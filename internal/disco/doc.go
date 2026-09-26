// Package disco es el adaptador de los puertos de internal/core/instalacion
// sobre el sistema de ficheros real (docs/ADR/0019; research.md D2 de H19):
//
//   - Lector implementa Disco, la lectura sin seguir nunca el último enlace de
//     una ruta: Examinar y Nombres miran con Lstat y Readlink, y solo para
//     decir si un enlace resuelve hacen un Stat sobre él, sin leer nada de lo
//     que encuentran; Huella y Leer abren solo un fichero regular, y
//     comprueban que lo abierto es lo examinado (research.md D6).
//   - Escritor implementa Escritor, las operaciones sueltas con las que el
//     dominio aplica un plan: crear un directorio, escribir un fichero de forma
//     atómica —un temporal del mismo directorio, volcado a disco y renombrado
//     encima—, retirar una entrada y enlazar con el Enlazador que recibe
//     (research.md D8).
//   - Enlazador implementa Enlazador, el creador de enlaces del sistema, que
//     dice si en un directorio se pueden crear enlaces con una sonda en ese
//     mismo directorio, que retira después (research.md D9).
//
// El adaptador no decide nada: qué se examina, se lee, se escribe o se
// retira, y en qué orden, lo decide el dominio, que ya ha examinado cada
// directorio por encima de lo que pide y lo que hay por encima de la raíz del
// ámbito lo usa tal cual (FR-027). Por eso Lstat no sigue la última componente
// de una ruta y sí las de encima.
//
// Los errores del Lector y del Enlazador nombran la operación y la ruta, «leer
// <ruta>: <error del sistema>», porque el dominio los devuelve tal cual; los
// del Escritor son solo el error del sistema, porque el dominio los envuelve
// con la operación y la ruta del plan (contracts/applet-skills.md §5). El
// error del sistema es la causa, sin la operación ni la ruta con que la
// envuelve el paquete os, para no nombrarlas dos veces.
//
// Límite conocido (research.md D6): entre el Lstat y el Open de un fichero
// regular, otro proceso con permiso de escritura en el directorio podría
// cambiarlo por una tubería con nombre, y abrirla bloquearía. Cerrarlo pediría
// O_NONBLOCK, que solo existe en Unix.
package disco
