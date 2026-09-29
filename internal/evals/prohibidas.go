package evals

import (
	"fmt"
	"regexp"
	"slices"
	"strings"
	"sync"

	"github.com/santhosh-tekuri/jsonschema/v6"

	"github.com/jmorenobl/kitlegal/internal/skills"
)

// ficheroDeExpresionesProhibidas es el nombre exacto de la lista de expresiones
// prohibidas en la carpeta de evals de una skill: LeerConjunto la reconoce por él
// y no la lee como eval (contrato lista-y-juicio §1; research D1).
const ficheroDeExpresionesProhibidas = "expresiones-prohibidas.yaml"

// rutaDelEsquemaDeExpresionesProhibidas es el esquema publicado de la lista,
// relativo al directorio de este paquete, como el de eval.
const rutaDelEsquemaDeExpresionesProhibidas = "../../schemas/expresiones-prohibidas.yaml.json"

// Los extremos de la expresión regular de una expresión prohibida: delante y
// detrás de sus palabras, el principio o el final del texto o algo que no es
// letra ni cifra (research D3).
const (
	antesDeLaExpresion   = `(?:^|[^\p{L}\p{N}])`
	despuesDeLaExpresion = `(?:$|[^\p{L}\p{N}])`
)

// ExpresionesProhibidas es la lista de expresiones prohibidas de una skill, las
// que no lleva la respuesta de una eval que la activa, en sus tres familias
// (FR-050 de H7.2 y FR-020 de H7.3; data-model §5). Se lee de
// evals/<skill>/expresiones-prohibidas.yaml, validada contra
// schemas/expresiones-prohibidas.yaml.json, que exige las tres; su valor cero es
// el de una skill sin lista.
type ExpresionesProhibidas struct {
	// Maquinaria son las de la maquinaria interna —la memoria de consultas,
	// kitlegal graph y sus verbos, los códigos de salida, los hallazgos y sus
	// clases, el JSON y el sobre—, en el orden del fichero.
	Maquinaria []string `yaml:"maquinaria"`

	// OtraConversacion son las que atribuyen a la skill algo dicho a quien
	// pregunta en otra conversación, en el orden del fichero.
	OtraConversacion []string `yaml:"otra_conversacion"`

	// Anuncio son las que anuncian a quien pregunta la respuesta que viene o el
	// estado de lo comprobado —que no hay nada que trasladar, que ya se puede
	// responder, que se tiene lo necesario—, en el orden del fichero. Ninguna
	// dice que la norma no está derogada o que no tiene avisos (FR-020).
	Anuncio []string `yaml:"anuncio"`
}

// esquemaDeExpresionesProhibidas compila una sola vez el esquema publicado de la
// lista, que no cambia mientras se ejecutan los tests.
var esquemaDeExpresionesProhibidas = sync.OnceValues(func() (*jsonschema.Schema, error) {
	return compilarEsquemaPublicado(rutaDelEsquemaDeExpresionesProhibidas, "de la lista de expresiones prohibidas")
})

// leerExpresionesProhibidas lee el contenido de expresiones-prohibidas.yaml con
// el lector común de documentos YAML de internal/skills —una clave repetida es un
// defecto con sus dos líneas, nunca la última que gana— y lo valida contra
// expresiones-prohibidas.yaml.json. Todo error empieza por el nombre del fichero
// y va con la lista vacía (FR-055).
func leerExpresionesProhibidas(contenido []byte) (ExpresionesProhibidas, error) {
	esquema, err := esquemaDeExpresionesProhibidas()
	if err != nil {
		return ExpresionesProhibidas{}, fmt.Errorf("%s: %w", ficheroDeExpresionesProhibidas, err)
	}

	lista, err := skills.ValidarDocumentoYAML[ExpresionesProhibidas](contenido, esquema)
	if err != nil {
		return ExpresionesProhibidas{}, fmt.Errorf("%s: %w", ficheroDeExpresionesProhibidas, err)
	}

	return lista, nil
}

// ExtraerExpresionesProhibidas devuelve las expresiones de la lista que lleva el
// texto, en el orden de la lista —la maquinaria, lo dicho en otra conversación y
// el anuncio— y sin repetir, o nil si no lleva ninguna. Una expresión se
// encuentra si su forma casa en algún punto del texto: sus palabras en su orden,
// cada una sin distinguir mayúsculas, con los blancos y el énfasis de Markdown
// entre dos que tolera la forma fija de los avisos (H5.1), y sin letra ni cifra a
// los lados. No pliega tildes ni admite un salto de línea entre dos palabras
// (FR-051 de H7.2 y FR-024 de H7.3; contratos lista-y-juicio §3 de H7.2 y
// lista-de-expresiones §3 de H7.3; research D3).
func ExtraerExpresionesProhibidas(texto string, lista ExpresionesProhibidas) []string {
	var encontradas []string

	for _, expresion := range slices.Concat(lista.Maquinaria, lista.OtraConversacion, lista.Anuncio) {
		if !slices.Contains(encontradas, expresion) && formasDeExpresiones.forma(expresion).MatchString(texto) {
			encontradas = append(encontradas, expresion)
		}
	}

	return encontradas
}

// formasDeExpresiones son las expresiones regulares ya compiladas de las
// expresiones prohibidas, por expresión: cada una se compila una sola vez, la
// primera vez que se busca, aunque la busquen a la vez tests en paralelo. Solo
// crecen con las expresiones distintas que se buscan, las de listas fijas.
var formasDeExpresiones = &formasCompiladas{porExpresion: map[string]*regexp.Regexp{}}

// formasCompiladas guarda, con su cerrojo, la expresión regular de cada
// expresión prohibida ya compilada.
type formasCompiladas struct {
	cerrojo      sync.Mutex
	porExpresion map[string]*regexp.Regexp
}

// forma es la expresión regular de la expresión: la de formaDeExpresion, que se
// compila la primera vez que se pide y se guarda para las siguientes.
func (f *formasCompiladas) forma(expresion string) *regexp.Regexp {
	f.cerrojo.Lock()
	defer f.cerrojo.Unlock()

	forma, compilada := f.porExpresion[expresion]
	if !compilada {
		forma = formaDeExpresion(expresion)
		f.porExpresion[expresion] = forma
	}

	return forma
}

// formaDeExpresion compila la expresión regular de una expresión prohibida: las
// palabras de la expresión en su orden, cada una sin distinguir mayúsculas y
// unidas por entrePalabrasDeAviso, entre antesDeLaExpresion y
// despuesDeLaExpresion (contrato lista-y-juicio §3). Las piezas son fijas y cada
// palabra pasa por regexp.QuoteMeta, así que la compilación no puede fallar.
func formaDeExpresion(expresion string) *regexp.Regexp {
	palabras := strings.Fields(expresion)
	for i, palabra := range palabras {
		palabras[i] = `(?i:` + regexp.QuoteMeta(palabra) + `)`
	}

	return regexp.MustCompile(antesDeLaExpresion + strings.Join(palabras, entrePalabrasDeAviso) + despuesDeLaExpresion)
}
