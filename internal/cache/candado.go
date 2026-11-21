package cache

import (
	"errors"
	"io/fs"
	"os"
	"path/filepath"
	"time"
)

// TomarCandado intenta tomar el candado con el que una sola invocación a la vez
// reinstala las skills globales: el fichero nombre, regular y vacío, creado en
// exclusiva con los permisos del fichero de la caché (0o600) en el directorio
// que da Directorio(), y nunca en el de las skills (H38: FR-010 y FR-016,
// contracts/aviso-y-reinstalacion.md §4, data-model §3).
//
// Si el directorio no existe, lo crea con los permisos con que la caché crea el
// suyo (0o700) y lo deja. Si el candado ya existe, decide su fecha de
// modificación: de hace menos que la caducidad es de otra invocación que está
// reinstalando, y no se toma, no se toca y no se espera; de hace la caducidad o
// más es un huérfano que dejó una invocación interrumpida, y se retira y se crea
// otra vez en exclusiva, una sola vez (research.md D5).
//
// Con tomado, liberar retira el candado. Sin él, liberar es nil y no hay nada
// que soltar: si Directorio() falla, si el directorio no se puede crear, o si
// la creación, la retirada del huérfano o la segunda creación fallan por otra
// causa, queda lo que hubiera (contracts/aviso-y-reinstalacion.md §4). Ningún
// fallo se devuelve: quien lo pide da entonces la línea de hoy, que es lo que
// FR-010 pide, y la decisión es la del aviso de H19 (research.md D11).
//
// No espera ni reintenta, y no recibe ni devuelve nada de la base de datos.
func TomarCandado(nombre string, caducidad time.Duration) (liberar func(), tomado bool) {
	return tomarCandado(nombre, caducidad, time.Now)
}

// tomarCandado es TomarCandado con el reloj que fecha el «ahora» con el que se
// compara la del candado. Existe para que la frontera de la caducidad se pueda
// probar exacta: con el reloj del sistema, entre fijar la fecha y compararla
// siempre pasa algo de tiempo.
func tomarCandado(nombre string, caducidad time.Duration, ahora func() time.Time) (liberar func(), tomado bool) {
	directorio, err := Directorio()
	if err != nil {
		return nil, false
	}

	if err := os.MkdirAll(directorio, permisosDelDirectorio); err != nil {
		return nil, false
	}

	ruta := filepath.Join(directorio, nombre)

	err = creaEnExclusiva(ruta)
	if errors.Is(err, fs.ErrExist) {
		err = sustituyeElHuerfano(ruta, caducidad, ahora)
	}

	if err != nil {
		return nil, false
	}

	return func() { retira(ruta) }, true
}

// errCandadoDeOtra es el motivo por el que no se sustituye un candado que
// existe: es de hace menos que la caducidad, así que otra invocación lo tiene.
// No sale del paquete: TomarCandado lo traduce, como cualquier otro fallo, a no
// tomado.
var errCandadoDeOtra = errors.New("el candado es de otra invocación")

// sustituyeElHuerfano retira el candado que existe y lo crea otra vez en
// exclusiva, solo si es huérfano: su fecha de modificación es de hace la
// caducidad o más. Si es de hace menos, no lo toca. Comprobar la fecha y
// retirarlo no son una sola operación del sistema de ficheros, y lo que eso deja
// abierto está acotado en research.md D5.
func sustituyeElHuerfano(ruta string, caducidad time.Duration, ahora func() time.Time) error {
	info, err := os.Lstat(ruta)
	if err != nil {
		return err
	}

	if ahora().Sub(info.ModTime()) < caducidad {
		return errCandadoDeOtra
	}

	if err := os.Remove(ruta); err != nil {
		return err
	}

	return creaEnExclusiva(ruta)
}

// creaEnExclusiva crea el candado vacío solo si no existe: con O_EXCL, la
// creación falla con fs.ErrExist si ya hay algo en la ruta, también en Windows,
// donde se traduce a CREATE_NEW (research.md D4).
func creaEnExclusiva(ruta string) error {
	candado, err := os.OpenFile(filepath.Clean(ruta), os.O_WRONLY|os.O_CREATE|os.O_EXCL, permisosDelFichero)
	if err != nil {
		return err
	}

	return candado.Close()
}

// retira borra el candado. El resultado no se comprueba, y no es un error
// silenciado: el candado que no se puede borrar se deja como está y lo sustituye
// la primera invocación que lo encuentre pasada la caducidad, y quien suelta o
// renuncia al candado no tiene a quién darle el fallo, porque una reinstalación
// nunca es un error de la invocación (FR-012, research.md D11).
func retira(ruta string) {
	_ = os.Remove(ruta)
}
