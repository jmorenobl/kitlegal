package evals

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

	// ghDelCierre es el gh que el cierre encuentra en el test: tiene sesión y, al
	// pedirle el registro de un trabajo, da el del caso. No toca la red.
	ghDelCierre = `#!/bin/sh
case "$1 $2" in
  "auth status") exit 0;;
  "run view") cat "$` + variableDelRegistro + `";;
  *) echo "gh sustituto: orden inesperada: $*" >&2; exit 1;;
esac
`

	// gitDelCierre responde a la única orden de git que el cierre ejecuta antes
	// de recoger los informes: la rama actual.
	gitDelCierre = `#!/bin/sh
echo rama-del-test
`

	directorioDelHitoDelCierre = "specs/000-hito-del-test"
	skillDelCierre             = "boe-legislacion"
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
		// roto deja el informe sin su última línea: lo que se recoge no es JSON.
		roto    bool
		recoge  bool
		deError string
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

			raiz := arbolDelCierre(t, registroDelTrabajo(lineas, caso.marcas))

			codigo, deError := ejecutarElCierre(t, raiz)

			assert.Equalf(t, 0, codigo, "recoger los informes nunca para el run; salida de error:\n%s", deError)
			assert.Contains(t, deError, caso.deError, "lo que el guion dice del informe")

			arbol, err := os.OpenRoot(raiz)
			require.NoError(t, err)

			defer func() { require.NoError(t, arbol.Close()) }()

			recogido, err := arbol.ReadFile(directorioDelHitoDelCierre + "/gates/evals/" + skillDelCierre + ".json")
			if !caso.recoge {
				require.ErrorIs(t, err, os.ErrNotExist, "un informe ilegible no deja fichero")

				return
			}

			require.NoError(t, err, "el informe queda en gates/evals/<skill>.json")
			assert.JSONEq(t, string(escrito), string(recogido), "lo recogido es el informe que el trabajo imprimió, sin horas ni marcas")
		})
	}
}

// registroDelTrabajo escribe el registro que `gh run view --log` da de un
// trabajo de evals: cada línea con el trabajo, el paso y la hora delante, y el
// informe entre sus dos marcas de texto. Las líneas de marcas llevan además la
// marca de orden de bytes delante de la hora, como la primera de cada trozo.
func registroDelTrabajo(informe []string, marcas []int) string {
	const (
		prefijo = "evals (" + skillDelCierre + ")\tEjecutar las evals\t"
		hora    = "2026-10-02T05:37:37.6558296Z "
	)

	var registro strings.Builder

	linea := func(indice int, texto string) {
		registro.WriteString(prefijo)

		for _, marca := range marcas {
			if marca == indice {
				registro.WriteString(marcaDeOrdenDeBytes)
			}
		}

		registro.WriteString(hora + texto + "\n")
	}

	linea(-1, "##[group]Run make evals")
	linea(-2, "--- inicio de informe.json ---")

	for indice, texto := range informe {
		linea(indice, texto)
	}

	linea(-2, "--- fin de informe.json ---")
	linea(-2, "Post job cleanup.")

	return registro.String()
}

// arbolDelCierre deja en un temporal lo que el cierre necesita para recoger los
// informes de un hito: sus dos guiones, el directorio del hito con una medición
// en la que el trabajo de evals de la skill terminó, los sustitutos de gh y de
// git y el registro que gh devuelve. Da la raíz del árbol.
func arbolDelCierre(t *testing.T, registro string) string {
	t.Helper()

	directorio := t.TempDir()

	raiz, err := os.OpenRoot(directorio)
	require.NoError(t, err)

	defer func() { require.NoError(t, raiz.Close()) }()

	for _, d := range []string{"scripts/workflow", ".specify", directorioDelHitoDelCierre + "/gates", "bin"} {
		require.NoError(t, raiz.MkdirAll(d, 0o750))
	}

	guiones, err := os.OpenRoot(guionesDelWorkflow)
	require.NoError(t, err)

	defer func() { require.NoError(t, guiones.Close()) }()

	for _, nombre := range []string{guionDelCierre, comunDelWorkflow} {
		guion, err := guiones.ReadFile(nombre)
		require.NoError(t, err)
		require.NoError(t, escribirEjecutable(raiz, "scripts/workflow/"+nombre, string(guion)))
	}

	require.NoError(t, escribirEjecutable(raiz, "bin/gh", ghDelCierre))
	require.NoError(t, escribirEjecutable(raiz, "bin/git", gitDelCierre))

	medicion, err := json.Marshal(map[string]any{"checks": []any{map[string]any{
		"workflow": "evals",
		"bucket":   "pass",
		"name":     "evals (" + skillDelCierre + ")",
		"link":     "https://github.com/jmorenobl/kitlegal/actions/runs/1/job/2",
	}}})
	require.NoError(t, err)

	for nombre, contenido := range map[string]string{
		".specify/feature.json":                           `{"feature_directory": "` + directorioDelHitoDelCierre + `"}`,
		directorioDelHitoDelCierre + "/gates/cierre.json": string(medicion),
		"registro.log":                                    registro,
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
	orden.Env = sobreLaBase(os.Environ(), []string{
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
