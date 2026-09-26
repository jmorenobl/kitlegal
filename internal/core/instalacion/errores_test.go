package instalacion_test

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/jmorenobl/kitlegal/internal/core/instalacion"
)

// TestManifiestoIlegibleSinCausa exige que el valor cero del error, que
// cualquiera puede construir porque el tipo se exporta para reconocerlo con
// errors.As, se pueda usar sin un panic: dice que el manifiesto es ilegible y
// no envuelve nada (FR-144).
func TestManifiestoIlegibleSinCausa(t *testing.T) {
	t.Parallel()

	var ilegible instalacion.ManifiestoIlegible

	assert.Equal(t, "manifiesto ilegible", ilegible.Error())
	assert.NoError(t, ilegible.Unwrap())
}
