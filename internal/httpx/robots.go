package httpx

import (
	"context"
	"errors"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"time"

	"github.com/temoto/robotstxt"
)

// rutaDelRobots es donde RFC 9309 §2.3 dice que un sitio publica sus reglas, y
// la única ruta que este paquete pide sin que nadie se la haya pedido.
const rutaDelRobots = "/robots.txt"

// agenteDeLasReglas es el nombre con el que el proyecto busca su grupo de reglas
// en el robots.txt: el nombre con que se identifica, sin la versión. FindGroup
// compara en minúsculas y se queda con el prefijo de agente más largo que casa,
// con «*» como comodín más débil, así que un «User-agent: kitlegal» gana a un
// «User-agent: *» y un «User-agent: kit» también casaría (D7).
const agenteDeLasReglas = "kitlegal"

// reglasDelSitio es el resultado de obtener e interpretar el robots.txt de un
// sitio, **o** el de no haber podido hacerlo. Las dos cosas se cachean igual:
// un sitio caído, limitado o inalcanzable no se vuelve a pedir por cada ruta
// (FR-015, data-model.md §5).
type reglasDelSitio struct {
	// datos son las reglas interpretadas, o el «todo permitido» de un sitio que
	// no las declara. Solo se consulta cuando denegado es nil.
	datos *robotstxt.RobotsData
	// denegado no es nil cuando el sitio quedó denegado entero, que es lo que
	// ocurre cuando el permiso no se pudo obtener: 429, 5xx o fallo de
	// transporte agotados los intentos, contenido ilegible, cadena de
	// redirecciones en bucle o excedida y estado no previsto (FR-015).
	denegado *Error
}

// permite dice si la ruta pedida se puede pedir. Un sitio denegado entero lo
// está para toda ruta; de un robots.txt interpretado solo se honran las reglas
// de permiso y denegación de rutas para el agente del proyecto, y ninguna otra
// directiva —Crawl-delay incluida, que el tipo Group expone y este paquete no
// lee (FR-015, D7).
//
// Lo que se evalúa es RequestURI() —la ruta codificada más la consulta, «/» si
// la ruta va vacía—, que es lo que RFC 9309 §2.2.2 compara.
func (r *reglasDelSitio) permite(peticion *http.Request) error {
	if r.denegado != nil {
		return r.denegado
	}

	if r.datos.TestAgent(peticion.URL.RequestURI(), agenteDeLasReglas) {
		return nil
	}

	implicada := Peticion{Metodo: peticion.Method, URL: peticion.URL.String()}

	return errorDeLimiteOTos(implicada, 0, nil,
		"el robots.txt del sitio no autoriza esta ruta al agente "+agenteDeLasReglas)
}

// decoradorDeRobots es el escalón que decide, por sitio, si la ruta se puede
// pedir, y que cachea la decisión mientras viva el cliente (FR-013, FR-015).
//
// Va **por encima** de los reintentos —y por tanto del ritmo—, de modo que la
// obtención del robots.txt pase por los dos: un 5xx suyo se reintenta como
// cualquier otro y cada petición suya espera su turno en el sitio (FR-017, D3).
// Y va por debajo de la identificación, que es la razón de que la petición del
// robots.txt la construya nuevaPeticionIdentificada: la fabrica este decorador,
// por debajo del que pone la cabecera, así que sin ese constructor saldría sin
// ella (FR-009, D3).
type decoradorDeRobots struct {
	// sitios es el registro de sitios del cliente, el mismo que usa el ritmo:
	// la clave de sitio es una y la exclusión tiene que ser compartida
	// (data-model.md §4, D14).
	sitios *sitios
	// siguiente es el escalón al que se entrega la petición autorizada, y
	// también por el que baja la del propio robots.txt.
	siguiente http.RoundTripper
}

// conRobots envuelve el escalón que recibe con el decorador, sobre el registro
// de sitios del cliente.
func conRobots(siguiente http.RoundTripper, sitios *sitios) http.RoundTripper {
	return &decoradorDeRobots{sitios: sitios, siguiente: siguiente}
}

// RoundTrip consulta las reglas del sitio de destino y solo entrega la petición
// al escalón siguiente si autorizan la ruta. Como cada salto de una redirección
// baja por la cadena entera, esto ocurre en el sitio de destino de cada salto y
// ninguno hereda el permiso del anterior (FR-011).
func (r *decoradorDeRobots) RoundTrip(peticion *http.Request) (*http.Response, error) {
	reglas, err := r.reglasDe(peticion)
	if err != nil {
		return nil, err
	}

	if err := reglas.permite(peticion); err != nil {
		return nil, err
	}

	return r.siguiente.RoundTrip(peticion)
}

// reglasDe devuelve las reglas del sitio de destino, obteniéndolas la primera
// vez que se ve ese sitio. La exclusión del sitio se mantiene tomada durante la
// obtención: N goroutines que piden a la vez el mismo sitio producen **una**
// petición de robots.txt y N esperas, en vez de N obtenciones (SC-003, D7).
//
// El error que devuelve es el de la operación que ha terminado antes de llegar a
// evaluar ninguna regla —el contexto vencido o cancelado, clase 4—, y es el
// único caso que **no** deja entrada en la caché: la petición siguiente volverá
// a intentarlo (FR-015, data-model.md §5).
func (r *decoradorDeRobots) reglasDe(peticion *http.Request) (*reglasDelSitio, error) {
	delSitio := r.sitios.de(peticion.URL)

	delSitio.mu.Lock()
	defer delSitio.mu.Unlock()

	if delSitio.reglas != nil {
		return delSitio.reglas, nil
	}

	reglas, err := r.obtener(peticion.Context(), peticion.URL)
	if err != nil {
		return nil, err
	}

	delSitio.reglas = reglas

	return reglas, nil
}

// obtener pide el robots.txt del sitio y lo interpreta. Sigue su **propia**
// cadena de redirecciones, con el mismo tope que la del recurso pedido (RFC 9309
// §2.3.1.2 pide seguir al menos cinco), pero ningún salto se evalúa contra
// ningún robots.txt —ni el del origen ni el del destino—, que es lo que impide
// la recursión (FR-016). Lo que sí se les aplica es la identificación, el ritmo
// del sitio de destino y el plazo del contexto (FR-017).
//
// Una cadena en bucle o que excede el tope significa que el permiso no se pudo
// obtener: deniega con la clase del límite y no con la clase 4 de FR-011, porque
// aquí lo que falta no es el recurso sino el permiso para pedirlo (FR-015).
func (r *decoradorDeRobots) obtener(ctx context.Context, delSitio *url.URL) (*reglasDelSitio, error) {
	destino := &url.URL{Scheme: delSitio.Scheme, Host: delSitio.Host, Path: rutaDelRobots}
	visitadas := make(map[string]struct{}, topeDeRedirecciones+1)

	for range topeDeRedirecciones + 1 {
		if _, repetida := visitadas[destino.String()]; repetida {
			return sinPermiso(destino, 0, nil,
				"la cadena de redirecciones del robots.txt vuelve sobre una dirección ya visitada: "+
					destino.String()), nil
		}

		visitadas[destino.String()] = struct{}{}

		respuesta, err := pedirElRobots(ctx, r.siguiente, destino)
		if err != nil {
			return sinRespuestaDelRobots(ctx, destino, err)
		}

		siguiente, hayQueSeguir := saltoDelRobots(destino, respuesta)
		if !hayQueSeguir {
			return interpretarElRobots(destino, respuesta), nil
		}

		destino = siguiente
	}

	return sinPermiso(destino, 0, nil,
		"la cadena de redirecciones del robots.txt supera los "+
			strconv.Itoa(topeDeRedirecciones)+" saltos"), nil
}

// pedirElRobots baja una petición del robots.txt por la cadena que queda por
// debajo del decorador y devuelve lo que la fuente respondió, con el cuerpo ya
// leído y el *http.Response ya cerrado aquí dentro.
//
// La petición la construye nuevaPeticionIdentificada, el único constructor del
// paquete: la fabrica este escalón, por debajo del que pone la cabecera, así que
// es la única forma de que nazca identificada (FR-009, D3).
//
// Nace de cero, y no copiada de la petición que ha provocado la consulta, y por
// eso no lleva más cabecera que la identificación: tampoco el formato que quien
// llama pidió para su recurso. Un robots.txt es texto plano (RFC 9309 §2.3), y
// pedirlo en XML o en JSON invitaría a un sitio que negocia el contenido a
// responder con otra cosa; clonar aquí la petición entrante rompería esa garantía,
// y TestPedirConAcepta lo detectaría (contrato httpx-acepta-e-instante §1 de H4).
func pedirElRobots(ctx context.Context, siguiente http.RoundTripper, destino *url.URL) (recibida, error) {
	peticion, err := nuevaPeticionIdentificada(ctx, http.MethodGet, destino.String())
	if err != nil {
		return recibida{}, err
	}

	respuesta, err := siguiente.RoundTrip(peticion)
	if err != nil {
		return recibida{}, err
	}

	cuerpo, errDeLectura := io.ReadAll(respuesta.Body)

	// El cierre se intenta siempre y su fallo no se pierde: sin cuerpo entero
	// —o sin haberlo podido soltar— no hay robots.txt que interpretar, y quien
	// llama convierte eso en la denegación que le corresponde.
	if errAlCerrar := respuesta.Body.Close(); errDeLectura == nil {
		errDeLectura = errAlCerrar
	}

	if errDeLectura != nil {
		return recibida{}, errDeLectura
	}

	return recibida{
		estado:    respuesta.StatusCode,
		cabeceras: Cabeceras(respuesta.Header),
		cuerpo:    cuerpo,
	}, nil
}

// saltoDelRobots dice si la respuesta es una redirección que hay que seguir y a
// dónde lleva. Un 3xx que no se puede seguir —uno de los que la biblioteca no
// sigue, uno sin Location o uno cuya Location no se puede interpretar— no es un
// salto: es una respuesta más, y como tal la clasifica interpretarElRobots, que
// la trata como el estado no previsto que es.
func saltoDelRobots(actual *url.URL, respuesta recibida) (*url.URL, bool) {
	if !esRedireccionSeguible(respuesta.estado) {
		return nil, false
	}

	destino := respuesta.cabeceras.Get("Location")
	if destino == "" {
		return nil, false
	}

	referencia, err := url.Parse(destino)
	if err != nil {
		return nil, false
	}

	return actual.ResolveReference(referencia), true
}

// interpretarElRobots clasifica la respuesta final de la cadena por la lista
// cerrada de FR-015. Las dos primeras filas no se le pasan a la biblioteca:
// implementa la regla de Google «todo 4xx es permitir», y el 429 es justo la
// excepción que el spec exige —un 429 es la fuente pidiendo que paremos, y abrir
// la consulta a ciegas por venir en un 4xx sería ir más deprisa de lo que
// tolera—; el 5xx lo trataría como «todo desautorizado», que es la misma
// decisión pero sin decir por qué (FR-015, FR-030, D7).
//
// Lo demás sí: 2xx interpretable son las reglas del sitio, 2xx vacío y 4xx son
// «sin reglas» y permiten, 2xx ilegible y estado no previsto devuelven error y
// deniegan.
func interpretarElRobots(destino *url.URL, respuesta recibida) *reglasDelSitio {
	switch {
	case respuesta.estado == http.StatusTooManyRequests:
		return &reglasDelSitio{denegado: errorDe429(peticionDelRobots(destino), respuesta.cabeceras, time.Now(),
			"la fuente ha alcanzado su límite de peticiones al pedirle su robots.txt")}

	case respuesta.estado >= http.StatusInternalServerError:
		return sinPermiso(destino, respuesta.estado, nil,
			"el sitio no ha entregado su robots.txt, y sin él no hay permiso para pedirle nada")
	}

	datos, err := robotstxt.FromStatusAndBytes(respuesta.estado, respuesta.cuerpo)
	if err != nil {
		return sinPermiso(destino, respuesta.estado, err,
			"el robots.txt del sitio no se ha podido interpretar")
	}

	return &reglasDelSitio{datos: datos}
}

// sinRespuestaDelRobots distingue las dos formas en que la obtención puede
// terminar sin respuesta, que no son de la misma clase.
//
// El contexto que vence o se cancela —directamente, o en el escalón de más abajo
// que ya declaró su clase: el turno que no llega y la espera entre reintentos
// interrumpida— termina la operación entera sin llegar a evaluar ninguna regla,
// con la clase de lo que no llegó a la fuente y **sin** dejar entrada en la
// caché: la petición siguiente volverá a intentarlo (FR-015, FR-029).
//
// Un fallo de transporte que ha sobrevivido a la política de reintentos es otra
// cosa: el permiso no se pudo obtener, y el sitio queda denegado con la clase del
// límite, que prevalece aquí sobre FR-029 (FR-015).
func sinRespuestaDelRobots(ctx context.Context, destino *url.URL, causa error) (*reglasDelSitio, error) {
	var declarado *Error
	if errors.As(causa, &declarado) {
		return nil, declarado
	}

	if ctx.Err() != nil {
		return nil, errorDeFuenteNoDisponible(peticionDelRobots(destino), 0, causa,
			"la operación ha terminado obteniendo el robots.txt del sitio")
	}

	return sinPermiso(destino, 0, causa,
		"no se ha podido obtener el robots.txt del sitio, y sin él no hay permiso para pedirle nada"), nil
}

// sinPermiso es el sitio que queda denegado entero porque su robots.txt no se
// pudo obtener ni interpretar: la clase es la del límite o los términos de uso
// (FR-014, FR-015), y la petición que el mensaje nombra es la del robots.txt,
// que es la que falló.
func sinPermiso(destino *url.URL, estado int, causa error, motivo string) *reglasDelSitio {
	return &reglasDelSitio{denegado: errorDeLimiteOTos(peticionDelRobots(destino), estado, causa, motivo)}
}

// peticionDelRobots es la petición que este decorador fabrica —siempre un GET, y
// siempre a la misma ruta del sitio— tal como la nombran sus errores: es la que
// falló cuando lo que falta es el permiso, y no la que quien llama pidió
// (FR-034).
func peticionDelRobots(destino *url.URL) Peticion {
	return Peticion{Metodo: http.MethodGet, URL: destino.String()}
}
