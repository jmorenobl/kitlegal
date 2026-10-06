package evals

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"slices"
	"strconv"
	"strings"
	"sync"

	"github.com/jmorenobl/kitlegal/internal/app"
	"github.com/jmorenobl/kitlegal/internal/cli"
	"github.com/jmorenobl/kitlegal/internal/core/schema"
)

// Ficheros que el guion de evals escribe siempre en el directorio de una sesión
// y que lee LeerSesion (contrato job-de-evals §3.2 y §4).
const (
	ficheroDelTranscript   = "sesion.jsonl"
	ficheroDelCodigo       = "codigo-de-la-sesion"
	ficheroDeSalidaDeError = "sesion.err"
)

// Códigos con los que timeout termina una sesión al agotar el tope: 124 si
// bastó con TERM y 137 si, pasados los 10 s de --kill-after, tuvo que enviar
// KILL. Si la sesión termina antes, el código es el de claude (contrato
// job-de-evals §3.2; research.md V52 y V53).
const (
	codigoDelTope          = 124
	codigoDeKillTrasElTope = 137
)

// Motivos de texto fijo de una sesión sin terminar (data-model §10.1); los
// otros dos llevan el código o el subtype.
const (
	motivoDelTope          = "tope de 240 s agotado (código 124)"
	motivoDeKillTrasElTope = "terminada por señal tras el tope (código 137)"
	motivoSinResult        = "sin mensaje result"
	motivoResultConError   = "result con is_error"
)

// finSinMensajes es el fin de una sesión cuyo transcript no tiene ningún
// mensaje (data-model §10.1).
const finSinMensajes = "sin mensajes"

// Tipos y subtipos de los mensajes del transcript stream-json de los que se lee
// algo, y la herramienta con la que Claude Code carga una skill (research.md V6,
// V7, V8 y V47).
const (
	mensajeSystem    = "system"
	mensajeAssistant = "assistant"
	mensajeResult    = "result"
	subtipoInit      = "init"
	subtipoAPIRetry  = "api_retry"
	subtipoSuccess   = "success"
	bloqueToolUse    = "tool_use"
	herramientaSkill = "Skill"
)

// Lo que se lee de una llamada a una herramienta del servidor MCP y de su
// resultado (research.md S5 y V21 de H21): el mensaje user que trae el bloque
// tool_result, el bloque de texto de su contenido y lo que separa el nombre de
// la herramienta del prefijo que le pone el agente, que en Claude Code es
// mcp__<servidor>__.
const (
	mensajeUser         = "user"
	bloqueToolResult    = "tool_result"
	bloqueDeTexto       = "text"
	separadorDelPrefijo = "__"
)

// herramientaBash es la herramienta con la que Claude Code ejecuta una orden:
// la que nombra kitlegal da uno de los textos de la sesión
// (contracts/juez-y-voto.md §2 de H24).
const herramientaBash = "Bash"

// Sesion es lo que se lee del directorio de una sesión de evals (data-model
// §10.1; contrato job-de-evals §4): lo que declara su transcript stream-json, el
// código con el que terminó, su salida de error y, desde su traza, sus
// invocaciones.
type Sesion struct {
	// Modelo es el campo model del mensaje system/init del transcript; vacío si
	// el transcript no tiene ese mensaje.
	Modelo string

	// VersionDeClaudeCode es el campo claude_code_version del mismo mensaje.
	VersionDeClaudeCode string

	// SkillsActivadas son las input.skill de los bloques tool_use de la
	// herramienta Skill, en el orden del transcript. Activada dice si una skill
	// está entre ellas.
	SkillsActivadas []string

	// Respuesta es el campo result del primer mensaje result del transcript si
	// ese mensaje tiene subtype success e is_error falso; vacía en otro caso. Es
	// la respuesta a la pregunta: el aviso de una tarea en segundo plano abre
	// otro turno, con su propio result, que no la sustituye (research.md D12 y S5
	// de H7.4).
	Respuesta string

	// Codigo es el entero de codigo-de-la-sesion. Nunca es 0 por omisión: si el
	// fichero falta o no es un entero, LeerSesion devuelve un error.
	Codigo int

	// Fin es el texto fijo del último mensaje del transcript: result <subtype>,
	// seguido de « con is_error» si is_error es verdadero; el type de cualquier
	// otro mensaje; o «sin mensajes».
	Fin string

	// Terminada dice si Codigo es 0 y el último mensaje es result con subtype
	// success e is_error falso. Una eval con la sesión sin terminar no pasa.
	Terminada bool

	// Cortada dice si Codigo es 124 o 137: el tope cortó la sesión, y es lo que
	// quien lee la traza pasa a LeerTrazas (data-model §9, regla 6).
	Cortada bool

	// MotivoSinTerminar es por qué no terminó la sesión, con el primero que
	// aplica de data-model §10.1; vacío si terminó. Con ErrorDelResultado, el
	// motivo «result con is_error» y el de un código que no es del tope llevan su
	// texto: «result con is_error: <texto>» y «código <n>: result con is_error:
	// <texto>» (data-model §2 de H7.3).
	MotivoSinTerminar string

	// SalidaDeError es el contenido entero de sesion.err.
	SalidaDeError string

	// Invocaciones son las de la traza de la sesión (data-model §9). LeerSesion
	// no lee la traza y la deja vacía: la lee LeerTrazas con Cortada, y quien
	// juzga la sesión se las pone (contrato job-de-evals §3.3).
	Invocaciones []Invocacion

	// Reintentos son los mensajes system/api_retry del transcript, en su orden
	// (data-model §2 de H7.3; research.md V2 de H7.3).
	Reintentos []ReintentoDeLaAPI

	// TerminaEnReintento dice si el último mensaje del transcript es un
	// system/api_retry.
	TerminaEnReintento bool

	// ErrorDelResultado es el result del último mensaje del transcript si es un
	// result con is_error verdadero, sea cual sea el código de la sesión; vacío
	// si no lo es o si no lleva result (research.md V3 y V18 de H7.3).
	ErrorDelResultado string

	// Llamadas son las llamadas a las herramientas del servidor MCP de kitlegal,
	// en el orden de sus bloques tool_use en el transcript; ninguna si la sesión
	// no llamó a ninguna (data-model §8 de H21).
	Llamadas []Llamada

	// Textos son los textos que devolvieron las herramientas de la sesión, en el
	// orden de sus bloques tool_use en el transcript: lo que el juez con modelo
	// recibe de ella; ninguno si ninguna devolvió nada (data-model §2 de H24;
	// contracts/juez-y-voto.md §2 de H24).
	Textos []Texto
}

// Texto es el texto que devolvió una herramienta de la sesión, leído del
// transcript: el bloque tool_use de un mensaje assistant y el bloque tool_result
// con su tool_use_id de un mensaje user (data-model §2 de H24;
// contracts/juez-y-voto.md §2 de H24; FR-001 de H24). La regla es la misma en
// los dos modos y en la eval sin binario ni servidor: dan texto la orden de Bash
// que nombra kitlegal y la llamada a una herramienta del registro, también
// cuando fallan, y nada más; un tool_use sin tool_result no da ninguno.
type Texto struct {
	// Orden es, en una orden de Bash, su input.command tal cual y, en una
	// llamada a una herramienta del registro, lo que el informe publica en
	// invocaciones[].orden de esa llamada: la herramienta sin el prefijo del
	// agente, seguida de sus argumentos.
	Orden string

	// Salida son los textos del content de su tool_result, unidos por un salto
	// de línea; vacía si no trae ninguno.
	Salida string
}

// Llamada es una llamada a una herramienta del servidor MCP de kitlegal leída
// del transcript de una sesión: el bloque tool_use de un mensaje assistant y,
// si el transcript lo trae, el bloque tool_result con su tool_use_id de un
// mensaje user (data-model §8 de H21; contracts/evals-en-dos-modos.md §3 de H21;
// research.md D17 y S5 de H21).
type Llamada struct {
	// Herramienta es el name del bloque tool_use sin lo que precede a su último
	// «__»: boe_articulo, la nombre el agente mcp__kitlegal__boe_articulo o a
	// secas.
	Herramienta string

	// Argumentos es su input, tal como está en el transcript.
	Argumentos json.RawMessage

	// ConResultado dice si el transcript trae su tool_result. No lo trae la
	// llamada de una sesión que se cortó antes.
	ConResultado bool

	// Error dice si su resultado es un error: su is_error es verdadero o su
	// contenido es un sobre con ok falso. No depende solo de is_error, que es lo
	// que el agente dice del resultado y no lo que el servidor devolvió.
	Error bool

	// Clase es data.clase de ese sobre con ok falso; vacía si el resultado no lo
	// trae o si de él no se lee.
	Clase schema.Clase
}

// ReintentoDeLaAPI es un mensaje system/api_retry del transcript: Claude Code lo
// emite antes de repetir una llamada a la API que ha fallado (research.md V2 de
// H7.3).
type ReintentoDeLaAPI struct {
	// Intento es su attempt: el número de este reintento, desde 1.
	Intento int

	// Maximo es su max_retries: los reintentos que Claude Code hará como mucho.
	Maximo int

	// Error es su error: la clase del fallo, rate_limit para un 429 y
	// overloaded para un 529, entre otras.
	Error string
}

// errorDeLimiteDeRitmo es el error de un reintento por un 429 (research.md V2 de
// H7.3).
const errorDeLimiteDeRitmo = "rate_limit"

// Activada dice si la sesión activó la skill: algún bloque tool_use de la
// herramienta Skill lleva su nombre en input.skill. La activación de otra skill
// no cuenta (data-model §10.1).
func (s Sesion) Activada(skill string) bool {
	return slices.Contains(s.SkillsActivadas, skill)
}

// ReintentosPorLimiteDeRitmo son los reintentos de la sesión con error
// rate_limit (FR-033 de H7.3).
func (s Sesion) ReintentosPorLimiteDeRitmo() int {
	n := 0

	for _, reintento := range s.Reintentos {
		if reintento.Error == errorDeLimiteDeRitmo {
			n++
		}
	}

	return n
}

// LeerSesion lee del directorio de una sesión de evals sesion.jsonl,
// codigo-de-la-sesion y sesion.err, los tres ficheros que el guion escribe
// siempre, y devuelve la sesión sin sus invocaciones (data-model §10.1; contrato
// job-de-evals §3.2 y §4), con sus llamadas a las herramientas del registro de
// producción (contracts/evals-en-dos-modos.md §3 de H21) y con los textos que
// devolvieron sus herramientas (contracts/juez-y-voto.md §2 de H24).
//
// Nada que falte se toma por vacío ni por 0: un fichero ausente o que no se
// puede leer, un código que no es un entero en una línea y una línea del
// transcript que no es un mensaje de stream-json con lo que se lee de él son un
// error, y la sesión es ilegible. Los ficheros se leen en ese orden y el error,
// el del primero que falla, empieza por su nombre —el del motivo «sesión
// ilegible: <fichero>: <error>» del informe (data-model §10.2)— y nombra su ruta
// y, en el transcript, el número de la línea. Un transcript vacío no es
// ilegible: es el de una sesión que acabó, o que el tope cortó, antes de emitir
// ningún mensaje, y su fin es «sin mensajes».
func LeerSesion(dir string) (Sesion, error) {
	herramientas, err := herramientasDelRegistro()
	if err != nil {
		return Sesion{}, err
	}

	verbos, err := verbosDeLasHerramientas()
	if err != nil {
		return Sesion{}, err
	}

	leido, err := leerTranscript(dir, herramientas, verbos)
	if err != nil {
		return Sesion{}, err
	}

	codigo, err := leerCodigo(dir)
	if err != nil {
		return Sesion{}, err
	}

	salidaDeError, err := leerFicheroDeSesion(dir, ficheroDeSalidaDeError)
	if err != nil {
		return Sesion{}, err
	}

	errorDelResultado := ""
	if leido.ultimo != nil {
		errorDelResultado = leido.ultimo.errorDelResultado
	}

	motivo := motivoSinTerminar(codigo, leido.ultimo, errorDelResultado)

	return Sesion{
		Modelo:              leido.modelo,
		VersionDeClaudeCode: leido.version,
		SkillsActivadas:     leido.skills,
		Respuesta:           leido.respuesta,
		Codigo:              codigo,
		Fin:                 finDelTranscript(leido.ultimo),
		Terminada:           motivo == "",
		Cortada:             codigo == codigoDelTope || codigo == codigoDeKillTrasElTope,
		MotivoSinTerminar:   motivo,
		SalidaDeError:       string(salidaDeError),
		Reintentos:          leido.reintentos,
		TerminaEnReintento:  leido.ultimo != nil && leido.ultimo.tipo == mensajeSystem && leido.ultimo.subtipo == subtipoAPIRetry,
		ErrorDelResultado:   errorDelResultado,
		Llamadas:            leido.llamadas,
		Textos:              leido.textosDeLasHerramientas(),
	}, nil
}

// herramientasDelRegistro son los nombres de las herramientas que el servidor
// MCP anuncia con el registro de producción, los de app.NombresDeHerramientas:
// ninguna lista paralela. El registro se construye una sola vez, con la versión
// vacía, la de quien no tiene ninguna: de él solo se leen nombres, que no
// cambian mientras dura el proceso.
var herramientasDelRegistro = sync.OnceValues(func() ([]string, error) {
	registro, err := app.RegistroDeProduccion("")
	if err != nil {
		return nil, fmt.Errorf("el registro de applets del binario no se puede construir: %w", err)
	}

	return app.NombresDeHerramientas(registro), nil
})

// modoDeLaSesion es el modo que da el directorio de una sesión (data-model §6 de
// H21; contracts/evals-en-dos-modos.md §3 de H21; research.md D16 de H21):
// ModoHerramienta si tiene servidor.json, que el repartidor escribe solo en las
// sesiones de ese modo; ninguno, el valor vacío, si no lo tiene y su eval es sin
// binario ni servidor; y ModoOrden en otro caso.
//
// Un servidor.json del que no se puede saber si está no se toma por ausente: es
// un error que empieza por su nombre y nombra su ruta, como los de LeerSesion.
func modoDeLaSesion(dir string, eval Eval) (Modo, error) {
	ruta := filepath.Join(dir, ficheroDelServidor)

	switch _, err := os.Stat(ruta); {
	case err == nil:
		return ModoHerramienta, nil
	case !errors.Is(err, fs.ErrNotExist):
		return "", fmt.Errorf("%s: %s no se puede comprobar: %w", ficheroDelServidor, ruta, err)
	case eval.SinBinarioNiServidor:
		return "", nil
	default:
		return ModoOrden, nil
	}
}

// transcriptLeido es lo que LeerSesion toma de sesion.jsonl.
type transcriptLeido struct {
	// herramientas son los nombres de las herramientas del registro de
	// producción: de los bloques tool_use, solo los de una de ellas son una
	// llamada.
	herramientas []string

	// verbos da, por el nombre de cada una de esas herramientas, su verbo: con
	// él se escribe la orden de una llamada como la publica el informe.
	verbos map[string]cli.Verbo

	modelo  string
	version string

	// conInit dice si ya se leyó un mensaje system/init: si hubiera más de uno,
	// el modelo y la versión son los del primero.
	conInit bool

	skills    []string
	respuesta string

	// conResult dice si ya se leyó un mensaje result: la respuesta es la del
	// primero.
	conResult bool

	reintentos []ReintentoDeLaAPI

	llamadas []Llamada

	// llamadaDelUso da, por el id de su bloque tool_use, la posición en llamadas
	// de cada llamada: con él se reconoce su tool_result, que lo trae en
	// tool_use_id.
	llamadaDelUso map[string]int

	// textos son los usos de herramienta que dan texto —la orden de Bash que
	// nombra kitlegal y la llamada a una herramienta del registro—, en el orden
	// de sus bloques tool_use.
	textos []textoDeUso

	// textoDelUso da, por el id de su bloque tool_use, la posición en textos de
	// cada uno: con él se reconoce su tool_result, como con llamadaDelUso.
	textoDelUso map[string]int

	// ultimo es el último mensaje leído; nil si el transcript no tiene ninguno.
	ultimo *mensaje
}

// textoDeUso es el texto de un uso de herramienta que da texto: su orden desde
// que se lee su bloque tool_use y su salida desde que se lee su tool_result.
type textoDeUso struct {
	Texto

	// conResultado dice si el transcript trae su tool_result. Sin él, el uso no
	// deja ningún texto: la sesión se cortó antes de que la herramienta
	// devolviera nada.
	conResultado bool
}

// mensaje es lo que decide el fin de la sesión de un mensaje del transcript: su
// type y, si es system o result, su subtype; si es result, además su is_error
// y, con is_error verdadero, su result.
type mensaje struct {
	tipo              string
	subtipo           string
	conError          bool
	errorDelResultado string
}

// bloqueDeContenido es un bloque del message.content de un mensaje assistant,
// con lo que dice si es la llamada que carga una skill o la llamada a una
// herramienta del registro, y el id por el que se reconoce su resultado.
type bloqueDeContenido struct {
	Type  string          `json:"type"`
	ID    string          `json:"id"`
	Name  string          `json:"name"`
	Input json.RawMessage `json:"input"`
}

// leerTranscript lee sesion.jsonl línea a línea: cada una, un mensaje JSON de
// stream-json. herramientas son los nombres de las herramientas cuyas llamadas
// se leen, y verbos, el verbo de cada una.
func leerTranscript(dir string, herramientas []string, verbos map[string]cli.Verbo) (transcriptLeido, error) {
	contenido, err := leerFicheroDeSesion(dir, ficheroDelTranscript)
	if err != nil {
		return transcriptLeido{}, err
	}

	leido := transcriptLeido{herramientas: herramientas, verbos: verbos}

	numero := 0

	for linea := range strings.Lines(string(contenido)) {
		numero++

		if err := leido.leerMensaje(strings.TrimSuffix(linea, "\n")); err != nil {
			return transcriptLeido{}, fmt.Errorf("%s: %s, línea %d: %w",
				ficheroDelTranscript, filepath.Join(dir, ficheroDelTranscript), numero, err)
		}
	}

	return leido, nil
}

// leerMensaje lee un mensaje del transcript: un objeto JSON con su type y, en
// los mensajes de los que se lee algo, esos campos con su forma.
func (t *transcriptLeido) leerMensaje(texto string) error {
	var cabecera struct {
		Type string `json:"type"`
	}

	if err := json.Unmarshal([]byte(texto), &cabecera); err != nil {
		return fmt.Errorf("no es un mensaje JSON: %w", err)
	}

	if cabecera.Type == "" {
		return errors.New("el mensaje no tiene type")
	}

	leido := mensaje{tipo: cabecera.Type}

	var err error

	switch cabecera.Type {
	case mensajeSystem:
		leido, err = t.leerSystem(texto)
	case mensajeAssistant:
		err = t.leerAssistant(texto)
	case mensajeUser:
		err = t.leerUser(texto)
	case mensajeResult:
		leido, err = t.leerResult(texto)
	}

	if err != nil {
		return err
	}

	t.ultimo = &leido

	return nil
}

// leerSystem lee del mensaje system/init el modelo y la versión de Claude Code
// (research.md V7) y de cada system/api_retry, el reintento (research.md V2 de
// H7.3); de los demás mensajes system no se lee nada más que su subtype.
func (t *transcriptLeido) leerSystem(texto string) (mensaje, error) {
	var system struct {
		Subtype           string  `json:"subtype"`
		Model             *string `json:"model"`
		ClaudeCodeVersion *string `json:"claude_code_version"`
	}

	if err := json.Unmarshal([]byte(texto), &system); err != nil {
		return mensaje{}, fmt.Errorf("el mensaje system no tiene la forma de stream-json: %w", err)
	}

	leido := mensaje{tipo: mensajeSystem, subtipo: system.Subtype}

	switch {
	case system.Subtype == subtipoAPIRetry:
		if err := t.leerReintento(texto); err != nil {
			return mensaje{}, err
		}

		return leido, nil
	case system.Subtype != subtipoInit || t.conInit:
		return leido, nil
	case system.Model == nil || system.ClaudeCodeVersion == nil:
		return mensaje{}, errors.New("el mensaje system/init no tiene model y claude_code_version")
	default:
		t.modelo, t.version, t.conInit = *system.Model, *system.ClaudeCodeVersion, true

		return leido, nil
	}
}

// leerReintento anota el reintento de un mensaje system/api_retry, que tiene que
// llevar su attempt, su max_retries y su error (research.md V2 de H7.3).
func (t *transcriptLeido) leerReintento(texto string) error {
	var reintento struct {
		Attempt    *int    `json:"attempt"`
		MaxRetries *int    `json:"max_retries"`
		Error      *string `json:"error"`
	}

	if err := json.Unmarshal([]byte(texto), &reintento); err != nil {
		return fmt.Errorf("el mensaje system/%s no tiene la forma de stream-json: %w", subtipoAPIRetry, err)
	}

	if reintento.Attempt == nil || reintento.MaxRetries == nil || reintento.Error == nil {
		return fmt.Errorf("el mensaje system/%s no tiene attempt, max_retries y error", subtipoAPIRetry)
	}

	t.reintentos = append(t.reintentos,
		ReintentoDeLaAPI{Intento: *reintento.Attempt, Maximo: *reintento.MaxRetries, Error: *reintento.Error})

	return nil
}

// leerAssistant anota las skills que activa un mensaje assistant: la
// input.skill de cada bloque tool_use de la herramienta Skill, cuya entrada es
// {skill, args?} (research.md V6). Una llamada a Skill sin skill no se ignora:
// podría ser la activación que decide una eval de no activación. De los demás
// bloques tool_use anota las órdenes de Bash que nombran kitlegal y las
// llamadas a las herramientas del registro.
func (t *transcriptLeido) leerAssistant(texto string) error {
	var assistant struct {
		Message struct {
			Content []bloqueDeContenido `json:"content"`
		} `json:"message"`
	}

	if err := json.Unmarshal([]byte(texto), &assistant); err != nil {
		return fmt.Errorf("el mensaje assistant no tiene la forma de stream-json: %w", err)
	}

	if assistant.Message.Content == nil {
		return errors.New("el mensaje assistant no tiene la lista message.content")
	}

	for _, bloque := range assistant.Message.Content {
		if bloque.Type != bloqueToolUse {
			continue
		}

		if bloque.Name == herramientaBash {
			t.leerOrdenDeBash(bloque)

			continue
		}

		if bloque.Name != herramientaSkill {
			t.leerLlamada(bloque)

			continue
		}

		var entrada struct {
			Skill string `json:"skill"`
		}

		if err := json.Unmarshal(bloque.Input, &entrada); err != nil {
			return fmt.Errorf("el bloque tool_use de %s no tiene la entrada {skill, args?}: %w", herramientaSkill, err)
		}

		if entrada.Skill == "" {
			return fmt.Errorf("el bloque tool_use de %s no nombra ninguna skill en input.skill", herramientaSkill)
		}

		t.skills = append(t.skills, entrada.Skill)
	}

	return nil
}

// leerLlamada anota la llamada de un bloque tool_use si su nombre, o lo que
// sigue a su último «__», es el de una herramienta del registro, con su input
// como argumentos y todavía sin resultado (contracts/evals-en-dos-modos.md §3
// de H21). El prefijo es del agente —Claude Code antepone mcp__<servidor>__
// (research.md V21 de H21)— y no dice de quién es la herramienta: lo dice el
// registro. Un tool_use de cualquier otra herramienta no se lee.
//
// La llamada da además uno de los textos de la sesión, con la orden que el
// informe publica de ella en invocaciones[].orden: la de invocacionDeLaLlamada,
// que no depende de su resultado (contracts/juez-y-voto.md §2 de H24).
func (t *transcriptLeido) leerLlamada(bloque bloqueDeContenido) {
	herramienta := bloque.Name
	if corte := strings.LastIndex(herramienta, separadorDelPrefijo); corte >= 0 {
		herramienta = herramienta[corte+len(separadorDelPrefijo):]
	}

	if !slices.Contains(t.herramientas, herramienta) {
		return
	}

	if t.llamadaDelUso == nil {
		t.llamadaDelUso = map[string]int{}
	}

	llamada := Llamada{Herramienta: herramienta, Argumentos: bloque.Input}

	t.llamadaDelUso[bloque.ID] = len(t.llamadas)
	t.llamadas = append(t.llamadas, llamada)
	t.anotarElTexto(bloque.ID, invocacionDeLaLlamada(llamada, t.verbos[herramienta]).orden)
}

// leerOrdenDeBash anota el texto que dará un bloque tool_use de Bash si su
// input.command lleva la palabra kitlegal, con la orden tal cual y todavía sin
// salida (contracts/juez-y-voto.md §2 de H24; research D2 de H24). La palabra
// es la de contieneComoPalabra: la orden que la lleva dentro de otra, como
// «mikitlegal», no la nombra.
func (t *transcriptLeido) leerOrdenDeBash(bloque bloqueDeContenido) {
	var entrada struct {
		Command string `json:"command"`
	}

	// Un input sin la orden como texto —el de una llamada que el modelo dio sin
	// su forma, y que Bash rechaza— no lleva la palabra: no es un defecto del
	// transcript, y no da texto.
	if json.Unmarshal(bloque.Input, &entrada) != nil || !contieneComoPalabra(entrada.Command, binarioMulticall) {
		return
	}

	t.anotarElTexto(bloque.ID, entrada.Command)
}

// anotarElTexto anota, detrás de los ya leídos, el uso de herramienta con ese
// id como uno que da texto, con su orden y todavía sin salida.
func (t *transcriptLeido) anotarElTexto(uso, orden string) {
	if t.textoDelUso == nil {
		t.textoDelUso = map[string]int{}
	}

	t.textoDelUso[uso] = len(t.textos)
	t.textos = append(t.textos, textoDeUso{Texto: Texto{Orden: orden}})
}

// anotarLaSalida deja, en el texto del uso con ese id, los textos del content
// de su tool_result unidos por un salto de línea, sea o no un error
// (contracts/juez-y-voto.md §2 de H24). El resultado de un uso que no da texto
// —el de Skill, el de Read, el de un Bash que no nombra kitlegal— no se anota.
func (t *transcriptLeido) anotarLaSalida(uso string, contenido any) {
	posicion, daTexto := t.textoDelUso[uso]
	if !daTexto {
		return
	}

	texto := &t.textos[posicion]
	texto.Salida = strings.Join(textosDelContenido(contenido), "\n")
	texto.conResultado = true
}

// textosDeLasHerramientas son los textos de los usos cuyo tool_result trae el
// transcript, en el orden de sus bloques tool_use; ninguno, y no una lista
// vacía, si no hay ninguno.
func (t *transcriptLeido) textosDeLasHerramientas() []Texto {
	var textos []Texto

	for _, texto := range t.textos {
		if texto.conResultado {
			textos = append(textos, texto.Texto)
		}
	}

	return textos
}

// leerUser anota el resultado de las llamadas ya leídas: de la lista
// message.content de un mensaje user, cada bloque tool_result cuyo tool_use_id
// es el id de una de ellas (research.md S5 de H21). La llamada queda con
// resultado; es un error si el bloque lleva is_error verdadero o si su contenido
// es un sobre con ok falso, y su clase es la de ese sobre (research.md D17 de
// H21).
//
// De un mensaje user no se exige más forma que la de la API, en la que su
// contenido es un texto o una lista de bloques: el de un turno de texto y el
// resultado de otra herramienta, como Skill, no aportan nada.
//
// Antes, cada bloque tool_result deja su salida en el texto de su uso, si es de
// los que dan texto: el de una orden de Bash que nombra kitlegal o el de una de
// esas llamadas (contracts/juez-y-voto.md §2 de H24).
func (t *transcriptLeido) leerUser(texto string) error {
	var user struct {
		Message struct {
			Content any `json:"content"`
		} `json:"message"`
	}

	if err := json.Unmarshal([]byte(texto), &user); err != nil {
		return fmt.Errorf("el mensaje user no tiene la forma de stream-json: %w", err)
	}

	bloques, _ := user.Message.Content.([]any)

	for _, bloque := range bloques {
		campos, _ := bloque.(map[string]any)
		if campos["type"] != bloqueToolResult {
			continue
		}

		uso, _ := campos["tool_use_id"].(string)

		t.anotarLaSalida(uso, campos["content"])

		posicion, deUnaLlamada := t.llamadaDelUso[uso]
		if !deUnaLlamada {
			continue
		}

		conError, _ := campos["is_error"].(bool)
		clase, conSobreDeFallo := sobreDeFallo(campos["content"])

		llamada := &t.llamadas[posicion]
		llamada.ConResultado = true
		llamada.Error = conError || conSobreDeFallo
		llamada.Clase = clase
	}

	return nil
}

// sobreDeFallo dice si el contenido de un tool_result es un sobre con ok falso,
// y da su data.clase, vacía si el sobre no la lleva como un texto. El sobre es
// el primero de los textos del contenido que es un objeto JSON con ok: el
// servidor devuelve uno por llamada, de éxito o de fallo (data-model §2 de H21).
// Un texto que no lo es —el aviso del agente de que la herramienta no existe o
// de que el servidor no responde— no es un sobre.
func sobreDeFallo(contenido any) (schema.Clase, bool) {
	for _, texto := range textosDelContenido(contenido) {
		var sobre struct {
			OK   *bool `json:"ok"`
			Data any   `json:"data"`
		}

		if err := json.Unmarshal([]byte(texto), &sobre); err != nil || sobre.OK == nil {
			continue
		}

		if *sobre.OK {
			return "", false
		}

		datos, _ := sobre.Data.(map[string]any)
		clase, _ := datos["clase"].(string)

		return schema.Clase(clase), true
	}

	return "", false
}

// textosDelContenido son los textos del content de un tool_result, que en la
// API es un texto o una lista de bloques: el propio texto o, de la lista, el de
// cada bloque de texto, en su orden. Un contenido de otra forma, o ninguno, no
// tiene textos.
func textosDelContenido(contenido any) []string {
	switch contenido := contenido.(type) {
	case string:
		return []string{contenido}
	case []any:
		var textos []string

		for _, bloque := range contenido {
			campos, _ := bloque.(map[string]any)
			if campos["type"] != bloqueDeTexto {
				continue
			}

			if texto, esTexto := campos["text"].(string); esTexto {
				textos = append(textos, texto)
			}
		}

		return textos
	default:
		return nil
	}
}

// leerResult lee un mensaje result: su subtype y su is_error, que deciden el fin
// de la sesión (research.md V47), y la respuesta, que es su result solo con
// subtype success e is_error falso y vacía en otro caso. Solo el primer result
// deja la respuesta, de modo que cuenta la del turno que abre la pregunta y no
// la réplica de un turno posterior; los demás se leen igual, y el último sigue
// decidiendo el fin (research.md D12 de H7.4). Con is_error verdadero, su
// result, si lo lleva, es el texto del error (research.md V3 y V18 de H7.3).
func (t *transcriptLeido) leerResult(texto string) (mensaje, error) {
	var result struct {
		Subtype string  `json:"subtype"`
		IsError *bool   `json:"is_error"`
		Result  *string `json:"result"`
	}

	if err := json.Unmarshal([]byte(texto), &result); err != nil {
		return mensaje{}, fmt.Errorf("el mensaje result no tiene la forma de stream-json: %w", err)
	}

	if result.Subtype == "" || result.IsError == nil {
		return mensaje{}, errors.New("el mensaje result no tiene subtype e is_error")
	}

	if result.Subtype == subtipoSuccess && !*result.IsError {
		if result.Result == nil {
			return mensaje{}, errors.New("el mensaje result con subtype success e is_error falso no tiene result")
		}

		if !t.conResult {
			t.respuesta = *result.Result
		}
	}

	t.conResult = true

	leido := mensaje{tipo: mensajeResult, subtipo: result.Subtype, conError: *result.IsError}

	if leido.conError && result.Result != nil {
		leido.errorDelResultado = *result.Result
	}

	return leido, nil
}

// leerCodigo lee de codigo-de-la-sesion el código de la sesión: un entero en una
// línea, como lo escribe el guion con printf '%s\n' (contrato job-de-evals
// §3.2).
func leerCodigo(dir string) (int, error) {
	contenido, err := leerFicheroDeSesion(dir, ficheroDelCodigo)
	if err != nil {
		return 0, err
	}

	codigo, err := strconv.Atoi(strings.TrimSuffix(string(contenido), "\n"))
	if err != nil {
		return 0, fmt.Errorf("%s: %s no es un entero en una línea: %w",
			ficheroDelCodigo, filepath.Join(dir, ficheroDelCodigo), err)
	}

	return codigo, nil
}

// leerFicheroDeSesion lee entero un fichero del directorio de la sesión. Si
// falta o no se puede leer, el error empieza por su nombre y nombra su ruta.
func leerFicheroDeSesion(dir, nombre string) ([]byte, error) {
	ruta := filepath.Join(dir, nombre)

	contenido, err := leerFichero(ruta)
	if err != nil {
		return nil, fmt.Errorf("%s: %s no se puede leer: %w", nombre, ruta, err)
	}

	return contenido, nil
}

// motivoSinTerminar es por qué no terminó una sesión, con el primero que aplica
// en el orden de data-model §10.1, o vacío si terminó. Si el último mensaje es
// un result con is_error y un texto, ese texto va en el motivo del código que no
// es del tope y en el del result con is_error (data-model §2 de H7.3).
func motivoSinTerminar(codigo int, ultimo *mensaje, errorDelResultado string) string {
	conTexto := ""
	if errorDelResultado != "" {
		conTexto = motivoResultConError + ": " + errorDelResultado
	}

	switch {
	case codigo == codigoDelTope:
		return motivoDelTope
	case codigo == codigoDeKillTrasElTope:
		return motivoDeKillTrasElTope
	case codigo != 0 && conTexto != "":
		return fmt.Sprintf("código %d: %s", codigo, conTexto)
	case codigo != 0:
		return fmt.Sprintf("código %d", codigo)
	case ultimo == nil || ultimo.tipo != mensajeResult:
		return motivoSinResult
	case ultimo.subtipo != subtipoSuccess:
		return "result con subtype " + ultimo.subtipo
	case ultimo.conError && conTexto != "":
		return conTexto
	case ultimo.conError:
		return motivoResultConError
	default:
		return ""
	}
}

// finDelTranscript es el texto fijo del último mensaje del transcript
// (data-model §10.1).
func finDelTranscript(ultimo *mensaje) string {
	switch {
	case ultimo == nil:
		return finSinMensajes
	case ultimo.tipo != mensajeResult:
		return ultimo.tipo
	case ultimo.conError:
		return mensajeResult + " " + ultimo.subtipo + " con is_error"
	default:
		return mensajeResult + " " + ultimo.subtipo
	}
}
