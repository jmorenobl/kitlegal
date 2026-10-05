package evals

import (
	"encoding/json/jsontext"
	"encoding/json/v2"
	"fmt"
	"io/fs"
	"maps"
	"path/filepath"
	"slices"
	"strings"
	"sync"

	"github.com/santhosh-tekuri/jsonschema/v6"

	"github.com/jmorenobl/kitlegal/internal/skills"
)

// carpetaDelJuez es el nombre exacto de la carpeta del juez en la carpeta de
// evals de una skill: LeerConjunto la reconoce por él y no la lee como eval
// (contracts/juez-y-voto.md §1 de H24; research D20 de H24).
const carpetaDelJuez = "juez"

// Los cinco ficheros de la carpeta del juez (contracts/juez-y-voto.md §1 de
// H24): la declaración de clases, que se escribe en ella, y las cuatro copias de
// lo que validó al juez.
const (
	ficheroDeClasesDelJuez  = "clases.yaml"
	ficheroDeRubricaDelJuez = "rubrica.md"
	ficheroDeEsquemaDelJuez = "esquema.json"
	ficheroDeCasosDelJuez   = "casos.yaml"
	ficheroDeMedidaDelJuez  = "medida.json"
)

// rutaDelEsquemaDeClasesDelJuez es el esquema publicado de la declaración de
// clases, relativo al directorio de este paquete, como el de eval.
const rutaDelEsquemaDeClasesDelJuez = "../../schemas/juez-clases.yaml.json"

// motivoNoDirectorio es el motivo de la entrada juez que no es un directorio.
const motivoNoDirectorio = "no es un directorio"

// Juez es el juez con modelo de una skill: lo que LeerConjunto lee de la carpeta
// juez de su carpeta de evals (data-model §1 y contracts/juez-y-voto.md §1 de
// H24; FR-020). Una skill sin esa carpeta no tiene juez.
type Juez struct {
	// Clases son las de clases.yaml, en su orden: al menos una y sin nombres
	// repetidos.
	Clases []ClaseDelJuez

	// Rubrica es el contenido de rubrica.md, entero: va tal cual como
	// instrucciones del juez.
	Rubrica string

	// Esquema es el contenido de esquema.json, entero: la forma de la respuesta
	// del juez, con la que se valida cada voto. Sus propiedades son exactamente
	// los nombres de Clases.
	Esquema string

	// Casos es la ruta de casos.yaml, los casos etiquetados con los que se mide
	// al juez.
	Casos string

	// Medida es la ruta de medida.json, la medida versionada del juez.
	Medida string
}

// ClaseDelJuez es una clase de la declaración de clases del juez de una skill,
// clases.yaml: lo que el juez mira en cada respuesta (FR-020 y FR-021 de H24).
type ClaseDelJuez struct {
	// Nombre es el de la clase, con la forma ^[a-z0-9_]+$: el de su propiedad en
	// el esquema de la respuesta del juez.
	Nombre string `yaml:"nombre"`

	// Decide dice si la clase decide —su umbral del informe hace fallar el job—
	// o solo se publica.
	Decide bool `yaml:"decide"`

	// Umbral es la proporción de respuestas marcadas en la clase que se admite,
	// de 0 a 1.
	Umbral float64 `yaml:"umbral"`
}

// declaracionDeClases es el documento clases.yaml de la carpeta del juez.
type declaracionDeClases struct {
	Clases []ClaseDelJuez `yaml:"clases"`
}

// esquemaDeClasesDelJuez compila una sola vez el esquema publicado de la
// declaración de clases, que no cambia mientras se ejecutan los tests.
var esquemaDeClasesDelJuez = sync.OnceValues(func() (*jsonschema.Schema, error) {
	return compilarEsquemaPublicado(rutaDelEsquemaDeClasesDelJuez, "de la declaración de clases del juez")
})

// leerJuez lee el juez de una skill de la entrada juez de su directorio de evals
// (contracts/juez-y-voto.md §1 de H24; FR-020). Si la entrada no es un
// directorio, es ella el fichero mal formado. Si lo es, lee sus cinco ficheros
// en orden de nombre y devuelve un fichero mal formado por cada uno que falta,
// que no se puede leer o que no tiene su forma, con su nombre dentro del
// directorio de evals, juez/<fichero>, que es por el que empieza su error:
//
//   - clases.yaml se lee con leerClasesDelJuez;
//   - esquema.json, con propiedadesDelEsquema, y, si la declaración se ha leído,
//     sus propiedades tienen que ser exactamente las clases declaradas;
//   - de rubrica.md, de casos.yaml y de medida.json solo se exige que se puedan
//     leer.
//
// El juez se devuelve solo si ninguno está mal formado: con la rúbrica y el
// esquema enteros y con la ruta de los casos y la de la medida.
func leerJuez(dir string, entrada fs.DirEntry) (*Juez, []FicheroMalFormado) {
	if !entrada.IsDir() {
		return nil, []FicheroMalFormado{{
			Fichero: carpetaDelJuez,
			Error:   fmt.Errorf("%s: %s", carpetaDelJuez, motivoNoDirectorio),
		}}
	}

	lectura := lecturaDelJuez{carpeta: filepath.Join(dir, carpetaDelJuez)}

	lectura.contenido(ficheroDeCasosDelJuez)
	clases := lectura.clases()
	esquema := lectura.esquema(clases)
	lectura.contenido(ficheroDeMedidaDelJuez)
	rubrica, _ := lectura.contenido(ficheroDeRubricaDelJuez)

	if len(lectura.malFormados) > 0 {
		return nil, lectura.malFormados
	}

	return &Juez{
		Clases:  clases,
		Rubrica: string(rubrica),
		Esquema: string(esquema),
		Casos:   filepath.Join(lectura.carpeta, ficheroDeCasosDelJuez),
		Medida:  filepath.Join(lectura.carpeta, ficheroDeMedidaDelJuez),
	}, nil
}

// lecturaDelJuez es la lectura de la carpeta del juez de una skill: su ruta y
// los ficheros mal formados que va encontrando, en el orden en que se leen.
type lecturaDelJuez struct {
	carpeta     string
	malFormados []FicheroMalFormado
}

// malFormado anota el fichero de la carpeta como mal formado por ese motivo. Su
// nombre es el que tiene dentro del directorio de evals, juez/<fichero>, y su
// error empieza por él, como el de una eval.
func (l *lecturaDelJuez) malFormado(fichero string, motivo error) {
	nombre := carpetaDelJuez + "/" + fichero

	l.malFormados = append(l.malFormados, FicheroMalFormado{Fichero: nombre, Error: fmt.Errorf("%s: %w", nombre, motivo)})
}

// contenido lee entero el fichero de la carpeta con ese nombre. Si falta o no se
// puede leer, lo anota como mal formado y leido es falso.
func (l *lecturaDelJuez) contenido(fichero string) (contenido []byte, leido bool) {
	contenido, err := leerFichero(filepath.Join(l.carpeta, fichero))
	if err != nil {
		l.malFormado(fichero, fmt.Errorf("no se puede leer: %w", err))

		return nil, false
	}

	return contenido, true
}

// clases lee la declaración de clases de la carpeta, clases.yaml, con
// leerClasesDelJuez. Si no se puede leer o está mal formada, la anota y no
// devuelve ninguna clase.
func (l *lecturaDelJuez) clases() []ClaseDelJuez {
	contenido, leido := l.contenido(ficheroDeClasesDelJuez)
	if !leido {
		return nil
	}

	clases, err := leerClasesDelJuez(contenido)
	if err != nil {
		l.malFormado(ficheroDeClasesDelJuez, err)

		return nil
	}

	return clases
}

// esquema lee el esquema de la respuesta del juez de la carpeta, esquema.json, y
// devuelve su contenido. Lo anota como mal formado si no se puede leer, si no es
// un documento JSON del que leer las propiedades o si sus propiedades no son
// exactamente las clases declaradas. Sin clases, que es que la declaración no se
// ha podido leer, no hay con qué compararlas: el fichero mal formado es la
// declaración, que ya está anotada.
func (l *lecturaDelJuez) esquema(clases []ClaseDelJuez) []byte {
	contenido, leido := l.contenido(ficheroDeEsquemaDelJuez)
	if !leido {
		return nil
	}

	propiedades, err := propiedadesDelEsquema(contenido)
	if err != nil {
		l.malFormado(ficheroDeEsquemaDelJuez, err)

		return nil
	}

	if len(clases) == 0 {
		return contenido
	}

	declaradas := make([]string, 0, len(clases))
	for _, clase := range clases {
		declaradas = append(declaradas, clase.Nombre)
	}

	slices.Sort(declaradas)

	if !slices.Equal(propiedades, declaradas) {
		l.malFormado(ficheroDeEsquemaDelJuez, fmt.Errorf("sus propiedades (%s) no son exactamente las clases declaradas (%s)",
			strings.Join(propiedades, ", "), strings.Join(declaradas, ", ")))

		return nil
	}

	return contenido
}

// leerClasesDelJuez lee el contenido de clases.yaml con el lector común de
// documentos YAML de internal/skills —una clave repetida es un defecto con sus
// dos líneas, nunca la última que gana— y lo valida contra
// juez-clases.yaml.json, como la lista de expresiones prohibidas. Devuelve las
// clases en el orden del fichero; un nombre que se repite es un error que lo
// dice, porque el esquema no puede exigir que sean distintos.
func leerClasesDelJuez(contenido []byte) ([]ClaseDelJuez, error) {
	esquema, err := esquemaDeClasesDelJuez()
	if err != nil {
		return nil, err
	}

	leida, err := skills.ValidarDocumentoYAML[declaracionDeClases](contenido, esquema)
	if err != nil {
		return nil, err
	}

	for posicion, clase := range leida.Clases {
		repetida := slices.ContainsFunc(leida.Clases[:posicion], func(anterior ClaseDelJuez) bool {
			return anterior.Nombre == clase.Nombre
		})
		if repetida {
			return nil, fmt.Errorf("la clase %s está repetida", clase.Nombre)
		}
	}

	return leida.Clases, nil
}

// propiedadesDelEsquema devuelve, en orden, los nombres de las propiedades del
// esquema de la respuesta del juez: las claves de properties de su raíz, ninguna
// si no la tiene. El error es el de un contenido que no es un objeto JSON con
// esa forma.
func propiedadesDelEsquema(contenido []byte) ([]string, error) {
	var esquema struct {
		Propiedades map[string]jsontext.Value `json:"properties"`
	}

	if err := json.Unmarshal(contenido, &esquema); err != nil {
		return nil, fmt.Errorf("no se puede leer como JSON: %w", err)
	}

	return slices.Sorted(maps.Keys(esquema.Propiedades)), nil
}
