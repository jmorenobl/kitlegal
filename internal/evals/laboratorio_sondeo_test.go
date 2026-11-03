package evals

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path"
	"path/filepath"
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

// casosDeUsoDelArnes son los casos de TestArnesDelLaboratorio que son un error de
// uso (contracts/arnes.md §2 de H39; FR-030, FR-031; SC-007).
func casosDeUsoDelArnes() []casoDeUsoDelArnes {
	fijo := func(mensaje string) func(ArgumentosDelLaboratorio) string {
		return func(ArgumentosDelLaboratorio) string { return mensaje }
	}

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
					sesionesDelReparto(repartoEsperado{porEval: 1})[0])

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

			exigirElSondeoDelArnes(t, caso)
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

// exigirElSondeoDelArnes ejecuta el arnés de un caso válido y exige sus
// sesiones, su sondeo.json, su servidor.json y su salida.
func exigirElSondeoDelArnes(t *testing.T, caso casoValidoDelArnes) {
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
		plan := sesionesDelReparto(reparto)
		require.LessOrEqual(t, reparto.abiertas, len(plan), "el caso abre de las sesiones de su plan")

		for _, sesion := range plan[:reparto.abiertas] {
			esperadas = append(esperadas, filepath.Join(fisico, reparto.variante, string(reparto.modo), sesion))
			exigirLaSesionDelArnes(t, arnes, reparto, filepath.Join(directorio, reparto.variante, string(reparto.modo),
				sesion))
		}
	}

	slices.Sort(esperadas)
	assert.Equal(t, esperadas, arnes.claude.sesiones(t), "se abren las sesiones de cada eval en cada variante y modo "+
		"pedidos, y ninguna más")
}

// sesionesDelReparto son los nombres de las sesiones que abre un reparto: las
// porEval de cada eval del laboratorio, <eval>-<nn>.
func sesionesDelReparto(reparto repartoEsperado) []string {
	var sesiones []string

	for _, eval := range evalsDelLaboratorioDelTest {
		for numero := 1; numero <= reparto.porEval; numero++ {
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
