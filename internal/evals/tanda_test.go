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
}

// TestSalidaDeLaTanda fija la línea que TestTandaDelCommit añade a -salida, el
// GITHUB_OUTPUT de su paso (contracts/tanda-del-job.md §3 de H7.4): medir=si o
// medir=no detrás de lo que ya tenga, creándolo si no existe; y un error que
// nombra la ruta si no se puede escribir.
func TestSalidaDeLaTanda(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()

	salida := filepath.Join(dir, "salida-del-paso")
	require.NoError(t, os.WriteFile(salida, []byte("antes=si\n"), 0o600))
	require.NoError(t, escribirLaDecision(salida, true))

	contenido, err := leerFichero(salida)
	require.NoError(t, err)
	assert.Equal(t, "antes=si\nmedir=si\n", string(contenido))

	nueva := filepath.Join(dir, "nueva")
	require.NoError(t, escribirLaDecision(nueva, false))

	contenido, err = leerFichero(nueva)
	require.NoError(t, err)
	assert.Equal(t, "medir=no\n", string(contenido))

	err = escribirLaDecision(dir, true)
	require.ErrorContains(t, err, dir)
}

// TestRegistroDeLaTanda fija lo que TestTandaDelCommit dice en su registro
// (contracts/tanda-del-job.md §3 de H7.4): de cada consulta, las ejecuciones
// anteriores a la propia que miró, en el orden de la lista, y lo que vio en
// cada una; y por qué mide o no.
func TestRegistroDeLaTanda(t *testing.T) {
	t.Parallel()

	ejecuciones := []ejecucionDelCommit{
		{id: 30, tanda: tandaSinDecidir},
		{id: ejecucionPropia, tanda: tandaSinDecidir},
		{id: 12, tanda: tandaQueMide},
		{id: 11, tanda: tandaQueNoMide},
		{id: 10, tanda: tandaSinDecidir},
		{id: 5, terminada: true, tanda: tandaSinDecidir},
	}

	assert.Equal(t, "consulta 2, ejecuciones anteriores a la 20 sobre el commit:\n"+
		"  12, sin terminar: su tanda mide el commit\n"+
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
			nombre:   "no-mide",
			decision: decisionTrasLaEspera{decisionDeLaTanda: decisionDeLaTanda{mide: false}},
			motivo:   "no mide el commit: lo mide la tanda de estas ejecuciones anteriores sin terminar: 12",
		},
		{
			nombre:   "mide",
			decision: decisionTrasLaEspera{decisionDeLaTanda: decisionDeLaTanda{mide: true}},
			motivo:   "mide el commit: ninguna ejecuci\xc3\xb3n anterior sin terminar lo mide",
		},
		{
			nombre: "espera-agotada",
			decision: decisionTrasLaEspera{
				decisionDeLaTanda: decisionDeLaTanda{mide: true, pendientes: []int64{10, 9}},
				agotada:           true,
			},
			motivo: "mide el commit: tras 10m0s, la tanda de estas ejecuciones anteriores sin terminar a\xc3\xban no ha " +
				"decidido: 10, 9; la concurrency del trabajo evals pone la de esta detr\xc3\xa1s de la que mida",
		},
	}

	for _, caso := range casos {
		t.Run(caso.nombre, func(t *testing.T) {
			t.Parallel()

			assert.Equal(t, caso.motivo, motivoDeLaDecision(ejecucionPropia, ejecuciones, caso.decision))
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
