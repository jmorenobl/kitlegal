package evals

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

// listaDelContrato es, nueva en cada llamada, la lista de expresiones prohibidas
// de contracts/lista-y-juicio.md §2 escrita en el test, con la que se leen las
// filas de la tabla de §3.
func listaDelContrato() ExpresionesProhibidas {
	return ExpresionesProhibidas{
		Maquinaria: []string{
			"memoria de consultas",
			"hallazgo",
			"hallazgos",
			"graph check",
			"graph show",
			"graph stats",
			"kitlegal graph",
			"version-obsoleta",
			"fuente-caducada",
			"c\xc3\xb3digo de salida",
			"c\xc3\xb3digos de salida",
			"c\xc3\xb3digo 0",
			"c\xc3\xb3digo 1",
			"c\xc3\xb3digo 2",
			"c\xc3\xb3digo 3",
			"c\xc3\xb3digo 4",
			"c\xc3\xb3digo 5",
			"c\xc3\xb3digo 6",
			"c\xc3\xb3digo 7",
			"exit code",
			"json",
			"sobre de salida",
		},
		OtraConversacion: []string{
			"te dije",
			"te respond\xc3\xad",
			"te confirm\xc3\xa9",
			"te indiqu\xc3\xa9",
			"te coment\xc3\xa9",
			"te contest\xc3\xa9",
			"te expliqu\xc3\xa9",
			"te habr\xc3\xada dicho",
			"te habr\xc3\xada respondido",
			"te habr\xc3\xada confirmado",
			"te habr\xc3\xada indicado",
			"te habr\xc3\xada comentado",
			"te habr\xc3\xada contestado",
			"te habr\xc3\xada explicado",
			"conversaci\xc3\xb3n anterior",
			"conversaciones anteriores",
		},
	}
}

// TestExtraerExpresionesProhibidas fija la comparación de la lista de
// expresiones prohibidas con una respuesta (contrato lista-y-juicio §3; FR-051;
// research D3): cada fila de la tabla de §3 con la lista de §2 escrita en el
// test —la tolerancia de H5.1 a mayúsculas, blancos y énfasis de Markdown, los
// extremos de palabra, y las formas de los avisos y de los hallazgos, que no
// llevan ninguna—; que las encontradas van en el orden de la lista, la
// maquinaria y después lo dicho en otra conversación, y no en el del texto, sin
// repetir aunque el texto o la lista repitan una; y que cada expresión se
// compila una sola vez.
func TestExtraerExpresionesProhibidas(t *testing.T) {
	t.Parallel()

	delContrato := listaDelContrato()

	casos := []struct {
		nombre string
		texto  string
		lista  ExpresionesProhibidas

		// encontradas son las que tiene que devolver; nil, ninguna.
		encontradas []string
	}{
		{
			nombre:      "memoria-de-consultas-y-hallazgos",
			texto:       "Sin hallazgos en la memoria de consultas. Ya tengo todo lo necesario para responder.",
			lista:       delContrato,
			encontradas: []string{"memoria de consultas", "hallazgos"},
		},
		{
			nombre:      "mayusculas",
			texto:       "SIN HALLAZGOS",
			lista:       delContrato,
			encontradas: []string{"hallazgos"},
		},
		{
			nombre:      "dos-espacios-y-enfasis-alrededor",
			texto:       "Sin  hallazgos en la **memoria de consultas**",
			lista:       delContrato,
			encontradas: []string{"memoria de consultas", "hallazgos"},
		},
		{
			nombre:      "enfasis-dentro",
			texto:       "la *memoria* de _consultas_",
			lista:       delContrato,
			encontradas: []string{"memoria de consultas"},
		},
		{
			nombre:      "comillas-invertidas-y-parentesis",
			texto:       "Comprobado con `graph check` (c\xc3\xb3digo 0).",
			lista:       delContrato,
			encontradas: []string{"graph check", "c\xc3\xb3digo 0"},
		},
		{
			nombre:      "mayusculas-con-tilde",
			texto:       "Sali\xc3\xb3 con C\xc3\x93DIGO 0.",
			lista:       delContrato,
			encontradas: []string{"c\xc3\xb3digo 0"},
		},
		{
			nombre:      "te-habria-confirmado",
			texto:       "Este texto coincide con lo que te habr\xc3\xada confirmado antes",
			lista:       delContrato,
			encontradas: []string{"te habr\xc3\xada confirmado"},
		},
		{
			nombre:      "te-confirme-y-punto",
			texto:       "te confirm\xc3\xa9.",
			lista:       delContrato,
			encontradas: []string{"te confirm\xc3\xa9"},
		},
		{
			nombre: "redaccion-modificada",
			texto: "\xe2\x9a\xa0 REDACCI\xc3\x93N MODIFICADA: la redacci\xc3\xb3n con fecha de vigencia 20180309, " +
				"la que se consult\xc3\xb3 antes, ha sido sustituida por la de 20200206, que es la que se cita.",
			lista: delContrato,
		},
		{
			nombre: "norma-derogada",
			texto:  "\xe2\x9a\xa0 NORMA DEROGADA: esta norma ha sido derogada.",
			lista:  delContrato,
		},
		{
			nombre: "sin-avisos-de-vigencia",
			texto:  "No hay avisos de vigencia sobre este bloque.",
			lista:  delContrato,
		},
		{
			nombre: "sin-comprobar-la-redaccion",
			texto:  "No se ha podido comprobar si la redacci\xc3\xb3n ha cambiado desde una consulta anterior.",
			lista:  delContrato,
		},
		{
			nombre: "dentro-de-otra-palabra",
			texto:  "hallazgoss",
			lista:  delContrato,
		},
		{
			// El salto de línea no es un blanco tolerado entre dos palabras.
			nombre:      "salto-de-linea-entre-palabras",
			texto:       "Sin hallazgos en la memoria\nde consultas",
			lista:       delContrato,
			encontradas: []string{"hallazgos"},
		},
		{
			// En el texto, primero lo dicho en otra conversación y dos veces cada
			// una; en la lista, primero la maquinaria.
			nombre:      "orden-de-la-lista-y-no-del-texto",
			texto:       "Te dije que no hay hallazgos en la memoria de consultas; te dije que no hay hallazgos.",
			lista:       delContrato,
			encontradas: []string{"memoria de consultas", "hallazgos", "te dije"},
		},
		{
			nombre: "lista-con-repeticiones",
			texto:  "Los hallazgos, en json: te dije.",
			lista: ExpresionesProhibidas{
				Maquinaria:       []string{"json", "hallazgos", "json"},
				OtraConversacion: []string{"hallazgos", "te dije"},
			},
			encontradas: []string{"json", "hallazgos", "te dije"},
		},
		{
			nombre: "sin-lista",
			texto:  "Sin hallazgos en la memoria de consultas; te dije.",
		},
	}

	for _, caso := range casos {
		t.Run(caso.nombre, func(t *testing.T) {
			t.Parallel()

			assert.Equal(t, caso.encontradas, ExtraerExpresionesProhibidas(caso.texto, caso.lista))
		})
	}

	t.Run("compilada-una-sola-vez", func(t *testing.T) {
		t.Parallel()

		// Una expresión que no busca ningún otro test: la primera vez se compila.
		const expresion = "compilada una sola vez"

		primera := formasDeExpresiones.forma(expresion)
		assert.Same(t, primera, formasDeExpresiones.forma(expresion),
			"la segunda vez que se pide la forma de una expresión es la compilada la primera")
	})
}
