# Evidencia del sondeo de la persona antes de fusionar H41: el relanzamiento (#248)

Material del sondeo que la entrada de H41 («Aceptación»), FR-052 y FR-053 dejan a una persona antes de fusionar,
lanzado el 2026-11-09 en una sesión interactiva, fuera de cualquier run (ADR 0032). Como todo `evidencias/`, un run del
workflow no la escribe.

No es evidencia de una fuente: ninguno de sus ficheros sale de una fila de `docs/SOURCES.md`. `manifiesto.json` da la
huella y el tamaño de cada fichero.

## El sondeo

`make evals-laboratorio` con el arnés que entrega H41, sobre la cabeza de la #248 (`7bca2cf8`, rama
`034-h41-las-skills-del`), sin cambios: `claude-sonnet-5-5`, 40 sesiones por variante y modo (diez por eval),
concurrencia 5, Claude Code 2.1.319. Son cuatro lanzamientos, los de `specs/034-h41-las-skills-del/contracts/arnes.md`
§1: uno por familia, salvo `normativa`, que lleva dos porque `normativa-remisiones-en-orden` solo se mide en modo
herramienta. Hay diecinueve filas y 760 sesiones, y todas terminaron: ninguna sale «sin terminar» en las tablas y
ningún `sesion.err` tiene contenido. Cada sesión tiene instalada solo su variante. `<skill>-producto` es
`skills/<skill>` de esa cabeza; en `plazos-y-territorio` es byte a byte `plazos-y-territorio-control` (FR-010), y su
fila es la de control contra control.

La versión de Claude Code no es la del sondeo de la #245 (2.1.317, `evidencias/sondeos-20261106/`). Por eso cada
comparación de la aceptación se hace dentro de este sondeo, `control` frente a sus variantes.

## La aceptación de la fusión (FR-052)

Intervalos de Wilson al 95 %. Sesiones de 40 que abren cada referencia a cuya condición llegan las evals:

| Variante | Modo | Activan antes | `con_herramientas.md` | `normas.md` | `redaccion_en_una_fecha.md` |
|---|---|---|---|---|---|
| `normativa-control` | orden | 40 | 14 [22,1–50,5 %] | 40 | 10 |
| `normativa-producto` | orden | 40 | **40** [91,2–100,0 %] | 40 | 10 |
| `normativa-control` | herramienta | 40 | 1 [0,4–12,9 %] | 5 [5,5–26,1 %] | 4 [4,0–23,1 %] |
| `normativa-producto` | herramienta | 40 | **34** [70,9–92,9 %] | **35** [73,9–94,5 %] | 10 [14,2–40,2 %] |
| `normativa-remisiones-en-orden` | herramienta | 40 | 35 [73,9–94,5 %] | 37 [80,1–97,4 %] | 10 |

| Variante | Modo | Activan antes | `consulta_preparada.md` | `documento_aportado.md` |
|---|---|---|---|---|
| `jurisprudencia-control` | orden | 40 | 30 | 20 |
| `jurisprudencia-producto` | orden | 40 | 30 | 20 |
| `jurisprudencia-control` | herramienta | 40 | 29 | 20 |
| `jurisprudencia-producto` | herramienta | 40 | 29 | 20 |

- **En modo herramienta, el intervalo de `normativa-producto` en `con_herramientas.md` y en `normas.md` no se toca con el
  de `normativa-control` y sí con el de `normativa-remisiones-en-orden`**: se cumple.
- **En `normativa` y `jurisprudencia`, en cada modo, `producto` no queda por debajo de `control`**, ni en «Activan
  antes» ni en ninguna referencia a cuya condición llegan las evals: se cumple. En `normativa` queda por encima en las
  dos referencias que siempre hacen falta y en `redaccion_en_una_fecha.md`, y en `jurisprudencia` sale igual celda a
  celda.

La propuesta cumple la aceptación de la fusión.

En modo herramienta, `normativa-producto` abre `con_herramientas.md` y `normas.md` en 34 y 35 de 40, frente a 38 y 38
en el sondeo de la #245 con el mismo texto (salvo el «o» de la orden de `redaccion_modificada.md`, a cuyo caso no llega
ninguna eval). Los intervalos se tocan. Lo que falta está en las evals 04 (6 y 7 de 10) y 05 (8 y 8); en las 15 y 22
abre las dos en 10 de 10. La eval 04 es también la de la serie de `sin_activar` en modo herramienta del job (research
D21 de H41).

## Las variantes de `plazos-y-territorio` (FR-053)

No deciden la fusión: su lectura va al ADR que dice qué forma aplica H41.1.

| Variante | Modo | Activan antes | `administrativo.md` | `calculo_de_plazos.md` | `jerarquia_normativa.md` | `municipios.md` |
|---|---|---|---|---|---|---|
| `control` | orden | 40 | 30 | 25 [47,0–75,8 %] | 0 | 3 [2,6–19,9 %] |
| `producto` (= control) | orden | 39 | 29 | 25 [47,0–75,8 %] | 0 | 10 [14,2–40,2 %] |
| `sin-condicion` | orden | 39 | 29 | 29 [57,2–83,9 %] | 10 | **39** [87,1–99,6 %] |
| `condicion-por-orden` | orden | 40 | 30 | 30 [59,8–85,8 %] | 0 | 8 [10,5–34,8 %] |
| `lista-con-verbo` | orden | 38 | 28 | 27 [52,0–79,9 %] | 2 | 13 [20,1–48,0 %] |
| `control` | herramienta | 40 | 30 | 28 [54,6–81,9 %] | 0 | 11 [16,1–42,8 %] |
| `producto` (= control) | herramienta | 40 | 30 | 26 [49,5–77,9 %] | 1 | 14 [22,1–50,5 %] |
| `sin-condicion` | herramienta | 40 | 30 | 30 [59,8–85,8 %] | 3 | **40** [91,2–100,0 %] |
| `condicion-por-orden` | herramienta | 40 | 30 | 28 [54,6–81,9 %] | 1 | 8 [10,5–34,8 %] |
| `lista-con-verbo` | herramienta | 40 | 30 | 28 [54,6–81,9 %] | 0 | 11 [16,1–42,8 %] |

Con la regla del ADR 0052, punto 3 (ninguna celda por debajo de `control` con intervalos que no se tocan), ninguna de
las tres queda por debajo. Solo `sin-condicion` queda por encima, y lo hace en `municipios.md` en los dos modos (39 y
40 de 40, frente a 3 y 11), con intervalos que no se tocan.

El desglose por eval (`recuento.md`):

- **`municipios.md`.**
  - `sin-condicion` la abre en 10 de 10 en cada eval y modo, salvo una sesión de la eval 10 en modo orden que no activa
    la skill.
  - En la eval 02, la del municipio no cubierto, a la que en control llega «al presentar un territorio» (research D15),
    `control` la abre en 2 de 10 en modo orden y en 10 de 10 en modo herramienta, y `lista-con-verbo`, en 7 y 10.
  - En `condicion-por-orden` ninguna eval llega a la condición de su orden: en las cuatro, el municipio se dice y
    `territorio resolver` sale con 0 (research D15, V6). Sus 8 de 40 por modo son lecturas que ninguna orden pide, y no
    una orden que no se sigue. Lo que esa forma da en el caso de su condición queda sin medir.
  - En las evals 05, 10 y 11, `control`, `producto`, `condicion-por-orden` y `lista-con-verbo` la abren en 0 a 5 de 10.
- **`calculo_de_plazos.md`.** En las evals 05, 10 y 11, `sin-condicion` la abre en 10 de 10 en modo herramienta y en 10,
  9 y 10 en modo orden, frente a entre 8 y 10 de `control`. `condicion-por-orden`, cuya orden «si alguno de los
  municipios cuenta» llega a las tres, la abre en 10, 10 y 10 en modo orden y en 8, 10 y 10 en modo herramienta. En la eval 05, la que cayó de 10 a 3 en la #245, todas las
  formas la abren en 8 a 10 de 10. En la eval 02 no la abre ninguna variante: ninguna eval de la familia pide ahí un
  plazo.
- **`jerarquia_normativa.md`**, cuya línea no cambia en ninguna variante, la abren en 10 de 40 `sin-condicion` en modo
  orden y en 0 a 3 de 40 las demás filas. Ninguna eval llega a su condición.

**Cuánto se mueve el control.** Entre `control` y `producto`, que son el mismo texto, `municipios.md` va de 3 a 10 de
40 en modo orden y de 11 a 14 en modo herramienta, con intervalos que se tocan. Entre sondeos se mueve más: en el de la
#245, con Claude Code 2.1.317, `plazos-y-territorio-control` abría `municipios.md` en modo orden en 17 de 40
[28,5–57,8 %], y en este, en 3 [2,6–19,9 %]. Los intervalos no se tocan, y en la eval 05 pasa de 9 a 1 de 10. La lista
de control no es una base estable entre medidas. La forma sin condición no depende de ella: es la única que no baja de
39 de 40 en ningún modo.

## Qué hay en cada fichero

- **`tabla-normativa.md`, `tabla-plazos-y-territorio.md`, `tabla-jurisprudencia.md`**: la tabla de cada familia, tal
  como la imprime `make laboratorio-tabla`, con el detalle de sus sesiones (200, 400 y 160).
- **`recuento.md`**: la cuenta de `guiones/recuento.py` (`python3 -I guiones/recuento.py tabla-normativa.md
  tabla-plazos-y-territorio.md tabla-jurisprudencia.md`).
- **`salida.txt`**: lo que imprimieron los cuatro lanzamientos.
- **`guiones/recuento.py`**: el guion del recuento, copia sin cambios del de `evidencias/sondeos-20261106/`.

Las sesiones, con sus trazas, se conservan en `~/Projects/kitlegal-sondeos/h41-relanzamiento/` de quien lo lanzó y no
se versionan aquí.
