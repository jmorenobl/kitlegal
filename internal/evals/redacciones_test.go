package evals

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

// TestExtraerRedaccionesModificadas fija lo que ExtraerRedaccionesModificadas lee
// de una respuesta (contracts/evals-y-juicio.md §2 de H7.4; data-model §2; research
// D10): de cada línea con la forma fija de version-obsoleta —con las tolerancias
// de la de los avisos—, una redacción con la primera cita de la línea y sus dos
// primeras fechas de ocho cifras con un mes y un día posibles y sin otra cifra a
// los lados, en el orden en que las lleva, y las de todas las líneas en el orden
// de la respuesta; nada de una línea sin la forma, sin cita o con menos de dos
// fechas, ni de una cita o unas fechas de otra línea.
func TestExtraerRedaccionesModificadas(t *testing.T) {
	t.Parallel()

	delArticulo118 := RedaccionEsperada{
		Norma: normaDeLaLCSP, Bloque: bloqueDelArticulo118,
		FechaVigencia: fechaDeLaRedaccionOriginal, FechaVigenciaReciente: fechaDelArticulo118Vigente,
	}
	deLaDA3 := RedaccionEsperada{
		Norma: normaDeLaLCSP, Bloque: bloqueDeLaDA3,
		FechaVigencia: fechaDeLaRedaccionOriginal, FechaVigenciaReciente: fechaDeLaDA3Vigente,
	}

	casos := []struct {
		nombre    string
		texto     string
		esperadas []RedaccionEsperada
	}{
		{nombre: "sin-la-forma", texto: cuerpoConLasDosCitas},
		{
			nombre:    "una-por-linea",
			texto:     lineaConLaCitaDel118() + "\n\n" + lineaConLaCitaDeLaDA3() + "\n\n" + cuerpoConLasDosCitas,
			esperadas: []RedaccionEsperada{delArticulo118, deLaDA3},
		},
		{
			nombre:    "en-el-orden-de-la-respuesta",
			texto:     lineaConLaCitaDeLaDA3() + "\n" + lineaConLaCitaDel118(),
			esperadas: []RedaccionEsperada{deLaDA3, delArticulo118},
		},
		{
			nombre:    "con-saltos-de-linea-de-windows",
			texto:     lineaConLaCitaDel118() + "\r\n" + lineaConLaCitaDeLaDA3() + "\r\n",
			esperadas: []RedaccionEsperada{delArticulo118, deLaDA3},
		},
		{
			// La forma con el selector de presentación U+FE0F y el énfasis
			// envolviendo la etiqueta en minúsculas.
			nombre: "con-la-forma-tolerada",
			texto: "**\xe2\x9a\xa0\xef\xb8\x8f Redacci\xc3\xb3n modificada**: " + citaDelArticulo118 +
				": de 20180309 a 20200206.",
			esperadas: []RedaccionEsperada{delArticulo118},
		},
		{
			nombre: "la-redaccion-dicha-sin-la-forma",
			texto: "La redacci\xc3\xb3n del " + citaDelArticulo118 + " con fecha de vigencia 20180309 ha sido " +
				"sustituida por la de 20200206.",
		},
		{
			nombre: "sin-cita",
			texto:  lineaDeRedaccion("", fechaDeLaRedaccionOriginal, fechaDelArticulo118Vigente),
		},
		{
			// La cita de la línea siguiente no es la de esta.
			nombre: "con-la-cita-en-otra-linea",
			texto: lineaDeRedaccion("", fechaDeLaRedaccionOriginal, fechaDelArticulo118Vigente) + "\n" +
				citaDelArticulo118,
		},
		{
			nombre: "con-una-sola-fecha",
			texto:  "\xe2\x9a\xa0 REDACCI\xc3\x93N MODIFICADA: " + citaDelArticulo118 + ": la vigente es la de 20200206.",
		},
		{
			// Las fechas de la línea siguiente no son las de esta.
			nombre: "con-las-fechas-en-otra-linea",
			texto: "\xe2\x9a\xa0 REDACCI\xc3\x93N MODIFICADA: " + citaDelArticulo118 + ": la redacci\xc3\xb3n ha " +
				"sido sustituida.\n20180309 y 20200206",
		},
		{
			// La primera cita de la línea y sus dos primeras fechas, aunque lleve
			// más: la línea de los dos bloques da solo la del primero.
			nombre:    "las-dos-en-una-sola-linea",
			texto:     lineaConLaCitaDel118() + " " + lineaConLaCitaDeLaDA3(),
			esperadas: []RedaccionEsperada{delArticulo118},
		},
		{
			// Un mes 13, un día 32 y un día 00 no son fechas: las dos primeras son
			// las que siguen.
			nombre: "sin-las-fechas-imposibles",
			texto: "\xe2\x9a\xa0 REDACCI\xc3\x93N MODIFICADA: " + citaDelArticulo118 +
				": 20181309, 20180332, 20180300, 20180309 y 20200206.",
			esperadas: []RedaccionEsperada{delArticulo118},
		},
		{
			// Ocho cifras con otra cifra delante o detrás —también de otra
			// escritura, como el uno de ancho completo U+FF11— no son una fecha, ni
			// lo son siete, ni las de una fecha escrita con guiones.
			nombre: "sin-otra-cifra-a-los-lados",
			texto: "\xe2\x9a\xa0 REDACCI\xc3\x93N MODIFICADA: " + citaDelArticulo118 +
				": 120180309, 201803091, \xef\xbc\x9120180309, 2018030, 2018-03-09, 20180309 y 20200206.",
			esperadas: []RedaccionEsperada{delArticulo118},
		},
	}

	for _, caso := range casos {
		t.Run(caso.nombre, func(t *testing.T) {
			t.Parallel()

			assert.Equal(t, caso.esperadas, ExtraerRedaccionesModificadas(caso.texto))
		})
	}
}

// TestTextoDeLaRedaccion fija el texto de una redacción esperada, el de los
// motivos, del informe y de sus formas exigidas: la norma, el bloque y las dos
// fechas, en ese orden y separados por un espacio (data-model §2 de H7.4).
func TestTextoDeLaRedaccion(t *testing.T) {
	t.Parallel()

	assert.Equal(t, redaccionDeLaDA3, RedaccionEsperada{
		Norma: normaDeLaLCSP, Bloque: bloqueDeLaDA3,
		FechaVigencia: fechaDeLaRedaccionOriginal, FechaVigenciaReciente: fechaDeLaDA3Vigente,
	}.texto())
}
