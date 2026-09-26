package disco

import (
	"io/fs"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"
)

// arbolDe es todo lo que hay por debajo de directorio, entrada a entrada y sin
// seguir ningún enlace: el tipo de cada una, los bytes de cada fichero regular
// y el destino literal de cada enlace. Con él se comprueba que una operación
// deja un directorio con las mismas entradas y los mismos bytes, y que no
// escribe ni retira nada a través de un enlace.
//
// Pasa por el árbol y lee con un os.Root, que no sigue los enlaces que salen
// de él: así ni el propio test puede tocar nada fuera del directorio.
func arbolDe(t *testing.T, directorio string) map[string]string {
	t.Helper()

	raiz, err := os.OpenRoot(directorio)
	require.NoError(t, err)

	arbol := map[string]string{}

	err = fs.WalkDir(raiz.FS(), ".", func(ruta string, entrada fs.DirEntry, err error) error {
		if err != nil {
			return err
		}

		arbol[ruta], err = descripcionDe(raiz, ruta, entrada.Type())

		return err
	})
	require.NoError(t, err)
	require.NoError(t, raiz.Close())

	return arbol
}

// descripcionDe es lo que arbolDe anota de la entrada de ruta: su tipo y, si
// es un fichero regular, sus bytes, o, si es un enlace, su destino literal.
func descripcionDe(raiz *os.Root, ruta string, tipo fs.FileMode) (string, error) {
	switch {
	case tipo&fs.ModeSymlink != 0:
		destino, err := raiz.Readlink(filepath.FromSlash(ruta))

		return "enlace -> " + destino, err
	case tipo.IsDir():
		return "directorio", nil
	case tipo.IsRegular():
		contenido, err := raiz.ReadFile(filepath.FromSlash(ruta))

		return "fichero: " + string(contenido), err
	default:
		return "otra: " + tipo.String(), nil
	}
}

// escribirDePrueba deja en ruta un fichero regular con contenido.
func escribirDePrueba(t *testing.T, ruta, contenido string) {
	t.Helper()

	require.NoError(t, os.WriteFile(ruta, []byte(contenido), 0o600))
}

// enlazarDePrueba deja en ruta un enlace simbólico con ese destino literal,
// resuelva o no.
func enlazarDePrueba(t *testing.T, destino, ruta string) {
	t.Helper()

	require.NoError(t, os.Symlink(destino, ruta))
}

// crearDirectorioDePrueba deja en ruta un directorio real.
func crearDirectorioDePrueba(t *testing.T, ruta string) {
	t.Helper()

	require.NoError(t, os.Mkdir(ruta, 0o750))
}

// cambiarPermisos deja ruta con los permisos de modo durante el test y le
// devuelve después los que tenía, para que t.TempDir pueda retirarla. El modo
// llega como parámetro y el que se restaura se lee del disco, no se escribe a
// mano.
func cambiarPermisos(t *testing.T, ruta string, modo fs.FileMode) {
	t.Helper()

	antes, err := os.Stat(ruta)
	require.NoError(t, err)

	require.NoError(t, os.Chmod(ruta, modo))

	t.Cleanup(func() {
		require.NoError(t, os.Chmod(ruta, antes.Mode().Perm()))
	})
}
