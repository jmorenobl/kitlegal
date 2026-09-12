// Package cache es el adaptador que implementa el puerto core.Cache sobre
// SQLite sin cgo, y el único lugar del módulo donde se abre, se migra y se
// consulta la base de datos de la caché (regla R3 de docs/ROADMAP.md §2). El
// dominio declara qué es una caché; este paquete es el único que sabe que
// detrás hay un fichero.
//
// Lo que garantiza:
//
//   - Lo guardado se devuelve íntegro y byte a byte igual que se escribió, y un
//     contenido de cero bytes se guarda y se distingue de la ausencia (FR-012).
//   - Una entrada se sirve solo mientras el reloj del cliente va por delante
//     del instante de expiración: en el instante exacto, y después, es ausencia
//     (FR-008).
//   - Fuera del modo de solo lectura, la ausencia —también la de una entrada
//     expirada— no es un fallo, sino el resultado normal que lleva a quien
//     llama a pedir a la fuente (FR-013, FR-014).
//   - Toda operación exige el contexto de quien llama y termina cuando ese
//     contexto termina, también mientras espera a que otra invocación suelte
//     el bloqueo: esa espera se hace por tramos cortos y entre tramo y tramo se
//     mira el contexto. El paquete no crea ninguno por su cuenta (FR-003).
//   - Todo fallo declara su clase por schema.ConClase y ninguna ruta termina en
//     panic (FR-033, FR-034).
//
// Lo que deliberadamente no hace:
//
//   - No lee ninguna bandera de la invocación ni inspecciona el contexto de
//     ejecución de cada llamada: el modo de solo lectura lo pone quien
//     construye el cliente —que lo toma de schema.Contexto.Offline— y lo
//     declara como opción del constructor, nunca como parámetro de cada
//     operación. Tiene que conocerse al construir, porque crear la base de
//     datos y migrarla ocurre al abrir y no al operar (FR-015).
//   - No borra la base de datos por su cuenta, ni siquiera para arreglarla: un
//     fichero que no es una base utilizable, o que trae un esquema de una
//     versión que este binario no conoce, es un fallo explícito y el fichero
//     queda intacto (FR-027, FR-028).
//   - No cifra nada: lo que se guarda va tal cual en el disco de la persona,
//     con el directorio y el fichero creados con acceso reservado a su cuenta
//     (FR-021). Quien no quiera guardar algo, no lo escribe.
//   - No interpreta la clave ni el contenido: los dos son opacos. La caché no
//     deriva, normaliza ni valida la clave más allá de rechazar la vacía, y no
//     sabe qué hay dentro del contenido ni le añade metadatos propios; el
//     esquema de claves por fuente es de quien llama (FR-006, FR-011).
//   - No fija ninguna vigencia por omisión ni la negocia: la vigencia llega en
//     cada escritura, y una vigencia menor o igual que cero se rechaza sin
//     escribir nada, porque «no guardar esto» se expresa no escribiendo
//     (FR-010).
//   - No decide ningún código de salida: declara la clase del fallo y es el
//     kernel quien la traduce, una sola vez por invocación (FR-033).
//   - No desaloja entradas, no las cuenta, no limita su tamaño y no ofrece
//     ningún verbo de mantenimiento: nada de eso está en el alcance del hito.
//
// Este fichero es la única excepción declarada a «un fichero de test por
// fichero de código»: no contiene ninguna declaración, solo este comentario.
package cache
