package core

import (
	"context"
	"time"

	"github.com/jmorenobl/kitlegal/internal/core/schema"
)

// Source es el puerto de una fuente pública: una consulta entra y sale un
// schema.Resultado con su procedencia —fuente, dirección y fecha de la
// consulta—. El dominio no sabe cómo se pide, se lee ni se guarda lo que
// responde; quien lo implementa es un adaptador de internal/source (FR-121,
// docs/ADR/0015).
//
// El puerto no lleva registrador de eventos, y no es un olvido: el dominio no
// importa log/slog, así que el adaptador lo recibe al construirse, uno por
// invocación.
type Source interface {
	// Name es el nombre de la fuente tal como va en `fuente` del sobre. Nunca
	// empieza por «kitlegal.», el espacio reservado de lo que se obtiene sin
	// consultar ninguna fuente pública (ADR 0006).
	Name() string

	// Fetch resuelve la consulta. ec lleva --offline y --dry-run, que la fuente
	// honra. En éxito, la procedencia declara la fecha de la consulta que
	// sostiene el contenido, también cuando sale de la caché (FR-096). En fallo
	// devuelve además un Resultado cuya procedencia nombra la petición que
	// falló —o el valor cero si no llegó a construirse ninguna—, y su error
	// declara su clase (schema.ConClase).
	Fetch(ctx context.Context, ec schema.Contexto, consulta Consulta) (schema.Resultado, error)

	// TTL es la vigencia con la que se guarda lo que responde esa consulta. Es
	// por consulta y no por fuente porque cada verbo envejece a su ritmo
	// (FR-091).
	TTL(consulta Consulta) time.Duration

	// Terms son los términos de uso de la fuente y el día en que una persona
	// los revisó (FR-121).
	Terms() Terminos
}

// Consulta es lo que se pide a una fuente. Cada fuente declara sus propios
// tipos de consulta, uno por verbo; una consulta de un tipo que la fuente no
// declara es un defecto de quien la compone, nunca de quien invoca.
type Consulta interface {
	// Verbo es el nombre del verbo que la consulta resuelve.
	Verbo() string
}

// Terminos son los términos de uso de una fuente pública: dónde se leen y
// cuándo los revisó una persona. Cada adaptador los ata con un test a la fila
// de su fuente en docs/SOURCES.md (FR-121, FR-122).
type Terminos struct {
	// URL es la dirección de los términos de uso.
	URL string
	// Revisados es el día de la revisión, a las 00:00 UTC.
	Revisados time.Time
}
