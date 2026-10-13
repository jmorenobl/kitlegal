# Feature Specification: H26 · El umbral de las clases del juez pasa a ser una tasa sobre los dos modos

**Feature Branch**: `021-h26-el-umbral-de`

**Created**: 2026-10-13

**Status**: Draft

**Input**: Sección «#### H26 · El umbral de las clases del juez pasa a ser una tasa sobre los dos modos (ADR 0038)» de `docs/ROADMAP.md`, en modo desatendido (ADR 0018).

## Resumen

El job de evals decide hoy `afirma_lo_no_leido` con 0 respuestas marcadas en cada modo, en `boe-legislacion` y en `jurisprudencia`. Ese cero no es un control:

- **Su verde no dice que la skill cumpla, y su rojo no dice nada del cambio que se mide.** Con la tasa medida de `boe-legislacion` desde H24 (5 de 536, un 0,9 %), una ejecución de 108 respuestas sale limpia el 36 % de las veces (ADR 0038, «Contexto y problema»). El cierre de H25 cayó en sus tres mediciones por una skill que no tocaba, y el de H24 pasó a la tercera sin que la tasa hubiera cambiado.
- **Ninguna de las cuatro condiciones probadas acerca la tasa a cero** en las preguntas por materia: con Sonnet, con Opus, con la rúbrica en el índice y con un paso de repaso quedan entre el 3 % y el 20 % (`evidencias/adr-0038/`; `docs/USO.md`, 2026-10-12).

El ADR 0038 ya decidió qué hacer: cero donde hay forma, tasa donde hay significado. Este hito hace que el job lo aplique. Por cada clase del juez que decide, el informe gana un umbral sobre las respuestas de los dos modos juntas, con la proporción que su `clases.yaml` admite, y ese es el que hace fallar el job. El recuento de cada modo se sigue publicando, sin decidir. Los valores son los del ADR: 0.028 en `boe-legislacion` (como mucho 3 marcadas de 108) y 0.017 en `jurisprudencia` (como mucho 1 de 60).

El hito no entrega skill nueva ni cambia ninguna. Protege las que hay: el job falla cuando la skill empeora y no por una respuesta suelta (principio VIII). Hasta que cierre no hay release.

**Cómo se citan los requisitos de otros hitos.** Los de H7.3 a H25 se citan sin guion: «H24 FR 010», «H25 FR 020». Los ids con guion (FR-001, SC-001…) son siempre de este spec. `<modelo>` es el id del modelo que decide (ADR 0031), hoy `claude-sonnet-5-5`; `<modo>` es `orden` u `herramienta` (H21); y `<clase>`, el nombre de una clase de `clases.yaml`. **Umbral conjunto** es el elemento `<clase>:<modelo>` de `umbrales`, y **recuento por modo**, el elemento `<clase>:<modelo>:<modo>`.

## Clarifications

### Session 2026-10-13

- Q: Cuando el umbral conjunto no se cumple pero su porcentaje, escrito con un decimal, coincide con el de la condición (3 marcadas de 107 respuestas juzgadas en `boe-legislacion`, 2,804 %, frente a 0.028; 1 de 58 en `jurisprudencia`, 1,724 %, frente a 0.017), ¿cómo escriben el motivo y la tabla «Umbrales» de `informe.md` la medida y la condición? → A: Opción A. Como todo umbral con `total`: la medida y la condición con un decimal, sin ningún texto más, en el motivo y en la fila de la tabla «Umbrales»; con 3 de 107 en `boe-legislacion`, «3 de 107 (2,8 %), y tiene que ser ≤ 2,8 %», y con 1 de 58 en `jurisprudencia`, «1 de 58 (1,7 %), y tiene que ser ≤ 1,7 %». El resultado lo dicen `cumple: false` y el veredicto `fallo`, que salen de la comparación sin redondear (FR-003), no el porcentaje escrito. No cambia cómo se escribe ningún umbral y no se añade ningún requisito. (auto: criterio a; fuente: `docs/ROADMAP.md`, H26, Alcance, «El umbral conjunto» («con un motivo que nombra el umbral y cada respuesta marcada […] como hoy») y «Fuera de alcance»; `internal/evals/umbrales.go` (`motivoDeUnUmbral`, `medidaEscrita`, `condicionEscrita`, `porcentajeConComa`); `docs/ADR/0029-umbrales-que-se-cumplen.md`, «Contrato de umbrales del informe del job de evals»; constitución, «Criterio de decisión autónoma», punto 2, y «Gates», proporcionalidad)
- Q: ¿Corrige el hito el comentario de `.github/workflows/evals.yml` que dice que el informe de `jurisprudencia` lleva, por modo, `afirma_lo_no_leido`, «que decide», y «doce umbrales, diez que deciden»? → A: Opción A. No se toca: el diff del hito no lleva nada bajo `.github/`, FR-055 y FR-062 no ganan ninguna excepción y SC-013 se queda en los cinco documentos de la Entrega 4. Queda en «Fuera de alcance» como pendiente para la persona, con sus datos: el comentario queda diciendo doce umbrales, diez que deciden y que `afirma_lo_no_leido` decide por modo, y lo que vale es trece, nueve que deciden, con el recuento por modo solo publicado (FR-012). (auto: conservadora; criterio d; fuente: `docs/ROADMAP.md`, H26, Entrega 4 y Alcance, «Los documentos» (lista de cinco, sin ese fichero); `docs/ADR/0038-el-umbral-de-una-clase-de-significado-es-una-tasa.md`, «Enmiendas que provoca»; `.specify/workflows/hito/workflow.yml`, paso `barrido`; `docs/WORKFLOW.md`, «Pasos globales»; `scripts/workflow/guardian-diff.sh`, modo global; constitución, «Criterio de decisión autónoma», puntos 2 y 4)

## Criterios del hito, literales

Transcripción literal de `docs/ROADMAP.md` §4, H26. El Objetivo y el Alcance se reflejan en los requisitos.

- **Entrega**:
  1. «en el job y en el sondeo, un umbral por cada clase del juez que decide, sobre las respuestas de los dos modos juntas, que es el que hace fallar el job;»
  2. «el recuento de cada modo, publicado sin decidir;»
  3. «los umbrales de `afirma_lo_no_leido` en `clases.yaml`: 0.028 en `boe-legislacion` y 0.017 en `jurisprudencia`;»
  4. «`CHANGELOG.md` (*Unreleased*) y lo que `CONTRIBUTING.md`, `docs/JURISPRUDENCIA.md`, `docs/WORKFLOW.md` y `CLAUDE.md` dicen del umbral.»
- **Controles**:
  - «`make ci` en verde, con `schema-check` y las reglas del conjunto;»
  - «la sección «Controles de umbral» del plan con una fila por umbral que decide: `afirma_lo_no_leido:<modelo>` en cada una de las dos skills, y los que ya decidían;»
  - «tests del umbral conjunto con sesiones sintéticas, en `boe-legislacion`: 3 marcadas de 108 se cumple y 4 da `fallo`, repartidas de cada forma entre los modos (3 y 0, 0 y 3, 2 y 1; 4 y 0, 2 y 2); el motivo nombra el umbral conjunto y cada respuesta con su modo; y con 104 respuestas, 2 se cumple y 3 da `fallo`;»
  - «los mismos en `jurisprudencia`: 1 de 60 se cumple y 2 da `fallo`, las dos en un modo o una en cada uno;»
  - «un test de que el recuento por modo se publica con `decide: false` y no cambia el veredicto: 3 marcadas en un solo modo, que pasan de su proporción en ese modo, dan `aprobado`;»
  - «un test de que una respuesta sin juzgar sigue dando `fallo` con su motivo, con el umbral conjunto cumplido;»
  - «un test de que una clase que solo se publica sigue por modo y sin decidir;»
  - «un test de los datos: `clases.yaml` de cada skill lleva el valor del ADR 0038, y un valor distinto hace fallar `make ci`;»
  - «un mutante por umbral: con la comparación sobre un solo modo, o con el `decide` del conjunto en falso, algún test falla;»
  - «`CHANGELOG.md` (*Unreleased*); y el job de evals en la propuesta de cambio.»
- **Aceptación**: «en el informe del job de cierre de `boe-legislacion`, `afirma_lo_no_leido:<modelo>` lleva `decide: true`, `umbral` 0.028 y `cumple: true`; los dos `afirma_lo_no_leido:<modelo>:<modo>` se publican con `decide: false`; los dos umbrales de la medida del juez, `sin_activar` y las duraciones se cumplen; ninguna respuesta queda «sin juzgar» y el veredicto es aprobado. En el de `jurisprudencia`, lo mismo con `umbral` 0.017, y `cita_sin_documento` cumplido. El de `legal-core` sigue aprobado y `red` está vacío. El informe final da los dos umbrales conjuntos como «comprobado por su control». Después, y fuera del run porque es humano: Jorge lee las respuestas marcadas del cierre, y con ello decide la release.»

Trazabilidad resumida (el detalle, en cada requisito):

| Criterio del hito | Dónde se cumple |
|---|---|
| Entrega 1, el umbral conjunto, en el job y en el sondeo | FR-001 a FR-005, FR-040, US1, US5 |
| Entrega 2, el recuento por modo | FR-010 a FR-012, US2 |
| Entrega 3, los valores | FR-020 a FR-022, US3 |
| Entrega 4, los documentos | FR-050 a FR-055, US6 |
| Lo que sigue igual | FR-030 a FR-036, US4 |
| Dentro del run | FR-060 a FR-062 |
| Controles | FR-070 a FR-079, SC-002 a SC-010, SC-013, SC-014 |
| Aceptación | SC-001, SC-011, SC-012 |

## Relación con H24 y H25

Ningún spec, plan ni informe de un hito anterior se edita: son el registro de sus runs (FR-055). La decisión de arquitectura es la del ADR 0038, ya aceptado: no hay ADR nuevo.

**Sustituye** (lo que dice el hito citado deja de valer, y vale lo de este spec):

- **«`afirma_lo_no_leido:<modelo>:<modo>` decide, con umbral 0»**, en `boe-legislacion` (H24 FR 030) y en `jurisprudencia` (H25 FR 020): pasa a ser el recuento por modo, que se publica con el umbral de la clase y `decide: false` (FR-010). Decide el umbral conjunto (FR-001).
- **«1 marcada en un modo y 0 en el otro: el veredicto es `fallo`. Los modos no se suman»** (caso límite de H25, y lo mismo en H24): los modos se suman en el umbral conjunto, y una marcada sola no da `fallo` en ninguna de las dos skills (FR-003).
- **El motivo de `afirma_lo_no_leido:<modelo>:<modo>`** (H24 FR 035; H25 FR 025): el motivo es el del umbral conjunto, y nombra las respuestas marcadas de todos los modos, cada una con su modo (FR-004).
- **Los doce elementos de `umbrales`, diez con `decide: true`**, de `jurisprudencia` (H25 FR 024) y, desde H24, de `boe-legislacion`: pasan a ser trece, nueve con `decide: true` (FR-012).
- **El `umbral: 0` de `afirma_lo_no_leido` en `clases.yaml`** (H24 FR 021; H25 FR 002): 0.028 y 0.017 (FR-020).
- **Que todo umbral de las respuestas se mida sobre las sesiones de un solo modo y decida en él** (H21 FR 043 y FR 044, y lo que `CONTRIBUTING.md` dice de ello): vale para todos menos para el de una clase del juez que decide, cuyo umbral conjunto suma los modos.

Los tests de `make ci` que hoy comprueban lo sustituido —que una marcada en un modo da `fallo`, el motivo por modo, los doce elementos con diez que deciden y el `umbral: 0` de las dos declaraciones de clases— pasan a comprobar lo que dice este spec, cada uno con su control de FR-070 a FR-079: ninguno se retira sin el que lo sustituye.

**Se queda**, sin que este spec lo vuelva a definir:

- **El voto y la regla** (H24 FR 003 a FR 011): una sesión nueva del modelo del juez por voto, la frase citada comprobada sin modelo, el voto nulo y su repetición, tres votos por orden en la clase que decide y el primero en la que solo se publica. Una respuesta queda **marcada** en una clase que decide con tres síes con su frase comprobada. El hito no añade ni quita ningún voto.
- **Qué respuestas se juzgan** (H24 FR 012 y FR 013; H25 FR 012): las del modelo que decide, en las evals que activan la skill, que terminaron con una respuesta a la pregunta. No se juzgan las del modelo informativo, y las de la eval sin binario ni servidor no cuentan en ningún umbral.
- **La respuesta «sin juzgar» y la sesión sin medir**, con su motivo «de la ejecución, no de la skill» (H24 FR 007; H7.3 FR 040 a FR 044).
- **La medida versionada del juez**, su comprobación sin modelo en `make ci` y antes de abrir ninguna sesión del job, y el informe del instrumento sin medir, que no lleva ningún umbral del juez (H24 FR 041 a FR 045; H25 FR 031 y FR 032).
- **El informe de los votos** (`juez` en `informe.json`, H24 FR 060 y FR 061; H25 FR 013).
- **El contrato de `umbrales`** (ADR 0029): `nombre`, `descripcion`, `medida`, `total`, `comparacion`, `umbral`, `cumple` y `decide`; `cumple` es la comparación en coma flotante, sin redondeos, de `medida` entre `total` (0 si `total` es 0).
- **El modelo que decide, las repeticiones y la regla por serie** (ADR 0016 y 0031), y la definición del job: sus topes, su concurrencia y sus versiones.

## User Scenarios & Testing *(mandatory)*

### User Story 1 - El cierre de un hito cae cuando la skill empeora, y no por una respuesta suelta (Priority: P1)

Quien lanza un hito lee el veredicto del job de evals de cada skill. Con `boe-legislacion` como está, que deja marcada alrededor de 1 de cada 100 respuestas, el job sale aprobado con 3 marcadas o menos de sus 108, estén en el modo que estén, y en `fallo` con 4. En `jurisprudencia`, aprobado con 1 de 60 y en `fallo` con 2. Si falla, el motivo nombra el umbral conjunto y cada respuesta marcada, con su modo y sus frases.

**Why this priority**: es el objetivo del hito. Con el umbral en cero, el cierre de cualquier hito es un sorteo y no hay release.

**Independent Test**: los tests del umbral conjunto con sesiones sintéticas, en `make ci` (SC-002, SC-003); el job de cierre (SC-001).

**Acceptance Scenarios**:

1. **Dado** un informe sintético de `boe-legislacion` con 54 respuestas juzgadas en cada modo y 3 marcadas en `afirma_lo_no_leido`, repartidas 3 y 0, 0 y 3 o 2 y 1 entre el modo `orden` y el modo `herramienta`, **Cuando** se escribe, **Entonces** `afirma_lo_no_leido:<modelo>` lleva `medida` 3, `total` 108, `umbral` 0.028, `cumple: true` y `decide: true`, y el veredicto no cambia por él (FR-001, FR-003).
2. **Dado** el mismo informe con 4 marcadas, repartidas 4 y 0 o 2 y 2, **Cuando** se escribe, **Entonces** `afirma_lo_no_leido:<modelo>` lleva `medida` 4 y `cumple: false`, el veredicto es `fallo` y su motivo nombra el umbral conjunto y cada una de las cuatro respuestas, con su modo y las frases de sus tres votos (FR-003, FR-004).
3. **Dado** un informe sintético de `boe-legislacion` con 104 respuestas juzgadas entre los dos modos, **Cuando** lleva 2 marcadas, **Entonces** el umbral conjunto se cumple; y **Cuando** lleva 3, **Entonces** no se cumple y el veredicto es `fallo` (FR-003).
4. **Dado** un informe sintético de `jurisprudencia` con 30 respuestas juzgadas en cada modo, **Cuando** lleva 1 marcada en `afirma_lo_no_leido`, **Entonces** `afirma_lo_no_leido:<modelo>` lleva `medida` 1, `total` 60, `umbral` 0.017 y `cumple: true`; y **Cuando** lleva 2, las dos en un modo o una en cada uno, **Entonces** `cumple: false` y el veredicto es `fallo`, con el motivo que nombra las dos respuestas con su modo (FR-001, FR-003, FR-004).
5. **Dado** el job de cierre de `boe-legislacion` y el de `jurisprudencia`, **Cuando** se lee `umbrales`, **Entonces** `afirma_lo_no_leido:<modelo>` lleva `decide: true`, `umbral` 0.028 o 0.017 y `cumple: true`, y el veredicto es aprobado (SC-001).

---

### User Story 2 - El recuento de cada modo se sigue viendo, y no decide (Priority: P1)

Quien lee el informe quiere saber en qué modo están las marcas. `afirma_lo_no_leido:<modelo>:<modo>` sigue en `umbrales`, con las marcadas y las juzgadas de ese modo y el umbral de la clase, y con `decide: false`. Tres marcadas en un solo modo pasan de la proporción en ese modo y el veredicto sigue siendo aprobado.

**Why this priority**: sin el recuento por modo se pierde lo que distingue el modo `orden` del modo `herramienta`; y si siguiera decidiendo, el umbral conjunto no cambiaría nada.

**Independent Test**: el test del recuento por modo, en `make ci` (SC-004); el job de cierre (SC-001).

**Acceptance Scenarios**:

1. **Dado** un informe sintético de `boe-legislacion` con 3 marcadas en el modo `herramienta` y 0 en el modo `orden`, sin ningún otro motivo de fallo, **Cuando** se escribe, **Entonces** `afirma_lo_no_leido:<modelo>:herramienta` lleva `medida` 3, `total` 54, `umbral` 0.028, `cumple: false` y `decide: false`; `afirma_lo_no_leido:<modelo>:orden`, `medida` 0, `cumple: true` y `decide: false`; y el veredicto es `aprobado`, sin ningún motivo (FR-010).
2. **Dado** un informe sintético de `jurisprudencia` con 1 marcada en un modo, **Cuando** se escribe, **Entonces** el recuento de ese modo lleva `medida` 1, `total` 30, `umbral` 0.017 y `cumple: false`, con `decide: false`, y el veredicto es `aprobado` (FR-010).
3. **Dado** el informe del job de `boe-legislacion` y el de `jurisprudencia`, **Cuando** se cuenta `umbrales`, **Entonces** cada uno lleva trece elementos, nueve con `decide: true`, y el de `legal-core`, ninguno (FR-012).

---

### User Story 3 - Los valores son los del ADR 0038, y no los elige el run (Priority: P1)

`evals/boe-legislacion/juez/clases.yaml` lleva `umbral: 0.028` en `afirma_lo_no_leido`, y `evals/jurisprudencia/juez/clases.yaml`, `umbral: 0.017`. Si alguien pone otro valor, `make ci` falla: el umbral solo lo estrecha una persona, con las cifras de la serie de cierres (ADR 0038, punto 4), y ese cambio toca también el test.

**Why this priority**: el umbral conjunto compara con este valor; con el 0 de hoy, el job seguiría como está.

**Independent Test**: el test de los datos, en `make ci` (SC-007).

**Acceptance Scenarios**:

1. **Dado** las dos declaraciones de clases con los valores del ADR 0038, **Cuando** corre `make ci`, **Entonces** pasa, con `schema-check` (FR-020, FR-021).
2. **Dado** `afirma_lo_no_leido` de `boe-legislacion` con un `umbral` distinto de 0.028, o el de `jurisprudencia` con uno distinto de 0.017, **Cuando** corre `make ci`, **Entonces** falla (FR-022).

---

### User Story 4 - Lo que no es una tasa sigue igual (Priority: P2)

El umbral conjunto no afloja nada más. Una respuesta sin juzgar sigue dando `fallo`, con su motivo de la ejecución, aunque el umbral conjunto se cumpla. `sin_activar` y `cita_sin_documento` siguen decidiendo con 0 en cada modo; los dos umbrales de la medida del juez, con 0; y las duraciones, con su cota. Una clase que solo se publica (`cuenta_su_proceso`, `afirma_que_existe`) sigue por modo, con el primer voto, y sin umbral conjunto.

**Why this priority**: el ADR 0038 cambia el umbral de lo que es significado y deja el cero de lo que tiene forma. Sin estos tests, el cambio podría arrastrar lo demás.

**Independent Test**: los tests de la respuesta sin juzgar y de la clase que solo se publica, y los de H23 a H25 que siguen como están, en `make ci` (SC-005, SC-006, SC-009).

**Acceptance Scenarios**:

1. **Dado** un informe sintético con el umbral conjunto cumplido y una respuesta que el juez dejó sin juzgar, **Cuando** se escribe, **Entonces** el veredicto es `fallo`, con el motivo «de la ejecución, no de la skill» que nombra esa respuesta, y el umbral conjunto lleva `cumple: true` (FR-030).
2. **Dado** un informe sintético de `boe-legislacion` con 3 respuestas con sí en `cuenta_su_proceso` en un modo, **Cuando** se escribe, **Entonces** `cuenta_su_proceso:<modelo>:<modo>` lleva `medida` 3 y `decide: false`, `umbrales` no lleva ningún `cuenta_su_proceso:<modelo>` sin modo, y el veredicto no cambia (FR-011).
3. **Dado** un informe sintético con una respuesta sin la skill activada en un modo, o, en `jurisprudencia`, con una cita sin documento cotejado, **Cuando** se escribe, **Entonces** `sin_activar:<modelo>:<modo>` o `cita_sin_documento:<modelo>:<modo>` lleva `cumple: false` y `decide: true`, y el veredicto es `fallo` (FR-031, FR-032).

---

### User Story 5 - El sondeo da lo mismo que hoy (Priority: P2)

Quien lanza `make evals-sondeo` sobre una skill con juez lee, por cada clase, las marcadas sobre todas las respuestas juzgadas de su sondeo. El sondeo juzga con el mismo código que el job, y su salida no gana umbral ni veredicto.

**Why this priority**: el sondeo es con lo que se mide un cambio de una skill antes de proponerlo (ADR 0038, punto 8); si dijera otra cosa que el job, mediría otra cosa.

**Independent Test**: los tests del sondeo, en `make ci`, sin cambiar lo que esperan de sus líneas (SC-014).

**Acceptance Scenarios**:

1. **Dado** un sondeo sintético de una skill con juez, con 9 respuestas juzgadas y 1 marcada en `afirma_lo_no_leido`, **Cuando** se escribe su salida, **Entonces** lleva la línea `- marcadas en afirma_lo_no_leido, con sus tres votos: 1 de 9 (11,1 %).`, ninguna línea con un umbral ni con un veredicto, y termina con `0` (FR-040).

---

### User Story 6 - Los documentos dicen el umbral que el job aplica (Priority: P3)

Quien lee `CONTRIBUTING.md` para entender un informe del job, `docs/JURISPRUDENCIA.md`, `docs/WORKFLOW.md` o `CLAUDE.md` encuentra el umbral de `afirma_lo_no_leido` como una tasa sobre los dos modos, con su valor en cada skill, y el recuento por modo como algo que se publica. `CHANGELOG.md` (*Unreleased*) registra el cambio.

**Why this priority**: no cambia lo que decide el job; un documento que dice «umbral 0» lleva a leer mal un informe aprobado con marcas.

**Independent Test**: la revisión final, por lectura contra el árbol (SC-013).

**Acceptance Scenarios**:

1. **Dado** los cinco documentos tras el hito, **Cuando** se buscan las afirmaciones de FR-050 a FR-054, **Entonces** ninguno dice que `afirma_lo_no_leido` decide con 0 o en cada modo, ni que el job todavía no aplica el ADR 0038, y cada cifra y cada nombre de umbral que dan existe en el árbol (SC-013).

---

### Edge Cases

- **3 marcadas en un solo modo de `boe-legislacion`**: son 3 de 54 en ese modo (5,6 %), por encima de 0.028, y 3 de 108 en el conjunto (2,8 %), que cumple. El recuento de ese modo lleva `cumple: false` y `decide: false`, y el veredicto es `aprobado` (FR-010).
- **2 marcadas en cada modo de `boe-legislacion`**: 4 de 108 (3,7 %); el umbral conjunto no se cumple y el veredicto es `fallo` (FR-003).
- **1 marcada en `jurisprudencia`**: 1 de 30 en su modo (3,3 %), por encima de 0.017, y 1 de 60 en el conjunto (1,7 %), que cumple. Con el umbral de H25 era `fallo`; ahora es `aprobado`, y la marca queda en el informe (ADR 0038, «En contra, y asumido»).
- **Menos respuestas juzgadas que las del conjunto de hoy**: la proporción se calcula sobre las que hay. Con 104, 3 marcadas son el 2,9 % y no cumplen; con 108, cumplen. Si faltan porque quedaron sesiones sin medir, el veredicto ya es `fallo` por su motivo (FR-030).
- **El porcentaje escrito y la comparación**: `informe.md` y el motivo escriben la medida y la condición con un decimal («3 de 108 (2,8 %)», «≤ 2,8 %»). La comparación se hace sin redondear: 3 de 104 se escribe «2,9 %» y 3 de 108, «2,8 %» (FR-003).
- **El porcentaje escrito coincide con la condición y no se cumple**: con 3 marcadas de 107 respuestas juzgadas en `boe-legislacion` (0,02804, «2,8 %») o 1 de 58 en `jurisprudencia` (0,01724, «1,7 %»), que pasan de 0.028 y de 0.017, el motivo y la fila de `informe.md` escriben los dos lados con el mismo porcentaje («3 de 107 (2,8 %), y tiene que ser ≤ 2,8 %»), sin ningún texto más: lo dicen `cumple: false` y el veredicto `fallo`. Se llega a 107 si una sesión del modelo que decide no termina y su serie pasa con 2 de 3, y a 58 si no terminan dos, de series distintas, sin otro motivo de fallo; con 59, 1 marcada todavía cumple (0,0169) (FR-003, FR-004).
- **Una respuesta sin juzgar**: sigue en el `total` del umbral conjunto y en el de su modo, y no está en ninguna `medida`. El veredicto es `fallo` por su motivo (FR-002, FR-030).
- **Sí, sí y no**: la respuesta no queda marcada y no cuenta en el umbral conjunto ni en el recuento de su modo (H24 FR 010).
- **Una respuesta marcada en `afirma_lo_no_leido` y con sí en la clase que solo se publica**: cuenta una vez en el umbral conjunto, una en el recuento de su modo y una en el de la otra clase.
- **Las respuestas de la eval sin binario ni servidor** de `boe-legislacion`: no son de ningún modo y no entran en el umbral conjunto, como no entran en ninguno (FR-002).
- **El instrumento sin medir**: el job termina sin abrir ninguna sesión y `umbrales` no lleva ningún umbral del juez, tampoco el conjunto (H24 FR 043; H25 FR 032).
- **Una skill sin juez** (`legal-core`): no tiene clases, ni umbral conjunto, ni recuento por modo; `umbrales` es `[]` (FR-012).
- **Ninguna respuesta juzgada**: `total` es 0, la proporción es 0 y el umbral conjunto se cumple, como todo umbral con `total` 0 (contrato del ADR 0029); el veredicto lo deciden los demás motivos.
- **Un plan con un solo modo**: el umbral conjunto lleva la medida y el total de ese modo, y es el que decide; el recuento de ese modo se publica (FR-001).
- **El cierre sale en rojo por `afirma_lo_no_leido:<modelo>`**: ninguna reparación toca una skill ni un umbral; queda en el informe final para la persona, con las respuestas marcadas y sus frases (FR-061).

## Requirements *(mandatory)*

### Functional Requirements

#### El umbral conjunto

Cada uno es un elemento de `umbrales` de `informe.json` de una skill con juez, con el contrato del ADR 0029.

- **FR-001**: Por cada clase del juez con `decide: true` en `clases.yaml`, `umbrales` MUST llevar exactamente un elemento de nombre `<clase>:<modelo>`, sin modo, con: `medida`, las respuestas del modelo que decide que el juez deja marcadas en esa clase con sus tres votos, en todos los modos del plan; `total`, las respuestas juzgadas de esos modos (FR-002); `comparacion`, `"<="`; `umbral`, el de la clase en `clases.yaml`, que es una proporción; `decide: true`; y una `descripcion` que dice que cuenta las respuestas de todos los modos. Vale para toda skill con juez, sin ningún caso por skill: hoy, `afirma_lo_no_leido:<modelo>` en `boe-legislacion` y en `jurisprudencia`. **Control**: si la proporción pasa del umbral, el veredicto es `fallo` y el job de evals de esa skill sale en rojo (FR-004).
- **FR-002**: El `total` del umbral conjunto MUST ser la suma de los totales de los recuentos por modo de esa clase: las respuestas medidas del modelo que decide en las evals que activan la skill, como hoy (H24 FR 012; H25 FR 012). Su `medida` MUST ser la suma de las medidas de esos recuentos. Una respuesta sin juzgar sigue en el `total` y no está en ninguna `medida`. MUST NOT contar las respuestas del modelo informativo ni las de la eval sin binario ni servidor. Con el conjunto de evals de hoy y todas las sesiones medidas son 108 respuestas en `boe-legislacion` (54 por modo) y 60 en `jurisprudencia` (30 por modo). Comprobable: FR-072, FR-073, FR-075.
- **FR-003**: El umbral conjunto MUST cumplirse si y solo si `medida` entre `total` es menor o igual que `umbral`, comparado en coma flotante de doble precisión y sin redondeos (0 si `total` es 0), y su `cumple` MUST decirlo. No depende de cómo se reparten las marcadas entre los modos. Con los valores de FR-020: en `boe-legislacion`, 3 de 108 cumple y 4 de 108 no; 2 de 104 cumple y 3 de 104 no; en `jurisprudencia`, 1 de 60 cumple y 2 de 60 no. **Control**: cuando no se cumple, el job de evals de esa skill sale en rojo (FR-004). Comprobable: FR-072, FR-073.
- **FR-004**: Un umbral conjunto que no se cumple MUST poner el veredicto en `fallo`, y el job MUST salir en rojo. Su motivo MUST nombrar el umbral con su medida y su condición, como el de todo umbral que decide («umbral `<clase>:<modelo>`: 4 de 108 (3,7 %), y tiene que ser ≤ 2,8 %»), y detrás, cada respuesta marcada de todos los modos: con el nombre de su modo, su sesión y las frases de los tres votos que la marcan, las del modo `orden` delante y, dentro de un modo, en orden de sesión. El modo de cada respuesta MUST leerse en el motivo sin deducirlo del nombre de la sesión. Uno que se cumple MUST NOT cambiar el veredicto ni dar un motivo. Comprobable: FR-072, FR-073.
- **FR-005**: El umbral conjunto MUST ir, en `umbrales`, detrás de los umbrales de las respuestas de los modos y delante de los dos de la medida del juez; con varias clases que deciden, en el orden de `clases.yaml`. `informe.md` lo da en su tabla «Umbrales» como una fila más, con la medida y la condición escritas como las de todo umbral con `total`. Comprobable: el test de los trece elementos los comprueba en su orden (FR-075).

#### El recuento por modo

- **FR-010**: `<clase>:<modelo>:<modo>` de una clase que decide MUST seguir en `umbrales`, uno por modo del plan, con su `medida` —las marcadas con tres votos de ese modo—, su `total` —las juzgadas de ese modo—, `comparacion` `"<="` y el mismo `umbral` que el conjunto, y MUST llevar `decide: false`. Su `cumple` sigue siendo la comparación del contrato, y solo informa: MUST NOT cambiar el veredicto ni dar un motivo, tampoco cuando pasa de su proporción. No es un requisito con umbral: el hito lo deja publicado. Comprobable: FR-074.
- **FR-011**: Una clase con `decide: false` (`cuenta_su_proceso` en `boe-legislacion`, `afirma_que_existe` en `jurisprudencia`) MUST seguir como está: un elemento por modo, con el primer voto, su `umbral` de `clases.yaml` y `decide: false`. MUST NOT tener umbral conjunto. Comprobable: FR-076.
- **FR-012**: Con el plan del job, `umbrales` de `boe-legislacion` MUST llevar exactamente trece elementos, nueve con `decide: true`: por modo, `sin_activar`, `afirma_lo_no_leido` y `cuenta_su_proceso`; el umbral conjunto; los dos de la medida del juez; y por modo, `duracion_de_las_sesiones` y `duracion_del_juez`. Los de `jurisprudencia`, exactamente trece, nueve con `decide: true`: por modo, `sin_activar`, `afirma_lo_no_leido`, `afirma_que_existe` y `cita_sin_documento`; el umbral conjunto; los dos de la medida; y por modo, `duracion_del_juez`. No deciden los recuentos por modo de ninguna clase. `umbrales` de `legal-core` MUST seguir siendo `[]`. **Control**: un test de `make ci` falla con otro número de elementos o con otro `decide` en alguno (FR-075).

#### Los valores

- **FR-020**: `afirma_lo_no_leido` MUST llevar `umbral: 0.028` en `evals/boe-legislacion/juez/clases.yaml` —como mucho 3 de 108— y `umbral: 0.017` en `evals/jurisprudencia/juez/clases.yaml` —como mucho 1 de 60—, con `decide: true`. Son los del ADR 0038 y el run MUST NOT elegir otros. `cuenta_su_proceso` y `afirma_que_existe` MUST seguir con `decide: false` y `umbral: 0`. El cambio va en tareas `[datos]`. **Control**: con otro valor, `make ci` falla (FR-022).
- **FR-021**: El esquema de la declaración de clases, que ya admite una proporción entre 0 y 1, MUST NOT cambiar, ni nada de `schemas/`. `clases.yaml` no es uno de los cuatro ficheros que `evidencias/` copia: `evidencias/` MUST NOT cambiar, las cuatro copias de cada carpeta del juez siguen idénticas (`TestCopiasDelJuez`) y la medida versionada del juez, que no depende de `clases.yaml`, MUST NOT repetirse ni escribirse. Comprobable: `schema-check`, `TestCopiasDelJuez` y la comprobación sin modelo de la medida versionada siguen pasando en `make ci` sin cambiar lo que esperan.
- **FR-022**: Un test de `make ci` MUST comprobar que `clases.yaml` de cada una de las dos skills lleva en `afirma_lo_no_leido` el valor de FR-020. **Control**: con un valor distinto en cualquiera de las dos, `make ci` falla. Comprobable: FR-077.

#### Lo que sigue igual

- **FR-030**: Una respuesta que el juez deja sin juzgar y una sesión sin medir MUST seguir dando `fallo`, cada una con su motivo, que lleva el prefijo «de la ejecución, no de la skill» y la nombra, también con el umbral conjunto cumplido. Comprobable: FR-075.
- **FR-031**: `sin_activar:<modelo>:<modo>` MUST seguir como está, por modo, con `umbral` 0 y `decide: true`, en las dos skills. **Control**: con una respuesta sin la skill activada, el veredicto es `fallo` y el job de evals sale en rojo.
- **FR-032**: `cita_sin_documento:<modelo>:<modo>` de `jurisprudencia` MUST seguir como está, por modo, con `umbral` 0 y `decide: true`. **Control**: con una respuesta que cuenta, el veredicto es `fallo` y el job de evals sale en rojo.
- **FR-033**: Los dos umbrales de la medida del juez, `medida_del_juez:afirma_lo_no_leido:defectos_sin_marcar` y `medida_del_juez:afirma_lo_no_leido:correctos_marcados`, MUST seguir como están en las dos skills, con `umbral` 0 y `decide: true`, y con los recuentos de la medida versionada: 0 de 212 y 0 de 47 en `boe-legislacion`, 0 de 125 y 0 de 124 en `jurisprudencia`. **Control**: con un recuento distinto de 0, `make ci` falla y el job termina con `fallo` sin abrir ninguna sesión, como hoy.
- **FR-034**: `duracion_de_las_sesiones:<modo>` de `boe-legislacion`, con su objetivo de 900 s, y `duracion_del_juez:<modo>` de las dos skills, con 900 s, MUST seguir como están, con `decide: true`. **Control**: por encima de su cota, el veredicto es `fallo`, con un motivo de la ejecución, y el job de evals sale en rojo.
- **FR-035**: La regla de los votos, la rúbrica, el esquema de la respuesta, los casos, la medida, el modelo del juez y la versión de Claude Code de sus votos MUST NOT cambiar. El hito MUST NOT añadir ni quitar ninguna sesión ni ningún voto: los peores casos que `TestDefinicionDelJob` recalcula, y la definición del job, no cambian. Comprobable: `TestDefinicionDelJob` y `TestCopiasDelJuez` siguen pasando en `make ci` sin cambiar lo que esperan.
- **FR-036**: `boe-legislacion`, `jurisprudencia` y `legal-core` MUST seguir como están: sus `SKILL.md`, sus `references/` y sus evals. `legal-core` MUST seguir sin juez. Comprobable: las reglas del conjunto pasan en `make ci`, y el diff del hito no toca `skills/` ni ninguna eval.

#### El sondeo

- **FR-040**: `make evals-sondeo` MUST juzgar con el mismo código que el job, y su salida MUST seguir dando, por cada clase de `clases.yaml`, las marcadas sobre todas las respuestas juzgadas del sondeo —con sus tres votos en la que decide y con el primero en la que solo se publica—, como hoy. MUST NOT llevar un umbral, decir si se cumple ni dar un veredicto, y MUST seguir terminando con `0` sean cuales sean sus recuentos. Comprobable: los tests del sondeo que ya hay siguen pasando sin cambiar las líneas que esperan (SC-014).

#### Documentos

- **FR-050**: `CHANGELOG.md` (*Unreleased*) MUST registrar el cambio para quien lee un informe del job: que `afirma_lo_no_leido` decide con una tasa sobre las respuestas de los dos modos juntas, con el nombre del umbral conjunto y su valor en cada skill (0.028, como mucho 3 de 108; 0.017, como mucho 1 de 60); que el recuento por modo se publica con `decide: false`; que `umbrales` pasa de doce elementos con diez que deciden a trece con nueve en cada una de las dos skills; y la forma del motivo. Ninguna línea de *Unreleased* MUST seguir diciendo que `afirma_lo_no_leido` decide con 0, que decide en cada modo o que una marcada en un modo da `fallo`. Las secciones de las versiones ya publicadas MUST NOT cambiar. Comprobable: SC-013.
- **FR-051**: `CONTRIBUTING.md`, en lo que dice de la carpeta del juez y del job de evals, MUST decir el umbral nuevo donde hoy dice el de antes: que `afirma_lo_no_leido` decide con la tasa de su `clases.yaml` sobre los dos modos juntos, con su valor en cada skill; las dos tablas de `umbrales`, con la fila del umbral conjunto, el `umbral` y el `decide` del recuento por modo, los trece elementos, los nueve que deciden y su orden; el ejemplo del motivo de una clase del juez; y que una marcada en un modo ya no da `fallo` por sí sola. Lo que dice del sondeo no cambia (FR-040).
- **FR-052**: `docs/JURISPRUDENCIA.md`, en «Lo que H25 deja hecho», MUST dejar de decir que resumir o caracterizar una sentencia que no se tenía delante lo decide el juez «con umbral 0 en cada modo» y que una respuesta así deja el job en rojo, y MUST decir lo que vale: como mucho 1 marcada de las 60 de los dos modos juntos, y con 2 el job sale en rojo.
- **FR-053**: `docs/WORKFLOW.md`, en lo que dice del cierre, MUST dejar de decir que en `boe-legislacion` deciden diez umbrales con `afirma_lo_no_leido:<modelo>:<modo>` entre ellos, y MUST decir lo que vale: deciden nueve, con `afirma_lo_no_leido:<modelo>`, y el recuento por modo solo se publica.
- **FR-054**: `CLAUDE.md` MUST dejar de decir, donde lo dice como algo que vale hoy, que `afirma_lo_no_leido` decide con umbral 0 en cada modo y que el job sigue con el umbral 0 hasta H26, y MUST decir lo que vale: el job aplica la tasa del ADR 0038 sobre los dos modos juntos. Las cifras de los cierres pasados que da (0 de 54, 2 de 108…) MUST NOT cambiar.
- **FR-055**: La corrección de FR-051 a FR-054 vale para toda afirmación sobre el umbral de `afirma_lo_no_leido`, o sobre cuántos umbrales llevan y cuáles deciden en las dos skills, que el hito deje falsa en esos cuatro ficheros, y para nada más: MUST NOT reescribir lo que sigue siendo cierto, prometer ninguna release ni añadir un test que ate ese texto al código. Cada cifra sale del árbol y no de memoria. MUST NOT tocarse `docs/ROADMAP.md` —tampoco sus secciones de H24 y de H25—, ningún ADR, ni el spec, el plan o los informes de un hito anterior. Comprobable: SC-013.

#### Dentro del run

- **FR-060**: Dentro del run, ninguna sesión MUST abrir el juez ni una sesión de evals (ADR 0032): los tests usan sesiones sintéticas y votos grabados. El run MUST NOT lanzar un sondeo ni la medida del juez, y solo pone la etiqueta `evals`.
- **FR-061**: Ninguna corrección ni reparación del run MUST cambiar una skill (`SKILL.md`, `references/`, evals), tampoco si el cierre da marcas; ni cumplir un umbral cambiando su valor, dejándolo en `decide: false`, sacando respuestas de su total, o cambiando la rúbrica, los casos o la medida (ADR 0029; ADR 0038, punto 4). Si el cierre sale en rojo por `afirma_lo_no_leido:<modelo>`, queda en el informe final para la persona, con las respuestas marcadas y sus frases. Un rojo del cierre por un defecto de lo que este hito entrega —el código del informe o un documento— sí se repara. Comprobable: lo leen los jueces de la revisión final contra el diff de cada reparación.
- **FR-062**: El hito MUST NOT cambiar el binario fuera de `internal/evals`, ni `scripts/workflow/`, ni la constitución, ni la definición del job. Lo que cambie bajo `testdata/` va en tareas `[datos]`. Comprobable: el diff del hito, que leen el guardián de diff y la revisión final.

#### Controles y Definition of Done

- **FR-070**: `make ci` MUST quedar en verde, con `schema-check` y las reglas del conjunto.
- **FR-071**: La sección «Controles de umbral» del plan MUST tener una fila por umbral que decide, cada una en `evals:<skill>:<nombre>`: `afirma_lo_no_leido:<modelo>` en cada una de las dos skills (FR-001), y los que ya decidían —`sin_activar:<modelo>:<modo>` en cada modo de las dos skills (FR-031), `cita_sin_documento:<modelo>:<modo>` en cada modo de `jurisprudencia` (FR-032), los dos de la medida del juez de cada skill (FR-033), `duracion_de_las_sesiones:<modo>` en cada modo de `boe-legislacion` y `duracion_del_juez:<modo>` en cada modo de las dos (FR-034)—: nueve por skill. MUST NOT tener fila el recuento por modo de ninguna clase, que no decide. La fila del test de los datos (FR-022) y la del recuento de elementos (FR-012) las pone el plan, con su control `ci:`.
- **FR-072**: Los tests del umbral conjunto con sesiones sintéticas de `boe-legislacion` MUST cubrir: 3 marcadas de 108 se cumple, repartidas 3 y 0, 0 y 3 y 2 y 1 entre los modos; 4 da `fallo`, repartidas 4 y 0 y 2 y 2; el motivo nombra el umbral conjunto y cada respuesta con su modo; y con 104 respuestas, 2 se cumple y 3 da `fallo`.
- **FR-073**: Los mismos tests con sesiones sintéticas de `jurisprudencia` MUST cubrir: 1 marcada de 60 se cumple, y 2 da `fallo`, las dos en un modo o una en cada uno, con el motivo que nombra el umbral conjunto y cada respuesta con su modo.
- **FR-074**: Un test MUST comprobar que el recuento por modo se publica con `decide: false` y no cambia el veredicto: 3 marcadas en un solo modo, que pasan de su proporción en ese modo, dan `aprobado`.
- **FR-075**: Un test MUST comprobar que una respuesta sin juzgar sigue dando `fallo` con su motivo, con el umbral conjunto cumplido; y otro, los trece elementos de `umbrales` de cada una de las dos skills, en su orden (FR-005), con los nueve que deciden (FR-012).
- **FR-076**: Un test MUST comprobar que una clase que solo se publica sigue por modo y sin decidir, y que no tiene umbral conjunto.
- **FR-077**: Un test de los datos MUST comprobar que `clases.yaml` de cada skill lleva el valor del ADR 0038, y un valor distinto MUST hacer fallar `make ci` (FR-022).
- **FR-078**: Un mutante por umbral conjunto: con la comparación hecha sobre las respuestas de un solo modo, y con el `decide` del conjunto en falso, cada uno por separado, algún test de `make ci` de FR-072 y alguno de FR-073 MUST fallar. La tarea que lo comprueba dice qué test falla con cada mutante, y ningún mutante queda en el árbol. Es la prueba de que el control de FR-001 falla por encima de su umbral (ADR 0029).
- **FR-079**: `CHANGELOG.md` (*Unreleased*) MUST llevar lo de FR-050, y el job de evals MUST ejecutarse en la propuesta de cambio (lo hace el workflow tras la revisión final); su informe da lo que pide SC-001.

### Key Entities

- **Clase que decide**: una clase de `clases.yaml` con `decide: true`. Hoy, `afirma_lo_no_leido` en las dos skills con juez.
- **Respuesta juzgada**: una respuesta del modelo que decide, en una eval que activa la skill, que terminó con una respuesta a la pregunta.
- **Respuesta marcada**: en una clase que decide, la respuesta juzgada con tres síes del juez, cada uno con su frase comprobada.
- **Umbral conjunto**: el elemento `<clase>:<modelo>` de `umbrales`: las marcadas de todos los modos sobre las juzgadas de todos los modos, frente a la proporción de la clase. Decide.
- **Recuento por modo**: el elemento `<clase>:<modelo>:<modo>`: las marcadas y las juzgadas de un modo, con el mismo umbral. Se publica.
- **Tasa de la clase**: el `umbral` de la clase en `clases.yaml`, una proporción entre 0 y 1 que fija el ADR 0038 y que solo estrecha una persona.

## Success Criteria *(mandatory)*

### Measurable Outcomes

- **SC-001**: En el informe del job de cierre de `boe-legislacion`:
  - `afirma_lo_no_leido:<modelo>` lleva `decide: true`, `umbral` 0.028 y `cumple: true`;
  - los dos `afirma_lo_no_leido:<modelo>:<modo>` se publican con `decide: false`;
  - los dos umbrales de la medida del juez, `sin_activar` y las duraciones se cumplen;
  - ninguna respuesta queda «sin juzgar» y el veredicto es aprobado.

  En el de `jurisprudencia`, lo mismo con `umbral` 0.017, y `cita_sin_documento` cumplido. El de `legal-core` sigue aprobado y `red` está vacío.

  **Control**: el job de evals de cada skill sale en rojo, con el veredicto `fallo`, en cualquiera de estos casos: las marcadas de los dos modos pasan de la tasa de la clase (FR-001); la medida versionada no corresponde o no se cumple (FR-033); `sin_activar` o `cita_sin_documento` no se cumplen (FR-031, FR-032); una duración pasa de su cota (FR-034); alguna respuesta queda sin juzgar o alguna sesión sin medir (FR-030); una serie que decide no pasa; o llega una petición a la red. El cierre del workflow cuenta cada rojo.
- **SC-002**: En `boe-legislacion`, con 108 respuestas juzgadas, el umbral conjunto se cumple con 3 marcadas en los 3 repartos (3 y 0, 0 y 3, 2 y 1) y da `fallo` con 4 en los 2 repartos (4 y 0, 2 y 2), con un motivo que nombra el umbral conjunto y las 4 respuestas con su modo; y con 104 respuestas, se cumple con 2 y da `fallo` con 3: 7 de 7 casos. **Control**: los tests de FR-072 fallan en `make ci`.
- **SC-003**: En `jurisprudencia`, con 60 respuestas juzgadas, el umbral conjunto se cumple con 1 marcada y da `fallo` con 2 en los 2 repartos (2 y 0, 1 y 1), con un motivo que nombra el umbral conjunto y las 2 respuestas con su modo: 3 de 3 casos. **Control**: los tests de FR-073 fallan en `make ci`.
- **SC-004**: Con 3 marcadas en un solo modo de `boe-legislacion`, el recuento de ese modo lleva `cumple: false` y `decide: false`, y el veredicto es `aprobado`, con 0 motivos. **Control**: el test de FR-074 falla en `make ci`.
- **SC-005**: Con el umbral conjunto cumplido y 1 respuesta sin juzgar, el veredicto es `fallo`, con 1 motivo de la ejecución que la nombra. **Control**: el test de FR-075 falla en `make ci`.
- **SC-006**: Una clase que solo se publica lleva 1 elemento por modo, con `decide: false`, y 0 elementos sin modo; con 3 respuestas con sí en un modo, el veredicto no cambia. **Control**: el test de FR-076 falla en `make ci`.
- **SC-007**: `afirma_lo_no_leido` lleva `umbral` 0.028 en `boe-legislacion` y 0.017 en `jurisprudencia`: 2 de 2. **Control**: el test de los datos falla en `make ci` con un valor distinto en cualquiera de las dos (FR-077).
- **SC-008**: Con cada uno de los 2 mutantes —la comparación sobre un solo modo y el `decide` del conjunto en falso— falla algún test de `boe-legislacion` y alguno de `jurisprudencia`: 4 de 4. **Control**: los tests de FR-072 y FR-073, en `make ci`; lo comprueban la tarea de FR-078 y la revisión final.
- **SC-009**: `umbrales` de `boe-legislacion` y de `jurisprudencia` llevan 13 elementos cada uno, 9 con `decide: true`, y el de `legal-core`, 0. **Control**: el test de FR-075 falla en `make ci`.
- **SC-010**: `make ci` en verde, con `schema-check` y las reglas del conjunto, y `CHANGELOG.md` (*Unreleased*) con el umbral nuevo. **Control**: `make ci`.
- **SC-011**: El informe final da los dos umbrales conjuntos como «comprobado por su control». **Control**: lo decide sin modelo el informe final del workflow, que solo lo dice de un umbral con fila en «Controles de umbral» del plan (FR-071) que el job publica, cumple y hace fallar el job; si el job no lo publica, no se cumple o lleva `decide: false`, el informe final lo da como sin verificar o como no cumplido.
- **SC-012**: Después del run, y fuera de él porque es humano: Jorge lee las respuestas marcadas del cierre, y con ello decide la release. Lo mide una persona: no tiene control en el run.
- **SC-013**: `CONTRIBUTING.md`, `docs/JURISPRUDENCIA.md`, `docs/WORKFLOW.md`, `CLAUDE.md` y *Unreleased* de `CHANGELOG.md` llevan 0 afirmaciones de que `afirma_lo_no_leido` decide con 0, de que decide en cada modo, de que una marcada en un modo da `fallo` o de que el job no aplica todavía el ADR 0038; y cada cifra, fichero, test y umbral que nombran al decir lo nuevo existe en el árbol. **Control**: la revisión final, por lectura contra el árbol; ningún test de `make ci` ata ese texto al código, y no se añade uno (FR-055).
- **SC-014**: La salida del sondeo lleva, por cada clase, 1 línea con las marcadas sobre todas las respuestas juzgadas, y 0 líneas con un umbral o un veredicto. **Control**: los tests del sondeo que ya hay fallan en `make ci` si sus líneas cambian (FR-040).

## Uso, de fuera adentro

Cada salida del hito, desde quien la consume (criterio de uso, ADR 0028). El hito dice lo que gana cada skill: no cambia, y el verde y el rojo de sus evals vuelven a decir algo. Ninguna skill pide ninguna de estas salidas, y ningún paso de un `SKILL.md` las lee: son 0 por pregunta. El hito no cambia el binario fuera de `internal/evals`, y nada de lo que entrega crece con el uso del kit —cientos de normas y miles de bloques consultados en meses—: cada job parte de las evals del repositorio y mide de nuevo sobre su commit.

| Salida | Quién la pide, cuántas veces y qué hace con ella | Tamaño | Cuándo deja de darse cada señal |
|---|---|---|---|
| `umbrales` del informe de una skill con juez, en `informe.json` y en la tabla «Umbrales» de `informe.md` (FR-001, FR-010, FR-012) | El job, una vez, que decide con ellos el veredicto; el informe final del workflow, una vez por medición del cierre, que los lee sin modelo (SC-011); y la persona que lee el informe. | 13 elementos por skill, uno más que hoy. Los 12 de hoy miden entre 238 y 388 bytes: 3 482 bytes en `boe-legislacion` y 3 757 en `jurisprudencia` en el último cierre; con el conjunto, unos 3,8 KB y 4,1 KB. Fijo: depende de las clases y de los modos, no del uso. | Cada job los mide de nuevo. El `cumple: false` del umbral conjunto deja de darse en el primer job en que las marcadas de los dos modos caben en la tasa. El `cumple: false` de un recuento por modo, que no decide, deja de darse cuando las marcadas de ese modo caben en ella. |
| El motivo de `fallo` del umbral conjunto (FR-004) | La persona, y la reparación del cierre, que lo lee y no cambia por él ninguna skill (FR-061). Uno por umbral conjunto incumplido: como mucho uno por skill y por job. | El umbral, su medida y su condición, y por respuesta marcada, su modo, su sesión y tres frases: el del último cierre, con 2 respuestas, mide 953 bytes, unos 0,5 KB por respuesta. Con las 108 y las 60 respuestas juzgadas, el más corto que puede darse lleva 4 respuestas en `boe-legislacion` (unos 2 KB) y 2 en `jurisprudencia` (1 KB); con menos juzgadas puede llevar menos: 3, de 107 o de 104, y 1, de 58. Acotado por las respuestas juzgadas, 108 y 60, no por el uso. | Deja de darse en el primer job en que el umbral conjunto se cumple. No se acumula de un job a otro. |
| Las líneas del juez en la salida de `make evals-sondeo` (FR-040) | La persona que lanza el sondeo, una vez por sondeo. | Una línea por clase, como hoy: dos, de unos 70 bytes. | Una por sondeo. No cambia con este hito. |
| El `umbral` de la clase en `clases.yaml` (FR-020) | El job y `make ci`, una vez por ejecución; y la persona que lo estrecha. | Un número por clase; dos clases por skill. | No da señales. Cambia solo con un diff de una persona. |
| Los cinco documentos (FR-050 a FR-054) | Quien contribuye o lee un informe del job. | Las frases y las filas que hoy dicen el umbral de antes; no crecen con el uso. | No dan señales. |

## Fuera de alcance

Del hito, literal:

- «cambiar `boe-legislacion`, `jurisprudencia` o `legal-core`, también si el cierre da marcas: mejorar una skill es trabajo aparte, con su sondeo antes (ADR 0038, punto 8). Si el cierre sale en rojo por `afirma_lo_no_leido`, ninguna reparación toca una skill ni un umbral: queda en el informe final para la persona;»
- «cambiar los valores del umbral, la rúbrica, los casos, la medida, el modelo del juez o su versión de Claude Code;»
- «la serie de cierres y la tasa acumulada (ADR 0038, puntos 4 y 6) y la línea de base sin la skill (punto 7): la serie la escribe y la enseña `scripts/workflow/`, que un hito no toca, y van en una propuesta de proceso aparte; la línea de base la mide una persona;»
- «que `cuenta_su_proceso` o `afirma_que_existe` decidan; extender el juez a `legal-core`;»
- «cambiar el modelo que decide, las repeticiones, la regla por serie o las evals (ADR 0016 y 0031);»
- «el binario fuera de `internal/evals`, y `scripts/workflow/`.»

De lo que el hito no especifica (constitución, «Criterio de decisión autónoma», punto 2):

- Un umbral conjunto para lo que no es una clase del juez: `sin_activar`, `cita_sin_documento` y las duraciones siguen por modo.
- Que el recuento por modo lleve otro umbral que el de la clase, o que deje de publicarse.
- Un umbral, un veredicto o una línea nueva en la salida del sondeo.
- Cambiar el contrato de `umbrales` (ADR 0029), `scripts/workflow/informe.sh` o lo que el informe final dice de un umbral.
- Cambiar el esquema de la declaración de clases, o una clave nueva en `clases.yaml`.
- Publicar en el informe del job la tasa acumulada, la clase de pregunta de cada respuesta marcada o la probabilidad de un rojo sin motivo.
- Mover la frontera de la rúbrica (ADR 0038, opción 5), o meter en los casos una marca que una persona da por errónea.
- La viñeta de H26 de «Estado del repositorio» de `CLAUDE.md` y su frase del siguiente hito: las escribe la persona al fusionar, como en los hitos anteriores. `README.md`, `docs/USO.md`, la web y `docs/ROADMAP.md`.
- Un test que ate lo que dicen los documentos al código, y reescribir de ellos lo que sigue siendo cierto (FR-055).
- Editar el spec, el plan o los informes de un hito anterior; un ADR nuevo; y la release, que decide una persona.
- Corregir el comentario de `.github/workflows/evals.yml` (hoy en sus líneas 152 a 154) que dice que el informe de `jurisprudencia` lleva «doce umbrales, diez que deciden» y que `afirma_lo_no_leido` decide por modo: tras el hito queda falso a sabiendas (son trece, nueve que deciden, y el recuento por modo solo se publica: FR-012) en un fichero que el hito no toca. El diff del hito no lleva nada bajo `.github/`, ninguna tarea declara ese fichero y ni el barrido, ni un corrector, ni una reparación del cierre lo cambian. **Pendiente para la persona**: se corrige en un cambio fuera del hito, o corrigiendo la entrada y relanzando.

## Assumptions

- **Lo que H24 y H25 ya definen vale sin volver a escribirse** («Relación con H24 y H25», «Se queda»). El hito dice «como hoy» de qué respuestas se juzgan, de la regla de los votos y del motivo; este spec solo define lo que cambia.
- **La posición del umbral conjunto en `umbrales`** (FR-005): el hito no la dice. Va detrás de los umbrales de las respuestas de los modos y delante de los dos de la medida, porque es de las respuestas y no es de un modo. No cambia ningún veredicto.
- **El modo de cada respuesta en el motivo** (FR-004): el hito pide que el motivo nombre «cada respuesta marcada, con su modo y sus frases». Hoy el modo va en el nombre del umbral, que deja de llevarlo; y el nombre de una sesión solo lo deja ver en el modo `herramienta`. Por eso el spec pide que se lea sin deducirlo. El texto exacto lo fija el plan.
- **El recuento por modo conserva su `cumple`** (FR-010): el hito dice que sigue «con su `medida`, su `total` y el mismo `umbral`», y el contrato del ADR 0029 pide `cumple` en todo elemento. Con `decide: false` solo informa, como el de `cuenta_su_proceso` hoy.
- **Trece elementos, nueve que deciden** (FR-012) sale de contar: los doce de hoy más el conjunto, y los dos recuentos por modo de `afirma_lo_no_leido` dejan de decidir.
- **Nueve filas por skill en «Controles de umbral»** (FR-071) sale de «una fila por umbral que decide […] y los que ya decidían». El informe final da por no verificable un control cuyo umbral lleva `decide: false` (`scripts/workflow/informe.sh`, leído y no ejecutado en esta sesión): por eso el recuento por modo no tiene fila.
- **Qué entra en los documentos** (FR-050 a FR-055): el hito pide que digan el umbral nuevo «donde hoy dicen «umbral 0» de `afirma_lo_no_leido`» y, en la Entrega, «lo que […] dicen del umbral». Se toma como en H25 (su clarificación sobre `CONTRIBUTING.md`, criterio a): toda afirmación de esa clase que el hito deje falsa, y nada más. Las que se han leído en esta sesión: en `CONTRIBUTING.md`, «decide, con umbral 0» de las dos skills, las dos tablas de `umbrales` con «son doce, diez de ellos decidiendo», el ejemplo del motivo con «≤ 0,0 %» y «una respuesta marcada […] en un modo da `fallo` aunque en el otro no haya ninguna»; en `docs/JURISPRUDENCIA.md`, «umbral 0 en cada modo»; en `docs/WORKFLOW.md`, «deciden diez» con `afirma_lo_no_leido:<modelo>:<modo>`; en `CLAUDE.md`, «decide con umbral 0 en cada modo» y «El job sigue con el umbral 0 hasta H26»; y en *Unreleased* de `CHANGELOG.md`, las entradas de H24 y de H25, que no están en ninguna versión publicada.
- **`CHANGELOG.md`** (FR-050): las entradas de H24 y de H25 siguen en *Unreleased* porque la última release, v0.5.0, lleva hasta H22. *Unreleased* describe lo que llevará la siguiente release, y H25 ya reescribió en ella lo que H23 decía y dejó de valer.
- **Los mutantes** (FR-078) los aplica una tarea del run sobre el código, uno cada vez, y los retira: el hito pide que «algún test falla» con cada uno. Cómo se aplican lo fija el plan.
- **El sondeo no cambia su salida** (FR-040): corre en un solo modo, y sus líneas ya dan las marcadas de cada clase sobre todas sus respuestas, sin umbral (leído en `internal/evals/sondeo.go` y en `CONTRIBUTING.md`). «Usa el mismo código» se cumple si el cambio del informe no rompe sus tests.
- **Sin aceptación e2e del binario**: el hito no cambia nada que vea quien usa `kitlegal`. Ningún guion congelado de un hito anterior nombra `afirma_lo_no_leido` (buscado en `specs/*/aceptacion/` en esta sesión). Qué describe la entrega lo fija el plan; la aceptación del hito es el job de cierre (SC-001).
- **Lo comprobado en esta sesión**, sin ejecutar ningún test ni `make ci`:
  - las dos `clases.yaml` llevan hoy `umbral: 0` en sus dos clases, y `schemas/juez-clases.yaml.json` admite un número entre 0 y 1;
  - ni `evidencias/adr-0037/` ni `evidencias/adr-0037-jurisprudencia/` tienen un `clases.yaml`;
  - los informes del último cierre versionado (`specs/020-h25-el-juez-de/gates/evals/`) llevan doce elementos en `umbrales` en cada una de las dos skills, con totales de 54 y de 30 por modo, y `[]` en `legal-core`; sus tamaños y el de su motivo se han medido con `jq`;
  - las proporciones se han calculado a mano: 3/108 = 0,0278 y 4/108 = 0,0370; 2/104 = 0,0192 y 3/104 = 0,0288; 1/60 = 0,0167 y 2/60 = 0,0333; 3/54 = 0,0556 y 1/30 = 0,0333;
  - `TestEvalsDelRepositorio` comprueba hoy el `Umbral: 0` de las dos declaraciones.

  Las demás cifras (5 de 536, el 36 %, los cuatro sondeos) son del hito y del ADR 0038.
- **Los nombres técnicos que aparecen** (`clases.yaml`, `umbrales`, `informe.json`, `informe.md`, `TestCopiasDelJuez`, `TestDefinicionDelJob`, `TestEvalsDelRepositorio`, `make evals-sondeo`) existen en el repositorio o los fija el hito. Quedan para el plan: los nombres de los tests nuevos, el texto de la `descripcion` del umbral conjunto, cómo escribe el motivo el modo de cada respuesta y cómo se aplican los mutantes.
