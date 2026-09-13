package boe

import (
	"encoding/json"
	"net/url"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/jmorenobl/kitlegal/internal/cli"
	"github.com/jmorenobl/kitlegal/internal/core/schema"
)

// Lo que json.dumps pone alrededor de la consulta en el cuerpo de la búsqueda
// —con los separadores «: » y «, » de Python—, y lo mismo pasado por quote. Los
// dos tests de este fichero los escriben aparte de la cadena de la consulta, que
// es lo que cambia de un caso a otro.
const (
	antesDeLaCadena         = `{"query": {"query_string": {"query": `
	despuesDeLaCadena       = `}}}`
	antesDeLaCadenaCitada   = `%7B%22query%22%3A%20%7B%22query_string%22%3A%20%7B%22query%22%3A%20`
	despuesDeLaCadenaCitada = `%7D%7D%7D`
)

// TestConsultaDeBusqueda fija la consulta que buscar envía, la de refs/boe.py
// 258-271 con las dos adaptaciones de FR-030: un texto con « AND », « OR »,
// « NOT », titulo:, materia: o una comilla doble va tal cual, sin recortar; si
// no, cada palabra —partida por el espacio en blanco de str.split, que además
// del de Unicode incluye U+001C a U+001F— va como titulo:<palabra>, unidas con
// « AND », también cuando es una sola; y sin ninguna palabra es «argumentos»,
// código 2, sin dirección. La dirección de «procedimiento administrativo común»
// es la del contrato verbos-y-salidas §1, byte a byte; la de los demás casos
// pide 10 resultados y lleva, una vez descodificada, esa misma consulta
// (research.md D8).
func TestConsultaDeBusqueda(t *testing.T) {
	t.Parallel()

	const (
		base = "https://www.boe.es/datosabiertos/api/legislacion-consolidada"
		// direccionDelContrato es la del ejemplo del contrato verbos-y-salidas §1,
		// copiada tal cual.
		direccionDelContrato = "https://www.boe.es/datosabiertos/api/legislacion-consolidada?limit=10&query=" +
			"%7B%22query%22%3A%20%7B%22query_string%22%3A%20%7B%22query%22%3A%20%22titulo%3Aprocedimiento%20AND%20" +
			"titulo%3Aadministrativo%20AND%20titulo%3Acom%5Cu00fan%22%7D%7D%7D"
	)

	casos := []struct {
		nombre    string
		texto     string
		consulta  string
		direccion string
	}{
		{nombre: "con AND va tal cual", texto: "ley AND tributaria", consulta: "ley AND tributaria"},
		{nombre: "con OR va tal cual", texto: "tributaria OR fiscal", consulta: "tributaria OR fiscal"},
		{nombre: "con NOT va tal cual", texto: "ley NOT orgánica", consulta: "ley NOT orgánica"},
		{nombre: "con titulo: va tal cual", texto: "titulo:presupuestos", consulta: "titulo:presupuestos"},
		{nombre: "con materia: va tal cual", texto: "materia:1234", consulta: "materia:1234"},
		{nombre: "con comilla doble va tal cual", texto: `"ley general tributaria"`, consulta: `"ley general tributaria"`},
		{nombre: "con operador no se recorta", texto: " \tley AND tributaria\n", consulta: " \tley AND tributaria\n"},
		{nombre: "un operador solo ya es una consulta", texto: " AND ", consulta: " AND "},
		{
			nombre:   "un operador en minúsculas no es operador",
			texto:    "ley and tributaria",
			consulta: "titulo:ley AND titulo:and AND titulo:tributaria",
		},
		{nombre: "AND sin espacios alrededor no es operador", texto: "AND", consulta: "titulo:AND"},
		{nombre: "Titulo: con mayúscula no es operador", texto: "Titulo:ley", consulta: "titulo:Titulo:ley"},
		{
			nombre:    "varias palabras, con la dirección del contrato",
			texto:     "procedimiento administrativo común",
			consulta:  "titulo:procedimiento AND titulo:administrativo AND titulo:común",
			direccion: direccionDelContrato,
		},
		{
			nombre:   "varias palabras con espacio en blanco repetido alrededor y entre ellas",
			texto:    "  ley\t\tgeneral \r\n tributaria  ",
			consulta: "titulo:ley AND titulo:general AND titulo:tributaria",
		},
		{nombre: "una palabra", texto: "tributaria", consulta: "titulo:tributaria"},
		{nombre: "una palabra recortada", texto: " \t tributaria\r\n", consulta: "titulo:tributaria"},
		{
			nombre:   "los separadores U+001C a U+001F son espacio en blanco de Python",
			texto:    "a\x1cb\x1dc\x1ed\x1fe",
			consulta: "titulo:a AND titulo:b AND titulo:c AND titulo:d AND titulo:e",
		},
		{
			nombre:   "el espacio en blanco de Unicode también separa",
			texto:    "ley\u00a0general\u3000tributaria\u2028estatal\u0085foral\vde\fcanarias",
			consulta: "titulo:ley AND titulo:general AND titulo:tributaria AND titulo:estatal AND titulo:foral AND titulo:de AND titulo:canarias",
		},
		{nombre: "lo que no es espacio en blanco no separa", texto: "ley\u200bgeneral", consulta: "titulo:ley\u200bgeneral"},
	}

	for _, caso := range casos {
		t.Run(caso.nombre, func(t *testing.T) {
			t.Parallel()

			consulta, err := consultaDeBusqueda(caso.texto)
			require.NoError(t, err)
			assert.Equal(t, caso.consulta, consulta)

			direccion, err := direccionDeBusqueda(caso.texto)
			require.NoError(t, err)
			compruebaDireccionDeLaFuente(t, direccion)

			if caso.direccion != "" {
				assert.Equal(t, caso.direccion, direccion)
			}

			assert.True(t, strings.HasPrefix(direccion, base+"?limit=10&query="),
				"la dirección %q no pide 10 resultados antes de la consulta", direccion)
			assert.Equal(t, caso.consulta, consultaEnviada(t, direccion))
		})
	}

	sinPalabras := []struct {
		nombre string
		texto  string
	}{
		{nombre: "vacía", texto: ""},
		{nombre: "un espacio", texto: " "},
		{nombre: "espacio en blanco ASCII", texto: " \t\n\v\f\r"},
		{nombre: "separadores U+001C a U+001F", texto: "\x1c\x1d\x1e\x1f"},
		{nombre: "espacio en blanco de Unicode", texto: "\u00a0\u0085\u2028\u3000"},
	}

	for _, caso := range sinPalabras {
		t.Run("sin ninguna palabra: "+caso.nombre, func(t *testing.T) {
			t.Parallel()

			consulta, err := consultaDeBusqueda(caso.texto)
			compruebaBusquedaSinPalabras(t, err)
			assert.Empty(t, consulta)

			direccion, err := direccionDeBusqueda(caso.texto)
			compruebaBusquedaSinPalabras(t, err)
			assert.Empty(t, direccion)
		})
	}
}

// TestCodificacionComoPython fija, byte a byte y contra salidas escritas a mano
// a partir de las reglas de CPython, cómo va la consulta en la dirección de la
// búsqueda (research.md D8): primero json.dumps con ensure_ascii, que escapa la
// comilla doble y la barra invertida, da el escape corto a \b \f \n \r \t y
// \u con cuatro dígitos hexadecimales en minúsculas al resto de controles, al
// carácter de borrado y a todo lo que no es ASCII, con un par de sustitutos
// fuera del plano básico (json/encoder.py 18-31 y 49-68); y después quote, que
// deja letras y dígitos ASCII, _ . - ~ y la barra, y codifica cada otro byte
// como %XX en mayúsculas (urllib/parse.py 790, 815, 819 y 890). Un byte que no
// forma UTF-8 válido va como el sustituto \udcXX que le da Python al leer los
// argumentos de la orden.
func TestCodificacionComoPython(t *testing.T) {
	t.Parallel()

	casos := []struct {
		nombre   string
		consulta string
		// cadena es la consulta tal como la escribe json.dumps, con sus comillas.
		cadena string
		// citada es cadena pasada por quote.
		citada string
	}{
		{
			nombre:   "ASCII sin nada que escapar",
			consulta: "titulo:ley",
			cadena:   `"titulo:ley"`,
			citada:   `%22titulo%3Aley%22`,
		},
		{
			nombre:   "ASCII que quote deja tal cual",
			consulta: "AZaz09_.-~",
			cadena:   `"AZaz09_.-~"`,
			citada:   `%22AZaz09_.-~%22`,
		},
		{
			nombre:   "ASCII que quote codifica",
			consulta: " !#$%&'()*+,:;<=>?@[]^`{|}",
			cadena:   "\" !#$%&'()*+,:;<=>?@[]^`{|}\"",
			citada:   `%22%20%21%23%24%25%26%27%28%29%2A%2B%2C%3A%3B%3C%3D%3E%3F%40%5B%5D%5E%60%7B%7C%7D%22`,
		},
		{
			nombre:   "no ASCII del plano básico",
			consulta: "común año Ñu € \ufffd \uffff",
			cadena:   `"com\u00fan a\u00f1o \u00d1u \u20ac \ufffd \uffff"`,
			citada:   `%22com%5Cu00fan%20a%5Cu00f1o%20%5Cu00d1u%20%5Cu20ac%20%5Cufffd%20%5Cuffff%22`,
		},
		{
			nombre:   "fuera del plano básico, con un par de sustitutos",
			consulta: "\U00010000 𝄞 😀 \U0010ffff",
			cadena:   `"\ud800\udc00 \ud834\udd1e \ud83d\ude00 \udbff\udfff"`,
			citada:   `%22%5Cud800%5Cudc00%20%5Cud834%5Cudd1e%20%5Cud83d%5Cude00%20%5Cudbff%5Cudfff%22`,
		},
		{
			nombre:   "comillas",
			consulta: `"ley" 'general'`,
			cadena:   `"\"ley\" 'general'"`,
			citada:   `%22%5C%22ley%5C%22%20%27general%27%22`,
		},
		{
			nombre:   "barra invertida",
			consulta: `C:\ley\u`,
			cadena:   `"C:\\ley\\u"`,
			citada:   `%22C%3A%5C%5Cley%5C%5Cu%22`,
		},
		{
			nombre:   "controles, borrado y controles de Latin-1",
			consulta: "\x00\x01\b\t\n\v\f\r\x1b\x1f\x7f\u0080\u0085",
			cadena:   `"\u0000\u0001\b\t\n\u000b\f\r\u001b\u001f\u007f\u0080\u0085"`,
			citada:   `%22%5Cu0000%5Cu0001%5Cb%5Ct%5Cn%5Cu000b%5Cf%5Cr%5Cu001b%5Cu001f%5Cu007f%5Cu0080%5Cu0085%22`,
		},
		{
			nombre:   "barra",
			consulta: "a/b /ley/",
			cadena:   `"a/b /ley/"`,
			citada:   `%22a/b%20/ley/%22`,
		},
		{
			nombre:   "UTF-8 inválido, byte a byte como sustitutos",
			consulta: "a\xffb\xe2\x82c\xed\xa0\x80",
			cadena:   `"a\udcffb\udce2\udc82c\udced\udca0\udc80"`,
			citada:   `%22a%5Cudcffb%5Cudce2%5Cudc82c%5Cudced%5Cudca0%5Cudc80%22`,
		},
	}

	for _, caso := range casos {
		t.Run(caso.nombre, func(t *testing.T) {
			t.Parallel()

			assert.Equal(t, caso.cadena, cadenaJSONComoPython(caso.consulta), "json.dumps de la consulta")
			assert.Equal(t, caso.citada, quoteComoPython(caso.cadena), "quote de la cadena")
			assert.Equal(t, antesDeLaCadena+caso.cadena+despuesDeLaCadena, cuerpoDeLaBusqueda(caso.consulta))
			assert.Equal(t, antesDeLaCadenaCitada+caso.citada+despuesDeLaCadenaCitada,
				quoteComoPython(cuerpoDeLaBusqueda(caso.consulta)))

			// Con ensure_ascii, json.dumps solo escribe ASCII imprimible; y quote no
			// pierde nada: descodificar lo citado devuelve la cadena entera.
			for indice := range len(caso.cadena) {
				assert.True(t, ' ' <= caso.cadena[indice] && caso.cadena[indice] <= '~',
					"el byte %d de %q no es ASCII imprimible", indice, caso.cadena)
			}

			descitada, err := url.PathUnescape(caso.citada)
			require.NoError(t, err)
			assert.Equal(t, caso.cadena, descitada)
		})
	}
}

// consultaEnviada devuelve la consulta que lleva la dirección de una búsqueda,
// leída con la biblioteca estándar y no con el código que la construyó: el
// parámetro query descodificado y, dentro de su JSON, query.query_string.query.
func consultaEnviada(t *testing.T, direccion string) string {
	t.Helper()

	destino, err := url.Parse(direccion)
	require.NoError(t, err)

	valores := destino.Query()
	assert.Equal(t, []string{"10"}, valores["limit"])

	var cuerpo struct {
		Query struct {
			QueryString struct {
				Query string `json:"query"`
			} `json:"query_string"`
		} `json:"query"`
	}
	require.NoError(t, json.Unmarshal([]byte(valores.Get("query")), &cuerpo))

	return cuerpo.Query.QueryString.Query
}

// compruebaBusquedaSinPalabras exige lo que el contrato errores-y-codigos fija
// para la búsqueda sin ninguna palabra (fila 4 y §2): clase «argumentos» y
// código 2, un mensaje que dice que no tiene ninguna palabra, y un *Error sin
// dirección, sin instante y sin causa, porque no se ha construido ninguna
// petición.
func compruebaBusquedaSinPalabras(t *testing.T, err error) {
	t.Helper()

	require.Error(t, err)
	assert.Equal(t, schema.ClaseArgumentos, cli.Clasificar(err))
	assert.Equal(t, 2, cli.CodigoSalida(err))
	assert.Equal(t, "la búsqueda no tiene ninguna palabra", err.Error())

	var fallo *Error
	require.ErrorAs(t, err, &fallo)
	assert.Empty(t, fallo.URL)
	assert.True(t, fallo.Instante.IsZero())
	require.NoError(t, fallo.Causa)
}
