# Evidencia del sondeo de la persona antes de fusionar H41.1 (#254)

Material del sondeo que la entrada de H41.1 («Aceptación») y SC-003 dejan a una persona antes de fusionar, lanzado el
2026-11-10 en una sesión interactiva, fuera de cualquier run (ADR 0032). Como todo `evidencias/`, un run del workflow
no la escribe.

No es evidencia de una fuente: ninguno de sus ficheros sale de una fila de `docs/SOURCES.md`. `manifiesto.json` da la
huella y el tamaño de cada fichero.

## El sondeo

`make evals-laboratorio` con el arnés de H41, sin cambios, sobre la cabeza de la #254 (`280b7cca`, rama
`035-h41-1-plazos-y-territorio`): `claude-sonnet-5-5`, Claude Code 2.1.322, `plazos-y-territorio-control` y
`plazos-y-territorio-producto` en los dos modos, las evals 02, 05, 10 y 11 con diez sesiones cada una, 40 por variante
y modo y concurrencia 5 (`specs/035-h41-1-plazos-y-territorio/quickstart.md` §8). Son cuatro filas y 160 sesiones, y
todas terminaron: ninguna sale «sin terminar» en la tabla y ningún `sesion.err` tiene contenido. `producto` es
`skills/plazos-y-territorio` de esa cabeza, la forma `sin-condicion` con M1 (ADR 0053).

La versión de Claude Code no es la del sondeo de H41 (2.1.319, `evidencias/sondeos-20261109/`). Por eso la aceptación
compara `producto` con `control` dentro de este sondeo (ADR 0053, «Consecuencias»).

## La aceptación de la fusión (ADR 0053, punto 2)

Intervalos de Wilson al 95 %. Sesiones de 40 que activan la skill antes de la primera herramienta y que abren cada
referencia a cuya condición llegan las evals:

| Variante | Modo | Activan antes | `administrativo.md` | `calculo_de_plazos.md` | `municipios.md` |
|---|---|---|---|---|---|
| `control` | orden | 40 [91,2–100,0 %] | 30 [59,8–85,8 %] | 30 [59,8–85,8 %] | 11 [16,1–42,8 %] |
| `producto` | orden | 39 [87,1–99,6 %] | 29 [57,2–83,9 %] | 29 [57,2–83,9 %] | **39** [87,1–99,6 %] |
| `control` | herramienta | 40 [91,2–100,0 %] | 30 [59,8–85,8 %] | 28 [54,6–81,9 %] | 10 [14,2–40,2 %] |
| `producto` | herramienta | 40 [91,2–100,0 %] | 30 [59,8–85,8 %] | 30 [59,8–85,8 %] | **40** [91,2–100,0 %] |

- **En cada modo, `producto` no queda por debajo de `control`**, ni en «Activan antes» ni en `administrativo.md` ni en
  `calculo_de_plazos.md`: en cada celda, los dos intervalos se tocan. Se cumple.
- **En `municipios.md`, `producto` queda por encima de `control` en los dos modos**, con 39 y 40 de 40 frente a 11 y 10,
  e intervalos que no se tocan. Se cumple.

La propuesta cumple la aceptación de la fusión.

`jerarquia_normativa.md`, cuya línea no cambia, la abren 1 y 0 sesiones de `control` y 3 y 0 de `producto`. Ninguna
eval llega a su condición, así que no entra en la aceptación.

El desglose por eval (`recuento.md`):

- **`municipios.md`.** `producto` la abre en 10 de 10 en cada eval y modo, salvo en la eval 10 en modo orden, con 9:
  la sesión 06 no activa la skill y no abre ninguna referencia. `control` la abre en 0 a 5 de 10 en las evals 05, 10 y
  11, en 1 de 10 en la eval 02 en modo orden y en 9 de 10 en la eval 02 en modo herramienta.
- **`calculo_de_plazos.md` y `administrativo.md`.** En las evals 05, 10 y 11, `producto` las abre en 10 de 10, salvo
  esa misma sesión de la eval 10 en modo orden (9 de 10); `control`, en 9 o 10 de 10. En la eval 02 no las abre
  ninguna variante: ninguna eval de la familia pide ahí un plazo.
- **Lo que no se lee.** En modo orden, 9 sesiones de `control` no abren ninguna referencia, frente a 1 de `producto`.

**Cuánto se mueve el control.** `control` abre `municipios.md` en 11 y 10 de 40, frente a 3 y 11 en el sondeo de H41
con Claude Code 2.1.319. `producto` repite lo que dio `plazos-y-territorio-sin-condicion` en aquel sondeo (39 y 40 de
40): la forma sin condición no depende de la versión de Claude Code en lo que se ha medido.

## Qué hay en cada fichero

- **`tabla-plazos-y-territorio.md`**: la tabla tal como la imprime `make laboratorio-tabla`, con el detalle de sus 160
  sesiones.
- **`recuento.md`**: la cuenta de `guiones/recuento.py` (`python3 -I guiones/recuento.py tabla-plazos-y-territorio.md`).
  Su título dice «del sondeo de H41» porque el guion es el mismo, sin cambios.
- **`salida.txt`**: lo que imprimió el lanzamiento.
- **`guiones/recuento.py`**: el guion del recuento, copia sin cambios del de `evidencias/sondeos-20261109/`.

Las sesiones, con sus trazas, se conservan en `~/Projects/kitlegal-sondeos/h41-1/` de quien lo lanzó y no se versionan
aquí.
