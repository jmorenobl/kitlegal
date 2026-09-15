package skills_test

import (
	"os"
	"path/filepath"
	"syscall"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/jmorenobl/kitlegal/internal/skills"
)

// TestContarLineas fija ContarLineas (data-model §1.2; FR-041): el número de
// saltos de línea, más uno si el contenido no está vacío y no termina en salto.
func TestContarLineas(t *testing.T) {
	t.Parallel()

	casos := []struct {
		nombre    string
		contenido string
		lineas    int
	}{
		{nombre: "vacio", contenido: "", lineas: 0},
		{nombre: "una-linea-sin-salto", contenido: "a", lineas: 1},
		{nombre: "una-linea-con-salto", contenido: "a\n", lineas: 1},
		{nombre: "dos-lineas-sin-salto-final", contenido: "a\nb", lineas: 2},
		{nombre: "dos-lineas-con-salto-final", contenido: "a\nb\n", lineas: 2},
		{nombre: "solo-saltos", contenido: "\n\n", lineas: 2},
	}

	for _, caso := range casos {
		t.Run(caso.nombre, func(t *testing.T) {
			t.Parallel()

			assert.Equal(t, caso.lineas, skills.ContarLineas([]byte(caso.contenido)))
		})
	}
}

// El SKILL.md de la skill alfa del árbol de TestListarYCargar.
const skillMdDeAlfa = "---\nname: alfa\n---\n# Alfa\n"

// TestListarYCargar fija Listar y Cargar (data-model §1; FR-040): todo directorio
// del directorio de skills del repositorio es una skill, y nada más lo es; una
// skill sin SKILL.md, o con un SKILL.md que no es un fichero regular, es un
// defecto que nombra la skill; y un repositorio sin directorio de skills no
// tiene ninguna, sin error.
func TestListarYCargar(t *testing.T) {
	t.Parallel()

	raiz := t.TempDir()
	escribirFicheroDePrueba(t, filepath.Join(raiz, "skills", "beta", "SKILL.md"), "# Beta\n")
	escribirFicheroDePrueba(t, filepath.Join(raiz, "skills", "alfa", "SKILL.md"), skillMdDeAlfa)
	escribirFicheroDePrueba(t, filepath.Join(raiz, "skills", "sin-skill-md", "references", "normas.md"), "# Normas\n")
	require.NoError(t, os.MkdirAll(filepath.Join(raiz, "skills", "skill-md-es-un-directorio", "SKILL.md"), 0o750))
	escribirFicheroDePrueba(t, filepath.Join(raiz, "skills", "notas.md"), "# Notas\n")

	t.Run("listar", func(t *testing.T) {
		t.Parallel()

		nombres, err := skills.Listar(raiz)
		require.NoError(t, err)
		assert.Equal(t, []string{"alfa", "beta", "sin-skill-md", "skill-md-es-un-directorio"}, nombres,
			"cada directorio es una skill, en orden de nombre; el fichero notas.md no lo es")
	})

	t.Run("cargar", func(t *testing.T) {
		t.Parallel()

		skill, err := skills.Cargar(raiz, "alfa")
		require.NoError(t, err)
		assert.Equal(t, skills.Skill{
			Nombre:     "alfa",
			Directorio: filepath.Join(raiz, "skills", "alfa"),
			Contenido:  []byte(skillMdDeAlfa),
		}, skill)
	})

	defectos := []struct {
		skill   string
		defecto string
	}{
		{skill: "sin-skill-md", defecto: "falta SKILL.md"},
		{skill: "skill-md-es-un-directorio", defecto: "SKILL.md no es un fichero regular"},
	}

	for _, caso := range defectos {
		t.Run(caso.skill, func(t *testing.T) {
			t.Parallel()

			skill, err := skills.Cargar(raiz, caso.skill)

			var defecto *skills.DefectoDeSkill
			require.ErrorAs(t, err, &defecto)
			assert.Equal(t, &skills.DefectoDeSkill{Skill: caso.skill, Defecto: caso.defecto}, defecto)
			require.EqualError(t, err, caso.skill+": "+caso.defecto)
			assert.Zero(t, skill, "una skill con defectos no se entrega a medias")
		})
	}

	t.Run("nombre-de-un-fichero", func(t *testing.T) {
		t.Parallel()

		// notas.md no es una skill: la ruta de su SKILL.md pasa por un fichero, y
		// eso no es que falte, sino que no se puede consultar.
		skill, err := skills.Cargar(raiz, "notas.md")
		require.ErrorIs(t, err, syscall.ENOTDIR)
		require.ErrorContains(t, err, "notas.md: SKILL.md no se puede consultar: ")

		var defecto *skills.DefectoDeSkill
		assert.NotErrorAs(t, err, &defecto)
		assert.Zero(t, skill)
	})

	t.Run("sin-directorio-de-skills", func(t *testing.T) {
		t.Parallel()

		nombres, err := skills.Listar(t.TempDir())
		require.NoError(t, err)
		assert.Empty(t, nombres)
	})

	t.Run("directorio-de-skills-que-no-es-un-directorio", func(t *testing.T) {
		t.Parallel()

		otraRaiz := t.TempDir()
		escribirFicheroDePrueba(t, filepath.Join(otraRaiz, "skills"), "no es un directorio\n")

		nombres, err := skills.Listar(otraRaiz)
		require.ErrorContains(t, err, "el directorio de skills "+filepath.Join(otraRaiz, "skills")+" no se puede listar")
		assert.Nil(t, nombres)
	})
}

// escribirFicheroDePrueba escribe el contenido en la ruta, dentro de un árbol
// temporal del test, creando antes lo que falte de la ruta.
func escribirFicheroDePrueba(t *testing.T, ruta, contenido string) {
	t.Helper()

	require.NoError(t, os.MkdirAll(filepath.Dir(ruta), 0o750))
	require.NoError(t, os.WriteFile(ruta, []byte(contenido), 0o600))
}
