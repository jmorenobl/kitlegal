// Command empaquetar es el paso que empaqueta las dos piezas con las que
// kitlegal se instala sin terminal —kitlegal.mcpb, la extensión de escritorio, y
// kitlegal-plugin.zip, el plugin de Claude— y que escribe el catálogo de una
// versión (contracts/paso.md §1 de H22). Es un programa de construcción del
// repositorio: lo ejecutan goreleaser y los flujos de la release con `go run`,
// no es un applet de kitlegal, no se instala y ningún archivo de la release lo
// lleva (H22 FR-001).
//
// No decide nada: inyecta en internal/empaquetado los argumentos y la salida de
// error, y termina el proceso con el código que Ejecutar devuelve.
package main

import (
	"os"

	"github.com/jmorenobl/kitlegal/internal/empaquetado"
)

func main() {
	os.Exit(empaquetado.Ejecutar(os.Args[1:], os.Stderr))
}
