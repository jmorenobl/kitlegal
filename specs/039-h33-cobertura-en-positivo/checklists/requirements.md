# Specification Quality Checklist: H33 · Cobertura en positivo por capacidad y materia

**Purpose**: Validate specification completeness and quality before proceeding to planning
**Created**: 2026-11-20
**Feature**: [spec.md](../spec.md)

## Content Quality

- [x] No implementation details (languages, frameworks, APIs) — los nombres de ficheros, tests y funciones que aparecen
  son los que fija el propio hito como entrega y control (FR-006, FR-007, FR-020, FR-021); el diseño (cómo se parte el
  generador, dónde vive cada comprobación nueva) queda para el plan.
- [x] Focused on user value and business needs — historias 1 a 3: la persona sabe qué no está cubierto, la frontera vive
  en datos, una sola plantilla.
- [x] Written for non-technical stakeholders — «Contexto» e historias en lenguaje del dominio; lo técnico es lo que el
  hito exige como control.
- [x] All mandatory sections completed — historias, requisitos, criterios de éxito, supuestos y «Fuera de alcance».

## Requirement Completeness

- [x] No [NEEDS CLARIFICATION] markers remain — el hito y el ADR 0058 fijan textos, datos y controles.
- [x] Requirements are testable and unambiguous — cada FR nombra su control (p. ej. FR-001 `TestLeerCobertura`, FR-013
  mutantes, FR-025 el guion e2e); FR-010 fija el cuerpo de `plazos-y-territorio` en el de FR-009, sin texto añadido ni
  movido a una referencia, así que no abre excepción a FR-005, FR-009, FR-020 ni «Fuera de alcance» (ver
  «Clarifications»).
- [x] Success criteria are measurable — SC-001 a SC-010 con recuentos, bytes o caracteres y su control.
- [x] Success criteria are technology-agnostic (no implementation details) — miden lo que ve quien instala o pregunta
  (líneas, filas, umbrales); los controles nombrados son los que pide el hito.
- [x] All acceptance scenarios are defined — Dado/Cuando/Entonces en las cinco historias.
- [x] Edge cases are identified — sin cobertura o con defectos, dos filas de una capacidad, cero o dos plantillas,
  lecturas fuera de lo grabado, y la regla genérica para lo manipulado a mano.
- [x] Scope is clearly bounded — «Fuera de alcance» recoge el «Fuera» del hito y lo que el run no hace.
- [x] Dependencies and assumptions identified — «Assumptions» (base `7a7526e8`, los 140 caracteres, el mensaje del
  marcador, las cifras no medidas) y dependencias H32 y H41.4 en el contexto de las suites congeladas (FR-020).

## Feature Readiness

- [x] All functional requirements have clear acceptance criteria — cada FR de FR-001 a FR-026 lleva su control o su
  verificación (FR-022, revisión y quickstart).
- [x] User scenarios cover primary flows — fuera de lo declarado (historia 1), lo cubierto en datos (2), lo regulado
  aparte (3), evals y umbrales (4), lo instalado (5).
- [x] Feature meets measurable outcomes defined in Success Criteria — SC-001 a SC-010 trazados a FR-001–FR-026.
- [x] No implementation details leak into specification — ver la primera marca: solo lo que el hito fija.

## Notes

- Los criterios «Controles» y «Aceptación» del hito van literales en «Criterios del hito (literales)».
- Cada umbral numérico medido sin persona nombra la comprobación que sale en rojo: 0 de `cobertura_sin_declarar` y de
  `sin_activar` (job de evals, FR-014, FR-024), 11 000 y 10 860 caracteres (FR-010), 15 013 ≤ 21 120 s (FR-019), al
  menos 2 normas (FR-017).
