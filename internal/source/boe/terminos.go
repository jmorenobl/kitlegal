package boe

import (
	"time"

	"github.com/jmorenobl/kitlegal/internal/core"
)

// Lo que la fuente declara de cómo se la consulta: el ritmo al que se le pide y
// los términos de uso bajo los que se hace. Los dos salen de la fila
// boe.legislacion-consolidada de docs/SOURCES.md, que revisa y fija una persona,
// y TestFuenteCoincideConSources falla si este fichero y esa fila divergen
// (FR-121, FR-122, FR-123; research.md D15). Viven solos en este fichero para
// que quien revisa la fila los alinee con ella sin tocar ningún otro código.

// IntervaloEntrePeticiones es la separación mínima entre dos peticiones al sitio
// del BOE: la celda «Ritmo» de la fila de la fuente, y no el valor por omisión
// de internal/httpx, aunque hoy coincidan. Con él graba TestGrabarFixtures
// (FR-122).
const IntervaloEntrePeticiones = time.Second

// terminosDeUso son los términos de uso de la fuente: la dirección de la celda
// «Términos de uso» de la fila y el día de su celda «Revisado», a las 00:00 UTC.
//
// Revisados cero es la fila con «Revisado» pendiente: la dirección es una
// propuesta que ninguna persona ha comprobado todavía. Quien revisa los
// términos y el robots.txt fija los dos valores a la vez que la fila, antes de
// grabar (contrato esquemas-fixtures-y-controles §3.2 y §8).
var terminosDeUso = core.Terminos{
	URL:       "https://www.boe.es/informacion/aviso_legal/index.php",
	Revisados: time.Time{},
}
