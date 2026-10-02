package httpx

import (
	"net/http"
	"sync"
	"time"

	"golang.org/x/time/rate"
)

// rafagaDelSitio es el tamaño del cubo de cada sitio: un solo token. Con una
// ráfaga mayor, las primeras peticiones saldrían todas a la vez y el ritmo solo
// empezaría a notarse después, que es ir más deprisa de lo que la fuente tolera
// justo al arrancar, que es cuando peor sienta (D8).
const rafagaDelSitio = 1

// Ritmo son los turnos por sitio de todo cliente que lo recibe con ConRitmo:
// todos esperan turno en el mismo limitador de cada sitio, de modo que el sitio
// no recibe de entre todos más de una petición por intervalo. Es lo que hace
// falta cuando un proceso construye más de un cliente —uno por llamada, en un
// servidor que atiende varias a la vez— y el ritmo tiene que ser del proceso y
// no de cada cliente. Lo demás de cada cliente sigue siendo suyo, también lo que
// recuerda del robots.txt de cada sitio.
//
// No tiene campos exportados: el intervalo se declara al construirlo y no cambia
// después, y los turnos no se negocian desde fuera. Es seguro para varias
// goroutines.
type Ritmo struct {
	// intervalo es la separación mínima entre dos peticiones a un mismo sitio,
	// una sola para todos los sitios: el ámbito del limitador es el sitio, pero
	// su valor no se negocia sitio a sitio (FR-019, FR-020).
	intervalo time.Duration

	// mu protege obtener o crear el limitador de un sitio, y nada más: la espera
	// del turno es del limitador, que tiene su propia exclusión.
	mu sync.Mutex
	// turnos es el limitador de cada sitio, por su clave, creado la primera vez
	// que un cliente ve una dirección de ese sitio y vivo lo que vive el Ritmo.
	turnos map[string]*rate.Limiter
}

// NuevoRitmo construye los turnos de un proceso, con la separación mínima entre
// dos peticiones a un mismo sitio. El intervalo se comprueba donde se comprueba
// el de ConIntervalo, al construir el cliente que recibe el Ritmo: uno nulo o
// negativo es allí un error de argumentos (ConRitmo).
func NuevoRitmo(intervalo time.Duration) *Ritmo {
	return &Ritmo{intervalo: intervalo, turnos: make(map[string]*rate.Limiter)}
}

// turnoDe devuelve el limitador del sitio de esa clave, creándolo la primera vez
// que se pide: un cubo de un token que se rellena cada intervalo. Todo cliente
// con el mismo Ritmo recibe el mismo, que es lo que los pone a esperar turno uno
// detrás de otro.
func (r *Ritmo) turnoDe(clave string) *rate.Limiter {
	r.mu.Lock()
	defer r.mu.Unlock()

	if existente, hay := r.turnos[clave]; hay {
		return existente
	}

	nuevo := rate.NewLimiter(rate.Every(r.intervalo), rafagaDelSitio)
	r.turnos[clave] = nuevo

	return nuevo
}

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
	// el limitador de cada destino, que el registro toma del Ritmo del cliente.
	// No hay un mapa por decorador: la clave de sitio es una (data-model.md §4,
	// D14).
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
