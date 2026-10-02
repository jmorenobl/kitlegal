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
	"regexp"
	"runtime"
	"slices"
	"strings"
	"sync"
	"testing"
	"testing/fstest"
	"time"

	"github.com/rogpeppe/go-internal/testscript"
	"github.com/rogpeppe/go-internal/txtar"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/jmorenobl/kitlegal/internal/cache"
	"github.com/jmorenobl/kitlegal/internal/core/instalacion"
	"github.com/jmorenobl/kitlegal/internal/mcp/mcptest"
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

	// derivadasDelRepositorio son las grabaciones derivadas de las de H4, relativas
	// al directorio de este paquete: una carpeta por caso, con la grabación que
	// sustituye a la de H4 del mismo nombre; y directorioDeDerivadas, la carpeta
	// de $WORK en la que cada guion tiene su copia (contracts/arnes-e2e.md §3 de
	// H7; research.md D22 de H7).
	derivadasDelRepositorio = "testdata/derivadas"
	directorioDeDerivadas   = "derivadas"

	// ordenCronometra es el nombre con el que los guiones escriben la orden que
	// mide una invocación.
	ordenCronometra = "cronometra"

	// ordenArbol es el nombre con el que los guiones escriben la orden que
	// lista un directorio sin seguir enlaces, para comparar el disco byte a byte
	// (contracts/arnes-e2e.md §4).
	ordenArbol = "arbol"

	// ordenMCP es el nombre con el que los guiones escriben la orden que habla
	// con un servidor MCP por su entrada y su salida estándar
	// (contracts/arnes-e2e.md §3 de H21).
	ordenMCP = "mcp"
)

// Las versiones de los binarios de e2e que no son el de desarrollo, los relojes
// de los que fijan el instante y los -ldflags con que se construyen
// (contracts/arnes-e2e.md §2; research.md D24; contracts/arnes-e2e.md §2 y
// research.md D23 de H7). Son constantes, como el resto de la orden de
// construcción: ni el entorno ni ningún argumento deciden con qué se construye
// un binario.
const (
	versionV1 = "v0.1.0"
	versionV2 = "v0.2.0"

	// enlazadorQueFalla es el valor de la variable de cadena enlazador del
	// package main de e2e que elige el creador de enlaces que siempre falla, y
	// es el de la constante del mismo nombre de ejemplo/kitlegal-e2e/main.go.
	// El enlazador de Go ignora sin avisar un -X que nombra una variable que no
	// existe: lo que comprueba que este llega es TestBinariosDelArnes.
	enlazadorQueFalla = "falla"

	// relojT0, relojT1 y relojT8 son los valores de la variable de cadena reloj
	// del package main de e2e: el instante con el que la reproducción de boe
	// fecha lo que sirve y el reloj del applet graph. T1 es un día después de
	// T0, y T8, ocho, cuando la consulta de una semana hecha en T0 ya ha
	// caducado. Como con enlazador, lo que comprueba que el -X llega es
	// TestBinariosDelArnes.
	relojT0 = "2026-09-28T12:00:00Z"
	relojT1 = "2026-09-29T12:00:00Z"
	relojT8 = "2026-10-06T12:00:00Z"

	ldflagsV1         = "-ldflags=-X main.version=" + versionV1
	ldflagsV2         = "-ldflags=-X main.version=" + versionV2
	ldflagsSinEnlaces = ldflagsV1 + " -X main.enlazador=" + enlazadorQueFalla
	ldflagsT0         = "-ldflags=-X main.reloj=" + relojT0
	ldflagsT1         = "-ldflags=-X main.reloj=" + relojT1
	ldflagsT8         = "-ldflags=-X main.reloj=" + relojT8
)

// Las variables de entorno que el arnés deja en cada guion, además de
// KITLEGAL_BIN y KITLEGAL_CACHE_DIR (contracts/arnes-e2e.md §3), y la que elige
// el origen del snapshot.
const (
	variableV1               = "KITLEGAL_V1_BIN"
	variableV2               = "KITLEGAL_V2_BIN"
	variableSinEnlaces       = "KITLEGAL_SIN_ENLACES_BIN"
	variableT0               = "KITLEGAL_T0_BIN"
	variableT1               = "KITLEGAL_T1_BIN"
	variableT8               = "KITLEGAL_T8_BIN"
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

// binarioDelArnes es una de las siete construcciones del binario de e2e de
// contracts/arnes-e2e.md §2: el mismo paquete, con la versión, el creador de
// enlaces y el reloj que fijan sus -ldflags.
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

// binariosDelArnes son los siete que TestMain construye. El de desarrollo, sin
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
	{
		variable: variableT0,
		carpeta:  "t0",
		orden: func(ctx context.Context) *exec.Cmd {
			return exec.CommandContext(ctx, "go", "install", ldflagsT0, paqueteDelBinario)
		},
	},
	{
		variable: variableT1,
		carpeta:  "t1",
		orden: func(ctx context.Context) *exec.Cmd {
			return exec.CommandContext(ctx, "go", "install", ldflagsT1, paqueteDelBinario)
		},
	},
	{
		variable: variableT8,
		carpeta:  "t8",
		orden: func(ctx context.Context) *exec.Cmd {
			return exec.CommandContext(ctx, "go", "install", ldflagsT8, paqueteDelBinario)
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

// prepararEn construye en el temporal los siete binarios de
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
// los siete binarios del arnés, y devuelve la ruta absoluta de cada uno por su
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
//
// Los guiones que cronometran no van aquí sino en TestMedidasDeTiempo: esta
// prueba corre en paralelo con el resto del paquete y de los demás paquetes, y
// una medida de reloj en una máquina cargada mide la carga, no el programa.
func TestEntregaDelHito(t *testing.T) {
	t.Parallel()

	if entorno.dist {
		require.NoError(t, hayGuionesDelInstalador(directorioDeGuiones))
	}

	require.NoError(t, entorno.err)

	_, sinMedidas, err := guionesPorMedida(directorioDeGuiones)
	require.NoError(t, err)

	ejecutarGuiones(t, sinMedidas)
}

// TestMedidasDeTiempo ejecuta los guiones que cronometran —la cota de 200 ms de
// boe articulo desde la caché y la de territorio resolver (FR-117, SC-002 de
// H4; H6)— con el reloj de pared, como los escribió su hito, pero sin nada más
// en marcha: no es paralela, así que en su paquete corre antes que las pruebas
// paralelas, y make ci la ejecuta en su propio paso (test-tiempos), después de
// test y test-integration, que la saltan. Así la medida es del programa y no de
// lo que la máquina esté haciendo a la vez: en el runner de GitHub, con todos
// los paquetes de go test -race ./... en marcha, una invocación de 8 ms llegó a
// medir 236 ms.
//
//nolint:paralleltest // Mide con el reloj de pared: con otra prueba a la vez, mediría también lo suyo.
func TestMedidasDeTiempo(t *testing.T) {
	require.NoError(t, entorno.err)

	conMedidas, _, err := guionesPorMedida(directorioDeGuiones)
	require.NoError(t, err)
	require.NotEmpty(t, conMedidas, "sin guiones que cronometren, esta prueba pasaría en vacío")

	ejecutarGuiones(t, conMedidas)
}

// ejecutarGuiones ejecuta los guiones de ficheros —rutas relativas al paquete—
// con los binarios, las variables y las órdenes de contracts/arnes-e2e.md.
func ejecutarGuiones(t *testing.T, ficheros []string) {
	t.Helper()

	testscript.Run(t, testscript.Params{
		Files:               ficheros,
		RequireExplicitExec: true,
		Setup: func(env *testscript.Env) error {
			for variable, valor := range entorno.variables {
				env.Setenv(variable, valor)
			}

			env.Setenv(cache.VariableDirectorio, filepath.Join(env.WorkDir, directorioDeLaCache))

			// Con todas las variables ya puestas: es de donde la orden mcp saca
			// el entorno del programa que arranca.
			env.Values[claveDeLasVariables{}] = nombresDeLasVariables(env.Vars)

			if err := copiarGrabaciones(env.WorkDir); err != nil {
				return err
			}

			return copiarDerivadas(env.WorkDir)
		},
		Cmds: map[string]func(ts *testscript.TestScript, neg bool, args []string){
			ordenCronometra: cronometra,
			ordenArbol:      listarArbol,
			ordenMCP:        conversarMCP,
		},
	})
}

// ordenCronometraEnUnaLinea reconoce una línea de guion que invoca cronometra,
// con o sin sangría y con o sin condición delante ([unix] cronometra …).
var ordenCronometraEnUnaLinea = regexp.MustCompile(`(?m)^\s*(\[[^\]]*\]\s*)*` + ordenCronometra + `\s`)

// guionesPorMedida reparte los guiones del directorio —lo que testscript
// ejecutaría con Dir: los .txtar y los .txt— entre los que cronometran y los
// demás, cada lista en el orden del directorio. Solo mira el guion, no los
// ficheros que lleva dentro: una sección -- … -- que nombrara cronometra no es
// una orden.
func guionesPorMedida(directorio string) ([]string, []string, error) {
	entradas, err := os.ReadDir(directorio)
	if err != nil {
		return nil, nil, fmt.Errorf("e2e: no se pudo listar %s: %w", directorio, err)
	}

	var conMedidas, sinMedidas []string

	for _, entrada := range entradas {
		nombre := entrada.Name()
		if entrada.IsDir() || (filepath.Ext(nombre) != ".txtar" && filepath.Ext(nombre) != ".txt") {
			continue
		}

		ruta := filepath.Join(directorio, nombre)

		archivo, err := txtar.ParseFile(ruta)
		if err != nil {
			return nil, nil, fmt.Errorf("e2e: no se pudo leer el guion %s: %w", ruta, err)
		}

		if ordenCronometraEnUnaLinea.Match(archivo.Comment) {
			conMedidas = append(conMedidas, ruta)
		} else {
			sinMedidas = append(sinMedidas, ruta)
		}
	}

	return conMedidas, sinMedidas, nil
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

// copiarDerivadas deja en el directorio de trabajo de un guion, en su carpeta
// derivadas, una copia de las grabaciones derivadas de las de H4, con la misma
// carpeta por caso: un guion pone una en juego copiándola sobre la de su
// reproducción (contracts/arnes-e2e.md §3 de H7; research.md D23 de H7).
//
// Se copian, y no se enlazan, por lo mismo que las grabaciones: nada de lo que
// haga un guion con su copia alcanza a las del repositorio.
func copiarDerivadas(trabajo string) error {
	destino := filepath.Join(trabajo, directorioDeDerivadas)

	if err := os.CopyFS(destino, os.DirFS(derivadasDelRepositorio)); err != nil {
		return fmt.Errorf("e2e: no se pudieron copiar las derivadas de %s a %s: %w",
			derivadasDelRepositorio, destino, err)
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

// TestGuionesPorMedida fija el reparto entre TestEntregaDelHito y
// TestMedidasDeTiempo: un guion que invoca cronometra —con sangría o tras una
// condición— va con las medidas; uno que solo la nombra en una sección de
// fichero o en un comentario, no; y lo que no es un guion no va a ninguna de
// las dos listas. Así ninguna medida vuelve a correr en paralelo sin que se
// note, y ningún guion se queda sin ejecutar.
func TestGuionesPorMedida(t *testing.T) {
	t.Parallel()

	directorio := t.TempDir()
	guiones := map[string]string{
		"a-mide.txtar":         "exec kitlegal version\ncronometra 200ms kitlegal version\n",
		"b-sangria.txtar":      "  cronometra 1s kitlegal version\n",
		"c-condicion.txt":      "[unix] cronometra 1s kitlegal version\n",
		"d-no-mide.txtar":      "exec kitlegal version\n# cronometra solo en un comentario\n",
		"e-en-fichero.txtar":   "exec kitlegal version\n-- nota.txt --\ncronometra 1s algo\n",
		"f-no-es-guion.md":     "cronometra 1s kitlegal version\n",
		"g-cronometrado.txtar": "exec cronometrador\n",
	}

	for nombre, contenido := range guiones {
		require.NoError(t, os.WriteFile(filepath.Join(directorio, nombre), []byte(contenido), 0o600))
	}

	require.NoError(t, os.Mkdir(filepath.Join(directorio, "subdirectorio.txtar"), 0o700))

	conMedidas, sinMedidas, err := guionesPorMedida(directorio)
	require.NoError(t, err)

	enDirectorio := func(nombres ...string) []string {
		rutas := make([]string, 0, len(nombres))
		for _, nombre := range nombres {
			rutas = append(rutas, filepath.Join(directorio, nombre))
		}

		return rutas
	}

	assert.Equal(t, enDirectorio("a-mide.txtar", "b-sangria.txtar", "c-condicion.txt"), conMedidas)
	assert.Equal(t, enDirectorio("d-no-mide.txtar", "e-en-fichero.txtar", "g-cronometrado.txtar"), sinMedidas)

	_, _, err = guionesPorMedida(filepath.Join(directorio, "no-existe"))
	require.Error(t, err)
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
// kitlegal, dice su versión en «version», solo el de KITLEGAL_SIN_ENLACES_BIN
// deja como copia la entrada de host que los demás enlazan (FR-024), y los de
// KITLEGAL_T0_BIN, KITLEGAL_T1_BIN y KITLEGAL_T8_BIN fechan con su instante lo
// que graph stats responde sobre un directorio vacío y lo que boe pide a la
// reproducción, mientras que los demás lo fechan con el reloj del sistema
// (contracts/arnes-e2e.md §2 de H7). Es lo que comprueba que cada -X llega: el
// enlazador de Go ignora sin avisar uno que nombra una variable que no existe.
func TestBinariosDelArnes(t *testing.T) {
	t.Parallel()

	require.NoError(t, entorno.err)

	casos := []struct {
		variable string
		version  string
		enlaza   bool
		// reloj es el instante que fija su construcción; vacío, el binario
		// fecha con el reloj del sistema.
		reloj string
	}{
		{variable: variableDelBinario, version: "dev", enlaza: true},
		{variable: variableV1, version: versionV1, enlaza: true},
		{variable: variableV2, version: versionV2, enlaza: true},
		{variable: variableSinEnlaces, version: versionV1, enlaza: false},
		{variable: variableT0, version: "dev", enlaza: true, reloj: relojT0},
		{variable: variableT1, version: "dev", enlaza: true, reloj: relojT1},
		{variable: variableT8, version: "dev", enlaza: true, reloj: relojT8},
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

			compruebaElReloj(t, binario, caso.reloj)
		})
	}

	assert.Equal(t, filepath.Dir(entorno.binario), strings.SplitN(os.Getenv("PATH"), string(os.PathListSeparator), 2)[0],
		"el binario de desarrollo va en el PATH, por delante de todo")
}

// compruebaElReloj exige que el binario feche con el reloj de su construcción
// el sobre de graph stats sobre un directorio vacío, que sigue vacío, y el de
// boe articulo servido desde una copia de las grabaciones de H4, con su caché y
// su world.db en el directorio de trabajo.
func compruebaElReloj(t *testing.T, binario, reloj string) {
	t.Helper()

	vacio := t.TempDir()
	stats := exec.CommandContext(t.Context(), binario, "graph", "stats", "--json")
	compruebaLaFecha(t, stats, vacio, vacio, "kitlegal.graph", reloj)

	entradas, err := os.ReadDir(vacio)
	require.NoError(t, err)
	assert.Empty(t, entradas, "graph stats no crea nada en el directorio vacío")

	trabajo := t.TempDir()
	require.NoError(t, copiarGrabaciones(trabajo))

	articulo := exec.CommandContext(t.Context(), binario, "boe", "articulo", "BOE-A-2015-10565", "a21", "--json")
	compruebaLaFecha(t, articulo, trabajo, filepath.Join(trabajo, directorioDeLaCache), boe.NombreDeLaFuente, reloj)
}

// compruebaLaFecha ejecuta la orden desde el directorio de trabajo con la
// carpeta de la caché como único entorno y exige que salga con 0 y un sobre de
// éxito de la fuente cuya fecha_consulta sea el reloj tal cual o, sin reloj, un
// instante del reloj del sistema tomado mientras la orden se ejecutaba. El
// límite inferior se toma al segundo: una fecha que se escribiera sin fracción
// no es anterior a la invocación.
func compruebaLaFecha(t *testing.T, orden *exec.Cmd, trabajo, carpetaDeLaCache, fuente, reloj string) {
	t.Helper()

	var errores bytes.Buffer

	orden.Dir = trabajo
	orden.Env = []string{cache.VariableDirectorio + "=" + carpetaDeLaCache}
	orden.Stderr = &errores

	antes := time.Now()
	salida, err := orden.Output()
	despues := time.Now()

	require.NoError(t, err, "%q: %s", orden.Args, errores.String())

	var sobre struct {
		OK            bool   `json:"ok"`
		Fuente        string `json:"fuente"`
		FechaConsulta string `json:"fecha_consulta"`
	}

	require.NoError(t, json.Unmarshal(salida, &sobre), "%q: %s", orden.Args, salida)
	assert.True(t, sobre.OK, "%q", orden.Args)
	assert.Equal(t, fuente, sobre.Fuente, "%q", orden.Args)

	if reloj != "" {
		assert.Equal(t, reloj, sobre.FechaConsulta, "%q fecha con el reloj de su construcción", orden.Args)

		return
	}

	fecha, err := time.Parse(time.RFC3339Nano, sobre.FechaConsulta)
	require.NoError(t, err, "%q: fecha_consulta en RFC 3339", orden.Args)
	assert.Truef(t, !fecha.Before(antes.Truncate(time.Second)) && !fecha.After(despues),
		"%q fecha con el reloj del sistema: %s no está entre %s y %s", orden.Args, fecha, antes, despues)
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
// TestMain deja para todos los guiones: las rutas, absolutas —también las de
// los tres binarios con reloj de H7—; el archivo de la plataforma y la versión
// del origen; y los proxies cerrados.
func TestVariablesDeLosGuiones(t *testing.T) {
	t.Parallel()

	require.NoError(t, entorno.err)

	rutas := []string{
		variableDelBinario, variableV1, variableV2, variableSinEnlaces, variableT0, variableT1, variableT8,
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

// Lo que la orden mcp lee del guion y lo que deja en su directorio de trabajo,
// además de mcp-<n>.json y mcp-<n>.txt (contracts/arnes-e2e.md §3 de H21), y
// lo que espera.
const (
	// operacionHerramientas es la operación del fichero de llamadas que lista
	// las herramientas; toda otra primera palabra de una línea es el nombre de
	// la herramienta a la que se llama.
	operacionHerramientas = "herramientas"

	// ficheroDeErroresDeMCP lleva la salida de error del programa, entera;
	// ficheroDeSalidaDeMCP, solo con -anterior, lo que escribió en su salida
	// estándar; y ficheroDeInstruccionesDeMCP, las instructions del servidor.
	ficheroDeErroresDeMCP       = "mcp.err"
	ficheroDeSalidaDeMCP        = "mcp.out"
	ficheroDeInstruccionesDeMCP = "mcp.instrucciones"

	// variableDeTrabajo es la que testscript da a cada guion con su directorio
	// de trabajo.
	variableDeTrabajo = "WORK"

	// topeDeLaConversacion es lo más que la orden deja vivir al programa, del
	// arranque a su final. Es generoso —una conversación de un guion dura
	// décimas de segundo—: solo se agota cuando el programa ni responde ni
	// termina, que es el defecto que se busca, y entonces el guion falla aquí y
	// no con el plazo de go test, que se llevaría por delante a todos los demás.
	topeDeLaConversacion = time.Minute

	// esperaDeLasTuberias es lo más que se espera, con el programa ya
	// terminado o interrumpido, a que se cierren su salida estándar y la de
	// error, que un proceso que hubiera dejado detrás mantendría abiertas.
	esperaDeLasTuberias = 5 * time.Second
)

// Los motivos por los que mcp hace fallar un guion que la orden distingue. Son
// centinelas para que TestOrdenMCP reconozca cada rama sin leer mensajes.
var (
	errUsoDeMCP = errors.New("uso: " + ordenMCP + " [-anterior] [-dir <directorio>] <llamadas> <programa> [<argumento>…]")

	// errOperacionDeMCP es una línea del fichero de llamadas que no es ninguna
	// de las tres operaciones.
	errOperacionDeMCP = errors.New(ordenMCP + ": el fichero de llamadas lleva una línea que no es una operación")

	// errSinRespuesta es el programa que termina sin completar el saludo: el
	// único fallo de la orden que un guion puede pedir, con «!».
	errSinRespuesta = errors.New(ordenMCP + ": el programa ha terminado sin completar el saludo")

	// errSaludoCompletado es el fallo de un guion que pidió con «!» que el
	// programa terminara sin completar el saludo, y lo completó.
	errSaludoCompletado = errors.New(ordenMCP + ": lleva «!» y el programa ha completado el saludo")

	// errTopeDeMCP es el programa que sigue en marcha cuando se agota el tope
	// de la conversación: el que no responde y el que no termina tras cerrarle
	// la entrada.
	errTopeDeMCP = errors.New(ordenMCP + ": el programa seguía en marcha al agotarse el tope")
)

// claveDeLasVariables es la clave con la que el arnés guarda, en los valores de
// cada guion, los nombres de sus variables de entorno.
type claveDeLasVariables struct{}

// conversarMCP es la orden `mcp [-anterior] [-dir <directorio>] <llamadas>
// <programa> [<argumento>…]` de los guiones (contracts/arnes-e2e.md §3 de H21):
// arranca el programa con el entorno del guion y el directorio pedido, le
// conecta por su entrada y su salida estándar un cliente MCP —el del SDK del
// protocolo, que negocia la versión vigente, o con -anterior el de la
// especificación 2025-11-25—, hace en orden las operaciones del fichero de
// llamadas esperando cada respuesta, le cierra la entrada y espera a que
// termine. Escribe en la salida estándar de la orden —la que ven stdout y cmp
// stdout— una línea por cosa, y deja en $WORK el sobre de cada llamada, la
// lista de herramientas, las instructions y lo que el programa escribió.
//
// Hace falta porque un servidor MCP no se puede ejercer con `stdin fichero` y
// `exec`: con todas las peticiones de golpe y la entrada cerrada no escribe
// ninguna respuesta (research.md V12 y D12 de H21).
//
// «!» pide una sola cosa, que el programa termine sin completar el saludo: es
// lo que hace un servidor que no sirve. Cualquier otro fallo —un uso
// incorrecto, lo que cada cliente comprueba por su cuenta, un programa que no
// termina— hace fallar el guion con «!» y sin él, porque negarlo no diría qué
// se esperaba.
//
// Pide la salida estándar siempre, como arbol, y escribe en ella lo que la
// conversación haya dado también si falla: es lo que el registro del guion
// enseña junto al fallo.
func conversarMCP(ts *testscript.TestScript, neg bool, args []string) {
	salida := ts.Stdout()

	nombres, hay := ts.Value(claveDeLasVariables{}).([]string)
	if !hay {
		ts.Fatalf("%s: el arnés no ha guardado los nombres de las variables del guion", ordenMCP)
	}

	conversacion, err := conversacionDeLosArgumentos(args, ts.Getenv(variableDeTrabajo))
	ts.Check(err)

	conversacion.entorno = entornoDelGuion(nombres, ts.Getenv, conversacion.directorio)
	conversacion.tope = topeDeLaConversacion

	lineas, err := conversacion.mantener()

	_, errDeLaSalida := io.WriteString(salida, lineas)
	ts.Check(errDeLaSalida)

	ts.Check(veredictoDeMCP(neg, err))
}

// veredictoDeMCP es el fallo del guion, o nulo, a partir de cómo ha acabado la
// conversación y de si la orden lleva «!»: con él, lo único que pasa es el
// programa que termina sin completar el saludo.
func veredictoDeMCP(neg bool, err error) error {
	switch {
	case neg && errors.Is(err, errSinRespuesta):
		return nil
	case neg && err == nil:
		return errSaludoCompletado
	default:
		return err
	}
}

// nombresDeLasVariables son los nombres, ordenados y sin repetir, de las
// variables de un entorno escrito como lo recibe un proceso: «nombre=valor».
func nombresDeLasVariables(variables []string) []string {
	nombres := make([]string, 0, len(variables))

	for _, variable := range variables {
		if nombre, _, hay := strings.Cut(variable, "="); hay {
			nombres = append(nombres, nombre)
		}
	}

	slices.Sort(nombres)

	return slices.Compact(nombres)
}

// entornoDelGuion es el entorno con el que la orden mcp arranca el programa: el
// valor que cada variable tiene en el guion cuando la orden se ejecuta —el que
// le haya dado una orden env anterior, si la hubo— y PWD con el directorio de
// trabajo del programa, como hace exec.
//
// testscript no da a una orden propia el entorno del guion, solo el valor de
// la variable que se le nombre: por eso se parte de los nombres que el arnés
// guardó al preparar el guion, que son todos los que testscript y él definen.
// Una variable con un nombre nuevo, que el guion definiera con env, no llega al
// programa.
func entornoDelGuion(nombres []string, valor func(string) string, directorio string) []string {
	variables := make([]string, 0, len(nombres)+1)

	for _, nombre := range nombres {
		variables = append(variables, nombre+"="+valor(nombre))
	}

	return append(variables, "PWD="+directorio)
}

// conversacionMCP es una ejecución de la orden mcp ya interpretada.
type conversacionMCP struct {
	// anterior elige el cliente de la especificación 2025-11-25.
	anterior bool

	// trabajo es $WORK, de donde se lee el fichero de llamadas y donde se
	// dejan los demás; directorio, el de trabajo del programa, absoluto.
	trabajo, directorio string

	// llamadas es el fichero de las operaciones, relativo a $WORK.
	llamadas string

	// programa es la ruta absoluta del que se arranca; argumentos, los suyos;
	// y entorno, el del guion.
	programa   string
	argumentos []string
	entorno    []string

	// tope es lo más que el programa vive.
	tope time.Duration
}

// conversacionDeLosArgumentos interpreta los argumentos de mcp: las banderas
// de la orden, que van delante del fichero de llamadas, el fichero, el programa
// y sus argumentos, que no se miran aunque empiecen por un guion. El directorio
// de -dir es relativo a $WORK si no es absoluto, y $WORK si no se da. El
// programa se da por su ruta absoluta: la orden no busca nada en el PATH, ni en
// el del guion ni en el del proceso de los tests.
func conversacionDeLosArgumentos(args []string, trabajo string) (*conversacionMCP, error) {
	conversacion := &conversacionMCP{trabajo: trabajo, directorio: trabajo}

	resto := args

	for len(resto) > 0 && strings.HasPrefix(resto[0], "-") {
		switch bandera := resto[0]; {
		case bandera == "-anterior":
			conversacion.anterior = true
			resto = resto[1:]
		case bandera == "-dir" && len(resto) > 1:
			conversacion.directorio = resto[1]
			if !filepath.IsAbs(conversacion.directorio) {
				conversacion.directorio = filepath.Join(trabajo, conversacion.directorio)
			}

			resto = resto[2:]
		default:
			return nil, fmt.Errorf("%w: %q no es una bandera de la orden, o le falta su valor, en %q",
				errUsoDeMCP, bandera, args)
		}
	}

	if len(resto) < 2 {
		return nil, fmt.Errorf("%w: faltan el fichero de llamadas o el programa en %q", errUsoDeMCP, args)
	}

	conversacion.llamadas, conversacion.programa, conversacion.argumentos = resto[0], resto[1], resto[2:]

	if !filepath.IsAbs(conversacion.programa) {
		return nil, fmt.Errorf("%w: el programa %q no es una ruta absoluta", errUsoDeMCP, conversacion.programa)
	}

	return conversacion, nil
}

// operacionMCP es una línea del fichero de llamadas, con su número.
type operacionMCP struct {
	numero int

	// lista dice que la operación es la que lista las herramientas; si no, es
	// una llamada a herramienta.
	lista       bool
	herramienta string

	// argumentos son los de la llamada, un objeto JSON; nulos, la llamada va
	// sin ellos.
	argumentos json.RawMessage
}

// operacionesDe son las operaciones del contenido de un fichero de llamadas,
// una por línea y numeradas desde 1; las líneas vacías no cuentan. Una línea es
// `herramientas`, `<herramienta> <objeto JSON>` o `<herramienta>` sola: lo que
// siga al nombre tiene que ser un objeto JSON, y a `herramientas` no la sigue
// nada.
func operacionesDe(llamadas string) ([]operacionMCP, error) {
	var operaciones []operacionMCP

	for linea := range strings.SplitSeq(llamadas, "\n") {
		linea = strings.TrimSpace(linea)
		if linea == "" {
			continue
		}

		nombre, argumentos, _ := strings.Cut(linea, " ")
		argumentos = strings.TrimSpace(argumentos)

		operacion := operacionMCP{numero: len(operaciones) + 1}

		switch {
		case nombre == operacionHerramientas && argumentos == "":
			operacion.lista = true
		case nombre == operacionHerramientas:
			return nil, fmt.Errorf("%w: %q: %s no lleva argumentos", errOperacionDeMCP, linea, operacionHerramientas)
		case argumentos == "":
			operacion.herramienta = nombre
		case strings.HasPrefix(argumentos, "{") && json.Valid([]byte(argumentos)):
			operacion.herramienta, operacion.argumentos = nombre, json.RawMessage(argumentos)
		default:
			return nil, fmt.Errorf("%w: %q: lo que sigue a la herramienta no es un objeto JSON", errOperacionDeMCP, linea)
		}

		operaciones = append(operaciones, operacion)
	}

	return operaciones, nil
}

// mantener hace la conversación entera y devuelve las líneas que la orden
// escribe en su salida estándar, también las que haya dado una conversación
// que falla. El error es errSinRespuesta si el programa termina sin completar
// el saludo, errTopeDeMCP si sigue en marcha al agotarse el tope, y cualquier
// otro si la conversación no sale: lo que cada cliente comprueba por su
// cuenta, entre ellos.
//
// Pase lo que pase, el programa no sobrevive a la orden, y lo que escribió
// queda en $WORK: mcp.err y, con -anterior, mcp.out. Si la conversación falla,
// el error lleva además la salida de error del programa, que es donde un
// servidor cuenta lo que le pasa.
func (c *conversacionMCP) mantener() (lineas string, err error) {
	raiz, err := os.OpenRoot(c.trabajo)
	if err != nil {
		return "", fmt.Errorf("%s: el directorio de trabajo del guion: %w", ordenMCP, err)
	}

	defer func() { err = errors.Join(err, raiz.Close()) }()

	llamadas, err := raiz.ReadFile(c.llamadas)
	if err != nil {
		return "", fmt.Errorf("%s: el fichero de llamadas: %w", ordenMCP, err)
	}

	operaciones, err := operacionesDe(string(llamadas))
	if err != nil {
		return "", err
	}

	ctx, cancelar := context.WithTimeout(context.Background(), c.tope)
	defer cancelar()

	programa, err := c.arrancar(ctx)
	if err != nil {
		return "", err
	}

	var escritas strings.Builder

	err = c.hablar(ctx, raiz, programa, operaciones, &escritas)

	// Si la conversación ha salido, el programa ya ha terminado y esto no hace
	// nada; si no, lo interrumpe.
	cancelar()

	err = errors.Join(err, programa.retirar(), c.dejarLoEscrito(raiz, programa))
	if err != nil && programa.errores.Len() > 0 {
		err = fmt.Errorf("%w\n[%s]\n%s", err, ficheroDeErroresDeMCP, programa.errores.String())
	}

	return escritas.String(), err
}

// hablar hace el saludo, las operaciones y el cierre con el programa ya
// arrancado, y escribe en lineas lo que la orden dice de cada cosa: el
// protocolo negociado, las capacidades, una línea por operación y el código
// con el que el programa termina.
func (c *conversacionMCP) hablar(
	ctx context.Context,
	raiz *os.Root,
	programa *programaEnMarcha,
	operaciones []operacionMCP,
	lineas *strings.Builder,
) error {
	sesion, err := mcptest.Abrir(ctx, programa.salida, programa.entrada, c.anterior)
	if err != nil {
		return c.sinSesion(ctx, programa, lineas, err)
	}

	lineas.WriteString("protocolo " + sesion.Protocolo() + "\n")
	lineas.WriteString(strings.Join(slices.Concat([]string{"capacidades"}, sesion.Capacidades()), " ") + "\n")

	if err := raiz.WriteFile(ficheroDeInstruccionesDeMCP, []byte(sesion.Instrucciones()+"\n"), 0o600); err != nil {
		return c.fallo(ctx, "las instructions del servidor", err)
	}

	for _, operacion := range operaciones {
		linea, err := operacion.hacer(ctx, raiz, sesion)
		if err != nil {
			return c.fallo(ctx, fmt.Sprintf("la operación %d", operacion.numero), err)
		}

		lineas.WriteString(linea)
	}

	if err := sesion.Cerrar(ctx); err != nil {
		return c.fallo(ctx, "el cierre de la entrada del programa", err)
	}

	codigo, err := programa.esperar(ctx)
	if err != nil {
		return c.fallo(ctx, "la espera del programa tras cerrarle la entrada", err)
	}

	fmt.Fprintf(lineas, "salida %d\n", codigo)

	return nil
}

// sinSesion es lo que queda por hacer cuando el saludo no sale. Si el programa
// ha cerrado la conexión antes de completarlo, la orden espera a que termine,
// sin cerrarle la entrada: es lo que distingue al que termina por su cuenta del
// que termina porque se la cierran (FR-022 de H21). Entonces dice `sin
// respuesta` y el código, y falla con errSinRespuesta. Cualquier otro saludo
// fallido es un fallo del guion sin más.
func (c *conversacionMCP) sinSesion(
	ctx context.Context,
	programa *programaEnMarcha,
	lineas *strings.Builder,
	errDelSaludo error,
) error {
	if !errors.Is(errDelSaludo, mcptest.ErrSinSaludo) {
		return c.fallo(ctx, "el saludo", errDelSaludo)
	}

	codigo, err := programa.esperar(ctx)
	if err != nil {
		return c.fallo(ctx, "la espera del programa, que ha cerrado la conexión sin completar el saludo",
			errors.Join(errDelSaludo, err))
	}

	fmt.Fprintf(lineas, "sin respuesta\nsalida %d\n", codigo)

	return fmt.Errorf("%w: %w", errSinRespuesta, errDelSaludo)
}

// fallo es el error de un paso de la conversación que no sale. Con el tope
// agotado lo que se cuenta es el tope: lo que el paso diga entonces es solo
// cómo se ha enterado de que al programa lo han interrumpido.
func (c *conversacionMCP) fallo(ctx context.Context, paso string, err error) error {
	if ctx.Err() != nil {
		return fmt.Errorf("%w, de %s, en %s: %w", errTopeDeMCP, c.tope, paso, err)
	}

	return fmt.Errorf("%s: %s: %w", ordenMCP, paso, err)
}

// dejarLoEscrito deja en $WORK, con el programa ya terminado, su salida de
// error entera y, con -anterior, lo que escribió en su salida estándar, tal
// cual: todas sus líneas, las haya leído el cliente o no.
func (c *conversacionMCP) dejarLoEscrito(raiz *os.Root, programa *programaEnMarcha) error {
	err := raiz.WriteFile(ficheroDeErroresDeMCP, programa.errores.Bytes(), 0o600)

	if c.anterior {
		err = errors.Join(err, raiz.WriteFile(ficheroDeSalidaDeMCP, programa.salidaEntera(), 0o600))
	}

	if err != nil {
		return fmt.Errorf("%s: lo que el programa escribió: %w", ordenMCP, err)
	}

	return nil
}

// hacer hace una operación con la sesión abierta, deja su fichero en $WORK y
// devuelve la línea que la orden escribe de ella.
func (o operacionMCP) hacer(ctx context.Context, raiz *os.Root, sesion *mcptest.Sesion) (string, error) {
	if o.lista {
		return o.listar(ctx, raiz, sesion)
	}

	return o.llamar(ctx, raiz, sesion)
}

// listar pide las herramientas y deja en mcp-<n>.txt una línea por cada una, en
// el orden del servidor, con su nombre y si se anuncia de solo lectura.
func (o operacionMCP) listar(ctx context.Context, raiz *os.Root, sesion *mcptest.Sesion) (string, error) {
	herramientas, err := sesion.Herramientas(ctx)
	if err != nil {
		return "", err
	}

	var lista strings.Builder

	for _, herramienta := range herramientas {
		fmt.Fprintf(&lista, "%s lectura=%t\n", herramienta.Nombre, herramienta.SoloLectura)
	}

	if err := raiz.WriteFile(fmt.Sprintf("mcp-%d.txt", o.numero), []byte(lista.String()), 0o600); err != nil {
		return "", err
	}

	return fmt.Sprintf("%d %s %d\n", o.numero, operacionHerramientas, len(herramientas)), nil
}

// llamar llama a la herramienta. Un error del protocolo es una línea y ningún
// fichero. Un resultado, que la sesión ya ha comprobado, deja en mcp-<n>.json
// el texto de su bloque de texto y un salto de línea, que es lo que la orden
// del mismo verbo escribe con --json; si viene marcado como error de
// herramienta, la línea dice la clase de su sobre de fallo.
func (o operacionMCP) llamar(ctx context.Context, raiz *os.Root, sesion *mcptest.Sesion) (string, error) {
	resultado, err := sesion.Llamar(ctx, o.herramienta, o.argumentos)

	var deProtocolo *mcptest.ErrorDeProtocolo
	if errors.As(err, &deProtocolo) {
		return fmt.Sprintf("%d %s protocolo %s\n", o.numero, o.herramienta, deProtocolo.Mensaje), nil
	}

	if err != nil {
		return "", err
	}

	sobre := slices.Concat(resultado.Sobre, []byte("\n"))
	if err := raiz.WriteFile(fmt.Sprintf("mcp-%d.json", o.numero), sobre, 0o600); err != nil {
		return "", err
	}

	if !resultado.Fallo {
		return fmt.Sprintf("%d %s resultado\n", o.numero, o.herramienta), nil
	}

	clase, err := claseDelFallo(resultado.Sobre)
	if err != nil {
		return "", err
	}

	return fmt.Sprintf("%d %s error %s\n", o.numero, o.herramienta, clase), nil
}

// claseDelFallo es la clase de un sobre de fallo, su data.clase.
func claseDelFallo(sobre []byte) (string, error) {
	var fallo struct {
		Data struct {
			Clase string `json:"clase"`
		} `json:"data"`
	}

	if err := json.Unmarshal(sobre, &fallo); err != nil {
		return "", fmt.Errorf("el sobre de fallo no se puede leer: %w", err)
	}

	if fallo.Data.Clase == "" {
		return "", fmt.Errorf("el sobre de fallo no lleva data.clase: %s", sobre)
	}

	return fallo.Data.Clase, nil
}

// programaEnMarcha es el programa de una conversación, ya arrancado.
type programaEnMarcha struct {
	// entrada es su entrada estándar, que cierra la sesión del cliente al
	// terminar la conversación y nadie más antes de que el programa termine.
	entrada io.WriteCloser

	// salida es por donde el cliente lee su salida estándar, que llega a su
	// final cuando el programa termina.
	salida *io.PipeReader

	// reparto guarda todo lo que escribe en su salida estándar y errores, en
	// la de error. Están completos cuando terminado se cierra.
	reparto *salidaRepartida
	errores *bytes.Buffer

	// terminado se cierra cuando el programa ha terminado y sus dos salidas se
	// han leído enteras; final es lo que dio entonces la espera, y solo se lee
	// después.
	terminado <-chan struct{}
	final     error
}

// salidaRepartida es a donde va la salida estándar del programa: guarda todo
// lo que escribe y se lo pasa al cliente mientras el cliente lo lee. Con un
// cliente que ya no lee —ha terminado, o su saludo ha fallado—, lo que el
// programa escriba después solo se guarda: ni el programa ni quien copia su
// salida se quedan esperando a un lector que no va a volver.
type salidaRepartida struct {
	guardada bytes.Buffer
	cliente  *io.PipeWriter
}

// Write guarda lo escrito y se lo pasa al cliente si sigue leyendo.
func (s *salidaRepartida) Write(escrito []byte) (int, error) {
	s.guardada.Write(escrito)

	if _, err := s.cliente.Write(escrito); err != nil && !errors.Is(err, io.ErrClosedPipe) {
		return 0, err
	}

	return len(escrito), nil
}

// arrancar arranca el programa con sus argumentos, el entorno del guion y su
// directorio de trabajo, con una tubería como entrada y su salida estándar
// repartida entre el cliente y lo que la orden guarda. Cuando ctx termina, el
// programa se interrumpe.
func (c *conversacionMCP) arrancar(ctx context.Context) (*programaEnMarcha, error) {
	desdeElPrograma, haciaElCliente := io.Pipe()

	programa := &programaEnMarcha{
		salida:  desdeElPrograma,
		reparto: &salidaRepartida{cliente: haciaElCliente},
		errores: &bytes.Buffer{},
	}

	// El programa y sus argumentos son los que escribe un guion, que es un
	// fichero del repositorio o un test de este paquete: no hay entrada de
	// usuario ni red en la orden, de modo que G204 no tiene aquí nada que
	// prevenir.
	//nolint:gosec // el programa y sus argumentos los escribe un guion del repositorio; no hay entrada externa.
	orden := exec.CommandContext(ctx, c.programa, c.argumentos...)
	orden.Dir = c.directorio
	orden.Env = c.entorno
	orden.Stdout = programa.reparto
	orden.Stderr = programa.errores
	orden.WaitDelay = esperaDeLasTuberias

	entrada, err := orden.StdinPipe()
	if err != nil {
		return nil, fmt.Errorf("%s: la entrada de %s: %w", ordenMCP, c.programa, err)
	}

	if err := orden.Start(); err != nil {
		return nil, fmt.Errorf("%s: no se pudo arrancar %s: %w", ordenMCP, c.programa, err)
	}

	terminado := make(chan struct{})
	programa.entrada, programa.terminado = entrada, terminado

	go func() {
		defer close(terminado)

		// La espera vuelve con el programa terminado y sus dos salidas leídas
		// hasta el final; el final de su salida estándar es entonces el de lo
		// que lee el cliente.
		programa.final = errors.Join(orden.Wait(), haciaElCliente.Close())
	}()

	return programa, nil
}

// retirar deja de leer la salida del programa y espera a que termine, que es lo
// que tarda en estar completo lo que escribió.
func (p *programaEnMarcha) retirar() error {
	err := p.salida.Close()

	<-p.terminado

	return err
}

// esperar espera, con el cliente ya sin leer, a que el programa termine por su
// cuenta, y da su código de salida. Si lo que lo ha terminado es el final de
// ctx, no ha terminado por su cuenta.
func (p *programaEnMarcha) esperar(ctx context.Context) (int, error) {
	if err := p.retirar(); err != nil {
		return 0, err
	}

	if err := ctx.Err(); err != nil {
		return 0, fmt.Errorf("el programa no ha terminado por su cuenta: %w", err)
	}

	var deSalida *exec.ExitError

	switch {
	case p.final == nil:
		return 0, nil
	case errors.As(p.final, &deSalida):
		return deSalida.ExitCode(), nil
	default:
		return 0, fmt.Errorf("la espera del programa: %w", p.final)
	}
}

// salidaEntera es todo lo que el programa escribió en su salida estándar. Solo
// se lee con el programa terminado.
func (p *programaEnMarcha) salidaEntera() []byte { return p.reparto.guardada.Bytes() }

// saludoDelSustituto es el principio de un guion de sh que hace de servidor MCP
// ante el cliente anterior: lee `initialize`, responde a la petición 1 —la
// primera de ese cliente— con una capacidad y unas instructions, y lee la
// notificación `initialized`. Lo que haga después lo dice cada prueba.
const saludoDelSustituto = `read -r saludo || exit 9
printf '%s\n' '` + respuestaDelSustituto + `'
read -r iniciada || exit 9
`

// respuestaDelSustituto es la línea con la que el sustituto responde al saludo.
const respuestaDelSustituto = `{"jsonrpc":"2.0","id":1,"result":{"protocolVersion":"2025-11-25",` +
	`"capabilities":{"tools":{}},"serverInfo":{"name":"sustituto","version":"0"},` +
	`"instructions":"Soy un sustituto."}}`

// TestOrdenMCP fija de la orden mcp lo que los guiones no pueden
// (contracts/arnes-e2e.md §3 de H21), sin lanzar ninguno, como TestCronometra y
// TestArbol: el uso, las operaciones del fichero de llamadas, el entorno que
// recibe el programa, qué deja pasar «!», y con sustitutos del servidor
// escritos en sh, el programa que no termina tras cerrarle la entrada, el que
// ni saluda ni termina y el que termina sin saludar. La forma de las líneas y
// de los ficheros de una conversación con el servidor de verdad la fijan los
// guiones, que la comparan con cmp, y lo que cada cliente comprueba por su
// cuenta, TestSesion (§5).
func TestOrdenMCP(t *testing.T) {
	t.Parallel()

	t.Run("el uso", testUsoDeMCP)
	t.Run("las operaciones del fichero de llamadas", testOperacionesDeMCP)
	t.Run("el entorno del guion", testEntornoDelGuion)
	t.Run("lo que deja pasar «!»", testVeredictoDeMCP)
	t.Run("un programa que saluda y termina al cerrarle la entrada", testConversacionConUnSustituto)
	t.Run("un programa que no termina tras cerrarle la entrada", testProgramaQueNoTermina)
	t.Run("un programa que ni saluda ni termina", testProgramaQueNoSaluda)
	t.Run("un programa que termina sin saludar", testProgramaSinRespuesta)
	t.Run("lo que impide arrancar la conversación", testConversacionQueNoArranca)
}

func testUsoDeMCP(t *testing.T) {
	t.Parallel()

	const trabajo = "/trabajo/del guion"

	validos := []struct {
		nombre string
		args   []string
		quiere conversacionMCP
	}{
		{
			nombre: "sin banderas, el directorio es el de trabajo del guion",
			args:   []string{"llamadas.txt", "/bin/kitlegal", "mcp", "serve"},
			quiere: conversacionMCP{
				trabajo: trabajo, directorio: trabajo, llamadas: "llamadas.txt",
				programa: "/bin/kitlegal", argumentos: []string{"mcp", "serve"},
			},
		},
		{
			nombre: "las dos banderas, con un directorio absoluto",
			args:   []string{"-anterior", "-dir", "/", "llamadas.txt", "/con espacios/kitlegal", "mcp", "serve", "--verbose"},
			quiere: conversacionMCP{
				anterior: true, trabajo: trabajo, directorio: "/", llamadas: "llamadas.txt",
				programa: "/con espacios/kitlegal", argumentos: []string{"mcp", "serve", "--verbose"},
			},
		},
		{
			nombre: "un directorio relativo lo es al de trabajo del guion",
			args:   []string{"-dir", "otro sitio", "-anterior", "llamadas.txt", "/bin/kitlegal"},
			quiere: conversacionMCP{
				anterior: true, trabajo: trabajo, directorio: filepath.Join(trabajo, "otro sitio"), llamadas: "llamadas.txt",
				programa: "/bin/kitlegal", argumentos: []string{},
			},
		},
		{
			nombre: "lo que sigue al programa es suyo, aunque se llame como una bandera de la orden",
			args:   []string{"llamadas.txt", "/bin/kitlegal", "-anterior", "-dir", "/"},
			quiere: conversacionMCP{
				trabajo: trabajo, directorio: trabajo, llamadas: "llamadas.txt",
				programa: "/bin/kitlegal", argumentos: []string{"-anterior", "-dir", "/"},
			},
		},
	}

	for _, caso := range validos {
		t.Run(caso.nombre, func(t *testing.T) {
			t.Parallel()

			conversacion, err := conversacionDeLosArgumentos(caso.args, trabajo)
			require.NoError(t, err)
			assert.Equal(t, &caso.quiere, conversacion)
		})
	}

	rechazados := map[string][]string{
		"sin nada":                        nil,
		"sin programa":                    {"llamadas.txt"},
		"solo banderas":                   {"-anterior"},
		"-dir sin su valor":               {"-dir"},
		"-dir sin fichero ni programa":    {"-dir", "/"},
		"una bandera que no es":           {"-vigente", "llamadas.txt", "/bin/kitlegal"},
		"una bandera con dos guiones":     {"--anterior", "llamadas.txt", "/bin/kitlegal"},
		"un programa nombrado sin ruta":   {"llamadas.txt", "kitlegal", "mcp", "serve"},
		"un programa de ruta relativa":    {"llamadas.txt", "bin/kitlegal", "mcp", "serve"},
		"las banderas detrás del fichero": {"llamadas.txt", "-anterior", "/bin/kitlegal"},
	}

	for nombre, args := range rechazados {
		t.Run(nombre, func(t *testing.T) {
			t.Parallel()

			_, err := conversacionDeLosArgumentos(args, trabajo)
			require.ErrorIs(t, err, errUsoDeMCP)
		})
	}
}

func testOperacionesDeMCP(t *testing.T) {
	t.Parallel()

	t.Run("las tres formas, numeradas desde 1 sin contar las líneas vacías", func(t *testing.T) {
		t.Parallel()

		operaciones, err := operacionesDe("herramientas\n\n" +
			"boe_articulo {\"norma\":\"BOE-A-2015-10565\",\"bloque\":\"a21\"}\n" +
			"  \t\n" +
			"version\n" +
			"graph_stats   {}  \n" +
			"herramientas")
		require.NoError(t, err)
		assert.Equal(t, []operacionMCP{
			{numero: 1, lista: true},
			{
				numero: 2, herramienta: "boe_articulo",
				argumentos: json.RawMessage(`{"norma":"BOE-A-2015-10565","bloque":"a21"}`),
			},
			{numero: 3, herramienta: "version"},
			{numero: 4, herramienta: "graph_stats", argumentos: json.RawMessage(`{}`)},
			{numero: 5, lista: true},
		}, operaciones)
	})

	t.Run("un fichero sin operaciones", func(t *testing.T) {
		t.Parallel()

		for _, llamadas := range []string{"", "\n", "\n  \n\n"} {
			operaciones, err := operacionesDe(llamadas)
			require.NoError(t, err)
			assert.Empty(t, operaciones, "%q", llamadas)
		}
	})

	rechazadas := map[string]string{
		"herramientas con argumentos":      "herramientas {}\n",
		"una lista en lugar de un objeto":  "boe_articulos [\"a21\"]\n",
		"una cadena en lugar de un objeto": "boe_articulo \"a21\"\n",
		"un objeto sin cerrar":             "boe_articulo {\"norma\":\n",
		"algo detrás del objeto":           "boe_articulo {} {}\n",
		"detrás de operaciones que valen":  "herramientas\nversion\nboe_articulo a21\n",
	}

	for nombre, llamadas := range rechazadas {
		t.Run(nombre, func(t *testing.T) {
			t.Parallel()

			operaciones, err := operacionesDe(llamadas)
			require.ErrorIs(t, err, errOperacionDeMCP)
			assert.Empty(t, operaciones, "un fichero con una línea que no vale no da ninguna operación")
		})
	}
}

func testEntornoDelGuion(t *testing.T) {
	t.Parallel()

	// Lo que testscript y el arnés dejan en un guion, con sus nombres raros, uno
	// repetido y una entrada que no es una variable.
	nombres := nombresDeLasVariables([]string{
		"WORK=/trabajo", "PATH=/bin:/usr/bin", "HOME=/no-home", "$=$", "/=/", "KITLEGAL_CACHE_DIR=/trabajo/cache",
		"HOME=/otra", "NO_PROXY=", "sin-igual",
	})
	assert.Equal(t, []string{"$", "/", "HOME", "KITLEGAL_CACHE_DIR", "NO_PROXY", "PATH", "WORK"}, nombres)

	// El valor de cada una es el que tiene el guion cuando la orden se ejecuta:
	// HOME, el que le dio después una orden env.
	valores := map[string]string{
		"WORK": "/trabajo", "PATH": "/bin:/usr/bin", "HOME": "/trabajo/home", "$": "$", "/": "/",
		"KITLEGAL_CACHE_DIR": "/trabajo/cache", "NO_PROXY": "",
	}

	assert.Equal(t, []string{
		"$=$", "/=/", "HOME=/trabajo/home", "KITLEGAL_CACHE_DIR=/trabajo/cache", "NO_PROXY=", "PATH=/bin:/usr/bin",
		"WORK=/trabajo", "PWD=/con espacios",
	}, entornoDelGuion(nombres, func(nombre string) string { return valores[nombre] }, "/con espacios"))
}

func testVeredictoDeMCP(t *testing.T) {
	t.Parallel()

	sinRespuesta := fmt.Errorf("%w: %w", errSinRespuesta, mcptest.ErrSinSaludo)
	noCumple := errors.New("mcp: la operación 2: el resultado de boe_articulo no cumple")
	tope := fmt.Errorf("%w, de 1m0s, en el saludo: %w", errTopeDeMCP, context.DeadlineExceeded)

	casos := []struct {
		nombre string
		neg    bool
		err    error
		// quiere es el error que el fallo del guion deja alcanzable; nulo, el
		// guion sigue.
		quiere error
	}{
		{nombre: "sin «!», la conversación que sale", err: nil, quiere: nil},
		{nombre: "sin «!», el programa que termina sin saludar", err: sinRespuesta, quiere: errSinRespuesta},
		{nombre: "sin «!», lo que un cliente comprueba", err: noCumple, quiere: noCumple},
		{nombre: "sin «!», el tope", err: tope, quiere: errTopeDeMCP},
		{nombre: "con «!», el programa que termina sin saludar", neg: true, err: sinRespuesta, quiere: nil},
		{nombre: "con «!», la conversación que sale", neg: true, err: nil, quiere: errSaludoCompletado},
		{nombre: "con «!», lo que un cliente comprueba", neg: true, err: noCumple, quiere: noCumple},
		{nombre: "con «!», el tope", neg: true, err: tope, quiere: errTopeDeMCP},
	}

	for _, caso := range casos {
		t.Run(caso.nombre, func(t *testing.T) {
			t.Parallel()

			err := veredictoDeMCP(caso.neg, caso.err)
			if caso.quiere == nil {
				require.NoError(t, err)

				return
			}

			require.ErrorIs(t, err, caso.quiere)
		})
	}
}

// conversacionCon es una conversación de la orden mcp con el programa nombrado,
// que se busca en el PATH del proceso, en un directorio de trabajo nuevo con un
// fichero de llamadas sin operaciones y otro directorio, distinto, para el
// programa. Su entorno es el PATH y una variable del guion.
func conversacionCon(t *testing.T, anterior bool, tope time.Duration, nombre string, argumentos ...string) *conversacionMCP {
	t.Helper()

	programa, err := exec.LookPath(nombre)
	require.NoError(t, err)
	require.True(t, filepath.IsAbs(programa), "%s tiene que estar en una entrada absoluta del PATH: %q", nombre, programa)

	trabajo, directorio := t.TempDir(), t.TempDir()
	require.NoError(t, os.WriteFile(filepath.Join(trabajo, "llamadas.txt"), []byte("\n"), 0o600))

	valores := map[string]string{"PATH": os.Getenv("PATH"), "DEL_GUION": "lo que dice el guion"}

	return &conversacionMCP{
		anterior:   anterior,
		trabajo:    trabajo,
		directorio: directorio,
		llamadas:   "llamadas.txt",
		programa:   programa,
		argumentos: argumentos,
		entorno:    entornoDelGuion([]string{"DEL_GUION", "PATH"}, func(nombre string) string { return valores[nombre] }, directorio),
		tope:       tope,
	}
}

// dejadoPorMCP es el contenido del fichero que la orden deja en el directorio
// de trabajo de la conversación.
func dejadoPorMCP(t *testing.T, conversacion *conversacionMCP, fichero string) string {
	t.Helper()

	contenido, err := fs.ReadFile(os.DirFS(conversacion.trabajo), fichero)
	require.NoError(t, err)

	return string(contenido)
}

// lineasDejadasPorMCP son las líneas de ese fichero, cada una con su salto: es
// como se compara mcp.out, que lleva cada línea que el programa escribió en su
// salida estándar, carácter a carácter.
func lineasDejadasPorMCP(t *testing.T, conversacion *conversacionMCP, fichero string) []string {
	t.Helper()

	return slices.Collect(strings.Lines(dejadoPorMCP(t, conversacion, fichero)))
}

func testConversacionConUnSustituto(t *testing.T) {
	t.Parallel()

	// Tras el saludo dice en su salida de error dónde está y lo que vale la
	// variable del guion, lee su entrada hasta que se la cierran y termina con
	// un código que no es 0: la orden lo cuenta, y no falla por él.
	conversacion := conversacionCon(t, true, topeDeLaConversacion, "sh", "-c", saludoDelSustituto+
		"pwd >&2\nprintf '%s\\n' \"$DEL_GUION\" >&2\nwhile read -r linea; do :; done\nexit 4\n")

	lineas, err := conversacion.mantener()
	require.NoError(t, err)
	assert.Equal(t, "protocolo 2025-11-25\ncapacidades tools\nsalida 4\n", lineas)

	assert.Equal(t, "Soy un sustituto.\n", dejadoPorMCP(t, conversacion, ficheroDeInstruccionesDeMCP))
	assert.Equal(t, []string{respuestaDelSustituto + "\n"}, lineasDejadasPorMCP(t, conversacion, ficheroDeSalidaDeMCP))

	donde, variable, hay := strings.Cut(dejadoPorMCP(t, conversacion, ficheroDeErroresDeMCP), "\n")
	require.True(t, hay)
	assert.Equal(t, "lo que dice el guion\n", variable, "el programa recibe el entorno del guion")

	// El directorio temporal puede nombrarse a través de un enlace: se compara
	// lo que nombran las dos rutas.
	quiere, err := filepath.EvalSymlinks(conversacion.directorio)
	require.NoError(t, err)

	tiene, err := filepath.EvalSymlinks(donde)
	require.NoError(t, err)
	assert.Equal(t, quiere, tiene, "el programa arranca en su directorio, que no es el de trabajo del guion")
}

func testProgramaQueNoTermina(t *testing.T) {
	t.Parallel()

	// El tope da tiempo de sobra al saludo, también con la máquina cargada: lo
	// que lo agota es el programa, que tras él se queda dormido sin mirar su
	// entrada.
	const tope = 5 * time.Second

	conversacion := conversacionCon(t, true, tope, "sh", "-c", saludoDelSustituto+"exec sleep 60\n")

	inicio := time.Now()
	lineas, err := conversacion.mantener()
	duracion := time.Since(inicio)

	require.ErrorIs(t, err, errTopeDeMCP)
	require.NotErrorIs(t, err, errSinRespuesta)
	assert.Equal(t, "protocolo 2025-11-25\ncapacidades tools\n", lineas,
		"el saludo salió, y de un programa interrumpido no se dice ningún código")
	assert.GreaterOrEqual(t, duracion, tope, "la orden espera el tope entero antes de darlo por colgado")
	assert.Equal(t, []string{respuestaDelSustituto + "\n"}, lineasDejadasPorMCP(t, conversacion, ficheroDeSalidaDeMCP))
}

func testProgramaQueNoSaluda(t *testing.T) {
	t.Parallel()

	for _, anterior := range []bool{false, true} {
		t.Run(fmt.Sprintf("anterior=%t", anterior), func(t *testing.T) {
			t.Parallel()

			conversacion := conversacionCon(t, anterior, 500*time.Millisecond, "sleep", "60")

			lineas, err := conversacion.mantener()
			require.ErrorIs(t, err, errTopeDeMCP)
			require.NotErrorIs(t, err, errSinRespuesta,
				"un programa al que hay que interrumpir no ha terminado sin completar el saludo")
			assert.Empty(t, lineas)
		})
	}
}

func testProgramaSinRespuesta(t *testing.T) {
	t.Parallel()

	for _, anterior := range []bool{false, true} {
		t.Run(fmt.Sprintf("anterior=%t", anterior), func(t *testing.T) {
			t.Parallel()

			conversacion := conversacionCon(t, anterior, topeDeLaConversacion, "sh", "-c", "echo 'no sirvo' >&2\nexit 3\n")

			lineas, err := conversacion.mantener()
			require.ErrorIs(t, err, errSinRespuesta)
			assert.Equal(t, "sin respuesta\nsalida 3\n", lineas)
			assert.Equal(t, "no sirvo\n", dejadoPorMCP(t, conversacion, ficheroDeErroresDeMCP))
			assert.NoFileExists(t, filepath.Join(conversacion.trabajo, ficheroDeInstruccionesDeMCP))

			if anterior {
				assert.Empty(t, dejadoPorMCP(t, conversacion, ficheroDeSalidaDeMCP))
			} else {
				assert.NoFileExists(t, filepath.Join(conversacion.trabajo, ficheroDeSalidaDeMCP))
			}
		})
	}
}

func testConversacionQueNoArranca(t *testing.T) {
	t.Parallel()

	casos := []struct {
		nombre string
		// estropear deja la conversación, que valía, como el caso la necesita.
		estropear func(t *testing.T, conversacion *conversacionMCP)
		quiere    error
	}{
		{
			nombre: "un fichero de llamadas que no existe",
			estropear: func(t *testing.T, conversacion *conversacionMCP) {
				t.Helper()

				conversacion.llamadas = "no-existe.txt"
			},
			quiere: fs.ErrNotExist,
		},
		{
			nombre: "un fichero de llamadas con una línea que no es una operación",
			estropear: func(t *testing.T, conversacion *conversacionMCP) {
				t.Helper()

				require.NoError(t, os.WriteFile(filepath.Join(conversacion.trabajo, conversacion.llamadas),
					[]byte("herramientas {}\n"), 0o600))
			},
			quiere: errOperacionDeMCP,
		},
		{
			nombre: "un programa que no existe",
			estropear: func(t *testing.T, conversacion *conversacionMCP) {
				t.Helper()

				conversacion.programa = filepath.Join(conversacion.trabajo, "no-existe")
			},
			quiere: fs.ErrNotExist,
		},
	}

	for _, caso := range casos {
		t.Run(caso.nombre, func(t *testing.T) {
			t.Parallel()

			conversacion := conversacionCon(t, true, topeDeLaConversacion, "sh", "-c", "exit 0")
			caso.estropear(t, conversacion)

			lineas, err := conversacion.mantener()
			require.ErrorIs(t, err, caso.quiere)
			require.NotErrorIs(t, err, errSinRespuesta)
			assert.Empty(t, lineas)
			assert.NoFileExists(t, filepath.Join(conversacion.trabajo, ficheroDeErroresDeMCP),
				"sin programa arrancado, la orden no deja nada suyo")
		})
	}
}
