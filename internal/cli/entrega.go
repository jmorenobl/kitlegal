package cli

import (
	"context"
	"strings"
	"time"

	"github.com/jmorenobl/kitlegal/internal/core"
	"github.com/jmorenobl/kitlegal/internal/core/schema"
)

// prefijoDeEntregaFallida encabeza la línea con la que el kernel avisa de que lo
// observado no llegó al grafo del mundo. La causa que la sigue la da el
// almacén, y todo mensaje de internal/graph nombra world.db, así que la línea
// nombra world.db y la causa (contracts/resultado-y-entrega.md §4, research.md
// D7).
const prefijoDeEntregaFallida = "kitlegal: lo observado no ha llegado al grafo del mundo: "

// LoteDe es el lote que el kernel entrega al grafo por una invocación: las
// operaciones y la vigencia de lo observado con la procedencia del sobre que se
// presentó. La fecha es el mismo texto que el sobre escribe en fecha_consulta
// —RFC 3339 con la fracción de segundo sin ceros a la derecha—, también cuando
// la puso el reloj del montador, porque se toma del sobre ya montado y no de
// otra lectura del reloj (FR-021; research.md D4, D5, V2, V18).
//
// Es el único sitio que construye un lote con procedencia: ninguna operación
// llega al grafo con una fuente distinta de la del sobre que la sostiene.
func LoteDe(sobre schema.Sobre, observado schema.Observado) core.Lote {
	return core.Lote{
		Fuente:        sobre.Fuente,
		URL:           sobre.URL,
		FechaConsulta: sobre.FechaConsulta.Format(time.RFC3339Nano),
		Vigencia:      observado.Vigencia,
		Operaciones:   observado.Operaciones,
	}
}

// entregar lleva al grafo del mundo lo que observó una invocación que ya se
// presentó con éxito. Sin almacén o sin operaciones no hace nada, antes de
// nada: ni se construye el lote ni el almacén resuelve ninguna ruta (FR-030;
// research.md D6).
//
// Una entrega que falla no cambia el desenlace: el sobre ya dijo `ok: true` y
// el código 0 es el único coherente con él, así que lo único que deja es una
// línea en la salida de error (FR-033). El error al escribir esa línea no se
// propaga, y es una decisión y no un silencio (research.md D7): con el mismo
// razonamiento que el aviso de versión de H19, la línea nunca cambia el código
// ni la salida estándar, y contar ese error iría al mismo descriptor que acaba
// de fallar.
func (m Montador) entregar(ctx context.Context, p Presentador, sobre schema.Sobre, observado schema.Observado) {
	if m.Grafo == nil || len(observado.Operaciones) == 0 {
		return
	}

	if err := m.Grafo.Apply(ctx, LoteDe(sobre, observado)); err != nil {
		_ = p.Aviso(lineaDeEntregaFallida(err))
	}
}

// lineaDeEntregaFallida es el aviso de una entrega que falló: una sola línea,
// con cada salto de línea de la causa sustituido por un espacio
// (contracts/resultado-y-entrega.md §4).
func lineaDeEntregaFallida(causa error) string {
	return prefijoDeEntregaFallida + strings.ReplaceAll(causa.Error(), "\n", " ")
}
