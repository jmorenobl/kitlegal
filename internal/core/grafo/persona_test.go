package grafo_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/jmorenobl/kitlegal/internal/core/grafo"
	"github.com/jmorenobl/kitlegal/internal/core/schema"
)

// Los motivos con los que una Persona no entra en el grafo del mundo (FR-025,
// constitución VII).
const (
	motivoIDConDocumento    = "su id tiene la forma de un DNI, un NIE o un NIF"
	motivoDatosConDocumento = "sus datos llevan una clave o un valor con la forma de un DNI, un NIE o un NIF"
)

// cadenaPropia es un tipo de cadena que no es string: sus valores se examinan
// igual, porque en los datos guardados son cadenas.
type cadenaPropia string

// documentos son los 17 valores ASCII que rechazan (H7.1 FR-074; H7 FR-025 y
// SC-010, que pide al menos 12): las tres formas —DNI, NIE y NIF de persona
// jurídica— con y sin separadores, en mayúsculas y en minúsculas, dentro de
// un texto y en la fracción de segundo de una fecha.
var documentos = []struct {
	nombre string
	valor  string
}{
	{"DNI", "12345678Z"},
	{"DNI con puntos y guion", "12.345.678-Z"},
	{"DNI con espacios y la letra en minuscula", "12 345 678 z"},
	{"DNI con espacios", "12 345 678 Z"},
	{"DNI con guion", "12345678-Z"},
	{"NIE en minusculas", "x1234567l"},
	{"NIE con guiones y puntos", "X-1.234.567-L"},
	{"NIE con guiones", "X-1234567-L"},
	{"NIE con espacios en minusculas", "y 1234567 l"},
	{"NIF de persona juridica", "B12345678"},
	{"NIF en minuscula con guion y puntos", "b-12.345.678"},
	{"NIF con guion y puntos", "B-12.345.678"},
	{"NIF con letra de control", "B1234567J"},
	{"NIF con espacios en minusculas", "b 1234567 j"},
	{"DNI detras de un nombre", "Ana 12345678Z"},
	{"fraccion de seis cifras de una fecha", "2026-09-28T12:00:00.123456Z"},
	{"fraccion de ocho cifras de una fecha", "2026-09-28T12:00:00.12345678Z"},
}

// sinDocumento son los 6 valores ASCII que entran (H7.1 FR-074; H7 FR-025 y
// SC-010, que pide al menos 6): un nombre, fechas, instantes sin fracción de
// segundo y una letra con siete cifras.
var sinDocumento = []struct {
	nombre string
	valor  string
}{
	{"nombre con guiones", "ana-garcia-lopez"},
	{"fecha", "1990-01-01"},
	{"dos fechas", "1990-01-01 y 2000-02-02"},
	{"instante en UTC", "2026-09-28T12:00:00Z"},
	{"instante con desplazamiento", "2026-09-28T12:00:00+02:00"},
	{"una letra y siete cifras", "A1234567"},
}

// posicion es un lugar de una Persona en el que se examina una cadena, con el
// motivo del rechazo si lleva un documento.
type posicion struct {
	nombre string
	nodo   func(valor string) schema.Nodo
	motivo string
}

// posiciones son los cuatro lugares que fija contracts/almacen-world-db.md §5:
// el id, un valor de primer nivel, dentro de una lista de un objeto anidado
// —detrás de otra cadena, para que se vea que se examina toda la lista— y una
// clave.
var posiciones = []posicion{
	{
		nombre: "en el id",
		nodo: func(valor string) schema.Nodo {
			return schema.Nodo{ID: valor, Tipo: grafo.TipoPersona, Datos: map[string]any{"nombre": "Ana"}}
		},
		motivo: motivoIDConDocumento,
	},
	{
		nombre: "en un valor de primer nivel",
		nodo: func(valor string) schema.Nodo {
			return schema.Nodo{ID: "persona:ana", Tipo: grafo.TipoPersona, Datos: map[string]any{"nota": valor}}
		},
		motivo: motivoDatosConDocumento,
	},
	{
		nombre: "en una lista de un objeto anidado",
		nodo: func(valor string) schema.Nodo {
			return schema.Nodo{ID: "persona:ana", Tipo: grafo.TipoPersona, Datos: map[string]any{
				"contacto": map[string]any{"documentos": []any{"ninguno", valor}},
			}}
		},
		motivo: motivoDatosConDocumento,
	},
	{
		nombre: "como clave",
		nodo: func(valor string) schema.Nodo {
			return schema.Nodo{ID: "persona:ana", Tipo: grafo.TipoPersona, Datos: map[string]any{valor: "nombre"}}
		},
		motivo: motivoDatosConDocumento,
	},
}

// TestPersonaSinDocumento fija el rechazo de una Persona con un documento de
// identidad (FR-025, SC-010; H7.1 FR-074; contracts/almacen-world-db.md §5):
// los 17 valores que rechazan y los 6 que entran, cada uno en el id, en un
// valor de primer nivel, dentro de una lista de un objeto anidado y como
// clave. El mensaje nombra la Persona por su tipo y nunca repite el documento.
func TestPersonaSinDocumento(t *testing.T) {
	t.Parallel()

	t.Run("rechazan", probarDocumentosQueRechazan)
	t.Run("entran", probarCadenasQueEntran)
	t.Run("toda cadena de los datos", probarCadenasDeLosDatos)
}

func probarDocumentosQueRechazan(t *testing.T) {
	t.Parallel()

	for _, documento := range documentos {
		for _, lugar := range posiciones {
			t.Run(documento.nombre+" "+lugar.nombre, func(t *testing.T) {
				t.Parallel()

				nodo := lugar.nodo(documento.valor)
				exigirPersonaRechazada(t, nodo, lugar.motivo, documento.valor)
			})
		}
	}
}

func probarCadenasQueEntran(t *testing.T) {
	t.Parallel()

	for _, cadena := range sinDocumento {
		for _, lugar := range posiciones {
			t.Run(cadena.nombre+" "+lugar.nombre, func(t *testing.T) {
				t.Parallel()

				lote := loteDeEjemplo()
				conOperacion(lugar.nodo(cadena.valor))(&lote)

				require.NoError(t, grafo.ValidarLote(lote))
			})
		}
	}
}

// probarCadenasDeLosDatos fija que se examina toda cadena de los datos tal
// como se guarda, en su forma JSON: a cualquier profundidad de listas y
// objetos y con cualquier tipo de Go que sea una lista, un objeto o una
// cadena; y que los números, los booleanos y los nulos no se examinan.
func probarCadenasDeLosDatos(t *testing.T) {
	t.Parallel()

	casos := []struct {
		nombre string
		datos  map[string]any
		motivo string
	}{
		{"una lista de Go de cadenas", map[string]any{"documentos": []string{"12345678Z"}}, motivoDatosConDocumento},
		{
			"un mapa de Go de cadenas como valor",
			map[string]any{"contacto": map[string]string{"dni": "12345678Z"}},
			motivoDatosConDocumento,
		},
		{
			"el documento como clave de un mapa de Go de numeros",
			map[string]any{"contacto": map[string]int{"12345678Z": 1}},
			motivoDatosConDocumento,
		},
		{"un tipo propio de cadena", map[string]any{"dni": cadenaPropia("12345678Z")}, motivoDatosConDocumento},
		{
			"una lista dentro de una lista",
			map[string]any{"a": []any{"b", []any{"c", []any{"12345678Z"}}}},
			motivoDatosConDocumento,
		},
		{
			"una clave tres niveles abajo",
			map[string]any{"a": map[string]any{"b": map[string]any{"12345678Z": nil}}},
			motivoDatosConDocumento,
		},
		{
			"un documento en medio de un texto",
			map[string]any{"nota": "Ana, con DNI 12.345.678-Z, vecina de Leganes"},
			motivoDatosConDocumento,
		},
		{"numeros, booleanos y nulos", map[string]any{"n": 12345678, "x": 1.5, "b": true, "z": nil}, ""},
		{"sin datos", nil, ""},
		{"datos vacios", map[string]any{}, ""},
	}

	for _, caso := range casos {
		t.Run(caso.nombre, func(t *testing.T) {
			t.Parallel()

			nodo := schema.Nodo{ID: "persona:ana", Tipo: grafo.TipoPersona, Datos: caso.datos}
			if caso.motivo == "" {
				lote := loteDeEjemplo()
				conOperacion(nodo)(&lote)
				require.NoError(t, grafo.ValidarLote(lote))

				return
			}

			exigirPersonaRechazada(t, nodo, caso.motivo, "12345678")
		})
	}
}

// exigirPersonaRechazada comprueba que la Persona rechaza el lote de ejemplo
// entero con el motivo, nombrada por su tipo, y que el mensaje no repite el
// documento.
func exigirPersonaRechazada(t *testing.T, persona schema.Nodo, motivo, documento string) {
	t.Helper()

	lote := loteDeEjemplo()
	conOperacion(persona)(&lote)

	err := grafo.ValidarLote(lote)
	exigirRechazoDelLote(t, err, persona, `el nodo de tipo "Persona": `+motivo)
	assert.NotContains(t, err.Error(), documento, "el mensaje nunca repite el documento")
}
