package evals

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestActivadaPor fija Eval.activadaPor (contracts/evals.md §3 de H33; FR-015;
// ADR 0058, punto 7): la activación de una skill de tambien_la_resuelven cuenta
// como la de la skill de la carpeta, y sin ella, o con la de una skill que la
// eval no nombra, solo cuenta la de la carpeta. La eval que lo declara se lee de
// un documento con la clave, como la 17 de normativa, así que el esquema de eval
// y su lector estricto la admiten; y juzgarEnModo fija con ella
// ResultadoDeEval.Activada, de la que salen el veredicto y sin_activar.
func TestActivadaPor(t *testing.T) {
	t.Parallel()

	const (
		skillDeLaCarpeta = "normativa"
		otra             = "plazos-y-territorio"
	)

	conLaOtra, err := LeerEval("17-ltaibg-plazo-por-materia.yaml", []byte(
		"pregunta: \"¿En qué plazo hay que resolver una solicitud de acceso a la información pública según la Ley "+
			"19/2013?\"\n"+
			"activa: true\n"+
			"tambien_la_resuelven:\n"+
			"  - "+otra+"\n"+
			"comandos:\n"+
			"  - applet: boe\n"+
			"    norma: BOE-A-2013-12887\n"+
			"    bloque: a20\n"+
			"citas:\n"+
			"  - norma: BOE-A-2013-12887\n"+
			"    bloque: a20\n"))
	require.NoError(t, err)
	require.Equal(t, []string{otra}, conLaOtra.TambienLaResuelven, "la eval se lee con tambien_la_resuelven")

	sinLaOtra := conLaOtra
	sinLaOtra.TambienLaResuelven = nil

	casos := []struct {
		nombre    string
		eval      Eval
		activadas []string
		activada  bool
	}{
		{nombre: "la-de-la-carpeta", eval: sinLaOtra, activadas: []string{"consulta-juridica", skillDeLaCarpeta}, activada: true},
		{nombre: "la-de-la-carpeta-con-la-otra-declarada", eval: conLaOtra, activadas: []string{skillDeLaCarpeta}, activada: true},
		{nombre: "la-otra-sin-declararla", eval: sinLaOtra, activadas: []string{"consulta-juridica", otra}},
		{nombre: "la-otra-declarada", eval: conLaOtra, activadas: []string{"consulta-juridica", otra}, activada: true},
		{nombre: "una-que-no-declara", eval: conLaOtra, activadas: []string{"consulta-juridica", "jurisprudencia"}},
		{nombre: "ninguna", eval: conLaOtra},
	}

	for _, caso := range casos {
		t.Run(caso.nombre, func(t *testing.T) {
			t.Parallel()

			sesion := Sesion{SkillsActivadas: caso.activadas, Terminada: true}

			assert.Equal(t, caso.activada, caso.eval.activadaPor(sesion, skillDeLaCarpeta))
			assert.Equal(t, caso.activada, juzgarEnModo(caso.eval, sesion, skillDeLaCarpeta, ModoOrden).Activada,
				"juzgarEnModo fija Activada con activadaPor")
		})
	}
}

// TestSkillsNombradasEnTambienLaResuelven fija que ComprobarLasSkillsNombradas
// mira también tambien_la_resuelven (contracts/evals.md §3 de H33; FR-015): un
// nombre que es una skill de skills/ no da defecto, y uno que no lo es, uno con
// su fichero y su clave, escrito aquí a mano, detrás de los de se_activan y de
// no_se_activan de la misma eval.
func TestSkillsNombradasEnTambienLaResuelven(t *testing.T) {
	t.Parallel()

	const (
		fichero  = "17-ltaibg-plazo-por-materia.yaml"
		retirada = "legal-core"
	)

	skillsDelRepositorio, err := directoriosDe(os.DirFS(raizDelRepositorio), directorioDeLasSkills)
	require.NoError(t, err)
	require.Contains(t, skillsDelRepositorio, "plazos-y-territorio", "premisa: plazos-y-territorio es una skill")
	require.NotContains(t, skillsDelRepositorio, retirada, "premisa: %s es un nombre retirado", retirada)

	assert.Empty(t, ComprobarLasSkillsNombradas(
		[]Eval{{Fichero: fichero, TambienLaResuelven: []string{"plazos-y-territorio"}}}, skillsDelRepositorio),
		"una skill de skills/ en tambien_la_resuelven no es un defecto")

	assert.Equal(t, []DefectoDelConjunto{
		{Regla: "skills nombradas", Mensaje: fichero + ": no_se_activan nombra boe-legislacion, que no es una skill de skills/"},
		{Regla: "skills nombradas", Mensaje: fichero + ": tambien_la_resuelven nombra " + retirada + ", que no es una skill de skills/"},
	}, ComprobarLasSkillsNombradas([]Eval{{
		Fichero:            fichero,
		NoSeActivan:        []string{"boe-legislacion"},
		TambienLaResuelven: []string{"plazos-y-territorio", retirada},
	}}, skillsDelRepositorio))
}
