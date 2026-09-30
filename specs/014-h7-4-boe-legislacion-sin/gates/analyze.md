# Specification Analysis Report — H7.4 (`014-h7-4-boe-legislacion-sin`)

Análisis de solo lectura de `spec.md`, `plan.md` y `tasks.md`, con `research.md`, `data-model.md`, `contracts/` y `quickstart.md` como apoyo, contra `.specify/memory/constitution.md` 2.8.0 y contra el estado del repositorio en `014-h7-4-boe-legislacion-sin`. Los hooks `before_analyze` y `after_analyze` no existen en `.specify/extensions.yml`.

## Resultado

Sin hallazgos CRÍTICOS ni ALTOS. El conjunto es coherente: cada FR y cada SC de este spec tiene una tarea o un control declarado, las cifras cuadran entre las tres piezas y no hay conflicto con la constitución. Quedan tres observaciones MEDIAS (riesgo de ejecución, no defectos de diseño) y cuatro BAJAS de redacción.

## Findings

| ID | Categoría | Severidad | Ubicación | Resumen | Recomendación |
|----|-----------|-----------|-----------|---------|---------------|
| E1 | Cobertura / riesgo | MEDIUM | spec FR-003, FR-004; tasks T002, T004, T010 | La `description` de v0.1.4 se amplía para activarse «siempre que la respuesta dependa de lo que dice una norma» (C1), mientras `legal-core` §6 delega en `boe-legislacion` el texto de los artículos. Que una sesión de `legal-core` que active `boe-legislacion` no pase (FR-003) solo lo mide el job del cierre; en `make ci` hay tests de `Juzgar` con sesiones sintéticas, pero ninguna tarea comprueba la `description` contra las preguntas de las tres evals de `legal-core`. Las tres evals de hoy no piden contenido de norma (01, 02 y 03), así que el riesgo es bajo. | Ninguna acción antes de implementar. Si el job de `legal-core` sale rojo en el cierre, el arreglo va a la `description` y al cuerpo de `boe-legislacion` (FR-002, FR-003), no a las evals de `legal-core` (fuera de alcance). |
| E2 | Ejecución | MEDIUM | tasks T010 | T010 junta la skill, la lista entera (87 expresiones), el calibrado sobre tres informes, cuatro subpruebas y tests nuevos y ajustes en nueve ficheros de test. Está declarada como excepción (con la lista entera, la prosa de v0.1.3 da 9 defectos, research V19) y trae sus mutantes, pero es la tarea con más superficie del hito y la que más probablemente agote intentos. | Sin cambio: partirla dejaría `make ci` en rojo entre las dos mitades (research V19). El workflow ya la trata como excepción declarada; la revisión final debe mirarla primero. |
| E3 | Cobertura de umbral | MEDIUM | spec SC-001; tasks T006, T008, T010 | Los cuatro umbrales que deciden se comprueban en `make ci` con sesiones sintéticas (T006) y se miden de verdad solo en el job del cierre (FR-102), que no es una tarea. Es la forma prevista por el ADR 0029 y la tabla «Controles de umbral → tareas» lo refleja, pero un umbral que no se cumple con el modelo real (por ejemplo `sin_activar`, 2 de 51 en la línea de base) solo aparece tras el run. | Sin cambio: es el diseño del ADR 0029. El informe final debe dar los cuatro como «comprobado por su control» solo si el job del cierre los mide cumplidos. |
| T1 | Terminología | LOW | spec FR-031 («Son tres» formas fijas); tasks T001 y plan («las dos formas definitivas» de `formas_fijas`) | El spec cuenta tres formas fijas (la línea `⚠ REDACCIÓN MODIFICADA:`, los avisos de vigencia y la regla 7); la lista y `formas_fijas` llevan dos, porque la de los avisos la quita el código con las expresiones de `ExtraerAvisos`. Tasks y plan lo dicen (T003, punto 2) pero no lo enlazan con el «tres» del spec. | Sin cambio necesario; basta con leer T003 junto a FR-031. |
| T2 | Terminología | LOW | spec FR-043 y FR-054 (96 sesiones); plan V13 y tasks T008 (97 sesiones) | El spec da 96 sesiones del plan; el cálculo del tope de 122 min usa 97, con la sesión de la prueba de red (research V13). Sin contradicción, pero el spec no nombra esa sesión al hablar de las 96. | Sin cambio; el spec ya excluye la prueba de red de los recuentos (H7.2 FR 053) y la incluye en la duración (H7.3 FR 050). |
| C1 | Cobertura | LOW | spec FR-002, FR-020 | Las exigencias de research (causa de la activación, trazado de cada cambio de `SKILL.md`) no tienen tarea propia: se cumplen con `research.md` ya escrito, «Cambios de `SKILL.md` trazados a la causa» del plan y T010, que las cita. | Sin cambio; las comprueba el juez del plan, no el bucle de tareas. |
| C2 | Cobertura | LOW | spec FR-092 y FR-102 | Sin tarea por diseño: FR-092 está ya escrito en el plan y T013 lo comprueba; FR-102 es el cierre del workflow, no una tarea. | Sin cambio. |

## Coverage Summary Table

| Requisito | ¿Tiene tarea? | Tareas | Notas |
|---|---|---|---|
| FR-001, FR-002 (activación) | Sí | T010 | C1 trazado en research |
| FR-003, FR-004 (`legal-core` no activa `boe-legislacion`) | Sí | T002, T004, T010 | ver E1 |
| FR-010, FR-011, FR-012 (clases A y B) | Sí | T003, T010 | |
| FR-020 a FR-026 (`SKILL.md` v0.1.4) | Sí | T010 | FR-023 también T004 y T009 |
| FR-027, FR-101 (CHANGELOG) | Sí | T012 | |
| FR-030 a FR-037 (la lista por clases) | Sí | T001, T003, T006, T009, T010 | |
| FR-040 a FR-048 (umbrales) | Sí | T006 (T008 para FR-043) | |
| FR-050 a FR-055 (eval 20) | Sí | T002, T004, T008, T009 | |
| FR-060 a FR-062 (respuesta a la pregunta) | Sí | T004, T005, T006 | |
| FR-070 a FR-072 (una tanda) | Sí | T007, T008, T013 | sin tocar `scripts/workflow/` |
| FR-080 a FR-082 (sondeo) | Sí | T006, T011 | |
| FR-090 (sin editar H7.1 a H7.3) | Sí | T010, T013 | |
| FR-091 | Sí | todas, T013 | |
| FR-092 | Plan | plan «Controles de umbral»; T013 | ver C2 |
| FR-093 a FR-100 | Sí | T003, T004, T005, T006, T007, T008, T009, T010, T011 | |
| FR-102 | Workflow | cierre tras la revisión final | ver C2 |
| SC-001 | Control + cierre | T006, T008, T010; job del cierre | ver E3 |
| SC-002 a SC-012 | Sí | T003 a T013 | cada uno con su test en `make ci` |

## Constitution Alignment Issues

Ninguno.

- Principio III («cada hito empieza por el test e2e»): el plan declara «Aceptación e2e: no aplica» y lo registra en *Complexity Tracking* con su motivo (el binario no cambia y `cmd/kitlegal` no enlaza `internal/evals`). Tiene precedente en H7.2 y H7.3.
- Principio II: la clase B y `redaccion_no_leida` con umbral 0 lo refuerzan; ningún umbral se rebaja (FR-047, ADR 0029).
- Principio VIII y Definition of Done §1.10: las evals y la lista (T001 a T004, T009) van antes que `SKILL.md` (T010).
- ADR 0029: las cuatro filas de FR-092 están en el plan y todas las de «Controles de umbral» tienen tarea y test que las ve fallar.
- Guardián de diff: las tareas que tocan `schemas/` o `testdata/` (T001, T002, T005, T009) llevan `[datos]`. Los casos de `internal/evals/testdata/sesiones/informe/` son solo entradas: las salidas esperadas están en el código de los tests, así que T004 y T006 no los modifican. El resto de rutas declaradas coincide con las que existen (`redacciones.go` y `tanda.go` son nuevos). `.github/workflows/evals.yml` está declarado en T008 y T012, y no lo protege el guardián en modo tarea.

## Unmapped Tasks

Ninguna. T012 (documentación) y T013 (cierre de la Definition of Done) no llevan historia, por diseño.

## Metrics

- Total de requisitos: 72 (60 FR + 12 SC).
- Total de tareas: 13 (4 `[datos]`, ninguna `[aceptacion]`, ninguna `[P]`).
- Cobertura (requisitos con al menos una tarea, o con control en el plan o en el cierre del workflow, por diseño): 100 %. FR-092 y FR-102 no tienen tarea (C2).
- Ambigüedades: 0. Placeholders sin resolver: 0.
- Duplicados: 0.
- Inconsistencias: 2 de terminología (T1, T2), sin efecto sobre la ejecución.
- Críticos: 0. Altos: 0. Medios: 3. Bajos: 4.

## Next Actions

- Sin CRÍTICOS: se puede pasar a `/speckit-implement` sin cambiar `spec.md`, `plan.md` ni `tasks.md`.
- Vigilar en la revisión final: T010 (E2) y la activación con el modelo que decide (E1, E3), que solo se mide de verdad en el job del cierre.
