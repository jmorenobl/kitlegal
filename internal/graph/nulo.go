package graph

import (
	"context"

	"github.com/jmorenobl/kitlegal/internal/core"
)

// Nulo es el almacén del grafo de --no-graph: descarta todo lo que se le
// entrega. No resuelve ninguna ruta ni abre nada, de modo que con la bandera
// world.db y sus ficheros auxiliares quedan como estaban o siguen sin existir,
// y un entorno que no da ningún directorio no produce ningún fallo (FR-031;
// research.md D6).
type Nulo struct{}

// Nulo es un core.GraphStore.
var _ core.GraphStore = Nulo{}

// Apply devuelve nil con cualquier lote y cualquier contexto, sin tocar nada.
func (Nulo) Apply(context.Context, core.Lote) error {
	return nil
}
