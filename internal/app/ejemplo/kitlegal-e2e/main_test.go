package main

import (
	"io/fs"
	"os"
	"path/filepath"
	"strconv"
	"testing"
	"time"

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

// TestDependenciasDelReloj fija cómo elige este binario su reloj: sin elección,
// el del sistema en el grafo, como el binario distribuido; con un instante RFC
// 3339, ese instante en cada lectura del reloj del grafo, con su desplazamiento;
// y cualquier otro valor, un defecto de la construcción que se nombra, nunca el
// reloj del sistema en silencio (contracts/arnes-e2e.md §2 de H7; research.md
// D23 de H7). Que la reproducción de boe fecha con el mismo instante lo
// comprueba TestBinariosDelArnes sobre los binarios construidos: la
// reproducción sirve de una carpeta relativa al directorio desde el que se
// invoca el binario, y el de estos tests no la tiene.
func TestDependenciasDelReloj(t *testing.T) {
	t.Parallel()

	t.Run("sin elección, el reloj del sistema", func(t *testing.T) {
		t.Parallel()

		deBoe, delGrafo, err := dependenciasDelReloj("")
		require.NoError(t, err)
		assert.NotNil(t, deBoe.Cliente)
		assert.Empty(t, delGrafo.Almacen)
		require.NotNil(t, delGrafo.Reloj)

		antes := time.Now()
		leido := delGrafo.Reloj()
		despues := time.Now()

		assert.Truef(t, !leido.Before(antes) && !leido.After(despues), "%s no está entre %s y %s", leido, antes, despues)
	})

	t.Run("un instante RFC 3339 fija el reloj del grafo", func(t *testing.T) {
		t.Parallel()

		instantes := map[string]time.Time{
			"2026-09-28T12:00:00Z":      time.Date(2026, time.September, 28, 12, 0, 0, 0, time.UTC),
			"2026-10-06T14:00:00+02:00": time.Date(2026, time.October, 6, 12, 0, 0, 0, time.UTC),
		}

		for eleccion, instante := range instantes {
			deBoe, delGrafo, err := dependenciasDelReloj(eleccion)
			require.NoError(t, err, eleccion)
			assert.NotNil(t, deBoe.Cliente)
			assert.Empty(t, delGrafo.Almacen)
			require.NotNil(t, delGrafo.Reloj)

			for range 2 {
				leido := delGrafo.Reloj()
				assert.Truef(t, instante.Equal(leido), "%s es %s", eleccion, leido)
				assert.Equal(t, eleccion, leido.Format(time.RFC3339), "el reloj conserva el desplazamiento de la construcción")
			}
		}
	})

	t.Run("un valor que no es RFC 3339 es un defecto de la construcción", func(t *testing.T) {
		t.Parallel()

		invalidos := []string{
			"2026-09-28",
			"2026-09-28T12:00:00",
			"2026-09-28 12:00:00Z",
			" 2026-09-28T12:00:00Z",
			"2026-09-28T12:00:00Z ",
			"2026-09-31T12:00:00Z",
			"ma\xc3\xb1ana",
		}

		for _, eleccion := range invalidos {
			_, _, err := dependenciasDelReloj(eleccion)
			require.ErrorIs(t, err, errRelojInvalido, "%q", eleccion)
			assert.ErrorContains(t, err, strconv.Quote(eleccion))
		}
	})
}

// TestRegistroDeE2E comprueba que el registro de este binario, construido sin
// ninguna elección de enlazador ni de reloj —como en los tests, sin -ldflags—,
// lleva los applets de ejemplo, boe, cita, graph, mcp, skills y territorio. Que
// además entrega al grafo del mundo lo ejercen los guiones del e2e, que ven
// world.db en el directorio de la caché de cada guion; y que da a las órdenes la
// entrada estándar del proceso, los de cita cotejar, que leen de ella el
// documento.
func TestRegistroDeE2E(t *testing.T) {
	t.Parallel()

	registro, err := registroDeE2E(versionDePrueba)
	require.NoError(t, err)
	assert.Equal(t, []string{"boe", "cita", "contar", "echo", "graph", "mcp", "skills", "territorio"}, registro.Nombres())
}
