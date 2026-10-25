# Evidencia de H31.1: las medidas de `boe-legislacion` antes y después de H31, y seis sondeos de la eval 15

Material del análisis y de los dos sondeos del 2026-10-25, hechos en una sesión interactiva a petición de Jorge, fuera
de cualquier run (ADR 0032), después de que la #195 y la #198 salieran en rojo por `afirma_lo_no_leido`. Es la
evidencia de la entrada de H31.1. Como todo `evidencias/`, un run del workflow no la escribe; `evidencias/h31-1/` queda
para el paso `grabar_datos` del run de H31.1.

No es evidencia de una fuente: ninguno de sus ficheros sale de una fila de `docs/SOURCES.md`. `manifiesto.json` da la
huella y el tamaño de cada fichero.

## Las medidas del job

`medidas-del-job/medidas.json` resume las 42 ejecuciones del job de evals que midieron `boe-legislacion` desde la
primera con juez (H24, 2026-10-06) hasta la #198: de cada una, su commit, el árbol de la skill que midió, las versiones
de Claude Code de sus sesiones y de su juez, `afirma_lo_no_leido` con los dos modos juntos y en la eval 15,
`cuenta_su_proceso` por modo, y las respuestas marcadas con sus frases. Lo sacan de los logs del job
`guiones/recoger.sh`, `guiones/extraer.py` y `guiones/tabla.py`.

| Respuestas de `claude-sonnet-5-5` | Antes de H31 (34 medidas) | Desde H31 (8 medidas) | Fisher unilateral |
|---|---|---|---|
| `afirma_lo_no_leido`, todas | 37 de 3.836 (0,96 %) | 23 de 960 (2,40 %) | p ≈ 0,0008 |
| `afirma_lo_no_leido`, eval 15 | 23 de 192 (12 %) | 14 de 48 (29 %) | p ≈ 0,005 |
| `afirma_lo_no_leido`, las demás | 14 de 3.644 (0,38 %) | 9 de 912 (0,99 %) | p ≈ 0,03 |
| `cuenta_su_proceso`, modo orden | 34 de 1.889 (1,8 %) | 127 de 480 (26,5 %) | |
| `cuenta_su_proceso`, modo herramienta | 53 de 1.862 (2,8 %) | 163 de 480 (34,0 %) | |

«Desde H31» son los árboles `92525ec` (H31) y `b4620bb` (H31 con la #193). Todas las sesiones son de Claude Code
2.1.284 y todos los votos del juez de la 2.1.289: la versión no cambió. En la eval 15, ninguna sesión de ninguna medida
leyó el art. 7 de la Ley 35/2006; todas leyeron el art. 17, y las marcas son casi siempre la materia del art. 7 pegada
a su número: «que regula las rentas exentas», «(rentas exentas)».

## Los sondeos

`make evals-sondeo` con el arnés de `parches/arnes-del-sondeo.diff` (el de `evidencias/adr-0038/`: los dos modos, el
temporal conservado y `votos.json`), sobre `main` en `8d00d67`: la eval 15, 20 repeticiones por modo, 40 respuestas por
brazo, `claude-sonnet-5-5`, concurrencia 5, y sesiones y juez (`claude-opus-5-5`) con Claude Code 2.1.289, el de la
medida del juez. Ninguna sesión quedó sin medir ni sin terminar, ninguna respuesta sin juzgar, y en todos la medida
versionada del juez corresponde. Los lanzó Jorge con `guiones/lanzar.sh` (A a D) y `guiones/lanzar-2.sh` (E y A2).

| Brazo | Skill | `afirma_lo_no_leido` | `cuenta_su_proceso` |
|---|---|---|---|
| `brazo-A/` | la de `main` (control) | 6 de 40 | 7 de 40 |
| `brazo-B/` | la regla 5 de antes de H31, tal cual (`parches/brazo-B.diff`) | 8 de 40 | 11 de 40 |
| `brazo-C/` | la regla 5 con una forma fija y ejemplos ✓/✗ en el cuerpo (`parches/brazo-C.diff`) | 7 de 40 | 5 de 40 |
| `brazo-D/` | la regla 5 corta y la forma y los ejemplos en `references/responder.md`, que el cuerpo manda leer antes de responder (`parches/brazo-D.diff`) | 15 de 40 | 4 de 40 |
| `brazo-E/` | la skill entera de antes de H31, el árbol `27aeccc` de `d14bb86^` | 8 de 40 | 0 de 40 |
| `brazo-A2/` | la de `main` otra vez, control a la hora de E | 10 de 40 | 10 de 40 |

Lo que dicen, con Fisher unilateral:

- **Ninguna redacción de la regla 5 baja `afirma_lo_no_leido`.** B y C no se distinguen de A (p = 0,39 y 0,5); D es
  peor (15 de 40 frente a 6 de 40, p = 0,02).
- **La skill de antes de H31 marca lo mismo que la de hoy**: E, 8 de 40; A y A2, 16 de 80 (p = 0,59). Hoy cualquier
  versión da en torno al 20 % en la eval 15, por encima del 12 % de las medidas del job antes de H31 (24 de 120 frente
  a 23 de 192, p = 0,04). Lo que cambió no es el texto de la skill.
- **`cuenta_su_proceso` sí empeoró con H31**: E, 0 de 40; A y A2, 17 de 80 (p ≈ 0,0005).
- **Lo que va a una referencia como guía no se lee.** Ninguna de las 40 sesiones de D abrió `responder.md`. En los seis
  brazos, las sesiones solo abren las referencias de las que necesitan un dato para seguir: `normas` (de 23 a 40
  sesiones por brazo) y, en los brazos con la skill de H31, `con_herramientas` (de 6 a 9), que E no tiene. Lo cuenta `guiones/analizar.py` sobre las transcripciones, que
  se conservaron en el temporal de cada sondeo y no se versionan aquí.
- **Casi todas las respuestas nombran el art. 7**: de 34 a 40 de 40 por brazo. La marca llega cuando le pegan la
  materia.

## Qué hay en cada carpeta

- **`brazo-<b>/`**: `salida.txt`, lo que imprimió el sondeo, y `votos.json`, la respuesta de cada sesión y el juicio
  del juez, voto a voto.
- **`parches/`**: el arnés y los cambios de la skill de B, C y D sobre `8d00d67`. Son prototipos, no parte del
  producto. E es la carpeta `skills/boe-legislacion` de `d14bb86^`, sin parche.
- **`guiones/`**: los dos lanzadores, que leen la credencial del llavero una vez y lo bloquean antes de abrir ninguna
  sesión; el análisis de los brazos; y los tres guiones de las medidas del job.
- **`medidas-del-job/medidas.json`**: la tabla de las 42 ejecuciones.
