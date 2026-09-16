package evals

import (
	"fmt"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/jmorenobl/kitlegal/internal/skills"
	"github.com/jmorenobl/kitlegal/internal/source/boe"
)

// TestExtraerAvisos fija la forma fija de un aviso con los casos del contrato de
// la forma fija §4, cada uno con su texto literal: la marca, la etiqueta y los
// dos puntos en una misma línea, con el selector de presentación, el énfasis de
// Markdown, los blancos, las mayúsculas y lo que rodea a la forma tolerados; y
// ningún aviso con otra redacción, la negación, una palabra de más, pegada o
// distinta, sin la marca o sin los dos puntos, con otros dos puntos, con dos
// selectores, con un salto de línea o un tachado dentro, con otro emoji o en un
// texto vacío. Los códigos salen en el orden de boe.CodigosDeAviso y sin repetir,
// y lo que sigue a la forma no se lee (FR-032, FR-033; US2, escenarios 3 a 6;
// SC-003; research D2 y D3).
func TestExtraerAvisos(t *testing.T) {
	t.Parallel()

	derogada := []string{"derogada"}

	casos := []struct {
		nombre string

		// textos son los textos del caso, y cada uno por separado da los
		// códigos.
		textos []string

		codigos []string
	}{
		{nombre: "forma-fija", textos: []string{"⚠ NORMA DEROGADA: esta norma ha sido derogada."}, codigos: derogada},
		{nombre: "selector-emoji", textos: []string{"⚠️ NORMA DEROGADA:"}, codigos: derogada},
		{nombre: "selector-texto", textos: []string{"⚠︎ NORMA DEROGADA:"}, codigos: derogada},
		{nombre: "enfasis-envolvente", textos: []string{"**⚠ NORMA DEROGADA:**"}, codigos: derogada},
		{nombre: "enfasis-en-la-etiqueta", textos: []string{"⚠ **NORMA DEROGADA**:"}, codigos: derogada},
		{
			nombre:  "enfasis-con-guiones-bajos",
			textos:  []string{"_⚠ NORMA DEROGADA:_", "⚠ __NORMA DEROGADA__:"},
			codigos: derogada,
		},
		{nombre: "espacios", textos: []string{"⚠  NORMA \t DEROGADA :"}, codigos: derogada},
		{
			nombre:  "blancos-de-unicode",
			textos:  []string{"⚠ NORMA DEROGADA:", "⚠ NORMA DEROGADA :"},
			codigos: derogada,
		},
		{nombre: "sin-blanco-tras-la-marca", textos: []string{"⚠NORMA DEROGADA:"}, codigos: derogada},
		{nombre: "minusculas", textos: []string{"⚠ norma derogada:"}, codigos: derogada},
		{nombre: "mayusculas-mezcladas", textos: []string{"⚠ Norma Derogada:"}, codigos: derogada},
		{nombre: "dentro-de-una-cita-en-bloque", textos: []string{"> ⚠ NORMA DEROGADA:"}, codigos: derogada},
		{
			nombre:  "en-otra-linea",
			textos:  []string{"El artículo 42 regula la obligación de resolver.\n⚠ NORMA DEROGADA:"},
			codigos: derogada,
		},
		{
			nombre:  "orden-de-los-codigos",
			textos:  []string{"⚠ VIGENCIA AGOTADA:\n⚠ NORMA DEROGADA:"},
			codigos: []string{"derogada", "vigencia-agotada"},
		},
		{nombre: "repetida", textos: []string{"⚠ NORMA DEROGADA:\n⚠ NORMA DEROGADA:"}, codigos: derogada},
		{
			// La limitación declarada: lo que sigue a la forma no se lee.
			nombre:  "forma-fija-y-lo-contrario",
			textos:  []string{"⚠ NORMA DEROGADA: pero sigue en vigor."},
			codigos: derogada,
		},
		{nombre: "otra-redaccion", textos: []string{"Esta ley fue derogada."}},
		{nombre: "negacion", textos: []string{"La Ley 30/1992 sigue en vigor."}},
		{nombre: "palabra-de-mas", textos: []string{"⚠ NORMA PARCIALMENTE DEROGADA:"}},
		{nombre: "palabras-pegadas", textos: []string{"⚠ NORMADEROGADA:"}},
		{nombre: "palabra-distinta", textos: []string{"⚠ NORMA DEROGADAS:"}},
		{nombre: "sin-marca", textos: []string{"NORMA DEROGADA:"}},
		{nombre: "sin-dos-puntos", textos: []string{"⚠ NORMA DEROGADA"}},
		{nombre: "dos-puntos-de-ancho-completo", textos: []string{"⚠ NORMA DEROGADA："}},
		{nombre: "dos-selectores", textos: []string{"⚠️️ NORMA DEROGADA:"}},
		{nombre: "salto-dentro", textos: []string{"⚠ NORMA\nDEROGADA:"}},
		{nombre: "tachado", textos: []string{"⚠ NORMA ~~DEROGADA~~:"}},
		{nombre: "otro-emoji", textos: []string{"🚨 NORMA DEROGADA:"}},
		{nombre: "vacio", textos: []string{""}},
	}

	for _, caso := range casos {
		t.Run(caso.nombre, func(t *testing.T) {
			t.Parallel()

			for _, texto := range caso.textos {
				assert.Equal(t, caso.codigos, ExtraerAvisos(texto), "texto %q", texto)
			}
		})
	}

	// Cada etiqueta del binario, con la marca delante y los dos puntos detrás, da
	// su código y solo el suyo.
	t.Run("las-tres-etiquetas", func(t *testing.T) {
		t.Parallel()

		etiquetas := boe.EtiquetasDeAviso()

		for _, codigo := range boe.CodigosDeAviso() {
			assert.Equal(t, []string{codigo}, ExtraerAvisos("⚠ "+etiquetas[codigo]+":"), "código %s", codigo)
		}
	})
}

// Las formas fijas de los tres avisos como las enseña una lista, una por línea, y
// la de derogada dentro de ella, que cada caso de TestComprobarFormasDeAviso sin
// una de sus partes sustituye.
const (
	formasDeLosAvisos = "- `⚠ TEXTO POSIBLEMENTE DESACTUALIZADO:` para el aviso `consolidacion-no-finalizada`.\n" +
		"- `⚠ NORMA DEROGADA:` para el aviso `derogada`.\n" +
		"- `⚠ VIGENCIA AGOTADA:` para el aviso `vigencia-agotada`.\n"
	formaDeDerogadaEnLaLista = "`⚠ NORMA DEROGADA:`"
)

// TestComprobarFormasDeAviso fija la comprobación de FR-014 sobre textos
// sintéticos: con las tres formas no hay error; sin la etiqueta, sin la marca o
// sin los dos puntos de la forma de derogada, el error es exactamente el de
// derogada con su forma; y sin ninguna forma, uno por código en el orden de
// boe.CodigosDeAviso (US4, escenario 3; SC-005; research D7).
func TestComprobarFormasDeAviso(t *testing.T) {
	t.Parallel()

	require.Equal(t, 1, strings.Count(formasDeLosAvisos, formaDeDerogadaEnLaLista),
		"los casos sustituyen la única forma de derogada del texto completo")

	const faltaDerogada = "falta la forma fija del aviso derogada: ⚠ NORMA DEROGADA:"

	casos := []struct {
		nombre string
		texto  string

		// error es el error exacto; vacío, no hay error.
		error string
	}{
		{nombre: "completo", texto: formasDeLosAvisos},
		{
			nombre: "falta-la-etiqueta",
			texto:  strings.Replace(formasDeLosAvisos, formaDeDerogadaEnLaLista, "`⚠ :`", 1),
			error:  faltaDerogada,
		},
		{
			nombre: "falta-la-marca",
			texto:  strings.Replace(formasDeLosAvisos, formaDeDerogadaEnLaLista, "`NORMA DEROGADA:`", 1),
			error:  faltaDerogada,
		},
		{
			nombre: "faltan-los-dos-puntos",
			texto:  strings.Replace(formasDeLosAvisos, formaDeDerogadaEnLaLista, "`⚠ NORMA DEROGADA`", 1),
			error:  faltaDerogada,
		},
		{
			nombre: "ninguna",
			texto:  "Traslada los avisos de vigencia del sobre.\n",
			error: "falta la forma fija del aviso consolidacion-no-finalizada: ⚠ TEXTO POSIBLEMENTE DESACTUALIZADO:\n" +
				faltaDerogada + "\n" +
				"falta la forma fija del aviso vigencia-agotada: ⚠ VIGENCIA AGOTADA:",
		},
	}

	for _, caso := range casos {
		t.Run(caso.nombre, func(t *testing.T) {
			t.Parallel()

			err := ComprobarFormasDeAviso(caso.texto)

			if caso.error == "" {
				assert.NoError(t, err)

				return
			}

			assert.EqualError(t, err, caso.error)
		})
	}
}

// TestComprobarCodigosDeAviso fija la comprobación de FR-013 sobre esquemas
// sintéticos compilados con skills.CompilarEsquema: con el enumerado de los tres
// códigos detrás de un $ref, como en el esquema publicado, no hay error; sin
// derogada, el error la nombra; con otro-aviso de más, el error lo nombra; y con
// items sin enumerado, con dos $ref que se apuntan entre sí —el compilador los
// acepta— o sin la propiedad avisos, uno por código en el orden de
// boe.CodigosDeAviso (US4, escenario 2; SC-005; research D5).
func TestComprobarCodigosDeAviso(t *testing.T) {
	t.Parallel()

	const ningunoEnumerado = "el esquema de eval no enumera el código de aviso consolidacion-no-finalizada en avisos\n" +
		"el esquema de eval no enumera el código de aviso derogada en avisos\n" +
		"el esquema de eval no enumera el código de aviso vigencia-agotada en avisos"

	casos := []struct {
		nombre  string
		esquema string

		// error es el error exacto; vacío, no hay error.
		error string
	}{
		{
			nombre: "exacto",
			esquema: `{
  "$schema": "https://json-schema.org/draft/2020-12/schema",
  "type": "object",
  "properties": {
    "avisos": { "type": "array", "minItems": 1, "items": { "$ref": "#/$defs/codigo-de-aviso" } }
  },
  "$defs": {
    "codigo-de-aviso": { "enum": ["consolidacion-no-finalizada", "derogada", "vigencia-agotada"] }
  }
}`,
		},
		{
			nombre: "falta-un-codigo",
			esquema: `{
  "$schema": "https://json-schema.org/draft/2020-12/schema",
  "type": "object",
  "properties": {
    "avisos": { "type": "array", "minItems": 1, "items": { "$ref": "#/$defs/codigo-de-aviso" } }
  },
  "$defs": {
    "codigo-de-aviso": { "enum": ["consolidacion-no-finalizada", "vigencia-agotada"] }
  }
}`,
			error: "el esquema de eval no enumera el código de aviso derogada en avisos",
		},
		{
			nombre: "sobra-un-codigo",
			esquema: `{
  "$schema": "https://json-schema.org/draft/2020-12/schema",
  "type": "object",
  "properties": {
    "avisos": { "type": "array", "minItems": 1, "items": { "$ref": "#/$defs/codigo-de-aviso" } }
  },
  "$defs": {
    "codigo-de-aviso": { "enum": ["consolidacion-no-finalizada", "derogada", "vigencia-agotada", "otro-aviso"] }
  }
}`,
			error: "el esquema de eval enumera en avisos el código otro-aviso, que no es un código de aviso del binario",
		},
		{
			nombre: "sin-enumerado",
			esquema: `{
  "$schema": "https://json-schema.org/draft/2020-12/schema",
  "type": "object",
  "properties": {
    "avisos": { "type": "array", "minItems": 1, "items": { "type": "string" } }
  },
  "$defs": {
    "codigo-de-aviso": { "enum": ["consolidacion-no-finalizada", "derogada", "vigencia-agotada"] }
  }
}`,
			error: ningunoEnumerado,
		},
		{
			nombre: "referencia-circular",
			esquema: `{
  "$schema": "https://json-schema.org/draft/2020-12/schema",
  "type": "object",
  "properties": {
    "avisos": { "type": "array", "minItems": 1, "items": { "$ref": "#/$defs/ida" } }
  },
  "$defs": {
    "ida": { "$ref": "#/$defs/vuelta" },
    "vuelta": { "$ref": "#/$defs/ida" }
  }
}`,
			error: ningunoEnumerado,
		},
		{
			nombre: "sin-avisos",
			esquema: `{
  "$schema": "https://json-schema.org/draft/2020-12/schema",
  "type": "object",
  "properties": {
    "pregunta": { "type": "string", "minLength": 1 }
  },
  "$defs": {
    "codigo-de-aviso": { "enum": ["consolidacion-no-finalizada", "derogada", "vigencia-agotada"] }
  }
}`,
			error: ningunoEnumerado,
		},
	}

	for _, caso := range casos {
		t.Run(caso.nombre, func(t *testing.T) {
			t.Parallel()

			esquema, err := skills.CompilarEsquema([]byte(caso.esquema))
			require.NoError(t, err)

			err = ComprobarCodigosDeAviso(esquema)

			if caso.error == "" {
				assert.NoError(t, err)

				return
			}

			assert.EqualError(t, err, caso.error)
		})
	}
}

// TestEtiquetasSoloDesdeBoe comprueba que ningún fichero de código del paquete
// —los *.go que no terminan en _test.go— contiene una etiqueta de
// boe.EtiquetasDeAviso, tampoco en un comentario: la forma fija se compone con
// las del binario y no con una copia. Para no pasar en vacío, entre los ficheros
// leídos tiene que estar avisos.go (FR-011; US4, escenario 5; SC-005; research
// D8).
func TestEtiquetasSoloDesdeBoe(t *testing.T) {
	t.Parallel()

	ficheros, err := filepath.Glob("*.go")
	require.NoError(t, err)

	etiquetas := boe.EtiquetasDeAviso()

	var leidos, copias []string

	for _, fichero := range ficheros {
		if strings.HasSuffix(fichero, "_test.go") {
			continue
		}

		contenido, err := leerFichero(fichero)
		require.NoError(t, err)

		leidos = append(leidos, fichero)

		for numero, linea := range strings.Split(string(contenido), "\n") {
			for _, codigo := range boe.CodigosDeAviso() {
				if strings.Contains(linea, etiquetas[codigo]) {
					copias = append(copias, fmt.Sprintf("%s:%d: la etiqueta del aviso %s", fichero, numero+1, codigo))
				}
			}
		}
	}

	assert.Contains(t, leidos, "avisos.go")
	assert.Empty(t, copias)
}
