// Command laboratorio es el programa del repositorio con el que se mide la
// skill de laboratorio de H39 (contracts/arnes.md §5 de H39; research.md D7 de
// H39). No es un applet de kitlegal, no se instala con él y ningún archivo de la
// release lo lleva: lo construye el arnés del sondeo del laboratorio.
//
// Su orden servidor es el filtro que, en las sesiones de las variantes
// instrucciones y parametro, se pone delante de `kitlegal mcp serve`; y su orden
// tabla, el guion que cuenta un sondeo (contracts/tabla.md de H39; research.md
// D9 de H39), sin red, modelo ni credencial:
//
//	laboratorio servidor -kitlegal <ruta> [-instrucciones <texto>] [-herramienta <h> -parametro <p> -descripcion <texto>]
//	laboratorio tabla <ruta>
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

// Las órdenes de laboratorio: el filtro del servidor de las sesiones y el
// guion que cuenta un sondeo.
const (
	ordenServidor = "servidor"
	ordenTabla    = "tabla"
)

// raizDelRepositorio es la del repositorio cuya declaración del laboratorio lee
// la orden tabla: el directorio en que se ejecuta, que es su raíz cuando lo
// lanzan make laboratorio-tabla o `go run ./cmd/laboratorio` (contracts/tabla.md
// §1 de H39).
const raizDelRepositorio = "."

// codigoDeUso es el código con el que termina sin orden o con una que no es
// suya: el del fallo inesperado, como un defecto de quien lo arranca.
const codigoDeUso = 1

func main() {
	os.Exit(ejecutar(context.Background(), os.Args, os.Stdin, os.Stdout, os.Stderr))
}

// ejecutar atiende la orden que nombra el argumento que sigue al del programa
// con los que le siguen. Sin orden, o con una que no es de laboratorio, escribe
// en la salida de error una línea de uso por cada orden y devuelve 1.
//
// El error de escribir esas líneas no se propaga, y es una decisión y no un
// silencio: el código ya dice que la orden no se atendió, que es lo que lee
// quien la da, y contarlo iría a la misma salida que acaba de fallar.
func ejecutar(ctx context.Context, args []string, entrada io.Reader, salida, errores io.Writer) int {
	if len(args) > 1 {
		switch args[1] {
		case ordenServidor:
			return evals.ServidorDelLaboratorio(ctx, args[2:], entrada, salida, errores)
		case ordenTabla:
			return evals.TablaDelLaboratorio(raizDelRepositorio, args[2:], salida, errores)
		}
	}

	for _, uso := range []string{evals.UsoDelServidorDelLaboratorio, evals.UsoDeLaTablaDelLaboratorio} {
		_, _ = fmt.Fprintf(errores, "%suso: %s\n", evals.PrefijoDelLaboratorio, uso)
	}

	return codigoDeUso
}
