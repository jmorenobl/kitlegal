package evals

import (
	"encoding/json"
	"errors"
	"io/fs"
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

// Lo que comparten los casos de TestDecisionDelJob: la propuesta cuyos ficheros
// da la plataforma en memoria; la medida de main sintética, su ejecución, la
// propuesta fusionada de cuya cabeza es y el tiempo entre su creación y la
// decisión;
// las repeticiones del job; y la eval de legal-core que en algún caso no tiene
// resultado en main o no tiene conjunto.
const (
	propuestaDelJob       = 210
	ejecucionDeMainDelJob = 301
	propuestaDeMainDelJob = 199
	antesDeLaMedidaDeMain = time.Hour
	repeticionesDelJob    = 3
	evalSinResultado      = "legal-core/05-plazo-reposicion-leganes"
)

// disparoDelJob es lo que lanza el job en un caso de TestDecisionDelJob
// (contracts/job-de-evals.md §1 de H30): la apertura de la propuesta, que solo
// corre los trabajos de evals si elige alguna; la etiqueta evals, que los corre
// siempre con reutilización; y la suite completa.
type disparoDelJob string

const (
	disparoDeLaApertura disparoDelJob = "apertura"
	disparoDeLaEtiqueta disparoDelJob = "etiqueta"
	disparoCompleto     disparoDelJob = "completa"
)

// medidasEsperadas son las evals que mide el trabajo de una skill en un caso:
// las de ficheros, o todas si no lleva ninguno, con su motivo y las rutas que
// las eligen.
type medidasEsperadas struct {
	motivo   motivoDeLaEleccion
	rutas    []string
	ficheros []string
}

// casoDeLaDecision es un caso de TestDecisionDelJob: lo que lo lanza, la
// propuesta con sus ficheros, si hay medida de main y con qué huellas
// distintas de las de la cabeza, los fallos de la plataforma, si el conjunto
// de evalSinResultado no se puede calcular, y lo que espera: si los trabajos
// corren, lo que mide cada skill —las que no están, todo reutilizado de la
// medida de main—, el error de la plataforma y las lecturas de cada trabajo.
type casoDeLaDecision struct {
	nombre       string
	disparo      disparoDelJob
	mutante      mutanteDeLaPropuesta
	sinMain      bool
	otrasHuellas map[string]string
	fallos       map[string]error
	incalculable bool

	sinTrabajos       bool
	medidas           map[string]medidasEsperadas
	plataformaFallida *errorDeLaPlataformaEnElOrigen
	sinLecturas       bool
	soloLasLecturas   []string
}

// repositorioDeLaDecision es lo que los casos de TestDecisionDelJob leen una
// vez del repositorio: su árbol, su grafo de paquetes, su árbol como base de la
// propuesta, lo común, sus skills con evals, en orden, y las evals de cada una.
type repositorioDeLaDecision struct {
	arbol  fs.FS
	grafo  grafoDePaquetes
	base   arbolDelRepositorio
	comun  loComun
	skills []string
	evals  map[string][]Eval
}

// TestDecisionDelJob fija lo que mide y lo que reutiliza cada trabajo del job
// (FR-012 a FR-014, FR-016 a FR-018; SC-006, SC-007; contracts/job-de-evals.md
// §1 y §2 de H30; research D17 de H30), sobre los conjuntos del repositorio, con
// la plataforma en memoria y una medida de main sintética con la huella de cada
// eval —una huella sintética por conjunto, la que da T006 en el job—, un
// subtest por caso. Cada uno compara el origen de la decisión de cada trabajo,
// eval por eval, sus líneas del registro y su fichero, las lecturas de la
// plataforma, cada una con el plazo de 10 minutos de la decisión, y cuenta las
// sesiones del plan de cada trabajo con las que espera,
// que salen del plan de las evals que espera medir: sale en rojo con una sola de
// más.
func TestDecisionDelJob(t *testing.T) {
	t.Parallel()

	repositorio := leerElRepositorioDeLaDecision(t)

	docs := conUnaLineaMas(t, repositorio.arbol, ficheroDelRepositorio(t, repositorio.arbol, "docs", esMarkdown))
	cuerpo := "skills/jurisprudencia/SKILL.md"
	deJurisprudencia := map[string]medidasEsperadas{
		"jurisprudencia": {motivo: motivoConjunto, rutas: []string{cuerpo}},
	}
	errDeLosFicheros := &errorDeUnaOrden{
		orden: ordenDeLosFicherosDeLaPropuesta, variables: []string{"PROPUESTA=" + strconv.Itoa(propuestaDelJob)},
		causa: errors.New("exit status 1"), salidaDeError: "HTTP 502: Bad Gateway (https://api.github.com/repos/o/r/pulls/210/files)\n",
	}

	casos := []casoDeLaDecision{
		{
			nombre: "solo-el-cuerpo-de-jurisprudencia", disparo: disparoDeLaApertura,
			mutante: conUnaLineaMas(t, repositorio.arbol, cuerpo), medidas: deJurisprudencia,
		},
		{
			nombre: "solo-el-cuerpo-de-jurisprudencia-sin-medida-de-main", disparo: disparoDeLaApertura,
			mutante: conUnaLineaMas(t, repositorio.arbol, cuerpo), sinMain: true,
			medidas: map[string]medidasEsperadas{
				"boe-legislacion": {motivo: motivoSinResultadoDeMain},
				"jurisprudencia":  {motivo: motivoConjunto, rutas: []string{cuerpo}},
				"legal-core":      {motivo: motivoSinResultadoDeMain},
			},
		},
		{nombre: "etiqueta-evals-sin-cambios", disparo: disparoDeLaEtiqueta, mutante: docs},
		{
			nombre: "solo-la-eval-sin-resultado", disparo: disparoDeLaEtiqueta, mutante: docs,
			otrasHuellas: map[string]string{evalSinResultado: otraHuella},
			medidas: map[string]medidasEsperadas{
				"legal-core": {motivo: motivoSinResultadoDeMain, ficheros: []string{"05-plazo-reposicion-leganes.yaml"}},
			},
		},
		{nombre: "apertura-sin-eleccion", disparo: disparoDeLaApertura, mutante: docs, sinMain: true, sinTrabajos: true},
		{
			nombre: "apertura-con-un-conjunto-incalculable", disparo: disparoDeLaApertura, mutante: docs, incalculable: true,
			medidas: todasMedidas(repositorio.skills, medidasEsperadas{motivo: motivoSinConjunto}),
			// Con todas elegidas, ningún trabajo busca nada en main.
			soloLasLecturas: []string{"ficheros de la propuesta 210"},
		},
		{
			nombre: "suite-completa-sin-ninguna-busqueda", disparo: disparoCompleto, mutante: docs,
			medidas:     todasMedidas(repositorio.skills, medidasEsperadas{motivo: motivoSuiteCompleta}),
			sinLecturas: true,
		},
		{
			nombre: "ficheros-de-la-propuesta-que-gh-no-da", disparo: disparoDeLaApertura,
			mutante: conUnaLineaMas(t, repositorio.arbol, cuerpo),
			fallos:  map[string]error{"ficheros de la propuesta 210": errDeLosFicheros},
			medidas: todasMedidas(repositorio.skills, medidasEsperadas{motivo: motivoPlataforma}),
			plataformaFallida: &errorDeLaPlataformaEnElOrigen{
				Orden: ordenDeLosFicherosDeLaPropuesta + ", con PROPUESTA=210",
				Error: "HTTP 502: Bad Gateway (https://api.github.com/repos/o/r/pulls/210/files)",
			},
			soloLasLecturas: []string{"ficheros de la propuesta 210"},
		},
	}

	for _, caso := range casos {
		t.Run(caso.nombre, func(t *testing.T) {
			t.Parallel()

			comprobarLaDecision(t, repositorio, caso)
		})
	}
}

// leerElRepositorioDeLaDecision lee del repositorio lo que comparten los casos
// de TestDecisionDelJob.
func leerElRepositorioDeLaDecision(t *testing.T) repositorioDeLaDecision {
	t.Helper()

	grafo, err := leerGrafoDePaquetes(t.Context(), raizDelRepositorio)
	require.NoError(t, err)

	arbol := os.DirFS(raizDelRepositorio)
	base := arbolConLosCambios(t, arbol, mutanteDeLaPropuesta{})

	comun, err := grafo.loComun(base.modulos)
	require.NoError(t, err)

	repositorio := repositorioDeLaDecision{arbol: arbol, grafo: grafo, base: base, comun: comun, evals: map[string][]Eval{}}

	for _, nombre := range nombresDeLasEvals(t, arbol) {
		skill, _, _ := strings.Cut(nombre, "/")
		if !slices.Contains(repositorio.skills, skill) {
			repositorio.skills = append(repositorio.skills, skill)
		}
	}

	for _, skill := range repositorio.skills {
		conjunto, err := LeerConjunto(filepath.Join(raizDelRepositorio, directorioDeLasEvals, skill))
		require.NoError(t, err)
		require.Emptyf(t, conjunto.MalFormados, "las evals de %s están bien formadas", skill)

		repositorio.evals[skill] = conjunto.Evals
	}

	return repositorio
}

// todasMedidas son las medidas esperadas de todas las evals de cada skill.
func todasMedidas(skills []string, medidas medidasEsperadas) map[string]medidasEsperadas {
	todas := map[string]medidasEsperadas{}
	for _, skill := range skills {
		todas[skill] = medidas
	}

	return todas
}

// calculo es el de la decisión sobre el repositorio con el mutante del caso:
// los conjuntos de la cabeza, lo común, los árboles de la cabeza y de la base,
// y una huella sintética por conjunto, la de su nombre; con el conjunto de
// evalSinResultado incalculable, si el caso lo pide, y entonces ninguna huella,
// como en el job.
func (r repositorioDeLaDecision) calculo(t *testing.T, caso casoDeLaDecision) calculoDeLaDecision {
	t.Helper()

	conjuntos, err := calcularDependencias(arbolConCambios{FS: r.arbol, cambiados: caso.mutante.cambiados}, r.grafo)
	require.NoError(t, err)

	calculo := calculoDeLaDecision{eleccion: calculoDeLaEleccion{
		conjuntos: conjuntos, comun: r.comun, cabeza: arbolConLosCambios(t, r.arbol, caso.mutante), base: r.base,
	}}

	if caso.incalculable {
		indice := slices.IndexFunc(conjuntos, func(conjunto dependenciasDeLaEval) bool {
			return conjunto.nombre() == evalSinResultado
		})
		require.GreaterOrEqual(t, indice, 0, "el repositorio tiene la eval %s", evalSinResultado)

		conjuntos[indice] = dependenciasDeLaEval{
			skill: conjuntos[indice].skill, eval: conjuntos[indice].eval,
			err: errors.New(evalSinResultado + `: el applet "inexistente" no tiene fichero en internal/app`),
		}

		return calculo
	}

	for _, conjunto := range conjuntos {
		calculo.huellas = append(calculo.huellas, huellaSintetica(conjunto.nombre()))
	}

	return calculo
}

// huellaSintetica es la huella de la eval con ese nombre en los casos de
// TestDecisionDelJob: 64 caracteres hexadecimales que solo dependen del nombre.
func huellaSintetica(nombre string) string {
	return huellaSHA256([]byte("huella sintética de " + nombre))
}

// plataforma es la plataforma en memoria del caso para el trabajo de la skill,
// o para el trabajo cambios si skill está vacía: los ficheros de la propuesta
// del mutante, sus fallos y, salvo sinMain, una medida de main con el informe
// de la skill, con la huella de cada eval de la base y, en la de cada eval de
// otrasHuellas, la suya.
func (r repositorioDeLaDecision) plataforma(t *testing.T, caso casoDeLaDecision, skill string) *plataformaEnMemoria {
	t.Helper()

	plataforma := &plataformaEnMemoria{
		propuestas: propuestasDeMain(),
		ficheros:   map[int][]cambioDeLaPropuesta{propuestaDelJob: caso.mutante.cambios},
		fallos:     caso.fallos,
	}

	if caso.sinMain || skill == "" {
		return plataforma
	}

	informe := Informe{
		Skill: skill, ModeloQueDecide: modeloQueDecideEnMain, ModelosInformativos: []string{modeloInformativoEnMain},
		Origen: &origenDeLasEvals{Modo: origenDeLaSuiteCompleta},
	}

	for _, eval := range r.evals[skill] {
		nombre := skill + "/" + strings.TrimSuffix(eval.Fichero, ".yaml")

		huella, otra := caso.otrasHuellas[nombre]
		if !otra {
			huella = huellaSintetica(nombre)
		}

		informe.Origen.Evals = append(informe.Origen.Evals, origenDeEval{
			Eval: eval.Fichero, Huella: huella, Medida: true, Motivo: motivoSuiteCompleta, Aportes: aportesDeLaEvalEnMain(),
		})
		informe.Origen.EvalsMedidas++
		informe.Tasas = append(informe.Tasas, seriesDeLaEval(eval.Fichero)...)
	}

	plataforma.ejecuciones = []ejecucionEnMemoria{medidaDeMainDelJob(t, r.skills, skill, informe)}

	return plataforma
}

// medidaDeMainDelJob es la ejecución del flujo que es medida de main para el
// trabajo de la skill: terminada en verde sobre la cabeza de
// propuestaDeMainDelJob, con el trabajo de cada skill terminado y con su
// informe escrito —el de la skill con el id que da trabajoDeLaSkillBuscada— y
// con el informe dado en el de la skill.
func medidaDeMainDelJob(t *testing.T, skills []string, skill string, informe Informe) ejecucionEnMemoria {
	t.Helper()

	contenido, err := json.MarshalIndent(informe, "", "  ")
	require.NoError(t, err)

	medida := ejecucionEnMemoria{
		ejecucion: ejecucionDelFlujo{
			ID: ejecucionDeMainDelJob, Commit: cabezaDeLaPropuesta(propuestaDeMainDelJob), Status: statusTerminado,
			Conclusion: conclusionDeExito, Creada: haceDesdeLaBusqueda(antesDeLaMedidaDeMain),
		},
		propuesta: propuestaDeMainDelJob,
		informe:   append(contenido, '\n'),
	}

	for posicion, otra := range skills {
		id := trabajoDeLaSkillBuscada(ejecucionDeMainDelJob)
		if otra != skill {
			id = ejecucionDeMainDelJob*100 + int64(posicion)
		}

		medida.trabajos = append(medida.trabajos, trabajoDeEvalsEnMemoria(id, otra, conclusionDeExito))
	}

	return medida
}

// procedenciaDeMainDelJob es la procedencia de una eval reutilizada de la
// medida de main de los casos de TestDecisionDelJob.
func procedenciaDeMainDelJob() *procedenciaDeLaMedida {
	return &procedenciaDeLaMedida{
		Ejecucion: ejecucionDeMainDelJob, Trabajo: trabajoDeLaSkillBuscada(ejecucionDeMainDelJob),
		Commit: cabezaDeLaPropuesta(propuestaDeMainDelJob), Propuesta: propuestaDeMainDelJob,
		Fecha: fechaDesdeLaBusqueda(antesDeLaMedidaDeMain),
	}
}

// modo es el de la ejecución del caso: la suite completa con su disparo, y con
// reutilización con la apertura y con la etiqueta.
func (c casoDeLaDecision) modo() modoDelOrigen {
	if c.disparo == disparoCompleto {
		return origenDeLaSuiteCompleta
	}

	return origenConReutilizacion
}

// esperado es el origen que espera el caso del trabajo de la skill, de las
// evals de su conjunto en la cabeza, en su orden: las de sus medidas, medidas
// con su motivo y sus rutas, y las demás, reutilizadas de la medida de main con
// sus aportes y sus series; cada una con su huella sintética, salvo sin
// conjunto.
func (c casoDeLaDecision) esperado(calculo calculoDeLaDecision, skill string) *origenDeLasEvals {
	origen := &origenDeLasEvals{
		Modo: c.modo(), SinConjunto: calculo.eleccion.sinConjunto(), ErrorDeLaPlataforma: c.plataformaFallida,
	}

	medidas, conMedidas := c.medidas[skill]

	for indice, conjunto := range calculo.eleccion.conjuntos {
		if conjunto.skill != skill {
			continue
		}

		fichero := conjunto.eval + ".yaml"
		eval := origenDeEval{Eval: fichero}

		if calculo.huellas != nil {
			eval.Huella = calculo.huellas[indice]
		}

		if conMedidas && (medidas.ficheros == nil || slices.Contains(medidas.ficheros, fichero)) {
			eval.Medida, eval.Motivo, eval.Rutas = true, medidas.motivo, medidas.rutas
			origen.EvalsMedidas++
		} else {
			eval.Procedencia, eval.Aportes, eval.Tasas = procedenciaDeMainDelJob(), aportesDeLaEvalEnMain(),
				seriesDeLaEval(fichero)
			origen.EvalsReutilizadas++
		}

		origen.Evals = append(origen.Evals, eval)
	}

	return origen
}

// sesionesEsperadas son las sesiones que el caso espera del trabajo de la
// skill: las del plan del job con solo las evals que espera medir, sin pasar
// por la decisión; ninguna si los trabajos no corren.
func (c casoDeLaDecision) sesionesEsperadas(evals []Eval, esperado *origenDeLasEvals) int {
	if c.sinTrabajos {
		return 0
	}

	var medidas []Eval

	for _, eval := range evals {
		if slices.ContainsFunc(esperado.Evals, func(decidida origenDeEval) bool {
			return decidida.Eval == eval.Fichero && decidida.Medida
		}) {
			medidas = append(medidas, eval)
		}
	}

	return len(planDelJob(medidas).Sesiones())
}

// planDelJob es el plan de sesiones del trabajo de una skill con esas evals:
// el modelo que decide y el informativo, las repeticiones del job y los dos
// modos.
func planDelJob(evals []Eval) PlanDeEvals {
	return PlanDeEvals{
		Evals: evals, ModeloQueDecide: modeloQueDecideEnMain, ModelosInformativos: []string{modeloInformativoEnMain},
		Repeticiones: repeticionesDelJob, Modos: []Modo{ModoOrden, ModoHerramienta},
	}
}

// comprobarLaDecision ejecuta el job del caso: con la apertura, la elección del
// trabajo cambios, que solo corre los trabajos si elige alguna; y la decisión
// de cada trabajo, con su plataforma en memoria. Compara, de cada skill, el
// origen, las líneas del registro, el fichero de la decisión, las lecturas de la
// plataforma con su plazo y las sesiones del plan.
func comprobarLaDecision(t *testing.T, repositorio repositorioDeLaDecision, caso casoDeLaDecision) {
	t.Helper()

	calculo := repositorio.calculo(t, caso)
	corren := repositorio.corren(t, caso, calculo)

	if caso.disparo == disparoDeLaApertura {
		assert.Equal(t, !caso.sinTrabajos, corren, "la apertura elige=si solo si elige alguna eval")
	}

	for _, skill := range repositorio.skills {
		esperado := caso.esperado(calculo, skill)
		sesiones := 0

		if corren {
			antes := time.Now()
			origen, plataforma := repositorio.decidir(t, caso, calculo, skill)
			despues := time.Now()

			comprobarElOrigenDeLaDecision(t, skill, esperado, origen)
			comprobarLasLecturas(t, skill, caso, plataforma.lecturas)
			comprobarElPlazoDeLasLecturas(t, skill, plataforma.plazos, antes, despues)

			sesiones = len(planDelJob(repositorio.evals[skill]).deLasMedidas(origen).Sesiones())
		}

		assert.Equalf(t, caso.sesionesEsperadas(repositorio.evals[skill], esperado), sesiones,
			"las sesiones del plan del trabajo de %s", skill)
	}
}

// corren dice si la ejecución del caso corre los trabajos de evals: con la
// apertura, solo si la elección del trabajo cambios, con su plataforma en
// memoria, elige alguna eval (FR-016); con la etiqueta y con la suite
// completa, siempre.
func (r repositorioDeLaDecision) corren(t *testing.T, caso casoDeLaDecision, calculo calculoDeLaDecision) bool {
	t.Helper()

	if caso.disparo != disparoDeLaApertura {
		return true
	}

	eleccion, _ := eleccionConLaPlataforma(t.Context(), r.plataforma(t, caso, ""), propuestaDelJob, calculo.eleccion)

	return eleccion.eligeAlguna()
}

// decidir es la decisión del trabajo de la skill en la ejecución del caso, con
// su plataforma en memoria, que también devuelve para ver lo que leyó.
func (r repositorioDeLaDecision) decidir(
	t *testing.T, caso casoDeLaDecision, calculo calculoDeLaDecision, skill string,
) (*origenDeLasEvals, *plataformaEnMemoria) {
	t.Helper()

	plataforma := r.plataforma(t, caso, skill)

	origen := decidirElTrabajo(t.Context(), trabajoADecidir{
		skill: skill, modo: caso.modo(), propuesta: propuestaDelJob, calculo: calculo, plataforma: plataforma,
		instante: instanteDeLaBusqueda,
	})

	return origen, plataforma
}

// comprobarElOrigenDeLaDecision compara el origen de la decisión del trabajo
// de la skill con el esperado, eval por eval; sus líneas del registro, la del
// error de la plataforma delante y una por eval como las de informe.md; y el
// fichero de la decisión, que es el origen sin sesiones abiertas y, leído y
// vuelto a escribir, da el mismo JSON.
func comprobarElOrigenDeLaDecision(t *testing.T, skill string, esperado, origen *origenDeLasEvals) {
	t.Helper()

	assert.Equalf(t, esperado, origen, "la decisión del trabajo de %s, eval por eval", skill)

	var lineas []string
	if fallida := esperado.ErrorDeLaPlataforma; fallida != nil {
		lineas = append(lineas, "error de la plataforma: "+fallida.Orden+": "+fallida.Error)
	}

	for _, eval := range esperado.Evals {
		lineas = append(lineas, lineaDelOrigen(eval, esperado.ErrorDeLaPlataforma != nil))
	}

	assert.Equalf(t, lineas, lineasDeLaDecision(origen), "el registro de la decisión del trabajo de %s", skill)

	directorio := t.TempDir()
	ruta, deNuevo := filepath.Join(directorio, "decision-de-evals.json"), filepath.Join(directorio, "de-nuevo.json")

	abiertas := 3
	conSesiones := *origen
	conSesiones.SesionesAbiertas = &abiertas
	require.NoError(t, escribirElFicheroDeLaDecision(ruta, &conSesiones))

	contenido, err := os.ReadFile(filepath.Clean(ruta))
	require.NoError(t, err)
	assert.NotContains(t, string(contenido), `"sesiones_abiertas"`, "el fichero de la decisión no lleva sesiones abiertas")

	leido, err := leerElFicheroDeLaDecision(ruta)
	require.NoError(t, err)
	assert.Lenf(t, leido.Evals, len(origen.Evals), "el fichero de la decisión de %s tiene todas sus evals", skill)
	require.NoError(t, escribirElFicheroDeLaDecision(deNuevo, leido))

	otraVez, err := os.ReadFile(filepath.Clean(deNuevo))
	require.NoError(t, err)
	assert.JSONEqf(t, string(contenido), string(otraVez), "el fichero de la decisión de %s se lee como se escribió", skill)
}

// comprobarLasLecturas compara las lecturas de la plataforma del trabajo de la
// skill con las que espera el caso: ninguna en la suite completa; solo las
// dadas, si las da; y si no, la de los ficheros de la propuesta delante y, con
// alguna eval que buscar, las de la búsqueda.
func comprobarLasLecturas(t *testing.T, skill string, caso casoDeLaDecision, lecturas []string) {
	t.Helper()

	switch {
	case caso.sinLecturas:
		assert.Emptyf(t, lecturas, "la suite completa de %s no lee nada de la plataforma", skill)
	case caso.soloLasLecturas != nil:
		assert.Equalf(t, caso.soloLasLecturas, lecturas, "las lecturas de la plataforma del trabajo de %s", skill)
	default:
		require.NotEmptyf(t, lecturas, "el trabajo de %s lee los ficheros de la propuesta", skill)
		assert.Equalf(t, "ficheros de la propuesta 210", lecturas[0],
			"el trabajo de %s lee primero los ficheros de la propuesta", skill)
	}
}

// comprobarElPlazoDeLasLecturas comprueba que cada lectura de la plataforma
// del trabajo de la skill, entre antes y después de su decisión, lleva el plazo
// de lo que la decisión lee, plazoDeLaPlataforma desde que la decisión empieza
// (FR-013; contracts/job-de-evals.md §2 de H30): sin él, una orden de gh que no
// responde la corta el tiempo del paso y el trabajo no deja decisión. El
// contexto del test no lleva plazo.
func comprobarElPlazoDeLasLecturas(t *testing.T, skill string, plazos []time.Time, antes, despues time.Time) {
	t.Helper()

	for posicion, plazo := range plazos {
		assert.Falsef(t, plazo.IsZero(), "la lectura %d de la plataforma del trabajo de %s lleva plazo", posicion, skill)
		assert.Falsef(t, plazo.Before(antes.Add(plazoDeLaPlataforma)) || plazo.After(despues.Add(plazoDeLaPlataforma)),
			"la lectura %d de la plataforma del trabajo de %s tiene de plazo %s, y es el de %s desde que empieza "+
				"la decisión, entre %s y %s", posicion, skill, plazo, plazoDeLaPlataforma,
			antes.Add(plazoDeLaPlataforma), despues.Add(plazoDeLaPlataforma))
	}
}

// TestFicheroDeLaDecisionQueNoSeLee fija que el fichero de la decisión que no
// existe o no es JSON es un error que lo nombra.
func TestFicheroDeLaDecisionQueNoSeLee(t *testing.T) {
	t.Parallel()

	directorio := t.TempDir()
	noEsJSON := filepath.Join(directorio, "no-es-json.json")
	require.NoError(t, os.WriteFile(noEsJSON, []byte("{\n  \"modo\": \n"), 0o600))

	for _, ruta := range []string{filepath.Join(directorio, "no-existe.json"), noEsJSON} {
		_, err := leerElFicheroDeLaDecision(ruta)
		require.Error(t, err)
		assert.Contains(t, err.Error(), ruta, "el error nombra el fichero de la decisión")
	}
}

// TestModoDeLaDecision fija los dos modos de la decisión de un trabajo, los
// del env de la tanda (contracts/job-de-evals.md §1 de H30), y el error que
// nombra cualquier otro.
func TestModoDeLaDecision(t *testing.T) {
	t.Parallel()

	for valor, esperado := range map[string]modoDelOrigen{
		"completa": origenDeLaSuiteCompleta, "reutilizacion": origenConReutilizacion,
	} {
		modo, err := modoDeLaDecision(valor)
		require.NoError(t, err)
		assert.Equal(t, esperado, modo)
	}

	for _, valor := range []string{"", "Completa", "reutilización", "todas"} {
		_, err := modoDeLaDecision(valor)
		require.Errorf(t, err, "%q no es un modo de la decisión", valor)
		assert.Contains(t, err.Error(), strconv.Quote(valor), "el error nombra el modo")
	}
}

// TestSalidaDeLaApertura fija lo que la apertura añade a su GITHUB_OUTPUT
// (contracts/job-de-evals.md §2 de H30): una línea, elige=si o elige=no, detrás
// de lo que ya tenía.
func TestSalidaDeLaApertura(t *testing.T) {
	t.Parallel()

	for elige, linea := range map[bool]string{true: "elige=si\n", false: "elige=no\n"} {
		salida := filepath.Join(t.TempDir(), "github-output")
		require.NoError(t, os.WriteFile(salida, []byte("otra=1\n"), 0o600))

		require.NoError(t, escribirLaApertura(salida, elige))

		contenido, err := os.ReadFile(filepath.Clean(salida))
		require.NoError(t, err)
		assert.Equal(t, "otra=1\n"+linea, string(contenido))
	}

	err := escribirLaApertura(filepath.Join(t.TempDir(), "no-existe", "github-output"), true)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "github-output", "el error nombra la salida")
}

// TestCalculoDeLaDecisionDelCheckout lee del checkout del repositorio, con
// HEAD como base, lo que la decisión necesita: el árbol de la base, el de la
// cabeza y una huella de 64 caracteres hexadecimales por conjunto, como
// TestHuellaDeLasEvals; sin base, ningún árbol de la base, el de la suite
// completa; y con un fijado vacío, el error que lo nombra, antes de leer nada.
func TestCalculoDeLaDecisionDelCheckout(t *testing.T) {
	t.Parallel()

	calculo, err := leerElCalculoDeLaDecision(t.Context(), raizDelRepositorio, "HEAD", fijadosDelTest)
	require.NoError(t, err)

	require.Empty(t, calculo.eleccion.sinConjunto(), "cada conjunto del repositorio se calcula")
	require.Len(t, calculo.huellas, len(calculo.eleccion.conjuntos), "una huella por conjunto")

	for _, huella := range calculo.huellas {
		assert.Regexp(t, patronDeUnaHuella, huella)
	}

	assert.NotEmpty(t, calculo.eleccion.base.skills, "con base, el árbol de la base")
	assert.NotEmpty(t, calculo.eleccion.base.versionDeGo, "con base, el go.mod de la base")

	sinBase, err := leerElCalculoDeLaDecision(t.Context(), raizDelRepositorio, "", fijadosDelTest)
	require.NoError(t, err)
	assert.Empty(t, sinBase.eleccion.base.skills, "sin base, ningún árbol de la base")
	assert.Equal(t, calculo.huellas, sinBase.huellas, "la base no cambia ninguna huella")

	sinModelo := fijadosDelTest
	sinModelo.modeloQueDecide = ""

	_, err = leerElCalculoDeLaDecision(t.Context(), filepath.Join(t.TempDir(), "no-existe"), "HEAD", sinModelo)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "el modelo que decide", "el error nombra el fijado vacío")
}
