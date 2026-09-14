package boe_test

import (
	"encoding/json"
	"fmt"
	"maps"
	"net/url"
	"regexp"
	"slices"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/jmorenobl/kitlegal/internal/cli"
	"github.com/jmorenobl/kitlegal/internal/core/schema"
	"github.com/jmorenobl/kitlegal/internal/httpx"
	"github.com/jmorenobl/kitlegal/internal/source/boe"
)

// Las formas esperadas que el mensaje de «argumentos» nombra para cada entrada
// fuera de su gramática (data-model.md §5, contrato errores-y-codigos §2, filas
// 2 y 3).
const (
	formaDeLaNorma = "BOE-A-<año>-<número>"
	formaDelBloque = "de 1 a 64 caracteres, el primero letra o dígito ASCII y los demás letras, " +
		"dígitos, guiones o puntos"
)

// TestValidarNorma fija la gramática de la norma, ^BOE-A-[0-9]{4}-[0-9]{1,9}$,
// sensible a mayúsculas: lo que casa es válido y todo lo demás —también lo que
// solo se diferencia en un salto de línea final o en dígitos que no son ASCII—
// es «argumentos», antes de abrir la caché y de pedir nada (FR-081, D9).
func TestValidarNorma(t *testing.T) {
	t.Parallel()

	casos := []struct {
		nombre string
		norma  string
		valida bool
	}{
		{nombre: "la ley 39 de 2015", norma: "BOE-A-2015-10565", valida: true},
		{nombre: "la Constitución", norma: "BOE-A-1978-31229", valida: true},
		{nombre: "número de un dígito", norma: "BOE-A-2015-1", valida: true},
		{nombre: "número de nueve dígitos", norma: "BOE-A-2015-123456789", valida: true},
		{nombre: "vacía", norma: ""},
		{nombre: "sin número", norma: "BOE-A-2015"},
		{nombre: "número vacío", norma: "BOE-A-2015-"},
		{nombre: "en minúsculas", norma: "boe-a-2015-10565"},
		{nombre: "otra sección del BOE", norma: "BOE-B-2015-10565"},
		{nombre: "año de dos dígitos", norma: "BOE-A-15-10565"},
		{nombre: "año de cinco dígitos", norma: "BOE-A-20155-10565"},
		{nombre: "número de diez dígitos", norma: "BOE-A-2015-1234567890"},
		{nombre: "dos guiones", norma: "BOE-A-2015--10565"},
		{nombre: "signo en el número", norma: "BOE-A-2015-+10565"},
		{nombre: "dígitos que no son ASCII", norma: "BOE-A-2015-١٠٥٦٥"},
		{nombre: "espacio delante", norma: " BOE-A-2015-10565"},
		{nombre: "espacio detrás", norma: "BOE-A-2015-10565 "},
		{nombre: "salto de línea detrás", norma: "BOE-A-2015-10565\n"},
		{nombre: "carácter nulo", norma: "BOE-A-2015-10565\x00"},
		{nombre: "separador de ruta", norma: "BOE-A-2015-10565/metadatos"},
		{nombre: "subida de ruta", norma: "../BOE-A-2015-10565"},
		{nombre: "interrogación", norma: "BOE-A-2015-10565?x=1"},
		{nombre: "almohadilla", norma: "BOE-A-2015-10565#a21"},
		{nombre: "porcentaje", norma: "BOE-A-2015-%31"},
	}

	for _, caso := range casos {
		t.Run(caso.nombre, func(t *testing.T) {
			t.Parallel()

			err := boe.ValidarNorma(caso.norma)
			if caso.valida {
				require.NoError(t, err)

				return
			}

			compruebaArgumentosInvalidos(t, err, caso.norma, formaDeLaNorma)
		})
	}
}

// TestValidarBloque fija la gramática del id de bloque,
// ^[A-Za-z0-9][A-Za-z0-9.-]{0,63}$: el primer carácter letra o dígito ASCII y
// los demás letras, dígitos, guiones o puntos, de uno a sesenta y cuatro. El
// guion y el punto entran porque son los que usa la fuente (a1-30 es el artículo
// 118 de la LCSP, a85bis. el 85 bis de la LRBRL) y son caracteres no reservados
// del RFC 3986 §2.3. Todo lo que podría alterar la petición en la que el id va
// como segmento —vacío, separadores de ruta, ?, #, %, espacios, controles,
// longitud desmedida, y el punto o los dos puntos al principio— es «argumentos»
// (FR-080, D9; la escalada del bloqueo 2 de gates/tarea-T009.md).
func TestValidarBloque(t *testing.T) {
	t.Parallel()

	casos := []struct {
		nombre string
		bloque string
		valido bool
	}{
		{nombre: "artículo", bloque: "a21", valido: true},
		{nombre: "disposición adicional", bloque: "da3", valido: true},
		{nombre: "disposición transitoria", bloque: "dt1", valido: true},
		{nombre: "preámbulo", bloque: "preambulo", valido: true},
		{nombre: "artículo bis", bloque: "a108bis", valido: true},
		{nombre: "mayúsculas", bloque: "A21", valido: true},
		{nombre: "un dígito", bloque: "1", valido: true},
		{nombre: "sesenta y cuatro caracteres", bloque: strings.Repeat("a", 64), valido: true},
		{nombre: "artículo de la LCSP con guion", bloque: "a1-30", valido: true},
		{nombre: "disposición adicional con guion", bloque: "da-3", valido: true},
		{nombre: "capítulo con guion", bloque: "ci-2", valido: true},
		{nombre: "sección con guion", bloque: "s1-2", valido: true},
		{nombre: "artículo bis con punto final", bloque: "a85bis.", valido: true},
		{nombre: "ordinal con guion", bloque: "primera-2", valido: true},
		{nombre: "vacío", bloque: ""},
		{nombre: "subida de ruta", bloque: "../a21"},
		{nombre: "dos puntos al principio", bloque: ".."},
		{nombre: "punto al principio", bloque: ".a21"},
		{nombre: "guion al principio", bloque: "-a21"},
		{nombre: "barra", bloque: "a21/x"},
		{nombre: "barra invertida", bloque: `a21\x`},
		{nombre: "punto", bloque: "."},
		{nombre: "interrogación", bloque: "a21?x"},
		{nombre: "almohadilla", bloque: "a21#x"},
		{nombre: "porcentaje", bloque: "%2e"},
		{nombre: "porcentaje que codifica una barra", bloque: "a21%2Fx"},
		{nombre: "espacio delante", bloque: " a21"},
		{nombre: "espacio detrás", bloque: "a21 "},
		{nombre: "espacio en medio", bloque: "a 21"},
		{nombre: "espacio sin salto de línea", bloque: "a21\u00a0"},
		{nombre: "salto de línea", bloque: "a21\n"},
		{nombre: "tabulador", bloque: "\ta21"},
		{nombre: "carácter nulo", bloque: "a21\x00"},
		{nombre: "carácter de borrado", bloque: "a21\x7f"},
		{nombre: "guion", bloque: "a-21", valido: true},
		{nombre: "guion bajo", bloque: "da_3"},
		{nombre: "letra que no es ASCII", bloque: "añadido"},
		{nombre: "dígitos de ancho completo", bloque: "a２１"},
		{nombre: "UTF-8 inválido", bloque: "a21\xff"},
		{nombre: "sesenta y cinco caracteres", bloque: strings.Repeat("a", 65)},
	}

	for _, caso := range casos {
		t.Run(caso.nombre, func(t *testing.T) {
			t.Parallel()

			err := boe.ValidarBloque(caso.bloque)
			if caso.valido {
				require.NoError(t, err)

				return
			}

			compruebaArgumentosInvalidos(t, err, caso.bloque, formaDelBloque)
		})
	}
}

// TestTipoDesdeID fija el porte de _tipo_from_id (refs/boe.py 155-176): las
// diez reglas sobre el id en minúsculas y en su orden, con las rarezas que se
// portan —ti es titulo, cv3 es capitulo, subseccion no la produce ninguna— y los
// tres ids de SC-007 (FR-040, D9, data-model.md §4).
func TestTipoDesdeID(t *testing.T) {
	t.Parallel()

	casos := []struct {
		regla int
		id    string
		tipo  string
	}{
		{regla: 1, id: "a21", tipo: "articulo"},
		{regla: 1, id: "A21", tipo: "articulo"},
		{regla: 1, id: "a٣", tipo: "articulo"},
		{regla: 2, id: "ti", tipo: "titulo"},
		{regla: 2, id: "t", tipo: "titulo"},
		{regla: 2, id: "tp", tipo: "titulo"},
		{regla: 3, id: "ci1", tipo: "capitulo"},
		{regla: 3, id: "cv3", tipo: "capitulo"},
		{regla: 3, id: "cx", tipo: "capitulo"},
		{regla: 3, id: "cm", tipo: "capitulo"},
		{regla: 4, id: "s1", tipo: "seccion"},
		{regla: 4, id: "se", tipo: "seccion"},
		{regla: 5, id: "preambulo", tipo: "preambulo"},
		{regla: 5, id: "PREAMBULO", tipo: "preambulo"},
		{regla: 6, id: "da3", tipo: "disposicion_adicional"},
		{regla: 7, id: "dt1", tipo: "disposicion_transitoria"},
		{regla: 8, id: "dd", tipo: "disposicion_derogatoria"},
		{regla: 9, id: "df1", tipo: "disposicion_final"},
		{regla: 10, id: "x1", tipo: ""},
		{regla: 10, id: "", tipo: ""},
		{regla: 10, id: "a", tipo: ""},
		{regla: 10, id: "ab", tipo: ""},
		{regla: 10, id: "c", tipo: ""},
		{regla: 10, id: "ca", tipo: ""},
		{regla: 10, id: "s", tipo: ""},
		{regla: 10, id: "subseccion", tipo: ""},
		{regla: 10, id: "preambulo1", tipo: ""},
		{regla: 10, id: "d1", tipo: ""},
	}

	for _, caso := range casos {
		t.Run(fmt.Sprintf("regla %d con %q", caso.regla, caso.id), func(t *testing.T) {
			t.Parallel()

			assert.Equal(t, caso.tipo, boe.TipoDesdeID(caso.id))
		})
	}
}

// FuzzIDDeBloque es el fuzz ligero de FR-082 y SC-007. Con las semillas corre en
// make test; el fuzz de verdad se lanza a mano (quickstart). Para cualquier
// entrada exige que nada entre en pánico, que el tipo inferido sea el mismo dos
// veces y uno de los diez posibles, y que ValidarBloque acepte exactamente lo que
// casa con la gramática; y para todo id aceptado, que sea seguro como segmento
// de la petición: nada que escapar, ninguno de / ? # %, no más de sesenta y
// cuatro bytes, y la dirección del bloque, analizada, termina exactamente en
// /<id> sin consulta ni fragmento (D9).
func FuzzIDDeBloque(f *testing.F) {
	for _, semilla := range []string{"a21", "da3", "dt1", "a1-30", "a85bis."} {
		f.Add(semilla)
	}

	const (
		sitio             = "www.boe.es"
		rutaHastaElBloque = "/datosabiertos/api/legislacion-consolidada/id/BOE-A-2015-10565/texto/bloque/"
	)

	gramatica := regexp.MustCompile(`^[A-Za-z0-9][A-Za-z0-9.-]{0,63}$`)
	tipos := []string{
		"articulo", "titulo", "capitulo", "seccion", "preambulo", "disposicion_adicional",
		"disposicion_transitoria", "disposicion_derogatoria", "disposicion_final", "",
	}

	f.Fuzz(func(t *testing.T, id string) {
		tipo := boe.TipoDesdeID(id)
		assert.Equal(t, tipo, boe.TipoDesdeID(id), "el tipo inferido de %q no es determinista", id)
		assert.Contains(t, tipos, tipo)

		err := boe.ValidarBloque(id)
		if !gramatica.MatchString(id) {
			compruebaArgumentosInvalidos(t, err, id, formaDelBloque)

			return
		}

		require.NoError(t, err, "ValidarBloque rechaza %q, que casa con la gramática", id)

		assert.Equal(t, id, url.PathEscape(id))
		assert.False(t, strings.ContainsAny(id, "/?#%"), "el id %q lleva un carácter de la petición", id)
		assert.LessOrEqual(t, len(id), 64)

		direccion, err := url.Parse("https://" + sitio + rutaHastaElBloque + id)
		require.NoError(t, err)
		assert.Equal(t, sitio, direccion.Host)
		assert.Equal(t, rutaHastaElBloque+id, direccion.Path)
		assert.True(t, strings.HasSuffix(direccion.EscapedPath(), "/"+id))
		assert.Empty(t, direccion.RawPath)
		assert.Empty(t, direccion.RawQuery)
		assert.Empty(t, direccion.Fragment)
	})
}

// compruebaArgumentosInvalidos exige lo que el contrato errores-y-codigos fija
// para una entrada fuera de su gramática (filas 2 y 3): clase «argumentos» y
// código 2, un mensaje que nombra el valor recibido y la forma esperada, y un
// *boe.Error sin dirección, sin instante y sin causa, porque no se ha
// construido ninguna petición.
func compruebaArgumentosInvalidos(t *testing.T, err error, valor, forma string) {
	t.Helper()

	require.Error(t, err)
	assert.Equal(t, schema.ClaseArgumentos, cli.Clasificar(err))
	assert.Equal(t, 2, cli.CodigoSalida(err))
	assert.Contains(t, err.Error(), fmt.Sprintf("%q", valor))
	assert.Contains(t, err.Error(), forma)

	var fallo *boe.Error
	require.ErrorAs(t, err, &fallo)
	assert.Empty(t, fallo.URL)
	assert.True(t, fallo.Instante.IsZero())
	require.NoError(t, fallo.Causa)
}

// Las grabaciones de índice a las que se ata la gramática del id de bloque: las
// de los recursos 3 a 5 del manifiesto (contrato esquemas-fixtures-y-controles
// §3.1), reproducidas desde la carpeta de grabaciones de la fuente. La dirección
// de cada índice es la que TestDirecciones fija byte a byte para
// direccionDelIndice.
const (
	carpetaDeGrabacionesDeLaFuente = "testdata/" + boe.NombreDeLaFuente
	direccionDeUnIndice            = "https://www.boe.es/datosabiertos/api/legislacion-consolidada/id/%s/texto/indice"
	estadoDeUnIndiceGrabado        = 200
	claveDelIDDeBloque             = "id"
)

// TestGramaticaCubreLosIndicesGrabados ata la gramática de ValidarBloque a los ids
// de bloque que usa la fuente (FR-080; research.md D9; data-model.md §5): lee con
// el decodificador JSON genérico, sin el código de lectura de la fuente, el cuerpo
// de las tres grabaciones de índice que sirve httpx.Replay y exige que todo id de
// bloque que traen case con la gramática, nombrando los que no. Si alguno no casa,
// la gramática se revisa dentro de FR-080; si hiciera falta un carácter que FR-080
// no admite, es un conflicto con el spec y se escala.
func TestGramaticaCubreLosIndicesGrabados(t *testing.T) {
	t.Parallel()

	cliente, err := httpx.Replay(carpetaDeGrabacionesDeLaFuente, httpx.ConFuente(boe.NombreDeLaFuente))
	require.NoError(t, err)

	for _, norma := range []string{"BOE-A-2015-10565", "BOE-A-1985-5392", "BOE-A-2017-12902"} {
		t.Run(norma, func(t *testing.T) {
			t.Parallel()

			respuesta, err := cliente.Pedir(t.Context(), schema.Contexto{}, httpx.Peticion{
				Metodo: "GET",
				URL:    fmt.Sprintf(direccionDeUnIndice, norma),
				Acepta: "application/json",
			})
			require.NoError(t, err)
			require.Equal(t, estadoDeUnIndiceGrabado, respuesta.Estado)

			var cuerpo any
			require.NoError(t, json.Unmarshal(respuesta.Cuerpo, &cuerpo))

			raiz, esObjeto := cuerpo.(map[string]any)
			require.Truef(t, esObjeto, "el cuerpo del índice grabado de %s no es un objeto JSON", norma)

			ids, err := idsDeBloque(raiz["data"])
			require.NoError(t, err)
			require.NotEmptyf(t, ids, "el índice grabado de %s no trae ningún id de bloque", norma)

			var fuera []string

			for _, id := range ids {
				if boe.ValidarBloque(id) != nil {
					fuera = append(fuera, id)
				}
			}

			assert.Emptyf(t, fuera, "%d de los %d ids de bloque del índice grabado de %s no casan con la gramática "+
				"de ValidarBloque: %s", len(fuera), len(ids), norma, strings.Join(fuera, " "))
		})
	}
}

// idsDeBloque devuelve, en su orden y con las claves de cada objeto en orden
// alfabético, el valor de cada clave id que hay a cualquier profundidad bajo data
// en la respuesta de un índice, donde todo id es el de un bloque, venga el índice
// anidado o plano (contrato verbos-y-salidas §2; data-model.md §3.1, J6). Un id
// que no es una cadena es un error que lo nombra.
func idsDeBloque(valor any) ([]string, error) {
	var ids []string

	switch nodo := valor.(type) {
	case []any:
		for _, elemento := range nodo {
			delElemento, err := idsDeBloque(elemento)
			if err != nil {
				return nil, err
			}

			ids = append(ids, delElemento...)
		}

	case map[string]any:
		for _, clave := range slices.Sorted(maps.Keys(nodo)) {
			if clave != claveDelIDDeBloque {
				delValor, err := idsDeBloque(nodo[clave])
				if err != nil {
					return nil, err
				}

				ids = append(ids, delValor...)

				continue
			}

			id, esCadena := nodo[clave].(string)
			if !esCadena {
				return nil, fmt.Errorf("un id de bloque del índice no es una cadena: %v", nodo[clave])
			}

			ids = append(ids, id)
		}
	}

	return ids, nil
}
