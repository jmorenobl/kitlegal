package evals

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os"
	"path"
	"path/filepath"
	"slices"
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

// Las sesiones fijas de dos sondeos de las variantes de normativa y la tabla
// que el guion da de ellas, escrita a mano a partir de contracts/tabla.md §3 y
// §4 de H40 y de esas sesiones, con los extremos de Wilson de research V21 de
// H40 (FR-020 a FR-022 de H40; SC-005 de H40).
//
// Los dos sondeos, h40-a y h40-b, con el mismo modelo y otra versión de Claude
// Code, cubren lo de contracts/tabla.md §7 de H40: los dos modos; una referencia
// abierta dos veces con Read por tramos, otra con Bash en cada modo y otra
// entera, en los turnos 2 y 3 de una misma fila; Grep y Glob sobre una
// referencia, y el fichero del mismo nombre de otra skill; referencias que
// ninguna sesión abre; Skill y la primera herramienta en el mismo turno, y Skill
// después de la primera orden que consulta; una sesión que activa sin ninguna
// herramienta; sesiones sin terminar por el tope, por un result con is_error y
// sin respuesta; dos filas sin ninguna sesión terminada; y, en
// normativa-testigo-en-referencia, sesiones que cumplen la regla testigo y una
// que no, y fuera de ella, sesiones que la cumplen.
const (
	sondeosFijosDeNormativa = "testdata/laboratorio/normativa/sondeo"
	tablaFijaDeNormativa    = "testdata/laboratorio/normativa/tabla.md"
)

// Las evals del sondeo de normativa de las sesiones fijas que copian los tests,
// las dos primeras de familiaDeNormativa.evalsDelSondeo, y el sondeo fijo en que
// están.
const (
	evalDeNormativaDeLaLGT   = "04-lgt-prescripcion.yaml"
	evalDeNormativaDelTRLRHL = "05-trlrhl-impuestos-municipales.yaml"
	sondeoFijoDeNormativa    = "h40-a"
)

// sesionDelLaboratorio es el nombre que el arnés del laboratorio da a la sesión
// de ese número de la eval de ese fichero (contracts/arnes.md §3 de H39): el del
// fichero sin su extensión y el número con dos cifras.
func sesionDelLaboratorio(fichero string, numero int) string {
	return fmt.Sprintf("%s-%02d", strings.TrimSuffix(fichero, extensionDeUnaEval), numero)
}

// sesionDeNormativa es la ruta, relativa a su sondeo, de la sesión de ese número
// de esa eval en esa variante y ese modo.
func sesionDeNormativa(variante string, modo Modo, eval string, numero int) string {
	return path.Join(variante, string(modo), sesionDelLaboratorio(eval, numero))
}

// sesionFijaDeNormativa es la ruta, relativa a las sesiones fijas de
// normativa, de la primera sesión de esa eval en esa variante y ese modo del
// sondeo fijo h40-a.
func sesionFijaDeNormativa(variante string, modo Modo, eval string) string {
	return path.Join(sondeoFijoDeNormativa, sesionDeNormativa(variante, modo, eval, 1))
}

// Dos sesiones fijas de normativa-testigo-en-referencia, en el modo
// herramienta, que activan normativa antes de la primera herramienta de
// kitlegal: la de la eval de la LGT abre con_herramientas.md dos veces, la
// primera en el turno 2, y cumple la regla testigo; la del TRLRHL no abre
// ninguna referencia y no la cumple.
var (
	sesionFijaDeNormativaQueCumpleLaRegla = sesionFijaDeNormativa(varianteConLaReglaDeNormativa, ModoHerramienta,
		evalDeNormativaDeLaLGT)
	sesionFijaDeNormativaQueNoCumpleLaRegla = sesionFijaDeNormativa(varianteConLaReglaDeNormativa, ModoHerramienta,
		evalDeNormativaDelTRLRHL)
)

// copiaDeNormativa es la ruta, relativa al sondeo de 40 de normativa, de la
// sesión de ese número, todas de la eval de la LGT en
// normativa-testigo-en-referencia, en el modo herramienta.
func copiaDeNormativa(numero int) string {
	return sesionDeNormativa(varianteConLaReglaDeNormativa, ModoHerramienta, evalDeNormativaDeLaLGT, numero)
}

// filaDeCuarentaSesionesDeNormativa es la fila de 40 sesiones de
// normativa-testigo-en-referencia: 20 copias de cada una de las dos sesiones
// fijas, todas activan antes, la mitad abre con_herramientas.md en el turno 2 y
// cumple la regla, y ninguna abre otra referencia. Los extremos de Wilson de 40,
// 0 y 20 de 40 son los de la historia 3 del spec de H40 (SC-005 de H40; research
// V21 de H40).
const filaDeCuarentaSesionesDeNormativa = "| normativa-testigo-en-referencia | herramienta | claude-sonnet-5-5 | 2.1.317 | " +
	"0 | 40 | 40 de 40 · 100,0 % [91,2–100,0 %] | 0 de 40 · 0,0 % [0,0–8,8 %] | 0 de 40 · 0,0 % [0,0–8,8 %] | " +
	"20 de 40 · 50,0 % [35,2–64,8 %] · turno 2 (2–2) | 0 de 40 · 0,0 % [0,0–8,8 %] | 0 de 40 · 0,0 % [0,0–8,8 %] | " +
	"0 de 40 · 0,0 % [0,0–8,8 %] | 0 de 40 · 0,0 % [0,0–8,8 %] | 0 de 40 · 0,0 % [0,0–8,8 %] | " +
	"20 de 40 · 50,0 % [35,2–64,8 %] |\n"

// TestTablaDeNormativa fija la tabla de un sondeo de las variantes de normativa
// (FR-020 a FR-022 de H40; SC-005 de H40; contracts/tabla.md §1 a §4 y §7 de
// H40): sobre las sesiones fijas, su salida es, byte a byte, la tabla escrita a
// mano; sobre un sondeo de 40 sesiones, las celdas de 0, 20 y 40 de 40 llevan
// los extremos de Wilson del spec; y en una tabla de normativa, un directorio de
// una variante de H39 o de un modo que no es de la variante es un defecto: lo
// dice con la ruta en la salida de error y da 1 sin tabla, porque una tabla es
// de una sola familia (FR-010 de H40).
func TestTablaDeNormativa(t *testing.T) {
	t.Parallel()

	t.Run("sondeos-fijos", probarLosSondeosFijosDeNormativa)
	t.Run("cuarenta-sesiones", probarCuarentaSesionesDeNormativa)
	t.Run("variante-de-h39-en-el-sondeo", probarUnaVarianteDeH39EnElSondeoDeNormativa)
	t.Run("variante-de-h39-en-otro-sondeo", probarUnaVarianteDeH39EnOtroSondeo)
	t.Run("modo-que-no-es-de-la-variante", probarUnModoQueNoEsDeLaVarianteDeNormativa)
}

// probarLosSondeosFijosDeNormativa fija la salida del guion sobre las sesiones
// fijas de normativa: la de su tabla.md, byte a byte, y nada en la salida de
// error.
func probarLosSondeosFijosDeNormativa(t *testing.T) {
	t.Parallel()

	esperada, err := os.ReadFile(tablaFijaDeNormativa)
	require.NoError(t, err)

	tabla := contarElSondeo(t, sondeosFijosDeNormativa)

	require.Zerof(t, tabla.codigo, "salida de error: %s", tabla.errores)
	assert.Equal(t, string(esperada), tabla.salida, "la tabla de las sesiones fijas de normativa, byte a byte")
	assert.Empty(t, tabla.errores)
}

// probarCuarentaSesionesDeNormativa fija los extremos de Wilson de 0, 20 y 40
// de 40 en un sondeo que el test monta con 40 copias de dos sesiones fijas de
// normativa-testigo-en-referencia, y el detalle de las 40: las 20 primeras,
// copias de la que abre con_herramientas.md y cumple la regla, y las 20
// siguientes, de la que no.
func probarCuarentaSesionesDeNormativa(t *testing.T) {
	t.Parallel()

	sondeo := t.TempDir()
	escribirElSondeo(t, sondeo)

	for numero := 1; numero <= sesionesDeCadaClaseDelSondeo; numero++ {
		copiarLaSesionFijaDe(t, sondeosFijosDeNormativa, sesionFijaDeNormativaQueCumpleLaRegla, sondeo,
			copiaDeNormativa(numero))
		copiarLaSesionFijaDe(t, sondeosFijosDeNormativa, sesionFijaDeNormativaQueNoCumpleLaRegla, sondeo,
			copiaDeNormativa(sesionesDeCadaClaseDelSondeo+numero))
	}

	tabla := contarElSondeo(t, sondeo)

	require.Zerof(t, tabla.codigo, "salida de error: %s", tabla.errores)
	assert.True(t, strings.HasPrefix(tabla.salida, "# Sondeo de las variantes de normativa\n"),
		"con el título de la tabla de normativa: %s", tabla.salida)
	assert.Contains(t, tabla.salida, filaDeCuarentaSesionesDeNormativa)
	assert.Equal(t, 1, strings.Count(tabla.salida, "| normativa-testigo-en-referencia | herramienta |"), "una sola fila")
	assert.Contains(t, tabla.salida, "| "+copiaDeNormativa(1)+
		" | sí | sí | no | no | turno 2 | no | no | no | no | no | sí |\n")
	assert.Contains(t, tabla.salida, "| "+copiaDeNormativa(2*sesionesDeCadaClaseDelSondeo)+
		" | sí | sí | no | no | no | no | no | no | no | no | no |\n")
	assert.Equal(t, 2*sesionesDeCadaClaseDelSondeo,
		strings.Count(tabla.salida, "\n| normativa-testigo-en-referencia/herramienta/"),
		"una línea del detalle por sesión, con la ruta relativa al sondeo")
	assert.Empty(t, tabla.errores)
}

// probarUnaVarianteDeH39EnElSondeoDeNormativa fija que, en un sondeo cuya
// primera variante es de normativa, el directorio de una variante de H39 que la
// sigue en orden de nombre es un defecto: 1, ninguna tabla y su ruta en la
// salida de error.
func probarUnaVarianteDeH39EnElSondeoDeNormativa(t *testing.T) {
	t.Parallel()

	sondeo := t.TempDir()
	escribirElSondeo(t, sondeo)
	copiarLaSesionFijaDe(t, sondeosFijosDeNormativa, sesionFijaDeNormativaQueCumpleLaRegla, sondeo,
		sesionDeNormativa(varianteDeControlDeNormativa, ModoHerramienta, evalDeNormativaDeLaLGT, 1))
	copiarLaSesionFija(t, "h39-a/referencia-paso/orden/03-lrjsp-principio-de-legalidad-01", sondeo,
		"referencia-paso/orden/03-lrjsp-principio-de-legalidad-01")

	tabla := contarElSondeo(t, sondeo)

	assert.Equal(t, 1, tabla.codigo)
	assert.Empty(t, tabla.salida, "sin tabla")
	assert.Equal(t, PrefijoDelLaboratorio+filepath.Join(sondeo, "referencia-paso")+
		" no es ninguna variante de normativa\n", tabla.errores)
}

// probarUnaVarianteDeH39EnOtroSondeo fija que, en una tabla de normativa, la
// variante de H39 de un sondeo que sigue en orden de nombre al primero, de
// normativa, es un defecto: 1, ninguna tabla y su ruta en la salida de error.
func probarUnaVarianteDeH39EnOtroSondeo(t *testing.T) {
	t.Parallel()

	ruta := t.TempDir()

	deNormativa := filepath.Join(ruta, "a")
	require.NoError(t, os.Mkdir(deNormativa, 0o750))
	escribirElSondeo(t, deNormativa)
	copiarLaSesionFijaDe(t, sondeosFijosDeNormativa, sesionFijaDeNormativaQueCumpleLaRegla, deNormativa,
		copiaDeNormativa(1))

	deH39 := filepath.Join(ruta, "b")
	require.NoError(t, os.Mkdir(deH39, 0o750))
	escribirElSondeo(t, deH39)
	copiarLaSesionFija(t, sesionFijaQueCumpleLaRegla, deH39, "control/herramienta/03-lrjsp-principio-de-legalidad-01")

	tabla := contarElSondeo(t, ruta)

	assert.Equal(t, 1, tabla.codigo)
	assert.Empty(t, tabla.salida, "sin tabla")
	assert.Equal(t, PrefijoDelLaboratorio+filepath.Join(deH39, "control")+" no es ninguna variante de normativa\n",
		tabla.errores)
}

// probarUnModoQueNoEsDeLaVarianteDeNormativa fija que el directorio de un modo
// que no es de la variante de normativa es el defecto de H39: 1, ninguna tabla y
// su ruta en la salida de error.
func probarUnModoQueNoEsDeLaVarianteDeNormativa(t *testing.T) {
	t.Parallel()

	sondeo := t.TempDir()
	escribirElSondeo(t, sondeo)
	copiarLaSesionFijaDe(t, sondeosFijosDeNormativa,
		sesionFijaDeNormativa(varianteDeControlDeNormativa, ModoOrden, evalDeNormativaDeLaLGT), sondeo,
		sesionDeNormativa("normativa-cuerpo-corto", ModoOrden, evalDeNormativaDeLaLGT, 1))

	tabla := contarElSondeo(t, sondeo)

	assert.Equal(t, 1, tabla.codigo)
	assert.Empty(t, tabla.salida, "sin tabla")
	assert.Equal(t, PrefijoDelLaboratorio+filepath.Join(sondeo, "normativa-cuerpo-corto", "orden")+
		" no es ningún modo de la variante normativa-cuerpo-corto\n", tabla.errores)
}

// ordenQueConsulta es una orden kitlegal que consulta con una herramienta del
// registro, boe_articulo, y palabrasQueConsultan, sus palabras.
const ordenQueConsulta = "kitlegal boe articulo BOE-A-1978-31229 a140"

var palabrasQueConsultan = []string{"kitlegal", "boe", "articulo", "BOE-A-1978-31229", "a140"}

// casoDeOrdenDeBash es una orden de Bash del modo orden, sus órdenes simples y
// si es la primera herramienta de kitlegal de la sesión.
type casoDeOrdenDeBash struct {
	nombre  string
	orden   string
	ordenes [][]palabraDeSh
	cuenta  bool
}

// TestOrdenesDeBash fija cómo lee el guion de la tabla la orden de un Bash en
// el modo orden (FR-041 de H39; contracts/tabla.md §2 de H39; research D10 de
// H39): las órdenes simples en que la parte ordenesSimples, cada una con sus
// palabras, y si la cuenta esDeKitlegal como primera herramienta de kitlegal,
// que es lo que decide «Activa antes». La separan |, ||, &&, ;, & y el salto de
// línea, pero no el & de 2>&1 ni lo escapado; la consulta es la de un verbo que
// da herramienta, sin --describe, --help ni --dry-run; kitlegal dentro de una
// subshell, de $( ) o de comillas invertidas no cuenta (alternativa c); un # es
// un carácter más; y unas comillas, un paréntesis o unas comillas invertidas
// sin cerrar llegan hasta el final de la orden.
func TestOrdenesDeBash(t *testing.T) {
	t.Parallel()

	herramientas, err := herramientasDelRegistro()
	require.NoError(t, err)

	casos := slices.Concat(casosDeOrdenesQueSeParten(), casosDeOrdenesQueNoConsultan(), casosDeOrdenesSinPartir(),
		casosDeOrdenesConEscapes())

	for _, caso := range casos {
		t.Run(caso.nombre, func(t *testing.T) {
			t.Parallel()

			// Un intérprete por caso: su analizador guarda lo que lee de cada
			// invocación, y los casos corren a la vez.
			interprete, err := nuevoInterprete()
			require.NoError(t, err)

			reconocedor := reconocedorDeKitlegal{herramientas: herramientas, interprete: interprete}

			entrada, err := json.Marshal(map[string]string{"command": caso.orden})
			require.NoError(t, err)

			uso := usoDelTranscript{turno: 1, nombre: herramientaBash, entrada: entrada}

			assert.Equal(t, caso.ordenes, ordenesSimples(caso.orden), "las órdenes simples de %q", caso.orden)
			assert.Equal(t, caso.cuenta, reconocedor.esDeKitlegal(uso, ModoOrden), "si %q cuenta", caso.orden)
		})
	}
}

// casosDeOrdenesQueSeParten son órdenes que se parten en órdenes simples y
// cuentan por la que consulta.
func casosDeOrdenesQueSeParten() []casoDeOrdenDeBash {
	return []casoDeOrdenDeBash{
		{nombre: "sola", orden: ordenQueConsulta, ordenes: [][]palabraDeSh{consultaCon()}, cuenta: true},
		{
			nombre: "tuberia", orden: "cat x | " + ordenQueConsulta,
			ordenes: [][]palabraDeSh{palabrasDe("cat", "x"), consultaCon()}, cuenta: true,
		},
		{
			nombre: "o", orden: "false || " + ordenQueConsulta,
			ordenes: [][]palabraDeSh{palabrasDe("false"), consultaCon()}, cuenta: true,
		},
		{
			nombre: "y", orden: "cd /tmp && " + ordenQueConsulta,
			ordenes: [][]palabraDeSh{palabrasDe("cd", "/tmp"), consultaCon()}, cuenta: true,
		},
		{
			nombre: "segundo-plano", orden: "sleep 1 & " + ordenQueConsulta,
			ordenes: [][]palabraDeSh{palabrasDe("sleep", "1"), consultaCon()}, cuenta: true,
		},
		{
			nombre: "punto-y-coma", orden: "echo x; " + ordenQueConsulta,
			ordenes: [][]palabraDeSh{palabrasDe("echo", "x"), consultaCon()}, cuenta: true,
		},
		{
			nombre: "salto-de-linea", orden: "echo x\n" + ordenQueConsulta,
			ordenes: [][]palabraDeSh{palabrasDe("echo", "x"), consultaCon()}, cuenta: true,
		},
		{
			nombre: "redireccion", orden: ordenQueConsulta + " 2>&1 | tail -n 40",
			ordenes: [][]palabraDeSh{consultaCon("2>&1"), palabrasDe("tail", "-n", "40")}, cuenta: true,
		},
		{
			nombre: "asignacion-y-ruta", orden: `LANG="es ES" "/usr/local/bin/kitlegal" boe articulo BOE-A-1978-31229 a140`,
			ordenes: [][]palabraDeSh{slices.Concat([]palabraDeSh{{texto: "LANG=es ES", asignacion: true}},
				palabrasDe("/usr/local/bin/kitlegal", "boe", "articulo", "BOE-A-1978-31229", "a140"))},
			cuenta: true,
		},
		{
			nombre: "comilla-sin-cerrar-en-la-linea-siguiente", orden: ordenQueConsulta + "\necho \"sin cerrar",
			ordenes: [][]palabraDeSh{consultaCon(), palabrasDe("echo", "sin cerrar")}, cuenta: true,
		},
		{
			nombre: "parentesis-entre-comillas", orden: "echo $(echo ')' ) ; " + ordenQueConsulta,
			ordenes: [][]palabraDeSh{palabrasDe("echo", "$(echo ')' )"), consultaCon()}, cuenta: true,
		},
		{
			nombre: "parentesis-escapado", orden: `echo $(echo \( ); ` + ordenQueConsulta,
			ordenes: [][]palabraDeSh{palabrasDe("echo", `$(echo \( )`), consultaCon()}, cuenta: true,
		},
	}
}

// casosDeOrdenesQueNoConsultan son órdenes kitlegal, o que lo nombran, que no
// consultan con un verbo que da herramienta.
func casosDeOrdenesQueNoConsultan() []casoDeOrdenDeBash {
	return []casoDeOrdenDeBash{
		{nombre: "describe", orden: ordenQueConsulta + " --describe", ordenes: [][]palabraDeSh{consultaCon("--describe")}},
		{nombre: "ayuda", orden: ordenQueConsulta + " --help", ordenes: [][]palabraDeSh{consultaCon("--help")}},
		{nombre: "dry-run", orden: ordenQueConsulta + " --dry-run", ordenes: [][]palabraDeSh{consultaCon("--dry-run")}},
		{nombre: "skills-list", orden: "kitlegal skills list", ordenes: [][]palabraDeSh{palabrasDe("kitlegal", "skills", "list")}},
		{nombre: "which", orden: "which kitlegal", ordenes: [][]palabraDeSh{palabrasDe("which", "kitlegal")}},
		{
			nombre: "solo-asignaciones", orden: "KITLEGAL=kitlegal",
			ordenes: [][]palabraDeSh{{{texto: "KITLEGAL=kitlegal", asignacion: true}}},
		},
		{
			nombre: "almohadilla-al-principio", orden: "# " + ordenQueConsulta,
			ordenes: [][]palabraDeSh{palabrasDe(slices.Concat([]string{"#"}, palabrasQueConsultan)...)},
		},
		{
			nombre: "almohadilla-detras", orden: "echo x # " + ordenQueConsulta,
			ordenes: [][]palabraDeSh{palabrasDe(slices.Concat([]string{"echo", "x", "#"}, palabrasQueConsultan)...)},
		},
		{nombre: "almohadilla-en-la-palabra", orden: "echo a#b", ordenes: [][]palabraDeSh{palabrasDe("echo", "a#b")}},
	}
}

// casosDeOrdenesSinPartir son órdenes con kitlegal dentro de un tramo que no se
// parte: una subshell, una sustitución de orden o unas comillas, cerradas o no.
func casosDeOrdenesSinPartir() []casoDeOrdenDeBash {
	return []casoDeOrdenDeBash{
		{
			nombre: "subshell", orden: "(" + ordenQueConsulta + ")",
			ordenes: [][]palabraDeSh{palabrasDe("(" + ordenQueConsulta + ")")},
		},
		{
			nombre: "subshell-con-secuencia", orden: "( cd /tmp && " + ordenQueConsulta + " )",
			ordenes: [][]palabraDeSh{palabrasDe("( cd /tmp && " + ordenQueConsulta + " )")},
		},
		{
			nombre: "sustitucion", orden: "echo $(" + ordenQueConsulta + ")",
			ordenes: [][]palabraDeSh{palabrasDe("echo", "$("+ordenQueConsulta+")")},
		},
		{
			nombre: "comillas-invertidas", orden: "echo `" + ordenQueConsulta + "`",
			ordenes: [][]palabraDeSh{palabrasDe("echo", "`"+ordenQueConsulta+"`")},
		},
		{
			nombre: "doble-sin-cerrar", orden: `echo "x; ` + ordenQueConsulta,
			ordenes: [][]palabraDeSh{palabrasDe("echo", "x; "+ordenQueConsulta)},
		},
		{
			nombre: "simple-sin-cerrar", orden: "echo 'x; " + ordenQueConsulta,
			ordenes: [][]palabraDeSh{palabrasDe("echo", "x; "+ordenQueConsulta)},
		},
		{
			nombre: "parentesis-sin-cerrar", orden: "echo $(cd /tmp; " + ordenQueConsulta,
			ordenes: [][]palabraDeSh{palabrasDe("echo", "$(cd /tmp; "+ordenQueConsulta)},
		},
		{
			nombre: "comilla-sin-cerrar-en-el-parentesis", orden: "echo $(echo 'x); " + ordenQueConsulta,
			ordenes: [][]palabraDeSh{palabrasDe("echo", "$(echo 'x); "+ordenQueConsulta)},
		},
	}
}

// casosDeOrdenesConEscapes son órdenes con barras invertidas, fuera y dentro de
// las comillas dobles.
func casosDeOrdenesConEscapes() []casoDeOrdenDeBash {
	return []casoDeOrdenDeBash{
		{
			nombre: "separador-escapado", orden: `echo x\;` + ordenQueConsulta,
			ordenes: [][]palabraDeSh{palabrasDe(slices.Concat([]string{"echo", "x;kitlegal"}, palabrasQueConsultan[1:])...)},
		},
		{
			nombre: "comilla-escapada-en-las-dobles", orden: `echo "x\"; ` + ordenQueConsulta + `"`,
			ordenes: [][]palabraDeSh{palabrasDe("echo", `x"; `+ordenQueConsulta)},
		},
		{
			nombre: "escapes-de-las-dobles", orden: "echo \"1\\$ 2\\` 3\\\\ 4\\q 5\\\n6\"",
			ordenes: [][]palabraDeSh{palabrasDe("echo", "1$ 2` 3\\ 4\\q 56")},
		},
		{
			nombre: "barra-al-final-de-las-dobles", orden: `echo "x\`,
			ordenes: [][]palabraDeSh{palabrasDe("echo", `x\`)},
		},
		{
			nombre: "barra-y-salto-de-linea", orden: "kitlegal boe \\\n  articulo BOE-A-1978-31229 a140",
			ordenes: [][]palabraDeSh{consultaCon()}, cuenta: true,
		},
		{
			nombre: "barra-al-final", orden: ordenQueConsulta + ` \`,
			ordenes: [][]palabraDeSh{consultaCon(`\`)}, cuenta: true,
		},
	}
}

// palabrasDe son las palabras de una orden simple, ninguna una asignación.
func palabrasDe(textos ...string) []palabraDeSh {
	palabras := make([]palabraDeSh, 0, len(textos))
	for _, texto := range textos {
		palabras = append(palabras, palabraDeSh{texto: texto})
	}

	return palabras
}

// consultaCon son las palabras de ordenQueConsulta seguidas de esas.
func consultaCon(detras ...string) []palabraDeSh {
	return palabrasDe(slices.Concat(palabrasQueConsultan, detras)...)
}

// casoDeLecturaConCd son las órdenes de los Bash de una sesión, una por turno y
// en su orden, la referencia que se mira y el turno de su primera apertura, 0 si
// no la abre.
type casoDeLecturaConCd struct {
	nombre     string
	ordenes    []string
	referencia string
	turno      int
}

// TestLecturaConCd fija cuándo una orden de Bash abre una referencia con la
// lectura con cd (FR-041; SC-006; contracts/tabla.md §2 y §5; research D16): en
// la familia de normativa, las trece órdenes de la tabla de contracts/tabla.md
// §2, orden a orden, con el turno de la apertura; y en la de H39, que se cuenta
// con su contrato, el cd a las referencias de su skill seguido del nombre solo no
// la abre, y su ruta entera, sí.
func TestLecturaConCd(t *testing.T) {
	t.Parallel()

	declarada, err := leerLaDeclaracionDelRepositorio(raizDelRepositorio)
	require.NoError(t, err)

	normativa := familiasDelProducto[0]
	require.Equal(t, skillDeNormativa, normativa.skill)

	familias := []struct {
		nombre  string
		familia familiaDeLaTabla
		casos   []casoDeLecturaConCd
	}{
		{
			nombre: normativa.skill,
			familia: tablaDeLaFamilia(normativa,
				map[string][]string{varianteDeControlDeNormativa: normativa.referencias}),
			casos: casosDeLecturaConCdDeNormativa(),
		},
		{nombre: "h39", familia: tablaDeH39(declarada), casos: casosDeLecturaConCdDeH39(declarada)},
	}

	for _, deLaFamilia := range familias {
		for _, caso := range deLaFamilia.casos {
			t.Run(deLaFamilia.nombre+"/"+caso.nombre, func(t *testing.T) {
				t.Parallel()

				posicion := slices.Index(deLaFamilia.familia.referencias, caso.referencia)
				require.GreaterOrEqualf(t, posicion, 0, "%s es una referencia de la familia", caso.referencia)

				contador := contadorDelLaboratorio{familia: deLaFamilia.familia}
				usos := make([]usoDelTranscript, 0, len(caso.ordenes))

				for indice, orden := range caso.ordenes {
					entrada, err := json.Marshal(map[string]string{"command": orden})
					require.NoError(t, err)

					usos = append(usos, usoDelTranscript{turno: indice + 1, nombre: herramientaBash, entrada: entrada})
				}

				var contada sesionContada

				require.NoError(t, contador.anotarLosUsos(&contada, usos, ModoHerramienta))
				assert.Equal(t, caso.turno, contada.aperturas[posicion],
					"el turno de la primera apertura de %s con %q", caso.referencia, caso.ordenes)
			})
		}
	}
}

// casosDeLecturaConCdDeNormativa son las trece órdenes de la tabla de
// contracts/tabla.md §2, en su orden, con normas.md: la quinta, el cd en un Bash
// y el nombre en el siguiente.
func casosDeLecturaConCdDeNormativa() []casoDeLecturaConCd {
	const (
		normas      = "normas.md"
		referencias = "/s/claude/skills/normativa/references"
	)

	casos := []casoDeLecturaConCd{
		{nombre: "cd-y-cat", ordenes: []string{"cd " + referencias + " && cat normas.md"}, turno: 1},
		{nombre: "barra-final-y-dobles", ordenes: []string{"cd " + referencias + `/ && cat "normas.md"`}, turno: 1},
		{
			nombre:  "variable-entre-dobles-y-simples",
			ordenes: []string{`cd "$HOME/x/normativa/references"; cat 'normas.md'`}, turno: 1,
		},
		{nombre: "head-y-punto", ordenes: []string{"cd " + referencias + " && head -n 40 ./normas.md"}, turno: 1},
		{nombre: "cd-en-otro-bash", ordenes: []string{"cd " + referencias, "cat normas.md"}},
		{
			nombre:  "skill-de-otra-familia",
			ordenes: []string{"cd /s/claude/skills/jurisprudencia/references && cat normas.md"},
		},
		{nombre: "directorio-de-la-skill", ordenes: []string{"cd /s/claude/skills/normativa && cat normas.md"}},
		{nombre: "nombre-antes-del-cd", ordenes: []string{"cat normas.md && cd " + referencias}},
		{nombre: "otra-carpeta", ordenes: []string{"cd " + referencias + " && cat otra/normas.md"}},
		{nombre: "ruta-absoluta", ordenes: []string{"cd " + referencias + " && cat /x/normas.md"}},
		{nombre: "prefijo", ordenes: []string{"cd " + referencias + " && cat mis-normas.md"}},
		{nombre: "sufijo", ordenes: []string{"cd " + referencias + " && cat normas.md.bak"}},
		{nombre: "ruta-entera-de-h40", ordenes: []string{"cat " + referencias + "/normas.md"}, turno: 1},
	}

	for posicion := range casos {
		casos[posicion].referencia = normas
	}

	return casos
}

// casosDeLecturaConCdDeH39 son, con la referencia de la declaración del
// laboratorio de H39, el cd a las referencias de su skill seguido del nombre
// solo, que no la abre (contracts/tabla.md §2), y su ruta entera, que sí y deja
// ver que la familia cuenta esa referencia.
func casosDeLecturaConCdDeH39(declarada declaracionDelLaboratorio) []casoDeLecturaConCd {
	referencias := path.Join("/s/claude/skills", declarada.Skill, carpetaDeLasReferenciasDeLaSkill)

	return []casoDeLecturaConCd{
		{
			nombre: "cd-y-cat", referencia: declarada.Referencia,
			ordenes: []string{"cd " + referencias + " && cat " + declarada.Referencia},
		},
		{
			nombre: "ruta-entera", referencia: declarada.Referencia,
			ordenes: []string{"cat " + path.Join(referencias, declarada.Referencia)}, turno: 1,
		},
	}
}

// Las sesiones fijas de las tres familias del producto, cada una en
// familias/<skill>/sondeo/ con la tabla que el guion da de ellas en
// familias/<skill>/tabla.md, escrita a mano a partir de contracts/tabla.md §2 y
// §3 y de esas sesiones, con los extremos de Wilson de H40 (FR-032, FR-040,
// FR-041; SC-006, SC-007; contracts/tabla.md §5): los dos modos; en
// plazos-y-territorio, las cinco variantes; la variante del producto, con la
// lista referencias_del_producto de su sondeo.json, y en normativa una
// referencia de la lista que la familia no declara, nueva.md, vacía en las
// variantes versionadas; lecturas con cd que cuentan —con comillas dobles y
// simples, con ./, con head, con barra final, con una variable, dos
// referencias en una orden y una tubería detrás— y que no —el cd en otro Bash,
// el nombre antes del cd, un nombre con prefijo o con sufijo, el cd al
// directorio de la skill y el de las referencias de otra skill—; una referencia
// abierta dos veces, por tramos, y Grep y Glob, que no abren nada; Skill
// después de la primera herramienta y en su mismo turno; filas de dos sesiones,
// de dos sondeos y de dos versiones de Claude Code; una sesión sin terminar por
// el tope en cada familia; y la regla testigo, solo en la de normativa.
//
// Y familias/mezcla/sondeo/h41/, sin tabla.md: un sondeo cuya primera variante,
// en orden de nombre, es de jurisprudencia, con un directorio de una variante de
// normativa (plan.md «Gates»).
const (
	familiasFijas        = "testdata/laboratorio/familias"
	sondeosDeLaFamilia   = "sondeo"
	tablaDeLaFamiliaFija = "tabla.md"
	familiaFijaDeMezcla  = "mezcla"
	sondeoFijoDeLaMezcla = "h41"
)

// familiasConTablaFija son las tres familias del producto cuyas sesiones fijas
// tienen su tabla, nombradas aquí y no sacadas de leer familias/.
var familiasConTablaFija = []string{skillDeNormativa, skillDeLaFamiliaDePlazos, skillDeLaFamiliaDeJurisprudencia}

// TestTablaDeLasFamilias fija la tabla de un sondeo de cada familia del producto
// (FR-032, FR-040, FR-041; SC-006, SC-007; contracts/tabla.md §1 a §3 y §5): sobre
// las sesiones fijas de cada una, su salida es, byte a byte, su tabla.md, sin
// nada en la salida de error; y en un sondeo de jurisprudencia, el directorio de
// una variante de normativa es un defecto: lo dice con la ruta en la salida de
// error y da 1 sin tabla, porque una tabla es de una sola familia.
func TestTablaDeLasFamilias(t *testing.T) {
	t.Parallel()

	for _, familia := range familiasConTablaFija {
		t.Run(familia, func(t *testing.T) {
			t.Parallel()

			// filepath.Clean es lo que el control de rutas reconoce como saneado
			// (gosec G304): la familia es una de las de familiasConTablaFija.
			esperada, err := os.ReadFile(filepath.Clean(filepath.Join(familiasFijas, familia, tablaDeLaFamiliaFija)))
			require.NoError(t, err)

			tabla := contarElSondeo(t, filepath.Join(familiasFijas, familia, sondeosDeLaFamilia))

			require.Zerof(t, tabla.codigo, "salida de error: %s", tabla.errores)
			assert.Equal(t, string(esperada), tabla.salida, "la tabla de las sesiones fijas de %s, byte a byte", familia)
			assert.Empty(t, tabla.errores)
		})
	}

	t.Run("variante-de-otra-familia", func(t *testing.T) {
		t.Parallel()

		sondeos := filepath.Join(familiasFijas, familiaFijaDeMezcla, sondeosDeLaFamilia)

		tabla := contarElSondeo(t, sondeos)

		assert.Equal(t, 1, tabla.codigo)
		assert.Empty(t, tabla.salida, "sin tabla")
		assert.Equal(t, PrefijoDelLaboratorio+filepath.Join(sondeos, sondeoFijoDeLaMezcla, varianteDeControlDeNormativa)+
			" no es ninguna variante de "+skillDeLaFamiliaDeJurisprudencia+"\n", tabla.errores)
	})
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
// fijas de H39, en la ruta dada del sondeo, y devuelve su directorio.
func copiarLaSesionFija(t *testing.T, fija, sondeo, ruta string) string {
	t.Helper()

	return copiarLaSesionFijaDe(t, sondeosFijosDelLaboratorio, fija, sondeo, ruta)
}

// copiarLaSesionFijaDe copia la sesión fija de esa ruta, relativa al directorio
// de sesiones fijas dado, en la ruta dada del sondeo, y devuelve su directorio.
func copiarLaSesionFijaDe(t *testing.T, fijas, fija, sondeo, ruta string) string {
	t.Helper()

	destino := filepath.Join(sondeo, filepath.FromSlash(ruta))
	require.NoError(t, os.MkdirAll(filepath.Dir(destino), 0o750))
	require.NoError(t, os.CopyFS(destino, os.DirFS(filepath.Join(fijas, filepath.FromSlash(fija)))))

	return destino
}
