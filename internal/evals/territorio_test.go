package evals

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/jmorenobl/kitlegal/internal/skills"
)

// TestExtraerTerritorio fija la forma fija de lo que una respuesta declara del
// territorio esperado (contrato de evals §2 de H6), con cada texto literal: la
// comunidad y la provincia por su nombre plegado con territorio.Plegar, como
// palabras enteras, sin distinguir mayúsculas ni diacríticos y con cualquier
// puntuación o blanco entre sus palabras, y no dentro de otra palabra; el boletín
// por su código exacto y como palabra, no en minúsculas ni dentro de otra
// palabra; y el aspecto de cobertura en la forma <aspecto>: <valor>, con blancos o
// sin ellos alrededor de los dos puntos y sin distinguir mayúsculas, y no con
// otra palabra, con otro valor, sin los dos puntos, partido en dos líneas ni fuera
// del vocabulario del applet. Lo declarado va en el orden del esperado y con sus
// repeticiones, y lo que no se declara queda vacío.
func TestExtraerTerritorio(t *testing.T) {
	t.Parallel()

	madrid := TerritorioEsperado{
		Comunidad: "Comunidad de Madrid",
		Provincia: "Madrid",
		Boletines: []string{"BOCM"},
		Cobertura: []string{"boletin_autonomico: configurado", "dir3: no-verificado"},
	}
	castillaYLeon := TerritorioEsperado{Comunidad: "Castilla y León", Provincia: "Ávila"}
	configurado := TerritorioEsperado{Cobertura: []string{"boletin_autonomico: configurado"}}
	bocm := TerritorioEsperado{Boletines: []string{"BOCM"}}

	casos := []struct {
		nombre   string
		esperado TerritorioEsperado

		// textos son los textos del caso, y cada uno por separado da declarado.
		textos    []string
		declarado TerritorioEsperado
	}{
		{
			nombre:   "todo-declarado",
			esperado: madrid,
			textos: []string{
				"Leganés está en la provincia de Madrid, en la Comunidad de Madrid. Boletines: BOE y BOCM.\n" +
					"boletin_autonomico: configurado\ndir3: no-verificado",
			},
			declarado: madrid,
		},
		{
			nombre:   "nombres-sin-tildes-y-en-mayusculas",
			esperado: castillaYLeon,
			textos: []string{
				"Tordesillas está en la provincia de AVILA, en CASTILLA Y LEON.",
				"castilla y leon, ávila",
			},
			declarado: castillaYLeon,
		},
		{
			nombre:   "nombres-con-puntuacion-y-blancos",
			esperado: castillaYLeon,
			textos: []string{
				"(Castilla y León) · «Ávila».",
				"Castilla  y\tLeón, Ávila",
				"**Castilla y León**: Ávila",
				"Castilla y\nLeón / Ávila",
			},
			declarado: castillaYLeon,
		},
		{
			nombre:   "nombres-dentro-de-otras-palabras",
			esperado: castillaYLeon,
			textos: []string{
				"Castilla y Leónesa, en Avilés.",
				"Castillay León y Ávilas",
			},
		},
		{
			nombre:   "nombre-a-medias",
			esperado: castillaYLeon,
			textos:   []string{"El municipio está en León, en Castilla."},
		},
		{
			nombre:    "boletin-como-palabra",
			esperado:  bocm,
			textos:    []string{"BOCM", "Se publica en el BOCM.", "(BOCM)", "`BOCM`", "BOCM-2026", "**BOCM**"},
			declarado: bocm,
		},
		{
			nombre:   "boletin-en-otras-mayusculas-o-dentro-de-otra-palabra",
			esperado: bocm,
			textos:   []string{"bocm", "Bocm", "BOCMA", "XBOCM", "BOCM2", "B O C M"},
		},
		{
			nombre:   "aspecto-con-blancos-y-mayusculas",
			esperado: configurado,
			textos: []string{
				"boletin_autonomico: configurado",
				"boletin_autonomico:configurado",
				"boletin_autonomico :  configurado",
				"boletin_autonomico\t:\tconfigurado",
				"BOLETIN_AUTONOMICO: Configurado",
				"- `boletin_autonomico: configurado`",
				"Cobertura (boletin_autonomico: configurado).",
			},
			declarado: configurado,
		},
		{
			nombre:   "aspecto-sin-su-forma",
			esperado: configurado,
			textos: []string{
				"boletin_autonomico: no-configurado",
				"boletin_autonomico: configurados",
				"boletin_autonomico: configurado-parcial",
				"boletin autonomico: configurado",
				"boletín_autonómico: configurado",
				"subboletin_autonomico: configurado",
				"boletin_autonomico configurado",
				"boletin_autonomico = configurado",
				"boletin_autonomico:\nconfigurado",
				"`boletin_autonomico`: `configurado`",
				"El boletín autonómico está configurado.",
			},
		},
		{
			nombre:    "valor-negado-del-vocabulario",
			esperado:  TerritorioEsperado{Cobertura: []string{"dir3: no-verificado"}},
			textos:    []string{"dir3: no-verificado", "DIR3 : NO-VERIFICADO"},
			declarado: TerritorioEsperado{Cobertura: []string{"dir3: no-verificado"}},
		},
		{
			nombre:   "valor-afirmado-no-esta-en-el-negado",
			esperado: TerritorioEsperado{Cobertura: []string{"dir3: verificado"}},
			textos:   []string{"dir3: no-verificado"},
		},
		{
			nombre:   "aspecto-fuera-del-vocabulario",
			esperado: TerritorioEsperado{Cobertura: []string{"regimen: foral"}},
			textos:   []string{"regimen: foral"},
		},
		{
			nombre:   "orden-y-repeticiones-del-esperado",
			esperado: TerritorioEsperado{Boletines: []string{"BOCM", "BOE", "BOCM"}},
			textos:   []string{"BOE y BOCM"},
			declarado: TerritorioEsperado{
				Boletines: []string{"BOCM", "BOE", "BOCM"},
			},
		},
		{
			nombre:   "solo-una-parte",
			esperado: madrid,
			textos: []string{
				"Leganés está en la Comunidad de Madrid y publica en el BOE.\nboletin_autonomico: no-configurado",
			},
			// La provincia se llama como la comunidad termina: su nombre está en el
			// de la comunidad como palabras enteras, y se da por declarada.
			declarado: TerritorioEsperado{Comunidad: "Comunidad de Madrid", Provincia: "Madrid"},
		},
		{
			nombre:   "texto-vacio",
			esperado: madrid,
			textos:   []string{""},
		},
		{
			nombre: "sin-esperado",
			textos: []string{"Comunidad de Madrid, Madrid, BOCM, boletin_autonomico: configurado"},
		},
	}

	for _, caso := range casos {
		t.Run(caso.nombre, func(t *testing.T) {
			t.Parallel()

			for _, texto := range caso.textos {
				assert.Equal(t, caso.declarado, ExtraerTerritorio(texto, caso.esperado), "texto %q", texto)
			}
		})
	}

	// Cada aspecto del vocabulario del applet, en su forma, da ese aspecto y solo
	// ese: ninguno es el otro valor de su misma clave.
	t.Run("las-seis-combinaciones", func(t *testing.T) {
		t.Parallel()

		todos := TerritorioEsperado{Cobertura: aspectosDeCobertura()}
		require.Len(t, todos.Cobertura, 6, "tres claves con dos valores cada una")

		for _, aspecto := range todos.Cobertura {
			assert.Equal(t, TerritorioEsperado{Cobertura: []string{aspecto}}, ExtraerTerritorio(aspecto, todos),
				"aspecto %s", aspecto)
		}
	})
}

// TestComprobarAspectosDeCobertura fija la comprobación de que el enumerado de
// cobertura del esquema de eval y el vocabulario del applet dicen lo mismo
// (contrato de evals §1.2 de H6), sobre esquemas sintéticos compilados con
// skills.CompilarEsquema: con las seis combinaciones detrás de un $ref, como en el
// esquema publicado, no hay error; sin una, el error la nombra; con otra de más,
// el error la nombra; y sin enumerado, con dos $ref que se apuntan entre sí, sin
// la propiedad cobertura o sin la propiedad territorio, uno por combinación del
// vocabulario, en su orden.
func TestComprobarAspectosDeCobertura(t *testing.T) {
	t.Parallel()

	const (
		// cabecera y pie envuelven la propiedad territorio de cada esquema y sus
		// definiciones.
		cabecera = `{
  "$schema": "https://json-schema.org/draft/2020-12/schema",
  "type": "object",
  "properties": {`
		pie = `
  }
}`

		conCobertura = `
    "territorio": {
      "type": "object",
      "properties": {
        "cobertura": { "type": "array", "minItems": 1, "items": { "$ref": "#/$defs/aspecto-de-cobertura" } }
      }
    }
  },
  "$defs": {`

		ningunoEnumerado = "el esquema de eval no enumera el aspecto de cobertura boletin_autonomico: configurado en territorio/cobertura\n" +
			"el esquema de eval no enumera el aspecto de cobertura boletin_autonomico: no-configurado en territorio/cobertura\n" +
			"el esquema de eval no enumera el aspecto de cobertura boletin_provincial: configurado en territorio/cobertura\n" +
			"el esquema de eval no enumera el aspecto de cobertura boletin_provincial: no-configurado en territorio/cobertura\n" +
			"el esquema de eval no enumera el aspecto de cobertura dir3: verificado en territorio/cobertura\n" +
			"el esquema de eval no enumera el aspecto de cobertura dir3: no-verificado en territorio/cobertura"
	)

	casos := []struct {
		nombre  string
		esquema string

		// error es el error exacto; vacío, no hay error.
		error string
	}{
		{
			nombre: "exacto",
			esquema: cabecera + conCobertura + `
    "aspecto-de-cobertura": { "enum": [
      "boletin_autonomico: configurado", "boletin_autonomico: no-configurado",
      "boletin_provincial: configurado", "boletin_provincial: no-configurado",
      "dir3: verificado", "dir3: no-verificado"
    ] }` + pie,
		},
		{
			nombre: "falta-un-aspecto",
			esquema: cabecera + conCobertura + `
    "aspecto-de-cobertura": { "enum": [
      "boletin_autonomico: configurado", "boletin_autonomico: no-configurado",
      "boletin_provincial: configurado", "boletin_provincial: no-configurado",
      "dir3: verificado"
    ] }` + pie,
			error: "el esquema de eval no enumera el aspecto de cobertura dir3: no-verificado en territorio/cobertura",
		},
		{
			nombre: "sobra-un-aspecto",
			esquema: cabecera + conCobertura + `
    "aspecto-de-cobertura": { "enum": [
      "boletin_autonomico: configurado", "boletin_autonomico: no-configurado",
      "boletin_provincial: configurado", "boletin_provincial: no-configurado",
      "dir3: verificado", "dir3: no-verificado", "regimen: foral", 7
    ] }` + pie,
			error: "el esquema de eval enumera en territorio/cobertura el aspecto regimen: foral, " +
				"que no es un aspecto de cobertura del binario\n" +
				"el esquema de eval enumera en territorio/cobertura el aspecto 7, " +
				"que no es un aspecto de cobertura del binario",
		},
		{
			nombre: "sin-enumerado",
			esquema: cabecera + conCobertura + `
    "aspecto-de-cobertura": { "type": "string" }` + pie,
			error: ningunoEnumerado,
		},
		{
			nombre: "referencia-circular",
			esquema: cabecera + `
    "territorio": {
      "type": "object",
      "properties": {
        "cobertura": { "type": "array", "items": { "$ref": "#/$defs/ida" } }
      }
    }
  },
  "$defs": {
    "ida": { "$ref": "#/$defs/vuelta" },
    "vuelta": { "$ref": "#/$defs/ida" }` + pie,
			error: ningunoEnumerado,
		},
		{
			nombre: "sin-cobertura",
			esquema: cabecera + `
    "territorio": { "type": "object", "properties": { "comunidad": { "type": "string" } } }` + pie,
			error: ningunoEnumerado,
		},
		{
			nombre: "sin-territorio",
			esquema: cabecera + `
    "pregunta": { "type": "string", "minLength": 1 }` + pie,
			error: ningunoEnumerado,
		},
	}

	for _, caso := range casos {
		t.Run(caso.nombre, func(t *testing.T) {
			t.Parallel()

			esquema, err := skills.CompilarEsquema([]byte(caso.esquema))
			require.NoError(t, err)

			err = ComprobarAspectosDeCobertura(esquema)

			if caso.error == "" {
				assert.NoError(t, err)

				return
			}

			assert.EqualError(t, err, caso.error)
		})
	}
}
