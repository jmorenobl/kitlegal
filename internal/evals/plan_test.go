package evals

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// Los modelos con los que TestPlan compone sus planes: dos ids con la forma de un
// id de modelo y uno sin ella.
const (
	modeloQueDecideDelPlan   = "claude-sonnet-5"
	modeloInformativoDelPlan = "claude-haiku-4-5-20251001"
	modeloOtroInformativo    = "claude-opus-5"
	modeloConFormaInvalida   = "Claude Sonnet 5"
)

// planDeTres es el plan de una eval que decide, una de no activación y una
// informativa, con un modelo en los informativos y dos repeticiones.
func planDeTres() PlanDeEvals {
	return PlanDeEvals{
		Evals: []Eval{
			{Fichero: "01-lpac-articulo-21.yaml", Activa: true},
			{Fichero: "11-no-activa-programacion.yaml"},
			{Fichero: "13-lrbrl-por-materia.yaml", Activa: true, Informativa: true},
		},
		ModeloQueDecide:     modeloQueDecideDelPlan,
		ModelosInformativos: []string{modeloInformativoDelPlan},
		Repeticiones:        2,
	}
}

// TestPlan fija PlanDeEvals (contrato job-de-evals §3.2; data-model §10.4): qué
// series pide el plan y con qué modelo, que las evals informativas solo las abre
// el modelo que decide y sin decidir, los nombres y el orden de las sesiones, la
// sesión de la prueba de red y lo que Comprobar no admite.
func TestPlan(t *testing.T) {
	t.Parallel()

	t.Run("series", func(t *testing.T) {
		t.Parallel()

		assert.Equal(t, []SerieDeSesiones{
			{Eval: "01-lpac-articulo-21.yaml", Modelo: modeloQueDecideDelPlan, Decide: true},
			{Eval: "01-lpac-articulo-21.yaml", Modelo: modeloInformativoDelPlan},
			{Eval: "11-no-activa-programacion.yaml", Modelo: modeloQueDecideDelPlan, Decide: true},
			{Eval: "11-no-activa-programacion.yaml", Modelo: modeloInformativoDelPlan},
			{Eval: "13-lrbrl-por-materia.yaml", Modelo: modeloQueDecideDelPlan},
		}, planDeTres().Series(),
			"cada eval con el modelo que decide y, si no es informativa, con cada uno de los "+
				"informativos; la informativa no decide y ninguno de ellos la abre")
	})

	t.Run("sesiones", func(t *testing.T) {
		t.Parallel()

		nombres := []string{}
		for _, sesion := range planDeTres().Sesiones() {
			nombres = append(nombres, sesion.Nombre)
		}

		assert.Equal(t, []string{
			"01-lpac-articulo-21-claude-sonnet-5-01",
			"01-lpac-articulo-21-claude-sonnet-5-02",
			"01-lpac-articulo-21-claude-haiku-4-5-20251001-01",
			"01-lpac-articulo-21-claude-haiku-4-5-20251001-02",
			"11-no-activa-programacion-claude-sonnet-5-01",
			"11-no-activa-programacion-claude-sonnet-5-02",
			"11-no-activa-programacion-claude-haiku-4-5-20251001-01",
			"11-no-activa-programacion-claude-haiku-4-5-20251001-02",
			"13-lrbrl-por-materia-claude-sonnet-5-01",
			"13-lrbrl-por-materia-claude-sonnet-5-02",
		}, nombres, "las repeticiones de cada serie, numeradas con dos cifras para que el orden de nombre sea el suyo")
	})

	t.Run("cada-sesion-lleva-su-eval-y-su-modelo", func(t *testing.T) {
		t.Parallel()

		sesiones := planDeTres().Sesiones()
		require.Len(t, sesiones, 10)

		assert.Equal(t, SesionPlanificada{
			Nombre: "01-lpac-articulo-21-claude-haiku-4-5-20251001-02", Fichero: "01-lpac-articulo-21.yaml",
			Modelo: modeloInformativoDelPlan,
		}, sesiones[3])
	})

	t.Run("prueba-de-red", func(t *testing.T) {
		t.Parallel()

		plan := planDeTres()
		plan.PruebaDeRed = true

		sesiones := plan.Sesiones()
		require.Len(t, sesiones, 11, "la prueba de red añade una sola sesión, que no se repite")
		assert.Equal(t, SesionPlanificada{
			Nombre:  "01-lpac-articulo-21-prueba-de-red-claude-sonnet-5-01",
			Fichero: "01-lpac-articulo-21.yaml", Modelo: modeloQueDecideDelPlan, PruebaDeRed: true,
		}, sesiones[10], "va al final, con el modelo que decide y con la primera eval")

		assert.Equal(t, planDeTres().Series(), plan.Series(),
			"la prueba de red no añade ninguna serie: no pregunta lo que pregunta la eval")
	})

	t.Run("sin-evals", func(t *testing.T) {
		t.Parallel()

		plan := planDeTres()
		plan.Evals = nil
		plan.PruebaDeRed = true

		assert.Empty(t, plan.Series())
		assert.Empty(t, plan.Sesiones(), "sin evals no hay ni siquiera sesión de prueba de red")
	})

	t.Run("sin-modelos-informativos", func(t *testing.T) {
		t.Parallel()

		plan := planDeTres()
		plan.ModelosInformativos = nil

		require.NoError(t, plan.Comprobar())
		assert.Len(t, plan.Series(), 3, "una serie por eval, la del modelo que decide")
	})
}

// TestPlanComprobar fija lo que PlanDeEvals.Comprobar no admite: un modelo sin la
// forma de un id, uno de los informativos que es el que decide o está repetido y
// menos de una repetición. El error nombra lo que falla.
func TestPlanComprobar(t *testing.T) {
	t.Parallel()

	casos := []struct {
		nombre    string
		modificar func(plan *PlanDeEvals)
		fragmento string
	}{
		{
			nombre:    "cumple",
			modificar: func(*PlanDeEvals) {},
		},
		{
			nombre:    "modelo-que-decide-con-otra-forma",
			modificar: func(plan *PlanDeEvals) { plan.ModeloQueDecide = modeloConFormaInvalida },
			fragmento: modeloConFormaInvalida,
		},
		{
			nombre:    "sin-modelo-que-decide",
			modificar: func(plan *PlanDeEvals) { plan.ModeloQueDecide = "" },
			fragmento: "el modelo que decide",
		},
		{
			nombre:    "modelos-informativos-con-otra-forma",
			modificar: func(plan *PlanDeEvals) { plan.ModelosInformativos = []string{modeloConFormaInvalida} },
			fragmento: modeloConFormaInvalida,
		},
		{
			nombre:    "modelos-informativos-con-el-que-decide",
			modificar: func(plan *PlanDeEvals) { plan.ModelosInformativos = []string{modeloQueDecideDelPlan} },
			fragmento: "es el que decide",
		},
		{
			nombre: "modelos-informativos-repetidos",
			modificar: func(plan *PlanDeEvals) {
				plan.ModelosInformativos = []string{modeloInformativoDelPlan, modeloOtroInformativo, modeloInformativoDelPlan}
			},
			fragmento: "dos veces",
		},
		{
			nombre:    "sin-repeticiones",
			modificar: func(plan *PlanDeEvals) { plan.Repeticiones = 0 },
			fragmento: "al menos una",
		},
	}

	for _, caso := range casos {
		t.Run(caso.nombre, func(t *testing.T) {
			t.Parallel()

			plan := planDeTres()
			caso.modificar(&plan)

			err := plan.Comprobar()
			if caso.fragmento == "" {
				assert.NoError(t, err)

				return
			}

			require.Error(t, err)
			assert.Contains(t, err.Error(), caso.fragmento)
		})
	}
}
