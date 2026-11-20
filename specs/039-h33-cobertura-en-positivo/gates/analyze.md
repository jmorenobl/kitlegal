# Análisis de consistencia · H33 (spec.md, plan.md, tasks.md)

Sesión de un paso, solo lectura sobre los artefactos del hito (este fichero es lo único escrito). No se ha ejecutado
`make ci` ni ningún test: lo que sigue es lectura cruzada de `spec.md`, `plan.md`, `tasks.md`, `contracts/` y la
constitución, más tres comprobaciones sobre el repositorio de hoy (nombradas en cada fila que las usa).

## Specification Analysis Report

| ID | Categoría | Severidad | Ubicación | Resumen | Recomendación |
|----|-----------|-----------|-----------|---------|---------------|
| C1 | Cobertura | LOW | spec.md FR-022; tasks.md T007, T008 | FR-022 lista el comentario del tope de `.github/workflows/evals.yml` entre la documentación; la tarea que lo hace es T007 (junto a FR-019) y T008 no lo nombra. Está cubierto, pero la traza de FR-022 apunta a T008 y T009 solamente. | Sin cambio necesario: T007 lo declara con FR-019. Si se quiere la traza completa, citar T007 en la cabecera de T008. |
| C2 | Cobertura | LOW | tasks.md «Controles de umbral»; spec.md FR-024, SC-004, SC-010 | FR-024, SC-004 y SC-010 no tienen tarea: los mide el cierre del workflow (job de evals) y el sandbox del run. Está dicho en el propio fichero (fila «lo mide el cierre del workflow; ninguna tarea lo ejecuta», «Estrategia: Cierre»). | Ninguna: es la lectura que fija el ADR 0032 y la constitución (el run no abre sesiones con modelo). |
| C3 | Cobertura | LOW | spec.md US1 y US4; tasks.md T001 | La suite de aceptación (T001) solo cubre US5 (lo instalado). US1 (evals de la frontera) y US4 (juicio sin modelo) no tienen guion e2e; los controles son tests de unidad en `make ci` (T006, T007). | Ninguna: el propio T001 lo razona (research D18: las claves nuevas de eval romperían `make ci` hasta T006). |
| I1 | Inconsistencia | LOW | tasks.md fases 3 a 5 | US2 aparece en dos fases (3 y 5) con US3 en medio, y US5 (P2) abre el fichero. Es el orden de dentro afuera del plan, no el de prioridad; no afecta a la ejecución. | Ninguna; ya lo explica «Organización». |
| I2 | Inconsistencia | LOW | tasks.md «Format» | La convención dice que `[datos]` es T002 a T006 «porque crean o modifican ficheros bajo `testdata/` o `schemas/`»; T007 crea `evals/**/*.yaml` y modifica `eleccion_test.go` (`goldenDeLaMatriz`) sin la etiqueta. No he podido comprobar si `goldenDeLaMatriz` está en `testdata/` o es una constante de Go; si fuera un fichero bajo `testdata/`, T007 necesitaría `[datos]`. | Comprobar al implementar T007 dónde vive `goldenDeLaMatriz`; si es un fichero de `testdata/`, etiquetar T007 `[datos]`. |
| U1 | Subespecificación | LOW | tasks.md T004 y T005 | Cada una toca 9 a 12 ficheros y su línea ocupa ~6 000 caracteres. Está justificado en cada tarea (por separado `make ci` queda en rojo) y la regla del guardián exige declarar todas las rutas. Riesgo práctico: una ruta olvidada en la línea es un rechazo del guardián, no un fallo de diseño. | Ninguna; al implementar, `export_test.go` aparece como «si un test externo lo necesita» y está declarado, así que se admite. |
| A1 | Ambigüedad | LOW | spec.md FR-010 / SC-005 | Los cuerpos (4481, 10 686, 10 756, 10 860, tope 11 000) son cifras previstas por el spec; esta sesión no las ha medido porque el cuerpo nuevo no existe. Las mide el escenario 6 del quickstart y el guion de T001. | Ninguna ahora; el control ya existe (`presupuestos`, guion de FR-025). |

Sin hallazgos CRITICAL ni HIGH.

## Comprobaciones hechas (con lo que he visto)

- **Recuento de FR-013**: «once respuestas» (FR-013 y `contracts/evals.md` §2) = 11 filas de la tabla de §2 (6 fuera de lo
  declarado, 5 de lo regulado aparte). Coincide con T006.
- **Mutantes del esquema (FR-002)**: enumerados 6 (sin `capacidad`, sin `alcance`, capacidad fuera del vocabulario, clave
  no declarada, sin motivo, motivo vacío); T003 pide los seis.
- **Cifras repetidas** (4481, 10 686, 10 756, 10 860, 15 013, 21 120, 8485): aparecen en spec, plan, tasks, research,
  quickstart y contratos sin discrepancia entre ellas (recuento con `grep -c`; el cálculo 15 013 s ≈ 250,2 min y
  21 120 s = 352 min es aritmética, no medida).
- **Orden T004 → T005 con `formas-compartidas`**: hoy `skills/consulta-juridica/SKILL.md` no tiene ninguna línea
  `⚠ NO CUBIERTO:` (`grep` vacío) y `formas-compartidas` compara solo las skills que llevan la misma etiqueta
  (`internal/app/skills_test.go`, `fallosDeFormasCompartidas`). Tras T004, con la plantilla nueva solo en
  `plazos-y-territorio`, no hay pareja que difiera: T004 no deja `make ci` en rojo por ese subtest. No hay contradicción de orden.
- **Defecto de la prosa (T005) sobre `plazos-y-territorio` (T004)**: el texto que T004 escribe en esa skill (FR-009)
  contiene «fuera de lo que kitlegal cubre» entre « », así que la comprobación que llega en T005 no la deja en rojo.
- **Punto de partida de FR-020**: `h32-capacidades.txtar` tiene hoy `plazos-y-territorio|⚠ NO CUBIERTO:|2` (línea 47) y
  `formas: 16 de 16` (línea 29), como parte el spec; T004 lo lleva a `|1`.
- Sin marcadores sin resolver (`TODO`, `TKTK`, `???`, `[NEEDS CLARIFICATION]`) en `spec.md`, `plan.md` ni `tasks.md`: las
  apariciones de `TODO` son la prohibición de escribirlos.

## Coverage Summary

| Requisito | ¿Tarea? | Tareas | Notas |
|---|---|---|---|
| FR-001 datos de la cobertura | sí | T003 | `TestLeerCobertura`, subtest `repositorio` |
| FR-002 esquema (6 casos) | sí | T003 | seis mutantes |
| FR-003 defectos de la sincronía | sí | T005 (materia, sin fichero, motivo en la prosa) | `TestRegionDeLoCubierto` |
| FR-004 plantilla única | sí | T001, T004, T005 | `formas-compartidas` y guion |
| FR-005 motivos | sí | T004 | referencia y vertical de ensayo |
| FR-006 generador | sí | T004 | `TestRegiones`, `TestRegionDeLaCobertura` |
| FR-007 región de lo cubierto | sí | T005 (guion T001) | `partirElCuerpo` en T005 |
| FR-008 `consulta-juridica` | sí | T005 | huella de H33 |
| FR-009 `plazos-y-territorio` | sí | T004 | declaración de la suite congelada |
| FR-010 tope del cuerpo | sí | T001, T004, T005 | 10 860 en el guion |
| FR-011 vertical de ensayo | sí | T004, T005 | `TestVerticalDeEnsayo` |
| FR-012 formato de eval | sí | T006 | `TestFormatoDeLaCobertura` |
| FR-013 juicio | sí | T006 | once respuestas |
| FR-014 umbrales | sí | T006 (descripción), T007 | cierre del workflow mide el resultado |
| FR-015 `tambien_la_resuelven` | sí | T006, T007 | eval 17 |
| FR-016 evals | sí | T007 | 01, 03, 04 |
| FR-017 regla «frontera en dos materias» | sí | T007 | |
| FR-018 lo grabado | sí | T002 (manifiesto), T007 | graba `grabar_datos` |
| FR-019 trabajo de evals | sí | T007 | comentario de `evals.yml` |
| FR-020 suites congeladas | sí | T004, T005 | los cuatro cambios |
| FR-021 tests actualizados | sí | T003 a T007 | |
| FR-022 documentación | sí | T007 (evals.yml), T008, T009 | véase C1 |
| FR-023 sin diff de lo generado | sí | T004, T005, T009 | `make skills-sync` |
| FR-024 medición del cierre | no (por diseño) | — | job de evals del workflow; véase C2 |
| FR-025 guion e2e | sí | T001 | |
| FR-026 sin sondeos ni modelo | sí | convenciones | sandbox; véase C2 |
| SC-001 a SC-003, SC-005 | sí | T001, T004, T005 | |
| SC-004 | no (por diseño) | — | lo mide el cierre |
| SC-006, SC-007 | sí | T007 | |
| SC-008, SC-009 | sí | T003 a T009 | |
| SC-010 | no (por diseño) | — | sandbox del run |

## Constitución

Sin conflictos. Comprobado contra el texto leído:

- **Tests primero (III) y gates**: T001 escribe la suite de aceptación, que queda congelada; cada tarea posterior trae su test.
- **Umbrales que se cumplen (ADR 0029)**: la tabla «Controles de umbral» da un control y un modo de verlo fallar a cada
  umbral numérico medible sin persona; los que mide el cierre (SC-004) están marcados como tales.
- **Modo desatendido**: ninguna tarea usa la red, graba, ni abre sesiones con modelo; la grabación del art. 59 va por
  `grabar_datos`, de una fuente con fila «Revisado» (2026-09-13).
- **Criterio de decisión autónoma (lo no especificado no se implementa)**: «Fuera de alcance» y las convenciones listan
  lo que no se toca (`Makefile`, `scripts/`, `cmd/`, `internal/core/`…); ninguna tarea añade comportamiento fuera de un FR.
- **Genericidad por materia (X)**: la regla «frontera en dos materias» se escribe sin nombrar norma ni materia; la
  `materia` de `cubierto` es el `vertical` de `data/normas.yaml`.

## Tareas sin requisito

Ninguna. T008 y T009 (pulido) trazan a FR-022, FR-023 y a la Definition of Done.

## Métricas

- Requisitos con trabajo construible: 26 FR y 7 SC con control construible (SC-001 a SC-003, SC-005 a SC-009); SC-004 y SC-010 los mide el cierre o el sandbox.
- Tareas: 9 (T001 a T009).
- Cobertura (requisitos con al menos una tarea): 24 de 26 FR (92 %); los dos restantes (FR-024, FR-026) están cubiertos por diseño fuera de las tareas.
- Ambigüedades: 1 (A1, LOW). Duplicaciones: 0. Incidencias CRITICAL: 0. HIGH: 0. MEDIUM: 0. LOW: 7.

## Próximas acciones

- Sin CRITICAL ni HIGH: se puede pasar a `/speckit-implement`.
- Único punto a vigilar al implementar: I2 (dónde vive `goldenDeLaMatriz`, para saber si T007 es `[datos]`).
- Esta sesión no ha ejecutado tests ni `make`: las cifras de cuerpos y el peor caso de 15 013 s las comprobarán los
  controles de las propias tareas, no este análisis.
