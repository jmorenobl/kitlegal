package disco

import (
	"io/fs"
	"path/filepath"
	"syscall"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/jmorenobl/kitlegal/internal/core/instalacion"
)

// TestExaminar fija research.md D6 para Examinar y Nombres sobre un árbol real:
// la entrada de una ruta se mira sin seguirla, y de un enlace se dice su
// destino literal y si resuelve, que es lo único que se mira al otro lado
// (FR-028; data-model §3); un directorio se lista solo si es un directorio
// real; y ninguna de las dos cambia nada.
func TestExaminar(t *testing.T) {
	t.Parallel()

	casos := []struct {
		nombre string
		// preparar deja en el directorio lo que se examina y devuelve su ruta.
		preparar func(t *testing.T, directorio string) string
		esperada instalacion.Entrada
	}{
		{
			nombre:   "lo que no existe es una entrada ausente, no un error",
			preparar: func(_ *testing.T, directorio string) string { return filepath.Join(directorio, "nada") },
			esperada: instalacion.Entrada{Tipo: instalacion.EntradaAusente},
		},
		{
			nombre:   "un directorio real",
			preparar: conDirectorio,
			esperada: instalacion.Entrada{Tipo: instalacion.EntradaDirectorio},
		},
		{
			nombre:   "un fichero regular",
			preparar: conFichero,
			esperada: instalacion.Entrada{Tipo: instalacion.EntradaFichero},
		},
		{
			nombre:   "un enlace a un fichero es un enlace que resuelve, no un fichero",
			preparar: conEnlaceA("fichero"),
			esperada: instalacion.Entrada{Tipo: instalacion.EntradaEnlace, Destino: "fichero", Resuelve: true},
		},
		{
			nombre:   "un enlace a un directorio es un enlace que resuelve, no un directorio",
			preparar: conEnlaceA("directorio"),
			esperada: instalacion.Entrada{Tipo: instalacion.EntradaEnlace, Destino: "directorio", Resuelve: true},
		},
		{
			nombre:   "el destino es el literal, sin limpiar",
			preparar: conEnlaceA("./directorio/../fichero"),
			esperada: instalacion.Entrada{
				Tipo: instalacion.EntradaEnlace, Destino: "./directorio/../fichero", Resuelve: true,
			},
		},
		{
			nombre:   "un enlace colgando no resuelve",
			preparar: conEnlaceA("../../.agents/skills/boe-legislacion"),
			esperada: instalacion.Entrada{
				Tipo: instalacion.EntradaEnlace, Destino: "../../.agents/skills/boe-legislacion",
			},
		},
		{
			nombre:   "un enlace que pasa por un fichero no resuelve",
			preparar: conEnlaceA("fichero/dentro"),
			esperada: instalacion.Entrada{Tipo: instalacion.EntradaEnlace, Destino: "fichero/dentro"},
		},
		{
			nombre:   "un enlace en un ciclo no resuelve",
			preparar: conCiclo,
			esperada: instalacion.Entrada{Tipo: instalacion.EntradaEnlace, Destino: "otro"},
		},
	}

	for _, caso := range casos {
		t.Run(caso.nombre, func(t *testing.T) {
			t.Parallel()

			directorio := t.TempDir()
			ruta := caso.preparar(t, directorio)
			antes := arbolDe(t, directorio)

			entrada, err := Lector{}.Examinar(ruta)
			require.NoError(t, err)

			assert.Equal(t, caso.esperada, entrada)
			assert.Equal(t, antes, arbolDe(t, directorio), "examinar no cambia nada")
		})
	}

	t.Run("por debajo de un fichero es un error que nombra la operación y la ruta", func(t *testing.T) {
		t.Parallel()

		ruta := filepath.Join(conFichero(t, t.TempDir()), "dentro")

		_, err := Lector{}.Examinar(ruta)

		require.ErrorIs(t, err, syscall.ENOTDIR)
		require.EqualError(t, err, "examinar "+ruta+": "+syscall.ENOTDIR.Error())
	})

	t.Run("un enlace que no se deja comprobar es un error, no un enlace colgando", func(t *testing.T) {
		t.Parallel()

		directorio := t.TempDir()
		cerrado := filepath.Join(directorio, "cerrado")
		crearDirectorioDePrueba(t, cerrado)
		escribirDePrueba(t, filepath.Join(cerrado, "fichero"), "dentro")
		ruta := filepath.Join(directorio, "enlace")
		enlazarDePrueba(t, "cerrado/fichero", ruta)
		cambiarPermisos(t, cerrado, 0)

		_, err := Lector{}.Examinar(ruta)

		require.ErrorIs(t, err, fs.ErrPermission)
		require.EqualError(t, err, "examinar "+ruta+": "+syscall.EACCES.Error())
	})
}

// TestNombres fija research.md D6 para Nombres: da las entradas de un
// directorio real, sin mirar dentro de ninguna, y no lista lo que no es un
// directorio real.
func TestNombres(t *testing.T) {
	t.Parallel()

	t.Run("las entradas de un directorio real, también las ocultas y los enlaces", func(t *testing.T) {
		t.Parallel()

		directorio := t.TempDir()
		conDirectorio(t, directorio)
		conFichero(t, directorio)
		escribirDePrueba(t, filepath.Join(directorio, ".oculto"), "oculto")
		enlazarDePrueba(t, "no-existe", filepath.Join(directorio, "colgando"))
		antes := arbolDe(t, directorio)

		nombres, err := Lector{}.Nombres(directorio)
		require.NoError(t, err)

		assert.ElementsMatch(t, []string{"directorio", "fichero", ".oculto", "colgando"}, nombres)
		assert.Equal(t, antes, arbolDe(t, directorio), "listar no cambia nada")
	})

	t.Run("un directorio vacío no tiene entradas", func(t *testing.T) {
		t.Parallel()

		nombres, err := Lector{}.Nombres(t.TempDir())
		require.NoError(t, err)

		assert.Empty(t, nombres)
	})

	for nombre, preparar := range map[string]func(t *testing.T, directorio string) string{
		"un enlace a un directorio no se lista": conEnlaceA("directorio"),
		"un fichero no se lista":                conFichero,
	} {
		t.Run(nombre, func(t *testing.T) {
			t.Parallel()

			ruta := preparar(t, t.TempDir())

			_, err := Lector{}.Nombres(ruta)

			require.ErrorIs(t, err, errNoEsDirectorioReal)
			require.EqualError(t, err, "listar "+ruta+": no es un directorio real")
		})
	}

	t.Run("lo que no existe no se lista", func(t *testing.T) {
		t.Parallel()

		ruta := filepath.Join(t.TempDir(), "nada")

		_, err := Lector{}.Nombres(ruta)

		require.ErrorIs(t, err, fs.ErrNotExist)
		require.EqualError(t, err, "listar "+ruta+": "+syscall.ENOENT.Error())
	})
}

// conDirectorio deja en el directorio un directorio real, «directorio», y
// devuelve su ruta.
func conDirectorio(t *testing.T, directorio string) string {
	t.Helper()

	ruta := filepath.Join(directorio, "directorio")
	crearDirectorioDePrueba(t, ruta)

	return ruta
}

// conFichero deja en el directorio un fichero regular, «fichero», y devuelve
// su ruta.
func conFichero(t *testing.T, directorio string) string {
	t.Helper()

	ruta := filepath.Join(directorio, "fichero")
	escribirDePrueba(t, ruta, "contenido")

	return ruta
}

// conEnlaceA prepara en el directorio un directorio real, un fichero regular y
// un enlace, «enlace», con el destino literal dado, y devuelve la ruta del
// enlace.
func conEnlaceA(destino string) func(t *testing.T, directorio string) string {
	return func(t *testing.T, directorio string) string {
		t.Helper()

		conDirectorio(t, directorio)
		conFichero(t, directorio)

		ruta := filepath.Join(directorio, "enlace")
		enlazarDePrueba(t, destino, ruta)

		return ruta
	}
}

// conCiclo deja en el directorio dos enlaces que apuntan el uno al otro y
// devuelve la ruta del primero.
func conCiclo(t *testing.T, directorio string) string {
	t.Helper()

	ruta := filepath.Join(directorio, "enlace")
	enlazarDePrueba(t, "otro", ruta)
	enlazarDePrueba(t, "enlace", filepath.Join(directorio, "otro"))

	return ruta
}
