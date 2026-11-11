package evals

import (
	"encoding/json"
	"fmt"
	"slices"
	"strings"
)

// Activacion es un bloque tool_use de Skill de una sesión (FR-001; data-model
// §1; contracts/informe.md §1): su input.skill tal cual y si va antes de la
// primera herramienta de kitlegal de la sesión.
type Activacion struct {
	Skill string `json:"skill"`
	Antes bool   `json:"antes"`
}

// usoDelTranscript es un bloque tool_use del transcript de una sesión con el
// turno de la respuesta que lo emite (research D10 de H39).
type usoDelTranscript struct {
	turno   int
	nombre  string
	entrada json.RawMessage
}

// activacionesDeLosUsos son las skills que activan los usos de una sesión, en
// su orden, cada una con si va antes de la primera herramienta de kitlegal, la
// del primer uso que esDeKitlegal reconoce (FR-001; research D1): en un turno
// anterior al suyo o, si no hay ninguna, en cualquiera; en el mismo turno, no.
// Cada uso de Skill es una entrada, también si repite skill, y sin ninguno la
// lista es vacía, nunca nil.
func activacionesDeLosUsos(usos []usoDelTranscript, esDeKitlegal func(usoDelTranscript) bool) ([]Activacion, error) {
	primera := 0

	for _, uso := range usos {
		if esDeKitlegal(uso) {
			primera = uso.turno

			break
		}
	}

	activaciones := []Activacion{}

	for _, uso := range usos {
		if uso.nombre != herramientaSkill {
			continue
		}

		skill, err := skillDelUso(uso)
		if err != nil {
			return nil, err
		}

		activaciones = append(activaciones, Activacion{Skill: skill, Antes: activaAntes(uso.turno, primera)})
	}

	return activaciones, nil
}

// activaAntes dice si una skill activada en ese turno lo está antes de la
// primera herramienta de kitlegal, la de ese turno (FR-041 de H39): en un turno
// anterior, o activada y sin ninguna herramienta. En el mismo turno no, porque
// las llamadas de una respuesta se emiten sin haber leído el resultado de las
// otras. El turno 0 es ninguno.
func activaAntes(activa, primera int) bool {
	return activa > 0 && (primera == 0 || activa < primera)
}

// skillDelUso es la input.skill de un uso de Skill, cuya entrada es {skill,
// args?} (research.md V6).
func skillDelUso(uso usoDelTranscript) (string, error) {
	var entrada struct {
		Skill string `json:"skill"`
	}

	if err := json.Unmarshal(uso.entrada, &entrada); err != nil {
		return "", fmt.Errorf("el bloque tool_use de %s del turno %d no tiene la entrada {skill, args?}: %w",
			herramientaSkill, uso.turno, err)
	}

	return entrada.Skill, nil
}

// reconocedorDeKitlegal reconoce la llamada a una herramienta de kitlegal del
// modo de una sesión (research D2): con las herramientas del registro y el
// intérprete de las invocaciones del binario. Su intérprete guarda lo que lee de
// cada invocación, así que un reconocedor no se usa desde varias gorrutinas a la
// vez.
type reconocedorDeKitlegal struct {
	herramientas []string
	interprete   interprete
}

// nuevoReconocedorDeKitlegal es el reconocedor con las herramientas del
// registro de producción y su intérprete.
func nuevoReconocedorDeKitlegal() (reconocedorDeKitlegal, error) {
	herramientas, err := herramientasDelRegistro()
	if err != nil {
		return reconocedorDeKitlegal{}, err
	}

	interprete, err := nuevoInterprete()
	if err != nil {
		return reconocedorDeKitlegal{}, err
	}

	return reconocedorDeKitlegal{herramientas: herramientas, interprete: interprete}, nil
}

// esDeKitlegal dice si el uso es la llamada a una herramienta de kitlegal del
// modo de la sesión (FR-041 de H39; research D10 de H39; research D2): en el modo
// herramienta, un uso cuyo nombre, o lo que sigue a su último «__», es una
// herramienta del registro, de modo que ToolSearch, que carga un esquema, no lo
// es; en el modo orden, una orden de Bash con una orden simple que invoca
// kitlegal con un verbo que da herramienta y que consulta; y en cualquier otro,
// como el de una eval sin binario ni servidor, ninguno.
func (r reconocedorDeKitlegal) esDeKitlegal(uso usoDelTranscript, modo Modo) bool {
	switch modo {
	case ModoHerramienta:
		nombre := uso.nombre
		if corte := strings.LastIndex(nombre, separadorDelPrefijo); corte >= 0 {
			nombre = nombre[corte+len(separadorDelPrefijo):]
		}

		return slices.Contains(r.herramientas, nombre)
	case ModoOrden:
		orden, deBash := ordenDelUso(uso)
		if !deBash {
			return false
		}

		for _, palabras := range ordenesSimples(orden) {
			if r.consultaKitlegal(palabras) {
				return true
			}
		}

		return false
	default:
		return false
	}
}

// consultaKitlegal dice si las palabras de una orden simple invocan kitlegal y
// consultan con un verbo que da herramienta: su primera palabra que no es una
// asignación NOMBRE=valor es kitlegal o una ruta que termina en /kitlegal, y su
// argv, desde ella, es para InterpretarInvocacion de un applet, consulta —sin
// ayuda, --describe ni --dry-run— y su <applet>_<verbo> es una herramienta del
// registro, así que no lo son ni which kitlegal, ni --help, ni skills list.
func (r reconocedorDeKitlegal) consultaKitlegal(palabras []palabraDeSh) bool {
	inicio := slices.IndexFunc(palabras, func(palabra palabraDeSh) bool { return !palabra.asignacion })
	if inicio < 0 {
		return false
	}

	programa := palabras[inicio].texto
	if programa != binarioMulticall && !strings.HasSuffix(programa, "/"+binarioMulticall) {
		return false
	}

	argv := make([]string, 0, len(palabras)-inicio)
	for _, palabra := range palabras[inicio:] {
		argv = append(argv, palabra.texto)
	}

	invocacion, deApplet := r.interprete.interpretar(argv)

	return deApplet && invocacion.Consulta &&
		slices.Contains(r.herramientas, invocacion.Applet+separadorDeHerramienta+invocacion.Verbo)
}

// leerLosUsos lee del transcript de la sesión sus usos de herramienta, cada uno
// con su turno (research D10 de H39): empieza un turno el primer mensaje
// assistant, y cada assistant que llega después de un user con algún
// tool_result y lleva un message.id distinto del del assistant anterior; así, los
// bloques de una misma respuesta que el transcript da en varios mensajes son del
// mismo turno aunque un resultado llegue entre ellos (research S2 de H39).
//
// LeerSesion ya ha leído el transcript, y como ella no exige el message.id de
// cada assistant: se compara tal como se lee, también vacío (research D1).
func leerLosUsos(directorio string) ([]usoDelTranscript, error) {
	contenido, err := leerFicheroDeSesion(directorio, ficheroDelTranscript)
	if err != nil {
		return nil, err
	}

	var (
		usos     []usoDelTranscript
		turno    int
		anterior string
		resuelto bool
	)

	numero := 0

	for linea := range strings.Lines(string(contenido)) {
		numero++

		var mensaje struct {
			Type    string `json:"type"`
			Message struct {
				ID      string          `json:"id"`
				Content json.RawMessage `json:"content"`
			} `json:"message"`
		}

		if err := json.Unmarshal([]byte(linea), &mensaje); err != nil {
			return nil, fmt.Errorf("%s, línea %d: no es un mensaje JSON: %w", ficheroDelTranscript, numero, err)
		}

		switch mensaje.Type {
		case mensajeAssistant:
			if turno == 0 || (resuelto && mensaje.Message.ID != anterior) {
				turno++
			}

			anterior, resuelto = mensaje.Message.ID, false

			delMensaje, err := usosDelMensaje(mensaje.Message.Content, turno)
			if err != nil {
				return nil, fmt.Errorf("%s, línea %d: %w", ficheroDelTranscript, numero, err)
			}

			usos = append(usos, delMensaje...)
		case mensajeUser:
			resuelto = resuelto || conResultado(mensaje.Message.Content)
		}
	}

	return usos, nil
}

// usosDelMensaje son los bloques tool_use del message.content de un mensaje
// assistant, del turno dado.
func usosDelMensaje(contenido json.RawMessage, turno int) ([]usoDelTranscript, error) {
	var bloques []bloqueDeContenido

	if err := json.Unmarshal(contenido, &bloques); err != nil {
		return nil, fmt.Errorf("el mensaje assistant no tiene la lista message.content: %w", err)
	}

	var usos []usoDelTranscript

	for _, bloque := range bloques {
		if bloque.Type == bloqueToolUse {
			usos = append(usos, usoDelTranscript{turno: turno, nombre: bloque.Name, entrada: bloque.Input})
		}
	}

	return usos, nil
}

// bloqueConTipo es lo que se lee de un bloque del contenido de un mensaje user:
// su type.
type bloqueConTipo struct {
	Type string `json:"type"`
}

// conResultado dice si el message.content de un mensaje user trae algún bloque
// tool_result. Como en LeerSesion, de un mensaje user no se exige más forma que
// la de la API, en la que su contenido es un texto o una lista de bloques: un
// texto no trae ninguno, y no es un defecto.
func conResultado(contenido json.RawMessage) bool {
	var bloques []bloqueConTipo

	if json.Unmarshal(contenido, &bloques) != nil {
		return false
	}

	return slices.ContainsFunc(bloques, func(bloque bloqueConTipo) bool { return bloque.Type == bloqueToolResult })
}
