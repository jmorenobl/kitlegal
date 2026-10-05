package evals

import (
	"fmt"
	"strconv"
	"strings"
)

// La proporción de las respuestas medidas del modelo que decide que admite el
// umbral de las que no activaron la skill: ninguna (FR-041 de H7.4). Va en el
// paquete, y no en la definición del job, que solo da el objetivo de duración
// (research D5 de H7.3; FR-002, FR-004 de H7.3).
const umbralDeRespuestasSinActivar = 0

// Lo que lleva cada umbral del informe (contrato informe-del-job §1 de H7.3;
// contracts/informe-del-job.md §2 de H7.4 y de H24; data-model §1;
// contracts/evals-en-dos-modos.md §5.1 de H21; data-model §10 de H21). Cada uno
// se mide sobre las sesiones de un solo modo: el de las respuestas de un modelo
// se nombra con lo que mide seguido del modelo y del modo, y el de la duración,
// con el modo; su descripción los nombra, primero el modelo y después el modo; y
// la única comparación es «<=», que informe.md escribe «≤».
const (
	prefijoDelUmbralSinActivar     = "sin_activar:"
	prefijoDelUmbralDeLaDuracion   = "duracion_de_las_sesiones:"
	separadorDelModoDelUmbral      = ":"
	descripcionDelUmbralSinActivar = "Respuestas de %s en el modo %s sin la skill activada, sobre sus respuestas " +
		"medidas en las evals que la activan"
	descripcionDelUmbralDeLaDuracion = "Segundos de la tanda del modo %s, desde que se prepara su primera sesión " +
		"hasta que termina la última"

	comparacionMenorOIgual = "<="
	signoMenorOIgual       = "≤ "
)

// Textos fijos de los umbrales en informe.md y en los motivos (contrato
// informe-del-job §2 y §4 de H7.3): la celda de uno que no decide, y lo que
// sigue a la medida y a la condición de uno que no se cumple.
const (
	soloSePublica      = "no: solo se publica"
	motivoDeUnUmbral   = "umbral %s: %s, y tiene que ser %s"
	motivoDeLaDuracion = "%s: %s s, y tiene que ser %s s"
)

// Umbral es un umbral del informe del job de evals, con los campos, las claves y
// los invariantes del contrato de umbrales del ADR 0029 (data-model §1 de H7.3).
// Sus claves JSON, en el orden de sus campos, son las de cada elemento de
// umbrales en informe.json.
type Umbral struct {
	// Nombre es lo que cita el plan en evals:<skill>:<nombre>, único en el
	// informe: sin_activar:<modelo>:<modo> o duracion_de_las_sesiones:<modo>, con
	// <modo> orden u herramienta (data-model §10 de H21).
	Nombre string `json:"nombre"`

	// Descripcion dice en una línea qué se mide, en qué modo y sobre qué
	// respuestas.
	Descripcion string `json:"descripcion"`

	// Medida es lo medido en las sesiones de su modo: las respuestas sin la
	// skill activada, o los segundos de su tanda.
	Medida float64 `json:"medida"`

	// Total, solo en los de las respuestas, son las respuestas medidas del
	// modelo en ese modo en las evals que activan la skill: con él, lo que se
	// compara es la proporción Medida/Total, 0 si Total es 0. Sin él, la propia
	// Medida.
	Total *int `json:"total,omitzero"`

	// Comparacion es siempre «<=» en este hito, y Umbral, el valor con el que se
	// compara, en la unidad de lo que se compara: una proporción, o segundos.
	Comparacion string  `json:"comparacion"`
	Umbral      float64 `json:"umbral"`

	// Cumple es el resultado de la comparación, en coma flotante de doble
	// precisión y sin redondeos.
	Cumple bool `json:"cumple"`

	// Decide dice si incumplirlo pone el veredicto en fallo; si no, solo se
	// publica.
	Decide bool `json:"decide"`
}

// umbralesDelInforme son los umbrales del informe, sin ningún caso por skill
// (research D5 de H7.3 y D14 de H7.4; FR-002, FR-004 a FR-006, FR-051 de H7.3;
// FR-040 a FR-048 de H7.4): por cada elemento del recuento —que solo existe si
// la skill tiene juez (research D13 de H24)—, el de las respuestas del modelo
// que decide sin la skill activada, con sus respuestas medidas como total, que
// decide con 0. Detrás de todos, los de la duración de las sesiones si el job
// da un objetivo mayor que 0, sin total y decidiendo. Nil, [] en informe.json,
// si no hay ninguno.
//
// Desde H21 (contracts/evals-en-dos-modos.md §5.1 de H21; data-model §10;
// research.md D19; FR-043 a FR-045), cada umbral es de un modo del plan y lo
// nombra: el recuento lleva un elemento por modo, el del modo orden delante, y
// detrás de los de las respuestas, el de la duración de cada modo, en su orden,
// que mide los segundos de la tanda de ese modo frente al mismo objetivo.
// Ninguno suma las medidas de dos modos ni cuenta las sesiones o la tanda de las
// evals sin binario ni servidor.
//
// Desde H24 (contracts/informe-del-job.md §2 y §6 de H24; FR-034), ninguno mide
// las expresiones de la lista: salen expresiones_prohibidas:<modelo>:<modo>, de
// todos los modelos, y redaccion_no_leida:<modelo>:<modo>. Con el plan del job,
// los dos modos, y el juez y el objetivo de boe-legislacion son cuatro, y todos
// deciden.
func umbralesDelInforme(e InformeAEscribir, recuento []recuentoDeRespuestas) []Umbral {
	var umbrales []Umbral

	for _, delModo := range recuento {
		umbrales = append(umbrales, umbralDeSinActivar(delModo))
	}

	if e.ObjetivoDeDuracion > 0 {
		for _, modo := range e.plan(nil).modos() {
			umbrales = append(umbrales, compararUmbral(Umbral{
				Nombre:      prefijoDelUmbralDeLaDuracion + string(modo),
				Descripcion: fmt.Sprintf(descripcionDelUmbralDeLaDuracion, modo),
				Medida:      float64(e.DuracionDeLosModos[modo]),
				Umbral:      float64(e.ObjetivoDeDuracion),
				Decide:      true,
			}))
		}
	}

	return umbrales
}

// umbralDeSinActivar es el umbral de las respuestas sin la skill activada del
// recuento de un modelo en un modo, sobre sus respuestas medidas, ya comparado:
// se nombra con su prefijo seguido del modelo y del modo, su descripción los
// nombra, y decide.
func umbralDeSinActivar(delModo recuentoDeRespuestas) Umbral {
	total := delModo.respuestas

	return compararUmbral(Umbral{
		Nombre:      prefijoDelUmbralSinActivar + delModo.modelo + separadorDelModoDelUmbral + string(delModo.modo),
		Descripcion: fmt.Sprintf(descripcionDelUmbralSinActivar, delModo.modelo, delModo.modo),
		Medida:      float64(delModo.sinActivar),
		Total:       &total,
		Umbral:      umbralDeRespuestasSinActivar,
		Decide:      true,
	})
}

// esDeLaDuracion dice si el umbral es el de la duración de las sesiones de un
// modo: su motivo no es de la skill sino de la ejecución.
func (u Umbral) esDeLaDuracion() bool {
	return strings.HasPrefix(u.Nombre, prefijoDelUmbralDeLaDuracion)
}

// compararUmbral pone al umbral su comparación, «<=», y su resultado: su valor
// comparado con Umbral en float64, sin redondear nada (contrato de umbrales del
// ADR 0029).
func compararUmbral(umbral Umbral) Umbral {
	umbral.Comparacion = comparacionMenorOIgual
	umbral.Cumple = umbral.valor() <= umbral.Umbral

	return umbral
}

// valor es lo que el umbral compara: la proporción Medida/Total —0 si Total es
// 0— si lleva total y, si no, la propia Medida.
func (u Umbral) valor() float64 {
	switch {
	case u.Total == nil:
		return u.Medida
	case *u.Total == 0:
		return 0
	default:
		return u.Medida / float64(*u.Total)
	}
}

// motivosDeLosUmbrales son los motivos de la raíz de los umbrales que deciden y no
// se cumplen, en su orden, sin los de la duración, que son de la ejecución
// (motivosDeLaDuracionDeLasSesiones): «umbral <nombre>: <medida>, y tiene que ser
// <condición>», con la medida y la condición como en informe.md (contrato
// informe-del-job §2.1 de H7.3; FR-003). El nombre lleva el modo del umbral, y
// con él lo nombra el motivo (contracts/evals-en-dos-modos.md §5.2 de H21;
// FR-044).
func motivosDeLosUmbrales(umbrales []Umbral) []string {
	var motivos []string

	for _, umbral := range umbrales {
		if umbral.esDeLaDuracion() || !umbral.Decide || umbral.Cumple {
			continue
		}

		motivos = append(motivos, fmt.Sprintf(motivoDeUnUmbral, umbral.Nombre, umbral.medidaEscrita(),
			umbral.condicionEscrita()))
	}

	return motivos
}

// motivosDeLaDuracionDeLasSesiones son los motivos de la raíz de los umbrales de
// la duración que deciden y no se cumplen, en su orden, uno por modo, con el
// prefijo de la ejecución: no son de la skill (contrato informe-del-job §2.3 de
// H7.3; contracts/evals-en-dos-modos.md §5.2 de H21; FR-051). Ninguno si no los
// hay.
func motivosDeLaDuracionDeLasSesiones(umbrales []Umbral) []string {
	var motivos []string

	for _, umbral := range umbrales {
		if umbral.esDeLaDuracion() && umbral.Decide && !umbral.Cumple {
			motivos = append(motivos, motivoDeLaEjecucion+fmt.Sprintf(motivoDeLaDuracion, umbral.Nombre,
				umbral.medidaEscrita(), umbral.condicionEscrita()))
		}
	}

	return motivos
}

// medidaEscrita es la medida del umbral como la escriben informe.md y los
// motivos: con total, «<medida> de <total> (<p> %)», con el porcentaje con un
// decimal y coma, 0 si total es 0; sin él, la medida sola.
func (u Umbral) medidaEscrita() string {
	medida := strconv.FormatFloat(u.Medida, 'f', -1, 64)
	if u.Total == nil {
		return medida
	}

	return medida + " de " + strconv.Itoa(*u.Total) + " (" + porcentajeConComa(u.valor()) + " %)"
}

// condicionEscrita es la condición del umbral como la escriben informe.md y los
// motivos: «≤ <u> %», con un decimal y coma, si compara una proporción, y
// «≤ <u>» si no.
func (u Umbral) condicionEscrita() string {
	if u.Total == nil {
		return signoMenorOIgual + strconv.FormatFloat(u.Umbral, 'f', -1, 64)
	}

	return signoMenorOIgual + porcentajeConComa(u.Umbral) + " %"
}

// porcentajeConComa es la proporción como porcentaje con un decimal y coma
// decimal, como el informe final: 3/51 es «5,9».
func porcentajeConComa(cociente float64) string {
	return strings.Replace(strconv.FormatFloat(cociente*100, 'f', 1, 64), ".", ",", 1)
}

// filasDeUmbrales son las filas de la tabla de los umbrales de informe.md
// (contrato informe-del-job §4 de H7.3): el nombre, como código, como en las
// filas del contrato; la medida, la condición, si se cumple y si hace fallar el
// veredicto, «no: solo se publica» en los que no deciden.
func filasDeUmbrales(umbrales []Umbral) [][]string {
	filas := make([][]string, 0, len(umbrales))

	for _, umbral := range umbrales {
		decide := siONo(umbral.Decide)
		if !umbral.Decide {
			decide = soloSePublica
		}

		filas = append(filas, []string{
			"`" + umbral.Nombre + "`", umbral.medidaEscrita(), umbral.condicionEscrita(), siONo(umbral.Cumple), decide,
		})
	}

	return filas
}
