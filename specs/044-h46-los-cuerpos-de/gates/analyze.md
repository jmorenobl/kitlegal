# Informe de análisis · H46 (`spec.md`, `plan.md`, `tasks.md`)

Solo lectura: no se ha modificado ningún fichero salvo este informe. Leídos: `spec.md`, `plan.md` y `tasks.md`
enteros; de la constitución, solo los encabezados de principios (el plan los recorre uno a uno en su «Constitution
Check»). No leídos en esta sesión: `research.md`, `data-model.md`, `quickstart.md` y `contracts/` (salvo un `grep` de
`contracts/aceptacion.md`).

**Lo ejecutado y visto en esta sesión**: `check-prerequisites.sh` (spec, plan y tasks presentes);
`git diff --name-only main f89093dd`; `git cat-file -t f89093dd` (`commit`) y `git branch -a` (solo existe la local
`cuerpos-prototipo`); `git grep` de los textos que fija la tabla de cuatro columnas fuera de `specs/044`;
`grep` de `contracts/aceptacion.md`; lectura del principio de `h45-paquete-de-mas.txtar`.

**No medido**: los caracteres de los cuerpos (9498, 9577, 9473, 9513, 10 117, 10 336…), los rojos y verdes de los
mutantes y de `make ci`. Las cifras son las de `plan.md` y `tasks.md`; no las he ejecutado. Sí comprobé a mano que son
coherentes entre sí: 10 750 − 1237 = 9513 y 9513 − 15 (FR-012) = 9498; 10 686 − 569 = 10 117; 10 842 − 506 = 10 336;
9577 − 94 = 9483; la tabla del plan de «216 menos con k = 1, 236 más con k = 2, 452 más por cada documento siguiente»
sale de sus propias cifras.

## Specification Analysis Report

| ID | Categoría | Severidad | Ubicación | Resumen | Recomendación |
|----|-----------|-----------|-----------|---------|---------------|
| D1 | Dependencia no declarada | MEDIUM | spec.md «Dependencias»; tasks.md T003, T006 | Los textos (FR-010) salen de `git checkout f89093dd -- skills/`, un commit que solo cuelga de la rama **local** `cuerpos-prototipo` (no hay remota ni está en `main`). «Dependencias» solo cita H45.1. Si el run se hace en un clon o worktree sin esa rama, o la rama se borra y el commit se recoge, T003, Q4 y T006 fallan. Aquí el objeto existe. | Tratarlo como entrada del run: que el workflow compruebe `git cat-file -e f89093dd` antes de T003 (causa mayor «dependencia inaccesible» de la constitución) y que «Dependencias» lo nombre. No exige cambiar el diseño. |
| D2 | Cobertura (por diseño) | LOW | FR-013, FR-019, SC-007, SC-008, US5 | Sin tarea propia que los construya: FR-013 queda por construcción (bytes del prototipo) y por `TestSkillsDelProductoDesdeControl`; FR-019 y SC-007 por los tests de cada applet sin tocar y por Q8 de T006; SC-008 y US5 por las evals del cierre del workflow (research D13). Está declarado en plan y tasks. | Ninguna. Solo se anota que SC-008 no lo cubre ninguna tarea ni se puede medir antes del cierre. |
| D3 | Ambigüedad de redacción | LOW | tasks.md T001 | T001 cita por adelantado números de línea de ficheros que aún no existen (`tabla-de-tres-columnas.txtar:20`, `textos-en-su-referencia.txtar:25`). Vienen de `contracts/aceptacion.md`, que no he releído. | Dejarlo: el criterio real es «cae por una aserción», no la línea. Si el run los escribe distintos, no es un defecto. |
| D4 | Margen | LOW | spec.md «Edge cases»; plan.md P1 | Con los textos del prototipo los cuerpos quedan en 9498 / 9483 / 9473 frente a 9600: holguras de 102, 117 y 127 caracteres, no los 1400 que el hito promete a las materias. Es coherente con el ADR 0063 (el tope mide el cuerpo sin las líneas de los paquetes), pero cualquier retoque de protocolo en estas tres skills antes de H42 se come esa holgura. | Ninguna acción en H46. Lo cubre el defecto nuevo, que es el objeto del hito. |
| D5 | Fuera del diff de la tarea | LOW | `git diff main f89093dd` | El prototipo cambia también `web/src/data/plan.ts`, `docs/` y `.specify/workflows/hito/workflow.yml`, que ni spec ni tasks incluyen. Es correcto: T003 solo hace checkout de `skills/`. | Ninguna. Recordar que «la rama del prototipo» no es lo que se fusiona. |

Sin hallazgos de duplicación, de marcadores sin resolver (`TODO`, `???`, `<placeholder>`) ni de terminología distinta
entre los tres ficheros. Los nombres coinciden: `maximoSinLosPaquetes`, `casosDePresupuestos`,
`declaracionesDelProducto`, `remisionesDelPlan`, `comprobarElSitioDeOtroPaquete`, las marcas de los paquetes.

## Alineación con la constitución

Sin violaciones. El «Constitution Check» del plan recorre I a X y R1 a R7. Contrastado con los tres ficheros:

- **I, VII**: ninguna petición ni fuente nueva; el hito no graba nada y `tasks.md` lo repite («Datos externos: ninguno»).
- **III (tests primero y offline)**: T001 escribe la suite congelada antes que nada. T002 a T004 llevan su test junto al
  código. No hay un paso de «rojo primero» propio dentro de T002 a T004. En T004 lo da el orden: tras T003 los casos de
  9601 y 11 001 ya caen sin la regla. Se acepta por el precedente de H41.1.
- **V (YAGNI)**: T002 retira lo que la tabla deja sin uso; ninguna dependencia nueva.
- **VIII, X**: el hito es su aplicación; la regla de 9600 no nombra ninguna skill ni materia (FR-005, FR-006).
- **Suite congelada**: el plan declara en «Complexity Tracking» la modificación de los guiones activados de H41 y H41.1
  y de la copia congelada de H41.1. Es una desviación consciente y mínima, con precedente (FR-013 de H41.1 y de H45.1).
  Es lo que la persona debe mirar en el informe final.

## Resumen de cobertura

| Clave | ¿Tarea? | Tareas | Notas |
|-------|---------|--------|-------|
| FR-001 tabla de tres columnas | Sí | T001, T002 | |
| FR-002 mismos errores de `--describe` | Sí | T002 | tests de hoy sin tocar; mutante en T006 |
| FR-003 tests con los bytes nuevos | Sí | T002 | |
| FR-004 `skills-sync` / `skills-check` | Sí | T002, T003 | |
| FR-005 defecto de 9600 | Sí | T004 | |
| FR-006 skills sin marcas | Sí | T004 | casos `consulta-juridica`; mutante en T006 |
| FR-007 cuerpos sin defecto | Sí | T004 | `make skills-check` |
| FR-008 `casosDePresupuestos` | Sí | T004 | |
| FR-009 `TestVerticalDeEnsayo` | Sí | T004 | |
| FR-010 a FR-012 textos | Sí | T003 | `git checkout f89093dd -- skills/` (véase D1) |
| FR-013 nada de `⚠`/formas/reglas/órdenes sale | Sí, por construcción | T003 | véase D2 |
| FR-014 declaraciones del producto | Sí | T003 | 15 entradas |
| FR-015 remisiones | Sí | T003 | 2 frases |
| FR-016 guiones de H41 y H41.1 | Sí | T003 | con la copia congelada |
| FR-017 `docs/SKILLS.md` | Sí | T002 (§1), T004 (§3, §5, ensayo) | |
| FR-018 `CHANGELOG.md` | Sí | T005 | |
| FR-019 ningún applet cambia | Control, sin tarea propia | T006 (Q8) | véase D2 |
| FR-020 orden de tareas | Sí | T001 a T006 | coincide con el orden del spec |
| FR-021 suite de aceptación | Sí | T001 | |
| SC-001 a SC-006 | Sí | T002 a T006 | tabla «Controles de umbral → tarea» de `tasks.md` |
| SC-007 | Control | T006 (Q8) | |
| SC-008 / US5 | Evals del cierre | — | las lanza el workflow, no una tarea |

## Tareas sin requisito

Ninguna. T005 → FR-018; T006 → SC-003 a SC-006 (registro de mutantes, asumido en «Assumptions»).

## Métricas

- Requisitos: 21 FR + 8 SC = 29 claves (más 5 historias).
- Tareas: 6.
- Cobertura con al menos una tarea o control declarado: 28/29 (96 %). Queda fuera SC-008, que decide el cierre de evals
  y no una tarea, por diseño (D2). 21/21 de los FR tienen tarea o control.
- Ambigüedades: 1 (D3, LOW).
- Duplicados: 0.
- Críticos: 0. Altos: 0. Medios: 1 (D1). Bajos: 4.

## Siguientes pasos

Sin hallazgos CRITICAL ni HIGH: puede pasar a `/speckit-implement`. Lo único con efecto en el run es D1, que se resuelve
comprobando que `f89093dd` es alcanzable antes de T003 (`git cat-file -e f89093dd`); aquí lo es. No se aplica ninguna
edición a `spec.md`, `plan.md` ni `tasks.md`.
