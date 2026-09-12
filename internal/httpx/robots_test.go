package httpx

import (
	"context"
	"io"
	"net/http"
	"net/http/httptest"
	"strconv"
	"sync/atomic"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/jmorenobl/kitlegal/internal/core/schema"
)

// TestRobotsDeniegaLaRuta es el control literal del hito: una ruta que el
// robots.txt del sitio desautoriza **no se pide**, y quien llama recibe la clase
// «límite de peticiones o términos de uso». Lo que lo demuestra es el contador de
// rutas del servidor, que se queda en cero: la diferencia entre «no se emitió» y
// «se emitió y se descartó» (FR-014, SC-011).
//
// Cada fila pide después una ruta que las mismas reglas sí autorizan, cuando la
// hay: es lo que fija que la denegación es de la ruta y no del sitio, y lo que
// permite comprobar que lo evaluado es la ruta **con su consulta** —RequestURI(),
// que es lo que RFC 9309 §2.2.2 compara—, porque la última fila solo desautoriza
// la consulta y la misma ruta sin ella tiene que pedirse.
func TestRobotsDeniegaLaRuta(t *testing.T) {
	t.Parallel()

	casos := []struct {
		nombre        string
		robots        string
		desautorizada string
		autorizada    string
	}{
		{
			nombre:        "el grupo del comodín desautoriza la ruta",
			robots:        "User-agent: *\nDisallow: /privado/\n",
			desautorizada: "/privado/norma",
			autorizada:    "/norma",
		},
		{
			nombre:        "el grupo del agente del proyecto gana al del comodín",
			robots:        "User-agent: *\nAllow: /\n\nUser-agent: kitlegal\nDisallow: /privado/\n",
			desautorizada: "/privado/norma",
			autorizada:    "/norma",
		},
		{
			nombre:        "un sitio que lo desautoriza todo no deja pedir nada",
			robots:        "User-agent: *\nDisallow: /\n",
			desautorizada: "/privado/norma",
		},
		{
			nombre:        "la consulta forma parte de la ruta que se evalúa",
			robots:        "User-agent: *\nDisallow: /norma?id=\n",
			desautorizada: "/norma?id=BOE-A-2015-10565",
			autorizada:    "/norma",
		},
	}

	for _, caso := range casos {
		t.Run(caso.nombre, func(t *testing.T) {
			t.Parallel()

			servidor, contador := servidorConRobots(t, robotsQueDice(caso.robots), redireccionesDePrueba())
			cliente := clienteDePrueba(t, ConIntervalo(time.Millisecond))

			respuesta, err := cliente.Pedir(t.Context(), schema.Contexto{},
				Peticion{Metodo: http.MethodGet, URL: servidor.URL + caso.desautorizada})

			fallo := falloDe(t, err)
			assert.Equal(t, schema.ClaseLimiteOTos, fallo.Clase(),
				"una ruta desautorizada es «límite de peticiones o términos de uso» (FR-014)")
			assert.Contains(t, fallo.Error(), "robots.txt", "el mensaje dice quién desautoriza la ruta (FR-034)")
			assert.Equal(t, Respuesta{}, respuesta, "una ruta desautorizada no entrega ninguna respuesta")
			assert.Zero(t, contador.total.Load(), "la ruta desautorizada no llega a pedirse (FR-014)")
			assert.Equal(t, int64(1), contador.robots.Load(), "y el permiso se consultó una vez (FR-013)")
			assert.Zero(t, contador.sinIdentificar.Load(), "también la del robots.txt llega identificada (FR-009)")

			if caso.autorizada == "" {
				return
			}

			respuesta, err = cliente.Pedir(t.Context(), schema.Contexto{},
				Peticion{Metodo: http.MethodGet, URL: servidor.URL + caso.autorizada})
			require.NoError(t, err,
				"las mismas reglas autorizan esta otra ruta: lo denegado es la ruta, no el sitio (FR-014)")
			assert.Equal(t, http.StatusOK, respuesta.Estado)
			assert.Equal(t, int64(1), contador.total.Load(), "la ruta autorizada sí se pide")
			assert.Equal(t, int64(1), contador.robots.Load(), "con las reglas que ya estaban cacheadas (FR-015)")
		})
	}
}

// TestRobotsSePideUnaVezPorSitio es el control literal de SC-003: diez rutas del
// mismo sitio producen **una sola** obtención de robots.txt, porque la decisión
// se cachea por sitio, en memoria y sin caducidad, mientras vive el cliente
// (FR-015, FR-018). Esa única obtención llega además con la identificación
// exacta del proyecto, que es la cuarta procedencia de SC-001 (FR-009).
func TestRobotsSePideUnaVezPorSitio(t *testing.T) {
	t.Parallel()

	const rutas = 10

	servidor, contador := servidorConRobots(t,
		robotsQueDice("User-agent: *\nAllow: /\n"), redireccionesDePrueba())
	cliente := clienteDePrueba(t, ConIntervalo(time.Millisecond))

	for consulta := range rutas {
		respuesta, err := cliente.Pedir(t.Context(), schema.Contexto{},
			Peticion{Metodo: http.MethodGet, URL: servidor.URL + "/norma?n=" + strconv.Itoa(consulta)})
		require.NoError(t, err)
		require.Equal(t, http.StatusOK, respuesta.Estado)
	}

	assert.Equal(t, int64(rutas), contador.total.Load(), "las diez rutas se piden")
	assert.Equal(t, int64(1), contador.robots.Load(),
		"y el robots.txt del sitio, una sola vez por sitio y cliente (SC-003, FR-015)")
	assert.Zero(t, contador.sinIdentificar.Load(),
		"la obtención del robots.txt lleva la identificación del proyecto (FR-009, SC-001)")
}

// TestRobotsCasosDeObtencion repasa la lista **cerrada** de FR-015: todo lo que
// puede terminar una petición de robots.txt y qué hace el cliente con ello. Cada
// fila pide dos rutas del sitio para comprobar de paso que el resultado —permitir
// o denegar— se cachea igual, de modo que un sitio caído, limitado o inalcanzable
// no se vuelva a pedir por cada ruta (FR-015, D7).
func TestRobotsCasosDeObtencion(t *testing.T) {
	t.Parallel()

	casos := []struct {
		nombre      string
		robots      http.HandlerFunc
		permitida   bool
		estado      int
		espera      time.Duration
		obtenciones int64
	}{
		{
			nombre:      "un «no encontrado» es un sitio sin reglas y permite",
			robots:      robotsConEstado(http.StatusNotFound),
			permitida:   true,
			obtenciones: 1,
		},
		{
			nombre:      "cualquier otra 4xx también permite",
			robots:      robotsConEstado(http.StatusForbidden),
			permitida:   true,
			obtenciones: 1,
		},
		{
			nombre:      "un cuerpo vacío no declara ninguna regla y permite",
			robots:      robotsQueDice(""),
			permitida:   true,
			obtenciones: 1,
		},
		{
			nombre:      "un robots.txt que no se puede interpretar deniega",
			robots:      robotsQueDice("Disallow: /norma\n"),
			estado:      http.StatusOK,
			obtenciones: 1,
		},
		{
			nombre:      "un error del servidor deniega agotados los intentos",
			robots:      robotsConEstado(http.StatusServiceUnavailable),
			estado:      http.StatusServiceUnavailable,
			obtenciones: intentosPorOmision,
		},
		{
			nombre:      "un 429 deniega sin reintentarlo y con lo que la fuente dijo que hay que esperar",
			robots:      robotsLimitado("120"),
			estado:      http.StatusTooManyRequests,
			espera:      2 * time.Minute,
			obtenciones: 1,
		},
		{
			nombre:      "un estado no previsto deniega: el permiso no se pudo obtener",
			robots:      robotsConEstado(http.StatusMultipleChoices),
			estado:      http.StatusMultipleChoices,
			obtenciones: 1,
		},
	}

	for _, caso := range casos {
		t.Run(caso.nombre, func(t *testing.T) {
			t.Parallel()

			esperas := &esperasAnotadas{}
			servidor, contador := servidorConRobots(t, caso.robots, redireccionesDePrueba())
			cliente := clienteDePrueba(t, ConIntervalo(time.Millisecond), conReloj(esperas.reloj))

			for consulta := range 2 {
				respuesta, err := cliente.Pedir(t.Context(), schema.Contexto{},
					Peticion{Metodo: http.MethodGet, URL: servidor.URL + "/norma?n=" + strconv.Itoa(consulta)})

				if caso.permitida {
					require.NoError(t, err, "un robots.txt sin reglas no impide pedir nada (FR-015)")
					require.Equal(t, http.StatusOK, respuesta.Estado)

					continue
				}

				fallo := falloDe(t, err)
				assert.Equal(t, schema.ClaseLimiteOTos, fallo.Clase(),
					"un robots.txt que no se obtiene ni se interpreta deniega con «límite o términos de uso» (FR-015)")
				assert.Equal(t, caso.estado, fallo.Estado, "y conserva el estado con que la fuente respondió")
				assert.Equal(t, caso.espera, fallo.Espera, "con la espera que la fuente declaró, si la declaró (FR-030)")
				assert.Equal(t, caso.espera != 0, fallo.EsperaConocida)
				assert.Equal(t, Respuesta{}, respuesta, "sin permiso no hay respuesta que entregar")
			}

			assert.Equal(t, caso.obtenciones, contador.robots.Load(),
				"el resultado de la obtención se cachea: la segunda ruta no vuelve a pedir el robots.txt (FR-015)")

			if caso.permitida {
				assert.Equal(t, int64(2), contador.total.Load(), "las dos rutas se piden")
			} else {
				assert.Zero(t, contador.total.Load(), "sin permiso no se emite ninguna petición del recurso (FR-014)")
			}
		})
	}

	t.Run("el vencimiento del contexto termina la operación sin evaluar ninguna regla", func(t *testing.T) {
		t.Parallel()

		const plazo = 50 * time.Millisecond

		var obtenciones atomic.Int64

		servidor, contador := servidorConRobots(t, func(escritor http.ResponseWriter, peticion *http.Request) {
			// La primera vez el robots.txt no llega nunca: quien termina la
			// operación es el contexto de quien la pidió. Las siguientes el sitio
			// sí responde, que es lo que deja ver si el cliente vuelve a preguntar
			// o dio el sitio por denegado.
			if obtenciones.Add(1) == 1 {
				<-peticion.Context().Done()

				return
			}

			robotsQueDice("User-agent: *\nAllow: /\n")(escritor, peticion)
		}, redireccionesDePrueba())

		cliente := clienteDePrueba(t, ConIntervalo(time.Millisecond))
		norma := Peticion{Metodo: http.MethodGet, URL: servidor.URL + "/norma"}

		ctx, cancelar := context.WithTimeout(t.Context(), plazo)
		defer cancelar()

		comienzo := time.Now()
		respuesta, err := cliente.Pedir(ctx, schema.Contexto{}, norma)

		fallo := falloDe(t, err)
		assert.Equal(t, schema.ClaseFuenteNoDisponible, fallo.Clase(),
			"el contexto que vence obteniendo el robots.txt es clase 4 y no denegación (FR-015, FR-029)")
		require.ErrorIs(t, err, context.DeadlineExceeded, "la causa del contexto llega intacta hasta quien llama")
		assert.Equal(t, Respuesta{}, respuesta)
		assert.Zero(t, contador.total.Load(), "no se llega a pedir el recurso")
		assert.Less(t, time.Since(comienzo), time.Second, "la operación termina con el contexto, no más tarde")

		// Y es el único caso que no deja nada en la caché: con el contexto vivo,
		// la petición siguiente vuelve a pedir el robots.txt en vez de dar el
		// sitio por denegado mientras viva el cliente (contrato de errores §3,
		// data-model.md §5).
		respuesta, err = cliente.Pedir(t.Context(), schema.Contexto{}, norma)
		require.NoError(t, err, "un plazo agotado obteniendo el robots.txt no deja el sitio denegado (FR-015)")
		assert.Equal(t, http.StatusOK, respuesta.Estado)
		assert.Equal(t, int64(2), contador.robots.Load(),
			"el robots.txt se vuelve a pedir: el vencimiento del contexto no se cachea")
		assert.Equal(t, int64(1), contador.total.Load(), "y esta vez el recurso sí se pide")
	})
}

// TestRobotsRedirigido cubre el escenario 8 de US1: el robots.txt detrás de una
// redirección. La obtención sigue su propia cadena con el mismo tope que el
// recurso, cada salto identificado, y la respuesta final se clasifica como
// cualquier otra. Una cadena que excede el tope no es «la fuente no sabe entregar
// el recurso» sino «el permiso no se pudo obtener»: deniega con clase 5 y no con
// la clase 4 de FR-011 (FR-015, FR-016).
func TestRobotsRedirigido(t *testing.T) {
	t.Parallel()

	t.Run("la cadena se sigue y las reglas del destino se aplican", func(t *testing.T) {
		t.Parallel()

		var saltos, sinIdentificar atomic.Int64

		servidor := servidorLocal(t, func(escritor http.ResponseWriter, peticion *http.Request) {
			if peticion.Header.Get("User-Agent") != AgenteDeUsuario() {
				sinIdentificar.Add(1)
			}

			switch peticion.URL.Path {
			case rutaDelRobots:
				saltos.Add(1)
				http.Redirect(escritor, peticion, "/reglas/nuevas", http.StatusMovedPermanently)
			case "/reglas/nuevas":
				saltos.Add(1)
				http.Redirect(escritor, peticion, "/reglas/definitivas", http.StatusFound)
			case "/reglas/definitivas":
				saltos.Add(1)
				robotsQueDice("User-agent: *\nDisallow: /privado/\n")(escritor, peticion)
			default:
				redireccionesDePrueba()(escritor, peticion)
			}
		})

		cliente := clienteDePrueba(t, ConIntervalo(time.Millisecond))

		respuesta, err := cliente.Pedir(t.Context(), schema.Contexto{},
			Peticion{Metodo: http.MethodGet, URL: servidor.URL + "/norma"})
		require.NoError(t, err, "las reglas obtenidas tras la cadena permiten esta ruta")
		assert.Equal(t, http.StatusOK, respuesta.Estado)

		_, err = cliente.Pedir(t.Context(), schema.Contexto{},
			Peticion{Metodo: http.MethodGet, URL: servidor.URL + "/privado/norma"})

		fallo := falloDe(t, err)
		assert.Equal(t, schema.ClaseLimiteOTos, fallo.Clase(),
			"y las mismas reglas desautorizan esta otra (FR-015)")

		assert.Equal(t, int64(3), saltos.Load(),
			"la cadena del robots.txt se sigue entera y una sola vez, aunque se pidan dos rutas (SC-003)")
		assert.Zero(t, sinIdentificar.Load(), "cada salto de la cadena del robots.txt va identificado (FR-009)")
	})

	t.Run("una cadena que excede el tope deniega, no declara la fuente caída", func(t *testing.T) {
		t.Parallel()

		var saltos atomic.Int64

		servidor := servidorLocal(t, func(escritor http.ResponseWriter, peticion *http.Request) {
			http.Redirect(escritor, peticion, "/reglas/"+strconv.FormatInt(saltos.Add(1), 10), http.StatusFound)
		})

		respuesta, err := clienteDePrueba(t, ConIntervalo(time.Millisecond)).Pedir(t.Context(), schema.Contexto{},
			Peticion{Metodo: http.MethodGet, URL: servidor.URL + "/norma"})

		fallo := falloDe(t, err)
		assert.Equal(t, schema.ClaseLimiteOTos, fallo.Clase(),
			"lo que falta no es el recurso sino el permiso para pedirlo (FR-015, no FR-011)")
		assert.Equal(t, Respuesta{}, respuesta)
		assert.Equal(t, int64(topeDeRedirecciones+1), saltos.Load(),
			"la petición del robots.txt más los diez saltos del tope, y ni una más (D10)")
	})

	t.Run("una cadena que vuelve sobre sí misma también deniega", func(t *testing.T) {
		t.Parallel()

		var saltos atomic.Int64

		servidor := servidorLocal(t, func(escritor http.ResponseWriter, peticion *http.Request) {
			saltos.Add(1)
			http.Redirect(escritor, peticion, rutaDelRobots, http.StatusFound)
		})

		_, err := clienteDePrueba(t, ConIntervalo(time.Millisecond)).Pedir(t.Context(), schema.Contexto{},
			Peticion{Metodo: http.MethodGet, URL: servidor.URL + "/norma"})

		fallo := falloDe(t, err)
		assert.Equal(t, schema.ClaseLimiteOTos, fallo.Clase(),
			"un robots.txt que se redirige a sí mismo tampoco se obtiene (FR-015)")
		assert.Equal(t, int64(1), saltos.Load(),
			"el bucle se detecta antes de emitir la petición repetida, sin agotar el tope")
	})
}

// TestRobotsNoSeEvaluaASiMismo comprueba FR-016 por sus dos mitades: la petición
// de robots.txt no se somete a sí misma —si lo hiciera, un sitio que lo
// desautoriza todo no podría ni obtener sus propias reglas, y la obtención
// entraría en recursión— y tampoco se somete ningún salto de su propia cadena de
// redirecciones, ni al robots.txt del origen ni al del destino.
func TestRobotsNoSeEvaluaASiMismo(t *testing.T) {
	t.Parallel()

	t.Run("un sitio que lo desautoriza todo entrega igual su robots.txt", func(t *testing.T) {
		t.Parallel()

		servidor, contador := servidorConRobots(t,
			robotsQueDice("User-agent: *\nDisallow: /\n"), redireccionesDePrueba())

		_, err := clienteDePrueba(t).Pedir(t.Context(), schema.Contexto{},
			Peticion{Metodo: http.MethodGet, URL: servidor.URL + "/norma"})

		fallo := falloDe(t, err)
		assert.Equal(t, schema.ClaseLimiteOTos, fallo.Clase(), "la ruta pedida sí queda desautorizada (FR-014)")
		assert.Equal(t, int64(1), contador.robots.Load(),
			"y el robots.txt se obtuvo una sola vez: no se evalúa contra sí mismo ni entra en recursión (FR-016)")
	})

	t.Run("el salto hacia otro sitio no consulta el robots.txt de ese sitio", func(t *testing.T) {
		t.Parallel()

		reglas, contadorDeLasReglas := servidorConRobots(t,
			robotsQueDice("User-agent: *\nDisallow: /\n"),
			robotsQueDice("User-agent: *\nAllow: /\n"))

		origen, contadorDelOrigen := servidorConRobots(t, func(escritor http.ResponseWriter, peticion *http.Request) {
			http.Redirect(escritor, peticion, reglas.URL+"/reglas", http.StatusFound)
		}, redireccionesDePrueba())

		respuesta, err := clienteDePrueba(t, ConIntervalo(time.Millisecond)).Pedir(t.Context(), schema.Contexto{},
			Peticion{Metodo: http.MethodGet, URL: origen.URL + "/norma"})
		require.NoError(t, err, "las reglas que la cadena entrega permiten la ruta")
		assert.Equal(t, http.StatusOK, respuesta.Estado)

		assert.Equal(t, int64(1), contadorDelOrigen.robots.Load(), "el robots.txt del origen se pidió una vez")
		assert.Zero(t, contadorDeLasReglas.robots.Load(),
			"y el del sitio de destino del salto no se pidió nunca: ningún salto de la cadena del "+
				"robots.txt se evalúa contra ningún robots.txt (FR-016)")
		assert.Equal(t, int64(1), contadorDeLasReglas.total.Load(), "solo se le pidió el salto de la cadena")
	})
}

// TestRobotsPorSitio comprueba que el ámbito de la caché es el sitio tal como lo
// define el spec —esquema, host y puerto—: dos servidores locales que solo se
// distinguen por el puerto tienen robots.txt distintos, cada uno con su propia
// obtención, y el permiso de uno no alcanza al otro (FR-013, FR-015, SC-004).
func TestRobotsPorSitio(t *testing.T) {
	t.Parallel()

	permisivo, contadorDelPermisivo := servidorConRobots(t,
		robotsQueDice("User-agent: *\nAllow: /\n"), redireccionesDePrueba())
	estricto, contadorDelEstricto := servidorConRobots(t,
		robotsQueDice("User-agent: *\nDisallow: /norma\n"), redireccionesDePrueba())

	require.Equal(t, direccionDePrueba(t, permisivo.URL).Hostname(), direccionDePrueba(t, estricto.URL).Hostname(),
		"los dos servidores del test están en el mismo host")
	require.NotEqual(t, claveDeSitio(direccionDePrueba(t, permisivo.URL)),
		claveDeSitio(direccionDePrueba(t, estricto.URL)),
		"y aun así son dos sitios, porque el puerto entra en la clave")

	cliente := clienteDePrueba(t, ConIntervalo(time.Millisecond))

	respuesta, err := cliente.Pedir(t.Context(), schema.Contexto{},
		Peticion{Metodo: http.MethodGet, URL: permisivo.URL + "/norma"})
	require.NoError(t, err)
	assert.Equal(t, http.StatusOK, respuesta.Estado)

	_, err = cliente.Pedir(t.Context(), schema.Contexto{},
		Peticion{Metodo: http.MethodGet, URL: estricto.URL + "/norma"})

	fallo := falloDe(t, err)
	assert.Equal(t, schema.ClaseLimiteOTos, fallo.Clase(),
		"el otro sitio tiene sus propias reglas y no hereda el permiso del primero (FR-013)")

	assert.Equal(t, int64(1), contadorDelPermisivo.robots.Load())
	assert.Equal(t, int64(1), contadorDelEstricto.robots.Load(), "cada sitio tiene su propia obtención")
	assert.Equal(t, int64(1), contadorDelPermisivo.total.Load())
	assert.Zero(t, contadorDelEstricto.total.Load(), "la ruta desautorizada del otro sitio no se pide")
}

// servidorConRobots levanta un servidor local que responde al robots.txt con el
// manejador que se le declara y atiende el resto de rutas con el otro, delante
// del contador que usan todas las tablas: así cada una sabe cuántas veces le
// pidieron el permiso y cuántas el recurso, y que ninguna de las dos llegó sin la
// identificación del proyecto.
func servidorConRobots(t *testing.T, robots, manejador http.HandlerFunc) (*httptest.Server, *contadorDeIdentificacion) {
	t.Helper()

	contador := &contadorDeIdentificacion{}

	return servidorLocal(t, contador.vigila(func(escritor http.ResponseWriter, peticion *http.Request) {
		if peticion.URL.Path == rutaDelRobots {
			robots(escritor, peticion)

			return
		}

		manejador(escritor, peticion)
	})), contador
}

// robotsQueDice entrega el robots.txt que se le declara, con el tipo de contenido
// que RFC 9309 §2.3 espera.
func robotsQueDice(contenido string) http.HandlerFunc {
	return func(escritor http.ResponseWriter, _ *http.Request) {
		escritor.Header().Set("Content-Type", "text/plain; charset=utf-8")
		_, _ = io.WriteString(escritor, contenido)
	}
}

// robotsConEstado responde a la petición del robots.txt con el estado que se le
// declara y sin cuerpo, que es como una fuente declara que no tiene reglas, que
// está caída o que no sabe de qué se le habla.
func robotsConEstado(estado int) http.HandlerFunc {
	return func(escritor http.ResponseWriter, _ *http.Request) {
		escritor.WriteHeader(estado)
	}
}

// robotsLimitado es la fuente que responde al robots.txt con su límite de
// peticiones alcanzado y dice cuándo se puede volver a pedir.
func robotsLimitado(cuando string) http.HandlerFunc {
	return func(escritor http.ResponseWriter, _ *http.Request) {
		escritor.Header().Set("Retry-After", cuando)
		escritor.WriteHeader(http.StatusTooManyRequests)
	}
}
