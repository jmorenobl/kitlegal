//go:build integration && unix

// Las pruebas de este fichero son los casos de la matriz de integración en que
// world.db es un enlace simbólico (contracts/almacen-world-db.md §2 y §7;
// research.md D33, V47). Fuera de Windows, SQLite abre el destino y nombra y
// crea los auxiliares junto a él y con su nombre, así que lo que se afirma de
// ellos solo vale en Unix; precedente de la restricción:
// internal/disco/noregular_unix_test.go. Los preparativos son los de la matriz,
// en el otro fichero con la etiqueta integration.
package graph_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestIntegracionEnlace fija world.db que es un enlace simbólico por la API
// (FR-004; research.md V47): a un destino de otro directorio con un -wal
// huérfano junto a él, los tres verbos leen lo confirmado en el WAL, el destino
// y su -wal quedan con los mismos bytes —su -shm existe al terminar: la
// desviación declarada de §3— y junto al enlace no aparece nada.
func TestIntegracionEnlace(t *testing.T) {
	t.Parallel()

	t.Run("a un destino con un -wal huérfano", func(t *testing.T) {
		t.Parallel()

		raiz := t.TempDir()
		destino := filepath.Join(raiz, "otro", "grafo.db")
		require.NoError(t, os.Mkdir(filepath.Dir(destino), permisosDeDirectorio))
		copiaConElWAL(t, destino, elTerritorioEnElWAL, "", sufijoWAL, sufijoMemoriaCompartida)

		directorio := enlazadoA(t, raiz, destino)
		antes, junto := huellasDe(t, destino, destino+sufijoWAL), estadoDelArbol(t, directorio)

		assert.Equal(t, leidoTras(t, elBloque(), elTerritorio()), leerElGrafo(t, directorio), "lo confirmado en el WAL")
		assert.Equal(t, antes, huellasDe(t, destino, destino+sufijoWAL), "el destino y su -wal no cambian")
		assert.FileExists(t, destino+sufijoMemoriaCompartida, "el -shm del destino: la desviación declarada")
		assert.Equal(t, junto, estadoDelArbol(t, directorio), "junto al enlace no aparece ningún auxiliar")
	})
}

// enlazadoA crea bajo la raíz el directorio de la caché con world.db como enlace
// simbólico al destino, y lo devuelve.
func enlazadoA(t *testing.T, raiz, destino string) string {
	t.Helper()

	directorio := filepath.Join(raiz, "cache")
	require.NoError(t, os.Mkdir(directorio, permisosDeDirectorio))
	require.NoError(t, os.Symlink(destino, rutaEn(directorio)))

	return directorio
}
