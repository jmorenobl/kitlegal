package evals

import (
	"context"
	"encoding/json/v2"
	"io/fs"
	"os"
	"path/filepath"
	"slices"
	"strconv"
	"strings"
	"syscall"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/jmorenobl/kitlegal/internal/cache"
)

// guionDeLaSesion es scripts/evals-sesion.sh, relativo al directorio de este
// paquete, que es donde go test ejecuta los tests.
const guionDeLaSesion = "../../scripts/evals-sesion.sh"

// El tope y el margen de los tests: los de verdad son 240 s y 10 s
// (contracts/ejecucion-del-job.md §3.3 de H7.3), y el repartidor los recibe.
// topeSinCorte es el de los tests que no miran el tope: ninguna sesión suya se
// acerca a él, así que no depende de la carga de la máquina.
const (
	topeDeLosTests   = time.Second
	margenDeLosTests = time.Second
	topeSinCorte     = time.Minute
)

// transcriptTerminado es el transcript stream-json de una sesión que termina con
// su respuesta, con el modelo de modeloDeLaSesion.
const transcriptTerminado = `{"type":"system","subtype":"init","model":"` + modeloDeLaSesion + `",` +
	`"claude_code_version":"2.1.270"}` + "\n" +
	`{"type":"result","subtype":"success","is_error":false,"result":"Respuesta del sustituto."}` + "\n"

// Las entradas del directorio de skills instaladas de los tests: una carpeta y
// un enlace a otra, como las deja make install.
const (
	skillEnCarpeta = "boe-legislacion"
	skillEnlazada  = "legal-core"
)

// TestTopeDeLaSesion fija el tope de la sesión, en Go y sin timeout
// (contracts/ejecucion-del-job.md §3.3 de H7.3; research D9 de H7.3; FR-031):
// con un tope de 1 s y un margen de 1 s, el sustituto de claude que duerme 5 s
// deja 124, porque bastó TERM, que llega a todo su grupo de procesos; el que
// además ignora TERM deja 137, porque hizo falta KILL; y el que termina antes
// deja su código, también el de una señal que no es del tope, como lo daría la
// shell. Los dos primeros no cumplen su espera: el tope los corta antes, y no
// se limita a esperarlos. LeerSesion lee cada código.
func TestTopeDeLaSesion(t *testing.T) {
	t.Parallel()

	casos := []struct {
		nombre    string
		variables []string
		codigo    int

		// cortada dice si el tope corta la sesión antes de que el sustituto
		// cumpla su espera.
		cortada bool

		// term dice si el sustituto recibe TERM y lo atiende.
		term bool

		// minimo es lo que tarda la sesión como poco: el tope si hace falta
		// TERM, y el tope más el margen si hace falta KILL.
		minimo time.Duration
	}{
		{
			nombre:    "basta-term",
			variables: []string{variableDeEspera + "=5"},
			codigo:    codigoDelTope,
			cortada:   true,
			term:      true,
			minimo:    topeDeLosTests,
		},
		{
			nombre:    "hace-falta-kill",
			variables: []string{variableDeEspera + "=5", variableDeIgnorarTERM + "=si"},
			codigo:    codigoDeKillTrasElTope,
			cortada:   true,
			minimo:    topeDeLosTests + margenDeLosTests,
		},
		{
			nombre:    "termina-antes",
			variables: []string{variableDeCodigo + "=3"},
			codigo:    3,
		},
		{
			nombre:    "muere-por-otra-senal",
			variables: []string{variableDeSenal + "=HUP"},
			codigo:    128 + int(syscall.SIGHUP),
		},
	}

	for _, caso := range casos {
		t.Run(caso.nombre, func(t *testing.T) {
			t.Parallel()

			s := escribirSustitutos(t)
			ejecucion := sesionesDelTest(t, s, false)
			ejecucion.Entorno = s.base(caso.variables...)
			sesion := ejecucion.Plan[0]

			inicio := time.Now()
			require.NoError(t, abrirSesion(t.Context().Done(), ejecucion, sesion))
			assert.GreaterOrEqual(t, time.Since(inicio), caso.minimo)

			dir := filepath.Join(ejecucion.Sesiones, sesion.Nombre)
			require.Truef(t, s.llego(t, sesion.Nombre), "el sustituto de claude se ejecuta en la sesion; sesion.err:\n%s",
				contenidoDeLaSesion(t, dir, ficheroDeSalidaDeError))
			assert.Equal(t, !caso.cortada, s.cumplioLaEspera(t, sesion.Nombre), "el sustituto cumple su espera")
			assert.Equal(t, caso.term, s.recibioTERM(t, sesion.Nombre), "el sustituto recibe TERM")
			assert.Equal(t, strconv.Itoa(caso.codigo)+"\n", contenidoDeLaSesion(t, dir, ficheroDelCodigo))

			leida, err := LeerSesion(dir)
			require.NoError(t, err)
			assert.Equal(t, caso.codigo, leida.Codigo)
			assert.Equal(t, caso.cortada, leida.Cortada)
		})
	}
}

// TestInterrumpirLaSesion fija la interrupción de una sesión (research D9 de
// H7.3; FR-037, FR-064): con interrupcion ya cerrado al terminar la
// preparación, el guion no se ejecuta; cerrado con la sesión abierta, el
// sustituto de claude que duerme 5 s se cierra con la secuencia del tope: TERM
// a todo su grupo de procesos y, al que lo ignora, KILL; ninguno cumple su
// espera. En los dos casos el error es errSesionInterrumpida y la sesión queda
// sin codigo-de-la-sesion.
func TestInterrumpirLaSesion(t *testing.T) {
	t.Parallel()

	casos := []struct {
		nombre    string
		variables []string

		// abierta dice si interrupcion se cierra con la sesión abierta, cuando
		// el sustituto ha llegado; si no, se cierra antes de abrirla.
		abierta bool

		// term dice si el sustituto recibe TERM y lo atiende.
		term bool
	}{
		{nombre: "antes-de-abrirla"},
		{
			nombre:    "abierta",
			variables: []string{variableDeEspera + "=5"},
			abierta:   true,
			term:      true,
		},
		{
			nombre:    "abierta-e-ignora-term",
			variables: []string{variableDeEspera + "=5", variableDeIgnorarTERM + "=si"},
			abierta:   true,
		},
	}

	for _, caso := range casos {
		t.Run(caso.nombre, func(t *testing.T) {
			t.Parallel()

			s := escribirSustitutos(t)
			ejecucion := sesionesDelTest(t, s, false)
			ejecucion.Tope = topeSinCorte
			ejecucion.Entorno = s.base(caso.variables...)
			sesion := ejecucion.Plan[0]

			interrupcion := make(chan struct{})
			if !caso.abierta {
				close(interrupcion)
			}

			resultado := make(chan error, 1)

			go func() { resultado <- abrirSesion(interrupcion, ejecucion, sesion) }()

			if caso.abierta {
				empezada := s.esperaEmpezada(sesion.Nombre)
				assert.Eventually(t, func() bool {
					_, err := os.Lstat(empezada)

					return err == nil
				}, topeSinCorte, 10*time.Millisecond, "el sustituto de claude empieza su espera")
				close(interrupcion)
			}

			require.ErrorIs(t, <-resultado, errSesionInterrumpida)

			dir := filepath.Join(ejecucion.Sesiones, sesion.Nombre)
			assert.NoFileExists(t, filepath.Join(dir, ficheroDelCodigo))
			assert.Equal(t, caso.abierta, s.llego(t, sesion.Nombre), "el sustituto llega solo a la sesion abierta")
			assert.False(t, s.cumplioLaEspera(t, sesion.Nombre), "el sustituto no cumple su espera")
			assert.Equal(t, caso.term, s.recibioTERM(t, sesion.Nombre), "el sustituto recibe TERM")

			if !caso.abierta {
				assert.FileExists(t, filepath.Join(dir, ficheroDeLaPregunta), "la sesion se ha preparado")
				assert.NoFileExists(t, filepath.Join(dir, ficheroDelTranscript), "el guion no se ha ejecutado")
			}
		})
	}
}

// TestSesionQueNoSePuedePreparar fija que una sesión cuya preparación da faltas
// no se abre (research D7 de H7.3; contrato job-de-evals §3.2 de H5): con una
// eval del art. 9998 de la LPAC, que no está en ninguna grabación, el error
// nombra la sesión, la eval y la consulta que falta, el sustituto de claude no
// llega y la sesión queda sin transcript y sin codigo-de-la-sesion.
func TestSesionQueNoSePuedePreparar(t *testing.T) {
	t.Parallel()

	s := escribirSustitutos(t)
	ejecucion := sesionesDelTest(t, s, false)
	ejecucion.Tope = topeSinCorte

	const deArticulo9998 = "02-lpac-articulo-9998.yaml"
	require.NoError(t, os.WriteFile(filepath.Join(ejecucion.Evals, deArticulo9998),
		[]byte(contenidoDelArticulo9998), 0o600))

	sesion := ejecucion.Plan[0]
	err := abrirSesion(t.Context().Done(), ejecucion, sesion)

	require.Error(t, err)
	require.ErrorContains(t, err, sesion.Nombre)
	require.ErrorContains(t, err, deArticulo9998)
	require.ErrorContains(t, err, "boe articulo BOE-A-2015-10565 a9998")

	dir := filepath.Join(ejecucion.Sesiones, sesion.Nombre)
	assert.False(t, s.llego(t, sesion.Nombre), "sin preparar no se abre")
	assert.NoFileExists(t, filepath.Join(dir, ficheroDelTranscript))
	assert.NoFileExists(t, filepath.Join(dir, ficheroDelCodigo))
}

// TestAbrirUnaSesion fija lo que abre una sesión (contracts/ejecucion-del-job.md
// §2, §3.2 y §4 de H7.3; research D8 y D10 de H7.3; FR-031, FR-032) con los
// sustitutos de claude y strace y la sesión de la prueba de red, cuya pregunta
// lleva varias líneas: su directorio, con trabajo/, cache/, traza/, tmp/ y
// claude/ en 0700 y un enlace absoluto por skill instalada en claude/skills/; la
// orden de la sesión, carácter a carácter, con la de strace delante solo con
// Traza; la traza en traza/, que LeerTrazas lee; el entorno de §4 dentro del
// directorio de la sesión, sobre una base que trae las mismas variables con
// otros valores, y sin KITLEGAL_EVALS_TRAZA sin Traza aunque la base la traiga;
// y la sesión, terminada con 0, que LeerSesion lee.
func TestAbrirUnaSesion(t *testing.T) {
	t.Parallel()

	for _, traza := range []bool{true, false} {
		t.Run("traza-"+strconv.FormatBool(traza), func(t *testing.T) {
			t.Parallel()

			s := escribirSustitutos(t)
			ejecucion := sesionesDelTest(t, s, traza)
			ejecucion.Tope = topeSinCorte
			ejecucion.Entorno = s.base(
				cache.VariableDirectorio+"=/otra/cache", "HTTP_PROXY=http://otro:3128", "no_proxy=*",
				"CLAUDE_CONFIG_DIR=/otro/claude", "TMPDIR=/otro/tmp", "KITLEGAL_EVALS_TRAZA=si", "HOME=/otro/home")
			sesion := ejecucion.Plan[len(ejecucion.Plan)-1]
			require.True(t, sesion.PruebaDeRed, "la ultima sesion del plan es la de la prueba de red")

			s.escribirTranscript(t, sesion.Nombre, transcriptTerminado)

			require.NoError(t, abrirSesion(t.Context().Done(), ejecucion, sesion))

			dir := filepath.Join(ejecucion.Sesiones, sesion.Nombre)
			require.Truef(t, s.llego(t, sesion.Nombre), "el sustituto de claude se ejecuta en la sesion; sesion.err:\n%s",
				contenidoDeLaSesion(t, dir, ficheroDeSalidaDeError))
			exigirLosDirectoriosDeLaSesion(t, dir, ejecucion.Skills)
			exigirLaOrdenDeLaSesion(t, s, dir, sesion.Nombre, traza)
			exigirLaTrazaDeLaSesion(t, dir, traza)
			exigirElEntornoDeLaSesion(t, s, dir, sesion.Nombre, traza)

			assert.Equal(t, "0\n", contenidoDeLaSesion(t, dir, ficheroDelCodigo))

			leida, err := LeerSesion(dir)
			require.NoError(t, err)
			assert.True(t, leida.Terminada)
			assert.Equal(t, modeloDeLaSesion, leida.Modelo)
			assert.Equal(t, "Respuesta del sustituto.", leida.Respuesta)
		})
	}
}

// exigirLosDirectoriosDeLaSesion exige trabajo/, cache/, traza/, tmp/ y claude/
// en el directorio de la sesión, en 0700, y, en claude/skills/, un enlace
// absoluto a cada entrada de las skills instaladas y nada más.
func exigirLosDirectoriosDeLaSesion(t *testing.T, dir, skills string) {
	t.Helper()

	for _, subdirectorio := range []string{"trabajo", "cache", "traza", "tmp", "claude"} {
		estado, err := os.Stat(filepath.Join(dir, subdirectorio))
		require.NoError(t, err)
		assert.Truef(t, estado.IsDir(), "%s es un directorio", subdirectorio)
		assert.Equalf(t, os.FileMode(0o700), estado.Mode().Perm(), "permisos de %s", subdirectorio)
	}

	enlaces, err := os.ReadDir(filepath.Join(dir, "claude", "skills"))
	require.NoError(t, err)

	nombres := make([]string, 0, len(enlaces))
	for _, enlace := range enlaces {
		nombres = append(nombres, enlace.Name())

		destino, err := os.Readlink(filepath.Join(dir, "claude", "skills", enlace.Name()))
		require.NoError(t, err)
		assert.Equal(t, filepath.Join(skills, enlace.Name()), destino)
		assert.Truef(t, filepath.IsAbs(destino), "el enlace de %s es absoluto", enlace.Name())
	}

	assert.ElementsMatch(t, []string{skillEnCarpeta, skillEnlazada}, nombres)
}

// exigirLaOrdenDeLaSesion exige que el sustituto de claude reciba la orden de la
// sesión carácter a carácter, con la pregunta entera de pregunta.txt sin el
// salto de línea final, y que, con traza, el de strace la reciba detrás de la
// suya, con su traza en ../traza/t; sin traza, strace no se ejecuta
// (contracts/ejecucion-del-job.md §2 de H7.3).
func exigirLaOrdenDeLaSesion(t *testing.T, s sustitutos, dir, sesion string, traza bool) {
	t.Helper()

	pregunta := preguntaDeLaSesion(t, dir)
	require.Contains(t, pregunta, "\n\n", "la pregunta de la prueba de red lleva varias lineas")

	exigirLosArgumentosDeLaSesion(t, s, sesion, ordenDeLaSesion(pregunta), traza)
}

// preguntaDeLaSesion es la pregunta de pregunta.txt del directorio de la
// sesión, sin su salto de línea final, que es como la recibe la orden.
func preguntaDeLaSesion(t *testing.T, dir string) string {
	t.Helper()

	return strings.TrimSuffix(contenidoDeLaSesion(t, dir, ficheroDeLaPregunta), "\n")
}

// ordenDeLaSesion son los argumentos de la orden de una sesión sin servidor.json
// con la pregunta dada y modeloDeLaSesion: la de siempre, carácter a carácter
// (contracts/ejecucion-del-job.md §2 de H7.3).
func ordenDeLaSesion(pregunta string) []string {
	return []string{
		"-p", pregunta, "--model", modeloDeLaSesion, "--output-format", "stream-json", "--verbose",
		"--max-turns", "30", "--no-session-persistence", "--setting-sources", "user",
		"--settings", `{"sandbox":{"enabled":false}}`, "--permission-mode", "bypassPermissions",
		"--disallowedTools", "WebFetch", "WebSearch",
	}
}

// conElServidor es la orden dada con lo que el guion le añade al final cuando
// la sesión tiene servidor.json (contracts/evals-en-dos-modos.md §2.3 de H21).
func conElServidor(orden []string) []string {
	return slices.Concat(orden, []string{"--mcp-config", "../servidor.json"})
}

// exigirLosArgumentosDeLaSesion exige que el sustituto de claude reciba en la
// sesión la orden dada y que, con traza, el de strace la reciba detrás de la
// suya, con su traza en ../traza/t; sin traza, strace no se ejecuta.
func exigirLosArgumentosDeLaSesion(t *testing.T, s sustitutos, sesion string, orden []string, traza bool) {
	t.Helper()

	assert.Equal(t, orden, s.argumentos(t, sustitutoClaude, sesion))

	if !traza {
		assert.NoDirExists(t, filepath.Join(s.comun, sustitutoStrace, sesion), "sin traza no hay strace")

		return
	}

	delante := []string{
		"-ff", "-e", "trace=execve,connect,clone,clone3,fork,vfork", "-s", "131072", "-o", "../traza/t", "--",
		sustitutoClaude,
	}
	assert.Equal(t, slices.Concat(delante, orden), s.argumentos(t, sustitutoStrace, sesion))
}

// exigirLaTrazaDeLaSesion exige, con traza, la del sustituto de strace en un
// fichero t.<n> de traza/, que LeerTrazas lee sin invocaciones; sin traza,
// traza/ vacío.
func exigirLaTrazaDeLaSesion(t *testing.T, dir string, traza bool) {
	t.Helper()

	directorio := filepath.Join(dir, "traza")

	ficheros, err := os.ReadDir(directorio)
	require.NoError(t, err)

	if !traza {
		assert.Empty(t, ficheros, "sin traza, traza/ queda vacio")

		return
	}

	require.Len(t, ficheros, 1)
	assert.Regexp(t, `^t\.[0-9]+$`, ficheros[0].Name())
	assert.Equal(t, trazaDelSustituto, contenidoDeLaSesion(t, directorio, ficheros[0].Name()))

	invocaciones, err := LeerTrazas(directorio, false)
	require.NoError(t, err)
	assert.Empty(t, invocaciones)
}

// exigirElEntornoDeLaSesion exige que el sustituto de claude se ejecute en
// trabajo/ y vea el entorno de contracts/ejecucion-del-job.md §4 de H7.3, con
// cada ruta dentro del directorio de la sesión y en lugar de la de la base, y
// el resto de la base sin cambios.
func exigirElEntornoDeLaSesion(t *testing.T, s sustitutos, dir, sesion string, traza bool) {
	t.Helper()

	anotado := s.entorno(t, sesion)

	trabajo, err := filepath.EvalSymlinks(filepath.Join(dir, "trabajo"))
	require.NoError(t, err)
	assert.Equal(t, trabajo, anotado[claveDelDirectorio])

	esperado := map[string]string{
		"KITLEGAL_CACHE_DIR": filepath.Join(dir, "cache"),
		"HTTP_PROXY":         "http://127.0.0.1:9",
		"HTTPS_PROXY":        "http://127.0.0.1:9",
		"http_proxy":         "http://127.0.0.1:9",
		"https_proxy":        "http://127.0.0.1:9",
		"NO_PROXY":           "api.anthropic.com",
		"no_proxy":           "api.anthropic.com",
		"CLAUDE_CODE_DISABLE_NONESSENTIAL_TRAFFIC": "1",
		"CLAUDE_CODE_SUBPROCESS_ENV_SCRUB":         "0",
		"CLAUDE_CONFIG_DIR":                        filepath.Join(dir, "claude"),
		"TMPDIR":                                   filepath.Join(dir, "tmp"),
		"CLAUDE_CODE_TMPDIR":                       filepath.Join(dir, "tmp"),
		"HOME":                                     "/otro/home",
	}
	if traza {
		esperado["KITLEGAL_EVALS_TRAZA"] = "si"
	}

	for variable, valor := range esperado {
		assert.Equalf(t, valor, anotado[variable], "la variable %s en la sesion", variable)
	}

	if !traza {
		assert.NotContains(t, anotado, "KITLEGAL_EVALS_TRAZA", "sin traza, la variable no llega a la sesion")
	}
}

// preguntaDelGuion es la pregunta de las sesiones de TestGuionDeLaSesion: con
// varias líneas, comillas y un dólar, que la orden recibe tal cual.
const preguntaDelGuion = "¿Qué dice el art. 21 de la Ley 39/2015?\n\nCon \"comillas\", 'apóstrofos' y $HOME."

// TestGuionDeLaSesion fija la orden que ejecuta scripts/evals-sesion.sh según
// haya o no servidor.json en el directorio de la sesión
// (contracts/evals-en-dos-modos.md §2.3 y §8 de H21; research V21 de H21;
// FR-040), con los sustitutos de claude y strace y una sesión creada a mano:
// sin el fichero, la orden es la de siempre, carácter a carácter; con él, la
// misma con --mcp-config ../servidor.json al final. Con traza, la de strace va
// delante de una y de otra.
func TestGuionDeLaSesion(t *testing.T) {
	t.Parallel()

	casos := []struct {
		nombre      string
		conServidor bool
		traza       bool
	}{
		{nombre: "sin-servidor"},
		{nombre: "sin-servidor-con-traza", traza: true},
		{nombre: "con-servidor", conServidor: true},
		{nombre: "con-servidor-y-traza", conServidor: true, traza: true},
	}

	guion, err := filepath.Abs(guionDeLaSesion)
	require.NoError(t, err)

	for _, caso := range casos {
		t.Run(caso.nombre, func(t *testing.T) {
			t.Parallel()

			s := escribirSustitutos(t)

			dir := filepath.Join(t.TempDir(), caso.nombre)
			require.NoError(t, crearDirectorioDeSesion(dir, skillsInstaladas(t)))

			escritos := map[string]string{
				ficheroDeLaPregunta: preguntaDelGuion + "\n",
				ficheroDelModelo:    modeloDeLaSesion + "\n",
			}
			if caso.conServidor {
				escritos[ficheroDelServidor] = "{}\n"
			}

			for fichero, contenido := range escritos {
				require.NoError(t, os.WriteFile(filepath.Join(dir, fichero), []byte(contenido), 0o600))
			}

			codigo, err := ejecutarElGuion(t.Context().Done(), guion, sesionEnMarcha{
				dir:     dir,
				entorno: entornoDeLaSesion(s.base(), dir, caso.traza),
				tope:    topeSinCorte,
				margen:  margenDeLosTests,
			})
			require.NoError(t, err)
			require.Zerof(t, codigo, "el guion termina con 0; sesion.err:\n%s",
				contenidoDeLaSesion(t, dir, ficheroDeSalidaDeError))

			orden := ordenDeLaSesion(preguntaDelGuion)
			if caso.conServidor {
				orden = conElServidor(orden)
			}

			exigirLosArgumentosDeLaSesion(t, s, caso.nombre, orden, caso.traza)
		})
	}
}

// TestSesionesPorModo fija lo que el repartidor da a la sesión de cada modo
// (contracts/evals-en-dos-modos.md §2.2 y §8 de H21; data-model §6 y §11 de H21;
// research D16 de H21; FR-041, FR-046), con el claude sustituto: la del modo
// orden, lo de siempre, con kitlegal en su PATH y sin servidor.json; la del modo
// herramienta, servidor.json con el contenido del contrato carácter a carácter y
// el PATH de la base sin el directorio del binario; y la de la eval sin binario
// ni servidor, ese mismo PATH y ningún servidor.json. Sin la ruta absoluta del
// binario no se abre ninguna sesión, y una sesión cuyo servidor.json no se puede
// escribir no se abre.
func TestSesionesPorModo(t *testing.T) {
	t.Parallel()

	t.Run("cada-modo", probarLaSesionDeCadaModo)
	t.Run("binario-sin-ruta-absoluta", probarElBinarioSinRutaAbsoluta)
	t.Run("servidor-que-no-se-escribe", probarElServidorQueNoSeEscribe)
}

// sesionDelModo es lo que el repartidor da a la sesión de un modo: su PATH, si
// en él está el binario y si tiene servidor.json.
type sesionDelModo struct {
	path        string
	conBinario  bool
	conServidor bool
}

// probarLaSesionDeCadaModo abre, con el repartidor, una sesión del modo orden,
// una del modo herramienta y una de la eval sin binario ni servidor, y exige lo
// de cada una.
func probarLaSesionDeCadaModo(t *testing.T) {
	t.Parallel()

	s := escribirSustitutos(t)
	ejecucion := sesionesEnDosModos(t, s)
	require.Equal(t, []Modo{ModoOrden, ModoHerramienta, ""}, modosDe(ejecucion.Plan),
		"premisa: el plan tiene una sesion de cada modo")

	for _, sesion := range ejecucion.Plan {
		s.escribirTranscript(t, sesion.Nombre, transcriptTerminado)
	}

	ejecutada, err := ejecutarSesiones(t.Context().Done(), ejecucion)
	require.NoError(t, err)
	require.Equal(t, nombresDe(ejecucion.Plan), ejecutada.Abiertas)

	delBinario := filepath.Dir(ejecucion.Binario)
	sinElBinario := s.bin + string(os.PathListSeparator) + os.Getenv("PATH")
	esperadas := map[Modo]sesionDelModo{
		ModoOrden:       {path: pathConElBinario(s, delBinario), conBinario: true},
		ModoHerramienta: {path: sinElBinario, conServidor: true},
		"":              {path: sinElBinario},
	}

	for _, sesion := range ejecucion.Plan {
		exigirLaSesionDelModo(t, s, ejecucion, sesion, esperadas[sesion.Modo])
	}
}

// exigirLaSesionDelModo exige que la sesión se abra con lo esperado de su modo:
// el PATH que ve el sustituto de claude y si resuelve en él el binario; la caché
// preparada y lo que es propio de cada sesión, como en todos los modos; y, según
// tenga o no servidor.json, su contenido, que repite la caché y los proxies que
// ve la sesión, y la orden que lo declara, o la orden de siempre.
func exigirLaSesionDelModo(
	t *testing.T, s sustitutos, ejecucion SesionesAEjecutar, sesion SesionPlanificada, esperada sesionDelModo,
) {
	t.Helper()

	dir := filepath.Join(ejecucion.Sesiones, sesion.Nombre)
	require.Truef(t, s.llego(t, sesion.Nombre), "el sustituto de claude se ejecuta en la sesion %s; sesion.err:\n%s",
		sesion.Nombre, contenidoDeLaSesion(t, dir, ficheroDeSalidaDeError))

	anotado := s.entorno(t, sesion.Nombre)
	assert.Equalf(t, esperada.path, anotado["PATH"], "el PATH de la sesion %s", sesion.Nombre)
	assert.Equalf(t, esperada.conBinario, anotado["kitlegal"] == ejecucion.Binario,
		"la sesion %s resuelve el binario en su PATH; resuelve %q", sesion.Nombre, anotado["kitlegal"])

	assert.FileExistsf(t, filepath.Join(dir, "cache", "cache.db"), "la cache de la sesion %s se prepara", sesion.Nombre)
	exigirLosDirectoriosPropios(t, s, dir, sesion.Nombre)

	orden := ordenDeLaSesion(preguntaDeLaSesion(t, dir))

	if !esperada.conServidor {
		assert.NoFileExists(t, filepath.Join(dir, "servidor.json"))
		exigirLosArgumentosDeLaSesion(t, s, sesion.Nombre, orden, false)

		return
	}

	escrito := contenidoDeLaSesion(t, dir, "servidor.json")
	assert.Equal(t, servidorEsperado(ejecucion.Binario, dir), escrito)
	exigirLosArgumentosDeLaSesion(t, s, sesion.Nombre, conElServidor(orden), false)

	var declarado struct {
		Servidores map[string]struct {
			Entorno map[string]string `json:"env"`
		} `json:"mcpServers"`
	}

	require.NoError(t, json.Unmarshal([]byte(escrito), &declarado))

	entorno := declarado.Servidores["kitlegal"].Entorno
	assert.Len(t, entorno, 7, "la cache y los seis proxies")

	for variable, valor := range entorno {
		assert.Equalf(t, anotado[variable], valor, "el servidor repite la variable %s de la sesion", variable)
	}
}

// servidorEsperado es el servidor.json de contracts/evals-en-dos-modos.md §2.2
// de H21 para el binario y el directorio de sesión dados, carácter a carácter y
// con su salto de línea final.
func servidorEsperado(binario, dir string) string {
	return `{"mcpServers":{"kitlegal":{"command":"` + binario + `","args":["mcp","serve"],"env":{` +
		`"KITLEGAL_CACHE_DIR":"` + dir + `/cache","HTTP_PROXY":"http://127.0.0.1:9",` +
		`"HTTPS_PROXY":"http://127.0.0.1:9","http_proxy":"http://127.0.0.1:9",` +
		`"https_proxy":"http://127.0.0.1:9","NO_PROXY":"api.anthropic.com","no_proxy":"api.anthropic.com"}}}}` + "\n"
}

// probarElBinarioSinRutaAbsoluta fija que, con una sesión que no es del modo
// orden en el plan, un Binario que no es una ruta absoluta es un error antes de
// abrir ninguna: sin él no hay directorio que quitar del PATH ni orden que
// declarar. El error lo nombra y el directorio de sesiones queda vacío.
func probarElBinarioSinRutaAbsoluta(t *testing.T) {
	t.Parallel()

	for nombre, binario := range map[string]string{"vacio": "", "relativo": "bin/kitlegal"} {
		t.Run(nombre, func(t *testing.T) {
			t.Parallel()

			s := escribirSustitutos(t)
			ejecucion := sesionesEnDosModos(t, s)
			ejecucion.Binario = binario

			_, err := ejecutarSesiones(t.Context().Done(), ejecucion)
			require.ErrorContains(t, err, "el binario de kitlegal es "+strconv.Quote(binario)+
				" y tiene que ser una ruta absoluta")

			entradas, err := os.ReadDir(ejecucion.Sesiones)
			require.NoError(t, err)
			assert.Empty(t, entradas, "no se abre ninguna sesion")
		})
	}
}

// probarElServidorQueNoSeEscribe fija que una sesión del modo herramienta cuyo
// servidor.json no se puede escribir no se abre: porque su ruta ya es un
// directorio, o porque la ruta del binario lleva octetos que no son UTF-8 y no
// cabe en un texto de JSON. El error nombra la sesión y el fichero, el
// sustituto de claude no llega y la sesión queda sin codigo-de-la-sesion.
func probarElServidorQueNoSeEscribe(t *testing.T) {
	t.Parallel()

	casos := []struct {
		nombre   string
		preparar func(t *testing.T, ejecucion *SesionesAEjecutar, dir string)

		// fragmento es lo que el error dice del fichero.
		fragmento string
	}{
		{
			nombre: "su-ruta-es-un-directorio",
			preparar: func(t *testing.T, _ *SesionesAEjecutar, dir string) {
				t.Helper()

				require.NoError(t, os.MkdirAll(filepath.Join(dir, "servidor.json"), 0o700))
			},
			fragmento: "servidor.json: is a directory",
		},
		{
			nombre: "binario-que-no-es-utf8",
			preparar: func(t *testing.T, ejecucion *SesionesAEjecutar, _ string) {
				t.Helper()

				ejecucion.Binario = "/kitlegal-\xff/kitlegal"
			},
			fragmento: "servidor.json no se puede codificar",
		},
	}

	for _, caso := range casos {
		t.Run(caso.nombre, func(t *testing.T) {
			t.Parallel()

			s := escribirSustitutos(t)
			ejecucion := sesionesEnDosModos(t, s)

			sesion := ejecucion.Plan[1]
			require.Equal(t, ModoHerramienta, sesion.Modo, "premisa: la segunda sesion del plan es del modo herramienta")

			ejecucion.Plan = []SesionPlanificada{sesion}
			dir := filepath.Join(ejecucion.Sesiones, sesion.Nombre)
			caso.preparar(t, &ejecucion, dir)

			_, err := ejecutarSesiones(t.Context().Done(), ejecucion)
			require.ErrorContains(t, err, "la sesión "+sesion.Nombre+":")
			require.ErrorContains(t, err, caso.fragmento)

			assert.False(t, s.llego(t, sesion.Nombre), "sin su servidor.json la sesion no se abre")
			assert.NoFileExists(t, filepath.Join(dir, ficheroDelCodigo))
		})
	}
}

// repeticionesDelRepartidor son las de la eval sintética en los tests del
// repartidor: con la prueba de red, un plan de ocho sesiones
// (contracts/ejecucion-del-job.md §8 de H7.3).
const repeticionesDelRepartidor = 7

// sesionesDelPlan son las sesiones del plan de los tests del repartidor.
const sesionesDelPlan = repeticionesDelRepartidor + 1

// duracionDelInforme son los segundos de duración con los que
// TestEjecutarSesionesEnParalelo escribe sus dos informes: los mismos en los dos.
const duracionDelInforme = 60

// TestEjecutarSesionesEnParalelo fija el reparto de las sesiones de un plan
// (contracts/ejecucion-del-job.md §3 y §8 de H7.3; research D7, D10 y D14 de
// H7.3; FR-030, FR-032, FR-094; SC-008; US3-1) con los sustitutos de claude y
// strace y un plan de ocho sesiones sintéticas con traza, en las que el
// sustituto de claude duerme 1 s: con Concurrencia 4, el máximo de sesiones
// abiertas a la vez que cuentan los sustitutos al llegar es como mucho 4 y al
// menos 2, y con 1, es 1. En los dos casos se abren todas en el orden del plan;
// la duración cubre al menos las esperas que la concurrencia no deja solapar y
// no pasa de lo que tarda la llamada; el directorio de trabajo,
// CLAUDE_CONFIG_DIR, TMPDIR, CLAUDE_CODE_TMPDIR y KITLEGAL_CACHE_DIR de cada
// sesión son los suyos, dentro de su directorio, y por tanto distintos de los de
// las demás; lo que escribe cada una en ellos no está en ningún directorio de
// otra; y claude/skills/ lleva un enlace por skill instalada. informe.json e
// informe.md de EscribirInforme, escritos con la misma duración, son iguales
// byte a byte con 4 y con 1.
func TestEjecutarSesionesEnParalelo(t *testing.T) {
	t.Parallel()

	conCuatro := ejecutarEnParalelo(t, 4, 2)
	conUna := ejecutarEnParalelo(t, 1, 1)

	assert.Equal(t, conUna.json, conCuatro.json, "informe.json con 4 sesiones a la vez y con 1")
	assert.Equal(t, conUna.md, conCuatro.md, "informe.md con 4 sesiones a la vez y con 1")
}

// informeEscrito son informe.json e informe.md tal como los escribe
// EscribirInforme.
type informeEscrito struct {
	json string
	md   string
}

// ejecutarEnParalelo reparte el plan de ocho sesiones con traza y la
// concurrencia dada, exige lo que TestEjecutarSesionesEnParalelo fija de cada
// reparto, con el mínimo dado del máximo de sesiones abiertas a la vez, y
// devuelve el informe de sus sesiones.
func ejecutarEnParalelo(t *testing.T, concurrencia, minimo int) informeEscrito {
	t.Helper()

	s := escribirSustitutos(t)
	ejecucion := sesionesDelRepartidor(t, s, concurrencia, 1, true)

	for _, sesion := range ejecucion.Plan {
		s.escribirTranscript(t, sesion.Nombre, transcriptTerminado)
	}

	inicio := time.Now()
	ejecutada, err := ejecutarSesiones(t.Context().Done(), ejecucion)
	transcurrido := time.Since(inicio)
	require.NoError(t, err)

	assert.Equal(t, nombresDe(ejecucion.Plan), ejecutada.Abiertas, "se abren todas, en el orden del plan")
	assert.Empty(t, ejecutada.SinAbrir)

	oleadas := (sesionesDelPlan + concurrencia - 1) / concurrencia
	assert.GreaterOrEqual(t, ejecutada.Duracion, time.Duration(oleadas)*time.Second,
		"la duración cubre las esperas que la concurrencia no deja solapar")
	assert.LessOrEqual(t, ejecutada.Duracion, transcurrido)

	maximo := 0

	for _, sesion := range ejecucion.Plan {
		dir := filepath.Join(ejecucion.Sesiones, sesion.Nombre)
		require.Truef(t, s.llego(t, sesion.Nombre), "el sustituto de claude se ejecuta en la sesion; sesion.err:\n%s",
			contenidoDeLaSesion(t, dir, ficheroDeSalidaDeError))

		maximo = max(maximo, s.abiertasAlLlegar(t, sesion.Nombre))

		exigirLosDirectoriosDeLaSesion(t, dir, ejecucion.Skills)
		exigirLosDirectoriosPropios(t, s, dir, sesion.Nombre)
	}

	assert.LessOrEqualf(t, maximo, concurrencia, "sesiones abiertas a la vez con concurrencia %d", concurrencia)
	assert.GreaterOrEqualf(t, maximo, minimo, "sesiones abiertas a la vez con concurrencia %d", concurrencia)

	exigirQueNingunaEscribeEnOtra(t, ejecucion)

	return informeDeLasSesiones(t, ejecucion, ejecutada)
}

// exigirLosDirectoriosPropios exige que el sustituto de claude se ejecute en
// trabajo/ de la sesión y vea CLAUDE_CONFIG_DIR, TMPDIR, CLAUDE_CODE_TMPDIR y
// KITLEGAL_CACHE_DIR en claude/, tmp/ y cache/ de su directorio.
func exigirLosDirectoriosPropios(t *testing.T, s sustitutos, dir, sesion string) {
	t.Helper()

	anotado := s.entorno(t, sesion)

	trabajo, err := filepath.EvalSymlinks(filepath.Join(dir, "trabajo"))
	require.NoError(t, err)
	assert.Equal(t, trabajo, anotado[claveDelDirectorio], "el directorio de trabajo de la sesion")

	propios := map[string]string{
		"CLAUDE_CONFIG_DIR":  "claude",
		"TMPDIR":             "tmp",
		"CLAUDE_CODE_TMPDIR": "tmp",
		"KITLEGAL_CACHE_DIR": "cache",
	}
	for variable, subdirectorio := range propios {
		assert.Equalf(t, filepath.Join(dir, subdirectorio), anotado[variable], "la variable %s en la sesion", variable)
	}
}

// exigirQueNingunaEscribeEnOtra exige que lo que escribe el sustituto de claude
// de cada sesión en cada directorio en el que escribiría Claude Code esté solo
// en cache/, claude/, tmp/ y trabajo/ de su sesión: ningún directorio lo
// escriben dos sesiones (FR-032; SC-008).
func exigirQueNingunaEscribeEnOtra(t *testing.T, ejecucion SesionesAEjecutar) {
	t.Helper()

	escritos := map[string][]string{}

	err := filepath.WalkDir(ejecucion.Sesiones, func(ruta string, entrada fs.DirEntry, err error) error {
		if err != nil {
			return err
		}

		sesion, escrito := strings.CutPrefix(entrada.Name(), prefijoDeLoEscrito)
		if !escrito {
			return nil
		}

		relativa, err := filepath.Rel(ejecucion.Sesiones, filepath.Dir(ruta))
		escritos[sesion] = append(escritos[sesion], relativa)

		return err
	})
	require.NoError(t, err)

	esperados := map[string][]string{}
	for _, sesion := range ejecucion.Plan {
		for _, subdirectorio := range []string{"cache", "claude", "tmp", "trabajo"} {
			esperados[sesion.Nombre] = append(esperados[sesion.Nombre], filepath.Join(sesion.Nombre, subdirectorio))
		}
	}

	assert.Equal(t, esperados, escritos, "lo que escribe cada sesion, por directorio")
}

// informeDeLasSesiones escribe con EscribirInforme el informe de las sesiones
// ejecutadas, con duracionDelInforme, y devuelve informe.json e informe.md.
func informeDeLasSesiones(t *testing.T, ejecucion SesionesAEjecutar, ejecutada EjecucionDeSesiones) informeEscrito {
	t.Helper()

	destino := t.TempDir()

	informe, err := EscribirInforme(InformeAEscribir{
		Skill:                 skillDeLasSesiones,
		Evals:                 ejecucion.Evals,
		Sesiones:              ejecucion.Sesiones,
		Destino:               destino,
		ModeloQueDecide:       modeloDeLaSesion,
		Repeticiones:          repeticionesDelRepartidor,
		Umbral:                1,
		Commit:                commitEvaluado,
		SinPython:             filepath.Join(casosDeInforme, ficheroSinPython),
		SinAbrir:              ejecutada.SinAbrir,
		DuracionDeLasSesiones: duracionDelInforme,
	})
	require.NoError(t, err)
	require.Len(t, informe.Evals, sesionesDelPlan, "el informe juzga todas las sesiones")

	return informeEscrito{
		json: contenidoDelInforme(t, destino, ficheroDelInformeJSON),
		md:   contenidoDelInforme(t, destino, ficheroDelInformeMD),
	}
}

// sesionDelLimite es la posición en el plan de la sesión que da el transcript de
// un límite en TestEjecutarSesionesTrasElLimiteDeUso: la tercera, que se abre
// cuando termina una de las dos primeras.
const sesionDelLimite = 2

// TestEjecutarSesionesTrasElLimiteDeUso fija que el repartidor no abre ninguna
// sesión más tras el mensaje del límite de uso, y sí tras los reintentos
// agotados (contracts/ejecucion-del-job.md §3 de H7.3; data-model §3 de H7.3;
// research D6 y D7 de H7.3; FR-044, FR-093; SC-007; US3-2, US3-3), con
// Concurrencia 2 y un plan de ocho sesiones en el que la tercera da el
// transcript del límite y termina enseguida y las demás duermen 1 s. Con el
// mensaje del límite de uso, (a), se abren la tercera y, como mucho, la que ya
// estaba abierta con ella: las abiertas son las primeras del plan, entre tres y
// cuatro; todas terminan y se leen; ninguna posterior llega a abrirse ni a tener
// directorio; y SinAbrir son exactamente las que faltan del plan, en su orden.
// Con los reintentos agotados, (b), se abren todas y SinAbrir queda vacía.
func TestEjecutarSesionesTrasElLimiteDeUso(t *testing.T) {
	t.Parallel()

	casos := []struct {
		nombre     string
		transcript func(t *testing.T) string
		clase      ClaseDeLimite
	}{
		{
			nombre: "mensaje-del-limite-de-uso",
			transcript: func(t *testing.T) string {
				t.Helper()

				return mensajeInit + mensajeResultConError(t, textoDelLimiteDeSesion)
			},
			clase: LimiteMensajeDeUso,
		},
		{
			nombre: "reintentos-agotados",
			transcript: func(t *testing.T) string {
				t.Helper()

				return mensajeInit + mensajeDeReintento(9, 10, 429, "rate_limit") +
					mensajeDeReintento(10, 10, 429, "rate_limit") + mensajeResultConError(t, textoDelError429)
			},
			clase: LimiteReintentosAgotados,
		},
	}

	for _, caso := range casos {
		t.Run(caso.nombre, func(t *testing.T) {
			t.Parallel()

			s := escribirSustitutos(t)
			ejecucion := sesionesDelRepartidor(t, s, 2, 1, false)

			for posicion, sesion := range ejecucion.Plan {
				transcript := transcriptTerminado
				if posicion == sesionDelLimite {
					transcript = caso.transcript(t)
					s.escribirEspera(t, sesion.Nombre, 0)
				}

				s.escribirTranscript(t, sesion.Nombre, transcript)
			}

			ejecutada, err := ejecutarSesiones(t.Context().Done(), ejecucion)
			require.NoError(t, err)

			limite, err := LeerSesion(filepath.Join(ejecucion.Sesiones, ejecucion.Plan[sesionDelLimite].Nombre))
			require.NoError(t, err)
			require.Equal(t, caso.clase, ClasificarElLimite(limite).Clase, "premisa: la sesion da el transcript de su clase")

			abiertas := len(ejecutada.Abiertas)
			if caso.clase == LimiteMensajeDeUso {
				assert.GreaterOrEqual(t, abiertas, sesionDelLimite+1, "se abre la sesion del limite")
				assert.LessOrEqual(t, abiertas, sesionDelLimite+ejecucion.Concurrencia,
					"tras el limite solo sigue la que ya estaba abierta con ella")
				assert.Equal(t, ejecucion.Plan[abiertas:], ejecutada.SinAbrir, "SinAbrir son las que faltan del plan")
			} else {
				assert.Equal(t, sesionesDelPlan, abiertas, "tras los reintentos agotados se abren todas")
				assert.Empty(t, ejecutada.SinAbrir)
			}

			assert.Equal(t, nombresDe(ejecucion.Plan[:abiertas]), ejecutada.Abiertas, "las abiertas, en el orden del plan")

			for _, sesion := range ejecucion.Plan[:abiertas] {
				exigirTerminada(t, s, ejecucion, sesion)
			}

			exigirSinAbrir(t, s, ejecucion, ejecucion.Plan[abiertas:])
		})
	}
}

// exigirTerminada exige que el sustituto de claude llegue a la sesión y cumpla
// su espera, que la sesión quede con un código que LeerSesion lee y que no
// quede ningún proceso en su grupo.
func exigirTerminada(t *testing.T, s sustitutos, ejecucion SesionesAEjecutar, sesion SesionPlanificada) {
	t.Helper()

	require.Truef(t, s.llego(t, sesion.Nombre), "la sesion %s se abre", sesion.Nombre)
	assert.Truef(t, s.cumplioLaEspera(t, sesion.Nombre), "la sesion %s termina", sesion.Nombre)

	_, err := LeerSesion(filepath.Join(ejecucion.Sesiones, sesion.Nombre))
	require.NoErrorf(t, err, "la sesion %s se lee", sesion.Nombre)

	exigirGrupoTerminado(t, s, sesion.Nombre)
}

// exigirCerrada exige que el sustituto de claude llegue a la sesión y la cierre
// TERM sin que cumpla su espera, que la sesión quede sin codigo-de-la-sesion y
// que no quede ningún proceso en su grupo.
func exigirCerrada(t *testing.T, s sustitutos, ejecucion SesionesAEjecutar, sesion SesionPlanificada) {
	t.Helper()

	require.Truef(t, s.llego(t, sesion.Nombre), "la sesion %s se abre", sesion.Nombre)
	assert.Truef(t, s.recibioTERM(t, sesion.Nombre), "la sesion %s recibe TERM", sesion.Nombre)
	assert.Falsef(t, s.cumplioLaEspera(t, sesion.Nombre), "la sesion %s no cumple su espera", sesion.Nombre)
	assert.NoFileExists(t, filepath.Join(ejecucion.Sesiones, sesion.Nombre, ficheroDelCodigo))

	exigirGrupoTerminado(t, s, sesion.Nombre)
}

// exigirGrupoTerminado exige que, poco después, no quede ningún proceso en el
// grupo del sustituto de claude de la sesión: los que el grupo deja sin esperar
// al terminar los recoge el sistema.
func exigirGrupoTerminado(t *testing.T, s sustitutos, sesion string) {
	t.Helper()

	grupo := s.grupoDeProcesos(t, sesion)
	assert.Eventuallyf(t, func() bool { return grupoTerminado(grupo) }, topeSinCorte, 10*time.Millisecond,
		"no queda ningun proceso en el grupo de la sesion %s", sesion)
}

// exigirSinAbrir exige que ninguna de las sesiones llegue al sustituto de claude
// ni tenga directorio.
func exigirSinAbrir(t *testing.T, s sustitutos, ejecucion SesionesAEjecutar, sesiones []SesionPlanificada) {
	t.Helper()

	for _, sesion := range sesiones {
		assert.Falsef(t, s.llego(t, sesion.Nombre), "la sesion %s no se abre", sesion.Nombre)
		assert.NoDirExists(t, filepath.Join(ejecucion.Sesiones, sesion.Nombre))
	}
}

// TestEjecutarSesionesConElContextoCancelado fija la interrupción del reparto
// (contracts/ejecucion-del-job.md §3 y §8 de H7.3; research D9 de H7.3; FR-037,
// FR-064): con Concurrencia 2 y un plan de ocho sesiones en las que el
// sustituto de claude duerme 5 s, cuando las dos primeras han empezado su
// espera se cancela el contexto cuyo Done() recibe el repartidor, lo que hacen
// SIGINT y SIGTERM en las entradas; vuelve con errSesionInterrumpida, las dos
// abiertas reciben TERM y no cumplen su espera ni escriben su código, ninguna
// otra llega a abrirse ni a tener directorio, y no queda ningún proceso en el
// grupo de ninguna.
func TestEjecutarSesionesConElContextoCancelado(t *testing.T) {
	t.Parallel()

	s := escribirSustitutos(t)
	ejecucion := sesionesDelRepartidor(t, s, 2, 5, false)

	contexto, cancelar := context.WithCancel(t.Context())
	defer cancelar()

	interrupcion := contexto.Done()
	resultado := make(chan error, 1)

	go func() {
		_, err := ejecutarSesiones(interrupcion, ejecucion)
		resultado <- err
	}()

	for _, sesion := range ejecucion.Plan[:ejecucion.Concurrencia] {
		empezada := s.esperaEmpezada(sesion.Nombre)
		require.Eventuallyf(t, func() bool {
			_, err := os.Lstat(empezada)

			return err == nil
		}, topeSinCorte, 10*time.Millisecond, "el sustituto de claude empieza su espera en %s", sesion.Nombre)
	}

	cancelar()

	require.ErrorIs(t, <-resultado, errSesionInterrumpida)

	for _, sesion := range ejecucion.Plan[:ejecucion.Concurrencia] {
		exigirCerrada(t, s, ejecucion, sesion)
	}

	exigirSinAbrir(t, s, ejecucion, ejecucion.Plan[ejecucion.Concurrencia:])
}

// TestEjecutarSesionesConUnError fija que un error que impide abrir una sesión
// cierra las abiertas con la secuencia del tope, como la interrupción, y vuelve
// con ese error y no con el de las sesiones que cierra
// (contracts/ejecucion-del-job.md §3 de H7.3; research D7 de H7.3; FR-037): con
// Concurrencia 2, la primera sesión duerme 5 s, la segunda 1 s y la tercera no
// se puede crear porque su directorio ya tiene trabajo/. La segunda termina, la
// tercera no llega a abrirse y la primera recibe TERM sin cumplir su espera;
// ninguna posterior a la tercera se abre ni tiene directorio, no queda ningún
// proceso en el grupo de ninguna, y el error nombra la tercera sesión.
func TestEjecutarSesionesConUnError(t *testing.T) {
	t.Parallel()

	s := escribirSustitutos(t)
	ejecucion := sesionesDelRepartidor(t, s, 2, 1, false)

	primera, tercera := ejecucion.Plan[0].Nombre, ejecucion.Plan[2].Nombre
	s.escribirEspera(t, primera, 5)
	require.NoError(t, os.MkdirAll(filepath.Join(ejecucion.Sesiones, tercera, "trabajo"), 0o700))

	_, err := ejecutarSesiones(t.Context().Done(), ejecucion)

	require.Error(t, err)
	require.ErrorContains(t, err, "la sesión "+tercera+":")
	require.NotErrorIs(t, err, errSesionInterrumpida, "el error es el de la sesion que no se abre")

	exigirCerrada(t, s, ejecucion, ejecucion.Plan[0])
	exigirTerminada(t, s, ejecucion, ejecucion.Plan[1])
	assert.False(t, s.llego(t, tercera), "la sesion que no se puede crear no se abre")
	exigirSinAbrir(t, s, ejecucion, ejecucion.Plan[3:])
}

// TestEjecutarSesionesSinConcurrencia fija que una Concurrencia menor que 1 es
// un error antes de abrir ninguna sesión (contracts/ejecucion-del-job.md §1 y §3
// de H7.3; FR-030): el error nombra la concurrencia y el directorio de sesiones
// queda vacío.
func TestEjecutarSesionesSinConcurrencia(t *testing.T) {
	t.Parallel()

	for _, concurrencia := range []int{0, -1} {
		t.Run(strconv.Itoa(concurrencia), func(t *testing.T) {
			t.Parallel()

			s := escribirSustitutos(t)
			ejecucion := sesionesDelRepartidor(t, s, concurrencia, 0, false)

			_, err := ejecutarSesiones(t.Context().Done(), ejecucion)
			require.ErrorContains(t, err, "la concurrencia es "+strconv.Itoa(concurrencia)+
				" y tiene que ser un entero mayor o igual que 1")

			entradas, err := os.ReadDir(ejecucion.Sesiones)
			require.NoError(t, err)
			assert.Empty(t, entradas, "no se abre ninguna sesion")
		})
	}
}

// sesionesDelRepartidor son las sesiones de los tests del repartidor: las ocho
// del plan de repeticionesDelRepartidor, con la concurrencia dada, un tope que
// no corta ninguna y la base de los sustitutos con la espera dada, en segundos,
// para las sesiones que no tengan la suya.
func sesionesDelRepartidor(t *testing.T, s sustitutos, concurrencia, espera int, traza bool) SesionesAEjecutar {
	t.Helper()

	ejecucion := sesionesConRepeticiones(t, s, traza, repeticionesDelRepartidor)
	require.Len(t, ejecucion.Plan, sesionesDelPlan, "premisa: el plan tiene ocho sesiones")

	ejecucion.Concurrencia = concurrencia
	ejecucion.Tope = topeSinCorte
	ejecucion.Entorno = s.base(variableDeEspera + "=" + strconv.Itoa(espera))

	return ejecucion
}

// nombresDe son los nombres de las sesiones, en su orden.
func nombresDe(sesiones []SesionPlanificada) []string {
	nombres := make([]string, 0, len(sesiones))
	for _, sesion := range sesiones {
		nombres = append(nombres, sesion.Nombre)
	}

	return nombres
}

// sesionesDelTest son las sesiones que abre un test con los sustitutos: el plan
// de la eval sintética del art. 21 de la LPAC con modeloDeLaSesion, una
// repetición y la prueba de red; su directorio de sesiones y el de las skills
// instaladas, temporales; el guion de la sesión; la base de los sustitutos; y el
// tope y el margen de los tests.
func sesionesDelTest(t *testing.T, s sustitutos, traza bool) SesionesAEjecutar {
	t.Helper()

	return sesionesConRepeticiones(t, s, traza, 1)
}

// sesionesConRepeticiones son las de sesionesDelTest con las repeticiones dadas
// de la eval sintética.
func sesionesConRepeticiones(t *testing.T, s sustitutos, traza bool, repeticiones int) SesionesAEjecutar {
	t.Helper()

	evals := crearConjunto(t, []entradaDeConjunto{{nombre: nombreDeEval, contenido: contenidoDelArticulo21}})

	conjunto, err := LeerConjunto(evals)
	require.NoError(t, err)
	require.Empty(t, conjunto.MalFormados)

	plan := PlanDeEvals{
		Evals: conjunto.Evals, ModeloQueDecide: modeloDeLaSesion, Repeticiones: repeticiones, PruebaDeRed: true,
	}
	require.NoError(t, plan.Comprobar())

	guion, err := filepath.Abs(guionDeLaSesion)
	require.NoError(t, err)

	return SesionesAEjecutar{
		Plan:          plan.Sesiones(),
		Concurrencia:  1,
		Evals:         evals,
		Sesiones:      t.TempDir(),
		Skills:        skillsInstaladas(t),
		Guion:         guion,
		Entorno:       s.base(),
		Traza:         traza,
		Tope:          topeDeLosTests,
		MargenDelTope: margenDeLosTests,
	}
}

// ficheroDeEvalSinBinario es la eval sin binario ni servidor de los tests del
// repartidor con los dos modos.
const ficheroDeEvalSinBinario = "02-sin-binario-ni-servidor.yaml"

// sesionesEnDosModos son las sesiones que abre un test del repartidor con los
// dos modos: el plan de la eval sintética del art. 21 de la LPAC y de una eval
// sin binario ni servidor, con modeloDeLaSesion y una repetición —una sesión del
// modo orden, una del modo herramienta y una sin binario ni servidor—; un
// kitlegal que no hace nada, en su propio directorio, como Binario; y la base de
// los sustitutos con el PATH de pathConElBinario. Lo demás, como en
// sesionesDelTest, sin traza y con un tope que no corta ninguna.
func sesionesEnDosModos(t *testing.T, s sustitutos) SesionesAEjecutar {
	t.Helper()

	evals := crearConjunto(t, []entradaDeConjunto{
		{nombre: nombreDeEval, contenido: contenidoDelArticulo21},
		{nombre: ficheroDeEvalSinBinario, contenido: evalSinBinarioDeBoeLegislacion},
	})

	conjunto, err := LeerConjunto(evals)
	require.NoError(t, err)
	require.Empty(t, conjunto.MalFormados)

	plan := PlanDeEvals{
		Evals: conjunto.Evals, ModeloQueDecide: modeloDeLaSesion, Repeticiones: 1,
		Modos: []Modo{ModoOrden, ModoHerramienta},
	}
	require.NoError(t, plan.Comprobar())

	guion, err := filepath.Abs(guionDeLaSesion)
	require.NoError(t, err)

	binario := escribirElKitlegalDeLaBase(t)

	return SesionesAEjecutar{
		Plan:          plan.Sesiones(),
		Concurrencia:  1,
		Evals:         evals,
		Sesiones:      t.TempDir(),
		Skills:        skillsInstaladas(t),
		Guion:         guion,
		Binario:       binario,
		Entorno:       sobreLaBase(s.base(), []string{"PATH=" + pathConElBinario(s, filepath.Dir(binario))}),
		Tope:          topeSinCorte,
		MargenDelTope: margenDeLosTests,
	}
}

// pathConElBinario es el PATH de la base de sesionesEnDosModos: el de la base
// de los sustitutos con el directorio del binario detrás del de los sustitutos
// y, otra vez y escrito con la barra final, al final. Sin ese directorio, las
// dos veces, es el de la base de los sustitutos.
func pathConElBinario(s sustitutos, delBinario string) string {
	return strings.Join([]string{
		s.bin, delBinario, os.Getenv("PATH"), delBinario + string(filepath.Separator),
	}, string(os.PathListSeparator))
}

// escribirElKitlegalDeLaBase escribe, en un directorio temporal del test y a
// través de un os.Root, un kitlegal que no hace nada, y devuelve su ruta: el
// Binario de los tests del repartidor con los dos modos.
func escribirElKitlegalDeLaBase(t *testing.T) string {
	t.Helper()

	dir := t.TempDir()

	raiz, err := os.OpenRoot(dir)
	require.NoError(t, err)

	defer func() { require.NoError(t, raiz.Close()) }()

	require.NoError(t, escribirEjecutable(raiz, programaDeLasConsultas, kitlegalQueNoHaceNada))

	return filepath.Join(dir, programaDeLasConsultas)
}

// skillsInstaladas crea un directorio temporal de skills instaladas con una
// carpeta y un enlace a otra, como lo deja make install, y devuelve su ruta.
func skillsInstaladas(t *testing.T) string {
	t.Helper()

	skills := t.TempDir()
	require.NoError(t, os.Mkdir(filepath.Join(skills, skillEnCarpeta), 0o750))

	otra := t.TempDir()
	require.NoError(t, os.Symlink(otra, filepath.Join(skills, skillEnlazada)))

	return skills
}
