# Specification Quality Checklist: H7.3 · El umbral de expresiones prohibidas decide, `boe-legislacion` sin el vocabulario que provoca el ruido, y evals en paralelo con un sondeo local

**Purpose**: Validate specification completeness and quality before proceeding to planning
**Created**: 2026-09-29
**Feature**: [spec.md](../spec.md)

## Content Quality

- [x] No implementation details (languages, frameworks, APIs) — el spec nombra solo lo que fija el hito o existe en el repositorio (`Juzgar`, `TestPrepararSesion`, `scripts/evals.sh`, los eventos de `stream-json` que el hito cita); cómo se aísla cada sesión, cómo evita el flujo la segunda tanda, las expresiones exactas de la tercera familia y los argumentos del sondeo quedan para el plan (Assumptions, última viñeta; FR-020, FR-034, FR-060).
- [x] Focused on user value and business needs — US1 (la respuesta sin el estado de la comprobación), US2 (el umbral lo hace cumplir el job), US3 (medir en minutos) y US4 (iterar en un Mac), trazados al Objetivo del hito.
- [x] Written for non-technical stakeholders — cada historia dice qué gana quien pregunta o quien ajusta la skill antes que el mecanismo; el vocabulario técnico es el del contrato del ADR 0029, que el hito exige.
- [x] All mandatory sections completed — User Scenarios & Testing, Requirements, Success Criteria, Uso de fuera adentro, Fuera de alcance y Assumptions.

## Requirement Completeness

- [x] No [NEEDS CLARIFICATION] markers remain — ninguno: las ambigüedades se cierran con el hito (p. ej., qué límite detiene el job: FR-044, el de uso de FR-040 a, que el hito describe como «el de la ventana, el semanal o el de la familia del modelo»).
- [x] Requirements are testable and unambiguous — cada FR lleva su comprobación (p. ej., FR-002/FR-003 con 3 y 2 de 51; FR-013 define qué es una fecha `AAAAMMDD`; FR-040 enumera los tres casos de límite; FR-061 compara eval a eval con el juicio del job); sin contradicciones entre requisitos: FR-032 exceptúa lo que cambia tras un límite de uso (FR-044), FR-065 publica las sesiones que no terminaron con su motivo (FR-063), y FR-034 dice qué deja el segundo disparo al cierre que lo consume.
- [x] Success criteria are measurable — SC-001 a SC-010 con recuentos o umbrales (≤ 2 de 51, ≤ 900 s, 35 y 10, ≥ 5 y ≤ 2 de 15, 0 expresiones, 0 fechas, < 300 líneas), y cada uno medido sin persona nombra el control que sale en rojo.
- [x] Success criteria are technology-agnostic (no implementation details) — se expresan como recuentos, duraciones y veredictos observables en el informe; los nombres de campos son los del contrato del ADR 0029 que exige el hito.
- [x] All acceptance scenarios are defined — cinco historias con escenarios Dado/Cuando/Entonces que cubren las seis piezas de la Entrega.
- [x] Edge cases are identified — «Edge Cases»: los bordes 3/51 y 2/51, la prueba de red, sesiones sin medir y el total, límites a mitad del job, un segundo disparo tardío, evals de no activación y límites en el sondeo, argumentos no válidos y paráfrasis.
- [x] Scope is clearly bounded — «Fuera de alcance» con las diez viñetas literales del hito y lo no especificado (workflow y `cierre.sh`, concurrencia adaptativa, reintentos de sesiones, otros límites, umbrales para `legal-core` o el sondeo).
- [x] Dependencies and assumptions identified — «Assumptions» y «Relación con H7.2 y H5» (lo que se sustituye y lo que se queda).

## Feature Readiness

- [x] All functional requirements have clear acceptance criteria — cada grupo de FR tiene su historia y su SC: FR-001 a FR-008 → US2 y SC-001/SC-006; FR-010 a FR-018 → US1 y SC-005/SC-010; FR-020 a FR-024 → US1 y SC-003/SC-004; FR-030 a FR-052 → US3 y SC-006 a SC-008; FR-060 a FR-068 → US4 y SC-009; FR-070 → US5 y SC-002.
- [x] User scenarios cover primary flows — la respuesta de la skill, el veredicto del job, el job en paralelo con límites y duración, el sondeo y el quickstart.
- [x] Feature meets measurable outcomes defined in Success Criteria — la Aceptación literal del hito es SC-001 y SC-002; los Controles literales son FR-090 a FR-099 y SC-003 a SC-010.
- [x] No implementation details leak into specification — ver el primer ítem; el plan decide mecanismos y los registra en «Decisiones».

## Rúbrica del juez del spec (autocomprobación)

- [x] a. alcance — las seis piezas de la Entrega, y nada más; ningún caso para estados manipulados a mano.
- [x] b. fuera_de_alcance — viñetas literales del hito y lo no especificado.
- [x] d. aceptación — «Criterios del hito, literales» copia Aceptación y Controles; SC-001 y SC-002 los recogen como condiciones con umbral.
- [x] e. decisiones cerradas — ADR 0016 (modelo que decide, repeticiones, regla por serie, Haiku informativo) y ADR 0029 (contrato) sin cambios; sin ADR nuevo (FR-080).
- [x] i. uso — tabla «Uso, de fuera adentro» con quién pide cada salida, cuántas veces, su tamaño y cuándo deja de darse cada señal.
- [x] j. umbral con control — FR-002/FR-003 y FR-051 (job en rojo), FR-043 (sesiones sin medir), y en `make ci` FR-091, FR-094, FR-095 y FR-096; el umbral de Haiku 4.5 es `decide: false` por el hito y no es umbral del hito (FR-004); SC-002 lo mide la persona.

## Notes

- Validado en una iteración contra `scripts/workflow/precheck.sh spec` y la rúbrica de `juez_spec`.
