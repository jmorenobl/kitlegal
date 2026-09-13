package boe

import (
	"strconv"
	"strings"
	"time"
	"unicode"
	"unicode/utf16"
	"unicode/utf8"
)

// Lo que fija la búsqueda de refs/boe.py (cmd_buscar, 252-271) y lo que dice su
// fallo (contrato errores-y-codigos §2, fila 4).
const (
	// limiteDeResultados es el limit=10 con el que buscar pide como mucho diez
	// resultados (refs/boe.py 252 y 271; FR-030).
	limiteDeResultados = 10
	// campoDelTitulo es el prefijo con el que cada palabra se exige en el título
	// de la norma (refs/boe.py 266).
	campoDelTitulo = "titulo:"
	// unionDeCondiciones es lo que une las condiciones de cada palabra
	// (refs/boe.py 266).
	unionDeCondiciones = " AND "
	// motivoSinPalabras es el mensaje de la búsqueda que no tiene ninguna
	// palabra.
	motivoSinPalabras = "la búsqueda no tiene ninguna palabra"
)

// Las constantes de la escritura de json.dumps y de quote.
const (
	// primerSustitutoDeByte es U+DC00, al que Python suma cada byte que no
	// forma UTF-8 válido al leer los argumentos de la orden (surrogateescape).
	primerSustitutoDeByte = 0xdc00
	// primerCaracterFueraDelPlanoBasico es U+10000, desde el que json.dumps
	// escribe un par de sustitutos (json/encoder.py 59-67).
	primerCaracterFueraDelPlanoBasico = 0x10000
	// hexadecimalEnMinusculas son los dígitos de '\\u{0:04x}' (json/encoder.py
	// 31, 60 y 67).
	hexadecimalEnMinusculas = "0123456789abcdef"
	// hexadecimalEnMayusculas son los dígitos de '%{:02X}' (urllib/parse.py 815).
	hexadecimalEnMayusculas = "0123456789ABCDEF"
	// seguroParaQuote es lo que quote deja sin codificar además de las letras y
	// los dígitos ASCII: el resto de _ALWAYS_SAFE y la barra, su safe por
	// omisión (urllib/parse.py 790-793 y 819).
	seguroParaQuote = "_.-~/"
)

// consultaDeBusqueda construye la consulta que buscar envía para un texto, el que
// resulta de unir sus argumentos con un espacio (refs/boe.py 784), como
// cmd_buscar (refs/boe.py 258-268) y con las dos adaptaciones de FR-030:
//
//   - si el texto lleva « AND », « OR », « NOT », titulo:, materia: o una
//     comilla doble, distinguiendo mayúsculas, ya es una consulta y va tal cual,
//     sin recortar;
//   - si no, cada palabra se exige en el título como titulo:<palabra> y las
//     condiciones se unen con « AND », también cuando hay una sola, que así va
//     sin el espacio en blanco que la rodea;
//   - y un texto sin ninguna palabra es un *Error de clase «argumentos», sin
//     dirección ni instante: no busca nada, y quien la llama lo hace antes de
//     abrir la caché y de pedir, así que un código 2 no deja rastro en disco ni
//     en red (contrato errores-y-codigos, fila 4).
//
// Las palabras son las de str.split sin argumentos: lo que queda entre los tramos
// de espacio en blanco de Python (esEspacioDePython).
func consultaDeBusqueda(texto string) (string, error) {
	if llevaOperadores(texto) {
		return texto, nil
	}

	palabras := strings.FieldsFunc(texto, esEspacioDePython)
	if len(palabras) == 0 {
		return "", errorDeArgumentos("", time.Time{}, nil, motivoSinPalabras)
	}

	condiciones := make([]string, len(palabras))
	for indice, palabra := range palabras {
		condiciones[indice] = campoDelTitulo + palabra
	}

	return strings.Join(condiciones, unionDeCondiciones), nil
}

// direccionDeBusqueda es la dirección con la que buscar pide la consulta de un
// texto (refs/boe.py 270-271; contrato verbos-y-salidas §1):
//
//	<base>?limit=10&query=<quote(json.dumps({"query": {"query_string": {"query": <consulta>}}}))>
//
// Es byte a byte la de boe.py, porque es la que identifica la búsqueda ante la
// fuente, en la caché, en las grabaciones y en el sobre (research.md D8). Un texto
// sin ninguna palabra da el error de consultaDeBusqueda y ninguna dirección.
func direccionDeBusqueda(texto string) (string, error) {
	consulta, err := consultaDeBusqueda(texto)
	if err != nil {
		return "", err
	}

	return baseDeLaAPI + "?limit=" + strconv.Itoa(limiteDeResultados) +
		"&query=" + quoteComoPython(cuerpoDeLaBusqueda(consulta)), nil
}

// llevaOperadores dice si el texto contiene alguno de _ES_OPERATORS
// (refs/boe.py 260-261), con sus espacios y distinguiendo mayúsculas.
func llevaOperadores(texto string) bool {
	for _, operador := range [...]string{" AND ", " OR ", " NOT ", "titulo:", "materia:", `"`} {
		if strings.Contains(texto, operador) {
			return true
		}
	}

	return false
}

// esEspacioDePython dice si el carácter es espacio en blanco para str.split y
// str.strip de Python: el de Unicode, que es el de unicode.IsSpace, y además los
// separadores de información U+001C a U+001F (supuesto (a) de doc.go).
func esEspacioDePython(caracter rune) bool {
	return unicode.IsSpace(caracter) || ('\x1c' <= caracter && caracter <= '\x1f')
}

// cuerpoDeLaBusqueda es json.dumps({"query": {"query_string": {"query":
// consulta}}}) tal como lo escribe Python con sus opciones por omisión
// (refs/boe.py 270): «: » entre cada clave y su valor (json/encoder.py 103) y, con
// una sola clave por objeto, ningún separador entre elementos; y la consulta como
// cadena con ensure_ascii (cadenaJSONComoPython). encoding/json no lo reproduce:
// escapa <, > y &, deja lo no ASCII sin escapar y no pone espacio tras los
// separadores (research.md D8).
func cuerpoDeLaBusqueda(consulta string) string {
	return `{"query": {"query_string": {"query": ` + cadenaJSONComoPython(consulta) + `}}}`
}

// cadenaJSONComoPython escribe el texto como cadena JSON, entre comillas dobles,
// como encode_basestring_ascii, la que json.dumps usa con ensure_ascii, su valor
// por omisión (json/__init__.py 183; json/encoder.py 18-31 y 49-68): el ASCII
// imprimible va tal cual salvo la comilla doble y la barra invertida, que van
// escapadas; \b, \f, \n, \r y \t, con su escape corto; y cualquier otro carácter
// —el resto de los controles, el de borrado y todo lo que no es ASCII— como \u y
// cuatro dígitos hexadecimales en minúsculas, con un par de sustitutos fuera del
// plano básico.
//
// Cada byte que no forma UTF-8 válido va como \udcXX, el sustituto con el que
// Python representa ese byte al leer los argumentos de la orden en Unix
// (surrogateescape): es la cadena que boe.py escribiría con los mismos bytes.
func cadenaJSONComoPython(texto string) string {
	cadena := make([]byte, 0, len(texto)+2)
	cadena = append(cadena, '"')

	for resto := texto; resto != ""; {
		caracter, medida := utf8.DecodeRuneInString(resto)
		if caracter == utf8.RuneError && medida == 1 {
			caracter = primerSustitutoDeByte + rune(resto[0])
		}

		cadena = agregaCaracterJSON(cadena, caracter)
		resto = resto[medida:]
	}

	return string(append(cadena, '"'))
}

// agregaCaracterJSON añade a la cadena un carácter escrito como lo escribe
// encode_basestring_ascii (json/encoder.py 49-68).
func agregaCaracterJSON(cadena []byte, caracter rune) []byte {
	if escape, conEscapeCorto := escapeCortoJSON(caracter); conEscapeCorto {
		return append(cadena, escape...)
	}

	switch {
	case ' ' <= caracter && caracter <= '~':
		return utf8.AppendRune(cadena, caracter)
	case caracter < primerCaracterFueraDelPlanoBasico:
		return agregaUnidadUTF16(cadena, caracter)
	default:
		alto, bajo := utf16.EncodeRune(caracter)

		return agregaUnidadUTF16(agregaUnidadUTF16(cadena, alto), bajo)
	}
}

// escapeCortoJSON es ESCAPE_DCT sin los controles que escribe como \u
// (json/encoder.py 21-29): el escape de la barra invertida, de la comilla doble y
// de los cinco controles que tienen uno corto.
func escapeCortoJSON(caracter rune) (string, bool) {
	switch caracter {
	case '\\':
		return `\\`, true
	case '"':
		return `\"`, true
	case '\b':
		return `\b`, true
	case '\f':
		return `\f`, true
	case '\n':
		return `\n`, true
	case '\r':
		return `\r`, true
	case '\t':
		return `\t`, true
	}

	return "", false
}

// agregaUnidadUTF16 añade \u y la unidad en cuatro dígitos hexadecimales en
// minúsculas, '\\u{0:04x}' (json/encoder.py 31, 60 y 67).
func agregaUnidadUTF16(cadena []byte, unidad rune) []byte {
	return append(cadena, '\\', 'u',
		hexadecimalEnMinusculas[(unidad>>12)&0xf], hexadecimalEnMinusculas[(unidad>>8)&0xf],
		hexadecimalEnMinusculas[(unidad>>4)&0xf], hexadecimalEnMinusculas[unidad&0xf])
}

// quoteComoPython es urllib.parse.quote con su safe por omisión, la barra
// (urllib/parse.py 819 y 890): cada byte que no es una letra o un dígito ASCII ni
// uno de _ . - ~ / va como %XX en mayúsculas (urllib/parse.py 815). Ninguna
// función de net/url lo reproduce: QueryEscape escribe el espacio como + y
// PathEscape deja sin codificar : @ & = + y $ (research.md D8).
func quoteComoPython(texto string) string {
	citado := make([]byte, 0, len(texto))

	for indice := range len(texto) {
		octeto := texto[indice]
		if esLetraODigitoASCII(octeto) || strings.IndexByte(seguroParaQuote, octeto) >= 0 {
			citado = append(citado, octeto)

			continue
		}

		citado = append(citado, '%', hexadecimalEnMayusculas[octeto>>4], hexadecimalEnMayusculas[octeto&0xf])
	}

	return string(citado)
}
