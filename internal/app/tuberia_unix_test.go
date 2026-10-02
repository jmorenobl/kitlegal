//go:build unix

package app_test

import (
	"bytes"
	"os"
	"os/exec"
	"strings"
	"syscall"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/jmorenobl/kitlegal/internal/cli"
)

// mensajeDeSalidaRota es lo que el mensaje para la persona tiene que decir
// cuando lo que falló fue el descriptor: lo escribe el presentador para todo lo
// que sale por él y el analizador para la ayuda del verbo, y en los dos casos
// nombra el descriptor y no la operación que se estaba haciendo.
const mensajeDeSalidaRota = "no se pudo escribir en la salida estándar"

// TestTuberiaCerrada ejerce el binario real con una tubería de verdad cuyo
// lector ya no existe, que es el caso que el spec nombra —«la salida estándar
// falla al escribirse (tubería cerrada)»— y que ningún escritor en memoria
// reproduce: con un escritor que devuelve un error, el proceso sigue vivo y
// el kernel traduce; con una tubería cerrada, el runtime de Go termina el
// proceso con SIGPIPE **antes** de que el error exista, salvo que la raíz de
// composición desarme la señal (FR-031, FR-033,
// contracts/banderas-y-exit-codes.md §4).
//
// Se comprueban las cinco cosas que se escriben en la salida estándar —el
// sobre, la tabla, las tres ayudas, «version» y el esquema— porque la garantía
// es del descriptor y no de lo que se estaba escribiendo: el proceso termina
// con el código del fallo inesperado, sin señal, y el mensaje para la persona
// sale por la salida de error, que sigue sana.
//
// Es un test del binario y no del kernel en memoria a propósito, y por eso vive
// junto al e2e y usa el binario que TestMain construye: la señal la desarma el
// proceso entero, y solo un proceso entero puede demostrar que está desarmada.
// Y vive en un fichero solo para Unix porque la tubería se pide con
// syscall.Pipe (tuberiaSinLector), que en Windows tiene otra forma.
func TestTuberiaCerrada(t *testing.T) {
	t.Parallel()

	require.NoError(t, entorno.err)

	escriben := [][]string{
		{"echo", "hola", "--json"},
		{"echo", "hola"},
		{"--help"},
		{"echo", "--help"},
		{"echo", "repetir", "--help"},
		{"version"},
		{"echo", "hola", "--describe"},
	}

	for _, argv := range escriben {
		t.Run(strings.Join(argv, " "), func(t *testing.T) {
			t.Parallel()

			codigo, errores := lanzarConLectorCerrado(t, entorno.binario, argv...)

			assert.Equal(t, 1, codigo,
				"una tubería cerrada es el fallo inesperado del contrato, y termina con su código")
			assert.Contains(t, errores, mensajeDeSalidaRota,
				"el mensaje para la persona nombra el descriptor que falló")
			assert.Contains(t, errores, syscall.EPIPE.Error(),
				"la causa es la del sistema: una tubería sin lector")
		})
	}
}

// lanzarConLectorCerrado ejecuta el binario con la salida estándar conectada a
// una tubería que no tiene lector desde **antes** de arrancarlo
// (tuberiaSinLector), de modo que la primera escritura falle con EPIPE.
// Devuelve el código con el que terminó y lo que dejó en la salida de error, y
// exige que haya terminado por su cuenta y no por una señal: un proceso muerto
// por SIGPIPE no tiene código de salida, que es justo lo que este test existe
// para descartar.
//
// El nivel del registro se fija vacío en el entorno del subproceso para que la
// salida de error lleve solo el mensaje, sea cual sea el entorno de quien
// ejecuta los tests.
func lanzarConLectorCerrado(t *testing.T, binario string, argv ...string) (int, string) {
	t.Helper()

	escritor := tuberiaSinLector(t)

	var errores bytes.Buffer

	// El ejecutable es el que TestMain acaba de construir en un temporal y los
	// argumentos son literales de este fichero: no hay entrada de usuario, red
	// ni variable de entorno en la orden, de modo que G204 no tiene aquí nada
	// que prevenir.
	//nolint:gosec // el binario lo construye TestMain y los argumentos son literales de este fichero; no hay entrada externa.
	orden := exec.CommandContext(t.Context(), binario, argv...)
	orden.Stdout = escritor
	orden.Stderr = &errores
	orden.Env = append(os.Environ(), cli.VariableNivel+"=")

	errEjecucion := orden.Run()
	require.NoError(t, escritor.Close())

	var fallo *exec.ExitError
	require.ErrorAs(t, errEjecucion, &fallo,
		"con la salida estándar rota el binario tiene que terminar con un fallo, nunca con 0")

	estado, esEstado := fallo.Sys().(syscall.WaitStatus)
	require.True(t, esEstado, "el estado de salida es el del sistema operativo")
	require.Falsef(t, estado.Signaled(),
		"el proceso murió por la señal %v en lugar de terminar con un código: "+
			"la raíz de composición no ha desarmado SIGPIPE", estado.Signal())

	return fallo.ExitCode(), errores.String()
}

// tuberiaSinLector devuelve el extremo de escritura de una tubería cuyo extremo
// de lectura no ha heredado ningún proceso: la crea y cierra el lector con
// syscall.ForkLock tomado para lectura. Un proceso que otro test en paralelo
// crea hereda, entre su fork y su exec, una copia de cada descriptor abierto;
// con os.Pipe y un Close después, el lector existe en esa ventana, y si el
// binario escribe mientras la copia sigue abierta, la escritura tiene quien la
// lea, no falla y el binario termina con 0: con 16 lanzamientos a la vez pasaba
// en entre 8 y 44 de cada 16 000. Go toma ForkLock para escritura en cada fork,
// así que con él tomado para lectura ningún fork ocurre mientras el lector
// existe, y cerrado ya no lo hereda ningún proceso. La tubería se pide con
// syscall.Pipe y no con os.Pipe porque os.Pipe toma ese mismo candado y lo
// suelta antes de volver, y un candado de lectura no se toma dos veces.
//
// El escritor queda marcado para cerrarse en el exec, como lo deja os.Pipe: el
// binario lo recibe como su salida estándar y ningún otro proceso lo conserva.
func tuberiaSinLector(t *testing.T) *os.File {
	t.Helper()

	var extremos [2]int

	syscall.ForkLock.RLock()

	err := syscall.Pipe(extremos[:])
	if err == nil {
		syscall.CloseOnExec(extremos[1])

		err = syscall.Close(extremos[0])
	}

	syscall.ForkLock.RUnlock()

	require.NoError(t, err, "la tubería se crea y su lector se cierra sin que nadie lo herede")

	return os.NewFile(uintptr(extremos[1]), "tubería sin lector")
}
