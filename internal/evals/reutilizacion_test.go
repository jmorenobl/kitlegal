package evals

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"slices"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// Lo que comparten las medidas de main sintéticas de TestBusquedaDeMedidasDeMain:
// la skill cuyo trabajo busca, la otra skill del flujo, el modelo que decide, el
// informativo y la vida de un resultado, la del registro de la plataforma.
const (
	skillDeLaBusqueda       = "normativa"
	otraSkillDeLaBusqueda   = "plazos-y-territorio"
	modeloQueDecideEnMain   = "claude-sonnet-5-5"
	modeloInformativoEnMain = "claude-haiku-4-5"
	noventaDiasDeRegistro   = 90 * 24 * time.Hour
)

// Las lecturas de la plataforma en memoria que no llevan ningún valor, como las
// anota.
const (
	lecturaDeLasEjecuciones = "ejecuciones del flujo"
	lecturaDeLasPropuestas  = "propuestas fusionadas"
)

// instanteDeLaBusqueda es el instante fijo de la decisión con el que la
// búsqueda compara la fecha de cada medida y de cada procedencia.
var instanteDeLaBusqueda = time.Date(2026, 10, 22, 12, 0, 0, 0, time.UTC)

// Las dos evals no elegidas que buscan su resultado en casi todos los casos,
// cada una con su huella, y otra huella, la de una versión anterior de la
// primera.
var (
	evalBuscadaA = evalABuscar{fichero: "01-articulo.yaml", huella: strings.Repeat("a", 64)}
	evalBuscadaB = evalABuscar{fichero: "02-indice.yaml", huella: strings.Repeat("b", 64)}
	otraHuella   = strings.Repeat("c", 64)
)

// errPlazoEnCurso es el fallo con el que una lectura de la plataforma en
// memoria sigue en curso hasta que se agota el plazo del contexto.
var errPlazoEnCurso = errors.New("la lectura sigue en curso hasta que se agota el plazo")

// ejecucionEnMemoria es una ejecución del flujo evals de la plataforma en
// memoria: la ejecución, la propuesta de cuya cabeza es, sus trabajos y el
// informe que imprimió el de la skill buscada.
type ejecucionEnMemoria struct {
	ejecucion ejecucionDelFlujo
	propuesta int
	trabajos  []trabajoDeLaEjecucion
	informe   []byte
}

// plataformaEnMemoria son las lecturas de la plataforma con respuestas
// sintéticas con la forma de las de gh: las propuestas fusionadas, las
// ejecuciones con sus trabajos y sus informes, los ficheros de cada propuesta y
// el fallo de cada lectura que falla. Anota cada lectura, en orden, y el plazo
// de su contexto, el instante cero si no lleva ninguno.
type plataformaEnMemoria struct {
	propuestas  []propuestaFusionada
	ejecuciones []ejecucionEnMemoria
	ficheros    map[int][]cambioDeLaPropuesta
	fallos      map[string]error
	lecturas    []string
	plazos      []time.Time
}

// La plataforma en memoria implementa las lecturas de la plataforma.
var _ lecturasDeLaPlataforma = (*plataformaEnMemoria)(nil)

// leer anota la lectura y el plazo de su contexto y da su fallo, si lo tiene:
// con errPlazoEnCurso, el del contexto en cuanto se agota.
func (p *plataformaEnMemoria) leer(ctx context.Context, lectura string) error {
	plazo, _ := ctx.Deadline()
	p.lecturas = append(p.lecturas, lectura)
	p.plazos = append(p.plazos, plazo)

	fallo := p.fallos[lectura]
	if errors.Is(fallo, errPlazoEnCurso) {
		<-ctx.Done()

		return ctx.Err()
	}

	return fallo
}

func (p *plataformaEnMemoria) ficherosDeLaPropuesta(ctx context.Context, propuesta int) ([]cambioDeLaPropuesta, error) {
	if err := p.leer(ctx, "ficheros de la propuesta "+strconv.Itoa(propuesta)); err != nil {
		return nil, err
	}

	return p.ficheros[propuesta], nil
}

func (p *plataformaEnMemoria) propuestasFusionadas(ctx context.Context) ([]propuestaFusionada, error) {
	if err := p.leer(ctx, lecturaDeLasPropuestas); err != nil {
		return nil, err
	}

	return p.propuestas, nil
}

func (p *plataformaEnMemoria) ejecucionesDelFlujo(ctx context.Context) ([]ejecucionDelFlujo, error) {
	if err := p.leer(ctx, lecturaDeLasEjecuciones); err != nil {
		return nil, err
	}

	ejecuciones := make([]ejecucionDelFlujo, 0, len(p.ejecuciones))
	for _, enMemoria := range p.ejecuciones {
		ejecuciones = append(ejecuciones, enMemoria.ejecucion)
	}

	return ejecuciones, nil
}

func (p *plataformaEnMemoria) ejecucionesDelCommit(ctx context.Context, commit string) ([]ejecucionDelFlujo, error) {
	if err := p.leer(ctx, "ejecuciones del commit "+commit); err != nil {
		return nil, err
	}

	var ejecuciones []ejecucionDelFlujo

	for _, enMemoria := range p.ejecuciones {
		if enMemoria.ejecucion.Commit == commit {
			ejecuciones = append(ejecuciones, enMemoria.ejecucion)
		}
	}

	return ejecuciones, nil
}

func (p *plataformaEnMemoria) trabajos(ctx context.Context, ejecucion int64) ([]trabajoDeLaEjecucion, error) {
	if err := p.leer(ctx, lecturaDeLosTrabajos(ejecucion)); err != nil {
		return nil, err
	}

	for _, enMemoria := range p.ejecuciones {
		if enMemoria.ejecucion.ID == ejecucion {
			return enMemoria.trabajos, nil
		}
	}

	return nil, fmt.Errorf("la plataforma en memoria no tiene la ejecución %d", ejecucion)
}

func (p *plataformaEnMemoria) informeDelTrabajo(ctx context.Context, ejecucion, trabajo int64) ([]byte, error) {
	if err := p.leer(ctx, lecturaDelInforme(ejecucion, trabajo)); err != nil {
		return nil, err
	}

	for _, enMemoria := range p.ejecuciones {
		if enMemoria.ejecucion.ID == ejecucion && trabajo == trabajoDeLaSkillBuscada(ejecucion) {
			return enMemoria.informe, nil
		}
	}

	return nil, fmt.Errorf("la plataforma en memoria no tiene el trabajo %d de la ejecución %d", trabajo, ejecucion)
}

// lecturaDeLosTrabajos es como anota la plataforma en memoria la lectura de los
// trabajos de una ejecución.
func lecturaDeLosTrabajos(ejecucion int64) string {
	return fmt.Sprintf("trabajos de %d", ejecucion)
}

// lecturaDelInforme es como anota la plataforma en memoria la lectura del
// informe de un trabajo de una ejecución.
func lecturaDelInforme(ejecucion, trabajo int64) string {
	return fmt.Sprintf("informe del trabajo %d de %d", trabajo, ejecucion)
}

// trabajoDeLaSkillBuscada y trabajoDeLaOtraSkill son los ids de los trabajos
// evals (<skill>) de las dos skills en una ejecución sintética.
func trabajoDeLaSkillBuscada(ejecucion int64) int64 { return ejecucion*10 + 1 }

func trabajoDeLaOtraSkill(ejecucion int64) int64 { return ejecucion*10 + 2 }

// cabezaDeLaPropuesta es el commit con el que se fusionó una propuesta
// sintética.
func cabezaDeLaPropuesta(propuesta int) string {
	return "cabeza-de-la-" + strconv.Itoa(propuesta)
}

// propuestasDeMain son las propuestas fusionadas de la plataforma en memoria,
// de la 199 a la 180, cada una con su cabeza.
func propuestasDeMain() []propuestaFusionada {
	var propuestas []propuestaFusionada

	for numero := 199; numero >= 180; numero-- {
		propuestas = append(propuestas, propuestaFusionada{
			numero: numero, cabeza: cabezaDeLaPropuesta(numero), fusion: instanteDeLaBusqueda.Add(-time.Duration(200-numero) * time.Hour),
		})
	}

	return propuestas
}

// haceDesdeLaBusqueda es el instante que queda el tiempo dado antes del de la
// búsqueda, y fechaDesdeLaBusqueda, el mismo como lo escribe una procedencia.
func haceDesdeLaBusqueda(antes time.Duration) time.Time {
	return instanteDeLaBusqueda.Add(-antes)
}

func fechaDesdeLaBusqueda(antes time.Duration) string {
	return haceDesdeLaBusqueda(antes).Format(time.RFC3339)
}

// seriesDeLaEval son las series de una eval en un informe sintético: la del
// modelo que decide y la del informativo.
func seriesDeLaEval(fichero string) []TasaDelInforme {
	return []TasaDelInforme{
		{
			Eval: fichero, Modelo: modeloQueDecideEnMain, Modo: ModoOrden, Planificada: true, Decide: true,
			Formas: []string{}, Sesiones: 3, Pasan: 3, Pasa: true,
		},
		{
			Eval: fichero, Modelo: modeloInformativoEnMain, Modo: ModoOrden, Planificada: true,
			Formas: []string{}, Sesiones: 3, Pasan: 2, Pasa: false,
		},
	}
}

// aportesDeLaEvalEnMain son los aportes de una eval en un informe sintético.
func aportesDeLaEvalEnMain() []aporteDeLaEval {
	return []aporteDeLaEval{
		{Umbral: "afirma_lo_no_leido:" + modeloQueDecideEnMain, Medida: 1, Total: 6},
		{Umbral: "sin_activar:" + modeloQueDecideEnMain + ":orden", Medida: 0, Total: 3},
	}
}

// medidaAqui es la entrada del origen de un informe de main de una eval que se
// midió en él, con la huella dada.
func medidaAqui(eval evalABuscar, huella string) origenDeEval {
	return origenDeEval{
		Eval: eval.fichero, Huella: huella, Medida: true, Motivo: motivoConjunto,
		Rutas: []string{"evals/" + skillDeLaBusqueda + "/" + eval.fichero}, Aportes: aportesDeLaEvalEnMain(),
	}
}

// reutilizadaEnMain es la entrada del origen de un informe de main de una eval
// que reutilizó el resultado de la ejecución dada, de la propuesta dada,
// creada el tiempo dado antes de la búsqueda.
func reutilizadaEnMain(eval evalABuscar, ejecucion int64, propuesta int, antes time.Duration) origenDeEval {
	return origenDeEval{
		Eval: eval.fichero, Huella: eval.huella,
		Procedencia: &procedenciaDeLaMedida{
			Ejecucion: ejecucion, Trabajo: trabajoDeLaSkillBuscada(ejecucion), Commit: cabezaDeLaPropuesta(propuesta),
			Propuesta: propuesta, Fecha: fechaDesdeLaBusqueda(antes),
		},
		Aportes: aportesDeLaEvalEnMain(), Tasas: seriesDeLaEval(eval.fichero),
	}
}

// informeDeMain es el informe de la skill buscada en una medida de main con las
// entradas dadas en su origen: de cada eval medida en él, sus series y una
// sesión con cada modelo.
func informeDeMain(entradas ...origenDeEval) Informe {
	informe := Informe{
		Skill: skillDeLaBusqueda, ModeloQueDecide: modeloQueDecideEnMain, ModelosInformativos: []string{modeloInformativoEnMain},
		Origen: &origenDeLasEvals{Modo: origenConReutilizacion, Evals: entradas},
		Juez:   &JuezInformado{Modelo: "claude-opus-5-5"},
	}

	for _, entrada := range entradas {
		if entrada.Medida {
			informe.Origen.EvalsMedidas++
			informe.Tasas = append(informe.Tasas, seriesDeLaEval(entrada.Eval)...)
			informe.Evals = append(informe.Evals,
				ResultadoDeEval{Sesion: sesionDeLaEval(entrada.Eval, modeloQueDecideEnMain), Eval: entrada.Eval, Modelo: modeloQueDecideEnMain},
				ResultadoDeEval{Sesion: sesionDeLaEval(entrada.Eval, modeloInformativoEnMain), Eval: entrada.Eval, Modelo: modeloInformativoEnMain},
			)
		} else {
			informe.Origen.EvalsReutilizadas++
		}
	}

	return informe
}

// sesionDeLaEval es el nombre de la sesión de una eval con un modelo en un
// informe sintético.
func sesionDeLaEval(fichero, modelo string) string {
	return strings.TrimSuffix(fichero, ".yaml") + "-" + modelo + "-1"
}

// conLaSesionSinMedir es el informe con una sesión de la eval dada sin medir,
// la del modelo informativo.
func conLaSesionSinMedir(informe Informe, eval evalABuscar) Informe {
	informe.SesionesSinMedir = append(slices.Clone(informe.SesionesSinMedir), SesionSinMedir{
		Sesion: sesionDeLaEval(eval.fichero, modeloInformativoEnMain), Eval: eval.fichero, Modelo: modeloInformativoEnMain,
		Motivo: "límite de uso de la cuenta",
	})

	return informe
}

// conLaRespuestaSinJuzgar es el informe con la respuesta del modelo que decide
// de la eval dada sin juzgar.
func conLaRespuestaSinJuzgar(informe Informe, eval evalABuscar) Informe {
	juez := *informe.Juez
	juez.SinJuzgar = append(slices.Clone(juez.SinJuzgar), RespuestaSinJuzgar{
		Sesion: sesionDeLaEval(eval.fichero, modeloQueDecideEnMain), Motivo: "voto 2: tope de 35 s agotado",
	})
	informe.Juez = &juez

	return informe
}

// medidaEnMemoria es una ejecución del flujo que es medida de main: con el
// número dado, terminada en verde sobre la cabeza de la propuesta dada, creada
// el tiempo dado antes de la búsqueda, con el trabajo de cada skill terminado
// —el de la otra en rojo— y con su informe escrito, y con el informe dado en el
// de la skill buscada.
func medidaEnMemoria(t *testing.T, numero int64, propuesta int, antes time.Duration, informe Informe) ejecucionEnMemoria {
	t.Helper()

	contenido, err := json.MarshalIndent(informe, "", "  ")
	require.NoError(t, err)

	return ejecucionEnMemoria{
		ejecucion: ejecucionDelFlujo{
			ID: numero, Commit: cabezaDeLaPropuesta(propuesta), Status: statusTerminado, Conclusion: conclusionDeExito,
			Creada: haceDesdeLaBusqueda(antes),
		},
		propuesta: propuesta,
		trabajos: []trabajoDeLaEjecucion{
			{ID: numero * 100, Name: "cambios", Status: statusTerminado, Conclusion: conclusionDeExito},
			trabajoDeEvalsEnMemoria(trabajoDeLaSkillBuscada(numero), skillDeLaBusqueda, conclusionDeExito),
			trabajoDeEvalsEnMemoria(trabajoDeLaOtraSkill(numero), otraSkillDeLaBusqueda, "failure"),
			{ID: numero*10 + 3, Name: "medida del juez (" + skillDeLaBusqueda + ")", Status: statusTerminado, Conclusion: "skipped"},
		},
		informe: append(contenido, '\n'),
	}
}

// trabajoDeEvalsEnMemoria es el trabajo evals (<skill>) de una ejecución
// sintética, terminado con la conclusión dada y con su informe escrito.
func trabajoDeEvalsEnMemoria(id int64, skill, conclusion string) trabajoDeLaEjecucion {
	return trabajoDeLaEjecucion{
		ID: id, Name: "evals (" + skill + ")", Status: statusTerminado, Conclusion: conclusion,
		Steps: []pasoDeLaEjecucion{
			{Name: "Ejecutar las evals", Conclusion: conclusion},
			{Name: "El informe del trabajo está escrito", Conclusion: conclusionDeExito},
		},
	}
}

// conLaEjecucion es la ejecución sintética con su ejecución cambiada.
func conLaEjecucion(enMemoria ejecucionEnMemoria, cambiar func(*ejecucionDelFlujo)) ejecucionEnMemoria {
	cambiar(&enMemoria.ejecucion)

	return enMemoria
}

// conLaMarca es la ejecución sintética con la marca del informe escrito del
// trabajo de la skill dada con la conclusión dada, o sin ella si es vacía.
func conLaMarca(enMemoria ejecucionEnMemoria, skill, conclusion string) ejecucionEnMemoria {
	enMemoria.trabajos = slices.Clone(enMemoria.trabajos)

	for indice, trabajo := range enMemoria.trabajos {
		if trabajo.Name != "evals ("+skill+")" {
			continue
		}

		trabajo.Steps = slices.DeleteFunc(slices.Clone(trabajo.Steps), func(paso pasoDeLaEjecucion) bool {
			return paso.Name == "El informe del trabajo está escrito"
		})
		if conclusion != "" {
			trabajo.Steps = append(trabajo.Steps, pasoDeLaEjecucion{Name: "El informe del trabajo está escrito", Conclusion: conclusion})
		}

		enMemoria.trabajos[indice] = trabajo
	}

	return enMemoria
}

// conElInforme es la ejecución sintética con el contenido dado como informe
// del trabajo de la skill buscada.
func conElInforme(enMemoria ejecucionEnMemoria, contenido string) ejecucionEnMemoria {
	enMemoria.informe = []byte(contenido)

	return enMemoria
}

// reutilizadaDe es el resultado de la eval dada que la búsqueda reutiliza de la
// medida dada, en la que se midió: con su procedencia, sus aportes y sus
// series.
func reutilizadaDe(eval evalABuscar, medida ejecucionEnMemoria) origenDeEval {
	return origenDeEval{
		Eval: eval.fichero, Huella: eval.huella,
		Procedencia: &procedenciaDeLaMedida{
			Ejecucion: medida.ejecucion.ID, Trabajo: trabajoDeLaSkillBuscada(medida.ejecucion.ID),
			Commit: medida.ejecucion.Commit, Propuesta: medida.propuesta,
			Fecha: medida.ejecucion.Creada.Format(time.RFC3339),
		},
		Aportes: aportesDeLaEvalEnMain(), Tasas: seriesDeLaEval(eval.fichero),
	}
}

// sinResultadoDeMain es la eval dada medida porque la búsqueda no le dio
// resultado.
func sinResultadoDeMain(eval evalABuscar) origenDeEval {
	return origenDeEval{Eval: eval.fichero, Huella: eval.huella, Medida: true, Motivo: motivoSinResultadoDeMain}
}

// lecturasDeLasMedidas son las lecturas de la búsqueda que pasa por las medidas
// dadas: las dos listas y, de cada medida, sus trabajos y el informe del de la
// skill buscada.
func lecturasDeLasMedidas(medidas ...int64) []string {
	lecturas := []string{lecturaDeLasEjecuciones, lecturaDeLasPropuestas}
	for _, medida := range medidas {
		lecturas = append(lecturas, lecturaDeLosTrabajos(medida), lecturaDelInforme(medida, trabajoDeLaSkillBuscada(medida)))
	}

	return lecturas
}

// casoDeLaBusqueda es un caso de TestBusquedaDeMedidasDeMain: las evals que
// buscan, la huella del juicio de la descripción que busca, vacía si no busca
// ninguno, las ejecuciones de la plataforma, los fallos de sus lecturas, si el
// plazo está agotado, y lo que espera: las evals, el juicio reutilizado, nil si
// no reutiliza ninguno, el error de la plataforma y las lecturas, en orden.
type casoDeLaBusqueda struct {
	nombre            string
	buscadas          []evalABuscar
	descripcion       string
	ejecuciones       []ejecucionEnMemoria
	fallos            map[string]error
	plazoAgotado      bool
	evals             []origenDeEval
	juicio            *descripcionDeLaDecision
	plataformaFallida *errorDeLaPlataformaEnElOrigen
	lecturas          []string
}

// TestBusquedaDeMedidasDeMain fija la búsqueda de resultados en las medidas de
// main (FR-010 a FR-013 de H30; data-model §7 y §8 de H30; research D8, D9 y
// D17 de H30), con la plataforma en memoria y un instante fijo, un subtest por
// caso: lo que no es medida de main —una ejecución cancelada o sin terminar, la
// de un commit que no es la cabeza de una propuesta fusionada o que es otro
// commit de ella, la que no tiene la marca del informe escrito en el trabajo de
// alguna skill y la de 90 días o más— no cuenta; gana la más reciente con la
// misma huella y la búsqueda para en cuanto cada eval tiene resultado; no pasa
// de la décima medida; una sesión sin medir o una respuesta sin juzgar de la
// eval no se reutilizan; la procedencia de una reutilizada en main
// es la original, y la de 90 días o más deja la eval sin resultado; y la
// primera lectura que falla, o el plazo agotado, paran la búsqueda: lo que
// queda se mide, con la orden y su error en el resultado. Y, en el mismo
// recorrido, el juicio de la descripción de la skill (FR-030 de H31; research
// D4 de H31; contracts/juez-de-las-descripciones.md §2 de H31): se reutiliza el
// de la medida más reciente con la misma huella y juicio completo, con su
// procedencia, y ninguno con otra huella, de otra skill, sin juzgar, de lo que
// no es medida de main o de una procedencia de 90 días o más.
func TestBusquedaDeMedidasDeMain(t *testing.T) {
	t.Parallel()

	casos := slices.Concat(casosDeLoQueNoEsMedida(t), casosDelOrdenDeLasMedidas(t), casosDeLosResultados(t),
		casosDeLosFallos(t), casosDeLaDescripcion(t), casosDeLaDescripcionQueNoSeReutiliza(t))

	for _, caso := range casos {
		t.Run(caso.nombre, func(t *testing.T) {
			t.Parallel()

			comprobarLaBusqueda(t, caso)
		})
	}
}

// comprobarLaBusqueda busca las evals del caso en su plataforma en memoria, con
// el plazo agotado si lo pide, y compara las evals, el error y las lecturas.
func comprobarLaBusqueda(t *testing.T, caso casoDeLaBusqueda) {
	t.Helper()

	plataforma := &plataformaEnMemoria{propuestas: propuestasDeMain(), ejecuciones: caso.ejecuciones, fallos: caso.fallos}

	ctx := t.Context()
	if caso.plazoAgotado {
		var cancelar context.CancelFunc

		ctx, cancelar = context.WithDeadline(ctx, time.Now())
		defer cancelar()
	}

	evals, juicio, plataformaFallida := buscarEnLasMedidasDeMain(ctx, plataforma, skillDeLaBusqueda, caso.buscadas,
		caso.descripcion, instanteDeLaBusqueda)

	assert.Equal(t, caso.evals, evals, "las evals de la búsqueda")
	assert.Equal(t, caso.juicio, juicio, "el juicio de la descripción reutilizado")
	assert.Equal(t, caso.plataformaFallida, plataformaFallida, "el error de la plataforma")
	assert.Equal(t, caso.lecturas, plataforma.lecturas, "las lecturas de la plataforma, en orden")
}

// casosDeLoQueNoEsMedida son los de lo que no es medida de main: cada uno, una
// ejecución más reciente que no cuenta, con resultado de las dos evals, delante
// de una medida de main de la que se reutilizan.
func casosDeLoQueNoEsMedida(t *testing.T) []casoDeLaBusqueda {
	t.Helper()

	reciente := medidaEnMemoria(t, 101, 186, time.Hour, informeDeMain(medidaAqui(evalBuscadaA, evalBuscadaA.huella),
		medidaAqui(evalBuscadaB, evalBuscadaB.huella)))
	anterior := medidaEnMemoria(t, 100, 185, 2*time.Hour, informeDeMain(medidaAqui(evalBuscadaA, evalBuscadaA.huella),
		medidaAqui(evalBuscadaB, evalBuscadaB.huella)))
	deLaAnterior := []origenDeEval{reutilizadaDe(evalBuscadaA, anterior), reutilizadaDe(evalBuscadaB, anterior)}

	sinContar := func(nombre string, noCuenta ejecucionEnMemoria, leeSusTrabajos bool) casoDeLaBusqueda {
		lecturas := lecturasDeLasMedidas(anterior.ejecucion.ID)
		if leeSusTrabajos {
			lecturas = slices.Insert(lecturas, 2, lecturaDeLosTrabajos(noCuenta.ejecucion.ID))
		}

		return casoDeLaBusqueda{
			nombre: nombre, buscadas: []evalABuscar{evalBuscadaA, evalBuscadaB},
			ejecuciones: []ejecucionEnMemoria{noCuenta, anterior}, evals: deLaAnterior, lecturas: lecturas,
		}
	}

	return []casoDeLaBusqueda{
		sinContar("cancelada", conLaEjecucion(reciente, func(ejecucion *ejecucionDelFlujo) {
			ejecucion.Conclusion = conclusionCancelada
		}), false),
		sinContar("sin-terminar", conLaEjecucion(reciente, func(ejecucion *ejecucionDelFlujo) {
			ejecucion.Status, ejecucion.Conclusion = "in_progress", ""
		}), false),
		sinContar("commit-que-no-es-la-cabeza-de-una-fusionada", conLaEjecucion(reciente, func(ejecucion *ejecucionDelFlujo) {
			ejecucion.Commit = "cabeza-de-una-propuesta-cerrada-sin-fusionar"
		}), false),
		sinContar("otro-commit-de-la-misma-propuesta", conLaEjecucion(reciente, func(ejecucion *ejecucionDelFlujo) {
			ejecucion.Commit = "anterior-a-la-" + cabezaDeLaPropuesta(186)
		}), false),
		sinContar("sin-la-marca", conLaMarca(reciente, skillDeLaBusqueda, ""), true),
		sinContar("con-la-marca-sin-success", conLaMarca(reciente, skillDeLaBusqueda, "skipped"), true),
		sinContar("sin-la-marca-en-el-trabajo-de-otra-skill", conLaMarca(reciente, otraSkillDeLaBusqueda, ""), true),
		medidaDeNoventaDias(t),
	}
}

// medidaDeNoventaDias es el caso de una medida de 90 días o más, de la que no
// se lee nada, detrás de una de un segundo menos, de la que sí.
func medidaDeNoventaDias(t *testing.T) casoDeLaBusqueda {
	t.Helper()

	casiNoventa := medidaEnMemoria(t, 101, 186, noventaDiasDeRegistro-time.Second,
		informeDeMain(medidaAqui(evalBuscadaA, evalBuscadaA.huella)))
	noventa := medidaEnMemoria(t, 100, 185, noventaDiasDeRegistro,
		informeDeMain(medidaAqui(evalBuscadaA, evalBuscadaA.huella), medidaAqui(evalBuscadaB, evalBuscadaB.huella)))

	return casoDeLaBusqueda{
		nombre: "medida-de-90-dias", buscadas: []evalABuscar{evalBuscadaA, evalBuscadaB},
		ejecuciones: []ejecucionEnMemoria{casiNoventa, noventa},
		evals:       []origenDeEval{reutilizadaDe(evalBuscadaA, casiNoventa), sinResultadoDeMain(evalBuscadaB)},
		lecturas:    lecturasDeLasMedidas(casiNoventa.ejecucion.ID),
	}
}

// casosDelOrdenDeLasMedidas son los del orden en que se recorren las medidas:
// de la más reciente a la más antigua, como mucho diez.
func casosDelOrdenDeLasMedidas(t *testing.T) []casoDeLaBusqueda {
	t.Helper()

	ambas := informeDeMain(medidaAqui(evalBuscadaA, evalBuscadaA.huella), medidaAqui(evalBuscadaB, evalBuscadaB.huella))
	reciente := medidaEnMemoria(t, 101, 186, time.Hour, ambas)
	anterior := medidaEnMemoria(t, 100, 185, 2*time.Hour, ambas)
	aLaMismaHora := medidaEnMemoria(t, 100, 185, time.Hour, ambas)

	conOtraHuella := medidaEnMemoria(t, 101, 186, time.Hour,
		informeDeMain(medidaAqui(evalBuscadaA, otraHuella), medidaAqui(evalBuscadaB, evalBuscadaB.huella)))

	return []casoDeLaBusqueda{
		{
			nombre: "la-mas-reciente-gana", buscadas: []evalABuscar{evalBuscadaA, evalBuscadaB},
			ejecuciones: []ejecucionEnMemoria{anterior, reciente},
			evals:       []origenDeEval{reutilizadaDe(evalBuscadaA, reciente), reutilizadaDe(evalBuscadaB, reciente)},
			lecturas:    lecturasDeLasMedidas(reciente.ejecucion.ID),
		},
		{
			nombre: "a-la-misma-hora-la-de-mayor-numero", buscadas: []evalABuscar{evalBuscadaA, evalBuscadaB},
			ejecuciones: []ejecucionEnMemoria{aLaMismaHora, reciente},
			evals:       []origenDeEval{reutilizadaDe(evalBuscadaA, reciente), reutilizadaDe(evalBuscadaB, reciente)},
			lecturas:    lecturasDeLasMedidas(reciente.ejecucion.ID),
		},
		{
			nombre: "con-otra-huella-sigue-con-la-anterior", buscadas: []evalABuscar{evalBuscadaA, evalBuscadaB},
			ejecuciones: []ejecucionEnMemoria{conOtraHuella, anterior},
			evals:       []origenDeEval{reutilizadaDe(evalBuscadaA, anterior), reutilizadaDe(evalBuscadaB, conOtraHuella)},
			lecturas:    lecturasDeLasMedidas(conOtraHuella.ejecucion.ID, anterior.ejecucion.ID),
		},
		laUndecimaQuedaFuera(t, false),
		laUndecimaQuedaFuera(t, true),
		{
			nombre: "sin-evals-que-buscar", buscadas: nil, ejecuciones: []ejecucionEnMemoria{reciente},
			evals: []origenDeEval{}, lecturas: nil,
		},
	}
}

// laUndecimaQuedaFuera es el caso de once medidas de main en las que la
// primera eval solo tiene resultado en la más antigua, por la que la búsqueda
// no pasa: la segunda se reutiliza de la más reciente y la primera se mide. Con
// una ejecución sin la marca entre ellas, que no es medida, la undécima
// ejecución es la décima medida y la primera eval se reutiliza de ella.
func laUndecimaQuedaFuera(t *testing.T, conUnaQueNoEsMedida bool) casoDeLaBusqueda {
	t.Helper()

	var (
		ejecuciones []ejecucionEnMemoria
		recorridas  []int64
	)

	soloLaSegunda := informeDeMain(medidaAqui(evalBuscadaB, evalBuscadaB.huella))

	for orden := range 10 {
		numero := int64(120 - orden)
		medida := medidaEnMemoria(t, numero, 199-orden, time.Duration(orden+1)*time.Hour, soloLaSegunda)

		if conUnaQueNoEsMedida && orden == 2 {
			medida = conLaMarca(medida, otraSkillDeLaBusqueda, "")
		} else {
			recorridas = append(recorridas, numero)
		}

		ejecuciones = append(ejecuciones, medida)
	}

	undecima := medidaEnMemoria(t, 110, 189, 11*time.Hour, informeDeMain(medidaAqui(evalBuscadaA, evalBuscadaA.huella)))
	ejecuciones = append(ejecuciones, undecima)

	caso := casoDeLaBusqueda{
		nombre: "la-undecima-queda-fuera", buscadas: []evalABuscar{evalBuscadaA, evalBuscadaB}, ejecuciones: ejecuciones,
		evals:    []origenDeEval{sinResultadoDeMain(evalBuscadaA), reutilizadaDe(evalBuscadaB, ejecuciones[0])},
		lecturas: lecturasDeLasMedidas(recorridas...),
	}

	if conUnaQueNoEsMedida {
		caso.nombre = "una-que-no-es-medida-no-cuenta-en-las-diez"
		caso.evals[0] = reutilizadaDe(evalBuscadaA, undecima)
		caso.lecturas = slices.Insert(lecturasDeLasMedidas(append(recorridas, undecima.ejecucion.ID)...), 6,
			lecturaDeLosTrabajos(118))
	}

	return caso
}

// casosDeLosResultados son los de qué resultado de un informe de main se
// reutiliza: no el incompleto, sí el reutilizado allí con su procedencia, y
// nunca uno de una procedencia de 90 días o más.
func casosDeLosResultados(t *testing.T) []casoDeLaBusqueda {
	t.Helper()

	ambas := informeDeMain(medidaAqui(evalBuscadaA, evalBuscadaA.huella), medidaAqui(evalBuscadaB, evalBuscadaB.huella))
	sinMedir := medidaEnMemoria(t, 101, 186, time.Hour, conLaSesionSinMedir(ambas, evalBuscadaA))
	sinJuzgar := medidaEnMemoria(t, 101, 186, time.Hour, conLaRespuestaSinJuzgar(ambas, evalBuscadaA))
	sinJuzgarDeOtraSesion := medidaEnMemoria(t, 101, 186, time.Hour,
		conLaRespuestaSinJuzgar(ambas, evalABuscar{fichero: "03-sin-sesion-en-el-informe.yaml"}))
	anterior := medidaEnMemoria(t, 100, 185, 2*time.Hour, ambas)

	reutilizadaA := reutilizadaEnMain(evalBuscadaA, 90, 180, 20*24*time.Hour)
	conReutilizada := medidaEnMemoria(t, 101, 186, time.Hour,
		informeDeMain(reutilizadaA, medidaAqui(evalBuscadaB, evalBuscadaB.huella)))

	casiNoventa := reutilizadaEnMain(evalBuscadaB, 80, 181, noventaDiasDeRegistro-time.Second)
	conProcedenciasViejas := medidaEnMemoria(t, 101, 186, time.Hour,
		informeDeMain(reutilizadaEnMain(evalBuscadaA, 79, 180, noventaDiasDeRegistro), casiNoventa))

	return []casoDeLaBusqueda{
		{
			nombre: "una-sesion-sin-medir", buscadas: []evalABuscar{evalBuscadaA, evalBuscadaB},
			ejecuciones: []ejecucionEnMemoria{sinMedir},
			evals:       []origenDeEval{sinResultadoDeMain(evalBuscadaA), reutilizadaDe(evalBuscadaB, sinMedir)},
			lecturas:    lecturasDeLasMedidas(sinMedir.ejecucion.ID),
		},
		{
			nombre: "una-respuesta-sin-juzgar", buscadas: []evalABuscar{evalBuscadaA, evalBuscadaB},
			ejecuciones: []ejecucionEnMemoria{sinJuzgar},
			evals:       []origenDeEval{sinResultadoDeMain(evalBuscadaA), reutilizadaDe(evalBuscadaB, sinJuzgar)},
			lecturas:    lecturasDeLasMedidas(sinJuzgar.ejecucion.ID),
		},
		{
			nombre:   "una-respuesta-sin-juzgar-de-una-sesion-que-el-informe-no-tiene",
			buscadas: []evalABuscar{evalBuscadaA, evalBuscadaB}, ejecuciones: []ejecucionEnMemoria{sinJuzgarDeOtraSesion},
			evals:    []origenDeEval{sinResultadoDeMain(evalBuscadaA), sinResultadoDeMain(evalBuscadaB)},
			lecturas: lecturasDeLasMedidas(sinJuzgarDeOtraSesion.ejecucion.ID),
		},
		{
			nombre: "incompleta-sigue-con-la-anterior", buscadas: []evalABuscar{evalBuscadaA, evalBuscadaB},
			ejecuciones: []ejecucionEnMemoria{sinMedir, anterior},
			evals:       []origenDeEval{reutilizadaDe(evalBuscadaA, anterior), reutilizadaDe(evalBuscadaB, sinMedir)},
			lecturas:    lecturasDeLasMedidas(sinMedir.ejecucion.ID, anterior.ejecucion.ID),
		},
		{
			nombre: "reutilizada-en-main-con-su-procedencia-original", buscadas: []evalABuscar{evalBuscadaA, evalBuscadaB},
			ejecuciones: []ejecucionEnMemoria{conReutilizada},
			evals:       []origenDeEval{reutilizadaA, reutilizadaDe(evalBuscadaB, conReutilizada)},
			lecturas:    lecturasDeLasMedidas(conReutilizada.ejecucion.ID),
		},
		{
			nombre: "procedencia-de-90-dias", buscadas: []evalABuscar{evalBuscadaA, evalBuscadaB},
			ejecuciones: []ejecucionEnMemoria{conProcedenciasViejas},
			evals:       []origenDeEval{sinResultadoDeMain(evalBuscadaA), casiNoventa},
			lecturas:    lecturasDeLasMedidas(conProcedenciasViejas.ejecucion.ID),
		},
		{
			nombre: "informe-anterior-al-origen", buscadas: []evalABuscar{evalBuscadaA},
			ejecuciones: []ejecucionEnMemoria{conElInforme(sinMedir, `{"skill": "normativa", "origen": null}`+"\n")},
			evals:       []origenDeEval{sinResultadoDeMain(evalBuscadaA)},
			lecturas:    lecturasDeLasMedidas(sinMedir.ejecucion.ID),
		},
	}
}

// casosDeLosFallos son los de las lecturas que fallan: la de un informe que no
// se puede sacar o leer no aporta nada y la búsqueda sigue; cualquier otra, y
// el plazo agotado, la paran con su orden y su error, y lo que queda se mide.
func casosDeLosFallos(t *testing.T) []casoDeLaBusqueda {
	t.Helper()

	soloLaPrimera := medidaEnMemoria(t, 101, 186, time.Hour, informeDeMain(medidaAqui(evalBuscadaA, evalBuscadaA.huella)))
	ambas := informeDeMain(medidaAqui(evalBuscadaA, evalBuscadaA.huella), medidaAqui(evalBuscadaB, evalBuscadaB.huella))
	anterior := medidaEnMemoria(t, 100, 185, 2*time.Hour, ambas)
	antesAun := medidaEnMemoria(t, 99, 184, 3*time.Hour, ambas)
	todas := []ejecucionEnMemoria{soloLaPrimera, anterior, antesAun}
	deLaPrimera := []origenDeEval{reutilizadaDe(evalBuscadaA, soloLaPrimera), sinResultadoDeMain(evalBuscadaB)}

	return []casoDeLaBusqueda{
		{
			nombre: "falla-la-primera-lectura", buscadas: []evalABuscar{evalBuscadaA, evalBuscadaB}, ejecuciones: todas,
			fallos: map[string]error{lecturaDeLasEjecuciones: errors.New("la lista de ejecuciones del flujo no se puede leer")},
			evals:  []origenDeEval{sinResultadoDeMain(evalBuscadaA), sinResultadoDeMain(evalBuscadaB)},
			plataformaFallida: &errorDeLaPlataformaEnElOrigen{
				Orden: ordenDeLasEjecucionesDelFlujo, Error: "la lista de ejecuciones del flujo no se puede leer",
			},
			lecturas: []string{lecturaDeLasEjecuciones},
		},
		{
			nombre: "falla-una-orden", buscadas: []evalABuscar{evalBuscadaA, evalBuscadaB}, ejecuciones: todas,
			fallos: map[string]error{lecturaDeLosTrabajos(100): &errorDeUnaOrden{
				orden: ordenDeLosTrabajos, variables: []string{"EJECUCION_ANTERIOR=100"}, causa: errors.New("exit status 1"),
				salidaDeError: "HTTP 502: Bad Gateway (https://api.github.com/repos/jmorenobl/kitlegal/actions/runs/100)\n",
			}},
			evals: deLaPrimera,
			plataformaFallida: &errorDeLaPlataformaEnElOrigen{
				Orden: ordenDeLosTrabajos + ", con EJECUCION_ANTERIOR=100",
				Error: "HTTP 502: Bad Gateway (https://api.github.com/repos/jmorenobl/kitlegal/actions/runs/100)",
			},
			lecturas: append(lecturasDeLasMedidas(101), lecturaDeLosTrabajos(100)),
		},
		{
			nombre: "falla-la-orden-del-registro-sin-salida-de-error", buscadas: []evalABuscar{evalBuscadaA, evalBuscadaB},
			ejecuciones: todas,
			fallos: map[string]error{lecturaDelInforme(100, trabajoDeLaSkillBuscada(100)): &errorDeUnaOrden{
				orden: ordenDelRegistroDeUnTrabajo, variables: []string{"EJECUCION=100", "TRABAJO=1001"},
				causa: errors.New("exit status 1"),
			}},
			evals: deLaPrimera,
			plataformaFallida: &errorDeLaPlataformaEnElOrigen{
				Orden: ordenDelRegistroDeUnTrabajo + ", con EJECUCION=100, TRABAJO=1001", Error: "exit status 1",
			},
			lecturas: lecturasDeLasMedidas(101, 100),
		},
		{
			nombre: "plazo-agotado", buscadas: []evalABuscar{evalBuscadaA, evalBuscadaB}, ejecuciones: todas,
			fallos:       map[string]error{lecturaDelInforme(100, trabajoDeLaSkillBuscada(100)): errPlazoEnCurso},
			plazoAgotado: true,
			evals:        deLaPrimera,
			plataformaFallida: &errorDeLaPlataformaEnElOrigen{
				Orden: ordenDelRegistroDeUnTrabajo + ", con EJECUCION=100, TRABAJO=1001", Error: "plazo de 10 minutos agotado",
			},
			lecturas: lecturasDeLasMedidas(101, 100),
		},
		{
			nombre: "registro-sin-informe", buscadas: []evalABuscar{evalBuscadaA, evalBuscadaB}, ejecuciones: todas,
			fallos: map[string]error{lecturaDelInforme(100, trabajoDeLaSkillBuscada(100)): fmt.Errorf(
				"el informe del trabajo 1001 de la ejecución 100: %w", errors.New(`el registro no trae la marca "--- inicio de informe.json ---"`))},
			evals:    []origenDeEval{reutilizadaDe(evalBuscadaA, soloLaPrimera), reutilizadaDe(evalBuscadaB, antesAun)},
			lecturas: lecturasDeLasMedidas(101, 100, 99),
		},
		{
			nombre: "informe-que-no-es-json", buscadas: []evalABuscar{evalBuscadaA, evalBuscadaB},
			ejecuciones: []ejecucionEnMemoria{soloLaPrimera, conElInforme(anterior, "{\n  \"skill\": \n"), antesAun},
			evals:       []origenDeEval{reutilizadaDe(evalBuscadaA, soloLaPrimera), reutilizadaDe(evalBuscadaB, antesAun)},
			lecturas:    lecturasDeLasMedidas(101, 100, 99),
		},
	}
}

// La huella del juicio de la descripción de la skill buscada en los casos de
// la descripción, y la frase y los motivos de los juicios de las medidas de
// main sintéticas.
const (
	huellaDeLaDescripcionBuscada = "d0d0d0d0d0d0d0d0d0d0d0d0d0d0d0d0d0d0d0d0d0d0d0d0d0d0d0d0d0d0d0d0"
	fraseMarcadaEnMain           = "Úsala siempre que la respuesta dependa de lo que dice una norma"
	motivoMarcadaEnMain          = "Da una orden a quien la lee con un imperativo con un pronombre pegado."
	motivoSinMarcarEnMain        = "Todas las frases hablan de la skill en tercera persona."
	motivoSinJuzgarEnMain        = "voto 1: la sesión del juez terminó con error: exit status 1"
	claseDeLaDescripcionMain     = "habla_en_primera_o_segunda_persona"
)

// bloqueJuzgadoEnMain es el bloque descripcion del informe de un trabajo que
// juzgó la descripción de la skill dada con esa huella, con la forma de
// contracts/juez-de-las-descripciones.md §6 de H31: marcada con tres «si» con
// su frase o sin marcar con un primer «no».
func bloqueJuzgadoEnMain(skill, huella string, marcada bool) map[string]any {
	bloque := map[string]any{
		"skill": skill, "clase": claseDeLaDescripcionMain, "huella": huella, "juzgada": true, "marcada": marcada,
		"frase": "", "motivo": motivoSinMarcarEnMain,
		"votos": []any{map[string]any{"respuesta": "no", "frase": "", "motivo": motivoSinMarcarEnMain}},
	}

	if marcada {
		voto := map[string]any{"respuesta": "si", "frase": fraseMarcadaEnMain, "motivo": motivoMarcadaEnMain}
		bloque["frase"], bloque["motivo"], bloque["votos"] = fraseMarcadaEnMain, motivoMarcadaEnMain, []any{voto, voto, voto}
	}

	return bloque
}

// bloqueSinJuzgarEnMain es el bloque descripcion del informe de un trabajo que
// dejó sin juzgar la descripción de la skill buscada con esa huella.
func bloqueSinJuzgarEnMain(huella string) map[string]any {
	return map[string]any{
		"skill": skillDeLaBusqueda, "clase": claseDeLaDescripcionMain, "huella": huella, "juzgada": true,
		"marcada": false, "frase": "", "motivo": "", "votos": []any{}, "sin_juzgar": motivoSinJuzgarEnMain,
	}
}

// bloqueReutilizadoEnMain es el bloque descripcion del informe de un trabajo
// que reutilizó, sin marcar, el juicio de la descripción de la skill buscada
// con esa huella de la ejecución dada, de la propuesta dada, creada el tiempo
// dado antes de la búsqueda.
func bloqueReutilizadoEnMain(huella string, ejecucion int64, propuesta int, antes time.Duration) map[string]any {
	return map[string]any{
		"skill": skillDeLaBusqueda, "clase": claseDeLaDescripcionMain, "huella": huella, "juzgada": false,
		"procedencia": map[string]any{
			"ejecucion": ejecucion, "trabajo": trabajoDeLaSkillBuscada(ejecucion), "commit": cabezaDeLaPropuesta(propuesta),
			"propuesta": propuesta, "fecha": fechaDesdeLaBusqueda(antes),
		},
		"marcada": false, "frase": "", "motivo": motivoSinMarcarEnMain,
	}
}

// conLaDescripcion es la ejecución sintética con el bloque descripcion dado en
// el informe del trabajo de la skill buscada.
func conLaDescripcion(t *testing.T, enMemoria ejecucionEnMemoria, bloque map[string]any) ejecucionEnMemoria {
	t.Helper()

	var informe map[string]any
	require.NoError(t, json.Unmarshal(enMemoria.informe, &informe))

	informe["descripcion"] = bloque

	contenido, err := json.MarshalIndent(informe, "", "  ")
	require.NoError(t, err)

	contenido = append(contenido, '\n')
	enMemoria.informe = contenido

	return enMemoria
}

// juicioReutilizadoDe es el juicio de la descripción de la skill buscada que
// la búsqueda reutiliza de la medida dada, en la que se juzgó: con su
// procedencia, la medida, y su marca, su frase y su motivo.
func juicioReutilizadoDe(medida ejecucionEnMemoria, marcada bool) *descripcionDeLaDecision {
	juicio := &juicioReutilizado{Motivo: motivoSinMarcarEnMain}
	if marcada {
		juicio = &juicioReutilizado{Marcada: true, Frase: fraseMarcadaEnMain, Motivo: motivoMarcadaEnMain}
	}

	return &descripcionDeLaDecision{
		Skill: skillDeLaBusqueda, Huella: huellaDeLaDescripcionBuscada,
		Procedencia: &procedenciaDeLaMedida{
			Ejecucion: medida.ejecucion.ID, Trabajo: trabajoDeLaSkillBuscada(medida.ejecucion.ID),
			Commit: medida.ejecucion.Commit, Propuesta: medida.propuesta,
			Fecha: medida.ejecucion.Creada.Format(time.RFC3339),
		},
		Juicio: juicio,
	}
}

// casosDeLaDescripcion son los del juicio de la descripción que la búsqueda
// reutiliza: el de la medida más reciente con su huella y juicio completo,
// marcado o no, con su procedencia, o con la original si allí se reutilizó; en
// el mismo recorrido que las evals, que no para hasta que todo lo buscado tiene
// resultado.
func casosDeLaDescripcion(t *testing.T) []casoDeLaBusqueda {
	t.Helper()

	ambas := informeDeMain(medidaAqui(evalBuscadaA, evalBuscadaA.huella), medidaAqui(evalBuscadaB, evalBuscadaB.huella))
	reciente := medidaEnMemoria(t, 101, 186, time.Hour, ambas)
	anterior := medidaEnMemoria(t, 100, 185, 2*time.Hour, ambas)
	marcadaReciente := conLaDescripcion(t, reciente, bloqueJuzgadoEnMain(skillDeLaBusqueda, huellaDeLaDescripcionBuscada, true))
	sinMarcarAnterior := conLaDescripcion(t, anterior,
		bloqueJuzgadoEnMain(skillDeLaBusqueda, huellaDeLaDescripcionBuscada, false))
	conOtraHuellaReciente := conLaDescripcion(t, reciente, bloqueJuzgadoEnMain(skillDeLaBusqueda, otraHuella, true))
	reutilizadaReciente := conLaDescripcion(t, reciente,
		bloqueReutilizadoEnMain(huellaDeLaDescripcionBuscada, 90, 180, 20*24*time.Hour))

	return []casoDeLaBusqueda{
		{
			nombre: "descripcion-marcada-con-la-misma-huella", descripcion: huellaDeLaDescripcionBuscada,
			ejecuciones: []ejecucionEnMemoria{marcadaReciente, sinMarcarAnterior}, evals: []origenDeEval{},
			juicio: juicioReutilizadoDe(marcadaReciente, true), lecturas: lecturasDeLasMedidas(101),
		},
		{
			nombre: "descripcion-con-otra-huella-sigue-con-la-anterior", descripcion: huellaDeLaDescripcionBuscada,
			ejecuciones: []ejecucionEnMemoria{conOtraHuellaReciente, sinMarcarAnterior}, evals: []origenDeEval{},
			juicio: juicioReutilizadoDe(sinMarcarAnterior, false), lecturas: lecturasDeLasMedidas(101, 100),
		},
		{
			nombre: "descripcion-reutilizada-en-main-con-su-procedencia-original", descripcion: huellaDeLaDescripcionBuscada,
			ejecuciones: []ejecucionEnMemoria{reutilizadaReciente}, evals: []origenDeEval{},
			juicio: &descripcionDeLaDecision{
				Skill: skillDeLaBusqueda, Huella: huellaDeLaDescripcionBuscada,
				Procedencia: &procedenciaDeLaMedida{
					Ejecucion: 90, Trabajo: trabajoDeLaSkillBuscada(90), Commit: cabezaDeLaPropuesta(180), Propuesta: 180,
					Fecha: fechaDesdeLaBusqueda(20 * 24 * time.Hour),
				},
				Juicio: &juicioReutilizado{Motivo: motivoSinMarcarEnMain},
			},
			lecturas: lecturasDeLasMedidas(101),
		},
		{
			nombre: "evals-y-descripcion-en-la-misma-medida", buscadas: []evalABuscar{evalBuscadaA, evalBuscadaB},
			descripcion: huellaDeLaDescripcionBuscada, ejecuciones: []ejecucionEnMemoria{marcadaReciente, sinMarcarAnterior},
			evals:  []origenDeEval{reutilizadaDe(evalBuscadaA, reciente), reutilizadaDe(evalBuscadaB, reciente)},
			juicio: juicioReutilizadoDe(marcadaReciente, true), lecturas: lecturasDeLasMedidas(101),
		},
		{
			nombre: "la-descripcion-sigue-el-recorrido-tras-las-evals", buscadas: []evalABuscar{evalBuscadaA, evalBuscadaB},
			descripcion: huellaDeLaDescripcionBuscada, ejecuciones: []ejecucionEnMemoria{reciente, sinMarcarAnterior},
			evals:  []origenDeEval{reutilizadaDe(evalBuscadaA, reciente), reutilizadaDe(evalBuscadaB, reciente)},
			juicio: juicioReutilizadoDe(sinMarcarAnterior, false), lecturas: lecturasDeLasMedidas(101, 100),
		},
		{
			nombre: "sin-descripcion-que-buscar", buscadas: []evalABuscar{evalBuscadaA, evalBuscadaB},
			ejecuciones: []ejecucionEnMemoria{marcadaReciente, sinMarcarAnterior},
			evals:       []origenDeEval{reutilizadaDe(evalBuscadaA, reciente), reutilizadaDe(evalBuscadaB, reciente)},
			lecturas:    lecturasDeLasMedidas(101),
		},
	}
}

// casosDeLaDescripcionQueNoSeReutiliza son los de 0 juicios reutilizados: con
// otra huella, de otra skill, sin juzgar, de una medida de main sin bloque, de
// lo que no es medida de main —una ejecución cancelada, la de 90 días o más o
// la que no tiene la marca del informe escrito— o de una procedencia de 90 días
// o más; y la primera lectura que falla, que para la búsqueda.
func casosDeLaDescripcionQueNoSeReutiliza(t *testing.T) []casoDeLaBusqueda {
	t.Helper()

	ambas := informeDeMain(medidaAqui(evalBuscadaA, evalBuscadaA.huella), medidaAqui(evalBuscadaB, evalBuscadaB.huella))
	reciente := medidaEnMemoria(t, 101, 186, time.Hour, ambas)
	juzgada := conLaDescripcion(t, reciente, bloqueJuzgadoEnMain(skillDeLaBusqueda, huellaDeLaDescripcionBuscada, true))
	noventa := conLaDescripcion(t, medidaEnMemoria(t, 101, 186, noventaDiasDeRegistro, ambas),
		bloqueJuzgadoEnMain(skillDeLaBusqueda, huellaDeLaDescripcionBuscada, true))

	ninguno := func(nombre string, ejecucion ejecucionEnMemoria, lecturas []string) casoDeLaBusqueda {
		return casoDeLaBusqueda{
			nombre: nombre, descripcion: huellaDeLaDescripcionBuscada, ejecuciones: []ejecucionEnMemoria{ejecucion},
			evals: []origenDeEval{}, lecturas: lecturas,
		}
	}

	return []casoDeLaBusqueda{
		ninguno("descripcion-con-otra-huella",
			conLaDescripcion(t, reciente, bloqueJuzgadoEnMain(skillDeLaBusqueda, otraHuella, false)),
			lecturasDeLasMedidas(101)),
		ninguno("descripcion-de-otra-skill",
			conLaDescripcion(t, reciente, bloqueJuzgadoEnMain(otraSkillDeLaBusqueda, huellaDeLaDescripcionBuscada, false)),
			lecturasDeLasMedidas(101)),
		ninguno("descripcion-sin-juzgar", conLaDescripcion(t, reciente, bloqueSinJuzgarEnMain(huellaDeLaDescripcionBuscada)),
			lecturasDeLasMedidas(101)),
		ninguno("medida-sin-bloque-de-la-descripcion", reciente, lecturasDeLasMedidas(101)),
		ninguno("descripcion-de-una-ejecucion-cancelada", conLaEjecucion(juzgada, func(ejecucion *ejecucionDelFlujo) {
			ejecucion.Conclusion = conclusionCancelada
		}), lecturasDeLasMedidas()),
		ninguno("descripcion-de-una-medida-de-90-dias", noventa, lecturasDeLasMedidas()),
		ninguno("descripcion-de-una-ejecucion-sin-la-marca", conLaMarca(juzgada, otraSkillDeLaBusqueda, ""),
			[]string{lecturaDeLasEjecuciones, lecturaDeLasPropuestas, lecturaDeLosTrabajos(101)}),
		ninguno("descripcion-de-una-procedencia-de-90-dias",
			conLaDescripcion(t, reciente, bloqueReutilizadoEnMain(huellaDeLaDescripcionBuscada, 79, 180, noventaDiasDeRegistro)),
			lecturasDeLasMedidas(101)),
		{
			nombre: "descripcion-con-la-primera-lectura-que-falla", descripcion: huellaDeLaDescripcionBuscada,
			ejecuciones: []ejecucionEnMemoria{juzgada},
			fallos:      map[string]error{lecturaDeLasEjecuciones: errors.New("la lista de ejecuciones del flujo no se puede leer")},
			evals:       []origenDeEval{},
			plataformaFallida: &errorDeLaPlataformaEnElOrigen{
				Orden: ordenDeLasEjecucionesDelFlujo, Error: "la lista de ejecuciones del flujo no se puede leer",
			},
			lecturas: []string{lecturaDeLasEjecuciones},
		},
	}
}

// TestPlazoDeLaPlataformaDelContrato fija el plazo de lo que la decisión de un
// trabajo lee de la plataforma en los 10 minutos que dan el contrato
// (contracts/job-de-evals.md §2 de H30) y CONTRIBUTING.md: que el -timeout del
// paso de la decisión lo supere lo exige plazoDeLaDecision.
func TestPlazoDeLaPlataformaDelContrato(t *testing.T) {
	t.Parallel()

	assert.Equal(t, 10*time.Minute, plazoDeLaPlataforma)
}
