package evals

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

// TestActivadaPor comprueba que la activación de una skill de
// tambien_la_resuelven cuenta como la de la skill de la eval, y que sin ella
// solo cuenta la de la eval (ADR 0058, punto 7).
func TestActivadaPor(t *testing.T) {
	t.Parallel()

	conLaOtra := Eval{TambienLaResuelven: []string{"plazos-y-territorio"}}

	casos := []struct {
		nombre    string
		eval      Eval
		activadas []string
		esperado  bool
	}{
		{"la-de-la-eval", Eval{}, []string{"consulta-juridica", "normativa"}, true},
		{"otra-sin-declararla", Eval{}, []string{"consulta-juridica", "plazos-y-territorio"}, false},
		{"otra-declarada", conLaOtra, []string{"consulta-juridica", "plazos-y-territorio"}, true},
		{"otra-no-declarada", conLaOtra, []string{"consulta-juridica", "jurisprudencia"}, false},
		{"ninguna", conLaOtra, nil, false},
	}

	for _, caso := range casos {
		t.Run(caso.nombre, func(t *testing.T) {
			t.Parallel()

			sesion := Sesion{SkillsActivadas: caso.activadas}
			assert.Equal(t, caso.esperado, caso.eval.activadaPor(sesion, "normativa"))
		})
	}
}

// TestSkillsNombradasEnTambienLaResuelven comprueba que un nombre de
// tambien_la_resuelven que no es una skill de skills/ es un defecto del conjunto.
func TestSkillsNombradasEnTambienLaResuelven(t *testing.T) {
	t.Parallel()

	skills := []string{"consulta-juridica", "jurisprudencia", "normativa", "plazos-y-territorio"}

	assert.Empty(t, ComprobarLasSkillsNombradas(
		[]Eval{{Fichero: "17.yaml", TambienLaResuelven: []string{"plazos-y-territorio"}}}, skills))

	defectos := ComprobarLasSkillsNombradas(
		[]Eval{{Fichero: "17.yaml", TambienLaResuelven: []string{"legal-core"}}}, skills)
	if assert.Len(t, defectos, 1) {
		assert.Contains(t, defectos[0].Mensaje, "tambien_la_resuelven nombra legal-core")
	}
}
