package boe

import (
	"encoding/json"
	"fmt"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/jmorenobl/kitlegal/internal/cli"
	"github.com/jmorenobl/kitlegal/internal/core/schema"
)

// TestLeerEnvoltorio fija la lectura del envoltorio de las respuestas JSON de
// buscar, indice, metadatos y analisis (data-model.md §3.1, research.md D7): el
// cuerpo es JSON en UTF-8 con un objeto como raíz (J1); data ausente o nula
// equivale a la lista vacía, y «vacío» es lista vacía, objeto vacío, cadena
// vacía o nulo (J2); y en metadatos y análisis cuenta el primer elemento de data,
// o data si es un objeto suelto, que tiene que ser un objeto (J5).
//
// Todo lo que no se puede leer es «fuente no disponible» con un mensaje literal
// que dice qué no se pudo interpretar y sin dirección ni instante, que pone quien
// pidió al envolverlo (contrato errores-y-codigos, filas 16 y 17): nunca el
// cuerpo ni un recorte suyo. Las entradas son JSON escrito en la tabla, sin
// ficheros de datos.
func TestLeerEnvoltorio(t *testing.T) {
	t.Parallel()

	t.Run("data del envoltorio", func(t *testing.T) {
		t.Parallel()

		casos := []struct {
			nombre string
			cuerpo string
			datos  string
		}{
			{
				nombre: "data como lista, la forma de la API",
				cuerpo: `{"status": {"code": "200", "text": "ok"}, "data": [{"identificador": "BOE-A-2015-10565"}]}`,
				datos:  `[{"identificador": "BOE-A-2015-10565"}]`,
			},
			{nombre: "data como objeto suelto", cuerpo: `{"data": {"titulo": "Ley 39/2015"}}`, datos: `{"titulo": "Ley 39/2015"}`},
			{
				nombre: "data como cadena vacía, la de la búsqueda sin resultados",
				cuerpo: `{"status": {"code": "200", "text": "ok"}, "data": ""}`,
				datos:  `""`,
			},
			{nombre: "sin data, la lista vacía", cuerpo: `{"status": {"code": "200", "text": "ok"}}`, datos: `[]`},
			{nombre: "data nula, la lista vacía", cuerpo: `{"data": null}`, datos: `[]`},
			{nombre: "espacio en blanco alrededor de la raíz", cuerpo: " \n\t{\"data\": []}\r\n ", datos: `[]`},
		}

		for _, caso := range casos {
			t.Run(caso.nombre, func(t *testing.T) {
				t.Parallel()

				datos, err := leerEnvoltorio([]byte(caso.cuerpo))

				require.NoError(t, err)
				assert.JSONEq(t, caso.datos, comoJSON(t, datos))
			})
		}
	})

	t.Run("los números llegan como su literal, sin pasar por la coma flotante", func(t *testing.T) {
		t.Parallel()

		datos, err := leerEnvoltorio([]byte(`{"data": [{"codigo": 12345678901234567890}]}`))

		require.NoError(t, err)
		assert.Equal(t, []any{map[string]any{"codigo": json.Number("12345678901234567890")}}, datos)
	})

	t.Run("lo que no se puede leer es fuente no disponible", func(t *testing.T) {
		t.Parallel()

		const (
			ilegible = "el cuerpo no es JSON legible"
			noUTF8   = "el cuerpo no es UTF-8 válido"
		)

		casos := []struct {
			nombre  string
			cuerpo  string
			mensaje string
		}{
			{nombre: "cuerpo vacío", cuerpo: "", mensaje: ilegible},
			{nombre: "solo espacio en blanco", cuerpo: " \n ", mensaje: ilegible},
			{
				nombre:  "XML, como el 404 de la API",
				cuerpo:  "<?xml version=\"1.0\" encoding=\"utf-8\"?>\n<response><data/></response>",
				mensaje: ilegible,
			},
			{nombre: "JSON cortado", cuerpo: `{"data": [`, mensaje: ilegible},
			{nombre: "otro valor detrás de la raíz", cuerpo: `{"data": []} {"data": []}`, mensaje: ilegible},
			{nombre: "un cierre sobrante detrás de la raíz", cuerpo: `{"data": []}]`, mensaje: ilegible},
			{nombre: "texto detrás de la raíz", cuerpo: `{"data": []} fin`, mensaje: ilegible},
			{nombre: "marca de orden de bytes delante", cuerpo: "\ufeff{\"data\": []}", mensaje: ilegible},
			{nombre: "un octeto que no es UTF-8 dentro de una cadena", cuerpo: "{\"data\": \"Ley \xff\"}", mensaje: noUTF8},
			{nombre: "raíz lista", cuerpo: `[{"data": []}]`, mensaje: "la raíz del JSON no es un objeto, sino una lista"},
			{nombre: "raíz cadena", cuerpo: `"data"`, mensaje: "la raíz del JSON no es un objeto, sino una cadena"},
			{nombre: "raíz número", cuerpo: `404`, mensaje: "la raíz del JSON no es un objeto, sino un número"},
			{nombre: "raíz booleano", cuerpo: `true`, mensaje: "la raíz del JSON no es un objeto, sino un booleano"},
			{nombre: "raíz nula", cuerpo: `null`, mensaje: "la raíz del JSON no es un objeto, sino nulo"},
		}

		for _, caso := range casos {
			t.Run(caso.nombre, func(t *testing.T) {
				t.Parallel()

				datos, err := leerEnvoltorio([]byte(caso.cuerpo))

				assert.Nil(t, datos)
				exigirIlegible(t, err, caso.mensaje)
			})
		}
	})

	t.Run("vacío", func(t *testing.T) {
		t.Parallel()

		casos := []struct {
			nombre string
			datos  string
			vacio  bool
		}{
			{nombre: "nulo", datos: `null`, vacio: true},
			{nombre: "lista vacía", datos: `[]`, vacio: true},
			{nombre: "objeto vacío", datos: `{}`, vacio: true},
			{nombre: "cadena vacía", datos: `""`, vacio: true},
			{nombre: "cadena con un espacio", datos: `" "`, vacio: false},
			{nombre: "lista con un nulo", datos: `[null]`, vacio: false},
			{nombre: "objeto con una clave nula", datos: `{"data": null}`, vacio: false},
			{nombre: "cero", datos: `0`, vacio: false},
			{nombre: "falso", datos: `false`, vacio: false},
		}

		for _, caso := range casos {
			t.Run(caso.nombre, func(t *testing.T) {
				t.Parallel()

				assert.Equal(t, caso.vacio, esVacio(valorJSON(t, caso.datos)))
			})
		}

		datos, err := leerEnvoltorio([]byte(`{"status": {"code": "200", "text": "ok"}}`))
		require.NoError(t, err)
		assert.True(t, esVacio(datos), "data ausente tiene que contar como vacía")
	})

	t.Run("primer elemento de data", func(t *testing.T) {
		t.Parallel()

		casos := []struct {
			nombre  string
			datos   string
			objeto  string
			mensaje string
		}{
			{nombre: "el primero de la lista", datos: `[{"titulo": "uno"}, {"titulo": "dos"}]`, objeto: `{"titulo": "uno"}`},
			{nombre: "el objeto suelto", datos: `{"titulo": "uno"}`, objeto: `{"titulo": "uno"}`},
			{nombre: "un primero sin claves sigue siendo un objeto", datos: `[{}]`, objeto: `{}`},
			{
				nombre:  "el primero no es objeto aunque el segundo lo sea",
				datos:   `["uno", {"titulo": "dos"}]`,
				mensaje: "el primer elemento de data no es un objeto, sino una cadena",
			},
			{
				nombre:  "el primero es nulo",
				datos:   `[null]`,
				mensaje: "el primer elemento de data no es un objeto, sino nulo",
			},
			{
				nombre:  "el primero es una lista",
				datos:   `[[{"titulo": "uno"}]]`,
				mensaje: "el primer elemento de data no es un objeto, sino una lista",
			},
			{
				nombre:  "data es una cadena",
				datos:   `"Ley 39/2015"`,
				mensaje: "el primer elemento de data no es un objeto, sino una cadena",
			},
			{nombre: "data es un número", datos: `4`, mensaje: "el primer elemento de data no es un objeto, sino un número"},
		}

		for _, caso := range casos {
			t.Run(caso.nombre, func(t *testing.T) {
				t.Parallel()

				objeto, err := primerElemento(valorJSON(t, caso.datos))

				if caso.mensaje != "" {
					assert.Nil(t, objeto)
					exigirIlegible(t, err, caso.mensaje)

					return
				}

				require.NoError(t, err)
				assert.JSONEq(t, caso.objeto, comoJSON(t, objeto))
			})
		}
	})
}

// TestComoLista fija dónde un objeto suelto se lee como lista de un elemento
// (J4, FR-070), los envoltorios (J6 y J7) y los elementos que no son objeto (J8),
// sobre JSON escrito en la tabla (data-model.md §3.1, research.md D7):
//
//   - comoLista es _ensure_list (refs/boe.py 218-225): nulo → lista vacía, lista
//     → tal cual, cualquier otro valor → lista de uno.
//   - Los resultados de buscar y los bloques del índice, anidado en el primer
//     elemento de data o plano (365-369), son listas de objetos.
//   - Las materias, las notas y las referencias anteriores y posteriores se
//     abren de su envoltorio —materia (531), nota (540), anterior y posterior
//     (550-554 y 570-574)— elemento a elemento, con el contenido también como
//     lista, y sin nada si están vacías (528, 539, 547 y 567). Una materia que no
//     es objeto se conserva porque aporta su texto (535); una referencia que no lo
//     es se salta (556 y 576).
//
// La salida se compara como JSON, que distingue la cadena "7" del número 7 y la
// lista vacía de nulo.
func TestComoLista(t *testing.T) {
	t.Parallel()

	type caso struct {
		nombre  string
		entrada string
		lista   string
	}

	lecturas := []struct {
		nombre string
		leer   func(t *testing.T, entrada string) any
		casos  []caso
	}{
		{
			nombre: "comoLista",
			leer: func(t *testing.T, entrada string) any {
				t.Helper()

				return comoLista(valorJSON(t, entrada))
			},
			casos: []caso{
				{nombre: "nulo, la lista vacía", entrada: `null`, lista: `[]`},
				{nombre: "lista, tal cual", entrada: `[{"id": "a1"}, "a2", 3, null]`, lista: `[{"id": "a1"}, "a2", 3, null]`},
				{nombre: "lista vacía, tal cual", entrada: `[]`, lista: `[]`},
				{nombre: "objeto suelto, lista de uno", entrada: `{"id": "a1"}`, lista: `[{"id": "a1"}]`},
				{nombre: "objeto sin claves, lista de uno", entrada: `{}`, lista: `[{}]`},
				{nombre: "cadena suelta, lista de una", entrada: `"a1"`, lista: `["a1"]`},
				{nombre: "número suelto, lista de uno", entrada: `7`, lista: `[7]`},
			},
		},
		{
			nombre: "resultados de buscar",
			leer: func(t *testing.T, entrada string) any {
				t.Helper()

				return resultadosDeBusqueda(valorJSON(t, entrada))
			},
			casos: []caso{
				{
					nombre:  "lista de resultados, en el orden de la fuente",
					entrada: `[{"identificador": "BOE-A-1992-26318"}, {"identificador": "BOE-A-2015-10565"}]`,
					lista:   `[{"identificador": "BOE-A-1992-26318"}, {"identificador": "BOE-A-2015-10565"}]`,
				},
				{
					nombre:  "un resultado suelto",
					entrada: `{"identificador": "BOE-A-2015-10565"}`,
					lista:   `[{"identificador": "BOE-A-2015-10565"}]`,
				},
				{
					nombre:  "los que no son objeto se saltan",
					entrada: `["BOE-A-1", {"identificador": "BOE-A-2"}, null, 3, true, [{"identificador": "BOE-A-3"}]]`,
					lista:   `[{"identificador": "BOE-A-2"}]`,
				},
				{nombre: "nulo, ninguno", entrada: `null`, lista: `[]`},
				{nombre: "la cadena vacía de la búsqueda sin resultados, ninguno", entrada: `""`, lista: `[]`},
			},
		},
		{
			nombre: "bloques del índice",
			leer: func(t *testing.T, entrada string) any {
				t.Helper()

				return bloquesDelIndice(valorJSON(t, entrada))
			},
			casos: []caso{
				{
					nombre:  "anidados en el primer elemento, la forma de la API",
					entrada: `[{"bloque": [{"id": "preambulo", "titulo": ""}, {"id": "a1", "titulo": "Artículo 1"}]}]`,
					lista:   `[{"id": "preambulo", "titulo": ""}, {"id": "a1", "titulo": "Artículo 1"}]`,
				},
				{nombre: "planos", entrada: `[{"id": "preambulo"}, {"id": "a1"}]`, lista: `[{"id": "preambulo"}, {"id": "a1"}]`},
				{nombre: "anidado un bloque suelto", entrada: `[{"bloque": {"id": "a1"}}]`, lista: `[{"id": "a1"}]`},
				{nombre: "el elemento que los anida, suelto", entrada: `{"bloque": [{"id": "a1"}, {"id": "a2"}]}`, lista: `[{"id": "a1"}, {"id": "a2"}]`},
				{nombre: "el elemento que anida, suelto, con un bloque suelto", entrada: `{"bloque": {"id": "a1"}}`, lista: `[{"id": "a1"}]`},
				{nombre: "plano, un bloque suelto", entrada: `{"id": "a1"}`, lista: `[{"id": "a1"}]`},
				{
					nombre:  "solo anida el primer elemento",
					entrada: `[{"id": "a1"}, {"bloque": [{"id": "a2"}]}]`,
					lista:   `[{"id": "a1"}, {"bloque": [{"id": "a2"}]}]`,
				},
				{
					nombre:  "si el primero no es objeto, son planos",
					entrada: `["a1", {"bloque": [{"id": "a2"}]}]`,
					lista:   `[{"bloque": [{"id": "a2"}]}]`,
				},
				{nombre: "los que no son objeto se saltan, también anidados", entrada: `[{"bloque": ["a1", {"id": "a2"}, null, 3]}]`, lista: `[{"id": "a2"}]`},
				{nombre: "anidados nulos, ninguno", entrada: `[{"bloque": null}]`, lista: `[]`},
				{nombre: "nulo, ninguno", entrada: `null`, lista: `[]`},
			},
		},
		{
			nombre: "materias del análisis",
			leer: func(t *testing.T, entrada string) any {
				t.Helper()

				return materiasDe(objetoJSON(t, entrada))
			},
			casos: []caso{
				{
					nombre: "envueltas, la forma de la API",
					entrada: `{"materias": [{"materia": {"codigo": "6499", "texto": "Seguridad Social"}},` +
						` {"materia": {"codigo": "4107", "texto": "Procedimiento administrativo"}}]}`,
					lista: `[{"codigo": "6499", "texto": "Seguridad Social"}, {"codigo": "4107", "texto": "Procedimiento administrativo"}]`,
				},
				{
					nombre:  "directas",
					entrada: `{"materias": [{"codigo": "6499", "texto": "Seguridad Social"}]}`,
					lista:   `[{"codigo": "6499", "texto": "Seguridad Social"}]`,
				},
				{
					nombre:  "una suelta y envuelta",
					entrada: `{"materias": {"materia": {"codigo": "6499", "texto": "Seguridad Social"}}}`,
					lista:   `[{"codigo": "6499", "texto": "Seguridad Social"}]`,
				},
				{
					nombre:  "un envoltorio con varias",
					entrada: `{"materias": [{"materia": [{"codigo": "6499"}, {"codigo": "4107"}]}]}`,
					lista:   `[{"codigo": "6499"}, {"codigo": "4107"}]`,
				},
				{
					nombre:  "las que no son objeto se conservan, porque aportan su texto",
					entrada: `{"materias": ["Seguridad Social", {"materia": "Procedimiento administrativo"}, 7]}`,
					lista:   `["Seguridad Social", "Procedimiento administrativo", 7]`,
				},
				{nombre: "un envoltorio nulo, ninguna", entrada: `{"materias": [{"materia": null}]}`, lista: `[]`},
				{nombre: "sin materias", entrada: `{}`, lista: `[]`},
				{nombre: "materias nulas", entrada: `{"materias": null}`, lista: `[]`},
				{nombre: "materias como lista vacía", entrada: `{"materias": []}`, lista: `[]`},
				{nombre: "materias como objeto vacío", entrada: `{"materias": {}}`, lista: `[]`},
				{nombre: "materias como cadena vacía", entrada: `{"materias": ""}`, lista: `[]`},
			},
		},
		{
			nombre: "notas del análisis",
			leer: func(t *testing.T, entrada string) any {
				t.Helper()

				return notasDe(objetoJSON(t, entrada))
			},
			casos: []caso{
				{
					nombre:  "envueltas dentro de una lista, la forma de la API",
					entrada: `{"notas": [{"nota": ["Entrada en vigor el 2 de octubre de 2016.", "Efectos: 2 de abril de 2021."]}]}`,
					lista:   `["Entrada en vigor el 2 de octubre de 2016.", "Efectos: 2 de abril de 2021."]`,
				},
				{nombre: "el envoltorio suelto con una cadena", entrada: `{"notas": {"nota": "Entrada en vigor"}}`, lista: `["Entrada en vigor"]`},
				{nombre: "el envoltorio suelto con una lista", entrada: `{"notas": {"nota": ["uno", "dos"]}}`, lista: `["uno", "dos"]`},
				{nombre: "una cadena directa, lista de una", entrada: `{"notas": "Entrada en vigor"}`, lista: `["Entrada en vigor"]`},
				{nombre: "una lista de cadenas directa, tal cual", entrada: `{"notas": ["uno", "dos"]}`, lista: `["uno", "dos"]`},
				{nombre: "sin notas", entrada: `{}`, lista: `[]`},
				{nombre: "notas nulas", entrada: `{"notas": null}`, lista: `[]`},
				{nombre: "notas como objeto vacío", entrada: `{"notas": {}}`, lista: `[]`},
				{nombre: "notas como cadena vacía", entrada: `{"notas": ""}`, lista: `[]`},
			},
		},
		{
			nombre: "referencias del análisis",
			leer: func(t *testing.T, entrada string) any {
				t.Helper()

				analisis := objetoJSON(t, entrada)

				return map[string]any{
					"anteriores":  referenciasAnterioresDe(analisis),
					"posteriores": referenciasPosterioresDe(analisis),
				}
			},
			casos: []caso{
				{
					nombre: "envueltas, la forma de la API",
					entrada: `{"referencias": {` +
						`"anteriores": [{"anterior": [{"id_norma": "BOE-A-2011-4117"}, {"id_norma": "BOE-A-2009-18358"}]}], ` +
						`"posteriores": [{"posterior": [{"id_norma": "BOE-A-2020-10491"}]}]}}`,
					lista: `{"anteriores": [{"id_norma": "BOE-A-2011-4117"}, {"id_norma": "BOE-A-2009-18358"}], ` +
						`"posteriores": [{"id_norma": "BOE-A-2020-10491"}]}`,
				},
				{
					nombre:  "directas",
					entrada: `{"referencias": {"anteriores": [{"id_norma": "BOE-A-2011-4117"}], "posteriores": [{"id_norma": "BOE-A-2020-10491"}]}}`,
					lista:   `{"anteriores": [{"id_norma": "BOE-A-2011-4117"}], "posteriores": [{"id_norma": "BOE-A-2020-10491"}]}`,
				},
				{
					nombre: "el envoltorio suelto con una referencia suelta",
					entrada: `{"referencias": {"anteriores": {"anterior": {"id_norma": "BOE-A-2011-4117"}}, ` +
						`"posteriores": {"posterior": {"id_norma": "BOE-A-2020-10491"}}}}`,
					lista: `{"anteriores": [{"id_norma": "BOE-A-2011-4117"}], "posteriores": [{"id_norma": "BOE-A-2020-10491"}]}`,
				},
				{
					nombre: "las que no son objeto se saltan, también dentro del envoltorio",
					entrada: `{"referencias": {"anteriores": ["BOE-A-1", {"anterior": ["BOE-A-2", {"id_norma": "BOE-A-3"}]}], ` +
						`"posteriores": [null, {"posterior": [7, {"id_norma": "BOE-A-4"}]}]}}`,
					lista: `{"anteriores": [{"id_norma": "BOE-A-3"}], "posteriores": [{"id_norma": "BOE-A-4"}]}`,
				},
				{
					nombre:  "el envoltorio de la otra lista no se abre",
					entrada: `{"referencias": {"anteriores": [{"posterior": [{"id_norma": "BOE-A-2011-4117"}]}]}}`,
					lista:   `{"anteriores": [{"posterior": [{"id_norma": "BOE-A-2011-4117"}]}], "posteriores": []}`,
				},
				{
					nombre:  "sin una de las dos listas",
					entrada: `{"referencias": {"posteriores": [{"id_norma": "BOE-A-2020-10491"}]}}`,
					lista:   `{"anteriores": [], "posteriores": [{"id_norma": "BOE-A-2020-10491"}]}`,
				},
				{nombre: "listas vacías", entrada: `{"referencias": {"anteriores": {}, "posteriores": ""}}`, lista: `{"anteriores": [], "posteriores": []}`},
				{
					nombre:  "referencias que no son un objeto, ninguna",
					entrada: `{"referencias": [{"anteriores": [{"id_norma": "BOE-A-2011-4117"}]}]}`,
					lista:   `{"anteriores": [], "posteriores": []}`,
				},
				{nombre: "sin referencias", entrada: `{}`, lista: `{"anteriores": [], "posteriores": []}`},
			},
		},
	}

	for _, lectura := range lecturas {
		t.Run(lectura.nombre, func(t *testing.T) {
			t.Parallel()

			for _, caso := range lectura.casos {
				t.Run(caso.nombre, func(t *testing.T) {
					t.Parallel()

					assert.JSONEq(t, caso.lista, comoJSON(t, lectura.leer(t, caso.entrada)))
				})
			}
		})
	}
}

// TestTextoDe fija la lectura de los campos de texto (J9, FR-016, research.md
// D7): una cadena es su valor; un campo ausente o nulo es la cadena vacía, nunca
// el marcador ? de refs/boe.py; y cualquier otro tipo —número, booleano, lista u
// objeto— es «fuente no disponible» con un mensaje literal que nombra el campo, y
// el subcampo si lo hay, sin dirección ni instante. rango, estado_consolidacion y
// relacion se leen de su objeto: en buscar, lo que no es objeto da la cadena
// vacía (refs/boe.py 288-291); en metadatos y análisis, la cadena es la cadena
// (466, 469, 559 y 579). Y envuelto por quien pidió, el mensaje nombra además la
// dirección (contrato errores-y-codigos, fila 17).
func TestTextoDe(t *testing.T) {
	t.Parallel()

	type caso struct {
		nombre   string
		entrada  string
		campo    string
		subcampo string
		texto    string
		mensaje  string
	}

	lecturas := []struct {
		nombre string
		leer   func(objeto map[string]any, campo, subcampo string) (string, error)
		casos  []caso
	}{
		{
			nombre: "textoDe",
			leer: func(objeto map[string]any, campo, _ string) (string, error) {
				return textoDe(objeto, campo)
			},
			casos: []caso{
				{nombre: "cadena, su valor", entrada: `{"titulo": "Ley 39/2015"}`, campo: "titulo", texto: "Ley 39/2015"},
				{nombre: "cadena vacía, vacía", entrada: `{"titulo": ""}`, campo: "titulo", texto: ""},
				{nombre: "la cadena ? es un valor como otro", entrada: `{"titulo": "?"}`, campo: "titulo", texto: "?"},
				{nombre: "ausente, vacío y no el marcador ?", entrada: `{"titulo": "Ley 39/2015"}`, campo: "url_eli", texto: ""},
				{nombre: "nulo, vacío", entrada: `{"url_eli": null}`, campo: "url_eli", texto: ""},
				{
					nombre:  "número, que el decodificador conserva como literal",
					entrada: `{"numero_oficial": 39}`,
					campo:   "numero_oficial",
					mensaje: `el campo "numero_oficial" no es texto, sino un número`,
				},
				{
					nombre:  "booleano",
					entrada: `{"vigencia_agotada": false}`,
					campo:   "vigencia_agotada",
					mensaje: `el campo "vigencia_agotada" no es texto, sino un booleano`,
				},
				{
					nombre:  "lista",
					entrada: `{"titulo": ["Ley 39/2015"]}`,
					campo:   "titulo",
					mensaje: `el campo "titulo" no es texto, sino una lista`,
				},
				{
					nombre:  "objeto",
					entrada: `{"titulo": {"texto": "Ley 39/2015"}}`,
					campo:   "titulo",
					mensaje: `el campo "titulo" no es texto, sino un objeto`,
				},
			},
		},
		{
			nombre: "texto del objeto, como en buscar",
			leer:   textoDelObjetoDe,
			casos: []caso{
				{
					nombre:  "objeto, su texto",
					entrada: `{"rango": {"codigo": "1300", "texto": "Ley"}}`,
					campo:   "rango", subcampo: "texto", texto: "Ley",
				},
				{
					nombre:  "el código del estado de consolidación",
					entrada: `{"estado_consolidacion": {"codigo": "3", "texto": "Finalizado"}}`,
					campo:   "estado_consolidacion", subcampo: "codigo", texto: "3",
				},
				{nombre: "objeto sin texto, vacío", entrada: `{"rango": {"codigo": "1300"}}`, campo: "rango", subcampo: "texto", texto: ""},
				{nombre: "objeto con texto nulo, vacío", entrada: `{"rango": {"texto": null}}`, campo: "rango", subcampo: "texto", texto: ""},
				{nombre: "cadena, vacío", entrada: `{"rango": "Ley"}`, campo: "rango", subcampo: "texto", texto: ""},
				{nombre: "número, vacío", entrada: `{"rango": 1300}`, campo: "rango", subcampo: "texto", texto: ""},
				{nombre: "lista, vacío", entrada: `{"rango": [{"texto": "Ley"}]}`, campo: "rango", subcampo: "texto", texto: ""},
				{nombre: "nulo, vacío", entrada: `{"rango": null}`, campo: "rango", subcampo: "texto", texto: ""},
				{nombre: "ausente, vacío", entrada: `{}`, campo: "rango", subcampo: "texto", texto: ""},
				{
					nombre:  "un texto del objeto que no es texto, nombrando campo y subcampo",
					entrada: `{"rango": {"texto": 1300}}`,
					campo:   "rango", subcampo: "texto",
					mensaje: `el campo "rango.texto" no es texto, sino un número`,
				},
			},
		},
		{
			nombre: "texto del objeto o la cadena, como en metadatos y análisis",
			leer:   textoDelObjetoOCadenaDe,
			casos: []caso{
				{
					nombre:  "objeto, su texto",
					entrada: `{"relacion": {"codigo": "210", "texto": "DEROGA"}}`,
					campo:   "relacion", subcampo: "texto", texto: "DEROGA",
				},
				{nombre: "objeto sin texto, vacío", entrada: `{"relacion": {"codigo": "210"}}`, campo: "relacion", subcampo: "texto", texto: ""},
				{nombre: "cadena, la cadena", entrada: `{"relacion": "DEROGA"}`, campo: "relacion", subcampo: "texto", texto: "DEROGA"},
				{nombre: "nulo, vacío", entrada: `{"relacion": null}`, campo: "relacion", subcampo: "texto", texto: ""},
				{nombre: "ausente, vacío", entrada: `{}`, campo: "rango", subcampo: "texto", texto: ""},
				{
					nombre:  "número, nombrando el campo",
					entrada: `{"relacion": 210}`,
					campo:   "relacion", subcampo: "texto",
					mensaje: `el campo "relacion" no es texto, sino un número`,
				},
				{
					nombre:  "lista, nombrando el campo",
					entrada: `{"rango": ["Ley"]}`,
					campo:   "rango", subcampo: "texto",
					mensaje: `el campo "rango" no es texto, sino una lista`,
				},
				{
					nombre:  "booleano, nombrando el campo",
					entrada: `{"estado_consolidacion": true}`,
					campo:   "estado_consolidacion", subcampo: "texto",
					mensaje: `el campo "estado_consolidacion" no es texto, sino un booleano`,
				},
				{
					nombre:  "un texto del objeto que no es texto, nombrando campo y subcampo",
					entrada: `{"relacion": {"texto": ["DEROGA"]}}`,
					campo:   "relacion", subcampo: "texto",
					mensaje: `el campo "relacion.texto" no es texto, sino una lista`,
				},
			},
		},
	}

	for _, lectura := range lecturas {
		t.Run(lectura.nombre, func(t *testing.T) {
			t.Parallel()

			for _, caso := range lectura.casos {
				t.Run(caso.nombre, func(t *testing.T) {
					t.Parallel()

					texto, err := lectura.leer(objetoJSON(t, caso.entrada), caso.campo, caso.subcampo)

					assert.Equal(t, caso.texto, texto)
					if caso.mensaje != "" {
						exigirIlegible(t, err, caso.mensaje)

						return
					}

					require.NoError(t, err)
				})
			}
		})
	}

	t.Run("un valor que no sale del decodificador también se nombra", func(t *testing.T) {
		t.Parallel()

		texto, err := textoDelValor(3.5, "codigo")

		assert.Empty(t, texto)
		exigirIlegible(t, err, `el campo "codigo" no es texto, sino un valor de tipo float64`)
	})

	t.Run("envuelto por quien pidió, el mensaje nombra el campo y la dirección", func(t *testing.T) {
		t.Parallel()

		const (
			direccion = "https://www.boe.es/datosabiertos/api/legislacion-consolidada/id/BOE-A-2015-10565/metadatos"
			contexto  = "no se pueden interpretar los metadatos"
		)

		instante := time.Date(2026, time.September, 13, 10, 30, 0, 123456789, time.UTC)

		_, deLaLectura := textoDe(objetoJSON(t, `{"titulo": 39}`), "titulo")
		require.Error(t, deLaLectura)

		fallo := errorDeFuenteNoDisponible(direccion, instante, deLaLectura, contexto)
		envuelto := fmt.Errorf("metadatos BOE-A-2015-10565: %w", fallo)

		assert.Equal(t, contexto+" ("+direccion+`): el campo "titulo" no es texto, sino un número`, fallo.Error())
		assert.Equal(t, schema.ClaseFuenteNoDisponible, cli.Clasificar(envuelto))

		var recuperado *Error
		require.ErrorAs(t, envuelto, &recuperado)
		assert.Same(t, fallo, recuperado)
		assert.Equal(t, direccion, recuperado.URL)
		assert.Equal(t, instante, recuperado.Instante)
	})
}

// exigirIlegible comprueba que el error sea el de una lectura que no se pudo
// hacer: un *Error de «fuente no disponible», con el mensaje literal esperado y
// sin dirección ni instante, que pone quien pidió al envolverlo.
func exigirIlegible(t *testing.T, err error, mensaje string) {
	t.Helper()

	var fallo *Error
	require.ErrorAs(t, err, &fallo)
	assert.Equal(t, mensaje, err.Error())
	assert.Equal(t, schema.ClaseFuenteNoDisponible, cli.Clasificar(err))
	assert.Empty(t, fallo.URL)
	assert.True(t, fallo.Instante.IsZero(), "la lectura no conoce el instante de la petición")
}

// valorJSON decodifica el JSON de una tabla como lo entrega la API a la lectura
// —con los números como json.Number—, sin pasar por leerEnvoltorio.
func valorJSON(t *testing.T, texto string) any {
	t.Helper()

	decodificador := json.NewDecoder(strings.NewReader(texto))
	decodificador.UseNumber()

	var valor any
	require.NoError(t, decodificador.Decode(&valor), "el JSON de la tabla no es válido: %s", texto)

	return valor
}

// objetoJSON es valorJSON para las tablas cuya entrada tiene que ser un objeto.
func objetoJSON(t *testing.T, texto string) map[string]any {
	t.Helper()

	objeto, esObjeto := valorJSON(t, texto).(map[string]any)
	require.True(t, esObjeto, "el JSON de la tabla no es un objeto: %s", texto)

	return objeto
}

// comoJSON codifica lo que devuelve una lectura para compararlo con el JSON de la
// tabla.
func comoJSON(t *testing.T, valor any) string {
	t.Helper()

	codificado, err := json.Marshal(valor)
	require.NoError(t, err)

	return string(codificado)
}
