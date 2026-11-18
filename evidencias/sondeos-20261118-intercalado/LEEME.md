# Evidencia del sondeo intercalado antes de fusionar H41.4 (#274)

Material del sondeo de aceptación que fija el ADR 0056, lanzado el 2026-11-18 en delorean (Linux) por una sesión
interactiva a petición de la persona, fuera de cualquier run (ADR 0032). Como todo `evidencias/`, un run del workflow
no la escribe.

No es evidencia de una fuente: ninguno de sus ficheros sale de una fila de `docs/SOURCES.md`. `manifiesto.json` da la
huella y el tamaño de cada fichero.

## El sondeo

`make evals-laboratorio` con `normativa-producto`, solo la eval 04 (`EVAL=04`), solo el modo herramienta,
`claude-sonnet-5-5`, Claude Code 2.1.326 (la versión del job, puesta delante en el `PATH`), `INSTALACION=producto` y
concurrencia 5, en 16 tandas de 40 sesiones, alternando las dos filas y empezando por `control`, de 09:28 a 09:53 UTC:

| Fila | Árbol desde el que se lanza | Skills |
|---|---|---|
| `control` | un árbol fuera del repositorio en `ab951c22`, la base de la rama | las cuatro de `main` |
| `producto` | `1ec12710`, la cabeza de `038-h41-4-las-skills-del` | las cuatro, con las descripciones cortas |

Cada tanda va en `~/code/kitlegal-sondeos/h41-4b/<fila>/<tanda>`. Son 640 sesiones y todas terminaron: ninguna sale
«sin terminar» en las tablas y ningún `sesion.err` tiene contenido.

## Lo que da

| Fila | No activan `normativa` antes | Activan otra skill | No activan ninguna |
|---|---|---|---|
| `control` | **33** [7,4–14,1 %] | 0 [0,0–1,2 %] | 33 |
| `producto` | **0** [0,0–1,2 %] | 0 [0,0–1,2 %] | 0 |

Con el criterio del ADR 0056, `producto` tiene como mucho 2 de 320 y no queda por encima de `control`: se cumple. La
lectura está en el ADR 0057.

## Qué hay en cada fichero

- **`tabla-control.md`** y **`tabla-producto.md`**: la tabla de cada fila tal como la imprime
  `make laboratorio-tabla DIRECTORIO=…/kitlegal-sondeos/h41-4b/<fila>`, que junta las 8 tandas de la fila, con el
  detalle de sus 320 sesiones. Regeneradas desde las sesiones, coinciden byte a byte con las del lanzamiento.

Las sesiones, con sus trazas, se conservan en `~/code/kitlegal-sondeos/h41-4b/` de quien lo lanzó y no se versionan
aquí.
