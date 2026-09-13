package core

import (
	"context"
	"time"
)

// Cache es el puerto de caché del dominio: guardar un contenido bajo una clave
// con una vigencia y recuperarlo mientras siga vigente. El dominio no sabe qué
// hay detrás ni dónde se guarda; quien lo implementa es un adaptador.
//
// El puerto no tiene Close, y no es un olvido: cierra quien lo construyó, de
// modo que una fuente a la que se le entrega una Cache no pueda cerrar la que
// no es suya (FR-004, research.md D1).
type Cache interface {
	// Get devuelve el contenido guardado bajo la clave mientras siga vigente,
	// con el idioma «coma ok» de Go. Presente y vigente es (contenido, true,
	// nil). La ausencia es (nil, false, nil): el resultado normal que lleva a
	// quien llama a pedirlo a la fuente, y también la forma en que se lee una
	// entrada expirada, que a efectos de lectura no está. El fallo es (nil,
	// false, err), y ese error declara su clase —implementa schema.ConClase—,
	// de manera que quien llama nunca compara errores para saber si había
	// entrada (FR-001, FR-013).
	//
	// Una implementación construida para no ir a la fuente —la de solo
	// lectura que impone --offline— no puede devolver la ausencia como
	// ausencia, porque no hay a dónde ir a buscar lo que falta: la informa
	// como un fallo de la clase «fuente no disponible», que quien llama
	// propaga tal cual y el kernel traduce al código 4. Fuera de ese modo la
	// ausencia nunca es un error (FR-013, FR-016).
	Get(ctx context.Context, clave string) (contenido []byte, presente bool, err error)

	// Put guarda el contenido bajo la clave con la vigencia dada, sustituyendo
	// por completo lo que hubiera. La clave y el contenido son opacos: el
	// puerto no los interpreta ni les supone forma alguna. Devuelve nil si y
	// solo si la entrada quedó guardada (FR-006, FR-007, FR-012).
	Put(ctx context.Context, clave string, contenido []byte, vigencia time.Duration) error
}
