// Package httpx es el único punto del módulo por el que sale una petición HTTP
// (regla R2 de docs/ROADMAP.md §2). Lo que garantiza no es una lista de buenas
// prácticas que cada fuente legal deba recordar, sino la forma misma de pedir:
// toda petición que sale de aquí —también las que el paquete origina por su
// cuenta, la de robots.txt y cada salto de una cadena de redirecciones— lleva
// la identificación del proyecto, ha consultado antes el robots.txt de su
// sitio, espera su turno en el ritmo de ese sitio, reintenta solo lo que se
// puede reintentar y termina cuando termina el contexto de quien la pidió
// (FR-001, FR-002).
//
// El ritmo de un sitio es de cada cliente, salvo que varios reciban el mismo
// Ritmo con ConRitmo: entonces esperan turno entre todos, que es lo que necesita
// un proceso que construye un cliente por llamada. Lo demás es siempre de cada
// uno: lo que un cliente recuerda del robots.txt de un sitio vive y termina con
// él.
//
// Por qué no se puede usar mal:
//
//   - Ninguna de esas garantías se declara como configuración, sino como la
//     cadena por la que toda petición baja, así que no hay opción que las
//     desactive ni descuido que las omita: la imposibilidad es por
//     construcción, y los controles mecánicos (noctx, bodyclose, la regla R2 y
//     el test de arquitectura) son la red secundaria, no el mecanismo.
//   - La única operación de red es Pedir —la del cliente y la de una Consulta
//     suya, que es el mismo pedido—, que exige el contexto de cancelación como
//     primer parámetro y el contexto de ejecución del kernel como segundo: pedir
//     algo sin contexto no compila (FR-003, FR-061).
//   - El paquete no exporta *http.Client, http.RoundTripper, *http.Request ni
//     *http.Response, ni ninguna forma de construir o emitir una petición desde
//     fuera. Lo que Pedir devuelve es Respuesta, un tipo de este paquete con el
//     cuerpo ya leído —el *http.Response se abre y se cierra aquí dentro—, de
//     modo que un adaptador de fuente pueda usar la respuesta sin importar la
//     biblioteca HTTP (FR-002, FR-053).
//
// La única excepción a GET y HEAD. No existe en el módulo ninguna llamada HTTP
// con otro método, con una sola excepción, que es la del principio I de la
// constitución: el envío del formulario de consulta de un buscador público,
// cuando la fila de la fuente en docs/SOURCES.md, revisada por una persona,
// declara ese formulario. Es una consulta sin identidad, que no presenta, firma
// ni cambia nada en la fuente; la emite solo este paquete, a la dirección que
// declara la fila, y ningún adaptador la usa para otra cosa (ADR 0036).
//
// Aquí esa excepción tiene una sola forma. El cliente de la fuente declara la
// dirección con ConFormulario; el envío es un POST con los campos de
// Peticion.Campos, y solo sale de una Consulta de ese cliente, a esa dirección
// exacta y con al menos un campo. Cualquier otro POST —desde Cliente.Pedir, a
// otra dirección o sin campos— y cualquier otro método siguen siendo un fallo de
// argumentos antes de abrir nada, también en ensayo. El envío baja por la misma
// cadena que toda petición, así que no pierde ninguna de sus garantías, y dentro
// de una Consulta no se sigue ninguna redirección: el formulario no se reenvía a
// ninguna parte. Las cookies que el sitio da en una Consulta acompañan a las
// peticiones de esa Consulta y a nada más —ni a la siguiente, ni a
// Cliente.Pedir, ni al robots.txt—, y no se escriben en disco (FR-030 a FR-033
// de H23).
//
// Lo que el paquete deliberadamente no hace: no presenta nada, no decide el
// código de salida —declara la clase del fallo y es el kernel quien la traduce
// (FR-031)— y no realiza ninguna acción que exija identidad humana.
//
// Este fichero es la única excepción declarada a «un fichero de test por
// fichero de código»: no contiene ninguna declaración, solo este comentario.
package httpx
