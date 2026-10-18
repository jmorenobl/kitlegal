// Package data lleva dentro del binario los ficheros congelados de data/ que
// se leen en ejecución: los de data/territorio/ (FR-056; research.md D2) y,
// desde H9, data/plazos.yaml y los de data/festivos/ (H9 FR-001, FR-028;
// research.md D1 y D12 de H9).
//
// Las directivas //go:embed viven aquí, junto a los ficheros, porque un patrón
// no puede subir de directorio: un paquete de internal/ no los alcanzaría. Cada
// fichero se nombra en su directiva, y un patrón que no casa con ningún fichero
// es un error de compilación, así que la ausencia de uno nunca llega a la
// ejecución; que su contenido valga lo
// comprueba make ci contra sus esquemas (contrato de datos §2 y §3). El paquete
// no interpreta nada: entrega los bytes tal como están escritos, y quien los
// analiza es el dominio (research.md D3).
package data

import (
	"embed"
	"fmt"
	"io/fs"
	"path"
	"strings"
)

// Municipios es data/territorio/municipios.yaml, la relación de municipios del
// INE.
//
//go:embed territorio/municipios.yaml
var Municipios []byte

// DIR3 es data/territorio/dir3.yaml, la correspondencia verificada INE→DIR3.
//
//go:embed territorio/dir3.yaml
var DIR3 []byte

// Estado es data/territorio/estado.yaml, lo nacional que no es de ninguna
// comunidad: el boletín estatal.
//
//go:embed territorio/estado.yaml
var Estado []byte

// comunidades es el subárbol data/territorio/comunidades/, con un fichero por
// comunidad y ciudad autónoma. Cada uno se embebe por su nombre, y no el
// directorio entero, para que la falta de cualquiera de los 19 sea un error de
// compilación y no un fallo en ejecución.
//
//go:embed territorio/comunidades/01.yaml territorio/comunidades/02.yaml territorio/comunidades/03.yaml territorio/comunidades/04.yaml
//go:embed territorio/comunidades/05.yaml territorio/comunidades/06.yaml territorio/comunidades/07.yaml territorio/comunidades/08.yaml
//go:embed territorio/comunidades/09.yaml territorio/comunidades/10.yaml territorio/comunidades/11.yaml territorio/comunidades/12.yaml
//go:embed territorio/comunidades/13.yaml territorio/comunidades/14.yaml territorio/comunidades/15.yaml territorio/comunidades/16.yaml
//go:embed territorio/comunidades/17.yaml territorio/comunidades/18.yaml territorio/comunidades/19.yaml
var comunidades embed.FS

// Dónde está el subárbol de comunidades dentro de lo embebido y cómo se llama
// cada uno de sus ficheros: <código>.yaml.
const (
	carpetaDeComunidades = "territorio/comunidades"
	extensionDeComunidad = ".yaml"
)

// Comunidades devuelve el contenido de cada fichero del subárbol de
// comunidades, indexado por el código que le da nombre: el de
// territorio/comunidades/13.yaml va en "13". Una entrada del subárbol que no es
// un fichero <código>.yaml es un error que la nombra, no algo que se descarta
// en silencio. Cada llamada devuelve un mapa nuevo.
func Comunidades() (map[string][]byte, error) {
	return porComunidad(comunidades, carpetaDeComunidades)
}

// Plazos es data/plazos.yaml, los tipos de plazo de plazos calcular con su
// cómputo, su duración y los preceptos que los fundan.
//
//go:embed plazos.yaml
var Plazos []byte

// FestivosNacionales es data/festivos/nacionales-y-autonomicos.yaml, los
// festivos nacionales y autonómicos de las resoluciones del BOE.
//
//go:embed festivos/nacionales-y-autonomicos.yaml
var FestivosNacionales []byte

// festivosLocales es el subárbol data/festivos/locales/, con un fichero por
// comunidad cuyos festivos locales se han derivado de su boletín. Se embebe por
// patrón y no fichero a fichero, como las comunidades, porque no hay un número
// fijo: los festivos locales llegan comunidad a comunidad, con su boletín
// (principio IX), y la que no tiene fichero es la que no tiene festivos
// locales cargados. Un patrón que no casa con ninguno sigue siendo un error de
// compilación.
//
//go:embed festivos/locales/*.yaml
var festivosLocales embed.FS

// carpetaDeFestivosLocales es dónde está ese subárbol dentro de lo embebido;
// cada uno de sus ficheros se llama como el de su comunidad en
// territorio/comunidades/.
const carpetaDeFestivosLocales = "festivos/locales"

// FestivosLocales devuelve el contenido de cada fichero de
// data/festivos/locales/, indexado por el código de la comunidad que le da
// nombre, como Comunidades: el de festivos/locales/13.yaml va en "13". Cada
// llamada devuelve un mapa nuevo.
func FestivosLocales() (map[string][]byte, error) {
	return porComunidad(festivosLocales, carpetaDeFestivosLocales)
}

// porComunidad es el recorrido de una carpeta con un fichero <código>.yaml por
// comunidad sobre un sistema de ficheros cualquiera, para que sus tres ramas de
// error —la carpeta que no se lista, la entrada que no es un fichero
// <código>.yaml y el fichero que no se lee— se puedan ejercer sin tocar el
// embebido.
func porComunidad(arbol fs.FS, carpeta string) (map[string][]byte, error) {
	entradas, err := fs.ReadDir(arbol, carpeta)
	if err != nil {
		return nil, fmt.Errorf("no se puede listar data/%s: %w", carpeta, err)
	}

	ficheros := make(map[string][]byte, len(entradas))

	for _, entrada := range entradas {
		ruta := path.Join(carpeta, entrada.Name())

		codigo, conExtension := strings.CutSuffix(entrada.Name(), extensionDeComunidad)
		if entrada.IsDir() || !conExtension {
			return nil, fmt.Errorf("data/%s no es el fichero de una comunidad, <código>%s", ruta, extensionDeComunidad)
		}

		contenido, err := fs.ReadFile(arbol, ruta)
		if err != nil {
			return nil, fmt.Errorf("no se puede leer data/%s: %w", ruta, err)
		}

		ficheros[codigo] = contenido
	}

	return ficheros, nil
}
