package workflow

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestLaSesionDeUnPasoArrancaSinPluginsNiMCP: cada plugin que los ajustes de
// usuario o los del proyecto tienen activado llega apagado a la sesión por los
// ajustes que le pone scripts/claude-modelo.sh, y la sesión no carga ningún
// servidor MCP de fuera de ellos. Lo que ya estaba apagado no se menciona, y
// sin ajustes de usuario ni de proyecto la sesión se abre igual.
func TestLaSesionDeUnPasoArrancaSinPluginsNiMCP(t *testing.T) {
	t.Parallel()

	casos := []struct {
		nombre   string
		proyecto string
		usuario  string
		apagados map[string]bool
	}{
		{
			nombre:   "activados en el proyecto y en el usuario",
			proyecto: `{"enabledPlugins": {"proyecto@mercado": true, "apagado@mercado": false, "compartido@mercado": true}}`,
			usuario:  `{"enabledPlugins": {"usuario@mercado": true, "compartido@mercado": true}}`,
			apagados: map[string]bool{"proyecto@mercado": false, "usuario@mercado": false, "compartido@mercado": false},
		},
		{
			nombre:   "solo en el usuario",
			usuario:  `{"enabledPlugins": {"usuario@mercado": true}, "model": "opus"}`,
			apagados: map[string]bool{"usuario@mercado": false},
		},
		{
			nombre:   "todos apagados ya",
			proyecto: `{"enabledPlugins": {"apagado@mercado": false}}`,
			usuario:  `{}`,
			apagados: map[string]bool{},
		},
		{nombre: "sin ajustes", apagados: map[string]bool{}},
	}

	for _, caso := range casos {
		t.Run(caso.nombre, func(t *testing.T) {
			t.Parallel()

			a := nuevoArbol(t)

			if caso.proyecto != "" {
				a.escribir(".claude/settings.json", caso.proyecto)
			}

			if caso.usuario != "" {
				a.escribir(".config-claude/settings.json", caso.usuario)
			}

			salida, argumentos := a.abrirSesion(sesionDelTest{modelo: modeloDelJuezA + "#juez_spec:modelo_juez"})
			require.Zero(t, salida.codigo, "la sesión se abre: %s", salida.errores)

			assert.Contains(t, argumentos, "--strict-mcp-config", "solo los servidores MCP que se le den, y no se le da ninguno")

			var ajustes struct {
				EnabledPlugins     map[string]bool `json:"enabledPlugins"`
				SyncClaudeAiSkills *bool           `json:"syncClaudeAiSkills"`
			}

			require.NoError(t, json.Unmarshal([]byte(valorDe(argumentos, "--settings")), &ajustes))
			assert.Equal(t, caso.apagados, ajustes.EnabledPlugins)

			// Las skills de la cuenta de claude.ai (anthropic-skills:*) no las
			// apaga enabledPlugins; con false en --settings, Claude Code no las
			// carga en esa invocación y no mueve las sincronizadas.
			require.NotNil(t, ajustes.SyncClaudeAiSkills, "los ajustes de la sesión dicen si sincroniza las skills de la cuenta")
			assert.False(t, *ajustes.SyncClaudeAiSkills, "la sesión arranca sin las skills de la cuenta")
		})
	}
}
