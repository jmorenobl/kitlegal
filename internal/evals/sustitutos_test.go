package evals

import (
	"errors"
	"fmt"
	"io/fs"
	"os"
	"os/exec"
	"path/filepath"
	"slices"
	"strconv"
	"strings"
	"syscall"
	"testing"

	"github.com/stretchr/testify/require"
)

// sustitutoDeClaude es el claude que los tests del repartidor ponen delante en
// el PATH en lugar del de verdad, que abriría una sesión con modelo (research
// D18 de H7.3; FR-068): toma el nombre de la sesión del directorio padre del de
// trabajo; marca su llegada creando, en el directorio común, claude/<sesión>,
// que falla si la sesión ya llegó; marca que está abierta creando
// abiertas/<sesión>, que retira al terminar, y anota cuántas sesiones abiertas
// cuenta en abiertas/ nada más llegar, la suya incluida; anota su pid, que
// encabeza su grupo de procesos, sus argumentos, cada uno terminado en NUL, y lo
// que ve —su directorio de trabajo, el kitlegal que resuelve, si ve
// ANTHROPIC_API_KEY y ANTHROPIC_AUTH_TOKEN y cada variable de su lista que está
// definida—; escribe escrito-por-<sesión> en cada directorio en el que
// escribiría Claude Code —el de trabajo, CLAUDE_CONFIG_DIR, TMPDIR,
// CLAUDE_CODE_TMPDIR y KITLEGAL_CACHE_DIR— y, con variableDeEscribirEnHome, en
// HOME; escribe en la salida estándar el transcript
// <sesión>.jsonl del directorio de transcripts, si lo hay; y, con la espera
// anotada al empezarla y al cumplirla, duerme, se envía una señal y sale con el
// código que le digan las variables KITLEGAL_SUSTITUTO_*. La espera de una
// sesión con <sesión>.espera en el directorio de transcripts es la de ese
// fichero. Con variableDeCalentar, termina con 0 sin hacer nada.
//
// Si no ignora TERM, duerme en una subshell que lo atiende, lo anota y sale, y
// la shell, al recibirlo, espera a la subshell antes de salir: la anotación
// dice que TERM llegó a otro proceso del grupo, y si llega solo a la shell, la
// subshell sigue durmiendo y la sesión no sale hasta el KILL. Las dos esperan
// con wait, que TERM interrumpe, y la subshell anota que empieza la espera
// cuando ya lo atiende: TERM no puede caer entre el fork y el exec de sleep,
// donde la shell hija aún tiene su manejador y la señal se pierde. La marca de
// abierta se retira antes de salir, también con TERM; con KILL queda.
const sustitutoDeClaude = `#!/bin/sh
if [ "${KITLEGAL_SUSTITUTO_CALENTAR:-}" = si ]; then exit 0; fi
set -eu
sesion=$(basename "$(cd .. && pwd -P)")
anotaciones="$KITLEGAL_SUSTITUTO_COMUN/claude/$sesion"
mkdir "$anotaciones"
abierta="$KITLEGAL_SUSTITUTO_COMUN/abiertas/$sesion"
mkdir "$abierta"
ls "$KITLEGAL_SUSTITUTO_COMUN/abiertas" | wc -l | tr -d ' ' > "$anotaciones/abiertas-al-llegar"
printf '%s\n' "$$" > "$anotaciones/pid"
printf '%s\0' "$@" > "$anotaciones/argumentos"
{
	printf 'directorio=%s\n' "$(pwd -P)"
	printf 'kitlegal=%s\n' "$(command -v kitlegal || true)"
	for nombre in ANTHROPIC_API_KEY ANTHROPIC_AUTH_TOKEN; do
		eval "definida=\${$nombre+si}"
		printf 've-%s=%s\n' "$nombre" "${definida:-no}"
	done
	for nombre in KITLEGAL_CACHE_DIR HTTP_PROXY HTTPS_PROXY http_proxy https_proxy NO_PROXY no_proxy \
		CLAUDE_CODE_DISABLE_NONESSENTIAL_TRAFFIC CLAUDE_CODE_SUBPROCESS_ENV_SCRUB CLAUDE_CONFIG_DIR TMPDIR \
		CLAUDE_CODE_TMPDIR KITLEGAL_EVALS_TRAZA HOME PATH LANG LC_ALL LC_CTYPE LC_MESSAGES TERM USER LOGNAME SHELL TZ \
		CLAUDE_CODE_OAUTH_TOKEN OTRA_DE_LA_BASE; do
		eval "definida=\${$nombre+si}"
		if [ "$definida" = si ]; then
			eval "valor=\$$nombre"
			printf '%s=%s\n' "$nombre" "$valor"
		fi
	done
} > "$anotaciones/entorno"
for escrito in . "$CLAUDE_CONFIG_DIR" "$TMPDIR" "$CLAUDE_CODE_TMPDIR" "$KITLEGAL_CACHE_DIR"; do
	: >> "$escrito/escrito-por-$sesion"
done
if [ "${KITLEGAL_SUSTITUTO_ESCRIBE_EN_HOME:-no}" = si ]; then : >> "$HOME/escrito-por-$sesion"; fi
transcript="$KITLEGAL_SUSTITUTO_TRANSCRIPTS/$sesion.jsonl"
if [ -f "$transcript" ]; then cat "$transcript"; fi
espera="${KITLEGAL_SUSTITUTO_ESPERA:-0}"
if [ -f "$KITLEGAL_SUSTITUTO_TRANSCRIPTS/$sesion.espera" ]; then
	espera=$(cat "$KITLEGAL_SUSTITUTO_TRANSCRIPTS/$sesion.espera")
fi
if [ "${KITLEGAL_SUSTITUTO_IGNORA_TERM:-no}" = si ]; then
	trap '' TERM
	: > "$anotaciones/espera-empezada"
	sleep "$espera"
else
	trap 'wait; rmdir "$abierta"; exit 143' TERM
	(
		trap ': > "$anotaciones/term-recibido"; exit 143' TERM
		: > "$anotaciones/espera-empezada"
		sleep "$espera" &
		wait $!
	) &
	wait $!
fi
: > "$anotaciones/espera-cumplida"
rmdir "$abierta"
if [ -n "${KITLEGAL_SUSTITUTO_SENAL:-}" ]; then kill -s "$KITLEGAL_SUSTITUTO_SENAL" $$; fi
exit "${KITLEGAL_SUSTITUTO_CODIGO:-0}"
`

// sustitutoDeStrace es el strace de los tests del repartidor (research D18 de
// H7.3): anota sus argumentos, cada uno terminado en NUL, en strace/<sesión> del
// directorio común; escribe en <-o>.<su pid> la traza fija de un solo proceso,
// trazaDelSustituto, como la dejaría strace -ff; y ejecuta lo que va tras --.
// Con variableDeCalentar, termina con 0 sin hacer nada.
const sustitutoDeStrace = `#!/bin/sh
if [ "${KITLEGAL_SUSTITUTO_CALENTAR:-}" = si ]; then exit 0; fi
set -eu
sesion=$(basename "$(cd .. && pwd -P)")
anotaciones="$KITLEGAL_SUSTITUTO_COMUN/strace/$sesion"
mkdir "$anotaciones"
printf '%s\0' "$@" > "$anotaciones/argumentos"
salida=
while [ $# -gt 0 ]; do
	case $1 in
	-o) salida=$2; shift 2 ;;
	--) shift; break ;;
	*) shift ;;
	esac
done
printf '%s' '` + trazaDelSustituto + `' > "$salida.$$"
exec "$@"
`

// trazaDelSustituto es la traza fija que escribe el sustituto de strace: la de
// un solo proceso que ejecuta claude y termina con 0, sin ninguna invocación de
// kitlegal. Va entre comillas simples en el guion, así que no lleva ninguna.
const trazaDelSustituto = `execve("/usr/local/bin/claude", ["claude"], 0x7ffd8f13a6c0 /* 25 vars */) = 0
+++ exited with 0 +++
`

// Los dos sustitutos de los tests del repartidor, por el nombre del programa
// que sustituyen, que es el suyo en el PATH y en el directorio común.
const (
	sustitutoClaude = "claude"
	sustitutoStrace = "strace"
)

// Variables con las que un test gobierna el sustituto de claude, además de las
// que le dicen dónde anotar y de dónde tomar el transcript.
const (
	variableDeEspera         = "KITLEGAL_SUSTITUTO_ESPERA"
	variableDeCodigo         = "KITLEGAL_SUSTITUTO_CODIGO"
	variableDeIgnorarTERM    = "KITLEGAL_SUSTITUTO_IGNORA_TERM"
	variableDeSenal          = "KITLEGAL_SUSTITUTO_SENAL"
	variableDelComun         = "KITLEGAL_SUSTITUTO_COMUN"
	variableDeTranscripts    = "KITLEGAL_SUSTITUTO_TRANSCRIPTS"
	variableDeEscribirEnHome = "KITLEGAL_SUSTITUTO_ESCRIBE_EN_HOME"
)

// accesosDeClaudeCode son las variables con las credenciales de Claude Code que
// un proceso de test puede tener en su entorno: la base de los sustitutos no las
// lleva, para que ningún sustituto anote la de quien ejecuta los tests. El
// nombre no dice «credencial»: gosec (G101) toma por secreto lo que se llama así.
var accesosDeClaudeCode = []string{variableDeLaSuscripcion, "ANTHROPIC_API_KEY", "ANTHROPIC_AUTH_TOKEN"}

// variableDeCalentar, con el valor si, hace que los sustitutos terminen con 0
// sin hacer nada: escribirSustitutos los ejecuta así una vez. En macOS, la
// primera ejecución de un fichero recién escrito tarda de 0,1 a 0,3 s más que
// las siguientes (medido en la tarea T006 de H7.3), y bajo la carga de make ci
// se come el tope de 1 s de TestTopeDeLaSesion antes de que el sustituto llegue
// a empezar.
const variableDeCalentar = "KITLEGAL_SUSTITUTO_CALENTAR"

// claveDelDirectorio es la clave con la que el sustituto de claude anota su
// directorio de trabajo, el físico, sin enlaces.
const claveDelDirectorio = "directorio"

// Lo que el sustituto de claude deja, además de sus anotaciones: en el
// directorio común, la marca de cada sesión abierta; y, en cada directorio en
// el que escribe, el fichero que nombra su sesión detrás de este prefijo.
const (
	directorioDeAbiertas = "abiertas"
	prefijoDeLoEscrito   = "escrito-por-"
)

// sustitutos es dónde están los sustitutos de claude y strace de un test: el
// directorio que se pone delante en el PATH, el común en el que anotan lo que
// ven y el de los transcripts que escribe claude, con la espera de cada sesión
// que no tiene la de la base.
type sustitutos struct {
	bin         string
	comun       string
	transcripts string
}

// escribirSustitutos escribe los sustitutos de claude y strace, ejecutables, en
// un directorio temporal del test, a través de un os.Root (research V8b de
// H7.3), y los ejecuta una vez con variableDeCalentar; y crea el directorio
// común, con claude/, strace/ y abiertas/ dentro, y el de los transcripts.
func escribirSustitutos(t *testing.T) sustitutos {
	t.Helper()

	bin := t.TempDir()

	raiz, err := os.OpenRoot(bin)
	require.NoError(t, err)

	defer func() { require.NoError(t, raiz.Close()) }()

	guiones := map[string]string{sustitutoClaude: sustitutoDeClaude, sustitutoStrace: sustitutoDeStrace}
	for programa, guion := range guiones {
		require.NoError(t, raiz.WriteFile(programa, []byte(guion), 0o755))
		calentar(t, filepath.Join(bin, programa))
	}

	comun := t.TempDir()
	for _, directorio := range []string{sustitutoClaude, sustitutoStrace, directorioDeAbiertas} {
		require.NoError(t, os.Mkdir(filepath.Join(comun, directorio), 0o750))
	}

	return sustitutos{bin: bin, comun: comun, transcripts: t.TempDir()}
}

// calentar ejecuta una vez el sustituto de la ruta dada con variableDeCalentar,
// que lo hace terminar con 0 sin hacer nada, y sin ninguna otra variable.
func calentar(t *testing.T, sustituto string) {
	t.Helper()

	orden := exec.CommandContext(t.Context(), sustituto)
	orden.Env = []string{variableDeCalentar + "=si"}

	salida, err := orden.CombinedOutput()
	require.NoErrorf(t, err, "el sustituto %s se ejecuta sin hacer nada: %s", sustituto, salida)
}

// base es el entorno base del repartidor en un test: el del proceso sin sus
// credenciales de Claude Code, con el directorio de los sustitutos delante en el
// PATH, las rutas del común y de los transcripts y las variables dadas, que
// sustituyen a las que el proceso tenga con el mismo nombre.
func (s sustitutos) base(variables ...string) []string {
	return sobreLaBase(os.Environ(), append([]string{
		"PATH=" + s.bin + string(os.PathListSeparator) + os.Getenv("PATH"),
		variableDelComun + "=" + s.comun,
		variableDeTranscripts + "=" + s.transcripts,
	}, variables...), accesosDeClaudeCode...)
}

// kitlegalQueNoHaceNada es un kitlegal que termina con 0 sin hacer nada: el que
// los tests del sondeo ponen en su lugar sin construir el binario, y el de la
// base, que las sesiones del sondeo no tienen que resolver.
const kitlegalQueNoHaceNada = "#!/bin/sh\nexit 0\n"

// escribirElClaudeDelSondeo escribe, en un directorio temporal del test y a
// través de un os.Root, el claude que ven las sesiones del sondeo en los tests
// de sondear y un kitlegal que no hace nada, y devuelve ese directorio, que va
// delante en el PATH de la base. Las sesiones del sondeo solo ven las variables
// de contracts/sondeo.md §5 de H7.3, y no las que dicen al sustituto de claude
// dónde anotar ni de dónde tomar el transcript: este claude las pone —las del
// común y los transcripts, la de escribir también en HOME y las dadas,
// nombre=valor—, con los valores que lleva escritos, y ejecuta con exec el de
// los sustitutos. Lo ejecuta una vez con variableDeCalentar, como
// escribirSustitutos.
func (s sustitutos) escribirElClaudeDelSondeo(t *testing.T, variables ...string) string {
	t.Helper()

	dir := t.TempDir()

	raiz, err := os.OpenRoot(dir)
	require.NoError(t, err)

	defer func() { require.NoError(t, raiz.Close()) }()

	var guion strings.Builder

	guion.WriteString("#!/bin/sh\n")

	for _, variable := range slices.Concat([]string{
		variableDelComun + "=" + s.comun,
		variableDeTranscripts + "=" + s.transcripts,
		variableDeEscribirEnHome + "=si",
	}, variables) {
		nombre, valor, _ := strings.Cut(variable, "=")
		fmt.Fprintf(&guion, "%s=%s\nexport %s\n", nombre, entreComillasSimples(valor), nombre)
	}

	fmt.Fprintf(&guion, "exec %s \"$@\"\n", entreComillasSimples(filepath.Join(s.bin, sustitutoClaude)))

	require.NoError(t, raiz.WriteFile(sustitutoClaude, []byte(guion.String()), 0o755))
	require.NoError(t, raiz.WriteFile(programaDeLasConsultas, []byte(kitlegalQueNoHaceNada), 0o755))
	calentar(t, filepath.Join(dir, sustitutoClaude))

	return dir
}

// entreComillasSimples es el valor entre comillas simples, como lo lee sh
// literalmente, con cada comilla simple que lleve cerrada, escapada y vuelta a
// abrir.
func entreComillasSimples(valor string) string {
	return "'" + strings.ReplaceAll(valor, "'", `'\''`) + "'"
}

// escribirTranscript deja el transcript que el sustituto de claude escribe en
// la sesión del nombre dado.
func (s sustitutos) escribirTranscript(t *testing.T, sesion, contenido string) {
	t.Helper()

	require.NoError(t, os.WriteFile(filepath.Join(s.transcripts, sesion+".jsonl"), []byte(contenido), 0o600))
}

// escribirEspera deja la espera, en segundos, del sustituto de claude en la
// sesión del nombre dado, en lugar de la de la base.
func (s sustitutos) escribirEspera(t *testing.T, sesion string, segundos int) {
	t.Helper()

	require.NoError(t, os.WriteFile(filepath.Join(s.transcripts, sesion+".espera"), []byte(strconv.Itoa(segundos)), 0o600))
}

// abiertasAlLlegar es cuántas sesiones abiertas contó el sustituto de claude al
// llegar a la sesión, la suya incluida.
func (s sustitutos) abiertasAlLlegar(t *testing.T, sesion string) int {
	t.Helper()

	return s.enteroAnotado(t, sesion, "abiertas-al-llegar")
}

// grupoDeProcesos es el del sustituto de claude en la sesión: el de su pid, que
// es el del guion de la sesión, abierto en su propio grupo.
func (s sustitutos) grupoDeProcesos(t *testing.T, sesion string) int {
	t.Helper()

	return s.enteroAnotado(t, sesion, "pid")
}

// enteroAnotado es el entero, en su línea, que el sustituto de claude anotó en
// la sesión con ese nombre.
func (s sustitutos) enteroAnotado(t *testing.T, sesion, anotacion string) int {
	t.Helper()

	contenido, err := leerFichero(filepath.Join(s.comun, sustitutoClaude, sesion, anotacion))
	require.NoError(t, err)

	entero, err := strconv.Atoi(strings.TrimSuffix(string(contenido), "\n"))
	require.NoErrorf(t, err, "la anotación %s de la sesión %s es un entero en su línea", anotacion, sesion)

	return entero
}

// grupoTerminado dice si ya no queda ningún proceso en el grupo: el sistema
// responde ESRCH a una señal 0 al grupo. No recibe el test porque es la
// condición de assert.Eventually, que la evalúa en otra gorrutina.
func grupoTerminado(grupo int) bool {
	return errors.Is(syscall.Kill(-grupo, 0), syscall.ESRCH)
}

// llego dice si el sustituto de claude llegó a ejecutarse en la sesión.
func (s sustitutos) llego(t *testing.T, sesion string) bool {
	t.Helper()

	return existe(t, filepath.Join(s.comun, sustitutoClaude, sesion))
}

// esperaEmpezada es la marca que el sustituto de claude deja en la sesión al
// empezar su espera, cuando ya atiende TERM, o lo ignora, en todos sus
// procesos.
func (s sustitutos) esperaEmpezada(sesion string) string {
	return filepath.Join(s.comun, sustitutoClaude, sesion, "espera-empezada")
}

// cumplioLaEspera dice si el sustituto de claude cumplió en la sesión la espera
// que le dijeron: uno que el tope corta, con TERM o con KILL, no la cumple.
func (s sustitutos) cumplioLaEspera(t *testing.T, sesion string) bool {
	t.Helper()

	return existe(t, filepath.Join(s.comun, sustitutoClaude, sesion, "espera-cumplida"))
}

// recibioTERM dice si TERM llegó en la sesión a la subshell en la que duerme el
// sustituto de claude, otro proceso de su grupo, y la subshell lo atendió.
func (s sustitutos) recibioTERM(t *testing.T, sesion string) bool {
	t.Helper()

	return existe(t, filepath.Join(s.comun, sustitutoClaude, sesion, "term-recibido"))
}

// existe dice si hay algo en la ruta; cualquier otro error al mirarla falla el
// test.
func existe(t *testing.T, ruta string) bool {
	t.Helper()

	_, err := os.Lstat(ruta)
	if errors.Is(err, fs.ErrNotExist) {
		return false
	}

	require.NoError(t, err)

	return true
}

// argumentos son los que recibió el sustituto del programa en la sesión, en su
// orden.
func (s sustitutos) argumentos(t *testing.T, programa, sesion string) []string {
	t.Helper()

	contenido, err := leerFichero(filepath.Join(s.comun, programa, sesion, "argumentos"))
	require.NoError(t, err)
	require.NotEmpty(t, contenido)
	require.Equal(t, byte(0), contenido[len(contenido)-1], "cada argumento termina en NUL")

	return strings.Split(strings.TrimSuffix(string(contenido), "\x00"), "\x00")
}

// entorno es lo que el sustituto de claude anotó que ve en la sesión, por
// clave: el directorio de trabajo, el kitlegal que resuelve, si ve
// ANTHROPIC_API_KEY y ANTHROPIC_AUTH_TOKEN (ve-<variable>, si o no) y cada
// variable de su lista que está definida.
func (s sustitutos) entorno(t *testing.T, sesion string) map[string]string {
	t.Helper()

	contenido, err := leerFichero(filepath.Join(s.comun, sustitutoClaude, sesion, "entorno"))
	require.NoError(t, err)

	anotado := map[string]string{}

	for linea := range strings.Lines(string(contenido)) {
		clave, valor, ok := strings.Cut(strings.TrimSuffix(linea, "\n"), "=")
		require.Truef(t, ok, "el entorno anotado lleva clave=valor en %q", linea)
		require.NotContainsf(t, anotado, clave, "la clave %s se anota una sola vez", clave)

		anotado[clave] = valor
	}

	return anotado
}
