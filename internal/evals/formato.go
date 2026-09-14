package evals

import (
	"fmt"
	"os"
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

	// Comandos son los comandos esperados de una eval que activa la skill, en el
	// orden del fichero; vacío en una de no activación (FR-061).
	Comandos []ComandoEsperado `yaml:"comandos"`

	// Citas son las citas esperadas de una eval que activa la skill, en el orden
	// del fichero; vacío en una de no activación (FR-061).
	Citas []CitaEsperada `yaml:"citas"`
}

// ComandoEsperado es un comando que la sesión tiene que ejecutar, en una de las
// tres formas excluyentes de data-model §6.1: bloque (Applet, Norma y Bloque, sin
// Verbo), consulta de norma (Applet, Verbo indice, metadatos o analisis, y Norma)
// o búsqueda (Applet, Verbo buscar y Terminos). Lo que su forma no lleva queda
// vacío.
type ComandoEsperado struct {
	// Applet es el applet que se invoca, como boe.
	Applet string `yaml:"applet"`

	// Verbo es el verbo de una consulta de norma o de una búsqueda; vacío en la
	// forma bloque.
	Verbo string `yaml:"verbo"`

	// Norma es el identificador de la norma de la forma bloque o de una
	// consulta de norma.
	Norma string `yaml:"norma"`

	// Bloque es el id del bloque de la forma bloque.
	Bloque string `yaml:"bloque"`

	// Terminos son los términos que tienen que aparecer, como palabras, en los
	// argumentos de una búsqueda.
	Terminos []string `yaml:"terminos"`
}

// CitaEsperada es una cita que la respuesta tiene que contener: la pareja exacta
// de norma y bloque (data-model §6.2).
type CitaEsperada struct {
	// Norma es el identificador de la norma citada.
	Norma string `yaml:"norma"`

	// Bloque es el id del bloque citado.
	Bloque string `yaml:"bloque"`
}

// esquemaDeEval compila una sola vez el esquema publicado del formato común de
// eval, que no cambia mientras se ejecutan los tests.
var esquemaDeEval = sync.OnceValues(func() (*jsonschema.Schema, error) {
	contenido, err := os.ReadFile(rutaDelEsquemaDeEval)
	if err != nil {
		return nil, fmt.Errorf("no se puede leer el esquema del formato de eval: %w", err)
	}

	esquema, err := skills.CompilarEsquema(contenido)
	if err != nil {
		return nil, fmt.Errorf("el esquema del formato de eval %s: %w", rutaDelEsquemaDeEval, err)
	}

	return esquema, nil
})

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
