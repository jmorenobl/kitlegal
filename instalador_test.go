package kitlegal_test

import (
	"context"
	"os"
	"os/exec"
	"regexp"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// variableSinLlavesAntesDeNoASCII casa una variable de shell escrita sin
// llaves, $nombre, seguida de un carácter que no es ASCII. El /bin/sh de macOS,
// bash 3.2, en un locale UTF-8 lee los bytes de ese carácter como parte del
// nombre y, con set -u, aborta por una variable sin definir: «$x»» es la
// variable «x\xc2», no x seguida de una comilla.
var variableSinLlavesAntesDeNoASCII = regexp.MustCompile(`\$[A-Za-z_][A-Za-z0-9_]*[^\x00-\x7f]`)

// localeUTF8 es el locale con el que se ejecutan los rechazos: existe en macOS
// y en toda glibc desde la 2.35 sin generarlo.
const localeUTF8 = "C.UTF-8"

// preludioDelInstalador va delante de install.sh en la entrada estándar de sh.
// Sustituye uname y curl por funciones, que ganan a las órdenes del PATH:
// uname da lo que digan FALSO_S y FALSO_M, y curl deja constancia de que se le
// llamó en la salida de error y falla, de modo que ningún caso descarga nada.
const preludioDelInstalador = `uname() {
	case $1 in
	-s) printf '%s\n' "$FALSO_S" ;;
	-m) printf '%s\n' "$FALSO_M" ;;
	esac
}
curl() {
	printf 'curl llamado\n' >&2
	return 1
}
`

// unameQueDejaConstancia sustituye, detrás de preludioDelInstalador, el uname
// falso por uno que deja constancia de que se le llamó: es lo primero que hace
// el guion fuera de sus funciones y de la guarda.
const unameQueDejaConstancia = `uname() {
	printf 'uname llamado\n' >&2
	return 1
}
`

// interpreteDelInstalador es un sh con el que se ejecuta install.sh leído de su
// entrada estándar, como en «curl … | sh»; la orden va escrita entera con
// constantes.
type interpreteDelInstalador struct {
	nombre string
	orden  func(ctx context.Context) *exec.Cmd
}

// interpretesDelInstalador son sh, que es el de «curl … | sh» —bash 3.2 en
// macOS, dash en Debian y Ubuntu—, y bash, si está.
var interpretesDelInstalador = []interpreteDelInstalador{
	{nombre: "sh", orden: func(ctx context.Context) *exec.Cmd { return exec.CommandContext(ctx, "sh", "-s") }},
	{nombre: "bash", orden: func(ctx context.Context) *exec.Cmd { return exec.CommandContext(ctx, "bash", "-s") }},
}

// TestInstaladorEnUTF8 fija que los rechazos de install.sh anteriores a
// descargar nada dan, también en un locale UTF-8, la única línea con el prefijo
// «install.sh: » de contracts/release.md §7 y código 1 (FR-100, FR-101,
// FR-103): ninguna variable del guion va sin llaves delante de un carácter que
// no es ASCII, y los rechazos de dos argumentos, de un sistema o una
// arquitectura que no se admiten y de un HOME relativo sin
// KITLEGAL_INSTALL_DIR, ejecutados con sh y con bash en C.UTF-8, dan esa línea
// y nada más. Los guiones instalador- los ejercen sin locale, con el guion
// entero y contra el origen local.
func TestInstaladorEnUTF8(t *testing.T) {
	t.Parallel()

	guion := leerElInstalador(t)

	t.Run("ninguna variable sin llaves delante de un carácter que no es ASCII", func(t *testing.T) {
		t.Parallel()

		assert.True(t, variableSinLlavesAntesDeNoASCII.MatchString("«$x»"), "control: la forma que falla")
		assert.False(t, variableSinLlavesAntesDeNoASCII.MatchString("«${x}» «$1» $x «"), "control: las que no")

		for numero, linea := range strings.Split(guion, "\n") {
			if variable := variableSinLlavesAntesDeNoASCII.FindString(linea); variable != "" {
				t.Errorf("install.sh:%d: %q va sin llaves delante de un carácter que no es ASCII: "+
					"escríbela ${…}", numero+1, variable)
			}
		}
	})

	casos := []struct {
		nombre     string
		argumentos string
		entorno    map[string]string
		nombra     string
	}{
		{nombre: "dos argumentos", argumentos: "0.1.0 0.2.0", nombra: "«0.1.0» «0.2.0»"},
		{nombre: "un sistema que no se admite", entorno: map[string]string{"FALSO_S": "FreeBSD"}, nombra: "«FreeBSD»"},
		{nombre: "una arquitectura que no se admite", entorno: map[string]string{"FALSO_M": "riscv64"}, nombra: "«riscv64»"},
		{
			nombre:  "un HOME relativo sin KITLEGAL_INSTALL_DIR",
			entorno: map[string]string{"HOME": "relativo", "KITLEGAL_INSTALL_DIR": ""},
			nombra:  "HOME («relativo»)",
		},
	}

	for _, interprete := range interpretesDelInstalador {
		if _, err := exec.LookPath(interprete.nombre); err != nil {
			require.NotEqual(t, "sh", interprete.nombre, "sin sh no se puede ejecutar install.sh: %v", err)

			continue
		}

		for _, caso := range casos {
			t.Run(interprete.nombre+"/"+caso.nombre, func(t *testing.T) {
				t.Parallel()

				entorno := entornoDelInstalador(t, caso.entorno)
				entrada := preludioDelInstalador + "set -- " + caso.argumentos + "\n" + guion

				codigo, errores := ejecutarElInstalador(t, interprete, entorno, entrada)

				assert.Equal(t, 1, codigo, errores)
				assert.NotContains(t, errores, "curl llamado", "ningún rechazo descarga nada")

				lineas := strings.Split(strings.TrimSuffix(errores, "\n"), "\n")
				require.Len(t, lineas, 1, "una sola línea en la salida de error: %q", errores)
				assert.True(t, strings.HasPrefix(lineas[0], "install.sh: "), lineas[0])
				assert.Contains(t, lineas[0], caso.nombra)
			})
		}
	}
}

// TestInstaladorCortado fija que un install.sh que llega cortado por
// «curl … | sh» no ejecuta nada a medias ni sale con 0 (FR-100, FR-106): cortado
// al final de cada línea y en medio de cada una, desde la guarda de la segunda
// línea hasta antes de la llamada a main, sh sale con un código distinto de 0 y
// sin llamar a uname ni a curl, que es lo primero que hace el guion.
func TestInstaladorCortado(t *testing.T) {
	t.Parallel()

	guion := leerElInstalador(t)
	lineas := strings.SplitAfter(guion, "\n")
	require.Greater(t, len(lineas), 3, "install.sh tiene más líneas que la guarda y la llamada")

	// Los cortes van desde el final de la guarda, la segunda línea, hasta el
	// último byte antes de que la llamada a main llegue entera.
	desde := len(lineas[0]) + len(lineas[1])
	llamada := strings.LastIndex(guion, "{ main \"$@\"; }")
	require.Positive(t, llamada, "install.sh termina llamando a main dentro de un grupo")

	var cortes []int

	for inicio := desde; inicio < len(guion); {
		fin := inicio + strings.IndexByte(guion[inicio:], '\n') + 1
		cortes = append(cortes, inicio+(fin-inicio)/2, fin)
		inicio = fin
	}

	entorno := entornoDelInstalador(t, nil)

	for _, corte := range cortes {
		if corte > llamada+len("{ main \"$@\"; ") {
			break
		}

		codigo, errores := ejecutarElInstalador(t, interpretesDelInstalador[0], entorno,
			preludioDelInstalador+unameQueDejaConstancia+guion[:corte])

		assert.NotEqual(t, 0, codigo, "cortado en el byte %d sale con 0: %q", corte, errores)
		assert.NotContains(t, errores, "uname llamado", "cortado en el byte %d ejecuta parte del guion", corte)
		assert.NotContains(t, errores, "curl llamado", "cortado en el byte %d ejecuta parte del guion", corte)
	}
}

// leerElInstalador es scripts/install.sh entero.
func leerElInstalador(t *testing.T) string {
	t.Helper()

	guion, err := os.ReadFile(instaladorEnLaRelease)
	require.NoError(t, err)

	return string(guion)
}

// entornoDelInstalador es el entorno de una ejecución de install.sh: el locale
// UTF-8, el PATH del proceso, un sistema y una arquitectura que se admiten, un
// HOME y un KITLEGAL_INSTALL_DIR en temporales y un origen que no existe, con
// lo que el caso cambie. Nada más del entorno de quien ejecuta los tests.
func entornoDelInstalador(t *testing.T, cambios map[string]string) []string {
	t.Helper()

	variables := map[string]string{
		"LC_ALL":               localeUTF8,
		"PATH":                 os.Getenv("PATH"),
		"FALSO_S":              "Linux",
		"FALSO_M":              "x86_64",
		"HOME":                 t.TempDir(),
		"KITLEGAL_INSTALL_DIR": t.TempDir(),
		"KITLEGAL_INSTALL_URL": "file:///no-existe",
	}

	for variable, valor := range cambios {
		variables[variable] = valor
	}

	entorno := make([]string, 0, len(variables))
	for variable, valor := range variables {
		entorno = append(entorno, variable+"="+valor)
	}

	return entorno
}

// ejecutarElInstalador ejecuta con el intérprete lo que lee de su entrada
// estándar, con ese entorno, y devuelve el código de salida y la salida de
// error.
func ejecutarElInstalador(t *testing.T, interprete interpreteDelInstalador, entorno []string, entrada string) (int, string) {
	t.Helper()

	orden := interprete.orden(t.Context())
	orden.Env = entorno
	orden.Stdin = strings.NewReader(entrada)

	var errores strings.Builder

	orden.Stderr = &errores

	if err := orden.Run(); err != nil {
		var terminado *exec.ExitError
		require.ErrorAs(t, err, &terminado, "%s no se pudo ejecutar", interprete.nombre)
	}

	return orden.ProcessState.ExitCode(), errores.String()
}
