package evals

import (
	"fmt"
	"sync"

	"github.com/santhosh-tekuri/jsonschema/v6"

	"github.com/jmorenobl/kitlegal/internal/skills"
)

// rutaDelEsquemaDeEval es el esquema publicado del formato común de eval,
// relativo al directorio de este paquete, que es donde go test ejecuta los
// tests desde los que se usa (research.md D2 y V46).
const rutaDelEsquemaDeEval = "../../schemas/eval.yaml.json"

// Eval es un fichero de evals/<skill>/ leído y validado contra el formato común
// (data-model §6; contrato evals-y-grabaciones §1).
type Eval struct {
	// Fichero es el nombre con el que se leyó el fichero. No es una clave del
	// YAML: lo pone LeerEval, y con él se nombra la eval.
	Fichero string `yaml:"-"`

	// Pregunta es lo que se pregunta en la sesión.
	Pregunta string `yaml:"pregunta"`

	// Activa dice si la skill debe activarse con la pregunta.
	Activa bool `yaml:"activa"`

	// Reproduce es el nombre de la skill cuya consulta reproduce la eval, o
	// vacío si no reproduce ninguna (FR-064).
	Reproduce string `yaml:"reproduce"`

	// Informativa dice que la eval se ejecuta y se publica, pero no decide el
	// veredicto: mide algo que la skill todavía no puede hacer con las
	// herramientas que hay (ADR 0016), o algo que todavía no hay datos para
	// exigir (H5.1). Solo la ejecuta el modelo que decide.
	Informativa bool `yaml:"informativa"`

	// Comandos son los comandos esperados de una eval que activa la skill, en el
	// orden del fichero; vacío en una de no activación (FR-061).
	Comandos []ComandoEsperado `yaml:"comandos"`

	// Citas son las citas esperadas de una eval que activa la skill, en el orden
	// del fichero; vacío en una de no activación (FR-061).
	Citas []CitaEsperada `yaml:"citas"`

	// Avisos son los códigos de aviso de vigencia cuya forma fija tiene que llevar la respuesta, en el orden del
	// fichero; vacío si la eval no los espera. Solo los admite una eval que activa la skill, y sus valores son los de
	// boe.CodigosDeAviso (FR-020 a FR-022 de H5.1).
	Avisos []string `yaml:"avisos"`

	// Territorio es lo que la respuesta tiene que declarar del territorio del
	// municipio, cada elemento por su forma fija (ExtraerTerritorio); vacío si la
	// eval no lo espera. Solo lo admite una eval que activa la skill, y con él
	// puede no llevar citas (contrato de evals §1.2 y §1.3 de H6).
	Territorio TerritorioEsperado `yaml:"territorio"`
}

// ComandoEsperado es un comando que la sesión tiene que ejecutar, en una de las
// cuatro formas excluyentes de data-model §6.1 de H5 y de H6, que decide
// formaDelComando: bloque (Applet, Norma y Bloque, sin Verbo), consulta de norma
// (Applet, Verbo indice, metadatos o analisis, y Norma), búsqueda (Applet, Verbo
// buscar y Terminos) o territorio (Applet, Verbo resolver y Municipio). Lo que su
// forma no lleva queda vacío.
type ComandoEsperado struct {
	// Applet es el applet que se invoca, como boe o territorio.
	Applet string `yaml:"applet"`

	// Verbo es el verbo de una consulta de norma, de una búsqueda o de un
	// comando de territorio; vacío en la forma bloque.
	Verbo string `yaml:"verbo"`

	// Norma es el identificador de la norma de la forma bloque o de una
	// consulta de norma.
	Norma string `yaml:"norma"`

	// Bloque es el id del bloque de la forma bloque.
	Bloque string `yaml:"bloque"`

	// Terminos son los términos que tienen que aparecer, como palabras, en los
	// argumentos de una búsqueda.
	Terminos []string `yaml:"terminos"`

	// Municipio es el municipio que resuelve un comando de territorio, tal como
	// lo escribe la eval.
	Municipio string `yaml:"municipio"`
}

// verboResolver es el verbo de un comando de territorio, el único del applet
// territorio (contrato de evals §1.1 de H6).
const verboResolver = "resolver"

// formaDeComando es una de las cuatro formas de un comando esperado.
type formaDeComando int

// Las cuatro formas de un comando esperado (data-model §6.1 de H5 y de H6).
const (
	// formaConsultaDeNorma es la de una consulta de norma: el verbo con la norma.
	formaConsultaDeNorma formaDeComando = iota

	// formaBloque es la de la lectura de un bloque de una norma, sin verbo.
	formaBloque

	// formaBusqueda es la de una búsqueda: buscar con sus términos.
	formaBusqueda

	// formaTerritorio es la de un comando de territorio: resolver con el
	// municipio.
	formaTerritorio
)

// formaDelComando es la forma del comando esperado, que decide su verbo, el que
// distingue las cuatro de data-model §6.1: sin verbo, la forma bloque; buscar, la
// búsqueda; resolver, el comando de territorio; y cualquier otro, la consulta de
// norma, porque el esquema de eval solo admite en esa forma los verbos de su
// enumerado. Es el único sitio que decide la variante: la consumen el
// juicio, el texto del comando y las consultas necesarias (research D21 de H6).
func formaDelComando(comando ComandoEsperado) formaDeComando {
	switch comando.Verbo {
	case "":
		return formaBloque
	case verboBuscar:
		return formaBusqueda
	case verboResolver:
		return formaTerritorio
	default:
		return formaConsultaDeNorma
	}
}

// CitaEsperada es una cita que la respuesta tiene que contener: la pareja exacta
// de norma y bloque (data-model §6.2).
type CitaEsperada struct {
	// Norma es el identificador de la norma citada.
	Norma string `yaml:"norma"`

	// Bloque es el id del bloque citado.
	Bloque string `yaml:"bloque"`
}

// TerritorioEsperado es lo que la respuesta tiene que declarar del territorio de
// un municipio (contrato de evals §1.2 de H6): cada clave que la eval escribe es
// un elemento que se busca en la respuesta por su forma fija, y lo que no escribe
// queda vacío.
type TerritorioEsperado struct {
	// Comunidad es el nombre de la comunidad o ciudad autónoma.
	Comunidad string `yaml:"comunidad"`

	// Provincia es el nombre de la provincia.
	Provincia string `yaml:"provincia"`

	// Boletines son los códigos de los boletines, en el orden del fichero.
	Boletines []string `yaml:"boletines"`

	// Cobertura tiene cada aspecto de cobertura en la forma <aspecto>: <valor>,
	// del vocabulario del applet territorio, en el orden del fichero.
	Cobertura []string `yaml:"cobertura"`
}

// esquemaDeEval compila una sola vez el esquema publicado del formato común de
// eval, que no cambia mientras se ejecutan los tests.
var esquemaDeEval = sync.OnceValues(func() (*jsonschema.Schema, error) {
	return compilarEsquemaDeEval(rutaDelEsquemaDeEval)
})

// compilarEsquemaDeEval lee el esquema del formato común de eval de la ruta y lo
// compila con skills.CompilarEsquema. El error nombra la ruta: la del fichero que
// no se puede leer o la del esquema que no compila.
func compilarEsquemaDeEval(ruta string) (*jsonschema.Schema, error) {
	contenido, err := leerFichero(ruta)
	if err != nil {
		return nil, fmt.Errorf("no se puede leer el esquema del formato de eval: %w", err)
	}

	esquema, err := skills.CompilarEsquema(contenido)
	if err != nil {
		return nil, fmt.Errorf("el esquema del formato de eval %s: %w", ruta, err)
	}

	return esquema, nil
}

// LeerEval lee el contenido de un fichero de eval con el lector común de
// documentos YAML de internal/skills —una clave repetida es un defecto con sus
// dos líneas, nunca la última que gana— y lo valida contra eval.yaml.json. Todo
// error empieza por el nombre del fichero y va con una Eval vacía; la Eval leída
// lleva ese nombre en Fichero (US4, escenario 4; contrato evals-y-grabaciones §1).
func LeerEval(nombre string, contenido []byte) (Eval, error) {
	esquema, err := esquemaDeEval()
	if err != nil {
		return Eval{}, fmt.Errorf("%s: %w", nombre, err)
	}

	eval, err := skills.ValidarDocumentoYAML[Eval](contenido, esquema)
	if err != nil {
		return Eval{}, fmt.Errorf("%s: %w", nombre, err)
	}

	eval.Fichero = nombre

	return eval, nil
}
