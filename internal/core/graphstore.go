package core

import (
	"context"
	"time"

	"github.com/jmorenobl/kitlegal/internal/core/schema"
)

// GraphStore es el puerto del grafo del mundo: recibe lo que una invocación
// observó, con la procedencia del sobre que la presentó, y lo guarda. El
// dominio no sabe dónde ni cómo; quien lo implementa es un adaptador
// (docs/ADR/0014, FR-020).
//
// El puerto no tiene Close, y no es un olvido: cada entrega abre, aplica y
// cierra, así que entre dos entregas no queda nada abierto que cerrar.
type GraphStore interface {
	// Apply guarda el lote entero o nada: es transaccional —si devuelve un
	// error, el grafo queda como estaba— e idempotente —aplicar dos veces el
	// mismo lote deja el grafo como lo dejó la primera— (FR-022).
	Apply(ctx context.Context, lote Lote) error
}

// Lote es lo que el kernel entrega al grafo por una invocación: las
// operaciones que el applet declaró en su Resultado y, para todas ellas, la
// procedencia del sobre presentado. Lo construye el kernel y nadie más, de
// modo que ninguna operación llega con una fuente distinta de la del sobre que
// la sostiene (FR-021).
type Lote struct {
	// Fuente es la `fuente` del sobre presentado.
	Fuente string
	// URL es la `url` del sobre presentado.
	URL string
	// FechaConsulta es la `fecha_consulta` del sobre presentado, escrita como
	// la escribe el sobre (RFC 3339 con fracción de segundo si la tiene): es un
	// texto y no un instante para que el grafo guarde lo mismo que el sobre
	// cita, y compara instantes interpretándolo (research.md D5).
	FechaConsulta string
	// Vigencia es la de Observado. Cero significa que la fuente no la declara
	// (FR-065).
	Vigencia time.Duration
	// Operaciones son las de Observado, en su orden.
	Operaciones []schema.Operacion
}
