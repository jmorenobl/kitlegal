package empaquetado_test

import (
	"archive/zip"
	"bytes"
	"io/fs"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/jmorenobl/kitlegal/internal/empaquetado"
)

// Lo que el catálogo fija, escrito aquí y no leído del paquete: es contra lo
// que se compara.
const (
	// rutaDelCatalogo es el documento del catálogo, y carpetaDelPluginEnElCatalogo,
	// la del plugin dentro de él.
	rutaDelCatalogo              = ".claude-plugin/marketplace.json"
	carpetaDelPluginEnElCatalogo = "plugins/kitlegal"
)

// catalogoLeido es el catálogo: sus campos, en su orden y ninguno más. Como
// manifiestoLeido, es la transcripción del contrato y no el tipo del paquete.
type catalogoLeido struct {
	Nombre  string               `json:"name"`
	Titular autoriaLeida         `json:"owner"`
	Plugins []entradaDelCatalogo `json:"plugins"`
}

// entradaDelCatalogo es la entrada de un plugin en el catálogo: su fuente es
// una ruta del propio catálogo.
type entradaDelCatalogo struct {
	Nombre      string       `json:"name"`
	Fuente      string       `json:"source"`
	Version     string       `json:"version"`
	Descripcion string       `json:"description"`
	Autoria     autoriaLeida `json:"author"`
	Pagina      string       `json:"homepage"`
	Licencia    string       `json:"license"`
}

// pluginDePrueba escribe con el paso las piezas de esa versión y devuelve la
// ruta de su kitlegal-plugin.zip y los ficheros que lleva, con su contenido.
func pluginDePrueba(t *testing.T, version string) (string, map[string][]byte) {
	t.Helper()

	piezas := piezasDePrueba(t)
	piezas.Version = version

	require.NoError(t, empaquetado.EscribirPiezas(piezas))

	ruta := filepath.Join(piezas.Salida, nombreDelPlugin)
	ficheros := map[string][]byte{}

	for _, entrada := range leerZip(t, ruta) {
		ficheros[entrada.Nombre] = entrada.Contenido
	}

	return ruta, ficheros
}

// ficherosBajo son los ficheros de una carpeta, con su ruta relativa a ella,
// con `/`, y su contenido.
func ficherosBajo(t *testing.T, carpeta string) map[string][]byte {
	t.Helper()

	ficheros := map[string][]byte{}

	require.NoError(t, filepath.WalkDir(carpeta, func(ruta string, entrada fs.DirEntry, err error) error {
		if err != nil || entrada.IsDir() {
			return err
		}

		relativa, err := filepath.Rel(carpeta, ruta)
		if err != nil {
			return err
		}

		ficheros[filepath.ToSlash(relativa)] = leerFichero(t, ruta)

		return nil
	}))

	return ficheros
}

// TestCatalogo es el control en `make ci` de la plantilla del catálogo (H22
// FR-030, FR-031): con una versión, el documento, leído de forma estricta,
// lleva una sola entrada, `kitlegal`, cuya fuente es la carpeta del plugin
// dentro del catálogo, con la versión y los textos de su único sitio. La app de
// escritorio de Claude no sincroniza un catálogo cuya entrada apunta a un zip
// de fuera (ADR 0035, «Prueba con la v0.4.0»).
func TestCatalogo(t *testing.T) {
	t.Parallel()

	for _, version := range []string{"0.4.1", "0.4.0-SNAPSHOT-751b76e"} {
		t.Run(version, func(t *testing.T) {
			t.Parallel()

			escrito, err := empaquetado.DocumentoDelCatalogo(version)
			require.NoError(t, err)

			esperado := catalogoLeido{
				Nombre:  "kitlegal-plugins",
				Titular: autoriaLeida{Nombre: empaquetado.Autoria},
				Plugins: []entradaDelCatalogo{{
					Nombre:      "kitlegal",
					Fuente:      "./" + carpetaDelPluginEnElCatalogo,
					Version:     version,
					Descripcion: empaquetado.Descripcion,
					Autoria:     autoriaLeida{Nombre: empaquetado.Autoria},
					Pagina:      "https://kitlegal.es",
					Licencia:    "EUPL-1.2",
				}},
			}

			var catalogo catalogoLeido

			leerEstricto(t, escrito, &catalogo)

			assert.Equal(t, esperado, catalogo,
				"el catálogo lleva una sola entrada, `kitlegal`, con la carpeta del plugin dentro del catálogo como "+
					"fuente, la versión y los textos, y ningún campo más (FR-030, FR-031)")
			assert.Equal(t, conLaFormaDelContrato(t, esperado), string(escrito),
				"el catálogo va con la forma de contracts/paso.md §5")
		})
	}
}

// TestCatalogoEscrito comprueba lo que EscribirCatalogo deja en su carpeta: el
// documento del catálogo y, en plugins/kitlegal, cada fichero del plugin de esa
// versión, byte a byte, y nada más.
func TestCatalogoEscrito(t *testing.T) {
	t.Parallel()

	const version = "0.4.1"

	plugin, delPlugin := pluginDePrueba(t, version)
	salida := t.TempDir()

	require.NoError(t, empaquetado.EscribirCatalogo(version, plugin, salida))

	documento, err := empaquetado.DocumentoDelCatalogo(version)
	require.NoError(t, err)

	esperados := map[string][]byte{rutaDelCatalogo: documento}
	for nombre, contenido := range delPlugin {
		esperados[carpetaDelPluginEnElCatalogo+"/"+nombre] = contenido
	}

	require.Contains(t, esperados, carpetaDelPluginEnElCatalogo+"/"+rutaDeLaFichaDelPlugin,
		"premisa: el plugin de prueba lleva su plugin.json")
	assert.Equal(t, esperados, ficherosBajo(t, salida),
		"la carpeta lleva el catálogo y, en %s, cada fichero del plugin, byte a byte, y nada más (FR-030)",
		carpetaDelPluginEnElCatalogo)
}

// TestCatalogoSinEscribir comprueba con qué no se escribe el catálogo, y que
// entonces no deja nada: un plugin de otra versión, uno que no es un zip, uno
// sin plugin.json, uno con una entrada que sale de su carpeta, y una carpeta de
// salida que ya lleva un catálogo.
func TestCatalogoSinEscribir(t *testing.T) {
	t.Parallel()

	const version = "0.4.1"

	t.Run("con el plugin de otra versión", func(t *testing.T) {
		t.Parallel()

		plugin, _ := pluginDePrueba(t, "0.4.0")
		salida := t.TempDir()

		require.EqualError(t, empaquetado.EscribirCatalogo(version, plugin, salida),
			plugin+" es de la versión «0.4.0», y el catálogo, de la «0.4.1»")
		assert.Empty(t, ficherosBajo(t, salida), "sin el plugin de su versión no hay catálogo")
	})

	t.Run("con un plugin que no es un zip", func(t *testing.T) {
		t.Parallel()

		plugin := escribirFichero(t, nombreDelPlugin, []byte("no es un zip"))
		salida := t.TempDir()

		require.EqualError(t, empaquetado.EscribirCatalogo(version, plugin, salida),
			"no se puede leer "+plugin+": "+zip.ErrFormat.Error())
		assert.Empty(t, ficherosBajo(t, salida))
	})

	t.Run("con un plugin sin plugin.json", func(t *testing.T) {
		t.Parallel()

		plugin := escribirFichero(t, nombreDelPlugin, zipDe(t, map[string]string{"skills/alfa/SKILL.md": "# alfa\n"}))
		salida := t.TempDir()

		require.EqualError(t, empaquetado.EscribirCatalogo(version, plugin, salida),
			plugin+" no lleva "+rutaDeLaFichaDelPlugin)
		assert.Empty(t, ficherosBajo(t, salida))
	})

	t.Run("con una entrada que sale de la carpeta del plugin", func(t *testing.T) {
		t.Parallel()

		plugin := escribirFichero(t, nombreDelPlugin, zipDe(t, map[string]string{
			rutaDeLaFichaDelPlugin:      `{"version": "` + version + `"}`,
			"../../../fuera/de/aqui.md": "fuera\n",
		}))
		salida := filepath.Join(t.TempDir(), "catalogo")
		require.NoError(t, os.Mkdir(salida, 0o700))

		require.Error(t, empaquetado.EscribirCatalogo(version, plugin, salida))
		assert.NoDirExists(t, filepath.Join(filepath.Dir(salida), "fuera"),
			"una entrada del zip no escribe fuera de la carpeta de salida")
	})

	for _, propia := range []string{".claude-plugin", "plugins"} {
		t.Run("con una carpeta de salida que ya lleva "+propia, func(t *testing.T) {
			t.Parallel()

			plugin, _ := pluginDePrueba(t, version)
			salida := t.TempDir()
			yaEsta := filepath.Join(salida, propia)
			require.NoError(t, os.Mkdir(yaEsta, 0o700))

			require.EqualError(t, empaquetado.EscribirCatalogo(version, plugin, salida),
				yaEsta+" ya existe: el catálogo se escribe en una carpeta que no lleva otro")
			assert.Empty(t, ficherosBajo(t, salida))
		})
	}
}

// zipDe es un zip con esas entradas, con su nombre tal cual.
func zipDe(t *testing.T, entradas map[string]string) []byte {
	t.Helper()

	var escrito bytes.Buffer

	escritor := zip.NewWriter(&escrito)

	for nombre, contenido := range entradas {
		entrada, err := escritor.Create(nombre)
		require.NoError(t, err)

		_, err = entrada.Write([]byte(contenido))
		require.NoError(t, err)
	}

	require.NoError(t, escritor.Close())

	return escrito.Bytes()
}
