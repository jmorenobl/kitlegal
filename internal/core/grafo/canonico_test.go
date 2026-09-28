package grafo_test

import (
	"math"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/jmorenobl/kitlegal/internal/core/grafo"
)

// TestDatosCanonicos fija la forma en la que se guardan y se comparan los datos
// identificativos de un nodo: la del esquema de canonicalización JSON de RFC
// 8785 (JCS), que es la que FR-023 nombra para desempatar dos observaciones del
// mismo instante (research.md D17). Los valores esperados son los del propio
// RFC donde los da —el ejemplo de §3.2.2.3 y el orden de claves de §3.2.3, que
// compara unidades UTF-16 y no bytes— y, en los números, los del algoritmo de
// ECMAScript que el RFC adopta: todo número es un doble de IEEE 754, sin
// exponente entre 1e-7 y 1e21, sin ceros a la derecha y con -0 escrito 0.
// Los caracteres que no son ASCII van con su escape de Go, para que se vean en
// el diff.
func TestDatosCanonicos(t *testing.T) {
	t.Parallel()

	casos := []struct {
		nombre    string
		datos     map[string]any
		canonicos string
	}{
		{nombre: "sin datos", datos: nil, canonicos: `{}`},
		{nombre: "datos vacios", datos: map[string]any{}, canonicos: `{}`},
		{
			nombre: "los datos de una version de bloque",
			datos: map[string]any{
				grafo.DatoHashTexto:         "sha256:ab",
				grafo.DatoFechaVigencia:     "20161002",
				grafo.DatoNormaModificadora: "",
				grafo.DatoFechaVersion:      "20161002",
			},
			canonicos: `{"fecha_version":"20161002","fecha_vigencia":"20161002","hash_texto":"sha256:ab",` +
				`"norma_modificadora":""}`,
		},
		{
			nombre: "las claves de todos los tipos de nodo, ordenadas",
			datos: map[string]any{
				grafo.DatoIdentificador:     "BOE-A-2015-10565",
				grafo.DatoBloque:            "a21",
				grafo.DatoFechaVigencia:     "20161002",
				grafo.DatoFechaVersion:      "20161002",
				grafo.DatoNormaModificadora: "BOE-A-2016-0001",
				grafo.DatoHashTexto:         "sha256:ab",
				grafo.DatoCodigoINE:         "28074",
				grafo.DatoNombre:            "Leganés",
				grafo.DatoDIR3:              "L01280745",
			},
			canonicos: `{"bloque":"a21","codigo_ine":"28074","dir3":"L01280745","fecha_version":"20161002",` +
				`"fecha_vigencia":"20161002","hash_texto":"sha256:ab","identificador":"BOE-A-2015-10565",` +
				`"nombre":"Leganés","norma_modificadora":"BOE-A-2016-0001"}`,
		},
		{
			nombre: "orden de las claves por unidades UTF-16 (RFC 8785, 3.2.3)",
			datos: map[string]any{
				"\u20ac":     "Euro Sign",
				"\r":         "Carriage Return",
				"\ufb33":     "Hebrew Letter Dalet With Dagesh",
				"1":          "One",
				"\U0001f600": "Emoji: Grinning Face",
				"\u0080":     "Control",
				"\u00f6":     "Latin Small Letter O With Diaeresis",
			},
			canonicos: `{"\r":"Carriage Return","1":"One","` + "\u0080" + `":"Control","` + "\u00f6" +
				`":"Latin Small Letter O With Diaeresis","` + "\u20ac" + `":"Euro Sign","` + "\U0001f600" +
				`":"Emoji: Grinning Face","` + "\ufb33" + `":"Hebrew Letter Dalet With Dagesh"}`,
		},
		{
			nombre: "objetos anidados, ordenados, y listas, que conservan su orden",
			datos: map[string]any{
				"b": []any{"z", "a"},
				"a": map[string]any{"d": false, "c": nil, "e": map[string]any{"g": true, "f": 1}},
			},
			canonicos: `{"a":{"c":null,"d":false,"e":{"f":1,"g":true}},"b":["z","a"]}`,
		},
		{
			nombre: "numeros y objetos de otros tipos de Go",
			datos: map[string]any{
				"m": map[string]string{"b": "2", "a": "1"},
				"n": uint16(1),
				"r": 1.0,
				"e": int64(1),
			},
			canonicos: `{"e":1,"m":{"a":"1","b":"2"},"n":1,"r":1}`,
		},
		{
			nombre: "el ejemplo de RFC 8785, 3.2.2.3",
			datos: map[string]any{
				"numbers":  []any{333333333.33333329, 1e30, 4.50, 2e-3, 0.000000000000000000000000001},
				"string":   "\u20ac$\u000F\u000aA'\u0042\u0022\u005c\\\"/",
				"literals": []any{nil, true, false},
			},
			canonicos: `{"literals":[null,true,false],"numbers":[333333333.3333333,1e+30,4.5,0.002,1e-27],` +
				`"string":"` + "\u20ac" + `$\u000f\nA'B\"\\\\\"/"}`,
		},
		{
			// 2^53 + 1 no es un doble: queda en el par más cercano, 2^53. Y 2^68
			// necesita diecisiete cifras para volver al mismo doble.
			nombre: "numeros como dobles de IEEE 754, con el formato de ECMAScript",
			datos: map[string]any{
				"a": 0,
				"b": math.Copysign(0, -1),
				"c": 1e20,
				"d": 1e21,
				"e": 1e-6,
				"f": 1e-7,
				"g": math.SmallestNonzeroFloat64,
				"h": math.MaxFloat64,
				"i": int64(1)<<53 + 1,
				"j": math.Ldexp(1, 68),
				"k": -1.5,
				"l": uint8(7),
			},
			canonicos: `{"a":0,"b":0,"c":100000000000000000000,"d":1e+21,"e":0.000001,"f":1e-7,"g":5e-324,` +
				`"h":1.7976931348623157e+308,"i":9007199254740992,"j":295147905179352830000,"k":-1.5,"l":7}`,
		},
		{
			// Solo se escapan la comilla, la barra invertida y los controles por
			// debajo de U+0020, con la forma corta si la tienen y si no con cuatro
			// cifras en minúscula; ni <, >, &, U+2028, U+2029 ni U+007F.
			nombre: "escapes minimos",
			datos: map[string]any{
				"s": "<>&\u2028\u2029\u007f\u001f\b\f\n\r\t\"\\/\u00f1",
			},
			canonicos: `{"s":"<>&` + "\u2028\u2029\u007f" + `\u001f\b\f\n\r\t\"\\/` + "\u00f1" + `"}`,
		},
	}

	for _, caso := range casos {
		t.Run(caso.nombre, func(t *testing.T) {
			t.Parallel()

			canonicos, err := grafo.DatosCanonicos(caso.datos)
			require.NoError(t, err)
			assert.Equal(t, caso.canonicos, canonicos)
		})
	}
}

// TestDatosCanonicosSinOrden fija que la forma canónica no depende de cómo se
// construyen los datos: ni del orden en el que Go visita las claves de un
// mapa, que cambia de una llamada a otra, ni del tipo de Go de un número o de
// un objeto. Compara resultados entre sí, byte a byte: dos datos iguales que
// dieran formas distintas romperían el desempate de FR-023. Qué forma es cada
// una lo fija TestDatosCanonicos.
func TestDatosCanonicosSinOrden(t *testing.T) {
	t.Parallel()

	t.Run("el mismo resultado en cada llamada", func(t *testing.T) {
		t.Parallel()

		datos := map[string]any{
			grafo.DatoIdentificador: "BOE-A-2015-10565",
			grafo.DatoBloque:        "a21",
			grafo.DatoFechaVigencia: "20161002",
			grafo.DatoFechaVersion:  "20161002",
			grafo.DatoHashTexto:     "sha256:ab",
			grafo.DatoCodigoINE:     "28074",
			grafo.DatoNombre:        "Leganés",
			grafo.DatoDIR3:          "L01280745",
			"anidado":               map[string]any{"z": 1, "y": 2, "x": 3, "w": 4, "v": 5, "u": 6},
		}

		primera, err := grafo.DatosCanonicos(datos)
		require.NoError(t, err)

		for range 50 {
			canonicos, err := grafo.DatosCanonicos(datos)
			require.NoError(t, err)
			require.Equal(t, primera, canonicos)
		}
	})

	t.Run("el mismo resultado con otros tipos de Go", func(t *testing.T) {
		t.Parallel()

		primera, err := grafo.DatosCanonicos(map[string]any{"n": 1, "m": map[string]any{"b": "2", "a": "1"}})
		require.NoError(t, err)

		otras := []map[string]any{
			{"m": map[string]string{"a": "1", "b": "2"}, "n": 1.0},
			{"n": int64(1), "m": map[string]any{"a": "1", "b": "2"}},
			{"m": map[string]string{"b": "2", "a": "1"}, "n": uint16(1)},
		}

		for _, datos := range otras {
			canonicos, err := grafo.DatosCanonicos(datos)
			require.NoError(t, err)
			assert.Equal(t, primera, canonicos)
		}
	})
}

// TestDatosCanonicosImposibles fija que unos datos que no tienen forma JSON
// —un número que no es finito, una cadena o una clave que no es UTF-8, un
// valor que no es de JSON— no se canonicalizan a medias ni se corrigen: dan un
// error que dice qué forma no pudieron tomar. RFC 8785 no admite números que
// no sean finitos, y un byte que no es UTF-8 sustituido por U+FFFD cambiaría
// los datos del nodo.
func TestDatosCanonicosImposibles(t *testing.T) {
	t.Parallel()

	casos := []struct {
		nombre string
		datos  map[string]any
	}{
		{"NaN", map[string]any{"n": math.NaN()}},
		{"infinito", map[string]any{"n": math.Inf(1)}},
		{"menos infinito", map[string]any{"n": map[string]any{"m": math.Inf(-1)}}},
		{"cadena que no es UTF-8", map[string]any{"s": "a\xffb"}},
		{"clave que no es UTF-8", map[string]any{"a\xffb": "s"}},
		{"valor que no es de JSON", map[string]any{"f": func() {}}},
	}

	for _, caso := range casos {
		t.Run(caso.nombre, func(t *testing.T) {
			t.Parallel()

			canonicos, err := grafo.DatosCanonicos(caso.datos)
			require.ErrorContains(t, err, "los datos no tienen forma JSON canónica (RFC 8785)")
			assert.Empty(t, canonicos)
		})
	}
}
