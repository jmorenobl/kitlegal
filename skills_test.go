package kitlegal_test

import (
	"io/fs"
	"maps"
	"os"
	"path"
	"slices"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/jmorenobl/kitlegal"
)

const (
	// carpetaDeSkills es la de las skills del repositorio, relativa a la raíz
	// del módulo, que es el directorio de este paquete y donde go test ejecuta
	// sus tests; y lo empotrado lleva las mismas rutas.
	carpetaDeSkills = "skills"

	// ficheroDeLaSkill es el que hace de un directorio de skills/ una skill
	// (FR-004).
	ficheroDeLaSkill = "SKILL.md"

	// prefijoDeReferencias es el de todo fichero de references/, relativo al
	// directorio de su skill.
	prefijoDeReferencias = "references/"
)

// skillsDeHoy son las skills que el repositorio tiene hoy (FR-004, SC-021): sin
// ellas, la comparación entre lo empotrado y el árbol pasaría en vacío. Una
// skill nueva no rompe nada: basta con que estén estas.
var skillsDeHoy = []string{"boe-legislacion", "legal-core"}

// TestSkillsEmpotradas fija que lo empotrado es exactamente lo que dice FR-001
// del árbol, byte a byte (FR-001, FR-003, FR-004; contracts/skills-e-invocacion.md
// §1): por cada directorio skills/<skill>/ con SKILL.md, ese SKILL.md y todo
// fichero bajo su references/, a cualquier profundidad y aunque su nombre
// empiece por punto, y nada más, ni de ese directorio ni de ningún otro.
//
// Una skill sin SKILL.md no entra. El patrón de //go:embed no puede exigirlo: si
// el árbol tuviera un directorio con references/ y sin SKILL.md, lo empotrado
// llevaría esas referencias y este test fallaría nombrándolas, porque no son de
// ninguna skill; quien lee lo empotrado, internal/app, lo ignora igualmente.
// También falla si el patrón deja fuera algo que FR-001 pide, como un fichero
// con punto en un subdirectorio de references/.
func TestSkillsEmpotradas(t *testing.T) {
	t.Parallel()

	delArbol := ficherosDeLasSkills(t, os.DirFS("."))
	empotrados := ficherosDe(t, kitlegal.Skills())

	require.Subset(t, skillsDe(delArbol), skillsDeHoy,
		"el árbol no tiene las skills de hoy: la comparación pasaría en vacío")

	require.Equal(t, slices.Sorted(maps.Keys(delArbol)), slices.Sorted(maps.Keys(empotrados)),
		"lo empotrado no es exactamente el SKILL.md y references/ de cada skill del árbol (FR-001, FR-004)")

	for ruta, contenido := range delArbol {
		assert.Equalf(t, contenido, empotrados[ruta], "%s empotrado no es byte a byte el del árbol (FR-003)", ruta)
	}
}

// ficherosDeLasSkills son, por su ruta skills/<skill>/…, los ficheros del árbol
// que FR-001 manda empotrar: el SKILL.md de cada directorio de skills/ que lo
// tiene y todo fichero bajo su references/. No sigue ningún enlace, que
// //go:embed tampoco empotra.
func ficherosDeLasSkills(t *testing.T, arbol fs.FS) map[string][]byte {
	t.Helper()

	ficheros := map[string][]byte{}

	err := fs.WalkDir(arbol, carpetaDeSkills, func(ruta string, entrada fs.DirEntry, err error) error {
		if err != nil || !entrada.Type().IsRegular() {
			return err
		}

		partes := strings.SplitN(ruta, "/", 3)
		if len(partes) < 3 || (partes[2] != ficheroDeLaSkill && !strings.HasPrefix(partes[2], prefijoDeReferencias)) {
			return nil
		}

		ficheros[ruta], err = fs.ReadFile(arbol, ruta)

		return err
	})
	require.NoError(t, err)

	maps.DeleteFunc(ficheros, func(ruta string, _ []byte) bool {
		partes := strings.SplitN(ruta, "/", 3)
		_, conSkillMd := ficheros[path.Join(partes[0], partes[1], ficheroDeLaSkill)]

		return !conSkillMd
	})

	return ficheros
}

// ficherosDe son todos los ficheros de un sistema de ficheros por su ruta, con
// sus bytes.
func ficherosDe(t *testing.T, arbol fs.FS) map[string][]byte {
	t.Helper()

	ficheros := map[string][]byte{}

	err := fs.WalkDir(arbol, ".", func(ruta string, entrada fs.DirEntry, err error) error {
		if err != nil || entrada.IsDir() {
			return err
		}

		ficheros[ruta], err = fs.ReadFile(arbol, ruta)

		return err
	})
	require.NoError(t, err)

	return ficheros
}

// skillsDe son los nombres de las skills cuyo SKILL.md, skills/<skill>/SKILL.md,
// está entre los ficheros.
func skillsDe(ficheros map[string][]byte) []string {
	var skills []string

	for ruta := range ficheros {
		if partes := strings.SplitN(ruta, "/", 3); len(partes) == 3 && partes[2] == ficheroDeLaSkill {
			skills = append(skills, partes[1])
		}
	}

	return skills
}
