package skills_test

import (
	"testing"

	"github.com/santhosh-tekuri/jsonschema/v6/kind"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/jmorenobl/kitlegal/internal/skills"
)

// esquemaDePrueba es el esquema en línea de TestValidarDocumentoYAML: un objeto
// cerrado con un texto obligatorio con patrón, un entero, dos mapas cuyas claves
// tienen patrón —etiquetas, con valores de texto o enteros, y marcas— y una
// lista de mapas de enteros. Acepta cualquiera de los dos valores de cada clave
// repetida de los casos, para que solo el recorrido del lector pueda rechazarlas.
const esquemaDePrueba = `{
  "$schema": "https://json-schema.org/draft/2020-12/schema",
  "type": "object",
  "additionalProperties": false,
  "required": ["nombre"],
  "properties": {
    "nombre": { "type": "string", "pattern": "^[a-z]+$" },
    "edad": { "type": "integer" },
    "etiquetas": {
      "type": "object",
      "propertyNames": { "pattern": "^[a-z]+$" },
      "additionalProperties": { "type": ["string", "integer"] }
    },
    "marcas": { "type": "object", "propertyNames": { "pattern": "^[a-z]+$" } },
    "piezas": {
      "type": "array",
      "items": { "type": "object", "additionalProperties": { "type": "integer" } }
    }
  }
}`

// documentoDePrueba es el tipo al que TestValidarDocumentoYAML lee cada
// documento.
type documentoDePrueba struct {
	Nombre    string           `yaml:"nombre"`
	Edad      int              `yaml:"edad"`
	Etiquetas map[string]any   `yaml:"etiquetas"`
	Piezas    []map[string]int `yaml:"piezas"`
}

// TestValidarDocumentoYAML fija el lector común de data-model («Lectura de
// documentos YAML»; research.md D8): un documento válido se lee entero; cada
// incumplimiento del esquema se presenta con la ruta dentro del documento y la
// línea del nodo que la ocupa; toda clave repetida se rechaza nombrando el mapa,
// la clave y sus dos líneas antes de validar, aunque el esquema aceptara
// cualquiera de los dos valores; y un valor que JSON no representa, empezando por
// un mapa con una clave que no es texto, es un defecto que nombra su ruta.
func TestValidarDocumentoYAML(t *testing.T) {
	t.Parallel()

	esquema, err := skills.CompilarEsquema([]byte(esquemaDePrueba))
	require.NoError(t, err)

	casos := []struct {
		nombre     string
		documento  string
		sinEsquema bool
		leido      documentoDePrueba
		error      string
		clave      *skills.ClaveRepetida
		defecto    *skills.DefectoEnElDocumento
	}{
		{
			nombre: "valido",
			documento: "nombre: ana\n" +
				"edad: 30\n" +
				"etiquetas:\n" +
				"  color: rojo\n" +
				"  talla: 42\n" +
				"piezas:\n" +
				"  - ancho: 3\n" +
				"    alto: 4\n",
			leido: documentoDePrueba{
				Nombre:    "ana",
				Edad:      30,
				Etiquetas: map[string]any{"color": "rojo", "talla": 42},
				Piezas:    []map[string]int{{"ancho": 3, "alto": 4}},
			},
		},
		{
			nombre:     "sin-esquema",
			documento:  "nombre: ANA\nsobrante: 1\n",
			sinEsquema: true,
			leido:      documentoDePrueba{Nombre: "ANA"},
		},
		{
			nombre:    "clave-desconocida",
			documento: "nombre: ana\nsobrante: 1\n",
			error:     "línea 1: additional properties 'sobrante' not allowed",
			defecto: &skills.DefectoEnElDocumento{
				Linea:  1,
				Motivo: "additional properties 'sobrante' not allowed",
				Tipo:   &kind.AdditionalProperties{Properties: []string{"sobrante"}},
			},
		},
		{
			nombre:    "tipo",
			documento: "nombre: ana\nedad: treinta\n",
			error:     "edad, línea 2: got string, want integer",
			defecto: &skills.DefectoEnElDocumento{
				Ruta:   []string{"edad"},
				Linea:  2,
				Motivo: "got string, want integer",
				Tipo:   &kind.Type{Got: "string", Want: []string{"integer"}},
			},
		},
		{
			nombre:    "patron",
			documento: "nombre: Ana\n",
			error:     "nombre, línea 1: 'Ana' does not match pattern '^[a-z]+$'",
			defecto: &skills.DefectoEnElDocumento{
				Ruta:   []string{"nombre"},
				Linea:  1,
				Motivo: "'Ana' does not match pattern '^[a-z]+$'",
				Tipo:   &kind.Pattern{Got: "Ana", Want: "^[a-z]+$"},
			},
		},
		{
			nombre:    "patron-de-clave",
			documento: "nombre: ana\netiquetas:\n  Color: rojo\n",
			error:     "etiquetas, línea 3: invalid propertyName 'Color': 'Color' does not match pattern '^[a-z]+$'",
			defecto: &skills.DefectoEnElDocumento{
				Ruta:   []string{"etiquetas"},
				Linea:  3,
				Motivo: "invalid propertyName 'Color': 'Color' does not match pattern '^[a-z]+$'",
				Tipo:   &kind.PropertyNames{Property: "Color"},
			},
		},
		{
			nombre:    "patron-de-clave-en-mapas-que-no-se-distinguen",
			documento: "nombre: ana\netiquetas:\n  Color: rojo\nmarcas:\n  Color: 1\n",
			error: "invalid propertyName 'Color': 'Color' does not match pattern '^[a-z]+$'\n" +
				"invalid propertyName 'Color': 'Color' does not match pattern '^[a-z]+$'",
			defecto: &skills.DefectoEnElDocumento{
				Motivo: "invalid propertyName 'Color': 'Color' does not match pattern '^[a-z]+$'",
				Tipo:   &kind.PropertyNames{Property: "Color"},
			},
		},
		{
			nombre:    "varios-incumplimientos-en-orden-de-linea",
			documento: "edad: treinta\nnombre: Ana\n",
			error: "edad, línea 1: got string, want integer\n" +
				"nombre, línea 2: 'Ana' does not match pattern '^[a-z]+$'",
			defecto: &skills.DefectoEnElDocumento{
				Ruta:   []string{"edad"},
				Linea:  1,
				Motivo: "got string, want integer",
				Tipo:   &kind.Type{Got: "string", Want: []string{"integer"}},
			},
		},
		{
			nombre:    "clave-repetida-en-la-raiz",
			documento: "nombre: ana\nedad: 30\nedad: 31\n",
			error:     "edad repetido en las líneas 2 y 3",
			clave:     &skills.ClaveRepetida{Clave: "edad", Lineas: [2]int{2, 3}},
		},
		{
			nombre:    "clave-repetida-en-un-mapa-anidado",
			documento: "nombre: ana\netiquetas:\n  color: rojo\n  color: azul\n",
			error:     "etiquetas: color repetido en las líneas 3 y 4",
			clave:     &skills.ClaveRepetida{Mapa: []string{"etiquetas"}, Clave: "color", Lineas: [2]int{3, 4}},
		},
		{
			nombre:    "clave-repetida-en-un-mapa-de-una-lista",
			documento: "nombre: ana\npiezas:\n  - ancho: 3\n    ancho: 4\n",
			error:     "piezas/0: ancho repetido en las líneas 3 y 4",
			clave:     &skills.ClaveRepetida{Mapa: []string{"piezas", "0"}, Clave: "ancho", Lineas: [2]int{3, 4}},
		},
		{
			nombre:    "clave-repetida-antes-que-el-esquema",
			documento: "nombre: Ana\nsobrante: 1\nedad: 30\nedad: 31\nedad: 32\n",
			error:     "edad repetido en las líneas 3 y 4\nedad repetido en las líneas 3 y 5",
			clave:     &skills.ClaveRepetida{Clave: "edad", Lineas: [2]int{3, 4}},
		},
		{
			nombre:    "mapa-con-una-clave-que-no-es-texto",
			documento: "nombre: ana\netiquetas:\n  color: rojo\n  1: uno\n",
			error:     "etiquetas, línea 3: mapa con claves que no son texto",
			defecto: &skills.DefectoEnElDocumento{
				Ruta:   []string{"etiquetas"},
				Linea:  3,
				Motivo: "mapa con claves que no son texto",
			},
		},
		{
			nombre:    "numero-que-json-no-representa",
			documento: "nombre: ana\netiquetas:\n  talla: .inf\n",
			error:     "etiquetas/talla, línea 3: número que JSON no puede representar",
			defecto: &skills.DefectoEnElDocumento{
				Ruta:   []string{"etiquetas", "talla"},
				Linea:  3,
				Motivo: "número que JSON no puede representar",
			},
		},
		{
			nombre:    "yaml-mal-formado",
			documento: "nombre: ana\n  edad: 30\n",
			error:     "no es YAML válido: yaml: line 2: mapping values are not allowed in this context",
		},
		{
			nombre:    "varios-documentos",
			documento: "nombre: ana\n---\nnombre: eva\n",
			error:     "línea 2: empieza un segundo documento YAML y solo se admite uno",
			defecto: &skills.DefectoEnElDocumento{
				Linea:  2,
				Motivo: "empieza un segundo documento YAML y solo se admite uno",
			},
		},
		{
			nombre:    "documento-vacio",
			documento: "",
			error:     "got null, want object",
			defecto: &skills.DefectoEnElDocumento{
				Motivo: "got null, want object",
				Tipo:   &kind.Type{Got: "null", Want: []string{"object"}},
			},
		},
	}

	for _, caso := range casos {
		t.Run(caso.nombre, func(t *testing.T) {
			t.Parallel()

			delCaso := esquema
			if caso.sinEsquema {
				delCaso = nil
			}

			leido, err := skills.ValidarDocumentoYAML[documentoDePrueba]([]byte(caso.documento), delCaso)
			if caso.error == "" {
				require.NoError(t, err)
				assert.Equal(t, caso.leido, leido)

				return
			}

			require.EqualError(t, err, caso.error)
			assert.Zero(t, leido, "un documento con defectos no se entrega a medias")

			if caso.clave != nil {
				var repetida *skills.ClaveRepetida
				require.ErrorAs(t, err, &repetida)
				assert.Equal(t, caso.clave, repetida)
			}

			if caso.defecto != nil {
				var defecto *skills.DefectoEnElDocumento
				require.ErrorAs(t, err, &defecto)
				assert.Equal(t, caso.defecto, defecto)
			}
		})
	}
}
