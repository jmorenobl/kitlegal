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
// de H7.3): el que decide y el que solo se publica, con los que
// TestUmbralesDelInforme y TestInformeMarkdownDeLosUmbrales arman sus
// ejecuciones.
const (
	modeloSonnet5 = "claude-sonnet-5"
	modeloHaiku45 = "claude-haiku-4-5-20251001"
)

// formaDelNombreDeUmbral es la del nombre de un umbral del contrato del ADR 0029.
var formaDelNombreDeUmbral = regexp.MustCompile(`^[a-z0-9_.:-]+$`)

// ejecucionConUmbrales es una ejecución sintética del job de evals, armada en un
// directorio temporal del test por escribirEjecucionConUmbrales con tres
// repeticiones por serie y umbral 2, como el job (ADR 0016). Sus evals son copias
// de la del art. 21 del caso aprobado, que activa la skill: primero las que
// deciden y después las informativas; cada sesión es una copia de la del art. 21
// del caso aprobado, que pasa, con su eval y su modelo.
type ejecucionConUmbrales struct {
	// queDeciden e informativas son cuántas evals de cada clase lleva.
	queDeciden, informativas int

	// conHaiku dice si modeloHaiku45 está entre los modelos informativos del job,
	// y abre solo las evals que deciden; el que decide es siempre modeloSonnet5.
	conHaiku bool

	// sinLista dice que la carpeta de evals no lleva la lista del repositorio.
	sinLista bool

	// conAlguna son, por modelo, cuántas de sus respuestas llevan delante la
	// transición de la memoria de consultas, que tiene expresiones de la lista: la
	// primera sesión de otras tantas de sus series, repartidas por igual entre
	// ellas, de modo que cada serie que decide sigue llegando al umbral con las
	// que pasan.
	conAlguna map[string]int

	// sinMedir es el modelo cuyas sesiones terminan todas con el mensaje del
	// límite de uso; vacío, ninguno.
	sinMedir string

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
// escritas con EscribirInforme en t.TempDir(): uno de expresiones por modelo del
// job si la skill tiene lista, con las respuestas medidas de las evals que activan
// la skill —las informativas incluidas— y decide solo el del modelo que decide; y
// el de la duración si hay objetivo, sin total y decidiendo. Un umbral que decide
// y no se cumple pone el veredicto en fallo con su motivo, detrás de los de
// siempre y delante del de las sesiones sin medir, y el de la duración, con el
// prefijo de la ejecución, detrás de él; uno que se cumple, o que no decide, no
// cambia nada. Los valores esperados son los del contrato escritos a mano —0.05,
// no la constante del paquete—, para que un umbral cambiado en el código no pase.
// En todos, rehacer la comparación de cada umbral da su cumple
// (exigirLosInvariantesDeLosUmbrales, desde leerInformeEscrito).
func TestUmbralesDelInforme(t *testing.T) {
	t.Parallel()

	// comoElJob es la ejecución del cierre de boe-legislacion: 17 evals que
	// activan la skill, 7 de ellas informativas, con Sonnet 5 y Haiku 4.5 (51 y
	// 30 respuestas); minima, con una de cada clase (6 y 3).
	comoElJob := ejecucionConUmbrales{queDeciden: 10, informativas: 7, conHaiku: true}
	minima := ejecucionConUmbrales{queDeciden: 1, informativas: 1, conHaiku: true}

	casos := []struct {
		nombre    string
		ejecucion ejecucionConUmbrales
		umbrales  []Umbral
		motivos   []string
		veredicto Veredicto
	}{
		{
			nombre:    "tres-de-51-de-sonnet-5",
			ejecucion: conCambios(comoElJob, func(e *ejecucionConUmbrales) { e.conAlguna = map[string]int{modeloSonnet5: 3} }),
			umbrales: []Umbral{
				umbralDeExpresiones(modeloSonnet5, 3, 51, false, true),
				umbralDeExpresiones(modeloHaiku45, 0, 30, true, false),
			},
			motivos:   []string{"umbral expresiones_prohibidas:claude-sonnet-5: 3 de 51 (5,9 %), y tiene que ser ≤ 5,0 %"},
			veredicto: VeredictoFallo,
		},
		{
			nombre:    "dos-de-51-de-sonnet-5",
			ejecucion: conCambios(comoElJob, func(e *ejecucionConUmbrales) { e.conAlguna = map[string]int{modeloSonnet5: 2} }),
			umbrales: []Umbral{
				umbralDeExpresiones(modeloSonnet5, 2, 51, true, true),
				umbralDeExpresiones(modeloHaiku45, 0, 30, true, false),
			},
			veredicto: VeredictoAprobado,
		},
		{
			nombre:    "dos-de-30-de-haiku-4-5",
			ejecucion: conCambios(comoElJob, func(e *ejecucionConUmbrales) { e.conAlguna = map[string]int{modeloHaiku45: 2} }),
			umbrales: []Umbral{
				umbralDeExpresiones(modeloSonnet5, 0, 51, true, true),
				umbralDeExpresiones(modeloHaiku45, 2, 30, false, false),
			},
			veredicto: VeredictoAprobado,
		},
		{
			// Sin lista, la respuesta con expresiones pasa; y sin objetivo, ninguna
			// duración es un umbral.
			nombre: "sin-lista-ni-objetivo",
			ejecucion: conCambios(minima, func(e *ejecucionConUmbrales) {
				e.sinLista, e.conAlguna, e.duracion = true, map[string]int{modeloSonnet5: 1}, 2000
			}),
			umbrales:  []Umbral{},
			veredicto: VeredictoAprobado,
		},
		{
			nombre:    "901-s-con-objetivo-900",
			ejecucion: conCambios(minima, func(e *ejecucionConUmbrales) { e.duracion, e.objetivo = 901, 900 }),
			umbrales: []Umbral{
				umbralDeExpresiones(modeloSonnet5, 0, 6, true, true),
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
				umbralDeExpresiones(modeloSonnet5, 0, 6, true, true),
				umbralDeExpresiones(modeloHaiku45, 0, 3, true, false),
				umbralDeDuracion(900, 900, true),
			},
			veredicto: VeredictoAprobado,
		},
		{
			// Sin ninguna respuesta medida, la proporción es 0 y el umbral se
			// cumple; el veredicto ya es fallo por las sesiones sin medir, y sus
			// series, sin medir, no dan el motivo de su tasa.
			nombre:    "todas-las-de-sonnet-5-sin-medir",
			ejecucion: conCambios(minima, func(e *ejecucionConUmbrales) { e.sinMedir = modeloSonnet5 }),
			umbrales: []Umbral{
				umbralDeExpresiones(modeloSonnet5, 0, 0, true, true),
				umbralDeExpresiones(modeloHaiku45, 0, 3, true, false),
			},
			motivos:   []string{motivoEsperadoDelLimite(sesionesSinMedirDe(modeloSonnet5, 2)...)},
			veredicto: VeredictoFallo,
		},
		{
			// Los tres motivos nuevos a la vez, en su orden: el del umbral que
			// decide, el de las sesiones sin medir —las de Haiku 4.5, cuyo umbral
			// queda en 0 de 0— y el de la duración.
			nombre: "los-tres-motivos-en-su-orden",
			ejecucion: conCambios(minima, func(e *ejecucionConUmbrales) {
				e.conAlguna, e.sinMedir, e.duracion, e.objetivo = map[string]int{modeloSonnet5: 1}, modeloHaiku45, 901, 900
			}),
			umbrales: []Umbral{
				umbralDeExpresiones(modeloSonnet5, 1, 6, false, true),
				umbralDeExpresiones(modeloHaiku45, 0, 0, true, false),
				umbralDeDuracion(901, 900, false),
			},
			motivos: []string{
				"umbral expresiones_prohibidas:claude-sonnet-5: 1 de 6 (16,7 %), y tiene que ser ≤ 5,0 %",
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

	modelos := []string{modeloSonnet5}
	if ejecucion.conHaiku {
		modelos = append(modelos, modeloHaiku45)
	}

	eval := contenidoDeLaSesion(t, filepath.Join(casosDeInforme, casoAprobado, "evals"), ficheroDeLaEval01)

	// series son, por modelo, los números de las evals de sus series, en su orden.
	series := map[string][]int{}

	for numero := 1; numero <= ejecucion.queDeciden+ejecucion.informativas; numero++ {
		informativa := numero > ejecucion.queDeciden

		contenido := eval
		if informativa {
			contenido += "informativa: true\n"
		}

		escribirEnLaCopia(t, evals, ficheroSintetico(numero), contenido)

		for _, modelo := range modelos {
			if informativa && modelo != modeloSonnet5 {
				continue
			}

			series[modelo] = append(series[modelo], numero)

			for vez := 1; vez <= repeticionesConUmbrales; vez++ {
				escribirSesionSintetica(t, copia, numero, modelo, vez, ejecucion.sinMedir == modelo)
			}
		}
	}

	for modelo, cuantas := range ejecucion.conAlguna {
		delModelo := series[modelo]
		require.LessOrEqual(t, cuantas, len(delModelo), "hay una serie de %s por respuesta con expresiones", modelo)

		for k := range cuantas {
			numero := delModelo[k*len(delModelo)/cuantas]
			anteponerALaRespuesta(t, filepath.Join(copia, "sesiones", sesionSintetica(numero, modelo, 1)),
				transicionDeLaMemoria+"\n\n")
		}
	}

	return informeDeLaCopia(t, copia, func(entradas *InformeAEscribir) {
		entradas.ModeloQueDecide = modeloSonnet5
		entradas.ModelosInformativos = modelos[1:]
		entradas.Repeticiones, entradas.Umbral = repeticionesConUmbrales, umbralConUmbrales
		entradas.DuracionDeLasSesiones, entradas.ObjetivoDeDuracion = ejecucion.duracion, ejecucion.objetivo
	})
}

// escribirSesionSintetica escribe la sesión de esa repetición de la eval de ese
// número con el modelo dado: una copia de la del art. 21 del caso aprobado que lo
// pide y lo declara, con eval.txt nombrando su eval y, si sinMedir, terminada
// con el mensaje del límite de uso.
func escribirSesionSintetica(t *testing.T, copia string, numero int, modelo string, vez int, sinMedir bool) {
	t.Helper()

	dir := filepath.Join(copia, "sesiones", sesionSintetica(numero, modelo, vez))
	copiarSesionConOtroModelo(t, filepath.Join(casosDeInforme, casoAprobado, "sesiones", sesionDelArticulo21), dir, modelo)
	escribirEnLaCopia(t, dir, "eval.txt", ficheroSintetico(numero)+"\n")

	if sinMedir {
		cambiarElFinalDeLaSesion(t, dir, 1, mensajeResultConError(t, textoDelLimiteDeSesion))
	}
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
