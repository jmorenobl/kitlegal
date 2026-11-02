package app

import (
	"slices"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/jmorenobl/kitlegal/internal/core/instalacion"
)

// skillsDeC991808 son los nombres de skills/ en c991808, la última entrega con
// los nombres de antes de H32.
var skillsDeC991808 = []string{"boe-legislacion", "jurisprudencia", "legal-core"}

// TestSkillsRetiradas fija la lista de los nombres retirados (FR-025;
// contracts/instalacion-retiradas.md §1; data-model §6): cada skill de
// c991808 que el binario ya no empotra está en ella, ninguna que empotra lo
// está y cada heredera es una skill que el binario empotra. Y la lista es la
// del contrato, cada nombre retirado con la skill que hereda sus preguntas.
func TestSkillsRetiradas(t *testing.T) {
	t.Parallel()

	empotradas, err := skillsEmpotradas()
	require.NoError(t, err)

	empotra := func(nombre string) bool {
		return slices.ContainsFunc(empotradas, func(skill instalacion.SkillEmpotrada) bool { return skill.Nombre == nombre })
	}

	for _, nombre := range skillsDeC991808 {
		if !empotra(nombre) {
			assert.Contains(t, skillsRetiradas, nombre, "%s ya no se empotra y falta en la lista de los nombres retirados", nombre)
		}
	}

	for nombre, herederas := range skillsRetiradas {
		assert.False(t, empotra(nombre), "%s está en la lista de los nombres retirados y el binario la empotra", nombre)
		assert.NotEmpty(t, herederas, "%s no tiene ninguna heredera", nombre)

		for _, heredera := range herederas {
			assert.True(t, empotra(heredera), "%s, heredera de %s, no es una skill que el binario empotre", heredera, nombre)
		}
	}

	assert.Equal(t, instalacion.Retiradas{
		"boe-legislacion": {"normativa"},
		"legal-core":      {"plazos-y-territorio"},
	}, skillsRetiradas, "la lista de contracts/instalacion-retiradas.md §1")
}
