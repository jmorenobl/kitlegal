package render

import (
	"bytes"
	"encoding/json"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/jmorenobl/kitlegal/internal/core/schema"
)

// TestTabla comprueba la tabla mínima: los cuatro datos de procedencia y
// después el contenido, de modo que la cita no se pierda por elegir la forma
// legible por una persona (FR-041, FR-043,
// contracts/sobre-de-salida.md §7).
func TestTabla(t *testing.T) {
	t.Parallel()

	t.Run("los cuatro datos de procedencia y después el contenido", func(t *testing.T) {
		t.Parallel()

		var salida bytes.Buffer
		require.NoError(t, escribirTabla(&salida, sobreDeEjemplo()))

		assert.Equal(t, ""+
			"fuente          kitlegal.echo\n"+
			"url             kitlegal:applet/echo\n"+
			"fecha_consulta  2026-09-11T10:12:00+02:00\n"+
			"hash            "+hashDeEjemplo+"\n"+
			"mensaje         hola\n",
			salida.String())
	})

	t.Run("la procedencia va siempre, aunque no haya contenido que mostrar", func(t *testing.T) {
		t.Parallel()

		sobre := sobreDeEjemplo()
		sobre.Data = map[string]any{}

		var salida bytes.Buffer
		require.NoError(t, escribirTabla(&salida, sobre))

		lineas := strings.Split(strings.TrimSuffix(salida.String(), "\n"), "\n")
		require.Len(t, lineas, 4, "un contenido vacío no aporta filas, pero la cita se conserva")
		for i, clave := range []string{"fuente", "url", "fecha_consulta", "hash"} {
			assert.True(t, strings.HasPrefix(lineas[i], clave),
				"la línea %d empieza por %s", i, clave)
		}
	})

	t.Run("el contenido anidado se aplana a pares ruta/valor", func(t *testing.T) {
		t.Parallel()

		sobre := sobreDeEjemplo()
		sobre.Data = map[string]any{
			"articulos": []any{
				map[string]any{"titulo": "Uno"},
				map[string]any{"titulo": "Dos"},
			},
		}

		var salida bytes.Buffer
		require.NoError(t, escribirTabla(&salida, sobre))

		assert.Contains(t, salida.String(), "articulos.0.titulo  Uno\n")
		assert.Contains(t, salida.String(), "articulos.1.titulo  Dos\n")
	})

	casos := []struct {
		nombre string
		data   any
		filas  []string
	}{
		{
			nombre: "las claves de un objeto se recorren en orden, como en la forma canónica",
			data:   map[string]any{"zeta": "última", "alfa": "primera"},
			filas:  []string{"alfa primera", "zeta última"},
		},
		{
			nombre: "los números conservan su literal, sin pasar por la coma flotante",
			data:   json.RawMessage(`{"expediente":9007199254740993,"importe":1234.50}`),
			filas:  []string{"expediente 9007199254740993", "importe 1234.50"},
		},
		{
			nombre: "los booleanos y los nulos se muestran con su literal",
			data:   map[string]any{"vigente": true, "derogada": nil},
			filas:  []string{"derogada null", "vigente true"},
		},
		{
			nombre: "un contenido que no es compuesto se nombra por la clave del sobre",
			data:   "hola",
			filas:  []string{"data hola"},
		},
		{
			nombre: "una lista en la raíz usa sus índices como ruta",
			data:   []any{"primera", "segunda"},
			filas:  []string{"0 primera", "1 segunda"},
		},
		{
			nombre: "una lista vacía no se imprime: los valores compuestos se recorren",
			data:   map[string]any{"articulos": []any{}, "norma": "BOE-A-2015-10565"},
			filas:  []string{"norma BOE-A-2015-10565"},
		},
		{
			nombre: "el contenido de un fallo son sus dos claves",
			data:   schema.DatosError{Clase: schema.ClaseNoEncontrado, Mensaje: "no existe el artículo 99"},
			filas:  []string{"clase no-encontrado", "mensaje no existe el artículo 99"},
		},
	}

	for _, caso := range casos {
		t.Run(caso.nombre, func(t *testing.T) {
			t.Parallel()

			sobre := sobreDeEjemplo()
			sobre.Data = caso.data

			var salida bytes.Buffer
			require.NoError(t, escribirTabla(&salida, sobre))

			contenido := lineasTrasLaProcedencia(t, salida.String())
			assert.Equal(t, caso.filas, contenido)
		})
	}

	t.Run("el fallo de escritura se propaga desde el vaciado final", func(t *testing.T) {
		t.Parallel()

		roto := &escritorRoto{}
		require.ErrorIs(t, escribirTabla(roto, sobreDeEjemplo()), errTuberiaCerrada)
	})

	t.Run("un contenido que no se puede serializar es un error, no un pánico", func(t *testing.T) {
		t.Parallel()

		sobre := sobreDeEjemplo()
		sobre.Data = make(chan int)

		var salida bytes.Buffer
		var err error
		require.NotPanics(t, func() { err = escribirTabla(&salida, sobre) })
		require.Error(t, err)
		assert.Empty(t, salida.String(),
			"un sobre que no se puede recorrer no deja media tabla en la salida estándar")
	})
}

// lineasTrasLaProcedencia devuelve las filas del contenido —las que siguen a
// las cuatro de procedencia— como el par ruta/valor separado por un solo
// espacio. El relleno con el que la tabla alinea la columna se comprueba en el
// caso de la forma completa; aquí estorbaría, porque depende de la ruta más
// larga de toda la tabla.
func lineasTrasLaProcedencia(t *testing.T, tabla string) []string {
	t.Helper()

	lineas := strings.Split(strings.TrimSuffix(tabla, "\n"), "\n")
	require.GreaterOrEqual(t, len(lineas), 4, "la tabla lleva siempre las cuatro de procedencia")

	contenido := make([]string, 0, len(lineas)-4)
	for _, linea := range lineas[4:] {
		ruta, relleno, hayValor := strings.Cut(linea, " ")
		require.True(t, hayValor, "toda fila de la tabla es un par ruta/valor")
		contenido = append(contenido, ruta+" "+strings.TrimLeft(relleno, " "))
	}

	return contenido
}
