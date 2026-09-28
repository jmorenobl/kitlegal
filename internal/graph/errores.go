package graph

import (
	"errors"
	"fmt"
	"io/fs"
	"strconv"
	"time"

	"github.com/jmorenobl/kitlegal/internal/core/schema"
)

// Error es el único error que el paquete produce, y cada situación de la tabla
// de contracts/almacen-world-db.md §6 lo construye con su clase: ninguna queda
// sin ella. Declara esa clase él mismo —implementa schema.ConClase— para que el
// kernel la encuentre con errors.As y la traduzca a código de salida sin que
// este paquete importe internal/cli.
//
// Los campos públicos son lo que quien llama necesita sin analizar ninguna
// cadena, y se leen con errors.As aunque el error llegue envuelto.
type Error struct {
	// Operacion es qué se estaba haciendo: «leer», para los verbos de graph, o
	// «escribir», para la entrega.
	Operacion string
	// Ruta es el world.db o el directorio implicado, y va vacía cuando todavía
	// no se conoce: la ruta no se pudo ubicar, o el fallo llegó antes de
	// resolverla.
	Ruta string
	// Causa es el error de origen —el del controlador, el del sistema de
	// ficheros, el del contexto, el de la caché, el rechazo del dominio—, que
	// Unwrap expone para que errors.Is y errors.As lo alcancen.
	Causa error

	// clase es la que el error declara al kernel; la fija el constructor de
	// cada situación y no se negocia desde fuera.
	clase schema.Clase
	// motivo es el mensaje ya compuesto, sin el encabezado: cada situación
	// nombra algo distinto y los campos públicos no alcanzan a llevarlo todo.
	motivo string
}

// Error implementa schema.ConClase: su clase la reconoce el kernel.
var _ schema.ConClase = (*Error)(nil)

const (
	// operacionLeer y operacionEscribir son el vocabulario cerrado de
	// Error.Operacion: leer es lo que hacen los verbos de graph; escribir, la
	// entrega de un lote.
	operacionLeer     = "leer"
	operacionEscribir = "escribir"

	// encabezado va delante de todo mensaje del paquete, también del de un
	// *Error que no construyó ningún constructor.
	encabezado = "grafo: "

	// motivoSinDeclarar es el mensaje de un *Error nulo o construido a cero
	// fuera de sus constructores: ni queda vacío ni leerlo entra en panic.
	motivoSinDeclarar = ficheroDelGrafo + " ha fallado sin declarar el motivo"
)

// Error devuelve el mensaje: en español, con el encabezado «grafo: » y
// nombrando world.db, por su ruta cuando ya se conoce (§6). Nunca devuelve la
// cadena vacía, ni sobre un *Error nulo.
func (e *Error) Error() string {
	if e == nil || e.motivo == "" {
		return encabezado + motivoSinDeclarar
	}

	return encabezado + e.motivo
}

// Unwrap expone la causa.
func (e *Error) Unwrap() error {
	if e == nil {
		return nil
	}

	return e.Causa
}

// Clase es la clase del fallo. Lo que ningún constructor declaró es
// «inesperado».
func (e *Error) Clase() schema.Clase {
	if e == nil || e.clase == "" {
		return schema.ClaseInesperado
	}

	return e.clase
}

// nuevoError compone un *Error de la situación.
func nuevoError(clase schema.Clase, operacion, ruta, motivo string, causa error) *Error {
	return &Error{Operacion: operacion, Ruta: ruta, Causa: causa, clase: clase, motivo: motivo}
}

// nombrar es cómo un mensaje nombra world.db: por su ruta entre comillas si ya
// se conoce —%q hace visibles los espacios y los caracteres de control— y por
// su nombre si todavía no.
func nombrar(ruta string) string {
	if ruta == "" {
		return ficheroDelGrafo
	}

	return strconv.Quote(ruta)
}

// errorDeUbicacion es la ruta no resoluble (§6, fila 1; FR-011): la opción
// ConDirectorio vacía, o la regla de la caché que no da un directorio. El
// mensaje repite el de la causa, que dice qué corregir.
func errorDeUbicacion(operacion string, causa error) *Error {
	return nuevoError(schema.ClaseArgumentos, operacion, "",
		"no se puede ubicar "+ficheroDelGrafo+": "+causa.Error(), causa)
}

// errorInutilizable es world.db que no es una base de datos, está dañado o no
// deja abrirse (§6, fila 2; FR-010). Si la causa es el acceso denegado, el
// mensaje lo dice, porque es lo que quien lo lee puede arreglar.
func errorInutilizable(operacion, ruta string, causa error) *Error {
	detalle := ""
	if errors.Is(causa, fs.ErrPermission) {
		detalle = ": acceso denegado"
	}

	return nuevoError(schema.ClaseInesperado, operacion, ruta,
		nombrar(ruta)+" no es una base de datos utilizable"+detalle+"; no se modifica", causa)
}

// errorEsDirectorio es la variante de la fila 2 en que world.db es un
// directorio (FR-010).
func errorEsDirectorio(operacion, ruta string) *Error {
	return nuevoError(schema.ClaseInesperado, operacion, ruta,
		nombrar(ruta)+" es un directorio y no una base de datos utilizable; no se modifica", nil)
}

// errorDeTransaccionInterrumpida es el diario de rollback caliente que la
// lectura en solo lectura no puede deshacer (§6, fila 3; research.md D10, V43).
func errorDeTransaccionInterrumpida(ruta string, causa error) *Error {
	return nuevoError(schema.ClaseInesperado, operacionLeer, ruta,
		nombrar(ruta)+" tiene una transacción interrumpida sin deshacer; no se modifica", causa)
}

// errorDeVersionPosterior es el esquema de una versión que este binario no
// conoce (§6, fila 4; FR-012): quien lo escribió sabía algo que aquí no se
// sabe, y no se toca.
func errorDeVersionPosterior(operacion, ruta string, encontrada, conocida int64) *Error {
	return nuevoError(schema.ClaseInesperado, operacion, ruta, fmt.Sprintf(
		"%s tiene el esquema en la versión %d y este binario conoce la %d: no se modifica",
		nombrar(ruta), encontrada, conocida), nil)
}

// errorDeBloqueo es la espera propia agotada mientras otra invocación retiene
// world.db (§6, fila 5; FR-014): la base no está estropeada, pero no se pudo
// esperar más.
func errorDeBloqueo(operacion, ruta string, espera time.Duration, causa error) *Error {
	return nuevoError(schema.ClaseInesperado, operacion, ruta, fmt.Sprintf(
		"%s está bloqueada por otra invocación y la espera de %s se agotó", nombrar(ruta), espera), causa)
}

// errorDePlazo es el contexto de quien llama terminado: el plazo de --timeout
// (§6, fila 6; FR-014). Es «fuente-no-disponible», como en la caché: no es un
// fallo del fichero.
func errorDePlazo(operacion, ruta string, causa error) *Error {
	return nuevoError(schema.ClaseFuenteNoDisponible, operacion, ruta,
		"el plazo terminó antes de "+operacion+" "+nombrar(ruta), causa)
}

// errorDeLoteRechazado es el lote que no entra (§6, fila 7; FR-024, FR-025): el
// motivo es el del rechazo del dominio, que nombra la operación y nunca
// repite el id de una Persona ni el cuerpo de un texto.
func errorDeLoteRechazado(ruta string, causa error) *Error {
	return nuevoError(schema.ClaseInesperado, operacionEscribir, ruta,
		"el lote no entra en "+nombrar(ruta)+": "+causa.Error(), causa)
}

// errorDeDirectorioNoEscribible es el directorio de world.db que la entrega no
// puede crear o en el que no puede escribir (§6, fila 8).
func errorDeDirectorioNoEscribible(directorio string, causa error) *Error {
	return nuevoError(schema.ClaseInesperado, operacionEscribir, directorio,
		"no se puede escribir "+ficheroDelGrafo+" en "+strconv.Quote(directorio)+": "+causa.Error(), causa)
}

// errorDeFicheroNoEscribible es world.db que existe y que el proceso no puede
// abrir para escribir (§6, fila 9; research.md V46): la entrega falla antes de
// abrir SQLite y no cambia nada.
func errorDeFicheroNoEscribible(ruta string, causa error) *Error {
	return nuevoError(schema.ClaseInesperado, operacionEscribir, ruta,
		"no se puede escribir "+nombrar(ruta)+": "+causa.Error()+"; no se modifica", causa)
}

// errorDePublicacion es la publicación del temporal con os.Link que falla por
// otra cosa que un world.db ya publicado por otra invocación (§6, fila 10;
// research.md S7).
func errorDePublicacion(directorio string, causa error) *Error {
	return nuevoError(schema.ClaseInesperado, operacionEscribir, directorio,
		"no se puede publicar "+ficheroDelGrafo+" en "+strconv.Quote(directorio)+": "+causa.Error(), causa)
}

// errorDeModoWAL es el PRAGMA journal_mode=WAL que no devuelve wal (§6, fila
// 11; research.md V42): la entrega falla antes de abrir la transacción.
func errorDeModoWAL(ruta, modo string) *Error {
	return nuevoError(schema.ClaseInesperado, operacionEscribir, ruta,
		"no se puede poner "+nombrar(ruta)+" en modo WAL: el modo sigue siendo "+modo, nil)
}

// errorDeEntradaSalida es un fallo sobrevenido del sistema de ficheros o del
// controlador que ninguna otra situación explica: comprobar si world.db está,
// cerrarlo, leerlo o escribirlo. Es «inesperado» y nombra la operación y la
// ruta; la causa va en el mensaje porque es lo único que lo explica.
func errorDeEntradaSalida(operacion, ruta string, causa error) *Error {
	return nuevoError(schema.ClaseInesperado, operacion, ruta,
		"no se pudo "+operacion+" "+nombrar(ruta)+": "+causa.Error(), causa)
}

// errorDeMigracion es una migración que falla dentro de su transacción
// (FR-013): nombra la migración y el fichero, y la transacción que la pidió la
// deshace entera.
func errorDeMigracion(ruta, migracion string, causa error) *Error {
	return nuevoError(schema.ClaseInesperado, operacionEscribir, ruta, fmt.Sprintf(
		"no se pudo aplicar la migración %q en %s: %s", migracion, nombrar(ruta), causa), causa)
}

// errorDeMigracionesIlegibles es lo que no debería poder pasar: las
// migraciones van dentro del ejecutable y leerlas no toca el disco. Si falla,
// el binario está mal construido.
func errorDeMigracionesIlegibles(ruta string, causa error) *Error {
	return nuevoError(schema.ClaseInesperado, operacionEscribir, ruta,
		"no se pueden leer las migraciones de "+ficheroDelGrafo+" que el binario trae dentro", causa)
}
