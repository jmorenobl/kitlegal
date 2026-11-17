# Specification Quality Checklist: H41.3 · Por qué `normativa` no se activa con todas las skills instaladas, y si unas descripciones cortas lo arreglan

**Purpose**: Validate specification completeness and quality before proceeding to planning
**Created**: 2026-11-16
**Feature**: [spec.md](../spec.md)

## Content Quality

- [x] No implementation details (languages, frameworks, APIs) — el spec nombra `make evals-laboratorio`,
  `make laboratorio-tabla`, `INSTALACION`, `PRESUPUESTO`, `SLASH_COMMAND_TOOL_CHAR_BUDGET` y `sondeo.json` porque los
  nombra el hito; el fichero y el formato de la variante en `laboratorio/` (FR-001), cómo se compone el `SKILL.md`
  (FR-011), dónde y con qué clave se anota la constancia (FR-012), dónde se declara en el código la regla de FR-010,
  los textos de los mensajes de error nuevos (FR-016; los de hoy no cambian), las cabeceras de las columnas (FR-020)
  y los cuatro textos de la variante los decide el plan.
- [x] Focused on user value and business needs — Contexto (la hipótesis del ADR 0054 y por qué medirla antes de tocar
  las descripciones del producto) y «Uso, de fuera adentro».
- [x] Written for non-technical stakeholders — cada historia dice qué ve la persona en `make ci`, en la sesión o en la
  tabla antes del control.
- [x] All mandatory sections completed — escenarios, requisitos, criterios de éxito, supuestos y «Fuera de alcance».

## Requirement Completeness

- [x] No [NEEDS CLARIFICATION] markers remain — cerrados por clarify: `INSTALACION=producto-descripciones-cortas` solo
  con `normativa-producto` (FR-010) y `sondeo.json` sin las dos opciones como hoy, byte a byte (FR-012).
- [x] Requirements are testable and unambiguous — FR-001 y FR-002 con la comprobación de la variante; FR-010 a FR-018
  con el test del arnés (FR-018), con la forma de `PRESUPUESTO` fijada en FR-013 (solo cifras, sin signo ni ceros
  delante, y que cabe en un `int`, como `SESIONES`; lo que no cabe es error de argumentos de FR-016) y los mensajes de hoy de una petición sin las dos opciones que no vale (FR-015, FR-016); FR-020 a FR-024 con el del guion (FR-024); FR-032 con las suites congeladas; FR-034
  con el job; FR-030 y FR-031 los comprueba la persona (SC-009), con la regla de «no se tocan» fijada en FR-031.
- [x] Success criteria are measurable — SC-001 (≤ 450, ≤ 1700, 0 etiquetas, 1 por skill), SC-002 (4 skills, 100 %,
  0 bytes), SC-003 (100 % y 0 sesiones), SC-004 (0 bytes), SC-005 (0 sesiones), SC-006 (valores exactos), SC-007
  (0 bytes), SC-008 (job de evals), SC-009 (capa 3).
- [x] Success criteria are technology-agnostic (no implementation details) — miden descripciones, skills, bytes,
  entornos, recuentos y resultados de comprobaciones, no cómo se implementan.
- [x] All acceptance scenarios are defined — US1 a US4, en Dado/Cuando/Entonces.
- [x] Edge cases are identified — las dos opciones juntas, `PRESUPUESTO` sin `INSTALACION`, descripciones que el YAML
  cita, variante desfasada de `skills/`, tabla que junta sondeos de H41.2 y de H41.3, fila con `n` = 0 y la regla
  genérica (ADR 0023).
- [x] Scope is clearly bounded — «Fuera de alcance» con la lista «Fuera» del hito entera y lo que el spec no añade
  (otras variantes de instalación, la familia de H39, medir el listado, otras versiones).
- [x] Dependencies and assumptions identified — H41.2 (arnés con `INSTALACION=producto` y `EVAL`, tabla con las
  activaciones, su suite congelada) y ADR 0054; «Assumptions».

## Feature Readiness

- [x] All functional requirements have clear acceptance criteria — FR-001 y FR-002 en US1 y SC-001; FR-010 a FR-018 en
  US2 y SC-002 a SC-005; FR-020 a FR-024 en US3 y SC-006/SC-007; FR-030 y FR-031 en US4 y SC-009; FR-032 en SC-007;
  FR-033 en «Fuera de alcance» y en las suites (no cambia ningún fichero de `skills/`); FR-034 en SC-008.
- [x] User scenarios cover primary flows — la variante (US1), el arnés con las dos opciones (US2), la tabla (US3), el
  sondeo y el ADR (US4).
- [x] Feature meets measurable outcomes defined in Success Criteria — SC-001 a SC-009, con los «Controles» y la
  «Aceptación» del hito literales en «Criterios del hito (literales)».
- [x] No implementation details leak into specification — ver el primer ítem.

## Notes

- Los dos marcadores (FR-010 y FR-012) los cerró clarify con el criterio de la constitución (sesión 2026-11-16).
- Los umbrales que se miden sin persona (≤ 450, ≤ 1700, 0 etiquetas, 100 %, 0 bytes, 0 sesiones, `sin_activar` ≤ 0)
  nombran la comprobación que sale en rojo; el de SC-009 lo mide la persona y lo hace cumplir la decisión de fusionar.
