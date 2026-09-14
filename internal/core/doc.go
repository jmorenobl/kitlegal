// Package core declara los puertos del dominio: las interfaces con las que el
// dominio nombra lo que necesita —guardar y recuperar, pedir a una fuente,
// anotar en el grafo— sin saber quién lo hace ni con qué. Quien las implementa
// es siempre un adaptador, y la dependencia va hacia dentro: el dominio no
// importa el kernel, ni ningún adaptador, ni la entrada y salida de la
// biblioteca estándar (regla R1 de docs/ROADMAP.md §2).
//
// Es la raíz del dominio que el §2 del roadmap reserva a los puertos, junto al
// subpaquete schema, el contrato del sobre de salida. Cada puerto nace con el
// hito que lo estrena: Cache nació con H3, y Source nace con H4 con la firma
// que fija docs/ADR/0015. Los que todavía no existen —Fetcher, GraphStore,
// Renderer— nacerán aquí con el suyo.
//
// Este fichero no contiene ninguna declaración, solo este comentario.
package core
