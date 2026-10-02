package mcp

// Instrucciones es el texto fijo que el servidor da como `instructions`: lo que
// un agente lee una vez por conexión, y lo único que tiene donde hay servidor y
// no hay skill. Son cinco frases, en este orden, una por cada contenido que
// pide FR-006 de H21: qué son las herramientas, que ningún contenido legal se
// afirma si no viene del texto devuelto, que cada afirmación lleva norma y
// bloque, que los avisos del sobre se trasladan y que el protocolo completo son
// las skills de kitlegal.
//
// Mide 485 bytes y sus cinco frases caben en los primeros 512, que es la parte
// que la documentación de Codex pide que se baste a sí misma. El texto lo fija
// contracts/servidor-mcp.md §5 de H21, carácter a carácter, y TestInstrucciones
// falla si una frase falta, cambia de sitio o queda detrás del byte 512
// (FR-077; research.md D10).
const Instrucciones = "Herramientas de kitlegal: consultan fuentes legales públicas españolas y devuelven un " +
	"sobre con el texto vigente de una norma consolidada del BOE o con el territorio de un municipio. " +
	"No afirmes ningún contenido legal que no venga del texto devuelto en esta conversación. " +
	"Cada afirmación lleva su cita, con la norma y el bloque: art. 21 de la Ley 39/2015 " +
	"[BOE-A-2015-10565, bloque a21]. " +
	"Traslada los avisos del sobre (data.avisos). " +
	"El protocolo completo son las skills de kitlegal."
