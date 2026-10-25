package evals

import (
	"cmp"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"slices"
	"strconv"
	"strings"
	"time"
)

// Lo que acota la búsqueda de resultados en las medidas de main (FR-012 de
// H30; data-model §7 y §8 de H30; research D8 y D17 de H30).
const (
	// medidasQueSeRecorren son las medidas de main más recientes por las que pasa
	// la búsqueda, como mucho.
	medidasQueSeRecorren = 10

	// vidaDeUnResultado es la del registro de la plataforma: una medida, o una
	// procedencia, de esta edad o más antes de la decisión no da ningún
	// resultado reutilizable.
	vidaDeUnResultado = 90 * 24 * time.Hour

	// plazoDeLaPlataforma es el de lo que la decisión de un trabajo lee de la
	// plataforma, los ficheros de la propuesta y la búsqueda juntos: lo pone en
	// el contexto quien decide.
	plazoDeLaPlataforma = 10 * time.Minute
)

// Lo que dice de una ejecución del flujo si es medida de main (FR-010 de H30;
// contracts/job-de-evals.md §2 de H30): la conclusión de una ejecución
// cancelada; el nombre de los trabajos de la matriz, evals (<skill>), y la
// conclusión con la que termina uno en rojo; y la marca, el último paso de cada
// uno, que solo termina en success si su informe está escrito.
const (
	conclusionCancelada       = "cancelled"
	conclusionDeFallo         = "failure"
	prefijoDeUnTrabajoDeEvals = "evals ("
	sufijoDeUnTrabajoDeEvals  = ")"
	marcaDelInformeEscrito    = "El informe del trabajo está escrito"
)

// evalABuscar es una eval no elegida que busca su resultado en las medidas de
// main: su fichero, el nombre con el que la nombran el origen y las series de
// un informe, y su huella.
type evalABuscar struct {
	fichero string
	huella  string
}

// medidaDeMain es una ejecución del flujo evals sobre la cabeza de una
// propuesta fusionada en main por la que pasa la búsqueda (data-model §7 de
// H30): la ejecución, el trabajo de la skill una vez leídos sus trabajos, su
// commit, la propuesta de cuya cabeza es y la fecha en que se creó.
type medidaDeMain struct {
	ejecucion int64
	trabajo   int64
	commit    string
	propuesta int
	fecha     time.Time
}

// buscarEnLasMedidasDeMain da el resultado de cada eval no elegida de la
// skill que hay en las medidas de main (FR-010 a FR-013 de H30; data-model §7
// y §8 de H30; research D8, D9 y D17 de H30): una entrada del origen por eval
// buscada, en su orden, que es la reutilizada de la medida de main más reciente
// que tiene de ella un resultado reutilizable (resultadoReutilizable) o, si
// ninguna lo tiene, la de la eval medida con sin_resultado_de_main. En el mismo
// recorrido, si la huella de la descripción no está vacía, busca el juicio de
// la descripción de la skill con esa huella (FR-030 de H31; research D4 de H31):
// el segundo resultado es el de la medida de main más reciente que tiene uno
// reutilizable (juicioReutilizable), o nil si ninguna lo tiene o no lo busca.
// Lee de la plataforma solo si hay algo que buscar y para en cuanto todo lo
// buscado tiene resultado. La primera lectura que falla, o el plazo del
// contexto agotado, la para: lo que queda se mide o se juzga, y el error, con
// la orden que falló o la que estaba en curso, es el tercer resultado; nil si
// no hubo ninguno. El instante es el de la decisión, con el que se compara la
// edad de cada medida y de cada procedencia.
func buscarEnLasMedidasDeMain(
	ctx context.Context, plataforma lecturasDeLaPlataforma, skill string, buscadas []evalABuscar,
	huellaDeLaDescripcion string, instante time.Time,
) ([]origenDeEval, *descripcionDeLaDecision, *errorDeLaPlataformaEnElOrigen) {
	busqueda := &busquedaEnMain{
		skill: skill, buscadas: buscadas, huellaDeLaDescripcion: huellaDeLaDescripcion, instante: instante,
		halladas: make(map[string]origenDeEval, len(buscadas)),
	}
	plataformaFallida := busqueda.recorrer(ctx, plataforma)

	evals := make([]origenDeEval, 0, len(buscadas))

	for _, buscada := range buscadas {
		hallada, ok := busqueda.halladas[buscada.fichero]
		if !ok {
			hallada = origenDeEval{Eval: buscada.fichero, Huella: buscada.huella, Medida: true, Motivo: motivoSinResultadoDeMain}
		}

		evals = append(evals, hallada)
	}

	return evals, busqueda.juicio, plataformaFallida
}

// busquedaEnMain es lo que la decisión de un trabajo busca en las medidas de
// main —el resultado de cada eval no elegida de la skill y, si la huella no
// está vacía, el juicio de su descripción con esa huella— con el instante de
// la decisión, y lo que lleva hallado: el resultado de cada eval por su fichero
// y el juicio.
type busquedaEnMain struct {
	skill                 string
	buscadas              []evalABuscar
	huellaDeLaDescripcion string
	instante              time.Time

	halladas map[string]origenDeEval
	juicio   *descripcionDeLaDecision
}

// terminada dice si todo lo buscado tiene resultado: cada eval y, si lo busca,
// el juicio de la descripción.
func (b *busquedaEnMain) terminada() bool {
	return len(b.halladas) == len(b.buscadas) && (b.huellaDeLaDescripcion == "" || b.juicio != nil)
}

// recorrer pasa por las medidas de main, de la más reciente a la más antigua y
// como mucho medidasQueSeRecorren, y anota en la búsqueda lo que cada una
// tiene de lo que aún no tiene resultado, hasta que todo lo tiene (FR-012 de
// H30; FR-030 de H31). Una posible medida que sus trabajos descartan
// (trabajoDeLaSkillEnLaMedida) no cuenta entre las recorridas, y una medida
// cuyo informe no se puede sacar del registro o leer no aporta nada (research
// D8 de H30). Devuelve el error de la lectura que la paró, o nil.
func (b *busquedaEnMain) recorrer(ctx context.Context, plataforma lecturasDeLaPlataforma) *errorDeLaPlataformaEnElOrigen {
	if b.terminada() {
		return nil
	}

	posibles, plataformaFallida := posiblesMedidasDeMain(ctx, plataforma, b.instante)
	if plataformaFallida != nil {
		return plataformaFallida
	}

	recorridas := 0

	for _, posible := range posibles {
		if recorridas == medidasQueSeRecorren || b.terminada() {
			break
		}

		informe, esMedida, plataformaFallida := leerLaMedida(ctx, plataforma, b.skill, &posible)
		if plataformaFallida != nil {
			return plataformaFallida
		}

		if !esMedida {
			continue
		}

		recorridas++

		if informe != nil {
			b.anotar(informe, posible)
		}
	}

	return nil
}

// anotar anota en la búsqueda lo que el informe de la medida de main tiene de
// lo que aún no tiene resultado: el resultado reutilizable de cada eval y el
// juicio reutilizable de la descripción, si lo busca.
func (b *busquedaEnMain) anotar(informe *Informe, medida medidaDeMain) {
	for _, buscada := range b.buscadas {
		if _, hallada := b.halladas[buscada.fichero]; hallada {
			continue
		}

		if resultado, ok := resultadoReutilizable(informe, medida, buscada, b.instante); ok {
			b.halladas[buscada.fichero] = resultado
		}
	}

	if b.huellaDeLaDescripcion != "" && b.juicio == nil {
		b.juicio = juicioReutilizable(informe.Descripcion, medida, b.skill, b.huellaDeLaDescripcion, b.instante)
	}
}

// posiblesMedidasDeMain son las ejecuciones del flujo que pueden ser medida
// de main por lo que dice de ellas la lista (data-model §7 de H30): terminadas,
// no canceladas, sobre la cabeza con que se fusionó una propuesta en main —su
// headSha es el head.sha de la propuesta, no otro commit suyo— y creadas menos
// de vidaDeUnResultado antes del instante; de la más reciente a la más antigua
// por su creación y, a la misma, por su número. Si una lista no se puede leer,
// el error de su lectura.
func posiblesMedidasDeMain(
	ctx context.Context, plataforma lecturasDeLaPlataforma, instante time.Time,
) ([]medidaDeMain, *errorDeLaPlataformaEnElOrigen) {
	ejecuciones, err := plataforma.ejecucionesDelFlujo(ctx)
	if err != nil {
		return nil, errorDeLaLectura(ctx, err, ordenDeLasEjecucionesDelFlujo)
	}

	propuestas, err := plataforma.propuestasFusionadas(ctx)
	if err != nil {
		return nil, errorDeLaLectura(ctx, err, ordenDeLasPropuestasFusionadas)
	}

	cabezas := make(map[string]int, len(propuestas))

	for _, propuesta := range propuestas {
		cabezas[propuesta.cabeza] = propuesta.numero
	}

	var posibles []medidaDeMain

	for _, ejecucion := range ejecuciones {
		propuesta, fusionada := cabezas[ejecucion.Commit]
		if !fusionada || ejecucion.Status != statusTerminado || ejecucion.Conclusion == conclusionCancelada ||
			!vigente(ejecucion.Creada, instante) {
			continue
		}

		posibles = append(posibles, medidaDeMain{
			ejecucion: ejecucion.ID, commit: ejecucion.Commit, propuesta: propuesta, fecha: ejecucion.Creada,
		})
	}

	slices.SortStableFunc(posibles, func(una, otra medidaDeMain) int {
		return cmp.Or(otra.fecha.Compare(una.fecha), cmp.Compare(otra.ejecucion, una.ejecucion))
	})

	return posibles, nil
}

// leerLaMedida lee los trabajos de la posible medida y, si lo es, pone en ella
// el trabajo de la skill y lee su informe, con su bloque descripcion si lo
// lleva (Informe.Descripcion). Devuelve el informe —nil si
// no se pudo sacar del registro o leer, que deja la medida sin aportar nada—, si
// es medida y el error de la lectura que falló, que para la búsqueda: la de los
// trabajos, y la del registro si su orden falló o se agotó el plazo.
func leerLaMedida(
	ctx context.Context, plataforma lecturasDeLaPlataforma, skill string, posible *medidaDeMain,
) (*Informe, bool, *errorDeLaPlataformaEnElOrigen) {
	ejecucion := strconv.FormatInt(posible.ejecucion, 10)

	trabajos, err := plataforma.trabajos(ctx, posible.ejecucion)
	if err != nil {
		return nil, false, errorDeLaLectura(ctx, err, ordenDeLosTrabajos, variableDeLaEjecucionAnterior+"="+ejecucion)
	}

	trabajo, esMedida := trabajoDeLaSkillEnLaMedida(trabajos, skill)
	if !esMedida {
		return nil, false, nil
	}

	posible.trabajo = trabajo

	contenido, err := plataforma.informeDelTrabajo(ctx, posible.ejecucion, trabajo)

	var deUnaOrden *errorDeUnaOrden

	switch {
	case err == nil:
	case errors.As(err, &deUnaOrden) || ctx.Err() != nil:
		return nil, true, errorDeLaLectura(ctx, err, ordenDelRegistroDeUnTrabajo,
			variableDeLaEjecucion+"="+ejecucion, variableDelTrabajo+"="+strconv.FormatInt(trabajo, 10))
	default:
		// El registro no trae el informe: la medida no aporta nada y lo que
		// buscaba en ella se busca en la anterior (research D8 de H30).
		return nil, true, nil
	}

	var informe Informe
	if json.Unmarshal(contenido, &informe) != nil {
		// Un informe que no es JSON con su forma tampoco aporta nada (FR-013 de
		// H30).
		return nil, true, nil
	}

	return &informe, true, nil
}

// trabajoDeLaSkillEnLaMedida es el id del trabajo evals (<skill>) de la skill
// en una ejecución, y si la ejecución es medida de main por sus trabajos
// (FR-010 de H30; data-model §7 de H30): lo es si tiene el trabajo de la skill
// y si cada trabajo evals (<skill>), el de cualquier skill, terminó en verde o
// en rojo con su marca del informe escrito en success.
func trabajoDeLaSkillEnLaMedida(trabajos []trabajoDeLaEjecucion, skill string) (int64, bool) {
	var (
		deLaSkill     int64
		conLaSkill    bool
		nombreBuscado = prefijoDeUnTrabajoDeEvals + skill + sufijoDeUnTrabajoDeEvals
	)

	for _, trabajo := range trabajos {
		if !strings.HasPrefix(trabajo.Name, prefijoDeUnTrabajoDeEvals) ||
			!strings.HasSuffix(trabajo.Name, sufijoDeUnTrabajoDeEvals) {
			continue
		}

		if !conElInformeEscrito(trabajo) {
			return 0, false
		}

		if trabajo.Name == nombreBuscado {
			deLaSkill, conLaSkill = trabajo.ID, true
		}
	}

	return deLaSkill, conLaSkill
}

// conElInformeEscrito dice si un trabajo evals (<skill>) terminó en verde o en
// rojo con su marca del informe escrito en success.
func conElInformeEscrito(trabajo trabajoDeLaEjecucion) bool {
	terminado := trabajo.Status == statusTerminado &&
		(trabajo.Conclusion == conclusionDeExito || trabajo.Conclusion == conclusionDeFallo)

	return terminado && slices.ContainsFunc(trabajo.Steps, func(paso pasoDeLaEjecucion) bool {
		return paso.Name == marcaDelInformeEscrito && paso.Conclusion == conclusionDeExito
	})
}

// resultadoReutilizable es el resultado de la eval buscada en el informe de una
// medida de main, y si es reutilizable (data-model §8 de H30; FR-011 y FR-012
// de H30): la entrada de la eval en su origen tiene que llevar su misma huella.
// Si se midió allí, su procedencia es esa medida, sus aportes los de la
// entrada y sus series las del informe con esa eval, y solo si está completa
// (completaEnElInforme). Si se reutilizó allí, su procedencia, sus aportes y
// sus series son los de la entrada, los de la medida en que se abrieron sus
// sesiones. En los dos casos, su procedencia tiene que ser de menos de
// vidaDeUnResultado antes del instante.
func resultadoReutilizable(
	informe *Informe, medida medidaDeMain, buscada evalABuscar, instante time.Time,
) (origenDeEval, bool) {
	if informe.Origen == nil {
		return origenDeEval{}, false
	}

	indice := slices.IndexFunc(informe.Origen.Evals, func(entrada origenDeEval) bool {
		return entrada.Eval == buscada.fichero
	})
	if indice < 0 || informe.Origen.Evals[indice].Huella != buscada.huella {
		return origenDeEval{}, false
	}

	entrada := informe.Origen.Evals[indice]
	resultado := origenDeEval{
		Eval: buscada.fichero, Huella: buscada.huella, Procedencia: entrada.Procedencia, Aportes: entrada.Aportes,
		Tasas: entrada.Tasas,
	}

	if entrada.Medida {
		if !completaEnElInforme(informe, buscada.fichero) {
			return origenDeEval{}, false
		}

		resultado.Procedencia = medida.procedencia()
		resultado.Tasas = slices.DeleteFunc(slices.Clone(informe.Tasas), func(serie TasaDelInforme) bool {
			return serie.Eval != buscada.fichero
		})
	}

	if !procedenciaVigente(resultado.Procedencia, instante) {
		return origenDeEval{}, false
	}

	return resultado, true
}

// juicioReutilizable es el juicio de la descripción de la skill con esa huella
// que se reutiliza del bloque descripcion del informe de una medida de main, o
// nil si no es reutilizable (FR-030 y FR-032 de H31; research D4 de H31): el
// bloque tiene que ser de la skill, con la misma huella, y su juicio, completo,
// no sin juzgar. Si se juzgó allí, su procedencia es esa medida; si se
// reutilizó allí, la del bloque, la de la medida en que se juzgó. En los dos
// casos, como la de un resultado de una eval, de menos de vidaDeUnResultado
// antes del instante.
func juicioReutilizable(
	bloque *DescripcionInformada, medida medidaDeMain, skill, huella string, instante time.Time,
) *descripcionDeLaDecision {
	if bloque == nil || bloque.Skill != skill || bloque.Huella != huella || bloque.SinJuzgar != "" {
		return nil
	}

	procedencia := bloque.Procedencia
	if bloque.Juzgada {
		procedencia = medida.procedencia()
	}

	if !procedenciaVigente(procedencia, instante) {
		return nil
	}

	return &descripcionDeLaDecision{
		Skill: skill, Huella: huella, Procedencia: procedencia,
		Juicio: &juicioReutilizado{Marcada: bloque.Marcada, Frase: bloque.Frase, Motivo: bloque.Motivo},
	}
}

// procedencia es la de lo que se midió o se juzgó en la medida de main: su
// ejecución, el trabajo de la skill, su commit, su propuesta y la fecha en que
// se creó, en UTC (data-model §7 y §8 de H30).
func (m medidaDeMain) procedencia() *procedenciaDeLaMedida {
	return &procedenciaDeLaMedida{
		Ejecucion: m.ejecucion, Trabajo: m.trabajo, Commit: m.commit, Propuesta: m.propuesta,
		Fecha: m.fecha.UTC().Format(time.RFC3339),
	}
}

// procedenciaVigente dice si hay procedencia y su fecha, que tiene que leerse
// en RFC 3339, es de menos de vidaDeUnResultado antes del instante.
func procedenciaVigente(procedencia *procedenciaDeLaMedida, instante time.Time) bool {
	if procedencia == nil {
		return false
	}

	fecha, err := time.Parse(time.RFC3339, procedencia.Fecha)

	return err == nil && vigente(fecha, instante)
}

// completaEnElInforme dice si el resultado de una eval medida en un informe
// está completo (FR-012 de H30; data-model §8 de H30): ninguna de sus sesiones
// quedó sin medir y, con juez, ninguna respuesta de una de sus sesiones del
// modelo que decide quedó sin juzgar. Una respuesta sin juzgar de una sesión
// que el informe no tiene puede ser de cualquier eval: ninguna está completa.
func completaEnElInforme(informe *Informe, fichero string) bool {
	if slices.ContainsFunc(informe.SesionesSinMedir, func(sesion SesionSinMedir) bool { return sesion.Eval == fichero }) {
		return false
	}

	if informe.Juez == nil {
		return true
	}

	return !slices.ContainsFunc(informe.Juez.SinJuzgar, func(respuesta RespuestaSinJuzgar) bool {
		indice := slices.IndexFunc(informe.Evals, func(sesion ResultadoDeEval) bool { return sesion.Sesion == respuesta.Sesion })

		return indice < 0 || informe.Evals[indice].Eval == fichero && informe.Evals[indice].Modelo == informe.ModeloQueDecide
	})
}

// vigente dice si una medida o una procedencia de esa fecha es de menos de
// vidaDeUnResultado antes del instante.
func vigente(fecha, instante time.Time) bool {
	return instante.Sub(fecha) < vidaDeUnResultado
}

// errorDeLaLectura es el error de la plataforma de una lectura que falló
// (contracts/informe-del-job.md §2 de H30; research D17 de H30): su orden, con
// las variables con las que recibe sus valores, como la nombra un
// *errorDeUnaOrden; y el plazo agotado si el contexto lo agotó, la salida de
// error de la orden —o, si no escribió nada, cómo terminó— si gh no terminó con
// 0, o el error de la lectura si su respuesta no se pudo leer.
func errorDeLaLectura(ctx context.Context, err error, orden string, variables ...string) *errorDeLaPlataformaEnElOrigen {
	if len(variables) > 0 {
		orden += ", con " + strings.Join(variables, ", ")
	}

	var deUnaOrden *errorDeUnaOrden

	switch {
	case errors.Is(ctx.Err(), context.DeadlineExceeded):
		return &errorDeLaPlataformaEnElOrigen{Orden: orden, Error: plazoDeLaPlataformaAgotado()}
	case errors.As(err, &deUnaOrden):
		return &errorDeLaPlataformaEnElOrigen{
			Orden: orden, Error: cmp.Or(strings.TrimSpace(deUnaOrden.salidaDeError), fmt.Sprint(deUnaOrden.causa)),
		}
	default:
		return &errorDeLaPlataformaEnElOrigen{Orden: orden, Error: err.Error()}
	}
}

// plazoDeLaPlataformaAgotado es el error de la plataforma cuando se agota
// plazoDeLaPlataforma (contracts/informe-del-job.md §2 de H30).
func plazoDeLaPlataformaAgotado() string {
	return fmt.Sprintf("plazo de %d minutos agotado", int(plazoDeLaPlataforma/time.Minute))
}
