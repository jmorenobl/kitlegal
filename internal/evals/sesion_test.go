package evals

import (
	"fmt"
	"os"
	"path/filepath"
	"slices"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// casosDeLeerSesion es el directorio de los casos de TestLeerSesion: cada
// subdirectorio es el de una sesión (contrato job-de-evals §9.1).
const casosDeLeerSesion = "testdata/sesiones/leer-sesion"

// Lo que declaran los transcripts sintéticos de leer-sesion.
const (
	skillDeLasSesiones   = "boe-legislacion"
	modeloDeLasSesiones  = "claude-haiku-4-5-20251001"
	versionDeLasSesiones = "2.1.270"
	respuestaConCita     = "El artículo 21 de la Ley 39/2015 regula la obligación de resolver.\n\n" +
		"[BOE-A-2015-10565, bloque a21]"

	// respuestaSinSkill es la del transcript de no-activada. Su primera palabra
	// va partida en dos literales porque misspell (locale US) la toma suelta por
	// una errata de «Recorder»; el texto es el mismo.
	respuestaSinSkill = "Rec" + "orre la lista con dos punteros, anterior y actual: en cada nodo guarda " +
		"actual.Next, haz que actual.Next apunte a anterior y avanza los dos; al terminar, anterior es la nueva cabeza."
)

// ficheroIlegible es lo que el error de una sesión ilegible tiene que nombrar:
// el fichero de la sesión por el que empieza, un fragmento del motivo y, si el
// defecto es de una línea, su número.
type ficheroIlegible struct {
	fichero string
	motivo  string
	linea   string
}

// TestLeerSesion fija la lectura del directorio de una sesión de data-model
// §10.1: la activación por el tool_use Skill de la skill, y no de otra; modelo y
// versión del init; la respuesta, solo del result success sin is_error; el
// código leído de codigo-de-la-sesion, nunca 0 por omisión; la sesión terminada,
// la cortada por el tope con 124 o 137 y el motivo de la que no terminó; el fin
// con su texto fijo; y los ficheros ausentes o las líneas ilegibles como error
// que los nombra, salvo el transcript vacío de una sesión que el tope cortó
// antes de su primer mensaje (contrato job-de-evals §4 y §9; FR-071, FR-072).
func TestLeerSesion(t *testing.T) {
	t.Parallel()

	activadaYTerminada := Sesion{
		Modelo:              modeloDeLasSesiones,
		VersionDeClaudeCode: versionDeLasSesiones,
		SkillsActivadas:     []string{skillDeLasSesiones},
		Respuesta:           respuestaConCita,
		Fin:                 "result success",
		Terminada:           true,
	}

	casos := []struct {
		nombre   string
		activada bool
		sesion   Sesion
		ilegible *ficheroIlegible
	}{
		{nombre: "activada", activada: true, sesion: activadaYTerminada},
		{
			nombre: "no-activada",
			sesion: Sesion{
				Modelo:              modeloDeLasSesiones,
				VersionDeClaudeCode: versionDeLasSesiones,
				Respuesta:           respuestaSinSkill,
				Fin:                 "result success",
				Terminada:           true,
			},
		},
		{
			nombre: "otra-skill-activada",
			sesion: Sesion{
				Modelo:              modeloDeLasSesiones,
				VersionDeClaudeCode: versionDeLasSesiones,
				SkillsActivadas:     []string{"boe-fiscal"},
				Respuesta:           "He cargado la skill boe-fiscal para responder.",
				Fin:                 "result success",
				Terminada:           true,
			},
		},
		{
			nombre:   "codigo-distinto-de-cero",
			activada: true,
			sesion: Sesion{
				Modelo:              modeloDeLasSesiones,
				VersionDeClaudeCode: versionDeLasSesiones,
				SkillsActivadas:     []string{skillDeLasSesiones},
				Respuesta:           respuestaConCita,
				Codigo:              1,
				Fin:                 "result success",
				MotivoSinTerminar:   "código 1",
				SalidaDeError:       "salida de error sintética: la sesión terminó con código 1\n",
			},
		},
		{
			nombre: "tope-agotado",
			sesion: Sesion{
				Modelo:              modeloDeLasSesiones,
				VersionDeClaudeCode: versionDeLasSesiones,
				Codigo:              124,
				Fin:                 "system",
				Cortada:             true,
				MotivoSinTerminar:   "tope de 240 s agotado (código 124)",
				SalidaDeError:       "salida de error sintética: la sesión se cortó a los 240 s\n",
			},
		},
		{
			nombre: "senal-tras-el-tope",
			sesion: Sesion{
				Modelo:              modeloDeLasSesiones,
				VersionDeClaudeCode: versionDeLasSesiones,
				Codigo:              137,
				Fin:                 "system",
				Cortada:             true,
				MotivoSinTerminar:   "terminada por señal tras el tope (código 137)",
				SalidaDeError:       "salida de error sintética: la sesión recibió KILL tras el tope\n",
			},
		},
		{
			nombre: "sin-result",
			sesion: Sesion{
				Modelo:              modeloDeLasSesiones,
				VersionDeClaudeCode: versionDeLasSesiones,
				Fin:                 "assistant",
				MotivoSinTerminar:   "sin mensaje result",
			},
		},
		{
			nombre:   "error-max-turns",
			activada: true,
			sesion: Sesion{
				Modelo:              modeloDeLasSesiones,
				VersionDeClaudeCode: versionDeLasSesiones,
				SkillsActivadas:     []string{skillDeLasSesiones},
				Fin:                 "result error_max_turns",
				MotivoSinTerminar:   "result con subtype error_max_turns",
			},
		},
		{
			nombre: "result-con-is-error",
			sesion: Sesion{
				Modelo:              modeloDeLasSesiones,
				VersionDeClaudeCode: versionDeLasSesiones,
				Fin:                 "result success con is_error",
				MotivoSinTerminar:   "result con is_error",
			},
		},
		{
			nombre:   "sin-fichero-de-codigo",
			ilegible: &ficheroIlegible{fichero: "codigo-de-la-sesion", motivo: "no se puede leer"},
		},
		{
			nombre:   "codigo-no-entero",
			ilegible: &ficheroIlegible{fichero: "codigo-de-la-sesion", motivo: `no es un entero en una línea: strconv.Atoi: parsing "0 segundos"`},
		},
		{
			nombre:   "sin-transcript",
			ilegible: &ficheroIlegible{fichero: "sesion.jsonl", motivo: "no se puede leer"},
		},
		{
			nombre:   "sin-salida-de-error",
			ilegible: &ficheroIlegible{fichero: "sesion.err", motivo: "no se puede leer"},
		},
		{
			nombre:   "linea-ilegible",
			ilegible: &ficheroIlegible{fichero: "sesion.jsonl", linea: "línea 3", motivo: "no es un mensaje JSON"},
		},
		{
			nombre: "sin-mensajes",
			sesion: Sesion{
				Codigo:            124,
				Fin:               "sin mensajes",
				Cortada:           true,
				MotivoSinTerminar: "tope de 240 s agotado (código 124)",
				SalidaDeError:     "salida de error sintética: la sesión se cortó a los 240 s antes de emitir ningún mensaje\n",
			},
		},
	}

	nombres := make([]string, 0, len(casos))
	for _, caso := range casos {
		nombres = append(nombres, caso.nombre)
	}

	assert.Equal(t, directoriosDeLosCasos(t, casosDeLeerSesion), slices.Sorted(slices.Values(nombres)),
		"cada directorio de %s es un caso de TestLeerSesion, y cada caso tiene el suyo", casosDeLeerSesion)

	for _, caso := range casos {
		t.Run(caso.nombre, func(t *testing.T) {
			t.Parallel()

			dir := filepath.Join(casosDeLeerSesion, caso.nombre)

			sesion, err := LeerSesion(dir)

			if caso.ilegible == nil {
				require.NoError(t, err)
				assert.Equal(t, caso.sesion, sesion)
				assert.Equal(t, caso.activada, sesion.Activada(skillDeLasSesiones))

				return
			}

			require.Error(t, err)
			assert.Zero(t, sesion, "una sesión ilegible no devuelve ningún dato, tampoco el código 0")
			assert.True(t, strings.HasPrefix(err.Error(), caso.ilegible.fichero+": "),
				"el error %q empieza por el fichero de la sesión, que da el motivo «sesión ilegible: <fichero>: …»", err)
			require.ErrorContains(t, err, filepath.Join(dir, caso.ilegible.fichero))
			require.ErrorContains(t, err, caso.ilegible.motivo)

			if caso.ilegible.linea != "" {
				require.ErrorContains(t, err, caso.ilegible.linea)
			}
		})
	}
}

// mensajeInit es el system/init de los transcripts que escriben los tests sobre
// t.TempDir(), con la forma de research.md V7 y su salto de línea.
const mensajeInit = `{"type":"system","subtype":"init","model":"` + modeloDeLasSesiones +
	`","claude_code_version":"` + versionDeLasSesiones + `"}` + "\n"

// TestLeerSesionConMensajesSinSuForma fija, sobre transcripts que el propio test
// escribe en t.TempDir() con mensajes literales, cada mensaje de stream-json del
// que LeerSesion no puede leer lo que necesita y que ningún caso de §9.1 tiene
// (data-model §10.1; contrato job-de-evals §4; FR-072): sin type; system/init sin
// model o sin claude_code_version, o con un campo de otro tipo; assistant sin la
// lista message.content o con otra cosa en ella; el tool_use de Skill sin entrada
// o sin input.skill; y result sin subtype o sin is_error, con un campo de otro
// tipo o, con subtype success e is_error falso, sin result. La sesión es ilegible
// y el error empieza por sesion.jsonl y nombra su ruta y la línea del mensaje.
func TestLeerSesionConMensajesSinSuForma(t *testing.T) {
	t.Parallel()

	casos := []struct {
		nombre     string
		transcript string
		linea      int
		motivo     string
	}{
		{
			nombre:     "sin-type",
			transcript: `{"subtype":"init","model":"` + modeloDeLasSesiones + `"}` + "\n",
			linea:      1,
			motivo:     "el mensaje no tiene type",
		},
		{
			nombre:     "init-con-model-que-no-es-una-cadena",
			transcript: `{"type":"system","subtype":"init","model":45,"claude_code_version":"` + versionDeLasSesiones + `"}` + "\n",
			linea:      1,
			motivo:     "el mensaje system no tiene la forma de stream-json",
		},
		{
			nombre:     "init-sin-model",
			transcript: `{"type":"system","subtype":"init","claude_code_version":"` + versionDeLasSesiones + `"}` + "\n",
			linea:      1,
			motivo:     "el mensaje system/init no tiene model y claude_code_version",
		},
		{
			nombre:     "init-sin-claude-code-version",
			transcript: `{"type":"system","subtype":"init","model":"` + modeloDeLasSesiones + `"}` + "\n",
			linea:      1,
			motivo:     "el mensaje system/init no tiene model y claude_code_version",
		},
		{
			nombre:     "assistant-con-content-que-no-es-una-lista",
			transcript: mensajeInit + `{"type":"assistant","message":{"role":"assistant","content":"Consulto el BOE."}}` + "\n",
			linea:      2,
			motivo:     "el mensaje assistant no tiene la forma de stream-json",
		},
		{
			nombre:     "assistant-sin-content",
			transcript: mensajeInit + `{"type":"assistant","message":{"role":"assistant"}}` + "\n",
			linea:      2,
			motivo:     "el mensaje assistant no tiene la lista message.content",
		},
		{
			nombre: "skill-sin-entrada",
			transcript: mensajeInit +
				`{"type":"assistant","message":{"content":[{"type":"tool_use","id":"toolu_01","name":"Skill"}]}}` + "\n",
			linea:  2,
			motivo: "el bloque tool_use de Skill no tiene la entrada {skill, args?}",
		},
		{
			nombre: "skill-sin-input-skill",
			transcript: mensajeInit + `{"type":"assistant","message":{"content":[{"type":"tool_use","id":"toolu_01",` +
				`"name":"Skill","input":{"args":"art. 21 de la Ley 39/2015"}}]}}` + "\n",
			linea:  2,
			motivo: "el bloque tool_use de Skill no nombra ninguna skill en input.skill",
		},
		{
			nombre:     "result-con-is-error-que-no-es-booleano",
			transcript: mensajeInit + `{"type":"result","subtype":"success","is_error":"false","result":"Hecho."}` + "\n",
			linea:      2,
			motivo:     "el mensaje result no tiene la forma de stream-json",
		},
		{
			nombre:     "result-sin-subtype",
			transcript: mensajeInit + `{"type":"result","is_error":false,"result":"Hecho."}` + "\n",
			linea:      2,
			motivo:     "el mensaje result no tiene subtype e is_error",
		},
		{
			nombre:     "result-sin-is-error",
			transcript: mensajeInit + `{"type":"result","subtype":"success","result":"Hecho."}` + "\n",
			linea:      2,
			motivo:     "el mensaje result no tiene subtype e is_error",
		},
		{
			nombre:     "result-success-sin-result",
			transcript: mensajeInit + `{"type":"result","subtype":"success","is_error":false}` + "\n",
			linea:      2,
			motivo:     "el mensaje result con subtype success e is_error falso no tiene result",
		},
	}

	for _, caso := range casos {
		t.Run(caso.nombre, func(t *testing.T) {
			t.Parallel()

			dir := escribirSesion(t, caso.transcript)

			sesion, err := LeerSesion(dir)

			require.Error(t, err)
			assert.Zero(t, sesion, "una sesión ilegible no devuelve ningún dato, tampoco el código 0")
			assert.True(t, strings.HasPrefix(err.Error(), "sesion.jsonl: "),
				"el error %q empieza por el fichero de la sesión, que da el motivo «sesión ilegible: <fichero>: …»", err)
			require.ErrorContains(t, err,
				fmt.Sprintf("%s, línea %d: %s", filepath.Join(dir, "sesion.jsonl"), caso.linea, caso.motivo))
		})
	}
}

// TestLeerSesionConVariosMensajesSystem fija, sobre un transcript que escribe el
// propio test, que de los mensajes system solo se lee el primer init: un system
// con otro subtype no aporta nada y un segundo init no cambia el modelo ni la
// versión de Claude Code de la sesión (data-model §10.1).
func TestLeerSesionConVariosMensajesSystem(t *testing.T) {
	t.Parallel()

	const respuesta = "El artículo 21 de la Ley 39/2015 regula la obligación de resolver."

	dir := escribirSesion(t, mensajeInit+
		`{"type":"system","subtype":"compact_boundary"}`+"\n"+
		`{"type":"system","subtype":"init","model":"claude-sonnet-5","claude_code_version":"2.2.0"}`+"\n"+
		`{"type":"result","subtype":"success","is_error":false,"result":"`+respuesta+`"}`+"\n")

	sesion, err := LeerSesion(dir)

	require.NoError(t, err)
	assert.Equal(t, Sesion{
		Modelo:              modeloDeLasSesiones,
		VersionDeClaudeCode: versionDeLasSesiones,
		Respuesta:           respuesta,
		Fin:                 "result success",
		Terminada:           true,
	}, sesion)
}

// escribirSesion crea en un directorio temporal del test los tres ficheros que el
// guion escribe siempre en el de una sesión: el transcript dado, el código 0 y la
// salida de error vacía. Devuelve su ruta.
func escribirSesion(t *testing.T, transcript string) string {
	t.Helper()

	dir := t.TempDir()

	require.NoError(t, os.WriteFile(filepath.Join(dir, "sesion.jsonl"), []byte(transcript), 0o600))
	require.NoError(t, os.WriteFile(filepath.Join(dir, "codigo-de-la-sesion"), []byte("0\n"), 0o600))
	require.NoError(t, os.WriteFile(filepath.Join(dir, "sesion.err"), nil, 0o600))

	return dir
}

// directoriosDeLosCasos son los nombres de los subdirectorios de dir, en orden.
func directoriosDeLosCasos(t *testing.T, dir string) []string {
	t.Helper()

	entradas, err := os.ReadDir(dir)
	require.NoError(t, err)

	nombres := make([]string, 0, len(entradas))

	for _, entrada := range entradas {
		require.True(t, entrada.IsDir(), "en %s cada caso es el directorio de una sesión", dir)

		nombres = append(nombres, entrada.Name())
	}

	return nombres
}
