## Specification Analysis Report — H41.2 (`specs/036-h41-2-por-qu-e`)

Alcance: `spec.md`, `plan.md`, `tasks.md`, contrastados con `contracts/tabla.md` (§3, §4, §6), la constitución
(vía el «Constitution Check» del plan) y el código de hoy (`internal/evals/sesion.go`, `laboratorio_tabla.go`,
`scripts/evals-laboratorio.sh`, `scripts/evals-sondeo-llavero.sh`, `Makefile`). Solo lectura: no se ha ejecutado
`make ci` ni ningún test en esta sesión; lo que sigue sale de leer los ficheros, y los cálculos de Wilson de la
spec y del contrato se han comprobado a mano (0/40 → 8,8 %; 1/40 → 0,4–12,9 %; 1/4, 3/4, 1/2 y 0/2 de la tabla de
§4 cuadran).

| ID | Category | Severity | Location(s) | Summary | Recommendation |
|----|----------|----------|-------------|---------|----------------|
| F1 | Inconsistency | LOW | spec.md FR-003/FR-004; tasks.md T002, T004; `sesion.go:606-624` vs `laboratorio_tabla.go:710-721` | Habrá dos lectores de los `tool_use` de `Skill`: `leerAssistant` (da `Activada`) y `leerLosUsos`/`skillDelUso` (da la lista). El primero falla con `input.skill` vacío; el segundo lo devuelve como `""`. En un transcript así la sesión es ilegible (el error sale antes), de modo que hoy no hay divergencia visible, pero FR-003 depende de que ambos coincidan. | Opcional: un caso de `TestActivacionesDeLosUsos` o `TestInforme` que fije que un `Skill` sin `input.skill` deja la sesión ilegible y sin lista (T004 ya lo cubre para la sesión ilegible en general). No bloquea. |
| F2 | Underspecification | LOW | spec.md FR-013/FR-030; plan.md (`EVAL`); Makefile | El nombre `EVAL` (singular) se parece a `EVALS` de `make evals-sondeo`. `make` ignora una variable desconocida: `EVALS=04` en la orden de FR-030 abriría `SESIONES/4` de cada una de las cuatro evals y la tabla las agregaría sin avisar. El control de `make ci` no puede verlo; solo lo vería la persona. | Que la lectura de `quickstart.md` §8 diga que, antes de leer la tabla, se compruebe en el detalle que las 40 sesiones de cada fila son `04-lgt-prescripcion-NN`. No toca el run. |
| F3 | Inconsistency (transitoria) | LOW | tasks.md T006, T007, T009; `evals-laboratorio.sh:15-18` | Entre T006/T007 y T009 el guion exige ocho o nueve argumentos mientras su cabecera y la del llavero siguen diciendo «siete». T009 las corrige y `make ci` no las comprueba, así que no rompe nada. | Ninguna; si T006/T007 lo ven cómodo, corregir la cabecera en la tarea que cambia el recuento (ya declaran el fichero). |
| F4 | Coverage (observación) | LOW | tasks.md T006, T007, T008 | T008 reúne código de la tabla, dos juegos de fixtures y su test; T006 y T007 son largas. Sin riesgo de contenido, pero una tarea que no sale va a cuarentena (workflow) y las tres pesan más que las demás. | Ninguna acción; vigilar el presupuesto de tiempo en el run. |

Sin hallazgos CRITICAL ni HIGH. Sin duplicados entre requisitos, sin placeholders (`TODO`, `???`, `<placeholder>`) y sin
adjetivos vagos sin criterio medible: cada FR/SC lleva su cifra o su control.

### Coverage Summary

| Requirement Key | Has Task? | Task IDs | Notes |
|-----------------|-----------|----------|-------|
| FR-001 (lista en `informe.json`) | Sí | T002, T004 | regla compartida + campo `activaciones` |
| FR-002 (párrafo en `informe.md`) | Sí | T004 | `textoDeLasActivaciones` |
| FR-003 (otra skill o ninguna en `sin_activar`) | Sí | T003, T004 | sesiones `otra` y `ninguna` |
| FR-004 (nada más cambia) | Sí | T003, T004 | aserciones antes del cambio; `TestInformeDeLaSuiteCompletaNoCambia` |
| FR-005 (sesión ilegible / informe anterior) | Sí | T004 | `omitzero`, `TestInforme`, `TestJuzgarSinLaLista` |
| FR-006 (control en `make ci`) | Sí | T002, T003, T004 | diez sesiones fijas |
| FR-010 (opción de instalación) | Sí | T006 | `INSTALACION=producto` |
| FR-011 (peticiones inválidas) | Sí | T006, T007 | 0 sesiones, nada construido |
| FR-012 (constancia en `sondeo.json`) | Sí | T005, T006 | `skills_instaladas` |
| FR-013 (una sola eval) | Sí | T007 | `EVAL=<nn>`; `SESIONES` > 0 |
| FR-014 (lo demás del arnés) | Sí | T006, T007 | sin cambiar comprobaciones; ninguna tarea abre sesiones |
| FR-015 (control del arnés) | Sí | T005, T006, T007 | sustituto de `claude` |
| FR-020 (columnas de la tabla) | Sí | T008 | |
| FR-021 (`n` = 0) | Sí | T001, T008 | fila de `h41` sin sesiones terminadas |
| FR-022 (sin constancia, igual que hoy) | Sí | T001, T002, T008 | tablas de H40 y H41 |
| FR-023 (tabla de H39 igual) | Sí | T002, T008 | `TestTablaDelLaboratorio` |
| FR-024 (control del guion) | Sí | T001, T008 | |
| FR-030 (sondeo de la persona) | No (por diseño) | — | fuera del run (ADR 0032); T006/T007 dejan las cuatro órdenes válidas |
| FR-031 (tabla y ADR en la PR) | No (por diseño) | — | `gates/supuestos.md` lo deja pendiente |
| FR-032 (suites congeladas) | Sí (transversal) | todas | ninguna tarea declara sus rutas |
| FR-033 (evals del cierre) | No (por diseño) | — | lo mide el job del cierre; el informe da las listas (T004) |
| SC-001 a SC-006, SC-009 | Sí | T002–T008 | cada uno con su test en «Controles de umbral» |
| SC-007 | No (por diseño) | — | job de evals del cierre |
| SC-008 | No (por diseño) | — | capa 3: la persona |

### Constitution Alignment Issues

Ninguno. Principio III: la primera tarea (T001) escribe la suite de aceptación antes de cualquier código, y las
aserciones de «lo que no cambia» del informe se escriben y pasan antes de tocar `informe.go` (T003 → T004). Principios
IV a VII y reglas 1 a 7: `internal/evals` no nombra `os.Std*` ni `os.Exit`, ninguna dependencia nueva, el binario no
cambia. Principios IX y X: sin municipio ni materia en código; `INSTALACION` y `EVAL` valen para cualquier familia.
Frontera humana y ADR 0032: el run no lanza el sondeo ni escribe en `evidencias/`.

### Unmapped Tasks

Ninguna. Cada tarea traza a al menos un FR/SC (T009 → documentación de lo que deja de ser cierto, DoD §1).

### Metrics

- Total Requirements: 30 (21 FR + 9 SC)
- Total Tasks: 9
- Coverage % (requisitos con ≥ 1 tarea): 25 de 30 (83 %): 17 FR y 7 SC con tarea propia, más FR-032 como guarda
  transversal. Los 5 restantes (FR-030, FR-031, FR-033, SC-007, SC-008) son de la persona o del cierre del workflow
  por diseño, declarados como tales en spec, plan y tasks. Sobre lo que puede construir el run: 100 %.
- Ambiguity Count: 0 bloqueantes (F2 es una trampa de uso, no una ambigüedad del texto)
- Duplication Count: 0
- Critical Issues Count: 0

### Next Actions

- Sin CRITICAL ni HIGH: se puede seguir con `/speckit-implement`.
- F1 a F4 son opcionales y ninguna exige editar `spec.md`, `plan.md` ni `tasks.md` antes de implementar.
- No se ha medido nada con `make ci` en esta sesión; el primer control real es el de T001 (rojo por una aserción) y el
  de cada tarea posterior.
