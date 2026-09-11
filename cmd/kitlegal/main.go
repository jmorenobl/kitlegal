// Command kitlegal es el punto de entrada del binario del proyecto. No decide
// nada y no escribe nada: inyecta en la raíz de composición del kernel los
// descriptores del sistema, el registro de producción y los datos de
// construcción, y termina el proceso con el código que esta devuelve (FR-035,
// FR-040, research.md D16 y D27).
//
// El contrato observable ya no es el de H0
// (specs/001-h0-esqueleto-del-repo/contracts/cli-version.md): de aquel se
// conservan las tres líneas de version y su código 0, y lo demás lo fija el
// despacho multicall de contracts/registro-y-describe.md §2 y §3.
package main

import (
	"os"

	"github.com/jmorenobl/kitlegal/internal/app"
)

// Datos de construcción. Los inyectan con -ldflags las órdenes build e install
// del Makefile; los valores por defecto solo se ven con `go run` y `go test`.
var (
	version = "dev"
	commit  = "none"
	fecha   = "unknown"
)

// main es el único sitio del binario que llama a os.Exit: app.Main devuelve el
// código de salida y nunca termina el proceso, que es lo que permite ejercer el
// contrato entero con escritores en memoria
// (contracts/reglas-de-arquitectura.md §2, R4).
func main() {
	os.Exit(app.Main(
		os.Args, app.RegistroDeProduccion(), os.Stdout, os.Stderr, version, commit, fecha,
	))
}
