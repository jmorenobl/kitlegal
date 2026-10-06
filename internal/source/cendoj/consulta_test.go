package cendoj

import (
	"net/url"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/jmorenobl/kitlegal/internal/core"
)

// TestConsultaResolver fija que la consulta del verbo implementa core.Consulta,
// por valor y por puntero, que declara el verbo resolver lleve o no una
// referencia, y que su vigencia son 30 días (contrato cita-resolver §5;
// data-model §2).
func TestConsultaResolver(t *testing.T) {
	t.Parallel()

	referencia, err := NuevaReferencia(ecliConocido, "", "", "")
	require.NoError(t, err)

	var (
		vacia    core.Consulta = ConsultaResolver{}
		conDatos core.Consulta = &ConsultaResolver{Referencia: referencia}
	)

	assert.Equal(t, "resolver", vacia.Verbo())
	assert.Equal(t, "resolver", conDatos.Verbo())
	assert.Equal(t, 2_592_000*time.Second, vigenciaDeResolver, "30 días")
}

// TestCamposDeLaConsulta fija el cuerpo del envío de cada forma de referencia,
// codificado como lo codifica internal/httpx (contrato
// fuente-cendoj-y-grabacion §2; research D9; FR-006, FR-020): los cinco campos
// fijos y el campo de la referencia —ECLI; ROJ, sin fechas también cuando lleva
// fecha; o NUMERORESOLUCION con la fecha como principio y como fin del
// intervalo, escrita dd/mm/aaaa—, y ninguno más. El del ECLI es el del ejemplo
// del contrato httpx-formulario §5, y los tres tamaños, los de research M2.
func TestCamposDeLaConsulta(t *testing.T) {
	t.Parallel()

	const fijos = "action=query&databasematch=AN&recordsPerPage=10&sort=IN_FECHARESOLUCION%3Adecreasing&start=1"

	casos := []struct {
		nombre     string
		ecli       string
		roj        string
		resolucion string
		fecha      string
		cuerpo     string
		bytes      int
	}{
		{
			nombre: "por ECLI",
			ecli:   ecliConocido,
			cuerpo: "ECLI=ECLI%3AES%3ATS%3A2023%3A3144&" + fijos,
			bytes:  126,
		},
		{
			nombre: "por ROJ",
			roj:    rojConocido,
			cuerpo: "ROJ=STS+3144%2F2023&" + fijos,
			bytes:  112,
		},
		{
			nombre: "por ROJ con fecha, sin fechas en la consulta",
			roj:    rojConocido,
			fecha:  fechaConocida,
			cuerpo: "ROJ=STS+3144%2F2023&" + fijos,
			bytes:  112,
		},
		{
			nombre:     "por número de resolución con su fecha",
			resolucion: resolucionConocida,
			fecha:      fechaConocida,
			cuerpo: "FECHARESOLUCIONDESDE=04%2F07%2F2023&FECHARESOLUCIONHASTA=04%2F07%2F2023&" +
				"NUMERORESOLUCION=1088%2F2023&" + fijos,
			bytes: 193,
		},
	}

	for _, caso := range casos {
		t.Run(caso.nombre, func(t *testing.T) {
			t.Parallel()

			referencia, err := NuevaReferencia(caso.ecli, caso.roj, caso.resolucion, caso.fecha)
			require.NoError(t, err)

			cuerpo := cuerpoCodificado(camposDe(referencia))

			assert.Equal(t, caso.cuerpo, cuerpo)
			assert.Len(t, cuerpo, caso.bytes)
		})
	}

	t.Run("el ejemplo del contrato del formulario", func(t *testing.T) {
		t.Parallel()

		referencia, err := NuevaReferencia(ecliConocido, "", "", "")
		require.NoError(t, err)

		assert.Equal(t,
			"ECLI=ECLI%3AES%3ATS%3A2023%3A3144&action=query&databasematch=AN&recordsPerPage=10&"+
				"sort=IN_FECHARESOLUCION%3Adecreasing&start=1",
			cuerpoCodificado(camposDe(referencia)))
	})

	t.Run("cada llamada da sus campos", func(t *testing.T) {
		t.Parallel()

		referencia, err := NuevaReferencia(ecliConocido, "", "", "")
		require.NoError(t, err)

		campos := camposDe(referencia)
		campos["start"] = "11"

		assert.Equal(t, "1", camposDe(referencia)["start"], "tocar los campos de un envío no cambia los del siguiente")
	})

	t.Run("sin referencia no hay nada que enviar", func(t *testing.T) {
		t.Parallel()

		assert.Empty(t, camposDe(Referencia{}),
			"una referencia que no construyó NuevaReferencia no da campos, y sin campos internal/httpx no envía nada")
	})
}

// cuerpoCodificado es el cuerpo que internal/httpx escribe con esos campos: los
// de un formulario, codificados y ordenados por clave (contrato
// httpx-formulario §3).
func cuerpoCodificado(campos map[string]string) string {
	valores := make(url.Values, len(campos))
	for nombre, valor := range campos {
		valores.Set(nombre, valor)
	}

	return valores.Encode()
}
