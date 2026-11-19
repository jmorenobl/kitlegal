package evals

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

// TestMotivoSinConcordancia comprueba que el motivo fijo fuera de lo declarado,
// «fuera de lo que kitlegal cubre», no lleva verbo que concordar con el asunto:
// la línea cuenta como declarada con cualquier verbo delante, y no con una
// paráfrasis del motivo (ADR 0058, punto 1).
func TestMotivoSinConcordancia(t *testing.T) {
	t.Parallel()

	const cola = ". kitlegal no da sus plazos, órganos, vías ni escritos.\n"

	esperado := NoCubiertoEsperado{FueraDeLoDeclarado: true}

	casos := []struct {
		nombre   string
		linea    string
		esperado bool
	}{
		{"tal-cual", "⚠ NO CUBIERTO: el plazo para reclamar: fuera de lo que kitlegal cubre" + cola, true},
		{"plural", "⚠ NO CUBIERTO: la vía y el plazo para reclamar están fuera de lo que kitlegal cubre" + cola, true},
		{"queda", "⚠ NO CUBIERTO: el plazo para reclamar queda fuera de lo que kitlegal cubre" + cola, true},
		{"son-asuntos", "⚠ NO CUBIERTO: la vía y el órgano: son asuntos fuera de lo que kitlegal cubre" + cola, true},
		{"parafrasis", "⚠ NO CUBIERTO: el plazo para reclamar: fuera de la cobertura de kitlegal" + cola, false},
		{"sin-cola", "⚠ NO CUBIERTO: el plazo para reclamar: fuera de lo que kitlegal cubre.\n", false},
	}

	for _, caso := range casos {
		t.Run(caso.nombre, func(t *testing.T) {
			t.Parallel()

			assert.Equal(t, caso.esperado, esperado.declarada(ExtraerNoCubierto(caso.linea)))
		})
	}
}
