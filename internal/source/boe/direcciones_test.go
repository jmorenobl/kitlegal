package boe

import (
	"net/url"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/jmorenobl/kitlegal/internal/core/schema"
)

// TestDirecciones fija byte a byte las direcciones que la fuente pide y cita
// (data-model.md §8, contrato verbos-y-salidas §0 a §6): la base de la API y,
// bajo ella, la norma, su bloque, su índice, sus metadatos y su análisis; y las
// páginas públicas act.php de la norma y del bloque, que van dentro de data.
// De todas exige lo que FR-002 y FR-003 piden a lo que sale de la fuente: https,
// de www.boe.es, fuera del espacio reservado kitlegal: y aceptable como
// procedencia de un sobre.
func TestDirecciones(t *testing.T) {
	t.Parallel()

	const (
		norma   = "BOE-A-2015-10565"
		bloque  = "a21"
		api     = "https://www.boe.es/datosabiertos/api/legislacion-consolidada"
		publica = "https://www.boe.es/buscar/act.php"
	)

	casos := []struct {
		nombre    string
		direccion string
		esperada  string
	}{
		{nombre: "la base de la API", direccion: baseDeLaAPI, esperada: api},
		{nombre: "la norma", direccion: direccionDeLaNorma(norma), esperada: api + "/id/BOE-A-2015-10565"},
		{
			nombre:    "el bloque",
			direccion: direccionDelBloque(norma, bloque),
			esperada:  api + "/id/BOE-A-2015-10565/texto/bloque/a21",
		},
		{nombre: "el índice", direccion: direccionDelIndice(norma), esperada: api + "/id/BOE-A-2015-10565/texto/indice"},
		{nombre: "los metadatos", direccion: direccionDeLosMetadatos(norma), esperada: api + "/id/BOE-A-2015-10565/metadatos"},
		{nombre: "el análisis", direccion: direccionDelAnalisis(norma), esperada: api + "/id/BOE-A-2015-10565/analisis"},
		{nombre: "la pública de la norma", direccion: direccionPublicaDeLaNorma(norma), esperada: publica + "?id=BOE-A-2015-10565"},
		{
			nombre:    "la pública del bloque",
			direccion: direccionPublicaDelBloque(norma, bloque),
			esperada:  publica + "?id=BOE-A-2015-10565#a21",
		},
		{
			// Un resultado de búsqueda sin identificador conserva su dirección
			// pública con el identificador vacío (data-model.md §2.3).
			nombre:    "la pública de un resultado sin identificador",
			direccion: direccionPublicaDeLaNorma(""),
			esperada:  publica + "?id=",
		},
	}

	for _, caso := range casos {
		t.Run(caso.nombre, func(t *testing.T) {
			t.Parallel()

			assert.Equal(t, caso.esperada, caso.direccion)
			compruebaDireccionDeLaFuente(t, caso.direccion)
		})
	}

	t.Run("la fuente que las firma no es del espacio reservado", func(t *testing.T) {
		t.Parallel()

		assert.Equal(t, "boe.legislacion-consolidada", NombreDeLaFuente)
		assert.False(t, strings.HasPrefix(NombreDeLaFuente, "kitlegal."),
			"la fuente %q firma con el espacio reservado de lo que no consulta ninguna fuente", NombreDeLaFuente)
	})
}

// compruebaDireccionDeLaFuente exige a una dirección lo que el sobre de la
// fuente necesita para citarla (FR-002, data-model.md §9, invariante 1): que
// empiece por https://www.boe.es/, que sea https y de ese sitio sin nada
// delante, que no sea del espacio reservado kitlegal: y que el kernel la acepte
// como procedencia junto al nombre de la fuente.
func compruebaDireccionDeLaFuente(t *testing.T, direccion string) {
	t.Helper()

	assert.True(t, strings.HasPrefix(direccion, "https://www.boe.es/"),
		"la dirección %q no es del sitio del BOE", direccion)
	assert.False(t, strings.HasPrefix(direccion, "kitlegal:"),
		"la dirección %q es del espacio reservado", direccion)

	destino, err := url.Parse(direccion)
	require.NoError(t, err)
	assert.Equal(t, "https", destino.Scheme)
	assert.Equal(t, "www.boe.es", destino.Host)
	assert.Nil(t, destino.User)
	require.NoError(t, schema.Procedencia{Fuente: NombreDeLaFuente, URL: direccion}.Validar())
}
