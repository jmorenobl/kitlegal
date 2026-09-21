package skills

import (
	"errors"
	"fmt"
	"maps"
	"slices"
	"sync"

	"github.com/santhosh-tekuri/jsonschema/v6"

	"github.com/jmorenobl/kitlegal/internal/core/territorio"
)

// Los esquemas publicados de los ficheros congelados de data/territorio/,
// relativos al directorio de un paquete de internal/, que es donde go test
// ejecuta los tests desde los que se leen esos ficheros (research.md V37 y V42).
const (
	rutaDelEsquemaDeMunicipios = "../../schemas/territorio-municipios.yaml.json"
	rutaDelEsquemaDeDIR3       = "../../schemas/territorio-dir3.yaml.json"
	rutaDelEsquemaDelEstado    = "../../schemas/territorio-estado.yaml.json"
	rutaDelEsquemaDeComunidad  = "../../schemas/territorio-comunidad.yaml.json"
)

// Dónde vive cada fichero congelado, relativo a la raíz del repositorio: el
// nombre con el que LeerTerritorio señala cada defecto (contrato de datos §1).
const (
	ficheroDeMunicipios  = "data/territorio/municipios.yaml"
	ficheroDeDIR3        = "data/territorio/dir3.yaml"
	ficheroDelEstado     = "data/territorio/estado.yaml"
	carpetaDeComunidades = "data/territorio/comunidades/"
	extensionDeComunidad = ".yaml"
)

// esquemasDelTerritorio son los cuatro esquemas de data/territorio/ compilados.
type esquemasDelTerritorio struct {
	municipios, dir3, estado, comunidad *jsonschema.Schema
}

// esquemasDeTerritorio compila una sola vez los esquemas publicados de los
// ficheros de data/territorio/, que no cambian mientras se ejecutan los tests.
var esquemasDeTerritorio = sync.OnceValues(func() (esquemasDelTerritorio, error) {
	var (
		esquemas esquemasDelTerritorio
		defectos []error
	)

	for _, esquema := range []struct {
		ruta    string
		destino **jsonschema.Schema
	}{
		{rutaDelEsquemaDeMunicipios, &esquemas.municipios},
		{rutaDelEsquemaDeDIR3, &esquemas.dir3},
		{rutaDelEsquemaDelEstado, &esquemas.estado},
		{rutaDelEsquemaDeComunidad, &esquemas.comunidad},
	} {
		compilado, err := compilarEsquemaDelTerritorio(esquema.ruta)
		defectos = append(defectos, err)
		*esquema.destino = compilado
	}

	return esquemas, errors.Join(defectos...)
})

// compilarEsquemaDelTerritorio lee el esquema de un fichero de data/territorio/
// de la ruta y lo compila con CompilarEsquema. El error nombra la ruta: la del
// fichero que no se puede leer o la del esquema que no compila.
func compilarEsquemaDelTerritorio(ruta string) (*jsonschema.Schema, error) {
	contenido, err := leerFichero(ruta)
	if err != nil {
		return nil, fmt.Errorf("no se puede leer un esquema de data/territorio/: %w", err)
	}

	esquema, err := CompilarEsquema(contenido)
	if err != nil {
		return nil, fmt.Errorf("el esquema de data/territorio/ %s: %w", ruta, err)
	}

	return esquema, nil
}

// LeerTerritorio lee los cuatro ficheros congelados de data/territorio/ con el
// lector común de documentos YAML, ValidarDocumentoYAML, cada uno contra su
// esquema publicado y en el tipo con el que lo decodifica el dominio, sin
// repetirlo (contrato de datos §2; FR-044, research.md D27): la relación de
// municipios, la correspondencia INE→DIR3, el estado y cada comunidad. El
// lector rechaza además toda clave repetida, en cualquier mapa del documento.
//
// Solo valida cada fichero por separado: la integridad entre ficheros la
// comprueba territorio.Cargar. Ningún defecto se descarta: van todos, cada uno
// con la ruta de su fichero delante, en el orden de los ficheros —las
// comunidades, por código— y unidos con errors.Join, y con ellos no se devuelve
// nada.
func LeerTerritorio(fuentes territorio.Fuentes) (territorio.Ficheros, error) {
	esquemas, err := esquemasDeTerritorio()
	if err != nil {
		return territorio.Ficheros{}, err
	}

	municipios, errDeMunicipios := leerFicheroDelTerritorio[territorio.FicheroDeMunicipios](
		ficheroDeMunicipios, fuentes.Municipios, esquemas.municipios)
	dir3, errDeDIR3 := leerFicheroDelTerritorio[territorio.FicheroDeDIR3](ficheroDeDIR3, fuentes.DIR3, esquemas.dir3)
	estado, errDelEstado := leerFicheroDelTerritorio[territorio.FicheroDeEstado](
		ficheroDelEstado, fuentes.Estado, esquemas.estado)

	defectos := []error{errDeMunicipios, errDeDIR3, errDelEstado}
	comunidades := make(map[string]territorio.FicheroDeComunidad, len(fuentes.Comunidades))

	for _, codigo := range slices.Sorted(maps.Keys(fuentes.Comunidades)) {
		comunidad, err := leerFicheroDelTerritorio[territorio.FicheroDeComunidad](
			carpetaDeComunidades+codigo+extensionDeComunidad, fuentes.Comunidades[codigo], esquemas.comunidad)
		defectos = append(defectos, err)
		comunidades[codigo] = comunidad
	}

	if err := errors.Join(defectos...); err != nil {
		return territorio.Ficheros{}, err
	}

	return territorio.Ficheros{Municipios: municipios, DIR3: dir3, Estado: estado, Comunidades: comunidades}, nil
}

// leerFicheroDelTerritorio lee un fichero con el lector común contra su
// esquema y pone su ruta delante de cada uno de sus defectos, que siguen siendo
// los del lector: un ClaveRepetida o un DefectoEnElDocumento se reconocen con
// errors.As a través de ella.
func leerFicheroDelTerritorio[T any](fichero string, contenido []byte, esquema *jsonschema.Schema) (T, error) {
	leido, err := ValidarDocumentoYAML[T](contenido, esquema)
	if err == nil {
		return leido, nil
	}

	defectos := []error{err}

	var unidos interface{ Unwrap() []error }
	if errors.As(err, &unidos) {
		defectos = unidos.Unwrap()
	}

	enSuFichero := make([]error, 0, len(defectos))
	for _, defecto := range defectos {
		enSuFichero = append(enSuFichero, fmt.Errorf("%s: %w", fichero, defecto))
	}

	return leido, errors.Join(enSuFichero...)
}
