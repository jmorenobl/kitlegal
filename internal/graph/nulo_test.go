package graph_test

import (
	"context"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/jmorenobl/kitlegal/internal/cache"
	"github.com/jmorenobl/kitlegal/internal/core"
	"github.com/jmorenobl/kitlegal/internal/core/schema"
	"github.com/jmorenobl/kitlegal/internal/graph"
)

// TestNulo fija el almacén nulo de --no-graph (FR-031; research.md D6;
// contracts/almacen-world-db.md §1): es un core.GraphStore que devuelve nil sin
// resolver ninguna ruta ni abrir nada, con cualquier lote y cualquier contexto.
// Lo mide con el entorno que haría fallar la ubicación —la variable declarada
// y vacía, sin directorio de la cuenta—, que un almacén que resolviera la ruta
// convertiría en un error, y con la variable que nombra un directorio que no
// existe, que uno que abriera algo crearía.
//
// No declara t.Parallel(): usa t.Setenv, y el entorno es lo que mide.
func TestNulo(t *testing.T) {
	raiz := t.TempDir()
	t.Setenv("HOME", "")
	t.Setenv("USERPROFILE", "")

	lotes := map[string]core.Lote{
		"un lote que el almacén aceptaría": {
			Fuente:        "boe.legislacion-consolidada",
			URL:           "https://www.boe.es/buscar/act.php?id=BOE-A-2015-10565",
			FechaConsulta: "2026-09-28T12:00:00Z",
			Vigencia:      7 * 24 * time.Hour,
			Operaciones:   []schema.Operacion{schema.Nodo{ID: "ine:28074", Tipo: "Municipio"}},
		},
		"un lote que el almacén rechazaría": {
			Operaciones: []schema.Operacion{schema.Arista{Origen: "a", Relacion: "r", Destino: "b"}},
		},
		"el lote vacío": {},
	}

	entornos := map[string]string{
		"con la variable declarada y vacía":           "",
		"con la variable que nombra lo que no existe": filepath.Join(raiz, "cache"),
	}

	var almacen core.GraphStore = graph.Nulo{}

	for entorno, variable := range entornos {
		t.Run(entorno, func(t *testing.T) {
			t.Setenv(cache.VariableDirectorio, variable)

			terminado, cancela := context.WithCancel(t.Context())
			cancela()

			for nombre, lote := range lotes {
				require.NoError(t, almacen.Apply(t.Context(), lote), "%s: el almacén nulo descarta sin fallar", nombre)
				require.NoError(t, almacen.Apply(terminado, lote), "%s: también con el contexto terminado", nombre)
			}

			entradas, err := os.ReadDir(raiz)
			require.NoError(t, err)
			assert.Empty(t, entradas, "el almacén nulo no crea nada")
		})
	}
}
