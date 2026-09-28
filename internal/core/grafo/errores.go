package grafo

import (
	"fmt"

	"github.com/jmorenobl/kitlegal/internal/core/schema"
)

// Rechazo es el motivo por el que un lote entero no entra en el grafo del
// mundo: una operación, o el propio lote, que incumple una regla (FR-024,
// FR-025; contracts/almacen-world-db.md §5). Un solo Rechazo basta para que no
// entre nada del lote.
//
// Se exporta porque lo construye también el adaptador, con lo que solo sabe
// dentro de la transacción —un extremo que no está, un tipo o un cuerpo que ya
// estaban guardados con otro valor—, y porque quien lo recibe llega con
// errors.As a la operación y al motivo.
type Rechazo struct {
	// Operacion es la operación rechazada, tal como la trae el lote; nula si
	// lo que se rechaza es el lote mismo: su procedencia o su vigencia.
	Operacion schema.Operacion
	// Motivo dice qué regla incumple.
	Motivo string
}

// Un rechazo declara su clase él mismo.
var _ schema.ConClase = (*Rechazo)(nil)

// Error nombra lo rechazado y el motivo, como «el nodo "ine:28074": <motivo>».
// Un nodo se nombra por su id, una arista por sus extremos y su relación y un
// texto por su huella, con %q, que hace visibles los espacios, los controles y
// los bytes que no son UTF-8. Nunca repite el cuerpo de un texto ni el id de
// una Persona: es donde el rechazo de FR-025 encuentra un documento de
// identidad, y el mensaje llega a la salida de error (constitución VII).
func (r *Rechazo) Error() string {
	return nombrar(r.Operacion) + ": " + r.Motivo
}

// Clase es «inesperado»: un lote lo construye un applet, no quien pregunta, y
// uno que no entra es un defecto de quien lo emite. En la entrega de un applet
// se convierte en la línea de aviso, y no cambia el código de salida
// (contracts/almacen-world-db.md §6).
func (*Rechazo) Clase() schema.Clase {
	return schema.ClaseInesperado
}

// nombrar describe la operación rechazada para el mensaje de un Rechazo. La
// interfaz está sellada, pero un puntero a un Nodo, una Arista o un Texto
// también la satisface: no es ninguno de los tres valores y se nombra por su
// tipo.
func nombrar(operacion schema.Operacion) string {
	switch op := operacion.(type) {
	case nil:
		return "el lote"
	case schema.Nodo:
		if op.Tipo == TipoPersona {
			return fmt.Sprintf("el nodo de tipo %q", op.Tipo)
		}

		return fmt.Sprintf("el nodo %q", op.ID)
	case schema.Arista:
		return fmt.Sprintf("la arista de %q a %q por %q", op.Origen, op.Destino, op.Relacion)
	case schema.Texto:
		return fmt.Sprintf("el texto %q", op.Huella)
	default:
		return fmt.Sprintf("la operación de tipo %T", op)
	}
}

// errorDeID es el id de `graph show` que no puede ser el de ningún nodo. No se
// exporta: quien lo recibe lo reconoce por su clase, con errors.As a
// schema.ConClase, que es como lo reconoce el kernel.
type errorDeID struct {
	// id es lo recibido, tal cual.
	id string
	// motivo es lo que el id tiene de malo.
	motivo string
}

// Un id que no puede serlo declara su clase él mismo.
var _ schema.ConClase = (*errorDeID)(nil)

// idNoValido construye el error de un id que no puede ser el de ningún nodo,
// con lo que tiene de malo.
func idNoValido(id, motivo string) error {
	return &errorDeID{id: id, motivo: motivo}
}

// Error nombra el id con %q —que hace visibles los espacios, los controles y
// los bytes que no son UTF-8— y lo que tiene de malo.
func (e *errorDeID) Error() string {
	return fmt.Sprintf("el id %q no puede ser el de ningún nodo: %s", e.id, e.motivo)
}

// Clase es «argumentos»: un id así es un fallo de quien lo escribe, que el
// kernel traduce a código 2 (FR-052).
func (*errorDeID) Clase() schema.Clase {
	return schema.ClaseArgumentos
}
