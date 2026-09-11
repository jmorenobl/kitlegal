// Command kitlegal es el punto de entrada del binario del proyecto. En H0 su
// única superficie observable es el verbo version, definido en
// specs/001-h0-esqueleto-del-repo/contracts/cli-version.md.
package main

import (
	"fmt"
	"io"
	"os"
)

// Datos de construcción. Los inyectan con -ldflags las órdenes build e install
// del Makefile; los valores por defecto solo se ven con `go run` y `go test`.
var (
	version = "dev"
	commit  = "none"
	fecha   = "unknown"
)

// uso es la única línea que el binario escribe cuando no reconoce la invocación.
const uso = "uso: kitlegal version\n"

func main() { os.Exit(run(os.Args[1:], os.Stdout, os.Stderr)) }

// run ejecuta la invocación y devuelve el código de salida: 0 cuando el verbo
// es version y 2 («args» en la tabla de códigos estables del proyecto) ante un
// verbo ausente, desconocido o un argumento sobrante. Nunca llama a os.Exit,
// nunca entra en pánico y no abre ninguna conexión ni escribe ningún fichero:
// toda su salida va por los escritores recibidos.
func run(args []string, stdout, stderr io.Writer) int {
	if len(args) != 1 || args[0] != "version" {
		fmt.Fprint(stderr, uso)

		return 2
	}

	fmt.Fprintf(stdout, "kitlegal %s\ncommit: %s\nfecha:  %s\n", version, commit, fecha)

	return 0
}
