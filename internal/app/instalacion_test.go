package app

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path"
	"path/filepath"
	"reflect"
	"slices"
	"strings"
	"testing"
	"testing/fstest"

	"github.com/santhosh-tekuri/jsonschema/v6"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/jmorenobl/kitlegal"
	"github.com/jmorenobl/kitlegal/internal/cli"
	"github.com/jmorenobl/kitlegal/internal/core/instalacion"
	"github.com/jmorenobl/kitlegal/internal/core/schema"
	"github.com/jmorenobl/kitlegal/internal/disco"
)

// El registro local de estos tests: el applet skills compuesto con dos skills
// inventadas, alfa, con una referencia, y beta, sin ninguna, en la forma de lo
// empotrado, y con la versión v0.1.0. Ninguna es una skill real, de modo que
// nada de lo que se comprueba aquí dependa de las del repositorio; las reales las
// ejercen los guiones e2e del hito.
const (
	versionDeSkills = "v0.1.0"
	skillMdDeAlfa   = "# alfa\n\nUna skill de prueba.\n"
	normasDeAlfa    = "Las normas de alfa.\n"
	skillMdDeBeta   = "# beta\n"
)

// carpetaAjena es el fichero de una carpeta con el nombre de una skill que no
// puso ahí el binario: el conflicto «carpeta ajena» de FR-041.
const carpetaAjena = "Una carpeta que no es de kitlegal.\n"

// cabeceraDeLosConflictos es la primera línea del mensaje con que install nombra
// cada conflicto (contracts/applet-skills.md §5). Su última palabra va en dos
// literales, como en el dominio, porque misspell, con su diccionario inglés, la
// marca entera como una errata de «conflicts».
const cabeceraDeLosConflictos = "skills install: nada se ha creado ni cambiado; conflict" + "os:"

// firmaDeSkills es la procedencia con la que firma el applet (FR-050). Se escribe
// aquí entera, y no con las constantes del applet, para que un cambio en ellas no
// pase por aquí en silencio.
var firmaDeSkills = schema.Procedencia{Fuente: "kitlegal.skills", URL: "kitlegal:applet/skills"}

// skillsDePrueba es lo empotrado del registro local, nuevo en cada llamada.
func skillsDePrueba() fstest.MapFS {
	return fstest.MapFS{
		"skills/alfa/SKILL.md":             {Data: []byte(skillMdDeAlfa)},
		"skills/alfa/references/normas.md": {Data: []byte(normasDeAlfa)},
		"skills/beta/SKILL.md":             {Data: []byte(skillMdDeBeta)},
	}
}

// dependenciasDePrueba son las del registro local con ese creador de enlaces.
func dependenciasDePrueba(enlazador instalacion.Enlazador) DependenciasDeSkills {
	return DependenciasDeSkills{Version: versionDeSkills, Skills: skillsDePrueba(), Enlazador: enlazador}
}

// registroDeSkills construye el registro de una invocación de prueba: el applet
// skills compuesto con esas dependencias, y nada más.
func registroDeSkills(t *testing.T, dependencias DependenciasDeSkills) *Registro {
	t.Helper()

	var registro Registro

	require.NoError(t, registro.Registrar(AppletSkills(dependencias)))

	return &registro
}

// delSistema es el registro local con el creador de enlaces del sistema.
func delSistema(t *testing.T) *Registro {
	t.Helper()

	return registroDeSkills(t, dependenciasDePrueba(disco.Enlazador{}))
}

// argvDeSkills es la invocación de «kitlegal skills» con los argumentos.
func argvDeSkills(argumentos ...string) []string {
	return slices.Concat([]string{"kitlegal", "skills"}, argumentos)
}

// enUnProyecto hace de un directorio temporal nuevo el directorio de trabajo del
// test, con las carpetas que se le pidan dentro, y lo devuelve: el ámbito local
// es el directorio de trabajo (FR-011), así que cada test necesita el suyo.
func enUnProyecto(t *testing.T, carpetas ...string) string {
	t.Helper()

	proyecto := t.TempDir()
	t.Chdir(proyecto)

	for _, directorio := range carpetas {
		require.NoError(t, os.MkdirAll(directorio, 0o750))
	}

	return proyecto
}

// escribirEnElProyecto deja en ruta, relativa al directorio de trabajo, un
// fichero regular con contenido, creando antes lo que falte hasta él.
func escribirEnElProyecto(t *testing.T, ruta, contenido string) {
	t.Helper()

	require.NoError(t, os.MkdirAll(filepath.Dir(ruta), 0o750))
	require.NoError(t, os.WriteFile(ruta, []byte(contenido), 0o600))
}

// arbolDelProyecto es todo lo que hay por debajo de directorio, entrada a entrada
// y sin seguir ningún enlace: el tipo de cada una, los bytes de cada fichero
// regular y el destino literal de cada enlace. Con él se comprueba que una
// invocación deja el disco byte a byte como estaba. Pasa por el árbol y lee con
// un os.Root, que no sigue los enlaces que salen de él.
func arbolDelProyecto(t *testing.T, directorio string) map[string]string {
	t.Helper()

	raiz, err := os.OpenRoot(directorio)
	require.NoError(t, err)

	arbol := map[string]string{}

	err = fs.WalkDir(raiz.FS(), ".", func(ruta string, entrada fs.DirEntry, err error) error {
		if err != nil {
			return err
		}

		switch tipo := entrada.Type(); {
		case tipo&fs.ModeSymlink != 0:
			destino, err := raiz.Readlink(filepath.FromSlash(ruta))
			arbol[ruta] = "enlace -> " + destino

			return err
		case tipo.IsDir():
			arbol[ruta] = "directorio"
		default:
			contenido, err := raiz.ReadFile(filepath.FromSlash(ruta))
			arbol[ruta] = "fichero: " + string(contenido)

			return err
		}

		return nil
	})
	require.NoError(t, err)
	require.NoError(t, raiz.Close())

	return arbol
}

// exitoDeSkills exige el sobre correcto del applet —código 0, nada en la salida
// de error, un único documento con las seis claves, firmado por él (FR-050)— y
// devuelve su data en el tipo del verbo. Una clave de data que el tipo no tiene
// hace fallar la lectura.
func exitoDeSkills[T any](t *testing.T, res invocacionDePrueba) T {
	t.Helper()

	require.Equal(t, 0, res.codigo, res.errores)
	assert.Empty(t, res.errores, "una invocación correcta no tiene nada que decir en la salida de error")

	sobre := sobreDelJSON(t, res.salida)

	assert.Equal(t, true, sobre["ok"])
	assert.Equal(t, firmaDeSkills.Fuente, sobre["fuente"])
	assert.Equal(t, firmaDeSkills.URL, sobre["url"])

	var crudo struct {
		Data json.RawMessage `json:"data"`
	}

	require.NoError(t, json.Unmarshal([]byte(res.salida), &crudo))

	decodificador := json.NewDecoder(bytes.NewReader(crudo.Data))
	decodificador.DisallowUnknownFields()

	var datos T

	require.NoError(t, decodificador.Decode(&datos))

	return datos
}

// falloDeSkills exige el sobre de fallo con su código, su clase y su
// procedencia, que el mismo mensaje vaya a la salida de error y devuelve el
// mensaje (FR-052; contracts/applet-skills.md §5).
func falloDeSkills(
	t *testing.T, res invocacionDePrueba, clase schema.Clase, codigo int, procedencia schema.Procedencia,
) string {
	t.Helper()

	exigirSobreDeFallo(t, res, clase, codigo, procedencia)

	mensaje, esTexto := datosDelSobre(t, sobreDelJSON(t, res.salida))["mensaje"].(string)
	require.True(t, esTexto, "el mensaje es un texto")
	assert.Contains(t, res.errores, mensaje, "el mismo mensaje sale en la salida de error")

	return mensaje
}

// Las entradas de la salida de install que se repiten en los casos.
func instaladaConEnlace(nombre string, estado instalacion.Estado, modo instalacion.Modo) instalacion.SkillInstalada {
	return instalacion.SkillInstalada{
		Nombre:  nombre,
		Ruta:    ".agents/skills/" + nombre,
		Estado:  estado,
		Enlaces: []instalacion.Enlace{{Host: "claude", Ruta: ".claude/skills/" + nombre, Modo: modo}},
	}
}

func instaladaSinEnlaces(ruta, nombre string, estado instalacion.Estado) instalacion.SkillInstalada {
	return instalacion.SkillInstalada{
		Nombre: nombre, Ruta: path.Join(ruta, nombre), Estado: estado, Enlaces: []instalacion.Enlace{},
	}
}

// enlazadorQueFalla es el creador de enlaces de un sistema de ficheros que no
// los admite: Disponible dice que no y Enlazar falla siempre (contracts/arnes-e2e.md
// §2).
type enlazadorQueFalla struct{}

// errSinEnlacesDePrueba es el fallo de Enlazar del enlazador que falla.
var errSinEnlacesDePrueba = errors.New("este sistema de ficheros no admite enlaces simbólicos")

func (enlazadorQueFalla) Disponible(string) (bool, error) { return false, nil }

func (enlazadorQueFalla) Enlazar(string, string) error { return errSinEnlacesDePrueba }

// enlazadorConSondaQueQueda es el creador de enlaces cuya sonda no se puede
// retirar, que es un fallo de la orden (data-model §3): lo dice como el del
// sistema, nombrando la sonda.
type enlazadorConSondaQueQueda struct{}

func (enlazadorConSondaQueQueda) Disponible(directorio string) (bool, error) {
	return false, fmt.Errorf("retirar la sonda %s: %w", path.Join(directorio, ".kitlegal-sonda-0123456789abcdef"),
		fs.ErrPermission)
}

func (enlazadorConSondaQueQueda) Enlazar(string, string) error { return errSinEnlacesDePrueba }

// enlazadorQueOcupaElManifiesto es el del sistema salvo que, al sondear, deja un
// directorio donde va el manifiesto: lo que otro proceso podría hacer entre la
// comprobación de install y su escritura, que falla así en la fase 3.
type enlazadorQueOcupaElManifiesto struct {
	disco.Enlazador

	manifiesto string
}

func (e enlazadorQueOcupaElManifiesto) Disponible(string) (bool, error) {
	return true, os.Mkdir(e.manifiesto, 0o750)
}

// TestAppletSkills fija el applet skills tal como lo declara
// contracts/applet-skills.md y lo que hace sobre el disco, por la raíz de
// composición entera y sobre el registro local, sin registrarlo en producción:
//
//   - declara install, list y doctor, ninguno por omisión, con sus banderas y el
//     tipo de su data (§1, §4);
//   - valida la invocación antes de tocar el disco, con la precedencia de FR-052,
//     y sale con 2 o, solo con -g sin HOME, con 7, un conflicto con el entorno
//     (§2; ADR 0023);
//   - instala en el ámbito local, en el global y en el de --dir, con las rutas
//     como se alcanzan desde el directorio de trabajo, enlaza en el host y repite
//     sin cambios (§3, §4.1);
//   - list y doctor dan su data, también sin manifiesto, y doctor con sus
//     hallazgos en ella y código 0 (§4.2, §4.3; ADR 0023);
//   - un conflicto sale con 7 y un fallo del disco con 1, con el sobre de fallo
//     del kernel y el mensaje del contrato (§5, §6; ADR 0023);
//   - --dry-run describe por Resultado.Ensayo sin cambiar nada (§7), y --describe
//     describe los tres verbos;
//   - con un creador de enlaces que siempre falla, las entradas de host quedan
//     como copia y doctor no las señala (FR-024, FR-069);
//   - y DependenciasDeSkillsDelSistema lleva la versión, lo empotrado y el
//     creador de enlaces del sistema.
//
// No es paralelo: el ámbito local es el directorio de trabajo, que cada subtest
// cambia con t.Chdir, y el global sale de HOME, que cada uno fija; y el kernel
// lee KITLEGAL_LOG del entorno del proceso, que se fija vacío para que la salida
// de error sea solo la del applet.
func TestAppletSkills(t *testing.T) {
	t.Setenv(cli.VariableNivel, "")

	for _, caso := range casosDelApplet() {
		t.Run(caso.nombre, func(t *testing.T) {
			t.Setenv(variableHome, t.TempDir())

			caso.comprueba(t)
		})
	}

	for _, caso := range casosDeValidacion() {
		t.Run("validacion/"+caso.nombre, func(t *testing.T) {
			t.Setenv(variableHome, "")

			caso.comprueba(t)
		})
	}

	for _, caso := range casosDeDefectoDeComposicion() {
		t.Run("defecto/"+caso.nombre, func(t *testing.T) {
			t.Setenv(variableHome, t.TempDir())

			compruebaDefectoDeSkills(t, caso.dependencias)
		})
	}
}

// casoDelApplet es un subtest de TestAppletSkills.
type casoDelApplet struct {
	nombre    string
	comprueba func(t *testing.T)
}

// casosDelApplet son los subtests de TestAppletSkills que no son de validación
// ni de defecto de composición.
func casosDelApplet() []casoDelApplet {
	return []casoDelApplet{
		{"declara-install-list-y-doctor", compruebaDeclaracionDeSkills},
		{"sin-verbo-sale-con-2", compruebaSkillsSinVerbo},
		{"install-local-enlaza-en-claude-y-repite-sin-cambios", compruebaInstallLocal},
		{"install-de-las-nombradas-una-vez-cada-una", compruebaInstallDeLasNombradas},
		{"list-y-doctor-sin-manifiesto", compruebaListYDoctorSinManifiesto},
		{"list-y-doctor-tras-install", compruebaListYDoctorTrasInstall},
		{"doctor-con-un-hallazgo-sale-con-0", compruebaDoctorConUnHallazgo},
		{"install-con-un-conflicto-sale-con-7", compruebaInstallConUnConflicto},
		{"dry-run-describe-sin-cambiar-nada", compruebaDryRunDeSkills},
		{"enlazador-que-falla-deja-copias-que-doctor-no-senala", compruebaEnlazadorQueFalla},
		{"ambito-global", compruebaAmbitoGlobal},
		{"ambito-dir", compruebaAmbitoDir},
		{"describe-los-tres-verbos", compruebaDescribeDeSkills},
		{"sonda-que-no-se-retira-sale-con-1", compruebaSondaQueNoSeRetira},
		{"fallo-al-escribir-sale-con-1", compruebaFalloAlEscribir},
		{"fallo-del-disco-nombra-el-verbo", compruebaFalloDelDiscoNombraElVerbo},
		{"version-que-el-manifiesto-no-admite", compruebaVersionQueNoSeDeclara},
		{"offline-y-sin-grafo-no-cambian-nada", compruebaOfflineYSinGrafo},
		{"dependencias-del-sistema", compruebaDependenciasDelSistema},
	}
}

// compruebaDeclaracionDeSkills exige la declaración del contrato §1: el nombre,
// los tres verbos en su orden, ninguno por omisión, sus banderas propias y el
// valor cero del tipo de su data; y que el registro lo admite.
func compruebaDeclaracionDeSkills(t *testing.T) {
	t.Helper()

	applet := AppletSkills(dependenciasDePrueba(disco.Enlazador{}))

	assert.Equal(t, "skills", applet.Nombre())
	assert.NotEmpty(t, applet.Descripcion(), "la ayuda del binario lo describe")

	contrato := []struct {
		nombre   string
		banderas []string
		salida   any
	}{
		{nombre: "install", banderas: []string{"skill", "global", "host", "dir"}, salida: []instalacion.SkillInstalada(nil)},
		{nombre: "list", banderas: []string{"global", "dir"}, salida: instalacion.Listado{}},
		{nombre: "doctor", banderas: []string{"global", "dir"}, salida: instalacion.Diagnostico{}},
	}

	verbos := applet.Verbos()
	require.Len(t, verbos, len(contrato))

	for i, verbo := range verbos {
		assert.Equal(t, contrato[i].nombre, verbo.Nombre)
		assert.NotEmpty(t, verbo.Descripcion, "la ayuda del applet describe %q", verbo.Nombre)
		assert.False(t, verbo.PorOmision, "%q no es por omisión: skills no tiene ninguno", verbo.Nombre)
		assert.Equal(t, reflect.TypeOf(contrato[i].salida), reflect.TypeOf(verbo.Salida))
		assert.True(t, reflect.ValueOf(verbo.Salida).IsZero(), "Salida es el valor cero de su tipo")

		require.NotNil(t, verbo.Argumentos)

		primero, segundo := verbo.Argumentos(), verbo.Argumentos()
		assert.NotSame(t, primero, segundo, "la fábrica da un valor nuevo en cada invocación")
		assert.Equal(t, contrato[i].banderas, banderasPropias(primero), "las de §1 y ninguna más")
	}

	var registro Registro

	require.NoError(t, registro.Registrar(applet), "el registro lo admite")
}

// banderasPropias son los nombres con que se escriben los campos exportados de
// los argumentos de un verbo, en su orden.
func banderasPropias(argumentos Argumentos) []string {
	tipo := reflect.TypeOf(argumentos).Elem()

	var nombres []string

	for i := range tipo.NumField() {
		campo := tipo.Field(i)
		if !campo.IsExported() {
			continue
		}

		nombre := strings.ToLower(campo.Name)
		if etiqueta, renombrado := campo.Tag.Lookup("name"); renombrado {
			nombre = etiqueta
		}

		nombres = append(nombres, nombre)
	}

	return nombres
}

// compruebaSkillsSinVerbo exige que skills sin verbo salga con 2, lo firme el
// kernel y nombre los tres verbos: ninguno es por omisión (§1).
func compruebaSkillsSinVerbo(t *testing.T) {
	t.Helper()

	enUnProyecto(t)

	res := invocar(t, delSistema(t), argvDeSkills("--json")...)

	falloDeSkills(t, res, schema.ClaseArgumentos, 2, cli.ProcedenciaKernel())

	for _, verbo := range []string{"install", "list", "doctor"} {
		assert.Contains(t, res.errores, verbo, "nombra el verbo %q", verbo)
	}
}

// compruebaInstallLocal instala las dos skills en un proyecto con .claude/: el
// sobre las da instaladas y enlazadas, el disco lleva los bytes empotrados y los
// enlaces relativos de FR-021, y la segunda ejecución no cambia nada y lo dice
// (FR-010, FR-014, FR-021, FR-022, FR-045, FR-051).
func compruebaInstallLocal(t *testing.T) {
	t.Helper()

	enUnProyecto(t, ".claude")

	registro := delSistema(t)

	instaladas := exitoDeSkills[[]instalacion.SkillInstalada](t, invocar(t, registro, argvDeSkills("install", "--json")...))
	assert.Equal(t, []instalacion.SkillInstalada{
		instaladaConEnlace("alfa", instalacion.EstadoInstalada, instalacion.ModoEnlace),
		instaladaConEnlace("beta", instalacion.EstadoInstalada, instalacion.ModoEnlace),
	}, instaladas)

	arbol := arbolDelProyecto(t, ".")

	assert.Equal(t, "fichero: "+skillMdDeAlfa, arbol[".agents/skills/alfa/SKILL.md"])
	assert.Equal(t, "fichero: "+normasDeAlfa, arbol[".agents/skills/alfa/references/normas.md"])
	assert.Equal(t, "fichero: "+skillMdDeBeta, arbol[".agents/skills/beta/SKILL.md"])
	assert.Equal(t, "enlace -> ../../.agents/skills/alfa", arbol[".claude/skills/alfa"])
	assert.Equal(t, "enlace -> ../../.agents/skills/beta", arbol[".claude/skills/beta"])
	assert.Contains(t, arbol[".agents/skills/kitlegal.json"], `"`+versionDeSkills+`"`, "el manifiesto con la versión")

	segunda := exitoDeSkills[[]instalacion.SkillInstalada](t, invocar(t, registro, argvDeSkills("install", "--json")...))
	assert.Equal(t, []instalacion.SkillInstalada{
		instaladaConEnlace("alfa", instalacion.EstadoSinCambios, instalacion.ModoEnlace),
		instaladaConEnlace("beta", instalacion.EstadoSinCambios, instalacion.ModoEnlace),
	}, segunda)
	assert.Equal(t, arbol, arbolDelProyecto(t, "."), "la segunda ejecución no cambia nada")
}

// compruebaInstallDeLasNombradas instala solo las skills nombradas, una vez
// cada una aunque se repitan, y sin .claude/ no enlaza nada (FR-010, FR-022).
func compruebaInstallDeLasNombradas(t *testing.T) {
	t.Helper()

	enUnProyecto(t)

	registro := delSistema(t)

	res := invocar(t, registro, argvDeSkills("install", "beta", "beta", "--json")...)
	assert.Equal(t, []instalacion.SkillInstalada{
		instaladaSinEnlaces(".agents/skills", "beta", instalacion.EstadoInstalada),
	}, exitoDeSkills[[]instalacion.SkillInstalada](t, res))
	assert.Contains(t, res.salida, `"enlaces":[]`, "una lista vacía, nunca null")

	arbol := arbolDelProyecto(t, ".")
	assert.Contains(t, arbol, ".agents/skills/beta/SKILL.md")
	assert.NotContains(t, arbol, ".agents/skills/alfa", "alfa no se pidió")
	assert.NotContains(t, arbol, ".claude", "sin .claude/ no se enlaza en ningún host")

	assert.Equal(t, []instalacion.SkillInstalada{
		instaladaSinEnlaces(".agents/skills", "alfa", instalacion.EstadoInstalada),
	}, exitoDeSkills[[]instalacion.SkillInstalada](t, invocar(t, registro, argvDeSkills("install", "alfa", "--json")...)))
}

// compruebaListYDoctorSinManifiesto exige que, sin manifiesto, list y doctor
// salgan con 0, sin nada en la salida de error, con manifiesto no, versión nula
// y listas vacías (FR-061, FR-067).
func compruebaListYDoctorSinManifiesto(t *testing.T) {
	t.Helper()

	enUnProyecto(t)

	registro := delSistema(t)

	list := invocar(t, registro, argvDeSkills("list", "--json")...)
	assert.Equal(t, instalacion.Listado{Directorio: ".agents/skills", Skills: []instalacion.SkillListada{}},
		exitoDeSkills[instalacion.Listado](t, list))
	assert.Contains(t, list.salida, `"version":null`)

	doctor := invocar(t, registro, argvDeSkills("doctor", "--json")...)
	assert.Equal(t, instalacion.Diagnostico{
		Directorio: ".agents/skills", VersionDelBinario: versionDeSkills, Hallazgos: []instalacion.Hallazgo{},
	}, exitoDeSkills[instalacion.Diagnostico](t, doctor))
	assert.Contains(t, doctor.salida, `"version":null`)

	assert.Equal(t, map[string]string{".": "directorio"}, arbolDelProyecto(t, "."), "ni list ni doctor escriben")
}

// compruebaListYDoctorTrasInstall exige lo que dan list y doctor de una
// instalación sana: cada skill con su versión, empotrada y con sus enlaces, y
// ningún hallazgo (FR-060, FR-067).
func compruebaListYDoctorTrasInstall(t *testing.T) {
	t.Helper()

	enUnProyecto(t, ".claude")

	registro := delSistema(t)

	require.Equal(t, 0, invocar(t, registro, argvDeSkills("install")...).codigo)

	arbol := arbolDelProyecto(t, ".")
	version := versionDeSkills

	listada := func(nombre string) instalacion.SkillListada {
		return instalacion.SkillListada{
			Nombre: nombre, Ruta: ".agents/skills/" + nombre, Version: versionDeSkills, Empotrada: true,
			Enlaces: []instalacion.Enlace{{Host: "claude", Ruta: ".claude/skills/" + nombre, Modo: instalacion.ModoEnlace}},
		}
	}

	assert.Equal(t, instalacion.Listado{
		Directorio: ".agents/skills", Manifiesto: true, Version: &version,
		Skills: []instalacion.SkillListada{listada("alfa"), listada("beta")},
	}, exitoDeSkills[instalacion.Listado](t, invocar(t, registro, argvDeSkills("list", "--json")...)))

	assert.Equal(t, instalacion.Diagnostico{
		Directorio: ".agents/skills", Manifiesto: true, Version: &version, VersionDelBinario: versionDeSkills,
		Hallazgos: []instalacion.Hallazgo{},
	}, exitoDeSkills[instalacion.Diagnostico](t, invocar(t, registro, argvDeSkills("doctor", "--json")...)))

	assert.Equal(t, arbol, arbolDelProyecto(t, "."), "ni list ni doctor cambian nada")
}

// compruebaDoctorConUnHallazgo edita un fichero instalado: doctor sale con 0,
// con el sobre correcto del applet y el hallazgo en su data —clase, ruta y la
// orden que lo arregla—, sin cambiar nada. Encontrar algo es el resultado de una
// verificación que ha funcionado, no un fallo (FR-065 a FR-068;
// contracts/applet-skills.md §5 y §6; ADR 0023).
func compruebaDoctorConUnHallazgo(t *testing.T) {
	t.Helper()

	enUnProyecto(t, ".claude")

	registro := delSistema(t)

	require.Equal(t, 0, invocar(t, registro, argvDeSkills("install")...).codigo)
	escribirEnElProyecto(t, ".agents/skills/alfa/SKILL.md", "editado a mano\n")

	arbol := arbolDelProyecto(t, ".")

	diagnostico := exitoDeSkills[instalacion.Diagnostico](t, invocar(t, registro, argvDeSkills("doctor", "--json")...))

	assert.Equal(t, []instalacion.Hallazgo{{
		Clase: instalacion.HallazgoFicheroEditado,
		Ruta:  ".agents/skills/alfa/SKILL.md",
		Orden: "rm -- '.agents/skills/alfa/SKILL.md' && kitlegal skills install alfa --host claude",
	}}, diagnostico.Hallazgos)
	assert.Equal(t, arbol, arbolDelProyecto(t, "."), "doctor no cambia nada")
}

// compruebaInstallConUnConflicto deja una carpeta ajena con el nombre de una
// skill: install sale con 7, un conflicto con el estado local, y la nombra, sin
// crear ni cambiar nada, con la bandera --dry-run y sin ella: el conflicto se
// conoce sin efectos, así que el ensayo predice el código de la orden real
// (FR-040 a FR-042, FR-048, FR-052; ADR 0023).
func compruebaInstallConUnConflicto(t *testing.T) {
	t.Helper()

	enUnProyecto(t, ".claude")
	escribirEnElProyecto(t, ".agents/skills/alfa/mio.md", carpetaAjena)

	registro := delSistema(t)
	arbol := arbolDelProyecto(t, ".")

	mensaje := falloDeSkills(t, invocar(t, registro, argvDeSkills("install", "--json")...),
		schema.ClaseConflicto, 7, firmaDeSkills)
	assert.Equal(t, cabeceraDeLosConflictos+"\ncarpeta ajena: .agents/skills/alfa", mensaje)
	assert.Equal(t, arbol, arbolDelProyecto(t, "."), "nada se crea ni se cambia")

	// Con --dry-run, el mismo fallo y el mismo código; el sobre lo firma el
	// kernel, que descarta el Resultado del applet en ensayo.
	ensayo := invocar(t, registro, argvDeSkills("install", "--dry-run", "--json")...)
	assert.Equal(t, mensaje, falloDeSkills(t, ensayo, schema.ClaseConflicto, 7, cli.ProcedenciaKernel()))
	assert.True(t, strings.HasPrefix(ensayo.errores, prefijoDeEnsayo+"no se ha ejecutado nada"), ensayo.errores)
	assert.NotContains(t, ensayo.errores, "se habría pedido", "con un conflicto no describe ninguna skill")
	assert.Equal(t, arbol, arbolDelProyecto(t, "."), "ni con --dry-run")
}

// compruebaDryRunDeSkills exige que install --dry-run describa en la salida de
// error una línea por skill pedida, con su estado y cada enlace con su modo, sin
// sobre y sin ningún cambio en disco, ni carpetas vacías; y que list y doctor
// tampoco cambien nada con la bandera (FR-048; contracts/applet-skills.md §7).
func compruebaDryRunDeSkills(t *testing.T) {
	t.Helper()

	enUnProyecto(t, ".claude")

	registro := delSistema(t)
	arbol := arbolDelProyecto(t, ".")

	compruebaEnsayoDeInstall(t, registro, []string{
		"instalar alfa en .agents/skills/alfa: instalada; Claude Code .claude/skills/alfa (enlace)",
		"instalar beta en .agents/skills/beta: instalada; Claude Code .claude/skills/beta (enlace)",
	})
	assert.Equal(t, arbol, arbolDelProyecto(t, "."), "--dry-run no cambia nada")

	require.Equal(t, 0, invocar(t, registro, argvDeSkills("install")...).codigo)

	arbol = arbolDelProyecto(t, ".")

	compruebaEnsayoDeInstall(t, registro, []string{
		"instalar alfa en .agents/skills/alfa: sin cambios; Claude Code .claude/skills/alfa (enlace)",
		"instalar beta en .agents/skills/beta: sin cambios; Claude Code .claude/skills/beta (enlace)",
	})

	for _, verbo := range []string{"list", "doctor"} {
		res := invocar(t, registro, argvDeSkills(verbo, "--dry-run", "--json")...)

		assert.Equal(t, 0, res.codigo, res.errores)
		assert.Empty(t, res.salida, "con --dry-run no hay sobre")
		assert.NotContains(t, res.errores, "se habría pedido", "%s no describe nada", verbo)
	}

	assert.Equal(t, arbol, arbolDelProyecto(t, "."), "ni install, ni list, ni doctor con --dry-run")
}

// compruebaEnsayoDeInstall exige que install --dry-run salga con 0, sin nada en
// la salida estándar, y con la línea del kernel seguida de una por línea
// esperada, con su prefijo.
func compruebaEnsayoDeInstall(t *testing.T, registro *Registro, esperadas []string) {
	t.Helper()

	res := invocar(t, registro, argvDeSkills("install", "--dry-run", "--json")...)

	require.Equal(t, 0, res.codigo, res.errores)
	assert.Empty(t, res.salida, "con --dry-run no hay sobre")

	lineas := strings.Split(strings.TrimSuffix(res.errores, "\n"), "\n")
	require.NotEmpty(t, lineas)
	assert.True(t, strings.HasPrefix(lineas[0], prefijoDeEnsayo+
		`no se ha ejecutado nada; se habría ejecutado el applet "skills", el verbo "install"`), lineas[0])

	descritas := make([]string, 0, len(esperadas))
	for _, esperada := range esperadas {
		descritas = append(descritas, prefijoDeEnsayo+"se habría pedido "+esperada)
	}

	assert.Equal(t, descritas, lineas[1:])
}

// compruebaEnlazadorQueFalla compone el applet con un creador de enlaces que
// siempre falla: --dry-run predice copia, install deja cada entrada de host como
// una copia real de lo empotrado con modo copia, list lo dice, doctor no lo
// señala porque el enlace sigue sin poder crearse, y repetir no cambia nada
// (FR-024, FR-045, FR-048, FR-069).
func compruebaEnlazadorQueFalla(t *testing.T) {
	t.Helper()

	enUnProyecto(t, ".claude")

	registro := registroDeSkills(t, dependenciasDePrueba(enlazadorQueFalla{}))

	compruebaEnsayoDeInstall(t, registro, []string{
		"instalar alfa en .agents/skills/alfa: instalada; Claude Code .claude/skills/alfa (copia)",
		"instalar beta en .agents/skills/beta: instalada; Claude Code .claude/skills/beta (copia)",
	})

	instaladas := exitoDeSkills[[]instalacion.SkillInstalada](t, invocar(t, registro, argvDeSkills("install", "--json")...))
	assert.Equal(t, []instalacion.SkillInstalada{
		instaladaConEnlace("alfa", instalacion.EstadoInstalada, instalacion.ModoCopia),
		instaladaConEnlace("beta", instalacion.EstadoInstalada, instalacion.ModoCopia),
	}, instaladas)

	arbol := arbolDelProyecto(t, ".")

	assert.Equal(t, "directorio", arbol[".claude/skills/alfa"], "la entrada de host es un directorio real")
	assert.Equal(t, "fichero: "+skillMdDeAlfa, arbol[".claude/skills/alfa/SKILL.md"])
	assert.Equal(t, "fichero: "+normasDeAlfa, arbol[".claude/skills/alfa/references/normas.md"])
	assert.Equal(t, "fichero: "+skillMdDeBeta, arbol[".claude/skills/beta/SKILL.md"])

	listado := exitoDeSkills[instalacion.Listado](t, invocar(t, registro, argvDeSkills("list", "--json")...))
	require.Len(t, listado.Skills, 2)

	for _, skill := range listado.Skills {
		assert.Equal(t, []instalacion.Enlace{
			{Host: "claude", Ruta: ".claude/skills/" + skill.Nombre, Modo: instalacion.ModoCopia},
		}, skill.Enlaces)
	}

	diagnostico := exitoDeSkills[instalacion.Diagnostico](t, invocar(t, registro, argvDeSkills("doctor", "--json")...))
	assert.Empty(t, diagnostico.Hallazgos, "una copia donde no se puede enlazar no es un hallazgo")

	segunda := exitoDeSkills[[]instalacion.SkillInstalada](t, invocar(t, registro, argvDeSkills("install", "--json")...))
	assert.Equal(t, []instalacion.SkillInstalada{
		instaladaConEnlace("alfa", instalacion.EstadoSinCambios, instalacion.ModoCopia),
		instaladaConEnlace("beta", instalacion.EstadoSinCambios, instalacion.ModoCopia),
	}, segunda)
	assert.Equal(t, arbol, arbolDelProyecto(t, "."), "la segunda ejecución no cambia nada")
}

// compruebaAmbitoGlobal instala con -g: el directorio neutro y el host cuelgan
// de HOME, las rutas se presentan absolutas y el directorio de trabajo queda
// intacto (FR-012; contracts/applet-skills.md §3).
func compruebaAmbitoGlobal(t *testing.T) {
	t.Helper()

	home := t.TempDir()
	t.Setenv(variableHome, home)
	require.NoError(t, os.Mkdir(filepath.Join(home, ".claude"), 0o750))

	enUnProyecto(t)

	registro := delSistema(t)
	neutro := home + "/.agents/skills"

	conEnlace := func(nombre string) instalacion.SkillInstalada {
		return instalacion.SkillInstalada{
			Nombre: nombre, Ruta: neutro + "/" + nombre, Estado: instalacion.EstadoInstalada,
			Enlaces: []instalacion.Enlace{
				{Host: "claude", Ruta: home + "/.claude/skills/" + nombre, Modo: instalacion.ModoEnlace},
			},
		}
	}

	assert.Equal(t, []instalacion.SkillInstalada{conEnlace("alfa"), conEnlace("beta")},
		exitoDeSkills[[]instalacion.SkillInstalada](t, invocar(t, registro, argvDeSkills("install", "-g", "--json")...)))

	enHome := arbolDelProyecto(t, home)
	assert.Equal(t, "fichero: "+skillMdDeAlfa, enHome[".agents/skills/alfa/SKILL.md"])
	assert.Equal(t, "enlace -> ../../.agents/skills/alfa", enHome[".claude/skills/alfa"])

	listado := exitoDeSkills[instalacion.Listado](t, invocar(t, registro, argvDeSkills("list", "--global", "--json")...))
	assert.Equal(t, neutro, listado.Directorio)
	assert.Len(t, listado.Skills, 2)

	diagnostico := exitoDeSkills[instalacion.Diagnostico](t, invocar(t, registro, argvDeSkills("doctor", "-g", "--json")...))
	assert.Equal(t, neutro, diagnostico.Directorio)
	assert.Empty(t, diagnostico.Hallazgos)

	assert.Equal(t, map[string]string{".": "directorio"}, arbolDelProyecto(t, "."),
		"-g no toca el directorio de trabajo")
}

// compruebaAmbitoDir instala con --dir relativo: el directorio neutro es esa
// ruta, tal como se pasó, y no se enlaza en ningún host aunque .claude/ exista
// (FR-013; contracts/applet-skills.md §3).
func compruebaAmbitoDir(t *testing.T) {
	t.Helper()

	enUnProyecto(t, ".claude")

	registro := delSistema(t)

	assert.Equal(t, []instalacion.SkillInstalada{
		instaladaSinEnlaces("mis-skills", "alfa", instalacion.EstadoInstalada),
		instaladaSinEnlaces("mis-skills", "beta", instalacion.EstadoInstalada),
	}, exitoDeSkills[[]instalacion.SkillInstalada](t, invocar(t, registro,
		argvDeSkills("install", "--dir", "mis-skills/", "--json")...)))

	arbol := arbolDelProyecto(t, ".")
	assert.Equal(t, "fichero: "+skillMdDeAlfa, arbol["mis-skills/alfa/SKILL.md"])
	assert.Contains(t, arbol, "mis-skills/kitlegal.json")
	assert.NotContains(t, arbol, ".claude/skills", "--dir no tiene hosts")
	assert.NotContains(t, arbol, ".agents", "--dir sustituye al directorio neutro")

	listado := exitoDeSkills[instalacion.Listado](t, invocar(t, registro, argvDeSkills("list", "--dir", "mis-skills", "--json")...))
	assert.Equal(t, "mis-skills", listado.Directorio)
	assert.Len(t, listado.Skills, 2)

	for _, skill := range listado.Skills {
		assert.Empty(t, skill.Enlaces)
	}

	diagnostico := exitoDeSkills[instalacion.Diagnostico](t, invocar(t, registro,
		argvDeSkills("doctor", "--dir", "mis-skills", "--json")...))
	assert.Empty(t, diagnostico.Hallazgos)
}

// compruebaDescribeDeSkills exige que --describe de cada verbo salga con 0 sin
// leer nada, con su título, sus banderas propias en la entrada, ninguna
// obligatoria, y una salida contra la que valida el sobre real del verbo; y el
// de fallo, el de install con un conflicto.
func compruebaDescribeDeSkills(t *testing.T) {
	t.Helper()

	enUnProyecto(t, ".claude")

	registro := delSistema(t)

	esquemas := map[string]*jsonschema.Schema{}

	for verbo, banderas := range map[string][]string{
		"install": {"skill", "global", "host", "dir"},
		"list":    {"global", "dir"},
		"doctor":  {"global", "dir"},
	} {
		esquemas[verbo] = esquemaDeSkillsEmitido(t, registro, verbo, banderas)
	}

	assert.Equal(t, map[string]string{".": "directorio", ".claude": "directorio"}, arbolDelProyecto(t, "."),
		"--describe no ejecuta nada")

	for _, verbo := range []string{"list", "doctor", "install", "list", "doctor"} {
		res := invocar(t, registro, argvDeSkills(verbo, "--json")...)
		require.Equal(t, 0, res.codigo, res.errores)
		require.NoError(t, esquemas[verbo].Validate(sobreValidable(t, res.salida)),
			"el sobre de %s valida contra su --describe", verbo)
	}

	escribirEnElProyecto(t, ".agents/skills/alfa/SKILL.md", "editado a mano\n")

	res := invocar(t, registro, argvDeSkills("doctor", "--json")...)
	require.Equal(t, 0, res.codigo, res.errores)
	require.NoError(t, esquemas["doctor"].Validate(sobreValidable(t, res.salida)),
		"el sobre con un hallazgo valida contra la rama then de su --describe (ADR 0023)")

	res = invocar(t, registro, argvDeSkills("install", "--json")...)
	require.Equal(t, 7, res.codigo, res.errores)
	require.NoError(t, esquemas["install"].Validate(sobreValidable(t, res.salida)),
		"el sobre de un conflicto valida contra la rama else de su --describe")
}

// esquemaDeSkillsEmitido pide el --describe del verbo, exige su título y sus
// banderas propias en la entrada, ninguna obligatoria, y devuelve su salida
// compilada.
func esquemaDeSkillsEmitido(t *testing.T, registro *Registro, verbo string, banderas []string) *jsonschema.Schema {
	t.Helper()

	res := invocar(t, registro, argvDeSkills(verbo, "--describe")...)
	require.Equal(t, 0, res.codigo, res.errores)

	documento, err := jsonschema.UnmarshalJSON(strings.NewReader(res.salida))
	require.NoError(t, err)

	raiz, esObjeto := documento.(map[string]any)
	require.True(t, esObjeto)
	assert.Equal(t, "skills "+verbo, raiz["title"])

	entrada, esObjeto := valorDelEsquema(t, documento, "properties", "entrada").(map[string]any)
	require.True(t, esObjeto)

	propiedades, esObjeto := entrada["properties"].(map[string]any)
	require.True(t, esObjeto)

	for _, bandera := range banderas {
		assert.Contains(t, propiedades, bandera, "la entrada de %s describe %s", verbo, bandera)
	}

	obligatorias, _ := entrada["required"].([]any)
	for _, bandera := range banderas {
		assert.NotContains(t, obligatorias, bandera, "%s no es obligatoria en %s", bandera, verbo)
	}

	const id = "https://kitlegal.es/schemas/skills/describe.json"

	compilador := jsonschema.NewCompiler()
	compilador.AssertFormat()
	require.NoError(t, compilador.AddResource(id, documento))

	esquema, err := compilador.Compile(id + "#/properties/salida")
	require.NoError(t, err, "el --describe de %s compila", verbo)

	return esquema
}

// compruebaSondaQueNoSeRetira compone el applet con un creador de enlaces cuya
// sonda no se puede retirar: install, también con --dry-run, y doctor salen con
// 1 nombrando el verbo y la sonda, sin ningún conflicto ni hallazgo y sin cambiar nada
// (contracts/applet-skills.md §5).
func compruebaSondaQueNoSeRetira(t *testing.T) {
	t.Helper()

	enUnProyecto(t, ".claude")

	queda := registroDeSkills(t, dependenciasDePrueba(enlazadorConSondaQueQueda{}))
	arbol := arbolDelProyecto(t, ".")

	mensaje := falloDeSkills(t, invocar(t, queda, argvDeSkills("install", "--json")...),
		schema.ClaseInesperado, 1, firmaDeSkills)
	assert.Equal(t, "skills install: retirar la sonda .claude/.kitlegal-sonda-0123456789abcdef: "+
		fs.ErrPermission.Error(), mensaje)

	ensayo := invocar(t, queda, argvDeSkills("install", "--dry-run", "--json")...)
	assert.Equal(t, mensaje, falloDeSkills(t, ensayo, schema.ClaseInesperado, 1, cli.ProcedenciaKernel()))
	assert.Equal(t, arbol, arbolDelProyecto(t, "."))

	// doctor solo sondea con una copia declarada: la deja un install con el
	// enlazador que falla.
	require.Equal(t, 0, invocar(t, registroDeSkills(t, dependenciasDePrueba(enlazadorQueFalla{})),
		argvDeSkills("install")...).codigo)

	mensaje = falloDeSkills(t, invocar(t, queda, argvDeSkills("doctor", "--json")...),
		schema.ClaseInesperado, 1, firmaDeSkills)
	assert.Equal(t, "skills doctor: retirar la sonda .claude/skills/.kitlegal-sonda-0123456789abcdef: "+
		fs.ErrPermission.Error(), mensaje)
}

// compruebaFalloAlEscribir hace fallar la escritura del manifiesto tras la
// comprobación: install sale con 1 y el mensaje nombra el verbo una sola vez, la
// operación y la ruta (FR-044; contracts/applet-skills.md §5).
func compruebaFalloAlEscribir(t *testing.T) {
	t.Helper()

	enUnProyecto(t, ".claude", ".agents/skills")

	manifiesto := ".agents/skills/kitlegal.json"
	registro := registroDeSkills(t, dependenciasDePrueba(enlazadorQueOcupaElManifiesto{manifiesto: manifiesto}))

	mensaje := falloDeSkills(t, invocar(t, registro, argvDeSkills("install", "--json")...),
		schema.ClaseInesperado, 1, firmaDeSkills)

	assert.True(t, strings.HasPrefix(mensaje, "skills install: escribir "+manifiesto+": "), mensaje)
	assert.Equal(t, 1, strings.Count(mensaje, "skills install:"), "el verbo, una sola vez: %s", mensaje)
}

// compruebaFalloDelDiscoNombraElVerbo hace fallar el disco al examinar la ruta
// de --dir, que pasa por un fichero regular: install, list y doctor salen con 1
// y el mensaje empieza por «skills <verbo>: » una sola vez, seguido de la
// operación y la ruta, sin cambiar nada (contracts/applet-skills.md §5). El
// fallo se provoca por la forma del árbol y no por un permiso, para que no
// dependa de quién ejecuta el test.
func compruebaFalloDelDiscoNombraElVerbo(t *testing.T) {
	t.Helper()

	enUnProyecto(t)
	escribirEnElProyecto(t, "fichero", "no soy un directorio\n")

	registro := delSistema(t)
	arbol := arbolDelProyecto(t, ".")

	for _, verbo := range []string{"install", "list", "doctor"} {
		mensaje := falloDeSkills(t, invocar(t, registro, argvDeSkills(verbo, "--dir", "fichero/sub", "--json")...),
			schema.ClaseInesperado, 1, firmaDeSkills)

		cabecera := "skills " + verbo + ": examinar fichero/sub: "
		assert.True(t, strings.HasPrefix(mensaje, cabecera), "%s: %s", verbo, mensaje)
		assert.NotEqual(t, cabecera, mensaje, "%s: el mensaje dice el error del sistema", verbo)
		assert.Equal(t, 1, strings.Count(mensaje, "skills "), "%s: el verbo, una sola vez: %s", verbo, mensaje)
	}

	assert.Equal(t, arbol, arbolDelProyecto(t, "."), "ninguno de los tres cambia nada")
}

// compruebaVersionQueNoSeDeclara compone el applet con una versión vacía, que
// el manifiesto no admitiría: install y doctor salen con 1 nombrando su verbo, y
// list, que no la declara ni la compara, sale con 0.
func compruebaVersionQueNoSeDeclara(t *testing.T) {
	t.Helper()

	enUnProyecto(t)

	dependencias := dependenciasDePrueba(disco.Enlazador{})
	dependencias.Version = ""
	registro := registroDeSkills(t, dependencias)

	for _, verbo := range []string{"install", "doctor"} {
		mensaje := falloDeSkills(t, invocar(t, registro, argvDeSkills(verbo, "--json")...),
			schema.ClaseInesperado, 1, firmaDeSkills)
		assert.True(t, strings.HasPrefix(mensaje,
			"skills "+verbo+": la versión del binario no se puede declarar en el manifiesto"), mensaje)
	}

	assert.Equal(t, 0, invocar(t, registro, argvDeSkills("list", "--json")...).codigo)
	assert.Equal(t, map[string]string{".": "directorio"}, arbolDelProyecto(t, "."))
}

// compruebaOfflineYSinGrafo exige que --offline, --no-graph y --asunto no
// cambien lo que dan los verbos: el applet no abre ninguna conexión (FR-016) ni
// emite operaciones de grafo (FR-054).
func compruebaOfflineYSinGrafo(t *testing.T) {
	t.Helper()

	enUnProyecto(t, ".claude")

	registro := delSistema(t)

	require.Equal(t, 0, invocar(t, registro, argvDeSkills("install", "--offline")...).codigo)

	listado := exitoDeSkills[instalacion.Listado](t, invocar(t, registro, argvDeSkills("list", "--json")...))

	for _, banderas := range [][]string{{"--offline"}, {"--no-graph"}, {"--asunto", "demo"}} {
		argv := argvDeSkills(slices.Concat([]string{"list", "--json"}, banderas)...)
		assert.Equal(t, listado, exitoDeSkills[instalacion.Listado](t, invocar(t, registro, argv...)), "%q", argv)

		argv = argvDeSkills(slices.Concat([]string{"install", "--json"}, banderas)...)
		for _, skill := range exitoDeSkills[[]instalacion.SkillInstalada](t, invocar(t, registro, argv...)) {
			assert.Equal(t, instalacion.EstadoSinCambios, skill.Estado, "%q", argv)
		}
	}
}

// compruebaDependenciasDelSistema exige que las dependencias del binario
// distribuido lleven la versión que reciben, lo empotrado en el binario y el
// creador de enlaces del sistema de internal/disco (FR-024; research.md D9).
func compruebaDependenciasDelSistema(t *testing.T) {
	t.Helper()

	dependencias := DependenciasDeSkillsDelSistema(versionDeSkills)

	assert.Equal(t, versionDeSkills, dependencias.Version)
	assert.Equal(t, "v9.9.9", DependenciasDeSkillsDelSistema("v9.9.9").Version, "la que recibe, no una fija")

	require.NotNil(t, dependencias.Skills)

	leidas, err := skillsEmpotradasDe(dependencias.Skills)
	require.NoError(t, err)

	empotradas, err := skillsEmpotradasDe(kitlegal.Skills())
	require.NoError(t, err)
	assert.Equal(t, empotradas, leidas, "lo empotrado en el binario")

	require.NotNil(t, dependencias.Enlazador)
	assert.IsType(t, disco.Enlazador{}, dependencias.Enlazador, "el creador de enlaces del sistema")
}

// casoDeValidacion es una invocación que la validación rechaza antes de tocar el
// disco, con su código, su clase y lo que dice su mensaje (contracts/applet-skills.md
// §2). Todas se hacen con HOME vacío y en un proyecto con una carpeta ajena con el
// nombre de alfa: la precedencia de FR-052 hace que un error de argumentos gane
// al de HOME y al conflicto, que no llegan a comprobarse.
type casoDeValidacion struct {
	nombre  string
	argv    []string
	codigo  int
	clase   schema.Clase
	mensaje string
	// delKernel dice que el rechazo es del analizador del kernel, anterior al
	// applet, que lo firma.
	delKernel bool
}

// Los mensajes de la tabla del contrato §2.
const (
	mensajeGlobalConDir   = "-g y --dir se excluyen"
	mensajeHostConDir     = "--host no se combina con --dir"
	mensajeHostNoAdmitido = "los hosts admitidos son claude y antigravity"
	mensajeSinHome        = "HOME no está definido o está vacío"
	mensajeSkillDeFuera   = `"desconocida" no es ninguna skill de este binario; skills disponibles: alfa, beta`
	mensajeSkillDeFueraB  = `"otra" no es ninguna skill de este binario; skills disponibles: alfa, beta`
)

// casosDeValidacion son las filas del contrato §2 en los tres verbos y cada
// combinación en la que decide la precedencia.
func casosDeValidacion() []casoDeValidacion {
	argumentos := func(nombre, mensaje string, argv ...string) casoDeValidacion {
		return casoDeValidacion{
			nombre: nombre, argv: argvDeSkills(argv...), codigo: 2, clase: schema.ClaseArgumentos, mensaje: mensaje,
		}
	}
	sinHome := func(nombre string, argv ...string) casoDeValidacion {
		return casoDeValidacion{
			nombre: nombre, argv: argvDeSkills(argv...), codigo: 7, clase: schema.ClaseConflicto, mensaje: mensajeSinHome,
		}
	}

	return []casoDeValidacion{
		argumentos("install-g-y-dir", mensajeGlobalConDir, "install", "-g", "--dir", "otro"),
		argumentos("list-g-y-dir", mensajeGlobalConDir, "list", "-g", "--dir", "otro"),
		argumentos("doctor-global-y-dir", mensajeGlobalConDir, "doctor", "--global", "--dir", "otro"),
		argumentos("g-y-dir-antes-que-el-host", mensajeGlobalConDir, "install", "-g", "--dir", "otro", "--host", "codex"),
		argumentos("g-y-dir-con-dry-run", mensajeGlobalConDir, "install", "-g", "--dir", "otro", "--dry-run"),
		argumentos("host-con-dir", mensajeHostConDir, "install", "--dir", "otro", "--host", "claude"),
		argumentos("host-con-dir-vacio", mensajeHostConDir, "install", "--dir", "", "--host", "claude"),
		argumentos("host-desconocido", mensajeHostNoAdmitido, "install", "--host", "codex"),
		argumentos("host-repetido-con-uno-desconocido", mensajeHostNoAdmitido,
			"install", "--host", "claude", "--host", "codex"),
		argumentos("host-con-coma-sin-partir", `"claude,antigravity"`, "install", "--host", "claude,antigravity"),
		argumentos("host-vacio", mensajeHostNoAdmitido, "install", "--host", ""),
		argumentos("host-antes-que-la-skill", mensajeHostNoAdmitido, "install", "desconocida", "--host", "codex"),
		argumentos("skill-desconocida", mensajeSkillDeFuera, "install", "desconocida"),
		argumentos("la-primera-desconocida", mensajeSkillDeFuera, "install", "desconocida", "otra"),
		argumentos("desconocida-tras-una-conocida", mensajeSkillDeFueraB, "install", "beta", "otra"),
		argumentos("skill-antes-que-el-conflicto", mensajeSkillDeFuera, "install", "alfa", "desconocida"),
		argumentos("skill-antes-que-home", mensajeSkillDeFuera, "install", "desconocida", "-g"),
		sinHome("install-g-sin-home", "install", "-g"),
		sinHome("list-g-sin-home", "list", "-g"),
		sinHome("doctor-g-sin-home", "doctor", "--global"),
		{
			nombre: "bandera-desconocida", argv: argvDeSkills("install", "--otra"), codigo: 2,
			clase: schema.ClaseArgumentos, mensaje: "--otra", delKernel: true,
		},
		{
			nombre: "list-no-admite-skills", argv: argvDeSkills("list", "alfa"), codigo: 2,
			clase: schema.ClaseArgumentos, mensaje: "alfa", delKernel: true,
		},
		{
			nombre: "doctor-no-admite-host", argv: argvDeSkills("doctor", "--host", "claude"), codigo: 2,
			clase: schema.ClaseArgumentos, mensaje: "--host", delKernel: true,
		},
	}
}

// comprueba invoca el caso en un proyecto con una carpeta ajena con el nombre de
// alfa y exige su sobre de fallo, su mensaje, en data y en la salida de error, y
// que el disco quede byte a byte como estaba.
func (caso casoDeValidacion) comprueba(t *testing.T) {
	t.Helper()

	enUnProyecto(t, ".claude")
	escribirEnElProyecto(t, ".agents/skills/alfa/mio.md", carpetaAjena)

	arbol := arbolDelProyecto(t, ".")

	// Con --dry-run el kernel descarta el Resultado del applet y el sobre lo
	// firma él, como un rechazo de su analizador.
	procedencia := firmaDeSkills
	if caso.delKernel || slices.Contains(caso.argv, "--dry-run") {
		procedencia = cli.ProcedenciaKernel()
	}

	argv := slices.Concat(caso.argv, []string{"--json"})
	res := invocar(t, delSistema(t), argv...)

	mensaje := falloDeSkills(t, res, caso.clase, caso.codigo, procedencia)

	assert.Contains(t, mensaje, caso.mensaje, "%q", argv)
	assert.NotContains(t, res.errores, "carpeta ajena", "%q: ningún conflicto llega a comprobarse", argv)
	assert.Equal(t, arbol, arbolDelProyecto(t, "."), "%q no lee ni escribe nada", argv)
}

// casoDeDefecto son unas dependencias con las que el applet se compone y que no
// sirven: un defecto de composición, que no puede darse en el binario publicado.
type casoDeDefecto struct {
	nombre       string
	dependencias DependenciasDeSkills
}

// casosDeDefectoDeComposicion son lo empotrado que no se puede leer, lo empotrado
// que falta y el creador de enlaces que falta.
func casosDeDefectoDeComposicion() []casoDeDefecto {
	sinSkills := dependenciasDePrueba(disco.Enlazador{})
	sinSkills.Skills = fstest.MapFS{}

	sinEmpotrado := dependenciasDePrueba(disco.Enlazador{})
	sinEmpotrado.Skills = nil

	return []casoDeDefecto{
		{nombre: "empotrado-que-no-se-lee", dependencias: sinSkills},
		{nombre: "sin-empotrado", dependencias: sinEmpotrado},
		{nombre: "sin-enlazador", dependencias: dependenciasDePrueba(nil)},
	}
}

// compruebaDefectoDeSkills exige que la ayuda y --describe respondan con unas
// dependencias que no sirven, y que ejecutar cualquier verbo sea el defecto de
// composición —código 1, clase inesperado y firma del kernel—, dos veces, sin
// ningún pánico y sin tocar el disco.
func compruebaDefectoDeSkills(t *testing.T, dependencias DependenciasDeSkills) {
	t.Helper()

	enUnProyecto(t, ".claude")

	registro := registroDeSkills(t, dependencias)

	for _, argv := range [][]string{
		argvDeSkills("--help"),
		argvDeSkills("install", "--help"),
		argvDeSkills("install", "--describe"),
		argvDeSkills("doctor", "--describe"),
	} {
		res := invocar(t, registro, argv...)

		assert.Equal(t, 0, res.codigo, "%q no lee lo empotrado: %s", argv, res.errores)
		assert.NotEmpty(t, res.salida, "%q responde", argv)
	}

	for _, verbo := range []string{"install", "install", "list", "doctor"} {
		mensaje := falloDeSkills(t, invocar(t, registro, argvDeSkills(verbo, "--json")...),
			schema.ClaseInesperado, 1, cli.ProcedenciaKernel())
		assert.True(t, strings.HasPrefix(mensaje, "skills: "), mensaje)
	}

	assert.Equal(t, map[string]string{".": "directorio", ".claude": "directorio"}, arbolDelProyecto(t, "."))
}

// TestEnsayoDeInstall fija la línea que --dry-run describe por skill, con cada
// entrada de host, sobre salidas de install escritas a mano: con varios enlaces,
// que el contrato de hoy no da, cada uno va con su modo, en su orden.
func TestEnsayoDeInstall(t *testing.T) {
	t.Parallel()

	skills := []instalacion.SkillInstalada{
		instaladaSinEnlaces(".agents/skills", "alfa", instalacion.EstadoActualizada),
		{
			Nombre: "beta", Ruta: "/home/x/.agents/skills/beta", Estado: instalacion.EstadoSinCambios,
			Enlaces: []instalacion.Enlace{
				{Host: "claude", Ruta: "/home/x/.claude/skills/beta", Modo: instalacion.ModoCopia},
				{Host: "claude", Ruta: "/home/x/otro/beta", Modo: instalacion.ModoEnlace},
			},
		},
	}

	assert.Equal(t, []string{
		"instalar alfa en .agents/skills/alfa: actualizada",
		"instalar beta en ~/.agents/skills/beta: sin cambios; Claude Code ~/.claude/skills/beta (copia);" +
			" Claude Code ~/otro/beta (enlace)",
	}, ensayoDeInstall(skills, "/home/x"), "con HOME, las rutas que cuelgan de él se abrevian")
	assert.Equal(t, []string{
		"instalar alfa en .agents/skills/alfa: actualizada",
		"instalar beta en /home/x/.agents/skills/beta: sin cambios; Claude Code /home/x/.claude/skills/beta (copia);" +
			" Claude Code /home/x/otro/beta (enlace)",
	}, ensayoDeInstall(skills, ""), "sin HOME, tal cual")
	assert.Empty(t, ensayoDeInstall(nil, "/home/x"))
}

// TestSalidaDeSkillsContraSchemas es el punto 4 de la Definition of Done sobre
// el applet skills (FR-053, FR-061, FR-067; contracts/applet-skills.md §4): el
// sobre real que emite el kernel con --json, sobre el registro local, valida
// contra la parte de su verbo leída de schemas/instalacion.json, y no contra lo
// que emite --describe mientras se ejecuta el test. Lo hace toda salida correcta
// de install, list y doctor —la de doctor, sin hallazgos—: cada estado y cada
// modo de install, la lista de enlaces vacía, los tres ámbitos, la skill
// declarada que el binario ya no lleva y, sin manifiesto, la de list y la de
// doctor con la versión nula y la lista vacía, que el contrato tiene que admitir.
// La validación restringe: el mismo sobre con una clave de más o de menos en su
// data no valida.
//
// No es paralelo, por lo mismo que TestAppletSkills: el ámbito local es el
// directorio de trabajo, que cada subtest cambia, y el global sale de HOME, que
// cada uno fija. Los esquemas se compilan antes de cambiar de directorio, porque
// schemas/ se alcanza desde el del paquete.
func TestSalidaDeSkillsContraSchemas(t *testing.T) {
	t.Setenv(cli.VariableNivel, "")

	esquemas := map[string]*jsonschema.Schema{}

	for _, verbo := range []string{"install", "list", "doctor"} {
		publicado, id := ficheroPublicadoDelVerbo(t, verbo)
		require.Equal(t, raizDeLosEsquemas+"instalacion.json", id, "la parte de %q la publica instalacion.json", verbo)

		esquemas[verbo] = salidaPublicada(t, publicado, id, verbo)
	}

	for _, caso := range salidasDeSkills() {
		t.Run(caso.nombre, func(t *testing.T) {
			t.Setenv(variableHome, t.TempDir())

			res := invocar(t, caso.prepara(t), argvDeSkills(slices.Concat(caso.argumentos, []string{"--json"})...)...)

			require.Equal(t, 0, res.codigo, res.errores)
			assert.Equal(t, true, sobreDelJSON(t, res.salida)["ok"],
				"ok decide la rama del esquema contra la que se valida data")

			for _, rasgo := range caso.rasgos {
				assert.Contains(t, res.salida, rasgo, "la salida es la que nombra el caso")
			}

			exigirSalidaPublicada(t, esquemas[caso.argumentos[0]], res.salida)
		})
	}
}

// salidaDeSkills es una invocación correcta de TestSalidaDeSkillsContraSchemas:
// lo que se prepara antes, los argumentos sin --json, el primero de ellos el
// verbo, y los fragmentos que su salida tiene que llevar para que el caso sea el
// que dice, y no otra salida que también valide.
type salidaDeSkills struct {
	nombre     string
	prepara    preparacionDeSkills
	argumentos []string
	rasgos     []string
}

// preparacionDeSkills deja el ámbito de un caso, en un proyecto nuevo que pasa a
// ser el directorio de trabajo, y devuelve el registro de la invocación cuya
// salida se valida.
type preparacionDeSkills func(t *testing.T) *Registro

// salidasDeSkills son las invocaciones correctas cuyo sobre se valida.
func salidasDeSkills() []salidaDeSkills {
	sistema := dependenciasDePrueba(disco.Enlazador{})
	sinEnlaces := dependenciasDePrueba(enlazadorQueFalla{})

	otraVersion := dependenciasDePrueba(disco.Enlazador{})
	otraVersion.Version = "v0.2.0"

	soloAlfa := dependenciasDePrueba(disco.Enlazador{})
	soloAlfa.Skills = sinLaSkill(skillsDePrueba(), "beta")

	enlazada := `"modo":"enlace"`
	copiada := `"modo":"copia"`
	absoluta := `"ruta":"/`
	conVersion := `"version":"` + versionDeSkills + `"`
	sinHallazgos := `"hallazgos":[]`
	sinManifiesto := []string{`"manifiesto":false`, `"version":null`}

	return []salidaDeSkills{
		{
			"install-instalada-y-enlazada", sinInstalar(".claude"),
			[]string{"install"},
			[]string{`"estado":"instalada"`, enlazada},
		},
		{
			"install-sin-cambios", instaladoCon(sistema, sistema, nil, ".claude"),
			[]string{"install"},
			[]string{`"estado":"sin cambios"`, enlazada},
		},
		{
			"install-actualizada", instaladoCon(sistema, otraVersion, nil, ".claude"),
			[]string{"install"},
			[]string{`"estado":"actualizada"`},
		},
		{
			"install-sin-hosts-con-enlaces-vacios", sinInstalar(),
			[]string{"install", "beta"},
			[]string{`"nombre":"beta"`, `"enlaces":[]`},
		},
		{
			"install-con-copia", instaladoCon(sinEnlaces, sinEnlaces, []string{"beta"}, ".claude"),
			[]string{"install"},
			[]string{`"estado":"instalada"`, `"estado":"sin cambios"`, copiada},
		},
		{"install-global", conClaudeEnHome(sinInstalar()), []string{"install", "-g"}, []string{absoluta, enlazada}},
		{
			"install-dir", sinInstalar(".claude"),
			[]string{"install", "--dir", "mis-skills"},
			[]string{`"ruta":"mis-skills/alfa"`, `"enlaces":[]`},
		},

		{
			"list-sin-manifiesto", sinInstalar(".claude"),
			[]string{"list"},
			slices.Concat(sinManifiesto, []string{`"skills":[]`}),
		},
		{
			"list-tras-install", instaladoCon(sistema, sistema, nil, ".claude"),
			[]string{"list"},
			[]string{`"manifiesto":true`, conVersion, `"empotrada":true`, enlazada},
		},
		{
			"list-con-una-skill-que-ya-no-se-empotra", instaladoCon(sistema, soloAlfa, nil, ".claude"),
			[]string{"list"},
			[]string{`"empotrada":true`, `"empotrada":false`},
		},
		{"list-con-copias", instaladoCon(sinEnlaces, sinEnlaces, nil, ".claude"), []string{"list"}, []string{copiada}},
		{
			"list-global", conClaudeEnHome(instaladoCon(sistema, sistema, []string{"-g"})),
			[]string{"list", "-g"},
			[]string{`"directorio":"/`, absoluta, enlazada},
		},
		{
			"list-dir", instaladoCon(sistema, sistema, []string{"--dir", "mis-skills"}),
			[]string{"list", "--dir", "mis-skills"},
			[]string{`"directorio":"mis-skills"`, `"enlaces":[]`},
		},

		{
			"doctor-sin-manifiesto", sinInstalar(".claude"),
			[]string{"doctor"},
			slices.Concat(sinManifiesto, []string{sinHallazgos}),
		},
		{
			"doctor-tras-install", instaladoCon(sistema, sistema, nil, ".claude"),
			[]string{"doctor"},
			[]string{`"manifiesto":true`, conVersion, sinHallazgos},
		},
		{
			"doctor-con-una-skill-que-ya-no-se-empotra", instaladoCon(sistema, soloAlfa, nil, ".claude"),
			[]string{"doctor"},
			[]string{`"manifiesto":true`, sinHallazgos},
		},
		{
			"doctor-con-copias", instaladoCon(sinEnlaces, sinEnlaces, nil, ".claude"),
			[]string{"doctor"},
			[]string{sinHallazgos},
		},
		{
			"doctor-global", conClaudeEnHome(instaladoCon(sistema, sistema, []string{"-g"})),
			[]string{"doctor", "--global"},
			[]string{`"directorio":"/`, sinHallazgos},
		},
		{
			"doctor-dir", instaladoCon(sistema, sistema, []string{"--dir", "mis-skills"}),
			[]string{"doctor", "--dir", "mis-skills"},
			[]string{`"directorio":"mis-skills"`, sinHallazgos},
		},
	}
}

// sinInstalar es un proyecto con esas carpetas y nada instalado, y la invocación
// va con el creador de enlaces del sistema.
func sinInstalar(carpetas ...string) preparacionDeSkills {
	return func(t *testing.T) *Registro {
		t.Helper()

		enUnProyecto(t, carpetas...)

		return delSistema(t)
	}
}

// instaladoCon es un proyecto con esas carpetas en el que install, con esos
// argumentos y las dependencias de antes, ya terminó con 0; la invocación va con
// las de después.
func instaladoCon(antes, despues DependenciasDeSkills, argumentos []string, carpetas ...string) preparacionDeSkills {
	return func(t *testing.T) *Registro {
		t.Helper()

		enUnProyecto(t, carpetas...)

		argv := argvDeSkills(slices.Concat([]string{"install"}, argumentos)...)

		res := invocar(t, registroDeSkills(t, antes), argv...)
		require.Equal(t, 0, res.codigo, res.errores)

		return registroDeSkills(t, despues)
	}
}

// conClaudeEnHome hace de HOME, antes de la preparación, un directorio nuevo con
// .claude/ dentro: el host del ámbito global existe, así que install -g enlaza
// en él.
func conClaudeEnHome(prepara preparacionDeSkills) preparacionDeSkills {
	return func(t *testing.T) *Registro {
		t.Helper()

		home := t.TempDir()
		t.Setenv(variableHome, home)
		require.NoError(t, os.Mkdir(filepath.Join(home, ".claude"), 0o750))

		return prepara(t)
	}
}

// sinLaSkill es lo empotrado sin la carpeta de esa skill.
func sinLaSkill(empotrado fstest.MapFS, nombre string) fstest.MapFS {
	for ruta := range empotrado {
		if strings.HasPrefix(ruta, "skills/"+nombre+"/") {
			delete(empotrado, ruta)
		}
	}

	return empotrado
}
