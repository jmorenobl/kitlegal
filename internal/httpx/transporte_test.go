package httpx

import (
	"errors"
	"net/http"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestTransporteConfiguracion fija el escalón inferior de la cadena y el cliente
// que la ejecuta, que son las dos piezas donde una garantía se pierde sin que
// nada lo note: un plazo propio acortaría o alargaría el del contexto (FR-004),
// unas redirecciones seguidas por la biblioteca dejarían la política fuera del
// paquete (FR-011, D10), un almacén de cookies guardaría estado entre peticiones
// y una configuración de TLS propia sería la única forma de que la verificación
// de certificados dejara de estar activa (FR-012).
//
// Lo que el clon hereda se comprueba con valores literales a propósito: son los
// que documenta `go doc net/http.DefaultTransport`, y verlos aquí es lo que
// distingue «clonado» de «construido a mano» (D11).
func TestTransporteConfiguracion(t *testing.T) {
	t.Parallel()

	t.Run("el transporte es un clon del de la biblioteca, con su TLS intacto", func(t *testing.T) {
		t.Parallel()

		porOmision, esDeLaBiblioteca := http.DefaultTransport.(*http.Transport)
		require.True(t, esDeLaBiblioteca, "la biblioteca declara http.DefaultTransport como *http.Transport")

		propio := nuevoTransporte()
		require.NotNil(t, propio)
		assert.NotSame(t, porOmision, propio, "el transporte del cliente es un clon, no el global compartido")

		exigeVerificacionDeCertificados(t, propio)
		assert.NotNil(t, propio.Proxy, "el clon conserva el proxy del entorno")
		assert.True(t, propio.ForceAttemptHTTP2)
		assert.Equal(t, 100, propio.MaxIdleConns)
		assert.Equal(t, 90*time.Second, propio.IdleConnTimeout)
		assert.Equal(t, 10*time.Second, propio.TLSHandshakeTimeout)
		assert.Equal(t, time.Second, propio.ExpectContinueTimeout)
	})

	t.Run("el cliente no pone plazo, no sigue redirecciones y no guarda cookies", func(t *testing.T) {
		t.Parallel()

		cadena := nuevoTransporte()
		cliente := nuevoClienteHTTP(cadena)
		require.NotNil(t, cliente)

		assert.Same(t, cadena, cliente.Transport, "el cliente ejecuta la cadena que se le entrega")
		assert.Zero(t, cliente.Timeout, "el plazo de la operación es solo el del contexto (FR-004)")
		assert.Nil(t, cliente.Jar, "sin almacén de cookies: ninguna petición hereda estado de otra")

		require.NotNil(t, cliente.CheckRedirect, "sin política propia la biblioteca seguiría diez saltos por su cuenta")
		assert.ErrorIs(t, cliente.CheckRedirect(nil, nil), http.ErrUseLastResponse,
			"las redirecciones las sigue Pedir, que es quien sabe someter cada salto a la cadena entera (FR-011, D10)")
	})

	t.Run("un transporte por omisión ajeno no se adopta", func(t *testing.T) {
		t.Parallel()

		propio := clonDelTransportePorOmision(transporteAjeno{})
		require.NotNil(t, propio, "el paquete construye el suyo en vez de adoptar el ajeno")
		exigeVerificacionDeCertificados(t, propio)
	})
}

// exigeVerificacionDeCertificados comprueba la forma que puede tener FR-012: o
// el transporte no trae configuración de TLS —y entonces rige la de la
// biblioteca— o la trae con la verificación activa. Lo que no puede es traerla
// desactivada, que es lo único que este paquete no ofrece forma de hacer; el
// control mecánico que lo remata es G402 de gosec, gate desde H0.
func exigeVerificacionDeCertificados(t *testing.T, transporte *http.Transport) {
	t.Helper()

	if transporte.TLSClientConfig == nil {
		return
	}

	assert.False(t, transporte.TLSClientConfig.InsecureSkipVerify,
		"ninguna opción del paquete llega a la configuración de TLS (FR-012)")
}

// transporteAjeno es lo que se encontraría el paquete si algo del proceso
// hubiera sustituido http.DefaultTransport, que es una variable pública: un
// transporte del que no se sabe nada y cuya configuración de TLS no controla
// este paquete.
type transporteAjeno struct{}

// RoundTrip nunca llega a llamarse: el transporte ajeno está aquí para que el
// clon lo rechace, no para emitir nada.
func (transporteAjeno) RoundTrip(*http.Request) (*http.Response, error) {
	return nil, errors.New("este transporte no emite ninguna petición")
}
