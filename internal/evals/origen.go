package evals

import (
	"cmp"
	"fmt"
	"slices"
	"strconv"
	"strings"
)

// modoDelOrigen es cómo mide una ejecución las evals de su skill
// (contracts/informe-del-job.md §2 de H30; data-model §8 de H30).
type modoDelOrigen string

const (
	// origenDeLaSuiteCompleta: las mide todas, sin buscar nada en main (FR-018).
	origenDeLaSuiteCompleta modoDelOrigen = "completa"

	// origenConReutilizacion: mide las que elige la propuesta, y cada una de las
	// demás reutiliza el resultado de la medida de main con su huella o, sin él,
	// se mide (FR-012 a FR-014).
	origenConReutilizacion modoDelOrigen = "reutilizacion"
)

// Los motivos de una eval medida que la propuesta no elige, detrás de los de la
// elección (motivoDeLaEleccion; contracts/informe-del-job.md §2 de H30): sin
// resultado reutilizable de main (FR-013, FR-014) y la suite completa (FR-018).
const (
	motivoSinResultadoDeMain motivoDeLaEleccion = "sin_resultado_de_main"
	motivoSuiteCompleta      motivoDeLaEleccion = "suite_completa"
)

// Textos fijos de la sección «Medidas y reutilizadas» de informe.md
// (contracts/informe-del-job.md §6 de H30).
const (
	tituloDeLasMedidasYReutilizadas = "## Medidas y reutilizadas"
	etiquetaDeLasSeriesReutilizadas = "Series reutilizadas"
	medidaPorUnErrorDeLaPlataforma  = ", la búsqueda paró por un error de la plataforma"
	reutilizadaDeLaEjecucion        = "reutilizada de la ejecución %d (propuesta #%d, %s, %s)"

	// cifrasDelCommitCorto son las del commit de una procedencia en informe.md,
	// las de su forma abreviada en git.
	cifrasDelCommitCorto = 7
)

// motivoEnPalabras es cómo dice informe.md el motivo de una eval medida y si
// lleva detrás, entre paréntesis, las rutas que la eligen.
type motivoEnPalabras struct {
	palabras string
	conRutas bool
}

// motivosEnPalabras son los de cada motivo de una eval medida
// (contracts/informe-del-job.md §6 de H30).
var motivosEnPalabras = map[motivoDeLaEleccion]motivoEnPalabras{
	motivoConjunto:           {palabras: "rutas de su conjunto", conRutas: true},
	motivoComun:              {palabras: "común", conRutas: true},
	motivoDescripcion:        {palabras: "descripción", conRutas: true},
	motivoSinAsignar:         {palabras: "sin asignar", conRutas: true},
	motivoSinConjunto:        {palabras: "sin conjunto calculable"},
	motivoPlataforma:         {palabras: "la plataforma no dio los ficheros de la propuesta"},
	motivoSinResultadoDeMain: {palabras: "sin resultado reutilizable de `main`"},
	motivoSuiteCompleta:      {palabras: "suite completa"},
}

// origenDeLasEvals es lo que una ejecución del job midió y lo que reutilizó de
// cada eval de su skill (data-model §8 de H30; contracts/informe-del-job.md §2
// de H30; FR-011, FR-022): el bloque origen de informe.json, siempre presente,
// y, sin SesionesAbiertas, la decisión del job. Sus claves JSON, en el orden de
// sus campos, son las del bloque.
type origenDeLasEvals struct {
	// Modo es el de la ejecución: la suite completa o con reutilización.
	Modo modoDelOrigen `json:"modo"`

	// SinConjunto es la causa por la que algún conjunto no se pudo calcular en
	// el job (FR-005); vacía, y ausente, si no hubo ninguna.
	SinConjunto string `json:"sin_conjunto,omitzero"`

	// ErrorDeLaPlataforma es el de la orden de gh de la decisión que falló, o el
	// plazo de la búsqueda agotado; nil, y ausente, si no lo hubo.
	ErrorDeLaPlataforma *errorDeLaPlataformaEnElOrigen `json:"error_de_la_plataforma,omitzero"`

	// EvalsMedidas y EvalsReutilizadas son cuántas de Evals se midieron y cuántas
	// se reutilizaron: suman las evals de la skill en el commit.
	EvalsMedidas      int `json:"evals_medidas"`
	EvalsReutilizadas int `json:"evals_reutilizadas"`

	// SesionesAbiertas son las sesiones que abrió la ejecución; nil, y ausente,
	// en la decisión, que es anterior a ellas.
	SesionesAbiertas *int `json:"sesiones_abiertas,omitzero"`

	// Evals son una entrada por eval de la skill, en orden de fichero.
	Evals []origenDeEval `json:"evals"`

	// Descripcion es el bloque descripcion de la decisión, el de la descripción
	// de la skill que el trabajo juzga o cuyo juicio reutiliza
	// (contracts/juez-de-las-descripciones.md §2 de H31; FR-030 de H31); nil, y
	// ausente, si la ejecución no juzga descripciones y en el bloque origen de
	// informe.json, que no lo lleva (origenRecibido).
	Descripcion *descripcionDeLaDecision `json:"descripcion,omitzero"`
}

// descripcionDeLaDecision es lo que la decisión de un trabajo dice de la
// descripción de su skill (contracts/juez-de-las-descripciones.md §2 de H31;
// data-model §4 de H31; research D3 y D4 de H31): la skill, la huella de su
// juicio (huellaDelJuicio) y si se juzga en la ejecución o, con reutilización,
// se reutiliza el de una medida de main con su huella, con la procedencia de
// ese juicio, con la forma de la de H30, y el juicio. Sus claves JSON, en el
// orden de sus campos, son las del bloque.
type descripcionDeLaDecision struct {
	Skill   string `json:"skill"`
	Huella  string `json:"huella"`
	Juzgada bool   `json:"juzgada"`

	// Procedencia y Juicio son los del juicio reutilizado; nil, y ausentes, si
	// se juzga en la ejecución.
	Procedencia *procedenciaDeLaMedida `json:"procedencia,omitzero"`
	Juicio      *juicioReutilizado     `json:"juicio,omitzero"`
}

// juicioReutilizado es el juicio de una descripción que la decisión reutiliza
// de una medida de main: si quedó marcada y la frase y el motivo de su primer
// voto (contracts/juez-de-las-descripciones.md §2 de H31).
type juicioReutilizado struct {
	Marcada bool   `json:"marcada"`
	Frase   string `json:"frase"`
	Motivo  string `json:"motivo"`
}

// origenDeEval es lo que el origen dice de una eval: su nombre, el mismo que
// en las series del informe; su huella, ausente solo sin conjunto; si se midió
// y, de una medida, su motivo y las rutas que la eligen o, de una reutilizada,
// la medida de main en que se abrieron sus sesiones y sus series; y lo que
// aportan sus respuestas a cada umbral de las respuestas (data-model §8 de H30;
// research D9 de H30).
type origenDeEval struct {
	Eval   string `json:"eval"`
	Huella string `json:"huella,omitzero"`
	Medida bool   `json:"medida"`

	Motivo motivoDeLaEleccion `json:"motivo,omitzero"`
	Rutas  []string           `json:"rutas,omitzero"`

	Procedencia *procedenciaDeLaMedida `json:"procedencia,omitzero"`

	// Aportes son, de cada umbral de las respuestas del informe —todos menos
	// los de la duración y los de la medida del juez—, la medida y el total que
	// dan las respuestas de la eval, solo los de total mayor que 0, en el orden
	// de los umbrales.
	Aportes []aporteDeLaEval `json:"aportes"`

	// Tasas son las series de una reutilizada, también las de los modelos
	// informativos, como en su medida; ninguna, y ausentes, en una medida, cuyas
	// series están en las del informe.
	Tasas []TasaDelInforme `json:"tasas,omitzero"`
}

// procedenciaDeLaMedida es la medida de main en que se abrieron las sesiones de
// una eval reutilizada: su ejecución, el trabajo de la skill, su commit, su
// propuesta y su fecha (data-model §7 y §8 de H30; FR-011).
type procedenciaDeLaMedida struct {
	Ejecucion int64  `json:"ejecucion"`
	Trabajo   int64  `json:"trabajo"`
	Commit    string `json:"commit"`
	Propuesta int    `json:"propuesta"`
	Fecha     string `json:"fecha"`
}

// aporteDeLaEval es lo que dan las respuestas de una eval a un umbral de las
// respuestas: su medida y su total (data-model §8 de H30).
type aporteDeLaEval struct {
	Umbral string  `json:"umbral"`
	Medida float64 `json:"medida"`
	Total  int     `json:"total"`
}

// errorDeLaPlataformaEnElOrigen es el error de la plataforma como lo escribe el
// origen (contracts/informe-del-job.md §2 de H30; research D17 de H30): la
// orden de gh que falló, o la que estaba en curso al agotarse el plazo, y su
// salida de error o el plazo agotado.
type errorDeLaPlataformaEnElOrigen struct {
	Orden string `json:"orden"`
	Error string `json:"error"`
}

// reutilizadas son las evals reutilizadas del origen, en su orden; ninguna sin
// origen.
func (o *origenDeLasEvals) reutilizadas() []origenDeEval {
	if o == nil {
		return nil
	}

	var reutilizadas []origenDeEval

	for _, eval := range o.Evals {
		if !eval.Medida {
			reutilizadas = append(reutilizadas, eval)
		}
	}

	return reutilizadas
}

// medidas son, de las evals dadas, las que el origen no reutiliza, en su orden:
// todas sin origen, que es la ejecución de todas.
func (o *origenDeLasEvals) medidas(evals []Eval) []Eval {
	reutilizadas := o.reutilizadas()
	if len(reutilizadas) == 0 {
		return evals
	}

	return slices.DeleteFunc(slices.Clone(evals), func(eval Eval) bool {
		return slices.ContainsFunc(reutilizadas, func(reutilizada origenDeEval) bool {
			return reutilizada.Eval == eval.Fichero
		})
	})
}

// origenRecibido es el origen que da la decisión, sin sus sesiones abiertas ni
// su bloque descripcion, que no es del origen de informe.json, y con sus
// recuentos: una copia, que no cambia la decisión. Sin decisión, el de la suite
// completa, con cada una de las evals dadas medida por eso, en su orden.
func (e InformeAEscribir) origenRecibido(evals []Eval) *origenDeLasEvals {
	origen := origenDeLasEvals{Modo: origenDeLaSuiteCompleta}

	if e.decision != nil {
		origen = *e.decision
		origen.Evals = slices.Clone(e.decision.Evals)
		origen.SesionesAbiertas = nil
		origen.Descripcion = nil
	} else {
		for _, eval := range evals {
			origen.Evals = append(origen.Evals, origenDeEval{Eval: eval.Fichero, Medida: true, Motivo: motivoSuiteCompleta})
		}
	}

	origen.EvalsReutilizadas = len(origen.reutilizadas())
	origen.EvalsMedidas = len(origen.Evals) - origen.EvalsReutilizadas

	return &origen
}

// origenDelInforme es el bloque origen del informe de una ejecución que abrió
// sus sesiones (FR-011, FR-022 de H30; data-model §8 de H30;
// contracts/informe-del-job.md §2 de H30): el recibido (origenRecibido), con
// las sesiones abiertas, que son las del directorio de sesiones, y, de cada
// eval medida, sus aportes, medidos aquí (aportesDeLaEval). Los de una
// reutilizada son los de la medida de la que viene, tal cual.
func origenDelInforme(
	e InformeAEscribir, evals []Eval, sesiones []sesionJuzgada, series []serieJuzgada, respuestas *respuestasMedidas,
) *origenDeLasEvals {
	origen := e.origenRecibido(evals)

	abiertas := len(sesiones)
	origen.SesionesAbiertas = &abiertas

	for posicion := range origen.Evals {
		if deLaEval := &origen.Evals[posicion]; deLaEval.Medida {
			deLaEval.Aportes = aportesDeLaEval(e, evals, deLaEval.Eval, sesiones, series, respuestas)
		}
	}

	return origen
}

// origenSinNadaMedido es el bloque origen del informe de una ejecución que no
// abrió ninguna sesión, la del instrumento sin medir: el recibido, sin ningún
// aporte medido y con cero sesiones abiertas.
func origenSinNadaMedido(e InformeAEscribir) *origenDeLasEvals {
	origen := e.origenRecibido(nil)

	abiertas := 0
	origen.SesionesAbiertas = &abiertas

	return origen
}

// aportesDeLaEval son los aportes de la eval medida con ese fichero a cada
// umbral de las respuestas, calculados con las mismas funciones que los
// umbrales del informe sobre sus respuestas, ya juzgadas, y sus sesiones, sin
// volver a votar (research D9 de H30): los de sus respuestas en cada modo
// (respuestasMedidas.umbrales), los conjuntos de las clases del juez que
// deciden (umbralesConjuntos) y los de las lecturas de sus sesiones
// (lecturasMedidas.umbrales); de cada uno, la medida y el total, solo con
// total mayor que 0. Como cada respuesta y cada sesión es de una sola eval, los
// aportes de las evals medidas a un umbral suman su medida y su total.
func aportesDeLaEval(
	e InformeAEscribir, evals []Eval, fichero string, sesiones []sesionJuzgada, series []serieJuzgada,
	respuestas *respuestasMedidas,
) []aporteDeLaEval {
	var suyas []serieJuzgada

	for _, serie := range series {
		if serie.clave.eval == fichero {
			suyas = append(suyas, serie)
		}
	}

	deLaEval := respuestas.deLasSesiones(func(posicion int) bool { return sesiones[posicion].clave.eval == fichero })
	umbrales := slices.Concat(deLaEval.umbrales(), deLaEval.juez().umbralesConjuntos(),
		medirLasLecturas(e, evals, sesiones, suyas).umbrales())

	var aportes []aporteDeLaEval

	for _, umbral := range umbrales {
		if umbral.Total != nil && *umbral.Total > 0 {
			aportes = append(aportes, aporteDeLaEval{Umbral: umbral.Nombre, Medida: umbral.Medida, Total: *umbral.Total})
		}
	}

	return aportes
}

// deLasSesiones es lo que se mide de las respuestas de las sesiones, por su
// posición, que cumplen la condición: los mismos grupos, cada uno con solo esas
// respuestas, que llevan ya su juicio; nil si la skill no tiene umbrales de sus
// respuestas.
func (r *respuestasMedidas) deLasSesiones(cumple func(posicion int) bool) *respuestasMedidas {
	if r == nil {
		return nil
	}

	suyas := *r
	suyas.grupos = make([]grupoJuzgado, 0, len(r.grupos))

	for _, grupo := range r.grupos {
		grupo.respuestas = slices.DeleteFunc(slices.Clone(grupo.respuestas), func(respuesta respuestaDelJuez) bool {
			return !cumple(respuesta.posicion)
		})
		suyas.grupos = append(suyas.grupos, grupo)
	}

	if r.delJuez != nil {
		delJuez := *r.delJuez
		delJuez.grupos = suyas.grupos
		suyas.delJuez = &delJuez
	}

	return &suyas
}

// medidasYReutilizadas añade la sección «Medidas y reutilizadas» de informe.md
// (contracts/informe-del-job.md §6 de H30): los recuentos —evals medidas,
// reutilizadas y sesiones abiertas—; el error de la plataforma, si lo hubo, en
// una línea antes de la lista; una línea por eval, en su orden
// (lineaDelOrigen); y la tabla de las series reutilizadas, con los encabezados
// de la de las tasas, o «ninguna». Nada sin origen.
func (d *documento) medidasYReutilizadas(origen *origenDeLasEvals) {
	if origen == nil {
		return
	}

	abiertas := 0
	if origen.SesionesAbiertas != nil {
		abiertas = *origen.SesionesAbiertas
	}

	d.parrafo(tituloDeLasMedidasYReutilizadas)
	d.parrafo("Evals medidas: " + strconv.Itoa(origen.EvalsMedidas))
	d.parrafo("Evals reutilizadas: " + strconv.Itoa(origen.EvalsReutilizadas))
	d.parrafo("Sesiones abiertas: " + strconv.Itoa(abiertas))

	if plataforma := origen.ErrorDeLaPlataforma; plataforma != nil {
		d.parrafo("Error de la plataforma: `" + plataforma.Orden + "`: " + plataforma.Error)
	}

	if len(origen.Evals) > 0 {
		lineas := make([]string, 0, len(origen.Evals))
		for _, eval := range origen.Evals {
			lineas = append(lineas, lineaDelOrigen(eval, origen.ErrorDeLaPlataforma != nil))
		}

		d.lista(lineas)
	}

	var series []TasaDelInforme
	for _, eval := range origen.reutilizadas() {
		series = append(series, eval.Tasas...)
	}

	d.tablaConEtiqueta(etiquetaDeLasSeriesReutilizadas, encabezadosDeTasas, filasDeTasas(series))
}

// lineaDelOrigen es la línea de informe.md de una eval del origen
// (contracts/informe-del-job.md §6 de H30): la de una medida, «medida:» y su
// motivo en palabras, con las rutas que la eligen entre paréntesis si las
// lleva, y, la de una sin resultado reutilizable de main, por qué paró la
// búsqueda si hubo un error de la plataforma; y la de una reutilizada, la
// ejecución de la que viene, con su propuesta, su commit abreviado y su fecha.
func lineaDelOrigen(eval origenDeEval, conErrorDeLaPlataforma bool) string {
	if !eval.Medida {
		return eval.Eval + ": " + eval.Procedencia.escrita()
	}

	enPalabras := motivosEnPalabras[eval.Motivo]
	linea := eval.Eval + ": medida: " + cmp.Or(enPalabras.palabras, string(eval.Motivo))

	if enPalabras.conRutas {
		linea += " (" + strings.Join(eval.Rutas, ", ") + ")"
	}

	if eval.Motivo == motivoSinResultadoDeMain && conErrorDeLaPlataforma {
		linea += medidaPorUnErrorDeLaPlataforma
	}

	return linea
}

// escrita es la procedencia como la escribe la línea de una eval reutilizada:
// «reutilizada de la ejecución <n> (propuesta #<p>, <commit abreviado>,
// <fecha>)»; «reutilizada», sin ella.
func (p *procedenciaDeLaMedida) escrita() string {
	if p == nil {
		return "reutilizada"
	}

	return fmt.Sprintf(reutilizadaDeLaEjecucion, p.Ejecucion, p.Propuesta, p.Commit[:min(len(p.Commit), cifrasDelCommitCorto)],
		p.Fecha)
}
