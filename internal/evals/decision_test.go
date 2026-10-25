package evals

import (
	"encoding/json"
	"errors"
	"io/fs"
	"maps"
	"os"
	"path/filepath"
	"slices"
	"strconv"
	"strings"
	"testing"
	"testing/fstest"
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

// casoDeLaDecision es un caso de TestDecisionDelJob: lo que lo lanza, si la
// ejecución no tiene propuesta (el despacho manual), la propuesta con sus
// ficheros, si hay medida de main y con qué huellas distintas de las de la
// cabeza, los fallos de la plataforma, si el conjunto de evalSinResultado no se
// puede calcular, otro juez que el de fijadosDelTest, y lo que espera: si los
// trabajos corren, lo que mide cada skill —las que no están, todo reutilizado de
// la medida de main—, las skills cuya descripción se juzga y aquellas cuyo
// juicio se reutiliza de la medida de main —las demás, sin bloque
// descripcion—, el error de la plataforma y las lecturas de cada trabajo.
type casoDeLaDecision struct {
	nombre       string
	disparo      disparoDelJob
	sinPropuesta bool
	mutante      mutanteDeLaPropuesta
	sinMain      bool
	otrasHuellas map[string]string
	fallos       map[string]error
	incalculable bool
	otroJuez     func(fijados *fijadosDeLaHuella)

	sinTrabajos       bool
	medidas           map[string]medidasEsperadas
	juzgadas          []string
	reutilizadas      []string
	plataformaFallida *errorDeLaPlataformaEnElOrigen
	sinLecturas       bool
	soloLasLecturas   []string
}

// repositorioDeLaDecision es lo que los casos de TestDecisionDelJob leen una
// vez del repositorio: su árbol, su grafo de paquetes, su árbol como base de la
// propuesta, lo común, sus skills con evals, en orden, las evals de cada una y
// el juez de las descripciones.
type repositorioDeLaDecision struct {
	arbol  fs.FS
	grafo  grafoDePaquetes
	base   arbolDelRepositorio
	comun  loComun
	skills []string
	evals  map[string][]Eval
	juez   *Juez
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
// más. Y el bloque descripcion de cada trabajo (FR-030 de H31; Clarificación Q2
// de H31; research D3 y D4 de H31; contracts/juez-de-las-descripciones.md §2 de
// H31): con una descripción cambiada, en los dos modos y con lo común cambiado a
// la vez, lo lleva el trabajo de cada skill, juzgada o, con reutilización, con
// el juicio de la medida de main con su huella; sin propuesta, con la etiqueta
// evals o la suite completa sin descripciones cambiadas y con un cambio solo de
// la rúbrica, el esquema, los casos, la medida, el modelo o la versión del
// juez, ninguno lo lleva.
func TestDecisionDelJob(t *testing.T) {
	t.Parallel()

	repositorio := leerElRepositorioDeLaDecision(t)

	casos := slices.Concat(casosDeLaDecisionDeH30(t, repositorio), casosDeLaDescripcionDelJob(t, repositorio),
		casosSinDescripcionQueJuzgar(t, repositorio))

	for _, caso := range casos {
		t.Run(caso.nombre, func(t *testing.T) {
			t.Parallel()

			comprobarLaDecision(t, repositorio, caso)
		})
	}

	t.Run("juez-de-las-descripciones-que-no-se-lee", func(t *testing.T) {
		t.Parallel()

		calculo := repositorio.calculo(t, casoDeLaDecision{
			disparo: disparoCompleto, sinPropuesta: true, mutante: conOtraDescripcionDeLegalCore(t, repositorio.arbol),
		})

		err := calculo.juzgarLasDescripciones(true, t.TempDir(), fijadosDelTest)
		require.Error(t, err, "con una descripción cambiada, el juez de las descripciones se lee")
		assert.Contains(t, err.Error(), "evidencias/adr-0046/rubrica.md", "el error nombra el fichero del juez")
		assert.Nil(t, calculo.huellasDeLasDescripciones, "sin juez, ninguna huella")
	})
}

// casosDeLaDecisionDeH30 son los casos de lo que mide y lo que reutiliza cada
// trabajo, los de H30, en los que ninguna descripción cambia.
func casosDeLaDecisionDeH30(t *testing.T, repositorio repositorioDeLaDecision) []casoDeLaDecision {
	t.Helper()

	docs := conUnaLineaMas(t, repositorio.arbol, ficheroDelRepositorio(t, repositorio.arbol, "docs", esMarkdown))
	cuerpo := "skills/jurisprudencia/SKILL.md"
	deJurisprudencia := map[string]medidasEsperadas{
		"jurisprudencia": {motivo: motivoConjunto, rutas: []string{cuerpo}},
	}

	return []casoDeLaDecision{
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
			mutante:           conUnaLineaMas(t, repositorio.arbol, cuerpo),
			fallos:            falloDeLosFicherosDeLaPropuesta(),
			medidas:           todasMedidas(repositorio.skills, medidasEsperadas{motivo: motivoPlataforma}),
			plataformaFallida: errorDeLosFicherosDeLaPropuesta(),
			soloLasLecturas:   []string{"ficheros de la propuesta 210"},
		},
	}
}

// casosDeLaDescripcionDelJob son los casos de una propuesta que cambia la
// descripción de legal-core, en la que el trabajo de cada skill lleva el bloque
// descripcion: con la apertura, sola y con lo común cambiado a la vez, juzga la
// de legal-core, cuya huella no tiene la medida de main, y reutiliza el juicio
// de las demás de la medida de main, que se lee aunque ninguna eval se busque;
// en la suite completa sobre la propuesta, las juzga todas sin leer nada de la
// plataforma; y si gh no da los ficheros de la propuesta, la búsqueda no
// empieza y las juzga todas.
func casosDeLaDescripcionDelJob(t *testing.T, repositorio repositorioDeLaDecision) []casoDeLaDecision {
	t.Helper()

	descripcion := conOtraDescripcionDeLegalCore(t, repositorio.arbol)
	skillMD := "skills/legal-core/SKILL.md"
	cli := ficheroDelRepositorio(t, repositorio.arbol, "internal/cli", esCodigoSinTests)
	lasDemas := slices.DeleteFunc(slices.Clone(repositorio.skills), func(skill string) bool { return skill == "legal-core" })
	conLaMedidaDeMain := slices.Concat([]string{"ficheros de la propuesta 210"}, lecturasDeLasMedidas(ejecucionDeMainDelJob))

	return []casoDeLaDecision{
		{
			nombre: "descripcion-de-legal-core-con-la-apertura", disparo: disparoDeLaApertura, mutante: descripcion,
			medidas:  todasMedidas(repositorio.skills, medidasEsperadas{motivo: motivoDescripcion, rutas: []string{skillMD}}),
			juzgadas: []string{"legal-core"}, reutilizadas: lasDemas, soloLasLecturas: conLaMedidaDeMain,
		},
		{
			nombre: "descripcion-de-legal-core-y-lo-comun-con-la-apertura", disparo: disparoDeLaApertura,
			mutante:  juntos(descripcion, conUnaLineaMas(t, repositorio.arbol, cli)),
			medidas:  todasMedidas(repositorio.skills, medidasEsperadas{motivo: motivoComun, rutas: []string{cli, skillMD}}),
			juzgadas: []string{"legal-core"}, reutilizadas: lasDemas, soloLasLecturas: conLaMedidaDeMain,
		},
		{
			nombre: "descripcion-de-legal-core-sin-medida-de-main", disparo: disparoDeLaEtiqueta, mutante: descripcion,
			sinMain:  true,
			medidas:  todasMedidas(repositorio.skills, medidasEsperadas{motivo: motivoDescripcion, rutas: []string{skillMD}}),
			juzgadas: repositorio.skills,
			soloLasLecturas: []string{
				"ficheros de la propuesta 210", lecturaDeLasEjecuciones, lecturaDeLasPropuestas,
			},
		},
		{
			nombre: "descripcion-de-legal-core-en-la-suite-completa", disparo: disparoCompleto, mutante: descripcion,
			medidas:  todasMedidas(repositorio.skills, medidasEsperadas{motivo: motivoSuiteCompleta}),
			juzgadas: repositorio.skills, sinLecturas: true,
		},
		{
			nombre: "descripcion-de-legal-core-sin-los-ficheros-de-la-propuesta", disparo: disparoDeLaApertura,
			mutante: descripcion, fallos: falloDeLosFicherosDeLaPropuesta(),
			medidas:           todasMedidas(repositorio.skills, medidasEsperadas{motivo: motivoPlataforma}),
			juzgadas:          repositorio.skills,
			plataformaFallida: errorDeLosFicherosDeLaPropuesta(),
			soloLasLecturas:   []string{"ficheros de la propuesta 210"},
		},
	}
}

// casosSinDescripcionQueJuzgar son los casos en que ningún trabajo lleva el
// bloque descripcion: la descripción de legal-core cambiada en la suite
// completa sin propuesta, el despacho manual; y, con la etiqueta evals, un
// cambio solo de uno de los cuatro ficheros del juez de las descripciones, o de
// su modelo o su versión de Claude Code, que no cambian ninguna descripción.
func casosSinDescripcionQueJuzgar(t *testing.T, repositorio repositorioDeLaDecision) []casoDeLaDecision {
	t.Helper()

	docs := conUnaLineaMas(t, repositorio.arbol, ficheroDelRepositorio(t, repositorio.arbol, "docs", esMarkdown))
	casos := []casoDeLaDecision{
		{
			nombre: "descripcion-de-legal-core-sin-propuesta", disparo: disparoCompleto, sinPropuesta: true,
			mutante:     conOtraDescripcionDeLegalCore(t, repositorio.arbol),
			medidas:     todasMedidas(repositorio.skills, medidasEsperadas{motivo: motivoSuiteCompleta}),
			sinLecturas: true,
		},
		{
			nombre: "etiqueta-con-otro-modelo-del-juez", disparo: disparoDeLaEtiqueta, mutante: docs,
			otroJuez: func(fijados *fijadosDeLaHuella) { fijados.modeloDelJuez = "claude-opus-5-6" },
		},
		{
			nombre: "etiqueta-con-otra-version-del-juez", disparo: disparoDeLaEtiqueta, mutante: docs,
			otroJuez: func(fijados *fijadosDeLaHuella) { fijados.claudeCodeDelJuez = "2.1.290" },
		},
	}

	for _, usado := range ficherosUsadosPorElJuez {
		casos = append(casos, casoDeLaDecision{
			nombre: "etiqueta-con-solo-" + usado.campo + "-del-juez", disparo: disparoDeLaEtiqueta,
			mutante: conUnaLineaMas(t, repositorio.arbol, enLaEvidencia(usado.fichero)),
		})
	}

	return casos
}

// falloDeLosFicherosDeLaPropuesta es el fallo de la lectura de los ficheros de
// la propuesta del job en la plataforma en memoria: gh termina con 1 y un 502.
func falloDeLosFicherosDeLaPropuesta() map[string]error {
	return map[string]error{"ficheros de la propuesta 210": &errorDeUnaOrden{
		orden: ordenDeLosFicherosDeLaPropuesta, variables: []string{"PROPUESTA=" + strconv.Itoa(propuestaDelJob)},
		causa: errors.New("exit status 1"), salidaDeError: "HTTP 502: Bad Gateway (https://api.github.com/repos/o/r/pulls/210/files)\n",
	}}
}

// errorDeLosFicherosDeLaPropuesta es el error de la plataforma en el origen
// por falloDeLosFicherosDeLaPropuesta.
func errorDeLosFicherosDeLaPropuesta() *errorDeLaPlataformaEnElOrigen {
	return &errorDeLaPlataformaEnElOrigen{
		Orden: ordenDeLosFicherosDeLaPropuesta + ", con PROPUESTA=210",
		Error: "HTTP 502: Bad Gateway (https://api.github.com/repos/o/r/pulls/210/files)",
	}
}

// juntos es la propuesta que cambia lo que cambian todos los mutantes dados,
// que no cambian el mismo fichero.
func juntos(mutantes ...mutanteDeLaPropuesta) mutanteDeLaPropuesta {
	junto := mutanteDeLaPropuesta{cambiados: fstest.MapFS{}}

	for _, mutante := range mutantes {
		maps.Copy(junto.cambiados, mutante.cambiados)
		junto.retirados = append(junto.retirados, mutante.retirados...)
		junto.cambios = append(junto.cambios, mutante.cambios...)
	}

	return junto
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

	juez, err := juezDeLasDescripciones(raizDelRepositorio)
	require.NoError(t, err)

	repositorio := repositorioDeLaDecision{
		arbol: arbol, grafo: grafo, base: base, comun: comun, evals: map[string][]Eval{}, juez: juez,
	}

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
// como en el job. Y la huella del juicio de cada descripción si la ejecución
// las juzga (juzgarLasDescripciones), con el juez del repositorio y los
// fijados del caso, como en el job.
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
	} else {
		for _, conjunto := range conjuntos {
			calculo.huellas = append(calculo.huellas, huellaSintetica(conjunto.nombre()))
		}
	}

	require.NoError(t, calculo.juzgarLasDescripciones(!caso.sinPropuesta, raizDelRepositorio, caso.fijados()))

	return calculo
}

// fijados son los del caso: los de fijadosDelTest, con el otro juez si lo
// lleva.
func (c casoDeLaDecision) fijados() fijadosDeLaHuella {
	fijados := fijadosDelTest
	if c.otroJuez != nil {
		c.otroJuez(&fijados)
	}

	return fijados
}

// huellaDelJuicioDeLaSkill es la huella del juicio de la descripción de la
// skill en el árbol dado, la de la cabeza o la de la base, con el juez del
// repositorio y los fijados dados.
func (r repositorioDeLaDecision) huellaDelJuicioDeLaSkill(
	arbol arbolDelRepositorio, skill string, fijados fijadosDeLaHuella,
) string {
	return huellaDelJuicio(skill, arbol.descripciones[skill], r.juez, fijados.modeloDelJuez, fijados.claudeCodeDelJuez)
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
// otrasHuellas, la suya, y con el juicio sin marcar de la descripción de la
// skill en la base, con fijadosDelTest.
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

	plataforma.ejecuciones = []ejecucionEnMemoria{conLaDescripcion(t, medidaDeMainDelJob(t, r.skills, skill, informe),
		bloqueJuzgadoEnMain(skill, r.huellaDelJuicioDeLaSkill(r.base, skill, fijadosDelTest), false))}

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
// conjunto. Si la skill es de las juzgadas, el bloque descripcion con la huella
// dada, juzgada; si es de las reutilizadas, con el juicio sin marcar de la
// medida de main y su procedencia; si no, sin bloque.
func (c casoDeLaDecision) esperado(calculo calculoDeLaDecision, skill, huella string) *origenDeLasEvals {
	origen := &origenDeLasEvals{
		Modo: c.modo(), SinConjunto: calculo.eleccion.sinConjunto(), ErrorDeLaPlataforma: c.plataformaFallida,
	}

	switch {
	case slices.Contains(c.juzgadas, skill):
		origen.Descripcion = &descripcionDeLaDecision{Skill: skill, Huella: huella, Juzgada: true}
	case slices.Contains(c.reutilizadas, skill):
		origen.Descripcion = &descripcionDeLaDecision{
			Skill: skill, Huella: huella, Procedencia: procedenciaDeMainDelJob(),
			Juicio: &juicioReutilizado{Motivo: motivoSinMarcarEnMain},
		}
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
		esperado := caso.esperado(calculo, skill,
			repositorio.huellaDelJuicioDeLaSkill(calculo.eleccion.cabeza, skill, caso.fijados()))
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
	comprobarLaDescripcionDelFichero(t, skill, esperado.Descripcion, contenido)

	leido, err := leerElFicheroDeLaDecision(ruta)
	require.NoError(t, err)
	assert.Lenf(t, leido.Evals, len(origen.Evals), "el fichero de la decisión de %s tiene todas sus evals", skill)
	require.NoError(t, escribirElFicheroDeLaDecision(deNuevo, leido))

	otraVez, err := os.ReadFile(filepath.Clean(deNuevo))
	require.NoError(t, err)
	assert.JSONEqf(t, string(contenido), string(otraVez), "el fichero de la decisión de %s se lee como se escribió", skill)
}

// comprobarLaDescripcionDelFichero compara el bloque descripcion del fichero
// de la decisión del trabajo de la skill con el esperado, clave a clave, con
// las de contracts/juez-de-las-descripciones.md §2 de H31: skill, huella y
// juzgada y, si se reutiliza, procedencia, con las claves de la de H30, y
// juicio, con marcada, frase y motivo. Sin el esperado, el fichero no lleva el
// bloque.
func comprobarLaDescripcionDelFichero(t *testing.T, skill string, esperada *descripcionDeLaDecision, contenido []byte) {
	t.Helper()

	var fichero map[string]any
	require.NoError(t, json.Unmarshal(contenido, &fichero))

	bloque, conBloque := fichero["descripcion"]
	if esperada == nil {
		assert.Falsef(t, conBloque, "el fichero de la decisión de %s no lleva el bloque descripcion", skill)

		return
	}

	contrato := map[string]any{"skill": esperada.Skill, "huella": esperada.Huella, "juzgada": esperada.Juzgada}

	if procedencia := esperada.Procedencia; procedencia != nil {
		contrato["procedencia"] = map[string]any{
			"ejecucion": float64(procedencia.Ejecucion), "trabajo": float64(procedencia.Trabajo),
			"commit": procedencia.Commit, "propuesta": float64(procedencia.Propuesta), "fecha": procedencia.Fecha,
		}
	}

	if juicio := esperada.Juicio; juicio != nil {
		contrato["juicio"] = map[string]any{"marcada": juicio.Marcada, "frase": juicio.Frase, "motivo": juicio.Motivo}
	}

	assert.Equalf(t, contrato, bloque, "el bloque descripcion del fichero de la decisión de %s", skill)
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
	assert.Nil(t, sinBase.huellasDeLasDescripciones, "sin base, que es sin propuesta, ninguna descripción que juzgar")

	sinModelo := fijadosDelTest
	sinModelo.modeloQueDecide = ""

	_, err = leerElCalculoDeLaDecision(t.Context(), filepath.Join(t.TempDir(), "no-existe"), "HEAD", sinModelo)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "el modelo que decide", "el error nombra el fijado vacío")
}
