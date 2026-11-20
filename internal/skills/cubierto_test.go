package skills_test

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/jmorenobl/kitlegal/internal/skills"
)

// Las marcas de la región de lo que cubre la base, cada una con su salto
// (contracts/datos-y-regiones.md §3 de H33).
const (
	inicioDeLoCubierto = "<!-- lo que cubre la base: lo genera make skills-sync -->\n"
	finDeLoCubierto    = "<!-- fin de lo que cubre la base -->\n"
)

// lineasDeLoCubierto son las de la región de lo cubierto de la skill de entrada
// con los datos del repositorio: una por fila de lo cubierto de
// data/cobertura.yaml, en su orden (contracts/skills.md §1 de H33).
var lineasDeLoCubierto = []string{
	"- **Normativa**: leer y citar cualquier norma consolidada del BOE, estatal o autonómica, de cualquier materia y " +
		"en cualquier fecha.",
	"- **Jurisprudencia**: preparar la consulta exacta de una sentencia, que la persona busca con su navegador, y " +
		"cotejar el documento que trae. No consulta ningún buscador ni dice si una sentencia existe.",
	"- **Plazos**: solo los del procedimiento administrativo común, los del acceso a la información pública y la " +
		"interposición del recurso contencioso-administrativo, con los festivos de los municipios que cuentan. Lo que " +
		"la ley regula aparte de ese procedimiento lo declara `plazos-y-territorio` con su línea.",
	"- **Territorio**: la provincia, la comunidad, el régimen, el DIR3 del ayuntamiento y los boletines de un municipio.",
}

// TestRegionDeLoCubierto fija la región de lo que cubre la base (FR-003 y FR-007;
// contracts/datos-y-regiones.md §3 y §4; contracts/controles.md §1 de H33) sobre
// una copia de consulta-juridica, plazos-y-territorio y los datos del
// repositorio: la región de la skill de entrada da las cuatro líneas de lo
// cubierto en su orden y la sincronía no cambia nada. Cada mutante, la medida
// pasada de su umbral de cero (ADR 0029), da los defectos exactos de cada skill
// de la copia, que con alguno no lleva nada regenerado: una materia que no es el
// vertical de ninguna norma, el de la materia, que da la sincronía y no el
// esquema; sin data/cobertura.yaml, en la skill de entrada el de su región y el
// de su plantilla, y en plazos-y-territorio el de su plantilla y el de su región
// cobertura_por_materia; y el motivo cambiado en la prosa de la skill de
// entrada, el del motivo.
func TestRegionDeLoCubierto(t *testing.T) {
	t.Parallel()

	t.Run("repositorio", func(t *testing.T) {
		t.Parallel()

		raiz := copiaDeLaEntradaYDePlazos(t)

		regenerado := regenerar(t, raiz, nil)
		require.Empty(t, presentarDefectos(regenerado.Defectos()))

		entrada := skillDe(t, regenerado, "consulta-juridica")
		assert.Equal(t, lineasDeLoCubierto, lineasEntreLasMarcasDeLoCubierto(t, string(entrada.Contenido)),
			"la región da una línea por fila de lo cubierto, en su orden")

		derivas, err := skills.Comparar(raiz, regenerado)
		require.NoError(t, err)
		assert.Empty(t, textosDeLosFallos(derivas), "la sincronía no cambia nada")
	})

	casos := []struct {
		nombre   string
		mutante  func(t *testing.T, raiz string)
		defectos map[string][]string
	}{
		{
			nombre: "materia-que-no-es-un-vertical",
			mutante: func(t *testing.T, raiz string) {
				t.Helper()

				cambiarFicheroDePrueba(t, filepath.Join(raiz, "data", "cobertura.yaml"), "    materia: administrativo\n",
					"    materia: Laboral\n")
			},
			defectos: map[string][]string{
				"consulta-juridica": {
					"consulta-juridica: data/cobertura.yaml: la fila de lo cubierto de la capacidad «plazos» lleva la " +
						"materia «Laboral», que no es el vertical de ninguna norma de data/normas.yaml",
				},
			},
		},
		{
			nombre: "sin-data-cobertura",
			mutante: func(t *testing.T, raiz string) {
				t.Helper()

				require.NoError(t, os.Remove(filepath.Join(raiz, "data", "cobertura.yaml")))
			},
			defectos: map[string][]string{
				"consulta-juridica": {
					"consulta-juridica: SKILL.md: la región lo que cubre la base sin data/cobertura.yaml",
					"consulta-juridica: SKILL.md: la plantilla ⚠ NO CUBIERTO: sin data/cobertura.yaml",
				},
				"plazos-y-territorio": {
					"plazos-y-territorio: SKILL.md: la plantilla ⚠ NO CUBIERTO: sin data/cobertura.yaml",
					"plazos-y-territorio: references/administrativo.md: la región cobertura_por_materia sin " +
						"data/cobertura.yaml",
				},
			},
		},
		{
			nombre: "motivo-cambiado-en-la-prosa",
			mutante: func(t *testing.T, raiz string) {
				t.Helper()

				cambiarFicheroDePrueba(t, rutaDeSkill(raiz, "consulta-juridica", "SKILL.md"),
					"«fuera\nde lo que kitlegal cubre»", "«fuera\nde lo que cubre kitlegal»")
			},
			defectos: map[string][]string{
				"consulta-juridica": {
					"consulta-juridica: SKILL.md: la prosa no dice el motivo de lo no declarado, «fuera de lo que " +
						"kitlegal cubre»",
				},
			},
		},
	}

	for _, caso := range casos {
		t.Run(caso.nombre, func(t *testing.T) {
			t.Parallel()

			raiz := copiaDeLaEntradaYDePlazos(t)
			caso.mutante(t, raiz)

			regenerado := regenerar(t, raiz, nil)

			nombres := make([]string, 0, len(regenerado.Skills))

			for _, skill := range regenerado.Skills {
				nombres = append(nombres, skill.Nombre)

				esperados := caso.defectos[skill.Nombre]
				assert.Equal(t, esperados, presentarDefectos(skill.Defectos), "los defectos de %s", skill.Nombre)

				if len(esperados) > 0 {
					assert.Nil(t, skill.Contenido, "%s, con defectos, no lleva nada regenerado", skill.Nombre)
				}
			}

			assert.Equal(t, []string{"consulta-juridica", "plazos-y-territorio"}, nombres, "las skills de la copia")
		})
	}
}

// copiaDeLaEntradaYDePlazos es la copia de plazos-y-territorio y de los datos de
// sus referencias de copiaDePlazosYTerritorio con la skill de entrada,
// consulta-juridica, que no declara applets: la de la región de lo cubierto y la
// de la lista de los paquetes de la base.
func copiaDeLaEntradaYDePlazos(t *testing.T) string {
	t.Helper()

	raiz := copiaDePlazosYTerritorio(t)
	require.NoError(t, os.CopyFS(rutaDeSkill(raiz, "consulta-juridica"), os.DirFS("../../skills/consulta-juridica")))

	return raiz
}

// skillDe es la skill regenerada de nombre.
func skillDe(t *testing.T, regenerado skills.Regenerado, nombre string) skills.SkillRegenerada {
	t.Helper()

	for _, skill := range regenerado.Skills {
		if skill.Nombre == nombre {
			return skill
		}
	}

	require.Failf(t, "falta la skill", "lo regenerado no tiene %s", nombre)

	return skills.SkillRegenerada{}
}

// lineasEntreLasMarcasDeLoCubierto son las líneas, sin su salto, de lo que hay
// entre las marcas de la región de lo cubierto del SKILL.md, que las lleva una
// vez cada una.
func lineasEntreLasMarcasDeLoCubierto(t *testing.T, contenido string) []string {
	t.Helper()

	require.Equal(t, 1, strings.Count(contenido, inicioDeLoCubierto), "una marca de inicio de lo cubierto")
	require.Equal(t, 1, strings.Count(contenido, finDeLoCubierto), "una marca de fin de lo cubierto")

	_, tras, _ := strings.Cut(contenido, inicioDeLoCubierto)
	entre, _, _ := strings.Cut(tras, finDeLoCubierto)

	if entre == "" {
		return nil
	}

	return strings.Split(strings.TrimSuffix(entre, "\n"), "\n")
}
