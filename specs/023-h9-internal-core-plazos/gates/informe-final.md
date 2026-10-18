# Informe del hito H9 · `internal/core/plazos` + applet `plazos` + skill `legal-core` v1

Generado por el workflow `hito` el 2026-10-18T23:24:28Z, sobre `92fc232` de `023-h9-internal-core-plazos`.
Lo escribe scripts/workflow/informe.sh sin modelo, desde los artefactos de `specs/023-h9-internal-core-plazos/`. Fusionar (squash-merge) es una decisión humana:
si algo de lo que sigue no es lo que se quería, se corrige la sección del hito en docs/ROADMAP.md y se relanza.

## 1. Estado

- **make ci local**: verde.
- **CI y evals remotos** sobre `92fc232` (medición 2): verde; es el producto de la cabeza (lo posterior solo toca `gates/`).
- **Evals por skill**: boe-legislacion aprobado, jurisprudencia aprobado, legal-core aprobado (tasas en la sección 3).
- **Umbrales del job**: boe-legislacion: 13 umbrales, los 9 que deciden se cumplen y 4 solo se publican (2 por encima de su umbral: `cuenta_su_proceso:claude-sonnet-5-5:orden`, `cuenta_su_proceso:claude-sonnet-5-5:herramienta`); jurisprudencia: 15 umbrales, los 11 que deciden se cumplen y 4 solo se publican (2 por encima de su umbral: `afirma_que_existe:claude-sonnet-5-5:orden`, `afirma_que_existe:claude-sonnet-5-5:herramienta`); legal-core: 6 umbrales, los 6 que deciden se cumplen (sección 3).
- **Revisión final**: juez A aprobado (con `claude-opus-5-5`), juez B aprobado (con `claude-fable-5-1`), 5 rondas en 2 ciclos (el primero juzga el hito; cada uno de los siguientes, lo que cambió después del último veredicto, o el hito entero si es un rejuicio).
- **Sesiones con otro modelo que el de su rol, o sin saber cuál: 23** de 85 (ADR 0033):
  - `implementar_tarea` (2026-10-18T06:09:45Z): modelo real desconocido; `workflow.yml` da `claude-opus-5-5` a `modelo_implementacion`. No se sabe con qué modelo corrió: la sesión no trae ninguna respuesta del modelo, solo 1 mensajes de Claude Code sin modelo (un error de la API o el límite de uso).
  - `implementar_tarea_escalada` (2026-10-18T11:17:12Z): modelo real desconocido; `workflow.yml` da `claude-fable-5-1` a `modelo_escalada`. No se sabe con qué modelo corrió: la sesión no trae ninguna respuesta del modelo, solo 1 mensajes de Claude Code sin modelo (un error de la API o el límite de uso).
  - `implementar_tarea_escalada` (2026-10-18T15:37:03Z): modelo real desconocido; `workflow.yml` da `claude-fable-5-1` a `modelo_escalada`. No se sabe con qué modelo corrió: la sesión no trae ninguna respuesta del modelo, solo 1 mensajes de Claude Code sin modelo (un error de la API o el límite de uso).
  - `implementar_tarea_escalada` (2026-10-18T15:52:05Z): modelo real desconocido; `workflow.yml` da `claude-fable-5-1` a `modelo_escalada`. No se sabe con qué modelo corrió: la sesión no trae ninguna respuesta del modelo, solo 1 mensajes de Claude Code sin modelo (un error de la API o el límite de uso).
  - `implementar_tarea` (2026-10-18T16:07:08Z): modelo real desconocido; `workflow.yml` da `claude-opus-5-5` a `modelo_implementacion`. No se sabe con qué modelo corrió: la sesión no trae ninguna respuesta del modelo, solo 1 mensajes de Claude Code sin modelo (un error de la API o el límite de uso).
  - `revision_juez_a` (2026-10-18T18:50:22Z): modelo real desconocido; `workflow.yml` da `claude-opus-5-5` a `modelo_revisor`. No se sabe con qué modelo corrió: la sesión no trae ninguna respuesta del modelo, solo 1 mensajes de Claude Code sin modelo (un error de la API o el límite de uso).
  - `revision_juez_b` (2026-10-18T18:50:22Z): modelo real desconocido; `workflow.yml` da `claude-fable-5-1` a `modelo_adversario`. No se sabe con qué modelo corrió: la sesión no trae ninguna respuesta del modelo, solo 1 mensajes de Claude Code sin modelo (un error de la API o el límite de uso).
  - `revision_juez_b` (2026-10-18T19:05:24Z): modelo real desconocido; `workflow.yml` da `claude-fable-5-1` a `modelo_adversario`. No se sabe con qué modelo corrió: la sesión no trae ninguna respuesta del modelo, solo 1 mensajes de Claude Code sin modelo (un error de la API o el límite de uso).
  - `revision_juez_a` (2026-10-18T19:05:24Z): modelo real desconocido; `workflow.yml` da `claude-opus-5-5` a `modelo_revisor`. No se sabe con qué modelo corrió: la sesión no trae ninguna respuesta del modelo, solo 1 mensajes de Claude Code sin modelo (un error de la API o el límite de uso).
  - `revision_juez_b` (2026-10-18T19:20:27Z): modelo real desconocido; `workflow.yml` da `claude-fable-5-1` a `modelo_adversario`. No se sabe con qué modelo corrió: la sesión no trae ninguna respuesta del modelo, solo 1 mensajes de Claude Code sin modelo (un error de la API o el límite de uso).
  - `revision_juez_a` (2026-10-18T19:20:27Z): modelo real desconocido; `workflow.yml` da `claude-opus-5-5` a `modelo_revisor`. No se sabe con qué modelo corrió: la sesión no trae ninguna respuesta del modelo, solo 1 mensajes de Claude Code sin modelo (un error de la API o el límite de uso).
  - `revision_juez_a` (2026-10-18T19:35:30Z): modelo real desconocido; `workflow.yml` da `claude-opus-5-5` a `modelo_revisor`. No se sabe con qué modelo corrió: la sesión no trae ninguna respuesta del modelo, solo 1 mensajes de Claude Code sin modelo (un error de la API o el límite de uso).
  - `revision_juez_b` (2026-10-18T19:35:30Z): modelo real desconocido; `workflow.yml` da `claude-fable-5-1` a `modelo_adversario`. No se sabe con qué modelo corrió: la sesión no trae ninguna respuesta del modelo, solo 1 mensajes de Claude Code sin modelo (un error de la API o el límite de uso).
  - `revision_juez_a` (2026-10-18T19:50:33Z): modelo real desconocido; `workflow.yml` da `claude-opus-5-5` a `modelo_revisor`. No se sabe con qué modelo corrió: la sesión no trae ninguna respuesta del modelo, solo 1 mensajes de Claude Code sin modelo (un error de la API o el límite de uso).
  - `revision_juez_b` (2026-10-18T19:50:33Z): modelo real desconocido; `workflow.yml` da `claude-fable-5-1` a `modelo_adversario`. No se sabe con qué modelo corrió: la sesión no trae ninguna respuesta del modelo, solo 1 mensajes de Claude Code sin modelo (un error de la API o el límite de uso).
  - `revision_juez_a` (2026-10-18T20:05:36Z): modelo real desconocido; `workflow.yml` da `claude-opus-5-5` a `modelo_revisor`. No se sabe con qué modelo corrió: la sesión no trae ninguna respuesta del modelo, solo 1 mensajes de Claude Code sin modelo (un error de la API o el límite de uso).
  - `revision_juez_b` (2026-10-18T20:05:36Z): modelo real desconocido; `workflow.yml` da `claude-fable-5-1` a `modelo_adversario`. No se sabe con qué modelo corrió: la sesión no trae ninguna respuesta del modelo, solo 1 mensajes de Claude Code sin modelo (un error de la API o el límite de uso).
  - `revision_juez_b` (2026-10-18T20:37:57Z): modelo real desconocido; `workflow.yml` da `claude-fable-5-1` a `modelo_adversario`. No se sabe con qué modelo corrió: la sesión no trae ninguna respuesta del modelo, solo 1 mensajes de Claude Code sin modelo (un error de la API o el límite de uso).
  - `revision_juez_a` (2026-10-18T20:37:57Z): modelo real desconocido; `workflow.yml` da `claude-opus-5-5` a `modelo_revisor`. No se sabe con qué modelo corrió: la sesión no trae ninguna respuesta del modelo, solo 1 mensajes de Claude Code sin modelo (un error de la API o el límite de uso).
  - `revision_juez_a` (2026-10-18T20:53:01Z): modelo real desconocido; `workflow.yml` da `claude-opus-5-5` a `modelo_revisor`. No se sabe con qué modelo corrió: la sesión no trae ninguna respuesta del modelo, solo 1 mensajes de Claude Code sin modelo (un error de la API o el límite de uso).
  - `revision_juez_b` (2026-10-18T20:53:01Z): modelo real desconocido; `workflow.yml` da `claude-fable-5-1` a `modelo_adversario`. No se sabe con qué modelo corrió: la sesión no trae ninguna respuesta del modelo, solo 1 mensajes de Claude Code sin modelo (un error de la API o el límite de uso).
  - `revision_juez_a` (2026-10-18T21:08:04Z): modelo real desconocido; `workflow.yml` da `claude-opus-5-5` a `modelo_revisor`. No se sabe con qué modelo corrió: la sesión no trae ninguna respuesta del modelo, solo 1 mensajes de Claude Code sin modelo (un error de la API o el límite de uso).
  - `revision_juez_b` (2026-10-18T21:08:04Z): modelo real desconocido; `workflow.yml` da `claude-fable-5-1` a `modelo_adversario`. No se sabe con qué modelo corrió: la sesión no trae ninguna respuesta del modelo, solo 1 mensajes de Claude Code sin modelo (un error de la API o el límite de uso).
- **Cambios que ningún juez vio**: ninguno.
- **Tareas**: 15 hechas, 1 en cuarentena, 0 pendientes sin cuarentena.
- **Diff**: 31 commits; 237 files changed, 60904 insertions(+), 502 deletions(-).

## 2. Supuestos y pendientes

Decisiones que el run tomó sin preguntar, ordenadas por impacto: cada paso que escribe una etiqueta la suya (ADR 0028).

### Cambian el comportamiento visible, el alcance o una skill

**Comportamiento (salida, códigos, ficheros, argumentos)**

- corrector_spec: el hito cuenta en días naturales «solo cuando quien llama lo dice» y deja a quien llama el plazo máximo del silencio, y el spec limitaba los naturales al tipo genérico → FR-004 y FR-009 admiten el silencio en meses, en días hábiles o en días naturales, y FR-012 sale con 2 si su duración no trae unidad; es el cómputo del hito, no alcance añadido.
- corrector_spec: FR-012 no fijaba el código con varios defectos en una llamada → primero tipo, fecha, duración y número de municipios (2); luego los municipios (2 con candidatos de cada uno, junto al municipio pedido); 3 solo si todo lo demás es válido; es la lectura que no esconde un error de argumentos tras un «no encontrado».
- plan: seis nombres del BOCM no son los del INE («Braojos de la Sierra», «Buitrago de Lozoya», «Talamanca del Jarama»…) y seis líneas tienen erratas de forma (sin dos puntos, sin «de» ante el mes), y con el pliegue de `territorio` y una gramática estricta FR-084 daría diferencias de más → el cruce usa la forma compacta (sin artículos ni preposiciones ni blancos) y, si no casa, el nombre de un único municipio seguido de «de»; la gramática tolera esas erratas; ninguna lista de alias (research D13, … (entera en `specs/023-h9-internal-core-plazos/gates/supuestos.md` o `clarify-respuestas.json`)
- plan: FR-009 deja al plan cómo se llega del vencimiento al día del silencio → el día que se da es el siguiente al vencimiento del plazo máximo (arts. 24.1 y 24.4 LPAC); se descarta dar el propio vencimiento, en el que la Administración aún puede notificar (research D8).
- plan: FR-010 deja al plan el descuento de agosto en un plazo de meses → si el tramo de fecha a fecha toca agosto, el vencimiento pasa al mismo día del mes siguiente, y una notificación de agosto cuenta desde el 31 de agosto; el último día solo lo mueven los inhábiles del art. 182.1 LOPJ, también tras una prórroga que cae en agosto: la fecha más temprana (research D9).
- plan: FR-012 no dice qué pasa si se da `--unidad` a un tipo que fija su duración → 2 (`argumentos`), la misma regla que su duración; aceptarla sin efecto escondería un error de quien llama (research D10).
- corrector_plan: (proporcionalidad) `data` llevaba `inicio`, `duracion` y `computo`, que ningún requisito pide → `data` lleva solo lo que enumera FR-001 (`fecha`, `tipo`, `preceptos`, `municipios`, `festivos`, `avisos`, `cobertura`); el sobre del caso (a) pasa de 854 a 772 bytes y el del (b), de 940 a 858 (research D5, M1, M2).
- T001: contracts/applet-plazos.md §3 fija `fuente` `kitlegal.plazos` y `url` `kitlegal:applet/plazos` sin distinguir el sobre de éxito del de fallo, y §6 no repite la procedencia de los errores → los guiones la afirman también en los sobres de error de §6 (órdenes y herramienta), como `territorio-matriz` y `h23-cita-herramientas` en los suyos: research D2 da los argumentos como cadenas que decide el dominio, «el patrón de `territorio` y `cita`», y no el kernel, que firmaría `kitlegal.cli`.
- T004: contracts/applet-plazos.md §5 escribe «los de <años>» y da `<años>` como «de 2028 a 2030» o «anteriores a 2025», que dejarían «los de de 2028 a 2030» y «los de anteriores a 2025» → la preposición va con los años: «de 2027», «de 2028 a 2030» y, antes del primer año cargado del ámbito, «anteriores a 2025»; con un solo tramo de años es el texto de §5, que es lo que miran los guiones. Los tramos de un mismo aviso (un año no comunicado y el siguiente sin cargar) van cada uno con su `<motivo>`, … (entera en `specs/023-h9-internal-core-plazos/gates/supuestos.md` o `clarify-respuestas.json`)
- T004: §5 da un aviso `festivo-de-una-parte` por cada festivo de una parte del tramo y su texto dice que el día «lo ha contado como hábil» → solo se da si el cómputo contó ese día como hábil: uno en sábado, en domingo o en un día inhábil por otra causa no avisa, porque el aviso diría algo falso. Uno por festivo y municipio de esa comunidad, por fecha, municipio en el orden dado y parte.
- T004: §4 ordena los avisos de un mismo código «de fecha y de municipio», y uno de festivos locales puede tener varios años → por el primer año que le falta y, con el mismo, en el orden dado de los municipios.
- T004: research D28 no fija una duración máxima y §4 da `fecha` como AAAA-MM-DD → un plazo que acabaría después del 9999-12-31, por la duración, por el inicio o por la prórroga, sale con 2 (`argumentos`) y un mensaje que nombra el tipo, el inicio y la duración, en lugar de una fecha de cinco cifras, un desbordamiento de la aritmética o un recorrido sin fin; `ComprobarPeticion` no lo mira, porque la prórroga depende de los festivos de los municipios, así que con un municipio que no existe esa peti… (entera en `specs/023-h9-internal-core-plazos/gates/supuestos.md` o `clarify-respuestas.json`)
- T004: research D2 da la duración y la unidad como cadenas → una cadena vacía es una bandera que no se da.
- T005: research S4 recordaba el art. 182.1 LOPJ con «los días 24 y 31 de diciembre» entre los inhábiles, y el bloque grabado `acientoochentaydos`, en la redacción de la LO 14/2022, no los enumera («Son inhábiles a efectos procesales los sábados y domingos, los días de fiesta nacional y los festivos a efectos laborales en la respectiva Comunidad Autónoma o localidad»); los días del 24 de diciembre al 6 de enero los declara inhábiles el art. 183 → como manda S4, el cómputo del contencioso ya no tra… (entera en `specs/023-h9-internal-core-plazos/gates/supuestos.md` o `clarify-respuestas.json`)
- T007: research S1 daba «***» o «**» por nacional y «*» por autonómica, y lo grabado dice lo contrario en las tres resoluciones: su leyenda («Códigos de las fiestas:») da «*» a la fiesta nacional no sustituible, «**» a la nacional respecto de la que no se ha ejercido la sustitución y «***» a la de comunidad autónoma, y el `title` de cada `<abbr>` de 2025 y 2026 dice lo mismo → manda lo grabado: `nacional` con «*» o «**» y `autonomico` con «***»; y `LeerResolucion` exige que la leyenda de cada res… (entera en `specs/023-h9-internal-core-plazos/gates/supuestos.md` o `clarify-respuestas.json`)
- T007: research S2 no se cumple con lo grabado: cinco nombres de la cabecera no casan por la forma compacta con los de data/territorio/comunidades/ —«Asturias», «Com. Madrid» (2025, 2026) y «Com. de Madrid» (2027), «C. Foral Navarra», «Ciudad de Ceuta» y «Ciudad de Melilla»—, y con la forma compacta sola la derivación falla, el fichero no sale y FR-020 queda sin cumplir; una lista de alias está rechazada (principio IX; plan.md, principio V) y cruzar por la posición de la columna sería una lista d… (entera en `specs/023-h9-internal-core-plazos/gates/supuestos.md` o `clarify-respuestas.json`)
- T009: contracts/applet-plazos.md §6 fija el mensaje del 2 de un municipio («municipio «<pedido>»: » y el mensaje de `territorio`, separados por « | ») y no el del 3 → el 3 lleva la misma forma con los municipios que no están en la relación, y solo ellos: es el mensaje de `territorio` de cada uno, que nombra la entrada como pide §6, y un 3 no nombra los que sí se resolvieron.
- corrector_revision: [i][f] el supuesto T007 dejaba fuera la nota de Arán y aplicaba a toda Cataluña el 26 de diciembre de 2025 y 2026 y el 29 de marzo de 2027, que la resolución no hace inhábiles en el territorio de Arán: `plazos calcular reposicion 2025-11-26 "Vielha e Mijaran" "Vielha e Mijaran"` daba el 2025-12-29 sin aviso, posterior al vencimiento real, contra FR-011, FR-025 y el ADR 0039, punto 3 → lectura conservadora: ningún día que la resolución no hace inhábil en toda una comunidad cue… (entera en `specs/023-h9-internal-core-plazos/gates/supuestos.md` o `clarify-respuestas.json`)
- clarify Q1 (criterio d, conservadora): De ninguna fuente. Para la derivación de un año, una línea del BOCM cuyo nombre no casa con ninguno de los municipios de la Comunidad de Madrid de `data/territorio/` (Cerceda, Mataelpino, El Espartal) queda fuera de cobertura, con un motivo de falta propio del vocabulario cerrado que fija el plan. No se atribuye a ningún municipio y ninguna de sus fechas entra en los festivos de un municipio. Así ninguna línea de lo grabado se pierde sin rastro. El cruce se… (entera en `specs/023-h9-internal-core-plazos/gates/supuestos.md` o `clarify-respuestas.json`)

**Alcance (lo que queda fuera o dentro del hito)**

- plan: FR-031 da norma solo a las entradas cuyo precepto aplicable cite el plan de lo grabado → llevan norma la aplicación de los tributos (LGT, art. 97) y los sancionadores de tráfico (texto refundido, art. 1); las otras siete, el precepto de la LPAC (research D17).
- T002: data-model §2 fija la disposición del manifiesto de cada fuente de festivos como `{id, anio, url}`, y la tarea declara en cada paquete nuevo solo `doc.go`, `grabacion_test.go` y el manifiesto, «y nada más» → el manifiesto es `{"fuente", "disposiciones": [{id, anio, url}]}` y su lector vive en `grabacion_test.go` (etiqueta `grabacion`): lee con `encoding/json/v2` y `RejectUnknownMembers`, exige la fuente del paquete, un identificador con su forma (`boe.ValidarNorma`; `BOCM-AAAAMMDD-N` con f… (entera en `specs/023-h9-internal-core-plazos/gates/supuestos.md` o `clarify-respuestas.json`)
- T007 (corregido en la revisión final; ver «corrector_revision» al final): las notas aclaratorias se leen cita a cita, y una cita que no limita una fiesta a una parte de la comunidad queda fuera sola: en 2025, la segunda cita de la nota 2 (una fiesta recuperable a escoger) y la nota 3 (el 24 de junio, recuperable, en la Comunitat Valenciana). Entran la nota 1 de cada año, la de Canarias, con siete fiestas de una parte por año (FR-025), y la nota 2 de cada año, la de Cataluña, que sustituye en el … (entera en `specs/023-h9-internal-core-plazos/gates/supuestos.md` o `clarify-respuestas.json`)
- barrido: FR-063 (no se tocan `docs/HISTORIAL.md` ni `CLAUDE.md`) → el diff del hito cambia una línea de cada uno («2.9.0» → «2.9.1»), con el commit `f266c12` del workflow 2.9.1 con el que se reanudó el run, ajeno a las tareas (cierre.md §1). El barrido no lo revierte.

**Skill (lo que pide, dice o comprueba una skill)**

- corrector_spec: la declaración de no cubierto debía nombrar siempre «la norma que lo rige», y ni la DA 1ª.2 ni el art. 112.4 LPAC la nombran → lectura conservadora: cada entrada de FR-031 lleva norma solo si el plan la fija citando de lo grabado el precepto que la hace aplicable (al menos LGT y texto refundido de Tráfico); sin ella, la declaración nombra el precepto de la LPAC que excluye el procedimiento y ninguna norma que no venga de la lista (FR-042, FR-048), para que ninguna respuesta nombr… (entera en `specs/023-h9-internal-core-plazos/gates/supuestos.md` o `clarify-respuestas.json`)
- corrector_spec: (uso) «hasta qué disposición llegan los datos» se repite igual en cada llamada hasta que se repite la tarea `[datos]` → la skill lo dice solo junto a un aviso de cobertura de festivos, del ámbito de ese aviso, y no lo copia en lo demás (FR-041 y tabla de «Uso»); se quitan del cálculo de avisos los de entidad menor, que FR-024 prohíbe.
- plan: FR-040 pone el territorio antes que la cobertura, y SC-003 pide responder a una pregunta sin municipio con la declaración → la skill comprueba la cobertura antes de preguntar los municipios, que solo necesita si va a pedir un plazo; en todo asunto con plazo el orden de FR-040 se cumple (contracts/skill-legal-core.md §2).
- corrector_plan: (uso) el borrador preguntaba el municipio antes de mirar la cobertura por materia, y la pregunta de SC-003 sin municipio no recibía la declaración → FR-040 pone la cobertura primero, antes de preguntar municipios o fecha; una pregunta de plazos o de recursos empieza por el paso 7, que declara lo regulado aparte o el acto presunto sin preguntar nada, y solo el paso 8 resuelve los municipios de un plazo que se calcula con los pasos 1 a 4. Sustituye la lectura de FR-040 del supuesto… (entera en `specs/023-h9-internal-core-plazos/gates/supuestos.md` o `clarify-respuestas.json`)
- T013: contracts/skill-legal-core.md §4 da «Precepto» y «Lo regula aparte» como `art. 124.1 LPAC (BOE-A-2015-10565, bloque a124)`, pero `data/plazos.yaml` y `data/cobertura.yaml` nombran cada norma solo por su identificador, y la tarea fija que cada referencia sale de su YAML (con `plazos.LeerTipos` y `skills.LeerCobertura`) con una cabecera que nombra solo ese fichero → la celda es `art. 124.1 (BOE-A-2015-10565, bloque a124)`, sin la abreviatura: sacarla de `data/normas.yaml` haría depender la r… (entera en `specs/023-h9-internal-core-plazos/gates/supuestos.md` o `clarify-respuestas.json`)
- T013: el contrato no fija las celdas «Duración», «Cómputo» y «Precepto» de `recursos_y_plazos.md` más allá de sus títulos → la duración que fija un tipo, con el nombre de su unidad en singular o en plural («1 mes», «2 meses»); la del que no la fija, «la que recibe, con su unidad» (el silencio) o, en cada fila del genérico, «la que recibe, en `<unidad>`», con el valor que se pasa a `plazos calcular`; el cómputo, el valor de `data/plazos.yaml` (`lpac`, `contencioso`); y en «Precepto», todos los pr… (entera en `specs/023-h9-internal-core-plazos/gates/supuestos.md` o `clarify-respuestas.json`)
- reparar_cierre: la medición del cierre sobre 3f14e3a (job de evals 39168409746) deja `plazo_sin_fecha:claude-sonnet-5-5` en 1 de 9 (orden) y 2 de 9 (herramienta), las tres de `09-plazo-contencioso-agosto`, que pasa 1 de 3 y 0 de 3 con el umbral 2: tres sesiones ponen los dos preceptos del contencioso en unos mismos corchetes («bloques a46 y a128»; «bloque a46; …, bloque a128»), que `formaDeCita` no lee y dejan la línea ⚠ PLAZO: sin cita; una pasa Madrid a `plazos` sin resolverlo con `territorio`… (entera en `specs/023-h9-internal-core-plazos/gates/supuestos.md` o `clarify-respuestas.json`)
- clarify Q2 (criterio c): Solo por la línea del plazo en su forma fija (FR-048). En una eval de cobertura, la sesión cuenta en el umbral de la cobertura de su modo si su respuesta no lleva la declaración de no cubierto en su forma fija con la norma esperada, o si lleva la línea del plazo en su forma fija, con la fecha que sea. Una fecha suelta en el texto libre no se busca. La prohibición de FR-042 de pedir a `plazos` el plazo del común se comprueba con el mecanismo que ya tiene el formato de eva… (entera en `specs/023-h9-internal-core-plazos/gates/supuestos.md` o `clarify-respuestas.json`)

### Del propio run

- 2026-10-18T06:09:45Z · Tarea T001 en cuarentena: 3 intentos sin quedar en verde. Su trabajo está en specs/023-h9-internal-core-plazos/gates/cuarentena/T001.patch y la causa en specs/023-h9-internal-core-plazos/gates/tarea-T001.md.
- 2026-10-18T07:05:24Z · Tarea T002 en cuarentena: 3 intentos sin quedar en verde. Su trabajo está en specs/023-h9-internal-core-plazos/gates/cuarentena/T002.patch y la causa en specs/023-h9-internal-core-plazos/gates/tarea-T002.md.
- 2026-10-18T07:48:21Z · Tarea T003 en cuarentena: 3 intentos sin quedar en verde. Su trabajo está en specs/023-h9-internal-core-plazos/gates/cuarentena/T003.patch y la causa en specs/023-h9-internal-core-plazos/gates/tarea-T003.md.
- 2026-10-18T16:07:08Z · Tarea T015 en cuarentena: 3 intentos sin quedar en verde. Su trabajo está en specs/023-h9-internal-core-plazos/gates/cuarentena/T015.patch y la causa en specs/023-h9-internal-core-plazos/gates/tarea-T015.md.
- 2026-10-18T22:29:46Z · La revisión final abre el ciclo 2 para juzgar lo que cambió fuera de gates/ después del veredicto sobre 0c9101a: e3d6426 «fix(H9): cierre en la plataforma».

- Observaciones de los jueces, por debajo del umbral y sin corregir: 15 en `spec-r2.json`, 13 en `plan-r2.json`, 8 en `tasks-r2.json`, 7 en `revision-a-r5.json`, 8 en `revision-b-r5.json`.

### Internos (45)

Decisiones que no cambian nada observable (técnica, estructura, tests), por autor: T002 (1), T003 (4), T004 (3), T005 (7), T006 (1), T007 (3), T008 (4), T009 (1), T010 (3), T011 (5), T012 (3), T014 (2), barrido (3), corrector_plan (2), corrector_spec (1), plan (2). Enteras en `specs/023-h9-internal-core-plazos/gates/supuestos.md`.

## 3. Evals sobre la cabeza

**boe-legislacion**: aprobado sobre `92fc232`; decide `claude-sonnet-5-5` con 2 de 3; 21 evals, 0 nuevas, 8 informativas.

| Eval | claude-sonnet-5-5 | claude-haiku-4-5-20251001 | claude-sonnet-5-5 (herramienta) | claude-haiku-4-5-20251001 (herramienta) | Marca |
|---|---|---|---|---|---|
| `01-lpac-articulo-21.yaml` | 3/3 | 3/3 | 3/3 | 3/3 |  |
| `02-lcsp-contrato-menor.yaml` | 3/3 | 3/3 | 3/3 | 3/3 |  |
| `03-lrbrl-atribuciones-del-pleno.yaml` | 3/3 | 3/3 | 3/3 | 2/3 |  |
| `04-lgt-prescripcion.yaml` | 3/3 | 3/3 | 3/3 | 2/3 |  |
| `05-trlrhl-impuestos-municipales.yaml` | 3/3 | 3/3 | 3/3 | 3/3 |  |
| `06-irpf-rendimientos-del-trabajo.yaml` | 3/3 | 3/3 | 3/3 | 1/3 |  |
| `07-lrjsp-principio-de-legalidad.yaml` | 3/3 | 3/3 | 3/3 | 1/3 |  |
| `08-ltaibg-plazo-de-resolucion.yaml` | 3/3 | 3/3 | 3/3 | 3/3 |  |
| `09-constitucion-articulo-140.yaml` | 3/3 | 3/3 | 3/3 | 3/3 |  |
| `10-et-vacaciones.yaml` | 3/3 | 3/3 | 3/3 | 2/3 |  |
| `11-no-activa-programacion.yaml` | 3/3 | 3/3 | 3/3 | 3/3 |  |
| `12-no-activa-acuerdo-entre-amigos.yaml` | 3/3 | 3/3 | 3/3 | 3/3 |  |
| `13-lrbrl-atribuciones-por-materia.yaml` | 3/3 | — | 3/3 | — | informativa |
| `14-trlrhl-impuestos-por-materia.yaml` | 3/3 | — | 3/3 | — | informativa |
| `15-irpf-rendimientos-por-materia.yaml` | 3/3 | — | 3/3 | — | informativa |
| `16-lrjsp-legalidad-por-materia.yaml` | 3/3 | — | 3/3 | — | informativa |
| `17-ltaibg-plazo-por-materia.yaml` | 3/3 | — | 3/3 | — | informativa |
| `18-lrjpac-norma-derogada.yaml` | 3/3 | — | 3/3 | — | informativa |
| `19-lcsp-contrato-menor-redaccion-cambiada.yaml` | 3/3 | — | 3/3 | — | informativa |
| `20-lcsp-dos-bloques-redaccion-cambiada.yaml` | 3/3 | — | 3/3 | — | informativa |
| `21-sin-binario-ni-servidor.yaml` | 3/3 | 3/3 | — | — |  |

Umbrales que publica el job (ADR 0029):

| Umbral | Medida | Condición | Cumple | Hace fallar el job |
|---|---|---|---|---|
| `sin_activar:claude-sonnet-5-5:orden` | 0 de 54 (0,0 %) | ≤ 0,0 % | sí | sí |
| `afirma_lo_no_leido:claude-sonnet-5-5:orden` | 0 de 54 (0,0 %) | ≤ 2,8 % | dentro | no: solo se publica, no es un control |
| `cuenta_su_proceso:claude-sonnet-5-5:orden` | 2 de 54 (3,7 %) | ≤ 0,0 % | por encima | no: solo se publica, no es un control |
| `sin_activar:claude-sonnet-5-5:herramienta` | 0 de 54 (0,0 %) | ≤ 0,0 % | sí | sí |
| `afirma_lo_no_leido:claude-sonnet-5-5:herramienta` | 0 de 54 (0,0 %) | ≤ 2,8 % | dentro | no: solo se publica, no es un control |
| `cuenta_su_proceso:claude-sonnet-5-5:herramienta` | 2 de 54 (3,7 %) | ≤ 0,0 % | por encima | no: solo se publica, no es un control |
| `afirma_lo_no_leido:claude-sonnet-5-5` | 0 de 108 (0,0 %) | ≤ 2,8 % | sí | sí |
| `medida_del_juez:afirma_lo_no_leido:defectos_sin_marcar` | 0 de 212 (0,0 %) | ≤ 0,0 % | sí | sí |
| `medida_del_juez:afirma_lo_no_leido:correctos_marcados` | 0 de 47 (0,0 %) | ≤ 0,0 % | sí | sí |
| `duracion_de_las_sesiones:orden` | 441 | ≤ 900 | sí | sí |
| `duracion_de_las_sesiones:herramienta` | 430 | ≤ 900 | sí | sí |
| `duracion_del_juez:orden` | 109 | ≤ 900 | sí | sí |
| `duracion_del_juez:herramienta` | 112 | ≤ 900 | sí | sí |

**jurisprudencia**: aprobado sobre `92fc232`; decide `claude-sonnet-5-5` con 2 de 3; 10 evals, 0 nuevas, 0 informativas.

| Eval | claude-sonnet-5-5 | claude-haiku-4-5-20251001 | claude-sonnet-5-5 (herramienta) | claude-haiku-4-5-20251001 (herramienta) | Marca |
|---|---|---|---|---|---|
| `01-existe-con-numero-y-fecha.yaml` | 3/3 | 3/3 | 3/3 | 2/3 |  |
| `02-resumen-sin-documento.yaml` | 3/3 | 2/3 | 3/3 | 3/3 |  |
| `03-por-materia.yaml` | 3/3 | 2/3 | 3/3 | 3/3 |  |
| `04-documento-pegado.yaml` | 3/3 | 3/3 | 3/3 | 3/3 |  |
| `05-tribunal-constitucional.yaml` | 3/3 | 3/3 | 3/3 | 3/3 |  |
| `06-documento-que-no-es-el-pedido.yaml` | 3/3 | 0/3 | 3/3 | 0/3 |  |
| `07-resumen-de-una-conocida.yaml` | 3/3 | 2/3 | 3/3 | 3/3 |  |
| `08-doctrina-con-el-fallo-delante.yaml` | 3/3 | 2/3 | 3/3 | 3/3 |  |
| `09-de-que-trata-con-la-ficha-sola.yaml` | 3/3 | 2/3 | 3/3 | 2/3 |  |
| `10-doctrina-dada-por-hecha.yaml` | 3/3 | 0/3 | 3/3 | 1/3 |  |

Umbrales que publica el job (ADR 0029):

| Umbral | Medida | Condición | Cumple | Hace fallar el job |
|---|---|---|---|---|
| `sin_activar:claude-sonnet-5-5:orden` | 0 de 30 (0,0 %) | ≤ 0,0 % | sí | sí |
| `afirma_lo_no_leido:claude-sonnet-5-5:orden` | 0 de 30 (0,0 %) | ≤ 1,7 % | dentro | no: solo se publica, no es un control |
| `afirma_que_existe:claude-sonnet-5-5:orden` | 2 de 30 (6,7 %) | ≤ 0,0 % | por encima | no: solo se publica, no es un control |
| `cita_sin_documento:claude-sonnet-5-5:orden` | 0 de 30 (0,0 %) | ≤ 0,0 % | sí | sí |
| `documento_sin_cotejar:claude-sonnet-5-5:orden` | 0 de 12 (0,0 %) | ≤ 0,0 % | sí | sí |
| `sin_activar:claude-sonnet-5-5:herramienta` | 0 de 30 (0,0 %) | ≤ 0,0 % | sí | sí |
| `afirma_lo_no_leido:claude-sonnet-5-5:herramienta` | 0 de 30 (0,0 %) | ≤ 1,7 % | dentro | no: solo se publica, no es un control |
| `afirma_que_existe:claude-sonnet-5-5:herramienta` | 2 de 30 (6,7 %) | ≤ 0,0 % | por encima | no: solo se publica, no es un control |
| `cita_sin_documento:claude-sonnet-5-5:herramienta` | 0 de 30 (0,0 %) | ≤ 0,0 % | sí | sí |
| `documento_sin_cotejar:claude-sonnet-5-5:herramienta` | 0 de 12 (0,0 %) | ≤ 0,0 % | sí | sí |
| `afirma_lo_no_leido:claude-sonnet-5-5` | 0 de 60 (0,0 %) | ≤ 1,7 % | sí | sí |
| `medida_del_juez:afirma_lo_no_leido:defectos_sin_marcar` | 0 de 125 (0,0 %) | ≤ 0,0 % | sí | sí |
| `medida_del_juez:afirma_lo_no_leido:correctos_marcados` | 0 de 124 (0,0 %) | ≤ 0,0 % | sí | sí |
| `duracion_del_juez:orden` | 43 | ≤ 900 | sí | sí |
| `duracion_del_juez:herramienta` | 41 | ≤ 900 | sí | sí |

**legal-core**: aprobado sobre `92fc232`; decide `claude-sonnet-5-5` con 2 de 3; 9 evals, 5 nuevas, 0 informativas.

| Eval | claude-sonnet-5-5 | claude-haiku-4-5-20251001 | claude-sonnet-5-5 (herramienta) | claude-haiku-4-5-20251001 (herramienta) | Marca |
|---|---|---|---|---|---|
| `05-plazo-reposicion-leganes.yaml` | 3/3 | 0/3 | 3/3 | 0/3 | **nueva** |
| `06-plazo-reposicion-tordesillas.yaml` | 3/3 | 0/3 | 3/3 | 0/3 | **nueva** |
| `07-cobertura-liquidacion-tributaria.yaml` | 3/3 | 0/3 | 3/3 | 0/3 | **nueva** |
| `08-cobertura-multa-de-trafico.yaml` | 3/3 | 0/3 | 3/3 | 0/3 | **nueva** |
| `09-plazo-contencioso-agosto.yaml` | 3/3 | 0/3 | 3/3 | 0/3 | **nueva** |
| `01-territorio-municipio-cubierto.yaml` | 3/3 | 0/3 | 3/3 | 0/3 |  |
| `02-territorio-municipio-no-cubierto.yaml` | 3/3 | 0/3 | 3/3 | 0/3 |  |
| `03-no-activa-receta-de-cocina.yaml` | 3/3 | 3/3 | 3/3 | 3/3 |  |
| `04-sin-binario-ni-servidor.yaml` | 3/3 | 0/3 | — | — |  |

Umbrales que publica el job (ADR 0029):

| Umbral | Medida | Condición | Cumple | Hace fallar el job |
|---|---|---|---|---|
| `sin_activar:claude-sonnet-5-5:orden` | 0 de 21 (0,0 %) | ≤ 0,0 % | sí | sí |
| `plazo_sin_fecha:claude-sonnet-5-5:orden` | 0 de 9 (0,0 %) | ≤ 0,0 % | sí | sí |
| `cobertura_sin_declarar:claude-sonnet-5-5:orden` | 0 de 6 (0,0 %) | ≤ 0,0 % | sí | sí |
| `sin_activar:claude-sonnet-5-5:herramienta` | 0 de 21 (0,0 %) | ≤ 0,0 % | sí | sí |
| `plazo_sin_fecha:claude-sonnet-5-5:herramienta` | 0 de 9 (0,0 %) | ≤ 0,0 % | sí | sí |
| `cobertura_sin_declarar:claude-sonnet-5-5:herramienta` | 0 de 6 (0,0 %) | ≤ 0,0 % | sí | sí |

Cada celda: sesiones que pasan de las abiertas; ✗, una serie que decide y no llega al umbral. Una eval informativa publica su tasa sin decidir el veredicto (ADR 0016). La regla por serie no hace cumplir un umbral agregado sobre todas las respuestas: eso solo lo hace un umbral del job que lo hace fallar (ADR 0029).

**Serie de cierres de boe-legislacion** (ADR 0038, punto 6): respuestas de las evals que activan la skill que el juez marca con sus tres votos, por modo y por clase de pregunta.

Clave de la cabeza: skill `eed129f` (árbol de `skills/boe-legislacion`), decide `claude-sonnet-5-5`, juez `claude-opus-5-5` con Claude Code 2.1.289.

`afirma_lo_no_leido`:

| Medición | Fecha | Commit | Skill | Orden | Herramienta | Nombra | Por materia | Total | En la tasa |
|---|---|---|---|---|---|---|---|---|---|
| H24-1 | 2026-10-06 | `a5bda45` | `6d65aa3` | 1/54 | 0/54 | 1/78 | 0/30 | 1/108 | no: otra skill |
| H24-2 | 2026-10-06 | `b0a7a9c` | `75e1d04` | 0/54 | 3/54 | 0/78 | 3/30 | 3/108 | no: 1 respuesta sin juzgar |
| H24-3 | 2026-10-06 | `3107ae2` | `6063317` | 0/54 | 0/54 | 0/78 | 0/30 | 0/108 | no: otra skill |
| H23-1 | 2026-10-07 | `bf4befd` | `6063317` | 0/54 | 0/50 | 0/76 | 0/28 | 0/104 | no: otra skill |
| H25-1 | 2026-10-10 | `3f99446` | `6063317` | 0/54 | 1/54 | 0/78 | 1/30 | 1/108 | no: otra skill |
| H25-2 | 2026-10-10 | `e380f0a` | `6063317` | 1/54 | 1/54 | 0/78 | 2/30 | 2/108 | no: otra skill |
| H25-3 | 2026-10-10 | `d51f3d7` | `eed129f` | 0/54 | 2/54 | 0/78 | 2/30 | 2/108 | sí |
| H26-1 | 2026-10-13 | `b29fe6c` | `eed129f` | 0/54 | 2/54 | 1/78 | 1/30 | 2/108 | sí |
| H23.1-1 | 2026-10-16 | `60e8678` | `eed129f` | 0/54 | 3/54 | 1/78 | 2/30 | 3/108 | sí |
| H9-1 | 2026-10-18 | `3f14e3a` | `eed129f` | 0/54 | 1/54 | 0/78 | 1/30 | 1/108 | sí |
| H9-2 | 2026-10-18 | `92fc232` | `eed129f` | 0/54 | 0/54 | 0/78 | 0/30 | 0/108 | sí |

Tasa acumulada con la skill como está: 8 de 540 (1,48 %), de H25-3, H26-1, H23.1-1, H9-1 y H9-2.
Escalón de la regla 3 con 108 respuestas, el total de hoy: como mucho 5 marcadas (vale hasta una tasa del 1,96 %). El umbral de hoy, 0.028 en `evals/boe-legislacion/juez/clases.yaml`, admite 3 de 108.
**Aviso**: la tasa acumulada, 1,48 %, pasa de la que vale para el umbral de hoy (3 de 108, hasta el 0,95 %): con ella, una ejecución sale en rojo sin motivo más de 1 de cada 50 veces. Este informe no propone cambiar el umbral: lo que hay que mirar es la skill (ADR 0038, puntos 4 y 8).

**Serie de cierres de jurisprudencia** (ADR 0038, punto 6): respuestas de las evals que activan la skill que el juez marca con sus tres votos, por modo y por clase de pregunta.

Clave de la cabeza: skill `4759aa4` (árbol de `skills/jurisprudencia`), decide `claude-sonnet-5-5`, juez `claude-opus-5-5` con Claude Code 2.1.289.

`afirma_lo_no_leido`:

| Medición | Fecha | Commit | Skill | Orden | Herramienta | Nombra | Por materia | Total | En la tasa |
|---|---|---|---|---|---|---|---|---|---|
| H25-1 | 2026-10-10 | `3f99446` | `9480bbc` | 0/30 | 0/30 | 0/48 | 0/12 | 0/60 | no: otra skill |
| H25-2 | 2026-10-10 | `e380f0a` | `9480bbc` | 0/30 | 0/30 | 0/48 | 0/12 | 0/60 | no: 1 respuesta sin juzgar |
| H25-3 | 2026-10-10 | `d51f3d7` | `9480bbc` | 0/30 | 0/30 | 0/48 | 0/12 | 0/60 | no: otra skill |
| H26-1 | 2026-10-13 | `b29fe6c` | `9480bbc` | 0/30 | 0/30 | 0/48 | 0/12 | 0/60 | no: otra skill |
| H23.1-1 | 2026-10-16 | `60e8678` | `4759aa4` | 0/30 | 0/30 | 0/48 | 0/12 | 0/60 | sí |
| H9-1 | 2026-10-18 | `3f14e3a` | `4759aa4` | 0/30 | 0/30 | 0/48 | 0/12 | 0/60 | sí |
| H9-2 | 2026-10-18 | `92fc232` | `4759aa4` | 0/30 | 0/30 | 0/48 | 0/12 | 0/60 | sí |

Tasa acumulada con la skill como está: 0 de 180 (0,00 %), de H23.1-1, H9-1 y H9-2.
Insuficiente: 180 de 500 respuestas. La regla 3 del ADR 0038 pide la tasa de 500 respuestas al menos: con menos no da escalón.

## 4. Revisión que la constitución reserva a la persona

- Adaptador de fuente nuevo: `internal/source/bocm/`.
- Fixture o esquema EXISTENTE modificado: `internal/app/testdata/script/argumentos.txtar`.
- Fixture o esquema EXISTENTE modificado: `internal/app/testdata/script/h21-mcp-herramientas.txtar`.
- Fixture o esquema EXISTENTE modificado: `internal/app/testdata/script/h21-mcp-proceso.txtar`.
- Fixture o esquema EXISTENTE modificado: `internal/app/testdata/script/h21-mcp-protocolo.txtar`.
- Fixture o esquema EXISTENTE modificado: `schemas/eval.yaml.json`.
- Fixtures nuevos (27), por directorio:
  - `internal/app/testdata/plazos/`: `caso-a.json`, `caso-b.json`, `caso-e.json`, `describe.json`
  - `internal/evals/testdata/boe.legislacion-consolidada/`: `GET_https_www.boe.es_datosabiertos_api_legislacion-consolidada_id_BOE-A-1889-4763_texto_bloque_art5.json`, `GET_https_www.boe.es_datosabiertos_api_legislacion-consolidada_id_BOE-A-1985-12666_texto_bloque_acientoochentaycinco.json`, `GET_https_www.boe.es_datosabiertos_api_legislacion-consolidada_id_BOE-A-1985-12666_texto_bloque_acientoochentaydos.json`, `GET_https_www.boe.es_datosabiertos_api_legislacion-consolidada_id_BOE-A-1985-12666_texto_bloque_acientoochentaytres.json`, `GET_https_www.boe.es_datosabiertos_api_legislacion-consolidada_id_BOE-A-1998-16718_texto_bloque_a128.json`, `GET_https_www.boe.es_datosabiertos_api_legislacion-consolidada_id_BOE-A-1998-16718_texto_bloque_a46.json`, `GET_https_www.boe.es_datosabiertos_api_legislacion-consolidada_id_BOE-A-2003-23186_texto_bloque_a97.json`, `GET_https_www.boe.es_datosabiertos_api_legislacion-consolidada_id_BOE-A-2013-12887_texto_bloque_a24.json`, `GET_https_www.boe.es_datosabiertos_api_legislacion-consolidada_id_BOE-A-2015-10565_texto_bloque_a112.json`, `GET_https_www.boe.es_datosabiertos_api_legislacion-consolidada_id_BOE-A-2015-10565_texto_bloque_a122.json`, `GET_https_www.boe.es_datosabiertos_api_legislacion-consolidada_id_BOE-A-2015-10565_texto_bloque_a124.json`, `GET_https_www.boe.es_datosabiertos_api_legislacion-consolidada_id_BOE-A-2015-10565_texto_bloque_a24.json`, `GET_https_www.boe.es_datosabiertos_api_legislacion-consolidada_id_BOE-A-2015-10565_texto_bloque_a30.json`, `GET_https_www.boe.es_datosabiertos_api_legislacion-consolidada_id_BOE-A-2015-10565_texto_bloque_daprimera.json`, `GET_https_www.boe.es_datosabiertos_api_legislacion-consolidada_id_BOE-A-2015-11722_metadatos.json`, `GET_https_www.boe.es_datosabiertos_api_legislacion-consolidada_id_BOE-A-2015-11722_texto_bloque_a1.json`, `GET_https_www.boe.es_datosabiertos_api_legislacion-consolidada_id_BOE-A-2015-11722_texto_indice.json`, `GET_https_www.boe.es_datosabiertos_api_legislacion-consolidada_q_limit_10_query_7B_22query_22_3A_20_-524cb2ad.json`, `GET_https_www.boe.es_datosabiertos_api_legislacion-consolidada_q_limit_10_query_7B_22query_22_3A_20_-744167c4.json`, `GET_https_www.boe.es_robots.txt.json`
  - `internal/evals/testdata/`: `grabaciones.json`
  - `internal/source/bocm/testdata/`: `grabaciones.json`
  - `internal/source/boe/fiestas/testdata/`: `grabaciones.json`
- Esquemas nuevos: `schemas/cobertura.yaml.json`, `schemas/plazo.json`.
- Suite de aceptación activada y congelada (5 guiones, escritos desde el spec en la primera tarea): `h9-plazos-calcular.txtar`, `h9-plazos-cobertura.txtar`, `h9-plazos-codigos.txtar`, `h9-plazos-herramienta.txtar`, `h9-plazos-sin-efectos.txtar`.
- 13 ficheros de evidencia en evidencias/ (manifiesto con huellas).
- Grabaciones del paso grabar_datos: `specs/023-h9-internal-core-plazos/gates/grabaciones.md`.

## 5. Tareas en cuarentena

- **T015**: 3 intentos sin quedar en verde. Parche `specs/023-h9-internal-core-plazos/gates/cuarentena/T015.patch`; causa en `specs/023-h9-internal-core-plazos/gates/tarea-T015.md`.

## 6. Trazabilidad

**Estado**: «tareas hechas» solo dice que las tareas que citan el requisito están marcadas; nada del run lo ha medido. «comprobado por su control» exige su fila en «Controles de umbral» de `plan.md` y que el control esté: un test o una comprobación de `make ci` que existe en la cabeza, con `make ci` en verde, o un umbral que el job de evals publica, cumple y hace fallar el job (ADR 0029). «UMBRAL NO CUMPLIDO» y «CONTROL SIN VERIFICAR» son lo que hay que mirar.

| Requisito | Tareas | Aceptación | Estado | Control de umbral |
|---|---|---|---|---|
| FR-001 | T001(hecha) T009(hecha) | plazos-calcular.txtar plazos-cobertura.txtar plazos-codigos.txtar plazos-herramienta.txtar plazos-sin-efectos.txtar  | comprobado por su control | `ci:internal/app/herramientas_test.go:TestHerramientasDelServidor`, en make ci (verde); `ci:internal/app/testdata/script/h9-plazos-herramienta.txtar`, en make ci (verde) |
| FR-002 | T001(hecha) T004(hecha) T010(hecha) | plazos-calcular.txtar  | tareas hechas | — |
| FR-003 | T001(hecha) T003(hecha) T004(hecha) T010(hecha) | plazos-calcular.txtar  | tareas hechas | — |
| FR-004 | T001(hecha) T004(hecha) T010(hecha) | — | tareas hechas | — |
| FR-005 | T001(hecha) T004(hecha) T010(hecha) | plazos-calcular.txtar  | tareas hechas | — |
| FR-006 | T001(hecha) T004(hecha) T010(hecha) | plazos-calcular.txtar  | tareas hechas | — |
| FR-007 | T001(hecha) T004(hecha) T009(hecha) T010(hecha) | plazos-codigos.txtar  | tareas hechas | — |
| FR-008 | T001(hecha) T003(hecha) T004(hecha) T005(hecha) T010(hecha) | — | tareas hechas | — |
| FR-009 | T001(hecha) T004(hecha) T010(hecha) | — | tareas hechas | — |
| FR-010 | T001(hecha) T004(hecha) T005(hecha) T009(hecha) T010(hecha) | plazos-calcular.txtar plazos-codigos.txtar  | comprobado por su control | `ci:internal/evals/grabaciones_test.go:TestPreceptosDeLosPlazos`, en make ci (verde) |
| FR-011 | T001(hecha) T003(hecha) T004(hecha) T007(hecha) T010(hecha) | — | comprobado por su control | `ci:internal/app/plazos_test.go:TestFestivoDeUnaParte`, en make ci (verde) |
| FR-012 | T001(hecha) T004(hecha) T009(hecha) | plazos-codigos.txtar  | comprobado por su control | `ci:internal/app/testdata/script/h9-plazos-codigos.txtar`, en make ci (verde); `ci:internal/app/testdata/script/h9-plazos-cobertura.txtar`, en make ci (verde) |
| FR-013 | T001(hecha) T009(hecha) | plazos-calcular.txtar plazos-sin-efectos.txtar  | comprobado por su control | `ci:internal/arch_test.go:TestArquitectura`, en make ci (verde); `ci:internal/app/plazos_test.go:TestAppletPlazos`, en make ci (verde); `ci:internal/app/testdata/script/h9-plazos-sin-efectos.txtar`, en make ci (verde); `ci:internal/app/testdata/script/h9-plazos-calcular.txtar`, en make ci (verde) |
| FR-020 | T002(hecha) T003(hecha) T007(hecha) T008(hecha) | — | comprobado por su control | `ci:internal/source/boe/fiestas/derivar_test.go:TestFestivosNacionalesDerivados`, en make ci (verde) |
| FR-021 | T008(hecha) | — | comprobado por su control | `ci:internal/source/bocm/derivar_test.go:TestFestivosLocalesDerivados`, en make ci (verde) |
| FR-022 | T003(hecha) T006(hecha) T008(hecha) | — | comprobado por su control | `ci:internal/source/bocm/derivar_test.go:TestFestivosLocalesDerivados`, en make ci (verde) |
| FR-023 | T008(hecha) | — | tareas hechas | — |
| FR-024 | T004(hecha) T008(hecha) T010(hecha) | — | tareas hechas | — |
| FR-025 | T003(hecha) T007(hecha) | — | tareas hechas | — |
| FR-026 | T001(hecha) T003(hecha) T004(hecha) T009(hecha) T010(hecha) | plazos-cobertura.txtar  | tareas hechas | — |
| FR-027 | T002(hecha) | — | tareas hechas | — |
| FR-028 | T001(hecha) T009(hecha) T016(hecha) | plazos-sin-efectos.txtar  | tareas hechas | — |
| FR-030 | T002(hecha) T003(hecha) T005(hecha) | — | comprobado por su control | `ci:internal/evals/grabaciones_test.go:TestPreceptosDeLosPlazos`, en make ci (verde) |
| FR-031 | T002(hecha) T005(hecha) | — | comprobado por su control | `ci:internal/evals/grabaciones_test.go:TestPreceptosDeLosPlazos`, en make ci (verde) |
| FR-032 | T013(hecha) | — | comprobado por su control | `ci:Makefile:skills-check`, en make ci (verde) |
| FR-040 | T013(hecha) | — | tareas hechas | — |
| FR-041 | T013(hecha) | — | tareas hechas | — |
| FR-042 | T013(hecha) | — | tareas hechas | — |
| FR-043 | T013(hecha) | — | tareas hechas | — |
| FR-044 | T013(hecha) | — | tareas hechas | — |
| FR-045 | T013(hecha) | — | comprobado por su control | `evals:legal-core:sin_activar:claude-sonnet-5-5:orden`: 0 de 21 (0,0 %), ≤ 0,0 %; `evals:legal-core:sin_activar:claude-sonnet-5-5:herramienta`: 0 de 21 (0,0 %), ≤ 0,0 % |
| FR-046 | T013(hecha) | — | comprobado por su control | `ci:Makefile:skills-check`, en make ci (verde); `ci:internal/app/skills_test.go:TestSkillsDelRepositorio`, en make ci (verde) |
| FR-047 | T013(hecha) | — | tareas hechas | — |
| FR-048 | T004(hecha) T011(hecha) T013(hecha) | — | tareas hechas | — |
| FR-050 | T011(hecha) | — | comprobado por su control | `ci:internal/evals/definicion_test.go:TestDefinicionDelJob`, en make ci (verde) |
| FR-051 | T011(hecha) | — | comprobado por su control | `ci:internal/evals/conjunto_test.go:TestEvalsDelRepositorio`, en make ci (verde); `ci:internal/evals/formato_test.go:TestFormatoDePlazos`, en make ci (verde) |
| FR-052 | T002(hecha) T005(hecha) T011(hecha) | — | comprobado por su control | `ci:internal/evals/conjunto_test.go:TestEvalsDelRepositorio`, en make ci (verde) |
| FR-053 | T012(hecha) | — | comprobado por su control | `evals:legal-core:plazo_sin_fecha:claude-sonnet-5-5:orden`: 0 de 9 (0,0 %), ≤ 0,0 %; `evals:legal-core:plazo_sin_fecha:claude-sonnet-5-5:herramienta`: 0 de 9 (0,0 %), ≤ 0,0 % |
| FR-054 | T012(hecha) | — | comprobado por su control | `evals:legal-core:cobertura_sin_declarar:claude-sonnet-5-5:orden`: 0 de 6 (0,0 %), ≤ 0,0 %; `evals:legal-core:cobertura_sin_declarar:claude-sonnet-5-5:herramienta`: 0 de 6 (0,0 %), ≤ 0,0 % |
| FR-055 | T012(hecha) | — | comprobado por su control | `ci:internal/evals/informe_test.go:TestUmbralesDeLegalCore`, en make ci (verde) |
| FR-056 | T012(hecha) | — | comprobado por su control | `ci:internal/evals/informe_test.go:TestUmbralesDeLasTresSkills`, en make ci (verde) |
| FR-057 | T011(hecha) | — | tareas hechas | — |
| FR-058 | T011(hecha) T016(hecha) | — | tareas hechas | — |
| FR-060 | T015(cuarentena) | — | EN CUARENTENA | — |
| FR-061 | T015(cuarentena) | — | EN CUARENTENA | — |
| FR-062 | T015(cuarentena) | — | EN CUARENTENA | — |
| FR-063 | T016(hecha) | — | tareas hechas | — |
| FR-070 | T016(hecha) | — | tareas hechas | — |
| FR-071 | T012(hecha) | — | comprobado por su control | `evals:boe-legislacion:sin_activar:claude-sonnet-5-5:orden`: 0 de 54 (0,0 %), ≤ 0,0 %; `evals:boe-legislacion:sin_activar:claude-sonnet-5-5:herramienta`: 0 de 54 (0,0 %), ≤ 0,0 %; `evals:boe-legislacion:afirma_lo_no_leido:claude-sonnet-5-5`: 0 de 108 (0,0 %), ≤ 2,8 %; `evals:boe-legislacion:medida_del_juez:afirma_lo_no_leido:defectos_sin_marcar`: 0 de 212 (0,0 %), ≤ 0,0 %; `evals:boe-legislacion:medida_del_juez:afirma_lo_no_leido:correctos_marcados`: 0 de 47 (0,0 %), ≤ 0,0 %; `evals:boe-legislacion:duracion_de_las_sesiones:orden`: 441, ≤ 900; `evals:boe-legislacion:duracion_de_las_sesiones:herramienta`: 430, ≤ 900; `evals:boe-legislacion:duracion_del_juez:orden`: 109, ≤ 900; `evals:boe-legislacion:duracion_del_juez:herramienta`: 112, ≤ 900; `evals:jurisprudencia:sin_activar:claude-sonnet-5-5:orden`: 0 de 30 (0,0 %), ≤ 0,0 %; `evals:jurisprudencia:sin_activar:claude-sonnet-5-5:herramienta`: 0 de 30 (0,0 %), ≤ 0,0 %; `evals:jurisprudencia:cita_sin_documento:claude-sonnet-5-5:orden`: 0 de 30 (0,0 %), ≤ 0,0 %; `evals:jurisprudencia:cita_sin_documento:claude-sonnet-5-5:herramienta`: 0 de 30 (0,0 %), ≤ 0,0 %; `evals:jurisprudencia:documento_sin_cotejar:claude-sonnet-5-5:orden`: 0 de 12 (0,0 %), ≤ 0,0 %; `evals:jurisprudencia:documento_sin_cotejar:claude-sonnet-5-5:herramienta`: 0 de 12 (0,0 %), ≤ 0,0 %; `evals:jurisprudencia:afirma_lo_no_leido:claude-sonnet-5-5`: 0 de 60 (0,0 %), ≤ 1,7 %; `evals:jurisprudencia:medida_del_juez:afirma_lo_no_leido:defectos_sin_marcar`: 0 de 125 (0,0 %), ≤ 0,0 %; `evals:jurisprudencia:medida_del_juez:afirma_lo_no_leido:correctos_marcados`: 0 de 124 (0,0 %), ≤ 0,0 %; `evals:jurisprudencia:duracion_del_juez:orden`: 43, ≤ 900; `evals:jurisprudencia:duracion_del_juez:herramienta`: 41, ≤ 900 |
| FR-072 | T012(hecha) | — | tareas hechas | — |
| FR-073 | T016(hecha) | — | tareas hechas | — |
| FR-080 | T016(hecha) | — | tareas hechas | — |
| FR-081 | T004(hecha) | — | comprobado por su control | `ci:internal/core/plazos/propiedades_test.go:TestPropiedadesDelComputo`, en make ci (verde) |
| FR-082 | T010(hecha) | — | comprobado por su control | `ci:internal/app/plazos_casos_test.go:TestCasosDePlazos`, en make ci (verde) |
| FR-083 | T007(hecha) T008(hecha) | — | comprobado por su control | `ci:internal/source/boe/fiestas/derivar_test.go:TestFestivosNacionalesDerivados`, en make ci (verde); `ci:internal/source/bocm/derivar_test.go:TestFestivosLocalesDerivados`, en make ci (verde) |
| FR-084 | T008(hecha) | — | comprobado por su control | `ci:internal/source/bocm/cotejo_test.go:TestCotejoConElCatalogo`, en make ci (verde) |
| FR-085 | T001(hecha) T009(hecha) | plazos-cobertura.txtar plazos-herramienta.txtar  | comprobado por su control | `ci:internal/app/testdata/script/h9-plazos-codigos.txtar`, en make ci (verde); `ci:internal/app/testdata/script/h9-plazos-cobertura.txtar`, en make ci (verde); `ci:internal/app/herramientas_test.go:TestHerramientasDelServidor`, en make ci (verde); `ci:internal/app/testdata/script/h9-plazos-herramienta.txtar`, en make ci (verde) |
| FR-086 | — | — | SIN TAREA | `evals:boe-legislacion:sin_activar:claude-sonnet-5-5:orden`: 0 de 54 (0,0 %), ≤ 0,0 %; `evals:boe-legislacion:sin_activar:claude-sonnet-5-5:herramienta`: 0 de 54 (0,0 %), ≤ 0,0 %; `evals:boe-legislacion:afirma_lo_no_leido:claude-sonnet-5-5`: 0 de 108 (0,0 %), ≤ 2,8 %; `evals:boe-legislacion:medida_del_juez:afirma_lo_no_leido:defectos_sin_marcar`: 0 de 212 (0,0 %), ≤ 0,0 %; `evals:boe-legislacion:medida_del_juez:afirma_lo_no_leido:correctos_marcados`: 0 de 47 (0,0 %), ≤ 0,0 %; `evals:boe-legislacion:duracion_de_las_sesiones:orden`: 441, ≤ 900; `evals:boe-legislacion:duracion_de_las_sesiones:herramienta`: 430, ≤ 900; `evals:boe-legislacion:duracion_del_juez:orden`: 109, ≤ 900; `evals:boe-legislacion:duracion_del_juez:herramienta`: 112, ≤ 900; `evals:jurisprudencia:sin_activar:claude-sonnet-5-5:orden`: 0 de 30 (0,0 %), ≤ 0,0 %; `evals:jurisprudencia:sin_activar:claude-sonnet-5-5:herramienta`: 0 de 30 (0,0 %), ≤ 0,0 %; `evals:jurisprudencia:cita_sin_documento:claude-sonnet-5-5:orden`: 0 de 30 (0,0 %), ≤ 0,0 %; `evals:jurisprudencia:cita_sin_documento:claude-sonnet-5-5:herramienta`: 0 de 30 (0,0 %), ≤ 0,0 %; `evals:jurisprudencia:documento_sin_cotejar:claude-sonnet-5-5:orden`: 0 de 12 (0,0 %), ≤ 0,0 %; `evals:jurisprudencia:documento_sin_cotejar:claude-sonnet-5-5:herramienta`: 0 de 12 (0,0 %), ≤ 0,0 %; `evals:jurisprudencia:afirma_lo_no_leido:claude-sonnet-5-5`: 0 de 60 (0,0 %), ≤ 1,7 %; `evals:jurisprudencia:medida_del_juez:afirma_lo_no_leido:defectos_sin_marcar`: 0 de 125 (0,0 %), ≤ 0,0 %; `evals:jurisprudencia:medida_del_juez:afirma_lo_no_leido:correctos_marcados`: 0 de 124 (0,0 %), ≤ 0,0 %; `evals:jurisprudencia:duracion_del_juez:orden`: 43, ≤ 900; `evals:jurisprudencia:duracion_del_juez:herramienta`: 41, ≤ 900 |
| FR-087 | T012(hecha) | — | comprobado por su control | `ci:internal/evals/informe_test.go:TestUmbralesDeLegalCore`, en make ci (verde) |
| FR-088 | T005(hecha) T011(hecha) T012(hecha) | — | comprobado por su control | `evals:legal-core:plazo_sin_fecha:claude-sonnet-5-5:orden`: 0 de 9 (0,0 %), ≤ 0,0 %; `evals:legal-core:plazo_sin_fecha:claude-sonnet-5-5:herramienta`: 0 de 9 (0,0 %), ≤ 0,0 %; `evals:legal-core:cobertura_sin_declarar:claude-sonnet-5-5:orden`: 0 de 6 (0,0 %), ≤ 0,0 %; `evals:legal-core:cobertura_sin_declarar:claude-sonnet-5-5:herramienta`: 0 de 6 (0,0 %), ≤ 0,0 % |
| FR-089 | T014(hecha) | — | comprobado por su control | `ci:internal/app/plazos_casos_test.go:TestCasosDePlazos`, en make ci (verde); `ci:internal/source/bocm/derivar_test.go:TestFestivosLocalesDerivados`, en make ci (verde); `ci:internal/evals/informe_test.go:TestUmbralesDeLegalCore`, en make ci (verde); `ci:internal/evals/informe_test.go:TestUmbralesDeLasTresSkills`, en make ci (verde) |
| FR-090 | T015(cuarentena) | — | EN CUARENTENA | — |
| FR-091 | T016(hecha) | — | tareas hechas | — |
| SC-001 | T011(hecha) T012(hecha) | — | comprobado por su control | `evals:legal-core:plazo_sin_fecha:claude-sonnet-5-5:orden`: 0 de 9 (0,0 %), ≤ 0,0 %; `evals:legal-core:plazo_sin_fecha:claude-sonnet-5-5:herramienta`: 0 de 9 (0,0 %), ≤ 0,0 %; `evals:legal-core:cobertura_sin_declarar:claude-sonnet-5-5:orden`: 0 de 6 (0,0 %), ≤ 0,0 %; `evals:legal-core:cobertura_sin_declarar:claude-sonnet-5-5:herramienta`: 0 de 6 (0,0 %), ≤ 0,0 %; `evals:legal-core:sin_activar:claude-sonnet-5-5:orden`: 0 de 21 (0,0 %), ≤ 0,0 %; `evals:legal-core:sin_activar:claude-sonnet-5-5:herramienta`: 0 de 21 (0,0 %), ≤ 0,0 %; `ci:internal/evals/definicion_test.go:TestDefinicionDelJob`, en make ci (verde) |
| SC-002 | — | — | SIN TAREA | — |
| SC-003 | — | — | SIN TAREA | — |
| SC-004 | T004(hecha) T016(hecha) | — | comprobado por su control | `ci:internal/core/plazos/propiedades_test.go:TestPropiedadesDelComputo`, en make ci (verde) |
| SC-005 | T010(hecha) T016(hecha) | — | comprobado por su control | `ci:internal/app/plazos_casos_test.go:TestCasosDePlazos`, en make ci (verde) |
| SC-006 | T007(hecha) T008(hecha) T016(hecha) | — | comprobado por su control | `ci:internal/source/boe/fiestas/derivar_test.go:TestFestivosNacionalesDerivados`, en make ci (verde); `ci:internal/source/bocm/derivar_test.go:TestFestivosLocalesDerivados`, en make ci (verde) |
| SC-007 | T008(hecha) T016(hecha) | — | comprobado por su control | `ci:internal/source/bocm/cotejo_test.go:TestCotejoConElCatalogo`, en make ci (verde) |
| SC-008 | T014(hecha) T016(hecha) | — | comprobado por su control | `ci:internal/app/plazos_casos_test.go:TestCasosDePlazos`, en make ci (verde); `ci:internal/source/bocm/derivar_test.go:TestFestivosLocalesDerivados`, en make ci (verde); `ci:internal/evals/informe_test.go:TestUmbralesDeLegalCore`, en make ci (verde); `ci:internal/evals/informe_test.go:TestUmbralesDeLasTresSkills`, en make ci (verde) |
| SC-009 | T001(hecha) T009(hecha) T016(hecha) | plazos-calcular.txtar plazos-sin-efectos.txtar  | comprobado por su control | `ci:internal/arch_test.go:TestArquitectura`, en make ci (verde); `ci:internal/app/plazos_test.go:TestAppletPlazos`, en make ci (verde); `ci:internal/app/testdata/script/h9-plazos-sin-efectos.txtar`, en make ci (verde); `ci:internal/app/testdata/script/h9-plazos-calcular.txtar`, en make ci (verde) |
| SC-010 | T013(hecha) T016(hecha) | — | comprobado por su control | `ci:Makefile:skills-check`, en make ci (verde); `ci:internal/app/skills_test.go:TestSkillsDelRepositorio`, en make ci (verde); `ci:Makefile:skills-check`, en make ci (verde) |
| SC-011 | T012(hecha) T016(hecha) | — | comprobado por su control | `ci:internal/evals/informe_test.go:TestUmbralesDeLasTresSkills`, en make ci (verde) |
| SC-012 | T012(hecha) T016(hecha) | — | comprobado por su control | `ci:internal/evals/informe_test.go:TestUmbralesDeLegalCore`, en make ci (verde) |
| SC-013 | T015(cuarentena) | — | EN CUARENTENA | — |

## 7. Cambios posteriores a la revisión final

Cada commit posterior a lo que juzgó la primera ronda de la revisión final (`24aaf80`), con la ronda que lo vio y su veredicto, y lo que toca fuera de `gates/` (ADR 0030):

- `9243c55` fix(H9): motivos de la revisión final: lo vio la ronda 2 (juez A: rechazado; juez B: rechazado). Toca `CHANGELOG.md`, `README.md`, `data/festivos/nacionales-y-autonomicos.yaml`, `internal/app/plazos_casos_test.go`, `internal/source/boe/fiestas/resolucion.go`, `internal/source/boe/fiestas/resolucion_test.go`, `contracts/casos-de-plazos.md`, `contracts/datos-de-plazos.md`, `research.md`.
- `0c9101a` fix(H9): motivos de la revisión final: lo vio la ronda 3 (juez A: aprobado; juez B: aprobado). Toca `plan.md`, `quickstart.md`.
- `a514aa6` docs(H9): veredictos de la revisión final: solo registros de `gates/`.
- `3f14e3a` docs(H9): registros del run: solo registros de `gates/`.
- `e3d6426` fix(H9): cierre en la plataforma: lo vio la ronda 4 (juez A: aprobado; juez B: rechazado). Toca `CHANGELOG.md`, `skills/legal-core/SKILL.md`, `contracts/skill-legal-core.md`, `quickstart.md`, `research.md`.
- `a172b3f` fix(H9): motivos de la revisión final: lo vio la ronda 5 (juez A: aprobado; juez B: aprobado). Toca `research.md`.
- `017260d` docs(H9): veredictos de la revisión final: solo registros de `gates/`.
- `92fc232` docs(H9): registros del run: solo registros de `gates/`.

### Cambios que ningún juez vio

Ninguno.

## 8. Cómo comprobarlo y consumo

- Escenarios manuales: `specs/023-h9-internal-core-plazos/quickstart.md`. Suite de aceptación congelada: `specs/023-h9-internal-core-plazos/aceptacion/` (activada en `internal/app/testdata/script/`).
- Run `1bd01ed1`: 21 h 5 min de reloj. Coste por paso y por rol: `scripts/coste-run.sh 1bd01ed1`.
