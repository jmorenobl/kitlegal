package evals

import (
	"encoding/json"
	"os"
	"path/filepath"
	"slices"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.yaml.in/yaml/v3"

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
		"  grabaciones: lpac-a21-version-anterior\n" +
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
func TestLeerEval(t *testing.T) {
	t.Parallel()

	citaDeLaLRBRL := []CitaEsperada{{Norma: "BOE-A-1985-5392", Bloque: "a85bis."}}

	// La eval de territorio del municipio cubierto y lo que se lee de su comando y
	// de su esperado.
	positivaDelMunicipio := preguntaDelMunicipio + "activa: true\n" + comandoDelMunicipio
	preguntaDelMunicipioLeida := "¿En qué boletines se publican las normas que afectan a Leganés?"
	comandoDelMunicipioLeido := []ComandoEsperado{{Applet: "territorio", Verbo: "resolver", Municipio: "Leganés"}}

	// La positiva del art. 21, sin avisos, y lo que se lee de su comando y de su
	// cita: los casos de avisos le añaden líneas al final.
	positivaDelArticulo21 := preguntaDelArticulo21 + "activa: true\n" + comandoDelArticulo21 + citaDelArticulo21
	comandoDelArticulo21Leido := []ComandoEsperado{{Applet: "boe", Norma: "BOE-A-2015-10565", Bloque: "a21"}}
	citaDelArticulo21Leida := []CitaEsperada{{Norma: "BOE-A-2015-10565", Bloque: "a21"}}

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
			// Sin citas ni territorio fallan las dos ramas del anyOf del esperado
			// verificable, y el defecto es una hoja de cada una.
			nombre:    "positiva-sin-esperado-verificable",
			documento: preguntaDelArticulo21 + "activa: true\n" + comandoDelArticulo21,
			error: nombreDeEval + ": línea 1: missing property 'citas'\n" +
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
				Fichero:  nombreDeEval,
				Pregunta: preguntaDelMunicipioLeida,
				Activa:   true,
				Comandos: comandoDelMunicipioLeido,
				Territorio: TerritorioEsperado{
					Comunidad: "Comunidad de Madrid",
					Provincia: "Madrid",
					Boletines: []string{"BOCM"},
					Cobertura: []string{"boletin_autonomico: configurado", "dir3: verificado"},
				},
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
				GrafoPrevio: GrafoPrevio{Grabaciones: "lpac-a21-version-anterior", Comandos: comandoDelArticulo21Leido},
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
				strings.Replace(grafoPrevioDelArticulo21, "lpac-a21-version-anterior", "../lpac-a21", 1),
			error: nombreDeEval + ": grafo_previo/grabaciones, línea 11: " +
				"'../lpac-a21' does not match pattern '^[a-z0-9]+(-[a-z0-9]+)*$'",
		},
		{
			nombre:    "grafo-previo-sin-comandos",
			documento: positivaDelArticulo21 + "grafo_previo:\n  grabaciones: lpac-a21-version-anterior\n",
			error:     nombreDeEval + ": grafo_previo, línea 11: missing property 'comandos'",
		},
		{
			nombre:    "grafo-previo-con-comandos-vacio",
			documento: positivaDelArticulo21 + "grafo_previo:\n  grabaciones: lpac-a21-version-anterior\n  comandos: []\n",
			error:     nombreDeEval + ": grafo_previo/comandos, línea 12: minItems: got 0, want 1",
		},
		{
			nombre: "grafo-previo-con-comando-de-comprobacion",
			documento: positivaDelArticulo21 + "grafo_previo:\n  grabaciones: lpac-a21-version-anterior\n  comandos:\n" +
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
				GrafoPrevio: GrafoPrevio{Grabaciones: "lpac-a21-version-anterior", Comandos: comandoDelArticulo21Leido},
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
// norma que consulta (contrato evals-y-skill §1 de H7.1).
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

			for posicion, comando := range eval.Comandos {
				assert.Equal(t, comando.Verbo == "check", formaDelComando(comando) == formaComprobacion,
					"el comando %d de %s tiene la forma de comprobación si y solo si su verbo es check", posicion, ruta)
			}

			exigirComprobacionesSinNormaLeidasSinElla(t, ruta, claves, eval.Comandos)
		}
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
// desde H7.1 el de comprobación, y la cita—, con una eval que solo puede fallar
// por ese valor: la misma eval con
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
