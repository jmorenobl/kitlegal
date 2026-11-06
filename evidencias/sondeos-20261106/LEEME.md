# Evidencia del sondeo de la persona antes de fusionar H41: las tres skills del producto sobre la #245

Material del sondeo que la entrada de H41 («Aceptación») y FR-052 dejan a una persona antes de fusionar, lanzado el
2026-11-06 en una sesión interactiva, fuera de cualquier run (ADR 0032). Como todo `evidencias/`, un run del workflow
no la escribe. De él salen las cifras del ADR 0052.

No es evidencia de una fuente: ninguno de sus ficheros sale de una fila de `docs/SOURCES.md`. `manifiesto.json` da la
huella y el tamaño de cada fichero.

## El sondeo

`make evals-laboratorio` con el arnés que entrega H41, sobre la cabeza de la #245 (`fb369637`, rama
`033-h41-las-skills-del`), sin cambios: `claude-sonnet-5-5`, 40 sesiones por variante y modo (diez por eval),
concurrencia 5 y Claude Code 2.1.317, la de los sondeos de H39 y H40. Cuatro lanzamientos, uno por familia salvo
`normativa`, que lleva dos porque `normativa-remisiones-en-orden` solo se mide en modo herramienta; son las órdenes de
`specs/033-h41-las-skills-del/quickstart.md` §8. Las 520 sesiones terminaron (ninguna «sin terminar» en las tablas) y
ningún `sesion.err` tiene contenido. Cada sesión tiene instalada solo su variante; `<skill>-producto` es
`skills/<skill>` de esa cabeza.

Las evals del sondeo son las del research de H41 (D12): en `normativa`, 04, 05, 15 y 22, las de H40; en
`plazos-y-territorio`, 02, 05, 10 y 11; en `jurisprudencia`, 01, 03, 04 y 06.

## Las tablas

`tabla-normativa.md`, `tabla-plazos-y-territorio.md` y `tabla-jurisprudencia.md`, tal como las imprime
`make laboratorio-tabla` sobre cada familia, con el detalle por sesión. En las tres, todas las sesiones de todas las
filas activan la skill antes de la primera herramienta (40 de 40), salvo una de `plazos-y-territorio-control` en modo
orden (39 de 40). En resumen, sesiones de 40 que abren cada referencia a cuya condición llegan las evals:

| Variante | Modo | Referencia | Abren |
|---|---|---|---|
| `normativa-control` | herramienta | `con_herramientas.md` · `normas.md` | 0 [0,0–8,8 %] · 9 [12,3–37,5 %] |
| `normativa-producto` | herramienta | `con_herramientas.md` · `normas.md` | **38** [83,5–98,6 %] · **38** [83,5–98,6 %] |
| `normativa-remisiones-en-orden` | herramienta | `con_herramientas.md` · `normas.md` | 34 [70,9–92,9 %] · 35 [73,9–94,5 %] |
| `normativa-control` | orden | `con_herramientas.md` · `normas.md` | 14 · 40 |
| `normativa-producto` | orden | `con_herramientas.md` · `normas.md` | 40 · 40 |
| `plazos-y-territorio-control` | orden | `administrativo.md` · `calculo_de_plazos.md` · `municipios.md` | 29 · 27 · **17** [28,5–57,8 %] |
| `plazos-y-territorio-producto` | orden | `administrativo.md` · `calculo_de_plazos.md` · `municipios.md` | 30 · 22 · **1** [0,4–12,9 %] |
| `plazos-y-territorio-control` | herramienta | `administrativo.md` · `calculo_de_plazos.md` · `municipios.md` | 30 · 26 [49,5–77,9 %] · 12 |
| `plazos-y-territorio-producto` | herramienta | `administrativo.md` · `calculo_de_plazos.md` · `municipios.md` | 30 · 15 [24,2–53,0 %] · 9 |
| `jurisprudencia-control` | orden · herramienta | `consulta_preparada.md` · `documento_aportado.md` | 30 · 20 · 27 · 20 |
| `jurisprudencia-producto` | orden · herramienta | `consulta_preparada.md` · `documento_aportado.md` | 30 · 20 · 26 · 20 |

`redaccion_en_una_fecha.md` se abre en las 10 sesiones de la eval 22 de `normativa-producto` en los dos modos (5 de 10
en `normativa-control` en modo herramienta). `como_se_cita.md` no se abre en `normativa-producto` (su orden pide avisos
o un id que termina en punto, y ninguna eval llega) y sí en 14 de 40 de `remisiones-en-orden`, que la manda leer antes
de escribir la respuesta. `jerarquia_normativa.md` no se abre en ninguna sesión: ninguna eval de la familia llega a su
condición (D12).

## La aceptación

- **En modo herramienta, `normativa-producto` abre `con_herramientas.md` y `normas.md` con un intervalo que no se toca
  con el de `normativa-control` y sí con el de `normativa-remisiones-en-orden`**: se cumple.
- **En cada familia y modo, `producto` no queda por debajo de `control`**: se cumple en `normativa` y en
  `jurisprudencia`, y no en `plazos-y-territorio`: en modo orden, `municipios.md` se abre en 1 de 40 frente a 17, con
  intervalos que no se tocan. La entrada dice qué sigue: la propuesta no se fusiona y lo medido va a un ADR que dice
  qué medir después (ADR 0052).

## El recuento por eval

`recuento.md`, de `guiones/recuento.py`, que lee solo el detalle de las tres tablas (no abre transcripciones), da cada
celda por eval y cuántas referencias abre cada sesión. Lo que se ve en `plazos-y-territorio`:

- `municipios.md` en modo orden: lo que pierde `producto` viene de las evals 05 y 11, que presentan un territorio
  sin que falte el municipio (9 y 5 de 10 en `control`, 0 y 0 en `producto`); en la 02, la del municipio no cubierto,
  3 y 1. En modo herramienta, la 02 se mantiene (8 y 9) y la 10 baja de 4 a 0.
- `calculo_de_plazos.md` en la eval 05, la de los municipios que cuentan sin plazo sin calcular ni lectura fallida: de
  10 a 3 de 10 en modo orden y de 6 a 0 en modo herramienta. En las evals 10 y 11 se mantiene (de 8–10 a 6–10).
- En modo orden, 14 sesiones de `control` abren tres referencias, en el turno 2 o 3; ninguna de `producto` abre más
  de dos. En las sesiones de la eval 05 que se han mirado, `control` lee `administrativo.md`, `calculo_de_plazos.md` y
  `municipios.md` seguidas tras activar la skill, antes de `kitlegal territorio resolver`, y `producto` lee solo
  `administrativo.md`, o esa y una más.

## Qué hay en cada fichero

- **`tabla-normativa.md`, `tabla-plazos-y-territorio.md`, `tabla-jurisprudencia.md`**: la tabla de cada familia y el
  detalle de sus sesiones (200, 160 y 160).
- **`recuento.md`**: la cuenta de `guiones/recuento.py` (`python3 -I guiones/recuento.py tabla-*.md`, en el orden de
  arriba).
- **`salida.txt`**: lo que imprimieron los cuatro lanzamientos.
- **`guiones/recuento.py`**: el guion del recuento.

Las sesiones, con sus trazas, se conservan en `~/Projects/kitlegal-sondeos/h41/` de quien lo lanzó y no se versionan
aquí.
