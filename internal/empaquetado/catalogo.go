package empaquetado

import (
	"archive/zip"
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"path"
	"path/filepath"
)

// Lo fijo del catálogo. Este fichero es su plantilla: lo fija todo salvo la
// versión y el plugin, que es el de la release de esa versión (H22 FR-030).
//
// El catálogo lleva el plugin dentro, como carpeta, y su entrada lo nombra con
// una ruta relativa. Nació en H22 con una fuente `archive` —la dirección de
// kitlegal-plugin.zip en la release y su huella—, que Claude Code admite y la
// app de escritorio de Claude no: al añadir el catálogo, la app daba un error
// de sincronización, y con el plugin dentro lo sincroniza (ADR 0035, «Prueba
// con la v0.4.0»).
const (
	// nombreDelCatalogo es el del repositorio que lo publica,
	// jmorenobl/kitlegal-plugins: el plugin se instala como
	// `kitlegal@kitlegal-plugins`.
	nombreDelCatalogo = "kitlegal-plugins"
	// rutaDelCatalogo es el documento del catálogo, y carpetaDeLosPlugins, la
	// que lleva dentro cada plugin, en una carpeta con su nombre.
	rutaDelCatalogo     = ".claude-plugin/marketplace.json"
	carpetaDeLosPlugins = "plugins"
	// fuenteRelativa es lo que una ruta del catálogo lleva delante para que
	// quien lo lee la tome por una carpeta del propio repositorio.
	fuenteRelativa = "./"
	// modoDeCarpeta es el de las carpetas que el catálogo crea.
	modoDeCarpeta = 0o755
)

// catalogo es `.claude-plugin/marketplace.json`: estos campos, en este orden,
// y ninguno más.
type catalogo struct {
	Nombre  string              `json:"name"`
	Titular autoria             `json:"owner"`
	Plugins []pluginDelCatalogo `json:"plugins"`
}

// pluginDelCatalogo es la entrada de un plugin: dónde está, dentro del
// catálogo, y cómo se presenta.
type pluginDelCatalogo struct {
	Nombre      string  `json:"name"`
	Fuente      string  `json:"source"`
	Version     string  `json:"version"`
	Descripcion string  `json:"description"`
	Autoria     autoria `json:"author"`
	Pagina      string  `json:"homepage"`
	Licencia    string  `json:"license"`
}

// carpetaDelPlugin es la del plugin dentro del catálogo: plugins/kitlegal.
func carpetaDelPlugin() string {
	return path.Join(carpetaDeLosPlugins, nombreDeLasPiezas)
}

// documentoDelCatalogo compone el catálogo de una versión: una sola entrada,
// `kitlegal`, cuya fuente es la carpeta del plugin dentro del catálogo, con la
// versión y los textos de textos.go (H22 FR-030, FR-031). La versión llega sin
// `v`.
func documentoDelCatalogo(version string) ([]byte, error) {
	var escrito bytes.Buffer

	err := escribirDocumento(&escrito, catalogo{
		Nombre:  nombreDelCatalogo,
		Titular: autoria{Nombre: Autoria},
		Plugins: []pluginDelCatalogo{{
			Nombre:      nombreDeLasPiezas,
			Fuente:      fuenteRelativa + carpetaDelPlugin(),
			Version:     version,
			Descripcion: Descripcion,
			Autoria:     autoria{Nombre: Autoria},
			Pagina:      paginaDelProyecto,
			Licencia:    licenciaDelProyecto,
		}},
	})

	// El error es el de escribir el documento, que quien llama mira antes que
	// los bytes.
	return escrito.Bytes(), err
}

// escribirCatalogo escribe en esa carpeta, que tiene que existir y no llevar
// ya un catálogo, el de esa versión: `.claude-plugin/marketplace.json` y, en
// plugins/kitlegal, cada fichero de ese kitlegal-plugin.zip, byte a byte. Cada
// etiqueta lo sustituye entero: el catálogo no acumula versiones (H22 FR-032).
//
// El plugin tiene que ser el de esa versión: si su plugin.json dice otra, el
// catálogo anunciaría una versión y entregaría otra, y no se escribe nada.
func escribirCatalogo(version, plugin, salida string) (err error) {
	ficheros, err := ficherosDelPlugin(plugin, version)
	if err != nil {
		return err
	}

	documento, err := documentoDelCatalogo(version)
	if err != nil {
		return err
	}

	raiz, err := os.OpenRoot(salida)
	if err != nil {
		return falloDeEscritura(salida, err)
	}

	defer func() { err = errors.Join(err, raiz.Close()) }()

	// Sin restos de otro catálogo: un fichero de una versión anterior que esta
	// ya no lleva se quedaría en el plugin.
	for _, propia := range []string{path.Dir(rutaDelCatalogo), carpetaDeLosPlugins} {
		if _, err := raiz.Lstat(propia); err == nil {
			return fmt.Errorf("%s ya existe: el catálogo se escribe en una carpeta que no lleva otro",
				filepath.Join(salida, propia))
		}
	}

	if err := escribirEnLaRaiz(raiz, salida, rutaDelCatalogo, documento); err != nil {
		return err
	}

	for _, fichero := range ficheros {
		ruta := path.Join(carpetaDelPlugin(), fichero.nombre)

		if err := escribirEnLaRaiz(raiz, salida, ruta, fichero.contenido); err != nil {
			return err
		}
	}

	return nil
}

// ficheroDelPlugin es un fichero del plugin: su ruta dentro de él, con `/`, y
// su contenido.
type ficheroDelPlugin struct {
	nombre    string
	contenido []byte
}

// ficherosDelPlugin lee ese kitlegal-plugin.zip y devuelve sus ficheros, en su
// orden. Falla si no se puede leer, si no lleva plugin.json o si el suyo no es
// de esa versión.
func ficherosDelPlugin(plugin, version string) ([]ficheroDelPlugin, error) {
	lector, err := zip.OpenReader(plugin)
	if err != nil {
		return nil, fmt.Errorf("no se puede leer %s: %w", plugin, causaDe(err))
	}

	ficheros, err := leerEntradas(plugin, lector.File)

	if err := errors.Join(err, lector.Close()); err != nil {
		return nil, err
	}

	for _, fichero := range ficheros {
		if fichero.nombre != rutaDeLaFichaDelPlugin {
			continue
		}

		var ficha fichaDelPlugin
		if err := json.Unmarshal(fichero.contenido, &ficha); err != nil {
			return nil, fmt.Errorf("no se puede leer %s de %s: %w", rutaDeLaFichaDelPlugin, plugin, err)
		}

		if ficha.Version != version {
			return nil, fmt.Errorf("%s es de la versión «%s», y el catálogo, de la «%s»", plugin, ficha.Version, version)
		}

		return ficheros, nil
	}

	return nil, fmt.Errorf("%s no lleva %s", plugin, rutaDeLaFichaDelPlugin)
}

// leerEntradas lee el contenido de cada entrada del zip que es un fichero.
func leerEntradas(plugin string, entradas []*zip.File) ([]ficheroDelPlugin, error) {
	ficheros := make([]ficheroDelPlugin, 0, len(entradas))

	for _, entrada := range entradas {
		if entrada.FileInfo().IsDir() {
			continue
		}

		contenido, err := leerEntrada(entrada)
		if err != nil {
			return nil, fmt.Errorf("no se puede leer %s de %s: %w", entrada.Name, plugin, err)
		}

		ficheros = append(ficheros, ficheroDelPlugin{nombre: entrada.Name, contenido: contenido})
	}

	return ficheros, nil
}

// leerEntrada lee una entrada del zip entera.
func leerEntrada(entrada *zip.File) (contenido []byte, err error) {
	abierta, err := entrada.Open()
	if err != nil {
		return nil, err
	}

	defer func() { err = errors.Join(err, abierta.Close()) }()

	return io.ReadAll(abierta)
}

// escribirEnLaRaiz escribe un fichero en esa ruta de la raíz, con `/`, creando
// las carpetas que le falten y sin salir de ella: una ruta que escapa de la
// raíz —la de una entrada de un zip que no es el del paso— la rechaza la propia
// raíz (research.md D16 de H22).
func escribirEnLaRaiz(raiz *os.Root, salida, ruta string, contenido []byte) error {
	local := filepath.FromSlash(ruta)

	if err := raiz.MkdirAll(filepath.Dir(local), modoDeCarpeta); err != nil {
		return falloDeEscritura(filepath.Join(salida, local), err)
	}

	if err := raiz.WriteFile(local, contenido, modoDeFichero); err != nil {
		return falloDeEscritura(filepath.Join(salida, local), err)
	}

	return nil
}

// causaDe es la causa de un error del sistema sin la operación ni la ruta con
// que la envuelve el paquete os, para no nombrar la ruta dos veces.
func causaDe(err error) error {
	var enRuta *os.PathError
	if errors.As(err, &enRuta) {
		return enRuta.Err
	}

	return err
}
