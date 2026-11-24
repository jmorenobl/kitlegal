# Specification Analysis Report — H45 (specs/042-h45-el-molde-admite)

Solo lectura, sobre `spec.md`, `plan.md` y `tasks.md` (con `research.md` y `contracts/aceptacion.md` donde hacía falta) y la
constitución. Nada de esto se ha escrito fuera de este informe.

## Qué se ha medido en esta sesión y qué no

Visto en esta sesión, sobre el repositorio:

- `skills/normativa/SKILL.md` líneas 38-41 son «En PowerShell 5.1:», las dos vallas y la orden: lo que T004 mueve (FR-009).
- Los mutantes `orden-con-su-forma-en-otra-referencia` y `sin-las-ordenes-de-articulos` ya existen en
  `internal/evals/conjunto_test.go:5244` y `:5259`; `sesionesALaVezComoMucho` vale hoy 11 (`definicion_test.go:162`), como
  dice FR-021.
- Los ficheros que llaman a `UnionDeGrabaciones` o `GrabacionesDeH36` (`grep`): `medida.go`, `sesiones.go`,
  `dependencias.go`, `grabaciones.go` y, en test, `conjunto_test.go`, `dependencias_test.go`, `grabacion_test.go`,
  `grabaciones_test.go`, `juzgar_test.go`, `laboratorio_test.go`, `preparar_test.go` e `internal/skills/normas_test.go`.
  Son exactamente los que enumera T007.
- El texto de FR-023 mide 429 caracteres (`wc -m` sobre el texto sin salto). Las variantes de hoy suman 1 694 con sus
  cuatro saltos, es decir 1 690 caracteres; con el texto nuevo, 1 692, lo que dice T017.
- Los cuerpos de `SKILL.md` de las tres skills de capacidad (`jurisprudencia`, `normativa`, `plazos-y-territorio`) son las
  que usa T006; la cuarta, `consulta-juridica`, no es de capacidad.

No medido (no se ha ejecutado nada de esto): ningún test, `make ci`, `make skills-sync`, el guion de aceptación ni ningún
mutante; los tiempos y el tamaño de la copia de `paquete-de-mas.txtar` son los de `contracts/aceptacion.md` y `research.md`
(V10, P3), no los de esta sesión; tampoco el guardián de diff del workflow (si admite que una tarea `[datos]` cambie las
copias congeladas de hitos anteriores: ver D1).

## Hallazgos

| ID | Categoría | Severidad | Ubicación | Resumen | Recomendación |
|----|-----------|-----------|-----------|---------|---------------|
| I1 | Inconsistencia | MEDIUM | tasks.md:129-130 vs 141, 150-152 | La meta de la fase 9 dice que «las descripciones (T010, T011) ya están». Las descripciones son T016 y T017, de la fase 10, que va después; T010 y T011 son los tests de `internal/skills` y los subtests de `TestEvalsDelRepositorio`. El propio punto de control (l. 141) dice lo contrario: «salvo los de la descripción, que cambian en la fase siguiente». | Corregir la frase a «(T008, T009)» sin las descripciones, o a «las descripciones, T016 y T017, vienen después». No cambia el orden de ejecución, que sale de las dependencias. |
| I2 | Inconsistencia | LOW | spec.md:353 (FR-012), 590-599 (§C.3) vs tasks.md:71 (T004) | FR-012 dice que los mutantes «crecen con los casos de §C.3, los cinco y ninguno más». La tabla §C.3 tiene seis filas (una sin defecto y cinco con él) y dos de ellas (`orden-con-su-forma-en-otra-referencia`, `sin-las-ordenes-de-articulos`) ya están en el repositorio. T004 añade cuatro y conserva dos, que es consistente con la tabla pero no con «crecen con los cinco». | Leer «los cinco» como los cinco con defecto de la tabla; T004 ya lo cumple (cuatro nuevos más los dos que existen, y ningún sexto caso). Sin cambio en tareas. |
| I3 | Inconsistencia | LOW | tasks.md:215-228 | El ejemplo paralelo de la fase 9 lista T010, T012, T013 y T014 y omite T011 y T015; la línea de «tras T006, en la fase 6-8» deja T009 fuera de la lista de `[P]` de esa fase aunque la fase 8 existe. El texto de cierre ya dice que el bucle ejecuta una a una por id, así que no hay efecto en la ejecución. | Opcional: completar el ejemplo. |
| D1 | Cobertura / proceso | LOW | tasks.md:139 (T015), spec.md:381-384 (FR-017) | T015 cambia copias congeladas de hitos anteriores (`specs/030-…/paquetes.txtar`, `specs/039-…/cobertura-instalada.txtar`, `specs/041-…/plazos-de-hoy.txtar`) y la constitución congela la suite de aceptación y protege `testdata/`. La tarea se declara `[datos]` y FR-017 dice que el guardián de diff las admite; no he leído los scripts del workflow para confirmarlo. | No se cambia nada: es una comprobación del guardián (FR-017). Si el guardián no admite esas rutas, T015 entraría en cuarentena. |
| U1 | Subespecificación | LOW | tasks.md:35 (T001) | T001 pide «comprobar que cada guion falla hoy por una aserción» pero los guiones viven en `specs/…/aceptacion/` y solo corren dentro de `TestEntregaDelHito` una vez activados (`aceptacion.sh activar`); la tarea no dice cómo se hace esa comprobación antes de implementar. El «rojo primero» del workflow lo cubre, y `contracts/aceptacion.md` §1 detalla el rojo esperado (`sincronía: 1`). | Sin cambio obligatorio; el ejecutor puede activarlos en una copia o dejarlos al «rojo primero» del workflow. |
| U2 | Subespecificación | LOW | tasks.md:35, contracts/aceptacion.md §1 | `paquete-de-mas.txtar` lanza dos `go test` anidados sobre una copia de ~150 MB (cifras de `research.md`, no medidas aquí) dentro de `make ci`; ningún artefacto fija un tope de tiempo para esa fase ni qué pasa si el reloj de CI es peor. | Sin cambio: el coste está declarado en el contrato; si el CI lo sufre, es dato para H42. |

Sin hallazgos CRITICAL ni HIGH. Sin duplicados entre requisitos; marcadores pendientes (`TODO`, `???`, `<placeholder>`)
en `spec.md`, `plan.md` y `tasks.md`: ninguno que sea un hueco (los `<…>` que aparecen son los de los textos de
contrato, p. ej. `<código>`, `<vertical>`).

## Constitución

Sin conflictos.

- *Suite de aceptación primero y congelada*: T001 es la primera tarea y escribe los dos guiones desde el spec; el cambio
  de arnés (`GOCACHE`, `GOMODCACHE` en `e2e_test.go`) está declarado en la misma tarea y en `plan.md` (l. 132, 202, 245).
- *Sin red, sin `KITLEGAL_RECORD`, sin modelo, ningún `data/` escrito a mano*: las reglas de cabecera de `tasks.md` lo
  repiten y ninguna tarea toca `data/` ni `evidencias/`.
- *Cambios a `testdata/` existentes* (T005, T015, T017): van como `[datos]`, que la persona revisa después (capa 3).
  T017 mezcla otros ficheros en la misma tarea y lo justifica (el binario empotra la descripción y `h41-4` la ve en el
  mismo commit), que es la excepción que prevé la definición de `[datos]`.
- *Umbrales que se cumplen* (ADR 0029): la sección «Controles de umbral → tarea» de `tasks.md` da una fila por requisito
  numérico y un control con forma; SC-010 (el cierre de evals) lo mide el job y no una tarea, y así lo dice la tabla.
- *Mutantes en una copia en `$TMPDIR`*, nunca en el árbol: regla de cabecera y T018.
- *Cobertura en positivo y genericidad por materia (principios II y X)*: el hito no añade ninguna norma ni materia; la
  vertical de ensayo es de prueba (`testdata/`).

## Cobertura de requisitos

| Requisito | ¿Tarea? | Tareas | Notas |
|---|---|---|---|
| FR-001 (paquete de un tipo) | Sí | T003 | |
| FR-002 (región solo con sus tipos) | Sí | T003, T006 | mutante «tipo asignado a todas» en T006 |
| FR-003, FR-004 (defectos de la sincronía) | Sí | T003 | `tipo-sin-paquete`, `tipo-de-dos-paquetes`, `computo-de-otro-paquete` |
| FR-005, FR-006 (paquete de una fila, `filasSinRegion`) | Sí | T006 | |
| FR-007 (línea de la fila sin cambios) | Sí | T006 | sin tarea propia: se comprueba con la línea que no cambia |
| FR-008 (lo de hoy, igual) | Sí | T003, T009, T018 | `make skills-sync` y contra `main` |
| FR-009, FR-010 (movimiento en `normativa` y su declaración) | Sí | T004 | |
| FR-011, FR-012, FR-013 (control de PowerShell, mutantes, `docs/SKILLS.md`) | Sí | T004 | |
| FR-014 (laboratorio sin referencias de un paquete) | Sí | T009 | |
| FR-015 (grabaciones por hito) | Sí | T007 | 11 ficheros de llamadas, todos listados |
| FR-016 (tests sin listas) | Sí | T002, T003, T006, T008-T015 | |
| FR-017 (guiones de `internal/app`) | Sí | T015, T017 | |
| FR-018 (reglas del conjunto de `normativa` sin cambio) | Sí | T011 (no toca `conjunto.go`), T018 | |
| FR-019, FR-020 (vertical de ensayo crecida) | Sí | T005, T006 | |
| FR-021, FR-022 (tope del trabajo de evals) | Sí | T008 | |
| FR-023 (descripción de 429) | Sí | T016, T017, T001 | 429 verificado en esta sesión |
| FR-024 (CHANGELOG) | Sí | T008, T017 | |
| FR-025 (ninguna eval cambia) | Sí | T018 | `git diff main -- evals/` vacío |
| FR-026 (`kitlegal plazos` igual) | Sí | T014, T015, T018 | |
| FR-027 (`mutantes.md`) | Sí | T018 | |
| SC-001 a SC-009, SC-011 | Sí | ver tabla de controles de `tasks.md` | |
| SC-010 (evals del cierre) | No (a propósito) | — | no es trabajo construible; lo mide el cierre del workflow; T008 y T017 dejan lo que juzga |

Historias: US1 sin tarea propia (la explica `tasks.md`: T003, T006, T014, T015, T017, T018); US2 T005-T006; US3 T003;
US4 T010-T015 y T001; US5 T004 y T001; US6 T009; US7 T007; US8 T008; US9 T016-T017.

## Tareas sin requisito

Ninguna: cada una cita FR, SC o historia. T002 es fundacional de FR-016 (datos de `TestRegiones`).

## Métricas

- Requisitos: 27 FR + 11 SC (10 construibles; SC-010 es del cierre).
- Tareas: 18 (T001-T018).
- Cobertura: 27/27 FR (100 %); 10/10 SC construibles (100 %).
- Ambigüedades: 1 (I2, redacción de «los cinco»).
- Duplicados: 0.
- Inconsistencias: 2 (I1, I3). Subespecificaciones: 2 (U1, U2). Cobertura/proceso: 1 (D1).
- Críticos: 0. Altos: 0.

## Siguientes pasos

- No hay CRITICAL ni HIGH: puede pasarse a la implementación.
- I1 es la única corrección de redacción que merece hacerse en `tasks.md` (una frase de la fase 9); el resto son
  observaciones que no cambian el orden ni el alcance de ninguna tarea.

## Extension Hooks

Hooks registrados en `.specify/extensions.yml` (`before_analyze` y `after_analyze`), ambos opcionales y de la extensión
`git`: `/speckit-git-commit` («Commit outstanding changes before analysis?» y «Commit analysis results?»). No se han
ejecutado: esta sesión es de solo lectura y la rama tiene sin confirmar `gates/modelos.jsonl` y `tasks.md`; los commits
los hace el workflow.
