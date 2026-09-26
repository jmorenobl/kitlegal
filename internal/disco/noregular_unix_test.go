//go:build unix

package disco

import (
	"os"
	"path/filepath"
	"syscall"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/jmorenobl/kitlegal/internal/core/instalacion"
)

// esperaMaxima es el plazo en que tiene que volver una operación sobre una
// tubería con nombre. Abrir una para leer, sin nadie al otro lado, no vuelve
// nunca; las operaciones que no la abren vuelven en microsegundos.
const esperaMaxima = 10 * time.Second

// TestNoAbreLoQueNoEsRegular fija research.md D6 sobre lo que no es un fichero
// regular ni un directorio real: una tubería con nombre y un dispositivo se
// examinan como «otra» entrada, y Leer, Huella y Nombres no los abren; y una
// tubería, que el rename sustituiría sin quejarse, no se sustituye al escribir
// (FR-028). Vive en un fichero solo para Unix porque syscall.Mkfifo no existe
// en Windows.
func TestNoAbreLoQueNoEsRegular(t *testing.T) {
	t.Parallel()

	t.Run("una tubería con nombre se examina sin abrirla ni leerla", func(t *testing.T) {
		t.Parallel()

		tuberia := tuberiaDePrueba(t)

		entrada, err := sinBloquear(t, func() (instalacion.Entrada, error) { return Lector{}.Examinar(tuberia) })
		require.NoError(t, err)
		assert.Equal(t, instalacion.Entrada{Tipo: instalacion.EntradaOtra}, entrada)

		_, err = sinBloquear(t, func() ([]byte, error) { return Lector{}.Leer(tuberia) })
		require.ErrorIs(t, err, errNoEsFicheroRegular)
		require.EqualError(t, err, "leer "+tuberia+": no es un fichero regular")

		_, err = sinBloquear(t, func() (string, error) { return Lector{}.Huella(tuberia) })
		require.ErrorIs(t, err, errNoEsFicheroRegular)

		_, err = sinBloquear(t, func() ([]string, error) { return Lector{}.Nombres(tuberia) })
		require.ErrorIs(t, err, errNoEsDirectorioReal)
		require.EqualError(t, err, "listar "+tuberia+": no es un directorio real")
	})

	t.Run("un dispositivo se examina sin abrirlo ni leerlo", func(t *testing.T) {
		t.Parallel()

		const dispositivo = "/dev/null"

		entrada, err := Lector{}.Examinar(dispositivo)
		require.NoError(t, err)
		assert.Equal(t, instalacion.Entrada{Tipo: instalacion.EntradaOtra}, entrada)

		_, err = Lector{}.Leer(dispositivo)
		require.ErrorIs(t, err, errNoEsFicheroRegular)

		_, err = Lector{}.Huella(dispositivo)
		require.ErrorIs(t, err, errNoEsFicheroRegular)
	})

	t.Run("un directorio no se lee como un fichero", func(t *testing.T) {
		t.Parallel()

		ruta := conDirectorio(t, t.TempDir())

		_, err := Lector{}.Leer(ruta)
		require.ErrorIs(t, err, errNoEsFicheroRegular)
		require.EqualError(t, err, "leer "+ruta+": no es un fichero regular")

		_, err = Lector{}.Huella(ruta)
		require.ErrorIs(t, err, errNoEsFicheroRegular)
	})

	t.Run("una tubería no se sustituye al escribir", func(t *testing.T) {
		t.Parallel()

		tuberia := tuberiaDePrueba(t)
		antes := arbolDe(t, filepath.Dir(tuberia))

		err := NuevoEscritor(Enlazador{}).EscribirFichero(tuberia, []byte("nuevo"))

		require.ErrorIs(t, err, errNoEsFicheroRegular)
		assert.Equal(t, antes, arbolDe(t, filepath.Dir(tuberia)), "la tubería sigue ahí y no queda ningún temporal")
	})
}

// tuberiaDePrueba deja una tubería con nombre en un directorio temporal y
// devuelve su ruta.
func tuberiaDePrueba(t *testing.T) string {
	t.Helper()

	ruta := filepath.Join(t.TempDir(), "tuberia")
	require.NoError(t, syscall.Mkfifo(ruta, 0o600))

	info, err := os.Lstat(ruta)
	require.NoError(t, err)
	require.NotZero(t, info.Mode()&os.ModeNamedPipe, "la ruta es una tubería con nombre")

	return ruta
}

// sinBloquear devuelve lo que devuelve la operación, y falla si no vuelve en
// esperaMaxima, que es lo que pasa cuando abre la tubería: entonces la
// operación se queda bloqueada hasta que termina el binario del test.
func sinBloquear[T any](t *testing.T, operacion func() (T, error)) (T, error) {
	t.Helper()

	type resultado struct {
		valor T
		err   error
	}

	hecho := make(chan resultado, 1)

	go func() {
		valor, err := operacion()
		hecho <- resultado{valor: valor, err: err}
	}()

	select {
	case r := <-hecho:
		return r.valor, r.err
	case <-time.After(esperaMaxima):
		t.Fatalf("la operación no volvió en %s: abrió lo que no es un fichero regular ni un directorio real (FR-028)",
			esperaMaxima)

		var cero T

		return cero, nil
	}
}
