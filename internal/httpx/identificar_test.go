package httpx

import (
	"go/ast"
	"go/parser"
	"go/token"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"sync/atomic"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/jmorenobl/kitlegal/internal/core/schema"
)

// constructorDeLasPeticiones es el único fichero del paquete autorizado a
// construir un *http.Request. Que sea uno solo es lo que hace que la
// identificación tenga un único dueño y que ninguna petición pueda nacer sin
// ella, venga de donde venga (FR-009, D3).
const constructorDeLasPeticiones = "identificar.go"

// TestIdentificacionEnTodaPeticion es el control literal de SC-001: el servidor
// cuenta cuántas peticiones le llegaron sin la identificación exacta del
// proyecto, y ese contador tiene que quedarse en cero por cada procedencia
// posible de una petición.
//
// Nació con las dos procedencias que la cadena de T006 permitía —el recurso
// pedido y cada salto de una redirección—, T008 le añadió la de los reintentos
// al entrar ese escalón en la cadena, y T009 la de «robots.txt», que es la cuarta
// y última (plan.md, control 6).
func TestIdentificacionEnTodaPeticion(t *testing.T) {
	t.Parallel()

	t.Run("recurso", func(t *testing.T) {
		t.Parallel()

		servidor, contador := servidorIdentificado(t, redireccionesDePrueba())
		cliente := clienteDePrueba(t)

		_, err := cliente.Pedir(t.Context(), schema.Contexto{},
			Peticion{Metodo: http.MethodGet, URL: servidor.URL + "/norma"})
		require.NoError(t, err)

		assert.Equal(t, int64(1), contador.total.Load(), "el recurso pedido es una petición")
		assert.Zero(t, contador.sinIdentificar.Load(), "la petición del recurso lleva la identificación (FR-006)")
	})

	t.Run("saltos", func(t *testing.T) {
		t.Parallel()

		servidor, contador := servidorIdentificado(t, redireccionesDePrueba())
		cliente := clienteDePrueba(t)

		_, err := cliente.Pedir(t.Context(), schema.Contexto{},
			Peticion{Metodo: http.MethodGet, URL: servidor.URL + "/antigua"})
		require.NoError(t, err)

		assert.Equal(t, int64(2), contador.total.Load(), "el salto es una petición más")
		assert.Zero(t, contador.sinIdentificar.Load(),
			"cada salto de la cadena baja por el decorador y nace identificado (FR-009)")
	})

	t.Run("reintentos", func(t *testing.T) {
		t.Parallel()

		esperas := &esperasAnotadas{}
		servidor, contador := servidorIdentificado(t, servidorQueFalla(2, http.StatusServiceUnavailable))
		cliente := clienteDePrueba(t, ConIntervalo(time.Millisecond), conReloj(esperas.reloj))

		_, err := cliente.Pedir(t.Context(), schema.Contexto{},
			Peticion{Metodo: http.MethodGet, URL: servidor.URL + "/norma"})
		require.NoError(t, err)

		assert.Equal(t, int64(3), contador.total.Load(), "cada reintento es una petición más")
		assert.Zero(t, contador.sinIdentificar.Load(),
			"cada reintento vuelve a bajar por el decorador y nace identificado (FR-009)")
	})

	t.Run("robots.txt", func(t *testing.T) {
		t.Parallel()

		servidor, contador := servidorIdentificado(t, redireccionesDePrueba())
		cliente := clienteDePrueba(t, ConIntervalo(time.Millisecond))

		_, err := cliente.Pedir(t.Context(), schema.Contexto{},
			Peticion{Metodo: http.MethodGet, URL: servidor.URL + "/norma"})
		require.NoError(t, err)

		assert.Equal(t, int64(1), contador.robots.Load(),
			"antes del recurso, el sitio recibe la petición de su robots.txt (FR-013)")
		assert.Zero(t, contador.sinIdentificar.Load(),
			"que el paquete fabrica por debajo del decorador y aun así nace identificada (FR-009, D3)")
	})
}

// TestSoloIdentificarConstruyePeticiones es la otra mitad de la garantía, y la
// que no depende de que alguien se acuerde de ampliar el test anterior: un
// decorador solo ve lo que baja desde arriba, así que una petición fabricada por
// debajo de él —el bucle de redirecciones de cliente.go, la de robots.txt en
// T009— saldría sin cabecera si la construyera por su cuenta. Recorrer el
// paquete y exigir que solo identificar.go construya un *http.Request convierte
// esa disciplina en un control mecánico (FR-009, D3, plan.md control 6).
func TestSoloIdentificarConstruyePeticiones(t *testing.T) {
	t.Parallel()

	entradas, err := os.ReadDir(".")
	require.NoError(t, err, "el test lee los ficheros del propio paquete")

	conjunto := token.NewFileSet()
	recorridos := 0

	for _, entrada := range entradas {
		nombre := entrada.Name()
		if entrada.IsDir() || filepath.Ext(nombre) != ".go" || nombre == constructorDeLasPeticiones {
			continue
		}

		fichero, err := parser.ParseFile(conjunto, nombre, nil, parser.SkipObjectResolution)
		require.NoError(t, err, "no se ha podido interpretar el fichero %s del paquete", nombre)

		recorridos++

		ast.Inspect(fichero, func(nodo ast.Node) bool {
			forma, construye := construccionDePeticion(nodo)
			if construye {
				t.Errorf("%s construye una petición HTTP con %s; solo %s puede hacerlo, que es lo que "+
					"garantiza que ninguna petición del paquete nazca sin identificación (FR-009, D3)",
					conjunto.Position(nodo.Pos()), forma, constructorDeLasPeticiones)
			}

			return true
		})
	}

	assert.Positive(t, recorridos, "el recorrido tiene que haber leído ficheros del paquete")
}

// construccionDePeticion reconoce las dos formas de construir una petición: la
// llamada al constructor de la biblioteca —cualquiera de sus dos nombres, para
// que el prohibido por noctx tampoco pase inadvertido aquí— y el literal del
// tipo, que la esquivaría.
func construccionDePeticion(nodo ast.Node) (string, bool) {
	switch tipo := nodo.(type) {
	case *ast.CallExpr:
		if nombre, esDeHTTP := simboloDeHTTP(tipo.Fun); esDeHTTP && strings.HasPrefix(nombre, "NewRequest") {
			return "http." + nombre, true
		}
	case *ast.CompositeLit:
		if nombre, esDeHTTP := simboloDeHTTP(tipo.Type); esDeHTTP && nombre == "Request" {
			return "un literal http.Request", true
		}
	}

	return "", false
}

// simboloDeHTTP devuelve el nombre del símbolo cuando la expresión es «http.X»,
// que es como se nombra la biblioteca HTTP en todo el paquete.
func simboloDeHTTP(expresion ast.Expr) (string, bool) {
	selector, esSelector := expresion.(*ast.SelectorExpr)
	if !esSelector {
		return "", false
	}

	paquete, esIdentificador := selector.X.(*ast.Ident)
	if !esIdentificador || paquete.Name != "http" {
		return "", false
	}

	return selector.Sel.Name, true
}

// contadorDeIdentificacion es lo que convierte SC-001 en una cifra: cuántas
// peticiones llegaron al servidor y cuántas de ellas lo hicieron sin la
// identificación exacta. La segunda tiene que ser siempre cero, y por eso se
// cuenta lo que falta y no lo que está: una cabecera distinta —la de la
// biblioteca, «Go-http-client/1.1»— suma igual que ninguna.
//
// Las peticiones del robots.txt del sitio se cuentan aparte de las del recurso:
// la identificación se les exige igual —son la cuarta procedencia de SC-001—,
// pero no son ninguna de las peticiones que cada tabla cuenta, que son las que
// quien llama pidió.
type contadorDeIdentificacion struct {
	total          atomic.Int64
	robots         atomic.Int64
	sinIdentificar atomic.Int64
}

// vigila envuelve el manejador de un servidor de prueba para contar cada
// petición antes de atenderla.
func (c *contadorDeIdentificacion) vigila(manejador http.HandlerFunc) http.HandlerFunc {
	return func(escritor http.ResponseWriter, peticion *http.Request) {
		if peticion.Header.Get("User-Agent") != AgenteDeUsuario() {
			c.sinIdentificar.Add(1)
		}

		if peticion.URL.Path == rutaDelRobots {
			c.robots.Add(1)
		} else {
			c.total.Add(1)
		}

		manejador(escritor, peticion)
	}
}

// servidorIdentificado levanta un servidor local —ningún test del hito conoce
// una dirección externa— con su contador de identificación delante.
//
// El sitio publica además un robots.txt que no declara ninguna regla, porque
// desde que el decorador entra en la cadena todo sitio recibe esa petición antes
// que ninguna otra (FR-013): sin ella, cada tabla del paquete estaría
// comprobando de paso qué hace su servidor con una ruta que no es la suya. Los
// sitios cuyo robots.txt **es** lo que se prueba los levanta
// servidorConRobots (robots_test.go).
func servidorIdentificado(t *testing.T, manejador http.HandlerFunc) (*httptest.Server, *contadorDeIdentificacion) {
	t.Helper()

	contador := &contadorDeIdentificacion{}

	return servidorLocal(t, contador.vigila(robotsSinReglas(manejador))), contador
}

// robotsSinReglas responde al robots.txt del sitio con un cuerpo vacío —«sin
// reglas», que permite toda ruta (FR-015)— y deja el resto al manejador que
// envuelve, que así nunca ve una petición que no es la que el test le hizo.
func robotsSinReglas(manejador http.HandlerFunc) http.HandlerFunc {
	return func(escritor http.ResponseWriter, peticion *http.Request) {
		if peticion.URL.Path == rutaDelRobots {
			escritor.Header().Set("Content-Type", "text/plain; charset=utf-8")

			return
		}

		manejador(escritor, peticion)
	}
}

// servidorLocal levanta el servidor de prueba de una tabla y lo cierra al
// terminar. Ningún test del hito conoce una dirección externa: todos se miden
// contra un servidor de este.
func servidorLocal(t *testing.T, manejador http.HandlerFunc) *httptest.Server {
	t.Helper()

	servidor := httptest.NewServer(manejador)
	t.Cleanup(servidor.Close)

	return servidor
}

// contenidoDePrueba es lo que entrega el servidor de las tablas de este paquete.
const contenidoDePrueba = "<norma>contenido</norma>"

// redireccionesDePrueba es el servidor que comparten los tests de la cadena: una
// dirección antigua que redirige con una Location relativa, el recurso al que
// lleva, una redirección hacia una ruta que ya no existe y el «no encontrado»
// para todo lo demás.
func redireccionesDePrueba() http.HandlerFunc {
	return func(escritor http.ResponseWriter, peticion *http.Request) {
		switch peticion.URL.Path {
		case "/antigua":
			http.Redirect(escritor, peticion, "/norma", http.StatusFound)
		case "/movida":
			http.Redirect(escritor, peticion, "/perdida", http.StatusMovedPermanently)
		case "/norma":
			escritor.Header().Set("Content-Type", "application/xml; charset=utf-8")
			_, _ = io.WriteString(escritor, contenidoDePrueba)
		default:
			http.NotFound(escritor, peticion)
		}
	}
}
