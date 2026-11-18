# Specification Quality Checklist: H41.4 · Las skills del producto, con las descripciones cortas

**Purpose**: Validate specification completeness and quality before proceeding to planning
**Created**: 2026-11-17
**Feature**: [spec.md](../spec.md)

## Content Quality

- [x] No implementation details (languages, frameworks, APIs) — el spec dice qué se compara y qué sale en rojo
  (FR-001 a FR-005); cómo se fija lo esperado de `consulta-juridica` y cómo se parte el bloque `>-` quedan al plan
  («Assumptions»). Los nombres de tests y de `skills.LeerFrontmatter` son los que da el propio hito.
- [x] Focused on user value and business needs — User Story 1 y «Uso, de fuera adentro»: `normativa` se activa antes
  de la primera herramienta `boe_…`.
- [x] Written for non-technical stakeholders — cada historia empieza por quien instala o mantiene kitlegal.
- [x] All mandatory sections completed — User Scenarios, Requirements, Success Criteria, Assumptions y «## Fuera de
  alcance».

## Requirement Completeness

- [x] No [NEEDS CLARIFICATION] markers remain — ninguno; «Clarifications» dice por qué.
- [x] Requirements are testable and unambiguous — FR-001 a FR-007, FR-009 y FR-010 nombran su control y cuándo sale
  en rojo; FR-005 dice qué espera cada subtest («desde control»: base y declaración; «frontmatter»: el de su control),
  en los dos con la `description` de la variante, sin contradecir FR-003 ni FR-009; FR-008 se verifica en el
  quickstart y la revisión, sin control en `make ci` («## Fuera de alcance»), porque la release cierra
  «[Unreleased]»; FR-011 es el sondeo de la persona, con su criterio literal.
- [x] Success criteria are measurable — SC-001 a SC-006 con recuentos, bytes, caracteres o intervalos.
- [x] Success criteria are technology-agnostic (no implementation details) — se expresan en descripciones, bytes,
  caracteres y sesiones; los controles nombrados son los de `make ci` y del job que el hito exige.
- [x] All acceptance scenarios are defined — Dado/Cuando/Entonces en las cuatro historias.
- [x] Edge cases are identified — quinta skill sin texto, variante cambiada en `main`, lo manipulado a mano (regla
  genérica, ADR 0023).
- [x] Scope is clearly bounded — «## Fuera de alcance» recoge la lista «Fuera» del hito y lo que no entra (web,
  release, sondeo, ADR siguiente).
- [x] Dependencies and assumptions identified — base de la rama `ab951c22`, H41.3 (variante y su suite) y
  «Assumptions».

## Feature Readiness

- [x] All functional requirements have clear acceptance criteria — FR-001 a FR-011 trazados a escenarios y a SC-001
  a SC-006.
- [x] User scenarios cover primary flows — aplicar las descripciones (US1), no cambiar nada más (US2), suites y
  generado (US3), decisión de la persona (US4).
- [x] Feature meets measurable outcomes defined in Success Criteria — SC-001 a SC-005 en `make ci` y el job; SC-006,
  la persona.
- [x] No implementation details leak into specification — ver el primer ítem.

## Rúbrica del juez de spec y precheck

- [x] Los criterios de «Aceptación» y «Controles» del hito están literales en «Criterios del hito (literales)».
- [x] Cada umbral que se mide sin persona nombra la comprobación que sale en rojo: 450 y 1700 (FR-001/FR-004), 900 y
  XML (`TestSkillsDelRepositorio`, FR-004), byte a byte (FR-001, FR-003, FR-005), evals y tercera persona (job de
  evals, FR-010). SC-006 lo mide la persona.
- [x] «Uso, de fuera adentro» da quién consume cada salida, cuántas veces, su tamaño con meses de uso y que no hay
  señales que caduquen.
- [x] Ids **FR-nnn** y **SC-nnn** únicos, sin referencias a ids no definidos.

## Notes

- La tercera persona de las descripciones no tiene control en `make ci`: la juzga el job (ADR 0046, punto 4), como
  dice el propio hito.
