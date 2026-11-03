package evals

import (
	"bytes"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// Las sesiones fijas de dos sondeos del laboratorio y la tabla que el guion da
// de ellas, escrita a mano a partir de contracts/tabla.md de H39 y de esas
// sesiones, con los extremos de Wilson calculados aparte (FR-043 de H39; SC-006
// de H39).
//
// Los dos sondeos, h39-a y h39-b, con el mismo modelo y otra versión de Claude
// Code, cubren lo de contracts/tabla.md §6 de H39: los dos modos; Skill y la
// primera herramienta en la misma respuesta, también en dos mensajes assistant
// con el mismo id y el resultado de Skill entre ellos; Skill y ToolSearch en un
// turno y boe_articulo en el siguiente; una orden kitlegal en una tubería, otra
// en una secuencia y otra con una asignación delante, la ruta entre comillas y
// una barra invertida al final de la línea; which kitlegal, --help, un grep,
// skills list y --describe antes de la primera orden que consulta; una sesión
// que activa sin ninguna herramienta; la referencia abierta dos veces, con Read
// por tramos y con Bash; Grep y Glob sobre ella, y el cierre.md de otra skill;
// una variante con referencia que ninguna sesión abre; sesiones sin terminar
// por el tope, por un result con is_error, sin result y sin respuesta; y una
// fila sin ninguna sesión terminada.
const (
	sondeosFijosDelLaboratorio = "testdata/laboratorio/sondeo"
	tablaFijaDelLaboratorio    = "testdata/laboratorio/tabla.md"
)

// Dos sesiones fijas del modo herramienta que activan la skill de laboratorio
// antes de la primera herramienta de kitlegal, o sin ninguna, y no abren
// ninguna referencia: la primera cumple la regla testigo y la segunda no.
const (
	sesionFijaQueCumpleLaRegla   = "h39-a/control/herramienta/03-lrjsp-principio-de-legalidad-01"
	sesionFijaQueNoCumpleLaRegla = "h39-a/entrada/herramienta/03-lrjsp-principio-de-legalidad-01"
)

// sesionDeLasCopias es el nombre de las sesiones del sondeo de 40, el de la eval
// de las dos sesiones fijas que copia, con su número.
const sesionDeLasCopias = "referencia-paso/herramienta/03-lrjsp-principio-de-legalidad-%02d"

// Lo que deja en un sondeo el arnés del laboratorio y que el guion lee
// (data-model §6 de H39): el sondeo.json con el modelo y la versión, y de cada
// sesión, los ficheros que lee LeerSesion.
const (
	sondeoDeCuarentaSesiones     = `{"modelo":"claude-sonnet-5-5","claude_code":"2.1.317"}` + "\n"
	sesionesDeCadaClaseDelSondeo = 20
)

// filaDeCuarentaSesiones es la fila de 40 sesiones de una variante con
// referencia en el modo herramienta: 20 copias de cada una de las dos sesiones
// fijas, todas activan antes, ninguna abre la referencia y la mitad cumple la
// regla. Los extremos de Wilson de las tres celdas son los de la historia 3 del
// spec de H39 para 40, 0 y 20 de 40 (SC-006 de H39; research V21 de H39).
const filaDeCuarentaSesiones = "| referencia-paso | herramienta | claude-sonnet-5-5 | 2.1.317 | 0 | 40 | " +
	"40 de 40 · 100,0 % [91,2–100,0 %] |  |  | 0 de 40 · 0,0 % [0,0–8,8 %] |  | " +
	"20 de 40 · 50,0 % [35,2–64,8 %] |\n"

// TestTablaDelLaboratorio fija el guion que cuenta un sondeo del laboratorio
// (FR-040 a FR-043 de H39; SC-006 de H39; contracts/tabla.md de H39): sobre las
// sesiones fijas, su salida es, byte a byte, la tabla escrita a mano; sobre un
// sondeo de 40 sesiones, las celdas de 40, 0 y 20 de 40 llevan los extremos de
// Wilson del spec; y un transcript truncado o un directorio sin ningún sondeo
// son un defecto: lo dice con la ruta en la salida de error y da 1 sin tabla.
func TestTablaDelLaboratorio(t *testing.T) {
	t.Parallel()

	t.Run("sondeos-fijos", probarLosSondeosFijos)
	t.Run("cuarenta-sesiones", probarCuarentaSesiones)
	t.Run("transcript-truncado", probarUnTranscriptTruncado)
	t.Run("sin-sondeo", probarUnDirectorioSinSondeo)
}

// probarLosSondeosFijos fija la salida del guion sobre las sesiones fijas: la
// de tabla.md, byte a byte, y nada en la salida de error.
func probarLosSondeosFijos(t *testing.T) {
	t.Parallel()

	esperada, err := os.ReadFile(tablaFijaDelLaboratorio)
	require.NoError(t, err)

	tabla := contarElSondeo(t, sondeosFijosDelLaboratorio)

	require.Zerof(t, tabla.codigo, "salida de error: %s", tabla.errores)
	assert.Equal(t, string(esperada), tabla.salida, "la tabla de las sesiones fijas, byte a byte")
	assert.Empty(t, tabla.errores)
}

// probarCuarentaSesiones fija los extremos de Wilson de 40, 0 y 20 de 40 en un
// sondeo que el test monta con 40 copias de dos sesiones fijas, en una variante
// con referencia, y el detalle de las 40: las 20 primeras, copias de la que
// cumple la regla, y las 20 siguientes, de la que no.
func probarCuarentaSesiones(t *testing.T) {
	t.Parallel()

	sondeo := t.TempDir()
	escribirElSondeo(t, sondeo)

	for numero := 1; numero <= sesionesDeCadaClaseDelSondeo; numero++ {
		copiarLaSesionFija(t, sesionFijaQueCumpleLaRegla, sondeo, fmt.Sprintf(sesionDeLasCopias, numero))
		copiarLaSesionFija(t, sesionFijaQueNoCumpleLaRegla, sondeo,
			fmt.Sprintf(sesionDeLasCopias, sesionesDeCadaClaseDelSondeo+numero))
	}

	tabla := contarElSondeo(t, sondeo)

	require.Zerof(t, tabla.codigo, "salida de error: %s", tabla.errores)
	assert.Contains(t, tabla.salida, filaDeCuarentaSesiones)
	assert.Equal(t, 1, strings.Count(tabla.salida, "| referencia-paso | herramienta |"), "una sola fila")
	assert.Contains(t, tabla.salida, "| "+fmt.Sprintf(sesionDeLasCopias, 1)+" | sí | sí |  | no | sí |\n")
	assert.Contains(t, tabla.salida,
		"| "+fmt.Sprintf(sesionDeLasCopias, 2*sesionesDeCadaClaseDelSondeo)+" | sí | sí |  | no | no |\n")
	assert.Equal(t, 2*sesionesDeCadaClaseDelSondeo, strings.Count(tabla.salida, "\n| referencia-paso/herramienta/"),
		"una línea del detalle por sesión, con la ruta relativa al sondeo")
	assert.Empty(t, tabla.errores)
}

// probarUnTranscriptTruncado fija que un transcript cortado a media línea,
// junto a una sesión que se lee, es un defecto: 1, ninguna tabla y la ruta de la
// sesión en la salida de error.
func probarUnTranscriptTruncado(t *testing.T) {
	t.Parallel()

	sondeo := t.TempDir()
	escribirElSondeo(t, sondeo)
	copiarLaSesionFija(t, sesionFijaQueCumpleLaRegla, sondeo, fmt.Sprintf(sesionDeLasCopias, 1))
	truncada := copiarLaSesionFija(t, sesionFijaQueNoCumpleLaRegla, sondeo, fmt.Sprintf(sesionDeLasCopias, 2))

	// El transcript, con su última línea cortada por la mitad.
	directorio, err := os.OpenRoot(truncada)
	require.NoError(t, err)

	t.Cleanup(func() { require.NoError(t, directorio.Close()) })

	contenido, err := directorio.ReadFile(ficheroDelTranscript)
	require.NoError(t, err)

	principio := bytes.LastIndexByte(bytes.TrimSuffix(contenido, []byte("\n")), '\n') + 1
	require.NoError(t, directorio.WriteFile(ficheroDelTranscript, contenido[:principio+(len(contenido)-principio)/2], 0o600))

	tabla := contarElSondeo(t, sondeo)

	assert.Equal(t, 1, tabla.codigo)
	assert.Empty(t, tabla.salida, "sin tabla")
	assert.Contains(t, tabla.errores, truncada)
	assert.True(t, strings.HasPrefix(tabla.errores, PrefijoDelLaboratorio), "con el prefijo del programa: %s",
		tabla.errores)
}

// probarUnDirectorioSinSondeo fija que un directorio sin ningún sondeo, ni él
// ni ninguno de sus subdirectorios, es un defecto: 1, ninguna tabla y su ruta
// en la salida de error. Es el de las respuestas fijas de la regla testigo, que
// no tiene ningún sondeo.json.
func probarUnDirectorioSinSondeo(t *testing.T) {
	t.Parallel()

	tabla := contarElSondeo(t, respuestasDeLaReglaTestigo)

	assert.Equal(t, 1, tabla.codigo)
	assert.Empty(t, tabla.salida, "sin tabla")
	assert.Contains(t, tabla.errores, respuestasDeLaReglaTestigo)
	assert.True(t, strings.HasPrefix(tabla.errores, PrefijoDelLaboratorio), "con el prefijo del programa: %s",
		tabla.errores)
}

// tablaContada es lo que devuelve el guion que cuenta un sondeo: su código, su
// salida y su salida de error.
type tablaContada struct {
	codigo  int
	salida  string
	errores string
}

// contarElSondeo ejecuta el guion sobre esa ruta con la declaración del
// laboratorio del repositorio.
func contarElSondeo(t *testing.T, ruta string) tablaContada {
	t.Helper()

	var salida, errores bytes.Buffer

	codigo := TablaDelLaboratorio(raizDelRepositorio, []string{ruta}, &salida, &errores)

	return tablaContada{codigo: codigo, salida: salida.String(), errores: errores.String()}
}

// escribirElSondeo escribe en ese directorio el sondeo.json de un sondeo.
func escribirElSondeo(t *testing.T, sondeo string) {
	t.Helper()

	require.NoError(t, os.WriteFile(filepath.Join(sondeo, ficheroDelSondeo), []byte(sondeoDeCuarentaSesiones), 0o600))
}

// copiarLaSesionFija copia la sesión fija de esa ruta, relativa a las sesiones
// fijas, en la ruta dada del sondeo, y devuelve su directorio.
func copiarLaSesionFija(t *testing.T, fija, sondeo, ruta string) string {
	t.Helper()

	destino := filepath.Join(sondeo, filepath.FromSlash(ruta))
	require.NoError(t, os.MkdirAll(filepath.Dir(destino), 0o750))
	require.NoError(t, os.CopyFS(destino, os.DirFS(filepath.Join(sondeosFijosDelLaboratorio, filepath.FromSlash(fija)))))

	return destino
}
