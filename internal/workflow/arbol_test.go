package workflow

import (
	"encoding/json"
	"errors"
	"os"
	"os/exec"
	"path/filepath"
	"slices"
	"strings"
	"syscall"
	"testing"

	"github.com/stretchr/testify/require"
)

const (
	// raizDelRepositorio es la del proyecto, vista desde este paquete.
	raizDelRepositorio = "../.."

	// hitoDelTest es el hito del árbol temporal, y directorioDelHito, su
	// directorio de feature.
	hitoDelTest       = "H0"
	directorioDelHito = "specs/000-h0-hito-del-test"
	gatesDelHito      = directorioDelHito + "/gates"

	// ficheroDelHitoDelTest es el fichero del hito en el docs/roadmap/ del
	// árbol temporal, y hitoDelRoadmapDelTest, su contenido: el informe final
	// saca de él el título del hito.
	ficheroDelHitoDelTest = "docs/roadmap/in-progress/H0-hito-del-test.md"
	hitoDelRoadmapDelTest = "---\nid: H0\nfase: \"1\"\n---\n\n# H0 · Hito del test\n\nLa entrada del hito.\n"

	// ficheroDelProducto es lo que el hito cambia fuera de gates/.
	ficheroDelProducto = "producto.txt"
)

// guionesDelArbol son los que el test copia al árbol temporal, con su ruta
// relativa a la raíz: los que ejerce y los que estos llaman.
var guionesDelArbol = []string{
	"scripts/claude-modelo.sh",
	"scripts/workflow/comun.sh",
	"scripts/workflow/gate.sh",
	"scripts/workflow/global.sh",
	"scripts/workflow/hito-a-done.sh",
	"scripts/workflow/informe.sh",
	"scripts/workflow/marcas.sh",
	"scripts/workflow/modelos.sh",
	"scripts/workflow/revision.sh",
	"scripts/workflow/sesion-unica.sh",
	"scripts/workflow/tarea.sh",
}

// arbol es un repositorio de git en un directorio temporal con las copias de
// los guiones del workflow y la de workflow.yml, un hito con su directorio de
// feature y una rama `hito` que cambia, respecto a main, lo que entrega el hito.
type arbol struct {
	t   *testing.T
	dir string
	// raiz da acceso a los ficheros del árbol sin salir de él.
	raiz *os.Root
	// config es el directorio de Claude Code del árbol (CLAUDE_CONFIG_DIR):
	// donde el sustituto de Claude Code escribe los transcripts.
	config string
	// fuera es un directorio temporal fuera del repositorio, para lo que el
	// test pone a disposición de los guiones sin que cuente como un cambio
	// del árbol: el Claude Code sustituto.
	fuera   string
	entorno []string
}

// salidaDeOrden es lo que deja una orden: su salida estándar, la de error y su
// código de salida.
type salidaDeOrden struct {
	estandar string
	errores  string
	codigo   int
}

// nuevoArbol monta el árbol del test. Se salta donde falta bash, git o jq.
func nuevoArbol(t *testing.T) *arbol {
	t.Helper()

	for _, herramienta := range []string{"bash", "git", "jq"} {
		if _, err := exec.LookPath(herramienta); err != nil {
			t.Skipf("los guiones del workflow necesitan %s, que no está en el PATH", herramienta)
		}
	}

	dir := t.TempDir()
	raiz, err := os.OpenRoot(dir)
	require.NoError(t, err)
	t.Cleanup(func() { require.NoError(t, raiz.Close()) })

	a := &arbol{t: t, dir: dir, raiz: raiz, config: filepath.Join(dir, ".config-claude"), fuera: t.TempDir()}
	a.entorno = entornoDelArbol(a.config)

	for _, guion := range guionesDelArbol {
		a.escribirEjecutable(guion, leerDelRepositorio(t, guion))
	}

	flujo := leerDelRepositorio(t, rutaDelWorkflow)

	a.escribir(rutaDelWorkflow, flujo)
	a.escribir(".specify/feature.json", `{"feature_directory": "`+directorioDelHito+`"}`)
	a.escribir(".gitignore", ".config-claude/\n.specify/feature.json\n.specify/sesiones-sin-feature.jsonl\n")
	a.escribir(ficheroDelHitoDelTest, hitoDelRoadmapDelTest)

	for _, artefacto := range []string{"spec.md", "plan.md", "tasks.md"} {
		a.escribir(directorioDelHito+"/"+artefacto, "")
	}

	a.git("init", "-q", "-b", "main")
	a.commit("main del test")
	a.git("switch", "-q", "-c", "hito")
	a.escribir(ficheroDelProducto, "lo que entrega el hito\n")
	a.commit("feat(H0): lo que entrega el hito")

	return a
}

// entornoDelArbol es el del proceso sin ninguna variable de git —con la de un
// gancho, git trabajaría en el repositorio del proyecto— ni de kitlegal —el
// testigo de un run, el repuesto del supervisor o el origen de una sesión
// cambiarían lo que se registra—, sin la configuración global de git, con un
// nombre y un correo fijos y con el directorio de Claude Code del árbol.
func entornoDelArbol(config string) []string {
	entorno := slices.DeleteFunc(os.Environ(), func(variable string) bool {
		return strings.HasPrefix(variable, "GIT_") || strings.HasPrefix(variable, "KITLEGAL_") ||
			strings.HasPrefix(variable, "CLAUDE_")
	})

	return append(entorno,
		"GIT_CONFIG_GLOBAL="+os.DevNull, "GIT_CONFIG_NOSYSTEM=1",
		"GIT_AUTHOR_NAME=kitlegal", "GIT_AUTHOR_EMAIL=test@kitlegal.invalid",
		"GIT_COMMITTER_NAME=kitlegal", "GIT_COMMITTER_EMAIL=test@kitlegal.invalid",
		"CLAUDE_CONFIG_DIR="+config)
}

func (a *arbol) ruta(relativa string) string {
	return filepath.Join(a.dir, filepath.FromSlash(relativa))
}

// leerDelRepositorio lee un fichero del proyecto, con su ruta relativa a la raíz.
func leerDelRepositorio(t *testing.T, relativa string) string {
	t.Helper()

	raiz, err := os.OpenRoot(raizDelRepositorio)
	require.NoError(t, err)

	defer func() { require.NoError(t, raiz.Close()) }()

	contenido, err := raiz.ReadFile(relativa)
	require.NoError(t, err)

	return string(contenido)
}

func (a *arbol) escribir(relativa, contenido string) {
	a.t.Helper()

	require.NoError(a.t, a.raiz.MkdirAll(filepath.Dir(relativa), 0o750))
	require.NoError(a.t, a.raiz.WriteFile(relativa, []byte(contenido), 0o600))
}

// escribirEjecutable escribe un guion con permiso de ejecución. Con el candado
// de lectura de los fork, ningún proceso que otro test lance a la vez hereda
// abierto el descriptor de escritura, que haría fallar la ejecución con
// ETXTBSY.
func (a *arbol) escribirEjecutable(relativa, contenido string) {
	a.t.Helper()

	require.NoError(a.t, a.raiz.MkdirAll(filepath.Dir(relativa), 0o750))
	escribirEjecutable(a.t, a.raiz, relativa, contenido)
}

// escribirEjecutable escribe un guion con permiso de ejecución, con el candado
// de lectura de los fork (arbol.escribirEjecutable).
func escribirEjecutable(t *testing.T, raiz *os.Root, relativa, contenido string) {
	t.Helper()

	syscall.ForkLock.RLock()
	defer syscall.ForkLock.RUnlock()

	require.NoError(t, raiz.WriteFile(relativa, []byte(contenido), 0o755))
}

func (a *arbol) leer(relativa string) string {
	a.t.Helper()

	contenido, err := a.raiz.ReadFile(relativa)
	require.NoError(a.t, err)

	return string(contenido)
}

func (a *arbol) existe(relativa string) bool {
	_, err := a.raiz.Stat(relativa)

	return err == nil
}

// escribirJSON escribe un valor serializado.
func (a *arbol) escribirJSON(relativa string, valor any) {
	a.t.Helper()

	contenido, err := json.Marshal(valor)
	require.NoError(a.t, err)
	a.escribir(relativa, string(contenido))
}

// leerJSON lee un fichero JSON en un valor.
func (a *arbol) leerJSON(relativa string, destino any) {
	a.t.Helper()

	require.NoError(a.t, json.Unmarshal([]byte(a.leer(relativa)), destino))
}

func (a *arbol) git(argumentos ...string) string {
	a.t.Helper()

	orden := exec.CommandContext(a.t.Context(), "git", argumentos...) //nolint:gosec // argumentos literales de los tests de este paquete.
	orden.Dir, orden.Env = a.dir, a.entorno

	salida, err := orden.CombinedOutput()
	require.NoErrorf(a.t, err, "git %v: %s", argumentos, salida)

	return strings.TrimSpace(string(salida))
}

func (a *arbol) commit(mensaje string) {
	a.t.Helper()

	a.git("add", "-A")
	a.git("commit", "-q", "--allow-empty", "-m", mensaje)
}

func (a *arbol) cabeza() string {
	a.t.Helper()

	return a.git("rev-parse", "HEAD")
}

// ejecutar lanza un guion del árbol con el entorno del árbol y, delante, las
// variables que se le den; devuelve lo que deja sin exigir que salga con 0.
func (a *arbol) ejecutar(variables []string, guion string, argumentos ...string) salidaDeOrden {
	a.t.Helper()

	// El guion es la copia que el test escribió en su temporal.
	orden := exec.CommandContext(a.t.Context(), a.ruta(guion), argumentos...) //nolint:gosec // la copia del guion que el test escribió en su temporal.
	orden.Dir, orden.Env = a.dir, append(slices.Clone(a.entorno), variables...)

	var estandar, errores strings.Builder

	orden.Stdout, orden.Stderr = &estandar, &errores

	codigo := 0
	if err := orden.Run(); err != nil {
		var salida *exec.ExitError
		require.ErrorAsf(a.t, err, &salida, "%s no llega a ejecutarse: %s", guion, errores.String())

		codigo = salida.ExitCode()
	}

	return salidaDeOrden{estandar: estandar.String(), errores: errores.String(), codigo: codigo}
}

// debeSalirBien lanza un guion del árbol y exige que salga con 0.
func (a *arbol) debeSalirBien(variables []string, guion string, argumentos ...string) salidaDeOrden {
	a.t.Helper()

	salida := a.ejecutar(variables, guion, argumentos...)
	require.Zerof(a.t, salida.codigo, "%s %v sale con %d: %s", guion, argumentos, salida.codigo, salida.errores)

	return salida
}

// noExiste dice si no existe un fichero del árbol.
func (a *arbol) noExiste(relativa string) bool {
	_, err := a.raiz.Stat(relativa)

	return errors.Is(err, os.ErrNotExist)
}
