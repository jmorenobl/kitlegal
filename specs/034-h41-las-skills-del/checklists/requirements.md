# Specification Quality Checklist: H41 · Las skills del producto mandan leer sus referencias

**Purpose**: Validate specification completeness and quality before proceeding to planning
**Created**: 2026-11-06
**Feature**: [spec.md](../spec.md)

## Content Quality

- [x] No implementation details (languages, frameworks, APIs) — el spec dice qué remisiones, variantes, filas y
  recuentos hay y qué comprobación sale en rojo; los nombres de las variantes de `plazos-y-territorio`, el texto exacto
  de cada orden, la forma de la lista de referencias del sondeo y dónde vive cada test los fija el plan (FR-001,
  FR-023, FR-032). Los nombres de `make`, `skills/`, `laboratorio/` y los tests que el hito nombra son la superficie del
  repositorio que el hito da, no diseño.
- [x] Focused on user value and business needs — lo que gana la skill (Contexto; «Uso, de fuera adentro»): que en modo
  herramienta `normativa` abra lo que necesita, que ninguna de las dos abra menos, y saber qué forma aplica H41.1.
- [x] Written for non-technical stakeholders — dentro de lo posible en un hito sobre el texto de las skills y su medida:
  cada historia dice quién hace qué y por qué.
- [x] All mandatory sections completed — historias, escenarios, casos límite, requisitos, entidades, criterios de
  éxito, supuestos y «Fuera de alcance».

## Requirement Completeness

- [x] No [NEEDS CLARIFICATION] markers remain — las dudas se cierran con la lectura conservadora en «Assumptions» y
  «Fuera de alcance» (el desglose por eval desde el detalle; ningún otro movimiento que los dos de FR-006).
- [x] Requirements are testable and unambiguous — cada FR nombra la comprobación de `make ci` que sale en rojo (FR-001,
  FR-006 a FR-010, FR-020 a FR-025, FR-030, FR-032, FR-040 a FR-042) o la medida de la persona (FR-052, FR-053).
  FR-033 y US4-1 piden las filas del sondeo en cuatro invocaciones (dos de `normativa`, porque `MODOS` vale para todas
  las variantes de una invocación según FR-030), sin contradecir a FR-030; SC-008 las cuenta igual.
- [x] Success criteria are measurable — SC-001 a SC-010 con recuentos o bytes y su control; SC-011 y SC-012 con la
  regla de los intervalos que lee la persona.
- [x] Success criteria are technology-agnostic (no implementation details) — miden frases, bytes, sesiones y recuentos,
  no cómo se calculan.
- [x] All acceptance scenarios are defined — las seis historias, en Dado/Cuando/Entonces, y los criterios literales del
  hito en «Criterios del hito (literales)».
- [x] Edge cases are identified — tope del cuerpo, variante por encima de 10 930, cambio en un control o en
  `skills/plazos-y-territorio`, cambio de `skills/` entre lanzamiento, sondeo y tabla, referencias sin eval, sesiones
  sin terminar, mezcla de familias y transcripción manipulada (regla genérica, ADR 0023).
- [x] Scope is clearly bounded — «Fuera de alcance» recoge el «Fuera» del hito y lo que no especifica (FR-010, FR-024,
  FR-031).
- [x] Dependencies and assumptions identified — H40 y la rama `033-h41-las-skills-del` (Contexto, Assumptions).

## Feature Readiness

- [x] All functional requirements have clear acceptance criteria — cada FR va a un SC o a un escenario (SC-001 a
  SC-012).
- [x] User scenarios cover primary flows — remisiones (US1), movimientos (US2), `plazos-y-territorio` y sus variantes
  (US3), arnés (US4), tabla (US5) y decisión de la persona (US6).
- [x] Feature meets measurable outcomes defined in Success Criteria — los umbrales medibles sin persona tienen su
  control en `make ci` o en el job de evals del cierre (SC-001 a SC-010); los de la persona se declaran de capa 3
  (SC-011, SC-012).
- [x] No implementation details leak into specification — ver el primer punto.

## Notes

- Rúbrica de `juez_spec` revisada antes de terminar: alcance (a) y «Fuera de alcance» (b) contra el hito; aceptación y
  controles literales (d); decisiones cerradas (e: tope de 11 000, descripciones, generador, CENDOJ, sin red en el
  run); uso de fuera adentro (i, con tamaños); y umbrales con control (j: cada umbral medible sin persona nombra el test
  que sale en rojo).
- `scripts/workflow/precheck.sh spec` no se pudo ejecutar en esta sesión (la orden pedía aprobación); sus
  comprobaciones de la fase spec se reprodujeron con `grep` y no dan ningún defecto: 0 marcas de clarificación en
  spec.md, «## Fuera de alcance» presente, 0 ítems sin marcar, 0 ids **FR**/**SC** definidos dos veces, 0 citados sin
  definir y 22 escenarios Dado/Cuando/Entonces.
