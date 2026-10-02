package httpx

import (
	"context"
	"errors"
	"fmt"
	"math"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"sync/atomic"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/jmorenobl/kitlegal/internal/cli"
	"github.com/jmorenobl/kitlegal/internal/core/schema"
)

// peticionDePrueba es la petición implicada en los fallos de estas tablas. Como
// toda dirección del paquete, es de un host ficticio: ningún test de H2 conoce
// una dirección externa (plan.md, obligación 11).
var peticionDePrueba = Peticion{Metodo: "GET", URL: "http://fuente.prueba/norma"}

// instanteDePrueba es el «ahora» contra el que se mide la espera que declara una
// fecha HTTP en Retry-After. Va fijo para que la tabla no dependa del reloj de
// la máquina (D20).
var instanteDePrueba = time.Date(2026, time.September, 12, 10, 0, 0, 0, time.UTC)

// TestErrorEnvueltoConservaLaClase comprueba la mitad del paquete en el mecanismo
// de D4: el error declara su clase y el kernel la reconoce con errors.As sin que
// internal/httpx importe internal/cli —es el test, que sí puede, quien llama a
// cli.Clasificar (contracts/errores-y-ensayo.md §4)—, y envolverlo con
// fmt.Errorf("%w") no cambia lo que el kernel devuelve, por muchas capas de
// contexto que se le añadan (FR-031, FR-063).
//
// Comprueba además las dos direcciones del envoltorio: hacia fuera, que
// errors.As recupera el *Error con sus datos intactos —que es como un adaptador
// lee Retry-After sin analizar ninguna cadena de texto (contrato §2)—; y hacia
// dentro, que Unwrap expone la causa, de modo que errors.Is la alcance a través
// del error del paquete.
func TestErrorEnvueltoConservaLaClase(t *testing.T) {
	t.Parallel()

	casos := []struct {
		nombre string
		fallo  *Error
		clase  schema.Clase
	}{
		{
			nombre: "argumentos: el método no está permitido",
			fallo:  errorDeArgumentos(peticionDePrueba, nil, "el método debe ser GET o HEAD"),
			clase:  schema.ClaseArgumentos,
		},
		{
			nombre: "fuente no disponible: 5xx agotados los intentos",
			fallo: errorDeFuenteNoDisponible(peticionDePrueba, http.StatusServiceUnavailable, nil,
				"el sitio no ha respondido correctamente tras agotar los intentos"),
			clase: schema.ClaseFuenteNoDisponible,
		},
		{
			nombre: "límite o términos de uso: el robots.txt deniega la ruta",
			fallo:  errorDeLimiteOTos(peticionDePrueba, 0, nil, "el robots.txt del sitio no permite la ruta"),
			clase:  schema.ClaseLimiteOTos,
		},
		{
			nombre: "límite o términos de uso: 429 con espera declarada",
			fallo: errorDe429(peticionDePrueba, Cabeceras{"Retry-After": {"120"}}, instanteDePrueba,
				"la fuente ha alcanzado su límite de peticiones"),
			clase: schema.ClaseLimiteOTos,
		},
		{
			nombre: "inesperado: el mecanismo de grabación",
			fallo:  errorInesperado(peticionDePrueba, nil, "no se ha podido escribir la grabación"),
			clase:  schema.ClaseInesperado,
		},
	}

	for _, caso := range casos {
		t.Run(caso.nombre, func(t *testing.T) {
			t.Parallel()

			// La clase que el error declara es del vocabulario del dominio: una
			// de fuera saldría como «inesperado» y el sobre no podría llevarla
			// (FR-063).
			assert.Contains(t, schema.Clases(), caso.fallo.Clase())
			assert.Equal(t, caso.clase, caso.fallo.Clase())
			assert.Equal(t, caso.clase, cli.Clasificar(caso.fallo))

			unaCapa := fmt.Errorf("consultando la norma: %w", caso.fallo)
			assert.Equal(t, caso.clase, cli.Clasificar(unaCapa))

			dosCapas := fmt.Errorf("el applet no ha podido terminar: %w", unaCapa)
			assert.Equal(t, caso.clase, cli.Clasificar(dosCapas))

			// Y el error sigue siendo alcanzable con sus datos bajo las dos
			// capas, que es el uso que documenta el contrato §2.
			var recuperado *Error
			require.ErrorAs(t, dosCapas, &recuperado)
			assert.Same(t, caso.fallo, recuperado)
		})
	}

	t.Run("la clase no es nunca «no encontrado» ni «identidad humana»", func(t *testing.T) {
		t.Parallel()

		for _, caso := range casos {
			assert.NotEqual(t, schema.ClaseNoEncontrado, caso.fallo.Clase(), caso.nombre)
			assert.NotEqual(t, schema.ClaseIdentidadHumana, caso.fallo.Clase(), caso.nombre)
		}
	})

	t.Run("Unwrap expone la causa y errors.Is la alcanza", func(t *testing.T) {
		t.Parallel()

		conCausa := errorDeFuenteNoDisponible(peticionDePrueba, 0, context.DeadlineExceeded,
			"la operación no ha terminado dentro del plazo")

		assert.Equal(t, context.DeadlineExceeded, conCausa.Unwrap())
		require.ErrorIs(t, conCausa, context.DeadlineExceeded)
		require.ErrorIs(t, fmt.Errorf("consultando la norma: %w", conCausa), context.DeadlineExceeded)
		// Y la causa no cambia la clase: la declara el error, no lo que envuelve.
		assert.Equal(t, schema.ClaseFuenteNoDisponible, cli.Clasificar(conCausa))
	})

	t.Run("sin causa, Unwrap no devuelve nada", func(t *testing.T) {
		t.Parallel()

		sinCausa := errorDeArgumentos(Peticion{}, nil, "falta la opción ConFuente")

		require.NoError(t, sinCausa.Unwrap())
		require.NotErrorIs(t, sinCausa, context.DeadlineExceeded)
	})
}

// TestErrorRetryAfter fija la lectura de Retry-After de D20: un entero son
// segundos; cualquier otro valor se intenta como fecha HTTP y lo que queda hasta
// ella son los segundos, con mínimo cero; EsperaConocida dice si la cabecera
// estaba y se pudo leer, y un valor ilegible equivale a su ausencia. Nada de
// esto cambia la clase —sigue siendo «límite o términos de uso» (FR-030)— ni
// hace esperar al cliente, que solo transporta el dato para que quien llama
// decida (US2 escenario 3).
func TestErrorRetryAfter(t *testing.T) {
	t.Parallel()

	// La mayor espera que cabe en una duración: el tope con el que un valor
	// enorme no se desborda en silencio.
	maxima := time.Duration(math.MaxInt64)

	casos := []struct {
		nombre    string
		cabeceras Cabeceras
		espera    time.Duration
		conocida  bool
	}{
		{
			nombre:    "segundos",
			cabeceras: Cabeceras{"Retry-After": {"120"}},
			espera:    2 * time.Minute,
			conocida:  true,
		},
		{
			nombre:    "cero segundos: la cabecera estaba y dice que ya",
			cabeceras: Cabeceras{"Retry-After": {"0"}},
			espera:    0,
			conocida:  true,
		},
		{
			nombre:    "segundos con espacios alrededor",
			cabeceras: Cabeceras{"Retry-After": {" 120 "}},
			espera:    2 * time.Minute,
			conocida:  true,
		},
		{
			nombre:    "segundos negativos: mínimo cero",
			cabeceras: Cabeceras{"Retry-After": {"-5"}},
			espera:    0,
			conocida:  true,
		},
		{
			nombre:    "segundos que desbordarían la duración: el tope",
			cabeceras: Cabeceras{"Retry-After": {"9223372036854775807"}},
			espera:    maxima,
			conocida:  true,
		},
		{
			nombre:    "fecha HTTP futura: lo que queda hasta ella",
			cabeceras: Cabeceras{"Retry-After": {instanteDePrueba.Add(90 * time.Second).Format(http.TimeFormat)}},
			espera:    90 * time.Second,
			conocida:  true,
		},
		{
			nombre:    "fecha HTTP pasada: mínimo cero",
			cabeceras: Cabeceras{"Retry-After": {instanteDePrueba.Add(-time.Hour).Format(http.TimeFormat)}},
			espera:    0,
			conocida:  true,
		},
		{
			nombre:    "fecha HTTP en otro de los formatos de HTTP/1.1",
			cabeceras: Cabeceras{"Retry-After": {instanteDePrueba.Add(time.Minute).UTC().Format(time.ANSIC)}},
			espera:    time.Minute,
			conocida:  true,
		},
		{
			nombre:    "sin la cabecera: no se conoce, y la clase no cambia",
			cabeceras: Cabeceras{"Date": {instanteDePrueba.Format(http.TimeFormat)}},
			espera:    0,
			conocida:  false,
		},
		{
			nombre:    "sin ninguna cabecera",
			cabeceras: nil,
			espera:    0,
			conocida:  false,
		},
		{
			nombre:    "cabecera presente y vacía",
			cabeceras: Cabeceras{"Retry-After": {""}},
			espera:    0,
			conocida:  false,
		},
		{
			nombre:    "valor ilegible: equivale a su ausencia",
			cabeceras: Cabeceras{"Retry-After": {"mañana"}},
			espera:    0,
			conocida:  false,
		},
		{
			nombre:    "entero fuera del rango de un int64: ilegible",
			cabeceras: Cabeceras{"Retry-After": {"99999999999999999999"}},
			espera:    0,
			conocida:  false,
		},
	}

	for _, caso := range casos {
		t.Run(caso.nombre, func(t *testing.T) {
			t.Parallel()

			fallo := errorDe429(peticionDePrueba, caso.cabeceras, instanteDePrueba,
				"la fuente ha alcanzado su límite de peticiones")

			assert.Equal(t, caso.espera, fallo.Espera)
			assert.Equal(t, caso.conocida, fallo.EsperaConocida)
			// El 429 es el estado del error, lo diga la cabecera o no, y la
			// clase es la misma en los trece casos (FR-030).
			assert.Equal(t, http.StatusTooManyRequests, fallo.Estado)
			assert.Equal(t, schema.ClaseLimiteOTos, fallo.Clase())
			assert.Equal(t, schema.ClaseLimiteOTos, cli.Clasificar(fallo))

			// Un adaptador lee la espera del error envuelto, sin analizar
			// ninguna cadena de texto (contrato §2, US2 escenario 3).
			var recuperado *Error
			require.ErrorAs(t, fmt.Errorf("consultando la norma: %w", fallo), &recuperado)
			assert.Equal(t, caso.espera, recuperado.Espera)
			assert.Equal(t, caso.conocida, recuperado.EsperaConocida)
		})
	}

	t.Run("la espera declarada aparece en el mensaje, y la desconocida no", func(t *testing.T) {
		t.Parallel()

		conEspera := errorDe429(peticionDePrueba, Cabeceras{"Retry-After": {"120"}}, instanteDePrueba,
			"la fuente ha alcanzado su límite de peticiones")
		sinEspera := errorDe429(peticionDePrueba, nil, instanteDePrueba,
			"la fuente ha alcanzado su límite de peticiones")

		assert.Contains(t, conEspera.Error(), "dentro de 2m0s")
		assert.NotContains(t, sinEspera.Error(), "dentro de")
	})
}

// TestErrorMensajesEnEspanol comprueba el mensaje que la persona lee, que es lo
// único que el sobre de fallo lleva junto a la clase: está en español y nombra
// el sitio y la ruta de la petición implicada; cuando no hay ninguna —los fallos
// de configuración— nombra en su lugar la opción que falta o que sobra (FR-034).
// El detalle técnico no entra en él: la causa viaja por Unwrap y el registro de
// eventos, no por el mensaje, de modo que un error de la biblioteca estándar no
// cuele una línea en inglés (FR-033, FR-063, D4, D20).
func TestErrorMensajesEnEspanol(t *testing.T) {
	t.Parallel()

	casos := []struct {
		nombre  string
		fallo   *Error
		mensaje string
	}{
		{
			nombre: "con petición: nombra método, ruta con su consulta y sitio, y el estado",
			fallo: errorDeFuenteNoDisponible(
				Peticion{Metodo: "GET", URL: "http://fuente.prueba/norma?id=BOE-A-2015-10565"},
				http.StatusServiceUnavailable, nil,
				"el sitio no ha respondido correctamente tras agotar los intentos"),
			mensaje: "GET /norma?id=BOE-A-2015-10565 en el sitio http://fuente.prueba:80: " +
				"el sitio no ha respondido correctamente tras agotar los intentos (estado 503)",
		},
		{
			nombre: "con petición y espera: el estado y lo que la fuente indica",
			fallo: errorDe429(Peticion{Metodo: "GET", URL: "http://fuente.prueba/limitada"},
				Cabeceras{"Retry-After": {"120"}}, instanteDePrueba,
				"la fuente ha alcanzado su límite de peticiones"),
			mensaje: "GET /limitada en el sitio http://fuente.prueba:80: " +
				"la fuente ha alcanzado su límite de peticiones (estado 429); " +
				"la fuente indica que se puede volver a pedir dentro de 2m0s",
		},
		{
			nombre: "sin estado: el del robots.txt que deniega, con el puerto del esquema",
			fallo: errorDeLimiteOTos(Peticion{Metodo: "HEAD", URL: "https://otra.prueba/privada"}, 0, nil,
				"el robots.txt del sitio no permite la ruta"),
			mensaje: "HEAD /privada en el sitio https://otra.prueba:443: " +
				"el robots.txt del sitio no permite la ruta",
		},
		{
			nombre: "dirección sin ruta: la ruta pedida es la raíz",
			fallo: errorDeLimiteOTos(Peticion{Metodo: "GET", URL: "http://fuente.prueba"}, 0, nil,
				"el robots.txt del sitio no permite la ruta"),
			mensaje: "GET / en el sitio http://fuente.prueba:80: el robots.txt del sitio no permite la ruta",
		},
		{
			nombre: "sin petición: nombra la opción que falta",
			fallo: errorDeArgumentos(Peticion{}, nil,
				"la grabación está activa pero falta la opción ConFuente"),
			mensaje: "la grabación está activa pero falta la opción ConFuente",
		},
		{
			nombre: "sin petición: nombra la opción que sobra",
			fallo: errorDeArgumentos(Peticion{}, nil,
				"sobra la opción ConIntervalo: en reproducción no hay ritmo que respetar"),
			mensaje: "sobra la opción ConIntervalo: en reproducción no hay ritmo que respetar",
		},
		{
			nombre:  "sin petición: el ritmo nulo nombra su opción",
			fallo:   falloDeLaOpcion(t, ConRitmo(nil)),
			mensaje: "el ritmo en el que el cliente espera turno no puede ser nulo (ConRitmo)",
		},
		{
			nombre: "sin petición: el ritmo sin intervalo nombra su opción y el intervalo que trae",
			fallo:  falloDeLaOpcion(t, ConRitmo(NuevoRitmo(-time.Second))),
			mensaje: "el intervalo del ritmo entre peticiones a un mismo sitio tiene que ser mayor que cero " +
				"(ConRitmo): -1s",
		},
		{
			nombre: "sin petición: el ritmo que sobra en reproducción nombra su opción",
			fallo: falloDe(t, comprobarOpcionesDeReproduccion(
				configuracionDelCliente{ritmo: NuevoRitmo(time.Second)})),
			mensaje: "ConRitmo no tiene sentido en reproducción: la reproducción no espera nunca (FR-048)",
		},
		{
			nombre:  "petición sin dirección: no hay sitio ni ruta que nombrar",
			fallo:   errorDeArgumentos(Peticion{Metodo: "GET"}, nil, "la dirección no puede estar vacía"),
			mensaje: "la dirección no puede estar vacía",
		},
		{
			nombre: "dirección inanalizable: se nombra tal como la escribió quien llama",
			fallo: errorDeArgumentos(Peticion{Metodo: "GET", URL: "http://[::1"},
				errors.New("parse error"), "la dirección no se puede interpretar"),
			mensaje: "GET http://[::1: la dirección no se puede interpretar",
		},
		{
			nombre: "dirección relativa: tampoco hay sitio que derivar",
			fallo: errorDeArgumentos(Peticion{Metodo: "GET", URL: "/norma"}, nil,
				"la dirección debe ser absoluta"),
			mensaje: "GET /norma: la dirección debe ser absoluta",
		},
		{
			nombre: "esquema que no es de red: no se le inventa ningún puerto",
			fallo: errorDeArgumentos(Peticion{Metodo: "GET", URL: "ftp://fuente.prueba/norma"}, nil,
				"el esquema debe ser http o https"),
			mensaje: "GET ftp://fuente.prueba/norma: el esquema debe ser http o https",
		},
		{
			nombre: "el mecanismo de reproducción nombra la petición y el fichero",
			fallo: errorInesperado(peticionDePrueba, nil,
				`no hay grabación de la petición en "GET_http_fuente.prueba_norma.json"`),
			mensaje: "GET /norma en el sitio http://fuente.prueba:80: " +
				`no hay grabación de la petición en "GET_http_fuente.prueba_norma.json"`,
		},
	}

	for _, caso := range casos {
		t.Run(caso.nombre, func(t *testing.T) {
			t.Parallel()

			assert.Equal(t, caso.mensaje, caso.fallo.Error())
		})
	}

	t.Run("la causa no entra en el mensaje, ni siquiera la de la biblioteca estándar", func(t *testing.T) {
		t.Parallel()

		fallo := errorDeFuenteNoDisponible(peticionDePrueba, 0, context.DeadlineExceeded,
			"la operación no ha terminado dentro del plazo")

		assert.Equal(t, "GET /norma en el sitio http://fuente.prueba:80: "+
			"la operación no ha terminado dentro del plazo", fallo.Error())
		assert.NotContains(t, fallo.Error(), context.DeadlineExceeded.Error())
	})

	t.Run("ninguna ruta termina en panic ni en un mensaje vacío", func(t *testing.T) {
		t.Parallel()

		// Error es un tipo exportado: nada impide que alguien de fuera lo
		// construya a cero o que una función devuelva un *Error nulo como
		// error. Ninguno de los dos casos lo produce este paquete, y ninguno
		// puede acabar en panic ni en un sobre con el mensaje vacío, que el
		// esquema de --describe prohíbe (FR-033, schema.DatosError).
		aCero := &Error{}

		var nulo *Error

		require.NotPanics(t, func() {
			assert.NotEmpty(t, aCero.Error())
			assert.Equal(t, schema.ClaseInesperado, aCero.Clase())
			assert.NoError(t, aCero.Unwrap())

			assert.NotEmpty(t, nulo.Error())
			assert.Equal(t, schema.ClaseInesperado, nulo.Clase())
			assert.NoError(t, nulo.Unwrap())
		})

		assert.Equal(t, schema.ClaseInesperado, cli.Clasificar(aCero))
	})
}

// TestClasesDeError es el control de que **ninguna ruta de fallo del paquete
// queda sin clase**, y lo comprueba donde de verdad importa: sobre el error tal
// como sale del cliente, no sobre uno construido a mano para la ocasión. Las
// tres tablas anteriores de este fichero miden el tipo Error; esta mide el
// paquete entero, provocando cada situación con el servidor local, un directorio
// temporal o un cliente de reproducción.
//
// Es la tabla **cerrada** del contrato de errores §3, fila por fila: las nueve
// situaciones de red de SC-006 —5xx del recurso agotado, contexto vencido en
// cualquier punto, cadena del recurso en bucle o excedida, 429 del recurso,
// robots.txt que deniega, 5xx o transporte del robots.txt agotados, 429 del
// robots.txt, cadena del robots.txt excedida y método no permitido— y las siete
// restantes de SC-016 y de research D19: dirección inválida, configuración de
// grabación incompleta o con un valor inválido, raíz o directorio inválidos,
// grabación y reproducción a la vez u opción sin sentido, entrada y salida
// sobrevenida al grabar, colisión de nombres y grabación ausente o ajena.
//
// De cada una se comprueban las dos mitades del mecanismo de D4: la clase que
// cli.Clasificar devuelve —es el test quien importa internal/cli, nunca el
// paquete (contrato §4)— y el código que cli.CodigoSalida produce, sobre el
// error tal cual y también envuelto, que es como llega desde un applet (FR-031).
// Y de las dieciséis juntas, las dos reglas que cierran la tabla: el cliente no
// produce nunca «no encontrado» (3) —quien sabe si lo pedido existe es la
// fuente— ni «requiere identidad humana» (6) —este paquete no hace nada que la
// exija—, y ninguna termina en panic (FR-029 a FR-033, FR-063, SC-006, SC-016).
//
// Esta tabla no declara t.Parallel(), y no es un descuido: usa t.Setenv, que
// «cannot be used in parallel tests» (go doc testing.T.Setenv), y partirla en
// dos rompería justamente lo que la hace un control: que las dieciséis
// situaciones se lean juntas y en el orden del contrato. Vale aquí lo que
// grabar_test.go documenta para sus tablas.
func TestClasesDeError(t *testing.T) {
	// La grabación va apagada para toda la tabla, y se declara en vez de darse
	// por supuesta: así ninguna fila depende de con qué variable se lanzara la
	// suite, y las cinco que la necesitan encendida la declaran ellas.
	t.Setenv(VariableGrabacion, "")

	casos := []struct {
		fila   int
		nombre string
		clase  schema.Clase
		codigo int
		// provocar devuelve un error por cada situación de la fila: las filas
		// que el contrato enuncia con un «o» —una cadena en bucle o excedida,
		// una grabación ausente o ajena— cubren todas las suyas, y todas tienen
		// que salir con la misma clase y el mismo código.
		provocar func(t *testing.T) []error
	}{
		{
			fila:   1,
			nombre: "5xx del recurso pedido agotados los intentos",
			clase:  schema.ClaseFuenteNoDisponible,
			codigo: 4,
			provocar: func(t *testing.T) []error {
				t.Helper()

				servidor, _ := servidorIdentificado(t, func(escritor http.ResponseWriter, _ *http.Request) {
					escritor.WriteHeader(http.StatusServiceUnavailable)
				})

				_, agotados := clienteSinEsperas(t).Pedir(t.Context(), schema.Contexto{},
					Peticion{Metodo: http.MethodGet, URL: servidor.URL + "/norma"})

				return []error{agotados}
			},
		},
		{
			fila:   2,
			nombre: "vencimiento del contexto en cualquier punto de la operación",
			clase:  schema.ClaseFuenteNoDisponible,
			codigo: 4,
			provocar: func(t *testing.T) []error {
				t.Helper()

				// Pidiendo el recurso, con el permiso del sitio ya concedido…
				delRecurso, _ := servidorIdentificado(t, esperaAlContexto())

				// …y obteniendo el robots.txt, que es el otro punto que la fila
				// nombra: los dos son «la operación no llegó a la fuente» y no
				// una denegación (FR-015, FR-029).
				delRobots, _ := servidorConRobots(t, esperaAlContexto(), redireccionesDePrueba())

				return []error{
					pedirConPlazo(t, delRecurso.URL+"/lenta"),
					pedirConPlazo(t, delRobots.URL+"/norma"),
				}
			},
		},
		{
			fila:   3,
			nombre: "cadena de redirecciones del recurso en bucle o por encima del tope",
			clase:  schema.ClaseFuenteNoDisponible,
			codigo: 4,
			provocar: func(t *testing.T) []error {
				t.Helper()

				enBucle, _ := servidorIdentificado(t, func(escritor http.ResponseWriter, peticion *http.Request) {
					http.Redirect(escritor, peticion, "/bucle", http.StatusFound)
				})

				sinFin, _ := servidorIdentificado(t, cadenaSinFin("/salto/"))

				cliente := clienteSinEsperas(t)

				_, repetida := cliente.Pedir(t.Context(), schema.Contexto{},
					Peticion{Metodo: http.MethodGet, URL: enBucle.URL + "/bucle"})

				_, excedida := cliente.Pedir(t.Context(), schema.Contexto{},
					Peticion{Metodo: http.MethodGet, URL: sinFin.URL + "/salto/0"})

				return []error{repetida, excedida}
			},
		},
		{
			fila:   4,
			nombre: "429 del recurso pedido",
			clase:  schema.ClaseLimiteOTos,
			codigo: 5,
			provocar: func(t *testing.T) []error {
				t.Helper()

				servidor, _ := servidorIdentificado(t, func(escritor http.ResponseWriter, _ *http.Request) {
					escritor.Header().Set("Retry-After", "120")
					escritor.WriteHeader(http.StatusTooManyRequests)
				})

				_, limitado := clienteSinEsperas(t).Pedir(t.Context(), schema.Contexto{},
					Peticion{Metodo: http.MethodGet, URL: servidor.URL + "/limitada"})

				return []error{limitado}
			},
		},
		{
			fila:   5,
			nombre: "el robots.txt del sitio deniega la ruta",
			clase:  schema.ClaseLimiteOTos,
			codigo: 5,
			provocar: func(t *testing.T) []error {
				t.Helper()

				servidor, _ := servidorConRobots(t,
					robotsQueDice("User-agent: *\nDisallow: /privado/\n"), redireccionesDePrueba())

				_, denegada := clienteSinEsperas(t).Pedir(t.Context(), schema.Contexto{},
					Peticion{Metodo: http.MethodGet, URL: servidor.URL + "/privado/norma"})

				return []error{denegada}
			},
		},
		{
			fila:   6,
			nombre: "5xx o fallo de transporte del robots.txt agotados los intentos",
			clase:  schema.ClaseLimiteOTos,
			codigo: 5,
			provocar: func(t *testing.T) []error {
				t.Helper()

				caido, _ := servidorConRobots(t,
					robotsConEstado(http.StatusServiceUnavailable), redireccionesDePrueba())

				// Un sitio que ya no escucha: la obtención del robots.txt no
				// llega a respuesta, y lo que falta entonces no es el recurso
				// sino el permiso para pedirlo (FR-015, no FR-029).
				apagado, _ := servidorIdentificado(t, redireccionesDePrueba())
				direccionDelApagado := apagado.URL + "/norma"

				apagado.Close()

				cliente := clienteSinEsperas(t)

				_, del5xx := cliente.Pedir(t.Context(), schema.Contexto{},
					Peticion{Metodo: http.MethodGet, URL: caido.URL + "/norma"})

				_, deTransporte := cliente.Pedir(t.Context(), schema.Contexto{},
					Peticion{Metodo: http.MethodGet, URL: direccionDelApagado})

				return []error{del5xx, deTransporte}
			},
		},
		{
			fila:   7,
			nombre: "429 del robots.txt",
			clase:  schema.ClaseLimiteOTos,
			codigo: 5,
			provocar: func(t *testing.T) []error {
				t.Helper()

				servidor, _ := servidorConRobots(t, robotsLimitado("120"), redireccionesDePrueba())

				_, limitado := clienteSinEsperas(t).Pedir(t.Context(), schema.Contexto{},
					Peticion{Metodo: http.MethodGet, URL: servidor.URL + "/norma"})

				return []error{limitado}
			},
		},
		{
			fila:   8,
			nombre: "cadena de redirecciones del robots.txt en bucle o por encima del tope",
			clase:  schema.ClaseLimiteOTos,
			codigo: 5,
			provocar: func(t *testing.T) []error {
				t.Helper()

				// Los dos servidores redirigen **toda** ruta, la del robots.txt
				// incluida, que es la primera que el cliente pide.
				sinFin := servidorLocal(t, cadenaSinFin("/reglas/"))

				enBucle := servidorLocal(t, func(escritor http.ResponseWriter, peticion *http.Request) {
					http.Redirect(escritor, peticion, rutaDelRobots, http.StatusFound)
				})

				cliente := clienteSinEsperas(t)

				_, excedida := cliente.Pedir(t.Context(), schema.Contexto{},
					Peticion{Metodo: http.MethodGet, URL: sinFin.URL + "/norma"})

				_, repetida := cliente.Pedir(t.Context(), schema.Contexto{},
					Peticion{Metodo: http.MethodGet, URL: enBucle.URL + "/norma"})

				return []error{excedida, repetida}
			},
		},
		{
			fila:   9,
			nombre: "método distinto de GET o HEAD",
			clase:  schema.ClaseArgumentos,
			codigo: 2,
			provocar: func(t *testing.T) []error {
				t.Helper()

				cliente := clienteSinEsperas(t)
				rechazados := make([]error, 0, 3)

				// Ninguna de las tres abre conexión, así que no hace falta
				// ningún servidor: el rechazo es anterior (FR-010).
				for _, metodo := range []string{http.MethodPost, http.MethodDelete, "get"} {
					_, rechazado := cliente.Pedir(t.Context(), schema.Contexto{},
						Peticion{Metodo: metodo, URL: peticionDePrueba.URL})

					rechazados = append(rechazados, rechazado)
				}

				return rechazados
			},
		},
		{
			fila:   10,
			nombre: "dirección no absoluta, de otro esquema o inanalizable",
			clase:  schema.ClaseArgumentos,
			codigo: 2,
			provocar: func(t *testing.T) []error {
				t.Helper()

				cliente := clienteSinEsperas(t)
				rechazadas := make([]error, 0, 4)

				for _, direccion := range []string{
					"", "/norma?id=1", "ftp://fuente.prueba/norma", "://fuente.prueba",
				} {
					_, rechazada := cliente.Pedir(t.Context(), schema.Contexto{},
						Peticion{Metodo: http.MethodGet, URL: direccion})

					rechazadas = append(rechazadas, rechazada)
				}

				return rechazadas
			},
		},
		{
			fila:   11,
			nombre: "grabación activa sin fuente o sin raíz, o con un valor de variable inválido",
			clase:  schema.ClaseArgumentos,
			codigo: 2,
			provocar: func(t *testing.T) []error {
				t.Helper()

				t.Setenv(VariableGrabacion, variableActiva)

				raiz := t.TempDir()

				_, sinFuente := New(ConRaizDeGrabacion(raiz))
				_, sinRaiz := New(ConFuente(fuenteDePrueba))

				t.Setenv(VariableGrabacion, "0")

				_, valorInvalido := New(ConFuente(fuenteDePrueba), ConRaizDeGrabacion(raiz))

				return []error{sinFuente, sinRaiz, valorInvalido}
			},
		},
		{
			fila:   12,
			nombre: "raíz de grabación o directorio de reproducción inválidos",
			clase:  schema.ClaseArgumentos,
			codigo: 2,
			provocar: func(t *testing.T) []error {
				t.Helper()

				t.Setenv(VariableGrabacion, variableActiva)

				noEsDirectorio := filepath.Join(t.TempDir(), "raiz.txt")
				require.NoError(t, os.WriteFile(noEsDirectorio, []byte("no soy un directorio"), permisoDeLaCopia))

				_, raizAusente := New(ConFuente(fuenteDePrueba),
					ConRaizDeGrabacion(filepath.Join(t.TempDir(), "ausente")))
				_, raizQueNoLoEs := New(ConFuente(fuenteDePrueba), ConRaizDeGrabacion(noEsDirectorio))

				// El directorio de Replay es la otra mitad de la fila, y la
				// reproducción exige la grabación apagada (FR-043).
				t.Setenv(VariableGrabacion, "")

				_, directorioAusente := Replay(filepath.Join(t.TempDir(), "tampoco"))

				return []error{raizAusente, raizQueNoLoEs, directorioAusente}
			},
		},
		{
			fila:   13,
			nombre: "grabación y reproducción a la vez, u opción sin sentido en la reproducción",
			clase:  schema.ClaseArgumentos,
			codigo: 2,
			provocar: func(t *testing.T) []error {
				t.Helper()

				directorio := grabacionesDePrueba(t)

				t.Setenv(VariableGrabacion, variableActiva)

				_, lasDosALaVez := Replay(directorio)

				t.Setenv(VariableGrabacion, "")

				_, conRaiz := Replay(directorio, ConRaizDeGrabacion(t.TempDir()))
				_, conRitmo := Replay(directorio, ConIntervalo(time.Second))
				_, conIntentos := Replay(directorio, ConIntentos(2))

				return []error{lasDosALaVez, conRaiz, conRitmo, conIntentos}
			},
		},
		{
			fila:   14,
			nombre: "entrada y salida sobrevenida al escribir la grabación",
			clase:  schema.ClaseInesperado,
			codigo: 1,
			provocar: func(t *testing.T) []error {
				t.Helper()

				t.Setenv(VariableGrabacion, variableActiva)

				raiz := t.TempDir()
				servidor, _ := servidorIdentificado(t, redireccionesDePrueba())
				cliente := clienteQueGraba(t, raiz)

				// La primera petición graba bien y deja decidido el robots.txt
				// del sitio, de modo que lo que falle luego sea la grabación del
				// recurso y no la obtención del permiso, que es la fila 6.
				primera, err := cliente.Pedir(t.Context(), schema.Contexto{},
					Peticion{Metodo: http.MethodGet, URL: servidor.URL + "/norma"})
				require.NoError(t, err, "la grabación de la primera petición sí se puede escribir")
				require.Equal(t, http.StatusOK, primera.Estado)

				// Y entonces el directorio que la construcción validó y creó
				// desaparece: escribir la grabación siguiente ya no es una raíz
				// mal declarada —eso es la fila 12— sino un tropiezo sobrevenido
				// del mecanismo (FR-042).
				require.NoError(t, os.RemoveAll(filepath.Join(raiz, fuenteDePrueba)))

				_, alGrabar := cliente.Pedir(t.Context(), schema.Contexto{},
					Peticion{Metodo: http.MethodGet, URL: servidor.URL + "/norma?otra=1"})

				return []error{alGrabar}
			},
		},
		{
			fila:   15,
			nombre: "colisión: el fichero guarda otra petición",
			clase:  schema.ClaseInesperado,
			codigo: 1,
			provocar: func(t *testing.T) []error {
				t.Helper()

				t.Setenv(VariableGrabacion, variableActiva)

				raiz := t.TempDir()
				servidor, _ := servidorIdentificado(t, servidorQueNumera())

				// «/a,b» y «/a_b» se sanean al mismo nombre: la grabación de la
				// primera ocupa el fichero de la segunda (contrato §2).
				ruta := rutaGrabada(raiz, nombreDelServidor(t, servidor, "GET", "_a_b"))
				require.NoError(t, os.MkdirAll(filepath.Dir(ruta), permisoDelDirectorioDeGrabacion))
				require.NoError(t, os.WriteFile(ruta,
					[]byte(grabacionAMano(servidor.URL+"/a,b")), permisoDeLaCopia))

				_, colision := clienteQueGraba(t, raiz).Pedir(t.Context(), schema.Contexto{},
					Peticion{Metodo: http.MethodGet, URL: servidor.URL + "/a_b"})

				return []error{colision}
			},
		},
		{
			fila:   16,
			nombre: "reproducción: la grabación falta o es de otra petición",
			clase:  schema.ClaseInesperado,
			codigo: 1,
			provocar: func(t *testing.T) []error {
				t.Helper()

				// La grabación ya va apagada para toda la tabla, que es lo que
				// la reproducción exige (FR-043).
				cliente := clienteDeReproduccion(t, grabacionesDePrueba(t))

				_, ausente := cliente.Pedir(t.Context(), schema.Contexto{},
					Peticion{Metodo: http.MethodGet, URL: sitioGrabado + "/inexistente"})

				_, ajena := cliente.Pedir(t.Context(), schema.Contexto{},
					Peticion{Metodo: http.MethodGet, URL: colisionBuscada})

				return []error{ausente, ajena}
			},
		},
	}

	for _, caso := range casos {
		t.Run(strconv.Itoa(caso.fila)+". "+caso.nombre, func(t *testing.T) {
			// Cada fila declara también la grabación apagada, y no le basta la
			// que declaró la tabla: así cualquiera de las dieciséis se puede
			// lanzar sola con -run sin depender del entorno de quien la lance.
			t.Setenv(VariableGrabacion, "")

			var provocados []error

			// Provocar la situación no puede terminar en panic: un fallo sin
			// clase se vería en las comprobaciones de abajo, pero uno que tumba
			// el proceso no llegaría siquiera a clasificarse (SC-016).
			require.NotPanics(t, func() { provocados = caso.provocar(t) },
				"la fila %d no puede terminar en panic", caso.fila)
			require.NotEmpty(t, provocados, "cada fila provoca al menos una situación")

			for _, err := range provocados {
				exigeClaseYCodigo(t, caso.clase, caso.codigo, err)
			}
		})
	}

	t.Run("la tabla es la cerrada del contrato: dieciséis filas y cuatro clases", func(t *testing.T) {
		filas := make(map[int]struct{}, len(casos))
		porClase := make(map[schema.Clase]int, len(casos))

		for _, caso := range casos {
			filas[caso.fila] = struct{}{}
			porClase[caso.clase]++
		}

		assert.Len(t, filas, 16, "una subprueba por fila del contrato §3, sin repetir ninguna")
		assert.Equal(t, map[schema.Clase]int{
			schema.ClaseFuenteNoDisponible: 3,
			schema.ClaseLimiteOTos:         5,
			schema.ClaseArgumentos:         5,
			schema.ClaseInesperado:         3,
		}, porClase,
			"el cliente produce estas cuatro clases y ninguna más: ni «no encontrado» ni «identidad humana» (FR-063)")
	})
}

// exigeClaseYCodigo comprueba las dos mitades del mecanismo de D4 sobre un error
// que de verdad salió del paquete: que declara la clase que le toca y que el
// kernel la traduce al código que le toca, tal cual y bajo las capas de contexto
// que un applet le añade al devolverlo (FR-031, contrato §4).
//
// Comprueba además, de cada error, lo que la tabla cierra: que la clase no es
// nunca «no encontrado» ni «requiere identidad humana», que el código no es
// nunca el 3 ni el 6, que no es el del éxito, y que ni clasificar el fallo ni
// leer su mensaje terminan en panic (FR-033, FR-063).
func exigeClaseYCodigo(t *testing.T, clase schema.Clase, codigo int, err error) {
	t.Helper()

	require.Error(t, err, "la situación tiene que fallar")

	fallo := falloDe(t, err)

	var (
		declarada   schema.Clase
		clasificada schema.Clase
		salida      int
		mensaje     string
	)

	require.NotPanics(t, func() {
		declarada = fallo.Clase()
		clasificada = cli.Clasificar(err)
		salida = cli.CodigoSalida(err)
		mensaje = fallo.Error()
	}, "clasificar el fallo y leer su mensaje no terminan nunca en panic (FR-033)")

	assert.Equal(t, clase, declarada, "el error declara la clase de su fila")
	assert.Equal(t, clase, clasificada, "y el kernel la reconoce con errors.As, sin que el paquete lo importe (D4)")
	assert.Equal(t, codigo, salida, "que es el código de salida de la tabla del contrato §3")
	assert.NotEmpty(t, mensaje, "ningún fallo llega sin mensaje que leer (FR-033)")

	// Envuelto por un applet —una capa o dos—, la clase y el código no cambian.
	unaCapa := fmt.Errorf("consultando la norma: %w", err)
	assert.Equal(t, clase, cli.Clasificar(unaCapa))
	assert.Equal(t, codigo, cli.CodigoSalida(unaCapa))

	dosCapas := fmt.Errorf("el applet no ha podido terminar: %w", unaCapa)
	assert.Equal(t, clase, cli.Clasificar(dosCapas))
	assert.Equal(t, codigo, cli.CodigoSalida(dosCapas))

	assert.NotEqual(t, schema.ClaseNoEncontrado, declarada,
		"quien decide si lo pedido existe es la fuente, no el cliente (FR-063)")
	assert.NotEqual(t, schema.ClaseIdentidadHumana, declarada,
		"este paquete no hace nada que exija identidad humana (FR-063)")
	assert.NotEqual(t, 3, salida, "el código 3 no lo produce ninguna ruta del cliente")
	assert.NotEqual(t, 6, salida, "y el 6 tampoco")
	assert.NotZero(t, salida, "un fallo no sale nunca con el código del éxito")
}

// falloDeLaOpcion devuelve el fallo con que la construcción de un cliente
// rechaza una opción inválida, tal como sale de New: el mensaje que se comprueba
// es el que lee quien la declaró, no uno escrito para la tabla.
func falloDeLaOpcion(t *testing.T, opcion Opcion) *Error {
	t.Helper()

	_, err := New(opcion)

	return falloDe(t, err)
}

// clienteSinEsperas es el cliente contra la red de esta tabla, con las dos
// esperas apartadas del camino: el turno del sitio, que con el intervalo por
// omisión separaría un segundo la petición del robots.txt de la del recurso, y
// el retardo entre reintentos, que en las filas que los agotan costaría
// segundos. Lo que aquí se mide es la clase del fallo; las dos políticas las
// fijan sus propias tablas sobre las duraciones pedidas (D8, D9).
func clienteSinEsperas(t *testing.T) *Cliente {
	t.Helper()

	return clienteDePrueba(t, ConIntervalo(time.Millisecond), conReloj(sinEsperar))
}

// sinEsperar es el reloj de esta tabla: no espera el retardo entre dos intentos,
// pero pasa por el mismo camino interrumpible que el de verdad, de modo que un
// contexto terminado siga cortando donde cortaría (FR-026).
func sinEsperar(ctx context.Context, _ time.Duration) error {
	return dormirInterrumpible(ctx, 0)
}

// plazoDeLaTabla es lo que dura el contexto de las filas que comprueba su
// vencimiento: lo justo para que la operación esté en marcha cuando venza.
const plazoDeLaTabla = 50 * time.Millisecond

// pedirConPlazo pide una dirección con un contexto que vence, y devuelve el
// fallo con que la operación terminó. Comprueba de paso las dos cosas que hacen
// del vencimiento lo que la fila 2 declara: que la causa del contexto llega
// intacta hasta quien llama y que el corte es en ese instante y no más tarde
// (FR-005, FR-029).
func pedirConPlazo(t *testing.T, direccion string) error {
	t.Helper()

	ctx, cancelar := context.WithTimeout(t.Context(), plazoDeLaTabla)
	defer cancelar()

	comienzo := time.Now()

	_, err := clienteSinEsperas(t).Pedir(ctx, schema.Contexto{},
		Peticion{Metodo: http.MethodGet, URL: direccion})

	require.ErrorIs(t, err, context.DeadlineExceeded, "la causa del contexto llega intacta hasta quien llama")
	assert.Less(t, time.Since(comienzo), time.Second, "la operación termina con el contexto, no más tarde")

	return err
}

// esperaAlContexto es el manejador que no responde nunca por su cuenta: quien
// termina la operación es el contexto de quien la pidió.
func esperaAlContexto() http.HandlerFunc {
	return func(_ http.ResponseWriter, peticion *http.Request) {
		<-peticion.Context().Done()
	}
}

// cadenaSinFin es el servidor que redirige toda ruta a una nueva, distinta cada
// vez: una cadena que no vuelve sobre sí misma y que por tanto solo la corta el
// tope de saltos (D10).
func cadenaSinFin(prefijo string) http.HandlerFunc {
	var saltos atomic.Int64

	return func(escritor http.ResponseWriter, peticion *http.Request) {
		http.Redirect(escritor, peticion, prefijo+strconv.FormatInt(saltos.Add(1), 10), http.StatusFound)
	}
}
