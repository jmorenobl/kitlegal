// Package mcp es el adaptador del protocolo MCP: el único árbol del módulo que
// importa github.com/modelcontextprotocol/go-sdk (regla R7; H21 FR-026 y
// FR-079, research.md D2). Sirve por la entrada y la salida que recibe las
// herramientas que recibe, y nada más: ni abre un puerto, ni atiende ninguna
// conexión de red, ni anuncia `prompts` ni `resources` (FR-001, FR-008).
//
// Recibe las herramientas ya hechas. No sabe de applets ni de sobres: de cada
// Herramienta anuncia el nombre, la descripción y los dos esquemas con los
// bytes que le dan, y de cada llamada devuelve los bytes que la herramienta le
// entrega, marcados como error solo si ella lo dice. Qué verbo hay detrás, qué
// argumentos valen y qué es un fallo lo decide quien las compone, que es
// internal/app, donde viven el registro y el kernel.
//
// Exporta cinco cosas: Instrucciones, el texto fijo que el servidor da como
// `instructions`; Herramienta y Resultado, lo que se sirve y lo que devuelve
// una llamada; y Servicio y Servir, lo que hace falta para atender una
// conexión y la función que la atiende hasta que la entrada se cierra
// (contracts/servidor-mcp.md §5 a §7 de H21).
//
// Todo lo que el SDK decide por su cuenta se le deja: las versiones del
// protocolo que negocia, sin estrechar la lista (FR-007), el orden de las
// herramientas en `tools/list` y la respuesta a una herramienta que no existe,
// que es un error del protocolo y no llega a ninguna Herramienta.
package mcp
