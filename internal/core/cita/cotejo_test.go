package cita_test

import (
	"reflect"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/jmorenobl/kitlegal/internal/core/cita"
)

// cotejado es el cotejo de la ficha de un texto con la referencia de esos
// argumentos, o sin ninguna si no dan ninguna.
func cotejado(t *testing.T, texto string, dados argumentos) cita.Cotejo {
	t.Helper()

	referencia, hay, err := dados.referencia()
	require.NoError(t, err)

	if !hay {
		return cita.Cotejar(leida(t, texto), nil)
	}

	return cita.Cotejar(leida(t, texto), &referencia)
}

// TestCotejar fija lo que cita cotejar da de la ficha de un documento, con
// las claves de contracts/applet-cita.md §4 en su orden: los ocho datos y la
// correspondencia de su ROJ con su ECLI, de tres valores; con una referencia,
// qué se pidió y si el documento es el pedido; y, cuando no lo es, un hallazgo
// documento-distinto, uno solo, con cada dato que difiere, el cruce cuando el
// número pedido es el otro número del documento y una explicación en español.
// Sin referencia no dice nada de lo pedido ni lleva hallazgo (H23, FR-022 a
// FR-025; los siete casos de FR-081).
func TestCotejar(t *testing.T) {
	t.Parallel()

	t.Run("los siete casos del fragmento", func(t *testing.T) {
		t.Parallel()

		for _, caso := range cotejosDelFragmento() {
			t.Run(caso.nombre, func(t *testing.T) {
				t.Parallel()

				cotejo := cotejado(t, fragmento(t), caso.dados)
				compruebaSerializacion(t, caso.serializado, cotejo)
				compruebaEtiquetas(t, cotejo)
			})
		}
	})

	t.Run("lo que difiere y su cruce", func(t *testing.T) {
		t.Parallel()
		compruebaDiferencias(t)
	})

	t.Run("la correspondencia del ROJ con el ECLI", func(t *testing.T) {
		t.Parallel()
		compruebaCorrespondencias(t)
	})

	t.Run("la referencia pedida es una copia", func(t *testing.T) {
		t.Parallel()

		referencia, hay, err := argumentos{roj: dado(rojConocido)}.referencia()
		require.NoError(t, err)
		require.True(t, hay)

		cotejo := cita.Cotejar(leida(t, fragmento(t)), &referencia)
		require.NotNil(t, cotejo.Pedida)
		assert.NotSame(t, &referencia, cotejo.Pedida, "el cotejo guarda la referencia de quien lo pide, y no una copia")
		assert.Equal(t, referencia, *cotejo.Pedida)
	})

	t.Run("lo que las etiquetas enumeran", func(t *testing.T) {
		t.Parallel()

		assert.Equal(t, []string{"se-corresponden", "no-se-corresponden", "no-se-deduce"},
			enumeradoDe(t, reflect.TypeFor[cita.Cotejo](), "Correspondencia"))
		assert.Equal(t, []string{"documento-distinto"}, enumeradoDe(t, reflect.TypeFor[cita.Hallazgo](), "Clase"))
		assert.Equal(t, []string{"numero-de-resolucion", "numero-del-roj"},
			enumeradoDe(t, reflect.TypeFor[cita.Hallazgo](), "Cruce"))
		assert.Equal(t, []string{"ecli", "roj", "resolucion", "fecha"},
			enumeradoDe(t, reflect.TypeFor[cita.Diferencia](), "Dato"))
	})
}

// cotejoDelFragmento es un cotejo del fragmento de la evidencia con lo que
// tiene que dar, byte a byte.
type cotejoDelFragmento struct {
	nombre      string
	dados       argumentos
	serializado string
}

// cotejosDelFragmento son los siete casos de FR-081: sin referencia; pedido
// por su ECLI, por su ROJ y por su número con su fecha, que es el pedido; y
// pedido como ROJ STS 1088/2023, con otra fecha y por el número 3144/2023, que
// no lo es, con el cruce en el primero y en el último.
func cotejosDelFragmento() []cotejoDelFragmento {
	const (
		delDocumento = `{"ficha":` + fichaSerializada + `,"correspondencia":"se-corresponden",`
		hallazgo     = `"es_la_pedida":false,"hallazgos":[{"clase":"documento-distinto","difiere":`
	)

	return []cotejoDelFragmento{
		{"sin referencia", argumentos{}, delDocumento + `"hallazgos":[]}`},
		{
			"pedido por su ECLI",
			argumentos{ecli: dado(ecliConocido)},
			delDocumento + `"pedida":{"forma":"ecli","valor":"ECLI:ES:TS:2023:3144"},"es_la_pedida":true,"hallazgos":[]}`,
		},
		{
			"pedido por su ROJ",
			argumentos{roj: dado(rojConocido)},
			delDocumento + `"pedida":{"forma":"roj","valor":"STS 3144/2023"},"es_la_pedida":true,"hallazgos":[]}`,
		},
		{
			"pedido por su número con su fecha",
			argumentos{resolucion: dado(resolucionConocida), fecha: dado(fechaConocida)},
			delDocumento + `"pedida":{"forma":"resolucion","valor":"1088/2023","fecha":"2023-07-04"},` +
				`"es_la_pedida":true,"hallazgos":[]}`,
		},
		{
			"pedido como ROJ STS 1088/2023, que es su número de resolución",
			argumentos{roj: dado("STS 1088/2023")},
			delDocumento + `"pedida":{"forma":"roj","valor":"STS 1088/2023"},` + hallazgo +
				`[{"dato":"roj","pedido":"STS 1088/2023","documento":"STS 3144/2023"}],` +
				`"cruce":"numero-de-resolucion",` +
				`"explicacion":"El documento no es el pedido: se pidió el ROJ STS 1088/2023 y el del documento es ` +
				`STS 3144/2023. 1088/2023 es el número de resolución del documento, no su ROJ."}]}`,
		},
		{
			"pedido con otra fecha",
			argumentos{resolucion: dado(resolucionConocida), fecha: dado("2023-01-01")},
			delDocumento + `"pedida":{"forma":"resolucion","valor":"1088/2023","fecha":"2023-01-01"},` + hallazgo +
				`[{"dato":"fecha","pedido":"2023-01-01","documento":"2023-07-04"}],` +
				`"explicacion":"El documento no es el pedido: se pidió la fecha 2023-01-01 y la del documento es ` +
				`2023-07-04."}]}`,
		},
		{
			"pedido por el número 3144/2023, que es el de su ROJ",
			argumentos{resolucion: dado("3144/2023"), fecha: dado(fechaConocida)},
			delDocumento + `"pedida":{"forma":"resolucion","valor":"3144/2023","fecha":"2023-07-04"},` + hallazgo +
				`[{"dato":"resolucion","pedido":"3144/2023","documento":"1088/2023"}],` +
				`"cruce":"numero-del-roj",` +
				`"explicacion":"El documento no es el pedido: se pidió el número de resolución 3144/2023 y el del ` +
				`documento es 1088/2023. 3144/2023 es el número del ROJ del documento (STS 3144/2023), no su número ` +
				`de resolución."}]}`,
		},
	}
}

// compruebaDiferencias fija qué lleva el hallazgo: una diferencia por dato
// distinto —el ECLI, el ROJ o, pedido por número y fecha, el número, la fecha
// o los dos, en ese orden—, y el cruce solo cuando el número difiere y es el
// otro número del documento. Ni el órgano ni las siglas intervienen.
func compruebaDiferencias(t *testing.T) {
	t.Helper()

	// conElNumeroDelROJ es una ficha cuyo número de resolución coincide con el
	// número de su ROJ: pedirla por él no es cruzar nada.
	conElNumeroDelROJ := conLaLinea(t, fichaDelFragmento(t), "Nº de Resolución", "Nº de Resolución: 3144/2023")

	casos := []struct {
		nombre  string
		texto   string
		dados   argumentos
		difiere []cita.Diferencia
		cruce   string
	}{
		{
			"otro ECLI",
			fragmento(t),
			argumentos{ecli: dado("ECLI:ES:TS:2023:1088")},
			[]cita.Diferencia{{Dato: "ecli", Pedido: "ECLI:ES:TS:2023:1088", Documento: "ECLI:ES:TS:2023:3144"}},
			"",
		},
		{
			"otro ROJ, sin cruce",
			fragmento(t),
			argumentos{roj: dado("STS 999/2023")},
			[]cita.Diferencia{{Dato: "roj", Pedido: "STS 999/2023", Documento: "STS 3144/2023"}},
			"",
		},
		{
			"un ROJ de otras siglas con el número de resolución",
			fragmento(t),
			argumentos{roj: dado("SAP M 1088/2023")},
			[]cita.Diferencia{{Dato: "roj", Pedido: "SAP M 1088/2023", Documento: "STS 3144/2023"}},
			"numero-de-resolucion",
		},
		{
			"otro número, sin cruce",
			fragmento(t),
			argumentos{resolucion: dado("999/2023"), fecha: dado(fechaConocida)},
			[]cita.Diferencia{{Dato: "resolucion", Pedido: "999/2023", Documento: "1088/2023"}},
			"",
		},
		{
			"el número con ceros por delante, que no se normaliza",
			fragmento(t),
			argumentos{resolucion: dado("01088/2023"), fecha: dado(fechaConocida)},
			[]cita.Diferencia{{Dato: "resolucion", Pedido: "01088/2023", Documento: "1088/2023"}},
			"",
		},
		{
			"otro número y otra fecha",
			fragmento(t),
			argumentos{resolucion: dado("999/2023"), fecha: dado("2023-01-01")},
			[]cita.Diferencia{
				{Dato: "resolucion", Pedido: "999/2023", Documento: "1088/2023"},
				{Dato: "fecha", Pedido: "2023-01-01", Documento: "2023-07-04"},
			},
			"",
		},
		{
			"el número del ROJ y otra fecha",
			fragmento(t),
			argumentos{resolucion: dado("3144/2023"), fecha: dado("2023-01-01")},
			[]cita.Diferencia{
				{Dato: "resolucion", Pedido: "3144/2023", Documento: "1088/2023"},
				{Dato: "fecha", Pedido: "2023-01-01", Documento: "2023-07-04"},
			},
			"numero-del-roj",
		},
		{
			"su número, que es también el de su ROJ, con otra fecha",
			conElNumeroDelROJ,
			argumentos{resolucion: dado("3144/2023"), fecha: dado("2023-01-01")},
			[]cita.Diferencia{{Dato: "fecha", Pedido: "2023-01-01", Documento: "2023-07-04"}},
			"",
		},
	}

	for _, caso := range casos {
		t.Run(caso.nombre, func(t *testing.T) {
			t.Parallel()

			cotejo := cotejado(t, caso.texto, caso.dados)
			require.NotNil(t, cotejo.EsLaPedida)
			assert.False(t, *cotejo.EsLaPedida)
			require.Len(t, cotejo.Hallazgos, 1, "el hallazgo es uno por invocación")

			hallazgo := cotejo.Hallazgos[0]
			assert.Equal(t, "documento-distinto", hallazgo.Clase)
			assert.Equal(t, caso.difiere, hallazgo.Difiere)
			assert.Equal(t, caso.cruce, hallazgo.Cruce)
			assert.Contains(t, hallazgo.Explicacion, "El documento no es el pedido")

			for _, diferencia := range caso.difiere {
				assert.Contains(t, hallazgo.Explicacion, diferencia.Pedido, "la explicación no nombra lo pedido")
				assert.Contains(t, hallazgo.Explicacion, diferencia.Documento, "la explicación no nombra lo del documento")
			}

			compruebaEtiquetas(t, cotejo)
		})
	}

	t.Run("su ROJ, cuyo número es también el de resolución, es el pedido", func(t *testing.T) {
		t.Parallel()

		cotejo := cotejado(t, conElNumeroDelROJ, argumentos{roj: dado(rojConocido)})
		require.NotNil(t, cotejo.EsLaPedida)
		assert.True(t, *cotejo.EsLaPedida)
		assert.Empty(t, cotejo.Hallazgos)
	})
}

// compruebaCorrespondencias fija la correspondencia de FR-023, de tres
// valores: la regla alcanza a la ficha solo si su ROJ y su ECLI son los dos de
// la pareja de FR-012 —siglas STS y órgano TS con el número de cifras—, y
// entonces se corresponden si el número y el año coinciden carácter a
// carácter. En cualquier otro caso no se deduce. Es un dato, no un hallazgo.
func compruebaCorrespondencias(t *testing.T) {
	t.Helper()

	casos := []struct{ nombre, linea, correspondencia string }{
		{"la ficha del fragmento", "Roj: STS 3144/2023 - ECLI:ES:TS:2023:3144", "se-corresponden"},
		{"otro número", "Roj: STS 3144/2023 - ECLI:ES:TS:2023:9999", "no-se-corresponden"},
		{"otro año", "Roj: STS 3144/2023 - ECLI:ES:TS:2022:3144", "no-se-corresponden"},
		{"el mismo número con un cero por delante", "Roj: STS 03144/2023 - ECLI:ES:TS:2023:3144", "no-se-corresponden"},
		{"una ficha de otro órgano", lineaDeOtroOrgano, "no-se-deduce"},
		{"un auto", "Roj: ATS 3144/2023 - ECLI:ES:TS:2023:3144A", "no-se-deduce"},
		{"el ECLI fuera de la pareja", "Roj: STS 3144/2023 - ECLI:ES:AN:2023:3144", "no-se-deduce"},
		{"el ROJ fuera de la pareja", "Roj: SAN 3144/2023 - ECLI:ES:TS:2023:3144", "no-se-deduce"},
	}

	for _, caso := range casos {
		t.Run(caso.nombre, func(t *testing.T) {
			t.Parallel()

			cotejo := cotejado(t, conLaLinea(t, fichaDelFragmento(t), "Roj", caso.linea), argumentos{})
			assert.Equal(t, caso.correspondencia, cotejo.Correspondencia)
			assert.Nil(t, cotejo.Pedida, "sin referencia no hay nada pedido")
			assert.Nil(t, cotejo.EsLaPedida, "sin referencia no se dice si es el pedido")
			assert.Empty(t, cotejo.Hallazgos, "la correspondencia es un dato, no un hallazgo")
			assert.NotNil(t, cotejo.Hallazgos, "hallazgos es siempre una lista")
			compruebaEtiquetas(t, cotejo)
		})
	}
}
