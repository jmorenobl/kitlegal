// Package grafo es el dominio del grafo del mundo: las reglas de lo que el
// binario recuerda de las fuentes públicas que consulta, con su procedencia
// (docs/ADR/0014, research.md D1). Aquí viven el vocabulario —los tipos de
// nodo, las relaciones, las claves de datos y las clases de hallazgo—, la forma
// canónica de los datos identificativos (RFC 8785), la validación de un lote y
// el rechazo de una Persona con un documento de identidad, la fusión de
// observaciones, las reglas de `graph check` con sus explicaciones y las formas
// de salida de los verbos de `graph`.
//
// Lo que garantiza:
//
//   - Los datos identificativos de un nodo tienen una sola forma canónica, la
//     de RFC 8785, que no depende de cómo se construyen ni del orden en el que
//     se recorren; unos datos sin forma JSON dan un error y no se corrigen
//     (FR-023, research.md D17).
//   - Un lote que incumple una regla se rechaza entero con un Rechazo que
//     nombra la operación y el motivo, de clase «inesperado», y que nunca
//     repite el cuerpo de un texto ni nombra por su id un nodo Persona, que
//     nombra por su tipo; una arista sí se nombra por los ids de sus extremos
//     (FR-024, FR-025).
//   - El id de `graph show` vacío, formado solo por caracteres de espacio en
//     blanco o con algún carácter de control es un error de clase
//     «argumentos», que nombra el id; cualquier otro se busca tal cual, sin
//     recortar (FR-052, research.md D35).
//
// Es dominio puro: no hace entrada ni salida y no importa más paquetes
// internos que internal/core —por core.Lote, lo que valida— e
// internal/core/schema (regla R1 de docs/ROADMAP.md §2). Los
// applets usan su vocabulario para declarar lo que observan, y el adaptador
// internal/graph, sus reglas para guardar y leer world.db.
package grafo
