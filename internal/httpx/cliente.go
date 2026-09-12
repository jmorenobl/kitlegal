package httpx

import (
	"context"
	"io"
	"log/slog"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"

	"github.com/jmorenobl/kitlegal/internal/core/schema"
)

// topeDeRedirecciones es el número de saltos que el cliente sigue antes de
// declarar que la fuente no sabe entregar el recurso. Es el mismo de la política
// por omisión de la biblioteca —«stop after 10 consecutive requests» (go doc
// net/http.Client)— y queda por encima de los cinco que RFC 9309 §2.3.1.2 pide
// como mínimo para el robots.txt, que reutiliza este tope (D10).
const topeDeRedirecciones = 10

// esquemasDeRed son los dos únicos esquemas que una dirección puede traer. Lo
// que no es uno de ellos no es una fuente caída sino una invocación que quien
// llama puede corregir (research D19).
const (
	esquemaHTTP  = "http"
	esquemaHTTPS = "https"
)

// Cliente es el único objeto del módulo capaz de emitir una petición HTTP, y
// Pedir su única operación: no hay ninguna otra forma de salir a la red desde
// este paquete, ni forma alguna de construir uno al que le falte una de sus
// garantías. Sus campos son privados porque ninguna de ellas se negocia desde
// fuera (FR-001, FR-002, D1).
//
// Es seguro para uso concurrente: no guarda nada de una petición a la siguiente.
type Cliente struct {
	// cliente ejecuta la cadena de decoradores, que se compone una vez en New y
	// no cambia después (D3).
	cliente *http.Client
	// fuente es el nombre lógico de la fuente que usa este cliente, el que
	// nombrará el directorio de las grabaciones (FR-039).
	fuente string
	// registrador es el destino de los eventos. Nunca es nulo: sin la opción,
	// descarta lo que se le entregue (FR-035, D15).
	registrador *slog.Logger
}

// configuracionDelCliente es lo que las opciones rellenan antes de que New
// construya el cliente. No se exporta: lo que se declara son las opciones, no la
// estructura que escriben.
type configuracionDelCliente struct {
	fuente      string
	registrador *slog.Logger
}

// Opcion declara una variación del cliente. Devuelve el error de la opción
// inválida, pero quien lo entrega es el constructor: validar en la construcción
// y no en cada uso es lo que hace que un cliente exista solo si es válido
// (patrón de golang-design-patterns; contrato §2).
type Opcion func(*configuracionDelCliente) error

// ConFuente declara el nombre lógico de la fuente —«boe», «placsp»…—, que es el
// que nombra el directorio de las grabaciones. Un nombre vacío o que pudiera
// escaparse de ese directorio es un error de argumentos (FR-039).
func ConFuente(nombre string) Opcion {
	return func(config *configuracionDelCliente) error {
		if nombre == "" {
			return errorDeArgumentos(Peticion{}, nil,
				"el nombre de la fuente no puede ir vacío (ConFuente)")
		}

		if strings.ContainsAny(nombre, `/\`) || strings.Contains(nombre, "..") {
			return errorDeArgumentos(Peticion{}, nil,
				`el nombre de la fuente no puede llevar «/», «\» ni «..» (ConFuente): `+nombre)
		}

		config.fuente = nombre

		return nil
	}
}

// ConRegistrador declara el destino de los eventos del cliente, que es el mismo
// registrador que el kernel entrega al applet. Sin esta opción los eventos se
// descartan; nunca se emiten por el registrador global ni por la salida estándar
// (FR-035, D15).
func ConRegistrador(registrador *slog.Logger) Opcion {
	return func(config *configuracionDelCliente) error {
		if registrador == nil {
			return errorDeArgumentos(Peticion{}, nil,
				"el registrador de eventos no puede ser nulo (ConRegistrador)")
		}

		config.registrador = registrador

		return nil
	}
}

// New construye el cliente contra la red, con sus garantías ya puestas sin
// declarar ninguna opción: identificación en toda petición, plazo del contexto
// y redirecciones seguidas por él mismo. Las opciones se aplican en orden —la última
// repetida gana— y la primera inválida termina la construcción con su clase
// (FR-001, contrato §2 y §3).
func New(opciones ...Opcion) (*Cliente, error) {
	config := configuracionDelCliente{registrador: slog.New(slog.DiscardHandler)}

	for _, opcion := range opciones {
		if err := opcion(&config); err != nil {
			return nil, err
		}
	}

	cadena := conIdentificacion(nuevoTransporte())

	return &Cliente{
		cliente:     nuevoClienteHTTP(cadena),
		fuente:      config.fuente,
		registrador: config.registrador,
	}, nil
}

// Pedir es la única operación de red del módulo. Exige el contexto de
// cancelación como primer parámetro —el plazo de la operación es el suyo y solo
// el suyo (FR-003, FR-004)— y el contexto de ejecución del kernel como segundo,
// que es lo que hace que --dry-run llegue hasta aquí sin que quien llama pueda
// olvidarlo (FR-050, D1).
//
// Fuera de la cadena de decoradores viven las tres cosas que necesitan decidir
// antes o después de ella: la comprobación del método y de la dirección, que
// ocurre antes de abrir nada; el ensayo, que devuelve sin bajar por la cadena; y
// el bucle de redirecciones con la clasificación del resultado (D3).
func (c *Cliente) Pedir(ctx context.Context, ejecucion schema.Contexto, p Peticion) (Respuesta, error) {
	direccion, err := comprobarPeticion(p)
	if err != nil {
		return Respuesta{}, err
	}

	if ejecucion.DryRun {
		c.registrador.DebugContext(ctx, "ensayo: la petición no se emite",
			"metodo", p.Metodo, "url", p.URL)

		return Respuesta{Peticion: p, URL: p.URL, Ensayo: true}, nil
	}

	return c.seguirLaCadena(ctx, p, direccion)
}

// comprobarPeticion aplica las dos comprobaciones que no necesitan red y que por
// eso rigen también en ensayo: el método es GET o HEAD (FR-010) y la dirección
// es absoluta, de esquema de red y con sitio (research D19). Las dos son de la
// clase «argumentos» y ninguna llega a abrir una conexión.
func comprobarPeticion(p Peticion) (*url.URL, error) {
	if p.Metodo != http.MethodGet && p.Metodo != http.MethodHead {
		return nil, errorDeArgumentos(p, nil, "el método de una petición solo puede ser GET o HEAD")
	}

	direccion, err := url.Parse(p.URL)
	if err != nil {
		return nil, errorDeArgumentos(p, err, "la dirección no se puede interpretar")
	}

	if direccion.Scheme != esquemaHTTP && direccion.Scheme != esquemaHTTPS {
		return nil, errorDeArgumentos(p, nil, "la dirección tiene que ser absoluta y de esquema http o https")
	}

	if direccion.Host == "" {
		return nil, errorDeArgumentos(p, nil, "la dirección no nombra ningún sitio")
	}

	return direccion, nil
}

// seguirLaCadena es el bucle de redirecciones: cada salto es una petición
// completa que baja por la cadena entera, de modo que el destino no hereda nada
// del origen (FR-011, D10). Termina cuando la respuesta ya no es una redirección
// seguible, cuando una dirección se repite o cuando se agota el tope; en los dos
// últimos casos, con la clase de quien no sabe entregar el recurso.
func (c *Cliente) seguirLaCadena(ctx context.Context, p Peticion, destino *url.URL) (Respuesta, error) {
	visitadas := make(map[string]struct{}, topeDeRedirecciones+1)

	for salto := 0; salto <= topeDeRedirecciones; salto++ {
		if _, repetida := visitadas[destino.String()]; repetida {
			return Respuesta{}, errorDeFuenteNoDisponible(p, 0, nil,
				"la cadena de redirecciones vuelve sobre una dirección ya visitada: "+destino.String())
		}

		visitadas[destino.String()] = struct{}{}

		recibida, err := c.emitir(ctx, p, destino)
		if err != nil {
			return Respuesta{}, err
		}

		if !esRedireccionSeguible(recibida.estado) {
			return entregar(p, destino, recibida)
		}

		siguiente, err := siguienteDestino(p, destino, recibida)
		if err != nil {
			return Respuesta{}, err
		}

		c.registrador.DebugContext(ctx, "redirección seguida",
			"de", destino.String(), "a", siguiente.String(), "estado", recibida.estado)

		destino = siguiente
	}

	return Respuesta{}, errorDeFuenteNoDisponible(p, 0, nil,
		"la cadena de redirecciones supera los "+strconv.Itoa(topeDeRedirecciones)+" saltos")
}

// recibida es lo que emitir devuelve: la respuesta con el cuerpo ya leído y el
// *http.Response ya cerrado. El único de todo el módulo se abre y se cierra
// aquí dentro, que es lo que impide que cerrarlo quede nunca en manos de quien
// llama (FR-002, D2).
type recibida struct {
	estado    int
	cabeceras Cabeceras
	cuerpo    []byte
}

// emitir baja una petición por la cadena y devuelve lo que la fuente respondió.
// La petición la construye nuevaPeticionIdentificada, que es el único
// constructor del paquete: así nace identificada aunque la origine el bucle y no
// quien llama (FR-009, D3).
func (c *Cliente) emitir(ctx context.Context, p Peticion, destino *url.URL) (recibida, error) {
	peticion, err := nuevaPeticionIdentificada(ctx, p.Metodo, destino.String())
	if err != nil {
		return recibida{}, errorDeArgumentos(p, err, "la dirección no se puede pedir")
	}

	comienzo := time.Now()

	respuesta, err := c.cliente.Do(peticion)
	if err != nil {
		return recibida{}, falloAlEmitir(ctx, p, err)
	}

	cuerpo, errDeLectura := io.ReadAll(respuesta.Body)

	if errAlCerrar := respuesta.Body.Close(); errAlCerrar != nil {
		c.registrador.DebugContext(ctx, "el cuerpo de la respuesta no se ha cerrado limpiamente",
			"url", destino.String(), "error", errAlCerrar.Error())
	}

	if errDeLectura != nil {
		return recibida{}, errorDeFuenteNoDisponible(p, respuesta.StatusCode, errDeLectura,
			"el cuerpo de la respuesta no se ha podido leer entero")
	}

	c.registrador.DebugContext(ctx, "petición emitida",
		"metodo", p.Metodo, "url", destino.String(),
		"estado", respuesta.StatusCode, "duracion", time.Since(comienzo))

	return recibida{
		estado:    respuesta.StatusCode,
		cabeceras: Cabeceras(respuesta.Header),
		cuerpo:    cuerpo,
	}, nil
}

// falloAlEmitir distingue las dos formas en que una petición puede no llegar a
// respuesta. Las dos son de la misma clase —la fuente no sabe entregar el
// recurso—, pero no son lo mismo para quien lee el mensaje: el contexto que
// vence o se cancela corta la operación en ese instante (FR-005, FR-029) y el
// fallo del transporte es el sitio que no responde.
func falloAlEmitir(ctx context.Context, p Peticion, causa error) error {
	if ctx.Err() != nil {
		return errorDeFuenteNoDisponible(p, 0, causa,
			"la operación ha terminado antes de recibir la respuesta")
	}

	return errorDeFuenteNoDisponible(p, 0, causa, "no se ha podido conectar con el sitio")
}

// esRedireccionSeguible dice si el estado es uno de los cinco que el cliente
// sigue, que son los mismos que sigue la biblioteca (go doc net/http.Client.Get).
// Cualquier otro 3xx —300, 304, 305— no se puede seguir y tampoco se entrega
// (FR-032, D10).
func esRedireccionSeguible(estado int) bool {
	switch estado {
	case http.StatusMovedPermanently, http.StatusFound, http.StatusSeeOther,
		http.StatusTemporaryRedirect, http.StatusPermanentRedirect:
		return true
	default:
		return false
	}
}

// siguienteDestino resuelve la dirección del salto sobre la del salto actual,
// que es lo que hace que una Location relativa funcione (RFC 3986 §5.2). Una
// redirección que no dice a dónde ir, o que lo dice con una dirección
// ininterpretable, no se puede seguir: es la fuente la que no sabe entregar el
// recurso (FR-011, D10).
func siguienteDestino(p Peticion, actual *url.URL, recibida recibida) (*url.URL, error) {
	destino := recibida.cabeceras.Get("Location")
	if destino == "" {
		return nil, errorDeFuenteNoDisponible(p, recibida.estado, nil,
			"la redirección no trae la cabecera Location y no dice a dónde ir")
	}

	referencia, err := url.Parse(destino)
	if err != nil {
		return nil, errorDeFuenteNoDisponible(p, recibida.estado, err,
			"la redirección apunta a una dirección que no se puede interpretar: "+destino)
	}

	return actual.ResolveReference(referencia), nil
}

// entregar clasifica la respuesta final de la cadena: el 429 y el 5xx tienen su
// clase, un 3xx que no se ha podido seguir no es entregable, y todo lo demás
// —el «no encontrado» incluido— se entrega a quien llama, que es quien sabe qué
// significa en su fuente (FR-029, FR-030, FR-032; tabla del contrato §3).
func entregar(p Peticion, direccion *url.URL, recibida recibida) (Respuesta, error) {
	switch {
	case recibida.estado == http.StatusTooManyRequests:
		return Respuesta{}, errorDe429(p, recibida.cabeceras, time.Now(),
			"la fuente ha alcanzado su límite de peticiones")

	case recibida.estado >= http.StatusInternalServerError:
		return Respuesta{}, errorDeFuenteNoDisponible(p, recibida.estado, nil,
			"el sitio no ha sabido entregar el recurso")

	case recibida.estado >= http.StatusMultipleChoices && recibida.estado < http.StatusBadRequest:
		return Respuesta{}, errorDeFuenteNoDisponible(p, recibida.estado, nil,
			"el sitio ha respondido con una redirección que no se puede seguir")

	default:
		return Respuesta{
			Peticion:  p,
			URL:       direccion.String(),
			Estado:    recibida.estado,
			Cabeceras: recibida.cabeceras,
			Cuerpo:    recibida.cuerpo,
		}, nil
	}
}
