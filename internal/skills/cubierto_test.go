package skills_test

import (
	"io/fs"
	"os"
	"path"
	"path/filepath"
	"regexp"
	"slices"
	"strings"
	"testing"
	"unicode"
	"unicode/utf8"

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

// lineasDeLoCubiertoDelRepositorio son las de la región de lo cubierto de la
// skill de entrada con los datos del repositorio (contracts/skills.md §1 de H33;
// FR-016 de H45): una por fila de lo cubierto de data/cobertura.yaml, en su
// orden, «- **<capacidad>**: <alcance>», con la capacidad con su primera letra en
// mayúscula y sin su materia.
func lineasDeLoCubiertoDelRepositorio(t *testing.T) []string {
	t.Helper()

	cobertura, err := skills.LeerCobertura(leerDatosDelRepositorio(t, "cobertura.yaml"))
	require.NoError(t, err)

	lineas := make([]string, 0, len(cobertura.Cubierto))

	for _, fila := range cobertura.Cubierto {
		primera, tamano := utf8.DecodeRuneInString(fila.Capacidad)
		lineas = append(lineas, "- **"+string(unicode.ToUpper(primera))+fila.Capacidad[tamano:]+"**: "+fila.Alcance)
	}

	return lineas
}

// TestRegionDeLoCubierto fija la región de lo que cubre la base (FR-003 y FR-007;
// contracts/datos-y-regiones.md §3 y §4; contracts/controles.md §1 de H33) sobre
// una copia de consulta-juridica, de cada skill con una referencia de un paquete
// y de los datos del repositorio: la región de la skill de entrada da una línea
// por fila de lo cubierto de data/cobertura.yaml, en su orden, y la sincronía no
// cambia nada. Cada mutante, la medida pasada de su umbral de cero (ADR 0029), da
// los defectos exactos de cada skill de la copia, que con alguno no lleva nada
// regenerado: una materia que no es el vertical de ninguna norma, el de la
// materia, que da la sincronía y no el esquema; sin data/cobertura.yaml, en la
// skill de entrada el de su región y el de su plantilla, y en cada otra skill el
// de su plantilla, si la tiene, y el de la región cobertura_por_materia de cada
// referencia que la lleva, en orden de nombre (FR-016 de H45); sin la marca de
// fin de la región en la skill de entrada, el de sus marcas que no van en pareja;
// y el motivo cambiado en la prosa de la skill de entrada, el del motivo.
func TestRegionDeLoCubierto(t *testing.T) {
	t.Parallel()

	t.Run("repositorio", func(t *testing.T) {
		t.Parallel()

		raiz, _ := copiaDeLaEntradaYDeLosPaquetes(t)

		regenerado := regenerar(t, raiz, nil)
		require.Empty(t, presentarDefectos(regenerado.Defectos()))

		entrada := skillDe(t, regenerado, skillDeEntrada)
		assert.Equal(t, lineasDeLoCubiertoDelRepositorio(t), lineasEntreLasMarcasDeLoCubierto(t, string(entrada.Contenido)),
			"la región da una línea por fila de lo cubierto, en su orden")

		derivas, err := skills.Comparar(raiz, regenerado)
		require.NoError(t, err)
		assert.Empty(t, textosDeLosFallos(derivas), "la sincronía no cambia nada")
	})

	casos := []struct {
		nombre   string
		mutante  func(t *testing.T, raiz string)
		defectos func(t *testing.T, raiz string, nombres []string) map[string][]string
	}{
		{
			nombre: "materia-que-no-es-un-vertical",
			mutante: func(t *testing.T, raiz string) {
				t.Helper()

				cambiarFicheroDePrueba(t, filepath.Join(raiz, "data", "cobertura.yaml"), "    materia: administrativo\n",
					"    materia: Laboral\n")
			},
			defectos: soloEnLaEntrada(
				"consulta-juridica: data/cobertura.yaml: la fila de lo cubierto de la capacidad «plazos» lleva la " +
					"materia «Laboral», que no es el vertical de ninguna norma de data/normas.yaml",
			),
		},
		{
			nombre: "sin-data-cobertura",
			mutante: func(t *testing.T, raiz string) {
				t.Helper()

				require.NoError(t, os.Remove(filepath.Join(raiz, "data", "cobertura.yaml")))
			},
			defectos: defectosSinLaCobertura,
		},
		{
			nombre: "marcas-sin-pareja",
			mutante: func(t *testing.T, raiz string) {
				t.Helper()

				cambiarFicheroDePrueba(t, rutaDeSkill(raiz, "consulta-juridica", "SKILL.md"),
					"<!-- fin de lo que cubre la base -->\n", "")
			},
			defectos: soloEnLaEntrada(
				"consulta-juridica: SKILL.md: las marcas de lo que cubre la base no van en pareja: falta la de fin",
			),
		},
		{
			nombre: "motivo-cambiado-en-la-prosa",
			mutante: func(t *testing.T, raiz string) {
				t.Helper()

				cambiarFicheroDePrueba(t, rutaDeSkill(raiz, "consulta-juridica", "SKILL.md"),
					"«fuera\nde lo que kitlegal cubre»", "«fuera\nde lo que cubre kitlegal»")
			},
			defectos: soloEnLaEntrada(
				"consulta-juridica: SKILL.md: la prosa no dice el motivo de lo no declarado, «fuera de lo que " +
					"kitlegal cubre»",
			),
		},
	}

	for _, caso := range casos {
		t.Run(caso.nombre, func(t *testing.T) {
			t.Parallel()

			raiz, deLaCopia := copiaDeLaEntradaYDeLosPaquetes(t)
			caso.mutante(t, raiz)
			defectos := caso.defectos(t, raiz, deLaCopia)

			regenerado := regenerar(t, raiz, nil)

			nombres := make([]string, 0, len(regenerado.Skills))

			for _, skill := range regenerado.Skills {
				nombres = append(nombres, skill.Nombre)

				esperados := defectos[skill.Nombre]
				assert.Equal(t, esperados, presentarDefectos(skill.Defectos), "los defectos de %s", skill.Nombre)

				if len(esperados) > 0 {
					assert.Nil(t, skill.Contenido, "%s, con defectos, no lleva nada regenerado", skill.Nombre)
				}
			}

			assert.Equal(t, deLaCopia, nombres, "las skills de la copia")
		})
	}
}

// skillDeEntrada es la skill que lleva la región de lo cubierto.
const skillDeEntrada = "consulta-juridica"

// soloEnLaEntrada son los defectos de un mutante que solo da defectos en la
// skill de entrada: los dados, en su orden.
func soloEnLaEntrada(defectos ...string) func(*testing.T, string, []string) map[string][]string {
	return func(*testing.T, string, []string) map[string][]string {
		return map[string][]string{skillDeEntrada: defectos}
	}
}

// defectosSinLaCobertura son los defectos de la copia de raíz, con las skills
// de nombres, sin data/cobertura.yaml: en la skill de entrada, el de su región y
// el de su plantilla; en cada otra, el de su plantilla, si la tiene, y detrás el
// de la región cobertura_por_materia de cada references/*.md que lleva su marca,
// en orden de nombre (FR-016 de H45).
func defectosSinLaCobertura(t *testing.T, raiz string, nombres []string) map[string][]string {
	t.Helper()

	defectos := map[string][]string{
		skillDeEntrada: {
			skillDeEntrada + ": SKILL.md: la región lo que cubre la base sin data/cobertura.yaml",
			skillDeEntrada + ": SKILL.md: la plantilla ⚠ NO CUBIERTO: sin data/cobertura.yaml",
		},
	}

	for _, skill := range nombres {
		if skill == skillDeEntrada {
			continue
		}

		var deLaSkill []string

		if conLaPlantillaDeNoCubierto(leerFicheroDePrueba(t, rutaDeSkill(raiz, skill, "SKILL.md"))) {
			deLaSkill = append(deLaSkill, skill+": SKILL.md: la plantilla ⚠ NO CUBIERTO: sin data/cobertura.yaml")
		}

		referencias, err := filepath.Glob(rutaDeSkill(raiz, skill, "references", "*.md"))
		require.NoError(t, err)

		for _, referencia := range referencias {
			if strings.Contains(leerFicheroDePrueba(t, referencia), "\n"+marcaDeLaCobertura) {
				deLaSkill = append(deLaSkill, skill+": references/"+filepath.Base(referencia)+
					": la región cobertura_por_materia sin data/cobertura.yaml")
			}
		}

		if len(deLaSkill) > 0 {
			defectos[skill] = deLaSkill
		}
	}

	return defectos
}

// marcaDeLaCobertura es el principio de la marca de inicio de la región
// cobertura_por_materia de una referencia de un paquete.
const marcaDeLaCobertura = "<!-- inicio de cobertura_por_materia: "

// conLaPlantillaDeNoCubierto dice si el cuerpo de un SKILL.md tiene la
// plantilla ⚠ NO CUBIERTO: (contracts/datos-y-regiones.md §5 de H33): una sola
// línea que, sin los blancos del principio, empieza por la etiqueta y lleva
// <motivo>.
func conLaPlantillaDeNoCubierto(cuerpo string) bool {
	plantillas := 0

	for linea := range strings.Lines(cuerpo) {
		linea = strings.TrimLeft(linea, " ")
		if strings.HasPrefix(linea, "⚠ NO CUBIERTO: ") && strings.Contains(linea, "<motivo>") {
			plantillas++
		}
	}

	return plantillas == 1
}

// copiaDeLaEntradaYDeLosPaquetes es la copia de plazos-y-territorio y de los
// datos de sus referencias de copiaDePlazosYTerritorio con la skill de entrada,
// consulta-juridica, que no declara applets —la de la región de lo cubierto y la
// de la lista de los paquetes de la base—, y con cada otra skill del repositorio
// que tiene una referencia de un paquete (skillsConUnPaquete), sin sus applets
// declarados, como plazos-y-territorio; y los nombres de sus skills, en orden.
func copiaDeLaEntradaYDeLosPaquetes(t *testing.T) (string, []string) {
	t.Helper()

	raiz := copiaDePlazosYTerritorio(t)
	nombres := []string{"plazos-y-territorio"}

	for _, skill := range slices.Concat([]string{skillDeEntrada}, skillsConUnPaquete(t)) {
		if slices.Contains(nombres, skill) {
			continue
		}

		require.NoError(t, os.CopyFS(rutaDeSkill(raiz, skill), os.DirFS(filepath.Join(skillsDelRepositorio, skill))))
		sinLosApplets(t, rutaDeSkill(raiz, skill, "SKILL.md"))

		nombres = append(nombres, skill)
	}

	slices.Sort(nombres)

	return raiz, nombres
}

// skillsConUnPaquete son las skills del repositorio, en orden de nombre, con
// alguna referencia de un paquete: references/<paquete>.md, con <paquete> el
// vertical de alguna norma de data/normas.yaml (data-model §1 de H45).
func skillsConUnPaquete(t *testing.T) []string {
	t.Helper()

	normas, err := skills.LeerNormas(leerDatosDelRepositorio(t, "normas.yaml"))
	require.NoError(t, err)

	referencias, err := fs.Glob(os.DirFS(skillsDelRepositorio), "*/references/*.md")
	require.NoError(t, err)

	var conPaquete []string

	for _, referencia := range referencias {
		skill, _, _ := strings.Cut(referencia, "/")
		paquete := strings.TrimSuffix(path.Base(referencia), ".md")

		if !slices.Contains(conPaquete, skill) && slices.ContainsFunc(normas, func(norma skills.Norma) bool {
			return norma.Vertical == paquete
		}) {
			conPaquete = append(conPaquete, skill)
		}
	}

	return conPaquete
}

// lineaDeLosApplets es la de los applets que declara el frontmatter de un
// SKILL.md: la tabla de comandos no se regenera sin sus descripciones.
var lineaDeLosApplets = regexp.MustCompile(`(?m)^  kitlegal-applets: [^\n]*\n`)

// sinLosApplets quita del SKILL.md de ruta la línea de los applets que declara,
// si la lleva.
func sinLosApplets(t *testing.T, ruta string) {
	t.Helper()

	escribirFicheroDePrueba(t, ruta, lineaDeLosApplets.ReplaceAllString(leerFicheroDePrueba(t, ruta), ""))
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
