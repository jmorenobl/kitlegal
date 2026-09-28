# Specification Quality Checklist: H7.1 · `graph check` acotado a la pregunta, con señales que se apagan y salida legible; H7 sin lo que no pasa el umbral

**Purpose**: Validate specification completeness and quality before proceeding to planning
**Created**: 2026-09-28
**Feature**: [spec.md](../spec.md)

## Content Quality

- [x] No implementation details (languages, frameworks, APIs) — el spec dice qué observa quien usa el binario y la skill; cómo se pasan los argumentos de `check`, los nombres de campos de `data` y del formato de eval y cómo se guardan las lecturas son del plan (FR-001, FR-012, FR-054, Assumptions). Los nombres de ficheros que aparecen los fija el hito (FR-071, FR-072, FR-080, FR-095).
- [x] Focused on user value and business needs — el objetivo es lo que gana `boe-legislacion` (US1, FR-040 a FR-048) y lo que ve una persona (US3, US5), con la medida de la bitácora (Resumen, «Uso, de fuera adentro»).
- [x] Written for non-technical stakeholders — cada historia empieza por quien pregunta o repasa su memoria; lo técnico queda en los requisitos.
- [x] All mandatory sections completed — historias con escenarios Dado/Cuando/Entonces, casos límite, requisitos, entidades, criterios de éxito, «Fuera de alcance» y supuestos.

## Requirement Completeness

- [x] No [NEEDS CLARIFICATION] markers remain — ninguno: la única duda (si la salida de `check` lleva la forma fija) la cierra el criterio 2 de la constitución y va a «Fuera de alcance».
- [x] Requirements are testable and unambiguous — cada FR lleva su criterio comprobable o un escenario que lo prueba; la secuencia de cinco lecturas (FR-025, SC-003) y la cota (FR-010, FR-013, SC-001) tienen resultados exactos; FR-044 (sin decir nada con 0 y sin `version-obsoleta`) y FR-046 (decir que no se pudo comprobar con otro código) se reparten los códigos de `graph check` sin solaparse.
- [x] Success criteria are measurable — SC-001 (≤ 40 000 bytes, 50 hallazgos), SC-003 y SC-004 (recuentos exactos), SC-005 (≤ k hallazgos, ≤ 3 800 bytes), SC-009, SC-011 y SC-012 (códigos y recuentos).
- [x] Success criteria are technology-agnostic (no implementation details) — hablan de códigos de salida, hallazgos, bytes de salida y respuestas de la skill; SC-013 nombra los ficheros que el hito manda retirar.
- [x] All acceptance scenarios are defined — US1 a US6, con los criterios literales de «Aceptación» y «Controles» del hito transcritos y trazados.
- [x] Edge cases are identified — norma sin ELI, bloque repetido, lectura que no llega al grafo, fechas iguales o no válidas, concurrencia, más de 50 hallazgos, bloque desconocido, grafo ausente, `world.db` de H7 y la regla genérica; solo estados a los que el producto llega (umbral de materialidad).
- [x] Scope is clearly bounded — «Fuera de alcance» con lo del hito y lo que el hito no especifica; «Relación con H7» dice qué se sustituye, qué se retira y qué se queda.
- [x] Dependencies and assumptions identified — H7 sin release, vigencia de la caché y `--offline` (H3, H4), cifras de la medida, redacciones derivadas de H4 y SC-007 fuera del run.

## Feature Readiness

- [x] All functional requirements have clear acceptance criteria — trazados en la tabla de «Criterios del hito» y en cada historia.
- [x] User scenarios cover primary flows — la skill (US1), el apagado (US2), el acotado y la cota (US3), `fuente-caducada` (US4), la salida legible (US5) y la retirada (US6).
- [x] Feature meets measurable outcomes defined in Success Criteria — SC-001 a SC-014 cubren la Aceptación y los Controles del hito.
- [x] No implementation details leak into specification — ver el primer punto.

## Criterio de uso (ADR 0028)

- [x] Cada salida dice quién la pide, cuántas veces por pregunta, su tamaño con meses de uso y cuándo deja de darse cada señal — tabla «Uso, de fuera adentro», FR-013, FR-024, FR-031, SC-005.
- [x] Ninguna salida crece con todo lo acumulado sin cota — `check` con cota de 50 (FR-010); `stats` por pares; `show` acotado por una norma y sin consumidor skill.

## Umbral de materialidad (constitución, «Gates»)

- [x] Lo que llega de fuera manipulado lo cubre la regla genérica, sin casos — FR-070, FR-072, FR-073 y el último punto de «Fuera de alcance».

## Notes

- Los requisitos de H7 se citan «H7 FR 060», sin guion, para que la comprobación mecánica de ids no los confunda con los de este spec.
