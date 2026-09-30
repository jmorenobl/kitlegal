package evals

import (
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"
)

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
