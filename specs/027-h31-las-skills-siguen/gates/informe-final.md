# Informe del hito H31 · Las skills siguen la especificación Agent Skills: guía de autoría y gates

Generado por el workflow `hito` el 2026-10-25T05:23:45Z, sobre `03cf8d9` de `027-h31-las-skills-siguen`.
Lo escribe scripts/workflow/informe.sh sin modelo, desde los artefactos de `specs/027-h31-las-skills-siguen/`. Fusionar (squash-merge) es una decisión humana:
si algo de lo que sigue no es lo que se quería, se corrige el fichero del hito en docs/roadmap/in-progress/ y se relanza.

## 1. Estado

- **make ci local**: verde.
- **CI y evals remotos** sobre `03cf8d9` (medición 1): verde; es el producto de la cabeza (lo posterior solo toca `gates/`).
- **Evals por skill**: boe-legislacion aprobado, jurisprudencia aprobado, legal-core aprobado (tasas en la sección 3).
- **Umbrales del job**: boe-legislacion: 20 umbrales, los 12 que deciden se cumplen y 8 solo se publican (5 por encima de su umbral: `cuenta_su_proceso:claude-sonnet-5-5:orden`, `cuenta_su_proceso:claude-sonnet-5-5:herramienta`, `fuera_de_lo_grabado:claude-haiku-4-5-20251001:orden`, `fuera_de_lo_grabado:claude-sonnet-5-5:herramienta`, `fuera_de_lo_grabado:claude-haiku-4-5-20251001:herramienta`); jurisprudencia: 20 umbrales, los 12 que deciden se cumplen y 8 solo se publican (2 por encima de su umbral: `afirma_que_existe:claude-sonnet-5-5:orden`, `afirma_que_existe:claude-sonnet-5-5:herramienta`); legal-core: 19 umbrales, los 13 que deciden se cumplen y 6 solo se publican (4 por encima de su umbral: `fuera_de_lo_grabado:claude-haiku-4-5-20251001:orden`, `lectura_sin_intentar:claude-haiku-4-5-20251001:orden`, `fuera_de_lo_grabado:claude-haiku-4-5-20251001:herramienta`, `lectura_sin_intentar:claude-haiku-4-5-20251001:herramienta`) (sección 3).
- **Revisión final**: juez A aprobado (con `claude-opus-5-5`), juez B aprobado (con `claude-fable-5-1`), 2 rondas.
- **Sesiones con otro modelo que el de su rol**: ninguna (43 sesiones registradas, todas con el modelo que `workflow.yml` da a su rol).
- **Cambios que ningún juez vio**: ninguno.
- **Tareas**: 18 hechas, 0 en cuarentena, 0 pendientes sin cuarentena.
- **Diff**: 25 commits; 112 files changed, 14597 insertions(+), 1351 deletions(-).

## 2. Supuestos y pendientes

Decisiones que el run tomó sin preguntar, ordenadas por impacto: cada paso que escribe una etiqueta la suya (ADR 0028).

### Cambian el comportamiento visible, el alcance o una skill

**Comportamiento (salida, códigos, ficheros, argumentos)**

- corrector_spec: FR-012 contradecía «Edge Cases» (un fichero declarado sin cabecera) → un fichero de `references/` se clasifica primero por la declaración en `kitlegal-referencias`: el declarado es la referencia generada con o sin cabecera (deriva si difiere; `skills-sync` lo rehace), el no declarado con cabecera es sobrante (deriva; lo retira) y solo el `.md` no declarado sin cabecera es una referencia a mano; es lo que hace hoy `skills-sync` y lo que dice la entrega 4 del hito. SC-003 lleva tam… (entera en `specs/027-h31-las-skills-siguen/gates/supuestos.md` o `clarify-respuestas.json`)
- corrector_spec: FR-032 frente a FR-028, qué voto que no llega a darse se repite → lectura conservadora, la del juez de las evals que se cita (H24): se repite una vez el voto nulo y el que corta el tope de un voto; la sesión con error, la salida sin la forma del esquema y el límite de uso dejan la descripción sin juzgar, sin repetición, y el job falla. SC-007 lleva el voto cortado por el tope y la sesión con error.
- T009: el contrato (§2) no fija el orden de los defectos nuevos de `name` entre los de hoy → tras el de la longitud y antes del de «distinto del nombre del directorio», que sigue el último, en el orden en que D11 los enumera (palabra reservada, etiqueta XML); un `name` que contiene `anthropic` y `claude` da un defecto por palabra, en ese orden; el de la etiqueta de `description` va donde iba el de «< o >», tras el de la longitud.
- T010: el cuerpo es lo que sigue a la línea `---` que cierra el frontmatter, y el spec no dice qué se mide en un `SKILL.md` sin esa línea (sin frontmatter o con uno sin cierre), en el que el límite de 300 líneas sí daba su defecto → no hay cuerpo que medir y el único defecto es el del frontmatter, que ya da `declaracionDeLaSkill`: es la lectura que menos añade (no inventa un cuerpo que FR-004 no define) y la skill sale en rojo igual; lo fija el caso `frontmatter-sin-cierre` de `TestCuerpoDeLaSkil… (entera en `specs/027-h31-las-skills-siguen/gates/supuestos.md` o `clarify-respuestas.json`)
- T011: el §2.3 dice que un bloque de código va hasta la línea que lo cierra «con la misma marca», sin decir si cuenta una valla más larga o con texto detrás → la cierra la línea con la misma marca, repetida al menos tantas veces como la que lo abrió y sin nada detrás salvo blancos, que es como lo lee un visor de Markdown (CommonMark): un «```text» dentro de un bloque no lo cierra, y un «`````» sí; con las skills de hoy, que abren y cierran con tres, da lo mismo que la marca exacta. Lo fija `TestE… (entera en `specs/027-h31-las-skills-siguen/gates/supuestos.md` o `clarify-respuestas.json`)
- T011: FR-008 dice «otro fichero de `references/` de la skill» sin decir qué pasa con un destino que no existe o que no es una referencia → cuenta solo el enlace a otra referencia de la skill, generada (como la regenera su fila, esté o no en el árbol) o a mano, y una sola vez cada una, en el orden en que aparece; un destino que no es ninguna referencia —un fichero que no existe, una entrada que no es `.md`, un sobrante— no es otra referencia y no da defecto. Es la lectura que menos añade: una ref… (entera en `specs/027-h31-las-skills-siguen/gates/supuestos.md` o `clarify-respuestas.json`)
- T011: el contrato no fija qué se mira cuando el frontmatter o los datos tienen defectos, ni el orden de los defectos nuevos → sin el frontmatter o los datos en orden no hay referencias regeneradas y FR-007 a FR-009 no se miran (`ReferenciasAMano` tampoco podría leer lo declarado); FR-006, que no depende de nada declarado, se mira siempre que hay `SKILL.md`. Los defectos de `references/` van tras los de los datos y antes del de `scripts/`: los de FR-006, por el nombre de la entrada; después, refe… (entera en `specs/027-h31-las-skills-siguen/gates/supuestos.md` o `clarify-respuestas.json`)
- T011: el §3 da como defecto de FR-006 «en este orden» toda entrada que no es un fichero regular `.md`, también la de una referencia declarada que hoy era la deriva `contenido-distinto (no es un fichero regular)` → es el defecto `references/<entrada>: no es un fichero .md` y `skills-sync` no la retira; la deriva queda solo para el árbol que cambia entre `Regenerar` y `Escribir`, que `Escribir` sigue deshaciendo sin escribir en el destino de un enlace.
- T011: FR-007 pide un enlace «cuyo destino es su ruta relativa `references/<fichero>.md`», y FR-008 resuelve con `path.Join`, sin decir si FR-007 también → FR-007 compara el destino, sin `#…`, con `references/<fichero>.md` tal cual (`./references/x.md` no cuenta), como lo dice el contrato; un enlace es `[texto](destino)` con el destino terminado en `)` o en un espacio, así que `[a](a.md` al final de la línea no lo es; y la lista del índice de FR-009 se lee, como los apartados, fuera de los bloque… (entera en `specs/027-h31-las-skills-siguen/gates/supuestos.md` o `clarify-respuestas.json`)
- T013: data-model §4 y contracts/juez-de-las-descripciones.md §6 dan `frase` y `motivo` «del primer voto» y cada voto de `votos` con solo `respuesta`, `frase` y `motivo`, sin decir qué es el primer voto cuando el primero es nulo o cortado, ni si el juicio lleva el número de cada voto, si es nulo o los votos que el tope cortó → del primer voto cuenta el que cuenta en la regla de los votos, el último con el número 1 (la repetición de un nulo o de un cortado, como `frasesQueCuentan`), porque la fras… (entera en `specs/027-h31-las-skills-siguen/gates/supuestos.md` o `clarify-respuestas.json`)
- T014: con una descripción cambiada y la lectura de los ficheros de la propuesta fallida, ¿se busca igualmente en `main` el juicio de la descripción? → no: como en H30, la primera lectura que falla para la búsqueda y lo que queda se mide; la descripción de cada skill se juzga (`juzgada: true`) y el trabajo no lee nada más de la plataforma. Lo fija el caso `descripcion-de-legal-core-sin-los-ficheros-de-la-propuesta` de `TestDecisionDelJob`.
- T015: contracts/juez-de-las-descripciones.md §6 escribe el motivo del umbral en rojo como «umbral habla_en_primera_o_segunda_persona: 1, y tiene que ser <= 0», pero T015 y research D7 piden «el motivo de hoy» y un umbral con `total` → es el de `motivosDeLosUmbrales`, el de cualquier umbral con total que decide y no se cumple: «umbral habla_en_primera_o_segunda_persona: 1 de 1 (100,0 %), y tiene que ser ≤ 0,0 %», sin nombrar la frase, que va en el bloque y en la sección «La descripción»; lo fijan… (entera en `specs/027-h31-las-skills-siguen/gates/supuestos.md` o `clarify-respuestas.json`)
- T015: el contrato no dice dónde van el bloque `descripcion` en `informe.json` ni la sección «La descripción» en `informe.md`, ni el texto de la sección → el bloque, detrás de `origen` y delante de `tasas`, y la sección, detrás de «Medidas y reutilizadas» y delante de «Tasas por eval», sin mover nada de lo que ya fija el orden del informe; la sección dice `Skill:`, `Juicio:` («juzgada en esta ejecución» o la procedencia como la escribe la línea de una eval reutilizada) y, según el juicio, `Sin ju… (entera en `specs/027-h31-las-skills-siguen/gates/supuestos.md` o `clarify-respuestas.json`)
- T015: el contrato da el motivo de la descripción sin juzgar sin el prefijo «de la ejecución, no de la skill: » que llevan el de las respuestas sin juzgar y el del instrumento sin medir → se escribe tal cual lo da el contrato y research D7, sin prefijo, detrás del motivo de las respuestas sin juzgar; no cita ningún umbral y se distingue sin modelo del de la marcada, que empieza por «umbral ».
- T015: con el bloque `descripcion` y la medida del juez de las descripciones sin corresponder o sin cumplirse, el informe del instrumento sin medir de una skill sin juez de las respuestas (hoy legal-core) llevaría el juez con el modelo y la versión fijados, como el de una skill con juez → no lo lleva: `juez` es null y la sección «Juez» dice que la skill no lo tiene, como en cualquier otro informe de esa skill (el contrato de `Informe.Juez`); y ese informe no lleva el bloque `descripcion`, ni su u… (entera en `specs/027-h31-las-skills-siguen/gates/supuestos.md` o `clarify-respuestas.json`)
- T015: el contrato ordena «las sesiones de las evals» y después «los votos de la descripción», sin decir cuándo se leen la descripción y quien da el votante → los dos se preparan tras comprobar las medidas y antes de abrir ninguna sesión, para que un SKILL.md que no se lee, la juzgada sin quien dé su votante o la reutilizada sin juicio paren el job con su error antes de abrir nada (como hoy un plan sin sentido); los votos se piden detrás de las sesiones. Con la reutilizada no se pide el votante.
- T017: contracts/juez-de-las-descripciones.md §7 pide una línea en el registro por caso sin juzgar, pero no dice si con alguno se escribe la medida → no se escribe, y el guion no imprime ninguna marca, como en la medida del juez de una skill (FR-053 de H24): una medida con casos sin juzgar no es una medida, y research D9 la llama «el flujo de la medida de H24 y H25 con otros casos».
- T017: el guion admite `descripciones` sin que nada diga si su línea de uso cambia → sigue siendo `scripts/evals-medir-juez.sh <skill>`, la lectura que menos añade; `descripciones` lo documentan la cabecera del guion, el trabajo `medida` y `CHANGELOG.md`, y con él el punto de entrada recibe las banderas de la medida de una skill sin `-skill`, porque el juez de las descripciones no es de ninguna.

**Alcance (lo que queda fuera o dentro del hito)**

- corrector_spec: lo que deja fuera la Clarificación Q4 solo estaba en FR-009 → «Fuera de alcance» lleva cambiar los generadores de las referencias generadas para que escriban su índice; lo adapta el cambio de `data/` que lleve una por encima de 100 líneas.
- T004: el control 8 dice «Ley N/AAAA», «LO…» y «RD…» sin decir si «LO» y «RD» son abreviaturas que también se comprueban → se leen como las formas que ya reconoce `normaNombrada` (Ley, Ley Orgánica, Real Decreto, Real Decreto-ley, Real Decreto Legislativo), sin ampliar la expresión: FR-016 amplía qué ficheros lee el control, no qué reconoce, y ninguna skill escribe hoy «LO N/AAAA» ni «RD N/AAAA»; la exigencia de no pasar en vacío cuenta las normas de la skill entera, que es lo que el control lee.
- clarify Q1 (criterio c): El cierre por nombre de la parte (b) lee `SKILL.md` y sus referencias a mano. Una skill nombrada solo en una referencia a mano de una skill de la eval (el nombre de otro directorio de `skills/` entre límites que no son [a-z0-9-], la regla `nombra` de hoy) entra en (b) con su `SKILL.md` y sus `references/`, y el cierre sigue desde ella como desde cualquier skill nombrada. Las referencias generadas no se leen para nombres: salen de `data/`, que entra por (c), y hoy ninguna… (entera en `specs/027-h31-las-skills-siguen/gates/supuestos.md` o `clarify-respuestas.json`)
- clarify Q4 (criterio e): No material: el spec no lo especifica, y los generadores no cambian en este hito. Ninguna vía real de este hito lleva una referencia generada por encima de 100 líneas: la más larga, `normas.md`, tiene 24, y el hito no cambia el contenido de las referencias generadas (FR-024). `data/` no lo cambia quien usa el kit, y ningún fichero suyo se escribe a mano: solo lo cambia código a partir de lo grabado en un hito. Con todo lo que este hito puede producir, las dos opciones da… (entera en `specs/027-h31-las-skills-siguen/gates/supuestos.md` o `clarify-respuestas.json`)

**Skill (lo que pide, dice o comprueba una skill)**

- T003: el control 7 compara la línea ⚠ de una referencia a mano, sin los blancos de los extremos, con «una línea del SKILL.md», sin decir si esta también va sin ellos → también sin ellos: el cuerpo enseña sus formas con la sangría de un elemento de lista (las de la redacción fechada y la línea sin consulta de boe-legislacion), y sin quitarla ninguna referencia podría repetirlas; el marcador de lista y el énfasis sí forman parte de la línea comparada («otro marcador» es una variante mal escrita), … (entera en `specs/027-h31-las-skills-siguen/gates/supuestos.md` o `clarify-respuestas.json`)
- T006: contracts/skills-reescritas.md da las nueve reglas de legal-core «sin cambio», y research.md («Tabla de reglas», L48 y L76) pone en la regla 9 «ni se exponen los de la norma que lo rige» y en la regla 8 «tampoco para «comprobarla»», que en v1.2.3 decían los pasos 7 y 11 → las reglas 8 y 9 llevan esas dos frases y las demás no cambian: es la tabla de reglas la que dice dónde queda cada una, y sin ellas L48 y L76 solo quedarían en una referencia o en ninguna parte.
- T006: el contrato pide las nueve formas `⚠` del plazo «byte a byte» sin decir si van en bloques `text` separados → cada línea igual que en v1.2.3, y la forma de la línea del plazo y su ejemplo del contencioso en un mismo bloque, con el ejemplo nombrado en la frase de encima: ningún control de legal-core lee los bloques del plazo (solo el de la línea sin consulta, en la regla 7), y así el cuerpo queda en 10 860 caracteres.
- T007: contracts/skills-reescritas.md, «Riesgos», propone devolver al paso 2 de boe-legislacion «con herramientas, `graph_check` con esa `norma` y el bloque en `bloques`» «si el margen lo permite», sin fijar cuánto margen → no vuelve: con el resto del reparto el cuerpo queda en 10 976 de 11 000 caracteres, y la frase está en references/con_herramientas.md, «Las órdenes del protocolo», que el párrafo del protocolo manda leer antes de usar una herramienta; es el reparto base del contrato, y lo deci… (entera en `specs/027-h31-las-skills-siguen/gates/supuestos.md` o `clarify-respuestas.json`)
- T007: la tabla de reglas lleva B79 («una por código de aviso») a la regla 3 del cuerpo, que con el margen solo cabe con las tres etiquetas → la regla 3 lleva las tres etiquetas, y la correspondencia de cada una con su código (`derogada`, `vigencia-agotada`, `consolidacion-no-finalizada`) va en references/como_se_cita.md, «Los avisos», el apartado al que la tabla lleva de dónde sale la etiqueta (B78).
- corrector_revision: el juez A pide devolver al cuerpo de `boe-legislacion` las condiciones de B40, B45, B56, B63, B64, B65 y B122, y el cuerpo, a 10 976 de 11 000, no las cabe → se devuelven con el texto de `main` (la regla 5 con «si se pregunta por un municipio», «siempre que lo preguntado les corresponda» y los ejemplos de la variación autonómica) y, para que quepa (10 993), sale del cuerpo solo detalle que no es regla invariante ni forma fija y que ya está en una referencia a mano que el cuer… (entera en `specs/027-h31-las-skills-siguen/gates/supuestos.md` o `clarify-respuestas.json`)
- corrector_revision: el juez A pide en `legal-core` devolver «de vencimiento» a los pasos 9 y 11 y marcar como ejemplo la línea del contencioso → «sin la línea del plazo ni ninguna fecha de vencimiento» (paso 9), «ninguna otra fecha de vencimiento (regla 8)» (paso 11) y «y, de ejemplo, la de los dos preceptos del recurso contencioso-administrativo»; el paso 7, que en `main` ya decía «ni ninguna fecha», no cambia.

### Del propio run


- Observaciones de los jueces, por debajo del umbral y sin corregir: 14 en `spec-r2.json`, 7 en `plan-r2.json`, 4 en `tasks-r2.json`, 10 en `revision-a-r2.json`, 8 en `revision-b-r2.json`.

### Internos (20)

Decisiones que no cambian nada observable (técnica, estructura, tests), por autor: T001 (1), T002 (2), T005 (1), T006 (1), T008 (3), T009 (1), T012 (3), T014 (1), T016 (1), T017 (1), T018 (1), corrector_plan (1), corrector_revision (1), corrector_spec (2). Enteras en `specs/027-h31-las-skills-siguen/gates/supuestos.md`.

## 3. Evals sobre la cabeza

**boe-legislacion**: aprobado sobre `03cf8d9`; decide `claude-sonnet-5-5` con 2 de 3; 23 evals, 0 nuevas, 8 informativas.

| Eval | claude-sonnet-5-5 | claude-haiku-4-5-20251001 | claude-sonnet-5-5 (herramienta) | claude-haiku-4-5-20251001 (herramienta) | Marca |
|---|---|---|---|---|---|
| `01-lpac-articulo-21.yaml` | 3/3 | 3/3 | 3/3 | 0/3 |  |
| `02-lcsp-contrato-menor.yaml` | 3/3 | 2/3 | 3/3 | 1/3 |  |
| `03-lrbrl-atribuciones-del-pleno.yaml` | 3/3 | 2/3 | 3/3 | 0/3 |  |
| `04-lgt-prescripcion.yaml` | 3/3 | 1/3 | 3/3 | 0/3 |  |
| `05-trlrhl-impuestos-municipales.yaml` | 3/3 | 2/3 | 3/3 | 0/3 |  |
| `06-irpf-rendimientos-del-trabajo.yaml` | 3/3 | 2/3 | 3/3 | 0/3 |  |
| `07-lrjsp-principio-de-legalidad.yaml` | 3/3 | 0/3 | 3/3 | 0/3 |  |
| `08-ltaibg-plazo-de-resolucion.yaml` | 3/3 | 1/3 | 3/3 | 1/3 |  |
| `09-constitucion-articulo-140.yaml` | 3/3 | 2/3 | 3/3 | 0/3 |  |
| `10-et-vacaciones.yaml` | 3/3 | 1/3 | 3/3 | 1/3 |  |
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
| `22-lcsp-contrato-menor-en-2019.yaml` | 3/3 | 2/3 | 3/3 | 0/3 |  |
| `23-lpac-articulo-21-en-2016.yaml` | 3/3 | 3/3 | 3/3 | 1/3 |  |

Umbrales que publica el job (ADR 0029):

| Umbral | Medida | Condición | Cumple | Hace fallar el job |
|---|---|---|---|---|
| `sin_activar:claude-sonnet-5-5:orden` | 0 de 60 (0,0 %) | ≤ 0,0 % | sí | sí |
| `afirma_lo_no_leido:claude-sonnet-5-5:orden` | 1 de 60 (1,7 %) | ≤ 2,8 % | dentro | no: solo se publica, no es un control |
| `cuenta_su_proceso:claude-sonnet-5-5:orden` | 10 de 60 (16,7 %) | ≤ 0,0 % | por encima | no: solo se publica, no es un control |
| `redaccion_sin_fechar:claude-sonnet-5-5:orden` | 0 de 6 (0,0 %) | ≤ 0,0 % | sí | sí |
| `sin_activar:claude-sonnet-5-5:herramienta` | 0 de 60 (0,0 %) | ≤ 0,0 % | sí | sí |
| `afirma_lo_no_leido:claude-sonnet-5-5:herramienta` | 0 de 60 (0,0 %) | ≤ 2,8 % | dentro | no: solo se publica, no es un control |
| `cuenta_su_proceso:claude-sonnet-5-5:herramienta` | 21 de 60 (35,0 %) | ≤ 0,0 % | por encima | no: solo se publica, no es un control |
| `redaccion_sin_fechar:claude-sonnet-5-5:herramienta` | 0 de 6 (0,0 %) | ≤ 0,0 % | sí | sí |
| `afirma_lo_no_leido:claude-sonnet-5-5` | 1 de 120 (0,8 %) | ≤ 2,8 % | sí | sí |
| `medida_del_juez:afirma_lo_no_leido:defectos_sin_marcar` | 0 de 212 (0,0 %) | ≤ 0,0 % | sí | sí |
| `medida_del_juez:afirma_lo_no_leido:correctos_marcados` | 0 de 47 (0,0 %) | ≤ 0,0 % | sí | sí |
| `duracion_de_las_sesiones:orden` | 410 | ≤ 900 | sí | sí |
| `duracion_de_las_sesiones:herramienta` | 348 | ≤ 900 | sí | sí |
| `duracion_del_juez:orden` | 98 | ≤ 900 | sí | sí |
| `duracion_del_juez:herramienta` | 103 | ≤ 900 | sí | sí |
| `fuera_de_lo_grabado:claude-sonnet-5-5:orden` | 0 de 66 (0,0 %) | ≤ 0,0 % | dentro | no: solo se publica, no es un control |
| `fuera_de_lo_grabado:claude-haiku-4-5-20251001:orden` | 35 de 42 (83,3 %) | ≤ 0,0 % | por encima | no: solo se publica, no es un control |
| `fuera_de_lo_grabado:claude-sonnet-5-5:herramienta` | 4 de 66 (6,1 %) | ≤ 0,0 % | por encima | no: solo se publica, no es un control |
| `fuera_de_lo_grabado:claude-haiku-4-5-20251001:herramienta` | 5 de 42 (11,9 %) | ≤ 0,0 % | por encima | no: solo se publica, no es un control |
| `habla_en_primera_o_segunda_persona` | 0 de 1 (0,0 %) | ≤ 0,0 % | sí | sí |

**jurisprudencia**: aprobado sobre `03cf8d9`; decide `claude-sonnet-5-5` con 2 de 3; 10 evals, 0 nuevas, 0 informativas.

| Eval | claude-sonnet-5-5 | claude-haiku-4-5-20251001 | claude-sonnet-5-5 (herramienta) | claude-haiku-4-5-20251001 (herramienta) | Marca |
|---|---|---|---|---|---|
| `01-existe-con-numero-y-fecha.yaml` | 3/3 | 3/3 | 3/3 | 3/3 |  |
| `02-resumen-sin-documento.yaml` | 3/3 | 3/3 | 3/3 | 2/3 |  |
| `03-por-materia.yaml` | 3/3 | 3/3 | 3/3 | 2/3 |  |
| `04-documento-pegado.yaml` | 3/3 | 3/3 | 3/3 | 3/3 |  |
| `05-tribunal-constitucional.yaml` | 3/3 | 3/3 | 3/3 | 3/3 |  |
| `06-documento-que-no-es-el-pedido.yaml` | 3/3 | 0/3 | 3/3 | 0/3 |  |
| `07-resumen-de-una-conocida.yaml` | 3/3 | 2/3 | 3/3 | 3/3 |  |
| `08-doctrina-con-el-fallo-delante.yaml` | 3/3 | 3/3 | 3/3 | 1/3 |  |
| `09-de-que-trata-con-la-ficha-sola.yaml` | 3/3 | 0/3 | 3/3 | 0/3 |  |
| `10-doctrina-dada-por-hecha.yaml` | 3/3 | 3/3 | 3/3 | 1/3 |  |

Umbrales que publica el job (ADR 0029):

| Umbral | Medida | Condición | Cumple | Hace fallar el job |
|---|---|---|---|---|
| `sin_activar:claude-sonnet-5-5:orden` | 0 de 30 (0,0 %) | ≤ 0,0 % | sí | sí |
| `afirma_lo_no_leido:claude-sonnet-5-5:orden` | 0 de 30 (0,0 %) | ≤ 1,7 % | dentro | no: solo se publica, no es un control |
| `afirma_que_existe:claude-sonnet-5-5:orden` | 3 de 30 (10,0 %) | ≤ 0,0 % | por encima | no: solo se publica, no es un control |
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
| `duracion_del_juez:orden` | 38 | ≤ 900 | sí | sí |
| `duracion_del_juez:herramienta` | 40 | ≤ 900 | sí | sí |
| `fuera_de_lo_grabado:claude-sonnet-5-5:orden` | 0 de 30 (0,0 %) | ≤ 0,0 % | dentro | no: solo se publica, no es un control |
| `fuera_de_lo_grabado:claude-haiku-4-5-20251001:orden` | 0 de 30 (0,0 %) | ≤ 0,0 % | dentro | no: solo se publica, no es un control |
| `fuera_de_lo_grabado:claude-sonnet-5-5:herramienta` | 0 de 30 (0,0 %) | ≤ 0,0 % | dentro | no: solo se publica, no es un control |
| `fuera_de_lo_grabado:claude-haiku-4-5-20251001:herramienta` | 0 de 30 (0,0 %) | ≤ 0,0 % | dentro | no: solo se publica, no es un control |
| `habla_en_primera_o_segunda_persona` | 0 de 1 (0,0 %) | ≤ 0,0 % | sí | sí |

**legal-core**: aprobado sobre `03cf8d9`; decide `claude-sonnet-5-5` con 2 de 3; 12 evals, 0 nuevas, 0 informativas.

| Eval | claude-sonnet-5-5 | claude-haiku-4-5-20251001 | claude-sonnet-5-5 (herramienta) | claude-haiku-4-5-20251001 (herramienta) | Marca |
|---|---|---|---|---|---|
| `01-territorio-municipio-cubierto.yaml` | 3/3 | 0/3 | 3/3 | 0/3 |  |
| `02-territorio-municipio-no-cubierto.yaml` | 3/3 | 0/3 | 3/3 | 0/3 |  |
| `03-no-activa-receta-de-cocina.yaml` | 3/3 | 3/3 | 3/3 | 3/3 |  |
| `04-sin-binario-ni-servidor.yaml` | 3/3 | 0/3 | — | — |  |
| `05-plazo-reposicion-leganes.yaml` | 3/3 | 0/3 | 3/3 | 0/3 |  |
| `06-plazo-reposicion-tordesillas.yaml` | 3/3 | 0/3 | 3/3 | 0/3 |  |
| `07-cobertura-liquidacion-tributaria.yaml` | 3/3 | 0/3 | 3/3 | 0/3 |  |
| `08-cobertura-multa-de-trafico.yaml` | 3/3 | 0/3 | 3/3 | 0/3 |  |
| `09-plazo-contencioso-agosto.yaml` | 3/3 | 0/3 | 3/3 | 0/3 |  |
| `10-plazo-reposicion-anterior-a-la-lpac.yaml` | 3/3 | 0/3 | 3/3 | 0/3 |  |
| `11-plazo-reposicion-sin-leer-el-precepto.yaml` | 3/3 | 0/3 | 3/3 | 0/3 |  |
| `12-cobertura-seguridad-social-sin-leer-la-da-primera.yaml` | 3/3 | 0/3 | 3/3 | 0/3 |  |

Umbrales que publica el job (ADR 0029):

| Umbral | Medida | Condición | Cumple | Hace fallar el job |
|---|---|---|---|---|
| `sin_activar:claude-sonnet-5-5:orden` | 0 de 30 (0,0 %) | ≤ 0,0 % | sí | sí |
| `plazo_sin_fecha:claude-sonnet-5-5:orden` | 0 de 12 (0,0 %) | ≤ 0,0 % | sí | sí |
| `cobertura_sin_declarar:claude-sonnet-5-5:orden` | 0 de 9 (0,0 %) | ≤ 0,0 % | sí | sí |
| `vigencia_sin_declarar:claude-sonnet-5-5:orden` | 0 de 3 (0,0 %) | ≤ 0,0 % | sí | sí |
| `redaccion_sin_fechar:claude-sonnet-5-5:orden` | 0 de 9 (0,0 %) | ≤ 0,0 % | sí | sí |
| `sin_activar:claude-sonnet-5-5:herramienta` | 0 de 30 (0,0 %) | ≤ 0,0 % | sí | sí |
| `plazo_sin_fecha:claude-sonnet-5-5:herramienta` | 0 de 12 (0,0 %) | ≤ 0,0 % | sí | sí |
| `cobertura_sin_declarar:claude-sonnet-5-5:herramienta` | 0 de 9 (0,0 %) | ≤ 0,0 % | sí | sí |
| `vigencia_sin_declarar:claude-sonnet-5-5:herramienta` | 0 de 3 (0,0 %) | ≤ 0,0 % | sí | sí |
| `redaccion_sin_fechar:claude-sonnet-5-5:herramienta` | 0 de 9 (0,0 %) | ≤ 0,0 % | sí | sí |
| `fuera_de_lo_grabado:claude-sonnet-5-5:orden` | 0 de 33 (0,0 %) | ≤ 0,0 % | dentro | no: solo se publica, no es un control |
| `lectura_sin_intentar:claude-sonnet-5-5:orden` | 0 de 3 (0,0 %) | ≤ 0,0 % | sí | sí |
| `fuera_de_lo_grabado:claude-haiku-4-5-20251001:orden` | 50 de 33 (151,5 %) | ≤ 0,0 % | por encima | no: solo se publica, no es un control |
| `lectura_sin_intentar:claude-haiku-4-5-20251001:orden` | 2 de 3 (66,7 %) | ≤ 0,0 % | por encima | no: solo se publica, no es un control |
| `fuera_de_lo_grabado:claude-sonnet-5-5:herramienta` | 0 de 33 (0,0 %) | ≤ 0,0 % | dentro | no: solo se publica, no es un control |
| `lectura_sin_intentar:claude-sonnet-5-5:herramienta` | 0 de 3 (0,0 %) | ≤ 0,0 % | sí | sí |
| `fuera_de_lo_grabado:claude-haiku-4-5-20251001:herramienta` | 19 de 33 (57,6 %) | ≤ 0,0 % | por encima | no: solo se publica, no es un control |
| `lectura_sin_intentar:claude-haiku-4-5-20251001:herramienta` | 2 de 3 (66,7 %) | ≤ 0,0 % | por encima | no: solo se publica, no es un control |
| `habla_en_primera_o_segunda_persona` | 0 de 1 (0,0 %) | ≤ 0,0 % | sí | sí |

Cada celda: sesiones que pasan de las abiertas; ✗, una serie que decide y no llega al umbral. Una eval informativa publica su tasa sin decidir el veredicto (ADR 0016). La regla por serie no hace cumplir un umbral agregado sobre todas las respuestas: eso solo lo hace un umbral del job que lo hace fallar (ADR 0029).

**Serie de cierres de boe-legislacion** (ADR 0038, punto 6): respuestas de las evals que activan la skill que el juez marca con sus tres votos, por modo y por clase de pregunta.

Clave de la cabeza: skill `92525ec` (árbol de `skills/boe-legislacion`), decide `claude-sonnet-5-5`, juez `claude-opus-5-5` con Claude Code 2.1.289.

`afirma_lo_no_leido`:

| Medición | Fecha | Commit | Skill | Orden | Herramienta | Nombra | Por materia | Total | En la tasa |
|---|---|---|---|---|---|---|---|---|---|
| H24-1 | 2026-10-06 | `a5bda45` | `6d65aa3` | 1/54 | 0/54 | 1/78 | 0/30 | 1/108 | no: otra skill |
| H24-2 | 2026-10-06 | `b0a7a9c` | `75e1d04` | 0/54 | 3/54 | 0/78 | 3/30 | 3/108 | no: 1 respuesta sin juzgar |
| H24-3 | 2026-10-06 | `3107ae2` | `6063317` | 0/54 | 0/54 | 0/78 | 0/30 | 0/108 | no: otra skill |
| H23-1 | 2026-10-07 | `bf4befd` | `6063317` | 0/54 | 0/50 | 0/76 | 0/28 | 0/104 | no: otra skill |
| H25-1 | 2026-10-10 | `3f99446` | `6063317` | 0/54 | 1/54 | 0/78 | 1/30 | 1/108 | no: otra skill |
| H25-2 | 2026-10-10 | `e380f0a` | `6063317` | 1/54 | 1/54 | 0/78 | 2/30 | 2/108 | no: otra skill |
| H25-3 | 2026-10-10 | `d51f3d7` | `eed129f` | 0/54 | 2/54 | 0/78 | 2/30 | 2/108 | no: otra skill |
| H26-1 | 2026-10-13 | `b29fe6c` | `eed129f` | 0/54 | 2/54 | 1/78 | 1/30 | 2/108 | no: otra skill |
| H23.1-1 | 2026-10-16 | `60e8678` | `eed129f` | 0/54 | 3/54 | 1/78 | 2/30 | 3/108 | no: otra skill |
| H9-1 | 2026-10-18 | `3f14e3a` | `eed129f` | 0/54 | 1/54 | 0/78 | 1/30 | 1/108 | no: otra skill |
| H9-2 | 2026-10-18 | `92fc232` | `eed129f` | 0/54 | 0/54 | 0/78 | 0/30 | 0/108 | no: otra skill |
| H20-1 | 2026-10-19 | `77211a9` | `49b59ac` | 0/60 | 2/60 | 0/90 | 2/30 | 2/120 | no: otra skill |
| H26.1-1 | 2026-10-21 | `444e674` | `27aeccc` | 0/60 | 1/60 | 0/90 | 1/30 | 1/120 | no: otra skill |
| H30-1 | 2026-10-22 | `2c8dda6` | `27aeccc` | 0/60 | 0/60 | 0/90 | 0/30 | 0/120 | no: otra skill |
| H31-1 | 2026-10-25 | `03cf8d9` | `92525ec` | 1/60 | 0/60 | 0/90 | 1/30 | 1/120 | sí |

Tasa acumulada con la skill como está: 1 de 120 (0,83 %), de H31-1.
Insuficiente: 120 de 500 respuestas. La regla 3 del ADR 0038 pide la tasa de 500 respuestas al menos: con menos no da escalón.

**Serie de cierres de jurisprudencia** (ADR 0038, punto 6): respuestas de las evals que activan la skill que el juez marca con sus tres votos, por modo y por clase de pregunta.

Clave de la cabeza: skill `5a9198f` (árbol de `skills/jurisprudencia`), decide `claude-sonnet-5-5`, juez `claude-opus-5-5` con Claude Code 2.1.289.

`afirma_lo_no_leido`:

| Medición | Fecha | Commit | Skill | Orden | Herramienta | Nombra | Por materia | Total | En la tasa |
|---|---|---|---|---|---|---|---|---|---|
| H25-1 | 2026-10-10 | `3f99446` | `9480bbc` | 0/30 | 0/30 | 0/48 | 0/12 | 0/60 | no: otra skill |
| H25-2 | 2026-10-10 | `e380f0a` | `9480bbc` | 0/30 | 0/30 | 0/48 | 0/12 | 0/60 | no: 1 respuesta sin juzgar |
| H25-3 | 2026-10-10 | `d51f3d7` | `9480bbc` | 0/30 | 0/30 | 0/48 | 0/12 | 0/60 | no: otra skill |
| H26-1 | 2026-10-13 | `b29fe6c` | `9480bbc` | 0/30 | 0/30 | 0/48 | 0/12 | 0/60 | no: otra skill |
| H23.1-1 | 2026-10-16 | `60e8678` | `4759aa4` | 0/30 | 0/30 | 0/48 | 0/12 | 0/60 | no: otra skill |
| H9-1 | 2026-10-18 | `3f14e3a` | `4759aa4` | 0/30 | 0/30 | 0/48 | 0/12 | 0/60 | no: otra skill |
| H9-2 | 2026-10-18 | `92fc232` | `4759aa4` | 0/30 | 0/30 | 0/48 | 0/12 | 0/60 | no: otra skill |
| H20-1 | 2026-10-19 | `77211a9` | `4759aa4` | 0/30 | 0/30 | 0/48 | 0/12 | 0/60 | no: otra skill |
| H26.1-1 | 2026-10-21 | `444e674` | `3bbd9cd` | 0/30 | 0/30 | 0/48 | 0/12 | 0/60 | no: otra skill |
| H30-1 | 2026-10-22 | `2c8dda6` | `3bbd9cd` | 0/30 | 0/30 | 0/48 | 0/12 | 0/60 | no: otra skill |
| H31-1 | 2026-10-25 | `03cf8d9` | `5a9198f` | 0/30 | 0/30 | 0/48 | 0/12 | 0/60 | sí |

Tasa acumulada con la skill como está: 0 de 60 (0,00 %), de H31-1.
Insuficiente: 60 de 500 respuestas. La regla 3 del ADR 0038 pide la tasa de 500 respuestas al menos: con menos no da escalón.

## 4. Revisión que la constitución reserva a la persona

Nada: el hito no toca fuentes, anomalías, adaptadores, fixtures, esquemas, grabaciones ni evidencias.

## 5. Tareas en cuarentena

Ninguna.

## 6. Trazabilidad

**Estado**: «tareas hechas» solo dice que las tareas que citan el requisito están marcadas; nada del run lo ha medido. «comprobado por su control» exige su fila en «Controles de umbral» de `plan.md` y que el control esté: un test o una comprobación de `make ci` que existe en la cabeza, con `make ci` en verde, o un umbral que el job de evals publica, cumple y hace fallar el job (ADR 0029). «UMBRAL NO CUMPLIDO» y «CONTROL SIN VERIFICAR» son lo que hay que mirar.

| Requisito | Tareas | Aceptación | Estado | Control de umbral |
|---|---|---|---|---|
| FR-001 | T018(hecha) | — | tareas hechas | — |
| FR-002 | T009(hecha) | — | comprobado por su control | `ci:internal/app/skills_test.go:TestSkillsDelRepositorio`, en make ci (verde) |
| FR-003 | T009(hecha) | — | comprobado por su control | `ci:internal/app/skills_test.go:TestSkillsDelRepositorio`, en make ci (verde) |
| FR-004 | T010(hecha) | — | comprobado por su control | `ci:internal/app/skills_test.go:TestSkillsDelRepositorio`, en make ci (verde) |
| FR-005 | T009(hecha) | — | comprobado por su control | `ci:internal/app/skills_test.go:TestSkillsDelRepositorio`, en make ci (verde) |
| FR-006 | T011(hecha) T018(hecha) | — | comprobado por su control | `ci:internal/app/skills_test.go:TestSkillsDelRepositorio`, en make ci (verde) |
| FR-007 | T011(hecha) T018(hecha) | — | comprobado por su control | `ci:internal/app/skills_test.go:TestSkillsDelRepositorio`, en make ci (verde) |
| FR-008 | T011(hecha) T018(hecha) | — | comprobado por su control | `ci:internal/app/skills_test.go:TestSkillsDelRepositorio`, en make ci (verde) |
| FR-009 | T011(hecha) T018(hecha) | — | comprobado por su control | `ci:internal/app/skills_test.go:TestSkillsDelRepositorio`, en make ci (verde) |
| FR-010 | T010(hecha) | — | comprobado por su control | `ci:internal/app/skills_test.go:TestSkillsDelRepositorio`, en make ci (verde) |
| FR-011 | T009(hecha) T010(hecha) T011(hecha) | — | comprobado por su control | `ci:internal/app/skills_test.go:TestSkillsDelRepositorio`, en make ci (verde); `ci:internal/app/skills_test.go:TestSkillsDelRepositorio`, en make ci (verde) |
| FR-012 | T001(hecha) T018(hecha) | — | comprobado por su control | `ci:internal/app/skills_test.go:TestSkillsDelRepositorio`, en make ci (verde) |
| FR-013 | T001(hecha) T005(hecha) T006(hecha) T007(hecha) | — | tareas hechas | — |
| FR-014 | T002(hecha) | — | comprobado por su control | `ci:internal/evals/dependencias_test.go:TestConjuntoDeCadaEval`, en make ci (verde); `ci:internal/evals/eleccion_test.go:TestEleccionDeLosMutantes`, en make ci (verde) |
| FR-015 | T005(hecha) T006(hecha) T007(hecha) | — | comprobado por su control | `ci:skills_test.go:TestSkillsEmpotradas`, en make ci (verde) |
| FR-016 | T003(hecha) T004(hecha) | — | comprobado por su control | `ci:internal/evals/conjunto_test.go:TestControlesSobreLasReferenciasAMano`, en make ci (verde); `ci:internal/evals/conjunto_test.go:TestEvalsDelRepositorio`, en make ci (verde); `ci:internal/app/skills_test.go:TestSkillsDelRepositorio`, en make ci (verde) |
| FR-017 | T003(hecha) T004(hecha) | — | comprobado por su control | `ci:internal/evals/conjunto_test.go:TestControlesSobreLasReferenciasAMano`, en make ci (verde); `ci:internal/evals/conjunto_test.go:TestEvalsDelRepositorio`, en make ci (verde) |
| FR-018 | T003(hecha) | — | comprobado por su control | `ci:internal/evals/conjunto_test.go:TestControlesSobreLasReferenciasAMano`, en make ci (verde); `ci:internal/evals/conjunto_test.go:TestEvalsDelRepositorio`, en make ci (verde) |
| FR-019 | T007(hecha) T008(hecha) | — | comprobado por su control | `ci:internal/evals/conjunto_test.go:TestEvalsDelRepositorio`, en make ci (verde); `ci:internal/evals/conjunto_test.go:TestOrdenesDeEjemplo`, en make ci (verde) |
| FR-020 | T005(hecha) T006(hecha) T007(hecha) | — | tareas hechas | — |
| FR-021 | T005(hecha) T006(hecha) T007(hecha) | — | tareas hechas | — |
| FR-022 | T005(hecha) T006(hecha) T007(hecha) | — | tareas hechas | — |
| FR-023 | T005(hecha) T006(hecha) T007(hecha) T011(hecha) T017(hecha) T018(hecha) | — | tareas hechas | — |
| FR-024 | T005(hecha) T006(hecha) T007(hecha) | — | tareas hechas | — |
| FR-025 | T012(hecha) | — | comprobado por su control | `ci:internal/evals/descripciones_test.go:TestJuezDeLasDescripciones`, en make ci (verde) |
| FR-026 | T012(hecha) T015(hecha) | — | comprobado por su control | `ci:internal/evals/medida_test.go:TestMedidaVersionada`, en make ci (verde); `ci:internal/evals/ejecucion_test.go:TestEjecucionSinMedir`, en make ci (verde) |
| FR-027 | T012(hecha) | — | comprobado por su control | `ci:internal/evals/descripciones_test.go:TestJuezDeLasDescripciones`, en make ci (verde) |
| FR-028 | T013(hecha) | — | comprobado por su control | `ci:internal/evals/descripciones_test.go:TestJuezDeLasDescripciones`, en make ci (verde) |
| FR-029 | T013(hecha) | — | comprobado por su control | `ci:internal/evals/descripciones_test.go:TestJuezDeLasDescripciones`, en make ci (verde) |
| FR-030 | T002(hecha) T014(hecha) T016(hecha) | — | comprobado por su control | `ci:internal/evals/eleccion_test.go:TestEleccionDeLosMutantes`, en make ci (verde); `ci:internal/evals/decision_test.go:TestDecisionDelJob`, en make ci (verde); `ci:internal/evals/reutilizacion_test.go:TestBusquedaDeMedidasDeMain`, en make ci (verde); `ci:internal/evals/definicion_test.go:TestDefinicionDelJob`, en make ci (verde) |
| FR-031 | T015(hecha) | — | comprobado por su control | `ci:internal/evals/informe_test.go:TestInformeConLaDescripcion`, en make ci (verde); `evals:boe-legislacion:habla_en_primera_o_segunda_persona`: 0 de 1 (0,0 %), ≤ 0,0 %; `evals:legal-core:habla_en_primera_o_segunda_persona`: 0 de 1 (0,0 %), ≤ 0,0 %; `evals:jurisprudencia:habla_en_primera_o_segunda_persona`: 0 de 1 (0,0 %), ≤ 0,0 % |
| FR-032 | T013(hecha) T015(hecha) | — | comprobado por su control | `ci:internal/evals/descripciones_test.go:TestJuezDeLasDescripciones`, en make ci (verde); `ci:internal/evals/informe_test.go:TestInformeConLaDescripcion`, en make ci (verde) |
| FR-033 | T012(hecha) T013(hecha) T015(hecha) T017(hecha) | — | comprobado por su control | `ci:internal/evals/descripciones_test.go:TestJuezDeLasDescripciones`, en make ci (verde) |
| FR-034 | T017(hecha) | — | tareas hechas | — |
| SC-001 | T018(hecha) | — | comprobado por su control | `ci:Makefile:ci`, en make ci (verde) |
| SC-002 | T009(hecha) T010(hecha) T011(hecha) | — | comprobado por su control | `ci:internal/app/skills_test.go:TestSkillsDelRepositorio`, en make ci (verde); `ci:internal/app/skills_test.go:TestSkillsDelRepositorio`, en make ci (verde); `ci:internal/app/skills_test.go:TestSkillsDelRepositorio`, en make ci (verde) |
| SC-003 | T001(hecha) | — | comprobado por su control | `ci:internal/app/skills_test.go:TestSkillsDelRepositorio`, en make ci (verde) |
| SC-004 | T002(hecha) | — | comprobado por su control | `ci:internal/evals/dependencias_test.go:TestConjuntoDeCadaEval`, en make ci (verde); `ci:internal/evals/eleccion_test.go:TestEleccionDeLosMutantes`, en make ci (verde) |
| SC-005 | T003(hecha) T004(hecha) T008(hecha) | — | comprobado por su control | `ci:internal/evals/conjunto_test.go:TestControlesSobreLasReferenciasAMano`, en make ci (verde); `ci:internal/app/skills_test.go:TestSkillsDelRepositorio`, en make ci (verde); `ci:internal/evals/conjunto_test.go:TestEvalsDelRepositorio`, en make ci (verde) |
| SC-006 | T012(hecha) | — | comprobado por su control | `ci:internal/evals/descripciones_test.go:TestJuezDeLasDescripciones`, en make ci (verde) |
| SC-007 | T013(hecha) T015(hecha) | — | comprobado por su control | `ci:internal/evals/descripciones_test.go:TestJuezDeLasDescripciones`, en make ci (verde) |
| SC-008 | T012(hecha) T015(hecha) | — | comprobado por su control | `ci:internal/evals/medida_test.go:TestMedidaVersionada`, en make ci (verde); `ci:internal/evals/ejecucion_test.go:TestEjecucionSinMedir`, en make ci (verde) |
| SC-009 | T005(hecha) T006(hecha) T007(hecha) T009(hecha) T010(hecha) T011(hecha) | — | comprobado por su control | `ci:internal/app/skills_test.go:TestSkillsDelRepositorio`, en make ci (verde); `ci:internal/app/skills_test.go:TestSkillsDelRepositorio`, en make ci (verde) |
| SC-010 | T002(hecha) T014(hecha) T015(hecha) | — | comprobado por su control | `ci:internal/evals/eleccion_test.go:TestEleccionDeLosMutantes`, en make ci (verde); `ci:internal/evals/decision_test.go:TestDecisionDelJob`, en make ci (verde); `evals:boe-legislacion:habla_en_primera_o_segunda_persona`: 0 de 1 (0,0 %), ≤ 0,0 %; `evals:legal-core:habla_en_primera_o_segunda_persona`: 0 de 1 (0,0 %), ≤ 0,0 %; `evals:jurisprudencia:habla_en_primera_o_segunda_persona`: 0 de 1 (0,0 %), ≤ 0,0 %; `evals:boe-legislacion:afirma_lo_no_leido:claude-sonnet-5-5`: 1 de 120 (0,8 %), ≤ 2,8 %; `evals:jurisprudencia:afirma_lo_no_leido:claude-sonnet-5-5`: 0 de 60 (0,0 %), ≤ 1,7 % |
| SC-011 | T017(hecha) | — | tareas hechas | — |
| SC-012 | T005(hecha) T006(hecha) T007(hecha) | — | tareas hechas | — |
| SC-013 | T018(hecha) | — | tareas hechas | — |
| SC-014 | T012(hecha) T013(hecha) T015(hecha) T017(hecha) | — | comprobado por su control | `ci:internal/evals/descripciones_test.go:TestJuezDeLasDescripciones`, en make ci (verde) |
| SC-015 | T016(hecha) T017(hecha) | — | comprobado por su control | `ci:internal/evals/definicion_test.go:TestDefinicionDelJob`, en make ci (verde) |

## 7. Cambios posteriores a la revisión final

Cada commit posterior a lo que juzgó la primera ronda de la revisión final (`22d7962`), con la ronda que lo vio y su veredicto, y lo que toca fuera de `gates/` (ADR 0030):

- `3b987fd` fix(H31): motivos de la revisión final: lo vio la ronda 2 (juez A: aprobado; juez B: aprobado). Toca `CHANGELOG.md`, `CONTRIBUTING.md`, `docs/SKILLS.md`, `skills/boe-legislacion/SKILL.md`, `skills/boe-legislacion/references/lecturas_que_fallan.md`, `skills/legal-core/SKILL.md`, `contracts/skills-reescritas.md`, `plan.md`, `research.md`.
- `5aa1953` docs(H31): veredictos de la revisión final: solo registros de `gates/`.
- `03cf8d9` docs(H31): registros del run: solo registros de `gates/`.

### Cambios que ningún juez vio

Ninguno.

## 8. Cómo comprobarlo y consumo

- Escenarios manuales: `specs/027-h31-las-skills-siguen/quickstart.md`. Suite de aceptación congelada: `specs/027-h31-las-skills-siguen/aceptacion/` (activada en `internal/app/testdata/script/`).
- Run `eed49d22`: 14 h 31 min de reloj. Coste por paso y por rol: `scripts/coste-run.sh eed49d22`.
