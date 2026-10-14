# Specification Analysis Report — H26 (`021-h26-el-umbral-de`)

Análisis de solo lectura de `spec.md`, `plan.md` y `tasks.md` (con `research.md`, `data-model.md`, `contracts/` y
`quickstart.md` como apoyo) contra la constitución 2.13.0. Modo desatendido: no se ofrecen remediaciones para después.

**Qué se ha ejecutado y visto en esta sesión**: lectura de los tres artefactos y de la constitución; `check-prerequisites.sh`
(FEATURE_DIR y `tasks.md` presentes); `grep` de cada test, ayudante y fichero que citan las tareas en `internal/evals`;
lectura de `internal/evals/umbrales.go` (orden de `umbralesDelInforme`, `umbralesDeLasClases`, `motivosDeLosUmbrales`,
`marcadasDelUmbral`), de las dos `clases.yaml`, del comentario de `.github/workflows/evals.yml` (líneas 148-158), de
`scripts/workflow/informe.sh` (búsqueda del control por nombre exacto, líneas 250-273) y de `CLAUDE.md`. **No se ha
ejecutado ningún test, ni `make ci`, ni el prototipo de research (que ya no existe)**: las cifras de research (tablas M)
y de los contratos no se han vuelto a medir aquí.

## Hallazgos

| ID | Categoría | Severidad | Ubicación | Resumen | Recomendación |
|----|-----------|-----------|-----------|---------|---------------|
| I1 | Inconsistencia | MEDIUM | spec.md SC-013 (L305) · FR-054 (L250) · contracts/documentos.md («CLAUDE.md», último párrafo) · tasks.md T004 · CLAUDE.md:23 | SC-013 exige «0 afirmaciones […] de que el job no aplica todavía el ADR 0038» en `CLAUDE.md`, sin cualificar. La viñeta de H25 (CLAUDE.md:23) dice «dentro de los umbrales del ADR 0038, que el job no aplica hasta H26», y contracts/documentos.md y T004 prohíben tocarla porque es historia de aquel cierre. FR-054 sí cualifica («donde lo dice como algo que vale hoy»); SC-013 no. Un juez de la revisión final que lea SC-013 al pie de la letra la marcará como afirmación que el hito deja falsa, y la corrección (editar la viñeta de H25) chocaría con T004 y FR-055. | Cualificar SC-013 como FR-054 («como algo que vale hoy») o nombrar la viñeta de H25 como excepción leída. Hoy: que la revisión final lea SC-013 a través de FR-054 y del contrato. |
| I2 | Inconsistencia | LOW | tasks.md T005 («Rutas») · plan.md «Project Structure» y «Obligaciones para tasks.md» | T005 se declara «sin código de producto» pero lista `internal/evals/umbrales.go` en «Rutas», lo que autoriza al guardián a aceptar un cambio en el código de producto. Solo hace falta para permitir `umbrales_test.go` (el guardián admite el `_test.go` de cada `x.go` declarado). Además, `cierre.md` no aparece en la estructura de documentación del plan (sí `mutantes.md`). | Aceptable como está (T005 dice que solo añade tests si la cobertura queda bajo su umbral); si se retoca, declarar `internal/evals/umbrales_test.go` y añadir `cierre.md` a la estructura del plan. |
| C1 | Cobertura | LOW | spec.md Edge Cases («El porcentaje escrito coincide con la condición y no se cumple», L193) · Clarifications · tasks.md | El caso 3 de 107 / 1 de 58 (porcentaje escrito igual al de la condición, `cumple: false`) no tiene caso de test en T001/T002. No hay requisito nuevo: la clarificación dice que no cambia cómo se escribe ningún umbral, y `medidaEscrita` y `condicionEscrita` (umbrales.go:567-579) ya usan el mismo `porcentajeConComa`. Los casos `104-con-3` (2,9 % frente a 2,8 %) ejercen la ruta. | Sin acción: queda cubierto por lo que ya hace el código y por `104-con-3`; la revisión final puede leerlo. |
| U1 | Infraespecificación / riesgo | LOW | tasks.md T001 | T001 concentra el código de `umbrales.go`, los comentarios de `doc.go` e `informe.go` y la actualización de siete funciones de test en un solo bloque. Está justificado (con los valores en 0 ningún veredicto cambia, y `make ci` tiene que quedar en verde por tarea) y las rutas son suficientes (`x_test.go` de cada fichero declarado), pero es la tarea con más superficie y la que más intentos del bucle puede consumir. | Sin acción: la división alternativa dejaría `make ci` en rojo a mitad (research D12). |
| A1 | Alcance (informativo) | LOW | spec.md «Fuera de alcance» último punto · `.github/workflows/evals.yml:152-154` | El comentario de `evals.yml` seguirá diciendo «doce umbrales, diez que deciden» y que `afirma_lo_no_leido` decide por modo, falso tras el hito. Está decidido y registrado (Clarification, «Pendiente para la persona»), coherente con FR-062 y con la regla «lo no especificado no se implementa». | Sin acción; queda en el informe final como pendiente. |
| V1 | Verificación no hecha | LOW | tasks.md T003 · research M9, M10, M13, M16 | Los mutantes de T003 se apoyan en un worktree bajo `$TMPDIR` y en las sustituciones de quickstart §6, que research dice haber probado sobre un prototipo ya inexistente. Aquí no se han aplicado. La propia tarea lleva la guarda («si dice que el mutante no se ha aplicado, se ajusta») y la condición de no marcarse si un mutante no hace fallar a las dos skills. | Sin acción: el riesgo está cubierto por la guarda de T003. |

Ningún hallazgo CRITICAL ni HIGH.

## Cobertura

Requisitos con tarea o con el cierre del workflow como control declarado. Los FR-0xx de cada grupo se listan juntos
cuando comparten tarea.

| Requisito | ¿Tarea? | Tareas | Notas |
|---|---|---|---|
| FR-001 a FR-005 (umbral conjunto) | Sí | T001, T002, T003 | Orden de `umbrales` (FR-005) coincide con `umbralesDelInforme` (respuestas por modo → medida → duraciones) y con contracts §1 |
| FR-010 a FR-012 (recuento por modo, trece y nueve) | Sí | T001, T002 | Recuento verificado a mano: `boe-legislacion` 13/9 y `jurisprudencia` 13/9 |
| FR-020 a FR-022 (valores, esquema, test de datos) | Sí | T002 (`[datos]`), T005 | `clases.yaml` lleva hoy `umbral: 0` en las dos clases de cada skill |
| FR-030 a FR-036 (lo que sigue igual) | Sí | T001, T002, T005 | Restricciones de batería |
| FR-040 (sondeo) | Sí | T001, T002 (sin cambiar), T004 | `sondeo.go` no se toca |
| FR-050 a FR-055 (documentos) | Sí | T004, T005 | `grep` en el árbol: los cinco ficheros son los únicos documentos de alcance con las frases viejas; `docs/USO.md`, ADR y `evals.yml` quedan fuera por decisión |
| FR-060 a FR-062 (dentro del run) | Sí | Batería de todas, T005 | FR-061 además lo leen los jueces de la revisión final |
| FR-070 a FR-079 (controles) | Sí | T001 a T005 | FR-079: el job de evals lo lanza el workflow |
| SC-001, SC-011 | Control del workflow | T001/T002 construyen los controles | Búsqueda por nombre exacto en `informe.sh:264` coincide con las 18 filas `evals:` del plan |
| SC-002 a SC-010, SC-013, SC-014 | Sí | T001 a T005 | SC-013, por lectura (ver I1) |
| SC-012 | No (humano, después del run) | — | Declarado por diseño |

Tareas sin requisito: ninguna. T005 (cierre) se justifica por la Definition of Done.

**Verificaciones cruzadas hechas** (todas contra el árbol): los 13 tests nombrados existen en el fichero que dicen, salvo
`TestUmbralConjunto`, que el plan fija como nuevo; los ayudantes que T001/T002 citan (`umbralesConJuez`,
`umbralesDeUnModo`, `exigirLosDoceDelContrato`, `ejecucionConUmbrales`, `cortarTrasLaRespuesta`, `nombreDelUmbralDeClase`,
`marcadasDelUmbral`, `nombradasEn`, `compararUmbral`, `deLosModos`, `marcadasEn`) existen; las nueve filas `evals:` por
skill del plan suman las que FR-071 pide y no hay fila para el recuento por modo; el motivo de `motivosDeLosUmbrales`
omite hoy los umbrales con `decide: false`, de modo que el recuento por modo no dará motivo; las citas de línea de
`contracts/documentos.md` son las de `869eaf8`, que la propia nota advierte.

## Constitución

Sin conflictos. Comprobado contra los gates de «Umbrales que se cumplen» (ADR 0029/0037/0038): hay control con forma para
cada umbral que decide (fila `evals:` y fila `ci:` en el plan), mutante por umbral conjunto (FR-078, T003), y ninguna
corrección puede rebajar un umbral (FR-061). Principio III: «Aceptación e2e: no aplica» está justificado en Complexity
Tracking (el binario no enlaza `internal/evals`, research V14) con precedente H24/H25. Principio V: sin dependencias
nuevas, sin tipos ni ficheros Go nuevos. Principio VIII: no cambia ninguna skill. Reglas del modo desatendido: ninguna
tarea abre sesiones con modelo ni usa red salvo la de las herramientas de Go; rutas declaradas por tarea.

## Métricas

- Requisitos funcionales: 38 (FR-001–005, 010–012, 020–022, 030–036, 040, 050–055, 060–062, 070–079); criterios de éxito: 14.
- Tareas: 5 (T001 a T005); ninguna `[P]`; ninguna `[aceptacion]`; una `[datos]` (T002).
- Cobertura de FR con ≥ 1 tarea: 38/38 (100 %). SC con tarea: 12/14; SC-001/SC-011 los mide el cierre del workflow (con sus controles construidos por T001/T002) y SC-012, una persona.
- Ambigüedades: 0 marcadores pendientes (el `precheck` de plan ya lo dio y `tasks.md` no lleva ninguno).
- Duplicaciones: 0.
- Issues CRITICAL: 0 · HIGH: 0 · MEDIUM: 1 (I1) · LOW: 5 (I2, C1, U1, A1, V1).

## Siguientes acciones

Sin CRITICAL ni HIGH: se puede pasar a `/speckit-implement`. El único MEDIUM (I1) no bloquea: es una lectura de SC-013
que la revisión final resuelve si lee FR-054 y `contracts/documentos.md`. Si se prefiere cerrarlo antes, es una frase en
SC-013 (cualificar como FR-054) y se haría con `/speckit-specify` sobre el spec, no desde este análisis.

No se aplica ninguna edición desde este paso (solo lectura; este fichero es el registro del análisis).

## Extension hooks

`.specify/extensions.yml` lleva `before_analyze` con `speckit.git.commit`, **opcional** («Commit outstanding changes
before analysis?»). En modo desatendido no se ejecuta: este paso no modifica el árbol salvo este informe, y confirmar o
commitear lo decide el workflow.
