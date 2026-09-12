package app_test

import (
	"context"
	"errors"
	"fmt"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"testing"

	"github.com/rogpeppe/go-internal/testscript"
	"github.com/stretchr/testify/require"
)

const (
	// nombreDelBinario es el nombre con el que el binario de e2e tiene que
	// aparecer en el PATH. No es cosmético: el despacho multicall lee el último
	// componente de os.Args[0], y cuatro de los guiones invocan «kitlegal» a
	// través del intérprete de órdenes para poder comprobar un código de salida
	// concreto (FR-002, FR-003).
	nombreDelBinario = "kitlegal"

	// paqueteDelBinario es la raíz de composición del binario de e2e: el kernel
	// real más el registro de los applets de ejemplo. Se construye de verdad, y
	// no se simula, porque SC-003 exige un enlace simbólico a un ejecutable real
	// (FR-009, research.md D20).
	paqueteDelBinario = "./ejemplo/kitlegal-e2e"

	// directorioDeGuiones es donde viven los guiones que describen la entrega del
	// hito, relativo al directorio de este paquete.
	directorioDeGuiones = "testdata/script"

	// variableDelBinario es con lo que cada guion nombra el binario por su ruta
	// absoluta. `symlink` no convierte su destino a absoluto, de modo que un
	// valor absoluto es lo que hace que el enlace apunte al ejecutable real
	// (research.md D20).
	variableDelBinario = "KITLEGAL_BIN"

	// nombreInstalado es como `go install` nombra el ejecutable que produce: por
	// el último componente del directorio del paquete.
	nombreInstalado = "kitlegal-e2e"
)

// entorno es lo que TestMain deja preparado para el e2e: la ruta del binario ya
// construido y el fallo que, de haberlo, impidió prepararlo.
//
// El fallo se **guarda** en lugar de terminar el proceso porque los tests del
// kernel que viven en este mismo binario de test no necesitan ningún ejecutable:
// quien no pudo hacer su trabajo es el e2e, y es el e2e quien lo cuenta con un
// fallo atribuido. Terminar aquí convertiría un problema de un test en el
// silencio de todos.
var entorno struct {
	binario string
	err     error
}

// TestMain construye el binario de e2e en un directorio temporal, lo antepone al
// PATH y **retorna**.
//
// Retornar no es un descuido: el envoltorio de testing termina el proceso con el
// resultado de m.Run por su cuenta —«If TestMain returns, the test wrapper will
// pass the result of m.Run to os.Exit itself», testing.go—, así que aquí no hay
// ningún os.Exit que se salte los defer. Es lo que hace que el temporal se borre
// de verdad y, de paso, lo que evita que este fichero necesite una sola
// excepción de lint (research.md D18, D20).
func TestMain(m *testing.M) {
	limpiar := preparar()
	defer limpiar()

	// El resultado se descarta a propósito: con TestMain definido, el envoltorio
	// de testing lee el código de m.Run y es él quien termina el proceso.
	m.Run()
}

// preparar construye el binario y deja el PATH listo, y devuelve lo que hay que
// deshacer al terminar. Nada de lo que falle aquí termina el proceso: se anota
// en `entorno` y lo cuenta el test que necesitaba el binario.
func preparar() func() {
	temporal, err := os.MkdirTemp("", "kitlegal-e2e-")
	if err != nil {
		entorno.err = fmt.Errorf("e2e: no se pudo crear el directorio temporal: %w", err)

		return func() {}
	}

	entorno.binario = filepath.Join(temporal, nombreDelBinario)
	entorno.err = errors.Join(construir(entorno.binario), anteponerAlPath(temporal))

	return func() {
		if err := os.RemoveAll(temporal); err != nil {
			// Ya no queda ningún test al que contárselo —m.Run ha terminado—, así
			// que el aviso va al único canal que sigue vivo. Callarlo dejaría un
			// temporal huérfano sin rastro de quién lo dejó.
			log.Printf("e2e: no se pudo borrar el directorio temporal %s: %v", temporal, err)
		}
	}
}

// construir compila la raíz de composición del binario de e2e y la deja en el
// destino que se le da.
//
// El directorio viaja en el entorno, por GOBIN, y no como argumento de `-o`: así
// la orden queda escrita **entera con constantes**, que es lo que el análisis de
// seguridad exige de un subproceso y lo que evita que este fichero necesite una
// excepción de lint. Lo que se construye es el binario real —no un arnés que
// simule el despacho—, que es lo único capaz de sostener el enlace simbólico de
// SC-003 (research.md D20).
//
// No toca la red: cuando `go test` llega hasta aquí, lo que el módulo necesita
// está ya en la caché de módulos.
func construir(destino string) error {
	directorio := filepath.Dir(destino)

	orden := exec.CommandContext(context.Background(), "go", "install", paqueteDelBinario)
	orden.Env = append(os.Environ(), "GOBIN="+directorio)

	if salida, err := orden.CombinedOutput(); err != nil {
		return fmt.Errorf("e2e: no se pudo construir %s: %w: %s",
			paqueteDelBinario, err, salida)
	}

	// El ejecutable se instala con el nombre del directorio del paquete, y el
	// despacho multicall lee precisamente ese nombre: renombrarlo es lo que hace
	// que los guiones puedan invocarlo como «kitlegal» (FR-002, FR-003).
	if err := os.Rename(filepath.Join(directorio, nombreInstalado), destino); err != nil {
		return fmt.Errorf("e2e: no se pudo nombrar el binario como %s: %w",
			nombreDelBinario, err)
	}

	return nil
}

// anteponerAlPath pone el directorio del binario recién construido por delante
// de todo lo demás, de modo que un `kitlegal` instalado en la máquina no pueda
// atender ninguna invocación de ningún guion.
func anteponerAlPath(directorio string) error {
	path := directorio + string(os.PathListSeparator) + os.Getenv("PATH")
	if err := os.Setenv("PATH", path); err != nil {
		return fmt.Errorf("e2e: no se pudo anteponer %s al PATH: %w", directorio, err)
	}

	return nil
}

// TestEntregaDelHito ejecuta los guiones que describen la entrega —la invocación
// literal del hito, el enlace simbólico, la ayuda, el código 2 con argumentos
// malos, la autodescripción, el verbo obligatorio del segundo applet y la salida
// estándar con solo el documento JSON— contra el binario que construyó TestMain
// (FR-055, SC-001, SC-002, SC-003, SC-009).
//
// RequireExplicitExec obliga a que cada orden externa de un guion se escriba con
// `exec`: sin él, una línea mal escrita podría ejecutar un programa de la máquina
// creyendo que ejecuta una orden integrada. Ningún guion accede a la red y
// ninguno escribe fuera del directorio de trabajo que testscript le da y borra.
func TestEntregaDelHito(t *testing.T) {
	t.Parallel()

	require.NoError(t, entorno.err)

	testscript.Run(t, testscript.Params{
		Dir:                 directorioDeGuiones,
		RequireExplicitExec: true,
		Setup: func(env *testscript.Env) error {
			env.Setenv(variableDelBinario, entorno.binario)

			return nil
		},
	})
}
