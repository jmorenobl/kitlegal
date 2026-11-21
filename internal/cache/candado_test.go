package cache

import (
	"io/fs"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// candadoDePrueba y caducidadDePrueba son los dos valores con los que la
// composición del aviso toma el candado (contracts/aviso-y-reinstalacion.md §4).
// La prueba los escribe como literales y no los toma de internal/app: el candado
// es de la caché y no sabe quién lo pide.
const (
	candadoDePrueba   = "skills-global.lock"
	caducidadDePrueba = 10 * time.Minute
)

// contenidoAjeno es lo que lleva el candado que la prueba deja de antemano. El
// candado que crea TomarCandado está vacío, así que un candado que conserva este
// contenido es el ajeno, intacto, y uno vacío es uno nuevo.
const contenidoAjeno = "de otra invocación"

// TestTomarCandado es el control de FR-010 sobre el adaptador: cada fila de la
// tabla del contrato del aviso y la reinstalación §4 que se puede provocar sin
// cambiar los permisos del sistema de ficheros, con lo que queda en el disco
// después (FR-013).
//
// No declara t.Parallel(), ni él ni sus subtests, y no es un descuido: el
// candado vive en el directorio que da Directorio(), que se declara con
// t.Setenv, y t.Setenv no se puede usar en pruebas paralelas. HOME y USERPROFILE
// apuntan a un directorio temporal, de modo que ningún subtest escribe bajo la
// cuenta real aunque la variable falte (research.md D7).
//
// Va en el paquete interno por un único subtest, el de la frontera de la
// caducidad: con el reloj del sistema, entre fijar la fecha del candado y
// compararla siempre pasa algo de tiempo, y la comparación nunca cae en la
// frontera exacta, así que solo un reloj parado distingue «menos de diez
// minutos» de «diez minutos o menos». Los demás llaman a TomarCandado, que es lo
// que llama la composición.
func TestTomarCandado(t *testing.T) {
	cuenta := t.TempDir()
	t.Setenv("HOME", cuenta)
	t.Setenv("USERPROFILE", cuenta)

	t.Run("sin candado se toma, vacío y solo para su dueño, y liberar lo retira", func(t *testing.T) {
		directorio := t.TempDir()
		t.Setenv(VariableDirectorio, directorio)
		ruta := filepath.Join(directorio, candadoDePrueba)

		liberar, tomado := TomarCandado(candadoDePrueba, caducidadDePrueba)

		require.True(t, tomado, "sin candado, se crea en exclusiva y se toma (FR-010)")
		require.NotNil(t, liberar, "el candado tomado se suelta con lo que devuelve TomarCandado")
		compruebaCandadoNuevo(t, ruta)

		liberar()

		compruebaSinCandado(t, ruta)
		assert.Equal(t, soloLaRaiz(), arbolDe(t, directorio), "liberar no deja nada en el directorio (FR-013)")
	})

	t.Run("uno de hace menos de diez minutos no se toma ni se toca", func(t *testing.T) {
		directorio := t.TempDir()
		t.Setenv(VariableDirectorio, directorio)
		ruta := filepath.Join(directorio, candadoDePrueba)
		dejaUnCandado(t, ruta, time.Now().Add(-(9*time.Minute + 30*time.Second)))

		antes, fecha := arbolDe(t, directorio), fechaDe(t, ruta)

		liberar, tomado := TomarCandado(candadoDePrueba, caducidadDePrueba)

		assert.False(t, tomado, "un candado reciente es de otra invocación que está reinstalando (FR-010)")
		assert.Nil(t, liberar, "sin candado tomado no hay nada que soltar (data-model §3)")
		assert.Equal(t, antes, arbolDe(t, directorio), "el candado ajeno queda igual, byte a byte")
		assert.True(t, fecha.Equal(fechaDe(t, ruta)), "y con la misma fecha de modificación: no se toca")
	})

	huerfanos := []struct {
		nombre string
		edad   time.Duration
	}{
		{nombre: "uno de hace diez minutos se sustituye", edad: caducidadDePrueba},
		{nombre: "uno de hace un año se sustituye", edad: 365 * 24 * time.Hour},
	}

	for _, huerfano := range huerfanos {
		t.Run(huerfano.nombre, func(t *testing.T) {
			directorio := t.TempDir()
			t.Setenv(VariableDirectorio, directorio)
			ruta := filepath.Join(directorio, candadoDePrueba)
			dejaUnCandado(t, ruta, time.Now().Add(-huerfano.edad))

			liberar, tomado := TomarCandado(candadoDePrueba, caducidadDePrueba)

			require.True(t, tomado, "un candado de hace diez minutos o más es huérfano y se sustituye (FR-010)")
			require.NotNil(t, liberar)
			compruebaCandadoNuevo(t, ruta)

			liberar()

			compruebaSinCandado(t, ruta)
		})
	}

	t.Run("la caducidad se cumple en su instante y ni un nanosegundo antes", func(t *testing.T) {
		directorio := t.TempDir()
		t.Setenv(VariableDirectorio, directorio)
		ruta := filepath.Join(directorio, candadoDePrueba)
		dejaUnCandado(t, ruta, instanteDePrueba)

		// La fecha se lee de vuelta porque es la que el sistema de ficheros
		// guardó, con su precisión, y es la que TomarCandado compara.
		fecha, antes := fechaDe(t, ruta), arbolDe(t, directorio)

		liberar, tomado := tomarCandado(candadoDePrueba, caducidadDePrueba,
			relojEn(fecha.Add(caducidadDePrueba-time.Nanosecond)).Ahora)

		assert.False(t, tomado, "a un nanosegundo de los diez minutos, el candado todavía es de otra invocación")
		assert.Nil(t, liberar)
		assert.Equal(t, antes, arbolDe(t, directorio), "y queda sin tocar")

		liberar, tomado = tomarCandado(candadoDePrueba, caducidadDePrueba,
			relojEn(fecha.Add(caducidadDePrueba)).Ahora)

		require.True(t, tomado, "a los diez minutos justos ya es huérfano: «de hace 10 minutos o más» (FR-010)")
		require.NotNil(t, liberar)
		compruebaCandadoNuevo(t, ruta)

		liberar()

		compruebaSinCandado(t, ruta)
	})

	t.Run("sin directorio lo crea con los permisos de la caché y lo deja", func(t *testing.T) {
		directorio := filepath.Join(t.TempDir(), "kitlegal")
		t.Setenv(VariableDirectorio, directorio)

		liberar, tomado := TomarCandado(candadoDePrueba, caducidadDePrueba)

		require.True(t, tomado, "si el directorio no existe, se crea y se sigue (FR-010)")
		require.NotNil(t, liberar)

		info, err := os.Lstat(directorio)
		require.NoError(t, err, "el directorio de la caché se creó")
		assert.True(t, info.IsDir())
		assert.Equal(t, fs.FileMode(0o700), info.Mode().Perm(), "con los permisos con que la caché crea el suyo (FR-010)")
		compruebaCandadoNuevo(t, filepath.Join(directorio, candadoDePrueba))

		liberar()

		assert.Equal(t, soloLaRaiz(), arbolDe(t, directorio),
			"liberar retira el candado y el directorio queda (contracts/aviso-y-reinstalacion.md §4)")
	})

	t.Run("la variable vacía no lo toma ni crea nada", func(t *testing.T) {
		t.Setenv(VariableDirectorio, "")

		antes := arbolDe(t, cuenta)

		liberar, tomado := TomarCandado(candadoDePrueba, caducidadDePrueba)

		assert.False(t, tomado, "si Directorio() falla, no se toma (FR-010)")
		assert.Nil(t, liberar)
		assert.Equal(t, antes, arbolDe(t, cuenta), "no cae al directorio de la cuenta")
		compruebaSinCandado(t, candadoDePrueba)
	})

	t.Run("la variable que apunta a un fichero no lo toma ni crea nada", func(t *testing.T) {
		padre := t.TempDir()
		fichero := filepath.Join(padre, "fichero")
		require.NoError(t, os.WriteFile(fichero, []byte("no es un directorio"), 0o600))
		t.Setenv(VariableDirectorio, fichero)

		antes := arbolDe(t, padre)

		liberar, tomado := TomarCandado(candadoDePrueba, caducidadDePrueba)

		assert.False(t, tomado, "si Directorio() falla, no se toma (FR-010)")
		assert.Nil(t, liberar)
		assert.Equal(t, antes, arbolDe(t, padre), "el fichero queda intacto y no aparece nada a su lado")
	})

	t.Run("el directorio que no se puede crear no lo toma ni crea nada", func(t *testing.T) {
		padre := t.TempDir()
		fichero := filepath.Join(padre, "fichero")
		require.NoError(t, os.WriteFile(fichero, []byte("no es un directorio"), 0o600))
		t.Setenv(VariableDirectorio, filepath.Join(fichero, "kitlegal"))

		antes := arbolDe(t, padre)

		liberar, tomado := TomarCandado(candadoDePrueba, caducidadDePrueba)

		assert.False(t, tomado, "si el directorio no se puede crear, no se toma (FR-010)")
		assert.Nil(t, liberar)
		assert.Equal(t, antes, arbolDe(t, padre))
	})

	t.Run("la creación que falla por otra causa no lo toma ni deja nada", func(t *testing.T) {
		directorio := t.TempDir()
		t.Setenv(VariableDirectorio, directorio)

		liberar, tomado := TomarCandado(filepath.Join("no-existe", candadoDePrueba), caducidadDePrueba)

		assert.False(t, tomado, "una creación que no falla por «ya existe» no toma el candado (contracts §4)")
		assert.Nil(t, liberar)
		assert.Equal(t, soloLaRaiz(), arbolDe(t, directorio))
	})

	t.Run("el huérfano que no se puede retirar no lo toma y queda", func(t *testing.T) {
		directorio := t.TempDir()
		t.Setenv(VariableDirectorio, directorio)
		ruta := filepath.Join(directorio, candadoDePrueba)
		require.NoError(t, os.Mkdir(ruta, 0o700))
		require.NoError(t, os.WriteFile(filepath.Join(ruta, "dentro"), []byte(contenidoAjeno), 0o600))
		require.NoError(t, os.Chtimes(ruta, instanteDePrueba, instanteDePrueba))

		antes := arbolDe(t, directorio)

		liberar, tomado := TomarCandado(candadoDePrueba, caducidadDePrueba)

		assert.False(t, tomado, "si la retirada del huérfano falla, no se toma (contracts §4)")
		assert.Nil(t, liberar)
		assert.Equal(t, antes, arbolDe(t, directorio), "y queda lo que hubiera")
	})
}

// dejaUnCandado deja en la ruta un candado ajeno, con contenido para poder ver
// después si se tocó, y con la fecha de modificación dada.
func dejaUnCandado(t *testing.T, ruta string, fecha time.Time) {
	t.Helper()

	require.NoError(t, os.WriteFile(ruta, []byte(contenidoAjeno), 0o600))
	require.NoError(t, os.Chtimes(ruta, fecha, fecha))
}

// fechaDe es la fecha de modificación del candado, la que decide si es de otra
// invocación o huérfano.
func fechaDe(t *testing.T, ruta string) time.Time {
	t.Helper()

	info, err := os.Lstat(ruta)
	require.NoError(t, err)

	return info.ModTime()
}

// compruebaCandadoNuevo comprueba que en la ruta está el candado que acaba de
// crear TomarCandado: un fichero regular, vacío —así que no es el ajeno, que
// lleva contenido— y solo para su dueño (data-model §3).
func compruebaCandadoNuevo(t *testing.T, ruta string) {
	t.Helper()

	info, err := os.Lstat(ruta)
	require.NoError(t, err, "el candado tomado existe en el directorio de la caché")
	assert.True(t, info.Mode().IsRegular(), "el candado es un fichero regular")
	assert.Equal(t, fs.FileMode(0o600), info.Mode().Perm(), "con los permisos del fichero de la caché")
	assert.Zero(t, info.Size(), "y vacío: es uno nuevo, no el que había")
}

// compruebaSinCandado comprueba que en la ruta no hay nada.
func compruebaSinCandado(t *testing.T, ruta string) {
	t.Helper()

	_, err := os.Lstat(ruta)
	assert.ErrorIs(t, err, fs.ErrNotExist, "no queda ningún candado en %s", ruta)
}

// soloLaRaiz es lo que arbolDe describe de un directorio vacío.
func soloLaRaiz() []string {
	return []string{"." + string(filepath.Separator)}
}
