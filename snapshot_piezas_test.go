//go:build snapshot

package kitlegal_test

import (
	"archive/tar"
	"archive/zip"
	"bytes"
	"compress/gzip"
	"context"
	"debug/macho"
	"encoding/json"
	"errors"
	"fmt"
	"image/png"
	"io"
	"io/fs"
	"os"
	"os/exec"
	"path"
	"path/filepath"
	"runtime"
	"slices"
	"strings"
	"testing"
	"time"
	"unicode/utf8"

	"github.com/santhosh-tekuri/jsonschema/v6"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	paso "github.com/jmorenobl/kitlegal/internal/empaquetado"
	"github.com/jmorenobl/kitlegal/internal/mcp/mcptest"
)

// Este fichero lleva las subpruebas de TestSnapshot que comprueban las dos
// piezas con las que kitlegal se instala sin terminal —kitlegal.mcpb, la
// extensión de escritorio, y kitlegal-plugin.zip, el plugin de Claude— contra
// el binario real del snapshot (H22 contracts/release.md §3; research.md D9).
// Y, fuera de TestSnapshot, TestPluginValido, que valida con Claude Code el
// plugin y el catálogo de su versión (H22 contracts/release.md §4; research.md
// D10). El paquete del paso se importa como `paso` porque release_test.go ya
// declara un tipo empaquetado (research.md D19).

// Lo que el contrato fija de las dos piezas, escrito aquí como lo escriben
// data-model §2 a §4 de H22 y no leído del paquete del paso: es contra lo que
// se compara.
const (
	// extensionDelSnapshot y pluginDelSnapshot son los dos ficheros que el paso
	// deja en dist/ (FR-002).
	extensionDelSnapshot = "kitlegal.mcpb"
	pluginDelSnapshot    = "kitlegal-plugin.zip"

	// Las cuatro entradas de la extensión (FR-010).
	manifiestoDeLaExtension = "manifest.json"
	iconoDeLaExtension      = "icon.png"
	servidorDeLaExtension   = "server/kitlegal"
	servidorDeWindows       = "server/kitlegal.exe"

	// fichaDelPlugin es la entrada del plugin que no es de ninguna skill
	// (FR-020).
	fichaDelPlugin = ".claude-plugin/plugin.json"

	// catalogoDelPlugin es el catálogo desde el que se instala el plugin, con
	// la ruta que tiene en su repositorio y en la que `claude plugin validate`
	// lo busca dentro de una carpeta (FR-030; research.md V12 de H22).
	catalogoDelPlugin = ".claude-plugin/marketplace.json"

	// paginaDeLasPiezas es la web que declaran el manifiesto y plugin.json, que
	// no es la página del repositorio que declaran el cask, el bucket y los
	// paquetes.
	paginaDeLasPiezas = "https://kitlegal.es"

	// carpetaDeLaExtension es lo que la app sustituye, en la orden de
	// `mcp_config`, por la carpeta en la que extrae la extensión.
	carpetaDeLaExtension = "${__dirname}"

	// iconoVersionado es el icono de la extensión en el árbol, y ladoDelIcono,
	// lo que mide cada lado, en píxeles (FR-016).
	iconoVersionado = "mcp/icon.png"
	ladoDelIcono    = 512

	// maximoDeLaDescripcion son los caracteres que la descripción corta puede
	// tener como mucho (FR-015).
	maximoDeLaDescripcion = 120

	// binarioDeWindows es el del archivo de Windows, y archivoDeWindows, el de
	// la única arquitectura de Windows que la extensión lleva (FR-011).
	binarioDeWindows = "kitlegal.exe"
	archivoDeWindows = "kitlegal_windows_amd64.zip"

	// carpetaDeLaInstalacion es donde `kitlegal skills install` deja las skills
	// en un proyecto, y manifiestoDeLaInstalacion, el fichero que escribe junto
	// a ellas y que no es de ninguna skill (H19).
	carpetaDeLaInstalacion    = ".agents/skills"
	manifiestoDeLaInstalacion = "kitlegal.json"

	// topeDelServidor es lo más que se espera a que el servidor de la extensión
	// salude, liste sus herramientas y termine, y esperaDeSusSalidas, lo más que
	// se espera después a que sus salidas se cierren.
	topeDelServidor    = time.Minute
	esperaDeSusSalidas = 5 * time.Second
)

// piezasDelSnapshot son los dos ficheros que el paso deja en dist/ (FR-060).
var piezasDelSnapshot = []string{extensionDelSnapshot, pluginDelSnapshot}

// entradasDeLaExtension son las cuatro entradas de la extensión, y ninguna más
// (FR-010, FR-062).
var entradasDeLaExtension = []string{
	manifiestoDeLaExtension, iconoDeLaExtension, servidorDeLaExtension, servidorDeWindows,
}

// argumentosDelServidor son los de `mcp_config`: `kitlegal mcp serve` (FR-012).
var argumentosDelServidor = []string{"mcp", "serve"}

// arquitecturasDeMacOS son, por su tipo de CPU, las dos arquitecturas del
// universal, con el nombre que les dan Go y los archivos de macOS (FR-011).
var arquitecturasDeMacOS = map[macho.Cpu]string{macho.CpuAmd64: "amd64", macho.CpuArm64: "arm64"}

// herramientasDeHoy son las diez herramientas que el servidor MCP anuncia hoy,
// escritas a mano (FR-014, SC-004). Tienen que estar; una herramienta nueva
// del registro llega a `tools` sin tocar este test (US5).
var herramientasDeHoy = []string{
	"boe_buscar", "boe_indice", "boe_articulo", "boe_articulos", "boe_metadatos", "boe_analisis",
	"graph_show", "graph_stats", "graph_check",
	"territorio_resolver",
}

// manifiestoLeido es el manifiesto como lo fija data-model §3 de H22: sus
// campos y ninguno más. Es la transcripción del contrato con la que se lee de
// forma estricta el del snapshot; no es el tipo del paquete del paso. Sin
// `user_config`, que la lectura estricta rechaza (FR-012, FR-017).
type manifiestoLeido struct {
	VersionDelManifiesto string              `json:"manifest_version"`
	Nombre               string              `json:"name"`
	NombreVisible        string              `json:"display_name"`
	Version              string              `json:"version"`
	Descripcion          string              `json:"description"`
	DescripcionLarga     string              `json:"long_description"`
	Autoria              autoriaLeida        `json:"author"`
	Pagina               string              `json:"homepage"`
	Licencia             string              `json:"license"`
	Icono                string              `json:"icon"`
	Servidor             servidorLeido       `json:"server"`
	Herramientas         []herramientaLeida  `json:"tools"`
	Compatibilidad       compatibilidadLeida `json:"compatibility"`
}

// autoriaLeida es `author` del manifiesto y de plugin.json.
type autoriaLeida struct {
	Nombre string `json:"name"`
}

// servidorLeido es `server` del manifiesto.
type servidorLeido struct {
	Tipo     string        `json:"type"`
	Entrada  string        `json:"entry_point"`
	Arranque arranqueLeido `json:"mcp_config"`
}

// arranqueLeido es `mcp_config`: la orden con la que la app arranca el
// servidor (FR-012, FR-064).
type arranqueLeido struct {
	Orden         string            `json:"command"`
	Argumentos    []string          `json:"args"`
	PorPlataforma plataformasLeidas `json:"platform_overrides"`
}

// plataformasLeidas es `platform_overrides`: solo Windows cambia la orden, y la
// lectura estricta rechaza cualquier otra plataforma.
type plataformasLeidas struct {
	Win32 ordenLeida `json:"win32"`
}

// ordenLeida es lo que una plataforma cambia de `mcp_config`.
type ordenLeida struct {
	Orden string `json:"command"`
}

// herramientaLeida es un elemento de `tools`, y también lo que el servidor
// anuncia de una herramienta que el manifiesto dice: su nombre y su
// descripción.
type herramientaLeida struct {
	Nombre      string `json:"name"`
	Descripcion string `json:"description"`
}

// compatibilidadLeida es `compatibility`.
type compatibilidadLeida struct {
	Plataformas []string `json:"platforms"`
}

// fichaDelPluginLeida es plugin.json como lo fija data-model §4 de H22: sin
// `mcpServers`, que la lectura estricta rechaza (FR-021, FR-022).
type fichaDelPluginLeida struct {
	Nombre      string       `json:"name"`
	Version     string       `json:"version"`
	Descripcion string       `json:"description"`
	Autoria     autoriaLeida `json:"author"`
	Pagina      string       `json:"homepage"`
	Licencia    string       `json:"license"`
}

// zipDelSnapshot es un zip de dist/, abierto: una de las dos piezas o el
// archivo de Windows.
type zipDelSnapshot struct {
	ruta     string
	ficheros []*zip.File
}

// abrirZip abre un zip de dist/, que tiene que estar.
func abrirZip(t *testing.T, snapshot snapshotLeido, nombre string) zipDelSnapshot {
	t.Helper()

	ruta := path.Join(carpetaDelSnapshot, nombre)

	contenido, err := fs.ReadFile(snapshot.raiz, ruta)
	require.NoErrorf(t, err, "falta %s en el snapshot que deja make release", ruta)

	lector, err := zip.NewReader(bytes.NewReader(contenido), int64(len(contenido)))
	require.NoErrorf(t, err, "%s no es un zip", ruta)

	return zipDelSnapshot{ruta: ruta, ficheros: lector.File}
}

// entradas son los nombres de las entradas del zip, en su orden.
func (z zipDelSnapshot) entradas() []string {
	nombres := make([]string, 0, len(z.ficheros))
	for _, fichero := range z.ficheros {
		nombres = append(nombres, fichero.Name)
	}

	return nombres
}

// entrada es la del zip con ese nombre, que tiene que estar.
func (z zipDelSnapshot) entrada(t *testing.T, nombre string) *zip.File {
	t.Helper()

	indice := slices.IndexFunc(z.ficheros, func(fichero *zip.File) bool { return fichero.Name == nombre })
	require.GreaterOrEqualf(t, indice, 0, "a %s le falta la entrada %s", z.ruta, nombre)

	return z.ficheros[indice]
}

// leer da el contenido de la entrada del zip con ese nombre, que tiene que
// estar.
func (z zipDelSnapshot) leer(t *testing.T, nombre string) []byte {
	t.Helper()

	abierta, err := z.entrada(t, nombre).Open()
	require.NoErrorf(t, err, "la entrada %s de %s no se puede abrir", nombre, z.ruta)

	contenido, err := io.ReadAll(abierta)
	require.NoErrorf(t, err, "la entrada %s de %s no se puede leer", nombre, z.ruta)
	require.NoError(t, abierta.Close())

	return contenido
}

// leerEstricto lee un documento JSON en destino sin admitir un campo que su
// tipo no declare ni nada detrás del documento, y falla con lo que el
// documento lleva de más.
func leerEstricto(t *testing.T, nombre string, documento []byte, destino any) {
	t.Helper()

	lector := json.NewDecoder(bytes.NewReader(documento))
	lector.DisallowUnknownFields()

	require.NoErrorf(t, lector.Decode(destino), "%s no es el documento del contrato:\n%s", nombre, documento)

	_, err := lector.Token()
	require.ErrorIsf(t, err, io.EOF, "%s lleva algo detrás del documento", nombre)
}

// manifiestoDe es el manifiesto de la extensión para la subprueba que solo
// necesita lo que dice: que no lleve ningún campo de más lo fija
// manifiesto-de-la-extension, y no cada subprueba que lo lee.
func manifiestoDe(t *testing.T, extension zipDelSnapshot) manifiestoLeido {
	t.Helper()

	var manifiesto manifiestoLeido

	require.NoErrorf(t, json.Unmarshal(extension.leer(t, manifiestoDeLaExtension), &manifiesto),
		"%s de %s no se puede leer", manifiestoDeLaExtension, extension.ruta)

	return manifiesto
}

// esquemaOficialDelManifiesto es el esquema JSON de la versión `0.3` del
// manifiesto de MCP Bundle, tal como lo publica su fuente
// (testdata/mcpb/README.md); que el fichero es el de la fuente lo fija
// TestEsquemaOficial, en make ci.
const esquemaOficialDelManifiesto = "testdata/mcpb/mcpb-manifest-v0.3.schema.json"

// validarConElEsquemaOficial devuelve el error de validar un manifiesto contra
// el esquema oficial de la versión `0.3`.
func validarConElEsquemaOficial(t *testing.T, manifiesto []byte) error {
	t.Helper()

	const url = "https://kitlegal.es/" + esquemaOficialDelManifiesto

	leido, err := os.ReadFile(filepath.FromSlash(esquemaOficialDelManifiesto))
	require.NoError(t, err)

	esquema, err := jsonschema.UnmarshalJSON(bytes.NewReader(leido))
	require.NoError(t, err)

	compilador := jsonschema.NewCompiler()
	require.NoError(t, compilador.AddResource(url, esquema))

	compilado, err := compilador.Compile(url)
	require.NoError(t, err)

	documento, err := jsonschema.UnmarshalJSON(bytes.NewReader(manifiesto))
	require.NoError(t, err)

	return compilado.Validate(documento)
}

// binarioDeLaPlataforma es la ruta del binario del snapshot de la plataforma
// que ejecuta el test, que tiene que ser uno. En el trabajo snapshot de la
// integración continua es el de linux/amd64.
func binarioDeLaPlataforma(t *testing.T, snapshot snapshotLeido) string {
	t.Helper()

	var rutas []string

	for _, binario := range snapshot.deTipo(tipoBinario) {
		if binario.Goos == runtime.GOOS && binario.Goarch == runtime.GOARCH {
			rutas = append(rutas, binario.Path)
		}
	}

	require.Lenf(t, rutas, 1, "el snapshot no da un solo binario para %s/%s, la plataforma que ejecuta el test",
		runtime.GOOS, runtime.GOARCH)

	ruta, err := filepath.Abs(rutas[0])
	require.NoError(t, err)

	return ruta
}

// versionDelBinario es la versión que el binario imprime con el verbo version,
// sin la `v` de una etiqueta: la forma en la que el manifiesto y plugin.json
// la llevan (FR-013).
func versionDelBinario(t *testing.T, binario string) string {
	t.Helper()

	primera, _, _ := strings.Cut(salidaDeVersion(t, binario), "\n")

	version, conNombre := strings.CutPrefix(primera, nombreDelProyecto+" ")
	require.Truef(t, conNombre, "la primera línea de version no es `%s <versión>`: %q", nombreDelProyecto, primera)

	return strings.TrimPrefix(version, "v")
}

// mismosBytes falla si los dos contenidos no son los mismos bytes, diciendo de
// qué son y el tamaño de cada uno, sin volcar ninguno: son binarios de decenas
// de megas.
func mismosBytes(t *testing.T, esperado, obtenido []byte, obtenidoDe, esperadoDe string) {
	t.Helper()

	assert.Truef(t, bytes.Equal(esperado, obtenido), "%s no es, byte a byte, %s: mide %d bytes, y lo esperado, %d",
		obtenidoDe, esperadoDe, len(obtenido), len(esperado))
}

// probarDosPiezas: la extensión y el plugin están en dist/, y checksums.txt
// lleva de cada uno una línea `<sha256>  <nombre>` con la huella del fichero
// (H22 FR-060, SC-003).
func probarDosPiezas(t *testing.T, snapshot snapshotLeido) {
	t.Helper()

	checksums := path.Join(carpetaDelSnapshot, checksumsDelSnapshot)

	contenido, err := fs.ReadFile(snapshot.raiz, checksums)
	require.NoError(t, err)

	lineas := strings.Split(strings.TrimSuffix(string(contenido), "\n"), "\n")

	for _, pieza := range piezasDelSnapshot {
		ruta := path.Join(carpetaDelSnapshot, pieza)

		// Sin una pieza, la subprueba sigue con la otra: nombra las dos que
		// falten, y no solo la primera.
		if _, err := fs.Stat(snapshot.raiz, ruta); err != nil {
			assert.Failf(t, "falta una pieza del snapshot", "falta %s en el snapshot que deja make release (FR-060): %v",
				ruta, err)

			continue
		}

		deLaPieza := slices.DeleteFunc(slices.Clone(lineas), func(linea string) bool {
			_, nombre, _ := strings.Cut(linea, separadorDeChecksums)

			return nombre != pieza
		})

		assert.Equalf(t, []string{huellaDelFichero(t, snapshot.raiz, ruta) + separadorDeChecksums + pieza}, deLaPieza,
			"%s no lleva de %s una línea `<sha256>  <nombre>` con la huella del fichero (FR-060)", checksums, pieza)
	}
}

// probarManifiestoDeLaExtension: manifest.json, leído de forma estricta, lleva
// cada campo de data-model §3 con su valor y ninguno más —tampoco
// `user_config`—: la versión `0.3` del manifiesto, lo fijo, los textos del
// paquete del paso y, como `version`, la que imprime el binario del snapshot,
// sin su `v`; y su descripción corta no pasa de 120 caracteres (H22 FR-061,
// FR-066; SC-004, SC-009). `tools` tiene que estar: que sea lo que el binario
// anuncia lo fija servidor-de-la-extension. Y cumple el esquema oficial de la
// versión `0.3`, versionado en testdata/mcpb/ (FR-017).
func probarManifiestoDeLaExtension(t *testing.T, snapshot snapshotLeido) {
	t.Helper()

	extension := abrirZip(t, snapshot, extensionDelSnapshot)

	var manifiesto manifiestoLeido

	leerEstricto(t, manifiestoDeLaExtension+" de "+extension.ruta, extension.leer(t, manifiestoDeLaExtension), &manifiesto)

	require.NoErrorf(t, validarConElEsquemaOficial(t, extension.leer(t, manifiestoDeLaExtension)),
		"%s de %s no cumple el esquema oficial de la versión `0.3` del manifiesto de MCP Bundle (%s)",
		manifiestoDeLaExtension, extension.ruta, esquemaOficialDelManifiesto)

	assert.NotEmpty(t, manifiesto.Herramientas, "al manifiesto le falta `tools` (FR-061)")

	assert.Equal(t, manifiestoLeido{
		VersionDelManifiesto: "0.3",
		Nombre:               "kitlegal",
		NombreVisible:        paso.NombreVisible,
		Version:              versionDelBinario(t, binarioDeLaPlataforma(t, snapshot)),
		Descripcion:          paso.Descripcion,
		DescripcionLarga:     paso.DescripcionLarga,
		Autoria:              autoriaLeida{Nombre: paso.Autoria},
		Pagina:               paginaDeLasPiezas,
		Licencia:             licenciaDelProyecto,
		Icono:                iconoDeLaExtension,
		Servidor: servidorLeido{
			Tipo:    "binary",
			Entrada: servidorDeLaExtension,
			Arranque: arranqueLeido{
				Orden:         carpetaDeLaExtension + "/" + servidorDeLaExtension,
				Argumentos:    argumentosDelServidor,
				PorPlataforma: plataformasLeidas{Win32: ordenLeida{Orden: carpetaDeLaExtension + "/" + servidorDeWindows}},
			},
		},
		Herramientas:   manifiesto.Herramientas,
		Compatibilidad: compatibilidadLeida{Plataformas: []string{"darwin", "win32"}},
	}, manifiesto,
		"el manifiesto no lleva cada campo de data-model §3 con su valor: lo fijo, los textos del paso y la versión "+
			"que imprime el binario del snapshot, sin su `v` (FR-061)")

	assert.LessOrEqualf(t, utf8.RuneCountInString(manifiesto.Descripcion), maximoDeLaDescripcion,
		"la descripción corta del manifiesto pasa de %d caracteres (FR-066)", maximoDeLaDescripcion)
}

// probarBinariosDeLaExtension: la extensión lleva exactamente sus cuatro
// entradas; server/kitlegal, con el bit de ejecución, es un universal de dos
// arquitecturas, una amd64 y una arm64, cada una byte a byte el kitlegal del
// archivo de macOS de su arquitectura; y server/kitlegal.exe es byte a byte el
// del archivo de Windows amd64 (H22 FR-062, SC-005).
func probarBinariosDeLaExtension(t *testing.T, snapshot snapshotLeido) {
	t.Helper()

	extension := abrirZip(t, snapshot, extensionDelSnapshot)

	require.ElementsMatchf(t, entradasDeLaExtension, extension.entradas(),
		"las entradas de %s no son exactamente las cuatro de data-model §2 (FR-062)", extension.ruta)

	assert.NotZerof(t, extension.entrada(t, servidorDeLaExtension).Mode()&0o100,
		"%s no lleva el bit de ejecución en %s (FR-062)", servidorDeLaExtension, extension.ruta)

	for arquitectura, delUniversal := range arquitecturasDelUniversal(t, extension.leer(t, servidorDeLaExtension)) {
		archivo := nombreDelProyecto + "_darwin_" + arquitectura + ".tar.gz"

		mismosBytes(t, binarioDelTarGz(t, snapshot, archivo), delUniversal,
			"la arquitectura "+arquitectura+" de "+servidorDeLaExtension, nombreDelProyecto+" de "+archivo+" (FR-062)")
	}

	mismosBytes(t, abrirZip(t, snapshot, archivoDeWindows).leer(t, binarioDeWindows),
		extension.leer(t, servidorDeWindows), servidorDeWindows, binarioDeWindows+" de "+archivoDeWindows+" (FR-062)")
}

// arquitecturasDelUniversal lee con debug/macho el binario universal de macOS
// y da, por el nombre de su arquitectura, los bytes de cada una. Tienen que ser
// exactamente dos, una amd64 y una arm64, y lo son por su tipo de CPU y no por
// su posición: el orden en el que goreleaser las escribe no está fijado
// (research.md V3 y V17 de H22).
func arquitecturasDelUniversal(t *testing.T, universal []byte) map[string][]byte {
	t.Helper()

	leido, err := macho.NewFatFile(bytes.NewReader(universal))
	require.NoErrorf(t, err, "%s no es un binario universal de macOS (FR-062)", servidorDeLaExtension)

	var nombres []string

	porArquitectura := map[string][]byte{}

	for _, arquitectura := range leido.Arches {
		nombre, deMacOS := arquitecturasDeMacOS[arquitectura.Cpu]
		if !deMacOS {
			nombre = arquitectura.Cpu.String()
		}

		contenido, err := io.ReadAll(io.NewSectionReader(
			bytes.NewReader(universal), int64(arquitectura.Offset), int64(arquitectura.Size)))
		require.NoError(t, err)

		nombres = append(nombres, nombre)
		porArquitectura[nombre] = contenido
	}

	require.ElementsMatchf(t, []string{"amd64", "arm64"}, nombres,
		"%s no lleva exactamente dos arquitecturas, una amd64 y una arm64 (FR-062)", servidorDeLaExtension)

	return porArquitectura
}

// binarioDelTarGz es el binario kitlegal que un archivo tar.gz del snapshot
// lleva en su raíz.
func binarioDelTarGz(t *testing.T, snapshot snapshotLeido, archivo string) []byte {
	t.Helper()

	ruta := path.Join(carpetaDelSnapshot, archivo)

	comprimido, err := fs.ReadFile(snapshot.raiz, ruta)
	require.NoErrorf(t, err, "falta %s en el snapshot que deja make release", ruta)

	descomprimido, err := gzip.NewReader(bytes.NewReader(comprimido))
	require.NoErrorf(t, err, "%s no es un gzip", ruta)

	entradas := tar.NewReader(descomprimido)

	for {
		cabecera, err := entradas.Next()
		require.NoErrorf(t, err, "%s no lleva %s en su raíz", ruta, nombreDelProyecto)

		if cabecera.Name != nombreDelProyecto {
			continue
		}

		contenido, err := io.ReadAll(entradas)
		require.NoErrorf(t, err, "%s de %s no se puede leer", nombreDelProyecto, ruta)

		return contenido
	}
}

// probarIconoDeLaExtension: icon.png de la extensión es, byte a byte, el icono
// versionado, un PNG de 512 × 512 px (H22 FR-066, SC-009).
func probarIconoDeLaExtension(t *testing.T, snapshot snapshotLeido) {
	t.Helper()

	extension := abrirZip(t, snapshot, extensionDelSnapshot)
	icono := extension.leer(t, iconoDeLaExtension)

	versionado, err := fs.ReadFile(snapshot.raiz, iconoVersionado)
	require.NoError(t, err)

	mismosBytes(t, versionado, icono, iconoDeLaExtension+" de "+extension.ruta, iconoVersionado+" (FR-066)")

	medidas, err := png.DecodeConfig(bytes.NewReader(icono))
	require.NoErrorf(t, err, "%s de %s no es un PNG (FR-066)", iconoDeLaExtension, extension.ruta)

	assert.Equalf(t, [2]int{ladoDelIcono, ladoDelIcono}, [2]int{medidas.Width, medidas.Height},
		"%s de %s no mide %d × %d px (FR-066)", iconoDeLaExtension, extension.ruta, ladoDelIcono, ladoDelIcono)
}

// probarServidorDeLaExtension: con la extensión extraída en una carpeta cuyo
// nombre lleva espacios, la orden de `mcp_config` con `${__dirname}` resuelto a
// esa carpeta, sus argumentos —que tienen que ser `mcp` y `serve`— y la raíz
// del sistema de ficheros como directorio de trabajo, el servidor completa el
// saludo y lista, como conjunto de nombres y descripciones, exactamente `tools`
// del manifiesto, que contiene las diez herramientas de hoy (H22 FR-061,
// FR-064; SC-004, SC-006). El orden de la lista lo pone el SDK y no es el del
// registro, así que se comparan conjuntos (research.md V18).
func probarServidorDeLaExtension(t *testing.T, snapshot snapshotLeido) {
	t.Helper()

	extension := abrirZip(t, snapshot, extensionDelSnapshot)
	manifiesto := manifiestoDe(t, extension)
	arranque := manifiesto.Servidor.Arranque

	require.Equal(t, argumentosDelServidor, arranque.Argumentos,
		"los argumentos de `mcp_config` no son `mcp` y `serve` (FR-064)")

	carpeta := extraerLaExtension(t, extension, binarioDeLaPlataforma(t, snapshot))

	anunciadas := herramientasDelServidor(t, strings.ReplaceAll(arranque.Orden, carpetaDeLaExtension, carpeta))

	assert.ElementsMatch(t, manifiesto.Herramientas, anunciadas,
		"lo que el servidor de la extensión lista, por nombre y descripción, no es exactamente `tools` del manifiesto "+
			"(FR-061, FR-064)")

	nombres := make([]string, 0, len(manifiesto.Herramientas))
	for _, herramienta := range manifiesto.Herramientas {
		nombres = append(nombres, herramienta.Nombre)
	}

	assert.Subset(t, nombres, herramientasDeHoy, "a `tools` del manifiesto le falta alguna de las diez herramientas de hoy (FR-061)")
}

// extraerLaExtension extrae la extensión en una carpeta temporal cuyo nombre
// lleva espacios y un carácter que no es ASCII, y devuelve su ruta. En el
// sitio del binario de macOS pone un enlace simbólico al binario del snapshot
// de la plataforma que ejecuta el test, que es el que puede arrancar aquí. Es
// un enlace y no una copia porque un ejecutable escrito desde el proceso del
// test y lanzado a continuación da a veces «text file busy» en Linux cuando
// otra subprueba lanza un proceso a la vez, y las de TestSnapshot corren en
// paralelo (research.md V32 de H22). Lo demás va con el modo de un fichero
// cualquiera: nada de ello se ejecuta.
func extraerLaExtension(t *testing.T, extension zipDelSnapshot, binario string) string {
	t.Helper()

	carpeta := filepath.Join(t.TempDir(), "la extensión extraída")
	require.NoError(t, os.Mkdir(carpeta, 0o700))

	raiz, err := os.OpenRoot(carpeta)
	require.NoError(t, err)

	defer func() { assert.NoError(t, raiz.Close()) }()

	for _, entrada := range extension.entradas() {
		if subcarpeta := path.Dir(entrada); subcarpeta != "." {
			require.NoError(t, raiz.MkdirAll(subcarpeta, 0o700))
		}

		if entrada == servidorDeLaExtension {
			require.NoError(t, raiz.Symlink(binario, entrada))

			continue
		}

		require.NoError(t, raiz.WriteFile(entrada, extension.leer(t, entrada), 0o600))
	}

	return carpeta
}

// herramientasDelServidor arranca el servidor con esa orden y `mcp serve`,
// como la app arranca el de la extensión —con la raíz del sistema de ficheros
// como directorio de trabajo—, con la caché y el HOME en carpetas temporales, y
// da el nombre y la descripción de cada herramienta que lista. El servidor
// tiene que completar el saludo, listar y, al cerrarle la entrada, terminar
// con 0. `mcp` y `serve` van escritos aquí y no leídos del manifiesto: que el
// manifiesto dice esos dos lo comprueba quien llama.
func herramientasDelServidor(t *testing.T, orden string) []herramientaLeida {
	t.Helper()

	ctx, cancelar := context.WithTimeout(t.Context(), topeDelServidor)
	defer cancelar()

	var errores bytes.Buffer

	servidor := exec.CommandContext(ctx, orden, "mcp", "serve")
	servidor.Dir = "/"
	servidor.Env = []string{"HOME=" + t.TempDir(), "KITLEGAL_CACHE_DIR=" + t.TempDir()}
	servidor.Stderr = &errores
	servidor.WaitDelay = esperaDeSusSalidas

	entrada, err := servidor.StdinPipe()
	require.NoError(t, err)

	salida, err := servidor.StdoutPipe()
	require.NoError(t, err)

	require.NoErrorf(t, servidor.Start(), "la orden de `mcp_config` no arranca: %s", orden)

	anunciadas, errDeLaSesion := listarHerramientas(ctx, salida, entrada)
	if errDeLaSesion != nil {
		// Sin sesión nadie le ha cerrado la entrada, y no va a terminar por su
		// cuenta.
		cancelar()
	}

	errDelServidor := servidor.Wait()

	require.NoErrorf(t, errDeLaSesion, "el servidor de la extensión, arrancado con %s:\n%s", orden, errores.String())
	require.NoErrorf(t, errDelServidor, "el servidor de la extensión no termina con 0 al cerrarle la entrada:\n%s",
		errores.String())

	return anunciadas
}

// listarHerramientas hace el saludo con el cliente del SDK del protocolo, pide
// la lista de herramientas y cierra la sesión, que es cerrar la entrada del
// servidor.
func listarHerramientas(ctx context.Context, salida io.Reader, entrada io.WriteCloser) ([]herramientaLeida, error) {
	sesion, err := mcptest.Abrir(ctx, salida, entrada, false)
	if err != nil {
		return nil, fmt.Errorf("no completa el saludo: %w", err)
	}

	listadas, err := sesion.Herramientas(ctx)
	if err != nil {
		return nil, errors.Join(fmt.Errorf("no lista sus herramientas: %w", err), sesion.Cerrar(ctx))
	}

	if err := sesion.Cerrar(ctx); err != nil {
		return nil, fmt.Errorf("no deja cerrar la sesión: %w", err)
	}

	anunciadas := make([]herramientaLeida, 0, len(listadas))
	for _, listada := range listadas {
		anunciadas = append(anunciadas, herramientaLeida{Nombre: listada.Nombre, Descripcion: listada.Descripcion})
	}

	return anunciadas, nil
}

// probarSkillsDelPlugin: las entradas del plugin son exactamente plugin.json
// y, bajo skills/, los ficheros que `kitlegal skills install` del binario del
// snapshot deja en un directorio vacío, sin su manifiesto, byte a byte —de
// modo que no lleva `bin/`, `.mcp.json` ni ningún `.mcpb`—; y plugin.json,
// leído de forma estricta, lleva cada campo de data-model §4 y ninguno más
// —tampoco `mcpServers`—, con la versión y los textos del manifiesto de la
// extensión (H22 FR-063, SC-007).
func probarSkillsDelPlugin(t *testing.T, snapshot snapshotLeido) {
	t.Helper()

	plugin := abrirZip(t, snapshot, pluginDelSnapshot)
	manifiesto := manifiestoDe(t, abrirZip(t, snapshot, extensionDelSnapshot))
	instaladas := skillsQueInstala(t, binarioDeLaPlataforma(t, snapshot))

	esperadas := []string{fichaDelPlugin}
	for ruta := range instaladas {
		esperadas = append(esperadas, path.Join(carpetaDeSkills, ruta))
	}

	require.ElementsMatchf(t, esperadas, plugin.entradas(),
		"las entradas de %s no son exactamente %s y, bajo %s/, los ficheros que instala `kitlegal skills install` (FR-063)",
		plugin.ruta, fichaDelPlugin, carpetaDeSkills)

	for ruta, instalada := range instaladas {
		entrada := path.Join(carpetaDeSkills, ruta)

		mismosBytes(t, instalada, plugin.leer(t, entrada), entrada+" de "+plugin.ruta,
			"el que instala `kitlegal skills install` (FR-063)")
	}

	var ficha fichaDelPluginLeida

	leerEstricto(t, fichaDelPlugin+" de "+plugin.ruta, plugin.leer(t, fichaDelPlugin), &ficha)

	assert.Equal(t, fichaDelPluginLeida{
		Nombre:      "kitlegal",
		Version:     manifiesto.Version,
		Descripcion: manifiesto.Descripcion,
		Autoria:     manifiesto.Autoria,
		Pagina:      paginaDeLasPiezas,
		Licencia:    licenciaDelProyecto,
	}, ficha,
		"plugin.json no lleva cada campo de data-model §4 con su valor, con la versión y los textos del manifiesto de "+
			"la extensión (FR-063)")
}

// skillsQueInstala ejecuta `kitlegal skills install` con ese binario en un
// directorio vacío, con el HOME y la caché en carpetas temporales vacías —sin
// ningún host que enlazar—, y da, por su ruta <skill>/…, los ficheros que deja
// en .agents/skills/, sin el manifiesto de la instalación (research.md V19 de
// H22). Las skills de hoy tienen que estar: sin ellas, la comparación con el
// plugin pasaría en vacío.
func skillsQueInstala(t *testing.T, binario string) map[string][]byte {
	t.Helper()

	proyecto := t.TempDir()

	instalacion := exec.CommandContext(t.Context(), binario, "skills", "install")
	instalacion.Dir = proyecto
	instalacion.Env = []string{"HOME=" + t.TempDir(), "KITLEGAL_CACHE_DIR=" + t.TempDir()}

	salida, err := instalacion.CombinedOutput()
	require.NoErrorf(t, err, "`kitlegal skills install` del binario del snapshot:\n%s", salida)

	instaladas := ficherosDe(t, os.DirFS(filepath.Join(proyecto, filepath.FromSlash(carpetaDeLaInstalacion))))

	require.Containsf(t, instaladas, manifiestoDeLaInstalacion, "`kitlegal skills install` no deja %s en %s",
		manifiestoDeLaInstalacion, carpetaDeLaInstalacion)
	delete(instaladas, manifiestoDeLaInstalacion)

	for _, skill := range skillsDeHoy {
		require.Containsf(t, instaladas, path.Join(skill, ficheroDeLaSkill),
			"`kitlegal skills install` no instala la skill %s: la comparación con el plugin pasaría en vacío", skill)
	}

	return instaladas
}

// TestPluginValido valida con Claude Code lo que kitlegal publica para el
// marketplace (H22 contracts/release.md §4; FR-023, FR-030, FR-065; SC-008): el
// plugin del snapshot, extraído, y el catálogo de la versión del snapshot, cada
// uno en su carpeta —con los dos en la misma, `claude plugin validate` solo
// mira el catálogo (research.md V12 de H22)—. Falla, con lo que la orden
// escribió, si `claude` no está en el PATH o no da por válido alguno de los
// dos. No abre ninguna sesión con modelo ni usa ninguna credencial.
//
// Va fuera de TestSnapshot porque es lo único del snapshot que necesita Claude
// Code: lo ejecuta make plugin-check, y no make snapshot-check ni make ci
// (research.md D10 de H22).
func TestPluginValido(t *testing.T) {
	t.Parallel()

	snapshot := leerSnapshot(t, os.DirFS("."))

	casos := []struct {
		nombre   string
		preparar func(*testing.T, snapshotLeido) string
	}{
		{"plugin", extraerElPlugin},
		{"catalogo", escribirElCatalogo},
	}

	for _, caso := range casos {
		t.Run(caso.nombre, func(t *testing.T) {
			t.Parallel()

			validarConClaude(t, caso.preparar(t, snapshot))
		})
	}
}

// extraerElPlugin extrae el plugin del snapshot en una carpeta temporal, con el
// modo de un fichero cualquiera —nada de él se ejecuta—, y devuelve su ruta.
func extraerElPlugin(t *testing.T, snapshot snapshotLeido) string {
	t.Helper()

	plugin := abrirZip(t, snapshot, pluginDelSnapshot)
	carpeta := t.TempDir()

	raiz, err := os.OpenRoot(carpeta)
	require.NoError(t, err)

	defer func() { assert.NoError(t, raiz.Close()) }()

	for _, entrada := range plugin.entradas() {
		require.NoError(t, raiz.MkdirAll(path.Dir(entrada), 0o700))
		require.NoError(t, raiz.WriteFile(entrada, plugin.leer(t, entrada), 0o600))
	}

	return carpeta
}

// escribirElCatalogo escribe en una carpeta temporal el catálogo que da la
// orden catalogo del paso para la versión de metadata.json y la huella que
// checksums.txt da para el plugin —la orden con la que la release compone el
// de cada etiqueta (H22 contracts/release.md §6.3)—, y devuelve la ruta de la
// carpeta. La versión del snapshot no lleva la `v` de una etiqueta.
func escribirElCatalogo(t *testing.T, snapshot snapshotLeido) string {
	t.Helper()

	carpeta := t.TempDir()
	require.NoError(t, os.Mkdir(filepath.Join(carpeta, path.Dir(catalogoDelPlugin)), 0o700))

	var errores bytes.Buffer

	codigo := paso.Ejecutar([]string{
		"catalogo",
		"-version", snapshot.metadatos.Version,
		"-sha256", huellaEnChecksums(t, snapshot, pluginDelSnapshot),
		"-salida", filepath.Join(carpeta, filepath.FromSlash(catalogoDelPlugin)),
	}, &errores)
	require.Zerof(t, codigo, "la orden catalogo del paso no escribe el catálogo de la versión del snapshot (FR-030):\n%s",
		errores.String())

	return carpeta
}

// huellaEnChecksums es la huella que checksums.txt da para ese fichero del
// snapshot, que tiene que tener su línea `<sha256>  <nombre>`. Es la que lee
// quien instala, y no la calculada sobre el fichero: que las dos coinciden lo
// fija dos-piezas.
func huellaEnChecksums(t *testing.T, snapshot snapshotLeido, nombre string) string {
	t.Helper()

	ruta := path.Join(carpetaDelSnapshot, checksumsDelSnapshot)

	contenido, err := fs.ReadFile(snapshot.raiz, ruta)
	require.NoError(t, err)

	for linea := range strings.Lines(string(contenido)) {
		huella, deQuien, _ := strings.Cut(strings.TrimSuffix(linea, "\n"), separadorDeChecksums)
		if deQuien == nombre {
			return huella
		}
	}

	require.Failf(t, "falta una línea de checksums.txt", "%s no lleva la línea de %s", ruta, nombre)

	return ""
}

// validarConClaude ejecuta `claude plugin validate .` en la carpeta y falla,
// con lo que la orden escribió, si claude no está en el PATH o sale con un
// código distinto de 0. La orden solo lee lo que hay en la carpeta, sin abrir
// ninguna sesión con modelo; y su entorno son solo el PATH —con el que da con
// su intérprete— y un HOME temporal y vacío, de modo que no recibe ninguna
// credencial ni la configuración de quien ejecuta el test (FR-065).
func validarConClaude(t *testing.T, carpeta string) {
	t.Helper()

	validacion := exec.CommandContext(t.Context(), "claude", "plugin", "validate", ".")
	validacion.Dir = carpeta
	validacion.Env = []string{"HOME=" + t.TempDir(), "PATH=" + os.Getenv("PATH")}

	salida, err := validacion.CombinedOutput()
	require.NoErrorf(t, err,
		"`claude plugin validate .` no está en el PATH o no da por válido lo que kitlegal publica (FR-065):\n%s", salida)
}
