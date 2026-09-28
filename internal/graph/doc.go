// Package graph es el adaptador del grafo del mundo: implementa el puerto
// core.GraphStore sobre SQLite sin cgo, en un único fichero world.db, y da a
// los verbos del applet graph la lectura de lo guardado (docs/ADR/0014, H7
// research.md D1). Es, con internal/cache, el lugar del módulo donde se abre,
// se migra y se consulta una base de datos (regla R3 de docs/ROADMAP.md §2). El
// dominio —qué se guarda, cómo se fusionan las observaciones, qué se rechaza y
// qué avisa graph check— vive en internal/core/grafo; este paquete sabe que
// detrás hay un fichero.
//
// Lo que garantiza:
//
//   - world.db vive en el directorio de ConDirectorio o, sin la opción, en el
//     de la caché, con su misma regla y sus mismos errores de clase
//     «argumentos» (cache.Directorio). La ruta se resuelve al leer o al
//     entregar, nunca al construir (FR-001, FR-011).
//   - La entrega crea world.db en su sitio, en 0600 y con cada directorio que
//     falta en 0700, sin temporal ni enlace: una creación interrumpida deja
//     como mucho world.db sin esquema, que la entrega siguiente completa (H7.1
//     FR-071).
//   - El esquema está versionado con migraciones que viajan dentro del binario
//     y se aplican dentro de la transacción que las pide: entran enteras o no
//     entra nada, y un esquema de una versión posterior no se toca (FR-003,
//     FR-012, FR-013).
//   - Toda espera por un bloqueo de otra invocación va por tramos de 100 ms
//     dentro del motor, mira el contexto entre tramo y tramo y dura como mucho
//     5 s: el contexto terminado es «fuente-no-disponible» y la espera propia
//     agotada, «inesperado» (FR-014).
//   - Todo fallo es un *Error que declara su clase por schema.ConClase y cuyo
//     mensaje empieza por «grafo: » y nombra world.db
//     (contracts/almacen-world-db.md §6). Ninguna ruta termina en panic.
//   - Nulo, el almacén de --no-graph, descarta cualquier lote sin resolver
//     ninguna ruta ni abrir nada (FR-031).
//   - Ninguna declaración exportada nombra database/sql ni el controlador de
//     SQLite: la conexión no sale nunca del paquete.
//
// Lo que deliberadamente no hace:
//
//   - No registra eventos con slog ni escribe en ningún descriptor: lo único
//     que una entrega fallida deja en la salida de error es la línea que el
//     kernel escribe con el error de Apply (FR-033).
//   - No borra world.db ni lo rehace: un fichero que el binario no puede usar
//     es un fallo explícito con su ruta y su causa, sin ninguna promesa sobre
//     sus bytes (H7.1 FR-070), y uno de otra versión, además, no se toca
//     (FR-012).
//   - No firma ningún sobre ni decide ningún código de salida: la procedencia
//     de cada operación es la del lote que construye el kernel, y la clase del
//     fallo la traduce el kernel (FR-021, ADR 0023).
//   - No importa internal/source ni internal/render (regla R6, FR-092): el
//     grafo no sabe de dónde viene lo observado ni cómo se presenta; de
//     internal/cache solo usa la regla de ubicación.
//
// Este fichero no contiene declaraciones, solo este comentario.
package graph
