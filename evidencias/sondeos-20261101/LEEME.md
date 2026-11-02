# Evidencia de los sondeos de la persona antes de fusionar H32: normativa y plazos-y-territorio sobre la #214

Material de los dos sondeos que la entrada de H32 («Aceptación») y el ADR 0047 dejan a una persona antes de fusionar,
hechos en sesiones interactivas entre el 2026-10-30 y el 2026-11-02, fuera de cualquier run (ADR 0032). Como todo
`evidencias/`, un run del workflow no la escribe.

No es evidencia de una fuente: ninguno de sus ficheros sale de una fila de `docs/SOURCES.md`. `manifiesto.json` da la
huella y el tamaño de cada fichero.

## Los sondeos

`make evals-sondeo` con el arnés de `evidencias/sondeos-20261025/parches/arnes-del-sondeo.diff`, sin cambios:
`claude-sonnet-5-5`, 40 sesiones por modo en cada eval de `normativa` y 20 en cada una de `plazos-y-territorio`,
concurrencia 5, y sesiones y juez (`claude-opus-5-5`) con Claude Code 2.1.289, el de la medida del juez. Ninguna sesión
quedó sin medir ni sin terminar, ninguna respuesta sin juzgar, y en los dos la medida versionada del juez corresponde.
Los lanzó una persona con `guiones/lanzar-280.sh`, `guiones/lanzar-160-854b60a.sh` y `guiones/lanzar-160-274a5ca.sh`.

`salida.txt` da el total de cada sondeo mezclando las evals, y los límites son por eval: las cifras de abajo salen de
`votos.json` con `guiones/analizar.py`.

### `normativa`: evals 05 y 15

| Brazo | Skill | Eval 15, `afirma_lo_no_leido` (límite 6) | Eval 15, `cuenta_su_proceso` (límite 8) | Eval 05, `cuenta_su_proceso` (límite 6) | ¿Pasa? |
|---|---|---|---|---|---|
| `brazo-ddfb799/` | v0.4.2 (`ddfb799`) | 9 de 80 (p = 0,095) | 2 de 80 (p = 0,0002) | **7 de 80** | no |
| `brazo-854b60a/` | v0.4.4 (`854b60a`) | **2 de 80** (p = 0,0003) | **0 de 80** (p < 0,00001) | **1 de 80** | sí |
| `brazo-274a5ca/` | v0.4.5 (`274a5ca`), la final | **2 de 80** (p = 0,0003) | **2 de 80** (p = 0,0002) | **4 de 80** | sí |

El control es el de `evidencias/sondeos-20261025/`: 16 de 80 en `afirma_lo_no_leido` y 17 de 80 en
`cuenta_su_proceso`, con Fisher unilateral. Por modo, orden y herramienta:

| Brazo | Eval 15, `afirma_lo_no_leido` | Eval 15, `cuenta_su_proceso` | Eval 05, `cuenta_su_proceso` |
|---|---|---|---|
| `brazo-ddfb799/` | 4 y 5 de 40 | 0 y 2 de 40 | 2 y 5 de 40 |
| `brazo-854b60a/` | 0 y 2 de 40 | 0 y 0 de 40 | 0 y 1 de 40 |
| `brazo-274a5ca/` | 2 y 0 de 40 | 2 y 0 de 40 | 0 y 4 de 40 |

Lo que dicen:

- **El reparto de v0.4 había quitado frases medidas.** Seis de las nueve marcas de `ddfb799` pegaban una materia a una
  remisión al avisar de lo no leído, y tres daban la regla del rendimiento neto. Las siete de la eval 05 eran la línea
  de la comprobación: «no consta ninguna redacción posterior».
- **Devueltas las frases, la skill cumple todos los límites**, y además por debajo de lo medido en
  `evidencias/sondeos-20261028/` (`brazo-P80/`, 2 y 4 de 80).
- **v0.4.5 (`274a5ca`), la skill que se fusiona, cumple también.** Después de `854b60a`, el job de `be86054` dio
  `⚠ REDACCIÓN DE HOY:` sin `⚠ REDACCIÓN A FECHA:` en la eval 22 (research D29). v0.4.5 dice en el cuerpo qué líneas
  van juntas, y por eso se repitió el sondeo. Marcas: en la eval 15, dos de una materia pegada a una remisión al avisar
  de lo no leído («qué prestaciones o becas quedan exentas»), y dos de la línea de la comprobación o los avisos. En la
  eval 05, cuatro de la línea de la comprobación («Una comprobación posterior no encontró otra redacción más reciente.»),
  todas en modo herramienta.
- **Lo que quedaba en `854b60a`**, todo en modo herramienta. Eval 15: «Si necesitas saber si una prestación concreta
  (pensión, beca, etc.) está exenta, hay que leerlos.» y «Los rendimientos netos se calculan después, aplicando los
  gastos deducibles.»; eval 05: «No consta una redacción posterior que sustituya a la consultada.».

### `plazos-y-territorio`: evals 07, 08 y 12

`brazo-ddfb799-plazos/`, 20 sesiones por modo sobre v2.0.1: ninguna de las 120 sin la línea `⚠ NO CUBIERTO:` con la
norma esperada y la cita del precepto, frente a un límite de 1. La skill no tiene juez, y la tasa de `salida.txt` es la
cifra. `normativa` cambió después, pero esta skill no, así que el sondeo sigue valiendo y no se repitió.

## Qué hay en cada carpeta

- **`brazo-<commit>/`**: `salida.txt`, lo que imprimió el sondeo, y `votos.json`, la respuesta de cada sesión y el juicio
  del juez, voto a voto. `brazo-ddfb799-plazos/` solo trae `salida.txt`: esa skill no tiene juez.
- **`guiones/`**: los dos lanzadores, que leen la credencial del llavero una vez y lo bloquean antes de abrir ninguna
  sesión, y `analizar.py`. Cada lanzador corre sobre un clon de la cabeza de la #214 que nombra, con el arnés aplicado.

Las transcripciones se conservaron en el temporal de cada sondeo y no se versionan aquí.
