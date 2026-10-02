//go:build snapshot

package kitlegal_test

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"io/fs"
	"os"
	"os/exec"
	"path"
	"path/filepath"
	"runtime"
	"slices"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// Este fichero solo se compila con la etiqueta snapshot: lo ejecuta make
// snapshot-check sobre el dist/ que deja make release, en local y en el trabajo
// snapshot de la integración continua; make test no lo ve, porque sin make
// release no hay dist/ (contracts/release.md §3 y §4).

const (
	// carpetaDelSnapshot es donde make release deja el snapshot, relativa a la
	// raíz del módulo, que es el directorio de este paquete y donde go test
	// ejecuta sus tests.
	carpetaDelSnapshot = "dist"

	// artefactosDelSnapshot y metadatosDelSnapshot son los ficheros en los que
	// goreleaser describe lo que construyó y con qué versión y commit
	// (research.md V15).
	artefactosDelSnapshot = "artifacts.json"
	metadatosDelSnapshot  = "metadata.json"

	// checksumsDelSnapshot es el fichero de checksums (contracts/release.md §2).
	checksumsDelSnapshot = "checksums.txt"

	// separadorDeChecksums separa la huella del nombre en cada línea de
	// checksums.txt: `<sha256>  <nombre>`.
	separadorDeChecksums = "  "

	// tipoArchivo y tipoBinario son los tipos de artifacts.json de un archivo de
	// distribución y de un binario (research.md V15).
	tipoArchivo = "Archive"
	tipoBinario = "Binary"
)

// archivosDeDistribucion son los seis archivos del snapshot y de la release:
// tar.gz en darwin y linux, zip en windows, sin versión en el nombre (SC-016;
// research.md D22).
var archivosDeDistribucion = []string{
	"kitlegal_darwin_amd64.tar.gz",
	"kitlegal_darwin_arm64.tar.gz",
	"kitlegal_linux_amd64.tar.gz",
	"kitlegal_linux_arm64.tar.gz",
	"kitlegal_windows_amd64.zip",
	"kitlegal_windows_arm64.zip",
}

// sufijosQueElSnapshotOmite son los de los SBOM y las firmas que la release
// genera y el snapshot no (FR-095, SC-016).
var sufijosQueElSnapshotOmite = []string{".sbom.json", ".sig", ".sigstore.json", ".pem"}

// artefactoDelSnapshot es una entrada de artifacts.json, con lo que el test lee
// de ella.
type artefactoDelSnapshot struct {
	Name   string `json:"name"`
	Path   string `json:"path"`
	Goos   string `json:"goos"`
	Goarch string `json:"goarch"`
	Type   string `json:"type"`
}

// metadatosDeLaConstruccion son la versión y el commit del snapshot, de
// metadata.json.
type metadatosDeLaConstruccion struct {
	Version string `json:"version"`
	Commit  string `json:"commit"`
}

// snapshotLeido es el dist/ que el test comprueba: la raíz del módulo, desde la
// que artifacts.json da las rutas, y lo que describen artifacts.json y
// metadata.json.
type snapshotLeido struct {
	raiz       fs.FS
	artefactos []artefactoDelSnapshot
	metadatos  metadatosDeLaConstruccion
}

// TestSnapshot comprueba el dist/ que deja make release (contracts/release.md
// §4; FR-095, FR-120, SC-016) y falla nombrando lo que falta o sobra: los tipos
// Archive de artifacts.json son exactamente los seis archivos, en dist/; cada
// uno tiene su línea en checksums.txt con la huella calculada sobre el fichero,
// sin fijar el total de líneas; en dist/ no hay SBOM ni firmas; y el binario de
// la plataforma que ejecuta el test imprime en version la versión y el commit de
// metadata.json. Sin archivo para esa plataforma, falla.
//
// Desde H22 comprueba además las dos piezas con las que kitlegal se instala sin
// terminal, kitlegal.mcpb y kitlegal-plugin.zip, contra ese mismo binario: sus
// seis subpruebas están en snapshot_piezas_test.go (H22 contracts/release.md
// §3; FR-060 a FR-064, FR-066).
func TestSnapshot(t *testing.T) {
	t.Parallel()

	snapshot := leerSnapshot(t, os.DirFS("."))

	casos := []struct {
		nombre string
		probar func(*testing.T, snapshotLeido)
	}{
		{"seis-archivos", probarSeisArchivos},
		{"checksums-de-los-archivos", probarChecksumsDeLosArchivos},
		{"sin-sbom-ni-firmas", probarSinSBOMNiFirmas},
		{"version-del-binario", probarVersionDelBinario},
		{"dos-piezas", probarDosPiezas},
		{"manifiesto-de-la-extension", probarManifiestoDeLaExtension},
		{"binarios-de-la-extension", probarBinariosDeLaExtension},
		{"icono-de-la-extension", probarIconoDeLaExtension},
		{"servidor-de-la-extension", probarServidorDeLaExtension},
		{"skills-del-plugin", probarSkillsDelPlugin},
	}

	for _, caso := range casos {
		t.Run(caso.nombre, func(t *testing.T) {
			t.Parallel()

			caso.probar(t, snapshot)
		})
	}
}

// leerSnapshot lee artifacts.json y metadata.json del dist/ de la raíz.
func leerSnapshot(t *testing.T, raiz fs.FS) snapshotLeido {
	t.Helper()

	leido := snapshotLeido{raiz: raiz}

	leerJSONDelSnapshot(t, raiz, artefactosDelSnapshot, &leido.artefactos)
	leerJSONDelSnapshot(t, raiz, metadatosDelSnapshot, &leido.metadatos)

	return leido
}

// leerJSONDelSnapshot decodifica un fichero JSON de dist/ en destino.
func leerJSONDelSnapshot(t *testing.T, raiz fs.FS, fichero string, destino any) {
	t.Helper()

	ruta := path.Join(carpetaDelSnapshot, fichero)

	contenido, err := fs.ReadFile(raiz, ruta)
	require.NoError(t, err, "%s: el snapshot lo construye make release", ruta)
	require.NoError(t, json.Unmarshal(contenido, destino), "%s no es el JSON de goreleaser", ruta)
}

// deTipo son los artefactos del snapshot del tipo dado, en el orden de
// artifacts.json.
func (s snapshotLeido) deTipo(tipo string) []artefactoDelSnapshot {
	return slices.DeleteFunc(slices.Clone(s.artefactos), func(artefacto artefactoDelSnapshot) bool {
		return artefacto.Type != tipo
	})
}

// probarSeisArchivos: los archivos de distribución son exactamente los seis, y
// cada uno está en dist/ (SC-016).
func probarSeisArchivos(t *testing.T, snapshot snapshotLeido) {
	t.Helper()

	var nombres []string

	for _, archivo := range snapshot.deTipo(tipoArchivo) {
		nombres = append(nombres, archivo.Name)

		assert.Equalf(t, path.Join(carpetaDelSnapshot, archivo.Name), filepath.ToSlash(archivo.Path),
			"el archivo %s no está en %s/", archivo.Name, carpetaDelSnapshot)
	}

	assert.ElementsMatch(t, archivosDeDistribucion, nombres,
		"los archivos de distribución del snapshot no son exactamente los seis (SC-016)")
}

// probarChecksumsDeLosArchivos: cada archivo tiene en checksums.txt una línea
// `<sha256>  <nombre>`, con su nombre exacto, cuya huella es la del fichero; el
// resto de líneas —paquetes, install.sh— no se fija (SC-016).
func probarChecksumsDeLosArchivos(t *testing.T, snapshot snapshotLeido) {
	t.Helper()

	ruta := path.Join(carpetaDelSnapshot, checksumsDelSnapshot)

	contenido, err := fs.ReadFile(snapshot.raiz, ruta)
	require.NoError(t, err)

	huellas := map[string]string{}

	for linea := range strings.Lines(string(contenido)) {
		huella, nombre, conSeparador := strings.Cut(strings.TrimSuffix(linea, "\n"), separadorDeChecksums)
		require.Truef(t, conSeparador, "%s tiene una línea sin la forma `<sha256>  <nombre>`: %q", ruta, linea)

		huellas[nombre] = huella
	}

	for _, archivo := range archivosDeDistribucion {
		huella, conLinea := huellas[archivo]
		if assert.Truef(t, conLinea, "%s no tiene la línea de %s (SC-016)", ruta, archivo) {
			assert.Equalf(t, huellaDelFichero(t, snapshot.raiz, path.Join(carpetaDelSnapshot, archivo)), huella,
				"la huella de %s en %s no es la del fichero (SC-016)", archivo, ruta)
		}
	}
}

// huellaDelFichero es el SHA-256 del fichero, en hexadecimal.
func huellaDelFichero(t *testing.T, raiz fs.FS, ruta string) string {
	t.Helper()

	contenido, err := fs.ReadFile(raiz, ruta)
	require.NoError(t, err)

	suma := sha256.Sum256(contenido)

	return hex.EncodeToString(suma[:])
}

// probarSinSBOMNiFirmas: en ningún lugar de dist/ hay un SBOM ni una firma
// (FR-095, SC-016).
func probarSinSBOMNiFirmas(t *testing.T, snapshot snapshotLeido) {
	t.Helper()

	var sobrantes []string

	err := fs.WalkDir(snapshot.raiz, carpetaDelSnapshot, func(ruta string, entrada fs.DirEntry, err error) error {
		if err != nil {
			return err
		}

		if slices.ContainsFunc(sufijosQueElSnapshotOmite, func(sufijo string) bool {
			return strings.HasSuffix(entrada.Name(), sufijo)
		}) {
			sobrantes = append(sobrantes, ruta)
		}

		return nil
	})
	require.NoError(t, err)

	assert.Empty(t, sobrantes, "el snapshot no genera SBOM ni firmas (FR-095, SC-016)")
}

// probarVersionDelBinario: la plataforma que ejecuta el test tiene su archivo, y
// su binario imprime en version la versión y el commit de metadata.json
// (FR-092, SC-016). En el trabajo snapshot de la integración continua, esa
// plataforma es linux/amd64, y el registro del test lo dice (research.md S15).
func probarVersionDelBinario(t *testing.T, snapshot snapshotLeido) {
	t.Helper()

	extension := ".tar.gz"
	if runtime.GOOS == "windows" {
		extension = ".zip"
	}

	archivo := nombreDelProyecto + "_" + runtime.GOOS + "_" + runtime.GOARCH + extension
	require.Containsf(t, nombresDe(snapshot.deTipo(tipoArchivo)), archivo,
		"el snapshot no tiene archivo para %s/%s, la plataforma que ejecuta el test", runtime.GOOS, runtime.GOARCH)

	binarios := slices.DeleteFunc(snapshot.deTipo(tipoBinario), func(binario artefactoDelSnapshot) bool {
		return binario.Goos != runtime.GOOS || binario.Goarch != runtime.GOARCH
	})
	require.Lenf(t, binarios, 1, "artifacts.json no da un solo binario para %s/%s", runtime.GOOS, runtime.GOARCH)

	require.NotEmpty(t, snapshot.metadatos.Version, "%s no da la versión", metadatosDelSnapshot)
	require.NotEmpty(t, snapshot.metadatos.Commit, "%s no da el commit", metadatosDelSnapshot)

	t.Logf("binario ejecutado: %s/%s, %s", runtime.GOOS, runtime.GOARCH, binarios[0].Path)

	binario, err := filepath.Abs(binarios[0].Path)
	require.NoError(t, err)

	lineas := strings.Split(salidaDeVersion(t, binario), "\n")
	require.GreaterOrEqual(t, len(lineas), 2, "version no escribe sus tres líneas")

	assert.Equal(t, "kitlegal "+snapshot.metadatos.Version, lineas[0],
		"el binario del snapshot no imprime la versión de %s (FR-092)", metadatosDelSnapshot)
	assert.Equal(t, "commit: "+snapshot.metadatos.Commit, lineas[1],
		"el binario del snapshot no imprime el commit de %s (FR-092)", metadatosDelSnapshot)
}

// nombresDe son los nombres de los artefactos.
func nombresDe(artefactos []artefactoDelSnapshot) []string {
	nombres := make([]string, 0, len(artefactos))
	for _, artefacto := range artefactos {
		nombres = append(nombres, artefacto.Name)
	}

	return nombres
}

// salidaDeVersion es lo que el binario escribe en la salida estándar con el
// verbo version.
func salidaDeVersion(t *testing.T, binario string) string {
	t.Helper()

	salida, err := exec.CommandContext(t.Context(), binario, "version").Output()
	require.NoError(t, err)

	return string(salida)
}
