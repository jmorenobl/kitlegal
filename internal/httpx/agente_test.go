package httpx

import (
	"regexp"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestAgenteDeUsuario fija la forma exacta de la identificación y la ata a la
// versión que el binario lleva inyectada (FR-006, FR-007, research.md D5).
//
// La expresión regular se construye **sobre la variable** version y no sobre el
// literal «dev»: así el caso vale igual para la construcción sin datos
// inyectados —la de go test y go run— y para cualquier valor que -ldflags
// ponga, que es lo que hace demostrable el camino completo. Un caso escrito
// contra «dev» fallaría con la versión inyectada y no comprobaría nada.
//
// El sufijo pasa por regexp.QuoteMeta en lugar de escaparse a mano por dos
// razones: el fuente dice entonces la identificación tal cual se envía, y la
// comprobación mecánica «solo direcciones locales» de quickstart.md extrae de
// él exactamente la dirección de la identificación, que descarta por su tercer
// filtro; escrita a mano dejaría aquí un token con barras invertidas que ese
// filtro no reconoce y la comprobación quedaría en rojo sin motivo.
//
// El t.Log no es traza de depuración: es lo que hace observable la inyección.
// `go test -v -ldflags '-X …/internal/httpx.version=9.9.9-prueba'` muestra la
// identificación resultante con el test en PASS (quickstart.md escenario 2).
func TestAgenteDeUsuario(t *testing.T) {
	t.Parallel()

	// Una versión vacía produciría «kitlegal/ (+…)», que es identificación
	// inválida: FR-007 prohíbe que quede sin versión, venga o no inyectada.
	require.NotEmpty(t, version,
		"la versión de la identificación nunca queda vacía, ni siquiera sin -ldflags (FR-007)")

	agente := AgenteDeUsuario()

	t.Log(agente)

	forma := "^kitlegal/" + regexp.QuoteMeta(version) +
		regexp.QuoteMeta(" (+https://kitlegal.es/bot)") + "$"

	assert.Regexp(t, forma, agente,
		"la identificación tiene la forma exacta kitlegal/<versión> (+<dirección del bot>) (FR-006)")
}
