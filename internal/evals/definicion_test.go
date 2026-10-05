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
	// (contracts/job-de-evals.md §3 de H24).
	condicionDeLaTanda = sinCancelar + " && needs.cambios.result != 'failure' && (\n" +
		"  " + despachoSinLaMedida + " ||\n" +
		"  github.event.label.name == 'evals' ||\n" +
		"  github.event.label.name == 'evals-prueba-de-red' ||\n" +
		"  needs.cambios.outputs.coincide == 'si'\n" +
		")"

	// permisoDeLasEjecuciones es el de actions: con él lee gh las ejecuciones
	// del flujo y sus trabajos.
	permisoDeLasEjecuciones = "read"

	// salidaDeLaTanda es su output medir, el del paso que decide.
	salidaDeLaTanda = "${{ steps.decidir.outputs.medir }}"

	// ordenDelPasoQueDecide es el run del paso que decide, el de id decidir:
	// TestTandaDelCommit con el commit evaluado, la ejecución y el
	// GITHUB_OUTPUT del paso. El -timeout supera los 10 min que la decisión
	// espera como mucho, más sus consultas, y cabe con la preparación en los 15
	// del trabajo: con los 10 de go test por omisión, el test acabaría en pánico
	// antes de medir tras la espera, y el trabajo, en rojo.
	ordenDelPasoQueDecide = `go test -tags evals -count=1 -timeout 12m -v -run '^TestTandaDelCommit$' ./internal/evals/ ` +
		`-args -commit "$COMMIT_EVALUADO" -ejecucion "$EJECUCION" -salida "$GITHUB_OUTPUT"`

	// condicionDeLaMarca es el if de la marca: solo corre si el paso que
	// decide dice que la ejecución mide.
	condicionDeLaMarca = "steps.decidir.outputs.medir == 'si'"

	// condicionDelTrabajo es el if del trabajo evals: sin medir=si de la
	// tanda, se salta entero. La función de estado quita el success()
	// implícito, que podría saltarlo por cambios, saltado en la etiqueta y en
	// el despacho (research.md S9 de H7.4).
	condicionDelTrabajo = "${{ " + sinCancelar + " && needs.tanda.outputs.medir == 'si' }}"
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

// skillsDelTrabajo son las skills de su matriz, en su orden.
var skillsDelTrabajo = []string{"boe-legislacion", "legal-core"}

// ajustesDelTrabajo son la concurrencia y el objetivo de duración que su
// include da a cada skill (FR-030 y FR-051 de H7.3; research.md D12 de H7.3).
var ajustesDelTrabajo = map[string]AjustesDeSkill{
	"boe-legislacion": {Concurrencia: 4, ObjetivoDeDuracion: 900},
	"legal-core":      {Concurrencia: 1, ObjetivoDeDuracion: 0},
}

// entornoDelTrabajo son las variables con las que su env pasa esos ajustes a
// scripts/evals.sh.
var entornoDelTrabajo = map[string]string{
	"CONCURRENCIA_DE_EVALS":         "${{ matrix.concurrencia }}",
	"OBJETIVO_DE_DURACION_DE_EVALS": "${{ matrix.objetivo_de_duracion }}",
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
// también sus votos.
func TestDefinicionDelJob(t *testing.T) {
	t.Parallel()

	t.Run("del-repositorio", probarLaDefinicionDelRepositorio)
	t.Run("sinteticas", probarLasDefinicionesSinteticas)
	t.Run("errores", probarLosErroresDeLaDefinicion)
	t.Run("peor-caso", probarElPeorCasoDelTrabajo)
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
			"contracts/ejecucion-del-job.md \xc2\xa77 de H7.3 y contracts/job-de-evals.md \xc2\xa75 de H24:\n%s",
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
// en el orden de contracts/job-de-evals.md §5 de H24. Sin ninguna línea, la
// definición es la de los contratos.
func comprobarLaDefinicion(leida DefinicionDelJob, evals string) []string {
	var fallos []string

	fallos = append(fallos, fallosDeLaTanda(leida.Tanda)...)
	fallos = append(fallos, fallosDeLaDependencia(leida)...)
	fallos = append(fallos, fallosDeLaConcurrencia(leida)...)
	fallos = append(fallos, fallosDeLaMatriz(leida)...)
	fallos = append(fallos, fallosDelEntorno(leida)...)
	fallos = append(fallos, fallosDelJuez(leida)...)
	fallos = append(fallos, fallosDelTope(leida, evals)...)
	fallos = append(fallos, fallosDeLaMedida(leida)...)
	fallos = append(fallos, fallosDeLaEntrada(leida)...)
	fallos = append(fallos, fallosDelTopeDeLaMedida(leida.Medida, evals)...)

	return fallos
}

// fallosDeLaTanda comprueba el trabajo tanda de contracts/tanda-del-job.md §4
// de H7.4: que está; que un name no cambia el nombre con el que gh lo da, su
// id, que es el que busca la decisión; su if; que no tiene concurrency; su
// permiso de actions; su salida medir; y sus pasos. Sin el trabajo, solo esa
// línea.
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

	if salida := tanda.Salidas["medir"]; salida != salidaDeLaTanda {
		fallos = append(fallos, fallo("jobs.tanda.outputs.medir", presentarTexto(salida), strconv.Quote(salidaDeLaTanda)))
	}

	return append(fallos, fallosDeLosPasosDeLaTanda(tanda.Pasos)...)
}

// fallosDeLosPasosDeLaTanda comprueba los pasos del trabajo tanda de
// contracts/tanda-del-job.md §4 de H7.4: el run del paso que decide, el de id
// decidir, y el name y el if del último, la marca que leen las ejecuciones
// posteriores.
func fallosDeLosPasosDeLaTanda(pasos []PasoDelTrabajo) []string {
	var fallos []string

	var decide, ultimo PasoDelTrabajo

	if indice := slices.IndexFunc(pasos, func(paso PasoDelTrabajo) bool { return paso.ID == "decidir" }); indice >= 0 {
		decide = pasos[indice]
	}

	if len(pasos) > 0 {
		ultimo = pasos[len(pasos)-1]
	}

	if decide.Orden != ordenDelPasoQueDecide {
		fallos = append(fallos, fallo("jobs.tanda.steps, id decidir, run", presentarTexto(decide.Orden),
			strconv.Quote(ordenDelPasoQueDecide)))
	}

	if ultimo.Nombre != marcaDeLaTanda {
		fallos = append(fallos, fallo("jobs.tanda.steps, el \xc3\xbaltimo, name", presentarTexto(ultimo.Nombre),
			strconv.Quote(marcaDeLaTanda)))
	}

	if ultimo.Condicion != condicionDeLaMarca {
		fallos = append(fallos, fallo("jobs.tanda.steps, el \xc3\xbaltimo, if", presentarTexto(ultimo.Condicion),
			strconv.Quote(condicionDeLaMarca)))
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

	skills := slices.Sorted(maps.Keys(leida.PorSkill))
	for skill := range ajustesDelTrabajo {
		if !slices.Contains(skills, skill) {
			skills = append(skills, skill)
		}
	}

	slices.Sort(skills)

	for _, skill := range skills {
		encontrados, estan := leida.PorSkill[skill]
		esperados, seEsperan := ajustesDelTrabajo[skill]

		if estan == seEsperan && encontrados == esperados {
			continue
		}

		esperado := queNoEste
		if seEsperan {
			esperado = presentarAjustes(esperados)
		}

		encontrado := noEsta
		if estan {
			encontrado = "vale " + presentarAjustes(encontrados)
		}

		fallos = append(fallos, fallo("jobs.evals.strategy.matrix.include, skill "+skill, encontrado, esperado))
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

// presentarAjustes presenta los ajustes de una skill como las claves de su
// entrada de include.
func presentarAjustes(ajustes AjustesDeSkill) string {
	return fmt.Sprintf("{concurrencia: %d, objetivo_de_duracion: %d}", ajustes.Concurrencia,
		ajustes.ObjetivoDeDuracion)
}

// trabajoDeLaTandaDelContrato es el trabajo tanda de definicionDelContrato, el
// de contracts/tanda-del-job.md §1 de H7.4 sin sus comentarios ni los pasos que
// no se comprueban.
const trabajoDeLaTandaDelContrato = `  tanda:
    needs: [cambios]
    if: >-
      ` + sinCancelar + ` && needs.cambios.result != 'failure' && (
        ` + despachoSinLaMedida + ` ||
        github.event.label.name == 'evals' ||
        github.event.label.name == 'evals-prueba-de-red' ||
        needs.cambios.outputs.coincide == 'si'
      )
    runs-on: ubuntu-24.04
    timeout-minutes: 15
    permissions:
      contents: read
      actions: read
    outputs:
      medir: ${{ steps.decidir.outputs.medir }}
    steps:
      - name: Mirar si otra tanda mide este commit
        id: decidir
        run: >-
          go test -tags evals -count=1 -timeout 12m -v -run '^TestTandaDelCommit$' ./internal/evals/
          -args -commit "$COMMIT_EVALUADO" -ejecucion "$EJECUCION" -salida "$GITHUB_OUTPUT"
      - name: ` + pasoDeLaMarca + `
        if: steps.decidir.outputs.medir == 'si'
        run: echo mide
`

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

// pasosDeEvalsDelContrato son los pasos del trabajo evals de
// definicionDelContrato: de los de contracts/job-de-evals.md §2 de H24, el
// único que se comprueba, el que instala los dos Claude Code.
const pasosDeEvalsDelContrato = `    steps:
      - name: Instalar strace y Claude Code
        run: |
          ` + instalacionDeLasSesiones + `
          claude --version
` + lineaDeLaInstalacionDelJuez + `          "$RUNNER_TEMP/claude-del-juez/node_modules/.bin/claude" --version
          echo "CLAUDE_DEL_JUEZ=$RUNNER_TEMP/claude-del-juez/node_modules/.bin/claude" >> "$GITHUB_ENV"
`

// trabajoDeLaMedidaDelContrato es el trabajo medida de definicionDelContrato, el
// de contracts/job-de-evals.md §3 de H24 sin sus pasos, que no se comprueban, y
// con el tope que cubre el peor caso de la medida con los casos de
// evalsSinteticas.
const trabajoDeLaMedidaDelContrato = `  medida:
    name: medida del juez (${{ matrix.skill }})
` + lineaDelIfDeLaMedida + `    strategy:
      fail-fast: false
      matrix:
        skill: [boe-legislacion]
        include:
          - skill: boe-legislacion
            concurrencia: 4
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
// sus sesiones es de 1573 s en boe-legislacion (⌈7 / 4⌉ + ⌈6 / 4⌉ = 4 tandas) y
// de 4021 s en legal-core (13 tandas). boe-legislacion tiene además juez, que
// juzga las 3 respuestas del modelo que decide en cada modo: 60 s de instalar su
// Claude Code y (⌈3 × 6 / 4⌉ + ⌈3 × 6 / 4⌉) × 40 s de sus votos, 2033 s en
// total. 120 minutos cubren los dos. Y el de la medida, con sus 7 casos
// etiquetados, es de 485 + 60 + ⌈7 × 6 / 4⌉ × 40 = 985 s, que cubren los 17
// minutos de su trabajo.
const definicionDelContrato = `name: evals
on:
  workflow_dispatch:
    inputs:
` + entradaDeLaMedidaDelFlujo + `  pull_request:
    types: [opened, reopened, labeled]
jobs:
` + trabajoDeLaTandaDelContrato + trabajoDeEvalsDelContrato + trabajoDeLaMedidaDelContrato

// trabajoDeEvalsDelContrato es el trabajo evals de definicionDelContrato.
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
        skill: [boe-legislacion, legal-core]
        include:
          - skill: boe-legislacion
            concurrencia: 4
            objetivo_de_duracion: 900
          - skill: legal-core
            concurrencia: 1
            objetivo_de_duracion: 0
    runs-on: ubuntu-24.04
    timeout-minutes: 120
    env:
      MODELO_DE_EVALS: claude-sonnet-5
` + entornoDelJuezDelContrato + `      MODELOS_INFORMATIVOS_DE_EVALS: claude-haiku-4-5-20251001
      REPETICIONES_DE_EVALS: 3
      UMBRAL_DE_EVALS: 2
      CONCURRENCIA_DE_EVALS: ${{ matrix.concurrencia }}
      OBJETIVO_DE_DURACION_DE_EVALS: ${{ matrix.objetivo_de_duracion }}
      VERSION_DE_CLAUDE_CODE: ` + versionDeLasSesionesDelContrato + `
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
	return cambioDeLaDefinicion{
		antes:   trabajoDeLaMedidaDelContrato,
		despues: strings.Replace(trabajoDeLaMedidaDelContrato, antes, despues, 1),
	}
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

			contenido := definicionDelContrato
			for _, cambio := range caso.cambios {
				require.Contains(t, contenido, cambio.antes, "premisa: la definici\xc3\xb3n tiene lo que el caso cambia")
				require.NotEqual(t, cambio.antes, cambio.despues, "premisa: el cambio del caso cambia algo")
				contenido = strings.Replace(contenido, cambio.antes, cambio.despues, 1)
			}

			ruta := filepath.Join(t.TempDir(), "evals.yml")
			require.NoError(t, os.WriteFile(ruta, []byte(contenido), 0o600))

			leida, err := leerDefinicionDelJob(ruta)
			require.NoError(t, err)

			assert.Equal(t, caso.fallos, comprobarLaDefinicion(leida, evals))
		})
	}
}

// definicionesSinteticas son los casos de probarLasDefinicionesSinteticas: uno
// por comprobación de contracts/ejecucion-del-job.md §7 de H7.3, los del tope
// que lo cubren y, detrás, los del juez y de su medida.
func definicionesSinteticas() []definicionSintetica {
	return slices.Concat(definicionesSinteticasDelJob(), definicionesSinteticasDelJuez(),
		definicionesSinteticasDeLaMedida())
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
			cambios: []cambioDeLaDefinicion{{pasosDeEvalsDelContrato, ""}},
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
// puede obtener.
func definicionesSinteticasDeLaMedida() []definicionSintetica {
	const (
		yLoEsperadoEs    = ", y lo esperado es "
		soloConEtiqueta  = "github.event.label.name == 'evals-medir-juez'"
		despachoDeAntes  = "github.event_name == 'workflow_dispatch'"
		conLaEtiqueta    = "  github.event.label.name == 'evals-medir-juez' ||\n"
		lineaDelDespacho = "  " + despachoSinLaMedida + " ||\n"
		evalsConEtiqueta = "${{ " + sinCancelar + " && (needs.tanda.outputs.medir == 'si' || " + soloConEtiqueta + ") }}"
		entradaDelFlujo  = "on.workflow_dispatch.inputs.medir_al_juez"
		topeDeLaMedida   = "jobs.medida.timeout-minutes: "
		concurrenciaDe4  = "            concurrencia: 4\n"
		peorCasoDeMedida = "985 s = 485 s + 60 s + (\xe2\x8c\x887 \xc3\x97 6 / 4\xe2\x8c\x89) \xc3\x97 (35 s + 5 s)"
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
			// 16 minutos son 960 s, y el peor caso de la medida, 985 s: los cubren
			// los 17 de la definición del contrato.
			nombre:  "tope-de-la-medida-por-debajo-de-su-peor-caso",
			cambios: []cambioDeLaDefinicion{{"timeout-minutes: 17", "timeout-minutes: 16"}},
			fallos: []string{topeDeLaMedida + "vale 16 (960 s)" + yLoEsperadoEs +
				"al menos el peor caso de boe-legislacion, " + peorCasoDeMedida},
		},
		{
			nombre:  "medida-sin-la-concurrencia-de-su-skill",
			cambios: []cambioDeLaDefinicion{enLaMedida(concurrenciaDe4, "")},
			fallos: []string{topeDeLaMedida + "el peor caso de boe-legislacion no se puede obtener: la concurrencia " +
				"de boe-legislacion es 0 y tiene que ser un entero mayor o igual que 1"},
		},
		{
			// legal-core no tiene juez: no hay medida suya que quepa en ningún tope.
			nombre: "medida-de-una-skill-sin-juez",
			cambios: []cambioDeLaDefinicion{enLaMedida(
				"skill: [boe-legislacion]\n        include:\n          - skill: boe-legislacion\n",
				"skill: [legal-core]\n        include:\n          - skill: legal-core\n")},
			fallos: []string{topeDeLaMedida + "el peor caso de legal-core no se puede obtener: legal-core no tiene " +
				"juez que medir"},
		},
	}
}

// definicionesSinteticasDelJob son las definiciones que se apartan de
// contracts/tanda-del-job.md §4 de H7.4 y de contracts/ejecucion-del-job.md §7
// de H7.3, y las del tope del trabajo evals que lo cubren.
func definicionesSinteticasDelJob() []definicionSintetica {
	const (
		grupoPorCommit   = "group: evals-${{ github.event.pull_request.head.sha || github.sha }}-${{ matrix.skill }}"
		cancelaEnCurso   = "      cancel-in-progress: false\n"
		matrizDelJob     = "skill: [boe-legislacion, legal-core]"
		includeDeLegal   = "          - skill: legal-core\n            concurrencia: 1\n            objetivo_de_duracion: 0\n"
		topeDelJob       = "timeout-minutes: 120"
		concurrenciaEnv  = "      CONCURRENCIA_DE_EVALS: ${{ matrix.concurrencia }}\n"
		objetivoEnv      = "OBJETIVO_DE_DURACION_DE_EVALS: ${{ matrix.objetivo_de_duracion }}"
		repeticionesEnv  = "REPETICIONES_DE_EVALS: 3"
		ajustesDeBoe     = "{concurrencia: 4, objetivo_de_duracion: 900}"
		ajustesDeLegal   = "{concurrencia: 1, objetivo_de_duracion: 0}"
		peorCasoDeLegal  = "4021 s = 485 s + (\xe2\x8c\x887 / 1\xe2\x8c\x89 + \xe2\x8c\x886 / 1\xe2\x8c\x89) \xc3\x97 (22 s + 240 s + 10 s)"
		esperadoDelGrupo = `"evals-${{ github.event.pull_request.head.sha || github.sha }}-${{ matrix.skill }}"`
		pruebaDeRed      = "  github.event.label.name == 'evals-prueba-de-red' ||\n"
		ejecucionDeGh    = ` -ejecucion "$EJECUCION"`
		topeDeLaTanda    = "    timeout-minutes: 15\n"
		elUltimo         = "jobs.tanda.steps, el \xc3\xbaltimo, "
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
			cambios: []cambioDeLaDefinicion{{"    outputs:\n      medir: ${{ steps.decidir.outputs.medir }}\n", ""}},
			fallos: []string{`jobs.tanda.outputs.medir: no est` + "\xc3\xa1" + `, y lo esperado es ` +
				`"${{ steps.decidir.outputs.medir }}"`},
		},
		{
			nombre:  "decidir-sin-la-ejecucion",
			cambios: []cambioDeLaDefinicion{{ejecucionDeGh, ""}},
			fallos: []string{"jobs.tanda.steps, id decidir, run: vale " +
				strconv.Quote(strings.Replace(ordenDelPasoQueDecide, ejecucionDeGh, "", 1)) + ", y lo esperado es " +
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
			fallos: []string{elUltimo + `name: vale "Mide el commit", y lo esperado es "Esta ejecuci` + "\xc3\xb3" +
				`n mide el commit"`},
		},
		{
			nombre:  "marca-sin-su-if",
			cambios: []cambioDeLaDefinicion{{"        if: steps.decidir.outputs.medir == 'si'\n", ""}},
			fallos: []string{elUltimo + `if: no est` + "\xc3\xa1" + `, y lo esperado es ` +
				`"steps.decidir.outputs.medir == 'si'"`},
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
			cambios: []cambioDeLaDefinicion{{matrizDelJob, "skill: [boe-legislacion]"}},
			fallos: []string{"jobs.evals.strategy.matrix.skill: vale [boe-legislacion], y lo esperado es " +
				"[boe-legislacion, legal-core]"},
		},
		{
			nombre:  "otra-concurrencia",
			cambios: []cambioDeLaDefinicion{{"concurrencia: 4", "concurrencia: 2"}},
			fallos: []string{"jobs.evals.strategy.matrix.include, skill boe-legislacion: vale " +
				"{concurrencia: 2, objetivo_de_duracion: 900}, y lo esperado es " + ajustesDeBoe},
		},
		{
			nombre:  "otro-objetivo",
			cambios: []cambioDeLaDefinicion{{"objetivo_de_duracion: 0", "objetivo_de_duracion: 900"}},
			fallos: []string{"jobs.evals.strategy.matrix.include, skill legal-core: vale " +
				"{concurrencia: 1, objetivo_de_duracion: 900}, y lo esperado es " + ajustesDeLegal},
		},
		{
			nombre:  "include-sin-legal-core",
			cambios: []cambioDeLaDefinicion{{includeDeLegal, ""}},
			fallos: []string{
				"jobs.evals.strategy.matrix.include, skill legal-core: no est\xc3\xa1, y lo esperado es " + ajustesDeLegal,
				"jobs.evals.timeout-minutes: el peor caso de legal-core no se puede obtener: la concurrencia de " +
					"legal-core es 0 y tiene que ser un entero mayor o igual que 1",
			},
		},
		{
			nombre:  "include-con-otra-skill",
			cambios: []cambioDeLaDefinicion{{includeDeLegal, includeDeLegal + "          - skill: cita\n            concurrencia: 1\n"}},
			fallos: []string{"jobs.evals.strategy.matrix.include, skill cita: vale " +
				"{concurrencia: 1, objetivo_de_duracion: 0}, y lo esperado es que no est\xc3\xa9"},
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
			// 67 minutos son 4020 s, uno menos que el peor caso de legal-core, y
			// cubren el de boe-legislacion.
			nombre:  "tope-por-debajo-del-peor-caso",
			cambios: []cambioDeLaDefinicion{{topeDelJob, "timeout-minutes: 67"}},
			fallos: []string{"jobs.evals.timeout-minutes: vale 67 (4020 s), y lo esperado es al menos el peor " +
				"caso de legal-core, " + peorCasoDeLegal},
		},
		{
			// 33 minutos son 1980 s, uno menos que el peor caso de boe-legislacion
			// con su juez: cubrirían el de sus sesiones solas, 1573 s, que es el de
			// una skill sin juez.
			nombre:  "tope-por-debajo-de-los-dos",
			cambios: []cambioDeLaDefinicion{{topeDelJob, "timeout-minutes: 33"}},
			fallos: []string{
				"jobs.evals.timeout-minutes: vale 33 (1980 s), y lo esperado es al menos el peor caso de " +
					"boe-legislacion, 2033 s = 485 s + (\xe2\x8c\x887 / 4\xe2\x8c\x89 + \xe2\x8c\x886 / 4\xe2\x8c\x89) " +
					"\xc3\x97 (22 s + 240 s + 10 s) + 60 s + (\xe2\x8c\x883 \xc3\x97 6 / 4\xe2\x8c\x89 + " +
					"\xe2\x8c\x883 \xc3\x97 6 / 4\xe2\x8c\x89) \xc3\x97 (35 s + 5 s)",
				"jobs.evals.timeout-minutes: vale 33 (1980 s), y lo esperado es al menos el peor caso de " +
					"legal-core, " + peorCasoDeLegal,
			},
		},
		{
			// 34 minutos, 2040 s, cubren el de boe-legislacion con su juez, y no el
			// de legal-core.
			nombre:  "tope-que-cubre-el-de-la-skill-con-juez",
			cambios: []cambioDeLaDefinicion{{topeDelJob, "timeout-minutes: 34"}},
			fallos: []string{"jobs.evals.timeout-minutes: vale 34 (2040 s), y lo esperado es al menos el peor " +
				"caso de legal-core, " + peorCasoDeLegal},
		},
		{
			nombre:  "tope-que-cubre-el-peor-caso",
			cambios: []cambioDeLaDefinicion{{topeDelJob, "timeout-minutes: 68"}},
		},
		{
			// Con una repetición, 3 sesiones por skill en el modo orden y 2 en el
			// modo herramienta: 485 + 5 × 272 = 1845 s en legal-core, que 31
			// minutos cubren.
			nombre: "tope-con-las-repeticiones-del-env",
			cambios: []cambioDeLaDefinicion{
				{topeDelJob, "timeout-minutes: 31"},
				{repeticionesEnv, "REPETICIONES_DE_EVALS: 1"},
			},
		},
		{
			// Sin modelos informativos, 4 sesiones por skill en el modo orden y 3
			// en el modo herramienta: 485 + 7 × 272 = 2389 s en legal-core, que 40
			// minutos cubren.
			nombre: "tope-con-los-modelos-del-env",
			cambios: []cambioDeLaDefinicion{
				{topeDelJob, "timeout-minutes: 40"},
				{"MODELOS_INFORMATIVOS_DE_EVALS: claude-haiku-4-5-20251001", `MODELOS_INFORMATIVOS_DE_EVALS: ""`},
			},
		},
	}
}

// skillConJuezSintetico es la skill de evalsSinteticas que tiene juez, y
// casosDelJuezSintetico, sus casos etiquetados.
const (
	skillConJuezSintetico = "boe-legislacion"
	casosDelJuezSintetico = 7
)

// evalsSinteticas crea un directorio temporal de evals con una carpeta por
// skill de la matriz, cada una con la eval sintética del art. 21 de la LPAC, y
// devuelve su ruta. La de boe-legislacion lleva además la carpeta del juez, con
// siete casos etiquetados; la de legal-core, no: es la skill sin juez.
func evalsSinteticas(t *testing.T) string {
	t.Helper()

	evals := t.TempDir()

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

	conJuez := filepath.Join(evals, skillConJuezSintetico)
	escribirLaCarpetaDelJuez(t, conJuez)
	crearEntradas(t, conJuez, []entradaDeConjunto{{nombre: casosEnElJuez, contenido: casos.String()}})

	leidos, err := leerCasosEtiquetados(filepath.Join(conJuez, casosEnElJuez))
	require.NoError(t, err)
	require.Len(t, leidos.Casos, casosDelJuezSintetico, "premisa: los casos etiquetados del juez sint\xc3\xa9tico")

	return evals
}

// probarLosErroresDeLaDefinicion fija lo que no se puede leer ni obtener: una
// definición que no existe, que no es YAML o cuyas repeticiones no son un
// entero no se lee, con un error que nombra la ruta; y el peor caso de una skill
// sin su carpeta de evals o con un plan que no se puede componer no se obtiene,
// con un error que dice por qué, ni el de la medida de una skill sin su carpeta
// de evals o con unos casos etiquetados que no se pueden leer.
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

	_, err = leida.peorCaso(t.TempDir(), "boe-legislacion")
	require.ErrorIs(t, err, os.ErrNotExist, "sin la carpeta de evals de la skill")

	evals := evalsSinteticas(t)

	leida.ModeloQueDecide = "Claude Sonnet"
	_, err = leida.peorCaso(evals, "boe-legislacion")
	require.ErrorContains(t, err, `el modelo que decide "Claude Sonnet" no tiene la forma de un id de modelo`)

	// El peor caso de la medida tampoco se obtiene sin la carpeta de evals de la
	// skill ni con unos casos etiquetados que no se pueden leer: su error nombra
	// el fichero.
	require.NotNil(t, leida.Medida, "premisa: la definici\xc3\xb3n del contrato tiene el trabajo medida")

	_, err = leida.Medida.peorCaso(t.TempDir(), skillConJuezSintetico)
	require.ErrorIs(t, err, os.ErrNotExist, "sin la carpeta de evals de la skill")

	etiquetados := filepath.Join(evals, skillConJuezSintetico, casosEnElJuez)
	require.NoError(t, os.WriteFile(etiquetados, []byte("clase: afirma_lo_no_leido\ncasos:\n  - etiqueta: dudoso\n"), 0o600))

	_, err = leida.Medida.peorCaso(evals, skillConJuezSintetico)
	require.ErrorContains(t, err, "los casos etiquetados del juez "+etiquetados+": ")
	require.ErrorContains(t, err, `tiene la etiqueta "dudoso"`)
}

// probarElPeorCasoDelTrabajo fija el peor caso del trabajo de una skill desde
// H21 (contracts/evals-en-dos-modos.md §6 de H21; research.md D20 de H21;
// FR-083 de H21): una tanda por cada grupo de sesiones —las del modo orden con
// la prueba de red, las del modo herramienta y las de las evals sin binario ni
// servidor— y no por la suma de todas. Con las evals del repositorio y la
// definición del job, las sesiones son 14 357 s en boe-legislacion y 12 181 s
// en legal-core, que cuenta la prueba de red aunque su trabajo no la lleve.
//
// Desde H24 (contracts/job-de-evals.md §4 de H24; research D6 y S3 de H24;
// FR-092 de H24; SC-012 de H24), el de una skill con juez suma los 60 s de
// instalar su Claude Code y, por grupo, una tanda de 40 s —el tope de un voto y
// su margen— por cada ⌈respuestas × 6 / concurrencia⌉, con las respuestas del
// modelo que decide en las evals que activan la skill: 54, 54 y 3 en
// boe-legislacion, que con ello llega a 21 097 s. El de legal-core, que no tiene
// juez, no cambia. Y el de la medida del juez de boe-legislacion, con sus 259
// casos etiquetados, es de 16 105 s. El tope de cada trabajo cubre el suyo.
func probarElPeorCasoDelTrabajo(t *testing.T) {
	t.Parallel()

	leida, err := leerDefinicionDelJob(rutaDeLaDefinicionDelJob)
	require.NoError(t, err)

	const (
		terminosDeUnaTanda = " \xc3\x97 (22 s + 240 s + 10 s)"
		terminosDeUnVoto   = " \xc3\x97 (35 s + 5 s)"
	)

	casos := []struct {
		skill    string
		peor     peorCasoDelTrabajo
		duracion time.Duration
		texto    string
	}{
		{
			skill: "boe-legislacion",
			peor: peorCasoDelTrabajo{
				SesionesPorGrupo: []int{97, 96, 6},
				Concurrencia:     4,
				Juez:             &peorCasoDelJuez{RespuestasPorGrupo: []int{54, 54, 3}, Concurrencia: 4},
			},
			duracion: 21097 * time.Second,
			texto: "21097 s = 485 s + (\xe2\x8c\x8897 / 4\xe2\x8c\x89 + \xe2\x8c\x8896 / 4\xe2\x8c\x89 + " +
				"\xe2\x8c\x886 / 4\xe2\x8c\x89)" + terminosDeUnaTanda + " + 60 s + (\xe2\x8c\x8854 \xc3\x97 6 / 4\xe2\x8c\x89 + " +
				"\xe2\x8c\x8854 \xc3\x97 6 / 4\xe2\x8c\x89 + \xe2\x8c\x883 \xc3\x97 6 / 4\xe2\x8c\x89)" + terminosDeUnVoto,
		},
		{
			skill:    "legal-core",
			peor:     peorCasoDelTrabajo{SesionesPorGrupo: []int{19, 18, 6}, Concurrencia: 1},
			duracion: 12181 * time.Second,
			texto: "12181 s = 485 s + (\xe2\x8c\x8819 / 1\xe2\x8c\x89 + \xe2\x8c\x8818 / 1\xe2\x8c\x89 + " +
				"\xe2\x8c\x886 / 1\xe2\x8c\x89)" + terminosDeUnaTanda,
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

	// Las 199 sesiones de boe-legislacion en un solo grupo serían 50 tandas, una
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

	// La medida del juez (contracts/job-de-evals.md §4 de H24): sin sesiones de
	// evals, lo de fuera de las sesiones, la instalación del Claude Code del
	// juez y los votos de sus casos etiquetados, que son los de la copia.
	require.NotNil(t, leida.Medida, "la definici\xc3\xb3n del job tiene el trabajo medida")

	deLaMedida, err := leida.Medida.peorCaso(directorioDeEvals, "boe-legislacion")
	require.NoError(t, err)

	assert.Equal(t, peorCasoDeLaMedida{Casos: 259, Concurrencia: 4}, deLaMedida)
	assert.Equal(t, 16105*time.Second, deLaMedida.Duracion())
	assert.Equal(t, "16105 s = 485 s + 60 s + (\xe2\x8c\x88259 \xc3\x97 6 / 4\xe2\x8c\x89)"+terminosDeUnVoto, deLaMedida.String())
	assert.GreaterOrEqual(t, time.Duration(leida.Medida.TopeEnMinutos)*time.Minute, deLaMedida.Duracion(),
		"el tope del trabajo medida cubre el peor caso de la medida")
}

// ejecucionPropia es la ejecución que decide en segundo-disparo: las de
// databaseId menor son anteriores y las de mayor, posteriores (research.md S2
// de H7.4).
const ejecucionPropia = 20

// Los nombres de la definición del job con los que gh da los trabajos y los
// pasos de una ejecución: el trabajo tanda, sin name, con su id; su paso que
// decide y su último paso, la marca, que solo corre en la ejecución que mide
// (contracts/tanda-del-job.md §1 de H7.4); cambios; y el trabajo evals saltado
// por su if, con el nombre sin la skill (research.md O2 de H7.4).
const (
	nombreDeLaTanda  = "tanda"
	pasoQueDecide    = "Mirar si otra tanda mide este commit"
	pasoDeLaMarca    = "Esta ejecuci\xc3\xb3n mide el commit"
	trabajoDeCambios = "cambios"
	evalsSaltadas    = "evals (${{ matrix.skill }})"
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
// conclusión skipped. El que mide deja cambios, su tanda con la marca y un
// trabajo evals por skill; el que no, la marca saltada y evals saltado entero.
var (
	cambiosTerminado      = trabajoSintetico(trabajoDeCambios, ghTerminado, ghConExito)
	tandaQueMideSintetica = trabajoSintetico(nombreDeLaTanda, ghTerminado, ghConExito,
		pasoSintetico(pasoQueDecide, ghTerminado, ghConExito), pasoSintetico(pasoDeLaMarca, ghTerminado, ghConExito))
	tandaQueNoMideSintetica = trabajoSintetico(nombreDeLaTanda, ghTerminado, ghConExito,
		pasoSintetico(pasoQueDecide, ghTerminado, ghConExito), pasoSintetico(pasoDeLaMarca, ghTerminado, ghSaltado))
	evalsSaltadasSinteticas = trabajoSintetico(evalsSaltadas, ghTerminado, ghSaltado)

	// trabajosSinDecidir son los de una ejecución que aún está en cambios: su
	// tanda no se ha creado.
	trabajosSinDecidir = trabajosSinteticos(trabajoSintetico(trabajoDeCambios, ghEnCurso, ""))

	// trabajosQueMidenYCorren son los de una ejecución que mide y cuyas
	// sesiones corren; trabajosQueMidenYEsperan, los de una que mide y cuyos
	// trabajos evals esperan por su concurrency; y trabajosQueMidieron, los de
	// una que ya midió.
	trabajosQueMidenYCorren  = trabajosQueMiden(ghEnCurso, "")
	trabajosQueMidenYEsperan = trabajosQueMiden(ghEnEspera, "")
	trabajosQueMidieron      = trabajosQueMiden(ghTerminado, ghConExito)

	// trabajosQueNoMiden son los de una ejecución cuya tanda decidió no medir.
	trabajosQueNoMiden = trabajosSinteticos(cambiosTerminado, tandaQueNoMideSintetica, evalsSaltadasSinteticas)
)

// trabajosQueMiden son los trabajos de una ejecución cuya tanda mide, con los
// trabajos evals de las dos skills en el estado y con la conclusión dados.
func trabajosQueMiden(estadoDeEvals, conclusionDeEvals string) string {
	return trabajosSinteticos(cambiosTerminado, tandaQueMideSintetica,
		trabajoSintetico("evals (boe-legislacion)", estadoDeEvals, conclusionDeEvals),
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
// H7.4: lo que gh da en cada consulta, la decisión que se espera, las consultas
// que se hacen y si se mide por agotar la espera.
type segundoDisparo struct {
	nombre     string
	rondas     []rondaSintetica
	mide       bool
	pendientes []int64
	consultas  int
	agotada    bool
}

// segundosDisparos son los casos de probarElSegundoDisparo, en el orden de la
// tabla de contracts/tanda-del-job.md §4 de H7.4.
func segundosDisparos() []segundoDisparo {
	propiaEnCurso := ejecucionSintetica(ejecucionPropia, ghEnCurso)
	conLaAnteriorEnCurso := ejecucionesSinteticas(propiaEnCurso, ejecucionSintetica(10, ghEnCurso))

	return []segundoDisparo{
		{
			nombre:    "anterior-que-mide-corre",
			rondas:    []rondaSintetica{{conLaAnteriorEnCurso, map[int64]string{10: trabajosQueMidenYCorren}}},
			consultas: 1,
		},
		{
			nombre:    "anterior-que-mide-espera",
			rondas:    []rondaSintetica{{conLaAnteriorEnCurso, map[int64]string{10: trabajosQueMidenYEsperan}}},
			consultas: 1,
		},
		{
			nombre: "anterior-sin-decidir-que-mide",
			rondas: []rondaSintetica{
				{conLaAnteriorEnCurso, map[int64]string{10: trabajosSinDecidir}},
				{conLaAnteriorEnCurso, map[int64]string{10: trabajosQueMidenYCorren}},
			},
			consultas: 2,
		},
		{
			nombre: "anterior-sin-decidir-que-no-mide",
			rondas: []rondaSintetica{
				{conLaAnteriorEnCurso, map[int64]string{10: trabajosSinDecidir}},
				{conLaAnteriorEnCurso, map[int64]string{10: trabajosQueNoMiden}},
			},
			mide:      true,
			consultas: 2,
		},
		{
			nombre:    "anterior-que-no-mide",
			rondas:    []rondaSintetica{{conLaAnteriorEnCurso, map[int64]string{10: trabajosQueNoMiden}}},
			mide:      true,
			consultas: 1,
		},
		{
			// FR-071 de H7.4: la etiqueta sobre un commit cuya tanda terminó
			// vuelve a medir.
			nombre: "anterior-terminada",
			rondas: []rondaSintetica{{
				ejecucionesSinteticas(propiaEnCurso, ejecucionSintetica(10, ghTerminado)),
				map[int64]string{10: trabajosQueMidieron},
			}},
			mide:      true,
			consultas: 1,
		},
		{
			nombre: "posterior-que-mide",
			rondas: []rondaSintetica{{
				ejecucionesSinteticas(ejecucionSintetica(30, ghEnCurso), propiaEnCurso),
				map[int64]string{30: trabajosQueMidenYCorren},
			}},
			mide:      true,
			consultas: 1,
		},
		{
			nombre:    "sola",
			rondas:    []rondaSintetica{{ejecucionesSinteticas(propiaEnCurso), nil}},
			mide:      true,
			consultas: 1,
		},
		{
			// Una consulta cada 10 s desde la primera, en el segundo 0, hasta
			// la del minuto 10, que ya no espera: 61.
			nombre:     "espera-agotada",
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
// SC-010 de H7.4): la ejecución propia consulta un gh sintético y espera con un
// reloj que no duerme. No mide si una ejecución anterior sin terminar mide,
// aunque sus trabajos evals esperen; espera, 10 s entre consulta y consulta, a
// que decida la anterior que aún no lo ha hecho; y mide si ninguna anterior sin
// terminar mide, si la que midió ya terminó, si la que mide es posterior o si
// pasan 10 min sin que la anterior decida. Un error de la consulta es un error
// que la nombra.
func probarElSegundoDisparo(t *testing.T) {
	t.Parallel()

	for _, caso := range segundosDisparos() {
		t.Run(caso.nombre, func(t *testing.T) {
			t.Parallel()

			gh := ghSintetico{rondas: caso.rondas}

			var reloj relojSintetico

			decision, err := esperarLaDecision(t.Context(), ejecucionPropia, gh.consultar, reloj.ahora, reloj.esperar)
			require.NoError(t, err)

			assert.Equal(t, decisionTrasLaEspera{
				decisionDeLaTanda: decisionDeLaTanda{mide: caso.mide, pendientes: caso.pendientes},
				agotada:           caso.agotada,
			}, decision)
			assert.Equal(t, caso.consultas, gh.consultas, "consultas")

			var esperas []time.Duration
			for range caso.consultas - 1 {
				esperas = append(esperas, 10*time.Second)
			}

			assert.Equal(t, esperas, reloj.esperas, "esperas")
		})
	}

	t.Run("consulta-que-falla", func(t *testing.T) {
		t.Parallel()

		errDeGh := errors.New("gh run list termin\xc3\xb3 con 1")
		fallida := func(context.Context) ([]ejecucionDelCommit, error) { return nil, errDeGh }

		var reloj relojSintetico

		_, err := esperarLaDecision(t.Context(), ejecucionPropia, fallida, reloj.ahora, reloj.esperar)
		require.ErrorIs(t, err, errDeGh)
		require.ErrorContains(t, err, "la consulta 1 de las ejecuciones del commit")
	})
}

// probarElEstadoDeLaTanda fija el estado de la tanda de una ejecución leído
// del JSON de sus trabajos (contracts/tanda-del-job.md §2 y §4 de H7.4;
// research.md S4 de H7.4): mide con su trabajo tanda terminado y la marca en
// success; no mide con la marca saltada o con la tanda saltada; y aún no ha
// decidido con la tanda en curso o sin ella.
func probarElEstadoDeLaTanda(t *testing.T) {
	t.Parallel()

	tandaEnCurso := trabajoSintetico(nombreDeLaTanda, ghEnCurso, "",
		pasoSintetico(pasoQueDecide, ghEnCurso, ""), pasoSintetico(pasoDeLaMarca, ghEnEspera, ""))
	tandaSaltada := trabajoSintetico(nombreDeLaTanda, ghTerminado, ghSaltado)

	casos := []struct {
		nombre   string
		trabajos string
		estado   estadoDeLaTanda
	}{
		{nombre: "marca-en-success", trabajos: trabajosQueMidenYCorren, estado: tandaQueMide},
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
