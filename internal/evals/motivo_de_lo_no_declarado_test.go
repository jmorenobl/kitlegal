package evals

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

// TestDeclaradaFueraDeLoDeclarado comprueba que, fuera de lo declarado, cuenta
// como declarada cualquier línea ⚠ NO CUBIERTO:, también con el motivo o la cola
// escritos con otras palabras, y que sin la línea no (ADR 0058, punto 1). Las dos
// primeras son las del despacho 44568601898; en lo regulado aparte sigue
// pidiéndose la norma y la cita de su precepto.
func TestDeclaradaFueraDeLoDeclarado(t *testing.T) {
	t.Parallel()

	fuera := NoCubiertoEsperado{FueraDeLoDeclarado: true}
	aparte := NoCubiertoEsperado{
		Norma: "BOE-A-2015-10565", Precepto: CitaEsperada{Norma: "BOE-A-2015-10565", Bloque: "daprimera"},
	}

	casos := []struct {
		nombre    string
		esperado  NoCubiertoEsperado
		respuesta string
		declarada bool
	}{
		{"cola-reescrita", fuera, "⚠ NO CUBIERTO: a quién reclamar y qué escrito presentar. Está fuera de lo que " +
			"kitlegal cubre, que no da plazos, órganos, vías ni escritos para esto.\n", true},
		{"motivo-propio", fuera, "⚠ NO CUBIERTO: el día en que vence el plazo: es un plazo laboral. kitlegal no da " +
			"sus plazos, órganos, vías ni escritos.\n", true},
		{"en-negrita", fuera, "**⚠️ NO CUBIERTO:** la vía: fuera de lo que kitlegal cubre.\n", true},
		{"sin-la-linea", fuera, "No está cubierto: kitlegal no da sus plazos, órganos, vías ni escritos.\n", false},
		{"aparte-sin-la-cita", aparte, "⚠ NO CUBIERTO: los tributos: fuera de lo que kitlegal cubre.\n", false},
		{"aparte-con-la-cita", aparte, "⚠ NO CUBIERTO: los tributos lo regula aparte la ley (DA 1ª.2 LPAC " +
			"[BOE-A-2015-10565, bloque daprimera]). kitlegal no da sus plazos, órganos, vías ni escritos.\n", true},
	}

	for _, caso := range casos {
		t.Run(caso.nombre, func(t *testing.T) {
			t.Parallel()

			assert.Equal(t, caso.declarada, caso.esperado.declarada(ExtraerNoCubierto(caso.respuesta)))
		})
	}
}
