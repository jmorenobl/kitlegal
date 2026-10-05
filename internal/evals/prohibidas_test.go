package evals

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

// listaDelContrato es, nueva en cada llamada, la lista de expresiones prohibidas
// escrita en el test —la de contracts/lista-y-juicio.md §2 de H7.2 con la familia
// del anuncio de contracts/lista-de-expresiones.md §1 de H7.3—, con la que se
// leen las filas de la tabla de §3 de H7.2 y los casos del anuncio.
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
		Anuncio: []string{
			"que trasladar",
			"hace falta trasladar",
			"ya puedo responder",
			"con esto puedo responder",
			"y puedo responder",
			"tengo todo lo necesario",
			"tengo lo necesario",
			"redacto la respuesta",
			"respondo con el texto",
			"respondo con el contenido",
			"ya tengo la respuesta",
			"ya tengo el texto",
			"as\xc3\xad que respondo",
			"sin redacciones cambiadas",
		},
	}
}

// Las dos formas fijas de la lista de contracts/lista-de-expresiones.md §1 de
// H7.4: la línea con la que la skill traslada un cambio de redacción, con los
// marcadores de la cita y de las dos fechas, y la frase de la regla 7 sin su
// punto final.
const (
	formaDeLaRedaccionModificada = "\xe2\x9a\xa0 REDACCI\xc3\x93N MODIFICADA: <cita>: la redacci\xc3\xb3n con fecha de " +
		"vigencia <fecha>, la que se consult\xc3\xb3 antes, ha sido sustituida por la de <fecha>, que es la que se cita."
	formaDeLaRegla7 = "No se ha podido comprobar si la redacci\xc3\xb3n ha cambiado desde una consulta anterior"
)

// Lo que TestExtraerExpresionesProhibidas quita con las formas fijas
// (contracts/lista-de-expresiones.md §3 y §6 de H7.4).
const (
	// lineaDelArticulo118 es la línea de la redacción modificada con la cita de
	// su bloque, la que enseña v0.1.4, sin la marca ni la etiqueta.
	lineaDelArticulo118 = "art. 118 de la Ley 9/2017 [BOE-A-2017-12902, bloque a1-30]: la redacci\xc3\xb3n con " +
		"fecha de vigencia 20180309, la que se consult\xc3\xb3 antes, ha sido sustituida por la de 20200206, que es " +
		"la que se cita."

	// sinCitaDeLaLCSP es el resto de la línea sin cita, la de v0.1.1 a v0.1.3.
	sinCitaDeLaLCSP = "la redacci\xc3\xb3n con fecha de vigencia 20180309, la que se consult\xc3\xb3 antes, ha sido " +
		"sustituida por la de 20200206, que es la que se cita."

	// marcaYEtiqueta son la marca y la etiqueta de la línea, con sus dos puntos.
	marcaYEtiqueta = "\xe2\x9a\xa0 REDACCI\xc3\x93N MODIFICADA: "
)

// listaConFormasFijas es, nueva en cada llamada, la lista del contrato con lo
// que contracts/lista-de-expresiones.md §1 de H7.4 le da para quitar y buscar:
// las dos formas fijas, se consultó antes y consulta anterior en el anuncio, con
// las palabras de esas formas, y dos expresiones de la redacción no leída.
func listaConFormasFijas() ExpresionesProhibidas {
	lista := listaDelContrato()
	lista.Anuncio = append(lista.Anuncio, "consulta anterior", "se consult\xc3\xb3 antes")
	lista.RedaccionNoLeida = []string{"ya no exige", "se elimin\xc3\xb3"}
	lista.FormasFijas = []string{formaDeLaRedaccionModificada, formaDeLaRegla7}

	return lista
}

// TestExtraerExpresionesProhibidas fija la comparación de la lista de
// expresiones prohibidas con una respuesta (contrato lista-y-juicio §3 y FR-051
// de H7.2; contracts/lista-de-expresiones.md §3 y §6, FR-020 y FR-024 de H7.3;
// contracts/lista-de-expresiones.md §3 y §6, FR-012, FR-030 y FR-031 de H7.4;
// research D3 de H7.2 y D7 de H7.4): cada fila de la tabla de §3 de H7.2 con la
// lista escrita en el test —la tolerancia de H5.1 a mayúsculas, blancos y
// énfasis de Markdown, los extremos de palabra, y las formas de los avisos y de
// los hallazgos, que no llevan ninguna—; el anuncio de lo que viene, que la
// familia del anuncio encuentra, y el de que no hay avisos, que no encuentra;
// que las encontradas van en el orden de la lista, la maquinaria, lo dicho en
// otra conversación, el anuncio y la redacción no leída, y no en el del texto,
// sin repetir aunque el texto o la lista repitan una; que antes de buscar se
// quitan las formas fijas de la lista —la línea de la redacción modificada con
// su cita, con énfasis, con otra cita y sin ella; la frase de la regla 7 con su
// punto, sin él y seguida de un paréntesis— y la marca, la etiqueta y los dos
// puntos de cada aviso, y se busca lo que sigue en la línea y las mismas
// palabras fuera de esas formas; que una lista sin formas fijas no quita nada;
// y que cada expresión y cada forma fija se compilan una sola vez.
func TestExtraerExpresionesProhibidas(t *testing.T) {
	t.Parallel()

	delContrato := listaDelContrato()
	conFormasFijas := listaConFormasFijas()

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
			encontradas: []string{"memoria de consultas", "hallazgos", "tengo todo lo necesario"},
		},
		{
			nombre:      "nada-que-trasladar",
			texto:       "Nada que trasladar. Ya puedo responder.",
			lista:       delContrato,
			encontradas: []string{"que trasladar", "ya puedo responder"},
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
			// Decir que no hay avisos no es anunciar el estado de lo comprobado:
			// ninguna expresión del anuncio lo dice (FR-020 de H7.3 y FR-012 de
			// H7.4).
			nombre: "sin-avisos-de-vigencia",
			texto:  "No hay avisos de vigencia sobre este bloque.",
			lista:  conFormasFijas,
		},
		{
			// La frase de la regla 7 es una forma fija: consulta anterior no
			// cuenta dentro de ella.
			nombre: "sin-comprobar-la-redaccion",
			texto:  formaDeLaRegla7 + ".",
			lista:  conFormasFijas,
		},
		{
			nombre: "sin-comprobar-la-redaccion-sin-punto",
			texto:  "Lo leído es el texto vigente. " + formaDeLaRegla7,
			lista:  conFormasFijas,
		},
		{
			nombre: "sin-comprobar-la-redaccion-y-un-parentesis",
			texto:  formaDeLaRegla7 + " (la comprobaci\xc3\xb3n termin\xc3\xb3 con otro c\xc3\xb3digo).",
			lista:  conFormasFijas,
		},
		{
			nombre: "redaccion-modificada-con-cita",
			texto:  marcaYEtiqueta + lineaDelArticulo118,
			lista:  conFormasFijas,
		},
		{
			// El énfasis que rodea la marca y la etiqueta, el de la forma fija de
			// los avisos (H5.1), también en la etiqueta.
			nombre: "redaccion-modificada-con-enfasis",
			texto: "**\xe2\x9a\xa0 REDACCI\xc3\x93N MODIFICADA:** " + lineaDelArticulo118 + "\n" +
				"\xe2\x9a\xa0\xef\xb8\x8f **Redacci\xc3\xb3n modificada**: " + lineaDelArticulo118,
			lista: conFormasFijas,
		},
		{
			nombre: "redaccion-modificada-con-otra-cita",
			texto: marcaYEtiqueta + "disposici\xc3\xb3n adicional tercera de la LCSP [BOE-A-2017-12902, bloque " +
				"da-3]: la redacci\xc3\xb3n con fecha de vigencia 20180309, la que se consult\xc3\xb3 antes, ha sido " +
				"sustituida por la de 20230101, que es la que se cita.",
			lista: conFormasFijas,
		},
		{
			nombre: "redaccion-modificada-sin-cita",
			texto:  marcaYEtiqueta + sinCitaDeLaLCSP,
			lista:  conFormasFijas,
		},
		{
			// Lo que sigue a la forma en la línea se busca, y no se junta con lo
			// que la precede: la forma se cambia por un salto de línea.
			nombre:      "redaccion-modificada-y-lo-que-la-rodea",
			texto:       "Nada que " + marcaYEtiqueta + lineaDelArticulo118 + " trasladar. Ya puedo responder.",
			lista:       conFormasFijas,
			encontradas: []string{"ya puedo responder"},
		},
		{
			// La línea escrita con otras palabras no es la forma: se busca entera.
			nombre:      "redaccion-modificada-con-otras-palabras",
			texto:       marcaYEtiqueta + "la redacci\xc3\xb3n que se consult\xc3\xb3 antes ha sido sustituida.",
			lista:       conFormasFijas,
			encontradas: []string{"se consult\xc3\xb3 antes"},
		},
		{
			// Unas fechas de nueve cifras no son las de la forma.
			nombre: "redaccion-modificada-con-otras-fechas",
			texto: marcaYEtiqueta + "la redacci\xc3\xb3n con fecha de vigencia 201803091, la que se consult\xc3\xb3 " +
				"antes, ha sido sustituida por la de 202002061, que es la que se cita.",
			lista:       conFormasFijas,
			encontradas: []string{"se consult\xc3\xb3 antes"},
		},
		{
			nombre:      "la-que-se-consulto-antes-fuera-de-la-forma",
			texto:       "Esta redacci\xc3\xb3n es distinta de la que se consult\xc3\xb3 antes.",
			lista:       conFormasFijas,
			encontradas: []string{"se consult\xc3\xb3 antes"},
		},
		{
			nombre: "las-palabras-de-las-dos-formas-fuera-de-ellas",
			texto: "La redacci\xc3\xb3n es la misma que la que se consult\xc3\xb3 antes y no ha cambiado desde una " +
				"consulta anterior.",
			lista:       conFormasFijas,
			encontradas: []string{"consulta anterior", "se consult\xc3\xb3 antes"},
		},
		{
			nombre:      "etiqueta-de-aviso-y-una-expresion-en-su-linea",
			texto:       "\xe2\x9a\xa0 NORMA DEROGADA: nada que trasladar.",
			lista:       conFormasFijas,
			encontradas: []string{"que trasladar"},
		},
		{
			// La marca, la etiqueta y los dos puntos de un aviso son una forma
			// fija de la skill: sus palabras no cuentan dentro de ella.
			nombre: "etiqueta-de-aviso-quitada",
			texto:  "**\xe2\x9a\xa0 VIGENCIA AGOTADA:** la vigencia de esta norma est\xc3\xa1 agotada.",
			lista: ExpresionesProhibidas{
				Anuncio:     []string{"vigencia agotada"},
				FormasFijas: []string{formaDeLaRegla7},
			},
		},
		{
			nombre: "sin-formas-fijas",
			texto: marcaYEtiqueta + lineaDelArticulo118 + "\n" + formaDeLaRegla7 + ".\n" +
				"\xe2\x9a\xa0 VIGENCIA AGOTADA: la vigencia de esta norma est\xc3\xa1 agotada.",
			lista: ExpresionesProhibidas{
				Anuncio: []string{"consulta anterior", "se consult\xc3\xb3 antes", "vigencia agotada"},
			},
			encontradas: []string{"consulta anterior", "se consult\xc3\xb3 antes", "vigencia agotada"},
		},
		{
			// En el texto, primero la redacción no leída y al final la
			// maquinaria; en la lista, al revés.
			nombre:      "cuatro-familias",
			texto:       "Ya no exige nada. Ya puedo responder: como te dije, no hay hallazgos.",
			lista:       conFormasFijas,
			encontradas: []string{"hallazgos", "te dije", "ya puedo responder", "ya no exige"},
		},
		{
			// Sin familias no hay nada que buscar, tampoco con formas fijas.
			nombre: "solo-formas-fijas",
			texto:  "Ya no exige nada; ya puedo responder.",
			lista:  ExpresionesProhibidas{FormasFijas: conFormasFijas.FormasFijas},
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
			// En el texto, primero el anuncio, después lo dicho en otra
			// conversación y al final la maquinaria; en la lista, al revés.
			nombre:      "maquinaria-otra-conversacion-y-anuncio",
			texto:       "Ya puedo responder: como te dije, no hay hallazgos.",
			lista:       delContrato,
			encontradas: []string{"hallazgos", "te dije", "ya puedo responder"},
		},
		{
			nombre: "lista-con-repeticiones",
			texto:  "Los hallazgos, en json: te dije. Ya puedo responder.",
			lista: ExpresionesProhibidas{
				Maquinaria:       []string{"json", "hallazgos", "json"},
				OtraConversacion: []string{"hallazgos", "te dije"},
				Anuncio:          []string{"te dije", "ya puedo responder", "json", "ya puedo responder"},
			},
			encontradas: []string{"json", "hallazgos", "te dije", "ya puedo responder"},
		},
		{
			nombre: "sin-lista",
			texto:  "Sin hallazgos en la memoria de consultas; te dije. Ya puedo responder.",
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

	t.Run("forma-fija-compilada-una-sola-vez", func(t *testing.T) {
		t.Parallel()

		// Una forma fija que no quita ningún otro test: la primera vez se compila.
		const forma = "Compilada una sola vez el <fecha>"

		primera := formasDeLasFormasFijas.forma(forma)
		assert.Same(t, primera, formasDeLasFormasFijas.forma(forma),
			"la segunda vez que se pide la expresión de una forma fija es la compilada la primera")
	})
}
