package schema

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestClase comprueba el vocabulario de clases de error que aparece en el sobre
// y en el esquema que emite --describe: las seis clases, con el valor exacto que
// declara contracts/sobre-de-salida.md §6, y los datos de error con sus dos
// únicas claves (FR-029, FR-045).
func TestClase(t *testing.T) {
	t.Parallel()

	casos := []struct {
		nombre string
		clase  Clase
		valor  string
	}{
		{nombre: "argumentos", clase: ClaseArgumentos, valor: "argumentos"},
		{nombre: "no encontrado", clase: ClaseNoEncontrado, valor: "no-encontrado"},
		{nombre: "fuente no disponible", clase: ClaseFuenteNoDisponible, valor: "fuente-no-disponible"},
		{nombre: "límite o términos de uso", clase: ClaseLimiteOTos, valor: "limite-o-tos"},
		{nombre: "identidad humana", clase: ClaseIdentidadHumana, valor: "identidad-humana"},
		{nombre: "inesperado", clase: ClaseInesperado, valor: "inesperado"},
	}

	for _, caso := range casos {
		t.Run("clase: "+caso.nombre, func(t *testing.T) {
			t.Parallel()

			assert.Equal(t, caso.valor, string(caso.clase))

			emitido, err := json.Marshal(DatosError{Clase: caso.clase, Mensaje: "mensaje para la persona"})
			require.NoError(t, err)
			assert.JSONEq(t,
				`{"clase":"`+caso.valor+`","mensaje":"mensaje para la persona"}`,
				string(emitido))
		})
	}

	t.Run("las seis clases son distintas entre sí", func(t *testing.T) {
		t.Parallel()

		vistas := make(map[Clase]struct{}, len(casos))
		for _, caso := range casos {
			vistas[caso.clase] = struct{}{}
		}
		assert.Len(t, vistas, len(casos))
	})

	t.Run("el vocabulario enumera las seis del contrato, en su orden, y ninguna más", func(t *testing.T) {
		t.Parallel()

		// La lista esperada sale de la tabla de arriba, que escribe los valores
		// del contrato a mano: si alguien añadiera una séptima constante y se
		// olvidara de Clases —o al revés—, el esquema de --describe y el sobre
		// dejarían de hablar del mismo vocabulario, y esto lo señalaría.
		esperadas := make([]Clase, 0, len(casos))
		for _, caso := range casos {
			esperadas = append(esperadas, Clase(caso.valor))
		}

		assert.Equal(t, esperadas, Clases())

		primera, segunda := Clases(), Clases()
		primera[0] = "alterada"
		assert.NotEqual(t, primera[0], segunda[0],
			"cada llamada devuelve una lista propia: el vocabulario no se altera desde fuera")
	})

	t.Run("los datos de error llevan exactamente clase y mensaje", func(t *testing.T) {
		t.Parallel()

		emitido, err := json.Marshal(DatosError{Clase: ClaseNoEncontrado, Mensaje: "no existe el bloque a99"})
		require.NoError(t, err)

		var claves map[string]json.RawMessage
		require.NoError(t, json.Unmarshal(emitido, &claves))

		emitidas := make([]string, 0, len(claves))
		for clave := range claves {
			emitidas = append(emitidas, clave)
		}
		assert.ElementsMatch(t, []string{"clase", "mensaje"}, emitidas)
	})
}
