package evals

import (
	"context"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"slices"
	"strconv"
	"strings"
)

// Los argumentos de make evals-sondeo, por su nombre, que es con lo que empieza
// cada error de su comprobación (contracts/sondeo.md §3.1 de H7.3; FR-067).
const (
	argumentoSkill        = "SKILL"
	argumentoEvals        = "EVALS"
	argumentoModelo       = "MODELO"
	argumentoRepeticiones = "REPETICIONES"
	argumentoConcurrencia = "CONCURRENCIA"
)

// Las formas de los argumentos del sondeo que no se comprueban con otra
// (contracts/sondeo.md §3.1 de H7.3): la de EVALS, números de eval de dos
// cifras separados por comas; y la de REPETICIONES y CONCURRENCIA, un entero
// mayor o igual que 1 escrito sin nada más, como exige scripts/evals.sh a la
// concurrencia del job. La skill tiene la forma de un nombre de skill, que es la
// del id de un modelo (formaDelModelo).
var (
	formaDeLasEvalsPedidas  = regexp.MustCompile(`^[0-9]{2}(,[0-9]{2})*$`)
	formaDeUnEnteroPositivo = regexp.MustCompile(`^[1-9][0-9]*$`)
)

// variableDeLaSuscripcion es la del token de la suscripción de Claude que da
// claude setup-token, la única credencial que acepta el sondeo y que pasa a sus
// sesiones (FR-063 de H7.3). El nombre no dice «token» ni «credencial»: gosec
// (G101) toma por secreto lo que se llama así.
const variableDeLaSuscripcion = "CLAUDE_CODE_OAUTH_TOKEN"

// errSinSuscripcion es el error del sondeo sin la credencial en el entorno, o
// con ella vacía (contracts/sondeo.md §3.2 de H7.3; FR-063).
var errSinSuscripcion = errors.New("falta la credencial: " + variableDeLaSuscripcion +
	", el token de la suscripción que da claude setup-token, no está en el entorno o está vacía")

// variablesQueVenLasSesionesDelSondeo son las variables de quien lanza el sondeo
// que ven sus sesiones, además del PATH, que ven con el bin/ del temporal
// delante (contracts/sondeo.md §5 de H7.3; research D16; FR-063): ninguna otra,
// ni ANTHROPIC_API_KEY ni ANTHROPIC_AUTH_TOKEN, que tendrían prioridad sobre la
// suscripción (research V5).
var variablesQueVenLasSesionesDelSondeo = []string{
	"LANG", "LC_ALL", "LC_CTYPE", "LC_MESSAGES", "TERM", "USER", "LOGNAME", "SHELL", "TZ", variableDeLaSuscripcion,
}

// El árbol de trabajo que el sondeo construye e instala (contracts/sondeo.md
// §3.3 de H7.3; data-model §8 de H7.3; FR-062).
const (
	// raizDelRepositorio es la del árbol de trabajo, relativa al directorio de
	// este paquete, que es donde go test ejecuta el punto de entrada del sondeo.
	raizDelRepositorio = "../.."

	// paqueteDelBinario es el paquete principal del binario, relativo a la raíz
	// del repositorio.
	paqueteDelBinario = "./cmd/kitlegal"

	// Las carpetas del temporal del sondeo: la de su binario, que es el GOBIN
	// del go install; la de su HOME, en la que el skills install deja las
	// skills; y la de sus sesiones.
	carpetaDeBinarios = "bin"
	carpetaPersonal   = "home"
	carpetaDeSesiones = "sesiones"

	// carpetaDeClaudeCode es, dentro de un HOME, la de Claude Code, en la que
	// skills install --host claude enlaza las skills.
	carpetaDeClaudeCode = ".claude"
)

// variableDeLaRuta es la de las carpetas en las que se buscan los ejecutables.
const variableDeLaRuta = "PATH"

// Textos fijos de la salida del sondeo (contracts/sondeo.md §4 de H7.3; FR-065):
// las dos primeras líneas, siempre las mismas, que dicen que no es un veredicto y
// lo que no comprueba; el título de las series y lo que lleva una serie
// informativa o una con alguna sesión sin medir; el recuento de las expresiones
// prohibidas, o la línea de la skill sin lista; y los títulos de los dos
// apartados de sesiones, con lo que dicen sin ninguna. El del segundo apartado va
// en dos literales partidos dentro de un verbo: misspell toma el pretérito de
// «terminar», suelto, por una errata inglesa.
const (
	primeraLineaDelSondeo = "Esto es un sondeo, no un veredicto: el veredicto de la skill lo da el job de evals."
	segundaLineaDelSondeo = "No comprueba lo que el job lee de la traza de strace —qué órdenes se ejecutaron " +
		"(los comandos esperados y los prohibidos de cada eval) y las llegadas a la red— ni la ausencia de Python: " +
		"por eso sus tasas no son las del job."

	tituloDeLasSeriesDelSondeo = "Tasa de cada serie (sesiones que pasan de las medidas):"
	serieInformativaDelSondeo  = " (informativa)"
	serieSinMedirDelSondeo     = "sin medir"

	recuentoDelSondeo = "Respuestas con alguna expresión prohibida en las evals que activan la skill: %s; " +
		"referencia: como mucho el %s %%."
	sondeoSinLista = "Respuestas con alguna expresión prohibida: " + sinListaDeExpresiones + "."

	tituloSinMedirDelSondeo    = "Sesiones sin medir por límite de uso"
	tituloSinTerminarDelSondeo = "Sesiones que no termin" + "aron por otra causa"
	ningunaEnElSondeo          = ningunaEnElInforme + "."
)

// SondeoAJuzgar es lo que el juicio del sondeo necesita de las sesiones que abrió
// el repartidor (contracts/sondeo.md §3.4 y §3.5 de H7.3; data-model §8): su plan
// —las evals pedidas con un solo modelo, sin modelos informativos ni prueba de
// red— y dónde están sus sesiones.
type SondeoAJuzgar struct {
	// Skill es la skill sondeada: la activación que se busca.
	Skill string

	// Evals es el directorio de las evals de la skill, y Pedidas, las bien
	// formadas que se sondean, en el orden del plan.
	Evals   string
	Pedidas []Eval

	// Prohibidas es la lista de expresiones prohibidas de la skill, la de su
	// directorio de evals: vacía si no la tiene.
	Prohibidas ExpresionesProhibidas

	// Sesiones es el directorio con un subdirectorio por sesión.
	Sesiones string

	// Modelo es el único modelo del sondeo, y Repeticiones, las sesiones que el
	// plan abre de cada eval con él.
	Modelo       string
	Repeticiones int

	// SinAbrir son las sesiones del plan que el repartidor no abrió tras una con
	// el mensaje del límite de uso (FR-066).
	SinAbrir []SesionPlanificada
}

// juicioDelSondeo es lo que el sondeo sabe de sus sesiones, juzgadas con el
// código del job y sin lo que el job saca de la traza (research D17 de H7.3;
// FR-061), y lo que publica de ellas.
type juicioDelSondeo struct {
	// resultados son los de cada sesión, en orden de sesión: los de informe.json
	// sin lo que sale de la traza.
	resultados []ResultadoDeEval

	// series son las del plan, en su orden.
	series []serieDelSondeo

	// recuento es el de las respuestas con alguna expresión prohibida de su
	// modelo, como el del informe (recontarExpresiones): nil si la skill no tiene
	// lista.
	recuento []RecuentoDeExpresiones

	// sinMedir son las sesiones que un límite de uso de la cuenta no dejó
	// terminar y las que el repartidor no abrió tras él, en orden de sesión, como
	// las del informe (sesionesSinMedir).
	sinMedir []SesionSinMedir

	// sinTerminar son las que quedaron sin terminar por otra causa, en orden de
	// sesión.
	sinTerminar []sesionSinTerminar
}

// serieDelSondeo es una serie del plan del sondeo con su tasa, sin umbral ni
// resultado: el sondeo no decide nada (FR-066).
type serieDelSondeo struct {
	eval   string
	modelo string

	// informativa dice que la serie no decide el veredicto del job: la de una
	// eval informativa.
	informativa bool

	// sesiones son las que se leyeron de la serie; pasan, cuántas de ellas pasan;
	// y sinMedir, las de la serie que quedaron sin medir, abiertas o no.
	sesiones int
	pasan    int
	sinMedir int
}

// sesionSinTerminar es una sesión juzgada que no terminó por otra causa que un
// límite de uso de la cuenta, con su motivo del job: «la sesión no terminó:
// <MotivoSinTerminar>».
type sesionSinTerminar struct {
	sesion string
	motivo string
}

// juzgarElSondeo juzga las sesiones del sondeo con el código del job, sin su
// traza y sin un juez propio (contracts/sondeo.md §3.5 de H7.3; research D17;
// FR-061):
//
//  1. cada entrada de s.Sesiones, en orden de nombre, con juzgarSesionDelSondeo:
//     Juzgar sobre la eval sin sus comandos ni sus prohibidos y la sesión sin
//     invocaciones, exigirElModeloPedido y ClasificarElLimite;
//  2. las series con repartirEnSeries y el recuento con recontarExpresiones,
//     como el informe, con el plan del sondeo: las evals pedidas con su modelo
//     como el que decide, sin modelos informativos ni prueba de red, y las
//     sesiones que el repartidor no abrió contadas como sin medir. Sin umbral:
//     el sondeo no dice si una serie llega a él (FR-066);
//  3. las sesiones sin medir con sesionesSinMedir, como el informe, y las que
//     quedaron sin terminar por otra causa (sesionesSinTerminar).
//
// El error es solo para el directorio de sesiones que no se puede listar.
func juzgarElSondeo(s SondeoAJuzgar) (juicioDelSondeo, error) {
	e := InformeAEscribir{
		Skill:           s.Skill,
		Evals:           s.Evals,
		Sesiones:        s.Sesiones,
		ModeloQueDecide: s.Modelo,
		Repeticiones:    s.Repeticiones,
		SinAbrir:        s.SinAbrir,
	}

	entradas, err := os.ReadDir(s.Sesiones)
	if err != nil {
		return juicioDelSondeo{}, fmt.Errorf("el sondeo no se puede juzgar: el directorio de sesiones %s no se puede listar: %w",
			s.Sesiones, err)
	}

	sesiones := make([]sesionJuzgada, 0, len(entradas))
	for _, entrada := range entradas {
		sesiones = append(sesiones, juzgarSesionDelSondeo(e, s.Pedidas, entrada.Name()))
	}

	var juicio juicioDelSondeo

	series := repartirEnSeries(e, s.Pedidas, sesiones)
	for _, serie := range series {
		for _, posicion := range serie.sesiones {
			sesiones[posicion].resultado.Decide = serie.tasa.Decide
		}

		if serie.tasa.Planificada {
			juicio.series = append(juicio.series, serieDelSondeo{
				eval: serie.tasa.Eval, modelo: serie.tasa.Modelo, informativa: !serie.tasa.Decide,
				sesiones: serie.tasa.Sesiones, pasan: serie.tasa.Pasan, sinMedir: serie.tasa.SinMedir,
			})
		}
	}

	juicio.recuento = recontarExpresiones(e, s.Prohibidas, sesiones, series)

	for _, juzgada := range sesiones {
		juicio.resultados = append(juicio.resultados, juzgada.resultado)
	}

	juicio.sinMedir = sesionesSinMedir(juicio.resultados, s.SinAbrir)
	juicio.sinTerminar = sesionesSinTerminar(sesiones)

	return juicio, nil
}

// juzgarSesionDelSondeo lee la sesión del subdirectorio nombre sin su traza y la
// juzga como el job, sin lo que el job saca de ella (research D17 de H7.3;
// FR-061): lee eval.txt, modelo.txt, pregunta.txt y la sesión con LeerSesion,
// que la deja sin invocaciones, y, si todo se leyó y eval.txt nombra una de las
// evals pedidas, la juzga con Juzgar sobre esa eval sin sus comandos ni sus
// prohibidos, le exige el modelo pedido y la clasifica con ClasificarElLimite,
// como EscribirInforme. Si no, la sesión no pasa, con un motivo «sesión ilegible:
// <fichero>: <error>» por cada fichero que no se pudo leer, en ese orden, y lleva
// lo que sí se leyó de ella, como en el informe.
func juzgarSesionDelSondeo(e InformeAEscribir, evals []Eval, nombre string) sesionJuzgada {
	dir := filepath.Join(e.Sesiones, nombre)

	nombreDeEval, eval, errDeEval := leerEvalDeLaSesion(dir, e.Evals, evals)
	modelo, errDelModelo := leerFicheroDeSesion(dir, ficheroDelModelo)
	pregunta, errDeLaPregunta := leerFicheroDeSesion(dir, ficheroDeLaPregunta)
	sesion, errDeLaSesion := LeerSesion(dir)

	juzgada := sesionJuzgada{
		sesion:        sesion,
		leida:         errDeLaSesion == nil,
		pregunta:      string(pregunta),
		preguntaLeida: errDeLaPregunta == nil,
		clave:         claveDeSerie{modelo: strings.TrimSuffix(string(modelo), "\n")},
	}

	if errDeEval == nil {
		juzgada.clave.eval = nombreDeEval
		juzgada.clave.preguntaAmpliada = juzgada.preguntaLeida && juzgada.pregunta != eval.Pregunta+"\n"
	}

	var motivos []string

	for _, err := range []error{errDeEval, errDelModelo, errDeLaPregunta, errDeLaSesion} {
		if err != nil {
			motivos = append(motivos, motivoDeSesionIlegible+err.Error())
		}
	}

	if len(motivos) > 0 {
		juzgada.ilegible = true
		juzgada.resultado = ResultadoDeEval{
			Sesion:           nombre,
			Eval:             nombreDeEval,
			Modelo:           juzgada.clave.modelo,
			ModeloDeLaSesion: sesion.Modelo,
			Activa:           eval.Activa,
			Motivos:          motivos,
		}

		if juzgada.leida {
			juzgada.resultado.observar(sesion, e.Skill)
		}

		return juzgada
	}

	juzgada.resultado = Juzgar(sinLoQueSaleDeLaTraza(eval), sesion, e.Skill)
	juzgada.resultado.Sesion = nombre
	juzgada.resultado.Modelo = juzgada.clave.modelo
	juzgada.resultado.ModeloDeLaSesion = sesion.Modelo
	juzgada.resultado.exigirElModeloPedido()
	juzgada.resultado.dejarSinMedir(ClasificarElLimite(sesion))

	return juzgada
}

// observar pone en el resultado de una sesión ilegible lo observado en ella, si
// LeerSesion la leyó: si activó la skill, su respuesta, su código, su fin, si
// terminó y sus reintentos por rate_limit, como el informe.
func (r *ResultadoDeEval) observar(sesion Sesion, skill string) {
	codigo := sesion.Codigo

	r.Activada = sesion.Activada(skill)
	r.Respuesta = sesion.Respuesta
	r.CodigoDeLaSesion = &codigo
	r.FinDeLaSesion = sesion.Fin
	r.SesionTerminada = sesion.Terminada
	r.ReintentosPorLimiteDeRitmo = sesion.ReintentosPorLimiteDeRitmo()
}

// sinLoQueSaleDeLaTraza es la eval sin lo que el job juzga con la traza: sus
// comandos esperados, incluida la comprobación de la redacción, y sus comandos
// prohibidos (research D17 de H7.3; FR-061). Con ella, Juzgar deja vacíos los
// comandos ejecutados y ausentes y los prohibidos ejecutados, y su juicio es el
// de antes en todo lo demás.
func sinLoQueSaleDeLaTraza(eval Eval) Eval {
	eval.Comandos, eval.Prohibidos = nil, nil

	return eval
}

// sesionesSinTerminar son las sesiones juzgadas que quedaron sin terminar por
// otra causa que un límite de uso de la cuenta, en orden de sesión, con su motivo
// del job (contracts/sondeo.md §4 de H7.3; FR-065). Las ilegibles no están: no se
// sabe si acabaron.
func sesionesSinTerminar(sesiones []sesionJuzgada) []sesionSinTerminar {
	var sinTerminar []sesionSinTerminar

	for _, juzgada := range sesiones {
		if juzgada.ilegible || juzgada.resultado.SinMedir != "" || juzgada.sesion.Terminada {
			continue
		}

		sinTerminar = append(sinTerminar, sesionSinTerminar{
			sesion: juzgada.resultado.Sesion,
			motivo: motivoDeSesionSinTerminar + juzgada.sesion.MotivoSinTerminar,
		})
	}

	return sinTerminar
}

// salida es la salida del sondeo, la que su guion imprime (contracts/sondeo.md
// §4 de H7.3; FR-065): las dos líneas fijas; la tasa de cada serie del plan,
// «<pasan> de <sesiones>», con « (informativa)» detrás del modelo en las que no
// deciden el veredicto del job, o «sin medir» con alguna sesión sin medir; el
// recuento de las respuestas con alguna expresión prohibida con el 5 % como
// referencia, o la línea de la skill sin lista; y las sesiones sin medir por
// límite de uso y las que quedaron sin terminar por otra causa, cada una con su
// motivo, o «ninguna.». Una línea en blanco separa cada parte. Nada por sesión de las
// expresiones prohibidas, y cada sesión en su línea.
func (j juicioDelSondeo) salida() string {
	lineas := []string{primeraLineaDelSondeo, segundaLineaDelSondeo, "", tituloDeLasSeriesDelSondeo}

	for _, serie := range j.series {
		lineas = append(lineas, "- "+serie.escrita())
	}

	lineas = append(lineas, "", j.recuentoEscrito())

	sinMedir := make([]string, 0, len(j.sinMedir))
	for _, sesion := range j.sinMedir {
		sinMedir = append(sinMedir, sesion.Sesion+": "+sesion.Motivo)
	}

	sinTerminar := make([]string, 0, len(j.sinTerminar))
	for _, sesion := range j.sinTerminar {
		sinTerminar = append(sinTerminar, sesion.sesion+": "+sesion.motivo)
	}

	lineas = append(lineas, apartadoDelSondeo(tituloSinMedirDelSondeo, sinMedir)...)
	lineas = append(lineas, apartadoDelSondeo(tituloSinTerminarDelSondeo, sinTerminar)...)

	return strings.Join(lineas, "\n") + "\n"
}

// escrita es la serie como la escribe la salida del sondeo: «<eval> con
// <modelo>», « (informativa)» si lo es, y su tasa o «sin medir».
func (s serieDelSondeo) escrita() string {
	serie := s.eval + " con " + s.modelo
	if s.informativa {
		serie += serieInformativaDelSondeo
	}

	if s.sinMedir > 0 {
		return serie + ": " + serieSinMedirDelSondeo
	}

	return serie + ": " + strconv.Itoa(s.pasan) + " de " + strconv.Itoa(s.sesiones)
}

// recuentoEscrito es la línea del recuento de la salida del sondeo: con la
// medida escrita como la del umbral de las expresiones del informe, «<n> de <m>
// (<p> %)», y el umbral del paquete como referencia; o la de la skill sin lista.
// El sondeo tiene un solo modelo, así que su recuento tiene un solo elemento.
func (j juicioDelSondeo) recuentoEscrito() string {
	if len(j.recuento) == 0 {
		return sondeoSinLista
	}

	respuestas := j.recuento[0].Respuestas
	medida := Umbral{Medida: float64(j.recuento[0].ConAlguna), Total: &respuestas}.medidaEscrita()
	referencia := strings.Replace(strconv.FormatFloat(umbralDeExpresionesProhibidas*100, 'f', -1, 64), ".", ",", 1)

	return fmt.Sprintf(recuentoDelSondeo, medida, referencia)
}

// apartadoDelSondeo son las líneas de un apartado de sesiones de la salida del
// sondeo, detrás de una en blanco: «<título>: ninguna.» sin ninguna, o el título
// con dos puntos y una línea «- <sesión>: <motivo>» por sesión, con los saltos
// de línea de su motivo como espacios.
func apartadoDelSondeo(titulo string, sesiones []string) []string {
	if len(sesiones) == 0 {
		return []string{"", titulo + ": " + ningunaEnElSondeo}
	}

	lineas := []string{"", titulo + ":"}
	for _, sesion := range sesiones {
		lineas = append(lineas, "- "+enUnaLinea.Replace(sesion))
	}

	return lineas
}

// ArgumentosDelSondeo son los de make evals-sondeo tal como los escribe quien
// lo lanza, sin interpretar (contracts/sondeo.md §1 y §3.1 de H7.3; data-model
// §8 de H7.3): la skill, los números de sus evals separados por comas, el id
// del modelo, las repeticiones y la concurrencia, vacía para la del job.
type ArgumentosDelSondeo struct {
	Skill        string
	Evals        string
	Modelo       string
	Repeticiones string
	Concurrencia string
}

// SondeoAEjecutar es lo que sondear necesita para abrir y juzgar las sesiones de
// un sondeo (contracts/sondeo.md §3 de H7.3; data-model §8 de H7.3).
type SondeoAEjecutar struct {
	// Argumentos son los de make, sin comprobar.
	Argumentos ArgumentosDelSondeo

	// Entorno es el de quien lanza el sondeo: de él salen la credencial, el
	// entorno del go install y lo poco que ven las sesiones.
	Entorno []string

	// EvalsDeLasSkills es el directorio con la carpeta de evals de cada skill.
	EvalsDeLasSkills string

	// Temporal es el directorio del sondeo, que crea y borra su guion: todo lo
	// que el sondeo escribe va dentro (FR-064).
	Temporal string

	// Guion es la ruta absoluta de scripts/evals-sesion.sh, que abre cada sesión.
	Guion string

	// PrepararElArbol deja en el temporal, que recibe con su ruta absoluta, el
	// binario del árbol de trabajo en bin/ y sus skills en el directorio de
	// Claude Code de home/, con el entorno de quien lanza el sondeo como base:
	// prepararElArbol en el punto de entrada, y en los tests, un árbol que no
	// construye nada.
	PrepararElArbol func(temporal string, base []string) error
}

// sondeoComprobado es un sondeo cuyos argumentos y credencial valen, con sus
// argumentos ya interpretados.
type sondeoComprobado struct {
	// skill es la skill sondeada, y evals, la carpeta de sus evals.
	skill string
	evals string

	// pedidas son las evals de EVALS, en su orden, y prohibidas, la lista de
	// expresiones prohibidas de la carpeta.
	pedidas    []Eval
	prohibidas ExpresionesProhibidas

	modelo       string
	repeticiones int
	concurrencia int
}

// comprobarElSondeo comprueba el sondeo antes de construir nada, en este orden
// (contracts/sondeo.md §3.1 y §3.2 de H7.3; FR-063, FR-067):
//
//  1. los argumentos, con la definición del job de evals del repositorio, que
//     dice qué skills ejecuta y con qué concurrencia (comprobar): con alguno que
//     no vale, el error nombra cada uno, uno por línea;
//  2. la credencial: CLAUDE_CODE_OAUTH_TOKEN en el entorno y no vacía, o
//     errSinSuscripcion. Sin llamar al servicio ni ninguna otra comprobación: una
//     credencial caducada se ve en la primera sesión.
func comprobarElSondeo(s SondeoAEjecutar) (sondeoComprobado, error) {
	job, err := leerDefinicionDelJob(rutaDeLaDefinicionDelJob)
	if err != nil {
		return sondeoComprobado{}, err
	}

	comprobado, err := s.Argumentos.comprobar(s.EvalsDeLasSkills, job)
	if err != nil {
		return sondeoComprobado{}, err
	}

	if valorEnElEntorno(s.Entorno, variableDeLaSuscripcion) == "" {
		return sondeoComprobado{}, errSinSuscripcion
	}

	return comprobado, nil
}

// comprobar comprueba los argumentos del sondeo con la carpeta de evals de cada
// skill en evalsDeLasSkills y la definición del job (contracts/sondeo.md §3.1 de
// H7.3; research D16; FR-060, FR-067): SKILL, la forma de un nombre de skill,
// una carpeta de evals legible y sin ficheros mal formados, y en la matriz del
// job; EVALS, números de dos cifras separados por comas, sin repetir, cada uno
// el de una eval de la skill; MODELO, no vacío y con la forma de un id de
// modelo; REPETICIONES, un entero mayor o igual que 1; y CONCURRENCIA, vacía,
// la del job para la skill, o un entero mayor o igual que 1. Con una skill que
// no vale, de EVALS solo se comprueba la forma. Los errores de todos los
// argumentos van juntos, uno por línea y en ese orden, para que quien lanza la
// orden los corrija de una vez.
func (a ArgumentosDelSondeo) comprobar(evalsDeLasSkills string, job DefinicionDelJob) (sondeoComprobado, error) {
	comprobado := sondeoComprobado{skill: a.Skill, evals: filepath.Join(evalsDeLasSkills, a.Skill), modelo: a.Modelo}

	conjunto, errDeLaSkill := conjuntoDeLaSkill(a.Skill, comprobado.evals, job)

	numeros, errDeLasEvals := numerosDeLasEvals(a.Evals)
	if errDeLaSkill == nil && errDeLasEvals == nil {
		comprobado.pedidas, errDeLasEvals = evalsPedidas(numeros, a.Skill, conjunto)
	}

	comprobado.prohibidas = conjunto.Prohibidas

	var errDeLasRepeticiones, errDeLaConcurrencia error

	comprobado.repeticiones, errDeLasRepeticiones = enteroDelArgumento(argumentoRepeticiones, a.Repeticiones)

	comprobado.concurrencia = job.PorSkill[a.Skill].Concurrencia
	if a.Concurrencia != "" {
		comprobado.concurrencia, errDeLaConcurrencia = enteroDelArgumento(argumentoConcurrencia, a.Concurrencia)
	}

	if err := errors.Join(errDeLaSkill, errDeLasEvals, comprobarElModelo(a.Modelo), errDeLasRepeticiones,
		errDeLaConcurrencia); err != nil {
		return sondeoComprobado{}, err
	}

	return comprobado, nil
}

// conjuntoDeLaSkill es el conjunto de evals de la skill, de su carpeta evals, si
// la skill tiene la forma de un nombre de skill, su carpeta se puede leer y no
// tiene ficheros mal formados y el job de evals la ejecuta.
func conjuntoDeLaSkill(skill, evals string, job DefinicionDelJob) (Conjunto, error) {
	sinEvals := fmt.Errorf("%s: «%s» no es ninguna skill con evals", argumentoSkill, skill)

	if !formaDelModelo.MatchString(skill) {
		return Conjunto{}, sinEvals
	}

	conjunto, err := LeerConjunto(evals)
	if err != nil || len(conjunto.MalFormados) > 0 {
		return Conjunto{}, sinEvals
	}

	if !slices.Contains(job.Skills, skill) {
		return Conjunto{}, fmt.Errorf("%s: el job de evals no ejecuta «%s»", argumentoSkill, skill)
	}

	return conjunto, nil
}

// numerosDeLasEvals son los números de eval de EVALS, en su orden, si tiene la
// forma de números de dos cifras separados por comas y ninguno se repite: uno
// repetido abriría dos veces las mismas sesiones.
func numerosDeLasEvals(valor string) ([]string, error) {
	numeros := strings.Split(valor, ",")
	repetido := len(slices.Compact(slices.Sorted(slices.Values(numeros)))) < len(numeros)

	if !formaDeLasEvalsPedidas.MatchString(valor) || repetido {
		return nil, fmt.Errorf("%s: «%s» no es una lista de números de eval de dos cifras separados por comas",
			argumentoEvals, valor)
	}

	return numeros, nil
}

// evalsPedidas son las evals del conjunto cuyo fichero empieza por cada número
// y un guion, en el orden de los números. Un número que no es el de ninguna es
// un error que lo nombra, uno por número.
func evalsPedidas(numeros []string, skill string, conjunto Conjunto) ([]Eval, error) {
	pedidas := make([]Eval, 0, len(numeros))

	var faltan []error

	for _, numero := range numeros {
		posicion := slices.IndexFunc(conjunto.Evals, func(eval Eval) bool {
			return strings.HasPrefix(eval.Fichero, numero+"-")
		})
		if posicion < 0 {
			faltan = append(faltan, fmt.Errorf("%s: %s no es ninguna eval de %s", argumentoEvals, numero, skill))

			continue
		}

		pedidas = append(pedidas, conjunto.Evals[posicion])
	}

	return pedidas, errors.Join(faltan...)
}

// comprobarElModelo exige un MODELO no vacío y con la forma de un id de modelo.
func comprobarElModelo(modelo string) error {
	switch {
	case modelo == "":
		return fmt.Errorf("%s: está vacío", argumentoModelo)
	case !formaDelModelo.MatchString(modelo):
		return fmt.Errorf("%s: «%s» no tiene la forma de un id de modelo", argumentoModelo, modelo)
	default:
		return nil
	}
}

// enteroDelArgumento es el entero del valor del argumento, si es un entero
// mayor o igual que 1 escrito sin nada más y cabe en un int.
func enteroDelArgumento(argumento, valor string) (int, error) {
	entero, err := strconv.Atoi(valor)
	if err != nil || !formaDeUnEnteroPositivo.MatchString(valor) {
		return 0, fmt.Errorf("%s: «%s» no es un entero mayor o igual que 1", argumento, valor)
	}

	return entero, nil
}

// valorEnElEntorno es el valor de la variable en el entorno, nombre=valor, o
// vacío si no está. Si está más de una vez, vale la última, como en el entorno
// que exec da a un proceso.
func valorEnElEntorno(entorno []string, nombre string) string {
	valor := ""

	for _, variable := range entorno {
		if n, v, _ := strings.Cut(variable, "="); n == nombre {
			valor = v
		}
	}

	return valor
}

// prepararElArbol deja en el temporal del sondeo, que recibe con su ruta
// absoluta, el binario y las skills del árbol de trabajo, no los que tenga
// instalados quien lo lanza (contracts/sondeo.md §3.3 de H7.3; research D8 y
// D16; FR-062):
//
//  1. go install -trimpath del paquete principal del binario, en la raíz del
//     repositorio, con el entorno de la base, el GOBIN en bin/ del temporal y
//     CGO_ENABLED=0: las cachés de Go son las de quien lo lanza (FR-064);
//  2. skills install -g --host claude con el kitlegal recién construido, con el
//     entorno de las sesiones del sondeo (entornoDelSondeo), que tiene el HOME
//     en home/ del temporal: las skills quedan en su .claude/skills, como las
//     deja make install, y nada del binario sale del temporal.
//
// Cada orden la construye una función que recibe su ruta (research D8). El
// error nombra lo que no se pudo hacer, con la orden y lo que escribió.
func prepararElArbol(temporal string, base []string) error {
	binarios := filepath.Join(temporal, carpetaDeBinarios)

	construir := ordenDeConstruirElBinario(raizDelRepositorio)
	construir.Env = sobreLaBase(base, []string{"GOBIN=" + binarios, "CGO_ENABLED=0"})

	if err := ejecutarLaOrdenDelArbol(construir); err != nil {
		return fmt.Errorf("el binario del sondeo no se pudo construir: %w", err)
	}

	personal := filepath.Join(temporal, carpetaPersonal)
	if err := os.Mkdir(personal, permisosDeLaSesion); err != nil {
		return fmt.Errorf("el HOME del sondeo %s no se puede crear: %w", personal, err)
	}

	instalar := ordenDeInstalarLasSkills(filepath.Join(binarios, programaDeLasConsultas))
	instalar.Env = entornoDelSondeo(base, temporal)

	if err := ejecutarLaOrdenDelArbol(instalar); err != nil {
		return fmt.Errorf("las skills del sondeo no se pudieron instalar: %w", err)
	}

	return nil
}

// ordenDeConstruirElBinario es go install -trimpath del paquete principal del
// binario en la raíz dada, toda con constantes (gosec G204).
func ordenDeConstruirElBinario(raiz string) *exec.Cmd {
	orden := exec.CommandContext(context.Background(), "go", "install", "-trimpath", paqueteDelBinario)
	orden.Dir = raiz

	return orden
}

// ordenDeInstalarLasSkills es skills install -g --host claude con el kitlegal
// de la ruta dada, el único argumento que no es constante (gosec G204).
func ordenDeInstalarLasSkills(kitlegal string) *exec.Cmd {
	return exec.CommandContext(context.Background(), kitlegal, "skills", "install", "-g", "--host", "claude")
}

// ejecutarLaOrdenDelArbol ejecuta la orden y, si no termina con 0, devuelve un
// error con la orden, cómo terminó y lo que escribió en sus dos salidas.
func ejecutarLaOrdenDelArbol(orden *exec.Cmd) error {
	salida, err := orden.CombinedOutput()
	if err != nil {
		return fmt.Errorf("%s: %w\n%s", orden, err, salida)
	}

	return nil
}

// entornoDelSondeo es el entorno de quien lanza el sondeo que ven sus sesiones y
// su skills install (contracts/sondeo.md §5 de H7.3; research D16; FR-063): de
// la base, solo las variables de variablesQueVenLasSesionesDelSondeo, en su
// orden, y detrás el PATH de la base con el bin/ del temporal delante —solo ese
// bin/ si la base no tiene PATH— y el HOME en home/ del temporal. Lo de cada
// sesión va encima (entornoDeLaSesion).
func entornoDelSondeo(base []string, temporal string) []string {
	binarios := filepath.Join(temporal, carpetaDeBinarios)
	ruta := binarios

	var entorno []string

	for _, variable := range base {
		nombre, valor, _ := strings.Cut(variable, "=")

		switch {
		case nombre == variableDeLaRuta:
			ruta = binarios + string(os.PathListSeparator) + valor
		case slices.Contains(variablesQueVenLasSesionesDelSondeo, nombre):
			entorno = append(entorno, variable)
		}
	}

	return append(entorno, variableDeLaRuta+"="+ruta, "HOME="+filepath.Join(temporal, carpetaPersonal))
}

// sondear abre y juzga las sesiones de un sondeo, y devuelve su salida
// (contracts/sondeo.md §3 y §4 de H7.3; research D10, D16 y D17; FR-060 a
// FR-066):
//
//  1. comprueba los argumentos y la credencial (comprobarElSondeo), antes de
//     construir nada;
//  2. prepara el árbol de trabajo en el temporal con PrepararElArbol;
//  3. abre con el repartidor, en sesiones/ del temporal, el plan de las evals
//     pedidas en su orden, con MODELO como el modelo que decide, sin modelos
//     informativos ni prueba de red, y las repeticiones y la concurrencia
//     comprobadas: sin traza, con las skills que el árbol deja en el directorio
//     de Claude Code de home/, el entorno de entornoDelSondeo debajo del de cada
//     sesión y el tope de 240 s con su margen de 10 s. Tras una sesión con el
//     mensaje del límite de uso, el repartidor no abre ninguna más;
//  4. juzga las sesiones con juzgarElSondeo, con las que no se abrieron como sin
//     medir, y compone la salida.
//
// Vuelve sin error sean cuales sean las tasas y aunque la cuenta no deje abrir
// todas las sesiones: no escribe informe ni veredicto (FR-066). El error es el
// de la comprobación, el del árbol, el del repartidor —también tras la
// interrupción, con interrupcion cerrado— o el del directorio de sesiones.
func sondear(interrupcion <-chan struct{}, s SondeoAEjecutar) (string, error) {
	comprobado, err := comprobarElSondeo(s)
	if err != nil {
		return "", err
	}

	temporal, err := filepath.Abs(s.Temporal)
	if err != nil {
		return "", fmt.Errorf("el temporal del sondeo %s no tiene ruta absoluta: %w", s.Temporal, err)
	}

	if err := s.PrepararElArbol(temporal, s.Entorno); err != nil {
		return "", err
	}

	sesiones := filepath.Join(temporal, carpetaDeSesiones)
	if err := os.Mkdir(sesiones, permisosDeLaSesion); err != nil {
		return "", fmt.Errorf("el directorio de sesiones del sondeo %s no se puede crear: %w", sesiones, err)
	}

	plan := PlanDeEvals{
		Evals: comprobado.pedidas, ModeloQueDecide: comprobado.modelo, Repeticiones: comprobado.repeticiones,
	}

	ejecucion, err := ejecutarSesiones(interrupcion, SesionesAEjecutar{
		Plan:          plan.Sesiones(),
		Concurrencia:  comprobado.concurrencia,
		Evals:         comprobado.evals,
		Sesiones:      sesiones,
		Skills:        filepath.Join(temporal, carpetaPersonal, carpetaDeClaudeCode, directorioDeSkills),
		Guion:         s.Guion,
		Entorno:       entornoDelSondeo(s.Entorno, temporal),
		Traza:         false,
		Tope:          topeDeUnaSesion,
		MargenDelTope: margenDelTopeDeUnaSesion,
	})
	if err != nil {
		return "", err
	}

	juicio, err := juzgarElSondeo(SondeoAJuzgar{
		Skill:        comprobado.skill,
		Evals:        comprobado.evals,
		Pedidas:      comprobado.pedidas,
		Prohibidas:   comprobado.prohibidas,
		Sesiones:     sesiones,
		Modelo:       comprobado.modelo,
		Repeticiones: comprobado.repeticiones,
		SinAbrir:     ejecucion.SinAbrir,
	})
	if err != nil {
		return "", err
	}

	return juicio.salida(), nil
}
