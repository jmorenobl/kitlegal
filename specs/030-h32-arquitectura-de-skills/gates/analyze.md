# Specification Analysis Report — H32 · Arquitectura de skills

Artefactos: `spec.md` (43 FR, 18 SC, 8 US), `plan.md`, `tasks.md` (T001–T018), `contracts/`, `research.md`,
`data-model.md`. Constitución: `.specify/memory/constitution.md`. Solo lectura: no se ha modificado ningún artefacto.

## Hallazgos

| ID | Categoría | Severidad | Ubicación | Resumen | Recomendación |
|----|-----------|-----------|-----------|---------|---------------|
| C1 | Constitución | MEDIUM | tasks.md T003; constitución «Reglas del modo desatendido» | La regla dice «ningún fichero de `data/` se escribe a mano ni de memoria: lo produce código a partir de lo grabado». T003 escribe a mano en `data/normas.yaml` `vertical: administrativo` en diez normas (clasificación del plan, sin fuente: plan «Datos externos») y la entrada de la Ley 50/1980 (título y rango copiados de lo grabado, no de memoria; `TestIdentificadoresDeLasNormas` los coteja). Hay precedente: `git log` da commits de H5, H5.1 y H6 sobre `data/normas.yaml`, y las tareas de H4 a H6, H9 y H26.1 lo nombran. | Registrar como supuesto que `data/normas.yaml` es el catálogo curado por una persona que la constitución no pretende generar, y que el título de la Ley 50/1980 se copia de lo grabado. No bloquea. |
| E1 | Cobertura | LOW | spec FR-033; tasks T004, T005 | El criterio «`git diff --find-renames c991808 -- evals/` da cada fichero como renombrado o sin cambios» solo está como criterio de tarea y de revisión final. Ni la suite de aceptación ni `make ci` lo comprueban (el CI corre sobre un clon superficial, T001). | Aceptable por diseño. La revisión final tiene que ejecutar el `git diff`. |
| E2 | Cobertura | LOW | spec FR-032; contracts/aceptacion.md | Ocurrencias de los nombres retirados fuera de los directorios de la búsqueda de FR-032: `scripts/workflow/politica-paso.sh` (ejemplos `SKILL=boe-legislacion`), `scripts/workflow/comun.sh` y `.golangci.yml` (comentarios), `testdata/evals/grabaciones.json` (lo grabado), `CLAUDE.md` y `CONTRIBUTING.md`. `scripts/workflow/` está fuera de alcance y las dos últimas las declara el spec como pendientes de una persona. | Sin acción en el hito. La persona decide el resto con el informe final. No he comprobado si la política de `politica-paso.sh` casa por patrón, de modo que `SKILL=normativa` quede cubierto igual. |
| I1 | Inconsistencia | LOW | tasks T004 | El texto cita `ejemplo/kitlegal-e2e/`. La ruta real es `internal/app/ejemplo/kitlegal-e2e/` (comprobado con `git ls-files`), que cae dentro de `internal/app/` declarado. | Citar la ruta completa para que el guardián no la tome por una ruta de la raíz. |
| I2 | Terminología | LOW | spec SC-010 | Nombra `evals/boe-legislacion/juez/clases.yaml`, la ruta de `c991808`. Tras T004 es `evals/normativa/juez/clases.yaml`. Es una referencia histórica y `specs/` queda fuera de la búsqueda de FR-032. | Sin acción. |
| U1 | Infraespecificación | LOW | tasks T015 | Las rutas de la vertical de ensayo (`data/normas.yaml`, `skills/normativa/references/ensayo.md`, `evals/normativa/24-…yaml`) están escritas relativas a `internal/skills/testdata/vertical-de-ensayo/`. Un extractor de rutas las lee como rutas de la raíz y admite de más. La ruta declarada de verdad es el directorio de `testdata`. | Prefijar con `internal/skills/testdata/vertical-de-ensayo/` al citarlas. |
| T1 | Orden | LOW | tasks T016 `[P]`, «Oportunidades de paralelismo» | T016 va `[P]` pero comparte el paquete `internal/app` con T004, T013 y T014, y la constitución fija una sola sesión sobre el árbol. En el run el `[P]` no tiene efecto. | Sin acción. El orden serie del fichero es correcto: T016 va tras T005 y T015. |
| R1 | Riesgo | LOW | tasks T004, T005 | Cada tarea renombra carpetas y toca decenas de ficheros en seis paquetes y en los workflows (el grep de FR-032 da hoy 51 ficheros bajo `internal/`, `skills/`, `.github/workflows/`, `README.md` y la web). El texto justifica la tarea única (en cuanto la carpeta cambia, el binario empotra otro nombre). Una cuarentena en T004 bloquea todo lo posterior. | Sin acción. Lo tiene presente quien supervise el run. |

Sin hallazgos CRITICAL ni HIGH. Duplicaciones: 0. Marcadores pendientes (`TODO`, `TKTK`, `???`, `NEEDS CLARIFICATION`):
0. Casillas sin marcar en `checklists/requirements.md`: 0.

## Coverage Summary

| Requisito | ¿Tarea? | Tareas | Notas |
|---|---|---|---|
| FR-001 a FR-004 (entrada) | sí | T013, T014; T001 (`entrada`) | Cuerpo `consulta-juridica` = borrador de 3 660 caracteres. |
| FR-005 (reparto, un applet por skill) | sí | T001, T004, T005, T013 | Tabla de evals en `contracts/reparto.md` §6. |
| FR-006 (cada skill responde sola) | sí | T001 (`capacidades`), T006, T007, T009 | |
| FR-007 (nombres) | sí | T004, T005, T001 | |
| FR-008 (descripción de plazos) | sí | T009 | |
| FR-009 (≤ 11 000) | sí | T006, T007, T009, T013 | Medido: cuerpos de los borradores = 10 858, 10 842, 10 896 y 3 660, como dicen las tareas. |
| FR-010, FR-013 (forma única y formas compartidas) | sí | T010 (controles y mutantes), T006, T009 | Las citas fuera de forma que hay hoy están en `legal-core/SKILL.md`, `como_se_cita.md` y las referencias generadas de `legal-core`. Las reescriben T006 y T009 antes de T010. |
| FR-011, FR-012 (`⚠ NO CUBIERTO:` por fila) | sí | T009 | |
| FR-014 (medida con la cita del precepto) | sí | T011 | |
| FR-015 a FR-018 (normativa) | sí | T006, T001 (`normativa`) | |
| FR-019 a FR-021 (paquete administrativo, genericidad) | sí | T003, T009, T001 (`paquetes`) | |
| FR-022 a FR-024 (molde, enlaces, ensayo) | sí | T008, T009, T015 | |
| FR-025 a FR-031 (nombres retirados al instalar) | sí | T016, T001 (`retiradas`) | |
| FR-032 (nombres de la entrega) | sí | T004, T005, T017, T001 (`nombres-retirados`) | Los ficheros con coincidencias hoy caen todos en T004, T005 o T017. |
| FR-033, FR-034 (evals) | sí | T004, T005 | Ver E1. |
| FR-035 a FR-037 (evals nuevas, `se_activan`) | sí | T002, T003, T012, T014 | |
| FR-038, FR-039 (umbrales, matriz, tope) | sí | T004, T005, T009, T014 | Hoy 5 + 4 + 2 = 11 sesiones; la entrega 1 + 3 + 5 + 2 = 11. Peor caso máximo 20 725 s frente a 21 120 s (`timeout-minutes: 352`). |
| FR-040 (retirar los guiones de H31.1) | sí | T004 | |
| FR-041 a FR-043 (guía, ADR, CHANGELOG) | sí | T018 | `docs/ADR/` llega a 0047, así que 0048 está libre. |
| SC-001 a SC-009, SC-011 a SC-015, SC-017, SC-018 | sí | ver «Notes» de tasks.md | Cada fila de «Controles de umbral» tiene la tarea que construye su control. |
| SC-010, SC-012 | sí (control de CI); el valor lo mide el job | T004, T005, T014 | |
| SC-016 (sondeo) | no (a propósito) | — | Lo lanza una persona; el plan lo lista en «Pendientes» y US8 no tiene tareas. |

## Constitution Alignment Issues

- C1 (MEDIUM). El resto de la constitución se cumple en la medida en que he podido cotejarlo:
  - La suite de aceptación la escribe T001 antes que cualquier código.
  - Las grabaciones las hace `grabar_datos` (T002 solo escribe el manifiesto).
  - Cada tarea declara sus rutas.
  - `schemas/` y `testdata/` modificados se dejan a una persona (plan, «Pendientes»).
  - No hay dependencias nuevas.
  - No se tocan `scripts/workflow/`, `.specify/` ni `internal/workflow/`.
  - Principio X (genericidad por materia): los cuerpos de capacidad quedan sin normas de un paquete (FR-021, con control).

## Unmapped Tasks

Ninguna. T002 y T003 sirven a FR-035 y FR-021/FR-022. T017 y T018 sirven a FR-032 y FR-041 a FR-043.

## Metrics

- Requisitos funcionales: 43. Criterios de éxito: 18, de los que 16 piden trabajo construible.
- Tareas: 18 (T001 `[aceptacion]`, 11 `[datos]`).
- Cobertura de FR con ≥ 1 tarea: 43/43 (100 %). SC sin tarea: solo SC-016, deliberado.
- Ambigüedades: 1 (U1). Duplicaciones: 0. Hallazgos CRITICAL: 0. HIGH: 0. MEDIUM: 1. LOW: 7.

## Qué no he podido medir

- No he ejecutado `make ci` ni ningún test: es un análisis de artefactos. Lo que digo del repositorio sale de `git grep`, `git ls-files`, `wc -m` y lecturas de esta sesión.
- Las ocho probabilidades binomiales de SC-016 no las he recalculado: la sesión no pudo ejecutar Python. Las tomo del spec.
- No he leído entera la parte de `spec.md` anterior a «Requirements» (Contexto y US 1 a 8), ni `research.md`, `data-model.md`, `quickstart.md` ni el plan completo. He cotejado solo las secciones citadas en los hallazgos.

## Next Actions

- No hay CRITICAL ni HIGH: se puede pasar a `/speckit-implement`.
- Opcional, sin bloquear: registrar C1 como supuesto en `gates/supuestos.md` y precisar las rutas de I1 y U1 en `tasks.md`.
