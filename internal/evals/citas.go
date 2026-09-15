package evals

import "regexp"

// formaDeCita es la expresión con la que se extrae la parte mecánica de una
// cita, [<identificador>, bloque <id de bloque>], con el identificador y el id
// tal como los da la fuente (contrato skill-boe-legislacion §3). El corchete de
// cierre delimita el id del bloque, que puede terminar en punto, como a85bis.
var formaDeCita = regexp.MustCompile(`\[(BOE-A-[0-9]{4}-[0-9]{1,9}), bloque ([A-Za-z0-9][A-Za-z0-9.-]{0,63})\]`)

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
// ninguna. Solo cuenta la parte entre corchetes: lo que no tiene los dos
// corchetes o la palabra bloque no es una cita, y la redacción legible de
// alrededor no se lee (FR-008, FR-072).
func ExtraerCitas(respuesta string) []Cita {
	var citas []Cita

	for _, partes := range formaDeCita.FindAllStringSubmatch(respuesta, -1) {
		citas = append(citas, Cita{Norma: partes[1], Bloque: partes[2]})
	}

	return citas
}
