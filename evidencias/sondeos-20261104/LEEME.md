# Evidencia del sondeo de la persona antes de fusionar H40: las variantes de `normativa` sobre la #228

Material del sondeo que la entrada de H40 («Aceptación») y FR-040 dejan a una persona antes de fusionar, lanzado el
2026-11-04 en una sesión interactiva, fuera de cualquier run (ADR 0032). Como todo `evidencias/`, un run del workflow
no la escribe. De él salen las cifras del ADR 0051.

No es evidencia de una fuente: ninguno de sus ficheros sale de una fila de `docs/SOURCES.md`. `manifiesto.json` da la
huella y el tamaño de cada fichero.

## El sondeo

`make evals-laboratorio` con el arnés que entrega H40, sobre la cabeza de la #228 (`be1dac4a`), sin cambios:
`claude-sonnet-5-5`, 40 sesiones por variante y modo (diez por cada una de las evals 04, 05, 15 y 22 de `normativa`),
concurrencia 5 y Claude Code 2.1.317, la del sondeo de H39. Dos lanzamientos, porque el arnés combina todas las
variantes con todos los modos que se le piden: la parte `a`, `normativa-control` en los dos modos, y la parte `b`, las
otras tres variantes solo en modo herramienta. `make laboratorio-tabla` los junta. Lo lanzó una persona con
`guiones/lanzar-200.sh`, que lee la credencial del llavero una vez y lo bloquea antes de abrir ninguna sesión. Las 200
sesiones terminaron; ningún registro de error tiene contenido. Cada sesión tiene instalada solo su variante.

## La tabla

`tabla.md`, tal como la imprime `make laboratorio-tabla`, con el detalle por sesión. En resumen, sesiones que abren
cada referencia de 40 (de las otras cinco, solo se abre `redaccion_en_una_fecha.md`, y solo en la eval 22, la de la
fecha):

| Variante | Modo | `con_herramientas.md` | `normas.md` | `como_se_cita.md` | Cumplen el testigo |
|---|---|---|---|---|---|
| `normativa-control` | orden | 21 | 40 | 0 | — |
| `normativa-control` | herramienta | **3** | **7** | 0 | — |
| `normativa-remisiones-en-orden` | herramienta | **38** | **39** | 20 | — |
| `normativa-cuerpo-corto` | herramienta | 10 | 13 | 0 | — |
| `normativa-testigo-en-referencia` | herramienta | 10 | 12 | 0 | 11 |

Todas las sesiones de todas las filas activan la skill antes de la primera herramienta (40 de 40), y las que abren una
referencia la abren en el turno 2, salvo tres aperturas en el 3: `normas.md` en una sesión de `control` y
`como_se_cita.md` en dos de `remisiones-en-orden`.

## El recuento por eval

`recuento.md`, de `guiones/recuento.py`, da cada celda por eval y un recuento ampliado. El contrato de la tabla
(`specs/032-h40-por-qu-e/contracts/tabla.md` §2) cuenta una lectura con `Bash` solo si la orden lleva
`references/<fichero>`; nueve sesiones leen con `cd …/references && cat <fichero>` y no cuentan. El recuento ampliado
las suma: una sesión más, como mucho, por celda y referencia, y ninguna conclusión cambia. Con él, en
`testigo-en-referencia` las 11 sesiones que leen `con_herramientas.md` son las 11 que cumplen la regla.

Lo que se ve por eval:

- **`cuerpo-corto`**: lo que sube sobre `control` es casi todo de la eval 22, la de la fecha (8 de 10 en las dos
  referencias); en las evals 04, 05 y 15 abre `con_herramientas.md` en 2 de 30 y `normas.md` en 5, como `control`.
- **`remisiones-en-orden`**: abre `con_herramientas.md` y `normas.md` en 9 o 10 de 10 en cada eval;
  `como_se_cita.md`, que manda leer antes de escribir la respuesta, en 1, 3, 7 y 9 de 10.

## Qué hay en cada fichero

- **`tabla.md`**: la tabla de las cinco filas y el detalle de las 200 sesiones.
- **`recuento.md`**: la cuenta de `guiones/recuento.py`.
- **`salida-a.txt`, `salida-b.txt`**: lo que imprimió cada lanzamiento.
- **`guiones/`**: el lanzador y `recuento.py`.

Las sesiones, con sus trazas, se conservan en `~/kitlegal-sondeos/20261104-pr228/h40/` de quien lo lanzó y no se
versionan aquí.
