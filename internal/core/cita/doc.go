// Package cita es el dominio de la cita de una sentencia (H23): la referencia
// con la que se nombra, la consulta que una persona hace en el buscador del
// CENDOJ para encontrarla, la ficha con la que el CENDOJ encabeza el documento
// que esa persona trae y el cotejo que dice si el documento es el pedido. Las
// citas de normas entran con su hito, no antes.
//
// Nada de lo que hace consulta una fuente (docs/ADR/0036): no pide nada al
// buscador, no descarga ningún documento y no comprueba que una dirección
// responde. De una referencia solo dice dónde buscarla y qué escribir, y de un
// documento, lo que lleva su ficha: ni que la sentencia existe ni que el
// CENDOJ la tiene.
//
// Lo que garantiza:
//
//   - Una referencia se da de una sola forma de tres —un ECLI español, un ROJ
//     o un número de resolución con su fecha— y se construye con
//     NuevaReferencia desde los cuatro argumentos, cada uno dado o no dado. El
//     que llega escrito con valor vacío está dado, y se comprueba con su forma
//     como cualquier otro (FR-005, FR-006; research.md D3).
//   - Preparar da la dirección del buscador y cada casilla con su valor, y
//     PrepararTexto, la dirección que abre el buscador con una búsqueda ya
//     hecha. El equivalente de una referencia va solo si se deduce sin
//     consultar nada, y la cobertura, solo con el ECLI del Tribunal
//     Constitucional, que el CENDOJ no tiene: lo que no aplica no está
//     (FR-010 a FR-014; research.md D4).
//   - LeerFicha lee la primera ficha del texto, desde su línea «Roj:», con sus
//     ocho datos, y Cotejar dice si su ROJ y su ECLI se corresponden y, con una
//     referencia, si el documento es el pedido. Que no lo sea es un hallazgo
//     del cotejo, no un fallo (FR-021 a FR-025; docs/ADR/0023).
//   - Todo rechazo es un error de clase «argumentos», que el kernel traduce a
//     código 2, con un mensaje en español que nombra lo recibido y lo que le
//     falta; nunca un panic ni un valor por omisión (FR-006, FR-015, FR-026).
//   - Las fechas van AAAA-MM-DD en todo lo que el paquete devuelve, también la
//     de la ficha, que en el documento se escribe dd/mm/aaaa; en el valor de
//     una casilla van dd/mm/aaaa, que es como las escribe quien la rellena
//     (research.md D8).
//
// Es dominio puro: no hace entrada ni salida, no guarda nada y no importa más
// paquetes internos que internal/core/ids e internal/core/schema (regla R1 de
// docs/ROADMAP.md §2; FR-002).
package cita
