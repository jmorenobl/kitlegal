package graph

import (
	"context"
	"crypto/sha256"
	"database/sql"
	"encoding/binary"
	"encoding/hex"
	"fmt"
	"path/filepath"
	"strconv"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/jmorenobl/kitlegal/internal/core/grafo"
	"github.com/jmorenobl/kitlegal/internal/core/schema"
)

const (
	// versionDeEsteControlador es la SQLITE_VERSION_NUMBER de la SQLite que
	// lleva modernc.org/sqlite v1.59.0 (3.53.4), la que escribe en los bytes
	// 96-99 de la cabecera al confirmar (research.md V41).
	versionDeEsteControlador = 3053004
	// versionDeOtroControlador es la de otra SQLite, más antigua (3.43.2).
	versionDeOtroControlador = 3043002

	// paginaDeSQLite es el tamaño de página de las bases de las pruebas, el
	// que SQLite usa por omisión.
	paginaDeSQLite = 4096

	// filasBorradas son las filas de zeroblob(3000) que se escriben y se borran
	// para dejar páginas libres: una por página (research.md V45).
	filasBorradas = 40

	// sentenciaDeLaTabla y sentenciaDeLaFila escriben la tabla de fuera de las
	// bases con cabecera, cada una en su propia confirmación.
	sentenciaDeLaTabla = `CREATE TABLE ajena (x TEXT)`
	sentenciaDeLaFila  = `INSERT INTO ajena VALUES ('de fuera')`
)

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

// TestFijarWAL fija el paso a WAL de los pasos 3.4 y 4 (contracts/almacen-world-db.md
// §4 y §6; research.md V36 A, V42): fuera de toda transacción, un fichero de 0
// bytes pasa a una base en WAL de una página; dentro de una, SQLite no lo fija y
// lo calla —el pragma devuelve delete sin error—, y el paso lo dice con su
// mensaje en lugar de seguir con la base fuera de WAL.
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

// baseDeFuera es un world.db sin el esquema del grafo, como lo deja otro
// programa, y lo que deja de él una entrega que falla después de ponerlo en WAL
// (contracts/almacen-world-db.md §4.1 y §7; research.md D11, V36, V41, V45).
type baseDeFuera struct {
	nombre string
	// construir deja la base en la ruta, con su diario si lo lleva, y
	// comprueba las premisas de su construcción.
	construir func(t *testing.T, ruta string)
	// tamano es el de world.db al terminar.
	tamano int
	// distintos son los bytes de world.db que cambian, entre los que quedan.
	distintos []int
}

// TestPasoAWALDeUnaBaseDeFuera fija lo que declara contracts/almacen-world-db.md
// §4.1 para un world.db que existe sin esquema y fuera de WAL, por los pasos
// internos, porque por la API no se alcanza de forma determinista: el paso 4
// fija WAL; otra conexión toma la transacción inmediata y la mantiene; el paso
// 5, con un plazo corto, sale con el plazo agotado; y se cierran las dos
// conexiones. En cada base se afirman las cotas —mismas tablas y filas, en WAL,
// versión 0 y ningún fichero que no estuviera antes— y su resultado exacto: el
// tamaño y la lista literal de los bytes que cambian. Si una versión futura del
// controlador cambia SQLITE_VERSION_NUMBER o su forma de confirmar, el caso que
// cambie deja de valer y esta prueba lo dice: su lista se actualiza con la
// sonda, nunca se relaja a un intervalo. Una base que ya estaba en WAL queda con
// los mismos bytes.
func TestPasoAWALDeUnaBaseDeFuera(t *testing.T) {
	t.Parallel()

	cabecera := []int{18, 19, 27, 95}

	casos := []baseDeFuera{
		{nombre: "de 0 bytes", construir: deCeroBytes(false), tamano: paginaDeSQLite},
		{nombre: "de 0 bytes con un diario vacío", construir: deCeroBytes(true), tamano: paginaDeSQLite},
		{
			nombre:    "tabla de fuera con la cabecera de este controlador",
			construir: conCabecera(nil),
			tamano:    2 * paginaDeSQLite,
			distintos: cabecera,
		},
		{
			nombre:    "escrita por otra versión de SQLite",
			construir: conCabecera(map[int]uint32{96: versionDeOtroControlador}),
			tamano:    2 * paginaDeSQLite,
			distintos: []int{18, 19, 27, 95, 98, 99},
		},
		{
			nombre:    "version-valid-for distinto del contador y tamaño 0 en la cabecera",
			construir: conCabecera(map[int]uint32{92: 7, 28: 0}),
			tamano:    2 * paginaDeSQLite,
			distintos: []int{18, 19, 27, 31, 95},
		},
		{
			nombre:    "con acarreo en el contador",
			construir: conCabecera(map[int]uint32{24: 511, 92: 511}),
			tamano:    2 * paginaDeSQLite,
			distintos: []int{18, 19, 26, 27, 94, 95},
		},
		{
			nombre:    "auto_vacuum=full con páginas libres al final",
			construir: conPaginasLibres(false, true),
			tamano:    12288,
			distintos: []int{18, 19, 27, 31, 35, 39, 95},
		},
		{
			nombre:    "auto_vacuum=full con una página en uso detrás de las libres",
			construir: conPaginasLibres(true, true),
			tamano:    24576,
			distintos: []int{
				18, 19, 27, 31, 35, 39, 95, 4106, 4110, 4111, 4115, 12299, 16382, 16384,
				16388, 16389, 16390, 16391, 16392, 16393, 16395,
				16399, 16403, 16407, 16411, 16415, 16419, 16423, 16427, 16431, 16435, 16439, 16443, 16447,
				16451, 16455, 16459, 16463, 16467, 16471, 16475, 16479, 16483, 16487, 16491, 16495, 16499,
				16503, 16507, 16511, 16515, 16519, 16523, 16527, 16531, 16535, 16539, 16543, 16547,
			},
		},
		{
			nombre:    "auto_vacuum=INCREMENTAL con páginas libres",
			construir: conPaginasLibres(false, false),
			tamano:    176128,
			distintos: cabecera,
		},
		{
			nombre:    "con el diario frío de PERSIST",
			construir: conDiarioDeFuera("persist", 8720),
			tamano:    2 * paginaDeSQLite,
			distintos: cabecera,
		},
		{
			nombre:    "con el diario vacío de TRUNCATE",
			construir: conDiarioDeFuera("truncate", 0),
			tamano:    2 * paginaDeSQLite,
			distintos: cabecera,
		},
		{nombre: "ya en WAL", construir: yaEnWAL, tamano: 2 * paginaDeSQLite},
	}

	require.Len(t, casos[7].distintos, 59, "premisa: los 59 bytes de research.md V45")

	for _, caso := range casos {
		t.Run(caso.nombre, func(t *testing.T) {
			t.Parallel()

			compruebaPasoAWAL(t, caso)
		})
	}
}

// compruebaPasoAWAL es el cuerpo de cada caso de TestPasoAWALDeUnaBaseDeFuera.
func compruebaPasoAWAL(t *testing.T, caso baseDeFuera) {
	t.Helper()

	directorio := t.TempDir()
	ruta := filepath.Join(directorio, "world.db")
	caso.construir(t, ruta)

	antes := leerFichero(t, ruta)
	contenidoDeAntes := contenidoDe(t, ruta)

	lote := loteDelBOE(urlDeLaNorma, fechaDelBloque, laNorma(nil))
	consolidado, err := grafo.Consolidar(lote)
	require.NoError(t, err)

	base, err := abrirParaEscribir(t.Context(), ruta, lote)
	require.NoError(t, err, "el paso 4 abre y deja world.db en WAL")

	suelta := retenerLaTransaccion(t, ruta)

	conPlazo, cancela := context.WithTimeout(t.Context(), 150*time.Millisecond)
	err = escribirLote(conPlazo, base, ruta, consolidado)

	cancela()
	compruebaFalloDeEntrega(t, err, schema.ClaseFuenteNoDisponible, ruta,
		"grafo: el plazo termin\xc3\xb3 antes de escribir "+strconv.Quote(ruta))

	require.NoError(t, base.Close())
	suelta()

	despues := leerFichero(t, ruta)

	assert.Equal(t, []string{"world.db"}, nombresEn(t, directorio),
		"ningún fichero que no estuviera antes; el diario que había, ya no")
	assert.Len(t, despues, caso.tamano)
	assert.Equal(t, caso.distintos, bytesDistintos(antes, despues))
	assert.Equal(t, []byte{2, 2}, despues[18:20], "world.db queda en WAL")
	assert.Equal(t, contenidoDeAntes, contenidoDe(t, ruta), "mismas tablas y filas")
	assert.Equal(t, "0", contenidoDe(t, ruta)["version"], "sigue sin esquema")
}

// retenerLaTransaccion abre otra conexión sobre world.db —la de otra
// invocación— y toma en ella la transacción inmediata, que retiene el bloqueo
// de escritura hasta que se llama a lo que devuelve, que la deshace y cierra la
// conexión.
func retenerLaTransaccion(t *testing.T, ruta string) (suelta func()) {
	t.Helper()

	otra := abrirBaseDePrueba(t, ruta, pragmaDelTramo+"&_txlock=immediate")

	tx, err := otra.BeginTx(context.WithoutCancel(t.Context()), nil)
	require.NoError(t, err, "la otra conexión toma la transacción inmediata")

	return func() {
		require.NoError(t, tx.Rollback())
		require.NoError(t, otra.Close())
	}
}

// bytesDistintos son las posiciones en que difieren los dos contenidos, entre
// las que tienen los dos.
func bytesDistintos(antes, despues []byte) []int {
	var distintos []int

	for i := range min(len(antes), len(despues)) {
		if antes[i] != despues[i] {
			distintos = append(distintos, i)
		}
	}

	return distintos
}

// contenidoDe describe lo que guarda la base: su versión, cada objeto de su
// esquema y, de cada tabla, cuántas filas tiene y la huella de todas ellas.
// Abre la base inmutable, que la lee tal cual sin crear ni cambiar ningún
// fichero (research.md V46).
func contenidoDe(t *testing.T, ruta string) map[string]string {
	t.Helper()

	base := abrirBaseDePrueba(t, ruta, "mode=ro&immutable=1")
	contenido := map[string]string{"version": strconv.FormatInt(versionDe(t, base), 10)}

	for _, objeto := range filasDe(t, base, `SELECT type, name, ifnull(sql, '') FROM sqlite_schema ORDER BY name`) {
		contenido[objeto[0]+" "+objeto[1]] = objeto[2]

		if objeto[0] == "table" {
			consulta, conocida := filasDeCadaTabla[objeto[1]]
			require.True(t, conocida, "la tabla %s es de las que construyen las pruebas", objeto[1])

			contenido["filas de "+objeto[1]] = filasDeLaTabla(t, base, consulta)
		}
	}

	require.NoError(t, base.Close())

	return contenido
}

// filasDeCadaTabla lee, de cada tabla que construyen las bases de fuera, todas
// sus filas en el orden en que las guarda.
var filasDeCadaTabla = map[string]string{
	"ajena": `SELECT * FROM ajena ORDER BY rowid`,
	"a":     `SELECT * FROM a ORDER BY rowid`,
	"otra":  `SELECT * FROM otra ORDER BY rowid`,
}

// filasDeLaTabla es cuántas filas da la consulta y la huella de todas ellas.
func filasDeLaTabla(t *testing.T, base *sql.DB, consulta string) string {
	t.Helper()

	filas, err := base.QueryContext(t.Context(), consulta)
	require.NoError(t, err)

	defer func() { require.NoError(t, filas.Close()) }()

	nombres, err := filas.Columns()
	require.NoError(t, err)

	suma := sha256.New()
	cuantas := 0

	for filas.Next() {
		valores := make([]any, len(nombres))
		destinos := make([]any, len(nombres))

		for i := range valores {
			destinos[i] = &valores[i]
		}

		require.NoError(t, filas.Scan(destinos...))

		_, err := fmt.Fprintf(suma, "%#v\n", valores)
		require.NoError(t, err)

		cuantas++
	}

	require.NoError(t, filas.Err())

	return strconv.Itoa(cuantas) + " filas, " + hex.EncodeToString(suma.Sum(nil))
}

// deCeroBytes construye un world.db de 0 bytes y, si se pide, un diario vacío
// junto a él.
func deCeroBytes(conDiario bool) func(*testing.T, string) {
	return func(t *testing.T, ruta string) {
		t.Helper()

		escribirFichero(t, ruta, nil)

		if conDiario {
			escribirFichero(t, ruta+sufijoDiario, nil)
		}
	}
}

// construirConSQL crea la base con el modo de diario dado y ejecuta cada
// sentencia en su propia confirmación; al cerrarla no queda ningún auxiliar,
// salvo el diario que ese modo conserva.
func construirConSQL(t *testing.T, ruta, diario string, sentencias ...string) {
	t.Helper()

	base := abrirBaseDePrueba(t, ruta, pragmaDelTramo)
	fijarDiario(t, base, diario)

	for _, sentencia := range sentencias {
		_, err := base.ExecContext(t.Context(), sentencia)
		require.NoError(t, err, sentencia)
	}

	require.NoError(t, base.Close())
}

// conCabecera construye una base en modo rollback con una tabla de fuera y una
// fila, escrita por este controlador —dos páginas, el contador de cambios y
// version-valid-for en 2, el tamaño en cabecera válido y la versión de este
// controlador en 96-99, lo que research.md V41 llama «la cabecera de este
// controlador»— y retoca después cada entero de 4 bytes de la cabecera que se
// pide, por la posición de su primer byte.
func conCabecera(retoques map[int]uint32) func(*testing.T, string) {
	return func(t *testing.T, ruta string) {
		t.Helper()

		construirConSQL(t, ruta, diarioClasico, sentenciaDeLaTabla, sentenciaDeLaFila)

		fichero := leerFichero(t, ruta)
		require.Len(t, fichero, 2*paginaDeSQLite, "premisa: dos páginas")

		for posicion, esperado := range map[int]uint32{24: 2, 28: 2, 92: 2, 96: versionDeEsteControlador} {
			require.Equal(t, esperado, binary.BigEndian.Uint32(fichero[posicion:posicion+4]),
				"premisa: la cabecera de este controlador en %d", posicion)
		}

		for posicion, valor := range retoques {
			binary.BigEndian.PutUint32(fichero[posicion:posicion+4], valor)
		}

		escribirFichero(t, ruta, fichero)
	}
}

// conPaginasLibres construye una base con auto_vacuum=INCREMENTAL y 40 páginas
// libres: la tabla a con 40 filas de zeroblob(3000), una por página, que se
// borran; con otra, la tabla otra, creada antes que las filas, recibe dos
// después de las de a, así que sus páginas quedan detrás de las libres. Con
// completo, los bytes 64-67 se ponen a 0, que es auto_vacuum=full con las
// páginas libres que deja una aplicación que usa sqlite3_autovacuum_pages
// (research.md V45 a).
func conPaginasLibres(otra, completo bool) func(*testing.T, string) {
	return func(t *testing.T, ruta string) {
		t.Helper()

		sentencias := []string{`PRAGMA auto_vacuum=INCREMENTAL`, `CREATE TABLE a (b BLOB)`}
		paginas := 43

		if otra {
			sentencias = append(sentencias, `CREATE TABLE otra (b BLOB)`)
			paginas = 46
		}

		for range filasBorradas {
			sentencias = append(sentencias, `INSERT INTO a VALUES (zeroblob(3000))`)
		}

		if otra {
			sentencias = append(sentencias, `INSERT INTO otra VALUES (zeroblob(3000))`, `INSERT INTO otra VALUES (zeroblob(3000))`)
		}

		construirConSQL(t, ruta, diarioClasico, append(sentencias, `DELETE FROM a`)...)

		fichero := leerFichero(t, ruta)
		require.Len(t, fichero, paginas*paginaDeSQLite, "premisa: el tamaño con las páginas libres")
		require.Equal(t, uint32(filasBorradas), binary.BigEndian.Uint32(fichero[36:40]), "premisa: las páginas libres")
		require.Equal(t, uint32(1), binary.BigEndian.Uint32(fichero[64:68]), "premisa: auto_vacuum=INCREMENTAL")

		if completo {
			binary.BigEndian.PutUint32(fichero[64:68], 0)
			escribirFichero(t, ruta, fichero)
		}
	}
}

// conDiarioDeFuera construye una base con una tabla de fuera y una fila escrita
// con el modo de diario dado, que deja junto a ella un diario frío del tamaño
// dado: el de PERSIST, con la cabecera a cero; el de TRUNCATE, vacío (research.md
// V45 b).
func conDiarioDeFuera(diario string, tamano int) func(*testing.T, string) {
	return func(t *testing.T, ruta string) {
		t.Helper()

		construirConSQL(t, ruta, diario, sentenciaDeLaTabla, sentenciaDeLaFila)

		frio := leerFichero(t, ruta+sufijoDiario)
		require.Len(t, frio, tamano, "premisa: el diario que deja %s", diario)

		if tamano > 0 {
			require.Equal(t, make([]byte, 8), frio[:8], "premisa: la cabecera del diario está a cero, así que está frío")
		}
	}
}

// yaEnWAL construye una base en WAL con una tabla de fuera y una fila, cerrada
// sin auxiliares.
func yaEnWAL(t *testing.T, ruta string) {
	t.Helper()

	construirConSQL(t, ruta, diarioWAL, sentenciaDeLaTabla, sentenciaDeLaFila)
	require.Equal(t, []byte{2, 2}, leerFichero(t, ruta)[18:20], "premisa: la base está en WAL")
}
