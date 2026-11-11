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
// las que quedan sin abrir tras el límite de uso; los transcripts que escribe
// el sustituto, por sesión <variante>/<modo>/<sesión>; y la skill de la
// variante <skill>-producto que pide, cuyas referencias de skills/<skill> lista
// su sondeo.json, vacía si no pide ninguna (contracts/arnes.md §3.4).
type casoValidoDelArnes struct {
	nombre      string
	argumentos  func(a *ArgumentosDelLaboratorio)
	repartos    []repartoEsperado
	transcripts func(t *testing.T) map[string]string
	producto    string
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

			exigirElSondeoDelArnes(t, caso, evalsDelLaboratorioDelTest, nil, exigirLaSesionDelArnes)
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
// ninguna más—, su sondeo.json, con las skills instaladas dadas, ninguna en uno
// de H39 (contracts/arnes.md §3.3 de H41.2; research D11 de H41.2), lo que ve
// cada sesión, con exigirLaSesion, y su salida.
func exigirElSondeoDelArnes(t *testing.T, caso casoValidoDelArnes, evals, instaladas []string,
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

	referencias := referenciasDelProductoDelTest(t, caso.producto)

	contenido, err := leerFichero(filepath.Join(directorio, ficheroDelSondeo))
	require.NoError(t, err)
	assert.Equal(t, sondeoEscritoDelTest(t, referencias, instaladas), string(contenido), "sondeo.json, byte a byte")

	leido, err := leerElSondeo(directorio)
	require.NoError(t, err, "el guion que cuenta el sondeo lee su sondeo.json")
	assert.Equal(t, sondeoLeido{
		Modelo: modeloDelLaboratorio, ClaudeCode: versionDelClaudeDelLaboratorio, ReferenciasDelProducto: referencias,
		SkillsInstaladas: instaladas,
	}, leido)

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

// sondeoSinElProductoDelTest es el sondeo.json de un sondeo que no pide ninguna
// variante <skill>-producto, el de H39 y H40, sin su salto de línea final
// (contracts/arnes.md §3.2 de H39; data-model §4).
const sondeoSinElProductoDelTest = `{"modelo":"claude-sonnet-5-5","claude_code":"2.1.317"}`

// referenciasDelProductoDelTest son los nombres de los ficheros de
// skills/<skill>/references del árbol, en orden de nombre, los que lista el
// sondeo.json de un sondeo que pide <skill>-producto (FR-032; data-model §4); sin
// skill, ninguno.
func referenciasDelProductoDelTest(t *testing.T, skill string) []string {
	t.Helper()

	if skill == "" {
		return nil
	}

	entradas, err := os.ReadDir(filepath.Join(raizDelRepositorio, "skills", skill, "references"))
	require.NoError(t, err)

	var referencias []string

	for _, entrada := range entradas {
		if entrada.Type().IsRegular() {
			referencias = append(referencias, entrada.Name())
		}
	}

	require.NotEmpty(t, referencias, "skills/%s tiene referencias", skill)
	slices.Sort(referencias)

	return referencias
}

// sondeoEscritoDelTest es el sondeo.json que deja el arnés, con su salto de
// línea final: el de H39 y H40 sin referencias del producto ni skills
// instaladas; con referencias, el mismo con referencias_del_producto detrás
// (data-model §4); y con skills instaladas, con skills_instaladas detrás de
// todo (data-model §3 de H41.2).
func sondeoEscritoDelTest(t *testing.T, referencias, instaladas []string) string {
	t.Helper()

	escrito := strings.TrimSuffix(sondeoSinElProductoDelTest, "}")

	for _, clave := range []struct {
		nombre string
		lista  []string
	}{
		{nombre: "referencias_del_producto", lista: referencias},
		{nombre: "skills_instaladas", lista: instaladas},
	} {
		if len(clave.lista) == 0 {
			continue
		}

		lista, err := json.Marshal(clave.lista)
		require.NoError(t, err)

		escrito += `,"` + clave.nombre + `":` + string(lista)
	}

	return escrito + "}\n"
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

			exigirElSondeoDelArnes(t, caso, evalsDeNormativaDelTest, []string{skillDeNormativaDelTest},
				exigirLaSesionDeNormativa)
		})
	}

	t.Run("sin-una-de-las-cuatro-evals", func(t *testing.T) {
		t.Parallel()

		exigirElSondeoSinUnaEvalDeNormativa(t, "15-irpf-rendimientos-por-materia.yaml")
	})
}

// exigirLaSesionDeNormativa exige que la sesión de ese directorio vea solo la
// skill normativa, enlazada a la copia de su variante y con los ficheros de
// laboratorio/normativa/<variante>/skills/normativa/ byte a byte; y que no tenga
// servidor.json en el modo orden y, en el modo herramienta, uno que arranca
// kitlegal mcp serve del bin/ del temporal, sin laboratorio servidor
// (contracts/arnes.md §3 y §5 de H40; FR-008): lo de exigirLaSesionDeLaFamilia
// con la familia de normativa, cuyas variantes de H40 son todas versionadas.
func exigirLaSesionDeNormativa(t *testing.T, arnes arnesDelTest, reparto repartoEsperado, sesion string) {
	t.Helper()

	exigirLaSesionDeLaFamilia(skillDeNormativaDelTest)(t, arnes, reparto, sesion)
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

// Las skills de las tres familias del producto y las evals del sondeo de las de
// plazos-y-territorio y jurisprudencia, de evals/<skill>, en su orden
// (data-model §1; research D15); las de normativa son evalsDeNormativaDelTest.
// El test las nombra, y no las lee de familiasDelProducto, para que una eval
// quitada del plan se vea.
const (
	skillDeNormativaDelTest      = "normativa"
	skillDePlazosDelTest         = "plazos-y-territorio"
	skillDeJurisprudenciaDelTest = "jurisprudencia"
)

var (
	evalsDePlazosDelTest = []string{
		"02-territorio-municipio-no-cubierto.yaml", "05-plazo-reposicion-leganes.yaml",
		"10-plazo-reposicion-anterior-a-la-lpac.yaml", "11-plazo-reposicion-sin-leer-el-precepto.yaml",
	}
	evalsDeJurisprudenciaDelTest = []string{
		"01-existe-con-numero-y-fecha.yaml", "03-por-materia.yaml", "04-documento-pegado.yaml",
		"06-documento-que-no-es-el-pedido.yaml",
	}
)

// sufijoDelProductoDelTest es el del nombre de la variante de una familia que
// se instala desde skills/<skill>, <skill>-producto (data-model §3).
const sufijoDelProductoDelTest = "-producto"

// evalsDeCadaFamiliaDelTest es cuántas evals planifica el sondeo de cada familia
// del producto, entre las que se reparten por igual las SESIONES de cada variante
// y modo (FR-030; data-model §1).
const evalsDeCadaFamiliaDelTest = 4

// casosDeUsoDeLasFamilias son los casos de TestArnesDeLasFamilias que son un
// error de uso (contracts/arnes.md §2 y §5; FR-030; SC-008): cada uno, con su
// mensaje y nada hecho.
func casosDeUsoDeLasFamilias() []casoDeUsoDelArnes {
	return []casoDeUsoDelArnes{
		{
			nombre:     "con-una-variante-que-no-existe",
			argumentos: func(_ *testing.T, a *ArgumentosDelLaboratorio) { a.Variantes = "plazos-y-territorio-otra" },
			error:      mensajeFijo("VARIANTES: «plazos-y-territorio-otra» no es ninguna variante del laboratorio"),
		},
		{
			nombre: "con-una-de-solo-herramienta-en-modo-orden",
			argumentos: func(_ *testing.T, a *ArgumentosDelLaboratorio) {
				a.Variantes, a.Modos = "normativa-remisiones-en-orden", "orden"
			},
			error: mensajeFijo("VARIANTES y MODOS: normativa-remisiones-en-orden no se mide en modo orden"),
		},
		{
			nombre: "con-dos-familias-del-producto",
			argumentos: func(_ *testing.T, a *ArgumentosDelLaboratorio) {
				a.Variantes = "normativa-producto,jurisprudencia-control"
			},
			error: mensajeFijo("VARIANTES: se mezclan variantes de normativa y variantes de jurisprudencia"),
		},
		{
			// La de H39 va primero, en el orden de las familias, aunque se pida
			// detrás.
			nombre: "con-una-del-producto-y-la-de-h39",
			argumentos: func(_ *testing.T, a *ArgumentosDelLaboratorio) {
				a.Variantes = "plazos-y-territorio-control,control"
			},
			error: mensajeFijo("VARIANTES: se mezclan variantes del laboratorio de H39 y variantes de " +
				"plazos-y-territorio"),
		},
		{
			nombre: "con-las-cuatro-familias",
			argumentos: func(_ *testing.T, a *ArgumentosDelLaboratorio) {
				a.Variantes = "jurisprudencia-producto,plazos-y-territorio-sin-condicion,normativa-control,control"
			},
			error: mensajeFijo("VARIANTES: se mezclan variantes del laboratorio de H39, variantes de normativa, " +
				"variantes de plazos-y-territorio y variantes de jurisprudencia"),
		},
		{
			nombre: "con-sesiones-que-no-son-multiplo-de-4",
			argumentos: func(_ *testing.T, a *ArgumentosDelLaboratorio) {
				a.Variantes, a.Modos, a.Sesiones = "jurisprudencia-control", "orden,herramienta", "6"
			},
			error: mensajeFijo("SESIONES: «6» no es un múltiplo de 4 mayor que 0"),
		},
		{
			nombre: "con-sesiones-0",
			argumentos: func(_ *testing.T, a *ArgumentosDelLaboratorio) {
				a.Variantes, a.Modos, a.Sesiones = "plazos-y-territorio-producto", "orden,herramienta", "0"
			},
			error: mensajeFijo("SESIONES: «0» no es un múltiplo de 4 mayor que 0"),
		},
		{
			// El ejemplo de contracts/arnes.md §2: todos los errores juntos, uno por
			// línea.
			nombre: "el-ejemplo-del-contrato",
			argumentos: func(_ *testing.T, a *ArgumentosDelLaboratorio) {
				a.Variantes, a.Modos, a.Sesiones = "normativa-producto,jurisprudencia-control,plazos-y-territorio-otra",
					"orden", "6"
			},
			error: mensajeFijo(strings.Join([]string{
				"VARIANTES: «plazos-y-territorio-otra» no es ninguna variante del laboratorio",
				"VARIANTES: se mezclan variantes de normativa y variantes de jurisprudencia",
				"SESIONES: «6» no es un múltiplo de 4 mayor que 0",
			}, "\n")),
		},
	}
}

// casoDeLasFamilias es un caso de TestArnesDeLasFamilias en el que el arnés abre
// sesiones: el caso, la skill de su familia, las evals de su sondeo, en su
// orden, y cuántas sesiones abre en total.
type casoDeLasFamilias struct {
	caso     casoValidoDelArnes
	skill    string
	evals    []string
	sesiones int
}

// casosValidosDeLasFamilias son los casos de TestArnesDeLasFamilias en los que
// el arnés abre sesiones (contracts/arnes.md §1 y §5; FR-033; SC-007): las cuatro
// invocaciones de la aceptación con SESIONES=4, que abren 16, 4, 40 y 16, y la
// segunda de normativa con SESIONES=8, dos de cada eval.
func casosValidosDeLasFamilias() []casoDeLasFamilias {
	return []casoDeLasFamilias{
		{
			caso: invocacionDeLasFamilias("la-primera-de-normativa", "normativa-control,normativa-producto",
				"orden,herramienta", 1, skillDeNormativaDelTest),
			skill: skillDeNormativaDelTest, evals: evalsDeNormativaDelTest, sesiones: 16,
		},
		{
			caso: invocacionDeLasFamilias("la-segunda-de-normativa", "normativa-remisiones-en-orden", "herramienta", 1,
				""),
			skill: skillDeNormativaDelTest, evals: evalsDeNormativaDelTest, sesiones: 4,
		},
		{
			caso: invocacionDeLasFamilias("la-de-plazos-y-territorio", "plazos-y-territorio-control,"+
				"plazos-y-territorio-producto,plazos-y-territorio-sin-condicion,plazos-y-territorio-condicion-por-orden,"+
				"plazos-y-territorio-lista-con-verbo", "orden,herramienta", 1, skillDePlazosDelTest),
			skill: skillDePlazosDelTest, evals: evalsDePlazosDelTest, sesiones: 40,
		},
		{
			caso: invocacionDeLasFamilias("la-de-jurisprudencia", "jurisprudencia-control,jurisprudencia-producto",
				"orden,herramienta", 1, skillDeJurisprudenciaDelTest),
			skill: skillDeJurisprudenciaDelTest, evals: evalsDeJurisprudenciaDelTest, sesiones: 16,
		},
		{
			caso: invocacionDeLasFamilias("la-segunda-de-normativa-con-8", "normativa-remisiones-en-orden",
				"herramienta", 2, ""),
			skill: skillDeNormativaDelTest, evals: evalsDeNormativaDelTest, sesiones: 8,
		},
	}
}

// invocacionDeLasFamilias es el caso válido de una invocación con esas
// VARIANTES, en el orden de su declaración, y esos MODOS, con porEval sesiones
// de cada una de las cuatro evals de la familia: un reparto por variante y modo
// pedidos, orden antes que herramienta, que abre todas las de su plan; producto
// es la skill de la variante <skill>-producto que pide, o vacía.
func invocacionDeLasFamilias(nombre, variantes, modos string, porEval int, producto string) casoValidoDelArnes {
	pedidos := strings.Split(modos, ",")
	abiertas := porEval * evalsDeCadaFamiliaDelTest

	var repartos []repartoEsperado

	for variante := range strings.SplitSeq(variantes, ",") {
		for _, modo := range []Modo{ModoOrden, ModoHerramienta} {
			if !slices.Contains(pedidos, string(modo)) {
				continue
			}

			repartos = append(repartos, repartoEsperado{
				variante: variante, modo: modo, porEval: porEval, abiertas: abiertas,
				linea: fmt.Sprintf("- %s, %s: %d abiertas.", variante, modo, abiertas),
			})
		}
	}

	return casoValidoDelArnes{
		nombre: nombre,
		argumentos: func(a *ArgumentosDelLaboratorio) {
			a.Variantes, a.Modos, a.Sesiones = variantes, modos, fmt.Sprint(abiertas)
		},
		repartos: repartos,
		producto: producto,
	}
}

// TestArnesDeLasFamilias fija el arnés del sondeo del laboratorio con las tres
// familias del producto, sin modelo (contracts/arnes.md §2 a §5; data-model §1,
// §3 y §4; research D14, D15 y V9; FR-022, FR-030, FR-032, FR-033; SC-007,
// SC-008), con la construcción sustituida y el sustituto de claude de
// TestArnesDelLaboratorio, que anota sus invocaciones.
//
// Con una variante que no existe, una que solo se mide en modo herramienta
// pedida en modo orden, variantes de dos familias del producto, o de una y la de
// H39, o de las cuatro, SESIONES=6 o SESIONES=0, devuelve un error de uso con su
// mensaje, no construye nada, no abre ninguna sesión y no crea DIRECTORIO
// (exigirElErrorDeUsoDelArnes).
//
// Con las cuatro invocaciones de la aceptación, con SESIONES=4, y con la
// segunda de normativa con SESIONES=8, abre SESIONES/4 sesiones de cada una de
// las cuatro evals de la familia en cada variante y modo pedidos, y ninguna más;
// cada una ve solo la skill de la familia, la de su variante
// (exigirLaSesionDeLaFamilia); y sondeo.json lista las referencias de
// skills/<skill> si se pidió <skill>-producto, y es el de H39 si no
// (exigirElSondeoDelArnes).
func TestArnesDeLasFamilias(t *testing.T) {
	t.Parallel()

	for _, caso := range casosDeUsoDeLasFamilias() {
		t.Run(caso.nombre, func(t *testing.T) {
			t.Parallel()

			exigirElErrorDeUsoDelArnes(t, caso)
		})
	}

	for _, deLaFamilia := range casosValidosDeLasFamilias() {
		t.Run(deLaFamilia.caso.nombre, func(t *testing.T) {
			t.Parallel()

			abiertas := 0
			for _, reparto := range deLaFamilia.caso.repartos {
				abiertas += reparto.abiertas
			}

			require.Equal(t, deLaFamilia.sesiones, abiertas, "las sesiones que abre la invocación")

			exigirElSondeoDelArnes(t, deLaFamilia.caso, deLaFamilia.evals, []string{deLaFamilia.skill},
				exigirLaSesionDeLaFamilia(deLaFamilia.skill))
		})
	}
}

// exigirLaSesionDeLaFamilia exige que una sesión de la familia de esa skill vea
// solo la skill, enlazada a la copia de su variante en el temporal, con los
// ficheros, ruta a ruta y byte a byte, de cada origen de su variante
// (origenesDeLaVariante); y que no tenga servidor.json en el modo orden y, en el
// modo herramienta, uno que arranca kitlegal mcp serve del bin/ del temporal
// (contracts/arnes.md §3 y §5; FR-022, FR-030).
func exigirLaSesionDeLaFamilia(skill string) exigirLaSesionDelReparto {
	return func(t *testing.T, arnes arnesDelTest, reparto repartoEsperado, sesion string) {
		t.Helper()

		skills := filepath.Join(sesion, directorioDeClaude, directorioDeSkills)

		enlazadas, err := os.ReadDir(skills)
		require.NoError(t, err)

		nombres := make([]string, 0, len(enlazadas))
		for _, enlazada := range enlazadas {
			nombres = append(nombres, enlazada.Name())
		}

		require.Equal(t, []string{skill}, nombres, "la sesión %s ve solo %s", sesion, skill)

		destino, err := os.Readlink(filepath.Join(skills, skill))
		require.NoError(t, err)
		assert.Equal(t, filepath.Join(arnes.sondeo.Temporal, carpetaDeLasSkillsDelSondeo, reparto.variante, skill),
			destino, "la sesión %s enlaza la copia de su variante", sesion)

		instalada := ficherosDeLaSkillDelTest(t, skills, skill)

		for _, origen := range origenesDeLaVariante(skill, reparto.variante) {
			assert.Equal(t, ficherosDeLaSkillDelTest(t, origen, skill), instalada,
				"la skill de la sesión %s es, ruta a ruta y byte a byte, la de %s", sesion, origen)
		}

		exigirElServidorDeLaFamilia(t, arnes, reparto, sesion)
	}
}

// exigirElServidorDeLaFamilia exige que una sesión de una familia del producto
// no tenga servidor.json en el modo orden y, en el modo herramienta, uno que
// arranca kitlegal mcp serve del bin/ del temporal (contracts/arnes.md §3 de
// H41; FR-030 de H41).
func exigirElServidorDeLaFamilia(t *testing.T, arnes arnesDelTest, reparto repartoEsperado, sesion string) {
	t.Helper()

	servidor := filepath.Join(sesion, ficheroDelServidor)
	if reparto.modo == ModoOrden {
		assert.NoFileExists(t, servidor, "una sesión del modo orden no tiene servidor")

		return
	}

	kitlegal := filepath.Join(arnes.sondeo.Temporal, carpetaDeBinarios, programaDeLasConsultas)
	assert.Equal(t, []string{kitlegal, "mcp", "serve"}, ordenDelServidorEscrito(t, servidor),
		"el servidor de %s es el de hoy", sesion)
}

// origenesDeLaVariante son los directorios de skills cuya skill de la familia
// tiene que ser, ruta a ruta y byte a byte, la de las sesiones de la variante
// (contracts/arnes.md §5 de H41): laboratorio/<skill>/<variante>/skills en una
// variante versionada, y skills/ del árbol en <skill>-producto (FR-022 de H41).
func origenesDeLaVariante(skill, variante string) []string {
	if variante != skill+sufijoDelProductoDelTest {
		return []string{filepath.Join(raizDelRepositorio, "laboratorio", skill, variante, "skills")}
	}

	return []string{filepath.Join(raizDelRepositorio, "skills")}
}

// ficherosDeLaSkillDelTest son los ficheros de la skill de ese directorio de
// skills, por su ruta relativa a la de la skill, con su contenido como texto,
// para que la diferencia se lea (leerLosFicheros). Exige alguno, para no
// comparar en vacío.
func ficherosDeLaSkillDelTest(t *testing.T, skills, skill string) map[string]string {
	t.Helper()

	ficheros, err := leerLosFicheros(os.DirFS(skills), skill)
	require.NoError(t, err)
	require.NotEmpty(t, ficheros, "la skill %s de %s tiene ficheros", skill, skills)

	comoTexto := make(map[string]string, len(ficheros))
	for ruta, contenido := range ficheros {
		comoTexto[ruta] = string(contenido)
	}

	return comoTexto
}

// skillsDelProductoDelTest son las cuatro skills de skills/ del árbol, en orden
// de nombre: las que ve cada sesión de un sondeo con INSTALACION=producto y las
// que anota su sondeo.json (FR-010, FR-012; contracts/arnes.md §3 y §5). El test
// las nombra, y no las lee de skills/, para que una que falta o sobra se vea.
var skillsDelProductoDelTest = []string{"consulta-juridica", "jurisprudencia", "normativa", "plazos-y-territorio"}

// casosDeUsoDeLaInstalacion son los casos de TestArnesConLaInstalacionDelProducto
// que son un error de uso (contracts/arnes.md §2 y §5; FR-011): un valor que no
// es producto, con una familia del producto, con la de H39 y sin familia;
// producto con la de H39; producto sin familia, porque ninguna variante vale o
// porque las que valen se mezclan, que no da el error de INSTALACION; y el de
// INSTALACION detrás del de DIRECTORIO. Cada uno, con su mensaje y nada hecho.
func casosDeUsoDeLaInstalacion() []casoDeUsoDelArnes {
	return []casoDeUsoDelArnes{
		{
			nombre: "con-instalacion-todas",
			argumentos: func(_ *testing.T, a *ArgumentosDelLaboratorio) {
				a.Variantes, a.Modos, a.Instalacion = "normativa-control", "herramienta", "todas"
			},
			error: mensajeFijo("INSTALACION: «todas» no es producto"),
		},
		{
			// Con la de H39, un valor que no es producto da solo su error.
			nombre:     "con-instalacion-todas-y-la-de-h39",
			argumentos: func(_ *testing.T, a *ArgumentosDelLaboratorio) { a.Instalacion = "todas" },
			error:      mensajeFijo("INSTALACION: «todas» no es producto"),
		},
		{
			// El segundo ejemplo de contracts/arnes.md §2, sin familia.
			nombre: "con-instalacion-todas-y-una-variante-que-no-existe",
			argumentos: func(_ *testing.T, a *ArgumentosDelLaboratorio) {
				a.Variantes, a.Modos, a.Instalacion = "otra", "herramienta", "todas"
			},
			error: mensajeFijo(strings.Join([]string{
				"VARIANTES: «otra» no es ninguna variante del laboratorio",
				"INSTALACION: «todas» no es producto",
			}, "\n")),
		},
		{
			nombre:     "con-instalacion-producto-y-la-de-h39",
			argumentos: func(_ *testing.T, a *ArgumentosDelLaboratorio) { a.Instalacion = "producto" },
			error: mensajeFijo("INSTALACION: las variantes del laboratorio de H39 no son de ninguna familia " +
				"del producto"),
		},
		{
			nombre: "con-instalacion-producto-y-una-variante-que-no-existe",
			argumentos: func(_ *testing.T, a *ArgumentosDelLaboratorio) {
				a.Variantes, a.Modos, a.Instalacion = "otra", "herramienta", "producto"
			},
			error: mensajeFijo("VARIANTES: «otra» no es ninguna variante del laboratorio"),
		},
		{
			nombre: "con-instalacion-producto-y-una-mezcla",
			argumentos: func(_ *testing.T, a *ArgumentosDelLaboratorio) {
				a.Variantes, a.Modos, a.Instalacion = "normativa-control,control", "herramienta", "producto"
			},
			error: mensajeFijo("VARIANTES: se mezclan variantes del laboratorio de H39 y variantes de normativa"),
		},
		{
			nombre: "con-instalacion-todas-detras-del-directorio",
			argumentos: func(_ *testing.T, a *ArgumentosDelLaboratorio) {
				a.Variantes, a.Modos, a.Directorio, a.Instalacion = "jurisprudencia-control", "herramienta", "", "todas"
			},
			error: mensajeFijo("DIRECTORIO: está vacío\nINSTALACION: «todas» no es producto"),
		},
	}
}

// casosValidosDeLaInstalacion son los casos de
// TestArnesConLaInstalacionDelProducto en los que el arnés abre sesiones
// (contracts/arnes.md §5): una invocación por familia del producto con
// INSTALACION=producto, VARIANTES=<skill>-control,<skill>-producto,
// MODOS=herramienta y SESIONES=4, que abre 8.
func casosValidosDeLaInstalacion() []casoDeLasFamilias {
	familias := []struct {
		skill string
		evals []string
	}{
		{skill: skillDeNormativaDelTest, evals: evalsDeNormativaDelTest},
		{skill: skillDePlazosDelTest, evals: evalsDePlazosDelTest},
		{skill: skillDeJurisprudenciaDelTest, evals: evalsDeJurisprudenciaDelTest},
	}

	casos := make([]casoDeLasFamilias, 0, len(familias))

	for _, familia := range familias {
		caso := invocacionDeLasFamilias("la-de-"+familia.skill+"-con-el-producto",
			familia.skill+"-control,"+familia.skill+sufijoDelProductoDelTest, string(ModoHerramienta), 1, familia.skill)

		sinLaInstalacion := caso.argumentos
		caso.argumentos = func(a *ArgumentosDelLaboratorio) {
			sinLaInstalacion(a)
			a.Instalacion = "producto"
		}

		casos = append(casos, casoDeLasFamilias{caso: caso, skill: familia.skill, evals: familia.evals, sesiones: 8})
	}

	return casos
}

// TestArnesConLaInstalacionDelProducto fija INSTALACION=producto en el arnés del
// sondeo del laboratorio, sin modelo (contracts/arnes.md §2, §3 y §5; data-model
// §3; research D7, D10 y D11; FR-010, FR-011, FR-012, FR-015; SC-003, SC-004),
// con la construcción sustituida y el sustituto de claude de
// TestArnesDelLaboratorio, que anota sus invocaciones.
//
// Con un valor que no vale, o con producto y la de H39, devuelve un error de uso
// con los mensajes de §2, no construye nada, no abre ninguna sesión y no crea
// DIRECTORIO (exigirElErrorDeUsoDelArnes).
//
// Con producto, en una invocación por familia con su <skill>-control y su
// <skill>-producto, cada sesión ve las cuatro skills del producto: la de su
// familia, la de su variante, y cada otra, la de skills/ del árbol, ruta a ruta
// y byte a byte (exigirLaSesionConElProducto); y sondeo.json anota las cuatro en
// skills_instaladas (exigirElSondeoDelArnes).
func TestArnesConLaInstalacionDelProducto(t *testing.T) {
	t.Parallel()

	for _, caso := range casosDeUsoDeLaInstalacion() {
		t.Run(caso.nombre, func(t *testing.T) {
			t.Parallel()

			exigirElErrorDeUsoDelArnes(t, caso)
		})
	}

	for _, deLaFamilia := range casosValidosDeLaInstalacion() {
		t.Run(deLaFamilia.caso.nombre, func(t *testing.T) {
			t.Parallel()

			abiertas := 0
			for _, reparto := range deLaFamilia.caso.repartos {
				abiertas += reparto.abiertas
			}

			require.Equal(t, deLaFamilia.sesiones, abiertas, "las sesiones que abre la invocación")

			exigirElSondeoDelArnes(t, deLaFamilia.caso, deLaFamilia.evals, skillsDelProductoDelTest,
				exigirLaSesionConElProducto(deLaFamilia.skill))
		})
	}
}

// exigirLaSesionConElProducto exige que una sesión de la familia de esa skill,
// con INSTALACION=producto, vea exactamente las cuatro skills del producto, cada
// una enlazada a la copia de su variante en el temporal: la de la familia, con
// los ficheros, ruta a ruta y byte a byte, de cada origen de su variante
// (origenesDeLaVariante), y cada otra, con los de skills/ del árbol; y su
// servidor, el de la familia (exigirElServidorDeLaFamilia). Falla con la primera
// sesión cuyas skills difieren (contracts/arnes.md §3.1 y §5; FR-010).
func exigirLaSesionConElProducto(skill string) exigirLaSesionDelReparto {
	return func(t *testing.T, arnes arnesDelTest, reparto repartoEsperado, sesion string) {
		t.Helper()

		skills := filepath.Join(sesion, directorioDeClaude, directorioDeSkills)

		enlazadas, err := os.ReadDir(skills)
		require.NoError(t, err)

		nombres := make([]string, 0, len(enlazadas))
		for _, enlazada := range enlazadas {
			nombres = append(nombres, enlazada.Name())
		}

		require.Equal(t, skillsDelProductoDelTest, nombres, "la sesión %s ve las cuatro skills del producto", sesion)

		copia := filepath.Join(arnes.sondeo.Temporal, carpetaDeLasSkillsDelSondeo, reparto.variante)

		for _, nombre := range nombres {
			destino, err := os.Readlink(filepath.Join(skills, nombre))
			require.NoError(t, err)
			require.Equal(t, filepath.Join(copia, nombre), destino, "la sesión %s enlaza la copia de su variante", sesion)

			origenes := []string{filepath.Join(raizDelRepositorio, "skills")}
			if nombre == skill {
				origenes = origenesDeLaVariante(skill, reparto.variante)
			}

			instalada := ficherosDeLaSkillDelTest(t, skills, nombre)

			for _, origen := range origenes {
				require.Equal(t, ficherosDeLaSkillDelTest(t, origen, nombre), instalada,
					"la skill %s de la sesión %s es, ruta a ruta y byte a byte, la de %s", nombre, sesion, origen)
			}
		}

		exigirElServidorDeLaFamilia(t, arnes, reparto, sesion)
	}
}

// errorDeEvalConLaDeH39 es el error de EVAL con variantes de la familia de H39,
// que no tiene el plan de una familia del producto (contracts/arnes.md §2).
const errorDeEvalConLaDeH39 = "EVAL: las variantes del laboratorio de H39 no tienen el plan de una familia del producto"

// casosDeUsoDeUnaEval son los casos de TestArnesConUnaEval que son un error de
// uso (contracts/arnes.md §2 y §5; FR-011, FR-013): un EVAL que no es el número
// de dos cifras de ninguna eval del plan de la familia pedida —07 en normativa,
// 4, 04-lgt y, en plazos-y-territorio, 04, que es de normativa—; EVAL con la de
// H39; EVAL con SESIONES=0, que con él pide un entero mayor que 0; EVAL sin
// familia, que no da su error; y el primer ejemplo de §2, con sus tres líneas en
// su orden. Cada uno, con su mensaje y nada hecho.
func casosDeUsoDeUnaEval() []casoDeUsoDelArnes {
	return []casoDeUsoDelArnes{
		{
			nombre: "con-eval-07-en-normativa",
			argumentos: func(_ *testing.T, a *ArgumentosDelLaboratorio) {
				a.Variantes, a.Modos, a.Eval = "normativa-producto", "herramienta", "07"
			},
			error: mensajeFijo("EVAL: «07» no es el número de ninguna eval del plan de normativa"),
		},
		{
			nombre: "con-eval-de-una-cifra",
			argumentos: func(_ *testing.T, a *ArgumentosDelLaboratorio) {
				a.Variantes, a.Modos, a.Eval = "normativa-producto", "herramienta", "4"
			},
			error: mensajeFijo("EVAL: «4» no es el número de ninguna eval del plan de normativa"),
		},
		{
			// El comienzo del nombre de la eval no es su número.
			nombre: "con-eval-y-parte-de-su-nombre",
			argumentos: func(_ *testing.T, a *ArgumentosDelLaboratorio) {
				a.Variantes, a.Modos, a.Eval = "normativa-producto", "herramienta", "04-lgt"
			},
			error: mensajeFijo("EVAL: «04-lgt» no es el número de ninguna eval del plan de normativa"),
		},
		{
			// 04 es de una eval del plan de normativa, no del de plazos-y-territorio.
			nombre: "con-eval-04-en-plazos-y-territorio",
			argumentos: func(_ *testing.T, a *ArgumentosDelLaboratorio) {
				a.Variantes, a.Modos, a.Eval = "plazos-y-territorio-control", "herramienta", "04"
			},
			error: mensajeFijo("EVAL: «04» no es el número de ninguna eval del plan de plazos-y-territorio"),
		},
		{
			nombre:     "con-eval-y-la-de-h39",
			argumentos: func(_ *testing.T, a *ArgumentosDelLaboratorio) { a.Eval = "04" },
			error:      mensajeFijo(errorDeEvalConLaDeH39),
		},
		{
			nombre: "con-eval-y-sesiones-0",
			argumentos: func(_ *testing.T, a *ArgumentosDelLaboratorio) {
				a.Variantes, a.Modos, a.Sesiones, a.Eval = "normativa-producto", "herramienta", "0", "04"
			},
			error: mensajeFijo("SESIONES: «0» no es un entero mayor que 0"),
		},
		{
			// Sin familia, EVAL no da error: ya lo da el de VARIANTES.
			nombre: "con-eval-y-una-variante-que-no-existe",
			argumentos: func(_ *testing.T, a *ArgumentosDelLaboratorio) {
				a.Variantes, a.Modos, a.Eval = "otra", "herramienta", "07"
			},
			error: mensajeFijo("VARIANTES: «otra» no es ninguna variante del laboratorio"),
		},
		{
			// El primer ejemplo de contracts/arnes.md §2: el error de SESIONES en su
			// sitio, y los de INSTALACION y EVAL detrás del de DIRECTORIO, en ese
			// orden.
			nombre: "el-ejemplo-del-contrato",
			argumentos: func(_ *testing.T, a *ArgumentosDelLaboratorio) {
				a.Variantes, a.Modos, a.Sesiones, a.Concurrencia, a.Instalacion, a.Eval = "control", "orden", "0", "1",
					"todas", "04"
			},
			error: mensajeFijo(strings.Join([]string{
				"SESIONES: «0» no es un entero mayor que 0",
				"INSTALACION: «todas» no es producto",
				errorDeEvalConLaDeH39,
			}, "\n")),
		},
	}
}

// casoDeUnaEval es un caso de TestArnesConUnaEval en el que el arnés abre
// sesiones: el caso, la eval pedida, las skills que anota su sondeo.json y lo
// que ve cada una de sus sesiones.
type casoDeUnaEval struct {
	caso           casoValidoDelArnes
	eval           string
	instaladas     []string
	exigirLaSesion exigirLaSesionDelReparto
}

// invocacionDeUnaEval es el caso válido de una invocación con esa variante de la
// familia de esa skill, en modo herramienta, con EVAL el número de dos cifras
// de esa eval, el comienzo de su nombre, esas SESIONES y esa INSTALACION: un
// reparto que abre sesiones sesiones de la eval, <eval>-01 a <eval>-<sesiones>,
// y ninguna más; cada una con las skills de hoy o, con INSTALACION=producto, con
// las cuatro del producto (contracts/arnes.md §3 y §5; FR-013).
func invocacionDeUnaEval(nombre, variante, skill, eval string, sesiones int, instalacion string) casoDeUnaEval {
	producto := ""
	if variante == skill+sufijoDelProductoDelTest {
		producto = skill
	}

	deUnaEval := casoDeUnaEval{
		caso: casoValidoDelArnes{
			nombre: nombre,
			argumentos: func(a *ArgumentosDelLaboratorio) {
				a.Variantes, a.Modos, a.Sesiones = variante, string(ModoHerramienta), fmt.Sprint(sesiones)
				a.Instalacion, a.Eval = instalacion, eval[:2]
			},
			repartos: []repartoEsperado{{
				variante: variante, modo: ModoHerramienta, porEval: sesiones, abiertas: sesiones,
				linea: fmt.Sprintf("- %s, herramienta: %d abiertas.", variante, sesiones),
			}},
			producto: producto,
		},
		eval:           eval,
		instaladas:     []string{skill},
		exigirLaSesion: exigirLaSesionDeLaFamilia(skill),
	}

	if instalacion == "producto" {
		deUnaEval.instaladas, deUnaEval.exigirLaSesion = skillsDelProductoDelTest, exigirLaSesionConElProducto(skill)
	}

	return deUnaEval
}

// casosValidosDeUnaEval son los casos de TestArnesConUnaEval en los que el arnés
// abre sesiones (contracts/arnes.md §1 y §5; FR-013, FR-030; SC-009): una
// variante de cada familia con SESIONES que no son múltiplo de 4, y las dos
// invocaciones de FR-030 con SESIONES=2, sin INSTALACION y con producto.
func casosValidosDeUnaEval() []casoDeUnaEval {
	return []casoDeUnaEval{
		invocacionDeUnaEval("normativa-con-la-04", "normativa-producto", skillDeNormativaDelTest,
			"04-lgt-prescripcion.yaml", 3, ""),
		invocacionDeUnaEval("plazos-y-territorio-con-la-05", "plazos-y-territorio-control", skillDePlazosDelTest,
			"05-plazo-reposicion-leganes.yaml", 2, ""),
		invocacionDeUnaEval("jurisprudencia-con-la-03", "jurisprudencia-producto", skillDeJurisprudenciaDelTest,
			"03-por-materia.yaml", 1, ""),
		invocacionDeUnaEval("la-de-fr-030-con-la-skill-sola", "normativa-producto", skillDeNormativaDelTest,
			"04-lgt-prescripcion.yaml", 2, ""),
		invocacionDeUnaEval("la-de-fr-030-con-el-producto", "normativa-producto", skillDeNormativaDelTest,
			"04-lgt-prescripcion.yaml", 2, "producto"),
	}
}

// TestArnesConUnaEval fija EVAL en el arnés del sondeo del laboratorio, sin
// modelo (contracts/arnes.md §2, §3 y §5; data-model §3; research D8 y D9;
// FR-011, FR-013, FR-015; SC-004, SC-009), con la construcción sustituida y el
// sustituto de claude de TestArnesDelLaboratorio, que anota sus invocaciones.
//
// Con un EVAL que no es el número de ninguna eval del plan de la familia, con
// la de H39 o con SESIONES=0, devuelve un error de uso con los mensajes de §2,
// no construye nada, no abre ninguna sesión y no crea DIRECTORIO
// (exigirElErrorDeUsoDelArnes).
//
// Con una eval del plan, abre SESIONES sesiones de esa eval en el par pedido,
// <eval>-01 a <eval>-<SESIONES>, y ninguna más, aunque SESIONES no sea múltiplo
// de 4; cada una, con las skills de hoy o, con INSTALACION=producto, con las
// cuatro del producto; y sondeo.json anota las skills instaladas
// (exigirElSondeoDelArnes).
func TestArnesConUnaEval(t *testing.T) {
	t.Parallel()

	for _, caso := range casosDeUsoDeUnaEval() {
		t.Run(caso.nombre, func(t *testing.T) {
			t.Parallel()

			exigirElErrorDeUsoDelArnes(t, caso)
		})
	}

	for _, deUnaEval := range casosValidosDeUnaEval() {
		t.Run(deUnaEval.caso.nombre, func(t *testing.T) {
			t.Parallel()

			exigirElSondeoDelArnes(t, deUnaEval.caso, []string{deUnaEval.eval}, deUnaEval.instaladas,
				deUnaEval.exigirLaSesion)
		})
	}
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
// test. Con nueve argumentos, crea en TMPDIR su temporal con la plantilla
// kitlegal-laboratorio.XXXXXX y ejecuta, en la raíz del repositorio y con el
// tmp/ de ese temporal como TMPDIR, la orden go test de TestSondeoDelLaboratorio
// con los nueve tras -args, DIRECTORIO con la raíz delante si es relativo,
// INSTALACION y EVAL detrás, y -temporal con ese temporal, el último, con sus
// dos salidas en su go-test.log. Si go test sale con 0 y deja uso.txt, el guion
// sale con 1 e imprime solo uso.txt, en la salida de error; si go test falla,
// aunque deje uso.txt, sale con 1 e imprime su registro; si no, sale con 0 e
// imprime salida.txt. Sin nueve argumentos, el uso y el código 1, sin ejecutar
// go (contracts/arnes.md §1 y §5 de H41.2; research D9 de H41.2). En todos los casos, el TMPDIR del test
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
		assert.NoDirExists(t, anotaciones, "sin nueve argumentos, el guion no ejecuta go")

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
// TestSondeoDelLaboratorio (contracts/arnes.md §1 y §5 de H39 y de H41.2), con
// los seis primeros de los nueve argumentos del guion, en su orden, ese
// DIRECTORIO, el octavo, INSTALACION, y el noveno, EVAL, tal cual, y su
// temporal, el último.
func ordenDelLaboratorio(argumentos []string, directorio, temporal string) []string {
	return []string{
		"test", "-tags", "evals", "-count=1", "-timeout", "0", "-run", "^TestSondeoDelLaboratorio$", "./internal/evals/",
		"-args", "-variantes", argumentos[0], "-modos", argumentos[1], "-sesiones", argumentos[2], "-modelo",
		argumentos[3], "-claude-code", argumentos[4], "-concurrencia", argumentos[5], "-directorio", directorio,
		"-instalacion", argumentos[7], "-eval", argumentos[8], "-temporal", temporal,
	}
}

// casosDelGuionDelLaboratorio son los casos de TestGuionDelLaboratorio, con nueve
// argumentos, el octavo producto en el de un DIRECTORIO absoluto, el noveno 04
// en el de uno relativo, con el octavo vacío, para que los dos se vean por
// separado, y los dos vacíos en los demás: go test sale con 0, con un DIRECTORIO
// relativo y con uno absoluto; sale con 1; sale con 0 y deja uso.txt, con los
// errores de SESIONES=6 y MODELO=sonnet o con el de DIRECTORIO vacío, que el
// guion pasa vacío; sale con 1 y deja uso.txt; y el guion recibe ocho
// argumentos, uno menos de los nueve.
func casosDelGuionDelLaboratorio() []casoDelGuionDelLaboratorio {
	registro := []string{registroDeGoEnSuSalida, registroDeGoEnLaDeError}

	return []casoDelGuionDelLaboratorio{
		{casoDelGuionDelSondeo: casoDelGuionDelSondeo{
			nombre: "go-test-sale-con-0-y-directorio-relativo",
			orden: func(ctx context.Context, guion string) *exec.Cmd {
				return exec.CommandContext(ctx, guion, "control,instrucciones", "orden,herramienta", "8",
					modeloDelLaboratorio, versionDelClaudeDelLaboratorio, "2", directorioRelativoDelLaboratorio, "", "04")
			},
			codigo: 0,
			salida: salidaDelSustitutoDeGo,
		}, desdeLaRaiz: true},
		{casoDelGuionDelSondeo: casoDelGuionDelSondeo{
			nombre: "go-test-sale-con-0-y-directorio-absoluto",
			orden: func(ctx context.Context, guion string) *exec.Cmd {
				return exec.CommandContext(ctx, guion, "parametro", "herramienta", "4", modeloDelLaboratorio,
					versionDelClaudeDelLaboratorio, "1", directorioAbsolutoDelLaboratorio, "producto", "")
			},
			codigo: 0,
			salida: salidaDelSustitutoDeGo,
		}},
		{casoDelGuionDelSondeo: casoDelGuionDelSondeo{
			nombre: "go-test-sale-con-1",
			orden: func(ctx context.Context, guion string) *exec.Cmd {
				return exec.CommandContext(ctx, guion, "control", "orden", "4", modeloDelLaboratorio,
					versionDelClaudeDelLaboratorio, "1", directorioRelativoDelLaboratorio, "", "")
			},
			codigoDeGo: 1,
			codigo:     1,
			deError:    registro,
		}, desdeLaRaiz: true},
		{casoDelGuionDelSondeo: casoDelGuionDelSondeo{
			nombre: "uso-con-dos-argumentos-que-no-valen",
			orden: func(ctx context.Context, guion string) *exec.Cmd {
				return exec.CommandContext(ctx, guion, "control", "orden", "6", "sonnet", versionDelClaudeDelLaboratorio,
					"1", directorioAbsolutoDelLaboratorio, "", "")
			},
			uso:    usoDelLaboratorioConDosArgumentos,
			codigo: 1,
		}},
		{casoDelGuionDelSondeo: casoDelGuionDelSondeo{
			nombre: "uso-sin-directorio",
			orden: func(ctx context.Context, guion string) *exec.Cmd {
				return exec.CommandContext(ctx, guion, "control", "orden", "4", modeloDelLaboratorio,
					versionDelClaudeDelLaboratorio, "1", "", "", "")
			},
			uso:    usoDelLaboratorioSinDirectorio,
			codigo: 1,
		}},
		{casoDelGuionDelSondeo: casoDelGuionDelSondeo{
			nombre: "go-test-sale-con-1-y-deja-uso",
			orden: func(ctx context.Context, guion string) *exec.Cmd {
				return exec.CommandContext(ctx, guion, "control", "orden", "6", "sonnet", versionDelClaudeDelLaboratorio,
					"1", directorioRelativoDelLaboratorio, "", "")
			},
			codigoDeGo: 1,
			uso:        usoDelLaboratorioConDosArgumentos,
			codigo:     1,
			deError:    registro,
		}, desdeLaRaiz: true},
		{casoDelGuionDelSondeo: casoDelGuionDelSondeo{
			// Los ocho de antes: el guion da su uso, que nombra <instalacion> y
			// <eval> detrás de <directorio>, y no ejecuta go.
			nombre: "sin-nueve-argumentos",
			orden: func(ctx context.Context, guion string) *exec.Cmd {
				return exec.CommandContext(ctx, guion, "control", "orden", "4", modeloDelLaboratorio,
					versionDelClaudeDelLaboratorio, "1", directorioRelativoDelLaboratorio, "")
			},
			codigo: 1,
			deError: []string{"evals-laboratorio: uso: scripts/evals-laboratorio.sh <variantes> <modos> <sesiones> " +
				"<modelo> <claude-code> <concurrencia> <directorio> <instalacion> <eval>"},
			sinGo: true,
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
// skill y los nueve del laboratorio (contracts/arnes.md §1 y §5 de H41.2).
const (
	ordenDeLaRecetaDelSondeo      = `scripts/evals-sondeo-llavero.sh sondeo "normativa" "01" "claude-sonnet-5-5" "2" "1"`
	ordenDeLaRecetaDelLaboratorio = `scripts/evals-sondeo-llavero.sh laboratorio "control,instrucciones" ` +
		`"orden,herramienta" "8" "claude-sonnet-5-5" "2.1.284" "2" "sondeo-de-prueba" "producto" "04"`
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
					"CONCURRENCIA=2", "DIRECTORIO=sondeo-de-prueba", "INSTALACION=producto", "EVAL=04")
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
