package grafo_test

import (
	"bytes"
	"encoding/json"
	"reflect"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/jmorenobl/kitlegal/internal/core/grafo"
)

// comoElKernel escribe valor como lo escribe el kernel dentro del sobre, con
// encoding/json y sin escapar los caracteres de HTML (internal/render), sin el
// salto de línea final.
func comoElKernel(t *testing.T, valor any) string {
	t.Helper()

	var escrito bytes.Buffer

	codificador := json.NewEncoder(&escrito)
	codificador.SetEscapeHTML(false)
	require.NoError(t, codificador.Encode(valor))

	return string(bytes.TrimSuffix(escrito.Bytes(), []byte("\n")))
}

// TestSalidaDelGrafoEnJSON fija la forma JSON del data de los verbos de graph
// (data-model §5; contracts/applet-graph.md §3; FR-053, FR-054, FR-060; H7.1
// data-model §6 y FR-012): las claves en español y en el orden del contrato; las
// listas de show, de stats y de check y los datos de un nodo, nunca null,
// aunque lleguen nulos; las claves propias de cada clase de hallazgo, solo en la
// suya; la url tal cual, con sus «&»; y la instantánea que lee check, sin
// ninguna etiqueta JSON, porque no sale.
func TestSalidaDelGrafoEnJSON(t *testing.T) {
	t.Parallel()

	procedencia := grafo.Procedencia{Fuente: fuenteBOE, URL: urlA21 + "?a=1&b=2", FechaConsulta: lunesEnMadrid}
	enJSON := `{"fuente":"boe.legislacion-consolidada","url":"` + urlA21 + `?a=1&b=2",` +
		`"fecha_consulta":"2026-09-28T14:00:00+02:00"}`
	nodoSinDatos := grafo.NodoDeFicha{
		ID: idBloque, Tipo: grafo.TipoBloque, PrimeraObservacion: lunes, UltimaObservacion: procedencia,
	}
	nodoSinDatosEnJSON := `{"id":"eli/es/l/2015/10/01/39#a21","tipo":"Bloque","datos":{},` +
		`"primera_observacion":"2026-09-28T12:00:00Z","ultima_observacion":` + enJSON + `}`

	casos := []struct {
		nombre   string
		valor    any
		esperado string
	}{
		{
			"una ficha con los datos y las aristas nulos",
			grafo.Ficha{Nodo: nodoSinDatos},
			`{"nodo":` + nodoSinDatosEnJSON + `,"salientes":[],"entrantes":[]}`,
		},
		{
			"una ficha por su puntero",
			&grafo.Ficha{Nodo: nodoSinDatos},
			`{"nodo":` + nodoSinDatosEnJSON + `,"salientes":[],"entrantes":[]}`,
		},
		{
			"una ficha con datos y aristas",
			grafo.Ficha{
				Nodo: grafo.NodoDeFicha{
					ID: idBloque, Tipo: grafo.TipoBloque, Datos: map[string]any{grafo.DatoBloque: "a21"},
					PrimeraObservacion: lunes, UltimaObservacion: procedencia,
				},
				Salientes: []grafo.AristaDeFicha{{
					Relacion: grafo.RelacionTieneVersion, ID: idBloque + "@20161002",
					PrimeraObservacion: lunes, UltimaObservacion: procedencia,
				}},
				Entrantes: []grafo.AristaDeFicha{},
			},
			`{"nodo":{"id":"eli/es/l/2015/10/01/39#a21","tipo":"Bloque","datos":{"bloque":"a21"},` +
				`"primera_observacion":"2026-09-28T12:00:00Z","ultima_observacion":` + enJSON + `},` +
				`"salientes":[{"relacion":"eli:has_version","id":"eli/es/l/2015/10/01/39#a21@20161002",` +
				`"primera_observacion":"2026-09-28T12:00:00Z","ultima_observacion":` + enJSON + `}],"entrantes":[]}`,
		},
		{
			"un recuento vacio",
			grafo.Recuento{},
			`{"nodos":0,"aristas":0,"textos":0,"nodos_por_tipo":[],"aristas_por_relacion":[]}`,
		},
		{
			"un recuento con sus pares",
			grafo.Recuento{
				Nodos: 2, Aristas: 1, Textos: 1,
				NodosPorTipo: []grafo.RecuentoDeNodos{
					{Tipo: grafo.TipoBloque, Fuente: fuenteBOE, Nodos: 1},
					{Tipo: grafo.TipoNorma, Fuente: fuenteBOE, Nodos: 1},
				},
				AristasPorRelacion: []grafo.RecuentoDeAristas{
					{Relacion: grafo.RelacionTieneParte, Fuente: fuenteBOE, Aristas: 1},
				},
			},
			`{"nodos":2,"aristas":1,"textos":1,"nodos_por_tipo":[` +
				`{"tipo":"Bloque","fuente":"boe.legislacion-consolidada","nodos":1},` +
				`{"tipo":"Norma","fuente":"boe.legislacion-consolidada","nodos":1}],"aristas_por_relacion":[` +
				`{"relacion":"eli:has_part","fuente":"boe.legislacion-consolidada","aristas":1}]}`,
		},
		{
			"un hallazgo de fuente caducada",
			grafo.Hallazgo{
				Clase: grafo.ClaseFuenteCaducada, ID: idBloque, Explicacion: "La consulta caduco.",
				Procedencia: procedencia, VigenciaSegundos: 604800,
			},
			`{"clase":"fuente-caducada","id":"eli/es/l/2015/10/01/39#a21","explicacion":"La consulta caduco.",` +
				`"procedencia":` + enJSON + `,"vigencia_segundos":604800}`,
		},
		{
			"un hallazgo de version obsoleta",
			grafo.Hallazgo{
				Clase: grafo.ClaseVersionObsoleta, ID: idBloque + "@20161002", Explicacion: "La version esta superada.",
				Procedencia: procedencia, FechaVigencia: "20161002", FechaVigenciaReciente: "20250101",
			},
			`{"clase":"version-obsoleta","id":"eli/es/l/2015/10/01/39#a21@20161002",` +
				`"explicacion":"La version esta superada.","procedencia":` + enJSON + `,` +
				`"fecha_vigencia":"20161002","fecha_vigencia_reciente":"20250101"}`,
		},
		{
			"una comprobacion sin ambito ni hallazgos, con las listas nulas",
			grafo.Comprobacion{},
			`{"norma":"","bloques":[],"version-obsoleta":0,"fuente-caducada":0,"omitidos":0,"hallazgos":[]}`,
		},
		{
			"una comprobacion por su puntero",
			&grafo.Comprobacion{},
			`{"norma":"","bloques":[],"version-obsoleta":0,"fuente-caducada":0,"omitidos":0,"hallazgos":[]}`,
		},
		{
			"una comprobacion con su ambito, sus totales y un hallazgo",
			grafo.Comprobacion{
				Norma: identificadorLPAC, Bloques: []string{"a21", "a99"}, VersionObsoleta: 1, FuenteCaducada: 52,
				Omitidos: 3,
				Hallazgos: []grafo.Hallazgo{{
					Clase: grafo.ClaseFuenteCaducada, ID: idBloque, Explicacion: "La consulta caduco.",
					Procedencia: procedencia, VigenciaSegundos: 604800,
				}},
			},
			`{"norma":"BOE-A-2015-10565","bloques":["a21","a99"],"version-obsoleta":1,"fuente-caducada":52,` +
				`"omitidos":3,"hallazgos":[{"clase":"fuente-caducada","id":"eli/es/l/2015/10/01/39#a21",` +
				`"explicacion":"La consulta caduco.","procedencia":` + enJSON + `,"vigencia_segundos":604800}]}`,
		},
	}

	for _, caso := range casos {
		t.Run(caso.nombre, func(t *testing.T) {
			t.Parallel()

			assert.Equal(t, caso.esperado, comoElKernel(t, caso.valor))
		})
	}

	t.Run("la instantanea no sale", probarInstantaneaSinJSON)
}

// probarInstantaneaSinJSON fija que ni la instantánea ni sus nodos llevan
// etiquetas JSON: son lo que lee check, no una salida (data-model §5).
func probarInstantaneaSinJSON(t *testing.T) {
	t.Parallel()

	for _, tipo := range []reflect.Type{reflect.TypeFor[grafo.Instantanea](), reflect.TypeFor[grafo.NodoDeInstantanea]()} {
		for campo := range tipo.Fields() {
			assert.Empty(t, campo.Tag, "%s.%s", tipo.Name(), campo.Name)
		}
	}
}
