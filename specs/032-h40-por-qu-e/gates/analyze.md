# Specification Analysis Report — H40 (`032-h40-por-qu-e`)

Alcance: `spec.md`, `plan.md` y `tasks.md`, contra `.specify/memory/constitution.md` (v2.15.0). Solo lectura. Lo que hay
aquí viene de leer los tres ficheros y de `ls`/`grep` sobre el repositorio en esta sesión; no se ha ejecutado `make ci`
ni ningún test (el hito aún no tiene código), así que nada de lo que sigue es una medida de la suite.

Hooks de extensión: `before_analyze` y `after_analyze` solo traen `speckit.git.commit`, opcional; en modo desatendido no
se ejecuta.

## Hallazgos

| ID | Categoría | Severidad | Ubicación | Resumen | Recomendación |
|----|-----------|-----------|-----------|---------|---------------|
| I1 | Inconsistencia | LOW | tasks.md T001; contracts/arnes-e2e.md:36; research V14 | Las fijas de H39 se cuentan como «63 ficheros», pero `internal/evals/testdata/laboratorio/sondeo/` tiene 65 (21 sesiones × 3 = 63, más `sondeo.json` en `h39-a` y `h39-b`; contado en esta sesión: 49 + 16). La suite congelada necesita los `sondeo.json` (modelo y versión) para que `tabla h39` dé la tabla de `main`; leído al pie de la letra, «los 63 ficheros de las sesiones» los dejaría fuera | Al escribir T001, copiar las dos carpetas enteras (65 ficheros). El prototipo que pide T001 («la parte de H39 pasa») lo detecta si se copian 63 |
| I2 | Inconsistencia | LOW | spec.md:9-10 | El spec cita como entrada `docs/roadmap/in-progress/H40-…md`; el fichero está hoy en `docs/roadmap/done/` (commit 3d6c3453, anterior al spec) y `in-progress/` solo tiene el README | Sin efecto en las tareas; no hay que cambiar nada en el spec, es lo que hace el workflow al lanzar |
| I3 | Terminología | LOW | spec.md FR-001, FR-040; plan.md Resumen; tasks.md | El spec nombra las variantes `control`, `remisiones-en-orden`…; el plan y las tareas, `normativa-control`, `normativa-remisiones-en-orden`… (el spec delega el nombre en el plan, FR-010, y el plan lo fija). El mapeo está en plan/data-model, no en el spec | Aceptable. La persona pide `VARIANTES=normativa-control,…` y `control` sigue siendo la de H39; lo cubre el caso `VARIANTES=control,normativa-control` de T004 |
| U1 | Infraespecificación | LOW | tasks.md T004 (test) | FR-011 («anota modelo y versión en la tabla») y FR-012 («conserva la transcripción con su variante, modo y eval») no tienen una aserción propia en `TestArnesDeNormativa`; los cubre el camino compartido con H39 (`TestArnesDelLaboratorio`, contracts/arnes.md línea 68-69) y T001/T006 leen ese formato | Añadir a `TestArnesDeNormativa`, si cabe sin salir de las rutas declaradas, que el sustituto deja `<variante>/<modo>/<eval>-<nn>/` y `sondeo.json`; si no, dejarlo como está |
| C1 | Cobertura | LOW | spec.md SC-008, FR-043; tasks.md | SC-008 pide las evals de producto en verde en el cierre, y FR-043 admite un rojo de partida en modo herramienta (eval 04 de `normativa`, ejecución 41474010633) que el run no puede reparar tocando `skills/`. Sin tarea, por diseño: lo mide el job del cierre | Sin cambio. El informe final debe dar las cifras si hay rojo (ya anotado en `gates/supuestos.md` y en plan.md «Cierre y pendientes») |
| P1 | Paralelismo | LOW | tasks.md T004, T005 | T005 es [P] con T004, pero T004 declara `laboratorio_normativa.go` y T005 `laboratorio_normativa_test.go`; por la regla «declarar `x.go` admite `x_test.go`», los dos admiten el mismo fichero de tests. Además, «Orden de las fases» dice que T005 solo necesita T002 y el ejemplo dice «tras T003» | Ejecutarlas en serie (T004 → T005) o aceptar el solape; no cambia el resultado |

Sin hallazgos CRITICAL, HIGH ni MEDIUM. Sin duplicaciones, sin marcadores pendientes (`TODO`, `???`, `<placeholder>`) en
los tres ficheros.

## Cobertura

| Requisito | ¿Tarea? | Tareas | Notas |
|-----------|---------|--------|-------|
| FR-001 (fuera del producto) | Sí | T002, T003 | `TestLaboratorioFueraDelProducto` generalizado (D12) |
| FR-002 (control byte a byte) | Sí | T002 | `TestVariantesDeNormativa`; mutantes |
| FR-003 (`remisiones-en-orden`) | Sí | T003 | Cinco sustituciones, textos del contrato |
| FR-004 (`cuerpo-corto`) | Sí | T003 | Regla por secciones + encabezado renombrado |
| FR-005 (`testigo-en-referencia`) | Sí | T003 | Inserción + sección con el párrafo de `cierre.md` |
| FR-006 (comparación por declaración) | Sí | T003 | Falla si la declaración ya no se aplica |
| FR-007 (tope no rige las variantes) | Sí | T002, T003 | `skills/` sin cambios; `TestSkillsDelRepositorio` |
| FR-008 (servidor y skills de la sesión) | Sí | T004 | Solo `normativa` de la variante; `servidor.json` de hoy |
| FR-010 (arnés, dos familias) | Sí | T004 | Peticiones inválidas → 0 sesiones; n/4 por eval |
| FR-011 (versión y modelo) | Sí | T004 | Ver U1 |
| FR-012 (lo lanza una persona) | Sí | T004, T007 | Sin receta nueva; ver U1 |
| FR-013 (el run no graba) | Sí | todas | Convención de las tareas |
| FR-020 (tabla) | Sí | T001, T006 | 16 columnas, detalle |
| FR-021 (sin terminar, n = 0) | Sí | T001, T006 | Fixtures con ambos casos |
| FR-022 (control del guion) | Sí | T006 | `TestTablaDeNormativa`, Wilson 0/20/40 de 40 |
| FR-023 (H39 no cambia) | Sí | T001, T006 | Guion congelado + `TestTablaDelLaboratorio` |
| FR-030 (regla testigo) | Sí | T005, T006 | Seis respuestas + la de H32 |
| FR-040 (sondeo de la persona) | No (por diseño) | — | Fuera del run; quickstart |
| FR-041 (ADR) | No (por diseño) | — | Fuera del run |
| FR-042 (el run no lanza ni escribe el ADR) | No (por diseño) | — | `gates/supuestos.md` e informe final |
| FR-043 (evals del cierre) | No (por diseño) | — | Job del cierre; ver C1 |
| SC-001 a SC-007 | Sí | T002 a T006 | Cada fila de «Controles de umbral» tiene control con forma |
| SC-008 | No (por diseño) | — | Job de evals del cierre |
| SC-009 | No (por diseño) | — | Capa 3, la persona |

## Alineación con la constitución

Sin conflictos. Comprobado:

- Constitution Check del plan completo (principios I a X y reglas 1 a 7), sin violaciones declaradas.
- III (tests primero): T001 [aceptacion] escribe la suite, con «rojo primero» (falla por la aserción de `tabla normativa`),
  y ninguna tarea posterior la toca.
- «Controles de umbral»: cada fila del plan trae umbral, control y lugar con forma `ci:…` o `evals:…`.
- Reglas del modo desatendido: cada tarea declara sus rutas; las dos [datos] (T005, T006) solo añaden ficheros nuevos en
  `internal/evals/testdata/`; ninguna toca `skills/`, `evals/`, `internal/mcp/`, `laboratorio/variantes*` ni las recetas
  del `Makefile`; ninguna graba ni abre sesiones con modelo.
- Checklists: no hay directorio `checklists/` en el feature; no se ha exigido ninguna.

## Tareas sin requisito

Ninguna. T007 (documentación y CHANGELOG) se traza a FR-010, FR-012, FR-020 y FR-040 y a la Definition of Done.

## Métricas

- Requisitos totales: 30 (21 FR + 9 SC)
- Tareas totales: 7
- Cobertura (≥ 1 tarea): 24 de 30 = 80 %; el 100 % de lo construible dentro del run. Los 6 restantes (FR-040 a FR-043,
  SC-008, SC-009) son del cierre del workflow o de la persona, y el spec y el plan lo dicen
- Ambigüedades: 0 materiales (I1 es un recuento mal escrito)
- Duplicaciones: 0
- CRITICAL: 0 · HIGH: 0 · MEDIUM: 0 · LOW: 6

## Próximos pasos

Sin CRITICAL ni HIGH: se puede pasar a implementar. Lo único con efecto práctico es I1: al escribir T001, copiar las dos
carpetas de H39 enteras, con sus `sondeo.json` (65 ficheros), no solo los 63 de las sesiones. Los demás hallazgos no
piden cambios en los artefactos.
