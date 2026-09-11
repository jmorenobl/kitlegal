package schema

import (
	"errors"
	"math"
	"regexp"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestHuella ejerce los tres pasos de la forma canónica y las propiedades de la
// huella que fija contracts/sobre-de-salida.md §4: claves ordenadas, números
// conservados como literales, caracteres HTML sin escapar, ni espacios ni
// saltos, y una huella que depende solo del contenido de data —nunca de
// fecha_consulta ni del orden de serialización— con error, y no pánico, cuando
// data no es serializable (FR-011, FR-012, SC-005).
//
// Es el test que invoca el escenario 4 de quickstart.md.
func TestHuella(t *testing.T) {
	t.Parallel()

	casosCanonicos := []struct {
		nombre   string
		data     any
		esperado string
	}{
		{
			nombre:   "el orden de declaración del struct no manda: las claves salen ordenadas",
			data:     ejemplo{Zeta: "z", Alfa: "a"},
			esperado: `{"alfa":"a","zeta":"z"}`,
		},
		{
			nombre:   "el mismo contenido como mapa produce la misma forma",
			data:     map[string]any{"zeta": "z", "alfa": "a"},
			esperado: `{"alfa":"a","zeta":"z"}`,
		},
		{
			nombre:   "un entero mayor que la precisión de la coma flotante conserva su literal",
			data:     map[string]any{"grande": int64(9007199254740993)},
			esperado: `{"grande":9007199254740993}`,
		},
		{
			nombre:   "los caracteres HTML no se escapan",
			data:     map[string]any{"html": "<a>&</a>"},
			esperado: `{"html":"<a>&</a>"}`,
		},
		{
			nombre: "las claves se ordenan también en los objetos anidados, sin espacios ni saltos",
			data: map[string]any{
				"articulos": []any{
					map[string]any{"titulo": "Objeto", "numero": 1},
					map[string]any{"titulo": "Ámbito", "numero": 2},
				},
			},
			esperado: `{"articulos":[{"numero":1,"titulo":"Objeto"},{"numero":2,"titulo":"Ámbito"}]}`,
		},
		{
			nombre:   "una cadena es un contenido válido",
			data:     "hola",
			esperado: `"hola"`,
		},
		{
			nombre:   "un número es un contenido válido",
			data:     123,
			esperado: `123`,
		},
		{
			nombre:   "un booleano es un contenido válido",
			data:     true,
			esperado: `true`,
		},
		{
			nombre:   "un contenido ausente es nulo",
			data:     nil,
			esperado: `null`,
		},
		{
			nombre:   "el objeto vacío se conserva",
			data:     map[string]any{},
			esperado: `{}`,
		},
		{
			nombre:   "la lista vacía se conserva",
			data:     []any{},
			esperado: `[]`,
		},
	}

	for _, caso := range casosCanonicos {
		t.Run("forma canónica: "+caso.nombre, func(t *testing.T) {
			t.Parallel()

			emitido, err := canonico(caso.data)
			require.NoError(t, err)
			assert.Equal(t, caso.esperado, string(emitido))
		})
	}

	casosPropiedades := []struct {
		nombre  string
		uno     any
		otro    any
		iguales bool
	}{
		{
			nombre:  "el mismo contenido con las claves en otro orden produce la misma huella",
			uno:     ejemplo{Zeta: "z", Alfa: "a"},
			otro:    map[string]any{"alfa": "a", "zeta": "z"},
			iguales: true,
		},
		{
			nombre:  "el mismo contenido escrito con otro tipo numérico produce la misma huella",
			uno:     map[string]any{"numero": 1},
			otro:    map[string]any{"numero": int64(1)},
			iguales: true,
		},
		{
			nombre:  "un byte de diferencia produce otra huella",
			uno:     "hola",
			otro:    "holb",
			iguales: false,
		},
		{
			nombre:  "una clave distinta produce otra huella",
			uno:     map[string]any{"mensaje": "hola"},
			otro:    map[string]any{"mensage": "hola"},
			iguales: false,
		},
	}

	for _, caso := range casosPropiedades {
		t.Run("propiedad: "+caso.nombre, func(t *testing.T) {
			t.Parallel()

			primera, err := Huella(caso.uno)
			require.NoError(t, err)
			segunda, err := Huella(caso.otro)
			require.NoError(t, err)

			if caso.iguales {
				assert.Equal(t, primera, segunda)

				return
			}
			assert.NotEqual(t, primera, segunda)
		})
	}

	t.Run("propiedad: la huella no depende de fecha_consulta", func(t *testing.T) {
		t.Parallel()

		data := map[string]any{"mensaje": "hola"}
		madrid := time.FixedZone("CEST", 2*60*60)
		primer := Sobre{FechaConsulta: time.Date(2026, time.September, 11, 10, 12, 0, 0, madrid), Data: data}
		segundo := Sobre{FechaConsulta: time.Date(2027, time.March, 1, 23, 59, 59, 0, time.UTC), Data: data}

		primera, err := Huella(primer.Data)
		require.NoError(t, err)
		segunda, err := Huella(segundo.Data)
		require.NoError(t, err)

		assert.Equal(t, primera, segunda)
	})

	// Huellas conocidas: sha256 de la forma canónica, calculado fuera de este
	// código, de modo que un cambio del algoritmo no pueda pasar inadvertido.
	casosConocidos := []struct {
		nombre   string
		data     any
		esperado string
	}{
		{
			nombre:   "objeto",
			data:     map[string]any{"mensaje": "hola"},
			esperado: "sha256:f2a2800485840f313f12e331f6627b274d7cc6ae006f2d79e821832990a53fac",
		},
		{
			nombre:   "cadena",
			data:     "hola",
			esperado: "sha256:9a15d9253c28ff08b87d8a03093634d2c38ef8b41731a1828778d083fecab5d6",
		},
		{
			nombre:   "contenido ausente",
			data:     nil,
			esperado: "sha256:74234e98afe7498fb5daf1f36ac2d78acc339464f950703b8c019892f982b90b",
		},
	}

	forma := regexp.MustCompile(`^sha256:[0-9a-f]{64}$`)

	for _, caso := range casosConocidos {
		t.Run("huella conocida: "+caso.nombre, func(t *testing.T) {
			t.Parallel()

			huella, err := Huella(caso.data)
			require.NoError(t, err)
			assert.Equal(t, caso.esperado, huella)
			assert.Regexp(t, forma, huella,
				"la huella lleva el prefijo del algoritmo y 64 dígitos hexadecimales en minúscula")
		})
	}

	casosNoSerializables := []struct {
		nombre string
		data   any
	}{
		{nombre: "un canal", data: make(chan int)},
		{nombre: "una función dentro de un objeto", data: map[string]any{"f": func() {}}},
		{nombre: "un número que no es un número", data: math.NaN()},
		{nombre: "un infinito", data: math.Inf(1)},
	}

	for _, caso := range casosNoSerializables {
		t.Run("data no serializable: "+caso.nombre, func(t *testing.T) {
			t.Parallel()

			var (
				huella string
				err    error
			)
			require.NotPanics(t, func() {
				huella, err = Huella(caso.data)
			}, "un contenido no serializable produce un error, nunca un pánico")

			require.ErrorContains(t, err, "no es serializable")
			require.Error(t, errors.Unwrap(err),
				"el fallo original de encoding/json sigue disponible, no se pierde")
			assert.Empty(t, huella)
		})
	}
}

// ejemplo declara sus campos al revés del orden en que la forma canónica los
// emite, que es lo que hace observable el paso de reordenación.
type ejemplo struct {
	Zeta string `json:"zeta"`
	Alfa string `json:"alfa"`
}
