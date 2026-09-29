package grafo_test

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/jmorenobl/kitlegal/internal/core/grafo"
)

// TestEtiquetaDeVersionObsoleta fija la única fuente de verdad de la etiqueta de
// la forma fija con la que una skill traslada un hallazgo: la de
// version-obsoleta, y ninguna para fuente-caducada; cada llamada da un mapa
// nuevo, que quien lo recibe puede cambiar sin cambiar el de nadie más (H7.1
// FR-045; data-model §7; research D12).
func TestEtiquetaDeVersionObsoleta(t *testing.T) {
	t.Parallel()

	const etiqueta = "REDACCI\xc3\x93N MODIFICADA"

	etiquetas := grafo.EtiquetasDeHallazgo()
	assert.Equal(t, map[grafo.ClaseDeHallazgo]string{grafo.ClaseVersionObsoleta: etiqueta}, etiquetas)

	delete(etiquetas, grafo.ClaseVersionObsoleta)
	etiquetas[grafo.ClaseFuenteCaducada] = etiqueta

	assert.Equal(t, map[grafo.ClaseDeHallazgo]string{grafo.ClaseVersionObsoleta: etiqueta}, grafo.EtiquetasDeHallazgo())
}
