package ids

import (
	"fmt"

	"github.com/jmorenobl/kitlegal/internal/core/schema"
)

// Lo que cada error nombra antes de la entrada: el identificador que se
// esperaba.
const (
	identificadorINE          = "el código INE"
	identificadorINEConDigito = "el código INE con dígito de control"
	identificadorDIR3         = "el código DIR3"
	identificadorECLI         = "el ECLI"
	identificadorROJ          = "el ROJ"
)

// errorDeEntrada es el rechazo de una entrada que no es el identificador que
// se esperaba. No se exporta —la superficie del paquete son sus tipos, sus
// analizadores y sus métodos, y nada más—: quien lo recibe lo reconoce por su
// clase, con errors.As a schema.ConClase, que es como lo reconoce el kernel.
type errorDeEntrada struct {
	// identificador es el que se esperaba, con su artículo.
	identificador string
	// entrada es lo recibido, tal cual.
	entrada string
	// motivo es lo que la entrada tiene de malo, dicho frente a la forma
	// esperada.
	motivo string
}

// Un rechazo de entrada declara su clase él mismo.
var _ schema.ConClase = (*errorDeEntrada)(nil)

// entradaInvalida construye el rechazo de entrada frente al identificador que
// se esperaba, con lo que tiene de malo.
func entradaInvalida(identificador, entrada, motivo string) error {
	return &errorDeEntrada{identificador: identificador, entrada: entrada, motivo: motivo}
}

// Error nombra el identificador que se esperaba, la entrada entrecomillada
// con %q —que hace visibles los espacios, los saltos de línea y los bytes que
// no son UTF-8— y lo que tiene de malo.
func (e *errorDeEntrada) Error() string {
	return fmt.Sprintf("%s %q no es válido: %s", e.identificador, e.entrada, e.motivo)
}

// Clase es siempre «argumentos»: una entrada mal formada es un fallo de quien
// la escribe, que el kernel traduce a código 2 (FR-033 y research.md V18 de
// H6; FR-006 de H23).
func (*errorDeEntrada) Clase() schema.Clase {
	return schema.ClaseArgumentos
}
