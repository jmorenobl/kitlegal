package workflow

import (
	"encoding/json"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestLaSesionDeUnPasoRetiraLasMarcasDeEscritura: las marcas de escritura
// vacías que deja Claude Code (`<directorio>/.claude/.cc-writes/`) salen del
// árbol al abrir la sesión de un paso, con el `.claude/` que las contiene si se
// queda vacío, y los ajustes de la sesión llevan el gancho PostToolUse que retira
// las que deje ella. Lo que tiene contenido se queda. En H9 una marca en un
// directorio de grabaciones puso `make ci` en rojo y mandó tres tareas a
// cuarentena (scripts/workflow/marcas.sh).
func TestLaSesionDeUnPasoRetiraLasMarcasDeEscritura(t *testing.T) {
	t.Parallel()

	const (
		grabaciones = "testdata/grabaciones/.claude"
		delHito     = directorioDelHito + "/.claude"
		aMedias     = "con-contenido/.claude/.cc-writes/a-medias.tmp"
		ajustes     = "con-ajustes/.claude"
		guion       = "scripts/workflow/marcas.sh"
	)

	a := nuevoArbol(t)
	ponerMarcas := func() {
		for _, dir := range []string{grabaciones, delHito, ajustes} {
			require.NoError(t, a.raiz.MkdirAll(dir+"/.cc-writes", 0o750))
		}
	}

	ponerMarcas()
	a.escribir(aMedias, "lo que escribe una sesión que no ha terminado")
	a.escribir(ajustes+"/settings.json", "{}")

	salida, argumentos := a.abrirSesion(sesionDelTest{modelo: modeloDelJuezA + "#juez_spec:modelo_juez"})
	require.Zero(t, salida.codigo, "la sesión se abre: %s", salida.errores)

	comprobar := func(cuando string) {
		assert.True(t, a.noExiste(grabaciones), "%s: la marca y su .claude/ vacío salen de las grabaciones", cuando)
		assert.True(t, a.noExiste(delHito), "%s: y del directorio del hito", cuando)
		assert.True(t, a.noExiste(ajustes+"/.cc-writes"), "%s: la marca sale de un .claude/ con ajustes", cuando)
		assert.True(t, a.existe(ajustes+"/settings.json"), "%s: y los ajustes se quedan", cuando)
		assert.True(t, a.existe(aMedias), "%s: una marca con contenido se queda", cuando)
	}
	comprobar("al abrir la sesión")

	var deLaSesion struct {
		Hooks struct {
			PostToolUse []struct {
				Matcher string `json:"matcher"`
				Hooks   []struct {
					Type    string `json:"type"`
					Command string `json:"command"`
				} `json:"hooks"`
			} `json:"PostToolUse"`
		} `json:"hooks"`
	}

	require.NoError(t, json.Unmarshal([]byte(valorDe(argumentos, "--settings")), &deLaSesion))
	require.Len(t, deLaSesion.Hooks.PostToolUse, 1, "un gancho PostToolUse")

	gancho := deLaSesion.Hooks.PostToolUse[0]
	assert.Equal(t, "*", gancho.Matcher, "tras cualquier herramienta")
	require.Len(t, gancho.Hooks, 1)

	raiz, err := filepath.EvalSymlinks(a.ruta("."))
	require.NoError(t, err)

	assert.Equal(t, "command", gancho.Hooks[0].Type)
	assert.Equal(t, "'"+filepath.Join(raiz, guion)+"'", gancho.Hooks[0].Command, "el guion del árbol, por su ruta absoluta")

	// Lo que deje la sesión lo retira el gancho, que es el mismo guion.
	ponerMarcas()
	a.debeSalirBien(nil, guion)
	comprobar("tras el gancho")
}
