# Bitácora de uso

Qué se le pidió al kit, qué falló y qué faltó. La rellena quien lo usa, en el momento, sin pulir. Es la
entrada del ritual de repriorización (`docs/ROADMAP.md` §6): cada tres o cuatro hitos se relee y el
siguiente hito sale de aquí o del backlog; ante discrepancia gana esta bitácora y el roadmap se actualiza
(ADR 0013).

Una entrada por sesión de uso, la más reciente arriba. Formato libre; basta con que quede claro qué se
quiso hacer y qué pasó. Las referencias a municipios concretos son bienvenidas aquí (es uso, no producto).

## Entradas

### 2026-09-20 · Instalarlo en otro proyecto, sin clonar, y actualizarlo con cada versión

- **Qué se pidió.** Cómo instalar kitlegal para usarlo en cualquier proyecto —en Claude Code, Codex o Antigravity— y
  actualizarlo conforme salgan versiones, con la menor fricción posible para alguien distinto de quien desarrolla.
- **Qué falló.** Nada del producto: `make install` hace lo que dice. Pero es una instalación para quien desarrolla:
  exige clonar y tener Go, la skill instalada es el árbol de trabajo del clon (cambia con la rama en la que esté) y
  `scripts/boe` es un enlace a un binario de fuera de la skill, que ningún zip ni ningún gestor de paquetes puede
  llevar. No había release (H19 estaba al cierre de la fase 3) y el repositorio es privado.
- **Qué faltó.** Un artefacto instalable sin clonar. Se descartó repartir bundles `.skill`/zip: el binario pesa 11 MB
  comprimido por plataforma, en macOS un ejecutable extraído de un zip descargado queda en cuarentena, y un bundle
  fino con un shim que descargue el binario mete un descargador propio y dos versiones (skill y binario) que pueden
  divergir.
- **Qué se hizo.** Decisión de Jorge (2026-09-20, ADR 0019): el binario se instala con el gestor de paquetes de cada
  plataforma (Homebrew en macOS y Linux, Scoop en Windows, `.deb`/`.rpm`, `install.sh`, `go install`), lleva las
  skills dentro y las instala él con `kitlegal skills install`: local por defecto en `./.agents/skills/`, `-g` en
  `~/.agents/skills/`, y los hosts como enlaces relativos (`--host claude` → `.claude/skills/`), detectados por su
  directorio de configuración cuando no se indica. Es el disparador del ADR 0013: H19 se adelanta a la fase 1, detrás
  de H6, con ese contrato. Entregado en H19 (#47) y publicado como v0.1.1 el 2026-09-27: en un directorio vacío,
  `curl … | sh` y `kitlegal skills install` bastan para que Claude Code responda al artículo 21 de la Ley 39/2015 con
  su cita.

### 2026-09-16 · Ninguna eval comprueba qué hace la skill ante una norma derogada

- **Qué se pidió.** Nada en concreto: la duda salió al explicar por qué el job de evals corre sin red. Si las
  respuestas del BOE están grabadas y la caché congelada, ¿cómo sabe uno que la ley que cita sigue viva?
- **Qué falló.** Nada. El binario hace lo que hay que hacer: `internal/source/boe/avisos.go` deriva de los metadatos
  los avisos `derogada` («⚠ NORMA DEROGADA: esta norma ha sido derogada.»), `vigencia-agotada` y
  `consolidacion-no-finalizada`, y `articulo.go` no emite ningún artículo con la vigencia sin comprobar: si los
  metadatos fallan, la invocación falla con «el bloque X se obtuvo, pero no se pudo comprobar su vigencia». En uso
  real, los metadatos caducan a los 300 s, precisamente porque son lo que dice si la norma sigue en vigor.
- **Qué faltó.** Que alguna eval lo mida. Las 17 de `boe-legislacion` leen artículos de normas vivas, así que ninguna
  comprueba que la skill **traslade el aviso a su respuesta**. `SKILL.md` dice que hay que señalarlo (FR-011, SC-004
  lo cuenta como regla presente en el texto), pero nadie comprueba que ocurra en una sesión real. Y el formato de eval
  tampoco lo permitiría hoy: solo sabe exigir comandos ejecutados y citas encontradas, no que la respuesta lleve un
  aviso.
- **Qué se hizo.** Anotarlo. Arreglarlo son tres piezas: un campo nuevo en el formato de eval (los avisos esperados),
  su comparación mecánica en `Juzgar` como la de las citas, y una eval que las use. La norma candidata ya está medio
  grabada: **`BOE-A-1992-26318`, la Ley 30/1992**, que H4 grabó para sus propios tests y cuyos metadatos dan
  `estatus_derogacion: "S"` y `vigencia_agotada: "S"` —o sea, dos avisos, `derogada` y `vigencia-agotada`—, con su
  bloque `a42` ya grabado. Falta grabar su `indice` (FR-074 lo exige de toda norma de una eval) y su `buscar`, si la
  norma entra en `data/normas.yaml`, que la regla «normas conocidas» obliga. Eso es la única pausa humana:
  `scripts/grabar-evals.sh`. No es un hito: cabe en una sesión.

- **Qué se pidió.** Cerrar H5 con las evals de `boe-legislacion` en verde. El job las ejecuta con
  `claude-haiku-4-5-20251001`, fijado por la clarificación Q5 del spec de H5 («un único modelo de gama
  económica») que viene de la tabla de controles del roadmap (`docs/ROADMAP.md` §4: «job semanal con modelo
  barato»). Al revisar las pausas del hito, Jorge preguntó si Haiku es el modelo adecuado para esto y para
  qué sirve la ejecución semanal.
- **Qué falló.** Nada del job ni de la skill, pero tres tareas de H5 existen solo para acomodar lo que el
  modelo hace de forma no fiable: T041 y T043 (las diez preguntas positivas nombran el artículo, porque el
  modelo no acierta el artículo por materia; entrada de abajo), T046 (la extracción de la cita admite la
  forma legible dentro de los corchetes, porque tras dos refuerzos del protocolo el modelo seguía metiéndola
  ahí en una sesión de cada diez). Las dos ejecuciones de cierre (35002104338 y 35023013878) dieron 10 de 10,
  pero «10 de 10 en una sola ejecución» es frágil con cualquier LLM. Y la ejecución semanal sobre `main`, con
  modelo, versión de Claude Code y respuestas del BOE fijados, mide sobre todo el azar del modelo: no hay
  cambios que comprobar, gasta suscripción y asume cada semana el riesgo del token en `/proc/<pid>/environ`
  (research D13 de H5).
- **Qué faltó.** Que el job decida con el modelo del uso real de la skill, que la medida no dependa de una
  sola tirada y que se ejecute cuando hay algo que medir.
- **Qué se hizo.** Decisión de Jorge (2026-09-15), pieza aparte tras fusionar H5 (#27), como ADR con enmienda
  de Q5, FR-070, SC-003 y research D13, y del roadmap:
  1. **Decide Sonnet 5** (`claude-sonnet-5`): si funciona con Sonnet funciona con Opus, y gasta menos
     suscripción que Opus. La elección es de calidad, no de coste por token (la suscripción no se factura por
     llamada).
  2. **Haiku 4.5 queda informativo**: se ejecuta y se publica en el informe como límite inferior, pero no
     hace fallar el job.
  3. **Cada eval se repite** N veces (p. ej. 3) con umbral (≥ 2 de 3) y la tasa en el informe, en lugar de
     exigir 10 de 10 en una sola ejecución.
  4. **Vuelven las preguntas por materia**, sin nombrar el artículo, junto a las que lo nombran. La
     herramienta que las haría posibles sigue en el backlog (entrada de abajo).
  5. **La extracción tolerante de la cita (T046) se queda**: compara por identificador, que es lo que SC-009
     quiere medir.
  6. **Disparador por cambios en vez de semanal**: quitar el `schedule` (`cron '41 4 * * 1'`) de
     `.github/workflows/evals.yml` y lanzar el job en las propuestas de cambio que toquen `skills/`,
     `evals/`, `data/`, el applet `boe` o el propio job, además del lanzamiento manual y del de etiqueta.

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
