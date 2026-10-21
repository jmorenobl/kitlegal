package evals

import (
	"encoding/json/jsontext"
	"encoding/json/v2"
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
	// sintética, que el test escribe en ella (escribirElJuezDelInforme): sin
	// ella, la skill no tiene juez. Y sinLista, que no lleva la lista de
	// expresiones prohibidas del repositorio.
	sinJuez, sinLista bool

	// medidaDelJuez es el contenido de la medida de la carpeta juez; vacío, el
	// de medidaDelJuezDelInforme, que se cumple.
	medidaDelJuez string

	// clasesDelJuez es el contenido de clases.yaml de la carpeta juez; vacío, el
	// de clasesDelJuezDelInforme, con el umbral 0.
	clasesDelJuez string

	// conTextos dice que el transcript de cada sesión del modo orden lleva la
	// orden con la que lee el bloque y lo que devolvió, que son los textos de
	// sus herramientas; las del modo herramienta llevan siempre los de su
	// llamada. Y conLaPruebaDeRed, que lleva además la sesión de la prueba de
	// red: la del modo orden de su primera eval con modeloSonnet55, con la
	// pregunta ampliada.
	conTextos, conLaPruebaDeRed bool

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
// objetivo, esos cuatro siguen decidiendo: cada uno incumplido sigue dando
// fallo.
//
// Con el juez (contracts/informe-del-job.md §1, §2 y §9 de H24; FR-030 a
// FR-034, FR-037, FR-044), una skill que lo tiene lleva además los suyos:
// detrás del de las respuestas sin activar de cada modo, el de cada clase del
// juez sobre esas mismas respuestas; detrás de los de los dos modos, los dos de
// la medida versionada del juez, con sus recuentos y sus totales, que no se
// cumplen si la medida no se cumple; y detrás de los de la duración de las
// sesiones, el de la duración del juez de cada modo. Aquí el juez dice no a
// todo —lo que hace con sus votos lo fija TestInformeConElJuez— y se le pide
// un voto por respuesta medida de las evals que activan la skill: ninguno de
// las sin medir, de las sin terminar, de la eval de no activación ni de una
// skill sin juez, y tres más por la eval sin binario ni servidor.
//
// Desde H26 (contracts/umbrales-del-informe.md §1, §2 y §6 de H26; data-model
// §3 de H26; FR-002, FR-005, FR-010, FR-012, FR-031, FR-033, FR-034), el de una
// clase del juez en un modo es un recuento que no decide, y la clase que decide
// tiene además su umbral conjunto, sin modo en su nombre, detrás de los de las
// respuestas de todos los modos y delante de los dos de la medida: su medida y
// su total son la suma de los de sus recuentos por modo y, con un plan de un
// solo modo, los de ese modo. Con el plan del job y el objetivo son los trece
// del contrato, en su orden, nueve de ellos decidiendo. Lo que ya decidía sigue
// dando su fallo entre ellos: una respuesta sin activar, 901 s de una tanda y
// la medida del juez sin cumplirse; y las respuestas de la eval sin binario ni
// servidor siguen fuera de toda medida, también de la del conjunto.
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
			umbrales:  umbralesConJuez(modeloSonnet55, nil, enOrden(54, 1)),
			motivos:   []string{"umbral sin_activar:claude-sonnet-5-5:orden: 1 de 54 (1,9 %), y tiene que ser ≤ 0,0 %"},
			veredicto: VeredictoFallo,
		},
		{
			// Las sesiones de la eval de no activación, que no activan la skill como
			// ella espera, no cuentan: ni como sin activar ni en el total.
			nombre:    "sin-activar-ninguna",
			ejecucion: conCambios(comoElJob, func(e *ejecucionConUmbrales) { e.conLaDeNoActivacion = true }),
			umbrales:  umbralesConJuez(modeloSonnet55, nil, enOrden(54, 0)),
			veredicto: VeredictoAprobado,
			exigir:    exigirLaDeNoActivacionSinContar,
		},
		{
			// Tres respuestas con expresiones de la lista, de 54: hasta H24, un 5,6 %
			// que no cumplía su umbral. La lista está en la carpeta y no juzga nada.
			nombre:    "tres-con-expresiones-de-la-lista",
			ejecucion: conCambios(comoElJob, func(e *ejecucionConUmbrales) { e.conExpresiones = 3 }),
			umbrales:  umbralesConJuez(modeloSonnet55, nil, enOrden(54, 0)),
			veredicto: VeredictoAprobado,
			exigir:    exigirLasTresConExpresionesQuePasan,
		},
		{
			// Con las seis sin terminar en el total, sería 1 de 60 (1,7 %).
			nombre: "una-sin-activar-y-seis-sin-terminar",
			ejecucion: conCambios(comoElJob, func(e *ejecucionConUmbrales) {
				e.sinActivar, e.sinTerminar = 1, 2
			}),
			umbrales:  umbralesConJuez(modeloSonnet55, nil, enOrden(54, 1)),
			motivos:   []string{"umbral sin_activar:claude-sonnet-5-5:orden: 1 de 54 (1,9 %), y tiene que ser ≤ 0,0 %"},
			veredicto: VeredictoFallo,
			exigir:    exigirLasSeisSinTerminar,
		},
		{
			nombre: "una-sin-activar-con-expresiones",
			ejecucion: conCambios(comoElJob, func(e *ejecucionConUmbrales) {
				e.sinActivar, e.conExpresiones = 1, 1
			}),
			umbrales:  umbralesConJuez(modeloSonnet55, nil, enOrden(54, 1)),
			motivos:   []string{"umbral sin_activar:claude-sonnet-5-5:orden: 1 de 54 (1,9 %), y tiene que ser ≤ 0,0 %"},
			veredicto: VeredictoFallo,
			exigir:    exigirLaMismaSesionSinActivarConExpresiones,
		},
		{
			// Sin juez, ni la respuesta sin activar es un umbral, aunque la carpeta
			// tenga lista; y sin objetivo, tampoco ninguna duración (FR-037 de
			// H24). Las evals sintéticas no declaran sentencias, plazo ni
			// no_cubierto, que darían umbrales de las respuestas sin juez.
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
			umbrales:  umbralesConJuez(modeloSonnet55, nil, enOrden(6, 1)),
			motivos:   []string{"umbral sin_activar:claude-sonnet-5-5:orden: 1 de 6 (16,7 %), y tiene que ser ≤ 0,0 %"},
			veredicto: VeredictoFallo,
		},
		{
			nombre:    "901-s-con-objetivo-900",
			ejecucion: conCambios(minima, func(e *ejecucionConUmbrales) { e.duracion, e.objetivo = 901, 900 }),
			umbrales: umbralesConJuez(modeloSonnet55, []Umbral{umbralDeDuracion(ModoOrden, 901, 900, false)},
				enOrden(6, 0)),
			motivos:   []string{"de la ejecución, no de la skill: duracion_de_las_sesiones:orden: 901 s, y tiene que ser ≤ 900 s"},
			veredicto: VeredictoFallo,
		},
		{
			nombre:    "900-s-con-objetivo-900",
			ejecucion: conCambios(minima, func(e *ejecucionConUmbrales) { e.duracion, e.objetivo = 900, 900 }),
			umbrales: umbralesConJuez(modeloSonnet55, []Umbral{umbralDeDuracion(ModoOrden, 900, 900, true)},
				enOrden(6, 0)),
			veredicto: VeredictoAprobado,
		},
		{
			// La medida versionada del juez no se cumple: sus dos umbrales llevan
			// los recuentos y los totales de la medida, sin votar ningún caso, y
			// cada uno pone el veredicto en fallo (FR-032, FR-044), detrás del
			// conjunto, que se cumple.
			nombre:    "la-medida-del-juez-sin-cumplirse",
			ejecucion: conCambios(minima, func(e *ejecucionConUmbrales) { e.medidaDelJuez = medidaDelJuezSinCumplirse }),
			umbrales: slices.Concat(
				umbralesDeUnModo(modeloSonnet55, enOrden(6, 0)),
				[]Umbral{umbralConjunto(modeloSonnet55, enOrden(6, 0))},
				umbralesDeLaMedidaDelJuez(1, 20, 2, 5),
				[]Umbral{umbralDeLaDuracionDelJuez(ModoOrden, 0)}),
			motivos: []string{
				"umbral medida_del_juez:afirma_lo_no_leido:defectos_sin_marcar: 1 de 20 (5,0 %), y tiene que ser ≤ 0,0 %",
				"umbral medida_del_juez:afirma_lo_no_leido:correctos_marcados: 2 de 5 (40,0 %), y tiene que ser ≤ 0,0 %",
			},
			veredicto: VeredictoFallo,
		},
		{
			// Sin ninguna respuesta medida, la proporción es 0 y el umbral se
			// cumple; el veredicto ya es fallo por las sesiones sin medir, y sus
			// series, sin medir, no dan el motivo de su tasa.
			nombre:    "todas-las-de-sonnet-5-5-sin-medir",
			ejecucion: conCambios(minima, func(e *ejecucionConUmbrales) { e.sinMedir = modeloSonnet55 }),
			umbrales:  umbralesConJuez(modeloSonnet55, nil, enOrden(0, 0)),
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
			umbrales: umbralesConJuez(modeloSonnet55, []Umbral{umbralDeDuracion(ModoOrden, 901, 900, false)},
				enOrden(6, 1)),
			motivos: []string{
				"umbral sin_activar:claude-sonnet-5-5:orden: 1 de 6 (16,7 %), y tiene que ser ≤ 0,0 %",
				motivoEsperadoDelLimite(sesionesSinMedirDe(modeloHaiku45, 1)...),
				"de la ejecución, no de la skill: duracion_de_las_sesiones:orden: 901 s, y tiene que ser ≤ 900 s",
			},
			veredicto: VeredictoFallo,
		},
		{
			// Las respuestas sin activar de un modo no se suman a las del otro
			// (FR-043 de H21).
			nombre:    "sin-activar-solo-en-orden",
			ejecucion: conCambios(enDosModos, func(e *ejecucionConUmbrales) { e.sinActivar = 1 }),
			umbrales:  umbralesConJuez(modeloSonnet55, nil, enOrden(54, 1), enHerramienta(54, 0)),
			motivos:   []string{"umbral sin_activar:claude-sonnet-5-5:orden: 1 de 54 (1,9 %), y tiene que ser ≤ 0,0 %"},
			veredicto: VeredictoFallo,
		},
		{
			nombre:    "sin-activar-solo-en-herramienta",
			ejecucion: conCambios(enDosModos, func(e *ejecucionConUmbrales) { e.enHerramienta.sinActivar = 1 }),
			umbrales:  umbralesConJuez(modeloSonnet55, nil, enOrden(54, 0), enHerramienta(54, 1)),
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
			umbrales: umbralesConJuez(modeloSonnet55, []Umbral{
				umbralDeDuracion(ModoOrden, 900, 900, true),
				umbralDeDuracion(ModoHerramienta, 901, 900, false),
			}, enOrden(6, 0), enHerramienta(6, 0)),
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
			umbrales: umbralesConJuez(modeloSonnet55, []Umbral{
				umbralDeDuracion(ModoOrden, 901, 900, false),
				umbralDeDuracion(ModoHerramienta, 900, 900, true),
			}, enOrden(6, 0), enHerramienta(6, 0)),
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
			umbrales: umbralesConJuez(modeloSonnet55, []Umbral{
				umbralDeDuracion(ModoOrden, 900, 900, true),
				umbralDeDuracion(ModoHerramienta, 900, 900, true),
			}, enOrden(6, 0), enHerramienta(6, 0)),
			veredicto: VeredictoAprobado,
			exigir:    exigirLosTreceDelContrato,
		},
		{
			// Las sesiones de la eval sin binario ni servidor no son de ningún modo:
			// ni la que no activa la skill cuenta en una medida, ni las seis en un
			// total —tampoco en el del conjunto, que es 108—, ni los 5000 s de su
			// tanda en una duración (FR-047 de H21; FR-002 de H26).
			nombre: "la-sin-binario-ni-servidor-fuera-de-toda-medida",
			ejecucion: conCambios(enDosModos, func(e *ejecucionConUmbrales) {
				e.conLaSinBinarioNiServidor, e.sinBinarioAlteradas = true, 1
				e.duracion, e.duracionEnHerramienta, e.duracionSinModo, e.objetivo = 500, 600, 5000, 900
			}),
			umbrales: umbralesConJuez(modeloSonnet55, []Umbral{
				umbralDeDuracion(ModoOrden, 500, 900, true),
				umbralDeDuracion(ModoHerramienta, 600, 900, true),
			}, enOrden(54, 0), enHerramienta(54, 0)),
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
			assert.Len(t, leido.votos, votosDeUnJuezQueDiceNo(caso.ejecucion, caso.umbrales),
				"un voto por respuesta medida de las evals que activan la skill, y ninguno más")

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
		motivosDeLosUmbrales(umbrales, nil, nil))
	assert.Empty(t, motivosDeLaDuracionDeLasSesiones(umbrales))
	assert.Equal(t, [][]string{
		{"`sin_activar:claude-sonnet-5-5:orden`", "1 de 54 (1,9 %)", "≤ 0,0 %", "no", "sí"},
		{"`se_publica:claude-haiku-4-5-20251001:orden`", "2 de 30 (6,7 %)", "≤ 5,0 %", "no", "no: solo se publica"},
	}, filasDeUmbrales(umbrales))
}

// TestUmbralesConJuezYConSentencias fija los umbrales de las respuestas de una
// skill que tiene juez y alguna eval que declara sentencias
// (contracts/evals-jurisprudencia.md §4 de H23; research D20 de H23). Desde
// H25, jurisprudencia tiene las dos cosas, y sus umbrales los fija
// TestUmbralesDeJurisprudencia con las evals del repositorio; este fija, con
// respuestas escritas aquí, el orden de los de cada modo y sus motivos. Por
// cada modo, y en este orden, el de las respuestas sin activar, el de cada
// clase del juez y el de las que llevan una cita sin documento cotejado o un
// ECLI sin origen, los tres sobre las mismas respuestas; las de las evals sin
// binario ni servidor no cuentan en ninguno; y el motivo de cada uno que
// decide y no se cumple nombra, si las tiene, las respuestas que cuentan en él.
//
// Desde H26 (contracts/umbrales-del-informe.md §1 de H26; FR-005, FR-010), el
// de la clase que decide no decide en ningún modo, y su umbral conjunto, con
// las respuestas de los dos modos y ninguna de las evals sin binario ni
// servidor, va detrás de cita_sin_documento del último modo y delante de los
// dos de la medida, que aquí es una sin casos, y de la duración de los votos de
// cada modo.
func TestUmbralesConJuezYConSentencias(t *testing.T) {
	t.Parallel()

	sinMarcar := JuicioDeRespuesta{Clases: []JuicioDeClase{{Clase: "afirma_lo_no_leido"}, {Clase: "cuenta_su_proceso"}}}
	sinOrigen := []ECLISinDocumento{{ECLI: "ECLI:ES:TS:2023:9999", Condicion: ECLISinOrigen}}

	grupos := []grupoJuzgado{
		{modo: ModoOrden, respuestas: []respuestaDelJuez{
			{sesion: "01-a-01", activada: true, juicio: sinMarcar},
			{sesion: "01-a-02", activada: true, sinDocumento: sinOrigen, juicio: sinMarcar},
			{sesion: "01-a-03", juicio: sinMarcar},
		}},
		{modo: ModoHerramienta, respuestas: []respuestaDelJuez{
			{sesion: "01-a-herramienta-01", activada: true, juicio: sinMarcar},
		}},
		{respuestas: []respuestaDelJuez{{sesion: "02-sin-binario-01", sinDocumento: sinOrigen, juicio: sinMarcar}}},
	}

	respuestas := &respuestasMedidas{
		modelo: modeloSonnet55,
		grupos: grupos,
		delJuez: &juicioDelJuez{
			modelo: modeloSonnet55,
			clases: []ClaseDelJuez{{Nombre: "afirma_lo_no_leido", Decide: true}, {Nombre: "cuenta_su_proceso"}},
			grupos: grupos,
		},
		conSentencias: true,
	}

	umbrales := umbralesDelInforme(InformeAEscribir{}, respuestas, nil)

	assert.Equal(t, slices.Concat(
		umbralesDeUnModo(modeloSonnet55, enOrden(3, 1)),
		[]Umbral{umbralCitaSinDocumento(modeloSonnet55, ModoOrden, 1, 3, false)},
		umbralesDeUnModo(modeloSonnet55, enHerramienta(1, 0)),
		[]Umbral{
			umbralCitaSinDocumento(modeloSonnet55, ModoHerramienta, 0, 1, true),
			umbralConjunto(modeloSonnet55, enOrden(3, 1), enHerramienta(1, 0)),
		},
		umbralesDeLaMedidaDelJuez(0, 0, 0, 0),
		[]Umbral{umbralDeLaDuracionDelJuez(ModoOrden, 0), umbralDeLaDuracionDelJuez(ModoHerramienta, 0)},
	), umbrales)

	assert.Equal(t, []string{
		"umbral sin_activar:claude-sonnet-5-5:orden: 1 de 3 (33,3 %), y tiene que ser ≤ 0,0 %",
		"umbral cita_sin_documento:claude-sonnet-5-5:orden: 1 de 3 (33,3 %), y tiene que ser ≤ 0,0 %: " +
			"01-a-02: ECLI:ES:TS:2023:9999 (sin origen)",
	}, motivosDeLosUmbrales(umbrales, respuestas, nil))
}

// TestUmbralConjunto fija lo que el umbral conjunto de la clase del juez que
// decide hace con la tasa de su declaración (contracts/umbrales-del-informe.md
// §1 a §4 de H26; data-model §1, §3 y §5 de H26; research D6 y D7 de H26;
// FR-001 a FR-005, FR-010, FR-072, FR-074; SC-002, SC-004), con la ejecución
// sintética «como el cierre» en dos modos —18 evals que activan la skill, 54
// respuestas del modelo que decide por modo, y el objetivo de duración del
// job—, escrita con EscribirInforme en t.TempDir(), y un votante de votos
// grabados que marca con tres síes las respuestas del caso. La carpeta del juez
// de esa ejecución lleva la clases.yaml de evals/boe-legislacion/juez, leída
// del repositorio: el umbral con el que el informe compara es el dato que leerá
// el job y no una copia, y lo que el test espera de él va escrito a mano
// —0.028, «≤ 2,8 %», 3 de 108—, de modo que otro valor en el repositorio no
// pasa.
//
//   - Con 3 marcadas de 108, el conjunto se cumple se repartan como se
//     repartan entre los modos —3 y 0, 0 y 3, 2 y 1—: el veredicto es aprobado
//     y no hay ningún motivo. El recuento del modo con las tres lleva 3 de 54,
//     no se cumple y no decide.
//   - Con 4, las cuatro en un modo o dos en cada uno, no se cumple: el
//     veredicto es fallo con un solo motivo, el del conjunto, que nombra las
//     cuatro respuestas, las del modo orden delante, cada una con su modo y
//     con sus tres frases.
//   - Con cuatro sesiones del modelo que decide que el tope cortó tras su
//     respuesta, una por serie y dos por modo, en series informativas, quedan
//     104 respuestas juzgadas sin ningún otro motivo de fallo: con 2 marcadas
//     el conjunto se cumple, y con 3, que de 108 lo cumplían, no.
//
// En todos, umbrales son los trece del contrato, en su orden, con el umbral
// 0.028 en el conjunto y en los dos recuentos por modo, y deciden nueve. Y con
// las tres en un modo, su recuento y el conjunto están escritos en informe.json
// como los dos ejemplos de boe-legislacion del contrato: tal cual en 0-y-3, que
// es el informe de esos ejemplos.
func TestUmbralConjunto(t *testing.T) {
	t.Parallel()

	comoElCierre := ejecucionConUmbrales{
		queDeciden: 10, informativas: 8, conHaiku: true, dosModos: true,
		duracion: 500, duracionEnHerramienta: 600, objetivo: 900,
		clasesDelJuez: string(contenidoDelFichero(t, filepath.Join(evalsDelRepositorio, clasesEnElJuez))),
	}

	for _, caso := range casosDelUmbralConjunto() {
		t.Run(caso.nombre, func(t *testing.T) {
			t.Parallel()

			exigirElUmbralConjunto(t, comoElCierre, caso)
		})
	}
}

// Los otros dos párrafos con preceptos que TestUmbralConjunto antepone a las
// respuestas que marca, con parrafoDeLosArticulos22Y24 y
// parrafoDeLosArticulos23Y25: cada uno distinto de los demás, con las tres
// frases suyas que citan sus votos y esas tres como las escribe el motivo de la
// respuesta que marcan.
const (
	parrafoDeLosArticulos29Y30 = "El artículo 29 obliga a cumplir los plazos, y el artículo 30, a contarlos por días."
	fraseDelArticulo29         = "El artículo 29 obliga a cumplir los plazos"
	fraseDeLosArticulos29Y30   = "obliga a cumplir los plazos, y el artículo 30"
	fraseDelArticulo30         = "el artículo 30, a contarlos por días"
	frasesDel29Y30             = "«" + fraseDelArticulo29 + "» · «" + fraseDeLosArticulos29Y30 + "» · «" + fraseDelArticulo30 + "»"

	parrafoDeLosArticulos32Y33 = "El artículo 32 permite ampliar los plazos, y el artículo 33, tramitar con urgencia."
	fraseDelArticulo32         = "El artículo 32 permite ampliar los plazos"
	fraseDeLosArticulos32Y33   = "permite ampliar los plazos, y el artículo 33"
	fraseDelArticulo33         = "el artículo 33, tramitar con urgencia"
	frasesDel32Y33             = "«" + fraseDelArticulo32 + "» · «" + fraseDeLosArticulos32Y33 + "» · «" + fraseDelArticulo33 + "»"
)

// Los dos ejemplos de boe-legislacion de contracts/umbrales-del-informe.md §2 y
// §3 de H26, tal cual: el umbral conjunto con 3 marcadas de 108, que se cumple
// y decide, y el recuento del modo herramienta de ese mismo informe, con las
// tres, que no se cumple y no decide.
const (
	conjuntoDelContrato = `{"nombre":"afirma_lo_no_leido:claude-sonnet-5-5","descripcion":"Respuestas de ` +
		`claude-sonnet-5-5, con todos los modos juntos, que el juez marca en afirma_lo_no_leido con sus tres votos, ` +
		`sobre sus respuestas juzgadas en las evals que activan la skill","medida":3,"total":108,"comparacion":"<=",` +
		`"umbral":0.028,"cumple":true,"decide":true}`

	recuentoDelContrato = `{"nombre":"afirma_lo_no_leido:claude-sonnet-5-5:herramienta","descripcion":"Respuestas de ` +
		`claude-sonnet-5-5 en el modo herramienta que el juez marca en afirma_lo_no_leido con sus tres votos, sobre sus ` +
		`respuestas juzgadas en las evals que activan la skill","medida":3,"total":54,"comparacion":"<=","umbral":0.028,` +
		`"cumple":false,"decide":false}`
)

// casoDelUmbralConjunto es un caso de TestUmbralConjunto: las respuestas que el
// juez marca en cada modo, si cuatro sesiones quedan sin terminar y lo que el
// informe tiene que decir, escrito a mano.
type casoDelUmbralConjunto struct {
	nombre string

	// enOrden y enHerramienta son cuántas respuestas de cada modo llevan tres
	// votos que dicen sí en afirma_lo_no_leido: la primera sesión del modelo que
	// decide de otras tantas evals, desde la primera.
	enOrden, enHerramienta int

	// conCortadas dice que el tope cortó tras su respuesta las cuatro sesiones
	// de sesionesCortadasDelUmbralConjunto.
	conCortadas bool

	// porModo son las respuestas juzgadas de cada modo; ordenCumple,
	// herramientaCumple y conjuntoCumple, si cumplen la tasa el recuento de
	// afirma_lo_no_leido de cada modo y su umbral conjunto; motivos, los de la
	// raíz; y veredicto, el del informe.
	porModo                                        int
	ordenCumple, herramientaCumple, conjuntoCumple bool
	motivos                                        []string
	veredicto                                      Veredicto

	// exigir, si no es nil, exige lo que el caso fija además de sus umbrales, sus
	// motivos y su veredicto.
	exigir func(t *testing.T, leido informeLeido)
}

// casosDelUmbralConjunto son los de TestUmbralConjunto. Lo que cumple cada
// elemento es lo de contracts/umbrales-del-informe.md §2 de H26: 3 de 108 y 2 de
// 104 caben en el 2,8 %, y 4 de 108 y 3 de 104, no; y en un modo, 1 de 54 y 1 de
// 52 caben, y 2 de 54 y 2 de 52, no.
func casosDelUmbralConjunto() []casoDelUmbralConjunto {
	return []casoDelUmbralConjunto{
		{
			nombre: "3-y-0", enOrden: 3,
			porModo: 54, ordenCumple: false, herramientaCumple: true, conjuntoCumple: true,
			veredicto: VeredictoAprobado,
			exigir:    exigirElRecuentoConLasTres(ModoOrden),
		},
		{
			nombre: "0-y-3", enHerramienta: 3,
			porModo: 54, ordenCumple: true, herramientaCumple: false, conjuntoCumple: true,
			veredicto: VeredictoAprobado,
			exigir:    exigirElRecuentoConLasTres(ModoHerramienta),
		},
		{
			nombre: "2-y-1", enOrden: 2, enHerramienta: 1,
			porModo: 54, ordenCumple: false, herramientaCumple: true, conjuntoCumple: true,
			veredicto: VeredictoAprobado,
		},
		{
			nombre: "4-y-0", enOrden: 4,
			porModo: 54, ordenCumple: false, herramientaCumple: true, conjuntoCumple: false,
			motivos: []string{
				"umbral afirma_lo_no_leido:claude-sonnet-5-5: 4 de 108 (3,7 %), y tiene que ser ≤ 2,8 %: " +
					"01-sintetica-claude-sonnet-5-5-01 (modo orden): " + frasesDel22Y24 +
					"; 02-sintetica-claude-sonnet-5-5-01 (modo orden): " + frasesDel23Y25 +
					"; 03-sintetica-claude-sonnet-5-5-01 (modo orden): " + frasesDel29Y30 +
					"; 04-sintetica-claude-sonnet-5-5-01 (modo orden): " + frasesDel32Y33,
			},
			veredicto: VeredictoFallo,
		},
		{
			nombre: "2-y-2", enOrden: 2, enHerramienta: 2,
			porModo: 54, ordenCumple: false, herramientaCumple: false, conjuntoCumple: false,
			motivos: []string{
				"umbral afirma_lo_no_leido:claude-sonnet-5-5: 4 de 108 (3,7 %), y tiene que ser ≤ 2,8 %: " +
					"01-sintetica-claude-sonnet-5-5-01 (modo orden): " + frasesDel22Y24 +
					"; 02-sintetica-claude-sonnet-5-5-01 (modo orden): " + frasesDel23Y25 +
					"; 01-sintetica-herramienta-claude-sonnet-5-5-01 (modo herramienta): " + frasesDel29Y30 +
					"; 02-sintetica-herramienta-claude-sonnet-5-5-01 (modo herramienta): " + frasesDel32Y33,
			},
			veredicto: VeredictoFallo,
		},
		{
			nombre: "104-con-2", enOrden: 1, enHerramienta: 1, conCortadas: true,
			porModo: 52, ordenCumple: true, herramientaCumple: true, conjuntoCumple: true,
			veredicto: VeredictoAprobado,
		},
		{
			nombre: "104-con-3", enOrden: 2, enHerramienta: 1, conCortadas: true,
			porModo: 52, ordenCumple: false, herramientaCumple: true, conjuntoCumple: false,
			motivos: []string{
				"umbral afirma_lo_no_leido:claude-sonnet-5-5: 3 de 104 (2,9 %), y tiene que ser ≤ 2,8 %: " +
					"01-sintetica-claude-sonnet-5-5-01 (modo orden): " + frasesDel22Y24 +
					"; 02-sintetica-claude-sonnet-5-5-01 (modo orden): " + frasesDel23Y25 +
					"; 01-sintetica-herramienta-claude-sonnet-5-5-01 (modo herramienta): " + frasesDel29Y30,
			},
			veredicto: VeredictoFallo,
		},
	}
}

// umbrales son los trece del informe del caso, en el orden de
// contracts/umbrales-del-informe.md §1 de H26, con sus valores escritos a mano:
// los de umbralesConJuez con las marcadas y las respuestas juzgadas de cada
// modo, que dan al conjunto la suma de unas y la de otras, y con las dos
// tandas en su objetivo; y en los tres de afirma_lo_no_leido —el 2, el 5 y el 7
// de los trece—, la tasa de la declaración de boe-legislacion, 0.028, y si la
// cumple cada uno.
func (c casoDelUmbralConjunto) umbrales() []Umbral {
	trece := umbralesConJuez(modeloSonnet55,
		[]Umbral{umbralDeDuracion(ModoOrden, 500, 900, true), umbralDeDuracion(ModoHerramienta, 600, 900, true)},
		medidasDeUnModo{modo: ModoOrden, respuestas: c.porModo, marcadas: c.enOrden},
		medidasDeUnModo{modo: ModoHerramienta, respuestas: c.porModo, marcadas: c.enHerramienta})

	for posicion, cumple := range map[int]bool{1: c.ordenCumple, 4: c.herramientaCumple, 6: c.conjuntoCumple} {
		trece[posicion].Umbral, trece[posicion].Cumple = 0.028, cumple
	}

	return trece
}

// grabados son los votos grabados de las respuestas que el caso marca: de cada
// una, tres que dicen sí en afirma_lo_no_leido, cada uno con una de las tres
// frases del párrafo que el test le antepone. La primera marcada del modo orden
// lleva el de los artículos 22 y 24, y las siguientes, las del modo herramienta
// detrás de las del modo orden, los de los 23 y 25, los 29 y 30 y los 32 y 33.
// Van con las del modo herramienta delante y en orden inverso al de sus
// sesiones: el motivo las nombra en el orden de los modos y de las sesiones, no
// en el de sus votos.
func (c casoDelUmbralConjunto) grabados(t *testing.T) []votosDeUnaRespuesta {
	t.Helper()

	marcas := []struct {
		parrafo string
		frases  []string
	}{
		{parrafoDeLosArticulos22Y24, []string{fraseDelArticulo22, fraseDeLosDosArticulos, fraseDelArticulo24}},
		{parrafoDeLosArticulos23Y25, []string{fraseDelArticulo23, fraseDeLosArticulos23Y25, fraseDelArticulo25}},
		{parrafoDeLosArticulos29Y30, []string{fraseDelArticulo29, fraseDeLosArticulos29Y30, fraseDelArticulo30}},
		{parrafoDeLosArticulos32Y33, []string{fraseDelArticulo32, fraseDeLosArticulos32Y33, fraseDelArticulo33}},
	}

	marcadas := c.enOrden + c.enHerramienta
	require.LessOrEqual(t, marcadas, len(marcas), "hay un párrafo por respuesta que se marca")

	grabados := make([]votosDeUnaRespuesta, 0, marcadas)

	for posicion, marca := range marcas[:marcadas] {
		sesion := sesionSinteticaEn(ModoOrden, posicion+1, modeloSonnet55, 1)
		if posicion >= c.enOrden {
			sesion = sesionSinteticaEn(ModoHerramienta, posicion-c.enOrden+1, modeloSonnet55, 1)
		}

		votos := make([]grabacion, 0, len(marca.frases))
		for _, frase := range marca.frases {
			votos = append(votos, votoDeLasDosClases(t, afirmaQueSi(frase), cuentaQueNo()))
		}

		grabados = append(grabados, votosDeUnaRespuesta{sesion: sesion, parrafo: marca.parrafo, votos: votos})
	}

	slices.Reverse(grabados)

	return grabados
}

// sesionesCortadasDelUmbralConjunto son las cuatro sesiones del modelo que
// decide que el tope corta tras su respuesta en los casos de 104 respuestas
// juzgadas de TestUmbralConjunto (research D7 de H26): la tercera de cuatro
// series informativas, una por serie y dos por modo. No cuentan como respuesta
// medida, y su serie, que sigue llegando al umbral con las otras dos, no da
// ningún motivo.
func sesionesCortadasDelUmbralConjunto() []string {
	return []string{
		sesionSinteticaEn(ModoOrden, 11, modeloSonnet55, 3),
		sesionSinteticaEn(ModoOrden, 12, modeloSonnet55, 3),
		sesionSinteticaEn(ModoHerramienta, 13, modeloSonnet55, 3),
		sesionSinteticaEn(ModoHerramienta, 14, modeloSonnet55, 3),
	}
}

// exigirElUmbralConjunto arma la ejecución en un directorio temporal del test,
// corta las sesiones del caso, antepone a cada respuesta que el caso marca su
// párrafo y escribe su informe con el votante de sus votos: exige sus trece
// umbrales, con los nombres, las descripciones y los nueve que deciden del
// contrato, sus motivos y su veredicto.
func exigirElUmbralConjunto(t *testing.T, ejecucion ejecucionConUmbrales, caso casoDelUmbralConjunto) {
	t.Helper()

	copia := armarEjecucionConUmbrales(t, ejecucion)

	if caso.conCortadas {
		for _, sesion := range sesionesCortadasDelUmbralConjunto() {
			cortarTrasLaRespuesta(t, filepath.Join(copia, "sesiones", sesion))
		}
	}

	leido := escribirLaCopiaConVotos(t, copia, ejecucion, nuevoVotanteDelInforme(t, caso.grabados(t)...))

	exigirUmbrales(t, leido, caso.umbrales())
	exigirLosTreceDelContrato(t, leido)
	exigirMotivosDeLaRaiz(t, leido, caso.motivos...)
	assert.Equal(t, caso.veredicto, leido.informe.Veredicto)

	if caso.exigir != nil {
		caso.exigir(t, leido)
	}
}

// exigirElRecuentoConLasTres da lo que exigen además los casos 3-y-0 y 0-y-3 de
// TestUmbralConjunto (FR-074; SC-004), con las tres respuestas marcadas en el
// modo dado: su recuento está escrito en informe.json con 3 de 54, cumple false
// y decide false, que es el ejemplo de contracts/umbrales-del-informe.md §3 de
// H26 con ese modo donde lleva el modo herramienta —en 0-y-3, tal cual—; y el
// conjunto, como el ejemplo de §2, tal cual, con 3 de 108, cumplido y
// decidiendo. El veredicto aprobado y la falta de motivos los exige el caso.
func exigirElRecuentoConLasTres(modo Modo) func(t *testing.T, leido informeLeido) {
	return func(t *testing.T, leido informeLeido) {
		t.Helper()

		exigirElUmbralEscrito(t, leido, strings.ReplaceAll(recuentoDelContrato, string(ModoHerramienta), string(modo)))
		exigirElUmbralEscrito(t, leido, conjuntoDelContrato)
	}
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

// medidasDeUnModo son las medidas de los umbrales de las respuestas del modelo
// que decide en un modo de una ejecución con juez, y los segundos de sus votos.
type medidasDeUnModo struct {
	modo Modo

	// respuestas son las medidas del modelo en ese modo, el total de sus tres
	// umbrales; y sinActivar, marcadas y conSi, las medidas de cada uno: las que
	// no activaron la skill, las que el juez marca en afirma_lo_no_leido y las
	// que tienen sí en cuenta_su_proceso.
	respuestas, sinActivar, marcadas, conSi int

	// segundosDelJuez son los de los votos de las respuestas del modo.
	segundosDelJuez int
}

// enOrden y enHerramienta son las medidas de un modo en el que el juez no marca
// ninguna respuesta ni dice sí de ninguna, con sus votos en 0 s.
func enOrden(respuestas, sinActivar int) medidasDeUnModo {
	return medidasDeUnModo{modo: ModoOrden, respuestas: respuestas, sinActivar: sinActivar}
}

func enHerramienta(respuestas, sinActivar int) medidasDeUnModo {
	return medidasDeUnModo{modo: ModoHerramienta, respuestas: respuestas, sinActivar: sinActivar}
}

// umbralesConJuez son los umbrales del informe de una skill con juez, en el
// orden de contracts/umbrales-del-informe.md §1 de H26: los de las respuestas
// de cada modo (umbralesDeUnModo); el conjunto de afirma_lo_no_leido, con las
// de todos los modos dados (umbralConjunto); los dos de la medida del juez, con
// los recuentos de medidaDelJuezDelInforme, 0 de 212 y 0 de 47; los de la
// duración de las sesiones dados; y, por cada modo, el de la duración del juez.
func umbralesConJuez(modelo string, deLaDuracion []Umbral, porModo ...medidasDeUnModo) []Umbral {
	var umbrales []Umbral

	for _, medidas := range porModo {
		umbrales = append(umbrales, umbralesDeUnModo(modelo, medidas)...)
	}

	umbrales = append(umbrales, umbralConjunto(modelo, porModo...))
	umbrales = append(umbrales, umbralesDeLaMedidaDelJuez(0, 212, 0, 47)...)
	umbrales = append(umbrales, deLaDuracion...)

	for _, medidas := range porModo {
		umbrales = append(umbrales, umbralDeLaDuracionDelJuez(medidas.modo, medidas.segundosDelJuez))
	}

	return umbrales
}

// umbralConjunto es el umbral conjunto de afirma_lo_no_leido, la clase que
// decide, tal como lo fija contracts/umbrales-del-informe.md §2 de H26, con sus
// valores escritos a mano: sin modo en su nombre, con la suma de las marcadas y
// la de las respuestas de los modos dados —con uno solo, las de ese modo—, y
// decidiendo con el 0 de la declaración de clases de los tests del informe.
func umbralConjunto(modelo string, porModo ...medidasDeUnModo) Umbral {
	marcadas, respuestas := 0, 0

	for _, medidas := range porModo {
		marcadas += medidas.marcadas
		respuestas += medidas.respuestas
	}

	return Umbral{
		Nombre: "afirma_lo_no_leido:" + modelo,
		Descripcion: "Respuestas de " + modelo + ", con todos los modos juntos, que el juez marca en " +
			"afirma_lo_no_leido con sus tres votos, sobre sus respuestas juzgadas en las evals que activan la skill",
		Medida:      float64(marcadas),
		Total:       &respuestas,
		Comparacion: "<=",
		Umbral:      0,
		Cumple:      marcadas == 0,
		Decide:      true,
	}
}

// umbralesDeUnModo son los tres umbrales de las respuestas del modelo que
// decide en un modo, tal como los fijan contracts/informe-del-job.md §2 de H24
// y contracts/umbrales-del-informe.md §3 de H26, con sus valores escritos a
// mano: el de las respuestas sin activar, que decide; y los recuentos de las
// dos clases del juez, que llevan 0 y ninguno decide —el de afirma_lo_no_leido,
// porque quien decide es su umbral conjunto, y el de cuenta_su_proceso, porque
// solo se publica—. Los tres, sobre las mismas respuestas, y cada uno cumplido
// si su medida es 0.
func umbralesDeUnModo(modelo string, medidas medidasDeUnModo) []Umbral {
	deLasEvals := "sobre sus respuestas juzgadas en las evals que activan la skill"
	delModo := "Respuestas de " + modelo + " en el modo " + string(medidas.modo)

	afirma, cuenta := medidas.respuestas, medidas.respuestas

	return []Umbral{
		umbralSinActivar(modelo, medidas.modo, medidas.sinActivar, medidas.respuestas, medidas.sinActivar == 0),
		{
			Nombre:      "afirma_lo_no_leido:" + modelo + ":" + string(medidas.modo),
			Descripcion: delModo + " que el juez marca en afirma_lo_no_leido con sus tres votos, " + deLasEvals,
			Medida:      float64(medidas.marcadas),
			Total:       &afirma,
			Comparacion: "<=",
			Umbral:      0,
			Cumple:      medidas.marcadas == 0,
			Decide:      false,
		},
		{
			Nombre:      "cuenta_su_proceso:" + modelo + ":" + string(medidas.modo),
			Descripcion: delModo + " con sí en cuenta_su_proceso en el primer voto del juez, " + deLasEvals,
			Medida:      float64(medidas.conSi),
			Total:       &cuenta,
			Comparacion: "<=",
			Umbral:      0,
			Cumple:      medidas.conSi == 0,
			Decide:      false,
		},
	}
}

// umbralesDeLaMedidaDelJuez son los dos umbrales de la medida del juez de
// afirma_lo_no_leido, la clase que decide, tal como los fija
// contracts/informe-del-job.md §2 de H24, con los recuentos y los totales de
// una medida: los dos deciden con 0.
func umbralesDeLaMedidaDelJuez(sinMarcar, defectos, marcados, correctos int) []Umbral {
	return []Umbral{
		{
			Nombre: "medida_del_juez:afirma_lo_no_leido:defectos_sin_marcar",
			Descripcion: "Casos etiquetados como defecto que el juez no marca en afirma_lo_no_leido, sobre los casos " +
				"etiquetados como defecto de su medida versionada",
			Medida:      float64(sinMarcar),
			Total:       &defectos,
			Comparacion: "<=",
			Umbral:      0,
			Cumple:      sinMarcar == 0,
			Decide:      true,
		},
		{
			Nombre: "medida_del_juez:afirma_lo_no_leido:correctos_marcados",
			Descripcion: "Casos etiquetados como correctos que el juez marca en afirma_lo_no_leido, sobre los casos " +
				"etiquetados como correctos de su medida versionada",
			Medida:      float64(marcados),
			Total:       &correctos,
			Comparacion: "<=",
			Umbral:      0,
			Cumple:      marcados == 0,
			Decide:      true,
		},
	}
}

// umbralDeLaDuracionDelJuez es el umbral de los segundos de los votos de las
// respuestas de un modo, tal como lo fija contracts/informe-del-job.md §2 de
// H24: sin total y decidiendo con 900 s.
func umbralDeLaDuracionDelJuez(modo Modo, segundos int) Umbral {
	return Umbral{
		Nombre: "duracion_del_juez:" + string(modo),
		Descripcion: "Segundos de los votos del juez sobre las respuestas del modo " + string(modo) +
			", desde que empieza el primero hasta que termina el último",
		Medida:      float64(segundos),
		Comparacion: "<=",
		Umbral:      900,
		Cumple:      segundos <= 900,
		Decide:      true,
	}
}

// votosDeUnJuezQueDiceNo son los votos que pide la ejecución con un juez que
// dice no a todo, uno por respuesta que se juzga: las del total de cada umbral
// de las respuestas sin activar —ninguno sin juez— y, con juez, las tres del
// modelo que decide de la eval sin binario ni servidor, si la lleva
// (contracts/informe-del-job.md §1 de H24).
func votosDeUnJuezQueDiceNo(ejecucion ejecucionConUmbrales, umbrales []Umbral) int {
	votos := 0

	for _, umbral := range umbrales {
		if strings.HasPrefix(umbral.Nombre, "sin_activar:") {
			votos += *umbral.Total
		}
	}

	if ejecucion.conLaSinBinarioNiServidor && !ejecucion.sinJuez {
		votos += repeticionesConUmbrales
	}

	return votos
}

// exigirLosTreceDelContrato exige, del caso 900-s-en-los-dos-modos y de cada
// uno de TestUmbralConjunto, los trece elementos de umbrales de boe-legislacion
// de contracts/umbrales-del-informe.md §1 de H26, en su orden, cada uno con su
// nombre, su descripción y su decide, escritos a mano: deciden nueve —el
// recuento de afirma_lo_no_leido de cada modo no, y su umbral conjunto, detrás
// de los de los dos modos, sí—, todos comparan con «<=» y ninguno es de los dos
// que salieron con la lista de expresiones (FR-030 a FR-034 de H24; FR-005,
// FR-010 y FR-012 de H26).
func exigirLosTreceDelContrato(t *testing.T, leido informeLeido) {
	t.Helper()

	type elemento struct {
		nombre, descripcion string
		decide              bool
	}

	const (
		deOrden       = "Respuestas de claude-sonnet-5-5 en el modo orden "
		deHerramienta = "Respuestas de claude-sonnet-5-5 en el modo herramienta "
		deLosModos    = "Respuestas de claude-sonnet-5-5, con todos los modos juntos, "
		sinActivar    = "sin la skill activada, sobre sus respuestas medidas en las evals que la activan"
		deLasEvals    = ", sobre sus respuestas juzgadas en las evals que activan la skill"
		queMarca      = "que el juez marca en afirma_lo_no_leido con sus tres votos" + deLasEvals
		conSi         = "con sí en cuenta_su_proceso en el primer voto del juez" + deLasEvals
		deLaTanda     = ", desde que se prepara su primera sesión hasta que termina la última"
		deLosVotos    = ", desde que empieza el primero hasta que termina el último"
	)

	esperados := []elemento{
		{"sin_activar:claude-sonnet-5-5:orden", deOrden + sinActivar, true},
		{"afirma_lo_no_leido:claude-sonnet-5-5:orden", deOrden + queMarca, false},
		{"cuenta_su_proceso:claude-sonnet-5-5:orden", deOrden + conSi, false},
		{"sin_activar:claude-sonnet-5-5:herramienta", deHerramienta + sinActivar, true},
		{"afirma_lo_no_leido:claude-sonnet-5-5:herramienta", deHerramienta + queMarca, false},
		{"cuenta_su_proceso:claude-sonnet-5-5:herramienta", deHerramienta + conSi, false},
		{"afirma_lo_no_leido:claude-sonnet-5-5", deLosModos + queMarca, true},
		{
			"medida_del_juez:afirma_lo_no_leido:defectos_sin_marcar",
			"Casos etiquetados como defecto que el juez no marca en afirma_lo_no_leido, sobre los casos etiquetados como " +
				"defecto de su medida versionada",
			true,
		},
		{
			"medida_del_juez:afirma_lo_no_leido:correctos_marcados",
			"Casos etiquetados como correctos que el juez marca en afirma_lo_no_leido, sobre los casos etiquetados como " +
				"correctos de su medida versionada",
			true,
		},
		{"duracion_de_las_sesiones:orden", "Segundos de la tanda del modo orden" + deLaTanda, true},
		{"duracion_de_las_sesiones:herramienta", "Segundos de la tanda del modo herramienta" + deLaTanda, true},
		{"duracion_del_juez:orden", "Segundos de los votos del juez sobre las respuestas del modo orden" + deLosVotos, true},
		{
			"duracion_del_juez:herramienta",
			"Segundos de los votos del juez sobre las respuestas del modo herramienta" + deLosVotos, true,
		},
	}

	publicados := make([]elemento, 0, len(leido.informe.Umbrales))
	deciden := 0

	for _, umbral := range leido.informe.Umbrales {
		publicados = append(publicados, elemento{umbral.Nombre, umbral.Descripcion, umbral.Decide})

		assert.Equal(t, "<=", umbral.Comparacion, "la comparación de %s", umbral.Nombre)

		if umbral.Decide {
			deciden++
		}

		for _, retirado := range []string{"expresiones_prohibidas:", "redaccion_no_leida:"} {
			assert.False(t, strings.HasPrefix(umbral.Nombre, retirado), "%s salió con la lista de expresiones", umbral.Nombre)
		}
	}

	assert.Equal(t, esperados, publicados, "los trece, en su orden")
	assert.Equal(t, 9, deciden, "nueve de los trece deciden")
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

// exigirElUmbralEscrito exige que el umbral que nombra el ejemplo esté escrito
// en informe.json como él, que es un elemento de un contrato en una línea: con
// sus mismos valores y, lo que JSONEq no mira, con sus claves en el mismo orden.
func exigirElUmbralEscrito(t *testing.T, leido informeLeido, ejemplo string) {
	t.Helper()

	var delEjemplo Umbral

	require.NoError(t, json.Unmarshal([]byte(ejemplo), &delEjemplo))

	var escritos []jsontext.Value

	require.NoError(t, json.Unmarshal(leido.crudo.Umbrales, &escritos))
	require.Len(t, escritos, len(leido.informe.Umbrales))

	posicion := slices.IndexFunc(leido.informe.Umbrales, func(umbral Umbral) bool {
		return umbral.Nombre == delEjemplo.Nombre
	})
	require.GreaterOrEqual(t, posicion, 0, "el informe tiene el umbral %s", delEjemplo.Nombre)

	assert.JSONEq(t, ejemplo, string(escritos[posicion]))
	assert.Equal(t, compacto(t, jsontext.Value(ejemplo)), compacto(t, escritos[posicion]),
		"las claves de %s van en el orden del contrato", delEjemplo.Nombre)
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
		escribirElJuezDelInforme(t, evals, ejecucion.clasesDelJuez, ejecucion.medidaDelJuez)
	}

	if !ejecucion.sinLista {
		copiarLaListaDelRepositorio(t, evals)
	}

	series := escribirLasEvalsConUmbrales(t, copia, ejecucion)

	if ejecucion.conLaPruebaDeRed {
		escribirLaSesionDeLaPruebaDeRed(t, copia)
	}

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

			// La que termina con el mensaje del límite de uso no llegó a pedir
			// ninguna orden.
			if ejecucion.conTextos && modo == ModoOrden && ejecucion.sinMedir != modelo {
				anotarLaOrdenEnElTranscript(t, dir)
			}

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

	insertarAntesDeLaRespuesta(t, dir, llamadaSinteticaAlArticulo21)

	sesion, err := LeerSesion(dir)
	require.NoError(t, err)
	require.Len(t, sesion.Llamadas, 1, "la sesión de %s lee el bloque con una llamada", dir)
}

// anotarLaOrdenEnElTranscript deja en el transcript de la sesión del modo orden
// del directorio, delante del mensaje con la respuesta, la orden de Bash con la
// que lee el bloque y lo que devolvió: los textos de sus herramientas, que su
// traza no da (contracts/juez-y-voto.md §2 de H24). El juicio sin modelo de la
// sesión no cambia con ellos.
func anotarLaOrdenEnElTranscript(t *testing.T, dir string) {
	t.Helper()

	insertarAntesDeLaRespuesta(t, dir,
		mensajeDeLlamadas(t, usoDeHerramienta{id: "toolu_sintetico_orden", nombre: "Bash", entrada: entradaDeBash(t, ordenDelArticulo)})+
			mensajeDeResultados(t, resultadoDeHerramienta{id: "toolu_sintetico_orden", contenido: cadenaJSON(t, sobreDelArticulo)}))

	sesion, err := LeerSesion(dir)
	require.NoError(t, err)
	require.Equal(t, []Texto{{Orden: ordenDelArticulo, Salida: sobreDelArticulo}}, sesion.Textos,
		"la sesión de %s tiene el texto de su orden", dir)
}

// insertarAntesDeLaRespuesta mete las líneas dadas en el transcript de la
// sesión del directorio delante de su último mensaje del asistente, que es el
// que lleva la respuesta.
func insertarAntesDeLaRespuesta(t *testing.T, dir, mensajes string) {
	t.Helper()

	lineas := strings.SplitAfter(contenidoDeLaSesion(t, dir, "sesion.jsonl"), "\n")

	conLaRespuesta := -1

	for posicion, linea := range lineas {
		if strings.HasPrefix(linea, principioDelMensajeDelAsistente) {
			conLaRespuesta = posicion
		}
	}

	require.GreaterOrEqual(t, conLaRespuesta, 0, "el transcript de %s tiene el mensaje con la respuesta", dir)
	escribirEnLaCopia(t, dir, "sesion.jsonl", strings.Join(slices.Insert(lineas, conLaRespuesta, mensajes), ""))
}

// escribirLaSesionDeLaPruebaDeRed escribe en la copia la sesión de la prueba de
// red de una ejecución sintética: la del caso de fuera de lo grabado, con el
// nombre que le da el plan, con modeloSonnet55 y con la primera eval sintética
// como la suya. Su pregunta no es la de la eval, así que no es de ninguna serie
// que pida el plan (contrato job-de-evals §6).
func escribirLaSesionDeLaPruebaDeRed(t *testing.T, copia string) {
	t.Helper()

	dir := filepath.Join(copia, "sesiones", nombreDeSesion(ficheroSintetico(1), modeloSonnet55, ModoOrden, 1, true))

	copiarSesionConOtroModelo(t, filepath.Join(casosDeInforme, "fuera-de-lo-grabado-no-cambia-el-veredicto", "sesiones",
		sesionDeLaPruebaDeRed), dir, modeloSonnet55)
	escribirEnLaCopia(t, dir, "eval.txt", ficheroSintetico(1)+"\n")
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

// Las evals de TestUmbralesDeLasLecturas, con la pregunta de la sesión del
// art. 21 del caso aprobado, que es la de sus sesiones
// (contracts/juicio-e-informe.md §4 y §5 de H26.1): la que pide el art. 21 con un
// comando de bloque y lo deja en su sin_servir, con la cita del art. 22, que un
// bloque de sin_servir no puede ser; la que deja sin servir el art. 22, que
// ningún comando suyo pide; y la que pide y cita el art. 21 sin sin_servir. Y la
// cita del art. 22, que se antepone a la respuesta de cada sesión de la primera.
const (
	evalConElComandoSinServir = "# El comando de bloque pide la lectura que la caché no sirve.\n" +
		"pregunta: \"¿qué dice el art. 21 de la Ley 39/2015?\"\n" +
		"activa: true\n" +
		"comandos:\n  - applet: boe\n    norma: BOE-A-2015-10565\n    bloque: a21\n" +
		"citas:\n  - norma: BOE-A-2015-10565\n    bloque: a22\n" +
		"sin_servir:\n  - norma: BOE-A-2015-10565\n    bloque: a21\n"
	evalConOtroBloqueSinServir = "# La caché no sirve un bloque que ningún comando pide.\n" +
		"pregunta: \"¿qué dice el art. 21 de la Ley 39/2015?\"\n" +
		"activa: true\n" +
		"comandos:\n  - applet: boe\n    norma: BOE-A-2015-10565\n    bloque: a21\n" +
		"citas:\n  - norma: BOE-A-2015-10565\n    bloque: a21\n" +
		"sin_servir:\n  - norma: BOE-A-2015-10565\n    bloque: a22\n"
	evalServida = "# El bloque se sirve.\n" +
		"pregunta: \"¿qué dice el art. 21 de la Ley 39/2015?\"\n" +
		"activa: true\n" +
		"comandos:\n  - applet: boe\n    norma: BOE-A-2015-10565\n    bloque: a21\n" +
		"citas:\n  - norma: BOE-A-2015-10565\n    bloque: a21\n"

	citaDelArticulo22 = "Lo completa el art. 22 [BOE-A-2015-10565, bloque a22].\n\n"
)

// El umbral lectura_sin_intentar de las sesiones del modelo que decide en el modo
// orden, como lo escribe informe.json, el ejemplo de
// contracts/juicio-e-informe.md §4 de H26.1; y lo que su motivo pone delante de
// la sesión que nombra cuando una no intenta la lectura.
const (
	umbralDeLecturaSinIntentarDelContrato = `{"nombre": "lectura_sin_intentar:claude-sonnet-5-5:orden",
 "descripcion": "Sesiones de claude-sonnet-5-5 en el modo orden que no intentaron la lectura que un comando esperado de su eval pide y su sin_servir no sirve, sobre sus sesiones medidas de las evals con un comando así",
 "medida": 0, "total": 3, "comparacion": "<=", "umbral": 0, "cumple": true, "decide": true}`

	motivoDeUnaSinIntentar = "umbral lectura_sin_intentar:claude-sonnet-5-5:orden: 1 de 3 (33,3 %), y tiene que " +
		"ser ≤ 0,0 %: "
)

// ejecucionDeLasLecturas es una ejecución sintética del job para
// TestUmbralesDeLasLecturas: la eval 01, con el contenido dado, y la 02, la del
// art. 21 servida; tres repeticiones de cada una con el modelo que decide y con
// el informativo, en el modo orden y, con dosModos, también en el modo
// herramienta. Cada sesión es una copia de la del art. 21 del caso aprobado, que
// lee el bloque; las de sinLeer, del modo orden y por su nombre, no lo leen, y
// las de cortadas, además, las cortó el tope tras su respuesta sin la skill
// activada.
type ejecucionDeLasLecturas struct {
	eval01   string
	dosModos bool
	sinLeer  []string
	cortadas []string
}

// TestUmbralesDeLasLecturas fija la serie y el umbral de las lecturas que una
// eval deja sin servir (contracts/juicio-e-informe.md §4 y §5 de H26.1;
// data-model §6 y §7 de H26.1; research D11 y D12 de H26.1; FR-015, FR-033 de
// H26.1; SC-005, SC-008, SC-013; US2, escenario 5), con sesiones sintéticas e
// informe escritos en el directorio temporal del test:
//
//   - con una eval con un comando de bloque en su sin_servir, el informe lleva,
//     detrás de los umbrales de hoy —ninguno en esta skill sin juez—, por cada
//     modo del plan y, en cada uno, por cada modelo, el que decide delante,
//     lectura_sin_intentar:<modelo>:<modo>, sobre las sesiones medidas de esa
//     eval, también con medida 0; decide solo el del modelo que decide;
//   - la sesión que no intenta la lectura cuenta, aunque el tope la cortara sin
//     la skill activada, y el motivo del umbral la nombra; la del modelo
//     informativo cuenta en el suyo, que no decide ni cambia el veredicto;
//   - el informe de una skill sin una eval así —la que deja sin servir un
//     bloque que ningún comando pide— no lo lleva;
//   - y la serie de una eval con sin_servir lleva en
//     sesiones_con_lectura_sin_servir cuántas de sus sesiones leídas tienen
//     alguna invocación en lecturas_sin_servir, también 0, y la de otra eval,
//     null.
func TestUmbralesDeLasLecturas(t *testing.T) {
	t.Parallel()

	t.Run("intentan-la-lectura", probarLasQueIntentanLaLectura)
	t.Run("una-no-la-intenta", probarUnaQueNoIntentaLaLectura)
	t.Run("el-informativo-no-la-intenta", probarElInformativoQueNoIntentaLaLectura)
	t.Run("ninguna-la-intenta", probarNingunaQueIntentaLaLectura)
	t.Run("en-los-dos-modos", probarLasLecturasEnLosDosModos)
	t.Run("sin-una-eval-asi", probarLaSkillSinUnaEvalAsi)
}

// probarLasQueIntentanLaLectura: las tres sesiones de cada modelo intentan la
// lectura, y los dos umbrales, con medida 0 sobre 3, se cumplen; el del modelo
// que decide, escrito como el ejemplo del contrato, y sus series con 3.
func probarLasQueIntentanLaLectura(t *testing.T) {
	t.Parallel()

	leido := escribirEjecucionDeLasLecturas(t, ejecucionDeLasLecturas{eval01: evalConElComandoSinServir})

	exigirUmbrales(t, leido, []Umbral{
		umbralDeLecturaSinIntentar(modeloSonnet55, ModoOrden, 0, true),
		umbralDeLecturaSinIntentar(modeloHaiku45, ModoOrden, 0, false),
	})
	exigirElUmbralEscrito(t, leido, umbralDeLecturaSinIntentarDelContrato)
	exigirLosInvariantesDeLosUmbrales(t, leido)
	exigirMotivosDeLaRaiz(t, leido)
	assert.Equal(t, VeredictoAprobado, leido.informe.Veredicto)

	exigirLasSesionesConLecturaSinServir(t, leido, map[string]string{
		modeloSonnet55: "3", modeloHaiku45: "3",
	})
	exigirLasFilasDeLasLecturas(t, leido)
}

// probarUnaQueNoIntentaLaLectura: la segunda sesión del modelo que decide no lee
// el bloque, y el tope la cortó tras su respuesta sin la skill activada: no es
// una respuesta medida, y es una sesión medida (research D11 de H26.1). Cuenta
// en el umbral, 1 de 3, cuyo motivo la nombra, y el veredicto es fallo aunque su
// serie, con las otras dos, pase.
func probarUnaQueNoIntentaLaLectura(t *testing.T) {
	t.Parallel()

	sinLeer := sesionSintetica(1, modeloSonnet55, 2)
	leido := escribirEjecucionDeLasLecturas(t, ejecucionDeLasLecturas{
		eval01: evalConElComandoSinServir, sinLeer: []string{sinLeer}, cortadas: []string{sinLeer},
	})

	exigirUmbrales(t, leido, []Umbral{
		umbralDeLecturaSinIntentar(modeloSonnet55, ModoOrden, 1, true),
		umbralDeLecturaSinIntentar(modeloHaiku45, ModoOrden, 0, false),
	})
	exigirLosInvariantesDeLosUmbrales(t, leido)
	exigirMotivosDeLaRaiz(t, leido, motivoDeUnaSinIntentar+sinLeer)
	assert.Equal(t, VeredictoFallo, leido.informe.Veredicto)

	resultado := resultadoDeLaSesion(t, leido.informe, sinLeer)
	assert.False(t, resultado.SesionTerminada, "el tope cortó %s", sinLeer)
	assert.False(t, resultado.Activada, "%s no activa la skill", sinLeer)
	assert.Empty(t, resultado.LecturasSinServir, "%s no intenta la lectura", sinLeer)
	assert.True(t, tasaDeLaSerie(t, leido.informe, ficheroSintetico(1), modeloSonnet55).Pasa,
		"la serie de %s pasa con las otras dos", sinLeer)

	exigirLasSesionesConLecturaSinServir(t, leido, map[string]string{
		modeloSonnet55: "2", modeloHaiku45: "3",
	})
	exigirLasFilasDeLasLecturas(t, leido)
}

// probarElInformativoQueNoIntentaLaLectura: la segunda sesión del modelo
// informativo no lee el bloque y cuenta en su umbral, 1 de 3, que no se cumple,
// no decide y deja el veredicto como estaba, sin ningún motivo.
func probarElInformativoQueNoIntentaLaLectura(t *testing.T) {
	t.Parallel()

	leido := escribirEjecucionDeLasLecturas(t, ejecucionDeLasLecturas{
		eval01: evalConElComandoSinServir, sinLeer: []string{sesionSintetica(1, modeloHaiku45, 2)},
	})

	exigirUmbrales(t, leido, []Umbral{
		umbralDeLecturaSinIntentar(modeloSonnet55, ModoOrden, 0, true),
		umbralDeLecturaSinIntentar(modeloHaiku45, ModoOrden, 1, false),
	})
	exigirLosInvariantesDeLosUmbrales(t, leido)
	exigirMotivosDeLaRaiz(t, leido)
	assert.Equal(t, VeredictoAprobado, leido.informe.Veredicto)

	exigirLasSesionesConLecturaSinServir(t, leido, map[string]string{
		modeloSonnet55: "3", modeloHaiku45: "2",
	})
	exigirLasFilasDeLasLecturas(t, leido)
}

// probarNingunaQueIntentaLaLectura: ninguna sesión del modelo que decide lee el
// bloque; las tres cuentan, el motivo del umbral las nombra en orden de sesión
// separadas por «; », y su serie lleva 0, no null.
func probarNingunaQueIntentaLaLectura(t *testing.T) {
	t.Parallel()

	sinLeer := []string{
		sesionSintetica(1, modeloSonnet55, 1), sesionSintetica(1, modeloSonnet55, 2), sesionSintetica(1, modeloSonnet55, 3),
	}
	leido := escribirEjecucionDeLasLecturas(t, ejecucionDeLasLecturas{eval01: evalConElComandoSinServir, sinLeer: sinLeer})

	exigirUmbrales(t, leido, []Umbral{
		umbralDeLecturaSinIntentar(modeloSonnet55, ModoOrden, 3, true),
		umbralDeLecturaSinIntentar(modeloHaiku45, ModoOrden, 0, false),
	})
	exigirLosInvariantesDeLosUmbrales(t, leido)
	assert.Contains(t, leido.informe.Motivos, "umbral lectura_sin_intentar:claude-sonnet-5-5:orden: 3 de 3 (100,0 %), "+
		"y tiene que ser ≤ 0,0 %: "+strings.Join(sinLeer, "; "))
	assert.Equal(t, VeredictoFallo, leido.informe.Veredicto)

	exigirLasSesionesConLecturaSinServir(t, leido, map[string]string{
		modeloSonnet55: "0", modeloHaiku45: "3",
	})
	exigirLasFilasDeLasLecturas(t, leido)
}

// probarLasLecturasEnLosDosModos: con los dos modos, los cuatro umbrales van por
// modo, el del modo orden delante, y en cada modo, el del modelo que decide
// delante del informativo; las sesiones del modo herramienta intentan la lectura
// con su llamada a boe_articulo, y sus series también llevan 3.
func probarLasLecturasEnLosDosModos(t *testing.T) {
	t.Parallel()

	leido := escribirEjecucionDeLasLecturas(t, ejecucionDeLasLecturas{eval01: evalConElComandoSinServir, dosModos: true})

	exigirUmbrales(t, leido, []Umbral{
		umbralDeLecturaSinIntentar(modeloSonnet55, ModoOrden, 0, true),
		umbralDeLecturaSinIntentar(modeloHaiku45, ModoOrden, 0, false),
		umbralDeLecturaSinIntentar(modeloSonnet55, ModoHerramienta, 0, true),
		umbralDeLecturaSinIntentar(modeloHaiku45, ModoHerramienta, 0, false),
	})
	exigirLosInvariantesDeLosUmbrales(t, leido)
	exigirMotivosDeLaRaiz(t, leido)

	exigirLasSesionesConLecturaSinServir(t, leido, map[string]string{
		modeloSonnet55: "3", modeloHaiku45: "3",
		modeloDeLaSerie(modeloSonnet55, ModoHerramienta): "3", modeloDeLaSerie(modeloHaiku45, ModoHerramienta): "3",
	})
	exigirLasFilasDeLasLecturas(t, leido)
}

// probarLaSkillSinUnaEvalAsi: la eval 01 deja sin servir un bloque que ningún
// comando suyo pide, y su comando pide uno que se sirve, así que el informe no
// lleva ningún umbral; su serie lleva igual sus sesiones con una lectura de
// sin_servir, ninguna, porque ninguna lee ese bloque.
func probarLaSkillSinUnaEvalAsi(t *testing.T) {
	t.Parallel()

	leido := escribirEjecucionDeLasLecturas(t, ejecucionDeLasLecturas{eval01: evalConOtroBloqueSinServir})

	exigirUmbrales(t, leido, nil)
	exigirMotivosDeLaRaiz(t, leido)
	assert.Equal(t, VeredictoAprobado, leido.informe.Veredicto)

	exigirLasSesionesConLecturaSinServir(t, leido, map[string]string{
		modeloSonnet55: "0", modeloHaiku45: "0",
	})
}

// umbralDeLecturaSinIntentar es el umbral lectura_sin_intentar del modelo en el
// modo, con las sesiones que no intentaron la lectura de las tres medidas de la
// eval 01 de ejecucionDeLasLecturas.
func umbralDeLecturaSinIntentar(modelo string, modo Modo, sinIntentar int, decide bool) Umbral {
	medidas := repeticionesConUmbrales

	return Umbral{
		Nombre: "lectura_sin_intentar:" + modelo + ":" + string(modo),
		Descripcion: "Sesiones de " + modelo + " en el modo " + string(modo) + " que no intentaron la lectura que un " +
			"comando esperado de su eval pide y su sin_servir no sirve, sobre sus sesiones medidas de las evals con un " +
			"comando así",
		Medida:      float64(sinIntentar),
		Total:       &medidas,
		Comparacion: comparacionMenorOIgual,
		Umbral:      0,
		Cumple:      sinIntentar == 0,
		Decide:      decide,
	}
}

// exigirLasSesionesConLecturaSinServir exige, de cada serie de la eval 01 por el
// modelo con el que la nombra el informe, sesiones_con_lectura_sin_servir tal
// como lo escribe informe.json, y null en cada serie de la 02.
func exigirLasSesionesConLecturaSinServir(t *testing.T, leido informeLeido, deLa01 map[string]string) {
	t.Helper()

	vistas := map[string]string{}

	for _, tasa := range leido.crudo.Tasas {
		escrito := string(tasa.SesionesConLecturaSinServir)

		switch tasa.Eval {
		case ficheroSintetico(1):
			vistas[tasa.Modelo] = escrito
		case ficheroSintetico(2):
			assert.Equal(t, "null", escrito, "sesiones_con_lectura_sin_servir de la serie de %s con %s", tasa.Eval,
				tasa.Modelo)
		default:
			assert.Failf(t, "serie inesperada", "el informe tiene la serie de %s con %s", tasa.Eval, tasa.Modelo)
		}
	}

	assert.Equal(t, deLa01, vistas, "sesiones_con_lectura_sin_servir de las series de %s", ficheroSintetico(1))
}

// exigirLasFilasDeLasLecturas exige la fila de cada umbral en la tabla de los
// umbrales de informe.md, como la de los demás: con su medida, su condición, si
// se cumple y si decide, «no: solo se publica» en el del modelo informativo.
func exigirLasFilasDeLasLecturas(t *testing.T, leido informeLeido) {
	t.Helper()

	deLosUmbrales := seccionDelInforme(t, leido.md, "Umbrales")

	for _, umbral := range leido.informe.Umbrales {
		decide := "sí"
		if !umbral.Decide {
			decide = soloSePublica
		}

		exigirLineas(t, deLosUmbrales, filaDeTabla("`"+umbral.Nombre+"`", umbral.medidaEscrita(), "≤ 0,0 %",
			siONo(umbral.Cumple), decide))
	}
}

// escribirEjecucionDeLasLecturas arma la ejecución en un directorio temporal del
// test y escribe su informe con el plan del job —el modelo que decide y el
// informativo, tres repeticiones, el umbral 2 y sus modos— y sin objetivo de
// duración. La respuesta de cada sesión de la eval 01 lleva delante la cita del
// art. 22, que la primera de sus evals espera.
func escribirEjecucionDeLasLecturas(t *testing.T, ejecucion ejecucionDeLasLecturas) informeLeido {
	t.Helper()

	copia := t.TempDir()
	evals := filepath.Join(copia, "evals")
	require.NoError(t, os.Mkdir(evals, 0o750))
	escribirEnLaCopia(t, evals, ficheroSintetico(1), ejecucion.eval01)
	escribirEnLaCopia(t, evals, ficheroSintetico(2), evalServida)

	modos := []Modo{ModoOrden}
	if ejecucion.dosModos {
		modos = append(modos, ModoHerramienta)
	}

	for _, modo := range modos {
		for numero := 1; numero <= 2; numero++ {
			for _, modelo := range []string{modeloSonnet55, modeloHaiku45} {
				for vez := 1; vez <= repeticionesConUmbrales; vez++ {
					dir := escribirSesionSinteticaEn(t, copia, modo, numero, modelo, vez, false)

					if numero == 1 {
						anteponerALaRespuesta(t, dir, citaDelArticulo22)
					}
				}
			}
		}
	}

	for _, sesion := range ejecucion.sinLeer {
		quitarLaLectura(t, filepath.Join(copia, "sesiones", sesion))
	}

	for _, sesion := range ejecucion.cortadas {
		dir := filepath.Join(copia, "sesiones", sesion)
		quitarLaActivacion(t, dir)
		cortarTrasLaRespuesta(t, dir)
	}

	return informeDeLaCopia(t, copia, func(entradas *InformeAEscribir) {
		entradas.ModeloQueDecide = modeloSonnet55
		entradas.ModelosInformativos = []string{modeloHaiku45}
		entradas.Repeticiones, entradas.Umbral = repeticionesConUmbrales, umbralConUmbrales

		if ejecucion.dosModos {
			entradas.Modos = modos
		}
	})
}

// quitarLaLectura deja la sesión del art. 21 del modo orden del directorio sin
// la orden que lee el bloque: su traza ya no crea el proceso ni lo tiene.
func quitarLaLectura(t *testing.T, dir string) {
	t.Helper()

	traza := filepath.Join(dir, directorioDeLaTraza)
	deClaude := contenidoDeLaSesion(t, traza, "t.1000")
	require.Equal(t, 1, strings.Count(deClaude, creacionDeLaLectura),
		"la traza de claude de %s crea una sola vez el proceso que lee el bloque", dir)
	escribirEnLaCopia(t, traza, "t.1000", strings.Replace(deClaude, creacionDeLaLectura, "", 1))
	require.NoError(t, os.Remove(filepath.Join(traza, "t.2000")))
}
