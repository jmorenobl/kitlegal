package httpx

import (
	"context"
	"errors"
	"fmt"
	"math"
	"net/http"
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
