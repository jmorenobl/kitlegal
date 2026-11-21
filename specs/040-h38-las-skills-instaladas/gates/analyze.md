## Specification Analysis Report — H38 · Las skills instaladas se actualizan con el binario

Alcance leído: `spec.md` (FR, SC, US, criterios literales, supuestos, fuera de alcance), `tasks.md` (completo: líneas de
tarea, DoD, controles de umbral, trazabilidad, dependencias), `plan.md` (tabla de constitución y de controles, por
`grep`), `gates/supuestos.md`, y `internal/arch_test.go` y `internal/cache` para comprobar una premisa del plan.
**No leído en su totalidad**: `.specify/memory/constitution.md` (la alineación se juzga por la tabla de constitución del
propio plan, plan.md:59-73), `research.md`, `data-model.md`, `contracts/*` y el cuerpo de las viñetas de T001 a T007
más allá de las líneas citadas. Hallazgos limitados a lo leído.

### Comprobación ejecutada en esta sesión

- `go list -f '{{.Imports}}' ./internal/cache ./internal/core ./internal/core/schema`: las importaciones de esos tres
  paquetes dentro del módulo son solo `internal/core` y `internal/core/schema`, y ninguno importa `net*` ni
  `internal/httpx` directamente. La subprueba «sin red» de `TestArquitectura` sigue solo aristas del módulo
  (`internal/arch_test.go:655-690`), y `app/aviso.go` ya está en `ficherosSinRed` (`arch_test.go:720`): la premisa del
  plan (plan.md:73) —que al importar `internal/cache` el aviso sigue pasando— se sostiene. `go list -deps` de
  `internal/cache` sí lista `net` transitivamente (vía terceros), que esa subprueba no recorre por diseño; no es un
  cambio de este hito.

### Hallazgos

| ID | Categoría | Gravedad | Ubicación | Resumen | Recomendación |
|----|-----------|----------|-----------|---------|---------------|
| F1 | Inconsistencia | LOW | tasks.md:258 (T006) vs. «Format» y «Parallel Opportunities» | T006 se describe como «solo tests» y se marca `[P]` con T003–T005, pero su línea declara también `internal/core/instalacion/reinstalacion.go` (producción), que T002 crea. No hay choque real: T006 depende de T002 y el orden secuencial la pone tras T005. | Aclarar en T006 si toca producción (p. ej. ganchos para forzar intercalados) o quitar `reinstalacion.go` de su lista. |
| F2 | Cobertura | LOW | spec.md SC-004; gates/supuestos.md; tasks.md T001/T006 | El guion `candado.txtar` lanza dos procesos «sin garantía de que se solapen» (supuesto del plan); la afirmación es «como mucho una reinstala», que se cumple también sin solape. El control real de la concurrencia es `TestReinstalacionConcurrente` (T006), con mutantes. SC-004 ya remite a «una comprobación de `make ci`»; coherente. | Sin cambio; solo conviene que el informe final no cite el guion como prueba de la exclusión. |
| F3 | Terminología | LOW | spec.md FR-003 (a) y FR-005 («FR 070 de H19», «FR 044 de H19») | Se escribe «FR 070/044» sin guion en las referencias a H19, frente a «FR-0nn» en el resto. | Unificar a `FR-070` / `FR-044` (de H19). |
| F4 | Duplicación | LOW | spec.md FR-003 (a) y FR-004 | «La versión del binario tiene forma SemVer» aparece en FR-003 (a) y de nuevo en FR-004 («si … la del binario no tiene forma SemVer, no reinstala»). | Dejar la condición solo en FR-004 y que FR-003 (b) la remita. |
| F5 | Cobertura | LOW | spec.md FR-017; tasks.md T007 | FR-017 (texto de `README.md` y `CHANGELOG.md`) no tiene control mecánico propio; solo `make ci` en verde (SC-007). Es documentación; lo habitual en el proyecto. | Aceptable; si hay un guardián de documentación en `make ci`, citarlo en T007. No lo he comprobado. |
| F6 | Cobertura | LOW | spec.md FR-016; tasks.md trazabilidad | FR-016 («no cambia `install`, `list`, `doctor`, el manifiesto ni los hosts») se cubre por ausencia de cambios y por los e2e de H19 existentes; ninguna tarea añade un control nuevo. | Aceptable: es una restricción, no una funcionalidad. |

Sin hallazgos CRITICAL ni HIGH. Sin marcadores sin resolver (`NEEDS`, `TODO`, `TBD`, `???`) en `spec.md`, `plan.md` ni
`tasks.md` (grep ejecutado).

### Coverage Summary

| Requisito | ¿Tarea? | Tareas | Notas |
|---|---|---|---|
| FR-001, FR-002 | Sí | T001, T004, T005 | permiso del kernel (T004) y composición (T005) |
| FR-003 a FR-007 | Sí | T001, T002, T005 | dominio en T002 |
| FR-008 | Sí | T001, T004 | |
| FR-009 | Sí | T001, T002, T004, T005 | |
| FR-010 | Sí | T001, T002, T003, T005, T006 | candado en `internal/cache` |
| FR-011 | Sí | T001, T002, T006 | |
| FR-012, FR-013 | Sí | T001, T002, T003, T005 | |
| FR-014 | Sí | T001, T002 | |
| FR-015 | Sí | T001, T004 | |
| FR-016 | Sí (por ausencia) | T003 | ninguna tarea toca `install`/`list`/`doctor` |
| FR-017 | Sí | T007 | |
| FR-018 | Sí | T005 | tarea `[datos]` indivisible con la composición |
| FR-019 | Sí | T002, T006 | |
| SC-001 a SC-007 | Sí | ver «Controles de umbral» de tasks.md | cada fila con su «lo ve fallar» |
| US1 a US5 | Sí | T001 (guiones), T002–T006, T007 | US2, US3 y US4 sin tarea propia, declarado en tasks.md |

Escenarios de US1.4–US1.7, US2.2–US2.4, US3.4, US3.5 y US4.5 aparecen citados en las líneas de T001 y T006.

### Constitution Alignment Issues

Ninguno detectado en lo leído. El plan trae su tabla de principios (plan.md:59-73) y la DoD de tasks.md los trata punto a
punto (no aplica: schemas, SOURCES, evals; sí: `make ci`, unitarios, cobertura de `internal/core/**`, e2e,
`CHANGELOG.md`). No he releído la constitución completa en esta sesión.

### Unmapped Tasks

Ninguna: T001 (aceptación), T002–T006 y T007 citan FR/SC/US.

### Metrics

- Requisitos: 19 FR + 7 SC = 26 (más 5 historias)
- Tareas: 7
- Cobertura (requisitos con ≥1 tarea): 100 %
- Ambigüedades (marcadores sin resolver): 0
- Duplicaciones: 1 (F4, LOW)
- Hallazgos CRITICAL: 0 · HIGH: 0 · MEDIUM: 0 · LOW: 6

### Next Actions

- Sin CRITICAL ni HIGH: se puede pasar a `/speckit-implement`.
- Los LOW (F1 a F6) no bloquean; F1 (T006: producción o solo tests) es el único que puede cambiar qué rutas toca la tarea
  y conviene resolverlo al implementar T006 declarando la ruta que de verdad modifica.
