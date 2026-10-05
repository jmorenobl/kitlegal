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

// errorDeUso es un error del sondeo que comete quien lo lanza: los de sus
// argumentos, unidos, uno por línea, o errSinSuscripcion, con el mismo texto
// (contracts/sondeo.md §1 de H7.4; research D18 de H7.4; FR-080). TestSondeo
// deja su mensaje en uso.txt y su guion lo imprime solo, sin el registro de go
// test. No lo son el de leer la definición del job, construir el binario,
// instalar las skills, repartir las sesiones ni crear el directorio de
// sesiones, que necesitan ese registro (FR-081).
type errorDeUso struct {
	causa error
}

// Error es el texto del error que envuelve.
func (e *errorDeUso) Error() string {
	return e.causa.Error()
}

// Unwrap devuelve el error que envuelve.
func (e *errorDeUso) Unwrap() error {
	return e.causa
}

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
// informativa o una con alguna sesión sin medir; y los títulos de los dos
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

	tituloSinMedirDelSondeo    = "Sesiones sin medir por límite de uso"
	tituloSinTerminarDelSondeo = "Sesiones que no termin" + "aron por otra causa"
	ningunaEnElSondeo          = ningunaEnElInforme + "."
)

// Las líneas del juez con modelo en la salida del sondeo, carácter a carácter
// (contracts/informe-del-job.md §8 de H24; FR-075, FR-076), que van donde hasta
// H24 iba la del recuento de las expresiones prohibidas: la que lo nombra, con
// su modelo; la de cada clase de clases.yaml, la que decide y la que solo se
// publica, con su nombre y su medida sobre las respuestas juzgadas; la de la
// medida versionada, que corresponde, con la versión de Claude Code del equipo,
// o que no corresponde, con lo que no coincide, o que no se puede comparar, sin
// esa versión; el título de las respuestas sin juzgar; y la de la skill sin
// juez.
const (
	tituloDelJuezDelSondeo     = "Juez (%s), sobre las respuestas de las evals que activan la skill:"
	claseQueDecideDelSondeo    = "- marcadas en %s, con sus tres votos: %s."
	claseQueSePublicaDelSondeo = "- con sí en %s, con un voto: %s."

	medidaQueCorrespondeDelSondeo = "La medida versionada del juez corresponde a lo que hay, " +
		"con el Claude Code de este equipo (%s)."
	medidaSinCorresponderDelSondeo = "La medida versionada del juez no corresponde a lo que hay: %s. " +
		"Estos recuentos no son los de un juez medido."
	medidaSinVersionDelSondeo = "La medida versionada del juez no se puede comparar con lo que hay: " +
		"ningún transcript de las sesiones declara la versión de Claude Code de este equipo. " +
		"No se sabe si estos recuentos son los de un juez medido."

	// versionQueNoEsLaDelEquipo es, en la línea de la medida que no
	// corresponde, lo que comprobarLaMedida dice con medidaDeOtraVersion: en el
	// sondeo no hay versión fijada para los votos del juez, sino la del equipo.
	versionQueNoEsLaDelEquipo = "la versión de Claude Code de sus votos es %s y la de este equipo es %s"

	// separadorDeLoQueNoCorresponde separa, en esa línea, lo que no coincide.
	separadorDeLoQueNoCorresponde = "; "

	tituloSinJuzgarDelSondeo = "Respuestas sin juzgar"
	sondeoSinJuez            = "La skill no tiene juez."
)

// SondeoAJuzgar es lo que el juicio del sondeo necesita de las sesiones que abrió
// el repartidor (contracts/sondeo.md §3.4 y §3.5 de H7.3; data-model §8): su plan
// —las evals pedidas con un solo modelo, sin modelos informativos ni prueba de
// red— y dónde están sus sesiones. Desde H24, además, con qué juzga sus
// respuestas el juez con modelo de la skill, si lo tiene
// (contracts/informe-del-job.md §8 de H24; FR-075).
type SondeoAJuzgar struct {
	// Skill es la skill sondeada: la activación que se busca.
	Skill string

	// Evals es el directorio de las evals de la skill, y Pedidas, las bien
	// formadas que se sondean, en el orden del plan.
	Evals   string
	Pedidas []Eval

	// Sesiones es el directorio con un subdirectorio por sesión.
	Sesiones string

	// Modelo es el único modelo del sondeo, y Repeticiones, las sesiones que el
	// plan abre de cada eval con él.
	Modelo       string
	Repeticiones int

	// SinAbrir son las sesiones del plan que el repartidor no abrió tras una con
	// el mensaje del límite de uso (FR-066).
	SinAbrir []SesionPlanificada

	// Juez es el juez con modelo de la skill, el de su carpeta de evals: nil si
	// no lo tiene, y entonces no se mira nada de lo que sigue.
	Juez *Juez

	// Votar es quien da cada voto del juez. Se le llama desde varias gorrutinas
	// a la vez, una por respuesta que se juzga.
	Votar Votante

	// ModeloDelJuez es el id del modelo del juez, el que fija la definición del
	// job, y ConcurrenciaDelJuez, las respuestas que se votan a la vez como
	// mucho, las de la concurrencia del sondeo.
	ModeloDelJuez       string
	ConcurrenciaDelJuez int
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

	// juez es lo que el juez con modelo de la skill deja de sus respuestas: nil
	// si la skill no tiene juez.
	juez *juezDelSondeo

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

// juezDelSondeo es lo que el juez con modelo de la skill deja de un sondeo
// (contracts/informe-del-job.md §8 de H24; research D12 de H24; FR-075,
// FR-076): solo lo agregado, sin los votos ni las frases de ninguna respuesta,
// y sin umbral ni veredicto, que el sondeo no decide nada.
type juezDelSondeo struct {
	// modelo es el id del modelo del juez: el que fija la definición del job.
	modelo string

	// clases son las del juez, en el orden de clases.yaml.
	clases []claseDelSondeo

	// juzgadas son las respuestas que el juez juzgó, el total de cada clase: las
	// del modelo del sondeo en las evals que activan la skill, terminadas y
	// medidas, menos las que quedaron sin juzgar.
	juzgadas int

	// sinJuzgar son las respuestas de las que un voto no llegó a darse, en orden
	// de sesión, cada una con el motivo de ese voto.
	sinJuzgar []RespuestaSinJuzgar

	// version es la de Claude Code del equipo de quien lanza el sondeo, la que
	// declaran los transcripts de sus sesiones: vacía si ninguno la declara.
	version string

	// sinCorresponder es lo que no coincide entre la medida versionada del juez y
	// lo que hay, con ese modelo y esa versión: nada si la medida corresponde y
	// se cumple, o si no hay versión con la que compararla.
	sinCorresponder []string
}

// claseDelSondeo es una clase del juez con las respuestas juzgadas del sondeo
// que deja marcadas en ella: con sus tres votos, si decide, y con el primero,
// si solo se publica.
type claseDelSondeo struct {
	nombre   string
	decide   bool
	marcadas int
}

// juzgarElSondeo juzga las sesiones del sondeo con el código del job, sin su
// traza y sin un juicio propio (contracts/sondeo.md §3.5 de H7.3; research D17;
// FR-061):
//
//  1. cada entrada de s.Sesiones, en orden de nombre, con juzgarSesionDelSondeo:
//     Juzgar sobre la eval sin sus comandos ni sus prohibidos y la sesión sin
//     invocaciones, exigirElModeloPedido y ClasificarElLimite;
//  2. las series con repartirEnSeries, como el informe, con el plan del sondeo:
//     las evals pedidas con su modelo como el que decide, en un solo modo, el
//     modo orden, sin modelos informativos ni prueba de red, y las sesiones que
//     el repartidor no abrió contadas como sin medir. Sin umbral: el sondeo no
//     dice si una serie llega a él (FR-066);
//  3. las sesiones sin medir con sesionesSinMedir, como el informe, y las que
//     quedaron sin terminar por otra causa (sesionesSinTerminar);
//  4. con una skill que tiene juez, sus respuestas con él (juzgarConElJuez),
//     que no cambia el juicio de ninguna sesión ni la tasa de ninguna serie
//     (contracts/informe-del-job.md §8 de H24; FR-075, FR-076).
//
// El error es solo para el directorio de sesiones que no se puede listar y para
// el esquema de la respuesta del juez que no sirve para validar sus votos.
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

	for _, juzgada := range sesiones {
		juicio.resultados = append(juicio.resultados, juzgada.resultado)
	}

	juicio.sinMedir = sesionesSinMedir(juicio.resultados, s.SinAbrir)
	juicio.sinTerminar = sesionesSinTerminar(sesiones)

	juicio.juez, err = s.juzgarConElJuez(e, sesiones, series)
	if err != nil {
		return juicioDelSondeo{}, fmt.Errorf("el sondeo no se puede juzgar: %w", err)
	}

	return juicio, nil
}

// juzgarConElJuez juzga con el juez con modelo de la skill las respuestas del
// sondeo, y da lo que deja de ellas (contracts/informe-del-job.md §8 de H24;
// contracts/medida-del-juez.md §2 de H24; research D12 de H24; FR-075, FR-076):
// nil si la skill no tiene juez, y entonces no se pide ningún voto.
//
//  1. Las respuestas son las que el juez juzga en el job (respuestasQueSeJuzgan)
//     con el plan del sondeo: las del modelo del sondeo en las evals que activan
//     la skill, terminadas y medidas. El plan es de un solo modo y sin evals sin
//     binario ni servidor, así que van en orden de sesión.
//  2. Cada una se juzga con el mensaje del voto y la regla de los votos, con el
//     votante recibido y como mucho tantas a la vez como diga la concurrencia
//     (juzgarTodas).
//  3. De cada clase lleva las que quedan marcadas, y de las que quedan sin
//     juzgar, su motivo: no están entre las juzgadas.
//  4. La versión de Claude Code del equipo es la que declaran los transcripts
//     (versionDelEquipo), y con ella y el modelo del juez se comprueba la medida
//     versionada solo para decirlo (medidaFrenteAlEquipo): corresponda o no,
//     las respuestas se juzgan igual.
//
// El error es el del esquema de la respuesta del juez que no compila.
func (s SondeoAJuzgar) juzgarConElJuez(
	e InformeAEscribir, sesiones []sesionJuzgada, series []serieJuzgada,
) (*juezDelSondeo, error) {
	if s.Juez == nil {
		return nil, nil
	}

	votacion, err := nuevaVotacion(s.Juez, s.Votar)
	if err != nil {
		return nil, err
	}

	var delSondeo grupoJuzgado
	for _, grupo := range respuestasQueSeJuzgan(e, sesiones, series) {
		delSondeo.respuestas = append(delSondeo.respuestas, grupo.respuestas...)
	}

	for posicion, juicio := range votacion.juzgarTodas(delSondeo.aJuzgar(sesiones), s.ConcurrenciaDelJuez) {
		delSondeo.respuestas[posicion].juicio = juicio
	}

	delJuez := &juezDelSondeo{modelo: s.ModeloDelJuez, version: versionDelEquipo(sesiones)}

	for posicion, clase := range s.Juez.Clases {
		delJuez.clases = append(delJuez.clases, claseDelSondeo{
			nombre: clase.Nombre, decide: clase.Decide, marcadas: len(delSondeo.marcadasEn(posicion)),
		})
	}

	for _, respuesta := range delSondeo.respuestas {
		if respuesta.juicio.SinJuzgar != "" {
			delJuez.sinJuzgar = append(delJuez.sinJuzgar,
				RespuestaSinJuzgar{Sesion: respuesta.sesion, Motivo: respuesta.juicio.SinJuzgar})
		}
	}

	delJuez.juzgadas = len(delSondeo.respuestas) - len(delJuez.sinJuzgar)

	if delJuez.version != "" {
		delJuez.sinCorresponder = medidaFrenteAlEquipo(s.Juez, s.ModeloDelJuez, delJuez.version)
	}

	return delJuez, nil
}

// versionDelEquipo es la versión de Claude Code del equipo de quien lanza el
// sondeo: la que declara el transcript de sus sesiones, que se abren todas con
// el claude de su PATH, el mismo que da los votos del juez (research D12 de
// H24). Es la de la primera sesión que la declara, en orden de sesión, y vacía
// si ninguna la declara: la de una sesión que no se pudo leer o cuyo claude no
// llegó a escribir nada.
func versionDelEquipo(sesiones []sesionJuzgada) string {
	for _, juzgada := range sesiones {
		if juzgada.sesion.VersionDeClaudeCode != "" {
			return juzgada.sesion.VersionDeClaudeCode
		}
	}

	return ""
}

// medidaFrenteAlEquipo es lo que no coincide entre la medida versionada del
// juez y lo que hay en el equipo de quien lanza el sondeo: las líneas de
// comprobarLaMedida con el modelo del juez y la versión de Claude Code de ese
// equipo, en su orden, y ninguna si la medida corresponde y se cumple
// (contracts/medida-del-juez.md §2 de H24; FR-076). La de la versión va con las
// palabras del sondeo (versionQueNoEsLaDelEquipo): comprobarLaMedida habla de la
// versión fijada para los votos del juez, que es la del job.
//
// Esa línea se reconoce por ser la que comprobarLaMedida da de la versión de la
// medida, que se lee de nuevo. Si la medida no se puede leer, no hay versión
// que nombrar, y comprobarLaMedida ya lo dice en su única línea.
func medidaFrenteAlEquipo(juez *Juez, modelo, version string) []string {
	lineas := comprobarLaMedida(juez, modelo, version)

	medida, err := leerMedidaDelJuez([]byte(juez.Medida))
	if err != nil {
		return lineas
	}

	deLaVersion := slices.Index(lineas, fmt.Sprintf(medidaDeOtraVersion, medida.VersionDeClaudeCode, version))
	if deLaVersion >= 0 {
		lineas[deLaVersion] = fmt.Sprintf(versionQueNoEsLaDelEquipo, medida.VersionDeClaudeCode, version)
	}

	return lineas
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
		// El plan del sondeo es de un solo modo, el modo orden: sus sesiones no
		// tienen servidor.json, y su serie es la de ese modo.
		clave: claveDeSerie{modelo: strings.TrimSuffix(string(modelo), "\n"), modo: ModoOrden},
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
			Modo:             juzgada.clave.modo,
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
// deciden el veredicto del job, o «sin medir» con alguna sesión sin medir; las
// líneas del juez; y las sesiones sin medir por límite de uso y las que quedaron
// sin terminar por otra causa, cada una con su motivo, o «ninguna.». Una línea
// en blanco separa cada parte, y cada sesión va en su línea. Desde H24 no lleva
// la línea del recuento de las respuestas con alguna expresión prohibida
// (FR-070 de H24): donde iba van las del juez con modelo de la skill
// (juezDelSondeo.lineas; contracts/informe-del-job.md §8 de H24; FR-075,
// FR-076).
func (j juicioDelSondeo) salida() string {
	lineas := []string{primeraLineaDelSondeo, segundaLineaDelSondeo, "", tituloDeLasSeriesDelSondeo}

	for _, serie := range j.series {
		lineas = append(lineas, "- "+serie.escrita())
	}

	lineas = append(lineas, "")
	lineas = append(lineas, j.juez.lineas()...)

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

// lineas son las del juez en la salida del sondeo, las de
// contracts/informe-del-job.md §8 de H24 (FR-075, FR-076): con una skill sin
// juez, la que lo dice; con él, la que lo nombra con su modelo; una por clase,
// en el orden de clases.yaml, con las respuestas que deja marcadas —con sus
// tres votos en la que decide, y con sí en su primer voto en la que solo se
// publica— sobre las juzgadas, escritas como la medida de un umbral del
// informe, «<n> de <t> (<p> %)»; la de la medida versionada (lineaDeLaMedida); y
// las respuestas sin juzgar, cada una en su línea con su motivo, o «ninguna.».
// Nada de los votos ni de las frases de ninguna respuesta.
func (j *juezDelSondeo) lineas() []string {
	if j == nil {
		return []string{sondeoSinJuez}
	}

	lineas := []string{fmt.Sprintf(tituloDelJuezDelSondeo, j.modelo)}

	for _, clase := range j.clases {
		forma := claseQueSePublicaDelSondeo
		if clase.decide {
			forma = claseQueDecideDelSondeo
		}

		medida := Umbral{Medida: float64(clase.marcadas), Total: &j.juzgadas}.medidaEscrita()
		lineas = append(lineas, fmt.Sprintf(forma, clase.nombre, medida))
	}

	sinJuzgar := make([]string, 0, len(j.sinJuzgar))
	for _, respuesta := range j.sinJuzgar {
		sinJuzgar = append(sinJuzgar, respuesta.Sesion+": "+respuesta.Motivo)
	}

	return append(append(lineas, j.lineaDeLaMedida()), listaDelSondeo(tituloSinJuzgarDelSondeo, sinJuzgar)...)
}

// lineaDeLaMedida es la línea de la salida del sondeo que dice si la medida
// versionada del juez corresponde al Claude Code con el que se ha votado
// (contracts/informe-del-job.md §8 de H24; FR-076): sin versión del equipo, que
// no se puede comparar; si nada deja de coincidir, que corresponde, con esa
// versión; y si no, lo que no coincide, en su orden, y que los recuentos no son
// los de un juez medido.
func (j *juezDelSondeo) lineaDeLaMedida() string {
	switch {
	case j.version == "":
		return medidaSinVersionDelSondeo
	case len(j.sinCorresponder) == 0:
		return fmt.Sprintf(medidaQueCorrespondeDelSondeo, j.version)
	default:
		return fmt.Sprintf(medidaSinCorresponderDelSondeo, strings.Join(j.sinCorresponder, separadorDeLoQueNoCorresponde))
	}
}

// apartadoDelSondeo son las líneas de un apartado de sesiones de la salida del
// sondeo, detrás de una en blanco (listaDelSondeo).
func apartadoDelSondeo(titulo string, sesiones []string) []string {
	return append([]string{""}, listaDelSondeo(titulo, sesiones)...)
}

// listaDelSondeo son las líneas de una lista de sesiones de la salida del
// sondeo: «<título>: ninguna.» sin ninguna, o el título con dos puntos y una
// línea «- <sesión>: <motivo>» por sesión, con los saltos de línea de su motivo
// como espacios.
func listaDelSondeo(titulo string, sesiones []string) []string {
	if len(sesiones) == 0 {
		return []string{titulo + ": " + ningunaEnElSondeo}
	}

	lineas := []string{titulo + ":"}
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

	// RutaDeLaDefinicionDelJob es la de la definición del job de evals, que dice
	// qué skills ejecuta y con qué concurrencia: rutaDeLaDefinicionDelJob en el
	// punto de entrada.
	RutaDeLaDefinicionDelJob string

	// Temporal es el directorio del sondeo, que crea y borra su guion: todo lo
	// que el sondeo escribe va dentro (FR-064), también lo que la preparación de
	// cada sesión y el go install escriben en el TMPDIR, que el guion pone en su
	// tmp/.
	Temporal string

	// Guion es la ruta absoluta de scripts/evals-sesion.sh, que abre cada sesión.
	Guion string

	// PrepararElArbol deja en el temporal, que recibe con su ruta absoluta, el
	// binario del árbol de trabajo en bin/ y sus skills en el directorio de
	// Claude Code de home/, con el entorno de quien lanza el sondeo como base:
	// prepararElArbol en el punto de entrada, y en los tests, un árbol que no
	// construye nada.
	PrepararElArbol func(temporal string, base []string) error

	// NuevoVotante da quien vota con el juez de la skill, que recibe con el id
	// de su modelo, el que fija la definición del job
	// (contracts/informe-del-job.md §8 de H24; FR-075): en el punto de entrada,
	// el votante que abre cada voto con el guion del voto, con el PATH y la
	// credencial de quien lanza el sondeo (nuevoVotanteDelGuion), y en los
	// tests, uno de salidas grabadas. Lo que cree para votar lo retira quien lo
	// da. Con una skill sin juez no se llama.
	NuevoVotante func(juez *Juez, modelo string) (Votante, error)
}

// sondeoComprobado es un sondeo cuyos argumentos y credencial valen, con sus
// argumentos ya interpretados.
type sondeoComprobado struct {
	// skill es la skill sondeada, y evals, la carpeta de sus evals.
	skill string
	evals string

	// pedidas son las evals de EVALS, en su orden.
	pedidas []Eval

	modelo       string
	repeticiones int
	concurrencia int

	// juez es el juez con modelo de la skill, el de la carpeta de sus evals, o
	// nil si no lo tiene, y modeloDelJuez, el id de su modelo, el que fija la
	// definición del job (contracts/informe-del-job.md §8 de H24).
	juez          *Juez
	modeloDelJuez string
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
//
// Los errores de los argumentos y el de la credencial van envueltos en
// errorDeUso, con su mismo texto (FR-080 de H7.4); el de una definición del job
// que no se puede leer, no (FR-081 de H7.4).
func comprobarElSondeo(s SondeoAEjecutar) (sondeoComprobado, error) {
	job, err := leerDefinicionDelJob(s.RutaDeLaDefinicionDelJob)
	if err != nil {
		return sondeoComprobado{}, err
	}

	comprobado, err := s.Argumentos.comprobar(s.EvalsDeLasSkills, job)
	if err != nil {
		return sondeoComprobado{}, &errorDeUso{causa: err}
	}

	if valorEnElEntorno(s.Entorno, variableDeLaSuscripcion) == "" {
		return sondeoComprobado{}, &errorDeUso{causa: errSinSuscripcion}
	}

	return comprobado, nil
}

// comprobar comprueba los argumentos del sondeo con la carpeta de evals de cada
// skill en evalsDeLasSkills y la definición del job (contracts/sondeo.md §3.1 de
// H7.3; research D16; FR-060, FR-067): SKILL, la forma de un nombre de skill,
// una carpeta de evals legible y sin ficheros mal formados, y en la matriz del
// job; EVALS, números de dos cifras separados por comas, sin repetir, cada uno
// el de una eval de la skill que no sea sin binario ni servidor (FR-050 de
// H21); MODELO, no vacío y con la forma de un id de modelo; REPETICIONES, un
// entero mayor o igual que 1; y CONCURRENCIA, vacía, la del job para la skill, o
// un entero mayor o igual que 1. Con una skill que no vale, de EVALS solo se
// comprueba la forma. Los errores de todos los argumentos van juntos, uno por
// línea y en ese orden, para que quien lanza la orden los corrija de una vez.
//
// Lo comprobado lleva además el juez de la carpeta de evals de la skill, si lo
// tiene, y el modelo del juez de la definición del job, con los que el sondeo
// juzga sus respuestas (contracts/informe-del-job.md §8 de H24; FR-075).
func (a ArgumentosDelSondeo) comprobar(evalsDeLasSkills string, job DefinicionDelJob) (sondeoComprobado, error) {
	comprobado := sondeoComprobado{
		skill: a.Skill, evals: filepath.Join(evalsDeLasSkills, a.Skill), modelo: a.Modelo,
		modeloDelJuez: job.ModeloDelJuez,
	}

	conjunto, errDeLaSkill := conjuntoDeLaSkill(a.Skill, comprobado.evals, job)
	comprobado.juez = conjunto.Juez

	numeros, errDeLasEvals := numerosDeLasEvals(a.Evals)
	if errDeLaSkill == nil && errDeLasEvals == nil {
		comprobado.pedidas, errDeLasEvals = evalsPedidas(numeros, a.Skill, conjunto)
	}

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
// un error que lo nombra, y el de una eval sin binario ni servidor, otro: el
// sondeo mide solo el modo orden, y esa eval solo la mide el job de evals
// (contracts/evals-en-dos-modos.md §7 de H21; FR-050 de H21). Uno por número,
// en el orden de los números.
func evalsPedidas(numeros []string, skill string, conjunto Conjunto) ([]Eval, error) {
	pedidas := make([]Eval, 0, len(numeros))

	var noValen []error

	for _, numero := range numeros {
		posicion := slices.IndexFunc(conjunto.Evals, func(eval Eval) bool {
			return strings.HasPrefix(eval.Fichero, numero+"-")
		})

		switch {
		case posicion < 0:
			noValen = append(noValen, fmt.Errorf("%s: %s no es ninguna eval de %s", argumentoEvals, numero, skill))
		case conjunto.Evals[posicion].SinBinarioNiServidor:
			noValen = append(noValen, fmt.Errorf("%s: %s es una eval sin binario ni servidor: solo la mide el job de evals",
				argumentoEvals, numero))
		default:
			pedidas = append(pedidas, conjunto.Evals[posicion])
		}
	}

	return pedidas, errors.Join(noValen...)
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
//     construir nada; y, con una skill que tiene juez, pide a NuevoVotante quien
//     vota con él y con el modelo del juez de la definición del job: sin con qué
//     votar no se construye nada ni se abre ninguna sesión
//     (contracts/informe-del-job.md §8 de H24; FR-075);
//  2. prepara el árbol de trabajo en el temporal con PrepararElArbol;
//  3. abre con el repartidor, en sesiones/ del temporal, el plan de las evals
//     pedidas en su orden, con MODELO como el modelo que decide, sin modelos
//     informativos ni prueba de red, y las repeticiones y la concurrencia
//     comprobadas: sin traza, con las skills que el árbol deja en el directorio
//     de Claude Code de home/, el entorno de entornoDelSondeo debajo del de cada
//     sesión y el tope de 240 s con su margen de 10 s. Tras una sesión con el
//     mensaje del límite de uso, el repartidor no abre ninguna más;
//  4. juzga las sesiones con juzgarElSondeo, con las que no se abrieron como sin
//     medir, y, con una skill que tiene juez, sus respuestas con él, con ese
//     votante y como mucho tantas a la vez como sesiones se abren a la vez; y
//     compone la salida.
//
// Vuelve sin error sean cuales sean las tasas y aunque la cuenta no deje abrir
// todas las sesiones: no escribe informe ni veredicto (FR-066). Tampoco lo dan
// lo que el juez marque, las respuestas que deje sin juzgar ni la medida
// versionada que no corresponda al Claude Code del equipo (FR-075 y FR-076 de
// H24): el sondeo no lista sus votos ni sus frases, no mide el modo herramienta
// ni lanza la medida del juez. El error es el de la comprobación, el de quien
// da el votante, el del árbol, el del repartidor —también tras la interrupción,
// con interrupcion cerrado—, el del directorio de sesiones o el del esquema de
// la respuesta del juez. Solo el de la comprobación de los argumentos o de la
// credencial es un errorDeUso, y con él no pide el votante, no prepara el árbol
// ni llama al repartidor (FR-080 y FR-081 de H7.4).
func sondear(interrupcion <-chan struct{}, s SondeoAEjecutar) (string, error) {
	comprobado, err := comprobarElSondeo(s)
	if err != nil {
		return "", err
	}

	var votar Votante
	if comprobado.juez != nil {
		votar, err = s.NuevoVotante(comprobado.juez, comprobado.modeloDelJuez)
		if err != nil {
			return "", fmt.Errorf("el votante del juez del sondeo no se puede preparar: %w", err)
		}
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
		Sesiones:     sesiones,
		Modelo:       comprobado.modelo,
		Repeticiones: comprobado.repeticiones,
		SinAbrir:     ejecucion.SinAbrir,

		Juez:                comprobado.juez,
		Votar:               votar,
		ModeloDelJuez:       comprobado.modeloDelJuez,
		ConcurrenciaDelJuez: comprobado.concurrencia,
	})
	if err != nil {
		return "", err
	}

	return juicio.salida(), nil
}
