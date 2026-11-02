package skills

import (
	"fmt"
	"sync"

	"github.com/santhosh-tekuri/jsonschema/v6"
)

// rutaDelEsquemaDeCobertura es el esquema publicado de data/cobertura.yaml,
// relativo al directorio de un paquete de internal/, que es donde go test
// ejecuta los tests desde los que se lee el fichero.
const rutaDelEsquemaDeCobertura = "../../schemas/cobertura.yaml.json"

// Cobertura es data/cobertura.yaml leído y validado contra su esquema
// (contracts/datos-de-plazos.md §3 de H9; FR-031; research D17 de H9): los
// procedimientos que la LPAC deja a su normativa específica, en el orden del
// documento. Solo lo lee la skill que tiene las formas ⚠ NO CUBIERTO:, por la
// región cobertura_por_materia de la referencia de cada paquete
// (contracts/paquetes.md §3 de H32); el binario no lo lleva.
type Cobertura struct {
	// ReguladosAparte son los procedimientos, al menos uno.
	ReguladosAparte []ReguladoAparte `yaml:"regulados_aparte"`
}

// ReguladoAparte es un procedimiento que la ley regula aparte del
// procedimiento común: el precepto de la LPAC que lo dice y, si se cita de lo
// grabado el precepto que la hace aplicable, la norma que lo rige.
type ReguladoAparte struct {
	// Procedimiento lo describe, como «reclamaciones económico-administrativas».
	Procedimiento string `yaml:"procedimiento"`

	// Precepto es el de la LPAC que lo regula aparte.
	Precepto PreceptoDeLaCobertura `yaml:"precepto"`

	// Norma es la que lo rige, o nil: una entrada sin norma no la toma de
	// ningún otro sitio, y la declaración nombra el precepto de la LPAC
	// (FR-031).
	Norma *NormaQueLoRige `yaml:"norma"`
}

// PreceptoDeLaCobertura es un apartado de una norma consolidada del BOE: la
// norma por su identificador, el id del bloque en su índice consolidado y el
// apartado que se cita, como «DA 1ª.2.a» o «112.4».
type PreceptoDeLaCobertura struct {
	Norma    string `yaml:"norma"`
	Bloque   string `yaml:"bloque"`
	Apartado string `yaml:"apartado"`
}

// NormaQueLoRige es la norma que rige un procedimiento regulado aparte: su
// identificador, el que da la búsqueda grabada y nunca uno escrito de memoria
// (research S5 de H9); el nombre con que se la nombra y el precepto que la
// hace aplicable.
type NormaQueLoRige struct {
	Identificador string            `yaml:"identificador"`
	Nombre        string            `yaml:"nombre"`
	Aplicable     PreceptoAplicable `yaml:"aplicable"`
}

// PreceptoAplicable es el precepto de la norma que la hace aplicable al
// procedimiento: el id del bloque en su índice consolidado y el apartado.
type PreceptoAplicable struct {
	Bloque   string `yaml:"bloque"`
	Apartado string `yaml:"apartado"`
}

// esquemaDeCobertura compila una sola vez el esquema publicado de
// data/cobertura.yaml, que no cambia mientras se ejecutan los tests.
var esquemaDeCobertura = sync.OnceValues(func() (*jsonschema.Schema, error) {
	return compilarEsquemaDeCobertura(rutaDelEsquemaDeCobertura)
})

// compilarEsquemaDeCobertura lee el esquema de data/cobertura.yaml de la ruta y
// lo compila con CompilarEsquema. El error nombra la ruta: la del fichero que
// no se puede leer o la del esquema que no compila.
func compilarEsquemaDeCobertura(ruta string) (*jsonschema.Schema, error) {
	contenido, err := leerFichero(ruta)
	if err != nil {
		return nil, fmt.Errorf("no se puede leer el esquema de data/cobertura.yaml: %w", err)
	}

	esquema, err := CompilarEsquema(contenido)
	if err != nil {
		return nil, fmt.Errorf("el esquema de data/cobertura.yaml %s: %w", ruta, err)
	}

	return esquema, nil
}

// LeerCobertura lee el contenido de data/cobertura.yaml con el lector común de
// documentos YAML, ValidarDocumentoYAML, contra su esquema publicado: el lector
// rechaza toda clave repetida, en cualquier mapa del documento, y el esquema,
// toda clave que no declara, todo campo que falta o va vacío, todo
// identificador o bloque con otra forma y una lista sin entradas. Devuelve las
// entradas en el orden del documento.
//
// Ningún defecto se descarta: van todos, como los da el lector —con el punto del
// documento y su línea—, en el orden del documento y unidos con errors.Join, y
// con ellos no se devuelve nada.
func LeerCobertura(contenido []byte) (Cobertura, error) {
	esquema, err := esquemaDeCobertura()
	if err != nil {
		return Cobertura{}, err
	}

	return ValidarDocumentoYAML[Cobertura](contenido, esquema)
}
