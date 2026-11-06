# Specification Analysis Report — H41 · Las skills del producto mandan leer sus referencias

Artefactos: `spec.md`, `plan.md`, `tasks.md` (más `research.md`, `data-model.md`, `contracts/*` y la constitución, leídos por tramos).
Solo lectura: no se ha modificado ningún fichero salvo este informe.

## Findings

| ID | Category | Severity | Location(s) | Summary | Recommendation |
|----|----------|----------|-------------|---------|----------------|
| F1 | Inconsistency | MEDIUM | spec.md:330-335 (FR-041) vs contracts/tabla.md §2 y tasks.md T006 | FR-041 dice que el `cd` va a un directorio «cuya ruta termina en `<skill>/references`»; el contrato y T006 exigen que sea `<skill>/references` o termine en `/<skill>/references`. Leída al pie de la letra, la spec admitiría `…/xnormativa/references`. | Dejar que el contrato (más estricto) rija y alinear la redacción de FR-041 con él (`/<skill>/references`). Sin efecto en el código previsto. |
| F2 | Ambiguity | LOW | spec.md:352 (FR-051), SC-009; plan.md:234-238 | FR-051/SC-009 hablan de «las evals de las skills del producto»; el plan las limita a `normativa`, `plazos-y-territorio` y `jurisprudencia` y excluye `consulta-juridica` (no publica umbrales con `decide: true`, su skill no cambia). La exclusión está razonada en el plan pero no en el spec. | Aceptable: `consulta-juridica` no cambia (FR-007). Si se quiere cerrar, una frase en el spec. |
| F3 | Coverage | LOW | spec.md:299-302 (FR-030), tasks.md T005 | FR-030 y los supuestos del clarify nombran `SESIONES=0` como inválido («mayor que 0»); la lista de peticiones inválidas de `TestArnesDeLasFamilias` en T005 no lo incluye (sí `SESIONES=6`). Puede estar ya cubierto por los tests de H40, pero T005 no lo dice. | Añadir `SESIONES=0` a la lista de inválidas de T005 o indicar que lo cubre `TestArnesDelLaboratorio`. |
| F4 | Inconsistency | LOW | tasks.md T003 (`[US2]`), T004 (`[US1]`) | T003 lleva la etiqueta US2 pero también reescribe las seis remisiones de `normativa` (FR-004, historia 1); T004 lleva US1 y no toca ningún movimiento. El mapeo historia→tarea es aproximado. | Sin efecto en la ejecución; si se corrige, etiquetar T003 «US1 y US2». |
| F5 | Underspecification | LOW | spec.md:215-222 (FR-001) vs contracts/remisiones.md §6.2 | El contrato añade límites de frase que el spec no enumera (marca `#` y `* ` además de `- ` y número), y el spec dice «marca de una viñeta o de un paso». Es un refinamiento, no un conflicto. | Ninguna acción; el contrato es el que se ejecuta. |
| F6 | Underspecification | LOW | plan.md:12; contracts/remisiones.md §1 | Tras el hito, el margen bajo el tope de 11 000 es de 18 caracteres (`plazos-y-territorio`: 10 982) y 19 (`normativa`: 10 981). El tope no se toca (Fuera de alcance), pero cualquier paquete nuevo que añada una línea a la región generada de `plazos-y-territorio` pondrá `TestSkillsDelRepositorio` en rojo. Antes del hito el margen era 70 (research D3) y ya no cabía un paquete grande. | Riesgo conocido y declarado como fuera de alcance (ADR 0046); se anota para el próximo hito de paquetes. |
| F7 | Duplication | LOW | tasks.md T001 vs T007 | Las tablas esperadas de las tres familias se escriben dos veces a mano (en `laboratorio-tabla.txtar` de T001 y en `testdata/.../tabla.md` de T007), con extremos de Wilson calculados por separado; si divergen, un control falla y el otro no. | Aceptable (T001 queda congelada y T007 se mide contra la salida de T006). Opcional: que T007 reutilice los valores de T001 byte a byte. |
| F8 | Style | LOW | spec.md:306-313 | FR-033 aparece entre FR-031 y FR-032 (orden de ids alterado). | Cosmético. |

Sin hallazgos CRITICAL ni HIGH.

## Coverage Summary

| Requirement Key | Has Task? | Task IDs | Notes |
|---|---|---|---|
| FR-001, FR-002, FR-003 (remisiones como órdenes; condición; sitio) | Sí | T003, T004 | Regla ejecutable en `TestRemisionesDelRepositorio` (T004), con subtest «control» y «frases» |
| FR-004 (normativa) | Sí | T003 | Seis cambios R1–R6 + movimiento; `como_se_cita.md` en el paso 2 (contracts §2, R5) |
| FR-005 (plazos-y-territorio) | Sí | T004 | P1–P3; `administrativo.md` y bloque generado intactos |
| FR-006 (jurisprudencia) | Sí | T004 | J1, J2 |
| FR-007 (nada más cambia) | Sí | T003, T004 | `TestSkillsDelProductoDesdeControl` |
| FR-008 (≤ 11 000) | Sí | T003 (mutante 11 001), T004 | `TestSkillsDelRepositorio` de hoy |
| FR-009 (movimientos enteros) | Sí | T003 | Texto en la referencia y fuera del cuerpo |
| FR-010 (frontmatter y `skills-sync`) | Sí | T003, T004 | |
| FR-020 (`<skill>-control` congelada) | Sí | T002 | Huellas SHA-256, 17 ficheros |
| FR-021 (variantes H40 vs control) | Sí | T002 | `TestVariantesDeNormativa` sin cambiar valores |
| FR-022 (`<skill>-producto` desde `skills/`) | Sí | T002, T005 | |
| FR-023 (variantes fuera del binario/instalación/job) | Sí | T002 | `TestLaboratorioFueraDelProducto` generalizado |
| FR-030 (arnés, argumentos, evals por familia) | Sí | T002, T005 | Ver F3 |
| FR-031 (el run no lanza ni graba) | Transversal | convenciones de todas las tareas | Sin tarea propia, por diseño |
| FR-032 (13 filas, 520 sesiones) | Posibilitado | T005 | Se mide fuera del run (persona) |
| FR-033 (lista de referencias en `sondeo.json`) | Sí | T005, T006 | |
| FR-040 (tabla de tres familias) | Sí | T006, T007 | |
| FR-041 (lectura con `cd`) | Sí | T006, T007 | Ver F1 |
| FR-042 (suites H39/H40 en verde) | Sí | T006 y todas | Comprobado en `TestEntregaDelHito` |
| FR-050 (CHANGELOG, `docs/SKILLS.md`) | Sí | T008 | |
| FR-051 (evals en verde en el cierre) | Cierre del workflow | T008 (línea de supuestos) | Ver F2 |
| FR-052 (sondeo de la persona antes de fusionar) | Fuera del run | T008 (`[alcance]`) | Capa 3, declarado |
| SC-001…SC-008 | Sí | ver «Controles de umbral» del plan | Cada uno con test o guion y mutante |
| SC-009, SC-010 | Capa 3 / job de cierre | — | No automatizables en el run (ADR 0032); declarado |

**Unmapped tasks**: ninguna. T001 y T008 son transversales (suite de aceptación; documentación y cierre).

## Constitution Alignment

Revisadas las «Reglas del modo desatendido» y los principios I–X con la tabla del plan:

- La primera tarea (T001) escribe la suite de aceptación, comprueba que falla hoy y queda congelada. ✔
- Cada tarea declara sus rutas; las dos que renombran o crean ficheros (T002, T004) las listan. ✔
- Ninguna tarea graba, usa la red más que la de Go, escribe en `evidencias/` ni en `data/`, ni abre sesiones con modelo. ✔
- Ningún fichero existente de `testdata/` ni de `schemas/` cambia (capa 3). ✔ (T007 solo añade fixtures nuevos.)
- Principio VIII (skills primero), X (genericidad por materia: lo generado se excluye del cálculo desde control): ✔.

Sin conflictos con la constitución.

## Metrics

- Requisitos funcionales: 24 (FR-001…FR-052); criterios de éxito con trabajo construible: 8 (SC-001…SC-008) más 2 de capa 3
- Tareas: 8 (T001…T008)
- Cobertura FR con ≥ 1 tarea: 23/24 (96 %); FR-031 es transversal por diseño
- Ambigüedades: 2 (F2, F5)
- Duplicados: 1 (F7)
- Incidencias críticas: 0

## Qué no se ha medido

- No se ha ejecutado `make ci` ni ningún test: no hay implementación todavía.
- Las cifras de cuerpo de las skills tras el hito (10 981, 10 982, 10 842) y los 17 ficheros de control no se han recalculado aquí sobre copias con los cambios; solo se comprobó que `normativa-control` tiene 9 ficheros, que existen las evals nombradas de `plazos-y-territorio` y `jurisprudencia`, que `go.mod` declara `go 1.27.0`/`toolchain go1.27.3` y que el nombre de `TestRemisionesDelRepositorio` aún no está en `skills-check` (lo añade T004).
- Las cifras de H40 (58 de 60, 9 sesiones) las cita el spec como supuestos no medidos aquí.
- El gancho opcional `before_analyze`/`after_analyze` (`speckit.git.commit`) no se ha ejecutado (modo autónomo, solo lectura).

## Next Actions

- No hay CRITICAL ni HIGH: se puede pasar a implementar.
- Antes o durante la implementación, sin bloquear: alinear la redacción de FR-041 con el contrato (F1) y añadir `SESIONES=0` a la lista de peticiones inválidas de T005 (F3).
