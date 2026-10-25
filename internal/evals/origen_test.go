package evals

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

// TestOrigenRecibidoSinLaDescripcion fija que el bloque descripcion de la
// decisión (contracts/juez-de-las-descripciones.md §2 de H31) no pasa al
// bloque origen de informe.json, que es el de H30 (contracts/informe-del-job.md
// §2 de H30), y que la copia no cambia la decisión.
func TestOrigenRecibidoSinLaDescripcion(t *testing.T) {
	t.Parallel()

	descripcion := &descripcionDeLaDecision{Skill: "legal-core", Huella: strings.Repeat("d", 64), Juzgada: true}
	decision := &origenDeLasEvals{
		Modo:        origenConReutilizacion,
		Evals:       []origenDeEval{{Eval: "01-plazo.yaml", Medida: true, Motivo: motivoDescripcion}},
		Descripcion: descripcion,
	}

	origen := InformeAEscribir{decision: decision}.origenRecibido(nil)

	assert.Nil(t, origen.Descripcion, "el origen de informe.json no lleva el bloque descripcion")
	assert.Equal(t, 1, origen.EvalsMedidas, "el origen lleva las evals de la decisión")
	assert.Same(t, descripcion, decision.Descripcion, "la decisión conserva su bloque descripcion")
}
