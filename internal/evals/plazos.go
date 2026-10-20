package evals

import (
	"errors"
	"fmt"
	"regexp"
	"slices"
	"strings"
	"sync"

	"github.com/santhosh-tekuri/jsonschema/v6"

	"github.com/jmorenobl/kitlegal/internal/core/plazos"
)

// Las etiquetas de las tres formas fijas con las que la respuesta de una eval de
// legal-core da un plazo, declara que un asunto no está cubierto
// (contracts/evals-legal-core.md §2 de H9; FR-048) y, desde el ADR 0042, que
// plazos no calcula un plazo. Las de los avisos de un plazo no van aquí: son las
// de plazos.EtiquetaDeAviso.
const (
	etiquetaDelPlazo            = "PLAZO"
	etiquetaDeNoCubierto        = "NO CUBIERTO"
	etiquetaDelPlazoSinCalcular = "PLAZO SIN CALCULAR"
)

// propiedadDelPlazo es la propiedad del esquema de eval con el plazo esperado, y
// avisosDelPlazo, la de sus avisos como la nombra un error.
const (
	propiedadDelPlazo = "plazo"
	avisosDelPlazo    = propiedadDelPlazo + "." + propiedadDeAvisos
)

// Las expresiones con las que se reconocen en una línea de la respuesta las dos
// formas fijas, con la tolerancia de las etiquetas de los avisos
// (patronDeEtiqueta): blancos y énfasis de Markdown alrededor de las partes, el
// selector de presentación de la marca y las minúsculas.
var (
	// formaDelPlazo casa con la etiqueta del plazo seguida, tras esos mismos
	// blancos y énfasis, de una fecha AAAA-MM-DD detrás de la cual no va otra
	// cifra: la que va justo detrás de los dos puntos. El grupo es la fecha.
	formaDelPlazo = regexp.MustCompile(patronDeEtiqueta(etiquetaDelPlazo) + separadorDeAviso +
		`([0-9]{4}-[0-9]{2}-[0-9]{2})(?:[^0-9]|$)`)

	// formaDeNoCubierto casa con la etiqueta de la declaración de no cubierto.
	formaDeNoCubierto = formaFija(etiquetaDeNoCubierto)

	// formaDelPlazoSinCalcular casa con la etiqueta de la declaración de un
	// plazo sin calcular. La del plazo no casa con ella: tras PLAZO pide los dos
	// puntos.
	formaDelPlazoSinCalcular = formaFija(etiquetaDelPlazoSinCalcular)

	// formaDeIdentificador casa con cada identificador BOE-A-… entero de un
	// texto, el de la norma que rige el procedimiento o el de la LPAC.
	formaDeIdentificador = regexp.MustCompile(`\bBOE-A-[0-9]{4}-[0-9]{1,9}\b`)
)

// formasDeAvisoDelPlazo son las expresiones de la forma fija de cada código de
// plazos.CodigosDeAviso, compiladas una sola vez desde plazos.EtiquetaDeAviso
// con formasFijas.
var formasDeAvisoDelPlazo = sync.OnceValue(func() map[string]*regexp.Regexp {
	etiquetas := map[string]string{}

	for _, codigo := range plazos.CodigosDeAviso() {
		if etiqueta, conEtiqueta := plazos.EtiquetaDeAviso(codigo); conEtiqueta {
			etiquetas[codigo] = etiqueta
		}
	}

	return formasFijas(etiquetas)
})

// Principio de los motivos por los que una eval de plazo o de cobertura no
// cumple lo esperado (contracts/evals-legal-core.md §2 de H9), seguidos de la
// fecha, el código o la norma de lo que falta o sobra.
const (
	motivoDePlazoAusente          = "plazo ausente: "
	motivoDeAvisoDelPlazoAusente  = "aviso del plazo ausente: "
	motivoDeNoCubiertoAusente     = "declaración de no cubierto ausente: "
	motivoDeLineaDelPlazoSobrante = "línea del plazo en un asunto no cubierto: "

	motivoDeSinCalcularAusente          = "declaración de plazo sin calcular ausente: "
	motivoDeLineaDelPlazoSinLaRedaccion = "línea del plazo en un plazo sin calcular: "
)

// ExtraerLineasDelPlazo devuelve la fecha de cada línea del plazo de la
// respuesta en su forma fija —la marca, la etiqueta PLAZO y los dos puntos, con
// la tolerancia de las etiquetas de los avisos, la fecha AAAA-MM-DD justo detrás
// y al menos una cita con su forma (formaDeCita) en la misma línea—, en el orden
// en que aparecen y con sus repeticiones, o nil si no lleva ninguna. Una fecha
// detrás de la cita y no de los dos puntos, una línea sin cita o con la cita en
// otra línea, y una fecha suelta en el texto no son líneas del plazo
// (contracts/evals-legal-core.md §2 de H9; FR-048).
func ExtraerLineasDelPlazo(respuesta string) []string {
	var fechas []string

	for linea := range strings.Lines(respuesta) {
		if !formaDeCita.MatchString(linea) {
			continue
		}

		for _, partes := range formaDelPlazo.FindAllStringSubmatch(linea, -1) {
			fechas = append(fechas, partes[1])
		}
	}

	return fechas
}

// ExtraerNoCubierto devuelve cada identificador BOE-A-… de las líneas de la
// respuesta con la forma fija de la declaración de no cubierto —la marca, la
// etiqueta NO CUBIERTO y los dos puntos, con la tolerancia de las etiquetas de
// los avisos—, en el orden en que aparecen y con sus repeticiones, o nil si no
// lleva ninguno. El identificador de otra línea no cuenta, y no lee ninguna otra
// redacción (contracts/evals-legal-core.md §2 de H9; FR-048).
func ExtraerNoCubierto(respuesta string) []string {
	return identificadoresDeLasLineas(respuesta, formaDeNoCubierto)
}

// ExtraerPlazoSinCalcular devuelve cada identificador BOE-A-… de las líneas de
// la respuesta con la forma fija de la declaración de un plazo sin calcular
// —la marca, la etiqueta PLAZO SIN CALCULAR y los dos puntos, con la tolerancia
// de las etiquetas de los avisos—, en el orden en que aparecen y con sus
// repeticiones, o nil si no lleva ninguno, como ExtraerNoCubierto (ADR 0042).
func ExtraerPlazoSinCalcular(respuesta string) []string {
	return identificadoresDeLasLineas(respuesta, formaDelPlazoSinCalcular)
}

// identificadoresDeLasLineas son los identificadores BOE-A-… de las líneas de la
// respuesta que casan con la forma, en su orden y con sus repeticiones.
func identificadoresDeLasLineas(respuesta string, forma *regexp.Regexp) []string {
	var normas []string

	for linea := range strings.Lines(respuesta) {
		if forma.MatchString(linea) {
			normas = append(normas, formaDeIdentificador.FindAllString(linea, -1)...)
		}
	}

	return normas
}

// ExtraerAvisosDelPlazo devuelve los códigos de plazos.CodigosDeAviso cuya forma
// fija —la marca, la etiqueta de plazos.EtiquetaDeAviso y los dos puntos, con
// las tolerancias de la gramática— lleva el texto, en el orden de
// CodigosDeAviso y sin repetir, o nil si no lleva ninguna. No lee lo que sigue a
// la forma (contracts/evals-legal-core.md §2 de H9; FR-048).
func ExtraerAvisosDelPlazo(texto string) []string {
	return extraerFormas(texto, plazos.CodigosDeAviso(), formasDeAvisoDelPlazo())
}

// ComprobarCodigosDeAvisoDelPlazo comprueba que los códigos que el esquema de
// eval compilado admite en plazo.avisos —los del enumerado al que llegan sus
// items y cada $ref— son exactamente los de plazos.CodigosDeAviso, como
// ComprobarCodigosDeAviso los de avisos: devuelve un error por código de plazos
// que el esquema no enumera, en el orden de CodigosDeAviso, y otro por código
// que enumera sin ser de plazos, en el orden del esquema, unidos con
// errors.Join; o nil si coinciden. Un esquema sin la propiedad plazo, sin la de
// sus avisos o sin enumerado no enumera ninguno (contracts/evals-legal-core.md
// §1 de H9; research D18 de H9).
func ComprobarCodigosDeAvisoDelPlazo(esquema *jsonschema.Schema) error {
	var enumerados []any
	if plazo, conPlazo := esquema.Properties[propiedadDelPlazo]; conPlazo {
		enumerados = enumeradoDeLaLista(plazo, propiedadDeAvisos)
	}

	codigos := plazos.CodigosDeAviso()

	var defectos []error

	for _, codigo := range codigos {
		if !slices.Contains(enumerados, any(codigo)) {
			defectos = append(defectos, fmt.Errorf("el esquema de eval no enumera el código de aviso %s en %s",
				codigo, avisosDelPlazo))
		}
	}

	for _, enumerado := range enumerados {
		if codigo, esCadena := enumerado.(string); !esCadena || !slices.Contains(codigos, codigo) {
			defectos = append(defectos, fmt.Errorf("el esquema de eval enumera en %s el código %v, que no es un "+
				"código de aviso de plazos", avisosDelPlazo, enumerado))
		}
	}

	return errors.Join(defectos...)
}

// juzgarElPlazo juzga lo que una eval de plazo o de cobertura espera de la
// respuesta (contracts/evals-legal-core.md §2 de H9; FR-048, FR-051, FR-057).
// En las dos anota las fechas de sus líneas del plazo (ExtraerLineasDelPlazo).
// En una de plazo, anota si alguna lleva la fecha esperada, con el motivo plazo
// ausente si no, y cada aviso esperado cuya forma fija no lleva la respuesta
// (ExtraerAvisosDelPlazo), con el suyo. En una de cobertura, anota si alguna
// declaración de no cubierto lleva la norma esperada (ExtraerNoCubierto), con
// el motivo de la declaración ausente si no, y una línea del plazo, de la fecha
// que sea, lleva el suyo. En una de vigencia, lo mismo con la declaración de un
// plazo sin calcular (ExtraerPlazoSinCalcular; ADR 0042). Las demás evals no se
// juzgan aquí y dejan vacío todo.
func (r *ResultadoDeEval) juzgarElPlazo(eval Eval, respuesta string) {
	if !eval.esDePlazo() && !eval.esDeCobertura() && !eval.esDeVigencia() {
		return
	}

	r.LineasDelPlazo = ExtraerLineasDelPlazo(respuesta)

	if eval.esDePlazo() {
		r.PlazoEncontrado = slices.Contains(r.LineasDelPlazo, eval.Plazo.Fecha)
		if !r.PlazoEncontrado {
			r.Motivos = append(r.Motivos, motivoDePlazoAusente+eval.Plazo.Fecha)
		}

		_, r.AvisosDelPlazoAusentes = r.repartirFormas(eval.Plazo.Avisos, ExtraerAvisosDelPlazo(respuesta),
			motivoDeAvisoDelPlazoAusente)
	}

	if eval.esDeCobertura() {
		r.NoCubiertoDeclarado = slices.Contains(ExtraerNoCubierto(respuesta), eval.NoCubierto.Norma)
		if !r.NoCubiertoDeclarado {
			r.Motivos = append(r.Motivos, motivoDeNoCubiertoAusente+eval.NoCubierto.Norma)
		}

		for _, fecha := range r.LineasDelPlazo {
			r.Motivos = append(r.Motivos, motivoDeLineaDelPlazoSobrante+fecha)
		}
	}

	if eval.esDeVigencia() {
		r.PlazoSinCalcularDeclarado = slices.Contains(ExtraerPlazoSinCalcular(respuesta), eval.PlazoSinCalcular.Norma)
		if !r.PlazoSinCalcularDeclarado {
			r.Motivos = append(r.Motivos, motivoDeSinCalcularAusente+eval.PlazoSinCalcular.Norma)
		}

		for _, fecha := range r.LineasDelPlazo {
			r.Motivos = append(r.Motivos, motivoDeLineaDelPlazoSinLaRedaccion+fecha)
		}
	}
}

// cumpleElPlazo dice si la respuesta cumple lo que la eval espera del plazo: en
// una de plazo, la línea con la fecha esperada y cada aviso esperado; en una de
// cobertura, la declaración con la norma esperada y ninguna línea del plazo; y en
// una de vigencia, la declaración del plazo sin calcular con la norma esperada y
// ninguna línea del plazo. Una eval que no es de plazo, de cobertura ni de
// vigencia lo cumple siempre.
func (r *ResultadoDeEval) cumpleElPlazo(eval Eval) bool {
	return (!eval.esDePlazo() || r.PlazoEncontrado) && len(r.AvisosDelPlazoAusentes) == 0 &&
		(!eval.esDeCobertura() || r.NoCubiertoDeclarado && len(r.LineasDelPlazo) == 0) &&
		(!eval.esDeVigencia() || r.PlazoSinCalcularDeclarado && len(r.LineasDelPlazo) == 0)
}

// pideElPlazo dice si la invocación es la de plazos calcular con el tipo del
// comando (contracts/evals-legal-core.md §1 de H9): su verbo y, como primer
// argumento de posición, ese tipo.
func pideElPlazo(invocacion Invocacion, comando ComandoEsperado) bool {
	return invocacion.Verbo == verboCalcular && tipoDelCalculo(invocacion.Argumentos) == comando.Tipo
}

// tipoDelCalculo es el tipo que pide una invocación de plazos calcular, su primer
// argumento de posición, o vacío si no tiene ninguno. Las dos banderas propias
// del verbo, --duracion y --unidad, llevan valor, pegado con «=» o en el
// argumento siguiente, e InterpretarInvocacion ya quitó las globales; en una
// llamada, cli.LineaDeLlamada da las banderas con su valor pegado delante de los
// argumentos de posición. Lo que va detrás de un «--» suelto son argumentos de
// posición.
func tipoDelCalculo(argumentos []string) string {
	for posicion := 0; posicion < len(argumentos); posicion++ {
		argumento := argumentos[posicion]

		switch {
		case argumento == terminadorDeLaLlamada:
			if posicion+1 < len(argumentos) {
				return argumentos[posicion+1]
			}

			return ""
		case strings.HasPrefix(argumento, prefijoDeBandera):
			if !strings.Contains(argumento, "=") {
				posicion++
			}
		default:
			return argumento
		}
	}

	return ""
}
