package skills_test

import (
	"encoding/json"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/jmorenobl/kitlegal/internal/skills"
)

// Los valores de los dos enumerados de schemas/jerarquia.yaml.json, en su orden
// (data-model §4.2; FR-066): los cinco niveles territoriales y las cuatro reglas
// de interpretación.
var (
	nivelesDeLaJerarquia = []string{"ue", "estado", "comunidad-autonoma", "provincia", "municipio"}

	reglasDeInterpretacion = []string{
		"competencia-antes-que-jerarquia", "ley-posterior", "ley-especial", "reglamento-nunca-contra-ley",
	}
)

// La jerarquía sintética de TestLeerJerarquia: un nivel por línea, de la 2 a la
// 6, y las reglas de la 7 a la 11, cada uno en el orden de su enumerado. Solo los
// valores de nivel y de regla son los del esquema, que no admite otros: los
// nombres, los boletines, los tipos de norma y los enunciados son inventados.
const (
	inicioDeLosNiveles = "niveles:\n"

	nivelDeLaUnion = "  - {nivel: ue, nombre: \"Unión de Prueba\", boletin: \"Diario de la Unión de Prueba\", " +
		"normas: [\"Reglamento de prueba\", \"Directiva de prueba\"]}\n"
	nivelDelEstado = "  - {nivel: estado, nombre: \"Estado de Prueba\", boletin: \"Boletín del Estado de Prueba\", " +
		"normas: [\"Ley de prueba\", \"Decreto de prueba\"]}\n"
	nivelDeLaComunidad = "  - {nivel: comunidad-autonoma, nombre: \"Comunidad de Prueba\", " +
		"boletin: \"Boletín de la Comunidad de Prueba\", normas: [\"Ley de prueba\"]}\n"
	nivelDeLaProvincia = "  - {nivel: provincia, nombre: \"Provincia de Prueba\", " +
		"boletin: \"Boletín de la Provincia de Prueba\", normas: [\"Ordenanza de prueba\"]}\n"
	nivelDelMunicipio = "  - {nivel: municipio, nombre: \"Municipio de Prueba\", " +
		"boletin: \"Boletín de la Provincia de Prueba\", normas: [\"Ordenanza de prueba\", \"Bando de prueba\"]}\n"

	reglasDePrueba = "reglas:\n" +
		"  - {regla: competencia-antes-que-jerarquia, enunciado: \"Primero la competencia de prueba.\"}\n" +
		"  - {regla: ley-posterior, enunciado: \"La posterior de prueba deroga a la anterior.\"}\n" +
		"  - {regla: ley-especial, enunciado: \"La especial de prueba prevalece sobre la general.\"}\n" +
		"  - {regla: reglamento-nunca-contra-ley, enunciado: \"Ningún reglamento de prueba contra una ley.\"}\n"

	jerarquiaDePrueba = inicioDeLosNiveles +
		nivelDeLaUnion + nivelDelEstado + nivelDeLaComunidad + nivelDeLaProvincia + nivelDelMunicipio +
		reglasDePrueba
)

// jerarquiaLeidaDePrueba es lo que LeerJerarquia lee de jerarquiaDePrueba.
func jerarquiaLeidaDePrueba() skills.Jerarquia {
	return skills.Jerarquia{
		Niveles: []skills.NivelNormativo{
			{
				Codigo:  "ue",
				Nombre:  "Unión de Prueba",
				Boletin: "Diario de la Unión de Prueba",
				Normas:  []string{"Reglamento de prueba", "Directiva de prueba"},
			},
			{
				Codigo:  "estado",
				Nombre:  "Estado de Prueba",
				Boletin: "Boletín del Estado de Prueba",
				Normas:  []string{"Ley de prueba", "Decreto de prueba"},
			},
			{
				Codigo:  "comunidad-autonoma",
				Nombre:  "Comunidad de Prueba",
				Boletin: "Boletín de la Comunidad de Prueba",
				Normas:  []string{"Ley de prueba"},
			},
			{
				Codigo:  "provincia",
				Nombre:  "Provincia de Prueba",
				Boletin: "Boletín de la Provincia de Prueba",
				Normas:  []string{"Ordenanza de prueba"},
			},
			{
				Codigo:  "municipio",
				Nombre:  "Municipio de Prueba",
				Boletin: "Boletín de la Provincia de Prueba",
				Normas:  []string{"Ordenanza de prueba", "Bando de prueba"},
			},
		},
		Reglas: []skills.ReglaDeInterpretacion{
			{Codigo: "competencia-antes-que-jerarquia", Enunciado: "Primero la competencia de prueba."},
			{Codigo: "ley-posterior", Enunciado: "La posterior de prueba deroga a la anterior."},
			{Codigo: "ley-especial", Enunciado: "La especial de prueba prevalece sobre la general."},
			{Codigo: "reglamento-nunca-contra-ley", Enunciado: "Ningún reglamento de prueba contra una ley."},
		},
	}
}

// TestLeerJerarquia fija LeerJerarquia sobre la jerarquía sintética (data-model
// §4.2; FR-044, FR-066): el documento válido da sus niveles y sus reglas en el
// orden de sus enumerados; y cada defecto —un nivel o una regla fuera de su
// enumerado, un nivel fuera de su sitio, una clave que el esquema no declara, una
// clave repetida— es un error que nombra el punto del documento y su línea, y con
// él no se entrega nada.
func TestLeerJerarquia(t *testing.T) {
	t.Parallel()

	casos := []struct {
		nombre    string
		documento func(t *testing.T) string
		error     string
		// repetida es la clave repetida que el error lleva dentro, si la hay.
		repetida *skills.ClaveRepetida
	}{
		{
			nombre:    "valido",
			documento: func(*testing.T) string { return jerarquiaDePrueba },
		},
		{
			// Un valor fuera del enumerado incumple a la vez el enumerado del nivel
			// y el valor fijo de su sitio en la lista.
			nombre: "nivel-fuera-del-enumerado",
			documento: func(t *testing.T) string {
				t.Helper()

				return cambiada(t, jerarquiaDePrueba, "nivel: provincia", "nivel: comarca")
			},
			error: "niveles/3/nivel, línea 5: value must be 'provincia'\n" +
				"niveles/3/nivel, línea 5: value must be one of " +
				"'ue', 'estado', 'comunidad-autonoma', 'provincia', 'municipio'",
		},
		{
			nombre: "niveles-en-otro-orden",
			documento: func(*testing.T) string {
				return inicioDeLosNiveles +
					nivelDeLaUnion + nivelDelEstado + nivelDeLaComunidad + nivelDelMunicipio + nivelDeLaProvincia +
					reglasDePrueba
			},
			error: "niveles/3/nivel, línea 5: value must be 'provincia'\n" +
				"niveles/4/nivel, línea 6: value must be 'municipio'",
		},
		{
			nombre: "regla-desconocida",
			documento: func(t *testing.T) string {
				t.Helper()

				return cambiada(t, jerarquiaDePrueba, "regla: ley-especial", "regla: ley-favorable")
			},
			error: "reglas/2/regla, línea 10: value must be 'ley-especial'\n" +
				"reglas/2/regla, línea 10: value must be one of " +
				"'competencia-antes-que-jerarquia', 'ley-posterior', 'ley-especial', 'reglamento-nunca-contra-ley'",
		},
		{
			nombre: "clave-desconocida",
			documento: func(t *testing.T) string {
				t.Helper()

				return cambiada(t, jerarquiaDePrueba, "nivel: ue,", "nivel: ue, vigencia: siempre,")
			},
			error: "niveles/0, línea 2: additional properties 'vigencia' not allowed",
		},
		{
			nombre:    "clave-repetida",
			documento: func(*testing.T) string { return jerarquiaDePrueba + reglasDePrueba },
			error:     "reglas repetido en las líneas 7 y 12",
			repetida:  &skills.ClaveRepetida{Clave: "reglas", Lineas: [2]int{7, 12}},
		},
	}

	for _, caso := range casos {
		t.Run(caso.nombre, func(t *testing.T) {
			t.Parallel()

			jerarquia, err := skills.LeerJerarquia([]byte(caso.documento(t)))
			if caso.error == "" {
				require.NoError(t, err)
				assert.Equal(t, jerarquiaLeidaDePrueba(), jerarquia)

				return
			}

			require.EqualError(t, err, caso.error)
			assert.Zero(t, jerarquia, "una jerarquía con defectos no se entrega a medias")

			if caso.repetida != nil {
				var repetida *skills.ClaveRepetida
				require.ErrorAs(t, err, &repetida)
				assert.Equal(t, caso.repetida, repetida)
			}
		})
	}
}

// esquemaPublicadoDeJerarquia es el esquema de data/jerarquia.yaml, relativo al
// directorio de este paquete, que es donde go test ejecuta sus tests: el mismo
// fichero con el que valida LeerJerarquia.
const esquemaPublicadoDeJerarquia = "../../schemas/jerarquia.yaml.json"

// listaEnOrden es lo que el esquema de la jerarquía declara de una de sus dos
// listas: cuántos elementos exige como mínimo, si admite alguno más allá de los
// de prefixItems y, de cada uno de estos, la definición a la que remite y el
// valor fijo que le pone a su clave.
type listaEnOrden struct {
	MinItems    int   `json:"minItems"`
	Items       *bool `json:"items"`
	PrefixItems []struct {
		Ref        string `json:"$ref"`
		Properties map[string]struct {
			Const string `json:"const"`
		} `json:"properties"`
	} `json:"prefixItems"`
}

// enumeradoDeLaClave es lo que una definición del esquema declara de su clave:
// los valores de su enumerado.
type enumeradoDeLaClave map[string]struct {
	Enum []string `json:"enum"`
}

// TestEsquemaDeJerarquia comprueba el esquema publicado de data/jerarquia.yaml
// (data-model §4.2; FR-044, FR-067): compila con las aserciones de formato
// activas, y cada una de sus dos listas —niveles y reglas— exige exactamente los
// valores de su enumerado, cada uno una vez y en el orden del enumerado, que es el
// de data-model; así, el orden en que LeerJerarquia los devuelve es siempre ese, y
// ningún valor del enumerado queda sin sitio en la lista.
func TestEsquemaDeJerarquia(t *testing.T) {
	t.Parallel()

	contenido, err := os.ReadFile(esquemaPublicadoDeJerarquia)
	require.NoError(t, err)

	var esquema struct {
		Properties struct {
			Niveles listaEnOrden `json:"niveles"`
			Reglas  listaEnOrden `json:"reglas"`
		} `json:"properties"`
		Defs struct {
			Nivel struct {
				Properties enumeradoDeLaClave `json:"properties"`
			} `json:"nivel"`
			Regla struct {
				Properties enumeradoDeLaClave `json:"properties"`
			} `json:"regla"`
		} `json:"$defs"`
	}
	require.NoError(t, json.Unmarshal(contenido, &esquema), "el esquema de la jerarquía tiene la forma de data-model §4.2")

	t.Run("compila", func(t *testing.T) {
		t.Parallel()

		compilado, err := skills.CompilarEsquema(contenido)
		require.NoError(t, err)
		assert.NotNil(t, compilado)
	})

	t.Run("niveles", func(t *testing.T) {
		t.Parallel()

		comprobarListaEnOrden(t, esquema.Properties.Niveles, esquema.Defs.Nivel.Properties,
			"nivel", nivelesDeLaJerarquia)
	})

	t.Run("reglas", func(t *testing.T) {
		t.Parallel()

		comprobarListaEnOrden(t, esquema.Properties.Reglas, esquema.Defs.Regla.Properties,
			"regla", reglasDeInterpretacion)
	})
}

// comprobarListaEnOrden exige que el enumerado de la clave entre las propiedades
// de la definición del mismo nombre sea el esperado, y que la lista del esquema
// traiga un elemento por cada valor, en su orden: cada uno remite a esa
// definición y le pone a la clave ese valor fijo; la lista exige al menos tantos
// como valores y no admite ninguno más.
func comprobarListaEnOrden(t *testing.T, lista listaEnOrden, propiedades enumeradoDeLaClave, clave string,
	esperados []string,
) {
	t.Helper()

	require.Equal(t, esperados, propiedades[clave].Enum,
		"el enumerado de %s es el de data-model §4.2, en su orden", clave)

	fijos := make([]string, 0, len(lista.PrefixItems))

	for posicion, elemento := range lista.PrefixItems {
		assert.Equal(t, "#/$defs/"+clave, elemento.Ref, "el elemento %d de la lista remite a la definición de %s",
			posicion, clave)

		fijos = append(fijos, elemento.Properties[clave].Const)
	}

	assert.Equal(t, esperados, fijos, "la lista trae cada valor de %s una vez y en el orden de su enumerado", clave)
	assert.Equal(t, len(esperados), lista.MinItems, "la lista exige todos los valores de %s", clave)

	require.NotNil(t, lista.Items, "la lista declara items")
	assert.False(t, *lista.Items, "la lista no admite más elementos que los valores de %s", clave)
}

// jerarquiaDelRepositorio es data/jerarquia.yaml, relativo al directorio de este
// paquete, que es donde go test ejecuta sus tests.
const jerarquiaDelRepositorio = "../../data/jerarquia.yaml"

// TestJerarquiaDelRepositorio comprueba la jerarquía normativa del repositorio,
// de la que sale references/jerarquia_normativa.md (data-model §4.2; FR-044,
// FR-066, FR-067; SC-009): LeerJerarquia la lee sin ningún defecto, y trae los
// cinco niveles en su orden y las cuatro reglas. Sin el fichero, o sin sus niveles
// o sus reglas, el test falla en vez de pasar en vacío.
func TestJerarquiaDelRepositorio(t *testing.T) {
	t.Parallel()

	contenido, err := os.ReadFile(jerarquiaDelRepositorio)
	require.NoError(t, err, "la jerarquía normativa del repositorio no se puede leer")

	jerarquia, err := skills.LeerJerarquia(contenido)
	require.NoError(t, err, "la jerarquía normativa del repositorio, %s", jerarquiaDelRepositorio)

	niveles := make([]string, 0, len(jerarquia.Niveles))
	for _, nivel := range jerarquia.Niveles {
		niveles = append(niveles, nivel.Codigo)
	}

	reglas := make([]string, 0, len(jerarquia.Reglas))
	for _, regla := range jerarquia.Reglas {
		reglas = append(reglas, regla.Codigo)
	}

	assert.Equal(t, nivelesDeLaJerarquia, niveles, "%s trae los cinco niveles en su orden", jerarquiaDelRepositorio)
	assert.Equal(t, reglasDeInterpretacion, reglas, "%s trae las cuatro reglas en su orden", jerarquiaDelRepositorio)
}

// TestCompilarEsquemaDeJerarquiaDesdeUnaRuta fija las dos ramas de error con
// las que no se obtiene el esquema de data/jerarquia.yaml, que la ruta constante
// del paquete no da nunca: el fichero que no se puede leer y el esquema que no
// compila. Es el equivalente de TestCompilarEsquemaDelTerritorioDesdeUnaRuta y
// de su gemelo de normas.
func TestCompilarEsquemaDeJerarquiaDesdeUnaRuta(t *testing.T) {
	t.Parallel()

	comprobarCompilarEsquemaDesdeUnaRuta(t, skills.CompilarEsquemaDeJerarquia, "../../schemas/jerarquia.yaml.json",
		"no se puede leer el esquema de data/jerarquia.yaml: ", "el esquema de data/jerarquia.yaml ", "niveles")
}
