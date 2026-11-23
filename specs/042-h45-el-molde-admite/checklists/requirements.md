# Specification Quality Checklist: H45 · El molde admite un paquete con sus plazos y su norma, sin código (ADR 0062)

**Purpose**: Validate specification completeness and quality before proceeding to planning
**Created**: 2026-11-23
**Feature**: [spec.md](../spec.md)

## Content Quality

- [x] No implementation details (languages, frameworks, APIs) — el spec nombra funciones, tests, ficheros y constantes
  (`filasDeLosRecursosYPlazos`, `filasSinRegion`, `UnionDeGrabaciones`, `sesionesALaVezComoMucho`, los tests de §C.5)
  porque el hito los nombra como lo que cambia o como el control que lo hace cumplir (FR-002, FR-006, FR-015, FR-016,
  FR-021); no fija tipos de Go, firmas, cómo se parametriza la raíz de un test ni cómo se cuenta el sitio del cuerpo,
  que son del plan. Los textos de §C.1 y §C.2 son salida que ve quien escribe un paquete o el modelo.
- [x] Focused on user value and business needs — US1 (lo de hoy igual), US2 (un paquete entra solo con datos), US3 (quien
  escribe un paquete recibe el defecto), US4 (un paquete no rompe tests ajenos), US9 (la skill se activa ante otra
  materia).
- [x] Written for non-technical stakeholders — cada historia empieza por quién lo ve: quien usa las skills, quien escribe
  un paquete, el modelo que usa `normativa`, la persona que lee el cierre.
- [x] All mandatory sections completed — escenarios, requisitos, criterios de éxito, criterios literales del hito,
  contratos, supuestos y «Fuera de alcance».

## Requirement Completeness

- [x] No [NEEDS CLARIFICATION] markers remain — ninguno: las dos lecturas que el hito no escribe del todo (el sitio de
  `TestVerticalDeEnsayo` y el rojo de la descripción de hoy) se derivan del propio hito y van en «Assumptions» con su
  cálculo.
- [x] Requirements are testable and unambiguous — cada FR termina en «Comprobable:» con su escenario, su mutante o su
  control; los defectos nuevos van byte a byte en §C.1, lo movido en §C.2, los casos del control en §C.3 y el orden de
  las raíces en §C.4. US1 no congela la entrada ni lo esperado de los tests que FR-010, FR-014, FR-016, FR-017 y FR-023
  cambian: exige que, con los datos de `main`, la salida del binario y lo que comprueba cada caso sean los de `main`.
- [x] Success criteria are measurable — SC-001 a SC-011 dan recuentos y topes (0 diferencias, 1 referencia, 2 defectos,
  4 y 5 raíces, 12 sesiones, 352 min, 429/450/1 700/900 caracteres, 11 000 con 70 de sitio, 5 mutantes) y cada uno
  nombra la comprobación de `make ci` o del cierre que sale en rojo (criterio j); SC-004, la de la suite de aceptación
  que ejecuta los tests de §C.5 sobre la copia de US4.1.
- [x] Success criteria are technology-agnostic (no implementation details) — se miden sobre lo que ve quien usa las
  skills o escribe un paquete (referencias, cuerpos, defectos, descripción, informe); los controles nombran el test que
  los hace cumplir, como pide el criterio j.
- [x] All acceptance scenarios are defined — US1 a US9 cubren la aceptación y cada control del hito (mapa en «Criterios
  del hito (literales)»).
- [x] Edge cases are identified — tipo con normas con y sin `vertical`, tipo solo con `unidades`, fila con norma sin
  `vertical` o fuera de `data/normas.yaml`, paquete sin la región, carpetas de `evidencias/` incompletas, y lo manipulado
  a mano con la regla genérica.
- [x] Scope is clearly bounded — «Fuera de alcance» con la lista cerrada del hito, el punto 6 del ADR, el ADR mismo y lo
  que el hito no declara (defecto de un paquete sin la región, control de materias en las descripciones).
- [x] Dependencies and assumptions identified — «Assumptions» (sitio del cuerpo, rojo de la descripción, nombre de las
  carpetas de hito, la copia de US4.1, lo de «Antes de lanzar») y las dependencias (H36 hecho; H42 detrás).

## Feature Readiness

- [x] All functional requirements have clear acceptance criteria — FR-001 a FR-027, cada uno con su escenario, su
  mutante o su SC.
- [x] User scenarios cover primary flows — lo de hoy (US1), el molde (US2, US3), los tests (US4), `normativa` (US5), el
  laboratorio (US6), las grabaciones (US7), el trabajo de evals (US8), la descripción (US9).
- [x] Feature meets measurable outcomes defined in Success Criteria — SC-001, SC-002, SC-004 y SC-010 son la aceptación
  literal del hito.
- [x] No implementation details leak into specification — ver la primera nota de «Content Quality».

## Notes

- Los requisitos de otros hitos se citan sin guion («FR 020 y SC 011 de H30») para que el precheck no los tome por
  requisitos de este spec sin definir.
- «Quién consume cada salida, cuánto y hasta cuándo» responde al criterio i: las referencias de un paquete crecen con
  los tipos y las filas de su paquete, no con los de todos ni con el uso (FR-002, FR-005); `con_herramientas.md` gana
  176 caracteres, leída una vez por conversación; la descripción, 2 caracteres; los defectos, una línea por tipo mal
  declarado, que desaparece al corregirlo.
