package cache

import (
	"errors"
	"fmt"
	"io/fs"
	"time"

	"github.com/jmorenobl/kitlegal/internal/core/schema"
)

// Error es el único error que el paquete produce, y el que cada ruta de fallo
// construye con la clase que le asigna la tabla cerrada de
// contracts/errores-y-codigos.md §3: ninguna queda sin clase (FR-033). Declara
// esa clase él mismo —implementa schema.ConClase— de modo que el kernel la
// encuentre con errors.As y la traduzca al código de salida sin que este
// paquete importe internal/cli: la dependencia sigue yendo hacia el dominio y
// no entre adaptadores (research.md D9).
//
// Los campos son públicos porque son el dato que quien llama necesita sin
// analizar ninguna cadena de texto, y se leen del error recuperado con
// errors.As aunque venga envuelto:
//
//	var fallo *cache.Error
//	if errors.As(err, &fallo) && fallo.Clase() == schema.ClaseFuenteNoDisponible { /* --offline sin entrada */ }
type Error struct {
	// Operacion es qué se estaba haciendo, del vocabulario cerrado
	// «construir», «migrar», «leer», «escribir» y «cerrar» (data-model.md §8).
	Operacion string
	// Ruta es el fichero o el directorio implicado, y va vacía cuando no lo hay
	// —una clave vacía o una vigencia inválida no tienen ninguno—.
	Ruta string
	// Origen es de dónde salió la ruta, del vocabulario cerrado «opción
	// ConDirectorio», «variable KITLEGAL_CACHE_DIR» y «ruta por omisión», y va
	// vacío cuando el fallo no es de ruta. Es lo que distingue «la opción que
	// trae el código está mal» de «lo que hay en el entorno está mal», y por eso
	// el mensaje lo nombra siempre que existe (FR-022, FR-035).
	Origen string
	// Clave es la clave implicada, y va vacía cuando no hay ninguna: la tienen
	// los fallos de lectura y de escritura, no los de apertura.
	Clave string
	// Causa es el error de origen —el del controlador, el del sistema de
	// ficheros, el del contexto—, que Unwrap expone para que errors.Is lo
	// alcance. No entra en el mensaje: el detalle técnico va al registro de
	// eventos, no al sobre (FR-035).
	Causa error

	// clase es la que el error declara al kernel, y es privada porque no se
	// negocia desde fuera: la fija el constructor de cada situación y Clase() la
	// devuelve. Una de las tres que la caché produce —«argumentos», «fuente no
	// disponible» e «inesperado»—, nunca «no encontrado», «límite o TOS» ni
	// «identidad humana» (FR-033).
	clase schema.Clase
	// motivo es qué falló, en español y ya compuesto, y es privado y lo pone el
	// constructor de la situación porque cada fila de la tabla nombra algo
	// distinto —el origen y la ruta, las dos versiones del esquema, la clave y
	// que solo se lee, los dos ficheros auxiliares del registro de escritura— y
	// los campos públicos no alcanzan a transportarlo todo
	// (contracts/errores-y-codigos.md §6).
	motivo string
}

// Error cumple la interfaz que el lenguaje espera, y con ella lo que el kernel
// presenta: el mensaje va en español, lo encabeza el nombre de la caché para
// que se sepa de dónde viene sin leer ninguna traza, y nombra lo que la columna
// «Mensaje nombra» de la tabla del contrato §3 exige en cada situación
// (FR-035).
//
// La causa no entra: es técnica, suele venir en inglés de la biblioteca
// estándar y su sitio es Unwrap y el registro, no el sobre.
//
// Nunca devuelve la cadena vacía ni entra en panic, ni siquiera sobre un *Error
// nulo o construido a cero desde fuera del paquete: un sobre de fallo sin
// mensaje no le diría nada a nadie y el esquema de --describe lo prohíbe
// (FR-034, schema.DatosError).
func (e *Error) Error() string {
	if e == nil || e.motivo == "" {
		return encabezado + motivoSinDeclarar
	}

	return encabezado + e.motivo
}

// Unwrap expone la causa, de modo que errors.Is la alcance a través del error
// del paquete y que envolver este error con %w no pierda nada de la cadena.
func (e *Error) Unwrap() error {
	if e == nil {
		return nil
	}

	return e.Causa
}

// Clase es lo que hace de este error un schema.ConClase: el propio error
// declara con qué clase hay que traducirlo a código de salida, y es el kernel
// quien la reconoce y quien decide el código. Este paquete no decide ninguno
// (FR-033, D9).
//
// Devuelve siempre una clase del vocabulario del dominio. Un *Error nulo o
// construido a cero desde fuera no tiene clase que declarar, y lo que nadie
// declaró es justamente lo inesperado: así ninguna ruta acaba en una clase que
// el sobre no pueda llevar.
func (e *Error) Clase() schema.Clase {
	if e == nil || e.clase == "" {
		return schema.ClaseInesperado
	}

	return e.clase
}

// Error implementa el puerto del dominio, y que lo siga implementando no
// depende de que alguien lo recuerde.
var _ schema.ConClase = (*Error)(nil)

const (
	// encabezado dice de quién es el fallo. Va delante de todo mensaje, también
	// del de un *Error que no construyó este paquete.
	encabezado = "caché: "
	// motivoSinDeclarar es el mensaje de un *Error que no produjo ningún
	// constructor de este paquete —uno nulo, o uno construido a cero desde
	// fuera—, para que ni el mensaje quede vacío ni leerlo entre en panic.
	motivoSinDeclarar = "ha fallado sin declarar el motivo"
	// sufijoRegistroDeEscritura y sufijoMemoriaCompartida son los dos ficheros
	// que SQLite pone junto a la base cuando el diario va en WAL. Solo se
	// nombran en el mensaje de la fila 13 del contrato §3, que es donde su
	// presencia y su ausencia explican por qué no se puede leer.
	sufijoRegistroDeEscritura = "-wal"
	sufijoMemoriaCompartida   = "-shm"
)

// errorDeArgumentos es la invocación mal formada, que quien llama puede
// corregir: clave vacía, vigencia menor o igual que cero, opción inválida,
// variable de entorno declarada y vacía, ruta inservible, directorio o fichero
// que en modo normal no se puede crear ni escribir, y ruta por omisión
// indeterminable (filas 1 a 7 de la tabla del contrato §3; código de salida 2).
func errorDeArgumentos(operacion, motivo string, causa error) *Error {
	return &Error{
		Operacion: operacion,
		Causa:     causa,
		clase:     schema.ClaseArgumentos,
		motivo:    motivo,
	}
}

// errorDeFuenteNoDisponible es lo que no está y no se puede traer: la ausencia
// de entrada vigente cuando la invocación es de solo lectura, y el contexto
// cancelado o vencido durante cualquier operación, que es la clase que H1
// reserva a la fuente que no responde y al plazo agotado (filas 8 y 15; código
// de salida 4).
func errorDeFuenteNoDisponible(operacion, motivo string, causa error) *Error {
	return &Error{
		Operacion: operacion,
		Causa:     causa,
		clase:     schema.ClaseFuenteNoDisponible,
		motivo:    motivo,
	}
}

// errorInesperado es lo que no debería haber pasado y no arregla quien invoca:
// esquema de otra versión, fichero que no es una base utilizable, escritura
// intentada en solo lectura, acceso denegado que impide saber siquiera si la
// base está o leerla, registro de escritura sin memoria compartida, bloqueo
// que dura más que la espera, fallo sobrevenido del sistema de ficheros o del
// controlador, y operación después de cerrar (filas 9 a 14; código de salida 1).
func errorInesperado(operacion, motivo string, causa error) *Error {
	return &Error{
		Operacion: operacion,
		Causa:     causa,
		clase:     schema.ClaseInesperado,
		motivo:    motivo,
	}
}

// errorDeRutaInservible es la fila 5: la ruta de la caché existe y no es un
// directorio, venga de la opción o de la variable, y en cualquier modo. El
// mensaje nombra origen y ruta porque sin las dos cosas no se sabe qué hay que
// corregir ni dónde (FR-022).
func errorDeRutaInservible(operacion, origen, ruta string, causa error) *Error {
	fallo := errorDeArgumentos(operacion,
		fmt.Sprintf("la %s apunta a %q, que existe y no es un directorio", origen, ruta), causa)
	fallo.Origen = origen
	fallo.Ruta = ruta

	return fallo
}

// errorDeDirectorioNoEscribible es la fila 6 cuando el culpable es el
// directorio: en modo normal no se puede crear, o existe y no deja crear la
// base ni sus auxiliares, y sin escribir la caché no sirve de nada. En modo de
// solo lectura esta situación no existe, porque allí no se crea ni se escribe
// nada (FR-015, FR-022).
func errorDeDirectorioNoEscribible(operacion, origen, ruta string, causa error) *Error {
	fallo := errorDeArgumentos(operacion,
		fmt.Sprintf("no se puede escribir en el directorio %q (%s)", ruta, origen), causa)
	fallo.Origen = origen
	fallo.Ruta = ruta

	return fallo
}

// errorDeFicheroNoEscribible es la fila 6 cuando el culpable es el propio
// fichero: cache.db ya existe en un directorio que sí admite escribir y es él
// el que no se deja abrir para escribir —sus permisos, o que lo que hay en la
// ruta no es un fichero—. La misma clase que la del directorio, porque quien
// invoca lo arregla igual —los permisos, u otro directorio—, y un mensaje que
// nombra al fichero y no a un directorio que no tiene la culpa (FR-022, FR-035).
func errorDeFicheroNoEscribible(operacion, origen, ruta string, causa error) *Error {
	fallo := errorDeArgumentos(operacion,
		fmt.Sprintf("no se puede abrir %q para escribir (%s)%s", ruta, origen, detalleDelAcceso(causa)), causa)
	fallo.Origen = origen
	fallo.Ruta = ruta

	return fallo
}

// errorDeFicheroIlegible es la fila 12 cuando el culpable es el propio fichero:
// en solo lectura cache.db está pero no se deja leer. No se sabe qué contiene,
// así que declararlo ausente sería mentir (FR-015); el mensaje nombra al
// fichero, y el acceso denegado cuando lo es, sin culpar al directorio (FR-035).
func errorDeFicheroIlegible(ruta string, causa error) *Error {
	fallo := errorInesperado("construir",
		fmt.Sprintf("no se puede leer %q en solo lectura%s", ruta, detalleDelAcceso(causa)), causa)
	fallo.Ruta = ruta

	return fallo
}

// errorDeBloqueo es la fila 14 cuando lo que sobreviene es que otra invocación
// retiene el bloqueo de escritura más tiempo que la espera: la base no está
// estropeada y el fichero no se toca, pero la operación no se pudo hacer. El
// mensaje nombra la operación y la clave cuando la hay, la ruta y la espera que
// se agotó, que es lo que distingue este fallo de un fichero inutilizable
// (FR-031, FR-035).
func errorDeBloqueo(operacion, ruta, clave string, espera time.Duration, causa error) *Error {
	motivo := fmt.Sprintf("%q está bloqueada por otra invocación y la espera de %s se agotó", ruta, espera)
	if clave != "" {
		motivo = fmt.Sprintf("no se pudo %s %q: %s", operacion, clave, motivo)
	}

	fallo := errorInesperado(operacion, motivo, causa)
	fallo.Ruta = ruta
	fallo.Clave = clave

	return fallo
}

// detalleDelAcceso es lo que el mensaje añade cuando el sistema de ficheros
// denegó el acceso: es la causa más común de que un fichero no se deje abrir y
// la que quien lee el fallo puede arreglar con sus permisos. Ante cualquier otra
// causa no se añade nada, para no afirmar lo que no se sabe (FR-035).
func detalleDelAcceso(causa error) string {
	if errors.Is(causa, fs.ErrPermission) {
		return ": acceso denegado"
	}

	return ""
}

// errorDeVersionAjena es la fila 9: el fichero trae un esquema que este binario
// no conoce. Las dos versiones no caben en ningún campo público y viajan con
// este constructor, que es lo que permite distinguir este mensaje del de la
// fila 10 aun compartiendo clase, operación y ruta. El fichero no se toca: lo
// dice el propio mensaje, para que quien lo lea sepa que puede volver con el
// binario que corresponda (FR-027).
func errorDeVersionAjena(operacion, ruta string, encontrada, conocida int64) *Error {
	fallo := errorInesperado(operacion, fmt.Sprintf(
		"%q tiene el esquema en la versión %d y este binario conoce la %d: no se modifica",
		ruta, encontrada, conocida), nil)
	fallo.Ruta = ruta

	return fallo
}

// errorDeFicheroInutilizable es la fila 10: en la ruta de la base hay un
// fichero que no es una base de datos que se pueda leer. Tampoco se borra, y el
// mensaje lo dice: lo que hay ahí puede ser de otro y perderlo sería peor que
// no poder usar la caché (FR-028).
func errorDeFicheroInutilizable(operacion, ruta string, causa error) *Error {
	fallo := errorInesperado(operacion,
		fmt.Sprintf("%q no es una base de datos utilizable; no se borra", ruta), causa)
	fallo.Ruta = ruta

	return fallo
}

// errorDeWALSinMemoriaCompartida es la fila 13: la base tiene un registro de
// escritura pendiente y el directorio no admite crear el fichero de memoria
// compartida que hace falta para leerlo, así que en solo lectura no hay forma
// de ver lo que ese registro contiene y declararlo ausente sería mentir. Los
// dos ficheros auxiliares se derivan de la ruta de la base y viajan con este
// constructor, porque son lo que explica el fallo y no caben en ningún campo
// público (D5).
func errorDeWALSinMemoriaCompartida(operacion, ruta string, causa error) *Error {
	fallo := errorInesperado(operacion, fmt.Sprintf(
		"%q tiene un registro de escritura %q y el directorio no permite crear %q: "+
			"no se puede leer en solo lectura",
		ruta, ruta+sufijoRegistroDeEscritura, ruta+sufijoMemoriaCompartida), causa)
	fallo.Ruta = ruta

	return fallo
}

// errorDeAusenciaEnSoloLectura es la fila 8: en modo de solo lectura no hay
// entrada vigente para la clave, sea porque no está, porque expiró, porque la
// base no tiene esquema o porque ni la base ni su directorio existen. Las
// cuatro son la misma cosa para quien pregunta —no hay nada que servir y no se
// puede ir a buscarlo—, y por eso comparten mensaje: nombra la clave y dice que
// la invocación es de solo lectura, que es lo que explica por qué una ausencia
// termina aquí en vez de en una petición a la fuente (FR-016, FR-018).
func errorDeAusenciaEnSoloLectura(ruta, clave string) *Error {
	fallo := errorDeFuenteNoDisponible("leer", fmt.Sprintf(
		"no hay entrada vigente para %q y la invocación es de solo lectura (--offline)", clave), nil)
	fallo.Ruta = ruta
	fallo.Clave = clave

	return fallo
}

// errorDeEscrituraEnSoloLectura es la fila 11: alguien intenta escribir en una
// caché construida para solo leer. Lo que está mal no es la invocación de la
// persona, sino el código que lo intenta, y por eso la clase es «inesperado» y
// no «argumentos». No se escribe nada (FR-017).
func errorDeEscrituraEnSoloLectura(ruta, clave string) *Error {
	fallo := errorInesperado("escribir",
		fmt.Sprintf("no se puede escribir %q en una caché de solo lectura", clave), nil)
	fallo.Ruta = ruta
	fallo.Clave = clave

	return fallo
}

// errorTrasCierre es la mitad de la fila 14: una operación llega después de
// Close. El mensaje nombra la operación —y la clave cuando la hay— porque es lo
// que sitúa la llamada que sobra en el código de quien usa la caché (FR-004).
func errorTrasCierre(operacion, ruta, clave string) *Error {
	motivo := "cliente cerrado; no se puede " + operacion
	if clave != "" {
		motivo += fmt.Sprintf(" %q", clave)
	}

	fallo := errorInesperado(operacion, motivo, nil)
	fallo.Ruta = ruta
	fallo.Clave = clave

	return fallo
}
