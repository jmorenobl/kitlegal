package evals

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

// TestExtraerCitas fija la extracción de la parte mecánica de una cita,
// [<identificador>, bloque <id de bloque>], con la expresión del contrato
// skill-boe-legislacion §3: dentro de la redacción que la rodea, con el id del
// bloque delimitado por el corchete de cierre aunque termine en punto, varias en
// una línea en el orden en que aparecen, y ninguna sin los dos corchetes o sin
// «bloque» (FR-008, FR-072; US4, escenario 3).
func TestExtraerCitas(t *testing.T) {
	t.Parallel()

	casos := []struct {
		nombre    string
		respuesta string
		citas     []Cita
	}{
		{
			nombre: "forma-fija-con-redaccion",
			respuesta: "Según el art. 21 de la Ley 39/2015 [BOE-A-2015-10565, bloque a21], la Administración " +
				"está obligada a dictar resolución expresa.",
			citas: []Cita{{Norma: normaDeLasTrazas, Bloque: "a21"}},
		},
		{
			nombre:    "bloque-que-termina-en-punto",
			respuesta: "El art. 85 bis de la Ley 7/1985 se cita [BOE-A-1985-5392, bloque a85bis.] con el punto de su id.",
			citas:     []Cita{{Norma: "BOE-A-1985-5392", Bloque: "a85bis."}},
		},
		{
			nombre: "sin-corchete",
			respuesta: "Sin corchetes: BOE-A-2015-10565, bloque a21.\n" +
				"Sin el de cierre: [BOE-A-2015-10565, bloque a22\n" +
				"Sin el de abrir: BOE-A-2015-10565, bloque a23]",
		},
		{
			nombre:    "sin-bloque",
			respuesta: "Sin la palabra: [BOE-A-2015-10565, a21], [BOE-A-2015-10565 a22] y [BOE-A-2015-10565, art. a23].",
		},
		{
			nombre: "varias-en-una-linea",
			respuesta: "El art. 21 [BOE-A-2015-10565, bloque a21] y el art. 22 [BOE-A-2015-10565, bloque a22], " +
				"y el expediente del contrato menor [BOE-A-2017-12902, bloque a1-30].",
			citas: []Cita{
				{Norma: normaDeLasTrazas, Bloque: "a21"},
				{Norma: normaDeLasTrazas, Bloque: "a22"},
				{Norma: "BOE-A-2017-12902", Bloque: "a1-30"},
			},
		},
	}

	for _, caso := range casos {
		t.Run(caso.nombre, func(t *testing.T) {
			t.Parallel()

			assert.Equal(t, caso.citas, ExtraerCitas(caso.respuesta))
		})
	}
}
