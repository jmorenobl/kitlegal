package disco

import (
	"errors"
	"fmt"
	"io/fs"
	"os"
)

// Lo que el adaptador se niega a abrir, porque abrirlo sería leer a través de
// un enlace o quedarse bloqueado en una tubería con nombre (FR-028).
var (
	// errNoEsFicheroRegular es la ruta que se pide leer, o sustituir al
	// escribir, y no es un fichero regular: un enlace, un directorio, una
	// tubería, un socket o un dispositivo, o un fichero que dejó de ser el
	// examinado antes de abrirlo.
	errNoEsFicheroRegular = errors.New("no es un fichero regular")

	// errNoEsDirectorioReal es la ruta que se pide listar y no es un
	// directorio real: un enlace a uno también, o un directorio que dejó de
	// ser el examinado antes de abrirlo.
	errNoEsDirectorioReal = errors.New("no es un directorio real")
)

// Las operaciones del Lector y del Enlazador, como las nombran sus errores.
const (
	operacionExaminar = "examinar"
	operacionLeer     = "leer"
	operacionListar   = "listar"
	operacionRetirar  = "retirar la sonda"
)

// fallo es el error de una operación sobre una ruta: «<operación> <ruta>:
// <causa>».
func fallo(operacion, ruta string, causa error) error {
	return fmt.Errorf("%s %s: %w", operacion, ruta, causa)
}

// delSistema es la causa de un error del paquete os, sin la operación y las
// rutas con que la envuelve: la operación y la ruta que importan las pone
// quien sabe qué se estaba haciendo, el propio adaptador o el dominio. Otro
// error se devuelve tal cual.
func delSistema(err error) error {
	var enRuta *fs.PathError
	if errors.As(err, &enRuta) {
		return enRuta.Err
	}

	var entreRutas *os.LinkError
	if errors.As(err, &entreRutas) {
		return entreRutas.Err
	}

	return err
}
