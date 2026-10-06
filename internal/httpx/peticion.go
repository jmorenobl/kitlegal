package httpx

import (
	"net/http"
	"time"
)

// Peticion es lo que un adaptador de fuente pide: un método, una dirección, el
// formato en que quiere el recurso y, solo cuando envía un formulario de
// consulta, sus campos; nada más. No lleva un juego de cabeceras porque la que no
// se negocia no es suya —la identificación la pone el paquete en toda petición
// (FR-009)— y la única que una fuente necesita elegir es la del formato (research
// D4 de H4); ni un cuerpo libre, porque el único que sale del paquete es el de
// los campos de un formulario declarado, y lo escribe él (research D2 de H23).
// Todo se valida antes de abrir nada, y un valor inválido es un fallo de la clase
// «argumentos» (data-model.md §2).
type Peticion struct {
	// Metodo es «GET» o «HEAD» (FR-010) o, solo desde una Consulta y hacia la
	// dirección que su cliente declaró con ConFormulario, «POST» (contrato
	// httpx-formulario §2 de H23).
	Metodo string
	// URL es la dirección absoluta, de esquema http o https (D19).
	URL string
	// Acepta es el tipo de contenido que se pide en la cabecera Accept
	// —«application/xml», «application/json»—, en la petición y en cada salto de
	// su cadena de redirecciones. Vacío, la petición no lleva Accept (contrato
	// httpx-acepta-e-instante §1 de H4).
	Acepta string
	// Campos son los campos del formulario que se envía: solo con «POST», y
	// entonces al menos uno. Salen en el cuerpo, codificados y ordenados por
	// clave; con «GET» o «HEAD» van vacíos (contrato httpx-formulario §2 y §3 de
	// H23).
	Campos map[string]string
}

// Respuesta es lo que el paquete devuelve cuando la petición no falla. Es un
// tipo propio y no un *http.Response, para que un adaptador de fuente pueda
// leer el estado, las cabeceras y el cuerpo sin importar la biblioteca HTTP
// (FR-002, R2, D2).
type Respuesta struct {
	// Peticion es la que se pidió, no la del último salto de la cadena de
	// redirecciones.
	Peticion Peticion
	// URL es la dirección final, la que de verdad entregó el contenido: es la
	// que el sobre tiene que citar cuando hubo redirecciones (FR-011).
	URL string
	// Estado es el estado HTTP de la respuesta final. Nunca es 429 ni 5xx, que
	// terminan en un fallo con su clase, ni 3xx fuera de una Consulta (FR-029,
	// FR-030, FR-032). Dentro de una Consulta ninguna redirección se sigue, y la
	// respuesta 3xx se entrega con su estado (contrato httpx-formulario §4 de
	// H23).
	Estado int
	// Cabeceras son las de la respuesta final, con los nombres canónicos.
	Cabeceras Cabeceras
	// Cuerpo es el cuerpo entero, ya leído dentro del paquete; vacío en HEAD.
	// No se entrega en streaming, para que cerrarlo no quede nunca en manos de
	// quien llama (D2).
	Cuerpo []byte
	// Ensayo es verdadero solo bajo --dry-run, y entonces la petición no se ha
	// emitido: Estado es 0 y Cuerpo es nil (FR-065).
	Ensayo bool
	// Instante es el de emisión de la petición que entregó esta respuesta —la
	// del último intento del último salto—, con la hora del cliente tomada justo
	// antes de entregarla al transporte; en reproducción, al servir la grabación.
	// Es la fecha de consulta que la fuente declara (FR-096). Cero en ensayo,
	// donde no se emitió nada (contrato httpx-acepta-e-instante §2 de H4).
	Instante time.Time
}

// Descripcion es la línea «<Metodo> <URL>» de la petición pedida: la que un
// adaptador copia en schema.Resultado.Ensayo para que el kernel la presente
// bajo --dry-run. Describe lo que se pidió, no la dirección final, porque en
// ensayo no se ha emitido nada y no hay ninguna otra (D6, FR-065).
func (r Respuesta) Descripcion() string {
	return r.Peticion.Metodo + " " + r.Peticion.URL
}

// Cabeceras son las cabeceras de una respuesta, indexadas por su nombre
// canónico. Es un tipo del paquete y no un http.Header para que nombrarlo no
// obligue a importar la biblioteca HTTP (R2, D2).
type Cabeceras map[string][]string

// Get devuelve el primer valor de la cabecera, o la cadena vacía si no está o
// no trae ninguno. Canonicaliza el nombre igual que la biblioteca HTTP, de modo
// que «content-type», «Content-Type» y «CONTENT-TYPE» son la misma cabecera.
func (c Cabeceras) Get(nombre string) string {
	valores := c[http.CanonicalHeaderKey(nombre)]
	if len(valores) == 0 {
		return ""
	}

	return valores[0]
}
