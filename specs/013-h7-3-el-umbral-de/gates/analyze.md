## Specification Analysis Report

Hito H7.3 (`specs/013-h7-3-el-umbral-de/`). Artefactos leídos: `spec.md` (FR-001 a FR-099, SC-001 a SC-010, US1 a US5), `plan.md`, `tasks.md` (T001 a T015), `research.md` (D1 a D22, en lo que toca a los hallazgos) y la constitución 2.7.0. Análisis de solo lectura; ningún fichero del hito se ha modificado salvo este informe.

| ID | Category | Severity | Location(s) | Summary | Recommendation |
|----|----------|----------|-------------|---------|----------------|
| F1 | Inconsistency | LOW | spec.md:L262 (FR-061), L371; plan.md «Decisiones»; tasks.md T009 | FR-061 pide preparar cada sesión «con `TestPrepararSesion`», y T009 retira esa entrada (`TestPrepararSesion`, `TestPlanDeSesiones`, `TestInformeDelJob`). El plan lo resuelve llamando a `PrepararSesion` en proceso, con las mismas evals y `UnionDeGrabaciones()`: la intención se cumple, la letra no. | No bloquea. Implementar como el plan; en el informe final, decir que «la misma preparación» es `PrepararSesion` en proceso. |
| F2 | Coverage / robustness | LOW | spec.md:L240 (FR-034); research.md D11; tasks.md T008 | FR-034 exige que el segundo disparo no deje «ninguna comprobación cancelada, elija lo que elija». D11 se apoya en el flujo del cierre (los únicos disparos son la apertura y la etiqueta) y en que `TestDefinicionDelJob` fija `cancel-in-progress: false`. Un tercer disparo sobre el mismo commit con uno en curso y otro pendiente cancelaría al pendiente (D11 lo reconoce, V6). | Aceptable por el umbral de materialidad (sin vía real en el run) y ya razonado en D11. Sin tarea nueva. |
| F3 | Inconsistency | LOW | spec.md:L326-327 (tabla «Uso») frente a plan.md L259-260 | El spec da ≤ 6 KB para las sesiones sin medir y ≤ 2 KB para los reintentos; el plan resume «≤ 22 KB en el peor caso» para sin medir, reintentos y duración. Son cotas de lo mismo con otro alcance (`informe.json` + `informe.md`), sin decir cuál es cuál. | Cosmético. Si el plan se toca, alinear la cifra con las de contracts/informe-del-job.md §7. |
| F4 | Underspecification | LOW | spec.md:L265 (FR-064); tasks.md T011, T012 | «Nada fuera del directorio temporal, salvo las cachés de compilación de Go»: en una máquina con el módulo aún sin descargar, `go install` escribe también en la caché de módulos, que no es de compilación. Los tests con sustitutos no lo ven. | No bloquea (el supuesto es un Go ya usado en el repositorio). Si la revisión final quiere afinarlo, decir «cachés de Go». |
| F5 | Underspecification | LOW | spec.md:L263-264 (FR-062, FR-063); tasks.md T011 | El sondeo comprueba argumentos y credencial antes de abrir sesiones, pero no que `claude` esté en el `PATH`; sin él, cada sesión sale como «no terminó» con su motivo. | Sin caso propio (constitución, proporcionalidad): el fallo ya es visible en la salida, con su motivo por sesión. |
| F6 | Ambiguity | LOW | tasks.md T013 (verificación) | La comprobación con la `SKILL.md` de `main` cita números de línea de v0.1.2 (73, 75, 88…, 191) y `git show main:`. Es exacta hoy y deja de serlo si `main` cambia esa skill antes de que corra la tarea. | Que la tarea lea los números de la propia salida del control, no de la lista; los 14 párrafos y la fecha son la comprobación, no las líneas. |
| F7 | Duplication | LOW | spec.md FR-044/FR-066; FR-061/FR-065; «Clarifications» | Requisitos del sondeo que repiten los del job (límite de uso) y aclaraciones que reescriben lo que ya dicen los FR. Es deliberado: el sondeo reutiliza la regla del job. | Ninguna. |
| F8 | Efficiency (tasks) | LOW | tasks.md T005, T011, T013 | Tres tareas de tamaño grande (informe con umbrales; argumentos + credencial + árbol + entorno + `sondear`; skill + comprobación de la prosa). Cada una es una rebanada vertical con `make ci` en verde y sus mutantes; el riesgo es gastar un intento del run, no un hueco de cobertura. | Sin acción: partirlas dejaría `make ci` en rojo entre tareas, como el propio tasks.md argumenta. |

Sin hallazgos CRITICAL ni HIGH. Nada conflictúa con la constitución 2.7.0.

### Coverage Summary Table

Requisitos con trabajo construible: 59 FR y 10 SC. Los que no tienen tarea la tienen por diseño (FR-098 y SC-002 los mide el cierre del workflow o la persona; ADR 0029).

| Requirement Key | Has Task? | Task IDs | Notes |
|---|---|---|---|
| FR-001 a FR-006 (`umbrales`, el de Sonnet 5 decide, el de Haiku informativo, duración, `[]` en `legal-core`) | Sí | T005 | Control: `TestUmbralesDelInforme` con 3/51, 2/51, 901 s y 900 s; fila FR-099 |
| FR-007 (regla por serie intacta) | Sí | T002, T004, T005 | Restricción, sin código propio |
| FR-008 (no rebajar umbrales) | Sí | T005, T015 y batería | Comprobado en el cierre |
| FR-010 a FR-017 (`SKILL.md` v0.1.3) | Sí | T013 | Control: subprueba `prosa-de-la-skill` y `TestProsaDeLaSkill` |
| FR-018, FR-097 (CHANGELOG) | Sí | T014 | |
| FR-020, FR-023, FR-024 (tercera familia, esquema, aplicación) | Sí | T001, T002 | Única tarea `[datos]` |
| FR-021, FR-022, FR-095 (calibrado 35 y 10; bloques y formas) | Sí | T002 | Control en `TestEvalsDelRepositorio` |
| FR-030 a FR-032, FR-036, FR-037 (repartidor, garantías, aislamiento) | Sí | T006, T007, T009 | Probado sin strace, sin sudo y en macOS con el `claude` sustituto |
| FR-033, FR-040 a FR-044 (reintentos, sin medir, motivo propio, parada) | Sí | T003, T004, T007 | Control: `TestInformeConSesionesSinMedir` |
| FR-034, FR-035, FR-052 (una tanda por commit, tope de trabajo) | Sí | T008 | Control: `TestDefinicionDelJob`; ver F2 |
| FR-050, FR-051 (duración y su umbral) | Sí | T005, T007, T008, T009 | Fila FR-099 (900 s) |
| FR-060 a FR-068 (sondeo) | Sí | T008, T010, T011, T012 | Control: `TestSondear`, `TestGuionDelSondeo`, `TestPrepararElArbolDelSondeo`; ver F1, F4, F5 |
| FR-070 (escenario del quickstart) | Sí | T015 | T015 comprueba que sus órdenes nombran lo que existe; lo ejecuta la persona |
| FR-080 (no editar H7.2, sin ADR) | Sí | T015 | `git diff --quiet main -- specs/012-…` |
| FR-090 a FR-096 (controles de `make ci`) | Sí | Todas; T005, T007, T008, T010 a T012 | |
| FR-098 (job de evals en la propuesta) | No (cierre del workflow) | — | Por diseño; T008 y T009 dejan el job listo |
| FR-099 (dos filas de «Controles de umbral») | Sí | Plan ya escrito; T015 lo comprueba | Las dos filas están en plan.md L234-235; la de Haiku no tiene fila, como pide el spec |
| SC-001 (job de cierre) | Sí (controles) | T004, T005, T008, T009 | Se mide en el job; control: el job sale en rojo con `fallo` |
| SC-002 (quickstart §6) | No (medida sin control) | — | Por diseño (ADR 0029) |
| SC-003 a SC-010 | Sí | T002, T013, T005, T003/T004/T007, T006-T008, T010-T012, T013-T015 | Cada uno con su control en `make ci` |

### Constitution Alignment Issues

Ninguno.

- «Aceptación e2e: no aplica» (principio III): el workflow lo admite con motivo (`workflow.yml`: «o "no aplica" con motivo»), el plan lo declara en *Complexity Tracking* y `cmd/kitlegal` no enlaza `internal/evals`; no hay tarea `[aceptacion]`.
- Una sola tarea `[datos]` (T001) con esquema, lista y tipo en el mismo diff: excepción declarada, con precedente en H7.2 T001; no graba nada de ninguna fuente y no hay fila nueva que revisar en `docs/SOURCES.md`.
- ADR 0029 (controles de umbral): las dos filas obligatorias del cierre están, y cada umbral medido en `make ci` tiene fila con su test. El umbral de Haiku no tiene fila, como dice FR-099.
- Sin dependencias nuevas, sin cambios en el binario, sin `//nolint` ni `t.Skip` previstos, sin telemetría.

### Unmapped Tasks

Ninguna. T014 (documentación) y T015 (cierre de la Definition of Done) mapean a FR-018, FR-097, FR-070, FR-080, FR-090, FR-099 y SC-010.

### Comprobaciones cruzadas

- Cifras de la evidencia coherentes entre spec, plan y tasks: 10 de 51 (03, 06, 13, 14, 15, 19: 1+1+2+2+3+1); 35 de 93 de H7.1 (suma por eval de FR-021); 9 de las 10 en las evals 03, 06, 13, 14 y 15 (SC-002); 51 = 17 evals × 3 y 30 de Haiku; 94 sesiones con la prueba de red.
- Peor caso de FR-035: boe-legislacion 485 + ⌈94/4⌉ × 272 = 7 013 s; legal-core 485 + 19 × 272 = 5 653 s, contando la prueba de red aunque esa skill no la lleve (research D13, deliberado y del lado seguro); ambos por debajo de `timeout-minutes: 120` = 7 200 s. El margen de boe-legislacion es de 187 s: una eval nueva que añada una tanda de cuatro sesiones hará fallar `TestDefinicionDelJob` hasta subir el tope, que es lo que pide el control.
- Orden de las tareas: T001 → T002 (lista antes de los umbrales), T003 → T004 → T005 (los límites antes del total del umbral), T006 → T007 → T008 → T009 (repartidor antes de la definición y del punto de entrada), T010 → T012 (sondeo), T013 (skill) después de sus evals (Definition of Done §1.10), T014 y T015 al final. Sin contradicciones.
- Rutas declaradas: los ficheros de test de cada tarea son `x_test.go` de un `x.go` declarado o están declarados (`sustitutos_test.go`, `sondeo_integracion_test.go`, `job_test.go`, `conjunto_test.go`); las únicas referencias vivas a lo que T009 retira (`plan.tsv`, `TestPlanDeSesiones`, `TestInformeDelJob`, `TestPrepararSesion`) están en `internal/evals/job_test.go` y `scripts/evals.sh`, ambos declarados en T009.
- `make help` y `README.md`: el hito solo añade `evals-sondeo` a `make help`; el spec excluye documentar el sondeo fuera de `make help`, del quickstart y de `CHANGELOG.md`.

### Metrics

- Requisitos: 69 (59 FR + 10 SC)
- Tareas: 15 (1 `[datos]`, 0 `[aceptacion]`, 0 `[P]`)
- Cobertura: 67 de 69 con al menos una tarea (97,1 %); los otros dos (FR-098, SC-002) están cubiertos por el cierre del workflow o por la persona, por diseño; 100 % contando esos
- Ambigüedades: 2 (F5, F6)
- Duplicidades: 1 (F7, deliberada)
- Hallazgos CRITICAL: 0 · HIGH: 0 · MEDIUM: 0 · LOW: 8

### Next Actions

- No hay CRITICAL ni HIGH: se puede pasar a `/speckit-implement`.
- Los ocho hallazgos LOW no piden cambios en `spec.md`, `plan.md` ni `tasks.md` antes de implementar. F1 y F6 conviene tenerlos presentes durante T009/T013 y en el informe final; F2, F4 y F5 son residuos que el umbral de materialidad deja fuera y que ya constan en research (D11) o en el spec.
