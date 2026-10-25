# Specification Quality Checklist: H31.1 · `boe-legislacion` habla de la norma, no de la consulta, y lee lo que nombra

**Purpose**: Validate specification completeness and quality before proceeding to planning
**Created**: 2026-10-25
**Feature**: [spec.md](../spec.md)

## Content Quality

- [x] No implementation details (languages, frameworks, APIs) — el spec fija lo observable: qué dice el cuerpo de la
  skill (FR-001 a FR-004), qué bloques llevan las `lecturas` (FR-009) y qué control sale en rojo (SC-001 a SC-011); los
  nombres de controles, tests y ficheros (`skills-check`, `TestDefinicionDelJob`, `grabaciones.json`) son los que nombra
  el hito o la superficie de hoy; cómo se reparte el cuerpo lo decide el plan (FR-006, FR-013).
- [x] Focused on user value and business needs — «Contexto» dice lo medido y qué gana cada respuesta; «Uso, de fuera
  adentro» lo traza por salida.
- [x] Written for non-technical stakeholders — en la medida en que los consumidores son el agente que carga la skill,
  quien escribe skills y quien decide la fusión; cada historia da su porqué antes del criterio.
- [x] All mandatory sections completed — seis historias, FR-001 a FR-014, «Uso, de fuera adentro», SC-001 a SC-013 con
  los criterios del hito tal como los escribe, supuestos y «Fuera de alcance».

## Requirement Completeness

- [x] No [NEEDS CLARIFICATION] markers remain — el de FR-013 (el tope de una sesión, 240 s) se cerró en `clarify`: no
  cambia, y queda en «Fuera de alcance».
- [x] Requirements are testable and unambiguous — cada FR lleva su «Criterio»; «Contexto» fija qué son «la regla 5», «el
  paso 3» y los dos pasajes de `27aeccc` (su paso 5 entero y la última viñeta de su paso 4), los mismos que nombran
  FR-001, US1 y SC-003; FR-001 enumera los seis ejemplos que se buscan y dice que en lo remitido y en el precepto no
  leído prevalecen FR-002 y FR-003, que dicen las dos excepciones (el segundo salto y la lectura que falla) sin
  contradecirse; FR-012 da las cifras de `LEEME.md`, con `con_herramientas` solo en los brazos con la skill de H31.
- [x] Success criteria are measurable — SC-001, SC-002, SC-004 y SC-005, en `make ci`; SC-007 a SC-011, en el job de
  cierre, con su veredicto; SC-008 con el control de SC-004 y la falta de la preparación; SC-012, el sondeo que lanza
  una persona; SC-003 y SC-013, la revisión final.
- [x] Success criteria are technology-agnostic (no implementation details) — se miden en `make ci`, en el job de evals,
  en la revisión y en el sondeo, que son la superficie del proyecto.
- [x] All acceptance scenarios are defined — seis historias en Dado/Cuando/Entonces.
- [x] Edge cases are identified — lo remitido de otra norma o que falla, lo remitido que no se nombra, la pregunta con
  fecha, el cuerpo que no cabe, un ejemplo o una orden que tumba un control, y lo manipulado a mano (regla genérica).
- [x] Scope is clearly bounded — «Fuera de alcance» lleva lo que excluye el hito y lo que el spec no especifica: la
  descripción (FR-008), otras evals, el binario y el arnés salvo el tope (FR-013), el sondeo, el resto de la guía y el
  proceso.
- [x] Dependencies and assumptions identified — «Assumptions»: ids en los índices grabados, `a24` ya grabado, serie de
  cierres desde cero, versión de Claude Code del job y del sondeo, versión de la skill en `CHANGELOG.md`.

## Feature Readiness

- [x] All functional requirements have clear acceptance criteria — «Criterio» en cada FR, con su SC.
- [x] User scenarios cover primary flows — la respuesta que habla de la norma (US1), leer lo que se nombra (US2), las
  lecturas grabadas (US3), el cierre (US4), la guía (US5) y la decisión de fusionar (US6).
- [x] Feature meets measurable outcomes defined in Success Criteria — cada historia apunta a sus SC.
- [x] No implementation details leak into specification — el reparto del cuerpo, el cálculo del tope y los tests los
  decide el plan.

## Notes

- El paso `clarify` (2026-10-25) cerró el único marcador (FR-013) y precisó el salto de FR-003; la lista queda completa.
