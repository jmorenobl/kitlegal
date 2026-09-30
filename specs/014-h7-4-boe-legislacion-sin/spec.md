# Feature Specification: H7.4 · `boe-legislacion` sin el estado de la comprobación ni lo que no ha leído, y un job que juzga la respuesta a la pregunta

**Feature Branch**: `014-h7-4-boe-legislacion-sin`

**Created**: 2026-09-30

**Status**: Draft

**Input**: Sección «#### H7.4 · `boe-legislacion` sin el estado de la comprobación ni lo que no ha leído, y un job que juzga la respuesta a la pregunta» de `docs/ROADMAP.md`, en modo desatendido (ADR 0018).

## Resumen

Es el pulido de `boe-legislacion` antes de la release: ni H7, ni H7.1, ni H7.2 ni H7.3 han salido en ninguna. Quien pregunta tiene que leer en la respuesta lo que dice la norma, leída con el binario y no de memoria, y, si una redacción cambió, la línea `⚠ REDACCIÓN MODIFICADA:` de ese precepto. No tiene que leer nada más sobre la comprobación ni sobre una redacción que la skill no ha leído. El job de evals tiene que medirlo sobre la respuesta que la sesión da a la pregunta.

Se parte de dos mediciones:

- **Con el modelo que decide ahora** (`claude-sonnet-5-5` con Claude Code 2.1.284, ADR 0031): la ejecución 36712391835 del job en la PR #88, sobre `ee8e7ee`, es la línea de base. Da `fallo` solo por la eval 04, que decide (1 de 3), con 0 de 51 respuestas con una expresión de la lista y 421 s. `main` tiene el job en rojo hasta este hito.
- **La del cierre de H7.3**, con el modelo que decidía entonces (Sonnet 5), que queda como contexto y como calibrado. Es `boe-legislacion` v0.1.3 (PR #85, `eafa386`), y el informe del job sobre `196ee05` (`specs/013-h7-3-el-umbral-de/gates/evals/boe-legislacion.json`) da 1 de 51 respuestas con una expresión de la lista, 0 de 30 del modelo informativo y 492 s. Las dos mediciones anteriores de ese cierre son `6ab3add` y `eb6b4c8`.

La bitácora `docs/USO.md` (entrada del 2026-09-30 sobre H7.3, que etiqueta cada respuesta) y el ADR 0031 («Medición del job en la propuesta de cambio») lo leen respuesta a respuesta:

1. **La skill no se activa ante un artículo que el modelo cree saber.** En la eval 04 (LGT, art. 66), las sesiones 01 y 03 del modelo que decide no activan la skill, no ejecutan ninguna orden y responden de memoria. Son 2 de las 51 respuestas en evals que activan la skill. Con el modelo que decidía antes eran 0 de 51, en los cierres de H7.2 y H7.3.
2. **La comprobación, contada con otras palabras (clase A).** En `196ee05`, 7 de 51 respuestas dicen el estado de la comprobación o anuncian lo que el agente tiene o va a hacer, y la lista marca una. En `6ab3add` son 11 (la lista marca 5) y en `eb6b4c8`, 4 (la lista no marca ninguna). Con el modelo que decide ahora son 28 de 51, y la lista no marca ninguna. Viene del vocabulario de `SKILL.md`: «desde una lectura anterior» aparece seis veces en su prosa y «lectura anterior», diez.
3. **Una redacción que ninguna orden devolvió (clase B).** En la eval 19, 5 de las 9 sesiones de esas tres mediciones describen la redacción superada del art. 118 LCSP, y dos lo hacen con datos falsos contra la respuesta grabada. Ninguna orden la devolvió: es contenido legal sin fuente (constitución, principio II). `SKILL.md` ya lo prohíbe, y la prohibición no basta. Contando las dos clases, el cierre de H7.3 da 9 de 51 (17,6 %) y la ejecución del ADR 0031, 28 de 51 (54,9 %).
4. **Ninguna eval lee dos bloques de una misma norma.** Es justo donde el protocolo de H7.3 se aparta del de H7.1, y donde dos líneas `⚠ REDACCIÓN MODIFICADA:` no dicen de qué precepto es cada una.
5. **La orden de lectura y comprobación no funciona en Windows PowerShell 5.1**, que no tiene `&&`. En Windows nativo, Claude Code puede usar su herramienta PowerShell, y el kit se distribuye por Scoop.
6. **El job juzga otra cosa que la respuesta.** Juzga el último `result` del transcript. En `eb6b4c8` juzgó, en la sesión 03-01, la réplica al aviso de una tarea en segundo plano, y además cuenta como limpia una sesión que no terminó.
7. **Dos tandas por commit.** Sobre `6ab3add`, la apertura de la propuesta de cambio y la etiqueta `evals` midieron dos veces el mismo commit: 35 min en lugar de 12.
8. **El sondeo dice sus errores de uso dentro de la traza de `go test`.**

H7.4 no entrega skill nueva: arregla la que existe y el instrumento que la mide (constitución, principio VIII). Va antes de H20 y de la release.

**Cómo se citan los requisitos de otros hitos.** Los requisitos y criterios de H5, H7, H7.1, H7.2 y H7.3 se citan sin guion: «H7.3 FR 034», «H5 FR 077». Los ids con guion (FR-001, SC-001…) son siempre de este spec. Las sesiones se nombran `<eval>-<sesión>` (por ejemplo, `19-02`) y, si hace falta, con el commit medido.

## Criterios del hito, literales

Transcripción literal de `docs/ROADMAP.md` §4, H7.4. El Objetivo y el Alcance se reflejan en los requisitos.

- **Entrega**: «(1) `boe-legislacion` v0.1.4, que se activa ante toda pregunta de su ámbito aunque el modelo crea saber la respuesta, no cuenta la comprobación ni anuncia la respuesta, no describe una redacción que no ha leído, da una línea `⚠ REDACCIÓN MODIFICADA:` por precepto cambiado que dice cuál es, y enseña la orden de lectura y comprobación también para PowerShell; (2) la lista de expresiones prohibidas mide las dos clases del Objetivo por lo que dicen, calibrada con la bitácora, y sin contar las formas fijas que la skill enseña; (3) dos umbrales nuevos que deciden: ninguna respuesta del modelo que decide sin la skill activada en una eval que la activa, y ninguna que describa una redacción que no ha leído; (4) una eval que lee dos bloques de la LCSP con la redacción cambiada; (5) el job y el sondeo juzgan la respuesta a la pregunta y no cuentan en los umbrales lo que no se midió; (6) una sola tanda de sesiones por commit y skill; (7) el sondeo dice sus errores de uso sin la traza; (8) `CHANGELOG.md` (*Unreleased*), con la versión de la skill.»
- **Controles**:
  - «`make ci` en verde, con `skills-check` y `schema-check` sin drift y las reglas del conjunto;»
  - «la sección «Controles de umbral» del plan con cuatro filas: `evals:boe-legislacion:expresiones_prohibidas:<modelo>`, `evals:boe-legislacion:sin_activar:<modelo>` (o el nombre que fije el plan), `evals:boe-legislacion:redaccion_no_leida:<modelo>` y `evals:boe-legislacion:duracion_de_las_sesiones`, con el id del modelo que decide en `<modelo>`;»
  - «el calibrado de la lista sobre los tres informes versionados (36, 11 y 9, eval por eval y con las marcas de hoy), en `TestEvalsDelRepositorio`; y las comprobaciones de H7.2 y H7.3 sobre los bloques grabados, las formas de la skill y la prosa de `SKILL.md`, extendidas a lo que la lista gane;»
  - «tests de `Juzgar`: cada frase de la bitácora, también las de `6ab3add` y `eb6b4c8`, que no están en `main`, y las del modelo que decide ahora que cita este Objetivo, marcada en su clase; las formas fijas de la skill, sin marcar; y «No hay avisos de vigencia sobre este bloque», sin marcar; y una sesión de una eval de `legal-core` que activa `boe-legislacion`, que no pasa;»
  - «una comprobación en `make ci` de que cada orden de lectura y comprobación de `SKILL.md` lleva su forma para PowerShell, con la misma norma y los mismos bloques;»
  - «el control de derivaciones (`TestGrabacionesDerivadas`) con la derivada nueva;»
  - «tests del informe con sesiones sintéticas: `sin_activar` con 1 da `fallo` con su motivo y con 0 se cumple; `redaccion_no_leida`, lo mismo; una sesión sin terminar no cuenta en ningún recuento ni umbral;»
  - «tests de la lectura de la sesión (`TestLeerSesion`): con la réplica a un aviso de una tarea en segundo plano después de la respuesta, lo que se juzga es la respuesta;»
  - «tests del guion del sondeo (`TestGuionDelSondeo`): sin credencial y con un argumento que no vale, la salida de error es solo el mensaje;»
  - «la comprobación de la definición del job (`TestDefinicionDelJob`): un segundo disparo mientras corre o espera una tanda del mismo commit y skill no abre sesiones ni deja una comprobación que el cierre lea como roja, y uno posterior vuelve a medir;»
  - «`CHANGELOG.md` (*Unreleased*), con `boe-legislacion` v0.1.4;»
  - «y el job de evals en la propuesta de cambio.»
- **Aceptación**: «en el informe del job de cierre de `boe-legislacion`, `umbrales` lleva `expresiones_prohibidas:<modelo>` con `decide: true` y `cumple: true` (como mucho 2 de 54, frente a 28 de 51 en la ejecución del ADR 0031 y 9 de 51 en el cierre de H7.3, contadas con las dos clases), `sin_activar:<modelo>` con `decide: true` y `cumple: true` (0, frente a 2 de 51 en la ejecución del ADR 0031), `redaccion_no_leida:<modelo>` con `decide: true` y `cumple: true` (0) y `duracion_de_las_sesiones` con `decide: true` y `cumple: true` (≤ 900 s); la eval 04 pasa, con su tasa (1 de 3 en la ejecución del ADR 0031); el job de `legal-core` sigue aprobado; la eval nueva declara sus dos líneas `⚠ REDACCIÓN MODIFICADA:` y su tasa; ninguna sesión se juzga por una réplica posterior a su respuesta ni queda sin medir por un límite; cada commit medido tiene una sola tanda de sesiones por skill; el veredicto es aprobado (el hito cambia `SKILL.md`, Definition of Done §1.10) y `red` está vacío; y el informe final da los cuatro umbrales como «comprobado por su control».»

En este spec, `<modelo>` es el id del modelo que decide (ADR 0031), hoy `claude-sonnet-5-5`.

Trazabilidad resumida (el detalle, en cada requisito):

| Criterio del hito | Dónde se cumple |
|---|---|
| Entrega (1), `boe-legislacion` v0.1.4 | FR-001 a FR-004, FR-010 a FR-012, FR-020 a FR-027, US1, US2, US3 |
| Entrega (2), la lista por clases | FR-030 a FR-037, US1 |
| Entrega (3), dos umbrales nuevos que deciden | FR-040 a FR-048, US4 |
| Entrega (4), la eval de dos bloques | FR-050 a FR-055, US3 |
| Entrega (5), el juicio sobre la respuesta a la pregunta | FR-060 a FR-062, US5 |
| Entrega (6), una tanda por commit y skill | FR-070 a FR-072, US6 |
| Entrega (7), los errores de uso del sondeo | FR-080 a FR-082, US7 |
| Entrega (8), `CHANGELOG.md` | FR-027, FR-101 |
| Relación con otros hitos | FR-090 |
| Controles | FR-091 a FR-102, SC-002 a SC-012 |
| Aceptación | SC-001 |

## Relación con H7.3, H7.2, H7.1 y H5

`specs/011-h7-1-graph-check-acotado/`, `specs/012-h7-2-la-consulta-repetida/` y `specs/013-h7-3-el-umbral-de/` no se editan: son el registro de sus runs (`plan.md:23` y `research.md:88-94` de H7.3 están desfasados, y siguen siendo registro). No cambia ninguna decisión de arquitectura: el veredicto sigue siendo del job con el modelo que decide (ADR 0031), las repeticiones y la regla por serie siguen como las fija el ADR 0016, y el contrato de `umbrales` es el del ADR 0029. Por eso no hay ADR nuevo (FR-090).

**Sustituye** (lo que dice el hito citado deja de valer, y vale lo de este spec):

- El calibrado de H7.3 (H7.3 FR 021, SC 003 y FR 095: 35 y 10 respuestas, eval por eval): lo sustituye FR-032. Ahora se calibra sobre tres informes, con 36, 11 y 9.
- La viñeta de «Fuera de alcance» de H7.2 y H7.3 que dejaba fuera las paráfrasis («juzgar la redacción libre con un modelo o por similitud […], incluidas las paráfrasis que la lista no recoge, que siguen siendo la limitación declarada de H7.2»), en lo que dice de las paráfrasis de las dos clases: FR-030 y FR-031. Ahora se miden por clases, por sus formas y sin modelo. Juzgar con un modelo o por similitud sigue fuera (ver «Fuera de alcance»).
- H7.3 FR 022, en lo que dice de las formas que enseña la skill: FR-033. Antes de buscar se quitan las formas fijas (FR-031), así que lo que se comprueba es que una respuesta hecha solo de esas formas no se marca. Lo que dice de los bloques grabados se queda y se extiende.
- El contrato del sondeo de H7.3 (`specs/013-h7-3-el-umbral-de/contracts/sondeo.md` §2, punto 3), en lo que dice del registro de `go test` en los errores de uso: FR-080 y FR-081. Los códigos de salida de su §6 se quedan.
- H7.3 FR 034, en la elección entre esperar y saltarse que dejaba al plan: FR-070. El segundo disparo termina sin medir. Lo demás de H7.3 FR 034 se queda: ninguna comprobación que el cierre lea en lugar de la de la tanda que mide.
- El «último `result`» de H5 (`specs/006-h5-skill-boe-legislacion/research.md`, V47, y la lectura de la respuesta de `leerResult`, «cuenta la del último»): FR-060.
- H7.3 FR 002, en su `total` (51) y en lo que cuenta: FR-040 y FR-045. El total pasa a 54 con la eval nueva, y cuentan solo las sesiones que terminaron con una respuesta a la pregunta.
- La viñeta de «Fuera de alcance» de H7.3 «Cambiar las formas de la cita, de los avisos de vigencia o de `⚠ REDACCIÓN MODIFICADA:`», en lo que dice de `⚠ REDACCIÓN MODIFICADA:`: FR-023. La línea dice ahora de qué precepto es. La cita, los avisos y la frase de la regla 7 no cambian.

**Se queda**: la regla por serie del ADR 0016 y H7.2 FR 054 (la lista decide en las positivas con 2 de 3 por serie); H7.2 FR 053 (expresiones por sesión y recuento por modelo, sin la sesión de la prueba de red); H5 FR 077 (la sesión ve la skill tal como la deja `make install`, y `SKILL.md` no nombra evals, el job ni modelos); H7.3 FR 001, FR 003, FR 004, FR 006 y FR 007 (el contrato de `umbrales`; el del modelo informativo con `decide: false`; `legal-core` con `[]`); H7.3 FR 010, FR 013 y FR 091 (la prosa sin expresiones de la lista y sin fechas de ocho cifras), que se extienden a lo que la lista gane; H7.3 FR 030 a FR 037 (el job en paralelo, con su tope, que cubre el peor caso recalculado con las evals del repositorio); H7.3 FR 040 a FR 044 (los límites de uso); H7.3 FR 050 a FR 052 (la duración); H7.3 FR 060 a FR 068 (el sondeo), salvo lo que sustituye FR-080; la comprobación en la misma orden que la lectura, «una vez por cada orden que lee bloques» (cierre de H7.3); y la cota de H7.1 (≤ k señales con k bloques leídos). Todo lo demás de H5 a H7.3 sigue en vigor.

## User Scenarios & Testing *(mandatory)*

### User Story 1 - Quien pregunta lee la norma, sin el estado de la comprobación ni una redacción que nadie leyó (Priority: P1)

Una persona pregunta a `boe-legislacion` por un artículo. La respuesta cita el texto vigente leído con el binario, con su cita y sus avisos de vigencia y, si la redacción cambió, con la línea `⚠ REDACCIÓN MODIFICADA:` de ese precepto. No dice en ninguna parte, ni en ningún idioma, si la comprobación encontró algo o no, ni lo que el agente tiene o va a hacer. Tampoco describe qué decía una redacción que ninguna orden devolvió. Si la persona quiere saber qué cambió, la respuesta dice que cita la redacción vigente y que no ha leído la anterior.

**Why this priority**: es lo que el hito dice que gana la skill. En la línea de base, 28 de 51 respuestas llevan la clase A, y en el cierre de H7.3 dos respuestas daban datos falsos de una redacción que nadie leyó.

**Independent Test**: los tests de `Juzgar` (SC-007), el calibrado (SC-002), las comprobaciones de la prosa y de las formas (SC-003, SC-004) en `make ci`; los umbrales `expresiones_prohibidas:<modelo>` y `redaccion_no_leida:<modelo>` en el job de cierre (SC-001).

**Acceptance Scenarios**:

1. **Dado** la respuesta «La comprobación de redacción no detecta cambios desde una lectura anterior.» (07-03 de la ejecución del ADR 0031), **Cuando** se juzga en una eval que activa la skill, **Entonces** se encuentra una expresión de la clase A y la sesión no pasa (FR-010, FR-030, FR-034).
2. **Dado** la respuesta «Article 59 already answers the question fully, so I don't need art. 60. I have enough to respond now.» (14-01 de `196ee05`), **Cuando** se juzga, **Entonces** se encuentra una expresión de la clase A (FR-010, FR-034).
3. **Dado** una respuesta con la frase de 19-02 de `196ee05` que da la redacción anterior del art. 118 por «vigente hasta el 9 de marzo de 2018», **Cuando** se juzga, **Entonces** se encuentra una expresión de la clase B (FR-011, FR-034).
4. **Dado** una respuesta con «No hay avisos de vigencia sobre este bloque» y nada más de la lista, **Cuando** se juzga, **Entonces** no se encuentra ninguna expresión (FR-012).
5. **Dado** una respuesta con la línea `⚠ REDACCIÓN MODIFICADA:` tal como la enseña la skill —su cita y sus dos fechas, con «la que se consultó antes»— y la frase de la regla 7, **Cuando** se juzga, **Entonces** no se encuentra ninguna expresión. **Dado** las mismas palabras fuera de esas formas, **Cuando** se juzga, **Entonces** se encuentran (FR-031, FR-033).
6. **Dado** los tres informes versionados de H7.1, H7.2 y H7.3, **Cuando** se les aplica la lista, **Entonces** marca exactamente 36, 11 y 9 respuestas, con el reparto por eval de FR-032, y ninguna otra (FR-032).
7. **Dado** `SKILL.md` v0.1.4, **Cuando** se comprueba su prosa fuera de las órdenes, de las formas fijas y de la región generada, **Entonces** no lleva «lectura anterior», «consulta anterior» ni ninguna otra expresión de la lista (FR-021).

---

### User Story 2 - La skill se activa ante toda pregunta de su ámbito, también cuando el modelo cree saber la respuesta (Priority: P1)

Una persona pregunta, con el modelo que decide, «¿En cuántos años prescribe el derecho de la Administración a liquidar una deuda tributaria según el artículo 66 de la Ley General Tributaria?». La skill se activa, lee el bloque con el binario y responde citándolo. No responde de memoria. Ante lo que no es de su ámbito, no se activa: tampoco en las sesiones de `legal-core`, que la tienen instalada.

**Why this priority**: sin activación no hay sobre ni cita, y la respuesta cita un precepto que ninguna orden leyó (principio II). Es el único motivo por el que el job está en rojo en `main`.

**Independent Test**: el umbral `sin_activar:<modelo>` y la serie de la eval 04 en el job de cierre (SC-001); en `make ci`, los tests de `Juzgar` (SC-007) y del informe (SC-006).

**Acceptance Scenarios**:

1. **Dado** `boe-legislacion` v0.1.4 y el job de cierre con el modelo que decide, **Cuando** se juzgan las 54 respuestas de las evals que activan la skill, **Entonces** `sin_activar:<modelo>` tiene `medida` 0, `cumple: true` y `decide: true`, y la serie de la eval 04 pasa (FR-001, FR-041).
2. **Dado** las evals 11 y 12 (no activación), **Cuando** las juzga el job de cierre, **Entonces** sus series pasan (FR-003).
3. **Dado** una sesión de una eval de `legal-core` en la que se activa `boe-legislacion`, **Cuando** se juzga, **Entonces** no pasa, con un motivo que nombra `boe-legislacion` (FR-003, FR-004).
4. **Dado** un informe sintético con 1 respuesta del modelo que decide sin la skill activada en una eval que la activa, **Cuando** se escribe, **Entonces** `sin_activar:<modelo>` lleva `cumple: false` y el veredicto es `fallo` con un motivo que nombra la medida; con 0, `cumple: true` y el veredicto no cambia (FR-041, FR-046).

---

### User Story 3 - Dos preceptos cambiados, dos líneas que dicen cuál es cada uno (Priority: P1)

Una persona que ya preguntó por el art. 118 y por la disposición adicional tercera de la LCSP vuelve a preguntar. Los dos preceptos han cambiado desde entonces. La respuesta lee y comprueba los dos, los cita y lleva dos líneas `⚠ REDACCIÓN MODIFICADA:`, una por precepto. Cada línea dice de qué precepto es, con la cita de su bloque, y lleva sus dos fechas. En Windows, con PowerShell, la skill tiene una orden de lectura y comprobación que funciona.

**Why this priority**: es el caso que ninguna eval cubre y donde dos líneas sin su precepto no se pueden leer.

**Independent Test**: la eval nueva en el job de cierre (SC-001); el control de derivaciones (SC-012) y la comprobación de las formas para PowerShell (SC-005) en `make ci`.

**Acceptance Scenarios**:

1. **Dado** la eval nueva, con el grafo previo de las dos redacciones originales y la caché con las grabadas, **Cuando** la sesión responde, **Entonces** pasa solo si lee `a1-30` y `da-3` de `BOE-A-2017-12902` (en una orden o en dos), comprueba con `graph check` y la norma, no pide `graph show`, cita los dos bloques y lleva dos líneas `⚠ REDACCIÓN MODIFICADA:`: una de `a1-30` con 20180309 y 20200206 y otra de `da-3` con 20180309 y 20230101, cada una con la cita de su bloque; y sin expresiones prohibidas (FR-052).
2. **Dado** el informe del job de cierre, **Cuando** se lee la eval nueva, **Entonces** declara sus dos líneas esperadas y publica su tasa, como informativa (FR-053, FR-054).
3. **Dado** `SKILL.md` v0.1.4, **Cuando** se comprueba en `make ci`, **Entonces** cada orden de lectura y comprobación (la de `articulo` y la de `articulos`) lleva al lado su forma para PowerShell, con la misma norma y los mismos bloques, y con la forma condicional que fija el plan, que no ejecuta la comprobación si la lectura falla (FR-024, FR-095).
4. **Dado** las derivadas del grafo previo, **Cuando** corre `TestGrabacionesDerivadas`, **Entonces** la de `da-3` sale de la respuesta grabada quitándole solo la redacción posterior (FR-051).

---

### User Story 4 - Los umbrales nuevos los hace cumplir el job (Priority: P1)

El informe del job de `boe-legislacion` publica, con el contrato del ADR 0029, cuatro umbrales que deciden: expresiones prohibidas (las dos clases), respuestas sin la skill activada, respuestas con una redacción no leída y duración. Si alguno no se cumple, el veredicto es `fallo` y el job sale en rojo.

**Why this priority**: sin control, una respuesta sin fuente en una de cada cincuenta volvería a salir en verde.

**Independent Test**: tests del informe con sesiones sintéticas en `make ci` (SC-006); el job de cierre (SC-001).

**Acceptance Scenarios**:

1. **Dado** 3 de 54 respuestas del modelo que decide con alguna expresión, **Cuando** se escribe el informe, **Entonces** `expresiones_prohibidas:<modelo>` lleva `cumple: false` y el veredicto es `fallo`; con 2 de 54, `cumple: true` (FR-040).
2. **Dado** 1 respuesta del modelo que decide con una expresión de la clase B en una eval que activa la skill, **Cuando** se escribe el informe, **Entonces** `redaccion_no_leida:<modelo>` lleva `medida` 1, `cumple: false` y el veredicto es `fallo` con un motivo que nombra la medida; con 0, `cumple: true` y el veredicto no cambia (FR-042, FR-046).
3. **Dado** 3 de 54 respuestas con alguna expresión y 6 sesiones más que no terminaron, **Cuando** se escribe el informe, **Entonces** `expresiones_prohibidas:<modelo>` lleva `medida` 3 y `total` 54, no 60, y no se cumple (FR-045).

---

### User Story 5 - El job juzga la respuesta que la sesión da a la pregunta (Priority: P2)

El agente lanza una orden en segundo plano, responde a la pregunta y, al llegar el aviso de que la tarea terminó, escribe una réplica. El job juzga la respuesta a la pregunta, no la réplica. Una sesión que no terminó se publica con su motivo y no cuenta como respuesta limpia.

**Why this priority**: sin esto, la tasa de una serie y el recuento de los umbrales miden un texto que nadie leyó como respuesta.

**Independent Test**: `TestLeerSesion` (SC-008) y los tests del informe (SC-006) en `make ci`.

**Acceptance Scenarios**:

1. **Dado** un transcript con la respuesta a la pregunta y, después, el aviso de una tarea en segundo plano y su réplica, **Cuando** se lee la sesión, **Entonces** la respuesta que se juzga es la de la pregunta (FR-060).
2. **Dado** una sesión que no terminó, **Cuando** se escribe el informe o la salida del sondeo, **Entonces** se publica con su motivo y no entra en ninguna medida ni en ningún total de recuentos o umbrales (FR-061).

---

### User Story 6 - Una tanda de sesiones por commit y skill (Priority: P2)

El cierre del workflow abre la propuesta de cambio y pone la etiqueta `evals`: dos disparos sobre el mismo commit. Solo uno abre sesiones. El otro termina sin medir y sin dejar una comprobación que el cierre lea como roja o en lugar de la buena. La etiqueta sobre un commit ya medido vuelve a medir.

**Why this priority**: cada tanda duplicada gasta el doble de sesiones de la suscripción y retrasa el cierre de 12 a 35 min.

**Independent Test**: `TestDefinicionDelJob` en `make ci` (SC-010).

**Acceptance Scenarios**:

1. **Dado** una tanda de `boe-legislacion` sobre un commit que corre o espera, **Cuando** llega un segundo disparo sobre el mismo commit, **Entonces** no abre sesiones, no espera para volver a medir, termina sin medir y no deja una comprobación que el cierre lea como roja o en lugar de la de la tanda que mide (FR-070).
2. **Dado** un commit cuya tanda ya terminó, **Cuando** se pone la etiqueta `evals`, **Entonces** se abre una tanda nueva que vuelve a medir (FR-071).

---

### User Story 7 - Quien lanza el sondeo lee su error de uso, no una traza (Priority: P3)

Una persona lanza el sondeo sin `CLAUDE_CODE_OAUTH_TOKEN`, o con un argumento que no vale. Lee solo el mensaje, un error por línea, sin nada de `go test` ni de testify.

**Why this priority**: es un error de quien lo usa; la traza lo esconde. No afecta al veredicto.

**Independent Test**: `TestGuionDelSondeo` en `make ci` (SC-009).

**Acceptance Scenarios**:

1. **Dado** el entorno sin `CLAUDE_CODE_OAUTH_TOKEN`, **Cuando** se lanza el sondeo, **Entonces** la salida de error es solo el mensaje de la credencial, sale con el código de hoy y no abre ninguna sesión (FR-080).
2. **Dado** `REPETICIONES=0` y `MODELO` vacío, **Cuando** se lanza el sondeo, **Entonces** la salida de error son las dos líneas de sus errores, en el orden de hoy, y nada más (FR-080).
3. **Dado** un binario que no se puede construir, **Cuando** se lanza el sondeo, **Entonces** imprime el registro de `go test`, como hoy (FR-081).

---

### Edge Cases

- **3 de 54 y 2 de 54**: 3/54 ≈ 0,0556 > 0,05, no cumple; 2/54 ≈ 0,0370, cumple. Con 3 de 54 y seis sesiones sin terminar, el total es 54 y no se cumple; contadas como limpias (3 de 60 = 0,05) se cumpliría (FR-045).
- **Sin respuestas que contar**: si todas las sesiones del modelo que decide en las evals que activan la skill quedan sin medir o sin terminar, el total de cada umbral es 0 y la proporción es 0 (contrato del ADR 0029). El veredicto ya es `fallo`, por las sesiones sin medir (H7.3 FR 043) o por las series que no pasan.
- **Una respuesta con una expresión de la clase B**: cuenta en `redaccion_no_leida:<modelo>` y en `expresiones_prohibidas:<modelo>`; una respuesta cuenta una sola vez en cada umbral, lleve las expresiones que lleve.
- **Una sesión sin la skill activada que además lleva una expresión**: cuenta en `sin_activar:<modelo>` y en los umbrales de la lista que correspondan.
- **La eval nueva leída en una sola orden** (`kitlegal boe articulos … a1-30 da-3 && kitlegal graph check …`): satisface la lectura de los dos bloques y la comprobación, igual que dos órdenes (FR-052).
- **Una sola línea `⚠ REDACCIÓN MODIFICADA:` en la eval nueva, o dos sin su bloque, o con las fechas cambiadas**: la sesión no pasa (FR-052).
- **Quien pregunta quiere saber qué cambió**: la respuesta dice que cita la redacción vigente y que no ha leído la anterior (FR-022). Eso no es clase B ni se marca (FR-034); describir qué decía la anterior, sí.
- **Palabras de la lista dentro de una forma fija y fuera de ella en la misma línea**: lo de la forma no cuenta; lo de fuera, sí (FR-031).
- **Que no hay avisos y, en la misma frase, el estado de la comprobación** («No hay avisos de vigencia sobre este bloque ni cambios de redacción respecto a una consulta anterior.», 01-02 de `196ee05`): lo de los avisos no cuenta; lo de la comprobación, sí, y la respuesta se marca en la clase A (FR-012).
- **La sesión de la prueba de red**: no entra en ninguna medida ni total de los umbrales de la lista ni en `sin_activar` (H7.2 FR 053); sí en la duración (H7.3 FR 050).
- **Una tarea en segundo plano cuya réplica llega antes de la respuesta a la pregunta**: lo que se juzga sigue siendo la respuesta del turno de la pregunta (FR-060).
- **Lo dicho con palabras que ninguna forma de la lista recoge**: no se detecta, y queda a la vista en el informe, que publica cada respuesta (FR-037).
- **El segundo disparo cuando la primera tanda ya terminó**: vuelve a medir (FR-071).

## Requirements *(mandatory)*

### Functional Requirements

#### La activación

- **FR-001**: `boe-legislacion` v0.1.4 MUST activarse ante toda pregunta de su ámbito: qué dice una norma o un artículo de la normativa consolidada del BOE, o dónde se regula una materia. También cuando el modelo cree saber la respuesta, porque la activación es lo que hace que la respuesta cite con fuente y no de memoria. Comprobable: en el job de cierre, `sin_activar:<modelo>` es 0 (FR-041) y la serie de la eval 04 pasa (SC-001).
- **FR-002**: El research MUST explicar por qué el modelo que decide no activa la skill en la eval 04, con las sesiones 04-01 y 04-03 de la ejecución 36712391835 y con lo que registra el ADR 0031 de su sondeo. El cambio MUST ir a lo que decide la activación —la `description` de `SKILL.md` y lo que haga falta de su cuerpo— y cada cambio se traza a esa causa. La eval 04 y su pregunta MUST NOT tocarse: se corrige la clase de defecto. Comprobable: el research da la causa con esa evidencia y el plan traza a ella cada cambio.
- **FR-003**: La skill MUST NOT activarse con lo que no es de su ámbito. En el job de cierre, las series de las evals de no activación del conjunto (11 y 12) pasan. En las evals de `legal-core`, cuyas sesiones ven también `boe-legislacion` (`make install` instala todas las skills del binario), una sesión que active `boe-legislacion` MUST NOT pasar, con un motivo que nombra `boe-legislacion`. El plan fija cómo lo declara el formato de eval. Lo que cambie bajo `schemas/` va en tareas `[datos]`.
- **FR-004**: Las tres evals de `legal-core` MUST declarar que `boe-legislacion` no se activa. Comprobable: el test de `Juzgar` de FR-094.

#### Lo que la respuesta no lleva

- **FR-010**: Salvo las formas fijas que enseña la skill (FR-031), la respuesta MUST NOT llevar, en ninguna parte ni en ningún idioma, la **clase A**: el estado de la comprobación o su anuncio. Es decir, que la redacción cambió o no respecto de una lectura o consulta anterior, que no hay nada que señalar, que la comprobación terminó, o lo que el agente tiene, necesita o va a hacer («tengo suficiente para responder», «no necesito el artículo 60»). Lo mide la lista (FR-030) en el umbral de FR-040.
- **FR-011**: La respuesta MUST NOT llevar la **clase B**: una redacción que ninguna orden de la sesión devolvió, es decir, qué decía o qué cambió respecto de ella. Sí puede decir lo que devuelven las órdenes: la norma modificadora y la fecha de vigencia de la redacción vigente, y las dos fechas de la línea `⚠ REDACCIÓN MODIFICADA:`. Lo mide la lista (FR-030) en los umbrales de FR-040 y FR-042.
- **FR-012**: Una frase que solo dice que no hay avisos de vigencia o que la norma no está derogada, como «No hay avisos de vigencia sobre este bloque» sola, MUST NOT contarse como clase A: es derecho (H7.3 FR 020). La lista no gana formas que casen con eso. Lo que la misma frase o línea diga además del estado de la comprobación o de una lectura o consulta anterior sí es clase A (FR-010), como en 01-02 de `196ee05` y 14-01 de `eb6b4c8` (FR-034).

#### `boe-legislacion` v0.1.4

- **FR-020**: El research MUST trazar cada cambio de `SKILL.md` a su causa, con los informes del cierre de H7.3 (los tres commits) y la ejecución del ADR 0031, como en H7.2 y H7.3: dónde enseña `SKILL.md` v0.1.3 el vocabulario de la clase A, línea a línea, y por qué el modelo describe una redacción que no ha leído. Una prohibición más no satisface este requisito, porque la de la clase B ya está («Lo único que la respuesta dice de una lectura anterior es esa línea»). Comprobable: el research da la causa de cada clase con esas medidas y el plan traza a ella cada cambio.
- **FR-021**: La prosa de `SKILL.md` —frontmatter incluido— MUST NOT llevar el vocabulario de la clase A («lectura anterior», «consulta anterior» y toda expresión que la lista gane). Quedan fuera de esta exigencia tres sitios: las órdenes, las formas fijas y la región generada de la tabla de comandos. La comprobación de H7.3 sobre la prosa (H7.3 FR 091) ya lo exige de toda expresión de la lista, y se extiende a la lista nueva. Comprobable: FR-093.
- **FR-022**: Para la clase B, `SKILL.md` MUST decir tres cosas. Primero, por qué la respuesta no puede describir la redacción superada: no la ha leído en esta pregunta, y el binario solo da la vigente y dos fechas. Segundo, qué sí puede decir (FR-011). Tercero, qué responde si quien pregunta quiere saber qué cambió: que la respuesta cita la redacción vigente y que no ha leído la anterior. Leer la anterior es H20. Comprobable: el research traza el cambio (FR-020); su efecto lo mide `redaccion_no_leida:<modelo>` (FR-042).
- **FR-023**: Cada línea `⚠ REDACCIÓN MODIFICADA:` MUST decir de qué precepto es, con la cita de su bloque (la forma de «Cómo se cita»), y seguir llevando en la misma línea la etiqueta y sus dos fechas tal como las da el binario. Hay una línea por bloque cuya redacción cambió. La Entrega (1) pide «una línea por precepto cambiado que dice cuál es», así que la forma es la misma con uno o con varios bloques cambiados. El plan fija la forma, con marcadores en lugar de datos en el ejemplo (H7.3 FR 013). Comprobable: la eval nueva (FR-052) y las formas de FR-033.
- **FR-024**: Cada orden de `SKILL.md` que lee y comprueba (`kitlegal boe articulo … --json && kitlegal graph check … --json`, y la de `articulos`) MUST llevar al lado su forma para PowerShell. Esa forma cumple cuatro condiciones: funciona en Windows PowerShell 5.1 y en PowerShell 7; lleva la misma norma y los mismos bloques que la orden de Bash; hace la lectura y, detrás, en la misma orden, la comprobación; y, como `&&`, no ejecuta la comprobación si la lectura termina con un código distinto de 0. El plan fija la forma. Comprobable: FR-095.
- **FR-025**: MUST quedarse de v0.1.3: la comprobación detrás de la lectura, en su misma orden y con su misma norma y sus mismos bloques, una vez por cada orden que lee bloques; la lectura de los bloques de uno en uno (paso 3); la forma de la cita y la de los avisos de vigencia; la regla 7 con su frase fija, «No se ha podido comprobar si la redacción ha cambiado desde una consulta anterior.»; lo que dicen las reglas 1 a 6 y «Nada de otra conversación»; y ninguna fecha `AAAAMMDD` escrita con cifras (H7.3 FR 013). La redacción de todo eso cambia solo donde lleve vocabulario de la lista (FR-021).
- **FR-026**: `SKILL.md` MUST seguir por debajo de 300 líneas, con frontmatter válido (la `description`, hasta 1024 caracteres), la región generada intacta (`make skills-check` sin drift) y sin nombrar evals, el job ni modelos (H5 FR 077).
- **FR-027**: `CHANGELOG.md` (*Unreleased*) MUST registrar `boe-legislacion` v0.1.4 con lo que cambia para quien la usa: la activación, la respuesta sin la clase A ni la B, la línea `⚠ REDACCIÓN MODIFICADA:` con su precepto y la orden para PowerShell.

#### La lista, por clases

- **FR-030**: `evals/boe-legislacion/expresiones-prohibidas.yaml` MUST ganar lo que haga falta para las dos clases, con las formas que las nombran: la lectura o consulta anterior, la comprobación, el cambio de redacción, el anuncio del agente, lo que decía o exigía la redacción anterior u original… El plan fija la lista y sus familias. No se ganan como frases sueltas de la bitácora. La lista MUST distinguir qué familias son de la clase B, porque son las que mide `redaccion_no_leida:<modelo>` (FR-042). Se compara como en H7.2 (H7.2 FR 051), sin juicio de ningún modelo. Las familias de hoy (maquinaria, otra conversación y anuncio) se quedan.
- **FR-031**: Antes de buscar, MUST quitarse de la respuesta las formas fijas de la skill, tal como la skill las enseña. Son tres: la línea `⚠ REDACCIÓN MODIFICADA:` (su etiqueta, su texto fijo, la cita de su bloque y sus dos fechas); la marca, la etiqueta y los dos puntos de cada aviso de vigencia; y la frase fija de la regla 7. Lo demás de cada línea se busca. Así, «la que se consultó antes» de la línea `⚠ REDACCIÓN MODIFICADA:` y «desde una consulta anterior» de la regla 7 no cuentan, y las mismas palabras fuera de esas formas sí.
- **FR-032**: El calibrado no empeora. Sobre los tres informes versionados, la lista MUST marcar, eval por eval, exactamente las respuestas que marca hoy más las que etiqueta la bitácora del 2026-09-30, y ninguna otra; ninguna que hoy marca deja de marcarse:
  - en el de H7.1 (`specs/011-h7-1-graph-check-acotado/gates/evals/boe-legislacion.json`), 36 de 93: las 35 de hoy —02 (1), 03 (3), 04 (3), 05 (3), 06 (3), 07 (3), 08 (2), 09 (2), 13 (3), 14 (3), 15 (3), 16 (2), 17 (3) y 19 (1)— y 19-01 (B), así que 19 (2);
  - en el de H7.2 (`specs/012-h7-2-la-consulta-repetida/gates/evals/boe-legislacion.json`), 11 de 93: las 10 de hoy —03 (1), 06 (1), 13 (2), 14 (2), 15 (3) y 19 (1), la 19-02— y 19-01 (B), así que 19 (2);
  - en el de H7.3 (`specs/013-h7-3-el-umbral-de/gates/evals/boe-legislacion.json`, `196ee05`), 9 de 93: 01 (1: 01-02), 05 (1: 05-02), 13 (2: 13-02 y 13-03), 14 (3: 14-01, 14-02 y 14-03), de la clase A, y 19 (2: 19-01 y 19-02), de la clase B.

  En los tres, las respuestas de la clase B las marca una familia de la clase B. Comprobable: FR-093.
- **FR-033**: Comparada como en H7.2 FR 051, ninguna expresión de la lista MUST aparecer en el texto de los bloques grabados para las evals de `boe-legislacion`, incluidos los de la eval nueva. Además, una respuesta hecha solo de las formas que la skill enseña a escribir, con datos de ejemplo, MUST NOT marcarse una vez quitadas sus formas fijas (FR-031). Esas formas son la cita, los avisos de vigencia, `⚠ REDACCIÓN MODIFICADA:` con su cita y la frase fija de la regla 7. Son las comprobaciones de H7.2 FR 085 y de H7.3 FR 022, extendidas a lo que la lista gane. Comprobable: FR-093.
- **FR-034**: `Juzgar` MUST marcar, cada una en su clase, estas frases, tomadas enteras de la respuesta de su informe cuando la bitácora cita solo un fragmento:
  - de `196ee05`, las de 01-02, 05-02, 13-02, 13-03, 14-01, 14-02 y 14-03 (A) y las de 19-01 y 19-02 (B);
  - de `6ab3add` y `eb6b4c8`, que no están en `main`: las de la clase A que cita la bitácora (03-03 y 13-03 de `6ab3add`, y 14-01 de `eb6b4c8`), y las de las tres sesiones que etiqueta de la clase B (19-01 de `6ab3add`, y 19-01 y 19-02 de `eb6b4c8`);
  - de los informes de H7.1 y H7.2, 19-01 de H7.1, y 19-01 y 19-02 de H7.2 (B);
  - del modelo que decide ahora, las que cita el Objetivo del hito: «La comprobación de redacción no detecta cambios desde una lectura anterior.» (07-03), «Es un extracto fiel del texto consolidado. Aquí va apartado por apartado.» (01-03) y «Respondo de memoria, sin haber consultado el texto consolidado del BOE.» (04), todas de la clase A.

  MUST NOT marcar las formas fijas de la skill (FR-031), «No hay avisos de vigencia sobre este bloque» (FR-012) ni lo que `SKILL.md` enseña a responder a quien quiere saber qué cambió (FR-022): que la respuesta cita la redacción vigente y que no ha leído la anterior. El hito no lo cuenta como clase B: con el modelo que decide, la 19-02 lo dice y no describe la redacción superada. Comprobable: FR-094.
- **FR-035**: Lo que cambie en `schemas/expresiones-prohibidas.yaml.json` (las familias nuevas y su clase) MUST ir en una tarea `[datos]`. Una lista que no cumple el esquema sigue siendo un fichero mal formado (H7.2 FR 055).
- **FR-036**: Las familias nuevas MUST aplicarse como las de hoy. Esto vale en `Juzgar`, a cada sesión de cada eval que activa la skill; en la regla por serie; en las expresiones por sesión y el recuento por modelo; en los umbrales (FR-040, FR-042 y el del modelo informativo, FR-044); y en el sondeo.
- **FR-037**: Lo dicho con palabras que ninguna forma de la lista recoge sigue sin detectarse. Queda a la vista en el informe, que publica cada respuesta (H7.2 FR 053).

#### Umbrales

Cada uno es un elemento de `umbrales` de `informe.json` de `boe-legislacion`, con el contrato del ADR 0029 (H7.3 FR 001). Con el conjunto de FR-054, las evals que activan la skill dan 54 respuestas del modelo que decide.

- **FR-040**: `expresiones_prohibidas:<modelo>` MUST seguir con `comparacion` `"<="`, `umbral` 0.05 y `decide: true`. Su `medida` son las respuestas del modelo que decide con alguna expresión de cualquier familia de la lista, ya con las dos clases. Su `total` son las respuestas del modelo que decide en las evals que activan la skill, informativas incluidas (54 con todas terminadas). Contadas las dos clases, el cierre de H7.3 da 9 de 51 (lo comprueba el calibrado de FR-032) y la ejecución del ADR 0031, 28 de 51 según la lectura del hito (su informe no está versionado): los dos, `fallo`. Con 54, el umbral admite 2.
- **FR-041**: `sin_activar:<modelo>` MUST ser un umbral nuevo. Su `medida` son las respuestas del modelo que decide, en las evals que activan la skill, en las que la skill no se activó. Su `total` son las respuestas del modelo que decide en esas evals. Lleva `comparacion` `"<="`, `umbral` 0, `decide: true` y una `descripcion` que nombra la medida. El umbral es 0 porque el principio II no admite una proporción de respuestas sin fuente, y el modelo que decidía antes dio 0 de 51 en dos cierres seguidos (H7.2 y H7.3). En la ejecución del ADR 0031 fueron 2 de 51.
- **FR-042**: `redaccion_no_leida:<modelo>` MUST ser un umbral nuevo. Su `medida` son las respuestas del modelo que decide con alguna expresión de una familia de la clase B (FR-030). Su `total` son las respuestas del modelo que decide en las evals que activan la skill. Lleva `comparacion` `"<="`, `umbral` 0, `decide: true` y una `descripcion` que nombra la medida. El umbral es 0 porque el principio II no admite una proporción. Solo es del modelo que decide: el informativo no abre las evals informativas (ADR 0016), que son las que traen una redacción cambiada. En el cierre de H7.3 dio 2, y en la ejecución del ADR 0031, 0.
- **FR-043**: `duracion_de_las_sesiones` MUST seguir con `comparacion` `"<="`, `umbral` 900, `decide: true` y sin `total` (H7.3 FR 051), ahora con las 96 sesiones del plan (FR-054). En la ejecución del ADR 0031 dio 421 s con 93 sesiones.
- **FR-044**: `expresiones_prohibidas:` del modelo informativo (`claude-haiku-4-5-20251001`) MUST seguir con `decide: false` (H7.3 FR 004), contado con la lista nueva. El modelo informativo no tiene `sin_activar` ni `redaccion_no_leida`.
- **FR-045**: En la `medida` y en el `total` de todos los umbrales de FR-040 a FR-042 y del recuento de expresiones por modelo, MUST contar solo las sesiones que terminaron con una respuesta a la pregunta (FR-060). No cuentan ni la sesión de la prueba de red (H7.2 FR 053) ni las sin medir (H7.3 FR 040). Las que no terminaron se publican con su motivo (H7.3) y no cuentan, ni arriba ni abajo de la proporción. Comprobable: con 3 de 54 y seis sesiones más sin terminar, `expresiones_prohibidas:<modelo>` da 3 de 54 y no se cumple (FR-097).
- **FR-046**: Un umbral de FR-040 a FR-043 que no se cumple MUST poner el veredicto en `fallo`, con un motivo que nombra el umbral, la medida y, si lo tiene, el total, y el job MUST salir en rojo (H7.3 FR 003). Uno que se cumple MUST NOT cambiar el veredicto.
- **FR-047**: Ninguna corrección del run MUST cumplir un umbral de FR-040 a FR-043 rebajándolo, dejándolo en `decide: false`, sacando evals del total o recortando la lista (ADR 0029).
- **FR-048**: `legal-core` MUST seguir sin umbrales: su `umbrales` es `[]` (H7.3 FR 006).

#### La eval de dos bloques

- **FR-050**: Una eval nueva de `boe-legislacion`, informativa (ADR 0016) y que activa la skill, MUST preguntar por el art. 118 (bloque `a1-30`) y por la disposición adicional tercera (bloque `da-3`, normas específicas de contratación en las entidades locales, que nombra los contratos menores) de la LCSP (`BOE-A-2017-12902`). La pregunta dice que ya se preguntó por los dos preceptos, por ejemplo: «Hace tiempo te pregunté qué exige la LCSP para el expediente de un contrato menor, en su artículo 118, y qué añade su disposición adicional tercera para los ayuntamientos. ¿Qué dicen ahora?». Nace sin tasa medida.
- **FR-051**: El grafo previo de la eval MUST tener una lectura de cada bloque que vio su redacción original: `a1-30` con vigencia 20180309 y `da-3` con vigencia 20180309. Cada lectura se deriva de su respuesta grabada quitándole solo la redacción posterior, sin nada escrito a mano. El control de derivaciones (`TestGrabacionesDerivadas`, H7.2) MUST cubrir la derivada nueva. La caché sirve las respuestas grabadas, que traen las dos redacciones de cada bloque: `a1-30`, 20180309 y 20200206 (`BOE-A-2020-1651`); `da-3`, 20180309 y 20230101 (`BOE-A-2022-22128`).
- **FR-052**: La eval MUST exigir, comparado sin modelo:
  - leer `a1-30` y `da-3` de `BOE-A-2017-12902`, en una orden con los dos o en una por bloque;
  - `graph check` con la norma;
  - no pedir `graph show`;
  - citar los dos bloques;
  - dos líneas `⚠ REDACCIÓN MODIFICADA:`, una por bloque, cada una con la cita de su bloque y sus dos fechas: `a1-30` con 20180309 y 20200206, y `da-3` con 20180309 y 20230101;
  - sin expresiones prohibidas.

  Una sola línea, dos líneas sin su bloque, o unas fechas que no son las de su bloque, hacen que la sesión no pase.
- **FR-053**: El formato de eval MUST poder declarar las líneas `⚠ REDACCIÓN MODIFICADA:` esperadas, cada una con su bloque y sus dos fechas. El informe MUST publicar, por sesión, las que encontró y las que faltan, y la tasa de la eval. El plan fija cómo. Lo que cambie bajo `schemas/` va en tareas `[datos]`. Las evals de hoy (también la 19) conservan su juicio.
- **FR-054**: El conjunto de `boe-legislacion` MUST pasar a 20 evals: 10 positivas que deciden (01 a 10), 2 de no activación que deciden (11 y 12) y 8 informativas (13 a 20). Da 18 evals que activan la skill, 54 respuestas del modelo que decide en ellas y 30 del informativo, y 96 sesiones (60 del modelo que decide y 36 del informativo). El tope del trabajo sigue cubriendo el peor caso, recalculado con las evals del repositorio (H7.3 FR 035).
- **FR-055**: Si la preparación de la eval necesita una grabación que aún no está bajo `testdata/evals/`, MUST grabarla el paso `grabar_datos` (fila `boe.legislacion-consolidada` revisada en `docs/SOURCES.md`). La derivada y la eval van en tareas `[datos]`.

#### El juicio sobre la respuesta a la pregunta

- **FR-060**: Lo que se juzga de una sesión —citas, avisos, formas y expresiones— MUST ser la respuesta que la sesión da a la pregunta: el resultado del turno que abre la pregunta. Lo que escribe después, como réplica al aviso de una tarea en segundo plano, MUST NOT sustituirla. Lo comparten el job y el sondeo. Comprobable: FR-098.
- **FR-061**: Los recuentos y los umbrales del informe y el recuento de la salida del sondeo MUST contar solo las sesiones que terminaron con una respuesta a la pregunta (FR-045). Una sesión que no terminó se sigue publicando con su motivo y sigue sin pasar en su serie (H5), pero no cuenta en ningún recuento ni umbral. Comprobable: FR-097.
- **FR-062**: MUST NOT cambiar nada más del juicio de cada sesión ni de la regla por serie.

#### Una tanda por commit y skill

- **FR-070**: Un segundo disparo del flujo de evals sobre un commit y una skill cuya tanda corre o espera MUST NOT abrir sesiones ni esperar para volver a medir: termina sin medir. MUST NOT dejar ninguna comprobación que el cierre del workflow lea como roja o en lugar de la de la tanda que mide. El cierre sigue esperando a esa tanda y recogiendo su informe: `scripts/workflow/cierre.sh`, en `medir`, espera mientras haya comprobaciones pendientes y cuenta como roja una cancelada, y en `recoger_evals` copia el informe de las de `evals` que acaban en `pass` o `fail`. Comprobable: FR-100.
- **FR-071**: La etiqueta `evals` sobre un commit cuya tanda ya terminó MUST volver a medir, como hoy: es el botón de «vuelve a medir» (ADR 0016).
- **FR-072**: FR-070 y FR-071 MUST conseguirse en el job (`.github/workflows/evals.yml` e `internal/evals`), sin cambiar `scripts/workflow/`, que es el proceso del workflow y no del hito. Si no se pudiera sin tocarlo, queda como supuesto en `gates/supuestos.md`, y el cambio del cierre va en su propia rama, fuera de este hito.

#### Los errores de uso del sondeo

- **FR-080**: Con un argumento que no vale o sin la credencial, la salida de error del sondeo MUST ser solo su mensaje, sin nada de `go test` ni de testify: cada error en una línea, con el texto y en el orden en que los da hoy `TestSondeo` (`specs/013-h7-3-el-umbral-de/contracts/sondeo.md` §3, puntos 1 y 2). MUST salir con el código de hoy (su §6: 1 el guion, 2 `make`) y MUST NOT abrir ninguna sesión. Comprobable: FR-099.
- **FR-081**: Un fallo al construir el binario, al instalar las skills o al repartir las sesiones MUST seguir imprimiendo el registro de `go test`, que es lo que necesita quien desarrolla.
- **FR-082**: La salida del sondeo MUST contar su recuento de expresiones con FR-060 y FR-061. Lo demás de su salida (H7.3 FR 065) no cambia.

#### Relación con otros hitos

- **FR-090**: `specs/011-…`, `specs/012-…` y `specs/013-…` MUST NOT editarse. Este spec nombra lo que sustituye de ellos y de H5 («Relación con H7.3, H7.2, H7.1 y H5»). No hay ADR nuevo.

#### Controles y Definition of Done

- **FR-091**: `make ci` MUST quedar en verde, con `skills-check` y `schema-check` sin drift y las reglas del conjunto.
- **FR-092**: La sección «Controles de umbral» del plan MUST tener cuatro filas: FR-040 en `evals:boe-legislacion:expresiones_prohibidas:<modelo>`; FR-041 en `evals:boe-legislacion:sin_activar:<modelo>`; FR-042 en `evals:boe-legislacion:redaccion_no_leida:<modelo>`; y FR-043 en `evals:boe-legislacion:duracion_de_las_sesiones`. En `<modelo>` va el id del modelo que decide, hoy `claude-sonnet-5-5`. El del modelo informativo (FR-044) no tiene fila.
- **FR-093**: `TestEvalsDelRepositorio` MUST comprobar en `make ci` tres cosas: el calibrado de FR-032 sobre los tres informes, eval por eval y con las marcas de hoy; las comprobaciones de FR-033 sobre los bloques grabados y las formas de la skill; y la de FR-021 sobre la prosa de `SKILL.md`, extendida a la lista nueva. Falla si la lista marca otra cosa.
- **FR-094**: Los tests de `Juzgar` MUST cubrir: cada frase de FR-034, marcada en su clase; las formas fijas de la skill, sin marcar; «No hay avisos de vigencia sobre este bloque» y la respuesta de FR-022 a quien quiere saber qué cambió, sin marcar; y una sesión de una eval de `legal-core` que activa `boe-legislacion`, que no pasa (FR-003).
- **FR-095**: Una comprobación en `make ci` MUST fallar si alguna orden de lectura y comprobación de `SKILL.md` no lleva su forma para PowerShell, si esa forma no lleva la misma norma y los mismos bloques, o si no es la forma condicional que fija el plan (FR-024).
- **FR-096**: El control de derivaciones (`TestGrabacionesDerivadas`) MUST cubrir la derivada nueva (FR-051).
- **FR-097**: Los tests del informe con sesiones sintéticas MUST cubrir:
  - `sin_activar:<modelo>` con 1, que da `cumple: false` y el veredicto `fallo` con su motivo, y con 0, que se cumple sin cambiar el veredicto;
  - lo mismo para `redaccion_no_leida:<modelo>`;
  - `expresiones_prohibidas:<modelo>` con 3 de 54 (`fallo`) y con 2 de 54 (se cumple);
  - una sesión sin terminar, que no cuenta en ningún recuento ni umbral: con 3 de 54 y seis sin terminar, el total es 54 y el umbral no se cumple.
- **FR-098**: `TestLeerSesion` MUST cubrir un transcript con la respuesta a la pregunta y, después, el aviso de una tarea en segundo plano y su réplica: lo que se juzga es la respuesta a la pregunta.
- **FR-099**: `TestGuionDelSondeo` MUST cubrir dos casos: sin credencial y con un argumento que no vale, la salida de error es solo el mensaje, con el código de hoy y sin abrir ninguna sesión; y un fallo de construcción sigue imprimiendo el registro.
- **FR-100**: `TestDefinicionDelJob` MUST fallar si un segundo disparo, mientras corre o espera una tanda del mismo commit y skill, puede abrir sesiones, esperar para volver a medir o dejar una comprobación que el cierre lea como roja o en lugar de la de la tanda que mide. También MUST fallar si un disparo posterior a la tanda terminada no vuelve a medir.
- **FR-101**: `CHANGELOG.md` (*Unreleased*) MUST registrar `boe-legislacion` v0.1.4 (FR-027), la lista por clases, los dos umbrales nuevos, la eval nueva, el juicio sobre la respuesta a la pregunta, la tanda única y los errores de uso del sondeo.
- **FR-102**: El job de evals MUST ejecutarse en la propuesta de cambio (lo hace el workflow tras la revisión final), y su informe da lo que pide SC-001.

### Key Entities

- **Clase A**: lo que la respuesta dice del estado de la comprobación de la redacción o de una lectura o consulta anterior, o lo que anuncia que el agente tiene, necesita o va a hacer.
- **Clase B**: lo que la respuesta dice de una redacción que ninguna orden de la sesión devolvió: qué decía o qué cambió respecto de ella.
- **Forma fija de la skill**: texto que la skill enseña a escribir tal cual: la línea `⚠ REDACCIÓN MODIFICADA:` con su cita y sus dos fechas, la etiqueta de cada aviso de vigencia y la frase de la regla 7. Se quita de la respuesta antes de buscar.
- **Familia de la lista**: un grupo de expresiones con su clase; las de la clase B alimentan `redaccion_no_leida`.
- **Respuesta a la pregunta**: el resultado del turno que abre la pregunta en una sesión.
- **Umbral del informe**: un elemento de `umbrales` con el contrato del ADR 0029.
- **Línea esperada de la eval**: un bloque y sus dos fechas, que la respuesta tiene que llevar en una línea `⚠ REDACCIÓN MODIFICADA:`.

## Success Criteria *(mandatory)*

### Measurable Outcomes

- **SC-001**: En el informe del job de cierre de `boe-legislacion`, `umbrales` lleva:
  - `expresiones_prohibidas:<modelo>` con `decide: true` y `cumple: true`: como mucho 2 de 54, frente a 28 de 51 en la ejecución del ADR 0031 y 9 de 51 en el cierre de H7.3, contadas con las dos clases;
  - `sin_activar:<modelo>` con `decide: true` y `cumple: true`: 0, frente a 2 de 51 en la ejecución del ADR 0031;
  - `redaccion_no_leida:<modelo>` con `decide: true` y `cumple: true`: 0;
  - `duracion_de_las_sesiones` con `decide: true` y `cumple: true`: ≤ 900 s.

  Además: la eval 04 pasa, con su tasa (1 de 3 en la ejecución del ADR 0031); el job de `legal-core` sigue aprobado; la eval nueva declara sus dos líneas `⚠ REDACCIÓN MODIFICADA:` y su tasa; ninguna sesión se juzga por una réplica posterior a su respuesta ni queda sin medir por un límite; cada commit medido tiene una sola tanda de sesiones por skill; el veredicto es aprobado (el hito cambia `SKILL.md`, Definition of Done §1.10) y `red` está vacío; y el informe final da los cuatro umbrales como «comprobado por su control».

  **Control**: el job de evals de `boe-legislacion` sale en rojo, con el veredicto `fallo`, en cualquiera de estos casos: pasan más de 2 de 54 respuestas con alguna expresión (FR-040); alguna respuesta sin la skill activada (FR-041) o con la clase B (FR-042); la duración pasa de 900 s (FR-043); alguna sesión queda sin medir (H7.3 FR 043); o una serie que decide no pasa, como la 04. El cierre del workflow cuenta ese rojo. Dos cosas se comprueban en `make ci`, no en la medición: que la respuesta juzgada es la de la pregunta (SC-008) y que no hay dos tandas por commit (SC-010).
- **SC-002**: Sobre los informes versionados, la lista marca exactamente 36 de 93 respuestas en H7.1, 11 de 93 en H7.2 y 9 de 93 en H7.3, con el reparto por eval de FR-032, y ninguna otra; ninguna que hoy marca deja de marcarse. **Control**: `TestEvalsDelRepositorio` falla en `make ci` (FR-093).
- **SC-003**: 0 expresiones de la lista en el texto de los bloques grabados para las evals de `boe-legislacion`, y 0 expresiones en una respuesta hecha solo de las formas de la skill, una vez quitadas sus formas fijas. **Control**: `TestEvalsDelRepositorio` falla en `make ci` (FR-093).
- **SC-004**: 0 expresiones de la lista en la prosa de `SKILL.md` fuera de las órdenes, las formas fijas y la región generada, y 0 fechas `AAAAMMDD` escritas con cifras. **Control**: la comprobación de la prosa falla en `make ci` (FR-093).
- **SC-005**: Las 2 órdenes de lectura y comprobación de `SKILL.md` (la de `articulo` y la de `articulos`), 2 de 2, llevan su forma para PowerShell con la misma norma y los mismos bloques. **Control**: la comprobación de FR-095 falla en `make ci`.
- **SC-006**: Los tests del informe distinguen los casos de FR-097: `sin_activar` con 1 y con 0; `redaccion_no_leida` con 1 y con 0; 3 de 54 y 2 de 54; y 3 de 54 con seis sesiones sin terminar, que no se cumple. **Control**: fallan en `make ci`.
- **SC-007**: Los tests de `Juzgar` marcan en su clase las frases de FR-034 (A y B), y 0 marcas en las formas fijas, en «No hay avisos de vigencia sobre este bloque» y en la respuesta de FR-022 a quien quiere saber qué cambió; la sesión de `legal-core` que activa `boe-legislacion` no pasa. **Control**: fallan en `make ci` (FR-094).
- **SC-008**: Con la réplica a un aviso de una tarea en segundo plano después de la respuesta, la respuesta juzgada es la de la pregunta. **Control**: `TestLeerSesion` falla en `make ci` (FR-098).
- **SC-009**: Sin credencial o con un argumento que no vale, la salida de error del sondeo tiene 0 líneas de `go test` o de testify, y el sondeo abre 0 sesiones. **Control**: `TestGuionDelSondeo` falla en `make ci` (FR-099).
- **SC-010**: La definición del job permite como mucho 1 tanda de sesiones a la vez por commit y skill, y 0 comprobaciones que el cierre lea como rojas o en lugar de la de la tanda que mide; un disparo posterior a una tanda terminada vuelve a medir. **Control**: `TestDefinicionDelJob` falla en `make ci` (FR-100).
- **SC-011**: `make ci` en verde, con `skills-check` y `schema-check` sin drift y las reglas del conjunto; `SKILL.md` de `boe-legislacion` por debajo de 300 líneas; y la entrada de `CHANGELOG.md` (*Unreleased*) con v0.1.4. **Control**: `make ci` (`skills-check` falla con 300 líneas o más, o con la `description` por encima de 1024 caracteres).
- **SC-012**: La derivada de `da-3` sale de su respuesta grabada quitándole solo la redacción posterior. **Control**: `TestGrabacionesDerivadas` falla en `make ci` (FR-096).

## Uso, de fuera adentro

Cada salida del hito, desde quien la consume (criterio de uso, ADR 0028). El hito no cambia el binario, así que lo que la skill lee de la comprobación sigue acotado como en H7.1: por norma y bloques, ≤ k señales con k bloques leídos y como mucho 50 hallazgos. Volumen de referencia: meses de uso diario, con cientos de normas y miles de bloques consultados, la mayoría hace más de una semana. Nada de lo que entrega el job o el sondeo depende de ese volumen, porque cada sesión empieza de un estado preparado.

| Salida | Quién la pide, cuántas veces y qué hace con ella | Tamaño | Cuándo deja de darse cada señal |
|---|---|---|---|
| La respuesta de `boe-legislacion` | La persona que pregunta; una por pregunta. Lee el texto citado, sus avisos y, si cambió, una línea `⚠ REDACCIÓN MODIFICADA:` por precepto (FR-010, FR-011, FR-023). | Lo que ocupa la norma, y 0 líneas sobre la comprobación. Hay una línea de ≈ 200 bytes (≈ 160 de la forma de H7.3 y ≈ 40 de la cita) por bloque cuya redacción cambió: 0 en la mayoría de las preguntas, 2 en la de la eval nueva y como mucho k × 200 con k bloques leídos. Con cientos de normas y miles de bloques consultados, es la misma: no cuenta lo acumulado. | La línea sale en la respuesta de la lectura que ve la redacción nueva, y no en la siguiente lectura de ese bloque (H7.1). Al cabo de un mes, solo sale si el BOE publica otra redacción. |
| La salida de la comprobación (`graph check <norma> <bloques>…`), sin cambios en el binario | La skill, una vez por cada orden que lee bloques, detrás de la lectura (FR-025). En la eval nueva, una o dos veces. | ≈ 325 bytes sin cambios y ≈ 1 000 por bloque cambiado (medido en la revisión final de H7.3): ≈ 2 KB en la pregunta de la eval nueva, en una orden o en dos; ≤ 3 800 bytes con cinco bloques cambiados (H7.1). Acotada por los bloques pedidos, no por los miles consultados. | Cada señal se da una vez y la apaga la siguiente lectura del bloque (H7.1). |
| La `description` de `SKILL.md` | El agente, en cada sesión que tiene la skill instalada, para decidir si la activa (FR-001, FR-003). | ≤ 1024 caracteres (FR-026). | No da señales. |
| El cuerpo de `SKILL.md` v0.1.4, con la orden para PowerShell | El modelo, una vez por conversación en que se activa la skill; en Windows con PowerShell, la orden de FR-024 en cada lectura. | < 300 líneas (FR-026). | No da señales. |
| `umbrales` del informe de `boe-legislacion` | El job, que decide con ellos el veredicto (FR-046); el informe final del workflow, que los lee sin modelo (ADR 0029); y la persona. Una vez por job y skill. | 5 elementos de ≈ 250 bytes (tres que deciden del modelo que decide, el del informativo y la duración); `[]` en `legal-core`. Es un tamaño fijo, que no crece con el uso. | Cada job los mide de nuevo sobre su commit. Un umbral incumplido deja de darse en el primer job que lo cumple. |
| Las líneas esperadas de la eval nueva en el informe | La persona que lee el informe; una vez por job. | 2 líneas esperadas por sesión, con las encontradas y las ausentes, en 3 sesiones: < 1 KB. | Cada job las mide de nuevo. |
| La salida de error del sondeo en un error de uso | Quien lanza el sondeo; una vez por lanzamiento que falla. | Una línea por error, como mucho una por argumento (5) y la de la credencial: < 1 KB. | Deja de darse al corregir el argumento o exportar la credencial. |
| La lista con las dos clases | El job, al juzgar cada sesión de las evals que activan la skill; el sondeo; y el calibrado y las comprobaciones de `make ci`. | Fija, del orden de decenas de expresiones; no crece con el uso. | No da señales. |

## Fuera de alcance

Del hito, literal:

- «cambiar el modelo que decide o reabrir el ADR 0031, cambiar las repeticiones o la regla por serie (ADR 0016), o que decida el umbral del modelo informativo;»
- «retocar la eval 04 o su pregunta para que la skill se active;»
- «cambiar el binario de `kitlegal`: la salida de `graph check` —también su sobre sin hallazgos, que la revisión final de H7.3 señala como origen de las palabras—, `boe articulo` y `--describe`; y leer o comparar la redacción superada, que es `boe articulo --fecha` (H20);»
- «juzgar la redacción libre con un modelo o por similitud (H5.1, *Decisión del mecanismo*): las clases se miden por sus formas;»
- «promover a decisoria ninguna eval informativa; extender la lista o los umbrales a `legal-core`;»
- «medir en Windows: ningún job ni test ejecuta PowerShell, y lo que se comprueba es la forma de `SKILL.md`;»
- «la release, H20, la cobertura de Codecov (pendiente del ADR 0029) y editar el spec, el plan, los guiones o las grabaciones de H7 a H7.3.»

De lo que el hito no especifica (constitución, «Criterio de decisión autónoma», punto 2):

- Cambiar el workflow `hito`, `scripts/workflow/` (también `cierre.sh`) o el informe final. Si la tanda única no se consigue en el job, queda como supuesto y va en su propia rama (FR-072).
- `sin_activar` o `redaccion_no_leida` del modelo informativo, publicados o que decidan.
- Que el sondeo publique `sin_activar`, `redaccion_no_leida` o recuentos por clase: su salida es la de H7.3, con el recuento de FR-082.
- Cambiar `legal-core`: su `SKILL.md`, su versión o las preguntas, comandos y juicio de sus evals, más allá de declarar en cada una que `boe-legislacion` no se activa (FR-004). Lo que evite que `boe-legislacion` se active en ellas va a su `description` y a su cuerpo (FR-002, FR-003).
- Cambiar la eval 19: su pregunta, sus comandos o la forma de su hallazgo esperado. Conserva su juicio (FR-053).
- Declarar líneas `⚠ REDACCIÓN MODIFICADA:` esperadas en otras evals que la nueva.
- Cambiar la cita, los avisos de vigencia o la frase fija de la regla 7 (FR-025).
- Una eval para PowerShell o una comprobación de la orden para PowerShell que la ejecute.
- Que una sesión sin terminar deje de contar como no pasada en su serie, o que se reintente (FR-061, FR-062).
- Otras formas de dar la respuesta a la pregunta que el resultado del turno que la abre (FR-060).
- Un ADR nuevo.

## Assumptions

- La evidencia es la de los tres informes versionados (H7.1, H7.2 y H7.3, 93 respuestas cada uno) y los de las mediciones de `6ab3add` y `eb6b4c8`, que están en commits de la rama de H7.3 fuera de `main`, alcanzables con git: el de `6ab3add`, en `eb6b4c8:specs/013-h7-3-el-umbral-de/gates/evals/boe-legislacion.json`, y el de `eb6b4c8`, en `196ee05:` con esa misma ruta (cada informe lleva en `commit` el que midió). También cuentan la bitácora `docs/USO.md` del 2026-09-30 y el ADR 0031. El informe de la ejecución 36712391835 se lee del registro del job, como lo lee el cierre (`recoger_evals`). Lleva la respuesta, la activación y las invocaciones de cada sesión, pero no los transcripts. Del sondeo del ADR 0031 queda solo lo que registra el ADR, porque el sondeo borra su directorio (H7.3 FR 064). Si el registro del job no se pudiera leer, el research lo dice y trabaja con el ADR y la bitácora.
- Las cifras de la línea de base (28 de 51 de la clase A con el modelo que decide ahora; 7, 11 y 4 en el cierre de H7.3) son de la lectura de la bitácora y del hito. El calibrado exigible (FR-032) es sobre los tres informes versionados.
- Los nombres `sin_activar:<modelo>` y `redaccion_no_leida:<modelo>` son los del hito y cumplen el patrón del contrato del ADR 0029 (`^[a-z0-9_.:-]+$`). El hito deja al plan fijar otro nombre para el primero. Si lo hace, sustituye al de este spec en FR-041, FR-092 y SC-001.
- «La respuesta a la pregunta» es el resultado del turno que abre la pregunta. Si el agente cierra ese turno sin responder, esperando una tarea en segundo plano, lo que se juzga es ese resultado. Es la lectura del hito («lo que escribe después, como réplica al aviso de una tarea en segundo plano, no la sustituye»). Juzgar el último resultado es lo que se sustituye.
- La línea `⚠ REDACCIÓN MODIFICADA:` lleva la cita de su bloque también con un solo bloque cambiado, porque la Entrega (1) pide «una línea por precepto cambiado que dice cuál es» sin condición. Así la eval 19, que exige la forma de la etiqueta, sigue igual.
- El hito pide dos cosas que chocan al pie de la letra. Por un lado, la lista gana formas de «consulta anterior», que están en la frase de la regla 7. Por otro, ninguna forma nueva debe aparecer en las formas de la skill. Se concilian con lo que el propio hito dispone: antes de buscar se quitan las formas fijas (FR-031). La comprobación de las formas de H7.3 pasa a ser que una respuesta hecha solo de ellas no se marca (FR-033).
- Las respuestas grabadas de `a1-30` y `da-3` están bajo `internal/source/boe/testdata/boe.legislacion-consolidada/` desde H4, con las dos redacciones de cada bloque. `testdata/evals/grafo-previo/` ya tiene la derivada de `a1-30` (eval 19).
- En Windows nativo, Claude Code usa `pwsh.exe` (PowerShell 7) si está y, si no, `powershell.exe` (5.1) (<https://code.claude.com/docs/en/tools-reference#powershell-tool>, <https://code.claude.com/docs/en/setup>). No se mide en Windows.
- Los nombres técnicos que aparecen (`Juzgar`, `TestEvalsDelRepositorio`, `TestLeerSesion`, `TestGuionDelSondeo`, `TestDefinicionDelJob`, `TestGrabacionesDerivadas`, `leerResult`, `recontarExpresiones`, `schemas/expresiones-prohibidas.yaml.json`, `.github/workflows/evals.yml`, `scripts/evals-sondeo.sh`) existen en el repositorio o los fija el hito. Quedan para el plan: las familias y expresiones de la lista, la forma de la línea `⚠ REDACCIÓN MODIFICADA:` con su cita, la forma para PowerShell, cómo declara el formato de eval la no activación de otra skill y las líneas esperadas, cómo se reconoce en el transcript el turno de la pregunta y cómo evita el job la segunda tanda.
