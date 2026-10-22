package evals

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"slices"
	"strings"
	"testing"
	"testing/fstest"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// Lo que comparten las ejecuciones sintéticas de TestComprobarLaSuiteCompleta:
// el commit etiquetado y la etiqueta, los de contracts/release.md §2 de H30, y
// otro commit, el de una ejecución que la plataforma da sin ser del etiquetado.
const (
	commitEtiquetado    = "0f3c2a9e8d7b6c5a4f3e2d1c0b9a8f7e6d5c4b3a"
	etiquetaDeLaRelease = "v0.9.0"
	otroCommitDelFlujo  = "9e8d7c6b5a4f3e2d1c0b9a8f7e6d5c4b3a2f1e0d"
)

// instanteDeLaRelease es el instante desde el que se fechan las ejecuciones
// sintéticas.
var instanteDeLaRelease = time.Date(2026, 10, 22, 12, 0, 0, 0, time.UTC)

// evalsDelCommitEtiquetado son las evals de cada skill con evals en el commit
// etiquetado, como las da evalsDelCommit.
var evalsDelCommitEtiquetado = []evalsDeUnaSkill{
	{
		skill:    "boe-legislacion",
		ficheros: []string{"01-articulo.yaml", "02-indice.yaml", "03-vigencia.yaml", "04-derogada.yaml"},
	},
	{skill: "legal-core", ficheros: []string{"01-territorio.yaml", "02-plazo.yaml"}},
}

// errLecturaDeOtraComprobacion es el fallo de las lecturas de la plataforma en
// memoria de la release que la comprobación de la suite completa no hace.
var errLecturaDeOtraComprobacion = errors.New("la comprobación de la suite completa no hace esta lectura")

// ejecucionDeLaRelease es una ejecución del flujo evals de la plataforma en
// memoria de la release: la ejecución, sus trabajos y el informe que imprimió el
// trabajo de cada skill, por skill. El registro del trabajo de una skill sin
// informe no lo trae.
type ejecucionDeLaRelease struct {
	ejecucion ejecucionDelFlujo
	trabajos  []trabajoDeLaEjecucion
	informes  map[string]*Informe
}

// plataformaDeLaRelease son las lecturas de la plataforma de la comprobación de
// la suite completa con respuestas sintéticas con la forma de las de gh: las
// ejecuciones sobre el commit, que da todas, también las de otro headSha, como
// si gh las diera; los trabajos de cada una; y el informe de cada trabajo,
// sacado de su registro con informeDelRegistro, como lo saca gh. Anota cada
// lectura, en orden, y da el fallo de la que lo tiene.
type plataformaDeLaRelease struct {
	ejecuciones []ejecucionDeLaRelease
	fallos      map[string]error
	lecturas    []string
}

// La plataforma en memoria de la release implementa las lecturas de la
// plataforma.
var _ lecturasDeLaPlataforma = (*plataformaDeLaRelease)(nil)

// leer anota la lectura y da su fallo, si lo tiene.
func (p *plataformaDeLaRelease) leer(lectura string) error {
	p.lecturas = append(p.lecturas, lectura)

	return p.fallos[lectura]
}

func (p *plataformaDeLaRelease) ficherosDeLaPropuesta(context.Context, int) ([]cambioDeLaPropuesta, error) {
	return nil, errLecturaDeOtraComprobacion
}

func (p *plataformaDeLaRelease) propuestasFusionadas(context.Context) ([]propuestaFusionada, error) {
	return nil, errLecturaDeOtraComprobacion
}

func (p *plataformaDeLaRelease) ejecucionesDelFlujo(context.Context) ([]ejecucionDelFlujo, error) {
	return nil, errLecturaDeOtraComprobacion
}

func (p *plataformaDeLaRelease) ejecucionesDelCommit(_ context.Context, commit string) ([]ejecucionDelFlujo, error) {
	if err := p.leer(lecturaDeLasEjecucionesDelCommit(commit)); err != nil {
		return nil, err
	}

	ejecuciones := make([]ejecucionDelFlujo, 0, len(p.ejecuciones))
	for _, enMemoria := range p.ejecuciones {
		ejecuciones = append(ejecuciones, enMemoria.ejecucion)
	}

	return ejecuciones, nil
}

func (p *plataformaDeLaRelease) trabajos(_ context.Context, ejecucion int64) ([]trabajoDeLaEjecucion, error) {
	if err := p.leer(lecturaDeLosTrabajos(ejecucion)); err != nil {
		return nil, err
	}

	enMemoria, err := p.ejecucionEnMemoria(ejecucion)
	if err != nil {
		return nil, err
	}

	return enMemoria.trabajos, nil
}

func (p *plataformaDeLaRelease) informeDelTrabajo(_ context.Context, ejecucion, trabajo int64) ([]byte, error) {
	if err := p.leer(lecturaDelInforme(ejecucion, trabajo)); err != nil {
		return nil, err
	}

	enMemoria, err := p.ejecucionEnMemoria(ejecucion)
	if err != nil {
		return nil, err
	}

	indice := slices.IndexFunc(enMemoria.trabajos, func(deLaEjecucion trabajoDeLaEjecucion) bool {
		return deLaEjecucion.ID == trabajo
	})
	if indice < 0 {
		return nil, fmt.Errorf("la plataforma en memoria no tiene el trabajo %d de la ejecución %d", trabajo, ejecucion)
	}

	skill := strings.TrimSuffix(strings.TrimPrefix(enMemoria.trabajos[indice].Name, prefijoDeUnTrabajoDeEvals),
		sufijoDeUnTrabajoDeEvals)
	registro := "Run scripts/evals.sh\n"

	if informe, ok := enMemoria.informes[skill]; ok {
		contenido, err := json.Marshal(informe)
		if err != nil {
			return nil, err
		}

		registro += marcaDeInicioDelInforme + "\n" + string(contenido) + "\n" + marcaDeFinDelInforme + "\n"
	}

	informe, err := informeDelRegistro([]byte(registro))
	if err != nil {
		return nil, fmt.Errorf("el informe del trabajo %d de la ejecución %d: %w", trabajo, ejecucion, err)
	}

	return informe, nil
}

// ejecucionEnMemoria es la ejecución de la plataforma en memoria con ese id.
func (p *plataformaDeLaRelease) ejecucionEnMemoria(ejecucion int64) (ejecucionDeLaRelease, error) {
	for _, enMemoria := range p.ejecuciones {
		if enMemoria.ejecucion.ID == ejecucion {
			return enMemoria, nil
		}
	}

	return ejecucionDeLaRelease{}, fmt.Errorf("la plataforma en memoria no tiene la ejecución %d", ejecucion)
}

// lecturaDeLasEjecucionesDelCommit es como anota la plataforma en memoria la
// lectura de las ejecuciones del flujo sobre un commit.
func lecturaDeLasEjecucionesDelCommit(commit string) string {
	return "ejecuciones del commit " + commit
}

// cambioDeLaEjecucion cambia una ejecución sintética de la release.
type cambioDeLaEjecucion func(*ejecucionDeLaRelease)

// ejecucionCompleta es una ejecución del flujo evals sobre el commit etiquetado,
// creada el tiempo dado antes de instanteDeLaRelease, terminada en verde, con el
// trabajo tanda y el de cada skill de evalsDelCommitEtiquetado en success, y el
// informe de cada skill de la suite completa, con cada eval suya medida; con los
// cambios dados encima, en orden. El trabajo de la skill en la posición i de
// evalsDelCommitEtiquetado es el id*10+i+1.
func ejecucionCompleta(id int64, antes time.Duration, cambios ...cambioDeLaEjecucion) ejecucionDeLaRelease {
	enMemoria := ejecucionDeLaRelease{
		ejecucion: ejecucionDelFlujo{
			ID: id, Commit: commitEtiquetado, Status: statusTerminado, Conclusion: conclusionDeExito,
			Creada: instanteDeLaRelease.Add(-antes),
		},
		trabajos: []trabajoDeLaEjecucion{{
			ID: id * 10, Name: trabajoDeLaTanda, Status: statusTerminado, Conclusion: conclusionDeExito,
			Steps: []pasoDeLaEjecucion{
				{Name: marcaDeLaTanda, Conclusion: conclusionDeExito},
				{Name: marcaDeLaSuiteCompleta, Conclusion: conclusionDeExito},
			},
		}},
		informes: map[string]*Informe{},
	}

	for posicion, deLaSkill := range evalsDelCommitEtiquetado {
		enMemoria.trabajos = append(enMemoria.trabajos, trabajoDeLaEjecucion{
			ID: id*10 + int64(posicion) + 1, Name: nombreDelTrabajoDeLaSkill(deLaSkill.skill), Status: statusTerminado,
			Conclusion: conclusionDeExito,
			Steps:      []pasoDeLaEjecucion{{Name: marcaDelInformeEscrito, Conclusion: conclusionDeExito}},
		})

		origen := &origenDeLasEvals{Modo: origenDeLaSuiteCompleta, EvalsMedidas: len(deLaSkill.ficheros)}
		for _, fichero := range deLaSkill.ficheros {
			origen.Evals = append(origen.Evals, origenDeEval{
				Eval: fichero, Huella: strings.Repeat("a", 64), Medida: true, Motivo: motivoSuiteCompleta,
				Aportes: []aporteDeLaEval{},
			})
		}

		enMemoria.informes[deLaSkill.skill] = &Informe{Skill: deLaSkill.skill, Commit: commitEtiquetado, Origen: origen}
	}

	for _, cambio := range cambios {
		cambio(&enMemoria)
	}

	return enMemoria
}

// deOtroCommit pone la ejecución sobre otro commit.
func deOtroCommit(enMemoria *ejecucionDeLaRelease) {
	enMemoria.ejecucion.Commit = otroCommitDelFlujo
}

// ejecucionSinTerminar deja la ejecución en curso, sin conclusión.
func ejecucionSinTerminar(enMemoria *ejecucionDeLaRelease) {
	enMemoria.ejecucion.Status, enMemoria.ejecucion.Conclusion = "in_progress", ""
}

// conElTrabajo pone el trabajo de la skill con ese estado y esa conclusión.
func conElTrabajo(skill, estado, conclusion string) cambioDeLaEjecucion {
	return func(enMemoria *ejecucionDeLaRelease) {
		for posicion := range enMemoria.trabajos {
			if trabajo := &enMemoria.trabajos[posicion]; trabajo.Name == nombreDelTrabajoDeLaSkill(skill) {
				trabajo.Status, trabajo.Conclusion = estado, conclusion
			}
		}
	}
}

// sinElTrabajo quita de la ejecución el trabajo de la skill.
func sinElTrabajo(skill string) cambioDeLaEjecucion {
	return func(enMemoria *ejecucionDeLaRelease) {
		enMemoria.trabajos = slices.DeleteFunc(enMemoria.trabajos, func(trabajo trabajoDeLaEjecucion) bool {
			return trabajo.Name == nombreDelTrabajoDeLaSkill(skill)
		})
	}
}

// sinElInforme deja el registro del trabajo de la skill sin su informe.
func sinElInforme(skill string) cambioDeLaEjecucion {
	return func(enMemoria *ejecucionDeLaRelease) {
		delete(enMemoria.informes, skill)
	}
}

// cambiandoElInforme cambia el informe del trabajo de la skill.
func cambiandoElInforme(skill string, cambio func(*Informe)) cambioDeLaEjecucion {
	return func(enMemoria *ejecucionDeLaRelease) {
		cambio(enMemoria.informes[skill])
	}
}

// reutilizando marca reutilizadas las evals dadas del origen del informe, con
// la procedencia de una medida de main, y deja los recuentos al día.
func reutilizando(ficheros ...string) func(*Informe) {
	return func(informe *Informe) {
		for posicion := range informe.Origen.Evals {
			if eval := &informe.Origen.Evals[posicion]; slices.Contains(ficheros, eval.Eval) {
				eval.Medida, eval.Motivo = false, ""
				eval.Procedencia = &procedenciaDeLaMedida{
					Ejecucion: 18700000001, Trabajo: 18700000001*10 + 1, Commit: otroCommitDelFlujo, Propuesta: 183,
					Fecha: instanteDeLaRelease.Add(-48 * time.Hour).Format(time.RFC3339),
				}
			}
		}

		informe.Origen.EvalsReutilizadas = len(informe.Origen.reutilizadas())
		informe.Origen.EvalsMedidas = len(informe.Origen.Evals) - informe.Origen.EvalsReutilizadas
	}
}

// conModo pone el modo del origen del informe.
func conModo(modo modoDelOrigen) func(*Informe) {
	return func(informe *Informe) {
		informe.Origen.Modo = modo
	}
}

// sinLaEvalEnElOrigen quita del origen del informe la entrada de la eval.
func sinLaEvalEnElOrigen(fichero string) func(*Informe) {
	return func(informe *Informe) {
		informe.Origen.Evals = slices.DeleteFunc(informe.Origen.Evals, func(eval origenDeEval) bool {
			return eval.Eval == fichero
		})
		informe.Origen.EvalsMedidas = len(informe.Origen.Evals)
	}
}

// sinOrigen quita del informe el bloque origen, como en uno anterior a H30.
func sinOrigen(informe *Informe) {
	informe.Origen = nil
}

// mensajeDeLaSuiteIncompleta es el mensaje de contracts/release.md §2 de H30
// con el commit etiquetado, la etiqueta y las líneas dadas.
func mensajeDeLaSuiteIncompleta(lineas ...string) string {
	return "release: el commit " + commitEtiquetado + " no tiene una ejecución del job de evals de la suite completa, " +
		"en verde y sin evals reutilizadas.\n" + strings.Join(lineas, "\n") + "\n" +
		"Lanza la suite completa sobre la etiqueta y, cuando termine en verde, vuelve a ejecutar la release:\n" +
		"  gh workflow run evals.yml --ref " + etiquetaDeLaRelease
}

// mensajeDelFalloDeLaPlataforma es el del error de la comprobación con el
// commit etiquetado cuando una lectura de la plataforma falla con el fallo dado.
func mensajeDelFalloDeLaPlataforma(fallo error) string {
	return "release: la comprobación de la suite completa del commit " + commitEtiquetado +
		" no pudo leer la plataforma: " + fallo.Error()
}

// casoDeLaSuiteCompleta es un caso de TestComprobarLaSuiteCompleta: las
// ejecuciones de la plataforma en memoria, en el orden en que las da, y los
// fallos de sus lecturas; y lo que tiene que dar la comprobación: la ejecución
// que vale, o el mensaje de su error y, si es de la plataforma, el fallo que
// envuelve.
type casoDeLaSuiteCompleta struct {
	nombre      string
	ejecuciones []ejecucionDeLaRelease
	fallos      map[string]error
	vale        int64
	mensaje     string
	fallo       error
	sinLeer     []string
}

// falloDeGh es el de una orden de gh que no terminó con 0, como lo da
// ejecutarGh.
func falloDeGh(orden string, variables ...string) error {
	return &errorDeUnaOrden{
		orden: orden, variables: variables, causa: errors.New("exit status 1"),
		salidaDeError: "HTTP 502: Bad Gateway\n",
	}
}

// casosDeLaSuiteCompleta son los casos de TestComprobarLaSuiteCompleta.
func casosDeLaSuiteCompleta() []casoDeLaSuiteCompleta {
	trabajosDe1001 := lecturaDeLosTrabajos(1001)
	falloDeLosTrabajos := falloDeGh(ordenDeLosTrabajos, variableDeLaEjecucionAnterior+"=1001")
	falloDelInforme := falloDeGh(ordenDelRegistroDeUnTrabajo, variableDeLaEjecucion+"=1001", variableDelTrabajo+"=10012")
	falloDeLaLista := falloDeGh(ordenDeLasEjecucionesDeUnCommit, variableDelCommitEvaluado+"="+commitEtiquetado)

	return []casoDeLaSuiteCompleta{
		{
			nombre:  "sin-ejecucion",
			mensaje: mensajeDeLaSuiteIncompleta("- ninguna ejecución del flujo evals sobre el commit"),
		},
		{
			// Las dos líneas de contracts/release.md §2 de H30, con las ejecuciones
			// dadas de la más antigua a la más reciente.
			nombre: "el-mensaje-del-contrato",
			ejecuciones: []ejecucionDeLaRelease{
				ejecucionCompleta(18733990002, 3*time.Hour, cambiandoElInforme("boe-legislacion", func(informe *Informe) {
					conModo(origenConReutilizacion)(informe)
					reutilizando("02-indice.yaml", "03-vigencia.yaml", "04-derogada.yaml")(informe)
				})),
				ejecucionCompleta(18734029311, time.Hour, conElTrabajo("legal-core", statusTerminado, conclusionDeFallo)),
			},
			mensaje: mensajeDeLaSuiteIncompleta(
				"- ejecución 18734029311: el trabajo evals (legal-core) terminó en failure",
				"- ejecución 18733990002: el informe de boe-legislacion es de una ejecución con reutilización "+
					"(3 evals reutilizadas)",
			),
		},
		{
			nombre:      "con-otro-headsha",
			ejecuciones: []ejecucionDeLaRelease{ejecucionCompleta(1001, time.Hour, deOtroCommit)},
			mensaje: mensajeDeLaSuiteIncompleta(
				"- ejecución 1001: su headSha es " + otroCommitDelFlujo + ", no el commit"),
			sinLeer: []string{trabajosDe1001},
		},
		{
			nombre:      "sin-terminar",
			ejecuciones: []ejecucionDeLaRelease{ejecucionCompleta(1001, time.Hour, ejecucionSinTerminar)},
			mensaje:     mensajeDeLaSuiteIncompleta("- ejecución 1001: no ha terminado (status in_progress)"),
			sinLeer:     []string{trabajosDe1001},
		},
		{
			nombre: "con-un-trabajo-en-failure",
			ejecuciones: []ejecucionDeLaRelease{
				ejecucionCompleta(1001, time.Hour, conElTrabajo("legal-core", statusTerminado, conclusionDeFallo)),
			},
			mensaje: mensajeDeLaSuiteIncompleta("- ejecución 1001: el trabajo evals (legal-core) terminó en failure"),
			sinLeer: []string{lecturaDelInforme(1001, 10011)},
		},
		{
			nombre: "con-un-trabajo-saltado",
			ejecuciones: []ejecucionDeLaRelease{
				ejecucionCompleta(1001, time.Hour, conElTrabajo("boe-legislacion", statusTerminado, "skipped")),
			},
			mensaje: mensajeDeLaSuiteIncompleta("- ejecución 1001: el trabajo evals (boe-legislacion) terminó en skipped"),
		},
		{
			nombre: "con-un-trabajo-sin-terminar",
			ejecuciones: []ejecucionDeLaRelease{
				ejecucionCompleta(1001, time.Hour, conElTrabajo("legal-core", "in_progress", "")),
			},
			mensaje: mensajeDeLaSuiteIncompleta(
				"- ejecución 1001: el trabajo evals (legal-core) no ha terminado (status in_progress)"),
		},
		{
			nombre:      "sin-el-trabajo-de-una-skill",
			ejecuciones: []ejecucionDeLaRelease{ejecucionCompleta(1001, time.Hour, sinElTrabajo("legal-core"))},
			mensaje:     mensajeDeLaSuiteIncompleta("- ejecución 1001: no tiene el trabajo evals (legal-core)"),
		},
		{
			nombre:      "con-un-informe-que-falta",
			ejecuciones: []ejecucionDeLaRelease{ejecucionCompleta(1001, time.Hour, sinElInforme("legal-core"))},
			mensaje: mensajeDeLaSuiteIncompleta(
				"- ejecución 1001: el registro del trabajo evals (legal-core) no trae su informe"),
		},
		{
			nombre:      "con-un-informe-sin-origen",
			ejecuciones: []ejecucionDeLaRelease{ejecucionCompleta(1001, time.Hour, cambiandoElInforme("legal-core", sinOrigen))},
			mensaje: mensajeDeLaSuiteIncompleta(
				"- ejecución 1001: el informe de legal-core no dice cómo midió sus evals: no tiene el bloque origen"),
		},
		{
			nombre: "con-una-eval-reutilizada",
			ejecuciones: []ejecucionDeLaRelease{
				ejecucionCompleta(1001, time.Hour, cambiandoElInforme("boe-legislacion", reutilizando("02-indice.yaml"))),
			},
			mensaje: mensajeDeLaSuiteIncompleta(
				"- ejecución 1001: el informe de boe-legislacion reutiliza la eval 02-indice.yaml"),
		},
		{
			nombre: "con-modo-reutilizacion",
			ejecuciones: []ejecucionDeLaRelease{
				ejecucionCompleta(1001, time.Hour, cambiandoElInforme("legal-core", func(informe *Informe) {
					conModo(origenConReutilizacion)(informe)
					reutilizando("02-plazo.yaml")(informe)
				})),
			},
			mensaje: mensajeDeLaSuiteIncompleta(
				"- ejecución 1001: el informe de legal-core es de una ejecución con reutilización (1 eval reutilizada)"),
		},
		{
			// Con reutilización y todas medidas, porque la propuesta las eligió
			// todas: tampoco es la suite completa.
			nombre: "con-modo-reutilizacion-y-todas-medidas",
			ejecuciones: []ejecucionDeLaRelease{
				ejecucionCompleta(1001, time.Hour, cambiandoElInforme("legal-core", conModo(origenConReutilizacion))),
			},
			mensaje: mensajeDeLaSuiteIncompleta(
				"- ejecución 1001: el informe de legal-core es de una ejecución con reutilización (0 evals reutilizadas)"),
		},
		{
			nombre: "con-otro-modo",
			ejecuciones: []ejecucionDeLaRelease{
				ejecucionCompleta(1001, time.Hour, cambiandoElInforme("legal-core", conModo("parcial"))),
			},
			mensaje: mensajeDeLaSuiteIncompleta(
				`- ejecución 1001: el informe de legal-core es de una ejecución de modo "parcial", no de la suite completa`),
		},
		{
			nombre: "con-una-eval-del-commit-que-no-esta-en-el-informe",
			ejecuciones: []ejecucionDeLaRelease{
				ejecucionCompleta(1001, time.Hour, cambiandoElInforme("legal-core", sinLaEvalEnElOrigen("02-plazo.yaml"))),
			},
			mensaje: mensajeDeLaSuiteIncompleta(
				"- ejecución 1001: el informe de legal-core no tiene la eval 02-plazo.yaml del commit"),
		},
		{
			nombre:      "completa-y-en-verde",
			ejecuciones: []ejecucionDeLaRelease{ejecucionCompleta(1001, time.Hour)},
			vale:        1001,
		},
		{
			// Vale la primera que vale, de la más reciente a la más antigua: la más
			// antigua de las dos completas ni se lee.
			nombre: "la-mas-reciente-de-las-que-valen",
			ejecuciones: []ejecucionDeLaRelease{
				ejecucionCompleta(1001, 3*time.Hour),
				ejecucionCompleta(1003, time.Hour, conElTrabajo("legal-core", statusTerminado, conclusionDeFallo)),
				ejecucionCompleta(1002, 2*time.Hour),
			},
			vale:    1002,
			sinLeer: []string{trabajosDe1001},
		},
		{
			// A la misma creación, la de número mayor va antes.
			nombre: "a-la-misma-creacion-por-numero",
			ejecuciones: []ejecucionDeLaRelease{
				ejecucionCompleta(1001, time.Hour, ejecucionSinTerminar),
				ejecucionCompleta(1002, time.Hour, deOtroCommit),
			},
			mensaje: mensajeDeLaSuiteIncompleta(
				"- ejecución 1002: su headSha es "+otroCommitDelFlujo+", no el commit",
				"- ejecución 1001: no ha terminado (status in_progress)",
			),
		},
		{
			nombre:  "la-lista-de-ejecuciones-falla",
			fallos:  map[string]error{lecturaDeLasEjecucionesDelCommit(commitEtiquetado): falloDeLaLista},
			mensaje: mensajeDelFalloDeLaPlataforma(falloDeLaLista),
			fallo:   falloDeLaLista,
		},
		{
			nombre:      "los-trabajos-de-una-ejecucion-fallan",
			ejecuciones: []ejecucionDeLaRelease{ejecucionCompleta(1001, time.Hour)},
			fallos:      map[string]error{trabajosDe1001: falloDeLosTrabajos},
			mensaje:     mensajeDelFalloDeLaPlataforma(falloDeLosTrabajos),
			fallo:       falloDeLosTrabajos,
		},
		{
			nombre:      "el-registro-de-un-trabajo-falla",
			ejecuciones: []ejecucionDeLaRelease{ejecucionCompleta(1001, time.Hour)},
			fallos:      map[string]error{lecturaDelInforme(1001, 10012): falloDelInforme},
			mensaje:     mensajeDelFalloDeLaPlataforma(falloDelInforme),
			fallo:       falloDelInforme,
		},
	}
}

// TestComprobarLaSuiteCompleta fija la comprobación de la release (FR-025 y
// SC-005 de H30; contracts/release.md §2 y §3 de H30; data-model §11 de H30)
// con la plataforma en memoria: falla sin ejecución, con otro headSha, sin
// terminar, con un trabajo de skill que no está o no terminó en success, con un
// informe que falta o sin origen, con una eval reutilizada, con modo
// reutilizacion o cualquier otro que no sea completa y con una eval del commit
// que no está en el informe, cada uno con su línea en el mensaje; sale bien con
// la más reciente que es completa y en verde, sin leer las anteriores; y con un
// fallo de la plataforma falla con él, sin decir que se lance la suite completa.
func TestComprobarLaSuiteCompleta(t *testing.T) {
	t.Parallel()

	for _, caso := range casosDeLaSuiteCompleta() {
		t.Run(caso.nombre, func(t *testing.T) {
			t.Parallel()

			comprobarElCasoDeLaSuiteCompleta(t, caso)
		})
	}
}

// comprobarElCasoDeLaSuiteCompleta comprueba un caso de
// TestComprobarLaSuiteCompleta.
func comprobarElCasoDeLaSuiteCompleta(t *testing.T, caso casoDeLaSuiteCompleta) {
	t.Helper()

	plataforma := &plataformaDeLaRelease{ejecuciones: caso.ejecuciones, fallos: caso.fallos}

	vale, err := comprobarLaSuiteCompleta(t.Context(), plataforma, commitEtiquetado, etiquetaDeLaRelease,
		evalsDelCommitEtiquetado)

	if caso.mensaje == "" {
		require.NoError(t, err)
		assert.Equal(t, caso.vale, vale)
	} else {
		require.EqualError(t, err, caso.mensaje)
		assert.Zero(t, vale)
	}

	if caso.fallo != nil {
		require.ErrorIs(t, err, caso.fallo)
		assert.NotContains(t, err.Error(), "gh workflow run", "un fallo de la plataforma no pide lanzar la suite")
	}

	for _, lectura := range caso.sinLeer {
		assert.NotContains(t, plataforma.lecturas, lectura)
	}
}

// TestComprobarLaSuiteCompletaLeeLoQueNecesita fija lo que lee la comprobación
// de una ejecución completa: la lista del commit etiquetado, los trabajos de la
// ejecución y el informe del trabajo de cada skill con evals, en el orden de las
// skills, y nada más.
func TestComprobarLaSuiteCompletaLeeLoQueNecesita(t *testing.T) {
	t.Parallel()

	plataforma := &plataformaDeLaRelease{ejecuciones: []ejecucionDeLaRelease{ejecucionCompleta(1001, time.Hour)}}

	vale, err := comprobarLaSuiteCompleta(t.Context(), plataforma, commitEtiquetado, etiquetaDeLaRelease,
		evalsDelCommitEtiquetado)
	require.NoError(t, err)
	assert.Equal(t, int64(1001), vale)
	assert.Equal(t, []string{
		lecturaDeLasEjecucionesDelCommit(commitEtiquetado),
		lecturaDeLosTrabajos(1001),
		lecturaDelInforme(1001, 10011),
		lecturaDelInforme(1001, 10012),
	}, plataforma.lecturas)
}

// TestComprobarLaSuiteCompletaConElPlazoAgotado fija que un informe que no se
// lee porque el contexto se agotó es un fallo de la plataforma, no un informe
// que falta.
func TestComprobarLaSuiteCompletaConElPlazoAgotado(t *testing.T) {
	t.Parallel()

	ctx, cancelar := context.WithCancel(t.Context())
	cancelar()

	plataforma := &plataformaDeLaRelease{
		ejecuciones: []ejecucionDeLaRelease{ejecucionCompleta(1001, time.Hour)},
		fallos:      map[string]error{lecturaDelInforme(1001, 10011): errors.New("signal: killed")},
	}

	_, err := comprobarLaSuiteCompleta(ctx, plataforma, commitEtiquetado, etiquetaDeLaRelease, evalsDelCommitEtiquetado)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "no pudo leer la plataforma: signal: killed")
	assert.NotContains(t, err.Error(), "gh workflow run")
}

// TestEvalsDelCommit fija las evals de cada skill con evals en el árbol de un
// commit: los subdirectorios de evals/ con algún fichero de eval, en orden, y
// sus ficheros de eval, en orden, sin la carpeta del juez, la lista de
// expresiones prohibidas ni otro fichero; un subdirectorio sin evals no es una
// skill con evals. Sin evals/, el error que lo nombra.
func TestEvalsDelCommit(t *testing.T) {
	t.Parallel()

	arbol := fstest.MapFS{
		"evals/legal-core/02-plazo.yaml":                       {},
		"evals/legal-core/01-territorio.yaml":                  {},
		"evals/boe-legislacion/01-articulo.yaml":               {},
		"evals/boe-legislacion/expresiones-prohibidas.yaml":    {},
		"evals/boe-legislacion/juez/clases.yaml":               {},
		"evals/boe-legislacion/LEEME.md":                       {},
		"evals/boe-legislacion/02-indice.yaml":                 {},
		"evals/sin-evals/LEEME.md":                             {},
		"evals/notas.yaml":                                     {},
		"evals/jurisprudencia/01-sentencia-sin-documento.yaml": {},
	}

	evals, err := evalsDelCommit(arbol)
	require.NoError(t, err)
	assert.Equal(t, []evalsDeUnaSkill{
		{skill: "boe-legislacion", ficheros: []string{"01-articulo.yaml", "02-indice.yaml"}},
		{skill: "jurisprudencia", ficheros: []string{"01-sentencia-sin-documento.yaml"}},
		{skill: "legal-core", ficheros: []string{"01-territorio.yaml", "02-plazo.yaml"}},
	}, evals)

	_, err = evalsDelCommit(fstest.MapFS{"skills/legal-core/SKILL.md": {}})
	require.ErrorContains(t, err, "evals/ no se puede listar")
}
