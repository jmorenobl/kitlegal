# Specification Quality Checklist: H38 · Las skills instaladas se actualizan con el binario

**Purpose**: Validate specification completeness and quality before proceeding to planning
**Created**: 2026-11-21
**Feature**: [spec.md](../spec.md)

## Content Quality

- [x] No implementation details (languages, frameworks, APIs) — el spec nombra `Planificar`, `ComprobarConflictos`,
  `cache.Directorio()` y `Enlazador.Disponible` porque el hito obliga a reutilizarlos (FR-005, FR-009, FR-010); no fija
  paquetes, tipos ni el orden de las escrituras, que son del plan.
- [x] Focused on user value and business needs — US1 (las skills al día sin ejecutar nada) y US2 (lo tocado no se pisa).
- [x] Written for non-technical stakeholders — cada historia empieza por lo que ve quien actualiza el binario.
- [x] All mandatory sections completed — escenarios, requisitos, criterios de éxito, supuestos y «Fuera de alcance».

## Requirement Completeness

- [x] No [NEEDS CLARIFICATION] markers remain — ninguno: el hito y el ADR 0060 fijan condiciones, líneas y candado.
- [x] Requirements are testable and unambiguous — cada FR tiene un escenario Dado/Cuando/Entonces o un control
  literal del hito (sección «Criterios del hito»); las líneas de FR-006 y FR-009 están dadas byte a byte. La salida de
  error se mide contra la de la misma orden sin skills instaladas («Una línea de aviso más»), así que la descripción
  de `--dry-run` y el mensaje de un error caben; «nada escrito» lo define FR-013 sobre lo que escribe la
  reinstalación, con la caché comparada con la de esa misma orden, así que el grafo del mundo que escribe la consulta
  no cuenta.
- [x] Success criteria are measurable — SC-001 a SC-006 dan recuentos de líneas, bytes, minutos o conexiones, y cada
  uno nombra la comprobación de `make ci` que sale en rojo.
- [x] Success criteria are technology-agnostic (no implementation details) — se miden sobre lo que ve quien usa el
  binario (salida de error, salida estándar, código, árbol); los controles nombran el test que los hace cumplir.
- [x] All acceptance scenarios are defined — US1 a US5 cubren todos los controles del hito (mapa en «Criterios del
  hito»).
- [x] Edge cases are identified — binario de desarrollo, `HOME`, local ilegible, empotradas nuevas, no empotradas,
  `KITLEGAL_CACHE_DIR` inválido y la regla genérica.
- [x] Scope is clearly bounded — FR-016 y «Fuera de alcance», con todo el «Fuera» del hito.
- [x] Dependencies and assumptions identified — «Assumptions» (el `chmod`, los binarios del arnés, los documentos de
  H19); el hito no tiene dependencias.

## Feature Readiness

- [x] All functional requirements have clear acceptance criteria — FR-001 a FR-019, cada uno con su escenario o su
  control (FR-017: US5; FR-018 y FR-019: SC-007; la cláusula de dos invocaciones a la vez de FR-010: US4.6 y SC-004).
- [x] User scenarios cover primary flows — reinstalar (US1), no pisar (US2), exclusiones y `--dry-run` (US3), candado
  y fallos (US4), documentación (US5).
- [x] Feature meets measurable outcomes defined in Success Criteria — SC-001 es la aceptación literal del hito.
- [x] No implementation details leak into specification — ver la primera nota de «Content Quality».

## Notes

- La sección «Quién consume cada salida, cuánto y hasta cuándo» responde al criterio i del juez: una línea de unos
  115 bytes, una vez por actualización del binario, independiente del uso acumulado.
- Los requisitos de H19 se citan sin guion («FR 044 de H19») para que el precheck no los tome por requisitos de este
  spec sin definir.
