package evals

import (
	"fmt"
	"strconv"
	"strings"
)

// umbralDeExpresionesProhibidas es la proporción de las respuestas de un modelo
// con alguna expresión prohibida que su umbral admite: el 5 % de H7.2 SC 001. Va
// con la lista, en el paquete, y no en la definición del job, que solo da el
// objetivo de duración (research D5 de H7.3; FR-002, FR-004).
const umbralDeExpresionesProhibidas = 0.05

// Lo que lleva cada umbral del informe (contrato informe-del-job §1 de H7.3;
// data-model §1). El de las expresiones de un modelo se nombra con el recuento
// que ya publica el informe seguido del modelo, y su descripción lo nombra; y la
// única comparación de este hito es «<=», que informe.md escribe «≤».
const (
	prefijoDelUmbralDeExpresiones  = "expresiones_prohibidas:"
	nombreDelUmbralDeLaDuracion    = "duracion_de_las_sesiones"
	descripcionDelUmbralDeUnModelo = "Respuestas de %s con alguna expresión prohibida, sobre sus respuestas medidas en " +
		"las evals que activan la skill"
	descripcionDelUmbralDeLaDuracion = "Segundos desde que se prepara la primera sesión hasta que termina la última"

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
	// informe: expresiones_prohibidas:<modelo> o duracion_de_las_sesiones.
	Nombre string `json:"nombre"`

	// Descripcion dice en una línea qué se mide y sobre qué respuestas.
	Descripcion string `json:"descripcion"`

	// Medida es lo medido: las respuestas con alguna expresión prohibida, o los
	// segundos de las sesiones.
	Medida float64 `json:"medida"`

	// Total, solo en los de las expresiones, son las respuestas medidas del
	// modelo en las evals que activan la skill: con él, lo que se compara es la
	// proporción Medida/Total, 0 si Total es 0. Sin él, la propia Medida.
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
// (research D5 de H7.3; FR-002, FR-004 a FR-006, FR-051): uno de las
// expresiones prohibidas por cada modelo del recuento —que solo existe si la
// skill tiene lista—, en su orden, el que decide delante de los informativos,
// con las respuestas con alguna como medida y las respuestas medidas como total,
// y que decide solo en el modelo que decide; y, detrás, el de la duración de las
// sesiones si el job da un objetivo mayor que 0, sin total y decidiendo. Nil,
// [] en informe.json, si no hay ninguno.
func umbralesDelInforme(e InformeAEscribir, recuento []RecuentoDeExpresiones) []Umbral {
	var umbrales []Umbral

	for _, delModelo := range recuento {
		total := delModelo.Respuestas

		umbrales = append(umbrales, compararUmbral(Umbral{
			Nombre:      prefijoDelUmbralDeExpresiones + delModelo.Modelo,
			Descripcion: fmt.Sprintf(descripcionDelUmbralDeUnModelo, delModelo.Modelo),
			Medida:      float64(delModelo.ConAlguna),
			Total:       &total,
			Umbral:      umbralDeExpresionesProhibidas,
			Decide:      delModelo.Modelo == e.ModeloQueDecide,
		}))
	}

	if e.ObjetivoDeDuracion > 0 {
		umbrales = append(umbrales, compararUmbral(Umbral{
			Nombre:      nombreDelUmbralDeLaDuracion,
			Descripcion: descripcionDelUmbralDeLaDuracion,
			Medida:      float64(e.DuracionDeLasSesiones),
			Umbral:      float64(e.ObjetivoDeDuracion),
			Decide:      true,
		}))
	}

	return umbrales
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
// se cumplen, en su orden, sin el de la duración, que es de la ejecución
// (motivoDeLaDuracionDeLasSesiones): «umbral <nombre>: <medida>, y tiene que ser
// <condición>», con la medida y la condición como en informe.md (contrato
// informe-del-job §2.1 de H7.3; FR-003).
func motivosDeLosUmbrales(umbrales []Umbral) []string {
	var motivos []string

	for _, umbral := range umbrales {
		if umbral.Nombre == nombreDelUmbralDeLaDuracion || !umbral.Decide || umbral.Cumple {
			continue
		}

		motivos = append(motivos, fmt.Sprintf(motivoDeUnUmbral, umbral.Nombre, umbral.medidaEscrita(),
			umbral.condicionEscrita()))
	}

	return motivos
}

// motivoDeLaDuracionDeLasSesiones es el motivo de la raíz del umbral de la
// duración si decide y no se cumple, con el prefijo de la ejecución: no es de la
// skill (contrato informe-del-job §2.3 de H7.3; FR-051). Ninguno si no lo hay.
func motivoDeLaDuracionDeLasSesiones(umbrales []Umbral) []string {
	for _, umbral := range umbrales {
		if umbral.Nombre == nombreDelUmbralDeLaDuracion && umbral.Decide && !umbral.Cumple {
			return []string{motivoDeLaEjecucion + fmt.Sprintf(motivoDeLaDuracion, umbral.Nombre,
				umbral.medidaEscrita(), umbral.condicionEscrita())}
		}
	}

	return nil
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
