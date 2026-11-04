# Specification Quality Checklist: H40 · Por qué `normativa` no lee sus referencias en modo herramienta

**Purpose**: Validate specification completeness and quality before proceeding to planning
**Created**: 2026-11-04
**Feature**: [spec.md](../spec.md)

## Content Quality

- [x] No implementation details (languages, frameworks, APIs) — el spec dice qué contienen las variantes, qué cuenta el
  guion y qué comprueba `make ci`, no dónde viven las variantes, cómo se declaran ni cómo se comparan; los nombres
  `make evals-laboratorio` y `make laboratorio-tabla` los da el hito.
- [x] Focused on user value and business needs — historias 1 a 4: saber la causa del defecto antes de reorganizar
  las skills (Contexto; FR-040, FR-041).
- [x] Written for non-technical stakeholders — cada historia explica su porqué; los términos de la medida (turno,
  activar, abrir una referencia) se definen en FR-020 por remisión a H39.
- [x] All mandatory sections completed — User Scenarios, Requirements, Success Criteria, Assumptions y Fuera de alcance.

## Requirement Completeness

- [x] No [NEEDS CLARIFICATION] markers remain — la de FR-008 (qué skills tiene instalada cada sesión del sondeo) la
  cerró el paso clarify: solo la variante.
- [x] Requirements are testable and unambiguous — cada FR nombra su comprobación: FR-001, FR-002, FR-006, FR-022,
  FR-023 y FR-030 en `make ci` (FR-023, con la suite de aceptación congelada contra la tabla de H39 en `main`); FR-010
  y FR-011 en el test del arnés de SC-006; FR-040 a FR-042, la persona. FR-007 no pide comprobación: solo dice que el
  tope del ADR 0046 no rige las variantes y sigue igual para `skills/`.
- [x] Success criteria are measurable — SC-001 a SC-008 con recuento, porcentaje o bytes y su control; SC-009, capa 3.
- [x] Success criteria are technology-agnostic (no implementation details) — miden bytes, ficheros, sesiones abiertas y
  cifras de la tabla, no cómo se calculan.
- [x] All acceptance scenarios are defined — escenarios Dado/Cuando/Entonces en las cuatro historias.
- [x] Edge cases are identified — cambio de `skills/normativa`, sesión sin herramienta, sin terminar, referencia abierta
  varias veces o ninguna, fila con n = 0, variante sin regla testigo, transcripción ilegible (regla genérica).
- [x] Scope is clearly bounded — «Fuera de alcance» con la reorganización, los topes, el laboratorio de H39, el juez,
  los otros modos, evals, modelos y hosts, y el sondeo y el ADR dentro del run.
- [x] Dependencies and assumptions identified — dependencia de H39 (arnés, guion, regla testigo; FR-010, FR-020,
  FR-030) y supuestos sobre las cifras de partida, las grabaciones y el tope del ADR 0046 (Assumptions).

## Feature Readiness

- [x] All functional requirements have clear acceptance criteria — escenarios de las historias 1 a 4 y SC-001 a
  SC-009, cada uno con los FR que cubre.
- [x] User scenarios cover primary flows — variantes, arnés, tabla y regla testigo, y la aceptación de la persona.
- [x] Feature meets measurable outcomes defined in Success Criteria — cada SC remite a su FR y a su control.
- [x] No implementation details leak into specification — ver el primer ítem.

## Criterios del hito y del juez

- [x] Los criterios de «Aceptación» y «Controles» del hito están literalmente en el spec — sección «Criterios del hito
  (literales)», con los FR y SC de cada uno.
- [x] Cada umbral que se mide sin persona nombra la comprobación que sale en rojo — SC-001 a SC-006, un test de
  `make ci`; SC-007, la comparación de la suite de aceptación congelada con la tabla esperada de H39 en `main`, que
  sale en rojo aunque el hito cambie esa tabla o los tests de H39; SC-008, el job de evals del cierre; SC-009 lo mide la persona y lo dice.
- [x] Uso de fuera adentro — sección «Uso, de fuera adentro»: quién pide cada salida, cuántas veces, su tamaño (cinco
  filas, 200 sesiones, variantes de tamaño fijo) y que no crece con el uso ni es una señal que caduque.
- [x] Ninguna decisión cerrada de CLAUDE.md se reabre — el tope de 11 000 del ADR 0046 sigue igual para `skills/`
  (FR-007, Fuera de alcance); las variantes no se distribuyen (FR-001); el juez no mide el sondeo (Fuera de alcance);
  la persona lanza el sondeo con su credencial (FR-012; ADR 0032).

## Notes

- La marca de FR-008 la cerró el paso clarify con el criterio de la constitución (sesión 2026-11-04).
