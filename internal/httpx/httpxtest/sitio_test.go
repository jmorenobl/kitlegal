package httpxtest_test

import (
	"bufio"
	"context"
	"fmt"
	"io"
	"net"
	"os"
	"slices"
	"strconv"
	"strings"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/jmorenobl/kitlegal/internal/core/schema"
	"github.com/jmorenobl/kitlegal/internal/httpx"
	"github.com/jmorenobl/kitlegal/internal/httpx/httpxtest"
)

// Los dos métodos que piden estas tablas. Van escritos aquí, y no tomados de la
// biblioteca HTTP, porque un test de fuera de internal/httpx —que es quien usa
// el sitio— no puede importarla (regla R2), y lo que estas tablas piden y
// comparan es lo que ese test podrá pedir y comparar.
const (
	metodoGET  = "GET"
	metodoPOST = "POST"
)

// Las rutas del buscador de prueba: su robots.txt, la página que da la cookie de
// sesión —pedida con su consulta, que también es parte de lo que se apunta— y el
// formulario que la espera.
const (
	rutaDelRobots     = "/robots.txt"
	rutaDeLaPagina    = "/pagina?idioma=es"
	rutaDelFormulario = "/buscar"
)

// Lo que entrega cada ruta del buscador de prueba, y la cookie que da su página.
const (
	paginaDePrueba     = "<html><form action=\"/buscar\"></form></html>"
	resultadosDePrueba = "<html><ul><li>un resultado</li></ul></html>"
	cookieDeLaPagina   = "sesion=1"
)

// cuerpoDelEnvio es el cuerpo de camposDelEnvio tal como sale de httpx: ordenado
// por clave —las mayúsculas van delante— y con «:» codificado. Está escrito a
// mano, y no derivado con el código que lo envía.
const cuerpoDelEnvio = "ECLI=ECLI%3AES%3ATS%3A2023%3A3144&action=query"

// plazoDeLectura acota lo que estas tablas esperan a leer de una conexión
// abierta a mano. No mide nada: solo impide que un test se quede colgado si el
// sitio no responde ni cierra.
const plazoDeLectura = 30 * time.Second

// camposDelEnvio son los campos del envío de estas tablas. Cada llamada da un
// mapa nuevo.
func camposDelEnvio() map[string]string {
	return map[string]string{"action": "query", "ECLI": "ECLI:ES:TS:2023:3144"}
}

// TestSitio es el contrato httpx-formulario §7 de H23: el sitio local de prueba
// apunta cada petición que recibe —método, ruta, cuerpo y cookie—, en su orden y
// sin carreras; lo que responder decide —estado, cabeceras y cuerpo— es lo que
// llega a quien pide; y el servidor se cierra con el test.
//
// Todo se pide con un cliente de httpx, por su cadena entera: el robots.txt del
// sitio es por eso la primera petición que cada cliente le hace, y el sitio la
// apunta como cualquier otra. Es lo que deja al control de FR-021 contar, desde
// el paquete del adaptador, las peticiones de una consulta.
func TestSitio(t *testing.T) {
	t.Parallel()

	t.Run("apunta el robots.txt, un GET y un envío con su cookie, en su orden", probarLoApuntado)
	t.Run("lo que responder decide llega al cliente", probarLoRespondido)
	t.Run("apunta sin carreras lo que le llega a la vez", probarSinCarreras)
	t.Run("un cuerpo que no llega entero se apunta y hace fallar al test", probarCuerpoSinTerminar)
	t.Run("se cierra con el test", probarCierre)
}

// buscadorDePrueba responde como el buscador que estas tablas necesitan: un
// robots.txt sin reglas, una página que da una cookie de sesión y un formulario
// que entrega su lista de resultados. Decide con lo que el sitio le da de cada
// petición, que es todo lo que un test de fuera de internal/httpx puede mirar.
func buscadorDePrueba(recibida httpxtest.Recibida) httpxtest.Respuesta {
	switch recibida.Ruta {
	case rutaDelRobots:
		return httpxtest.Respuesta{}
	case rutaDeLaPagina:
		return httpxtest.Respuesta{
			Cabeceras: httpx.Cabeceras{"Set-Cookie": {cookieDeLaPagina + "; Path=/"}},
			Cuerpo:    paginaDePrueba,
		}
	case rutaDelFormulario:
		return httpxtest.Respuesta{Cuerpo: resultadosDePrueba}
	default:
		return httpxtest.Respuesta{Estado: 404}
	}
}

// probarLoApuntado hace la secuencia de una consulta —la página y, con la cookie
// que da, el envío del formulario— y compara de una vez todo lo que el sitio
// apuntó: las tres peticiones, el robots.txt delante, cada una con su método, su
// ruta, su cuerpo y su cookie.
func probarLoApuntado(t *testing.T) {
	t.Parallel()

	sitio := httpxtest.NuevoSitio(t, buscadorDePrueba)
	formulario := sitio.URL() + rutaDelFormulario
	consulta := consultaDePrueba(t, formulario)

	pagina := pedir(t, consulta.Pedir, httpx.Peticion{Metodo: metodoGET, URL: sitio.URL() + rutaDeLaPagina})
	envio := pedir(t, consulta.Pedir, httpx.Peticion{Metodo: metodoPOST, URL: formulario, Campos: camposDelEnvio()})

	assert.Equal(t, paginaDePrueba, string(pagina.Cuerpo))
	assert.Equal(t, resultadosDePrueba, string(envio.Cuerpo))

	apuntado := []httpxtest.Recibida{
		{Metodo: metodoGET, Ruta: rutaDelRobots},
		{Metodo: metodoGET, Ruta: rutaDeLaPagina},
		{Metodo: metodoPOST, Ruta: rutaDelFormulario, Cuerpo: cuerpoDelEnvio, Cookie: cookieDeLaPagina},
	}

	recibidas := sitio.Recibidas()
	assert.Equal(t, apuntado, recibidas,
		"el sitio apunta cada petición que recibe, en su orden: el robots.txt que el cliente pide antes que "+
			"nada, la página con su consulta y el envío con su cuerpo y con la cookie que la página dio")

	require.NotEmpty(t, recibidas)
	recibidas[0] = httpxtest.Recibida{Metodo: "DELETE", Ruta: "/otra"}

	assert.Equal(t, apuntado, sitio.Recibidas(),
		"lo apuntado se entrega en una copia: quien la toca no cambia lo que el sitio recibió")
}

// probarLoRespondido declara una respuesta por ruta y comprueba que cada una
// llega al cliente como responder la decidió: su estado, cada una de sus
// cabeceras con todos sus valores y su cuerpo. Se piden dentro de una consulta,
// que es donde una redirección se entrega en vez de seguirse.
func probarLoRespondido(t *testing.T) {
	t.Parallel()

	casos := []struct {
		nombre string
		ruta   string
		decide httpxtest.Respuesta
		estado int
	}{
		{
			nombre: "un 200 con sus cabeceras, también la de dos valores, y su cuerpo",
			ruta:   "/entregada",
			decide: httpxtest.Respuesta{
				Estado: 200,
				Cabeceras: httpx.Cabeceras{
					"Content-Type": {"application/xml; charset=utf-8"},
					"X-Prueba":     {"uno", "dos"},
				},
				Cuerpo: "<norma>contenido</norma>",
			},
			estado: 200,
		},
		{
			nombre: "el valor cero es un 200 sin cuerpo",
			ruta:   "/vacia",
			decide: httpxtest.Respuesta{},
			estado: 200,
		},
		{
			nombre: "una redirección con su Location",
			ruta:   "/movida",
			decide: httpxtest.Respuesta{Estado: 302, Cabeceras: httpx.Cabeceras{"Location": {"/destino"}}},
			estado: 302,
		},
		{
			nombre: "un 403 con su cuerpo",
			ruta:   "/vetada",
			decide: httpxtest.Respuesta{Estado: 403, Cuerpo: "acceso denegado"},
			estado: 403,
		},
		{
			nombre: "un 404 con su cuerpo",
			ruta:   "/ausente",
			decide: httpxtest.Respuesta{Estado: 404, Cuerpo: "no está"},
			estado: 404,
		},
	}

	// El robots.txt no es de ningún caso: recibe el valor cero, que es un 200
	// sin cuerpo y, para el cliente, un sitio sin reglas.
	sitio := httpxtest.NuevoSitio(t, func(recibida httpxtest.Recibida) httpxtest.Respuesta {
		for _, caso := range casos {
			if caso.ruta == recibida.Ruta {
				return caso.decide
			}
		}

		return httpxtest.Respuesta{}
	})
	formulario := sitio.URL() + rutaDelFormulario

	for _, caso := range casos {
		t.Run(caso.nombre, func(t *testing.T) {
			t.Parallel()

			consulta := consultaDePrueba(t, formulario)
			respuesta := pedir(t, consulta.Pedir, httpx.Peticion{Metodo: metodoGET, URL: sitio.URL() + caso.ruta})

			assert.Equal(t, caso.estado, respuesta.Estado)
			assert.Equal(t, caso.decide.Cuerpo, string(respuesta.Cuerpo))

			for nombre, valores := range caso.decide.Cabeceras {
				assert.Equal(t, valores, respuesta.Cabeceras[nombre], "la cabecera %s llega con todos sus valores", nombre)
			}
		})
	}
}

// probarSinCarreras pide al sitio a la vez desde varias gorrutinas —cada una con
// un cliente suyo, con su ritmo y con su propio robots.txt por delante— y lee lo
// apuntado mientras el sitio sigue apuntando. Con el detector de carreras, que
// es como make test lo ejecuta, un apunte sin proteger no pasa de aquí.
//
// Cada gorrutina deja lo suyo en su casilla y las aserciones se hacen después,
// en la del test.
func probarSinCarreras(t *testing.T) {
	t.Parallel()

	const cuantas = 8

	sitio := httpxtest.NuevoSitio(t, func(httpxtest.Recibida) httpxtest.Respuesta {
		return httpxtest.Respuesta{}
	})

	var (
		propias   = make([]httpxtest.Recibida, cuantas)
		fallos    = make([]error, cuantas)
		apuntadas = make([][]httpxtest.Recibida, cuantas)
		esperadas = make([]httpxtest.Recibida, 0, 2*cuantas)

		grupo sync.WaitGroup
	)

	for numero := range cuantas {
		propias[numero] = httpxtest.Recibida{Metodo: metodoGET, Ruta: "/norma/" + strconv.Itoa(numero)}
		esperadas = append(esperadas, httpxtest.Recibida{Metodo: metodoGET, Ruta: rutaDelRobots}, propias[numero])

		cliente := clienteDePrueba(t)
		peticion := httpx.Peticion{Metodo: metodoGET, URL: sitio.URL() + propias[numero].Ruta}

		grupo.Go(func() {
			_, fallos[numero] = cliente.Pedir(t.Context(), schema.Contexto{}, peticion)
			apuntadas[numero] = sitio.Recibidas()
		})
	}

	grupo.Wait()

	for numero := range cuantas {
		require.NoError(t, fallos[numero], "GET %s se entrega", propias[numero].Ruta)
		assert.Contains(t, apuntadas[numero], propias[numero],
			"el sitio apunta antes de responder: quien ya tiene su respuesta encuentra su petición apuntada")
	}

	assert.ElementsMatch(t, esperadas, sitio.Recibidas(),
		"ni una petición de más ni una de menos: un robots.txt y un GET por gorrutina")
}

// probarCuerpoSinTerminar envía a mano un formulario que declara más cuerpo del
// que trae y cierra su mitad de la conexión. Es una petición que el sitio ha
// recibido, y la apunta con lo que llegó; pero no es la que quien la envió quiso
// hacer, así que no se la da a responder: se lo dice al test, que falla, y a
// quien pide, con un 400.
func probarCuerpoSinTerminar(t *testing.T) {
	t.Parallel()

	var decididas atomic.Int64

	falso := &testigo{TB: t}
	sitio := httpxtest.NuevoSitio(falso, func(httpxtest.Recibida) httpxtest.Respuesta {
		decididas.Add(1)

		return httpxtest.Respuesta{}
	})

	t.Cleanup(falso.terminar)

	conexion := conectar(t, sitio)

	_, err := io.WriteString(conexion, metodoPOST+" "+rutaDelFormulario+" HTTP/1.1\r\n"+
		"Host: sitio.prueba\r\n"+
		"Content-Type: application/x-www-form-urlencoded\r\n"+
		"Content-Length: "+strconv.Itoa(len(cuerpoDelEnvio))+"\r\n"+
		"Cookie: "+cookieDeLaPagina+"\r\n"+
		"\r\n"+
		"ECLI=")
	require.NoError(t, err)
	require.NoError(t, conexion.CloseWrite())

	estado, err := bufio.NewReader(conexion).ReadString('\n')
	require.NoError(t, err, "el sitio responde también a la petición que no le llegó entera")
	assert.Equal(t, "HTTP/1.1 400 Bad Request\r\n", estado)

	assert.Equal(t, []httpxtest.Recibida{
		{Metodo: metodoPOST, Ruta: rutaDelFormulario, Cuerpo: "ECLI=", Cookie: cookieDeLaPagina},
	}, sitio.Recibidas(), "la petición se recibió, y se apunta con el cuerpo que llegó")
	assert.Zero(t, decididas.Load(), "responder no decide nada de una petición que no llegó entera")

	fallos := falso.fallosApuntados()
	require.Len(t, fallos, 1, "el test que usa el sitio falla: lo que cuenta no es lo que quiso enviar")
	assert.Contains(t, fallos[0], metodoPOST+" "+rutaDelFormulario, "y el fallo nombra la petición")
}

// probarCierre mira que el sitio deja el cierre de su servidor para cuando el
// test termine, y que ese cierre lo cierra de verdad. Lo mide en una conexión
// abierta antes y que no pide nada: mientras el test dura, el sitio atiende; en
// cuanto termina, el sitio la cierra. No se mide pidiendo otra vez a la misma
// dirección, que otro servidor de otro test podría haber ocupado entre tanto.
func probarCierre(t *testing.T) {
	t.Parallel()

	falso := &testigo{TB: t}
	sitio := httpxtest.NuevoSitio(falso, func(httpxtest.Recibida) httpxtest.Respuesta {
		return httpxtest.Respuesta{}
	})

	// Si una aserción termina el test antes de tiempo, el servidor se cierra
	// igual.
	t.Cleanup(falso.terminar)

	conexion := conectar(t, sitio)

	pedir(t, clienteDePrueba(t).Pedir, httpx.Peticion{Metodo: metodoGET, URL: sitio.URL() + "/abierto"})
	assert.Equal(t, []httpxtest.Recibida{
		{Metodo: metodoGET, Ruta: rutaDelRobots},
		{Metodo: metodoGET, Ruta: "/abierto"},
	}, sitio.Recibidas(), "mientras el test dura, el sitio atiende")

	// Sin nada que ejecutar al final no hay cierre que medir, y la lectura de más
	// abajo solo agotaría su plazo.
	require.Equal(t, 1, falso.pendientes(), "el sitio deja una cosa para el final del test: cerrar su servidor")

	falso.terminar()

	_, err := conexion.Read(make([]byte, 1))
	require.Error(t, err, "con el test terminado, el sitio ha cerrado la conexión que tenía abierta")
	assert.NotErrorIs(t, err, os.ErrDeadlineExceeded,
		"la lectura termina porque el sitio cerró, no porque se acabara la espera")
}

// testigo es el testing.TB que se le da al sitio cuando lo que se mira es lo que
// el sitio hace con su test: guarda lo que el sitio deja para cuando el test
// termine y los fallos que le apunta, en vez de cerrarlo cuando no toca o de
// hacer fallar al test que lo mira. Todo lo demás es del test de verdad, que
// lleva dentro.
type testigo struct {
	testing.TB

	mu      sync.Mutex
	finales []func()
	fallos  []string
}

// Helper no marca nada: el testigo no informa de ninguna línea.
func (g *testigo) Helper() {}

// Cleanup guarda lo que el sitio deja para el final del test.
func (g *testigo) Cleanup(final func()) {
	g.mu.Lock()
	defer g.mu.Unlock()

	g.finales = append(g.finales, final)
}

// Errorf guarda el fallo que el sitio apunta. Le llega desde la gorrutina del
// servidor que atiende la petición.
func (g *testigo) Errorf(formato string, argumentos ...any) {
	g.mu.Lock()
	defer g.mu.Unlock()

	g.fallos = append(g.fallos, fmt.Sprintf(formato, argumentos...))
}

// pendientes dice cuántas cosas ha dejado el sitio para el final del test.
func (g *testigo) pendientes() int {
	g.mu.Lock()
	defer g.mu.Unlock()

	return len(g.finales)
}

// fallosApuntados devuelve una copia de los fallos que el sitio ha apuntado.
func (g *testigo) fallosApuntados() []string {
	g.mu.Lock()
	defer g.mu.Unlock()

	return slices.Clone(g.fallos)
}

// terminar hace lo que haría el final del test: ejecuta lo que el sitio dejó, lo
// último primero, y una sola vez aunque se le llame dos.
func (g *testigo) terminar() {
	g.mu.Lock()
	finales := g.finales
	g.finales = nil
	g.mu.Unlock()

	for _, final := range slices.Backward(finales) {
		final()
	}
}

// pedidor es la firma que comparten Cliente.Pedir y Consulta.Pedir.
type pedidor func(context.Context, schema.Contexto, httpx.Peticion) (httpx.Respuesta, error)

// pedir pide y exige que la petición se entregue, con el estado que sea: lo que
// estas tablas miran es lo que el sitio recibió y lo que respondió.
func pedir(t *testing.T, por pedidor, p httpx.Peticion) httpx.Respuesta {
	t.Helper()

	respuesta, err := por(t.Context(), schema.Contexto{}, p)
	require.NoError(t, err, "%s %s se entrega", p.Metodo, p.URL)

	return respuesta
}

// clienteDePrueba construye un cliente de httpx contra la red, con sus garantías
// puestas —identificación, robots.txt, reintentos— y un intervalo corto entre
// peticiones, que es lo único que estas tablas no necesitan esperar.
func clienteDePrueba(t *testing.T, opciones ...httpx.Opcion) *httpx.Cliente {
	t.Helper()

	cliente, err := httpx.New(append([]httpx.Opcion{httpx.ConIntervalo(time.Millisecond)}, opciones...)...)
	require.NoError(t, err, "el cliente de la tabla debe construirse sin error")

	return cliente
}

// consultaDePrueba abre una consulta de un cliente nuevo que declara ese
// formulario.
func consultaDePrueba(t *testing.T, formulario string) *httpx.Consulta {
	t.Helper()

	consulta, err := clienteDePrueba(t, httpx.ConFormulario(formulario)).Consulta()
	require.NoError(t, err, "un cliente con formulario declarado abre consultas")

	return consulta
}

// conectar abre una conexión con el sitio sin pasar por ningún cliente HTTP, que
// es lo que deja a un test enviarle lo que un cliente no enviaría y ver cuándo la
// cierra. Se cierra con el test, y su lectura no espera más que plazoDeLectura.
func conectar(t *testing.T, sitio *httpxtest.Sitio) *net.TCPConn {
	t.Helper()

	var marcador net.Dialer

	abierta, err := marcador.DialContext(t.Context(), "tcp", strings.TrimPrefix(sitio.URL(), "http://"))
	require.NoError(t, err, "el sitio escucha en la dirección que da")

	t.Cleanup(func() { assert.NoError(t, abierta.Close()) })

	conexion, esTCP := abierta.(*net.TCPConn)
	require.True(t, esTCP, "el sitio escucha en TCP")
	require.NoError(t, conexion.SetReadDeadline(time.Now().Add(plazoDeLectura)))

	return conexion
}
