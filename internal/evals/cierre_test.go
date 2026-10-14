package evals

import (
	"encoding/json"
	jsonv2 "encoding/json/v2"
	"fmt"
	"maps"
	"os"
	"os/exec"
	"path/filepath"
	"slices"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const (
	// guionesDelWorkflow es el directorio de los guiones del workflow, y
	// guionDelCierre y comunDelWorkflow, los dos que el test copia a un árbol
	// temporal: el cierre se sitúa en la raíz que deduce de su propia ruta y
	// carga el común desde ella.
	guionesDelWorkflow = "../../scripts/workflow"
	guionDelCierre     = "cierre.sh"
	comunDelWorkflow   = "comun.sh"

	// marcaDeOrdenDeBytes es la que GitHub Actions pone al empezar cada trozo
	// del registro de un trabajo, y que `gh run view --log` deja delante de la
	// hora de esa línea.
	marcaDeOrdenDeBytes = "\xef\xbb\xbf"

	// variableDelRegistro lleva al gh sustituto la ruta del registro que
	// devuelve como el de un trabajo.
	variableDelRegistro = "KITLEGAL_TEST_REGISTRO_DEL_TRABAJO"

	// ghDelCierre es el gh que el cierre encuentra en el test: tiene sesión, al
	// pedirle el registro de un trabajo da el del caso y la propuesta de la rama
	// es la 140. No toca la red.
	ghDelCierre = `#!/bin/sh
case "$1 $2" in
  "auth status") exit 0;;
  "run view") cat "$` + variableDelRegistro + `";;
  "pr view") echo 140;;
  *) echo "gh sustituto: orden inesperada: $*" >&2; exit 1;;
esac
`

	directorioDelHitoDelCierre = "specs/000-hito-del-test"
	skillDelCierre             = "boe-legislacion"

	// lineaDeOtraSalida es la que make escribe en su salida de error cuando el
	// guion de las evals termina en fallo, y avisoDeLasDescartadas, lo que el
	// cierre dice de las que encuentra dentro del informe, delante de su número.
	lineaDeOtraSalida     = "make: *** [Makefile:135: evals] Error 1"
	avisoDeLasDescartadas = "líneas de otra salida, descartadas: "
)

// TestRecogerLasEvalsDelCierre ejerce `scripts/workflow/cierre.sh evals`, el
// paso sin modelo que copia a gates/evals/<skill>.json el informe que el trabajo
// de evals imprime en su registro, con un gh que devuelve un registro escrito
// aquí.
//
// Existe por el run de H21: el informe de boe-legislacion en dos modos pesaba
// 821 kB, el registro lo partió en dos trozos y el segundo empezaba a mitad del
// JSON, con la marca de orden de bytes delante de la hora. El guion no quitaba
// la marca, la hora se quedaba dentro del informe, que dejaba de ser JSON, y el
// informe final salió sin los umbrales de la skill, que el trabajo cumplía.
//
// Y por el de H25: el trabajo de boe-legislacion falló, y la línea con la que
// make lo dice en su salida de error quedó en el registro dentro del informe,
// que el runner aún no había terminado de entregar por la salida estándar. Lo
// recogido no era JSON, el cierre no dejó fichero y la reparación trabajó sin
// las respuestas ni los umbrales de la medición.
//
// Necesita jq, que es con lo que el guion lee y valida: donde no está, el caso
// se salta diciéndolo; la CI lo tiene.
func TestRecogerLasEvalsDelCierre(t *testing.T) {
	t.Parallel()

	if _, err := exec.LookPath("jq"); err != nil {
		t.Skip("el guion del cierre lee con jq, que no está en el PATH")
	}

	informe := map[string]any{
		"skill":     skillDelCierre,
		"veredicto": "aprobado",
		"tasas":     []any{map[string]any{"eval": "01.yaml", "pasan": 3}, map[string]any{"eval": "02.yaml", "pasan": 2}},
		"umbrales":  []any{map[string]any{"nombre": "sin_activar:modelo:orden", "cumple": true, "decide": true}},
	}

	casos := []struct {
		nombre string
		// marcas son las líneas del informe, contadas desde 0, que empiezan un
		// trozo del registro; -1 es la primera línea del registro.
		marcas []int
		// ajenas son las líneas del informe, contadas desde 0, delante de las
		// que el registro trae una línea de otra salida; -1 la pone antes de la
		// marca de inicio.
		ajenas []int
		// roto deja el informe sin su última línea: lo que se recoge no es JSON.
		roto    bool
		recoge  bool
		deError string
		// descartadas es el número de líneas de otra salida que el guion dice
		// haber quitado del informe, o vacío si no dice nada de ellas.
		descartadas string
	}{
		{
			nombre: "un solo trozo", marcas: []int{-1},
			recoge: true, deError: "evals: informe de " + skillDelCierre,
		},
		{
			nombre: "un trozo empieza a mitad del informe", marcas: []int{-1, 6},
			recoge: true, deError: "evals: informe de " + skillDelCierre,
		},
		{
			nombre: "un trozo empieza en cada línea del informe", marcas: []int{-1, 0, 1, 2, 3, 4, 5, 6, 7, 8, 9},
			recoge: true, deError: "evals: informe de " + skillDelCierre,
		},
		{
			nombre: "el informe no es legible", marcas: []int{-1}, roto: true,
			recoge: false, deError: "no trae un informe.json legible",
		},
		{
			nombre: "una línea de otra salida cae antes del informe", marcas: []int{-1}, ajenas: []int{-1},
			recoge: true, deError: "evals: informe de " + skillDelCierre,
		},
		{
			nombre: "una línea de otra salida cae dentro del informe", marcas: []int{-1}, ajenas: []int{6},
			recoge: true, deError: "evals: informe de " + skillDelCierre, descartadas: "1",
		},
		{
			nombre: "dos líneas de otra salida caen dentro del informe, una al empezar un trozo",
			marcas: []int{-1, 6}, ajenas: []int{0, 6},
			recoge: true, deError: "evals: informe de " + skillDelCierre, descartadas: "2",
		},
	}

	for _, caso := range casos {
		t.Run(caso.nombre, func(t *testing.T) {
			t.Parallel()

			escrito, err := json.MarshalIndent(informe, "", "  ")
			require.NoError(t, err)

			lineas := strings.Split(string(escrito), "\n")
			require.Greater(t, len(lineas), 10, "el informe tiene líneas bastantes para partirlo donde dicen los casos")

			if caso.roto {
				lineas = lineas[:len(lineas)-1]
			}

			raiz := arbolDelCierre(t, registroDelTrabajo(lineas, caso.marcas, caso.ajenas), nil)

			codigo, deError := ejecutarElCierre(t, raiz)

			assert.Equalf(t, 0, codigo, "recoger los informes nunca para el run; salida de error:\n%s", deError)
			assert.Contains(t, deError, caso.deError, "lo que el guion dice del informe")

			if caso.descartadas == "" {
				assert.NotContains(t, deError, avisoDeLasDescartadas, "sin líneas de otra salida dentro, el guion no dice nada de ellas")
			} else {
				assert.Contains(t, deError, avisoDeLasDescartadas+caso.descartadas+"\n",
					"el guion dice cuántas líneas de otra salida quitó del informe")
			}

			arbol, err := os.OpenRoot(raiz)
			require.NoError(t, err)

			defer func() { require.NoError(t, arbol.Close()) }()

			recogido, err := arbol.ReadFile(directorioDelHitoDelCierre + "/gates/evals/" + skillDelCierre + ".json")
			if !caso.recoge {
				require.ErrorIs(t, err, os.ErrNotExist, "un informe ilegible no deja fichero")

				return
			}

			require.NoError(t, err, "el informe queda en gates/evals/<skill>.json")
			assert.JSONEq(t, string(escrito), string(recogido),
				"lo recogido es el informe que el trabajo imprimió, sin horas, marcas ni líneas de otra salida")
		})
	}
}

// TestSalidasDeLosPasosQueImprimen fija que los dos pasos del flujo de evals
// que imprimen en el registro un fichero entre dos marcas —el informe del job,
// que recoge el cierre de un hito, y la medida del juez, que versiona una
// persona— llevan su salida de error por la misma tubería que la estándar.
//
// El runner lee las dos por separado y las junta en el registro según le
// llegan. Sin el 2>&1, cuando el guion termina en fallo, la línea con la que
// make lo dice queda en un punto cualquiera de lo que la salida estándar aún
// no había entregado: en la medición 1 del cierre de H25, dentro del informe.
func TestSalidasDeLosPasosQueImprimen(t *testing.T) {
	t.Parallel()

	const esperado = `make evals SKILL="$SKILL_EVALUADA" 2>&1` + "\n" +
		`make evals-medir-juez SKILL="$SKILL_EVALUADA" 2>&1`

	flujo, err := leerFichero(rutaDeLaDefinicionDelJob)
	require.NoError(t, err)

	var ordenes []string

	for linea := range strings.Lines(string(flujo)) {
		if orden, es := strings.CutPrefix(strings.TrimSpace(linea), "run: "); es && strings.HasPrefix(orden, "make evals") {
			ordenes = append(ordenes, orden)
		}
	}

	assert.Equal(t, esperado, strings.Join(ordenes, "\n"),
		"los pasos que ejecutan las evals y la medida del juez, en su orden, con las dos salidas por una tubería")
}

// registroDelTrabajo escribe el registro que `gh run view --log` da de un
// trabajo de evals: cada línea con el trabajo, el paso y la hora delante, y el
// informe entre sus dos marcas de texto. Las líneas de marcas llevan además la
// marca de orden de bytes delante de la hora, como la primera de cada trozo, y
// delante de cada línea de ajenas va una de otra salida: si esa línea empieza un
// trozo, la marca de orden de bytes la lleva la de otra salida.
func registroDelTrabajo(informe []string, marcas, ajenas []int) string {
	const (
		prefijo = "evals (" + skillDelCierre + ")\tEjecutar las evals\t"
		hora    = "2026-10-02T05:37:37.6558296Z "
	)

	var registro strings.Builder

	linea := func(indice int, texto string) {
		registro.WriteString(prefijo)

		if slices.Contains(marcas, indice) {
			registro.WriteString(marcaDeOrdenDeBytes)
		}

		registro.WriteString(hora + texto + "\n")
	}

	// conLaAjena escribe la línea del informe del índice dado y, si el caso lo
	// pide, una de otra salida delante, que es entonces la que empieza el trozo.
	conLaAjena := func(indice int, texto string) {
		if !slices.Contains(ajenas, indice) {
			linea(indice, texto)

			return
		}

		linea(indice, lineaDeOtraSalida)
		linea(-2, texto)
	}

	linea(-1, "##[group]Run make evals")

	if slices.Contains(ajenas, -1) {
		linea(-2, lineaDeOtraSalida)
	}

	linea(-2, "--- inicio de informe.json ---")

	for indice, texto := range informe {
		conLaAjena(indice, texto)
	}

	linea(-2, "--- fin de informe.json ---")
	linea(-2, "Post job cleanup.")

	return registro.String()
}

// arbolDelCierre deja en un temporal lo que el cierre necesita para recoger los
// informes de un hito: un repositorio de git en la rama del hito, con sus tres
// guiones —el cierre, el común y el de la serie, que anota la fila de cada skill
// con juez— y los ficheros commiteados dados; y, fuera del commit, el directorio
// del hito con una medición en la que el trabajo de evals de la skill terminó, el
// sustituto de gh y el registro que gh devuelve. Da la raíz del árbol.
func arbolDelCierre(t *testing.T, registro string, commiteados map[string]string) string {
	t.Helper()

	directorio := t.TempDir()

	raiz, err := os.OpenRoot(directorio)
	require.NoError(t, err)

	defer func() { require.NoError(t, raiz.Close()) }()

	for _, d := range []string{"scripts/workflow", ".specify", directorioDelHitoDelCierre + "/gates", "bin"} {
		require.NoError(t, raiz.MkdirAll(d, 0o750))
	}

	copiarGuiones(t, raiz, guionDelCierre, comunDelWorkflow, guionDeLaSerie)

	ficheros := map[string]string{".specify/feature.json": `{"feature_directory": "` + directorioDelHitoDelCierre + `"}`}
	maps.Copy(ficheros, commiteados)

	for nombre, contenido := range ficheros {
		require.NoError(t, raiz.MkdirAll(filepath.Dir(nombre), 0o750))
		require.NoError(t, raiz.WriteFile(nombre, []byte(contenido), 0o600))
	}

	ejecutarGit(t, directorio, "init", "-q", "-b", "rama-del-test")
	ejecutarGit(t, directorio, "add", "-A")
	ejecutarGit(t, directorio, "commit", "-q", "-m", "el hito del test")

	require.NoError(t, escribirEjecutable(raiz, "bin/gh", ghDelCierre))

	medicion, err := json.Marshal(map[string]any{"checks": []any{map[string]any{
		"workflow": "evals",
		"bucket":   "pass",
		"name":     "evals (" + skillDelCierre + ")",
		"link":     "https://github.com/jmorenobl/kitlegal/actions/runs/1/job/2",
	}}})
	require.NoError(t, err)

	for nombre, contenido := range map[string]string{
		directorioDelHitoDelCierre + "/gates/cierre.json": string(medicion),
		"registro.log": registro,
	} {
		require.NoError(t, raiz.WriteFile(nombre, []byte(contenido), 0o600))
	}

	return directorio
}

// ejecutarElCierre ejecuta `cierre.sh evals` del árbol dado con sus sustitutos
// delante en el PATH y devuelve su código y su salida de error.
func ejecutarElCierre(t *testing.T, raiz string) (codigo int, deError string) {
	t.Helper()

	// El guion es la copia que el test acaba de escribir en su temporal y los
	// argumentos son literales: no hay entrada externa en la orden.
	//nolint:gosec // el guion lo escribe el test en su temporal y los argumentos son literales; no hay entrada externa.
	orden := exec.CommandContext(t.Context(), filepath.Join(raiz, "scripts", "workflow", guionDelCierre), "evals", "H0")
	orden.Env = sobreLaBase(entornoDeGitDelTest(), []string{
		"PATH=" + filepath.Join(raiz, "bin") + string(os.PathListSeparator) + os.Getenv("PATH"),
		variableDelRegistro + "=" + filepath.Join(raiz, "registro.log"),
	})

	var errores strings.Builder

	orden.Stderr = &errores

	err := orden.Run()
	if err != nil {
		var fallo *exec.ExitError

		require.ErrorAs(t, err, &fallo, "el guion se ejecuta")

		return fallo.ExitCode(), errores.String()
	}

	return 0, errores.String()
}

// TestSerieDelCierre ejerce lo que `cierre.sh evals` añade a la serie de cierres
// (ADR 0038, punto 6) al recoger los informes de cada medición de un cierre de
// tres, con el informe de la skill con juez en el registro de su trabajo:
//
//   - las tres mediciones dan tres filas en gates/serie.jsonl, cada una con su
//     medición, su ejecución, su trabajo, su fecha y la propuesta de la rama:
//     cuentan todas, no solo la que pasó;
//   - la segunda, con una respuesta medida sin juzgar, se registra y no cuenta;
//   - recoger otra vez la tercera sustituye su fila, sin duplicarla;
//   - y una cuarta cuya fila no casa con los umbrales de su informe sale con 1,
//     lo dice y deja la serie como estaba.
func TestSerieDelCierre(t *testing.T) {
	t.Parallel()

	if _, err := exec.LookPath("jq"); err != nil {
		t.Skip("el guion del cierre lee con jq, que no está en el PATH")
	}

	raiz := arbolDelCierre(t, "", ficherosDeLaSerie(umbralDeBoe))
	commit := ejecutarGit(t, raiz, "rev-parse", "HEAD")
	serie := filepath.Join(raiz, directorioDelHitoDelCierre, "gates", "serie.jsonl")

	completo := func() Informe {
		informe := informeDeLaSerie(commit, sesionesDeCierre(), true)
		informe.Tasas = []TasaDelInforme{}

		return informe
	}

	sinJuzgar := completo()
	sinJuzgar.Juez.SinJuzgar = []RespuestaSinJuzgar{{Sesion: sesionesDeCierre()[0].nombre(0), Motivo: "voto 3: tope agotado"}}

	noCasa := completo()
	noCasa.Umbrales[0].Medida++

	medir := func(ronda int, informe Informe) (int, string) {
		t.Helper()

		medicion, err := json.Marshal(map[string]any{
			"sha": commit, "ronda": ronda, "fecha": "2026-10-13T13:00:00Z",
			"checks": []any{map[string]any{
				"workflow":  "evals",
				"bucket":    "fail",
				"name":      "evals (" + skillDelCierre + ")",
				"link":      fmt.Sprintf("https://github.com/jmorenobl/kitlegal/actions/runs/%d/job/%d", 10+ronda, 20+ronda),
				"startedAt": fmt.Sprintf("2026-10-1%dT12:33:08Z", ronda),
			}},
		})
		require.NoError(t, err)

		escrito, err := json.MarshalIndent(informe, "", "  ")
		require.NoError(t, err)

		for nombre, contenido := range map[string]string{
			filepath.Join(directorioDelHitoDelCierre, "gates", "cierre.json"): string(medicion),
			"registro.log": registroDelTrabajo(strings.Split(string(escrito), "\n"), []int{-1}, nil),
		} {
			require.NoError(t, os.WriteFile(filepath.Join(raiz, nombre), []byte(contenido), 0o600))
		}

		return ejecutarElCierre(t, raiz)
	}

	for ronda, informe := range []Informe{completo(), sinJuzgar, completo(), completo()} {
		ronda = min(ronda+1, 3)

		codigo, deError := medir(ronda, informe)
		require.Equalf(t, 0, codigo, "la medición %d se recoge: %s", ronda, deError)
		assert.Contains(t, deError, "evals: fila de "+skillDelCierre+" en ")
	}

	escrita, err := leerFichero(serie)
	require.NoError(t, err)

	lineas := strings.Split(strings.TrimSuffix(string(escrita), "\n"), "\n")
	require.Len(t, lineas, 3, "una fila por medición, y la que se recoge otra vez se sustituye")

	for n, linea := range lineas {
		validarContraElEsquemaDeLaSerie(t, []byte(linea))

		var fila filaDeLaSerie

		require.NoError(t, jsonv2.Unmarshal([]byte(linea), &fila))

		propuesta := 140
		assert.Equal(t, procedenciaDeLaFila{
			Hito: "H0", Medicion: n + 1, Fecha: fmt.Sprintf("2026-10-1%dT12:33:08Z", n+1), Commit: commit,
			Propuesta: &propuesta, Ejecucion: 11 + n, Trabajo: 21 + n,
		}, fila.Procedencia)
		assert.Equal(t, n != 1, fila.Cuenta, "la medición %d cuenta salvo la incompleta", n+1)
		require.Len(t, fila.Clases, 1)
		assert.Equal(t, 6, fila.Clases[0].Total, "cada medición registra lo medido")
	}

	codigo, deError := medir(4, noCasa)
	assert.Equal(t, 1, codigo, "una fila que no casa con su informe para el cierre")
	assert.Contains(t, deError, "no casa con los umbrales de su informe")

	despues, err := leerFichero(serie)
	require.NoError(t, err)
	assert.Equal(t, string(escrita), string(despues), "la serie queda como estaba")
}
