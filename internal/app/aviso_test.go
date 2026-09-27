package app

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io/fs"
	"log/slog"
	"strings"
	"testing"
	"testing/fstest"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/jmorenobl/kitlegal/internal/cli"
	"github.com/jmorenobl/kitlegal/internal/core/schema"
	"github.com/jmorenobl/kitlegal/internal/disco"
)

// lineaDeAvisoDePrueba es la que da el avisador sintético. El kernel la escribe
// tal cual, sin interpretarla: qué dice la línea lo decide la composición
// (contracts/aviso.md §4), y cuándo se escribe, el kernel (§1).
const lineaDeAvisoDePrueba = "aviso: la línea del avisador de prueba"

// errAvisoRoto es el fallo de la salida de error al escribir el aviso.
var errAvisoRoto = errors.New("la salida de error no admite el aviso")

// fechaDeConsultaDelAviso es la que declaran los applets de la tabla del aviso,
// para que el sobre no dependa del reloj y la salida estándar de dos
// invocaciones iguales se pueda comparar byte a byte.
var fechaDeConsultaDelAviso = time.Date(2026, time.September, 1, 0, 0, 0, 0, time.UTC)

// normaQueNoResponde es la norma con la que el verbo de la tabla del aviso falla
// como una fuente que no responde, código 4.
const normaQueNoResponde = "sin-respuesta"

// appletDelAviso es un applet de la tabla del aviso: sus verbos, todos con una
// norma obligatoria, y como mucho uno por omisión. Declara lo que declara
// cualquier applet y nada más.
type appletDelAviso struct {
	nombre     string
	verbos     []string
	porOmision string
}

func (a appletDelAviso) Nombre() string { return a.nombre }

func (a appletDelAviso) Descripcion() string { return "applet de la tabla del aviso" }

func (a appletDelAviso) Verbos() []Verbo {
	verbos := make([]Verbo, 0, len(a.verbos))

	for _, nombre := range a.verbos {
		verbos = append(verbos, Verbo{
			Nombre:      nombre,
			Descripcion: "verbo de la tabla del aviso",
			Argumentos:  func() Argumentos { return &argumentosDelAviso{} },
			Salida:      map[string]any{},
			PorOmision:  nombre == a.porOmision,
		})
	}

	return verbos
}

// argumentosDelAviso son los de un verbo de la tabla del aviso: una norma que no
// puede faltar, de modo que su ausencia sea un error de argumentos que detecta el
// propio verbo una vez resuelto (FR-070).
type argumentosDelAviso struct {
	Norma string `arg:"" help:"Norma que se consulta."`
}

// Ejecutar devuelve la norma con una procedencia fechada, describe la petición
// que haría con --dry-run y falla como una fuente que no responde con
// normaQueNoResponde.
func (a *argumentosDelAviso) Ejecutar(
	_ context.Context, ec schema.Contexto, _ *slog.Logger,
) (schema.Resultado, error) {
	resultado := schema.Resultado{
		Procedencia: schema.Procedencia{
			Fuente:        procedenciaDePrueba.Fuente,
			URL:           procedenciaDePrueba.URL,
			FechaConsulta: fechaDeConsultaDelAviso,
		},
		Datos: map[string]any{"norma": a.Norma},
	}

	if ec.DryRun {
		resultado.Ensayo = []string{"GET https://fuente.prueba/" + a.Norma}
	}

	if a.Norma == normaQueNoResponde {
		return resultado, fmt.Errorf("la fuente no ha respondido: %w", cli.ErrFuenteNoDisponible)
	}

	return resultado, nil
}

var (
	_ Applet     = appletDelAviso{}
	_ Argumentos = (*argumentosDelAviso)(nil)
)

// registroDelAviso es el registro de la tabla del aviso con ese avisador: un
// applet sin verbo por omisión, como boe; uno con verbo por omisión; y uno que
// se llama skills, que nunca avisa (FR-074). Con un avisador nulo, el registro
// no tiene aviso.
func registroDelAviso(t *testing.T, avisador Avisador) *Registro {
	t.Helper()

	var registro Registro

	applets := []Applet{
		appletDelAviso{nombre: "consulta", verbos: []string{"leer", "listar"}},
		appletDelAviso{nombre: "resumen", verbos: []string{"leer"}, porOmision: "leer"},
		appletDelAviso{nombre: "skills", verbos: []string{"install", "list", "doctor"}},
	}

	for _, applet := range applets {
		require.NoError(t, registro.Registrar(applet))
	}

	registro.Avisar(avisador)

	return &registro
}

// erroresQueRechazanElAviso es una salida de error que admite todo salvo la
// escritura del aviso, que falla: la de un descriptor que se rompe justo
// entonces. Lo admitido queda en admitido.
type erroresQueRechazanElAviso struct {
	admitido bytes.Buffer
}

func (e *erroresQueRechazanElAviso) Write(p []byte) (int, error) {
	if bytes.HasPrefix(p, []byte(lineaDeAvisoDePrueba)) {
		return 0, errAvisoRoto
	}

	return e.admitido.Write(p)
}

// casoDelAvisoDelKernel es una invocación de la tabla del aviso: si la busca
// (contracts/aviso.md §1) y con qué código termina, con o sin aviso.
type casoDelAvisoDelKernel struct {
	nombre string
	argv   []string
	avisa  bool
	codigo int
}

// casosDelAvisoDelKernel son las invocaciones de FR-070: las que resuelven un
// applet distinto de skills con un verbo, nombrado o por omisión, termine como
// termine, avisan; version, toda ayuda, los fallos anteriores al despacho y
// skills, no.
func casosDelAvisoDelKernel() []casoDelAvisoDelKernel {
	return []casoDelAvisoDelKernel{
		// Avisan: el applet está resuelto con su verbo.
		{"verbo-con-json", []string{"kitlegal", "consulta", "leer", "BOE-1", "--json"}, true, 0},
		{"verbo-en-tabla", []string{"kitlegal", "consulta", "leer", "BOE-1"}, true, 0},
		{"verbo-por-omisión", []string{"kitlegal", "resumen", "BOE-1", "--json"}, true, 0},
		{"por-nombre-de-invocacion", []string{"resumen", "BOE-1", "--json"}, true, 0},
		{"fallo-del-verbo", []string{"kitlegal", "consulta", "leer", normaQueNoResponde, "--json"}, true, 4},
		{"error-de-argumentos-del-verbo", []string{"kitlegal", "consulta", "leer"}, true, 2},
		{"error-de-argumentos-del-verbo-por-omisión", []string{"kitlegal", "resumen"}, true, 2},
		{"argumento-de-mas", []string{"kitlegal", "consulta", "listar", "BOE-1", "BOE-2"}, true, 2},
		{"bandera-desconocida", []string{"kitlegal", "consulta", "leer", "BOE-1", "--no-existe"}, true, 2},
		{"plazo-invalido", []string{"kitlegal", "consulta", "leer", "BOE-1", "--timeout", "0s"}, true, 2},
		{"dry-run", []string{"kitlegal", "consulta", "leer", "BOE-1", "--dry-run"}, true, 0},
		{"dry-run-que-falla", []string{"kitlegal", "consulta", "leer", normaQueNoResponde, "--dry-run"}, true, 4},
		{"describe", []string{"kitlegal", "consulta", "leer", "BOE-1", "--describe"}, true, 0},

		// No avisan: version, la ayuda, lo anterior al despacho y skills.
		{"version", []string{"kitlegal", "version"}, false, 0},
		{"version-con-argumentos", []string{"kitlegal", "version", "extra"}, false, 2},
		{"sin-applet", []string{"kitlegal"}, false, 2},
		{"applet-desconocido", []string{"kitlegal", "noexiste", "leer"}, false, 2},
		{"applet-sin-verbo-ni-verbo-por-omisión", []string{"kitlegal", "consulta"}, false, 2},
		{"ayuda-del-binario", []string{"kitlegal", "--help"}, false, 0},
		{"ayuda-corta-del-binario", []string{"kitlegal", "-h"}, false, 0},
		{"ayuda-del-applet", []string{"kitlegal", "consulta", "--help"}, false, 0},
		{"ayuda-del-applet-con-verbo-por-omisión", []string{"kitlegal", "resumen", "--help"}, false, 0},
		{"ayuda-del-verbo", []string{"kitlegal", "consulta", "leer", "--help"}, false, 0},
		{"ayuda-corta-del-verbo", []string{"kitlegal", "consulta", "leer", "BOE-1", "-h"}, false, 0},
		{"ayuda-del-verbo-con-json", []string{"kitlegal", "consulta", "leer", "--help", "--json"}, false, 0},
		{"ayuda-con-valor-falso", []string{"kitlegal", "consulta", "leer", "BOE-1", "--help=false"}, false, 0},
		{"ayuda-rechazada-por-el-analisis", []string{"kitlegal", "consulta", "leer", "--no-existe", "--help"}, false, 2},
		{"skills-install", []string{"kitlegal", "skills", "install", "BOE-1", "--json"}, false, 0},
		{"skills-list-dry-run", []string{"kitlegal", "skills", "list", "BOE-1", "--dry-run"}, false, 0},
		{"skills-doctor-con-error-de-argumentos", []string{"kitlegal", "skills", "doctor"}, false, 2},
		{"skills-describe", []string{"kitlegal", "skills", "install", "BOE-1", "--describe"}, false, 0},
		{"skills-por-nombre-de-invocacion", []string{"skills", "list", "BOE-1", "--json"}, false, 0},
	}
}

// TestAvisoDelKernel fija dónde escribe el kernel el aviso que la composición
// registra (contracts/aviso.md §1 y §5; research.md D5), con un avisador
// sintético y sin disco:
//
//   - lo llaman, una vez, exactamente las invocaciones de FR-070, y ninguna más;
//   - la salida estándar, byte a byte, y el código de salida son los mismos con y
//     sin aviso, y la salida de error lleva exactamente la línea del aviso
//     delante de lo que llevaría sin él (FR-072, FR-075);
//   - un avisador que no avisa no escribe nada;
//   - y un aviso cuya escritura falla no cambia nada: ni el código, ni la salida
//     estándar, ni el resto de la salida de error (research.md D5).
//
// No es paralelo: el kernel lee KITLEGAL_LOG del entorno del proceso, que se fija
// en nivelSinEventos.
func TestAvisoDelKernel(t *testing.T) {
	t.Setenv(cli.VariableNivel, nivelSinEventos)

	for _, caso := range casosDelAvisoDelKernel() {
		t.Run(caso.nombre, func(t *testing.T) {
			t.Setenv(cli.VariableNivel, nivelSinEventos)

			caso.comprobar(t)
		})
	}

	t.Run("avisador-que-no-avisa", func(t *testing.T) {
		t.Setenv(cli.VariableNivel, nivelSinEventos)

		compruebaAvisadorQueNoAvisa(t)
	})

	t.Run("ayuda-que-no-se-puede-escribir", func(t *testing.T) {
		t.Setenv(cli.VariableNivel, nivelSinEventos)

		compruebaAyudaQueNoSePuedeEscribir(t)
	})
}

// nivelSinEventos es el nivel del registro de eventos con el que la salida de
// error de dos invocaciones iguales se puede comparar byte a byte: calla el
// evento de la invocación, que lleva su duración, y nada de lo que escribe el
// presentador, el aviso incluido, que ningún nivel filtra (research.md D10 de
// H1).
const nivelSinEventos = "error"

// comprobar invoca el caso sin avisador, con uno que avisa y con uno que avisa
// en una salida de error que rechaza el aviso, y compara las tres.
func (caso casoDelAvisoDelKernel) comprobar(t *testing.T) {
	t.Helper()

	sin := invocar(t, registroDelAviso(t, nil), caso.argv...)
	require.Equal(t, caso.codigo, sin.codigo, "la invocación es la que el caso dice: %s", sin.errores)

	llamadas := 0
	avisador := func() (string, bool) {
		llamadas++

		return lineaDeAvisoDePrueba, true
	}

	con := invocar(t, registroDelAviso(t, avisador), caso.argv...)

	assert.Equal(t, sin.codigo, con.codigo, "el aviso no cambia el código de salida (FR-072, FR-075)")
	assert.Equal(t, sin.salida, con.salida, "el aviso no cambia la salida estándar, byte a byte (FR-072)")

	if caso.avisa {
		assert.Equal(t, 1, llamadas, "la invocación busca el aviso, una sola vez (FR-070)")
		assert.Equal(t, lineaDeAvisoDePrueba+"\n"+sin.errores, con.errores,
			"exactamente una línea más en la salida de error, la del aviso, delante de todo lo demás")
	} else {
		assert.Zero(t, llamadas, "la invocación no busca el aviso (FR-070, FR-074)")
		assert.Equal(t, sin.errores, con.errores, "sin aviso, la salida de error no cambia")
	}

	var salida bytes.Buffer

	errores := &erroresQueRechazanElAviso{}
	codigo := Main(caso.argv, registroDelAviso(t, avisador), &salida, errores,
		versionDePrueba, commitDePrueba, fechaDePrueba)

	assert.Equal(t, sin.codigo, codigo, "el fallo al escribir el aviso no cambia el código (research.md D5)")
	assert.Equal(t, sin.salida, salida.String(), "ni la salida estándar")
	assert.Equal(t, sin.errores, errores.admitido.String(), "ni el resto de la salida de error")
}

// compruebaAvisadorQueNoAvisa: el avisador que dice que no hay aviso se llama
// igual, y la invocación queda como sin avisador.
func compruebaAvisadorQueNoAvisa(t *testing.T) {
	t.Helper()

	argv := []string{"kitlegal", "consulta", "leer", "BOE-1", "--json"}
	sin := invocar(t, registroDelAviso(t, nil), argv...)

	llamadas := 0
	con := invocar(t, registroDelAviso(t, func() (string, bool) {
		llamadas++

		return lineaDeAvisoDePrueba, false
	}), argv...)

	assert.Equal(t, 1, llamadas)
	assert.Equal(t, sin, con, "sin aviso que dar, la invocación es la misma")
}

// compruebaAyudaQueNoSePuedeEscribir: la ayuda del verbo pedida con la salida
// estándar rota hace fallar el análisis, con el código 1 de la escritura, y
// sigue siendo una petición de ayuda, que no busca el aviso (FR-070).
func compruebaAyudaQueNoSePuedeEscribir(t *testing.T) {
	t.Helper()

	argv := []string{"kitlegal", "consulta", "leer", "--help"}

	var sin, con strings.Builder

	codigoSin := Main(argv, registroDelAviso(t, nil), escritorRoto{}, &sin,
		versionDePrueba, commitDePrueba, fechaDePrueba)

	llamadas := 0
	codigoCon := Main(argv, registroDelAviso(t, func() (string, bool) {
		llamadas++

		return lineaDeAvisoDePrueba, true
	}), escritorRoto{}, &con, versionDePrueba, commitDePrueba, fechaDePrueba)

	assert.Equal(t, 1, codigoSin, "la ayuda que no se puede escribir es el fallo inesperado")
	assert.Equal(t, codigoSin, codigoCon)
	assert.Zero(t, llamadas, "una petición de ayuda no busca el aviso, aunque el análisis falle")
	assert.Equal(t, sin.String(), con.String())
}

// versionDelAvisoNueva es la versión de un binario posterior al que dejó
// instaladas las skills de prueba, que llevan versionDeSkills.
const versionDelAvisoNueva = "v0.2.0"

// avisoDeOtraVersion es la línea de contracts/aviso.md §4 para las skills de
// prueba instaladas con versionDeSkills y un binario versionDelAvisoNueva, sin
// las banderas del ámbito.
const avisoDeOtraVersion = "aviso: las skills instaladas son de kitlegal " + versionDeSkills +
	" y este binario es kitlegal " + versionDelAvisoNueva + "; ejecuta: kitlegal skills install"

// instalarLasDePrueba instala con esa versión las skills de prueba, con los
// argumentos de install que se le den, en el directorio de trabajo.
func instalarLasDePrueba(t *testing.T, version string, argumentos ...string) {
	t.Helper()

	dependencias := dependenciasDePrueba(disco.Enlazador{})
	dependencias.Version = version

	res := invocar(t, registroDeSkills(t, dependencias), argvDeSkills(append([]string{"install"}, argumentos...)...)...)
	require.Equal(t, 0, res.codigo, res.errores)
}

// avisoDeLasDePrueba es el aviso que compone AvisoDeVersion con las skills de
// prueba empotradas y esa versión del binario.
func avisoDeLasDePrueba(version string) (string, bool) {
	dependencias := dependenciasDePrueba(disco.Enlazador{})
	dependencias.Version = version

	return AvisoDeVersion(dependencias)()
}

// TestAvisoDeVersion comprueba la composición del aviso (contracts/aviso.md;
// research.md D5): el dominio de T009 sobre el disco real, con HOME del entorno
// en cada llamada, la versión y lo empotrado de las dependencias de skills, las
// mismas que instalan:
//
//   - un manifiesto local de otra versión da la línea del contrato, sin -g, y
//     uno global, con -g; sin manifiesto, o con el de la misma versión —también
//     sin la v inicial—, no hay aviso (FR-070, FR-071, FR-077);
//   - la versión de una skill declarada cuenta si el binario la empotra, y no si
//     no la empotra (FR-036, FR-071);
//   - un binario de desarrollo no compara (FR-073);
//   - sin HOME, solo el local (FR-070);
//   - y sin lo empotrado, o con lo empotrado que no se puede leer, no hay aviso:
//     el aviso nunca es un error (data-model §7).
//
// No es paralelo: el ámbito local es el directorio de trabajo, que cada subtest
// cambia con t.Chdir, y el global sale de HOME, que cada uno fija.
func TestAvisoDeVersion(t *testing.T) {
	t.Setenv(cli.VariableNivel, "")

	casos := []struct {
		nombre    string
		comprueba func(t *testing.T)
	}{
		{"sin-manifiesto-no-avisa", compruebaAvisoSinManifiesto},
		{"manifiesto-local-de-otra-version", compruebaAvisoLocal},
		{"misma-version-no-avisa", compruebaAvisoDeLaMismaVersion},
		{"binario-de-desarrollo-no-compara", compruebaAvisoDeDesarrollo},
		{"skill-declarada-de-otra-version", compruebaAvisoDeUnaSkill},
		{"manifiesto-global", compruebaAvisoGlobal},
		{"sin-home", compruebaAvisoSinHome},
		{"sin-lo-empotrado", compruebaAvisoSinLoEmpotrado},
	}

	for _, caso := range casos {
		t.Run(caso.nombre, func(t *testing.T) {
			t.Setenv(variableHome, t.TempDir())

			caso.comprueba(t)
		})
	}
}

func compruebaAvisoSinManifiesto(t *testing.T) {
	t.Helper()

	enUnProyecto(t)

	_, hay := avisoDeLasDePrueba(versionDelAvisoNueva)
	assert.False(t, hay, "sin manifiesto local ni global no hay nada que comparar")
}

func compruebaAvisoLocal(t *testing.T) {
	t.Helper()

	enUnProyecto(t)
	instalarLasDePrueba(t, versionDeSkills)

	linea, hay := avisoDeLasDePrueba(versionDelAvisoNueva)
	require.True(t, hay, "el manifiesto local es de otra versión")
	assert.Equal(t, avisoDeOtraVersion, linea, "la línea de contracts/aviso.md §4, sin -g")
}

func compruebaAvisoDeLaMismaVersion(t *testing.T) {
	t.Helper()

	enUnProyecto(t)
	instalarLasDePrueba(t, versionDeSkills)

	for _, version := range []string{versionDeSkills, strings.TrimPrefix(versionDeSkills, "v")} {
		_, hay := avisoDeLasDePrueba(version)
		assert.False(t, hay, "%s es la versión instalada, con o sin la v inicial (FR-077)", version)
	}
}

// arbolQueCuentaLecturas es lo empotrado que cuenta cada entrada que se abre.
// Solo tiene Open, y no las formas rápidas de fstest.MapFS, para que toda
// lectura pase por él.
type arbolQueCuentaLecturas struct {
	arbol    fs.FS
	lecturas *int
}

func (a arbolQueCuentaLecturas) Open(nombre string) (fs.File, error) {
	*a.lecturas++

	return a.arbol.Open(nombre)
}

// compruebaAvisoDeDesarrollo exige que un binario cuya versión no tiene forma
// SemVer no compare, y que lo decida antes de leer lo empotrado (FR-073;
// AvisoDeVersion): con un manifiesto de otra versión instalado, no avisa y no
// abre ninguna entrada de lo empotrado, que una versión con forma sí lee.
func compruebaAvisoDeDesarrollo(t *testing.T) {
	t.Helper()

	enUnProyecto(t)
	instalarLasDePrueba(t, versionDeSkills)

	for _, version := range []string{"dev", "", "0123abc"} {
		lecturas := 0
		arbol := arbolQueCuentaLecturas{arbol: skillsDePrueba(), lecturas: &lecturas}

		_, hay := AvisoDeVersion(DependenciasDeSkills{Version: version, Skills: arbol})()
		assert.False(t, hay, "%q no tiene forma SemVer: un binario de desarrollo no compara (FR-073)", version)
		assert.Zero(t, lecturas, "%q: un binario de desarrollo no lee lo empotrado", version)
	}

	lecturas := 0
	arbol := arbolQueCuentaLecturas{arbol: skillsDePrueba(), lecturas: &lecturas}

	_, hay := AvisoDeVersion(DependenciasDeSkills{Version: versionDelAvisoNueva, Skills: arbol})()
	assert.True(t, hay, "control: con forma SemVer y otra versión instalada, avisa")
	assert.Positive(t, lecturas, "control: con forma SemVer, lee lo empotrado")
}

func compruebaAvisoDeUnaSkill(t *testing.T) {
	t.Helper()

	enUnProyecto(t)
	instalarLasDePrueba(t, versionDeSkills)
	// alfa pasa a la versión nueva, y con ella el manifiesto; beta sigue
	// declarada con la anterior.
	instalarLasDePrueba(t, versionDelAvisoNueva, "alfa")

	linea, hay := avisoDeLasDePrueba(versionDelAvisoNueva)
	require.True(t, hay, "beta, declarada y empotrada, es de otra versión (FR-071)")
	assert.Equal(t, avisoDeOtraVersion, linea, "la línea nombra la versión de beta")

	// Un binario que empotra solo alfa no compara la versión de beta (FR-036).
	soloAlfa := fstest.MapFS{}

	for ruta, fichero := range skillsDePrueba() {
		if strings.HasPrefix(ruta, "skills/alfa/") {
			soloAlfa[ruta] = fichero
		}
	}

	_, hay = AvisoDeVersion(DependenciasDeSkills{Version: versionDelAvisoNueva, Skills: soloAlfa})()
	assert.False(t, hay, "la versión de una skill que el binario no empotra no se compara")
}

func compruebaAvisoGlobal(t *testing.T) {
	t.Helper()

	enUnProyecto(t)
	instalarLasDePrueba(t, versionDeSkills, "-g")

	// En otro directorio de trabajo, sin nada local, cuenta el global.
	enUnProyecto(t)

	linea, hay := avisoDeLasDePrueba(versionDelAvisoNueva)
	require.True(t, hay, "el manifiesto global es de otra versión")
	assert.Equal(t, avisoDeOtraVersion+" -g", linea, "la orden que lo arregla es la del ámbito global")
}

func compruebaAvisoSinHome(t *testing.T) {
	t.Helper()

	enUnProyecto(t)
	instalarLasDePrueba(t, versionDeSkills, "-g")
	t.Setenv(variableHome, "")

	// Sin HOME, el global no se busca: sin nada local, no hay efecto.
	enUnProyecto(t)

	_, hay := avisoDeLasDePrueba(versionDelAvisoNueva)
	assert.False(t, hay, "sin HOME no se busca el manifiesto global (FR-070)")

	// El local se compara igual, y la línea no lleva -g.
	instalarLasDePrueba(t, versionDeSkills)

	linea, hay := avisoDeLasDePrueba(versionDelAvisoNueva)
	require.True(t, hay, "sin HOME, el manifiesto local se compara como siempre")
	assert.Equal(t, avisoDeOtraVersion, linea)
}

func compruebaAvisoSinLoEmpotrado(t *testing.T) {
	t.Helper()

	enUnProyecto(t)
	instalarLasDePrueba(t, versionDeSkills)

	_, hay := AvisoDeVersion(DependenciasDeSkills{Version: versionDelAvisoNueva})()
	assert.False(t, hay, "sin lo empotrado no hay con qué comparar, y no es un error")

	ilegible := arbolQueFallaEn{FS: skillsDePrueba(), ruta: "skills/beta"}
	_, hay = AvisoDeVersion(DependenciasDeSkills{Version: versionDelAvisoNueva, Skills: ilegible})()
	assert.False(t, hay, "lo empotrado que no se puede leer no da aviso, y no es un error")
}
