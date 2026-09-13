package boe

import (
	"errors"
	"time"

	"github.com/jmorenobl/kitlegal/internal/core/schema"
)

// Error es el error con clase del adaptador, y el que cada ruta de fallo de la
// fuente construye con la clase que le asigna la tabla cerrada del contrato
// errores-y-codigos §1: ninguna queda sin clase y ninguna es la de identidad
// humana (FR-100). Declara esa clase él mismo —implementa schema.ConClase, como
// httpx.Error y cache.Error— de modo que el kernel la encuentre con errors.As
// aunque vaya envuelto con %w y la traduzca al código de salida sin que este
// paquete importe internal/cli (research.md D10).
//
// Los campos son públicos porque son lo que el applet necesita para fechar y
// firmar el sobre de fallo sin analizar ningún texto, y se leen del error
// recuperado con errors.As aunque venga envuelto:
//
//	var fallo *boe.Error
//	if errors.As(err, &fallo) { /* fallo.URL, fallo.Instante */ }
type Error struct {
	// URL es la dirección de la API cuya consulta falló: la de la petición que
	// falló, o la del recurso que falta o que no se pudo leer de la caché, y
	// nunca la del robots.txt (FR-101). Va vacía cuando el fallo es de los
	// argumentos, que no llegan a construir ninguna dirección (filas 2 a 4).
	URL string
	// Instante es el de emisión de la petición que falló, el que entrega httpx
	// (FR-096). Va a cero cuando no hubo petición —argumentos, --offline, caché—,
	// y entonces la fecha del sobre la pone el montaje del kernel.
	Instante time.Time
	// Causa es el error de origen, que Unwrap expone para que errors.Is y
	// errors.As lo alcancen. Solo entra en el mensaje si declara su propia clase,
	// como el de httpx o el de la caché: esos van escritos para la persona y el
	// contrato §2 obliga a conservar su texto. El detalle técnico de cualquier
	// otra —el del analizador de XML o de JSON— va al registro, no al sobre.
	Causa error

	// clase es la que el error declara al kernel, y es privada porque no se
	// negocia desde fuera: la fija el constructor de cada clase y Clase() la
	// devuelve.
	clase schema.Clase
	// motivo es qué falló, en español y ya compuesto por quien construye el
	// error, porque cada fila del contrato §2 nombra algo distinto —el valor y la
	// forma esperada, la norma y el bloque, el estado, la clave— y ningún campo
	// público alcanza a transportarlo.
	motivo string
}

// Error cumple la interfaz que el lenguaje espera, y con ella lo que el kernel
// presenta. El mensaje es el motivo; si hay dirección, la dirección entre
// paréntesis detrás, para que los fallos de las filas que la exigen la nombren
// sin que ningún motivo tenga que repetirla; y si la causa declara su clase, su
// texto detrás de dos puntos. El instante no entra: es dato del sobre, no del
// mensaje.
//
// Nunca devuelve la cadena vacía ni entra en panic, ni siquiera sobre un *Error
// nulo o construido a cero desde fuera del paquete: un sobre de fallo sin
// mensaje no le diría nada a nadie y el esquema de --describe lo prohíbe
// (schema.DatosError).
func (e *Error) Error() string {
	if e == nil {
		return motivoSinDeclarar
	}

	mensaje := e.motivo
	if mensaje == "" {
		mensaje = motivoSinDeclarar
	}

	if e.URL != "" {
		mensaje += " (" + e.URL + ")"
	}

	if texto := textoParaLaPersona(e.Causa); texto != "" {
		mensaje += ": " + texto
	}

	return mensaje
}

// Unwrap expone la causa, de modo que errors.Is y errors.As la alcancen a
// través del error del adaptador: el de la caché, por ejemplo, sigue siendo
// alcanzable con errors.Is después de envolverlo.
func (e *Error) Unwrap() error {
	if e == nil {
		return nil
	}

	return e.Causa
}

// Clase es lo que hace de este error un schema.ConClase: el propio error
// declara con qué clase hay que traducirlo a código de salida, y es el kernel
// quien la reconoce y quien decide el código. Este paquete no decide ninguno.
//
// Devuelve siempre una clase del vocabulario del dominio. Un *Error nulo o
// construido a cero desde fuera no tiene clase que declarar, y lo que nadie
// declaró es justamente lo inesperado.
func (e *Error) Clase() schema.Clase {
	if e == nil || e.clase == "" {
		return schema.ClaseInesperado
	}

	return e.clase
}

// Error implementa el puerto del dominio, y que lo siga implementando no depende
// de que alguien lo recuerde.
var _ schema.ConClase = (*Error)(nil)

// motivoSinDeclarar es el mensaje de un *Error que no produjo ningún constructor
// de este paquete —uno nulo, o uno construido a cero desde fuera—, para que ni
// el mensaje quede vacío ni leerlo entre en panic.
const motivoSinDeclarar = "el adaptador del BOE ha fallado sin declarar el motivo"

// textoParaLaPersona es el texto de la causa que entra en el mensaje: el entero,
// con el contexto que la envuelva, cuando en su cadena hay un error que declara
// su clase, y ninguno en otro caso. Un error con clase es de un adaptador del
// proyecto, cuyo contrato obliga a escribirlo en español y para la persona; uno
// sin ella es técnico y su sitio es el registro.
func textoParaLaPersona(causa error) string {
	var conClase schema.ConClase
	if !errors.As(causa, &conClase) {
		return ""
	}

	return causa.Error()
}

// errorDeArgumentos es la invocación mal formada, que quien invoca puede
// corregir: la norma o el bloque fuera de su gramática y la búsqueda sin ninguna
// palabra (filas 2 a 4), y el fallo de la caché que ella declara de esta clase
// (fila 21). Código de salida 2.
func errorDeArgumentos(direccion string, instante time.Time, causa error, motivo string) *Error {
	return nuevoError(schema.ClaseArgumentos, direccion, instante, causa, motivo)
}

// errorDeNoEncontrado es lo pedido que la fuente dice que no existe: el 404 de
// un bloque, de un índice, de unos metadatos o de un análisis, y la respuesta
// con data vacío de los tres últimos (filas 6 a 8). Una búsqueda nunca lo es.
// Código de salida 3.
func errorDeNoEncontrado(direccion string, instante time.Time, causa error, motivo string) *Error {
	return nuevoError(schema.ClaseNoEncontrado, direccion, instante, causa, motivo)
}

// errorDeFuenteNoDisponible es la fuente que no entrega lo pedido en una forma
// que se pueda usar: la entrada ausente o caducada con --offline, el 404 de una
// búsqueda y cualquier otro estado no 2xx, el cuerpo que no se puede interpretar,
// el fallo de httpx de esta clase y el de la caché que ella declara así (filas
// 5, 9 a 13, 16, 17 y 21). Código de salida 4.
func errorDeFuenteNoDisponible(direccion string, instante time.Time, causa error, motivo string) *Error {
	return nuevoError(schema.ClaseFuenteNoDisponible, direccion, instante, causa, motivo)
}

// errorDeLimiteOTos es el límite de peticiones o la restricción de los términos
// de uso, que el adaptador no produce por su cuenta sino que recibe de httpx: el
// 429 y el robots.txt que deniega la ruta o que no se pudo obtener (filas 14 y
// 15). Código de salida 5.
func errorDeLimiteOTos(direccion string, instante time.Time, causa error, motivo string) *Error {
	return nuevoError(schema.ClaseLimiteOTos, direccion, instante, causa, motivo)
}

// errorInesperado es lo que no debería haber pasado y no arregla quien invoca:
// la entrada de caché con la clave correcta que no se puede leer, el fallo de la
// caché que ella declara de esta clase y el defecto al componer la fuente (filas
// 20 a 22). Código de salida 1.
func errorInesperado(direccion string, instante time.Time, causa error, motivo string) *Error {
	return nuevoError(schema.ClaseInesperado, direccion, instante, causa, motivo)
}

// nuevoError es lo que comparten los cinco constructores: todo menos la clase,
// que es lo único que los distingue y lo que cada uno fija.
func nuevoError(clase schema.Clase, direccion string, instante time.Time, causa error, motivo string) *Error {
	return &Error{
		URL:      direccion,
		Instante: instante,
		Causa:    causa,
		clase:    clase,
		motivo:   motivo,
	}
}
