package workflow

import (
	"encoding/json"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"slices"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestLaSesionDeUnPasoSoloEscribeDondeDebe: Write, Edit y NotebookEdit no
// pasan por el sandbox, y en el run de H39 (50b213d0) la sesión `reparar` de
// T010 editó con Edit un fichero de la memoria automática del proyecto
// (~/.claude/projects/<proyecto>/memory), que la sesión tenía encendida y el
// gancho admitía. scripts/claude-modelo.sh abre la sesión con la memoria
// automática apagada, aunque quien la lance la pida encendida; con las
// escrituras de esas herramientas cerradas por Claude Code a los directorios de
// trabajo, que son el repositorio y los temporales que el sandbox ya deja
// escribir, y ninguno más; y con el directorio de Claude Code vedado a esas
// herramientas y el de sus proyectos a toda orden de Bash.
func TestLaSesionDeUnPasoSoloEscribeDondeDebe(t *testing.T) {
	t.Parallel()

	a := nuevoArbol(t)
	entorno := filepath.Join(a.fuera, "entorno")
	delSandbox := filepath.Join(a.fuera, temporalDelSandbox)

	salida, argumentos := a.abrirSesion(sesionDelTest{
		modelo: modeloDelJuezA + "#juez_spec:modelo_juez",
		variables: []string{
			"KITLEGAL_TEST_ENTORNO=" + entorno,
			"CLAUDE_CODE_DISABLE_AUTO_MEMORY=0",
		},
	})
	require.Zero(t, salida.codigo, "la sesión se abre: %s", salida.errores)

	fuera, err := os.OpenRoot(a.fuera)
	require.NoError(t, err)

	defer func() { require.NoError(t, fuera.Close()) }()

	recibido, err := fuera.ReadFile("entorno")
	require.NoError(t, err)

	var memoria []string

	for variable := range strings.Lines(string(recibido)) {
		if valor, ok := strings.CutPrefix(strings.TrimSuffix(variable, "\n"), "CLAUDE_CODE_DISABLE_AUTO_MEMORY="); ok {
			memoria = append(memoria, valor)
		}
	}

	assert.Equal(t, []string{"1"}, memoria, "la sesión arranca sin la memoria automática, aunque quien la lanza la pida encendida")

	var ajustes struct {
		Permissions struct {
			BlockWritesOutsideWorkingDirectories *bool    `json:"blockWritesOutsideWorkingDirectories"`
			Deny                                 []string `json:"deny"`
		} `json:"permissions"`
		Sandbox struct {
			Filesystem struct {
				DenyWrite []string `json:"denyWrite"`
			} `json:"filesystem"`
		} `json:"sandbox"`
	}

	require.NoError(t, json.Unmarshal([]byte(valorDe(argumentos, "--settings")), &ajustes))

	require.NotNil(t, ajustes.Permissions.BlockWritesOutsideWorkingDirectories, "los ajustes dicen si Write y Edit salen de los directorios de trabajo")
	assert.True(t, *ajustes.Permissions.BlockWritesOutsideWorkingDirectories, "Write, Edit y NotebookEdit no salen de los directorios de trabajo")
	assert.Contains(t, ajustes.Permissions.Deny, "Edit(/"+a.config+"/**)", "el directorio de Claude Code, vedado a Write y Edit")
	assert.Contains(t, ajustes.Sandbox.Filesystem.DenyWrite, a.config+"/projects", "el de sus proyectos, vedado a toda orden de Bash")

	delSandboxReal, err := filepath.EvalSymlinks(delSandbox)
	require.NoError(t, err, "el wrapper crea el temporal del sandbox si no existe")

	esperados := []string{delSandboxReal}

	if runtime.GOOS == "darwin" {
		temporal, err := exec.CommandContext(t.Context(), "getconf", "DARWIN_USER_TEMP_DIR").Output()
		require.NoError(t, err)

		temporalReal, err := filepath.EvalSymlinks(strings.TrimSpace(string(temporal)))
		require.NoError(t, err)

		esperados = append(esperados, temporalReal)
	}

	assert.ElementsMatch(t, esperados, valoresDe(argumentos, "--add-dir"),
		"los únicos directorios de trabajo además del repositorio son los temporales que el sandbox ya deja escribir")
}

// TestLaSesionDeUnPasoPideUnClaudeCodeQueCierreLasEscrituras: antes de la
// 2.1.314, Claude Code no conoce blockWritesOutsideWorkingDirectories y no la
// aplica: con esa versión, o con una que no se puede leer, la sesión no se abre
// y el wrapper dice por qué.
func TestLaSesionDeUnPasoPideUnClaudeCodeQueCierreLasEscrituras(t *testing.T) {
	t.Parallel()

	casos := []struct {
		version string
		abre    bool
	}{
		{version: "2.1.313 (Claude Code)"},
		{version: "1.9.999 (Claude Code)"},
		{version: "no es una versión"},
		{version: "2.1.314 (Claude Code)", abre: true},
		{version: "2.2.0 (Claude Code)", abre: true},
		{version: "3.0.0", abre: true},
	}

	for _, caso := range casos {
		t.Run(caso.version, func(t *testing.T) {
			t.Parallel()

			a := nuevoArbol(t)
			sesion := sesionDelTest{
				modelo:    modeloDelJuezA + "#juez_spec:modelo_juez",
				variables: []string{"KITLEGAL_TEST_VERSION=" + caso.version},
			}

			if caso.abre {
				salida, _ := a.abrirSesion(sesion)
				assert.Zero(t, salida.codigo, "la sesión se abre: %s", salida.errores)

				return
			}

			fuera, err := os.OpenRoot(a.fuera)
			require.NoError(t, err)

			defer func() { require.NoError(t, fuera.Close()) }()

			escribirEjecutable(t, fuera, "claude", claudeSustituto)

			salida := a.ejecutar(append([]string{
				"KITLEGAL_CLAUDE_BIN=" + filepath.Join(a.fuera, "claude"),
				"KITLEGAL_TEST_ARGUMENTOS=" + filepath.Join(a.fuera, "argumentos"),
			}, sesion.variables...), guionDelWrapper, "-p", "el prompt", "--model", sesion.modelo)

			assert.Equal(t, 2, salida.codigo, "la sesión no se abre")
			assert.Contains(t, salida.errores, "2.1.314", "dice qué versión hace falta")

			_, err = fuera.Stat("argumentos")
			assert.ErrorIs(t, err, os.ErrNotExist, "Claude Code no llega a abrir la sesión")
		})
	}
}

// TestLaPoliticaDelPasoNoAdmiteEscribirEnLaMemoria: el gancho de las sesiones
// de un paso deniega la llamada de H39 —Edit sobre un fichero de la memoria
// del proyecto— con el motivo, y su tabla de casos, que hito.sh ejecuta al
// arrancar, da lo esperado.
func TestLaPoliticaDelPasoNoAdmiteEscribirEnLaMemoria(t *testing.T) {
	t.Parallel()

	a := nuevoArbol(t)
	a.debeSalirBien(nil, "scripts/workflow/politica-paso.sh", "prueba")

	raiz, err := filepath.EvalSymlinks(a.dir)
	require.NoError(t, err)

	hogar, err := os.UserHomeDir()
	require.NoError(t, err)

	proyecto := strings.Map(func(r rune) rune {
		if ('a' <= r && r <= 'z') || ('A' <= r && r <= 'Z') || ('0' <= r && r <= '9') {
			return r
		}

		return '-'
	}, raiz)

	for _, memoria := range []string{
		filepath.Join(hogar, ".claude", "projects", proyecto, "memory", "cotas-de-tiempo-runner-vs-local.md"),
		filepath.Join(a.config, "projects", proyecto, "memory", "MEMORY.md"),
	} {
		llamada, err := json.Marshal(map[string]any{
			"tool_name":  "Edit",
			"tool_input": map[string]string{"file_path": memoria, "old_string": "a", "new_string": "b"},
		})
		require.NoError(t, err)

		orden := exec.CommandContext(t.Context(), a.ruta("scripts/workflow/sesion-unica.sh"), "gancho") //nolint:gosec // la copia del guion que el test escribió en su temporal.
		orden.Dir = a.dir
		orden.Env = append(slices.Clone(a.entorno), "KITLEGAL_PASO_DE_WORKFLOW=1")
		orden.Stdin = strings.NewReader(string(llamada))

		var errores strings.Builder

		orden.Stderr = &errores

		err = orden.Run()

		var salida *exec.ExitError
		require.ErrorAs(t, err, &salida, "el gancho deniega Edit sobre %s", memoria)
		assert.Equal(t, 2, salida.ExitCode(), "con el código con el que Claude Code deniega la llamada")
		assert.Contains(t, errores.String(), "ADR 0032", "y con el motivo")
	}
}

// valoresDe devuelve los argumentos que siguen a cada aparición de una bandera.
func valoresDe(argumentos []string, bandera string) []string {
	var valores []string

	for i, argumento := range argumentos[:max(len(argumentos)-1, 0)] {
		if argumento == bandera {
			valores = append(valores, argumentos[i+1])
		}
	}

	return valores
}
