# Specification Quality Checklist: H1 · Kernel CLI: multicall, flags globales, exit codes, sobre de salida

**Purpose**: Validar la completitud y calidad del spec antes de pasar a planificación
**Created**: 2026-09-11
**Feature**: [spec.md](../spec.md)

## Content Quality

- [x] No implementation details (languages, frameworks, APIs)
- [x] Focused on user value and business needs
- [x] Written for non-technical stakeholders
- [x] All mandatory sections completed

## Requirement Completeness

- [x] No [NEEDS CLARIFICATION] markers remain
- [x] Requirements are testable and unambiguous
- [x] Success criteria are measurable
- [x] Success criteria are technology-agnostic (no implementation details)
- [x] All acceptance scenarios are defined
- [x] Edge cases are identified
- [x] Scope is clearly bounded
- [x] Dependencies and assumptions identified

## Feature Readiness

- [x] All functional requirements have clear acceptance criteria
- [x] User scenarios cover primary flows
- [x] Feature meets measurable outcomes defined in Success Criteria
- [x] No implementation details leak into specification

## Notas de validación

- **«No implementation details» / «technology-agnostic»**. Igual que en H0, el reparto es deliberado:
  - **Los nombres de paquete (`internal/cli`, `internal/app`, `internal/core/schema`, `internal/render`), las banderas, las claves del sobre y los códigos de salida no son decisiones de diseño de este spec**: son el objeto literal del hito, fijados por `docs/ROADMAP.md` («Alcance» y «Aceptación» de H1), por las decisiones cerradas de `CLAUDE.md` y por la constitución (§IV y §VI). El criterio de aceptación del hito se enuncia sobre ellos (`cobertura internal/cli ≥ 90 %`), de modo que omitirlos haría el spec no trazable.
  - **Ningún requisito prescribe la forma del código.** FR-011 exige que la huella sea reproducible y dependa solo del contenido, no cómo se canonicaliza; FR-029 a FR-032 exigen clases de error, traducción única y exhaustiva y preservación de la clase al envolver, no un mecanismo; FR-041 exige dos formas de presentación, no su estructura interna; FR-048 exige que el esquema se derive de la definición del applet, no con qué herramienta.
  - **El *cómo* se ha empujado a *Assumptions* y al plan**: forma canónica de `data`, valor por omisión de `--timeout`, código concreto para el fallo inesperado de FR-031, columnas de la tabla mínima. Un cambio en cualquiera de ellos no obliga a tocar ningún FR.
  - Las herramientas nombradas en *Assumptions* (`kong`, `testify`, `go-internal`, `invopop/jsonschema`, `santhosh-tekuri/jsonschema`) están fijadas por la constitución §V; el spec las registra para acotar FR-060, no las elige.
- **«Written for non-technical stakeholders»**: las seis historias de usuario y los **quince** criterios de éxito (SC-001 … SC-015, tras la ronda de corrección del gate `spec` que añadió SC-014 y SC-015) describen resultados observables (una respuesta siempre citable, un binario que responde a varios nombres, fallos distinguibles sin leer el mensaje, un juego único de banderas, un applet que se describe solo, reglas que se hacen cumplir solas). El detalle técnico está confinado a los requisitos funcionales, cuyo destinatario es quien planifica.
- **[NEEDS CLARIFICATION] resueltos (3, el máximo permitido) en `/speckit-clarify` (modo desatendido, sesión 2026-09-11)**. El spec se redactó en modo desatendido con tres cuestiones de alcance o contrato sin respuesta determinada por el hito, `CLAUDE.md`, `refs/` o la constitución; las cinco preguntas evaluadas (incluidas dos adicionales sobre `--dry-run` y `--json --help` que no llevaban marcador explícito pero sí ambigüedad material) están en `gates/clarify-preguntas.json`, sus respuestas en `gates/clarify-respuestas.json`, y quedan integradas en `## Clarifications`:
  1. **FR-009 — el binario distribuido NO registra `echo`.** El «Alcance» del hito decía que el applet de ejemplo «vive solo en tests (`internal/app/testdata`)», y la «Entrega» exige que `kitlegal echo hola --json` y `ln -s kitlegal echo && ./echo hola` funcionen; la entrega literal se demuestra contra el binario que compila el test e2e, no contra el binario publicado. Ese acotamiento está en FR-009, FR-055 y también en SC-001, SC-003, SC-007, US1 escenario 1, US2 escenario 1 y US5 escenario 1, para que ningún criterio leído por separado prometa la entrega sobre el artefacto de release.
  2. **FR-016 — `fuente` y `url` en un applet sin fuente externa nunca van vacías.** Usan el espacio de nombres reservado `kitlegal.`/`kitlegal:` (p. ej. `kitlegal.echo` / `kitlegal:applet/echo`), inaccesible a los adaptadores de `internal/source/*`.
  3. **FR-045 — representación de un fallo bajo `--json`.** La salida estándar lleva el sobre con `ok: false` y la causa dentro de `data`; lo emite el kernel desde el único punto que traduce error a código de salida. Las cinco reglas de la respuesta están recogidas en el cuerpo del spec: quién lo emite y coherencia de `ok` con el código (FR-045), procedencia no vacía del sobre de fallo (FR-045, apoyado en FR-016 y FR-017), mensaje también en la salida de error (FR-045, US3 escenario 7) y esquema de salida condicionado a `ok` (FR-047).
- **Ítem «No [NEEDS CLARIFICATION] markers remain»**: ya se marca. Los tres marcadores del spec en modo desatendido se resolvieron en la sesión de clarificación y están integrados en FR-009, FR-016 y FR-045 y en la sección `## Clarifications`.
- **Trazabilidad de la aceptación literal del hito**: «cobertura `internal/cli` ≥ 90 %» → SC-004 y FR-056; «el e2e demuestra que stdout solo contiene JSON cuando `--json`» → SC-002, US1 escenario 2 y FR-042. Trazabilidad de los controles literales: «unit (mapeo error→exit, sobre, dispatch, `--describe`)» → FR-054, SC-005 y SC-006; «e2e testscript (`--help`, symlink, exit 2 con args malos, salida JSON parseable)» → FR-055 y SC-009; «test de arquitectura; `depguard` activo» → FR-050 a FR-053 y SC-008. La transcripción literal está en la sección *Criterios del hito, literales* del spec.
- **Trazabilidad de la entrega literal del hito**: el sobre de `kitlegal echo hola --json` → US1 escenario 1 y SC-001; el enlace simbólico → US2 escenario 1 y SC-003; `--describe` → US5 escenario 1 y SC-007. Los tres se verifican contra el binario que compila el e2e (FR-009).
- **Trazabilidad del contrato de fallo** (añadida tras la revisión del gate `spec`): FR-045 → US3 escenarios 9, 10 y 11, SC-014 y FR-054; procedencia del sobre de fallo (FR-045 + FR-016 + FR-017) → US3 escenario 10 y SC-015; esquema condicionado a `ok` (FR-047) → US5 escenarios 2 y 5 y SC-015.
- **Ítem «Requirements are testable and unambiguous»**: los tres puntos donde la ambigüedad era material se resolvieron en clarificación (FR-009, FR-016, FR-045), junto con dos ambigüedades adicionales sin marcador explícito pero con el mismo impacto (`--dry-run --json` en FR-022; `--json --help` en FR-042/FR-028 y en *Edge Cases*). Tras la revisión del gate `spec`, lo que la clarificación decidió está ahora **completo en el cuerpo del spec**, no solo en `## Clarifications`:
  - FR-045 fija la procedencia del sobre de fallo (la de la fuente consultada cuando se conoce; en su defecto el espacio de nombres reservado del kernel, `kitlegal.cli` / `kitlegal:cli`), de modo que los fallos anteriores al applet (FR-006, FR-027) cumplen FR-016 y FR-017 y su sobre es validable.
  - FR-047 describe `data` condicionado a `ok`, de modo que el esquema de `--describe` cubre también la salida fallida que FR-045 obliga a emitir.
  - FR-022 recoge que la descripción de `--dry-run` va a la salida de error y es siempre visible (no depende de `--verbose` ni de `KITLEGAL_LOG`) y que el kernel no corta antes del applet: la bandera viaja en el Contexto de ejecución, igual que en FR-021, FR-023 y FR-024.
  - SC-001, SC-003 y SC-007 acotan la entrega literal al binario que compila el e2e (FR-009), como ya hacía FR-055; sobre el binario distribuido, `echo` no está registrado.
  El resto de decisiones abiertas son de diseño, no de contrato, y están listadas en *Assumptions* como delegadas al plan; ninguna cambia el enunciado de un requisito.
- **Ítem «All functional requirements have clear acceptance criteria»**: cada requisito tiene escenario o criterio que lo comprueba. En particular, el contrato de fallo ya no queda sin verificar: US3 escenarios 9, 10 y 11 comprueban el sobre con `ok: false`, la procedencia reservada del kernel en los fallos previos al applet y la correspondencia entre clase y código de salida; SC-014 lo exige para las cinco clases de error y para el fallo inesperado de FR-031; SC-015 y US5 escenario 5 comprueban que ese sobre valida contra la descripción formal (FR-017) y contra el esquema de `--describe` (FR-047); FR-054 incluye el sobre de fallo entre los tests unitarios. El comportamiento de `--dry-run` de FR-022 se comprueba en US4 escenarios 3 y 7.
- **Ítem «Scope is clearly bounded»**: la sección *Fuera de alcance* es obligatoria en este proyecto y enumera, hito a hito, lo que H1 no construye, con atención especial a las tres banderas (`--offline`, `--no-graph`, `--asunto`) cuyo objeto no existe todavía: se declaran y se propagan, y su semántica llega en H3, H12 y H14.
- **Ronda de corrección (gate `spec` rechazado, 2026-09-11)**: se integraron en el cuerpo del spec las reglas de las respuestas Q3 y Q4 que solo figuraban en `## Clarifications` (FR-045, FR-047, FR-022), se añadieron los escenarios y criterios que faltaban para el sobre de fallo (US3 9–11, US5 5, SC-014, SC-015), se acotó al binario del e2e la entrega literal en SC-001, SC-003, SC-007 y en los escenarios correspondientes, y se corrigió en *Fuera de alcance* el recuento de applets de H11 (regla de los tres applets: `boe`, `cita`, `plazos`; `echo` no cuenta). Ninguna corrección requirió decisión humana, por lo que no hay `gates/spec-pendiente.md`.
- **Ronda de corrección (gate `tasks` rechazado, 2026-09-11)**: el veredicto alcanzó sobre todo a `tasks.md` y `quickstart.md`, pero dos de sus motivos tenían su raíz en el spec y se corrigieron aquí:
  - **FR-051 y SC-008** afirmaban que **las cinco** reglas de dependencia se comprueban «tanto en el lint como en el test de arquitectura». No es cierto ni puede serlo: las dos reglas de símbolo —terminar el proceso, escribir en la salida estándar— no son observables en un grafo de dependencias, porque el paquete que las contiene lo importa todo el mundo por otros motivos. Ahora ambos enunciados distinguen las **tres reglas de importación** (dos capas) de las **dos de símbolo** (lint con análisis de tipos), como ya hacía `contracts/reglas-de-arquitectura.md` §2. No se rebaja ninguna garantía: las cinco siguen fallando y nombrando la regla; lo que se corrige es la afirmación sobre **dónde** falla cada una.
  - **El recuento de criterios de éxito** de esta checklist decía «trece» y eran quince: residuo de la redacción anterior a la ronda del gate `spec`.
  Ningún motivo requirió decisión humana, por lo que no hay `gates/tasks-pendiente.md`.
- Con los tres marcadores resueltos, sin `[NEEDS CLARIFICATION]` pendientes y aplicados los motivos del gate, el spec está listo para `/speckit-plan`.
