package evals

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"slices"
	"strconv"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// Los modelos del job de cierre de boe-legislacion (contrato informe-del-job §1
// de H7.3; ADR 0031): el que decide y el que solo se publica, con los que
// TestUmbralesDelInforme y TestInformeMarkdownDeLosUmbrales arman sus
// ejecuciones; y modeloSonnet5, el que decidía hasta el ADR 0031, con el que
// los tests del sondeo, que admite cualquier modelo, arman los suyos.
const (
	modeloSonnet55 = "claude-sonnet-5-5"
	modeloHaiku45  = "claude-haiku-4-5-20251001"
	modeloSonnet5  = "claude-sonnet-5"
)

// Lo que escribirEjecucionConUmbrales pone en las sesiones que altera
// (contracts/informe-del-job.md §6 de H7.4).
const (
	// redaccionNoLeida cuenta una redacción que ninguna orden devolvió: lleva ya
	// no exige, la expresión de redaccion_no_leida de la lista del repositorio, y
	// ninguna otra. Con la transición de la memoria de consultas delante, es lo
	// que se antepone a una respuesta con expresiones de la lista.
	redaccionNoLeida = "Tras la reforma, el precepto ya no exige esa firma."

	// notificacionDeUnaTarea es la notificación de una tarea en segundo plano que
	// llega detrás del result (research V4 de H7.4): con ella esperaba la sesión
	// cuando el tope la cortó.
	notificacionDeUnaTarea = `{"type":"system","subtype":"task_notification","task_id":"tarea_sintetica_01",` +
		`"tool_use_id":"toolu_sintetico_02","session_id":"00000000-0000-4000-8000-000000000101"}` + "\n"

	// motivoDelTopeEsperado es el único motivo de una sesión que el tope cortó
	// tras su respuesta.
	motivoDelTopeEsperado = "la sesión no terminó: tope de 240 s agotado (código 124)"
)

// formaDelNombreDeUmbral es la del nombre de un umbral del contrato del ADR 0029.
var formaDelNombreDeUmbral = regexp.MustCompile(`^[a-z0-9_.:-]+$`)

// ejecucionConUmbrales es una ejecución sintética del job de evals, armada en un
// directorio temporal del test por escribirEjecucionConUmbrales con tres
// repeticiones por serie y umbral 2, como el job (ADR 0016). Sus evals son copias
// de la del art. 21 del caso aprobado, que activa la skill: primero las que
// deciden, después las informativas y detrás las informativas sin terminar; y,
// si se pide, la de no activación del caso aprobado, que decide, y una eval sin
// binario ni servidor. Cada sesión es una copia de la de su eval en el caso
// aprobado, que pasa, con su eval y su modelo; con dosModos, cada eval de modo
// tiene además sus sesiones del modo herramienta, que consultan con una llamada
// y pasan igual (contracts/evals-en-dos-modos.md §2.1 de H21).
type ejecucionConUmbrales struct {
	// queDeciden e informativas son cuántas evals de cada clase lleva.
	queDeciden, informativas int

	// conHaiku dice si modeloHaiku45 está entre los modelos informativos del job,
	// y abre solo las evals que deciden; el que decide es siempre modeloSonnet55.
	conHaiku bool

	// sinJuez dice que la carpeta de evals no lleva la carpeta juez de una skill
	// sintética, que el test escribe en ella (escribirLaCarpetaDelJuez): sin
	// ella, la skill no tiene juez. Y sinLista, que no lleva la lista de
	// expresiones prohibidas del repositorio.
	sinJuez, sinLista bool

	// respuestasAlteradas son las del modo orden.
	respuestasAlteradas

	// dosModos dice que el plan es el del job, con el modo orden y el modo
	// herramienta, y enHerramienta, las respuestas alteradas de las sesiones del
	// modo herramienta.
	dosModos      bool
	enHerramienta respuestasAlteradas

	// sinTerminar son cuántas evals informativas más lleva, detrás de las otras,
	// cuyas tres sesiones de modeloSonnet55 cortó el tope tras su respuesta, con
	// el código 124, mientras esperaban una tarea en segundo plano.
	sinTerminar int

	// conLaDeNoActivacion dice que lleva, detrás de todas, la eval de no
	// activación del caso aprobado, que decide, con sus sesiones de cada modelo,
	// que no activan la skill y pasan.
	conLaDeNoActivacion bool

	// conLaSinBinarioNiServidor dice que lleva, detrás de todas, una eval sin
	// binario ni servidor, que decide, con sus sesiones de cada modelo, una sola
	// vez y sin modo, que responden con la línea ⚠ SIN CONSULTA AL BOE: y pasan;
	// y sinBinarioAlteradas, cuántas de las de modeloSonnet55, las primeras,
	// llevan además delante la transición de la memoria de consultas, con
	// expresiones de la lista, y no activan la skill, que es por lo que no pasan.
	conLaSinBinarioNiServidor bool
	sinBinarioAlteradas       int

	// sinMedir es el modelo cuyas sesiones copiadas de la del art. 21 terminan
	// todas con el mensaje del límite de uso; vacío, ninguno.
	sinMedir string

	// anadidoALaEval es lo que se añade al final de cada eval, y
	// prefijoDeLasRespuestas, lo que se antepone a la respuesta de cada sesión;
	// vacíos, nada.
	anadidoALaEval, prefijoDeLasRespuestas string

	// duracion, duracionEnHerramienta y duracionSinModo son los segundos de las
	// tres tandas del repartidor —la del modo orden, la del modo herramienta y la
	// de las evals sin binario ni servidor—, y objetivo, ObjetivoDeDuracion.
	duracion, duracionEnHerramienta, duracionSinModo, objetivo int
}

// respuestasAlteradas son las respuestas de un modo de una ejecucionConUmbrales
// que no son la copia que pasa.
type respuestasAlteradas struct {
	// conExpresiones son cuántas respuestas de modeloSonnet55 llevan delante la
	// transición de la memoria de consultas y redaccionNoLeida, con expresiones
	// de la lista del repositorio: la primera sesión de otras tantas de sus
	// series, repartidas por igual entre ellas.
	conExpresiones int

	// sinActivar son cuántas respuestas de modeloSonnet55 no activan la skill
	// —su transcript no la carga—: como las de conExpresiones, la primera sesión
	// de otras tantas de sus series, de modo que cada serie que decide sigue
	// llegando al umbral con las que pasan, y con una de cada es la misma sesión.
	sinActivar int
}

// modos son los modos del plan de la ejecución: los dos del job con dosModos y,
// si no, solo el modo orden.
func (e ejecucionConUmbrales) modos() []Modo {
	if e.dosModos {
		return []Modo{ModoOrden, ModoHerramienta}
	}

	return []Modo{ModoOrden}
}

// alteradas son las respuestas alteradas de la ejecución en el modo dado.
func (e ejecucionConUmbrales) alteradas(modo Modo) respuestasAlteradas {
	if modo == ModoHerramienta {
		return e.enHerramienta
	}

	return e.respuestasAlteradas
}

// duracionDeLasTandas es la suma de los segundos de las tres tandas de la
// ejecución: lo que el informe publica en duracion_de_las_sesiones.
func (e ejecucionConUmbrales) duracionDeLasTandas() int {
	return e.duracion + e.duracionEnHerramienta + e.duracionSinModo
}

// repeticionesConUmbrales y umbralConUmbrales son las repeticiones y el umbral de
// las ejecuciones de ejecucionConUmbrales: los del job (ADR 0016).
const (
	repeticionesConUmbrales = 3
	umbralConUmbrales       = 2
)

// TestUmbralesDelInforme fija los umbrales del informe y lo que hacen con el
// veredicto (contrato informe-del-job §1, §2 y §6 de H7.3; contrato de umbrales
// del ADR 0029; data-model §1; research D5 y D14; FR-001 a FR-006, FR-050,
// FR-051, FR-092; SC-006; US2-1 a US2-5, US3-5), con ejecuciones sintéticas
// escritas con EscribirInforme en t.TempDir(): el de las respuestas del modelo
// que decide sin la skill activada, con sus respuestas medidas de las evals que
// activan la skill —las informativas incluidas— como total; y el de la duración
// si hay objetivo, sin total y decidiendo. Un umbral que decide y no se cumple
// pone el veredicto en fallo con su motivo, detrás de los de siempre y delante
// del de las sesiones sin medir, y el de la duración, con el prefijo de la
// ejecución, detrás de él; uno que se cumple no cambia nada. Los valores
// esperados son los del contrato escritos a mano —0, no la constante del
// paquete—, para que un umbral cambiado en el código no pase. En todos, rehacer
// la comparación de cada umbral da su cumple
// (exigirLosInvariantesDeLosUmbrales, desde leerInformeEscrito).
//
// Desde H7.4 (contracts/informe-del-job.md §1, §2, §3 y §6; data-model §5 y §6;
// research D13 y D14; FR-041, FR-045, FR-048, FR-061, FR-097; SC-006; US4-1 a
// US4-3, US5-2), sobre una ejecución como la del cierre —18 evals que activan la
// skill, 8 de ellas informativas: 54 respuestas del modelo que decide—, cuenta
// como respuesta medida solo la sesión terminada: seis más que el tope cortó
// tras su respuesta se publican con su motivo y no cuentan en el total; y una
// sesión de una eval que no activa la skill no cuenta como sin activar.
//
// Desde H21 (contracts/evals-en-dos-modos.md §5.1, §5.2 y §8; data-model §10;
// research D19; FR-043, FR-044, FR-047, FR-080; SC-011; US5-1, US5-2, US5-5),
// cada umbral se mide sobre las sesiones de un solo modo y lo nombra: un plan de
// un modo, el de los casos de antes, da los suyos del modo orden, y el del job,
// los de los dos. Cada uno incumplido en un solo modo pone el veredicto en fallo
// con el motivo que lo nombra con su modo; 901 s en la tanda de un modo y 900 en
// la otra, también; y las sesiones de la eval sin binario ni servidor no entran
// en ninguna medida, en ningún total ni en ninguna duración, aunque
// duracion_de_las_sesiones sume las tres tandas.
//
// Desde H24 (contracts/informe-del-job.md §2 y §6 de H24; research D13 y D14;
// FR-034, FR-037, FR-070), umbrales no lleva ningún elemento de las expresiones
// de la lista —ni expresiones_prohibidas ni redaccion_no_leida, de ningún modelo
// ni modo—, y las respuestas que las llevan pasan y no cuentan en nada; el de las
// respuestas sin activar existe si la skill tiene juez, la carpeta juez que la
// ejecución sintética escribe junto a sus evals, y no si tiene lista: con juez y
// sin lista está, y con lista y sin juez, no. Con el plan del job, el juez y el
// objetivo son cuatro, y todos deciden: cada uno incumplido sigue dando fallo.
func TestUmbralesDelInforme(t *testing.T) {
	t.Parallel()

	// comoElJob es la ejecución del cierre de boe-legislacion desde H7.4: 18
	// evals que activan la skill, 8 de ellas informativas, con Sonnet 5.5 y Haiku
	// 4.5 (54 respuestas del que decide); minima, con una de cada clase (6). Las
	// dos, en un solo modo; enDosModos y minimaEnDosModos, las mismas con el plan
	// del job desde H21: 54 respuestas, o 6, en cada modo.
	comoElJob := ejecucionConUmbrales{queDeciden: 10, informativas: 8, conHaiku: true}
	minima := ejecucionConUmbrales{queDeciden: 1, informativas: 1, conHaiku: true}
	enDosModos := conCambios(comoElJob, func(e *ejecucionConUmbrales) { e.dosModos = true })
	minimaEnDosModos := conCambios(minima, func(e *ejecucionConUmbrales) { e.dosModos = true })

	casos := []struct {
		nombre    string
		ejecucion ejecucionConUmbrales
		umbrales  []Umbral
		motivos   []string
		veredicto Veredicto

		// exigir, si no es nil, exige lo que el caso fija además de sus umbrales,
		// sus motivos y su veredicto.
		exigir func(t *testing.T, leido informeLeido)
	}{
		{
			nombre:    "sin-activar-una",
			ejecucion: conCambios(comoElJob, func(e *ejecucionConUmbrales) { e.sinActivar = 1 }),
			umbrales:  []Umbral{umbralSinActivar(modeloSonnet55, ModoOrden, 1, 54, false)},
			motivos:   []string{"umbral sin_activar:claude-sonnet-5-5:orden: 1 de 54 (1,9 %), y tiene que ser ≤ 0,0 %"},
			veredicto: VeredictoFallo,
		},
		{
			// Las sesiones de la eval de no activación, que no activan la skill como
			// ella espera, no cuentan: ni como sin activar ni en el total.
			nombre:    "sin-activar-ninguna",
			ejecucion: conCambios(comoElJob, func(e *ejecucionConUmbrales) { e.conLaDeNoActivacion = true }),
			umbrales:  []Umbral{umbralSinActivar(modeloSonnet55, ModoOrden, 0, 54, true)},
			veredicto: VeredictoAprobado,
			exigir:    exigirLaDeNoActivacionSinContar,
		},
		{
			// Tres respuestas con expresiones de la lista, de 54: hasta H24, un 5,6 %
			// que no cumplía su umbral. La lista está en la carpeta y no juzga nada.
			nombre:    "tres-con-expresiones-de-la-lista",
			ejecucion: conCambios(comoElJob, func(e *ejecucionConUmbrales) { e.conExpresiones = 3 }),
			umbrales:  []Umbral{umbralSinActivar(modeloSonnet55, ModoOrden, 0, 54, true)},
			veredicto: VeredictoAprobado,
			exigir:    exigirLasTresConExpresionesQuePasan,
		},
		{
			// Con las seis sin terminar en el total, sería 1 de 60 (1,7 %).
			nombre: "una-sin-activar-y-seis-sin-terminar",
			ejecucion: conCambios(comoElJob, func(e *ejecucionConUmbrales) {
				e.sinActivar, e.sinTerminar = 1, 2
			}),
			umbrales:  []Umbral{umbralSinActivar(modeloSonnet55, ModoOrden, 1, 54, false)},
			motivos:   []string{"umbral sin_activar:claude-sonnet-5-5:orden: 1 de 54 (1,9 %), y tiene que ser ≤ 0,0 %"},
			veredicto: VeredictoFallo,
			exigir:    exigirLasSeisSinTerminar,
		},
		{
			nombre: "una-sin-activar-con-expresiones",
			ejecucion: conCambios(comoElJob, func(e *ejecucionConUmbrales) {
				e.sinActivar, e.conExpresiones = 1, 1
			}),
			umbrales:  []Umbral{umbralSinActivar(modeloSonnet55, ModoOrden, 1, 54, false)},
			motivos:   []string{"umbral sin_activar:claude-sonnet-5-5:orden: 1 de 54 (1,9 %), y tiene que ser ≤ 0,0 %"},
			veredicto: VeredictoFallo,
			exigir:    exigirLaMismaSesionSinActivarConExpresiones,
		},
		{
			// Sin juez, ni la respuesta sin activar es un umbral, aunque la carpeta
			// tenga lista; y sin objetivo, tampoco ninguna duración: como en
			// legal-core (FR-037 de H24).
			nombre: "sin-juez-ni-objetivo",
			ejecucion: conCambios(minima, func(e *ejecucionConUmbrales) {
				e.sinJuez, e.conExpresiones, e.sinActivar, e.duracion = true, 1, 1, 2000
			}),
			umbrales:  []Umbral{},
			veredicto: VeredictoAprobado,
		},
		{
			// El objetivo de duración no depende del juez: sigue como hasta H24.
			nombre: "sin-juez-y-con-objetivo",
			ejecucion: conCambios(minima, func(e *ejecucionConUmbrales) {
				e.sinJuez, e.sinActivar, e.duracion, e.objetivo = true, 1, 900, 900
			}),
			umbrales:  []Umbral{umbralDeDuracion(ModoOrden, 900, 900, true)},
			veredicto: VeredictoAprobado,
		},
		{
			// Con juez y sin lista, el de las respuestas sin activar está: no
			// depende de la lista (research D13 de H24).
			nombre:    "con-juez-y-sin-lista",
			ejecucion: conCambios(minima, func(e *ejecucionConUmbrales) { e.sinLista, e.sinActivar = true, 1 }),
			umbrales:  []Umbral{umbralSinActivar(modeloSonnet55, ModoOrden, 1, 6, false)},
			motivos:   []string{"umbral sin_activar:claude-sonnet-5-5:orden: 1 de 6 (16,7 %), y tiene que ser ≤ 0,0 %"},
			veredicto: VeredictoFallo,
		},
		{
			nombre:    "901-s-con-objetivo-900",
			ejecucion: conCambios(minima, func(e *ejecucionConUmbrales) { e.duracion, e.objetivo = 901, 900 }),
			umbrales: []Umbral{
				umbralSinActivar(modeloSonnet55, ModoOrden, 0, 6, true),
				umbralDeDuracion(ModoOrden, 901, 900, false),
			},
			motivos:   []string{"de la ejecución, no de la skill: duracion_de_las_sesiones:orden: 901 s, y tiene que ser ≤ 900 s"},
			veredicto: VeredictoFallo,
		},
		{
			nombre:    "900-s-con-objetivo-900",
			ejecucion: conCambios(minima, func(e *ejecucionConUmbrales) { e.duracion, e.objetivo = 900, 900 }),
			umbrales: []Umbral{
				umbralSinActivar(modeloSonnet55, ModoOrden, 0, 6, true),
				umbralDeDuracion(ModoOrden, 900, 900, true),
			},
			veredicto: VeredictoAprobado,
		},
		{
			// Sin ninguna respuesta medida, la proporción es 0 y el umbral se
			// cumple; el veredicto ya es fallo por las sesiones sin medir, y sus
			// series, sin medir, no dan el motivo de su tasa.
			nombre:    "todas-las-de-sonnet-5-5-sin-medir",
			ejecucion: conCambios(minima, func(e *ejecucionConUmbrales) { e.sinMedir = modeloSonnet55 }),
			umbrales:  []Umbral{umbralSinActivar(modeloSonnet55, ModoOrden, 0, 0, true)},
			motivos:   []string{motivoEsperadoDelLimite(sesionesSinMedirDe(modeloSonnet55, 2)...)},
			veredicto: VeredictoFallo,
		},
		{
			// Los tres motivos de H7.3 a la vez, en su orden: el del umbral que
			// decide, el de las sesiones sin medir —las de Haiku 4.5— y el de la
			// duración.
			nombre: "los-tres-motivos-en-su-orden",
			ejecucion: conCambios(minima, func(e *ejecucionConUmbrales) {
				e.sinActivar, e.sinMedir, e.duracion, e.objetivo = 1, modeloHaiku45, 901, 900
			}),
			umbrales: []Umbral{
				umbralSinActivar(modeloSonnet55, ModoOrden, 1, 6, false),
				umbralDeDuracion(ModoOrden, 901, 900, false),
			},
			motivos: []string{
				"umbral sin_activar:claude-sonnet-5-5:orden: 1 de 6 (16,7 %), y tiene que ser ≤ 0,0 %",
				motivoEsperadoDelLimite(sesionesSinMedirDe(modeloHaiku45, 1)...),
				"de la ejecución, no de la skill: duracion_de_las_sesiones:orden: 901 s, y tiene que ser ≤ 900 s",
			},
			veredicto: VeredictoFallo,
		},
		{
			// Las medidas de un modo no se suman a las del otro (FR-043 de H21).
			nombre:    "sin-activar-solo-en-orden",
			ejecucion: conCambios(enDosModos, func(e *ejecucionConUmbrales) { e.sinActivar = 1 }),
			umbrales: []Umbral{
				umbralSinActivar(modeloSonnet55, ModoOrden, 1, 54, false),
				umbralSinActivar(modeloSonnet55, ModoHerramienta, 0, 54, true),
			},
			motivos:   []string{"umbral sin_activar:claude-sonnet-5-5:orden: 1 de 54 (1,9 %), y tiene que ser ≤ 0,0 %"},
			veredicto: VeredictoFallo,
		},
		{
			nombre:    "sin-activar-solo-en-herramienta",
			ejecucion: conCambios(enDosModos, func(e *ejecucionConUmbrales) { e.enHerramienta.sinActivar = 1 }),
			umbrales: []Umbral{
				umbralSinActivar(modeloSonnet55, ModoOrden, 0, 54, true),
				umbralSinActivar(modeloSonnet55, ModoHerramienta, 1, 54, false),
			},
			motivos: []string{
				"umbral sin_activar:claude-sonnet-5-5:herramienta: 1 de 54 (1,9 %), y tiene que ser ≤ 0,0 %",
			},
			veredicto: VeredictoFallo,
		},
		{
			// Las dos tandas suman 1801 s, y cada umbral mide solo la suya.
			nombre: "901-s-en-herramienta-y-900-en-orden",
			ejecucion: conCambios(minimaEnDosModos, func(e *ejecucionConUmbrales) {
				e.duracion, e.duracionEnHerramienta, e.objetivo = 900, 901, 900
			}),
			umbrales: []Umbral{
				umbralSinActivar(modeloSonnet55, ModoOrden, 0, 6, true),
				umbralSinActivar(modeloSonnet55, ModoHerramienta, 0, 6, true),
				umbralDeDuracion(ModoOrden, 900, 900, true),
				umbralDeDuracion(ModoHerramienta, 901, 900, false),
			},
			motivos: []string{
				"de la ejecución, no de la skill: duracion_de_las_sesiones:herramienta: 901 s, y tiene que ser ≤ 900 s",
			},
			veredicto: VeredictoFallo,
		},
		{
			nombre: "901-s-en-orden-y-900-en-herramienta",
			ejecucion: conCambios(minimaEnDosModos, func(e *ejecucionConUmbrales) {
				e.duracion, e.duracionEnHerramienta, e.objetivo = 901, 900, 900
			}),
			umbrales: []Umbral{
				umbralSinActivar(modeloSonnet55, ModoOrden, 0, 6, true),
				umbralSinActivar(modeloSonnet55, ModoHerramienta, 0, 6, true),
				umbralDeDuracion(ModoOrden, 901, 900, false),
				umbralDeDuracion(ModoHerramienta, 900, 900, true),
			},
			motivos: []string{
				"de la ejecución, no de la skill: duracion_de_las_sesiones:orden: 901 s, y tiene que ser ≤ 900 s",
			},
			veredicto: VeredictoFallo,
		},
		{
			// Cada tanda de un modo, en su objetivo, aunque las dos sumen el doble.
			nombre: "900-s-en-los-dos-modos",
			ejecucion: conCambios(minimaEnDosModos, func(e *ejecucionConUmbrales) {
				e.duracion, e.duracionEnHerramienta, e.objetivo = 900, 900, 900
			}),
			umbrales: []Umbral{
				umbralSinActivar(modeloSonnet55, ModoOrden, 0, 6, true),
				umbralSinActivar(modeloSonnet55, ModoHerramienta, 0, 6, true),
				umbralDeDuracion(ModoOrden, 900, 900, true),
				umbralDeDuracion(ModoHerramienta, 900, 900, true),
			},
			veredicto: VeredictoAprobado,
		},
		{
			// Las sesiones de la eval sin binario ni servidor no son de ningún modo:
			// ni la que no activa la skill cuenta en una medida, ni las seis en un
			// total, ni los 5000 s de su tanda en una duración (FR-047 de H21).
			nombre: "la-sin-binario-ni-servidor-fuera-de-toda-medida",
			ejecucion: conCambios(enDosModos, func(e *ejecucionConUmbrales) {
				e.conLaSinBinarioNiServidor, e.sinBinarioAlteradas = true, 1
				e.duracion, e.duracionEnHerramienta, e.duracionSinModo, e.objetivo = 500, 600, 5000, 900
			}),
			umbrales: []Umbral{
				umbralSinActivar(modeloSonnet55, ModoOrden, 0, 54, true),
				umbralSinActivar(modeloSonnet55, ModoHerramienta, 0, 54, true),
				umbralDeDuracion(ModoOrden, 500, 900, true),
				umbralDeDuracion(ModoHerramienta, 600, 900, true),
			},
			veredicto: VeredictoAprobado,
			exigir:    exigirLaSinBinarioNiServidorSinContar,
		},
	}

	for _, caso := range casos {
		t.Run(caso.nombre, func(t *testing.T) {
			t.Parallel()

			leido := escribirEjecucionConUmbrales(t, caso.ejecucion)

			exigirUmbrales(t, leido, caso.umbrales)
			exigirMotivosDeLaRaiz(t, leido, caso.motivos...)
			assert.Equal(t, caso.veredicto, leido.informe.Veredicto)
			assert.Equal(t, caso.ejecucion.duracionDeLasTandas(), leido.informe.DuracionDeLasSesiones,
				"duracion_de_las_sesiones es la suma de las tres tandas")

			if caso.exigir != nil {
				caso.exigir(t, leido)
			}
		})
	}

	t.Run("objetivo-negativo", func(t *testing.T) {
		t.Parallel()

		// Un objetivo negativo impide escribir el informe, como un umbral de
		// series fuera de rango: ni informe.md ni informe.json.
		entradas := entradasDelCaso(casoAprobado, t.TempDir())
		entradas.ObjetivoDeDuracion = -1

		informe, err := EscribirInforme(entradas)
		require.Error(t, err)
		require.ErrorContains(t, err, "el informe no se puede escribir")
		require.ErrorContains(t, err, "-1")
		assert.Zero(t, informe, "sin informe, EscribirInforme no devuelve ningún veredicto")
		assert.NoFileExists(t, filepath.Join(entradas.Destino, "informe.md"))
		assert.NoFileExists(t, filepath.Join(entradas.Destino, "informe.json"))
	})
}

// TestUmbralQueSoloSePublica fija lo que el informe hace con un umbral que no
// decide (contrato de umbrales del ADR 0029; contrato informe-del-job §2 y §4 de
// H7.3; FR-003), con umbrales escritos aquí: desde H24, el plan del job no da
// ninguno hasta que el juez publique una clase. Sin cumplirse, no da ningún
// motivo, ni de la skill ni de la ejecución, y su fila de informe.md dice «no:
// solo se publica» donde la del que decide dice «sí».
func TestUmbralQueSoloSePublica(t *testing.T) {
	t.Parallel()

	total := 30
	publicado := compararUmbral(Umbral{
		Nombre: "se_publica:claude-haiku-4-5-20251001:orden", Medida: 2, Total: &total, Umbral: 0.05,
	})
	queDecide := umbralSinActivar(modeloSonnet55, ModoOrden, 1, 54, false)

	require.False(t, publicado.Cumple, "2 de 30 no cumple el 5 %%")

	umbrales := []Umbral{queDecide, publicado}

	assert.Equal(t, []string{"umbral sin_activar:claude-sonnet-5-5:orden: 1 de 54 (1,9 %), y tiene que ser ≤ 0,0 %"},
		motivosDeLosUmbrales(umbrales))
	assert.Empty(t, motivosDeLaDuracionDeLasSesiones(umbrales))
	assert.Equal(t, [][]string{
		{"`sin_activar:claude-sonnet-5-5:orden`", "1 de 54 (1,9 %)", "≤ 0,0 %", "no", "sí"},
		{"`se_publica:claude-haiku-4-5-20251001:orden`", "2 de 30 (6,7 %)", "≤ 5,0 %", "no", "no: solo se publica"},
	}, filasDeUmbrales(umbrales))
}

// conCambios es la ejecución dada con lo que cambie cambiar.
func conCambios(ejecucion ejecucionConUmbrales, cambiar func(e *ejecucionConUmbrales)) ejecucionConUmbrales {
	cambiar(&ejecucion)

	return ejecucion
}

// umbralSinActivar es el umbral de las respuestas del modelo que decide sin la
// skill activada en un modo tal como lo fijan contracts/informe-del-job.md §2 de
// H7.4 y de H24 y contracts/evals-en-dos-modos.md §5.1 de H21, con sus valores
// escritos a mano: 0 y decidiendo.
func umbralSinActivar(modelo string, modo Modo, sinActivar, respuestas int, cumple bool) Umbral {
	return Umbral{
		Nombre: "sin_activar:" + modelo + ":" + string(modo),
		Descripcion: "Respuestas de " + modelo + " en el modo " + string(modo) + " sin la skill activada, sobre sus " +
			"respuestas medidas en las evals que la activan",
		Medida:      float64(sinActivar),
		Total:       &respuestas,
		Comparacion: "<=",
		Umbral:      0,
		Cumple:      cumple,
		Decide:      true,
	}
}

// exigirLaSinBinarioNiServidorSinContar exige, del caso
// la-sin-binario-ni-servidor-fuera-de-toda-medida, que el caso no pase en vacío
// (FR-047): la eval sin binario ni servidor tiene sus sesiones de los dos
// modelos, una sola vez y sin modo; la primera del modelo que decide no activa
// la skill, y no pasa; y su serie, con las otras dos, llega al umbral. Ningún
// umbral cuenta ninguna de las seis: los del caso tienen 0 de 54.
func exigirLaSinBinarioNiServidorSinContar(t *testing.T, leido informeLeido) {
	t.Helper()

	sinBinario := ficheroSintetico(19)

	var sesiones []string

	for _, resultado := range leido.informe.Evals {
		if resultado.Eval != sinBinario {
			continue
		}

		sesiones = append(sesiones, resultado.Sesion)

		assert.Empty(t, resultado.Modo, "%s no es de ningún modo", resultado.Sesion)
		assert.True(t, resultado.Activa, "%s es de una eval que activa la skill", resultado.Sesion)
	}

	alterada := sesionSintetica(19, modeloSonnet55, 1)

	var esperadas []string

	for _, modelo := range []string{modeloHaiku45, modeloSonnet55} {
		for vez := 1; vez <= repeticionesConUmbrales; vez++ {
			esperadas = append(esperadas, sesionSintetica(19, modelo, vez))
		}
	}

	assert.Equal(t, esperadas, sesiones, "la eval sin binario ni servidor tiene sus sesiones de los dos modelos, una vez")

	resultado := resultadoDeLaSesion(t, leido.informe, alterada)
	assert.False(t, resultado.Activada, "%s no activa la skill", alterada)
	assert.True(t, resultado.LineaSinConsulta, "%s lleva la línea con su dirección", alterada)
	assert.False(t, resultado.Pasa, "%s no pasa", alterada)

	for _, modelo := range []string{modeloSonnet55, modeloHaiku45} {
		assert.Equal(t, TasaDelInforme{
			Eval: sinBinario, Modelo: modelo, Planificada: true, Decide: modelo == modeloSonnet55,
			Formas: sinFormasExigidas, Sesiones: 3, Pasan: map[string]int{modeloSonnet55: 2, modeloHaiku45: 3}[modelo],
			Pasa: true,
		}, tasaDeLaSerie(t, leido.informe, sinBinario, modelo))
	}
}

// exigirLaDeNoActivacionSinContar exige, del caso sin-activar-ninguna, que el
// caso no pase en vacío: las sesiones de la eval de no activación, de los dos
// modelos, no activan la skill y pasan. El umbral del caso no las cuenta: tiene
// 0 de 54.
func exigirLaDeNoActivacionSinContar(t *testing.T, leido informeLeido) {
	t.Helper()

	deNoActivacion := ficheroSintetico(19)

	var sesiones int

	for _, resultado := range leido.informe.Evals {
		if resultado.Eval != deNoActivacion {
			continue
		}

		sesiones++

		assert.False(t, resultado.Activa, "%s es de una eval que no activa la skill", resultado.Sesion)
		assert.False(t, resultado.Activada, "%s no activa la skill", resultado.Sesion)
		assert.True(t, resultado.Pasa, "%s pasa", resultado.Sesion)
	}

	assert.Equal(t, 2*repeticionesConUmbrales, sesiones, "la eval de no activación tiene sus sesiones de los dos modelos")
}

// exigirLasTresConExpresionesQuePasan exige, del caso
// tres-con-expresiones-de-la-lista, que el caso no pase en vacío: tres sesiones
// del modelo que decide, y solo ellas, llevan en su respuesta lo que la lista
// del repositorio marca, y las tres pasan, sin ningún motivo.
func exigirLasTresConExpresionesQuePasan(t *testing.T, leido informeLeido) {
	t.Helper()

	conjunto, err := LeerConjunto(evalsDelRepositorio)
	require.NoError(t, err)

	lista := listaDelRepositorio(t, conjunto)

	var conExpresiones []string

	for _, resultado := range leido.informe.Evals {
		if len(ExtraerExpresionesProhibidas(resultado.Respuesta, lista)) == 0 {
			continue
		}

		conExpresiones = append(conExpresiones, resultado.Sesion)

		assert.True(t, resultado.Pasa, "%s pasa con expresiones de la lista en su respuesta", resultado.Sesion)
		assert.Empty(t, resultado.Motivos, "%s no tiene ningún motivo", resultado.Sesion)
	}

	assert.Equal(t, []string{
		sesionSintetica(1, modeloSonnet55, 1), sesionSintetica(7, modeloSonnet55, 1), sesionSintetica(13, modeloSonnet55, 1),
	}, conExpresiones, "las tres respuestas con expresiones, repartidas entre las series del modelo que decide")
}

// exigirLasSeisSinTerminar exige, del caso una-sin-activar-y-seis-sin-terminar,
// que las seis sesiones que el tope cortó se publiquen con su respuesta y su
// motivo, sin pasar (FR-045, FR-061). El umbral del caso no las cuenta en su
// total: tiene 54, y no 60.
func exigirLasSeisSinTerminar(t *testing.T, leido informeLeido) {
	t.Helper()

	var cortadas []string

	for numero := 19; numero <= 20; numero++ {
		for vez := 1; vez <= repeticionesConUmbrales; vez++ {
			sesion := sesionSintetica(numero, modeloSonnet55, vez)
			cortadas = append(cortadas, sesion)

			resultado := resultadoDeLaSesion(t, leido.informe, sesion)
			assert.False(t, resultado.SesionTerminada, "%s no terminó", sesion)
			assert.False(t, resultado.Pasa, "%s no pasa", sesion)
			assert.NotEmpty(t, resultado.Respuesta, "%s tiene la respuesta a la pregunta", sesion)
			assert.Equal(t, []string{motivoDelTopeEsperado}, resultado.Motivos, "el motivo de %s", sesion)
		}
	}

	var sinTerminar []string

	for _, resultado := range leido.informe.Evals {
		if !resultado.SesionTerminada {
			sinTerminar = append(sinTerminar, resultado.Sesion)
		}
	}

	assert.Equal(t, cortadas, sinTerminar, "las seis son las únicas sin terminar")
}

// exigirLaMismaSesionSinActivarConExpresiones exige, del caso
// una-sin-activar-con-expresiones, que la sesión sin la skill activada y la de
// las expresiones de la lista sean la misma, la primera de la primera eval con
// el modelo que decide, y que su único motivo sea el de la activación: por las
// expresiones no tiene ninguno.
func exigirLaMismaSesionSinActivarConExpresiones(t *testing.T, leido informeLeido) {
	t.Helper()

	var sinActivar, conExpresiones []string

	for _, resultado := range leido.informe.Evals {
		if resultado.Activa && !resultado.Activada {
			sinActivar = append(sinActivar, resultado.Sesion)
		}

		if strings.Contains(resultado.Respuesta, redaccionNoLeida) {
			conExpresiones = append(conExpresiones, resultado.Sesion)
		}
	}

	misma := sesionSintetica(1, modeloSonnet55, 1)
	assert.Equal(t, []string{misma}, sinActivar)
	assert.Equal(t, []string{misma}, conExpresiones)
	assert.Equal(t,
		[]string{"la activación no coincide: se esperaba que la skill boe-legislacion se activara y no se activó"},
		resultadoDeLaSesion(t, leido.informe, misma).Motivos)
}

// umbralDeDuracion es el umbral de la duración de las sesiones de un modo, la de
// su tanda, tal como lo fijan el contrato informe-del-job §1 de H7.3 y
// contracts/evals-en-dos-modos.md §5.1 de H21: sin total y decidiendo.
func umbralDeDuracion(modo Modo, segundos, objetivo int, cumple bool) Umbral {
	return Umbral{
		Nombre: "duracion_de_las_sesiones:" + string(modo),
		Descripcion: "Segundos de la tanda del modo " + string(modo) + ", desde que se prepara su primera sesión " +
			"hasta que termina la última",
		Medida:      float64(segundos),
		Comparacion: "<=",
		Umbral:      float64(objetivo),
		Cumple:      cumple,
		Decide:      true,
	}
}

// exigirUmbrales exige los umbrales esperados, en su orden: los del Informe y los
// escritos en informe.json, con sus claves en el orden del contrato del ADR 0029,
// total solo si lo llevan y una lista vacía, no null, si no hay ninguno.
func exigirUmbrales(t *testing.T, leido informeLeido, esperados []Umbral) {
	t.Helper()

	if len(esperados) == 0 {
		assert.Empty(t, leido.informe.Umbrales)
	} else {
		assert.Equal(t, esperados, leido.informe.Umbrales)
	}

	escritos := make([]string, 0, len(esperados))
	for _, umbral := range esperados {
		escritos = append(escritos, umbralEscrito(t, umbral))
	}

	assert.Equal(t, "["+strings.Join(escritos, ",")+"]", compacto(t, leido.crudo.Umbrales))
}

// umbralEscrito es el umbral tal como lo escribe informe.json, sin blancos: las
// claves del contrato del ADR 0029, en su orden, y total solo si lo lleva.
func umbralEscrito(t *testing.T, umbral Umbral) string {
	t.Helper()

	total := ""
	if umbral.Total != nil {
		total = `"total":` + strconv.Itoa(*umbral.Total) + ","
	}

	return fmt.Sprintf(`{"nombre":%s,"descripcion":%s,"medida":%s,%s"comparacion":%s,"umbral":%s,"cumple":%t,"decide":%t}`,
		cadenaJSON(t, umbral.Nombre), cadenaJSON(t, umbral.Descripcion), strconv.FormatFloat(umbral.Medida, 'f', -1, 64),
		total, cadenaJSON(t, umbral.Comparacion), strconv.FormatFloat(umbral.Umbral, 'f', -1, 64), umbral.Cumple,
		umbral.Decide)
}

// exigirLosInvariantesDeLosUmbrales exige, en cualquier informe, lo que el
// informe final comprueba sin modelo de sus umbrales (contrato de umbrales del
// ADR 0029; FR-001; US2-4): cada nombre con su forma y sin repetir; rehacer la
// comparación de cada uno con su medida, su total, su comparación y su umbral da
// su cumple; y uno que decide y no se cumple implica el veredicto fallo.
func exigirLosInvariantesDeLosUmbrales(t *testing.T, leido informeLeido) {
	t.Helper()

	var nombres []string

	for _, umbral := range leido.informe.Umbrales {
		assert.Regexp(t, formaDelNombreDeUmbral, umbral.Nombre)
		assert.NotContains(t, nombres, umbral.Nombre, "el nombre del umbral es único en el informe")
		nombres = append(nombres, umbral.Nombre)

		assert.Equal(t, rehacerLaComparacion(t, umbral), umbral.Cumple, "cumple de %s es la comparación rehecha",
			umbral.Nombre)

		if umbral.Decide && !umbral.Cumple {
			assert.Equal(t, VeredictoFallo, leido.informe.Veredicto, "%s decide y no se cumple", umbral.Nombre)
		}
	}
}

// rehacerLaComparacion es la comparación del umbral rehecha como la rehace el
// informe final: la proporción medida/total —0 si total es 0— si lleva total y,
// si no, la medida, comparada con el umbral en coma flotante de doble precisión y
// sin redondeos.
func rehacerLaComparacion(t *testing.T, umbral Umbral) bool {
	t.Helper()

	valor := umbral.Medida
	if umbral.Total != nil {
		valor = 0
		if *umbral.Total != 0 {
			valor = umbral.Medida / float64(*umbral.Total)
		}
	}

	switch umbral.Comparacion {
	case "<=":
		return valor <= umbral.Umbral
	case "<":
		return valor < umbral.Umbral
	case ">=":
		return valor >= umbral.Umbral
	case ">":
		return valor > umbral.Umbral
	}

	require.Failf(t, "comparación desconocida", "el umbral %s compara con %q", umbral.Nombre, umbral.Comparacion)

	return false
}

// escribirEjecucionConUmbrales arma la ejecución en un directorio temporal del
// test, escribe su informe con EscribirInforme y lo lee con leerInformeEscrito.
func escribirEjecucionConUmbrales(t *testing.T, ejecucion ejecucionConUmbrales) informeLeido {
	t.Helper()

	return informeDeLaCopia(t, armarEjecucionConUmbrales(t, ejecucion), ejecucion.ajustar)
}

// modelos son los modelos del job de la ejecución: modeloSonnet55, que decide, y
// detrás, con conHaiku, modeloHaiku45, el único de sus modelos informativos.
func (e ejecucionConUmbrales) modelos() []string {
	if e.conHaiku {
		return []string{modeloSonnet55, modeloHaiku45}
	}

	return []string{modeloSonnet55}
}

// ajustar pone en las entradas de EscribirInforme las de la ejecución: sus
// modelos, las tres repeticiones y el umbral 2 del job, sus modos —ninguno, que
// es solo el modo orden, si no es de dos modos—, la suma de sus tres tandas, la
// de la tanda de cada modo y su objetivo.
func (e ejecucionConUmbrales) ajustar(entradas *InformeAEscribir) {
	entradas.ModeloQueDecide = modeloSonnet55
	entradas.ModelosInformativos = e.modelos()[1:]
	entradas.Repeticiones, entradas.Umbral = repeticionesConUmbrales, umbralConUmbrales

	if e.dosModos {
		entradas.Modos = e.modos()
	}

	entradas.DuracionDeLasSesiones = e.duracionDeLasTandas()
	entradas.DuracionDeLosModos = map[Modo]int{ModoOrden: e.duracion, ModoHerramienta: e.duracionEnHerramienta}
	entradas.ObjetivoDeDuracion = e.objetivo
}

// armarEjecucionConUmbrales arma la ejecución en un directorio temporal del test
// —sus evals, con la carpeta juez de una skill sintética y la lista del
// repositorio si las lleva, y sus sesiones, con las respuestas alteradas de cada
// modo— y devuelve su ruta.
func armarEjecucionConUmbrales(t *testing.T, ejecucion ejecucionConUmbrales) string {
	t.Helper()

	copia := t.TempDir()
	evals := filepath.Join(copia, "evals")
	require.NoError(t, os.Mkdir(evals, 0o750))

	if !ejecucion.sinJuez {
		escribirLaCarpetaDelJuez(t, evals)
	}

	if !ejecucion.sinLista {
		copiarLaListaDelRepositorio(t, evals)
	}

	series := escribirLasEvalsConUmbrales(t, copia, ejecucion)

	for _, modo := range ejecucion.modos() {
		alteradas := ejecucion.alteradas(modo)

		alterarLaPrimeraDeLasSeries(t, copia, modo, series, alteradas.conExpresiones, func(dir string) {
			anteponerALaRespuesta(t, dir, transicionDeLaMemoria+" "+redaccionNoLeida+"\n\n")
		})
		alterarLaPrimeraDeLasSeries(t, copia, modo, series, alteradas.sinActivar, func(dir string) {
			quitarLaActivacion(t, dir)
		})
	}

	return copia
}

// escribirLasEvalsConUmbrales escribe en la copia las evals de la ejecución y
// sus sesiones con cada modelo en cada modo —el que decide abre todas, y los
// informativos, solo las que deciden— y devuelve los números de las evals de las
// series del modelo que decide con respuestas medidas, en su orden, que son las
// mismas en cada modo: sin las sin terminar, la de no activación ni la eval sin
// binario ni servidor.
func escribirLasEvalsConUmbrales(t *testing.T, copia string, ejecucion ejecucionConUmbrales) []int {
	t.Helper()

	evals := filepath.Join(copia, "evals")
	delCasoAprobado := filepath.Join(casosDeInforme, casoAprobado)
	eval := contenidoDeLaSesion(t, filepath.Join(delCasoAprobado, "evals"), ficheroDeLaEval01)

	var series []int

	medidas := ejecucion.queDeciden + ejecucion.informativas

	for numero := 1; numero <= medidas+ejecucion.sinTerminar; numero++ {
		informativa := numero > ejecucion.queDeciden

		contenido := eval + ejecucion.anadidoALaEval
		if informativa {
			contenido += "informativa: true\n"
		}

		escribirEnLaCopia(t, evals, ficheroSintetico(numero), contenido)

		if numero <= medidas {
			series = append(series, numero)
		}

		for _, modelo := range ejecucion.modelos() {
			if informativa && modelo != modeloSonnet55 {
				continue
			}

			escribirLasSesionesDeLaSerie(t, copia, ejecucion, numero, modelo, numero > medidas)
		}
	}

	siguiente := medidas + ejecucion.sinTerminar + 1

	if ejecucion.conLaDeNoActivacion {
		escribirLaEvalDeNoActivacion(t, copia, ejecucion, siguiente)

		siguiente++
	}

	if ejecucion.conLaSinBinarioNiServidor {
		escribirLaEvalSinBinarioNiServidor(t, copia, ejecucion, siguiente)
	}

	return series
}

// escribirLaEvalDeNoActivacion escribe en la copia, con ese número, la eval de no
// activación del caso aprobado y sus sesiones de cada modelo en cada modo de la
// ejecución, que no activan la skill y pasan: las del modo herramienta, con su
// servidor.json y sin ninguna llamada.
func escribirLaEvalDeNoActivacion(t *testing.T, copia string, ejecucion ejecucionConUmbrales, numero int) {
	t.Helper()

	escribirEnLaCopia(t, filepath.Join(copia, "evals"), ficheroSintetico(numero),
		contenidoDeLaSesion(t, filepath.Join(casosDeInforme, casoAprobado, "evals"), ficheroDeNoActivacion))

	for _, modelo := range ejecucion.modelos() {
		for _, modo := range ejecucion.modos() {
			for vez := 1; vez <= repeticionesConUmbrales; vez++ {
				dir := escribirCopiaDeLaSesion(t, copia, sesionDeNoActivacion, sesionSinteticaEn(modo, numero, modelo, vez),
					numero, modelo)

				if modo == ModoHerramienta {
					escribirEnLaCopia(t, dir, "servidor.json", servidorDeLaSesionSintetica)
				}
			}
		}
	}
}

// escribirLaEvalSinBinarioNiServidor escribe en la copia, con ese número, la eval
// sin binario ni servidor y sus sesiones de cada modelo, una sola vez y sin modo,
// que pasan; las sinBinarioAlteradas primeras de modeloSonnet55 llevan además
// delante de su respuesta la transición de la memoria de consultas y no activan
// la skill.
func escribirLaEvalSinBinarioNiServidor(t *testing.T, copia string, ejecucion ejecucionConUmbrales, numero int) {
	t.Helper()

	escribirEnLaCopia(t, filepath.Join(copia, "evals"), ficheroSintetico(numero), evalSinBinarioNiServidor)

	for _, modelo := range ejecucion.modelos() {
		for vez := 1; vez <= repeticionesConUmbrales; vez++ {
			dir := escribirSesionSinBinarioNiServidor(t, copia, numero, modelo, vez)

			if modelo == modeloSonnet55 && vez <= ejecucion.sinBinarioAlteradas {
				anteponerALaRespuesta(t, dir, transicionDeLaMemoria+"\n\n")
				quitarLaActivacion(t, dir)
			}
		}
	}
}

// escribirLasSesionesDeLaSerie escribe las sesiones de la eval de ese número con
// el modelo dado en cada modo de la ejecución, con el prefijo de las respuestas
// de la ejecución delante de la suya y, con cortadas, como las que el tope cortó
// tras su respuesta.
func escribirLasSesionesDeLaSerie(
	t *testing.T, copia string, ejecucion ejecucionConUmbrales, numero int, modelo string, cortadas bool,
) {
	t.Helper()

	for _, modo := range ejecucion.modos() {
		for vez := 1; vez <= repeticionesConUmbrales; vez++ {
			dir := escribirSesionSinteticaEn(t, copia, modo, numero, modelo, vez, ejecucion.sinMedir == modelo)

			if ejecucion.prefijoDeLasRespuestas != "" {
				anteponerALaRespuesta(t, dir, ejecucion.prefijoDeLasRespuestas)
			}

			if cortadas {
				cortarTrasLaRespuesta(t, dir)
			}
		}
	}
}

// alterarLaPrimeraDeLasSeries aplica alterar al directorio de la primera sesión
// del modo dado de cuantas de las series de modeloSonnet55 dadas, repartidas por
// igual entre ellas, de modo que cada serie que decide sigue llegando al umbral
// con las que pasan.
func alterarLaPrimeraDeLasSeries(
	t *testing.T, copia string, modo Modo, series []int, cuantas int, alterar func(dir string),
) {
	t.Helper()

	require.LessOrEqual(t, cuantas, len(series), "hay una serie de %s por respuesta alterada", modeloSonnet55)

	for k := range cuantas {
		alterar(filepath.Join(copia, "sesiones",
			sesionSinteticaEn(modo, series[k*len(series)/cuantas], modeloSonnet55, 1)))
	}
}

// quitarLaActivacion quita del transcript de la sesión del directorio la carga
// de la skill —la llamada a Skill y su resultado—, de modo que la sesión
// responde igual sin activarla.
func quitarLaActivacion(t *testing.T, dir string) {
	t.Helper()

	var sinActivar []string

	lineas := strings.SplitAfter(contenidoDeLaSesion(t, dir, "sesion.jsonl"), "\n")
	for _, linea := range lineas {
		if !strings.Contains(linea, `"name":"Skill"`) && !strings.Contains(linea, `"content":"Launching skill: `) {
			sinActivar = append(sinActivar, linea)
		}
	}

	require.Len(t, sinActivar, len(lineas)-2, "el transcript de %s carga la skill en dos mensajes", dir)
	escribirEnLaCopia(t, dir, "sesion.jsonl", strings.Join(sinActivar, ""))

	sesion, err := LeerSesion(dir)
	require.NoError(t, err)
	require.Empty(t, sesion.SkillsActivadas, "la sesión de %s no activa ninguna skill", dir)
}

// cortarTrasLaRespuesta deja la sesión del directorio como una que el tope cortó
// tras su respuesta a la pregunta, mientras esperaba una tarea en segundo plano
// (research V4 de H7.4): su transcript sigue con la notificación de la tarea, y
// su código es el del tope.
func cortarTrasLaRespuesta(t *testing.T, dir string) {
	t.Helper()

	escribirEnLaCopia(t, dir, "sesion.jsonl", contenidoDeLaSesion(t, dir, "sesion.jsonl")+notificacionDeUnaTarea)
	escribirEnLaCopia(t, dir, "codigo-de-la-sesion", "124\n")
}

// escribirSesionSintetica escribe la sesión del modo orden de esa repetición de
// la eval de ese número con el modelo dado (escribirSesionSinteticaEn).
func escribirSesionSintetica(t *testing.T, copia string, numero int, modelo string, vez int, sinMedir bool) {
	t.Helper()

	escribirSesionSinteticaEn(t, copia, ModoOrden, numero, modelo, vez, sinMedir)
}

// escribirSesionSinteticaEn escribe la sesión de esa repetición de la eval de ese
// número con el modelo dado en el modo dado, y devuelve su directorio: una copia
// de la del art. 21 del caso aprobado (escribirCopiaDeLaSesion), pasada al modo
// herramienta si es de ese modo (pasarAlModoHerramienta) y, si sinMedir,
// terminada con el mensaje del límite de uso.
func escribirSesionSinteticaEn(
	t *testing.T, copia string, modo Modo, numero int, modelo string, vez int, sinMedir bool,
) string {
	t.Helper()

	dir := escribirCopiaDeLaSesion(t, copia, sesionDelArticulo21, sesionSinteticaEn(modo, numero, modelo, vez), numero,
		modelo)

	if modo == ModoHerramienta {
		pasarAlModoHerramienta(t, dir)
	}

	if sinMedir {
		cambiarElFinalDeLaSesion(t, dir, 1, mensajeResultConError(t, textoDelLimiteDeSesion))
	}

	return dir
}

// escribirCopiaDeLaSesion escribe, con el nombre dado, una sesión de la eval de
// ese número con el modelo dado, y devuelve su directorio: una copia de la
// sesión dada del caso aprobado que lo pide y lo declara, con eval.txt nombrando
// su eval.
func escribirCopiaDeLaSesion(t *testing.T, copia, delCasoAprobado, sesion string, numero int, modelo string) string {
	t.Helper()

	dir := filepath.Join(copia, "sesiones", sesion)
	copiarSesionConOtroModelo(t, filepath.Join(casosDeInforme, casoAprobado, "sesiones", delCasoAprobado), dir, modelo)
	escribirEnLaCopia(t, dir, "eval.txt", ficheroSintetico(numero)+"\n")

	return dir
}

// Lo que escribirEjecucionConUmbrales pone en las sesiones del modo herramienta y
// en las de la eval sin binario ni servidor (contracts/evals-en-dos-modos.md
// §2.2, §3 y §4 de H21).
const (
	// servidorDeLaSesionSintetica es el servidor.json de una sesión del modo
	// herramienta: su directorio lo tiene, y por eso es de ese modo.
	servidorDeLaSesionSintetica = `{"mcpServers":{"kitlegal":{"command":"` + kitlegalDelServidor +
		`","args":["mcp","serve"]}}}` + "\n"

	// trazaDelServidorSintetico es, en una sesión del modo herramienta, el
	// proceso que crea claude: el del servidor, y no el de la orden que lee el
	// bloque.
	trazaDelServidorSintetico = `execve("` + kitlegalDelServidor + `", ["` + kitlegalDelServidor +
		`", "mcp", "serve"], 0x7ffd8f13a6c0 /* 25 vars */) = 0` + "\n+++ exited with 0 +++\n"

	// llamadaSinteticaAlArticulo21 son los dos mensajes del transcript de una
	// sesión del modo herramienta con los que lee el bloque de la eval del
	// art. 21: la llamada a la herramienta, con el prefijo que le pone el agente,
	// y su resultado, el sobre de éxito.
	llamadaSinteticaAlArticulo21 = `{"type":"assistant","message":{"id":"msg_sintetico_llamada","type":"message",` +
		`"role":"assistant","model":"claude-haiku-4-5-20251001","content":[{"type":"tool_use",` +
		`"id":"toolu_sintetico_llamada","name":"mcp__kitlegal__boe_articulo",` +
		`"input":{"norma":"BOE-A-2015-10565","bloque":"a21"}}],"stop_reason":null,"stop_sequence":null},` +
		`"parent_tool_use_id":null,"session_id":"00000000-0000-4000-8000-000000000101"}` + "\n" +
		`{"type":"user","message":{"role":"user","content":[{"tool_use_id":"toolu_sintetico_llamada",` +
		`"type":"tool_result","content":[{"type":"text","text":"{\"ok\":true,` +
		`\"fuente\":\"boe.legislacion-consolidada\",\"data\":{\"bloque\":\"a21\"}}"}]}]},` +
		`"parent_tool_use_id":null,"session_id":"00000000-0000-4000-8000-000000000101"}` + "\n"

	// principioDelMensajeDelAsistente es el de cada mensaje assistant de un
	// transcript sintético.
	principioDelMensajeDelAsistente = `{"type":"assistant",`

	// evalSinBinarioNiServidor es la eval sin binario ni servidor de una
	// ejecución sintética, con la pregunta de la del art. 21 del caso aprobado,
	// que es la de sus sesiones.
	evalSinBinarioNiServidor = "# Eval sin binario ni servidor: la sesión tiene la skill y nada más.\n" +
		"pregunta: \"¿qué dice el art. 21 de la Ley 39/2015?\"\n" +
		"activa: true\n" +
		"sin_binario_ni_servidor: true\n"
)

// pasarAlModoHerramienta deja la sesión del art. 21 del directorio como una del
// modo herramienta que pasa igual: con servidor.json; con el proceso del
// servidor, y no el de la orden que lee el bloque, como el único que crea claude
// en su traza; y con la llamada a boe_articulo y su resultado en su transcript,
// delante del mensaje con la respuesta.
func pasarAlModoHerramienta(t *testing.T, dir string) {
	t.Helper()

	escribirEnLaCopia(t, dir, "servidor.json", servidorDeLaSesionSintetica)
	escribirEnLaCopia(t, filepath.Join(dir, directorioDeLaTraza), "t.2000", trazaDelServidorSintetico)

	lineas := strings.SplitAfter(contenidoDeLaSesion(t, dir, "sesion.jsonl"), "\n")

	conLaRespuesta := -1

	for posicion, linea := range lineas {
		if strings.HasPrefix(linea, principioDelMensajeDelAsistente) {
			conLaRespuesta = posicion
		}
	}

	require.GreaterOrEqual(t, conLaRespuesta, 0, "el transcript de %s tiene el mensaje con la respuesta", dir)
	escribirEnLaCopia(t, dir, "sesion.jsonl",
		strings.Join(slices.Insert(lineas, conLaRespuesta, llamadaSinteticaAlArticulo21), ""))

	sesion, err := LeerSesion(dir)
	require.NoError(t, err)
	require.Len(t, sesion.Llamadas, 1, "la sesión de %s lee el bloque con una llamada", dir)
}

// escribirSesionSinBinarioNiServidor escribe la sesión de esa repetición de la
// eval sin binario ni servidor de ese número con el modelo dado, y devuelve su
// directorio: una copia de la del art. 21 del caso aprobado que no tiene
// servidor.json, cuya traza no tiene más proceso que claude y que responde con la
// línea ⚠ SIN CONSULTA AL BOE: y su dirección, sin ninguna cita.
func escribirSesionSinBinarioNiServidor(t *testing.T, copia string, numero int, modelo string, vez int) string {
	t.Helper()

	dir := escribirCopiaDeLaSesion(t, copia, sesionDelArticulo21, sesionSintetica(numero, modelo, vez), numero, modelo)

	traza := filepath.Join(dir, directorioDeLaTraza)
	deClaude := contenidoDeLaSesion(t, traza, "t.1000")
	require.Equal(t, 1, strings.Count(deClaude, creacionDeLaLectura),
		"la traza de claude crea una sola vez el proceso que lee el bloque")
	escribirEnLaCopia(t, traza, "t.1000", strings.Replace(deClaude, creacionDeLaLectura, "", 1))
	require.NoError(t, os.Remove(filepath.Join(traza, "t.2000")))

	escribirEnLaCopia(t, dir, "sesion.jsonl", sustituirDosVeces(t, contenidoDeLaSesion(t, dir, "sesion.jsonl"),
		cadenaJSON(t, respuestaConCita), cadenaJSON(t, lineaSinConsultaAlBOE)))

	return dir
}

// ficheroSintetico es el fichero de la eval de ese número de una ejecución de
// ejecucionConUmbrales.
func ficheroSintetico(numero int) string {
	return fmt.Sprintf("%02d-sintetica.yaml", numero)
}

// sesionSintetica es el nombre de la sesión de esa repetición de la eval de ese
// número con el modelo dado, como los nombra el plan en el modo orden y en una
// eval sin binario ni servidor.
func sesionSintetica(numero int, modelo string, vez int) string {
	return fmt.Sprintf("%02d-sintetica-%s-%02d", numero, modelo, vez)
}

// sesionSinteticaEn es el nombre de la sesión de esa repetición de la eval de ese
// número con el modelo dado en el modo dado, como los nombra el plan: en el modo
// herramienta, con el modo detrás de la eval.
func sesionSinteticaEn(modo Modo, numero int, modelo string, vez int) string {
	if modo == ModoHerramienta {
		return fmt.Sprintf("%02d-sintetica-herramienta-%s-%02d", numero, modelo, vez)
	}

	return sesionSintetica(numero, modelo, vez)
}

// sesionesSinMedirDe son las sesiones sin medir que da el mensaje del límite de
// uso en todas las sesiones del modelo de las primeras evals de una ejecución de
// ejecucionConUmbrales, en orden de sesión.
func sesionesSinMedirDe(modelo string, evals int) []SesionSinMedir {
	var sinMedir []SesionSinMedir

	for numero := 1; numero <= evals; numero++ {
		for vez := 1; vez <= repeticionesConUmbrales; vez++ {
			sinMedir = append(sinMedir, SesionSinMedir{
				Sesion: sesionSintetica(numero, modelo, vez), Eval: ficheroSintetico(numero),
				Modelo: modelo, Motivo: sinMedirPorElMensaje,
			})
		}
	}

	return sinMedir
}
