package schema

import "time"

// Contexto lleva al applet las opciones globales ya interpretadas, para que
// ningún applet tenga que leer una bandera (FR-018). Es de solo lectura para el
// applet: se construye una vez por invocación y no se reutiliza en otra.
//
// El registrador de eventos no viaja aquí, y no es un descuido: es un puerto
// con entrada y salida detrás, y el dominio no importa log/slog. El applet lo
// recibe ya montado y con el nivel ya resuelto como parámetro explícito de su
// método de ejecución (data-model.md §4, research.md D2). Tampoco viaja
// --describe, que excluye la ejecución y por tanto el applet nunca llega a ver
// (FR-049).
type Contexto struct {
	// JSON es la forma de presentación que pidió --json. El applet no la usa;
	// la usa el presentador.
	JSON bool
	// Timeout es el plazo de --timeout, ya aplicado como plazo del
	// context.Context que acompaña a la llamada.
	Timeout time.Duration
	// Offline declara que la operación no puede acceder a la red (--offline).
	Offline bool
	// DryRun pide describir la operación en lugar de realizarla (--dry-run).
	DryRun bool
	// SinGrafo pide no alimentar el grafo con esta invocación (--no-graph).
	SinGrafo bool
	// Asunto nombra el asunto sobre el que se trabaja (--asunto).
	Asunto string
}
