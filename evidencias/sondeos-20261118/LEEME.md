# Evidencia del sondeo de la persona antes de fusionar H41.4 (#274)

Material del sondeo de aceptación que pide el ADR 0055, punto 4, antes de fusionar H41.4, lanzado el 2026-11-18 en
delorean (Linux) por una sesión interactiva a petición de la persona, fuera de cualquier run (ADR 0032). Como todo
`evidencias/`, un run del workflow no la escribe.

No es evidencia de una fuente: ninguno de sus ficheros sale de una fila de `docs/SOURCES.md`. `manifiesto.json` da la
huella y el tamaño de cada fichero.

## El sondeo

`make evals-laboratorio` con `normativa-producto`, solo la eval 04 (`EVAL=04`), solo el modo herramienta,
`claude-sonnet-5-5`, Claude Code 2.1.326 (la versión del job, puesta delante en el `PATH`, porque la instalada se había
actualizado sola a la 2.1.327), `INSTALACION=producto`, 320 sesiones por fila y concurrencia 5:

| Fila | Árbol desde el que se lanza | Skills |
|---|---|---|
| `control` | un árbol fuera del repositorio en `ab951c22`, la base de la rama | las cuatro de `main` |
| `producto` | `77d6ed7c`, la cabeza de `038-h41-4-las-skills-del` | las cuatro, con las descripciones cortas |

`control` corrió de 07:53 a 08:05 UTC, y `producto`, de 08:05 a 08:17. Son 640 sesiones y todas terminaron: ninguna
sale «sin terminar» en las tablas y ningún `sesion.err` tiene contenido.

## Lo que da

| Fila | No activan `normativa` antes | Activan otra skill | No activan ninguna |
|---|---|---|---|
| `control` | **7** [1,1–4,4 %] | 0 [0,0–1,2 %] | 7 |
| `producto` | **0** [0,0–1,2 %] | 26 [5,6–11,6 %] | 0 |

Las 7 caídas de `control` (sesiones 70, 198, 215, 225, 233, 252 y 270) piden `boe_articulo` con `ToolSearch`, lo
llaman y responden sin ningún `tool_use` de `Skill`. Las 26 de `producto` que activan otra skill activan
`consulta-juridica` y `normativa`, las dos antes de la primera herramienta.

Con el criterio del ADR 0055, punto 4, los intervalos se tocan (1,06 % frente a 1,19 %) y el criterio no se cumple.
Qué se hace con eso lo dice el ADR 0056.

## Qué hay en cada fichero

- **`tabla-control.md`** y **`tabla-producto.md`**: la tabla de cada fila tal como la imprime
  `make laboratorio-tabla DIRECTORIO=…/kitlegal-sondeos/h41-4/<fila>`, con el detalle de sus 320 sesiones. Se sacan
  por fila porque las dos tienen la misma configuración, y la tabla de la carpeta común las suma en una sola de 640.

Las sesiones, con sus trazas, se conservan en `~/code/kitlegal-sondeos/h41-4/` de quien lo lanzó y no se versionan
aquí.
