package httpx

import (
	"net/http"
)

// decoradorDeRitmo es el escalón que impide que el cliente vaya más deprisa de
// lo que un sitio tolera: toda petición que baja por él espera antes el turno
// del sitio al que va, y ninguna sale sin haberlo esperado (FR-019, FR-021).
//
// Va por debajo de la identificación, de modo que lo que espera turno es una
// petición ya identificada, y por debajo de los reintentos y del robots.txt
// cuando entren, para que también cada reintento y la propia obtención del
// robots.txt ocupen su turno en el sitio (FR-017, D3).
type decoradorDeRitmo struct {
	// sitios es el registro de sitios del cliente —uno solo, el mismo que
	// compartirán los decoradores que necesiten estado por sitio—, del que sale
	// el limitador de cada destino. No hay un mapa por decorador: la clave de
	// sitio es una (data-model.md §4, D14).
	sitios *sitios
	// siguiente es el escalón al que se entrega la petición cuando le llega el
	// turno.
	siguiente http.RoundTripper
}

// conRitmo envuelve el escalón que recibe con el decorador, sobre el registro de
// sitios del cliente.
func conRitmo(siguiente http.RoundTripper, sitios *sitios) http.RoundTripper {
	return &decoradorDeRitmo{sitios: sitios, siguiente: siguiente}
}

// RoundTrip espera el turno del sitio de destino y solo entonces entrega la
// petición al escalón siguiente. La espera es la del limitador de ese sitio
// —Wait «blocks until [a token] can be obtained or its associated
// context.Context is canceled» (go doc golang.org/x/time/rate Limiter)—, así que
// el ritmo no necesita ni goroutines ni relojes propios: el decorador no emite
// nada en paralelo ni por su cuenta (FR-021).
//
// Mientras se espera no hay ninguna petición emitida, y por eso el contexto que
// termina durante la espera deja la operación sin petición y con la clase que le
// corresponde (FR-022).
func (r *decoradorDeRitmo) RoundTrip(peticion *http.Request) (*http.Response, error) {
	if err := r.sitios.de(peticion.URL).limitador.Wait(peticion.Context()); err != nil {
		return nil, sinTurno(peticion, err)
	}

	return r.siguiente.RoundTrip(peticion)
}

// sinTurno declara la clase del fallo aquí mismo, que es donde se sabe qué pasó:
// la espera terminó sin turno y sin petición emitida, de modo que es la
// operación la que no ha podido llegar a la fuente (clase «fuente no
// disponible», FR-022, tabla de clases §3 fila 2). Quien clasifica el resultado
// de la cadena lo entrega tal cual, porque desde fuera esto no se distingue de
// un sitio que no responde.
//
// Las dos formas en que la espera puede terminar sin turno son de la misma
// clase, pero no son lo mismo para quien lee el mensaje: el contexto que vence o
// se cancela mientras se espera, y el turno que ni siquiera cabe en el plazo que
// quedaba, que Wait resuelve sin llegar a esperar —«it returns an error if …
// the expected wait time exceeds the Context's Deadline» (go doc
// golang.org/x/time/rate Limiter.WaitN)—.
func sinTurno(peticion *http.Request, causa error) *Error {
	implicada := Peticion{Metodo: peticion.Method, URL: peticion.URL.String()}

	if peticion.Context().Err() != nil {
		return errorDeFuenteNoDisponible(implicada, 0, causa,
			"la operación ha terminado esperando turno en el sitio")
	}

	return errorDeFuenteNoDisponible(implicada, 0, causa,
		"el turno en el sitio no llega dentro del plazo de la operación")
}
