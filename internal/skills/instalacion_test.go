//go:build integration

// Las pruebas de este fichero ejecutan `make install` de verdad —la receta del
// Makefile: el go install y, con el binario recién instalado,
// `kitlegal skills install -g --host claude`— y comprueban lo que deja: el
// binario en el directorio de binarios de Go, las skills y su manifiesto en
// ~/.agents/skills/ y un enlace relativo por skill en ~/.claude/skills/ (H19:
// FR-125, FR-126, SC-018; contracts/skills-e-invocacion.md §5). También fijan
// qué código lleva la copia del árbol en la que se instala, que enumera go list.
//
// Llevan la etiqueta integration porque ejecutan make y el go command y compilan
// el binario: no son tests unitarios rápidos, y por eso make ci las ejecuta con
// test-integration y el lint las alcanza con run.build-tags (research.md D21).
//
// Nunca se ejecutan sobre el repositorio real: cada guion instala una copia
// mínima del árbol, con el directorio personal y los de binarios de Go dentro de
// su propio directorio de trabajo, que testscript borra al terminar (FR-055,
// research.md D15).
package skills_test

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/fs"
	"os"
	"os/exec"
	"path/filepath"
	"slices"
	"strings"
	"testing"

	"github.com/rogpeppe/go-internal/testscript"
	"github.com/stretchr/testify/require"
)

const (
	// raizDelRepositorio es la del árbol de quien ejecuta los tests, relativa al
	// directorio de este paquete. Solo se lee de ella: ningún guion escribe en
	// ese árbol (research.md D15 de H5).
	raizDelRepositorio = "../.."

	// paqueteDelBinario es el paquete principal del binario que instala la
	// receta, relativo a la raíz del repositorio.
	paqueteDelBinario = "./cmd/kitlegal"

	// directorioDeGuiones es donde viven los guiones de la instalación, relativo
	// al directorio de este paquete.
	directorioDeGuiones = "testdata/script"

	// Las carpetas del directorio de trabajo de cada guion que hacen de
	// repositorio, de directorio personal, de directorio de binarios de Go y de
	// GOPATH (contrato instalacion §4). Los guiones las nombran igual.
	carpetaDelRepositorio = "repo"
	carpetaPersonal       = "home"
	carpetaDeBinarios     = "gobin"
	carpetaDeGo           = "gopath"

	// permisosDelPropietario son los bits de permiso que la copia conserva de
	// cada fichero: los de su propietario, que dejan ejecutable lo que lo es en
	// el árbol y cualquier otro fichero en 0o600.
	permisosDelPropietario fs.FileMode = 0o700
)

// rutasDeLaInstalacion es lo que la copia lleva del repositorio además del
// código del binario, que da ficherosDelBinario con lo que empotra skills.go: los
// ficheros que lee la receta y skills/ entera, tal cual, como la tiene un clon,
// que es contra lo que los guiones comparan lo instalado y adonde apuntaba el
// enlace que dejaba el make install anterior (contracts/skills-e-invocacion.md
// §5).
var rutasDeLaInstalacion = []string{"Makefile", "go.mod", "go.sum", "skills"}

// cachesDeGo son las cachés de módulos y de construcción del proceso de test,
// que cada guion reutiliza para no descargar nada y no recompilar lo que el
// propio `go test` ya compiló (contrato instalacion §4).
type cachesDeGo struct {
	Modulos    string `json:"GOMODCACHE"`
	Compilados string `json:"GOCACHE"`
}

// paqueteListado es lo que se lee de cada paquete de `go list -deps -json`.
type paqueteListado struct {
	ImportPath string
	Dir        string
	GoFiles    []string
	EmbedFiles []string
	Module     *struct{ Main bool }
}

// guionDelEnlaceRoto es el guion de la instalación contra un enlace roto con el
// nombre de la skill, que TestInstalacion escribe en un directorio temporal y
// ejecuta como los de directorioDeGuiones (contracts/skills-e-invocacion.md §5).
const guionDelEnlaceRoto = `# Un enlace roto con el nombre de la skill en ~/.claude/skills/ —el que dejaba
# el make install anterior cuando el clon ya no está— también es un conflicto:
# make install termina con código distinto de 0 nombrándolo como «enlace roto»,
# no lo modifica y no crea nada en ~/.agents (FR-125, FR-126; FR-041;
# contracts/skills-e-invocacion.md §5).
#
# HOME, GOBIN y GOPATH son carpetas de $WORK y el árbol es la copia mínima de
# $WORK/repo (FR-055 de H5; contrato instalacion §4 de H5).

mkdir $HOME/.claude/skills
symlink $HOME/.claude/skills/boe-legislacion -> $WORK/sin-destino

! exec make -C $WORK/repo install
stderr '^enlace roto: '${WORK@R}'/home/\.claude/skills/boe-legislacion$'

# El enlace conserva su destino, que sigue sin existir, y es lo único que hay
# en ~/.claude/skills/.
exec readlink $HOME/.claude/skills/boe-legislacion
stdout '\A'${WORK@R}'/sin-destino\n\z'
! exists $WORK/sin-destino
exec ls -A $HOME/.claude/skills
stdout '\Aboe-legislacion\n\z'

# Y nada se ha instalado en ~/.agents.
! exists $HOME/.agents
`

// TestInstalacion ejecuta los guiones de la instalación —en limpio, repetida,
// contra el enlace de la instalación anterior, sin GOBIN y contra un enlace
// roto— sobre una copia mínima del repositorio en $WORK/repo (H19: US2 escenario
// 3, FR-125, FR-126, SC-018; contracts/skills-e-invocacion.md §5). Los cuatro
// primeros son los de directorioDeGuiones; el del enlace roto,
// guionDelEnlaceRoto, lo escribe el test en un directorio temporal.
//
// Setup deja en cada directorio de trabajo el árbol que basta para make install
// y el entorno de la tabla del contrato: HOME, GOBIN y GOPATH son carpetas de
// $WORK, las cachés son las del proceso de test, GOPROXY=off hace fallar el guion
// antes que tocar la red, GOENV=off no lee la configuración de Go de ninguna
// cuenta y GOFLAGS=-mod=readonly no reescribe go.mod ni go.sum de la copia.
// RequireExplicitExec obliga a escribir con `exec` cada orden externa, de modo
// que una línea mal escrita no ejecute un programa creyendo que es una orden
// integrada.
func TestInstalacion(t *testing.T) {
	t.Parallel()

	rutas := slices.Concat(rutasDeLaInstalacion, ficherosDelBinario(t, raizDelRepositorio, os.Environ()))
	caches := leerCachesDeGo(t)

	temporales := t.TempDir()
	require.NoError(t, os.WriteFile(filepath.Join(temporales, "instalar-con-enlace-roto.txtar"),
		[]byte(guionDelEnlaceRoto), 0o600))

	for _, guiones := range []string{directorioDeGuiones, temporales} {
		testscript.Run(t, testscript.Params{
			Dir:                 guiones,
			RequireExplicitExec: true,
			Setup: func(env *testscript.Env) error {
				return prepararInstalacion(env, rutas, caches)
			},
		})
	}
}

// prepararInstalacion deja en el directorio de trabajo de un guion la copia
// mínima del repositorio y un directorio personal vacío, y fija el entorno de la
// instalación (contrato instalacion §4).
func prepararInstalacion(env *testscript.Env, rutas []string, caches cachesDeGo) error {
	personal := filepath.Join(env.WorkDir, carpetaPersonal)

	env.Setenv("HOME", personal)
	env.Setenv("GOBIN", filepath.Join(env.WorkDir, carpetaDeBinarios))
	env.Setenv("GOPATH", filepath.Join(env.WorkDir, carpetaDeGo))
	env.Setenv("GOMODCACHE", caches.Modulos)
	env.Setenv("GOCACHE", caches.Compilados)
	env.Setenv("GOPROXY", "off")
	env.Setenv("GOENV", "off")
	env.Setenv("GOFLAGS", "-mod=readonly")

	return errors.Join(
		os.Mkdir(personal, 0o750),
		copiarArbol(raizDelRepositorio, filepath.Join(env.WorkDir, carpetaDelRepositorio), rutas),
	)
}

// TestFicherosDelBinario fija que ficherosDelBinario da los mismos ficheros
// cuando el repositorio se alcanza por una ruta con un enlace simbólico, sea la
// de la raíz o la de los Dir que devuelve go list (contrato instalacion §4). En
// macOS basta con clonar bajo /tmp o bajo el temporal de mktemp, que cuelgan de
// /private. Los dos casos usan un enlace de t.TempDir() al repositorio:
//
//   - raiz-por-un-enlace: la raíz es el enlace y go list corre sin PWD, así que
//     da cada Dir por la ruta física;
//   - dir-por-un-enlace: la raíz es la ruta física y PWD nombra el enlace, que
//     go list toma porque es su directorio de trabajo, así que da cada Dir por el
//     enlace.
//
// La referencia es la lista desde la ruta física con go list sin PWD, en la que
// la raíz y cada Dir ya se escriben igual.
func TestFicherosDelBinario(t *testing.T) {
	t.Parallel()

	fisica := rutaFisica(t, raizDelRepositorio)
	enlace := filepath.Join(t.TempDir(), "repositorio")
	require.NoError(t, os.Symlink(fisica, enlace))

	sinPWD := slices.DeleteFunc(os.Environ(), func(variable string) bool {
		return strings.HasPrefix(variable, prefijoDePWD)
	})
	referencia := ficherosDelBinario(t, fisica, sinPWD)

	casos := []struct {
		nombre  string
		raiz    string
		entorno []string
	}{
		{nombre: "raiz-por-un-enlace", raiz: enlace, entorno: sinPWD},
		{nombre: "dir-por-un-enlace", raiz: fisica, entorno: slices.Concat(sinPWD, []string{prefijoDePWD + enlace})},
	}

	for _, caso := range casos {
		t.Run(caso.nombre, func(t *testing.T) {
			t.Parallel()

			require.Equal(t, referencia, ficherosDelBinario(t, caso.raiz, caso.entorno))
		})
	}
}

// prefijoDePWD es el de la variable de entorno con la que un proceso hereda la
// ruta de su directorio de trabajo, que puede llevar enlaces simbólicos.
const prefijoDePWD = "PWD="

// ficherosDelBinario son los ficheros .go y embebidos de cada paquete del módulo
// del que depende el paquete principal del binario, por su ruta dentro del
// repositorio y con barras (contrato instalacion §4).
//
// Los enumera `go list -deps` desde la raíz, con el entorno que recibe y
// CGO_ENABLED=0, como compila la receta, de modo que la copia lleve exactamente
// los ficheros que esa compilación elige. Exige que la lista traiga el propio
// punto de entrada: sin eso, una copia vacía de código haría fallar los guiones
// por una razón que no es la instalación.
//
// La raíz y el Dir de cada paquete se comparan por su ruta física. Cuando el
// repositorio se alcanza por una ruta con un enlace simbólico, cada lado la
// escribe a su manera: la raíz relativa se resuelve contra el directorio de
// trabajo del test, que go test escribe por el enlace, y go list escribe cada Dir
// por la ruta de su PWD si nombra su directorio de trabajo y por la física si no.
// Sin resolver los enlaces de los dos, un paquete del repositorio parecería fuera
// de él.
func ficherosDelBinario(t *testing.T, raiz string, entorno []string) []string {
	t.Helper()

	fisica := rutaFisica(t, raiz)

	orden := exec.CommandContext(t.Context(), "go", "list", "-deps", "-json=ImportPath,Dir,GoFiles,EmbedFiles,Module",
		paqueteDelBinario)
	orden.Dir = raiz
	orden.Env = slices.Concat(entorno, []string{"CGO_ENABLED=0"})

	var (
		ficheros     []string
		carpetas     []string
		decodificado = json.NewDecoder(bytes.NewReader(salidaDeGo(t, orden)))
	)

	for {
		var paquete paqueteListado

		err := decodificado.Decode(&paquete)
		if errors.Is(err, io.EOF) {
			break
		}

		require.NoError(t, err)

		if paquete.Module == nil || !paquete.Module.Main {
			continue
		}

		carpeta, err := filepath.Rel(fisica, rutaFisica(t, paquete.Dir))
		require.NoError(t, err)
		require.Truef(t, filepath.IsLocal(carpeta), "el paquete %s del módulo está fuera del repositorio %s: %s",
			paquete.ImportPath, fisica, paquete.Dir)

		carpetas = append(carpetas, filepath.ToSlash(carpeta))

		for _, fichero := range slices.Concat(paquete.GoFiles, paquete.EmbedFiles) {
			ficheros = append(ficheros, filepath.ToSlash(filepath.Join(carpeta, fichero)))
		}
	}

	require.Contains(t, carpetas, paqueteDelBinario[len("./"):],
		"go list -deps no devolvió el propio paquete principal del binario")

	return ficheros
}

// rutaFisica es la ruta absoluta de un directorio que existe, sin ningún enlace
// simbólico.
func rutaFisica(t *testing.T, ruta string) string {
	t.Helper()

	absoluta, err := filepath.Abs(ruta)
	require.NoError(t, err)

	fisica, err := filepath.EvalSymlinks(absoluta)
	require.NoError(t, err)

	return fisica
}

// leerCachesDeGo son las cachés de módulos y de construcción que da `go env`
// en el proceso de test.
func leerCachesDeGo(t *testing.T) cachesDeGo {
	t.Helper()

	orden := exec.CommandContext(t.Context(), "go", "env", "-json", "GOMODCACHE", "GOCACHE")
	orden.Dir = raizDelRepositorio

	var caches cachesDeGo

	require.NoError(t, json.Unmarshal(salidaDeGo(t, orden), &caches))
	require.NotEmpty(t, caches.Modulos, "go env no devolvió GOMODCACHE")
	require.NotEmpty(t, caches.Compilados, "go env no devolvió GOCACHE")

	return caches
}

// salidaDeGo ejecuta una orden del go command y devuelve su salida estándar, o
// hace fallar el test con su salida de error.
func salidaDeGo(t *testing.T, orden *exec.Cmd) []byte {
	t.Helper()

	salida, err := orden.Output()
	if err != nil {
		var fallo *exec.ExitError
		if errors.As(err, &fallo) {
			t.Fatalf("%s falló: %v\n%s", orden, err, fallo.Stderr)
		}

		t.Fatalf("%s falló: %v", orden, err)
	}

	return salida
}

// copiarArbol copia del árbol de origen al de destino cada ruta —un fichero o
// una carpeta entera—, creando lo que falte de las carpetas que la contienen.
// Los dos árboles se abren como os.Root, que no deja salir de ellos por un
// enlace mientras se copia.
func copiarArbol(origen, destino string, rutas []string) (err error) {
	if err := os.MkdirAll(destino, 0o750); err != nil {
		return fmt.Errorf("instalación: no se pudo crear la copia %s: %w", destino, err)
	}

	desde, err := os.OpenRoot(origen)
	if err != nil {
		return fmt.Errorf("instalación: no se pudo abrir el repositorio %s: %w", origen, err)
	}

	defer func() { err = errors.Join(err, desde.Close()) }()

	hacia, err := os.OpenRoot(destino)
	if err != nil {
		return fmt.Errorf("instalación: no se pudo abrir la copia %s: %w", destino, err)
	}

	defer func() { err = errors.Join(err, hacia.Close()) }()

	for _, ruta := range rutas {
		err := fs.WalkDir(desde.FS(), ruta, func(visitada string, entrada fs.DirEntry, err error) error {
			if err != nil {
				return err
			}

			return copiarEntrada(desde, hacia, filepath.FromSlash(visitada), entrada)
		})
		if err != nil {
			return fmt.Errorf("instalación: no se pudo copiar %s de %s en %s: %w", ruta, origen, destino, err)
		}
	}

	return nil
}

// copiarEntrada copia una entrada del árbol de origen en la misma ruta del de
// destino. Un enlace se recrea con su destino literal, sin seguirlo, porque en
// el árbol puede no resolver, y un fichero conserva los permisos de su
// propietario.
func copiarEntrada(desde, hacia *os.Root, ruta string, entrada fs.DirEntry) error {
	estado, err := entrada.Info()
	if err != nil {
		return err
	}

	if err := hacia.MkdirAll(filepath.Dir(ruta), 0o750); err != nil {
		return err
	}

	switch modo := estado.Mode(); {
	case modo.IsDir():
		return hacia.MkdirAll(ruta, 0o750)
	case modo&fs.ModeSymlink != 0:
		enlazado, err := desde.Readlink(ruta)
		if err != nil {
			return err
		}

		return hacia.Symlink(enlazado, ruta)
	case modo.IsRegular():
		contenido, err := desde.ReadFile(ruta)
		if err != nil {
			return err
		}

		return hacia.WriteFile(ruta, contenido, modo.Perm()&permisosDelPropietario)
	default:
		return fmt.Errorf("%s no es un directorio, un fichero regular ni un enlace simbólico", ruta)
	}
}
