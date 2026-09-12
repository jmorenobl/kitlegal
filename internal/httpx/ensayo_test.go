package httpx

import (
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/jmorenobl/kitlegal/internal/core/schema"
)

// peticionDeEnsayo es la que describen estas tablas. Va contra un host ficticio
// a propósito: en ensayo no se emite nada, así que ni siquiera hace falta un
// servidor al que no llamar.
var peticionDeEnsayo = Peticion{Metodo: http.MethodGet, URL: "http://fuente.prueba/norma?id=BOE-A-2015-10565"}

// TestEnsayoNoAbreConexion es el control literal de SC-013 y del escenario 1 de
// US6: con --dry-run el cliente no abre ninguna conexión, ni al recurso ni a
// robots.txt. El contador del servidor tiene que quedarse en cero, que es lo que
// distingue «no se emitió» de «se emitió y se descartó» (FR-050).
func TestEnsayoNoAbreConexion(t *testing.T) {
	t.Parallel()

	servidor, contador := servidorIdentificado(t, redireccionesDePrueba())

	respuesta, err := clienteDePrueba(t).Pedir(t.Context(), schema.Contexto{DryRun: true},
		Peticion{Metodo: http.MethodGet, URL: servidor.URL + "/antigua"})
	require.NoError(t, err, "el ensayo no es un fallo (FR-065)")

	assert.True(t, respuesta.Ensayo)
	assert.Zero(t, contador.total.Load(), "ni el recurso ni ninguna redirección suya se piden en ensayo (FR-050)")
}

// TestEnsayoDevuelveRespuestaDeclarada comprueba lo que quien llama recibe:
// una respuesta del propio paquete que **declara** que no se emitió nada, sin
// estado de la fuente y sin cuerpo simulado, con la descripción que el adaptador
// copiará en schema.Resultado.Ensayo para que el kernel la presente (FR-052,
// FR-065, D6).
func TestEnsayoDevuelveRespuestaDeclarada(t *testing.T) {
	t.Parallel()

	respuesta, err := clienteDePrueba(t).Pedir(t.Context(), schema.Contexto{DryRun: true}, peticionDeEnsayo)
	require.NoError(t, err, "--dry-run no es un fallo: un error sin clase saldría con código 1 (FR-065)")

	assert.True(t, respuesta.Ensayo, "la respuesta declara que la petición no se emitió")
	assert.Equal(t, peticionDeEnsayo, respuesta.Peticion)
	assert.Equal(t, peticionDeEnsayo.URL, respuesta.URL, "sin saltos, la dirección final es la pedida")
	assert.Zero(t, respuesta.Estado, "no hay estado porque no hubo fuente que respondiera")
	assert.Nil(t, respuesta.Cuerpo, "no se entrega ningún cuerpo simulado (FR-052)")
	assert.Empty(t, respuesta.Cabeceras)
	assert.Equal(t, "GET http://fuente.prueba/norma?id=BOE-A-2015-10565", respuesta.Descripcion())
}

// TestEnsayoCompruebaArgumentos fija la lista cerrada de lo que sí ocurre en
// ensayo: se comprueban el método, la dirección y la configuración, que son las
// tres cosas que no necesitan red. Una invocación inválida sigue siendo inválida
// bajo --dry-run, y su código de salida sigue siendo el 2 (FR-050, FR-065).
func TestEnsayoCompruebaArgumentos(t *testing.T) {
	t.Parallel()

	ensayo := schema.Contexto{DryRun: true}

	t.Run("el método se comprueba igual", func(t *testing.T) {
		t.Parallel()

		servidor, contador := servidorIdentificado(t, redireccionesDePrueba())

		exigeArgumentosSinConexion(t, clienteDePrueba(t), contador, ensayo,
			Peticion{Metodo: http.MethodPost, URL: servidor.URL + "/norma"})
	})

	t.Run("la dirección se comprueba igual", func(t *testing.T) {
		t.Parallel()

		_, contador := servidorIdentificado(t, redireccionesDePrueba())

		exigeArgumentosSinConexion(t, clienteDePrueba(t), contador, ensayo,
			Peticion{Metodo: http.MethodGet, URL: "/norma"})
	})

	t.Run("la configuración se comprueba antes, en la construcción", func(t *testing.T) {
		t.Parallel()

		cliente, err := New(ConFuente(""))

		assert.Nil(t, cliente)

		fallo := falloDe(t, err)
		assert.Equal(t, schema.ClaseArgumentos, fallo.Clase(),
			"una configuración mal formada no llega a ensayarse: el cliente no se construye (FR-065)")
	})
}
