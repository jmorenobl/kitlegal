# Specification Quality Checklist: H7.4 · `boe-legislacion` sin el estado de la comprobación ni lo que no ha leído, y un job que juzga la respuesta a la pregunta

**Purpose**: Validate specification completeness and quality before proceeding to planning
**Created**: 2026-09-30
**Feature**: [spec.md](../spec.md)

## Content Quality

- [x] No implementation details (languages, frameworks, APIs) — el spec dice qué se mide y qué ve quien pregunta; la forma de la lista, de la línea `⚠ REDACCIÓN MODIFICADA:` con su cita, de la orden para PowerShell, del formato de eval y de la tanda única queda para el plan (Assumptions, última viñeta). Los nombres de ficheros, tests y umbrales que aparecen son los que fija el hito o existen en el repositorio.
- [x] Focused on user value and business needs — Resumen y US1 a US3: lo que lee quien pregunta; US4 a US7: el instrumento que lo mide.
- [x] Written for non-technical stakeholders — cada historia empieza por lo que ve una persona; el detalle técnico está en los requisitos, como en H7.3.
- [x] All mandatory sections completed — User Scenarios, Requirements, Success Criteria, Assumptions y «Fuera de alcance».

## Requirement Completeness

- [x] No [NEEDS CLARIFICATION] markers remain — ninguno; las dos lecturas que había que fijar (la respuesta a la pregunta y la cita en la línea con un solo bloque) salen del hito y están en Assumptions.
- [x] Requirements are testable and unambiguous — cada FR tiene con qué comprobarse, pero no todos lo dicen igual. Unos llevan «Comprobable» o remiten a un control de FR-091 a FR-102. La activación y los umbrales (FR-001, FR-003, FR-040 a FR-048) los comprueba SC-001 en el job de cierre. La lista, sus clases y las formas fijas (FR-010 a FR-012, FR-030 a FR-037) los comprueban SC-002, SC-003 y SC-007. La eval nueva (FR-050 a FR-055) la comprueban SC-001 y SC-012. FR-036 y FR-082 no llevan ninguna de las dos cosas: se comprueban con los tests de `Juzgar`, del informe y de `TestLeerSesion` (FR-094, FR-097, FR-098), porque el job y el sondeo comparten el juicio (FR-060). FR-072 tampoco las lleva: se comprueba en la revisión, que ve si el diff toca `scripts/workflow/`. FR-012 ya no choca con FR-010 ni con FR-034: solo excluye la frase que dice únicamente que no hay avisos o que la norma no está derogada.
- [x] Success criteria are measurable — SC-001 a SC-012 con recuentos, proporciones o segundos.
- [x] Success criteria are technology-agnostic (no implementation details) — miden respuestas, umbrales, tandas y salidas; los controles nombran la comprobación que falla, como exige el ADR 0029.
- [x] All acceptance scenarios are defined — US1 a US7, en forma Dado/Cuando/Entonces.
- [x] Edge cases are identified — «Edge Cases»: 3 de 54 frente a 2 de 54, total 0, sesiones sin terminar, una orden o dos, formas fijas junto a texto libre, «no hay avisos» en la misma frase que el estado de la comprobación, prueba de red, réplica a una tarea en segundo plano.
- [x] Scope is clearly bounded — «Fuera de alcance», literal del hito y lo no especificado.
- [x] Dependencies and assumptions identified — «Assumptions» y «Relación con H7.3, H7.2, H7.1 y H5».

## Feature Readiness

- [x] All functional requirements have clear acceptance criteria — trazabilidad de «Criterios del hito, literales» y controles FR-091 a FR-102.
- [x] User scenarios cover primary flows — la respuesta (US1), la activación (US2), dos bloques y PowerShell (US3), los umbrales (US4), el juicio (US5), la tanda (US6) y el sondeo (US7).
- [x] Feature meets measurable outcomes defined in Success Criteria — SC-001 recoge la Aceptación literal; SC-002 a SC-012, los Controles.
- [x] No implementation details leak into specification — ver el primer ítem.

## Criterios del juez del spec (workflow `hito`, `juez_spec`)

- [x] a. Alcance: cubre Entrega (1) a (8) y nada más (tabla de trazabilidad).
- [x] b. Fuera de alcance: las siete viñetas literales del hito y, de lo no especificado, también cambiar `legal-core` más allá de declarar en sus evals que `boe-legislacion` no se activa (FR-004).
- [x] c. Clarificaciones: `gates/clarify-preguntas.json` y `gates/clarify-respuestas.json` existen y están vacíos, así que no hay respuestas que reflejar. Las lecturas que fija el spec están en Assumptions.
- [x] d. Aceptación: «Criterios del hito, literales» con Aceptación y Controles transcritos; SC-001 a SC-012.
- [x] e. Decisiones cerradas: ni modelo, ni repeticiones, ni regla por serie, ni binario (FR-047, «Fuera de alcance»).
- [x] f. Checklist veraz: cada marca cita dónde se cumple.
- [x] g. Sin implementación: las formas concretas son del plan.
- [x] h. Derivable: FR-012 excluye solo la frase que dice únicamente que no hay avisos o que la norma no está derogada, y lo que la misma frase dice además del estado de la comprobación es clase A, como en 01-02 y 14-01 (FR-034, Edge Cases); umbrales con medida y total (FR-040 a FR-045), líneas esperadas con fechas (FR-052), frases concretas (FR-034) y códigos de salida (FR-080).
- [x] i. Uso: «Uso, de fuera adentro», con quién pide cada salida, su tamaño y cuándo deja de darse cada señal.
- [x] j. Umbral con control: SC-001 nombra el job en rojo por cada umbral (FR-040 a FR-043, FR-046); SC-002 a SC-012 nombran su test en `make ci`; FR-092 fija las cuatro filas de «Controles de umbral».

## Notes

- La conciliación de «consulta anterior» en la lista con la frase de la regla 7 (Assumptions) se resuelve con lo que dispone el propio hito: quitar las formas fijas antes de buscar (FR-031, FR-033).
