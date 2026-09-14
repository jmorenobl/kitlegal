package boe

import (
	"maps"
	"slices"
	"strconv"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/jmorenobl/kitlegal/internal/cli"
	"github.com/jmorenobl/kitlegal/internal/core/schema"
	"github.com/jmorenobl/kitlegal/internal/render"
)

// Los punteros JSON, dentro del documento de --describe, de los dos esquemas de
// data que declaran un enumerado: el código de un aviso y el tipo de una entrada
// del índice (research.md D11).
const (
	punteroDelCodigoDeAviso = "/$defs/boe.Aviso/properties/codigo"
	punteroDelTipoDeBloque  = "/$defs/boe.EntradaDeIndice/properties/tipo"
)

// verboDescrito es un verbo de boe con el valor cero del tipo de su data, el que
// el applet declara como Salida (contrato puerto-y-applet §4).
type verboDescrito struct {
	verbo  string
	salida any
}

// verbosDescritos son los seis verbos con la Salida del contrato puerto-y-applet
// §4.
func verbosDescritos() []verboDescrito {
	return []verboDescrito{
		{verbo: "buscar", salida: []ResultadoDeBusqueda(nil)},
		{verbo: "indice", salida: Indice{}},
		{verbo: "articulo", salida: Articulo{}},
		{verbo: "articulos", salida: []Articulo(nil)},
		{verbo: "metadatos", salida: Metadatos{}},
		{verbo: "analisis", salida: Analisis{}},
	}
}

// TestEnumeradosDeLosDatos fija los enumerados del esquema de data (FR-012,
// research.md D11): el documento que genera --describe, con la biblioteca de
// esquemas fijada en go.mod y la configuración del kernel, declara en el data de
// cada verbo exactamente los enumerados de su tipo —el código de los avisos en
// articulo, articulos y metadatos, igual a CodigosDeAviso(); el tipo de las
// entradas en indice, igual a TiposDeBloque()— y ningún otro; y todo tipo que da
// TipoDesdeID está en TiposDeBloque(), que no enumera ninguno que no dé.
func TestEnumeradosDeLosDatos(t *testing.T) {
	t.Parallel()

	delKernel := enumeradosDe(t, documentoDescrito(t, "sin-datos", nil))
	require.NotEmpty(t, delKernel, "el documento sin data ya enumera las clases de error: sin ellas, "+
		"restarlas no demostraría nada")

	esperados := map[string]map[string][]string{
		"buscar":    {},
		"indice":    {punteroDelTipoDeBloque: TiposDeBloque()},
		"articulo":  {punteroDelCodigoDeAviso: CodigosDeAviso()},
		"articulos": {punteroDelCodigoDeAviso: CodigosDeAviso()},
		"metadatos": {punteroDelCodigoDeAviso: CodigosDeAviso()},
		"analisis":  {},
	}

	for _, descrito := range verbosDescritos() {
		t.Run(descrito.verbo, func(t *testing.T) {
			t.Parallel()

			propios := enumeradosDe(t, documentoDescrito(t, descrito.verbo, descrito.salida))
			for puntero, valores := range delKernel {
				require.Equalf(t, valores, propios[puntero], "el enumerado %s del kernel no puede cambiar con data",
					puntero)
				delete(propios, puntero)
			}

			assert.Equal(t, esperados[descrito.verbo], propios)
		})
	}

	t.Run("TiposDeBloque son los nueve tipos de TipoDesdeID en el orden de sus reglas y el vacío", func(t *testing.T) {
		t.Parallel()

		assert.Equal(t, []string{
			"articulo", "titulo", "capitulo", "seccion", "preambulo", "disposicion_adicional",
			"disposicion_transitoria", "disposicion_derogatoria", "disposicion_final", "",
		}, TiposDeBloque())
	})

	t.Run("todo tipo que da TipoDesdeID está en TiposDeBloque, y todo valor de TiposDeBloque lo da algún id", func(t *testing.T) {
		t.Parallel()

		// Un id, al menos, por cada una de las diez reglas de data-model.md §4, con
		// las rarezas portadas y los ids sin regla.
		ids := []string{
			"a21", "A108bis", "a1-30", "t1", "ti", "tp", "ci", "cv3", "cx", "s1", "se", "preambulo", "PREAMBULO",
			"da3", "da-3", "dt1", "dd", "df1", "subseccion", "sa", "a", "ab", "c", "c1", "d", "x", "",
		}

		dados := map[string]bool{}

		for _, id := range ids {
			tipo := TipoDesdeID(id)
			assert.Containsf(t, TiposDeBloque(), tipo, "TipoDesdeID(%q) = %q", id, tipo)

			dados[tipo] = true
		}

		assert.ElementsMatch(t, TiposDeBloque(), slices.Collect(maps.Keys(dados)))
	})

	t.Run("cada llamada a TiposDeBloque devuelve su propia lista", func(t *testing.T) {
		t.Parallel()

		tipos := TiposDeBloque()
		tipos[0] = "otro"

		assert.Equal(t, "articulo", TiposDeBloque()[0])
	})
}

// TestClavesDeLosDatos fija las claves del data de cada verbo (data-model.md §2,
// research.md D11): en el esquema que genera --describe, cada tipo de data
// declara exactamente sus claves, todas obligatorias y en su orden, sin admitir
// otras; y cada verbo describe exactamente los tipos de su data.
func TestClavesDeLosDatos(t *testing.T) {
	t.Parallel()

	clavesDe := map[string][]string{
		"boe.ResultadoDeBusqueda": {"identificador", "titulo", "rango", "vigencia_agotada", "estado_consolidacion", "url"},
		"boe.Indice":              {"norma", "url", "bloques"},
		"boe.EntradaDeIndice":     {"id", "titulo", "tipo"},
		"boe.Articulo": {
			"norma", "bloque", "titulo", "tipo", "fecha_version", "fecha_vigencia", "norma_modificadora", "texto",
			"hash_texto", "avisos", "url", "url_eli",
		},
		"boe.Aviso": {"codigo", "texto"},
		"boe.Metadatos": {
			"norma", "titulo", "rango", "numero_oficial", "fecha_disposicion", "fecha_publicacion", "fecha_vigencia",
			"estatus_derogacion", "vigencia_agotada", "estado_consolidacion", "url_eli", "avisos",
		},
		"boe.EstadoDeConsolidacion": {"codigo", "texto"},
		"boe.Analisis":              {"norma", "materias", "notas", "referencias"},
		"boe.Materia":               {"codigo", "texto"},
		"boe.Referencias":           {"anteriores", "posteriores"},
		"boe.ReferenciaAnterior":    {"relacion", "norma", "texto"},
		"boe.ReferenciaPosterior":   {"relacion", "norma"},
	}

	tiposDe := map[string][]string{
		"buscar":    {"boe.ResultadoDeBusqueda"},
		"indice":    {"boe.Indice", "boe.EntradaDeIndice"},
		"articulo":  {"boe.Articulo", "boe.Aviso"},
		"articulos": {"boe.Articulo", "boe.Aviso"},
		"metadatos": {"boe.Metadatos", "boe.EstadoDeConsolidacion", "boe.Aviso"},
		"analisis": {
			"boe.Analisis", "boe.Materia", "boe.Referencias", "boe.ReferenciaAnterior", "boe.ReferenciaPosterior",
		},
	}

	delKernel := definicionesDe(t, documentoDescrito(t, "sin-datos", nil))

	for _, descrito := range verbosDescritos() {
		t.Run(descrito.verbo, func(t *testing.T) {
			t.Parallel()

			definiciones := definicionesDe(t, documentoDescrito(t, descrito.verbo, descrito.salida))
			for nombre := range delKernel {
				delete(definiciones, nombre)
			}

			assert.ElementsMatch(t, tiposDe[descrito.verbo], slices.Collect(maps.Keys(definiciones)))

			for nombre, esquema := range definiciones {
				claves := clavesDe[nombre]
				propiedades, esObjeto := esquema["properties"].(map[string]any)
				require.Truef(t, esObjeto, "%s no declara sus propiedades", nombre)

				assert.ElementsMatchf(t, claves, slices.Collect(maps.Keys(propiedades)), "%s: claves", nombre)
				assert.Equalf(t, claves, cadenasDe(t, nombre+"/required", esquema["required"]),
					"%s: todas las claves obligatorias, en su orden", nombre)

				adicionales, esBooleano := esquema["additionalProperties"].(bool)
				assert.Truef(t, esBooleano && !adicionales, "%s admite claves que no declara", nombre)
			}
		})
	}
}

// TestHashTexto fija la huella del texto de un artículo (FR-015, SC-013,
// data-model.md §2.1): sha256: seguido del hexadecimal en minúsculas del SHA-256
// de los bytes UTF-8 del texto, que cambia si y solo si cambia el texto y cumple
// el patrón que el esquema declara para hash_texto, el mismo que el dominio da a
// toda huella.
func TestHashTexto(t *testing.T) {
	t.Parallel()

	t.Run("el SHA-256 de los bytes UTF-8 del texto, con el prefijo del algoritmo", func(t *testing.T) {
		t.Parallel()

		// Resúmenes calculados fuera de Go, con shasum -a 256 sobre los bytes
		// UTF-8 de cada texto.
		casos := []struct {
			nombre string
			texto  string
			huella string
		}{
			{
				nombre: "vacío",
				texto:  "",
				huella: "sha256:e3b0c44298fc1c149afbf4c8996fb92427ae41e4649b934ca495991b7852b855",
			},
			{
				nombre: "ASCII",
				texto:  "abc",
				huella: "sha256:ba7816bf8f01cfea414140de5dae2223b00361a396177a9cb410ff61f20015ad",
			},
			{
				nombre: "no ASCII, en sus bytes UTF-8",
				texto:  "Artículo 21",
				huella: "sha256:e33e5562bf43589d8605edbe82b4667cf94c933656c3f207873630081d0da56a",
			},
		}

		for _, caso := range casos {
			t.Run(caso.nombre, func(t *testing.T) {
				t.Parallel()

				assert.Equal(t, caso.huella, huellaDelTexto(caso.texto))
			})
		}
	})

	t.Run("cambia si y solo si cambia el texto", func(t *testing.T) {
		t.Parallel()

		const articulo = "Artículo 21. Obligación de resolver.\n" +
			"1. La Administración está obligada a dictar resolución expresa."

		textos := []string{
			articulo,
			strings.Join([]string{"Artículo 21. Obligación de resolver.", "1. La Administración está obligada " +
				"a dictar resolución expresa."}, "\n"),
			articulo + "\n",
			" " + articulo,
			strings.Replace(articulo, "expresa.", "expresa;", 1),
			strings.Replace(articulo, "Artículo", "artículo", 1),
			strings.Replace(articulo, "resolver.", "resolver. ", 1),
			strings.Replace(articulo, "Artículo", "Arti\u0301culo", 1),
			"",
			" ",
		}

		iguales := 0

		for i := range textos {
			for j := i + 1; j < len(textos); j++ {
				mismoTexto := textos[i] == textos[j]
				if mismoTexto {
					iguales++
				}

				assert.Equalf(t, mismoTexto, huellaDelTexto(textos[i]) == huellaDelTexto(textos[j]),
					"textos %q y %q", textos[i], textos[j])
			}
		}

		assert.Equal(t, 1, iguales, "la tabla lleva exactamente un par de textos iguales construidos por separado")
	})

	t.Run("cumple el patrón que el esquema declara para hash_texto, el del dominio", func(t *testing.T) {
		t.Parallel()

		documento := documentoDescrito(t, "articulo", Articulo{})
		definiciones := definicionesDe(t, documento)

		articulo, definido := definiciones["boe.Articulo"]
		require.True(t, definido, "el documento de articulo no define boe.Articulo")

		propiedades, esObjeto := articulo["properties"].(map[string]any)
		require.True(t, esObjeto, "boe.Articulo no declara sus propiedades")

		hashTexto, esObjeto := propiedades["hash_texto"].(map[string]any)
		require.True(t, esObjeto, "boe.Articulo no declara hash_texto")

		assert.Equal(t, schema.PatronHuella, hashTexto["pattern"])
		assert.Regexp(t, schema.PatronHuella, huellaDelTexto("Artículo 21"))
	})
}

// documentoDescrito es el documento que emite --describe para un verbo de boe
// cuyo data es del tipo de salida: el mismo camino que sigue el binario, con
// el presentador real.
func documentoDescrito(t *testing.T, verbo string, salida any) map[string]any {
	t.Helper()

	var emitido, errores strings.Builder

	presentador := render.Nuevo(&emitido, &errores)
	require.NoError(t, cli.Describir(presentador, cli.Verbo{Applet: "boe", Verbo: verbo, Salida: salida}))
	require.Empty(t, errores.String())

	return objetoJSON(t, emitido.String())
}

// definicionesDe son los $defs del documento, cada uno como objeto.
func definicionesDe(t *testing.T, documento map[string]any) map[string]map[string]any {
	t.Helper()

	crudas, esObjeto := documento["$defs"].(map[string]any)
	require.True(t, esObjeto, "el documento no lleva $defs")

	definiciones := make(map[string]map[string]any, len(crudas))

	for nombre, cruda := range crudas {
		esquema, esObjeto := cruda.(map[string]any)
		require.Truef(t, esObjeto, "la definición %s no es un objeto", nombre)

		definiciones[nombre] = esquema
	}

	return definiciones
}

// enumeradosDe explora el documento entero y devuelve cada enum que declara, por
// el puntero JSON del esquema que lo lleva.
func enumeradosDe(t *testing.T, documento map[string]any) map[string][]string {
	t.Helper()

	enumerados := map[string][]string{}
	recogerEnumerados(t, "", documento, enumerados)

	return enumerados
}

// recogerEnumerados es el recorrido de enumeradosDe desde el valor que está en
// el puntero.
func recogerEnumerados(t *testing.T, puntero string, valor any, enumerados map[string][]string) {
	t.Helper()

	switch valor := valor.(type) {
	case map[string]any:
		for clave, hijo := range valor {
			if clave == "enum" {
				enumerados[puntero] = cadenasDe(t, puntero+"/enum", hijo)

				continue
			}

			recogerEnumerados(t, puntero+"/"+tramoDePuntero(clave), hijo, enumerados)
		}
	case []any:
		for indice, hijo := range valor {
			recogerEnumerados(t, puntero+"/"+strconv.Itoa(indice), hijo, enumerados)
		}
	}
}

// tramoDePuntero escapa una clave como tramo de un puntero JSON (RFC 6901 §3).
func tramoDePuntero(clave string) string {
	return strings.NewReplacer("~", "~0", "/", "~1").Replace(clave)
}

// cadenasDe exige que el valor del documento sea una lista de cadenas y la
// devuelve.
func cadenasDe(t *testing.T, donde string, valor any) []string {
	t.Helper()

	lista, esLista := valor.([]any)
	require.Truef(t, esLista, "%s no es una lista, sino %T", donde, valor)

	cadenas := make([]string, 0, len(lista))

	for _, elemento := range lista {
		cadena, esCadena := elemento.(string)
		require.Truef(t, esCadena, "%s lleva %v, que no es una cadena", donde, elemento)

		cadenas = append(cadenas, cadena)
	}

	return cadenas
}
