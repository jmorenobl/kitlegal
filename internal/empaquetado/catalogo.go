package empaquetado

import (
	"bytes"
	"errors"
	"fmt"
	"path/filepath"
	"regexp"
)

// Lo fijo del catálogo (data-model §6 de H22). Este fichero es su plantilla: lo
// fija todo salvo la versión, la dirección —que sale de ella— y la huella (H22
// FR-030; research.md D5).
const (
	// nombreDelCatalogo es el del repositorio que lo publica,
	// jmorenobl/kitlegal-plugins: el plugin se instala como
	// `kitlegal@kitlegal-plugins`.
	nombreDelCatalogo = "kitlegal-plugins"
	// fuenteDeArchivo es la clase de fuente de la entrada: un zip en una
	// dirección, con su huella.
	fuenteDeArchivo = "archive"
	// descargasDeLasReleases es de donde cuelgan los ficheros de cada release
	// del proyecto: detrás va la etiqueta y, detrás, el nombre del fichero.
	descargasDeLasReleases = "https://github.com/jmorenobl/kitlegal/releases/download/"
	// prefijoDeLaEtiqueta es lo que la etiqueta de una release lleva delante de
	// su versión.
	prefijoDeLaEtiqueta = "v"
)

// formaDeLaHuella es la de una huella SHA-256 como la da checksums.txt: 64
// dígitos hexadecimales en minúsculas (data-model §6 de H22).
var formaDeLaHuella = regexp.MustCompile(`^[0-9a-f]{64}$`)

// catalogo es `.claude-plugin/marketplace.json`: los campos de data-model §6 de
// H22, en ese orden, y ninguno más.
type catalogo struct {
	Nombre  string              `json:"name"`
	Titular autoria             `json:"owner"`
	Plugins []pluginDelCatalogo `json:"plugins"`
}

// pluginDelCatalogo es la entrada de un plugin: de dónde se descarga y cómo se
// presenta.
type pluginDelCatalogo struct {
	Nombre      string            `json:"name"`
	Fuente      fuenteDelCatalogo `json:"source"`
	Version     string            `json:"version"`
	Descripcion string            `json:"description"`
	Autoria     autoria           `json:"author"`
	Pagina      string            `json:"homepage"`
	Licencia    string            `json:"license"`
}

// fuenteDelCatalogo es la fuente de una entrada: la clase, la dirección del zip
// y la huella con la que se comprueba cada descarga.
type fuenteDelCatalogo struct {
	Fuente    string `json:"source"`
	Direccion string `json:"url"`
	Huella    string `json:"sha256"`
}

// documentoDelCatalogo compone el catálogo de una versión: una sola entrada,
// `kitlegal`, de fuente `archive`, con la dirección de kitlegal-plugin.zip en
// la release de esa versión —y no en la última—, su huella, la versión y los
// textos de textos.go (contracts/paso.md §4 de H22; H22 FR-030, FR-031). La
// versión llega sin `v`; la dirección lleva la etiqueta, que la tiene.
//
// Rechaza una versión vacía y una huella que no tiene su forma: un catálogo
// con cualquiera de las dos apuntaría a un plugin que nadie puede instalar.
func documentoDelCatalogo(version, huella string) ([]byte, error) {
	if version == "" {
		return nil, errors.New("la versión del catálogo está vacía")
	}

	if !formaDeLaHuella.MatchString(huella) {
		return nil, fmt.Errorf("«%s» no es una huella SHA-256: 64 dígitos hexadecimales en minúsculas", huella)
	}

	var escrito bytes.Buffer

	err := escribirDocumento(&escrito, catalogo{
		Nombre:  nombreDelCatalogo,
		Titular: autoria{Nombre: Autoria},
		Plugins: []pluginDelCatalogo{{
			Nombre: nombreDeLasPiezas,
			Fuente: fuenteDelCatalogo{
				Fuente:    fuenteDeArchivo,
				Direccion: descargasDeLasReleases + prefijoDeLaEtiqueta + version + "/" + nombreDelPlugin,
				Huella:    huella,
			},
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

// escribirCatalogo escribe en ese fichero, de una carpeta que tiene que
// existir, el catálogo de esa versión y esa huella. Cada etiqueta lo sustituye
// entero: el catálogo no acumula versiones (H22 FR-032).
func escribirCatalogo(version, huella, salida string) error {
	escrito, err := documentoDelCatalogo(version, huella)
	if err != nil {
		return err
	}

	return escribirEn(filepath.Dir(salida), filepath.Base(salida), escrito)
}
