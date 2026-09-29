package graph

import (
	"context"
	"database/sql"
	"fmt"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/jmorenobl/kitlegal/internal/core/schema"
)

// columna es una fila de pragma_table_info: nombre, tipo declarado, NOT NULL y
// posición en la clave primaria (0 si no es parte de ella).
type columna struct {
	nombre  string
	tipo    string
	noNula  bool
	enClave int
}

// esquemaDeLaVersion1 es el de data-model §3, tabla a tabla y columna a
// columna. La columna INTEGER PRIMARY KEY de schema_version es el rowid y
// SQLite no la declara NOT NULL.
var esquemaDeLaVersion1 = map[string][]columna{
	"schema_version": {
		{"version", "INTEGER", false, 1},
		{"aplicada_en", "TEXT", true, 0},
	},
	"nodes": {
		{"id", "TEXT", true, 1},
		{"type", "TEXT", true, 0},
		{"props", "TEXT", true, 0},
		{"first_seen", "TEXT", true, 0},
		{"first_source", "TEXT", true, 0},
		{"first_url", "TEXT", true, 0},
		{"last_seen", "TEXT", true, 0},
		{"source", "TEXT", true, 0},
		{"url", "TEXT", true, 0},
		{"ttl", "INTEGER", false, 0},
	},
	"edges": {
		{"src", "TEXT", true, 1},
		{"rel", "TEXT", true, 2},
		{"dst", "TEXT", true, 3},
		{"first_seen", "TEXT", true, 0},
		{"first_source", "TEXT", true, 0},
		{"first_url", "TEXT", true, 0},
		{"last_seen", "TEXT", true, 0},
		{"source", "TEXT", true, 0},
		{"url", "TEXT", true, 0},
		{"ttl", "INTEGER", false, 0},
	},
	"texts": {
		{"hash", "TEXT", true, 1},
		{"body", "TEXT", true, 0},
		{"fetched_at", "TEXT", true, 0},
		{"source", "TEXT", true, 0},
		{"url", "TEXT", true, 0},
	},
}

// TestMigracionesEmbebidas fija lo que el binario trae dentro: una sola
// migración, 0001_grafo.sql, que es la versión 1 (FR-013). El número del nombre
// es la versión, y la versión conocida es el número de migraciones.
func TestMigracionesEmbebidas(t *testing.T) {
	t.Parallel()

	lista, err := migracionesEmbebidas()
	require.NoError(t, err)
	require.Len(t, lista, 1)

	for posicion, cada := range lista {
		assert.Equal(t, int64(posicion)+1, cada.version)
		assert.Equal(t, fmt.Sprintf("%04d_", cada.version), cada.nombre[:5], "el nombre empieza por su versión")
		assert.NotEmpty(t, cada.sentencias)
	}

	assert.Equal(t, "0001_grafo.sql", lista[0].nombre)

	conocida, err := versionConocida()
	require.NoError(t, err)
	assert.Equal(t, int64(1), conocida)
}

// TestMigrar fija la migración de world.db dentro de la transacción que la pide
// (FR-003, FR-013; contracts/almacen-world-db.md §4, pasos 3.4 y 5): sobre una
// base sin esquema deja la versión 1 con el esquema de data-model §3 al
// confirmar, y nada al deshacer; mientras la transacción no se confirma, otra
// conexión no ve nada; una migración que falla a medias no deja nada; sobre la
// versión 1 no ejecuta nada; y un esquema posterior no se toca.
func TestMigrar(t *testing.T) {
	t.Parallel()

	t.Run("confirmada, la base queda en la versión 1 con el esquema entero", func(t *testing.T) {
		t.Parallel()

		ruta, base := baseEnWAL(t)
		otra := abrirBaseDePrueba(t, ruta, "mode=ro")

		antes := time.Now().UTC().Truncate(time.Second)
		tx := empezar(t, base)

		require.NoError(t, migrar(t.Context(), tx, ruta))
		assert.Equal(t, int64(0), versionDe(t, otra), "sin confirmar, otra conexión no ve nada")

		require.NoError(t, tx.Commit())

		despues := time.Now().UTC()

		assert.Equal(t, int64(1), versionDe(t, base))
		assert.Equal(t, int64(1), versionDe(t, otra), "confirmada, la ve cualquiera")
		compruebaEsquema(t, base)

		var aplicada string

		require.NoError(t, base.QueryRowContext(t.Context(),
			"SELECT aplicada_en FROM schema_version WHERE version = 1").Scan(&aplicada))

		instante, err := time.Parse(time.RFC3339, aplicada)
		require.NoError(t, err, "aplicada_en es RFC 3339")
		assert.Equal(t, instante.UTC().Format(time.RFC3339), aplicada, "en UTC")
		assert.False(t, instante.Before(antes), "no antes de migrar")
		assert.False(t, instante.After(despues), "ni después de confirmar")
	})

	t.Run("deshecha, no queda nada", func(t *testing.T) {
		t.Parallel()

		ruta, base := baseEnWAL(t)
		tx := empezar(t, base)

		require.NoError(t, migrar(t.Context(), tx, ruta))
		require.NoError(t, tx.Rollback())

		assert.Empty(t, tablasDe(t, base))
		assert.Equal(t, int64(0), versionDe(t, base))
	})

	t.Run("una migración que falla a medias no deja nada", func(t *testing.T) {
		t.Parallel()

		ruta, base := baseEnWAL(t)

		// edges la crea la migración después de schema_version y de nodes: el
		// fallo llega con las dos primeras ya creadas dentro de la transacción.
		_, err := base.ExecContext(t.Context(), "CREATE TABLE edges (ajena TEXT)")
		require.NoError(t, err)

		tx := empezar(t, base)
		err = migrar(t.Context(), tx, ruta)
		require.NoError(t, tx.Rollback())

		var fallo *Error

		require.ErrorAs(t, err, &fallo)
		assert.Equal(t, schema.ClaseInesperado, fallo.Clase())
		assert.Equal(t, ruta, fallo.Ruta)
		assert.Contains(t, fallo.Error(), `la migración "0001_grafo.sql"`)

		assert.Equal(t, []string{"edges"}, tablasDe(t, base), "solo la tabla que ya estaba")
		assert.Equal(t, int64(0), versionDe(t, base))
	})

	t.Run("sobre la versión 1 no ejecuta nada", func(t *testing.T) {
		t.Parallel()

		ruta, base := baseEnWAL(t)

		tx := empezar(t, base)
		require.NoError(t, migrar(t.Context(), tx, ruta))
		require.NoError(t, tx.Commit())

		tx = empezar(t, base)
		require.NoError(t, migrar(t.Context(), tx, ruta))
		require.NoError(t, tx.Commit())

		var filas int

		require.NoError(t, base.QueryRowContext(t.Context(), "SELECT count(*) FROM schema_version").Scan(&filas))
		assert.Equal(t, 1, filas, "una sola fila por versión")
		compruebaEsquema(t, base)
	})

	t.Run("un esquema posterior no se toca", func(t *testing.T) {
		t.Parallel()

		ruta, base := baseEnWAL(t)

		_, err := base.ExecContext(t.Context(),
			"CREATE TABLE schema_version (version INTEGER PRIMARY KEY, aplicada_en TEXT NOT NULL);"+
				"INSERT INTO schema_version VALUES (2, '2030-01-01T00:00:00Z')")
		require.NoError(t, err)

		tx := empezar(t, base)
		err = migrar(t.Context(), tx, ruta)
		require.NoError(t, tx.Rollback())

		var fallo *Error

		require.ErrorAs(t, err, &fallo)
		assert.Equal(t, schema.ClaseInesperado, fallo.Clase())
		assert.Equal(t, fmt.Sprintf("grafo: %q tiene el esquema en la versión 2 y este binario conoce la 1: no se modifica", ruta),
			fallo.Error())
		assert.Equal(t, []string{"schema_version"}, tablasDe(t, base))
		assert.Equal(t, int64(2), versionDe(t, base))
	})

	t.Run("el contexto terminado es el plazo agotado y no deja nada", func(t *testing.T) {
		t.Parallel()

		ruta, base := baseEnWAL(t)
		tx := empezar(t, base)

		ctx, cancela := context.WithCancel(t.Context())
		cancela()

		err := migrar(ctx, tx, ruta)
		require.NoError(t, tx.Rollback())

		var fallo *Error

		require.ErrorAs(t, err, &fallo)
		assert.Equal(t, schema.ClaseFuenteNoDisponible, fallo.Clase())
		require.ErrorIs(t, err, context.Canceled)
		assert.Empty(t, tablasDe(t, base))
	})
}

// TestVersionRegistrada fija la lectura de la versión: 0 sin la tabla
// schema_version —una base recién creada o de 0 bytes—, el máximo si la hay, y
// un fallo sobre un fichero que no es una base de datos, que no se toca.
func TestVersionRegistrada(t *testing.T) {
	t.Parallel()

	t.Run("sin esquema, la versión 0", func(t *testing.T) {
		t.Parallel()

		_, base := baseEnWAL(t)

		version, err := versionRegistrada(t.Context(), base)
		require.NoError(t, err)
		assert.Equal(t, int64(0), version)
	})

	t.Run("con varias filas, la mayor", func(t *testing.T) {
		t.Parallel()

		_, base := baseEnWAL(t)

		_, err := base.ExecContext(t.Context(),
			"CREATE TABLE schema_version (version INTEGER PRIMARY KEY, aplicada_en TEXT NOT NULL);"+
				"INSERT INTO schema_version VALUES (1, 'a'), (3, 'b'), (2, 'c')")
		require.NoError(t, err)

		version, err := versionRegistrada(t.Context(), base)
		require.NoError(t, err)
		assert.Equal(t, int64(3), version)
	})

	t.Run("un fichero que no es una base de datos", func(t *testing.T) {
		t.Parallel()

		ruta := filepath.Join(t.TempDir(), "world.db")
		require.NoError(t, os.WriteFile(ruta, []byte("no soy una base de datos, soy un texto largo"), 0o600))

		base := abrirBaseDePrueba(t, ruta, "mode=ro")

		_, err := versionRegistrada(t.Context(), base)
		require.Error(t, err)
	})
}

// baseEnWAL crea una base vacía en WAL en un directorio temporal, como la deja
// el paso previo a la transacción de una entrega, y devuelve su ruta y la
// conexión con la cadena de escritura que importa aquí: claves ajenas y
// transacciones inmediatas.
func baseEnWAL(t *testing.T) (string, *sql.DB) {
	t.Helper()

	ruta := filepath.Join(t.TempDir(), "world.db")
	base := abrirBaseDePrueba(t, ruta, pragmaDelTramo+"&_pragma=foreign_keys(1)&_txlock=immediate")

	var modo string

	require.NoError(t, base.QueryRowContext(t.Context(), "PRAGMA journal_mode=WAL").Scan(&modo))
	require.Equal(t, "wal", modo)

	return ruta, base
}

// empezar abre la transacción inmediata de una entrega, sin la cancelación del
// contexto, como la abre el almacén.
func empezar(t *testing.T, base *sql.DB) *sql.Tx {
	t.Helper()

	tx, err := base.BeginTx(context.WithoutCancel(t.Context()), nil)
	require.NoError(t, err)

	return tx
}

// versionDe es la versión registrada en la base, leída fuera de toda
// transacción.
func versionDe(t *testing.T, base *sql.DB) int64 {
	t.Helper()

	version, err := versionRegistrada(t.Context(), base)
	require.NoError(t, err)

	return version
}

// tablasDe son las tablas de la base, ordenadas, sin las internas de SQLite.
func tablasDe(t *testing.T, base *sql.DB) []string {
	t.Helper()

	filas, err := base.QueryContext(t.Context(),
		"SELECT name FROM sqlite_schema WHERE type = 'table' AND name NOT LIKE 'sqlite_%' ORDER BY name")
	require.NoError(t, err)

	defer func() { require.NoError(t, filas.Close()) }()

	var tablas []string

	for filas.Next() {
		var nombre string

		require.NoError(t, filas.Scan(&nombre))

		tablas = append(tablas, nombre)
	}

	require.NoError(t, filas.Err())

	return tablas
}

// compruebaEsquema compara el esquema de la base con el de data-model §3:
// tablas, cada columna, STRICT, índices y claves ajenas.
func compruebaEsquema(t *testing.T, base *sql.DB) {
	t.Helper()

	assert.Equal(t, []string{"edges", "nodes", "schema_version", "texts"}, tablasDe(t, base))

	for tabla, definidas := range esquemaDeLaVersion1 {
		assert.Equal(t, definidas, columnasDe(t, base, tabla), "cada columna de %s", tabla)

		var estricta bool

		require.NoError(t, base.QueryRowContext(t.Context(),
			"SELECT strict FROM pragma_table_list WHERE schema = 'main' AND name = ?", tabla).Scan(&estricta))
		assert.True(t, estricta, "%s es STRICT", tabla)
	}

	assert.Equal(t, [][3]string{
		{"edges_dst_rel", "edges", "dst,rel"},
		{"nodes_type_source", "nodes", "type,source"},
	}, filasDe(t, base,
		"SELECT m.name, m.tbl_name, group_concat(i.name, ',') FROM sqlite_schema AS m, pragma_index_info(m.name) AS i "+
			"WHERE m.type = 'index' AND m.sql IS NOT NULL GROUP BY m.name ORDER BY m.name"),
		"los dos índices, cada uno sobre sus campos en orden")

	assert.Equal(t, [][3]string{
		{"dst", "nodes", "id"},
		{"src", "nodes", "id"},
	}, filasDe(t, base, `SELECT "from", "table", "to" FROM pragma_foreign_key_list('edges') ORDER BY "from"`),
		"los dos extremos de una arista son nodos")
}

// columnasDe describe cada columna de la tabla, en su orden.
func columnasDe(t *testing.T, base *sql.DB, tabla string) []columna {
	t.Helper()

	filas, err := base.QueryContext(t.Context(),
		`SELECT name, type, "notnull", pk FROM pragma_table_info(?) ORDER BY cid`, tabla)
	require.NoError(t, err)

	defer func() { require.NoError(t, filas.Close()) }()

	var leidas []columna

	for filas.Next() {
		var cada columna

		require.NoError(t, filas.Scan(&cada.nombre, &cada.tipo, &cada.noNula, &cada.enClave))

		leidas = append(leidas, cada)
	}

	require.NoError(t, filas.Err())

	return leidas
}

// filasDe devuelve las filas de la consulta, de tres valores de texto cada una.
func filasDe(t *testing.T, base *sql.DB, consulta string) [][3]string {
	t.Helper()

	filas, err := base.QueryContext(t.Context(), consulta)
	require.NoError(t, err)

	defer func() { require.NoError(t, filas.Close()) }()

	var todas [][3]string

	for filas.Next() {
		var fila [3]string

		require.NoError(t, filas.Scan(&fila[0], &fila[1], &fila[2]))

		todas = append(todas, fila)
	}

	require.NoError(t, filas.Err())

	return todas
}
