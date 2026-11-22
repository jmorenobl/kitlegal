## Specification Analysis Report — H36 · El motor de cómputo descrito por datos

Artefactos leídos: `spec.md` (FR-001 a FR-025, SC-001 a SC-009, US1 a US6, §C), `plan.md` (Diseño, Orden y piezas
indivisibles, Controles, Mutantes, Controles de umbral), `tasks.md` (T001 a T013) y, para contrastar, `research.md`,
`contracts/`, `gates/supuestos.md`, `internal/core/plazos/tipos.go`, `.golangci.yml` y la constitución (solo el pasaje de
umbrales y mutantes, ADR 0029). No se ha ejecutado ningún test ni `make`: lo medido es solo lectura y `grep` del
repositorio. Hook `before_analyze`/`after_analyze` de `git` (opcional, commit): no se ejecuta, es solo de lectura.

| ID | Category | Severity | Location(s) | Summary | Recommendation |
|----|----------|----------|-------------|---------|----------------|
| F1 | Inconsistency | MEDIUM | tasks.md T007 (línea 148); `internal/core/plazos/tipos.go:38` | T007 dice «`tieneDuracion` pasa a mirar la duración del tipo», pero esa función no existe: la que decide por código es `fijaSuDuracion(codigo string)` (`tipos.go:38`, con `TipoSilencio` y `TipoGenerico`). Quien implemente T007 no la encuentra por el nombre. | Nombrar `fijaSuDuracion` en T007 (o decir «la que hoy decide por código»). |
| F2 | Coverage / lint | MEDIUM | tasks.md T004; `tipos.go:240,246` | T004 quita las reglas de `LeerTipos` que dependen de los códigos («fija su duración y no la lleva», «no fija…»), que son los únicos usos de `fijaSuDuracion` (grep: solo `tipos.go:240` y `:246`). Con `unused` en `golangci-lint` (`default: standard`) la función queda sin uso al terminar T004 y `make ci` sale en rojo, y T004 exige `make ci` en verde. T007 la menciona como si siguiera viva. | T004 debe retirar `fijaSuDuracion` (o T007 el único que la use); decirlo en la línea de T004. |
| F3 | Inconsistency | MEDIUM | spec.md FR-017 (l. 328-341) vs tasks.md T002, T004, T013 | FR-017 dice que los tests que cambian «MUST ser solo estos» y T013 comprueba «los tests de H9 cambiados, solo los de FR-017». Pero T002 cambia `internal/evals/grabacion_test.go`, T004 `internal/skills/regiones_test.go` y T009 la lista de `TestGrabacionesSinSolape`; ninguno está en FR-017. Las tres desviaciones están declaradas (research D10, `gates/supuestos.md`, «Convenciones» de tasks.md), pero la comprobación de T013 tal como está escrita las marcaría. | En T013, comprobar contra FR-017 más las tres desviaciones declaradas (D10 y supuestos). |
| F4 | Underspecification | LOW | tasks.md T005, T010, T011 | Tres tareas dicen «sin código de producto»: sus tests nacen en verde y el control se ve fallar solo con un mutante (ADR 0029), no con rojo primero. Es coherente con la constitución (umbrales con mutante), pero no hay ningún paso en rojo natural. | Sin acción: queda cubierto por los mutantes de T005, T010, T011 y T013. |
| F5 | Ambiguity | LOW | tasks.md T001–T013 | Cada tarea es un único párrafo de 1.500 a 4.000 caracteres; difícil de revisar y de cotejar con el guardián de rutas. | Estilo; no bloquea. |
| F6 | Coverage | LOW | tasks.md «Trazabilidad» | FR-025 y SC-008 no tienen tarea; las cubre el cierre del workflow (job de evals). Declarado, consistente con ADR 0018/0031 y con el plan («Autonomía»). | Sin acción. |

**Coverage Summary Table**

| Requirement Key | Has Task? | Task IDs | Notes |
|-----------------|-----------|----------|-------|
| FR-001, FR-002, FR-003 | Sí | T004 | |
| FR-004 a FR-008 | Sí | T007 (T010 con el civil) | FR-007: ver F1 |
| FR-009, FR-010 | Sí | T008 | |
| FR-011 | Sí | T001, T004, T007 | |
| FR-012 | Sí | T001, T003, T005 | |
| FR-013 | Sí | T004 | |
| FR-014 | Sí | T009, T010 | |
| FR-015, FR-016 | Sí | T002, T009 | |
| FR-017 | Sí | T003, T004, T007, T008, T009; T013 | ver F3 |
| FR-018, FR-020 | Sí | T013 | |
| FR-019 | Sí | T012 | |
| FR-021 | Sí | T001 | |
| FR-022 | Sí | T006, T007 | |
| FR-023 | Sí | T006, T010 | |
| FR-024 | Sí | T011 | |
| FR-025 | Cierre del workflow | — | F6 |
| SC-001 a SC-007, SC-009 | Sí | ver «Trazabilidad» de tasks.md | cada fila de «Controles de umbral» tiene tarea y mutante |
| SC-008 | Cierre del workflow | — | F6 |

**Constitution Alignment Issues**: ninguno. Cada control con umbral tiene su mutante registrado en T013 (ADR 0029);
la fuente única (BOE consolidado) tiene su fila revisada; ningún ADR, orden, código de salida ni dependencia nuevos; las
sesiones con modelo y la red quedan fuera de las tareas.

**Unmapped Tasks**: ninguna. T002 es fundacional (sin historia), T012 y T013 son de pulido.

**Orden y dependencias**: sin contradicciones. T001 antes de cualquier cambio de producto; T002 antes de T009; T003 antes de
T004 a T008; T004 antes de T005 a T011; T006 antes de T007, T008, T010, T011; T009 antes de T010. Placeholders: ninguno
(el único «TODO» de tasks.md es la prohibición de dejarlos).

**Metrics**

- Total Requirements: 25 FR + 9 SC (34)
- Total Tasks: 13
- Coverage % (requisitos con ≥ 1 tarea): 32 de 34 por tarea (94 %); los 2 restantes (FR-025, SC-008) los cubre el cierre del workflow, declarado
- Ambiguity Count: 1 (F5)
- Duplication Count: 0
- Critical Issues Count: 0

## Next Actions

- No hay CRITICAL. Se puede pasar a `/speckit-implement`; conviene corregir antes F1, F2 y F3 en `tasks.md` (nombre de la
  función de T007, retirada de `fijaSuDuracion` en T004 para no dejar `make ci` en rojo por `unused`, y la comprobación de
  T013 con las desviaciones declaradas).
- Ninguna de las tres cambia el alcance ni el spec; son ajustes de redacción de `tasks.md`.
