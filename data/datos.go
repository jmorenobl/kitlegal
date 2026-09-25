// Package data lleva dentro del binario los ficheros congelados de data/ que
// se leen en ejecución: los de data/territorio/ (FR-056; research.md D2).
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
	return comunidadesDe(comunidades)
}

// comunidadesDe es el recorrido de Comunidades sobre un sistema de ficheros
// cualquiera, para que sus tres ramas de error —el subárbol que no se lista, la
// entrada que no es un fichero <código>.yaml y el fichero que no se lee— se
// puedan ejercer sin tocar el embebido.
func comunidadesDe(arbol fs.FS) (map[string][]byte, error) {
	entradas, err := fs.ReadDir(arbol, carpetaDeComunidades)
	if err != nil {
		return nil, fmt.Errorf("no se puede listar data/%s: %w", carpetaDeComunidades, err)
	}

	ficheros := make(map[string][]byte, len(entradas))

	for _, entrada := range entradas {
		ruta := path.Join(carpetaDeComunidades, entrada.Name())

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
