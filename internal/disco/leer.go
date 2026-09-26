package disco

import (
	"errors"
	"io"
	"io/fs"
	"os"
	"path/filepath"

	"github.com/jmorenobl/kitlegal/internal/core/instalacion"
)

// Huella es la huella de los bytes del fichero regular de ruta, con la forma
// de instalacion.HuellaDe, y con las mismas garantías que Leer.
func (l Lector) Huella(ruta string) (string, error) {
	contenido, err := l.Leer(ruta)
	if err != nil {
		return "", err
	}

	return instalacion.HuellaDe(contenido), nil
}

// Leer son los bytes del fichero regular de ruta. Lo que no es un fichero
// regular —un enlace, aunque apunte a uno— no se abre, y un fichero que deja
// de ser el examinado antes de abrirlo no se lee: los dos son el error «no es
// un fichero regular» (research.md D6).
func (Lector) Leer(ruta string) ([]byte, error) {
	fichero, err := abrir(operacionLeer, ruta, fs.FileMode.IsRegular, errNoEsFicheroRegular)
	if err != nil {
		return nil, err
	}

	contenido, errAlLeer := io.ReadAll(fichero)
	errAlCerrar := fichero.Close()

	if err := errors.Join(delSistema(errAlLeer), delSistema(errAlCerrar)); err != nil {
		return nil, fallo(operacionLeer, ruta, err)
	}

	return contenido, nil
}

// abrir abre ruta si, examinada sin seguirla, es lo que dice es: un fichero
// regular o un directorio real. Si no lo es, no la abre y devuelve noLoEs.
func abrir(operacion, ruta string, es func(fs.FileMode) bool, noLoEs error) (*os.File, error) {
	examinado, err := os.Lstat(ruta)
	if err != nil {
		return nil, fallo(operacion, ruta, delSistema(err))
	}

	return abrirLoExaminado(operacion, ruta, examinado, es, noLoEs)
}

// abrirLoExaminado abre ruta, que al examinarla era examinado, si examinado es
// lo que dice es; y la da por abierta solo si lo abierto también lo es y es el
// mismo fichero que se examinó. Así, cambiarlo por un enlace, o por otro, entre
// el examen y la apertura no hace leer lo que no se examinó: se cierra y
// devuelve noLoEs (research.md D6).
func abrirLoExaminado(operacion, ruta string, examinado fs.FileInfo, es func(fs.FileMode) bool,
	noLoEs error,
) (*os.File, error) {
	if !es(examinado.Mode()) {
		return nil, fallo(operacion, ruta, noLoEs)
	}

	abierto, err := os.Open(filepath.Clean(ruta))
	if err != nil {
		return nil, fallo(operacion, ruta, delSistema(err))
	}

	loAbierto, err := abierto.Stat()

	switch {
	case err != nil:
		return nil, fallo(operacion, ruta, errors.Join(delSistema(err), delSistema(abierto.Close())))
	case !es(loAbierto.Mode()) || !os.SameFile(examinado, loAbierto):
		return nil, fallo(operacion, ruta, errors.Join(noLoEs, delSistema(abierto.Close())))
	default:
		return abierto, nil
	}
}
