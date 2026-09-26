package instalacion

import (
	"fmt"
	"strings"

	"github.com/jmorenobl/kitlegal/internal/core/schema"
)

// errorDeInvocacion es el rechazo de una invocación antes de tocar el disco
// (contracts/applet-skills.md §2): uno de argumentos, que el kernel traduce a
// código 2, o -g sin HOME, que traduce a 1 (data-model §9). No se exporta:
// quien lo recibe lo reconoce por su clase, con errors.As a schema.ConClase,
// que es como lo reconoce el kernel.
type errorDeInvocacion struct {
	// clase es schema.ClaseArgumentos o, solo para -g sin HOME,
	// schema.ClaseInesperado.
	clase schema.Clase
	// mensaje contiene la frase de su fila de la tabla y dice por qué.
	mensaje string
}

// Un rechazo de la invocación declara su clase él mismo.
var _ schema.ConClase = (*errorDeInvocacion)(nil)

// Error devuelve el mensaje, que llega literal al sobre de fallo y a la salida
// de error.
func (e *errorDeInvocacion) Error() string {
	return e.mensaje
}

// Clase es la del rechazo.
func (e *errorDeInvocacion) Clase() schema.Clase {
	return e.clase
}

// argumentosInvalidos es un rechazo de clase «argumentos», el de las filas 1
// a 4.
func argumentosInvalidos(mensaje string) error {
	return &errorDeInvocacion{clase: schema.ClaseArgumentos, mensaje: mensaje}
}

// globalConDir es la fila 1: -g y --dir a la vez, que serían dos ámbitos
// (FR-013).
func globalConDir() error {
	return argumentosInvalidos("-g y --dir se excluyen: cada invocación actúa en un único ámbito, " +
		"el global o el de --dir")
}

// hostConDir es la fila 2: --host junto a --dir, cuyo ámbito no tiene hosts
// (FR-013).
func hostConDir() error {
	return argumentosInvalidos("--host no se combina con --dir: el ámbito de --dir es solo ese directorio, " +
		"sin ningún host")
}

// hostNoAdmitido es la fila 3: --host con un valor distinto de claude
// (FR-020), que nombra con %q.
func hostNoAdmitido(host string) error {
	return argumentosInvalidos(fmt.Sprintf("el host %q no se admite: el único host admitido es %s", host, hostClaude))
}

// skillDesconocida es la fila 4: nombre, con %q, no es el de ninguna skill
// empotrada; disponibles son las que sí, en orden de nombre (FR-010).
func skillDesconocida(nombre string, disponibles []string) error {
	return argumentosInvalidos(fmt.Sprintf("%q no es ninguna skill de este binario; skills disponibles: %s",
		nombre, strings.Join(disponibles, ", ")))
}

// homeSinDefinir es la fila 5: -g con HOME sin definir o vacío (FR-012). No
// es un error de argumentos —la orden está bien escrita y es el entorno el que
// no la permite—, así que declara la clase «inesperado» (data-model §9).
func homeSinDefinir() error {
	return &errorDeInvocacion{
		clase: schema.ClaseInesperado,
		mensaje: "HOME no está definido o está vacío: el ámbito global de -g está en " +
			"$HOME/" + directorioNeutro,
	}
}

// ManifiestoIlegible es un kitlegal.json que existe y no se puede usar: no es
// un fichero regular, no se puede leer o no respeta la forma de
// contracts/manifiesto.md (FR-035). Quien no puede leer el manifiesto no sabe
// qué es suyo, así que install lo nombra como conflicto y no toca nada, list y
// doctor salen nombrándolo y el aviso no tiene ningún efecto.
//
// No declara una clase a propósito: nunca llega así al kernel. El conflicto
// de install y el ámbito ilegible de list y doctor son los errores que la
// declaran (data-model §9). Se exporta para reconocerlo con errors.As.
type ManifiestoIlegible struct {
	// causa dice qué lo hace ilegible; puede envolver el error de
	// encoding/json/v2 que rechazó el documento.
	causa error
}

// Error dice que el manifiesto es ilegible y por qué, si se sabe.
func (e *ManifiestoIlegible) Error() string {
	if e.causa == nil {
		return "manifiesto ilegible"
	}

	return "manifiesto ilegible: " + e.causa.Error()
}

// Unwrap devuelve lo que lo hace ilegible.
func (e *ManifiestoIlegible) Unwrap() error {
	return e.causa
}
