package evals

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"slices"
	"strings"
	"testing"

	"github.com/santhosh-tekuri/jsonschema/v6"
	"github.com/santhosh-tekuri/jsonschema/v6/kind"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.yaml.in/yaml/v3"

	"github.com/jmorenobl/kitlegal/internal/core/plazos"
	"github.com/jmorenobl/kitlegal/internal/skills"
	"github.com/jmorenobl/kitlegal/internal/source/boe"
)

// nombreDeEval es el nombre de fichero con el que los tests leen cada eval
// sintética, y por el que tiene que empezar todo error (US4, escenario 4).
const nombreDeEval = "01-lpac-articulo-21.yaml"

// Trozos de las evals sintéticas de TestLeerEval, cada uno con sus líneas
// completas: la pregunta del contrato evals-y-grabaciones §1, el comando y la
// cita del artículo 21 de la LPAC y los dos avisos de una norma derogada.
const (
	preguntaDelArticulo21 = "pregunta: \"¿qué dice el art. 21 de la Ley 39/2015?\"\n"
	comandoDelArticulo21  = "comandos:\n" +
		"  - applet: boe\n" +
		"    norma: BOE-A-2015-10565\n" +
		"    bloque: a21\n"
	citaDelArticulo21 = "citas:\n" +
		"  - norma: BOE-A-2015-10565\n" +
		"    bloque: a21\n"
	avisosDeNormaDerogada = "avisos:\n" +
		"  - derogada\n" +
		"  - vigencia-agotada\n"
)

// Trozos de las evals sintéticas de territorio de TestLeerEval y
// TestFormaDelComando, cada uno con sus líneas completas: la pregunta de un
// municipio, su comando de territorio y el esperado con las cuatro claves (contrato
// de evals §1.1 y §1.2 de H6).
const (
	preguntaDelMunicipio = "pregunta: \"¿En qué boletines se publican las normas que afectan a Leganés?\"\n"
	comandoDelMunicipio  = "comandos:\n" +
		"  - applet: territorio\n" +
		"    verbo: resolver\n" +
		"    municipio: Leganés\n"
	territorioDelMunicipio = "territorio:\n" +
		"  comunidad: Comunidad de Madrid\n" +
		"  provincia: Madrid\n" +
		"  boletines: [BOCM]\n" +
		"  cobertura:\n" +
		"    - \"boletin_autonomico: configurado\"\n" +
		"    - \"dir3: verificado\"\n"
)

// Trozos de las evals sintéticas del formato ampliado de TestLeerEval y
// TestFormaDelComando, cada uno con sus líneas completas: el grafo previo con el
// bloque del art. 21 de la LPAC, el comando de comprobación de graph —sin la
// línea comandos:, para ir detrás de otro comando— y el prohibido graph show
// (contrato evals-y-skill §1 y §4 de H7).
const (
	grafoPrevioDelArticulo21 = "grafo_previo:\n" +
		"  grabaciones: lcsp-a1-30-redaccion-original\n" +
		"  comandos:\n" +
		"    - applet: boe\n" +
		"      norma: BOE-A-2015-10565\n" +
		"      bloque: a21\n"
	comprobacionDelGrafo = "  - applet: graph\n" +
		"    verbo: check\n"
	prohibidoGraphShow = "prohibidos:\n" +
		"  - applet: graph\n" +
		"    verbo: show\n"
)

// Trozos de las evals sintéticas del formato de H7.1 de TestLeerEval, cada uno
// con sus líneas completas: la norma que consulta el comando de comprobación
// —para ir detrás de comprobacionDelGrafo— y el hallazgo version-obsoleta
// esperado (contrato evals-y-skill §1 y §5 de H7.1).
const (
	normaDeLaComprobacion     = "    norma: BOE-A-2015-10565\n"
	hallazgoDeVersionObsoleta = "hallazgos:\n" +
		"  - version-obsoleta\n"
)

// Trozos de las evals sintéticas del formato de H7.4 de TestLeerEval y
// TestLeerConjunto, cada uno con sus líneas completas: una eval de no activación
// y la skill que la sesión no puede activar, para ir detrás de activa, como en
// las evals de legal-core (contracts/evals-y-juicio.md §1 y §5 de H7.4; FR-003,
// FR-004); y la eval de los dos bloques de la LCSP de contracts §3, sin su
// comentario, con las dos redacciones modificadas que la respuesta tiene que
// llevar, cada una con la fecha de la redacción superada y la de la leída
// (FR-053).
const (
	evalDeLaReceta                 = "pregunta: \"¿Cómo se hace una tortilla de patatas jugosa?\"\nactiva: false\n"
	noSeActivaBoeLegislacion       = "no_se_activan: [boe-legislacion]\n"
	redaccionesModificadasDeLaLCSP = "redacciones_modificadas:\n" +
		"  - norma: BOE-A-2017-12902\n" +
		"    bloque: a1-30\n" +
		"    fecha_vigencia: \"20180309\"\n" +
		"    fecha_vigencia_reciente: \"20200206\"\n" +
		"  - norma: BOE-A-2017-12902\n" +
		"    bloque: da-3\n" +
		"    fecha_vigencia: \"20180309\"\n" +
		"    fecha_vigencia_reciente: \"20230101\"\n"
	evalDeDosBloquesDeLaLCSP = "pregunta: \"¿Qué dicen ahora el artículo 118 y la disposición adicional tercera de la LCSP?\"\n" +
		"activa: true\n" +
		"informativa: true\n" +
		"grafo_previo:\n" +
		"  grabaciones: lcsp-a1-30-y-da-3-redaccion-original\n" +
		"  comandos:\n" +
		"    - applet: boe\n      norma: BOE-A-2017-12902\n      bloque: a1-30\n" +
		"    - applet: boe\n      norma: BOE-A-2017-12902\n      bloque: da-3\n" +
		"comandos:\n" +
		"  - applet: boe\n    norma: BOE-A-2017-12902\n    bloque: a1-30\n" +
		"  - applet: boe\n    norma: BOE-A-2017-12902\n    bloque: da-3\n" +
		"  - applet: graph\n    verbo: check\n    norma: BOE-A-2017-12902\n" +
		prohibidoGraphShow +
		"citas:\n" +
		"  - norma: BOE-A-2017-12902\n    bloque: a1-30\n" +
		"  - norma: BOE-A-2017-12902\n    bloque: da-3\n" +
		hallazgoDeVersionObsoleta + redaccionesModificadasDeLaLCSP
)

// Las dos evals sin binario ni servidor de contracts/evals-en-dos-modos.md §1 de
// H21, sin su comentario y con sus líneas completas, y la pregunta de cada una
// tal como se lee: la de boe-legislacion y la de legal-core, que además declara
// la skill que su sesión no puede activar (FR-046).
const (
	preguntaSinBinarioDeBoeLegislacion = "\xc2\xbfQu\xc3\xa9 dice el art. 53 de la Ley 39/2015 sobre los derechos " +
		"del interesado en el procedimiento administrativo?"
	preguntaSinBinarioDeLegalCore = "\xc2\xbfQu\xc3\xa9 comunidad aut\xc3\xb3noma, provincia y boletines oficiales " +
		"corresponden al Ayuntamiento de Getafe?"
	sinBinarioNiServidor           = "sin_binario_ni_servidor: true\n"
	evalSinBinarioDeBoeLegislacion = "pregunta: \"" + preguntaSinBinarioDeBoeLegislacion + "\"\n" +
		"activa: true\n" + sinBinarioNiServidor
	evalSinBinarioDeLegalCore = "pregunta: \"" + preguntaSinBinarioDeLegalCore + "\"\n" +
		"activa: true\n" + noSeActivaBoeLegislacion + sinBinarioNiServidor
)

// leidaDeDosBloquesDeLaLCSP es lo que se lee de evalDeDosBloquesDeLaLCSP con el
// nombre de fichero dado: sus dos bloques en el grafo previo y en los comandos,
// la comprobación de la norma, graph show prohibido, las dos citas, el hallazgo
// version-obsoleta y las dos redacciones modificadas, en el orden del fichero.
func leidaDeDosBloquesDeLaLCSP(fichero string) Eval {
	const lcsp = "BOE-A-2017-12902"

	bloques := []ComandoEsperado{{Applet: "boe", Norma: lcsp, Bloque: "a1-30"}, {Applet: "boe", Norma: lcsp, Bloque: "da-3"}}

	return Eval{
		Fichero:     fichero,
		Pregunta:    "¿Qué dicen ahora el artículo 118 y la disposición adicional tercera de la LCSP?",
		Activa:      true,
		Informativa: true,
		GrafoPrevio: GrafoPrevio{Grabaciones: "lcsp-a1-30-y-da-3-redaccion-original", Comandos: bloques},
		Comandos:    slices.Concat(bloques, []ComandoEsperado{{Applet: "graph", Verbo: "check", Norma: lcsp}}),
		Prohibidos:  []ComandoProhibido{{Applet: "graph", Verbo: "show"}},
		Citas:       []CitaEsperada{{Norma: lcsp, Bloque: "a1-30"}, {Norma: lcsp, Bloque: "da-3"}},
		Hallazgos:   []string{"version-obsoleta"},
		RedaccionesModificadas: []RedaccionEsperada{
			{Norma: lcsp, Bloque: "a1-30", FechaVigencia: "20180309", FechaVigenciaReciente: "20200206"},
			{Norma: lcsp, Bloque: "da-3", FechaVigencia: "20180309", FechaVigenciaReciente: "20230101"},
		},
	}
}

// TestLeerEval fija la lectura de una eval del contrato evals-y-grabaciones §1 y
// de sus avisos (contrato de formato, juicio e informe §6): las tres formas de
// comando, con y sin reproduce, la que espera avisos, informativa o no, y la de no
// activación se leen enteras y con su Fichero; avisos vacío o con un código
// repetido se acepta o se rechaza igual que citas; cada fichero inválido, con una
// clave repetida incluida, da un error que empieza por su nombre y dice qué falla
// y dónde.
//
// Desde H6, la cuarta forma de comando, la de territorio, se lee con su Municipio,
// y el esperado de territorio, con sus cuatro claves en Territorio: con él, una
// eval activa sin citas es válida, y con citas también; un territorio vacío, con
// una clave o un aspecto de cobertura que el formato no tiene, o en una eval de
// no activación, y un comando de territorio sin municipio o con una norma, se
// rechazan (contrato de evals §1 y §2 de H6; FR-084).
//
// Desde H7, la eval de la consulta repetida se lee entera: el grafo previo con
// sus grabaciones y sus comandos de bloque en GrafoPrevio, el comando de
// comprobación en Comandos con su Verbo y los prohibidos en Prohibidos, y los
// prohibidos también sin grafo previo; un comando de comprobación con otro verbo
// o sin applet, unos prohibidos vacíos, sin verbo, con un verbo que
// no es de minúsculas o con otra clave, un grafo previo sin grabaciones, con un
// nombre de grabaciones que no es de minúsculas y guiones, sin comandos, con los
// comandos vacíos, con un comando que no es de bloque o con otra clave, y unos
// prohibidos o un grafo previo en una eval de no activación se rechazan
// (contrato evals-y-skill §1 de H7; FR-085, FR-086).
//
// Desde H7.1, el comando de comprobación se lee también con la norma que
// consulta, en su Norma, y sigue sin admitir bloques; los hallazgos esperados se
// leen en Hallazgos, también en la eval de la redacción cambiada entera; y unos
// hallazgos vacíos, con una clase que el formato no tiene —fuente-caducada, que
// ninguna skill traslada con forma fija— o en una eval de no activación se
// rechazan, como los avisos (contrato evals-y-skill §1 y §5 de H7.1; FR-050,
// FR-054).
//
// Desde H7.4, las skills que la sesión no puede activar se leen en NoSeActivan,
// en una eval que activa la skill y en una de no activación; y las redacciones
// modificadas esperadas, en RedaccionesModificadas, en la eval de los dos
// bloques de la LCSP entera. Unas skills que no se activan vacías, con un nombre
// que no es de skill o con uno repetido, unas redacciones modificadas en una
// eval de no activación, y una sin fecha_vigencia_reciente, con una fecha de
// siete cifras o con el mes 13 se rechazan: cada documento es el bien formado
// con solo ese defecto (contracts/evals-y-juicio.md §1; FR-003, FR-004, FR-053).
//
// Desde H21, la eval sin binario ni servidor se lee con SinBinarioNiServidor,
// con activa: true y nada más, y también con las skills que su sesión no puede
// activar: las dos de contracts/evals-en-dos-modos.md §1. Con comandos, con
// citas, con informativa, en una eval de no activación o con el valor false se
// rechaza (contracts/evals-en-dos-modos.md §1 y §8; FR-046).
func TestLeerEval(t *testing.T) {
	t.Parallel()

	citaDeLaLRBRL := []CitaEsperada{{Norma: "BOE-A-1985-5392", Bloque: "a85bis."}}

	// La eval de territorio del municipio cubierto y lo que se lee de su comando y
	// de su esperado.
	positivaDelMunicipio := preguntaDelMunicipio + "activa: true\n" + comandoDelMunicipio
	preguntaDelMunicipioLeida := "¿En qué boletines se publican las normas que afectan a Leganés?"
	comandoDelMunicipioLeido := []ComandoEsperado{{Applet: "territorio", Verbo: "resolver", Municipio: "Leganés"}}
	territorioDelMunicipioLeido := TerritorioEsperado{
		Comunidad: "Comunidad de Madrid",
		Provincia: "Madrid",
		Boletines: []string{"BOCM"},
		Cobertura: []string{"boletin_autonomico: configurado", "dir3: verificado"},
	}

	// patronDeFecha es el de una fecha AAAAMMDD de una redacción modificada, tal
	// como lo cita el rechazo (contracts/evals-y-juicio.md §1 de H7.4).
	const patronDeFecha = "'^[0-9]{4}(0[1-9]|1[0-2])(0[1-9]|[12][0-9]|3[01])$'"

	// La positiva del art. 21, sin avisos, y lo que se lee de su comando y de su
	// cita: los casos de avisos le añaden líneas al final.
	positivaDelArticulo21 := preguntaDelArticulo21 + "activa: true\n" + comandoDelArticulo21 + citaDelArticulo21
	comandoDelArticulo21Leido := []ComandoEsperado{{Applet: "boe", Norma: "BOE-A-2015-10565", Bloque: "a21"}}
	citaDelArticulo21Leida := []CitaEsperada{{Norma: "BOE-A-2015-10565", Bloque: "a21"}}

	// La eval de territorio del municipio cubierto con las skills que no se
	// activan en la línea que se le pase, detrás de activa, como en las evals de
	// legal-core: esa línea es la 3.
	municipioQueNoActiva := func(noSeActivan string) string {
		return preguntaDelMunicipio + "activa: true\n" + noSeActivan + comandoDelMunicipio + territorioDelMunicipio
	}

	casos := []struct {
		nombre     string
		documento  string
		leida      Eval
		error      string
		fragmentos []string
	}{
		{
			nombre: "comando-de-bloque",
			documento: "# Eval positiva: debe activar la skill, leer el bloque y citarlo.\n" +
				preguntaDelArticulo21 + "activa: true\n" + comandoDelArticulo21 + citaDelArticulo21,
			leida: Eval{
				Fichero:  nombreDeEval,
				Pregunta: "¿qué dice el art. 21 de la Ley 39/2015?",
				Activa:   true,
				Comandos: []ComandoEsperado{{Applet: "boe", Norma: "BOE-A-2015-10565", Bloque: "a21"}},
				Citas:    []CitaEsperada{{Norma: "BOE-A-2015-10565", Bloque: "a21"}},
			},
		},
		{
			nombre: "comandos-de-consulta-de-norma",
			documento: "pregunta: \"¿Qué servicios presta un municipio de más de 5000 habitantes?\"\n" +
				"activa: true\n" +
				"comandos:\n" +
				"  - applet: boe\n    verbo: indice\n    norma: BOE-A-1985-5392\n" +
				"  - applet: boe\n    verbo: metadatos\n    norma: BOE-A-1985-5392\n" +
				"  - applet: boe\n    verbo: analisis\n    norma: BOE-A-1985-5392\n" +
				"citas:\n  - norma: BOE-A-1985-5392\n    bloque: a85bis.\n",
			leida: Eval{
				Fichero:  nombreDeEval,
				Pregunta: "¿Qué servicios presta un municipio de más de 5000 habitantes?",
				Activa:   true,
				Comandos: []ComandoEsperado{
					{Applet: "boe", Verbo: "indice", Norma: "BOE-A-1985-5392"},
					{Applet: "boe", Verbo: "metadatos", Norma: "BOE-A-1985-5392"},
					{Applet: "boe", Verbo: "analisis", Norma: "BOE-A-1985-5392"},
				},
				Citas: citaDeLaLRBRL,
			},
		},
		{
			nombre: "comando-de-busqueda-con-reproduce",
			documento: "pregunta: \"¿Qué ley regula las bases del régimen local?\"\n" +
				"activa: true\n" +
				"reproduce: boe-fiscal\n" +
				"comandos:\n  - applet: boe\n    verbo: buscar\n    terminos: [bases, régimen, local]\n" +
				"citas:\n  - norma: BOE-A-1985-5392\n    bloque: a85bis.\n",
			leida: Eval{
				Fichero:   nombreDeEval,
				Pregunta:  "¿Qué ley regula las bases del régimen local?",
				Activa:    true,
				Reproduce: "boe-fiscal",
				Comandos:  []ComandoEsperado{{Applet: "boe", Verbo: "buscar", Terminos: []string{"bases", "régimen", "local"}}},
				Citas:     citaDeLaLRBRL,
			},
		},
		{
			nombre: "no-activa",
			documento: "# Eval de no activación: pregunta ajena a la normativa del BOE.\n" +
				"pregunta: \"¿Cómo invierto una lista enlazada en Go?\"\n" +
				"activa: false\n",
			leida: Eval{Fichero: nombreDeEval, Pregunta: "¿Cómo invierto una lista enlazada en Go?"},
		},
		{
			nombre:    "sin-pregunta",
			documento: "activa: false\n",
			error:     nombreDeEval + ": línea 1: missing property 'pregunta'",
		},
		{
			nombre:    "sin-activa",
			documento: preguntaDelArticulo21,
			error:     nombreDeEval + ": línea 1: missing property 'activa'",
		},
		{
			// Sin citas, territorio ni, desde H9, plazo, ni, desde el ADR 0042,
			// plazo_sin_calcular fallan las cuatro ramas del anyOf del esperado
			// verificable, y el defecto es una hoja de cada una.
			nombre:    "positiva-sin-esperado-verificable",
			documento: preguntaDelArticulo21 + "activa: true\n" + comandoDelArticulo21,
			error: nombreDeEval + ": línea 1: missing property 'citas'\n" +
				"línea 1: missing property 'plazo'\n" +
				"línea 1: missing property 'plazo_sin_calcular'\n" +
				"línea 1: missing property 'territorio'",
		},
		{
			nombre:    "no-activa-con-comandos",
			documento: preguntaDelArticulo21 + "activa: false\n" + comandoDelArticulo21,
			error:     nombreDeEval + ": línea 1: 'not' failed",
		},
		{
			nombre: "comando-con-verbo-y-bloque",
			documento: preguntaDelArticulo21 + "activa: true\n" +
				"comandos:\n" +
				"  - applet: boe\n" +
				"    verbo: indice\n" +
				"    norma: BOE-A-2015-10565\n" +
				"    bloque: a21\n" +
				citaDelArticulo21,
			fragmentos: []string{
				"comandos/0, línea 4: additional properties 'verbo' not allowed",
				"comandos/0, línea 4: additional properties 'bloque' not allowed",
			},
		},
		{
			nombre: "cita-sin-identificador-valido",
			documento: preguntaDelArticulo21 + "activa: true\n" + comandoDelArticulo21 +
				"citas:\n  - norma: BOE-A-15-1\n    bloque: a21\n",
			error: nombreDeEval + ": citas/0/norma, línea 8: " +
				"'BOE-A-15-1' does not match pattern '^BOE-A-[0-9]{4}-[0-9]{1,9}$'",
		},
		{
			nombre: "bloque-con-punto-inicial",
			documento: preguntaDelArticulo21 + "activa: true\n" + comandoDelArticulo21 +
				"citas:\n  - norma: BOE-A-2015-10565\n    bloque: .a1\n",
			error: nombreDeEval + ": citas/0/bloque, línea 9: " +
				"'.a1' does not match pattern '^[A-Za-z0-9][A-Za-z0-9.-]{0,63}$'",
		},
		{
			nombre:    "clave-desconocida",
			documento: preguntaDelArticulo21 + "activa: false\nmateria: programación\n",
			error:     nombreDeEval + ": línea 1: additional properties 'materia' not allowed",
		},
		{
			nombre:    "activa-repetida",
			documento: preguntaDelArticulo21 + "activa: true\nactiva: false\n",
			error:     nombreDeEval + ": activa repetido en las líneas 2 y 3",
		},
		{
			nombre: "bloque-repetido-en-una-cita",
			documento: preguntaDelArticulo21 + "activa: true\n" + comandoDelArticulo21 +
				"citas:\n  - norma: BOE-A-2015-10565\n    bloque: a21\n    bloque: a22\n",
			error: nombreDeEval + ": citas/0: bloque repetido en las líneas 9 y 10",
		},
		{
			nombre:    "avisos",
			documento: positivaDelArticulo21 + avisosDeNormaDerogada,
			leida: Eval{
				Fichero:  nombreDeEval,
				Pregunta: "¿qué dice el art. 21 de la Ley 39/2015?",
				Activa:   true,
				Comandos: comandoDelArticulo21Leido,
				Citas:    citaDelArticulo21Leida,
				Avisos:   []string{"derogada", "vigencia-agotada"},
			},
		},
		{
			nombre:    "aviso-desconocido",
			documento: positivaDelArticulo21 + "avisos:\n  - otro\n",
			error: nombreDeEval + ": avisos/0, línea 11: " +
				"value must be one of 'consolidacion-no-finalizada', 'derogada', 'vigencia-agotada'",
		},
		{
			nombre:    "no-activa-con-avisos",
			documento: preguntaDelArticulo21 + "activa: false\n" + avisosDeNormaDerogada,
			error:     nombreDeEval + ": línea 1: 'not' failed",
		},
		{
			nombre:    "avisos-vacio",
			documento: positivaDelArticulo21 + "avisos: []\n",
			error:     nombreDeEval + ": avisos, línea 10: minItems: got 0, want 1",
		},
		{
			nombre:    "citas-vacio",
			documento: preguntaDelArticulo21 + "activa: true\n" + comandoDelArticulo21 + "citas: []\n",
			error:     nombreDeEval + ": citas, línea 7: minItems: got 0, want 1",
		},
		{
			nombre:    "aviso-repetido",
			documento: positivaDelArticulo21 + "avisos:\n  - derogada\n  - derogada\n",
			leida: Eval{
				Fichero:  nombreDeEval,
				Pregunta: "¿qué dice el art. 21 de la Ley 39/2015?",
				Activa:   true,
				Comandos: comandoDelArticulo21Leido,
				Citas:    citaDelArticulo21Leida,
				Avisos:   []string{"derogada", "derogada"},
			},
		},
		{
			nombre:    "cita-repetida",
			documento: positivaDelArticulo21 + "  - norma: BOE-A-2015-10565\n    bloque: a21\n",
			leida: Eval{
				Fichero:  nombreDeEval,
				Pregunta: "¿qué dice el art. 21 de la Ley 39/2015?",
				Activa:   true,
				Comandos: comandoDelArticulo21Leido,
				Citas:    slices.Concat(citaDelArticulo21Leida, citaDelArticulo21Leida),
			},
		},
		{
			nombre: "informativa-con-avisos",
			documento: preguntaDelArticulo21 + "activa: true\ninformativa: true\n" +
				comandoDelArticulo21 + citaDelArticulo21 + avisosDeNormaDerogada,
			leida: Eval{
				Fichero:     nombreDeEval,
				Pregunta:    "¿qué dice el art. 21 de la Ley 39/2015?",
				Activa:      true,
				Informativa: true,
				Comandos:    comandoDelArticulo21Leido,
				Citas:       citaDelArticulo21Leida,
				Avisos:      []string{"derogada", "vigencia-agotada"},
			},
		},
		{
			nombre:    "comando-de-territorio",
			documento: positivaDelMunicipio + territorioDelMunicipio,
			leida: Eval{
				Fichero:    nombreDeEval,
				Pregunta:   preguntaDelMunicipioLeida,
				Activa:     true,
				Comandos:   comandoDelMunicipioLeido,
				Territorio: territorioDelMunicipioLeido,
			},
		},
		{
			nombre: "territorio-con-citas",
			documento: preguntaDelArticulo21 + "activa: true\n" + comandoDelArticulo21 +
				"  - applet: territorio\n    verbo: resolver\n    municipio: Leganés\n" +
				citaDelArticulo21 + "territorio:\n  comunidad: Comunidad de Madrid\n",
			leida: Eval{
				Fichero:    nombreDeEval,
				Pregunta:   "¿qué dice el art. 21 de la Ley 39/2015?",
				Activa:     true,
				Comandos:   slices.Concat(comandoDelArticulo21Leido, comandoDelMunicipioLeido),
				Citas:      citaDelArticulo21Leida,
				Territorio: TerritorioEsperado{Comunidad: "Comunidad de Madrid"},
			},
		},
		{
			nombre:    "territorio-vacio",
			documento: positivaDelMunicipio + "territorio: {}\n",
			error:     nombreDeEval + ": territorio, línea 7: minProperties: got 0, want 1",
		},
		{
			nombre:    "territorio-con-clave-desconocida",
			documento: positivaDelMunicipio + "territorio:\n  municipio: Leganés\n",
			error:     nombreDeEval + ": territorio, línea 8: additional properties 'municipio' not allowed",
		},
		{
			nombre:    "aspecto-de-cobertura-desconocido",
			documento: positivaDelMunicipio + "territorio:\n  cobertura:\n    - \"boletin_autonomico: si\"\n",
			error: nombreDeEval + ": territorio/cobertura/0, línea 9: value must be one of " +
				"'boletin_autonomico: configurado', 'boletin_autonomico: no-configurado', " +
				"'boletin_provincial: configurado', 'boletin_provincial: no-configurado', " +
				"'dir3: verificado', 'dir3: no-verificado'",
		},
		{
			nombre:    "no-activa-con-territorio",
			documento: preguntaDelMunicipio + "activa: false\n" + territorioDelMunicipio,
			error:     nombreDeEval + ": línea 1: 'not' failed",
		},
		{
			nombre: "comando-de-territorio-sin-municipio",
			documento: preguntaDelMunicipio + "activa: true\n" +
				"comandos:\n  - applet: territorio\n    verbo: resolver\n" + territorioDelMunicipio,
			fragmentos: []string{"comandos/0, línea 4: missing property 'municipio'"},
		},
		{
			nombre:     "comando-de-territorio-con-norma",
			documento:  positivaDelMunicipio + "    norma: BOE-A-2015-10565\n" + territorioDelMunicipio,
			fragmentos: []string{"comandos/0, línea 4: additional properties 'norma' not allowed"},
		},
		{
			// La eval del contrato evals-y-skill §4 de H7, sin su comentario.
			nombre: "consulta-repetida",
			documento: preguntaDelArticulo21 + "activa: true\ninformativa: true\n" + grafoPrevioDelArticulo21 +
				comandoDelArticulo21 + comprobacionDelGrafo + prohibidoGraphShow + citaDelArticulo21,
			leida: Eval{
				Fichero:     nombreDeEval,
				Pregunta:    "¿qué dice el art. 21 de la Ley 39/2015?",
				Activa:      true,
				Informativa: true,
				GrafoPrevio: GrafoPrevio{Grabaciones: "lcsp-a1-30-redaccion-original", Comandos: comandoDelArticulo21Leido},
				Comandos: slices.Concat(comandoDelArticulo21Leido,
					[]ComandoEsperado{{Applet: "graph", Verbo: "check"}}),
				Prohibidos: []ComandoProhibido{{Applet: "graph", Verbo: "show"}},
				Citas:      citaDelArticulo21Leida,
			},
		},
		{
			nombre:    "prohibidos-sin-grafo-previo",
			documento: positivaDelArticulo21 + prohibidoGraphShow + "  - applet: boe\n    verbo: buscar\n",
			leida: Eval{
				Fichero:    nombreDeEval,
				Pregunta:   "¿qué dice el art. 21 de la Ley 39/2015?",
				Activa:     true,
				Comandos:   comandoDelArticulo21Leido,
				Prohibidos: []ComandoProhibido{{Applet: "graph", Verbo: "show"}, {Applet: "boe", Verbo: "buscar"}},
				Citas:      citaDelArticulo21Leida,
			},
		},
		{
			nombre: "comprobacion-con-otro-verbo",
			documento: preguntaDelArticulo21 + "activa: true\n" + comandoDelArticulo21 +
				"  - applet: graph\n    verbo: stats\n" + citaDelArticulo21,
			fragmentos: []string{"comandos/1/verbo, línea 8: value must be 'check'"},
		},
		{
			nombre: "comprobacion-con-norma",
			documento: preguntaDelArticulo21 + "activa: true\n" + comandoDelArticulo21 + comprobacionDelGrafo +
				normaDeLaComprobacion + citaDelArticulo21,
			leida: Eval{
				Fichero:  nombreDeEval,
				Pregunta: "¿qué dice el art. 21 de la Ley 39/2015?",
				Activa:   true,
				Comandos: slices.Concat(comandoDelArticulo21Leido,
					[]ComandoEsperado{{Applet: "graph", Verbo: "check", Norma: "BOE-A-2015-10565"}}),
				Citas: citaDelArticulo21Leida,
			},
		},
		{
			nombre: "comprobacion-con-bloque",
			documento: preguntaDelArticulo21 + "activa: true\n" + comandoDelArticulo21 + comprobacionDelGrafo +
				"    norma: BOE-A-2015-10565\n    bloque: a21\n" + citaDelArticulo21,
			fragmentos: []string{"comandos/1, línea 7: additional properties 'bloque' not allowed"},
		},
		{
			nombre: "comprobacion-sin-applet",
			documento: preguntaDelArticulo21 + "activa: true\n" + comandoDelArticulo21 + "  - verbo: check\n" +
				citaDelArticulo21,
			fragmentos: []string{"comandos/1, línea 7: missing property 'applet'"},
		},
		{
			nombre:    "prohibidos-vacio",
			documento: positivaDelArticulo21 + "prohibidos: []\n",
			error:     nombreDeEval + ": prohibidos, línea 10: minItems: got 0, want 1",
		},
		{
			nombre:    "prohibido-sin-verbo",
			documento: positivaDelArticulo21 + "prohibidos:\n  - applet: graph\n",
			error:     nombreDeEval + ": prohibidos/0, línea 11: missing property 'verbo'",
		},
		{
			nombre:    "prohibido-con-verbo-mal-formado",
			documento: positivaDelArticulo21 + "prohibidos:\n  - applet: graph\n    verbo: Show\n",
			error:     nombreDeEval + ": prohibidos/0/verbo, línea 12: 'Show' does not match pattern '^[a-z]+$'",
		},
		{
			nombre:    "prohibido-con-norma",
			documento: positivaDelArticulo21 + prohibidoGraphShow + "    norma: BOE-A-2015-10565\n",
			error:     nombreDeEval + ": prohibidos/0, línea 11: additional properties 'norma' not allowed",
		},
		{
			nombre: "grafo-previo-sin-grabaciones",
			documento: positivaDelArticulo21 + "grafo_previo:\n  comandos:\n" +
				"    - applet: boe\n      norma: BOE-A-2015-10565\n      bloque: a21\n",
			error: nombreDeEval + ": grafo_previo, línea 11: missing property 'grabaciones'",
		},
		{
			nombre: "grafo-previo-con-grabaciones-mal-formadas",
			documento: positivaDelArticulo21 +
				strings.Replace(grafoPrevioDelArticulo21, "lcsp-a1-30-redaccion-original", "../lcsp-a1-30", 1),
			error: nombreDeEval + ": grafo_previo/grabaciones, línea 11: " +
				"'../lcsp-a1-30' does not match pattern '^[a-z0-9]+(-[a-z0-9]+)*$'",
		},
		{
			nombre:    "grafo-previo-sin-comandos",
			documento: positivaDelArticulo21 + "grafo_previo:\n  grabaciones: lcsp-a1-30-redaccion-original\n",
			error:     nombreDeEval + ": grafo_previo, línea 11: missing property 'comandos'",
		},
		{
			nombre:    "grafo-previo-con-comandos-vacio",
			documento: positivaDelArticulo21 + "grafo_previo:\n  grabaciones: lcsp-a1-30-redaccion-original\n  comandos: []\n",
			error:     nombreDeEval + ": grafo_previo/comandos, línea 12: minItems: got 0, want 1",
		},
		{
			nombre: "grafo-previo-con-comando-de-comprobacion",
			documento: positivaDelArticulo21 + "grafo_previo:\n  grabaciones: lcsp-a1-30-redaccion-original\n  comandos:\n" +
				"    - applet: graph\n      verbo: check\n",
			fragmentos: []string{"grafo_previo/comandos/0, línea 13: additional properties 'verbo' not allowed"},
		},
		{
			nombre:    "grafo-previo-con-clave-desconocida",
			documento: positivaDelArticulo21 + grafoPrevioDelArticulo21 + "  skill: boe-legislacion\n",
			error:     nombreDeEval + ": grafo_previo, línea 11: additional properties 'skill' not allowed",
		},
		{
			nombre:    "no-activa-con-prohibidos",
			documento: preguntaDelArticulo21 + "activa: false\n" + prohibidoGraphShow,
			error:     nombreDeEval + ": línea 1: 'not' failed",
		},
		{
			nombre:    "no-activa-con-grafo-previo",
			documento: preguntaDelArticulo21 + "activa: false\n" + grafoPrevioDelArticulo21,
			error:     nombreDeEval + ": línea 1: 'not' failed",
		},
		{
			nombre:    "hallazgos",
			documento: positivaDelArticulo21 + hallazgoDeVersionObsoleta,
			leida: Eval{
				Fichero:   nombreDeEval,
				Pregunta:  preguntaDelArticulo21Eval,
				Activa:    true,
				Comandos:  comandoDelArticulo21Leido,
				Citas:     citaDelArticulo21Leida,
				Hallazgos: []string{"version-obsoleta"},
			},
		},
		{
			// La eval del contrato evals-y-skill §5 de H7.1, sin su comentario.
			nombre: "redaccion-cambiada",
			documento: preguntaDelArticulo21 + "activa: true\ninformativa: true\n" + grafoPrevioDelArticulo21 +
				comandoDelArticulo21 + comprobacionDelGrafo + normaDeLaComprobacion + prohibidoGraphShow +
				citaDelArticulo21 + hallazgoDeVersionObsoleta,
			leida: Eval{
				Fichero:     nombreDeEval,
				Pregunta:    preguntaDelArticulo21Eval,
				Activa:      true,
				Informativa: true,
				GrafoPrevio: GrafoPrevio{Grabaciones: "lcsp-a1-30-redaccion-original", Comandos: comandoDelArticulo21Leido},
				Comandos: slices.Concat(comandoDelArticulo21Leido,
					[]ComandoEsperado{{Applet: "graph", Verbo: "check", Norma: "BOE-A-2015-10565"}}),
				Prohibidos: []ComandoProhibido{{Applet: "graph", Verbo: "show"}},
				Citas:      citaDelArticulo21Leida,
				Hallazgos:  []string{"version-obsoleta"},
			},
		},
		{
			nombre:    "hallazgo-desconocido",
			documento: positivaDelArticulo21 + "hallazgos:\n  - fuente-caducada\n",
			error:     nombreDeEval + ": hallazgos/0, línea 11: value must be 'version-obsoleta'",
		},
		{
			nombre:    "hallazgos-vacio",
			documento: positivaDelArticulo21 + "hallazgos: []\n",
			error:     nombreDeEval + ": hallazgos, línea 10: minItems: got 0, want 1",
		},
		{
			nombre:    "no-activa-con-hallazgos",
			documento: preguntaDelArticulo21 + "activa: false\nhallazgos:\n  - version-obsoleta\n",
			error:     nombreDeEval + ": línea 1: 'not' failed",
		},
		{
			nombre:    "no-se-activan",
			documento: municipioQueNoActiva(noSeActivaBoeLegislacion),
			leida: Eval{
				Fichero:     nombreDeEval,
				Pregunta:    preguntaDelMunicipioLeida,
				Activa:      true,
				NoSeActivan: []string{"boe-legislacion"},
				Comandos:    comandoDelMunicipioLeido,
				Territorio:  territorioDelMunicipioLeido,
			},
		},
		{
			nombre:    "no-activa-con-no-se-activan",
			documento: evalDeLaReceta + noSeActivaBoeLegislacion,
			leida: Eval{
				Fichero:     nombreDeEval,
				Pregunta:    "¿Cómo se hace una tortilla de patatas jugosa?",
				NoSeActivan: []string{"boe-legislacion"},
			},
		},
		{
			nombre:    "no-se-activan-vacio",
			documento: municipioQueNoActiva("no_se_activan: []\n"),
			error:     nombreDeEval + ": no_se_activan, línea 3: minItems: got 0, want 1",
		},
		{
			nombre:    "no-se-activan-con-un-nombre-que-no-es-de-skill",
			documento: municipioQueNoActiva("no_se_activan: [Boe-Legislacion]\n"),
			error: nombreDeEval + ": no_se_activan/0, línea 3: " +
				"'Boe-Legislacion' does not match pattern '^[a-z0-9]+(-[a-z0-9]+)*$'",
		},
		{
			nombre:    "no-se-activan-con-un-nombre-repetido",
			documento: municipioQueNoActiva("no_se_activan: [boe-legislacion, boe-legislacion]\n"),
			error:     nombreDeEval + ": no_se_activan, línea 3: items at 0 and 1 are equal",
		},
		{
			// La eval de contracts/evals-y-juicio.md §3 de H7.4, sin su comentario.
			nombre:    "redacciones-modificadas",
			documento: evalDeDosBloquesDeLaLCSP,
			leida:     leidaDeDosBloquesDeLaLCSP(nombreDeEval),
		},
		{
			nombre:    "no-activa-con-redacciones-modificadas",
			documento: preguntaDelArticulo21 + "activa: false\n" + redaccionesModificadasDeLaLCSP,
			error:     nombreDeEval + ": línea 1: 'not' failed",
		},
		{
			nombre:    "redaccion-modificada-sin-fecha-vigencia-reciente",
			documento: strings.Replace(evalDeDosBloquesDeLaLCSP, "    fecha_vigencia_reciente: \"20200206\"\n", "", 1),
			error:     nombreDeEval + ": redacciones_modificadas/0, línea 34: missing property 'fecha_vigencia_reciente'",
		},
		{
			nombre:    "redaccion-modificada-con-una-fecha-de-siete-cifras",
			documento: strings.Replace(evalDeDosBloquesDeLaLCSP, "\"20180309\"", "\"2018039\"", 1),
			error: nombreDeEval + ": redacciones_modificadas/0/fecha_vigencia, línea 36: " +
				"'2018039' does not match pattern " + patronDeFecha,
		},
		{
			nombre:    "redaccion-modificada-con-el-mes-13",
			documento: strings.Replace(evalDeDosBloquesDeLaLCSP, "\"20200206\"", "\"20201306\"", 1),
			error: nombreDeEval + ": redacciones_modificadas/0/fecha_vigencia_reciente, línea 37: " +
				"'20201306' does not match pattern " + patronDeFecha,
		},
		{
			// La eval de boe-legislacion de contracts/evals-en-dos-modos.md §1 de
			// H21, sin su comentario.
			nombre:    "sin-binario-ni-servidor",
			documento: evalSinBinarioDeBoeLegislacion,
			leida: Eval{
				Fichero:              nombreDeEval,
				Pregunta:             preguntaSinBinarioDeBoeLegislacion,
				Activa:               true,
				SinBinarioNiServidor: true,
			},
		},
		{
			// La de legal-core del mismo contrato, sin su comentario.
			nombre:    "sin-binario-ni-servidor-con-no-se-activan",
			documento: evalSinBinarioDeLegalCore,
			leida: Eval{
				Fichero:              nombreDeEval,
				Pregunta:             preguntaSinBinarioDeLegalCore,
				Activa:               true,
				NoSeActivan:          []string{"boe-legislacion"},
				SinBinarioNiServidor: true,
			},
		},
		{
			nombre:    "sin-binario-ni-servidor-con-comandos",
			documento: evalSinBinarioDeBoeLegislacion + comandoDelArticulo21,
			error:     nombreDeEval + ": línea 1: 'not' failed",
		},
		{
			nombre:    "sin-binario-ni-servidor-con-citas",
			documento: evalSinBinarioDeBoeLegislacion + citaDelArticulo21,
			error:     nombreDeEval + ": línea 1: 'not' failed",
		},
		{
			nombre:    "sin-binario-ni-servidor-con-informativa",
			documento: evalSinBinarioDeBoeLegislacion + "informativa: true\n",
			error:     nombreDeEval + ": línea 1: 'not' failed",
		},
		{
			nombre:    "no-activa-sin-binario-ni-servidor",
			documento: evalDeLaReceta + sinBinarioNiServidor,
			error:     nombreDeEval + ": línea 1: 'not' failed",
		},
		{
			// Con false, la eval activa no llevaría nada que comprobar: la clave
			// solo se escribe para declararlo.
			nombre:    "sin-binario-ni-servidor-falso",
			documento: preguntaDelArticulo21 + "activa: true\nsin_binario_ni_servidor: false\n",
			error:     nombreDeEval + ": sin_binario_ni_servidor, línea 3: value must be true",
		},
		{
			// Desde la serie de cierres (ADR 0038, punto 6), la clase de pregunta:
			// una pregunta por materia lo declara, con informativa o sin ella.
			nombre:    "por-materia",
			documento: preguntaDelArticulo21 + "activa: true\npor_materia: true\n" + comandoDelArticulo21 + citaDelArticulo21,
			leida: Eval{
				Fichero:    nombreDeEval,
				Pregunta:   "¿qué dice el art. 21 de la Ley 39/2015?",
				Activa:     true,
				PorMateria: true,
				Comandos:   comandoDelArticulo21Leido,
				Citas:      citaDelArticulo21Leida,
			},
		},
		{
			// Con false es el caso normal, el de la eval que no lo escribe.
			nombre:    "por-materia-falso",
			documento: preguntaDelArticulo21 + "activa: true\npor_materia: false\n" + comandoDelArticulo21 + citaDelArticulo21,
			leida: Eval{
				Fichero:  nombreDeEval,
				Pregunta: "¿qué dice el art. 21 de la Ley 39/2015?",
				Activa:   true,
				Comandos: comandoDelArticulo21Leido,
				Citas:    citaDelArticulo21Leida,
			},
		},
		{
			nombre:     "por-materia-no-booleano",
			documento:  positivaDelArticulo21 + "por_materia: si\n",
			fragmentos: []string{"por_materia"},
		},
		{
			// Una eval de no activación no tiene respuestas que repartir.
			nombre:    "no-activa-por-materia",
			documento: evalDeLaReceta + "por_materia: true\n",
			error:     nombreDeEval + ": línea 1: 'not' failed",
		},
		{
			// Las respuestas de una eval sin binario ni servidor no son de ningún
			// modo y no cuentan en la serie.
			nombre:    "sin-binario-ni-servidor-con-por-materia",
			documento: evalSinBinarioDeBoeLegislacion + "por_materia: true\n",
			error:     nombreDeEval + ": línea 1: 'not' failed",
		},
	}

	for _, caso := range casos {
		t.Run(caso.nombre, func(t *testing.T) {
			t.Parallel()

			leida, err := LeerEval(nombreDeEval, []byte(caso.documento))
			if caso.error == "" && len(caso.fragmentos) == 0 {
				require.NoError(t, err)
				assert.Equal(t, caso.leida, leida)

				return
			}

			require.Error(t, err)
			assert.Zero(t, leida, "una eval mal formada no se entrega a medias")
			assert.True(t, strings.HasPrefix(err.Error(), nombreDeEval+": "),
				"el error empieza por el nombre del fichero: %q", err.Error())

			if caso.error != "" {
				require.EqualError(t, err, caso.error)
			}

			for _, fragmento := range caso.fragmentos {
				assert.ErrorContains(t, err, fragmento)
			}
		})
	}
}

// TestFormaDelComando fija que formaDelComando decide la variante de un comando
// esperado por su verbo (data-model §6.1 de H6; research D21): cada una de las
// cinco formas del esquema, leída con LeerEval de una eval que solo lleva ese
// comando, tiene la suya —la de consulta de norma, con cualquiera de los tres
// verbos de su enumerado— y ni el comando de territorio ni, desde H7, el de
// comprobación, con el verbo check, son una consulta de norma (contrato
// evals-y-skill §1 de H7); tampoco, desde H7.1, la comprobación que lleva la
// norma que consulta (contrato evals-y-skill §1 de H7.1), ni, desde H23, los dos
// comandos de cita, que tienen cada uno la suya (contracts/evals-jurisprudencia.md
// §1 de H23), ni, desde H9, el comando calcular, con su tipo
// (contracts/evals-legal-core.md §1 de H9).
func TestFormaDelComando(t *testing.T) {
	t.Parallel()

	casos := []struct {
		nombre string

		// comando son las líneas del único comando de la eval.
		comando string

		forma formaDeComando
	}{
		{
			nombre:  "bloque",
			comando: "  - applet: boe\n    norma: BOE-A-2015-10565\n    bloque: a21\n",
			forma:   formaBloque,
		},
		{
			nombre:  "consulta-del-indice",
			comando: "  - applet: boe\n    verbo: indice\n    norma: BOE-A-2015-10565\n",
			forma:   formaConsultaDeNorma,
		},
		{
			nombre:  "consulta-de-los-metadatos",
			comando: "  - applet: boe\n    verbo: metadatos\n    norma: BOE-A-2015-10565\n",
			forma:   formaConsultaDeNorma,
		},
		{
			nombre:  "consulta-del-analisis",
			comando: "  - applet: boe\n    verbo: analisis\n    norma: BOE-A-2015-10565\n",
			forma:   formaConsultaDeNorma,
		},
		{
			nombre:  "busqueda",
			comando: "  - applet: boe\n    verbo: buscar\n    terminos: [procedimiento, común]\n",
			forma:   formaBusqueda,
		},
		{
			nombre:  "territorio",
			comando: "  - applet: territorio\n    verbo: resolver\n    municipio: Leganés\n",
			forma:   formaTerritorio,
		},
		{
			nombre:  "comprobacion",
			comando: comprobacionDelGrafo,
			forma:   formaComprobacion,
		},
		{
			nombre:  "comprobacion-con-norma",
			comando: comprobacionDelGrafo + normaDeLaComprobacion,
			forma:   formaComprobacion,
		},
		{
			nombre:  "preparar",
			comando: "  - applet: cita\n    verbo: preparar\n    con_texto: true\n",
			forma:   formaPreparar,
		},
		{
			nombre:  "cotejar",
			comando: "  - applet: cita\n    verbo: cotejar\n    roj: STS 1088/2023\n",
			forma:   formaCotejar,
		},
		{
			nombre:  "calcular",
			comando: strings.TrimPrefix(comandoDelCalculo, "comandos:\n"),
			forma:   formaCalcular,
		},
	}

	for _, caso := range casos {
		t.Run(caso.nombre, func(t *testing.T) {
			t.Parallel()

			eval, err := LeerEval(nombreDeEval,
				[]byte(preguntaDelArticulo21+"activa: true\ncomandos:\n"+caso.comando+citaDelArticulo21))
			require.NoError(t, err)
			require.Len(t, eval.Comandos, 1)

			assert.Equal(t, caso.forma, formaDelComando(eval.Comandos[0]))
		})
	}
}

// TestFormatoDeSentencias fija lo que H23 da al formato de eval
// (contracts/evals-jurisprudencia.md §1; FR-052): las siete claves de
// sentencias y los dos comandos de cita, con su roj y con con_texto, se leen
// enteros, y una eval con sentencias puede no llevar comandos; una eval de las
// otras dos skills se lee igual, sin sentencias y sin ningún comando de cita, y
// la positiva sin nada que esperar se sigue rechazando con el error de hoy; y
// sentencias en una eval que no activa la skill o que es sin binario ni
// servidor, vacías, con citas y ninguna_cita a la vez, con un valor falso, con
// un ECLI que no es español, con un ROJ, una dirección, una cita o una casilla
// que no tienen su forma o con otra clave, y un comando de cita con lo que su
// forma no admite, se rechazan: cada documento es el bien formado con solo ese
// defecto.
func TestFormatoDeSentencias(t *testing.T) {
	t.Parallel()

	const (
		preguntaLeida = "\xc2\xbfexiste la STS 1088/2023, de 4 de julio?"
		pregunta      = "pregunta: \"" + preguntaLeida + "\"\n"
		activa        = pregunta + "activa: true\n"
		ningunaCita   = "sentencias:\n  ninguna_cita: true\n"
		preparar      = "comandos:\n  - applet: cita\n    verbo: preparar\n"
		cotejar       = "comandos:\n  - applet: cita\n    verbo: cotejar\n"

		ecliDelEsquema = "'^ECLI:ES:[A-Z][A-Z0-9]{0,6}:[0-9]{4}:[A-Z0-9.]{1,25}$'"
		rojDelEsquema  = "'^[A-Z]+( [A-Z]+)* [0-9]+/[0-9]{4}$'"
	)

	// sentencias es el documento que activa la skill, sin comandos, con esas
	// líneas bajo sentencias, que es su línea 3.
	sentencias := func(lineas string) string { return activa + "sentencias:\n" + lineas }

	casos := []struct {
		nombre     string
		documento  string
		leida      Eval
		error      string
		fragmentos []string
	}{
		{
			nombre: "todas-las-claves",
			documento: "# Eval de jurisprudencia: lo que espera de la sesi\xc3\xb3n y de su respuesta.\n" + activa +
				"comandos:\n" +
				"  - applet: cita\n    verbo: cotejar\n    roj: STS 1088/2023\n" +
				"  - applet: cita\n    verbo: preparar\n    roj: STS 1088/2023\n" +
				"  - applet: cita\n    verbo: preparar\n    con_texto: true\n" +
				"  - applet: cita\n    verbo: cotejar\n" +
				"sentencias:\n" +
				"  citas:\n    - ecli: ECLI:ES:TS:2023:3144\n      roj: STS 3144/2023\n" +
				"    - ecli: ECLI:ES:APM:2020:1A.2\n      roj: SAP M 15/2020\n" +
				"  sin_cita_del_roj:\n    - STS 1088/2023\n" +
				"  no_comprobada: true\n" +
				"  direcciones:\n    - " + direccionDelCendoj + "\n    - " + direccionDelTC + "\n" +
				"  casillas:\n    - nombre: " + casillaDeROJ + "\n      valor: STS 1088/2023\n" +
				"    - nombre: " + casillaDeFecha + "\n      valor: 04/07/2023\n" +
				"  direccion_de_busqueda: true\n",
			leida: Eval{
				Fichero:  nombreDeEval,
				Pregunta: preguntaLeida,
				Activa:   true,
				Comandos: []ComandoEsperado{
					{Applet: "cita", Verbo: "cotejar", ROJ: rojQueNoEsElSuyo},
					{Applet: "cita", Verbo: "preparar", ROJ: rojQueNoEsElSuyo},
					{Applet: "cita", Verbo: "preparar", ConTexto: true},
					{Applet: "cita", Verbo: "cotejar"},
				},
				Sentencias: SentenciasEsperadas{
					Citas: []CitaDeSentenciaEsperada{
						{ECLI: ecliDelFragmento, ROJ: rojDelFragmento}, {ECLI: "ECLI:ES:APM:2020:1A.2", ROJ: "SAP M 15/2020"},
					},
					SinCitaDelROJ: []string{rojQueNoEsElSuyo},
					NoComprobada:  true,
					Direcciones:   []string{direccionDelCendoj, direccionDelTC},
					Casillas: []CasillaEsperada{
						{Nombre: casillaDeROJ, Valor: rojQueNoEsElSuyo}, {Nombre: casillaDeFecha, Valor: "04/07/2023"},
					},
					DireccionDeBusqueda: true,
				},
			},
		},
		{
			// Con sentencias, comandos deja de ser obligatorio.
			nombre:    "ninguna-cita-sin-comandos",
			documento: activa + ningunaCita,
			leida: Eval{
				Fichero: nombreDeEval, Pregunta: preguntaLeida, Activa: true,
				Sentencias: SentenciasEsperadas{NingunaCita: true},
			},
		},
		{
			// Y las claves de hoy siguen valiendo a su lado.
			nombre:    "con-una-cita-del-boe",
			documento: activa + preparar + citaDelArticulo21 + ningunaCita,
			leida: Eval{
				Fichero: nombreDeEval, Pregunta: preguntaLeida, Activa: true,
				Comandos:   []ComandoEsperado{{Applet: "cita", Verbo: "preparar"}},
				Citas:      []CitaEsperada{{Norma: "BOE-A-2015-10565", Bloque: "a21"}},
				Sentencias: SentenciasEsperadas{NingunaCita: true},
			},
		},
		{
			// Sin sentencias, la que activa la skill sigue necesitando sus comandos
			// y algo que esperar, con el error de hoy.
			nombre:    "sin-sentencias-ni-comandos",
			documento: activa,
			error: nombreDeEval + ": l\xc3\xadnea 1: missing property 'citas'\n" +
				"l\xc3\xadnea 1: missing property 'comandos'\n" +
				"l\xc3\xadnea 1: missing property 'plazo'\n" +
				"l\xc3\xadnea 1: missing property 'plazo_sin_calcular'\n" +
				"l\xc3\xadnea 1: missing property 'territorio'",
		},
		{
			nombre:    "sin-sentencias-y-con-un-comando-de-cita",
			documento: activa + preparar,
			error: nombreDeEval + ": l\xc3\xadnea 1: missing property 'citas'\n" +
				"l\xc3\xadnea 1: missing property 'plazo'\n" +
				"l\xc3\xadnea 1: missing property 'plazo_sin_calcular'\n" +
				"l\xc3\xadnea 1: missing property 'territorio'",
		},
		{
			nombre:    "en-una-eval-que-no-activa-la-skill",
			documento: pregunta + "activa: false\n" + ningunaCita,
			error:     nombreDeEval + ": l\xc3\xadnea 1: 'not' failed",
		},
		{
			nombre:    "en-una-eval-sin-binario-ni-servidor",
			documento: activa + sinBinarioNiServidor + ningunaCita,
			error:     nombreDeEval + ": l\xc3\xadnea 1: 'not' failed",
		},
		{
			nombre:    "vacias",
			documento: activa + "sentencias: {}\n",
			error:     nombreDeEval + ": sentencias, l\xc3\xadnea 3: minProperties: got 0, want 1",
		},
		{
			nombre: "con-citas-y-ninguna-cita",
			documento: sentencias("  citas:\n    - ecli: ECLI:ES:TS:2023:3144\n      roj: STS 3144/2023\n" +
				"  ninguna_cita: true\n"),
			error: nombreDeEval + ": sentencias, l\xc3\xadnea 4: 'not' failed",
		},
		{
			nombre:    "ninguna-cita-falsa",
			documento: sentencias("  ninguna_cita: false\n"),
			error:     nombreDeEval + ": sentencias/ninguna_cita, l\xc3\xadnea 4: value must be true",
		},
		{
			nombre:    "no-comprobada-falsa",
			documento: sentencias("  no_comprobada: false\n"),
			error:     nombreDeEval + ": sentencias/no_comprobada, l\xc3\xadnea 4: value must be true",
		},
		{
			nombre:    "direccion-de-busqueda-falsa",
			documento: sentencias("  direccion_de_busqueda: false\n"),
			error:     nombreDeEval + ": sentencias/direccion_de_busqueda, l\xc3\xadnea 4: value must be true",
		},
		{
			nombre:    "citas-vacias",
			documento: sentencias("  citas: []\n"),
			error:     nombreDeEval + ": sentencias/citas, l\xc3\xadnea 4: minItems: got 0, want 1",
		},
		{
			nombre:    "cita-con-un-ecli-de-otro-pais",
			documento: sentencias("  citas:\n    - ecli: ECLI:EU:C:2019:123\n      roj: STJUE 123/2019\n"),
			error: nombreDeEval + ": sentencias/citas/0/ecli, l\xc3\xadnea 5: " +
				"'ECLI:EU:C:2019:123' does not match pattern " + ecliDelEsquema,
		},
		{
			nombre:    "cita-sin-roj",
			documento: sentencias("  citas:\n    - ecli: ECLI:ES:TS:2023:3144\n"),
			error:     nombreDeEval + ": sentencias/citas/0, l\xc3\xadnea 5: missing property 'roj'",
		},
		{
			nombre:    "cita-con-otra-clave",
			documento: sentencias("  citas:\n    - ecli: ECLI:ES:TS:2023:3144\n      roj: STS 3144/2023\n      fecha: 2023-07-04\n"),
			error:     nombreDeEval + ": sentencias/citas/0, l\xc3\xadnea 5: additional properties 'fecha' not allowed",
		},
		{
			nombre:    "roj-que-no-se-cita-en-minusculas",
			documento: sentencias("  sin_cita_del_roj:\n    - sts 1088/2023\n"),
			error: nombreDeEval + ": sentencias/sin_cita_del_roj/0, l\xc3\xadnea 5: " +
				"'sts 1088/2023' does not match pattern " + rojDelEsquema,
		},
		{
			nombre:     "direccion-que-no-lo-es",
			documento:  sentencias("  direcciones:\n    - \"el buscador del CENDOJ\"\n"),
			fragmentos: []string{"sentencias/direcciones/0, l\xc3\xadnea 5: 'el buscador del CENDOJ' is not valid uri"},
		},
		{
			nombre:    "casilla-sin-valor",
			documento: sentencias("  casillas:\n    - nombre: " + casillaDeROJ + "\n"),
			error:     nombreDeEval + ": sentencias/casillas/0, l\xc3\xadnea 5: missing property 'valor'",
		},
		{
			nombre:    "casilla-con-el-valor-vacio",
			documento: sentencias("  casillas:\n    - nombre: " + casillaDeROJ + "\n      valor: \"\"\n"),
			error:     nombreDeEval + ": sentencias/casillas/0/valor, l\xc3\xadnea 6: minLength: got 0, want 1",
		},
		{
			nombre:    "con-otra-clave",
			documento: sentencias("  contiene:\n    - STS 1088/2023\n"),
			error:     nombreDeEval + ": sentencias, l\xc3\xadnea 4: additional properties 'contiene' not allowed",
		},
		{
			nombre:     "preparar-con-una-norma",
			documento:  activa + preparar + "    norma: BOE-A-2015-10565\n" + ningunaCita,
			fragmentos: []string{"comandos/0, l\xc3\xadnea 4: additional properties 'norma' not allowed"},
		},
		{
			nombre:     "preparar-con-texto-falso",
			documento:  activa + preparar + "    con_texto: false\n" + ningunaCita,
			fragmentos: []string{"comandos/0/con_texto, l\xc3\xadnea 6: value must be true"},
		},
		{
			nombre:     "cotejar-con-texto",
			documento:  activa + cotejar + "    con_texto: true\n" + ningunaCita,
			fragmentos: []string{"comandos/0, l\xc3\xadnea 4: additional properties 'con_texto' not allowed"},
		},
		{
			nombre:     "cotejar-con-un-roj-sin-su-forma",
			documento:  activa + cotejar + "    roj: STS-1088\n" + ningunaCita,
			fragmentos: []string{"comandos/0/roj, l\xc3\xadnea 6: 'STS-1088' does not match pattern " + rojDelEsquema},
		},
	}

	for _, caso := range casos {
		t.Run(caso.nombre, func(t *testing.T) {
			t.Parallel()

			leida, err := LeerEval(nombreDeEval, []byte(caso.documento))
			if caso.error == "" && len(caso.fragmentos) == 0 {
				require.NoError(t, err)
				assert.Equal(t, caso.leida, leida)
				assert.True(t, leida.Sentencias.declaradas(), "la eval declara sentencias")

				return
			}

			require.Error(t, err)
			assert.Zero(t, leida, "una eval mal formada no se entrega a medias")

			if caso.error != "" {
				require.EqualError(t, err, caso.error)
			}

			for _, fragmento := range caso.fragmentos {
				assert.ErrorContains(t, err, fragmento)
			}
		})
	}

	t.Run("las-evals-de-hoy", func(t *testing.T) {
		t.Parallel()

		for _, dir := range []string{evalsDelRepositorio, evalsDeLegalCore} {
			conjunto, err := LeerConjunto(dir)
			require.NoError(t, err)
			require.Empty(t, conjunto.MalFormados, "cada fichero de %s es una eval bien formada", dir)
			require.NotEmpty(t, conjunto.Evals, "%s tiene evals", dir)

			for _, eval := range conjunto.Evals {
				assert.False(t, eval.Sentencias.declaradas(), "%s de %s no declara sentencias", eval.Fichero, dir)
				assert.Zero(t, eval.Sentencias, "%s de %s se lee sin sentencias", eval.Fichero, dir)

				for posicion, comando := range eval.Comandos {
					assert.Empty(t, comando.ROJ, "el comando %d de %s de %s se lee sin ROJ", posicion, eval.Fichero, dir)
					assert.False(t, comando.ConTexto, "el comando %d de %s de %s se lee sin texto", posicion, eval.Fichero, dir)
					assert.NotContains(t, []formaDeComando{formaPreparar, formaCotejar}, formaDelComando(comando),
						"el comando %d de %s de %s conserva su forma", posicion, eval.Fichero, dir)
				}
			}
		}
	})
}

// Trozos de las evals sintéticas de TestFormatoDePlazos y TestFormaDelComando,
// cada uno con sus líneas completas: la pregunta de un plazo, su comando
// calcular, su plazo esperado sin avisos, la declaración de no cubierto con la
// Ley General Tributaria y plazos calcular prohibido (contracts/evals-legal-core.md
// §1 y §5 de H9).
const (
	preguntaDelPlazo  = "pregunta: \"\xc2\xbfHasta cu\xc3\xa1ndo puedo presentar el recurso de reposici\xc3\xb3n?\"\n"
	comandoDelCalculo = "comandos:\n" +
		"  - applet: plazos\n" +
		"    verbo: calcular\n" +
		"    tipo: reposicion\n"
	plazoSinAvisos = "plazo:\n" +
		"  fecha: \"2026-10-13\"\n" +
		"  avisos: []\n" +
		"  razonamiento: Un mes desde el d\xc3\xada siguiente, prorrogado al primer d\xc3\xada h\xc3\xa1bil.\n"
	noCubiertoPorLaLGT = "no_cubierto:\n" +
		"  norma: BOE-A-2003-23186\n"
	prohibidoElCalculo = "prohibidos:\n" +
		"  - applet: plazos\n" +
		"    verbo: calcular\n"
	sinCalcularPorLaLPAC = "plazo_sin_calcular:\n" +
		"  norma: BOE-A-2015-10565\n"
)

// TestFormatoDePlazos fija lo que H9 da al formato de eval
// (contracts/evals-legal-core.md §1; data-model §3 de H9; FR-051): plazo, con su
// fecha, sus avisos y su razonamiento, no_cubierto, con su norma, y el comando
// calcular, con su tipo, se leen enteros; con plazo, una eval puede no llevar
// citas, y con no_cubierto, ni comandos ni citas. Se rechazan: plazo o
// no_cubierto en una eval que no activa la skill o que es sin binario ni
// servidor, los dos juntos, plazo sin comandos, una fecha que no es AAAA-MM-DD,
// unos avisos repetidos o que no son de plazos, plazo sin avisos o sin
// razonamiento, o con un razonamiento vacío o con otra clave, una norma que no es
// BOE-A-… o que falta, y un comando calcular sin tipo, con el tipo vacío o con
// una norma: cada documento es el bien formado con solo ese defecto. Y el
// enumerado de plazo.avisos del esquema publicado es exactamente
// plazos.CodigosDeAviso, comprobado como ComprobarCodigosDeAviso comprueba el de
// avisos (ComprobarCodigosDeAvisoDelPlazo); las evals de boe-legislacion y de
// jurisprudencia se leen sin plazo, sin no_cubierto y sin ningún tipo.
func TestFormatoDePlazos(t *testing.T) {
	t.Parallel()

	const (
		preguntaLeida = "\xc2\xbfHasta cu\xc3\xa1ndo puedo presentar el recurso de reposici\xc3\xb3n?"
		activa        = preguntaDelPlazo + "activa: true\n"
		razonamiento  = "Un mes desde el d\xc3\xada siguiente, prorrogado al primer d\xc3\xada h\xc3\xa1bil."
		lgt           = "BOE-A-2003-23186"

		fechaDelEsquema = "'^[0-9]{4}-(0[1-9]|1[0-2])-(0[1-9]|[12][0-9]|3[01])$'"
		normaDelEsquema = "'^BOE-A-[0-9]{4}-[0-9]{1,9}$'"
	)

	// conPlazo es el documento que activa la skill con el comando calcular y esas
	// líneas bajo plazo, que es su línea 7.
	conPlazo := func(lineas string) string { return activa + comandoDelCalculo + "plazo:\n" + lineas }

	calculo := ComandoEsperado{Applet: "plazos", Verbo: "calcular", Tipo: "reposicion"}

	probarLasLecturasDeEval(t, []casoDeLecturaDeEval{
		{
			nombre: "plazo-con-sus-claves",
			documento: "# Eval de plazo: lo que espera de la sesi\xc3\xb3n y de su respuesta.\n" + activa +
				comandoDelCalculo + "  - applet: boe\n    norma: BOE-A-2015-10565\n    bloque: a124\n" +
				"citas:\n  - norma: BOE-A-2015-10565\n    bloque: a124\n" +
				"plazo:\n  fecha: \"2026-10-09\"\n  avisos: [faltan-festivos-locales, dias-del-art-183-lopj]\n" +
				"  razonamiento: >-\n    Un mes desde el d\xc3\xada siguiente,\n    prorrogado al primer d\xc3\xada h\xc3\xa1bil.\n",
			leida: Eval{
				Fichero: nombreDeEval, Pregunta: preguntaLeida, Activa: true,
				Comandos: []ComandoEsperado{calculo, {Applet: "boe", Norma: "BOE-A-2015-10565", Bloque: "a124"}},
				Citas:    []CitaEsperada{{Norma: "BOE-A-2015-10565", Bloque: "a124"}},
				Plazo: PlazoEsperado{
					Fecha:        "2026-10-09",
					Avisos:       []string{"faltan-festivos-locales", "dias-del-art-183-lopj"},
					Razonamiento: razonamiento,
				},
			},
		},
		{
			// Con plazo, la eval puede no llevar citas: el plazo es lo que espera.
			nombre:    "plazo-sin-citas-ni-avisos",
			documento: activa + comandoDelCalculo + plazoSinAvisos,
			leida: Eval{
				Fichero: nombreDeEval, Pregunta: preguntaLeida, Activa: true,
				Comandos: []ComandoEsperado{calculo},
				Plazo:    PlazoEsperado{Fecha: "2026-10-13", Avisos: []string{}, Razonamiento: razonamiento},
			},
		},
		{
			// Con no_cubierto, la eval puede no llevar comandos ni citas.
			nombre:    "no-cubierto-sin-comandos",
			documento: activa + prohibidoElCalculo + noCubiertoPorLaLGT,
			leida: Eval{
				Fichero: nombreDeEval, Pregunta: preguntaLeida, Activa: true,
				Prohibidos: []ComandoProhibido{{Applet: "plazos", Verbo: "calcular"}},
				NoCubierto: NoCubiertoEsperado{Norma: lgt},
			},
		},
		{
			// Con plazo_sin_calcular, la eval lleva comandos y puede no llevar
			// citas, y su comando calcular, sin_calcular (ADR 0042).
			nombre:    "plazo-sin-calcular",
			documento: activa + comandoDelCalculo + "    sin_calcular: true\n" + sinCalcularPorLaLPAC,
			leida: Eval{
				Fichero: nombreDeEval, Pregunta: preguntaLeida, Activa: true,
				Comandos:         []ComandoEsperado{{Applet: "plazos", Verbo: "calcular", Tipo: "reposicion", SinCalcular: true}},
				PlazoSinCalcular: PlazoSinCalcularEsperado{Norma: "BOE-A-2015-10565"},
			},
		},
		{
			nombre:    "plazo-sin-calcular-sin-comandos",
			documento: activa + sinCalcularPorLaLPAC,
			error:     nombreDeEval + ": l\xc3\xadnea 1: missing property 'comandos'",
		},
		{
			nombre:    "plazo-y-plazo-sin-calcular",
			documento: activa + comandoDelCalculo + plazoSinAvisos + sinCalcularPorLaLPAC,
			error:     nombreDeEval + ": l\xc3\xadnea 1: 'not' failed",
		},
		{
			nombre:    "no-cubierto-y-plazo-sin-calcular",
			documento: activa + comandoDelCalculo + noCubiertoPorLaLGT + sinCalcularPorLaLPAC,
			error:     nombreDeEval + ": l\xc3\xadnea 1: 'not' failed",
		},
		{
			nombre:    "plazo-sin-calcular-en-una-eval-que-no-activa-la-skill",
			documento: preguntaDelPlazo + "activa: false\n" + sinCalcularPorLaLPAC,
			error:     nombreDeEval + ": l\xc3\xadnea 1: 'not' failed",
		},
		{
			nombre:    "plazo-sin-calcular-en-una-eval-sin-binario-ni-servidor",
			documento: activa + sinBinarioNiServidor + sinCalcularPorLaLPAC,
			error:     nombreDeEval + ": l\xc3\xadnea 1: 'not' failed",
		},
		{
			nombre:    "plazo-sin-calcular-sin-su-forma",
			documento: activa + comandoDelCalculo + "plazo_sin_calcular:\n  norma: LPAC\n",
			error: nombreDeEval + ": plazo_sin_calcular/norma, l\xc3\xadnea 8: 'LPAC' does not match pattern " +
				normaDelEsquema,
		},
		{
			nombre:     "calcular-con-sin-calcular-falso",
			documento:  activa + comandoDelCalculo + "    sin_calcular: false\n" + plazoSinAvisos,
			fragmentos: []string{"comandos/0/sin_calcular, l\xc3\xadnea 7: "},
		},
		{
			nombre:    "plazo-sin-comandos",
			documento: activa + plazoSinAvisos,
			error:     nombreDeEval + ": l\xc3\xadnea 1: missing property 'comandos'",
		},
		{
			nombre:    "plazo-y-no-cubierto",
			documento: activa + comandoDelCalculo + plazoSinAvisos + noCubiertoPorLaLGT,
			error:     nombreDeEval + ": l\xc3\xadnea 1: 'not' failed",
		},
		{
			nombre:    "plazo-en-una-eval-que-no-activa-la-skill",
			documento: preguntaDelPlazo + "activa: false\n" + plazoSinAvisos,
			error:     nombreDeEval + ": l\xc3\xadnea 1: 'not' failed",
		},
		{
			nombre:    "no-cubierto-en-una-eval-que-no-activa-la-skill",
			documento: preguntaDelPlazo + "activa: false\n" + noCubiertoPorLaLGT,
			error:     nombreDeEval + ": l\xc3\xadnea 1: 'not' failed",
		},
		{
			nombre:    "plazo-en-una-eval-sin-binario-ni-servidor",
			documento: activa + sinBinarioNiServidor + plazoSinAvisos,
			error:     nombreDeEval + ": l\xc3\xadnea 1: 'not' failed",
		},
		{
			nombre:    "no-cubierto-en-una-eval-sin-binario-ni-servidor",
			documento: activa + sinBinarioNiServidor + noCubiertoPorLaLGT,
			error:     nombreDeEval + ": l\xc3\xadnea 1: 'not' failed",
		},
		{
			nombre:    "fecha-sin-guiones",
			documento: conPlazo("  fecha: \"20261013\"\n  avisos: []\n  razonamiento: " + razonamiento + "\n"),
			error: nombreDeEval + ": plazo/fecha, l\xc3\xadnea 8: '20261013' does not match pattern " +
				fechaDelEsquema,
		},
		{
			nombre:    "fecha-con-el-mes-13",
			documento: conPlazo("  fecha: \"2026-13-09\"\n  avisos: []\n  razonamiento: " + razonamiento + "\n"),
			error: nombreDeEval + ": plazo/fecha, l\xc3\xadnea 8: '2026-13-09' does not match pattern " +
				fechaDelEsquema,
		},
		{
			nombre: "avisos-repetidos",
			documento: conPlazo("  fecha: \"2026-10-09\"\n  avisos: [faltan-festivos-locales, faltan-festivos-locales]\n" +
				"  razonamiento: " + razonamiento + "\n"),
			fragmentos: []string{"plazo/avisos, l\xc3\xadnea 9: "},
		},
		{
			// derogada es un código de aviso de vigencia del BOE, de la clave
			// avisos, y no de plazos.
			nombre:     "aviso-de-vigencia",
			documento:  conPlazo("  fecha: \"2026-10-09\"\n  avisos: [derogada]\n  razonamiento: " + razonamiento + "\n"),
			fragmentos: []string{"plazo/avisos/0, l\xc3\xadnea 9: "},
		},
		{
			nombre:    "sin-avisos",
			documento: conPlazo("  fecha: \"2026-10-13\"\n  razonamiento: " + razonamiento + "\n"),
			error:     nombreDeEval + ": plazo, l\xc3\xadnea 8: missing property 'avisos'",
		},
		{
			nombre:    "sin-razonamiento",
			documento: conPlazo("  fecha: \"2026-10-13\"\n  avisos: []\n"),
			error:     nombreDeEval + ": plazo, l\xc3\xadnea 8: missing property 'razonamiento'",
		},
		{
			nombre:    "razonamiento-vacio",
			documento: conPlazo("  fecha: \"2026-10-13\"\n  avisos: []\n  razonamiento: \"\"\n"),
			error:     nombreDeEval + ": plazo/razonamiento, l\xc3\xadnea 10: minLength: got 0, want 1",
		},
		{
			nombre: "plazo-con-otra-clave",
			documento: conPlazo("  fecha: \"2026-10-13\"\n  avisos: []\n  razonamiento: " + razonamiento + "\n" +
				"  tipo: reposicion\n"),
			error: nombreDeEval + ": plazo, l\xc3\xadnea 8: additional properties 'tipo' not allowed",
		},
		{
			nombre:    "no-cubierto-sin-su-forma",
			documento: activa + "no_cubierto:\n  norma: LGT\n",
			error:     nombreDeEval + ": no_cubierto/norma, l\xc3\xadnea 4: 'LGT' does not match pattern " + normaDelEsquema,
		},
		{
			nombre:    "no-cubierto-vacio",
			documento: activa + "no_cubierto: {}\n",
			error:     nombreDeEval + ": no_cubierto, l\xc3\xadnea 3: missing property 'norma'",
		},
		{
			nombre:     "calcular-sin-tipo",
			documento:  activa + "comandos:\n  - applet: plazos\n    verbo: calcular\n" + plazoSinAvisos,
			fragmentos: []string{"comandos/0, l\xc3\xadnea 4: missing property 'tipo'"},
		},
		{
			nombre:     "calcular-con-el-tipo-vacio",
			documento:  activa + "comandos:\n  - applet: plazos\n    verbo: calcular\n    tipo: \"\"\n" + plazoSinAvisos,
			fragmentos: []string{"comandos/0/tipo, l\xc3\xadnea 6: minLength: got 0, want 1"},
		},
		{
			nombre:     "calcular-con-una-norma",
			documento:  activa + comandoDelCalculo + "    norma: BOE-A-2015-10565\n" + plazoSinAvisos,
			fragmentos: []string{"comandos/0, l\xc3\xadnea 4: additional properties 'norma' not allowed"},
		},
	})

	t.Run("enumerado-de-los-avisos", probarElEnumeradoDeLosAvisosDelPlazo)

	t.Run("las-evals-de-las-otras-skills", func(t *testing.T) {
		t.Parallel()

		for _, dir := range []string{evalsDelRepositorio, evalsDeJurisprudencia} {
			conjunto, err := LeerConjunto(dir)
			require.NoError(t, err)
			require.Empty(t, conjunto.MalFormados, "cada fichero de %s es una eval bien formada", dir)
			require.NotEmpty(t, conjunto.Evals, "%s tiene evals", dir)

			for _, eval := range conjunto.Evals {
				assert.Zero(t, eval.Plazo, "%s de %s se lee sin plazo", eval.Fichero, dir)
				assert.Zero(t, eval.NoCubierto, "%s de %s se lee sin no_cubierto", eval.Fichero, dir)
				assert.Zero(t, eval.PlazoSinCalcular, "%s de %s se lee sin plazo_sin_calcular", eval.Fichero, dir)

				for posicion, comando := range eval.Comandos {
					assert.Empty(t, comando.Tipo, "el comando %d de %s de %s se lee sin tipo", posicion, eval.Fichero, dir)
					assert.NotEqual(t, formaCalcular, formaDelComando(comando),
						"el comando %d de %s de %s conserva su forma", posicion, eval.Fichero, dir)
				}
			}
		}
	})
}

// casoDeLecturaDeEval es un documento de eval y lo que LeerEval tiene que dar de
// él: sin error ni fragmentos, la eval leída; con ellos, la eval vacía y un error
// que es el error dado o que contiene cada fragmento.
type casoDeLecturaDeEval struct {
	nombre     string
	documento  string
	leida      Eval
	error      string
	fragmentos []string
}

// probarLasLecturasDeEval lee con LeerEval el documento de cada caso, en su
// subtest, y lo compara con lo que el caso exige.
func probarLasLecturasDeEval(t *testing.T, casos []casoDeLecturaDeEval) {
	t.Helper()

	for _, caso := range casos {
		t.Run(caso.nombre, func(t *testing.T) {
			t.Parallel()

			leida, err := LeerEval(nombreDeEval, []byte(caso.documento))
			if caso.error == "" && len(caso.fragmentos) == 0 {
				require.NoError(t, err)
				assert.Equal(t, caso.leida, leida)

				return
			}

			require.Error(t, err)
			assert.Zero(t, leida, "una eval mal formada no se entrega a medias")

			if caso.error != "" {
				require.EqualError(t, err, caso.error)
			}

			for _, fragmento := range caso.fragmentos {
				assert.ErrorContains(t, err, fragmento)
			}
		})
	}
}

// probarElEnumeradoDeLosAvisosDelPlazo fija ComprobarCodigosDeAvisoDelPlazo: el
// esquema publicado enumera en plazo.avisos exactamente plazos.CodigosDeAviso; y
// sobre esquemas sintéticos compilados con skills.CompilarEsquema, con el
// enumerado de los cuatro detrás de un $ref no hay error; sin
// festivo-de-una-parte, el error lo nombra; con derogada de más, el error la
// nombra; y sin la propiedad plazo, uno por código en el orden de
// plazos.CodigosDeAviso.
func probarElEnumeradoDeLosAvisosDelPlazo(t *testing.T) {
	t.Parallel()

	publicado, err := esquemaDeEval()
	require.NoError(t, err)
	require.NoError(t, ComprobarCodigosDeAvisoDelPlazo(publicado),
		"el esquema publicado %s admite en plazo.avisos exactamente los c\xc3\xb3digos de aviso de plazos", rutaDelEsquemaDeEval)

	// conEnumerado es un esquema con plazo.avisos y ese enumerado detrás de un
	// $ref, como en el esquema publicado.
	conEnumerado := func(enumerado string) string {
		return `{
  "$schema": "https://json-schema.org/draft/2020-12/schema",
  "type": "object",
  "properties": {
    "plazo": {
      "type": "object",
      "properties": {
        "avisos": { "type": "array", "uniqueItems": true, "items": { "$ref": "#/$defs/codigo-de-aviso-del-plazo" } }
      }
    }
  },
  "$defs": { "codigo-de-aviso-del-plazo": { "enum": [` + enumerado + `] } }
}`
	}

	const losCuatro = `"faltan-festivos-nacionales-y-autonomicos", "faltan-festivos-locales", "dias-del-art-183-lopj"` +
		`, "festivo-de-una-parte"`

	casos := []struct {
		nombre  string
		esquema string

		// error es el error exacto; vacío, no hay error.
		error string
	}{
		{nombre: "exacto", esquema: conEnumerado(losCuatro)},
		{
			nombre: "falta-un-codigo",
			esquema: conEnumerado(`"faltan-festivos-nacionales-y-autonomicos", "faltan-festivos-locales", ` +
				`"dias-del-art-183-lopj"`),
			error: "el esquema de eval no enumera el c\xc3\xb3digo de aviso festivo-de-una-parte en plazo.avisos",
		},
		{
			nombre:  "sobra-un-codigo",
			esquema: conEnumerado(losCuatro + `, "derogada"`),
			error: "el esquema de eval enumera en plazo.avisos el c\xc3\xb3digo derogada, que no es un c\xc3\xb3digo " +
				"de aviso de plazos",
		},
		{
			nombre: "sin-plazo",
			esquema: `{
  "$schema": "https://json-schema.org/draft/2020-12/schema",
  "type": "object",
  "properties": { "pregunta": { "type": "string", "minLength": 1 } }
}`,
			error: "el esquema de eval no enumera el c\xc3\xb3digo de aviso faltan-festivos-nacionales-y-autonomicos " +
				"en plazo.avisos\n" +
				"el esquema de eval no enumera el c\xc3\xb3digo de aviso faltan-festivos-locales en plazo.avisos\n" +
				"el esquema de eval no enumera el c\xc3\xb3digo de aviso dias-del-art-183-lopj en plazo.avisos\n" +
				"el esquema de eval no enumera el c\xc3\xb3digo de aviso festivo-de-una-parte en plazo.avisos",
		},
	}

	for _, caso := range casos {
		t.Run(caso.nombre, func(t *testing.T) {
			t.Parallel()

			esquema, err := skills.CompilarEsquema([]byte(caso.esquema))
			require.NoError(t, err)

			err = ComprobarCodigosDeAvisoDelPlazo(esquema)
			if caso.error == "" {
				assert.NoError(t, err)

				return
			}

			assert.EqualError(t, err, caso.error)
		})
	}

	assert.Equal(t, []string{
		"faltan-festivos-nacionales-y-autonomicos", "faltan-festivos-locales", "dias-del-art-183-lopj",
		"festivo-de-una-parte",
	}, plazos.CodigosDeAviso(), "premisa: los cuatro c\xc3\xb3digos de los esquemas sint\xc3\xa9ticos son los de plazos")
}

// Trozos de las evals sintéticas de TestFormatoDeLasRedaccionesFechadas, cada
// uno con sus líneas completas: la pregunta de la eval (a) de H20, su comando de
// bloque con fecha, su cita y el principio de su redaccion_fechada, con la norma,
// el bloque y la fecha (contracts/evals-redaccion-fechada.md §1.3 de H20).
const (
	preguntaConFecha = "pregunta: \"¿qué exigía el artículo 118 de la LCSP el 1 de junio de 2019?\"\n"
	comandoConFecha  = "comandos:\n" +
		"  - applet: boe\n" +
		"    norma: BOE-A-2017-12902\n" +
		"    bloque: a1-30\n" +
		"    fecha: \"2019-06-01\"\n"
	citaDelBloqueConFecha = "citas:\n" +
		"  - norma: BOE-A-2017-12902\n" +
		"    bloque: a1-30\n"
	redaccionFechadaDel118 = "redaccion_fechada:\n" +
		"  norma: BOE-A-2017-12902\n" +
		"  bloque: a1-30\n" +
		"  fecha: \"2019-06-01\"\n"
)

// TestFormatoDeLasRedaccionesFechadas fija lo que H20 da al formato de eval
// (contracts/evals-redaccion-fechada.md §1.1 y §1.2; data-model §7.1 y §7.2;
// research D11; FR-051, FR-052): fecha en un comando de bloque de comandos, y
// redaccion_fechada, con su norma, su bloque y su fecha y, o fecha_vigencia con
// posterior o sin ella, o vigente_desde, se leen enteros. Se rechazan:
// fecha_vigencia con vigente_desde, ninguna de las dos, posterior sin
// fecha_vigencia o sin su norma modificadora, una fecha que no es AAAA-MM-DD en
// el comando o en lo esperado, una fecha de vigencia que no es AAAAMMDD, fecha en
// un comando de otra forma o en un comando de grafo_previo, y redaccion_fechada
// en una eval que no activa la skill o que es sin binario ni servidor, sin su
// fecha o con otra clave: cada documento es el bien formado con solo ese
// defecto.
func TestFormatoDeLasRedaccionesFechadas(t *testing.T) {
	t.Parallel()

	const (
		preguntaLeida = "¿qué exigía el artículo 118 de la LCSP el 1 de junio de 2019?"
		activa        = preguntaConFecha + "activa: true\n"

		diaDelEsquema   = "'^[0-9]{4}-(0[1-9]|1[0-2])-(0[1-9]|[12][0-9]|3[01])$'"
		fechaDelEsquema = "'^[0-9]{4}(0[1-9]|1[0-2])(0[1-9]|[12][0-9]|3[01])$'"

		// enLaRedaccionFechada es el principio de un defecto de redaccion_fechada,
		// con la línea de su primera clave.
		enLaRedaccionFechada = "redaccion_fechada, línea 12: "
	)

	// conRedaccion es el documento de la eval (a) con esas líneas bajo
	// redaccion_fechada, detrás de su fecha, que es la línea 14.
	conRedaccion := func(lineas string) string {
		return activa + comandoConFecha + citaDelBloqueConFecha + redaccionFechadaDel118 + lineas
	}

	// leida es la eval (a) leída con esa redacción fechada esperada.
	leida := func(redaccion RedaccionFechadaEsperada) Eval {
		return Eval{
			Fichero: nombreDeEval, Pregunta: preguntaLeida, Activa: true,
			Comandos: []ComandoEsperado{{
				Applet: "boe", Norma: normaDeLaLCSP, Bloque: bloqueDelArticulo118, Fecha: "2019-06-01",
			}},
			Citas:            []CitaEsperada{{Norma: normaDeLaLCSP, Bloque: bloqueDelArticulo118}},
			RedaccionFechada: &redaccion,
		}
	}

	delArticulo118 := RedaccionFechadaEsperada{Norma: normaDeLaLCSP, Bloque: bloqueDelArticulo118, Fecha: "2019-06-01"}

	probarLasLecturasDeEval(t, []casoDeLecturaDeEval{
		{
			// La eval (a): la redacción citada y la de hoy.
			nombre: "con-la-redaccion-posterior",
			documento: conRedaccion("  fecha_vigencia: \"20180309\"\n  posterior:\n    fecha_vigencia: \"20200206\"\n" +
				"    norma_modificadora: BOE-A-2020-1651\n"),
			leida: leida(cambiadaLaRedaccion(delArticulo118, func(r *RedaccionFechadaEsperada) {
				r.FechaVigencia = fechaDeLaRedaccionOriginal
				r.Posterior = &RedaccionPosteriorEsperada{
					FechaVigencia: fechaDelArticulo118Vigente, NormaModificadora: "BOE-A-2020-1651",
				}
			})),
		},
		{
			// La redacción citada es la de hoy: no hay posterior.
			nombre:    "sin-redaccion-posterior",
			documento: conRedaccion("  fecha_vigencia: \"20200206\"\n"),
			leida: leida(cambiadaLaRedaccion(delArticulo118, func(r *RedaccionFechadaEsperada) {
				r.FechaVigencia = fechaDelArticulo118Vigente
			})),
		},
		{
			// La forma de la eval (b): ese día el bloque no tenía redacción vigente.
			nombre:    "sin-redaccion-vigente",
			documento: conRedaccion("  vigente_desde: \"20180309\"\n"),
			leida: leida(cambiadaLaRedaccion(delArticulo118, func(r *RedaccionFechadaEsperada) {
				r.VigenteDesde = fechaDeLaRedaccionOriginal
			})),
		},
		{
			// Sin redaccion_fechada, el comando con fecha se lee igual.
			nombre:    "comando-con-fecha-sin-redaccion-fechada",
			documento: activa + comandoConFecha + citaDelBloqueConFecha,
			leida: Eval{
				Fichero: nombreDeEval, Pregunta: preguntaLeida, Activa: true,
				Comandos: []ComandoEsperado{{
					Applet: "boe", Norma: normaDeLaLCSP, Bloque: bloqueDelArticulo118, Fecha: "2019-06-01",
				}},
				Citas: []CitaEsperada{{Norma: normaDeLaLCSP, Bloque: bloqueDelArticulo118}},
			},
		},
		{
			// Cada rama del oneOf falla por su not: la primera, por vigente_desde,
			// y la segunda, por fecha_vigencia.
			nombre:    "fecha-vigencia-con-vigente-desde",
			documento: conRedaccion("  fecha_vigencia: \"20180309\"\n  vigente_desde: \"20180309\"\n"),
			error: nombreDeEval + ": " + enLaRedaccionFechada + "'not' failed\n" +
				enLaRedaccionFechada + "'not' failed",
		},
		{
			nombre:    "sin-fecha-vigencia-ni-vigente-desde",
			documento: conRedaccion(""),
			error: nombreDeEval + ": " + enLaRedaccionFechada + "missing property 'fecha_vigencia'\n" +
				enLaRedaccionFechada + "missing property 'vigente_desde'",
		},
		{
			// La segunda rama falla también por posterior.
			nombre: "posterior-sin-fecha-vigencia",
			documento: conRedaccion("  posterior:\n    fecha_vigencia: \"20200206\"\n" +
				"    norma_modificadora: BOE-A-2020-1651\n"),
			error: nombreDeEval + ": " + enLaRedaccionFechada + "'not' failed\n" +
				enLaRedaccionFechada + "missing property 'fecha_vigencia'\n" +
				enLaRedaccionFechada + "missing property 'vigente_desde'",
		},
		{
			nombre: "posterior-con-vigente-desde",
			documento: conRedaccion("  vigente_desde: \"20180309\"\n  posterior:\n    fecha_vigencia: \"20200206\"\n" +
				"    norma_modificadora: BOE-A-2020-1651\n"),
			error: nombreDeEval + ": " + enLaRedaccionFechada + "'not' failed\n" +
				enLaRedaccionFechada + "'not' failed\n" +
				enLaRedaccionFechada + "missing property 'fecha_vigencia'",
		},
		{
			nombre:    "posterior-sin-su-norma",
			documento: conRedaccion("  fecha_vigencia: \"20180309\"\n  posterior:\n    fecha_vigencia: \"20200206\"\n"),
			error:     nombreDeEval + ": redaccion_fechada/posterior, línea 17: missing property 'norma_modificadora'",
		},
		{
			nombre: "posterior-con-otra-norma",
			documento: conRedaccion("  fecha_vigencia: \"20180309\"\n  posterior:\n    fecha_vigencia: \"20200206\"\n" +
				"    norma_modificadora: Ley 9/2017\n"),
			error: nombreDeEval + ": redaccion_fechada/posterior/norma_modificadora, línea 18: " +
				"'Ley 9/2017' does not match pattern '^BOE-A-[0-9]{4}-[0-9]{1,9}$'",
		},
		{
			nombre: "fecha-del-comando-sin-guiones",
			documento: activa + "comandos:\n  - applet: boe\n    norma: BOE-A-2017-12902\n    bloque: a1-30\n" +
				"    fecha: \"20190601\"\n" + citaDelBloqueConFecha,
			fragmentos: []string{"comandos/0/fecha, línea 7: '20190601' does not match pattern " + diaDelEsquema},
		},
		{
			nombre: "fecha-del-comando-sin-ceros",
			documento: activa + "comandos:\n  - applet: boe\n    norma: BOE-A-2017-12902\n    bloque: a1-30\n" +
				"    fecha: \"2019-6-1\"\n" + citaDelBloqueConFecha,
			fragmentos: []string{"comandos/0/fecha, línea 7: '2019-6-1' does not match pattern " + diaDelEsquema},
		},
		{
			nombre: "fecha-esperada-sin-guiones",
			documento: activa + comandoConFecha + citaDelBloqueConFecha + "redaccion_fechada:\n" +
				"  norma: BOE-A-2017-12902\n  bloque: a1-30\n  fecha: \"20190601\"\n  fecha_vigencia: \"20180309\"\n",
			error: nombreDeEval + ": redaccion_fechada/fecha, línea 14: '20190601' does not match pattern " +
				diaDelEsquema,
		},
		{
			nombre:    "fecha-de-vigencia-con-guiones",
			documento: conRedaccion("  vigente_desde: \"2018-03-09\"\n"),
			fragmentos: []string{
				"redaccion_fechada/vigente_desde, línea 15: '2018-03-09' does not match pattern " + fechaDelEsquema,
			},
		},
		{
			nombre: "fecha-en-un-comando-de-otra-forma",
			documento: activa + "comandos:\n  - applet: boe\n    verbo: indice\n    norma: BOE-A-2017-12902\n" +
				"    fecha: \"2019-06-01\"\n" + citaDelBloqueConFecha,
			fragmentos: []string{"comandos/0, línea 4: additional properties 'fecha' not allowed"},
		},
		{
			nombre: "fecha-en-el-grafo-previo",
			documento: activa + "grafo_previo:\n  grabaciones: lcsp-redaccion-cambiada\n  comandos:\n" +
				"    - applet: boe\n      norma: BOE-A-2017-12902\n      bloque: a1-30\n      fecha: \"2019-06-01\"\n" +
				comandoConFecha + citaDelBloqueConFecha,
			error: nombreDeEval + ": grafo_previo/comandos/0, línea 6: additional properties 'fecha' not allowed",
		},
		{
			nombre:    "en-una-eval-que-no-activa-la-skill",
			documento: preguntaConFecha + "activa: false\n" + redaccionFechadaDel118 + "  fecha_vigencia: \"20180309\"\n",
			error:     nombreDeEval + ": línea 1: 'not' failed",
		},
		{
			nombre: "en-una-eval-sin-binario-ni-servidor",
			documento: activa + sinBinarioNiServidor + redaccionFechadaDel118 +
				"  fecha_vigencia: \"20180309\"\n",
			error: nombreDeEval + ": línea 1: 'not' failed",
		},
		{
			nombre: "sin-la-fecha",
			documento: activa + comandoConFecha + citaDelBloqueConFecha + "redaccion_fechada:\n" +
				"  norma: BOE-A-2017-12902\n  bloque: a1-30\n  fecha_vigencia: \"20180309\"\n",
			error: nombreDeEval + ": " + enLaRedaccionFechada + "missing property 'fecha'",
		},
		{
			nombre:    "con-otra-clave",
			documento: conRedaccion("  fecha_vigencia: \"20180309\"\n  avisos: [derogada]\n"),
			error:     nombreDeEval + ": " + enLaRedaccionFechada + "additional properties 'avisos' not allowed",
		},
	})
}

// cambiadaLaRedaccion es una copia de la redacción fechada esperada con los
// cambios aplicados.
func cambiadaLaRedaccion(
	redaccion RedaccionFechadaEsperada, cambiar func(*RedaccionFechadaEsperada),
) RedaccionFechadaEsperada {
	cambiar(&redaccion)

	return redaccion
}

// Trozos de las evals sintéticas de TestFormatoDeLecturas, cada uno con sus
// líneas completas (contracts/formato-de-eval.md §1 y §2 de H26.1): la pregunta
// del recurso de reposición, que activa la skill y ocupa las líneas 1 y 2; sus
// dos comandos de bloque de la LPAC, el art. 124 en una fecha y el art. 112 sin
// ella, y su cita, la del art. 21, que van de la línea 3 a la 13, así que la
// clave que va detrás es la línea 14 y su primera entrada, la 15; y una lectura
// del art. 25 de la LRJSP, que no es ni comando ni cita de la eval.
const (
	preguntaDeLasLecturas      = "pregunta: \"¿Hasta cuándo puedo recurrir en reposición?\"\nactiva: true\n"
	comandosYCitaDeLasLecturas = "comandos:\n" +
		"  - applet: boe\n    norma: BOE-A-2015-10565\n    bloque: a124\n    fecha: \"2026-09-01\"\n" +
		"  - applet: boe\n    norma: BOE-A-2015-10565\n    bloque: a112\n" +
		"citas:\n  - norma: BOE-A-2015-10565\n    bloque: a21\n"
	lecturaDelArticulo25 = "  - norma: BOE-A-2015-10566\n    bloque: a25\n" +
		"    motivo: La pregunta por la legalidad lleva al principio de la potestad sancionadora.\n"
)

// TestFormatoDeLecturas fija lo que H26.1 da al formato de eval
// (contracts/formato-de-eval.md §1 y §2; data-model §1, reglas 1 a 3; FR-008,
// FR-009, FR-012, FR-014; SC-004, SC-005): lecturas, con su norma, su bloque y su
// motivo, y sin_servir, con un bloque o con una consulta de norma, se leen
// enteros, con LeerEval y con LeerConjunto, y un bloque de sin_servir que es un
// comando de bloque de la eval se admite (FR-016). Se rechazan, cada uno con el
// nombre del fichero y la entrada o la ruta del esquema: una lectura sin motivo,
// con el motivo vacío o solo de blancos; una cuyo bloque ya es un comando de
// bloque de la eval, con fecha o sin ella, aunque no sea la primera, o una de sus
// citas; un bloque de sin_servir que es una cita; un verbo que no es indice ni
// metadatos; una entrada con bloque y verbo; y lecturas o sin_servir vacías: cada
// documento es el bien formado con solo ese defecto.
func TestFormatoDeLecturas(t *testing.T) {
	t.Parallel()

	const (
		lpac  = "BOE-A-2015-10565"
		lrjsp = "BOE-A-2015-10566"

		motivo = "La pregunta por la legalidad lleva al principio de la potestad sancionadora."

		// enLaPrimeraLectura y enLaPrimeraSinServir son el principio de un defecto
		// de la primera entrada de cada clave, con su línea, y enElMotivo, el de
		// un defecto del motivo de la primera lectura.
		enLaPrimeraLectura   = "lecturas/0, línea 15: "
		enLaPrimeraSinServir = "sin_servir/0, línea 15: "
		enElMotivo           = "lecturas/0/motivo, línea 17: "

		// patronDelMotivo es el pattern del motivo en el esquema, \S, como lo
		// escribe el error.
		patronDelMotivo = "'\\\\S'"
	)

	// conLas es el documento bien formado de la eval con esas líneas detrás de su
	// cita, desde la línea 14.
	conLas := func(lineas string) string { return preguntaDeLasLecturas + comandosYCitaDeLasLecturas + lineas }

	// leidaCon es lo que se lee del documento de conLas, con esas lecturas y esas
	// lecturas sin servir.
	leidaCon := func(lecturas []LecturaDeLaEval, sinServir []LecturaSinServir) Eval {
		return Eval{
			Fichero: nombreDeEval, Pregunta: "¿Hasta cuándo puedo recurrir en reposición?", Activa: true,
			Comandos: []ComandoEsperado{
				{Applet: "boe", Norma: lpac, Bloque: "a124", Fecha: "2026-09-01"},
				{Applet: "boe", Norma: lpac, Bloque: "a112"},
			},
			Citas:     []CitaEsperada{{Norma: lpac, Bloque: "a21"}},
			Lecturas:  lecturas,
			SinServir: sinServir,
		}
	}

	casos := []casoDeLecturaDeEval{
		{
			nombre: "las-dos-claves",
			documento: conLas("lecturas:\n" + lecturaDelArticulo25 +
				"sin_servir:\n  - norma: BOE-A-2015-10565\n    bloque: a30\n" +
				"  - norma: BOE-A-2015-10565\n    verbo: indice\n  - norma: BOE-A-2015-10566\n    verbo: metadatos\n"),
			leida: leidaCon(
				[]LecturaDeLaEval{{Norma: lrjsp, Bloque: "a25", Motivo: motivo}},
				[]LecturaSinServir{
					{Norma: lpac, Bloque: "a30"}, {Norma: lpac, Verbo: "indice"}, {Norma: lrjsp, Verbo: "metadatos"},
				},
			),
		},
		{
			// Es lo que la eval quiere provocar: que la lectura de un comando
			// suyo, con fecha o sin ella, no se sirva (FR-016).
			nombre: "sin-servir-los-comandos-de-bloque",
			documento: conLas("sin_servir:\n  - norma: BOE-A-2015-10565\n    bloque: a124\n" +
				"  - norma: BOE-A-2015-10565\n    bloque: a112\n"),
			leida: leidaCon(nil, []LecturaSinServir{{Norma: lpac, Bloque: "a124"}, {Norma: lpac, Bloque: "a112"}}),
		},
		{
			nombre:    "lectura-sin-motivo",
			documento: conLas("lecturas:\n  - norma: BOE-A-2015-10566\n    bloque: a25\n"),
			error:     nombreDeEval + ": " + enLaPrimeraLectura + "missing property 'motivo'",
		},
		{
			nombre:    "lectura-con-el-motivo-vacio",
			documento: conLas("lecturas:\n  - norma: BOE-A-2015-10566\n    bloque: a25\n    motivo: \"\"\n"),
			error:     nombreDeEval + ": " + enElMotivo + "'' does not match pattern " + patronDelMotivo,
		},
		{
			nombre:    "lectura-con-el-motivo-de-blancos",
			documento: conLas("lecturas:\n  - norma: BOE-A-2015-10566\n    bloque: a25\n    motivo: \" \\t \"\n"),
			error:     nombreDeEval + ": " + enElMotivo + "' \\t ' does not match pattern " + patronDelMotivo,
		},
		{
			nombre:    "lectura-de-un-comando-con-fecha",
			documento: conLas("lecturas:\n  - norma: BOE-A-2015-10565\n    bloque: a124\n    motivo: " + motivo + "\n"),
			error:     nombreDeEval + ": lecturas: BOE-A-2015-10565 a124: ya es un comando esperado de la eval",
		},
		{
			nombre: "lectura-de-un-comando-sin-fecha-detras-de-otra",
			documento: conLas("lecturas:\n" + lecturaDelArticulo25 +
				"  - norma: BOE-A-2015-10565\n    bloque: a112\n    motivo: " + motivo + "\n"),
			error: nombreDeEval + ": lecturas: BOE-A-2015-10565 a112: ya es un comando esperado de la eval",
		},
		{
			nombre:    "lectura-de-una-cita",
			documento: conLas("lecturas:\n  - norma: BOE-A-2015-10565\n    bloque: a21\n    motivo: " + motivo + "\n"),
			error:     nombreDeEval + ": lecturas: BOE-A-2015-10565 a21: ya es una cita esperada de la eval",
		},
		{
			nombre:    "sin-servir-una-cita",
			documento: conLas("sin_servir:\n  - norma: BOE-A-2015-10565\n    bloque: a21\n"),
			error:     nombreDeEval + ": sin_servir: BOE-A-2015-10565 a21: es una cita esperada de la eval",
		},
		{
			// analisis es un verbo de consulta de norma de los comandos, pero no
			// de sin_servir.
			nombre:     "sin-servir-con-otro-verbo",
			documento:  conLas("sin_servir:\n  - norma: BOE-A-2015-10565\n    verbo: analisis\n"),
			fragmentos: []string{"sin_servir/0/verbo, línea 16: value must be one of 'indice', 'metadatos'"},
		},
		{
			nombre:    "sin-servir-con-bloque-y-verbo",
			documento: conLas("sin_servir:\n  - norma: BOE-A-2015-10565\n    bloque: a30\n    verbo: indice\n"),
			fragmentos: []string{
				enLaPrimeraSinServir + "additional properties 'bloque' not allowed",
				enLaPrimeraSinServir + "additional properties 'verbo' not allowed",
			},
		},
		{
			nombre:    "lecturas-vacias",
			documento: conLas("lecturas: []\n"),
			error:     nombreDeEval + ": lecturas, línea 14: minItems: got 0, want 1",
		},
		{
			nombre:    "sin-servir-vacio",
			documento: conLas("sin_servir: []\n"),
			error:     nombreDeEval + ": sin_servir, línea 14: minItems: got 0, want 1",
		},
	}

	probarLasLecturasDeEval(t, casos)

	t.Run("con-leer-conjunto", func(t *testing.T) {
		t.Parallel()

		probarLasLecturasDelConjunto(t, casos)
	})
}

// probarLasLecturasDelConjunto escribe el documento de cada caso en un mismo
// directorio temporal, cada uno con su número y su nombre, <nn>-<nombre>.yaml, y
// los lee con LeerConjunto: el bien formado está entre las evals, con su nombre
// en Fichero, y el mal formado, entre los ficheros mal formados, con el error
// del caso, que empieza por su nombre y no por nombreDeEval.
func probarLasLecturasDelConjunto(t *testing.T, casos []casoDeLecturaDeEval) {
	t.Helper()

	fichero := func(posicion int, caso casoDeLecturaDeEval) string {
		return fmt.Sprintf("%02d-%s.yaml", posicion+1, caso.nombre)
	}

	entradas := make([]entradaDeConjunto, 0, len(casos))
	for posicion, caso := range casos {
		entradas = append(entradas, entradaDeConjunto{nombre: fichero(posicion, caso), contenido: caso.documento})
	}

	conjunto, err := LeerConjunto(crearConjunto(t, entradas))
	require.NoError(t, err)

	var bienFormadas, malFormados int

	for posicion, caso := range casos {
		nombre := fichero(posicion, caso)

		if caso.error == "" && len(caso.fragmentos) == 0 {
			bienFormadas++
			leida := caso.leida
			leida.Fichero = nombre
			assert.Contains(t, conjunto.Evals, leida, "%s se lee entera", nombre)

			continue
		}

		malFormados++
		indice := slices.IndexFunc(conjunto.MalFormados, func(mal FicheroMalFormado) bool { return mal.Fichero == nombre })
		require.NotEqual(t, -1, indice, "%s es un fichero mal formado", nombre)

		errDelFichero := conjunto.MalFormados[indice].Error
		require.Error(t, errDelFichero)
		assert.True(t, strings.HasPrefix(errDelFichero.Error(), nombre+": "), "el error de %s empieza por su nombre", nombre)

		if caso.error != "" {
			require.EqualError(t, errDelFichero, nombre+strings.TrimPrefix(caso.error, nombreDeEval))
		}

		for _, fragmento := range caso.fragmentos {
			require.ErrorContains(t, errDelFichero, fragmento)
		}
	}

	assert.Len(t, conjunto.Evals, bienFormadas)
	assert.Len(t, conjunto.MalFormados, malFormados)
}

// TestEsquemaDeEval comprueba que el esquema publicado del formato común de eval
// compila con las aserciones de formato activas. Lo que dicen sus patrones de
// norma y de bloque lo fija TestGramaticasCoincidenConBoe.
//
// Desde H7, comprueba además que el formato ampliado es compatible hacia atrás
// (FR-086; contrato evals-y-skill §1 de H7): cada eval del repositorio de
// boe-legislacion y de legal-core se lee y valida; la que no escribe prohibidos
// ni grafo_previo se lee sin prohibidos y sin grafo previo; y solo el comando que
// escribe el verbo check tiene la forma de comprobación, de modo que el resto
// conserva la suya.
//
// Desde H7.1, la que no escribe hallazgos se lee sin hallazgos, y el comando de
// comprobación que no escribe la norma, sin Norma: se leen igual que en H7
// (contrato evals-y-skill §1 de H7.1; FR-050, FR-054).
//
// Desde H7.4, la que no escribe no_se_activan ni redacciones_modificadas se lee
// sin ellas, así que su juicio no cambia (contracts/evals-y-juicio.md §1; FR-053).
//
// Desde H21, la que no escribe sin_binario_ni_servidor se lee sin ella, y la
// rama del esquema de esa clave es la de contracts/evals-en-dos-modos.md §1
// (probarLaRamaSinBinarioNiServidor; FR-046).
func TestEsquemaDeEval(t *testing.T) {
	t.Parallel()

	esquema, err := esquemaDeEval()
	require.NoError(t, err)
	assert.NotNil(t, esquema)

	for _, dir := range []string{evalsDelRepositorio, evalsDeLegalCore} {
		conjunto, err := LeerConjunto(dir)
		require.NoError(t, err)
		require.Empty(t, conjunto.MalFormados, "cada fichero de %s es una eval bien formada", dir)
		require.NotEmpty(t, conjunto.Evals, "%s tiene evals", dir)

		for _, eval := range conjunto.Evals {
			ruta := filepath.Join(dir, eval.Fichero)

			contenido, err := leerFichero(ruta)
			require.NoError(t, err)

			var claves map[string]any
			require.NoError(t, yaml.Unmarshal(contenido, &claves), "%s es un documento YAML", ruta)

			if _, escribe := claves["prohibidos"]; !escribe {
				assert.Nil(t, eval.Prohibidos, "%s no escribe prohibidos y se lee sin ellos", ruta)
			}

			if _, escribe := claves["grafo_previo"]; !escribe {
				assert.Zero(t, eval.GrafoPrevio, "%s no escribe grafo_previo y se lee sin grafo previo", ruta)
			}

			if _, escribe := claves["hallazgos"]; !escribe {
				assert.Nil(t, eval.Hallazgos, "%s no escribe hallazgos y se lee sin ellos", ruta)
			}

			if _, escribe := claves["no_se_activan"]; !escribe {
				assert.Nil(t, eval.NoSeActivan, "%s no escribe no_se_activan y se lee sin ellas", ruta)
			}

			if _, escribe := claves["redacciones_modificadas"]; !escribe {
				assert.Nil(t, eval.RedaccionesModificadas, "%s no escribe redacciones_modificadas y se lee sin ellas", ruta)
			}

			_, escribeSinBinario := claves["sin_binario_ni_servidor"]
			assert.Equal(t, escribeSinBinario, eval.SinBinarioNiServidor,
				"%s se lee sin binario ni servidor si y solo si escribe sin_binario_ni_servidor", ruta)

			for posicion, comando := range eval.Comandos {
				assert.Equal(t, comando.Verbo == "check", formaDelComando(comando) == formaComprobacion,
					"el comando %d de %s tiene la forma de comprobación si y solo si su verbo es check", posicion, ruta)
			}

			exigirComprobacionesSinNormaLeidasSinElla(t, ruta, claves, eval.Comandos)
		}
	}

	t.Run("sin-binario-ni-servidor", probarLaRamaSinBinarioNiServidor)
}

// probarLaRamaSinBinarioNiServidor fija la rama del esquema de la clave
// sin_binario_ni_servidor (contracts/evals-en-dos-modos.md §1 y §8 de H21;
// FR-046): con activa: true y nada más, la eval es válida, también con las
// skills que su sesión no puede activar; con cualquiera de las diez claves de lo
// que una sesión consulta o de lo que su respuesta tiene que llevar, con activa:
// false o con el valor false, no. Cada clave se prueba con un valor que el
// formato admite en una eval positiva, de modo que el rechazo solo puede ser el
// de la rama.
func probarLaRamaSinBinarioNiServidor(t *testing.T) {
	t.Parallel()

	const norma, bloque = "BOE-A-2015-10565", "a21"

	comando := map[string]any{"applet": "boe", "norma": norma, "bloque": bloque}
	cita := map[string]any{"norma": norma, "bloque": bloque}

	// La eval sin binario ni servidor bien formada, con esa clave puesta a ese
	// valor: una de las suyas o una que no lleva.
	sinBinario := func(clave string, valor any) map[string]any {
		documento := map[string]any{
			"pregunta": preguntaSinBinarioDeBoeLegislacion, "activa": true, "sin_binario_ni_servidor": true,
		}
		documento[clave] = valor

		return documento
	}

	assert.True(t, aceptaLaEval(t, sinBinario("activa", true)),
		"la clave con activa: true y nada m\xc3\xa1s es una eval v\xc3\xa1lida")
	assert.True(t, aceptaLaEval(t, sinBinario("no_se_activan", []any{"boe-legislacion"})),
		"la eval sin binario ni servidor admite las skills que su sesi\xc3\xb3n no puede activar")
	assert.False(t, aceptaLaEval(t, sinBinario("activa", false)),
		"una eval de no activaci\xc3\xb3n no puede ser sin binario ni servidor")
	assert.False(t, aceptaLaEval(t, sinBinario("sin_binario_ni_servidor", false)),
		"la clave solo admite el valor true")

	excluidas := map[string]any{
		"comandos":     []any{comando},
		"prohibidos":   []any{map[string]any{"applet": "graph", "verbo": "show"}},
		"grafo_previo": map[string]any{"grabaciones": "lcsp-a1-30-redaccion-original", "comandos": []any{comando}},
		"citas":        []any{cita},
		"avisos":       []any{"derogada"},
		"hallazgos":    []any{"version-obsoleta"},
		"redacciones_modificadas": []any{map[string]any{
			"norma": norma, "bloque": bloque, "fecha_vigencia": "20180309", "fecha_vigencia_reciente": "20200206",
		}},
		"territorio":  map[string]any{"comunidad": "Comunidad de Madrid"},
		"informativa": true,
		"reproduce":   "boe-fiscal",
		"por_materia": true,
	}

	for clave, valor := range excluidas {
		positiva := evalPositiva(comando, cita)
		positiva[clave] = valor

		require.True(t, aceptaLaEval(t, positiva),
			"una eval positiva admite %s con ese valor: sin eso, un rechazo no dir\xc3\xada nada de la rama", clave)
		assert.False(t, aceptaLaEval(t, sinBinario(clave, valor)),
			"una eval sin binario ni servidor no admite %s", clave)
	}
}

// exigirComprobacionesSinNormaLeidasSinElla exige que cada comando de
// comprobación que la eval de la ruta escribe sin la clave norma se lea sin
// Norma: las claves son las del documento YAML de la eval, y los comandos, los
// que se leyeron de él, en el mismo orden.
func exigirComprobacionesSinNormaLeidasSinElla(t *testing.T, ruta string, claves map[string]any,
	comandos []ComandoEsperado,
) {
	t.Helper()

	// Sin la clave comandos, como en una eval de no activación, no hay ninguno.
	escritos, _ := claves["comandos"].([]any)
	require.Len(t, escritos, len(comandos), "%s escribe tantos comandos como se leen de ella", ruta)

	for posicion, comando := range comandos {
		escrito, esMapa := escritos[posicion].(map[string]any)
		require.True(t, esMapa, "el comando %d de %s es un mapa", posicion, ruta)

		if _, escribe := escrito["norma"]; comando.Verbo == verboCheck && !escribe {
			assert.Empty(t, comando.Norma, "el comando %d de %s no escribe la norma y se lee sin ella", posicion, ruta)
		}
	}
}

// TestCompilarEsquemaDeEvalQueNoSirve fija los dos errores con los que no se
// obtiene el esquema del formato de eval, sobre la ruta de un directorio temporal
// del test: una carpeta donde se espera el fichero del esquema y un fichero JSON
// que no es un JSON Schema válido. Los dos nombran la ruta.
func TestCompilarEsquemaDeEvalQueNoSirve(t *testing.T) {
	t.Parallel()

	casos := []struct {
		nombre string

		// contenido es el del fichero del esquema; vacío, en su lugar hay una
		// carpeta.
		contenido string

		motivo string
	}{
		{nombre: "carpeta", motivo: "no se puede leer el esquema del formato de eval"},
		{
			nombre:    "tipo-que-no-es-de-json-schema",
			contenido: `{"$schema": "https://json-schema.org/draft/2020-12/schema", "type": "entero"}`,
			motivo:    "el esquema no compila",
		},
	}

	for _, caso := range casos {
		t.Run(caso.nombre, func(t *testing.T) {
			t.Parallel()

			ruta := filepath.Join(t.TempDir(), "eval.yaml.json")
			if caso.contenido == "" {
				require.NoError(t, os.Mkdir(ruta, 0o750))
			} else {
				require.NoError(t, os.WriteFile(ruta, []byte(caso.contenido), 0o600))
			}

			esquema, err := compilarEsquemaDeEval(ruta)

			require.ErrorContains(t, err, ruta)
			require.ErrorContains(t, err, caso.motivo)
			assert.Nil(t, esquema)
		})
	}
}

// patronDeExpresion es el patrón de cada expresión de la lista de expresiones
// prohibidas: una o más palabras separadas por un espacio, sin blancos en los
// extremos y sin * ni _ (contrato lista-y-juicio §1; research D2).
const patronDeExpresion = `^[^\s*_]+( [^\s*_]+)*$`

// patronDeFormaFija es el patrón de cada forma fija de la lista: una línea con
// algún carácter que no es un blanco ni un corchete angular de los marcadores
// (contracts/lista-de-expresiones.md §2 de H7.4).
const patronDeFormaFija = `^[^\n]*[^\s<>][^\n]*$`

// patronDeSalidaDeLasHerramientas es el patrón de cada expresión de la clave
// salida_de_las_herramientas de la lista: una o más palabras separadas por un
// espacio, sin blancos en los extremos y sin * ni acento grave. Admite el guion
// bajo de un nombre de campo, que el de las familias no admite
// (contracts/skill-boe-legislacion.md §4 de H24).
const patronDeSalidaDeLasHerramientas = "^[^\\s*`]+( [^\\s*`]+)*$"

// Claves bien formadas de una lista de expresiones prohibidas, cada una con sus
// líneas completas, para los documentos de TestEsquemaDeExpresionesProhibidas y
// de TestLeerConjunto: expresiones de una palabra y de varias, y con letras que
// no son ASCII; lo que devuelven las herramientas, con varias palabras y con el
// nombre de un campo; y formas fijas con los dos marcadores y sin ninguno.
const (
	maquinariaBienFormada = "maquinaria:\n" +
		"  - memoria de consultas\n" +
		"  - hallazgos\n" +
		"  - c\xc3\xb3digo de salida\n"
	otraConversacionBienFormada = "otra_conversacion:\n" +
		"  - te dije\n" +
		"  - conversaci\xc3\xb3n anterior\n"
	anuncioBienFormado = "anuncio:\n" +
		"  - que trasladar\n" +
		"  - ya puedo responder\n" +
		"  - as\xc3\xad que respondo\n"
	redaccionNoLeidaBienFormada = "redaccion_no_leida:\n" +
		"  - ya no exige\n" +
		"  - se elimin\xc3\xb3\n"
	salidaDeLasHerramientasBienFormada = "salida_de_las_herramientas:\n" +
		"  - el sobre\n" +
		"  - fecha_vigencia\n"
	formasFijasBienFormadas = "formas_fijas:\n" +
		"  - \"\xe2\x9a\xa0 REDACCI\xc3\x93N MODIFICADA: <cita>: vigente desde <fecha>.\"\n" +
		"  - \"No se ha podido comprobar si la redacci\xc3\xb3n ha cambiado\"\n"

	// listaBienFormada es la lista con sus seis claves bien formadas, en el
	// orden del fichero del repositorio.
	listaBienFormada = maquinariaBienFormada + otraConversacionBienFormada + anuncioBienFormado +
		redaccionNoLeidaBienFormada + salidaDeLasHerramientasBienFormada + formasFijasBienFormadas
)

// TestEsquemaDeExpresionesProhibidas fija el esquema publicado de la lista de
// expresiones prohibidas de una skill (contrato lista-y-juicio §1 de H7.2 y
// lista-de-expresiones §2 de H7.3; FR-050, FR-055 de H7.2 y FR-023 de H7.3): lo
// lee de su ruta, lo compila con skills.CompilarEsquema y valida contra él, con
// el lector común de documentos YAML, listas escritas en el test. Una lista con
// sus seis claves valida y se lee entera. Sin una familia —también sin la de
// anuncio, la lista de dos familias de H7.2—, con una familia vacía, con una
// clave de más y con una expresión con un blanco en un extremo, con dos blancos
// entre palabras, con * o con _ no valida, y cada rechazo es el incumplimiento
// de ese defecto en su sitio: la lista de cada caso es la bien formada con solo
// ese defecto, así que no la rechaza otro motivo.
//
// Desde H7.4 (contracts/lista-de-expresiones.md §2; FR-030, FR-031), la lista
// tiene además la familia redaccion_no_leida, con la misma forma que las otras
// tres, y las formas_fijas, las dos obligatorias: la lista del repositorio
// valida; sin redaccion_no_leida o sin formas_fijas, con cualquiera de las dos
// vacía, con una expresión suya con * y con una forma fija de dos líneas, no.
//
// Desde H24 (contracts/skill-boe-legislacion.md §4; FR-085, FR-110), la lista
// tiene además la clave salida_de_las_herramientas, obligatoria, con el patrón
// de las familias salvo que admite el guion bajo de un nombre de campo y no el
// acento grave: con fecha_vigencia valida; sin la clave —la lista de cinco
// claves de H7.4—, con ella vacía y con un elemento con acento grave o con *,
// no. Las demás claves no cambian: el guion bajo sigue sin valer en una familia.
func TestEsquemaDeExpresionesProhibidas(t *testing.T) {
	t.Parallel()

	// rutaDelEsquema es el esquema publicado de la lista, relativo al directorio
	// de este paquete, que es donde go test ejecuta los tests.
	const rutaDelEsquema = "../../schemas/expresiones-prohibidas.yaml.json"

	contenido, err := leerFichero(rutaDelEsquema)
	require.NoError(t, err, "el esquema publicado %s se lee", rutaDelEsquema)

	esquema, err := skills.CompilarEsquema(contenido)
	require.NoError(t, err, "el esquema publicado %s compila", rutaDelEsquema)

	leida, err := skills.ValidarDocumentoYAML[map[string][]string]([]byte(listaBienFormada), esquema)
	require.NoError(t, err, "la lista con sus seis claves valida: sin eso, un rechazo no diría nada del esquema")
	require.Equal(t, map[string][]string{
		"maquinaria":                 {"memoria de consultas", "hallazgos", "c\xc3\xb3digo de salida"},
		"otra_conversacion":          {"te dije", "conversaci\xc3\xb3n anterior"},
		"anuncio":                    {"que trasladar", "ya puedo responder", "as\xc3\xad que respondo"},
		"redaccion_no_leida":         {"ya no exige", "se elimin\xc3\xb3"},
		"salida_de_las_herramientas": {"el sobre", "fecha_vigencia"},
		"formas_fijas": {
			"\xe2\x9a\xa0 REDACCI\xc3\x93N MODIFICADA: <cita>: vigente desde <fecha>.",
			"No se ha podido comprobar si la redacci\xc3\xb3n ha cambiado",
		},
	}, leida, "la lista con sus seis claves se lee entera")

	delRepositorio := filepath.Join(evalsDelRepositorio, ficheroDeExpresionesProhibidas)
	contenidoDelRepositorio, err := leerFichero(delRepositorio)
	require.NoError(t, err, "la lista del repositorio %s se lee", delRepositorio)

	_, err = skills.ValidarDocumentoYAML[map[string][]string](contenidoDelRepositorio, esquema)
	require.NoError(t, err, "la lista del repositorio %s cumple el esquema", delRepositorio)

	// tresFamilias son las tres familias de H7.3 bien formadas; conLaClaveDeH24,
	// el documento de un caso con la clave de H24 bien formada detrás; y
	// conLasDemasClaves, el de un caso de esas tres familias con las dos claves de
	// H7.4 y la de H24 bien formadas detrás: así el defecto de cada caso es el
	// único.
	const tresFamilias = maquinariaBienFormada + otraConversacionBienFormada + anuncioBienFormado

	conLaClaveDeH24 := func(documento string) string {
		return documento + salidaDeLasHerramientasBienFormada
	}

	conLasDemasClaves := func(documento string) string {
		return conLaClaveDeH24(documento + redaccionNoLeidaBienFormada + formasFijasBienFormadas)
	}

	casos := []struct {
		nombre    string
		documento string

		// ruta es la del punto de la lista donde el esquema la rechaza; nil, en
		// la raíz.
		ruta []string

		// rechazo es el incumplimiento con el que el esquema la rechaza.
		rechazo jsonschema.ErrorKind
	}{
		{
			nombre:    "sin-maquinaria",
			documento: conLasDemasClaves(otraConversacionBienFormada + anuncioBienFormado),
			rechazo:   &kind.Required{Missing: []string{"maquinaria"}},
		},
		{
			nombre:    "sin-otra-conversacion",
			documento: conLasDemasClaves(maquinariaBienFormada + anuncioBienFormado),
			rechazo:   &kind.Required{Missing: []string{"otra_conversacion"}},
		},
		{
			nombre:    "sin-anuncio",
			documento: conLasDemasClaves(maquinariaBienFormada + otraConversacionBienFormada),
			rechazo:   &kind.Required{Missing: []string{"anuncio"}},
		},
		{
			nombre:    "sin-redaccion-no-leida",
			documento: conLaClaveDeH24(tresFamilias + formasFijasBienFormadas),
			rechazo:   &kind.Required{Missing: []string{"redaccion_no_leida"}},
		},
		{
			nombre:    "sin-formas-fijas",
			documento: conLaClaveDeH24(tresFamilias + redaccionNoLeidaBienFormada),
			rechazo:   &kind.Required{Missing: []string{"formas_fijas"}},
		},
		{
			nombre:    "maquinaria-vacia",
			documento: conLasDemasClaves("maquinaria: []\n" + otraConversacionBienFormada + anuncioBienFormado),
			ruta:      []string{"maquinaria"},
			rechazo:   &kind.MinItems{Got: 0, Want: 1},
		},
		{
			nombre:    "otra-conversacion-vacia",
			documento: conLasDemasClaves(maquinariaBienFormada + "otra_conversacion: []\n" + anuncioBienFormado),
			ruta:      []string{"otra_conversacion"},
			rechazo:   &kind.MinItems{Got: 0, Want: 1},
		},
		{
			nombre:    "anuncio-vacio",
			documento: conLasDemasClaves(maquinariaBienFormada + otraConversacionBienFormada + "anuncio: []\n"),
			ruta:      []string{"anuncio"},
			rechazo:   &kind.MinItems{Got: 0, Want: 1},
		},
		{
			nombre:    "redaccion-no-leida-vacia",
			documento: conLaClaveDeH24(tresFamilias + "redaccion_no_leida: []\n" + formasFijasBienFormadas),
			ruta:      []string{"redaccion_no_leida"},
			rechazo:   &kind.MinItems{Got: 0, Want: 1},
		},
		{
			nombre:    "formas-fijas-vacias",
			documento: conLaClaveDeH24(tresFamilias + redaccionNoLeidaBienFormada + "formas_fijas: []\n"),
			ruta:      []string{"formas_fijas"},
			rechazo:   &kind.MinItems{Got: 0, Want: 1},
		},
		{
			nombre:    "clave-de-mas",
			documento: conLasDemasClaves(tresFamilias) + "avisos:\n  - derogada\n",
			rechazo:   &kind.AdditionalProperties{Properties: []string{"avisos"}},
		},
		{
			nombre: "blanco-al-principio",
			documento: conLasDemasClaves("maquinaria:\n  - \" hallazgos\"\n" + otraConversacionBienFormada +
				anuncioBienFormado),
			ruta:    []string{"maquinaria", "0"},
			rechazo: &kind.Pattern{Got: " hallazgos", Want: patronDeExpresion},
		},
		{
			nombre: "blanco-al-final",
			documento: conLasDemasClaves(maquinariaBienFormada + "otra_conversacion:\n  - \"te dije \"\n" +
				anuncioBienFormado),
			ruta:    []string{"otra_conversacion", "0"},
			rechazo: &kind.Pattern{Got: "te dije ", Want: patronDeExpresion},
		},
		{
			nombre: "dos-blancos-entre-palabras",
			documento: conLasDemasClaves("maquinaria:\n  - \"memoria  de consultas\"\n" + otraConversacionBienFormada +
				anuncioBienFormado),
			ruta:    []string{"maquinaria", "0"},
			rechazo: &kind.Pattern{Got: "memoria  de consultas", Want: patronDeExpresion},
		},
		{
			nombre: "asterisco",
			documento: conLasDemasClaves(maquinariaBienFormada + "otra_conversacion:\n  - \"te *dije*\"\n" +
				anuncioBienFormado),
			ruta:    []string{"otra_conversacion", "0"},
			rechazo: &kind.Pattern{Got: "te *dije*", Want: patronDeExpresion},
		},
		{
			nombre: "asterisco-en-anuncio",
			documento: conLasDemasClaves(maquinariaBienFormada + otraConversacionBienFormada +
				"anuncio:\n  - \"ya *puedo* responder\"\n"),
			ruta:    []string{"anuncio", "0"},
			rechazo: &kind.Pattern{Got: "ya *puedo* responder", Want: patronDeExpresion},
		},
		{
			nombre: "asterisco-en-redaccion-no-leida",
			documento: conLaClaveDeH24(tresFamilias + "redaccion_no_leida:\n  - \"ya no *exige*\"\n" +
				formasFijasBienFormadas),
			ruta:    []string{"redaccion_no_leida", "0"},
			rechazo: &kind.Pattern{Got: "ya no *exige*", Want: patronDeExpresion},
		},
		{
			nombre: "guion-bajo",
			documento: conLasDemasClaves("maquinaria:\n  - \"_hallazgos_\"\n" + otraConversacionBienFormada +
				anuncioBienFormado),
			ruta:    []string{"maquinaria", "0"},
			rechazo: &kind.Pattern{Got: "_hallazgos_", Want: patronDeExpresion},
		},
		{
			// Una forma fija partida en dos líneas: el salto, escrito con su escape
			// en la cadena entre comillas del YAML.
			nombre: "forma-fija-de-dos-lineas",
			documento: conLaClaveDeH24(tresFamilias + redaccionNoLeidaBienFormada +
				"formas_fijas:\n  - \"No se ha podido comprobar\\nsi la redacci\xc3\xb3n ha cambiado\"\n"),
			ruta: []string{"formas_fijas", "0"},
			rechazo: &kind.Pattern{
				Got:  "No se ha podido comprobar\nsi la redacci\xc3\xb3n ha cambiado",
				Want: patronDeFormaFija,
			},
		},
		{
			// La lista de cinco claves de H7.4, sin la de H24.
			nombre:    "sin-salida-de-las-herramientas",
			documento: tresFamilias + redaccionNoLeidaBienFormada + formasFijasBienFormadas,
			rechazo:   &kind.Required{Missing: []string{"salida_de_las_herramientas"}},
		},
		{
			nombre: "salida-de-las-herramientas-vacia",
			documento: tresFamilias + redaccionNoLeidaBienFormada + "salida_de_las_herramientas: []\n" +
				formasFijasBienFormadas,
			ruta:    []string{"salida_de_las_herramientas"},
			rechazo: &kind.MinItems{Got: 0, Want: 1},
		},
		{
			// El nombre de un campo escrito como código en línea: en la lista va
			// sin los acentos graves, que es como se busca en la prosa.
			nombre: "acento-grave-en-salida-de-las-herramientas",
			documento: tresFamilias + redaccionNoLeidaBienFormada +
				"salida_de_las_herramientas:\n  - \"`fecha_vigencia`\"\n" + formasFijasBienFormadas,
			ruta:    []string{"salida_de_las_herramientas", "0"},
			rechazo: &kind.Pattern{Got: "`fecha_vigencia`", Want: patronDeSalidaDeLasHerramientas},
		},
		{
			nombre: "asterisco-en-salida-de-las-herramientas",
			documento: tresFamilias + redaccionNoLeidaBienFormada +
				"salida_de_las_herramientas:\n  - \"el *sobre*\"\n" + formasFijasBienFormadas,
			ruta:    []string{"salida_de_las_herramientas", "0"},
			rechazo: &kind.Pattern{Got: "el *sobre*", Want: patronDeSalidaDeLasHerramientas},
		},
	}

	for _, caso := range casos {
		t.Run(caso.nombre, func(t *testing.T) {
			t.Parallel()

			_, err := skills.ValidarDocumentoYAML[map[string][]string]([]byte(caso.documento), esquema)

			var defecto *skills.DefectoEnElDocumento
			require.ErrorAs(t, err, &defecto, "el esquema no valida la lista")
			assert.Equal(t, caso.ruta, defecto.Ruta, "el esquema la rechaza en su sitio: %v", err)
			assert.Equal(t, caso.rechazo, defecto.Tipo, "el esquema la rechaza por ese defecto: %v", err)
		})
	}
}

// casoDeGramatica es un valor límite de una gramática de identificador y si
// boe lo acepta (data-model, cabecera).
type casoDeGramatica struct {
	nombre string
	valor  string
	valido bool
}

// casosDeNorma son los casos límite de NORMA, ^BOE-A-[0-9]{4}-[0-9]{1,9}$.
var casosDeNorma = []casoDeGramatica{
	{nombre: "lpac", valor: "BOE-A-2015-10565", valido: true},
	{nombre: "número-de-una-cifra", valor: "BOE-A-2015-1", valido: true},
	{nombre: "número-de-nueve-cifras", valor: "BOE-A-2015-123456789", valido: true},
	{nombre: "número-de-diez-cifras", valor: "BOE-A-2015-1234567890"},
	{nombre: "año-de-dos-cifras", valor: "BOE-A-15-1"},
	{nombre: "año-de-tres-cifras", valor: "BOE-A-201-10565"},
	{nombre: "año-de-cinco-cifras", valor: "BOE-A-20150-10565"},
	{nombre: "sin-número", valor: "BOE-A-2015-"},
	{nombre: "vacía", valor: ""},
	{nombre: "minúsculas", valor: "boe-a-2015-10565"},
	{nombre: "otra-serie", valor: "BOE-B-2015-10565"},
	{nombre: "espacio-inicial", valor: " BOE-A-2015-10565"},
	{nombre: "salto-de-línea-final", valor: "BOE-A-2015-10565\n"},
	{nombre: "cifras-de-otra-escritura", valor: "BOE-A-２０１５-10565"},
}

// casosDeBloque son los casos límite de BLOQUE, ^[A-Za-z0-9][A-Za-z0-9.-]{0,63}$.
var casosDeBloque = []casoDeGramatica{
	{nombre: "artículo", valor: "a21", valido: true},
	{nombre: "punto-final", valor: "a85bis.", valido: true},
	{nombre: "guion", valor: "a1-30", valido: true},
	{nombre: "preámbulo", valor: "preambulo", valido: true},
	{nombre: "mayúscula-inicial", valor: "A21", valido: true},
	{nombre: "una-cifra", valor: "1", valido: true},
	{nombre: "sesenta-y-cuatro-caracteres", valor: "a" + strings.Repeat("1", 63), valido: true},
	{nombre: "sesenta-y-cinco-caracteres", valor: "a" + strings.Repeat("1", 64)},
	{nombre: "vacío", valor: ""},
	{nombre: "punto-inicial", valor: ".a1"},
	{nombre: "guion-inicial", valor: "-a1"},
	{nombre: "dos-puntos", valor: ".."},
	{nombre: "barra", valor: "a/1"},
	{nombre: "espacio", valor: "a 1"},
	{nombre: "porcentaje", valor: "a%201"},
	{nombre: "salto-de-línea-final", valor: "a21\n"},
	{nombre: "letra-no-ascii", valor: "á21"},
	{nombre: "guion-bajo", valor: "a_1"},
}

// ubicacionDeGramatica es un sitio del formato de eval donde va un valor de la
// gramática: el documento de una eval positiva con el valor ahí y valores
// válidos en todo lo demás.
type ubicacionDeGramatica struct {
	nombre    string
	documento func(valor string) map[string]any
}

// evalPositiva es una eval positiva con un único comando y una única cita.
func evalPositiva(comando, cita map[string]any) map[string]any {
	return map[string]any{
		"pregunta": "¿qué dice el art. 21 de la Ley 39/2015?",
		"activa":   true,
		"comandos": []any{comando},
		"citas":    []any{cita},
	}
}

// TestGramaticasCoincidenConBoe comprueba que los patrones de NORMA y BLOQUE
// escritos en los esquemas aceptan y rechazan exactamente lo mismo que
// boe.ValidarNorma y boe.ValidarBloque sobre los casos límite de data-model
// (cabecera; control 10 del plan). En eval.yaml.json lo comprueba en cada sitio
// del formato donde va el valor —el comando de bloque, el de consulta de norma,
// desde H7.1 el de comprobación, la cita, desde H7.4, la redacción modificada y,
// desde H20, la redacción fechada y la norma modificadora de su posterior—, con
// una eval que solo puede fallar por ese valor: la misma eval con
// un valor válido se lee sin error. En normas.yaml.json lo comprueba en el único
// sitio donde va una norma, el nombre de cada entrada de normas, con una tabla
// que solo puede fallar por ese nombre, y exige que cada rechazo sea el defecto
// «identificador con otra forma» de esa entrada: el del patrón de
// propertyNames, y no otro.
func TestGramaticasCoincidenConBoe(t *testing.T) {
	t.Parallel()

	t.Run("eval.yaml.json", func(t *testing.T) {
		t.Parallel()

		const normaValida, bloqueValido = "BOE-A-2015-10565", "a21"

		comandoDeBloque := func(norma, bloque string) map[string]any {
			return map[string]any{"applet": "boe", "norma": norma, "bloque": bloque}
		}
		cita := func(norma, bloque string) map[string]any {
			return map[string]any{"norma": norma, "bloque": bloque}
		}
		conRedaccionModificada := func(norma, bloque string) map[string]any {
			documento := evalPositiva(comandoDeBloque(normaValida, bloqueValido), cita(normaValida, bloqueValido))
			documento["redacciones_modificadas"] = []any{map[string]any{
				"norma": norma, "bloque": bloque, "fecha_vigencia": "20180309", "fecha_vigencia_reciente": "20200206",
			}}

			return documento
		}
		conRedaccionFechada := func(norma, bloque, modificadora string) map[string]any {
			documento := evalPositiva(comandoDeBloque(normaValida, bloqueValido), cita(normaValida, bloqueValido))
			documento["redaccion_fechada"] = map[string]any{
				"norma": norma, "bloque": bloque, "fecha": "2019-06-01", "fecha_vigencia": "20180309",
				"posterior": map[string]any{"fecha_vigencia": "20200206", "norma_modificadora": modificadora},
			}

			return documento
		}

		gramaticas := []struct {
			nombre      string
			casos       []casoDeGramatica
			valida      func(string) error
			valorValido string
			ubicaciones []ubicacionDeGramatica
		}{
			{
				nombre:      "norma",
				casos:       casosDeNorma,
				valida:      boe.ValidarNorma,
				valorValido: normaValida,
				ubicaciones: []ubicacionDeGramatica{
					{nombre: "comando de bloque", documento: func(valor string) map[string]any {
						return evalPositiva(comandoDeBloque(valor, bloqueValido), cita(normaValida, bloqueValido))
					}},
					{nombre: "comando de consulta de norma", documento: func(valor string) map[string]any {
						return evalPositiva(map[string]any{"applet": "boe", "verbo": "metadatos", "norma": valor},
							cita(normaValida, bloqueValido))
					}},
					{nombre: "comando de comprobación", documento: func(valor string) map[string]any {
						return evalPositiva(map[string]any{"applet": "graph", "verbo": "check", "norma": valor},
							cita(normaValida, bloqueValido))
					}},
					{nombre: "cita", documento: func(valor string) map[string]any {
						return evalPositiva(comandoDeBloque(normaValida, bloqueValido), cita(valor, bloqueValido))
					}},
					{nombre: "redacción modificada", documento: func(valor string) map[string]any {
						return conRedaccionModificada(valor, bloqueValido)
					}},
					{nombre: "redacción fechada", documento: func(valor string) map[string]any {
						return conRedaccionFechada(valor, bloqueValido, normaValida)
					}},
					{nombre: "norma modificadora de la redacción posterior", documento: func(valor string) map[string]any {
						return conRedaccionFechada(normaValida, bloqueValido, valor)
					}},
				},
			},
			{
				nombre:      "bloque",
				casos:       casosDeBloque,
				valida:      boe.ValidarBloque,
				valorValido: bloqueValido,
				ubicaciones: []ubicacionDeGramatica{
					{nombre: "comando de bloque", documento: func(valor string) map[string]any {
						return evalPositiva(comandoDeBloque(normaValida, valor), cita(normaValida, bloqueValido))
					}},
					{nombre: "cita", documento: func(valor string) map[string]any {
						return evalPositiva(comandoDeBloque(normaValida, bloqueValido), cita(normaValida, valor))
					}},
					{nombre: "redacción modificada", documento: func(valor string) map[string]any {
						return conRedaccionModificada(normaValida, valor)
					}},
					{nombre: "redacción fechada", documento: func(valor string) map[string]any {
						return conRedaccionFechada(normaValida, valor, normaValida)
					}},
				},
			},
		}

		for _, gramatica := range gramaticas {
			t.Run(gramatica.nombre, func(t *testing.T) {
				t.Parallel()

				for _, ubicacion := range gramatica.ubicaciones {
					require.True(t, aceptaLaEval(t, ubicacion.documento(gramatica.valorValido)),
						"con %q en %s la eval es válida: sin eso, un rechazo no diría nada del patrón",
						gramatica.valorValido, ubicacion.nombre)
				}

				for _, caso := range gramatica.casos {
					t.Run(caso.nombre, func(t *testing.T) {
						t.Parallel()

						aceptaBoe := gramatica.valida(caso.valor) == nil
						require.Equal(t, caso.valido, aceptaBoe,
							"la tabla dice de %q lo mismo que boe", caso.valor)

						for _, ubicacion := range gramatica.ubicaciones {
							assert.Equal(t, aceptaBoe, aceptaLaEval(t, ubicacion.documento(caso.valor)),
								"los patrones de %s del esquema en %s aceptan %q si y solo si boe lo acepta",
								gramatica.nombre, ubicacion.nombre, caso.valor)
						}
					})
				}
			})
		}
	})

	t.Run("normas.yaml.json", func(t *testing.T) {
		t.Parallel()

		const normaValida = "BOE-A-2015-10565"

		require.NoError(t, leerNormasConIdentificador(t, normaValida),
			"con %q la tabla de normas es válida: sin eso, un rechazo no diría nada del patrón", normaValida)

		for _, caso := range casosDeNorma {
			t.Run(caso.nombre, func(t *testing.T) {
				t.Parallel()

				aceptaBoe := boe.ValidarNorma(caso.valor) == nil
				require.Equal(t, caso.valido, aceptaBoe, "la tabla dice de %q lo mismo que boe", caso.valor)

				err := leerNormasConIdentificador(t, caso.valor)
				if aceptaBoe {
					require.NoError(t, err, "el patrón de propertyNames de normas.yaml.json acepta %q, como boe", caso.valor)

					return
				}

				var defecto *skills.DefectoDeNorma
				require.ErrorAs(t, err, &defecto,
					"el patrón de propertyNames de normas.yaml.json rechaza %q, como boe", caso.valor)
				assert.Equal(t, &skills.DefectoDeNorma{Norma: caso.valor, Defecto: "identificador con otra forma"}, defecto,
					"el rechazo de %q es el del patrón de propertyNames", caso.valor)
			})
		}
	})
}

// leerNormasConIdentificador lee con skills.LeerNormas una tabla de normas,
// escrita en JSON, que es YAML válido y conserva cada texto tal cual, con una
// sola norma cuyo nombre es el identificador y que es válida en todo lo demás, y
// devuelve su error.
func leerNormasConIdentificador(t *testing.T, identificador string) error {
	t.Helper()

	contenido, err := json.Marshal(map[string]any{
		"normas": map[string]any{
			identificador: map[string]any{
				"titulo":   "Ley 39/2015, de 1 de octubre, del Procedimiento Administrativo Común de las Administraciones Públicas.",
				"rango":    "Ley",
				"materias": []any{"procedimiento administrativo"},
			},
		},
	})
	require.NoError(t, err)

	_, err = skills.LeerNormas(contenido)

	return err
}

// aceptaLaEval dice si LeerEval acepta el documento, escrito en JSON, que es
// YAML válido y conserva cada texto tal cual.
func aceptaLaEval(t *testing.T, documento map[string]any) bool {
	t.Helper()

	contenido, err := json.Marshal(documento)
	require.NoError(t, err)

	_, err = LeerEval(nombreDeEval, contenido)

	return err == nil
}
