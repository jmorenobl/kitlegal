---

description: "Tareas de H31.1 · boe-legislacion habla de la norma, no de la consulta, y lee lo que nombra"
---

# Tasks: H31.1 · `boe-legislacion` habla de la norma, no de la consulta, y lee lo que nombra

**Input**: `specs/029-h31-1-boe-legislacion-habla/` — spec.md, plan.md, research.md, data-model.md, quickstart.md y
contracts/ (`aceptacion.md`, `lecturas-y-grabaciones.md`, `skill-md-v0.3.1.md`, `skill-boe-legislacion-v0-3-1.md`,
`guia-y-changelog.md`).

**Tests**: la suite de aceptación la escribe T001 y queda congelada; cada tarea posterior lleva su test o el control de
`make ci` que la ve fallar, y termina con `make ci` en verde (SC-001). Ninguna tarea usa la red, el modelo ni
credenciales; las grabaciones las hace el paso `grabar_datos` del workflow tras el commit de T002.

**Formato**: `- [ ] Tnnn [P?] [etiqueta?] [USn?] Descripción con todas las rutas que toca`. El guardián rechaza
cualquier fichero fuera de las rutas declaradas; `go.mod`, `go.sum`, `CHANGELOG.md` y el directorio del feature
siempre están permitidos.

**La skill medida**: es la del sondeo del 26 de octubre, cuya ruta dan plan.md («Summary») y
`contracts/aceptacion.md` (líneas 6-7). Se lee; ninguna tarea la escribe.

## Fase 1: Aceptación (SC-003)

**Objetivo**: fijar, antes de tocar nada, lo que la skill de la entrega conserva de la medida.

- [X] T001 [aceptacion] Escribir los tres guiones testscript de la suite copiando, palabra por palabra, los tres textos de `specs/029-h31-1-boe-legislacion-habla/contracts/aceptacion.md` §3 (sin el prefijo `zz-`) en `specs/029-h31-1-boe-legislacion-habla/aceptacion/h31-1-responder.txtar` (US1, escenario 1 · FR-002 · SC-003), `specs/029-h31-1-boe-legislacion-habla/aceptacion/h31-1-lee-lo-que-nombra.txtar` (US2, escenario 1 · FR-002 · SC-003) y `specs/029-h31-1-boe-legislacion-habla/aceptacion/h31-1-sitio-en-el-cuerpo.txtar` (US3, escenarios 2 y 3 · FR-001, FR-003, FR-004, FR-006, FR-007, FR-009, FR-010 · SC-003); cada guion nombra en su comentario «Cubre:» sus US, FR y SC, usa solo `exec awk`, `exec ls` y órdenes de testscript, compara la skill de `$KITLEGAL_SKILLS/boe-legislacion/` con la medida y deja la aserción en `stdout`/`cmp`/`! grep`; con la skill de hoy fallan por esas aserciones («distinto: 6 líneas frente a 4», «distinto: 12 líneas frente a 4», «encabezados: distinto: 14 frente a 13»; research V5). Sin código de producto y sin evals nuevas: el hito no añade evals, y las nueve `lecturas` de FR-011 van en T003 porque `TestEvalsDelRepositorio/grabado` y `TestCacheDeLasSesiones/lecturas` las exigen grabadas (research V11), y las grabaciones llegan con `grabar_datos` tras T002

**Checkpoint**: el workflow comprueba el rojo de cada guion (`aceptacion.sh rojo-primero`) y los congela.

## Fase 2: US4 — Las evals por materia sirven lo que sus respuestas nombran (P1)

**Objetivo**: que la caché de cada sesión de las evals 14, 15 y 17 sirva sin red cada bloque que la regla de leer lo
que se nombra pide (FR-011 a FR-013; SC-004, SC-008).

**Prueba independiente**: `go test -count=1 -v -run '^(TestEvalsDelRepositorio|TestCacheDeLasSesiones|TestLecturaDeLaEvalSinGrabar)$' ./internal/evals/`
con `PASS` en `lecturas`, `grabado` y `TestLecturaDeLaEvalSinGrabar` (quickstart §6).

- [X] T002 [datos] [US4] Ampliar el manifiesto `internal/evals/testdata/grabaciones.json` (`"fuente": "boe.legislacion-consolidada"`) según `specs/029-h31-1-boe-legislacion-habla/contracts/lecturas-y-grabaciones.md` §2: en la entrada `Ley 35/2006,`, `bloques` pasa a `["a18", "a7", "a19", "a20", "dadecimoctava"]`; en `Real Decreto Legislativo 2/2004,`, a `["a15", "a60", "a92", "a100", "a104"]`; el `para` de cada una gana la frase de §2 antes del punto final, si lo hay; la entrada `Ley 19/2013,` (con `a24`) y el resto del fichero no cambian (FR-012). Sin otro trabajo: el test de grabación es `TestGrabarEvals` de `internal/evals/grabacion_test.go` (`//go:build grabacion`), que ya existe y no cambia, y las ocho respuestas en `internal/evals/testdata/boe.legislacion-consolidada/` las graba `grabar_datos` tras el commit. Verificación: `go test -count=1 -run '^(TestManifiestoDeGrabaciones|TestIdentificadoresDeLasNormas|TestGrabacionesSinSolape)$' ./internal/evals/` y `make ci` en verde (el manifiesto lo lee `grabar_datos`; `grabado` comprueba lo grabado, research V11)
- [X] T003 [US4] Añadir las nueve lecturas de FR-011 con su test, en rojo primero: en `internal/evals/conjunto_test.go` (`probarLasLecturasDeLasEvalsDelRepositorio`, subtest `lecturas` de `TestEvalsDelRepositorio`) el mapa escrito a mano gana, con una constante para el `motivo` de `contracts/lecturas-y-grabaciones.md` §1, `a7`, `a19`, `a20` y `dadecimoctava` de `BOE-A-2006-20764` en la 15, `a60`, `a92`, `a100` y `a104` de `BOE-A-2004-4214` en la 14 y la eval 17 con `a24` de `BOE-A-2013-12887`; en `internal/evals/preparar_test.go` (`probarLasLecturasDeLasEvals`, subtest `lecturas` de `TestCacheDeLasSesiones`) la tabla gana los nueve bloques, servidos con `boe articulo … --offline` sin fecha y con `--fecha 9999-12-31`, con código 0; los comentarios que enumeran las evals con lecturas ganan la 17; ver el subtest `lecturas` de `TestEvalsDelRepositorio` en rojo; después, las nueve entradas `{norma, bloque, motivo}`, detrás de las de hoy y en el orden del contrato §1, en `evals/boe-legislacion/15-irpf-rendimientos-por-materia.yaml`, `evals/boe-legislacion/14-trlrhl-impuestos-por-materia.yaml` y, en una clave `lecturas` nueva al final, `evals/boe-legislacion/17-ltaibg-plazo-por-materia.yaml`, sin otro cambio en `evals/` (FR-011, FR-013; SC-004, SC-008). Verificación: quickstart §6 (seis `PASS`, ocho grabaciones) y §9 (tres ficheros, 12, 12 y 4 líneas añadidas, 0 borradas), `TestDefinicionDelJob` en verde sin tocar `.github/workflows/evals.yml` (FR-015, SC-005; research D8) y `make ci` en verde

**Checkpoint**: las evals 14, 15 y 17 llevan sus lecturas y la caché de sus sesiones las sirve sin red.

## Fase 3: US1, US2 y US3 — La skill v0.3.1 (P1)

**Objetivo**: el `SKILL.md` de la entrega, que conserva los pasajes medidos (US1: paso 4, primera viñeta, y paso 6;
US2: pasos 3 y 5) y cabe en 11 000 caracteres sin sacar nada a una referencia (US3).

**Prueba independiente**: quickstart §1 a §3, §5 y §8; con la suite activada, los tres guiones de T001 en `PASS`
(quickstart §4).

- [X] T004 [US3] Sustituir `skills/boe-legislacion/SKILL.md` por `specs/029-h31-1-boe-legislacion-habla/contracts/skill-md-v0.3.1.md` byte a byte (cuerpo de 10 995 caracteres; frontmatter, tabla de comandos, formas fijas, encabezados `##`/`###` y número y encabezado de cada regla como en la skill medida; los pasos 3, 4 primera viñeta, 5 y 6 idénticos a los suyos; sin título `#`) y `skills/boe-legislacion/references/buscar_la_norma.md` y `skills/boe-legislacion/references/redaccion_modificada.md` por las de la skill medida, también byte a byte, con lo que la #200 cambió en ellas; ninguna referencia nueva ni otro fichero de `skills/boe-legislacion/` cambia (FR-001 a FR-007, FR-009, FR-010; US1, US2, US3; SC-002, SC-003). En la misma rebanada, por consecuencia de FR-001 (research D10; contrato de la skill §6), en `CONTRIBUTING.md` (líneas 742-743, partida en dos líneas) «la frase de la regla 7» pasa a «la frase de `references/lecturas_que_fallan.md`» y nada más del fichero cambia. Verificación: `make skills-sync` sin cambios en lo generado, `make skills-check` en verde (gate del cuerpo de H31 —`TestCuerpoDeLaSkill` lo ve fallar con 11 001 caracteres—, `referencias`, tabla sin drift, `normas-nombradas`, `sin-instrucciones-de-evals` y los ocho controles de contenido de `TestEvalsDelRepositorio`, `ordenes-de-ejemplo` incluido; FR-004, FR-005, FR-007, FR-008), `TestSkillsEmpotradas` en verde, quickstart §2 (`igual`), §3 (`iguales`) y §8 (`0`), que son las comparaciones que los tres guiones de T001 hacen (research V3 a V5), y `make ci` en verde

**Checkpoint**: la skill de la entrega es la del contrato y los gates de H31 siguen en verde sobre ella.

## Fase 4: US6 — La guía dice qué referencias se leen (P2)

**Objetivo**: que quien escribe una skill sepa que lo que toda respuesta necesita va en el cuerpo (FR-014; SC-013).

**Prueba independiente**: quickstart §10.

- [X] T005 [P] [US6] Añadir a `docs/SKILLS.md` §1 el párrafo «Lo que se lee y lo que no» de `specs/029-h31-1-boe-legislacion-habla/contracts/guia-y-changelog.md` §1, tal cual y en el sitio que dice (en «Qué va en el cuerpo y qué en una referencia», detrás del párrafo que nombra el contrato), con las cifras de lo medido, la ruta de su evidencia, `https://agentskills.io/` y el enlace a «Las dos fuentes» (`#las-dos-fuentes`); ninguna otra URL ni ningún otro cambio de `docs/SKILLS.md` (FR-014; SC-013; research D9). Verificación: quickstart §10 (solo `https://agentskills.io/`, que está en `web/src/components/Pie.astro:30`), el ancla `las-dos-fuentes` existe en `docs/SKILLS.md` (`grep -n '^#.*Las dos fuentes' docs/SKILLS.md`) y `make ci` en verde, con los controles de contenido de `TestEvalsDelRepositorio` que nombra §1 sin cambios

**Checkpoint**: §1 de la guía dice lo medido con sus cifras y sus enlaces.

## Fase 5: Cierre del artefacto — CHANGELOG (FR-016; SC-006)

- [ ] T006 [P] Añadir a `CHANGELOG.md`, como primera entrada de «### Cambiado» en *Unreleased*, la de `boe-legislacion` v0.3.1 de `specs/029-h31-1-boe-legislacion-habla/contracts/guia-y-changelog.md` §2, tal cual: la respuesta empieza por lo que se pregunta, lo que nombra se lee antes, se repasa antes de enviar y el cuerpo cabe en 11 000 caracteres; la entrada de v0.3 no se toca (FR-016; SC-006; research D11; Definition of Done §1.6). Verificación: quickstart §11 (`1`) y `make ci` en verde

---

## Controles de umbral: cada fila de plan.md y su tarea

| Requisito (fila de plan.md) | Control | Tarea que lo construye o lo ejerce | Su rojo |
|---|---|---|---|
| SC-001 | `make ci` | todas: cada una termina con `make ci` en verde | cualquier control de abajo |
| FR-005, SC-002 | gate del cuerpo de H31, sin cambios | T004 lo ejerce sobre el cuerpo nuevo | `TestCuerpoDeLaSkill`, caso `once-mil-y-un-caracteres` (research V1) |
| FR-007 | subtest `referencias` de `TestSkillsDelRepositorio`, sin cambios | T004 | una referencia sin enlace desde `SKILL.md` |
| FR-004 | regeneración de la tabla (`make skills-check`), sin cambios | T004 | drift entre la tabla y `--describe` |
| FR-008, SC-001 | los ocho controles de `TestEvalsDelRepositorio`, `normas-nombradas` y `sin-instrucciones-de-evals`, sin cambios | T004 | cualquier defecto de contenido en `SKILL.md` o una referencia |
| FR-001 … FR-010, SC-003 | los tres guiones de la suite | T001 los construye; T004 los pone en verde; el workflow los activa en `TestEntregaDelHito` | hoy, rojo en `stdout` (research V5); mutantes de la regla 5 y de «TRLRHL» |
| FR-011, SC-004 | subtest `lecturas` de `TestEvalsDelRepositorio`, mapa escrito a mano | T003 lo amplía y lo ve en rojo antes de tocar las evals | una lectura de más, de menos o con otro `motivo` |
| FR-012, FR-013, SC-004 | subtest `grabado`, sin cambios | T002 (manifiesto) y `grabar_datos`; T003 lo ejerce con las lecturas nuevas | una lectura sin su grabación (research V11) |
| FR-013, SC-004 | subtest `lecturas` de `TestCacheDeLasSesiones` | T003 amplía la tabla con los nueve bloques | un bloque que la caché no sirve sin red, sin fecha o con `9999-12-31` |
| FR-012 | `TestLecturaDeLaEvalSinGrabar`, sin cambios | T003 lo ejerce | una lectura sin grabar: sus dos faltas |
| SC-008 | las tres filas anteriores | T002, T003 | las de esas filas |
| FR-015, SC-005 | `TestDefinicionDelJob` (`fallosDelTope`), sin cambios | T003 lo comprueba en verde con el tope de hoy (research D8) | un `timeout-minutes` menor que el peor caso |
| SC-007, SC-009, SC-011 | umbrales que deciden del job de evals, sin cambios | ninguna tarea: los mide el cierre del workflow tras la revisión final (ADR 0018) | veredicto `fallo` y job en rojo |

## Definition of Done (docs/ROADMAP.md §1) que aplica

- §1.1 `make ci`: todas las tareas.
- §1.2 fixtures grabados: T002 (manifiesto) y el paso `grabar_datos`.
- §1.6 e2e y CHANGELOG: T001 (suite) y T006.
- §1.10 evals y guía de autoría: T003 (las evals cambian antes que la skill) y T004 (cuerpo ≤ 11 000, referencias enlazadas, tabla y generados sin drift).
- No aplican: ADR (ninguna decisión de arquitectura cambia; research D12), `docs/SOURCES.md` (la fuente ya tiene fila revisada en `main`, 2026-09-13; research V9), `schemas/` (ningún sobre cambia), código Go de producción (no hay), territorio y cobertura (§1.11 y §1.13: no cambian).

## Historias sin tarea

- **US5** (el cierre mide la skill con los umbrales de hoy; SC-007, SC-009 a SC-011): la medida la hace el workflow tras la revisión final; ningún umbral cambia.
- **US7** (el sondeo de la eval 15 antes de fusionar; SC-012): fuera del run, por decisión de quien fusiona (ADR 0018); ninguna tarea lo hace.

## Dependencias y orden

- T001 → T002 → (`grabar_datos`) → T003 → T004 → T005 → T006, en secuencia.
- T003 depende de las grabaciones que `grabar_datos` deja tras T002.
- T004 no depende de T002 ni T003, pero va detrás para que las evals cambien antes que la skill (Definition of Done §1.10) y la suite de T001 se ponga en verde con la skill final.
- T005 y T006 [P]: tocan ficheros distintos de los de T003 y T004 y no dependen de ellos; se ejecutan igualmente en orden.

## Ejecución en paralelo

Las tareas se ejecutan una a una. Lo único independiente entre sí es T005 (`docs/SKILLS.md`) y T006 (`CHANGELOG.md`).

## Estrategia

- **MVP**: T001 a T004 — la skill de la entrega con sus lecturas servidas (US1 a US4, todas P1).
- **Incremento**: T005 y T006 — la guía y el CHANGELOG (US6 y FR-016).
- El cierre (push, CI y evals del job) lo hace el workflow tras la revisión final.
