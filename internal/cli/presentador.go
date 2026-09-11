package cli

import (
	"io"

	"github.com/jmorenobl/kitlegal/internal/core/schema"
)

// Presentador es todo lo que el kernel necesita para escribir, declarado por el
// consumidor y no por quien lo implementa. La implementación vive en
// internal/render y la inyecta la raíz de composición, de modo que internal/cli
// no importa internal/render y la dirección de dependencia del roadmap
// —cmd → app → cli → core/schema— se mantiene sin ciclos (research.md D2, D15).
//
// Los cinco métodos son los que el kernel usa y ninguno más: el sobre, las dos
// clases de texto para personas y los dos escritores crudos. Añadir una forma de
// presentación nueva es implementar esta misma interfaz de otro modo, sin tocar
// applet alguno (FR-044).
//
// Toda escritura devuelve su error y ninguna lo descarta. Un descriptor roto es
// un fallo como cualquier otro: sube hasta el único punto que traduce errores a
// código de salida, en lugar de convertirse en un pánico o en silencio
// (FR-033, FR-040, research.md D15).
type Presentador interface {
	// Presentar escribe el sobre en la salida estándar en la forma que pide
	// enJSON: el documento JSON cuando es verdadero y la tabla mínima cuando no.
	// Es el único camino por el que un sobre llega a la salida estándar, tanto en
	// éxito como en fallo (FR-041, FR-042, FR-045).
	Presentar(sobre schema.Sobre, enJSON bool) error
	// Texto escribe en la salida estándar el texto dirigido a una persona que no
	// es un resultado: las tres líneas de version y la ayuda derivada del
	// registro. No lo altera --json, que solo elige la forma del sobre (FR-026,
	// FR-042).
	Texto(texto string) error
	// Aviso escribe en la salida de error el mensaje dirigido a una persona: la
	// causa de un fallo, la lista de applets disponibles, la descripción de
	// --dry-run y el aviso de un KITLEGAL_LOG inválido. No pasa por el registro
	// de eventos, y por eso ningún nivel puede ocultarlo (FR-022, research.md
	// D10, D14).
	Aviso(texto string) error
	// Salida es el escritor crudo de la salida estándar. Existe solo para
	// entregárselo a quien escribe por su cuenta y no puede recibir un
	// Presentador —el analizador de la línea de órdenes, que emite ahí su ayuda—;
	// el kernel no escribe por él (research.md D11).
	Salida() io.Writer
	// Error es el escritor crudo de la salida de error. Además de lo anterior, es
	// el destino del registro de eventos, que se monta sobre él y nunca sobre la
	// salida estándar (FR-036).
	Error() io.Writer
}
