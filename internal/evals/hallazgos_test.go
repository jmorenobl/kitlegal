package evals

import (
	"maps"
	"slices"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/jmorenobl/kitlegal/internal/core/grafo"
	"github.com/jmorenobl/kitlegal/internal/skills"
	"github.com/jmorenobl/kitlegal/internal/source/boe"
)

// TestEtiquetasDeHallazgo comprueba el vocabulario de las formas fijas de los
// hallazgos contra el de los avisos: solo version-obsoleta está etiquetada
// —fuente-caducada no se traslada—, ninguna etiqueta de hallazgo es la de un
// aviso, ni siquiera con otras mayúsculas, y la forma fija escrita de cada una no
// se reconoce como la de ningún aviso, ni la de un aviso como la de ningún
// hallazgo (FR-045, FR-054; contrato evals-y-skill §3; research D12).
func TestEtiquetasDeHallazgo(t *testing.T) {
	t.Parallel()

	etiquetas := grafo.EtiquetasDeHallazgo()
	avisos := boe.EtiquetasDeAviso()

	t.Run("solo-version-obsoleta", func(t *testing.T) {
		t.Parallel()

		assert.Equal(t, []grafo.ClaseDeHallazgo{grafo.ClaseVersionObsoleta}, slices.Sorted(maps.Keys(etiquetas)))
		assert.NotContains(t, etiquetas, grafo.ClaseFuenteCaducada)
	})

	t.Run("ninguna-es-de-un-aviso", func(t *testing.T) {
		t.Parallel()

		require.NotEmpty(t, avisos, "boe tiene etiquetas de aviso con las que comparar")

		for clase, etiqueta := range etiquetas {
			for codigo, deAviso := range avisos {
				assert.False(t, strings.EqualFold(etiqueta, deAviso),
					"la etiqueta del hallazgo %s es la del aviso %s", clase, codigo)
			}

			assert.Empty(t, ExtraerAvisos(formaEscrita(etiqueta)),
				"la forma fija del hallazgo %s se reconoce como la de un aviso", clase)
		}

		for codigo, deAviso := range avisos {
			assert.Empty(t, ExtraerHallazgos(formaEscrita(deAviso)),
				"la forma fija del aviso %s se reconoce como la de un hallazgo", codigo)
		}
	})
}

// TestExtraerHallazgos fija la forma fija de un hallazgo con las tolerancias de
// la de los avisos de H5.1, cada caso con su texto literal: la marca, la etiqueta
// y los dos puntos en una misma línea, con el selector de presentación, el
// énfasis de Markdown, los blancos, las mayúsculas y lo que rodea a la forma
// tolerados; y ninguna clase con la etiqueta dicha con otras palabras, con una
// palabra de más, pegada o distinta, sin la marca o sin los dos puntos, partida
// en dos líneas, con la forma de un aviso o en un texto vacío. Las clases salen
// sin repetir (FR-051, FR-052; SC-008; contrato evals-y-skill §2; research V14).
func TestExtraerHallazgos(t *testing.T) {
	t.Parallel()

	obsoleta := []string{string(grafo.ClaseVersionObsoleta)}

	casos := []struct {
		nombre string

		// textos son los textos del caso, y cada uno por separado da las clases.
		textos []string

		clases []string
	}{
		{
			nombre: "forma-fija",
			textos: []string{
				"\xe2\x9a\xa0 REDACCI\xc3\x93N MODIFICADA: la redacci\xc3\xb3n con fecha de vigencia 20161002, " +
					"la que se consult\xc3\xb3 antes, ha sido sustituida por la de 20250101, que es la que se cita.",
			},
			clases: obsoleta,
		},
		{
			nombre: "selector-de-variante",
			textos: []string{
				"\xe2\x9a\xa0\xef\xb8\x8f REDACCI\xc3\x93N MODIFICADA:",
				"\xe2\x9a\xa0\xef\xb8\x8e REDACCI\xc3\x93N MODIFICADA:",
			},
			clases: obsoleta,
		},
		{
			nombre: "enfasis-de-markdown",
			textos: []string{
				"**\xe2\x9a\xa0 REDACCI\xc3\x93N MODIFICADA:**",
				"\xe2\x9a\xa0 **REDACCI\xc3\x93N MODIFICADA**:",
				"_\xe2\x9a\xa0 REDACCI\xc3\x93N MODIFICADA:_",
				"\xe2\x9a\xa0 __REDACCI\xc3\x93N MODIFICADA__:",
			},
			clases: obsoleta,
		},
		{
			nombre: "espacios",
			textos: []string{
				"\xe2\x9a\xa0  REDACCI\xc3\x93N \t MODIFICADA :",
				"\xe2\x9a\xa0\xc2\xa0REDACCI\xc3\x93N\xe2\x80\x83MODIFICADA\xc2\xa0:",
				"\xe2\x9a\xa0REDACCI\xc3\x93N MODIFICADA:",
			},
			clases: obsoleta,
		},
		{
			nombre: "mayusculas",
			textos: []string{
				"\xe2\x9a\xa0 redacci\xc3\xb3n modificada:",
				"\xe2\x9a\xa0 Redacci\xc3\xb3n Modificada:",
			},
			clases: obsoleta,
		},
		{
			// El caso que nombra el contrato: el selector del emoji, el énfasis
			// alrededor de la etiqueta y las mayúsculas a la vez.
			nombre: "tolerancias-a-la-vez",
			textos: []string{"**\xe2\x9a\xa0\xef\xb8\x8f Redacci\xc3\xb3n modificada**:"},
			clases: obsoleta,
		},
		{
			nombre: "en-otra-linea",
			textos: []string{
				"El art\xc3\xadculo 21 regula la obligaci\xc3\xb3n de resolver.\n> \xe2\x9a\xa0 REDACCI\xc3\x93N MODIFICADA:",
			},
			clases: obsoleta,
		},
		{
			nombre: "repetida",
			textos: []string{"\xe2\x9a\xa0 REDACCI\xc3\x93N MODIFICADA:\n\xe2\x9a\xa0 REDACCI\xc3\x93N MODIFICADA:"},
			clases: obsoleta,
		},
		{nombre: "con-otras-palabras", textos: []string{"La redacci\xc3\xb3n ha cambiado desde la consulta anterior."}},
		{nombre: "palabra-de-mas", textos: []string{"\xe2\x9a\xa0 REDACCI\xc3\x93N YA MODIFICADA:"}},
		{nombre: "palabras-pegadas", textos: []string{"\xe2\x9a\xa0 REDACCI\xc3\x93NMODIFICADA:"}},
		{
			// Exacta en la etiqueta: sin la tilde o con otra palabra no es la forma.
			nombre: "palabra-distinta",
			textos: []string{"\xe2\x9a\xa0 REDACCION MODIFICADA:", "\xe2\x9a\xa0 REDACCI\xc3\x93N MODIFICADAS:"},
		},
		{nombre: "sin-marca", textos: []string{"REDACCI\xc3\x93N MODIFICADA:"}},
		{nombre: "sin-dos-puntos", textos: []string{"\xe2\x9a\xa0 REDACCI\xc3\x93N MODIFICADA"}},
		{
			nombre: "partida-en-dos-lineas",
			textos: []string{"\xe2\x9a\xa0 REDACCI\xc3\x93N\nMODIFICADA:", "\xe2\x9a\xa0\nREDACCI\xc3\x93N MODIFICADA:"},
		},
		{nombre: "forma-de-un-aviso", textos: []string{"\xe2\x9a\xa0 NORMA DEROGADA:"}},
		{nombre: "vacio", textos: []string{""}},
	}

	for _, caso := range casos {
		t.Run(caso.nombre, func(t *testing.T) {
			t.Parallel()

			for _, texto := range caso.textos {
				assert.Equal(t, caso.clases, ExtraerHallazgos(texto), "texto %q", texto)
			}
		})
	}

	// Cada etiqueta del binario, con la marca delante y los dos puntos detrás, da
	// su clase y solo la suya.
	t.Run("cada-etiqueta", func(t *testing.T) {
		t.Parallel()

		for clase, etiqueta := range grafo.EtiquetasDeHallazgo() {
			assert.Equal(t, []string{string(clase)}, ExtraerHallazgos(formaEscrita(etiqueta)), "clase %s", clase)
		}
	})
}

// TestComprobarFormasDeHallazgo fija la comprobación mecánica de FR-045 sobre
// textos sintéticos: con la forma fija de version-obsoleta no hay error, también
// con sus tolerancias; y sin ella —sin ninguna forma, con la de un aviso o con la
// etiqueta dicha con otras palabras—, el error exacto nombra version-obsoleta y
// su forma (SC-008; contrato evals-y-skill §2 y §3).
func TestComprobarFormasDeHallazgo(t *testing.T) {
	t.Parallel()

	const faltaObsoleta = "falta la forma fija del hallazgo version-obsoleta: \xe2\x9a\xa0 REDACCI\xc3\x93N MODIFICADA:"

	casos := []struct {
		nombre string
		texto  string

		// error es el error exacto; vacío, no hay error.
		error string
	}{
		{
			nombre: "con-la-forma",
			texto: "Trasl\xc3\xa1dalo con su forma fija:\n" +
				"`\xe2\x9a\xa0 REDACCI\xc3\x93N MODIFICADA: la redacci\xc3\xb3n con fecha de vigencia 20161002 ha sido " +
				"sustituida por la de 20250101.`\n",
		},
		{nombre: "con-tolerancias", texto: "**\xe2\x9a\xa0\xef\xb8\x8f Redacci\xc3\xb3n modificada**: \xe2\x80\xa6\n"},
		{nombre: "sin-ninguna-forma", texto: "Traslada lo que encuentre la memoria de consultas.\n", error: faltaObsoleta},
		{
			nombre: "con-la-de-un-aviso",
			texto:  "- `\xe2\x9a\xa0 NORMA DEROGADA:` para el aviso `derogada`.\n",
			error:  faltaObsoleta,
		},
		{nombre: "con-otras-palabras", texto: "Di que la redacci\xc3\xb3n ha cambiado.\n", error: faltaObsoleta},
		{nombre: "vacio", texto: "", error: faltaObsoleta},
	}

	for _, caso := range casos {
		t.Run(caso.nombre, func(t *testing.T) {
			t.Parallel()

			err := ComprobarFormasDeHallazgo(caso.texto)

			if caso.error == "" {
				assert.NoError(t, err)

				return
			}

			assert.EqualError(t, err, caso.error)
		})
	}
}

// TestComprobarClasesDeHallazgo fija la comprobación de hallazgos-del-esquema
// sobre esquemas sintéticos compilados con skills.CompilarEsquema: con el
// enumerado exacto detrás de un $ref, como en el esquema publicado, no hay error;
// con una clase de más, el error la nombra —también si es fuente-caducada, que no
// tiene etiqueta—; y con una de menos, sin items con enumerado, con dos $ref que
// se apuntan entre sí o sin la propiedad hallazgos, el error nombra
// version-obsoleta (FR-054; contrato evals-y-skill §1 y §3).
func TestComprobarClasesDeHallazgo(t *testing.T) {
	t.Parallel()

	const faltaObsoleta = "el esquema de eval no enumera la clase de hallazgo version-obsoleta en hallazgos"

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
    "hallazgos": { "type": "array", "minItems": 1, "items": { "$ref": "#/$defs/clase-de-hallazgo" } }
  },
  "$defs": {
    "clase-de-hallazgo": { "enum": ["version-obsoleta"] }
  }
}`,
		},
		{
			nombre: "una-de-mas",
			esquema: `{
  "$schema": "https://json-schema.org/draft/2020-12/schema",
  "type": "object",
  "properties": {
    "hallazgos": { "type": "array", "minItems": 1, "items": { "$ref": "#/$defs/clase-de-hallazgo" } }
  },
  "$defs": {
    "clase-de-hallazgo": { "enum": ["version-obsoleta", "fuente-caducada"] }
  }
}`,
			error: "el esquema de eval enumera en hallazgos la clase fuente-caducada, que no es una clase de hallazgo " +
				"etiquetada por el binario",
		},
		{
			nombre: "una-de-menos",
			esquema: `{
  "$schema": "https://json-schema.org/draft/2020-12/schema",
  "type": "object",
  "properties": {
    "hallazgos": { "type": "array", "minItems": 1, "items": { "$ref": "#/$defs/clase-de-hallazgo" } }
  },
  "$defs": {
    "clase-de-hallazgo": { "enum": ["fuente-caducada"] }
  }
}`,
			error: faltaObsoleta + "\n" +
				"el esquema de eval enumera en hallazgos la clase fuente-caducada, que no es una clase de hallazgo " +
				"etiquetada por el binario",
		},
		{
			nombre: "vacio",
			esquema: `{
  "$schema": "https://json-schema.org/draft/2020-12/schema",
  "type": "object",
  "properties": {
    "hallazgos": { "type": "array", "minItems": 1, "items": { "$ref": "#/$defs/clase-de-hallazgo" } }
  },
  "$defs": {
    "clase-de-hallazgo": { "enum": [] }
  }
}`,
			error: faltaObsoleta,
		},
		{
			nombre: "sin-enumerado",
			esquema: `{
  "$schema": "https://json-schema.org/draft/2020-12/schema",
  "type": "object",
  "properties": {
    "hallazgos": { "type": "array", "minItems": 1, "items": { "type": "string" } }
  }
}`,
			error: faltaObsoleta,
		},
		{
			nombre: "referencia-circular",
			esquema: `{
  "$schema": "https://json-schema.org/draft/2020-12/schema",
  "type": "object",
  "properties": {
    "hallazgos": { "type": "array", "minItems": 1, "items": { "$ref": "#/$defs/ida" } }
  },
  "$defs": {
    "ida": { "$ref": "#/$defs/vuelta" },
    "vuelta": { "$ref": "#/$defs/ida" }
  }
}`,
			error: faltaObsoleta,
		},
		{
			nombre: "sin-hallazgos",
			esquema: `{
  "$schema": "https://json-schema.org/draft/2020-12/schema",
  "type": "object",
  "properties": {
    "pregunta": { "type": "string", "minLength": 1 }
  },
  "$defs": {
    "clase-de-hallazgo": { "enum": ["version-obsoleta"] }
  }
}`,
			error: faltaObsoleta,
		},
	}

	for _, caso := range casos {
		t.Run(caso.nombre, func(t *testing.T) {
			t.Parallel()

			esquema, err := skills.CompilarEsquema([]byte(caso.esquema))
			require.NoError(t, err)

			err = ComprobarClasesDeHallazgo(esquema)

			if caso.error == "" {
				assert.NoError(t, err)

				return
			}

			assert.EqualError(t, err, caso.error)
		})
	}
}
