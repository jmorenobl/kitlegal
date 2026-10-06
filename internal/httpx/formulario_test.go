package httpx

import (
	"context"
	"io"
	"net/http"
	"net/http/httptest"
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
)

// Las tres rutas del buscador de prueba: la página que da la cookie de sesión,
// el formulario que la espera y el destino de sus redirecciones, que ninguna
// consulta debe llegar a pedir.
const (
	rutaDeLaPagina    = "/pagina"
	rutaDelFormulario = "/buscar"
	rutaDelDestino    = "/destino"
)

// Lo que entrega cada ruta del buscador de prueba.
const (
	paginaDePrueba     = "<html><form action=\"/buscar\"></form></html>"
	resultadosDePrueba = "<html><ul><li>un resultado</li></ul></html>"
)

// cuerpoDePrueba es el cuerpo de camposDePrueba tal como tiene que salir:
// ordenado por clave —las mayúsculas van delante— y con «:», «/» y el espacio
// codificados. Está escrito a mano, y no derivado con el código que se prueba
// (contrato httpx-formulario §3, research V5).
const cuerpoDePrueba = "ECLI=ECLI%3AES%3ATS%3A2023%3A3144&action=query&desde=04%2F07%2F2023" +
	"&sort=IN_FECHARESOLUCION%3Adecreasing&texto=a+b"

// camposDePrueba son los campos de un envío, declarados en un orden que no es el
// de su cuerpo. Cada llamada da un mapa nuevo, para que ninguna tabla comparta el
// suyo con otra.
func camposDePrueba() map[string]string {
	return map[string]string{
		"sort":   "IN_FECHARESOLUCION:decreasing",
		"texto":  "a b",
		"action": "query",
		"desde":  "04/07/2023",
		"ECLI":   "ECLI:ES:TS:2023:3144",
	}
}

// pedidor es la firma que comparten Cliente.Pedir y Consulta.Pedir, que es lo que
// deja a una tabla pedir lo mismo desde los dos.
type pedidor func(context.Context, schema.Contexto, Peticion) (Respuesta, error)

// quienPide nombra desde dónde se pide en una fila de la tabla de admisión
// (contrato httpx-formulario §2).
type quienPide string

const (
	clienteSinFormulario quienPide = "Cliente.Pedir, sin formulario declarado"
	clienteConFormulario quienPide = "Cliente.Pedir, con el formulario declarado"
	desdeUnaConsulta     quienPide = "Consulta.Pedir"
)

// TestFormularioAdmitido es la tabla de admisión del contrato httpx-formulario
// §2, fila a fila, y el control de FR-030, FR-031 y SC-006: el único método que
// no es GET ni HEAD y que sale del paquete es el POST de una consulta a la
// dirección declarada, con sus campos. Todo lo demás es de la clase «argumentos»
// antes de abrir nada, también en ensayo, y eso lo demuestra el sitio de prueba,
// que no recibe ni la petición ni el robots.txt que la precedería.
func TestFormularioAdmitido(t *testing.T) {
	t.Parallel()

	t.Run("lo que se rechaza no emite ninguna petición", probarRechazosDelFormulario)
	t.Run("lo que se admite llega al sitio", probarAdmisionDelFormulario)
	t.Run("la dirección se compara tal como la escribe url.URL.String", probarDireccionDelFormulario)
	t.Run("un cliente sin formulario no da consultas", probarConsultaSinFormulario)
}

// probarRechazosDelFormulario pasa por las filas de rechazo. Cada una tiene sus
// dos sitios —el del formulario declarado y un vecino—, de modo que la que deje
// salir una petición es la única que falla, y lo que exige es que ninguno de los
// dos haya recibido nada.
func probarRechazosDelFormulario(t *testing.T) {
	t.Parallel()

	delCliente := []quienPide{clienteSinFormulario, clienteConFormulario}
	deCualquiera := []quienPide{clienteSinFormulario, clienteConFormulario, desdeUnaConsulta}
	deLaConsulta := []quienPide{desdeUnaConsulta}

	casos := []struct {
		nombre  string
		quienes []quienPide
		metodo  string
		// ruta es la de la petición en el sitio del formulario o, con alVecino, en
		// el otro.
		ruta     string
		alVecino bool
		campos   map[string]string
	}{
		{
			nombre:  "POST a la dirección declarada, con sus campos",
			quienes: delCliente,
			metodo:  http.MethodPost,
			ruta:    rutaDelFormulario,
			campos:  camposDePrueba(),
		},
		{
			nombre:  "POST a otra dirección",
			quienes: delCliente,
			metodo:  http.MethodPost,
			ruta:    rutaDeLaPagina,
			campos:  camposDePrueba(),
		},
		{
			nombre:  "POST sin campos",
			quienes: deCualquiera,
			metodo:  http.MethodPost,
			ruta:    rutaDelFormulario,
		},
		{
			nombre:  "POST con los campos vacíos",
			quienes: deCualquiera,
			metodo:  http.MethodPost,
			ruta:    rutaDelFormulario,
			campos:  map[string]string{},
		},
		{
			nombre:  "PUT",
			quienes: deCualquiera,
			metodo:  http.MethodPut,
			ruta:    rutaDelFormulario,
			campos:  camposDePrueba(),
		},
		{
			nombre:  "DELETE",
			quienes: deCualquiera,
			metodo:  http.MethodDelete,
			ruta:    rutaDelFormulario,
		},
		{
			nombre:  "PATCH",
			quienes: deCualquiera,
			metodo:  http.MethodPatch,
			ruta:    rutaDelFormulario,
			campos:  camposDePrueba(),
		},
		{
			nombre:  "un POST en minúsculas, que no es ningún método",
			quienes: deCualquiera,
			metodo:  "post",
			ruta:    rutaDelFormulario,
			campos:  camposDePrueba(),
		},
		{
			nombre:  "GET con campos",
			quienes: deCualquiera,
			metodo:  http.MethodGet,
			ruta:    rutaDeLaPagina,
			campos:  camposDePrueba(),
		},
		{
			nombre:  "GET con campos a la dirección declarada",
			quienes: deCualquiera,
			metodo:  http.MethodGet,
			ruta:    rutaDelFormulario,
			campos:  camposDePrueba(),
		},
		{
			nombre:  "HEAD con campos",
			quienes: deCualquiera,
			metodo:  http.MethodHead,
			ruta:    rutaDeLaPagina,
			campos:  camposDePrueba(),
		},
		{
			nombre:  "POST a otra ruta del mismo sitio",
			quienes: deLaConsulta,
			metodo:  http.MethodPost,
			ruta:    rutaDeLaPagina,
			campos:  camposDePrueba(),
		},
		{
			nombre:  "POST a la dirección declarada con una consulta añadida",
			quienes: deLaConsulta,
			metodo:  http.MethodPost,
			ruta:    rutaDelFormulario + "?start=11",
			campos:  camposDePrueba(),
		},
		{
			nombre:  "POST a la dirección declarada con una barra de más",
			quienes: deLaConsulta,
			metodo:  http.MethodPost,
			ruta:    rutaDelFormulario + "/",
			campos:  camposDePrueba(),
		},
		{
			nombre:   "POST a la misma ruta de otro sitio",
			quienes:  deLaConsulta,
			metodo:   http.MethodPost,
			ruta:     rutaDelFormulario,
			alVecino: true,
			campos:   camposDePrueba(),
		},
	}

	for _, caso := range casos {
		for _, quien := range caso.quienes {
			t.Run(caso.nombre+" desde "+string(quien), func(t *testing.T) {
				t.Parallel()

				servidor, sitio := sitioDePrueba(t, buscadorDePrueba())
				vecino, delVecino := sitioDePrueba(t, buscadorDePrueba())

				destino := servidor.URL
				if caso.alVecino {
					destino = vecino.URL
				}

				exigeRechazoSinPeticion(t, pedidorDe(t, quien, servidor.URL+rutaDelFormulario),
					Peticion{Metodo: caso.metodo, URL: destino + caso.ruta, Campos: caso.campos}, sitio, delVecino)
			})
		}
	}
}

// probarAdmisionDelFormulario pasa por las filas que sí se piden, cada una contra
// su propio sitio: lo que llega es el robots.txt y esa petición, y ninguna más.
func probarAdmisionDelFormulario(t *testing.T) {
	t.Parallel()

	casos := []struct {
		nombre  string
		quienes []quienPide
		metodo  string
		ruta    string
		campos  map[string]string
	}{
		{
			nombre:  "GET sin campos",
			quienes: []quienPide{clienteSinFormulario, clienteConFormulario, desdeUnaConsulta},
			metodo:  http.MethodGet,
			ruta:    rutaDeLaPagina,
		},
		{
			nombre:  "HEAD sin campos",
			quienes: []quienPide{clienteSinFormulario, clienteConFormulario, desdeUnaConsulta},
			metodo:  http.MethodHead,
			ruta:    rutaDeLaPagina,
		},
		{
			nombre:  "GET sin campos a la dirección declarada",
			quienes: []quienPide{clienteConFormulario, desdeUnaConsulta},
			metodo:  http.MethodGet,
			ruta:    rutaDelFormulario,
		},
		{
			nombre:  "POST a la dirección declarada, con sus campos",
			quienes: []quienPide{desdeUnaConsulta},
			metodo:  http.MethodPost,
			ruta:    rutaDelFormulario,
			campos:  camposDePrueba(),
		},
		{
			nombre:  "POST a la dirección declarada, con un solo campo",
			quienes: []quienPide{desdeUnaConsulta},
			metodo:  http.MethodPost,
			ruta:    rutaDelFormulario,
			campos:  map[string]string{"ROJ": "STS 3144/2023"},
		},
	}

	for _, caso := range casos {
		for _, quien := range caso.quienes {
			t.Run(caso.nombre+" desde "+string(quien), func(t *testing.T) {
				t.Parallel()

				servidor, sitio := sitioDePrueba(t, buscadorDePrueba())
				pedir := pedidorDe(t, quien, servidor.URL+rutaDelFormulario)

				respuesta, err := pedir(t.Context(), schema.Contexto{},
					Peticion{Metodo: caso.metodo, URL: servidor.URL + caso.ruta, Campos: caso.campos})
				require.NoError(t, err, "la fila se admite (contrato httpx-formulario §2)")

				assert.Equal(t, http.StatusOK, respuesta.Estado)
				assert.Equal(t, []string{http.MethodGet + " " + rutaDelRobots, caso.metodo + " " + caso.ruta},
					sitio.lineas(), "al sitio llegan su robots.txt y la petición admitida, y nada más")
			})
		}
	}
}

// probarDireccionDelFormulario fija con qué se compara la dirección de un envío:
// con la declarada, las dos tal como las escribe url.URL.String. Un esquema
// declarado en mayúsculas es la misma dirección, porque la biblioteca lo escribe
// en minúsculas; todo lo demás se compara carácter a carácter, que es lo que la
// tabla de rechazos comprueba (contrato httpx-formulario §2).
func probarDireccionDelFormulario(t *testing.T) {
	t.Parallel()

	servidor, sitio := sitioDePrueba(t, buscadorDePrueba())
	formulario := servidor.URL + rutaDelFormulario

	sinEsquema, esHTTP := strings.CutPrefix(formulario, "http://")
	require.True(t, esHTTP, "el servidor de prueba es de esquema http")

	declarada := "HTTP://" + sinEsquema
	consulta := consultaDePrueba(t, clienteDePrueba(t, ConFormulario(declarada), ConIntervalo(time.Millisecond)))

	respuesta, err := consulta.Pedir(t.Context(), schema.Contexto{},
		Peticion{Metodo: http.MethodPost, URL: formulario, Campos: camposDePrueba()})
	require.NoError(t, err)

	assert.Equal(t, http.StatusOK, respuesta.Estado)
	assert.Equal(t, []string{http.MethodGet + " " + rutaDelRobots, http.MethodPost + " " + rutaDelFormulario},
		sitio.lineas())
}

// probarConsultaSinFormulario es la fila que no pide nada: un cliente que no
// declaró ningún formulario no abre consultas, ni contra la red ni en
// reproducción, y el fallo es de argumentos, sin petición implicada y nombrando
// la opción que falta (FR-031, FR-034 de H2).
func probarConsultaSinFormulario(t *testing.T) {
	t.Parallel()

	casos := []struct {
		nombre  string
		cliente *Cliente
	}{
		{nombre: "contra la red", cliente: clienteDePrueba(t)},
		{nombre: "en reproducción", cliente: clienteDeReproduccion(t, grabacionesDePrueba(t))},
	}

	for _, caso := range casos {
		t.Run(caso.nombre, func(t *testing.T) {
			t.Parallel()

			consulta, err := caso.cliente.Consulta()

			assert.Nil(t, consulta, "sin formulario declarado no hay consulta")

			fallo := falloDe(t, err)
			assert.Equal(t, schema.ClaseArgumentos, fallo.Clase(), "lo que quien llama puede corregir es «argumentos»")
			assert.Empty(t, fallo.Peticion.URL, "abrir una consulta no tiene petición implicada")
			assert.Contains(t, fallo.Error(), "ConFormulario", "el mensaje nombra la opción que falta")
		})
	}
}

// TestEnvioDelFormulario es el contrato httpx-formulario §3 y el control de
// FR-033: un envío es un POST con el cuerpo de sus campos y sus dos cabeceras, y
// baja por la misma cadena que cualquier otra petición —identificado, con el
// robots.txt del sitio por delante, en su turno y con su instante de emisión—;
// en ensayo no sale nada.
func TestEnvioDelFormulario(t *testing.T) {
	t.Parallel()

	t.Run("lleva su cuerpo y sus cabeceras, tras el robots.txt y en su turno", probarEnvioDelFormulario)
	t.Run("en ensayo no sale nada y la respuesta describe el envío", probarEnsayoDelFormulario)
	t.Run("con el contexto terminado no sale nada", probarContextoTerminadoEnConsulta)
}

// probarEnvioDelFormulario mide el envío donde lo ve la fuente: en lo que el
// sitio recibe. El turno se mide como en TestRitmoSeparaPeticionesDelMismoSitio,
// contra un comienzo tomado antes de pedir nada: el robots.txt ocupa el turno
// cero y el envío no puede adelantarse al siguiente, una cota que fija el
// limitador y que por eso no lleva holgura.
func probarEnvioDelFormulario(t *testing.T) {
	t.Parallel()

	const intervalo = 150 * time.Millisecond

	instante := time.Date(2026, time.October, 6, 10, 0, 5, 0, time.UTC)

	servidor, sitio := sitioDePrueba(t, buscadorDePrueba())
	formulario := servidor.URL + rutaDelFormulario
	envio := Peticion{Metodo: http.MethodPost, URL: formulario, Campos: camposDePrueba()}

	consulta := consultaDePrueba(t, clienteDePrueba(t, ConFormulario(formulario), ConIntervalo(intervalo),
		ConHora(func() time.Time { return instante })))

	comienzo := time.Now()

	respuesta, err := consulta.Pedir(t.Context(), schema.Contexto{}, envio)
	require.NoError(t, err)

	assert.Equal(t, http.StatusOK, respuesta.Estado)
	assert.Equal(t, resultadosDePrueba, string(respuesta.Cuerpo), "lo que el formulario responde llega entero")
	assert.Equal(t, envio, respuesta.Peticion, "la petición devuelta es la que se pidió, con sus campos")
	assert.Equal(t, formulario, respuesta.URL)
	assert.Equal(t, instante, respuesta.Instante, "con el instante en que salió, como toda petición (FR-033)")
	assert.False(t, respuesta.Ensayo)

	recibido := sitio.recibido()
	require.Len(t, recibido, 2, "el robots.txt del sitio y el envío")

	delRobots, delEnvio := recibido[0], recibido[1]

	assert.Equal(t, http.MethodGet+" "+rutaDelRobots, delRobots.linea(),
		"el robots.txt del sitio se consulta antes que el envío (FR-033)")
	assert.Empty(t, delRobots.cuerpo, "y lo pide el paquete: no es un envío")
	assert.Empty(t, delRobots.cabeceras.Values("X-Requested-With"))

	assert.Equal(t, http.MethodPost+" "+rutaDelFormulario, delEnvio.linea())
	assert.Equal(t, cuerpoDePrueba, delEnvio.cuerpo, "el cuerpo son los campos, ordenados por clave y codificados")
	assert.Equal(t, []string{"application/x-www-form-urlencoded"}, delEnvio.cabeceras.Values("Content-Type"))
	assert.Equal(t, []string{"XMLHttpRequest"}, delEnvio.cabeceras.Values("X-Requested-With"))
	assert.Equal(t, int64(len(cuerpoDePrueba)), delEnvio.longitud, "y declara su longitud: no sale troceado")

	for _, apuntada := range recibido {
		assert.Equal(t, []string{AgenteDeUsuario()}, apuntada.cabeceras.Values("User-Agent"),
			"%s lleva la identificación del proyecto, y solo esa (FR-033)", apuntada.linea())
	}

	assert.GreaterOrEqual(t, delEnvio.instante.Sub(comienzo), intervalo,
		"el envío espera su turno en el ritmo del sitio: el robots.txt ocupó el anterior (FR-033)")
}

// probarEnsayoDelFormulario comprueba que el ensayo es el de Cliente.Pedir: nada
// sale —tampoco el robots.txt— y la respuesta declara que no se emitió, con la
// línea que el adaptador copiará en su resultado (contrato httpx-formulario §3).
func probarEnsayoDelFormulario(t *testing.T) {
	t.Parallel()

	servidor, sitio := sitioDePrueba(t, buscadorDePrueba())
	formulario := servidor.URL + rutaDelFormulario
	consulta := consultaDePrueba(t, clienteDePrueba(t, ConFormulario(formulario)))

	casos := []struct {
		peticion    Peticion
		descripcion string
	}{
		{
			peticion:    Peticion{Metodo: http.MethodPost, URL: formulario, Campos: camposDePrueba()},
			descripcion: "POST " + formulario,
		},
		{
			peticion:    Peticion{Metodo: http.MethodGet, URL: servidor.URL + rutaDeLaPagina},
			descripcion: "GET " + servidor.URL + rutaDeLaPagina,
		},
	}

	for _, caso := range casos {
		respuesta, err := consulta.Pedir(t.Context(), schema.Contexto{DryRun: true}, caso.peticion)
		require.NoError(t, err, "el ensayo no es un fallo")

		assert.True(t, respuesta.Ensayo, "la respuesta declara que la petición no se emitió")
		assert.Equal(t, caso.descripcion, respuesta.Descripcion(), "con el método y la dirección de lo que no salió")
		assert.Equal(t, caso.peticion, respuesta.Peticion)
		assert.Zero(t, respuesta.Estado, "no hay estado porque no hubo fuente que respondiera")
		assert.Nil(t, respuesta.Cuerpo)
		assert.True(t, respuesta.Instante.IsZero(), "ni instante de emisión, porque no se emitió nada")
	}

	assert.Empty(t, sitio.recibido(), "en ensayo no sale ni el envío, ni la página, ni el robots.txt (FR-033)")
}

// probarContextoTerminadoEnConsulta comprueba que el plazo de quien pide corta
// una consulta como corta cualquier petición: con el contexto ya terminado no
// sale nada y el fallo es «fuente no disponible», con la causa del contexto.
// También en reproducción, donde abajo no hay ninguna conexión que lo notara: sin
// esa comprobación, lo que se vería sería la grabación que falta (FR-033; FR-005
// y FR-049 a de H2).
func probarContextoTerminadoEnConsulta(t *testing.T) {
	t.Parallel()

	servidor, sitio := sitioDePrueba(t, buscadorDePrueba())
	formulario := servidor.URL + rutaDelFormulario
	envio := Peticion{Metodo: http.MethodPost, URL: formulario, Campos: camposDePrueba()}

	casos := []struct {
		nombre  string
		cliente *Cliente
	}{
		{nombre: "contra la red", cliente: clienteDePrueba(t, ConFormulario(formulario))},
		{nombre: "en reproducción", cliente: clienteDeReproduccion(t, grabacionesDePrueba(t), ConFormulario(formulario))},
	}

	for _, caso := range casos {
		t.Run(caso.nombre, func(t *testing.T) {
			t.Parallel()

			terminado, cancelar := context.WithCancel(t.Context())
			cancelar()

			respuesta, err := consultaDePrueba(t, caso.cliente).Pedir(terminado, schema.Contexto{}, envio)

			fallo := falloDe(t, err)
			assert.Equal(t, schema.ClaseFuenteNoDisponible, fallo.Clase())
			require.ErrorIs(t, err, context.Canceled, "la causa del contexto llega intacta hasta quien llama")
			assert.Equal(t, Respuesta{}, respuesta)
			assert.Empty(t, sitio.lineas(), "y al sitio no llega nada")
		})
	}
}

// TestCookiesDeLaConsulta es el contrato httpx-formulario §4 y el control de
// FR-032: la cookie que el sitio da en una consulta acompaña a las peticiones de
// esa consulta y a nada más. El sitio da una cookie distinta cada vez que se le
// pide la página, de modo que lo que llega en cada petición dice de qué consulta
// es.
func TestCookiesDeLaConsulta(t *testing.T) {
	t.Parallel()

	t.Run("la cookie de la página llega al formulario de su consulta, y a nada más", probarCookiesDeCadaConsulta)
	t.Run("la petición del robots.txt no lleva la cookie de ninguna consulta", probarRobotsSinCookies)
	t.Run("cada consulta tiene su almacén, sobre la cadena del cliente, que no tiene ninguno", probarAlmacenDeLaConsulta)
}

// probarCookiesDeCadaConsulta intercala dos consultas y el propio cliente contra
// el mismo sitio y compara de una vez, en el orden en que el sitio lo recibió,
// con qué cookie llegó cada petición.
func probarCookiesDeCadaConsulta(t *testing.T) {
	t.Parallel()

	servidor, sitio := sitioDePrueba(t, buscadorDePrueba())
	formulario := servidor.URL + rutaDelFormulario
	pagina := Peticion{Metodo: http.MethodGet, URL: servidor.URL + rutaDeLaPagina}
	envio := Peticion{Metodo: http.MethodPost, URL: formulario, Campos: camposDePrueba()}

	cliente := clienteDePrueba(t, ConFormulario(formulario), ConIntervalo(time.Millisecond))
	primera := consultaDePrueba(t, cliente)

	exigeEntregada(t, primera.Pedir, pagina) // el sitio da sesion=1
	exigeEntregada(t, primera.Pedir, envio)
	exigeEntregada(t, cliente.Pedir, pagina) // da sesion=2, que nadie guarda
	exigeEntregada(t, cliente.Pedir, pagina) // da sesion=3

	segunda := consultaDePrueba(t, cliente)

	exigeEntregada(t, segunda.Pedir, pagina) // da sesion=4
	exigeEntregada(t, segunda.Pedir, envio)
	exigeEntregada(t, primera.Pedir, envio)

	assert.Equal(t, []string{
		"GET /robots.txt · sin cookie",
		"GET /pagina · sin cookie",
		"POST /buscar · sesion=1",
		"GET /pagina · sin cookie",
		"GET /pagina · sin cookie",
		"GET /pagina · sin cookie",
		"POST /buscar · sesion=4",
		"POST /buscar · sesion=1",
	}, sitio.lineasConCookie(),
		"la cookie de la página llega al formulario de su consulta; Cliente.Pedir ni la lleva ni guarda la que "+
			"le dan; y la consulta siguiente empieza sin ninguna y se queda con la suya (FR-032)")
}

// probarRobotsSinCookies pone al paquete a pedir un robots.txt cuando la consulta
// ya tiene una cookie que valdría para él. Las cookies no distinguen puertos
// (RFC 6265 §8.5) y el robots.txt sí: el otro servidor local es el mismo host
// para el almacén de la consulta y un sitio nuevo para el cliente, que le pide
// su robots.txt antes que nada.
func probarRobotsSinCookies(t *testing.T) {
	t.Parallel()

	servidor, sitio := sitioDePrueba(t, buscadorDePrueba())
	vecino, delVecino := sitioDePrueba(t, redireccionesDePrueba())

	consulta := consultaDePrueba(t,
		clienteDePrueba(t, ConFormulario(servidor.URL+rutaDelFormulario), ConIntervalo(time.Millisecond)))

	exigeEntregada(t, consulta.Pedir, Peticion{Metodo: http.MethodGet, URL: servidor.URL + rutaDeLaPagina})
	exigeEntregada(t, consulta.Pedir, Peticion{Metodo: http.MethodGet, URL: vecino.URL + "/norma"})

	assert.Equal(t, []string{"GET /robots.txt · sin cookie", "GET /pagina · sin cookie"}, sitio.lineasConCookie())
	assert.Equal(t, []string{"GET /robots.txt · sin cookie", "GET /norma · sesion=1"}, delVecino.lineasConCookie(),
		"la cookie de la consulta valía para el vecino, y aun así su robots.txt llegó sin ella")
}

// probarAlmacenDeLaConsulta mira la forma de lo que las otras dos subpruebas
// miden: el almacén es de cada consulta, el cliente no tiene ninguno y las dos
// piden por la cadena del cliente, que es una sola (research D3).
func probarAlmacenDeLaConsulta(t *testing.T) {
	t.Parallel()

	cliente := clienteDePrueba(t, ConFormulario("http://fuente.prueba/buscar"))
	una, otra := consultaDePrueba(t, cliente), consultaDePrueba(t, cliente)

	assert.Nil(t, cliente.cliente.Jar, "Cliente.Pedir sigue sin almacén de cookies, también con formulario declarado")
	require.NotNil(t, una.emisor.Jar)
	require.NotNil(t, otra.emisor.Jar)
	assert.NotSame(t, una.emisor.Jar, otra.emisor.Jar, "el almacén es de la consulta, no del cliente")
	assert.Same(t, cliente.cliente.Transport, una.emisor.Transport,
		"la consulta pide por la misma cadena que su cliente: una sola identificación, un solo registro de "+
			"sitios, un solo ritmo (FR-033)")
	assert.Same(t, cliente.cliente.Transport, otra.emisor.Transport)
}

// TestConsultaSinRedirecciones es la otra mitad del contrato httpx-formulario §4
// y de FR-021: dentro de una consulta ninguna redirección se sigue —ni la de la
// página ni, sobre todo, la del formulario, que no se reenvía a ninguna parte—,
// y la respuesta 3xx se entrega con su estado y su Location. Fuera de una
// consulta, el mismo cliente las sigue como siempre.
func TestConsultaSinRedirecciones(t *testing.T) {
	t.Parallel()

	seguibles := []int{
		http.StatusMovedPermanently, http.StatusFound, http.StatusSeeOther,
		http.StatusTemporaryRedirect, http.StatusPermanentRedirect,
	}

	for _, estado := range seguibles {
		t.Run("un "+strconv.Itoa(estado)+" se entrega y su destino no recibe nada", func(t *testing.T) {
			t.Parallel()

			exigeRedireccionEntregada(t, estado)
		})
	}

	t.Run("el resto de la clasificación no cambia", probarClasificacionEnConsulta)
}

// exigeRedireccionEntregada comprueba un estado de redirección contra su propio
// sitio: la página y el formulario redirigen los dos al mismo destino, que solo
// recibe una petición cuando quien pide es el cliente, fuera de la consulta.
func exigeRedireccionEntregada(t *testing.T, estado int) {
	t.Helper()

	servidor, sitio := sitioDePrueba(t, buscadorQueResponde(estado))
	formulario := servidor.URL + rutaDelFormulario
	pagina := Peticion{Metodo: http.MethodGet, URL: servidor.URL + rutaDeLaPagina}
	envio := Peticion{Metodo: http.MethodPost, URL: formulario, Campos: camposDePrueba()}

	cliente := clienteDePrueba(t, ConFormulario(formulario), ConIntervalo(time.Millisecond))
	consulta := consultaDePrueba(t, cliente)

	for _, pedida := range []Peticion{pagina, envio} {
		respuesta, err := consulta.Pedir(t.Context(), schema.Contexto{}, pedida)
		require.NoError(t, err, "una redirección dentro de una consulta se entrega, no es un fallo (FR-021)")

		assert.Equal(t, estado, respuesta.Estado, "con su estado")
		assert.Equal(t, rutaDelDestino, respuesta.Cabeceras.Get("Location"), "y sus cabeceras")
		assert.Equal(t, pedida.URL, respuesta.URL, "la dirección es la pedida: no se ha ido a ninguna otra")
		assert.Equal(t, pedida, respuesta.Peticion)
	}

	enConsulta := []string{"GET /robots.txt", "GET /pagina", "POST /buscar"}
	assert.Equal(t, enConsulta, sitio.lineas(),
		"el destino no recibe nada: ni la página se sigue ni el formulario se reenvía (FR-021)")

	respuesta, err := cliente.Pedir(t.Context(), schema.Contexto{}, pagina)
	require.NoError(t, err)

	assert.Equal(t, http.StatusOK, respuesta.Estado, "fuera de una consulta la redirección se sigue como hoy")
	assert.Equal(t, servidor.URL+rutaDelDestino, respuesta.URL)
	assert.Equal(t, append(enConsulta, "GET /pagina", "GET /destino"), sitio.lineas())
}

// probarClasificacionEnConsulta fija lo que una consulta no cambia: un 429 sigue
// siendo límite o términos de uso, un 5xx la fuente que no sabe entregar, y los
// demás estados se entregan. Lo único nuevo es el 3xx, que se entrega aunque no
// sea de los cinco que fuera de una consulta se siguen (contrato
// httpx-formulario §4).
//
// El cliente declara un solo intento, como el de la fuente que usa el formulario
// (research D7): así el 5xx termina a la primera.
func probarClasificacionEnConsulta(t *testing.T) {
	t.Parallel()

	casos := []struct {
		estado int
		clase  schema.Clase
	}{
		{estado: http.StatusTooManyRequests, clase: schema.ClaseLimiteOTos},
		{estado: http.StatusServiceUnavailable, clase: schema.ClaseFuenteNoDisponible},
		{estado: http.StatusForbidden},
		{estado: http.StatusNotFound},
		{estado: http.StatusMultipleChoices},
		{estado: http.StatusNotModified},
	}

	for _, caso := range casos {
		t.Run(strconv.Itoa(caso.estado), func(t *testing.T) {
			t.Parallel()

			servidor, sitio := sitioDePrueba(t, buscadorQueResponde(caso.estado))
			formulario := servidor.URL + rutaDelFormulario
			consulta := consultaDePrueba(t,
				clienteDePrueba(t, ConFormulario(formulario), ConIntervalo(time.Millisecond), ConIntentos(1)))

			pedidas := []Peticion{
				{Metodo: http.MethodGet, URL: servidor.URL + rutaDeLaPagina},
				{Metodo: http.MethodPost, URL: formulario, Campos: camposDePrueba()},
			}

			for _, pedida := range pedidas {
				respuesta, err := consulta.Pedir(t.Context(), schema.Contexto{}, pedida)

				if caso.clase == "" {
					require.NoError(t, err, "el estado %d se entrega a quien pide", caso.estado)
					assert.Equal(t, caso.estado, respuesta.Estado)

					continue
				}

				fallo := falloDe(t, err)
				assert.Equal(t, caso.clase, fallo.Clase())
				assert.Equal(t, caso.estado, fallo.Estado)
				assert.Equal(t, Respuesta{}, respuesta)
			}

			assert.Equal(t, []string{"GET /robots.txt", "GET /pagina", "POST /buscar"}, sitio.lineas(),
				"una petición por dirección, y ninguna a otra")
		})
	}
}

// pedidorDe construye quien pide en una fila de la tabla de admisión: un cliente
// nuevo, con el formulario declarado o sin él, o una consulta de un cliente que
// lo declara.
func pedidorDe(t *testing.T, quien quienPide, formulario string) pedidor {
	t.Helper()

	switch quien {
	case clienteSinFormulario:
		return clienteDePrueba(t, ConIntervalo(time.Millisecond)).Pedir
	case clienteConFormulario:
		return clienteDePrueba(t, ConFormulario(formulario), ConIntervalo(time.Millisecond)).Pedir
	default:
		return consultaDePrueba(t, clienteDePrueba(t, ConFormulario(formulario), ConIntervalo(time.Millisecond))).Pedir
	}
}

// consultaDePrueba abre una consulta del cliente. Un fallo aquí es de la
// construcción, no de lo que el test comprueba.
func consultaDePrueba(t *testing.T, cliente *Cliente) *Consulta {
	t.Helper()

	consulta, err := cliente.Consulta()
	require.NoError(t, err, "un cliente con formulario declarado abre consultas")
	require.NotNil(t, consulta)

	return consulta
}

// exigeRechazoSinPeticion comprueba las dos mitades de un rechazo de la tabla de
// admisión, y las dos también en ensayo: que la clase es «argumentos» y que
// ningún sitio ha recibido nada, ni siquiera el robots.txt que precedería a la
// petición si hubiera llegado a bajar por la cadena (FR-031).
func exigeRechazoSinPeticion(t *testing.T, pedir pedidor, p Peticion, sitios ...*sitioQueApunta) {
	t.Helper()

	for _, ejecucion := range []schema.Contexto{{}, {DryRun: true}} {
		respuesta, err := pedir(t.Context(), ejecucion, p)

		fallo := falloDe(t, err)
		assert.Equal(t, schema.ClaseArgumentos, fallo.Clase(),
			"lo que quien llama puede corregir es «argumentos» (ensayo: %t)", ejecucion.DryRun)
		assert.Equal(t, Respuesta{}, respuesta, "un rechazo no entrega ninguna respuesta")
		assert.Equal(t, p, fallo.Peticion, "y nombra la petición rechazada")
	}

	for _, sitio := range sitios {
		assert.Empty(t, sitio.lineas(), "una petición rechazada no abre ninguna conexión, tampoco en ensayo")
	}
}

// exigeEntregada pide y exige que la petición se entregue con un 200: lo que
// estas tablas miran no es la respuesta, sino lo que el sitio recibió.
func exigeEntregada(t *testing.T, pedir pedidor, p Peticion) {
	t.Helper()

	respuesta, err := pedir(t.Context(), schema.Contexto{}, p)
	require.NoError(t, err, "%s %s se entrega", p.Metodo, p.URL)
	require.Equal(t, http.StatusOK, respuesta.Estado)
}

// apunte es lo que el sitio de prueba se queda de cada petición que recibe.
type apunte struct {
	metodo    string
	ruta      string
	cuerpo    string
	longitud  int64
	cabeceras http.Header
	instante  time.Time
}

// linea es el método y la ruta, separados por un espacio: la forma en que las
// tablas comparan lo recibido.
func (a apunte) linea() string {
	return a.metodo + " " + a.ruta
}

// sitioQueApunta apunta, en su orden, todo lo que un servidor de prueba recibe:
// método, ruta con su consulta, cuerpo con la longitud que declaraba, cabeceras e
// instante de llegada. Es lo que deja a estas tablas afirmar qué salió del
// paquete y qué no, que es donde FR-030 a FR-033 se miden.
type sitioQueApunta struct {
	mu      sync.Mutex
	apuntes []apunte
}

// sitioDePrueba levanta un servidor local con un robots.txt sin reglas y lo que
// atienda el manejador, y apunta todo lo que recibe, el robots.txt incluido.
func sitioDePrueba(t *testing.T, manejador http.HandlerFunc) (*httptest.Server, *sitioQueApunta) {
	t.Helper()

	sitio := &sitioQueApunta{}

	return servidorLocal(t, sitio.apunta(robotsSinReglas(manejador))), sitio
}

// apunta envuelve un manejador para apuntar cada petición antes de atenderla. Lee
// el cuerpo entero, que es lo que hace falta para compararlo.
func (s *sitioQueApunta) apunta(manejador http.HandlerFunc) http.HandlerFunc {
	return func(escritor http.ResponseWriter, peticion *http.Request) {
		cuerpo, err := io.ReadAll(peticion.Body)
		if err != nil {
			http.Error(escritor, err.Error(), http.StatusBadRequest)

			return
		}

		s.mu.Lock()
		s.apuntes = append(s.apuntes, apunte{
			metodo:    peticion.Method,
			ruta:      peticion.URL.RequestURI(),
			cuerpo:    string(cuerpo),
			longitud:  peticion.ContentLength,
			cabeceras: peticion.Header.Clone(),
			instante:  time.Now(),
		})
		s.mu.Unlock()

		manejador(escritor, peticion)
	}
}

// recibido devuelve una copia de lo apuntado hasta ahora, en su orden.
func (s *sitioQueApunta) recibido() []apunte {
	s.mu.Lock()
	defer s.mu.Unlock()

	return slices.Clone(s.apuntes)
}

// lineas devuelve la línea de cada petición recibida, en su orden.
func (s *sitioQueApunta) lineas() []string {
	recibido := s.recibido()
	lineas := make([]string, 0, len(recibido))

	for _, apuntada := range recibido {
		lineas = append(lineas, apuntada.linea())
	}

	return lineas
}

// lineasConCookie devuelve la línea de cada petición recibida con la cookie que
// traía, o «sin cookie» si llegó sin ninguna.
func (s *sitioQueApunta) lineasConCookie() []string {
	recibido := s.recibido()
	lineas := make([]string, 0, len(recibido))

	for _, apuntada := range recibido {
		cookie := "sin cookie"
		if valores := apuntada.cabeceras.Values("Cookie"); len(valores) > 0 {
			cookie = valores[0]
		}

		lineas = append(lineas, apuntada.linea()+" · "+cookie)
	}

	return lineas
}

// buscadorDePrueba imita lo que estas tablas necesitan de un buscador: una
// página que da una cookie de sesión —distinta cada vez, numerada desde 1, para
// que cada petición diga de qué consulta es— y un formulario que responde con su
// lista de resultados.
func buscadorDePrueba() http.HandlerFunc {
	var sesiones atomic.Int64

	return func(escritor http.ResponseWriter, peticion *http.Request) {
		switch peticion.URL.Path {
		case rutaDeLaPagina:
			escritor.Header().Set("Set-Cookie", "sesion="+strconv.FormatInt(sesiones.Add(1), 10)+"; Path=/")
			_, _ = io.WriteString(escritor, paginaDePrueba)
		case rutaDelFormulario:
			_, _ = io.WriteString(escritor, resultadosDePrueba)
		default:
			http.NotFound(escritor, peticion)
		}
	}
}

// buscadorQueResponde es el buscador cuya página y cuyo formulario responden con
// el estado que se le declara, y con una Location que apunta a un destino que sí
// existe: si alguien siguiera la redirección, el sitio lo apuntaría.
func buscadorQueResponde(estado int) http.HandlerFunc {
	return func(escritor http.ResponseWriter, peticion *http.Request) {
		switch peticion.URL.Path {
		case rutaDeLaPagina, rutaDelFormulario:
			escritor.Header().Set("Location", rutaDelDestino)
			escritor.WriteHeader(estado)
		case rutaDelDestino:
			_, _ = io.WriteString(escritor, resultadosDePrueba)
		default:
			http.NotFound(escritor, peticion)
		}
	}
}
