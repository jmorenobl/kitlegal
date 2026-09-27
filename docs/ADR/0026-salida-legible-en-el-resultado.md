# 0026 · La forma para personas viaja en el `Resultado`: `skills` la cuenta, el kernel la escribe

- **Estado**: aceptada
- **Fecha**: 2026-09-27
- **Hito**: transversal (tras v0.3.0, antes de H7). Amplía `schema.Resultado` sin tocar el contrato del ADR 0005, como
  hizo el ADR 0014 con las operaciones de grafo, y precisa la regla «sin `--json`, la tabla mínima» del ADR 0006 y de
  H1.

## Contexto y problema

Sin `--json`, todo applet sale por la **tabla mínima** de `internal/render/tabla.go`: las cuatro líneas de procedencia
y el contenido de `data` aplanado a pares ruta/valor (`0.enlaces.0.host  claude`). Es una forma genérica, que no
conoce applets y que conserva la cita, y para `boe` y `territorio` es la adecuada: los usan los agentes con `--json`, y
quien la mira a mano quiere ver `fuente`, `url`, `fecha_consulta` y `hash`.

`kitlegal skills install`, `list` y `doctor` son otra cosa. Los ejecuta a mano una persona que acaba de instalar
kitlegal, muchas veces sin perfil técnico, y lo que ve —`fuente kitlegal.skills`, `hash sha256:…`,
`0.enlaces.0.modo enlace`— no le dice dónde han quedado las skills, qué agente las verá ni qué hacer ahora
(`docs/USO.md`, 2026-09-27). Hace falta un texto para esa persona, y hace falta decidir **por dónde llega**, porque
el presentador no conoce applets por diseño (cabecera de `internal/render/render.go`: añadir una forma de presentación
no obliga a tocar ningún applet) y el applet no escribe en ningún descriptor (ADR 0005).

Restricciones que no se negocian: `--json` no cambia ni un byte, ni el sobre, ni `schemas/instalacion.json`, ni los
códigos de salida (ADR 0023); y H7 va a añadir al `Resultado` las operaciones de grafo (ADR 0014), así que el
mecanismo tiene que convivir con ese campo sin anticiparlo.

## Opciones consideradas

1. **Que el applet escriba el texto por su cuenta.** Rechazada por el ADR 0005: un applet no tiene por dónde emitir,
   y en cuanto uno escribiera, la forma de la salida dependería de quién la emitió.
2. **Que `render` conozca los tipos de `data` de `skills`** y los presente con una plantilla por tipo. Rechazada: el
   presentador dejaría de ser genérico, importaría `internal/core/instalacion`, y cada applet nuevo que quisiera una
   forma legible tendría que tocar `render`, justo lo contrario de lo que su cabecera promete.
3. **Que `data` implemente una interfaz `Legible()`** que `render` detecte por aserción de tipo. Es genérica, pero
   pone la presentación en los tipos del dominio, que no saben ni HOME ni la versión del binario, y la vuelve
   implícita: un tipo la tiene o no, y nada en el flujo lo dice. Rechazada.
4. **Un campo más en `schema.Resultado`**, `Legible string`: el applet compone el texto a partir de los mismos
   valores que van en `Datos`, en el mismo instante, y el kernel, en el único punto que emite, lo escribe sin `--json`
   en lugar de la tabla. Es el mismo camino por el que ya viaja `Ensayo` (ADR 0011): lo que el applet quiere que una
   persona lea sale por el `Resultado` y lo presenta el kernel. Elegida.

## Decisión

1. `schema.Resultado` gana **`Legible string`**: el contenido contado para una persona. Vacío significa «la tabla
   mínima», que sigue siendo la forma por omisión y la de todo applet que consulta fuentes. Su valor cero es válido y
   ningún applet tiene que rellenarlo.
2. **El kernel decide, no `render`**: `cli.Montador.Emitir` monta el sobre siempre —la procedencia válida y la huella
   se exigen igual aunque no se escriban— y, si no se pidió `--json` y `Legible` no está vacío, lo escribe por
   `Presentador.Texto` en lugar de presentar el sobre. `render` no cambia: sigue sin conocer applets y `Texto` sigue
   siendo el texto para una persona que va a la salida estándar. La interfaz `Presentador` no gana métodos.
3. **Con `--json` nada cambia**: el sobre es el único camino de la forma legible por máquina, `Legible` no entra en el
   sobre ni en la huella y los códigos de salida son los del ADR 0023. Un fallo sigue sin tener forma legible distinta
   del mensaje en la salida de error.
4. **Solo `skills` lo usa**, en `install`, `list` y `doctor` (`internal/app/instalacion_legible.go`): dónde han
   quedado las skills y quién las lee de ahí, cada skill con su estado o su versión y sus entradas de host con el
   nombre de la marca (Claude Code, Antigravity), HOME abreviado a `~`, el agente que no las verá con la orden que lo
   enlaza, y en `doctor` «todo en orden» o cada hallazgo explicado en una frase con su orden lista para copiar —la
   orden tal cual, sin `~`, porque es la que se copia y se pega—. Sin jerga del sobre, sin colores ni secuencias de
   escape, y determinista. `boe` y `territorio` no lo usan: ahí la procedencia es lo que hace citable la respuesta.
5. `--dry-run` se alinea con el mismo vocabulario: cada entrada de host lleva el nombre de la marca y las rutas se
   abrevian igual. Sigue en la salida de error con su prefijo (ADR 0011).

## Consecuencias

- Quien instala ve un texto que le dice qué ha pasado y qué hacer; los agentes y las evals, que leen `--json`, no
  notan nada.
- El mecanismo queda disponible para cualquier applet cuyo consumidor sea una persona, sin tocar `render` ni la
  interfaz `Presentador`: basta rellenar `Legible`. Un applet que lo use tiene que componer el texto de los mismos
  valores que pone en `Datos`, y ese es el único punto que la disciplina vigila; el kernel garantiza el resto.
- H7 añade su campo al mismo `Resultado` (ADR 0014); `Legible` no interfiere: el kernel escribe uno y aplica el otro,
  y ninguno entra en el sobre.
- `Resultado` tiene cuatro campos y no dos, y el ADR 0005 lo anota, como anotó el ADR 0014. Es el precio de que la
  presentación para personas no viva ni en el applet ni en el presentador.
- La regla «sin `--json`, la tabla mínima» de H1 pasa a ser «sin `--json`, la tabla mínima salvo que el applet cuente
  su resultado». Un consumidor que analizara la tabla de `skills` deja de poder hacerlo; la tabla nunca fue un
  contrato para máquinas, que tienen `--json`.
