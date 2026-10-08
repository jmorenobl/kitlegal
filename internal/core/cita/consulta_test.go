package cita_test

import (
	"reflect"
	"strconv"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/jmorenobl/kitlegal/internal/core/cita"
)

// preparada es la consulta de una referencia bien dada.
func preparada(t *testing.T, dados argumentos) cita.Consulta {
	t.Helper()

	referencia, hay, err := dados.referencia()
	require.NoError(t, err)
	require.True(t, hay, "el caso no da ninguna referencia")

	return cita.Preparar(referencia)
}

// TestPreparar fija lo que cita preparar da de una referencia y de un texto
// de búsqueda, con las claves de contracts/applet-cita.md §3 en su orden: la
// dirección del buscador y las casillas de la forma dada, con la fecha escrita
// como la escribe quien rellena la casilla; el equivalente solo en la pareja
// de FR-012; la cobertura solo con el ECLI del Tribunal Constitucional, sin
// dirección ni casillas; y, con un texto, la dirección de la búsqueda ya hecha,
// codificada como un segmento de ruta. Lo que no aplica no está, y casillas es
// siempre una lista (H23, FR-010 a FR-015; research.md D4).
func TestPreparar(t *testing.T) {
	t.Parallel()

	t.Run("las salidas del spec", func(t *testing.T) {
		t.Parallel()

		casos := []struct {
			nombre      string
			dados       argumentos
			serializada string
		}{
			{
				"un ECLI",
				argumentos{ecli: dado(ecliConocido)},
				`{"referencia":{"forma":"ecli","valor":"ECLI:ES:TS:2023:3144"},` +
					`"equivalente":{"forma":"roj","valor":"STS 3144/2023"},` +
					`"direccion":"https://www.poderjudicial.es/search/indexAN.jsp",` +
					`"casillas":[{"nombre":"ECLI","valor":"ECLI:ES:TS:2023:3144"}]}`,
			},
			{
				"un ROJ",
				argumentos{roj: dado(rojConocido)},
				`{"referencia":{"forma":"roj","valor":"STS 3144/2023"},` +
					`"equivalente":{"forma":"ecli","valor":"ECLI:ES:TS:2023:3144"},` +
					`"direccion":"https://www.poderjudicial.es/search/indexAN.jsp",` +
					`"casillas":[{"nombre":"Nº ROJ","valor":"STS 3144/2023"}]}`,
			},
			{
				"un número con su fecha",
				argumentos{resolucion: dado(resolucionConocida), fecha: dado(fechaConocida)},
				`{"referencia":{"forma":"resolucion","valor":"1088/2023","fecha":"2023-07-04"},` +
					`"direccion":"https://www.poderjudicial.es/search/indexAN.jsp",` +
					`"casillas":[{"nombre":"Nº Resolución","valor":"1088/2023"},` +
					`{"nombre":"Fecha resolución","campo":"Desde","valor":"04/07/2023"},` +
					`{"nombre":"Fecha resolución","campo":"Hasta","valor":"04/07/2023"}]}`,
			},
		}

		for _, caso := range casos {
			t.Run(caso.nombre, func(t *testing.T) {
				t.Parallel()

				consulta := preparada(t, caso.dados)
				compruebaSerializacion(t, caso.serializada, consulta)
				compruebaEtiquetas(t, consulta)
			})
		}
	})

	t.Run("sin equivalente y sin cobertura", func(t *testing.T) {
		t.Parallel()

		casos := []struct {
			nombre      string
			dados       argumentos
			serializada string
		}{
			{
				"un ECLI de otro órgano, con puntos y letras en el número",
				argumentos{ecli: dado("ECLI:ES:AN:2019:1.2.A")},
				`{"referencia":{"forma":"ecli","valor":"ECLI:ES:AN:2019:1.2.A"},` +
					`"direccion":"https://www.poderjudicial.es/search/indexAN.jsp",` +
					`"casillas":[{"nombre":"ECLI","valor":"ECLI:ES:AN:2019:1.2.A"}]}`,
			},
			{
				"un ECLI del Tribunal Supremo cuyo número termina en letra",
				argumentos{ecli: dado("ECLI:ES:TS:2023:3144A")},
				`{"referencia":{"forma":"ecli","valor":"ECLI:ES:TS:2023:3144A"},` +
					`"direccion":"https://www.poderjudicial.es/search/indexAN.jsp",` +
					`"casillas":[{"nombre":"ECLI","valor":"ECLI:ES:TS:2023:3144A"}]}`,
			},
			{
				"un ECLI cuyo órgano empieza como el del Tribunal Constitucional",
				argumentos{ecli: dado("ECLI:ES:TCT:1985:1")},
				`{"referencia":{"forma":"ecli","valor":"ECLI:ES:TCT:1985:1"},` +
					`"direccion":"https://www.poderjudicial.es/search/indexAN.jsp",` +
					`"casillas":[{"nombre":"ECLI","valor":"ECLI:ES:TCT:1985:1"}]}`,
			},
			{
				"un ROJ de otras siglas",
				argumentos{roj: dado("SAP M 1234/2020")},
				`{"referencia":{"forma":"roj","valor":"SAP M 1234/2020"},` +
					`"direccion":"https://www.poderjudicial.es/search/indexAN.jsp",` +
					`"casillas":[{"nombre":"Nº ROJ","valor":"SAP M 1234/2020"}]}`,
			},
			{
				"el ROJ de un auto",
				argumentos{roj: dado("ATS 3144/2023")},
				`{"referencia":{"forma":"roj","valor":"ATS 3144/2023"},` +
					`"direccion":"https://www.poderjudicial.es/search/indexAN.jsp",` +
					`"casillas":[{"nombre":"Nº ROJ","valor":"ATS 3144/2023"}]}`,
			},
			{
				"un ROJ con las siglas del Tribunal Constitucional, que se prepara como cualquier otro",
				argumentos{roj: dado("STC 79/2024")},
				`{"referencia":{"forma":"roj","valor":"STC 79/2024"},` +
					`"direccion":"https://www.poderjudicial.es/search/indexAN.jsp",` +
					`"casillas":[{"nombre":"Nº ROJ","valor":"STC 79/2024"}]}`,
			},
		}

		for _, caso := range casos {
			t.Run(caso.nombre, func(t *testing.T) {
				t.Parallel()

				consulta := preparada(t, caso.dados)
				compruebaSerializacion(t, caso.serializada, consulta)
				compruebaEtiquetas(t, consulta)
			})
		}
	})

	t.Run("cobertura solo con el ECLI de órgano TC", func(t *testing.T) {
		t.Parallel()

		consulta := preparada(t, argumentos{ecli: dado("ECLI:ES:TC:2024:79")})
		compruebaSerializacion(t,
			`{"referencia":{"forma":"ecli","valor":"ECLI:ES:TC:2024:79"},`+
				`"cobertura":{"cendoj":"no-cubierto",`+
				`"motivo":"Las resoluciones del Tribunal Constitucional no están en el CENDOJ: su buscador no las tiene."},`+
				`"casillas":[]}`,
			consulta)
		compruebaEtiquetas(t, consulta)
	})

	t.Run("un texto de búsqueda", func(t *testing.T) {
		t.Parallel()
		compruebaPrepararTexto(t)
	})

	t.Run("un texto vacío o solo de blancos", func(t *testing.T) {
		t.Parallel()
		compruebaTextoRechazado(t)
	})

	t.Run("lo que las etiquetas enumeran", func(t *testing.T) {
		t.Parallel()

		assert.Equal(t, []string{"ecli", "roj"}, enumeradoDe(t, reflect.TypeFor[cita.Equivalente](), "Forma"))
		assert.Equal(t, []string{"no-cubierto"}, enumeradoDe(t, reflect.TypeFor[cita.Cobertura](), "CENDOJ"))
	})
}

// compruebaPrepararTexto fija la dirección de la búsqueda por texto de FR-013:
// el texto tal como se dio, en UTF-8, con cada octeto que no es una letra
// ASCII, una cifra, «-», «.», «_» o «~» escrito como % y sus dos cifras
// hexadecimales en mayúsculas, entre el prefijo y el sufijo del buscador; sin
// casillas, sin referencia y sin nada añadido al texto.
func compruebaPrepararTexto(t *testing.T) {
	t.Helper()

	const (
		prefijo = "https://www.poderjudicial.es/search/sentencias/"
		sufijo  = "/1/AN"
	)

	consulta, err := cita.PrepararTexto("cláusula suelo")
	require.NoError(t, err)
	compruebaSerializacion(t,
		`{"texto":"cláusula suelo",`+
			`"direccion":"https://www.poderjudicial.es/search/sentencias/cl%C3%A1usula%20suelo/1/AN",`+
			`"casillas":[]}`,
		consulta)
	compruebaEtiquetas(t, consulta)

	casos := []struct{ nombre, texto, codificado string }{
		{"lo que no se codifica", "AZaz09-._~", "AZaz09-._~"},
		{"un espacio", "a b", "a%20b"},
		{"los blancos de los extremos, que no se recortan", " desahucio ", "%20desahucio%20"},
		{"lo reservado de una dirección", "a/b?c#d&e=f+g", "a%2Fb%3Fc%23d%26e%3Df%2Bg"},
		{"el propio signo de porcentaje", "100%", "100%25"},
		{"comillas y dos puntos", `"IVA": tipo`, "%22IVA%22%3A%20tipo"},
		{"una eñe, en sus dos octetos", "daño", "da%C3%B1o"},
		{"el hexadecimal, en mayúsculas", "\n\x7f", "%0A%7F"},
		{"un octeto que no es UTF-8", "a\xffb", "a%FFb"},
	}

	for _, caso := range casos {
		t.Run(caso.nombre, func(t *testing.T) {
			t.Parallel()

			consulta, err := cita.PrepararTexto(caso.texto)
			require.NoError(t, err)
			assert.Equal(t, caso.texto, consulta.Texto, "el texto no va tal como se dio")
			assert.Equal(t, prefijo+caso.codificado+sufijo, consulta.Direccion)
			assert.Empty(t, consulta.Casillas, "la búsqueda ya está hecha: no hay casillas que rellenar")
			assert.NotNil(t, consulta.Casillas, "casillas es siempre una lista")
			assert.Nil(t, consulta.Referencia)
			assert.Nil(t, consulta.Equivalente)
			assert.Nil(t, consulta.Cobertura)
		})
	}
}

// compruebaTextoRechazado fija que un texto vacío o solo de blancos es un
// error de argumentos que nombra el texto entre comillas, con el valor cero
// al lado (FR-015).
func compruebaTextoRechazado(t *testing.T) {
	t.Helper()

	// espacioDuro es U+00A0, que también es un blanco.
	const espacioDuro = "\xc2\xa0"

	casos := []struct{ nombre, texto string }{
		{"vacío", ""},
		{"un espacio", " "},
		{"tabulador y saltos de línea", "\t\r\n"},
		{"un espacio duro", espacioDuro},
	}

	for _, caso := range casos {
		t.Run(caso.nombre, func(t *testing.T) {
			t.Parallel()

			consulta, err := cita.PrepararTexto(caso.texto)
			compruebaErrorDeArgumentos(t, err, "el texto de la búsqueda "+strconv.Quote(caso.texto))
			assert.Equal(t, cita.Consulta{}, consulta, "un texto rechazado no puede acompañarse de una consulta")
		})
	}
}
