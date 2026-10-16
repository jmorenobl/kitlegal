# Specification Quality Checklist: H23.1 · `jurisprudencia` coteja el documento antes de decir nada de él, y un umbral sin modelo lo hace cumplir

**Purpose**: Validate specification completeness and quality before proceeding to planning
**Created**: 2026-10-16
**Feature**: [spec.md](../spec.md)

## Content Quality

- [x] No implementation details (languages, frameworks, APIs) — el spec dice qué cuenta el umbral, sobre qué sesiones, con qué valor y qué dice la skill. Los nombres que aparecen (`umbrales`, `informe.json`, `comandos_ejecutados`, `cita_cotejar`, `make skills-check`, `TestDefinicionDelJob`, `TestCopiasDelJuez`) los fija el hito o existen en el repositorio. Los nombres de los tests, el texto de la `descripcion` y del motivo, el orden de los quince elementos y cómo se aplican los mutantes quedan para el plan (Assumptions, última viñeta).
- [x] Focused on user value and business needs — Resumen y US2: que un despacho no cite una sentencia que no es la que cree; US1, que el job lo haga fallar.
- [x] Written for non-technical stakeholders — cada historia empieza por lo que ve quien pregunta o quien lee un informe; el detalle está en los requisitos, como en H25 y H26.
- [x] All mandatory sections completed — User Scenarios, Requirements, Success Criteria, Assumptions y «Fuera de alcance».

## Requirement Completeness

- [x] No [NEEDS CLARIFICATION] markers remain — vale tras el paso de clarify, que corre antes del juez del spec. La única que dejó el spec, en FR-033 (el sondeo de la Aceptación humana), la cerró clarify el 2026-10-16 (Q1, opción A: el sondeo no cambia) y está en «Clarifications» y en «Fuera de alcance»; `grep -c` da 0.
- [x] Requirements are testable and unambiguous — cada FR lleva un **Control** o un «Comprobable»; FR-001 y FR-002 dicen qué sesiones entran en el `total` y qué es cotejar; FR-062 enumera los casos de los tests; FR-020 a FR-022 y FR-040 a FR-045, que no tienen test, los comprueba la revisión final por lectura (SC-007, SC-008).
- [x] Success criteria are measurable — SC-001 a SC-009 con recuentos o número de elementos; SC-010 es la parte humana de la Aceptación.
- [x] Success criteria are technology-agnostic (no implementation details) — miden sesiones, elementos de `umbrales`, líneas y veredictos; los controles nombran la comprobación que falla, como exige el ADR 0029.
- [x] All acceptance scenarios are defined — US1 a US4, en forma Dado/Cuando/Entonces (15 escenarios).
- [x] Edge cases are identified — «Edge Cases»: el cotejo sin la referencia, el que termina con 2, el de `--dry-run`, varios cotejos, una sesión con cotejo y otra sin él, la que contradice al cotejo, la sesión cortada, la que no activó la skill, ninguna sesión medida, un solo modo, una skill sin evals que pidan el cotejo y el cierre en rojo.
- [x] Scope is clearly bounded — «Fuera de alcance», con las cuatro viñetas literales del hito y lo no especificado, incluido que el sondeo no cambia (FR-033).
- [x] Dependencies and assumptions identified — «Assumptions» y «Relación con H23, H25 y H26».

## Feature Readiness

- [x] All functional requirements have clear acceptance criteria — tabla de trazabilidad de «Criterios del hito, literales» y controles FR-060 a FR-066.
- [x] User scenarios cover primary flows — el umbral (US1), la skill (US2), lo que sigue igual (US3) y los documentos (US4).
- [x] Feature meets measurable outcomes defined in Success Criteria — SC-001, SC-009 y SC-010 recogen la Aceptación literal; SC-002 a SC-008, los Controles.
- [x] No implementation details leak into specification — ver el primer ítem.

## Criterios del juez del spec (workflow `hito`, `juez_spec`)

- [x] a. Alcance: cubre las tres partes de la Entrega y nada más (tabla de trazabilidad). No especifica caso a caso ficheros manipulados a mano.
- [x] b. Fuera de alcance: las cuatro viñetas literales del hito y lo no especificado. Lo no especificado es: exigir la referencia en el cotejo, otro umbral, la `description` y los demás pasos de la skill, el contrato de `umbrales`, `scripts/workflow/`, el esquema de eval, el comentario de `.github/workflows/evals.yml` como pendiente para la persona, la viñeta de H23.1 de `CLAUDE.md`, un test que ate los documentos y la release.
- [x] c. Clarificaciones: la pregunta de FR-033 está respondida en «Clarifications» (sesión 2026-10-16, Q1 = A).
- [x] d. Aceptación: «Criterios del hito, literales» lleva Entrega, Controles y Aceptación transcritos. En la sesión que escribió el spec se comprobó sin modelo, con `grep -cF` sobre `docs/ROADMAP.md` y `spec.md`, que la Entrega, la Aceptación, tres viñetas de Controles y tres de «Fuera de alcance» están en los dos ficheros, byte a byte. SC-001 a SC-010.
- [x] e. Decisiones cerradas: el umbral con 0 es el del ADR 0038 para lo que comprueba un guion; no hay juez para esto (FR-002 lo mide sin modelo); ninguna sesión del run abre una sesión con modelo (FR-051); el modelo que decide, las repeticiones y la regla por serie no cambian (FR-032; «Fuera de alcance»); `boe-legislacion` y `legal-core` no se tocan (FR-034, FR-050).
- [x] f. Checklist veraz: cada marca cita dónde se cumple; ningún ítem queda sin marcar.
- [x] g. Sin implementación: el texto del motivo, la `descripcion`, el orden de los elementos y la redacción de la skill son del plan y del run. Lo que el spec fija del informe es el contrato de una salida que leen el job, el informe final y una persona: nombre, medida, total, umbral, `decide` y número de elementos.
- [x] h. Derivable: FR-001 define el `total` y la `medida`; FR-002 dice qué es cotejar, con los casos que pasan el umbral de materialidad (código 2, `--dry-run`, sin referencia); FR-003 y FR-004, el motivo y el `cumple`; FR-005, la sesión sin medir; FR-006 y FR-007, los elementos de las tres skills.
- [x] i. Uso: «Uso, de fuera adentro» dice quién pide cada salida, cuántas veces y de qué tamaño. Para la salida de `cita cotejar` que consume la skill: 949 bytes por documento, medidos en esta sesión, una vez por documento delante. Para los umbrales: 4 108 bytes hoy, medidos en el último cierre versionado, y ~4,8 KB con los dos nuevos. Para el motivo: ~0,8 KB como mucho, acotado por las 12 sesiones del modo. Y dice cuándo deja de darse cada señal.
- [x] j. Umbral con control: FR-001 y FR-003 nombran el job en rojo; FR-006 y FR-023, su comprobación de `make ci`; SC-001 a SC-008 llevan **Control**; FR-061 fija las once filas de «Controles de umbral»; FR-064 pide la prueba, con mutantes, de que el control de FR-001 falla por encima de su umbral. SC-010 lo mide una persona.

## Notes

- Queda una marca [NEEDS CLARIFICATION], en FR-033: si el sondeo de la Aceptación humana tiene que contar las sesiones sin cotejo y medir el modo herramienta, o no cambia.
- En esta sesión no se ha ejecutado ningún test ni `make ci`. Se ejecutó `go run ./cmd/kitlegal cita cotejar` una vez, sin red ni grafo, para medir el tamaño de su sobre. Lo comprobado está en la viñeta «Lo comprobado en esta sesión» de Assumptions.
