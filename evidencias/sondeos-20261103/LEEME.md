# Evidencia del sondeo de la persona antes de fusionar H39: la skill de laboratorio sobre la #222

Material del sondeo que la entrada de H39 («Aceptación») y FR-050 dejan a una persona antes de fusionar, lanzado el
2026-11-03 en una sesión interactiva, fuera de cualquier run (ADR 0032). Como todo `evidencias/`, un run del workflow
no la escribe. De él salen las cifras del ADR 0050.

No es evidencia de una fuente: ninguno de sus ficheros sale de una fila de `docs/SOURCES.md`. `manifiesto.json` da la
huella y el tamaño de cada fichero.

## El sondeo

`make evals-laboratorio` con el arnés que entrega H39, sobre la cabeza de la #222 (`a593631b`), sin cambios:
`claude-sonnet-5-5`, 40 sesiones por variante y modo (diez por cada una de las cuatro evals de `laboratorio/evals/`),
concurrencia 5 y Claude Code 2.1.317, el más reciente el 2026-11-03. Dos lanzamientos, porque el arnés combina todas las
variantes con todos los modos que se le piden: la parte `a`, seis variantes en los dos modos, y la parte `b`,
`instrucciones` y `parametro`, solo en modo herramienta. `make laboratorio-tabla` los junta. Lo lanzó una persona con
`guiones/lanzar-560.sh`, que lee la credencial del llavero una vez y lo bloquea antes de abrir ninguna sesión. Las 560
sesiones terminaron; ningún registro de error tiene contenido.

Los sondeos de partida de la #214 usaron Claude Code 2.1.289. Esta tabla no se compara con ellos fila a fila: la
variante `control` se mide en el mismo sondeo que las demás.

## La tabla

`tabla.md`, tal como la imprime `make laboratorio-tabla`, con el detalle por sesión. En resumen:

| Variante | Modo | Activan antes | Abren la referencia | Cumplen la regla testigo |
|---|---|---|---|---|
| `control` | orden · herramienta | 40 · 40 | — | 40 · 40 |
| `instrucciones` | herramienta | 40 | — | 40 |
| `entrada` | orden · herramienta | 40 · 40 (la de entrada antes: 27 · 17) | — | 40 · 40 |
| `referencia-secundaria` | orden · herramienta | 40 · 40 | 40 · 39 | 40 · **37** [80,1–97,4 %] |
| `parametro` | herramienta | 40 | — | **25** [47,0–75,8 %] |
| `referencia-paso` | orden · herramienta | 40 · 40 | 40 · 40 | 39 · 40 |
| `referencia-en-el-punto` | orden · herramienta | 40 · 40 | 40 · 40 | 40 · 40 |
| `descripcion-sin-herramientas` | orden · herramienta | 40 · 40 | — | 40 · 40 |

Cada cifra es de 40 sesiones; los intervalos son de Wilson al 95 %, y el de 40 de 40 es [91,2–100 %].

Lo que se ve en las sesiones marcadas:

- **`parametro`**: las 40 sesiones llaman a `boe_articulo`, cuyo parámetro `bloque` lleva la regla en su descripción, y
  la cargan antes con `ToolSearch`. Las 15 que no la cumplen terminan sin la línea: el agente rellena el parámetro y no
  toma su descripción como una instrucción sobre la respuesta. Por eval: 9, 5, 6 y 5 de 10.
- **`referencia-secundaria`, modo herramienta**: una sesión no abre la referencia y no cumple; dos la abren en el turno 2
  y terminan sin la línea.
- **`entrada`**: la skill de entrada se salta sobre todo en la eval 01, 0 de 10 en los dos modos, en la que el agente
  activa directamente `articulos-del-boe`. La regla se cumple igual, porque esa skill es la de control.
- **`referencia-paso`, modo orden**: una sesión abre la referencia en el turno 2 y no cumple.

## ToolSearch en el turno de Skill

`turnos.md`, de `guiones/turnos.py`: en modo herramienta, `ToolSearch` va en el mismo turno que `Skill` en 24 a 40 de 40
sesiones según la variante (control, 29). En `normativa` fue en 240 de 240 y sus referencias casi no se abrían; aquí,
con el mismo orden, las referencias se abren y la regla se cumple. Quitar las herramientas de la descripción no lo
cambia (31 de 40).

## Qué hay en cada fichero

- **`tabla.md`**: la tabla de las 14 filas y el detalle de las 560 sesiones.
- **`turnos.md`**: la cuenta de `guiones/turnos.py`.
- **`salida-a.txt`, `salida-b.txt`**: lo que imprimió cada lanzamiento.
- **`guiones/`**: el lanzador y `turnos.py`.

Las sesiones (473 MB, con sus trazas) se conservan en `~/kitlegal-sondeos/20261103-pr222/h39/` de quien lo lanzó y no se
versionan aquí.
