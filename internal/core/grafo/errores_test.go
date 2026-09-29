package grafo_test

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/jmorenobl/kitlegal/internal/core/grafo"
	"github.com/jmorenobl/kitlegal/internal/core/schema"
)

// TestRechazo fija el error con el que un lote entero no entra en el grafo del
// mundo (FR-024, FR-025; H7.1 FR-075; contracts/almacen-world-db.md §5 y §6):
// nombra la operación rechazada —el propio lote, un nodo por su id, un texto
// por su huella— y el motivo; declara la clase «inesperado», la de un lote
// rechazado en la entrega, y la conserva envuelto. Nunca repite el id de una
// Persona, que es donde el rechazo de FR-025 encuentra un documento de
// identidad, ni el cuerpo de un texto.
func TestRechazo(t *testing.T) {
	t.Parallel()

	const motivo = "el motivo"

	casos := []struct {
		nombre    string
		operacion schema.Operacion
		mensaje   string
	}{
		{"el lote", nil, "el lote: el motivo"},
		{
			"un nodo",
			schema.Nodo{ID: "ine:28074", Tipo: grafo.TipoMunicipio, Datos: map[string]any{grafo.DatoNombre: "Leganés"}},
			`el nodo "ine:28074": el motivo`,
		},
		{
			"un nodo con un control en el id",
			schema.Nodo{ID: "a\x00b", Tipo: grafo.TipoNorma},
			`el nodo "a\x00b": el motivo`,
		},
		{
			"una Persona, sin su id",
			schema.Nodo{ID: "12345678Z", Tipo: grafo.TipoPersona, Datos: map[string]any{"dni": "12345678Z"}},
			`el nodo de tipo "Persona": el motivo`,
		},
		{
			"un texto, sin su cuerpo",
			schema.Texto{Huella: "sha256:ab", Cuerpo: "Artículo 21. Obligación de resolver."},
			`el texto "sha256:ab": el motivo`,
		},
	}

	for _, caso := range casos {
		t.Run(caso.nombre, func(t *testing.T) {
			t.Parallel()

			rechazo := &grafo.Rechazo{Operacion: caso.operacion, Motivo: motivo}
			require.EqualError(t, rechazo, caso.mensaje)

			envuelto := fmt.Errorf("grafo: el lote no entra en world.db: %w", rechazo)

			var conClase schema.ConClase
			require.ErrorAs(t, envuelto, &conClase)
			assert.Equal(t, schema.ClaseInesperado, conClase.Clase())

			var recibido *grafo.Rechazo
			require.ErrorAs(t, envuelto, &recibido)
			assert.Same(t, rechazo, recibido, "quien lo recibe llega a la operación y al motivo")
		})
	}
}
