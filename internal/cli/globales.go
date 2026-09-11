package cli

import (
	"time"

	"github.com/jmorenobl/kitlegal/internal/core/schema"
)

// timeoutPorOmision es el plazo de toda la operación cuando nadie lo fija: lo
// bastante para una consulta lenta a una fuente pública y lo bastante poco para
// que un agente no se quede esperando (research.md D9).
//
// La etiqueta de la bandera lo escribe otra vez, porque Kong necesita el valor
// por omisión como texto; que las dos escrituras digan lo mismo lo comprueba
// TestGlobales.
const timeoutPorOmision = 30 * time.Second

// Globales son las ocho banderas que el kernel ofrece a todos los applets con
// idéntica sintaxis y semántica. Se declaran **aquí y una sola vez**: toda
// gramática de invocación las embebe, de modo que un applet nuevo que solo
// declara su nombre, sus verbos y el contenido de su data las hereda todas sin
// escribir ninguna (FR-018, SC-010, contracts/banderas-y-exit-codes.md §1).
//
// Que ningún applet pueda redefinir una de ellas no es una promesa de este
// comentario: son la misma declaración embebida, y si un verbo declarara una
// bandera con uno de estos ocho nombres la gramática no se construiría.
//
// Tres de las ocho —--offline, --no-graph y --asunto— no tienen todavía objeto:
// H1 fija su sintaxis y las hace llegar al applet, y no les inventa una
// semántica que ningún hito ha definido (FR-021, FR-023, FR-024).
type Globales struct {
	// JSON elige la forma legible por máquina: el sobre en JSON en lugar de la
	// tabla mínima para personas (FR-019).
	JSON bool `help:"Emite el sobre en JSON en lugar de la tabla mínima."`
	// Timeout es el plazo de toda la operación, no el de una petición suelta.
	// Agotarlo es código 4 (FR-020).
	Timeout time.Duration `help:"Plazo total de la operación." default:"30s"`
	// Offline declara que la operación no puede acceder a la red. Se acepta y se
	// propaga; responder solo desde caché es alcance de H3 (FR-021).
	Offline bool `help:"Declara que la operación no puede acceder a la red."`
	// DryRun pide describir la operación en lugar de realizarla. No corta el
	// análisis: viaja en el contexto de ejecución (FR-022, research.md D10).
	DryRun bool `help:"Describe la operación en lugar de realizarla."`
	// Describe pide el esquema JSON de entrada y salida del verbo y excluye la
	// ejecución. No viaja al applet, que nunca llega a verlo (FR-046, FR-049).
	Describe bool `help:"Emite el esquema JSON de entrada y salida, sin ejecutar nada."`
	// SinGrafo declara que la ejecución no altera el grafo. Se acepta y se
	// propaga; el grafo llega en H12 (FR-023).
	SinGrafo bool `name:"no-graph" help:"Declara que la ejecución no altera el grafo."`
	// Asunto declara sobre qué asunto se trabaja. Se acepta y se propaga; abrir
	// o crear un asunto es alcance de H14 (FR-024).
	Asunto string `help:"Asunto sobre el que se trabaja."`
	// Verbose sube el detalle del registro de eventos, que va siempre a la
	// salida de error. No altera la salida estándar en absoluto, y por eso
	// tampoco viaja al applet (FR-025, FR-036).
	Verbose bool `help:"Sube el detalle del registro de eventos en la salida de error."`
}

// Contexto convierte las banderas en el contexto de ejecución que recibe el
// applet, para que ningún applet tenga que leer una bandera (FR-018).
//
// Dos de las ocho se quedan en el kernel y no es un olvido: --describe excluye
// la ejecución, así que el applet nunca llega a ver esa invocación (FR-049), y
// --verbose solo fija el nivel del registro de eventos, que viaja al applet como
// registrador ya montado y no como decisión que interpretar (research.md D14,
// data-model.md §8).
func (g Globales) Contexto() schema.Contexto {
	return schema.Contexto{
		JSON:     g.JSON,
		Timeout:  g.Timeout,
		Offline:  g.Offline,
		DryRun:   g.DryRun,
		SinGrafo: g.SinGrafo,
		Asunto:   g.Asunto,
	}
}
