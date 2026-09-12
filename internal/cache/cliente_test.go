package cache

import (
	"log/slog"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/jmorenobl/kitlegal/internal/cli"
	"github.com/jmorenobl/kitlegal/internal/core/schema"
)

// TestNewRechazaOpcionesInvalidas fija la fila 3 de la tabla cerrada del
// contrato de errores §3: una opción con un valor que no sirve es «argumentos»
// (2) y el mensaje nombra **la opción**, porque lo que hay que corregir está en
// el código que la pasa y no en lo que la persona escribió (D2).
//
// Fija además que las opciones se validan **en orden** y que la primera
// inválida es la que decide: la última fila pasa dos opciones inválidas y exige
// que el mensaje nombre la primera y no la segunda. Sin eso, validarlas en
// cualquier orden pasaría igual y el fallo señalaría un sitio que quien lo lee
// no tocó.
//
// El control positivo de la cabecera es lo que impide que la tabla pase con un
// New que fallara siempre: con las cuatro opciones bien construidas no falla, y
// el cliente que devuelve se cierra —dos veces, porque Close es idempotente
// (FR-004)— sin error.
func TestNewRechazaOpcionesInvalidas(t *testing.T) {
	t.Parallel()

	cliente, err := New(t.Context(),
		ConDirectorio(t.TempDir()),
		SoloLectura(),
		ConReloj(time.Now),
		ConRegistrador(slog.New(slog.DiscardHandler)))
	require.NoError(t, err)
	require.NotNil(t, cliente)
	require.NoError(t, cliente.Close())
	require.NoError(t, cliente.Close(),
		"cerrar dos veces no es un fallo: el defer de quien lo construyó y el cierre "+
			"explícito conviven (FR-004)")

	casos := []struct {
		nombre   string
		opciones []Opcion
		nombra   string
		noNombra string
	}{
		{
			nombre:   "un directorio sin ninguna ruta dentro",
			opciones: []Opcion{ConDirectorio("")},
			nombra:   "ConDirectorio",
		},
		{
			nombre:   "un reloj nulo",
			opciones: []Opcion{ConReloj(nil)},
			nombra:   "ConReloj",
		},
		{
			nombre:   "un registrador nulo",
			opciones: []Opcion{ConRegistrador(nil)},
			nombra:   "ConRegistrador",
		},
		{
			nombre:   "con dos inválidas, la que se nombra es la primera",
			opciones: []Opcion{ConReloj(nil), ConRegistrador(nil)},
			nombra:   "ConReloj",
			noNombra: "ConRegistrador",
		},
	}

	for _, caso := range casos {
		t.Run(caso.nombre, func(t *testing.T) {
			t.Parallel()

			cliente, err := New(t.Context(), caso.opciones...)

			require.Error(t, err)
			assert.Nil(t, cliente)
			assert.Equal(t, schema.ClaseArgumentos, cli.Clasificar(err))
			assert.Equal(t, 2, cli.CodigoSalida(err))
			assert.Contains(t, err.Error(), "opción "+caso.nombra,
				"el mensaje nombra la opción que hay que corregir")

			if caso.noNombra != "" {
				assert.NotContains(t, err.Error(), caso.noNombra,
					"las opciones se validan en orden: la segunda inválida no se llega a mirar")
			}
		})
	}
}

// TestCloseEsIdempotente fija FR-004: la primera llamada cierra y las
// siguientes no hacen nada y no fallan. No es una comodidad, es lo que hace que
// el defer de quien construyó el cliente y un cierre explícito antes de tiempo
// convivan sin que el segundo parezca un error.
//
// Las dos filas son las dos formas de cliente que existen: la que abrió una base
// de datos y la que no llegó a abrir ninguna, porque una invocación de solo
// lectura no encontró cache.db. En la segunda no hay nada que cerrar, y cerrarla
// tampoco puede fallar.
func TestCloseEsIdempotente(t *testing.T) {
	t.Parallel()

	casos := []struct {
		nombre   string
		opciones []Opcion
	}{
		{nombre: "con base de datos"},
		{nombre: "sin base de datos", opciones: []Opcion{SoloLectura()}},
	}

	for _, caso := range casos {
		t.Run(caso.nombre, func(t *testing.T) {
			t.Parallel()

			// El directorio está vacío: en modo normal la base se crea, y en
			// solo lectura el cliente queda sin ninguna.
			opciones := append([]Opcion{ConDirectorio(t.TempDir())}, caso.opciones...)

			cliente, err := New(t.Context(), opciones...)
			require.NoError(t, err)

			assert.NoError(t, cliente.Close())
			assert.NoError(t, cliente.Close(), "cerrar dos veces no es un fallo (FR-004)")
			assert.NoError(t, cliente.Close(), "ni tres")
		})
	}
}

// TestNewPorOmisionUsaElDirectorioDeLaCuenta fija FR-019 y la última
// precedencia de FR-023: sin opción y sin variable, la caché vive en
// <cuenta>/.cache/kitlegal y en ningún otro sitio.
//
// «En ningún otro sitio» es la mitad que de verdad protege algo: un descuido en
// la resolución de la ruta escribiría en el directorio de trabajo, o en la raíz
// de la cuenta, y la base aparecería igual donde se la busca. Por eso el test
// enumera lo que hay bajo la cuenta y exige que no haya nada más que .cache,
// dentro nada más que kitlegal, y dentro cache.db con —como mucho— los dos
// auxiliares que SQLite pone a su lado.
//
// El directorio de la cuenta se redirige con t.Setenv: ningún test del hito toca
// la caché real de quien lo ejecuta (SC-004, SC-009, FR-040).
func TestNewPorOmisionUsaElDirectorioDeLaCuenta(t *testing.T) {
	cuenta := t.TempDir()
	t.Setenv("HOME", cuenta)
	t.Setenv("USERPROFILE", cuenta)
	sinVariable(t)

	cliente, err := New(t.Context())
	require.NoError(t, err)
	t.Cleanup(func() { require.NoError(t, cliente.Close()) })

	directorio := filepath.Join(cuenta, ".cache", "kitlegal")
	require.FileExists(t, filepath.Join(directorio, ficheroDeLaBase),
		"la base por omisión vive en <cuenta>/.cache/kitlegal/cache.db (FR-019, FR-020)")

	assert.Equal(t, []string{".cache"}, contenidoDe(t, cuenta),
		"bajo la cuenta no aparece nada más que .cache")
	assert.Equal(t, []string{"kitlegal"}, contenidoDe(t, filepath.Join(cuenta, ".cache")),
		"dentro de .cache, nada más que kitlegal")

	auxiliares := []string{ficheroDeLaBase, ficheroDeLaBase + "-wal", ficheroDeLaBase + "-shm"}
	for _, nombre := range contenidoDe(t, directorio) {
		assert.Contains(t, auxiliares, nombre,
			"en el directorio de la caché solo están la base y sus auxiliares")
	}
}

// contenidoDe enumera, ordenados, los nombres que hay en un directorio.
func contenidoDe(t *testing.T, directorio string) []string {
	t.Helper()

	entradas, err := os.ReadDir(directorio)
	require.NoError(t, err)

	nombres := make([]string, 0, len(entradas))
	for _, entrada := range entradas {
		nombres = append(nombres, entrada.Name())
	}

	return nombres
}
