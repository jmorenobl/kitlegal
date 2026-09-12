package httpx

import (
	"math"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"

	"github.com/jmorenobl/kitlegal/internal/core/schema"
)

// Error es el único error que el paquete produce, y el que cada ruta de fallo
// construye con la clase que le asignan FR-029, FR-030 y FR-063: ninguna queda
// sin clase. Declara esa clase él mismo —implementa schema.ConClase— de modo que
// el kernel la encuentre con errors.As y la traduzca al código de salida sin que
// este paquete importe internal/cli: la dependencia sigue yendo hacia el dominio
// y no entre adaptadores (research.md D4, contracts/errores-y-ensayo.md).
//
// Los campos son públicos porque son el dato que quien llama necesita sin
// analizar ninguna cadena de texto —Espera y EsperaConocida son el Retry-After
// de FR-030, que US2 escenario 3 obliga a entregar como dato—, y se leen del
// error recuperado con errors.As aunque venga envuelto:
//
//	var fallo *httpx.Error
//	if errors.As(err, &fallo) && fallo.EsperaConocida { /* fallo.Espera */ }
type Error struct {
	// Peticion es el método y la dirección implicados. Va vacía en los fallos de
	// configuración, que no tienen ninguna petición (FR-034).
	Peticion Peticion
	// Estado es el estado con que respondió la fuente (429, 503…) y 0 cuando no
	// hubo respuesta: ni el fallo de transporte ni el de configuración lo tienen.
	Estado int
	// Espera es lo que la cabecera Retry-After declara, 0 si no la había o no se
	// pudo leer. El cliente no la espera ni reintenta: solo la transporta (D20).
	Espera time.Duration
	// EsperaConocida distingue «la fuente dijo que volviéramos ya» de «la fuente
	// no dijo nada», que Espera a secas no puede separar.
	EsperaConocida bool
	// Causa es el error de origen —el del transporte, el del contexto, el de
	// entrada y salida al grabar—, que Unwrap expone para que errors.Is lo
	// alcance. No entra en el mensaje: el detalle técnico va al registro de
	// eventos, no al sobre (FR-045).
	Causa error

	// clase es la que el error declara al kernel, y es privada porque no se
	// negocia desde fuera: la fija el constructor de cada situación y Clase() la
	// devuelve. Una de las cuatro que el cliente produce, nunca «no encontrado»
	// ni «identidad humana» (FR-063, data-model.md §9).
	clase schema.Clase
	// motivo es qué falló, en español y sin sitio, ruta ni estado: de eso se
	// encarga Error(). Es privado y lo pone el constructor porque cada fila de la
	// tabla de clases nombra algo distinto —la opción que falta, la ruta de la
	// raíz, las dos peticiones de una colisión, el fichero de la grabación
	// ausente—, y ningún campo de los públicos podría transportarlo
	// (contracts/errores-y-ensayo.md §3).
	motivo string
}

// Error cumple la interfaz que el lenguaje espera, y con ella lo que el kernel
// presenta: el mensaje va en español, nombra el sitio y la ruta de la petición
// implicada y, cuando no hay ninguna, lo que el motivo nombre en su lugar —la
// opción que falta o que sobra— (FR-034). Lleva además el estado de la fuente
// cuando lo hubo y lo que la fuente dijo que hay que esperar cuando lo dijo, que
// es lo que hace falta para entender el fallo sin leer el registro de eventos.
//
// La causa no entra: es técnica, suele venir en inglés de la biblioteca estándar
// y su sitio es Unwrap y el registro, no el sobre (FR-045).
//
// Nunca devuelve la cadena vacía ni entra en panic, ni siquiera sobre un *Error
// nulo o construido a cero desde fuera del paquete: un sobre de fallo sin
// mensaje no le diría nada a nadie y el esquema de --describe lo prohíbe
// (FR-033, schema.DatosError).
func (e *Error) Error() string {
	if e == nil {
		return motivoSinDeclarar
	}

	mensaje := e.motivo
	if mensaje == "" {
		mensaje = motivoSinDeclarar
	}

	if contexto := e.contextoDeLaPeticion(); contexto != "" {
		mensaje = contexto + ": " + mensaje
	}

	if e.Estado != 0 {
		mensaje += " (estado " + strconv.Itoa(e.Estado) + ")"
	}

	if e.EsperaConocida {
		mensaje += "; la fuente indica que se puede volver a pedir dentro de " + e.Espera.String()
	}

	return mensaje
}

// Unwrap expone la causa, de modo que errors.Is la alcance a través del error
// del paquete y que envolver este error con %w no pierda nada de la cadena.
func (e *Error) Unwrap() error {
	if e == nil {
		return nil
	}

	return e.Causa
}

// Clase es lo que hace de este error un schema.ConClase: el propio error declara
// con qué clase hay que traducirlo a código de salida, y es el kernel quien la
// reconoce (D4).
//
// Devuelve siempre una clase del vocabulario del dominio. Un *Error nulo o
// construido a cero desde fuera no tiene clase que declarar, y lo que nadie
// declaró es justamente lo inesperado: así ninguna ruta acaba en una clase que
// el sobre no pueda llevar (FR-063).
func (e *Error) Clase() schema.Clase {
	if e == nil || e.clase == "" {
		return schema.ClaseInesperado
	}

	return e.clase
}

// Error implementa el puerto del dominio, y que lo siga implementando no depende
// de que alguien lo recuerde.
var _ schema.ConClase = (*Error)(nil)

// motivoSinDeclarar es el mensaje de un *Error que no lo produjo ningún
// constructor de este paquete —uno nulo, o uno construido a cero desde fuera—,
// para que ni el mensaje quede vacío ni leerlo entre en panic (FR-033).
const motivoSinDeclarar = "el cliente HTTP ha fallado sin declarar el motivo"

// errorDeArgumentos es la invocación o la configuración mal formada, que quien
// llama puede corregir: método distinto de GET o HEAD, dirección que no es
// absoluta ni de esquema http o https, grabación activa sin fuente o sin raíz
// declaradas, raíz o directorio de reproducción inválidos, y grabación y
// reproducción a la vez. La petición va vacía cuando el fallo es de
// configuración y no hay ninguna (filas 9 a 13 de la tabla de clases).
func errorDeArgumentos(peticion Peticion, causa error, motivo string) *Error {
	return &Error{
		Peticion: peticion,
		Causa:    causa,
		clase:    schema.ClaseArgumentos,
		motivo:   motivo,
	}
}

// errorDeFuenteNoDisponible es la fuente que no sabe entregar el recurso pedido:
// 5xx agotados los intentos, cadena de redirecciones en bucle o excedida, y el
// vencimiento o la cancelación del contexto en cualquier punto de la operación,
// también obteniendo el robots.txt o esperando turno del limitador (filas 1 a 3,
// FR-029). El estado es el del último intento, 0 si no hubo respuesta.
func errorDeFuenteNoDisponible(peticion Peticion, estado int, causa error, motivo string) *Error {
	return &Error{
		Peticion: peticion,
		Estado:   estado,
		Causa:    causa,
		clase:    schema.ClaseFuenteNoDisponible,
		motivo:   motivo,
	}
}

// errorDeLimiteOTos es el límite de peticiones o la restricción de los términos
// de uso: el robots.txt que deniega la ruta y todo lo que termina su obtención
// sin vencimiento del contexto —5xx o transporte agotados los intentos, 2xx
// ilegible, estado no previsto y cadena de redirecciones excedida— (filas 5, 6 y
// 8, FR-014, FR-015). El 429, que es la otra fuente de esta clase, tiene su
// propio constructor porque es el único que trae espera.
func errorDeLimiteOTos(peticion Peticion, estado int, causa error, motivo string) *Error {
	return &Error{
		Peticion: peticion,
		Estado:   estado,
		Causa:    causa,
		clase:    schema.ClaseLimiteOTos,
		motivo:   motivo,
	}
}

// errorDe429 es la respuesta 429, del recurso pedido o del robots.txt, que no se
// reintenta y que lleva al error lo que la fuente dijo que hay que esperar, si
// lo dijo (filas 4 y 7, FR-030, D20). El instante es el que mide la espera
// cuando la cabecera declara una fecha, y lo pone quien llama para que la cuenta
// no dependa de ningún reloj escondido.
func errorDe429(peticion Peticion, cabeceras Cabeceras, ahora time.Time, motivo string) *Error {
	espera, conocida := esperaDeRetryAfter(cabeceras, ahora)

	fallo := errorDeLimiteOTos(peticion, http.StatusTooManyRequests, nil, motivo)
	fallo.Espera = espera
	fallo.EsperaConocida = conocida

	return fallo
}

// errorInesperado es el fallo del propio mecanismo de grabación o reproducción,
// que no es de la invocación ni de la fuente: escritura que falla después de
// validar la raíz, colisión de nombres con otra petición, y grabación ausente,
// ajena, ilegible o de otro formato (filas 14 a 16). Se declara así a propósito,
// para que un fixture ausente o cambiado no se confunda con la respuesta que el
// test esperaba (FR-063).
func errorInesperado(peticion Peticion, causa error, motivo string) *Error {
	return &Error{
		Peticion: peticion,
		Causa:    causa,
		clase:    schema.ClaseInesperado,
		motivo:   motivo,
	}
}

// contextoDeLaPeticion nombra la petición implicada: el método, la ruta pedida
// —con su consulta, que es lo que de verdad se pidió— y el sitio, que es la
// unidad del robots.txt y del ritmo y por tanto lo que hace falta para entender
// el fallo (FR-034).
//
// Devuelve la cadena vacía cuando no hay petición, y entonces el mensaje es solo
// el motivo, que es el que nombra la opción que falta o que sobra. Cuando la
// dirección es precisamente lo que falla —no se puede interpretar, no es
// absoluta o su esquema no es de red (fila 10)— no hay sitio ni ruta que
// derivar: se nombra tal como la escribió quien llamó, sin inventarle a ningún
// esquema un puerto que no tiene.
func (e *Error) contextoDeLaPeticion() string {
	if e.Peticion.URL == "" {
		return ""
	}

	comoLlego := strings.TrimSpace(e.Peticion.Metodo + " " + e.Peticion.URL)

	direccion, err := url.Parse(e.Peticion.URL)
	if err != nil || direccion.Host == "" {
		return comoLlego
	}

	if direccion.Scheme != "http" && direccion.Scheme != "https" {
		return comoLlego
	}

	return strings.TrimSpace(e.Peticion.Metodo+" "+direccion.RequestURI()) +
		" en el sitio " + claveDeSitio(direccion)
}

// esperaDeRetryAfter lee la cabecera Retry-After en las dos formas que HTTP
// admite: un entero de segundos, o una fecha —cualquiera de los tres formatos de
// HTTP/1.1, que es lo que http.ParseTime acepta—, de la que se toma lo que queda
// hasta ella con mínimo cero (D20).
//
// El segundo valor dice si la cabecera estaba y se pudo leer, y es lo que
// distingue «volver ya» de «la fuente no dijo nada». Un valor ilegible equivale
// a su ausencia y no cambia la clase del fallo, que sigue siendo la del 429
// (FR-030).
func esperaDeRetryAfter(cabeceras Cabeceras, ahora time.Time) (time.Duration, bool) {
	valor := strings.TrimSpace(cabeceras.Get("Retry-After"))
	if valor == "" {
		return 0, false
	}

	if segundos, err := strconv.ParseInt(valor, 10, 64); err == nil {
		return enSegundos(segundos), true
	}

	fecha, err := http.ParseTime(valor)
	if err != nil {
		return 0, false
	}

	if espera := fecha.Sub(ahora); espera > 0 {
		return espera, true
	}

	return 0, true
}

// enSegundos convierte en duración los segundos que declara Retry-After, con
// mínimo cero y con el tope de lo que cabe en una duración: un valor enorme
// —roto o malintencionado— no puede desbordarse en silencio y salir como una
// espera negativa.
func enSegundos(segundos int64) time.Duration {
	if segundos <= 0 {
		return 0
	}

	if segundos > segundosMaximos {
		return time.Duration(math.MaxInt64)
	}

	return time.Duration(segundos) * time.Second
}

// segundosMaximos son los segundos que caben en una time.Duration, que los mide
// en nanosegundos: unos 292 años.
const segundosMaximos = int64(math.MaxInt64 / int64(time.Second))
