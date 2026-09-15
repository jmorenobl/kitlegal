# Bitácora de uso

Qué se le pidió al kit, qué falló y qué faltó. La rellena quien lo usa, en el momento, sin pulir. Es la
entrada del ritual de repriorización (`docs/ROADMAP.md` §6): cada tres o cuatro hitos se relee y el
siguiente hito sale de aquí o del backlog; ante discrepancia gana esta bitácora y el roadmap se actualiza
(ADR 0013).

Una entrada por sesión de uso, la más reciente arriba. Formato libre; basta con que quede claro qué se
quiso hacer y qué pasó. Las referencias a municipios concretos son bienvenidas aquí (es uso, no producto).

## Entradas

### 2026-09-15 · `boe-legislacion`: encontrar un artículo por su materia dentro de una norma

- **Qué se pidió.** Dos preguntas por materia, sin el número del artículo, en la prueba de red de las evals
  de H5 (job `evals`, ejecución 34941499481, sesiones con `claude-haiku-4-5-20251001`): «¿Qué impuestos
  pueden exigir los ayuntamientos según el texto refundido de la Ley reguladora de las Haciendas Locales?»
  (eval 05; se esperaba el artículo 59, `BOE-A-2004-4214` bloque `a59`) y «¿Qué dice la Ley 40/2015 sobre el
  principio de legalidad en la potestad sancionadora?» (eval 07; se esperaba el artículo 25,
  `BOE-A-2015-10566` bloque `a25`).
- **Qué falló.** La skill identificó la norma y leyó su índice (código 0), pero pidió otro artículo: en la
  05, `a2` (dos veces); en la 07, `a140` a `a145` con `articulos` y después `a140`. En el job solo responden
  los bloques grabados, así que esas lecturas terminaron con código 5 y las dos respuestas dijeron que no
  pudieron consultar la fuente, sin texto ni cita.
- **Qué faltó.** Una herramienta que, dentro de una norma, encuentre el artículo que trata una materia. El
  índice del BOE (`boe indice`) solo da «Artículo N», sin rúbrica, y las divisiones («TÍTULO I», «CAPÍTULO
  III»…): sirve para pasar del número de un artículo a su id, no para saber qué artículo regula qué. Hoy lo
  decide lo que el modelo sabe de memoria o, con red, leer bloques a tientas.
- **Qué se hizo.** Las preguntas de la 05 y la 07 nombran ahora el artículo (H5, T041), para que midan el
  protocolo de la skill y no la memoria del modelo. La herramienta queda en el backlog, como candidata del
  grupo «Profundidad del BOE y de los escritos» (`docs/ROADMAP.md` §4).
