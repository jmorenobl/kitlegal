package app_test

import (
	"encoding/json"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// variableDirDePrueba es la del entorno por la que la ruta de --dir llega a los
// guiones de sh de TestOrdenesDeDoctorConElBinario, sin ir en ningún argumento
// de un subproceso.
const variableDirDePrueba = "KITLEGAL_DIR_DE_PRUEBA"

// TestOrdenesDeDoctorConElBinario fija la garantía (i) de FR-066
// (contracts/applet-skills.md §6) con el binario de e2e de desarrollo, es decir,
// con el análisis de la invocación real, que las pruebas del dominio imitan: con
// cada ruta de --dir —también una que empieza por «-», que ese análisis leería
// como otra bandera si fuera en la palabra siguiente a --dir—, instalada
// legal-core y editado su SKILL.md, doctor sale con 1 y un solo hallazgo; su
// orden, ejecutada por sh en el mismo directorio de trabajo, sale con 0 y deja
// SKILL.md como lo empotra el binario; y el doctor siguiente sale con 0 y sin
// hallazgos.
func TestOrdenesDeDoctorConElBinario(t *testing.T) {
	t.Parallel()

	require.NoError(t, entorno.err)

	empotrado, err := os.ReadFile(filepath.Join(entorno.variables[variableSkills], "legal-core", "SKILL.md"))
	require.NoError(t, err)

	for _, dir := range []string{"destino", "-raro", "--", "-o'tro y más"} {
		t.Run(dir, func(t *testing.T) {
			t.Parallel()

			proyecto := t.TempDir()

			_, errores, codigo := enSh(t, proyecto, dir,
				`kitlegal skills install legal-core --dir="$`+variableDirDePrueba+`" --json`)
			require.Equal(t, 0, codigo, errores)

			// El fichero se edita y se relee por un os.Root del proyecto, que no
			// deja salir de él.
			raiz, err := os.OpenRoot(proyecto)
			require.NoError(t, err)
			t.Cleanup(func() { require.NoError(t, raiz.Close()) })

			skillMd := filepath.Join(dir, "legal-core", "SKILL.md")
			require.NoError(t, raiz.WriteFile(skillMd, []byte("editado a mano\n"), 0o600))

			orden := ordenDelUnicoHallazgo(t, proyecto, dir, filepath.ToSlash(skillMd))

			salida, errores, codigo := enSh(t, proyecto, dir, orden)
			require.Equal(t, 0, codigo, "la orden de doctor sale con 0 (FR-066 (i)): %s\n%s%s", orden, salida, errores)

			arreglado, err := raiz.ReadFile(skillMd)
			require.NoError(t, err)
			assert.Equal(t, string(empotrado), string(arreglado), "la orden deja SKILL.md como lo empotra el binario")

			salida, errores, codigo = enSh(t, proyecto, dir, `kitlegal skills doctor --dir="$`+variableDirDePrueba+`" --json`)
			require.Equal(t, 0, codigo, "tras la orden, doctor no encuentra nada: %s", errores)

			var sobre struct {
				OK   bool `json:"ok"`
				Data struct {
					Hallazgos []json.RawMessage `json:"hallazgos"`
				} `json:"data"`
			}

			require.NoError(t, json.Unmarshal([]byte(salida), &sobre), salida)
			assert.True(t, sobre.OK, salida)
			assert.NotNil(t, sobre.Data.Hallazgos, "doctor da la lista de hallazgos, vacía: %s", salida)
			assert.Empty(t, sobre.Data.Hallazgos, salida)
		})
	}
}

// ordenDelUnicoHallazgo ejecuta doctor con la ruta de --dir en el proyecto,
// exige que salga con 1 con un solo hallazgo, el del fichero editado de ruta, y
// devuelve su orden, tal como la escribe el mensaje del sobre de fallo:
// «<clase>: <ruta>: <orden>» (contracts/applet-skills.md §5).
func ordenDelUnicoHallazgo(t *testing.T, proyecto, dir, ruta string) string {
	t.Helper()

	salida, errores, codigo := enSh(t, proyecto, dir, `kitlegal skills doctor --dir="$`+variableDirDePrueba+`" --json`)
	require.Equal(t, 1, codigo, "doctor con un fichero editado sale con 1: %s%s", salida, errores)

	var sobre struct {
		OK   bool `json:"ok"`
		Data struct {
			Mensaje string `json:"mensaje"`
		} `json:"data"`
	}

	require.NoError(t, json.Unmarshal([]byte(salida), &sobre), salida)
	require.False(t, sobre.OK, salida)

	cabecera, linea, hay := strings.Cut(sobre.Data.Mensaje, "\n")
	require.True(t, hay, "la cabecera y una línea por hallazgo: %q", sobre.Data.Mensaje)
	require.Equal(t, "skills doctor: 1 hallazgo:", cabecera)
	require.NotContains(t, linea, "\n", "un solo hallazgo: %q", sobre.Data.Mensaje)

	orden, hay := strings.CutPrefix(linea, "fichero editado: "+ruta+": ")
	require.True(t, hay, "el hallazgo es el del fichero editado %s: %q", ruta, linea)

	return orden
}

// enSh ejecuta el guion con sh, que lo lee de su entrada estándar, en el
// directorio de trabajo, con el PATH del proceso —que TestMain encabeza con el
// binario de desarrollo—, la ruta de --dir en variableDirDePrueba y nada más
// del entorno: ni HOME, de modo que nada de la cuenta de quien ejecuta los
// tests entra en la prueba. Devuelve la salida estándar, la de error y el
// código de salida.
func enSh(t *testing.T, directorio, dir, guion string) (string, string, int) {
	t.Helper()

	orden := exec.CommandContext(t.Context(), "sh")
	orden.Dir = directorio
	orden.Env = []string{"PATH=" + os.Getenv("PATH"), variableDirDePrueba + "=" + dir}
	orden.Stdin = strings.NewReader(guion + "\n")

	var salida, errores strings.Builder

	orden.Stdout = &salida
	orden.Stderr = &errores

	if err := orden.Run(); err != nil {
		var terminado *exec.ExitError
		require.ErrorAs(t, err, &terminado, "sh no se pudo ejecutar")
	}

	return salida.String(), errores.String(), orden.ProcessState.ExitCode()
}
