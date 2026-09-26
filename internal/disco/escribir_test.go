package disco

import (
	"io/fs"
	"os"
	"path/filepath"
	"syscall"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestEscrituraAtomica fija research.md D8: EscribirFichero deja el fichero
// entero o no toca nada, porque escribe un temporal del mismo directorio y lo
// renombra encima; sin permiso de ejecución (FR-015); sin dejar ningún
// temporal, tampoco cuando falla; y sin escribir a través de un enlace ni
// sustituir lo que no es un fichero regular (FR-028, FR-047).
func TestEscrituraAtomica(t *testing.T) {
	t.Parallel()

	t.Run("un fichero nuevo, sin permiso de ejecución y sin temporales", func(t *testing.T) {
		t.Parallel()

		directorio := t.TempDir()
		ruta := filepath.Join(directorio, "SKILL.md")

		require.NoError(t, NuevoEscritor(Enlazador{}).EscribirFichero(ruta, []byte("nuevo")))

		assert.Equal(t, map[string]string{".": "directorio", "SKILL.md": "fichero: nuevo"}, arbolDe(t, directorio))
		assert.Equal(t, fs.FileMode(0o600), examinarDePrueba(t, ruta).Mode().Perm())
	})

	t.Run("un fichero que ya existe se sustituye entero por otro", func(t *testing.T) {
		t.Parallel()

		directorio := t.TempDir()
		ruta := filepath.Join(directorio, "SKILL.md")
		escribirDePrueba(t, ruta, "viejo, más largo que el nuevo")
		cambiarPermisos(t, ruta, 0o700)
		anterior := examinarDePrueba(t, ruta)

		require.NoError(t, NuevoEscritor(Enlazador{}).EscribirFichero(ruta, []byte("nuevo")))

		assert.Equal(t, map[string]string{".": "directorio", "SKILL.md": "fichero: nuevo"}, arbolDe(t, directorio))

		actual := examinarDePrueba(t, ruta)
		assert.False(t, os.SameFile(anterior, actual),
			"el nuevo es otro fichero puesto encima con un rename, no el viejo reescrito en su sitio")
		assert.Equal(t, fs.FileMode(0o600), actual.Mode().Perm(), "sin el permiso de ejecución que tenía el viejo")
	})

	for nombre, destino := range map[string]string{
		"no escribe a través de un enlace": "../fuera/ajeno",
		"no sustituye un enlace colgando":  "../fuera/no-existe",
	} {
		t.Run(nombre, func(t *testing.T) {
			t.Parallel()

			directorio := t.TempDir()
			crearDirectorioDePrueba(t, filepath.Join(directorio, "fuera"))
			escribirDePrueba(t, filepath.Join(directorio, "fuera", "ajeno"), "ajeno")
			crearDirectorioDePrueba(t, filepath.Join(directorio, "ambito"))
			ruta := filepath.Join(directorio, "ambito", "SKILL.md")
			enlazarDePrueba(t, destino, ruta)
			antes := arbolDe(t, directorio)

			err := NuevoEscritor(Enlazador{}).EscribirFichero(ruta, []byte("nuevo"))

			require.ErrorIs(t, err, errNoEsFicheroRegular)
			assert.Equal(t, antes, arbolDe(t, directorio), "ni el enlace ni su destino cambian")
		})
	}

	t.Run("donde hay un directorio no escribe nada y retira el temporal", func(t *testing.T) {
		t.Parallel()

		directorio := t.TempDir()
		ruta := conDirectorio(t, directorio)
		conFichero(t, ruta)
		antes := arbolDe(t, directorio)

		err := NuevoEscritor(Enlazador{}).EscribirFichero(ruta, []byte("nuevo"))

		require.Error(t, err)
		assert.Equal(t, antes, arbolDe(t, directorio), "ni el directorio cambia ni queda el temporal")
	})

	t.Run("por debajo de un fichero no escribe nada", func(t *testing.T) {
		t.Parallel()

		directorio := t.TempDir()
		ruta := filepath.Join(conFichero(t, directorio), "SKILL.md")
		antes := arbolDe(t, directorio)

		err := NuevoEscritor(Enlazador{}).EscribirFichero(ruta, []byte("nuevo"))

		require.ErrorIs(t, err, syscall.ENOTDIR)
		assert.Equal(t, antes, arbolDe(t, directorio))
	})

	t.Run("sin el directorio de encima no escribe nada", func(t *testing.T) {
		t.Parallel()

		directorio := t.TempDir()

		err := NuevoEscritor(Enlazador{}).EscribirFichero(filepath.Join(directorio, "falta", "SKILL.md"), []byte("nuevo"))

		require.ErrorIs(t, err, fs.ErrNotExist)
		assert.NotContains(t, err.Error(), directorio,
			"el error es la causa del sistema: la operación y la ruta las pone quien aplica el plan")
		assert.Equal(t, map[string]string{".": "directorio"}, arbolDe(t, directorio))
	})
}

// TestCrearDirectorio fija el CrearDirectorio del puerto Escritor: crea un
// directorio real, sin permiso de escritura para los demás (research.md D8),
// solo si su padre existe y no hay nada en la ruta.
func TestCrearDirectorio(t *testing.T) {
	t.Parallel()

	t.Run("un directorio real con los permisos del repositorio", func(t *testing.T) {
		t.Parallel()

		directorio := t.TempDir()
		ruta := filepath.Join(directorio, "skills")

		require.NoError(t, NuevoEscritor(Enlazador{}).CrearDirectorio(ruta))

		assert.Equal(t, map[string]string{".": "directorio", "skills": "directorio"}, arbolDe(t, directorio))
		assert.Equal(t, fs.FileMode(0o750), examinarDePrueba(t, ruta).Mode().Perm())
	})

	for nombre, preparar := range map[string]func(t *testing.T, directorio string) string{
		"donde hay un fichero no crea nada":         conFichero,
		"donde hay un enlace colgando no crea nada": conEnlaceA("no-existe"),
		"donde hay un directorio no crea nada":      conDirectorio,
	} {
		t.Run(nombre, func(t *testing.T) {
			t.Parallel()

			directorio := t.TempDir()
			ruta := preparar(t, directorio)
			antes := arbolDe(t, directorio)

			err := NuevoEscritor(Enlazador{}).CrearDirectorio(ruta)

			require.ErrorIs(t, err, fs.ErrExist)
			assert.Equal(t, antes, arbolDe(t, directorio))
		})
	}

	t.Run("sin el directorio de encima no crea nada", func(t *testing.T) {
		t.Parallel()

		directorio := t.TempDir()

		err := NuevoEscritor(Enlazador{}).CrearDirectorio(filepath.Join(directorio, "falta", "skills"))

		require.ErrorIs(t, err, fs.ErrNotExist)
		assert.Equal(t, map[string]string{".": "directorio"}, arbolDe(t, directorio))
	})
}

// TestRetirar fija el Retirar del puerto Escritor: retira un fichero, un
// enlace sin seguirlo —su destino se queda como estaba— o un directorio
// vacío, y nada más.
func TestRetirar(t *testing.T) {
	t.Parallel()

	casos := []struct {
		nombre   string
		preparar func(t *testing.T, directorio string) string
		// queda es lo que hay en el directorio tras retirar.
		queda map[string]string
	}{
		{
			nombre:   "un fichero",
			preparar: conFichero,
			queda:    map[string]string{".": "directorio"},
		},
		{
			nombre:   "un enlace a un fichero, sin tocar el fichero",
			preparar: conEnlaceA("fichero"),
			queda:    map[string]string{".": "directorio", "directorio": "directorio", "fichero": "fichero: contenido"},
		},
		{
			nombre: "un enlace a un directorio, sin tocar el directorio ni lo que tiene dentro",
			preparar: func(t *testing.T, directorio string) string {
				t.Helper()

				ruta := conEnlaceA("directorio")(t, directorio)
				escribirDePrueba(t, filepath.Join(directorio, "directorio", "dentro"), "dentro")

				return ruta
			},
			queda: map[string]string{
				".": "directorio", "directorio": "directorio", "directorio/dentro": "fichero: dentro",
				"fichero": "fichero: contenido",
			},
		},
		{
			nombre:   "un enlace colgando",
			preparar: conEnlaceA("no-existe"),
			queda:    map[string]string{".": "directorio", "directorio": "directorio", "fichero": "fichero: contenido"},
		},
		{
			nombre:   "un directorio vacío",
			preparar: conDirectorio,
			queda:    map[string]string{".": "directorio"},
		},
	}

	for _, caso := range casos {
		t.Run(caso.nombre, func(t *testing.T) {
			t.Parallel()

			directorio := t.TempDir()
			ruta := caso.preparar(t, directorio)

			require.NoError(t, NuevoEscritor(Enlazador{}).Retirar(ruta))

			assert.Equal(t, caso.queda, arbolDe(t, directorio))
		})
	}

	t.Run("un directorio con algo dentro no se retira", func(t *testing.T) {
		t.Parallel()

		directorio := t.TempDir()
		ruta := conDirectorio(t, directorio)
		conFichero(t, ruta)
		antes := arbolDe(t, directorio)

		require.Error(t, NuevoEscritor(Enlazador{}).Retirar(ruta))

		assert.Equal(t, antes, arbolDe(t, directorio))
	})

	t.Run("lo que no existe no se retira", func(t *testing.T) {
		t.Parallel()

		err := NuevoEscritor(Enlazador{}).Retirar(filepath.Join(t.TempDir(), "nada"))

		require.ErrorIs(t, err, fs.ErrNotExist)
	})
}
