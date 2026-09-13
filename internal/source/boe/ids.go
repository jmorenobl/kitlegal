package boe

import (
	"fmt"
	"strings"
	"time"
	"unicode"
	"unicode/utf8"
)

// Las dos gramáticas de entrada de data-model.md §5, escritas como constantes
// para que el mensaje de «argumentos» y la comprobación no puedan decir cosas
// distintas.
const (
	// prefijoDeNorma es lo que tienen delante todas las normas de la API de
	// Legislación Consolidada, también las autonómicas (FR-081).
	prefijoDeNorma = "BOE-A-"
	// digitosDelAnio y maximoDeDigitosDelNumero son las dos partes numéricas de
	// la norma: ^BOE-A-[0-9]{4}-[0-9]{1,9}$.
	digitosDelAnio           = 4
	maximoDeDigitosDelNumero = 9
	// maximoDeCaracteresDelBloque acota el id de bloque, ^[A-Za-z0-9]{1,64}$: 64
	// cubre con holgura los ids del BOE (a21, da3, preambulo, a108bis) y deja
	// fuera la longitud desmedida (FR-080, research.md D9).
	maximoDeCaracteresDelBloque = 64
)

// ValidarNorma comprueba que la norma tenga la forma BOE-A-<año>-<número>, con
// cuatro dígitos ASCII en el año y de uno a nueve en el número, y distinguiendo
// mayúsculas. Fuera de esa gramática devuelve un *Error de clase «argumentos»
// que nombra el valor recibido y la forma esperada, sin dirección ni instante:
// quien la llama lo hace antes de abrir la caché y de construir ninguna
// petición, así que un código 2 no deja rastro en disco ni en red (FR-081,
// data-model.md §5).
//
// Se aparta de refs/boe.py, que concatena la norma en la ruta sin comprobarla
// (entrada 28 del porte anotado en doc.go).
func ValidarNorma(norma string) error {
	if !esNormaValida(norma) {
		return errorDeArgumentos("", time.Time{}, nil, fmt.Sprintf(
			"la norma %q no tiene la forma BOE-A-<año>-<número>, "+
				"con cuatro dígitos en el año y de uno a nueve en el número", norma))
	}

	return nil
}

// ValidarBloque comprueba que el id de bloque tenga de 1 a 64 caracteres, todos
// letras o dígitos ASCII: la forma de los ids del índice de la fuente, y nada de
// lo que podría alterar la petición en la que va como segmento —separadores de
// ruta, ?, #, %, espacios, controles—. Fuera de esa gramática devuelve un *Error
// de clase «argumentos» que nombra el valor recibido y la forma esperada, sin
// dirección ni instante, antes de abrir la caché y de construir ninguna
// petición (FR-080, data-model.md §5).
//
// Se aparta de refs/boe.py, que concatena el id en la ruta sin comprobarlo
// (entrada 28 del porte anotado en doc.go).
func ValidarBloque(bloque string) error {
	if bloque == "" || len(bloque) > maximoDeCaracteresDelBloque || !sonLetrasODigitosASCII(bloque) {
		return errorDeArgumentos("", time.Time{}, nil, fmt.Sprintf(
			"el bloque %q no tiene la forma de un id de bloque: "+
				"de 1 a %d caracteres, todos letras o dígitos ASCII, como a21, da3 o preambulo",
			bloque, maximoDeCaracteresDelBloque))
	}

	return nil
}

// Los tipos que TipoDesdeID infiere, con los nombres de refs/boe.py y en el
// orden de sus reglas. sinTipo es el de un id que no casa con ninguna.
const (
	tipoArticulo               = "articulo"
	tipoTitulo                 = "titulo"
	tipoCapitulo               = "capitulo"
	tipoSeccion                = "seccion"
	tipoPreambulo              = "preambulo"
	tipoDisposicionAdicional   = "disposicion_adicional"
	tipoDisposicionTransitoria = "disposicion_transitoria"
	tipoDisposicionDerogatoria = "disposicion_derogatoria"
	tipoDisposicionFinal       = "disposicion_final"
	sinTipo                    = ""
)

// TipoDesdeID infiere el tipo de un bloque a partir de su id, con el porte
// literal de _tipo_from_id (refs/boe.py 155-176): sobre el id en minúsculas
// Unicode, mirando runas y con «dígito» como unicode.IsDigit, la primera de estas
// diez reglas que casa decide (data-model.md §4):
//
//  1. primera runa a y segunda un dígito → articulo
//  2. empieza por t → titulo
//  3. empieza por ci o cv, o primera runa c y segunda una de i v x l c d m → capitulo
//  4. empieza por s y la segunda runa es un dígito o e → seccion
//  5. es exactamente preambulo → preambulo
//  6. empieza por da → disposicion_adicional
//  7. empieza por dt → disposicion_transitoria
//  8. empieza por dd → disposicion_derogatoria
//  9. empieza por df → disposicion_final
//  10. ninguna de las anteriores → vacío
//
// Se portan también sus rarezas: ti es titulo porque la regla 2 va antes, cv3 es
// capitulo y subseccion no la produce ninguna regla (entradas 24 y 25 del porte
// anotado en doc.go). La usa indice con los ids que entrega la fuente, que no
// pasan por ValidarBloque, así que acepta cualquier cadena (FR-040).
func TipoDesdeID(id string) string {
	bid := strings.ToLower(id)
	primera, segunda := dosPrimerasRunas(bid)

	switch {
	case primera == 'a' && unicode.IsDigit(segunda):
		return tipoArticulo
	case strings.HasPrefix(bid, "t"):
		return tipoTitulo
	case esCapitulo(bid, primera, segunda):
		return tipoCapitulo
	case esSeccion(bid, segunda):
		return tipoSeccion
	case bid == tipoPreambulo:
		return tipoPreambulo
	case strings.HasPrefix(bid, "da"):
		return tipoDisposicionAdicional
	case strings.HasPrefix(bid, "dt"):
		return tipoDisposicionTransitoria
	case strings.HasPrefix(bid, "dd"):
		return tipoDisposicionDerogatoria
	case strings.HasPrefix(bid, "df"):
		return tipoDisposicionFinal
	}

	return sinTipo
}

// sinRuna es lo que dosPrimerasRunas devuelve en lugar de una runa que la cadena
// no tiene. No es una runa válida, así que no es un dígito ni casa con ninguna
// letra: equivale al bid[1:2] vacío de refs/boe.py.
const sinRuna rune = -1

// esNormaValida es la gramática de la norma, ^BOE-A-[0-9]{4}-[0-9]{1,9}$.
func esNormaValida(norma string) bool {
	resto, conPrefijo := strings.CutPrefix(norma, prefijoDeNorma)
	anio, numero, conGuion := strings.Cut(resto, "-")

	return conPrefijo && conGuion &&
		len(anio) == digitosDelAnio && sonDigitosASCII(anio) &&
		numero != "" && len(numero) <= maximoDeDigitosDelNumero && sonDigitosASCII(numero)
}

// sonDigitosASCII dice si todos los bytes del texto son dígitos ASCII. Mira
// bytes y no runas a propósito: un dígito de otra escritura, que unicode.IsDigit
// sí acepta, no forma parte de ninguna gramática de entrada.
func sonDigitosASCII(texto string) bool {
	for indice := range len(texto) {
		if !esDigitoASCII(texto[indice]) {
			return false
		}
	}

	return true
}

// sonLetrasODigitosASCII dice si todos los bytes del texto son letras o dígitos
// ASCII. Un byte de una letra con tilde o de UTF-8 inválido no lo es.
func sonLetrasODigitosASCII(texto string) bool {
	for indice := range len(texto) {
		if !esLetraODigitoASCII(texto[indice]) {
			return false
		}
	}

	return true
}

// esLetraODigitoASCII dice si el byte es una letra ASCII, mayúscula o
// minúscula, o un dígito del 0 al 9.
func esLetraODigitoASCII(caracter byte) bool {
	return esDigitoASCII(caracter) || ('a' <= caracter && caracter <= 'z') || ('A' <= caracter && caracter <= 'Z')
}

// esDigitoASCII dice si el byte es un dígito del 0 al 9.
func esDigitoASCII(caracter byte) bool {
	return '0' <= caracter && caracter <= '9'
}

// dosPrimerasRunas devuelve las dos primeras runas del texto, y sinRuna en lugar
// de la que falte.
func dosPrimerasRunas(texto string) (primera, segunda rune) {
	primera, medida := utf8.DecodeRuneInString(texto)
	if medida == 0 {
		return sinRuna, sinRuna
	}

	segunda, medida = utf8.DecodeRuneInString(texto[medida:])
	if medida == 0 {
		return primera, sinRuna
	}

	return primera, segunda
}

// esCapitulo es la regla 3 tal como la escribe refs/boe.py, con las dos
// comprobaciones de prefijo que la tercera ya cubre: ci y cv empiezan por c
// seguida de una de i v x l c d m.
func esCapitulo(bid string, primera, segunda rune) bool {
	return strings.HasPrefix(bid, "ci") || strings.HasPrefix(bid, "cv") ||
		(primera == 'c' && strings.ContainsRune("ivxlcdm", segunda))
}

// esSeccion es la regla 4: empieza por s y la segunda runa es un dígito o e.
func esSeccion(bid string, segunda rune) bool {
	return strings.HasPrefix(bid, "s") && (unicode.IsDigit(segunda) || segunda == 'e')
}
