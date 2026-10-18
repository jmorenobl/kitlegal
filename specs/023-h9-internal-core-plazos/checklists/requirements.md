# Specification Quality Checklist: H9 · `internal/core/plazos` + applet `plazos` + skill `legal-core` v1

**Purpose**: Validate specification completeness and quality before proceeding to planning
**Created**: 2026-10-18
**Feature**: [spec.md](../spec.md)

## Content Quality

- [x] No implementation details (languages, frameworks, APIs) — el spec nombra solo lo que fija el hito o el contrato del repositorio (el applet, su verbo y su herramienta, `data/festivos/`, `umbrales`, `make ci`); los argumentos, las claves de `data`, los códigos de aviso, las claves de eval, los nombres de los umbrales, el texto de las formas fijas y los métodos de agosto y del silencio quedan para el plan (FR-001, FR-009, FR-010, FR-048, FR-051, FR-053, FR-054; Assumptions, último punto).
- [x] Focused on user value and business needs — Resumen y US1 a US6: la fecha de un recurso calculada por el binario, con su precepto y sus razones, y la cobertura declarada en lo que la ley regula aparte.
- [x] Written for non-technical stakeholders — las historias y los escenarios hablan de la persona que recurre y del despacho; lo técnico va en los requisitos, como en los specs anteriores del proyecto.
- [x] All mandatory sections completed — User Scenarios & Testing, Requirements, Success Criteria, Uso, Fuera de alcance y Assumptions.

## Requirement Completeness

- [x] No [NEEDS CLARIFICATION] markers remain — los dos que había (entidad menor en FR-024; «da una fecha» en FR-054) están resueltos en `## Clarifications`, junto con la comprobación de los avisos (FR-041, FR-051).
- [x] Requirements are testable and unambiguous — cada FR lleva su «Comprobable» o su «Control», o es él mismo un control de la Definition of Done (FR-080 a FR-085, FR-087, FR-089, FR-090) o la nota del informe final (FR-091). Tras el rechazo del juez: FR-027 ya no contradice a FR-010, FR-030, FR-031 y FR-052; FR-012 fija el orden de los códigos con varios defectos; FR-031, FR-042 y FR-048 dicen qué nombra la declaración cuando la entrada no trae norma; FR-004 y FR-009 admiten el silencio en días naturales.
- [x] Success criteria are measurable — SC-001 y SC-004 a SC-012, cada uno con su cifra y la comprobación que sale en rojo; SC-003 lo mide una persona y lo dice; SC-002 y SC-013 no tienen una comprobación que salga en rojo: SC-002 lo decide sin modelo el informe final del workflow, y SC-013, la revisión final por lectura.
- [x] Success criteria are technology-agnostic (no implementation details) — se miden por lo que da el job, `make ci` o una persona (SC-001 a SC-013), sin fijar cómo se construye.
- [x] All acceptance scenarios are defined — Dado/Cuando/Entonces en US1 a US6; la Aceptación literal del hito en SC-001 a SC-003.
- [x] Edge cases are identified — «Edge Cases»: art. 30.6, último día inhábil, mes sin día equivalente, festivo autonómico parcial, entidades menores, línea ilegible, acto presunto, argumentos inválidos y la regla genérica.
- [x] Scope is clearly bounded — «Fuera de alcance», con el del hito literal y lo que el hito no especifica.
- [x] Dependencies and assumptions identified — «Relación con hitos anteriores» y «Assumptions».

## Feature Readiness

- [x] All functional requirements have clear acceptance criteria — cada FR remite a un test, un control de `make ci`, el job de evals o la revisión final (FR-080 a FR-090).
- [x] User scenarios cover primary flows — plazo con festivo local (US1), sin festivos locales (US2), derivación de los festivos (US3), cobertura por materia (US4), contencioso (US5) y umbrales (US6).
- [x] Feature meets measurable outcomes defined in Success Criteria — trazabilidad en «Criterios del hito, literales».
- [x] No implementation details leak into specification — como en el primer punto de «Content Quality».

## Notes

- Items marked incomplete require spec updates before `/speckit-clarify` or `/speckit-plan`. Ninguno queda incompleto tras el paso de clarify del workflow y la corrección del rechazo del juez del spec.
