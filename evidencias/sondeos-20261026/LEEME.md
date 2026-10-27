# Evidencia del relanzamiento de H31.1: cinco sondeos de la eval 15 sobre la skill de la #200

Material de los sondeos del 2026-10-26 y el 2026-10-27, hechos en una sesión interactiva a petición de Jorge, fuera de
cualquier run (ADR 0032), después de que el run `ce64a33e` de H31.1 cerrara la #200 en verde. El primero es el que la
entrada de H31.1 dejaba a una persona antes de fusionar. Es la evidencia de la entrada corregida. Como todo
`evidencias/`, un run del workflow no la escribe.

No es evidencia de una fuente: ninguno de sus ficheros sale de una fila de `docs/SOURCES.md`. `manifiesto.json` da la
huella y el tamaño de cada fichero.

## Los sondeos

`make evals-sondeo` con el arnés de `evidencias/sondeos-20261025/parches/arnes-del-sondeo.diff`, sin cambios, sobre la
cabeza de la #200 (`47c3c24`, cerrada sin fusionar): la eval 15, 20 repeticiones por modo, 40 respuestas por brazo,
`claude-sonnet-5-5`, concurrencia 5, y sesiones y juez (`claude-opus-5-5`) con Claude Code 2.1.289, el de la medida
del juez. Ninguna sesión quedó sin medir ni sin terminar, ninguna respuesta sin juzgar, y en todos la medida
versionada del juez corresponde. Los lanzó Jorge con `guiones/lanzar.sh` (H y V, 2026-10-26) y `guiones/lanzar-2.sh`
(V2, W y X, 2026-10-27).

| Brazo | Skill | `afirma_lo_no_leido` | `cuenta_su_proceso` | Leen el art. 7 | Lo nombran | Hablan de exenciones |
|---|---|---|---|---|---|---|
| `brazo-H/` | la de la #200, tal cual | 11 de 40 | 10 de 40 | 0 | 28 | 11 |
| `brazo-V/` | la #200 con el paso 6, «Repasar antes de enviar» (`parches/brazo-V.diff`) | 8 de 40 | 0 de 40 | 0 | 11 | 12 |
| `brazo-V2/` | V otra vez, control a la hora de W y X | 6 de 40 | 4 de 40 | 0 | 9 | 16 |
| `brazo-W/` | V con un paso 3 que lee por defecto lo remitido, con su porqué y su «ojo», y el repaso de todo lo que se dice de un precepto, con número o sin él (`parches/brazo-W.diff`; la skill entera, en `brazo-W/skill/`) | 1 de 40 | 2 de 40 | 0 | 25 | 5 |
| `brazo-X/` | W con la regla 5 corta, sin sus prohibiciones concretas (`parches/brazo-X.diff`) | 6 de 40 | 0 de 40 | 2 | 38 | 13 |

El control es el de `evidencias/sondeos-20261025/`: la skill de `main` antes de H31.1, 16 de 80 en
`afirma_lo_no_leido` y 17 de 80 en `cuenta_su_proceso`. Lo que dicen, con Fisher unilateral:

- **La #200 no cumple el criterio de su entrada.** H da 11 de 40 en `afirma_lo_no_leido` y 10 de 40 en
  `cuenta_su_proceso`, frente a como mucho 2 de 40 en cada una. El cuerpo nuevo de la respuesta no basta y la regla de
  leer lo que se nombra no se cumple: ninguna sesión lee el art. 7.
- **El repaso baja `cuenta_su_proceso`.** V y V2 dan 4 de 80 (5 %), frente a 17 de 80 del control (p ≈ 0,002) y 10 de
  40 de H. Las marcas que quedan, en V2 y W, son la línea de `kitlegal graph check`: «la comprobación no detecta
  cambios».
- **W baja `afirma_lo_no_leido`.** 1 de 40, frente a 16 de 80 del control (p = 0,006), a V y V2 juntos, 14 de 80
  (p = 0,014), y a V2, a la misma hora, 6 de 40 (p = 0,054).
- **W no lo baja leyendo.** Ninguna sesión de W lee el art. 7. 25 de 40 lo nombran con las palabras del art. 17 («sin
  perjuicio de lo dispuesto en el artículo 7 de esta Ley») y solo 5 hablan de exenciones, frente a 12 a 16 en V y V2.
  La respuesta deja de pegar al número lo que el modelo sabe de memoria y avisa menos de que puede haber rentas exentas.
  Es lo que la regla 5 pide de lo no leído.
- **La regla 5 corta no sirve.** X, sin las prohibiciones concretas de la regla 5, vuelve a 6 de 40, el nivel de V
  (p = 0,62 frente a V2): 38 de 40 nombran el art. 7 y vuelve la materia pegada a su número («el art. 7 de la Ley, que
  recoge rentas exentas»).
- **La duración apenas cambia.** La mediana de las sesiones es de 19 s en H, 20 s en V y V2, 23 s en W y 24 s en X.

`guiones/analizar.py` saca todas las cifras de la tabla, salvo las de las dos clases, que dan las salidas, y las frases
marcadas, de las transcripciones y los votos de cada temporal. Las transcripciones se conservaron en el temporal de
cada sondeo y no se versionan aquí.

## Qué hay en cada carpeta

- **`brazo-<b>/`**: `salida.txt`, lo que imprimió el sondeo, y `votos.json`, la respuesta de cada sesión y el juicio
  del juez, voto a voto; en `brazo-W/skill/`, además, la skill medida, `SKILL.md` y referencias: la de `47c3c24` con
  `parches/brazo-W.diff`.
- **`parches/`**: los cambios de la skill de V, W y X sobre `47c3c24`. Son prototipos, no parte del producto.
- **`guiones/`**: los dos lanzadores, que leen la credencial del llavero una vez y lo bloquean antes de abrir ninguna
  sesión, y el análisis de los brazos.
