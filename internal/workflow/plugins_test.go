package workflow

import (
	"encoding/json"
	"path/filepath"
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

			assert.Contains(t, argumentos, "--strict-mcp-config", "solo los servidores MCP que se le den")
			assert.NotContains(t, argumentos, "--mcp-config", "sin índice de CodeGraph no se le da ninguno")

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

// TestLaSesionDeUnPasoCargaCodeGraph: con el índice de CodeGraph en el
// repositorio y su programa en el equipo, la sesión de un paso carga ese
// servidor MCP y ninguno más, atado a la raíz del repositorio y sin red
// —sin telemetría, sin comprobar versiones y sin descargas—, y sus herramientas
// se admiten sin preguntar. Sin el índice o sin el programa, la sesión se abre
// como antes, sin ningún servidor MCP.
func TestLaSesionDeUnPasoCargaCodeGraph(t *testing.T) {
	t.Parallel()

	const (
		indice    = ".codegraph/codegraph.db"
		sustituto = "sustitutos/codegraph"
		regla     = "mcp__codegraph__*"
	)

	casos := []struct {
		nombre    string
		conIndice bool
		programa  string
		carga     bool
	}{
		{nombre: "con el indice y el programa", conIndice: true, programa: sustituto, carga: true},
		{nombre: "sin el indice", programa: sustituto},
		{nombre: "sin el programa", conIndice: true, programa: "codegraph-que-no-existe"},
	}

	for _, caso := range casos {
		t.Run(caso.nombre, func(t *testing.T) {
			t.Parallel()

			a := nuevoArbol(t)
			a.escribirEjecutable(sustituto, "#!/bin/sh\nexit 0\n")

			if caso.conIndice {
				a.escribir(indice, "")
			}

			programa := caso.programa
			if programa == sustituto {
				programa = a.ruta(sustituto)
			}

			salida, argumentos := a.abrirSesion(sesionDelTest{
				modelo:    modeloDelJuezA + "#juez_spec:modelo_juez",
				variables: []string{"KITLEGAL_CODEGRAPH_BIN=" + programa},
			})
			require.Zero(t, salida.codigo, "la sesión se abre: %s", salida.errores)

			assert.Contains(t, argumentos, "--strict-mcp-config", "solo los servidores MCP que se le den")

			var ajustes struct {
				Permissions struct {
					Allow []string `json:"allow"`
				} `json:"permissions"`
			}

			require.NoError(t, json.Unmarshal([]byte(valorDe(argumentos, "--settings")), &ajustes))

			if !caso.carga {
				assert.NotContains(t, argumentos, "--mcp-config", "sin índice o sin programa no se le da ningún servidor")
				assert.NotContains(t, ajustes.Permissions.Allow, regla)

				return
			}

			var servidores struct {
				MCPServers map[string]struct {
					Command string            `json:"command"`
					Args    []string          `json:"args"`
					Env     map[string]string `json:"env"`
				} `json:"mcpServers"`
			}

			require.NoError(t, json.Unmarshal([]byte(valorDe(argumentos, "--mcp-config")), &servidores))
			require.Len(t, servidores.MCPServers, 1, "CodeGraph es el único servidor")

			servidor, ok := servidores.MCPServers["codegraph"]
			require.True(t, ok, "el servidor se llama codegraph, el prefijo de la regla de sus herramientas")

			raiz, err := filepath.EvalSymlinks(a.ruta("."))
			require.NoError(t, err)

			programaReal, err := filepath.EvalSymlinks(programa)
			require.NoError(t, err)

			ordenReal, err := filepath.EvalSymlinks(servidor.Command)
			require.NoError(t, err)

			assert.Equal(t, programaReal, ordenReal, "la ruta absoluta del programa")
			assert.Equal(t, []string{"serve", "--mcp", "--path", raiz}, servidor.Args, "atado a la raíz del repositorio")
			assert.Equal(t, map[string]string{
				"DO_NOT_TRACK":              "1",
				"CODEGRAPH_TELEMETRY":       "0",
				"CODEGRAPH_NO_UPDATE_CHECK": "1",
				"CODEGRAPH_NO_DOWNLOAD":     "1",
			}, servidor.Env, "sin telemetría, sin comprobar versiones y sin descargas")
			assert.Contains(t, ajustes.Permissions.Allow, regla, "sus herramientas se admiten sin preguntar")
		})
	}
}
