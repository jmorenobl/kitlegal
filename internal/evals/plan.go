package evals

import (
	"fmt"
	"regexp"
	"slices"
	"strings"
)

// formaDelModelo es la forma del id de un modelo, la misma que la del nombre de
// una skill: minúsculas, cifras y guiones. Con ella, el nombre de una sesión no
// lleva ningún separador de ruta ni ningún tabulador, y el plan se puede escribir
// como una tabla (contrato job-de-evals §3.2).
var formaDelModelo = regexp.MustCompile(`^[a-z0-9]+(-[a-z0-9]+)*$`)

// sufijoDeLaPruebaDeRed distingue en el nombre la sesión que lleva, además de la
// pregunta de su eval, el texto de la prueba de red (contrato job-de-evals §6).
const sufijoDeLaPruebaDeRed = "-prueba-de-red"

// sufijoDelModoHerramienta distingue en el nombre las sesiones del modo
// herramienta de las del modo orden de la misma eval y el mismo modelo
// (contracts/evals-en-dos-modos.md §2.1 de H21).
const sufijoDelModoHerramienta = "-herramienta"

// Modo es lo que una sesión de eval tiene para consultar (data-model §6 de H21;
// research.md D16 de H21). El valor vacío es ninguno: la sesión de una eval sin
// binario ni servidor, que no tiene kitlegal en el PATH ni el servidor
// declarado.
type Modo string

const (
	// ModoOrden es el de la sesión con kitlegal en el PATH y ningún servidor:
	// consulta con órdenes.
	ModoOrden Modo = "orden"

	// ModoHerramienta es el de la sesión con el servidor MCP de kitlegal
	// declarado y kitlegal fuera del PATH: consulta con herramientas.
	ModoHerramienta Modo = "herramienta"
)

// PlanDeEvals es lo que una ejecución del job tiene que abrir: cada eval bien
// formada, con el modelo que decide y con cada uno de los modelos informativos,
// repetida Repeticiones veces (ADR 0016; contrato job-de-evals §3.2), en cada
// uno de sus Modos (contracts/evals-en-dos-modos.md §2.1 de H21).
//
// Lo usan el guion, que abre las sesiones que da Sesiones, y EscribirInforme, que
// exige que cada serie de Series tenga exactamente Repeticiones sesiones: así el
// informe comprueba que se ejecutó lo planificado sin que el plan se escriba dos
// veces.
type PlanDeEvals struct {
	// Evals son las evals bien formadas con las que se juzga, en su orden.
	Evals []Eval

	// ModeloQueDecide es el modelo cuyas sesiones deciden el veredicto: el del
	// uso real de la skill.
	ModeloQueDecide string

	// ModelosInformativos son los modelos que se ejecutan y se publican como
	// límite inferior, sin decidir el veredicto, en su orden y sin repetir.
	// No ejecutan las evals informativas.
	ModelosInformativos []string

	// Repeticiones son las sesiones que se abren de cada eval con cada modelo.
	Repeticiones int

	// Modos son los modos en los que se abre cada eval de modo, que son todas
	// salvo las evals sin binario ni servidor. Vacío es solo ModoOrden: el plan
	// del sondeo. El job da los dos.
	Modos []Modo

	// PruebaDeRed añade, con el modelo que decide, una sesión del modo orden de
	// la primera eval de modo con el texto de la prueba de red. No se repite ni
	// forma serie: no mide la calidad de la skill sino la garantía de red
	// (contrato job-de-evals §6).
	PruebaDeRed bool
}

// SerieDeSesiones son las sesiones de una eval con un modelo en un modo: lo que
// el umbral juzga a la vez (data-model §10.4).
type SerieDeSesiones struct {
	// Eval es el Fichero de la eval.
	Eval string `json:"eval"`

	// Modelo es el id del modelo con el que se abren sus sesiones.
	Modelo string `json:"modelo"`

	// Modo es el modo en el que se abren; ninguno, en las de una eval sin
	// binario ni servidor.
	Modo Modo `json:"modo"`

	// Decide dice si la serie decide el veredicto: solo la del modelo que decide
	// sobre una eval que no es informativa.
	Decide bool `json:"decide"`
}

// SesionPlanificada es una sesión del plan: el directorio que el guion prepara y
// abre.
type SesionPlanificada struct {
	// Nombre es el del directorio de la sesión: <eval sin .yaml>[-prueba-de-red]-<modelo>-<nn>
	// en el modo orden y en las evals sin binario ni servidor, y
	// <eval sin .yaml>-herramienta-<modelo>-<nn> en el modo herramienta.
	Nombre string

	// Fichero es el de la eval con la que se juzga.
	Fichero string

	// Modelo es el id del modelo con el que se abre.
	Modelo string

	// Modo es el modo en el que se abre; ninguno, en la de una eval sin binario
	// ni servidor.
	Modo Modo

	// PruebaDeRed dice si la pregunta lleva además el texto de la prueba de red.
	PruebaDeRed bool
}

// Comprobar exige lo que el plan necesita para tener sentido: un modelo que
// decide con la forma de un id, modelos informativos con esa misma forma, sin
// repetir y distintos del que decide, y al menos una repetición. El error lo
// devuelven el guion, antes de la primera sesión, y EscribirInforme, sin escribir
// ningún informe.
func (p PlanDeEvals) Comprobar() error {
	if !formaDelModelo.MatchString(p.ModeloQueDecide) {
		return fmt.Errorf("el modelo que decide %q no tiene la forma de un id de modelo", p.ModeloQueDecide)
	}

	for posicion, modelo := range p.ModelosInformativos {
		if !formaDelModelo.MatchString(modelo) {
			return fmt.Errorf("los modelos informativos llevan %q, que no tiene la forma de un id de modelo", modelo)
		}

		if modelo == p.ModeloQueDecide {
			return fmt.Errorf("los modelos informativos llevan %q, que es el que decide", modelo)
		}

		if slices.Contains(p.ModelosInformativos[:posicion], modelo) {
			return fmt.Errorf("los modelos informativos llevan %q dos veces", modelo)
		}
	}

	if p.Repeticiones < 1 {
		return fmt.Errorf("las repeticiones son %d y hace falta al menos una", p.Repeticiones)
	}

	return nil
}

// modos son los modos en los que el plan abre cada eval de modo, en el orden en
// que se abren: el modo orden y el modo herramienta, los que estén en Modos; sin
// ninguno en Modos, solo el modo orden.
func (p PlanDeEvals) modos() []Modo {
	if len(p.Modos) == 0 {
		return []Modo{ModoOrden}
	}

	var modos []Modo

	for _, modo := range []Modo{ModoOrden, ModoHerramienta} {
		if slices.Contains(p.Modos, modo) {
			modos = append(modos, modo)
		}
	}

	return modos
}

// Series son las series del plan (contracts/evals-en-dos-modos.md §2.1 de H21):
// por cada modo, y en el orden de los modos, las de cada eval de modo, y al
// final, una sola vez y sin modo, las de cada eval sin binario ni servidor.
// Dentro de cada grupo van en orden de eval y, dentro de cada eval, con el
// modelo que decide delante de los informativos.
func (p PlanDeEvals) Series() []SerieDeSesiones {
	var series []SerieDeSesiones

	for _, modo := range p.modos() {
		for _, eval := range p.Evals {
			if !eval.SinBinarioNiServidor {
				series = append(series, p.seriesDe(eval, modo)...)
			}
		}
	}

	for _, eval := range p.Evals {
		if eval.SinBinarioNiServidor {
			series = append(series, p.seriesDe(eval, "")...)
		}
	}

	return series
}

// seriesDe son las series de la eval en el modo: la del modelo que decide y, si
// la eval no es informativa, la de cada uno de los modelos informativos. A las
// evals informativas no las abre ninguno de ellos: lo que miden es lo que el
// modelo del uso real hace con ellas, y el límite inferior es el de las evals
// que deciden (ADR 0016).
func (p PlanDeEvals) seriesDe(eval Eval, modo Modo) []SerieDeSesiones {
	series := []SerieDeSesiones{{Eval: eval.Fichero, Modelo: p.ModeloQueDecide, Modo: modo, Decide: !eval.Informativa}}
	if eval.Informativa {
		return series
	}

	for _, modelo := range p.ModelosInformativos {
		series = append(series, SerieDeSesiones{Eval: eval.Fichero, Modelo: modelo, Modo: modo})
	}

	return series
}

// Sesiones son las sesiones del plan, en el orden en que se abren: las
// Repeticiones de cada serie, numeradas desde 1, en el orden de Series, que las
// deja en tres tandas —las del modo orden, las del modo herramienta y las de las
// evals sin binario ni servidor—. Con PruebaDeRed, la tanda del modo orden
// termina con la sesión de la prueba de red de su primera eval con el modelo que
// decide; un plan sin ninguna serie del modo orden no la lleva. Los nombres se
// numeran con dos cifras para que el orden de nombre sea el de la repetición.
func (p PlanDeEvals) Sesiones() []SesionPlanificada {
	series := p.Series()

	// Las series del modo orden van delante de todas las demás.
	delModoOrden := 0
	for delModoOrden < len(series) && series[delModoOrden].Modo == ModoOrden {
		delModoOrden++
	}

	sesiones := p.sesionesDe(series[:delModoOrden])

	if p.PruebaDeRed && delModoOrden > 0 {
		fichero := series[0].Eval
		sesiones = append(sesiones, SesionPlanificada{
			Nombre:      nombreDeSesion(fichero, p.ModeloQueDecide, ModoOrden, 1, true),
			Fichero:     fichero,
			Modelo:      p.ModeloQueDecide,
			Modo:        ModoOrden,
			PruebaDeRed: true,
		})
	}

	return append(sesiones, p.sesionesDe(series[delModoOrden:])...)
}

// sesionesDe son las Repeticiones de cada una de las series, en su orden.
func (p PlanDeEvals) sesionesDe(series []SerieDeSesiones) []SesionPlanificada {
	var sesiones []SesionPlanificada

	for _, serie := range series {
		for numero := 1; numero <= p.Repeticiones; numero++ {
			sesiones = append(sesiones, SesionPlanificada{
				Nombre:  nombreDeSesion(serie.Eval, serie.Modelo, serie.Modo, numero, false),
				Fichero: serie.Eval,
				Modelo:  serie.Modelo,
				Modo:    serie.Modo,
			})
		}
	}

	return sesiones
}

// nombreDeSesion compone el nombre del directorio de una sesión: el de siempre
// en el modo orden y en las evals sin binario ni servidor, y con el modo detrás
// de la eval en el modo herramienta.
func nombreDeSesion(fichero, modelo string, modo Modo, numero int, pruebaDeRed bool) string {
	nombre := strings.TrimSuffix(fichero, ".yaml")
	if pruebaDeRed {
		nombre += sufijoDeLaPruebaDeRed
	}

	if modo == ModoHerramienta {
		nombre += sufijoDelModoHerramienta
	}

	return fmt.Sprintf("%s-%s-%02d", nombre, modelo, numero)
}
