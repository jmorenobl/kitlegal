# Specification Quality Checklist: H31.1 · `boe-legislacion` habla de la norma, no de la consulta, y lee lo que nombra

**Purpose**: Validate specification completeness and quality before proceeding to planning
**Created**: 2026-10-27
**Feature**: [spec.md](../spec.md)

## Content Quality

- [x] No implementation details (languages, frameworks, APIs) — el spec fija qué texto lleva la skill y qué controles
  existentes lo comprueban (FR-002 a FR-008); no fija diseño de código: dónde va cada texto movido, su clase y cómo se
  recalcula el tope los decide el plan (FR-006, FR-007, FR-015).
- [x] Focused on user value and business needs — US1 y US2 parten de lo que lee quien pregunta; US7, de la decisión de
  fusionar.
- [x] Written for non-technical stakeholders — cada historia dice por qué importa con las cifras medidas; los nombres de
  controles son los de `docs/SKILLS.md` §1.
- [x] All mandatory sections completed — escenarios, requisitos, criterios de éxito, supuestos y «Fuera de alcance».

## Requirement Completeness

- [x] No [NEEDS CLARIFICATION] markers remain — ninguno: el hito fija la skill de partida, los pasajes intactos, las
  lecturas y los umbrales.
- [x] Requirements are testable and unambiguous — cada FR lleva su «Criterio»; los pasajes medidos se fijan por líneas
  del `SKILL.md` de la skill medida (Contexto, FR-002), y «texto de una clase de pregunta» frente a «lo que toda
  respuesta necesita» se define en el Contexto y se lista en FR-006; las remisiones de las referencias al cuerpo siguen
  valiendo porque FR-006 fija encabezados y reglas, con su control en la suite de aceptación (SC-003); y los enlaces de
  FR-014 salen del repositorio, cotejables por la revisión final (SC-013).
- [x] Success criteria are measurable — SC-002 (11 000 caracteres), SC-007 (0.028), SC-008 (0 sesiones), SC-012 (2 y 4
  de 40), y cada umbral sin persona nombra la comprobación que sale en rojo (SC-001 a SC-009, SC-011).
- [x] Success criteria are technology-agnostic (no implementation details) — miden la skill y el cierre por sus
  resultados; las comprobaciones que nombran son las que ya existen.
- [x] All acceptance scenarios are defined — US1 a US7, en Dado/Cuando/Entonces.
- [x] Edge cases are identified — lo remitido fuera de la tabla, la fecha, el sitio que no da, textos movidos con forma
  fija u orden concreta, y la regla genérica para lo manipulado a mano.
- [x] Scope is clearly bounded — «Fuera de alcance» recoge el del hito y lo que no está en él (descripción, referencias más
  allá de recibir textos, `topeDeUnaSesion`, binario, workflow).
- [x] Dependencies and assumptions identified — «Assumptions»: lo reescrito no medido, la falta de red para las guías
  («Las dos fuentes» no da ninguna URL; el repositorio solo registra `https://agentskills.io/`), `a24` ya grabado, la
  serie que empieza de cero.

## Feature Readiness

- [x] All functional requirements have clear acceptance criteria — FR-001 a FR-016, cada uno con su «Criterio».
- [x] User scenarios cover primary flows — la respuesta (US1), lo nombrado (US2), el sitio (US3), las evals (US4), el
  cierre (US5), la guía (US6) y la decisión de la persona (US7).
- [x] Feature meets measurable outcomes defined in Success Criteria — SC-001 a SC-013 se trazan a los FR y a los criterios
  del hito, copiados literalmente en «Criterios del hito, tal como los escribe».
- [x] No implementation details leak into specification — ver «Content Quality».

## Notes

- Comprobado contra la rúbrica de `juez_spec` (criterios a a j) y contra `scripts/workflow/precheck.sh spec`.
- Criterio j: SC-012 lo mide una persona (fuera del criterio); SC-010 no tiene umbral; los demás umbrales nombran su
  comprobación en rojo.
