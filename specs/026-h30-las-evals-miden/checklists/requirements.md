# Specification Quality Checklist: H30 · Las evals miden solo lo que cambia

**Purpose**: Validate specification completeness and quality before proceeding to planning
**Created**: 2026-10-22
**Feature**: [spec.md](../spec.md)

## Content Quality

- [x] No implementation details (languages, frameworks, APIs) — el spec fija contratos observables (qué elige cada
  ruta, qué dice el informe, qué exige la release); las rutas del repositorio que nombra son las que el propio hito
  impone (`internal/cli`, `go list -deps`, `release.yml`, `serie.sh` de la 2.11.0) o las de hoy que FR-002 y FR-004
  describen por su papel.
- [x] Focused on user value and business needs — «Contexto» dice qué gana cada skill y quién consume cada salida.
- [x] Written for non-technical stakeholders — en la medida en que el consumidor es quien revisa una propuesta, el
  cierre del workflow y quien etiqueta una release; cada historia dice el porqué antes del criterio.
- [x] All mandatory sections completed — escenarios (cinco historias), requisitos (FR-001 a FR-027), criterios de éxito
  (SC-001 a SC-016), «Uso, de fuera adentro», supuestos y «Fuera de alcance».

## Requirement Completeness

- [x] No [NEEDS CLARIFICATION] markers remain — los dos que había (FR-014 y la materia de una eval, en «Assumptions»)
  los cerró clarify el 2026-10-22; ver «Clarifications» del spec.
- [x] Requirements are testable and unambiguous — cada FR lleva su «Criterio» o lo cubre un SC con su test de `make ci`;
  FR-014 quedó cerrado por clarify. Tras la primera revisión: FR-001(d) define los paquetes de un applet sin pasar por
  la composición (el paquete raíz y sus skills solo llegan por (b); `_test.go` y `testdata/` fuera, salvo las
  grabaciones); `go.mod` y `go.sum` entran en el conjunto, en lo común y en la huella (FR-001(d), FR-002, FR-004,
  FR-009); y FR-024 deriva la tasa de referencia del `umbral` vigente de `clases.yaml`.
- [x] Success criteria are measurable — SC-002 a SC-012 nombran su cifra y el test de `make ci` que sale en rojo;
  SC-014, la comprobación de `cierre.sh` que detiene el cierre; SC-015 es una restricción del diff del hito
  (FR-027), no una medida con umbral.
- [x] Success criteria are technology-agnostic (no implementation details) — se miden en `make ci`, en el informe del
  job y en la release, que son la superficie del proyecto.
- [x] All acceptance scenarios are defined — cinco historias en Dado/Cuando/Entonces.
- [x] Edge cases are identified — «Edge Cases», cada uno con su requisito: eval nueva o retirada, skills añadidas o
  renombradas, renombrados, dos medidas con la misma huella, reutilización encadenada, resultado incompleto, prueba de
  red, dos etiquetas sobre el mismo commit y lo manipulado a mano (regla genérica).
- [x] Scope is clearly bounded — «Fuera de alcance» lleva lo que el hito deja fuera y lo que el spec no especifica.
- [x] Dependencies and assumptions identified — «Assumptions»: squash, medidas sin huella hasta la fusión, retención de
  la plataforma, la etiqueta que crea una persona, la mención de `boe-legislacion` en `jurisprudencia`.

## Feature Readiness

- [x] All functional requirements have clear acceptance criteria — «Criterio» en cada FR o su SC (FR-011, FR-013,
  FR-016, FR-018 y FR-022, por SC-004, SC-007 y SC-012); FR-014 quedó cerrado por clarify.
- [x] User scenarios cover primary flows — elección (US1), reutilización (US2), release (US3), suite completa (US4) y el
  ADR del umbral en tasa (US5).
- [x] Feature meets measurable outcomes defined in Success Criteria — los «Controles» y la «Aceptación» del hito, tal
  como los escribe, están en SC-002 a SC-007, y el «Alcance» en FR-020 y FR-026.
- [x] No implementation details leak into specification — ver el primer punto.

## Notes

- Los dos [NEEDS CLARIFICATION] los cerró el paso `clarify` del workflow (2026-10-22); este fichero se actualizó tras él.
