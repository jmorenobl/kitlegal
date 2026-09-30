package evals

import (
	"regexp"
	"strings"

	"github.com/jmorenobl/kitlegal/internal/core/grafo"
)

// fechasDeUnaRedaccion son las fechas de una redacción modificada que se leen de
// su línea: la de la superada y la de la leída.
const fechasDeUnaRedaccion = 2

// cifrasSeguidas es cada tramo de cifras de un texto, entero: una fecha de
// vigencia es uno de ellos, así que no tiene otra cifra a ninguno de sus lados.
var cifrasSeguidas = regexp.MustCompile(`\p{Nd}+`)

// formaDeFechaDeVigencia es la de una fecha de vigencia, AAAAMMDD con un mes y un
// día posibles: la del esquema de eval (contracts/evals-y-juicio.md §1 de H7.4).
var formaDeFechaDeVigencia = regexp.MustCompile(`^[0-9]{4}(?:0[1-9]|1[0-2])(?:0[1-9]|[12][0-9]|3[01])$`)

// ExtraerRedaccionesModificadas devuelve las redacciones modificadas que traslada
// la respuesta, en el orden de sus líneas, o nil si no traslada ninguna
// (contracts/evals-y-juicio.md §2 de H7.4; data-model §2; research D10): de cada
// línea con la forma fija de version-obsoleta —la de formasDeHallazgo, con la que
// ExtraerHallazgos la encuentra—, una con la norma y el bloque de su primera cita
// (ExtraerCitas) y, como fecha de vigencia y fecha de vigencia reciente, sus dos
// primeras fechas de ocho cifras con un mes y un día posibles y sin otra cifra a
// los lados, en el orden en que las lleva. Una línea sin cita o con menos de dos
// fechas no da ninguna: la cita y las fechas de otra línea no son las suyas.
func ExtraerRedaccionesModificadas(texto string) []RedaccionEsperada {
	forma := formasDeHallazgo()[string(grafo.ClaseVersionObsoleta)]

	var redacciones []RedaccionEsperada

	for linea := range strings.Lines(texto) {
		if !forma.MatchString(linea) {
			continue
		}

		citas := ExtraerCitas(linea)
		fechas := fechasDeVigencia(linea)

		if len(citas) == 0 || len(fechas) < fechasDeUnaRedaccion {
			continue
		}

		redacciones = append(redacciones, RedaccionEsperada{
			Norma:                 citas[0].Norma,
			Bloque:                citas[0].Bloque,
			FechaVigencia:         fechas[0],
			FechaVigenciaReciente: fechas[1],
		})
	}

	return redacciones
}

// fechasDeVigencia son las primeras fechas de vigencia de la línea, como mucho
// las de una redacción, en su orden: los tramos de cifras que tienen la forma de
// una fecha.
func fechasDeVigencia(linea string) []string {
	var fechas []string

	for _, cifras := range cifrasSeguidas.FindAllString(linea, -1) {
		if !formaDeFechaDeVigencia.MatchString(cifras) {
			continue
		}

		fechas = append(fechas, cifras)
		if len(fechas) == fechasDeUnaRedaccion {
			break
		}
	}

	return fechas
}

// texto es el de la redacción esperada en los motivos, en el informe y en las
// formas exigidas de su serie: <norma> <bloque> <fecha_vigencia>
// <fecha_vigencia_reciente> (data-model §2 de H7.4).
func (r RedaccionEsperada) texto() string {
	return strings.Join([]string{r.Norma, r.Bloque, r.FechaVigencia, r.FechaVigenciaReciente}, " ")
}
