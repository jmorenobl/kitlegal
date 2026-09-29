# Specification Quality Checklist: H7.2 · La consulta repetida con evidencia coherente, y respuestas sin la maquinaria interna

**Purpose**: Validate specification completeness and quality before proceeding to planning
**Created**: 2026-09-29
**Feature**: [spec.md](../spec.md)

## Content Quality

- [x] No implementation details (languages, frameworks, APIs) — el spec dice qué exige la eval, qué dice la respuesta y qué publica el informe; las expresiones exactas de la lista, dónde vive, su formato, los campos del informe y el nombre del fichero de la eval nueva son del plan (FR-050, Assumptions). Los nombres técnicos que aparecen (`Juzgar`, `TestGrabacionesDerivadas`, `grabacionesDerivadas()`, los ficheros que se retiran) los fija el hito (FR-020, FR-081 a FR-083).
- [x] Focused on user value and business needs — parte de lo que lee quien pregunta (US1, FR-040 a FR-044) y de la medida de la bitácora (Resumen, «Uso, de fuera adentro»).
- [x] Written for non-technical stakeholders — cada historia empieza por quien pregunta, por la eval o por la persona que repite la aceptación; lo técnico queda en los requisitos.
- [x] All mandatory sections completed — historias con escenarios Dado/Cuando/Entonces, casos límite, requisitos, entidades, criterios de éxito, «Uso, de fuera adentro», «Fuera de alcance» y supuestos.

## Requirement Completeness

- [x] No [NEEDS CLARIFICATION] markers remain — ninguno: el umbral del 5 % se lee como criterio de aceptación sobre el recuento del informe y no como regla nueva del veredicto (FR-054, SC-001, «Fuera de alcance»), la lectura que menos comportamiento añade; la pregunta de la eval es el ejemplo literal del hito (FR-001).
- [x] Requirements are testable and unambiguous — cada FR lleva su criterio comprobable o un escenario: la eval nueva con sus comandos, cita y forma (FR-001 a FR-003, US2.1), el `version-obsoleta` exacto de su grafo previo (FR-002), la derivada y su control (FR-010 a FR-012, SC-005), el juicio de la lista (FR-051 a FR-055, US3), el calibrado con su reparto por eval (FR-084, SC-003) y el quickstart con sus tres condiciones (FR-061). FR-041 y FR-043 no se contradicen: con otro código se dice que no se pudo comprobar, sin afirmar que cambió ni que no. El veredicto de una serie que no pasa o de una lista mal formada es el `fallo` que ya da el job (US3.7, US3.9, FR-055); y FR-060 enumera lo que las órdenes del quickstart no cambian fuera del directorio temporal (árbol, índice e historial de git, instalación de kitlegal, skills de la persona, conversaciones guardadas), sin promesa universal.
- [x] Success criteria are measurable — SC-001 (≤ 5 % por modelo: ≤ 2 de 51 y ≤ 1 de 30), SC-002 (fechas 20180309 y 20200206, 0 expresiones), SC-003 (35 de 93 y 0 de 58), SC-004 (0), SC-007 (0 ficheros, 19 evals, 10 que deciden), SC-008 (< 300 líneas).
- [x] Success criteria are technology-agnostic (no implementation details) — hablan de respuestas, recuentos, sesiones y veredictos; SC-005 y SC-006 nombran los controles que el hito nombra.
- [x] All acceptance scenarios are defined — US1 a US5, con los criterios literales de «Aceptación» y «Controles» transcritos y trazados (SC-001, SC-002, FR-080 a FR-087).
- [x] Edge cases are identified — pregunta que dice «te pregunté» sin lectura previa, dos normas, transcripción de un artículo, formas de los avisos, paráfrasis, fallos de `graph check` o `kitlegal boe` y lista mal escrita; solo estados a los que el producto llega.
- [x] Scope is clearly bounded — «Fuera de alcance» con lo literal del hito y lo que el hito no especifica; «Relación con H7 y H7.1» dice qué se sustituye y qué se queda (FR-070).
- [x] Dependencies and assumptions identified — informes de H7 y H7.1, redacciones de la grabada de H4, modelos y repeticiones del ADR 0016, preparación del quickstart (Assumptions).

## Feature Readiness

- [x] All functional requirements have clear acceptance criteria — trazados en la tabla de «Criterios del hito» y en cada historia.
- [x] User scenarios cover primary flows — la respuesta sin maquinaria (US1), la eval coherente (US2), la lista que decide (US3), lo dicho en otra conversación (US4) y el quickstart (US5).
- [x] Feature meets measurable outcomes defined in Success Criteria — SC-001 a SC-008 cubren la Aceptación y los Controles del hito.
- [x] No implementation details leak into specification — ver el primer punto.

## Criterio de uso (ADR 0028)

- [x] Cada salida dice quién la pide, cuántas veces, su tamaño con meses de uso y cuándo deja de darse cada señal — tabla «Uso, de fuera adentro»: la respuesta (0 líneas de maquinaria, ≤ k × 150 bytes de forma, sin crecer con lo acumulado; FR-040 a FR-044), el informe (≤ 1 KB por sesión, dos enteros por modelo; FR-053) y la lista fija; la forma se apaga con la lectura siguiente (H7.1 FR 024, SC-002).
- [x] Ninguna salida crece con todo lo acumulado sin cota — el binario no cambia (H7.1 SC 005); la respuesta no cuenta la memoria (FR-040, FR-041).

## Umbral de materialidad (constitución, «Gates»)

- [x] Lo que llega de fuera manipulado lo cubre la regla genérica, sin casos — la lista mal escrita es un fichero mal formado (FR-055), sin enumerar formas; ningún requisito trata estados a los que el producto no llega.

## Notes

- Los requisitos de H5, H7 y H7.1 se citan «H7 FR 085», sin guion, para que la comprobación mecánica de ids no los confunda con los de este spec; en la Entrega literal, el id de H7.1 va escrito así por la misma razón.
