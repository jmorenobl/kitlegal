package app_test

import (
	"context"
	"errors"
	"fmt"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"slices"
	"testing"
	"time"

	"github.com/rogpeppe/go-internal/testscript"
	"github.com/stretchr/testify/require"

	"github.com/jmorenobl/kitlegal/internal/cache"
	"github.com/jmorenobl/kitlegal/internal/source/boe"
)

const (
	// nombreDelBinario es el nombre con el que el binario de e2e tiene que
	// aparecer en el PATH. No es cosmético: el despacho multicall lee el último
	// componente de os.Args[0], y cuatro de los guiones invocan «kitlegal» a
	// través del intérprete de órdenes para poder comprobar un código de salida
	// concreto (FR-002, FR-003).
	nombreDelBinario = "kitlegal"

	// paqueteDelBinario es la raíz de composición del binario de e2e: el kernel
	// real con los applets de ejemplo y boe sobre la reproducción de sus
	// grabaciones. Se construye de verdad, y no se simula, porque SC-003 exige un
	// enlace simbólico a un ejecutable real (FR-009, FR-114, research.md D20).
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

	// grabacionesDeLaFuente son las grabaciones de la API del BOE, relativas al
	// directorio de este paquete: las mismas que sirven a los tests del adaptador
	// y del applet. Ningún guion las modifica, porque cada uno trabaja sobre su
	// propia copia.
	grabacionesDeLaFuente = "../source/boe/testdata/" + boe.NombreDeLaFuente

	// directorioDeReproduccion es la carpeta del directorio de trabajo de cada
	// guion de la que el binario de e2e sirve las grabaciones, con una subcarpeta
	// por fuente. Es la misma ruta relativa que escribe su raíz de composición, y
	// el binario la resuelve contra el directorio desde el que se lo invoca, que
	// en un guion es $WORK (research.md D13 de H4).
	directorioDeReproduccion = "reproduccion"

	// directorioDeLaCache es la carpeta de $WORK en la que vive la caché de cada
	// guion: empieza vacía, se borra con el guion y ninguna invocación toca la
	// caché de la cuenta de quien ejecuta los tests.
	directorioDeLaCache = "cache"

	// ordenCronometra es el nombre con el que los guiones escriben la orden que
	// mide una invocación.
	ordenCronometra = "cronometra"
)

// Los motivos por los que cronometra hace fallar un guion: un uso que no se
// puede medir, un programa que no termina bien y una medida que no se cumple.
// Son centinelas para que TestCronometra distinga cada rama sin leer mensajes.
var (
	errUsoDeCronometra      = errors.New("uso: " + ordenCronometra + " <máximo> <programa> <argumentos…>")
	errCodigoDistintoDeCero = errors.New(ordenCronometra + ": el programa no terminó con código 0")
	errMaximoAlcanzado      = errors.New(ordenCronometra + ": el programa tardó el máximo o más")
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
// ninguno escribe fuera del directorio de trabajo que testscript le da y borra:
// boe responde desde la copia de sus grabaciones que Setup deja en ese
// directorio, y su caché vive en él (FR-114, research.md D13 de H4).
func TestEntregaDelHito(t *testing.T) {
	t.Parallel()

	require.NoError(t, entorno.err)

	testscript.Run(t, testscript.Params{
		Dir:                 directorioDeGuiones,
		RequireExplicitExec: true,
		Setup: func(env *testscript.Env) error {
			env.Setenv(variableDelBinario, entorno.binario)
			env.Setenv(cache.VariableDirectorio, filepath.Join(env.WorkDir, directorioDeLaCache))

			return copiarGrabaciones(env.WorkDir)
		},
		Cmds: map[string]func(ts *testscript.TestScript, neg bool, args []string){
			ordenCronometra: cronometra,
		},
	})
}

// copiarGrabaciones deja en el directorio de trabajo de un guion una copia de
// las grabaciones de la fuente, en la carpeta de reproducción de la que las
// sirve el binario de e2e.
//
// Se copian, y no se enlazan, porque un guion puede vaciar o borrar su
// reproducción —para comprobar que lo servido desde la caché no pide nada, o
// --offline sin grabaciones— y eso no puede alcanzar a las que comparten todos
// los tests (research.md D13 de H4).
func copiarGrabaciones(trabajo string) error {
	destino := filepath.Join(trabajo, directorioDeReproduccion, boe.NombreDeLaFuente)

	if err := os.CopyFS(destino, os.DirFS(grabacionesDeLaFuente)); err != nil {
		return fmt.Errorf("e2e: no se pudieron copiar las grabaciones de %s a %s: %w",
			grabacionesDeLaFuente, destino, err)
	}

	return nil
}

// cronometra es la orden `cronometra <máximo> <programa> <argumentos…>` de los
// guiones: ejecuta el programa como lo haría `exec` —con el entorno y el
// directorio del guion, y dejando sus salidas para las órdenes siguientes— y
// hace fallar el guion si no termina con código 0 o si tarda el máximo o más
// (FR-117, SC-002, research.md D13 de H4).
//
// Mide alrededor de TestScript.Exec, dentro del proceso de los tests: así la
// invocación entera —arranque del proceso incluido— queda dentro de la medida,
// y la ruta del binario no llega nunca a un exec.Command escrito aquí.
//
// No admite «!»: negar una medida no dice qué se esperaba —un fallo, o solo que
// tardara—, y un guion que lo escribiera pasaría por la razón equivocada.
func cronometra(ts *testscript.TestScript, neg bool, args []string) {
	if neg {
		ts.Fatalf("%s no admite «!»: la medida se cumple o hace fallar el guion", ordenCronometra)
	}

	ts.Check(cronometrar(args, ts.Exec, time.Now))
}

// cronometrar interpreta los argumentos de cronometra, ejecuta el programa con
// ejecutar y mide con ahora lo que tarda. Devuelve por qué la medida no se
// cumple, o nulo si el programa terminó bien por debajo del máximo.
//
// Recibe la ejecución y la hora para que su contrato —el uso, el código y el
// límite exacto— se pruebe sin lanzar ningún guion (TestCronometra). Un uso
// incorrecto no ejecuta nada.
func cronometrar(
	args []string,
	ejecutar func(programa string, argumentos ...string) error,
	ahora func() time.Time,
) error {
	if len(args) < 2 {
		return fmt.Errorf("%w: faltan el máximo o el programa en %q", errUsoDeCronometra, args)
	}

	maximo, err := time.ParseDuration(args[0])
	if err != nil {
		return fmt.Errorf("%w: el máximo %q no es una duración: %w", errUsoDeCronometra, args[0], err)
	}

	if maximo <= 0 {
		return fmt.Errorf("%w: el máximo %s tiene que ser positivo", errUsoDeCronometra, maximo)
	}

	programa, argumentos := args[1], args[2:]

	inicio := ahora()
	err = ejecutar(programa, argumentos...)
	duracion := ahora().Sub(inicio)

	if err != nil {
		return fmt.Errorf("%w: %s %q terminó tras %s: %w",
			errCodigoDistintoDeCero, programa, argumentos, duracion, err)
	}

	if duracion >= maximo {
		return fmt.Errorf("%w: %s %q tardó %s y el máximo es %s",
			errMaximoAlcanzado, programa, argumentos, duracion, maximo)
	}

	return nil
}

// TestCronometra fija el contrato de la orden cronometra sin lanzar ningún
// guion: la ejecución y la hora se sustituyen, de modo que el límite exacto
// —tardar justo el máximo ya es fallar— se comprueba sin depender de lo que
// tarde de verdad nada, y un uso incorrecto no llega a ejecutar el programa
// (FR-117, SC-002, research.md D13 de H4).
func TestCronometra(t *testing.T) {
	t.Parallel()

	invocacion := []string{"200ms", nombreDelBinario, "boe", "articulo", "BOE-A-2015-10565", "a21", "--json"}
	falloDelPrograma := errors.New("exit status 1")

	casos := []struct {
		nombre string
		args   []string
		// tarda es lo que avanza la hora mientras se ejecuta el programa, y
		// devuelve, lo que devuelve su ejecución.
		tarda    time.Duration
		devuelve error
		// quiere son los errores que el fallo tiene que dejar alcanzables; vacío,
		// la medida se cumple.
		quiere []error
		// ejecuta dice si el programa llega a ejecutarse, con exactamente el
		// programa y los argumentos que siguen al máximo.
		ejecuta bool
	}{
		{nombre: "por-debajo-del-maximo", args: invocacion, tarda: 199 * time.Millisecond, ejecuta: true},
		{nombre: "sin-argumentos-del-programa", args: []string{"1s", nombreDelBinario}, tarda: time.Millisecond, ejecuta: true},
		{
			nombre: "alcanza-el-maximo", args: invocacion, tarda: 200 * time.Millisecond,
			quiere: []error{errMaximoAlcanzado}, ejecuta: true,
		},
		{
			nombre: "supera-el-maximo", args: invocacion, tarda: time.Second,
			quiere: []error{errMaximoAlcanzado}, ejecuta: true,
		},
		{
			nombre: "codigo-distinto-de-cero", args: invocacion, tarda: time.Millisecond, devuelve: falloDelPrograma,
			quiere: []error{errCodigoDistintoDeCero, falloDelPrograma}, ejecuta: true,
		},
		{nombre: "sin-nada", quiere: []error{errUsoDeCronometra}},
		{nombre: "sin-programa", args: []string{"200ms"}, quiere: []error{errUsoDeCronometra}},
		{nombre: "maximo-que-no-es-una-duracion", args: []string{"rapido", nombreDelBinario}, quiere: []error{errUsoDeCronometra}},
		{nombre: "maximo-nulo", args: []string{"0s", nombreDelBinario}, quiere: []error{errUsoDeCronometra}},
		{nombre: "maximo-negativo", args: []string{"-1s", nombreDelBinario}, quiere: []error{errUsoDeCronometra}},
	}

	for _, caso := range casos {
		t.Run(caso.nombre, func(t *testing.T) {
			t.Parallel()

			hora := time.Date(2026, time.September, 13, 10, 0, 0, 0, time.UTC)

			var ejecutadas [][]string

			ejecutar := func(programa string, argumentos ...string) error {
				ejecutadas = append(ejecutadas, slices.Concat([]string{programa}, argumentos))
				hora = hora.Add(caso.tarda)

				return caso.devuelve
			}

			err := cronometrar(caso.args, ejecutar, func() time.Time { return hora })

			if len(caso.quiere) == 0 {
				require.NoError(t, err)
			}

			for _, quiere := range caso.quiere {
				require.ErrorIs(t, err, quiere)
			}

			if caso.ejecuta {
				require.Equal(t, [][]string{caso.args[1:]}, ejecutadas)
			} else {
				require.Empty(t, ejecutadas, "un uso incorrecto no ejecuta nada")
			}
		})
	}
}
