package schema

import (
	"go/build"
	"reflect"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestContexto comprueba las dos invariantes del contexto de ejecución
// (data-model.md §4): lleva las seis opciones globales ya interpretadas y
// ninguna más —en particular ningún registrador, que es un puerto con entrada y
// salida detrás— y el paquete que lo declara sigue siendo dominio puro,
// importando solo crypto/sha256, encoding/hex, encoding/json y time (FR-018,
// R1 de contracts/reglas-de-arquitectura.md).
func TestContexto(t *testing.T) {
	t.Parallel()

	contexto := reflect.TypeOf(Contexto{})

	campos := []struct {
		nombre string
		tipo   string
	}{
		{nombre: "JSON", tipo: "bool"},
		{nombre: "Timeout", tipo: "time.Duration"},
		{nombre: "Offline", tipo: "bool"},
		{nombre: "DryRun", tipo: "bool"},
		{nombre: "SinGrafo", tipo: "bool"},
		{nombre: "Asunto", tipo: "string"},
	}

	for _, caso := range campos {
		t.Run("campo: "+caso.nombre, func(t *testing.T) {
			t.Parallel()

			campo, existe := contexto.FieldByName(caso.nombre)
			require.True(t, existe, "el contexto de ejecución declara el campo")
			assert.Equal(t, caso.tipo, campo.Type.String())
		})
	}

	t.Run("seis campos y ninguno más: ningún registrador viaja en el contexto", func(t *testing.T) {
		t.Parallel()

		assert.Equal(t, len(campos), contexto.NumField())
	})

	t.Run("el dominio importa solo la biblioteca estándar que no hace entrada ni salida", func(t *testing.T) {
		t.Parallel()

		// build.ImportDir lee los ficheros de producción del paquete —no los de
		// test— y devuelve sus importaciones ya deduplicadas.
		paquete, err := build.ImportDir(".", 0)
		require.NoError(t, err)

		assert.ElementsMatch(t,
			[]string{"crypto/sha256", "encoding/hex", "encoding/json", "time"},
			paquete.Imports)
	})
}
