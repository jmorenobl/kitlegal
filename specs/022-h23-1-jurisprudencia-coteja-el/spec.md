# Feature Specification: H23.1 · `jurisprudencia` coteja el documento antes de decir nada de él, y un umbral sin modelo lo hace cumplir

**Feature Branch**: `022-h23-1-jurisprudencia-coteja-el`

**Created**: 2026-10-16

**Status**: Draft

**Input**: Sección «#### H23.1 · `jurisprudencia` coteja el documento antes de decir nada de él, y un umbral sin modelo lo hace cumplir (ADR 0038, punto 1)» de `docs/ROADMAP.md`, en modo desatendido (ADR 0018).

## Resumen

La promesa de `jurisprudencia` desde H23 es que una sentencia solo se cita si su documento está en la conversación y `cita cotejar` ha leído su ficha. En la eval 06 la persona pide el ROJ STS 1088/2023 y pega el documento del ROJ STS 3144/2023, cuyo número de resolución es 1088/2023. En las ocho tandas medibles desde el 2026-10-10, 6 de las 48 respuestas de `claude-sonnet-5-5` a esa eval no ejecutan `cita cotejar` y dicen que el documento es la sentencia pedida: 4 en el modo orden y 2 en el modo herramienta (`docs/USO.md`, 2026-10-15). Es el error que más daño hace a un despacho, citar una sentencia que no es la que cree, y hoy ningún control lo hace fallar:

- **La regla de 2 de 3 por eval** (ADR 0016) lo esconde: la eval solo cayó cuando dos fallos coincidieron en una serie, en la #145.
- **`cita_sin_documento`** cuenta la cita con su forma fija cuyo ECLI no se cotejó, y estas respuestas no citan con esa forma.

Si se ejecutó `cita cotejar` es un hecho de la sesión que un guion comprueba. Por el ADR 0038, un umbral así va con cero. El hito hace dos cosas:

1. **Corrige la skill** (`jurisprudencia` v0.1.1). Con un documento en la conversación, lo primero es pasar su ficha a `cita cotejar`, antes de decir nada de él. Si es o no la sentencia pedida lo dice solo lo que devuelve la operación, nunca una comparación hecha a ojo.
2. **Lo hace cumplir con un umbral nuevo del informe del job**, `documento_sin_cotejar:<modelo>:<modo>`. Cuenta, en cada modo, las sesiones que no ejecutaron ningún `cita cotejar` en una eval que lo pide. Decide con 0.

No entrega skill nueva: corrige una y protege su promesa con un control (principio VIII). Va antes de H9 porque es un defecto de la v0.6.0 publicada en lo que la skill promete.

**Cómo se citan los requisitos de otros hitos.** Los de H5 a H26 se citan sin guion: «H23 FR 060», «H26 FR 012». Los ids con guion (FR-001, SC-001…) son siempre de este spec. `<modelo>` es el id del modelo que decide (ADR 0031), hoy `claude-sonnet-5-5`; `<modo>` es `orden` u `herramienta` (H21). **Eval que pide el cotejo** es una eval cuyos `comandos` llevan un comando con `applet: cita` y `verbo: cotejar`. **Cotejo** es una ejecución de `cita cotejar` en la sesión, como orden o como herramienta, en el sentido de FR-002.

## Clarifications

### Session 2026-10-16

- Q: ¿Entra en el hito cambiar `make evals-sondeo` para que dé lo que pide la Aceptación fuera del run (SC-010: «15 repeticiones por modo» y «0 de 30 sin cotejar» en la eval 06), sabiendo que hoy el sondeo mide solo el modo orden, juzga las sesiones sin sus invocaciones ni los comandos esperados y su salida no dice qué sesiones ejecutaron `cita cotejar`? → A: No. El sondeo no cambia y FR-033 se queda tal cual; el run no toca `internal/evals/sondeo.go` ni `scripts/evals-sondeo.sh`. SC-010 sigue siendo una medida humana sin procedimiento fijado por el spec; la vía sin tocar el sondeo es la de la línea base (motivos de `documento_sin_cotejar` en los informes del job). Va a «Fuera de alcance» y como pendiente para la persona. (auto: conservadora; criterio d; fuente: docs/ROADMAP.md, H23.1; constitución, «Gates», «Umbrales que se cumplen» y «Criterio de decisión autónoma», puntos 2 y 4; ADR 0032; CONTRIBUTING.md, «Sondeo local»; docs/USO.md, 2026-10-15)

## Criterios del hito, literales

Transcripción literal de `docs/ROADMAP.md` §4, H23.1. El Objetivo y el Alcance se reflejan en los requisitos.

- **Entrega**: «(1) `jurisprudencia` v0.1.1, que con un documento en la conversación pasa su ficha a `cita cotejar` antes de decir nada de él, y dice si es la sentencia pedida solo con lo que devuelve; (2) un umbral nuevo del informe del job, `documento_sin_cotejar:<modelo>:<modo>`, sin modelo, con 0, que decide; (3) `CHANGELOG.md` (*Unreleased*) y lo que `CONTRIBUTING.md`, `docs/JURISPRUDENCIA.md`, `docs/WORKFLOW.md` y `CLAUDE.md` dicen de los umbrales de `jurisprudencia`.»
- **Controles**:
  - «`make ci` en verde, con `schema-check` y las reglas del conjunto;»
  - «la sección «Controles de umbral» del plan con una fila por cada `documento_sin_cotejar:<modelo>:<modo>`, y las de los que ya decidían;»
  - «tests del umbral con sesiones sintéticas: una sesión de una eval que pide `cita cotejar` sin ningún cotejo cuenta y da `fallo`; con `cita cotejar` como orden o como herramienta no cuenta; una eval que no lo pide no entra en el total; una sesión sin medir no cuenta y da su motivo de la ejecución; una sesión de Haiku no entra;»
  - «un mutante por control: con la comparación en «ninguna sesión con cotejo» en lugar de «cada sesión», o con el `decide` en falso, algún test falla;»
  - «`CHANGELOG.md` (*Unreleased*); y el job de evals en la propuesta de cambio.»
- **Aceptación**: «en el informe del job de cierre de `jurisprudencia`, los dos `documento_sin_cotejar:<modelo>:<modo>` llevan `decide: true`, 0 de 12 y `cumple: true`; `cita_sin_documento`, `sin_activar`, `afirma_lo_no_leido:<modelo>`, la medida del juez y las duraciones se cumplen; ninguna respuesta queda «sin juzgar» y el veredicto es aprobado. `legal-core` sigue aprobado y `red` está vacío; `boe-legislacion`, si sale en rojo, no es motivo de este hito y queda en el informe. El informe final da los dos umbrales nuevos como «comprobado por su control». Después, y fuera del run porque es humano: con 6 sesiones de la eval 06 por cierre no se distingue un arreglo de la suerte, así que Jorge lanza antes de fusionar un sondeo de la eval 06 con la skill del hito (`make evals-sondeo`, 15 repeticiones por modo, `claude-sonnet-5-5`): 0 de 30 sin cotejar, frente a 6 de 48 de partida; si no, no se fusiona y decide qué sigue.»

Trazabilidad resumida (el detalle, en cada requisito):

| Criterio del hito | Dónde se cumple |
|---|---|
| Entrega 1, la skill | FR-020 a FR-024, US2 |
| Entrega 2, el umbral | FR-001 a FR-007, US1 |
| Entrega 3, los documentos | FR-040 a FR-045, US4 |
| Lo que sigue igual | FR-030 a FR-034, US3 |
| El cierre y el run | FR-050 a FR-052 |
| Controles | FR-060 a FR-066, SC-002 a SC-008 |
| Aceptación | SC-001, SC-009, SC-010 |

## Relación con H23, H25 y H26

Ningún spec, plan ni informe de un hito anterior se edita: son el registro de sus runs. No hay ADR nuevo: el umbral con cero de un hecho de la sesión es el del ADR 0038 («cero donde hay forma»).

**Sustituye** (lo que dice el hito citado deja de valer, y vale lo de este spec):

- **Los trece elementos de `umbrales` de `jurisprudencia`, nueve con `decide: true`** (H26 FR 012): pasan a ser quince, once con `decide: true` (FR-006).
- **El paso 3 de `jurisprudencia`** («Con el documento delante: cotejar su ficha», H23 y v0.1 de H25): pasa a ser lo primero que se hace con un documento (FR-020).

**Se queda**, sin que este spec lo vuelva a definir:

- **Qué sesiones se miden y qué es una sesión sin medir** (H7.3 FR 040 a FR 044; H21 FR 043 y FR 044): las del modelo que decide, por modo, en las evals que activan la skill. Una sesión sin medir da su motivo, con el prefijo «de la ejecución, no de la skill».
- **Qué ejecutó una sesión** (H5 y H21): lo que `internal/evals` ya extrae de su traza, en el modo orden, y de sus llamadas a las herramientas del servidor, en el modo herramienta. Con ello el juicio de cada sesión reparte hoy sus comandos esperados entre `comandos_ejecutados` y `comandos_ausentes`.
- **El contrato de `umbrales`** (ADR 0029): `nombre`, `descripcion`, `medida`, `total`, `comparacion`, `umbral`, `cumple` y `decide`. `cumple` es la comparación en coma flotante, sin redondeos, de `medida` entre `total`, que vale 0 si `total` es 0.
- **`cita_sin_documento`, `sin_activar`, las dos clases del juez, el umbral conjunto, la medida del juez y la duración de sus votos** (H23 FR 060 a FR 063; H25; H26), con sus valores y su `decide`.
- **El modelo que decide, las repeticiones y la regla por serie** (ADR 0016 y 0031), y la definición del job: sus topes, su concurrencia y sus versiones.

## User Scenarios & Testing *(mandatory)*

### User Story 1 - El job sale en rojo si una sesión da por bueno un documento sin cotejarlo (Priority: P1)

Quien lanza un hito, o una propuesta que toca `jurisprudencia`, lee el veredicto de su job de evals. Si una sesión de una eval con un documento delante no ejecutó ningún `cita cotejar`, el job sale en `fallo`, con un motivo que nombra el umbral del modo y cada sesión que no cotejó. Es así aunque la serie de esa eval pase con 2 de 3 y aunque la respuesta no cite con la forma fija.

**Why this priority**: es lo que el hito añade al producto. Sin este control, el defecto de la v0.6.0 solo se ve cuando coinciden dos fallos en una serie.

**Independent Test**: los tests del umbral con sesiones sintéticas, en `make ci` (SC-002 a SC-006); el job de cierre (SC-001).

**Acceptance Scenarios**:

1. **Dado** un informe sintético de `jurisprudencia` con 12 sesiones medidas de `claude-sonnet-5-5` en el modo orden en evals que piden el cotejo, una de ellas sin ningún cotejo, **Cuando** se escribe, **Entonces** `documento_sin_cotejar:claude-sonnet-5-5:orden` lleva `medida` 1, `total` 12, `umbral` 0, `cumple: false` y `decide: true`; el veredicto es `fallo`; y el motivo nombra el umbral y esa sesión (FR-001, FR-003, FR-004).
2. **Dado** el mismo informe con esa sesión ejecutando `cita cotejar` como orden, o, en el modo herramienta, con una llamada a `cita_cotejar`, **Cuando** se escribe, **Entonces** la sesión no cuenta: `medida` 0 de 12 y `cumple: true`, y el umbral no da motivo (FR-002).
3. **Dado** un informe con una sesión sin cotejo de una eval que no pide el cotejo (la 01, que pide `cita preparar`), **Cuando** se escribe, **Entonces** esa sesión no está en el `total` de ningún `documento_sin_cotejar` ni en su `medida` (FR-001).
4. **Dado** un informe con una sesión sin medir de la eval 06, **Cuando** se escribe, **Entonces** no está en el `total` ni en la `medida` del umbral de su modo, y el veredicto es `fallo` por su motivo de la ejecución, que la nombra (FR-005).
5. **Dado** un informe con una sesión de `claude-haiku-4-5-20251001`, el modelo informativo, sin cotejo en la eval 06, **Cuando** se escribe, **Entonces** no entra en ningún `documento_sin_cotejar` y no cambia el veredicto (FR-001).
6. **Dado** el job de cierre de `jurisprudencia`, **Cuando** se lee `umbrales`, **Entonces** los dos `documento_sin_cotejar:<modelo>:<modo>` llevan `decide: true`, 0 de 12 y `cumple: true` (SC-001).

---

### User Story 2 - La skill coteja el documento antes de decir nada de él (Priority: P1)

Una persona de un despacho dice que trae una sentencia y pega su documento. Antes de decir nada del documento, la skill pasa su ficha a `cita cotejar`, con la referencia que se pidió. No dice antes que es la pedida, ni que coincide, ni de qué trata. Si es la pedida o no lo dice `es_la_pedida`. Cuando el número de resolución del documento es el ROJ pedido, la respuesta dice que no es la pedida, con las palabras del hallazgo, y prepara la consulta de la pedida.

**Why this priority**: es la corrección del defecto. El umbral de US1 la hace cumplir.

**Independent Test**: el job de cierre, con los dos `documento_sin_cotejar` a 0 (SC-001); y, fuera del run, el sondeo de la persona (SC-010). `make ci` comprueba la forma de la skill (SC-007).

**Acceptance Scenarios**:

1. **Dado** `skills/jurisprudencia/SKILL.md` del hito, **Cuando** se lee su protocolo, **Entonces** con un documento en la conversación lo primero que pide es pasar su ficha a `cita cotejar`, y ninguna instrucción le permite decir del documento que es la pedida, que coincide o de qué trata antes de ese cotejo (FR-020, FR-021).
2. **Dado** el mismo `SKILL.md`, **Cuando** se lee lo que dice de si el documento es el pedido, **Entonces** lo decide `es_la_pedida` de la salida de `cita cotejar`, y prohíbe compararlo a ojo, también cuando el número de resolución del documento coincide con el ROJ pedido (FR-022).
3. **Dado** el mismo `SKILL.md`, **Cuando** pasa `make ci`, **Entonces** tiene menos de 300 líneas, su tabla de comandos es la que regenera `make skills-sync` y no nombra ninguna herramienta nueva (FR-023).
4. **Dado** el job de cierre, **Cuando** se lee su informe, **Entonces** ninguna sesión de `claude-sonnet-5-5` en las evals 04, 06, 08 y 09 queda sin cotejo, en ninguno de los dos modos (SC-001).

---

### User Story 3 - Lo que ya decidía sigue decidiendo igual (Priority: P2)

Quien lee el informe de `jurisprudencia` sigue viendo `sin_activar`, `cita_sin_documento`, las dos clases del juez, el umbral conjunto, la medida y la duración de los votos con los mismos valores, la misma medida y el mismo `decide`. `boe-legislacion` y `legal-core` no ganan ningún umbral.

**Why this priority**: el hito no puede aflojar ni mover lo que ya protege a las tres skills.

**Independent Test**: el test de los elementos de `umbrales` de las tres skills, en `make ci` (SC-005).

**Acceptance Scenarios**:

1. **Dado** un informe sintético de `jurisprudencia` con el plan del job, **Cuando** se escribe, **Entonces** `umbrales` lleva exactamente 15 elementos, 11 con `decide: true`, y los 13 de antes siguen con su nombre, su `umbral`, su `decide` y su medida (FR-006, FR-030).
2. **Dado** un informe sintético de `boe-legislacion` y otro de `legal-core`, **Cuando** se escriben, **Entonces** `umbrales` sigue llevando 13 elementos con 9 que deciden y `[]`, respectivamente, y ninguno es `documento_sin_cotejar` (FR-007).
3. **Dado** una sesión de la eval 06 sin cotejo que además no activó la skill, **Cuando** se escribe el informe, **Entonces** cuenta en `sin_activar` y en `documento_sin_cotejar` de su modo, cada uno con su motivo (FR-001, FR-030).

---

### User Story 4 - Los documentos dicen los umbrales que el job aplica (Priority: P3)

Quien contribuye, o lee un informe del job, encuentra en `CONTRIBUTING.md`, `docs/JURISPRUDENCIA.md`, `CLAUDE.md` y *Unreleased* de `CHANGELOG.md` el umbral nuevo, con lo que mide, su valor, que decide y cuántos umbrales lleva `jurisprudencia`.

**Why this priority**: sin ello, el informe dice una cosa y la documentación otra.

**Independent Test**: la revisión final, por lectura contra el árbol (SC-008).

**Acceptance Scenarios**:

1. **Dado** `CONTRIBUTING.md` del hito, **Cuando** se lee la tabla de `umbrales` de `jurisprudencia`, **Entonces** tiene una fila para `documento_sin_cotejar:<modelo>:<modo>` y dice que son quince y deciden once (FR-041).
2. **Dado** *Unreleased* de `CHANGELOG.md`, **Cuando** se lee, **Entonces** registra `jurisprudencia` v0.1.1 y el umbral nuevo (FR-040).

---

### Edge Cases

- **Un cotejo sin la referencia pedida** (en la eval 06, `cita cotejar` sin `--roj`): la sesión no cuenta en `documento_sin_cotejar`, que pide «algún cotejo». El comando esperado `cita cotejar` con `roj: STS 1088/2023` queda en `comandos_ausentes` de la sesión, con su motivo, y lo decide la regla de la serie, como hoy (FR-002, FR-031).
- **Un cotejo que termina con el código 2**, porque la ficha pasada no era entera, sin otro cotejo que termine bien: la operación no leyó ninguna ficha, y la sesión cuenta (FR-002).
- **Un cotejo con `--dry-run` o `--describe`, o que pide la ayuda**: no coteja nada, y la sesión cuenta si no hay otro (FR-002).
- **Varios cotejos en una sesión**: la sesión cuenta 0 o 1 vez, nunca más (FR-001).
- **Una sesión con cotejo y otra sin él, en el mismo modo**: la primera no cuenta y la segunda sí; `medida` 1 y `fallo`. Es el caso que distingue «cada sesión» de «ninguna sesión con cotejo» (FR-001, FR-064).
- **Una sesión que coteja y luego contradice lo que devuelve `cita cotejar`**: no cuenta. Es significado, y queda fuera de alcance (FR-002).
- **Una sesión que el tope de la sesión cortó**: entra o no en el `total` igual que entra hoy en el de `sin_activar` de su modo. Si entra y no llegó a cotejar, cuenta (FR-001).
- **Una sesión sin la skill activada**: entra en el `total` y cuenta si no cotejó, además de contar en `sin_activar` (FR-001).
- **Ninguna sesión medida en un modo** (todas sin medir): `total` 0, la proporción es 0 y el umbral se cumple, como todo umbral con `total` 0 (contrato del ADR 0029). El veredicto es `fallo` por los motivos de las sesiones sin medir (FR-005).
- **Un plan con un solo modo**: un solo `documento_sin_cotejar`, el de ese modo (FR-001).
- **Una skill sin ninguna eval que pida el cotejo** (`boe-legislacion`, `legal-core`): no lleva `documento_sin_cotejar` (FR-007).
- **El cierre sale en rojo por `documento_sin_cotejar`**: lo trata FR-050.

## Requirements *(mandatory)*

### Functional Requirements

#### El umbral

Cada uno es un elemento de `umbrales` de `informe.json` de una skill, con el contrato del ADR 0029.

- **FR-001**: Por cada modo del plan, `umbrales` MUST llevar exactamente un elemento de nombre `documento_sin_cotejar:<modelo>:<modo>`, siempre que la skill tenga alguna eval que active la skill y pida el cotejo, sin ningún caso por skill. Lleva:
  - `total`: las sesiones de ese modo que cuentan hoy en el `total` de `sin_activar:<modelo>:<modo>` —las medidas del modelo que decide en las evals que activan la skill— y son de una eval que pide el cotejo;
  - `medida`: las de esas sesiones en las que no hubo ningún cotejo (FR-002), cada sesión como mucho una vez;
  - `comparacion` `"<="`, `umbral` 0 y `decide: true`;
  - una `descripcion` que dice qué sesiones cuenta y sobre cuáles.

  Qué evals entran lo dicen sus `comandos`, no una lista en el código. Hoy entran la 04, la 06, la 08 y la 09 de `jurisprudencia`: 12 sesiones por modo con 3 repeticiones. No entran las sesiones de un modelo informativo (hoy `claude-haiku-4-5-20251001`) ni las de una eval sin binario ni servidor. **Control**: con una sesión que cuenta, el veredicto es `fallo` y el job de evals de esa skill sale en rojo (FR-003). Comprobable: FR-062.
- **FR-002**: Una sesión **coteja** si ejecutó `cita cotejar`, con cualquier argumento, de una de dos formas: como orden, en su traza, o como llamada a la herramienta `cita_cotejar` del servidor. La ejecución tiene que consultar y terminar bien, con el mismo criterio con el que el juicio de la sesión da hoy por ejecutado un comando esperado de `cita` (H5 data-model §6.1; H23 contracts/evals-jurisprudencia.md §2), sin mirar sus argumentos: con la orden, código 0 y sin `--dry-run`, `--describe` ni la ayuda; con la herramienta, un resultado que no viene marcado como error. Lo que la respuesta diga después no cambia si la sesión cotejó. Se mide sin modelo y sin el juez. Comprobable: FR-062.
- **FR-003**: Un `documento_sin_cotejar` que no se cumple MUST poner el veredicto en `fallo`, y el job MUST salir en rojo. Su motivo MUST nombrar el umbral con su medida y su condición, como el de todo umbral que decide (por ejemplo, «umbral `documento_sin_cotejar:claude-sonnet-5-5:orden`: 1 de 12 (8,3 %), y tiene que ser ≤ 0,0 %»). Detrás MUST nombrar cada sesión que cuenta, por su nombre y en orden de sesión. El texto exacto lo fija el plan. Uno que se cumple MUST NOT dar motivo ni cambiar el veredicto. Comprobable: FR-062.
- **FR-004**: `cumple` MUST ser la comparación del contrato: `medida` entre `total` menor o igual que 0, que vale 0 si `total` es 0. Con 1 sesión sin cotejo de 12, `cumple: false`; con 0 de 12, `cumple: true`. Comprobable: FR-062.
- **FR-005**: Una sesión sin medir MUST NOT estar en el `total` ni en la `medida` de ningún `documento_sin_cotejar`, y MUST seguir dando su motivo «de la ejecución, no de la skill», que la nombra y pone el veredicto en `fallo`, como hoy. Comprobable: FR-062.
- **FR-006**: Con el plan del job, `umbrales` de `jurisprudencia` MUST llevar exactamente quince elementos, once con `decide: true`. Son los trece de H26 FR 012 y los dos `documento_sin_cotejar:<modelo>:<modo>`. Su orden lo fija el plan, y `informe.md` da cada uno en su tabla «Umbrales» como una fila más. **Control**: un test de `make ci` falla con otro número de elementos, otro orden o otro `decide` en alguno (FR-063).
- **FR-007**: `umbrales` de `boe-legislacion` MUST seguir con sus trece elementos, nueve con `decide: true`, y el de `legal-core`, `[]`: ninguna de sus evals pide el cotejo. Comprobable: FR-063.

#### La skill

- **FR-020**: En `skills/jurisprudencia/SKILL.md`, el cotejo de la ficha MUST ser lo primero que la skill hace con un documento que está en la conversación, antes de cualquier frase sobre él. La redacción la escribe el run. La skill MUST seguir diciendo lo que ya dice del cotejo: la ficha y nada más, la referencia en la forma en que se pidió y la orden con la entrada estándar. Comprobable: lo leen los jueces de la revisión final (SC-007); su efecto lo mide el job (SC-001).
- **FR-021**: Ninguna instrucción de `SKILL.md` MUST permitir decir de un documento, antes de cotejarlo, que es la sentencia pedida, que coincide con ella o de qué trata. Comprobable: como FR-020.
- **FR-022**: `SKILL.md` MUST decir que si el documento es o no la sentencia pedida lo dice `es_la_pedida`, y MUST prohibir la comparación hecha a ojo. Nombra el caso en que falla: el número de resolución del documento coincide con el ROJ pedido. Comprobable: como FR-020.
- **FR-023**: `SKILL.md` MUST seguir por debajo de 300 líneas, sin herramientas nuevas, con la tabla de comandos que regenera `make skills-sync` sin cambios y con la misma `description`. **Control**: `make skills-check`, dentro de `make ci`, falla con 300 líneas o más, o con una tabla distinta de la regenerada.
- **FR-024**: Lo que el hito no pide de la skill MUST NOT cambiar: los pasos 2 y 4 a 7 y las reglas, salvo la frase de alguno que hiciera hablar del documento antes del cotejo. La versión de la skill pasa a v0.1.1, que registra `CHANGELOG.md` (FR-040). Comprobable: el diff del hito, que leen el guardián de diff y la revisión final.

#### Lo que sigue igual

- **FR-030**: Los trece umbrales que `jurisprudencia` ya lleva MUST seguir como están: nombre, `descripcion`, `medida`, `total`, `umbral` y `decide`. Son `sin_activar` y `cita_sin_documento` de cada modo, con 0, que deciden; los recuentos por modo de `afirma_lo_no_leido` y de `afirma_que_existe`, que no deciden; `afirma_lo_no_leido:<modelo>`, con 0.017, que decide; los dos de la medida del juez; y `duracion_del_juez` de cada modo. **Control**: los tests de `make ci` que ya los comprueban siguen pasando sin cambiar lo que esperan de ellos, y el de FR-063 falla si alguno cambia.
- **FR-031**: El juicio de cada sesión MUST NOT cambiar: sus `comandos_ejecutados`, `comandos_ausentes` y motivos, y la tasa y el veredicto de cada serie (ADR 0016). El umbral nuevo no cambia lo que mide ningún otro. Comprobable: los tests del juicio y de las series que ya hay siguen pasando sin cambiar lo que esperan.
- **FR-032**: Las evals de `jurisprudencia` y sus repeticiones, la rúbrica, los casos, `clases.yaml`, la medida versionada, el modelo del juez y su versión de Claude Code MUST NOT cambiar. `evidencias/` y `schemas/` tampoco. El hito no añade ninguna sesión ni ningún voto, y la definición del job no cambia. Comprobable: `TestDefinicionDelJob`, `TestCopiasDelJuez` y `schema-check` siguen pasando en `make ci` sin cambiar lo que esperan.
- **FR-033**: `make evals-sondeo` MUST seguir como está, con las mismas líneas de salida y el mismo código de salida. Comprobable: los tests del sondeo que ya hay siguen pasando sin cambiar lo que esperan. El hito no toca `internal/evals/sondeo.go` ni `scripts/evals-sondeo.sh`, y no entrega nada para que el sondeo cuente las sesiones sin cotejo ni para que mida el modo herramienta. Comprobable: el diff del hito no toca esos dos ficheros.
- **FR-034**: `boe-legislacion` y `legal-core` MUST seguir como están: sus `SKILL.md`, sus `references/`, sus evals y sus umbrales (FR-007). Comprobable: el diff del hito no toca `skills/boe-legislacion/`, `skills/legal-core/` ni sus evals.

#### Documentos

- **FR-040**: `CHANGELOG.md` (*Unreleased*) MUST registrar `jurisprudencia` v0.1.1, que sustituye a la v0.1: con un documento delante, coteja su ficha antes de decir nada de él, y si es la pedida lo dice `es_la_pedida`. Lo hace con las cifras del defecto que corrige (6 de 48 en la eval 06). MUST registrar también el umbral `documento_sin_cotejar:<modelo>:<modo>`: qué cuenta, sobre qué sesiones, con 0 y que decide; y que `umbrales` de `jurisprudencia` pasa de trece elementos con nueve que deciden a quince con once. Las secciones de las versiones publicadas MUST NOT cambiar. Comprobable: SC-008.
- **FR-041**: `CONTRIBUTING.md`, en lo que dice de los umbrales de `jurisprudencia` en «Job de evals», MUST llevar:
  - una fila de `documento_sin_cotejar:<modelo>:<modo>` en su tabla, con su `medida`, su `total` (12), su `umbral` (`0`) y `decide` sí;
  - «quince» y «once» donde hoy dice «trece» y «nueve»;
  - el orden de los quince y cuántos no deciden;
  - un párrafo que dice que se mide sin modelo, con lo que el juicio ya extrae de cada sesión, y que lo tiene toda skill con una eval que pide el cotejo;
  - la forma de su motivo.

  Comprobable: SC-008.
- **FR-042**: `docs/JURISPRUDENCIA.md`, donde dice con qué umbral sin modelo decide el job sobre la skill («el job decide con `cita_sin_documento`, un hecho de la sesión»), MUST decir también `documento_sin_cotejar`. Comprobable: SC-008.
- **FR-043**: `CLAUDE.md`, donde enumera lo que comprueba un guion con umbral 0 («la cita, las líneas `⚠`, la activación, `cita_sin_documento`, la medida del juez»), MUST nombrar también el cotejo del documento (`documento_sin_cotejar`). Comprobable: SC-008.
- **FR-044**: `docs/WORKFLOW.md` MUST corregirse donde diga algo de los umbrales de `jurisprudencia` que el hito deje falso, y solo ahí. Hoy no lo dice en ningún sitio: en «`medir_cierre`» enumera los nueve que deciden en `boe-legislacion`, que siguen siendo ciertos. Si no hay nada falso, no cambia. Comprobable: SC-008.
- **FR-045**: La corrección de FR-041 a FR-044 vale para toda afirmación sobre los umbrales de `jurisprudencia` que el hito deje falsa o incompleta en esos ficheros, y para nada más. MUST NOT reescribir lo que sigue siendo cierto, ni prometer ninguna release, ni añadir un test que ate ese texto al código. Cada cifra sale del árbol y no de memoria. MUST NOT tocarse `docs/ROADMAP.md`, ningún ADR, ni el spec, el plan o los informes de un hito anterior. Comprobable: SC-008.

#### El cierre y el run

- **FR-050**: Ninguna corrección ni reparación del run MUST cumplir un umbral cambiando su valor, dejándolo en `decide: false`, sacando sesiones de su `total` o cambiando las evals, la rúbrica, los casos o la medida (ADR 0029). Ninguna MUST tocar `boe-legislacion`, `legal-core` ni sus umbrales. Si el cierre sale en rojo por un umbral de `boe-legislacion`, queda en el informe final para la persona y no es motivo de este hito (ADR 0038, punto 8). Un rojo por un defecto de lo que el hito entrega —el código del umbral, el paso del cotejo de `SKILL.md` o un documento— sí se repara. Comprobable: lo leen los jueces de la revisión final contra el diff de cada reparación.
- **FR-051**: Dentro del run, ninguna sesión MUST abrir una sesión de evals ni el juez (ADR 0032): los tests usan sesiones sintéticas. El run MUST NOT lanzar un sondeo ni la medida del juez, y solo pone la etiqueta `evals`.
- **FR-052**: El hito MUST NOT cambiar el binario fuera de `internal/evals`, ni `scripts/workflow/`, ni la constitución, ni nada bajo `.github/`. Lo que cambie bajo `testdata/` va en tareas `[datos]`. Comprobable: el diff del hito, que leen el guardián de diff y la revisión final.

#### Controles y Definition of Done

- **FR-060**: `make ci` MUST quedar en verde, con `schema-check` y las reglas del conjunto.
- **FR-061**: La sección «Controles de umbral» del plan MUST tener una fila, en `evals:jurisprudencia:<nombre>`, por cada `documento_sin_cotejar:<modelo>:<modo>` (FR-001). Y las de los que ya decidían en `jurisprudencia`: `sin_activar:<modelo>:<modo>` y `cita_sin_documento:<modelo>:<modo>` de cada modo, `afirma_lo_no_leido:<modelo>`, los dos de la medida del juez y `duracion_del_juez:<modo>` de cada modo. Son once filas. MUST NOT tener fila un umbral que no decide. La fila del test de los elementos (FR-006) la pone el plan, con su control `ci:`.
- **FR-062**: Los tests del umbral con sesiones sintéticas MUST cubrir cada uno de estos casos:
  - una sesión de una eval que pide el cotejo, sin ningún cotejo, cuenta y da `fallo`, con el motivo de FR-003;
  - con `cita cotejar` como orden no cuenta, y con una llamada a `cita_cotejar` en el modo herramienta tampoco;
  - una eval que no pide el cotejo no entra en el `total`;
  - una sesión sin medir no cuenta y da su motivo de la ejecución;
  - una sesión de Haiku no entra.
- **FR-063**: Un test MUST comprobar los quince elementos de `umbrales` de `jurisprudencia`, en su orden y con los once que deciden (FR-006), y que `boe-legislacion` sigue con trece, nueve que deciden, y `legal-core` con `[]` (FR-007).
- **FR-064**: Un mutante por control. El primero hace la comparación en «ninguna sesión con cotejo» en lugar de «cada sesión»: cuenta las sesiones sin cotejo solo si ninguna sesión del modo cotejó. El segundo pone el `decide` de `documento_sin_cotejar` en falso. Con cada uno, por separado, algún test de FR-062 MUST fallar. Para el primero hace falta un test con, en el mismo modo, una sesión con cotejo y otra sin él. La tarea que lo comprueba dice qué test falla con cada mutante, y ningún mutante queda en el árbol. Es la prueba de que el control de FR-001 falla por encima de su umbral (ADR 0029).
- **FR-065**: `CHANGELOG.md` (*Unreleased*) MUST llevar lo de FR-040.
- **FR-066**: El job de evals MUST ejecutarse en la propuesta de cambio (lo hace el workflow tras la revisión final), y su informe da lo que pide SC-001.

### Key Entities

- **Eval que pide el cotejo**: una eval con `activa: true` cuyos `comandos` llevan un comando con `applet: cita` y `verbo: cotejar`. Hoy, la 04, la 06, la 08 y la 09 de `jurisprudencia`.
- **Cotejo**: una ejecución de `cita cotejar` en la sesión, como orden o como herramienta, que consultó y terminó bien (FR-002).
- **Sesión sin cotejo**: una sesión medida del modelo que decide, en una eval que pide el cotejo, sin ningún cotejo.
- **`documento_sin_cotejar:<modelo>:<modo>`**: el umbral de las sesiones sin cotejo de un modo, sobre las sesiones medidas de ese modo en las evals que piden el cotejo. Decide con 0.

## Success Criteria *(mandatory)*

### Measurable Outcomes

- **SC-001**: En el informe del job de cierre de `jurisprudencia`:
  - los dos `documento_sin_cotejar:<modelo>:<modo>` llevan `decide: true`, 0 de 12 y `cumple: true`;
  - `cita_sin_documento`, `sin_activar`, `afirma_lo_no_leido:<modelo>`, la medida del juez y las duraciones se cumplen;
  - ninguna respuesta queda «sin juzgar» y el veredicto es aprobado.

  `legal-core` sigue aprobado y `red` está vacío. `boe-legislacion`, si sale en rojo, no es motivo de este hito y queda en el informe (FR-050).

  **Control**: el job de evals de `jurisprudencia` sale en rojo, con el veredicto `fallo`, en cualquiera de estos casos: una sesión sin cotejo en un modo (FR-001, FR-003); `sin_activar` o `cita_sin_documento` no se cumplen; `afirma_lo_no_leido:<modelo>` pasa de su tasa; la medida versionada no corresponde o no se cumple; la duración de los votos pasa de su cota; una respuesta queda sin juzgar o una sesión sin medir (FR-005); o una serie que decide no pasa. El job de `legal-core` sale en rojo si su veredicto no es aprobado, y el de cualquier skill, si llega una petición a la red. El cierre del workflow cuenta cada rojo.
- **SC-002**: Con 12 sesiones medidas en un modo y 1 sin cotejo, `documento_sin_cotejar` de ese modo da `medida` 1, `cumple: false` y veredicto `fallo`, con 1 motivo que nombra el umbral y la sesión; y con 0 sin cotejo, `cumple: true` y 0 motivos por él. **Control**: los tests de FR-062 fallan en `make ci`.
- **SC-003**: Un cotejo como orden, en el modo orden, y una llamada a `cita_cotejar`, en el modo herramienta, hacen que la sesión no cuente: 2 de 2 casos. **Control**: los tests de FR-062 fallan en `make ci`.
- **SC-004**: Las sesiones de una eval que no pide el cotejo, las sesiones sin medir y las de Haiku suman 0 en el `total` y en la `medida` de `documento_sin_cotejar`. La sesión sin medir da 1 motivo de la ejecución. **Control**: los tests de FR-062 fallan en `make ci`.
- **SC-005**: `umbrales` lleva 15 elementos con 11 que deciden en `jurisprudencia`, 13 con 9 en `boe-legislacion` y 0 en `legal-core`. **Control**: el test de FR-063 falla en `make ci`.
- **SC-006**: Con cada uno de los 2 mutantes de FR-064, falla al menos 1 test de FR-062: 2 de 2. **Control**: los tests de FR-062, en `make ci`; lo comprueban la tarea de FR-064 y la revisión final.
- **SC-007**: `skills/jurisprudencia/SKILL.md` tiene menos de 300 líneas y 0 herramientas nuevas, y su tabla de comandos es la regenerada. **Control**: `make skills-check`, en `make ci` (FR-023). Que el cotejo sea lo primero y que no se compare a ojo (FR-020 a FR-022) lo leen los jueces de la revisión final, y lo mide el job con SC-001.
- **SC-008**: `CONTRIBUTING.md`, `docs/JURISPRUDENCIA.md`, `docs/WORKFLOW.md`, `CLAUDE.md` y *Unreleased* de `CHANGELOG.md` llevan 0 afirmaciones de que `jurisprudencia` tiene trece umbrales con nueve que deciden, o de que solo `cita_sin_documento` decide sin modelo sobre ella. Y cada cifra, fichero, test y umbral que nombran al decir lo nuevo existe en el árbol. **Control**: la revisión final, por lectura contra el árbol. Ningún test de `make ci` ata ese texto al código, y no se añade uno (FR-045).
- **SC-009**: El informe final da los dos `documento_sin_cotejar:<modelo>:<modo>` como «comprobado por su control». **Control**: lo decide sin modelo el informe final del workflow. Solo lo dice de un umbral con fila en «Controles de umbral» del plan (FR-061) que el job publica, cumple y hace fallar el job. Si el job no lo publica, no lo cumple o lleva `decide: false`, el informe final lo da como sin verificar o como no cumplido.
- **SC-010**: Después del run, y fuera de él porque es humano, antes de fusionar: Jorge lanza un sondeo de la eval 06 con la skill del hito (`make evals-sondeo`, 15 repeticiones por modo, `claude-sonnet-5-5`), y espera 0 de 30 sin cotejar, frente a 6 de 48 de partida. Si no, no se fusiona y decide qué sigue. Lo mide una persona: no tiene control en el run (FR-033), y el spec no fija su procedimiento. Hoy `make evals-sondeo` mide solo el modo orden, juzga sin lo que el job lee de la traza y no dice qué sesiones ejecutaron `cita cotejar`, así que el «0 de 30» no se lee en su salida. Sin tocar el sondeo, la vía que ya hay es la de la línea base (6 de 48, sacada de los informes del job de ocho tandas, `docs/USO.md`, 2026-10-15): el motivo de `documento_sin_cotejar` nombra cada sesión sin cotejo de cada modo (FR-003), y cada job da 3 sesiones de la eval 06 por modo. Cómo mide la persona lo decide ella.

## Uso, de fuera adentro

Cada salida del hito, desde quien la consume (criterio de uso, ADR 0028). Lo que gana la skill es que no da por buena una sentencia que no es la pedida. La salida que consume una skill es lo que devuelve `cita cotejar`, que no cambia; el hito cambia cuándo lo pide. Las demás son del job y de quien lo lee. Nada de lo que entrega crece con el uso del kit (cientos de normas y miles de bloques consultados en meses): cada job parte de las evals del repositorio y mide de nuevo sobre su commit.

| Salida | Quién la pide, cuántas veces y qué hace con ella | Tamaño | Cuándo deja de darse cada señal |
|---|---|---|---|
| La salida de `cita cotejar` (`ficha`, `correspondencia`, `pedida`, `es_la_pedida`, `hallazgos`), que no cambia (FR-020 a FR-022) | La skill, en su paso del cotejo: una vez por documento delante, antes de decir nada de él. Con un documento por pregunta, una vez; con varios, una por documento. Con `es_la_pedida` decide si cita el documento o prepara la consulta de la pedida (paso 2). | Un sobre por documento: el de la ficha de la eval 06 con `--roj "STS 1088/2023"`, que lleva su hallazgo, mide 949 bytes (medido en esta sesión con `go run ./cmd/kitlegal cita cotejar … --json --no-graph`). Con un documento por pregunta, ~1 KB; con *n* documentos, *n* × ~1 KB. Lo que se pasa es la ficha y no el texto (lo dice ya el paso 3). No depende del uso del kit: lo acota el número de documentos de la pregunta. | Una por documento y pregunta. No se acumula: un documento nuevo es un cotejo nuevo. |
| Los dos `documento_sin_cotejar:<modelo>:<modo>` de `umbrales`, en `informe.json` y en la tabla «Umbrales» de `informe.md` (FR-001, FR-006) | El job, una vez, que decide con ellos el veredicto. El informe final del workflow, una vez por medición del cierre, que los lee sin modelo (SC-009). Y la persona que lee el informe. | 2 elementos más en `jurisprudencia`. Los 13 de hoy miden entre 249 y 387 bytes, 4 108 en total, en el último cierre versionado (`specs/021-h26-el-umbral-de/gates/evals/jurisprudencia.json`, medido con `jq`). Con estos dos, unos 4,8 KB. Fijo: depende de los modos y de las evals, no del uso. | Cada job los mide de nuevo. El `cumple: false` deja de darse en el primer job en que todas las sesiones del modo cotejan. |
| El motivo de `fallo` de `documento_sin_cotejar` (FR-003) | La persona, y la reparación del cierre (FR-050). Uno por modo con alguna sesión sin cotejo: como mucho dos por job. | El umbral, su medida y su condición, unos 90 bytes, y el nombre de cada sesión que cuenta, de unos 55 bytes («06-documento-que-no-es-el-pedido-claude-sonnet-5-5-01»). Acotado por las 12 sesiones del modo: como mucho unos 0,8 KB. | Deja de darse en el primer job en que el umbral se cumple. No se acumula de un job a otro. |
| `SKILL.md` de `jurisprudencia` (FR-020 a FR-024) | El agente que activa la skill, una vez por conversación. | Menos de 300 líneas (197 en `main`; 207 con el hito). | No da señales. |
| Los documentos (FR-040 a FR-045) | Quien contribuye o lee un informe del job. | Una fila de tabla, unas frases y una entrada de *Unreleased*; no crecen con el uso. | No dan señales. |

## Fuera de alcance

Del hito, literal:

- «`boe-legislacion` y `legal-core`;»
- «que una respuesta que sí cotejó contradiga lo que devolvió `cita cotejar`: es significado y, si hiciera falta, una clase del juez con su rúbrica, sus casos y su medida (ADR 0037), en otro hito;»
- «evals nuevas, salvo que el spec encuentre una forma del defecto que la 06 no cubre; si añade alguna, recalcula el tope del trabajo de evals de `jurisprudencia` antes de cerrar el plan;». El spec no encuentra ninguna (Assumptions), y no añade evals.
- «cambiar el modelo que decide, las repeticiones o la regla de 2 de 3 por eval (ADR 0016 y 0031).»

De lo que el hito no especifica (constitución, «Criterio de decisión autónoma», punto 2):

- Cambiar `make evals-sondeo` (`internal/evals/sondeo.go`, `scripts/evals-sondeo.sh`) para que cuente las sesiones sin cotejo o mida el modo herramienta (FR-033). SC-010 sigue siendo una medida humana, fuera del run y sin control en él. **Pendiente para la persona**, en el informe final: el «0 de 30» de la Aceptación no se puede leer en la salida del sondeo; si quiere que lo dé el sondeo, es un cambio fuera de este hito o una entrada corregida.
- Exigir en `documento_sin_cotejar` que el cotejo lleve la referencia pedida, o que sea del documento de la pregunta. El hito pide «ningún `cita cotejar`», y lo de la referencia lo mide ya el comando esperado de la eval 06 con su serie.
- Un umbral parecido para `cita preparar`, o para otra skill.
- Cambiar la `description`, la tabla de comandos, los pasos 2 y 4 a 7 o las reglas de `jurisprudencia` (FR-024), o el applet `cita`.
- Cambiar el contrato de `umbrales` (ADR 0029), `scripts/workflow/` (también `informe.sh` y `serie.sh`) o lo que el informe final dice de un umbral.
- Cambiar el esquema del formato de eval, o una clave nueva en una eval.
- Corregir el comentario de `.github/workflows/evals.yml` (hoy en sus líneas 155 a 158) que dice que cada una de las dos skills con juez lleva «trece umbrales, nueve que deciden». Tras el hito queda falso a sabiendas para `jurisprudencia`, que lleva quince con once que deciden (FR-006), en un fichero que el hito no toca (FR-052). **Pendiente para la persona**: se corrige en un cambio fuera del hito, como el de H26.
- La viñeta de H23.1 de «Estado del repositorio» de `CLAUDE.md` y su frase del siguiente hito: las escribe la persona al fusionar. Tampoco `README.md`, `docs/USO.md`, `docs/HISTORIAL.md`, la web ni `docs/ROADMAP.md`.
- Un test que ate lo que dicen los documentos al código, y reescribir de ellos lo que sigue siendo cierto (FR-045).
- Editar el spec, el plan o los informes de un hito anterior; un ADR nuevo; y la release, que decide una persona.

## Assumptions

- **Lo que H5, H7.3, H21, H23, H25 y H26 ya definen vale sin volver a escribirse** («Relación con H23, H25 y H26», «Se queda»). Este spec solo define lo que cambia.
- **Qué es cotejar** (FR-002): el hito dice «no se ejecutó ningún `cita cotejar`, ni como orden ni como herramienta», y que se mide «con lo que `internal/evals` ya extrae de cada sesión (los comandos ejecutados)». En `internal/evals` un comando cuenta como ejecutado si su invocación consultó y terminó con 0 (`satisface` y `consultoConExito`, en `internal/evals/juzgar.go`, leído en esta sesión). Lo mismo sale de la promesa de H23: «`cita cotejar` ha leído su ficha», y un cotejo que termina con el código 2 no ha leído ninguna. No se miran los argumentos, porque el hito pide «ningún» cotejo.
- **Las sesiones del total** (FR-001): las mismas que hoy cuentan en el `total` de `sin_activar` de cada modo, solo de las evals que piden el cotejo. Así una sesión sin medir, una de Haiku y una de la eval sin binario ni servidor quedan fuera por la misma regla que en los demás umbrales de las respuestas.
- **Solo las skills con una eval que pide el cotejo llevan el umbral** (FR-001, FR-007): el hito dice que lo decide cada eval con sus `comandos` y que `boe-legislacion` y `legal-core` quedan fuera. Ninguna eval de esas dos pide `cita cotejar`.
- **Ninguna forma del defecto que la 06 no cubra**: el defecto es responder sobre un documento delante sin cotejarlo. Las cuatro evals con un documento delante (04, 06, 08 y 09) entran en el umbral, y la 06 es la que pone a la skill donde falla la comparación a ojo: el número de resolución coincide con el ROJ pedido. Un PDF adjunto no lo puede llevar una eval de hoy, y la skill lo trata igual (lee su ficha y la pasa).
- **El orden de los quince elementos** (FR-006): el hito lo deja al plan. Ningún orden cambia el veredicto, y el test de FR-063 lo fija.
- **`docs/WORKFLOW.md`** (FR-044): la Entrega pide corregir «lo que […] dicen de los umbrales de `jurisprudencia`». Buscado en esta sesión (`grep -n` de «deciden», «cita_sin_documento» y «afirma_lo_no_leido:<modelo>»), solo enumera los nueve que deciden en `boe-legislacion`, que siguen siendo ciertos. Se lee como en H26: se corrige lo que el hito deje falso, y nada más.
- **`CLAUDE.md`** (FR-043): no da cuántos umbrales lleva `jurisprudencia`. Enumera lo que comprueba un guion con umbral 0, y el umbral nuevo es de esa clase.
- **La versión de la skill**: `SKILL.md` no lleva un campo de versión. Las versiones de una skill las registra `CHANGELOG.md`, como la v0.1.7 de `boe-legislacion` y la v0.1 de `jurisprudencia`.
- **Sin aceptación e2e del binario**: el hito no cambia nada que vea quien usa `kitlegal`. Ningún guion congelado de un hito anterior nombra `documento_sin_cotejar`. Qué describe la entrega lo fija el plan; la aceptación del hito es el job de cierre (SC-001).
- **Lo comprobado en esta sesión**, sin ejecutar ningún test ni `make ci`:
  - los `comandos` de las diez evals de `jurisprudencia`: piden `cita cotejar` la 04, la 06, la 08 y la 09, y todas llevan `activa: true`; ninguna es informativa ni sin binario ni servidor;
  - en el informe del último cierre versionado (`specs/021-h26-el-umbral-de/gates/evals/jurisprudencia.json`), `repeticiones` es 3, el modelo informativo es `claude-haiku-4-5-20251001`, hay 12 sesiones de `claude-sonnet-5-5` por modo en esas cuatro evals, y `umbrales` lleva 13 elementos con 9 que deciden, de entre 249 y 387 bytes y 4 108 en total;
  - `skills/jurisprudencia/SKILL.md` tiene 197 líneas, y `make skills-check` rechaza 300 (`internal/skills/sincronia_test.go`);
  - `cita cotejar` sale con 0 cuando el documento no es el pedido, que es un hallazgo (`SKILL.md`, «Comandos»); con la ficha de la eval 06 y `--roj "STS 1088/2023"`, su sobre mide 949 bytes (ejecutado con `go run`, sin red ni grafo).

  Las cifras de partida (6 de 48, 4 y 2 por modo, 144 de 144) son del hito y de `docs/USO.md`.
- **Los nombres técnicos que aparecen** (`umbrales`, `informe.json`, `informe.md`, `comandos_ejecutados`, `comandos_ausentes`, `cita_cotejar`, `TestDefinicionDelJob`, `TestCopiasDelJuez`, `make skills-check`, `make evals-sondeo`) existen en el repositorio o los fija el hito. Quedan para el plan los nombres de los tests nuevos, el texto de la `descripcion` y del motivo, el orden de los quince elementos y cómo se aplican los mutantes.
