package skills_test

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/jmorenobl/kitlegal/internal/skills"
)

// destinoDeLosEnlacesDePrueba es el destino literal de todo enlace de scripts/ de
// una skill, escrito aquí tal cual lo fija data-model §3 (research.md D5).
const destinoDeLosEnlacesDePrueba = "../../../bin/instalado/kitlegal"

// TestEnlacesEsperados fija EnlacesEsperados (data-model §3; research.md D4 y
// D5; FR-035, FR-036): un enlace por cada applet de kitlegal-applets, en el
// orden de la declaración, con el nombre del applet y el destino literal
// ../../../bin/instalado/kitlegal; y ninguno si la skill no declara applets,
// aunque declare referencias.
func TestEnlacesEsperados(t *testing.T) {
	t.Parallel()

	casos := []struct {
		nombre     string
		deKitlegal skills.DeclaracionDeKitlegal
		enlaces    []skills.Enlace
	}{
		{
			nombre:     "un-applet",
			deKitlegal: skills.DeclaracionDeKitlegal{Applets: []string{"boe"}, Referencias: []string{"normas"}},
			enlaces:    []skills.Enlace{{Nombre: "boe", Destino: destinoDeLosEnlacesDePrueba}},
		},
		{
			nombre:     "varios-applets-en-el-orden-declarado",
			deKitlegal: skills.DeclaracionDeKitlegal{Applets: []string{"placsp", "boe"}},
			enlaces: []skills.Enlace{
				{Nombre: "placsp", Destino: destinoDeLosEnlacesDePrueba},
				{Nombre: "boe", Destino: destinoDeLosEnlacesDePrueba},
			},
		},
		{
			nombre:     "solo-referencias",
			deKitlegal: skills.DeclaracionDeKitlegal{Referencias: []string{"normas"}},
		},
		{
			nombre: "nada-declarado",
		},
	}

	for _, caso := range casos {
		t.Run(caso.nombre, func(t *testing.T) {
			t.Parallel()

			assert.Equal(t, caso.enlaces, skills.EnlacesEsperados(caso.deKitlegal))
		})
	}
}
