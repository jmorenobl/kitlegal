package evals

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path"
	"path/filepath"
	"regexp"
	"slices"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// Lo que fijan los casos de TestArnesDelLaboratorio: la versión que contesta el
// sustituto de claude a --version, el modelo completo que se pide y los
// ficheros de las cuatro evals del laboratorio, en su orden
// (contracts/laboratorio.md §6 de H39).
const (
	versionDelClaudeDelLaboratorio = "2.1.317"
	modeloDelLaboratorio           = "claude-sonnet-5-5"
)

var evalsDelLaboratorioDelTest = []string{
	"01-constitucion-articulo-140.yaml", "02-lgt-prescripcion.yaml", "03-lrjsp-principio-de-legalidad.yaml",
	"04-et-vacaciones.yaml",
}

// guionDelClaudeDelLaboratorio es el claude que TestArnesDelLaboratorio pone
// delante en el PATH de quien lanza el arnés, en lugar del de verdad, que
// abriría una sesión con modelo: con variableDeCalentar termina con 0 sin hacer
// nada; con --version, y solo con él, anota la llamada y contesta lo que le
// digan, que de ordinario es la versión como la escribe Claude Code
// (respuestaConLaVersion; research V22 y S1 de H39); y con cualquier otra cosa,
// que es una sesión que abre el guion de la sesión en trabajo/, anota la ruta
// física del directorio de la sesión y escribe el transcript que el test deja
// para ella, si lo hay, en <transcripts>/<variante>/<modo>/<sesión>.jsonl. Las
// sesiones no ven las variables del test: las rutas van escritas en el guion, y
// cada anotación es un fichero propio, de modo que dos sesiones a la vez no se
// pisan. Los tres %s son las anotaciones, entre comillas simples, las órdenes de
// sh de la respuesta a --version, que terminan con exit, y los transcripts,
// entre comillas simples.
const guionDelClaudeDelLaboratorio = `#!/bin/sh
if [ "${` + variableDeCalentar + `:-}" = si ]; then exit 0; fi
set -eu
anotaciones=%s
if [ "$#" -eq 1 ] && [ "$1" = --version ]; then
	: > "$(mktemp "$anotaciones/version.XXXXXX")"
	%s
fi
sesion=$(cd .. && pwd -P)
printf '%%s\n' "$sesion" > "$(mktemp "$anotaciones/sesion.XXXXXX")"
relativa=$(printf '%%s\n' "$sesion" | awk -F/ '{ print $(NF-2) "/" $(NF-1) "/" $NF }')
transcript=%s/"$relativa".jsonl
if [ -f "$transcript" ]; then cat "$transcript"; fi
`

// claudeDelLaboratorio es dónde está el sustituto de claude de un caso de
// TestArnesDelLaboratorio: el directorio que va delante en el PATH, el de sus
// anotaciones y el de los transcripts que escribe.
type claudeDelLaboratorio struct {
	bin         string
	anotaciones string
	transcripts string
}

// respuestaConLaVersion es la respuesta de ordinario del sustituto de claude a
// --version: la versión del test como la escribe Claude Code, y 0.
var respuestaConLaVersion = fmt.Sprintf("printf '%%s (Claude Code)\\n' %s; exit 0",
	entreComillasSimples(versionDelClaudeDelLaboratorio))

// escribirElClaudeDelLaboratorio escribe el sustituto de claude, ejecutable, en
// un directorio temporal del test, a través de un os.Root, con la respuesta
// dada a --version o, vacía, respuestaConLaVersion, y lo ejecuta una vez con
// variableDeCalentar, como escribirSustitutos.
func escribirElClaudeDelLaboratorio(t *testing.T, respuesta string) claudeDelLaboratorio {
	t.Helper()

	if respuesta == "" {
		respuesta = respuestaConLaVersion
	}

	claude := claudeDelLaboratorio{bin: t.TempDir(), anotaciones: t.TempDir(), transcripts: t.TempDir()}

	raiz, err := os.OpenRoot(claude.bin)
	require.NoError(t, err)

	defer func() { require.NoError(t, raiz.Close()) }()

	guion := fmt.Sprintf(guionDelClaudeDelLaboratorio, entreComillasSimples(claude.anotaciones), respuesta,
		entreComillasSimples(claude.transcripts))
	require.NoError(t, escribirEjecutable(raiz, sustitutoClaude, guion))
	calentar(t, filepath.Join(claude.bin, sustitutoClaude))

	return claude
}

// anotadas son las anotaciones del sustituto con ese prefijo: una por llamada.
func (c claudeDelLaboratorio) anotadas(t *testing.T, prefijo string) []string {
	t.Helper()

	entradas, err := os.ReadDir(c.anotaciones)
	require.NoError(t, err)

	var anotadas []string

	for _, entrada := range entradas {
		if strings.HasPrefix(entrada.Name(), prefijo+".") {
			anotadas = append(anotadas, filepath.Join(c.anotaciones, entrada.Name()))
		}
	}

	return anotadas
}

// sesiones son las rutas físicas de las carpetas de las sesiones que abrió el
// sustituto, en orden de ruta: una por sesión abierta, también si se abre dos
// veces.
func (c claudeDelLaboratorio) sesiones(t *testing.T) []string {
	t.Helper()

	var sesiones []string

	for _, anotada := range c.anotadas(t, "sesion") {
		contenido, err := leerFichero(anotada)
		require.NoError(t, err)

		sesiones = append(sesiones, strings.TrimSuffix(string(contenido), "\n"))
	}

	slices.Sort(sesiones)

	return sesiones
}

// versiones es cuántas veces se le preguntó la versión al sustituto.
func (c claudeDelLaboratorio) versiones(t *testing.T) int {
	t.Helper()

	return len(c.anotadas(t, "version"))
}

// escribirTranscript deja el transcript que el sustituto escribe en la sesión de
// esa ruta, <variante>/<modo>/<sesión>.
func (c claudeDelLaboratorio) escribirTranscript(t *testing.T, sesion, contenido string) {
	t.Helper()

	ruta := filepath.Join(c.transcripts, filepath.FromSlash(sesion)+".jsonl")
	require.NoError(t, os.MkdirAll(filepath.Dir(ruta), 0o700))
	require.NoError(t, os.WriteFile(ruta, []byte(contenido), 0o600))
}

// arnesDelTest es un sondeo del laboratorio de TestArnesDelLaboratorio con lo
// que el test mira después: su sustituto de claude y si se construyeron
// kitlegal y laboratorio.
type arnesDelTest struct {
	sondeo     SondeoDelLaboratorio
	claude     claudeDelLaboratorio
	construido *bool
}

// argumentosValidos son unos argumentos que valen, con el DIRECTORIO dado, que
// no existe.
func argumentosValidos(directorio string) ArgumentosDelLaboratorio {
	return ArgumentosDelLaboratorio{
		Variantes: "control", Modos: "orden", Sesiones: "4", Modelo: modeloDelLaboratorio,
		ClaudeCode: versionDelClaudeDelLaboratorio, Concurrencia: "2", Directorio: directorio,
	}
}

// nuevoArnesDelTest arma el sondeo del laboratorio de un caso: el sustituto de
// claude delante en el PATH de la base, que es la del proceso sin sus
// credenciales de Claude Code, con un HOME y un TMPDIR propios y la credencial
// del sondeo; el laboratorio del repositorio; un temporal vacío; el guion de la
// sesión; y una construcción que no construye nada: deja un kitlegal y un
// laboratorio que no hacen nada en bin/ y en laboratorio-bin/ del temporal, y
// anota que se llamó. Los argumentos son los que da el ajuste sobre
// argumentosValidos, con DIRECTORIO dentro de un directorio temporal del test, y
// la respuesta del sustituto a --version, la dada (escribirElClaudeDelLaboratorio).
func nuevoArnesDelTest(t *testing.T, ajustar func(*ArgumentosDelLaboratorio), respuesta string) arnesDelTest {
	t.Helper()

	guion, err := filepath.Abs(guionDeLaSesion)
	require.NoError(t, err)

	arnes := arnesDelTest{claude: escribirElClaudeDelLaboratorio(t, respuesta), construido: new(bool)}

	argumentos := argumentosValidos(filepath.Join(t.TempDir(), "sondeos", "h39-a"))
	if ajustar != nil {
		ajustar(&argumentos)
	}

	arnes.sondeo = SondeoDelLaboratorio{
		Argumentos: argumentos,
		Entorno: sobreLaBase(os.Environ(), []string{
			"PATH=" + arnes.claude.bin + string(os.PathListSeparator) + os.Getenv("PATH"),
			"HOME=" + t.TempDir(), "TMPDIR=" + t.TempDir(), variableDeLaSuscripcion + "=" + valorDeLaSuscripcion,
		}, accesosDeClaudeCode...),
		Raiz:                  raizDelRepositorio,
		Temporal:              t.TempDir(),
		Guion:                 guion,
		ConstruirLosProgramas: programasSinConstruir(arnes.construido),
	}

	return arnes
}

// programasSinConstruir es la construcción de kitlegal y laboratorio de
// TestArnesDelLaboratorio, que no construye nada: deja, a través de un os.Root,
// un kitlegal y un laboratorio que no hacen nada donde los deja go install, y
// anota que se llamó.
func programasSinConstruir(construido *bool) func(raiz, temporal string, base []string) error {
	return func(_, temporal string, _ []string) (err error) {
		*construido = true

		raiz, err := os.OpenRoot(temporal)
		if err != nil {
			return err
		}

		defer func() { err = errors.Join(err, raiz.Close()) }()

		return errors.Join(
			raiz.Mkdir(carpetaDeBinarios, 0o700),
			escribirEjecutable(raiz, filepath.Join(carpetaDeBinarios, programaDeLasConsultas), kitlegalQueNoHaceNada),
			raiz.Mkdir(carpetaDelProgramaDelLaboratorio, 0o700),
			escribirEjecutable(raiz, filepath.Join(carpetaDelProgramaDelLaboratorio, programaDelLaboratorio),
				kitlegalQueNoHaceNada),
		)
	}
}

// casoDeUsoDelArnes es un caso de TestArnesDelLaboratorio en el que el arnés no
// abre nada: el ajuste de sus argumentos y de su sondeo, la respuesta del
// sustituto de claude a --version (escribirElClaudeDelLaboratorio), el mensaje de
// su error de uso, y si llega a preguntar la versión de Claude Code.
type casoDeUsoDelArnes struct {
	nombre     string
	argumentos func(t *testing.T, a *ArgumentosDelLaboratorio)
	sondeo     func(t *testing.T, s *SondeoDelLaboratorio)
	respuesta  string
	error      func(a ArgumentosDelLaboratorio) string
	version    bool
}

// mensajeFijo es el mensaje de error de un caso de uso que no depende de sus
// argumentos.
func mensajeFijo(mensaje string) func(ArgumentosDelLaboratorio) string {
	return func(ArgumentosDelLaboratorio) string { return mensaje }
}

// casosDeUsoDelArnes son los casos de TestArnesDelLaboratorio que son un error de
// uso (contracts/arnes.md §2 de H39; FR-030, FR-031; SC-007).
func casosDeUsoDelArnes() []casoDeUsoDelArnes {
	fijo := mensajeFijo

	return []casoDeUsoDelArnes{
		{
			nombre: "con-la-version-distinta",
			argumentos: func(_ *testing.T, a *ArgumentosDelLaboratorio) {
				a.ClaudeCode = "2.1.284"
			},
			error:   fijo("CLAUDE_CODE: la versión de Claude Code instalada es 2.1.317 y se pidió 2.1.284"),
			version: true,
		},
		{
			nombre:     "con-un-alias-del-modelo",
			argumentos: func(_ *testing.T, a *ArgumentosDelLaboratorio) { a.Modelo = "sonnet" },
			error: fijo("MODELO: «sonnet» no es un identificador completo de modelo " +
				"(claude-<familia>-<versión>, como claude-sonnet-5-5)"),
		},
		{
			nombre: "con-una-variante-en-un-modo-que-no-es-suyo",
			argumentos: func(_ *testing.T, a *ArgumentosDelLaboratorio) {
				a.Variantes, a.Modos = "instrucciones", "orden"
			},
			error: fijo("VARIANTES y MODOS: instrucciones no se mide en modo orden"),
		},
		{
			nombre:     "con-sesiones-que-no-son-multiplo-de-4",
			argumentos: func(_ *testing.T, a *ArgumentosDelLaboratorio) { a.Sesiones = "6" },
			error:      fijo("SESIONES: «6» no es un múltiplo de 4 mayor que 0"),
		},
		{
			nombre:     "todos-los-argumentos-juntos",
			argumentos: todosLosArgumentosMal,
			error: func(a ArgumentosDelLaboratorio) string {
				return strings.Join([]string{
					"VARIANTES: «nada» no es ninguna variante del laboratorio",
					"VARIANTES: «control» se repite",
					"MODOS: «todo» no es orden ni herramienta",
					"VARIANTES y MODOS: parametro no se mide en modo orden",
					"SESIONES: «0» no es un múltiplo de 4 mayor que 0",
					"MODELO: «» no es un identificador completo de modelo (claude-<familia>-<versión>, como " +
						"claude-sonnet-5-5)",
					"CLAUDE_CODE: está vacía",
					"CONCURRENCIA: «0» no es un entero mayor o igual que 1",
					"DIRECTORIO: " + a.Directorio + " ya existe",
				}, "\n")
			},
		},
		{
			nombre:     "con-el-directorio-vacio",
			argumentos: func(_ *testing.T, a *ArgumentosDelLaboratorio) { a.Directorio = "" },
			error:      fijo("DIRECTORIO: está vacío"),
		},
		{
			nombre: "sin-la-credencial",
			sondeo: func(_ *testing.T, s *SondeoDelLaboratorio) {
				s.Entorno = sobreLaBase(s.Entorno, []string{variableDeLaSuscripcion + "="})
			},
			error: fijo(errSinSuscripcion.Error()),
		},
		{
			nombre: "sin-claude-en-el-path",
			sondeo: func(t *testing.T, s *SondeoDelLaboratorio) {
				t.Helper()

				s.Entorno = sobreLaBase(s.Entorno, []string{variableDelPATH + "=" + t.TempDir()})
			},
			error: fijo("CLAUDE_CODE: claude --version no se pudo ejecutar: claude no está en ninguna carpeta del PATH"),
		},
		{
			nombre:    "con-un-claude-que-falla",
			respuesta: "printf 'no arranca\\n' >&2; exit 3",
			error:     fijo("CLAUDE_CODE: claude --version no se pudo ejecutar: exit status 3: no arranca"),
			version:   true,
		},
		{
			nombre:    "con-un-claude-que-no-escribe-nada",
			respuesta: "exit 0",
			error:     fijo("CLAUDE_CODE: claude --version no se pudo ejecutar: no escribió ninguna versión"),
			version:   true,
		},
		{
			nombre:     "con-un-directorio-que-no-se-puede-mirar",
			argumentos: directorioBajoUnFichero,
			error: func(a ArgumentosDelLaboratorio) string {
				_, err := os.Lstat(a.Directorio)

				return "DIRECTORIO: " + a.Directorio + " no se puede mirar: " + err.Error()
			},
		},
	}
}

// directorioBajoUnFichero da a DIRECTORIO una ruta dentro de un fichero, que no
// se puede mirar: ni existe ni deja de existir.
func directorioBajoUnFichero(t *testing.T, a *ArgumentosDelLaboratorio) {
	t.Helper()

	fichero := filepath.Join(t.TempDir(), "fichero")
	require.NoError(t, os.WriteFile(fichero, nil, 0o600))

	a.Directorio = filepath.Join(fichero, "h39-a")
}

// todosLosArgumentosMal da a cada argumento un valor que no vale: una variante
// que no existe y otra repetida, un modo que no existe, una variante que no se
// mide en el modo pedido, y un DIRECTORIO que ya existe.
func todosLosArgumentosMal(t *testing.T, a *ArgumentosDelLaboratorio) {
	t.Helper()

	a.Variantes, a.Modos, a.Sesiones, a.Modelo = "control,nada,control,parametro", "orden,todo", "0", ""
	a.ClaudeCode, a.Concurrencia, a.Directorio = "", "0", t.TempDir()
}

// casoValidoDelArnes es un caso de TestArnesDelLaboratorio en el que el arnés
// abre sesiones: sus argumentos, la variante y el modo de cada reparto, en el
// orden en que se abren, las sesiones que se abren de cada eval y el número de
// las que quedan sin abrir tras el límite de uso; y los transcripts que escribe
// el sustituto, por sesión <variante>/<modo>/<sesión>.
type casoValidoDelArnes struct {
	nombre      string
	argumentos  func(a *ArgumentosDelLaboratorio)
	repartos    []repartoEsperado
	transcripts func(t *testing.T) map[string]string
}

// repartoEsperado es lo que se espera de un reparto: su variante y su modo, las
// sesiones de su plan de cada eval, numeradas desde 1, cuántas abre de las
// primeras del plan y la línea de la salida.
type repartoEsperado struct {
	variante string
	modo     Modo
	porEval  int
	abiertas int
	linea    string
}

// casosValidosDelArnes son los casos de TestArnesDelLaboratorio en los que el
// arnés abre sesiones (contracts/arnes.md §3 y §4 de H39; FR-030, FR-033;
// SC-007).
func casosValidosDelArnes() []casoValidoDelArnes {
	return []casoValidoDelArnes{
		{
			// Dos variantes, una que se mide en un modo y otra en los dos, pedidas
			// al revés de la declaración: van en el orden de la declaración.
			nombre: "dos-variantes-en-el-orden-declarado",
			argumentos: func(a *ArgumentosDelLaboratorio) {
				a.Variantes, a.Modos, a.Sesiones, a.Concurrencia = "instrucciones,control", "herramienta", "8", "3"
			},
			repartos: []repartoEsperado{
				{
					variante: varianteDeControl, modo: ModoHerramienta, porEval: 2, abiertas: 8,
					linea: "- control, herramienta: 8 abiertas.",
				},
				{
					variante: "instrucciones", modo: ModoHerramienta, porEval: 2, abiertas: 8,
					linea: "- instrucciones, herramienta: 8 abiertas.",
				},
			},
		},
		{
			// Una variante en los dos modos, pedidos al revés: orden antes que
			// herramienta, con su skill de entrada.
			nombre: "una-variante-en-los-dos-modos",
			argumentos: func(a *ArgumentosDelLaboratorio) {
				a.Variantes, a.Modos = "entrada", "herramienta,orden"
			},
			repartos: []repartoEsperado{
				{variante: "entrada", modo: ModoOrden, porEval: 1, abiertas: 4, linea: "- entrada, orden: 4 abiertas."},
				{
					variante: "entrada", modo: ModoHerramienta, porEval: 1, abiertas: 4,
					linea: "- entrada, herramienta: 4 abiertas.",
				},
			},
		},
		{
			nombre:     "la-variante-del-parametro",
			argumentos: func(a *ArgumentosDelLaboratorio) { a.Variantes, a.Modos = "parametro", "herramienta" },
			repartos: []repartoEsperado{
				{
					variante: "parametro", modo: ModoHerramienta, porEval: 1, abiertas: 4,
					linea: "- parametro, herramienta: 4 abiertas.",
				},
			},
		},
		{
			// La primera sesión da el mensaje del límite de uso: su reparto no abre
			// ninguna más, y el siguiente, ninguna.
			nombre: "tras-el-limite-de-uso",
			argumentos: func(a *ArgumentosDelLaboratorio) {
				a.Modos, a.Concurrencia = "orden,herramienta", "1"
			},
			repartos: []repartoEsperado{
				{
					variante: varianteDeControl, modo: ModoOrden, porEval: 1, abiertas: 1,
					linea: "- control, orden: 1 abiertas, 3 sin abrir tras el límite de uso.",
				},
				{
					variante: varianteDeControl, modo: ModoHerramienta, porEval: 1, abiertas: 0,
					linea: "- control, herramienta: 0 abiertas, 4 sin abrir tras el límite de uso.",
				},
			},
			transcripts: func(t *testing.T) map[string]string {
				t.Helper()

				primera := path.Join(varianteDeControl, string(ModoOrden),
					sesionesDelReparto(evalsDelLaboratorioDelTest, 1)[0])

				return map[string]string{primera: mensajeInit + mensajeResultConError(t, textoDelLimiteDeSesion)}
			},
		},
	}
}

// TestArnesDelLaboratorio fija el arnés del sondeo del laboratorio sin modelo
// (contracts/arnes.md §2, §3, §4 y §6 de H39; data-model §6 de H39; research D8
// de H39; FR-030, FR-031, FR-033; SC-007), con la construcción de kitlegal y
// laboratorio sustituida y un sustituto de claude en el PATH que contesta a --version y
// anota cada sesión que abre.
//
// Con un argumento que no vale, la credencial que falta, una versión de Claude
// Code distinta de la pedida o un claude que no se puede ejecutar, devuelve un
// error de uso con el mensaje de §2 —los de los argumentos, juntos, uno por
// línea y en su orden—, no pregunta la versión si no valen los argumentos ni la
// credencial, no construye nada, no abre ninguna sesión, no crea DIRECTORIO y
// no escribe nada en el temporal.
//
// Con una petición válida abre SESIONES/4 sesiones de cada eval en cada variante
// y modo pedidos, en DIRECTORIO/<variante>/<modo>/<eval>-<nn>/, y ninguna más;
// en el orden de la declaración y con orden antes que herramienta; cada una con
// solo las skills de la copia de su variante; con servidor.json solo en el modo
// herramienta, con la orden de laboratorio servidor delante de kitlegal solo en
// instrucciones y parametro; deja sondeo.json con el modelo y la versión; y
// devuelve la salida de §4. Tras un reparto con sesiones sin abrir por el
// límite de uso no abre ninguna más.
func TestArnesDelLaboratorio(t *testing.T) {
	t.Parallel()

	for _, caso := range casosDeUsoDelArnes() {
		t.Run(caso.nombre, func(t *testing.T) {
			t.Parallel()

			exigirElErrorDeUsoDelArnes(t, caso)
		})
	}

	for _, caso := range casosValidosDelArnes() {
		t.Run(caso.nombre, func(t *testing.T) {
			t.Parallel()

			exigirElSondeoDelArnes(t, caso, evalsDelLaboratorioDelTest, exigirLaSesionDelArnes)
		})
	}
}

// exigirElErrorDeUsoDelArnes ejecuta el arnés de un caso de uso y exige su error
// de uso y que no haya hecho nada.
func exigirElErrorDeUsoDelArnes(t *testing.T, caso casoDeUsoDelArnes) {
	t.Helper()

	arnes := nuevoArnesDelTest(t, nil, caso.respuesta)
	if caso.argumentos != nil {
		caso.argumentos(t, &arnes.sondeo.Argumentos)
	}

	if caso.sondeo != nil {
		caso.sondeo(t, &arnes.sondeo)
	}

	estado, err := os.Lstat(arnes.sondeo.Argumentos.Directorio)
	directorioExistia := err == nil && estado.IsDir()

	salida, err := sondearElLaboratorio(t.Context().Done(), arnes.sondeo)
	require.EqualError(t, err, caso.error(arnes.sondeo.Argumentos))
	assert.Empty(t, salida)

	var uso *errorDeUso
	require.ErrorAs(t, err, &uso, "el error del caso es un error de uso")

	assert.Empty(t, arnes.claude.sesiones(t), "no se abre ninguna sesión")
	assert.Equal(t, caso.version, arnes.claude.versiones(t) == 1, "la versión se pregunta solo tras los argumentos y la "+
		"credencial, y una vez")
	assert.False(t, *arnes.construido, "no se construye nada")

	entradas, err := os.ReadDir(arnes.sondeo.Temporal)
	require.NoError(t, err)
	assert.Empty(t, entradas, "no se escribe nada en el temporal")

	if directorioExistia {
		entradas, err := os.ReadDir(arnes.sondeo.Argumentos.Directorio)
		require.NoError(t, err)
		assert.Empty(t, entradas, "no se escribe nada en el DIRECTORIO que ya existía")

		return
	}

	if arnes.sondeo.Argumentos.Directorio != "" {
		assert.NoDirExists(t, filepath.Dir(arnes.sondeo.Argumentos.Directorio), "no se crea DIRECTORIO ni su padre")
	}
}

// exigirLaSesionDelReparto exige lo que ve la sesión de ese directorio, de un
// reparto de un caso válido del arnés.
type exigirLaSesionDelReparto func(t *testing.T, arnes arnesDelTest, reparto repartoEsperado, sesion string)

// exigirElSondeoDelArnes ejecuta el arnés de un caso válido y exige sus
// sesiones —las de cada una de las evals dadas, en su orden, en cada reparto, y
// ninguna más—, su sondeo.json, lo que ve cada sesión, con exigirLaSesion, y su
// salida.
func exigirElSondeoDelArnes(t *testing.T, caso casoValidoDelArnes, evals []string,
	exigirLaSesion exigirLaSesionDelReparto,
) {
	t.Helper()

	arnes := nuevoArnesDelTest(t, caso.argumentos, "")

	if caso.transcripts != nil {
		for sesion, transcript := range caso.transcripts(t) {
			arnes.claude.escribirTranscript(t, sesion, transcript)
		}
	}

	salida, err := sondearElLaboratorio(t.Context().Done(), arnes.sondeo)
	require.NoError(t, err)

	directorio := arnes.sondeo.Argumentos.Directorio

	lineas := []string{"Sondeo del laboratorio en " + directorio + ", con claude-sonnet-5-5 y Claude Code 2.1.317."}
	for _, reparto := range caso.repartos {
		lineas = append(lineas, reparto.linea)
	}

	lineas = append(lineas, "La tabla: make laboratorio-tabla DIRECTORIO="+directorio)
	assert.Equal(t, strings.Join(lineas, "\n")+"\n", salida)

	assert.True(t, *arnes.construido, "se construyen kitlegal y laboratorio")
	assert.Equal(t, 1, arnes.claude.versiones(t), "la versión se pregunta una vez")

	contenido, err := leerFichero(filepath.Join(directorio, ficheroDelSondeo))
	require.NoError(t, err)
	assert.JSONEq(t, `{"modelo":"claude-sonnet-5-5","claude_code":"2.1.317"}`, string(contenido))

	leido, err := leerElSondeo(directorio)
	require.NoError(t, err, "el guion que cuenta el sondeo lee su sondeo.json")
	assert.Equal(t, sondeoLeido{Modelo: modeloDelLaboratorio, ClaudeCode: versionDelClaudeDelLaboratorio}, leido)

	fisico, err := filepath.EvalSymlinks(directorio)
	require.NoError(t, err)

	var esperadas []string

	for _, reparto := range caso.repartos {
		plan := sesionesDelReparto(evals, reparto.porEval)
		require.LessOrEqual(t, reparto.abiertas, len(plan), "el caso abre de las sesiones de su plan")

		for _, sesion := range plan[:reparto.abiertas] {
			esperadas = append(esperadas, filepath.Join(fisico, reparto.variante, string(reparto.modo), sesion))
			exigirLaSesion(t, arnes, reparto, filepath.Join(directorio, reparto.variante, string(reparto.modo), sesion))
		}
	}

	slices.Sort(esperadas)
	assert.Equal(t, esperadas, arnes.claude.sesiones(t), "se abren las sesiones de cada eval en cada variante y modo "+
		"pedidos, y ninguna más")
}

// sesionesDelReparto son los nombres de las sesiones que abre un reparto: las
// porEval de cada una de las evals dadas, en su orden, <eval>-<nn>.
func sesionesDelReparto(evals []string, porEval int) []string {
	var sesiones []string

	for _, eval := range evals {
		for numero := 1; numero <= porEval; numero++ {
			sesiones = append(sesiones, fmt.Sprintf("%s-%02d", strings.TrimSuffix(eval, ".yaml"), numero))
		}
	}

	return sesiones
}

// exigirLaSesionDelArnes exige que la sesión de ese directorio vea solo las
// skills de la copia de su variante, y que su servidor.json, solo en el modo
// herramienta, arranque kitlegal mcp serve o, en instrucciones y parametro, el
// filtro de laboratorio servidor con el texto de la variante.
func exigirLaSesionDelArnes(t *testing.T, arnes arnesDelTest, reparto repartoEsperado, sesion string) {
	t.Helper()

	variante := directorioDeLaVariante(raizDelRepositorio, reparto.variante)

	deLaVariante, err := os.ReadDir(filepath.Join(variante, carpetaDeLasSkillsDeLaVariante))
	require.NoError(t, err)

	enlazadas, err := os.ReadDir(filepath.Join(sesion, directorioDeClaude, directorioDeSkills))
	require.NoError(t, err)
	require.Len(t, enlazadas, len(deLaVariante), "la sesión %s ve las skills de su variante", sesion)

	copia := filepath.Join(arnes.sondeo.Temporal, carpetaDeLasSkillsDelSondeo, reparto.variante)

	for posicion, skill := range deLaVariante {
		destino, err := os.Readlink(filepath.Join(sesion, directorioDeClaude, directorioDeSkills, enlazadas[posicion].Name()))
		require.NoError(t, err)
		assert.Equal(t, filepath.Join(copia, skill.Name()), destino, "la sesión %s enlaza la copia de su variante", sesion)

		original, err := leerFichero(filepath.Join(variante, carpetaDeLasSkillsDeLaVariante, skill.Name(),
			ficheroDeLaSkillDeLaVariante))
		require.NoError(t, err)

		copiado, err := leerFichero(filepath.Join(destino, ficheroDeLaSkillDeLaVariante))
		require.NoError(t, err)
		assert.Equal(t, string(original), string(copiado), "la copia de %s es la de su variante", skill.Name())
	}

	servidor := filepath.Join(sesion, ficheroDelServidor)
	if reparto.modo == ModoOrden {
		assert.NoFileExists(t, servidor, "una sesión del modo orden no tiene servidor")

		return
	}

	assert.Equal(t, ordenEsperadaDelServidor(t, arnes.sondeo.Temporal, reparto.variante), ordenDelServidorEscrito(t,
		servidor), "la orden del servidor de %s", sesion)
}

// evalsDeNormativaDelTest son los ficheros de las cuatro evals del sondeo de la
// familia de normativa, de evals/normativa, en su orden (contracts/arnes.md §3
// de H40; data-model §1 de H40).
var evalsDeNormativaDelTest = []string{
	"04-lgt-prescripcion.yaml", "05-trlrhl-impuestos-municipales.yaml", "15-irpf-rendimientos-por-materia.yaml",
	"22-lcsp-contrato-menor-en-2019.yaml",
}

// Lo que piden los casos de TestArnesDeNormativa (contracts/arnes.md §1 y §5 de
// H40): la variante de control de normativa, que se mide en los dos modos, y las
// otras tres, que solo se miden en modo herramienta, en el orden de su
// declaración; y la línea del error de una mezcla de las dos familias.
const (
	controlDeNormativaDelTest         = "normativa-control"
	soloHerramientaDeNormativaDelTest = "normativa-remisiones-en-orden,normativa-cuerpo-corto," +
		"normativa-testigo-en-referencia"
	mezclaDeLasFamiliasDelTest = "VARIANTES: se mezclan variantes del laboratorio de H39 y variantes de normativa"
)

// casosDeUsoDeNormativa son los casos de TestArnesDeNormativa que son un error
// de uso (contracts/arnes.md §2 y §5 de H40; FR-010, FR-011; SC-006): cada uno,
// con su mensaje y nada hecho.
func casosDeUsoDeNormativa() []casoDeUsoDelArnes {
	return []casoDeUsoDelArnes{
		{
			nombre:     "con-una-variante-que-no-es-de-ninguna-familia",
			argumentos: func(_ *testing.T, a *ArgumentosDelLaboratorio) { a.Variantes = "normativa-otra" },
			error:      mensajeFijo("VARIANTES: «normativa-otra» no es ninguna variante del laboratorio"),
		},
		{
			// Los tres pares que no se miden, una línea por variante, en el orden de
			// su declaración.
			nombre: "las-tres-de-solo-herramienta-en-modo-orden",
			argumentos: func(_ *testing.T, a *ArgumentosDelLaboratorio) {
				a.Variantes, a.Modos = soloHerramientaDeNormativaDelTest, "orden"
			},
			error: mensajeFijo(strings.Join([]string{
				"VARIANTES y MODOS: normativa-remisiones-en-orden no se mide en modo orden",
				"VARIANTES y MODOS: normativa-cuerpo-corto no se mide en modo orden",
				"VARIANTES y MODOS: normativa-testigo-en-referencia no se mide en modo orden",
			}, "\n")),
		},
		{
			nombre: "con-sesiones-que-no-son-multiplo-de-4",
			argumentos: func(_ *testing.T, a *ArgumentosDelLaboratorio) {
				a.Variantes, a.Modos, a.Sesiones = controlDeNormativaDelTest, "orden,herramienta", "6"
			},
			error: mensajeFijo("SESIONES: «6» no es un múltiplo de 4 mayor que 0"),
		},
		{
			nombre: "con-las-dos-familias",
			argumentos: func(_ *testing.T, a *ArgumentosDelLaboratorio) {
				a.Variantes = "control," + controlDeNormativaDelTest
			},
			error: mensajeFijo(mezclaDeLasFamiliasDelTest),
		},
		{
			// Con las dos familias, los pares de cada una se comprueban con su
			// declaración, primero los de H39, aunque se pidan detrás.
			nombre: "con-las-dos-familias-y-sus-pares",
			argumentos: func(_ *testing.T, a *ArgumentosDelLaboratorio) {
				a.Variantes = "normativa-cuerpo-corto,instrucciones"
			},
			error: mensajeFijo(strings.Join([]string{
				mezclaDeLasFamiliasDelTest,
				"VARIANTES y MODOS: instrucciones no se mide en modo orden",
				"VARIANTES y MODOS: normativa-cuerpo-corto no se mide en modo orden",
			}, "\n")),
		},
		{
			// El ejemplo de contracts/arnes.md §2 de H40: la mezcla, tras las líneas
			// de los nombres y antes de las de los pares.
			nombre: "el-ejemplo-del-contrato",
			argumentos: func(_ *testing.T, a *ArgumentosDelLaboratorio) {
				a.Variantes, a.Sesiones = "control,normativa-cuerpo-corto,normativa-otra", "6"
			},
			error: mensajeFijo(strings.Join([]string{
				"VARIANTES: «normativa-otra» no es ninguna variante del laboratorio",
				mezclaDeLasFamiliasDelTest,
				"VARIANTES y MODOS: normativa-cuerpo-corto no se mide en modo orden",
				"SESIONES: «6» no es un múltiplo de 4 mayor que 0",
			}, "\n")),
		},
		{
			nombre: "con-la-version-distinta",
			argumentos: func(_ *testing.T, a *ArgumentosDelLaboratorio) {
				a.Variantes, a.ClaudeCode = controlDeNormativaDelTest, "2.1.284"
			},
			error:   mensajeFijo("CLAUDE_CODE: la versión de Claude Code instalada es 2.1.317 y se pidió 2.1.284"),
			version: true,
		},
		{
			nombre: "con-un-alias-del-modelo",
			argumentos: func(_ *testing.T, a *ArgumentosDelLaboratorio) {
				a.Variantes, a.Modelo = controlDeNormativaDelTest, "sonnet"
			},
			error: mensajeFijo("MODELO: «sonnet» no es un identificador completo de modelo " +
				"(claude-<familia>-<versión>, como claude-sonnet-5-5)"),
		},
	}
}

// casosValidosDeNormativa son los casos de TestArnesDeNormativa en los que el
// arnés abre sesiones: los dos lanzamientos de la aceptación con menos sesiones
// (contracts/arnes.md §1 y §5 de H40; SC-006), que entre los dos piden los cinco
// pares de la familia que se miden.
func casosValidosDeNormativa() []casoValidoDelArnes {
	enHerramienta := func(variante string) repartoEsperado {
		return repartoEsperado{
			variante: variante, modo: ModoHerramienta, porEval: 2, abiertas: 8,
			linea: "- " + variante + ", herramienta: 8 abiertas.",
		}
	}

	return []casoValidoDelArnes{
		{
			nombre: "el-primer-lanzamiento",
			argumentos: func(a *ArgumentosDelLaboratorio) {
				a.Variantes, a.Modos, a.Sesiones = controlDeNormativaDelTest, "orden,herramienta", "4"
			},
			repartos: []repartoEsperado{
				{
					variante: controlDeNormativaDelTest, modo: ModoOrden, porEval: 1, abiertas: 4,
					linea: "- normativa-control, orden: 4 abiertas.",
				},
				{
					variante: controlDeNormativaDelTest, modo: ModoHerramienta, porEval: 1, abiertas: 4,
					linea: "- normativa-control, herramienta: 4 abiertas.",
				},
			},
		},
		{
			nombre: "el-segundo-lanzamiento",
			argumentos: func(a *ArgumentosDelLaboratorio) {
				a.Variantes, a.Modos, a.Sesiones = soloHerramientaDeNormativaDelTest, "herramienta", "8"
			},
			repartos: []repartoEsperado{
				enHerramienta("normativa-remisiones-en-orden"), enHerramienta("normativa-cuerpo-corto"),
				enHerramienta("normativa-testigo-en-referencia"),
			},
		},
	}
}

// TestArnesDeNormativa fija el arnés del sondeo del laboratorio con la familia
// de normativa, sin modelo (contracts/arnes.md §2, §3 y §5 de H40; data-model
// §1, §4 y §5 de H40; research D7, D8 y D14 de H40; FR-008, FR-010, FR-011;
// SC-006), con la construcción sustituida y el sustituto de claude de
// TestArnesDelLaboratorio.
//
// Con una variante que no es de ninguna familia, las tres que solo se miden en
// modo herramienta pedidas en modo orden, SESIONES=6, variantes de las dos
// familias, una versión de Claude Code distinta o un alias del modelo, devuelve
// un error de uso con su mensaje, no construye nada, no abre ninguna sesión y no
// crea DIRECTORIO (exigirElErrorDeUsoDelArnes).
//
// Con los dos lanzamientos de la aceptación, con menos sesiones, abre
// SESIONES/4 sesiones de cada una de las cuatro evals del sondeo de normativa en
// cada variante y modo pedidos, y ninguna más; cada una ve solo normativa, la de
// su variante, sin servidor.json en el modo orden y con kitlegal mcp serve en el
// modo herramienta (exigirLaSesionDeNormativa).
//
// Sin una de las cuatro evals en evals/normativa, falla con su nombre antes de
// construir nada, y no es un error de uso.
func TestArnesDeNormativa(t *testing.T) {
	t.Parallel()

	for _, caso := range casosDeUsoDeNormativa() {
		t.Run(caso.nombre, func(t *testing.T) {
			t.Parallel()

			exigirElErrorDeUsoDelArnes(t, caso)
		})
	}

	for _, caso := range casosValidosDeNormativa() {
		t.Run(caso.nombre, func(t *testing.T) {
			t.Parallel()

			exigirElSondeoDelArnes(t, caso, evalsDeNormativaDelTest, exigirLaSesionDeNormativa)
		})
	}

	t.Run("sin-una-de-las-cuatro-evals", func(t *testing.T) {
		t.Parallel()

		exigirElSondeoSinUnaEvalDeNormativa(t, "15-irpf-rendimientos-por-materia.yaml")
	})
}

// exigirLaSesionDeNormativa exige que la sesión de ese directorio vea solo la
// skill normativa, enlazada a la copia de su variante y con el SKILL.md de
// laboratorio/normativa/<variante>/ byte a byte; y que no tenga servidor.json en
// el modo orden y, en el modo herramienta, uno que arranca kitlegal mcp serve
// del bin/ del temporal, sin laboratorio servidor (contracts/arnes.md §3 y §5 de
// H40; FR-008).
func exigirLaSesionDeNormativa(t *testing.T, arnes arnesDelTest, reparto repartoEsperado, sesion string) {
	t.Helper()

	skills := filepath.Join(sesion, directorioDeClaude, directorioDeSkills)

	enlazadas, err := os.ReadDir(skills)
	require.NoError(t, err)

	nombres := make([]string, 0, len(enlazadas))
	for _, enlazada := range enlazadas {
		nombres = append(nombres, enlazada.Name())
	}

	require.Equal(t, []string{"normativa"}, nombres, "la sesión %s ve solo normativa", sesion)

	destino, err := os.Readlink(filepath.Join(skills, "normativa"))
	require.NoError(t, err)
	assert.Equal(t, filepath.Join(arnes.sondeo.Temporal, carpetaDeLasSkillsDelSondeo, reparto.variante, "normativa"),
		destino, "la sesión %s enlaza la copia de su variante", sesion)

	original, err := leerFichero(filepath.Join(raizDelRepositorio, "laboratorio", "normativa", reparto.variante,
		"skills", "normativa", "SKILL.md"))
	require.NoError(t, err)

	copiado, err := leerFichero(filepath.Join(destino, "SKILL.md"))
	require.NoError(t, err)
	assert.Equal(t, string(original), string(copiado), "el SKILL.md de la sesión %s es el de %s", sesion,
		reparto.variante)

	servidor := filepath.Join(sesion, ficheroDelServidor)
	if reparto.modo == ModoOrden {
		assert.NoFileExists(t, servidor, "una sesión del modo orden no tiene servidor")

		return
	}

	kitlegal := filepath.Join(arnes.sondeo.Temporal, carpetaDeBinarios, programaDeLasConsultas)
	assert.Equal(t, []string{kitlegal, "mcp", "serve"}, ordenDelServidorEscrito(t, servidor),
		"el servidor de %s es el de hoy", sesion)
}

// exigirElSondeoSinUnaEvalDeNormativa ejecuta el arnés con normativa-control en
// los dos modos sobre un repositorio con la declaración de H39 y una copia de
// evals/normativa sin la eval dada, y exige que falle con su nombre, sin que sea
// un error de uso, antes de construir nada, abrir ninguna sesión, escribir en el
// temporal o crear DIRECTORIO (contracts/arnes.md §3.2 de H40).
func exigirElSondeoSinUnaEvalDeNormativa(t *testing.T, falta string) {
	t.Helper()

	raiz := t.TempDir()
	evals := filepath.Join(raiz, "evals", "normativa")
	require.NoError(t, os.CopyFS(evals, os.DirFS(filepath.Join(raizDelRepositorio, "evals", "normativa"))))
	require.NoError(t, os.Remove(filepath.Join(evals, falta)))

	deH39, err := leerFichero(filepath.Join(raizDelRepositorio, "laboratorio", "variantes.yaml"))
	require.NoError(t, err)
	require.NoError(t, os.Mkdir(filepath.Join(raiz, "laboratorio"), 0o700))
	require.NoError(t, os.WriteFile(filepath.Join(raiz, "laboratorio", "variantes.yaml"), deH39, 0o600))

	arnes := nuevoArnesDelTest(t, func(a *ArgumentosDelLaboratorio) {
		a.Variantes, a.Modos = controlDeNormativaDelTest, "orden,herramienta"
	}, "")
	arnes.sondeo.Raiz = raiz

	salida, err := sondearElLaboratorio(t.Context().Done(), arnes.sondeo)
	require.EqualError(t, err, "evals/normativa no tiene la eval "+falta+" del sondeo")
	assert.Empty(t, salida)

	var uso *errorDeUso
	assert.NotErrorAs(t, err, &uso, "una eval del sondeo que falta no es un error de uso")

	assert.Empty(t, arnes.claude.sesiones(t), "no se abre ninguna sesión")
	assert.False(t, *arnes.construido, "no se construye nada")

	entradas, err := os.ReadDir(arnes.sondeo.Temporal)
	require.NoError(t, err)
	assert.Empty(t, entradas, "no se escribe nada en el temporal")
	assert.NoDirExists(t, filepath.Dir(arnes.sondeo.Argumentos.Directorio), "no se crea DIRECTORIO ni su padre")
}

// TestConstruirLosProgramasDelLaboratorio construye de verdad kitlegal y
// laboratorio para el arnés del laboratorio en un temporal del test (contracts/arnes.md
// §3.1 de H39): kitlegal en bin/ y laboratorio en laboratorio-bin/, cada uno un
// fichero ejecutable, y el de laboratorio-bin/ es el programa del laboratorio,
// que sin orden escribe su uso y sale con 1. La base es la del proceso, con sus
// cachés de Go y sin red (GOPROXY=off, GOFLAGS=-mod=readonly).
func TestConstruirLosProgramasDelLaboratorio(t *testing.T) {
	t.Parallel()

	temporal := t.TempDir()
	base := sobreLaBase(os.Environ(), []string{"GOPROXY=off", "GOFLAGS=-mod=readonly"})

	require.NoError(t, construirLosProgramasDelLaboratorio(raizDelRepositorio, temporal, base))

	kitlegal := filepath.Join(temporal, carpetaDeBinarios, programaDeLasConsultas)
	laboratorio := filepath.Join(temporal, carpetaDelProgramaDelLaboratorio, programaDelLaboratorio)

	for _, programa := range []string{kitlegal, laboratorio} {
		estado, err := os.Stat(programa)
		require.NoError(t, err, "%s existe", programa)
		assert.True(t, estado.Mode().IsRegular(), "%s es un fichero", programa)
		assert.NotZero(t, estado.Mode().Perm()&0o100, "%s es ejecutable", programa)
	}

	codigo, salida, deError := codigoYSalidas(t, ordenSinArgumentos(t, laboratorio))
	assert.Equal(t, 1, codigo, "laboratorio sin orden sale con 1")
	assert.Empty(t, salida)
	assert.Contains(t, deError, PrefijoDelLaboratorio+"uso: "+UsoDelServidorDelLaboratorio+"\n")
}

// ordenSinArgumentos es la orden que ejecuta el programa de esa ruta sin ningún
// argumento, el único que no es constante (gosec G204), con el contexto del
// test.
func ordenSinArgumentos(t *testing.T, programa string) *exec.Cmd {
	t.Helper()

	return exec.CommandContext(t.Context(), programa)
}

// ordenEsperadaDelServidor es la orden con la que servidor.json arranca el
// servidor en el modo herramienta de la variante: kitlegal mcp serve del bin/
// del temporal y, en instrucciones y parametro, laboratorio servidor delante,
// con el texto de su fichero sin su salto de línea final (contracts/arnes.md
// §3.4 de H39).
func ordenEsperadaDelServidor(t *testing.T, temporal, variante string) []string {
	t.Helper()

	kitlegal := filepath.Join(temporal, carpetaDeBinarios, programaDeLasConsultas)
	filtro := []string{
		filepath.Join(temporal, carpetaDelProgramaDelLaboratorio, programaDelLaboratorio), "servidor", "-kitlegal", kitlegal,
	}

	texto := func(fichero string) string {
		contenido, err := leerFichero(filepath.Join(directorioDeLaVariante(raizDelRepositorio, variante), fichero))
		require.NoError(t, err)

		return strings.TrimSuffix(string(contenido), "\n")
	}

	switch variante {
	case "instrucciones":
		return append(filtro, "-instrucciones", texto(ficheroDeLasInstrucciones))
	case "parametro":
		return append(filtro, "-herramienta", "boe_articulo", "-parametro", "bloque", "-descripcion",
			texto(ficheroDelParametro))
	default:
		return []string{kitlegal, "mcp", "serve"}
	}
}

// ordenDelServidorEscrito es la orden, command y args, del servidor kitlegal de
// ese servidor.json.
func ordenDelServidorEscrito(t *testing.T, servidor string) []string {
	t.Helper()

	contenido, err := leerFichero(servidor)
	require.NoError(t, err)

	var declarado struct {
		MCPServers map[string]struct {
			Command string   `json:"command"`
			Args    []string `json:"args"`
		} `json:"mcpServers"`
	}

	require.NoError(t, json.Unmarshal(contenido, &declarado))
	require.Contains(t, declarado.MCPServers, nombreDelServidor)

	kitlegal := declarado.MCPServers[nombreDelServidor]

	return append([]string{kitlegal.Command}, kitlegal.Args...)
}

// guionDelLaboratorio es scripts/evals-laboratorio.sh, relativo al directorio
// de este paquete, que es donde go test ejecuta los tests.
const guionDelLaboratorio = "../../scripts/evals-laboratorio.sh"

// nombreDelTemporalDelLaboratorio es el del directorio que el guion del
// laboratorio crea en TMPDIR: la plantilla de mktemp,
// kitlegal-laboratorio.XXXXXX, con sus seis caracteres sustituidos.
var nombreDelTemporalDelLaboratorio = regexp.MustCompile(`^kitlegal-laboratorio\.[A-Za-z0-9]{6}$`)

// Lo que el go que deja el uso escribe en uso.txt en los casos de
// TestGuionDelLaboratorio, con el salto de línea final con el que lo escribe
// TestSondeoDelLaboratorio (contracts/arnes.md §2 de H39): los errores de
// SESIONES=6 y MODELO=sonnet, uno por línea y en el orden de los argumentos, y
// el de DIRECTORIO vacío.
const (
	usoDelLaboratorioConDosArgumentos = "SESIONES: «6» no es un múltiplo de 4 mayor que 0\n" +
		"MODELO: «sonnet» no es un identificador completo de modelo (claude-<familia>-<versión>, como " +
		"claude-sonnet-5-5)\n"
	usoDelLaboratorioSinDirectorio = "DIRECTORIO: está vacío\n"
)

// Los DIRECTORIO de los casos de TestGuionDelLaboratorio: uno relativo, que el
// guion resuelve desde la raíz del repositorio, y uno absoluto, que pasa tal
// cual. Ninguno se crea: lo crearía el arnés, y el sustituto de go no lo
// ejecuta.
const (
	directorioRelativoDelLaboratorio = "sondeos/h39-a"
	directorioAbsolutoDelLaboratorio = "/sondeos/del-laboratorio/h39-b"
)

// casoDelGuionDelLaboratorio es un caso de TestGuionDelLaboratorio: uno del
// guion del sondeo, con la orden del guion del laboratorio, y si el DIRECTORIO
// que recibe TestSondeoDelLaboratorio es el séptimo argumento del guion con la
// raíz del repositorio delante, por ser relativo, o el séptimo tal cual.
type casoDelGuionDelLaboratorio struct {
	casoDelGuionDelSondeo

	desdeLaRaiz bool
}

// TestGuionDelLaboratorio fija scripts/evals-laboratorio.sh (contracts/arnes.md
// §1 de H39; FR-030, FR-032 de H39), como TestGuionDelSondeo el guion del sondeo
// de una skill, con el sustituto de go delante en el PATH y un TMPDIR vacío del
// test. Con siete argumentos, crea en TMPDIR su temporal con la plantilla
// kitlegal-laboratorio.XXXXXX y ejecuta, en la raíz del repositorio y con el
// tmp/ de ese temporal como TMPDIR, la orden go test de TestSondeoDelLaboratorio
// con los siete tras -args, DIRECTORIO con la raíz delante si es relativo, y
// -temporal con ese temporal, con sus dos salidas en su go-test.log. Si go test
// sale con 0 y deja uso.txt, el guion sale con 1 e imprime solo uso.txt, en la
// salida de error; si go test falla, aunque deje uso.txt, sale con 1 e imprime
// su registro; si no, sale con 0 e imprime salida.txt. Sin siete argumentos, el
// uso y el código 1, sin ejecutar go. En todos los casos, el TMPDIR del test
// queda vacío: el guion borra su temporal.
func TestGuionDelLaboratorio(t *testing.T) {
	t.Parallel()

	guion, err := filepath.Abs(guionDelLaboratorio)
	require.NoError(t, err)

	raiz, err := filepath.EvalSymlinks(filepath.Dir(filepath.Dir(guion)))
	require.NoError(t, err)

	for _, caso := range casosDelGuionDelLaboratorio() {
		t.Run(caso.nombre, func(t *testing.T) {
			t.Parallel()

			exigirElGuionDelLaboratorio(t, caso, guion, raiz)
		})
	}
}

// exigirElGuionDelLaboratorio ejecuta el guion del laboratorio de esa ruta con
// la orden del caso, con un directorio común y un TMPDIR suyos, y exige lo que
// el caso espera de su código, sus dos salidas, su TMPDIR y la orden de go.
func exigirElGuionDelLaboratorio(t *testing.T, caso casoDelGuionDelLaboratorio, guion, raiz string) {
	t.Helper()

	comun, temporal := t.TempDir(), t.TempDir()
	orden := caso.orden(t.Context(), guion)

	codigo, salida, deError := ejecutarElGuionDelSondeo(t, orden, caso.casoDelGuionDelSondeo, comun, temporal)

	assert.Equalf(t, caso.codigo, codigo, "el código del guion, con esta salida de error:\n%s", deError)
	assert.Equal(t, caso.salida, salida, "la salida estándar del guion")
	exigirLaSalidaDeErrorDelGuion(t, caso.casoDelGuionDelSondeo, deError)

	entradas, err := os.ReadDir(temporal)
	require.NoError(t, err)
	assert.Empty(t, entradas, "el guion borra su temporal y el TMPDIR queda vacío")

	anotaciones := filepath.Join(comun, sustitutoGo)
	if caso.sinGo {
		assert.NoDirExists(t, anotaciones, "sin siete argumentos, el guion no ejecuta go")

		return
	}

	delGuion := orden.Args[1:]

	directorio := delGuion[6]
	if caso.desdeLaRaiz {
		directorio = filepath.Join(raiz, directorio)
	}

	exigirElGoDelLaboratorio(t, leerElGoAnotado(t, anotaciones), delGuion, directorio, temporal, raiz)
}

// exigirElGoDelLaboratorio exige que el sustituto de go se haya ejecutado en la
// raíz del repositorio con la orden de TestSondeoDelLaboratorio
// (ordenDelLaboratorio), con los argumentos del guion, ese DIRECTORIO y un
// temporal en el TMPDIR dado, con el nombre de la plantilla y que ya tenía su
// go-test.log y su tmp/, y con ese tmp/ como TMPDIR.
func exigirElGoDelLaboratorio(t *testing.T, anotado goAnotado, delGuion []string, directorio, temporal, raiz string) {
	t.Helper()

	require.NotEmpty(t, anotado.argumentos)

	temporalDelGuion := anotado.argumentos[len(anotado.argumentos)-1]

	fisico, err := filepath.EvalSymlinks(temporal)
	require.NoError(t, err)

	assert.Equal(t, fisico, filepath.Dir(temporalDelGuion), "el temporal del guion está en TMPDIR")
	assert.Regexp(t, nombreDelTemporalDelLaboratorio, filepath.Base(temporalDelGuion))
	assert.Equal(t, ordenDelLaboratorio(delGuion, directorio, temporalDelGuion), anotado.argumentos)
	assert.Equal(t, raiz, anotado.directorio, "go test se ejecuta en la raíz del repositorio")
	assert.Equal(t, filepath.Join(temporalDelGuion, "tmp"), anotado.tmpdir, "el TMPDIR de go test es el tmp/ del temporal")
	assert.Equal(t, "go-test.log\ntmp", anotado.temporal, "el temporal ya tiene su go-test.log y su tmp/")
}

// ordenDelLaboratorio son los argumentos de go de la orden de
// TestSondeoDelLaboratorio (contracts/arnes.md §1 de H39), con los seis
// primeros del guion, en su orden, ese DIRECTORIO y su temporal.
func ordenDelLaboratorio(argumentos []string, directorio, temporal string) []string {
	return []string{
		"test", "-tags", "evals", "-count=1", "-timeout", "0", "-run", "^TestSondeoDelLaboratorio$", "./internal/evals/",
		"-args", "-variantes", argumentos[0], "-modos", argumentos[1], "-sesiones", argumentos[2], "-modelo",
		argumentos[3], "-claude-code", argumentos[4], "-concurrencia", argumentos[5], "-directorio", directorio,
		"-temporal", temporal,
	}
}

// casosDelGuionDelLaboratorio son los casos de TestGuionDelLaboratorio: go test
// sale con 0, con un DIRECTORIO relativo y con uno absoluto; sale con 1; sale
// con 0 y deja uso.txt, con los errores de SESIONES=6 y MODELO=sonnet o con el
// de DIRECTORIO vacío, que el guion pasa vacío; sale con 1 y deja uso.txt; y el
// guion recibe seis argumentos.
func casosDelGuionDelLaboratorio() []casoDelGuionDelLaboratorio {
	registro := []string{registroDeGoEnSuSalida, registroDeGoEnLaDeError}

	return []casoDelGuionDelLaboratorio{
		{casoDelGuionDelSondeo: casoDelGuionDelSondeo{
			nombre: "go-test-sale-con-0-y-directorio-relativo",
			orden: func(ctx context.Context, guion string) *exec.Cmd {
				return exec.CommandContext(ctx, guion, "control,instrucciones", "orden,herramienta", "8",
					modeloDelLaboratorio, versionDelClaudeDelLaboratorio, "2", directorioRelativoDelLaboratorio)
			},
			codigo: 0,
			salida: salidaDelSustitutoDeGo,
		}, desdeLaRaiz: true},
		{casoDelGuionDelSondeo: casoDelGuionDelSondeo{
			nombre: "go-test-sale-con-0-y-directorio-absoluto",
			orden: func(ctx context.Context, guion string) *exec.Cmd {
				return exec.CommandContext(ctx, guion, "parametro", "herramienta", "4", modeloDelLaboratorio,
					versionDelClaudeDelLaboratorio, "1", directorioAbsolutoDelLaboratorio)
			},
			codigo: 0,
			salida: salidaDelSustitutoDeGo,
		}},
		{casoDelGuionDelSondeo: casoDelGuionDelSondeo{
			nombre: "go-test-sale-con-1",
			orden: func(ctx context.Context, guion string) *exec.Cmd {
				return exec.CommandContext(ctx, guion, "control", "orden", "4", modeloDelLaboratorio,
					versionDelClaudeDelLaboratorio, "1", directorioRelativoDelLaboratorio)
			},
			codigoDeGo: 1,
			codigo:     1,
			deError:    registro,
		}, desdeLaRaiz: true},
		{casoDelGuionDelSondeo: casoDelGuionDelSondeo{
			nombre: "uso-con-dos-argumentos-que-no-valen",
			orden: func(ctx context.Context, guion string) *exec.Cmd {
				return exec.CommandContext(ctx, guion, "control", "orden", "6", "sonnet", versionDelClaudeDelLaboratorio,
					"1", directorioAbsolutoDelLaboratorio)
			},
			uso:    usoDelLaboratorioConDosArgumentos,
			codigo: 1,
		}},
		{casoDelGuionDelSondeo: casoDelGuionDelSondeo{
			nombre: "uso-sin-directorio",
			orden: func(ctx context.Context, guion string) *exec.Cmd {
				return exec.CommandContext(ctx, guion, "control", "orden", "4", modeloDelLaboratorio,
					versionDelClaudeDelLaboratorio, "1", "")
			},
			uso:    usoDelLaboratorioSinDirectorio,
			codigo: 1,
		}},
		{casoDelGuionDelSondeo: casoDelGuionDelSondeo{
			nombre: "go-test-sale-con-1-y-deja-uso",
			orden: func(ctx context.Context, guion string) *exec.Cmd {
				return exec.CommandContext(ctx, guion, "control", "orden", "6", "sonnet", versionDelClaudeDelLaboratorio,
					"1", directorioRelativoDelLaboratorio)
			},
			codigoDeGo: 1,
			uso:        usoDelLaboratorioConDosArgumentos,
			codigo:     1,
			deError:    registro,
		}, desdeLaRaiz: true},
		{casoDelGuionDelSondeo: casoDelGuionDelSondeo{
			nombre: "sin-siete-argumentos",
			orden: func(ctx context.Context, guion string) *exec.Cmd {
				return exec.CommandContext(ctx, guion, "control", "orden", "4", modeloDelLaboratorio,
					versionDelClaudeDelLaboratorio, "1")
			},
			codigo:  1,
			deError: []string{"uso: scripts/evals-laboratorio.sh"},
			sinGo:   true,
		}},
	}
}

// prefijoDeLaOrdenDelLlavero es con lo que empieza la orden de las recetas de
// make evals-sondeo y make evals-laboratorio: el guion del llavero, que toma la
// credencial (ADR 0032; contracts/arnes.md §1 de H39).
const prefijoDeLaOrdenDelLlavero = "scripts/evals-sondeo-llavero.sh"

// Las órdenes que TestRecetasDelSondeo espera de make -n en cada receta, con
// los valores fijos de sus casos (contracts/arnes.md §1 de H39): el guion del
// llavero con el nombre del guion que ejecuta y los argumentos de make detrás,
// cada uno entre comillas dobles y en su orden; los cinco del sondeo de una
// skill y los siete del laboratorio.
const (
	ordenDeLaRecetaDelSondeo      = `scripts/evals-sondeo-llavero.sh sondeo "normativa" "01" "claude-sonnet-5-5" "2" "1"`
	ordenDeLaRecetaDelLaboratorio = `scripts/evals-sondeo-llavero.sh laboratorio "control,instrucciones" ` +
		`"orden,herramienta" "8" "claude-sonnet-5-5" "2.1.284" "2" "sondeo-de-prueba"`
)

// variablesDelMakeQueLanzaElTest son las que el make que lanza go test, el de
// make ci, pasa a sus hijos: TestRecetasDelSondeo las quita del entorno de su
// make -n para que ni sus banderas ni su nivel cambien lo que imprime.
var variablesDelMakeQueLanzaElTest = []string{"MAKEFLAGS", "MFLAGS", "MAKELEVEL"}

// casoDeLasRecetasDelSondeo es un caso de TestRecetasDelSondeo: la receta, la
// orden make -n que la imprime con valores fijos, escrita entera con
// constantes (gosec G204), y la orden del guion del llavero que tiene que
// imprimir.
type casoDeLasRecetasDelSondeo struct {
	receta   string
	orden    func(ctx context.Context) *exec.Cmd
	esperada string
}

// TestRecetasDelSondeo es el control de las recetas de make evals-sondeo y make
// evals-laboratorio (contracts/arnes.md §1 y §6 de H39; FR-030, FR-033 de H39),
// sin modelo: ejecuta en la raíz del repositorio, sin las variables del make
// que lanza el test, make -n de cada una con valores fijos —-n solo imprime: ni
// guion, ni credencial, ni sesión— y exige en su salida una sola línea que
// empieza por el guion del llavero, igual carácter a carácter a la orden de la
// cadena con esos valores. Las demás líneas, las de la receta de check-tools,
// no cuentan. Es su único control: la política de las sesiones del run veda
// make evals-sondeo como orden.
func TestRecetasDelSondeo(t *testing.T) {
	t.Parallel()

	casos := []casoDeLasRecetasDelSondeo{
		{
			receta: "evals-sondeo",
			orden: func(ctx context.Context) *exec.Cmd {
				return exec.CommandContext(ctx, "make", "-n", "evals-sondeo", "SKILL=normativa", "EVALS=01",
					"MODELO=claude-sonnet-5-5", "REPETICIONES=2", "CONCURRENCIA=1")
			},
			esperada: ordenDeLaRecetaDelSondeo,
		},
		{
			receta: "evals-laboratorio",
			orden: func(ctx context.Context) *exec.Cmd {
				return exec.CommandContext(ctx, "make", "-n", "evals-laboratorio", "VARIANTES=control,instrucciones",
					"MODOS=orden,herramienta", "SESIONES=8", "MODELO=claude-sonnet-5-5", "CLAUDE_CODE=2.1.284",
					"CONCURRENCIA=2", "DIRECTORIO=sondeo-de-prueba")
			},
			esperada: ordenDeLaRecetaDelLaboratorio,
		},
	}

	for _, caso := range casos {
		t.Run(caso.receta, func(t *testing.T) {
			t.Parallel()

			orden := caso.orden(t.Context())
			orden.Dir = raizDelRepositorio
			orden.Env = sobreLaBase(os.Environ(), nil, variablesDelMakeQueLanzaElTest...)

			codigo, salida, deError := codigoYSalidas(t, orden)
			require.Equalf(t, 0, codigo, "make -n %s sale con 0; su salida de error:\n%s", caso.receta, deError)

			var delLlavero []string

			for linea := range strings.Lines(salida) {
				if strings.HasPrefix(linea, prefijoDeLaOrdenDelLlavero) {
					delLlavero = append(delLlavero, strings.TrimSuffix(linea, "\n"))
				}
			}

			assert.Equalf(t, []string{caso.esperada}, delLlavero,
				"make -n %s imprime una sola orden del guion del llavero, la de la cadena; su salida:\n%s", caso.receta,
				salida)
		})
	}
}
