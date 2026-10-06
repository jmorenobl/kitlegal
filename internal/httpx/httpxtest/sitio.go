// Package httpxtest lleva el sitio local de prueba contra el que un test cuenta
// lo que un cliente de httpx pide de verdad: un servidor en la propia máquina
// que apunta cada petición que recibe —método, ruta, cuerpo y cookie—, en su
// orden, y responde lo que el test decide (H23 contracts/httpx-formulario.md §7;
// research.md D32).
//
// No es un paquete de test y solo lo usan tests: vive aquí porque la biblioteca
// HTTP solo se importa bajo internal/httpx (regla R2), y el control que cuenta
// las peticiones de una consulta está en el paquete de su adaptador, que no
// puede levantar un servidor (plan.md de H23, Complexity Tracking). Por eso
// nada de lo que exporta nombra un tipo de esa biblioteca: quien lo usa decide
// las respuestas y compara lo recibido con los tipos de este paquete y con los
// de httpx. El binario no lo enlaza, y TestElBinarioNoEnlazaLosEjemplos lo
// comprueba.
//
// Lo que da y la reproducción de httpx.Replay no dan lo mismo: aquí las
// peticiones bajan por la cadena entera del cliente —el robots.txt del sitio,
// el ritmo, los reintentos— y llegan a un servidor, de modo que el robots.txt
// y un reintento se cuentan como lo que son, una petición más.
package httpxtest

import (
	"io"
	"net/http"
	"net/http/httptest"
	"slices"
	"strings"
	"sync"
	"testing"

	"github.com/jmorenobl/kitlegal/internal/httpx"
)

// cabeceraDeCookies es la cabecera en la que una petición trae sus cookies.
const cabeceraDeCookies = "Cookie"

// Recibida es lo que el sitio apunta de una petición que recibe. Es comparable:
// un test compara lo apuntado con la lista que espera, de una vez.
type Recibida struct {
	// Metodo es el de la petición: «GET», «HEAD», «POST».
	Metodo string
	// Ruta es la que se pidió, con su consulta si la trae: «/buscar»,
	// «/pagina?idioma=es».
	Ruta string
	// Cuerpo es el de la petición, entero y tal como llegó: el de un
	// formulario, sus campos codificados. Vacío si la petición no trae ninguno.
	Cuerpo string
	// Cookie es lo que traía su cabecera Cookie —«sesion=1»—, y todas sus
	// líneas unidas con «; » si traía más de una. Vacía si llegó sin cookies.
	Cookie string
}

// Respuesta es lo que el sitio responde a una petición. El valor cero es un 200
// sin cabeceras propias ni cuerpo, que para un robots.txt es un sitio sin
// reglas.
type Respuesta struct {
	// Estado es el estado HTTP. Cero es 200, como en un manejador que no
	// declara ninguno.
	Estado int
	// Cabeceras son las que la respuesta lleva, cada una con todos sus valores.
	// Son las que el cliente lee después en httpx.Respuesta.
	Cabeceras httpx.Cabeceras
	// Cuerpo es el de la respuesta.
	Cuerpo string
}

// Responder decide lo que el sitio responde a cada petición que recibe, el
// robots.txt incluido, con lo que el sitio ha apuntado de ella.
//
// El servidor atiende cada petición en su gorrutina: si el test pide a la vez,
// Responder se llama a la vez, y lo que comparta tiene que protegerlo él.
type Responder func(Recibida) Respuesta

// Sitio es un sitio local de prueba. Lo levanta NuevoSitio, y es seguro para uso
// concurrente: se le puede pedir y leer lo apuntado desde varias gorrutinas.
type Sitio struct {
	// direccion es la del servidor, que no cambia.
	direccion string
	// responder y fallar son lo que el sitio recibió al nacer: quien decide cada
	// respuesta y la forma de hacer fallar al test que lo levantó.
	responder Responder
	fallar    func(formato string, argumentos ...any)

	mu        sync.Mutex
	recibidas []Recibida
}

// NuevoSitio levanta un servidor local que apunta cada petición que recibe y le
// responde lo que responder decide, y lo cierra cuando el test termina. El
// cierre espera a las peticiones que estén en curso.
//
// No pone nada de su parte: ni un robots.txt ni un «no encontrado». Toda
// petición, también la del robots.txt que un cliente de httpx hace antes que
// ninguna otra, se apunta y pasa por responder.
func NuevoSitio(tb testing.TB, responder Responder) *Sitio {
	tb.Helper()

	sitio := &Sitio{responder: responder, fallar: tb.Errorf}

	servidor := httptest.NewServer(http.HandlerFunc(sitio.atender))
	tb.Cleanup(servidor.Close)

	sitio.direccion = servidor.URL

	return sitio
}

// URL es la dirección del sitio —esquema, host y puerto, sin barra final—, a la
// que quien pide añade la ruta: «http://127.0.0.1:49152».
func (s *Sitio) URL() string {
	return s.direccion
}

// Recibidas devuelve lo que el sitio ha apuntado hasta ahora, en el orden en
// que lo recibió. Es una copia: tocarla no cambia lo apuntado.
func (s *Sitio) Recibidas() []Recibida {
	s.mu.Lock()
	defer s.mu.Unlock()

	return slices.Clone(s.recibidas)
}

// atender apunta la petición y escribe lo que responder decide de ella. La
// apunta antes de responder: quien ya tiene su respuesta encuentra su petición
// entre las recibidas.
//
// Una petición cuyo cuerpo no llega entero se apunta también, con lo que llegó,
// porque el sitio la ha recibido y lo que cuenta es eso. Pero no es la que
// quien la envió quiso hacer, y un recuento hecho con ella no sería el de lo que
// el test cree haber pedido: el test falla, quien pide recibe un 400 y
// responder no decide nada.
func (s *Sitio) atender(escritor http.ResponseWriter, peticion *http.Request) {
	cuerpo, err := io.ReadAll(peticion.Body)

	recibida := Recibida{
		Metodo: peticion.Method,
		Ruta:   peticion.URL.RequestURI(),
		Cuerpo: string(cuerpo),
		Cookie: strings.Join(peticion.Header.Values(cabeceraDeCookies), "; "),
	}

	s.mu.Lock()
	s.recibidas = append(s.recibidas, recibida)
	s.mu.Unlock()

	if err != nil {
		s.fallar("el sitio de prueba no ha recibido entero el cuerpo de %s %s: %v", recibida.Metodo, recibida.Ruta, err)
		http.Error(escritor, "el cuerpo de la petición no ha llegado entero", http.StatusBadRequest)

		return
	}

	escribir(escritor, s.responder(recibida))
}

// escribir entrega la respuesta decidida: sus cabeceras, cada una con todos sus
// valores, su estado y su cuerpo.
//
// El fallo al escribir el cuerpo no se recoge, y no por descuido: solo puede
// decir que quien pidió ya no está —su plazo venció, o cerró— o que el estado o
// el método no admiten cuerpo, y en los dos casos lo que cuenta es lo que el
// cliente recibió, que es lo que su test mira. Hacer fallar aquí al test daría
// por roto al que comprueba precisamente un plazo vencido.
func escribir(escritor http.ResponseWriter, respuesta Respuesta) {
	cabeceras := escritor.Header()

	for nombre, valores := range respuesta.Cabeceras {
		for _, valor := range valores {
			cabeceras.Add(nombre, valor)
		}
	}

	if respuesta.Estado != 0 {
		escritor.WriteHeader(respuesta.Estado)
	}

	_, _ = io.WriteString(escritor, respuesta.Cuerpo)
}
