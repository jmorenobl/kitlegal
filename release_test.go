package kitlegal_test

import (
	"bytes"
	"fmt"
	"io/fs"
	"maps"
	"os"
	"os/exec"
	"path"
	"regexp"
	"slices"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.yaml.in/yaml/v3"
)

const (
	// carpetaDeGitHub es la de los flujos de GitHub Actions y la configuración
	// de GitHub del repositorio.
	carpetaDeGitHub = ".github"

	// ficheroDelMakefile es la única superficie de invocación del repositorio.
	ficheroDelMakefile = "Makefile"

	// flujoDeEvals es el del job de evals, el último que nombró bin/instalado
	// (contracts/skills-e-invocacion.md §6).
	flujoDeEvals = ".github/workflows/evals.yml"

	// entradaDeScripts es la que una skill ya no lleva (FR-082, ADR 0019).
	entradaDeScripts = "scripts"
)

// restosDeLaInstalacionPorEnlaces son las cadenas que nombran la instalación por
// enlaces que ADR 0019 retira: los enlaces de scripts/ de cada skill y el
// directorio del binario al que apuntaban (FR-083, SC-014).
var restosDeLaInstalacionPorEnlaces = []string{"scripts/boe", "scripts/territorio", "bin/instalado"}

// retiradosConLosEnlaces son los ficheros de la instalación por enlaces que ya
// no existen: el guion de enlaces y lo que los regeneraba en skills-sync, con su
// test (FR-083; contracts/skills-e-invocacion.md §4).
var retiradosConLosEnlaces = []string{
	"scripts/instalar-skills.sh",
	"internal/skills/enlaces.go",
	"internal/skills/enlaces_test.go",
}

// TestSinInstalacionPorEnlaces fija que de la instalación por enlaces no queda
// nada (FR-083, SC-014; contracts/skills-e-invocacion.md §4): ningún fichero de
// skills/, ni el Makefile, ni ninguno de .github/ nombra scripts/boe,
// scripts/territorio ni bin/instalado, que es la búsqueda de quickstart.md §6b;
// no existen el guion de enlaces ni internal/skills/enlaces.go con su test; y
// ninguna skill tiene una entrada scripts, vista sin seguir enlaces.
func TestSinInstalacionPorEnlaces(t *testing.T) {
	t.Parallel()

	raiz := os.DirFS(".")

	casos := []struct {
		nombre string
		probar func(*testing.T, fs.FS)
	}{
		{"nadie-nombra-la-instalacion-por-enlaces", probarQueNadieNombraLaInstalacionPorEnlaces},
		{"retirados", probarRetiradosConLosEnlaces},
		{"ninguna-skill-lleva-scripts", probarNingunaSkillLlevaScripts},
	}

	for _, caso := range casos {
		t.Run(caso.nombre, func(t *testing.T) {
			t.Parallel()

			caso.probar(t, raiz)
		})
	}
}

// probarQueNadieNombraLaInstalacionPorEnlaces busca, línea a línea, cada resto
// de la instalación por enlaces en los ficheros de quickstart.md §6b, y los
// nombra todos como ruta:línea: texto.
func probarQueNadieNombraLaInstalacionPorEnlaces(t *testing.T, raiz fs.FS) {
	t.Helper()

	ficheros := ficherosDeLaBusqueda(t, raiz)

	vigilados := []string{ficheroDelMakefile, flujoDeEvals}
	for _, skill := range skillsDeHoy {
		vigilados = append(vigilados, path.Join(carpetaDeSkills, skill, ficheroDeLaSkill))
	}

	require.Subset(t, slices.Collect(maps.Keys(ficheros)), vigilados,
		"la búsqueda no lee lo que vigila: pasaría en vacío")

	var restos []string

	for _, ruta := range slices.Sorted(maps.Keys(ficheros)) {
		for numero, linea := range strings.Split(string(ficheros[ruta]), "\n") {
			nombraUnResto := slices.ContainsFunc(restosDeLaInstalacionPorEnlaces, func(resto string) bool {
				return strings.Contains(linea, resto)
			})
			if nombraUnResto {
				restos = append(restos, fmt.Sprintf("%s:%d: %s", ruta, numero+1, strings.TrimSpace(linea)))
			}
		}
	}

	assert.Empty(t, restos, "%s nombran la instalación por enlaces (FR-083, SC-014)",
		strings.Join(restosDeLaInstalacionPorEnlaces, ", "))
}

// ficherosDeLaBusqueda son, por su ruta, los ficheros regulares de skills/ y de
// .github/, a cualquier profundidad, y el Makefile, con sus bytes. No sigue
// ningún enlace, como grep -r.
func ficherosDeLaBusqueda(t *testing.T, raiz fs.FS) map[string][]byte {
	t.Helper()

	ficheros := map[string][]byte{}

	for _, carpeta := range []string{carpetaDeSkills, carpetaDeGitHub} {
		err := fs.WalkDir(raiz, carpeta, func(ruta string, entrada fs.DirEntry, err error) error {
			if err != nil || !entrada.Type().IsRegular() {
				return err
			}

			ficheros[ruta], err = fs.ReadFile(raiz, ruta)

			return err
		})
		require.NoError(t, err)
	}

	makefile, err := fs.ReadFile(raiz, ficheroDelMakefile)
	require.NoError(t, err)

	ficheros[ficheroDelMakefile] = makefile

	return ficheros
}

// probarRetiradosConLosEnlaces exige que cada fichero retirado con la
// instalación por enlaces no exista, ni siquiera como enlace colgando.
func probarRetiradosConLosEnlaces(t *testing.T, raiz fs.FS) {
	t.Helper()

	for _, ruta := range retiradosConLosEnlaces {
		_, err := fs.Lstat(raiz, ruta)
		assert.ErrorIsf(t, err, fs.ErrNotExist, "%s sigue existiendo (FR-083)", ruta)
	}
}

// probarNingunaSkillLlevaScripts exige que ningún directorio de skills/ tenga
// una entrada scripts de ningún tipo, vista con Lstat: un enlace colgando
// también cuenta (FR-082, FR-083; ADR 0019).
func probarNingunaSkillLlevaScripts(t *testing.T, raiz fs.FS) {
	t.Helper()

	entradas, err := fs.ReadDir(raiz, carpetaDeSkills)
	require.NoError(t, err)

	var skills []string

	for _, entrada := range entradas {
		if entrada.IsDir() {
			skills = append(skills, entrada.Name())
		}
	}

	require.Subset(t, skills, skillsDeHoy, "skills/ no tiene las skills de hoy: la comprobación pasaría en vacío")

	for _, skill := range skills {
		_, err := fs.Lstat(raiz, path.Join(carpetaDeSkills, skill, entradaDeScripts))
		assert.ErrorIsf(t, err, fs.ErrNotExist, "la skill %s tiene %s: una skill no lleva scripts/ (FR-082, FR-083)",
			skill, entradaDeScripts)
	}
}

const (
	// ficheroDeGoreleaser es la configuración de la release, en la raíz del
	// módulo (contracts/release.md §2).
	ficheroDeGoreleaser = ".goreleaser.yaml"

	// nombreDelProyecto es el del proyecto, el binario, el cask, el manifiesto
	// del bucket y los paquetes.
	nombreDelProyecto = "kitlegal"

	// plantillaDeVersion es la versión que goreleaser inyecta: la etiqueta tal
	// cual en una release (v0.1.0, la misma que da el Makefile en ese commit) y la
	// que asigna al snapshot en un snapshot (FR-092; research.md D21, V6).
	plantillaDeVersion = "{{ if .IsSnapshot }}{{ .Version }}{{ else }}{{ .Tag }}{{ end }}"

	// simboloDeLaVersion y simboloDeLaVersionDeRed son las dos inyecciones que
	// llevan la versión: la que imprime el verbo version y la del User-Agent
	// (FR-091).
	simboloDeLaVersion      = "main.version"
	simboloDeLaVersionDeRed = "github.com/jmorenobl/kitlegal/internal/httpx.version"

	// variableDelPublicador nombra el único secreto de la publicación en el tap
	// y en el bucket, y plantillaDelPublicador es la que lo lee del entorno, que
	// goreleaser solo evalúa al publicar (FR-097; research.md V11, V12).
	variableDelPublicador  = "PUBLISHER_TOKEN"
	plantillaDelPublicador = "{{ .Env.PUBLISHER_TOKEN }}"

	// propietarioEnGitHub es la cuenta de los tres repositorios de la
	// publicación, y paginaDelProyecto, la página que declaran el cask, el bucket
	// y los paquetes (contracts/release.md §2).
	propietarioEnGitHub = "jmorenobl"
	paginaDelProyecto   = "https://github.com/jmorenobl/kitlegal"

	// licenciaDelProyecto es la de LICENSE, con su identificador SPDX.
	licenciaDelProyecto = "EUPL-1.2"

	// mantenedorDeLosPaquetes es el maintainer de los .deb y .rpm: el nombre y
	// la dirección que ya publica el historial del repositorio (research.md D26;
	// supuesto anotado en gates/supuestos.md).
	mantenedorDeLosPaquetes = "Jorge <jmorenobl@gmail.com>"

	// instaladorEnLaRelease es scripts/install.sh como lo nombran
	// checksum.extra_files y release.extra_files: una ruta literal que tiene que
	// existir, porque sin él el snapshot falla (research.md D23, V48).
	instaladorEnLaRelease = "./scripts/install.sh"

	// extensionEnLaRelease y pluginEnLaRelease son las dos piezas de H22 como
	// las nombran checksum.extra_files y release.extra_files: las deja en dist/
	// el paso que empaqueta, antes de que goreleaser calcule las huellas (H22
	// FR-002; contracts/release.md §1 de H22).
	extensionEnLaRelease = "./dist/kitlegal.mcpb"
	pluginEnLaRelease    = "./dist/kitlegal-plugin.zip"

	// idDeLaConstruccion es el de la única construcción, que nombran el
	// universal y los archivos, e idDelUniversal, el del binario universal de
	// macOS: uno propio, para que ningún archivo lo lleve (H22 FR-006; research.md
	// D7 de H22).
	idDeLaConstruccion = "kitlegal"
	idDelUniversal     = "kitlegal-universal"

	// ordenDelPaso es la del gancho del universal, carácter a carácter: el paso
	// que empaqueta, con la versión y la ruta del universal que le da goreleaser,
	// el binario de Windows amd64 de la construcción, el icono versionado y dist/
	// como salida (H22 FR-001, FR-002; contracts/release.md §1 de H22).
	ordenDelPaso = "go run ./cmd/empaquetar piezas -version {{ .Version }} -macos {{ .Path }} " +
		"-windows dist/kitlegal_windows_amd64_v1/kitlegal.exe -icono mcp/icon.png -salida dist"

	// herramientaGoreleaser es la invocación de goreleaser, como la de las demás
	// herramientas: su versión la fija tools/goreleaser/go.mod (FR-096).
	herramientaGoreleaser = "go tool -modfile=tools/goreleaser/go.mod goreleaser"

	// variableDeLaHerramienta y variableDeInyecciones son las variables del
	// Makefile con esa invocación y con las cuatro -X de make build.
	variableDeLaHerramienta = "GORELEASER"
	variableDeInyecciones   = "LDFLAGS"

	// objetivoDeCI es el veredicto del repositorio, y objetivosFalsos, la
	// regla que declara los objetivos que no son ficheros.
	objetivoDeCI    = "ci"
	objetivosFalsos = ".PHONY"
)

// inyeccionesDeLaRelease son las cuatro -X de builds[0].ldflags, de símbolo a
// plantilla: los símbolos del LDFLAGS del Makefile (FR-091), con la misma
// versión en main.version y en la identificación de red (FR-092).
var inyeccionesDeLaRelease = map[string]string{
	simboloDeLaVersion:      plantillaDeVersion,
	"main.commit":           "{{ .FullCommit }}",
	"main.fecha":            "{{ .Date }}",
	simboloDeLaVersionDeRed: plantillaDeVersion,
}

// ficherosExtraDeLaRelease son, en orden, los de checksum.extra_files y los de
// release.extra_files: install.sh y, detrás, las dos piezas de H22. Lo que no
// está en las dos listas no se publica con su huella bajo la firma (H22
// FR-002).
var ficherosExtraDeLaRelease = []ficheroExtra{
	{Glob: instaladorEnLaRelease},
	{Glob: extensionEnLaRelease},
	{Glob: pluginEnLaRelease},
}

// recetasDeLaRelease son, por objetivo, las recetas de contracts/release.md §3,
// línea a línea como las escribe el Makefile, y la de plugin-check, que es de
// H22: solo TestPluginValido, que ejecuta `claude plugin validate` (H22
// FR-065; contracts/release.md §2 de H22).
var recetasDeLaRelease = map[string][]string{
	"goreleaser-check": {"$(GORELEASER) check"},
	"release":          {"$(GORELEASER) release --snapshot --clean --skip=publish,sign,sbom"},
	"plugin-check":     {"go test -count=1 -tags=snapshot -run '^TestPluginValido$$' ."},
	"snapshot-check": {
		"go test -count=1 -tags=snapshot -run '^TestSnapshot$$' .",
		"KITLEGAL_DIST=$(CURDIR)/dist go test -count=1 -run '^TestEntregaDelHito$$/instalador-' ./internal/app/",
	},
	"install": {
		`CGO_ENABLED=0 go install -trimpath -ldflags "$(LDFLAGS)" ./cmd/kitlegal`,
		`"$$(go list -f '{{.Target}}' ./cmd/kitlegal)" skills install -g --host claude`,
	},
}

// controlesDeCI son los prerrequisitos de ci, en orden: goreleaser-check entra
// (FR-094) y ni release ni snapshot-check, que construyen y leen dist/
// (contracts/release.md §3), ni plugin-check, que además necesita Claude Code
// (H22 FR-065).
var controlesDeCI = []string{
	"fmt-check", "lint", "test", "test-integration", "test-tiempos", "vuln", "schema-check", "skills-check", "goreleaser-check",
	"secrets", "mod-verify", "mod-tidy-check",
}

// ayudaDeLosObjetivos son, por objetivo, los fragmentos que su línea de make
// help tiene que llevar para describir lo que hace ahora (FR-121;
// contracts/release.md §3). Desde H22, la de release nombra las dos piezas que
// el snapshot deja además, y la de plugin-check es la de su contrato, entera
// (contracts/release.md §2 de H22).
var ayudaDeLosObjetivos = map[string][]string{
	"goreleaser-check": {"goreleaser check"},
	"install":          {"skills install -g --host claude"},
	"plugin-check": {
		"valida con claude plugin validate el plugin del snapshot y el catálogo de su versión " +
			"(requiere Claude Code y el dist/ de make release; fuera de ci)",
	},
	"release":        {"no publica", path.Base(extensionEnLaRelease), path.Base(pluginEnLaRelease)},
	"skills-sync":    {"scripts/"},
	"snapshot-check": {"instalador-"},
}

// commitsDeEjemplo son mensajes de commit y el grupo de changelog.groups que
// los recoge: feat, fix y el resto, por Conventional Commits (FR-093).
var commitsDeEjemplo = []struct {
	mensaje string
	grupo   int
}{
	{"feat(H19): T023", 0},
	{"feat: la release", 0},
	{"feat(boe)!: otra salida", 0},
	{"fix(httpx): reintentar el 429", 1},
	{"fix: el aviso", 1},
	{"docs: README", 2},
	{"chore(deps): actualizar", 2},
	{"feature: no es feat", 2},
	{"prefix: no es fix", 2},
	{"Merge pull request #43 from jmorenobl/h6", 2},
}

var (
	// lineaDeRegla, lineaDeVariable y lineaDeAyuda reconocen, en una línea
	// lógica del Makefile, una regla con sus prerrequisitos, la definición de
	// una variable y la línea que imprime make help.
	lineaDeRegla    = regexp.MustCompile(`^([A-Za-z0-9_.-]+):(?:\s+(.*))?$`)
	lineaDeVariable = regexp.MustCompile(`^(?:export\s+)?([A-Za-z_][A-Za-z0-9_]*)\s*(?::=|\?=|=)\s*(.*)$`)
	lineaDeAyuda    = regexp.MustCompile(`^## ([^:]+): (.*)$`)

	// inyeccionDelMakefile es cada -X del LDFLAGS del Makefile, e
	// inyeccionDeGoreleaser, una entrada de builds[0].ldflags que es una sola
	// -X; en las dos, el símbolo y su valor.
	inyeccionDelMakefile  = regexp.MustCompile(`-X\s+([^=\s]+)=(\S+)`)
	inyeccionDeGoreleaser = regexp.MustCompile(`^-X ([^=\s]+)=(.+)$`)
)

// configuracionDeGoreleaser es .goreleaser.yaml entero. Se lee de forma
// estricta: una clave que no está aquí —otra sección, una propiedad obsoleta
// como brews o archives.format, o una que cambie lo que fija la tabla de
// contracts/release.md §2, como builds.ignore o archives.wrap_in_directory—
// hace fallar el test. Desde H22 lleva también lo que fija la tabla de su
// contracts/release.md §1: el id de la construcción, el universal de macOS con
// su gancho y los ids de los archivos.
type configuracionDeGoreleaser struct {
	Version           int                  `yaml:"version"`
	ProjectName       string               `yaml:"project_name"`
	Builds            []binarioDeLaRelease `yaml:"builds"`
	UniversalBinaries []binarioUniversal   `yaml:"universal_binaries"`
	Archives          []empaquetado        `yaml:"archives"`
	Checksum          sumasDeLaRelease     `yaml:"checksum"`
	SBOMs             []inventario         `yaml:"sboms"`
	Signs             []firma              `yaml:"signs"`
	HomebrewCasks     []cask               `yaml:"homebrew_casks"`
	Scoops            []manifiestoDeScoop  `yaml:"scoops"`
	NFPMs             []paqueteDelSistema  `yaml:"nfpms"`
	Release           publicacionEnGitHub  `yaml:"release"`
	Changelog         notasDeLaPublicacion `yaml:"changelog"`
}

type binarioDeLaRelease struct {
	ID      string   `yaml:"id"`
	Main    string   `yaml:"main"`
	Binary  string   `yaml:"binary"`
	Env     []string `yaml:"env"`
	Flags   []string `yaml:"flags"`
	Goos    []string `yaml:"goos"`
	Goarch  []string `yaml:"goarch"`
	Ldflags []string `yaml:"ldflags"`
}

// binarioUniversal es una entrada de universal_binaries. Replace es un puntero
// para distinguir el `replace: false` escrito del que falta: los dos valen lo
// mismo para goreleaser, y el contrato lo pide escrito.
type binarioUniversal struct {
	ID      string              `yaml:"id"`
	IDs     []string            `yaml:"ids"`
	Replace *bool               `yaml:"replace"`
	Hooks   ganchosDelUniversal `yaml:"hooks"`
}

type ganchosDelUniversal struct {
	Pre  []ordenDeGancho `yaml:"pre"`
	Post []ordenDeGancho `yaml:"post"`
}

// ordenDeGancho es un gancho de goreleaser escrito con su forma larga: la orden
// y si su salida se enseña en el registro de la release.
type ordenDeGancho struct {
	Cmd    string `yaml:"cmd"`
	Output bool   `yaml:"output"`
}

type empaquetado struct {
	IDs             []string            `yaml:"ids"`
	NameTemplate    string              `yaml:"name_template"`
	Formats         []string            `yaml:"formats"`
	FormatOverrides []formatoPorSistema `yaml:"format_overrides"`
}

type formatoPorSistema struct {
	Goos    string   `yaml:"goos"`
	Formats []string `yaml:"formats"`
}

type ficheroExtra struct {
	Glob string `yaml:"glob"`
}

type sumasDeLaRelease struct {
	NameTemplate string         `yaml:"name_template"`
	ExtraFiles   []ficheroExtra `yaml:"extra_files"`
}

type inventario struct {
	Artifacts string `yaml:"artifacts"`
}

type firma struct {
	Cmd       string   `yaml:"cmd"`
	Artifacts string   `yaml:"artifacts"`
	Signature string   `yaml:"signature"`
	Args      []string `yaml:"args"`
}

type repositorioDePublicacion struct {
	Owner string `yaml:"owner"`
	Name  string `yaml:"name"`
	Token string `yaml:"token"`
}

type cask struct {
	Name        string                   `yaml:"name"`
	Repository  repositorioDePublicacion `yaml:"repository"`
	Homepage    string                   `yaml:"homepage"`
	Description string                   `yaml:"description"`
	Hooks       ganchosDelCask           `yaml:"hooks"`
}

type ganchosDelCask struct {
	Post ganchoDelCask `yaml:"post"`
}

type ganchoDelCask struct {
	Install string `yaml:"install"`
}

type manifiestoDeScoop struct {
	Name        string                   `yaml:"name"`
	Repository  repositorioDePublicacion `yaml:"repository"`
	Homepage    string                   `yaml:"homepage"`
	Description string                   `yaml:"description"`
	License     string                   `yaml:"license"`
}

type paqueteDelSistema struct {
	PackageName string   `yaml:"package_name"`
	Formats     []string `yaml:"formats"`
	Maintainer  string   `yaml:"maintainer"`
	Description string   `yaml:"description"`
	Homepage    string   `yaml:"homepage"`
	License     string   `yaml:"license"`
}

type repositorioDeGitHub struct {
	Owner string `yaml:"owner"`
	Name  string `yaml:"name"`
}

type publicacionEnGitHub struct {
	GitHub     repositorioDeGitHub `yaml:"github"`
	ExtraFiles []ficheroExtra      `yaml:"extra_files"`
}

type notasDeLaPublicacion struct {
	Use    string           `yaml:"use"`
	Sort   string           `yaml:"sort"`
	Groups []grupoDeCambios `yaml:"groups"`
}

type grupoDeCambios struct {
	Title  string `yaml:"title"`
	Regexp string `yaml:"regexp"`
	Order  int    `yaml:"order"`
}

// makefile es lo que el test lee del Makefile: su texto, sus variables, sus
// reglas por objetivo y, por objetivo, la descripción de su línea de ayuda
// («## objetivo: descripción», la que imprime make help).
type makefile struct {
	texto     string
	variables map[string]string
	reglas    map[string]regla
	ayudas    map[string]string
}

// regla son los prerrequisitos de un objetivo y su receta, una entrada por línea
// lógica, sin el tabulador.
type regla struct {
	prerrequisitos []string
	receta         []string
}

// laRelease es lo que TestConfiguracionDeLaRelease comprueba: .goreleaser.yaml
// leído de forma estricta y como documento YAML, el Makefile, ci.yml y
// release.yml leídos de forma estricta fuera de sus trabajos, y cada flujo de
// .github/workflows como documento YAML, por su ruta.
type laRelease struct {
	raiz        fs.FS
	goreleaser  configuracionDeGoreleaser
	documento   *yaml.Node
	makefile    makefile
	ci          flujoDeGitHub
	publicacion flujoDeGitHub
	flujos      map[string]*yaml.Node
}

// TestConfiguracionDeLaRelease fija lo que contracts/release.md §2, §3, §5 y §6
// dicen de la release y se puede comprobar sin construirla ni ejecutar ningún
// flujo (research.md D28): .goreleaser.yaml tiene exactamente las secciones de
// la tabla de §2 con sus valores —seis plataformas sin cgo y con -trimpath, las
// cuatro inyecciones del Makefile con la plantilla de etiqueta y de snapshot,
// archivos sin versión en el nombre, checksums con install.sh, SBOM, firma con
// cosign, cask con el gancho de la cuarentena, bucket, paquetes, release con
// install.sh y notas por Conventional Commits— y nombra PUBLISHER_TOKEN solo en
// los dos token; el Makefile tiene los objetivos de §3 con sus recetas y su
// línea de ayuda, y goreleaser-check en ci sin release ni snapshot-check
// (FR-090 a FR-097, FR-121); release.yml solo se dispara con una etiqueta v*,
// con los permisos de cada trabajo y ninguno del flujo, PUBLISHER_TOKEN solo en
// el entorno del paso que publica, la atestación de los seis archivos y de
// checksums.txt y el humo con sus seis comprobaciones (FR-110 a FR-115); y el
// trabajo snapshot de ci.yml, sin id-token ni ningún secreto, ejecuta make
// release y make snapshot-check (FR-120). Que ninguna propiedad esté obsoleta lo
// comprueba goreleaser check, en make ci (FR-094).
//
// Desde H22 fija además lo que su contracts/release.md §1 y §7 dicen de
// .goreleaser.yaml: el id de la construcción, el universal de macOS —con un id
// propio, sin sustituir a los dos binarios de macOS y con el paso que empaqueta
// como único gancho—, los archivos acotados a la construcción y las dos piezas,
// kitlegal.mcpb y kitlegal-plugin.zip, en checksums.txt y en la release (H22
// FR-002, FR-006, FR-068). Y lo que sus §2 y §5 dicen de la validación del
// plugin: el objetivo plugin-check, con su receta y su línea de ayuda y fuera
// de ci, y los dos pasos que el trabajo snapshot gana para ejecutarlo, con
// Claude Code en la versión que fija el job de evals (H22 FR-065).
func TestConfiguracionDeLaRelease(t *testing.T) {
	t.Parallel()

	release := leerLaRelease(t, os.DirFS("."))

	casos := []struct {
		nombre string
		probar func(*testing.T, laRelease)
	}{
		{"proyecto", probarProyecto},
		{"plataformas", probarPlataformas},
		{"universal", probarUniversal},
		{"inyecciones", probarInyecciones},
		{"archivos", probarArchivos},
		{"checksums", probarChecksums},
		{"sbom-y-firma", probarSBOMYFirma},
		{"cask", probarCask},
		{"bucket", probarBucket},
		{"paquetes", probarPaquetes},
		{"publicacion", probarPublicacion},
		{"notas", probarNotas},
		{"secreto-del-publicador", probarSecretoDelPublicador},
		{"objetivos-del-makefile", probarObjetivosDelMakefile},
		{"ci", probarCI},
		{"ayuda", probarAyuda},
		{"flujo-de-la-release", probarFlujoDeLaRelease},
		{"trabajo-de-publicacion", probarTrabajoDePublicacion},
		{"tokens-de-la-publicacion", probarTokensDeLaPublicacion},
		{"atestacion", probarAtestacion},
		{"humo", probarHumo},
		{"trabajo-de-snapshot", probarTrabajoDeSnapshot},
	}

	for _, caso := range casos {
		t.Run(caso.nombre, func(t *testing.T) {
			t.Parallel()

			caso.probar(t, release)
		})
	}
}

// leerLaRelease lee .goreleaser.yaml de forma estricta y como documento, el
// Makefile, ci.yml y release.yml de forma estricta fuera de sus trabajos, y cada
// flujo como documento.
func leerLaRelease(t *testing.T, raiz fs.FS) laRelease {
	t.Helper()

	contenido, err := fs.ReadFile(raiz, ficheroDeGoreleaser)
	require.NoError(t, err)

	leida := laRelease{
		raiz:        raiz,
		documento:   &yaml.Node{},
		makefile:    leerMakefile(t, raiz),
		ci:          leerFlujo(t, raiz, flujoDeCI),
		publicacion: leerFlujo(t, raiz, flujoDeLaRelease),
		flujos:      leerLosFlujos(t, raiz),
	}

	require.NoError(t, decodificarEstricto(contenido, &leida.goreleaser),
		"%s tiene algo que la tabla de contracts/release.md §2 no fija", ficheroDeGoreleaser)
	require.NoError(t, yaml.Unmarshal(contenido, leida.documento))

	return leida
}

// decodificarEstricto lee el YAML en el destino y falla con una clave que el
// destino no declara.
func decodificarEstricto(contenido []byte, destino any) error {
	estricto := yaml.NewDecoder(bytes.NewReader(contenido))
	estricto.KnownFields(true)

	return estricto.Decode(destino)
}

// leerMakefile lee las variables, las reglas y las líneas de ayuda del
// Makefile, sobre sus líneas lógicas.
func leerMakefile(t *testing.T, raiz fs.FS) makefile {
	t.Helper()

	contenido, err := fs.ReadFile(raiz, ficheroDelMakefile)
	require.NoError(t, err)

	leido := makefile{
		texto:     string(contenido),
		variables: map[string]string{},
		reglas:    map[string]regla{},
		ayudas:    map[string]string{},
	}

	objetivo := ""

	for _, linea := range lineasLogicas(leido.texto) {
		if receta, esReceta := strings.CutPrefix(linea, "\t"); esReceta && objetivo != "" {
			actual := leido.reglas[objetivo]
			actual.receta = append(actual.receta, strings.TrimSpace(receta))
			leido.reglas[objetivo] = actual

			continue
		}

		objetivo = ""

		if partes := lineaDeAyuda.FindStringSubmatch(linea); partes != nil {
			leido.ayudas[partes[1]] = partes[2]
		} else if partes := lineaDeRegla.FindStringSubmatch(linea); partes != nil {
			objetivo = partes[1]
			leido.reglas[objetivo] = regla{prerrequisitos: strings.Fields(partes[2])}
		} else if partes := lineaDeVariable.FindStringSubmatch(linea); partes != nil {
			leido.variables[partes[1]] = partes[2]
		}
	}

	return leido
}

// lineasLogicas son las líneas del texto con cada continuación —una barra
// invertida al final— unida a la siguiente por un espacio y sin la sangría de
// esta, como make lee las variables y los prerrequisitos.
func lineasLogicas(texto string) []string {
	var (
		lineas     []string
		pendiente  string
		continuada bool
	)

	for _, fisica := range strings.Split(texto, "\n") {
		if continuada {
			fisica = pendiente + " " + strings.TrimLeft(fisica, " \t")
		}

		pendiente, continuada = strings.CutSuffix(fisica, `\`)
		if !continuada {
			lineas = append(lineas, fisica)
		}
	}

	return lineas
}

// unica es el único elemento de una sección de lista que contracts/release.md
// §2 fija con una sola entrada.
func unica[T any](t *testing.T, seccion string, elementos []T) T {
	t.Helper()

	require.Lenf(t, elementos, 1, "%s: contracts/release.md §2 fija exactamente una entrada", seccion)

	return elementos[0]
}

func probarProyecto(t *testing.T, release laRelease) {
	t.Helper()

	assert.Equal(t, 2, release.goreleaser.Version)
	assert.Equal(t, nombreDelProyecto, release.goreleaser.ProjectName)
}

// probarPlataformas: kitlegal para darwin, linux y windows en amd64 y arm64,
// sin cgo y con -trimpath (FR-090), con el id que nombran el universal y los
// archivos (H22 FR-006).
func probarPlataformas(t *testing.T, release laRelease) {
	t.Helper()

	construida := unica(t, "builds", release.goreleaser.Builds)

	assert.Equal(t, idDeLaConstruccion, construida.ID,
		"builds[0].id es el que nombran universal_binaries[0].ids y archives[0].ids (H22 FR-006)")
	assert.Equal(t, "./cmd/kitlegal", construida.Main)
	assert.Equal(t, nombreDelProyecto, construida.Binary)
	assert.Equal(t, []string{"CGO_ENABLED=0"}, construida.Env)
	assert.Equal(t, []string{"-trimpath"}, construida.Flags)
	assert.ElementsMatch(t, []string{"darwin", "linux", "windows"}, construida.Goos)
	assert.ElementsMatch(t, []string{"amd64", "arm64"}, construida.Goarch)
}

// probarUniversal: un solo binario universal de macOS, de la construcción
// kitlegal, con un id propio y `replace: false` escrito —con otro id, ningún
// archivo lo lleva; sin sustituir, los dos binarios de macOS siguen en los
// suyos—, sin ningún gancho `pre` y con un solo gancho `post`: el paso que
// empaqueta las dos piezas, con la orden del contrato carácter a carácter y su
// salida en el registro de la release (H22 FR-001, FR-002, FR-006, FR-068;
// contracts/release.md §1 y §7 de H22).
func probarUniversal(t *testing.T, release laRelease) {
	t.Helper()

	universal := unica(t, "universal_binaries", release.goreleaser.UniversalBinaries)

	assert.Equal(t, idDelUniversal, universal.ID,
		"universal_binaries[0].id es uno propio: con el de la construcción, archives lo empaquetaría (H22 FR-006)")
	assert.Equal(t, []string{idDeLaConstruccion}, universal.IDs)

	if assert.NotNil(t, universal.Replace, "universal_binaries[0].replace se escribe (H22 contracts/release.md §1)") {
		assert.False(t, *universal.Replace,
			"universal_binaries[0].replace es falso: los dos binarios de macOS siguen en sus archivos (H22 FR-006)")
	}

	assert.Empty(t, universal.Hooks.Pre, "universal_binaries[0].hooks.pre: ningún gancho antes del universal")
	assert.Equal(t, []ordenDeGancho{{Cmd: ordenDelPaso, Output: true}}, universal.Hooks.Post,
		"universal_binaries[0].hooks.post es un solo gancho, el paso que empaqueta (H22 FR-002)")
}

// probarInyecciones: builds[0].ldflags son exactamente cuatro -X, una por
// entrada, con los símbolos del LDFLAGS del Makefile y la plantilla de etiqueta
// y de snapshot en las dos que llevan la versión, que en el Makefile también
// comparten valor (FR-091, FR-092).
func probarInyecciones(t *testing.T, release laRelease) {
	t.Helper()

	ldflags := unica(t, "builds", release.goreleaser.Builds).Ldflags
	deGoreleaser := map[string]string{}

	for _, ldflag := range ldflags {
		partes := inyeccionDeGoreleaser.FindStringSubmatch(ldflag)
		require.NotNilf(t, partes, "builds[0].ldflags lleva %q, que no es una sola -X (FR-091)", ldflag)

		deGoreleaser[partes[1]] = partes[2]
	}

	assert.Len(t, ldflags, len(inyeccionesDeLaRelease), "builds[0].ldflags no son exactamente cuatro -X (FR-091)")
	assert.Equal(t, inyeccionesDeLaRelease, deGoreleaser, "las -X de builds[0].ldflags (FR-091, FR-092)")

	delMakefile := map[string]string{}
	for _, partes := range inyeccionDelMakefile.FindAllStringSubmatch(release.makefile.variables[variableDeInyecciones], -1) {
		delMakefile[partes[1]] = partes[2]
	}

	assert.ElementsMatch(t, slices.Collect(maps.Keys(delMakefile)), slices.Collect(maps.Keys(deGoreleaser)),
		"las -X de goreleaser no son las del %s del Makefile (FR-091)", variableDeInyecciones)
	assert.Equal(t, delMakefile[simboloDeLaVersion], delMakefile[simboloDeLaVersionDeRed],
		"el Makefile no inyecta la misma versión en %s y en %s (FR-091)", simboloDeLaVersion, simboloDeLaVersionDeRed)
}

// probarArchivos: un archivo por plataforma, tar.gz y zip en Windows, sin
// versión en el nombre y con el binario en la raíz, que es lo que hace
// goreleaser sin wrap_in_directory, que la lectura estricta no admite (FR-093;
// research.md D22). Solo de la construcción kitlegal: sin ids, el universal
// daría un séptimo archivo (H22 FR-006, FR-068).
func probarArchivos(t *testing.T, release laRelease) {
	t.Helper()

	assert.Equal(t, empaquetado{
		IDs:             []string{idDeLaConstruccion},
		NameTemplate:    "{{ .ProjectName }}_{{ .Os }}_{{ .Arch }}",
		Formats:         []string{"tar.gz"},
		FormatOverrides: []formatoPorSistema{{Goos: "windows", Formats: []string{"zip"}}},
	}, unica(t, "archives", release.goreleaser.Archives))
}

// probarChecksums: checksums.txt, con el SHA-256 por omisión, lleva también la
// huella de install.sh, que existe con la ruta que lo nombra (FR-093;
// research.md D23, V48), y, detrás y en ese orden, las de las dos piezas de
// H22 (H22 FR-002, FR-068).
func probarChecksums(t *testing.T, release laRelease) {
	t.Helper()

	assert.Equal(t, sumasDeLaRelease{
		NameTemplate: "checksums.txt",
		ExtraFiles:   ficherosExtraDeLaRelease,
	}, release.goreleaser.Checksum)

	estado, err := fs.Stat(release.raiz, path.Clean(instaladorEnLaRelease))
	require.NoError(t, err, "checksum.extra_files nombra %s, que no existe", instaladorEnLaRelease)
	assert.True(t, estado.Mode().IsRegular(), "%s no es un fichero regular", instaladorEnLaRelease)
}

// probarSBOMYFirma: un SBOM de syft por archivo, con los valores por omisión, y
// la firma keyless de checksums.txt con cosign en un bundle (FR-093;
// research.md V9, V10).
func probarSBOMYFirma(t *testing.T, release laRelease) {
	t.Helper()

	assert.Equal(t, []inventario{{Artifacts: "archive"}}, release.goreleaser.SBOMs)
	assert.Equal(t, []firma{{
		Cmd:       "cosign",
		Artifacts: "checksum",
		Signature: "${artifact}.sigstore.json",
		Args:      []string{"sign-blob", "--bundle=${signature}", "${artifact}", "--yes"},
	}}, release.goreleaser.Signs)
}

// probarCask: el cask kitlegal en jmorenobl/homebrew-tap, con PUBLISHER_TOKEN,
// y el gancho que, solo en macOS, retira la cuarentena del binario instalado
// (FR-093, FR-097; research.md D25).
func probarCask(t *testing.T, release laRelease) {
	t.Helper()

	delTap := unica(t, "homebrew_casks", release.goreleaser.HomebrewCasks)

	assert.Equal(t, nombreDelProyecto, delTap.Name)
	assert.Equal(t, repositorioDePublicacion{
		Owner: propietarioEnGitHub, Name: "homebrew-tap", Token: plantillaDelPublicador,
	}, delTap.Repository)
	assert.Equal(t, paginaDelProyecto, delTap.Homepage)
	assert.NotEmpty(t, delTap.Description)

	for _, fragmento := range []string{"if OS.mac?", `"/usr/bin/xattr"`, `"com.apple.quarantine"`, `"#{staged_path}/kitlegal"`} {
		assert.Containsf(t, delTap.Hooks.Post.Install, fragmento,
			"homebrew_casks[0].hooks.post.install no retira la cuarentena del binario en macOS (research.md D25)")
	}
}

// probarBucket: el manifiesto kitlegal en jmorenobl/scoop-bucket, con
// PUBLISHER_TOKEN (FR-093, FR-097).
func probarBucket(t *testing.T, release laRelease) {
	t.Helper()

	delBucket := unica(t, "scoops", release.goreleaser.Scoops)

	assert.Equal(t, nombreDelProyecto, delBucket.Name)
	assert.Equal(t, repositorioDePublicacion{
		Owner: propietarioEnGitHub, Name: "scoop-bucket", Token: plantillaDelPublicador,
	}, delBucket.Repository)
	assert.Equal(t, paginaDelProyecto, delBucket.Homepage)
	assert.NotEmpty(t, delBucket.Description)
	assert.Equal(t, licenciaDelProyecto, delBucket.License)
}

// probarPaquetes: .deb y .rpm con nfpm, con maintainer, sin el cual goreleaser
// da la configuración por obsoleta (FR-093; research.md D26, V13).
func probarPaquetes(t *testing.T, release laRelease) {
	t.Helper()

	paquete := unica(t, "nfpms", release.goreleaser.NFPMs)

	assert.Equal(t, nombreDelProyecto, paquete.PackageName)
	assert.Equal(t, []string{"deb", "rpm"}, paquete.Formats)
	assert.Equal(t, mantenedorDeLosPaquetes, paquete.Maintainer)
	assert.NotEmpty(t, paquete.Description)
	assert.Equal(t, paginaDelProyecto, paquete.Homepage)
	assert.Equal(t, licenciaDelProyecto, paquete.License)
}

// probarPublicacion: la release en jmorenobl/kitlegal, fijado para que goreleaser
// check no dependa de git ni de un remote, con install.sh adjunto (FR-093,
// FR-109; research.md V2) y, detrás y en ese orden, las dos piezas de H22 (H22
// FR-002, FR-068).
func probarPublicacion(t *testing.T, release laRelease) {
	t.Helper()

	assert.Equal(t, publicacionEnGitHub{
		GitHub:     repositorioDeGitHub{Owner: propietarioEnGitHub, Name: nombreDelProyecto},
		ExtraFiles: ficherosExtraDeLaRelease,
	}, release.goreleaser.Release)
}

// probarNotas: las notas de la release salen de git en orden ascendente y
// agrupadas por Conventional Commits en feat, fix y el resto, que goreleaser
// reparte en el orden de la configuración: cada commit va al primer grupo cuya
// expresión casa con su mensaje, y el grupo sin expresión recoge los demás
// (FR-093; research.md V14).
func probarNotas(t *testing.T, release laRelease) {
	t.Helper()

	notas := release.goreleaser.Changelog

	assert.Equal(t, "git", notas.Use)
	assert.Equal(t, "asc", notas.Sort)
	require.Len(t, notas.Groups, 3, "changelog.groups: feat, fix y el resto")

	assert.Empty(t, notas.Groups[2].Regexp, "el último grupo, sin expresión, recoge el resto")
	assert.Less(t, notas.Groups[0].Order, notas.Groups[1].Order, "los grupos se presentan en el orden feat, fix, resto")
	assert.Less(t, notas.Groups[1].Order, notas.Groups[2].Order, "los grupos se presentan en el orden feat, fix, resto")

	for _, grupo := range notas.Groups {
		assert.NotEmpty(t, grupo.Title)
	}

	for _, commit := range commitsDeEjemplo {
		assert.Equalf(t, commit.grupo, grupoDelCommit(t, notas.Groups, commit.mensaje),
			"el commit %q va a otro grupo de changelog.groups", commit.mensaje)
	}
}

// grupoDelCommit es el índice del grupo al que goreleaser lleva el mensaje: el
// primero, en el orden de la configuración, sin expresión o cuya expresión casa
// con él; -1 si ninguno (research.md V14).
func grupoDelCommit(t *testing.T, grupos []grupoDeCambios, mensaje string) int {
	t.Helper()

	for indice, grupo := range grupos {
		if grupo.Regexp == "" {
			return indice
		}

		expresion, err := regexp.Compile(grupo.Regexp)
		require.NoError(t, err, "la expresión del grupo %q", grupo.Title)

		if expresion.MatchString(mensaje) {
			return indice
		}
	}

	return -1
}

// probarSecretoDelPublicador: .goreleaser.yaml nombra PUBLISHER_TOKEN
// exactamente en los dos token, como clave o como valor, y el Makefile no lo
// nombra: ni make ci ni make release lo necesitan (FR-097).
func probarSecretoDelPublicador(t *testing.T, release laRelease) {
	t.Helper()

	assert.Equal(t, []string{"homebrew_casks[0].repository.token", "scoops[0].repository.token"},
		rutasQueNombran(release.documento, variableDelPublicador, ""),
		"%s solo se nombra en los token del tap y del bucket (FR-097)", variableDelPublicador)
	assert.NotContains(t, release.makefile.texto, variableDelPublicador,
		"ningún objetivo del Makefile necesita %s (FR-097)", variableDelPublicador)
}

// rutasQueNombran son, en el orden del documento, las rutas —claves separadas
// por punto e índices entre corchetes— de cada clave o valor escalar que
// contiene el texto.
func rutasQueNombran(nodo *yaml.Node, texto, ruta string) []string {
	var rutas []string

	switch nodo.Kind {
	case yaml.DocumentNode:
		for _, hijo := range nodo.Content {
			rutas = append(rutas, rutasQueNombran(hijo, texto, ruta)...)
		}
	case yaml.SequenceNode:
		for indice, hijo := range nodo.Content {
			rutas = append(rutas, rutasQueNombran(hijo, texto, fmt.Sprintf("%s[%d]", ruta, indice))...)
		}
	case yaml.MappingNode:
		for indice := 0; indice+1 < len(nodo.Content); indice += 2 {
			clave := nodo.Content[indice].Value
			deLaClave := strings.TrimPrefix(ruta+"."+clave, ".")

			if strings.Contains(clave, texto) {
				rutas = append(rutas, deLaClave)
			}

			rutas = append(rutas, rutasQueNombran(nodo.Content[indice+1], texto, deLaClave)...)
		}
	case yaml.ScalarNode:
		if strings.Contains(nodo.Value, texto) {
			rutas = append(rutas, ruta)
		}
	case yaml.AliasNode:
		rutas = append(rutas, rutasQueNombran(nodo.Alias, texto, ruta)...)
	}

	return rutas
}

// probarObjetivosDelMakefile: GORELEASER invoca el goreleaser de tools/, y
// goreleaser-check, release, snapshot-check e install tienen las recetas de
// contracts/release.md §3 y están en .PHONY (FR-094, FR-095, FR-096, FR-120,
// FR-125); y plugin-check, la de contracts/release.md §2 de H22, también en
// .PHONY (H22 FR-065).
func probarObjetivosDelMakefile(t *testing.T, release laRelease) {
	t.Helper()

	assert.Equal(t, herramientaGoreleaser, release.makefile.variables[variableDeLaHerramienta])

	for _, objetivo := range slices.Sorted(maps.Keys(recetasDeLaRelease)) {
		actual, existe := release.makefile.reglas[objetivo]
		if assert.Truef(t, existe, "el Makefile no tiene el objetivo %s", objetivo) {
			assert.Equalf(t, recetasDeLaRelease[objetivo], actual.receta, "la receta de %s (contracts/release.md §3)", objetivo)
		}
	}

	assert.Subset(t, release.makefile.reglas[objetivosFalsos].prerrequisitos,
		slices.Collect(maps.Keys(recetasDeLaRelease)), "%s no declara los objetivos de la release", objetivosFalsos)
}

// probarCI: ci ejecuta goreleaser-check con los demás controles, y no release
// ni snapshot-check (FR-094; contracts/release.md §3) ni plugin-check, que
// depende de Claude Code (H22 FR-065).
func probarCI(t *testing.T, release laRelease) {
	t.Helper()

	actual, existe := release.makefile.reglas[objetivoDeCI]
	require.True(t, existe, "el Makefile no tiene el objetivo %s", objetivoDeCI)

	assert.Equal(t, controlesDeCI, actual.prerrequisitos,
		"los prerrequisitos de %s: ni release, ni snapshot-check ni plugin-check, que leen dist/ (FR-094; H22 FR-065)",
		objetivoDeCI)
}

// probarAyuda: make help describe cada objetivo de la release, install y
// skills-sync con lo que hacen ahora (FR-121), y plugin-check con la línea de
// su contrato (H22 FR-065).
func probarAyuda(t *testing.T, release laRelease) {
	t.Helper()

	for _, objetivo := range slices.Sorted(maps.Keys(ayudaDeLosObjetivos)) {
		for _, fragmento := range ayudaDeLosObjetivos[objetivo] {
			assert.Containsf(t, release.makefile.ayudas[objetivo], fragmento,
				"la línea de make help de %s no describe lo que hace (FR-121; H22 contracts/release.md §2)", objetivo)
		}

		assert.Containsf(t, release.makefile.reglas, objetivo, "make help describe %s, que no es un objetivo", objetivo)
	}
}

const (
	// carpetaDeFlujos es la de los flujos de GitHub Actions; flujoDeCI, el de la
	// integración continua, con el trabajo snapshot (contracts/release.md §5), y
	// flujoDeLaRelease, el que publica una etiqueta v* (§6).
	carpetaDeFlujos  = ".github/workflows"
	flujoDeCI        = ".github/workflows/ci.yml"
	flujoDeLaRelease = ".github/workflows/release.yml"

	// trabajoDeCI es el de make ci, cuya preparación de Go repite el trabajo
	// snapshot; trabajoDePublicacion y trabajoDeHumo son los de release.yml.
	trabajoDeCI          = "ci"
	trabajoDeSnapshot    = "snapshot"
	trabajoDePublicacion = "publicar"
	trabajoDeHumo        = "humo"

	// runnerDeLosFlujos es el de los tres trabajos: linux/amd64, el binario que
	// ejecutan TestSnapshot y el humo (research.md S15).
	runnerDeLosFlujos = "ubuntu-latest"

	// accionDeCheckout, accionDeSetupGo, accionDeSyft, accionDeCosign y
	// accionDeAtestacion son las acciones de los trabajos, sin su etiqueta.
	accionDeCheckout   = "actions/checkout"
	accionDeSetupGo    = "actions/setup-go"
	accionDeSyft       = "anchore/sbom-action/download-syft"
	accionDeCosign     = "sigstore/cosign-installer"
	accionDeAtestacion = "actions/attest-build-provenance"

	// ordenDePublicacion es la del paso que publica: el goreleaser de tools/, el
	// mismo que el de make release, sin --snapshot (FR-096; contracts/release.md
	// §6).
	ordenDePublicacion = herramientaGoreleaser + " release --clean"

	// ordenDelSnapshot y ordenDeSuComprobacion son las dos órdenes del trabajo
	// snapshot: construirlo y comprobarlo (FR-120).
	ordenDelSnapshot      = "make release"
	ordenDeSuComprobacion = "make snapshot-check"

	// trabajoDeEvals es el trabajo de evals.yml que fija, en su entorno y con
	// variableDeClaudeCode, la versión de Claude Code de las sesiones de las
	// evals (ADR 0031).
	trabajoDeEvals       = "evals"
	variableDeClaudeCode = "VERSION_DE_CLAUDE_CODE"

	// ordenDeClaudeCode y ordenDeLaValidacion son las dos órdenes que H22 añade
	// al trabajo snapshot, detrás de las de H19: instalar Claude Code en la
	// versión de variableDeClaudeCode y validar con él el plugin del snapshot y
	// el catálogo de su versión (H22 FR-065; contracts/release.md §5 de H22).
	ordenDeClaudeCode   = `npm install -g "@anthropic-ai/claude-code@${` + variableDeClaudeCode + `}"`
	ordenDeLaValidacion = "make plugin-check"

	// lecturaDeSecretos es el contexto del que un flujo lee un secreto, y
	// accesoDelFlujo, el token que la plataforma da a cada ejecución con los
	// permisos del trabajo; los dos se leen con expresion.
	lecturaDeSecretos = "secrets"
	accesoDelFlujo    = "github.token"

	// variableDeLaRelease es la del token con el que goreleaser publica la
	// release en GitHub.
	variableDeLaRelease = "GITHUB_TOKEN"

	// repositorioDelProyecto es el de la release, contra el que el humo la
	// descarga y verifica su atestación.
	repositorioDelProyecto = propietarioEnGitHub + "/" + nombreDelProyecto

	// carpetaDeLaRelease es donde goreleaser deja lo que construye, y
	// checksumsDeLaRelease, el fichero de checksums que publica
	// (contracts/release.md §2).
	carpetaDeLaRelease   = "dist"
	checksumsDeLaRelease = "checksums.txt"

	// archivoDelHumo es el archivo publicado para el runner del humo.
	archivoDelHumo = "kitlegal_linux_amd64.tar.gz"

	// shellDelHumo, directorioDelHumo y opcionesDelHumo son el shell de cada paso
	// del humo, el directorio temporal del runner en el que se ejecuta, que se lee
	// con expresion, y su primera orden: una orden que falla, una variable sin
	// valor o una tubería rota detienen el paso.
	shellDelHumo      = "bash"
	directorioDelHumo = "runner.temp"
	opcionesDelHumo   = "set -euo pipefail"
)

// archivosDeLaRelease son los seis archivos que la release publica y atesta
// (FR-112): los mismos que TestSnapshot exige en el snapshot, que solo se
// compila con la etiqueta snapshot (SC-016; research.md D22).
var archivosDeLaRelease = []string{
	"kitlegal_darwin_amd64.tar.gz",
	"kitlegal_darwin_arm64.tar.gz",
	"kitlegal_linux_amd64.tar.gz",
	"kitlegal_linux_arm64.tar.gz",
	"kitlegal_windows_amd64.zip",
	"kitlegal_windows_arm64.zip",
}

// comprobacionesDelHumo son, en orden y un paso cada una, las seis
// comprobaciones del trabajo humo, con las órdenes que el guion de su paso
// tiene que llevar, cada una como una línea entera: una orden comentada,
// precedida de un echo o seguida de un || true ya no comprueba nada
// (contracts/release.md §6; FR-113).
var comprobacionesDelHumo = []struct {
	nombre  string
	ordenes []string
}{
	{"descarga", []string{
		`gh release download "$GITHUB_REF_NAME" --repo ` + repositorioDelProyecto +
			" --pattern " + archivoDelHumo + " --pattern " + checksumsDeLaRelease,
	}},
	{"huella", []string{
		`awk '$2 == "` + archivoDelHumo + `"' ` + checksumsDeLaRelease + " > huella.txt",
		"sha256sum --check --strict huella.txt",
	}},
	{"atestacion", []string{
		"gh attestation verify " + archivoDelHumo + " --repo " + repositorioDelProyecto,
	}},
	{"version", []string{
		"tar -xzf " + archivoDelHumo + " " + nombreDelProyecto,
		"salida=$(./kitlegal version)",
		`primera=$(head -n 1 <<< "$salida")`,
		`if [ "$primera" != "kitlegal $GITHUB_REF_NAME" ]; then`,
		"exit 1",
	}},
	{"boe-sin-red", []string{
		"cache=$(mktemp -d)",
		`KITLEGAL_CACHE_DIR="$cache" ./kitlegal boe articulo BOE-A-2015-10565 a21 --offline || codigo=$?`,
		`if [ "$codigo" -ne 4 ]; then`,
		"exit 1",
	}},
	{"skills-install", []string{
		"proyecto=$(mktemp -d)",
		`cd "$proyecto"`,
		`"$RUNNER_TEMP/kitlegal" skills install`,
		"if [ ! -f .agents/skills/boe-legislacion/SKILL.md ]; then",
		"exit 1",
	}},
}

// etiquetaDeAccion es como los flujos fijan una acción: por su etiqueta mayor,
// como el resto de flujos del repositorio (research.md V36), o por su versión
// completa cuando la acción no publica etiqueta mayor. sigstore/cosign-installer
// solo publica vX.Y.Z: con `@v4` la release v0.1.0 no pudo ni resolver la acción.
var etiquetaDeAccion = regexp.MustCompile(`^v[0-9]+(\.[0-9]+\.[0-9]+)?$`)

// flujoDeGitHub es un flujo de GitHub Actions leído de forma estricta fuera de
// sus trabajos: otra clave del flujo —un env, un concurrency, unos defaults—
// hace fallar el test. Cada trabajo lo lee después la comprobación que lo mira;
// va como valor y no como puntero, porque yaml.v3 solo deja sin leer un
// yaml.Node.
type flujoDeGitHub struct {
	ruta        string
	Name        string                     `yaml:"name"`
	On          map[string]*eventoDelFlujo `yaml:"on"`
	Permissions map[string]string          `yaml:"permissions"`
	Jobs        map[string]yaml.Node       `yaml:"jobs"`
}

// eventoDelFlujo es el filtro de un evento: sin ninguno, el evento no tiene
// valor.
type eventoDelFlujo struct {
	Branches []string `yaml:"branches"`
	Tags     []string `yaml:"tags"`
}

// trabajoDelFlujo es un trabajo con las claves que usan los de
// contracts/release.md §5 y §6. Leído de forma estricta, un if, un
// continue-on-error o un timeout que el contrato no fija hacen fallar el test.
type trabajoDelFlujo struct {
	Needs       []string          `yaml:"needs"`
	RunsOn      string            `yaml:"runs-on"`
	Permissions map[string]string `yaml:"permissions"`
	Env         map[string]string `yaml:"env"`
	Steps       []pasoDelFlujo    `yaml:"steps"`
}

type pasoDelFlujo struct {
	Name             string            `yaml:"name"`
	Uses             string            `yaml:"uses"`
	With             map[string]string `yaml:"with"`
	Env              map[string]string `yaml:"env"`
	Shell            string            `yaml:"shell"`
	WorkingDirectory string            `yaml:"working-directory"`
	Run              string            `yaml:"run"`
}

// leerFlujo lee un flujo de forma estricta fuera de sus trabajos.
func leerFlujo(t *testing.T, raiz fs.FS, ruta string) flujoDeGitHub {
	t.Helper()

	contenido, err := fs.ReadFile(raiz, ruta)
	require.NoError(t, err)

	leido := flujoDeGitHub{ruta: ruta}
	require.NoErrorf(t, decodificarEstricto(contenido, &leido),
		"%s tiene, fuera de sus trabajos, algo que contracts/release.md §5 y §6 no fijan", ruta)

	return leido
}

// leerLosFlujos lee cada flujo de .github/workflows como documento, por su ruta.
func leerLosFlujos(t *testing.T, raiz fs.FS) map[string]*yaml.Node {
	t.Helper()

	flujos := map[string]*yaml.Node{}

	for _, patron := range []string{"*.yml", "*.yaml"} {
		rutas, err := fs.Glob(raiz, path.Join(carpetaDeFlujos, patron))
		require.NoError(t, err)

		for _, ruta := range rutas {
			contenido, err := fs.ReadFile(raiz, ruta)
			require.NoError(t, err)

			flujos[ruta] = &yaml.Node{}
			require.NoError(t, yaml.Unmarshal(contenido, flujos[ruta]), ruta)
		}
	}

	require.Subset(t, slices.Collect(maps.Keys(flujos)), []string{flujoDeCI, flujoDeLaRelease},
		"la búsqueda no lee los flujos que vigila: pasaría en vacío")

	return flujos
}

// leerTrabajo lee de forma estricta un trabajo del flujo.
func leerTrabajo(t *testing.T, flujo flujoDeGitHub, nombre string) trabajoDelFlujo {
	t.Helper()

	nodo, existe := flujo.Jobs[nombre]
	require.Truef(t, existe, "%s no tiene el trabajo %s", flujo.ruta, nombre)

	contenido, err := yaml.Marshal(&nodo)
	require.NoError(t, err)

	var leido trabajoDelFlujo
	require.NoErrorf(t, decodificarEstricto(contenido, &leido),
		"el trabajo %s de %s tiene algo que contracts/release.md §5 y §6 no fijan", nombre, flujo.ruta)

	return leido
}

// secuencia es, paso a paso, lo que hace cada paso: la acción que usa, sin su
// etiqueta, o la orden que ejecuta. Exige que cada acción vaya fijada por su
// etiqueta mayor o, si no la publica, por su versión completa (research.md V36).
func secuencia(t *testing.T, pasos []pasoDelFlujo) []string {
	t.Helper()

	hechos := make([]string, 0, len(pasos))

	for _, paso := range pasos {
		if paso.Uses == "" {
			hechos = append(hechos, strings.TrimSpace(paso.Run))

			continue
		}

		accion, etiqueta, _ := strings.Cut(paso.Uses, "@")
		assert.Regexpf(t, etiquetaDeAccion, etiqueta, "%s no va fijada por su etiqueta mayor ni por su versión completa (research.md V36)", paso.Uses)

		hechos = append(hechos, accion)
	}

	return hechos
}

// pasosEnOrden exige que los pasos hagan exactamente lo esperado y en ese orden,
// y los da por lo que hacen (secuencia).
func pasosEnOrden(t *testing.T, pasos []pasoDelFlujo, esperados ...string) map[string]pasoDelFlujo {
	t.Helper()

	hechos := secuencia(t, pasos)
	require.Equal(t, esperados, hechos, "los pasos del trabajo y su orden (contracts/release.md §5 y §6)")

	porLoQueHacen := map[string]pasoDelFlujo{}
	for indice, hecho := range hechos {
		porLoQueHacen[hecho] = pasos[indice]
	}

	return porLoQueHacen
}

// indiceDelPaso es la posición del paso que hace lo pedido (secuencia).
func indiceDelPaso(t *testing.T, trabajo trabajoDelFlujo, hecho string) int {
	t.Helper()

	indice := slices.Index(secuencia(t, trabajo.Steps), hecho)
	require.GreaterOrEqualf(t, indice, 0, "ningún paso hace %s", hecho)

	return indice
}

// rutasEnLosFlujos son, flujo a flujo en orden de ruta, las rutas que nombran el
// texto (rutasQueNombran), una vez cada una y tras la del flujo.
func rutasEnLosFlujos(flujos map[string]*yaml.Node, texto string) []string {
	var rutas []string

	for _, fichero := range slices.Sorted(maps.Keys(flujos)) {
		for _, ruta := range slices.Compact(rutasQueNombran(flujos[fichero], texto, "")) {
			rutas = append(rutas, fichero+": "+ruta)
		}
	}

	return rutas
}

// ordenesDelGuion son las líneas del guion de un paso sin su sangría, sin las
// vacías y sin los comentarios, que no comprueban nada.
func ordenesDelGuion(guion string) []string {
	var ordenes []string

	for _, linea := range strings.Split(guion, "\n") {
		linea = strings.TrimSpace(linea)
		if linea != "" && !strings.HasPrefix(linea, "#") {
			ordenes = append(ordenes, linea)
		}
	}

	return ordenes
}

// probarFlujoDeLaRelease: release.yml solo se dispara al empujar una etiqueta
// v*, con ningún otro evento (FR-110, FR-115); no da permisos a nivel de flujo,
// de modo que cada trabajo tiene solo los que declara (FR-111); y sus trabajos
// son publicar y humo.
func probarFlujoDeLaRelease(t *testing.T, release laRelease) {
	t.Helper()

	flujo := release.publicacion

	assert.Equal(t, map[string]*eventoDelFlujo{"push": {Tags: []string{"v*"}}}, flujo.On,
		"%s se dispara solo al empujar una etiqueta v* (FR-110)", flujo.ruta)
	assert.Nil(t, flujo.Permissions, "%s no da permisos a nivel de flujo (contracts/release.md §6)", flujo.ruta)
	assert.ElementsMatch(t, []string{trabajoDePublicacion, trabajoDeHumo}, slices.Collect(maps.Keys(flujo.Jobs)))
}

// probarTrabajoDePublicacion: publicar corre en ubuntu-latest con contents,
// id-token y attestations de escritura y sin entorno propio (FR-111, FR-114);
// obtiene el código con toda su historia sin dejar la credencial en el clon,
// instala Go sin restaurar ninguna caché, syft y cosign, publica con el
// goreleaser de tools/ con GITHUB_TOKEN y PUBLISHER_TOKEN en el entorno de ese
// paso, y atesta, en ese orden (contracts/release.md §6).
func probarTrabajoDePublicacion(t *testing.T, release laRelease) {
	t.Helper()

	publicar := leerTrabajo(t, release.publicacion, trabajoDePublicacion)

	assert.Equal(t, runnerDeLosFlujos, publicar.RunsOn)
	assert.Empty(t, publicar.Needs)
	assert.Equal(t, map[string]string{"contents": "write", "id-token": "write", "attestations": "write"},
		publicar.Permissions, "los permisos de %s (FR-111)", trabajoDePublicacion)
	assert.Empty(t, publicar.Env, "%s no tiene entorno propio: cada secreto va en el paso que publica (FR-114)",
		trabajoDePublicacion)

	pasos := pasosEnOrden(t, publicar.Steps,
		accionDeCheckout, accionDeSetupGo, accionDeSyft, accionDeCosign, ordenDePublicacion, accionDeAtestacion)

	assert.Equal(t, map[string]string{"fetch-depth": "0", "persist-credentials": "false"}, pasos[accionDeCheckout].With,
		"el código con toda su historia, para las notas de la release, y sin la credencial en el clon")
	assert.Equal(t, map[string]string{"go-version-file": "go.mod", "cache": "false"}, pasos[accionDeSetupGo].With,
		"Go de go.mod, sin restaurar nada de otra ejecución en lo que se publica y se firma")
	assert.Equal(t, map[string]string{
		variableDeLaRelease:   expresion(lecturaDeSecretos + "." + variableDeLaRelease),
		variableDelPublicador: expresion(lecturaDeSecretos + "." + variableDelPublicador),
	}, pasos[ordenDePublicacion].Env, "el entorno del paso que publica (FR-097, FR-114)")
}

// probarTokensDeLaPublicacion: de todos los flujos, solo el paso que publica
// nombra PUBLISHER_TOKEN (FR-097, FR-114), y release.yml no lee ningún secreto
// fuera del entorno de ese paso.
func probarTokensDeLaPublicacion(t *testing.T, release laRelease) {
	t.Helper()

	publicar := leerTrabajo(t, release.publicacion, trabajoDePublicacion)
	entorno := fmt.Sprintf("jobs.%s.steps[%d].env.", trabajoDePublicacion,
		indiceDelPaso(t, publicar, ordenDePublicacion))

	assert.Equal(t, []string{flujoDeLaRelease + ": " + entorno + variableDelPublicador},
		rutasEnLosFlujos(release.flujos, variableDelPublicador),
		"%s solo se usa en el paso que publica (FR-097, FR-114)", variableDelPublicador)
	assert.ElementsMatch(t, []string{entorno + variableDeLaRelease, entorno + variableDelPublicador},
		rutasQueNombran(release.flujos[flujoDeLaRelease], lecturaDeSecretos, ""),
		"%s solo lee un secreto en el entorno del paso que publica (FR-114)", flujoDeLaRelease)
}

// expresion es como un flujo lee un valor de un contexto de GitHub Actions:
// expresion("github.token") es ${{ github.token }}.
func expresion(valor string) string {
	return "${{ " + valor + " }}"
}

// probarAtestacion: tras publicar, attest-build-provenance atesta la procedencia
// de exactamente los seis archivos y checksums.txt de dist/, y ningún otro flujo
// la usa (FR-112).
func probarAtestacion(t *testing.T, release laRelease) {
	t.Helper()

	publicar := leerTrabajo(t, release.publicacion, trabajoDePublicacion)
	indice := indiceDelPaso(t, publicar, accionDeAtestacion)

	assert.Greater(t, indice, indiceDelPaso(t, publicar, ordenDePublicacion), "la atestación va tras publicar")

	var sujetos []string
	for _, fichero := range append(slices.Clone(archivosDeLaRelease), checksumsDeLaRelease) {
		sujetos = append(sujetos, path.Join(carpetaDeLaRelease, fichero))
	}

	atestacion := publicar.Steps[indice]

	assert.Equal(t, []string{"subject-path"}, slices.Collect(maps.Keys(atestacion.With)))
	assert.ElementsMatch(t, sujetos, strings.Fields(atestacion.With["subject-path"]),
		"los seis archivos y %s (FR-112)", checksumsDeLaRelease)
	assert.Equal(t, []string{fmt.Sprintf("%s: jobs.%s.steps[%d].uses", flujoDeLaRelease, trabajoDePublicacion, indice)},
		rutasEnLosFlujos(release.flujos, accionDeAtestacion), "solo release.yml atesta (FR-112)")
}

// probarHumo: humo espera a publicar y corre en ubuntu-latest, con permiso para
// leer el contenido y las atestaciones y el token del flujo para gh; hace las
// seis comprobaciones de contracts/release.md §6, un paso cada una y en orden,
// ninguno con una acción, de modo que no obtiene el código del repositorio
// (FR-113).
func probarHumo(t *testing.T, release laRelease) {
	t.Helper()

	humo := leerTrabajo(t, release.publicacion, trabajoDeHumo)

	assert.Equal(t, []string{trabajoDePublicacion}, humo.Needs)
	assert.Equal(t, runnerDeLosFlujos, humo.RunsOn)
	assert.Equal(t, map[string]string{"contents": "read", "attestations": "read"}, humo.Permissions,
		"los permisos de %s (contracts/release.md §6)", trabajoDeHumo)
	assert.Equal(t, map[string]string{"GH_TOKEN": expresion(accesoDelFlujo)}, humo.Env,
		"gh usa el token de la ejecución, con los permisos de %s", trabajoDeHumo)
	require.Len(t, humo.Steps, len(comprobacionesDelHumo), "un paso por comprobación del humo (FR-113)")

	for indice, comprobacion := range comprobacionesDelHumo {
		probarPasoDelHumo(t, humo.Steps[indice], comprobacion.nombre, comprobacion.ordenes)
	}
}

// probarPasoDelHumo: el paso de una comprobación no usa ninguna acción, corre con
// bash en el directorio temporal del runner, empieza fijando sus opciones de
// shell, lleva cada orden de la comprobación como una línea entera fuera de sus
// comentarios y bash lo lee sin errores de sintaxis, sin ejecutarlo.
func probarPasoDelHumo(t *testing.T, paso pasoDelFlujo, comprobacion string, esperadas []string) {
	t.Helper()

	assert.Emptyf(t, paso.Uses, "el paso %q del humo usa una acción", comprobacion)
	assert.Equalf(t, shellDelHumo, paso.Shell, "el shell del paso %q del humo", comprobacion)
	assert.Equalf(t, expresion(directorioDelHumo), paso.WorkingDirectory, "el directorio del paso %q del humo",
		comprobacion)

	ordenes := ordenesDelGuion(paso.Run)
	require.NotEmptyf(t, ordenes, "el paso %q del humo no ejecuta nada", comprobacion)
	assert.Equalf(t, opcionesDelHumo, ordenes[0], "el paso %q del humo no empieza fijando sus opciones", comprobacion)

	for _, orden := range esperadas {
		assert.Containsf(t, ordenes, orden, "el paso %q del humo no hace lo que fija contracts/release.md §6 (FR-113)",
			comprobacion)
	}

	sintaxis := exec.CommandContext(t.Context(), shellDelHumo, "-n")
	sintaxis.Stdin = strings.NewReader(paso.Run)

	salida, err := sintaxis.CombinedOutput()
	assert.NoErrorf(t, err, "bash no lee el paso %q del humo: %s", comprobacion, salida)
}

// probarTrabajoDeSnapshot: ci.yml se dispara en cada propuesta de cambio y en
// cada push a main, sin permisos a nivel de flujo; su trabajo snapshot corre en
// esos eventos sin condición ni espera, en ubuntu-latest, solo con permiso para
// leer el contenido —sin id-token— y sin ningún secreto ni el token del flujo;
// obtiene el código con toda su historia, prepara Go y la caché como el trabajo
// ci, y ejecuta make release y make snapshot-check (FR-120;
// contracts/release.md §5). Desde H22 son seis pasos: detrás instala Claude
// Code, en la versión que fija el job de evals, y ejecuta make plugin-check, y
// nada más (H22 FR-065; contracts/release.md §5 y §7 de H22).
func probarTrabajoDeSnapshot(t *testing.T, release laRelease) {
	t.Helper()

	assert.Equal(t, map[string]*eventoDelFlujo{"pull_request": nil, "push": {Branches: []string{"main"}}},
		release.ci.On, "los eventos de %s (contracts/release.md §5)", flujoDeCI)
	assert.Nil(t, release.ci.Permissions, "%s no da permisos a nivel de flujo", flujoDeCI)

	snapshot := leerTrabajo(t, release.ci, trabajoDeSnapshot)

	assert.Equal(t, runnerDeLosFlujos, snapshot.RunsOn)
	assert.Empty(t, snapshot.Needs, "%s corre junto a %s, sin esperarlo", trabajoDeSnapshot, trabajoDeCI)
	assert.Equal(t, map[string]string{"contents": "read"}, snapshot.Permissions, "sin id-token (FR-120)")

	nodo := release.ci.Jobs[trabajoDeSnapshot]
	for _, lectura := range []string{lecturaDeSecretos, accesoDelFlujo} {
		assert.Emptyf(t, rutasQueNombran(&nodo, lectura, "jobs."+trabajoDeSnapshot),
			"%s no usa ningún secreto (FR-097, FR-120)", trabajoDeSnapshot)
	}

	pasos := pasosEnOrden(t, snapshot.Steps, accionDeCheckout, accionDeSetupGo, ordenDelSnapshot, ordenDeSuComprobacion,
		ordenDeClaudeCode, ordenDeLaValidacion)

	assert.Equal(t, map[string]string{"fetch-depth": "0"}, pasos[accionDeCheckout].With,
		"el código con toda su historia, de la que sale la versión del snapshot (research.md D30)")
	assert.Equalf(t, map[string]string{variableDeClaudeCode: versionDeClaudeCodeDeLasEvals(t, release)},
		pasos[ordenDeClaudeCode].Env,
		"el entorno del paso que instala Claude Code: solo %s, con la versión de jobs.%s.env de %s (H22 FR-065)",
		variableDeClaudeCode, trabajoDeEvals, flujoDeEvals)

	// El trabajo ci se lee tal cual, sin exigirle nada más: lo fija su propio
	// contrato, y aquí solo importa su preparación de Go.
	var delCI trabajoDelFlujo

	nodoDeCI := release.ci.Jobs[trabajoDeCI]
	require.NoError(t, nodoDeCI.Decode(&delCI))

	indice := slices.IndexFunc(delCI.Steps, func(paso pasoDelFlujo) bool {
		return strings.HasPrefix(paso.Uses, accionDeSetupGo+"@")
	})
	require.GreaterOrEqualf(t, indice, 0, "el trabajo %s no usa %s", trabajoDeCI, accionDeSetupGo)

	assert.Equal(t, delCI.Steps[indice].Uses, pasos[accionDeSetupGo].Uses)
	assert.Equal(t, delCI.Steps[indice].With, pasos[accionDeSetupGo].With,
		"la preparación de Go y la caché del trabajo %s (contracts/release.md §5)", trabajoDeCI)
}

// entornoDeLosTrabajos es lo único que el test lee de un flujo que no fija: el
// entorno de cada uno de sus trabajos.
type entornoDeLosTrabajos struct {
	Jobs map[string]struct {
		Env map[string]string `yaml:"env"`
	} `yaml:"jobs"`
}

// versionDeClaudeCodeDeLasEvals es la versión de Claude Code que fija el job de
// evals, en jobs.evals.env de evals.yml: la única escrita en el repositorio, y
// la que el trabajo snapshot tiene que repetir (H22 FR-065; research.md V26 de
// H22). Tiene que estar: sin ella, compararla con la del trabajo snapshot
// pasaría en vacío.
func versionDeClaudeCodeDeLasEvals(t *testing.T, release laRelease) string {
	t.Helper()

	documento, leido := release.flujos[flujoDeEvals]
	require.Truef(t, leido, "falta %s, que fija la versión de Claude Code", flujoDeEvals)

	var evals entornoDeLosTrabajos

	require.NoErrorf(t, documento.Decode(&evals), "el entorno de los trabajos de %s no se puede leer", flujoDeEvals)

	version := evals.Jobs[trabajoDeEvals].Env[variableDeClaudeCode]
	require.NotEmptyf(t, version, "jobs.%s.env de %s no fija %s: la comparación pasaría en vacío",
		trabajoDeEvals, flujoDeEvals, variableDeClaudeCode)

	return version
}
