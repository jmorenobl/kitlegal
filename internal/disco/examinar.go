package disco

import (
	"errors"
	"io/fs"
	"os"
	"syscall"

	"github.com/jmorenobl/kitlegal/internal/core/instalacion"
)

// Lector es el puerto Disco del dominio sobre el sistema de ficheros: mira
// cada ruta sin seguir su último enlace y solo abre un fichero regular o un
// directorio real, comprobando que lo abierto es lo examinado (research.md
// D6). Su valor cero es el que se usa; no guarda nada entre llamadas.
type Lector struct{}

// El Lector es un Disco.
var _ instalacion.Disco = Lector{}

// Examinar dice qué hay en ruta sin seguirla. Que no haya nada no es un error.
// De un enlace da su destino literal y si resuelve, que es lo único que mira
// al otro lado, y sin leer nada (FR-028).
func (Lector) Examinar(ruta string) (instalacion.Entrada, error) {
	examinado, err := os.Lstat(ruta)

	switch {
	case errors.Is(err, fs.ErrNotExist):
		return instalacion.Entrada{Tipo: instalacion.EntradaAusente}, nil
	case err != nil:
		return instalacion.Entrada{}, fallo(operacionExaminar, ruta, delSistema(err))
	}

	modo := examinado.Mode()

	switch {
	case modo&fs.ModeSymlink != 0:
		return examinarEnlace(ruta)
	case modo.IsDir():
		return instalacion.Entrada{Tipo: instalacion.EntradaDirectorio}, nil
	case modo.IsRegular():
		return instalacion.Entrada{Tipo: instalacion.EntradaFichero}, nil
	default:
		return instalacion.Entrada{Tipo: instalacion.EntradaOtra}, nil
	}
}

// examinarEnlace es la entrada del enlace de ruta: su destino literal, tal
// como se escribió, y si resuelve.
func examinarEnlace(ruta string) (instalacion.Entrada, error) {
	destino, err := os.Readlink(ruta)
	if err != nil {
		return instalacion.Entrada{}, fallo(operacionExaminar, ruta, delSistema(err))
	}

	resuelve, err := resuelve(ruta)
	if err != nil {
		return instalacion.Entrada{}, fallo(operacionExaminar, ruta, delSistema(err))
	}

	return instalacion.Entrada{Tipo: instalacion.EntradaEnlace, Destino: destino, Resuelve: resuelve}, nil
}

// resuelve dice si seguir el enlace de ruta, con los que encadene, llega a
// algo que existe. Es la única excepción de FR-028 (research.md D6): un Stat,
// que sigue el enlace sin abrir ni leer nada. No resuelve si cuelga, si pasa
// por algo que no es un directorio o si está en un ciclo; cualquier otro fallo
// —un permiso que no deja comprobarlo— no dice ni una cosa ni la otra, y es un
// error.
func resuelve(ruta string) (bool, error) {
	_, err := os.Stat(ruta)

	switch {
	case err == nil:
		return true, nil
	case errors.Is(err, fs.ErrNotExist), errors.Is(err, syscall.ENOTDIR), errors.Is(err, syscall.ELOOP):
		return false, nil
	default:
		return false, err
	}
}

// Nombres son los de las entradas del directorio real de ruta, en el orden en
// que las da el sistema. Lo que no es un directorio real, un enlace a uno
// incluido, no se abre y es un error.
func (Lector) Nombres(ruta string) ([]string, error) {
	directorio, err := abrir(operacionListar, ruta, fs.FileMode.IsDir, errNoEsDirectorioReal)
	if err != nil {
		return nil, err
	}

	nombres, errAlListar := directorio.Readdirnames(-1)
	errAlCerrar := directorio.Close()

	if err := errors.Join(delSistema(errAlListar), delSistema(errAlCerrar)); err != nil {
		return nil, fallo(operacionListar, ruta, err)
	}

	return nombres, nil
}
