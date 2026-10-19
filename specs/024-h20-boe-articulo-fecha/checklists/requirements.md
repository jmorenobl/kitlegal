# Specification Quality Checklist: H20 · `boe articulo --fecha`: la redacción vigente en una fecha + `boe-legislacion` v0.2

**Purpose**: Validate specification completeness and quality before proceeding to planning
**Created**: 2026-10-19
**Feature**: [spec.md](../spec.md)

## Content Quality

- [x] No implementation details (languages, frameworks, APIs) — el spec nombra solo lo que fija el hito o el contrato del repositorio (`--fecha`, `boe_articulo` con `fecha`, los campos de hoy de `data`, `umbrales`, `make ci`, y `articuloDelBloque` y `Consolidado.Lecturas`, que cita el hito); los nombres de los campos nuevos, la forma del mensaje de FR-003, la fila de FR-004, cómo se guarda y se emite la lectura fechada, las claves de las evals, el nombre del umbral y el texto de las formas fijas quedan para el plan (FR-003, FR-004, FR-010, FR-022, FR-030, FR-031, FR-048, FR-051, FR-052, FR-054; Assumptions, último punto).
- [x] Focused on user value and business needs — Resumen y US1 a US5: la redacción de la fecha del acto citada con su fecha, el aviso de que hoy rige otra y la fecha en que el precepto aún no regía.
- [x] Written for non-technical stakeholders — las historias y los escenarios hablan del despacho y de quien pregunta; lo técnico va en los requisitos, como en los specs anteriores del proyecto.
- [x] All mandatory sections completed — User Scenarios & Testing, Requirements, Success Criteria, Uso, Fuera de alcance y Assumptions.

## Requirement Completeness

- [x] No [NEEDS CLARIFICATION] markers remain — las dos del spec (FR-004 y FR-054) se resolvieron en el paso de clarify, junto con la redacción posterior de vigencia futura (FR-041) y el año o mes sin día (FR-045); sus respuestas están en «Clarifications».
- [x] Requirements are testable and unambiguous — cada FR lleva su «Comprobable» o su «Control», o es él mismo un control de la Definition of Done (FR-080 a FR-087, FR-089 a FR-091) o la nota del informe final (FR-092); las ambigüedades que cambiaban algo observable (FR-004, FR-054) se resolvieron en «Clarifications», y las del primer veredicto del spec, en el texto: FR-049 ya no restringe a la lectura sin fecha que ninguna lectura trae el fin de una redacción, y FR-041 prohíbe dar la vigencia de la posterior como ese fin; US1.2 pide `20200206` y `BOE-A-2020-1651` tal como los da el binario, como FR-041 y FR-054; FR-040 pide el día ante un hecho sin fecha y FR-046 es su complemento, con FR-045: rige cuando la pregunta no lleva una fecha, ni un año o un mes, ni sitúa el momento por un hecho, de modo que una fecha que es la de hoy sigue FR-040 y «qué decía antes» sin fecha sigue FR-046; SC-010 cuenta los diez casos de FR-089.
- [x] Success criteria are measurable — SC-001 y SC-004 a SC-011, cada uno con su cifra y la comprobación que sale en rojo (SC-010, 10 de 10: los nueve casos de sesiones sintéticas de FR-089 y el de los elementos de `umbrales`); SC-003 lo mide una persona y lo dice; SC-002 lo decide sin modelo el informe final del workflow, y SC-012, la revisión final por lectura.
- [x] Success criteria are technology-agnostic (no implementation details) — se miden por lo que da el job, `make ci` o una persona (SC-001 a SC-012), sin fijar cómo se construye.
- [x] All acceptance scenarios are defined — Dado/Cuando/Entonces en US1 a US5; la Aceptación literal del hito en SC-001 a SC-003.
- [x] Edge cases are identified — «Edge Cases»: el día de entrada en vigor, entre publicación y vigencia, anterior a la primera vigencia, posterior a la última, versión sin fecha o bloque sin versiones, fecha inválida, errores de la lectura de hoy, `--offline` sin nada guardado, norma hoy derogada, año o mes sin día, hecho sin fecha y la regla genérica.
- [x] Scope is clearly bounded — «Fuera de alcance», con el del hito literal y lo que el hito no especifica.
- [x] Dependencies and assumptions identified — «Relación con hitos anteriores» y «Assumptions».

## Feature Readiness

- [x] All functional requirements have clear acceptance criteria — cada FR remite a un golden, un e2e, un test de `make ci`, el job de evals o la revisión final (FR-080 a FR-091).
- [x] User scenarios cover primary flows — redacción de la fecha con posterior (US1), fecha anterior al bloque (US2), pregunta sin fecha (US3), grafo (US4) y umbral (US5).
- [x] Feature meets measurable outcomes defined in Success Criteria — trazabilidad en «Criterios del hito, literales».
- [x] No implementation details leak into specification — como en el primer punto de «Content Quality».

## Notes

- Items marked incomplete require spec updates before `/speckit-clarify` or `/speckit-plan`. Ninguno queda incompleto tras el paso de clarify.
