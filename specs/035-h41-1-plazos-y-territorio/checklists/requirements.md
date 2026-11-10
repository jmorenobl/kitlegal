# Specification Quality Checklist: H41.1 · `plazos-y-territorio` manda leer sus referencias sin condición

**Purpose**: Validate specification completeness and quality before proceeding to planning
**Created**: 2026-11-10
**Feature**: [spec.md](../spec.md)

## Content Quality

- [x] No implementation details (languages, frameworks, APIs) — el spec fija textos, ficheros y los nombres de los
  controles porque los da el hito («Controles»); cómo se escriben los tests (declaraciones, funciones) queda al plan
  (FR-005 a FR-008).
- [x] Focused on user value and business needs — «Lo que gana la skill» en Contexto; US1 y «Uso, de fuera adentro».
- [x] Written for non-technical stakeholders — cada historia dice qué ve la persona o el agente antes del control.
- [x] All mandatory sections completed — escenarios, requisitos, criterios de éxito, supuestos y «Fuera de alcance».

## Requirement Completeness

- [x] No [NEEDS CLARIFICATION] markers remain — el de FR-007 (la viñeta de `jerarquia_normativa.md` frente a la
  regla del verbo de `TestRemisionesDelRepositorio`) lo cerró clarify con la opción C.
- [x] Requirements are testable and unambiguous — FR-001 a FR-013, cada uno con su comprobación; FR-007 fija la tabla
  de fallos admitidos que eligió clarify (opción C), y FR-012 se acota a lo que no piden FR-001 a FR-011 y FR-013, de
  modo que el cambio del guion `h41-skills-instaladas.txtar` (FR-013) y `make ci` en verde (SC-001) se cumplen a la vez.
- [x] Success criteria are measurable — SC-001 (mutaciones en rojo), SC-002 (10 882, 10 952, 11 000), SC-003 (Wilson al
  95 %), SC-004 (job de evals), SC-005 (cero bytes).
- [x] Success criteria are technology-agnostic (no implementation details) — miden textos, recuentos y resultados de
  comprobaciones, no cómo se implementan.
- [x] All acceptance scenarios are defined — US1 a US5, en Dado/Cuando/Entonces.
- [x] Edge cases are identified — región de los paquetes, pregunta solo de territorio, paso 7 que para, `jerarquia_normativa.md`,
  ficheros cambiados a mano (regla genérica, ADR 0023).
- [x] Scope is clearly bounded — FR-012, FR-013 (el guion de H41 que afirmaba v0.9.0) y «Fuera de alcance», con la
  lista «Fuera» del hito entera y el ejemplo de remisión de `docs/SKILLS.md` §1 que el hito no toca.
- [x] Dependencies and assumptions identified — H41 (contrato de variantes, arnés, ADR 0053); «Assumptions».

## Feature Readiness

- [x] All functional requirements have clear acceptance criteria — FR-001 a FR-003 en US1 y US2; FR-004 a FR-008 en
  US3 y US4; FR-009 y FR-010 en US5; FR-011 en SC-004; FR-012 en SC-005; FR-013 en US3, escenario 1, y SC-001.
- [x] User scenarios cover primary flows — pregunta de territorio y de plazos (US1), M1 (US2), controles (US3), sondeo
  (US4), guía y registro (US5).
- [x] Feature meets measurable outcomes defined in Success Criteria — SC-001 a SC-005, con los «Controles» y la
  «Aceptación» del hito literales.
- [x] No implementation details leak into specification — ver el primer ítem.

## Notes

- La contradicción aparente de FR-007 sale del propio hito: «La línea de `jerarquia_normativa.md` … queda como en
  v0.9.0» y «`TestRemisionesDelRepositorio` … deja de declararla fuera hasta H41.1». Con la exclusión quitada en una
  copia de trabajo, `go test ./internal/skills -run 'TestRemisionesDelRepositorio$'` dio, entre otros, «plazos-y-territorio:
  references/jerarquia_normativa.md: un enlace suyo no va en una frase que lo mande leer («lee» antes del enlace)»; el
  cambio se deshizo con `git checkout`.
- Los umbrales que se miden sin persona (11 000, 10 882, 10 952, cero bytes, evals en verde) nombran la comprobación que
  sale en rojo; el de SC-003 lo mide la persona y lo hace cumplir la decisión de fusionar.
