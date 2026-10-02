package empaquetado_test

import (
	"path/filepath"
	"strings"
	"testing"
	"unicode/utf8"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/jmorenobl/kitlegal/internal/empaquetado"
)

// maximoDeLaDescripcion es el umbral de FR-015, escrito aquí como lo escribe el
// spec y no leído del paquete.
const maximoDeLaDescripcion = 120

// TestDescripcionCorta es el control en `make ci` del umbral de la descripción
// corta (FR-015, FR-066; SC-009): la del repositorio tiene 120 caracteres como
// mucho, y el paso acepta una de 120 y rechaza una de 121 con la línea del
// contrato (contracts/paso.md §1; research.md D20). Las dos de prueba son de
// caracteres de dos bytes: la de 120 mide 240 bytes y pasa, y la línea de la de
// 121 dice 121 y no 242, así que lo que se cuenta son caracteres.
func TestDescripcionCorta(t *testing.T) {
	t.Parallel()

	t.Run("la del repositorio tiene 120 caracteres como mucho", func(t *testing.T) {
		t.Parallel()

		assert.NotEmpty(t, empaquetado.Descripcion)
		assert.LessOrEqual(t, utf8.RuneCountInString(empaquetado.Descripcion), maximoDeLaDescripcion,
			"la descripción corta del repositorio pasa de 120 caracteres (FR-015)")
	})

	t.Run("una de 120 caracteres pasa", func(t *testing.T) {
		t.Parallel()

		descripcion := strings.Repeat("ñ", maximoDeLaDescripcion)
		require.Greater(t, len(descripcion), maximoDeLaDescripcion, "la de prueba mide más de 120 bytes")

		piezas := piezasDePrueba(t)
		piezas.Descripcion = descripcion

		require.NoError(t, empaquetado.EscribirPiezas(piezas))

		var manifiesto manifiestoLeido

		leerEstricto(t,
			contenidoDe(t, leerZip(t, filepath.Join(piezas.Salida, nombreDeLaExtension)), rutaDelManifiesto), &manifiesto)

		assert.Equal(t, descripcion, manifiesto.Descripcion, "la descripción del manifiesto es la que se le da")

		var ficha fichaDelPluginLeida

		leerEstricto(t,
			contenidoDe(t, leerZip(t, filepath.Join(piezas.Salida, nombreDelPlugin)), rutaDeLaFichaDelPlugin), &ficha)

		assert.Equal(t, descripcion, ficha.Descripcion, "la descripción de plugin.json es la del manifiesto")
	})

	t.Run("una de 121 caracteres falla", func(t *testing.T) {
		t.Parallel()

		piezas := piezasDePrueba(t)
		piezas.Descripcion = strings.Repeat("ñ", maximoDeLaDescripcion+1)

		require.EqualError(t, empaquetado.EscribirPiezas(piezas),
			"la descripción corta tiene 121 caracteres y el máximo es 120")
	})
}
