package territorio

import (
	"slices"
	"strings"
	"unicode"
)

// Lo que el pliegue y las formas del nombre reconocen en un nombre oficial.
const (
	// puntoVolado es el de la ele geminada catalana, «l·l»: une la ele de cada
	// lado, así que el pliegue lo quita en lugar de separar por él.
	puntoVolado = '·'
	// separadorBilingue parte el nombre oficial de un municipio bilingüe en sus
	// dos lados, como en «Nombre/Izena».
	separadorBilingue = "/"
	// separadorDelArticulo es el que el INE pone delante de un artículo
	// pospuesto, como en «Nombre, El».
	separadorDelArticulo = ","
)

// Plegar lleva un nombre a la forma en la que se comparan la consulta y los
// nombres de la relación (data-model §2.7, research.md D10): en minúsculas;
// cada letra con diacrítico en su letra base —á, à, ä y â en a, y así las
// demás vocales; ñ en n; ç en c— y sin marcas combinantes; sin el punto volado
// de «l·l», que une la ele de cada lado; y los espacios y los signos de
// puntuación, colapsados en un espacio entre palabras y quitados al principio
// y al final.
//
// Lo que la tabla no cubre no se aproxima: queda en minúscula, fuera del
// alfabeto del pliegue —minúsculas y cifras ASCII y el espacio—, que es como
// el control del corpus congelado ve una runa nueva (research.md D27). Plegar
// lo plegado no lo cambia.
//
// Es la única definición del pliegue en el árbol: la usan el registro, el
// juicio de evals y el control del corpus.
func Plegar(nombre string) string {
	var plegado strings.Builder
	plegado.Grow(len(nombre))

	separar := false

	for _, runa := range nombre {
		letra := unicode.ToLower(runa)

		switch {
		case letra == puntoVolado, unicode.Is(unicode.Mn, letra):
			// Ni letra ni separador: el diacrítico suelto se quita y el punto
			// volado une lo que tiene a los lados.
		case unicode.IsSpace(letra), unicode.IsPunct(letra):
			separar = plegado.Len() > 0
		default:
			if separar {
				plegado.WriteByte(' ')
				separar = false
			}

			plegado.WriteRune(letraBase(letra))
		}
	}

	return plegado.String()
}

// letraBase es la tabla del pliegue: devuelve la letra base de una minúscula
// con diacrítico de las lenguas de España, o la propia letra si no lo lleva o
// si la tabla no la cubre.
func letraBase(letra rune) rune {
	switch letra {
	case 'á', 'à', 'ä', 'â':
		return 'a'
	case 'é', 'è', 'ë', 'ê':
		return 'e'
	case 'í', 'ì', 'ï', 'î':
		return 'i'
	case 'ó', 'ò', 'ö', 'ô':
		return 'o'
	case 'ú', 'ù', 'ü', 'û':
		return 'u'
	case 'ñ':
		return 'n'
	case 'ç':
		return 'c'
	default:
		return letra
	}
}

// formasDelNombre devuelve las formas conocidas de un municipio, plegadas y sin
// repetir, derivadas solo de su nombre oficial y sin ningún caso especial
// (FR-015, FR-024, data-model §2.7):
//
//   - el nombre oficial tal como lo escribe el INE, «Nombre, El»;
//   - con el artículo pospuesto antepuesto, «El Nombre»;
//   - cada lado de un nombre bilingüe partido por «/» —«Nombre» e «Izena» de
//     «Nombre/Izena»—, y la forma con el artículo antepuesto de cada uno.
//
// Una forma que se pliega a nada no es una forma: ninguna consulta la
// alcanzaría.
func formasDelNombre(oficial string) []string {
	var formas []string

	anadir := func(forma string) {
		if plegada := Plegar(forma); plegada != "" && !slices.Contains(formas, plegada) {
			formas = append(formas, plegada)
		}
	}

	anadir(oficial)

	for lado := range strings.SplitSeq(oficial, separadorBilingue) {
		anadir(lado)

		if antepuesto, pospuesto := conArticuloAntepuesto(lado); pospuesto {
			anadir(antepuesto)
		}
	}

	return formas
}

// conArticuloAntepuesto devuelve el nombre con el artículo delante si el INE lo
// escribe pospuesto: una sola palabra tras la última coma, como «Nombre, El» o
// «Nombre, L'». Una coma seguida de varias palabras no pospone un artículo
// —«Núcleo, Otro Núcleo i Tercero» es el nombre de un municipio de varios
// núcleos— y el nombre se queda como está.
func conArticuloAntepuesto(nombre string) (string, bool) {
	coma := strings.LastIndex(nombre, separadorDelArticulo)
	if coma < 0 {
		return "", false
	}

	articulo := strings.TrimSpace(nombre[coma+len(separadorDelArticulo):])
	if articulo == "" || strings.ContainsFunc(articulo, unicode.IsSpace) {
		return "", false
	}

	return articulo + " " + strings.TrimSpace(nombre[:coma]), true
}
