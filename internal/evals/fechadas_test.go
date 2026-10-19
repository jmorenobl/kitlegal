package evals

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

// Las líneas de las tres formas fijas de una redacción fechada con las que
// TestExtraerRedaccionesFechadas lee una respuesta, las de research D13 de H20
// con los valores de las evals (a) y (b) de contracts/evals-redaccion-fechada.md
// §1.3.
const (
	// lineaDeLaRedaccionAFecha es la forma A del art. 118 de la LCSP el
	// 2019-06-01, con la fecha de vigencia de la redacción que se cita.
	lineaDeLaRedaccionAFecha = "⚠ REDACCIÓN A FECHA: " + citaDelArticulo118 + ": se cita la redacción vigente el " +
		"2019-06-01, la de fecha de vigencia 20180309."

	// lineaDeLaRedaccionDeHoy es la forma B del mismo bloque, con la norma y la
	// fecha de vigencia de la redacción de hoy.
	lineaDeLaRedaccionDeHoy = "⚠ REDACCIÓN DE HOY: " + citaDelArticulo118 + ": hoy la redacción es otra, la de " +
		"BOE-A-2020-1651, con fecha de vigencia 20200206."

	// lineaSinRedaccionVigente es la forma C del art. 21 de la Ley 39/2015 el
	// 2016-06-01, con la fecha de vigencia de su primera redacción.
	lineaSinRedaccionVigente = "⚠ SIN REDACCIÓN VIGENTE: art. 21 de la Ley 39/2015 [BOE-A-2015-10565, bloque a21]: " +
		"el 2016-06-01 no tenía redacción vigente; la tiene desde el 20161002, la fecha de vigencia de su primera " +
		"redacción."
)

// TestExtraerRedaccionesFechadas fija lo que ExtraerRedaccionesFechadas lee de
// una respuesta (contracts/evals-redaccion-fechada.md §2 de H20; data-model §7.3;
// research D13; FR-052): de cada línea con una de las tres formas fijas —la de
// los avisos, también con las tolerancias de su gramática— y al menos una cita,
// su forma, la norma y el bloque de su primera cita y lo que lleva detrás de
// ella: en la redacción a fecha y en la línea sin redacción vigente, la primera
// fecha AAAA-MM-DD y la primera AAAAMMDD; en la redacción de hoy, el primer
// identificador BOE-A-… y la primera AAAAMMDD; lo que la línea no lleva queda
// vacío, y las de todas las líneas van en el orden de la respuesta; una línea con
// dos formas tiene la que empieza antes. Nada de una línea sin cita, de lo que va
// delante de la cita, de una forma sin la marca ni de la forma de un aviso o de
// un hallazgo de hoy.
func TestExtraerRedaccionesFechadas(t *testing.T) {
	t.Parallel()

	aFecha := LineaFechada{
		Forma: RedaccionAFecha, Norma: normaDeLaLCSP, Bloque: bloqueDelArticulo118,
		Fecha: "2019-06-01", FechaVigencia: fechaDeLaRedaccionOriginal,
	}
	deHoy := LineaFechada{
		Forma: RedaccionDeHoy, Norma: normaDeLaLCSP, Bloque: bloqueDelArticulo118,
		FechaVigencia: fechaDelArticulo118Vigente, NormaModificadora: "BOE-A-2020-1651",
	}
	sinVigente := LineaFechada{
		Forma: SinRedaccionVigente, Norma: normaDeLasTrazas, Bloque: "a21",
		Fecha: "2016-06-01", FechaVigencia: "20161002",
	}

	casos := []struct {
		nombre string
		texto  string
		lineas []LineaFechada
	}{
		{nombre: "sin-ninguna-forma", texto: cuerpoConLasDosCitas},
		{nombre: "la-redaccion-a-fecha", texto: lineaDeLaRedaccionAFecha, lineas: []LineaFechada{aFecha}},
		{nombre: "la-redaccion-de-hoy", texto: lineaDeLaRedaccionDeHoy, lineas: []LineaFechada{deHoy}},
		{nombre: "sin-redaccion-vigente", texto: lineaSinRedaccionVigente, lineas: []LineaFechada{sinVigente}},
		{
			// Las dos líneas de la eval (a), en el orden de la respuesta, entre el
			// cuerpo, que cita el bloque fuera de ellas.
			nombre: "dos-lineas",
			texto: cuerpoConLasDosCitas + "\n\n" + lineaDeLaRedaccionDeHoy + "\r\n" + lineaDeLaRedaccionAFecha +
				"\n",
			lineas: []LineaFechada{deHoy, aFecha},
		},
		{
			// Con el selector de presentación U+FE0F, el énfasis de Markdown y las
			// etiquetas en minúsculas o con otras mayúsculas.
			nombre: "con-mayusculas-y-enfasis",
			texto: "**⚠\xef\xb8\x8f redacción a fecha:** " + citaDelArticulo118 + ": el 2019-06-01, 20180309.\n" +
				"_⚠ Redacción De Hoy_: " + citaDelArticulo118 + ": BOE-A-2020-1651, 20200206.\n" +
				"⚠ **SIN**  _redacción_ VIGENTE : [BOE-A-2015-10565, bloque a21] el 2016-06-01, 20161002.",
			lineas: []LineaFechada{aFecha, deHoy, sinVigente},
		},
		{
			nombre: "una-linea-sin-cita",
			texto: "⚠ REDACCIÓN A FECHA: el art. 118 de la LCSP, en la redacción vigente el 2019-06-01, la de " +
				"fecha de vigencia 20180309.",
		},
		{
			// La cita de la línea siguiente no es la de esta.
			nombre: "con-la-cita-en-otra-linea",
			texto:  "⚠ SIN REDACCIÓN VIGENTE: el 2016-06-01 no la tenía; la tiene desde el 20161002.\n" + citaDelArticulo118,
		},
		{
			// Las fechas y el identificador de delante de la cita no se leen: son los
			// de detrás, y la norma de la propia cita no es la de la redacción de hoy.
			nombre: "lo-leido-detras-de-la-cita",
			texto: "⚠ REDACCIÓN A FECHA: el 2019-06-02, con 20200206, " + citaDelArticulo118 +
				": se cita la redacción vigente el 2019-06-01, la de fecha de vigencia 20180309.\n" +
				"⚠ REDACCIÓN DE HOY: BOE-A-2020-9999, 20200205, " + citaDelArticulo118 +
				": hoy la redacción es la de BOE-A-2020-1651, con fecha de vigencia 20200206.",
			lineas: []LineaFechada{aFecha, deHoy},
		},
		{
			// Con más de una cita, la norma y el bloque son los de la primera, y lo
			// leído va detrás de ella.
			nombre: "la-primera-cita-de-la-linea",
			texto: "⚠ REDACCIÓN A FECHA: " + citaDelArticulo118 + " y " + citaDeLaDA3 +
				": la vigente el 2019-06-01, la de 20180309.",
			lineas: []LineaFechada{aFecha},
		},
		{
			// Lo que la línea no lleva detrás de la cita queda vacío, y una fecha de
			// la redacción a fecha no es lo que lee la de hoy.
			nombre: "sin-lo-que-se-lee",
			texto: "⚠ REDACCIÓN A FECHA: " + citaDelArticulo118 + ": la vigente entonces.\n" +
				"⚠ REDACCIÓN DE HOY: " + citaDelArticulo118 + ": el 2019-06-01 era otra.",
			lineas: []LineaFechada{
				{Forma: RedaccionAFecha, Norma: normaDeLaLCSP, Bloque: bloqueDelArticulo118},
				{Forma: RedaccionDeHoy, Norma: normaDeLaLCSP, Bloque: bloqueDelArticulo118},
			},
		},
		{
			// Una fecha con otra cifra a un lado no es la pedida, ni ocho cifras
			// con un mes imposible son una fecha de vigencia: son las que siguen.
			nombre: "sin-otra-cifra-a-los-lados",
			texto: "⚠ SIN REDACCIÓN VIGENTE: [BOE-A-2015-10565, bloque a21]: 12016-06-01, 2016-06-012, " +
				"20161302, el 2016-06-01, desde el 20161002.",
			lineas: []LineaFechada{sinVigente},
		},
		{
			// Una línea con dos formas tiene la que empieza antes, aunque sea la de
			// hoy.
			nombre: "dos-formas-en-una-linea",
			texto: "⚠ REDACCIÓN DE HOY: " + citaDelArticulo118 + ": la de BOE-A-2020-1651, de 20200206; " +
				"⚠ REDACCIÓN A FECHA: el 2019-06-01, la de 20180309.",
			lineas: []LineaFechada{deHoy},
		},
		{
			nombre: "la-forma-sin-la-marca",
			texto:  "REDACCIÓN A FECHA: " + citaDelArticulo118 + ": la vigente el 2019-06-01, la de 20180309.",
		},
		{
			// Las formas de hoy, la de un aviso y la de un hallazgo, con la cita y
			// las fechas, no son ninguna de las tres.
			nombre: "las-formas-de-hoy",
			texto: "⚠ NORMA DEROGADA: " + citaDelArticulo118 + ": derogada el 2019-06-01.\n" +
				"⚠ REDACCIÓN MODIFICADA: " + citaDelArticulo118 + ": la de 20180309 ha sido sustituida por la de " +
				"20200206, BOE-A-2020-1651.",
		},
	}

	for _, caso := range casos {
		t.Run(caso.nombre, func(t *testing.T) {
			t.Parallel()

			assert.Equal(t, caso.lineas, ExtraerRedaccionesFechadas(caso.texto))
		})
	}
}
