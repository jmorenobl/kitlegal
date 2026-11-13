# Specification Quality Checklist: H41.2 · Por qué la eval 04 no activa `normativa` en modo herramienta en el job

**Purpose**: Validate specification completeness and quality before proceeding to planning
**Created**: 2026-11-11
**Feature**: [spec.md](../spec.md)

## Content Quality

- [x] No implementation details (languages, frameworks, APIs) — el spec nombra `informe.json`, `informe.md`,
  `make evals-laboratorio` y `make laboratorio-tabla` porque los nombra el hito; el nombre y los valores de la opción
  (FR-010), dónde queda la constancia de la instalación (FR-012) y la forma de la lista en cada fichero los decide el
  plan.
- [x] Focused on user value and business needs — Contexto (por qué medir antes de tocar el umbral) y «Uso, de fuera
  adentro».
- [x] Written for non-technical stakeholders — cada historia dice qué ve la persona en el informe, en la sesión o en la
  tabla antes del control.
- [x] All mandatory sections completed — escenarios, requisitos, criterios de éxito, supuestos y «Fuera de alcance».

## Requirement Completeness

- [x] No [NEEDS CLARIFICATION] markers remain — el de FR-013 (cómo se pide solo la eval 04 al arnés) lo cerró
  clarify con la opción A: un argumento nuevo del arnés.
- [x] Requirements are testable and unambiguous — FR-001 a FR-006 con su test de informe; FR-010 a FR-015 con el del
  arnés (lo que instala depende de la opción de FR-010; las sesiones que abre, solo del argumento de FR-013);
  FR-020 a FR-024 con el del guion; FR-032 y FR-033 con las suites congeladas y el job; FR-030 y FR-031 los
  comprueba la persona (SC-008).
- [x] Success criteria are measurable — SC-001 (100 %), SC-002 (0 diferencias), SC-003 (0 bytes, 4 skills, 0
  repetidas), SC-004 (0 sesiones), SC-005 (valores exactos), SC-006 (0 bytes), SC-007 (job de evals), SC-008 (capa 3).
- [x] Success criteria are technology-agnostic (no implementation details) — miden listas, recuentos, bytes y
  resultados de comprobaciones, no cómo se implementan.
- [x] All acceptance scenarios are defined — US1 a US4, en Dado/Cuando/Entonces.
- [x] Edge cases are identified — activación después de la primera herramienta, repetida, en el mismo turno, sesión
  sin herramientas, sesión ilegible, informe anterior, sondeo sin terminar, sondeo anterior junto a uno nuevo y la regla
  genérica (ADR 0023).
- [x] Scope is clearly bounded — «Fuera de alcance» con la lista «Fuera» del hito entera y lo que el spec no añade
  (otras columnas, otras instalaciones, la familia de H39).
- [x] Dependencies and assumptions identified — H41 (arnés, tabla y suites congeladas); «Assumptions».

## Feature Readiness

- [x] All functional requirements have clear acceptance criteria — FR-001 a FR-006 en US1 y SC-001/SC-002; FR-010 a
  FR-015 en US2 y SC-003/SC-004/SC-009; FR-020 a FR-024 en US3 y SC-005/SC-006; FR-030 y FR-031 en US4 y SC-008;
  FR-032 en SC-006; FR-033 en SC-007 (evals en verde y el informe final con las cifras del umbral sin cumplir, como
  hoy) y, para la lista de activaciones del cierre, en SC-001 (la del informe del job, FR-001 y FR-002).
- [x] User scenarios cover primary flows — informe del job (US1), instalación del arnés (US2), tabla (US3), sondeo y
  ADR (US4).
- [x] Feature meets measurable outcomes defined in Success Criteria — SC-001 a SC-008, con los «Controles» y la
  «Aceptación» del hito literales.
- [x] No implementation details leak into specification — ver el primer ítem.

## Notes

- El marcador de FR-013 lo cerró clarify (Session 2026-11-11, opción A); no queda ningún ítem sin marcar.
- Los umbrales que se miden sin persona (100 %, 0 diferencias, 0 bytes, 0 sesiones, `sin_activar` ≤ 0) nombran la
  comprobación que sale en rojo; el de SC-008 lo mide la persona y lo hace cumplir la decisión de fusionar.
