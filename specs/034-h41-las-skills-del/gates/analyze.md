# Specification Analysis Report — H41 · Las skills del producto mandan leer sus referencias

Artefactos: `spec.md` (606 líneas), `plan.md` (294), `tasks.md` (252), más `research.md`, `data-model.md` y `contracts/`
como apoyo. Constitución: `.specify/memory/constitution.md`. Análisis de solo lectura; no se ha ejecutado `make ci`
(sandbox de un paso, sin tiempo ni necesidad para ello en un análisis de artefactos): lo que sigue es lectura cruzada y
las comprobaciones puntuales que se indican.

## Comprobaciones hechas en esta sesión (con su resultado)

- `git diff --stat v0.9.0 HEAD -- skills/` sin salida: `skills/` es hoy la de `v0.9.0` (premisa de FR-020 y FR-010).
- Ficheros por skill: `normativa` 9, `plazos-y-territorio` 5, `jurisprudencia` 3; `laboratorio/normativa/normativa-control`
  9. Coinciden con los 17 de la tabla de huellas de T002 (9 + 5 + 3).
- Cuerpos hoy (`awk … | wc -m`): `normativa` 10 929, `jurisprudencia` 10 842, `plazos-y-territorio` 10 930. Coinciden con
  `contracts/remisiones.md` §1 y `research.md` V3.
- Las evals nombradas por T002/FR-030 existen en `evals/normativa` (04, 05, 15, 22), `evals/plazos-y-territorio` (02, 05,
  10, 11) y `evals/jurisprudencia` (01, 03, 04, 06).
- Existen los ficheros y tests que las tareas dicen tocar o dejar en verde: `internal/evals/laboratorio_*.go`,
  `TestEntregaDelHito`, `TestSkillsDelRepositorio`, `TestVerticalDeEnsayo`, `TestArnesDeNormativa`,
  `TestArnesDelLaboratorio`, `TestTablaDeNormativa`, `TestTablaDelLaboratorio`, `TestReglaTestigo`,
  `TestReglaTestigoDeNormativa`, `TestOrdenesDeBash`, `TestRecetasDelSondeo`, `TestServidorDelLaboratorio`, los subtests
  «directorios de la familia» y «control es skills/normativa», `cmd/laboratorio/main.go`,
  `scripts/evals-laboratorio.sh`, `CONTRIBUTING.md`, `docs/SKILLS.md`.
- `internal/evals` ya importa `internal/skills` (`consultas.go`, `definicion.go`): la dependencia de `skills.Frases` (T004,
  T005) no es nueva.
- La rama `033-h41-las-skills-del` existe en local (`git show 033-h41-las-skills-del:internal/evals/laboratorio_familias.go`
  devuelve el fichero). Que esté también en `origin` no se ha podido comprobar: `git ls-remote` lo bloquea el sandbox.
- `scripts/workflow/guardian-diff.sh`: `testdata/` solo con `[datos]`, y `verificar.sh` pide manifiesto de grabación solo a
  una tarea `[datos]` «con manifiesto»: T008 (`[datos]`, sin manifiesto) es coherente con el workflow.
- El mensaje que T001 espera hoy de `laboratorio tabla normativa` (`… no es ninguna variante de normativa`) es el de H40
  (`specs/032-h40-por-qu-e/contracts/tabla.md` §1).

## Hallazgos

| ID | Categoría | Severidad | Ubicación | Resumen | Recomendación |
|----|-----------|-----------|-----------|---------|---------------|
| F1 | Inconsistency | LOW | tasks.md:L98-99, L105-106, L118 (T003/T004/T005 y checkpoints) vs Makefile `skills-check` | Los criterios de T003 («`TestSkillsDelProductoDesdeControl`, `TestFrases`… en verde con `make skills-check`»), el checkpoint de la fase 3 («`plazos-y-territorio` sigue siendo su control; `make skills-check` lo comprueba») y T005 dan a `make skills-check` la comprobación de esos tests. La expresión `-run` de `skills-check` (Makefile:127) solo ganará `TestRemisionesDelRepositorio` (T004): `TestSkillsDelProductoDesdeControl` y `TestFrases` corren en `make test`, dentro de `make ci`, no en `skills-check`. | Al implementar, validar con `make test` (o `make ci`), que sí los corre; no depender de `skills-check` para ellos. No cambia el alcance (ni el Makefile ni la receta de T004 piden más). |
| F2 | Coverage Gap | LOW | spec.md:L231-232 (caso límite «fila sin ninguna sesión terminada») vs tasks.md T008 | T008 pide una sesión sin terminar (código 124) en los fixtures, pero ninguna fila cuyas sesiones estén todas sin terminar. El caso se declara «como en H40», que ya lo cubre `TestTablaDeNormativa`/`TestTablaDelLaboratorio` sin cambios. | Sin acción obligatoria: el comportamiento es el de H40 y sus tests siguen en verde (T007). Si se quiere blindar, una sesión más en un fixture de T008. |
| F3 | Coverage Gap | LOW | spec.md:L365-368 (FR-030) vs tasks.md T006 | La lista de peticiones inválidas de T006 prueba `normativa-remisiones-en-orden` con `MODOS=orden`, pero no `normativa-cuerpo-corto` ni `normativa-testigo-en-referencia` con `MODOS=orden`. | Sin acción: la regla es la misma por familia y las suites de H40 (`TestArnesDeNormativa`) se mantienen sin cambiar un valor. |
| F4 | Underspecification | LOW | tasks.md:L34-39 (research D20) | El punto de partida (`git show 033-h41-las-skills-del:<ruta>`) depende de una rama que consta en local; que esté en `origin` («local y en `origin`») no se ha podido verificar en esta sesión. | Si el run corre en un clon sin la rama local, las tareas no tienen su punto de partida; el script de lanzamiento puede comprobarlo antes de T002. |
| F5 | Underspecification | LOW | spec.md:L373-377 (FR-032), tasks.md T007, gates/supuestos.md | Un `sondeo.json` sin `referencias_del_producto` junto a un directorio `<skill>-producto` no tiene comprobación ni test (decisión registrada en `supuestos.md`: defecto `inesperado` de hoy). FR-032 dice que `make ci` sale en rojo «si la tabla no toma de ella las columnas»: eso lo cubre el fixture de `normativa` de T008, no el estado sin lista. | Sin acción: decisión conservadora ya registrada, coherente con el criterio de decisión autónoma. |

Sin hallazgos de duplicación, ambigüedad con marcadores pendientes (`TODO`, `???`, `<placeholder>`) ni de deriva
terminológica relevante entre `spec.md`, `plan.md` y `tasks.md`. Las cifras de FR-033 cuadran (19 filas, 760 sesiones;
T006: 16, 4, 40 y 16 sesiones con `SESIONES=4`) y las de contenido también (7 remisiones de `normativa` para 8
referencias, dos de ellas en una misma frase; 8, 4 y 2 referencias por familia).

## Coverage Summary

| Requisito | ¿Tarea? | Tareas | Notas |
|-----------|---------|--------|-------|
| FR-001, FR-002, FR-003 | Sí | T003, T004 | Regla y lista en `TestRemisionesDelRepositorio` (T004); skill de `normativa` en T003 |
| FR-004 | Sí | T003 | |
| FR-005 | Sí | T004 | |
| FR-006 | Sí | T003 | Movimientos, ver `TestSkillsDelProductoDesdeControl` |
| FR-007 | Sí | T003, T004 | |
| FR-008 | Sí | T003, T004 | `TestSkillsDelRepositorio` de hoy, con mutante de 11 001 caracteres |
| FR-009 | Sí | T003, T004 | |
| FR-010 | Sí | T003, T004 | Subtest `plazos-y-territorio` y `fueraDeLasRemisiones` |
| FR-020 | Sí | T002 | |
| FR-021 | Sí | T002, T005 | |
| FR-022 | Sí | T002, T006 | Familias con `producto`; copia desde `skills/` |
| FR-023 | Sí | T005 | |
| FR-024 | Sí | T005 | No es umbral; lo cumplen los cuerpos (contracts/variantes.md §3) |
| FR-025 | Sí | T001, T002, T005 | |
| FR-030 | Sí | T006 | |
| FR-031 | Sí (transversal) | Convenciones de todas | Sin control propio: lo exige el diseño de las tareas |
| FR-032 | Sí | T006, T007, T008 | |
| FR-033 | Sí | T006 | Cuatro invocaciones |
| FR-040 | Sí | T001, T007, T008 | |
| FR-041 | Sí | T007, T008 | |
| FR-042 | Sí | T007 y todas | |
| FR-050 | Sí | T009 | |
| FR-051 | Sí (cierre) | T009 (línea) | Lo mide el job de evals del cierre |
| FR-052, FR-053 | Sí (pendiente) | T009 (`gates/supuestos.md`) | Fuera del run por diseño (ADR 0032) |
| SC-001 a SC-009 | Sí | T001 a T008 (ver «Controles de umbral» de plan.md y Notes de tasks.md) | Cada umbral tiene su control y su mutante |
| SC-010 | Cierre | — | Job de evals con los umbrales de hoy |
| SC-011, SC-012 | Fuera del run | — | Capa 3 (persona) |

## Constitution Alignment Issues

Ninguno. Comprobado contra la constitución: test e2e primero (T001, `[aceptacion]`); tests offline sin red y sin
grabaciones (sin tarea de `grabar_datos`; `testdata/` solo bajo `[datos]`, T008); sin sesiones con modelo en el run
(ADR 0032); skills sin código y con cuerpo ≤ 11 000 caracteres (ADR 0046); genericidad por materia y por municipio
(ningún caso especial); el run no fusiona ni escribe ADR de decisión humana (FR-052, FR-053); CHANGELOG en T009.

## Unmapped Tasks

Ninguna: T001 a T009 mapean a historias y requisitos. T009 es transversal (documentación y cierre); US6 no tiene tareas
por diseño (se hace fuera del run).

## Metrics

- Requisitos funcionales: 27 (FR-001 a FR-010, FR-020 a FR-025, FR-030 a FR-033, FR-040 a FR-042, FR-050 a FR-053).
- Criterios de éxito: 12; 9 con trabajo construible (SC-001 a SC-009); SC-010 a SC-012 son medidas del cierre o de la
  persona.
- Tareas: 9 (T001 a T009), ninguna `[P]`.
- Cobertura (requisitos con ≥ 1 tarea o declaración explícita): 27 de 27 (100 %); FR-031, FR-051, FR-052 y FR-053 por
  convención transversal o por línea de cierre, no por una tarea propia.
- Ambigüedades: 0. Duplicaciones: 0. Hallazgos CRITICAL: 0. HIGH: 0. MEDIUM: 0. LOW: 5.

## Next Actions

- Sin CRITICAL ni HIGH: se puede pasar a `/speckit-implement`.
- F1: al validar T003 a T005, correr `make test` o `make ci` (no solo `make skills-check`) para ver `TestSkillsDelProductoDesdeControl`
  y `TestFrases`.
- F4: comprobar la rama `033-h41-las-skills-del` en el clon donde corra el run antes de T002.
- F2, F3 y F5 no piden cambio: son decisiones ya declaradas o cubiertas por suites que no cambian.

## Extension Hooks

`.specify/extensions.yml` registra `speckit.git.commit` como opcional en `before_analyze` y `after_analyze`. En modo
autónomo no se ha ejecutado ninguno (commits fuera de esta sesión de un paso).
