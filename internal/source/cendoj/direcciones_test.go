package cendoj

import (
	"errors"
	"net/url"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/jmorenobl/kitlegal/internal/core/schema"
	"github.com/jmorenobl/kitlegal/internal/httpx"
	"github.com/jmorenobl/kitlegal/internal/httpx/httpxtest"
)

// Los métodos de las peticiones de los tests. Se escriben aquí y no con
// net/http, que solo importa internal/httpx (R2).
const (
	metodoGET  = "GET"
	metodoPOST = "POST"
)

// TestDirecciones fija byte a byte las direcciones que la fuente pide y cita, y
// los tres campos de consulta del formulario (contrato
// fuente-cendoj-y-grabacion §1). De cada dirección exige lo que el sobre de la
// fuente necesita para citarla: https, del sitio del buscador, fuera del
// espacio reservado y aceptable como procedencia junto al nombre de la fuente.
func TestDirecciones(t *testing.T) {
	t.Parallel()

	casos := []struct {
		nombre    string
		direccion string
		esperada  string
	}{
		{nombre: "la base", direccion: baseDelBuscador, esperada: "https://www.poderjudicial.es/search"},
		{
			nombre:    "la página del buscador",
			direccion: paginaDelBuscador,
			esperada:  "https://www.poderjudicial.es/search/indexAN.jsp",
		},
		{
			nombre:    "el formulario",
			direccion: formularioDeConsulta,
			esperada:  "https://www.poderjudicial.es/search/search.action",
		},
	}

	for _, caso := range casos {
		t.Run(caso.nombre, func(t *testing.T) {
			t.Parallel()

			assert.Equal(t, caso.esperada, caso.direccion)
			assert.False(t, strings.HasPrefix(caso.direccion, "kitlegal:"),
				"la dirección %q es del espacio reservado", caso.direccion)

			destino, err := url.Parse(caso.direccion)
			require.NoError(t, err)
			assert.Equal(t, "https", destino.Scheme)
			assert.Equal(t, "www.poderjudicial.es", destino.Host)
			assert.Nil(t, destino.User)
			require.NoError(t, schema.Procedencia{Fuente: NombreDeLaFuente, URL: caso.direccion}.Validar())
		})
	}

	t.Run("la fuente que las firma no es del espacio reservado", func(t *testing.T) {
		t.Parallel()

		assert.Equal(t, "cendoj.jurisprudencia", NombreDeLaFuente)
		assert.False(t, strings.HasPrefix(NombreDeLaFuente, "kitlegal."),
			"la fuente %q firma con el espacio reservado de lo que no consulta ninguna fuente", NombreDeLaFuente)
	})

	t.Run("los campos de consulta", func(t *testing.T) {
		t.Parallel()

		assert.Equal(t, "ECLI", campoECLI)
		assert.Equal(t, "ROJ", campoROJ)
		assert.Equal(t, "NUMERORESOLUCION", campoNumeroDeResolucion)
	})
}

// TestOpcionesDeLaFuente fija las opciones de internal/httpx con las que la
// fuente se pide (contrato fuente-cendoj-y-grabacion §1; data-model §7): las de
// red y las de reproducción construyen su cliente y declaran el formulario, de
// modo que el cliente abre consultas y solo admite el envío a esa dirección; y
// las de red declaran además un solo intento, que por eso no valen para
// reproducir y que se cuenta en un sitio de prueba.
func TestOpcionesDeLaFuente(t *testing.T) {
	t.Parallel()

	t.Run("las de red declaran el formulario", func(t *testing.T) {
		t.Parallel()

		cliente, err := httpx.New(OpcionesDeRed()...)
		require.NoError(t, err)

		compruebaElFormularioDeclarado(t, cliente)
	})

	t.Run("las de reproducción declaran el formulario", func(t *testing.T) {
		t.Parallel()

		cliente, err := httpx.Replay(t.TempDir(), OpcionesDeReproduccion()...)
		require.NoError(t, err)

		compruebaElFormularioDeclarado(t, cliente)
	})

	t.Run("las de red declaran los intentos, que la reproducción no admite", func(t *testing.T) {
		t.Parallel()

		_, err := httpx.Replay(t.TempDir(), OpcionesDeRed()...)

		compruebaLaClase(t, err, schema.ClaseArgumentos)
		require.ErrorContains(t, err, "ConIntentos")
	})

	t.Run("las de red piden una sola vez", func(t *testing.T) {
		t.Parallel()

		const rutaDeLaPagina = "/search/indexAN.jsp"

		sitio := httpxtest.NuevoSitio(t, func(recibida httpxtest.Recibida) httpxtest.Respuesta {
			if recibida.Ruta == rutaDeLaPagina {
				return httpxtest.Respuesta{Estado: 503}
			}

			return httpxtest.Respuesta{}
		})

		cliente, err := httpx.New(append(OpcionesDeRed(), httpx.ConIntervalo(time.Millisecond))...)
		require.NoError(t, err)

		_, err = cliente.Pedir(t.Context(), schema.Contexto{},
			httpx.Peticion{Metodo: metodoGET, URL: sitio.URL() + rutaDeLaPagina})

		compruebaLaClase(t, err, schema.ClaseFuenteNoDisponible)
		assert.Equal(t, []httpxtest.Recibida{
			{Metodo: metodoGET, Ruta: "/robots.txt"},
			{Metodo: metodoGET, Ruta: rutaDeLaPagina},
		}, sitio.Recibidas(), "un 503 no se vuelve a pedir: el robots.txt y la página, una vez cada uno")
	})
}

// compruebaElFormularioDeclarado exige que el cliente abra consultas y que, en
// ensayo y sin que salga nada, admita el envío al formulario de la fuente y
// rechace el envío a cualquier otra dirección: el formulario que declaran sus
// opciones es ese y ninguno más.
func compruebaElFormularioDeclarado(t *testing.T, cliente *httpx.Cliente) {
	t.Helper()

	consulta, err := cliente.Consulta()
	require.NoError(t, err)

	ensayo := schema.Contexto{DryRun: true}
	campos := map[string]string{campoECLI: "ECLI:ES:TS:2023:3144"}

	respuesta, err := consulta.Pedir(t.Context(), ensayo,
		httpx.Peticion{Metodo: metodoPOST, URL: formularioDeConsulta, Campos: campos})
	require.NoError(t, err)
	assert.True(t, respuesta.Ensayo)
	assert.Equal(t, "POST https://www.poderjudicial.es/search/search.action", respuesta.Descripcion())

	_, err = consulta.Pedir(t.Context(), ensayo,
		httpx.Peticion{Metodo: metodoPOST, URL: paginaDelBuscador, Campos: campos})
	compruebaLaClase(t, err, schema.ClaseArgumentos)
}

// compruebaLaClase exige que el error exista y declare esa clase, que es con la
// que el kernel lo traduce a su código de salida.
func compruebaLaClase(t *testing.T, err error, clase schema.Clase) {
	t.Helper()

	require.Error(t, err)

	conClase, laDeclara := errors.AsType[schema.ConClase](err)
	require.True(t, laDeclara, "el error %q no declara ninguna clase", err)
	assert.Equal(t, clase, conClase.Clase())
}
