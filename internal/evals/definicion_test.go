package evals

import (
	"context"
	"errors"
	"fmt"
	"maps"
	"os"
	"path/filepath"
	"regexp"
	"slices"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// Lo que TestDefinicionDelJob exige al trabajo evals de la definición del job
// (contracts/ejecucion-del-job.md §6 y §7 de H7.3; data-model §6 de H7.3).
const (
	// nombreDelTrabajo es su name: el cierre reconoce el informe de cada skill
	// por evals (<skill>) (research.md D12 y V15 de H7.3).
	nombreDelTrabajo = "evals (${{ matrix.skill }})"

	// grupoDelTrabajo es el group de su concurrency: una tanda por commit y por
	// skill (research.md D11 de H7.3).
	grupoDelTrabajo = "evals-${{ github.event.pull_request.head.sha || github.sha }}-${{ matrix.skill }}"
)

// Cómo presentan las líneas de la comprobación una clave que falta y una que
// sobra.
const (
	noEsta    = "no est\xc3\xa1"
	queNoEste = "que no est\xc3\xa9"
)

// Lo que TestDefinicionDelJob exige al trabajo tanda y al trabajo evals que
// depende de él (contracts/tanda-del-job.md §1 y §4 de H7.4; research.md D15,
// D16 y S9 de H7.4).
const (
	// sinCancelar es la función de estado de GitHub Actions que dice que la
	// ejecución no se ha cancelado, negada, partida en dos literales: misspell,
	// con locale US, lee su nombre británico como una errata.
	sinCancelar = "!cance" + "lled()"

	// condicionDeLaTanda es el if del trabajo tanda, el que llevaba evals hasta
	// H7.3: solo corre en una ejecución que quiere medir. Es el texto del
	// escalar plegado de la definición, que conserva los saltos de línea de las
	// líneas más sangradas. Desde H24, el despacho con la entrada de la medida
	// del juez no la corre: lanza la medida, y no las sesiones de evals
	// (contracts/job-de-evals.md §3 de H24). Desde H30, corre también con la
	// etiqueta evals-completa, y la apertura la corre si cambios elige alguna
	// eval (contracts/job-de-evals.md §2 de H30; FR-016 a FR-018).
	condicionDeLaTanda = sinCancelar + " && needs.cambios.result != 'failure' && (\n" +
		"  " + despachoSinLaMedida + " ||\n" +
		"  github.event.label.name == 'evals' ||\n" +
		"  github.event.label.name == 'evals-completa' ||\n" +
		"  github.event.label.name == 'evals-prueba-de-red' ||\n" +
		"  needs.cambios.outputs.elige == 'si'\n" +
		")"

	// permisoDeLasEjecuciones es el de actions: con él lee gh las ejecuciones
	// del flujo y sus trabajos.
	permisoDeLasEjecuciones = "read"

	// salidaDeLaTanda es su output medir, el del paso que decide, y
	// salidaDelModoDeLaTanda, su output modo, del mismo paso (contracts/job-de-evals.md
	// §2 de H30).
	salidaDeLaTanda        = "${{ steps.decidir.outputs.medir }}"
	salidaDelModoDeLaTanda = "${{ steps.decidir.outputs.modo }}"

	// variableDelModoDeEvals es la del env del trabajo tanda que fija el modo de
	// la ejecución, y modoDeLaTanda, su valor: completa solo con el despacho, la
	// etiqueta evals-completa y la etiqueta evals-prueba-de-red, y con
	// reutilización con cualquier otro disparo (contracts/job-de-evals.md §1 de
	// H30; FR-016 a FR-018).
	variableDelModoDeEvals = "MODO_DE_EVALS"
	modoDeLaTanda          = "${{ (github.event_name == 'workflow_dispatch' || " +
		"github.event.label.name == 'evals-completa' || github.event.label.name == 'evals-prueba-de-red') && " +
		"'completa' || 'reutilizacion' }}"

	// ordenDelPasoQueDecide es el run del paso que decide, el de id decidir:
	// TestTandaDelCommit con el commit evaluado, la ejecución, el modo y el
	// GITHUB_OUTPUT del paso. El -timeout supera los 10 min que la decisión
	// espera como mucho, más sus consultas, y cabe con la preparación en los 15
	// del trabajo: con los 10 de go test por omisión, el test acabaría en pánico
	// antes de medir tras la espera, y el trabajo, en rojo.
	ordenDelPasoQueDecide = `go test -tags evals -count=1 -timeout 12m -v -run '^TestTandaDelCommit$' ./internal/evals/ ` +
		`-args -commit "$COMMIT_EVALUADO" -ejecucion "$EJECUCION" -modo "$MODO_DE_EVALS" -salida "$GITHUB_OUTPUT"`

	// condicionDeLaMarca es el if de la marca: solo corre si el paso que
	// decide dice que la ejecución mide. Y condicionDeLaMarcaDeLaSuiteCompleta,
	// el de la segunda marca, la que solo corre si además mide la suite completa
	// (contracts/job-de-evals.md §2 de H30; research D13 de H30).
	condicionDeLaMarca                  = "steps.decidir.outputs.medir == 'si'"
	condicionDeLaMarcaDeLaSuiteCompleta = condicionDeLaMarca + " && steps.decidir.outputs.modo == 'completa'"

	// condicionDelTrabajo es el if del trabajo evals: sin medir=si de la
	// tanda, se salta entero. La función de estado quita el success()
	// implícito, que podría saltarlo por cambios, saltado en la etiqueta y en
	// el despacho (research.md S9 de H7.4).
	condicionDelTrabajo = "${{ " + sinCancelar + " && needs.tanda.outputs.medir == 'si' }}"
)

// Lo que TestDefinicionDelJob exige al trabajo cambios desde H30
// (contracts/job-de-evals.md §2 de H30; FR-016; research D5 y D17 de H30): solo
// corre al abrir o reabrir una propuesta, con su tope y sus permisos de
// lectura; su checkout lleva toda la historia, de la que sale el árbol de la
// base; y su paso de id elegir ejecuta TestEleccionDeLaPropuesta y da la salida
// elige.
const (
	condicionDeCambios = "github.event_name == 'pull_request' && github.event.action != 'labeled'"
	topeDeCambios      = 15
	salidaDeCambios    = "${{ steps.elegir.outputs.elige }}"
	pasoDeLaEleccion   = "Elegir las evals que la propuesta ha podido cambiar"
	ordenDeLaEleccion  = `go test -tags evals -count=1 -timeout 12m -v -run '^TestEleccionDeLaPropuesta$' ` +
		`./internal/evals/ -args -propuesta "$PROPUESTA" -base "$BASE_DE_LA_PROPUESTA" -salida "$GITHUB_OUTPUT"`

	// accionDelCheckout y accionDeGo son las acciones del paso que obtiene el
	// código y del que instala Go, y todaLaHistoria, el fetch-depth del primero
	// que trae todos los commits.
	accionDelCheckout = "actions/checkout@v7"
	accionDeGo        = "actions/setup-go@v7"
	todaLaHistoria    = "0"
)

// Lo que TestDefinicionDelJob exige desde H30 a los pasos del trabajo evals
// (contracts/job-de-evals.md §2 y §3 de H30; FR-010, FR-012, FR-019; research
// D17 y D18 de H30): tras instalar Go, el paso de la decisión, con su plazo de
// 15 minutos y el fichero que escribe; el que ejecuta las evals, con ese
// fichero; y, el último, la marca del informe escrito, que corre aunque el
// anterior falle y es la que lee la búsqueda de medidas de main.
const (
	ordenDeLaDecision = `go test -tags evals -count=1 -timeout 15m -v -run '^TestDecisionDelTrabajo$' ./internal/evals/ ` +
		`-args -skill "$SKILL_EVALUADA" -modo "$MODO_DE_EVALS" -commit "$COMMIT_EVALUADO" -propuesta "$PROPUESTA" ` +
		`-base "$BASE_DE_LA_PROPUESTA" -modelo-que-decide "$MODELO_DE_EVALS" ` +
		`-version-de-claude-code "$VERSION_DE_CLAUDE_CODE" -modelo-del-juez "$MODELO_DEL_JUEZ" ` +
		`-version-del-juez "$VERSION_DE_CLAUDE_CODE_DEL_JUEZ" -salida "$RUNNER_TEMP/decision-de-evals.json"`
	plazoDeLaDecisionDelContrato = 15 * time.Minute

	pasoQueEjecutaLasEvals  = "Ejecutar las evals"
	ordenQueEjecutaLasEvals = `DECISION_DE_EVALS="$RUNNER_TEMP/decision-de-evals.json" make evals SKILL="$SKILL_EVALUADA" 2>&1`

	condicionDeLaMarcaDelInforme = "${{ " + sinCancelar + " }}"
	ordenDeLaMarcaDelInforme     = `test -s "$RUNNER_TEMP/kitlegal-evals-$SKILL_EVALUADA/informe.json"`
)

// Las dos variables con las que gh lee la plataforma: en el trabajo evals, solo
// en el paso de la decisión (contracts/job-de-evals.md §2 de H30).
const (
	variableDelAccesoDeGh = "GH_TOKEN"
	variableDelRepoDeGh   = "GH_REPO"
)

// sesionesALaVezComoMucho son las que el trabajo evals abre a la vez como mucho
// entre todas sus combinaciones: 5, 4 y 2, las de hoy (FR-020 y SC-011 de H30).
const sesionesALaVezComoMucho = 11

// permisosDeCambios y permisosDelTrabajo son los permissions del trabajo
// cambios y del trabajo evals: de lectura, de los ficheros de la propuesta y,
// en evals, también de las ejecuciones del flujo (contracts/job-de-evals.md §2
// de H30).
var (
	permisosDeCambios  = map[string]string{"contents": "read", "pull-requests": "read"}
	permisosDelTrabajo = map[string]string{"contents": "read", "actions": "read", "pull-requests": "read"}
)

// entornoDeLaEleccion y entornoDeLaDecision son los env del paso de la elección
// del trabajo cambios y del paso de la decisión del trabajo evals: los dos con
// lo que gh necesita, y el de la elección, además, la propuesta y su base, que
// el trabajo evals lleva en el suyo.
var (
	entornoDeLaEleccion = map[string]string{
		variableDelAccesoDeGh:  "${{ secrets.GITHUB_TOKEN }}",
		variableDelRepoDeGh:    "${{ github.repository }}",
		"PROPUESTA":            "${{ github.event.pull_request.number }}",
		"BASE_DE_LA_PROPUESTA": "${{ github.event.pull_request.base.sha }}",
	}
	entornoDeLaDecision = map[string]string{
		variableDelAccesoDeGh: "${{ secrets.GITHUB_TOKEN }}",
		variableDelRepoDeGh:   "${{ github.repository }}",
	}
)

// Lo que TestDefinicionDelJob exige al juez con modelo del trabajo evals y a la
// ejecución de su medida (contracts/job-de-evals.md §2, §3 y §5 de H24; FR-050,
// FR-090, FR-091 y FR-112 de H24).
const (
	// etiquetaDeLaMedida es la etiqueta que lanza la medida del juez en una
	// propuesta de cambio, y entradaDeLaMedida, la entrada del despacho que la
	// lanza a mano: las dos formas de lanzarla, y ninguna más.
	etiquetaDeLaMedida = "evals-medir-juez"
	entradaDeLaMedida  = "medir_al_juez"

	// despachoSinLaMedida es, en el if del trabajo tanda, el despacho que quiere
	// medir la skill: el que no lleva la entrada de la medida del juez.
	despachoSinLaMedida = "(github.event_name == 'workflow_dispatch' && inputs." + entradaDeLaMedida + " != true)"

	// condicionDeLaMedida es el if del trabajo medida: solo corre con su
	// etiqueta o con su entrada.
	condicionDeLaMedida = "github.event.label.name == '" + etiquetaDeLaMedida + "' || inputs." + entradaDeLaMedida +
		" == true"

	// claudeCodeDelJuez es el Claude Code de los votos del juez como lo nombra
	// npm: el paquete en la versión fijada para él, que no es la de las
	// sesiones. E instalacionDelJuez, la línea del paso de instalación que lo
	// instala en su prefijo, aparte del de las sesiones.
	claudeCodeDelJuez   = paqueteDeClaudeCode + "@${" + variableDeLaVersionDelJuez + "}"
	instalacionDelJuez  = `npm install --prefix "$RUNNER_TEMP/claude-del-juez" "` + claudeCodeDelJuez + `"`
	pasoDeLaInstalacion = "jobs.evals.steps, el que instala Claude Code, run"

	// modeloDelJuezEsperado y versionDelJuezEsperada son lo que se espera de las
	// dos variables del juez del env del trabajo evals.
	modeloDelJuezEsperado = "el id completo de un modelo (^claude-[a-z]+(-[0-9]+)+$), distinto del de " +
		variableDelModeloQueDecide
	versionDelJuezEsperada = "una versi\xc3\xb3n de Claude Code, <n>.<n>.<n>"
)

// formaDelIDCompleto es la forma del id completo de un modelo, la que no tiene
// un alias, y formaDeLaVersion, la de una versión de Claude Code
// (contracts/job-de-evals.md §5 de H24).
var (
	formaDelIDCompleto = regexp.MustCompile(`^claude-[a-z]+(-[0-9]+)+$`)
	formaDeLaVersion   = regexp.MustCompile(`^[0-9]+\.[0-9]+\.[0-9]+$`)
)

// dependenciasDelTrabajo son las de su needs: solo la tanda.
var dependenciasDelTrabajo = []string{trabajoDeLaTanda}

// skillsDelTrabajo son las skills de su matriz, en su orden. Desde H23,
// jurisprudencia es la tercera (contracts/evals-jurisprudencia.md §6 de H23;
// FR-063 de H23).
var skillsDelTrabajo = []string{"normativa", "jurisprudencia", "legal-core"}

// ajustesDelTrabajo son la concurrencia y el objetivo de duración que su
// include da a cada skill (FR-030 y FR-051 de H7.3; research.md D12 de H7.3).
// jurisprudencia va sin objetivo de duración, como legal-core (research.md D22
// de H23; FR-063 de H23), y, desde H25, a cuatro sesiones a la vez: con sus diez
// evals de una en una, el tope del trabajo no cubre su peor caso
// (contracts/job-de-evals.md §1 y §4 de H25; FR-070). legal-core va a dos desde
// antes de H9: con sus cinco evals de plazos, de una en una el tope tampoco
// cubriría su peor caso (entrada de H9, docs/roadmap/done/).
var ajustesDelTrabajo = map[string]AjustesDeSkill{
	"normativa":      {Concurrencia: 5, ObjetivoDeDuracion: 900},
	"jurisprudencia": {Concurrencia: 4, ObjetivoDeDuracion: 0},
	"legal-core":     {Concurrencia: 2, ObjetivoDeDuracion: 0},
}

// entornoDelTrabajo son las variables con las que su env pasa esos ajustes a
// scripts/evals.sh y, desde H30, el modo que decide la tanda, la propuesta y su
// base al paso de la decisión (contracts/job-de-evals.md §2 de H30).
var entornoDelTrabajo = map[string]string{
	"CONCURRENCIA_DE_EVALS":         "${{ matrix.concurrencia }}",
	"OBJETIVO_DE_DURACION_DE_EVALS": "${{ matrix.objetivo_de_duracion }}",
	variableDelModoDeEvals:          "${{ needs.tanda.outputs.modo }}",
	"PROPUESTA":                     "${{ github.event.pull_request.number }}",
	"BASE_DE_LA_PROPUESTA":          "${{ github.event.pull_request.base.sha }}",
}

// TestDefinicionDelJob comprueba en make ci la definición del job de evals
// (contracts/ejecucion-del-job.md §7 de H7.3; FR-030, FR-034, FR-035, FR-051 y
// FR-094 de H7.3; SC-008): la de .github/workflows/evals.yml es la del
// contrato; cada definición sintética que se aparta de él en una sola clave da
// una línea que nombra esa clave, el valor encontrado y el esperado; y una
// definición que no se puede leer, o un peor caso que no se puede obtener, es un
// error. El peor caso suma una tanda por cada grupo de sesiones —modo orden,
// modo herramienta y sin binario ni servidor—, y el tope de la definición lo
// cubre con las evals del repositorio (contracts/evals-en-dos-modos.md §6 de
// H21; FR-083 de H21). Comprueba también la tanda (contracts/tanda-del-job.md §4 de H7.4;
// FR-054, FR-070, FR-071 y FR-100 de H7.4; SC-010 de H7.4): el trabajo tanda y
// la dependencia del trabajo evals de él son los del contrato, un segundo
// disparo no mide mientras una ejecución anterior sin terminar mide, y la
// etiqueta sobre un commit cuya tanda terminó vuelve a medir. Y, desde H24, el
// juez con modelo y la ejecución de su medida (contracts/job-de-evals.md §4 y §5
// de H24; FR-050, FR-090, FR-091, FR-092 y FR-112 de H24; SC-012 de H24): el
// modelo del juez es un id completo y no es el que decide; la versión de Claude
// Code de sus votos se fija aparte de la de las sesiones y solo instala el del
// juez, en su prefijo; el trabajo medida repite las dos, solo corre con su
// etiqueta o con su entrada y no depende de la tanda, que no corre con ninguna
// de las dos; y cada tope cubre su peor caso, que en una skill con juez cuenta
// también sus votos. Y, desde H23, la tercera skill de la matriz
// (contracts/evals-jurisprudencia.md §6 de H23; FR-063 de H23): jurisprudencia
// está en ella, su include la deja sin objetivo de duración, y el tope del
// trabajo cubre el peor caso de las tres con las evals del repositorio. Y, desde
// H25 (contracts/job-de-evals.md §1, §3 y §4 de H25; FR-070, FR-072, FR-103;
// SC-003), su include la lleva a cuatro sesiones a la vez: con sus diez evals y
// de una en una, el tope ya no cubre su peor caso, y la definición que la deja
// así da la línea de su include y la del tope. Y, con su juez
// (contracts/job-de-evals.md §4 de H25; research D12 de H25; FR-050, FR-071), la
// matriz del trabajo medida: sus skills son las de la matriz del trabajo evals
// que tienen carpeta de juez, en su orden, cada una con la concurrencia de su
// trabajo evals —lo esperado se deriva de las evals, y no se escribe aquí—, de
// modo que la definición sin jurisprudencia en esa matriz, o con ella de una en
// una, da su línea; y el tope de cada trabajo cubre su peor caso, que en
// jurisprudencia cuenta ya los votos de su juez y los de sus casos etiquetados.
//
// Y, desde H30 (contracts/job-de-evals.md §1 a §3 de H30; FR-010, FR-012,
// FR-016 a FR-020; SC-011), la elección, la decisión y la reutilización: el
// trabajo cambios, con su tope, sus permisos de lectura, toda la historia en su
// checkout, TestEleccionDeLaPropuesta y la salida elige; la tanda, que solo
// corre con el despacho, con una de las etiquetas evals, evals-completa o
// evals-prueba-de-red o con elige=si, y cuyo modo es completa solo con el
// despacho y las dos últimas etiquetas, con su salida modo y la marca de la
// suite completa; y, en el trabajo evals, sus permisos, el modo, la propuesta y
// su base en el env, toda la historia en su checkout, el paso de la decisión
// tras instalar Go y antes de ejecutar las evals, el único con GH_TOKEN y
// GH_REPO, DECISION_DE_EVALS en el que las ejecuta y, el último, la marca del
// informe escrito, que corre salvo que se cancele la ejecución. La matriz da una
// sola combinación por skill, con 5, 4 y 2 sesiones a la vez, y 11 entre todas:
// la definición que admite más da su línea. Y el peor caso de cada trabajo suma
// el plazo del paso de la decisión, el -timeout de su go test, 15 minutos, que
// el tope cubre.
func TestDefinicionDelJob(t *testing.T) {
	t.Parallel()

	t.Run("del-repositorio", probarLaDefinicionDelRepositorio)
	t.Run("sinteticas", probarLasDefinicionesSinteticas)
	t.Run("errores", probarLosErroresDeLaDefinicion)
	t.Run("plazo-de-la-decision", probarElPlazoDeLaDecision)
	t.Run("peor-caso", probarElPeorCasoDelTrabajo)
	t.Run("tope-con-las-evals-del-repositorio", probarElTopeConLasEvalsDelRepositorio)
	t.Run("segundo-disparo", probarElSegundoDisparo)
	t.Run("estado-de-la-tanda", probarElEstadoDeLaTanda)
}

// TestJuezDeLaDefinicionDelJob fija de dónde se leen el modelo del juez y la
// versión de Claude Code de sus votos (contracts/job-de-evals.md §1 de H24;
// data-model §6 de H24; FR-090, FR-091): de MODELO_DEL_JUEZ y de
// VERSION_DE_CLAUDE_CODE_DEL_JUEZ del env del trabajo evals, tal cual. La
// versión de las sesiones, VERSION_DE_CLAUDE_CODE, va aparte y no es la de los
// votos; el modelo que decide no cambia; y una definición sin esas dos
// variables los deja vacíos. Las del trabajo medida, que las repite, no son las
// que se leen: en los dos casos se quedan como están.
func TestJuezDeLaDefinicionDelJob(t *testing.T) {
	t.Parallel()

	const (
		modeloDelJuez     = "claude-juez-9-8"
		versionDelJuez    = "9.8.7"
		modeloDelContrato = "claude-sonnet-5"
		entornoConOtros   = "      MODELO_DEL_JUEZ: " + modeloDelJuez + "\n" +
			"      VERSION_DE_CLAUDE_CODE_DEL_JUEZ: " + versionDelJuez + "\n"
	)

	casos := []struct {
		nombre  string
		entorno string
		modelo  string
		version string
	}{
		{nombre: "con-las-dos-variables", entorno: entornoConOtros, modelo: modeloDelJuez, version: versionDelJuez},
		{nombre: "sin-las-dos-variables"},
	}

	for _, caso := range casos {
		t.Run(caso.nombre, func(t *testing.T) {
			t.Parallel()

			contenido := strings.Replace(definicionDelContrato, trabajoDeEvalsDelContrato,
				strings.Replace(trabajoDeEvalsDelContrato, entornoDelJuezDelContrato, caso.entorno, 1), 1)
			require.NotEqual(t, definicionDelContrato, contenido, "premisa: el caso cambia el env del trabajo evals")

			ruta := filepath.Join(t.TempDir(), "evals.yml")
			require.NoError(t, os.WriteFile(ruta, []byte(contenido), 0o600))

			leida, err := leerDefinicionDelJob(ruta)
			require.NoError(t, err)

			require.Equal(t, versionDeLasSesionesDelContrato, leida.Env["VERSION_DE_CLAUDE_CODE"],
				"premisa: la definición fija la versión de las sesiones, que no es la de los votos")
			require.NotNil(t, leida.Medida, "premisa: la definición tiene el trabajo medida")
			require.Equal(t, modeloDelJuezDelContrato, leida.Medida.Env[variableDelModeloDelJuez],
				"premisa: el trabajo medida repite el modelo del juez del contrato")

			assert.Equal(t, caso.modelo, leida.ModeloDelJuez)
			assert.Equal(t, caso.version, leida.VersionDelJuez)
			assert.Equal(t, modeloDelContrato, leida.ModeloQueDecide)
		})
	}
}

// probarLaDefinicionDelRepositorio lee la definición real del job y falla, con
// una línea por clave, si no es la del contrato o si su tope no cubre el peor
// caso de alguna skill con las evals del repositorio.
func probarLaDefinicionDelRepositorio(t *testing.T) {
	t.Parallel()

	leida, err := leerDefinicionDelJob(rutaDeLaDefinicionDelJob)
	require.NoError(t, err)

	if fallos := comprobarLaDefinicion(leida, directorioDeEvals); len(fallos) > 0 {
		t.Fatalf("la definici\xc3\xb3n del job no es la de contracts/tanda-del-job.md \xc2\xa74 de H7.4, "+
			"contracts/ejecucion-del-job.md \xc2\xa77 de H7.3, contracts/job-de-evals.md \xc2\xa75 de H24 y "+
			"contracts/evals-jurisprudencia.md \xc2\xa76 de H23:\n%s",
			strings.Join(fallos, "\n"))
	}
}

// comprobarLaDefinicion devuelve una línea por cada clave del trabajo tanda y
// de la dependencia del trabajo evals de él que no es la de
// contracts/tanda-del-job.md §4 de H7.4, y por cada clave del trabajo evals que
// no es la de contracts/ejecucion-del-job.md §7 de H7.3, en el orden de los
// contratos, y una por cada skill de la matriz cuyo peor caso, con las evals de
// su carpeta dentro de evals, no cubre timeout-minutes o no se puede obtener.
// Detrás de las del env van las del juez con modelo, y al final, las del
// trabajo medida, la de la entrada del despacho que lo lanza y las de su tope,
// en el orden de contracts/job-de-evals.md §5 de H24; desde H25, detrás de las
// del trabajo medida van las de su matriz (contracts/job-de-evals.md §4 de H25).
// Desde H30, delante de todas van las del trabajo cambios, y detrás de las del
// env, las de los permisos y los pasos del trabajo evals (contracts/job-de-evals.md
// §2 de H30). Sin ninguna línea, la definición es la de los contratos.
func comprobarLaDefinicion(leida DefinicionDelJob, evals string) []string {
	var fallos []string

	fallos = append(fallos, fallosDeCambios(leida.cambios)...)
	fallos = append(fallos, fallosDeLaTanda(leida.Tanda)...)
	fallos = append(fallos, fallosDeLaDependencia(leida)...)
	fallos = append(fallos, fallosDeLaConcurrencia(leida)...)
	fallos = append(fallos, fallosDeLaMatriz(leida)...)
	fallos = append(fallos, fallosDelEntorno(leida)...)
	fallos = append(fallos, fallosDeUnMapa("jobs.evals.permissions", leida.permisos, permisosDelTrabajo)...)
	fallos = append(fallos, fallosDeLosPasosDelTrabajo(leida)...)
	fallos = append(fallos, fallosDelJuez(leida)...)
	fallos = append(fallos, fallosDelTope(leida, evals)...)
	fallos = append(fallos, fallosDeLaMedida(leida)...)
	fallos = append(fallos, fallosDeLaMatrizDeLaMedida(leida.Medida, evals)...)
	fallos = append(fallos, fallosDeLaEntrada(leida)...)
	fallos = append(fallos, fallosDelTopeDeLaMedida(leida.Medida, evals)...)

	return fallos
}

// fallosDeCambios comprueba el trabajo cambios de contracts/job-de-evals.md §2
// de H30: que está; su if, que solo lo corre al abrir o reabrir una propuesta;
// su tope; sus permisos; su salida elige; toda la historia en su checkout; y su
// paso de id elegir, con su name, su env y su run. Sin el trabajo, solo esa
// línea.
func fallosDeCambios(cambios *trabajoDelFlujo) []string {
	if cambios == nil {
		return []string{fallo("jobs.cambios", noEsta, "el trabajo que elige las evals que la propuesta ha podido cambiar")}
	}

	var fallos []string

	if cambios.Condicion != condicionDeCambios {
		fallos = append(fallos, fallo("jobs.cambios.if", presentarTexto(cambios.Condicion),
			strconv.Quote(condicionDeCambios)))
	}

	if cambios.TopeEnMinutos != topeDeCambios {
		fallos = append(fallos, fallo("jobs.cambios.timeout-minutes", fmt.Sprintf("vale %d", cambios.TopeEnMinutos),
			strconv.Itoa(topeDeCambios)))
	}

	fallos = append(fallos, fallosDeUnMapa("jobs.cambios.permissions", cambios.Permisos, permisosDeCambios)...)

	if salida := cambios.Salidas["elige"]; salida != salidaDeCambios {
		fallos = append(fallos, fallo("jobs.cambios.outputs.elige", presentarTexto(salida), strconv.Quote(salidaDeCambios)))
	}

	fallos = append(fallos, fallosDelCheckout("jobs.cambios.steps", cambios.Pasos)...)

	var elige pasoDelFlujo
	if indice := slices.IndexFunc(cambios.Pasos, func(paso pasoDelFlujo) bool { return paso.ID == "elegir" }); indice >= 0 {
		elige = cambios.Pasos[indice]
	}

	const delPaso = "jobs.cambios.steps, id elegir, "

	if elige.Nombre != pasoDeLaEleccion {
		fallos = append(fallos, fallo(delPaso+"name", presentarTexto(elige.Nombre), strconv.Quote(pasoDeLaEleccion)))
	}

	fallos = append(fallos, fallosDeUnMapa(delPaso+"env", elige.Env, entornoDeLaEleccion)...)

	if elige.Orden != ordenDeLaEleccion {
		fallos = append(fallos, fallo(delPaso+"run", presentarTexto(elige.Orden), strconv.Quote(ordenDeLaEleccion)))
	}

	return fallos
}

// fallosDeUnMapa devuelve, con esa clave, la línea de un mapa de la definición
// —unos permissions, el env de un paso— que no es exactamente el esperado: ni
// una clave más ni una menos, cada una con su valor.
func fallosDeUnMapa(clave string, encontrado, esperado map[string]string) []string {
	if maps.Equal(encontrado, esperado) {
		return nil
	}

	presentado := noEsta
	if encontrado != nil {
		presentado = "vale " + presentarMapa(encontrado)
	}

	return []string{fallo(clave, presentado, presentarMapa(esperado))}
}

// fallosDelCheckout comprueba que el primer paso de los dados que obtiene el
// código, el de accionDelCheckout, trae toda la historia, de la que la elección
// y la decisión sacan el árbol de la base (research D5 de H30), con la clave
// de esos pasos.
func fallosDelCheckout(clave string, pasos []pasoDelFlujo) []string {
	clave += ", el de " + accionDelCheckout + ", with.fetch-depth"

	indice := slices.IndexFunc(pasos, func(paso pasoDelFlujo) bool { return paso.Accion == accionDelCheckout })
	if indice < 0 {
		return []string{fallo(clave, "el paso "+noEsta, todaLaHistoria)}
	}

	if profundidad := pasos[indice].Con["fetch-depth"]; profundidad != todaLaHistoria {
		return []string{fallo(clave, presentarTexto(profundidad), todaLaHistoria)}
	}

	return nil
}

// fallosDeLosPasosDelTrabajo comprueba los pasos del trabajo evals de
// contracts/job-de-evals.md §2 de H30: toda la historia en su checkout; el paso
// de la decisión, que está entre el que instala Go y el que ejecuta las evals,
// con su run y con GH_TOKEN y GH_REPO en su env, que no están en el del
// trabajo ni en el de ningún otro paso; el run del que ejecuta las evals; y el
// name, el if y el run del último, la marca del informe escrito.
func fallosDeLosPasosDelTrabajo(leida DefinicionDelJob) []string {
	const clave = "jobs.evals.steps"

	decide := slices.IndexFunc(leida.pasos, func(paso pasoDelFlujo) bool { return paso.Nombre == pasoDeLaDecision })

	fallos := slices.Concat(fallosDelCheckout(clave, leida.pasos), fallosDeLaDecision(leida.pasos, decide),
		fallosDeGhFueraDeLaDecision(leida, decide))

	var ejecucion, ultimo pasoDelFlujo
	if ejecuta := slices.IndexFunc(leida.pasos, func(paso pasoDelFlujo) bool {
		return paso.Nombre == pasoQueEjecutaLasEvals
	}); ejecuta >= 0 {
		ejecucion = leida.pasos[ejecuta]
	}

	if len(leida.pasos) > 0 {
		ultimo = leida.pasos[len(leida.pasos)-1]
	}

	elUltimo := clave + ", el \xc3\xbaltimo, "

	for _, comprobada := range []struct{ clave, encontrado, esperado string }{
		{clave + ", " + strconv.Quote(pasoQueEjecutaLasEvals) + ", run", ejecucion.Orden, ordenQueEjecutaLasEvals},
		{elUltimo + "name", ultimo.Nombre, marcaDelInformeEscrito},
		{elUltimo + "if", ultimo.Condicion, condicionDeLaMarcaDelInforme},
		{elUltimo + "run", ultimo.Orden, ordenDeLaMarcaDelInforme},
	} {
		if comprobada.encontrado != comprobada.esperado {
			fallos = append(fallos, fallo(comprobada.clave, presentarTexto(comprobada.encontrado),
				strconv.Quote(comprobada.esperado)))
		}
	}

	return fallos
}

// fallosDeLaDecision comprueba el paso de la decisión del trabajo evals, el de
// la posición decide entre sus pasos (-1, ninguno): que está; que va tras el que
// instala Go y antes del que ejecuta las evals; y su env y su run.
func fallosDeLaDecision(pasos []pasoDelFlujo, decide int) []string {
	delPaso := "jobs.evals.steps, " + strconv.Quote(pasoDeLaDecision)

	if decide < 0 {
		return []string{fallo(delPaso, noEsta, "el paso que decide qu\xc3\xa9 mide y qu\xc3\xa9 reutiliza")}
	}

	var fallos []string

	instalaGo := slices.IndexFunc(pasos, func(paso pasoDelFlujo) bool { return paso.Accion == accionDeGo })
	ejecuta := slices.IndexFunc(pasos, func(paso pasoDelFlujo) bool { return paso.Nombre == pasoQueEjecutaLasEvals })

	if instalaGo < 0 || decide < instalaGo || ejecuta < decide {
		fallos = append(fallos, fallo(delPaso, "est\xc3\xa1 en el paso "+strconv.Itoa(decide+1),
			"que vaya tras el de "+accionDeGo+" y antes de "+strconv.Quote(pasoQueEjecutaLasEvals)))
	}

	decision := pasos[decide]

	fallos = append(fallos, fallosDeUnMapa(delPaso+", env", decision.Env, entornoDeLaDecision)...)

	if decision.Orden != ordenDeLaDecision {
		fallos = append(fallos, fallo(delPaso+", run", presentarTexto(decision.Orden), strconv.Quote(ordenDeLaDecision)))
	}

	return fallos
}

// fallosDeGhFueraDeLaDecision devuelve una línea por cada sitio del trabajo
// evals, aparte del paso de la decisión, cuyo env lleva GH_TOKEN o GH_REPO: el
// env del trabajo y el de cada uno de sus otros pasos, en su orden. El de la
// posición decide es el de la decisión; -1, ninguno.
func fallosDeGhFueraDeLaDecision(leida DefinicionDelJob, decide int) []string {
	type sitio struct {
		clave string
		env   map[string]string
	}

	var fallos []string

	sitios := []sitio{{clave: "jobs.evals.env", env: leida.Env}}

	for posicion, paso := range leida.pasos {
		if posicion != decide {
			sitios = append(sitios, sitio{clave: "jobs.evals.steps, " + strconv.Quote(paso.Nombre) + ", env", env: paso.Env})
		}
	}

	for _, sitio := range sitios {
		for _, variable := range []string{variableDelAccesoDeGh, variableDelRepoDeGh} {
			if _, esta := sitio.env[variable]; esta {
				fallos = append(fallos, fallo(sitio.clave+"."+variable, "est\xc3\xa1",
					queNoEste+": solo el paso de la decisi\xc3\xb3n lee la plataforma"))
			}
		}
	}

	return fallos
}

// fallosDeLaTanda comprueba el trabajo tanda de contracts/tanda-del-job.md §4
// de H7.4: que está; que un name no cambia el nombre con el que gh lo da, su
// id, que es el que busca la decisión; su if; que no tiene concurrency; su
// permiso de actions; su salida medir; y sus pasos. Desde H30, también el modo
// de su env y su salida modo (contracts/job-de-evals.md §1 y §2 de H30). Sin el
// trabajo, solo esa línea.
func fallosDeLaTanda(tanda *TrabajoDeLaTanda) []string {
	if tanda == nil {
		return []string{fallo("jobs."+trabajoDeLaTanda, noEsta,
			"el trabajo que decide si la ejecuci\xc3\xb3n mide el commit")}
	}

	var fallos []string

	if tanda.Nombre != "" && tanda.Nombre != trabajoDeLaTanda {
		fallos = append(fallos, fallo("jobs.tanda.name", presentarTexto(tanda.Nombre),
			queNoEste+": gh da el trabajo por su id, tanda, que es por el que lo busca la decisi\xc3\xb3n"))
	}

	if tanda.Condicion != condicionDeLaTanda {
		fallos = append(fallos, fallo("jobs.tanda.if", presentarTexto(tanda.Condicion),
			strconv.Quote(condicionDeLaTanda)))
	}

	if tanda.ConConcurrencia {
		fallos = append(fallos, fallo("jobs.tanda.concurrency", "est\xc3\xa1",
			queNoEste+": la tanda ni espera ni se cancela, y no deja una comprobaci\xc3\xb3n roja por decidir no medir"))
	}

	if permiso := tanda.Permisos["actions"]; permiso != permisoDeLasEjecuciones {
		fallos = append(fallos, fallo("jobs.tanda.permissions.actions", presentarTexto(permiso),
			strconv.Quote(permisoDeLasEjecuciones)))
	}

	if modo := tanda.env[variableDelModoDeEvals]; modo != modoDeLaTanda {
		fallos = append(fallos, fallo("jobs.tanda.env."+variableDelModoDeEvals, presentarTexto(modo),
			strconv.Quote(modoDeLaTanda)))
	}

	for _, salida := range []struct{ nombre, esperada string }{
		{"medir", salidaDeLaTanda},
		{"modo", salidaDelModoDeLaTanda},
	} {
		if encontrada := tanda.Salidas[salida.nombre]; encontrada != salida.esperada {
			fallos = append(fallos, fallo("jobs.tanda.outputs."+salida.nombre, presentarTexto(encontrada),
				strconv.Quote(salida.esperada)))
		}
	}

	return append(fallos, fallosDeLosPasosDeLaTanda(tanda.Pasos)...)
}

// fallosDeLosPasosDeLaTanda comprueba los pasos del trabajo tanda de
// contracts/tanda-del-job.md §4 de H7.4: el run del paso que decide, el de id
// decidir, y el name y el if de las marcas que leen las ejecuciones
// posteriores, que desde H30 son las dos últimas: la penúltima, la de que mide
// el commit, y la última, la de que mide la suite completa
// (contracts/job-de-evals.md §2 de H30; research D13 de H30).
func fallosDeLosPasosDeLaTanda(pasos []PasoDelTrabajo) []string {
	var fallos []string

	var decide, penultimo, ultimo PasoDelTrabajo

	if indice := slices.IndexFunc(pasos, func(paso PasoDelTrabajo) bool { return paso.ID == "decidir" }); indice >= 0 {
		decide = pasos[indice]
	}

	if len(pasos) > 1 {
		penultimo = pasos[len(pasos)-2]
	}

	if len(pasos) > 0 {
		ultimo = pasos[len(pasos)-1]
	}

	if decide.Orden != ordenDelPasoQueDecide {
		fallos = append(fallos, fallo("jobs.tanda.steps, id decidir, run", presentarTexto(decide.Orden),
			strconv.Quote(ordenDelPasoQueDecide)))
	}

	for _, comprobado := range []struct {
		clave             string
		paso              PasoDelTrabajo
		nombre, condicion string
	}{
		{"jobs.tanda.steps, el pen\xc3\xbaltimo, ", penultimo, marcaDeLaTanda, condicionDeLaMarca},
		{"jobs.tanda.steps, el \xc3\xbaltimo, ", ultimo, marcaDeLaSuiteCompleta, condicionDeLaMarcaDeLaSuiteCompleta},
	} {
		if comprobado.paso.Nombre != comprobado.nombre {
			fallos = append(fallos, fallo(comprobado.clave+"name", presentarTexto(comprobado.paso.Nombre),
				strconv.Quote(comprobado.nombre)))
		}

		if comprobado.paso.Condicion != comprobado.condicion {
			fallos = append(fallos, fallo(comprobado.clave+"if", presentarTexto(comprobado.paso.Condicion),
				strconv.Quote(comprobado.condicion)))
		}
	}

	return fallos
}

// fallosDeLaDependencia comprueba la dependencia del trabajo evals de la tanda
// de contracts/tanda-del-job.md §4 de H7.4: su needs y su if.
func fallosDeLaDependencia(leida DefinicionDelJob) []string {
	var fallos []string

	if !slices.Equal(leida.Dependencias, dependenciasDelTrabajo) {
		encontradas := noEsta
		if leida.Dependencias != nil {
			encontradas = "vale " + presentarLista(leida.Dependencias)
		}

		fallos = append(fallos, fallo("jobs.evals.needs", encontradas, presentarLista(dependenciasDelTrabajo)))
	}

	if leida.Condicion != condicionDelTrabajo {
		fallos = append(fallos, fallo("jobs.evals.if", presentarTexto(leida.Condicion), strconv.Quote(condicionDelTrabajo)))
	}

	return fallos
}

// fallosDeLaConcurrencia comprueba §7.1: el group y cancel-in-progress del
// trabajo, y que el flujo no tenga concurrency propio.
func fallosDeLaConcurrencia(leida DefinicionDelJob) []string {
	var fallos []string

	if leida.Grupo != grupoDelTrabajo {
		fallos = append(fallos, fallo("jobs.evals.concurrency.group", presentarTexto(leida.Grupo),
			strconv.Quote(grupoDelTrabajo)))
	}

	switch {
	case leida.CancelaLaEnCurso == nil:
		fallos = append(fallos, fallo("jobs.evals.concurrency.cancel-in-progress", noEsta, "false"))
	case *leida.CancelaLaEnCurso:
		fallos = append(fallos, fallo("jobs.evals.concurrency.cancel-in-progress", "vale true", "false"))
	}

	if leida.ConcurrenciaDeFlujo {
		fallos = append(fallos, fallo("concurrency", "est\xc3\xa1 en el flujo",
			queNoEste+": la tanda por commit la da el del trabajo evals, por skill"))
	}

	return fallos
}

// fallosDeLaMatriz comprueba §7.2 y la primera mitad de §7.3: el name del
// trabajo, las skills de su matriz y lo que include da a cada una.
func fallosDeLaMatriz(leida DefinicionDelJob) []string {
	var fallos []string

	if leida.Nombre != nombreDelTrabajo {
		fallos = append(fallos, fallo("jobs.evals.name", presentarTexto(leida.Nombre),
			strconv.Quote(nombreDelTrabajo)))
	}

	if !slices.Equal(leida.Skills, skillsDelTrabajo) {
		fallos = append(fallos, fallo("jobs.evals.strategy.matrix.skill", "vale "+presentarLista(leida.Skills),
			presentarLista(skillsDelTrabajo)))
	}

	fallos = append(fallos,
		fallosDelInclude("jobs.evals.strategy.matrix.include", leida.PorSkill, ajustesDelTrabajo, presentarAjustes)...)

	return append(fallos, fallosDeLasSesionesALaVez(leida)...)
}

// fallosDeLasSesionesALaVez comprueba lo que la matriz del trabajo evals deja
// abrir a la vez (FR-020 y SC-011 de H30): una sola combinación por skill
// —ninguna clave de la matriz además de skill e include, que multiplicaría sus
// trabajos— y, entre todas sus combinaciones, como mucho
// sesionesALaVezComoMucho sesiones, la suma de la concurrencia de cada una: la
// de cada skill de la matriz y la de cada skill que solo está en include, que
// es otra combinación.
func fallosDeLasSesionesALaVez(leida DefinicionDelJob) []string {
	var fallos []string

	for _, dimension := range leida.dimensionesDeLaMatriz {
		fallos = append(fallos, fallo("jobs.evals.strategy.matrix."+dimension, "est\xc3\xa1",
			queNoEste+": la matriz da un solo trabajo por skill"))
	}

	combinaciones := slices.Clone(leida.Skills)
	for skill := range leida.PorSkill {
		if !slices.Contains(combinaciones, skill) {
			combinaciones = append(combinaciones, skill)
		}
	}

	aLaVez := 0
	for _, skill := range combinaciones {
		aLaVez += leida.PorSkill[skill].Concurrencia
	}

	if aLaVez > sesionesALaVezComoMucho {
		fallos = append(fallos, fallo("jobs.evals.strategy.matrix",
			fmt.Sprintf("admite %d sesiones a la vez entre sus %d combinaciones", aLaVez, len(combinaciones)),
			fmt.Sprintf("como mucho %d", sesionesALaVezComoMucho)))
	}

	return fallos
}

// fallosDelInclude devuelve una línea, con esa clave y la skill, por cada skill
// cuya entrada del include de la matriz de un trabajo no es la esperada, en
// orden de skill: la que no tiene lo esperado, la que falta y la que sobra.
func fallosDelInclude[T comparable](clave string, porSkill, esperadosPorSkill map[string]T, presentar func(T) string) []string {
	var fallos []string

	skills := slices.Sorted(maps.Keys(porSkill))
	for skill := range esperadosPorSkill {
		if !slices.Contains(skills, skill) {
			skills = append(skills, skill)
		}
	}

	slices.Sort(skills)

	for _, skill := range skills {
		encontrados, estan := porSkill[skill]
		esperados, seEsperan := esperadosPorSkill[skill]

		if estan == seEsperan && encontrados == esperados {
			continue
		}

		esperado := queNoEste
		if seEsperan {
			esperado = presentar(esperados)
		}

		encontrado := noEsta
		if estan {
			encontrado = "vale " + presentar(encontrados)
		}

		fallos = append(fallos, fallo(clave+", skill "+skill, encontrado, esperado))
	}

	return fallos
}

// fallosDelEntorno comprueba la segunda mitad de §7.3: las variables con las que
// el env del trabajo pasa la concurrencia y el objetivo de su skill.
func fallosDelEntorno(leida DefinicionDelJob) []string {
	var fallos []string

	for _, nombre := range slices.Sorted(maps.Keys(entornoDelTrabajo)) {
		esperado := entornoDelTrabajo[nombre]

		encontrado, esta := leida.Env[nombre]
		if esta && encontrado == esperado {
			continue
		}

		presentado := noEsta
		if esta {
			presentado = "vale " + strconv.Quote(encontrado)
		}

		fallos = append(fallos, fallo("jobs.evals.env."+nombre, presentado, strconv.Quote(esperado)))
	}

	return fallos
}

// fallosDelJuez comprueba las dos primeras filas de contracts/job-de-evals.md §5
// de H24: el modelo del juez del env del trabajo evals está, tiene la forma de
// un id completo, que un alias no tiene, y no es el modelo que decide; y la
// versión de Claude Code de sus votos está, es <n>.<n>.<n> y es la del Claude
// Code que el paso de instalación instala en su prefijo, y de ningún otro.
func fallosDelJuez(leida DefinicionDelJob) []string {
	var fallos []string

	if modelo := leida.ModeloDelJuez; !formaDelIDCompleto.MatchString(modelo) || modelo == leida.ModeloQueDecide {
		fallos = append(fallos, fallo("jobs.evals.env."+variableDelModeloDelJuez, presentarTexto(modelo),
			modeloDelJuezEsperado))
	}

	if version := leida.VersionDelJuez; !formaDeLaVersion.MatchString(version) {
		fallos = append(fallos, fallo("jobs.evals.env."+variableDeLaVersionDelJuez, presentarTexto(version),
			versionDelJuezEsperada))
	}

	return append(fallos, fallosDeLaInstalacion(leida.OrdenDeInstalacion)...)
}

// fallosDeLaInstalacion comprueba el run del paso del trabajo evals que instala
// Claude Code (contracts/job-de-evals.md §2 y §5 de H24), línea a línea y sin
// su sangría: una de ellas instala el del juez en su prefijo, y ninguna otra
// instala Claude Code en la versión del juez, que sería instalar con ella el de
// las sesiones.
func fallosDeLaInstalacion(orden string) []string {
	var (
		fallos []string
		lineas []string
	)

	for linea := range strings.SplitSeq(orden, "\n") {
		lineas = append(lineas, strings.TrimSpace(linea))
	}

	if !slices.Contains(lineas, instalacionDelJuez) {
		fallos = append(fallos, fallo(pasoDeLaInstalacion, presentarTexto(orden),
			"que tenga la l\xc3\xadnea "+strconv.Quote(instalacionDelJuez)))
	}

	for _, linea := range lineas {
		if linea != instalacionDelJuez && strings.Contains(linea, claudeCodeDelJuez) {
			fallos = append(fallos, fallo(pasoDeLaInstalacion, "tiene la l\xc3\xadnea "+strconv.Quote(linea),
				"que con "+variableDeLaVersionDelJuez+" solo instale el del juez, en su prefijo"))
		}
	}

	return fallos
}

// fallosDelTope comprueba §7.4: para cada skill de la matriz, timeout-minutes
// cubre su peor caso, que nombra con sus términos. Desde H24, el de una skill
// con juez lleva también los de sus votos (contracts/job-de-evals.md §4 de H24).
func fallosDelTope(leida DefinicionDelJob, evals string) []string {
	return fallosDeUnTope("jobs.evals.timeout-minutes", leida.TopeEnMinutos, leida.Skills,
		func(skill string) (peorCasoConTerminos, error) { return leida.peorCaso(evals, skill) })
}

// fallosDeLaMedida comprueba el trabajo medida de contracts/job-de-evals.md §3
// y §5 de H24: que está; que su env repite el modelo del juez y la versión de
// Claude Code de sus votos del env del trabajo evals; su if, que es exactamente
// el de su etiqueta o su entrada; y que no tiene needs. Sin el trabajo, solo esa
// línea.
func fallosDeLaMedida(leida DefinicionDelJob) []string {
	medida := leida.Medida
	if medida == nil {
		return []string{fallo("jobs.medida", noEsta, "el trabajo que ejecuta la medida del juez")}
	}

	var fallos []string

	fijadas := []struct {
		variable string
		enEvals  string
	}{
		{variable: variableDelModeloDelJuez, enEvals: leida.ModeloDelJuez},
		{variable: variableDeLaVersionDelJuez, enEvals: leida.VersionDelJuez},
	}

	for _, fijada := range fijadas {
		if enLaMedida := medida.Env[fijada.variable]; enLaMedida != fijada.enEvals {
			fallos = append(fallos, fallo("jobs.medida.env."+fijada.variable, presentarTexto(enLaMedida),
				"lo de jobs.evals.env."+fijada.variable+", que "+presentarTexto(fijada.enEvals)))
		}
	}

	if medida.Condicion != condicionDeLaMedida {
		fallos = append(fallos, fallo("jobs.medida.if", presentarTexto(medida.Condicion),
			strconv.Quote(condicionDeLaMedida)))
	}

	if medida.ConDependencias {
		fallos = append(fallos, fallo("jobs.medida.needs", "est\xc3\xa1",
			queNoEste+": la medida no abre sesiones de evals, y no depende de la tanda que las decide"))
	}

	return fallos
}

// fallosDeLaMatrizDeLaMedida comprueba la matriz del trabajo medida
// (contracts/job-de-evals.md §4 de H25; research D12 de H25; FR-050, FR-071,
// FR-103): sus skills son las de la matriz del trabajo evals del contrato que
// tienen juez, con las evals de su carpeta dentro de evals, en su orden; y su
// include da a cada una la concurrencia de su trabajo evals, la del contrato, y
// a ninguna más. Lo esperado se deriva, y no se escribe aquí: una skill con juez
// que falte en la medida da su línea. Desde H31, detrás de ellas va la entrada
// descripciones, la de la medida del juez de las descripciones, a cinco casos
// a la vez (contracts/juez-de-las-descripciones.md §7 de H31; research D9 de
// H31; FR-034). Sin el trabajo no hay matriz que comprobar: su línea la da
// fallosDeLaMedida.
func fallosDeLaMatrizDeLaMedida(medida *TrabajoDeLaMedida, evals string) []string {
	if medida == nil {
		return nil
	}

	const clave = "jobs.medida.strategy.matrix."

	conJuez, err := skillsConJuez(evals)
	if err != nil {
		return []string{fmt.Sprintf("%sskill: las skills con juez no se pueden obtener: %v", clave, err)}
	}

	esperadas := append(slices.Clone(conJuez), entradaDeLasDescripciones)

	var fallos []string

	if !slices.Equal(medida.Skills, esperadas) {
		fallos = append(fallos, fallo(clave+"skill", "vale "+presentarLista(medida.Skills), presentarLista(esperadas)))
	}

	concurrencias := make(map[string]int, len(esperadas))
	for _, skill := range conJuez {
		concurrencias[skill] = ajustesDelTrabajo[skill].Concurrencia
	}

	concurrencias[entradaDeLasDescripciones] = concurrenciaDeLaMedidaDeLasDescripciones

	return append(fallos, fallosDelInclude(clave+"include", medida.ConcurrenciaPorSkill, concurrencias, presentarConcurrencia)...)
}

// concurrenciaDeLaMedidaDeLasDescripciones es la de la entrada descripciones del
// trabajo medida, cuántos casos de la medida del juez de las descripciones se
// votan a la vez como mucho: cinco, como normativa (research D9 de H31).
const concurrenciaDeLaMedidaDeLasDescripciones = 5

// skillsConJuez son las skills de la matriz del trabajo evals del contrato que
// tienen juez, en su orden: aquellas cuyas evals, las de su carpeta dentro de
// evals, se leen con la carpeta del juez (research D12 de H25).
func skillsConJuez(evals string) ([]string, error) {
	var conJuez []string

	for _, skill := range skillsDelTrabajo {
		conjunto, err := LeerConjunto(filepath.Join(evals, skill))
		if err != nil {
			return nil, err
		}

		if conjunto.Juez != nil {
			conJuez = append(conJuez, skill)
		}
	}

	return conJuez, nil
}

// fallosDeLaEntrada comprueba la entrada del despacho que lanza la medida del
// juez (contracts/job-de-evals.md §3 y §5 de H24): que está y que su valor por
// omisión es false, de modo que un despacho que no la da mide la skill, como
// hasta ahora.
func fallosDeLaEntrada(leida DefinicionDelJob) []string {
	const clave = "on.workflow_dispatch.inputs." + entradaDeLaMedida

	entrada, esta := leida.EntradasDelDespacho[entradaDeLaMedida]
	if !esta {
		return []string{fallo(clave, noEsta, "la entrada que lanza la medida del juez, con default: false")}
	}

	if porOmision, esBooleano := entrada.PorOmision.(bool); esBooleano && !porOmision {
		return nil
	}

	encontrado := noEsta
	if entrada.PorOmision != nil {
		encontrado = fmt.Sprintf("vale %#v", entrada.PorOmision)
	}

	return []string{fallo(clave+".default", encontrado, "false")}
}

// fallosDelTopeDeLaMedida comprueba que el timeout-minutes del trabajo medida
// cubre, para cada skill de su matriz, el peor caso de su medida, que nombra
// con sus términos (contracts/job-de-evals.md §4 de H24; FR-092 de H24). Sin el
// trabajo no hay tope que comprobar: su línea la da fallosDeLaMedida.
func fallosDelTopeDeLaMedida(medida *TrabajoDeLaMedida, evals string) []string {
	if medida == nil {
		return nil
	}

	return fallosDeUnTope("jobs.medida.timeout-minutes", medida.TopeEnMinutos, medida.Skills,
		func(skill string) (peorCasoConTerminos, error) { return medida.peorCaso(evals, skill) })
}

// peorCasoConTerminos es el peor caso de un trabajo: lo que dura y sus
// términos.
type peorCasoConTerminos interface {
	Duracion() time.Duration
	fmt.Stringer
}

// fallosDeUnTope devuelve una línea, con esa clave, por cada skill cuyo peor
// caso no cubre el tope en minutos de su trabajo, con el peor caso y sus
// términos, o no se puede obtener, con el porqué.
func fallosDeUnTope(clave string, minutos int, skills []string,
	peorCaso func(skill string) (peorCasoConTerminos, error),
) []string {
	var fallos []string

	tope := time.Duration(minutos) * time.Minute

	for _, skill := range skills {
		peor, err := peorCaso(skill)
		if err != nil {
			fallos = append(fallos, fmt.Sprintf("%s: el peor caso de %s no se puede obtener: %v", clave, skill, err))

			continue
		}

		if tope < peor.Duracion() {
			fallos = append(fallos, fallo(clave, fmt.Sprintf("vale %d (%d s)", minutos, int(tope/time.Second)),
				fmt.Sprintf("al menos el peor caso de %s, %s", skill, peor)))
		}
	}

	return fallos
}

// fallo es la línea de una clave cuyo valor no es el esperado.
func fallo(clave, encontrado, esperado string) string {
	return fmt.Sprintf("%s: %s, y lo esperado es %s", clave, encontrado, esperado)
}

// presentarTexto presenta el texto encontrado en una clave: entre comillas o,
// vacío, que la clave no está.
func presentarTexto(texto string) string {
	if texto == "" {
		return noEsta
	}

	return "vale " + strconv.Quote(texto)
}

// presentarLista presenta una lista de YAML en flujo, como [a, b].
func presentarLista(elementos []string) string {
	return "[" + strings.Join(elementos, ", ") + "]"
}

// presentarMapa presenta un mapa de YAML en flujo, con sus claves en orden,
// como {a: b, c: d}.
func presentarMapa(mapa map[string]string) string {
	pares := make([]string, 0, len(mapa))
	for _, clave := range slices.Sorted(maps.Keys(mapa)) {
		pares = append(pares, clave+": "+mapa[clave])
	}

	return "{" + strings.Join(pares, ", ") + "}"
}

// presentarAjustes presenta los ajustes de una skill como las claves de su
// entrada de include.
func presentarAjustes(ajustes AjustesDeSkill) string {
	return fmt.Sprintf("{concurrencia: %d, objetivo_de_duracion: %d}", ajustes.Concurrencia,
		ajustes.ObjetivoDeDuracion)
}

// presentarConcurrencia presenta la concurrencia de una skill como la clave de
// su entrada de include del trabajo medida.
func presentarConcurrencia(concurrencia int) string {
	return fmt.Sprintf("{concurrencia: %d}", concurrencia)
}

// trabajoDeLaTandaDelContrato es el trabajo tanda de definicionDelContrato, el
// de contracts/tanda-del-job.md §1 de H7.4 sin sus comentarios ni los pasos que
// no se comprueban, con lo que le añade contracts/job-de-evals.md §2 de H30: el
// modo en su env, la salida modo y la marca de la suite completa.
const trabajoDeLaTandaDelContrato = `  tanda:
    needs: [cambios]
    if: >-
      ` + sinCancelar + ` && needs.cambios.result != 'failure' && (
        ` + despachoSinLaMedida + ` ||
        github.event.label.name == 'evals' ||
        github.event.label.name == 'evals-completa' ||
        github.event.label.name == 'evals-prueba-de-red' ||
        needs.cambios.outputs.elige == 'si'
      )
    runs-on: ubuntu-24.04
    timeout-minutes: 15
    permissions:
      contents: read
      actions: read
    env:
      MODO_DE_EVALS: >-
        ${{ (github.event_name == 'workflow_dispatch' || github.event.label.name == 'evals-completa' ||
        github.event.label.name == 'evals-prueba-de-red') && 'completa' || 'reutilizacion' }}
    outputs:
      medir: ${{ steps.decidir.outputs.medir }}
      modo: ${{ steps.decidir.outputs.modo }}
    steps:
      - name: Mirar si otra tanda mide este commit
        id: decidir
        run: >-
          go test -tags evals -count=1 -timeout 12m -v -run '^TestTandaDelCommit$' ./internal/evals/
          -args -commit "$COMMIT_EVALUADO" -ejecucion "$EJECUCION" -modo "$MODO_DE_EVALS" -salida "$GITHUB_OUTPUT"
      - name: ` + pasoDeLaMarca + `
        if: steps.decidir.outputs.medir == 'si'
        run: echo mide
      - name: ` + pasoDeLaSuiteCompleta + `
        if: steps.decidir.outputs.medir == 'si' && steps.decidir.outputs.modo == 'completa'
        run: echo mide la suite completa
`

// trabajoDeCambiosDelContrato es el trabajo cambios de definicionDelContrato, el
// de contracts/job-de-evals.md §2 de H30 sin el paso que instala Go, que no se
// comprueba. Va el último de la definición: así un fragmento que está también
// en la tanda, como su tope, se cambia en la tanda, y los cambios de este
// trabajo los hace enCambios.
const trabajoDeCambiosDelContrato = `  cambios:
    if: ` + condicionDeCambios + `
    runs-on: ubuntu-24.04
    timeout-minutes: 15
    permissions:
      contents: read
      pull-requests: read
    outputs:
      elige: ${{ steps.elegir.outputs.elige }}
    steps:
      - name: Obtener la cabeza de la propuesta
        uses: ` + accionDelCheckout + `
        with:
          ref: ${{ github.event.pull_request.head.sha }}
          fetch-depth: 0
      - name: ` + pasoDeLaEleccion + `
        id: elegir
        env:
          GH_TOKEN: ${{ secrets.GITHUB_TOKEN }}
          GH_REPO: ${{ github.repository }}
          PROPUESTA: ${{ github.event.pull_request.number }}
          BASE_DE_LA_PROPUESTA: ${{ github.event.pull_request.base.sha }}
        run: >-
          go test -tags evals -count=1 -timeout 12m -v -run '^TestEleccionDeLaPropuesta$' ./internal/evals/
          -args -propuesta "$PROPUESTA" -base "$BASE_DE_LA_PROPUESTA" -salida "$GITHUB_OUTPUT"
`

// enCambios es el cambio de un fragmento por otro dentro del trabajo cambios de
// definicionDelContrato, también si el fragmento está antes en otro trabajo.
func enCambios(antes, despues string) cambioDeLaDefinicion {
	return cambioDeLaDefinicion{
		antes:   trabajoDeCambiosDelContrato,
		despues: strings.Replace(trabajoDeCambiosDelContrato, antes, despues, 1),
	}
}

// Lo que definicionDelContrato fija para el juez con modelo: su modelo y la
// versión de Claude Code de sus votos, las dos líneas con las que los repiten
// el env del trabajo evals y el del trabajo medida, y la versión de las
// sesiones, que va aparte.
const (
	modeloDelJuezDelContrato        = "claude-opus-5-5"
	versionDelJuezDelContrato       = "2.1.289"
	versionDeLasSesionesDelContrato = "2.1.284"

	modeloDelJuezEnElEnv  = "      MODELO_DEL_JUEZ: " + modeloDelJuezDelContrato + "\n"
	versionDelJuezEnElEnv = "      VERSION_DE_CLAUDE_CODE_DEL_JUEZ: " + versionDelJuezDelContrato + "\n"

	entornoDelJuezDelContrato = modeloDelJuezEnElEnv + versionDelJuezEnElEnv
)

// Las líneas de definicionDelContrato que cambian las definiciones sintéticas
// del juez y de su medida: las dos del paso de instalación que instalan un
// Claude Code, la del if del trabajo medida y las tres de la entrada del
// despacho que la lanza.
const (
	instalacionDeLasSesiones = `npm install -g "` + paqueteDeClaudeCode + `@${VERSION_DE_CLAUDE_CODE}"`

	lineaDeLaInstalacionDelJuez = "          " + instalacionDelJuez + "\n"
	lineaDelIfDeLaMedida        = "    if: " + condicionDeLaMedida + "\n"
	valorPorOmisionDeLaEntrada  = "        default: false\n"
	entradaDeLaMedidaDelFlujo   = "      " + entradaDeLaMedida + ":\n        type: boolean\n" + valorPorOmisionDeLaEntrada
)

// pasoDeInstalacionDelContrato es el paso del trabajo evals de
// definicionDelContrato que instala los dos Claude Code, el de
// contracts/job-de-evals.md §2 de H24.
const pasoDeInstalacionDelContrato = `      - name: Instalar strace y Claude Code
        run: |
          ` + instalacionDeLasSesiones + `
          claude --version
` + lineaDeLaInstalacionDelJuez + `          "$RUNNER_TEMP/claude-del-juez/node_modules/.bin/claude" --version
          echo "CLAUDE_DEL_JUEZ=$RUNNER_TEMP/claude-del-juez/node_modules/.bin/claude" >> "$GITHUB_ENV"
`

// pasoDeLaDecisionDelContrato es el paso del trabajo evals de
// definicionDelContrato que decide qué mide y qué reutiliza, el de
// contracts/job-de-evals.md §2 de H30, con su plazo de 15 minutos.
const pasoDeLaDecisionDelContrato = `      - name: ` + pasoDeLaDecision + `
        env:
          GH_TOKEN: ${{ secrets.GITHUB_TOKEN }}
          GH_REPO: ${{ github.repository }}
        run: >-
          go test -tags evals -count=1 -timeout 15m -v -run '^TestDecisionDelTrabajo$' ./internal/evals/
          -args -skill "$SKILL_EVALUADA" -modo "$MODO_DE_EVALS" -commit "$COMMIT_EVALUADO" -propuesta "$PROPUESTA"
          -base "$BASE_DE_LA_PROPUESTA" -modelo-que-decide "$MODELO_DE_EVALS"
          -version-de-claude-code "$VERSION_DE_CLAUDE_CODE" -modelo-del-juez "$MODELO_DEL_JUEZ"
          -version-del-juez "$VERSION_DE_CLAUDE_CODE_DEL_JUEZ" -salida "$RUNNER_TEMP/decision-de-evals.json"
`

// pasosDeEvalsDelContrato son los pasos del trabajo evals de
// definicionDelContrato: de los de contracts/job-de-evals.md §2 de H24, el que
// instala los dos Claude Code, y, desde H30, los de contracts/job-de-evals.md
// §2 de H30: el checkout con toda la historia, el que instala Go, el de la
// decisión tras él, el que ejecuta las evals con su fichero y, el último, la
// marca del informe escrito.
const pasosDeEvalsDelContrato = `    steps:
      - name: Obtener el commit evaluado
        uses: ` + accionDelCheckout + `
        with:
          ref: ${{ github.event.pull_request.head.sha || github.sha }}
          fetch-depth: 0
      - name: Instalar Go
        uses: ` + accionDeGo + `
        with:
          go-version-file: go.mod
` + pasoDeLaDecisionDelContrato + pasoDeInstalacionDelContrato + `      - name: ` + pasoQueEjecutaLasEvals + `
        env:
          CLAUDE_CODE_OAUTH_TOKEN: ${{ secrets.CLAUDE_CODE_OAUTH_TOKEN }}
        run: ` + ordenQueEjecutaLasEvals + `
      - name: ` + marcaDelInformeEscrito + `
        if: ` + condicionDeLaMarcaDelInforme + `
        run: ` + ordenDeLaMarcaDelInforme + `
`

// trabajoDeLaMedidaDelContrato es el trabajo medida de definicionDelContrato, el
// de contracts/job-de-evals.md §3 de H24 sin sus pasos, que no se comprueban, y
// con el tope que cubre el peor caso de la medida con los casos de
// evalsSinteticas. Desde H25 su matriz lleva las dos skills con juez, cada una
// con la concurrencia de su trabajo evals (contracts/job-de-evals.md §1 de H25);
// y, desde H31, detrás de ellas, la entrada descripciones, la de la medida del
// juez de las descripciones, a cinco casos a la vez
// (contracts/juez-de-las-descripciones.md §7 de H31; research D9 de H31).
const trabajoDeLaMedidaDelContrato = `  medida:
    name: medida del juez (${{ matrix.skill }})
` + lineaDelIfDeLaMedida + `    strategy:
      fail-fast: false
      matrix:
        skill: [normativa, jurisprudencia, descripciones]
        include:
          - skill: normativa
            concurrencia: 5
          - skill: jurisprudencia
            concurrencia: 4
          - skill: descripciones
            concurrencia: 5
    runs-on: ubuntu-24.04
    timeout-minutes: 17
    permissions:
      contents: read
    env:
` + entornoDelJuezDelContrato + `      CONCURRENCIA_DE_EVALS: ${{ matrix.concurrencia }}
      SKILL_EVALUADA: ${{ matrix.skill }}
      COMMIT_EVALUADO: ${{ github.event.pull_request.head.sha || github.sha }}
`

// definicionDelContrato es una definición sintética del job con las claves de
// los contratos y el env de hoy (contracts/ejecucion-del-job.md §6 de H7.3;
// contracts/tanda-del-job.md §1 de H7.4; contracts/job-de-evals.md §1 a §3 de
// H24). Con las evals de evalsSinteticas, cada skill tiene 7 sesiones en el modo
// orden —la eval con el modelo que decide y con el de Haiku, tres veces con cada
// uno, y la prueba de red— y 6 en el modo herramienta, así que el peor caso de
// sus sesiones es de 1573 s en normativa, que va a cinco a la vez desde
// antes de H20, y en jurisprudencia, que va a cuatro (⌈7 / 5⌉ + ⌈6 / 5⌉ =
// ⌈7 / 4⌉ + ⌈6 / 4⌉ = 4 tandas), y de 2389 s en legal-core, que va de dos en dos
// (⌈7 / 2⌉ + ⌈6 / 2⌉ = 7 tandas). normativa y, desde H25, jurisprudencia
// tienen además juez, que juzga las 3 respuestas del modelo que decide en cada
// modo: 60 s de instalar su Claude Code y, de sus votos, (⌈3 × 6 / 5⌉ +
// ⌈3 × 6 / 5⌉) × 40 s en normativa, 1953 s en total, y (⌈3 × 6 / 4⌉ +
// ⌈3 × 6 / 4⌉) × 40 s en jurisprudencia, 2033 s. Desde H30 cada uno suma los
// 900 s del plazo de la decisión (contracts/job-de-evals.md §2 de H30): 2853 s
// en normativa, 2933 s en jurisprudencia y 3289 s en legal-core. Desde
// H31, los 6 × 40 s de los votos de su descripción y, legal-core, que no tiene
// juez de las respuestas, los 60 s de instalar el Claude Code del juez
// (research D10 de H31): 3093 s, 3173 s y 3589 s, y 120 minutos cubren los
// tres. Y el de la medida de cada una, con sus 7 casos
// etiquetados, es de 485 + 60 + ⌈7 × 6 / 5⌉ × 40 = 905 s en normativa y
// 485 + 60 + ⌈7 × 6 / 4⌉ × 40 = 985 s en jurisprudencia, y, desde H31, el de la
// del juez de las descripciones, con los 7 casos de la evidencia de
// evalsSinteticas a cinco a la vez, de 905 s, que cubren los 17 minutos de su
// trabajo: el trabajo medida no decide nada.
const definicionDelContrato = `name: evals
on:
  workflow_dispatch:
    inputs:
` + entradaDeLaMedidaDelFlujo + `  pull_request:
    types: [opened, reopened, labeled]
jobs:
` + trabajoDeLaTandaDelContrato + trabajoDeEvalsDelContrato + trabajoDeLaMedidaDelContrato + trabajoDeCambiosDelContrato

// trabajoDeEvalsDelContrato es el trabajo evals de definicionDelContrato, con
// los permisos y el env que le añade contracts/job-de-evals.md §2 de H30.
const trabajoDeEvalsDelContrato = `  evals:
    name: evals (${{ matrix.skill }})
    needs: [tanda]
    if: ` + condicionDelTrabajo + `
    concurrency:
      group: evals-${{ github.event.pull_request.head.sha || github.sha }}-${{ matrix.skill }}
      cancel-in-progress: false
    strategy:
      fail-fast: false
      matrix:
        skill: [normativa, jurisprudencia, legal-core]
        include:
          - skill: normativa
            concurrencia: 5
            objetivo_de_duracion: 900
          - skill: jurisprudencia
            concurrencia: 4
            objetivo_de_duracion: 0
          - skill: legal-core
            concurrencia: 2
            objetivo_de_duracion: 0
    runs-on: ubuntu-24.04
    timeout-minutes: 120
    permissions:
      contents: read
      actions: read
      pull-requests: read
    env:
      MODELO_DE_EVALS: claude-sonnet-5
` + entornoDelJuezDelContrato + `      MODELOS_INFORMATIVOS_DE_EVALS: claude-haiku-4-5-20251001
      REPETICIONES_DE_EVALS: 3
      UMBRAL_DE_EVALS: 2
      CONCURRENCIA_DE_EVALS: ${{ matrix.concurrencia }}
      OBJETIVO_DE_DURACION_DE_EVALS: ${{ matrix.objetivo_de_duracion }}
      VERSION_DE_CLAUDE_CODE: ` + versionDeLasSesionesDelContrato + `
      MODO_DE_EVALS: ${{ needs.tanda.outputs.modo }}
      PROPUESTA: ${{ github.event.pull_request.number }}
      BASE_DE_LA_PROPUESTA: ${{ github.event.pull_request.base.sha }}
` + pasosDeEvalsDelContrato

// cambioDeLaDefinicion sustituye, en definicionDelContrato, un fragmento por
// otro: el primero que haya. El trabajo evals va delante del trabajo medida,
// así que un fragmento que está en los dos se cambia en el trabajo evals.
type cambioDeLaDefinicion struct {
	antes   string
	despues string
}

// enLaMedida es el cambio de un fragmento por otro dentro del trabajo medida de
// definicionDelContrato, también si el fragmento está antes en otro trabajo.
func enLaMedida(antes, despues string) cambioDeLaDefinicion {
	return cambiosEnLaMedida(cambioDeLaDefinicion{antes, despues})
}

// cambiosEnLaMedida es el cambio que hace esos cambios, en su orden, dentro del
// trabajo medida de definicionDelContrato: dos cambios de enLaMedida no se
// pueden hacer uno detrás de otro, porque el segundo ya no encuentra el trabajo
// del contrato.
func cambiosEnLaMedida(cambios ...cambioDeLaDefinicion) cambioDeLaDefinicion {
	medida := trabajoDeLaMedidaDelContrato
	for _, cambio := range cambios {
		medida = strings.Replace(medida, cambio.antes, cambio.despues, 1)
	}

	return cambioDeLaDefinicion{antes: trabajoDeLaMedidaDelContrato, despues: medida}
}

// enLosDosTrabajos son los cambios de un fragmento por otro en el trabajo evals
// y en el trabajo medida de definicionDelContrato, que lo tienen los dos: el
// env de la medida repite el modelo y la versión del juez del de evals.
func enLosDosTrabajos(antes, despues string) []cambioDeLaDefinicion {
	return []cambioDeLaDefinicion{{antes, despues}, enLaMedida(antes, despues)}
}

// definicionSintetica es una definición que parte de la del contrato con sus
// cambios y las líneas que su comprobación da, en su orden.
type definicionSintetica struct {
	nombre  string
	cambios []cambioDeLaDefinicion
	fallos  []string
}

// probarLasDefinicionesSinteticas escribe en un directorio temporal cada
// definición sintética, la lee con leerDefinicionDelJob y exige que su
// comprobación, con las evals de evalsSinteticas, dé exactamente sus líneas: la
// del contrato, ninguna; y cada una que se aparta de él en una clave, la línea
// de esa clave, con el valor encontrado y el esperado.
func probarLasDefinicionesSinteticas(t *testing.T) {
	t.Parallel()

	evals := evalsSinteticas(t)

	for _, caso := range definicionesSinteticas() {
		t.Run(caso.nombre, func(t *testing.T) {
			t.Parallel()

			assert.Equal(t, caso.fallos, comprobarLaDefinicion(leerLaSintetica(t, caso.cambios), evals))
		})
	}
}

// leerLaSintetica escribe en un directorio temporal la definición del contrato
// con los cambios, en su orden, y la lee con leerDefinicionDelJob.
func leerLaSintetica(t *testing.T, cambios []cambioDeLaDefinicion) DefinicionDelJob {
	t.Helper()

	contenido := definicionDelContrato
	for _, cambio := range cambios {
		require.Contains(t, contenido, cambio.antes, "premisa: la definici\xc3\xb3n tiene lo que el caso cambia")
		require.NotEqual(t, cambio.antes, cambio.despues, "premisa: el cambio del caso cambia algo")
		contenido = strings.Replace(contenido, cambio.antes, cambio.despues, 1)
	}

	ruta := filepath.Join(t.TempDir(), "evals.yml")
	require.NoError(t, os.WriteFile(ruta, []byte(contenido), 0o600))

	leida, err := leerDefinicionDelJob(ruta)
	require.NoError(t, err)

	return leida
}

// probarElTopeConLasEvalsDelRepositorio fija, con definiciones sintéticas y la
// carpeta de evals del repositorio, qué tope cubre el peor caso de cada una de
// las tres skills de la matriz (contracts/evals-jurisprudencia.md §6 de H23;
// research.md M5 de H23; FR-063 de H23). Desde H25 (contracts/job-de-evals.md §3
// y §4 de H25; research M6 de H25; FR-071, FR-072, FR-103; SC-003), el de
// jurisprudencia, con sus diez evals a cuatro a la vez y con su juez, es de
// 12 577 s —61 sesiones en el modo orden, con la prueba de red, que se cuenta
// aunque su trabajo no la lleve, y 60 en el modo herramienta, y las 30
// respuestas que su juez juzga en cada modo—. Desde H9, el de legal-core, con
// sus nueve evals, era de 14 629 s y, con la de vigencia del ADR 0042, de
// 16 261 s, entre el de normativa y el de jurisprudencia (FR-050 de
// H9). Desde antes de H20, normativa va a cinco
// a la vez y, desde H20, con sus veintitrés evals, el suyo es de 18 977 s
// (research D16 y S1 de H20). Desde H26.1, con la 11 y la 12, el de legal-core
// es de 19 525 s, el mayor de los tres (research V10 y D13 de H26.1). Desde
// H30, cada uno suma los 900 s del plazo de la decisión (contracts/job-de-evals.md
// §2 de H30): 19 877 s en normativa, 13 477 s en jurisprudencia y
// 20 425 s en legal-core, que sigue siendo el mayor. Desde H31, cada uno suma
// el peor caso de su descripción, 6 × 40 s y, en legal-core, que no tiene juez
// de las respuestas, los 60 s de instalar el Claude Code del juez (research D10
// de H31; SC-015): 20 117 s, 13 717 s y 20 725 s. El tope
// de la definición del job cubre los tres; con 345 minutos, 20 700 s, ya no
// cubre el de legal-core, y con 333, 19 980 s, tampoco el de normativa,
// aunque los dos cubrirían el de sus evals; con 288 minutos, 17 280 s, ya no
// cubre el de normativa ni el de legal-core; con 209,
// que son 12 540 s, ninguno, ni el de legal-core ni el de jurisprudencia; y con
// 107, 6 420 s, tampoco.
//
// Y las tres mutaciones de contracts/job-de-evals.md §4 de H25, con los topes
// del repositorio, 352 y 269 minutos, que cubren el peor caso de cada trabajo
// con las evals y los casos etiquetados del repositorio: la definición que deja
// a jurisprudencia de una en una en el trabajo evals da la línea de su include
// y la del tope, con su peor caso de 48 757 s; la que la deja de una en una en
// el trabajo medida, la de su include y la del tope de la medida, con el de sus
// 249 casos, 60 305 s; y la que la quita de la matriz del trabajo medida, la de
// esa matriz. Desde H31, la de SC-015 en la entrada descripciones del trabajo
// medida: con los 99 casos de la evidencia del juez de las descripciones de uno
// en uno, la línea de su include y la del tope de la medida, con su peor caso de
// 24 305 s, que a cinco a la vez es de 5305 s (research D9 de H31). En los demás
// casos el trabajo medida va con el tope de la
// definición del job, que cubre su peor caso: así toda línea es del trabajo
// evals.
func probarElTopeConLasEvalsDelRepositorio(t *testing.T) {
	t.Parallel()

	const (
		topeDelJob      = "timeout-minutes: 120"
		topeDeLaMedida  = "timeout-minutes: 17"
		enMinutos       = "timeout-minutes: "
		topeDeEvals     = "jobs.evals.timeout-minutes: vale "
		alMenosElPeorDe = ", y lo esperado es al menos el peor caso de "
		sinElDeBoe      = topeDeEvals + "288 (17280 s)" + alMenosElPeorDe
		sinElDeJuris    = topeDeEvals + "209 (12540 s)" + alMenosElPeorDe
		sinNinguno      = topeDeEvals + "107 (6420 s)" + alMenosElPeorDe
		sinLaDeBoe      = topeDeEvals + "333 (19980 s)" + alMenosElPeorDe
		sinLaDeLegal    = topeDeEvals + "345 (20700 s)" + alMenosElPeorDe
		deBoe           = "normativa, " + peorCasoDeNormativa
		deJuris         = "jurisprudencia, " + peorCasoDeJurisprudencia
		deLegal         = "legal-core, " + peorCasoDeLegalCore

		// La entrada de include de jurisprudencia de la definición del contrato,
		// que es la misma en los dos trabajos, y la que la deja de una en una, como
		// antes de H25, con las dos líneas que da en cada trabajo con los topes del
		// repositorio (contracts/job-de-evals.md §4 de H25).
		topeDelRepositorio = 352
		medidaDelRepo      = 269
		aCuatroALaVez      = "          - skill: jurisprudencia\n            concurrencia: 4\n"
		deUnaEnUna         = "          - skill: jurisprudencia\n            concurrencia: 1\n"
		includeDeUnaEnUna  = "jobs.evals.strategy.matrix.include, skill jurisprudencia: vale " +
			"{concurrencia: 1, objetivo_de_duracion: 0}, y lo esperado es {concurrencia: 4, objetivo_de_duracion: 0}"
		topeDeUnaEnUna = topeDeEvals + "352 (21120 s)" + alMenosElPeorDe + "jurisprudencia, " +
			peorCasoDeJurisprudenciaDeUnaEnUna
		includeDeLaMedidaDeUnaEnUna = "jobs.medida.strategy.matrix.include, skill jurisprudencia: vale " +
			"{concurrencia: 1}, y lo esperado es {concurrencia: 4}"
		topeDeLaMedidaDeUnaEnUna = "jobs.medida.timeout-minutes: vale 269 (16140 s)" + alMenosElPeorDe +
			"jurisprudencia, " + peorCasoDeLaMedidaDeJurisprudenciaDeUnaEnUna

		// La matriz del trabajo medida de la definición del contrato y la de
		// antes de H25, sin jurisprudencia, con su línea.
		matrizDeLaMedida = "skill: [normativa, jurisprudencia, descripciones]"
		matrizSinJuris   = "skill: [normativa, descripciones]"
		medidaSinJuris   = "jobs.medida.strategy.matrix.skill: vale [normativa, descripciones], y lo esperado es " +
			"[normativa, jurisprudencia, descripciones]"

		// La entrada de include de las descripciones del trabajo medida de la
		// definición del contrato y la que deja sus casos de uno en uno, con las
		// dos líneas que da con los casos de la evidencia del repositorio y su
		// tope (research D9 de H31; SC-015).
		descripcionesACinco                 = "          - skill: descripciones\n            concurrencia: 5\n"
		descripcionesDeUnoEnUno             = "          - skill: descripciones\n            concurrencia: 1\n"
		includeDeLasDescripcionesDeUnoEnUno = "jobs.medida.strategy.matrix.include, skill descripciones: vale " +
			"{concurrencia: 1}, y lo esperado es {concurrencia: 5}"
		topeDeLasDescripcionesDeUnoEnUno = "jobs.medida.timeout-minutes: vale 269 (16140 s)" + alMenosElPeorDe +
			"descripciones, " + peorCasoDeLaMedidaDeLasDescripcionesDeUnoEnUno
	)

	delRepositorio, err := leerDefinicionDelJob(rutaDeLaDefinicionDelJob)
	require.NoError(t, err)
	require.NotNil(t, delRepositorio.Medida, "premisa: la definici\xc3\xb3n del job tiene el trabajo medida")
	require.Equal(t, topeDelRepositorio, delRepositorio.TopeEnMinutos,
		"premisa: el tope del trabajo evals es el de contracts/job-de-evals.md \xc2\xa72 de H25")
	require.Equal(t, medidaDelRepo, delRepositorio.Medida.TopeEnMinutos,
		"premisa: el tope del trabajo medida es el de contracts/job-de-evals.md \xc2\xa72 de H25")

	// La definición sintética reparte las sesiones y los votos como la del
	// repositorio: las mismas skills en cada trabajo, con su concurrencia, las
	// mismas repeticiones y los mismos modelos, que solo cuentan por cuántos son.
	// Los casos solo le cambian los topes y, en tres, lo que dice su nombre.
	delContrato := leerLaSintetica(t, nil)
	require.Equal(t, delRepositorio.Skills, delContrato.Skills, "premisa: las skills de la matriz")
	require.Equal(t, delRepositorio.Repeticiones, delContrato.Repeticiones, "premisa: las repeticiones")
	require.Len(t, delContrato.ModelosInformativos, len(delRepositorio.ModelosInformativos),
		"premisa: los modelos informativos")

	for _, skill := range delRepositorio.Skills {
		require.Equal(t, delRepositorio.PorSkill[skill].Concurrencia, delContrato.PorSkill[skill].Concurrencia,
			"premisa: la concurrencia de %s", skill)
	}

	plazoDelRepositorio, err := delRepositorio.plazoDeLaDecision()
	require.NoError(t, err)

	plazoDelContrato, err := delContrato.plazoDeLaDecision()
	require.NoError(t, err)
	require.Equal(t, plazoDelRepositorio, plazoDelContrato, "premisa: el plazo de la decisi\xc3\xb3n")

	require.NotNil(t, delContrato.Medida, "premisa: la definici\xc3\xb3n del contrato tiene el trabajo medida")
	require.Equal(t, delRepositorio.Medida.Skills, delContrato.Medida.Skills, "premisa: las skills de la matriz de la medida")
	require.Equal(t, delRepositorio.Medida.ConcurrenciaPorSkill, delContrato.Medida.ConcurrenciaPorSkill,
		"premisa: la concurrencia de cada skill en la medida")

	// conElTope son los cambios que ponen ese tope al trabajo evals y el del
	// repositorio al trabajo medida, con los demás cambios dados en este último.
	conElTope := func(minutos int, enLaMedidaAdemas ...cambioDeLaDefinicion) []cambioDeLaDefinicion {
		delTope := cambioDeLaDefinicion{topeDeLaMedida, enMinutos + strconv.Itoa(medidaDelRepo)}

		return []cambioDeLaDefinicion{
			{topeDelJob, enMinutos + strconv.Itoa(minutos)},
			cambiosEnLaMedida(slices.Concat([]cambioDeLaDefinicion{delTope}, enLaMedidaAdemas)...),
		}
	}

	casos := []definicionSintetica{
		{nombre: "el-del-repositorio", cambios: conElTope(topeDelRepositorio)},
		{
			// Los dos mutantes de SC-015 de H31: 345 minutos cubrirían los 20 425 s
			// de las evals de legal-core y no cubren los 20 725 s que suman con su
			// descripción; 333, los 19 877 s de las de normativa y no sus
			// 20 117 s, ni el de legal-core.
			nombre:  "por-debajo-del-de-legal-core-con-su-descripcion",
			cambios: conElTope(345),
			fallos:  []string{sinLaDeLegal + deLegal},
		},
		{
			nombre:  "por-debajo-del-de-normativa-con-su-descripcion",
			cambios: conElTope(333),
			fallos:  []string{sinLaDeBoe + deBoe, sinLaDeBoe + deLegal},
		},
		{
			nombre:  "por-debajo-del-de-normativa-y-del-de-legal-core",
			cambios: conElTope(288),
			fallos:  []string{sinElDeBoe + deBoe, sinElDeBoe + deLegal},
		},
		{
			nombre:  "por-debajo-del-de-jurisprudencia",
			cambios: conElTope(209),
			fallos:  []string{sinElDeJuris + deBoe, sinElDeJuris + deJuris, sinElDeJuris + deLegal},
		},
		{
			nombre:  "por-debajo-de-los-tres",
			cambios: conElTope(107),
			fallos:  []string{sinNinguno + deBoe, sinNinguno + deJuris, sinNinguno + deLegal},
		},
		{
			// La mutación de SC-003 de H25 en el trabajo evals: con las diez evals
			// de una en una son 121 tandas de sesiones y 360 de votos, que el tope
			// del repositorio no cubre. El fragmento está en los dos trabajos, y el
			// que cambia es el del trabajo evals, que va delante.
			nombre:  "jurisprudencia-de-una-en-una",
			cambios: slices.Concat(conElTope(topeDelRepositorio), []cambioDeLaDefinicion{{aCuatroALaVez, deUnaEnUna}}),
			fallos:  []string{includeDeUnaEnUna, topeDeUnaEnUna},
		},
		{
			// La de SC-003 en el trabajo medida: con los 249 casos de uno en uno son
			// 1494 tandas de votos, que su tope no cubre.
			nombre:  "jurisprudencia-de-una-en-una-en-la-medida",
			cambios: conElTope(topeDelRepositorio, cambioDeLaDefinicion{aCuatroALaVez, deUnaEnUna}),
			fallos:  []string{includeDeLaMedidaDeUnaEnUna, topeDeLaMedidaDeUnaEnUna},
		},
		{
			// La matriz de antes de H25: sin jurisprudencia, la etiqueta no mide a
			// su juez. Lo esperado sale de las skills con carpeta de juez.
			nombre:  "medida-sin-jurisprudencia",
			cambios: conElTope(topeDelRepositorio, cambioDeLaDefinicion{matrizDeLaMedida, matrizSinJuris}),
			fallos:  []string{medidaSinJuris},
		},
		{
			// La de SC-015 de H31 en el trabajo medida: con los 99 casos de la
			// evidencia de uno en uno son 594 tandas de votos, que su tope no
			// cubre; a cinco a la vez, sus 5305 s, sí.
			nombre:  "descripciones-de-uno-en-uno-en-la-medida",
			cambios: conElTope(topeDelRepositorio, cambioDeLaDefinicion{descripcionesACinco, descripcionesDeUnoEnUno}),
			fallos:  []string{includeDeLasDescripcionesDeUnoEnUno, topeDeLasDescripcionesDeUnoEnUno},
		},
	}

	for _, caso := range casos {
		t.Run(caso.nombre, func(t *testing.T) {
			t.Parallel()

			assert.Equal(t, caso.fallos, comprobarLaDefinicion(leerLaSintetica(t, caso.cambios), directorioDeEvals))
		})
	}
}

// definicionesSinteticas son los casos de probarLasDefinicionesSinteticas: uno
// por comprobación de contracts/ejecucion-del-job.md §7 de H7.3, los del tope
// que lo cubren y, detrás, los del juez y de su medida y, desde H30, los de la
// elección y la decisión.
func definicionesSinteticas() []definicionSintetica {
	return slices.Concat(definicionesSinteticasDelJob(), definicionesSinteticasDelJuez(),
		definicionesSinteticasDeLaMedida(), definicionesSinteticasDeLaEleccion())
}

// definicionesSinteticasDelJuez son las definiciones que se apartan de
// contracts/job-de-evals.md §5 de H24 en el juez del trabajo evals, cada una con
// su línea: el modelo del juez que falta, que es un alias o que es el que
// decide; la versión de Claude Code de sus votos que falta o que no es
// <n>.<n>.<n>; y el paso de instalación que no instala el del juez en su
// prefijo, que no está o que instala con su versión el de las sesiones. El
// modelo y la versión cambian a la vez en el trabajo medida, que los repite:
// así la única línea es la suya.
func definicionesSinteticasDelJuez() []definicionSintetica {
	const (
		modeloEnElEnv  = "jobs.evals.env.MODELO_DEL_JUEZ: "
		versionEnElEnv = "jobs.evals.env.VERSION_DE_CLAUDE_CODE_DEL_JUEZ: "
		yLoEsperadoEs  = ", y lo esperado es "
		conLaDelJuez   = `npm install -g "` + claudeCodeDelJuez + `"`
		soloElDelJuez  = "que con VERSION_DE_CLAUDE_CODE_DEL_JUEZ solo instale el del juez, en su prefijo"
		ordenSinElJuez = instalacionDeLasSesiones + "\nclaude --version\n" +
			"\"$RUNNER_TEMP/claude-del-juez/node_modules/.bin/claude\" --version\n" +
			"echo \"CLAUDE_DEL_JUEZ=$RUNNER_TEMP/claude-del-juez/node_modules/.bin/claude\" >> \"$GITHUB_ENV\"\n"
	)

	return []definicionSintetica{
		{
			nombre:  "sin-el-modelo-del-juez",
			cambios: enLosDosTrabajos(modeloDelJuezEnElEnv, ""),
			fallos:  []string{modeloEnElEnv + noEsta + yLoEsperadoEs + modeloDelJuezEsperado},
		},
		{
			nombre:  "modelo-del-juez-que-es-un-alias",
			cambios: enLosDosTrabajos(modeloDelJuezEnElEnv, "      MODELO_DEL_JUEZ: opus\n"),
			fallos:  []string{modeloEnElEnv + `vale "opus"` + yLoEsperadoEs + modeloDelJuezEsperado},
		},
		{
			nombre:  "modelo-del-juez-que-es-el-que-decide",
			cambios: enLosDosTrabajos(modeloDelJuezEnElEnv, "      MODELO_DEL_JUEZ: claude-sonnet-5\n"),
			fallos:  []string{modeloEnElEnv + `vale "claude-sonnet-5"` + yLoEsperadoEs + modeloDelJuezEsperado},
		},
		{
			nombre:  "sin-la-version-del-juez",
			cambios: enLosDosTrabajos(versionDelJuezEnElEnv, ""),
			fallos:  []string{versionEnElEnv + noEsta + yLoEsperadoEs + versionDelJuezEsperada},
		},
		{
			nombre:  "version-del-juez-sin-su-tercer-numero",
			cambios: enLosDosTrabajos(versionDelJuezEnElEnv, "      VERSION_DE_CLAUDE_CODE_DEL_JUEZ: 2.1\n"),
			fallos:  []string{versionEnElEnv + `vale "2.1"` + yLoEsperadoEs + versionDelJuezEsperada},
		},
		{
			nombre:  "version-del-juez-que-es-una-etiqueta",
			cambios: enLosDosTrabajos(versionDelJuezEnElEnv, "      VERSION_DE_CLAUDE_CODE_DEL_JUEZ: latest\n"),
			fallos:  []string{versionEnElEnv + `vale "latest"` + yLoEsperadoEs + versionDelJuezEsperada},
		},
		{
			nombre:  "instalacion-sin-el-del-juez-en-su-prefijo",
			cambios: []cambioDeLaDefinicion{{lineaDeLaInstalacionDelJuez, ""}},
			fallos: []string{pasoDeLaInstalacion + ": vale " + strconv.Quote(ordenSinElJuez) + yLoEsperadoEs +
				"que tenga la l\xc3\xadnea " + strconv.Quote(instalacionDelJuez)},
		},
		{
			nombre:  "sin-el-paso-de-instalacion",
			cambios: []cambioDeLaDefinicion{{pasoDeInstalacionDelContrato, ""}},
			fallos: []string{pasoDeLaInstalacion + ": " + noEsta + yLoEsperadoEs + "que tenga la l\xc3\xadnea " +
				strconv.Quote(instalacionDelJuez)},
		},
		{
			nombre:  "instalacion-de-las-sesiones-con-la-version-del-juez",
			cambios: []cambioDeLaDefinicion{{instalacionDeLasSesiones, conLaDelJuez}},
			fallos: []string{pasoDeLaInstalacion + ": tiene la l\xc3\xadnea " + strconv.Quote(conLaDelJuez) +
				yLoEsperadoEs + soloElDelJuez},
		},
	}
}

// definicionesSinteticasDeLaMedida son las definiciones que se apartan de
// contracts/job-de-evals.md §5 de H24 en la ejecución de la medida del juez,
// cada una con su línea: el trabajo medida que falta, con otro modelo, con otra
// versión, con otro if o con needs; el trabajo tanda o el trabajo evals que
// nombran la etiqueta de la medida, o la tanda que admite el despacho con su
// entrada; la entrada que falta, sin valor por omisión o con otro; y el tope de
// la medida un minuto por debajo de su peor caso, o con un peor caso que no se
// puede obtener. Y, desde H25, las que se apartan de contracts/job-de-evals.md
// §4 de H25 en la matriz del trabajo medida (FR-071, FR-103): con una skill sin
// juez, sin jurisprudencia, con ella de una en una o sin su entrada de include.
// Y, desde H31, las que se apartan de contracts/juez-de-las-descripciones.md §7
// de H31 en la entrada descripciones de esa matriz (research D9 de H31; FR-034):
// sin ella, con sus casos de uno en uno o sin su entrada de include.
func definicionesSinteticasDeLaMedida() []definicionSintetica {
	return slices.Concat(definicionesSinteticasDelTrabajoMedida(), definicionesSinteticasDeLaMatrizDeLaMedida())
}

// definicionesSinteticasDelTrabajoMedida son las definiciones de
// definicionesSinteticasDeLaMedida que se apartan de contracts/job-de-evals.md
// §5 de H24 en el trabajo medida, en el flujo y en su tope.
func definicionesSinteticasDelTrabajoMedida() []definicionSintetica {
	const (
		yLoEsperadoEs    = ", y lo esperado es "
		soloConEtiqueta  = "github.event.label.name == 'evals-medir-juez'"
		despachoDeAntes  = "github.event_name == 'workflow_dispatch'"
		conLaEtiqueta    = "  github.event.label.name == 'evals-medir-juez' ||\n"
		lineaDelDespacho = "  " + despachoSinLaMedida + " ||\n"
		evalsConEtiqueta = "${{ " + sinCancelar + " && (needs.tanda.outputs.medir == 'si' || " + soloConEtiqueta + ") }}"
		entradaDelFlujo  = "on.workflow_dispatch.inputs.medir_al_juez"
		topeDeLaMedida   = "jobs.medida.timeout-minutes: "
		concurrenciaDe5  = "            concurrencia: 5\n"
		peorCasoDeMedida = "985 s = 485 s + 60 s + (\xe2\x8c\x887 \xc3\x97 6 / 4\xe2\x8c\x89) \xc3\x97 (35 s + 5 s)"

		// El de la medida de normativa, que va a cinco a la vez desde antes
		// de H20, que es también el de la de las descripciones, con sus siete
		// casos a cinco a la vez.
		peorCasoDeMedidaDeBoe = "905 s = 485 s + 60 s + (\xe2\x8c\x887 \xc3\x97 6 / 5\xe2\x8c\x89) \xc3\x97 (35 s + 5 s)"

		// La clave de una entrada del include del trabajo medida.
		includeDeLaMedida = "jobs.medida.strategy.matrix.include, skill "
	)

	return []definicionSintetica{
		{
			nombre:  "sin-medida",
			cambios: []cambioDeLaDefinicion{{trabajoDeLaMedidaDelContrato, ""}},
			fallos: []string{"jobs.medida: " + noEsta + yLoEsperadoEs +
				"el trabajo que ejecuta la medida del juez"},
		},
		{
			nombre:  "medida-con-otro-modelo",
			cambios: []cambioDeLaDefinicion{enLaMedida(modeloDelJuezEnElEnv, "      MODELO_DEL_JUEZ: claude-opus-5\n")},
			fallos: []string{`jobs.medida.env.MODELO_DEL_JUEZ: vale "claude-opus-5"` + yLoEsperadoEs +
				`lo de jobs.evals.env.MODELO_DEL_JUEZ, que vale "claude-opus-5-5"`},
		},
		{
			nombre: "medida-con-otra-version",
			cambios: []cambioDeLaDefinicion{
				enLaMedida(versionDelJuezEnElEnv, "      VERSION_DE_CLAUDE_CODE_DEL_JUEZ: 2.1.284\n"),
			},
			fallos: []string{`jobs.medida.env.VERSION_DE_CLAUDE_CODE_DEL_JUEZ: vale "2.1.284"` + yLoEsperadoEs +
				`lo de jobs.evals.env.VERSION_DE_CLAUDE_CODE_DEL_JUEZ, que vale "2.1.289"`},
		},
		{
			nombre:  "medida-sin-la-version",
			cambios: []cambioDeLaDefinicion{enLaMedida(versionDelJuezEnElEnv, "")},
			fallos: []string{"jobs.medida.env.VERSION_DE_CLAUDE_CODE_DEL_JUEZ: " + noEsta + yLoEsperadoEs +
				`lo de jobs.evals.env.VERSION_DE_CLAUDE_CODE_DEL_JUEZ, que vale "2.1.289"`},
		},
		{
			nombre:  "medida-solo-con-su-etiqueta",
			cambios: []cambioDeLaDefinicion{{lineaDelIfDeLaMedida, "    if: " + soloConEtiqueta + "\n"}},
			fallos: []string{"jobs.medida.if: vale " + strconv.Quote(soloConEtiqueta) + yLoEsperadoEs +
				strconv.Quote(condicionDeLaMedida)},
		},
		{
			nombre:  "medida-sin-if",
			cambios: []cambioDeLaDefinicion{{lineaDelIfDeLaMedida, ""}},
			fallos:  []string{"jobs.medida.if: " + noEsta + yLoEsperadoEs + strconv.Quote(condicionDeLaMedida)},
		},
		{
			nombre:  "medida-con-needs",
			cambios: []cambioDeLaDefinicion{{lineaDelIfDeLaMedida, lineaDelIfDeLaMedida + "    needs: [tanda]\n"}},
			fallos: []string{"jobs.medida.needs: est\xc3\xa1" + yLoEsperadoEs + queNoEste +
				": la medida no abre sesiones de evals, y no depende de la tanda que las decide"},
		},
		{
			nombre:  "tanda-que-nombra-la-etiqueta-de-la-medida",
			cambios: []cambioDeLaDefinicion{{"      " + lineaDelDespacho, "      " + lineaDelDespacho + "      " + conLaEtiqueta}},
			fallos: []string{"jobs.tanda.if: vale " +
				strconv.Quote(strings.Replace(condicionDeLaTanda, lineaDelDespacho, lineaDelDespacho+conLaEtiqueta, 1)) +
				yLoEsperadoEs + strconv.Quote(condicionDeLaTanda)},
		},
		{
			nombre:  "tanda-que-admite-el-despacho-con-la-entrada",
			cambios: []cambioDeLaDefinicion{{despachoSinLaMedida, despachoDeAntes}},
			fallos: []string{"jobs.tanda.if: vale " +
				strconv.Quote(strings.Replace(condicionDeLaTanda, despachoSinLaMedida, despachoDeAntes, 1)) +
				yLoEsperadoEs + strconv.Quote(condicionDeLaTanda)},
		},
		{
			nombre:  "evals-que-nombra-la-etiqueta-de-la-medida",
			cambios: []cambioDeLaDefinicion{{condicionDelTrabajo, evalsConEtiqueta}},
			fallos: []string{"jobs.evals.if: vale " + strconv.Quote(evalsConEtiqueta) + yLoEsperadoEs +
				strconv.Quote(condicionDelTrabajo)},
		},
		{
			nombre:  "sin-la-entrada-de-la-medida",
			cambios: []cambioDeLaDefinicion{{entradaDeLaMedidaDelFlujo, ""}},
			fallos: []string{entradaDelFlujo + ": " + noEsta + yLoEsperadoEs +
				"la entrada que lanza la medida del juez, con default: false"},
		},
		{
			nombre:  "entrada-de-la-medida-sin-default",
			cambios: []cambioDeLaDefinicion{{valorPorOmisionDeLaEntrada, ""}},
			fallos:  []string{entradaDelFlujo + ".default: " + noEsta + yLoEsperadoEs + "false"},
		},
		{
			nombre:  "entrada-de-la-medida-con-default-true",
			cambios: []cambioDeLaDefinicion{{valorPorOmisionDeLaEntrada, "        default: true\n"}},
			fallos:  []string{entradaDelFlujo + ".default: vale true" + yLoEsperadoEs + "false"},
		},
		{
			// Un texto no es el booleano false, aunque se escriba igual.
			nombre:  "entrada-de-la-medida-con-un-texto-de-default",
			cambios: []cambioDeLaDefinicion{{valorPorOmisionDeLaEntrada, "        default: \"false\"\n"}},
			fallos:  []string{entradaDelFlujo + `.default: vale "false"` + yLoEsperadoEs + "false"},
		},
		{
			// 15 minutos son 900 s, y el peor caso de la medida de normativa,
			// 905 s, el de jurisprudencia, 985 s, y el de las descripciones, 905 s:
			// los cubren los 17 de la definición del contrato.
			nombre:  "tope-de-la-medida-por-debajo-de-su-peor-caso",
			cambios: []cambioDeLaDefinicion{{"timeout-minutes: 17", "timeout-minutes: 15"}},
			fallos: []string{
				topeDeLaMedida + "vale 15 (900 s)" + yLoEsperadoEs + "al menos el peor caso de normativa, " +
					peorCasoDeMedidaDeBoe,
				topeDeLaMedida + "vale 15 (900 s)" + yLoEsperadoEs + "al menos el peor caso de jurisprudencia, " + peorCasoDeMedida,
				topeDeLaMedida + "vale 15 (900 s)" + yLoEsperadoEs + "al menos el peor caso de descripciones, " +
					peorCasoDeMedidaDeBoe,
			},
		},
		{
			// La entrada sin su concurrencia da la línea de su include y, sin
			// tandas que contar, la de su peor caso.
			nombre:  "medida-sin-la-concurrencia-de-su-skill",
			cambios: []cambioDeLaDefinicion{enLaMedida(concurrenciaDe5, "")},
			fallos: []string{
				includeDeLaMedida + "normativa: vale {concurrencia: 0}" + yLoEsperadoEs + "{concurrencia: 5}",
				topeDeLaMedida + "el peor caso de normativa no se puede obtener: la concurrencia " +
					"de normativa es 0 y tiene que ser un entero mayor o igual que 1",
			},
		},
	}
}

// definicionesSinteticasDeLaMatrizDeLaMedida son las definiciones de
// definicionesSinteticasDeLaMedida que se apartan de la matriz del trabajo
// medida de contracts/job-de-evals.md §4 de H25 y, desde H31, de su entrada
// descripciones (contracts/juez-de-las-descripciones.md §7 de H31), cada una con
// sus líneas: la de la matriz, la de su include y, si su peor caso no se cubre o
// no se puede obtener, la del tope.
func definicionesSinteticasDeLaMatrizDeLaMedida() []definicionSintetica {
	const (
		yLoEsperadoEs     = ", y lo esperado es "
		topeDeLaMedida    = "jobs.medida.timeout-minutes: "
		matrizEsperada    = yLoEsperadoEs + "[normativa, jurisprudencia, descripciones]"
		includeDeLaMedida = "jobs.medida.strategy.matrix.include, skill "

		// La matriz del trabajo medida de la definición del contrato y las
		// entradas de include de jurisprudencia y de las descripciones.
		matrizDeLaMedida  = "skill: [normativa, jurisprudencia, descripciones]"
		includeDeJuris    = "          - skill: jurisprudencia\n            concurrencia: 4\n"
		includeDeLasDescr = "          - skill: descripciones\n            concurrencia: 5\n"

		// El peor caso de una medida con los siete casos de uno en uno: 42 tandas
		// de votos, 2225 s, que los 17 minutos de la definición del contrato no
		// cubren.
		tope17DeUnoEnUno = topeDeLaMedida + "vale 17 (1020 s)" + yLoEsperadoEs + "al menos el peor caso de %s, " +
			"2225 s = 485 s + 60 s + (\xe2\x8c\x887 \xc3\x97 6 / 1\xe2\x8c\x89) \xc3\x97 (35 s + 5 s)"
		sinConcurrencia = topeDeLaMedida + "el peor caso de %[1]s no se puede obtener: la concurrencia " +
			"de %[1]s es 0 y tiene que ser un entero mayor o igual que 1"
	)

	return []definicionSintetica{
		{
			// legal-core no tiene juez: no es de la matriz de la medida, ni hay
			// medida suya que quepa en ningún tope.
			nombre: "medida-de-una-skill-sin-juez",
			cambios: []cambioDeLaDefinicion{enLaMedida(
				matrizDeLaMedida+"\n        include:\n",
				"skill: [normativa, jurisprudencia, descripciones, legal-core]\n        include:\n"+
					"          - skill: legal-core\n            concurrencia: 1\n")},
			fallos: []string{
				"jobs.medida.strategy.matrix.skill: vale [normativa, jurisprudencia, descripciones, legal-core]" +
					matrizEsperada,
				includeDeLaMedida + "legal-core: vale {concurrencia: 1}" + yLoEsperadoEs + queNoEste,
				topeDeLaMedida + "el peor caso de legal-core no se puede obtener: legal-core no tiene juez que medir",
			},
		},
		{
			// La matriz de antes de H25: jurisprudencia tiene juez, y la matriz
			// esperada, que sale de las skills con carpeta de juez, la lleva.
			nombre:  "medida-sin-jurisprudencia",
			cambios: []cambioDeLaDefinicion{enLaMedida(matrizDeLaMedida, "skill: [normativa, descripciones]")},
			fallos:  []string{"jobs.medida.strategy.matrix.skill: vale [normativa, descripciones]" + matrizEsperada},
		},
		{
			// La matriz de antes de H31: sin la entrada descripciones, la etiqueta
			// no mide al juez de las descripciones (FR-034 de H31).
			nombre:  "medida-sin-descripciones",
			cambios: []cambioDeLaDefinicion{enLaMedida(matrizDeLaMedida, "skill: [normativa, jurisprudencia]")},
			fallos:  []string{"jobs.medida.strategy.matrix.skill: vale [normativa, jurisprudencia]" + matrizEsperada},
		},
		{
			nombre: "medida-con-descripciones-de-uno-en-uno",
			cambios: []cambioDeLaDefinicion{
				enLaMedida(includeDeLasDescr, "          - skill: descripciones\n            concurrencia: 1\n"),
			},
			fallos: []string{
				includeDeLaMedida + "descripciones: vale {concurrencia: 1}" + yLoEsperadoEs + "{concurrencia: 5}",
				fmt.Sprintf(tope17DeUnoEnUno, entradaDeLasDescripciones),
			},
		},
		{
			nombre:  "medida-sin-el-include-de-descripciones",
			cambios: []cambioDeLaDefinicion{enLaMedida(includeDeLasDescr, "")},
			fallos: []string{
				includeDeLaMedida + "descripciones: " + noEsta + yLoEsperadoEs + "{concurrencia: 5}",
				fmt.Sprintf(sinConcurrencia, entradaDeLasDescripciones),
			},
		},
		{
			nombre:  "medida-con-jurisprudencia-de-una-en-una",
			cambios: []cambioDeLaDefinicion{enLaMedida(includeDeJuris, "          - skill: jurisprudencia\n            concurrencia: 1\n")},
			fallos: []string{
				includeDeLaMedida + "jurisprudencia: vale {concurrencia: 1}" + yLoEsperadoEs + "{concurrencia: 4}",
				fmt.Sprintf(tope17DeUnoEnUno, "jurisprudencia"),
			},
		},
		{
			nombre:  "medida-sin-el-include-de-jurisprudencia",
			cambios: []cambioDeLaDefinicion{enLaMedida(includeDeJuris, "")},
			fallos: []string{
				includeDeLaMedida + "jurisprudencia: " + noEsta + yLoEsperadoEs + "{concurrencia: 4}",
				fmt.Sprintf(sinConcurrencia, "jurisprudencia"),
			},
		},
	}
}

// definicionesSinteticasDelJob son las definiciones que se apartan de
// contracts/tanda-del-job.md §4 de H7.4 y de contracts/ejecucion-del-job.md §7
// de H7.3, y las del tope del trabajo evals que lo cubren. Desde H23, también
// las que se apartan de contracts/evals-jurisprudencia.md §6 de H23 (FR-063 de
// H23): jurisprudencia fuera de la matriz, con objetivo de duración o sin su
// entrada de include; y, desde H25, la que la deja de una en una
// (contracts/job-de-evals.md §4 de H25; FR-103). Con las evals de
// evalsSinteticas, jurisprudencia tiene la misma eval, la misma concurrencia y
// el mismo juez que normativa: su peor caso es el de esta, con los votos
// de su juez.
func definicionesSinteticasDelJob() []definicionSintetica {
	const (
		grupoPorCommit   = "group: evals-${{ github.event.pull_request.head.sha || github.sha }}-${{ matrix.skill }}"
		cancelaEnCurso   = "      cancel-in-progress: false\n"
		matrizDelJob     = "skill: [normativa, jurisprudencia, legal-core]"
		matrizEsperada   = "[normativa, jurisprudencia, legal-core]"
		deUnaEnUna       = "            concurrencia: 1\n"
		deDosEnDos       = "            concurrencia: 2\n"
		deCuatroEnCuatro = "            concurrencia: 4\n"
		sinObjetivo      = "            objetivo_de_duracion: 0\n"
		conObjetivo      = "            objetivo_de_duracion: 900\n"
		skillDeJuris     = "          - skill: jurisprudencia\n"
		entradaDeJuris   = skillDeJuris + deCuatroEnCuatro
		entradaDeLegal   = "          - skill: legal-core\n" + deDosEnDos
		includeDeJuris   = entradaDeJuris + sinObjetivo
		includeDeLegal   = entradaDeLegal + sinObjetivo
		includeDeSkill   = "jobs.evals.strategy.matrix.include, skill "
		topeDelJob       = "timeout-minutes: 120"
		topeDeEvals      = "jobs.evals.timeout-minutes: "
		concurrenciaEnv  = "      CONCURRENCIA_DE_EVALS: ${{ matrix.concurrencia }}\n"
		objetivoEnv      = "OBJETIVO_DE_DURACION_DE_EVALS: ${{ matrix.objetivo_de_duracion }}"
		repeticionesEnv  = "REPETICIONES_DE_EVALS: 3"
		ajustesDeBoe     = "{concurrencia: 5, objetivo_de_duracion: 900}"
		ajustesDeJuris   = "{concurrencia: 4, objetivo_de_duracion: 0}"
		ajustesDeLegal   = "{concurrencia: 2, objetivo_de_duracion: 0}"
		peorCasoDeLegal  = "3589 s = 485 s + 900 s + (\xe2\x8c\x887 / 2\xe2\x8c\x89 + \xe2\x8c\x886 / 2\xe2\x8c\x89) \xc3\x97 (22 s + 240 s + 10 s)" +
			" + 60 s" + terminosDeLaDescripcion
		peorCasoConJuez = "3173 s = 485 s + 900 s + (\xe2\x8c\x887 / 4\xe2\x8c\x89 + \xe2\x8c\x886 / 4\xe2\x8c\x89) " +
			"\xc3\x97 (22 s + 240 s + 10 s) + 60 s + (\xe2\x8c\x883 \xc3\x97 6 / 4\xe2\x8c\x89 + " +
			"\xe2\x8c\x883 \xc3\x97 6 / 4\xe2\x8c\x89) \xc3\x97 (35 s + 5 s)" + terminosDeLaDescripcion
		peorCasoDeBoe = "3093 s = 485 s + 900 s + (\xe2\x8c\x887 / 5\xe2\x8c\x89 + \xe2\x8c\x886 / 5\xe2\x8c\x89) " +
			"\xc3\x97 (22 s + 240 s + 10 s) + 60 s + (\xe2\x8c\x883 \xc3\x97 6 / 5\xe2\x8c\x89 + " +
			"\xe2\x8c\x883 \xc3\x97 6 / 5\xe2\x8c\x89) \xc3\x97 (35 s + 5 s)" + terminosDeLaDescripcion
		alMenosElPeorDe    = ", y lo esperado es al menos el peor caso de "
		esperadoDelGrupo   = `"evals-${{ github.event.pull_request.head.sha || github.sha }}-${{ matrix.skill }}"`
		pruebaDeRed        = "  github.event.label.name == 'evals-prueba-de-red' ||\n"
		evalsCompleta      = "  github.event.label.name == 'evals-completa' ||\n"
		ejecucionDeGh      = ` -ejecucion "$EJECUCION"`
		modoDeGh           = ` -modo "$MODO_DE_EVALS"`
		topeDeLaTanda      = "    timeout-minutes: 15\n"
		elPenultimo        = "jobs.tanda.steps, el pen\xc3\xbaltimo, "
		elUltimo           = "jobs.tanda.steps, el \xc3\xbaltimo, "
		yLoEsperadoEs      = ", y lo esperado es "
		sesionesDeLaMatriz = "jobs.evals.strategy.matrix: "
	)

	return []definicionSintetica{
		{nombre: "la-del-contrato"},
		{
			nombre:  "sin-tanda",
			cambios: []cambioDeLaDefinicion{{trabajoDeLaTandaDelContrato, ""}},
			fallos: []string{"jobs.tanda: no est\xc3\xa1, y lo esperado es el trabajo que decide si la ejecuci\xc3\xb3n " +
				"mide el commit"},
		},
		{
			nombre:  "tanda-con-nombre",
			cambios: []cambioDeLaDefinicion{{"  tanda:\n", "  tanda:\n    name: Decidir la tanda\n"}},
			fallos: []string{`jobs.tanda.name: vale "Decidir la tanda", y lo esperado es que no est` + "\xc3\xa9: gh da " +
				"el trabajo por su id, tanda, que es por el que lo busca la decisi\xc3\xb3n"},
		},
		{
			// Un name igual a su id no cambia el nombre con el que gh lo da.
			nombre:  "tanda-con-su-id-como-nombre",
			cambios: []cambioDeLaDefinicion{{"  tanda:\n", "  tanda:\n    name: tanda\n"}},
		},
		{
			nombre:  "tanda-sin-la-prueba-de-red",
			cambios: []cambioDeLaDefinicion{{"      " + pruebaDeRed, ""}},
			fallos: []string{"jobs.tanda.if: vale " + strconv.Quote(strings.Replace(condicionDeLaTanda, pruebaDeRed, "", 1)) +
				", y lo esperado es " + strconv.Quote(condicionDeLaTanda)},
		},
		{
			nombre:  "tanda-con-concurrency",
			cambios: []cambioDeLaDefinicion{{topeDeLaTanda, topeDeLaTanda + "    concurrency: tanda-${{ github.sha }}\n"}},
			fallos: []string{"jobs.tanda.concurrency: est\xc3\xa1, y lo esperado es que no est\xc3\xa9: la tanda ni espera " +
				"ni se cancela, y no deja una comprobaci\xc3\xb3n roja por decidir no medir"},
		},
		{
			nombre:  "tanda-sin-actions-read",
			cambios: []cambioDeLaDefinicion{{"      actions: read\n", ""}},
			fallos:  []string{`jobs.tanda.permissions.actions: no est` + "\xc3\xa1" + `, y lo esperado es "read"`},
		},
		{
			nombre:  "tanda-sin-la-salida-medir",
			cambios: []cambioDeLaDefinicion{{"      medir: ${{ steps.decidir.outputs.medir }}\n", ""}},
			fallos: []string{`jobs.tanda.outputs.medir: no est` + "\xc3\xa1" + `, y lo esperado es ` +
				`"${{ steps.decidir.outputs.medir }}"`},
		},
		{
			// Sin la salida modo, el trabajo evals no sabe con qué modo decidir
			// (contracts/job-de-evals.md §2 de H30).
			nombre:  "tanda-sin-la-salida-modo",
			cambios: []cambioDeLaDefinicion{{"      modo: ${{ steps.decidir.outputs.modo }}\n", ""}},
			fallos: []string{`jobs.tanda.outputs.modo: no est` + "\xc3\xa1" + yLoEsperadoEs +
				strconv.Quote(salidaDelModoDeLaTanda)},
		},
		{
			// La etiqueta evals-completa sin tanda no mediría nada (FR-018 de H30).
			nombre:  "tanda-sin-evals-completa",
			cambios: []cambioDeLaDefinicion{{"      " + evalsCompleta, ""}},
			fallos: []string{"jobs.tanda.if: vale " + strconv.Quote(strings.Replace(condicionDeLaTanda, evalsCompleta, "", 1)) +
				yLoEsperadoEs + strconv.Quote(condicionDeLaTanda)},
		},
		{
			// La salida de antes de H30, coincide, ya no la da cambios: la apertura
			// que elige alguna eval no mediría (FR-016 de H30).
			nombre:  "tanda-con-la-salida-de-antes",
			cambios: []cambioDeLaDefinicion{{"needs.cambios.outputs.elige == 'si'", "needs.cambios.outputs.coincide == 'si'"}},
			fallos: []string{"jobs.tanda.if: vale " + strconv.Quote(strings.Replace(condicionDeLaTanda, "elige", "coincide", 1)) +
				yLoEsperadoEs + strconv.Quote(condicionDeLaTanda)},
		},
		{
			// La etiqueta evals mide con reutilización (FR-017 de H30): con ella en
			// la condición del modo mediría la suite completa.
			nombre: "modo-completa-tambien-con-la-etiqueta-evals",
			cambios: []cambioDeLaDefinicion{{
				"(github.event_name == 'workflow_dispatch' || github.event.label.name == 'evals-completa' ||",
				"(github.event_name == 'workflow_dispatch' || github.event.label.name == 'evals' || " +
					"github.event.label.name == 'evals-completa' ||",
			}},
			fallos: []string{"jobs.tanda.env.MODO_DE_EVALS: vale " + strconv.Quote(strings.Replace(modoDeLaTanda,
				"'workflow_dispatch' || ", "'workflow_dispatch' || github.event.label.name == 'evals' || ", 1)) +
				yLoEsperadoEs + strconv.Quote(modoDeLaTanda)},
		},
		{
			nombre:  "tanda-sin-el-modo",
			cambios: []cambioDeLaDefinicion{{"      MODO_DE_EVALS: >-\n", "      OTRO_MODO: >-\n"}},
			fallos:  []string{"jobs.tanda.env.MODO_DE_EVALS: no est\xc3\xa1" + yLoEsperadoEs + strconv.Quote(modoDeLaTanda)},
		},
		{
			nombre:  "decidir-sin-la-ejecucion",
			cambios: []cambioDeLaDefinicion{{ejecucionDeGh, ""}},
			fallos: []string{"jobs.tanda.steps, id decidir, run: vale " +
				strconv.Quote(strings.Replace(ordenDelPasoQueDecide, ejecucionDeGh, "", 1)) + ", y lo esperado es " +
				strconv.Quote(ordenDelPasoQueDecide)},
		},
		{
			// Sin -modo, TestTandaDelCommit no decide (contracts/job-de-evals.md §4
			// de H30).
			nombre:  "decidir-sin-el-modo",
			cambios: []cambioDeLaDefinicion{{modoDeGh, ""}},
			fallos: []string{"jobs.tanda.steps, id decidir, run: vale " +
				strconv.Quote(strings.Replace(ordenDelPasoQueDecide, modoDeGh, "", 1)) + yLoEsperadoEs +
				strconv.Quote(ordenDelPasoQueDecide)},
		},
		{
			nombre:  "decidir-sin-su-id",
			cambios: []cambioDeLaDefinicion{{"        id: decidir\n", ""}},
			fallos: []string{"jobs.tanda.steps, id decidir, run: no est\xc3\xa1, y lo esperado es " +
				strconv.Quote(ordenDelPasoQueDecide)},
		},
		{
			nombre:  "marca-con-otro-nombre",
			cambios: []cambioDeLaDefinicion{{"      - name: " + pasoDeLaMarca + "\n", "      - name: Mide el commit\n"}},
			fallos: []string{elPenultimo + `name: vale "Mide el commit", y lo esperado es "Esta ejecuci` + "\xc3\xb3" +
				`n mide el commit"`},
		},
		{
			nombre:  "marca-sin-su-if",
			cambios: []cambioDeLaDefinicion{{"        if: steps.decidir.outputs.medir == 'si'\n", ""}},
			fallos: []string{elPenultimo + `if: no est` + "\xc3\xa1" + `, y lo esperado es ` +
				`"steps.decidir.outputs.medir == 'si'"`},
		},
		{
			nombre: "marca-de-la-suite-completa-con-otro-nombre",
			cambios: []cambioDeLaDefinicion{{
				"      - name: " + pasoDeLaSuiteCompleta + "\n", "      - name: Mide la suite completa\n",
			}},
			fallos: []string{elUltimo + `name: vale "Mide la suite completa"` + yLoEsperadoEs +
				strconv.Quote(marcaDeLaSuiteCompleta)},
		},
		{
			// Sin el modo en su if, la marca de la suite completa correría también
			// en una ejecución con reutilización, y la tanda de otra ejecución de la
			// suite completa no mediría (FR-021 de H30).
			nombre:  "marca-de-la-suite-completa-sin-el-modo",
			cambios: []cambioDeLaDefinicion{{" && steps.decidir.outputs.modo == 'completa'", ""}},
			fallos: []string{elUltimo + `if: vale "steps.decidir.outputs.medir == 'si'"` + yLoEsperadoEs +
				strconv.Quote(condicionDeLaMarcaDeLaSuiteCompleta)},
		},
		{
			// La de antes de H30: la tanda con una sola marca, que ya no es la
			// última.
			nombre: "tanda-sin-la-marca-de-la-suite-completa",
			cambios: []cambioDeLaDefinicion{{
				"      - name: " + pasoDeLaSuiteCompleta + "\n" +
					"        if: " + condicionDeLaMarcaDeLaSuiteCompleta + "\n        run: echo mide la suite completa\n",
				"",
			}},
			fallos: []string{
				elPenultimo + `name: vale "Mirar si otra tanda mide este commit"` + yLoEsperadoEs +
					strconv.Quote(marcaDeLaTanda),
				elPenultimo + `if: no est` + "\xc3\xa1" + yLoEsperadoEs + strconv.Quote(condicionDeLaMarca),
				elUltimo + `name: vale "Esta ejecuci` + "\xc3\xb3" + `n mide el commit"` + yLoEsperadoEs +
					strconv.Quote(marcaDeLaSuiteCompleta),
				elUltimo + `if: vale "steps.decidir.outputs.medir == 'si'"` + yLoEsperadoEs +
					strconv.Quote(condicionDeLaMarcaDeLaSuiteCompleta),
			},
		},
		{
			nombre:  "evals-sin-needs-tanda",
			cambios: []cambioDeLaDefinicion{{"needs: [tanda]", "needs: [cambios]"}},
			fallos:  []string{"jobs.evals.needs: vale [cambios], y lo esperado es [tanda]"},
		},
		{
			nombre:  "evals-con-otro-if",
			cambios: []cambioDeLaDefinicion{{condicionDelTrabajo, "${{ " + sinCancelar + " }}"}},
			fallos: []string{`jobs.evals.if: vale "${{ ` + sinCancelar + ` }}", y lo esperado es ` +
				`"${{ ` + sinCancelar + ` && needs.tanda.outputs.medir == 'si' }}"`},
		},
		{
			nombre:  "grupo-sin-la-cabeza-del-evento",
			cambios: []cambioDeLaDefinicion{{grupoPorCommit, "group: evals-${{ github.sha }}-${{ matrix.skill }}"}},
			fallos: []string{`jobs.evals.concurrency.group: vale "evals-${{ github.sha }}-${{ matrix.skill }}", ` +
				"y lo esperado es " + esperadoDelGrupo},
		},
		{
			nombre:  "grupo-sin-la-skill",
			cambios: []cambioDeLaDefinicion{{grupoPorCommit, "group: evals-${{ github.sha }}"}},
			fallos: []string{`jobs.evals.concurrency.group: vale "evals-${{ github.sha }}", y lo esperado es ` +
				esperadoDelGrupo},
		},
		{
			nombre:  "cancela-la-que-corre",
			cambios: []cambioDeLaDefinicion{{cancelaEnCurso, "      cancel-in-progress: true\n"}},
			fallos:  []string{"jobs.evals.concurrency.cancel-in-progress: vale true, y lo esperado es false"},
		},
		{
			nombre:  "sin-cancel-in-progress",
			cambios: []cambioDeLaDefinicion{{cancelaEnCurso, ""}},
			fallos:  []string{"jobs.evals.concurrency.cancel-in-progress: no est\xc3\xa1, y lo esperado es false"},
		},
		{
			nombre:  "concurrency-de-flujo",
			cambios: []cambioDeLaDefinicion{{"name: evals\n", "name: evals\nconcurrency: evals-${{ github.sha }}\n"}},
			fallos: []string{"concurrency: est\xc3\xa1 en el flujo, y lo esperado es que no est\xc3\xa9: la tanda " +
				"por commit la da el del trabajo evals, por skill"},
		},
		{
			nombre:  "sin-nombre",
			cambios: []cambioDeLaDefinicion{{"    name: evals (${{ matrix.skill }})\n", ""}},
			fallos:  []string{"jobs.evals.name: no est\xc3\xa1, y lo esperado es \"evals (${{ matrix.skill }})\""},
		},
		{
			nombre:  "matriz-sin-legal-core",
			cambios: []cambioDeLaDefinicion{{matrizDelJob, "skill: [normativa, jurisprudencia]"}},
			fallos: []string{"jobs.evals.strategy.matrix.skill: vale [normativa, jurisprudencia], y lo esperado es " +
				matrizEsperada},
		},
		{
			// La matriz de antes de H23: sin jurisprudencia, el job no la mide.
			nombre:  "matriz-sin-jurisprudencia",
			cambios: []cambioDeLaDefinicion{{matrizDelJob, "skill: [normativa, legal-core]"}},
			fallos: []string{"jobs.evals.strategy.matrix.skill: vale [normativa, legal-core], y lo esperado es " +
				matrizEsperada},
		},
		{
			nombre:  "otra-concurrencia",
			cambios: []cambioDeLaDefinicion{{"concurrencia: 5", "concurrencia: 2"}},
			fallos: []string{includeDeSkill + "normativa: vale " +
				"{concurrencia: 2, objetivo_de_duracion: 900}, y lo esperado es " + ajustesDeBoe},
		},
		{
			nombre:  "otro-objetivo",
			cambios: []cambioDeLaDefinicion{{includeDeLegal, entradaDeLegal + conObjetivo}},
			fallos: []string{includeDeSkill + "legal-core: vale " +
				"{concurrencia: 2, objetivo_de_duracion: 900}, y lo esperado es " + ajustesDeLegal},
		},
		{
			// jurisprudencia va sin objetivo de duración, como legal-core: no hay
			// medida de la que sacar otra cifra (research.md D22 y S4 de H23).
			nombre:  "jurisprudencia-con-objetivo",
			cambios: []cambioDeLaDefinicion{{includeDeJuris, entradaDeJuris + conObjetivo}},
			fallos: []string{includeDeSkill + "jurisprudencia: vale " +
				"{concurrencia: 4, objetivo_de_duracion: 900}, y lo esperado es " + ajustesDeJuris},
		},
		{
			// jurisprudencia va a cuatro a la vez desde H25. Con la eval de
			// evalsSinteticas, de una en una son 6661 s —los 4021 de sus sesiones,
			// los 900 de la decisión, 60 + 36 × 40 de su juez y, desde H31, 6 × 40
			// de su descripción—, que el tope cubre:
			// la única línea es la de su include. La del tope, con las diez evals del
			// repositorio, la da probarElTopeConLasEvalsDelRepositorio.
			nombre:  "jurisprudencia-de-una-en-una",
			cambios: []cambioDeLaDefinicion{{includeDeJuris, skillDeJuris + deUnaEnUna + sinObjetivo}},
			fallos: []string{includeDeSkill + "jurisprudencia: vale " +
				"{concurrencia: 1, objetivo_de_duracion: 0}, y lo esperado es " + ajustesDeJuris},
		},
		{
			nombre:  "include-sin-legal-core",
			cambios: []cambioDeLaDefinicion{{includeDeLegal, ""}},
			fallos: []string{
				includeDeSkill + "legal-core: no est\xc3\xa1, y lo esperado es " + ajustesDeLegal,
				topeDeEvals + "el peor caso de legal-core no se puede obtener: la concurrencia de " +
					"legal-core es 0 y tiene que ser un entero mayor o igual que 1",
			},
		},
		{
			nombre:  "include-sin-jurisprudencia",
			cambios: []cambioDeLaDefinicion{{includeDeJuris, ""}},
			fallos: []string{
				includeDeSkill + "jurisprudencia: no est\xc3\xa1, y lo esperado es " + ajustesDeJuris,
				topeDeEvals + "el peor caso de jurisprudencia no se puede obtener: la concurrencia de " +
					"jurisprudencia es 0 y tiene que ser un entero mayor o igual que 1",
			},
		},
		{
			// Una skill que solo está en include es otra combinación, con sus
			// sesiones a la vez además de las once de las tres (SC-011 de H30).
			nombre:  "include-con-otra-skill",
			cambios: []cambioDeLaDefinicion{{includeDeLegal, includeDeLegal + "          - skill: cita\n            concurrencia: 1\n"}},
			fallos: []string{
				includeDeSkill + "cita: vale {concurrencia: 1, objetivo_de_duracion: 0}, y lo esperado es que no est\xc3\xa9",
				sesionesDeLaMatriz + "admite 12 sesiones a la vez entre sus 4 combinaciones" + yLoEsperadoEs +
					"como mucho 11",
			},
		},
		{
			// Una sesión más en una skill son doce entre las tres (FR-020 y SC-011 de
			// H30): la línea de su include y la de las sesiones a la vez.
			nombre:  "una-sesion-mas-en-normativa",
			cambios: []cambioDeLaDefinicion{{"concurrencia: 5", "concurrencia: 6"}},
			fallos: []string{
				includeDeSkill + "normativa: vale {concurrencia: 6, objetivo_de_duracion: 900}" + yLoEsperadoEs +
					ajustesDeBoe,
				sesionesDeLaMatriz + "admite 12 sesiones a la vez entre sus 3 combinaciones" + yLoEsperadoEs +
					"como mucho 11",
			},
		},
		{
			// Otra clave en la matriz multiplica los trabajos de cada skill: dos a
			// la vez de la misma skill, cada uno con su concurrencia (SC-011 de
			// H30).
			nombre:  "matriz-con-otra-dimension",
			cambios: []cambioDeLaDefinicion{{matrizDelJob, matrizDelJob + "\n        modo: [orden, herramienta]"}},
			fallos: []string{"jobs.evals.strategy.matrix.modo: est\xc3\xa1" + yLoEsperadoEs + queNoEste +
				": la matriz da un solo trabajo por skill"},
		},
		{
			nombre:  "env-sin-la-concurrencia",
			cambios: []cambioDeLaDefinicion{{concurrenciaEnv, ""}},
			fallos: []string{"jobs.evals.env.CONCURRENCIA_DE_EVALS: no est\xc3\xa1, y lo esperado es " +
				`"${{ matrix.concurrencia }}"`},
		},
		{
			nombre:  "env-con-el-objetivo-escrito",
			cambios: []cambioDeLaDefinicion{{objetivoEnv, "OBJETIVO_DE_DURACION_DE_EVALS: 900"}},
			fallos: []string{`jobs.evals.env.OBJETIVO_DE_DURACION_DE_EVALS: vale "900", y lo esperado es ` +
				`"${{ matrix.objetivo_de_duracion }}"`},
		},
		{
			// 59 minutos son 3540 s, menos que el peor caso de legal-core, 3589 s, y
			// cubren el de normativa, 3093 s, y el de jurisprudencia, 3173 s.
			// Desde H30, todos llevan los 900 s de la decisión y, desde H31, los
			// 240 s de los votos de su descripción; el de legal-core, que no tiene
			// juez de las respuestas, también los 60 s de instalar el Claude Code
			// del juez, sin los que 59 minutos lo cubrirían (research D10 de H31).
			nombre:  "tope-por-debajo-del-peor-caso",
			cambios: []cambioDeLaDefinicion{{topeDelJob, "timeout-minutes: 59"}},
			fallos:  []string{topeDeEvals + "vale 59 (3540 s)" + alMenosElPeorDe + "legal-core, " + peorCasoDeLegal},
		},
		{
			// 41 minutos son 2460 s, menos que los 2473 s de las sesiones solas de
			// una skill a cuatro a la vez con la decisión: no cubren el peor caso de
			// ninguna de las tres.
			nombre:  "tope-por-debajo-de-los-tres",
			cambios: []cambioDeLaDefinicion{{topeDelJob, "timeout-minutes: 41"}},
			fallos: []string{
				topeDeEvals + "vale 41 (2460 s)" + alMenosElPeorDe + "normativa, " + peorCasoDeBoe,
				topeDeEvals + "vale 41 (2460 s)" + alMenosElPeorDe + "jurisprudencia, " + peorCasoConJuez,
				topeDeEvals + "vale 41 (2460 s)" + alMenosElPeorDe + "legal-core, " + peorCasoDeLegal,
			},
		},
		{
			// 47 minutos son 2820 s, menos que el peor caso de normativa con
			// su juez, 3093 s, y que el de jurisprudencia, 3173 s: cubrirían el de
			// sus sesiones solas con la decisión y su descripción, 2713 s, y no
			// cubren el que cuenta sus votos.
			nombre:  "tope-por-debajo-del-de-la-skill-con-juez",
			cambios: []cambioDeLaDefinicion{{topeDelJob, "timeout-minutes: 47"}},
			fallos: []string{
				topeDeEvals + "vale 47 (2820 s)" + alMenosElPeorDe + "normativa, " + peorCasoDeBoe,
				topeDeEvals + "vale 47 (2820 s)" + alMenosElPeorDe + "jurisprudencia, " + peorCasoConJuez,
				topeDeEvals + "vale 47 (2820 s)" + alMenosElPeorDe + "legal-core, " + peorCasoDeLegal,
			},
		},
		{
			// 50 minutos son 3000 s: cubrirían el de las evals de normativa
			// con su juez, 2853 s, y el de jurisprudencia, 2933 s, y no cubren lo
			// que suman con los 240 s de su descripción (SC-015 de H31).
			nombre:  "tope-por-debajo-del-de-la-descripcion",
			cambios: []cambioDeLaDefinicion{{topeDelJob, "timeout-minutes: 50"}},
			fallos: []string{
				topeDeEvals + "vale 50 (3000 s)" + alMenosElPeorDe + "normativa, " + peorCasoDeBoe,
				topeDeEvals + "vale 50 (3000 s)" + alMenosElPeorDe + "jurisprudencia, " + peorCasoConJuez,
				topeDeEvals + "vale 50 (3000 s)" + alMenosElPeorDe + "legal-core, " + peorCasoDeLegal,
			},
		},
		{
			// 53 minutos, 3180 s, cubren el de normativa y el de
			// jurisprudencia con su juez y su descripción, y no el de legal-core.
			nombre:  "tope-que-cubre-el-de-la-skill-con-juez",
			cambios: []cambioDeLaDefinicion{{topeDelJob, "timeout-minutes: 53"}},
			fallos:  []string{topeDeEvals + "vale 53 (3180 s)" + alMenosElPeorDe + "legal-core, " + peorCasoDeLegal},
		},
		{
			nombre:  "tope-que-cubre-el-peor-caso",
			cambios: []cambioDeLaDefinicion{{topeDelJob, "timeout-minutes: 60"}},
		},
		{
			// Con una repetición, 3 sesiones por skill en el modo orden y 2 en el
			// modo herramienta: 485 + 900 + (⌈3 / 2⌉ + ⌈2 / 2⌉) × 272 + 60 + 6 × 40
			// = 2501 s en legal-core, que va de dos en dos y tiene el mayor de los
			// tres, que 42 minutos cubren.
			nombre: "tope-con-las-repeticiones-del-env",
			cambios: []cambioDeLaDefinicion{
				{topeDelJob, "timeout-minutes: 42"},
				{repeticionesEnv, "REPETICIONES_DE_EVALS: 1"},
			},
		},
		{
			// Sin modelos informativos, 4 sesiones por skill en el modo orden y 3
			// en el modo herramienta: 485 + 900 + (⌈4 / 2⌉ + ⌈3 / 2⌉) × 272 + 60 +
			// 6 × 40 = 2773 s en legal-core, que va de dos en dos y tiene el mayor
			// de los tres, que 47 minutos cubren.
			nombre: "tope-con-los-modelos-del-env",
			cambios: []cambioDeLaDefinicion{
				{topeDelJob, "timeout-minutes: 47"},
				{"MODELOS_INFORMATIVOS_DE_EVALS: claude-haiku-4-5-20251001", `MODELOS_INFORMATIVOS_DE_EVALS: ""`},
			},
		},
	}
}

// definicionesSinteticasDeLaEleccion son las definiciones que se apartan de
// contracts/job-de-evals.md §2 de H30 en el trabajo cambios y en los permisos,
// el env y los pasos que el trabajo evals gana con la decisión, cada una con su
// línea (FR-010, FR-012, FR-016, FR-019; research D5, D8 y D17 de H30).
func definicionesSinteticasDeLaEleccion() []definicionSintetica {
	const (
		yLoEsperadoEs         = ", y lo esperado es "
		deCambios             = "jobs.cambios"
		delElegir             = "jobs.cambios.steps, id elegir, "
		deLosPasos            = "jobs.evals.steps"
		todaLaHist            = "          fetch-depth: 0\n"
		accesoDeGh            = "GH_TOKEN: ${{ secrets.GITHUB_TOKEN }}\n"
		conElAccesoDeGh       = "          " + accesoDeGh
		conRepoDeGh           = "          GH_REPO: ${{ github.repository }}\n"
		modoEnElTrabajo       = "      MODO_DE_EVALS: ${{ needs.tanda.outputs.modo }}\n"
		suscripcionDelPaso    = "          CLAUDE_CODE_OAUTH_TOKEN: ${{ secrets.CLAUDE_CODE_OAUTH_TOKEN }}\n"
		permisosDeEvals       = "      contents: read\n      actions: read\n      pull-requests: read\n"
		elDeLaDecision        = deLosPasos + ", \"Decidir qu\xc3\xa9 mide y qu\xc3\xa9 reutiliza\""
		laMarcaFinal          = deLosPasos + ", el \xc3\xbaltimo, "
		ejecutarConLaDecision = "        run: " + ordenQueEjecutaLasEvals + "\n"
	)

	// La decisión que va detrás del paso que ejecuta las evals, en lugar de
	// delante: el paso de la instalación y el que ejecuta las evals, y después el
	// de la decisión.
	pasosDeEvals := strings.SplitN(pasosDeEvalsDelContrato, pasoDeLaDecisionDelContrato, 2)
	conLaDecisionDetras := pasoDeLaDecisionDelContrato + pasosDeEvals[1]
	detrasDeLaEjecucion := strings.Replace(pasosDeEvals[1], "      - name: "+marcaDelInformeEscrito+"\n",
		pasoDeLaDecisionDelContrato+"      - name: "+marcaDelInformeEscrito+"\n", 1)

	return []definicionSintetica{
		{
			nombre:  "sin-cambios",
			cambios: []cambioDeLaDefinicion{{trabajoDeCambiosDelContrato, ""}},
			fallos: []string{deCambios + ": " + noEsta + yLoEsperadoEs +
				"el trabajo que elige las evals que la propuesta ha podido cambiar"},
		},
		{
			// cambios no corre con la etiqueta: la etiqueta evals mide aunque la
			// propuesta no elija nada (FR-017 de H30).
			nombre:  "cambios-tambien-con-la-etiqueta",
			cambios: []cambioDeLaDefinicion{enCambios(" && github.event.action != 'labeled'", "")},
			fallos: []string{deCambios + `.if: vale "github.event_name == 'pull_request'"` + yLoEsperadoEs +
				strconv.Quote(condicionDeCambios)},
		},
		{
			// Los 5 minutos de antes de H30 no cubren la elección, con sus 12
			// minutos de go test.
			nombre:  "cambios-con-el-tope-de-antes",
			cambios: []cambioDeLaDefinicion{enCambios("    timeout-minutes: 15\n", "    timeout-minutes: 5\n")},
			fallos:  []string{deCambios + ".timeout-minutes: vale 5" + yLoEsperadoEs + "15"},
		},
		{
			nombre:  "cambios-sin-pull-requests-read",
			cambios: []cambioDeLaDefinicion{enCambios("      pull-requests: read\n", "")},
			fallos: []string{deCambios + ".permissions: vale {contents: read}" + yLoEsperadoEs +
				"{contents: read, pull-requests: read}"},
		},
		{
			nombre:  "cambios-con-permiso-de-escritura",
			cambios: []cambioDeLaDefinicion{enCambios("      contents: read\n", "      contents: write\n")},
			fallos: []string{deCambios + ".permissions: vale {contents: write, pull-requests: read}" + yLoEsperadoEs +
				"{contents: read, pull-requests: read}"},
		},
		{
			// La salida de antes de H30, coincide, no la lee la tanda.
			nombre: "cambios-con-la-salida-de-antes",
			cambios: []cambioDeLaDefinicion{enCambios("      elige: ${{ steps.elegir.outputs.elige }}\n",
				"      coincide: ${{ steps.mirar.outputs.coincide }}\n")},
			fallos: []string{deCambios + ".outputs.elige: " + noEsta + yLoEsperadoEs + strconv.Quote(salidaDeCambios)},
		},
		{
			// Sin toda la historia no está el árbol de la base (research D5 de H30).
			nombre:  "cambios-sin-toda-la-historia",
			cambios: []cambioDeLaDefinicion{enCambios(todaLaHist, "")},
			fallos: []string{deCambios + ".steps, el de " + accionDelCheckout + ", with.fetch-depth: " + noEsta +
				yLoEsperadoEs + todaLaHistoria},
		},
		{
			nombre:  "cambios-con-un-solo-commit",
			cambios: []cambioDeLaDefinicion{enCambios(todaLaHist, "          fetch-depth: 1\n")},
			fallos: []string{deCambios + ".steps, el de " + accionDelCheckout + `, with.fetch-depth: vale "1"` +
				yLoEsperadoEs + todaLaHistoria},
		},
		{
			nombre: "elegir-con-otra-prueba",
			cambios: []cambioDeLaDefinicion{enCambios("'^TestEleccionDeLaPropuesta$'",
				"'^TestDecisionDelTrabajo$'")},
			fallos: []string{delElegir + "run: vale " + strconv.Quote(strings.Replace(ordenDeLaEleccion,
				"TestEleccionDeLaPropuesta", "TestDecisionDelTrabajo", 1)) + yLoEsperadoEs + strconv.Quote(ordenDeLaEleccion)},
		},
		{
			nombre:  "elegir-sin-el-token-de-gh",
			cambios: []cambioDeLaDefinicion{enCambios(conElAccesoDeGh, "")},
			fallos: []string{delElegir + "env: vale {BASE_DE_LA_PROPUESTA: ${{ github.event.pull_request.base.sha }}, " +
				"GH_REPO: ${{ github.repository }}, PROPUESTA: ${{ github.event.pull_request.number }}}" + yLoEsperadoEs +
				presentarMapa(entornoDeLaEleccion)},
		},
		{
			nombre:  "elegir-con-otro-nombre",
			cambios: []cambioDeLaDefinicion{enCambios("      - name: "+pasoDeLaEleccion+"\n", "      - name: Elegir\n")},
			fallos:  []string{delElegir + `name: vale "Elegir"` + yLoEsperadoEs + strconv.Quote(pasoDeLaEleccion)},
		},
		{
			// El de antes de H30: solo con contents.
			nombre:  "evals-sin-sus-permisos",
			cambios: []cambioDeLaDefinicion{{permisosDeEvals, "      contents: read\n"}},
			fallos: []string{"jobs.evals.permissions: vale {contents: read}" + yLoEsperadoEs +
				presentarMapa(permisosDelTrabajo)},
		},
		{
			nombre:  "evals-sin-el-modo",
			cambios: []cambioDeLaDefinicion{{"      MODO_DE_EVALS: ${{ needs.tanda.outputs.modo }}\n", ""}},
			fallos: []string{"jobs.evals.env.MODO_DE_EVALS: " + noEsta + yLoEsperadoEs +
				`"${{ needs.tanda.outputs.modo }}"`},
		},
		{
			nombre:  "evals-sin-toda-la-historia",
			cambios: []cambioDeLaDefinicion{{todaLaHist, ""}},
			fallos: []string{deLosPasos + ", el de " + accionDelCheckout + ", with.fetch-depth: " + noEsta +
				yLoEsperadoEs + todaLaHistoria},
		},
		{
			// Sin la decisión, el job lo mediría todo, y scripts/evals.sh no
			// tendría su fichero (FR-012 y research D18 de H30).
			nombre:  "sin-la-decision",
			cambios: []cambioDeLaDefinicion{{pasoDeLaDecisionDelContrato, ""}},
			fallos: []string{elDeLaDecision + ": " + noEsta + yLoEsperadoEs +
				"el paso que decide qu\xc3\xa9 mide y qu\xc3\xa9 reutiliza"},
		},
		{
			// Antes de instalar Go no hay go con el que decidir.
			nombre: "decision-antes-de-instalar-go",
			cambios: []cambioDeLaDefinicion{{
				pasosDeEvals[0] + pasoDeLaDecisionDelContrato,
				strings.Replace(pasosDeEvals[0], "      - name: Instalar Go\n", pasoDeLaDecisionDelContrato+
					"      - name: Instalar Go\n", 1),
			}},
			fallos: []string{elDeLaDecision + ": est\xc3\xa1 en el paso 2" + yLoEsperadoEs +
				"que vaya tras el de " + accionDeGo + ` y antes de "Ejecutar las evals"`},
		},
		{
			// Tras ejecutar las evals, la decisión no llega a ellas.
			nombre:  "decision-tras-ejecutar-las-evals",
			cambios: []cambioDeLaDefinicion{{conLaDecisionDetras, detrasDeLaEjecucion}},
			fallos: []string{elDeLaDecision + ": est\xc3\xa1 en el paso 5" + yLoEsperadoEs +
				"que vaya tras el de " + accionDeGo + ` y antes de "Ejecutar las evals"`},
		},
		{
			// La decisión sin su plazo no tiene peor caso: go test sin -timeout
			// espera sus 10 minutos por omisión, y el tope no sabría qué sumar
			// (research D17 de H30).
			nombre:  "decision-sin-su-plazo",
			cambios: []cambioDeLaDefinicion{{"-count=1 -timeout 15m -v", "-count=1 -v"}},
			fallos: slices.Concat(
				[]string{elDeLaDecision + ", run: vale " + strconv.Quote(strings.Replace(ordenDeLaDecision,
					"-timeout 15m ", "", 1)) + yLoEsperadoEs + strconv.Quote(ordenDeLaDecision)},
				lineasSinElPlazo("jobs.evals.timeout-minutes: el peor caso de %s no se puede obtener: el paso "+
					"\"Decidir qu\xc3\xa9 mide y qu\xc3\xa9 reutiliza\" no fija su plazo con -timeout"),
			),
		},
		{
			nombre:  "decision-sin-el-token-de-gh",
			cambios: []cambioDeLaDefinicion{{"        env:\n" + conElAccesoDeGh + conRepoDeGh + "        run: >-\n", "        run: >-\n"}},
			fallos:  []string{elDeLaDecision + ", env: " + noEsta + yLoEsperadoEs + presentarMapa(entornoDeLaDecision)},
		},
		{
			// GH_TOKEN en el env del trabajo lo verían todos sus pasos, también las
			// sesiones con modelo (contracts/job-de-evals.md §2 de H30).
			nombre:  "token-de-gh-en-el-trabajo",
			cambios: []cambioDeLaDefinicion{{modoEnElTrabajo, modoEnElTrabajo + "      " + accesoDeGh}},
			fallos: []string{"jobs.evals.env.GH_TOKEN: est\xc3\xa1" + yLoEsperadoEs + queNoEste +
				": solo el paso de la decisi\xc3\xb3n lee la plataforma"},
		},
		{
			nombre:  "token-de-gh-al-ejecutar-las-evals",
			cambios: []cambioDeLaDefinicion{{suscripcionDelPaso, suscripcionDelPaso + conElAccesoDeGh}},
			fallos: []string{deLosPasos + `, "Ejecutar las evals", env.GH_TOKEN: est` + "\xc3\xa1" + yLoEsperadoEs +
				queNoEste + ": solo el paso de la decisi\xc3\xb3n lee la plataforma"},
		},
		{
			// Sin DECISION_DE_EVALS, scripts/evals.sh termina con 1 antes de la
			// primera sesión (research D18 de H30).
			nombre:  "ejecutar-sin-la-decision",
			cambios: []cambioDeLaDefinicion{{ejecutarConLaDecision, "        run: make evals SKILL=\"$SKILL_EVALUADA\" 2>&1\n"}},
			fallos: []string{deLosPasos + `, "Ejecutar las evals", run: vale "make evals SKILL=\"$SKILL_EVALUADA\" 2>&1"` +
				yLoEsperadoEs + strconv.Quote(ordenQueEjecutaLasEvals)},
		},
		{
			// La marca solo con success() no correría con las evals en rojo, y la
			// búsqueda de medidas de main no vería su informe (FR-010 de H30).
			nombre:  "marca-del-informe-sin-cancelled",
			cambios: []cambioDeLaDefinicion{{"        if: " + condicionDeLaMarcaDelInforme + "\n", ""}},
			fallos:  []string{laMarcaFinal + "if: " + noEsta + yLoEsperadoEs + strconv.Quote(condicionDeLaMarcaDelInforme)},
		},
		{
			nombre:  "sin-la-marca-del-informe",
			cambios: []cambioDeLaDefinicion{{"      - name: " + marcaDelInformeEscrito + "\n", "      - name: Otra marca\n"}},
			fallos:  []string{laMarcaFinal + `name: vale "Otra marca"` + yLoEsperadoEs + strconv.Quote(marcaDelInformeEscrito)},
		},
		{
			nombre:  "marca-del-informe-que-no-mira-su-contenido",
			cambios: []cambioDeLaDefinicion{{"test -s \"$RUNNER_TEMP", "test -e \"$RUNNER_TEMP"}},
			fallos: []string{laMarcaFinal + "run: vale " + strconv.Quote(strings.Replace(ordenDeLaMarcaDelInforme,
				"-s", "-e", 1)) + yLoEsperadoEs + strconv.Quote(ordenDeLaMarcaDelInforme)},
		},
	}
}

// lineasSinElPlazo son las líneas del tope del trabajo evals de las tres
// skills de la matriz, en su orden, con el formato dado, que lleva el %s de la
// skill.
func lineasSinElPlazo(formato string) []string {
	lineas := make([]string, 0, len(skillsDelTrabajo))
	for _, skill := range skillsDelTrabajo {
		lineas = append(lineas, fmt.Sprintf(formato, skill))
	}

	return lineas
}

// skillConJuezSintetico es la skill de evalsSinteticas cuyo juez cambian los
// tests de los errores, y casosDelJuezSintetico, los casos etiquetados del juez
// de cada skill que lo tiene.
const (
	skillConJuezSintetico = "normativa"
	casosDelJuezSintetico = 7
)

// skillsConJuezSintetico son las skills de evalsSinteticas que tienen juez, en
// el orden de la matriz: las dos que lo tienen en el repositorio desde H25
// (contracts/job-de-evals.md §4 de H25).
var skillsConJuezSintetico = []string{skillConJuezSintetico, "jurisprudencia"}

// evalsSinteticas crea un directorio temporal de evals con una carpeta por
// skill de la matriz, cada una con la eval sintética del art. 21 de la LPAC, y
// devuelve su ruta. Las de normativa y jurisprudencia llevan además la
// carpeta del juez, con siete casos etiquetados cada una; la de legal-core, no:
// es la skill sin juez. Desde H31, el directorio de evals es el evals de una
// raíz temporal, como en el repositorio, que lleva en la evidencia del juez de
// las descripciones siete casos etiquetados suyos: los de la medida de la
// entrada descripciones del trabajo medida (research D9 de H31).
func evalsSinteticas(t *testing.T) string {
	t.Helper()

	raiz := t.TempDir()
	evals := filepath.Join(raiz, directorioDeLasEvals)
	require.NoError(t, os.Mkdir(evals, 0o750))

	evidencia := carpetaDeLaEvidenciaEn(raiz)
	require.NoError(t, os.MkdirAll(evidencia, 0o750))
	require.NoError(t, os.WriteFile(filepath.Join(evidencia, ficheroDeCasosDelJuez),
		[]byte(casosSinteticosDeLaEvidencia(casosDelJuezSintetico)), 0o600))

	for _, skill := range skillsDelTrabajo {
		require.NoError(t, os.Mkdir(filepath.Join(evals, skill), 0o750))
		require.NoError(t, os.WriteFile(filepath.Join(evals, skill, nombreDeEval), []byte(contenidoDelArticulo21), 0o600))
	}

	var casos strings.Builder

	casos.WriteString("clase: afirma_lo_no_leido\ncasos:\n")

	for numero := range casosDelJuezSintetico {
		fmt.Fprintf(&casos, "  - informe: specs/sintetico/informe.json\n    sesion: sesion-%02d\n"+
			"    grupo: medida\n    etiqueta: defecto\n    procedencia: lectura\n", numero+1)
	}

	for _, skill := range skillsConJuezSintetico {
		conJuez := filepath.Join(evals, skill)
		escribirLaCarpetaDelJuez(t, conJuez)
		crearEntradas(t, conJuez, []entradaDeConjunto{{nombre: casosEnElJuez, contenido: casos.String()}})
	}

	leidos, err := leerCasosEtiquetados([]byte(casos.String()))
	require.NoError(t, err)
	require.Len(t, leidos.Casos, casosDelJuezSintetico, "premisa: los casos etiquetados del juez sint\xc3\xa9tico")

	return evals
}

// casosSinteticosDeLaEvidencia son tantos casos etiquetados del juez de las
// descripciones, con la forma de casos.yaml de su evidencia: cada uno con su
// id, una skill de la matriz, su etiqueta y una descripción de una línea.
func casosSinteticosDeLaEvidencia(cuantos int) string {
	var casos strings.Builder

	casos.WriteString("casos:\n")

	for numero := range cuantos {
		fmt.Fprintf(&casos, "- id: '%03d'\n  skill: %s\n  etiqueta: defecto\n  descripcion: Consulta normativa.\n",
			numero+1, skillsDelTrabajo[numero%len(skillsDelTrabajo)])
	}

	return casos.String()
}

// probarLosErroresDeLaDefinicion fija lo que no se puede leer ni obtener: una
// definición que no existe, que no es YAML o cuyas repeticiones no son un
// entero no se lee, con un error que nombra la ruta; y el peor caso de una skill
// sin su carpeta de evals o con un plan que no se puede componer no se obtiene,
// con un error que dice por qué, ni el de la medida de una skill sin su carpeta
// de evals o con unos casos etiquetados que no tienen su forma, ni, desde H31, el
// de la medida de las descripciones sin los casos de su evidencia o con unos que
// no tienen su forma.
func probarLosErroresDeLaDefinicion(t *testing.T) {
	t.Parallel()

	directorio := t.TempDir()

	casos := []struct {
		nombre    string
		contenido string
		fragmento string
	}{
		{nombre: "no-es-yaml", contenido: "jobs: [\n", fragmento: "no es YAML"},
		{
			nombre:    "repeticiones-que-no-son-un-entero",
			contenido: strings.Replace(definicionDelContrato, "REPETICIONES_DE_EVALS: 3", "REPETICIONES_DE_EVALS: tres", 1),
			fragmento: `jobs.evals.env.REPETICIONES_DE_EVALS vale "tres" y no es un entero`,
		},
	}

	for _, caso := range casos {
		ruta := filepath.Join(directorio, caso.nombre+".yml")
		require.NoError(t, os.WriteFile(ruta, []byte(caso.contenido), 0o600))

		_, err := leerDefinicionDelJob(ruta)
		require.ErrorContains(t, err, ruta, caso.nombre)
		require.ErrorContains(t, err, caso.fragmento, caso.nombre)
	}

	sinFichero := filepath.Join(directorio, "no-existe.yml")
	_, err := leerDefinicionDelJob(sinFichero)
	require.ErrorIs(t, err, os.ErrNotExist)
	require.ErrorContains(t, err, sinFichero)

	ruta := filepath.Join(directorio, "evals.yml")
	require.NoError(t, os.WriteFile(ruta, []byte(definicionDelContrato), 0o600))

	leida, err := leerDefinicionDelJob(ruta)
	require.NoError(t, err)

	_, err = leida.peorCaso(t.TempDir(), "normativa")
	require.ErrorIs(t, err, os.ErrNotExist, "sin la carpeta de evals de la skill")

	evals := evalsSinteticas(t)

	leida.ModeloQueDecide = "Claude Sonnet"
	_, err = leida.peorCaso(evals, "normativa")
	require.ErrorContains(t, err, `el modelo que decide "Claude Sonnet" no tiene la forma de un id de modelo`)

	// El peor caso de la medida tampoco se obtiene sin la carpeta de evals de la
	// skill ni con unos casos etiquetados que no tienen su forma: su error nombra
	// el fichero.
	require.NotNil(t, leida.Medida, "premisa: la definici\xc3\xb3n del contrato tiene el trabajo medida")

	_, err = leida.Medida.peorCaso(t.TempDir(), skillConJuezSintetico)
	require.ErrorIs(t, err, os.ErrNotExist, "sin la carpeta de evals de la skill")

	etiquetados := filepath.Join(evals, skillConJuezSintetico, casosEnElJuez)
	require.NoError(t, os.WriteFile(etiquetados, []byte("clase: afirma_lo_no_leido\ncasos:\n  - etiqueta: dudoso\n"), 0o600))

	_, err = leida.Medida.peorCaso(evals, skillConJuezSintetico)
	require.ErrorContains(t, err, "los casos etiquetados del juez "+etiquetados+": ")
	require.ErrorContains(t, err, `tiene la etiqueta "dudoso"`)

	// Ni el de la entrada descripciones sin los casos de la evidencia del juez
	// de las descripciones en la raíz de las evals, o con unos que no tienen su
	// forma: su error nombra el fichero (research D9 de H31).
	_, err = leida.Medida.peorCaso(filepath.Join(t.TempDir(), directorioDeLasEvals), entradaDeLasDescripciones)
	require.ErrorIs(t, err, os.ErrNotExist, "sin los casos de la evidencia del juez de las descripciones")
	require.ErrorContains(t, err, rutaDeLaEvidenciaDeLasDescripciones+"/"+ficheroDeCasosDelJuez)

	deLaEvidencia := filepath.Join(carpetaDeLaEvidenciaEn(filepath.Dir(evals)), ficheroDeCasosDelJuez)
	require.NoError(t, os.WriteFile(deLaEvidencia, []byte("casos:\n  - etiqueta: dudoso\n"), 0o600))

	_, err = leida.Medida.peorCaso(evals, entradaDeLasDescripciones)
	require.ErrorContains(t, err, "los casos etiquetados del juez de las descripciones ("+
		rutaDeLaEvidenciaDeLasDescripciones+"/"+ficheroDeCasosDelJuez+"): ")
	require.ErrorContains(t, err, `tiene la etiqueta "dudoso"`)
}

// probarElPlazoDeLaDecision fija de dónde sale el plazo del paso de la decisión
// del trabajo evals que el peor caso suma (contracts/job-de-evals.md §2 de H30;
// research D17 de H30): del -timeout del go test de su run, escrito con un
// espacio o con un igual; sin el paso, ninguno; y, sin -timeout, sin su valor o
// con un valor que no es una duración mayor que 0, que en go test es no tener
// plazo, un error que nombra el paso.
func probarElPlazoDeLaDecision(t *testing.T) {
	t.Parallel()

	const (
		conSuPlazo = "-count=1 -timeout 15m -v -run '^TestDecisionDelTrabajo$'"
		delPaso    = `el paso "Decidir qu` + "\xc3\xa9" + ` mide y qu` + "\xc3\xa9" + ` reutiliza" `
		sinPlazo   = delPaso + "no fija su plazo con -timeout"
	)

	casos := []struct {
		nombre  string
		cambios []cambioDeLaDefinicion
		plazo   time.Duration
		error   string
	}{
		{nombre: "el-del-contrato", plazo: plazoDeLaDecisionDelContrato},
		{
			nombre:  "con-un-igual",
			cambios: []cambioDeLaDefinicion{{conSuPlazo, "-count=1 -timeout=20m -v -run '^TestDecisionDelTrabajo$'"}},
			plazo:   20 * time.Minute,
		},
		{nombre: "sin-el-paso", cambios: []cambioDeLaDefinicion{{pasoDeLaDecisionDelContrato, ""}}},
		{
			nombre:  "sin-timeout",
			cambios: []cambioDeLaDefinicion{{conSuPlazo, "-count=1 -v -run '^TestDecisionDelTrabajo$'"}},
			error:   sinPlazo,
		},
		{
			nombre: "con-timeout-al-final-sin-valor",
			cambios: []cambioDeLaDefinicion{
				{conSuPlazo, "-count=1 -v -run '^TestDecisionDelTrabajo$'"},
				{`-salida "$RUNNER_TEMP/decision-de-evals.json"`, `-salida "$RUNNER_TEMP/decision-de-evals.json" -timeout`},
			},
			error: sinPlazo,
		},
		{
			nombre:  "sin-limite",
			cambios: []cambioDeLaDefinicion{{conSuPlazo, "-count=1 -timeout 0 -v -run '^TestDecisionDelTrabajo$'"}},
			error:   delPaso + `tiene el plazo -timeout "0", y tiene que ser una duraci` + "\xc3\xb3" + "n mayor que 0",
		},
		{
			nombre:  "con-un-plazo-que-no-es-una-duracion",
			cambios: []cambioDeLaDefinicion{{conSuPlazo, "-count=1 -timeout quince -v -run '^TestDecisionDelTrabajo$'"}},
			error:   delPaso + `tiene el plazo -timeout "quince", y tiene que ser una duraci` + "\xc3\xb3" + "n mayor que 0",
		},
		{
			nombre:  "con-el-plazo-de-lo-que-lee",
			cambios: []cambioDeLaDefinicion{{conSuPlazo, "-count=1 -timeout 10m -v -run '^TestDecisionDelTrabajo$'"}},
			error: delPaso + "tiene el plazo -timeout 10m, y tiene que ser mayor que los 10 minutos de lo que la decisi" +
				"\xc3\xb3" + "n lee",
		},
	}

	for _, caso := range casos {
		t.Run(caso.nombre, func(t *testing.T) {
			t.Parallel()

			plazo, err := leerLaSintetica(t, caso.cambios).plazoDeLaDecision()
			if caso.error != "" {
				require.EqualError(t, err, caso.error)

				return
			}

			require.NoError(t, err)
			assert.Equal(t, caso.plazo, plazo)
		})
	}
}

// Los peores casos de los trabajos de las tres skills con las evals del
// repositorio y los ajustes de la definición del job, con sus términos: los de
// una tanda de sesiones, que llevan los tres, y los de una tanda de votos, que
// llevan los de las dos skills con juez. Y los de la medida del juez de
// jurisprudencia, con sus casos etiquetados. Los fija
// probarElPeorCasoDelTrabajo. Desde H30, los de los trabajos de las skills
// llevan, tras lo de fuera de las sesiones, los 900 s del plazo de la decisión
// (contracts/job-de-evals.md §2 de H30); los de la medida, no. Desde H31, los
// de los trabajos de las skills llevan detrás los seis votos de 40 s de la
// descripción de su skill y, el de legal-core, que no tiene juez de las
// respuestas, los 60 s de instalar el Claude Code del juez delante (research
// D10 de H31; SC-015).
const (
	terminosDeUnaTanda      = " \xc3\x97 (22 s + 240 s + 10 s)"
	terminosDeUnVoto        = " \xc3\x97 (35 s + 5 s)"
	terminosDeLaDescripcion = " + 6" + terminosDeUnVoto
	fueraYLaDecision        = " = 485 s + 900 s + "
	segundosDeLaDecision    = 900 * time.Second

	// El de normativa, el de research S1 de H20: sus veintitrés evals a
	// cinco a la vez y las 60 respuestas que su juez juzga en cada modo, con la
	// decisión, 19 877 s, el de contracts/job-de-evals.md §2 de H30, y, con su
	// descripción, 20 117 s (research D10 de H31).
	peorCasoDeNormativa = "20117 s" + fueraYLaDecision + "(\xe2\x8c\x88109 / 5\xe2\x8c\x89 + \xe2\x8c\x88108 / 5\xe2\x8c\x89 + " +
		"\xe2\x8c\x886 / 5\xe2\x8c\x89)" + terminosDeUnaTanda + " + 60 s + (\xe2\x8c\x8860 \xc3\x97 6 / 5\xe2\x8c\x89 + " +
		"\xe2\x8c\x8860 \xc3\x97 6 / 5\xe2\x8c\x89 + \xe2\x8c\x883 \xc3\x97 6 / 5\xe2\x8c\x89)" + terminosDeUnVoto +
		terminosDeLaDescripcion

	// El de jurisprudencia, el de contracts/job-de-evals.md §3 de H25: sus diez
	// evals a cuatro a la vez y las 30 respuestas que su juez juzga en cada modo.
	peorCasoDeJurisprudencia = "13717 s" + fueraYLaDecision + "(\xe2\x8c\x8861 / 4\xe2\x8c\x89 + \xe2\x8c\x8860 / 4\xe2\x8c\x89)" +
		terminosDeUnaTanda + " + 60 s + (\xe2\x8c\x8830 \xc3\x97 6 / 4\xe2\x8c\x89 + \xe2\x8c\x8830 \xc3\x97 6 / 4\xe2\x8c\x89)" +
		terminosDeUnVoto + terminosDeLaDescripcion

	// peorCasoDeJurisprudenciaDeUnaEnUna es el de la definición que la deja con
	// la concurrencia de antes de H25, que el tope del trabajo no cubre.
	peorCasoDeJurisprudenciaDeUnaEnUna = "48997 s" + fueraYLaDecision + "(\xe2\x8c\x8861 / 1\xe2\x8c\x89 + \xe2\x8c\x8860 / 1\xe2\x8c\x89)" +
		terminosDeUnaTanda + " + 60 s + (\xe2\x8c\x8830 \xc3\x97 6 / 1\xe2\x8c\x89 + \xe2\x8c\x8830 \xc3\x97 6 / 1\xe2\x8c\x89)" +
		terminosDeUnVoto + terminosDeLaDescripcion

	// El de legal-core, el de FR-050 de H9: sus doce evals, con la de vigencia
	// del ADR 0042 y la 11 y la 12 de H26.1, a dos a la vez (research V10 y D13
	// de H26.1; FR-020).
	peorCasoDeLegalCore = "20725 s" + fueraYLaDecision + "(\xe2\x8c\x8867 / 2\xe2\x8c\x89 + \xe2\x8c\x8866 / 2\xe2\x8c\x89 + " +
		"\xe2\x8c\x886 / 2\xe2\x8c\x89)" + terminosDeUnaTanda + " + 60 s" + terminosDeLaDescripcion

	// El de la medida del juez de jurisprudencia, con sus 249 casos etiquetados
	// a cuatro a la vez, y el de la definición que los deja de uno en uno, que el
	// tope del trabajo medida no cubre.
	peorCasoDeLaMedidaDeJurisprudencia = "15505 s = 485 s + 60 s + (\xe2\x8c\x88249 \xc3\x97 6 / 4\xe2\x8c\x89)" +
		terminosDeUnVoto
	peorCasoDeLaMedidaDeJurisprudenciaDeUnaEnUna = "60305 s = 485 s + 60 s + (\xe2\x8c\x88249 \xc3\x97 6 / 1\xe2\x8c\x89)" +
		terminosDeUnVoto

	// El de la medida del juez de las descripciones, con los 99 casos de la
	// evidencia a cinco a la vez, el de research D9 de H31, y el de la
	// definición que los deja de uno en uno, que el tope del trabajo medida no
	// cubre.
	peorCasoDeLaMedidaDeLasDescripciones = "5305 s = 485 s + 60 s + (\xe2\x8c\x8899 \xc3\x97 6 / 5\xe2\x8c\x89)" +
		terminosDeUnVoto
	peorCasoDeLaMedidaDeLasDescripcionesDeUnoEnUno = "24305 s = 485 s + 60 s + (\xe2\x8c\x8899 \xc3\x97 6 / 1\xe2\x8c\x89)" +
		terminosDeUnVoto
)

// probarElPeorCasoDelTrabajo fija el peor caso del trabajo de una skill desde
// H21 (contracts/evals-en-dos-modos.md §6 de H21; research.md D20 de H21;
// FR-083 de H21): una tanda por cada grupo de sesiones —las del modo orden con
// la prueba de red, las del modo herramienta y las de las evals sin binario ni
// servidor— y no por la suma de todas. Con las evals del repositorio y la
// definición del job, las sesiones son 14 357 s en normativa y, a dos a la
// vez, en legal-core, que cuenta la prueba de red aunque su trabajo no la lleve,
// 6469 s antes de H9 y, desde H9, con sus nueve evals, 14 629 s (FR-050 de H9).
//
// Desde H24 (contracts/job-de-evals.md §4 de H24; research D6 y S3 de H24;
// FR-092 de H24; SC-012 de H24), el de una skill con juez suma los 60 s de
// instalar su Claude Code y, por grupo, una tanda de 40 s —el tope de un voto y
// su margen— por cada ⌈respuestas × 6 / concurrencia⌉, con las respuestas del
// modelo que decide en las evals que activan la skill: 54, 54 y 3 en
// normativa, que con ello llega a 21 097 s. El de legal-core, que no tiene
// juez, no cambia. Y el de la medida del juez de normativa, con sus 259
// casos etiquetados, es de 16 105 s. El tope de cada trabajo cubre el suyo.
//
// Desde H23 (contracts/evals-jurisprudencia.md §6 de H23; research.md M5 y V29
// de H23; FR-063 de H23), el de jurisprudencia: sus evals son de modo, así que
// sus sesiones van en dos grupos y no en tres, con la prueba de red contada en
// el modo orden. Desde H25 (contracts/job-de-evals.md §3 de H25; research M6 de
// H25; FR-070 a FR-072), con sus diez evals —60 sesiones por modo— a cuatro a la
// vez y con su juez, que juzga las 30 respuestas del modelo que decide en cada
// modo: 31 tandas de 272 s y, con los 60 s de instalar el Claude Code del juez,
// 90 tandas de votos de 40 s, 12 577 s, que el tope del trabajo cubre sin
// cambiar. Y el de la medida de su juez, con sus 249 casos etiquetados a cuatro
// a la vez, es de 15 505 s, que cubre el tope del trabajo medida. Los de
// normativa y legal-core no cambian. Desde H9, el de legal-core es el de
// sus nueve evals, que el tope del trabajo cubre sin cambiar (FR-050, FR-058).
//
// Desde H20 (research D16 y S1 de H20; FR-050, FR-059 de H20), el de
// normativa es el de sus veintitrés evals, con las dos con fecha: 109, 108
// y 6 sesiones y 60, 60 y 3 respuestas del modelo que decide, 18 977 s, que el
// tope del trabajo cubre sin cambiar la concurrencia, el objetivo ni el tope.
//
// Desde H26.1 (research V10 y D13 de H26.1; FR-020; SC-009), el de legal-core
// es el de sus doce evals, con la 11 y la 12, de modo: 67, 66 y 6 sesiones,
// 19 525 s, el mayor de los tres, que el tope del trabajo cubre sin cambiar.
//
// Desde H30 (contracts/job-de-evals.md §2 de H30; research D17 de H30; FR-020),
// cada uno suma los 900 s del plazo del paso de la decisión, el -timeout de su
// go test: 19 877 s en normativa, 13 477 s en jurisprudencia y 20 425 s
// en legal-core, el mayor, que el tope del trabajo, 352 minutos, cubre sin
// cambiar. Los de la medida no cambian: su trabajo no decide nada.
//
// Desde H31 (research D1 y D10 de H31; FR-030; SC-015), todo trabajo evals
// puede juzgar la descripción de su skill, una, con los seis votos que como
// mucho pide, uno tras otro, de 40 s, 240 s; y, si la skill no tiene juez de
// las respuestas, que no la cuenta, los 60 s de instalar el Claude Code del
// juez: 20 117 s en normativa, 13 717 s en jurisprudencia y 20 725 s en
// legal-core, el mayor, que el tope del trabajo, 352 minutos, 21 120 s, cubre
// sin cambiar.
func probarElPeorCasoDelTrabajo(t *testing.T) {
	t.Parallel()

	leida, err := leerDefinicionDelJob(rutaDeLaDefinicionDelJob)
	require.NoError(t, err)

	casos := []struct {
		skill    string
		peor     peorCasoDelTrabajo
		duracion time.Duration
		texto    string
	}{
		{
			skill: "normativa",
			peor: peorCasoDelTrabajo{
				Decision:           segundosDeLaDecision,
				SesionesPorGrupo:   []int{109, 108, 6},
				Concurrencia:       5,
				Juez:               &peorCasoDelJuez{RespuestasPorGrupo: []int{60, 60, 3}, Concurrencia: 5},
				JuzgaLaDescripcion: true,
			},
			duracion: 20117 * time.Second,
			texto:    peorCasoDeNormativa,
		},
		{
			skill: "jurisprudencia",
			peor: peorCasoDelTrabajo{
				Decision:           segundosDeLaDecision,
				SesionesPorGrupo:   []int{61, 60},
				Concurrencia:       4,
				Juez:               &peorCasoDelJuez{RespuestasPorGrupo: []int{30, 30}, Concurrencia: 4},
				JuzgaLaDescripcion: true,
			},
			duracion: 13717 * time.Second,
			texto:    peorCasoDeJurisprudencia,
		},
		{
			skill: "legal-core",
			peor: peorCasoDelTrabajo{
				Decision: segundosDeLaDecision, SesionesPorGrupo: []int{67, 66, 6}, Concurrencia: 2,
				JuzgaLaDescripcion: true,
			},
			duracion: 20725 * time.Second,
			texto:    peorCasoDeLegalCore,
		},
	}

	for _, caso := range casos {
		peor, err := leida.peorCaso(directorioDeEvals, caso.skill)
		require.NoError(t, err, caso.skill)

		assert.Equal(t, caso.peor, peor, caso.skill)
		assert.Equal(t, caso.duracion, peor.Duracion(), caso.skill)
		assert.Equal(t, caso.texto, peor.String(), caso.skill)
		assert.GreaterOrEqual(t, time.Duration(leida.TopeEnMinutos)*time.Minute, peor.Duracion(),
			"el tope de la definici\xc3\xb3n cubre el peor caso de %s", caso.skill)
	}

	// Las 199 sesiones de normativa en un solo grupo serían 50 tandas, una
	// menos que las 51 de sus tres grupos: el peor caso no es el de la suma.
	enUnSoloGrupo := peorCasoDelTrabajo{SesionesPorGrupo: []int{199}, Concurrencia: 4}
	assert.Equal(t, 14085*time.Second, enUnSoloGrupo.Duracion())
	assert.Equal(t, "14085 s = 485 s + (\xe2\x8c\x88199 / 4\xe2\x8c\x89)"+terminosDeUnaTanda, enUnSoloGrupo.String())

	// Los términos del juez son los del tope de un voto y su margen, 35 s y 5 s,
	// y los de los seis votos que una respuesta puede pedir como mucho: tres,
	// cada uno con su repetición por nulo.
	assert.Equal(t, 40*time.Second, topeDelVoto+margenDelVoto)
	assert.Equal(t, 6, votosPorRespuestaComoMucho)

	// Los votos de un grupo tampoco llenan su última tanda con los del
	// siguiente: 3 y 3 respuestas, a 4 a la vez, son 5 + 5 tandas de votos, y las
	// 6 en un solo grupo serían 9.
	delJuez := peorCasoDelJuez{RespuestasPorGrupo: []int{3, 3}, Concurrencia: 4}
	assert.Equal(t, 460*time.Second, delJuez.Duracion())
	assert.Equal(t, 420*time.Second, peorCasoDelJuez{RespuestasPorGrupo: []int{6}, Concurrencia: 4}.Duracion())

	// La descripción de la skill (research D10 de H31): sus seis votos, uno tras
	// otro, de 40 s, que el trabajo de una skill con juez de las respuestas suma
	// tal cual, y el de una sin él, con los 60 s de instalar el Claude Code del
	// juez, que entonces no cuenta nadie más.
	assert.Equal(t, 240*time.Second, peorCasoDeLaDescripcion)

	conJuez := peorCasoDelTrabajo{SesionesPorGrupo: []int{3}, Concurrencia: 4, Juez: &delJuez}
	assert.Equal(t, 1217*time.Second, conJuez.Duracion())

	conJuez.JuzgaLaDescripcion = true
	assert.Equal(t, 1457*time.Second, conJuez.Duracion())
	assert.Equal(t, "1457 s = 485 s + (\xe2\x8c\x883 / 4\xe2\x8c\x89)"+terminosDeUnaTanda+" + 60 s + (\xe2\x8c\x883 \xc3\x97 6 / 4\xe2\x8c\x89 + "+
		"\xe2\x8c\x883 \xc3\x97 6 / 4\xe2\x8c\x89)"+terminosDeUnVoto+terminosDeLaDescripcion, conJuez.String())

	sinJuez := enUnSoloGrupo
	sinJuez.JuzgaLaDescripcion = true
	assert.Equal(t, 14385*time.Second, sinJuez.Duracion())
	assert.Equal(t, "14385 s = 485 s + (\xe2\x8c\x88199 / 4\xe2\x8c\x89)"+terminosDeUnaTanda+" + 60 s"+terminosDeLaDescripcion,
		sinJuez.String())

	// La medida del juez (contracts/job-de-evals.md §4 de H24): sin sesiones de
	// evals, lo de fuera de las sesiones, la instalación del Claude Code del
	// juez y los votos de sus casos etiquetados, que son los de la copia.
	require.NotNil(t, leida.Medida, "la definici\xc3\xb3n del job tiene el trabajo medida")

	medidas := []struct {
		skill    string
		peor     peorCasoDeLaMedida
		duracion time.Duration
		texto    string
	}{
		{
			skill:    "normativa",
			peor:     peorCasoDeLaMedida{Casos: 259, Concurrencia: 5},
			duracion: 12985 * time.Second,
			texto:    "12985 s = 485 s + 60 s + (\xe2\x8c\x88259 \xc3\x97 6 / 5\xe2\x8c\x89)" + terminosDeUnVoto,
		},
		{
			skill:    "jurisprudencia",
			peor:     peorCasoDeLaMedida{Casos: 249, Concurrencia: 4},
			duracion: 15505 * time.Second,
			texto:    peorCasoDeLaMedidaDeJurisprudencia,
		},
		{
			// La entrada descripciones, con los 99 casos de la evidencia del juez
			// de las descripciones (research D9 de H31; SC-015).
			skill:    entradaDeLasDescripciones,
			peor:     peorCasoDeLaMedida{Casos: 99, Concurrencia: concurrenciaDeLaMedidaDeLasDescripciones},
			duracion: 5305 * time.Second,
			texto:    peorCasoDeLaMedidaDeLasDescripciones,
		},
	}

	for _, caso := range medidas {
		deLaMedida, err := leida.Medida.peorCaso(directorioDeEvals, caso.skill)
		require.NoError(t, err, caso.skill)

		assert.Equal(t, caso.peor, deLaMedida, caso.skill)
		assert.Equal(t, caso.duracion, deLaMedida.Duracion(), caso.skill)
		assert.Equal(t, caso.texto, deLaMedida.String(), caso.skill)
		assert.GreaterOrEqual(t, time.Duration(leida.Medida.TopeEnMinutos)*time.Minute, deLaMedida.Duracion(),
			"el tope del trabajo medida cubre el peor caso de la medida de %s", caso.skill)
	}
}

// ejecucionPropia es la ejecución que decide en segundo-disparo: las de
// databaseId menor son anteriores y las de mayor, posteriores (research.md S2
// de H7.4).
const ejecucionPropia = 20

// Los nombres de la definición del job con los que gh da los trabajos y los
// pasos de una ejecución: el trabajo tanda, sin name, con su id; su paso que
// decide y sus dos marcas, la que solo corre en la ejecución que mide
// (contracts/tanda-del-job.md §1 de H7.4) y la que solo corre en la que mide la
// suite completa (contracts/job-de-evals.md §2 de H30); cambios; y el trabajo
// evals saltado por su if, con el nombre sin la skill (research.md O2 de H7.4).
const (
	nombreDeLaTanda       = "tanda"
	pasoQueDecide         = "Mirar si otra tanda mide este commit"
	pasoDeLaMarca         = "Esta ejecuci\xc3\xb3n mide el commit"
	pasoDeLaSuiteCompleta = "Esta ejecuci\xc3\xb3n mide la suite completa"
	trabajoDeCambios      = "cambios"
	evalsSaltadas         = "evals (${{ matrix.skill }})"
)

// Los valores de status y conclusion con los que gh da una ejecución, un
// trabajo o un paso.
const (
	ghTerminado = "completed"
	ghEnCurso   = "in_progress"
	ghEnEspera  = "pending"
	ghConExito  = "success"
	ghSaltado   = "skipped"
)

// Los trabajos sintéticos de una ejecución del flujo evals, como los da gh run
// view --json jobs (research.md O1 de H7.4): un trabajo saltado por su if sale
// terminado, con la conclusión skipped y sin pasos, y un paso saltado, con la
// conclusión skipped. El que mide deja cambios, su tanda con la marca —y, si
// mide la suite completa, también la segunda— y un trabajo evals por skill; el
// que no, las dos marcas saltadas y evals saltado entero.
var (
	cambiosTerminado = trabajoSintetico(trabajoDeCambios, ghTerminado, ghConExito)

	// tandaQueMideConReutilizacionSintetica es la tanda de una ejecución que
	// mide con reutilización, con la marca de la suite completa saltada;
	// tandaQueMideLaSuiteCompletaSintetica, la de una que mide la suite
	// completa, con las dos marcas; y tandaQueNoMideSintetica, la de una que no
	// mide, con las dos saltadas.
	tandaQueMideConReutilizacionSintetica = tandaSintetica(ghConExito, ghSaltado)
	tandaQueMideLaSuiteCompletaSintetica  = tandaSintetica(ghConExito, ghConExito)
	tandaQueNoMideSintetica               = tandaSintetica(ghSaltado, ghSaltado)
	evalsSaltadasSinteticas               = trabajoSintetico(evalsSaltadas, ghTerminado, ghSaltado)

	// trabajosSinDecidir son los de una ejecución que aún está en cambios: su
	// tanda no se ha creado.
	trabajosSinDecidir = trabajosSinteticos(trabajoSintetico(trabajoDeCambios, ghEnCurso, ""))

	// trabajosQueMidenYCorren son los de una ejecución que mide con
	// reutilización y cuyas sesiones corren; trabajosQueMidenYEsperan, los de
	// una que mide con reutilización y cuyos trabajos evals esperan por su
	// concurrency; trabajosQueMidieron, los de una que ya midió; y
	// trabajosQueMidenLaSuiteCompletaYCorren, los de una que mide la suite
	// completa y cuyas sesiones corren.
	trabajosQueMidenYCorren                = trabajosQueMiden(tandaQueMideConReutilizacionSintetica, ghEnCurso, "")
	trabajosQueMidenYEsperan               = trabajosQueMiden(tandaQueMideConReutilizacionSintetica, ghEnEspera, "")
	trabajosQueMidieron                    = trabajosQueMiden(tandaQueMideConReutilizacionSintetica, ghTerminado, ghConExito)
	trabajosQueMidenLaSuiteCompletaYCorren = trabajosQueMiden(tandaQueMideLaSuiteCompletaSintetica, ghEnCurso, "")

	// trabajosQueNoMiden son los de una ejecución cuya tanda decidió no medir.
	trabajosQueNoMiden = trabajosSinteticos(cambiosTerminado, tandaQueNoMideSintetica, evalsSaltadasSinteticas)
)

// tandaSintetica es el trabajo tanda terminado de una ejecución, con su paso
// que decide y sus dos marcas, cada una con la conclusion dada: primero la
// marca de que mide el commit y después la de que mide la suite completa.
func tandaSintetica(conclusionDeLaMarca, conclusionDeLaSuiteCompleta string) string {
	return trabajoSintetico(nombreDeLaTanda, ghTerminado, ghConExito,
		pasoSintetico(pasoQueDecide, ghTerminado, ghConExito),
		pasoSintetico(pasoDeLaMarca, ghTerminado, conclusionDeLaMarca),
		pasoSintetico(pasoDeLaSuiteCompleta, ghTerminado, conclusionDeLaSuiteCompleta))
}

// trabajosQueMiden son los trabajos de una ejecución cuya tanda, la dada, mide,
// con los trabajos evals de las dos skills en el estado y con la conclusión
// dados.
func trabajosQueMiden(tanda, estadoDeEvals, conclusionDeEvals string) string {
	return trabajosSinteticos(cambiosTerminado, tanda,
		trabajoSintetico("evals (normativa)", estadoDeEvals, conclusionDeEvals),
		trabajoSintetico("evals (legal-core)", estadoDeEvals, conclusionDeEvals))
}

// trabajosSinteticos es el JSON de gh run view --json jobs con los trabajos
// dados.
func trabajosSinteticos(trabajos ...string) string {
	return `{"jobs":[` + strings.Join(trabajos, ",") + "]}"
}

// trabajoSintetico es un trabajo del JSON de gh run view --json jobs, con sus
// pasos.
func trabajoSintetico(nombre, estado, conclusion string, pasos ...string) string {
	return fmt.Sprintf(`{"name":%q,"status":%q,"conclusion":%q,"steps":[%s]}`, nombre, estado, conclusion,
		strings.Join(pasos, ","))
}

// pasoSintetico es un paso de un trabajo del JSON de gh run view --json jobs.
func pasoSintetico(nombre, estado, conclusion string) string {
	return fmt.Sprintf(`{"name":%q,"status":%q,"conclusion":%q}`, nombre, estado, conclusion)
}

// ejecucionesSinteticas es el JSON de gh run list --json databaseId,status
// con las ejecuciones dadas.
func ejecucionesSinteticas(ejecuciones ...string) string {
	return "[" + strings.Join(ejecuciones, ",") + "]"
}

// ejecucionSintetica es una ejecución del JSON de gh run list --json
// databaseId,status.
func ejecucionSintetica(id int64, estado string) string {
	return fmt.Sprintf(`{"databaseId":%d,"status":%q}`, id, estado)
}

// rondaSintetica es lo que gh da en una consulta: la lista de ejecuciones del
// flujo sobre el commit y los trabajos de cada ejecución, por su databaseId.
type rondaSintetica struct {
	lista    string
	trabajos map[int64]string
}

// listar da la lista de ejecuciones de la ronda, como gh run list.
func (r rondaSintetica) listar(context.Context) ([]byte, error) {
	return []byte(r.lista), nil
}

// verLosTrabajos da los trabajos de la ejecución id en la ronda, como gh run
// view, y un error si la ronda no los tiene.
func (r rondaSintetica) verLosTrabajos(_ context.Context, id int64) ([]byte, error) {
	trabajos, estan := r.trabajos[id]
	if !estan {
		return nil, fmt.Errorf("la ronda no tiene los trabajos de la ejecuci\xc3\xb3n %d", id)
	}

	return []byte(trabajos), nil
}

// ghSintetico da en cada consulta lo que gh daría en su ronda, y desde la
// última, lo mismo que en ella; y cuenta las consultas.
type ghSintetico struct {
	rondas    []rondaSintetica
	consultas int
}

// consultar hace la consulta siguiente con consultarLasEjecuciones.
func (g *ghSintetico) consultar(ctx context.Context) ([]ejecucionDelCommit, error) {
	ronda := g.rondas[min(g.consultas, len(g.rondas)-1)]
	g.consultas++

	return consultarLasEjecuciones(ctx, ejecucionPropia, ronda.listar, ronda.verLosTrabajos)
}

// relojSintetico es un reloj que no duerme: cada espera lo adelanta lo que
// dura y queda anotada.
type relojSintetico struct {
	instante time.Time
	esperas  []time.Duration
}

// ahora es el instante del reloj.
func (r *relojSintetico) ahora() time.Time {
	return r.instante
}

// esperar adelanta el reloj la duración y la anota.
func (r *relojSintetico) esperar(_ context.Context, duracion time.Duration) error {
	r.instante = r.instante.Add(duracion)
	r.esperas = append(r.esperas, duracion)

	return nil
}

// segundoDisparo es un caso de la tabla de contracts/tanda-del-job.md §4 de
// H7.4, con el modo de la ejecución propia (data-model §10 de H30): lo que gh
// da en cada consulta, la decisión que se espera, las consultas que se hacen y
// si se mide por agotar la espera.
type segundoDisparo struct {
	nombre     string
	modo       modoDelOrigen
	rondas     []rondaSintetica
	mide       bool
	pendientes []int64
	consultas  int
	agotada    bool
}

// segundosDisparos son los casos de probarElSegundoDisparo, en el orden de la
// tabla de contracts/tanda-del-job.md §4 de H7.4: la ejecución propia y las
// anteriores que miden, todas con reutilización, el mismo modo; los de modos
// distintos son los de TestTandaConLaSuiteCompleta.
func segundosDisparos() []segundoDisparo {
	propiaEnCurso := ejecucionSintetica(ejecucionPropia, ghEnCurso)
	conLaAnteriorEnCurso := ejecucionesSinteticas(propiaEnCurso, ejecucionSintetica(10, ghEnCurso))

	return []segundoDisparo{
		{
			nombre:    "anterior-que-mide-corre",
			modo:      origenConReutilizacion,
			rondas:    []rondaSintetica{{conLaAnteriorEnCurso, map[int64]string{10: trabajosQueMidenYCorren}}},
			consultas: 1,
		},
		{
			nombre:    "anterior-que-mide-espera",
			modo:      origenConReutilizacion,
			rondas:    []rondaSintetica{{conLaAnteriorEnCurso, map[int64]string{10: trabajosQueMidenYEsperan}}},
			consultas: 1,
		},
		{
			nombre: "anterior-sin-decidir-que-mide",
			modo:   origenConReutilizacion,
			rondas: []rondaSintetica{
				{conLaAnteriorEnCurso, map[int64]string{10: trabajosSinDecidir}},
				{conLaAnteriorEnCurso, map[int64]string{10: trabajosQueMidenYCorren}},
			},
			consultas: 2,
		},
		{
			nombre: "anterior-sin-decidir-que-no-mide",
			modo:   origenConReutilizacion,
			rondas: []rondaSintetica{
				{conLaAnteriorEnCurso, map[int64]string{10: trabajosSinDecidir}},
				{conLaAnteriorEnCurso, map[int64]string{10: trabajosQueNoMiden}},
			},
			mide:      true,
			consultas: 2,
		},
		{
			nombre:    "anterior-que-no-mide",
			modo:      origenConReutilizacion,
			rondas:    []rondaSintetica{{conLaAnteriorEnCurso, map[int64]string{10: trabajosQueNoMiden}}},
			mide:      true,
			consultas: 1,
		},
		{
			// FR-071 de H7.4: la etiqueta sobre un commit cuya tanda terminó
			// vuelve a medir.
			nombre: "anterior-terminada",
			modo:   origenConReutilizacion,
			rondas: []rondaSintetica{{
				ejecucionesSinteticas(propiaEnCurso, ejecucionSintetica(10, ghTerminado)),
				map[int64]string{10: trabajosQueMidieron},
			}},
			mide:      true,
			consultas: 1,
		},
		{
			nombre: "posterior-que-mide",
			modo:   origenConReutilizacion,
			rondas: []rondaSintetica{{
				ejecucionesSinteticas(ejecucionSintetica(30, ghEnCurso), propiaEnCurso),
				map[int64]string{30: trabajosQueMidenYCorren},
			}},
			mide:      true,
			consultas: 1,
		},
		{
			nombre:    "sola",
			modo:      origenConReutilizacion,
			rondas:    []rondaSintetica{{ejecucionesSinteticas(propiaEnCurso), nil}},
			mide:      true,
			consultas: 1,
		},
		{
			// Una consulta cada 10 s desde la primera, en el segundo 0, hasta
			// la del minuto 10, que ya no espera: 61.
			nombre:     "espera-agotada",
			modo:       origenConReutilizacion,
			rondas:     []rondaSintetica{{conLaAnteriorEnCurso, map[int64]string{10: trabajosSinDecidir}}},
			mide:       true,
			pendientes: []int64{10},
			consultas:  61,
			agotada:    true,
		},
	}
}

// probarElSegundoDisparo fija la decisión de la tanda con la tabla de
// contracts/tanda-del-job.md §4 de H7.4 (FR-070, FR-071 y FR-100 de H7.4;
// SC-010 de H7.4), con el modo (data-model §10 de H30): la ejecución propia
// consulta un gh sintético y espera con un reloj que no duerme. No mide si una
// ejecución anterior sin terminar mide con su modo, aunque sus trabajos evals
// esperen; espera, 10 s entre consulta y consulta, a que decida la anterior que
// aún no lo ha hecho; y mide si ninguna anterior sin terminar mide, si la que
// midió ya terminó, si la que mide es posterior o si pasan 10 min sin que la
// anterior decida. Un error de la consulta es un error que la nombra.
func probarElSegundoDisparo(t *testing.T) {
	t.Parallel()

	for _, caso := range segundosDisparos() {
		t.Run(caso.nombre, func(t *testing.T) {
			t.Parallel()

			probarUnDisparo(t, caso)
		})
	}

	t.Run("consulta-que-falla", func(t *testing.T) {
		t.Parallel()

		errDeGh := errors.New("gh run list termin\xc3\xb3 con 1")
		fallida := func(context.Context) ([]ejecucionDelCommit, error) { return nil, errDeGh }

		var reloj relojSintetico

		_, err := esperarLaDecision(t.Context(), ejecucionPropia, origenConReutilizacion, fallida, reloj.ahora,
			reloj.esperar)
		require.ErrorIs(t, err, errDeGh)
		require.ErrorContains(t, err, "la consulta 1 de las ejecuciones del commit")
	})
}

// probarUnDisparo decide la tanda de la ejecución propia con el modo y las
// rondas de gh del caso, con un reloj que no duerme, y comprueba la decisión
// —si mide, con su modo, sus pendientes y si agotó la espera—, las consultas y
// los 10 s de espera entre una y la siguiente.
func probarUnDisparo(t *testing.T, caso segundoDisparo) {
	t.Helper()

	gh := ghSintetico{rondas: caso.rondas}

	var reloj relojSintetico

	decision, err := esperarLaDecision(t.Context(), ejecucionPropia, caso.modo, gh.consultar, reloj.ahora,
		reloj.esperar)
	require.NoError(t, err)

	assert.Equal(t, decisionTrasLaEspera{
		decisionDeLaTanda: decisionDeLaTanda{mide: caso.mide, modo: caso.modo, pendientes: caso.pendientes},
		agotada:           caso.agotada,
	}, decision)
	assert.Equal(t, caso.consultas, gh.consultas, "consultas")

	var esperas []time.Duration
	for range caso.consultas - 1 {
		esperas = append(esperas, 10*time.Second)
	}

	assert.Equal(t, esperas, reloj.esperas, "esperas")
}

// probarElEstadoDeLaTanda fija el estado de la tanda de una ejecución leído
// del JSON de sus trabajos (contracts/tanda-del-job.md §2 y §4 de H7.4;
// research.md S4 de H7.4), con su modo (data-model §10 de H30): mide con su
// trabajo tanda terminado y la marca en success, la suite completa si también
// la marca de la suite completa está en success y con reutilización si no; no
// mide con la marca saltada, aunque la de la suite completa esté en success, o
// con la tanda saltada; y aún no ha decidido con la tanda en curso o sin ella.
// Los pasos con los que gh da la tanda en estos casos son, por su nombre y en
// su orden, los tres últimos de la tanda de la definición del job: el que
// decide y sus dos marcas.
func probarElEstadoDeLaTanda(t *testing.T) {
	t.Parallel()

	delRepositorio, err := leerDefinicionDelJob(rutaDeLaDefinicionDelJob)
	require.NoError(t, err)
	require.NotNil(t, delRepositorio.Tanda, "premisa: la definici\xc3\xb3n del job tiene la tanda")

	nombres := make([]string, 0, len(delRepositorio.Tanda.Pasos))
	for _, paso := range delRepositorio.Tanda.Pasos {
		nombres = append(nombres, paso.Nombre)
	}

	require.GreaterOrEqual(t, len(nombres), 3, "premisa: la tanda tiene el paso que decide y sus dos marcas")
	require.Equal(t, []string{pasoQueDecide, pasoDeLaMarca, pasoDeLaSuiteCompleta}, nombres[len(nombres)-3:],
		"premisa: los pasos de la tanda sint\xc3\xa9tica son los de la definici\xc3\xb3n del job")

	tandaEnCurso := trabajoSintetico(nombreDeLaTanda, ghEnCurso, "",
		pasoSintetico(pasoQueDecide, ghEnCurso, ""), pasoSintetico(pasoDeLaMarca, ghEnEspera, ""),
		pasoSintetico(pasoDeLaSuiteCompleta, ghEnEspera, ""))
	tandaSaltada := trabajoSintetico(nombreDeLaTanda, ghTerminado, ghSaltado)

	casos := []struct {
		nombre   string
		trabajos string
		estado   estadoDeLaTanda
	}{
		{nombre: "marca-en-success", trabajos: trabajosQueMidenYCorren, estado: tandaQueMideConReutilizacion},
		{
			nombre:   "las-dos-marcas-en-success",
			trabajos: trabajosQueMidenLaSuiteCompletaYCorren,
			estado:   tandaQueMideLaSuiteCompleta,
		},
		{
			nombre:   "solo-la-marca-de-la-suite-completa",
			trabajos: trabajosSinteticos(cambiosTerminado, tandaSintetica(ghSaltado, ghConExito), evalsSaltadasSinteticas),
			estado:   tandaQueNoMide,
		},
		{nombre: "marca-saltada", trabajos: trabajosQueNoMiden, estado: tandaQueNoMide},
		{nombre: "tanda-en-curso", trabajos: trabajosSinteticos(cambiosTerminado, tandaEnCurso), estado: tandaSinDecidir},
		{nombre: "sin-tanda", trabajos: trabajosSinDecidir, estado: tandaSinDecidir},
		{
			nombre:   "tanda-saltada",
			trabajos: trabajosSinteticos(cambiosTerminado, tandaSaltada, evalsSaltadasSinteticas),
			estado:   tandaQueNoMide,
		},
	}

	for _, caso := range casos {
		t.Run(caso.nombre, func(t *testing.T) {
			t.Parallel()

			estado, err := leerLaTandaDeLaEjecucion([]byte(caso.trabajos))
			require.NoError(t, err)
			assert.Equal(t, caso.estado, estado)
		})
	}
}
