# Evidencia del sondeo de la persona antes de fusionar H41.3 (#270)

Material del sondeo que la entrada de H41.3 («Aceptación») y FR-030 y FR-031 dejan a una persona antes de fusionar,
lanzado el 2026-11-17 en delorean (Linux), en una terminal interactiva y fuera de cualquier run (ADR 0032). Como todo
`evidencias/`, un run del workflow no la escribe.

No es evidencia de una fuente: ninguno de sus ficheros sale de una fila de `docs/SOURCES.md`. `manifiesto.json` da la
huella y el tamaño de cada fichero.

## El sondeo

`make evals-laboratorio` con el arnés de H41.3 sobre la cabeza de la #270 (`e12146f1`, rama
`037-h41-3-por-que-normativa`): `normativa-producto`, solo la eval 04 (`EVAL=04`), solo el modo herramienta,
`claude-sonnet-5-5`, 160 sesiones por fila y concurrencia 5, en las cuatro filas de la entrada:

| Fila | Skills instaladas | Instalación | Presupuesto |
|---|---|---|---|
| `sola` | `normativa` | `normativa` | — |
| `control` | las cuatro del producto | `producto` | — |
| `diagnostico` | las cuatro del producto | `producto` | `SLASH_COMMAND_TOOL_CHAR_BUDGET=1000000` |
| `remedio` | las cuatro, con las descripciones de la variante | `producto-descripciones-cortas` | — |

**Versión de Claude Code: 2.1.326, no la 2.1.284 que dice la entrada.** Lo decidió Jorge el 2026-11-16: medir con la
versión del uso real, la misma que pasa a fijar el job en la #271 (`docs/USO.md`, entrada de ese día). Cada
`sondeo.json` dice el modelo y la versión de su fila, y la tabla los da por fila. `SLASH_COMMAND_TOOL_CHAR_BUDGET` está
en el binario de 2.1.326 como en el de 2.1.284.

Son 640 sesiones y todas terminaron: ninguna sale «sin terminar» en la tabla y ningún `sesion.err` tiene contenido.

## Lo que da

Sesiones de 160, con intervalos de Wilson al 95 %:

| Fila | Activan antes | Activan otra skill | No activan ninguna |
|---|---|---|---|
| `sola` | 159 [96,5–99,9 %] | 0 [0,0–2,3 %] | 1 [0,1–3,5 %] |
| `control` | 146 [85,8–94,7 %] | 0 [0,0–2,3 %] | **14** [5,3–14,2 %] |
| `diagnostico` | 153 [91,2–97,9 %] | 0 [0,0–2,3 %] | **7** [2,1–8,8 %] |
| `remedio` | 160 [97,7–100,0 %] | **18** [7,2–17,1 %] | 0 [0,0–2,3 %] |

Las 22 sesiones que no activan ninguna skill (`sola` 82; `control` 24, 32, 38, 39, 50, 69, 75, 85, 101, 114, 116, 119,
147 y 155; `diagnostico` 12, 30, 62, 83, 95, 115 y 142) son del tipo de las de H41.2 y del job: piden `boe_articulo`
con `ToolSearch`, lo llaman y responden sin ningún `tool_use` de `Skill`. Las 18 de `remedio` que activan otra skill
activan `consulta-juridica` y `normativa`, las dos antes de la primera herramienta.

## La lectura, con las reglas de la entrada

Cada comparación es entre los intervalos de dos filas en «No activan ninguna», salvo donde la entrada dice otra cosa.

- **La instalación: la caída es de la instalación completa.** `control` [5,3–14,2 %] queda por encima de `sola`
  [0,1–3,5 %] y los intervalos no se tocan. El control reproduce la caída, así que las otras dos comparaciones
  deciden.
- **El presupuesto: no se demuestra que sea la causa.** `diagnostico` [2,1–8,8 %] queda por debajo de `control`
  [5,3–14,2 %], pero los intervalos se tocan. Con el listado sin recortar siguen cayendo 7 de 160.
- **El remedio: no se cumple la regla, así que el hito siguiente no lo aplica tal cual.** En «No activan ninguna»,
  `remedio` [0,0–2,3 %] queda por debajo de `control` [5,3–14,2 %] y los intervalos no se tocan. Pero su «Activan
  otra skill» [7,2–17,1 %] queda por encima de la del control [0,0–2,3 %], y la entrada pide que no quede por encima.
  Esas 18 sesiones activan también `normativa` antes de la primera herramienta. Que eso cuente o no como un
  defecto, y qué se mide después, lo dice el ADR que lee esta tabla. Ese ADR debe saber dos cosas de la
  descripción corta de `consulta-juridica`. Quita «también cuando el agente cree saber la respuesta», la frase que la
  investigación del plan (D2) conserva en `normativa` («aunque el agente crea saberlo») porque es lo que mide la eval
  04. Y conserva «antes que las demás skills de kitlegal», que D2 dejó para no confundir el remedio con otra causa en
  esta columna. Quitar la frase haría pensar en menos activaciones de `consulta-juridica`, no en más, así que el texto
  no explica las 18 por sí solo.

Esta tabla no decide la fusión, que depende de los controles (entrada de H41.3, «Aceptación»).

## Qué hay en cada fichero

- **`tabla-normativa.md`**: la tabla tal como la imprime `make laboratorio-tabla DIRECTORIO=…/kitlegal-sondeos/h41-3`,
  con el detalle de las 640 sesiones. Regenerada sin modelo desde las sesiones, coincide byte a byte con la que
  imprimió el lanzamiento.

No se guarda la salida de los cuatro lanzamientos: se imprimió en la terminal de quien los lanzó y no se conservó.

Las sesiones, con sus trazas, se conservan en `~/code/kitlegal-sondeos/h41-3/` de quien lo lanzó y no se versionan
aquí.
