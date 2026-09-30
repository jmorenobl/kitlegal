package evals

import (
	"encoding/json"
	"errors"
	"fmt"
	"path/filepath"
	"slices"
	"strconv"
	"strings"
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
// job-de-evals §3.2 y §4).
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
	leido, err := leerTranscript(dir)
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
	}, nil
}

// transcriptLeido es lo que LeerSesion toma de sesion.jsonl.
type transcriptLeido struct {
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

	// ultimo es el último mensaje leído; nil si el transcript no tiene ninguno.
	ultimo *mensaje
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
// con lo que dice si es la llamada que carga una skill.
type bloqueDeContenido struct {
	Type  string          `json:"type"`
	Name  string          `json:"name"`
	Input json.RawMessage `json:"input"`
}

// leerTranscript lee sesion.jsonl línea a línea: cada una, un mensaje JSON de
// stream-json.
func leerTranscript(dir string) (transcriptLeido, error) {
	contenido, err := leerFicheroDeSesion(dir, ficheroDelTranscript)
	if err != nil {
		return transcriptLeido{}, err
	}

	var leido transcriptLeido

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
// podría ser la activación que decide una eval de no activación.
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
		if bloque.Type != bloqueToolUse || bloque.Name != herramientaSkill {
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
