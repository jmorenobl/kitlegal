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
// fichero, y la de una sesión con <sesión>.tras, que lleva una ruta, no empieza
// hasta que esa ruta existe: así una sesión no termina antes de que otra haya
// empezado su espera, llegue cuando llegue. Con variableDeCalentar, termina con
// 0 sin hacer nada.
//
// Si no ignora TERM, quien duerme es un sh aparte, hijo de una subshell que
// atiende TERM, lo anota y sale, y la shell, al recibirlo, sale cuando ha
// terminado la subshell: la anotación dice que TERM llegó a otro proceso del
// grupo, y si llega solo a la shell, la subshell sigue esperando y la sesión no
// sale hasta el KILL. Todo va en primer plano, porque una shell atiende la señal
// en cuanto termina la orden que tiene en marcha, llegue cuando llegue, y wait
// solo se interrumpe si ya ha empezado a esperar: con sleep en segundo plano y
// wait, un TERM que caía justo antes se quedaba sin atender hasta que terminaba
// sleep, y con 16 sustitutos a la vez 15 de 3000 no anotaban el TERM. Y la marca
// de que empieza la espera la anota quien duerme, tras su exec y sin manejador
// de TERM: desde que existe, TERM lo termina siempre, y no puede caer entre el
// fork y el exec, donde la shell hija aún tiene el manejador de su madre y la
// señal se pierde. La subshell deja dicho en la salida de error que su hijo
// terminó por una señal. La marca de abierta se retira antes de salir, también
// con TERM; con KILL queda.
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
if [ -f "$KITLEGAL_SUSTITUTO_TRANSCRIPTS/$sesion.tras" ]; then
	tras=$(cat "$KITLEGAL_SUSTITUTO_TRANSCRIPTS/$sesion.tras")
	while [ ! -e "$tras" ]; do sleep 0.05; done
fi
if [ "${KITLEGAL_SUSTITUTO_IGNORA_TERM:-no}" = si ]; then
	trap '' TERM
	: > "$anotaciones/espera-empezada"
	sleep "$espera"
else
	trap 'rmdir "$abierta"; exit 143' TERM
	(
		trap ': > "$anotaciones/term-recibido"; exit 143' TERM
		/bin/sh -c ': > "$1"; exec sleep "$2"' sh "$anotaciones/espera-empezada" "$espera" || exit $?
	) || exit $?
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
		require.NoError(t, escribirEjecutable(raiz, programa, guion))
		calentar(t, filepath.Join(bin, programa))
	}

	comun := t.TempDir()
	for _, directorio := range []string{sustitutoClaude, sustitutoStrace, directorioDeAbiertas} {
		require.NoError(t, os.Mkdir(filepath.Join(comun, directorio), 0o750))
	}

	return sustitutos{bin: bin, comun: comun, transcripts: t.TempDir()}
}

// escribirEjecutable escribe, a través de la raíz dada, el ejecutable del
// nombre dado con el guion dado, con syscall.ForkLock tomado para lectura. Un
// proceso que otro test en paralelo crea hereda, entre su fork y su exec, una
// copia de cada descriptor abierto, también el de un ejecutable a medio
// escribir; y en Linux, ejecutar un fichero que algún proceso tiene abierto
// para escribir falla con ETXTBSY, «text file busy» (golang/go#22315): el
// calentar que sigue fallaba así en 11 de 50 ejecuciones del paquete. Go toma
// ForkLock para escritura en cada fork, así que con él tomado para lectura
// ningún fork ocurre mientras el descriptor está abierto, y cerrado ya no lo
// hereda ningún proceso.
func escribirEjecutable(raiz *os.Root, nombre, guion string) error {
	syscall.ForkLock.RLock()
	defer syscall.ForkLock.RUnlock()

	return raiz.WriteFile(nombre, []byte(guion), 0o755)
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

	require.NoError(t, escribirEjecutable(raiz, sustitutoClaude, guion.String()))
	require.NoError(t, escribirEjecutable(raiz, programaDeLasConsultas, kitlegalQueNoHaceNada))
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

// empezarLaEsperaCuandoExista deja dicho al sustituto de claude de la sesión
// del nombre dado que no empiece su espera hasta que exista la ruta dada.
func (s sustitutos) empezarLaEsperaCuandoExista(t *testing.T, sesion, ruta string) {
	t.Helper()

	require.NoError(t, os.WriteFile(filepath.Join(s.transcripts, sesion+".tras"), []byte(ruta), 0o600))
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

// recibioTERM dice si TERM llegó en la sesión a la subshell que espera a quien
// duerme en el sustituto de claude, otro proceso de su grupo, y la subshell lo
// atendió.
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

	return argumentosAnotados(t, filepath.Join(s.comun, programa, sesion))
}

// argumentosAnotados son los que un sustituto anotó en el directorio de sus
// anotaciones, cada uno terminado en NUL, en su orden.
func argumentosAnotados(t *testing.T, anotaciones string) []string {
	t.Helper()

	contenido, err := leerFichero(filepath.Join(anotaciones, "argumentos"))
	require.NoError(t, err)
	require.NotEmpty(t, contenido)
	require.Equal(t, byte(0), contenido[len(contenido)-1], "cada argumento termina en NUL")

	return strings.Split(strings.TrimSuffix(string(contenido), "\x00"), "\x00")
}

// principioDelSustitutoDeGo es con lo que empieza el go que los tests de los
// guiones que ejecutan go test ponen delante en el PATH en lugar del de verdad,
// que construiría el paquete de evals y abriría sesiones con modelo (research
// D18 de H7.3; FR-068 de H7.3; FR-093 de H24): crea go/ en el directorio común,
// que falla si ya se ejecutó, y anota en él su directorio de trabajo, su TMPDIR
// y sus argumentos, cada uno terminado en NUL; y deja un directorio en su
// TMPDIR, como el de trabajo de go y la preparación de cada sesión, y no lo
// borra, como cuando go test muere sin llegar a hacerlo.
const principioDelSustitutoDeGo = `#!/bin/sh
set -eu
anotaciones="$KITLEGAL_SUSTITUTO_COMUN/go"
mkdir "$anotaciones"
pwd -P > "$anotaciones/directorio"
printf '%s\n' "$TMPDIR" > "$anotaciones/tmpdir"
mkdir "$TMPDIR/go-build-del-sustituto"
printf '%s\0' "$@" > "$anotaciones/argumentos"
`

// finalDelSustitutoDeGo es con lo que termina: escribe una línea en su salida
// estándar y otra en la de error, que son el registro de go test, y sale con el
// código de KITLEGAL_SUSTITUTO_CODIGO.
const finalDelSustitutoDeGo = `echo '` + registroDeGoEnSuSalida + `'
echo '` + registroDeGoEnLaDeError + `' >&2
exit "${KITLEGAL_SUSTITUTO_CODIGO:-0}"
`

// sustitutoDeGo es el go de TestGuionDelSondeo: entre su principio y su final,
// anota lo que hay en el -temporal que recibe y escribe en él salida.txt con
// salidaDelSustitutoDeGo. Nada de lo que escribe lleva comillas simples, porque
// va entre ellas en el guion.
const sustitutoDeGo = principioDelSustitutoDeGo + `temporal=
while [ $# -gt 0 ]; do
	case $1 in
	-temporal) temporal=$2; shift 2 ;;
	*) shift ;;
	esac
done
ls "$temporal" > "$anotaciones/temporal"
printf '%s' '` + salidaDelSustitutoDeGo + `' > "$temporal/` + ficheroDeLaSalidaDelSondeo + `"
` + finalDelSustitutoDeGo

// sustitutoDeGoDeLaMedida es el go de TestGuionDeLaMedida: entre su principio y
// su final, anota lo que hay en el directorio del fichero de -salida que recibe
// y, si KITLEGAL_SUSTITUTO_MEDIDA nombra un fichero, escribe en el de -salida lo
// que hay en él, como TestMedidaDelJuez cuando medirAlJuez da una medida; sin
// esa variable no lo escribe, como cuando no da ninguna.
const sustitutoDeGoDeLaMedida = principioDelSustitutoDeGo + `salida=
while [ $# -gt 0 ]; do
	case $1 in
	-salida) salida=$2; shift 2 ;;
	*) shift ;;
	esac
done
ls "${salida%/*}" > "$anotaciones/temporal"
if [ -n "${` + variableDeLaMedida + `:-}" ]; then
	cat "$` + variableDeLaMedida + `" > "$salida"
fi
` + finalDelSustitutoDeGo

// Lo que escribe el sustituto de go: la salida del sondeo, en salida.txt del
// temporal, y una línea en cada una de sus dos salidas, que son el registro de
// go test.
const (
	salidaDelSustitutoDeGo  = "Salida del sondeo que deja el sustituto de go\n- una serie: 3 de 3\n"
	registroDeGoEnSuSalida  = "go test escribe esto en su salida"
	registroDeGoEnLaDeError = "go test escribe esto en su salida de error"
)

// variableDeLaMedida es la variable con la que TestGuionDeLaMedida da al
// sustituto de go de la medida el fichero con la medida que escribe; sin ella,
// no escribe ninguna.
const variableDeLaMedida = "KITLEGAL_SUSTITUTO_MEDIDA"

// sustitutoGo es el nombre del go que sustituyen sustitutoDeGo y
// sustitutoDeGoDeLaMedida, el suyo en el PATH y en el directorio común.
const sustitutoGo = "go"

// escribirElGoDelSondeo escribe el sustituto de go del sondeo y devuelve su
// directorio, que va delante en el PATH del guion del sondeo.
func escribirElGoDelSondeo(t *testing.T) string {
	t.Helper()

	return escribirElGo(t, sustitutoDeGo)
}

// escribirElGoDeLaMedida escribe el sustituto de go de la medida y devuelve su
// directorio, que va delante en el PATH del guion de la medida.
func escribirElGoDeLaMedida(t *testing.T) string {
	t.Helper()

	return escribirElGo(t, sustitutoDeGoDeLaMedida)
}

// escribirElGo escribe ese sustituto de go, ejecutable, en un directorio
// temporal del test, a través de un os.Root, y devuelve ese directorio.
func escribirElGo(t *testing.T, guion string) string {
	t.Helper()

	bin := t.TempDir()

	raiz, err := os.OpenRoot(bin)
	require.NoError(t, err)

	defer func() { require.NoError(t, raiz.Close()) }()

	require.NoError(t, escribirEjecutable(raiz, sustitutoGo, guion))

	return bin
}

// goAnotado es lo que un sustituto de go anotó de su ejecución: sus
// argumentos, en su orden; su directorio de trabajo, el físico; su TMPDIR; y lo
// que había, una entrada por línea, en el directorio que mira —el -temporal
// del sondeo o el del fichero de -salida de la medida—.
type goAnotado struct {
	argumentos []string
	directorio string
	tmpdir     string
	temporal   string
}

// leerElGoAnotado lee lo que un sustituto de go anotó en el directorio de sus
// anotaciones, cada anotación sin su salto de línea final.
func leerElGoAnotado(t *testing.T, anotaciones string) goAnotado {
	t.Helper()

	anotado := goAnotado{argumentos: argumentosAnotados(t, anotaciones)}

	for anotacion, destino := range map[string]*string{
		"directorio": &anotado.directorio,
		"tmpdir":     &anotado.tmpdir,
		"temporal":   &anotado.temporal,
	} {
		contenido, err := leerFichero(filepath.Join(anotaciones, anotacion))
		require.NoError(t, err)

		*destino = strings.TrimSuffix(string(contenido), "\n")
	}

	return anotado
}

// codigoYSalidas ejecuta la orden, ya con su entorno, y devuelve el código con
// el que termina y lo que escribe en su salida estándar y en la de error. Que
// no llegue a ejecutarse falla el test.
func codigoYSalidas(t *testing.T, orden *exec.Cmd) (codigo int, salida, deError string) {
	t.Helper()

	var estandar, deErr strings.Builder

	orden.Stdout, orden.Stderr = &estandar, &deErr

	var terminada *exec.ExitError

	if err := orden.Run(); errors.As(err, &terminada) {
		codigo = terminada.ExitCode()
	} else {
		require.NoError(t, err)
	}

	return codigo, estandar.String(), deErr.String()
}

// entorno es lo que el sustituto de claude anotó que ve en la sesión, por
// clave: el directorio de trabajo, el kitlegal que resuelve, si ve
// ANTHROPIC_API_KEY y ANTHROPIC_AUTH_TOKEN (ve-<variable>, si o no) y cada
// variable de su lista que está definida.
func (s sustitutos) entorno(t *testing.T, sesion string) map[string]string {
	t.Helper()

	return entornoAnotado(t, filepath.Join(s.comun, sustitutoClaude, sesion))
}

// entornoAnotado es lo que un sustituto anotó en el fichero entorno del
// directorio de sus anotaciones, clave=valor en cada línea, por clave.
func entornoAnotado(t *testing.T, anotaciones string) map[string]string {
	t.Helper()

	contenido, err := leerFichero(filepath.Join(anotaciones, "entorno"))
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

// sustitutoDeClaudeDelJuez es el cuerpo del claude que los tests del votante
// ponen delante en el PATH del voto en lugar del de verdad, que abriría una
// sesión del juez con modelo (contracts/juez-y-voto.md §9 de H24; FR-093 de
// H24). El voto solo ve sus cuatro variables de entorno, así que ninguna le
// dice dónde anotar ni qué hacer: escribirElClaudeDelJuez le pone delante la
// línea que da a comun el directorio común del test, y lo gobiernan los
// ficheros de ese directorio. Crea votos/<su pid>, que falla si ya existe, y
// anota en él sus argumentos, cada uno terminado en NUL, su directorio de
// trabajo, el físico, lo que hay en él, su entorno entero, como lo da env, y su
// entrada estándar, hasta que se cierra; escribe en la salida de error lo que
// haya en salida-de-error y en la estándar lo que haya en salida, la salida
// grabada de una sesión del juez; y después:
//
//   - con hijo, deja en segundo plano un sleep de esos segundos, que hereda sus
//     dos salidas y las tiene abiertas cuando él ya no está, y anota su pid;
//   - con espera, deja la marca espera-empezada y se cambia con exec por un
//     sleep de esos segundos: quien duerme es el proceso del voto, con su pid,
//     y no un hijo que lo sobreviva;
//   - con codigo, sale con ese código.
//
// Con variableDeCalentar, termina con 0 sin hacer nada.
const sustitutoDeClaudeDelJuez = `if [ "${KITLEGAL_SUSTITUTO_CALENTAR:-}" = si ]; then exit 0; fi
set -eu
anotaciones="$comun/` + directorioDeVotos + `/$$"
mkdir "$anotaciones"
printf '%s\0' "$@" > "$anotaciones/argumentos"
pwd -P > "$anotaciones/directorio"
ls -A > "$anotaciones/en-el-directorio"
env > "$anotaciones/entorno"
cat > "$anotaciones/entrada"
if [ -f "$comun/` + salidaDeErrorDelClaudeDelJuez + `" ]; then cat "$comun/` + salidaDeErrorDelClaudeDelJuez + `" >&2; fi
if [ -f "$comun/` + salidaDelClaudeDelJuez + `" ]; then cat "$comun/` + salidaDelClaudeDelJuez + `"; fi
if [ -f "$comun/` + hijoDelClaudeDelJuez + `" ]; then
	sleep "$(cat "$comun/` + hijoDelClaudeDelJuez + `")" &
	printf '%s\n' "$!" > "$anotaciones/pid-del-hijo"
fi
if [ -f "$comun/` + esperaDelClaudeDelJuez + `" ]; then
	: > "$anotaciones/espera-empezada"
	exec sleep "$(cat "$comun/` + esperaDelClaudeDelJuez + `")"
fi
if [ -f "$comun/` + codigoDelClaudeDelJuez + `" ]; then exit "$(cat "$comun/` + codigoDelClaudeDelJuez + `")"; fi
`

// Lo que gobierna al sustituto de claude del juez, por el nombre de su fichero
// en el directorio común: la salida grabada que escribe en su salida estándar,
// lo que escribe en la de error, el código con el que sale, los segundos que
// espera sin terminar y los que dura el hijo que deja con sus salidas abiertas.
// Y directorioDeVotos, donde anota cada voto.
const (
	salidaDelClaudeDelJuez        = "salida"
	salidaDeErrorDelClaudeDelJuez = "salida-de-error"
	codigoDelClaudeDelJuez        = "codigo"
	esperaDelClaudeDelJuez        = "espera"
	hijoDelClaudeDelJuez          = "hijo"
	directorioDeVotos             = "votos"
)

// claudeDelJuez es dónde está el sustituto de claude de un test del votante: el
// directorio que va delante en el PATH del voto y el común, con los ficheros
// que lo gobiernan y, en votos/, lo que anota de cada voto.
type claudeDelJuez struct {
	bin   string
	comun string
}

// escribirElClaudeDelJuez escribe el sustituto de claude del juez, ejecutable,
// en un directorio temporal del test, a través de un os.Root, con el directorio
// común que crea para él escrito en su segunda línea, y lo ejecuta una vez con
// variableDeCalentar, como escribirSustitutos. Lo gobiernan los ficheros dados,
// por su nombre, que deja en el común antes de que nadie lo ejecute.
func escribirElClaudeDelJuez(t *testing.T, gobierno map[string]string) claudeDelJuez {
	t.Helper()

	c := claudeDelJuez{bin: t.TempDir(), comun: t.TempDir()}
	require.NoError(t, os.Mkdir(filepath.Join(c.comun, directorioDeVotos), 0o750))

	for fichero, contenido := range gobierno {
		require.NoError(t, os.WriteFile(filepath.Join(c.comun, fichero), []byte(contenido), 0o600))
	}

	raiz, err := os.OpenRoot(c.bin)
	require.NoError(t, err)

	defer func() { require.NoError(t, raiz.Close()) }()

	guion := "#!/bin/sh\ncomun=" + entreComillasSimples(c.comun) + "\n" + sustitutoDeClaudeDelJuez
	require.NoError(t, escribirEjecutable(raiz, sustitutoClaude, guion))
	calentar(t, filepath.Join(c.bin, sustitutoClaude))

	return c
}

// path es el PATH de un voto del test: el directorio del sustituto delante del
// PATH del proceso, del que el guion del voto y el sustituto toman bash, env,
// cat y las demás órdenes que usan.
func (c claudeDelJuez) path() string {
	return c.bin + string(os.PathListSeparator) + os.Getenv(variableDelPATH)
}

// votoAnotado es lo que el sustituto de claude del juez anotó de un voto: su
// pid, sus argumentos, su directorio de trabajo, el físico, lo que había en él,
// su entorno entero, por variable, y su entrada estándar.
type votoAnotado struct {
	pid            int
	argumentos     []string
	directorio     string
	enElDirectorio string
	entorno        map[string]string
	entrada        string
}

// votos son los votos que el sustituto de claude del juez anotó, por orden de
// pid, que no es el orden en que se pidieron.
func (c claudeDelJuez) votos(t *testing.T) []votoAnotado {
	t.Helper()

	entradas, err := os.ReadDir(filepath.Join(c.comun, directorioDeVotos))
	require.NoError(t, err)

	votos := make([]votoAnotado, 0, len(entradas))

	for _, entrada := range entradas {
		anotaciones := filepath.Join(c.comun, directorioDeVotos, entrada.Name())

		pid, err := strconv.Atoi(entrada.Name())
		require.NoErrorf(t, err, "el sustituto anota cada voto en el directorio de su pid: %s", entrada.Name())

		votos = append(votos, votoAnotado{
			pid:            pid,
			argumentos:     argumentosAnotados(t, anotaciones),
			directorio:     strings.TrimSuffix(anotacionDelVoto(t, anotaciones, "directorio"), "\n"),
			enElDirectorio: anotacionDelVoto(t, anotaciones, "en-el-directorio"),
			entorno:        entornoAnotado(t, anotaciones),
			entrada:        anotacionDelVoto(t, anotaciones, "entrada"),
		})
	}

	slices.SortFunc(votos, func(a, b votoAnotado) int { return a.pid - b.pid })

	return votos
}

// anotacionDelVoto es lo que el sustituto de claude del juez anotó de un voto en
// el fichero de ese nombre, entero.
func anotacionDelVoto(t *testing.T, anotaciones, nombre string) string {
	t.Helper()

	contenido, err := leerFichero(filepath.Join(anotaciones, nombre))
	require.NoError(t, err)

	return string(contenido)
}

// esperando es cuántos votos tienen al sustituto de claude del juez en su
// espera, con la marca que deja al empezarla. No recibe el test porque es la
// condición de assert.Eventually, que la evalúa en otra gorrutina.
func (c claudeDelJuez) esperando() int {
	marcas, err := filepath.Glob(filepath.Join(c.comun, directorioDeVotos, "*", "espera-empezada"))
	if err != nil {
		return 0
	}

	return len(marcas)
}

// terminarLosHijos termina, con KILL, el hijo que el sustituto de claude del
// juez dejó en cada voto con sus salidas abiertas, que sobrevive al voto: el
// votante termina el proceso del voto y no a sus descendientes. Un hijo que ya
// no está no es un error.
func (c claudeDelJuez) terminarLosHijos(t *testing.T) {
	t.Helper()

	anotados, err := filepath.Glob(filepath.Join(c.comun, directorioDeVotos, "*", "pid-del-hijo"))
	require.NoError(t, err)

	for _, anotado := range anotados {
		contenido, err := leerFichero(anotado)
		require.NoError(t, err)

		pid, err := strconv.Atoi(strings.TrimSuffix(string(contenido), "\n"))
		require.NoErrorf(t, err, "el pid del hijo es un entero en su línea: %s", anotado)

		if err := syscall.Kill(pid, syscall.SIGKILL); err != nil {
			require.ErrorIs(t, err, syscall.ESRCH)
		}
	}
}

// procesoTerminado dice si ya no hay ningún proceso con ese pid: el sistema
// responde ESRCH a una señal 0.
func procesoTerminado(pid int) bool {
	return errors.Is(syscall.Kill(pid, 0), syscall.ESRCH)
}

// sustitutoDeGh es el gh que TestOrdenesDeLaTanda y TestOrdenesDeLaPlataforma
// ponen en el PATH de las órdenes en lugar del de verdad, que consultaría la
// API de GitHub. Solo usa órdenes internas de sh, porque es el único programa
// de ese PATH:
//   - con variableDeGhQueFalla, escribe en sus dos salidas y termina con 4, como
//     gh ante un error de la API;
//   - con variableDeLaRespuestaDeGh, que nombra un directorio, anota en él sus
//     argumentos, uno por línea, en argumentosDeGh, y escribe en su salida
//     estándar, línea a línea, lo que el test dejó en respuestaDeGh: la
//     respuesta de gh a la orden;
//   - sin ninguna de las dos, escribe sus argumentos, uno por línea.
//
// Y termina con 0.
const sustitutoDeGh = `#!/bin/sh
if [ -n "$GH_SUSTITUTO_FALLA" ]; then
	printf 'lo escrito antes del error\n'
	printf 'HTTP 403: Resource not accessible by integration\n' >&2
	exit 4
fi
if [ -n "$GH_SUSTITUTO_RESPUESTA" ]; then
	printf '%s\n' "$@" > "$GH_SUSTITUTO_RESPUESTA/` + argumentosDeGh + `"
	while IFS= read -r linea; do
		printf '%s\n' "$linea"
	done < "$GH_SUSTITUTO_RESPUESTA/` + respuestaDeGh + `"
	exit 0
fi
for argumento in "$@"; do
	printf '%s\n' "$argumento"
done
`

// Las variables que gobiernan al sustituto de gh y, en el directorio de la
// segunda, los ficheros de sus argumentos y de su respuesta.
const (
	variableDeGhQueFalla      = "GH_SUSTITUTO_FALLA"
	variableDeLaRespuestaDeGh = "GH_SUSTITUTO_RESPUESTA"
	argumentosDeGh            = "argumentos"
	respuestaDeGh             = "respuesta"
)

// escribirElGh escribe el sustituto de gh, ejecutable, en un directorio
// temporal del test, a través de un os.Root, y devuelve ese directorio, que el
// test pone como único PATH de las órdenes.
func escribirElGh(t *testing.T) string {
	t.Helper()

	bin := t.TempDir()

	raiz, err := os.OpenRoot(bin)
	require.NoError(t, err)

	defer func() { require.NoError(t, raiz.Close()) }()

	require.NoError(t, escribirEjecutable(raiz, "gh", sustitutoDeGh))

	return bin
}

// guionDeLasEvals es scripts/evals.sh, relativo al directorio de este paquete,
// que es donde go test ejecuta los tests.
const guionDeLasEvals = "../../scripts/evals.sh"

// variableDeLaDecisionDeEvals es la variable con la que scripts/evals.sh recibe
// el fichero de la decisión de su trabajo, y faltaLaDecisionDeEvals, la línea
// con la que termina sin ella (contracts/job-de-evals.md §3 de H30; research D18
// de H30).
const (
	variableDeLaDecisionDeEvals = "DECISION_DE_EVALS"
	faltaLaDecisionDeEvals      = "evals: falta " + variableDeLaDecisionDeEvals + "\n"
)

// lasDemasDelGuionDeLasEvals son las variables obligatorias de scripts/evals.sh
// además de la decisión, con valores que pasan sus comprobaciones: las del plan
// de sesiones, la concurrencia y el commit evaluado.
var lasDemasDelGuionDeLasEvals = []string{
	"MODELO_DE_EVALS=claude-sonnet-5-5",
	"MODELOS_INFORMATIVOS_DE_EVALS=claude-haiku-4-5-20251001",
	"REPETICIONES_DE_EVALS=3",
	"UMBRAL_DE_EVALS=2",
	variableDeLaConcurrencia + "=2",
	variableDelCommitEvaluado + "=0123456789abcdef0123456789abcdef01234567",
}

// delGuionDeLasEvals son las variables que lee scripts/evals.sh: el test las
// quita del entorno del proceso, y el guion solo ve las que le da cada caso.
var delGuionDeLasEvals = []string{
	variableDeLaDecisionDeEvals, "MODELO_DE_EVALS", "MODELOS_INFORMATIVOS_DE_EVALS", "REPETICIONES_DE_EVALS",
	"UMBRAL_DE_EVALS", variableDeLaConcurrencia, variableDelCommitEvaluado, variableDelModeloDelJuez,
	variableDeLaVersionDelJuez, variableDelClaudeDelJuez, "OBJETIVO_DE_DURACION_DE_EVALS", "PRUEBA_DE_RED",
	"GITHUB_STEP_SUMMARY", "RUNNER_TEMP", "TMPDIR",
}

// casoDelGuionSinLaDecision es un caso de TestGuionDeLasEvalsSinLaDecision: el
// valor de DECISION_DE_EVALS, que no está si decision da falso; si el guion
// recibe además las demás variables obligatorias; y lo que dice en su salida de
// error.
type casoDelGuionSinLaDecision struct {
	nombre      string
	decision    func(t *testing.T) (string, bool)
	conLasDemas bool
	deError     string
}

// TestGuionDeLasEvalsSinLaDecision fija que scripts/evals.sh exige el fichero
// de la decisión de su trabajo (contracts/job-de-evals.md §3 de H30; research
// D18 de H30; FR-012): sin DECISION_DE_EVALS, con ella vacía o con una ruta que
// no es la de un fichero legible —una en la que no hay nada, la de un
// directorio o la de un fichero sin permiso de lectura—, termina con 1 y la
// línea «evals: falta DECISION_DE_EVALS», con la salida estándar vacía, sin que
// llegue ninguna sesión al claude sustituto, que va delante en el PATH, y sin
// dejar nada en su RUNNER_TEMP. Lo comprueba tras su argumento y antes de nada
// más que dependa del entorno: también sin las demás variables obligatorias,
// cuya falta diría otra cosa. Y con un fichero legible sigue, y la primera
// comprobación que no pasa es otra: sin las demás variables, la de
// MODELO_DE_EVALS.
func TestGuionDeLasEvalsSinLaDecision(t *testing.T) {
	t.Parallel()

	guion, err := filepath.Abs(guionDeLasEvals)
	require.NoError(t, err)

	for _, caso := range casosDelGuionSinLaDecision() {
		t.Run(caso.nombre, func(t *testing.T) {
			t.Parallel()

			s := escribirSustitutos(t)
			temporal := t.TempDir()

			variables := []string{"RUNNER_TEMP=" + temporal, "TMPDIR=" + temporal}
			if caso.conLasDemas {
				variables = append(variables, lasDemasDelGuionDeLasEvals...)
			}

			if decision, esta := caso.decision(t); esta {
				variables = append(variables, variableDeLaDecisionDeEvals+"="+decision)
			}

			orden := exec.CommandContext(t.Context(), guion, skillDeLegalCore)
			orden.Env = sobreLaBase(s.base(), variables, delGuionDeLasEvals...)

			codigo, salida, deError := codigoYSalidas(t, orden)

			require.Equalf(t, 1, codigo, "el código del guion, con esta salida de error:\n%s", deError)
			require.Empty(t, salida, "la salida estándar del guion")
			require.Equal(t, caso.deError, deError, "la salida de error del guion")

			sesiones, err := os.ReadDir(filepath.Join(s.comun, sustitutoClaude))
			require.NoError(t, err)
			require.Empty(t, sesiones, "ninguna sesión llega al claude sustituto")

			escrito, err := os.ReadDir(temporal)
			require.NoError(t, err)
			require.Empty(t, escrito, "el guion no deja nada en su RUNNER_TEMP")
		})
	}
}

// casosDelGuionSinLaDecision son los casos de TestGuionDeLasEvalsSinLaDecision:
// los de una decisión que falta o no es un fichero legible, con las demás
// variables y sin ellas, y el de un fichero legible, que pasa a la comprobación
// siguiente.
func casosDelGuionSinLaDecision() []casoDelGuionSinLaDecision {
	sinLaDecision := func(*testing.T) (string, bool) { return "", false }

	return []casoDelGuionSinLaDecision{
		{nombre: "sin-la-decision", decision: sinLaDecision, conLasDemas: true, deError: faltaLaDecisionDeEvals},
		{nombre: "sin-la-decision-ni-las-demas", decision: sinLaDecision, deError: faltaLaDecisionDeEvals},
		{
			nombre:      "con-la-decision-vacia",
			decision:    func(*testing.T) (string, bool) { return "", true },
			conLasDemas: true,
			deError:     faltaLaDecisionDeEvals,
		},
		{
			nombre: "con-una-decision-que-no-esta",
			decision: func(t *testing.T) (string, bool) {
				t.Helper()

				return filepath.Join(t.TempDir(), "decision-de-evals.json"), true
			},
			conLasDemas: true,
			deError:     faltaLaDecisionDeEvals,
		},
		{
			nombre: "con-una-decision-que-es-un-directorio",
			decision: func(t *testing.T) (string, bool) {
				t.Helper()

				return t.TempDir(), true
			},
			conLasDemas: true,
			deError:     faltaLaDecisionDeEvals,
		},
		{
			nombre:      "con-una-decision-que-no-se-puede-leer",
			decision:    decisionQueNoSePuedeLeer,
			conLasDemas: true,
			deError:     faltaLaDecisionDeEvals,
		},
		{
			nombre:   "con-una-decision-que-no-se-puede-leer-ni-las-demas",
			decision: decisionQueNoSePuedeLeer,
			deError:  faltaLaDecisionDeEvals,
		},
		{
			nombre:   "con-una-decision-legible-sigue",
			decision: decisionLegible,
			deError:  "evals: falta MODELO_DE_EVALS\n",
		},
	}
}

// decisionLegible es la ruta de un fichero de la decisión que se puede leer,
// en un directorio temporal del test.
func decisionLegible(t *testing.T) (string, bool) {
	t.Helper()

	ruta := filepath.Join(t.TempDir(), "decision-de-evals.json")
	require.NoError(t, os.WriteFile(ruta, []byte("{}\n"), 0o600))

	return ruta, true
}

// decisionQueNoSePuedeLeer es la ruta de un fichero de la decisión sin permiso
// de lectura, en un directorio temporal del test; que no se puede abrir para
// leer es una premisa, que con permisos de root no se cumple.
func decisionQueNoSePuedeLeer(t *testing.T) (string, bool) {
	t.Helper()

	const nombre = "decision-de-evals.json"

	dir := t.TempDir()

	raiz, err := os.OpenRoot(dir)
	require.NoError(t, err)

	defer func() { require.NoError(t, raiz.Close()) }()

	require.NoError(t, raiz.WriteFile(nombre, []byte("{}\n"), 0o200))

	fichero, err := raiz.Open(nombre)
	if err == nil {
		require.NoError(t, fichero.Close())
	}

	require.ErrorIs(t, err, fs.ErrPermission, "premisa: el fichero de la decisi\xc3\xb3n no se puede leer")

	return filepath.Join(dir, nombre), true
}

// casoDelGuionConLoDelJuez es un caso de TestGuionDeLasEvalsConLoDelJuez: lo
// del juez que el guion recibe, nombre=valor, y lo que dice en su salida de
// error.
type casoDelGuionConLoDelJuez struct {
	nombre  string
	delJuez func(t *testing.T) []string
	deError string
}

// TestGuionDeLasEvalsConLoDelJuez fija que scripts/evals.sh pide lo del juez en
// todo trabajo, y no solo con la carpeta juez en las evals de la skill
// (contracts/juez-de-las-descripciones.md §3 de H31; FR-026 de H31): con la
// skill legal-core, que no la tiene, la decisión legible y las demás variables
// obligatorias, termina con 1 y la línea de la primera de las tres que falta
// —MODELO_DEL_JUEZ, VERSION_DE_CLAUDE_CODE_DEL_JUEZ o CLAUDE_DEL_JUEZ, que falta
// también si su ruta no es la de un fichero ejecutable—, con la salida estándar
// vacía, sin que llegue ninguna sesión al claude sustituto y sin dejar nada en
// su RUNNER_TEMP. Así las tres banderas del juez llegan a TestEjecucionDelJob en
// todo trabajo, que es donde vota la descripción de su skill si la decisión lo
// pide.
func TestGuionDeLasEvalsConLoDelJuez(t *testing.T) {
	t.Parallel()

	require.NoDirExists(t, filepath.Join(raizDelRepositorio, directorioDeLasEvals, skillDeLegalCore, carpetaDelJuez),
		"premisa: las evals de legal-core no tienen la carpeta juez")

	guion, err := filepath.Abs(guionDeLasEvals)
	require.NoError(t, err)

	for _, caso := range casosDelGuionConLoDelJuez() {
		t.Run(caso.nombre, func(t *testing.T) {
			t.Parallel()

			s := escribirSustitutos(t)
			temporal := t.TempDir()
			decision, _ := decisionLegible(t)

			variables := slices.Concat([]string{
				"RUNNER_TEMP=" + temporal, "TMPDIR=" + temporal, variableDeLaDecisionDeEvals + "=" + decision,
			}, lasDemasDelGuionDeLasEvals, caso.delJuez(t))

			orden := exec.CommandContext(t.Context(), guion, skillDeLegalCore)
			orden.Env = sobreLaBase(s.base(), variables, delGuionDeLasEvals...)

			codigo, salida, deError := codigoYSalidas(t, orden)

			require.Equalf(t, 1, codigo, "el código del guion, con esta salida de error:\n%s", deError)
			require.Empty(t, salida, "la salida estándar del guion")
			require.Equal(t, caso.deError, deError, "la salida de error del guion")

			sesiones, err := os.ReadDir(filepath.Join(s.comun, sustitutoClaude))
			require.NoError(t, err)
			require.Empty(t, sesiones, "ninguna sesión llega al claude sustituto")

			escrito, err := os.ReadDir(temporal)
			require.NoError(t, err)
			require.Empty(t, escrito, "el guion no deja nada en su RUNNER_TEMP")
		})
	}
}

// casosDelGuionConLoDelJuez son los casos de TestGuionDeLasEvalsConLoDelJuez:
// sin nada del juez, con el modelo, con el modelo y la versión, y con los tres
// y un claude del juez que no es ejecutable.
func casosDelGuionConLoDelJuez() []casoDelGuionConLoDelJuez {
	const (
		modelo  = variableDelModeloDelJuez + "=claude-opus-5-5"
		version = variableDeLaVersionDelJuez + "=2.1.289"
	)

	falta := func(variable string) string { return "evals: falta " + variable + "\n" }

	return []casoDelGuionConLoDelJuez{
		{
			nombre:  "sin-lo-del-juez",
			delJuez: func(*testing.T) []string { return nil },
			deError: falta(variableDelModeloDelJuez),
		},
		{
			nombre:  "con-el-modelo",
			delJuez: func(*testing.T) []string { return []string{modelo} },
			deError: falta(variableDeLaVersionDelJuez),
		},
		{
			nombre:  "con-el-modelo-y-la-version",
			delJuez: func(*testing.T) []string { return []string{modelo, version} },
			deError: falta(variableDelClaudeDelJuez),
		},
		{
			nombre: "con-un-claude-del-juez-que-no-es-ejecutable",
			delJuez: func(t *testing.T) []string {
				t.Helper()

				claude := filepath.Join(t.TempDir(), "claude")
				require.NoError(t, os.WriteFile(claude, []byte("#!/bin/sh\nexit 0\n"), 0o600))

				return []string{modelo, version, variableDelClaudeDelJuez + "=" + claude}
			},
			deError: falta(variableDelClaudeDelJuez),
		},
	}
}
