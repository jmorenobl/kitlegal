//go:build fuentes

package app

import (
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/jmorenobl/kitlegal/internal/cache"
)

// TestVerificarFuentes es la verificación nocturna contra la fuente real
// (FR-115, SC-009; contrato esquemas-fixtures-y-controles §7): el caso
// «boe articulo» de verificarArticulo con el applet boe del binario distribuido,
// AppletBoe(DependenciasDeRed()), y la caché en una carpeta temporal, de modo que
// la invocación pide a la fuente en vez de servirse de lo que haya en la caché
// de la cuenta, y no deja nada en ella.
//
// Toca la red. Solo lo ejecuta scripts/verify-sources.sh, que pone la etiqueta
// fuentes, desde make verify-sources y el flujo nocturno; make ci no lo compila.
func TestVerificarFuentes(t *testing.T) {
	t.Parallel()

	dependencias := DependenciasDeRed()
	dependencias.Cache = []cache.Opcion{cache.ConDirectorio(t.TempDir())}

	require.NoError(t, verificarArticulo(registroDeBoe(t, dependencias), esquemaDeArticulo(t)))
}
