package cendoj

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/jmorenobl/kitlegal/internal/httpx"
)

// Las respuestas del formulario que TestClasificar clasifica, escritas aquí: la
// estructura real de la página de resultados no se ha visto todavía (research
// S1), y la clase solo depende de las dos marcas que documenta
// docs/JURISPRUDENCIA.md §3.
const (
	paginaSinResultados = `<div class="sinResultados">No se ha encontrado ningún resultado</div>`
	paginaConLista      = `<div class="resultado">` +
		`<a href="/search/AN/openDocument/0123456789abcdef0123456789abcdef/20230714">STS 3144/2023</a></div>`
	paginaSinMarcas = `<html><head><link rel="stylesheet" href="/search/css/captcha.css"></head>` +
		`<body><form action="/search/validar.action">Escriba el texto de la imagen</form></body></html>`
)

// TestClasificar fija la clase de una respuesta del formulario (contrato
// fuente-cendoj-y-grabacion §3; FR-022): con estado 200, el texto «No se ha
// encontrado ningún resultado» es «sin resultados»; sin ese texto y con algún
// enlace /search/AN/openDocument/, «lista»; y cualquier otra, con el estado que
// sea, «no reconocida». El «no encontrado» nunca sale de un estado.
func TestClasificar(t *testing.T) {
	t.Parallel()

	casos := []struct {
		nombre    string
		respuesta httpx.Respuesta
		esperada  claseDeRespuesta
	}{
		{
			nombre:    "sin resultados",
			respuesta: httpx.Respuesta{Estado: 200, Cuerpo: []byte(paginaSinResultados)},
			esperada:  respuestaSinResultados,
		},
		{
			nombre:    "una lista",
			respuesta: httpx.Respuesta{Estado: 200, Cuerpo: []byte(paginaConLista)},
			esperada:  respuestaConLista,
		},
		{
			// La lista es la respuesta que no lleva el texto: con los dos, manda él.
			nombre:    "el texto de sin resultados junto a un enlace",
			respuesta: httpx.Respuesta{Estado: 200, Cuerpo: []byte(paginaSinResultados + paginaConLista)},
			esperada:  respuestaSinResultados,
		},
		{
			// Como llega un CAPTCHA: un 200 que no es ninguna de las dos.
			nombre:    "una página sin ninguna marca",
			respuesta: httpx.Respuesta{Estado: 200, Cuerpo: []byte(paginaSinMarcas)},
			esperada:  respuestaNoReconocida,
		},
		{
			nombre:    "un 200 sin cuerpo",
			respuesta: httpx.Respuesta{Estado: 200},
			esperada:  respuestaNoReconocida,
		},
		{
			nombre:    "un 403",
			respuesta: httpx.Respuesta{Estado: 403, Cuerpo: []byte("Forbidden")},
			esperada:  respuestaNoReconocida,
		},
		{
			nombre:    "un 403 con el texto de sin resultados",
			respuesta: httpx.Respuesta{Estado: 403, Cuerpo: []byte(paginaSinResultados)},
			esperada:  respuestaNoReconocida,
		},
		{
			nombre:    "un 404",
			respuesta: httpx.Respuesta{Estado: 404, Cuerpo: []byte(paginaSinResultados)},
			esperada:  respuestaNoReconocida,
		},
		{
			nombre:    "un 404 con un enlace",
			respuesta: httpx.Respuesta{Estado: 404, Cuerpo: []byte(paginaConLista)},
			esperada:  respuestaNoReconocida,
		},
		{
			nombre: "una redirección",
			respuesta: httpx.Respuesta{
				Estado:    302,
				Cabeceras: httpx.Cabeceras{"Location": {"https://www.poderjudicial.es/search/indexAN.jsp"}},
				Cuerpo:    []byte(paginaConLista),
			},
			esperada: respuestaNoReconocida,
		},
		{
			nombre:    "un 204",
			respuesta: httpx.Respuesta{Estado: 204},
			esperada:  respuestaNoReconocida,
		},
		{
			// Un ensayo no emite nada y no trae estado: no hay respuesta que
			// reconocer.
			nombre:    "un ensayo",
			respuesta: httpx.Respuesta{Ensayo: true},
			esperada:  respuestaNoReconocida,
		},
	}

	for _, caso := range casos {
		t.Run(caso.nombre, func(t *testing.T) {
			t.Parallel()

			assert.Equal(t, caso.esperada, clasificar(caso.respuesta))
		})
	}
}
