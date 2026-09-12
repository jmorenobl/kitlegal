package httpx

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// caracteresProhibidosEnWindows son los que ningún nombre de fichero puede
// llevar en Windows. El nombre derivado tiene que ser portable, así que se
// comprueban en **todas** las filas y no solo en las que traen un carácter
// raro: la portabilidad es una propiedad del nombre, no de la fila (contrato de
// grabación §2, research.md D12).
const caracteresProhibidosEnWindows = `<>:"/\|?*`

// largoMaximoDeLaTabla es el tope que el contrato §2 pone al nombre sin la
// extensión: pasado de ahí, el nombre se recorta y lleva el resumen de la
// petición. Escrito aquí a mano, y no tomado del código, para que rebajar el
// tope en nombre.go se vea en este fichero.
const largoMaximoDeLaTabla = 120

// extensionDeLaTabla es la que el contrato §2 da a toda grabación; también va a
// mano, por la misma razón.
const extensionDeLaTabla = ".json"

// filaDeNombre es una fila de las dos tablas de este fichero: una petición
// —método y dirección completa, que es todo lo que entra en la derivación— y el
// fichero que le corresponde.
type filaDeNombre struct {
	nombre    string
	metodo    string
	direccion string
	fichero   string
}

// TestNombreDeGrabacion fija el nombre con el que se graba una petición y con
// el que la reproducción la busca después: la única clave que une las dos
// mitades del mecanismo, derivada del método y de la dirección completa y de
// nada más, porque nadie fuera del paquete la declara (FR-039, D12).
//
// Son dos tablas que tienen que decir lo mismo desde dos sitios distintos: la
// del contrato de grabación §2, que fija las reglas, y la de los diez ficheros
// escritos a mano de plan.md §«Fixtures», que fija los nombres que T012 va a
// escribir en el árbol. Un fichero mal nombrado falla aquí antes de que ningún
// test de reproducción lo busque.
//
// Toda dirección de las dos tablas es de un host ficticio o local
// —`fuente.prueba`, `otra.prueba` o `127.0.0.1`—, que es lo único que admite la
// comprobación «solo direcciones locales» del quickstart (obligación 11 del
// plan).
func TestNombreDeGrabacion(t *testing.T) {
	t.Parallel()

	t.Run("la tabla del contrato de grabación", func(t *testing.T) {
		t.Parallel()

		exigeLaTablaDeNombres(t, filasDelContrato())
	})

	t.Run("los ficheros escritos a mano de la reproducción", func(t *testing.T) {
		t.Parallel()

		filas := filasDeLosFixtures()
		require.Len(t, filas, 10, "la lista de plan.md «Fixtures» es cerrada: diez ficheros y ninguno más")

		exigeLaTablaDeNombres(t, filas)
	})

	t.Run("dos direcciones distintas pueden compartir nombre", func(t *testing.T) {
		t.Parallel()

		conComa := nombreDeGrabacion("GET", direccionDePrueba(t, "http://fuente.prueba/a,b"))
		conGuionBajo := nombreDeGrabacion("GET", direccionDePrueba(t, "http://fuente.prueba/a_b"))

		assert.Equal(t, conComa, conGuionBajo,
			"el nombre no es único por construcción: la colisión se resuelve por contenido, "+
				"comparando la petición guardada (contrato de grabación §3 y §4, FR-039)")
	})
}

// filasDelContrato es la tabla de ejemplos del contrato de grabación §2, copiada
// tal cual. Las cinco primeras filas fijan las reglas de la derivación; las dos
// últimas son el par que colisiona.
func filasDelContrato() []filaDeNombre {
	return []filaDeNombre{
		{
			nombre:    "el puerto explícito va tras un guion, y la barra de la ruta se vuelve guion bajo",
			metodo:    "GET",
			direccion: "http://127.0.0.1:53211/robots.txt",
			fichero:   "GET_http_127.0.0.1-53211_robots.txt.json",
		},
		{
			nombre:    "la consulta va tras _q_ y sus separadores se vuelven guiones bajos",
			metodo:    "GET",
			direccion: "http://fuente.prueba/norma?id=BOE-A-2015-10565",
			fichero:   "GET_http_fuente.prueba_norma_q_id_BOE-A-2015-10565.json",
		},
		{
			nombre:    "el método entra en el nombre, y una ruta vacía no aporta nada",
			metodo:    "HEAD",
			direccion: "https://fuente.prueba/",
			fichero:   "HEAD_https_fuente.prueba.json",
		},
		{
			nombre:    "una ruta larga que no pasa de 120 caracteres se escribe entera",
			metodo:    "GET",
			direccion: "https://fuente.prueba/datosabiertos/api/legislacion-consolidada/id/BOE-A-2015-10565/texto/bloque/a21",
			fichero: "GET_https_fuente.prueba_datosabiertos_api_legislacion-consolidada_id_BOE-A-2015-10565_texto_bloque_a21" +
				".json",
		},
		{
			nombre: "pasada de 120 caracteres, los 100 primeros y el resumen de la petición",
			metodo: "GET",
			direccion: "https://otra.prueba:8443/publicaciones/2026/09/12/resolucion-de-adjudicacion-definitiva" +
				"/expediente-de-contratacion-abierto-simplificado/anexo-i/documento-1",
			fichero: "GET_https_otra.prueba-8443_publicaciones_2026_09_12_resolucion-de-adjudicacion-definitiva" +
				"_expediente-d11bc93b.json",
		},
		{
			nombre:    "la coma, que net/url conserva sin escapar, queda fuera del conjunto admitido",
			metodo:    "GET",
			direccion: "http://fuente.prueba/a,b",
			fichero:   "GET_http_fuente.prueba_a_b.json",
		},
		{
			nombre:    "y la dirección que ya traía el guion bajo produce ese mismo nombre",
			metodo:    "GET",
			direccion: "http://fuente.prueba/a_b",
			fichero:   "GET_http_fuente.prueba_a_b.json",
		},
	}
}

// filasDeLosFixtures son los diez pares petición → fichero de plan.md
// §«Fixtures»: la lista cerrada de grabaciones escritas a mano que T012 deja en
// el directorio de reproducción de internal/httpx/testdata/. Cada nombre es el
// que la derivación produce para su petición, de modo que el directorio y el
// código no puedan separarse sin que este test lo diga.
func filasDeLosFixtures() []filaDeNombre {
	return []filaDeNombre{
		{
			nombre:    "1 · la norma en XML, que sirve de destino y de respuesta de referencia",
			metodo:    "GET",
			direccion: "http://fuente.prueba/norma?id=BOE-A-2015-10565",
			fichero:   "GET_http_fuente.prueba_norma_q_id_BOE-A-2015-10565.json",
		},
		{
			nombre:    "2 · la misma dirección por HEAD, que es otra petición y otro fichero",
			metodo:    "HEAD",
			direccion: "http://fuente.prueba/norma?id=BOE-A-2015-10565",
			fichero:   "HEAD_http_fuente.prueba_norma_q_id_BOE-A-2015-10565.json",
		},
		{
			nombre:    "3 · el documento binario, cuyo punto ya es un carácter admitido",
			metodo:    "GET",
			direccion: "http://fuente.prueba/documento.pdf",
			fichero:   "GET_http_fuente.prueba_documento.pdf.json",
		},
		{
			nombre:    "4 · la dirección antigua, que redirige con una Location relativa",
			metodo:    "GET",
			direccion: "http://fuente.prueba/antigua",
			fichero:   "GET_http_fuente.prueba_antigua.json",
		},
		{
			nombre:    "5 · la dirección movida, cuyo destino no tiene grabación",
			metodo:    "GET",
			direccion: "http://fuente.prueba/movida",
			fichero:   "GET_http_fuente.prueba_movida.json",
		},
		{
			nombre:    "6 · el bucle de redirecciones",
			metodo:    "GET",
			direccion: "http://fuente.prueba/bucle",
			fichero:   "GET_http_fuente.prueba_bucle.json",
		},
		{
			nombre:    "7 · el 503 grabado",
			metodo:    "GET",
			direccion: "http://fuente.prueba/caida",
			fichero:   "GET_http_fuente.prueba_caida.json",
		},
		{
			nombre:    "8 · el 429 grabado, con su Retry-After",
			metodo:    "GET",
			direccion: "http://fuente.prueba/limitada",
			fichero:   "GET_http_fuente.prueba_limitada.json",
		},
		{
			nombre:    "9 · la que colisiona con /a_b, y por eso se resuelve por contenido",
			metodo:    "GET",
			direccion: "http://fuente.prueba/a,b",
			fichero:   "GET_http_fuente.prueba_a_b.json",
		},
		{
			nombre:    "10 · el robots.txt que la reproducción deja sin usar",
			metodo:    "GET",
			direccion: "http://fuente.prueba/robots.txt",
			fichero:   "GET_http_fuente.prueba_robots.txt.json",
		},
	}
}

// exigeLaTablaDeNombres comprueba fila a fila que la derivación produce el
// fichero declarado y que ese fichero es portable. Un fallo aquí es del código
// bajo prueba o de la tabla; nunca del entorno, porque la derivación no depende
// de nada más que de sus dos argumentos.
func exigeLaTablaDeNombres(t *testing.T, filas []filaDeNombre) {
	t.Helper()

	for _, fila := range filas {
		t.Run(fila.nombre, func(t *testing.T) {
			t.Parallel()

			obtenido := nombreDeGrabacion(fila.metodo, direccionDePrueba(t, fila.direccion))

			assert.Equal(t, fila.fichero, obtenido,
				"%s %s se graba en un solo fichero, el que fija el contrato §2", fila.metodo, fila.direccion)
			exigeNombrePortable(t, obtenido)
		})
	}
}

// exigeNombrePortable comprueba lo que hace utilizable un nombre derivado: que
// se pueda escribir en cualquier sistema de ficheros y que no crezca sin
// límite. Es la mitad de la regla que la igualdad de la tabla no cubre —una
// tabla mal copiada pasaría la igualdad y no esto—.
func exigeNombrePortable(t *testing.T, nombre string) {
	t.Helper()

	assert.NotContains(t, nombre, " ", "un nombre con espacios no es cómodo de teclear ni de citar")

	for _, prohibido := range caracteresProhibidosEnWindows {
		assert.NotContainsf(t, nombre, string(prohibido),
			"%q no puede aparecer en un nombre de fichero de Windows", prohibido)
	}

	assert.LessOrEqual(t, len(strings.TrimSuffix(nombre, extensionDeLaTabla)), largoMaximoDeLaTabla,
		"pasado el tope, el contrato §2 recorta y añade el resumen de la petición")
}
