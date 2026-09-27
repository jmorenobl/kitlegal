# Specification Analysis Report — H19 (009-h19-instalar-sin-clonar)

Análisis de solo lectura de `spec.md`, `plan.md` y `tasks.md` frente a `.specify/memory/constitution.md` 2.1.0.
Artefactos leídos: spec (criterios literales, historias, clarificaciones, FR/SC por sus referencias y trazabilidad), plan
(completo) y tasks (completo). Los cuerpos de FR-0xx individuales de spec.md se contrastaron por identificador y por la
tabla de trazabilidad, no línea a línea.

## Hallazgos

| ID | Categoría | Severidad | Ubicación | Resumen | Recomendación |
|----|-----------|-----------|-----------|---------|---------------|
| C1 | Inconsistencia | LOW | tasks.md T009 (`[P]`) frente a T005-T008 | T009 se marca `[P]` («no comparte fichero con sus vecinas») pero declara `internal/core/instalacion/discoEnMemoria_test.go`, que T005, T006, T007 y T008 también declaran. La sección de paralelismo lo contradice al decir que solo depende de T002, T003 y T005 | Quitar `[P]` de T009 o no declarar el fichero compartido; no afecta al bucle secuencial del workflow |
| C2 | Inconsistencia | LOW | tasks.md T012 frente a plan.md, «Tests existentes que cambian» (paso 5) | T012 declara `cmd/kitlegal/main.go` entre sus rutas, pero el plan dice que ese fichero «sin cambio de código» y que `cmd/kitlegal/main_test.go:128` compila sin cambios | Inofensivo (declarar de más no rompe el guardián); anotar «sin cambio» o retirarlo de la línea |
| C3 | Inconsistencia | LOW | spec.md «Criterios del hito, literales» (Release) frente a FR-095, plan y T021 | El literal del roadmap fija `make release` con `--skip=publish`; FR-095, la clarificación 3, el plan y T021 usan `--skip=publish,sign,sbom` | Ya resuelto por la clarificación de la sesión 2026-09-26 (auto, criterio a); dejar constancia de que el literal es el del roadmap y no se edita. Sin acción sobre las tareas |
| C4 | Tamaño de tarea | MEDIUM | tasks.md T018 | Una sola tarea reescribe `comandos.go`, `sincronia.go`, retira 5 rutas (`enlaces.go`, `enlaces_test.go`, dos `scripts/`, `instalar-skills.sh`), edita dos `SKILL.md`, regenera tablas, añade `TestOrdenesDeLasSkillsEmpotradas` y cambia tres ficheros de test grandes. En un bucle de una tarea por intento, con guardián de diff y `make ci`, es la que más probabilidad tiene de agotar el intento | Mantener (dividir dejaría `make ci` en rojo por los tests que se rompen juntos, como razona el plan); el ejecutor debe seguir el orden de «Tests existentes que cambian», paso 11. Sin cambio de artefactos |
| C5 | Cobertura | LOW | tasks.md T011 | Declara `internal/skills/instalacion_test.go` «solo si `make test-integration` lo exige»; T017 lo reescribe de todos modos. Ruta condicional en una tarea con guardián de diff | Aceptable: está declarada, y el guardián solo rechaza lo no declarado |
| C6 | Cobertura | LOW | spec.md FR-150 / SC-023 | Sin tareas por diseño (humano, tras fusionar) | Ninguna; conforme a la frontera humana (principio I, ADR 0018) |
| C7 | Dependencias | LOW | tasks.md, tabla de dependencias, T012 → T011 | T012 (firma con la versión) no usa `skills.go` ni `empotradas.go`; la dependencia solo ordena la secuencia del plan (pasos 4 y 5) | Sin acción: el orden es el del plan y es inocuo |

## Coverage Summary

Todos los FR de spec.md tienen al menos una tarea, contando los rangos («FR-040 a FR-043», «FR-090 a FR-097»,
«FR-100 a FR-107», «FR-110 a FR-115»…), que es como las tareas los citan. El barrido literal que no encuentra el
identificador suelto (FR-002, 003, 011, 012, 031, 032, 041, 042, 046, 053, 068, 071, 074, 075, 092-096, 101-106,
111-114) cae siempre dentro de un rango de una tarea.

| Grupo de requisitos | ¿Tarea? | Tareas | Notas |
|---|---|---|---|
| FR-001 a FR-004, FR-122 (empotrado, `go install`) | Sí | T011 (+T001) | SC-021 |
| FR-010 a FR-016 (verbos, ámbitos, `--dir`, sin red) | Sí | T004, T006, T013, T016 | |
| FR-020 a FR-028 (hosts, enlaces, copia, rutas) | Sí | T004-T006, T010, T015 | |
| FR-030 a FR-036 (manifiesto, skill no empotrada) | Sí | T003, T006 | fuzz en T003 |
| FR-040 a FR-048 (atomicidad, conflictos, dry-run) | Sí | T005-T007 | `TestFalloAMitadSeCompleta` |
| FR-050 a FR-054 (sobre, esquema) | Sí | T013 | `[datos]` justificada |
| FR-060 a FR-062, FR-065 a FR-069 (`list`, `doctor`) | Sí | T008, T013 | |
| FR-070 a FR-077 (aviso) | Sí | T002, T009, T014 | |
| FR-080 a FR-086 (skills sin `scripts/`, evals intactas) | Sí | T018, T019 | |
| FR-090 a FR-098 (`.goreleaser.yaml`, `PUBLISHER_TOKEN`, sin red) | Sí | T020, T021 | |
| FR-100 a FR-109 (`install.sh`) | Sí | T020 | verificación por copias momentáneas |
| FR-110 a FR-115 (`release.yml`) | Sí | T022 | no se ejecuta en el run |
| FR-120, FR-121 (snapshot en CI, `make help`) | Sí | T021, T022 | |
| FR-125 a FR-128 (`make install`, evals) | Sí | T017, T019 | veredicto de evals en el cierre |
| FR-130 a FR-133 (documentación) | Sí | T018, T023 | |
| FR-140 a FR-146 (DoD, datos, arnés, sin red) | Sí | T003, T010, T015, T016, T024 | |
| FR-150 | No (humano) | — | fuera del run, por diseño |
| SC-001 a SC-014, SC-017, SC-021, SC-022 | Sí | T001 (suite) + tareas de código | |
| SC-015, SC-016, SC-018, SC-019, SC-020 | Sí | T017, T019, T021, T022, T023, T024 | SC-019 lo mide el cierre |
| SC-023 | No (humano) | — | tras fusionar |

## Constitution Alignment Issues

Ninguno. El plan pasa las tablas de principios I-IX y de reglas de dependencia con justificación en *Complexity
Tracking* para lo nuevo (`internal/disco`, `internal/core/instalacion`, goreleaser, syft, cosign, las acciones de la
release, las dos tareas `[datos]` mixtas, el aviso sin propagar el error y el cambio de firma de `Arrancar`). Ninguna
tarea publica, etiqueta, usa `PUBLISHER_TOKEN`, graba datos ni toca `evals/`. Sin dependencias nuevas del módulo.

## Unmapped Tasks

Ninguna. T001 (suite) y T024 (cierre DoD) mapean a FR-140 a FR-145 y a la aceptación; T012 mapea a FR-073 y FR-091.

## Metrics

- Total de requisitos: 108 FR y 23 SC (131)
- Total de tareas: 24
- Cobertura (FR con ≥ 1 tarea, salvo los humanos por diseño): 107 de 108 (99 %); FR-150 es humano por diseño
- Ambigüedades: 0 (sin marcas de aclaración pendiente, TODO ni placeholders)
- Duplicados: 0
- Incidencias CRITICAL: 0 · HIGH: 0 · MEDIUM: 1 · LOW: 6

## Next Actions

- No hay CRITICAL ni HIGH: se puede proceder a `/speckit-implement` sin cambios de artefactos.
- C1 y C2 son erratas de marcado en `tasks.md` sin efecto en la ejecución; C3-C7 no piden acción.
- Riesgo operativo a vigilar en la ejecución: T018 (C4), por su tamaño; seguir el orden por fichero de
  «Tests existentes que cambian», paso 11.
