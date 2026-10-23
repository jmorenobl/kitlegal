package evals

import (
	"cmp"
	"errors"
	"maps"
	"slices"
	"strconv"
	"strings"
)

// cambioDeLaPropuesta es un fichero de la propuesta, como lo da la plataforma
// (FR-007; data-model §5): su ruta y, si se renombró, la de antes.
type cambioDeLaPropuesta struct {
	ruta     string
	anterior string
}

// rutas son las rutas cambiadas del fichero: la suya y, en un renombrado, la
// de antes, que cuentan las dos (FR-007).
func (c cambioDeLaPropuesta) rutas() []string {
	if c.anterior == "" || c.anterior == c.ruta {
		return []string{c.ruta}
	}

	return []string{c.ruta, c.anterior}
}

// motivoDeLaEleccion es por qué se elige una eval (FR-008; research D19).
type motivoDeLaEleccion string

// Los motivos de la elección (data-model §5; contracts/informe-del-job.md §2).
const (
	// motivoComun: cambia una ruta de lo común, o en go.mod la versión de Go o
	// la de un módulo de lo común (FR-002).
	motivoComun motivoDeLaEleccion = "comun"
	// motivoSinConjunto: el conjunto de alguna eval no se puede calcular
	// (FR-005).
	motivoSinConjunto motivoDeLaEleccion = "sin_conjunto"
	// motivoPlataforma: la plataforma no da los ficheros de la propuesta
	// (research D17).
	motivoPlataforma motivoDeLaEleccion = "plataforma"
	// motivoSinAsignar: cambia una ruta que miden las evals y que no es de
	// ningún conjunto ni de lo común, o un directorio de skills/ o de evals/ que
	// está en uno solo de los dos árboles (FR-004).
	motivoSinAsignar motivoDeLaEleccion = "sin_asignar"
	// motivoDescripcion: cambia el description de la cabecera de algún SKILL.md
	// (FR-003).
	motivoDescripcion motivoDeLaEleccion = "descripcion"
	// motivoConjunto: cambia una ruta del conjunto de la eval, o en go.mod la
	// versión de uno de sus módulos (FR-001, FR-004).
	motivoConjunto motivoDeLaEleccion = "conjunto"
)

// prioridadDeLosMotivos son los motivos de más a menos prioridad: el de una
// eval es el de más prioridad entre los que la eligen (research D19).
var prioridadDeLosMotivos = []motivoDeLaEleccion{
	motivoComun, motivoSinConjunto, motivoPlataforma, motivoSinAsignar, motivoDescripcion, motivoConjunto,
}

// antesQue dice si el motivo tiene más prioridad que el otro.
func (m motivoDeLaEleccion) antesQue(otro motivoDeLaEleccion) bool {
	return slices.Index(prioridadDeLosMotivos, m) < slices.Index(prioridadDeLosMotivos, otro)
}

// evalElegida es una eval que elige la propuesta: su skill, su fichero sin
// .yaml, su motivo y las rutas cambiadas que la eligen, ordenadas (FR-008).
type evalElegida struct {
	skill  string
	eval   string
	motivo motivoDeLaEleccion
	rutas  []string
}

// nombre es el de la eval, <skill>/<fichero sin .yaml>.
func (e evalElegida) nombre() string {
	return e.skill + "/" + e.eval
}

// errorDeLaPlataforma es por qué la plataforma no dio los ficheros de la
// propuesta (research D17): la orden de gh y su salida de error, o «plazo de
// 10 minutos agotado» con la orden en curso.
type errorDeLaPlataforma struct {
	orden string
	causa string
}

// eleccionDeLaPropuesta es la elección de una propuesta (data-model §5): las
// evals elegidas, en orden de skill y de fichero; la causa por la que algún
// conjunto no se puede calcular, si la hay; y el error de la plataforma, si no
// dio los ficheros de la propuesta.
type eleccionDeLaPropuesta struct {
	elegidas    []evalElegida
	sinConjunto string
	plataforma  *errorDeLaPlataforma
}

// lineas son las del registro de la elección (contracts/eleccion-y-huella.md
// §3): la del error de la plataforma, si lo hay, y una por eval elegida, en su
// orden, con su motivo y, si las tiene, sus rutas, separadas por comas.
func (e eleccionDeLaPropuesta) lineas() []string {
	var lineas []string

	if e.plataforma != nil {
		lineas = append(lineas, "error de la plataforma: "+e.plataforma.orden+": "+e.plataforma.causa)
	}

	for _, elegida := range e.elegidas {
		linea := "elegida " + elegida.nombre() + ": " + string(elegida.motivo)
		if len(elegida.rutas) > 0 {
			linea += ": " + strings.Join(elegida.rutas, ", ")
		}

		lineas = append(lineas, linea)
	}

	return lineas
}

// calculoDeLaEleccion es lo que la elección lee (data-model §4 y §5): el
// conjunto de cada eval y lo común, calculados sobre la cabeza, y los árboles
// de la cabeza y de la base de la propuesta.
type calculoDeLaEleccion struct {
	conjuntos []dependenciasDeLaEval
	comun     loComun
	cabeza    arbolDelRepositorio
	base      arbolDelRepositorio
}

// elegir es la elección de la propuesta con sus ficheros (FR-002 a FR-008;
// data-model §5): cada ruta cambiada —las dos de un renombrado— elige por la
// primera de sus reglas que se aplica, y un conjunto que no se puede calcular
// elige todas. Una propuesta que no cambia nada de lo que miden las evals no
// elige ninguna.
func (c calculoDeLaEleccion) elegir(cambios []cambioDeLaPropuesta) eleccionDeLaPropuesta {
	var rutas []string
	for _, cambio := range cambios {
		rutas = append(rutas, cambio.rutas()...)
	}

	slices.Sort(rutas)

	return c.eleccion(slices.Compact(rutas), nil)
}

// elegirSinLosFicheros es la elección de una propuesta cuyos ficheros no da la
// plataforma (research D17): todas, con motivo plataforma y sin rutas, salvo
// que algún conjunto no se pueda calcular.
func (c calculoDeLaEleccion) elegirSinLosFicheros(plataforma errorDeLaPlataforma) eleccionDeLaPropuesta {
	return c.eleccion(nil, &plataforma)
}

// eleccion es la elección de las rutas cambiadas, ordenadas y sin repetir, o
// la de todas sin rutas si la plataforma no las dio: cada eval, con el motivo
// de más prioridad entre los que la eligen y todas las rutas que la eligen.
func (c calculoDeLaEleccion) eleccion(rutas []string, plataforma *errorDeLaPlataforma) eleccionDeLaPropuesta {
	votos := make([]votoDeLaEleccion, len(c.conjuntos))
	eleccion := eleccionDeLaPropuesta{sinConjunto: c.sinConjunto(), plataforma: plataforma}

	for indice := range votos {
		if eleccion.sinConjunto != "" {
			votos[indice].anotar(motivoSinConjunto, "")
		}

		if plataforma != nil {
			votos[indice].anotar(motivoPlataforma, "")
		}
	}

	paquetes := paquetesDeLosConjuntos(c.conjuntos)

	for _, ruta := range rutas {
		motivo, elegidas := c.eligeLaRuta(ruta, paquetes)
		for _, indice := range elegidas {
			votos[indice].anotar(motivo, ruta)
		}
	}

	for indice, voto := range votos {
		if voto.motivo != "" {
			conjunto := c.conjuntos[indice]
			eleccion.elegidas = append(eleccion.elegidas,
				evalElegida{skill: conjunto.skill, eval: conjunto.eval, motivo: voto.motivo, rutas: voto.rutas})
		}
	}

	slices.SortFunc(eleccion.elegidas, func(una, otra evalElegida) int {
		return cmp.Or(strings.Compare(una.skill, otra.skill), strings.Compare(una.eval, otra.eval))
	})

	return eleccion
}

// votoDeLaEleccion es lo que eligen de una eval las reglas: el motivo de más
// prioridad, vacío si nada la elige, y las rutas que la eligen, en el orden en
// que se anotan.
type votoDeLaEleccion struct {
	motivo motivoDeLaEleccion
	rutas  []string
}

// anotar anota que el motivo elige la eval, por la ruta si no está vacía.
func (v *votoDeLaEleccion) anotar(motivo motivoDeLaEleccion, ruta string) {
	if v.motivo == "" || motivo.antesQue(v.motivo) {
		v.motivo = motivo
	}

	if ruta != "" {
		v.rutas = append(v.rutas, ruta)
	}
}

// sinConjunto es la causa por la que no se pueden calcular los conjuntos que
// no se calculan, cada una con su eval delante, o vacía si se calculan todos
// (FR-005).
func (c calculoDeLaEleccion) sinConjunto() string {
	var errores []error

	for _, conjunto := range c.conjuntos {
		if conjunto.err != nil {
			errores = append(errores, conjunto.err)
		}
	}

	if err := errors.Join(errores...); err != nil {
		return err.Error()
	}

	return ""
}

// eligeLaRuta es lo que elige la ruta cambiada, por la primera regla de
// data-model §5 que se aplica: el motivo y el índice de cada conjunto que
// elige, o nada si la ruta no la miden las evals.
func (c calculoDeLaEleccion) eligeLaRuta(ruta string, paquetes []string) (motivoDeLaEleccion, []int) {
	switch {
	case !rutaQueMide(ruta, paquetes):
		return "", nil
	case slices.Contains(ficherosQueMiden, ruta):
		return c.eligeElGoMod()
	case esDeLoComun(ruta):
		return motivoComun, c.todas()
	case c.cambiaLaDescripcion(ruta):
		return motivoDescripcion, c.todas()
	case c.directorioDeUnSoloArbol(ruta):
		return motivoSinAsignar, c.todas()
	}

	var elegidas []int

	for indice, conjunto := range c.conjuntos {
		if conjunto.pertenece(ruta) {
			elegidas = append(elegidas, indice)
		}
	}

	if len(elegidas) == 0 {
		return motivoSinAsignar, c.todas()
	}

	return motivoConjunto, elegidas
}

// eligeElGoMod es lo que elige un cambio de go.mod o de go.sum (FR-004;
// data-model §5, regla 2), por lo que cambia entre los go.mod de la base y de
// la cabeza: todas, con comun, si cambian go, toolchain o la versión de un
// módulo de lo común; las evals de cuyo conjunto cambia la versión de algún
// módulo, con conjunto; y, si no cambia ninguno de los dos, todas, sin asignar.
func (c calculoDeLaEleccion) eligeElGoMod() (motivoDeLaEleccion, []int) {
	cambiados := modulosCambiados(c.base.modulos, c.cabeza.modulos)
	deUnosModulos := func(modulos map[string]string) bool {
		return slices.ContainsFunc(cambiados, func(modulo string) bool {
			_, esSuyo := modulos[modulo]

			return esSuyo
		})
	}

	if c.base.versionDeGo != c.cabeza.versionDeGo || c.base.toolchain != c.cabeza.toolchain ||
		deUnosModulos(c.comun.modulos) {
		return motivoComun, c.todas()
	}

	var elegidas []int

	for indice, conjunto := range c.conjuntos {
		if deUnosModulos(conjunto.modulos) {
			elegidas = append(elegidas, indice)
		}
	}

	if len(elegidas) == 0 {
		return motivoSinAsignar, c.todas()
	}

	return motivoConjunto, elegidas
}

// modulosCambiados son los módulos requeridos cuya versión difiere entre los
// dos go.mod, también los que solo requiere uno, ordenados.
func modulosCambiados(antes, despues map[string]string) []string {
	modulos := slices.Concat(slices.Collect(maps.Keys(antes)), slices.Collect(maps.Keys(despues)))
	slices.Sort(modulos)

	return slices.DeleteFunc(slices.Compact(modulos), func(modulo string) bool {
		version, enLaBase := antes[modulo]
		otra, enLaCabeza := despues[modulo]

		return enLaBase == enLaCabeza && version == otra
	})
}

// cambiaLaDescripcion dice si la ruta es el SKILL.md de una skill cuyo
// description difiere entre la base y la cabeza, también si la skill está en
// uno solo de los dos árboles (FR-003). El cuerpo no cuenta.
func (c calculoDeLaEleccion) cambiaLaDescripcion(ruta string) bool {
	skill, esElSuyo := skillDelFichero(ruta)
	if !esElSuyo {
		return false
	}

	antes, enLaBase := c.base.descripciones[skill]
	despues, enLaCabeza := c.cabeza.descripciones[skill]

	return enLaBase != enLaCabeza || antes != despues
}

// directorioDeUnSoloArbol dice si el primer componente de la ruta bajo
// skills/ o bajo evals/ está en uno solo de los dos árboles: un directorio que
// se añade, se retira o se renombra (FR-004).
func (c calculoDeLaEleccion) directorioDeUnSoloArbol(ruta string) bool {
	if skill, bajoSkills := primerComponente(ruta, directorioDeLasSkills); bajoSkills {
		return slices.Contains(c.base.skills, skill) != slices.Contains(c.cabeza.skills, skill)
	}

	if directorio, bajoEvals := primerComponente(ruta, directorioDeLasEvals); bajoEvals {
		return slices.Contains(c.base.directoriosDeEvals, directorio) !=
			slices.Contains(c.cabeza.directoriosDeEvals, directorio)
	}

	return false
}

// todas son los índices de todos los conjuntos.
func (c calculoDeLaEleccion) todas() []int {
	indices := make([]int, len(c.conjuntos))
	for indice := range indices {
		indices[indice] = indice
	}

	return indices
}

// trabajoDecidido es un trabajo evals (<skill>) de una ejecución del job
// (FR-019 de H30): su skill y su decisión, el bloque origen sin
// sesiones_abiertas.
type trabajoDecidido struct {
	skill  string
	origen *origenDeLasEvals
}

// sinTrabajosDeEvals es la matriz de una ejecución que no corre ningún trabajo
// de evals: la apertura de una propuesta que no elige ninguna (FR-016 de H30).
const sinTrabajosDeEvals = "ningún trabajo de evals\n"

// matrizDeLaEjecucion es la matriz de una ejecución en texto (FR-019 de H30;
// contracts/eleccion-y-huella.md §3 y §5 de H30): por trabajo, en orden de
// skill, la línea «evals (<skill>): medidas <m>, reutilizadas <r>» y, con dos
// espacios delante, las del registro de su decisión (lineasDeLaDecision): el
// error de la plataforma, si lo hubo, y una por eval en orden de fichero,
// medida con su motivo o reutilizada con su procedencia. Sin sesiones: un
// trabajo que lo reutiliza todo tiene su línea con medidas 0. Sin trabajos, la
// línea sinTrabajosDeEvals.
func matrizDeLaEjecucion(trabajos []trabajoDecidido) string {
	if len(trabajos) == 0 {
		return sinTrabajosDeEvals
	}

	ordenados := slices.SortedStableFunc(slices.Values(trabajos), func(uno, otro trabajoDecidido) int {
		return strings.Compare(uno.skill, otro.skill)
	})

	var matriz strings.Builder

	for _, trabajo := range ordenados {
		matriz.WriteString("evals (" + trabajo.skill + "): medidas " + strconv.Itoa(trabajo.origen.EvalsMedidas) +
			", reutilizadas " + strconv.Itoa(trabajo.origen.EvalsReutilizadas) + "\n")

		for _, linea := range lineasDeLaDecision(trabajo.origen) {
			matriz.WriteString("  " + linea + "\n")
		}
	}

	return matriz.String()
}
