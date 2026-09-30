package evals

import (
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

	pregunta := strings.TrimSuffix(contenidoDeLaSesion(t, dir, ficheroDeLaPregunta), "\n")
	require.Contains(t, pregunta, "\n\n", "la pregunta de la prueba de red lleva varias lineas")

	orden := []string{
		"-p", pregunta, "--model", modeloDeLaSesion, "--output-format", "stream-json", "--verbose",
		"--max-turns", "30", "--no-session-persistence", "--setting-sources", "user",
		"--settings", `{"sandbox":{"enabled":false}}`, "--permission-mode", "bypassPermissions",
		"--disallowedTools", "WebFetch", "WebSearch",
	}
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

// sesionesDelTest son las sesiones que abre un test con los sustitutos: el plan
// de la eval sintética del art. 21 de la LPAC con modeloDeLaSesion, una
// repetición y la prueba de red; su directorio de sesiones y el de las skills
// instaladas, temporales; el guion de la sesión; la base de los sustitutos; y el
// tope y el margen de los tests.
func sesionesDelTest(t *testing.T, s sustitutos, traza bool) SesionesAEjecutar {
	t.Helper()

	evals := crearConjunto(t, []entradaDeConjunto{{nombre: nombreDeEval, contenido: contenidoDelArticulo21}})

	conjunto, err := LeerConjunto(evals)
	require.NoError(t, err)
	require.Empty(t, conjunto.MalFormados)

	plan := PlanDeEvals{Evals: conjunto.Evals, ModeloQueDecide: modeloDeLaSesion, Repeticiones: 1, PruebaDeRed: true}
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
