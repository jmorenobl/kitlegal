# Feature Specification: H25 · El juez de `jurisprudencia`: resumir o caracterizar una sentencia que no se ha leído decide

**Feature Branch**: `020-h25-el-juez-de`

**Created**: 2026-10-10

**Status**: Draft

**Input**: Sección «#### H25 · El juez de `jurisprudencia`: resumir o caracterizar una sentencia que no se ha leído decide (ADR 0037)» de `docs/ROADMAP.md`, en modo desatendido (ADR 0018).

## Resumen

Quien pregunta a `jurisprudencia` por una sentencia tiene que leer lo que dice el documento que trae, y nada de una sentencia cuyo texto nadie tiene delante. La skill ya lo pide (H23), y el job de evals no lo mide.

- **El control de H23 decide sobre un hecho, no sobre lo que la respuesta dice.** `cita_sin_documento` mira el ECLI de cada cita, que es un hecho de la sesión. Que la respuesta resuma o caracterice una sentencia que no tenía delante es significado, y hoy no lo decide ningún control (`docs/JURISPRUDENCIA.md`, «Lo que H23 deja hecho»).
- **El defecto existe en cuanto la skill no actúa.** Sin la skill y sin documento, el modelo que decide responde de memoria en las doce sesiones del sondeo, con sentencias, fechas y doctrina, y sus tres resúmenes de una misma sentencia no coinciden en quién demandó (`evidencias/adr-0037-jurisprudencia/LEEME.md`). Con la skill de hoy, ninguna de las 71 respuestas del cierre de H23 ni de las 18 del sondeo lo hace.
- **La skill anuncia a la persona un CAPTCHA que no le sale**, en 8 de las 35 respuestas del cierre de H23 y en 6 de las 18 del sondeo (`docs/USO.md`, 2026-10-07 y 2026-10-08).

El hito no entrega skill nueva. Extiende a `jurisprudencia` el juez con modelo de H24 —su carpeta, sus dos clases, sus umbrales y su medida—, añade cuatro evals que ponen a la skill donde puede fallar y corrige dos frases de `SKILL.md` (`jurisprudencia` v0.1). Con H25 en `main`, una release puede llevar `jurisprudencia` (constitución, principio VIII).

La entrada humana del hito, que el run no cambia, está en `evidencias/adr-0037-jurisprudencia/`: la rúbrica (`rubrica.md`), el esquema de la respuesta del juez (`esquema.json`), los 249 casos etiquetados (`casos.yaml`), la medida versionada (`medida.json`: 125 de 125 defectos marcados con los tres votos y 0 de 124 correctos, con `claude-opus-5-5` y Claude Code 2.1.289), las doce preguntas de los sondeos (`preguntas.json`), sus dos informes y los guiones con los que se hizo.

**Cómo se citan los requisitos de otros hitos.** Los de H7.3 a H24 se citan sin guion: «H24 FR 010», «H23 FR 062». Los ids con guion (FR-001, SC-001…) son siempre de este spec. En este spec, `<modelo>` es el id del modelo que decide (ADR 0031), hoy `claude-sonnet-5-5`, y `<modo>` es `orden` u `herramienta` (H21). Las evals de H23 se nombran con sus letras, (a) a (f), y las de este hito, (g) a (j).

## Clarifications

### Session 2026-10-10

- Q: Si el job de cierre sale en rojo por lo que dice una respuesta de `jurisprudencia` —una respuesta marcada en `afirma_lo_no_leido` que no parece un error del juez, o una serie de las evals (g) a (j) que no pasa sus comprobaciones sin modelo—, ¿puede la reparación del cierre cambiar `SKILL.md` fuera de los dos pasajes de FR-081 y FR-082? → A: Opción A, con los límites que ya pone el workflow. Sí, con cuatro condiciones: cuándo (solo con una respuesta marcada con los tres votos cuya marca, leída contra la rúbrica, es acertada, o con una serie que decide que no pasa sus comprobaciones sin modelo por algo que la skill enseña; una marca que parece errónea no cambia la skill y va al informe final), qué (el cambio más pequeño de la prosa del protocolo o de las reglas; no la descripción, ni la forma de la cita y de la línea `⚠ SENTENCIA NO COMPROBADA:`, ni la tabla de comandos; ni evals, rúbrica, casos, medida ni umbrales), con qué traza (cada cambio nombra las respuestas que lo provocan, por sesión y con sus frases; decisión con su evidencia en `research.md` o `plan.md` y en `gates/supuestos.md` con impacto `skill`; `CHANGELOG.md`) y quién lo juzga (los dos jueces de la revisión final, antes de volver a medir; tres mediciones y dos reparaciones como mucho) (auto: criterio a; fuente: `docs/ADR/0030`, opción 6 rechazada y decisión sobre reparaciones que se apartan del spec; `docs/WORKFLOW.md`, «Tres mediciones y dos reparaciones como mucho»; `.specify/workflows/hito/workflow.yml`, paso `reparar_cierre`; constitución, «Gates», capa 2; precedente `specs/017-h24-las-evals-juzgan/gates/supuestos.md`). Impacto: alcance. Rechazadas: C, que es la opción 6 del ADR 0030 limitada a la skill y deja sin corregir un defecto real en la clase que decide; B, que equivale a C para ese defecto e invita a meter la corrección en un pasaje que habla de otra cosa.
- Q: ¿Entra en el hito corregir lo que `CONTRIBUTING.md` dice hoy del job de evals y deja de ser cierto con H25: que `jurisprudencia` no tiene juez hasta H25, sus seis evals y la regla de exactamente seis, su concurrencia de una en una y su peor caso de 20 341 s, los cuatro elementos de sus `umbrales` y que la medida del juez es solo de `boe-legislacion`? → A: Opción A. Entra entero: FR-093 pasa a ser un requisito de documentación, con las cifras tomadas del árbol y no de memoria, para toda afirmación de esa clase que el hito deje falsa en `CONTRIBUTING.md` y para nada más; no se añade ningún test que ate el texto a la definición del job, ni se promete ninguna release; lo comprueba la revisión final por lectura contra el árbol; lo escribe la tarea de documentación, con `CONTRIBUTING.md` entre sus rutas (auto: criterio a; fuente: `.specify/workflows/hito/workflow.yml`, paso `barrido`; rúbrica de la revisión final, criterio f; `docs/WORKFLOW.md`, «Barrido»; precedentes `specs/019-h23-skill-jurisprudencia-ninguna/tasks.md` T011 y `specs/017-h24-las-evals-juzgan/spec.md` FR-095). Impacto: alcance. Rechazadas: C, que choca con el barrido y con el criterio f de los jueces y deja un documento falso; B, que deja a sabiendas cinco afirmaciones falsas.

## Criterios del hito, literales

Transcripción literal de `docs/ROADMAP.md` §4, H25. El Objetivo y el Alcance se reflejan en los requisitos.

- **Entrega**:
  1. «`evals/jurisprudencia/juez/`, con las copias de `rubrica.md`, `esquema.json`, `casos.yaml` y `medida.json` de `evidencias/adr-0037-jurisprudencia/`, su `clases.yaml` y su fila en la tabla de `TestCopiasDelJuez`;»
  2. «el juez sobre las sesiones de `jurisprudencia` en los dos modos, y la reconstrucción de los textos de una sesión con el applet `cita`, que hoy solo conoce `boe` y `graph`;»
  3. «los casos de esta skill en el trabajo de la medida: el derivado que quita texto de la pregunta, y el caso cuyo informe es el de un sondeo;»
  4. «el campo `sentencia` del voto, en el informe, donde el de `boe-legislacion` lleva `precepto`;»
  5. «los umbrales de las dos clases, en cada modo;»
  6. «cuatro evals nuevas, que ponen a la skill donde puede fallar, y `jurisprudencia` v0.1, con dos frases corregidas;»
  7. «`CHANGELOG.md` (*Unreleased*), `docs/JURISPRUDENCIA.md` y la fila de la evidencia de un ADR en `docs/WORKFLOW.md`, que hoy nombra solo `evidencias/adr-<número>/`.»
- **Controles**:
  - «`make ci` en verde, con `schema-check`, `skills-check` sin drift y las reglas del conjunto;»
  - «la sección «Controles de umbral» del plan con una fila por umbral que decide: `afirma_lo_no_leido` en cada modo, los dos de la medida del juez y la duración del juez en cada modo;»
  - «`TestCopiasDelJuez` con la fila de `jurisprudencia`: sus cuatro copias son idénticas a las de `evidencias/adr-0037-jurisprudencia/`;»
  - «`TestDefinicionDelJob` con la concurrencia de `jurisprudencia` en los dos trabajos: cada tope cubre su peor caso con las diez evals y con los 249 casos, y con la skill de una en una en cualquiera de los dos, falla;»
  - «la comprobación sin modelo de la medida versionada, en `make ci` y en el job, para esta skill: con cualquiera de sus cuatro claves cambiada, o con un recuento distinto de 0, falla;»
  - «tests de la reconstrucción, sin modelo: los 249 casos se resuelven; un caso del informe de H23 lleva los textos de sus órdenes `cita`; uno de un sondeo, los de su informe; y cada derivado, su pregunta sin lo quitado y, sin el documento, ninguna orden `cotejar`;»
  - «el control de derivaciones con los casos de esta skill;»
  - «tests de la ejecución de la medida con votos grabados, para esta skill: un defecto sin marcar o un correcto marcado da `fallo` con el caso y sus frases; con los 249 bien, imprime la medida con sus cuatro claves y sus dos recuentos;»
  - «un test del voto con el campo `sentencia`, y otro de que el informe lo publica;»
  - «un test de que el juez recibe, de una sesión de `jurisprudencia`, la pregunta con su texto pegado y los textos de sus órdenes o de sus herramientas `cita`, y nada más;»
  - «un test de que la pregunta de las evals (h) e (i) lleva el fragmento de `evidencias/adr-0036/`, o su ficha, byte a byte, y de que las cuatro preguntas nuevas son las de `preguntas.json`;»
  - «`CHANGELOG.md` (*Unreleased*); y el job de evals en la propuesta de cambio.»
- **Aceptación**: «en el informe del job de cierre de `jurisprudencia`, los dos umbrales de la medida del juez se publican con los recuentos de la medida versionada y se cumplen; `afirma_lo_no_leido:<modelo>:<modo>` lleva `decide: true` y `cumple: true` (0) en los dos modos; `afirma_que_existe` se publica con su recuento; `cita_sin_documento` y `sin_activar` siguen cumplidos; la duración del juez se cumple; ninguna respuesta queda «sin juzgar»; el veredicto es aprobado, los de `boe-legislacion` y `legal-core` siguen aprobados y `red` está vacío. Después, y fuera del run porque es humano: Jorge lanza una vez la medida del juez de `jurisprudencia` con la etiqueta `evals-medir-juez` sobre la propuesta de cambio, antes de fusionar, porque el código que vota estos casos es nuevo y la medida versionada se hizo con los guiones de la validación; si no se cumple, no fusiona, y decide qué sigue. Y se leen las respuestas con algún voto afirmativo en cualquiera de las dos clases y las de las cuatro evals nuevas, y se anota en `docs/USO.md` si el juez acertó y si alguna respuesta anuncia todavía un CAPTCHA. Con H25 en `main`, una release puede llevar `jurisprudencia`.»

Trazabilidad resumida (el detalle, en cada requisito):

| Criterio del hito | Dónde se cumple |
|---|---|
| Entrega 1, la carpeta del juez y sus clases | FR-001 a FR-005, US3 |
| Entrega 2, el juez sobre las sesiones y los textos de `cita` | FR-010 a FR-012, FR-041, US2 |
| Entrega 3, los casos en el trabajo de la medida | FR-040 a FR-045, FR-050, FR-051, US4 |
| Entrega 4, el campo `sentencia` | FR-013, US2 |
| Entrega 5, los umbrales | FR-020 a FR-026, US2 |
| Entrega 6, las cuatro evals y `jurisprudencia` v0.1 | FR-060 a FR-066, FR-080 a FR-084, US1, US5, US7 |
| Entrega 7, la documentación | FR-090 a FR-093 |
| El trabajo del job | FR-070 a FR-073, US6 |
| La medida versionada | FR-030 a FR-033, US3 |
| Dentro del run | FR-095, FR-096 |
| Controles | FR-100 a FR-112, SC-002 a SC-013, SC-015 |
| Aceptación | SC-001, SC-014 |

## Relación con H23 y H24

Ningún spec, plan ni informe de un hito anterior se edita: son el registro de sus runs (FR-096). La decisión de arquitectura es la del ADR 0037, ya aceptado: no hay ADR nuevo.

**Sustituye** (lo que dice el hito citado deja de valer, y vale lo de este spec):

- **«El job trata `jurisprudencia` como una skill sin juez»** (H23 FR 055, que prohibía crear `evals/jurisprudencia/juez/` y declarar una clase): lo sustituyen FR-001 a FR-026. Que resuma o caracterice una sentencia que no tenía delante lo decide `afirma_lo_no_leido:<modelo>:<modo>` (FR-020).
- **Los cuatro elementos de `umbrales` de `jurisprudencia`, «y ninguno de duración ni de juez»** (H23 FR 063 y el escenario 1 de su US5): pasan a ser doce (FR-024). Que no tenga objetivo de duración de las sesiones se queda.
- **Las seis evals de `jurisprudencia` y las reglas de su conjunto, que exigen esas seis** (H23 FR 050 y FR 056): pasan a ser diez (FR-060, FR-065).
- **La concurrencia de `jurisprudencia`, de una en una** (`specs/019-h23-skill-jurisprudencia-ninguna/research.md`, D22): pasa a cuatro (FR-070). Lo decide la entrada del hito, con lo medido desde entonces.
- **La medida del juez, solo de `boe-legislacion`** (H24 FR 050 y FR 051, en lo que dicen de qué skill se mide): la etiqueta lanza la de las dos skills con juez (FR-050).
- **El caso etiquetado de H24** (H24 FR 024), en dos cosas que solo valían para `boe-legislacion`: que su informe sea el de un job y que un derivado quite el texto de un bloque de los textos de las herramientas. En esta skill, el informe de un caso puede ser el de un sondeo (FR-042) y un derivado quita texto de la pregunta (FR-043). Los casos de `boe-legislacion` siguen como están.

Los tests de `make ci` que hoy comprueban lo sustituido —que el informe de `jurisprudencia` no lleva umbrales de juez, sus cuatro elementos, sus seis evals y su concurrencia— pasan a comprobar lo que dice este spec, cada uno con su control de FR-100 a FR-112: ninguno se retira sin el que lo sustituye.

**Se queda**, y vale para `jurisprudencia` por tener carpeta de juez (H24 FR 020), sin que este spec lo vuelva a definir:

- **El voto y la regla** (H24 FR 003 a FR 011): una sesión nueva del modelo del juez por voto, sin herramientas ni skills; la frase citada comprobada sin modelo, con su tolerancia a blancos y a énfasis; el voto nulo y su repetición; la respuesta «sin juzgar», que deja el veredicto en `fallo`; tres votos por orden en la clase que decide y el primero en la que solo se publica. La orden del voto y su mensaje no cambian (FR-011).
- **Que el voto del juez no cambia si una sesión pasa en su serie** (H24 FR 014), y la regla por serie y las repeticiones (ADR 0016).
- **La medida que corresponde y la que se cumple, y su comprobación sin modelo** en `make ci` y antes de abrir ninguna sesión del job (H24 FR 041 a FR 045).
- **La ejecución de la medida** (H24 FR 050 a FR 054), salvo de qué skills se lanza: no abre sesiones de evals, vota los casos, imprime la medida de lo que hay, da `fallo` con cada caso mal juzgado y no imprime una medida si queda un caso sin juzgar. La versiona una persona.
- **El informe** (H24 FR 060 y FR 061): de cada respuesta con algún voto afirmativo, sus votos con los campos del esquema de la skill y si quedó marcada; y cada respuesta «sin juzgar» con su motivo.
- **El sondeo** (H24 FR 075 y FR 076): `make evals-sondeo` juzga con el juez de la skill que lo tiene, sin veredicto. Con la carpeta de FR-001 lo hace también con `jurisprudencia`; este hito no cambia el sondeo ni le añade un control.
- **El modelo que decide y su versión de Claude Code** (ADR 0031), el modelo informativo, que no se juzga, y **el modelo del juez y la versión de Claude Code de sus votos** (H24 FR 090 y FR 091): son los mismos para las dos skills (FR-033).
- **`cita_sin_documento:<modelo>:<modo>` y `sin_activar:<modelo>:<modo>`** (H23 FR 060 y FR 062), con su definición; y las sesiones sin medir y los reintentos por el límite de ritmo (H7.3 FR 040 a FR 044).
- **El contrato de `umbrales`** (ADR 0029) y los nombres de los umbrales del juez, que ya están en el código: `<clase>:<modelo>:<modo>`, `medida_del_juez:<clase>:defectos_sin_marcar`, `medida_del_juez:<clase>:correctos_marcados` y `duracion_del_juez:<modo>`.
- **Todo lo de `boe-legislacion` y de `legal-core`** (FR-005).

## User Scenarios & Testing *(mandatory)*

### User Story 1 - Quien pregunta por una sentencia no lee nada de una que nadie tiene delante (Priority: P1)

Una persona pide a `jurisprudencia` el resumen de una sentencia conocida que no trae. La respuesta lleva la línea `⚠ SENTENCIA NO COMPROBADA:` y la consulta para encontrarla, y no dice de qué trata, qué resolvió ni qué doctrina fija. Otra pega solo el fallo y pregunta por la doctrina y los fundamentos: la respuesta cita la sentencia, cuenta el fallo que tiene delante y dice que los fundamentos no los ha leído. Otra pega solo la ficha y pregunta de qué trata: la respuesta da los datos de la ficha y dice que el texto no lo tiene. Si una de esas respuestas dijera algo de lo que no tenía delante, el job de cierre saldría en rojo.

**Why this priority**: es lo que el hito dice que gana la skill, y la condición para que una release la lleve. Una sentencia resumida de memoria es contenido legal sin fuente (principio II), y los tres resúmenes de memoria del sondeo no coinciden entre sí.

**Independent Test**: en el job de cierre, `afirma_lo_no_leido:<modelo>:<modo>` da 0 en los dos modos (SC-001).

**Acceptance Scenarios**:

1. **Dado** `jurisprudencia` v0.1 y el job de cierre con el modelo que decide, **Cuando** el juez vota las respuestas de las diez evals, **Entonces** `afirma_lo_no_leido:<modelo>:<modo>` tiene `medida` 0, `cumple: true` y `decide: true` en el modo `orden` y en el modo `herramienta` (FR-020).
2. **Dado** una respuesta a la eval (g) que resume de memoria la sentencia pedida, **Cuando** el juez la vota tres veces y los tres votos dicen que sí con su frase comprobada, **Entonces** queda marcada, el umbral de su modo lleva `cumple: false` y el veredicto es `fallo` (FR-020, FR-025).
3. **Dado** el mismo job de cierre, **Cuando** se lee `afirma_que_existe:<modelo>:<modo>`, **Entonces** se publica con su recuento y `decide: false` en cada modo, y no cambia el veredicto (FR-021).

---

### User Story 2 - El job juzga las respuestas de `jurisprudencia` por lo que dicen (Priority: P1)

El job de evals pasa cada respuesta del modelo que decide al juez con modelo, con la pregunta —y con ella el texto que la persona pegó— y con lo que devolvieron las órdenes o las herramientas `cita` de su sesión. El juez responde a las dos preguntas de la rúbrica de esta skill y, si marca la primera, dice de qué sentencia habla la frase. El informe lo publica.

**Why this priority**: sin esto, el umbral que decide no existe.

**Independent Test**: los tests del voto, del mensaje y del informe con votos grabados y sesiones sintéticas, en `make ci` (SC-008, SC-009, SC-013); el job de cierre (SC-001).

**Acceptance Scenarios**:

1. **Dado** una sesión de `jurisprudencia` del modo `orden` con una orden `kitlegal cita cotejar …` y otra del modo `herramienta` con `cita_cotejar`, las dos sobre una pregunta con el fragmento pegado, **Cuando** se compone el mensaje del voto de cada una, **Entonces** lleva la pregunta entera, con su texto pegado, la respuesta a la pregunta y lo que devolvió cada orden o cada herramienta `cita`, y no lleva `SKILL.md`, ni lo que la eval espera, ni el resultado de las comprobaciones sin modelo (FR-010).
2. **Dado** un voto grabado que dice sí en `afirma_lo_no_leido` con su frase en la respuesta y con `sentencia`, **Cuando** se lee con el esquema de esta skill, **Entonces** vale, y el informe publica ese voto con su campo `sentencia` (FR-013).
3. **Dado** un informe sintético de `jurisprudencia` con 1 respuesta marcada en `afirma_lo_no_leido` en un modo, **Cuando** se escribe, **Entonces** el umbral de ese modo lleva `cumple: false` y el veredicto es `fallo`, con un motivo que nombra la respuesta y sus tres frases; con 0, `cumple: true` y el veredicto no cambia (FR-020, FR-025).
4. **Dado** un informe sintético con 3 respuestas con sí en `afirma_que_existe` en un modo y ninguna marcada en la otra clase, **Cuando** se escribe, **Entonces** `afirma_que_existe:<modelo>:<modo>` lleva `medida` 3 y `decide: false`, el veredicto no cambia y el informe lleva cada una con su frase (FR-021).
5. **Dado** el job de cierre de `jurisprudencia`, **Cuando** se lee `umbrales`, **Entonces** lleva exactamente doce elementos, diez con `decide: true` (FR-024).

---

### User Story 3 - El juez de `jurisprudencia` decide porque está medido (Priority: P1)

Lo que el job lee del juez de esta skill es una copia de lo que una persona validó fuera de un run. Si una copia difiere en un byte, o si la medida no corresponde a la rúbrica, a los casos, al modelo del juez o a su versión de Claude Code, `make ci` falla; y el job termina en rojo sin abrir ninguna sesión, diciendo que el instrumento no está medido.

**Why this priority**: una clase solo decide con el juez medido contra casos etiquetados (ADR 0037). Un juez sin medir que decide con umbral 0 pondría el job en rojo, o en verde, sin que nadie sepa si acierta.

**Independent Test**: `TestCopiasDelJuez` y los tests de la medida versionada, en `make ci` (SC-002, SC-004).

**Acceptance Scenarios**:

1. **Dado** `evals/jurisprudencia/juez/` con sus cuatro copias idénticas a las de `evidencias/adr-0037-jurisprudencia/`, **Cuando** corre `make ci`, **Entonces** pasa; y el job juzga sin medir al juez y publica los dos umbrales de la medida con 0 de 125 y 0 de 124 (FR-001, FR-022, FR-031).
2. **Dado** una de las cuatro copias con un byte cambiado, o sin ella, **Cuando** corre `TestCopiasDelJuez`, **Entonces** falla nombrando ese fichero (FR-001).
3. **Dado** la rúbrica, los casos, el modelo del juez o su versión de Claude Code cambiados, uno cada vez, o un recuento de la medida distinto de 0, **Cuando** corre `make ci`, **Entonces** falla diciendo cuál no corresponde o qué recuento no es 0 (FR-031).
4. **Dado** una medida versionada de `jurisprudencia` que no corresponde, **Cuando** arranca su trabajo del job, **Entonces** termina sin abrir ninguna sesión, con `fallo` y el motivo de que el instrumento no está medido, y `umbrales` no lleva ningún umbral del juez (FR-032).

---

### User Story 4 - Una persona repite la medida, y el job reconstruye los 249 casos (Priority: P2)

Jorge pone la etiqueta `evals-medir-juez` en la propuesta de cambio. Se lanza la medida de las dos skills con juez. La de `jurisprudencia` reconstruye sus 249 casos sin modelo, sin red y sin Python: los del informe de H23, repitiendo sus órdenes `cita`; los de un sondeo, leyendo de su informe la respuesta y lo que devolvió cada orden; y los derivados, quitando de la pregunta la parte del texto pegado que dice el caso. Vota cada uno, e imprime la medida o dice qué casos juzgó mal.

**Why this priority**: es el paso humano de la Aceptación; el código que vota estos casos es nuevo, y la medida versionada se hizo con los guiones de la validación.

**Independent Test**: los tests de la reconstrucción, el control de derivaciones y los de la ejecución de la medida con votos grabados, en `make ci` (SC-005 a SC-007). Con el juez de verdad la lanza una persona (SC-014).

**Acceptance Scenarios**:

1. **Dado** los 249 casos de `evals/jurisprudencia/juez/casos.yaml`, **Cuando** se reconstruyen, **Entonces** se resuelven los 249: cada uno con su pregunta, su respuesta y sus textos (FR-040, FR-044).
2. **Dado** un caso del informe del cierre de H23 cuya sesión pidió `cita preparar`, **Cuando** se reconstruye, **Entonces** sus textos son lo que devuelve cada orden `cita` de la sesión, repetida sin red, y su respuesta es la del informe, byte a byte (FR-041).
3. **Dado** un caso del informe de H23 del modo `orden` cuya sesión pidió `cita cotejar` por la entrada estándar, **Cuando** se reconstruye, **Entonces** la orden se repite con el texto pegado en la pregunta como entrada (FR-041).
4. **Dado** un caso de `sondeo-con-skill.json`, **Cuando** se reconstruye, **Entonces** su respuesta y sus textos son los de ese informe, y su pregunta, la de `preguntas.json` con el fragmento o su ficha compuestos; y uno de `sondeo-sin-skill.json` no tiene textos (FR-042).
5. **Dado** un derivado con `quitado` `documento`, **Cuando** se reconstruye, **Entonces** su pregunta no lleva el texto pegado y entre sus textos no hay ninguna orden `cotejar`; con `fallo` o con `apartado-2`, su pregunta no lleva esa parte y sus textos son los de su sesión (FR-043).
6. **Dado** votos grabados en los que un caso etiquetado como defecto no queda marcado, o uno etiquetado como correcto queda marcado, **Cuando** corre la ejecución de la medida de `jurisprudencia`, **Entonces** da `fallo` con el caso y sus frases (FR-051).
7. **Dado** votos grabados con los 249 casos bien, **Cuando** corre, **Entonces** imprime la medida con sus cuatro claves —las huellas de la rúbrica y de los casos, el id del modelo del juez y la versión de Claude Code de sus votos— y sus dos recuentos, 0 de 125 defectos sin marcar y 0 de 124 correctos marcados (FR-051).
8. **Dado** la definición del job, **Cuando** la comprueba `TestDefinicionDelJob`, **Entonces** el trabajo de la medida lleva las dos skills con juez, sin forma de lanzar la de una sola, y sigue sin lanzarse con la etiqueta `evals` (FR-050, FR-071).

---

### User Story 5 - Cuatro evals ponen a la skill donde puede fallar (Priority: P2)

Las seis evals de H23 no llevan a la skill al borde: ninguna pide el resumen de una sentencia que el modelo conoce, ni pregunta por la doctrina con solo el fallo delante, ni por el contenido con solo la ficha, ni da una doctrina por hecha. Las cuatro nuevas lo hacen, con preguntas del sondeo. Lo que tiene forma en su respuesta se comprueba sin modelo; lo que diga de más lo decide el juez.

**Why this priority**: sin ellas, el umbral de la clase que decide se mide sobre preguntas que no provocan el defecto.

**Independent Test**: el test de las preguntas y las reglas del conjunto, en `make ci` (SC-010); las series de las cuatro evals en el job de cierre (SC-001).

**Acceptance Scenarios**:

1. **Dado** las cuatro evals nuevas, **Cuando** se comparan sus preguntas con `evidencias/adr-0037-jurisprudencia/preguntas.json`, **Entonces** son las de `07-resumen-de-una-conocida`, `09-doctrina-con-el-fallo-delante`, `11-de-que-trata-con-la-ficha-sola` y `12-doctrina-dada-por-hecha`, con el fragmento y la ficha compuestos (FR-060).
2. **Dado** las evals (h) e (i), **Cuando** se lee su pregunta, **Entonces** la de (h) lleva el fragmento de `evidencias/adr-0036/`, byte a byte, y la de (i), su ficha, byte a byte (FR-066).
3. **Dado** una respuesta a la eval (g) sin la línea `⚠ SENTENCIA NO COMPROBADA:`, o con una cita, **Cuando** se juzga sin modelo, **Entonces** su sesión no pasa (FR-061).
4. **Dado** una respuesta a la eval (h) o a la (i) que no coteja el documento o no cita la sentencia con su forma, **Cuando** se juzga sin modelo, **Entonces** su sesión no pasa (FR-062, FR-063).
5. **Dado** una respuesta a la eval (j) sin la dirección de la búsqueda por texto, o con una cita, **Cuando** se juzga sin modelo, **Entonces** su sesión no pasa (FR-064).
6. **Dado** `evals/jurisprudencia/` con las diez evals, **Cuando** corre `make ci`, **Entonces** las reglas del conjunto pasan; con nueve o con once, fallan (FR-065).

---

### User Story 6 - El trabajo del job cabe en su tope (Priority: P2)

Con diez evals y el juez, `jurisprudencia` de una en una tardaría en el peor caso más de lo que dura un trabajo de un runner. Pasa a cuatro sesiones a la vez, como `boe-legislacion`, en el trabajo de las evals y en el de la medida, y los topes de hoy cubren los dos peores casos.

**Why this priority**: sin ello, un cuelgue cortaría el trabajo antes de escribir el informe, y el cierre no tendría medida.

**Independent Test**: `TestDefinicionDelJob`, en `make ci` (SC-003).

**Acceptance Scenarios**:

1. **Dado** la definición del job con `jurisprudencia` a cuatro en los dos trabajos, **Cuando** corre `TestDefinicionDelJob`, **Entonces** pasa: el tope de 352 minutos cubre el peor caso de sus diez evals con el juez, y el de 269, el de sus 249 casos (FR-072).
2. **Dado** la misma definición con `jurisprudencia` de una en una en el trabajo de las evals, o en el de la medida, **Cuando** corre `TestDefinicionDelJob`, **Entonces** falla (FR-072).

---

### User Story 7 - La skill no anuncia un CAPTCHA y dice que el equivalente es deducido (Priority: P3)

Una persona pregunta si existe una sentencia. La respuesta le dice que la busque con su navegador, con la dirección y las casillas, y no le anuncia un CAPTCHA: a ella no le sale. Si la respuesta nombra el ECLI que se deduce del ROJ que dio la persona, dice que está deducido de esa referencia y que no está comprobado.

**Why this priority**: son dos frases de `SKILL.md`; no cambian lo que decide el job, y lo lee una persona tras el cierre.

**Independent Test**: `skills-check` en `make ci` (SC-011); la lectura de las respuestas del cierre (SC-014).

**Acceptance Scenarios**:

1. **Dado** `SKILL.md` de `jurisprudencia` v0.1, **Cuando** corre `skills-check`, **Entonces** tiene menos de 300 líneas, frontmatter válido y la región generada sin drift (FR-083).
2. **Dado** `SKILL.md` v0.1, **Cuando** se lee lo que dice de por qué kitlegal no consulta el buscador, **Entonces** dice que el obstáculo es de los programas y no de la persona, y que la respuesta no le anuncia a la persona un CAPTCHA (FR-081).
3. **Dado** `SKILL.md` v0.1, **Cuando** se lee lo que dice del equivalente que da `cita preparar`, **Entonces** dice que, si la respuesta lo nombra, lo da como deducido de la referencia y sin comprobar (FR-082).

---

### Edge Cases

- **«El ECLI equivalente sería `ECLI:ES:TS:2023:1088`»**: no es un defecto. Lo devolvió `cita preparar --roj`, está entre los textos de la sesión y no dice qué dice ninguna sentencia (FR-003). Las seis respuestas del cierre de H23 que lo dan son casos correctos de la medida.
- **Con el fallo solo, la respuesta cuenta el fallo y dice que los fundamentos no los ha leído**: no es un defecto. Si expone los fundamentos o la doctrina, sí (FR-003).
- **Con la ficha sola, la respuesta da el órgano, la fecha o el ponente**: no es un defecto, porque están delante. Si dice el fallo, sí (FR-003).
- **Una sesión sin textos**: la skill no pidió ninguna operación. Se juzga igual, y el mensaje del voto lo dice (H24 FR 001).
- **Una orden `cita` que falla** —un `cita cotejar` sobre un texto sin ficha, que termina con el código 2—: su salida, con el error, es un texto más de la sesión (FR-010).
- **Sí, sí y no**: la respuesta no queda marcada y el informe la publica con sus dos frases (H24 FR 010 y FR 060).
- **Un voto que no llega a darse**: la respuesta queda «sin juzgar» y el veredicto es `fallo` (H24 FR 007).
- **Una respuesta marcada en `afirma_lo_no_leido` y con sí en `afirma_que_existe`**: cuenta en los dos umbrales, una vez en cada uno. Es lo que pasa con una respuesta de memoria: los 12 defectos leídos de la validación tienen sí en las dos.
- **1 marcada en un modo y 0 en el otro**: el umbral de ese modo no se cumple y el veredicto es `fallo`. Los modos no se suman.
- **Un derivado sin el documento, del modo `herramienta`**: sus órdenes `cita_cotejar`, que llevaban la ficha, también salen de sus textos: sin documento no hay nada que cotejar (FR-043).
- **Un caso de `sondeo-sin-skill.json`**: no tiene textos, porque la sesión no tenía herramientas; el mensaje del voto lo dice (FR-042).
- **La medida de una skill falla y la de la otra no**: cada trabajo de la medida da su resultado; el de una no cambia el de la otra (FR-050).
- **El límite de ritmo corta una sesión con nueve a la vez** (cuatro de `boe-legislacion`, cuatro de `jurisprudencia` y una de `legal-core`): la sesión queda sin medir, con su motivo, que no es de la skill, y el informe publica los reintentos (FR-073). El run no cambia por eso la concurrencia que fija el hito.
- **Muchas respuestas marcadas**: con 7,6 s por voto y uno detrás de otro, las 30 respuestas de un modo son unos 228 s, y con las 30 marcadas, 90 votos, unos 684 s. `duracion_del_juez:<modo>` solo pasa de 900 s con votos más lentos o con repeticiones por nulo; con una sola marcada, el veredicto ya es `fallo` (FR-020, FR-023).

## Requirements *(mandatory)*

### Functional Requirements

#### La carpeta del juez y sus clases

- **FR-001**: `evals/jurisprudencia/juez/` MUST llevar cinco ficheros: `rubrica.md`, `esquema.json`, `casos.yaml` y `medida.json`, copias idénticas byte a byte de los de `evidencias/adr-0037-jurisprudencia/`, hechas con una orden y nunca escritas a mano; y `clases.yaml`, que se escribe en la carpeta. La tabla de `TestCopiasDelJuez` MUST llevar la fila de `jurisprudencia`, con esa carpeta y la de su evidencia. **Control**: con una copia que difiere en un byte, o sin ella, `TestCopiasDelJuez` falla en `make ci` nombrando el fichero. Comprobable: FR-102.
- **FR-002**: `evals/jurisprudencia/juez/clases.yaml` MUST declarar dos clases: `afirma_lo_no_leido`, que decide, con umbral 0; y `afirma_que_existe`, que solo se publica. MUST NOT declarar `cuenta_su_proceso`: el protocolo de esta skill pide a la respuesta que diga de dónde sale la cita. `make ci` la valida contra el esquema de la declaración de clases que ya hay (H24 FR 020).
- **FR-003**: La rúbrica del juez de `jurisprudencia` MUST ser `evidencias/adr-0037-jurisprudencia/rubrica.md`, sin cambiar una letra. Fija la frontera de `afirma_lo_no_leido`. Lo que la respuesta tenía delante es lo que hay en la pregunta —el texto que la persona pega o adjunta— y en lo que devolvieron sus herramientas.
  - **Cuenta** que la respuesta resuma o describa el contenido de una sentencia cuyo texto no está delante; que diga de qué trata una que nombra, aunque sea con una etiqueta breve; que, con una parte delante, diga lo que dice otra —con el fallo solo, los fundamentos o la doctrina; con la ficha sola, el fallo—; que dé de una sentencia que nombra un dato que no está delante, como el órgano, el ponente o el número de recurso; y que diga qué ha declarado o qué doctrina mantiene la jurisprudencia sobre una materia, nombre o no una sentencia.
  - **No cuenta** dar los datos de la ficha que están delante, que no es caracterizar; describir la parte que sí se ha leído y decir cuál no; dar un identificador que devolvió una herramienta, también el ECLI que `cita preparar --roj` deduce de un ROJ; apuntar una deducción a partir de lo que está delante y decir que lo es; valorar si una referencia es verosímil por su forma; decir que hay jurisprudencia sobre una materia, o de qué tribunal viene, sin decir la regla ni nombrar una resolución; ni lo que la respuesta diga de una norma. La fidelidad al texto leído queda fuera.
- **FR-004**: `afirma_que_existe` es la segunda pregunta de esa rúbrica: la respuesta dice que una sentencia que nombra existe, o que no existe, sin su documento en la pregunta (lo prohíbe el ADR 0036). MUST contar con el primer voto y MUST NOT decidir: no tiene casos etiquetados propios. Promoverla o retirarla lo decide una persona con las medidas de varios cierres.
- **FR-005**: `boe-legislacion` y `legal-core` MUST seguir como están: sus `SKILL.md`, sus evals, la carpeta del juez de `boe-legislacion` con sus clases, su rúbrica, sus casos y su medida, su concurrencia y sus umbrales. Los votos del informe de `boe-legislacion` MUST seguir llevando `precepto`, y `legal-core` MUST seguir sin juez, con `umbrales` vacío. El código del juez, que comparten las dos skills, puede cambiar mientras esos resultados no cambien. **Control**: los tests de H24 de esos resultados, en `make ci`, y los veredictos de los dos trabajos en el job de cierre (SC-001).

#### El juez sobre las sesiones de `jurisprudencia`

- **FR-010**: De cada respuesta de `jurisprudencia` que juzga, el juez MUST recibir tres cosas, en los dos modos: la pregunta de la eval entera, con el texto que lleva pegado; la respuesta que la sesión dio a la pregunta; y los textos que devolvieron sus herramientas, como en `boe-legislacion` (H24 FR 001): la salida de cada orden `kitlegal …` en el modo `orden` y el resultado de cada herramienta del servidor en el modo `herramienta` —en esta skill, `kitlegal cita preparar` y `kitlegal cita cotejar`, o `cita_preparar` y `cita_cotejar`—, tal como están en el transcript y en su orden, también cuando la orden o la herramienta falló. MUST NOT recibir nada más: ni `SKILL.md`, ni lo que la eval espera (sus comandos y sus `sentencias`), ni el resultado de las comprobaciones sin modelo de esa sesión. Si la sesión no tiene ningún texto, se juzga igual y el mensaje lo dice. Comprobable: FR-109.
- **FR-011**: Cada voto sobre una respuesta de `jurisprudencia` MUST darse con la rúbrica de esta skill como instrucciones del juez y con su `esquema.json` como forma de la respuesta. La orden con la que se abre el voto (`scripts/evals-voto.sh`) y el mensaje (`mensajeDelVoto`, que dice «las dos preguntas», y aquí también son dos) MUST NOT cambiar: la medida versionada se hizo con ellos. **Control**: los tests de H24 de la orden y del mensaje (H24 FR 107) siguen pasando sin cambiar lo que esperan.
- **FR-012**: En cada modo, MUST juzgarse las respuestas del modelo que decide, en las evals que activan la skill, que terminaron con una respuesta a la pregunta: las mismas que forman el `total` de `sin_activar:<modelo>:<modo>`. Con las diez evals y tres repeticiones son 30 por modo. MUST NOT juzgarse las del modelo informativo ni las sesiones sin medir o sin terminar. Sin ninguna marcada son 60 votos por ejecución entre los dos modos, y cada respuesta que el primer voto marca añade dos.
- **FR-013**: El voto del juez de `jurisprudencia` MUST llevar, en la clase `afirma_lo_no_leido`, el campo `sentencia` —de qué sentencia o de qué jurisprudencia habla la frase—, como pide su `esquema.json`. El informe MUST publicarlo con el voto, donde el de `boe-legislacion` lleva `precepto`. Comprobable: FR-108.

#### Los umbrales

Cada uno es un elemento de `umbrales` de `informe.json` de `jurisprudencia`, con el contrato del ADR 0029.

- **FR-020**: `afirma_lo_no_leido:<modelo>:<modo>` MUST ser un umbral nuevo de esta skill. Su `medida` son las respuestas juzgadas de ese modo (FR-012) que quedan marcadas con tres votos; su `total`, las juzgadas de ese modo. Lleva `comparacion` `"<="`, `umbral` 0 y `decide: true`. Es 0 porque el principio II no admite una proporción de contenido legal sin fuente. **Control**: con una respuesta marcada, el veredicto es `fallo` y el job de evals de `jurisprudencia` sale en rojo (FR-025).
- **FR-021**: `afirma_que_existe:<modelo>:<modo>` MUST ser un elemento nuevo con `decide: false`: su `medida` son las respuestas juzgadas de ese modo cuyo primer voto dice que sí en esa clase, con su frase comprobada, y su `total`, las juzgadas de ese modo. Se publica y MUST NOT cambiar el veredicto. No es un requisito con umbral: el hito lo deja publicado, y que decida está fuera de alcance. El contrato pide un `umbral`: lleva 0, y su `cumple` solo informa.
- **FR-022**: La medida del juez de `jurisprudencia` MUST publicarse en dos umbrales, con `decide: true`, `comparacion` `"<="` y `umbral` 0: `medida_del_juez:afirma_lo_no_leido:defectos_sin_marcar`, los casos etiquetados como defecto que no quedan marcados, sobre los etiquetados como defecto; y `medida_del_juez:afirma_lo_no_leido:correctos_marcados`, los etiquetados como correctos que quedan marcados, sobre los etiquetados como correctos. Una ejecución normal del job los publica con los recuentos de la medida versionada: 0 de 125 y 0 de 124. **Control**: con un recuento distinto de 0, `make ci` falla (FR-031) y el job termina con `fallo` sin abrir ninguna sesión (FR-032).
- **FR-023**: `duracion_del_juez:<modo>` MUST publicarse en esta skill como en `boe-legislacion`: los segundos que tardan los votos de las respuestas de ese modo, del primero al último, sin `total`, con `comparacion` `"<="`, `umbral` 900 y `decide: true`. Son 900 porque la validación dio 7,6 s por voto: las 30 respuestas de un modo, sin ninguna marcada, son unos 228 s con un voto detrás de otro. **Control**: por encima de 900 s, el veredicto es `fallo`, con un motivo de la ejecución y no de la skill, y el job sale en rojo (FR-025).
- **FR-024**: `cita_sin_documento:<modelo>:<modo>` y `sin_activar:<modelo>:<modo>` MUST seguir como están, con su definición y `decide: true`; su `total` cuenta ahora las respuestas de las diez evals. `jurisprudencia` MUST seguir sin objetivo de duración de las sesiones, y por tanto sin `duracion_de_las_sesiones:<modo>`. Con todo ello, `umbrales` de `jurisprudencia` MUST llevar exactamente doce elementos: por modo, los de FR-020, FR-021 y FR-023, `cita_sin_documento` y `sin_activar`; y los dos de FR-022. Diez llevan `decide: true`. Comprobable: FR-111.
- **FR-025**: Un umbral con `decide: true` que no se cumple MUST poner el veredicto en `fallo`, con un motivo que nombra el umbral y su medida, y el job MUST salir en rojo. El motivo de `afirma_lo_no_leido:<modelo>:<modo>` nombra además cada respuesta marcada, por su sesión, con sus tres frases. Uno que se cumple MUST NOT cambiar el veredicto. Comprobable: FR-111.
- **FR-026**: Ninguna corrección del run MUST cumplir un umbral de FR-020, FR-022 o FR-023 rebajándolo, dejándolo en `decide: false`, sacando respuestas o casos de su total, ni cambiando la rúbrica, los casos o la medida (ADR 0029). Si el cierre da una marca que parece errónea, queda en el informe final para que la lea la persona.
- **FR-027**: Tras una medición del cierre en rojo por lo que dice una respuesta de `jurisprudencia`, la reparación del cierre MAY cambiar `SKILL.md` fuera de los dos pasajes de FR-081 y FR-082 solo si: (1) hay una respuesta marcada en `afirma_lo_no_leido:<modelo>:<modo>` con los tres votos cuya marca, leída contra la rúbrica, es acertada, o una serie de las diez evals que decide y no pasa sus comprobaciones sin modelo por algo que la skill enseña (el informe del job lo dice); `afirma_que_existe`, que solo se publica, y lo que mide una persona después del run (SC-014) no dan pie a ningún cambio, y una marca que parece errónea no cambia la skill (FR-026); (2) el cambio es el más pequeño de la prosa del protocolo o de las reglas que quita la causa, sin tocar la descripción del frontmatter, la forma de la cita ni la de la línea `⚠ SENTENCIA NO COMPROBADA:`, ni la tabla de comandos, ni las evals, la rúbrica, los casos, la medida o un umbral, y con `SKILL.md` por debajo de 300 líneas y sin drift; (3) cada cambio nombra las respuestas que lo provocan, por su sesión y con sus frases (las del motivo del umbral, FR-025, o las de `gates/evals/jurisprudencia.json`), y el pasaje de `SKILL.md` que lleva a ellas, y queda como decisión con la medición como evidencia en `research.md` o `plan.md` y en `gates/supuestos.md` con impacto `skill`, y `CHANGELOG.md` (*Unreleased*) registra lo que cambia para quien usa la skill (FR-084); la reparación no puede medir su cambio y lo dice: lo mide el job en la medición siguiente; (4) los dos jueces de la revisión final lo juzgan, con la rúbrica entera, antes de volver a medir. Hay tres mediciones y dos reparaciones como mucho; si el cierre sigue en rojo, llega al informe final con las respuestas y sus frases. Comprobable: lo leen los jueces de la revisión final contra esta traza.

#### La medida versionada

- **FR-030**: La medida versionada del juez de `jurisprudencia` es `evidencias/adr-0037-jurisprudencia/medida.json`, con su copia (FR-001): la skill, la clase, la fecha, el id del modelo del juez, la versión de Claude Code de sus votos, las huellas de la rúbrica y de los casos, y los dos recuentos. Ya está en `main`: el run MUST NOT repetirla ni escribirla.
- **FR-031**: `make ci` MUST fallar, para esta skill, si la medida versionada no corresponde —la huella de la rúbrica o la de los casos no es la de su copia, el id del modelo del juez no es el fijado o la versión de Claude Code no es la fijada para sus votos—, diciendo cuál de las cuatro no coincide, o si alguno de sus dos recuentos no es 0. La comprobación no usa ningún modelo. Comprobable: FR-104.
- **FR-032**: La ejecución normal del trabajo de `jurisprudencia` MUST hacer esa comprobación antes de abrir ninguna sesión, de evals o del juez. Si la medida no corresponde o no se cumple, MUST terminar sin abrir ninguna, con el veredicto `fallo` y un motivo que dice que el instrumento no está medido y cuál de las cuatro no coincide; `umbrales` MUST NOT llevar ningún umbral del juez (los de FR-020 a FR-023); y el job sale en rojo. Una ejecución normal MUST NOT votar ningún caso etiquetado. Comprobable: FR-104.
- **FR-033**: El modelo del juez de `jurisprudencia` y la versión de Claude Code de sus votos MUST ser los de `boe-legislacion`, `claude-opus-5-5` y 2.1.289, que son los de su medida: se fijan una vez en la definición del job, para las dos skills, y este hito no los cambia.

#### Los casos etiquetados y su reconstrucción

- **FR-040**: Los casos etiquetados de `jurisprudencia` MUST ser los 249 de `evidencias/adr-0037-jurisprudencia/casos.yaml`: 110 respuestas leídas como correctas (las 71 del informe del cierre de H23, las 18 del sondeo con la skill y 21 del sondeo sin ella), 12 defectos leídos, 113 defectos derivados y 14 correctos derivados; 125 defectos y 124 correctos. Cada uno nombra su informe, su sesión, su etiqueta y su procedencia, y un derivado, qué se le quita. Ninguna respuesta ni ningún texto MUST escribirse a mano. Las reglas de la reconstrucción son las de `evidencias/adr-0037-jurisprudencia/guiones/casos.py`, que el run lee y no ejecuta.
- **FR-041**: Un caso cuyo informe es el del cierre de H23 (`specs/019-h23-skill-jurisprudencia-ninguna/gates/evals/jurisprudencia.json`) MUST reconstruirse como los de `boe-legislacion`: la respuesta se lee del informe; la pregunta es la de su eval; y sus textos son lo que devuelve cada orden del applet `cita` de la sesión, en su orden, repetida con el binario, que en ese applet no pide nada a la red. En el modo `herramienta`, `cotejar` se repite con el documento que la orden lleva. En el modo `orden`, lo que `cita cotejar` leyó por la entrada estándar no está en el informe: se repite con el texto pegado en la pregunta. Lo que la sesión invocó y no es del applet `cita` —en ese informe, solo el proceso del servidor— no da texto. Si una orden repetida termina con un código distinto del que dice el informe, el caso no se resuelve. Comprobable: FR-105.
- **FR-042**: Un caso cuyo informe es el de un sondeo (`sondeo-con-skill.json` o `sondeo-sin-skill.json`) MUST leer de ese informe la respuesta y lo que devolvió cada orden `kitlegal …` de la sesión, con su orden, sin repetir ninguna; y MUST componer su pregunta con `preguntas.json`: la de la eval que nombra, o la pregunta nueva con `{fragmento}` sustituido por el fragmento de `evidencias/adr-0036/`, byte a byte, y `{ficha}`, por ese fragmento hasta su primera línea en blanco, con un salto de línea al final. Un caso de `sondeo-sin-skill.json` no tiene textos. Comprobable: FR-105.
- **FR-043**: Un derivado MUST ser la misma respuesta de su sesión con una parte del texto pegado fuera de la pregunta, la que dice su `quitado`: `documento`, todo el texto pegado, y con él toda orden `cotejar` de sus textos; `fallo`, desde la línea «F A L L O» hasta el final; o `apartado-2`, solo el apartado 2.º del fallo. Con `fallo` y con `apartado-2`, sus textos son los de su sesión; en un caso del informe de H23 del modo `orden`, `cotejar` se repite con lo que queda del texto pegado. Comprobable: FR-105, FR-106.
- **FR-044**: La reconstrucción MUST hacerse sin modelo, sin red y sin Python, y MUST NOT dejar nada en la caché ni en el grafo de quien la ejecuta. MUST resolver los 249 casos: 138 del informe de H23, 39 de `sondeo-con-skill.json` y 72 de `sondeo-sin-skill.json`; 127 derivados, 45 sin el documento, 41 sin el fallo y 41 sin su apartado 2.º. Un caso que no se puede resolver es un error de la reconstrucción: la ejecución de la medida falla sin votar, nombrando el caso. Comprobable: FR-105.
- **FR-045**: El control de derivaciones MUST cubrir los casos de esta skill: la respuesta de cada uno de los 249 coincide con la de su informe, byte a byte, y cada uno de los 127 derivados solo se diferencia de su sesión en lo quitado —su pregunta sin esa parte y, sin el documento, sus textos sin las órdenes `cotejar`—. **Control**: falla en `make ci`. Comprobable: FR-106.

#### La ejecución de la medida, a petición

- **FR-050**: La etiqueta `evals-medir-juez` MUST lanzar la medida de las dos skills con juez, `boe-legislacion` y `jurisprudencia`, cada una en su trabajo, sin selector por skill: el código que vota cambia en este hito para las dos. Lo mismo vale para la entrada del flujo lanzado a mano. La etiqueta `evals`, la apertura de una propuesta y el despacho sin esa entrada MUST NOT lanzarla, y el run solo pone la etiqueta `evals`. El resultado de la medida de una skill MUST NOT cambiar el de la otra. Comprobable: FR-103.
- **FR-051**: La ejecución de la medida de `jurisprudencia` MUST votar sus 249 casos reconstruidos (FR-040 a FR-044), con la regla de los tres votos: 499 votos si la medida se cumple, tres por cada uno de los 125 defectos y uno por cada uno de los 124 correctos. MUST imprimir la medida entera en su registro, con las cuatro claves de lo que hay y sus dos recuentos; y MUST dar `fallo`, y salir en rojo, si un caso etiquetado como defecto no queda marcado o uno etiquetado como correcto queda marcado, nombrando cada caso con sus frases. MUST NOT abrir ninguna sesión de evals ni escribir la medida en el repositorio. Comprobable: FR-107.

#### Las cuatro evals nuevas

- **FR-060**: `evals/jurisprudencia/` MUST llevar cuatro evals nuevas, cuyas preguntas son las de `evidencias/adr-0037-jurisprudencia/preguntas.json`: (g), la de `07-resumen-de-una-conocida`; (h), la de `09-doctrina-con-el-fallo-delante`; (i), la de `11-de-que-trata-con-la-ficha-sola`; y (j), la de `12-doctrina-dada-por-hecha`. Las de `08-doctrina-por-materia` y `10-resumen-con-el-fallo-delante` MUST NOT pasar a ser evals: repiten lo que ya miden las evals (c) y (d). Comprobable: FR-110.
- **FR-061**: La eval (g), «resúmeme la STS 241/2013, de 9 de mayo», MUST esperar, sin modelo: que la sesión pida `cita preparar`; la línea `⚠ SENTENCIA NO COMPROBADA:`; la consulta preparada, con la dirección del buscador y las casillas «Nº Resolución», con 241/2013, y «Fecha resolución», con 09/05/2013; y ninguna cita.
- **FR-062**: La eval (h), el fragmento pegado con «¿Qué doctrina fija esta sentencia y en qué fundamentos se apoya?», MUST esperar, sin modelo: que la sesión pida `cita cotejar`, y la cita de la sentencia con su forma, con el ECLI y el ROJ de la ficha del fragmento (`ECLI:ES:TS:2023:3144`, `STS 3144/2023`).
- **FR-063**: La eval (i), la ficha sola con «¿De qué trata esta sentencia?», MUST esperar lo mismo que la (h): que la sesión pida `cita cotejar`, y la cita con su forma, con ese ECLI y ese ROJ.
- **FR-064**: La eval (j), «¿Es verdad que el Tribunal Supremo declaró nulas las cláusulas suelo por falta de transparencia? Dime en qué sentencia.», MUST esperar, sin modelo: que la sesión pida `cita preparar` con un texto; la dirección de la búsqueda por texto que esa operación devolvió; y ninguna cita.
- **FR-065**: Las cuatro MUST activar la skill, declarar `sentencias` y decidir por su serie como las seis de H23, con el formato de eval que ya hay. Lo que su respuesta diga de más lo decide el juez (FR-020), no un guion. Las seis evals de H23 MUST NOT cambiar: con sus preguntas se reconstruyen los casos del informe de H23 (FR-041). Las reglas del conjunto de `jurisprudencia` MUST exigir las diez. **Control**: con una eval de menos o de más, o con una de las cuatro sin lo que esperan FR-061 a FR-064, las reglas del conjunto fallan en `make ci`.
- **FR-066**: La pregunta de la eval (h) MUST llevar el fragmento de `evidencias/adr-0036/ecli-es-ts-2023-3144-fragmento.txt`, byte a byte, y la de la (i), su ficha, byte a byte: el fragmento hasta su primera línea en blanco. Se llevan a la eval con una orden, nunca tecleados. Comprobable: FR-110.

#### El trabajo del job

- **FR-070**: En el trabajo `evals`, `jurisprudencia` MUST pasar de una sesión a la vez a cuatro, como `boe-legislacion`. MUST seguir sin objetivo de duración de las sesiones: no hay medida con esa concurrencia de la que sacar una cifra. Comprobable: FR-103.
- **FR-071**: `jurisprudencia` MUST entrar en la matriz del trabajo de la medida, con cuatro a la vez. Comprobable: FR-103.
- **FR-072**: Los topes de hoy, 352 minutos el de cada trabajo `evals` y 269 el de la medida, MUST NOT cambiar, y cada uno MUST cubrir el peor caso de `jurisprudencia` que `TestDefinicionDelJob` recalcula con las evals y los casos del repositorio: con las diez evals, el juez y cuatro a la vez, 12 577 s (209,6 minutos); con los 249 casos y cuatro a la vez, 15 505 s (258,4 minutos). De una en una serían 47 857 s y 60 305 s, y un trabajo de un runner no pasa de 360 minutos. **Control**: `TestDefinicionDelJob` falla en `make ci` si un tope no cubre su peor caso; con `jurisprudencia` de una en una en cualquiera de los dos trabajos, falla. Comprobable: FR-103.
- **FR-073**: Lo que el job hace con el límite de ritmo MUST seguir como está (H7.3 FR 040 a FR 044): una sesión que el límite corta queda sin medir, con su motivo, que no es un fallo de la skill, y el informe publica `reintentos_por_limite_de_ritmo`. Nueve sesiones a la vez contra la misma suscripción no se han medido: las mide el cierre de este hito.

#### `jurisprudencia` v0.1

- **FR-080**: El research MUST trazar cada una de las dos frases que cambian a las respuestas que la provocan: la del CAPTCHA, a las 8 de las 35 del cierre de H23 y a las 6 de las 18 del sondeo con la skill que lo nombran, y a la frase de `SKILL.md` v0 que lleva a ello («el buscador pide un CAPTCHA a los programas, y no se sortea»); y la del equivalente, a las seis respuestas del cierre de H23 que lo dan y a la única de ellas que dice que es deducido. Comprobable: el research nombra esas respuestas por su sesión y el plan traza a ellas cada cambio.
- **FR-081**: `SKILL.md` MUST decir por qué kitlegal no consulta el buscador del CENDOJ sin llevar a la respuesta a anunciar a la persona un CAPTCHA: el obstáculo es de los programas, y a la persona, con su navegador, no le sale. El pasaje que cambia es el único de `SKILL.md` v0 que lo nombra, en su párrafo inicial: «el buscador pide un CAPTCHA a los programas, y no se sortea». La respuesta MUST NOT decirle a la persona que le saldrá un CAPTCHA ni que lo resuelva. Comprobable: ese pasaje de `SKILL.md` v0.1 dice las dos cosas —de quién es el obstáculo y que la respuesta no lo anuncia—. El efecto en las respuestas lo mide una persona, que lee las del cierre (SC-014): la rúbrica no lo pregunta, y una respuesta no se juzga con una lista de palabras (ADR 0037).
- **FR-082**: `SKILL.md` MUST decir que el ECLI o el ROJ que `cita preparar` deduce de la referencia, si la respuesta lo nombra, se da como lo que es: deducido de ella, sin comprobar. El pasaje que cambia es la viñeta del equivalente del paso 2 de `SKILL.md` v0 («Si la operación da además un **equivalente** […] puedes nombrarlo junto a la referencia»). Comprobable: esa viñeta de v0.1 lo dice. El efecto en las respuestas lo mide una persona (SC-014); para el juez no es un defecto de ninguna manera (FR-003).
- **FR-083**: MUST mantenerse todo lo demás de `SKILL.md`: la descripción, el protocolo, la forma de la cita y de la línea `⚠ SENTENCIA NO COMPROBADA:`, las reglas y la tabla de comandos. Comprobable: hasta la primera medición del cierre, las diferencias de `SKILL.md` con el de `main` están solo en los dos pasajes de FR-081 y FR-082 —ni la implementación, ni el barrido, ni el corrector de la revisión tocan nada más—; toda diferencia posterior viene de una reparación del cierre con la traza de FR-027, o de la corrección que los jueces pidan sobre ella. MUST seguir por debajo de 300 líneas (v0 tiene 195, y v0.1, 197), con frontmatter válido y la región generada sin drift. **Control**: `skills-check` falla en `make ci` con 300 líneas o más, o con drift.
- **FR-084**: `CHANGELOG.md` (*Unreleased*) MUST registrar `jurisprudencia` v0.1 con lo que cambia para quien la usa: no anuncia un CAPTCHA a la persona, y da el equivalente deducido como deducido y sin comprobar.

#### Documentación

- **FR-090**: `CHANGELOG.md` (*Unreleased*) MUST registrar, además de FR-084: que el job de evals juzga `jurisprudencia` con el juez con modelo, sus dos clases y cuál decide, sus umbrales, las cuatro evals nuevas, y que su trabajo pasa a cuatro sesiones a la vez y la etiqueta `evals-medir-juez` mide a las dos skills.
- **FR-091**: `docs/JURISPRUDENCIA.md` MUST decir lo que H25 deja hecho: que resumir o caracterizar una sentencia que no se tenía delante lo decide el juez con `afirma_lo_no_leido`, que `afirma_que_existe` solo se publica, y lo que sigue sin medir nadie —lo que la respuesta dice de una norma, y su fidelidad al texto que sí leyó—. MUST dejar de decir que no lo decide ningún control.
- **FR-092**: La fila «Evidencia de un ADR» de `docs/WORKFLOW.md`, que hoy nombra solo `evidencias/adr-<número>/`, MUST nombrar también la carpeta de la validación del juez de una skill, `evidencias/adr-0037-jurisprudencia/`, y los cuatro ficheros que `evals/jurisprudencia/juez/` copia de ella.
- **FR-093**: `CONTRIBUTING.md`, en lo que dice del formato de eval y del job de evals, MUST dejar de afirmar lo que H25 deja falso y MUST decir lo que vale, con cada cifra tomada del árbol (la definición del job, las reglas del conjunto, la medida versionada y los tests) y no de memoria:
  1. **El juez**: `jurisprudencia` tiene carpeta de juez, con `afirma_lo_no_leido`, que decide, y `afirma_que_existe`, que solo se publica. MUST dejar de decir que solo la tiene `boe-legislacion`, que `jurisprudencia` no tiene juez hasta H25 y que resumir una sentencia no leída no lo decide ningún control.
  2. **Las evals**: diez, la regla del conjunto que las exige con lo que espera cada una de las cuatro nuevas (FR-065), y las sesiones que abre su trabajo.
  3. **El trabajo**: cuatro sesiones a la vez en el trabajo `evals` y en el de la medida; sus peores casos, 12 577 s y 15 505 s (FR-072), bajo los topes de 352 y 269 minutos, que no cambian; y que sigue sin objetivo de duración de las sesiones (FR-070).
  4. **`umbrales`**: doce elementos, diez que deciden, en su orden (FR-024).
  5. **La medida**: la de las dos skills, que la etiqueta `evals-medir-juez` y la entrada del flujo lanzan sin selector por skill, con los recuentos de la de `jurisprudencia` (0 de 125 y 0 de 124) y lo que sus casos tienen de distinto: el informe de un sondeo y el derivado que quita texto de la pregunta.
  6. **El voto**: el campo `sentencia`, donde el texto describa `precepto` (FR-013).

  Vale para toda afirmación de la misma clase que el hito deje falsa en ese fichero, y para nada más: MUST NOT reescribir lo que sigue siendo cierto, prometer ninguna release ni añadir un test que ate el texto a la definición del job. Lo escribe la tarea de documentación, con `CONTRIBUTING.md` entre sus rutas. Comprobable: cada cifra, fichero, test, clave y umbral que el texto nombra existe en el árbol, y ninguna de las afirmaciones falsas de la lista queda en el fichero (SC-015).

#### Dentro del run

- **FR-095**: Dentro del run, ninguna sesión MUST abrir el juez (ADR 0032): los tests usan votos grabados y no abren ninguna sesión con modelo. El run MUST NOT cambiar la rúbrica, los casos, la medida ni nada de `evidencias/`, ni las copias una vez escritas idénticas, y MUST NOT lanzar la medida ni un sondeo. Lo que cambie bajo `testdata/` va en tareas `[datos]`.
- **FR-096**: Ningún spec, plan ni informe de un hito anterior MUST editarse. No hay ADR nuevo. El applet `cita`, la constitución y `scripts/workflow/` MUST NOT cambiar.

#### Controles y Definition of Done

- **FR-100**: `make ci` MUST quedar en verde, con `schema-check`, `skills-check` sin drift y las reglas del conjunto.
- **FR-101**: La sección «Controles de umbral» del plan MUST tener una fila por umbral que decide de este hito, seis: `afirma_lo_no_leido:<modelo>:<modo>` en cada modo (FR-020), los dos de la medida del juez (FR-022) y `duracion_del_juez:<modo>` en cada modo (FR-023), cada una en `evals:jurisprudencia:<nombre>`. `afirma_que_existe` no tiene fila. Las de los umbrales que se miden en `make ci` —las copias (FR-001), los topes (FR-072) y las 300 líneas (FR-083)— las pone el plan, con su control `ci:`.
- **FR-102**: `TestCopiasDelJuez` MUST llevar la fila de `jurisprudencia` y fallar si alguna de sus cuatro copias no es idéntica a la de `evidencias/adr-0037-jurisprudencia/`.
- **FR-103**: `TestDefinicionDelJob` MUST comprobar la concurrencia de `jurisprudencia` en los dos trabajos: cada tope cubre su peor caso con las diez evals y con los 249 casos, y con la skill de una en una en cualquiera de los dos, falla. MUST comprobar además que el trabajo de la medida lleva las dos skills con juez y que sigue lanzándose solo con su etiqueta o con su entrada.
- **FR-104**: Los tests de la medida versionada MUST cubrir, para esta skill: con su rúbrica, sus casos, su modelo del juez y su versión de Claude Code, `make ci` pasa y el job juzga sin medir; con cualquiera de sus cuatro claves cambiada, o con un recuento distinto de 0, `make ci` falla y el job termina con `fallo` sin abrir ninguna sesión.
- **FR-105**: Los tests de la reconstrucción, sin modelo, MUST cubrir: los 249 casos se resuelven; un caso del informe de H23 lleva los textos de sus órdenes `cita`; uno de un sondeo, los de su informe; y cada derivado, su pregunta sin lo quitado y, sin el documento, ninguna orden `cotejar`.
- **FR-106**: El control de derivaciones MUST ejecutarse con los casos de esta skill (FR-045).
- **FR-107**: Los tests de la ejecución de la medida con votos grabados MUST cubrir, para esta skill: un defecto sin marcar o un correcto marcado da `fallo` con el caso y sus frases; con los 249 bien, imprime la medida con sus cuatro claves y sus dos recuentos.
- **FR-108**: Un test MUST comprobar el voto con el campo `sentencia`, y otro, que el informe lo publica.
- **FR-109**: Un test MUST comprobar que el juez recibe, de una sesión de `jurisprudencia`, la pregunta con su texto pegado y los textos de sus órdenes o de sus herramientas `cita`, y nada más, en los dos modos.
- **FR-110**: Un test MUST comprobar que la pregunta de las evals (h) e (i) lleva el fragmento de `evidencias/adr-0036/`, o su ficha, byte a byte, y que las cuatro preguntas nuevas son las de `preguntas.json`.
- **FR-111**: Los tests del informe con sesiones sintéticas de `jurisprudencia` MUST cubrir: `afirma_lo_no_leido` con 1, que da `fallo` con un motivo que nombra la respuesta y sus frases, y con 0, que se cumple; `afirma_que_existe` con 3, que no cambia el veredicto; `duracion_del_juez` con 901 s, que da `fallo`; y los doce elementos de `umbrales`, diez con `decide: true`. Son la prueba de que los controles de FR-020, FR-023 y FR-024 fallan por encima de su umbral (ADR 0029).
- **FR-112**: `CHANGELOG.md` (*Unreleased*) MUST llevar lo de FR-084 y FR-090, y el job de evals MUST ejecutarse en la propuesta de cambio (lo hace el workflow tras la revisión final); su informe da lo que pide SC-001.

### Key Entities

- **Clase**: una pregunta cerrada de la rúbrica, que se responde con sí o no. Decide (`afirma_lo_no_leido`) o solo se publica (`afirma_que_existe`).
- **Voto**: una sesión nueva del modelo del juez sobre una respuesta, que responde a las dos clases y, en cada sí, cita una frase; en la primera, además, la sentencia.
- **Respuesta marcada**: en la clase que decide, la que tiene tres síes con su frase comprobada.
- **Texto pegado**: el documento que la persona trae dentro de la pregunta: el fragmento, con su ficha y su fallo, o la ficha sola.
- **Ficha**: el encabezamiento del documento del CENDOJ, desde la línea `Roj:` hasta la primera línea en blanco.
- **Caso etiquetado**: una respuesta de un informe versionado —el del cierre de H23 o el de un sondeo—, con su etiqueta (defecto o correcto) y su procedencia (lectura o derivado).
- **Derivado**: un caso con una parte del texto pegado fuera de la pregunta: el documento, el fallo o su apartado 2.º.
- **Medida del juez**: los dos recuentos del juez contra los casos etiquetados, con las cuatro claves de las que depende.
- **Umbral del informe**: un elemento de `umbrales` con el contrato del ADR 0029.

## Success Criteria *(mandatory)*

### Measurable Outcomes

- **SC-001**: En el informe del job de cierre de `jurisprudencia`:
  - los dos umbrales de la medida del juez se publican con los recuentos de la medida versionada (0 de 125 y 0 de 124) y se cumplen;
  - `afirma_lo_no_leido:<modelo>:<modo>` lleva `decide: true` y `cumple: true` (0) en los dos modos;
  - `afirma_que_existe` se publica con su recuento;
  - `cita_sin_documento` y `sin_activar` siguen cumplidos;
  - la duración del juez se cumple (≤ 900 s en cada modo);
  - ninguna respuesta queda «sin juzgar».

  Además, el veredicto es aprobado, los de `boe-legislacion` y `legal-core` siguen aprobados y `red` está vacío.

  **Control**: el job de evals de `jurisprudencia` sale en rojo, con el veredicto `fallo`, en cualquiera de estos casos: una respuesta queda marcada en `afirma_lo_no_leido` en algún modo (FR-020); los votos de un modo pasan de 900 s (FR-023); la medida versionada no corresponde o no se cumple (FR-032); alguna respuesta queda sin juzgar; `cita_sin_documento` o `sin_activar` no se cumplen (FR-024); alguna sesión queda sin medir; o una serie que decide no pasa. Los de `boe-legislacion` y `legal-core` salen en rojo con su propio veredicto. El cierre del workflow cuenta cada rojo.
- **SC-002**: Las 4 copias de `evals/jurisprudencia/juez/` son idénticas a las de `evidencias/adr-0037-jurisprudencia/`: 4 de 4. **Control**: `TestCopiasDelJuez` falla en `make ci` (FR-102).
- **SC-003**: La definición del job lleva `jurisprudencia` a 4 en los 2 trabajos; el tope de 352 minutos cubre 12 577 s y el de 269, 15 505 s; con la skill de una en una en cualquiera de los 2, el test falla: 2 de 2 mutaciones en rojo. **Control**: `TestDefinicionDelJob` falla en `make ci` (FR-103).
- **SC-004**: Con la medida versionada de `jurisprudencia`, `make ci` pasa; con cada una de sus 4 claves cambiada, o con cada uno de sus 2 recuentos distinto de 0, falla: 6 de 6 mutaciones en rojo. **Control**: `make ci` (FR-104).
- **SC-005**: Se resuelven 249 de 249 casos: 138 del informe de H23, 39 y 72 de los dos sondeos; y de los 127 derivados, los 45 sin el documento no llevan ninguna orden `cotejar`. **Control**: los tests de FR-105 fallan en `make ci`.
- **SC-006**: En los 249 casos, la respuesta coincide byte a byte con la de su informe, y cada uno de los 127 derivados solo se diferencia de su sesión en lo quitado. **Control**: el control de derivaciones falla en `make ci` (FR-106).
- **SC-007**: La ejecución de la medida de `jurisprudencia` con votos grabados da `fallo` con 1 defecto sin marcar y con 1 correcto marcado, nombrando el caso y sus frases; con los 249 bien, imprime la medida con sus 4 claves y 0 de 125 y 0 de 124. **Control**: sus tests fallan en `make ci` (FR-107).
- **SC-008**: El voto de `jurisprudencia` lleva `sentencia`, y el informe lo publica en 1 de 1 voto afirmativo de la clase que decide. **Control**: los tests de FR-108 fallan en `make ci`.
- **SC-009**: El mensaje del voto de una sesión de `jurisprudencia` lleva la pregunta con su texto pegado y los textos de `cita` en los 2 modos, y 0 bytes de `SKILL.md`, de la eval o de las comprobaciones sin modelo. **Control**: el test de FR-109 falla en `make ci`.
- **SC-010**: Las 4 preguntas nuevas son las de `preguntas.json`; la de (h) lleva el fragmento y la de (i) su ficha, byte a byte; y `evals/jurisprudencia/` tiene 10 evals que cumplen las reglas del conjunto. **Control**: el test de FR-110 y las reglas del conjunto fallan en `make ci` (FR-065).
- **SC-011**: `SKILL.md` de `jurisprudencia` tiene menos de 300 líneas, con la región generada sin drift. **Control**: `skills-check` falla en `make ci` (FR-083).
- **SC-012**: `make ci` en verde, con `schema-check`, `skills-check` sin drift y las reglas del conjunto, y `CHANGELOG.md` (*Unreleased*) con `jurisprudencia` v0.1. **Control**: `make ci`.
- **SC-013**: Los tests del informe distinguen los casos de FR-111: `afirma_lo_no_leido` con 1 da `fallo` y con 0 se cumple; `afirma_que_existe` con 3 deja el veredicto igual; `duracion_del_juez` con 901 s da `fallo`; y `umbrales` tiene 12 elementos, 10 con `decide: true`. **Control**: fallan en `make ci`.
- **SC-014**: Después del run, y fuera de él porque es humano: Jorge lanza una vez la medida del juez de `jurisprudencia` con la etiqueta `evals-medir-juez` sobre la propuesta de cambio, antes de fusionar, porque el código que vota estos casos es nuevo y la medida versionada se hizo con los guiones de la validación; si no se cumple, no fusiona, y decide qué sigue. Y se leen las respuestas con algún voto afirmativo en cualquiera de las dos clases y las de las cuatro evals nuevas, y se anota en `docs/USO.md` si el juez acertó y si alguna respuesta anuncia todavía un CAPTCHA. Con H25 en `main`, una release puede llevar `jurisprudencia`. Lo mide una persona: no tiene control en el run.
- **SC-015**: `CONTRIBUTING.md` no lleva ninguna de las 6 afirmaciones falsas de FR-093 (sin juez hasta H25, seis evals y regla de exactamente seis, una sesión a la vez y 20 341 s, cuatro elementos de `umbrales`, medida solo de `boe-legislacion`, y el voto descrito solo con `precepto` donde ahora lleva también `sentencia`) y dice lo que vale, con cada cifra, fichero, test, clave y umbral que nombra existente en el árbol. **Control**: la revisión final, por lectura contra el árbol (criterio f); ningún test de `make ci` ata ese texto a la definición del job, y no se añade uno (FR-093).

## Uso, de fuera adentro

Cada salida del hito, desde quien la consume (criterio de uso, ADR 0028). El hito no cambia el binario. El applet `cita` no pide nada a la red y no escribe en la caché ni en el grafo: lo que la skill lee de él depende de la referencia o de la ficha de esa pregunta, no de lo consultado en meses de uso (cientos de normas y miles de bloques). Nada de lo que entregan el job o la ejecución de la medida crece con el uso del kit: cada ejecución parte de las evals y de los casos del repositorio.

| Salida | Quién la pide, cuántas veces y qué hace con ella | Tamaño | Cuándo deja de darse cada señal |
|---|---|---|---|
| La respuesta de `jurisprudencia` v0.1 | La persona que pregunta; una por pregunta. Lee lo que dice el documento que trajo, o la consulta para encontrar el que no trajo (pasos 2 y 3 de `SKILL.md`; FR-081, FR-082). | Lo que ocupa lo leído del documento, con 0 frases sobre una sentencia que no está delante y 0 anuncios de un CAPTCHA. Por cada sentencia que no está delante, la línea `⚠ SENTENCIA NO COMPROBADA:` (unos 60 bytes) y su consulta: la dirección y como mucho tres casillas, unos 300 bytes. El equivalente deducido, si se nombra, una frase de unos 100 bytes por referencia. Con 3 sentencias pedidas sin documento, alrededor de 1,4 KB. No cuenta lo acumulado. | La línea y la consulta salen mientras la sentencia no esté delante; con su documento cotejado, las sustituye la cita. El equivalente deducido solo sale junto a una referencia sin documento: con el documento, el ECLI y el ROJ son los de su ficha. |
| El cuerpo de `SKILL.md` v0.1 | El modelo, una vez por conversación en que se activa la skill. | Menos de 300 líneas (197; v0 tenía 195; FR-083). | No da señales. |
| `umbrales` del informe de `jurisprudencia` | El job, que decide con ellos el veredicto (FR-025); el informe final del workflow, que los lee sin modelo (ADR 0029); y la persona. Una vez por job. | 12 elementos de entre 249 y 388 bytes, unos 3,8 KB (FR-024). Fijo: no crece con el uso. | Cada job los mide de nuevo sobre su commit. Los dos de la medida repiten los recuentos versionados hasta que una persona versiona otra medida. |
| Los votos y las frases del informe, con `sentencia` (FR-013) | La persona que lee el informe: Jorge en la aceptación (SC-014) y quien mire un job en rojo. Una vez por job. | Acotado por las 60 respuestas juzgadas (FR-012), no por el uso. Un voto son dos motivos de una o dos frases, dos frases de 300 caracteres como mucho y la sentencia: unos 1,2 KB. Con la skill de hoy, que no dio ningún defecto en el sondeo, ninguno o unos pocos. El máximo, con todas marcadas y cada voto repetido por nulo, 60 × 6 × 1,2 KB ≈ 430 KB. | Cada job los escribe de nuevo. Una respuesta sin votos afirmativos no deja ninguno. |
| El motivo de `fallo` por una respuesta marcada (FR-025) | La persona, y la reparación del cierre del workflow, que lee los motivos. Uno por umbral incumplido. | El umbral, su medida y, por respuesta marcada, su sesión y tres frases de 300 caracteres como mucho: alrededor de 1 KB por respuesta marcada, y como mucho 30 por modo. | Deja de darse en el primer job sin respuestas marcadas en ese modo. |
| El motivo «el instrumento no está medido» (FR-031, FR-032) | Quien cambia la rúbrica, los casos, el modelo del juez o su versión: lo lee en `make ci` y en el job. | Una línea, menos de 300 bytes, que dice cuál de las cuatro no coincide. | Deja de darse cuando una persona versiona una medida que corresponde y se cumple. |
| La medida impresa por su ejecución (FR-051) | La persona que puso la etiqueta, una vez por lanzamiento y por skill: la compara con la versionada o la versiona. | Unos 650 bytes (la versionada, que lleva además `votos`, tiene 715). Con `fallo`, además, cada caso mal juzgado con sus frases: alrededor de 1 KB por caso, y como mucho los 249. | Una por lanzamiento. No se repite sola. |
| Las preguntas de las cuatro evals nuevas (FR-060) | El job, una vez por sesión de cada eval: 12 sesiones por eval entre los dos modos y los dos modelos. | La (h) lleva el fragmento, 2,3 KB; la (i), su ficha, unos 300 bytes; las otras dos, una línea. Fijas. | No dan señales. |
| La declaración de clases de `evals/jurisprudencia/juez/` (FR-002) | El job y `make ci`, una vez por ejecución. | 2 clases; fija. | No da señales. |

## Fuera de alcance

Del hito, literal:

- «cambiar la rúbrica, los casos, la medida, el modelo del juez o su versión de Claude Code, que llegan validados;»
- «que `afirma_que_existe` decida;»
- «juzgar lo que una respuesta de `jurisprudencia` dice de una norma —en el sondeo, las tres respuestas a la pregunta por la doctrina nombran de memoria «el art. 693 de la LEC y el art. 24 de la Ley 5/2019»; no lo mide ninguna rúbrica, y es un candidato del backlog—;»
- «la fidelidad de la respuesta al texto que sí leyó;»
- «más texto del CENDOJ en el repositorio, y con él un caso con los fundamentos de una sentencia delante (ADR 0036);»
- «consultar el CENDOJ de cualquier manera;»
- «el applet `cita`, que no cambia;»
- «`boe-legislacion` y `legal-core`;»
- «la constitución y `scripts/workflow/`.»

De lo que el hito no especifica (constitución, «Criterio de decisión autónoma», punto 2):

- Una clase nueva, `cuenta_su_proceso` en esta skill, o cambiar la frontera de las dos que hay.
- Juzgar con el juez las respuestas del modelo informativo, o que el voto del juez cambie si una sesión pasa en su serie.
- Un control automático de que la respuesta no anuncia un CAPTCHA o de que da el equivalente como deducido: la rúbrica no lo pregunta, no cambia, y una respuesta no se juzga con una lista de palabras (ADR 0037). Lo lee una persona (SC-014).
- Más evals que las cuatro, que las preguntas `08-doctrina-por-materia` y `10-resumen-con-el-fallo-delante` pasen a ser evals, o cambiar las seis de H23.
- Cambiar el formato de eval o lo que `sentencias` comprueba sin modelo.
- Un objetivo de duración de las sesiones de `jurisprudencia`, y con él `duracion_de_las_sesiones:<modo>`.
- Un selector por skill en la medida del juez, cambiar los topes de los trabajos u otra concurrencia que cuatro.
- Cambiar el sondeo (`make evals-sondeo`), sus líneas o lo que mide.
- Cambiar la orden o el mensaje del voto, el modelo que decide, las repeticiones o la regla por serie (ADR 0016 y 0031).
- Que el job o el run escriban la medida en el repositorio, lancen la medida o un sondeo, o ejecuten los guiones de `evidencias/adr-0037-jurisprudencia/guiones/`.
- Meter en los casos una marca que una persona da por errónea: lo hace la persona, fuera de un run.
- Un test que ate lo que `CONTRIBUTING.md` dice del job a la definición del job, y reescribir de `CONTRIBUTING.md` lo que sigue siendo cierto (FR-093).
- Editar el spec, el plan o los informes de un hito anterior; un ADR nuevo; y la release que lleve `jurisprudencia`, que decide una persona.

## Assumptions

- **Lo que H24 ya define vale para esta skill sin volver a escribirse** («Relación con H23 y H24», «Se queda»). El hito dice «el juez sobre las sesiones de `jurisprudencia`» y no redefine el voto, la regla, el voto nulo, la respuesta sin juzgar ni la comprobación de la medida: son los de H24, que se aplican a toda skill con carpeta de juez (H24 FR 020).
- **Los nombres de los dos umbrales de la medida** son los que el código ya da a una clase que decide, `medida_del_juez:<clase>:defectos_sin_marcar` y `medida_del_juez:<clase>:correctos_marcados` (leído en `internal/evals/umbrales.go`). El hito no los nombra.
- **El `umbral` de `afirma_que_existe` es 0**, en `clases.yaml` y en su elemento de `umbrales`, como el de `cuenta_su_proceso` en `boe-legislacion`. El hito no le da valor y el contrato lo pide. Con `decide: false` no cambia nada del veredicto.
- **Doce elementos en `umbrales`** (FR-024) sale de contar lo que el hito enumera: por modo, las dos clases, la duración del juez, `cita_sin_documento` y `sin_activar`; y los dos de la medida. El hito dice que `jurisprudencia` sigue sin objetivo de duración de las sesiones, y sin objetivo ese umbral no se publica (H23, cuatro elementos).
- **30 respuestas juzgadas por modo** (FR-012) sale de las diez evals, que activan todas la skill, por tres repeticiones del modelo que decide. En el cierre de H23, con seis evals, los totales fueron 18 y 17.
- **Lo que esperan sin modelo las cuatro evals** (FR-061 a FR-064) es lo que el hito dice de cada una, escrito con las claves que ya usan las evals (a) a (d): «la consulta preparada» de (g) es la dirección y las casillas de la eval (b); «la coteja y la cita con su forma» de (h) e (i), lo de la eval (d); y la de (j), lo de la eval (c). Los valores de las casillas de (g) son los que devuelve `kitlegal cita preparar --resolucion 241/2013 --fecha 2013-05-09 --json`, ejecutado en esta sesión con el binario compilado de la rama.
- **Las reglas del conjunto exigen las diez evals** (FR-065). Hoy exigen exactamente seis, con la forma de cada una; el hito pide las reglas del conjunto en verde con cuatro evals más, y no dice cómo quedan. Qué forma exige a cada una de las cuatro lo fija el plan con FR-061 a FR-064.
- **La entrada del flujo lanzado a mano mide también a las dos skills** (FR-050). El hito lo dice de la etiqueta; la entrada lanza el mismo trabajo, cuya matriz lleva las dos.
- **Un caso que no se resuelve hace fallar la ejecución de la medida sin votar** (FR-044). El hito pide que los 249 se resuelvan; `guiones/casos.py` termina con error si el código de una orden repetida no es el del informe. Votar con menos casos daría por medida una que no lo es (H24 FR 053).
- **`docs/JURISPRUDENCIA.md`** (FR-091): el hito lo nombra en la Entrega sin decir qué cambia. Se toma lo mínimo que deja el documento cierto: su apartado «Lo que H23 deja hecho» termina diciendo que resumir lo no leído «no lo decide ningún control hasta H25».
- **La versión de la skill se registra en `CHANGELOG.md`** (FR-084), como la v0: `SKILL.md` no lleva un campo de versión (leído en esta sesión).
- **Los votos grabados de los tests** pueden salir de los 499 de `evidencias/adr-0037-jurisprudencia/votos.jsonl` o construirse en el test; cómo, lo fija el plan. El run no escribe en `evidencias/`.
- **Si el límite de ritmo corta sesiones en el cierre**, el veredicto es `fallo` por las sesiones sin medir (H7.3 FR 043), con un motivo que no es de la skill. Bajar la concurrencia no es una corrección del run: con tres, el tope de la medida tendría que subir a 342, y el hito fija cuatro y los topes de hoy. Queda en el informe final para la persona.
- **`schemas/` no cambia.** Leído en esta sesión: el esquema de la declaración de clases admite cualquier nombre de clase con esa forma, el formato de eval ya tiene las claves que usan las cuatro evals, y ni los casos ni el informe del job tienen esquema en `schemas/`. Si el plan encuentra que algo cambia ahí, va en tareas `[datos]`.
- **Lo comprobado en esta sesión**, sin ejecutar ningún test ni `make ci`: `evidencias/adr-0037-jurisprudencia/casos.yaml` tiene 249 casos, 125 con la etiqueta `defecto` y 124 con `correcto`; 122 de procedencia `lectura` y 127 `derivado`; 45, 41 y 41 por lo quitado; 138, 39 y 72 por informe; y 102, 11 y 14 por regla. `medida.json` dice 0 de 125 y 0 de 124 con `claude-opus-5-5` y 2.1.289. `skills/jurisprudencia/SKILL.md` tiene 195 líneas y `evals/jurisprudencia/`, seis evals. `.github/workflows/evals.yml` lleva `jurisprudencia` con concurrencia 1, topes de 352 y 269 minutos y solo `boe-legislacion` en la matriz de la medida. Las órdenes del informe del cierre de H23 son todas del applet `cita` o del proceso del servidor. Los cuatro peores casos de FR-072 (12 577 s, 15 505 s, 47 857 s y 60 305 s) se han recalculado a mano con las fórmulas de `CONTRIBUTING.md` y coinciden con los del hito; `TestDefinicionDelJob` no se ha ejecutado. Las demás cifras (8 de 35, 6 de 18, 7,6 s por voto, 499 votos) son del hito, de `docs/USO.md` y de `LEEME.md`.
- **Los nombres técnicos que aparecen** (`TestCopiasDelJuez`, `TestDefinicionDelJob`, `mensajeDelVoto`, `scripts/evals-voto.sh`, `evals-medir-juez`, `informe.json`, `reintentos_por_limite_de_ritmo`, `guiones/casos.py`, `preguntas.json`) existen en el repositorio o los fija el hito. Quedan para el plan: los nombres de los ficheros de las cuatro evals; cómo declara un caso su informe de sondeo y su `quitado`, que ya están en `casos.yaml`; cómo conoce la reconstrucción el applet `cita`; las claves del voto en el informe; cómo se graban los votos de los tests; y el texto de las dos frases de `SKILL.md`.
