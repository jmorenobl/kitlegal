package kitlegal_test

import (
	"bytes"
	"fmt"
	"io/fs"
	"maps"
	"os"
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
	licenciaDelProyecto = "Apache-2.0"

	// mantenedorDeLosPaquetes es el maintainer de los .deb y .rpm: el nombre y
	// la dirección que ya publica el historial del repositorio (research.md D26;
	// supuesto anotado en gates/supuestos.md).
	mantenedorDeLosPaquetes = "Jorge <jmorenobl@gmail.com>"

	// instaladorEnLaRelease es scripts/install.sh como lo nombran
	// checksum.extra_files y release.extra_files: una ruta literal que tiene que
	// existir, porque sin él el snapshot falla (research.md D23, V48).
	instaladorEnLaRelease = "./scripts/install.sh"

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

// recetasDeLaRelease son, por objetivo, las recetas de contracts/release.md §3,
// línea a línea como las escribe el Makefile.
var recetasDeLaRelease = map[string][]string{
	"goreleaser-check": {"$(GORELEASER) check"},
	"release":          {"$(GORELEASER) release --snapshot --clean --skip=publish,sign,sbom"},
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
// (contracts/release.md §3).
var controlesDeCI = []string{
	"fmt-check", "lint", "test", "test-integration", "vuln", "schema-check", "skills-check", "goreleaser-check",
	"secrets", "mod-verify", "mod-tidy-check",
}

// ayudaDeLosObjetivos es, por objetivo, un fragmento que su línea de make help
// tiene que llevar para describir lo que hace ahora (FR-121; contracts/release.md
// §3).
var ayudaDeLosObjetivos = map[string]string{
	"goreleaser-check": "goreleaser check",
	"install":          "skills install -g --host claude",
	"release":          "no publica",
	"skills-sync":      "scripts/",
	"snapshot-check":   "instalador-",
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
// hace fallar el test.
type configuracionDeGoreleaser struct {
	Version       int                  `yaml:"version"`
	ProjectName   string               `yaml:"project_name"`
	Builds        []binarioDeLaRelease `yaml:"builds"`
	Archives      []empaquetado        `yaml:"archives"`
	Checksum      sumasDeLaRelease     `yaml:"checksum"`
	SBOMs         []inventario         `yaml:"sboms"`
	Signs         []firma              `yaml:"signs"`
	HomebrewCasks []cask               `yaml:"homebrew_casks"`
	Scoops        []manifiestoDeScoop  `yaml:"scoops"`
	NFPMs         []paqueteDelSistema  `yaml:"nfpms"`
	Release       publicacionEnGitHub  `yaml:"release"`
	Changelog     notasDeLaPublicacion `yaml:"changelog"`
}

type binarioDeLaRelease struct {
	Main    string   `yaml:"main"`
	Binary  string   `yaml:"binary"`
	Env     []string `yaml:"env"`
	Flags   []string `yaml:"flags"`
	Goos    []string `yaml:"goos"`
	Goarch  []string `yaml:"goarch"`
	Ldflags []string `yaml:"ldflags"`
}

type empaquetado struct {
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
// leído de forma estricta y como documento YAML, y el Makefile.
type laRelease struct {
	raiz       fs.FS
	goreleaser configuracionDeGoreleaser
	documento  *yaml.Node
	makefile   makefile
}

// TestConfiguracionDeLaRelease fija lo que contracts/release.md §2 y §3 dicen de
// la release y se puede comprobar sin construirla (research.md D28):
// .goreleaser.yaml tiene exactamente las secciones de la tabla de §2 con sus
// valores —seis plataformas sin cgo y con -trimpath, las cuatro inyecciones del
// Makefile con la plantilla de etiqueta y de snapshot, archivos sin versión en
// el nombre, checksums con install.sh, SBOM, firma con cosign, cask con el
// gancho de la cuarentena, bucket, paquetes, release con install.sh y notas por
// Conventional Commits— y nombra PUBLISHER_TOKEN solo en los dos token; y el
// Makefile tiene los objetivos de §3 con sus recetas y su línea de ayuda, y
// goreleaser-check en ci sin release ni snapshot-check (FR-090 a FR-097,
// FR-121). Que ninguna propiedad esté obsoleta lo comprueba goreleaser check, en
// make ci (FR-094).
func TestConfiguracionDeLaRelease(t *testing.T) {
	t.Parallel()

	release := leerLaRelease(t, os.DirFS("."))

	casos := []struct {
		nombre string
		probar func(*testing.T, laRelease)
	}{
		{"proyecto", probarProyecto},
		{"plataformas", probarPlataformas},
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
	}

	for _, caso := range casos {
		t.Run(caso.nombre, func(t *testing.T) {
			t.Parallel()

			caso.probar(t, release)
		})
	}
}

// leerLaRelease lee .goreleaser.yaml de forma estricta y como documento, y el
// Makefile.
func leerLaRelease(t *testing.T, raiz fs.FS) laRelease {
	t.Helper()

	contenido, err := fs.ReadFile(raiz, ficheroDeGoreleaser)
	require.NoError(t, err)

	leida := laRelease{raiz: raiz, documento: &yaml.Node{}, makefile: leerMakefile(t, raiz)}

	estricto := yaml.NewDecoder(bytes.NewReader(contenido))
	estricto.KnownFields(true)

	require.NoError(t, estricto.Decode(&leida.goreleaser),
		"%s tiene algo que la tabla de contracts/release.md §2 no fija", ficheroDeGoreleaser)
	require.NoError(t, yaml.Unmarshal(contenido, leida.documento))

	return leida
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
// sin cgo y con -trimpath (FR-090).
func probarPlataformas(t *testing.T, release laRelease) {
	t.Helper()

	construida := unica(t, "builds", release.goreleaser.Builds)

	assert.Equal(t, "./cmd/kitlegal", construida.Main)
	assert.Equal(t, nombreDelProyecto, construida.Binary)
	assert.Equal(t, []string{"CGO_ENABLED=0"}, construida.Env)
	assert.Equal(t, []string{"-trimpath"}, construida.Flags)
	assert.ElementsMatch(t, []string{"darwin", "linux", "windows"}, construida.Goos)
	assert.ElementsMatch(t, []string{"amd64", "arm64"}, construida.Goarch)
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
// research.md D22).
func probarArchivos(t *testing.T, release laRelease) {
	t.Helper()

	assert.Equal(t, empaquetado{
		NameTemplate:    "{{ .ProjectName }}_{{ .Os }}_{{ .Arch }}",
		Formats:         []string{"tar.gz"},
		FormatOverrides: []formatoPorSistema{{Goos: "windows", Formats: []string{"zip"}}},
	}, unica(t, "archives", release.goreleaser.Archives))
}

// probarChecksums: checksums.txt, con el SHA-256 por omisión, lleva también la
// huella de install.sh, que existe con la ruta que lo nombra (FR-093;
// research.md D23, V48).
func probarChecksums(t *testing.T, release laRelease) {
	t.Helper()

	assert.Equal(t, sumasDeLaRelease{
		NameTemplate: "checksums.txt",
		ExtraFiles:   []ficheroExtra{{Glob: instaladorEnLaRelease}},
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
// FR-109; research.md V2).
func probarPublicacion(t *testing.T, release laRelease) {
	t.Helper()

	assert.Equal(t, publicacionEnGitHub{
		GitHub:     repositorioDeGitHub{Owner: propietarioEnGitHub, Name: nombreDelProyecto},
		ExtraFiles: []ficheroExtra{{Glob: instaladorEnLaRelease}},
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
// FR-125).
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
// ni snapshot-check (FR-094; contracts/release.md §3).
func probarCI(t *testing.T, release laRelease) {
	t.Helper()

	actual, existe := release.makefile.reglas[objetivoDeCI]
	require.True(t, existe, "el Makefile no tiene el objetivo %s", objetivoDeCI)

	assert.Equal(t, controlesDeCI, actual.prerrequisitos)
}

// probarAyuda: make help describe cada objetivo de la release, install y
// skills-sync con lo que hacen ahora (FR-121).
func probarAyuda(t *testing.T, release laRelease) {
	t.Helper()

	for _, objetivo := range slices.Sorted(maps.Keys(ayudaDeLosObjetivos)) {
		assert.Containsf(t, release.makefile.ayudas[objetivo], ayudaDeLosObjetivos[objetivo],
			"la línea de make help de %s no describe lo que hace (FR-121)", objetivo)
		assert.Containsf(t, release.makefile.reglas, objetivo, "make help describe %s, que no es un objetivo", objetivo)
	}
}
