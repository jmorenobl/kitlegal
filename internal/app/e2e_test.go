package app_test

import (
	"archive/tar"
	"bytes"
	"compress/gzip"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/fs"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"slices"
	"strings"
	"sync"
	"testing"
	"testing/fstest"
	"time"

	"github.com/rogpeppe/go-internal/testscript"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/jmorenobl/kitlegal/internal/cache"
	"github.com/jmorenobl/kitlegal/internal/core/instalacion"
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
	// real con los applets de ejemplo, boe sobre la reproducción de sus
	// grabaciones, skills y territorio. Se construye de verdad, y no se simula,
	// porque SC-003 exige un enlace simbólico a un ejecutable real (FR-009,
	// FR-114, research.md D20).
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

	// ordenArbol es el nombre con el que los guiones escriben la orden que
	// lista un directorio sin seguir enlaces, para comparar el disco byte a byte
	// (contracts/arnes-e2e.md §4).
	ordenArbol = "arbol"
)

// Las versiones de los binarios de e2e que no son el de desarrollo, y los
// -ldflags con que se construyen (contracts/arnes-e2e.md §2; research.md D24).
// Son constantes, como el resto de la orden de construcción: ni el entorno ni
// ningún argumento deciden con qué se construye un binario.
const (
	versionV1 = "v0.1.0"
	versionV2 = "v0.2.0"

	// enlazadorQueFalla es el valor de la variable de cadena enlazador del
	// package main de e2e que elige el creador de enlaces que siempre falla, y
	// es el de la constante del mismo nombre de ejemplo/kitlegal-e2e/main.go.
	// El enlazador de Go ignora sin avisar un -X que nombra una variable que no
	// existe: lo que comprueba que este llega es TestBinariosDelArnes.
	enlazadorQueFalla = "falla"

	ldflagsV1         = "-ldflags=-X main.version=" + versionV1
	ldflagsV2         = "-ldflags=-X main.version=" + versionV2
	ldflagsSinEnlaces = ldflagsV1 + " -X main.enlazador=" + enlazadorQueFalla
)

// Las variables de entorno que el arnés deja en cada guion, además de
// KITLEGAL_BIN y KITLEGAL_CACHE_DIR (contracts/arnes-e2e.md §3), y la que elige
// el origen del snapshot.
const (
	variableV1               = "KITLEGAL_V1_BIN"
	variableV2               = "KITLEGAL_V2_BIN"
	variableSinEnlaces       = "KITLEGAL_SIN_ENLACES_BIN"
	variableSkills           = "KITLEGAL_SKILLS"
	variableInstalador       = "KITLEGAL_INSTALADOR"
	variableOrigen           = "KITLEGAL_ORIGEN"
	variableVersionDelOrigen = "KITLEGAL_ORIGEN_VERSION"
	variableArchivoDelOrigen = "KITLEGAL_ORIGEN_ARCHIVO"

	// variableDist es la ruta de dist/ tras make release (make
	// snapshot-check): con ella, el origen sirve el snapshot (§5).
	variableDist = "KITLEGAL_DIST"

	// skillsDelRepositorio e instaladorDelRepositorio son, relativos al
	// directorio de este paquete, el skills/ del repositorio, con el que un
	// guion compara lo instalado byte a byte, y scripts/install.sh, que el arnés
	// nombra sin exigir que exista: llega después que el arnés en el orden de
	// implementación, y solo lo usan los guiones instalador- (§3).
	skillsDelRepositorio     = "../../skills"
	instaladorDelRepositorio = "../../scripts/install.sh"

	// proxyCerrado es a donde apuntan los proxies de cada guion: un puerto de la
	// propia máquina que no atiende HTTP, de modo que toda petición HTTP(S) que
	// se colara fallaría sin salir de ella (§3; research.md V32).
	proxyCerrado = "http://127.0.0.1:9"
)

// variablesDeProxy son las que apuntan a proxyCerrado en cada guion, y
// variablesSinProxy, las excepciones que quedan vacías para que ninguna
// dirección se libre del proxy (contracts/arnes-e2e.md §3).
var (
	variablesDeProxy  = []string{"http_proxy", "https_proxy", "HTTP_PROXY", "HTTPS_PROXY", "ALL_PROXY"}
	variablesSinProxy = []string{"NO_PROXY", "no_proxy"}
)

// Los motivos por los que cronometra hace fallar un guion: un uso que no se
// puede medir, un programa que no termina bien y una medida que no se cumple.
// Son centinelas para que TestCronometra distinga cada rama sin leer mensajes.
var (
	errUsoDeCronometra      = errors.New("uso: " + ordenCronometra + " <máximo> <programa> <argumentos…>")
	errCodigoDistintoDeCero = errors.New(ordenCronometra + ": el programa no terminó con código 0")
	errMaximoAlcanzado      = errors.New(ordenCronometra + ": el programa tardó el máximo o más")
)

// binarioDelArnes es una de las cuatro construcciones del binario de e2e de
// contracts/arnes-e2e.md §2: el mismo paquete, con la versión y el creador de
// enlaces que fijan sus -ldflags.
type binarioDelArnes struct {
	// variable es la que lleva su ruta absoluta a los guiones.
	variable string
	// carpeta es la del directorio temporal del arnés en la que se instala,
	// siempre con el nombre kitlegal.
	carpeta string
	// orden es la que lo construye, escrita entera con constantes: es lo que el
	// análisis de seguridad exige de un subproceso. El directorio de destino
	// viaja en el entorno, por GOBIN, y no como argumento de -o.
	orden func(ctx context.Context) *exec.Cmd
}

// binariosDelArnes son los cuatro que TestMain construye. El de desarrollo, sin
// -ldflags y con la versión dev de su código, es el que va en el PATH.
var binariosDelArnes = [...]binarioDelArnes{
	{
		variable: variableDelBinario,
		carpeta:  "bin",
		orden: func(ctx context.Context) *exec.Cmd {
			return exec.CommandContext(ctx, "go", "install", paqueteDelBinario)
		},
	},
	{
		variable: variableV1,
		carpeta:  versionV1,
		orden: func(ctx context.Context) *exec.Cmd {
			return exec.CommandContext(ctx, "go", "install", ldflagsV1, paqueteDelBinario)
		},
	},
	{
		variable: variableV2,
		carpeta:  versionV2,
		orden: func(ctx context.Context) *exec.Cmd {
			return exec.CommandContext(ctx, "go", "install", ldflagsV2, paqueteDelBinario)
		},
	},
	{
		variable: variableSinEnlaces,
		carpeta:  "sin-enlaces",
		orden: func(ctx context.Context) *exec.Cmd {
			return exec.CommandContext(ctx, "go", "install", ldflagsSinEnlaces, paqueteDelBinario)
		},
	},
}

// entorno es lo que TestMain deja preparado para el e2e: la ruta del binario de
// desarrollo ya construido, las variables que comparten todos los guiones, si
// el origen es el del snapshot y el fallo que, de haberlo, impidió prepararlo.
//
// El fallo se **guarda** en lugar de terminar el proceso porque los tests del
// kernel que viven en este mismo binario de test no necesitan ningún ejecutable:
// quien no pudo hacer su trabajo es el e2e, y es el e2e quien lo cuenta con un
// fallo atribuido. Terminar aquí convertiría un problema de un test en el
// silencio de todos.
var entorno struct {
	binario   string
	variables map[string]string
	dist      bool
	err       error
}

// TestMain construye los binarios de e2e y el origen de release local en un
// directorio temporal, antepone el de desarrollo al PATH y **retorna**.
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

// preparar construye los binarios y el origen y deja el PATH listo, y devuelve
// lo que hay que deshacer al terminar. Nada de lo que falle aquí termina el
// proceso: se anota en `entorno` y lo cuenta el test que necesitaba el binario.
//
// KITLEGAL_DIST se lee aquí y solo aquí: decide de dónde sale el origen y que
// TestEntregaDelHito exija algún guion instalador- (contracts/arnes-e2e.md §5).
// Vacía cuenta como sin definir.
func preparar() func() {
	dist := os.Getenv(variableDist)
	entorno.dist = dist != ""

	temporal, err := os.MkdirTemp("", "kitlegal-e2e-")
	if err != nil {
		entorno.err = fmt.Errorf("e2e: no se pudo crear el directorio temporal: %w", err)

		return func() {}
	}

	entorno.variables, entorno.err = prepararEn(temporal, dist)
	entorno.binario = entorno.variables[variableDelBinario]

	return func() {
		if err := retirarTemporal(temporal); err != nil {
			// Ya no queda ningún test al que contárselo —m.Run ha terminado—, así
			// que el aviso va al único canal que sigue vivo. Callarlo dejaría un
			// temporal huérfano sin rastro de quién lo dejó.
			log.Printf("e2e: no se pudo borrar el directorio temporal %s: %v", temporal, err)
		}
	}
}

// prepararEn construye en el temporal los cuatro binarios de
// contracts/arnes-e2e.md §2 y el origen de release local de §5, antepone el de
// desarrollo al PATH y devuelve las variables de §3 que comparten todos los
// guiones. Con dist, el origen es el del snapshot; sin él, el del binario
// v0.1.0.
func prepararEn(temporal, dist string) (map[string]string, error) {
	variables, err := construirBinarios(temporal)
	if err != nil {
		return nil, err
	}

	if err := anteponerAlPath(filepath.Dir(variables[variableDelBinario])); err != nil {
		return nil, err
	}

	destino := filepath.Join(temporal, "origen")

	var origen origenDeRelease
	if dist == "" {
		origen, err = origenDelBinario(destino, variables[variableV1])
	} else {
		origen, err = origenDelSnapshot(destino, dist)
	}

	if err != nil {
		return nil, err
	}

	skills, err := filepath.Abs(skillsDelRepositorio)
	if err != nil {
		return nil, fmt.Errorf("e2e: no se pudo resolver %s: %w", skillsDelRepositorio, err)
	}

	instalador, err := filepath.Abs(instaladorDelRepositorio)
	if err != nil {
		return nil, fmt.Errorf("e2e: no se pudo resolver %s: %w", instaladorDelRepositorio, err)
	}

	variables[variableSkills] = skills
	variables[variableInstalador] = instalador
	variables[variableOrigen] = origen.directorio
	variables[variableVersionDelOrigen] = origen.version
	variables[variableArchivoDelOrigen] = origen.archivo

	for _, proxy := range variablesDeProxy {
		variables[proxy] = proxyCerrado
	}

	for _, excepcion := range variablesSinProxy {
		variables[excepcion] = ""
	}

	return variables, nil
}

// construirBinarios construye a la vez, cada uno en su carpeta del temporal,
// los cuatro binarios del arnés, y devuelve la ruta absoluta de cada uno por su
// variable. Construirlos a la vez solo reparte la espera: la caché de
// compilación de Go admite varias órdenes a la vez, y cada una escribe en su
// propia carpeta.
func construirBinarios(temporal string) (map[string]string, error) {
	var (
		grupo   sync.WaitGroup
		errores [len(binariosDelArnes)]error
	)

	rutas := make(map[string]string, len(binariosDelArnes))

	for i, binario := range binariosDelArnes {
		carpeta := filepath.Join(temporal, binario.carpeta)
		rutas[binario.variable] = filepath.Join(carpeta, nombreDelBinario)

		grupo.Go(func() { errores[i] = construir(binario, carpeta) })
	}

	grupo.Wait()

	if err := errors.Join(errores[:]...); err != nil {
		return nil, err
	}

	return rutas, nil
}

// construir compila la raíz de composición del binario de e2e con la orden del
// binario y la deja en la carpeta que se le da, con el nombre kitlegal.
//
// La carpeta viaja en el entorno, por GOBIN, y no como argumento de `-o`: así
// la orden queda escrita **entera con constantes**, que es lo que el análisis de
// seguridad exige de un subproceso y lo que evita que este fichero necesite una
// excepción de lint. Lo que se construye es el binario real —no un arnés que
// simule el despacho—, que es lo único capaz de sostener el enlace simbólico de
// SC-003 (research.md D20).
//
// No toca la red: cuando `go test` llega hasta aquí, lo que el módulo necesita
// está ya en la caché de módulos.
func construir(binario binarioDelArnes, carpeta string) error {
	if err := os.Mkdir(carpeta, 0o750); err != nil {
		return fmt.Errorf("e2e: no se pudo crear la carpeta de %s: %w", binario.variable, err)
	}

	orden := binario.orden(context.Background())
	orden.Env = append(os.Environ(), "GOBIN="+carpeta)

	if salida, err := orden.CombinedOutput(); err != nil {
		return fmt.Errorf("e2e: no se pudo construir %s para %s: %w: %s",
			paqueteDelBinario, binario.variable, err, salida)
	}

	// El ejecutable se instala con el nombre del directorio del paquete, y el
	// despacho multicall lee precisamente ese nombre: renombrarlo es lo que hace
	// que los guiones puedan invocarlo como «kitlegal» (FR-002, FR-003). Con
	// ese nombre, que no es el de ningún applet, el despacho toma el applet del
	// primer argumento también cuando un guion lo invoca por su ruta absoluta.
	if err := os.Rename(filepath.Join(carpeta, nombreInstalado), filepath.Join(carpeta, nombreDelBinario)); err != nil {
		return fmt.Errorf("e2e: no se pudo nombrar el binario de %s como %s: %w",
			binario.variable, nombreDelBinario, err)
	}

	return nil
}

// retirarTemporal borra el temporal del arnés entero, devolviendo antes a su
// propietario el permiso de escritura en cada directorio: el origen de release
// es de solo lectura y, sin él, no se podría vaciar.
func retirarTemporal(temporal string) error {
	return errors.Join(hacerEscribible(temporal), os.RemoveAll(temporal))
}

// hacerEscribible devuelve a su propietario el permiso de escritura en cada
// directorio del árbol, sin seguir enlaces. Opera a través de un os.Root: nada
// de lo que haga puede salir del árbol.
func hacerEscribible(directorio string) (err error) {
	raiz, err := os.OpenRoot(directorio)
	if err != nil {
		return fmt.Errorf("e2e: no se pudo abrir %s para hacerlo escribible: %w", directorio, err)
	}

	defer func() { err = errors.Join(err, raiz.Close()) }()

	return fs.WalkDir(raiz.FS(), ".", func(ruta string, entrada fs.DirEntry, err error) error {
		if err != nil || !entrada.IsDir() {
			return err
		}

		return raiz.Chmod(filepath.FromSlash(ruta), 0o700)
	})
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

// TestEntregaDelHito ejecuta los guiones que describen la entrega de cada hito
// —de la invocación literal del hito, el enlace simbólico, la ayuda, el código
// 2 con argumentos malos, la autodescripción, el verbo obligatorio del segundo
// applet y la salida estándar con solo el documento JSON (FR-055, SC-001,
// SC-002, SC-003, SC-009 de H1) a los h19- del applet skills, el aviso e
// install.sh (FR-143, SC-001 a SC-014, SC-017 de H19)— contra los binarios y el
// origen que preparó TestMain, con las variables y las órdenes de
// contracts/arnes-e2e.md.
//
// Con KITLEGAL_DIST, antes de ejecutar ningún guion, exige que haya alguno
// instalador- que ejecutar contra el snapshot: sin él, make snapshot-check
// pasaría en vacío (§5; research.md D28).
//
// RequireExplicitExec obliga a que cada orden externa de un guion se escriba con
// `exec`: sin él, una línea mal escrita podría ejecutar un programa de la máquina
// creyendo que ejecuta una orden integrada. Ningún guion accede a la red —los
// proxies de cada guion están cerrados (FR-146)— y ninguno escribe fuera del
// directorio de trabajo que testscript le da y borra: boe responde desde la
// copia de sus grabaciones que Setup deja en ese directorio, y su caché vive en
// él (FR-114, research.md D13 de H4).
func TestEntregaDelHito(t *testing.T) {
	t.Parallel()

	if entorno.dist {
		require.NoError(t, hayGuionesDelInstalador(directorioDeGuiones))
	}

	require.NoError(t, entorno.err)

	testscript.Run(t, testscript.Params{
		Dir:                 directorioDeGuiones,
		RequireExplicitExec: true,
		Setup: func(env *testscript.Env) error {
			for variable, valor := range entorno.variables {
				env.Setenv(variable, valor)
			}

			env.Setenv(cache.VariableDirectorio, filepath.Join(env.WorkDir, directorioDeLaCache))

			return copiarGrabaciones(env.WorkDir)
		},
		Cmds: map[string]func(ts *testscript.TestScript, neg bool, args []string){
			ordenCronometra: cronometra,
			ordenArbol:      listarArbol,
		},
	})
}

// errSinGuionesDelInstalador es el fallo de TestEntregaDelHito con KITLEGAL_DIST
// cuando no hay ningún guion instalador- que ejecutar, con el mensaje literal de
// contracts/arnes-e2e.md §5.
var errSinGuionesDelInstalador = errors.New(variableDist + ": ningún guion instalador- que ejecutar")

// hayGuionesDelInstalador dice, sin error, que en el directorio hay al menos un
// guion —un fichero .txtar o .txt, lo que testscript ejecuta— cuyo nombre, sin
// la extensión, contiene «instalador-»: la copia momentánea zz-instalador-… del
// bucle de tareas o el activado como h19-instalador-… (research.md D28).
func hayGuionesDelInstalador(directorio string) error {
	entradas, err := os.ReadDir(directorio)
	if err != nil {
		return fmt.Errorf("e2e: no se pudo listar %s: %w", directorio, err)
	}

	for _, entrada := range entradas {
		nombre, esGuion := strings.CutSuffix(entrada.Name(), ".txtar")
		if !esGuion {
			nombre, esGuion = strings.CutSuffix(entrada.Name(), ".txt")
		}

		if esGuion && strings.Contains(nombre, "instalador-") {
			return nil
		}
	}

	return errSinGuionesDelInstalador
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

// Los motivos por los que arbol hace fallar un guion además de los del disco: un
// uso con otro número de argumentos y un directorio que no es un directorio real.
// Son centinelas para que TestArbol distinga cada rama sin leer mensajes.
var (
	errUsoDeArbol         = errors.New("uso: " + ordenArbol + " <directorio>")
	errNoEsDirectorioReal = errors.New(ordenArbol + ": no es un directorio real")
)

// listarArbol es la orden `arbol <directorio>` de los guiones
// (contracts/arnes-e2e.md §4): escribe en la salida estándar de la orden —la
// que ven stdout, cp stdout y cmp stdout— una línea por cada entrada por debajo
// del directorio, sin seguir enlaces y en orden de bytes de la ruta, y hace
// fallar el guion si el directorio no existe o no es un directorio real.
//
// Pide la salida estándar siempre, también sin ninguna línea que escribir: una
// orden propia que no la pide deja a las siguientes la de la orden anterior, y
// `! stdout .` sobre un directorio vacío pasaría o fallaría por lo que escribió
// otra.
//
// No admite «!»: negar un listado no dice qué se esperaba, y un guion que lo
// escribiera pasaría por la razón equivocada.
func listarArbol(ts *testscript.TestScript, neg bool, args []string) {
	if neg {
		ts.Fatalf("%s no admite «!»: el listado se escribe o hace fallar el guion", ordenArbol)
	}

	salida := ts.Stdout()

	lineas, err := arbolDeLosArgumentos(args, ts.MkAbs)
	ts.Check(err)

	_, err = io.WriteString(salida, lineas)
	ts.Check(err)
}

// arbolDeLosArgumentos interpreta los argumentos de arbol, resuelve la ruta con
// absoluta —la del directorio actual del guion— y devuelve sus líneas. Un uso
// incorrecto no resuelve ninguna ruta.
func arbolDeLosArgumentos(args []string, absoluta func(string) string) (string, error) {
	if len(args) != 1 {
		return "", fmt.Errorf("%w: recibió %q", errUsoDeArbol, args)
	}

	return arbolDe(absoluta(args[0]))
}

// arbolDe son las líneas de arbol del directorio, que tiene que existir y ser un
// directorio real: un enlace a un directorio no lo es. Lo lista a través de un
// os.Root, que no deja que nada de lo que lee salga de él.
func arbolDe(directorio string) (lineas string, err error) {
	estado, err := os.Lstat(directorio)
	if err != nil {
		return "", fmt.Errorf("%s: %w", ordenArbol, err)
	}

	if !estado.IsDir() {
		return "", fmt.Errorf("%w: %s", errNoEsDirectorioReal, directorio)
	}

	raiz, err := os.OpenRoot(directorio)
	if err != nil {
		return "", fmt.Errorf("%s: %w", ordenArbol, err)
	}

	defer func() { err = errors.Join(err, raiz.Close()) }()

	return arbolDelSistemaDeFicheros(raiz.FS())
}

// entradaDelArbol es una línea de arbol junto a la ruta por la que se ordena.
type entradaDelArbol struct {
	ruta  string
	linea string
}

// arbolDelSistemaDeFicheros son las líneas de arbol de cada entrada del sistema
// de ficheros por debajo de su raíz, cada una terminada en un salto de línea,
// en orden de bytes de la ruta entera —que no es el del recorrido: el recorrido
// entra en «a» antes de pasar a «a-b», y «a-b» va antes que «a/b»—.
//
// El recorrido no sigue enlaces: un enlace a un directorio es una línea `l`, y
// lo que hay dentro de su destino no se lista.
func arbolDelSistemaDeFicheros(sistema fs.FS) (string, error) {
	var entradas []entradaDelArbol

	err := fs.WalkDir(sistema, ".", func(ruta string, entrada fs.DirEntry, err error) error {
		if err != nil || ruta == "." {
			return err
		}

		linea, err := lineaDelArbol(sistema, ruta, entrada)
		if err != nil {
			return err
		}

		entradas = append(entradas, entradaDelArbol{ruta: ruta, linea: linea})

		return nil
	})
	if err != nil {
		return "", fmt.Errorf("%s: %w", ordenArbol, err)
	}

	slices.SortFunc(entradas, func(a, b entradaDelArbol) int { return strings.Compare(a.ruta, b.ruta) })

	var lineas strings.Builder
	for _, entrada := range entradas {
		lineas.WriteString(entrada.linea)
		lineas.WriteByte('\n')
	}

	return lineas.String(), nil
}

// lineaDelArbol es la línea de una entrada según su tipo (contracts/arnes-e2e.md
// §4): `d` directorio, `f` fichero regular con sus permisos en octal y la
// huella SHA-256 de su contenido, `l` enlace simbólico con su destino literal,
// `p` tubería con nombre y `o` cualquier otra cosa.
func lineaDelArbol(sistema fs.FS, ruta string, entrada fs.DirEntry) (string, error) {
	switch tipo := entrada.Type(); {
	case tipo.IsDir():
		return "d " + ruta, nil
	case tipo&fs.ModeSymlink != 0:
		destino, err := fs.ReadLink(sistema, ruta)
		if err != nil {
			return "", err
		}

		return "l " + ruta + " -> " + destino, nil
	case tipo.IsRegular():
		estado, err := entrada.Info()
		if err != nil {
			return "", err
		}

		huella, err := huellaDelFichero(sistema, ruta)
		if err != nil {
			return "", err
		}

		return fmt.Sprintf("f %s %03o %s", ruta, estado.Mode().Perm(), huella), nil
	case tipo&fs.ModeNamedPipe != 0:
		return "p " + ruta, nil
	default:
		return "o " + ruta, nil
	}
}

// huellaDelFichero es la huella SHA-256 en hexadecimal del fichero regular de la
// ruta, leído por partes: los binarios que instala install.sh no tienen por
// qué caber enteros en memoria para compararse.
func huellaDelFichero(sistema fs.FS, ruta string) (huella string, err error) {
	fichero, err := sistema.Open(ruta)
	if err != nil {
		return "", err
	}

	defer func() { err = errors.Join(err, fichero.Close()) }()

	resumen := sha256.New()
	if _, err := io.Copy(resumen, fichero); err != nil {
		return "", err
	}

	return hex.EncodeToString(resumen.Sum(nil)), nil
}

// huellaSHA256 es la huella SHA-256 en hexadecimal del contenido, la de las
// líneas de checksums.txt.
func huellaSHA256(contenido []byte) string {
	resumen := sha256.Sum256(contenido)

	return hex.EncodeToString(resumen[:])
}

// Los nombres del origen de release local (contracts/arnes-e2e.md §5;
// research.md D22).
const (
	// archivoDeLaPlataforma es el archivo de la release para la plataforma que
	// ejecuta el test, sin versión en el nombre.
	archivoDeLaPlataforma = nombreDelBinario + "_" + runtime.GOOS + "_" + runtime.GOARCH + ".tar.gz"

	// nombreDelSenuelo es el de la línea señuelo de checksums.txt: contiene el
	// del archivo, para que una búsqueda por subcadena o por posición falle.
	nombreDelSenuelo = archivoDeLaPlataforma + ".sbom.json"

	// ficheroDeChecksums y metadatosDelSnapshot son los ficheros de goreleaser
	// que el origen sirve y del que saca la versión del snapshot.
	ficheroDeChecksums   = "checksums.txt"
	metadatosDelSnapshot = "metadata.json"

	// descargaDeLaUltima y prefijoDeDescarga son las dos rutas de descarga, sin
	// versión y con ella, relativas a la base que install.sh recibe en
	// KITLEGAL_INSTALL_URL.
	descargaDeLaUltima = "latest/download"
	prefijoDeDescarga  = "download/v"

	// permisosDeSoloLectura son los de cada fichero del origen, y
	// permisosDeCarpetaDeSoloLectura, los de cada directorio: los guiones lo
	// leen, y para alterarlo lo copian antes (§3).
	permisosDeSoloLectura          = 0o400
	permisosDeCarpetaDeSoloLectura = 0o500
)

// errMetadatosDelSnapshot es un metadata.json del snapshot que no da la versión
// de la release: no es JSON, o su version no tiene forma SemVer o empieza por v
// (KITLEGAL_ORIGEN_VERSION va sin v, §3).
var errMetadatosDelSnapshot = errors.New(variableDist + ": el metadata.json no da una versión SemVer sin v")

// origenDeRelease es un origen de release local construido: dónde está —lo que
// install.sh recibe como file://<directorio>— y qué versión y qué archivo
// sirve.
type origenDeRelease struct {
	directorio string
	version    string
	archivo    string
}

// publicacion es lo que un origen sirve en sus dos rutas de descarga: la
// versión de su release, sin v, el archivo de la plataforma y los checksums.
type publicacion struct {
	version   string
	archivo   []byte
	checksums []byte
}

// origenDelBinario construye en destino, que no existe, el origen de make ci
// (contracts/arnes-e2e.md §5): el binario de la ruta, que es el de versión
// v0.1.0, empaquetado como el miembro kitlegal en la raíz del archivo de la
// plataforma, y checksums.txt con la línea correcta del archivo y la señuelo.
func origenDelBinario(destino, binario string) (origenDeRelease, error) {
	archivo, err := archivoConElBinario(binario)
	if err != nil {
		return origenDeRelease{}, err
	}

	return publicarOrigen(destino, publicacion{
		version:   strings.TrimPrefix(versionV1, "v"),
		archivo:   archivo,
		checksums: checksumsConSenuelo(archivo),
	})
}

// archivoConElBinario es un tar.gz cuyo único miembro es el fichero de la ruta,
// en la raíz, con el nombre kitlegal y permiso de ejecución, como lo empaqueta
// goreleaser (contracts/release.md §2). La fecha del miembro es la del fichero,
// también como goreleaser: una fecha nula haría quejarse a algunos tar.
func archivoConElBinario(binario string) (archivo []byte, err error) {
	fichero, err := os.Open(filepath.Clean(binario))
	if err != nil {
		return nil, fmt.Errorf("e2e: no se pudo abrir %s para empaquetarlo: %w", binario, err)
	}

	defer func() { err = errors.Join(err, fichero.Close()) }()

	estado, err := fichero.Stat()
	if err != nil {
		return nil, fmt.Errorf("e2e: no se pudo leer %s para empaquetarlo: %w", binario, err)
	}

	var empaquetado bytes.Buffer

	// La compresión más rápida: el archivo no se publica, y el binario es de
	// decenas de megas.
	comprimido, err := gzip.NewWriterLevel(&empaquetado, gzip.BestSpeed)
	if err != nil {
		return nil, fmt.Errorf("e2e: no se pudo comprimir %s: %w", binario, err)
	}

	tarDelArchivo := tar.NewWriter(comprimido)

	cabecera := &tar.Header{
		Typeflag: tar.TypeReg,
		Name:     nombreDelBinario,
		Mode:     0o755,
		Size:     estado.Size(),
		ModTime:  estado.ModTime(),
	}

	if err := tarDelArchivo.WriteHeader(cabecera); err != nil {
		return nil, fmt.Errorf("e2e: no se pudo empaquetar %s: %w", binario, err)
	}

	if _, err := io.Copy(tarDelArchivo, fichero); err != nil {
		return nil, fmt.Errorf("e2e: no se pudo empaquetar %s: %w", binario, err)
	}

	if err := errors.Join(tarDelArchivo.Close(), comprimido.Close()); err != nil {
		return nil, fmt.Errorf("e2e: no se pudo cerrar el archivo de %s: %w", binario, err)
	}

	return empaquetado.Bytes(), nil
}

// checksumsConSenuelo es el checksums.txt del origen de make ci, con el formato
// de goreleaser —`<sha256>  <nombre>`, ordenado por nombre—: la línea correcta
// del archivo y, detrás, la señuelo, cuyo nombre contiene el del archivo y cuya
// huella, la de ese nombre, no es la del archivo.
func checksumsConSenuelo(archivo []byte) []byte {
	return fmt.Appendf(nil, "%s  %s\n%s  %s\n",
		huellaSHA256(archivo), archivoDeLaPlataforma,
		huellaSHA256([]byte(nombreDelSenuelo)), nombreDelSenuelo)
}

// origenDelSnapshot construye en destino, que no existe, el origen de make
// snapshot-check (contracts/arnes-e2e.md §5): el archivo de la plataforma y
// checksums.txt del dist/ de la ruta, tal cual, y la versión de su
// metadata.json. Una ruta relativa se resuelve contra el directorio de este
// paquete. Lee todo lo que necesita antes de crear nada: un snapshot al que le
// falta algo no deja ningún origen.
func origenDelSnapshot(destino, dist string) (origenDeRelease, error) {
	absoluta, err := filepath.Abs(dist)
	if err != nil {
		return origenDeRelease{}, fmt.Errorf("%s: no se pudo resolver %s: %w", variableDist, dist, err)
	}

	servida, err := publicacionDelSnapshot(os.DirFS(absoluta))
	if err != nil {
		return origenDeRelease{}, fmt.Errorf("%s=%s: %w", variableDist, absoluta, err)
	}

	return publicarOrigen(destino, servida)
}

// publicacionDelSnapshot es lo que el origen sirve de un dist/: la versión de
// metadata.json, que tiene que tener forma SemVer y no empezar por v, y el
// archivo de la plataforma y checksums.txt, tal cual.
func publicacionDelSnapshot(dist fs.FS) (publicacion, error) {
	metadatos, err := fs.ReadFile(dist, metadatosDelSnapshot)
	if err != nil {
		return publicacion{}, err
	}

	var leidos struct {
		Version string `json:"version"`
	}

	if err := json.Unmarshal(metadatos, &leidos); err != nil {
		return publicacion{}, fmt.Errorf("%w: %w", errMetadatosDelSnapshot, err)
	}

	if !instalacion.FormaSemVer(leidos.Version) || strings.HasPrefix(leidos.Version, "v") {
		return publicacion{}, fmt.Errorf("%w: version %q", errMetadatosDelSnapshot, leidos.Version)
	}

	archivo, err := fs.ReadFile(dist, archivoDeLaPlataforma)
	if err != nil {
		return publicacion{}, err
	}

	checksums, err := fs.ReadFile(dist, ficheroDeChecksums)
	if err != nil {
		return publicacion{}, err
	}

	return publicacion{version: leidos.Version, archivo: archivo, checksums: checksums}, nil
}

// publicarOrigen crea en destino, que no existe, la disposición de las URL de
// descarga de GitHub (research.md D22) con la publicación en sus dos rutas,
// latest/download y download/v<versión>, y deja el origen entero de solo
// lectura. Escribe a través de un os.Root: nada de lo que escribe puede salir de
// destino, sea cual sea la versión.
func publicarOrigen(destino string, servida publicacion) (origen origenDeRelease, err error) {
	if err := os.Mkdir(destino, 0o700); err != nil {
		return origenDeRelease{}, fmt.Errorf("e2e: no se pudo crear el origen %s: %w", destino, err)
	}

	raiz, err := os.OpenRoot(destino)
	if err != nil {
		return origenDeRelease{}, fmt.Errorf("e2e: no se pudo abrir el origen %s: %w", destino, err)
	}

	defer func() { err = errors.Join(err, raiz.Close()) }()

	descargas := []string{descargaDeLaUltima, prefijoDeDescarga + servida.version}

	for _, descarga := range descargas {
		carpeta := filepath.FromSlash(descarga)

		err := errors.Join(
			raiz.MkdirAll(carpeta, 0o700),
			raiz.WriteFile(filepath.Join(carpeta, archivoDeLaPlataforma), servida.archivo, permisosDeSoloLectura),
			raiz.WriteFile(filepath.Join(carpeta, ficheroDeChecksums), servida.checksums, permisosDeSoloLectura),
		)
		if err != nil {
			return origenDeRelease{}, fmt.Errorf("e2e: no se pudo publicar %s en el origen %s: %w", descarga, destino, err)
		}
	}

	// De solo lectura también cada directorio, para que ningún guion pueda
	// crear, retirar ni sustituir nada en el origen que comparten todos.
	for _, carpeta := range slices.Concat(descargas, []string{"latest", "download", "."}) {
		if err := raiz.Chmod(filepath.FromSlash(carpeta), permisosDeCarpetaDeSoloLectura); err != nil {
			return origenDeRelease{}, fmt.Errorf("e2e: no se pudo dejar de solo lectura el origen %s: %w", destino, err)
		}
	}

	return origenDeRelease{directorio: destino, version: servida.version, archivo: archivoDeLaPlataforma}, nil
}

// Las huellas SHA-256 de los contenidos con los que TestArbol y
// TestOrigenDeRelease llenan sus ficheros, escritas a mano para que el formato
// no se compruebe con el mismo cálculo que lo produce.
const (
	huellaDeHola   = "133ee989293f92736301280c6f14c89d521200c17dcdcecca30cd20705332d44"
	huellaDeDentro = "a1b165a39ca1690df074796d606892f7d0919a1b55703f320162d1c109af7bd4"
	huellaVacia    = "e3b0c44298fc1c149afbf4c8996fb92427ae41e4649b934ca495991b7852b855"
)

// TestArbol fija el contrato de la orden arbol sin lanzar ningún guion, como
// TestCronometra (contracts/arnes-e2e.md §4): la línea de cada tipo de entrada,
// los enlaces con su destino literal y sin seguirlos, el orden de bytes de la
// ruta entera —que no es el del recorrido— y los fallos del uso y de un
// directorio que no existe o no es un directorio real.
func TestArbol(t *testing.T) {
	t.Parallel()

	t.Run("una línea por tipo de entrada", func(t *testing.T) {
		t.Parallel()

		// Un sistema de ficheros en memoria, porque una tubería con nombre no se
		// puede crear igual en todas las plataformas y un dispositivo o un socket
		// no se pueden crear en un test.
		sistema := fstest.MapFS{
			"carpeta/nota.md": {Data: []byte("hola\n"), Mode: 0o640},
			"ejecutable":      {Mode: 0o755},
			"enlace":          {Data: []byte("../fuera/del/arbol"), Mode: fs.ModeSymlink},
			"tuberia":         {Mode: fs.ModeNamedPipe},
			"socket":          {Mode: fs.ModeSocket},
			"dispositivo":     {Mode: fs.ModeDevice},
			"vacia":           {Mode: fs.ModeDir | 0o700},
			"solo-grupo":      {Mode: 0o040},
		}

		lineas, err := arbolDelSistemaDeFicheros(sistema)
		require.NoError(t, err)
		assert.Equal(t, "d carpeta\n"+
			"f carpeta/nota.md 640 "+huellaDeHola+"\n"+
			"o dispositivo\n"+
			"f ejecutable 755 "+huellaVacia+"\n"+
			"l enlace -> ../fuera/del/arbol\n"+
			"o socket\n"+
			"f solo-grupo 040 "+huellaVacia+"\n"+
			"p tuberia\n"+
			"d vacia\n", lineas)
	})

	t.Run("los enlaces se listan con su destino literal y no se siguen", func(t *testing.T) {
		t.Parallel()

		directorio := t.TempDir()
		require.NoError(t, os.Mkdir(filepath.Join(directorio, "real"), 0o700))
		require.NoError(t, os.WriteFile(filepath.Join(directorio, "real", "dentro.md"), []byte("dentro\n"), 0o600))
		require.NoError(t, os.Symlink("real", filepath.Join(directorio, "a-carpeta")))
		require.NoError(t, os.Symlink("real/dentro.md", filepath.Join(directorio, "a-fichero")))
		require.NoError(t, os.Symlink("no-existe", filepath.Join(directorio, "colgando")))
		require.NoError(t, os.Symlink("../..", filepath.Join(directorio, "real", "arriba")))

		lineas, err := arbolDe(directorio)
		require.NoError(t, err)
		assert.Equal(t, "l a-carpeta -> real\n"+
			"l a-fichero -> real/dentro.md\n"+
			"l colgando -> no-existe\n"+
			"d real\n"+
			"l real/arriba -> ../..\n"+
			"f real/dentro.md 600 "+huellaDeDentro+"\n", lineas)
	})

	t.Run("en orden de bytes de la ruta entera y no en el del recorrido", func(t *testing.T) {
		t.Parallel()

		// El recorrido entra en «a» antes de pasar a «a-b»; el orden de bytes pone
		// «a-b» y «a.txt» antes que «a/b», porque «-» y «.» van antes que «/», y
		// las mayúsculas antes que las minúsculas.
		directorio := t.TempDir()
		require.NoError(t, os.Mkdir(filepath.Join(directorio, "a"), 0o700))

		for _, fichero := range []string{"a/b", "a-b", "a.txt", "B"} {
			require.NoError(t, os.WriteFile(filepath.Join(directorio, filepath.FromSlash(fichero)), nil, 0o600))
		}

		lineas, err := arbolDe(directorio)
		require.NoError(t, err)
		assert.Equal(t, "f B 600 "+huellaVacia+"\n"+
			"d a\n"+
			"f a-b 600 "+huellaVacia+"\n"+
			"f a.txt 600 "+huellaVacia+"\n"+
			"f a/b 600 "+huellaVacia+"\n", lineas)
	})

	t.Run("un directorio vacío no da ninguna línea", func(t *testing.T) {
		t.Parallel()

		lineas, err := arbolDe(t.TempDir())
		require.NoError(t, err)
		assert.Empty(t, lineas)
	})

	t.Run("la ruta relativa es la del directorio del guion", func(t *testing.T) {
		t.Parallel()

		directorio := t.TempDir()
		require.NoError(t, os.Mkdir(filepath.Join(directorio, "proyecto"), 0o700))
		require.NoError(t, os.WriteFile(filepath.Join(directorio, "proyecto", "nota.md"), []byte("hola\n"), 0o600))

		lineas, err := arbolDeLosArgumentos([]string{"proyecto"}, func(ruta string) string {
			return filepath.Join(directorio, ruta)
		})
		require.NoError(t, err)
		assert.Equal(t, "f nota.md 600 "+huellaDeHola+"\n", lineas)
	})

	t.Run("lo que no es un directorio real hace fallar la orden", func(t *testing.T) {
		t.Parallel()

		directorio := t.TempDir()
		require.NoError(t, os.WriteFile(filepath.Join(directorio, "fichero"), nil, 0o600))
		require.NoError(t, os.Mkdir(filepath.Join(directorio, "real"), 0o700))
		require.NoError(t, os.Symlink("real", filepath.Join(directorio, "enlace")))

		_, err := arbolDe(filepath.Join(directorio, "no-existe"))
		require.ErrorIs(t, err, fs.ErrNotExist)

		_, err = arbolDe(filepath.Join(directorio, "fichero"))
		require.ErrorIs(t, err, errNoEsDirectorioReal)

		_, err = arbolDe(filepath.Join(directorio, "enlace"))
		require.ErrorIs(t, err, errNoEsDirectorioReal, "un enlace a un directorio tampoco es un directorio real")
	})

	t.Run("un uso con otro número de argumentos no lista nada", func(t *testing.T) {
		t.Parallel()

		var resueltas []string

		absoluta := func(ruta string) string {
			resueltas = append(resueltas, ruta)

			return ruta
		}

		for _, args := range [][]string{nil, {".", "otro"}} {
			_, err := arbolDeLosArgumentos(args, absoluta)
			require.ErrorIs(t, err, errUsoDeArbol)
		}

		assert.Empty(t, resueltas, "un uso incorrecto no resuelve ninguna ruta")
	})
}

// TestOrigenDeRelease fija el origen de release local de contracts/arnes-e2e.md
// §5 construido en un temporal: con el binario, el archivo con él dentro como
// su único miembro y los checksums con su línea señuelo; con un snapshot
// sintético, el archivo y los checksums de dist/ tal cual y la versión de su
// metadata.json. En los dos, la misma disposición de las URL de descarga
// (research.md D22), de solo lectura; y lo que falta o no sirve del snapshot es
// un error que no deja nada.
func TestOrigenDeRelease(t *testing.T) {
	t.Parallel()

	t.Run("del binario", func(t *testing.T) {
		t.Parallel()

		binario := filepath.Join(t.TempDir(), nombreDelBinario)
		// Sin permiso de ejecución: el del miembro lo pone el archivo, no el
		// fichero.
		require.NoError(t, os.WriteFile(binario, []byte("#!/bin/sh\necho hola\n"), 0o600))

		destino := destinoDelOrigen(t)

		origen, err := origenDelBinario(destino, binario)
		require.NoError(t, err)
		assert.Equal(t, origenDeRelease{directorio: destino, version: "0.1.0", archivo: archivoDeLaPlataforma}, origen)

		archivo := leerDelOrigen(t, destino, "latest/download/"+archivoDeLaPlataforma)
		assert.Equal(t, []miembroDeArchivo{{nombre: nombreDelBinario, modo: 0o755, contenido: "#!/bin/sh\necho hola\n"}},
			miembrosDe(t, archivo), "el binario, con el nombre kitlegal, es el único miembro y está en la raíz")

		checksums := leerDelOrigen(t, destino, "latest/download/"+ficheroDeChecksums)
		lineas := strings.Split(strings.TrimSuffix(string(checksums), "\n"), "\n")
		require.Len(t, lineas, 2, "la línea del archivo y la señuelo")
		assert.Equal(t, huellaSHA256(archivo)+"  "+archivoDeLaPlataforma, lineas[0])

		huella, nombre, cortada := strings.Cut(lineas[1], "  ")
		require.True(t, cortada)
		assert.Equal(t, archivoDeLaPlataforma+".sbom.json", nombre, "el nombre señuelo contiene el del archivo")
		assert.Regexp(t, `\A[0-9a-f]{64}\z`, huella)
		assert.NotEqual(t, huellaSHA256(archivo), huella, "la señuelo lleva otra huella")

		exigirDisposicion(t, origen, archivo, checksums)
	})

	t.Run("del snapshot", func(t *testing.T) {
		t.Parallel()

		const version = "0.1.1-SNAPSHOT-0123abc"

		dist := snapshotSintetico(t, `{"project_name":"kitlegal","tag":"v0.1.0","previous_tag":"",`+
			`"version":"`+version+`","commit":"0123abc0123abc0123abc0123abc0123abc0123a",`+
			`"date":"2026-09-27T10:00:00Z","runtime":{"goos":"linux","goarch":"amd64"}}`)

		destino := destinoDelOrigen(t)

		origen, err := origenDelSnapshot(destino, dist)
		require.NoError(t, err)
		assert.Equal(t, origenDeRelease{directorio: destino, version: version, archivo: archivoDeLaPlataforma}, origen)

		exigirDisposicion(t, origen, []byte("archivo del snapshot"), []byte("checksums del snapshot\n"))
	})

	t.Run("del snapshot, lo que falta o no sirve es un error", func(t *testing.T) {
		t.Parallel()

		const metadatos = `{"version":"0.1.1-SNAPSHOT-0123abc"}`

		casos := []struct {
			nombre string
			// preparar deja el snapshot como el caso lo necesita y devuelve su ruta.
			preparar func(t *testing.T) string
			quiere   error
		}{
			{
				nombre:   "un dist que no existe",
				preparar: func(t *testing.T) string { t.Helper(); return filepath.Join(t.TempDir(), "dist") },
				quiere:   fs.ErrNotExist,
			},
			{
				nombre:   "sin metadata.json",
				preparar: func(t *testing.T) string { t.Helper(); return sinDelSnapshot(t, metadatos, metadatosDelSnapshot) },
				quiere:   fs.ErrNotExist,
			},
			{
				nombre:   "sin el archivo de la plataforma",
				preparar: func(t *testing.T) string { t.Helper(); return sinDelSnapshot(t, metadatos, archivoDeLaPlataforma) },
				quiere:   fs.ErrNotExist,
			},
			{
				nombre:   "sin checksums.txt",
				preparar: func(t *testing.T) string { t.Helper(); return sinDelSnapshot(t, metadatos, ficheroDeChecksums) },
				quiere:   fs.ErrNotExist,
			},
			{
				nombre:   "metadata.json que no es JSON",
				preparar: func(t *testing.T) string { t.Helper(); return snapshotSintetico(t, `{"version":`) },
				quiere:   errMetadatosDelSnapshot,
			},
			{
				nombre:   "sin versión",
				preparar: func(t *testing.T) string { t.Helper(); return snapshotSintetico(t, `{"tag":"v0.1.0"}`) },
				quiere:   errMetadatosDelSnapshot,
			},
			{
				nombre:   "una versión sin forma SemVer",
				preparar: func(t *testing.T) string { t.Helper(); return snapshotSintetico(t, `{"version":"0.1"}`) },
				quiere:   errMetadatosDelSnapshot,
			},
			{
				nombre:   "una versión con v",
				preparar: func(t *testing.T) string { t.Helper(); return snapshotSintetico(t, `{"version":"v0.1.0"}`) },
				quiere:   errMetadatosDelSnapshot,
			},
		}

		for _, caso := range casos {
			t.Run(caso.nombre, func(t *testing.T) {
				t.Parallel()

				destino := filepath.Join(t.TempDir(), "origen")

				_, err := origenDelSnapshot(destino, caso.preparar(t))
				require.ErrorIs(t, err, caso.quiere)

				_, err = os.Lstat(destino)
				require.ErrorIs(t, err, fs.ErrNotExist, "un snapshot que no sirve no deja ningún origen")
			})
		}
	})
}

// miembroDeArchivo es lo que TestOrigenDeRelease mira de cada miembro de un
// archivo tar.gz: su nombre, sus permisos y su contenido.
type miembroDeArchivo struct {
	nombre    string
	modo      int64
	contenido string
}

// miembrosDe son los miembros del archivo tar.gz, en el orden en que están.
func miembrosDe(t *testing.T, archivo []byte) []miembroDeArchivo {
	t.Helper()

	comprimido, err := gzip.NewReader(bytes.NewReader(archivo))
	require.NoError(t, err)

	empaquetado := tar.NewReader(comprimido)

	var miembros []miembroDeArchivo

	for {
		cabecera, err := empaquetado.Next()
		if errors.Is(err, io.EOF) {
			break
		}

		require.NoError(t, err)
		require.Equal(t, byte(tar.TypeReg), cabecera.Typeflag, "%s no es un fichero regular", cabecera.Name)

		contenido, err := io.ReadAll(empaquetado)
		require.NoError(t, err)

		miembros = append(miembros, miembroDeArchivo{nombre: cabecera.Name, modo: cabecera.Mode, contenido: string(contenido)})
	}

	require.NoError(t, comprimido.Close())

	return miembros
}

// destinoDelOrigen es la ruta, dentro de un temporal del test y todavía sin
// crear, donde construir un origen. Como el origen queda de solo lectura, le
// devuelve el permiso de escritura antes de que el test retire su temporal.
func destinoDelOrigen(t *testing.T) string {
	t.Helper()

	destino := filepath.Join(t.TempDir(), "origen")

	t.Cleanup(func() { assert.NoError(t, hacerEscribible(destino)) })

	return destino
}

// leerDelOrigen es el contenido del fichero de la ruta, con barras, dentro del
// origen.
func leerDelOrigen(t *testing.T, origen, ruta string) []byte {
	t.Helper()

	contenido, err := fs.ReadFile(os.DirFS(origen), ruta)
	require.NoError(t, err)

	return contenido
}

// exigirDisposicion comprueba que el origen tiene exactamente la disposición de
// las URL de descarga de research.md D22, con el archivo y los checksums dados
// en las dos, y que es de solo lectura: ningún fichero ni directorio admite
// escritura.
func exigirDisposicion(t *testing.T, origen origenDeRelease, archivo, checksums []byte) {
	t.Helper()

	descarga := "download/v" + origen.version

	lineas, err := arbolDe(origen.directorio)
	require.NoError(t, err)
	assert.Equal(t, "d download\n"+
		"d "+descarga+"\n"+
		"f "+descarga+"/checksums.txt 400 "+huellaSHA256(checksums)+"\n"+
		"f "+descarga+"/"+origen.archivo+" 400 "+huellaSHA256(archivo)+"\n"+
		"d latest\n"+
		"d latest/download\n"+
		"f latest/download/checksums.txt 400 "+huellaSHA256(checksums)+"\n"+
		"f latest/download/"+origen.archivo+" 400 "+huellaSHA256(archivo)+"\n", lineas)

	for _, carpeta := range []string{".", "download", descarga, "latest", "latest/download"} {
		estado, err := os.Lstat(filepath.Join(origen.directorio, filepath.FromSlash(carpeta)))
		require.NoError(t, err)
		assert.Equalf(t, fs.FileMode(0o500), estado.Mode().Perm(), "%s tiene que ser de solo lectura", carpeta)
	}
}

// snapshotSintetico deja en un temporal lo que make release deja en dist/ para
// el origen: el metadata.json dado, el archivo de la plataforma que ejecuta el
// test y checksums.txt; y, alrededor, lo que el origen no copia —el archivo de
// otra plataforma y install.sh—. Devuelve su ruta.
func snapshotSintetico(t *testing.T, metadatos string) string {
	t.Helper()

	dist := t.TempDir()

	ficheros := map[string]string{
		metadatosDelSnapshot:        metadatos,
		archivoDeLaPlataforma:       "archivo del snapshot",
		ficheroDeChecksums:          "checksums del snapshot\n",
		"kitlegal_plan9_386.tar.gz": "archivo de otra plataforma",
		"install.sh":                "#!/bin/sh\n",
	}

	for nombre, contenido := range ficheros {
		require.NoError(t, os.WriteFile(filepath.Join(dist, nombre), []byte(contenido), 0o600))
	}

	return dist
}

// sinDelSnapshot es un snapshot sintético con los metadatos dados al que le
// falta el fichero nombrado.
func sinDelSnapshot(t *testing.T, metadatos, falta string) string {
	t.Helper()

	dist := snapshotSintetico(t, metadatos)
	require.NoError(t, os.Remove(filepath.Join(dist, falta)))

	return dist
}

// TestGuionesDelInstalador fija la garantía de contracts/arnes-e2e.md §5 para
// que make snapshot-check no pase en vacío: con KITLEGAL_DIST, tiene que haber
// en el directorio de los guiones al menos uno —un fichero .txtar o .txt, que
// es lo que testscript ejecuta— cuyo nombre contenga «instalador-», sea la
// copia momentánea o el activado; si no, el fallo con su mensaje literal.
func TestGuionesDelInstalador(t *testing.T) {
	t.Parallel()

	casos := []struct {
		nombre   string
		ficheros []string
		hay      bool
	}{
		{nombre: "ningún guion"},
		{nombre: "solo guiones de otras órdenes", ficheros: []string{"boe-codigos.txtar", "h19-skills-aviso.txtar"}},
		{nombre: "instalador sin el guion", ficheros: []string{"instalador.txtar", "h19-instalador.txt"}},
		{nombre: "con otra extensión", ficheros: []string{"instalador-correcto.md", "instalador-correcto.txtar.orig"}},
		{nombre: "la copia momentánea", ficheros: []string{"boe-codigos.txtar", "zz-instalador-correcto.txtar"}, hay: true},
		{nombre: "el activado", ficheros: []string{"h19-instalador-rechazos.txtar"}, hay: true},
		{nombre: "un guion .txt", ficheros: []string{"instalador-correcto.txt"}, hay: true},
	}

	for _, caso := range casos {
		t.Run(caso.nombre, func(t *testing.T) {
			t.Parallel()

			directorio := t.TempDir()

			for _, fichero := range caso.ficheros {
				require.NoError(t, os.WriteFile(filepath.Join(directorio, fichero), nil, 0o600))
			}

			err := hayGuionesDelInstalador(directorio)
			if caso.hay {
				require.NoError(t, err)

				return
			}

			require.ErrorIs(t, err, errSinGuionesDelInstalador)
			require.EqualError(t, err, "KITLEGAL_DIST: ningún guion instalador- que ejecutar")
		})
	}

	t.Run("un directorio de guiones que no existe", func(t *testing.T) {
		t.Parallel()

		err := hayGuionesDelInstalador(filepath.Join(t.TempDir(), "script"))
		require.ErrorIs(t, err, fs.ErrNotExist)
	})
}

// TestBinariosDelArnes fija la tabla de contracts/arnes-e2e.md §2 sobre los
// binarios que TestMain construye: cada uno está en una ruta absoluta y se llama
// kitlegal, dice su versión en «version», y solo el de KITLEGAL_SIN_ENLACES_BIN
// deja como copia la entrada de host que los demás enlazan (FR-024). Es lo que
// comprueba que cada -X llega: el enlazador de Go ignora sin avisar uno que
// nombra una variable que no existe.
func TestBinariosDelArnes(t *testing.T) {
	t.Parallel()

	require.NoError(t, entorno.err)

	casos := []struct {
		variable string
		version  string
		enlaza   bool
	}{
		{variable: variableDelBinario, version: "dev", enlaza: true},
		{variable: variableV1, version: versionV1, enlaza: true},
		{variable: variableV2, version: versionV2, enlaza: true},
		{variable: variableSinEnlaces, version: versionV1, enlaza: false},
	}

	for _, caso := range casos {
		t.Run(caso.variable, func(t *testing.T) {
			t.Parallel()

			binario := entorno.variables[caso.variable]
			require.True(t, filepath.IsAbs(binario), "%s tiene que ser una ruta absoluta: %q", caso.variable, binario)
			assert.Equal(t, nombreDelBinario, filepath.Base(binario))

			assert.Equal(t, "kitlegal "+caso.version, primeraLineaDeVersion(t, binario))

			proyecto := t.TempDir()
			require.NoError(t, os.Mkdir(filepath.Join(proyecto, ".claude"), 0o700))
			instalarBoeLegislacion(t, binario, proyecto)

			estado, err := os.Lstat(filepath.Join(proyecto, ".claude", "skills", "boe-legislacion"))
			require.NoError(t, err)

			if caso.enlaza {
				assert.Equal(t, fs.ModeSymlink, estado.Mode().Type(), "la entrada de host es un enlace")
			} else {
				assert.True(t, estado.IsDir(), "sin enlaces, la entrada de host es una copia")
			}
		})
	}

	assert.Equal(t, filepath.Dir(entorno.binario), strings.SplitN(os.Getenv("PATH"), string(os.PathListSeparator), 2)[0],
		"el binario de desarrollo va en el PATH, por delante de todo")
}

// primeraLineaDeVersion es la primera línea de lo que el binario escribe en
// «version».
func primeraLineaDeVersion(t *testing.T, binario string) string {
	t.Helper()

	salida, err := exec.CommandContext(t.Context(), binario, "version").Output()
	require.NoError(t, err)

	primera, _, _ := strings.Cut(string(salida), "\n")

	return primera
}

// instalarBoeLegislacion instala boe-legislacion con el binario en el proyecto,
// con un entorno vacío: el ámbito local no necesita nada de él, y así nada de
// la cuenta de quien ejecuta los tests entra en la prueba.
func instalarBoeLegislacion(t *testing.T, binario, proyecto string) {
	t.Helper()

	orden := exec.CommandContext(t.Context(), binario, "skills", "install", "boe-legislacion")
	orden.Dir = proyecto
	orden.Env = []string{}

	salida, err := orden.CombinedOutput()
	require.NoError(t, err, "%s", salida)
}

// TestVariablesDeLosGuiones fija las variables de contracts/arnes-e2e.md §3 que
// TestMain deja para todos los guiones: las rutas, absolutas; el archivo de la
// plataforma y la versión del origen; y los proxies cerrados.
func TestVariablesDeLosGuiones(t *testing.T) {
	t.Parallel()

	require.NoError(t, entorno.err)

	rutas := []string{
		variableDelBinario, variableV1, variableV2, variableSinEnlaces,
		variableSkills, variableInstalador, variableOrigen,
	}
	for _, variable := range rutas {
		assert.Truef(t, filepath.IsAbs(entorno.variables[variable]), "%s tiene que ser una ruta absoluta: %q",
			variable, entorno.variables[variable])
	}

	assert.FileExists(t, filepath.Join(entorno.variables[variableSkills], "boe-legislacion", "SKILL.md"))
	assert.Equal(t, filepath.Join("scripts", "install.sh"),
		filepath.Join(filepath.Base(filepath.Dir(entorno.variables[variableInstalador])),
			filepath.Base(entorno.variables[variableInstalador])))
	assert.DirExists(t, entorno.variables[variableOrigen])
	assert.Equal(t, archivoDeLaPlataforma, entorno.variables[variableArchivoDelOrigen])
	assert.True(t, instalacion.FormaSemVer(entorno.variables[variableVersionDelOrigen]))

	for _, proxy := range []string{"http_proxy", "https_proxy", "HTTP_PROXY", "HTTPS_PROXY", "ALL_PROXY"} {
		assert.Equal(t, "http://127.0.0.1:9", entorno.variables[proxy], proxy)
	}

	for _, excepcion := range []string{"NO_PROXY", "no_proxy"} {
		valor, definida := entorno.variables[excepcion]
		assert.True(t, definida, excepcion)
		assert.Empty(t, valor, excepcion)
	}
}
