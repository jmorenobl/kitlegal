package httpx

import (
	"context"
	"io"
	"net/http"
	"net/http/cookiejar"
	"net/url"
	"strconv"
	"strings"

	"github.com/jmorenobl/kitlegal/internal/core/schema"
)

// Las dos cabeceras que lleva todo envío de un formulario, además de la
// identificación del proyecto. La primera dice cómo va codificado el cuerpo, que
// es la forma en que un navegador envía un formulario (go doc
// net/url.Values.Encode). La segunda es la que pone la propia página del
// buscador al consultar y la que llevaba la petición que se probó a mano: es una
// constante del paquete para todo envío, y no cambia quién dice ser el que pide,
// que sigue siendo el agente del proyecto (research D5 y V29 de H23).
const (
	cabeceraDelCuerpo    = "Content-Type"
	cuerpoDeUnFormulario = "application/x-www-form-urlencoded"

	cabeceraDeQuienEnvia = "X-Requested-With"
	enviadoPorLaPagina   = "XMLHttpRequest"
)

// ConFormulario declara la única dirección a la que el cliente puede enviar un
// formulario de consulta: la de la fila de su fuente en docs/SOURCES.md. Es lo
// que abre la única excepción a GET y HEAD del módulo, y la abre solo para esa
// dirección y solo desde una Consulta de ese cliente; sin la opción, el cliente
// no abre consultas y todo POST sigue siendo un fallo de argumentos
// (constitución, principio I; FR-030, FR-031).
//
// Vale para New y para Replay, porque una consulta se reproduce como se pide.
// Una dirección vacía, que no se puede interpretar, que no es absoluta o que no
// es de esquema http ni https es un error de argumentos, como en las demás
// opciones. Se guarda tal como la escribe url.URL.String, que es con lo que se
// compara, entera, la de cada envío (contrato httpx-formulario §1 y §2).
func ConFormulario(direccion string) Opcion {
	return func(config *configuracionDelCliente) error {
		if direccion == "" {
			return errorDeArgumentos(Peticion{}, nil,
				"la dirección del formulario de consulta no puede ir vacía (ConFormulario)")
		}

		interpretada, err := url.Parse(direccion)
		if err != nil {
			return errorDeArgumentos(Peticion{}, err,
				"la dirección del formulario de consulta no se puede interpretar (ConFormulario): "+
					strconv.Quote(direccion))
		}

		if !interpretada.IsAbs() {
			return errorDeArgumentos(Peticion{}, nil,
				"la dirección del formulario de consulta tiene que ser absoluta (ConFormulario): "+direccion)
		}

		if interpretada.Scheme != esquemaHTTP && interpretada.Scheme != esquemaHTTPS {
			return errorDeArgumentos(Peticion{}, nil,
				"la dirección del formulario de consulta tiene que ser de esquema http o https (ConFormulario): "+
					direccion)
		}

		config.formulario = interpretada.String()

		return nil
	}
}

// Consulta es donde viven las cookies: las peticiones que se piden por ella —la
// página del buscador y, con la cookie de sesión que esa página da, el envío de
// su formulario— llevan las cookies que el sitio les haya dado a ellas, y nadie
// más las lleva. Otra Consulta del mismo cliente empieza sin ninguna, y
// Cliente.Pedir ni las lleva ni las guarda (FR-032).
//
// Vive lo que quien la abrió la tenga en la mano: su almacén está en memoria, no
// se escribe en ningún sitio y se va con ella. Es segura para uso concurrente,
// como su cliente, aunque una consulta es una secuencia.
type Consulta struct {
	// cliente es el que la abrió. Suyos son el formulario declarado, la hora y el
	// registrador, y suya es la cadena: una consulta no tiene otra.
	cliente *Cliente
	// emisor ejecuta la cadena del cliente con el almacén de cookies de esta
	// consulta, que es lo único que es de ella y no de él.
	emisor *http.Client
}

// Consulta abre una consulta: sus peticiones comparten las cookies que el sitio
// les dé, y con nadie más. Solo la da un cliente con formulario; en cualquier
// otro es un error de argumentos, porque no hay ningún envío para el que guardar
// una cookie (contrato httpx-formulario §2).
//
// La consulta pide por la misma cadena que su cliente —la identificación, el
// robots.txt y el ritmo de cada sitio, los reintentos, la grabación—, con un
// cliente HTTP propio que solo añade el almacén. El robots.txt no recibe sus
// cookies: lo pide su escalón de la cadena, por debajo de ese cliente HTTP
// (research D3 de H23).
//
// El almacén nace sin lista de sufijos públicos, que sería una dependencia nueva
// (ADR 0036). Sin ella, un sitio podría dar una cookie que valiera para otros de
// su mismo sufijo; aquí no llega a ninguno, porque el almacén es de una sola
// consulta, que no sigue redirecciones y solo pide lo que su adaptador le pide.
func (c *Cliente) Consulta() (*Consulta, error) {
	if c.formulario == sinFormulario {
		return nil, errorDeArgumentos(Peticion{}, nil,
			"el cliente no declara ningún formulario de consulta (ConFormulario): no hay consulta que abrir")
	}

	almacen, err := cookiejar.New(nil)
	if err != nil {
		return nil, errorInesperado(Peticion{}, err, "el almacén de cookies de la consulta no se ha podido crear")
	}

	return &Consulta{cliente: c, emisor: nuevoClienteHTTPConCookies(c.cliente.Transport, almacen)}, nil
}

// Pedir es Cliente.Pedir dentro de la consulta: mismas comprobaciones, mismo
// ensayo, misma cadena y mismo instante; con sus cookies y sin seguir
// redirecciones.
//
// Es además lo único del módulo que emite un POST: el envío del formulario, a la
// dirección que el cliente declaró con ConFormulario y con al menos un campo.
// Cualquier otro POST y cualquier otro método son un fallo de argumentos antes
// de abrir nada, también en ensayo (FR-030, FR-031; contrato httpx-formulario
// §2).
func (q *Consulta) Pedir(ctx context.Context, ejecucion schema.Contexto, p Peticion) (Respuesta, error) {
	return q.cliente.pedir(ctx, ejecucion, p, q.cliente.formulario, q.pedirSinSeguir)
}

// pedirSinSeguir es el recorrido de una consulta: una sola petición, cuya
// respuesta se entrega también cuando es una redirección. No se pide nada a su
// destino, de modo que la consulta no va a otra dirección y el formulario no se
// reenvía a ninguna parte (FR-021; research D4 de H23).
//
// La marca de emisión no se toca: sin saltos, la que el pedido trae vacía es ya
// la de esta única petición.
func (q *Consulta) pedirSinSeguir(
	ctx context.Context, p Peticion, destino *url.URL, _ *marcaDeEmision,
) (Respuesta, *Error) {
	if fallo := terminadaAntesDePedir(ctx, p, destino); fallo != nil {
		return Respuesta{}, fallo
	}

	recibida, fallo := q.cliente.emitir(ctx, q.emisor, p, destino)
	if fallo != nil {
		return Respuesta{}, fallo
	}

	return entregarSinSeguir(p, destino, recibida)
}

// entregarSinSeguir clasifica la respuesta de una petición de una consulta. La
// única diferencia con entregar es la redirección, que aquí se entrega con su
// estado y sus cabeceras, sea o no de las que fuera de una consulta se siguen; el
// 429, el 5xx y los demás estados se clasifican igual (contrato
// httpx-formulario §4).
func entregarSinSeguir(p Peticion, direccion *url.URL, recibida recibida) (Respuesta, *Error) {
	if recibida.estado >= http.StatusMultipleChoices && recibida.estado < http.StatusBadRequest {
		return respuestaDe(p, direccion, recibida), nil
	}

	return entregar(p, direccion, recibida)
}

// ponerFormulario convierte en el envío de un formulario la petición que
// nuevaPeticionIdentificada ya construyó: le pone el cuerpo —los campos,
// codificados y ordenados por clave— y sus dos cabeceras. No construye ninguna
// petición, que sigue naciendo identificada de su único constructor (FR-009).
//
// El cuerpo se declara como lo haría la biblioteca con un lector en memoria: con
// su longitud, para que no salga troceado, y con GetBody, que deja leerlo otra
// vez sin consumir el que se envía (go doc net/http.NewRequestWithContext). No
// hay nada que cerrar en él: un escalón de la cadena que rechace la petición
// antes de entregarla al transporte no deja nada abierto.
func ponerFormulario(peticion *http.Request, campos map[string]string) {
	valores := make(url.Values, len(campos))
	for nombre, valor := range campos {
		valores.Set(nombre, valor)
	}

	cuerpo := valores.Encode()

	peticion.Body = io.NopCloser(strings.NewReader(cuerpo))
	peticion.ContentLength = int64(len(cuerpo))
	peticion.GetBody = func() (io.ReadCloser, error) {
		return io.NopCloser(strings.NewReader(cuerpo)), nil
	}

	peticion.Header.Set(cabeceraDelCuerpo, cuerpoDeUnFormulario)
	peticion.Header.Set(cabeceraDeQuienEnvia, enviadoPorLaPagina)
}
