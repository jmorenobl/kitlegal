package evals

import (
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
