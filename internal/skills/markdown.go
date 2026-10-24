package skills

import (
	"fmt"
	"path"
	"regexp"
	"strings"
)

// Lo que los gates de references/ leen de un fichero Markdown
// (contracts/gates-de-skills-check.md §2.3; research D14; FR-007 a FR-009 de
// H31): los bloques de código, los enlaces fuera de ellos, los apartados «## » y
// el índice de una referencia larga. No es un analizador de Markdown (V7): son
// las tres formas que escriben las skills.
const (
	// sangriaMaximaDeLaValla y largoMinimoDeLaValla son la sangría con la que una
	// línea todavía abre una valla de un bloque de código y las marcas que la
	// forman.
	sangriaMaximaDeLaValla = 3
	largoMinimoDeLaValla   = 3

	// marcadorDeApartado abre la línea de un apartado de nivel 2.
	marcadorDeApartado = "## "

	// maximoDeLineasSinIndice son las líneas que una referencia puede tener sin
	// abrir con su índice (FR-009 de H31).
	maximoDeLineasSinIndice = 100
)

// enlaceDeMarkdown es un [texto](destino) dentro de una línea: el destino llega
// hasta el primer ) o espacio, que tiene que estar.
var enlaceDeMarkdown = regexp.MustCompile(`\[[^\]]*\]\(([^)\s]*)[)\s]`)

// destinoDeUnEnlace es la parte «](…)» de un enlace, que el índice no compara.
var destinoDeUnEnlace = regexp.MustCompile(`\]\([^)]*\)`)

// sinMarcasDeMarkdown quita de un texto lo que el índice no compara de un
// elemento ni de un apartado: el corchete que abre un enlace, el énfasis y el
// código.
var sinMarcasDeMarkdown = strings.NewReplacer("[", "", "*", "", "_", "", "`", "")

// elementoDeLista es una línea de una lista: tras su sangría, «- », «* » o un
// número y «. », y el texto del elemento.
var elementoDeLista = regexp.MustCompile(`^[ \t]*(?:[-*]|[0-9]+\.) (.*)$`)

// lineasFueraDeBloques son las líneas del contenido, sin su salto, que no están
// dentro de un bloque de código ni lo abren o lo cierran, en su orden. Un bloque
// va de la línea que abre una valla a la que la cierra con la misma marca
// —la misma, repetida al menos tantas veces y sin nada detrás—, o hasta el
// final si ninguna la cierra.
func lineasFueraDeBloques(contenido []byte) []string {
	var (
		fuera   []string
		abierta string
	)

	for linea := range strings.Lines(string(contenido)) {
		linea = strings.TrimSuffix(linea, "\n")
		valla, resto := vallaDeLaLinea(linea)

		switch {
		case abierta == "" && valla != "":
			abierta = valla
		case abierta == "":
			fuera = append(fuera, linea)
		case valla != "" && valla[0] == abierta[0] && len(valla) >= len(abierta) && strings.TrimSpace(resto) == "":
			abierta = ""
		}
	}

	return fuera
}

// vallaDeLaLinea es la valla que forma la línea —tras hasta tres espacios, tres
// o más acentos graves o virgulillas, todas iguales— y lo que la sigue, o nada
// si la línea no la forma.
func vallaDeLaLinea(linea string) (string, string) {
	tras := strings.TrimLeft(linea, " ")
	if len(linea)-len(tras) > sangriaMaximaDeLaValla || tras == "" || (tras[0] != '`' && tras[0] != '~') {
		return "", ""
	}

	resto := strings.TrimLeft(tras, tras[:1])

	valla := tras[:len(tras)-len(resto)]
	if len(valla) < largoMinimoDeLaValla {
		return "", ""
	}

	return valla, resto
}

// enlacesDeMarkdown son los destinos de los enlaces [texto](destino) del
// contenido que están fuera de sus bloques de código, en su orden y sin lo que
// sigue a #: el que solo lleva a un apartado, sin otro destino, no da ninguno.
func enlacesDeMarkdown(contenido []byte) []string {
	var destinos []string

	for _, linea := range lineasFueraDeBloques(contenido) {
		for _, enlace := range enlaceDeMarkdown.FindAllStringSubmatch(linea, -1) {
			if destino, _, _ := strings.Cut(enlace[1], "#"); destino != "" {
				destinos = append(destinos, destino)
			}
		}
	}

	return destinos
}

// ficherosEnlazados son las rutas a las que llevan los enlaces del contenido de
// un fichero del directorio dado: cada destino de enlacesDeMarkdown resuelto
// desde el directorio con path.Join (FR-008 de H31), en su orden.
func ficherosEnlazados(contenido []byte, directorio string) []string {
	destinos := enlacesDeMarkdown(contenido)
	rutas := make([]string, 0, len(destinos))

	for _, destino := range destinos {
		rutas = append(rutas, path.Join(directorio, destino))
	}

	return rutas
}

// faltaDelIndice es lo que le falta a una referencia de más de 100 líneas para
// abrir con su índice (FR-009 de H31; Clarificación Q5), con sus líneas y el
// límite, o nada si no le falta nada: algún apartado —una línea fuera de los
// bloques de código que empieza por «## »— y, antes del primero, una lista en la
// que, por orden, cada apartado tenga un elemento que contenga su texto, los dos
// sin la sintaxis de enlace, el énfasis ni el código. Los elementos de más no
// estorban. Con 100 líneas o menos no se exige nada.
func faltaDelIndice(contenido []byte) string {
	lineas := ContarLineas(contenido)
	if lineas <= maximoDeLineasSinIndice {
		return ""
	}

	var apartados, elementos []string

	for _, linea := range lineasFueraDeBloques(contenido) {
		if apartado, es := strings.CutPrefix(linea, marcadorDeApartado); es {
			apartados = append(apartados, textoComparable(apartado))

			continue
		}

		if elemento := elementoDeLista.FindStringSubmatch(linea); elemento != nil && len(apartados) == 0 {
			elementos = append(elementos, textoComparable(elemento[1]))
		}
	}

	switch {
	case len(apartados) == 0:
		return fmt.Sprintf("%d líneas sin ningún apartado %s(más de %d)", lineas, marcadorDeApartado,
			maximoDeLineasSinIndice)
	case !nombraCadaApartado(elementos, apartados):
		return fmt.Sprintf("%d líneas sin un índice que nombre sus apartados %santes del primero (más de %d)", lineas,
			marcadorDeApartado, maximoDeLineasSinIndice)
	}

	return ""
}

// nombraCadaApartado dice si los elementos nombran, por orden, cada apartado:
// cada uno, el primer elemento que contiene su texto tras el del anterior.
func nombraCadaApartado(elementos, apartados []string) bool {
	siguiente := 0

	for _, apartado := range apartados {
		for siguiente < len(elementos) && !strings.Contains(elementos[siguiente], apartado) {
			siguiente++
		}

		if siguiente == len(elementos) {
			return false
		}

		siguiente++
	}

	return true
}

// textoComparable es el texto de un elemento o de un apartado como lo compara el
// índice: sin «](…)», «[», «*», «_» ni «`», y sin los blancos de los extremos.
func textoComparable(texto string) string {
	return strings.TrimSpace(sinMarcasDeMarkdown.Replace(destinoDeUnEnlace.ReplaceAllString(texto, "")))
}
