package skills

import (
	"errors"
	"fmt"
	"strconv"
	"strings"
	"sync"
	"unicode"

	"github.com/santhosh-tekuri/jsonschema/v6"
	"github.com/santhosh-tekuri/jsonschema/v6/kind"
	"go.yaml.in/yaml/v3"
)

// rutaDelEsquemaDeNormas es el esquema publicado de data/normas.yaml, relativo
// al directorio de un paquete de internal/, que es donde go test ejecuta los
// tests desde los que se lee la tabla (research.md D2 y V46).
const rutaDelEsquemaDeNormas = "../../schemas/normas.yaml.json"

// Lo que LeerNormas reconoce en la tabla para atribuir cada defecto a su norma.
const (
	// claveDeLasNormas es la única clave de la raíz de la tabla: el mapa de las
	// normas, indexado por identificador (data-model §4.1).
	claveDeLasNormas = "normas"

	// etiquetaDeFusion es la etiqueta de la clave << de YAML, que fusiona en un
	// mapa las claves de otro.
	etiquetaDeFusion = "!!merge"
)

// Norma es una entrada de data/normas.yaml leída y validada contra su esquema
// (data-model §4.1; contrato normas-y-referencias §1 y §2).
type Norma struct {
	// Identificador es la clave de la entrada, BOE-A-<año>-<número>. No es un
	// campo de la entrada: lo pone LeerNormas.
	Identificador string `yaml:"-"`

	// Titulo es el título oficial, tal como lo da la búsqueda grabada.
	Titulo string `yaml:"titulo"`

	// Rango es uno de los rangos que da la fuente (enum de rango del esquema).
	Rango string `yaml:"rango"`

	// Abreviatura es el nombre corto con que se la nombra, o vacío si no lo
	// tiene.
	Abreviatura string `yaml:"abreviatura"`

	// Materias son las materias que sirven para elegirla, al menos una y sin
	// repetir, en el orden del documento.
	Materias []string `yaml:"materias"`
}

// DefectoDeNorma es un defecto de una entrada de data/normas.yaml: nombra la
// norma y dice qué le falla (contrato normas-y-referencias §3; FR-043).
type DefectoDeNorma struct {
	// Norma es la clave de la entrada: su identificador o, si no tiene la forma
	// de uno, la clave tal como la escribe la tabla.
	Norma string

	// Defecto dice qué falla, como «campo no declarado: vertical», «falta
	// titulo», «materias vacía», «identificador con otra forma» o «repetido en
	// las líneas 12 y 20».
	Defecto string
}

// Error presenta la norma y el defecto, como «BOE-A-2015-10565: falta titulo».
// Una norma vacía, o con espacios o caracteres que no se ven, va entre comillas.
func (d *DefectoDeNorma) Error() string {
	return presentarNorma(d.Norma) + ": " + d.Defecto
}

// presentarNorma escribe la clave de una norma tal cual o, si vacía o con
// espacios o caracteres que no se ven no se distinguiría en un mensaje, entre
// comillas y con sus caracteres escapados.
func presentarNorma(norma string) string {
	noSeVe := func(caracter rune) bool { return unicode.IsSpace(caracter) || !unicode.IsGraphic(caracter) }
	if norma == "" || strings.ContainsFunc(norma, noSeVe) {
		return strconv.Quote(norma)
	}

	return norma
}

// esquemaDeNormas compila una sola vez el esquema publicado de data/normas.yaml,
// que no cambia mientras se ejecutan los tests.
var esquemaDeNormas = sync.OnceValues(func() (*jsonschema.Schema, error) {
	return compilarEsquemaDeNormas(rutaDelEsquemaDeNormas)
})

// compilarEsquemaDeNormas lee el esquema de data/normas.yaml de la ruta y lo
// compila con CompilarEsquema. El error nombra la ruta: la del fichero que no se
// puede leer o la del esquema que no compila.
func compilarEsquemaDeNormas(ruta string) (*jsonschema.Schema, error) {
	contenido, err := leerFichero(ruta)
	if err != nil {
		return nil, fmt.Errorf("no se puede leer el esquema de data/normas.yaml: %w", err)
	}

	esquema, err := CompilarEsquema(contenido)
	if err != nil {
		return nil, fmt.Errorf("el esquema de data/normas.yaml %s: %w", ruta, err)
	}

	return esquema, nil
}

// LeerNormas lee el contenido de data/normas.yaml y devuelve sus normas en el
// orden del documento (contrato normas-y-referencias §3; FR-025, FR-043):
//
//  1. lo lee con el lector común de documentos YAML, ValidarDocumentoYAML, que
//     rechaza toda clave repetida antes de convertir nada y valida el documento
//     contra normas.yaml.json;
//  2. atribuye cada defecto del lector a su norma, con un DefectoDeNorma que la
//     nombra: «repetido en las líneas N y M» para un identificador repetido,
//     «campo no declarado: vertical», «falta titulo», «materias vacía»,
//     «identificador con otra forma» y los demás incumplimientos del esquema
//     con sus propias palabras;
//  3. un defecto que no es de ninguna norma —la tabla sin normas, con una clave
//     de más en la raíz o con normas vacía— es un DefectoEnElDocumento con su
//     línea, y el que no es del documento, como un YAML mal formado, se
//     devuelve tal cual.
//
// Ningún defecto se descarta: van todos, en el orden en que los da el lector y
// unidos con errors.Join, y con ellos no se devuelve ninguna norma.
func LeerNormas(contenido []byte) ([]Norma, error) {
	esquema, err := esquemaDeNormas()
	if err != nil {
		return nil, err
	}

	tabla, err := ValidarDocumentoYAML[tablaDeNormas](contenido, esquema)
	if err != nil {
		return nil, atribuirDefectos(err)
	}

	return tabla.Normas, nil
}

// tablaDeNormas es data/normas.yaml leído, con sus normas en el orden del
// documento.
type tablaDeNormas struct {
	Normas normasEnOrden `yaml:"normas"`
}

// normasEnOrden son las normas de la tabla en el orden en que las escribe el
// documento, que un mapa de Go no conserva.
type normasEnOrden []Norma

// La tabla decodifica sus normas con su propio lector, el que conserva el orden
// y rechaza fusiones y alias repetidos.
var _ yaml.Unmarshaler = (*normasEnOrden)(nil)

// UnmarshalYAML lee el mapa de normas entrada a entrada. ValidarDocumentoYAML
// solo lo decodifica después de validarlo contra el esquema, así que llega un
// mapa con claves de texto y valores con la forma de una norma. Lo que el
// esquema no puede ver, porque valida el documento ya convertido, lo rechaza
// aquí:
//
//   - una clave << de fusión, que trae normas sin escribir su identificador en
//     la tabla;
//   - un identificador que ya había aparecido y que llega por un alias: el
//     recorrido del lector común compara las claves como las escribe el
//     documento, y al convertirlo la segunda entrada habría sustituido a la
//     primera sin decir nada.
func (n *normasEnOrden) UnmarshalYAML(nodo *yaml.Node) error {
	lineas := make(map[string]int, len(nodo.Content)/2)
	normas := make(normasEnOrden, 0, len(nodo.Content)/2)

	for indice := 0; indice+1 < len(nodo.Content); indice += 2 {
		clave := nodo.Content[indice]
		if clave.ShortTag() == etiquetaDeFusion {
			return &DefectoEnElDocumento{
				Ruta:   []string{claveDeLasNormas},
				Linea:  clave.Line,
				Motivo: "clave de fusión << no admitida: cada norma se escribe con su identificador",
			}
		}

		// El lector común ya convirtió la tabla a un mapa con claves de texto, así
		// que cada clave es un escalar de texto, o un alias de uno, y el
		// identificador es su texto, como en las claves del frontmatter.
		norma := Norma{Identificador: sinEnvoltorio(clave).Value}

		if primera, repetido := lineas[norma.Identificador]; repetido {
			return &DefectoDeNorma{Norma: norma.Identificador, Defecto: repetidoEnLasLineas(primera, clave.Line)}
		}

		lineas[norma.Identificador] = clave.Line

		if err := nodo.Content[indice+1].Decode(&norma); err != nil {
			return fmt.Errorf("%s: la norma no se puede leer: %w", presentarNorma(norma.Identificador), err)
		}

		normas = append(normas, norma)
	}

	*n = normas

	return nil
}

// atribuirDefectos presenta cada defecto que da el lector común al leer la
// tabla con atribuir, en el mismo orden, y los devuelve unidos.
func atribuirDefectos(err error) error {
	var unidos interface{ Unwrap() []error }

	defectos := []error{err}
	if errors.As(err, &unidos) {
		defectos = unidos.Unwrap()
	}

	var atribuidos []error
	for _, defecto := range defectos {
		atribuidos = append(atribuidos, atribuir(defecto)...)
	}

	return errors.Join(atribuidos...)
}

// atribuir presenta un defecto del lector común en los términos de la tabla: el
// de una norma, como DefectoDeNorma; el del documento, como
// DefectoEnElDocumento con su línea y sin ruta, que ya va en el motivo; y
// cualquier otro error, tal cual. Los defectos de UnmarshalYAML llegan
// envueltos por el lector y se presentan igual que los suyos.
func atribuir(defecto error) []error {
	var (
		deNorma       *DefectoDeNorma
		repetida      *ClaveRepetida
		enElDocumento *DefectoEnElDocumento
	)

	switch {
	case errors.As(defecto, &deNorma):
		return []error{deNorma}
	case errors.As(defecto, &repetida):
		return []error{atribuirRepeticion(repetida)}
	case errors.As(defecto, &enElDocumento):
		return atribuirDefectoDelDocumento(enElDocumento)
	default:
		return []error{defecto}
	}
}

// atribuirRepeticion es la clave repetida en términos de la tabla: dentro de
// normas, un DefectoDeNorma de la norma repetida, o del campo repetido dentro de
// una norma; fuera, la propia ClaveRepetida, que ya nombra la clave y sus líneas.
func atribuirRepeticion(repetida *ClaveRepetida) error {
	norma, resto, deUnaNorma := normaDeLaRuta(rutaHija(repetida.Mapa, repetida.Clave))
	if !deUnaNorma {
		return repetida
	}

	defecto := repetidoEnLasLineas(repetida.Lineas[0], repetida.Lineas[1])
	if len(resto) > 0 {
		defecto = presentarRuta(resto) + " " + defecto
	}

	return &DefectoDeNorma{Norma: norma, Defecto: defecto}
}

// atribuirDefectoDelDocumento presenta un DefectoEnElDocumento con las palabras
// de textosDelDefecto: de la norma de su ruta o, si es de propertyNames, de la
// clave que nombra —la ruta de ese incumplimiento no es de fiar
// (defectoDelEsquema)—; y, si no es de ninguna norma, del documento, con su
// línea.
func atribuirDefectoDelDocumento(defecto *DefectoEnElDocumento) []error {
	norma, resto, deUnaNorma := normaDeLaRuta(defecto.Ruta)
	if nombre, deNombreDeClave := defecto.Tipo.(*kind.PropertyNames); deNombreDeClave {
		norma, resto, deUnaNorma = nombre.Property, nil, true
	}

	textos := textosDelDefecto(defecto, resto)
	atribuidos := make([]error, 0, len(textos))

	for _, texto := range textos {
		if deUnaNorma {
			atribuidos = append(atribuidos, &DefectoDeNorma{Norma: norma, Defecto: texto})

			continue
		}

		atribuidos = append(atribuidos, &DefectoEnElDocumento{Linea: defecto.Linea, Motivo: texto, Tipo: defecto.Tipo})
	}

	return atribuidos
}

// normaDeLaRuta separa, de una ruta dentro de la tabla, la norma y lo que queda
// de ruta dentro de ella. Una ruta que no pasa por una entrada de normas no es
// de ninguna norma y queda entera como resto.
func normaDeLaRuta(ruta []string) (norma string, resto []string, deUnaNorma bool) {
	if len(ruta) < 2 || ruta[0] != claveDeLasNormas {
		return "", ruta, false
	}

	return ruta[1], ruta[2:], true
}

// textosDelDefecto dice con palabras de la tabla qué falla en un defecto, uno
// por campo cuando el incumplimiento nombra varios. El campo es el resto de la
// ruta, dentro de la norma o desde la raíz. Un incumplimiento de otro tipo va
// con el texto del lector, detrás de su campo.
func textosDelDefecto(defecto *DefectoEnElDocumento, resto []string) []string {
	campo := presentarRuta(resto)

	switch tipo := defecto.Tipo.(type) {
	case *kind.PropertyNames:
		return []string{"identificador con otra forma"}
	case *kind.AdditionalProperties:
		return textosPorCampo("campo no declarado: ", resto, tipo.Properties)
	case *kind.Required:
		return textosPorCampo("falta ", resto, tipo.Missing)
	case *kind.MinItems, *kind.MinProperties:
		return []string{campo + " vacía"}
	case *kind.MinLength:
		return []string{campo + " sin texto"}
	case *kind.UniqueItems:
		return []string{fmt.Sprintf("%[1]s/%[2]d y %[1]s/%[3]d repetidas", campo, tipo.Duplicates[0], tipo.Duplicates[1])}
	case *kind.Enum:
		return []string{fmt.Sprintf("%s no admitido: %v", campo, tipo.Got)}
	default:
		if campo == "" {
			return []string{defecto.Motivo}
		}

		return []string{campo + ": " + defecto.Motivo}
	}
}

// textosPorCampo es un texto por campo, con el prefijo delante de la ruta de
// cada campo dentro del resto.
func textosPorCampo(prefijo string, resto, campos []string) []string {
	textos := make([]string, 0, len(campos))
	for _, campo := range campos {
		textos = append(textos, prefijo+presentarRuta(rutaHija(resto, campo)))
	}

	return textos
}

// repetidoEnLasLineas es el defecto de una clave repetida, con la línea de su
// primera aparición y la de la repetición, como lo presenta ClaveRepetida.
func repetidoEnLasLineas(primera, segunda int) string {
	return fmt.Sprintf("repetido en las líneas %d y %d", primera, segunda)
}
