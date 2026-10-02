package evals

import (
	"context"
	"errors"
	"fmt"
	"maps"
	"os"
	"path/filepath"
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
	// líneas más sangradas.
	condicionDeLaTanda = sinCancelar + " && needs.cambios.result != 'failure' && (\n" +
		"  github.event_name == 'workflow_dispatch' ||\n" +
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
// etiqueta sobre un commit cuya tanda terminó vuelve a medir.
func TestDefinicionDelJob(t *testing.T) {
	t.Parallel()

	t.Run("del-repositorio", probarLaDefinicionDelRepositorio)
	t.Run("sinteticas", probarLasDefinicionesSinteticas)
	t.Run("errores", probarLosErroresDeLaDefinicion)
	t.Run("peor-caso", probarElPeorCasoDelTrabajo)
	t.Run("segundo-disparo", probarElSegundoDisparo)
	t.Run("estado-de-la-tanda", probarElEstadoDeLaTanda)
}

// probarLaDefinicionDelRepositorio lee la definición real del job y falla, con
// una línea por clave, si no es la del contrato o si su tope no cubre el peor
// caso de alguna skill con las evals del repositorio.
func probarLaDefinicionDelRepositorio(t *testing.T) {
	t.Parallel()

	leida, err := leerDefinicionDelJob(rutaDeLaDefinicionDelJob)
	require.NoError(t, err)

	if fallos := comprobarLaDefinicion(leida, directorioDeEvals); len(fallos) > 0 {
		t.Fatalf("la definici\xc3\xb3n del job no es la de contracts/tanda-del-job.md \xc2\xa74 de H7.4 y "+
			"contracts/ejecucion-del-job.md \xc2\xa77 de H7.3:\n%s", strings.Join(fallos, "\n"))
	}
}

// comprobarLaDefinicion devuelve una línea por cada clave del trabajo tanda y
// de la dependencia del trabajo evals de él que no es la de
// contracts/tanda-del-job.md §4 de H7.4, y por cada clave del trabajo evals que
// no es la de contracts/ejecucion-del-job.md §7 de H7.3, en el orden de los
// contratos, y una por cada skill de la matriz cuyo peor caso, con las evals de
// su carpeta dentro de evals, no cubre timeout-minutes o no se puede obtener.
// Sin ninguna línea, la definición es la de los contratos.
func comprobarLaDefinicion(leida DefinicionDelJob, evals string) []string {
	var fallos []string

	fallos = append(fallos, fallosDeLaTanda(leida.Tanda)...)
	fallos = append(fallos, fallosDeLaDependencia(leida)...)
	fallos = append(fallos, fallosDeLaConcurrencia(leida)...)
	fallos = append(fallos, fallosDeLaMatriz(leida)...)
	fallos = append(fallos, fallosDelEntorno(leida)...)
	fallos = append(fallos, fallosDelTope(leida, evals)...)

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

// fallosDelTope comprueba §7.4: para cada skill de la matriz, timeout-minutes
// cubre su peor caso, que nombra con sus términos.
func fallosDelTope(leida DefinicionDelJob, evals string) []string {
	var fallos []string

	tope := time.Duration(leida.TopeEnMinutos) * time.Minute

	for _, skill := range leida.Skills {
		peor, err := leida.peorCaso(evals, skill)
		if err != nil {
			fallos = append(fallos, fmt.Sprintf("jobs.evals.timeout-minutes: el peor caso de %s no se puede obtener: %v",
				skill, err))

			continue
		}

		if tope < peor.Duracion() {
			fallos = append(fallos, fallo("jobs.evals.timeout-minutes",
				fmt.Sprintf("vale %d (%d s)", leida.TopeEnMinutos, int(tope/time.Second)),
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
        github.event_name == 'workflow_dispatch' ||
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

// definicionDelContrato es una definición sintética del job con las claves de
// los contratos y el env de hoy (contracts/ejecucion-del-job.md §6 de H7.3;
// contracts/tanda-del-job.md §1 de H7.4). Con las evals de evalsSinteticas,
// cada skill tiene 7 sesiones en el modo orden —la eval con el modelo que decide
// y con el de Haiku, tres veces con cada uno, y la prueba de red— y 6 en el modo
// herramienta, así que el peor caso es de 1573 s en boe-legislacion (⌈7 / 4⌉ +
// ⌈6 / 4⌉ = 4 tandas) y de 4021 s en legal-core (13 tandas): 120 minutos los
// cubren.
const definicionDelContrato = `name: evals
on:
  pull_request:
    types: [opened, reopened, labeled]
jobs:
` + trabajoDeLaTandaDelContrato + `  evals:
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
      MODELOS_INFORMATIVOS_DE_EVALS: claude-haiku-4-5-20251001
      REPETICIONES_DE_EVALS: 3
      UMBRAL_DE_EVALS: 2
      CONCURRENCIA_DE_EVALS: ${{ matrix.concurrencia }}
      OBJETIVO_DE_DURACION_DE_EVALS: ${{ matrix.objetivo_de_duracion }}
`

// cambioDeLaDefinicion sustituye, en definicionDelContrato, un fragmento por
// otro.
type cambioDeLaDefinicion struct {
	antes   string
	despues string
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
// por comprobación de contracts/ejecucion-del-job.md §7 de H7.3, y los del tope
// que lo cubren.
func definicionesSinteticas() []definicionSintetica {
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
			nombre:  "tope-por-debajo-de-los-dos",
			cambios: []cambioDeLaDefinicion{{topeDelJob, "timeout-minutes: 26"}},
			fallos: []string{
				"jobs.evals.timeout-minutes: vale 26 (1560 s), y lo esperado es al menos el peor caso de " +
					"boe-legislacion, 1573 s = 485 s + (\xe2\x8c\x887 / 4\xe2\x8c\x89 + \xe2\x8c\x886 / 4\xe2\x8c\x89) " +
					"\xc3\x97 (22 s + 240 s + 10 s)",
				"jobs.evals.timeout-minutes: vale 26 (1560 s), y lo esperado es al menos el peor caso de " +
					"legal-core, " + peorCasoDeLegal,
			},
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

// evalsSinteticas crea un directorio temporal de evals con una carpeta por
// skill de la matriz, cada una con la eval sintética del art. 21 de la LPAC, y
// devuelve su ruta.
func evalsSinteticas(t *testing.T) string {
	t.Helper()

	evals := t.TempDir()

	for _, skill := range skillsDelTrabajo {
		require.NoError(t, os.Mkdir(filepath.Join(evals, skill), 0o750))
		require.NoError(t, os.WriteFile(filepath.Join(evals, skill, nombreDeEval), []byte(contenidoDelArticulo21), 0o600))
	}

	return evals
}

// probarLosErroresDeLaDefinicion fija lo que no se puede leer ni obtener: una
// definición que no existe, que no es YAML o cuyas repeticiones no son un
// entero no se lee, con un error que nombra la ruta; y el peor caso de una skill
// sin su carpeta de evals o con un plan que no se puede componer no se obtiene,
// con un error que dice por qué.
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

	leida.ModeloQueDecide = "Claude Sonnet"
	_, err = leida.peorCaso(evalsSinteticas(t), "boe-legislacion")
	require.ErrorContains(t, err, `el modelo que decide "Claude Sonnet" no tiene la forma de un id de modelo`)
}

// probarElPeorCasoDelTrabajo fija el peor caso del trabajo de una skill desde
// H21 (contracts/evals-en-dos-modos.md §6 de H21; research.md D20 de H21;
// FR-083 de H21): una tanda por cada grupo de sesiones —las del modo orden con
// la prueba de red, las del modo herramienta y las de las evals sin binario ni
// servidor— y no por la suma de todas. Con las evals del repositorio y la
// definición del job son 14 357 s en boe-legislacion y 12 181 s en legal-core,
// que cuenta la prueba de red aunque su trabajo no la lleve, y el tope de la
// definición los cubre.
func probarElPeorCasoDelTrabajo(t *testing.T) {
	t.Parallel()

	leida, err := leerDefinicionDelJob(rutaDeLaDefinicionDelJob)
	require.NoError(t, err)

	const terminosDeUnaTanda = " \xc3\x97 (22 s + 240 s + 10 s)"

	casos := []struct {
		skill    string
		peor     peorCasoDelTrabajo
		duracion time.Duration
		texto    string
	}{
		{
			skill:    "boe-legislacion",
			peor:     peorCasoDelTrabajo{SesionesPorGrupo: []int{97, 96, 6}, Concurrencia: 4},
			duracion: 14357 * time.Second,
			texto: "14357 s = 485 s + (\xe2\x8c\x8897 / 4\xe2\x8c\x89 + \xe2\x8c\x8896 / 4\xe2\x8c\x89 + " +
				"\xe2\x8c\x886 / 4\xe2\x8c\x89)" + terminosDeUnaTanda,
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
