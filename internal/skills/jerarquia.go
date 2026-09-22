package skills

import (
	"fmt"
	"sync"

	"github.com/santhosh-tekuri/jsonschema/v6"
)

// rutaDelEsquemaDeJerarquia es el esquema publicado de data/jerarquia.yaml,
// relativo al directorio de un paquete de internal/, que es donde go test
// ejecuta los tests desde los que se lee el fichero.
const rutaDelEsquemaDeJerarquia = "../../schemas/jerarquia.yaml.json"

// Jerarquia es data/jerarquia.yaml leído y validado contra su esquema
// (data-model §4.2; FR-066, FR-067): los niveles territoriales y las reglas de
// interpretación, cada uno una vez y en el orden de su enumerado, que es el que
// el esquema exige al documento.
type Jerarquia struct {
	// Niveles son los cinco niveles, de la Unión Europea al municipio.
	Niveles []NivelNormativo `yaml:"niveles"`

	// Reglas son las cuatro reglas de interpretación, de la competencia al
	// reglamento que nunca va contra la ley.
	Reglas []ReglaDeInterpretacion `yaml:"reglas"`
}

// NivelNormativo es un nivel territorial de la jerarquía normativa: quién
// regula, en qué clase de boletín se publica y con qué tipos de norma.
type NivelNormativo struct {
	// Codigo es su valor del enumerado nivel: ue, estado, comunidad-autonoma,
	// provincia o municipio.
	Codigo string `yaml:"nivel"`

	// Nombre es el nombre del nivel, como «Comunidad autónoma».
	Nombre string `yaml:"nombre"`

	// Boletin es la clase de boletín que publica sus normas, no un boletín
	// concreto: el de cada comunidad y provincia lo da territorio resolver.
	Boletin string `yaml:"boletin"`

	// Normas son sus tipos de norma, de mayor a menor rango y sin repetir.
	Normas []string `yaml:"normas"`
}

// ReglaDeInterpretacion es una regla con la que se decide entre dos normas que
// se contradicen.
type ReglaDeInterpretacion struct {
	// Codigo es su valor del enumerado regla: competencia-antes-que-jerarquia,
	// ley-posterior, ley-especial o reglamento-nunca-contra-ley.
	Codigo string `yaml:"regla"`

	// Enunciado dice la regla con palabras.
	Enunciado string `yaml:"enunciado"`
}

// esquemaDeJerarquia compila una sola vez el esquema publicado de
// data/jerarquia.yaml, que no cambia mientras se ejecutan los tests.
var esquemaDeJerarquia = sync.OnceValues(func() (*jsonschema.Schema, error) {
	return compilarEsquemaDeJerarquia(rutaDelEsquemaDeJerarquia)
})

// compilarEsquemaDeJerarquia lee el esquema de data/jerarquia.yaml de la ruta y
// lo compila con CompilarEsquema. El error nombra la ruta: la del fichero que no
// se puede leer o la del esquema que no compila.
func compilarEsquemaDeJerarquia(ruta string) (*jsonschema.Schema, error) {
	contenido, err := leerFichero(ruta)
	if err != nil {
		return nil, fmt.Errorf("no se puede leer el esquema de data/jerarquia.yaml: %w", err)
	}

	esquema, err := CompilarEsquema(contenido)
	if err != nil {
		return nil, fmt.Errorf("el esquema de data/jerarquia.yaml %s: %w", ruta, err)
	}

	return esquema, nil
}

// LeerJerarquia lee el contenido de data/jerarquia.yaml con el lector común de
// documentos YAML, ValidarDocumentoYAML, contra su esquema publicado
// (data-model §4.2; FR-044): el lector rechaza toda clave repetida, en
// cualquier mapa del documento, y el esquema, todo nivel o regla fuera de su
// enumerado o de su sitio, toda clave que no declara y todo campo que falta o
// va vacío. Devuelve los niveles y las reglas en el orden del documento, que el
// esquema obliga a ser el de sus enumerados.
//
// Ningún defecto se descarta: van todos, como los da el lector —con el punto del
// documento y su línea—, en el orden del documento y unidos con errors.Join, y
// con ellos no se devuelve nada.
func LeerJerarquia(contenido []byte) (Jerarquia, error) {
	esquema, err := esquemaDeJerarquia()
	if err != nil {
		return Jerarquia{}, err
	}

	return ValidarDocumentoYAML[Jerarquia](contenido, esquema)
}
