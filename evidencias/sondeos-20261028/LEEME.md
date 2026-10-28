# Evidencia del sondeo de la persona antes de fusionar H31.1: la eval 15 sobre la skill de la #204

Material de los dos sondeos del 2026-10-28, hechos en una sesión interactiva a petición de Jorge, fuera de cualquier
run (ADR 0032), después de que el run `c0ccd6c3` de H31.1 cerrara la #204 en verde. Son el sondeo que la entrada de
H31.1 deja a una persona antes de fusionar. Como todo `evidencias/`, un run del workflow no la escribe.

No es evidencia de una fuente: ninguno de sus ficheros sale de una fila de `docs/SOURCES.md`. `manifiesto.json` da la
huella y el tamaño de cada fichero.

## Los sondeos

`make evals-sondeo` con el arnés de `evidencias/sondeos-20261025/parches/arnes-del-sondeo.diff`, sin cambios: la
eval 15 en los dos modos, `claude-sonnet-5-5`, concurrencia 5, y sesiones y juez (`claude-opus-5-5`) con Claude Code
2.1.289, el de la medida del juez. Ninguna sesión quedó sin medir ni sin terminar, ninguna respuesta sin juzgar, y en
los dos la medida versionada del juez corresponde. Los lanzó Jorge con `guiones/lanzar-40.sh` y
`guiones/lanzar-80.sh`.

| Brazo | Skill | Respuestas | `afirma_lo_no_leido` | `cuenta_su_proceso` | Leen el art. 7 | Lo nombran | Hablan de exenciones |
|---|---|---|---|---|---|---|---|
| `brazo-P40/` | la de la #204 tal como la cerró el run (`1d7944f`) | 40 | 4 de 40 | 3 de 40 | 0 | 30 | 7 |
| `brazo-P80/` | la #204 con las tres frases de `brazo-W` devueltas al cuerpo (`36fddc7`) | 80 | **2 de 80** | **4 de 80** | 0 | 57 | 11 |

El control es el de `evidencias/sondeos-20261025/`: la skill de `main` antes de H31.1, 16 de 80 en
`afirma_lo_no_leido` y 17 de 80 en `cuenta_su_proceso`. Lo que dicen, con Fisher unilateral:

- **La skill que cerró el run no cumplía el criterio de entonces.** P40 dio 4 de 40 en `afirma_lo_no_leido`, frente
  a como mucho 2 de 40. Para caber en 11 000 caracteres, el run había quitado tres frases que no estaban en los pasos
  medidos y que sostenían el paso 3: «ni para afirmar, confirmar o desmentir», «de lo que `kitlegal boe` devolvió en
  esta conversación» y «también del remitido; repásalas». Dos de sus cuatro marcas pegaban al art. 7 la materia
  («la remisión que permite que algunas de estas cantidades estén exentas»).
- **El criterio de 40 respuestas suspendía 1 de cada 3 veces a una skill en el objetivo.** El ADR 0047 lo cambia: 80
  respuestas, `afirma_lo_no_leido` como mucho 6 y `cuenta_su_proceso` como mucho 8, los dos por debajo del control.
- **La skill de la #204, con las frases devueltas, cumple el criterio del ADR 0047.** P80 da 2 de 80 en
  `afirma_lo_no_leido` (p = 0,0003 frente al control) y 4 de 80 en `cuenta_su_proceso` (p = 0,002).
- **Lo que queda.** Las dos marcas de `afirma_lo_no_leido` son una materia pegada a una remisión («sin perjuicio de
  otra regla de la misma Ley sobre rentas exentas») y la regla del rendimiento neto dicha de memoria («antes de
  tributar se restan los gastos deducibles»). Las cuatro de `cuenta_su_proceso` son la línea de la comprobación: «no
  consta ninguna redacción posterior».
- **La duración apenas cambia.** La mediana de las sesiones es de 19 s en P40 y 23 s en P80.

`evidencias/sondeos-20261026/guiones/analizar.py` saca de las transcripciones y los votos de cada temporal las
lecturas del art. 7, las respuestas que lo nombran o hablan de exenciones, la mediana y las frases marcadas; las cifras
de las dos clases salen de `salida.txt`. Las transcripciones se conservaron en el temporal de cada sondeo y no se
versionan aquí.

## Qué hay en cada carpeta

- **`brazo-<b>/`**: `salida.txt`, lo que imprimió el sondeo, y `votos.json`, la respuesta de cada sesión y el juicio
  del juez, voto a voto.
- **`guiones/`**: los dos lanzadores, que leen la credencial del llavero una vez y lo bloquean antes de abrir ninguna
  sesión. Cada uno corre sobre un worktree de la cabeza de la #204 que nombra, con el arnés aplicado.
