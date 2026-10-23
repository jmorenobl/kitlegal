package evals

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestOrdenesDeLaTanda fija las órdenes de gh con las que TestTandaDelCommit
// consulta las ejecuciones del flujo sobre el commit
// (contracts/tanda-del-job.md §2 de H7.4; research.md D16 y V12 de H7.4), con
// el sustituto de gh como único programa del PATH de las órdenes: cada una da a
// gh los argumentos de su orden constante, con el valor de su variable entero y
// sin interpretarlo, aunque lleve espacios, comillas o $( ); ese valor
// sustituye al que el entorno tenga con el mismo nombre; y una orden que
// termina con otro código que 0 es un error que la nombra, con su variable y lo
// que escribió en sus dos salidas.
func TestOrdenesDeLaTanda(t *testing.T) {
	t.Parallel()

	// El valor que el entorno de la ejecución ya tiene no es el de la orden.
	base := sobreLaBase(os.Environ(), []string{"PATH=" + escribirElGh(t), "COMMIT_EVALUADO=el-del-entorno"})

	const commit = `6ab3add $(printf inyectado) "con comillas"`

	gh := ghDeLaTanda{entorno: base, commit: commit}

	lista, err := gh.listar(t.Context())
	require.NoError(t, err)
	assert.Equal(t, []string{
		"run", "list", "--workflow", "evals.yml", "--commit", commit, "--limit", "100", "--json", "databaseId,status",
	}, strings.Split(strings.TrimSuffix(string(lista), "\n"), "\n"))

	trabajos, err := gh.verLosTrabajos(t.Context(), 36672667544)
	require.NoError(t, err)
	assert.Equal(t, []string{"run", "view", "36672667544", "--json", "jobs"},
		strings.Split(strings.TrimSuffix(string(trabajos), "\n"), "\n"))

	fallido := ghDeLaTanda{entorno: sobreLaBase(base, []string{variableDeGhQueFalla + "=si"}), commit: "6ab3add"}

	_, err = fallido.listar(t.Context())
	require.ErrorContains(t, err, ordenDeLasEjecuciones)
	require.ErrorContains(t, err, "COMMIT_EVALUADO=6ab3add")
	require.ErrorContains(t, err, "exit status 4")
	require.ErrorContains(t, err, "lo escrito antes del error")
	require.ErrorContains(t, err, "HTTP 403: Resource not accessible by integration")

	_, err = fallido.verLosTrabajos(t.Context(), 36672667544)
	require.ErrorContains(t, err, ordenDeLosTrabajos)
	require.ErrorContains(t, err, "EJECUCION_ANTERIOR=36672667544")
	require.ErrorContains(t, err, "exit status 4")
	require.ErrorContains(t, err, "HTTP 403: Resource not accessible by integration")

	// Lo que da la orden de los trabajos lleva los pasos de la tanda con sus dos
	// marcas, de las que sale su modo (data-model §10 de H30).
	respuesta := t.TempDir()
	require.NoError(t, os.WriteFile(filepath.Join(respuesta, respuestaDeGh),
		[]byte(trabajosQueMidenLaSuiteCompletaYCorren+"\n"), 0o600))

	conLosTrabajos := ghDeLaTanda{
		entorno: sobreLaBase(base, []string{variableDeLaRespuestaDeGh + "=" + respuesta}), commit: "6ab3add",
	}

	trabajos, err = conLosTrabajos.verLosTrabajos(t.Context(), 36672667544)
	require.NoError(t, err)

	estado, err := leerLaTandaDeLaEjecucion(trabajos)
	require.NoError(t, err)
	assert.Equal(t, tandaQueMideLaSuiteCompleta, estado)
}

// TestSalidaDeLaTanda fija las líneas que TestTandaDelCommit añade a -salida,
// el GITHUB_OUTPUT de su paso (contracts/tanda-del-job.md §3 de H7.4;
// contracts/job-de-evals.md §2 de H30): medir=si o medir=no y modo=<modo>
// detrás de lo que ya tenga, creándolo si no existe; y un error que nombra la
// ruta si no se puede escribir.
func TestSalidaDeLaTanda(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()

	salida := filepath.Join(dir, "salida-del-paso")
	require.NoError(t, os.WriteFile(salida, []byte("antes=si\n"), 0o600))
	require.NoError(t, escribirLaDecision(salida, decisionDeLaTanda{mide: true, modo: origenDeLaSuiteCompleta}))

	contenido, err := leerFichero(salida)
	require.NoError(t, err)
	assert.Equal(t, "antes=si\nmedir=si\nmodo=completa\n", string(contenido))

	nueva := filepath.Join(dir, "nueva")
	require.NoError(t, escribirLaDecision(nueva, decisionDeLaTanda{mide: false, modo: origenConReutilizacion}))

	contenido, err = leerFichero(nueva)
	require.NoError(t, err)
	assert.Equal(t, "medir=no\nmodo=reutilizacion\n", string(contenido))

	err = escribirLaDecision(dir, decisionDeLaTanda{mide: true, modo: origenConReutilizacion})
	require.ErrorContains(t, err, dir)
}

// TestRegistroDeLaTanda fija lo que TestTandaDelCommit dice en su registro
// (contracts/tanda-del-job.md §3 de H7.4), con el modo (data-model §10 de H30):
// de cada consulta, las ejecuciones anteriores a la propia que miró, en el
// orden de la lista, y lo que vio en cada una, con el modo de las que miden; y
// por qué mide o no con su modo, nombrando, si no mide, solo las anteriores que
// miden con un modo igual o mayor.
func TestRegistroDeLaTanda(t *testing.T) {
	t.Parallel()

	ejecuciones := []ejecucionDelCommit{
		{id: 30, tanda: tandaSinDecidir},
		{id: ejecucionPropia, tanda: tandaSinDecidir},
		{id: 13, tanda: tandaQueMideLaSuiteCompleta},
		{id: 12, tanda: tandaQueMideConReutilizacion},
		{id: 11, tanda: tandaQueNoMide},
		{id: 10, tanda: tandaSinDecidir},
		{id: 5, terminada: true, tanda: tandaQueMideLaSuiteCompleta},
	}

	assert.Equal(t, "consulta 2, ejecuciones anteriores a la 20 sobre el commit:\n"+
		"  13, sin terminar: su tanda mide el commit en modo completa\n"+
		"  12, sin terminar: su tanda mide el commit en modo reutilizacion\n"+
		"  11, sin terminar: su tanda no mide el commit\n"+
		"  10, sin terminar: su tanda a\xc3\xban no ha decidido\n"+
		"  5, terminada: no cuenta, su tanda ya termin\xc3\xb3",
		registroDeLaConsulta(2, ejecucionPropia, ejecuciones))

	assert.Equal(t, "consulta 1: ninguna ejecuci\xc3\xb3n anterior a la 20 sobre el commit",
		registroDeLaConsulta(1, ejecucionPropia, ejecuciones[:2]))

	casos := []struct {
		nombre   string
		decision decisionTrasLaEspera
		motivo   string
	}{
		{
			nombre: "no-mide-con-reutilizacion",
			decision: decisionTrasLaEspera{
				decisionDeLaTanda: decisionDeLaTanda{mide: false, modo: origenConReutilizacion},
			},
			motivo: "no mide el commit en modo reutilizacion: lo mide con un modo igual o mayor la tanda de estas " +
				"ejecuciones anteriores sin terminar: 13, 12",
		},
		{
			nombre: "no-mide-la-suite-completa",
			decision: decisionTrasLaEspera{
				decisionDeLaTanda: decisionDeLaTanda{mide: false, modo: origenDeLaSuiteCompleta},
			},
			motivo: "no mide el commit en modo completa: lo mide con un modo igual o mayor la tanda de estas " +
				"ejecuciones anteriores sin terminar: 13",
		},
		{
			nombre: "mide",
			decision: decisionTrasLaEspera{
				decisionDeLaTanda: decisionDeLaTanda{mide: true, modo: origenDeLaSuiteCompleta},
			},
			motivo: "mide el commit en modo completa: ninguna ejecuci\xc3\xb3n anterior sin terminar lo mide con un " +
				"modo igual o mayor",
		},
		{
			nombre: "espera-agotada",
			decision: decisionTrasLaEspera{
				decisionDeLaTanda: decisionDeLaTanda{
					mide: true, modo: origenConReutilizacion, pendientes: []int64{10, 9},
				},
				agotada: true,
			},
			motivo: "mide el commit en modo reutilizacion: tras 10m0s, la tanda de estas ejecuciones anteriores sin " +
				"terminar a\xc3\xban no ha decidido: 10, 9; la concurrency del trabajo evals pone la de esta " +
				"detr\xc3\xa1s de la que mida",
		},
	}

	for _, caso := range casos {
		t.Run(caso.nombre, func(t *testing.T) {
			t.Parallel()

			assert.Equal(t, caso.motivo, motivoDeLaDecision(ejecucionPropia, ejecuciones, caso.decision))
		})
	}
}

// TestTandaConLaSuiteCompleta fija la decisión de la tanda con ejecuciones de
// modos distintos sobre el mismo commit (FR-021; data-model §10 de H30;
// contracts/job-de-evals.md §2 de H30): una ejecución no mide si una anterior
// sin terminar mide con un modo igual o mayor, con completa mayor que
// reutilizacion, y la de una anterior que mide con un modo menor no cuenta. Así
// una con reutilización no mide con una de la suite completa en curso, y una de
// la suite completa sí mide con una con reutilización en curso; también cuando
// la anterior aún no ha decidido y decide en la consulta siguiente.
func TestTandaConLaSuiteCompleta(t *testing.T) {
	t.Parallel()

	propiaEnCurso := ejecucionSintetica(ejecucionPropia, ghEnCurso)
	conLaAnteriorEnCurso := ejecucionesSinteticas(propiaEnCurso, ejecucionSintetica(10, ghEnCurso))
	conLasDosAnterioresEnCurso := ejecucionesSinteticas(propiaEnCurso, ejecucionSintetica(11, ghEnCurso),
		ejecucionSintetica(10, ghEnCurso))

	casos := []segundoDisparo{
		{
			nombre: "con-reutilizacion-y-una-completa-en-curso",
			modo:   origenConReutilizacion,
			rondas: []rondaSintetica{{
				conLaAnteriorEnCurso, map[int64]string{10: trabajosQueMidenLaSuiteCompletaYCorren},
			}},
			consultas: 1,
		},
		{
			nombre:    "completa-y-una-con-reutilizacion-en-curso",
			modo:      origenDeLaSuiteCompleta,
			rondas:    []rondaSintetica{{conLaAnteriorEnCurso, map[int64]string{10: trabajosQueMidenYCorren}}},
			mide:      true,
			consultas: 1,
		},
		{
			nombre: "completa-y-una-completa-en-curso",
			modo:   origenDeLaSuiteCompleta,
			rondas: []rondaSintetica{{
				conLaAnteriorEnCurso, map[int64]string{10: trabajosQueMidenLaSuiteCompletaYCorren},
			}},
			consultas: 1,
		},
		{
			nombre:    "con-reutilizacion-y-una-con-reutilizacion-en-curso",
			modo:      origenConReutilizacion,
			rondas:    []rondaSintetica{{conLaAnteriorEnCurso, map[int64]string{10: trabajosQueMidenYCorren}}},
			consultas: 1,
		},
		{
			// La primera de la lista, la más reciente, mide con un modo menor:
			// no decide nada, y la siguiente, completa, sí.
			nombre: "completa-con-una-con-reutilizacion-y-una-completa-en-curso",
			modo:   origenDeLaSuiteCompleta,
			rondas: []rondaSintetica{{conLasDosAnterioresEnCurso, map[int64]string{
				11: trabajosQueMidenYCorren, 10: trabajosQueMidenLaSuiteCompletaYCorren,
			}}},
			consultas: 1,
		},
		{
			nombre: "con-reutilizacion-tras-decidir-una-completa",
			modo:   origenConReutilizacion,
			rondas: []rondaSintetica{
				{conLaAnteriorEnCurso, map[int64]string{10: trabajosSinDecidir}},
				{conLaAnteriorEnCurso, map[int64]string{10: trabajosQueMidenLaSuiteCompletaYCorren}},
			},
			consultas: 2,
		},
		{
			nombre: "completa-tras-decidir-una-con-reutilizacion",
			modo:   origenDeLaSuiteCompleta,
			rondas: []rondaSintetica{
				{conLaAnteriorEnCurso, map[int64]string{10: trabajosSinDecidir}},
				{conLaAnteriorEnCurso, map[int64]string{10: trabajosQueMidenYCorren}},
			},
			mide:      true,
			consultas: 2,
		},
	}

	for _, caso := range casos {
		t.Run(caso.nombre, func(t *testing.T) {
			t.Parallel()

			probarUnDisparo(t, caso)
		})
	}
}

// TestDormir fija la espera de verdad de TestTandaDelCommit entre consulta y
// consulta: vuelve sin error al cumplirse, y con el error del contexto si este
// termina antes.
func TestDormir(t *testing.T) {
	t.Parallel()

	require.NoError(t, dormir(t.Context(), time.Millisecond))

	cancelado, cancelar := context.WithCancel(t.Context())
	cancelar()

	require.ErrorIs(t, dormir(cancelado, time.Hour), context.Canceled)
}
