package evals

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
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
	// redaccionNoLeida es lo que se antepone a una respuesta con una redacción
	// que ninguna orden devolvió: lleva ya no exige, la expresión de
	// redaccion_no_leida de la lista del repositorio, y ninguna otra.
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
// si se pide, la de no activación del caso aprobado, que decide. Cada sesión es
// una copia de la de su eval en el caso aprobado, que pasa, con su eval y su
// modelo.
type ejecucionConUmbrales struct {
	// queDeciden e informativas son cuántas evals de cada clase lleva.
	queDeciden, informativas int

	// conHaiku dice si modeloHaiku45 está entre los modelos informativos del job,
	// y abre solo las evals que deciden; el que decide es siempre modeloSonnet55.
	conHaiku bool

	// sinLista dice que la carpeta de evals no lleva la lista del repositorio.
	sinLista bool

	// conAlguna son, por modelo, cuántas de sus respuestas llevan delante la
	// transición de la memoria de consultas, que tiene expresiones de la lista: la
	// primera sesión de otras tantas de sus series, repartidas por igual entre
	// ellas, de modo que cada serie que decide sigue llegando al umbral con las
	// que pasan.
	conAlguna map[string]int

	// sinActivar y conRedaccionNoLeida son cuántas respuestas de modeloSonnet55
	// no activan la skill —su transcript no la carga— y cuántas llevan delante
	// redaccionNoLeida: como las de conAlguna, la primera sesión de otras tantas
	// de sus series, de modo que con una de cada es la misma sesión.
	sinActivar, conRedaccionNoLeida int

	// sinTerminar son cuántas evals informativas más lleva, detrás de las otras,
	// cuyas tres sesiones de modeloSonnet55 cortó el tope tras su respuesta, con
	// el código 124, mientras esperaban una tarea en segundo plano.
	sinTerminar int

	// conLaDeNoActivacion dice que lleva, detrás de todas, la eval de no
	// activación del caso aprobado, que decide, con sus sesiones de cada modelo,
	// que no activan la skill y pasan.
	conLaDeNoActivacion bool

	// sinMedir es el modelo cuyas sesiones copiadas de la del art. 21 terminan
	// todas con el mensaje del límite de uso; vacío, ninguno.
	sinMedir string

	// anadidoALaEval es lo que se añade al final de cada eval, y
	// prefijoDeLasRespuestas, lo que se antepone a la respuesta de cada sesión;
	// vacíos, nada.
	anadidoALaEval, prefijoDeLasRespuestas string

	// duracion y objetivo son DuracionDeLasSesiones y ObjetivoDeDuracion.
	duracion, objetivo int
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
// escritas con EscribirInforme en t.TempDir(): si la skill tiene lista, los tres
// del modelo que decide —el de las expresiones, el de las respuestas sin la skill
// activada y el de las que llevan una expresión de redaccion_no_leida—, con sus
// respuestas medidas de las evals que activan la skill —las informativas
// incluidas— como total, y detrás los de las expresiones de los informativos,
// que no deciden; y el de la duración si hay objetivo, sin total y decidiendo. Un
// umbral que decide y no se cumple pone el veredicto en fallo con su motivo,
// detrás de los de siempre y delante del de las sesiones sin medir, y el de la
// duración, con el prefijo de la ejecución, detrás de él; uno que se cumple, o que
// no decide, no cambia nada. Los valores esperados son los del contrato escritos
// a mano —0.05 y 0, no las constantes del paquete—, para que un umbral cambiado en
// el código no pase. En todos, rehacer la comparación de cada umbral da su cumple
// (exigirLosInvariantesDeLosUmbrales, desde leerInformeEscrito).
//
// Desde H7.4 (contracts/informe-del-job.md §1, §2, §3 y §6; data-model §5 y §6;
// research D13 y D14; FR-040 a FR-048, FR-061, FR-097; SC-006; US4-1 a US4-3,
// US5-2), sobre una ejecución como la del cierre —18 evals que activan la skill,
// 8 de ellas informativas: 54 respuestas del modelo que decide—, cuenta como
// respuesta medida solo la sesión terminada: seis más que el tope cortó tras su
// respuesta se publican con su motivo y no cuentan ni en la medida ni en el
// total; una sesión de una eval que no activa la skill no cuenta como sin
// activar; y la sesión sin activar y con una expresión de redaccion_no_leida
// cuenta una vez en cada umbral.
func TestUmbralesDelInforme(t *testing.T) {
	t.Parallel()

	// comoElJob es la ejecución del cierre de boe-legislacion desde H7.4: 18
	// evals que activan la skill, 8 de ellas informativas, con Sonnet 5.5 y Haiku
	// 4.5 (54 y 30 respuestas); minima, con una de cada clase (6 y 3).
	comoElJob := ejecucionConUmbrales{queDeciden: 10, informativas: 8, conHaiku: true}
	minima := ejecucionConUmbrales{queDeciden: 1, informativas: 1, conHaiku: true}

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
			umbrales: []Umbral{
				umbralDeExpresiones(modeloSonnet55, 0, 54, true, true),
				umbralSinActivar(modeloSonnet55, 1, 54, false),
				umbralDeRedaccionNoLeida(modeloSonnet55, 0, 54, true),
				umbralDeExpresiones(modeloHaiku45, 0, 30, true, false),
			},
			motivos:   []string{"umbral sin_activar:claude-sonnet-5-5: 1 de 54 (1,9 %), y tiene que ser ≤ 0,0 %"},
			veredicto: VeredictoFallo,
		},
		{
			// Las sesiones de la eval de no activación, que no activan la skill como
			// ella espera, no cuentan: ni como sin activar ni en el total.
			nombre:    "sin-activar-ninguna",
			ejecucion: conCambios(comoElJob, func(e *ejecucionConUmbrales) { e.conLaDeNoActivacion = true }),
			umbrales: []Umbral{
				umbralDeExpresiones(modeloSonnet55, 0, 54, true, true),
				umbralSinActivar(modeloSonnet55, 0, 54, true),
				umbralDeRedaccionNoLeida(modeloSonnet55, 0, 54, true),
				umbralDeExpresiones(modeloHaiku45, 0, 30, true, false),
			},
			veredicto: VeredictoAprobado,
			exigir:    exigirLaDeNoActivacionSinContar,
		},
		{
			nombre:    "redaccion-no-leida-una",
			ejecucion: conCambios(comoElJob, func(e *ejecucionConUmbrales) { e.conRedaccionNoLeida = 1 }),
			umbrales: []Umbral{
				umbralDeExpresiones(modeloSonnet55, 1, 54, true, true),
				umbralSinActivar(modeloSonnet55, 0, 54, true),
				umbralDeRedaccionNoLeida(modeloSonnet55, 1, 54, false),
				umbralDeExpresiones(modeloHaiku45, 0, 30, true, false),
			},
			motivos:   []string{"umbral redaccion_no_leida:claude-sonnet-5-5: 1 de 54 (1,9 %), y tiene que ser ≤ 0,0 %"},
			veredicto: VeredictoFallo,
		},
		{
			// Una expresión de otra familia cuenta en el de las expresiones y no en
			// el de redaccion_no_leida.
			nombre: "redaccion-no-leida-ninguna",
			ejecucion: conCambios(comoElJob, func(e *ejecucionConUmbrales) {
				e.conAlguna = map[string]int{modeloSonnet55: 1}
			}),
			umbrales: []Umbral{
				umbralDeExpresiones(modeloSonnet55, 1, 54, true, true),
				umbralSinActivar(modeloSonnet55, 0, 54, true),
				umbralDeRedaccionNoLeida(modeloSonnet55, 0, 54, true),
				umbralDeExpresiones(modeloHaiku45, 0, 30, true, false),
			},
			veredicto: VeredictoAprobado,
		},
		{
			nombre: "tres-de-54",
			ejecucion: conCambios(comoElJob, func(e *ejecucionConUmbrales) {
				e.conAlguna = map[string]int{modeloSonnet55: 3}
			}),
			umbrales: []Umbral{
				umbralDeExpresiones(modeloSonnet55, 3, 54, false, true),
				umbralSinActivar(modeloSonnet55, 0, 54, true),
				umbralDeRedaccionNoLeida(modeloSonnet55, 0, 54, true),
				umbralDeExpresiones(modeloHaiku45, 0, 30, true, false),
			},
			motivos:   []string{"umbral expresiones_prohibidas:claude-sonnet-5-5: 3 de 54 (5,6 %), y tiene que ser ≤ 5,0 %"},
			veredicto: VeredictoFallo,
		},
		{
			nombre: "dos-de-54",
			ejecucion: conCambios(comoElJob, func(e *ejecucionConUmbrales) {
				e.conAlguna = map[string]int{modeloSonnet55: 2}
			}),
			umbrales: []Umbral{
				umbralDeExpresiones(modeloSonnet55, 2, 54, true, true),
				umbralSinActivar(modeloSonnet55, 0, 54, true),
				umbralDeRedaccionNoLeida(modeloSonnet55, 0, 54, true),
				umbralDeExpresiones(modeloHaiku45, 0, 30, true, false),
			},
			veredicto: VeredictoAprobado,
		},
		{
			// Con las seis sin terminar en el total, 3 de 60 se cumpliría.
			nombre: "tres-de-54-y-seis-sin-terminar",
			ejecucion: conCambios(comoElJob, func(e *ejecucionConUmbrales) {
				e.conAlguna, e.sinTerminar = map[string]int{modeloSonnet55: 3}, 2
			}),
			umbrales: []Umbral{
				umbralDeExpresiones(modeloSonnet55, 3, 54, false, true),
				umbralSinActivar(modeloSonnet55, 0, 54, true),
				umbralDeRedaccionNoLeida(modeloSonnet55, 0, 54, true),
				umbralDeExpresiones(modeloHaiku45, 0, 30, true, false),
			},
			motivos:   []string{"umbral expresiones_prohibidas:claude-sonnet-5-5: 3 de 54 (5,6 %), y tiene que ser ≤ 5,0 %"},
			veredicto: VeredictoFallo,
			exigir:    exigirLasSeisSinTerminar,
		},
		{
			nombre: "una-sin-activar-con-expresion",
			ejecucion: conCambios(comoElJob, func(e *ejecucionConUmbrales) {
				e.sinActivar, e.conRedaccionNoLeida = 1, 1
			}),
			umbrales: []Umbral{
				umbralDeExpresiones(modeloSonnet55, 1, 54, true, true),
				umbralSinActivar(modeloSonnet55, 1, 54, false),
				umbralDeRedaccionNoLeida(modeloSonnet55, 1, 54, false),
				umbralDeExpresiones(modeloHaiku45, 0, 30, true, false),
			},
			motivos: []string{
				"umbral sin_activar:claude-sonnet-5-5: 1 de 54 (1,9 %), y tiene que ser ≤ 0,0 %",
				"umbral redaccion_no_leida:claude-sonnet-5-5: 1 de 54 (1,9 %), y tiene que ser ≤ 0,0 %",
			},
			veredicto: VeredictoFallo,
			exigir:    exigirLaMismaSesionSinActivarConExpresion,
		},
		{
			nombre: "dos-de-30-de-haiku-4-5",
			ejecucion: conCambios(comoElJob, func(e *ejecucionConUmbrales) {
				e.conAlguna = map[string]int{modeloHaiku45: 2}
			}),
			umbrales: []Umbral{
				umbralDeExpresiones(modeloSonnet55, 0, 54, true, true),
				umbralSinActivar(modeloSonnet55, 0, 54, true),
				umbralDeRedaccionNoLeida(modeloSonnet55, 0, 54, true),
				umbralDeExpresiones(modeloHaiku45, 2, 30, false, false),
			},
			veredicto: VeredictoAprobado,
		},
		{
			// Sin lista, la respuesta con expresiones pasa; y sin objetivo, ninguna
			// duración es un umbral: ni siquiera el de las respuestas sin activar,
			// como en legal-core (FR-048).
			nombre: "sin-lista-ni-objetivo",
			ejecucion: conCambios(minima, func(e *ejecucionConUmbrales) {
				e.sinLista, e.conAlguna, e.sinActivar, e.duracion = true, map[string]int{modeloSonnet55: 1}, 1, 2000
			}),
			umbrales:  []Umbral{},
			veredicto: VeredictoAprobado,
		},
		{
			nombre:    "901-s-con-objetivo-900",
			ejecucion: conCambios(minima, func(e *ejecucionConUmbrales) { e.duracion, e.objetivo = 901, 900 }),
			umbrales: []Umbral{
				umbralDeExpresiones(modeloSonnet55, 0, 6, true, true),
				umbralSinActivar(modeloSonnet55, 0, 6, true),
				umbralDeRedaccionNoLeida(modeloSonnet55, 0, 6, true),
				umbralDeExpresiones(modeloHaiku45, 0, 3, true, false),
				umbralDeDuracion(901, 900, false),
			},
			motivos:   []string{"de la ejecución, no de la skill: duracion_de_las_sesiones: 901 s, y tiene que ser ≤ 900 s"},
			veredicto: VeredictoFallo,
		},
		{
			nombre:    "900-s-con-objetivo-900",
			ejecucion: conCambios(minima, func(e *ejecucionConUmbrales) { e.duracion, e.objetivo = 900, 900 }),
			umbrales: []Umbral{
				umbralDeExpresiones(modeloSonnet55, 0, 6, true, true),
				umbralSinActivar(modeloSonnet55, 0, 6, true),
				umbralDeRedaccionNoLeida(modeloSonnet55, 0, 6, true),
				umbralDeExpresiones(modeloHaiku45, 0, 3, true, false),
				umbralDeDuracion(900, 900, true),
			},
			veredicto: VeredictoAprobado,
		},
		{
			// Sin ninguna respuesta medida, las proporciones son 0 y los umbrales se
			// cumplen; el veredicto ya es fallo por las sesiones sin medir, y sus
			// series, sin medir, no dan el motivo de su tasa.
			nombre:    "todas-las-de-sonnet-5-5-sin-medir",
			ejecucion: conCambios(minima, func(e *ejecucionConUmbrales) { e.sinMedir = modeloSonnet55 }),
			umbrales: []Umbral{
				umbralDeExpresiones(modeloSonnet55, 0, 0, true, true),
				umbralSinActivar(modeloSonnet55, 0, 0, true),
				umbralDeRedaccionNoLeida(modeloSonnet55, 0, 0, true),
				umbralDeExpresiones(modeloHaiku45, 0, 3, true, false),
			},
			motivos:   []string{motivoEsperadoDelLimite(sesionesSinMedirDe(modeloSonnet55, 2)...)},
			veredicto: VeredictoFallo,
		},
		{
			// Los tres motivos de H7.3 a la vez, en su orden: el del umbral que
			// decide, el de las sesiones sin medir —las de Haiku 4.5, cuyo umbral
			// queda en 0 de 0— y el de la duración.
			nombre: "los-tres-motivos-en-su-orden",
			ejecucion: conCambios(minima, func(e *ejecucionConUmbrales) {
				e.conAlguna, e.sinMedir, e.duracion, e.objetivo = map[string]int{modeloSonnet55: 1}, modeloHaiku45, 901, 900
			}),
			umbrales: []Umbral{
				umbralDeExpresiones(modeloSonnet55, 1, 6, false, true),
				umbralSinActivar(modeloSonnet55, 0, 6, true),
				umbralDeRedaccionNoLeida(modeloSonnet55, 0, 6, true),
				umbralDeExpresiones(modeloHaiku45, 0, 0, true, false),
				umbralDeDuracion(901, 900, false),
			},
			motivos: []string{
				"umbral expresiones_prohibidas:claude-sonnet-5-5: 1 de 6 (16,7 %), y tiene que ser ≤ 5,0 %",
				motivoEsperadoDelLimite(sesionesSinMedirDe(modeloHaiku45, 1)...),
				"de la ejecución, no de la skill: duracion_de_las_sesiones: 901 s, y tiene que ser ≤ 900 s",
			},
			veredicto: VeredictoFallo,
		},
	}

	for _, caso := range casos {
		t.Run(caso.nombre, func(t *testing.T) {
			t.Parallel()

			leido := escribirEjecucionConUmbrales(t, caso.ejecucion)

			exigirUmbrales(t, leido, caso.umbrales)
			exigirMotivosDeLaRaiz(t, leido, caso.motivos...)
			assert.Equal(t, caso.veredicto, leido.informe.Veredicto)
			assert.Equal(t, caso.ejecucion.duracion, leido.informe.DuracionDeLasSesiones)

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

// conCambios es la ejecución dada con lo que cambie cambiar.
func conCambios(ejecucion ejecucionConUmbrales, cambiar func(e *ejecucionConUmbrales)) ejecucionConUmbrales {
	cambiar(&ejecucion)

	return ejecucion
}

// umbralDeExpresiones es el umbral de las expresiones prohibidas de un modelo tal
// como lo fija el contrato informe-del-job §1 de H7.3, con sus valores escritos a
// mano.
func umbralDeExpresiones(modelo string, conAlguna, respuestas int, cumple, decide bool) Umbral {
	return Umbral{
		Nombre: "expresiones_prohibidas:" + modelo,
		Descripcion: "Respuestas de " + modelo + " con alguna expresión prohibida, sobre sus respuestas medidas en " +
			"las evals que activan la skill",
		Medida:      float64(conAlguna),
		Total:       &respuestas,
		Comparacion: "<=",
		Umbral:      0.05,
		Cumple:      cumple,
		Decide:      decide,
	}
}

// umbralSinActivar es el umbral de las respuestas del modelo que decide sin la
// skill activada tal como lo fija contracts/informe-del-job.md §2 de H7.4, con
// sus valores escritos a mano: 0 y decidiendo.
func umbralSinActivar(modelo string, sinActivar, respuestas int, cumple bool) Umbral {
	return Umbral{
		Nombre: "sin_activar:" + modelo,
		Descripcion: "Respuestas de " + modelo + " sin la skill activada, sobre sus respuestas medidas en las evals " +
			"que la activan",
		Medida:      float64(sinActivar),
		Total:       &respuestas,
		Comparacion: "<=",
		Umbral:      0,
		Cumple:      cumple,
		Decide:      true,
	}
}

// umbralDeRedaccionNoLeida es el umbral de las respuestas del modelo que decide
// con alguna expresión de redaccion_no_leida tal como lo fija
// contracts/informe-del-job.md §2 de H7.4, con sus valores escritos a mano: 0 y
// decidiendo.
func umbralDeRedaccionNoLeida(modelo string, conRedaccionNoLeida, respuestas int, cumple bool) Umbral {
	return Umbral{
		Nombre: "redaccion_no_leida:" + modelo,
		Descripcion: "Respuestas de " + modelo + " con alguna expresión de redaccion_no_leida (una redacción que " +
			"ninguna orden devolvió), sobre sus respuestas medidas en las evals que activan la skill",
		Medida:      float64(conRedaccionNoLeida),
		Total:       &respuestas,
		Comparacion: "<=",
		Umbral:      0,
		Cumple:      cumple,
		Decide:      true,
	}
}

// exigirLaDeNoActivacionSinContar exige, del caso sin-activar-ninguna, que el
// caso no pase en vacío: las sesiones de la eval de no activación, de los dos
// modelos, no activan la skill y pasan, y el recuento no las cuenta.
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
	assert.Equal(t, []RecuentoDeExpresiones{
		{Modelo: modeloSonnet55, ConAlguna: 0, Respuestas: 54},
		{Modelo: modeloHaiku45, ConAlguna: 0, Respuestas: 30},
	}, leido.informe.ExpresionesProhibidasPorModelo)
}

// exigirLasSeisSinTerminar exige, del caso tres-de-54-y-seis-sin-terminar, que
// las seis sesiones que el tope cortó se publiquen con su respuesta y su motivo,
// sin pasar, y que expresiones_prohibidas_por_modelo no las cuente (FR-045,
// FR-061).
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
	assert.Equal(t, []RecuentoDeExpresiones{
		{Modelo: modeloSonnet55, ConAlguna: 3, Respuestas: 54},
		{Modelo: modeloHaiku45, ConAlguna: 0, Respuestas: 30},
	}, leido.informe.ExpresionesProhibidasPorModelo)
}

// exigirLaMismaSesionSinActivarConExpresion exige, del caso
// una-sin-activar-con-expresion, que la sesión sin la skill activada y la de la
// expresión de redaccion_no_leida sean la misma y la única de cada clase: la
// primera de la primera eval con el modelo que decide.
func exigirLaMismaSesionSinActivarConExpresion(t *testing.T, leido informeLeido) {
	t.Helper()

	var sinActivar, conExpresiones []string

	for _, resultado := range leido.informe.Evals {
		if resultado.Activa && !resultado.Activada {
			sinActivar = append(sinActivar, resultado.Sesion)
		}

		if len(resultado.ExpresionesProhibidas) > 0 {
			conExpresiones = append(conExpresiones, resultado.Sesion)
		}
	}

	misma := sesionSintetica(1, modeloSonnet55, 1)
	assert.Equal(t, []string{misma}, sinActivar)
	assert.Equal(t, []string{misma}, conExpresiones)
	assert.Equal(t, []string{"ya no exige"}, resultadoDeLaSesion(t, leido.informe, misma).ExpresionesProhibidas)
}

// umbralDeDuracion es el umbral de la duración de las sesiones tal como lo fija
// el contrato informe-del-job §1 de H7.3: sin total y decidiendo.
func umbralDeDuracion(segundos, objetivo int, cumple bool) Umbral {
	return Umbral{
		Nombre:      "duracion_de_las_sesiones",
		Descripcion: "Segundos desde que se prepara la primera sesión hasta que termina la última",
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

	copia := t.TempDir()
	evals := filepath.Join(copia, "evals")
	require.NoError(t, os.Mkdir(evals, 0o750))

	if !ejecucion.sinLista {
		copiarLaListaDelRepositorio(t, evals)
	}

	modelos := []string{modeloSonnet55}
	if ejecucion.conHaiku {
		modelos = append(modelos, modeloHaiku45)
	}

	series := escribirLasEvalsConUmbrales(t, copia, ejecucion, modelos)

	for modelo, cuantas := range ejecucion.conAlguna {
		alterarLaPrimeraDeLasSeries(t, copia, modelo, series[modelo], cuantas, func(dir string) {
			anteponerALaRespuesta(t, dir, transicionDeLaMemoria+"\n\n")
		})
	}

	alterarLaPrimeraDeLasSeries(t, copia, modeloSonnet55, series[modeloSonnet55], ejecucion.conRedaccionNoLeida,
		func(dir string) { anteponerALaRespuesta(t, dir, redaccionNoLeida+"\n\n") })
	alterarLaPrimeraDeLasSeries(t, copia, modeloSonnet55, series[modeloSonnet55], ejecucion.sinActivar,
		func(dir string) { quitarLaActivacion(t, dir) })

	return informeDeLaCopia(t, copia, func(entradas *InformeAEscribir) {
		entradas.ModeloQueDecide = modeloSonnet55
		entradas.ModelosInformativos = modelos[1:]
		entradas.Repeticiones, entradas.Umbral = repeticionesConUmbrales, umbralConUmbrales
		entradas.DuracionDeLasSesiones, entradas.ObjetivoDeDuracion = ejecucion.duracion, ejecucion.objetivo
	})
}

// escribirLasEvalsConUmbrales escribe en la copia las evals de la ejecución y
// sus sesiones con cada modelo —el que decide abre todas, y los informativos,
// solo las que deciden— y devuelve, por modelo, los números de las evals de sus
// series con respuestas medidas, en su orden: sin las sin terminar ni la de no
// activación.
func escribirLasEvalsConUmbrales(
	t *testing.T, copia string, ejecucion ejecucionConUmbrales, modelos []string,
) map[string][]int {
	t.Helper()

	evals := filepath.Join(copia, "evals")
	delCasoAprobado := filepath.Join(casosDeInforme, casoAprobado)
	eval := contenidoDeLaSesion(t, filepath.Join(delCasoAprobado, "evals"), ficheroDeLaEval01)

	series := map[string][]int{}
	medidas := ejecucion.queDeciden + ejecucion.informativas

	for numero := 1; numero <= medidas+ejecucion.sinTerminar; numero++ {
		informativa := numero > ejecucion.queDeciden

		contenido := eval + ejecucion.anadidoALaEval
		if informativa {
			contenido += "informativa: true\n"
		}

		escribirEnLaCopia(t, evals, ficheroSintetico(numero), contenido)

		for _, modelo := range modelos {
			if informativa && modelo != modeloSonnet55 {
				continue
			}

			if numero <= medidas {
				series[modelo] = append(series[modelo], numero)
			}

			for vez := 1; vez <= repeticionesConUmbrales; vez++ {
				escribirSesionSintetica(t, copia, numero, modelo, vez, ejecucion.sinMedir == modelo)

				dir := filepath.Join(copia, "sesiones", sesionSintetica(numero, modelo, vez))
				if ejecucion.prefijoDeLasRespuestas != "" {
					anteponerALaRespuesta(t, dir, ejecucion.prefijoDeLasRespuestas)
				}

				if numero > medidas {
					cortarTrasLaRespuesta(t, dir)
				}
			}
		}
	}

	if ejecucion.conLaDeNoActivacion {
		numero := medidas + ejecucion.sinTerminar + 1
		escribirEnLaCopia(t, evals, ficheroSintetico(numero),
			contenidoDeLaSesion(t, filepath.Join(delCasoAprobado, "evals"), ficheroDeNoActivacion))

		for _, modelo := range modelos {
			for vez := 1; vez <= repeticionesConUmbrales; vez++ {
				escribirCopiaDeLaSesion(t, copia, sesionDeNoActivacion, numero, modelo, vez)
			}
		}
	}

	return series
}

// alterarLaPrimeraDeLasSeries aplica alterar al directorio de la primera sesión
// de cuantas de las series del modelo dadas, repartidas por igual entre ellas,
// de modo que cada serie que decide sigue llegando al umbral con las que pasan.
func alterarLaPrimeraDeLasSeries(t *testing.T, copia, modelo string, series []int, cuantas int, alterar func(dir string)) {
	t.Helper()

	require.LessOrEqual(t, cuantas, len(series), "hay una serie de %s por respuesta alterada", modelo)

	for k := range cuantas {
		alterar(filepath.Join(copia, "sesiones", sesionSintetica(series[k*len(series)/cuantas], modelo, 1)))
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

// escribirSesionSintetica escribe la sesión de esa repetición de la eval de ese
// número con el modelo dado: una copia de la del art. 21 del caso aprobado
// (escribirCopiaDeLaSesion) y, si sinMedir, terminada con el mensaje del límite
// de uso.
func escribirSesionSintetica(t *testing.T, copia string, numero int, modelo string, vez int, sinMedir bool) {
	t.Helper()

	escribirCopiaDeLaSesion(t, copia, sesionDelArticulo21, numero, modelo, vez)

	if sinMedir {
		cambiarElFinalDeLaSesion(t, filepath.Join(copia, "sesiones", sesionSintetica(numero, modelo, vez)), 1,
			mensajeResultConError(t, textoDelLimiteDeSesion))
	}
}

// escribirCopiaDeLaSesion escribe la sesión de esa repetición de la eval de ese
// número con el modelo dado: una copia de la sesión dada del caso aprobado que lo
// pide y lo declara, con eval.txt nombrando su eval.
func escribirCopiaDeLaSesion(t *testing.T, copia, delCasoAprobado string, numero int, modelo string, vez int) {
	t.Helper()

	dir := filepath.Join(copia, "sesiones", sesionSintetica(numero, modelo, vez))
	copiarSesionConOtroModelo(t, filepath.Join(casosDeInforme, casoAprobado, "sesiones", delCasoAprobado), dir, modelo)
	escribirEnLaCopia(t, dir, "eval.txt", ficheroSintetico(numero)+"\n")
}

// ficheroSintetico es el fichero de la eval de ese número de una ejecución de
// ejecucionConUmbrales.
func ficheroSintetico(numero int) string {
	return fmt.Sprintf("%02d-sintetica.yaml", numero)
}

// sesionSintetica es el nombre de la sesión de esa repetición de la eval de ese
// número con el modelo dado, como los nombra el plan.
func sesionSintetica(numero int, modelo string, vez int) string {
	return fmt.Sprintf("%02d-sintetica-%s-%02d", numero, modelo, vez)
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
