package grafo_test

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/jmorenobl/kitlegal/internal/core/grafo"
)

// TestVocabularioDeLaResolucion fija el tipo de nodo de una resolución judicial
// y las ocho claves de sus datos, una a una: las de `data` en la salida de
// `cita`, de modo que el nodo y el sobre nombran igual cada metadato (H23
// FR-042; data-model §5; contracts/cita-resolver.md §6).
func TestVocabularioDeLaResolucion(t *testing.T) {
	t.Parallel()

	assert.Equal(t, "Resolucion", grafo.TipoResolucion)

	casos := []struct {
		nombre string
		clave  string
		valor  string
	}{
		{"DatoECLI", grafo.DatoECLI, "ecli"},
		{"DatoROJ", grafo.DatoROJ, "roj"},
		{"DatoOrgano", grafo.DatoOrgano, "organo"},
		{"DatoFecha", grafo.DatoFecha, "fecha"},
		{"DatoNumeroResolucion", grafo.DatoNumeroResolucion, "numero_resolucion"},
		{"DatoNumeroRecurso", grafo.DatoNumeroRecurso, "numero_recurso"},
		{"DatoPonente", grafo.DatoPonente, "ponente"},
		{"DatoURL", grafo.DatoURL, "url"},
	}

	for _, caso := range casos {
		t.Run(caso.nombre, func(t *testing.T) {
			t.Parallel()

			assert.Equal(t, caso.valor, caso.clave)
		})
	}
}

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
