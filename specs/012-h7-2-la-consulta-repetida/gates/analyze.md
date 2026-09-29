# Specification Analysis Report — H7.2 (012-h7-2-la-consulta-repetida)

Análisis de solo lectura de `spec.md`, `plan.md` y `tasks.md` (con `research.md`, `data-model.md`, `contracts/` y
`quickstart.md` como apoyo), contrastado con la constitución 2.6.0 y con el repositorio (símbolos, ficheros y cifras que
las tareas nombran).

## Hallazgos

| ID | Categoría | Severidad | Ubicación | Resumen | Recomendación |
|----|-----------|-----------|-----------|---------|---------------|
| E1 | Cobertura | LOW | spec FR-085, SC-004; tasks T003, T007 | `expresiones-en-los-bloques` (T003) lee los bloques con `boe articulo`, que solo da la redacción vigente del bloque `a1-30`; la original (que FR-085 nombra) solo entra en la comprobación cuando T007 pone la derivada en las grabaciones de un grafo previo. Entre T003 y T007 la comprobación es parcial. T007 ya declara que `expresiones-en-los-bloques` «cuenta ya la redacción original», así que el cierre es correcto. | Sin cambios; opcional: T003 nombra que la redacción original se cubre en T007. |
| E2 | Inconsistencia | LOW | tasks T011; CHANGELOG.md líneas 124 y 130; quickstart §4 | `CHANGELOG.md` nombra hoy la ruta de la eval retirada y la de su grafo previo. FR-020 solo exige que no las nombre código ni tests, y el `git grep` de quickstart §4 se limita a `*.go`, `*.txtar`, `evals/`, `testdata/` y `scripts/`, así que la nueva entrada de T011 puede seguir nombrándolas sin que nada lo detecte. No incumple el spec. | T011 puede describir lo retirado sin la ruta literal; no hace falta control mecánico. |
| E3 | Ambigüedad | LOW | spec FR-053, SC-001; tasks T005 | El recuento por modelo cuenta las sesiones «juzgadas (no ilegibles)» (T005), mientras SC-001 razona sobre 51 y 30. Con una sesión ilegible o sin terminar, el denominador sería menor. Ese estado ya hace fallar el job por sus reglas (ADR 0016), así que no pasa el umbral de materialidad. | Sin cambios. |

No hay hallazgos CRITICAL ni HIGH.

## Tabla de cobertura (requisitos → tareas)

| Requisito | ¿Tiene tarea? | Tareas | Notas |
|---|---|---|---|
| FR-001 a FR-004 (la eval nueva) | Sí | T007 | FR-004 lo declara el informe desde `hallazgos` (mecanismo de H7.1). |
| FR-010 a FR-014 (derivada y control de derivaciones) | Sí | T006, T008 | FR-014: sin grabación nueva, dicho en T006. |
| FR-020, FR-021 (lo retirado) | Sí | T007, T008 | |
| FR-030 (19 evals, 10 deciden) | Sí | T007 | |
| FR-040 a FR-047 (`SKILL.md` v0.1.2) | Sí | T009 | FR-043 también en T003 (`expresiones-de-la-skill`). |
| FR-048, FR-086 (`CHANGELOG.md`) | Sí | T011 | |
| FR-050 a FR-055 (la lista y su juicio) | Sí | T001 a T005, T011 | Las 38 expresiones (22 + 16) coinciden con el contrato. |
| FR-056 (frontera de `internal/evals`) | Sí | T002, T010, T012 | |
| FR-060 a FR-062 (escenario del quickstart) | Sí | T010, T012 | El escenario está en quickstart §6; ninguna tarea lo ejecuta (FR-062). |
| FR-070 (no editar H7 ni H7.1) | Sí | T012 | |
| FR-080 a FR-085 (controles) | Sí | todas, T003, T006, T008 | FR-084: el reparto suma 35 (34 + 1) y el informe de H7.1 tiene 93 respuestas (verificado). |
| FR-087 (job de evals) | Cierre del workflow | — | No es una tarea, por diseño. |
| SC-001 a SC-008 | Sí | T003 a T012 y el job | SC-001 y SC-002 se miden fuera de `make ci`, por diseño. |

## Alineación con la constitución

Sin conflictos.

- **III (tests primero)**: «Aceptación e2e: no aplica» está justificada, porque el binario no cambia y `cmd/kitlegal` no enlaza `internal/evals`. Consta en Complexity Tracking, y cada pieza entra con tests en `make ci`.
- **VIII (skills primero)**: `SKILL.md` va detrás de sus evals y de su juicio (T003, T004, T007 → T009; DoD §1.10).
- **Criterio de uso y proporcionalidad (ADR 0028)**: el spec y el plan traen «Uso, de fuera adentro» con bytes y apagado de cada señal, y ninguna tarea añade casos para estados sin vía real.
- **I, IV, V, VI, VII, IX**: sin dependencias nuevas, sin cambios de binario, de grafo ni de códigos de salida, y nada particularizado por municipio.
- **Reglas de dependencia R1-R6**: sin cambios.

## Tareas sin requisito

Ninguna. T012 es el cierre de la Definition of Done y está trazada a FR-056, FR-062, FR-070 y FR-080.

## Comprobaciones contra el repositorio

- Los símbolos que las tareas reutilizan existen: `entrePalabrasDeAviso`, `formasDeHallazgo`, `contieneComoPalabra`, `ExtraerHallazgos`, `MotivoSinTerminar`, `exigirBanderas`, `FicheroMalFormado`, `TestPrepararSesion`, `ConsultasNecesarias`, `UnionDeGrabaciones`, `EtiquetasDeAviso`, `EtiquetasDeHallazgo`, `formaEscrita`, `grafosPreviosDeLasEvals`, `derivadasDelE2E`, `compruebaElGrafoPrevio`, `ficheroDeLaConsultaRepetida`.
- Las grabaciones de H4 de `BOE-A-2017-12902` (índice, metadatos y los bloques `a1-30` y `da-3`) están en `internal/source/boe/testdata/`.
- `specs/011-h7-1-graph-check-acotado/gates/evals/boe-legislacion.json` está versionado y trae 93 respuestas.
- Los textos que `SKILL.md` v0.1.1 enseña a escribir (bloques `text`) no contienen ninguna expresión de la lista, así que `expresiones-de-la-skill` de T003 pasa antes de T009.
- Los nombres retirados aparecen hoy en `formato_test.go`, `consultas_test.go`, `juzgar_test.go`, `grafo_test.go` y `preparar_test.go` (por `ficheroDeLaConsultaRepetida`), todos declarados en T007 y T008.
- Los casos versionados de `internal/evals/testdata/sesiones/informe/` son entradas, no informes esperados, así que T005 no toca `testdata/`.
- El orden de tareas es coherente: T008 no deja `make ci` en rojo y T006 va antes de T007.

## Métricas

- Requisitos: 48 (40 FR y 8 SC).
- Tareas: 12 (3 `[datos]`, 0 `[aceptacion]`).
- Cobertura (requisitos con al menos una tarea o cierre explícito): 100 %.
- Ambigüedades: 1 (E3, LOW).
- Duplicaciones: 0.
- Incidencias críticas: 0.
- Hallazgos: 3, todos LOW.

## Siguientes pasos

No hay CRITICAL ni HIGH: se puede proceder a `/speckit-implement`. Los tres hallazgos LOW no cambian el orden ni el alcance de las tareas y no requieren edición previa; E2 es solo una precaución para la redacción de T011.
