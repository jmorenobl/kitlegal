# Specification Analysis Report — H5.1 · Avisos de vigencia en las evals

**Modo desatendido.** Análisis de consistencia entre `spec.md`, `plan.md` y `tasks.md` (con apoyo de `research.md`,
`data-model.md`, `contracts/`, `quickstart.md` y `checklists/requirements.md`), tras `/speckit-tasks`. Solo lectura;
sin remediaciones diferidas: cualquier hallazgo accionable se resuelve o se declara aquí mismo.

## Hallazgos

| ID | Categoría | Severidad | Ubicación(es) | Resumen | Recomendación |
|----|----------|----------|-------------|---------|----------------|
| U1 | Underspecification | LOW | tasks.md T005 (lista de subtests); spec.md US2 escenario 7 | El escenario 7 de US2 exige que, sin ninguna forma fija en la respuesta, los dos avisos esperados (`derogada` y `vigencia-agotada`) queden ausentes **en el orden de la eval**. Ninguno de los nombres de subtest de `TestJuzgar` en T005 (`aviso-ausente`, `avisos-en-el-orden-de-la-eval`, …) dice explícitamente que cubre el caso de *ambos* ausentes a la vez por falta total de forma fija; `avisos-en-el-orden-de-la-eval` podría cubrir solo el caso en que ambos se encuentran, no el caso en que ambos faltan. | Ninguna: es una ambigüedad de nombre, no un hueco de cobertura — `aviso-ausente` generalizado a dos códigos ausentes ya ejercita el mismo camino de `repartirAvisos`, y el orden de ausentes lo fija el propio bucle sobre `eval.Avisos`. No se detiene el hito por esto. |
| N1 | Nota | — | spec.md *Fuera de alcance*, línea con «FR-062»; plan.md líneas 90, 92 (Constitution Check, principio III) | Dos referencias cruzadas a la especificación de H5 (`specs/006-h5-skill-boe-legislacion/spec.md`) — `FR-062` (regla «positivas») y varias `FR-0XX de H5` (072, 074-077, 083) — se resuelven correctamente contra ese fichero: no es una inconsistencia, solo se deja constancia de que la trazabilidad de este hito depende de un documento externo al directorio `specs/007-...`. | Ninguna acción; verificado. |

## Coverage Summary Table

| Requirement Key | Has Task? | Task IDs | Notes |
|-----------------|-----------|----------|-------|
| FR-001–FR-005 (forma fija en SKILL.md) | Sí | T011 (T013 valida) | — |
| FR-010–FR-012 (etiquetas en el binario) | Sí | T001 | — |
| FR-013 (esquema ↔ CodigosDeAviso) | Sí | T002, T004 | — |
| FR-014 (SKILL.md ↔ forma fija) | Sí | T002, T011 | — |
| FR-020–FR-023 (formato `avisos`) | Sí | T003, T004 | T003 es `[datos]` con pausa |
| FR-030–FR-035 (juicio) | Sí | T002, T005 | — |
| FR-040–FR-043 (informe) | Sí | T006 | — |
| FR-050–FR-057 (eval de la norma derogada) | Sí | T007–T010 | T008 es `[datos]` con pausa |
| FR-060 (make ci en verde) | Sí | Cada tarea; T013 | — |
| FR-061 (documentación) | Sí | T012 | — |
| FR-070–FR-073 (aceptación) | Sí | T013 (cuerpo), T014 | T014 es `[plataforma]` |
| SC-001–SC-009 | Sí | Ver tabla «Trazabilidad» de tasks.md | Todas mapeadas a ≥1 tarea |

Cobertura: **100 %** de los FR (001–073, con los huecos de numeración intencionales entre bloques temáticos) y de los
SC (001–009) tienen al menos una tarea; **0** tareas sin requisito mapeado (`tasks.md`, tabla «Trazabilidad»).

## Constitution Alignment Issues

Ninguno. `plan.md` recorre los nueve principios y las reglas de dependencia con veredicto «✅ Cumple» en todos salvo el
principio III, que lleva «✅ Cumple con justificación» y su desviación (ningún `testscript` nuevo, el primer paso no es
el e2e) está registrada y justificada en *Complexity Tracking* con la alternativa más simple rechazada y por qué. No
hay ningún requisito, tarea o decisión de diseño que contradiga un MUST de la constitución (`.specify/memory/constitution.md`).

## Unmapped Tasks

Ninguna. Las 14 tareas (T001–T014) están citadas en la tabla «Trazabilidad: requisito → tarea» y en «Obligaciones que
el plan trasladó a este fichero»; T012 (documentación) y T013–T014 (cierre y plataforma) no llevan historia de usuario
por diseño (spec, sección *Requirements*, no describe documentación ni el ritual como historias), consistente con la
leyenda del formato de tareas.

## Metrics

- **Total Requirements (FR)**: 73 numerados (FR-001 a FR-073, en bloques temáticos con huecos de numeración intencionales)
- **Total Success Criteria (SC)**: 9 (SC-001 a SC-009)
- **Total Tasks**: 14 (T001–T014)
- **Coverage %**: 100 % (FR y SC con ≥1 tarea)
- **Ambiguity Count**: 1 (U1, severidad LOW, no accionable)
- **Duplication Count**: 0
- **Critical Issues Count**: 0
- **High Issues Count**: 0

## Next Actions

No hay hallazgos CRITICAL ni HIGH. El único hallazgo (U1) es LOW y no bloquea: es una observación sobre la
legibilidad del nombre de un subtest, no un hueco de cobertura, porque `repartirAvisos` (research D3) procesa cada
aviso esperado de la eval independientemente y en su orden, así que un caso con un aviso ausente ya ejercita la misma
ruta que dos. No se recomienda ninguna edición de `spec.md`, `plan.md` ni `tasks.md`.

**Se puede proceder a `/speckit-implement`** con `tasks.md` tal como está.

## Extension Hooks

`.specify/extensions.yml` no existe en la raíz del proyecto: no hay hooks `before_analyze` ni `after_analyze` que ejecutar.
