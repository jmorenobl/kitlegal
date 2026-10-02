package empaquetado_test

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/jmorenobl/kitlegal/internal/empaquetado"
)

// huellaDePrueba es una huella SHA-256 con su forma: 64 dígitos hexadecimales
// en minúsculas. Es la del ejemplo de contracts/paso.md §4.
const huellaDePrueba = "84881b8db79242e901b527420fa62b1568c52be48edb321ad3aa85002461fcd6"

// catalogoLeido es el catálogo como lo fija data-model §6: sus campos, en su
// orden y ninguno más. Como manifiestoLeido, es la transcripción del contrato y
// no el tipo del paquete.
type catalogoLeido struct {
	Nombre  string               `json:"name"`
	Titular autoriaLeida         `json:"owner"`
	Plugins []entradaDelCatalogo `json:"plugins"`
}

// entradaDelCatalogo es la entrada de un plugin en el catálogo.
type entradaDelCatalogo struct {
	Nombre      string       `json:"name"`
	Fuente      fuenteLeida  `json:"source"`
	Version     string       `json:"version"`
	Descripcion string       `json:"description"`
	Autoria     autoriaLeida `json:"author"`
	Pagina      string       `json:"homepage"`
	Licencia    string       `json:"license"`
}

// fuenteLeida es la fuente `archive` de una entrada: de dónde se descarga el
// plugin y con qué huella se comprueba.
type fuenteLeida struct {
	Fuente    string `json:"source"`
	Direccion string `json:"url"`
	Huella    string `json:"sha256"`
}

// TestCatalogo es el control en `make ci` de la plantilla del catálogo
// (contracts/paso.md §4; data-model §6; FR-030, FR-031): con una versión y una
// huella, el documento, leído de forma estricta, lleva una sola entrada,
// `kitlegal`, de fuente `archive`, con la dirección de la release de esa
// versión y no la de la última, la huella, la versión y los textos de su único
// sitio; sin versión, o con una huella que no tiene su forma, error.
func TestCatalogo(t *testing.T) {
	t.Parallel()

	versiones := []struct {
		nombre    string
		version   string
		direccion string
	}{
		{
			nombre:    "el de una etiqueta",
			version:   "0.4.0",
			direccion: "https://github.com/jmorenobl/kitlegal/releases/download/v0.4.0/kitlegal-plugin.zip",
		},
		{
			nombre:    "el de un snapshot",
			version:   "0.3.2-SNAPSHOT-751b76e",
			direccion: "https://github.com/jmorenobl/kitlegal/releases/download/v0.3.2-SNAPSHOT-751b76e/kitlegal-plugin.zip",
		},
	}

	for _, caso := range versiones {
		t.Run(caso.nombre, func(t *testing.T) {
			t.Parallel()

			escrito, err := empaquetado.DocumentoDelCatalogo(caso.version, huellaDePrueba)
			require.NoError(t, err)

			esperado := catalogoLeido{
				Nombre:  "kitlegal-plugins",
				Titular: autoriaLeida{Nombre: empaquetado.Autoria},
				Plugins: []entradaDelCatalogo{{
					Nombre:      "kitlegal",
					Fuente:      fuenteLeida{Fuente: "archive", Direccion: caso.direccion, Huella: huellaDePrueba},
					Version:     caso.version,
					Descripcion: empaquetado.Descripcion,
					Autoria:     autoriaLeida{Nombre: empaquetado.Autoria},
					Pagina:      "https://kitlegal.es",
					Licencia:    "EUPL-1.2",
				}},
			}

			var catalogo catalogoLeido

			leerEstricto(t, escrito, &catalogo)

			assert.Equal(t, esperado, catalogo,
				"el catálogo lleva una sola entrada, `kitlegal`, de fuente `archive`, con la dirección de la release de "+
					"esa versión, la huella, la versión y los textos, y ningún campo más (FR-030, FR-031)")
			assert.Equal(t, conLaFormaDelContrato(t, esperado), string(escrito),
				"el catálogo va con la forma de contracts/paso.md §5")
		})
	}

	noValen := []struct {
		nombre  string
		version string
		huella  string
		mensaje string
	}{
		{
			nombre:  "sin versión",
			huella:  huellaDePrueba,
			mensaje: "la versión del catálogo está vacía",
		},
		{
			nombre:  "con una huella de 63 dígitos",
			version: "0.4.0",
			huella:  huellaDePrueba[1:],
			mensaje: "«" + huellaDePrueba[1:] + "» no es una huella SHA-256: 64 dígitos hexadecimales en minúsculas",
		},
		{
			nombre:  "con una huella de 65 dígitos",
			version: "0.4.0",
			huella:  huellaDePrueba + "0",
			mensaje: "«" + huellaDePrueba + "0» no es una huella SHA-256: 64 dígitos hexadecimales en minúsculas",
		},
		{
			nombre:  "con una huella en mayúsculas",
			version: "0.4.0",
			huella:  strings.ToUpper(huellaDePrueba),
			mensaje: "«" + strings.ToUpper(huellaDePrueba) +
				"» no es una huella SHA-256: 64 dígitos hexadecimales en minúsculas",
		},
		{
			nombre:  "con una huella con un dígito que no es hexadecimal",
			version: "0.4.0",
			huella:  "g" + huellaDePrueba[1:],
			mensaje: "«g" + huellaDePrueba[1:] + "» no es una huella SHA-256: 64 dígitos hexadecimales en minúsculas",
		},
		{
			nombre:  "con una huella seguida de un salto de línea",
			version: "0.4.0",
			huella:  huellaDePrueba + "\n",
			mensaje: "«" + huellaDePrueba + "\n» no es una huella SHA-256: 64 dígitos hexadecimales en minúsculas",
		},
		{
			nombre:  "sin huella",
			version: "0.4.0",
			mensaje: "«» no es una huella SHA-256: 64 dígitos hexadecimales en minúsculas",
		},
	}

	for _, caso := range noValen {
		t.Run(caso.nombre, func(t *testing.T) {
			t.Parallel()

			escrito, err := empaquetado.DocumentoDelCatalogo(caso.version, caso.huella)

			require.EqualError(t, err, caso.mensaje)
			assert.Empty(t, escrito, "sin versión o sin huella no hay catálogo")
		})
	}
}
