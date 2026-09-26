package main

import (
	"io/fs"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/jmorenobl/kitlegal/internal/app"
)

// versionDePrueba es la versión con la que se componen las dependencias en estos
// tests: la del primer binario de e2e con versión (contracts/arnes-e2e.md §2).
const versionDePrueba = "v0.1.0"

// TestDependenciasDeSkills fija cómo elige este binario el creador de enlaces
// de skills: sin elección, las dependencias del sistema tal cual; con la del
// enlazador que falla, las mismas con solo el Enlazador sustituido; y cualquier
// otro valor, un defecto de la construcción que se nombra, nunca el creador del
// sistema en silencio (FR-024; research.md D24).
func TestDependenciasDeSkills(t *testing.T) {
	t.Parallel()

	sistema := app.DependenciasDeSkillsDelSistema(versionDePrueba)

	t.Run("sin elección son las del sistema tal cual", func(t *testing.T) {
		t.Parallel()

		dependencias, err := dependenciasDeSkills(versionDePrueba, "")
		require.NoError(t, err)
		assert.Equal(t, sistema, dependencias)
	})

	t.Run("el enlazador que falla sustituye solo el Enlazador", func(t *testing.T) {
		t.Parallel()

		dependencias, err := dependenciasDeSkills(versionDePrueba, enlazadorQueFalla)
		require.NoError(t, err)
		assert.Equal(t, sistema.Version, dependencias.Version)
		assert.Equal(t, sistema.Skills, dependencias.Skills)
		assert.Equal(t, enlazadorFallido{}, dependencias.Enlazador)
	})

	t.Run("una elección desconocida es un defecto de la construcción", func(t *testing.T) {
		t.Parallel()

		_, err := dependenciasDeSkills(versionDePrueba, "sistema")
		require.ErrorIs(t, err, errEnlazadorDesconocido)
		assert.ErrorContains(t, err, `"sistema"`)
	})
}

// TestEnlazadorFallido fija el creador de enlaces que siempre falla de
// contracts/arnes-e2e.md §2: ningún directorio admite enlaces, sin error, y
// crear uno es siempre un error que no deja nada en el disco.
func TestEnlazadorFallido(t *testing.T) {
	t.Parallel()

	directorio := t.TempDir()

	disponible, err := enlazadorFallido{}.Disponible(directorio)
	require.NoError(t, err)
	assert.False(t, disponible)

	ruta := filepath.Join(directorio, "boe-legislacion")

	err = enlazadorFallido{}.Enlazar("../../.agents/skills/boe-legislacion", ruta)
	require.ErrorIs(t, err, errSinEnlaces)
	require.ErrorContains(t, err, ruta)

	_, err = os.Lstat(ruta)
	require.ErrorIs(t, err, fs.ErrNotExist)

	entradas, err := os.ReadDir(directorio)
	require.NoError(t, err)
	assert.Empty(t, entradas, "ni Disponible ni Enlazar tocan el disco")
}

// TestRegistroDeE2E comprueba que el registro de este binario, construido sin
// ninguna elección de enlazador —como en los tests, sin -ldflags—, lleva los
// applets de ejemplo, boe, skills y territorio.
func TestRegistroDeE2E(t *testing.T) {
	t.Parallel()

	registro, err := registroDeE2E(versionDePrueba)
	require.NoError(t, err)
	assert.Equal(t, []string{"boe", "contar", "echo", "skills", "territorio"}, registro.Nombres())
}
