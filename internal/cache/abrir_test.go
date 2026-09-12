package cache

import (
	"crypto/sha256"
	"encoding/hex"
	"io/fs"
	"os"
	"path/filepath"
	"slices"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/jmorenobl/kitlegal/internal/cli"
	"github.com/jmorenobl/kitlegal/internal/core/schema"
)

// Los dos modos de apertura, escritos una sola vez: cada tabla que mide algo
// que tiene que valer igual se escriba o solo se lea pasa por los dos, y así
// ninguna comprobación queda hecha para un modo y olvidada para el otro
// (FR-015).
var modosDeApertura = []struct {
	nombre   string
	opciones []Opcion
}{
	{nombre: "normal"},
	{nombre: "solo-lectura", opciones: []Opcion{SoloLectura()}},
}

// TestAbrirCreaDirectorioYFicheroConPermisosReservados fija FR-021 sobre el
// disco: el directorio de la caché nace solo para su dueño y el fichero solo
// legible y escribible por él.
//
// Que el fichero lo cree el propio paquete antes de abrirlo con SQLite no es un
// adorno: dejándoselo al controlador nacería a 0644 —legible por cualquiera de
// la máquina— y los auxiliares del registro de escritura heredarían esos
// permisos, porque heredan los del fichero de la base (D4, sonda 1 I).
//
// El directorio no existe cuando empieza el test: crearlo es parte de lo que se
// mide, y sobre un directorio que ya estuviera con otros permisos no se sabría
// si los puso New.
func TestAbrirCreaDirectorioYFicheroConPermisosReservados(t *testing.T) {
	t.Parallel()

	directorio := filepath.Join(t.TempDir(), "kitlegal")

	cliente, err := New(t.Context(), ConDirectorio(directorio))
	require.NoError(t, err)
	t.Cleanup(func() { require.NoError(t, cliente.Close()) })

	delDirectorio, err := os.Stat(directorio)
	require.NoError(t, err, "el modo normal crea el directorio que no existe")
	assert.True(t, delDirectorio.IsDir())
	assert.Equal(t, fs.FileMode(0o700), delDirectorio.Mode().Perm(),
		"el directorio de la caché es solo de su dueño (FR-021)")

	delFichero, err := os.Stat(filepath.Join(directorio, ficheroDeLaBase))
	require.NoError(t, err, "el modo normal crea la base de datos")
	assert.Equal(t, fs.FileMode(0o600), delFichero.Mode().Perm(),
		"la base nace con acceso reservado a la cuenta y no con los permisos que "+
			"le pondría SQLite (FR-021, D4)")
}

// TestAbrirAplicaLosPragma mide en la propia base los cuatro PRAGMA del
// contrato de apertura §4, que es lo que SC-006 pide: no que el DSN los
// nombre, sino que la conexión los tenga puestos.
//
// Cada uno protege algo distinto: el diario en WAL es lo que permite leer
// mientras otra invocación escribe; la confirmación sincronizada con el disco
// es la integridad que FR-031 prohíbe cambiar por velocidad; la espera ante
// bloqueo es lo que hace que dos invocaciones simultáneas se turnen en vez de
// fallar; y en solo lectura query_only refuerza dentro de la conexión lo que el
// modo ya promete (FR-030, FR-031).
//
// Los PRAGMA se consultan por la conexión del cliente y no abriendo otra: el
// contrato prohíbe exponerla (FR-005), y una conexión nueva no sería la que New
// configuró, que es justo lo que aquí se mide.
func TestAbrirAplicaLosPragma(t *testing.T) {
	t.Parallel()

	directorio := t.TempDir()

	normal, err := New(t.Context(), ConDirectorio(directorio))
	require.NoError(t, err)
	t.Cleanup(func() { require.NoError(t, normal.Close()) })

	assert.Equal(t, "wal", pragmaTexto(t, normal, "PRAGMA journal_mode"),
		"el diario va en WAL: un lector no bloquea al escritor ni al revés")
	assert.Equal(t, int64(2), pragmaEntero(t, normal, "PRAGMA synchronous"),
		"cada confirmación se sincroniza con el disco (FULL)")
	assert.Equal(t, int64(5000), pragmaEntero(t, normal, "PRAGMA busy_timeout"),
		"ante un bloqueo se espera, no se falla de inmediato")
	assert.Equal(t, int64(0), pragmaEntero(t, normal, "PRAGMA query_only"),
		"el modo normal escribe: query_only es de la otra apertura")

	require.NoError(t, normal.Close())

	soloLectura, err := New(t.Context(), ConDirectorio(directorio), SoloLectura())
	require.NoError(t, err)
	t.Cleanup(func() { require.NoError(t, soloLectura.Close()) })

	assert.Equal(t, int64(1), pragmaEntero(t, soloLectura, "PRAGMA query_only"),
		"en solo lectura toda escritura falla en la conexión aunque el modo no bastara")
	assert.Equal(t, int64(5000), pragmaEntero(t, soloLectura, "PRAGMA busy_timeout"),
		"el lector también espera: no falla porque otra invocación esté escribiendo")
	assert.Equal(t, "wal", pragmaTexto(t, soloLectura, "PRAGMA journal_mode"),
		"el diario es del fichero, y el lector lo ve tal como está")
}

// TestFicheroInutilizable fija FR-028 y la fila 10 del contrato de errores: en
// la ruta de la base hay un fichero que no es una base de datos que se pueda
// leer, y la caché lo dice y lo deja intacto.
//
// No borrarlo ni rehacerlo es la parte que importa y la que se mide con el
// SHA-256: lo que hay ahí puede ser de otro programa o de otra persona, y
// perderlo sería peor que no poder usar la caché. Vale igual en los dos modos,
// porque el fichero no se toca ni cuando la invocación podría escribir.
func TestFicheroInutilizable(t *testing.T) {
	t.Parallel()

	for _, modo := range modosDeApertura {
		t.Run(modo.nombre, func(t *testing.T) {
			t.Parallel()

			directorio := t.TempDir()
			ruta := filepath.Join(directorio, ficheroDeLaBase)
			require.NoError(t, os.WriteFile(ruta,
				[]byte("esto no es una base de datos de SQLite y nadie va a borrarlo"), 0o600))

			antes := huella(t, ruta)

			cliente, err := New(t.Context(),
				append(slices.Clone(modo.opciones), ConDirectorio(directorio))...)

			require.Error(t, err)
			assert.Nil(t, cliente)
			assert.Equal(t, schema.ClaseInesperado, cli.Clasificar(err))
			assert.Equal(t, 1, cli.CodigoSalida(err))
			assert.Contains(t, err.Error(), ruta, "el mensaje nombra el fichero implicado")
			assert.Contains(t, err.Error(), "no se borra",
				"el mensaje dice que el fichero queda donde está (FR-028)")

			require.FileExists(t, ruta)
			assert.Equal(t, antes, huella(t, ruta),
				"un fichero que no sirve como base no se borra ni se rehace")
		})
	}
}

// pragmaTexto y pragmaEntero consultan un PRAGMA por la conexión del cliente.
// La consulta llega entera desde quien llama —y es siempre una constante del
// test— porque componerla aquí sería construir SQL con una cadena de fuera.
func pragmaTexto(t *testing.T, cliente *Cliente, consulta string) string {
	t.Helper()

	var valor string
	require.NoError(t, cliente.db.QueryRowContext(t.Context(), consulta).Scan(&valor))

	return valor
}

func pragmaEntero(t *testing.T, cliente *Cliente, consulta string) int64 {
	t.Helper()

	var valor int64
	require.NoError(t, cliente.db.QueryRowContext(t.Context(), consulta).Scan(&valor))

	return valor
}

// huella es el SHA-256 del fichero en hexadecimal, y es la forma de comprobar
// que algo no cambió ni un byte: que un fichero que no se puede usar no se
// borra ni se rehace (FR-028) y que una invocación de solo lectura deja la base
// idéntica (SC-003).
func huella(t *testing.T, ruta string) string {
	t.Helper()

	contenido, err := os.ReadFile(filepath.Clean(ruta))
	require.NoError(t, err)

	suma := sha256.Sum256(contenido)

	return hex.EncodeToString(suma[:])
}
