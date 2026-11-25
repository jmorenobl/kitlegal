# Specification Quality Checklist: H46 · Los cuerpos de las skills de capacidad, con sitio para las materias

**Purpose**: Validate specification completeness and quality before proceeding to planning
**Created**: 2026-11-25
**Feature**: [spec.md](../spec.md)

## Content Quality

- [x] No implementation details (languages, frameworks, APIs) — el spec dice qué escribe la tabla (FR-001), qué
  defecto da la comprobación (FR-005) y qué textos lleva `skills/` (FR-010 a FR-012); los tests y funciones que nombra
  (FR-003, FR-008, FR-009, FR-014, FR-015) son los que nombra el hito como entrega, no un diseño. Cómo se cuentan y se
  quitan las líneas de la región, y los mutantes, quedan para el plan (Assumptions).
- [x] Focused on user value and business needs — lo que gana cada skill de capacidad: la materia siguiente cabe sin
  tocar su protocolo, con las mismas respuestas (Contexto; US2, US5).
- [x] Written for non-technical stakeholders — cada historia dice quién la necesita y por qué (Why this priority).
- [x] All mandatory sections completed — User Scenarios, Requirements, Success Criteria, más Fuera de alcance.

## Requirement Completeness

- [x] No clarification markers remain — ninguno; lo que el hito deja a elegir está en «Assumptions».
- [x] Requirements are testable and unambiguous — cada FR nombra su comprobación (p. ej. FR-005 con su mensaje y el
  límite 9600/9601; FR-010 con `git diff f89093dd -- skills/` vacío); FR-019 y SC-007 acotan «ningún applet devuelve
  otra cosa» a `--describe`, esquema, sobre, códigos y forma de `data`, y nombran como excepción los tests y guiones
  que fijan el texto de las skills (FR-003, FR-008, FR-009, FR-014 a FR-016), sin choque con ellos.
- [x] Success criteria are measurable — SC-001 a SC-008, cada uno con su control que sale en rojo.
- [x] Success criteria are technology-agnostic (no implementation details) — miden `make ci`, `skills-check`, la suite
  y el job de evals, las superficies del proyecto, no un diseño.
- [x] All acceptance scenarios are defined — US1 a US5 en forma Dado/Cuando/Entonces.
- [x] Edge cases are identified — límite exacto, skill sin marcas, región vacía, marcas sin pareja (el defecto de la sincronía de hoy;
  un `SKILL.md` editado a mano pasa por `skills-check` como cualquier otro, FR-005 y US2).
- [x] Scope is clearly bounded — «Fuera de alcance» con la lista cerrada del hito y lo que hace la persona.
- [x] Dependencies and assumptions identified — «Dependencias» (H45.1; H42 depende de este) y «Assumptions».

## Feature Readiness

- [x] All functional requirements have clear acceptance criteria — FR-001 a FR-021, trazados a US y SC.
- [x] User scenarios cover primary flows — tabla (US1), reserva (US2), textos movidos (US3), controles del texto
  (US4), respuestas (US5).
- [x] Feature meets measurable outcomes defined in Success Criteria — «Criterios del hito (literales)» los traza.
- [x] No implementation details leak into specification — ver el primer ítem.

## Rúbrica del juez de entrada (`juez_spec`)

- [x] a. alcance — la entrega 1 a 6 del hito en FR-001 a FR-018; tareas y suite en FR-020 y FR-021; nada más.
- [x] b. fuera de alcance — la lista cerrada del hito, la medida previa, lo de la persona y la copia de H41.
- [x] c. clarificaciones — no hay `clarify-respuestas.json` todavía; «Clarifications» lo dice.
- [x] d. aceptación — «Controles» y «Aceptación» literales, trazados a SC-001 a SC-008.
- [x] e. decisiones cerradas — no se tocan los topes de 11 000/900/500, la forma de la línea, §4, ni lo que devuelve
  un applet (FR-013, FR-019).
- [x] f. checklist veraz — cada marca cita su FR, SC o sección.
- [x] g. sin implementación — ver «Content Quality».
- [x] h. derivable — 9600 en verde, 9601 en rojo; con y sin marcas; orden de los defectos (Assumptions); las marcas
  sin pareja, con el defecto de la sincronía (Edge Cases); FR-019 y SC-007, con lo que fijan y sus excepciones
  nombradas.
- [x] i. uso — «Quién consume cada salida, cuánto y hasta cuándo»: cuerpo una vez por activación, ≤ 9600 + 1400, no
  crece con el uso; las tres referencias, una vez por su orden de leer; el defecto, una línea por skill y `make ci`.
- [x] j. umbral con control — 9600 (SC-002, SC-003: `skills-check` y `casosDePresupuestos` en `make ci`); 11 000
  (SC-002); umbrales de evals y 2 de 3 (SC-008: el job del cierre).

## Notes

- Las cifras de los cuerpos son las del hito y del ADR 0063; esta sesión no pudo medirlas (el sandbox no dejó correr
  el guion de medida). La medida que cuenta es la de `make ci`.
