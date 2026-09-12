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
// Por qué no se puede usar mal:
//
//   - Ninguna de esas garantías se declara como configuración, sino como la
//     cadena por la que toda petición baja, así que no hay opción que las
//     desactive ni descuido que las omita: la imposibilidad es por
//     construcción, y los controles mecánicos (noctx, bodyclose, la regla R2 y
//     el test de arquitectura) son la red secundaria, no el mecanismo.
//   - La única operación de red es Pedir, que exige el contexto de cancelación
//     como primer parámetro y el contexto de ejecución del kernel como segundo:
//     pedir algo sin contexto no compila (FR-003, FR-061).
//   - El paquete no exporta *http.Client, http.RoundTripper, *http.Request ni
//     *http.Response, ni ninguna forma de construir o emitir una petición desde
//     fuera. Lo que Pedir devuelve es Respuesta, un tipo de este paquete con el
//     cuerpo ya leído —el *http.Response se abre y se cierra aquí dentro—, de
//     modo que un adaptador de fuente pueda usar la respuesta sin importar la
//     biblioteca HTTP (FR-002, FR-053).
//
// Lo que el paquete deliberadamente no hace: no presenta nada, no decide el
// código de salida —declara la clase del fallo y es el kernel quien la traduce
// (FR-031)— y no realiza ninguna acción que exija identidad humana.
//
// Este fichero es la única excepción declarada a «un fichero de test por
// fichero de código»: no contiene ninguna declaración, solo este comentario.
package httpx
