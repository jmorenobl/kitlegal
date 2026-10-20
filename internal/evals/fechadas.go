package evals

import (
	"regexp"
	"strings"
	"sync"
)

// FormaFechada es una de las tres formas fijas con las que la respuesta a una
// pregunta con fecha dice la redacción de un bloque en ese día
// (contracts/evals-redaccion-fechada.md §2 de H20; research D13; FR-041,
// FR-043). Su valor es la etiqueta de la forma, la que va entre la marca y los
// dos puntos.
type FormaFechada string

// Las tres formas fijas de una redacción fechada.
const (
	// RedaccionAFecha es la forma A, la de la redacción citada, la vigente el día
	// pedido (FR-041).
	RedaccionAFecha FormaFechada = "REDACCIÓN A FECHA"

	// RedaccionDeHoy es la forma B, la de la redacción de hoy cuando es otra
	// (FR-041).
	RedaccionDeHoy FormaFechada = "REDACCIÓN DE HOY"

	// SinRedaccionVigente es la forma C, la del día en que el bloque no tenía
	// redacción vigente (FR-043).
	SinRedaccionVigente FormaFechada = "SIN REDACCIÓN VIGENTE"
)

// LineaFechada es lo que se lee de una línea de la respuesta con una forma fija
// de una redacción fechada (ExtraerRedaccionesFechadas): su forma, la norma y el
// bloque de su primera cita y lo que lleva detrás de esa cita. Lo que la forma no
// lee, o la línea no lleva detrás de la cita, queda vacío.
type LineaFechada struct {
	// Forma es la forma fija de la línea.
	Forma FormaFechada

	// Norma es el identificador de la norma de su primera cita.
	Norma string

	// Bloque es el id del bloque de su primera cita.
	Bloque string

	// Fecha es, en las formas A y C, la primera fecha AAAA-MM-DD detrás de la
	// cita: el día pedido.
	Fecha string

	// FechaVigencia es la primera fecha AAAAMMDD detrás de la cita: en la forma A,
	// la de la redacción citada; en la B, la de la de hoy; y en la C, la de la
	// primera redacción del bloque.
	FechaVigencia string

	// NormaModificadora es, en la forma B, el primer identificador BOE-A-… detrás
	// de la cita: la norma que dio la redacción de hoy.
	NormaModificadora string
}

// formasFechadas son las tres formas, en el orden de su letra.
var formasFechadas = []FormaFechada{RedaccionAFecha, RedaccionDeHoy, SinRedaccionVigente}

// expresionesFechadas son las expresiones de la forma fija de cada una de las
// tres, con la gramática de las etiquetas de los avisos (formaFija): la marca,
// las palabras de la etiqueta sin distinguir mayúsculas, con los blancos y el
// énfasis de Markdown alrededor de las partes, y los dos puntos.
var expresionesFechadas = sync.OnceValue(func() map[FormaFechada]*regexp.Regexp {
	expresiones := make(map[FormaFechada]*regexp.Regexp, len(formasFechadas))

	for _, forma := range formasFechadas {
		expresiones[forma] = formaFija(string(forma))
	}

	return expresiones
})

// formaDeDia casa con una fecha AAAA-MM-DD sin otra cifra a ninguno de sus
// lados, como la de la línea ⚠ PLAZO: (formaDelPlazo). El grupo es la fecha.
var formaDeDia = regexp.MustCompile(`(?:^|[^0-9])([0-9]{4}-[0-9]{2}-[0-9]{2})(?:[^0-9]|$)`)

// ExtraerRedaccionesFechadas devuelve lo que se lee de cada línea de la
// respuesta con una de las tres formas fijas de una redacción fechada y al menos
// una cita (formaDeCita), en el orden de sus líneas, o nil si no lleva ninguna
// (contracts/evals-redaccion-fechada.md §2 de H20; data-model §7.3; research
// D13): su forma, la norma y el bloque de su primera cita y, detrás de esa
// cita, en las formas A y C, la primera fecha AAAA-MM-DD y la primera AAAAMMDD
// (fechasDeVigencia), y en la B, el primer identificador BOE-A-… y la primera
// AAAAMMDD. Lo que va delante de la cita no se lee, ni la cita, las fechas o el
// identificador de otra línea: una línea sin cita no da nada. Una línea con más
// de una forma tiene la que empieza antes.
func ExtraerRedaccionesFechadas(respuesta string) []LineaFechada {
	var lineas []LineaFechada

	for linea := range strings.Lines(respuesta) {
		forma, conForma := formaDeLaLinea(linea)
		if !conForma {
			continue
		}

		cita := formaDeCita.FindStringSubmatchIndex(linea)
		if cita == nil {
			continue
		}

		leida := LineaFechada{Forma: forma, Norma: linea[cita[2]:cita[3]], Bloque: linea[cita[4]:cita[5]]}
		detras := linea[cita[1]:]

		if fechas := fechasDeVigencia(detras); len(fechas) > 0 {
			leida.FechaVigencia = fechas[0]
		}

		if forma == RedaccionDeHoy {
			leida.NormaModificadora = formaDeIdentificador.FindString(detras)
		} else if dia := formaDeDia.FindStringSubmatch(detras); dia != nil {
			leida.Fecha = dia[1]
		}

		lineas = append(lineas, leida)
	}

	return lineas
}

// formaDeLaLinea es la forma fija de una redacción fechada que lleva la línea,
// la que empieza antes si lleva más de una, y si lleva alguna.
func formaDeLaLinea(linea string) (FormaFechada, bool) {
	var (
		laPrimera FormaFechada
		inicio    = -1
	)

	for _, forma := range formasFechadas {
		encontrada := expresionesFechadas()[forma].FindStringIndex(linea)
		if encontrada != nil && (inicio < 0 || encontrada[0] < inicio) {
			laPrimera, inicio = forma, encontrada[0]
		}
	}

	return laPrimera, inicio >= 0
}
