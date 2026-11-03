// Command laboratorio es el programa del repositorio con el que se mide la
// skill de laboratorio de H39 (contracts/arnes.md §5 de H39; research.md D7 de
// H39). No es un applet de kitlegal, no se instala con él y ningún archivo de la
// release lo lleva: lo construye el arnés del sondeo del laboratorio.
//
// Su orden servidor es el filtro que, en las sesiones de las variantes
// instrucciones y parametro, se pone delante de `kitlegal mcp serve`:
//
//	laboratorio servidor -kitlegal <ruta> [-instrucciones <texto>] [-herramienta <h> -parametro <p> -descripcion <texto>]
//
// No decide nada: elige la orden por su primer argumento, inyecta en
// internal/evals los demás, la entrada, la salida y la salida de error, y
// termina el proceso con el código que devuelve.
package main

import (
	"context"
	"fmt"
	"io"
	"os"

	"github.com/jmorenobl/kitlegal/internal/evals"
)

// ordenServidor es la orden del filtro del servidor de las sesiones.
const ordenServidor = "servidor"

// codigoDeUso es el código con el que termina sin orden o con una que no es
// suya: el del fallo inesperado, como un defecto de quien lo arranca.
const codigoDeUso = 1

func main() {
	os.Exit(ejecutar(context.Background(), os.Args, os.Stdin, os.Stdout, os.Stderr))
}

// ejecutar atiende la orden que nombra el argumento que sigue al del programa
// con los que le siguen. Sin orden, o con una que no es de laboratorio, escribe
// su línea de uso en la salida de error y devuelve 1.
//
// El error de escribir esa línea no se propaga, y es una decisión y no un
// silencio: el código ya dice que la orden no se atendió, que es lo que lee
// quien la da, y contarlo iría a la misma salida que acaba de fallar.
func ejecutar(ctx context.Context, args []string, entrada io.Reader, salida, errores io.Writer) int {
	if len(args) > 1 && args[1] == ordenServidor {
		return evals.ServidorDelLaboratorio(ctx, args[2:], entrada, salida, errores)
	}

	_, _ = fmt.Fprintf(errores, "%suso: %s\n", evals.PrefijoDelLaboratorio, evals.UsoDelServidorDelLaboratorio)

	return codigoDeUso
}
