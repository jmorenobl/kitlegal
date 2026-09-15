# Bitácora de uso

Qué se le pidió al kit, qué falló y qué faltó. La rellena quien lo usa, en el momento, sin pulir. Es la
entrada del ritual de repriorización (`docs/ROADMAP.md` §6): cada tres o cuatro hitos se relee y el
siguiente hito sale de aquí o del backlog; ante discrepancia gana esta bitácora y el roadmap se actualiza
(ADR 0013).

Una entrada por sesión de uso, la más reciente arriba. Formato libre; basta con que quede claro qué se
quiso hacer y qué pasó. Las referencias a municipios concretos son bienvenidas aquí (es uso, no producto).

## Entradas

### 2026-09-15 · `boe-legislacion`: encontrar un artículo por su materia dentro de una norma

- **Qué se pidió.** Preguntas por materia, sin el número del artículo, en las pruebas de red de las evals
  de H5 (job `evals`, sesiones con `claude-haiku-4-5-20251001`). En la ejecución 34941499481: «¿Qué impuestos
  pueden exigir los ayuntamientos según el texto refundido de la Ley reguladora de las Haciendas Locales?»
  (eval 05; se esperaba el artículo 59, `BOE-A-2004-4214` bloque `a59`) y «¿Qué dice la Ley 40/2015 sobre el
  principio de legalidad en la potestad sancionadora?» (eval 07; se esperaba el artículo 25,
  `BOE-A-2015-10566` bloque `a25`). En la ejecución 34956596912, tres que en la anterior habían salido bien:
  «¿Qué atribuciones tiene el Pleno del ayuntamiento según la Ley reguladora de las Bases del Régimen Local?»
  (eval 03; se esperaba el artículo 22, `BOE-A-1985-5392` bloque `a22`), «¿Qué rendimientos se consideran
  rendimientos íntegros del trabajo en la ley del IRPF?» (eval 06; se esperaba el artículo 17,
  `BOE-A-2006-20764` bloque `a17`) y «¿En qué plazo hay que resolver una solicitud de acceso a la información
  pública según la Ley 19/2013?» (eval 08; se esperaba el artículo 20, `BOE-A-2013-12887` bloque `a20`).
- **Qué falló.** La skill identificó la norma y leyó su índice (código 0), pero pidió otro artículo: en la
  05, `a2` (dos veces); en la 07, `a140` a `a145` con `articulos` y después `a140`; en la 03, `a21`
  (atribuciones del Alcalde); en la 06, `a21` (rendimientos del capital); en la 08, `a12` (derecho de
  acceso), dos veces. En el job solo responden los bloques grabados, así que esas lecturas terminaron con
  código 5 (la 03 obtuvo 4 al repetirla con `--offline`, y la 06, 2 con `--timeout 10000`) y las respuestas
  dijeron que no pudieron consultar la fuente, sin texto ni cita. La 03, la 06 y la 08 habían pedido el
  artículo esperado en la ejecución anterior: el número que recuerda el modelo cambia de una sesión a otra.
- **Qué faltó.** Una herramienta que, dentro de una norma, encuentre el artículo que trata una materia. El
  índice del BOE (`boe indice`) solo da «Artículo N», sin rúbrica, y las divisiones («TÍTULO I», «CAPÍTULO
  III»…): sirve para pasar del número de un artículo a su id, no para saber qué artículo regula qué. Hoy lo
  decide lo que el modelo sabe de memoria o, con red, leer bloques a tientas.
- **Qué se hizo.** Las preguntas de la 05 y la 07 nombran el artículo desde T041, y desde T043 lo nombran
  las diez positivas de las evals de `boe-legislacion` (H5), para que midan el protocolo de la skill y no la
  memoria del modelo. La herramienta queda en el backlog, como candidata del grupo «Profundidad del BOE y de
  los escritos» (`docs/ROADMAP.md` §4).
