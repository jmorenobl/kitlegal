package kitlegal_test

import (
	"fmt"
	"io/fs"
	"maps"
	"os"
	"path"
	"slices"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const (
	// carpetaDeGitHub es la de los flujos de GitHub Actions y la configuración
	// de GitHub del repositorio.
	carpetaDeGitHub = ".github"

	// ficheroDelMakefile es la única superficie de invocación del repositorio.
	ficheroDelMakefile = "Makefile"

	// flujoDeEvals es el del job de evals, el último que nombró bin/instalado
	// (contracts/skills-e-invocacion.md §6).
	flujoDeEvals = ".github/workflows/evals.yml"

	// entradaDeScripts es la que una skill ya no lleva (FR-082, ADR 0019).
	entradaDeScripts = "scripts"
)

// restosDeLaInstalacionPorEnlaces son las cadenas que nombran la instalación por
// enlaces que ADR 0019 retira: los enlaces de scripts/ de cada skill y el
// directorio del binario al que apuntaban (FR-083, SC-014).
var restosDeLaInstalacionPorEnlaces = []string{"scripts/boe", "scripts/territorio", "bin/instalado"}

// retiradosConLosEnlaces son los ficheros de la instalación por enlaces que ya
// no existen: el guion de enlaces y lo que los regeneraba en skills-sync, con su
// test (FR-083; contracts/skills-e-invocacion.md §4).
var retiradosConLosEnlaces = []string{
	"scripts/instalar-skills.sh",
	"internal/skills/enlaces.go",
	"internal/skills/enlaces_test.go",
}

// TestSinInstalacionPorEnlaces fija que de la instalación por enlaces no queda
// nada (FR-083, SC-014; contracts/skills-e-invocacion.md §4): ningún fichero de
// skills/, ni el Makefile, ni ninguno de .github/ nombra scripts/boe,
// scripts/territorio ni bin/instalado, que es la búsqueda de quickstart.md §6b;
// no existen el guion de enlaces ni internal/skills/enlaces.go con su test; y
// ninguna skill tiene una entrada scripts, vista sin seguir enlaces.
func TestSinInstalacionPorEnlaces(t *testing.T) {
	t.Parallel()

	raiz := os.DirFS(".")

	casos := []struct {
		nombre string
		probar func(*testing.T, fs.FS)
	}{
		{"nadie-nombra-la-instalacion-por-enlaces", probarQueNadieNombraLaInstalacionPorEnlaces},
		{"retirados", probarRetiradosConLosEnlaces},
		{"ninguna-skill-lleva-scripts", probarNingunaSkillLlevaScripts},
	}

	for _, caso := range casos {
		t.Run(caso.nombre, func(t *testing.T) {
			t.Parallel()

			caso.probar(t, raiz)
		})
	}
}

// probarQueNadieNombraLaInstalacionPorEnlaces busca, línea a línea, cada resto
// de la instalación por enlaces en los ficheros de quickstart.md §6b, y los
// nombra todos como ruta:línea: texto.
func probarQueNadieNombraLaInstalacionPorEnlaces(t *testing.T, raiz fs.FS) {
	t.Helper()

	ficheros := ficherosDeLaBusqueda(t, raiz)

	vigilados := []string{ficheroDelMakefile, flujoDeEvals}
	for _, skill := range skillsDeHoy {
		vigilados = append(vigilados, path.Join(carpetaDeSkills, skill, ficheroDeLaSkill))
	}

	require.Subset(t, slices.Collect(maps.Keys(ficheros)), vigilados,
		"la búsqueda no lee lo que vigila: pasaría en vacío")

	var restos []string

	for _, ruta := range slices.Sorted(maps.Keys(ficheros)) {
		for numero, linea := range strings.Split(string(ficheros[ruta]), "\n") {
			nombraUnResto := slices.ContainsFunc(restosDeLaInstalacionPorEnlaces, func(resto string) bool {
				return strings.Contains(linea, resto)
			})
			if nombraUnResto {
				restos = append(restos, fmt.Sprintf("%s:%d: %s", ruta, numero+1, strings.TrimSpace(linea)))
			}
		}
	}

	assert.Empty(t, restos, "%s nombran la instalación por enlaces (FR-083, SC-014)",
		strings.Join(restosDeLaInstalacionPorEnlaces, ", "))
}

// ficherosDeLaBusqueda son, por su ruta, los ficheros regulares de skills/ y de
// .github/, a cualquier profundidad, y el Makefile, con sus bytes. No sigue
// ningún enlace, como grep -r.
func ficherosDeLaBusqueda(t *testing.T, raiz fs.FS) map[string][]byte {
	t.Helper()

	ficheros := map[string][]byte{}

	for _, carpeta := range []string{carpetaDeSkills, carpetaDeGitHub} {
		err := fs.WalkDir(raiz, carpeta, func(ruta string, entrada fs.DirEntry, err error) error {
			if err != nil || !entrada.Type().IsRegular() {
				return err
			}

			ficheros[ruta], err = fs.ReadFile(raiz, ruta)

			return err
		})
		require.NoError(t, err)
	}

	makefile, err := fs.ReadFile(raiz, ficheroDelMakefile)
	require.NoError(t, err)

	ficheros[ficheroDelMakefile] = makefile

	return ficheros
}

// probarRetiradosConLosEnlaces exige que cada fichero retirado con la
// instalación por enlaces no exista, ni siquiera como enlace colgando.
func probarRetiradosConLosEnlaces(t *testing.T, raiz fs.FS) {
	t.Helper()

	for _, ruta := range retiradosConLosEnlaces {
		_, err := fs.Lstat(raiz, ruta)
		assert.ErrorIsf(t, err, fs.ErrNotExist, "%s sigue existiendo (FR-083)", ruta)
	}
}

// probarNingunaSkillLlevaScripts exige que ningún directorio de skills/ tenga
// una entrada scripts de ningún tipo, vista con Lstat: un enlace colgando
// también cuenta (FR-082, FR-083; ADR 0019).
func probarNingunaSkillLlevaScripts(t *testing.T, raiz fs.FS) {
	t.Helper()

	entradas, err := fs.ReadDir(raiz, carpetaDeSkills)
	require.NoError(t, err)

	var skills []string

	for _, entrada := range entradas {
		if entrada.IsDir() {
			skills = append(skills, entrada.Name())
		}
	}

	require.Subset(t, skills, skillsDeHoy, "skills/ no tiene las skills de hoy: la comprobación pasaría en vacío")

	for _, skill := range skills {
		_, err := fs.Lstat(raiz, path.Join(carpetaDeSkills, skill, entradaDeScripts))
		assert.ErrorIsf(t, err, fs.ErrNotExist, "la skill %s tiene %s: una skill no lleva scripts/ (FR-082, FR-083)",
			skill, entradaDeScripts)
	}
}
