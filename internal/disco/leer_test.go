package disco

import (
	"io/fs"
	"os"
	"path/filepath"
	"syscall"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/jmorenobl/kitlegal/internal/core/instalacion"
)

// TestLeerNoSigueEnlaces fija research.md D6 para Leer y Huella: dan los bytes
// y la huella de un fichero regular, y nunca leen a través de un enlace, ni
// cuando es la propia ruta ni cuando el fichero examinado se cambia por un
// enlace o por otro fichero antes de abrirlo (FR-028).
func TestLeerNoSigueEnlaces(t *testing.T) {
	t.Parallel()

	t.Run("los bytes y la huella de un fichero regular", func(t *testing.T) {
		t.Parallel()

		ruta := filepath.Join(t.TempDir(), "SKILL.md")
		escribirDePrueba(t, ruta, "---\nname: prueba\n---\n")

		contenido, err := Lector{}.Leer(ruta)
		require.NoError(t, err)
		assert.Equal(t, "---\nname: prueba\n---\n", string(contenido))

		huella, err := Lector{}.Huella(ruta)
		require.NoError(t, err)
		assert.Equal(t, instalacion.HuellaDe(contenido), huella)
	})

	t.Run("las rutas de encima se siguen: las examina antes el dominio", func(t *testing.T) {
		t.Parallel()

		directorio := t.TempDir()
		conFichero(t, conDirectorio(t, directorio))
		enlazarDePrueba(t, "directorio", filepath.Join(directorio, "enlace"))

		contenido, err := Lector{}.Leer(filepath.Join(directorio, "enlace", "fichero"))
		require.NoError(t, err)
		assert.Equal(t, "contenido", string(contenido))
	})

	for nombre, destino := range map[string]string{
		"un enlace a un fichero no se lee": "fichero",
		"un enlace colgando no se lee":     "no-existe",
	} {
		t.Run(nombre, func(t *testing.T) {
			t.Parallel()

			ruta := conEnlaceA(destino)(t, t.TempDir())

			_, err := Lector{}.Leer(ruta)
			require.ErrorIs(t, err, errNoEsFicheroRegular)
			require.EqualError(t, err, "leer "+ruta+": no es un fichero regular")

			_, err = Lector{}.Huella(ruta)
			require.ErrorIs(t, err, errNoEsFicheroRegular)
			require.EqualError(t, err, "leer "+ruta+": no es un fichero regular")
		})
	}

	t.Run("lo que no existe no se lee", func(t *testing.T) {
		t.Parallel()

		ruta := filepath.Join(t.TempDir(), "nada")

		_, err := Lector{}.Leer(ruta)
		require.ErrorIs(t, err, fs.ErrNotExist)
		require.EqualError(t, err, "leer "+ruta+": "+syscall.ENOENT.Error())

		_, err = Lector{}.Huella(ruta)
		require.ErrorIs(t, err, fs.ErrNotExist)
	})

	t.Run("un fichero que no se deja abrir es un error que nombra la operación y la ruta", func(t *testing.T) {
		t.Parallel()

		ruta := conFichero(t, t.TempDir())
		cambiarPermisos(t, ruta, 0)

		_, err := Lector{}.Leer(ruta)
		require.ErrorIs(t, err, fs.ErrPermission)
		require.EqualError(t, err, "leer "+ruta+": "+syscall.EACCES.Error())
	})

	t.Run("un fichero cambiado por un enlace tras examinarlo no se lee", func(t *testing.T) {
		t.Parallel()

		directorio := t.TempDir()
		ruta := conFichero(t, directorio)
		escribirDePrueba(t, filepath.Join(directorio, "otro"), "de otro")
		examinado := examinarDePrueba(t, ruta)

		require.NoError(t, os.Remove(ruta))
		enlazarDePrueba(t, "otro", ruta)

		_, err := abrirLoExaminado(operacionLeer, ruta, examinado, fs.FileMode.IsRegular, errNoEsFicheroRegular)
		require.ErrorIs(t, err, errNoEsFicheroRegular)
		require.EqualError(t, err, "leer "+ruta+": no es un fichero regular")
	})

	t.Run("un fichero cambiado por otro tras examinarlo no se lee", func(t *testing.T) {
		t.Parallel()

		directorio := t.TempDir()
		ruta := conFichero(t, directorio)
		otro := filepath.Join(directorio, "otro")
		escribirDePrueba(t, otro, "de otro")
		examinado := examinarDePrueba(t, ruta)

		require.NoError(t, os.Rename(otro, ruta))

		_, err := abrirLoExaminado(operacionLeer, ruta, examinado, fs.FileMode.IsRegular, errNoEsFicheroRegular)
		require.ErrorIs(t, err, errNoEsFicheroRegular)
		require.EqualError(t, err, "leer "+ruta+": no es un fichero regular")
	})

	t.Run("un directorio cambiado por un enlace tras examinarlo no se lista", func(t *testing.T) {
		t.Parallel()

		directorio := t.TempDir()
		ruta := conDirectorio(t, directorio)
		crearDirectorioDePrueba(t, filepath.Join(directorio, "otro"))
		examinado := examinarDePrueba(t, ruta)

		require.NoError(t, os.Remove(ruta))
		enlazarDePrueba(t, "otro", ruta)

		_, err := abrirLoExaminado(operacionListar, ruta, examinado, fs.FileMode.IsDir, errNoEsDirectorioReal)
		require.ErrorIs(t, err, errNoEsDirectorioReal)
		require.EqualError(t, err, "listar "+ruta+": no es un directorio real")
	})
}

// examinarDePrueba es lo que da Lstat de ruta, como lo examina el adaptador
// antes de abrirla.
func examinarDePrueba(t *testing.T, ruta string) fs.FileInfo {
	t.Helper()

	examinado, err := os.Lstat(ruta)
	require.NoError(t, err)

	return examinado
}
