# Specification Analysis Report — H45.1 (`043-h45-1-los-tests-del`)

Análisis de solo lectura de `spec.md`, `plan.md` y `tasks.md` (con `research.md`, `data-model.md`, `contracts/` y
`quickstart.md` como apoyo) contra `.specify/memory/constitution.md`.

**Qué se ejecutó y vio en esta sesión**: lectura de los tres artefactos y de la constitución; `grep`/`sed` sobre el
repositorio para comprobar que lo que el spec y el plan citan existe (líneas de `cobertura_test.go`,
`conjunto_test.go` y `definicion_test.go`; `data/cobertura.yaml:39`; la línea 36 y la región vacía de
`skills/normativa/SKILL.md`; las dos filas de la base que cambia el guion, una vez cada una; los identificadores de test
que las tareas usan); `cmp` de los dos guiones `h45-paquete-de-mas` (iguales); lectura de
`scripts/workflow/guardian-diff.sh`, `tarea.sh` y `precheck.sh`.
**Qué no se midió**: no se corrió `make ci`, ningún test ni ningún mutante (es un análisis previo a implementar y no hay
cambio que probar); las cifras de peor caso (5 h 22 min…), los tiempos del guion y los rojos de L06/L07/T1/L29/L33/P6
son los del prototipo de `research.md` (P1) y no se han repetido aquí.

## Hallazgos

| ID | Categoría | Severidad | Ubicación | Resumen | Recomendación |
|----|-----------|-----------|-----------|---------|---------------|
| F1 | Inconsistencia | MEDIUM | tasks.md T001, T004, T005, T006 · plan.md «Controles de umbral» (último párrafo) · `scripts/workflow/tarea.sh:154` | El plan dice que el guardián de diff hace cumplir que el diff no toca `data/`, `evals/`, `skills/` ni `.github/` «con las rutas que declara cada tarea». Pero `tarea.sh` saca las rutas de **todo token con forma de ruta del texto de la tarea**, y los textos nombran esas rutas para describir mutantes y comprobaciones: T001 y T004 `data/cobertura.yaml`; T005 `data/cobertura.yaml`, `evals/plazos-y-territorio/`, `skills/normativa/references/prueba.md`, `skills/plazos-y-territorio/references/`; T006 `skills/plazos-y-territorio/references/administrativo.md`. Con eso el guardián por tarea **permite** escribir en esos ficheros y la garantía de US1.3/FR-014 queda solo en la orden Q3 de T006 (y en `make skills-check` y los tests de datos). | Redactar los mutantes y las comprobaciones sin ruta literal de lo que no se cambia («la fila de tráfico del YAML de cobertura», «la carpeta de evals de la skill, en la copia»), o aceptar de forma explícita en el plan que la garantía es Q3 y no el guardián. T006 ya corre Q3 al final, así que no bloquea. |
| F2 | Inconsistencia | LOW | spec.md:210, 339, 350 · plan.md «Uso» · contracts/copia-con-un-paquete-mas.md:73-80 · tasks.md T005 | El spec habla de **cuatro** líneas de salida del guion (`copia`, `sincronía`, `tests enteros`, `subtests`); plan, contrato y tareas hacen **cinco**: añaden `orden del paso 2 de normativa, líneas más abajo: 1`. La línea extra sirve a US6.4 y el spec no la nombra. | Decir «cuatro líneas, más la de la orden del paso 2» en el spec o contar cinco en US6.1 y SC-003. No cambia el orden de ejecución. |
| F3 | Ambigüedad | LOW | tasks.md T003 · contracts/tests-genericos.md:86 | «`carpeta.Remove(eval.Fichero)` de cada eval de `LeerConjunto`» no dice que solo son las de la carpeta de `plazos-y-territorio`; `LeerConjunto` devuelve las de todas las skills, y `Remove` de una de otra skill sobre una raíz abierta en esta carpeta falla. El prototipo lo resolvió (research P1), pero el texto de la tarea no lo recoge. | Escribir «cada eval de `LeerConjunto` cuya skill es `plazos-y-territorio`». |
| F4 | Constitución (observación) | LOW | plan.md «Complexity Tracking» · constitución, principio III | «Aceptación e2e: no aplica» es una desviación del «cada hito empieza por el test e2e». Está justificada y el plan la registra (`precheck.sh:123` la admite y rechaza que haya además una tarea `[aceptacion]`; consistente con tasks.md, que no la lleva). No es violación sin justificar. | Ninguna. |
| F5 | Constitución (observación) | LOW | tasks.md T005 · constitución, capa 3 | Modificar ficheros existentes de `testdata/` lo revisa una persona al leer el informe. T005 es la única tarea `[datos]` y el spec lo dice en «Assumptions»; el guardián exige la etiqueta (`guardian-diff.sh:94`). Coherente. | Ninguna. |
| F6 | Cobertura | LOW | tasks.md T005, T006 | T005 concentra en una tarea el guion entero (programa, dos ficheros nuevos del archivo, filtros, comprobaciones, copia byte a byte) y T006, ~15 mutantes con `make ci`. Es más riesgo de agotar intentos que de incoherencia: no hay forma de partirlas sin romper «cada tarea deja `make ci` en verde» (research D6). | Sin cambio; es una observación para quien lee el informe si alguna entra en cuarentena. |

Sin hallazgos CRITICAL ni HIGH. Sin marcadores pendientes (`TODO`, `TKTK`, `???`, `<placeholder>`) en spec, plan ni
tareas. Sin duplicados entre requisitos.

## Coverage Summary

| Requisito | ¿Tiene tarea? | Tareas | Notas |
|---|---|---|---|
| FR-001, FR-002, FR-003 (`TestCoberturaDelRepositorio`) | Sí | T001 (T006 registra L06/L07) | |
| FR-004, SC-007 (L04/L05 en `h32-paquetes`) | Sí | T006 (T001 comprueba que el test sigue en verde) | `h32-paquetes` no cambia |
| FR-005 (línea del paso 2) | Sí | T002 (T006) | |
| FR-006, FR-007, SC-006 (`cuatro-evals-mas`, 16 evals) | Sí | T003 (T006) | F3 |
| FR-008 (`tope-con-las-evals-del-repositorio`) | Sí | T003, T006 (ocho evals de más) | sin cambio de código |
| FR-009 (`llevaNoCubierto`) | Sí | T004 (T006) | |
| FR-010, FR-011, FR-013 (copia con un paquete más) | Sí | T005 | F2 |
| FR-012, SC-005 (cada test devuelto a `main`) | Sí | T005 (construye), T006 (registra) | |
| FR-014, SC-001, SC-002 (`make ci`, `make skills-check`, diff) | Sí | T001-T005 (verde), T006 (Q1-Q3) | F1 |
| SC-003 | Sí | T005 | |
| SC-004 | Sí | T001-T004, T006 | |
| US1-US6 | Sí | US1 → T006; US2 → T001; US3 → T002; US4 → T003; US5 → T004; US6 → T005 | |

## Constitution Alignment Issues

Ninguna violación. Comprobado contra los principios I-X y las reglas R1-R7: ninguna petición HTTP, ninguna importación
nueva, ningún `.go` de producción en el diff previsto, ningún `t.Skip` ni `//nolint` nuevo, «Controles de umbral» con un
control con forma por fila y el criterio de decisión autónoma aplicado en las Clarifications `(auto)` y en
`gates/supuestos.md`. La desviación del principio III (F4) está en Complexity Tracking.

## Unmapped Tasks

Ninguna: T001-T006 se trazan a historias y requisitos.

## Metrics

- Requisitos funcionales: 14 (FR-001 a FR-014) · criterios de éxito: 7 (SC-001 a SC-007) · historias: 6
- Tareas: 6 (T001-T006; una `[datos]`, ninguna `[aceptacion]` por «Aceptación e2e: no aplica»)
- Cobertura de requisitos con al menos una tarea: 100 % (14/14 FR y 7/7 SC)
- Ambigüedades: 1 (F3) · Inconsistencias: 2 (F1, F2) · Duplicados: 0
- Problemas críticos: 0 · altos: 0 · medios: 1 · bajos: 5

## Next Actions

- No hay CRITICAL ni HIGH: se puede pasar a `/speckit-implement`.
- F1 es el único hallazgo con efecto sobre un control (el guardián por tarea no impide tocar `data/`, `evals/` ni
  `skills/` desde T001, T004, T005 ni T006); queda cubierto por Q3 en T006 y por `make skills-check`, que sí se ejecutan,
  pero el plan atribuye esa garantía al guardián.
- F2 y F3 son de redacción y no cambian ninguna tarea ni su orden.
- No se ofrecen remediaciones para después (modo autónomo); este informe es solo lectura y no modifica ningún artefacto.
