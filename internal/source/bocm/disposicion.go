package bocm

import (
	"encoding/xml"
	"errors"
	"fmt"
	"slices"
	"strconv"
	"strings"
	"time"
	"unicode"

	"github.com/jmorenobl/kitlegal/internal/core/plazos"
)

// Lo que fija la forma de una disposición de fiestas locales en el XML del
// BOCM y la gramática de sus líneas (contracts/datos-de-plazos.md §1.3;
// research D13, V8).
const (
	// principioDeLinea es con lo que empieza, en el texto, la línea de cada
	// municipio o entidad: un punto y un tabulador.
	principioDeLinea = ".\t"
	// separadorDelNombre separa el nombre del cuerpo cuando la línea lo lleva;
	// sin él, el nombre llega hasta la primera cifra.
	separadorDelNombre = ":"
	// puntoSuelto es lo que el cuerpo tolera suelto, delante de una cifra o
	// antes del final: se lee como un blanco.
	puntoSuelto = "."
	// noComunicado es el cuerpo de la línea de un municipio que no comunicó sus
	// fiestas, en minúscula: se compara sin distinguir mayúsculas.
	noComunicado = "no comunicado por el ayuntamiento"
	// nexo une los dos días de un mes y los grupos de fechas.
	nexo = "y"
	// preposicionDelMes va, si va, entre el día y el mes.
	preposicionDelMes = "de"
	// principioDelAnio precede, en el título de los metadatos, al año cuyas
	// fiestas declara o corrige la disposición.
	principioDelAnio = "para el año "
	// cifrasDelAnio son las del año en el título.
	cifrasDelAnio = 4
	// diaMayor es el mayor día de un mes.
	diaMayor = 31
	// anioBisiesto es un año con el 29 de febrero, con el que se comprueba que
	// un día existe en su mes antes de saber el año de la disposición.
	anioBisiesto = 2000
)

// nombresDeLosMeses son los meses en minúscula, de enero a diciembre: el índice
// más uno es el time.Month.
var nombresDeLosMeses = []string{
	"enero", "febrero", "marzo", "abril", "mayo", "junio",
	"julio", "agosto", "septiembre", "octubre", "noviembre", "diciembre",
}

// Resolucion es una disposición leída: la resolución que declara las fiestas
// locales de un año o una que la modifica.
type Resolucion struct {
	// ID es su identificador, BOCM-AAAAMMDD-N, el de sus metadatos.
	ID string
	// Anio es el que dice su título, «para el año <aaaa>».
	Anio int
	// Lineas son las líneas de municipio o entidad del texto, en su orden.
	Lineas []Linea
}

// Linea es una línea `.⇥<nombre>[:] <cuerpo>[.]` del texto, ya leída: da
// festivos, o es la de un municipio que no comunicó sus fiestas, o no se lee
// sin interpretarla (FR-023; ADR 0039, punto 3).
type Linea struct {
	// Nombre es el del municipio o la entidad tal como lo escribe el boletín.
	Nombre string
	// Cuerpo es lo que sigue al nombre, tal cual.
	Cuerpo string
	// Fechas son las fiestas de la línea, en su orden, solo si da festivos.
	Fechas []Fecha
	// Falta es plazos.FaltaNoComunicado o plazos.FaltaLineaSinLeer, o vacía si
	// la línea da festivos.
	Falta plazos.Falta
}

// Fecha es un día y un mes: el año lo pone la disposición en el manifiesto.
type Fecha struct {
	Dia int
	Mes time.Month
}

// documento es lo que se lee del XML de una disposición.
type documento struct {
	Identificador string `xml:"metadatos>identificador"`
	Titulo        string `xml:"metadatos>titulo"`
	Texto         string `xml:"texto"`
}

// LeerDisposicion lee una disposición de fiestas locales del XML del BOCM tal
// como lo graba grabar_datos: su identificador y el año de su título, de los
// metadatos, y cada línea `.⇥<nombre>[:] <cuerpo>[.]` del texto con la
// gramática de research D13. Lo que no tiene la forma de una disposición —sin
// XML, sin identificador, sin el año en el título, sin ninguna línea— es un
// error: no hay nada que leer. La línea que no se lee sin interpretarla queda
// como tal, con su motivo (FR-023).
func LeerDisposicion(contenido []byte) (Resolucion, error) {
	var doc documento
	if err := xml.Unmarshal(contenido, &doc); err != nil {
		return Resolucion{}, fmt.Errorf("el documento no es una disposición: no se lee como XML: %w", err)
	}

	if doc.Identificador == "" {
		return Resolucion{}, errors.New("el documento no es una disposición: no tiene identificador en sus metadatos")
	}

	anio, err := anioDelTitulo(doc.Titulo)
	if err != nil {
		return Resolucion{}, fmt.Errorf("la disposición %s: %w", doc.Identificador, err)
	}

	var lineas []Linea

	for linea := range strings.SplitSeq(doc.Texto, "\n") {
		if resto, esLinea := strings.CutPrefix(strings.TrimSuffix(linea, "\r"), principioDeLinea); esLinea {
			lineas = append(lineas, leerLinea(resto))
		}
	}

	if len(lineas) == 0 {
		return Resolucion{}, fmt.Errorf("la disposición %s no tiene ninguna línea de municipio (%q)",
			doc.Identificador, principioDeLinea)
	}

	return Resolucion{ID: doc.Identificador, Anio: anio, Lineas: lineas}, nil
}

// anioDelTitulo es el año que sigue a «para el año» en el título.
func anioDelTitulo(titulo string) (int, error) {
	_, resto, tiene := strings.Cut(titulo, principioDelAnio)
	if !tiene || len(resto) < cifrasDelAnio {
		return 0, errors.New("el título no dice de qué año es («para el año <aaaa>»)")
	}

	anio, err := strconv.Atoi(resto[:cifrasDelAnio])
	if err != nil || anio <= 0 {
		return 0, fmt.Errorf("el título no dice de qué año es: tras «para el año» va %q", resto[:cifrasDelAnio])
	}

	return anio, nil
}

// leerLinea lee lo que sigue al principio de una línea: el nombre, hasta los
// dos puntos o, sin ellos, hasta la primera cifra, y el cuerpo, que es «no
// comunicado por el ayuntamiento» o una lista de fechas; lo demás no se lee.
func leerLinea(texto string) Linea {
	nombre, cuerpo := partirLinea(texto)
	linea := Linea{Nombre: nombre, Cuerpo: cuerpo}

	palabras := strings.Fields(strings.ToLower(strings.ReplaceAll(cuerpo, puntoSuelto, " ")))
	if strings.Join(palabras, " ") == noComunicado {
		linea.Falta = plazos.FaltaNoComunicado

		return linea
	}

	fechas, leidas := leerFechas(palabras)
	if !leidas {
		linea.Falta = plazos.FaltaLineaSinLeer

		return linea
	}

	linea.Fechas = fechas

	return linea
}

// partirLinea separa el nombre del cuerpo.
func partirLinea(texto string) (nombre, cuerpo string) {
	if antes, despues, conSeparador := strings.Cut(texto, separadorDelNombre); conSeparador {
		return strings.TrimSpace(antes), despues
	}

	if cifra := strings.IndexFunc(texto, unicode.IsDigit); cifra >= 0 {
		return strings.TrimSpace(texto[:cifra]), texto[cifra:]
	}

	return strings.TrimSpace(texto), ""
}

// lectura avanza por las palabras del cuerpo de una línea.
type lectura struct {
	palabras []string
	posicion int
}

// palabra es la que está a esa distancia de la posición, o vacía si no la hay.
func (l *lectura) palabra(distancia int) string {
	if l.posicion+distancia >= len(l.palabras) {
		return ""
	}

	return l.palabras[l.posicion+distancia]
}

// avanzar consume palabras.
func (l *lectura) avanzar(cuantas int) {
	l.posicion += cuantas
}

// acabada dice si no queda ninguna palabra.
func (l *lectura) acabada() bool {
	return l.posicion >= len(l.palabras)
}

// leerFechas lee una lista de grupos `<d>[ y <d>] [de ]<mes>` unidos por «y»,
// ya sin puntos sueltos y en minúscula, y dice si la ha leído entera.
func leerFechas(palabras []string) ([]Fecha, bool) {
	l := &lectura{palabras: palabras}

	var fechas []Fecha

	for {
		grupo, leido := l.grupo()
		if !leido {
			return nil, false
		}

		fechas = append(fechas, grupo...)

		if l.acabada() {
			return fechas, true
		}

		if l.palabra(0) != nexo {
			return nil, false
		}

		l.avanzar(1)
	}
}

// grupo lee `<d>[ y <d>] [de ]<mes>`: uno o dos días que existen en ese mes.
func (l *lectura) grupo() ([]Fecha, bool) {
	primero, esDia := diaDe(l.palabra(0))
	if !esDia {
		return nil, false
	}

	dias := []int{primero}

	l.avanzar(1)

	if l.palabra(0) == nexo {
		if segundo, esSegundo := diaDe(l.palabra(1)); esSegundo {
			dias = append(dias, segundo)

			l.avanzar(2)
		}
	}

	if l.palabra(0) == preposicionDelMes {
		l.avanzar(1)
	}

	mes, esMes := mesDe(l.palabra(0))
	if !esMes {
		return nil, false
	}

	l.avanzar(1)

	fechas := make([]Fecha, 0, len(dias))

	for _, dia := range dias {
		if !existe(dia, mes) {
			return nil, false
		}

		fechas = append(fechas, Fecha{Dia: dia, Mes: mes})
	}

	return fechas, true
}

// diaDe es el día que escribe una palabra, si es toda cifras y va del 1 al 31.
func diaDe(palabra string) (int, bool) {
	if palabra == "" || strings.ContainsFunc(palabra, func(r rune) bool { return r < '0' || r > '9' }) {
		return 0, false
	}

	dia, err := strconv.Atoi(palabra)
	if err != nil || dia < 1 || dia > diaMayor {
		return 0, false
	}

	return dia, true
}

// mesDe es el mes que nombra una palabra en minúscula, si es uno.
func mesDe(palabra string) (time.Month, bool) {
	indice := slices.Index(nombresDeLosMeses, palabra)
	if indice < 0 {
		return 0, false
	}

	return time.Month(indice + 1), true
}

// existe dice si el mes tiene ese día en algún año.
func existe(dia int, mes time.Month) bool {
	return time.Date(anioBisiesto, mes, dia, 0, 0, 0, 0, time.UTC).Day() == dia
}
