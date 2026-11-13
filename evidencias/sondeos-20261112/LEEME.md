# Evidencia del sondeo de la persona antes de fusionar H41.2 (#256)

Material del sondeo que la entrada de H41.2 («Aceptación») y FR-030 y FR-031 dejan a una persona antes de fusionar,
lanzado el 2026-11-12 en una sesión interactiva, fuera de cualquier run (ADR 0032). Como todo `evidencias/`, un run del
workflow no la escribe.

No es evidencia de una fuente: ninguno de sus ficheros sale de una fila de `docs/SOURCES.md`. `manifiesto.json` da la
huella y el tamaño de cada fichero.

## El sondeo

`make evals-laboratorio` con el arnés de H41.2 sobre la cabeza de la #256 (`8191a90e`, rama `036-h41-2-por-qu-e`):
`normativa-producto`, solo la eval 04 (`EVAL=04`), solo el modo herramienta, `claude-sonnet-5-5`, 40 sesiones por fila
y concurrencia 5, en las cuatro filas de FR-030 (`specs/036-h41-2-por-qu-e/contracts/arnes.md` §1):

| Fila | Claude Code | Skills instaladas |
|---|---|---|
| `sola-2.1.284` | 2.1.284, la del job, instalada aparte con `npm` | `normativa` |
| `todas-2.1.284` | 2.1.284 | las cuatro del producto (`INSTALACION=producto`) |
| `sola-2.1.322` | 2.1.322, la del día | `normativa` |
| `todas-2.1.322` | 2.1.322 | las cuatro del producto |

Son 160 sesiones y todas terminaron: ninguna sale «sin terminar» en la tabla y ningún `sesion.err` tiene contenido. Cada
`sondeo.json` dice el modelo, la versión y las skills instaladas de su fila, y la tabla las da por fila.

## Lo que da

Sesiones de 40 que activan `normativa` antes de la primera herramienta, que activan otra skill y que no activan
ninguna, con intervalos de Wilson al 95 %:

| Fila | Activan antes | Activan otra skill | No activan ninguna |
|---|---|---|---|
| `sola-2.1.284` | 40 [91,2–100,0 %] | 0 [0,0–8,8 %] | 0 [0,0–8,8 %] |
| `todas-2.1.284` | 38 [83,5–98,6 %] | 0 [0,0–8,8 %] | **2** [1,4–16,5 %] |
| `sola-2.1.322` | 40 [91,2–100,0 %] | 0 [0,0–8,8 %] | 0 [0,0–8,8 %] |
| `todas-2.1.322` | 36 [76,9–96,0 %] | 0 [0,0–8,8 %] | **4** [4,0–23,1 %] |

- **Las seis caídas son del mismo tipo que la del job.** `todas-2.1.284` 27 y 40 y `todas-2.1.322` 09, 26, 27 y 39
  piden la herramienta con `ToolSearch` (`select:mcp__kitlegal__boe_articulo`), leen el art. 66 de la LGT con
  `boe_articulo` y responden sin ningún `tool_use` de `Skill`. Ninguna activa otra skill, tampoco
  `consulta-juridica`. Es lo que hizo la sesión `-02` de la primera medición del cierre de H41.2 (`activaciones: []`).
- **Solo caen con las cuatro skills instaladas**, con las dos versiones: 0 de 80 con `normativa` sola [0,0–4,6 %] y 6
  de 80 con las cuatro [3,5–15,4 %]. Por versión: 2 de 80 con 2.1.284 [0,7–8,7 %] y 4 de 80 con 2.1.322 [2,0–12,2 %].
- **Con el criterio que fijó el hito, ninguna fila separa la caída de otra.** El criterio son los intervalos de cada
  fila. Los de las dos filas con caídas tocan los de `normativa` sola: 4,0 % frente a 8,8 % en 2.1.322, y 1,4 % frente
  a 8,8 % en 2.1.284. Tampoco se separan juntando las filas por instalación: 3,5 % frente a 4,6 %. Fuera del criterio,
  la prueba exacta de Fisher unilateral da p ≈ 0,014 para que las seis caigan con las cuatro instaladas por azar, y
  p ≈ 0,34 para que la 2.1.322 caiga más que la 2.1.284.
- **La tasa con la instalación del job.** Las dos filas `todas` son la instalación del job: 6 de 80 [3,5–15,4 %]. Las
  160 sesiones juntas dan 6 de 160 [1,7–7,9 %]. Cada medición del job abre 3 sesiones de la eval 04 en modo
  herramienta con `claude-sonnet-5-5`.

Lo que se hace con esto lo decide el ADR que lee esta tabla (entrada de H41.2, «Aceptación»). Esta tabla no decide la
fusión, que depende de los controles.

## Qué hay en cada fichero

- **`tabla-normativa.md`**: la tabla tal como la imprime `make laboratorio-tabla DIRECTORIO=…/kitlegal-sondeos/h41-2`,
  regenerada sin modelo desde las sesiones, con el detalle de las 160. Coincide con la que imprimió el lanzamiento.
- **`salida.txt`**: lo que imprimieron los cuatro lanzamientos.

Las sesiones, con sus trazas, se conservan en `~/Projects/kitlegal-sondeos/h41-2/` de quien lo lanzó y no se versionan
aquí.
