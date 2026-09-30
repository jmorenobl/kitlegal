//go:build integration

// La preparación del árbol del sondeo de verdad: el go install del binario y,
// con él, el skills install, en un temporal del test (contracts/sondeo.md §3.3 y
// §7 de H7.3; FR-062, FR-064). Lleva la etiqueta integration porque ejecuta el
// go command y compila el binario: make ci lo ejecuta con test-integration, y el
// lint lo alcanza con run.build-tags.
package evals

import (
	"encoding/json"
	"errors"
	"io/fs"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// skillDelArbol es la SKILL.md de boe-legislacion del árbol de trabajo,
// relativa al directorio de este paquete: la que el sondeo tiene que instalar.
const skillDelArbol = "../../skills/boe-legislacion/SKILL.md"

// TestPrepararElArbolDelSondeo prepara el árbol del sondeo de verdad en un
// directorio temporal del test (contracts/sondeo.md §3.3 y §7 de H7.3; FR-062,
// FR-064): el kitlegal del bin/ del temporal existe y es ejecutable, y la
// SKILL.md de boe-legislacion que ve Claude Code en el HOME del temporal es,
// byte a byte, la del árbol. El HOME de la base es otro temporal, con las cachés
// de Go del proceso de test y sin red (GOPROXY=off, GOFLAGS=-mod=readonly), y
// no gana nada fuera de lo que el go command guarda de sí mismo, su telemetría,
// que no se apaga desde el entorno y que ejecutar el sondeo, que es un go test,
// escribe igual: ni el binario, ni las skills, ni su manifiesto.
func TestPrepararElArbolDelSondeo(t *testing.T) {
	t.Parallel()

	temporal := t.TempDir()
	personal := t.TempDir()
	caches := leerLasCachesDelGoCommand(t)

	base := sobreLaBase(os.Environ(), []string{
		"HOME=" + personal,
		"GOCACHE=" + caches.Compilados,
		"GOMODCACHE=" + caches.Modulos,
		"GOPROXY=off",
		"GOFLAGS=-mod=readonly",
	})

	require.NoError(t, prepararElArbol(temporal, base))

	binario, err := os.Stat(filepath.Join(temporal, "bin", programaDeLasConsultas))
	require.NoError(t, err, "el kitlegal del temporal existe")
	assert.True(t, binario.Mode().IsRegular(), "el kitlegal del temporal es un fichero")
	assert.NotZero(t, binario.Mode().Perm()&0o100, "el kitlegal del temporal es ejecutable")

	instalada, err := leerFichero(filepath.Join(temporal, "home", ".claude", "skills", "boe-legislacion", "SKILL.md"))
	require.NoError(t, err)

	delArbol, err := leerFichero(skillDelArbol)
	require.NoError(t, err)

	assert.Equal(t, string(delArbol), string(instalada), "la SKILL.md instalada es la del árbol")

	exigirSoloLoDelGoCommand(t, personal, base)
}

// cachesDelGoCommand son las cachés de módulos y de construcción del go command
// del proceso de test, que la base del test reutiliza para no descargar nada.
type cachesDelGoCommand struct {
	Modulos    string `json:"GOMODCACHE"`
	Compilados string `json:"GOCACHE"`
}

// leerLasCachesDelGoCommand lee con go env las cachés del proceso de test.
func leerLasCachesDelGoCommand(t *testing.T) cachesDelGoCommand {
	t.Helper()

	orden := exec.CommandContext(t.Context(), "go", "env", "-json", "GOMODCACHE", "GOCACHE")
	orden.Dir = raizDelRepositorio

	var caches cachesDelGoCommand

	require.NoError(t, json.Unmarshal(salidaDelGoCommand(t, orden), &caches))
	require.NotEmpty(t, caches.Modulos, "go env da GOMODCACHE")
	require.NotEmpty(t, caches.Compilados, "go env da GOCACHE")

	return caches
}

// exigirSoloLoDelGoCommand exige que en el HOME de la base no haya nada que no
// sea el directorio de telemetría del go command, que da go env GOTELEMETRYDIR
// con la base, lo que hay dentro o una carpeta que lo contiene.
func exigirSoloLoDelGoCommand(t *testing.T, personal string, base []string) {
	t.Helper()

	orden := exec.CommandContext(t.Context(), "go", "env", "GOTELEMETRYDIR")
	orden.Dir = raizDelRepositorio
	orden.Env = base

	telemetria := strings.TrimSuffix(string(salidaDelGoCommand(t, orden)), "\n")
	require.True(t, filepath.IsAbs(telemetria), "go env da GOTELEMETRYDIR: %q", telemetria)

	var ajenos []string

	err := fs.WalkDir(os.DirFS(personal), ".", func(ruta string, entrada fs.DirEntry, err error) error {
		if err != nil || ruta == "." {
			return err
		}

		absoluta := filepath.Join(personal, filepath.FromSlash(ruta))
		dentro := absoluta == telemetria || strings.HasPrefix(absoluta, telemetria+string(filepath.Separator))
		contiene := entrada.IsDir() && strings.HasPrefix(telemetria, absoluta+string(filepath.Separator))

		if !dentro && !contiene {
			ajenos = append(ajenos, ruta)
		}

		return nil
	})
	require.NoError(t, err)

	assert.Empty(t, ajenos, "el HOME de la base solo gana la telemetría del go command, en %s", telemetria)
}

// salidaDelGoCommand ejecuta una orden del go command y devuelve su salida
// estándar, o hace fallar el test con su salida de error.
func salidaDelGoCommand(t *testing.T, orden *exec.Cmd) []byte {
	t.Helper()

	salida, err := orden.Output()
	if err != nil {
		var fallo *exec.ExitError
		if errors.As(err, &fallo) {
			t.Fatalf("%s: %v\n%s", orden, err, fallo.Stderr)
		}

		t.Fatalf("%s: %v", orden, err)
	}

	return salida
}
