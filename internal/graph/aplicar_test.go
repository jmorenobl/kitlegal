package graph

import (
	"context"
	"path/filepath"
	"strconv"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/jmorenobl/kitlegal/internal/core/schema"
)

// paginaDeSQLite es el tamaño de página de las bases de las pruebas, el que
// SQLite usa por omisión.
const paginaDeSQLite = 4096

// TestCadenaDeEscritura fija, carácter a carácter, la cadena de conexión de
// toda entrega (contracts/almacen-world-db.md §4; research.md D11, V10): sin
// journal_mode —el WAL se fija aparte, fuera de toda transacción—, con el tramo
// de espera, synchronous(FULL), las claves ajenas y la transacción inmediata;
// y la ruta con lo que el URI de SQLite interpreta protegido.
func TestCadenaDeEscritura(t *testing.T) {
	t.Parallel()

	assert.Equal(t,
		"file:/a/b%3Fc%23d%25e/world.db?_pragma=busy_timeout(100)&_pragma=synchronous(FULL)"+
			"&_pragma=foreign_keys(1)&_txlock=immediate",
		cadenaDeEscritura(filepath.FromSlash("/a/b?c#d%e/world.db")))
}

// TestFijarWAL fija el paso a WAL del paso 4 (contracts/almacen-world-db.md §4
// y §6; research.md V42): fuera de toda transacción, un fichero de 0 bytes pasa
// a una base en WAL de una página; dentro de una, SQLite no lo fija y lo calla
// —el pragma devuelve delete sin error—, y el paso lo dice con su mensaje en
// lugar de seguir con la base fuera de WAL.
func TestFijarWAL(t *testing.T) {
	t.Parallel()

	t.Run("fuera de toda transacción", func(t *testing.T) {
		t.Parallel()

		ruta := filepath.Join(t.TempDir(), "world.db")
		escribirFichero(t, ruta, nil)

		base, err := abrirConexion(operacionEscribir, ruta, cadenaDeEscritura(ruta))
		require.NoError(t, err)
		require.NoError(t, fijarWAL(t.Context(), base, ruta))
		require.NoError(t, fijarWAL(t.Context(), base, ruta), "una base que ya está en WAL se queda así")
		require.NoError(t, base.Close())

		fichero := leerFichero(t, ruta)
		assert.Len(t, fichero, paginaDeSQLite)
		assert.Equal(t, []byte{2, 2}, fichero[18:20])
	})

	t.Run("dentro de una transacción", func(t *testing.T) {
		t.Parallel()

		ruta := filepath.Join(t.TempDir(), "world.db")
		escribirFichero(t, ruta, nil)

		base, err := abrirConexion(operacionEscribir, ruta, cadenaDeEscritura(ruta))
		require.NoError(t, err)

		tx, err := base.BeginTx(context.WithoutCancel(t.Context()), nil)
		require.NoError(t, err)

		err = fijarWAL(t.Context(), tx, ruta)
		compruebaFalloDeEntrega(t, err, schema.ClaseInesperado, ruta,
			"grafo: no se puede poner "+strconv.Quote(ruta)+" en modo WAL: el modo sigue siendo delete")

		require.NoError(t, tx.Rollback())
		require.NoError(t, base.Close())
	})
}
