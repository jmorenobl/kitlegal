package httpx

import (
	"net/url"
	"strings"
	"sync"
)

// Puertos que el esquema implica cuando la dirección no declara ninguno. Son
// los dos únicos esquemas que el paquete llega a ver: cualquier otro es un
// fallo de la clase «argumentos» antes de que se resuelva ningún sitio (D19).
const (
	puertoHTTP  = "80"
	puertoHTTPS = "443"
)

// sitio es lo que el cliente recuerda de un sitio mientras vive: hoy solo su
// identidad. Le cuelgan el limitador de ritmo y las reglas de su robots.txt
// cuando entra el decorador que usa cada uno, porque los dos tienen que
// compartir esta entrada y su exclusión —no hay un mapa por decorador
// (data-model.md §4, D7, D14).
type sitio struct {
	// clave es la que indexa el sitio en el mapa; la guarda también aquí para
	// que un sitio se pueda nombrar sin volver a derivarla de una dirección.
	clave string
}

// sitios es el mapa de sitios de un cliente: una entrada por clave de sitio,
// creada la primera vez que se ve una dirección de ese sitio y viva lo que vive
// el cliente. El mutex es de grano corto —solo protege obtener o crear la
// entrada— porque el cliente se usa desde varias goroutines a la vez (FR-021,
// D14).
type sitios struct {
	mu       sync.Mutex
	porClave map[string]*sitio
}

// nuevosSitios construye el mapa vacío de sitios de un cliente.
func nuevosSitios() *sitios {
	return &sitios{porClave: make(map[string]*sitio)}
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

	nuevo := &sitio{clave: clave}
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
