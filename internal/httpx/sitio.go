package httpx

import (
	"net/url"
	"strings"
	"sync"
	"time"

	"golang.org/x/time/rate"
)

// Puertos que el esquema implica cuando la dirección no declara ninguno. Son
// los dos únicos esquemas que el paquete llega a ver: cualquier otro es un
// fallo de la clase «argumentos» antes de que se resuelva ningún sitio (D19).
const (
	puertoHTTP  = "80"
	puertoHTTPS = "443"
)

// sitio es lo que el cliente recuerda de un sitio mientras vive: su identidad,
// su ritmo y las reglas de su robots.txt. Los dos decoradores que necesitan
// estado por sitio comparten esta entrada y su exclusión —no hay un mapa por
// decorador (data-model.md §4, D7, D14).
type sitio struct {
	// clave es la que indexa el sitio en el mapa; la guarda también aquí para
	// que un sitio se pueda nombrar sin volver a derivarla de una dirección.
	clave string
	// limitador es el ritmo de este sitio y solo de este sitio: un cubo de un
	// token que se rellena cada intervalo, de modo que dos peticiones seguidas
	// al mismo sitio salgan separadas al menos ese intervalo y las de otro
	// sitio no compitan por él (FR-019, D8). Nace con el sitio y no cambia
	// después; es seguro para varias goroutines —«a Limiter is safe for
	// simultaneous use by multiple goroutines» (go doc golang.org/x/time/rate
	// Limiter)—, así que no necesita la exclusión del mapa más allá de aquí.
	limitador *rate.Limiter

	// mu protege las reglas de este sitio y se mantiene tomada mientras se
	// obtienen, que es lo que hace que varias goroutines que piden a la vez el
	// mismo sitio produzcan **una** petición de robots.txt y no una cada una
	// (SC-003, D7). Es la del sitio y no la del mapa: obtener el robots.txt de
	// un sitio no puede dejar en espera a los demás, que tienen su propio cupo
	// (FR-019).
	mu sync.Mutex
	// reglas es lo que se obtuvo del robots.txt del sitio —o el resultado de no
	// haber podido obtenerlo—, nil hasta la primera evaluación e inmutable
	// desde entonces: la decisión se cachea en memoria, sin caducidad y sin
	// tocar el disco, así que dura lo que vive el cliente (FR-015, FR-018).
	reglas *reglasDelSitio
}

// rafagaDelSitio es el tamaño del cubo de cada sitio: un solo token. Con una
// ráfaga mayor, las primeras peticiones saldrían todas a la vez y el ritmo solo
// empezaría a notarse después, que es ir más deprisa de lo que la fuente tolera
// justo al arrancar, que es cuando peor sienta (D8).
const rafagaDelSitio = 1

// sitios es el mapa de sitios de un cliente: una entrada por clave de sitio,
// creada la primera vez que se ve una dirección de ese sitio y viva lo que vive
// el cliente. El mutex es de grano corto —solo protege obtener o crear la
// entrada— porque el cliente se usa desde varias goroutines a la vez (FR-021,
// D14).
type sitios struct {
	mu       sync.Mutex
	porClave map[string]*sitio
	// intervalo es la separación mínima entre dos peticiones a un mismo sitio.
	// Vive aquí, y no en cada entrada, porque es uno solo para todos los sitios
	// de un cliente: el ámbito del limitador es el sitio, pero su valor no se
	// negocia sitio a sitio (FR-019, FR-020).
	intervalo time.Duration
}

// nuevosSitios construye el mapa vacío de sitios de un cliente, con el intervalo
// con el que nacerá el limitador de cada uno.
func nuevosSitios(intervalo time.Duration) *sitios {
	return &sitios{porClave: make(map[string]*sitio), intervalo: intervalo}
}

// de devuelve el sitio de una dirección, creándolo la primera vez que se ve.
// Dos direcciones con la misma clave reciben siempre la misma entrada, que es
// lo que hace que su ritmo y su robots.txt sean uno solo por cliente.
func (s *sitios) de(direccion *url.URL) *sitio {
	clave := claveDeSitio(direccion)

	s.mu.Lock()
	defer s.mu.Unlock()

	if existente, hay := s.porClave[clave]; hay {
		return existente
	}

	nuevo := &sitio{
		clave:     clave,
		limitador: rate.NewLimiter(rate.Every(s.intervalo), rafagaDelSitio),
	}
	s.porClave[clave] = nuevo

	return nuevo
}

// claveDeSitio es la identidad de un sitio —esquema, host y puerto—, que es la
// unidad del robots.txt (RFC 9309) y la del ritmo (FR-013, FR-015, FR-019). Ni
// la ruta ni la consulta entran en ella: dos rutas del mismo origen son el
// mismo sitio, y dos direcciones que difieren en el esquema, en el host o en el
// puerto son sitios distintos, cada uno con su primera petición.
//
// url.Parse ya devuelve el esquema en minúsculas; el host no, así que se
// normaliza aquí.
func claveDeSitio(direccion *url.URL) string {
	puerto := direccion.Port()
	if puerto == "" {
		puerto = puertoPorOmision(direccion.Scheme)
	}

	return direccion.Scheme + "://" + strings.ToLower(direccion.Hostname()) + ":" + puerto
}

// puertoPorOmision es el puerto que el esquema implica cuando la dirección no
// declara ninguno, de modo que http://fuente y http://fuente:80 sean un solo
// sitio.
func puertoPorOmision(esquema string) string {
	if esquema == "https" {
		return puertoHTTPS
	}

	return puertoHTTP
}
