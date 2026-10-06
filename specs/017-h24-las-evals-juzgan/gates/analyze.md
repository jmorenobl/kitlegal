# Specification Analysis Report — H24 (`017-h24-las-evals-juzgan`)

Análisis de `spec.md` (448 líneas), `plan.md` (426) y `tasks.md` (442) contra la constitución y el repositorio. Solo lectura,
salvo este fichero. **Lo que se ejecutó en esta sesión**: `check-prerequisites.sh`, lecturas de ficheros y `grep`/`git cat-file`/
`git show` locales. **Lo que no se ejecutó**: `make ci`, ningún test, `claude`, `npm`, `python3`, ningún guion `evals*.sh`; las
cifras de tiempo (8 s por voto, 40 s de tope, 60 s de instalación) y las banderas de Claude Code 2.1.289 no las he podido medir
y se tratan como supuestos del research (S1–S8), no como evidencia.

Hooks de extensión: hay `before_analyze` y `after_analyze` (`speckit.git.commit`), los dos `optional: true`. En modo autónomo no se
preguntan ni se ejecutan; no se ha hecho ningún commit.

## Findings

| ID | Category | Severity | Location(s) | Summary | Recommendation |
|----|----------|----------|-------------|---------|----------------|
| F1 | Inconsistency | MEDIUM | spec.md «Relación con H5.1 a H22», FR-024, Assumptions («Los 259 casos se reconstruyen con las grabaciones del repositorio»); plan.md «Decisiones» (D15), Complexity Tracking fila 3; tasks.md T008, T009 | H7.2 FR-020 mandó que la eval 19 anterior, su grafo previo, su derivada y su párrafo sintético **no existieran ni los nombrara ningún fichero de código o de test** (`specs/012-h7-2-la-consulta-repetida/spec.md:208`). T008 los restaura bajo `testdata/evals/retiradas/` y T008/T009 los nombran en código y en tests (`grafo_test.go`, la constante de la carpeta en `grabaciones.go`). El plan y los supuestos lo registran como desviación, pero el spec sigue diciendo que todo está en `main` y su lista «Sustituye» no incluye H7.2 FR-020. Ningún test del repositorio impone hoy esa regla (`grep` del nombre antiguo: 0 coincidencias en `.go`), así que no hay rojo, pero la decisión queda sin trazar en el spec. | Añadir H7.2 FR-020 a «Sustituye» (y corregir el supuesto de los 259 casos) en el spec, o dejar constancia en `cierre.md` (T017) de que H24 lo sustituye a sabiendas. |
| F2 | Underspecification | MEDIUM | tasks.md T005 («Salen las subpruebas `expresiones-calibradas`… con lo que solo ellas usan») y T005 `TestJuzgarSinLaLista`; `internal/evals/conjunto_test.go:1555-1630` | T005 retira los subtests de calibrado «con lo que solo ellas usan», que incluye `informesCalibrados`, `informeDeH71…` y `marcadasEnElInforme`; y a la vez pide que `TestJuzgarSinLaLista` use «los informes versionados de H7.1, H7.2 y H7.3» (los mismos tres de ese calibrado). La tarea no dice si esas rutas y el lector de entradas se conservan para el test nuevo. FR-111 lo llama «el calibrado de H7.4»; T005, «informes de H7.1–H7.3»: mismo conjunto con dos nombres. | Decir en T005 qué helper (rutas de los tres informes, lectura de sus respuestas) se queda para `TestJuzgarSinLaLista`; unificar el nombre («los tres informes del calibrado de H7.4»). |
| F3 | Underspecification / riesgo | MEDIUM | plan.md «Decisiones» (tope de un voto 35 s + 5 s, sin reintento), research D6/S6; spec FR-007, FR-051–053; `evidencias/adr-0037/guiones/juez.py:69-109` | La medida versionada se hizo con `voto_real` con `timeout=900` y **esperando y repitiendo** ante un límite de uso (`votar`, línea 108). El código de H24 usa 35 s (+5 s) y, ante cualquier voto que no llega, deja la respuesta «sin juzgar» sin reintentar (FR-007, fuera de alcance reintentar). En la ejecución de la medida (683 votos) un solo voto lento o un límite de uso da «caso sin juzgar» y **no se imprime medida** (FR-053), lo que bloquea SC-014 y obliga a relanzar. La propia plan lo reconoce («su cola no está medida», S6) y la decisión es consciente; no hay ninguna medida de cuántos votos de la validación pasaron de 35 s. | Aceptable como está si se asume. Los votos versionados (`votos-medida.jsonl`) no guardan la duración de cada voto (0 coincidencias de «duracion»), así que no hay forma de comprobar sin modelo cuántos pasaron de 35 s; solo lo mide el primer job. No bloquea. |
| F4 | Coverage gap | LOW | `.github/workflows/evals.yml` (job `cambios`, `case` de rutas, líneas 52-60); tasks.md T003, T012, T013 | El filtro de «qué toca la propuesta» lista `scripts/evals.sh` y `scripts/evals-sesion.sh` pero no `scripts/evals-voto.sh` ni `scripts/evals-medir-juez.sh`, que T003/T012 crean. Una propuesta que solo cambie el guion que abre el voto no arrancaría el job al abrirse (la etiqueta `evals` sí). Ninguna tarea lo toca; T013 ya edita `evals.yml`. | Añadir las dos rutas al `case` en T013 (o declarar que se deja). |
| F5 | Underspecification | LOW | plan.md «Performance Goals» / contracts/job-de-evals.md §4; tasks.md T013; research S4 | Peor caso de `evals` = 21 097 s → `timeout-minutes: 352`, a 8 minutos del tope de 360 de un trabajo de GitHub. `TestDefinicionDelJob` comprueba que cada tope **cubre** su peor caso, pero no que el tope no pase de 360; la siguiente eval añadida (o una concurrencia menor) dejaría un peor caso imposible de cubrir sin que ningún control lo diga. | Opcional: añadir a T013 un caso de `TestDefinicionDelJob` «peor caso > 360 min = fallo». |
| F6 | Inconsistency (cifras) | LOW | spec.md «Uso, de fuera adentro» (líneas 394-401) frente a plan.md «Uso» (278-289) | Las magnitudes difieren: elementos de `umbrales` ~300 B (spec) / ~325 B (plan); aviso de lo que no cubre ~150 B / 80 B; medida impresa ~700 B / 656 B; máximo del bloque `juez` ≈ 800 KB / 530 KB. El plan las midió (research M4); el spec las estimó. No afecta a ninguna tarea. | Dejar la del plan como la vigente; opcionalmente alinear el spec. |
| F7 | Underspecification | LOW | tasks.md T015 «Rutas» | Rutas declara `informe_test.go`, `juzgar_test.go`, `sondeo_test.go`, `umbrales_test.go` aunque el `grep` de `otra_conversacion` solo encuentra listas sintéticas en `informe_test.go`, `formato_test.go`, `conjunto_test.go` y `prohibidas.go`. Es más ancho de lo comprobado (inofensivo para el guardián de diff, pero la tarea no dice qué cambia en esos tres). | Reducir a lo que cambie o decir qué se alinea en cada uno. |
| F8 | Constitution (informativo) | LOW | plan.md «Aceptación e2e: no aplica», Complexity Tracking fila 1; tasks.md cabecera | Principio III («cada hito empieza por el test e2e») queda sin tarea `[aceptacion]`. Está declarado y justificado (el hito no cambia el binario) y tiene precedente en H7.2–H7.4; no es violación de un MUST. | Ninguna. |

Sin hallazgos CRITICAL ni HIGH. Ninguno de los puntos anteriores contradice una regla del «Criterio de decisión autónoma» ni de
«Reglas del modo desatendido»: ninguna tarea usa red, `KITLEGAL_RECORD`, abre sesiones con modelo ni escribe en `evidencias/` o
`scripts/workflow/`; las tres tareas que tocan `schemas/` o `testdata/` (T001, T008, T015) llevan `[datos]`; cada tarea declara
rutas; el orden es secuencial y sin dependencias hacia delante.

## Coverage Summary

71 requisitos funcionales (FR-001–007, 010–014, 020–024, 030–037, 040–045, 050–054, 060–062, 070–071, 075–076, 080–087, 090–093,
095–096, 100–113) y 14 criterios (SC-001–014). 17 tareas (T001–T017).

| Requisito | ¿Tarea? | Tareas | Notas |
|---|---|---|---|
| FR-001, 002, 004, 005, 006, 007, 010, 011 | Sí | T002 (+T003 FR-003/004/007) | Mensaje, frase, voto nulo y regla, sin proceso |
| FR-003, FR-004 (orden) | Sí | T003 | Orden y entorno idénticos a `voto_real` (comprobado en `juez.py:69-82`) |
| FR-012, 013, 014 | Sí | T006 (T007 publica lo de FR-013) | Eval 21 aparte y sin umbral |
| FR-020–023 | Sí | T001 | Carpeta, esquema, copias |
| FR-024 | Sí | T008, T009 | F1 |
| FR-030–033, 035, 037 | Sí | T006 | Doce umbrales y motivos |
| FR-034 | Sí | T005, T006 | |
| FR-036, 045, 093, 096 | Restricción de toda tarea | batería; T016/T017 lo constatan | |
| FR-040–042 | Sí | T001 (copia), T004 | |
| FR-043, 044 | Sí | T006, T011 | |
| FR-050–054 | Sí | T010, T012, T013 | |
| FR-060, 061 | Sí | T007 | |
| FR-062, 070, 071 | Sí | T005 | F2 |
| FR-075, 076 | Sí | T014 | |
| FR-080 | research/plan (hecho) | T015 aplica C1–C8 | |
| FR-081–086 | Sí | T015 | Los mide el job de cierre |
| FR-087, 095 | Sí | T016 | |
| FR-090–092 | Sí | T004, T013 | F5 |
| FR-100 | Todas | T017 | |
| FR-101 | Plan («Controles de umbral») | — | 17 filas en el plan, las 6 de FR-101 incluidas |
| FR-102–112 | Sí | T002–T015 según la tabla de tasks.md | Todos con test nombrado como `^func Test…\(` |
| FR-113 | T016 y el job del cierre | | |
| SC-001 | Cierre del workflow | | No es tarea |
| SC-002–013 | Sí | tabla de tasks.md | |
| SC-014 | Persona, después del run | | No es tarea |

**Unmapped tasks**: ninguna. Las 17 citan FR/SC.

**Constitution alignment issues**: ninguno de nivel MUST. Informativo: F8.

## Metrics

- Requisitos (FR): 71; criterios (SC): 14; tareas: 17
- Cobertura de FR con ≥ 1 tarea o control declarado: 71/71 (100 %); los que no tienen tarea propia lo son por diseño (FR-080,
  FR-101, SC-001, SC-014)
- Ambigüedades: 2 (F2, F3)
- Duplicaciones: 0 (F6 es deriva de cifras, no duplicación)
- Hallazgos CRITICAL: 0 · HIGH: 0 · MEDIUM: 3 · LOW: 5

## Next Actions

- No hay CRITICAL: se puede ir a `/speckit-implement`.
- Antes, si se quiere dejar limpio el rastro: F1 (declarar la sustitución de H7.2 FR-020) y F2 (nombrar qué helper de calibrado
  conserva T005) son ediciones de una o dos líneas en `spec.md` y `tasks.md`.
- F3 y F5 son riesgos conocidos y asumidos en el plan; se resuelven o no con una decisión de la persona, no con una edición.
- F4, F6 y F7 son mejoras menores; ninguna cambia el orden ni la cobertura.
