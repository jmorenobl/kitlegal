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

// PlanDeEvals es lo que una ejecución del job tiene que abrir: cada eval bien
// formada, con el modelo que decide y con cada uno de los modelos informativos,
// repetida Repeticiones veces (ADR 0016; contrato job-de-evals §3.2).
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

	// PruebaDeRed añade, con el modelo que decide, una sesión de la primera eval
	// con el texto de la prueba de red. No se repite ni forma serie: no mide la
	// calidad de la skill sino la garantía de red (contrato job-de-evals §6).
	PruebaDeRed bool
}

// SerieDeSesiones son las sesiones de una eval con un modelo: lo que el umbral
// juzga a la vez (data-model §10.4).
type SerieDeSesiones struct {
	// Eval es el Fichero de la eval.
	Eval string `json:"eval"`

	// Modelo es el id del modelo con el que se abren sus sesiones.
	Modelo string `json:"modelo"`

	// Decide dice si la serie decide el veredicto: solo la del modelo que decide
	// sobre una eval que no es informativa.
	Decide bool `json:"decide"`
}

// SesionPlanificada es una sesión del plan: el directorio que el guion prepara y
// abre.
type SesionPlanificada struct {
	// Nombre es el del directorio de la sesión, <eval sin .yaml>[-prueba-de-red]-<modelo>-<nn>.
	Nombre string

	// Fichero es el de la eval con la que se juzga.
	Fichero string

	// Modelo es el id del modelo con el que se abre.
	Modelo string

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

// Series son las series del plan, en orden de eval y, dentro de cada eval, con el
// modelo que decide delante de los informativos: la de cada eval con el modelo
// que decide, y la de cada eval que no es informativa con cada uno de los modelos
// informativos. A las evals informativas no las abre ninguno de ellos: lo que
// miden es lo que el modelo del uso real hace con ellas, y el límite inferior es
// el de las evals que deciden (ADR 0016).
func (p PlanDeEvals) Series() []SerieDeSesiones {
	var series []SerieDeSesiones

	for _, eval := range p.Evals {
		series = append(series, SerieDeSesiones{
			Eval: eval.Fichero, Modelo: p.ModeloQueDecide, Decide: !eval.Informativa,
		})

		if eval.Informativa {
			continue
		}

		for _, modelo := range p.ModelosInformativos {
			series = append(series, SerieDeSesiones{Eval: eval.Fichero, Modelo: modelo})
		}
	}

	return series
}

// Sesiones son las sesiones del plan, en el orden en que el guion las abre: las
// Repeticiones de cada serie, numeradas desde 1, y al final, con PruebaDeRed, la
// sesión de la prueba de red de la primera eval con el modelo que decide. Los
// nombres se numeran con dos cifras para que el orden de nombre sea el de la
// repetición.
func (p PlanDeEvals) Sesiones() []SesionPlanificada {
	var sesiones []SesionPlanificada

	for _, serie := range p.Series() {
		for numero := 1; numero <= p.Repeticiones; numero++ {
			sesiones = append(sesiones, SesionPlanificada{
				Nombre:  nombreDeSesion(serie.Eval, serie.Modelo, numero, false),
				Fichero: serie.Eval,
				Modelo:  serie.Modelo,
			})
		}
	}

	if p.PruebaDeRed && len(p.Evals) > 0 {
		fichero := p.Evals[0].Fichero
		sesiones = append(sesiones, SesionPlanificada{
			Nombre:      nombreDeSesion(fichero, p.ModeloQueDecide, 1, true),
			Fichero:     fichero,
			Modelo:      p.ModeloQueDecide,
			PruebaDeRed: true,
		})
	}

	return sesiones
}

// nombreDeSesion compone el nombre del directorio de una sesión.
func nombreDeSesion(fichero, modelo string, numero int, pruebaDeRed bool) string {
	nombre := strings.TrimSuffix(fichero, ".yaml")
	if pruebaDeRed {
		nombre += sufijoDeLaPruebaDeRed
	}

	return fmt.Sprintf("%s-%s-%02d", nombre, modelo, numero)
}
