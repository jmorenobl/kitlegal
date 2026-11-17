# Specification Analysis Report — H41.3 · Por qué `normativa` no se activa con todas las skills instaladas

Artefactos leídos: `spec.md` (requisitos FR/SC), `tasks.md` (íntegro), `plan.md`, `data-model.md`, `research.md` y
`contracts/tabla.md` solo por búsquedas dirigidas (`grep -n`), y la constitución por las reglas citadas en `tasks.md` y
`CLAUDE.md`. Solo lectura: no se ha modificado ningún fichero salvo este informe. No se ha ejecutado `make ci` ni nada
con modelo: lo que sigue es análisis estático de los artefactos y comprobaciones de existencia en el repositorio.

## Hallazgos

| ID | Categoría | Severidad | Ubicación | Resumen | Recomendación |
|----|-----------|-----------|-----------|---------|---------------|
| F1 | Inconsistencia | LOW | spec.md FR-020/FR-022 · tasks.md T009 · research.md D11 · contracts/tabla.md:26 | El spec dice que la instalación de una fila sin constancia es `producto` si sus `skills_instaladas` «son las cuatro del producto» (FR-022). T009, `data-model.md` y `contracts/tabla.md` la derivan de «alguna skill además de la de la familia» (research D11 rechaza expresamente la lista fija de cuatro nombres). Coinciden en los casos del hito (cuatro skills, solo `normativa`, sin la clave); difieren en un conjunto intermedio (p. ej. dos skills), que el spec no define. | Ninguna acción bloqueante: la regla de D11 es deliberada y T009 la fija con su mutante «`producto` solo con cuatro nombres fijos». Al implementar, prevalece la regla de los contratos; el redactado de FR-022 es una simplificación del caso de hoy. |
| F2 | Ambigüedad menor | LOW | spec.md FR-021 · tasks.md T001/T010 | FR-021 escribe la fila `diagnóstico` (con tilde); los directorios de los sondeos fijos son `h41-3-diagnostico` (sin tilde). Es solo un nombre de fila en prosa; ningún valor medido depende de él. | Sin acción. |
| F3 | Cobertura (informativa) | LOW | tasks.md «Requisitos sin tarea propia» | FR-030, FR-031, FR-034, SC-008 y SC-009 no tienen tarea: son de la persona (US4) o del cierre del workflow. FR-017, FR-032 y FR-033 se guardan por las convenciones y el guardián de diff, no por una tarea propia. `tasks.md` lo declara, con quién los cubre. | Sin acción: coherente con ADR 0018/0020 y con el spec (FR-030, FR-031). |

Sin hallazgos CRITICAL ni HIGH. Búsqueda de marcadores sin resolver (`NEEDS CLARIFICATION`, `TODO`, `TKTK`, `???`) en
`spec.md`, `plan.md` y `tasks.md`: 0 coincidencias.

## Resumen de cobertura

Mapeo obtenido de las referencias `FR-`/`SC-` de cada línea de tarea de `tasks.md`.

| Requisito | ¿Tarea? | Tareas | Notas |
|-----------|---------|--------|-------|
| FR-001 | Sí | T003, T011 | Variante y sus cuatro textos (417+420+426+427 = 1690 runas, ≤ 1700) |
| FR-002 | Sí | T002, T003 | Comprobación en `make ci`, con `PrimeraEtiquetaXML` exportada |
| FR-010 | Sí | T006, T011 | `INSTALACION=producto-descripciones-cortas`, solo con `normativa-producto` |
| FR-011 | Sí | T004, T006 | Composición del `SKILL.md` |
| FR-012 | Sí | T006, T007, T011 | Constancia en `sondeo.json`; sin opciones, byte a byte como hoy |
| FR-013 | Sí | T007, T008, T011 | `PRESUPUESTO` y su cadena hasta `go test` |
| FR-014 | Sí | T005, T007 | Sin `PRESUPUESTO`, sin variable |
| FR-015 | Sí | T005, T006 | Sin opciones, lo de hoy |
| FR-016 | Sí | T006, T007 | Errores de argumentos |
| FR-017 | Parcial (por diseño) | T008 + convenciones | Ninguna tarea lanza el arnés con modelo |
| FR-018 | Sí | T005, T006, T007 | Sustituto de `claude` sin red ni modelo |
| FR-020 | Sí | T001, T009, T010, T011 | Columnas y orden |
| FR-021 | Sí | T001, T009, T010 | Las cuatro filas |
| FR-022 | Sí | T001, T009, T010 | Ver F1 |
| FR-023 | Sí | T009 | La tabla de H39 no cambia |
| FR-024 | Sí | T001, T010 | Sondeos fijos |
| FR-030 | No (fuera del run) | — | US4, la persona |
| FR-031 | Parcial | T011 (documentación) | El ADR y la tabla en la PR, fuera del run |
| FR-032 | Sí | T009 + convenciones | Suites congeladas intactas |
| FR-033 | Sí | T001, T003 + convenciones | Ninguna skill cambia |
| FR-034 | No (cierre del workflow) | — | Lo mide el job de evals |
| SC-001 a SC-007 | Sí | T001 a T010 | Según «Controles de umbral» de `tasks.md` (Notes), cada uno con su test y mutantes |
| SC-008, SC-009 | No (cierre y persona) | — | Como FR-034 y FR-030 |

## Alineación con la constitución

Sin conflictos detectados en lo que se ha podido contrastar con `tasks.md` y `CLAUDE.md`: la primera tarea escribe la
suite de aceptación congelada (T001), cada tarea trae su test y `make ci` en verde, los datos de T010 no se graban de la
red, ninguna skill ni `data/` ni umbral cambia, no hay sesión con modelo dentro del run, y la frontera humana (medir y
escribir el ADR) queda fuera del run. No he releído el texto íntegro de la constitución en esta sesión.

## Tareas sin requisito

Ninguna. T002 sirve a FR-002 (dependencia de T003); T011 es documentación (DoD §1).

## Comprobaciones en el repositorio (vistas en esta sesión)

- Existen `internal/evals/laboratorio_tabla.go` (`compararClaves` en `:910`, `algunoConLaConstancia` en `:377`),
  `internal/skills/frontmatter.go` (`etiquetaXML` en `:71`, `defectosDeLaEtiqueta` en `:481`), `CONTRIBUTING.md:1660`
  y `scripts/evals-sondeo-llavero.sh:5` con «sus nueve», como dicen T002, T008, T009 y T011.
- `laboratorio/` y `internal/evals/testdata/laboratorio/familias/` existen; el directorio nuevo
  `producto-descripciones-cortas/` y `familias/instalacion/` aún no, como corresponde antes de implementar.
- Las rutas que declaran las tareas son coherentes con las dependencias de «Orden de las fases» (T003/T004/T006 sobre
  `laboratorio_descripciones.go`; T006/T007/T009 sobre `laboratorio_tabla.go`), por lo que ninguna lleva `[P]`.

## Métricas

- Requisitos con tarea o con cobertura declarada: 21 FR (FR-001, 002, 010 a 018, 020 a 024, 030 a 034) y 9 SC (SC-001 a SC-009).
- Tareas: 11 (T001 a T011).
- FR con al menos una tarea: 17 de 21 (81 %); los 4 restantes (FR-030, FR-031 parcial, FR-034, y FR-017 parcial) son de la persona o del cierre, declarados así en `tasks.md`.
- Ambigüedades: 1 (F2, menor). Duplicados: 0. Inconsistencias: 1 (F1, LOW). Críticos: 0.

## Siguientes pasos

Sin CRITICAL ni HIGH: se puede pasar a `/speckit-implement`. F1 y F2 no requieren cambio; si se quiere, al implementar T009
sigue la regla de `contracts/tabla.md:26`, no el redactado de FR-022.
