package evals

import (
	"regexp"
	"strings"
)

// formaDeCita es la expresión con la que se extrae la parte mecánica de una
// cita: los corchetes, abiertos y cerrados en la misma línea, que terminan en
// <identificador>, bloque <id de bloque>], con el identificador y el id tal como
// los da la fuente (contrato skill-boe-legislacion §3). Delante del identificador
// puede ir, dentro de los corchetes, cualquier texto sin corchetes que no acabe
// en letra ni en cifra, como la forma legible de la norma y del bloque; si ese
// texto abre otro corchete, la cita es la del corchete interior. El corchete de
// cierre delimita el id del bloque, que puede terminar en punto, como a85bis.
var formaDeCita = regexp.MustCompile(`\[(?:[^\[\]\n]*[^\[\]\n\p{L}\p{N}])?` +
	`(BOE-A-[0-9]{4}-[0-9]{1,9}), bloque ([A-Za-z0-9][A-Za-z0-9.-]{0,63})\]`)

// Cita es la parte mecánica de una cita de la respuesta de una sesión: la norma y
// el bloque citados. Se compara con una CitaEsperada por igualdad exacta de la
// pareja (data-model §6.2).
type Cita struct {
	// Norma es el identificador de la norma citada.
	Norma string

	// Bloque es el id del bloque citado.
	Bloque string
}

// ExtraerCitas devuelve las citas de la respuesta con la forma fija del contrato
// skill-boe-legislacion §3, en el orden en que aparecen, o nil si no hay
// ninguna. Solo cuenta la pareja con la que terminan los corchetes: lo que no
// tiene los dos corchetes en la misma línea, la palabra bloque o el id justo
// antes del corchete de cierre no es una cita, y la redacción legible, dentro o
// fuera de los corchetes, no se lee (FR-008, FR-072).
func ExtraerCitas(respuesta string) []Cita {
	var citas []Cita

	for _, partes := range formaDeCita.FindAllStringSubmatch(respuesta, -1) {
		citas = append(citas, Cita{Norma: partes[1], Bloque: partes[2]})
	}

	return citas
}

// Lo que lleva la línea con la que una respuesta dice que no ha consultado nada
// porque el agente no tiene ni la herramienta ni el binario
// (contracts/skills.md §3 de H21; data-model §9 de H21).
const (
	// etiquetaSinConsulta es la etiqueta de su forma fija, la que va entre la
	// marca y los dos puntos.
	etiquetaSinConsulta = "SIN CONSULTA AL BOE"

	// direccionParaInstalar es la dirección que la línea lleva detrás de la causa.
	direccionParaInstalar = "https://kitlegal.es/instalar/"
)

// lineaSinConsulta casa con cada línea que empieza por la forma fija de
// etiquetaSinConsulta, con la tolerancia de las etiquetas de los avisos
// (patronDeEtiqueta): blancos y énfasis de Markdown delante de la marca y
// alrededor de las partes. El grupo es lo que sigue a los dos puntos hasta el
// final de la línea.
var lineaSinConsulta = regexp.MustCompile(`(?m)^` + separadorDeAviso + patronDeEtiqueta(etiquetaSinConsulta) +
	`([^\n]*)$`)

// ExtraerSinConsulta dice si la respuesta tiene una línea que empieza por la
// forma fija ⚠ SIN CONSULTA AL BOE: —la marca, la etiqueta y los dos puntos, con
// la tolerancia de las etiquetas de los avisos— y si alguna de esas líneas lleva
// detrás, en la misma línea, https://kitlegal.es/instalar/. La dirección en otra
// línea no cuenta, y sin la línea no hay dirección. No lee la causa ni ninguna
// otra redacción (H21 FR-035, FR-047).
func ExtraerSinConsulta(respuesta string) (conLinea, conDireccion bool) {
	for _, partes := range lineaSinConsulta.FindAllStringSubmatch(respuesta, -1) {
		conLinea = true

		if strings.Contains(partes[1], direccionParaInstalar) {
			conDireccion = true
		}
	}

	return conLinea, conDireccion
}
