package evals

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"slices"
	"strconv"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/jmorenobl/kitlegal/internal/core/schema"
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
// versión del init; la respuesta, solo del result success sin is_error, y la del
// primer result, la de la pregunta, aunque el aviso de una tarea en segundo
// plano abra otro turno con su réplica (contrato evals-y-juicio §6 de H7.4;
// FR-060); el código leído de codigo-de-la-sesion, nunca 0 por omisión; la
// sesión terminada, la cortada por el tope con 124 o 137 y el motivo de la que
// no terminó; el fin con su texto fijo; y los ficheros ausentes o las líneas
// ilegibles como error que los nombra, salvo el transcript vacío de una sesión
// que el tope cortó antes de su primer mensaje (contrato job-de-evals §4 y §9;
// FR-071, FR-072). La sesión que lanza una orden kitlegal en segundo plano lleva
// en sus textos esa orden con lo que Bash devolvió de ella, y las demás, que no
// ejecutan ninguna, no llevan ninguno (contracts/juez-y-voto.md §2 de H24).
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

	conUnaOrdenEnSegundoPlano := activadaYTerminada
	conUnaOrdenEnSegundoPlano.Textos = []Texto{{
		Orden:  "kitlegal boe buscar 'obligación de resolver'",
		Salida: "Command running in background with ID: tarea_sintetica_01",
	}}

	casos := []struct {
		nombre   string
		activada bool
		sesion   Sesion
		ilegible *ficheroIlegible
	}{
		{nombre: "activada", activada: true, sesion: activadaYTerminada},
		{nombre: "respuesta-antes-de-una-tarea-en-segundo-plano", activada: true, sesion: conUnaOrdenEnSegundoPlano},
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
				MotivoSinTerminar:   "result con is_error: " + textoDelError529,
				ErrorDelResultado:   textoDelError529,
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

// casosDeLeerLasLlamadas es el directorio de los casos de TestLeerLasLlamadas:
// cada subdirectorio es el de una sesión (contracts/evals-en-dos-modos.md §3 y
// §8 de H21).
const casosDeLeerLasLlamadas = "testdata/sesiones/leer-llamadas"

// Lo que declaran los transcripts sintéticos de leer-llamadas; la skill del que
// llama a territorio_resolver es skillDeTerritorio.
const (
	argumentosDelArticulo21 = `{"norma":"BOE-A-2015-10565","bloque":"a21"}`
	argumentosDelArticulo24 = `{"norma":"BOE-A-2015-10565","bloque":"a24"}`
	argumentosDelMunicipio  = `{"consulta":"Villainexistente"}`

	respuestaDelArticulo24 = "El artículo 24 de la Ley 39/2015 regula el silencio administrativo en " +
		"procedimientos iniciados a solicitud del interesado.\n\n[BOE-A-2015-10565, bloque a24]"
	respuestaSinMunicipio = "La relación de municipios no tiene ninguno que se llame Villainexistente."
	respuestaSinFuente    = "No he podido leer el artículo 21 de la Ley 39/2015: la fuente no está disponible."
	respuestaSinLlamada   = "No tengo ninguna herramienta que liste las skills instaladas."
)

// TestLeerLasLlamadas fija lo que LeerSesion toma de las llamadas a las
// herramientas del servidor MCP de kitlegal y el modo que da el directorio de
// una sesión (contracts/evals-en-dos-modos.md §3 de H21; data-model §6 y §8 de
// H21; research.md D17 y S5 de H21; FR-041, FR-042), con una sesión sintética
// por caso: la herramienta sin el prefijo que le pone el agente, y la misma
// cuando el agente no le pone ninguno; sus argumentos, que son su input; el
// resultado que es un error por su is_error y el que lo es solo porque su
// contenido es un sobre con ok falso, cada uno con la clase de su sobre, venga
// el contenido como texto o como lista de bloques; la llamada de una sesión que
// se cortó antes de su resultado; y ninguna llamada de un tool_use que no es de
// una herramienta del registro, sea la de un verbo de un applet que el servidor
// no anuncia o Bash, aunque su resultado sea un sobre. El modo es herramienta
// en las sesiones cuyo directorio tiene servidor.json y orden en las demás.
//
// Cada llamada con resultado deja además su texto, con la orden que el informe
// publica de ella y lo que devolvió, sea un sobre de éxito o de fallo; lo deja
// también la orden kitlegal de Bash de modo-orden, que no es una llamada; y no
// deja ninguno la llamada sin resultado ni la de una herramienta que no es del
// registro (contracts/juez-y-voto.md §2 de H24).
func TestLeerLasLlamadas(t *testing.T) {
	t.Parallel()

	// usoDeLaConsulta es el id del bloque tool_use de la consulta de cada
	// transcript de leer-llamadas: el primero es el de Skill.
	const usoDeLaConsulta = "toolu_sintetico_02"

	devuelto := func(caso string) string {
		return textoDelResultado(t, filepath.Join(casosDeLeerLasLlamadas, caso), usoDeLaConsulta)
	}

	conLlamada := func(skill, respuesta string, llamada Llamada, texto Texto) Sesion {
		return Sesion{
			Modelo:              modeloDeLasSesiones,
			VersionDeClaudeCode: versionDeLasSesiones,
			SkillsActivadas:     []string{skill},
			Respuesta:           respuesta,
			Fin:                 "result success",
			Terminada:           true,
			Llamadas:            []Llamada{llamada},
			Textos:              []Texto{texto},
		}
	}

	sinLlamadas := func(respuesta string, textos ...Texto) Sesion {
		return Sesion{
			Modelo:              modeloDeLasSesiones,
			VersionDeClaudeCode: versionDeLasSesiones,
			SkillsActivadas:     []string{skillDeLasSesiones},
			Respuesta:           respuesta,
			Fin:                 "result success",
			Terminada:           true,
			Textos:              textos,
		}
	}

	casos := []struct {
		nombre string
		sesion Sesion
		modo   Modo
	}{
		{
			nombre: "con-prefijo",
			sesion: conLlamada(skillDeLasSesiones, respuestaConCita, Llamada{
				Herramienta:  "boe_articulo",
				Argumentos:   json.RawMessage(argumentosDelArticulo21),
				ConResultado: true,
			}, Texto{Orden: "boe_articulo BOE-A-2015-10565 a21", Salida: devuelto("con-prefijo")}),
			modo: ModoHerramienta,
		},
		{
			nombre: "sin-prefijo",
			sesion: conLlamada(skillDeLasSesiones, respuestaDelArticulo24, Llamada{
				Herramienta:  "boe_articulo",
				Argumentos:   json.RawMessage(argumentosDelArticulo24),
				ConResultado: true,
			}, Texto{Orden: "boe_articulo BOE-A-2015-10565 a24", Salida: devuelto("sin-prefijo")}),
			modo: ModoHerramienta,
		},
		{
			nombre: "error-por-is-error",
			sesion: conLlamada(skillDeTerritorio, respuestaSinMunicipio, Llamada{
				Herramienta:  "territorio_resolver",
				Argumentos:   json.RawMessage(argumentosDelMunicipio),
				ConResultado: true,
				Error:        true,
				Clase:        schema.ClaseNoEncontrado,
			}, Texto{Orden: "territorio_resolver Villainexistente", Salida: devuelto("error-por-is-error")}),
			modo: ModoOrden,
		},
		{
			nombre: "error-por-ok-falso",
			sesion: conLlamada(skillDeLasSesiones, respuestaSinFuente, Llamada{
				Herramienta:  "boe_articulo",
				Argumentos:   json.RawMessage(argumentosDelArticulo21),
				ConResultado: true,
				Error:        true,
				Clase:        schema.ClaseFuenteNoDisponible,
			}, Texto{Orden: "boe_articulo BOE-A-2015-10565 a21", Salida: devuelto("error-por-ok-falso")}),
			modo: ModoOrden,
		},
		{
			nombre: "sin-resultado",
			sesion: Sesion{
				Modelo:              modeloDeLasSesiones,
				VersionDeClaudeCode: versionDeLasSesiones,
				SkillsActivadas:     []string{skillDeLasSesiones},
				Codigo:              124,
				Fin:                 "assistant",
				Cortada:             true,
				MotivoSinTerminar:   "tope de 240 s agotado (código 124)",
				SalidaDeError:       "salida de error sintética: la sesión se cortó a los 240 s\n",
				Llamadas: []Llamada{
					{Herramienta: "boe_articulo", Argumentos: json.RawMessage(argumentosDelArticulo21)},
				},
			},
			modo: ModoOrden,
		},
		{nombre: "herramienta-ajena", sesion: sinLlamadas(respuestaSinLlamada), modo: ModoOrden},
		{
			nombre: "modo-orden",
			sesion: sinLlamadas(respuestaConCita, Texto{
				Orden:  "kitlegal boe articulo BOE-A-2015-10565 a21 --json",
				Salida: devuelto("modo-orden"),
			}),
			modo: ModoOrden,
		},
	}

	nombres := make([]string, 0, len(casos))
	for _, caso := range casos {
		nombres = append(nombres, caso.nombre)
	}

	assert.Equal(t, directoriosDeLosCasos(t, casosDeLeerLasLlamadas), slices.Sorted(slices.Values(nombres)),
		"cada directorio de %s es un caso de TestLeerLasLlamadas, y cada caso tiene el suyo", casosDeLeerLasLlamadas)

	for _, caso := range casos {
		t.Run(caso.nombre, func(t *testing.T) {
			t.Parallel()

			dir := filepath.Join(casosDeLeerLasLlamadas, caso.nombre)

			sesion, err := LeerSesion(dir)

			require.NoError(t, err)
			assert.Equal(t, caso.sesion, sesion)

			modo, err := modoDeLaSesion(dir, Eval{})

			require.NoError(t, err)
			assert.Equal(t, caso.modo, modo, "el modo lo da servidor.json, y no las llamadas de la sesión")
		})
	}
}

// TestLeerLasLlamadasDeVariosMensajes fija, sobre transcripts que el propio test
// escribe en t.TempDir(), lo que los siete casos de TestLeerLasLlamadas, con una
// llamada cada uno, no pueden ver (contracts/evals-en-dos-modos.md §3 de H21;
// data-model §8 de H21): las llamadas quedan en el orden de sus tool_use, cada
// una con el tool_result de su tool_use_id aunque los resultados lleguen en
// otro orden o en un mismo mensaje; la clase se lee solo de un sobre con ok
// falso; un mensaje user sin bloques tool_result, el de un turno de texto, no
// aporta nada; y un resultado del que no se lee ningún sobre —un texto que no
// lo es, un bloque que no es de texto o ningún content— es un error solo por su
// is_error, sin clase, como queda sin clase el sobre de fallo que no la lleva.
func TestLeerLasLlamadasDeVariosMensajes(t *testing.T) {
	t.Parallel()

	const (
		sobreCorrecto = `{"ok":true,"fuente":"kitlegal.territorio","data":{"clase":"no-encontrado"}}`
		sobreDeLimite = `{"ok":false,"fuente":"boe.legislacion-consolidada","data":{"clase":"limite-o-tos","mensaje":"límite"}}`
		sobreSinClase = `{"ok":false,"fuente":"boe.legislacion-consolidada","data":"sin forma"}`
	)

	articulo := func(argumentos string, conResultado, conError bool, clase schema.Clase) Llamada {
		return Llamada{
			Herramienta:  "boe_articulo",
			Argumentos:   json.RawMessage(argumentos),
			ConResultado: conResultado,
			Error:        conError,
			Clase:        clase,
		}
	}

	casos := []struct {
		nombre     string
		transcript string
		llamadas   []Llamada
	}{
		{
			nombre: "resultados-en-otro-orden",
			transcript: mensajeDeLlamadas(t,
				usoDeHerramienta{id: "toolu_01", nombre: "mcp__kitlegal__boe_articulo", entrada: argumentosDelArticulo21},
				usoDeHerramienta{id: "toolu_02", nombre: "mcp__kitlegal__boe_articulo", entrada: argumentosDelArticulo24},
				usoDeHerramienta{id: "toolu_03", nombre: "mcp__kitlegal__territorio_resolver", entrada: argumentosDelMunicipio},
			) +
				mensajeDeResultados(t, resultadoDeHerramienta{id: "toolu_03", contenido: cadenaJSON(t, sobreCorrecto)}) +
				mensajeDeResultados(t, resultadoDeHerramienta{id: "toolu_01", contenido: cadenaJSON(t, sobreDeLimite)}),
			llamadas: []Llamada{
				articulo(argumentosDelArticulo21, true, true, schema.ClaseLimiteOTos),
				articulo(argumentosDelArticulo24, false, false, ""),
				{Herramienta: "territorio_resolver", Argumentos: json.RawMessage(argumentosDelMunicipio), ConResultado: true},
			},
		},
		{
			nombre: "resultados-en-un-mismo-mensaje",
			transcript: mensajeDeLlamadas(t,
				usoDeHerramienta{id: "toolu_01", nombre: "boe_articulo", entrada: argumentosDelArticulo21},
				usoDeHerramienta{id: "toolu_02", nombre: "boe_articulo", entrada: argumentosDelArticulo24},
			) +
				mensajeDeResultados(t,
					resultadoDeHerramienta{id: "toolu_02", contenido: cadenaJSON(t, sobreDeLimite)},
					resultadoDeHerramienta{id: "toolu_01", contenido: cadenaJSON(t, sobreCorrecto)},
				),
			llamadas: []Llamada{
				articulo(argumentosDelArticulo21, true, false, ""),
				articulo(argumentosDelArticulo24, true, true, schema.ClaseLimiteOTos),
			},
		},
		{
			nombre: "turno-de-texto",
			transcript: mensajeDeLlamadas(t,
				usoDeHerramienta{id: "toolu_01", nombre: "boe_articulo", entrada: argumentosDelArticulo21}) +
				`{"type":"user","message":{"role":"user","content":"¿Y el artículo 24?"}}` + "\n" +
				`{"type":"user","message":{"role":"user","content":[{"type":"text","text":"¿Y el 30?"}]}}` + "\n",
			llamadas: []Llamada{articulo(argumentosDelArticulo21, false, false, "")},
		},
		{
			nombre: "error-sin-sobre",
			transcript: mensajeDeLlamadas(t,
				usoDeHerramienta{id: "toolu_01", nombre: "boe_articulo", entrada: argumentosDelArticulo21},
				usoDeHerramienta{id: "toolu_02", nombre: "boe_articulo", entrada: argumentosDelArticulo24},
				usoDeHerramienta{id: "toolu_03", nombre: "territorio_resolver", entrada: argumentosDelMunicipio},
			) +
				mensajeDeResultados(t,
					resultadoDeHerramienta{id: "toolu_01", contenido: `"MCP error -32000: Connection closed"`, conError: true},
					resultadoDeHerramienta{id: "toolu_02", contenido: `[{"type":"image","source":{"type":"base64"}}]`, conError: true},
					resultadoDeHerramienta{id: "toolu_03", conError: true},
				),
			llamadas: []Llamada{
				articulo(argumentosDelArticulo21, true, true, ""),
				articulo(argumentosDelArticulo24, true, true, ""),
				{Herramienta: "territorio_resolver", Argumentos: json.RawMessage(argumentosDelMunicipio), ConResultado: true, Error: true},
			},
		},
		{
			nombre: "resultado-sin-contenido",
			transcript: mensajeDeLlamadas(t,
				usoDeHerramienta{id: "toolu_01", nombre: "boe_articulo", entrada: argumentosDelArticulo21}) +
				mensajeDeResultados(t, resultadoDeHerramienta{id: "toolu_01"}),
			llamadas: []Llamada{articulo(argumentosDelArticulo21, true, false, "")},
		},
		{
			nombre: "sobre-de-fallo-sin-clase",
			transcript: mensajeDeLlamadas(t,
				usoDeHerramienta{id: "toolu_01", nombre: "boe_articulo", entrada: argumentosDelArticulo21}) +
				mensajeDeResultados(t, resultadoDeHerramienta{
					id: "toolu_01",
					contenido: `[{"type":"text","text":"aviso"},{"type":"text","text":"{\"aviso\":true}"},` +
						`{"type":"text","text":` + cadenaJSON(t, sobreSinClase) + `}]`,
				}),
			llamadas: []Llamada{articulo(argumentosDelArticulo21, true, true, "")},
		},
	}

	for _, caso := range casos {
		t.Run(caso.nombre, func(t *testing.T) {
			t.Parallel()

			sesion, err := LeerSesion(escribirSesion(t, mensajeInit+caso.transcript+mensajeResultCorrecto))

			require.NoError(t, err)
			assert.Equal(t, caso.llamadas, sesion.Llamadas)
		})
	}
}

// usoDeHerramienta es un bloque tool_use de un mensaje assistant: su id, el
// nombre de la herramienta y su input, que es un documento JSON.
type usoDeHerramienta struct {
	id      string
	nombre  string
	entrada string
}

// mensajeDeLlamadas es un mensaje assistant con un bloque tool_use por cada uso
// dado, en su orden, y su salto de línea.
func mensajeDeLlamadas(t *testing.T, usos ...usoDeHerramienta) string {
	t.Helper()

	bloques := make([]string, 0, len(usos))
	for _, uso := range usos {
		bloques = append(bloques, `{"type":"tool_use","id":`+cadenaJSON(t, uso.id)+
			`,"name":`+cadenaJSON(t, uso.nombre)+`,"input":`+uso.entrada+`}`)
	}

	return `{"type":"assistant","message":{"role":"assistant","content":[` + strings.Join(bloques, ",") + `]}}` + "\n"
}

// resultadoDeHerramienta es un bloque tool_result de un mensaje user: el id de
// su tool_use; su content, que es un documento JSON, y sin él el bloque no
// lleva content; y si lleva is_error verdadero, y sin él no lleva is_error.
type resultadoDeHerramienta struct {
	id        string
	contenido string
	conError  bool
}

// mensajeDeResultados es un mensaje user con un bloque tool_result por cada
// resultado dado, en su orden, y su salto de línea.
func mensajeDeResultados(t *testing.T, resultados ...resultadoDeHerramienta) string {
	t.Helper()

	bloques := make([]string, 0, len(resultados))

	for _, resultado := range resultados {
		bloque := `{"tool_use_id":` + cadenaJSON(t, resultado.id) + `,"type":"tool_result"`

		if resultado.contenido != "" {
			bloque += `,"content":` + resultado.contenido
		}

		if resultado.conError {
			bloque += `,"is_error":true`
		}

		bloques = append(bloques, bloque+`}`)
	}

	return `{"type":"user","message":{"role":"user","content":[` + strings.Join(bloques, ",") + `]}}` + "\n"
}

// textoDelResultado es el texto del bloque tool_result con ese tool_use_id en
// el transcript versionado del directorio, leído sin LeerSesion: su content, si
// es un texto, o el de su único bloque, si es una lista.
func textoDelResultado(t *testing.T, dir, uso string) string {
	t.Helper()

	for linea := range strings.Lines(contenidoDeLaSesion(t, dir, ficheroDelTranscript)) {
		if !strings.Contains(linea, `"tool_use_id":"`+uso+`"`) {
			continue
		}

		var mensaje struct {
			Message struct {
				Content []struct {
					Content any `json:"content"`
				} `json:"content"`
			} `json:"message"`
		}

		require.NoError(t, json.Unmarshal([]byte(linea), &mensaje))
		require.Len(t, mensaje.Message.Content, 1, "el mensaje lleva solo ese tool_result")

		if texto, esTexto := mensaje.Message.Content[0].Content.(string); esTexto {
			return texto
		}

		bloques, esLista := mensaje.Message.Content[0].Content.([]any)
		require.True(t, esLista, "el content de un tool_result es un texto o una lista de bloques")
		require.Len(t, bloques, 1, "el content lleva un solo bloque")

		campos, esObjeto := bloques[0].(map[string]any)
		require.True(t, esObjeto)

		texto, esTexto := campos["text"].(string)
		require.True(t, esTexto, "el bloque es de texto")

		return texto
	}

	require.Failf(t, "sin resultado", "%s no tiene ningún tool_result con el tool_use_id %s", dir, uso)

	return ""
}

// entradaDeBash es el input de un bloque tool_use de Bash con esa orden.
func entradaDeBash(t *testing.T, orden string) string {
	t.Helper()

	return `{"command":` + cadenaJSON(t, orden) + `,"description":"Consulta el BOE"}`
}

// contenidoDeTextos es el content de un tool_result que es una lista con un
// bloque de texto por cada texto dado, en su orden.
func contenidoDeTextos(t *testing.T, textos ...string) string {
	t.Helper()

	bloques := make([]string, 0, len(textos))
	for _, texto := range textos {
		bloques = append(bloques, `{"type":"text","text":`+cadenaJSON(t, texto)+`}`)
	}

	return "[" + strings.Join(bloques, ",") + "]"
}

// Lo que piden y devuelven las herramientas de los transcripts de
// TestTextosDeLaSesion: dos órdenes kitlegal de Bash, la segunda dentro de una
// orden compuesta y con el binario por su ruta, y los sobres que devuelven, con
// uno de fallo.
const (
	ordenDelIndice   = "kitlegal boe indice BOE-A-2015-10565 --json"
	ordenDelArticulo = "cd /tmp/trabajo && /usr/local/bin/kitlegal boe articulo BOE-A-2015-10565 a21 --json | head -c 4000"

	sobreDelIndice = `{"ok":true,"fuente":"boe.legislacion-consolidada","data":{"norma":"BOE-A-2015-10565",` +
		`"bloques":[{"id":"a21","titulo":"Artículo 21"}]}}`
	sobreDelArticulo = `{"ok":true,"fuente":"boe.legislacion-consolidada","data":{"norma":"BOE-A-2015-10565",` +
		`"bloque":"a21","texto":"Artículo 21. Obligación de resolver.\n1. La Administración está obligada a dictar ` +
		`resolución expresa."}}`
	sobreSinFuente = `{"ok":false,"fuente":"boe.legislacion-consolidada","data":{"clase":"fuente-no-disponible",` +
		`"mensaje":"con --offline no se pide nada a la fuente"}}`

	activacionDeLaSkill = `{"skill":"` + skillDeLasSesiones + `"}`
	argumentosDelIndice = `{"norma":"BOE-A-2015-10565"}`
)

// TestTextosDeLaSesion fija los textos de las herramientas que LeerSesion deja
// en Sesion.Textos (contracts/juez-y-voto.md §2 y §9 de H24; data-model §2 de
// H24; research D2 de H24; FR-001, FR-107), sobre transcripts que el propio test
// escribe en t.TempDir(). La regla es una por herramienta, la misma en los dos
// modos: da texto la orden de Bash que lleva la palabra kitlegal, con la orden
// tal cual y los textos del content de su tool_result unidos por un salto de
// línea, y la llamada a una herramienta del registro, con el prefijo del agente
// o sin él, con la orden que el informe publica de ella. Quedan en el orden de
// sus tool_use, lleguen sus resultados como lleguen y se mezclen o no las dos
// herramientas. La orden y la llamada que fallan dejan su texto, con el error;
// el tool_use sin tool_result, ninguno; el tool_result sin ningún texto, uno con
// la salida vacía; y no dan texto Skill, Read, otra herramienta, una que no es
// del registro ni el Bash que no nombra kitlegal en su orden —tampoco el que no
// trae ninguna—, digan lo que digan su entrada y su resultado.
func TestTextosDeLaSesion(t *testing.T) {
	t.Parallel()

	const (
		errorDeBash = "Exit code 127\n/bin/bash: line 1: kitlegal: command not found"
		skillMD     = "---\nname: boe-legislacion\n---\n\nConsulta el BOE con kitlegal boe articulo."
	)

	skill := usoDeHerramienta{id: "toolu_01", nombre: herramientaSkill, entrada: activacionDeLaSkill}
	skillLanzada := resultadoDeHerramienta{id: "toolu_01", contenido: `"Launching skill: boe-legislacion"`}

	casos := []struct {
		nombre     string
		transcript string
		textos     []Texto
	}{
		{
			nombre: "modo-orden-en-su-orden",
			transcript: mensajeDeLlamadas(t, skill) + mensajeDeResultados(t, skillLanzada) +
				mensajeDeLlamadas(t,
					usoDeHerramienta{id: "toolu_02", nombre: "Bash", entrada: entradaDeBash(t, ordenDelIndice)},
					usoDeHerramienta{id: "toolu_03", nombre: "Bash", entrada: entradaDeBash(t, ordenDelArticulo)},
				) +
				mensajeDeResultados(t, resultadoDeHerramienta{
					id:        "toolu_03",
					contenido: contenidoDeTextos(t, sobreDelArticulo, "Shell cwd was reset to /tmp/trabajo"),
				}) +
				mensajeDeResultados(t, resultadoDeHerramienta{id: "toolu_02", contenido: cadenaJSON(t, sobreDelIndice)}),
			textos: []Texto{
				{Orden: ordenDelIndice, Salida: sobreDelIndice},
				{Orden: ordenDelArticulo, Salida: sobreDelArticulo + "\nShell cwd was reset to /tmp/trabajo"},
			},
		},
		{
			nombre: "modo-herramienta-en-su-orden",
			transcript: mensajeDeLlamadas(t, skill) + mensajeDeResultados(t, skillLanzada) +
				mensajeDeLlamadas(t,
					usoDeHerramienta{id: "toolu_02", nombre: "mcp__kitlegal__boe_indice", entrada: argumentosDelIndice},
					usoDeHerramienta{id: "toolu_03", nombre: "boe_articulo", entrada: argumentosDelArticulo21},
				) +
				mensajeDeResultados(t,
					resultadoDeHerramienta{id: "toolu_03", contenido: contenidoDeTextos(t, sobreDelArticulo)},
					resultadoDeHerramienta{id: "toolu_02", contenido: cadenaJSON(t, sobreDelIndice)},
				),
			textos: []Texto{
				{Orden: "boe_indice BOE-A-2015-10565", Salida: sobreDelIndice},
				{Orden: "boe_articulo BOE-A-2015-10565 a21", Salida: sobreDelArticulo},
			},
		},
		{
			nombre: "la-orden-y-la-llamada-que-fallan",
			transcript: mensajeDeLlamadas(t,
				usoDeHerramienta{id: "toolu_01", nombre: "mcp__kitlegal__boe_indice", entrada: argumentosDelIndice}) +
				mensajeDeResultados(t, resultadoDeHerramienta{id: "toolu_01", contenido: cadenaJSON(t, sobreDelIndice)}) +
				mensajeDeLlamadas(t,
					usoDeHerramienta{id: "toolu_02", nombre: "Bash", entrada: entradaDeBash(t, ordenDelArticulo)}) +
				mensajeDeResultados(t,
					resultadoDeHerramienta{id: "toolu_02", contenido: cadenaJSON(t, errorDeBash), conError: true}) +
				mensajeDeLlamadas(t,
					usoDeHerramienta{id: "toolu_03", nombre: "mcp__kitlegal__boe_articulo", entrada: argumentosDelArticulo21}) +
				mensajeDeResultados(t,
					resultadoDeHerramienta{id: "toolu_03", contenido: cadenaJSON(t, sobreSinFuente), conError: true}),
			textos: []Texto{
				{Orden: "boe_indice BOE-A-2015-10565", Salida: sobreDelIndice},
				{Orden: ordenDelArticulo, Salida: errorDeBash},
				{Orden: "boe_articulo BOE-A-2015-10565 a21", Salida: sobreSinFuente},
			},
		},
		{
			nombre: "ni-skill-ni-read-ni-otra-herramienta-ni-bash-sin-kitlegal",
			transcript: mensajeDeLlamadas(t,
				skill,
				usoDeHerramienta{id: "toolu_02", nombre: "Read", entrada: `{"file_path":"/tmp/trabajo/.claude/skills/boe-legislacion/SKILL.md"}`},
				usoDeHerramienta{id: "toolu_03", nombre: "Bash", entrada: entradaDeBash(t, "ls -la /tmp/trabajo")},
				usoDeHerramienta{id: "toolu_04", nombre: "Bash", entrada: entradaDeBash(t, "echo kitlegales; ls mikitlegal; env | sort")},
				usoDeHerramienta{id: "toolu_05", nombre: "Grep", entrada: `{"pattern":"kitlegal","path":"."}`},
				usoDeHerramienta{id: "toolu_06", nombre: "mcp__kitlegal__skills_list", entrada: `{}`},
				usoDeHerramienta{id: "toolu_07", nombre: "Bash", entrada: `{"description":"kitlegal boe articulo"}`},
			) +
				mensajeDeResultados(t,
					resultadoDeHerramienta{id: "toolu_01", contenido: cadenaJSON(t, skillMD)},
					resultadoDeHerramienta{id: "toolu_02", contenido: cadenaJSON(t, skillMD)},
					resultadoDeHerramienta{id: "toolu_03", contenido: cadenaJSON(t, "kitlegal\nSKILL.md")},
					resultadoDeHerramienta{id: "toolu_04", contenido: cadenaJSON(t, sobreDelArticulo)},
					resultadoDeHerramienta{id: "toolu_05", contenido: cadenaJSON(t, "SKILL.md: kitlegal boe articulo")},
					resultadoDeHerramienta{id: "toolu_06", contenido: cadenaJSON(t, sobreDelIndice), conError: true},
					resultadoDeHerramienta{
						id:        "toolu_07",
						contenido: cadenaJSON(t, "InputValidationError: Bash: falta command (kitlegal)"),
						conError:  true,
					},
				),
		},
		{
			nombre: "uso-sin-resultado",
			transcript: mensajeDeLlamadas(t,
				usoDeHerramienta{id: "toolu_01", nombre: "Bash", entrada: entradaDeBash(t, ordenDelIndice)}) +
				mensajeDeResultados(t, resultadoDeHerramienta{id: "toolu_01", contenido: cadenaJSON(t, sobreDelIndice)}) +
				mensajeDeLlamadas(t,
					usoDeHerramienta{id: "toolu_02", nombre: "Bash", entrada: entradaDeBash(t, ordenDelArticulo)},
					usoDeHerramienta{id: "toolu_03", nombre: "mcp__kitlegal__boe_articulo", entrada: argumentosDelArticulo21},
				),
			textos: []Texto{{Orden: ordenDelIndice, Salida: sobreDelIndice}},
		},
		{
			nombre: "resultado-sin-ningun-texto",
			transcript: mensajeDeLlamadas(t,
				usoDeHerramienta{id: "toolu_01", nombre: "Bash", entrada: entradaDeBash(t, ordenDelIndice)},
				usoDeHerramienta{id: "toolu_02", nombre: "boe_articulo", entrada: argumentosDelArticulo21},
			) +
				mensajeDeResultados(t,
					resultadoDeHerramienta{id: "toolu_01", contenido: `""`},
					resultadoDeHerramienta{id: "toolu_02"},
				),
			textos: []Texto{{Orden: ordenDelIndice}, {Orden: "boe_articulo BOE-A-2015-10565 a21"}},
		},
		{nombre: "sin-textos"},
	}

	for _, caso := range casos {
		t.Run(caso.nombre, func(t *testing.T) {
			t.Parallel()

			sesion, err := LeerSesion(escribirSesion(t, mensajeInit+caso.transcript+mensajeResultCorrecto))

			require.NoError(t, err)
			assert.Equal(t, caso.textos, sesion.Textos)
		})
	}

	t.Run("la-orden-de-una-llamada-es-la-que-publica-el-informe", func(t *testing.T) {
		t.Parallel()

		sesion, err := LeerSesion(escribirSesion(t, mensajeInit+
			mensajeDeLlamadas(t,
				usoDeHerramienta{id: "toolu_01", nombre: "mcp__kitlegal__boe_indice", entrada: argumentosDelIndice},
				usoDeHerramienta{id: "toolu_02", nombre: "mcp__kitlegal__boe_articulo", entrada: argumentosDelArticulo21},
				usoDeHerramienta{id: "toolu_03", nombre: "mcp__kitlegal__boe_buscar", entrada: `{"terminos":["silencio","administrativo"]}`},
			)+
			mensajeDeResultados(t,
				resultadoDeHerramienta{id: "toolu_01", contenido: cadenaJSON(t, sobreDelIndice)},
				resultadoDeHerramienta{id: "toolu_02", contenido: cadenaJSON(t, sobreDelArticulo)},
				resultadoDeHerramienta{id: "toolu_03", contenido: cadenaJSON(t, sobreSinFuente), conError: true},
			)+mensajeResultCorrecto))
		require.NoError(t, err)

		var publicadas []string

		for _, invocacion := range Juzgar(Eval{}, sesion, skillDeLasSesiones).Invocaciones {
			require.True(t, invocacion.Llamada, "la sesión no tiene más invocaciones que sus llamadas")

			publicadas = append(publicadas, invocacion.Orden)
		}

		ordenes := make([]string, 0, len(sesion.Textos))
		for _, texto := range sesion.Textos {
			ordenes = append(ordenes, texto.Orden)
		}

		require.Len(t, publicadas, 3)
		assert.Equal(t, publicadas, ordenes)
	})
}

// TestModoDeLaSesion fija el modo que da el directorio de una sesión (data-model
// §6 de H21; contracts/evals-en-dos-modos.md §3 de H21; research.md D16 de H21):
// herramienta si tiene servidor.json; ninguno si no lo tiene y su eval es sin
// binario ni servidor; y orden en otro caso. Si no se puede saber si lo tiene,
// es un error que empieza por servidor.json y nombra su ruta, y no el modo orden.
func TestModoDeLaSesion(t *testing.T) {
	t.Parallel()

	sinBinarioNiServidor := Eval{SinBinarioNiServidor: true}

	casos := []struct {
		nombre string
		dir    string
		eval   Eval
		modo   Modo
	}{
		{nombre: "con servidor.json", dir: "con-prefijo", modo: ModoHerramienta},
		{nombre: "sin servidor.json", dir: "modo-orden", modo: ModoOrden},
		{nombre: "sin servidor.json y de una eval sin binario ni servidor", dir: "modo-orden", eval: sinBinarioNiServidor},
	}

	for _, caso := range casos {
		t.Run(caso.nombre, func(t *testing.T) {
			t.Parallel()

			modo, err := modoDeLaSesion(filepath.Join(casosDeLeerLasLlamadas, caso.dir), caso.eval)

			require.NoError(t, err)
			assert.Equal(t, caso.modo, modo)
		})
	}

	t.Run("sin poder saber si tiene servidor.json", func(t *testing.T) {
		t.Parallel()

		// Un fichero donde se espera el directorio de la sesión: lo que hay
		// debajo no «no existe», sino que no se puede mirar.
		dir := filepath.Join(casosDeLeerLasLlamadas, "modo-orden", "sesion.jsonl")

		modo, err := modoDeLaSesion(dir, Eval{})

		require.Error(t, err)
		assert.Empty(t, modo)
		assert.True(t, strings.HasPrefix(err.Error(), "servidor.json: "), "el error %q empieza por el fichero", err)
		require.ErrorContains(t, err, filepath.Join(dir, "servidor.json"))
	})
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
// o sin input.skill; user con un message que no es un objeto; y result sin
// subtype o sin is_error, con un campo de otro tipo o, con subtype success e
// is_error falso, sin result. La sesión es ilegible
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
			nombre:     "user-con-message-que-no-es-un-objeto",
			transcript: mensajeInit + `{"type":"user","message":"¿Qué dice el artículo 21 de la Ley 39/2015?"}` + "\n",
			linea:      2,
			motivo:     "el mensaje user no tiene la forma de stream-json",
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

// Textos de error de la API con los que Claude Code cierra una sesión en un
// result con is_error: el de la sobrecarga del transcript versionado de
// result-con-is-error y el de una credencial que no sirve (research.md V18).
const (
	textoDelError529      = `API Error: 529 {"type":"error","error":{"type":"overloaded_error","message":"Overloaded"}}`
	textoDeLaCredencial   = "Failed to authenticate. API Error: 401 OAuth access token is invalid."
	mensajeResultCorrecto = `{"type":"result","subtype":"success","is_error":false,"result":"Hecho."}` + "\n"
)

// TestLeerSesionConReintentos fija, sobre transcripts que el propio test escribe
// en t.TempDir(), lo que LeerSesion toma de los límites (data-model §2; research
// D6, V2, V3 y V18; FR-033, FR-040, FR-063, FR-065): cada mensaje
// system/api_retry, en su orden, con su attempt, su max_retries y su error, y
// los que son de rate_limit; si el último mensaje es uno de ellos; y el texto de
// un último result con is_error, en ErrorDelResultado y en el motivo sin
// terminar, sea cual sea el código de la sesión, salvo con los del tope, cuyo
// motivo no cambia. Un api_retry sin alguno de sus tres campos, o con uno de otro
// tipo, hace la sesión ilegible nombrando su línea.
func TestLeerSesionConReintentos(t *testing.T) {
	t.Parallel()

	casos := []struct {
		nombre           string
		transcript       string
		codigo           int
		sesion           Sesion
		porLimiteDeRitmo int
		linea            int
		ilegible         string
	}{
		{
			nombre: "reintentos-en-orden",
			transcript: mensajeInit + mensajeDeReintento(1, 10, 429, "rate_limit") +
				mensajeDeReintento(2, 10, 529, "overloaded") + mensajeDeReintento(3, 10, 429, "rate_limit") +
				mensajeResultCorrecto,
			sesion: Sesion{
				Modelo:              modeloDeLasSesiones,
				VersionDeClaudeCode: versionDeLasSesiones,
				Respuesta:           "Hecho.",
				Fin:                 "result success",
				Terminada:           true,
				Reintentos: []ReintentoDeLaAPI{
					{Intento: 1, Maximo: 10, Error: "rate_limit"},
					{Intento: 2, Maximo: 10, Error: "overloaded"},
					{Intento: 3, Maximo: 10, Error: "rate_limit"},
				},
			},
			porLimiteDeRitmo: 2,
		},
		{
			nombre:     "termina-en-reintento",
			transcript: mensajeInit + mensajeDeReintento(1, 10, 429, "rate_limit") + mensajeDeReintento(2, 10, 429, "rate_limit"),
			codigo:     124,
			sesion: Sesion{
				Modelo:              modeloDeLasSesiones,
				VersionDeClaudeCode: versionDeLasSesiones,
				Codigo:              124,
				Fin:                 "system",
				Cortada:             true,
				MotivoSinTerminar:   "tope de 240 s agotado (código 124)",
				Reintentos: []ReintentoDeLaAPI{
					{Intento: 1, Maximo: 10, Error: "rate_limit"},
					{Intento: 2, Maximo: 10, Error: "rate_limit"},
				},
				TerminaEnReintento: true,
			},
			porLimiteDeRitmo: 2,
		},
		{
			nombre:     "result-con-is-error-y-codigo-0",
			transcript: mensajeInit + mensajeResultConError(t, textoDelError529),
			sesion: Sesion{
				Modelo:              modeloDeLasSesiones,
				VersionDeClaudeCode: versionDeLasSesiones,
				Fin:                 "result success con is_error",
				MotivoSinTerminar:   "result con is_error: " + textoDelError529,
				ErrorDelResultado:   textoDelError529,
			},
		},
		{
			nombre:     "result-con-is-error-y-codigo-1",
			transcript: mensajeInit + mensajeResultConError(t, textoDeLaCredencial),
			codigo:     1,
			sesion: Sesion{
				Modelo:              modeloDeLasSesiones,
				VersionDeClaudeCode: versionDeLasSesiones,
				Codigo:              1,
				Fin:                 "result success con is_error",
				MotivoSinTerminar:   "código 1: result con is_error: Failed to authenticate. API Error: 401 OAuth access token is invalid.",
				ErrorDelResultado:   textoDeLaCredencial,
			},
		},
		{
			nombre:     "codigo-1-sin-result-con-is-error",
			transcript: mensajeInit + mensajeResultCorrecto,
			codigo:     1,
			sesion: Sesion{
				Modelo:              modeloDeLasSesiones,
				VersionDeClaudeCode: versionDeLasSesiones,
				Respuesta:           "Hecho.",
				Codigo:              1,
				Fin:                 "result success",
				MotivoSinTerminar:   "código 1",
			},
		},
		{
			nombre:     "result-con-is-error-y-el-tope",
			transcript: mensajeInit + mensajeResultConError(t, textoDelError529),
			codigo:     137,
			sesion: Sesion{
				Modelo:              modeloDeLasSesiones,
				VersionDeClaudeCode: versionDeLasSesiones,
				Codigo:              137,
				Fin:                 "result success con is_error",
				Cortada:             true,
				MotivoSinTerminar:   "terminada por señal tras el tope (código 137)",
				ErrorDelResultado:   textoDelError529,
			},
		},
		{
			nombre: "result-con-is-error-que-no-es-el-ultimo",
			transcript: mensajeInit + mensajeResultConError(t, textoDelError529) +
				`{"type":"system","subtype":"compact_boundary"}` + "\n",
			sesion: Sesion{
				Modelo:              modeloDeLasSesiones,
				VersionDeClaudeCode: versionDeLasSesiones,
				Fin:                 "system",
				MotivoSinTerminar:   "sin mensaje result",
			},
		},
		{
			nombre:     "result-con-is-error-sin-result",
			transcript: mensajeInit + `{"type":"result","subtype":"error_during_execution","is_error":true}` + "\n",
			codigo:     1,
			sesion: Sesion{
				Modelo:              modeloDeLasSesiones,
				VersionDeClaudeCode: versionDeLasSesiones,
				Codigo:              1,
				Fin:                 "result error_during_execution con is_error",
				MotivoSinTerminar:   "código 1",
			},
		},
		{
			nombre:     "reintento-sin-attempt",
			transcript: mensajeInit + `{"type":"system","subtype":"api_retry","max_retries":10,"error":"rate_limit"}` + "\n",
			linea:      2,
			ilegible:   "el mensaje system/api_retry no tiene attempt, max_retries y error",
		},
		{
			nombre: "reintento-sin-max-retries",
			transcript: mensajeInit + mensajeDeReintento(1, 10, 429, "rate_limit") +
				`{"type":"system","subtype":"api_retry","attempt":2,"error":"rate_limit"}` + "\n",
			linea:    3,
			ilegible: "el mensaje system/api_retry no tiene attempt, max_retries y error",
		},
		{
			nombre:     "reintento-sin-error",
			transcript: mensajeInit + `{"type":"system","subtype":"api_retry","attempt":1,"max_retries":10,"error_status":429}` + "\n",
			linea:      2,
			ilegible:   "el mensaje system/api_retry no tiene attempt, max_retries y error",
		},
		{
			nombre:     "reintento-con-attempt-que-no-es-un-entero",
			transcript: mensajeInit + `{"type":"system","subtype":"api_retry","attempt":"1","max_retries":10,"error":"rate_limit"}` + "\n",
			linea:      2,
			ilegible:   "el mensaje system/api_retry no tiene la forma de stream-json",
		},
	}

	for _, caso := range casos {
		t.Run(caso.nombre, func(t *testing.T) {
			t.Parallel()

			dir := escribirSesionConCodigo(t, caso.transcript, caso.codigo)

			sesion, err := LeerSesion(dir)

			if caso.ilegible == "" {
				require.NoError(t, err)
				assert.Equal(t, caso.sesion, sesion)
				assert.Equal(t, caso.porLimiteDeRitmo, sesion.ReintentosPorLimiteDeRitmo())

				return
			}

			require.Error(t, err)
			assert.Zero(t, sesion, "una sesión ilegible no devuelve ningún dato, tampoco el código 0")
			require.ErrorContains(t, err,
				fmt.Sprintf("sesion.jsonl: %s, línea %d: %s", filepath.Join(dir, "sesion.jsonl"), caso.linea, caso.ilegible))
		})
	}
}

// mensajeDeReintento es un mensaje system/api_retry con la forma que le da
// Claude Code 2.1.270 (research.md V2) y su salto de línea: el intento, el
// máximo de reintentos, el estado HTTP y la clase del error.
func mensajeDeReintento(intento, maximo, estado int, clase string) string {
	return fmt.Sprintf(`{"type":"system","subtype":"api_retry","attempt":%d,"max_retries":%d,"retry_delay_ms":%d,`+
		`"error_status":%d,"error":"%s","session_id":"00000000-0000-4000-8000-000000000020"}`+"\n",
		intento, maximo, 500*intento, estado, clase)
}

// mensajeResultConError es el result con is_error y el texto dado en result con
// que Claude Code cierra una sesión tras un error de la API (research.md V3), y
// su salto de línea.
func mensajeResultConError(t *testing.T, texto string) string {
	t.Helper()

	return `{"type":"result","subtype":"success","is_error":true,"num_turns":1,"result":` + cadenaJSON(t, texto) + `}` + "\n"
}

// escribirSesion crea en un directorio temporal del test los tres ficheros que el
// guion escribe siempre en el de una sesión: el transcript dado, el código 0 y la
// salida de error vacía. Devuelve su ruta.
func escribirSesion(t *testing.T, transcript string) string {
	t.Helper()

	return escribirSesionConCodigo(t, transcript, 0)
}

// escribirSesionConCodigo es escribirSesion con el código de la sesión dado.
func escribirSesionConCodigo(t *testing.T, transcript string, codigo int) string {
	t.Helper()

	dir := t.TempDir()

	require.NoError(t, os.WriteFile(filepath.Join(dir, "sesion.jsonl"), []byte(transcript), 0o600))
	require.NoError(t, os.WriteFile(filepath.Join(dir, "codigo-de-la-sesion"), []byte(strconv.Itoa(codigo)+"\n"), 0o600))
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
