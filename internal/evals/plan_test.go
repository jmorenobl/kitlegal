package evals

import (
	"maps"
	"path/filepath"
	"slices"
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
			{Eval: "01-lpac-articulo-21.yaml", Modelo: modeloQueDecideDelPlan, Modo: ModoOrden, Decide: true},
			{Eval: "01-lpac-articulo-21.yaml", Modelo: modeloInformativoDelPlan, Modo: ModoOrden},
			{Eval: "11-no-activa-programacion.yaml", Modelo: modeloQueDecideDelPlan, Modo: ModoOrden, Decide: true},
			{Eval: "11-no-activa-programacion.yaml", Modelo: modeloInformativoDelPlan, Modo: ModoOrden},
			{Eval: "13-lrbrl-por-materia.yaml", Modelo: modeloQueDecideDelPlan, Modo: ModoOrden},
		}, planDeTres().Series(),
			"cada eval con el modelo que decide y, si no es informativa, con cada uno de los "+
				"informativos; la informativa no decide y ninguno de ellos la abre; sin modos, el modo es orden")
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
			Modelo: modeloInformativoDelPlan, Modo: ModoOrden,
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
			Fichero: "01-lpac-articulo-21.yaml", Modelo: modeloQueDecideDelPlan, Modo: ModoOrden, PruebaDeRed: true,
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

// ficheroSinBinarioDelPlan es la eval sin binario ni servidor de planEnDosModos.
const ficheroSinBinarioDelPlan = "05-sin-binario-ni-servidor.yaml"

// planEnDosModos es el plan de planDeTres con los dos modos y con una eval sin
// binario ni servidor, que va la segunda para que se vea que sale de entre las
// de modo.
func planEnDosModos() PlanDeEvals {
	plan := planDeTres()
	plan.Evals = slices.Insert(plan.Evals, 1, Eval{
		Fichero: ficheroSinBinarioDelPlan, Activa: true, SinBinarioNiServidor: true,
	})
	plan.Modos = []Modo{ModoOrden, ModoHerramienta}

	return plan
}

// TestPlanEnDosModos fija el plan del job desde H21 (contracts/evals-en-dos-modos.md
// §2.1 y §8; data-model §6; FR-040, FR-046): una eval de modo da por cada modo
// las series de hoy; una eval sin binario ni servidor da las suyas una sola vez
// y sin modo; las sesiones van en tres tandas —modo orden, con la prueba de red,
// modo herramienta y sin binario ni servidor—, con el nombre de hoy salvo en el
// modo herramienta; con las evals del repositorio, las series y las sesiones de
// cada tanda son las del contrato; y un plan sin modos, el del sondeo, es el de
// hoy.
func TestPlanEnDosModos(t *testing.T) {
	t.Parallel()

	t.Run("textos", probarLosTextosDeLosModos)
	t.Run("series", probarLasSeriesEnDosModos)
	t.Run("sesiones", probarLasSesionesEnDosModos)
	t.Run("prueba-de-red", probarLaPruebaDeRedEnDosModos)
	t.Run("del-repositorio", probarElPlanDelRepositorio)
	t.Run("sondeo", probarElPlanDelSondeo)
}

// probarLosTextosDeLosModos fija el texto de cada modo, que es el que llevan el
// informe y sus umbrales (data-model §6).
func probarLosTextosDeLosModos(t *testing.T) {
	t.Parallel()

	assert.Equal(t, "orden", string(ModoOrden))
	assert.Equal(t, "herramienta", string(ModoHerramienta))
}

// probarLasSeriesEnDosModos fija las series: las de hoy por cada modo, en el
// orden de los modos, y al final las de la eval sin binario ni servidor, una vez
// y sin modo, con el modelo que decide, que decide, y con los informativos.
func probarLasSeriesEnDosModos(t *testing.T) {
	t.Parallel()

	var esperadas []SerieDeSesiones

	for _, modo := range []Modo{ModoOrden, ModoHerramienta} {
		esperadas = append(esperadas,
			SerieDeSesiones{Eval: "01-lpac-articulo-21.yaml", Modelo: modeloQueDecideDelPlan, Modo: modo, Decide: true},
			SerieDeSesiones{Eval: "01-lpac-articulo-21.yaml", Modelo: modeloInformativoDelPlan, Modo: modo},
			SerieDeSesiones{Eval: "11-no-activa-programacion.yaml", Modelo: modeloQueDecideDelPlan, Modo: modo, Decide: true},
			SerieDeSesiones{Eval: "11-no-activa-programacion.yaml", Modelo: modeloInformativoDelPlan, Modo: modo},
			SerieDeSesiones{Eval: "13-lrbrl-por-materia.yaml", Modelo: modeloQueDecideDelPlan, Modo: modo},
		)
	}

	esperadas = append(esperadas,
		SerieDeSesiones{Eval: ficheroSinBinarioDelPlan, Modelo: modeloQueDecideDelPlan, Decide: true},
		SerieDeSesiones{Eval: ficheroSinBinarioDelPlan, Modelo: modeloInformativoDelPlan},
	)

	assert.Equal(t, esperadas, planEnDosModos().Series())
}

// probarLasSesionesEnDosModos fija las sesiones, con la prueba de red: sus
// nombres, las tres tandas en su orden y lo que lleva una sesión de cada una.
func probarLasSesionesEnDosModos(t *testing.T) {
	t.Parallel()

	plan := planEnDosModos()
	plan.PruebaDeRed = true

	sesiones := plan.Sesiones()

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
		"01-lpac-articulo-21-prueba-de-red-claude-sonnet-5-01",
		"01-lpac-articulo-21-herramienta-claude-sonnet-5-01",
		"01-lpac-articulo-21-herramienta-claude-sonnet-5-02",
		"01-lpac-articulo-21-herramienta-claude-haiku-4-5-20251001-01",
		"01-lpac-articulo-21-herramienta-claude-haiku-4-5-20251001-02",
		"11-no-activa-programacion-herramienta-claude-sonnet-5-01",
		"11-no-activa-programacion-herramienta-claude-sonnet-5-02",
		"11-no-activa-programacion-herramienta-claude-haiku-4-5-20251001-01",
		"11-no-activa-programacion-herramienta-claude-haiku-4-5-20251001-02",
		"13-lrbrl-por-materia-herramienta-claude-sonnet-5-01",
		"13-lrbrl-por-materia-herramienta-claude-sonnet-5-02",
		"05-sin-binario-ni-servidor-claude-sonnet-5-01",
		"05-sin-binario-ni-servidor-claude-sonnet-5-02",
		"05-sin-binario-ni-servidor-claude-haiku-4-5-20251001-01",
		"05-sin-binario-ni-servidor-claude-haiku-4-5-20251001-02",
	}, nombresDe(sesiones), "los nombres de hoy, salvo en el modo herramienta, que lo lleva tras la eval")

	assert.Equal(t, slices.Concat(
		slices.Repeat([]Modo{ModoOrden}, 11), slices.Repeat([]Modo{ModoHerramienta}, 10), slices.Repeat([]Modo{""}, 4),
	), modosDe(sesiones), "tres tandas: modo orden, con la prueba de red, modo herramienta y sin binario ni servidor")

	assert.Equal(t, SesionPlanificada{
		Nombre:  "01-lpac-articulo-21-prueba-de-red-claude-sonnet-5-01",
		Fichero: "01-lpac-articulo-21.yaml", Modelo: modeloQueDecideDelPlan, Modo: ModoOrden, PruebaDeRed: true,
	}, sesiones[10], "la prueba de red cierra la tanda del modo orden")
	assert.Equal(t, SesionPlanificada{
		Nombre:  "01-lpac-articulo-21-herramienta-claude-haiku-4-5-20251001-02",
		Fichero: "01-lpac-articulo-21.yaml", Modelo: modeloInformativoDelPlan, Modo: ModoHerramienta,
	}, sesiones[14])
	assert.Equal(t, SesionPlanificada{
		Nombre:  "05-sin-binario-ni-servidor-claude-haiku-4-5-20251001-01",
		Fichero: ficheroSinBinarioDelPlan, Modelo: modeloInformativoDelPlan,
	}, sesiones[23], "la eval sin binario ni servidor se abre una sola vez, sin modo")
}

// probarLaPruebaDeRedEnDosModos fija que la prueba de red es una sesión del
// modo orden y no añade ninguna serie: su eval es la primera de modo, también
// si la primera del conjunto es la eval sin binario ni servidor, y un plan sin
// ninguna eval de modo no la lleva.
func probarLaPruebaDeRedEnDosModos(t *testing.T) {
	t.Parallel()

	sinBinario := Eval{Fichero: "00-sin-binario-ni-servidor.yaml", Activa: true, SinBinarioNiServidor: true}

	plan := planEnDosModos()
	plan.PruebaDeRed = true

	assert.Equal(t, planEnDosModos().Series(), plan.Series(), "la prueba de red no añade ninguna serie")

	plan.Evals = slices.Concat([]Eval{sinBinario}, planDeTres().Evals)

	deRed := slices.DeleteFunc(plan.Sesiones(), func(sesion SesionPlanificada) bool { return !sesion.PruebaDeRed })
	assert.Equal(t, []SesionPlanificada{{
		Nombre:  "01-lpac-articulo-21-prueba-de-red-claude-sonnet-5-01",
		Fichero: "01-lpac-articulo-21.yaml", Modelo: modeloQueDecideDelPlan, Modo: ModoOrden, PruebaDeRed: true,
	}}, deRed, "una sola, del modo orden y con la primera eval de modo")

	plan.Evals = []Eval{sinBinario}

	assert.Equal(t, []Modo{"", "", "", ""}, modosDe(plan.Sesiones()),
		"sin evals de modo no hay sesión de prueba de red, y la eval sin binario ni servidor no la lleva")
}

// probarElPlanDelRepositorio fija, con las evals del repositorio y los modelos y
// las repeticiones de la definición del job, las series y las sesiones de cada
// tanda del plan de cada skill, las de la tabla de contracts/evals-en-dos-modos.md
// §2.1: 96, 96 y 6 sesiones en boe-legislacion y 18, 18 y 6 en legal-core, y
// una más en el modo orden con la prueba de red.
func probarElPlanDelRepositorio(t *testing.T) {
	t.Parallel()

	delJob, err := leerDefinicionDelJob(rutaDeLaDefinicionDelJob)
	require.NoError(t, err)

	casos := []struct {
		skill    string
		series   map[Modo]int
		sesiones map[Modo]int
	}{
		{
			skill:    "boe-legislacion",
			series:   map[Modo]int{ModoOrden: 32, ModoHerramienta: 32, "": 2},
			sesiones: map[Modo]int{ModoOrden: 96, ModoHerramienta: 96, "": 6},
		},
		{
			skill:    "legal-core",
			series:   map[Modo]int{ModoOrden: 6, ModoHerramienta: 6, "": 2},
			sesiones: map[Modo]int{ModoOrden: 18, ModoHerramienta: 18, "": 6},
		},
	}

	for _, caso := range casos {
		t.Run(caso.skill, func(t *testing.T) {
			t.Parallel()

			conjunto, err := LeerConjunto(filepath.Join(directorioDeEvals, caso.skill))
			require.NoError(t, err)
			require.Empty(t, conjunto.MalFormados)

			plan := PlanDeEvals{
				Evals:               conjunto.Evals,
				ModeloQueDecide:     delJob.ModeloQueDecide,
				ModelosInformativos: delJob.ModelosInformativos,
				Repeticiones:        delJob.Repeticiones,
				Modos:               []Modo{ModoOrden, ModoHerramienta},
			}
			require.NoError(t, plan.Comprobar())

			series := map[Modo]int{}
			for _, serie := range plan.Series() {
				series[serie.Modo]++
			}

			assert.Equal(t, caso.series, series, "series por modo")
			assert.Equal(t, caso.sesiones, sesionesPorModo(plan.Sesiones()), "sesiones por modo")

			plan.PruebaDeRed = true

			conLaDeRed := maps.Clone(caso.sesiones)
			conLaDeRed[ModoOrden]++

			assert.Equal(t, conLaDeRed, sesionesPorModo(plan.Sesiones()), "sesiones por modo con la prueba de red")
			assert.Equal(t, []Modo{ModoOrden, ModoHerramienta, ""}, slices.Compact(modosDe(plan.Sesiones())),
				"tres tandas, en su orden")
		})
	}
}

// probarElPlanDelSondeo fija que un plan sin modos, el del sondeo, es el de hoy:
// el mismo que el que pide solo el modo orden, con todas sus sesiones en ese
// modo y con los nombres de hoy, que fija TestPlan.
func probarElPlanDelSondeo(t *testing.T) {
	t.Parallel()

	delSondeo := planDeTres()
	delSondeo.PruebaDeRed = true
	require.Empty(t, delSondeo.Modos, "premisa: el plan del sondeo no da modos")

	soloOrden := delSondeo
	soloOrden.Modos = []Modo{ModoOrden}

	assert.Equal(t, soloOrden.Series(), delSondeo.Series())
	assert.Equal(t, soloOrden.Sesiones(), delSondeo.Sesiones())
	assert.Equal(t, slices.Repeat([]Modo{ModoOrden}, 11), modosDe(delSondeo.Sesiones()))
}

// modosDe son los modos de las sesiones, en su orden.
func modosDe(sesiones []SesionPlanificada) []Modo {
	modos := make([]Modo, 0, len(sesiones))
	for _, sesion := range sesiones {
		modos = append(modos, sesion.Modo)
	}

	return modos
}

// sesionesPorModo es cuántas sesiones hay de cada modo.
func sesionesPorModo(sesiones []SesionPlanificada) map[Modo]int {
	porModo := map[Modo]int{}
	for _, sesion := range sesiones {
		porModo[sesion.Modo]++
	}

	return porModo
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
