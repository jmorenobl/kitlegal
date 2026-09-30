# Informe del hito H7.4 · `boe-legislacion` sin el estado de la comprobación ni lo que no ha leído, y un job que juzga la respuesta a la pregunta

Generado por el workflow `hito` el 2026-09-30T22:01:37Z, sobre `19167ad` de `014-h7-4-boe-legislacion-sin`.
Lo escribe scripts/workflow/informe.sh sin modelo, desde los artefactos de `specs/014-h7-4-boe-legislacion-sin/`. Fusionar (squash-merge) es una decisión humana:
si algo de lo que sigue no es lo que se quería, se corrige la sección del hito en docs/ROADMAP.md y se relanza.

## 1. Estado

- **make ci local**: verde.
- **CI y evals remotos** sobre `19167ad` (medición 3): verde; es el producto de la cabeza (lo posterior solo toca `gates/`).
- **Evals por skill**: boe-legislacion aprobado, legal-core aprobado (tasas en la sección 3).
- **Umbrales del job**: boe-legislacion: 5 umbrales, **1 sin cumplir o solo publicados** (`expresiones_prohibidas:claude-haiku-4-5-20251001`); legal-core: sin umbrales (sección 3).
- **Revisión final**: juez A aprobado, juez B aprobado, 4 rondas en 3 ciclos (el primero juzga el hito; cada uno de los siguientes, lo que cambió después del último veredicto).
- **Cambios que ningún juez vio**: ninguno.
- **Tareas**: 13 hechas, 0 en cuarentena, 0 pendientes sin cuarentena.
- **Diff**: 24 commits; 106 files changed, 18409 insertions(+), 727 deletions(-).

## 2. Supuestos y pendientes

Decisiones que el run tomó sin preguntar, ordenadas por impacto: cada paso que escribe una etiqueta la suya (ADR 0028).

### Cambian el comportamiento visible, el alcance o una skill

**Alcance (lo que queda fuera o dentro del hito)**

- corrector_spec: «Fuera de alcance» no listaba cambiar `legal-core`, y un rojo de su job tras FR-003/FR-004 podía arreglarse tocando su `SKILL.md`, su versión o sus evals → queda fuera cambiar `legal-core` más allá de declarar en cada eval que `boe-legislacion` no se activa (FR-004); lo que evite esa activación va a la `description` y al cuerpo de `boe-legislacion` (FR-002, FR-003), porque el hito no pide cambiar la respuesta de `legal-core`.
- T007: `esperaMaximaDeLaDecision` es de 10 min, que es también el `-timeout` por defecto de `go test`, y el `run` del paso `decidir` de contracts/tanda-del-job.md §1 no lleva `-timeout`: con una anterior que no decide en 10 min, `go test` terminaría por su tope antes de que el bucle mida, y el paso fallaría en lugar de medir → T007 deja los 10 min del contrato; el `-timeout` que deja llegar a la espera agotada es del `run` de la definición y de T008, fuera de las rutas de esta tarea.
- T008: la tarea pide el `run` del paso `decidir` de contracts/tanda-del-job.md §1 carácter a carácter, y ese `run` no lleva `-timeout` (el supuesto de T007): con una anterior que no decide en 10 min, `go test` acabaría en pánico a los 10 min por omisión antes de que el bucle mida tras la espera, y `tanda` quedaría en rojo, contra FR-070 → el `run` lleva `-timeout 12m` (los 10 min de espera más sus consultas; con la preparación, cabe en los 15 del trabajo), contracts §1 lo recoge con su motivo, y … (entera en `specs/014-h7-4-boe-legislacion-sin/gates/supuestos.md` o `clarify-respuestas.json`)
- T013: la tarea pide tests solo si la cobertura global (≥ 70 %) o la del dominio (≥ 85 %) quedan bajo su umbral, y quedan en 97,4 % y 98,6 %; pero la estimación local del diff, que Codecov mide con `patch` (`target: auto`, bloqueante), da 92,5 % (335 de 362 líneas), por debajo de la base (97,4 % en el cierre de H7.3) → no se añaden tests: las 27 líneas sin cubrir son ramas de error de la regla genérica (20 de `tanda.go`, entre ellas las respuestas de `gh` que no se pueden leer, que tasks.md, «Pro… (entera en `specs/014-h7-4-boe-legislacion-sin/gates/supuestos.md` o `clarify-respuestas.json`)

**Skill (lo que pide, dice o comprueba una skill)**

- corrector_plan: FR-031 y US1 (escenario 5) piden que «la que se consultó antes» fuera de la línea `⚠ REDACCIÓN MODIFICADA:` cuente, y ninguna forma de la lista la casaba → `anuncio` gana `se consultó antes` (25 formas de la clase A, 87 en la lista), y `<cita>` es opcional en la forma fija de la línea para que la línea sin cita de los informes de H7.1 a H7.3 se siga quitando: el calibrado sigue en 36, 11 y 9 y solo sube `anuncio` en la 19; una respuesta que escriba la línea sin su cita no la marc… (entera en `specs/014-h7-4-boe-legislacion-sin/gates/supuestos.md` o `clarify-respuestas.json`)
- T010: contracts/skill-boe-legislacion.md C9 da el «Cuerpo de v0.1.4» de «Redacción modificada» hasta su lista, y §1 dice que todo lo que no nombra queda como en v0.1.3; no dice si el párrafo que cerraba la sección, «La etiqueta `REDACCIÓN MODIFICADA` no es la de ningún aviso de vigencia.», se va → se queda, como en v0.1.3: es la lectura que menos cambia lo que la skill dice, y separa la línea de los avisos de vigencia, cuya forma fija comparte (295 líneas, bajo las 300 de FR-026).
- reparar_cierre: la medición del cierre sobre `4cb52cb` da 15 de 54 respuestas del modelo que decide con una expresión de la lista (umbral 2), 14 por contar que `kitlegal graph check` no encontró nada, con las palabras de las formas fijas, y contracts/skill-boe-legislacion.md fijaba v0.1.4 en C1-C10 («todo lo que no nombra queda como en v0.1.3») sin decir qué hace la respuesta con un sobre sin entradas → `SKILL.md` gana C11: el caso sin entrada con su razón (paso 3, paso 5 y «Redacción modificada… (entera en `specs/014-h7-4-boe-legislacion-sin/gates/supuestos.md` o `clarify-respuestas.json`)
- reparar_cierre: la medición del cierre sobre `ad68a70` da 1 de 54 respuestas del modelo que decide con una expresión de `redaccion_no_leida` (umbral 0): la 18-01 presenta la última redacción del art. 42 de la Ley 30/1992, derogada, como «vigente hasta la derogación», porque `SKILL.md` enseñaba a decir de una redacción la norma que la dio y «desde cuándo está vigente», y la regla 3 a no presentar como vigente una norma derogada, sin decir nunca que ningún sobre trae hasta cuándo rige una redacció… (entera en `specs/014-h7-4-boe-legislacion-sin/gates/supuestos.md` o `clarify-respuestas.json`)

### Del propio run

- 2026-09-30T20:36:28Z · La revisión final abre el ciclo 2 para juzgar lo que cambió fuera de gates/ después del veredicto sobre 885a688: 5fcdb6f «fix(H7.4): cierre en la plataforma».
- 2026-09-30T21:21:16Z · La revisión final abre el ciclo 3 para juzgar lo que cambió fuera de gates/ después del veredicto sobre 5fcdb6f: 53dfec5 «fix(H7.4): cierre en la plataforma».

- Observaciones de los jueces, por debajo del umbral y sin corregir: 10 en `spec-r2.json`, 15 en `plan-r2.json`, 7 en `tasks-r1.json`, 7 en `revision-a-r4.json`, 7 en `revision-b-r4.json`.

### Internos (14)

Decisiones que no cambian nada observable (técnica, estructura, tests), por autor: T002 (2), T003 (1), T004 (1), T005 (1), T006 (1), T007 (1), T008 (2), T010 (1), T011 (2), T012 (1), corrector_revision (1). Enteras en `specs/014-h7-4-boe-legislacion-sin/gates/supuestos.md`.

## 3. Evals sobre la cabeza

**boe-legislacion**: aprobado sobre `19167ad`; decide `claude-sonnet-5-5` con 2 de 3; 20 evals, 1 nuevas, 8 informativas.

| Eval | claude-sonnet-5-5 | claude-haiku-4-5-20251001 | Marca |
|---|---|---|---|
| `20-lcsp-dos-bloques-redaccion-cambiada.yaml` | 3/3 | — | **nueva** · informativa |
| `01-lpac-articulo-21.yaml` | 3/3 | 3/3 |  |
| `02-lcsp-contrato-menor.yaml` | 3/3 | 3/3 |  |
| `03-lrbrl-atribuciones-del-pleno.yaml` | 3/3 | 3/3 |  |
| `04-lgt-prescripcion.yaml` | 3/3 | 3/3 |  |
| `05-trlrhl-impuestos-municipales.yaml` | 3/3 | 3/3 |  |
| `06-irpf-rendimientos-del-trabajo.yaml` | 3/3 | 3/3 |  |
| `07-lrjsp-principio-de-legalidad.yaml` | 3/3 | 3/3 |  |
| `08-ltaibg-plazo-de-resolucion.yaml` | 3/3 | 3/3 |  |
| `09-constitucion-articulo-140.yaml` | 3/3 | 3/3 |  |
| `10-et-vacaciones.yaml` | 2/3 | 2/3 |  |
| `11-no-activa-programacion.yaml` | 3/3 | 3/3 |  |
| `12-no-activa-acuerdo-entre-amigos.yaml` | 3/3 | 3/3 |  |
| `13-lrbrl-atribuciones-por-materia.yaml` | 3/3 | — | informativa |
| `14-trlrhl-impuestos-por-materia.yaml` | 3/3 | — | informativa |
| `15-irpf-rendimientos-por-materia.yaml` | 3/3 | — | informativa |
| `16-lrjsp-legalidad-por-materia.yaml` | 3/3 | — | informativa |
| `17-ltaibg-plazo-por-materia.yaml` | 3/3 | — | informativa |
| `18-lrjpac-norma-derogada.yaml` | 3/3 | — | informativa |
| `19-lcsp-contrato-menor-redaccion-cambiada.yaml` | 3/3 | — | informativa |

Respuestas con alguna expresión prohibida, en las evals que activan la skill (`expresiones_prohibidas_por_modelo`):

| Modelo | Con alguna | Respuestas | Porcentaje |
|---|---|---|---|
| `claude-sonnet-5-5` | 0 | 54 | 0,0 % |
| `claude-haiku-4-5-20251001` | 0 | 30 | 0,0 % |

Umbrales que publica el job (ADR 0029):

| Umbral | Medida | Condición | Cumple | Hace fallar el job |
|---|---|---|---|---|
| `expresiones_prohibidas:claude-sonnet-5-5` | 0 de 54 (0,0 %) | ≤ 5,0 % | sí | sí |
| `sin_activar:claude-sonnet-5-5` | 0 de 54 (0,0 %) | ≤ 0,0 % | sí | sí |
| `redaccion_no_leida:claude-sonnet-5-5` | 0 de 54 (0,0 %) | ≤ 0,0 % | sí | sí |
| `expresiones_prohibidas:claude-haiku-4-5-20251001` | 0 de 30 (0,0 %) | ≤ 5,0 % | sí | no: solo se publica, no es un control |
| `duracion_de_las_sesiones` | 452 | ≤ 900 | sí | sí |

**legal-core**: aprobado sobre `19167ad`; decide `claude-sonnet-5-5` con 2 de 3; 3 evals, 0 nuevas, 0 informativas.

| Eval | claude-sonnet-5-5 | claude-haiku-4-5-20251001 | Marca |
|---|---|---|---|
| `01-territorio-municipio-cubierto.yaml` | 3/3 | 3/3 | cambiada |
| `02-territorio-municipio-no-cubierto.yaml` | 3/3 | 3/3 | cambiada |
| `03-no-activa-receta-de-cocina.yaml` | 3/3 | 3/3 | cambiada |

Umbrales: el job no publica ninguno para esta skill.

Cada celda: sesiones que pasan de las abiertas; ✗, una serie que decide y no llega al umbral. Una eval informativa publica su tasa sin decidir el veredicto (ADR 0016). La regla por serie no hace cumplir un umbral agregado sobre todas las respuestas: eso solo lo hace un umbral del job que lo hace fallar (ADR 0029).

## 4. Revisión que la constitución reserva a la persona

- Fixture o esquema EXISTENTE modificado: `schemas/eval.yaml.json`.
- Fixture o esquema EXISTENTE modificado: `schemas/expresiones-prohibidas.yaml.json`.
- Fixtures nuevos (5), por directorio:
  - `internal/evals/testdata/sesiones/leer-sesion/respuesta-antes-de-una-tarea-en-segundo-plano/`: `codigo-de-la-sesion`, `sesion.err`, `sesion.jsonl`
  - `testdata/evals/grafo-previo/lcsp-a1-30-y-da-3-redaccion-original/`: `GET_https_www.boe.es_datosabiertos_api_legislacion-consolidada_id_BOE-A-2017-12902_texto_bloque_a1-30.json`, `GET_https_www.boe.es_datosabiertos_api_legislacion-consolidada_id_BOE-A-2017-12902_texto_bloque_da-3.json`

## 5. Tareas en cuarentena

Ninguna.

## 6. Trazabilidad

**Estado**: «tareas hechas» solo dice que las tareas que citan el requisito están marcadas; nada del run lo ha medido. «comprobado por su control» exige su fila en «Controles de umbral» de `plan.md` y que el control esté: un test o una comprobación de `make ci` que existe en la cabeza, con `make ci` en verde, o un umbral que el job de evals publica, cumple y hace fallar el job (ADR 0029). «UMBRAL NO CUMPLIDO» y «CONTROL SIN VERIFICAR» son lo que hay que mirar.

| Requisito | Tareas | Aceptación | Estado | Control de umbral |
|---|---|---|---|---|
| FR-001 | T010(hecha) | — | tareas hechas | — |
| FR-002 | T010(hecha) | — | tareas hechas | — |
| FR-003 | T002(hecha) T004(hecha) T010(hecha) | — | tareas hechas | — |
| FR-004 | T002(hecha) T004(hecha) | — | tareas hechas | — |
| FR-010 | T010(hecha) | — | tareas hechas | — |
| FR-011 | T010(hecha) | — | tareas hechas | — |
| FR-012 | T003(hecha) T010(hecha) | — | tareas hechas | — |
| FR-020 | T010(hecha) | — | tareas hechas | — |
| FR-021 | T010(hecha) | — | comprobado por su control | `ci:internal/evals/conjunto_test.go:TestEvalsDelRepositorio`, en make ci (verde) |
| FR-022 | T010(hecha) | — | tareas hechas | — |
| FR-023 | T004(hecha) T010(hecha) | — | tareas hechas | — |
| FR-024 | T010(hecha) | — | comprobado por su control | `ci:internal/evals/conjunto_test.go:TestEvalsDelRepositorio`, en make ci (verde); `ci:internal/evals/conjunto_test.go:TestOrdenesParaPowerShell`, en make ci (verde) |
| FR-025 | T010(hecha) | — | tareas hechas | — |
| FR-026 | T010(hecha) | — | comprobado por su control | `ci:Makefile:skills-check`, en make ci (verde) |
| FR-027 | T012(hecha) | — | tareas hechas | — |
| FR-030 | T001(hecha) T003(hecha) T010(hecha) | — | tareas hechas | — |
| FR-031 | T001(hecha) T003(hecha) | — | tareas hechas | — |
| FR-032 | T010(hecha) | — | comprobado por su control | `ci:internal/evals/conjunto_test.go:TestEvalsDelRepositorio`, en make ci (verde) |
| FR-033 | T010(hecha) | — | comprobado por su control | `ci:internal/evals/conjunto_test.go:TestEvalsDelRepositorio`, en make ci (verde) |
| FR-034 | T010(hecha) | — | comprobado por su control | `ci:internal/evals/juzgar_test.go:TestJuzgarLasClasesDeLaRespuesta`, en make ci (verde) |
| FR-035 | T001(hecha) | — | tareas hechas | — |
| FR-036 | T003(hecha) T006(hecha) | — | tareas hechas | — |
| FR-037 | T006(hecha) T010(hecha) | — | tareas hechas | — |
| FR-040 | T006(hecha) T013(hecha) | — | comprobado por su control | `evals:boe-legislacion:expresiones_prohibidas:claude-sonnet-5-5`: 0 de 54 (0,0 %), ≤ 5,0 % |
| FR-041 | T006(hecha) T013(hecha) | — | comprobado por su control | `evals:boe-legislacion:sin_activar:claude-sonnet-5-5`: 0 de 54 (0,0 %), ≤ 0,0 %; `ci:internal/evals/umbrales_test.go:TestUmbralesDelInforme`, en make ci (verde) |
| FR-042 | T006(hecha) T013(hecha) | — | comprobado por su control | `evals:boe-legislacion:redaccion_no_leida:claude-sonnet-5-5`: 0 de 54 (0,0 %), ≤ 0,0 %; `ci:internal/evals/umbrales_test.go:TestUmbralesDelInforme`, en make ci (verde) |
| FR-043 | T006(hecha) T008(hecha) T013(hecha) | — | comprobado por su control | `evals:boe-legislacion:duracion_de_las_sesiones`: 452, ≤ 900 |
| FR-044 | T006(hecha) | — | tareas hechas | — |
| FR-045 | T006(hecha) | — | comprobado por su control | `ci:internal/evals/umbrales_test.go:TestUmbralesDelInforme`, en make ci (verde) |
| FR-046 | T006(hecha) | — | tareas hechas | — |
| FR-047 | T006(hecha) T013(hecha) | — | tareas hechas | — |
| FR-048 | T006(hecha) | — | tareas hechas | — |
| FR-050 | T009(hecha) | — | tareas hechas | — |
| FR-051 | T009(hecha) | — | tareas hechas | — |
| FR-052 | T004(hecha) T009(hecha) | — | tareas hechas | — |
| FR-053 | T002(hecha) T004(hecha) | — | tareas hechas | — |
| FR-054 | T008(hecha) T009(hecha) | — | comprobado por su control | `ci:internal/evals/definicion_test.go:TestDefinicionDelJob`, en make ci (verde) |
| FR-055 | T009(hecha) | — | tareas hechas | — |
| FR-060 | T005(hecha) | — | tareas hechas | — |
| FR-061 | T006(hecha) | — | tareas hechas | — |
| FR-062 | T004(hecha) T005(hecha) T006(hecha) | — | tareas hechas | — |
| FR-070 | T007(hecha) T008(hecha) | — | comprobado por su control | `ci:internal/evals/definicion_test.go:TestDefinicionDelJob`, en make ci (verde) |
| FR-071 | T007(hecha) T008(hecha) | — | comprobado por su control | `ci:internal/evals/definicion_test.go:TestDefinicionDelJob`, en make ci (verde) |
| FR-072 | T008(hecha) T013(hecha) | — | tareas hechas | — |
| FR-080 | T011(hecha) | — | comprobado por su control | `ci:internal/evals/sondeo_test.go:TestGuionDelSondeo`, en make ci (verde); `ci:internal/evals/sondeo_test.go:TestSondear`, en make ci (verde) |
| FR-081 | T011(hecha) | — | tareas hechas | — |
| FR-082 | T006(hecha) T011(hecha) | — | tareas hechas | — |
| FR-090 | T010(hecha) T013(hecha) | — | tareas hechas | — |
| FR-091 | T013(hecha) | — | tareas hechas | — |
| FR-092 | T013(hecha) | — | tareas hechas | — |
| FR-093 | T003(hecha) T010(hecha) | — | comprobado por su control | `ci:internal/evals/conjunto_test.go:TestEvalsDelRepositorio`, en make ci (verde); `ci:internal/evals/conjunto_test.go:TestEvalsDelRepositorio`, en make ci (verde); `ci:internal/evals/conjunto_test.go:TestEvalsDelRepositorio`, en make ci (verde) |
| FR-094 | T004(hecha) T010(hecha) | — | comprobado por su control | `ci:internal/evals/juzgar_test.go:TestJuzgarLasClasesDeLaRespuesta`, en make ci (verde) |
| FR-095 | T010(hecha) | — | comprobado por su control | `ci:internal/evals/conjunto_test.go:TestEvalsDelRepositorio`, en make ci (verde); `ci:internal/evals/conjunto_test.go:TestOrdenesParaPowerShell`, en make ci (verde) |
| FR-096 | T009(hecha) | — | tareas hechas | — |
| FR-097 | T006(hecha) | — | comprobado por su control | `ci:internal/evals/umbrales_test.go:TestUmbralesDelInforme`, en make ci (verde) |
| FR-098 | T005(hecha) | — | tareas hechas | — |
| FR-099 | T011(hecha) | — | comprobado por su control | `ci:internal/evals/sondeo_test.go:TestGuionDelSondeo`, en make ci (verde); `ci:internal/evals/sondeo_test.go:TestSondear`, en make ci (verde) |
| FR-100 | T007(hecha) T008(hecha) | — | comprobado por su control | `ci:internal/evals/definicion_test.go:TestDefinicionDelJob`, en make ci (verde) |
| FR-101 | T012(hecha) | — | tareas hechas | — |
| FR-102 | — | — | SIN TAREA | — |
| SC-001 | T006(hecha) | — | comprobado por su control | `evals:boe-legislacion:expresiones_prohibidas:claude-sonnet-5-5`: 0 de 54 (0,0 %), ≤ 5,0 %; `evals:boe-legislacion:sin_activar:claude-sonnet-5-5`: 0 de 54 (0,0 %), ≤ 0,0 %; `evals:boe-legislacion:redaccion_no_leida:claude-sonnet-5-5`: 0 de 54 (0,0 %), ≤ 0,0 %; `evals:boe-legislacion:duracion_de_las_sesiones`: 452, ≤ 900; `ci:internal/evals/informe_test.go:TestInformeConSesionesSinMedir`, en make ci (verde) |
| SC-002 | T001(hecha) T003(hecha) T010(hecha) | — | comprobado por su control | `ci:internal/evals/conjunto_test.go:TestEvalsDelRepositorio`, en make ci (verde) |
| SC-003 | T003(hecha) T009(hecha) T010(hecha) | — | comprobado por su control | `ci:internal/evals/conjunto_test.go:TestEvalsDelRepositorio`, en make ci (verde) |
| SC-004 | T010(hecha) | — | comprobado por su control | `ci:internal/evals/conjunto_test.go:TestEvalsDelRepositorio`, en make ci (verde) |
| SC-005 | T010(hecha) | — | comprobado por su control | `ci:internal/evals/conjunto_test.go:TestEvalsDelRepositorio`, en make ci (verde); `ci:internal/evals/conjunto_test.go:TestOrdenesParaPowerShell`, en make ci (verde) |
| SC-006 | T006(hecha) | — | comprobado por su control | `ci:internal/evals/umbrales_test.go:TestUmbralesDelInforme`, en make ci (verde) |
| SC-007 | T004(hecha) T010(hecha) | — | comprobado por su control | `ci:internal/evals/juzgar_test.go:TestJuzgarLasClasesDeLaRespuesta`, en make ci (verde) |
| SC-008 | T005(hecha) | — | tareas hechas | — |
| SC-009 | T011(hecha) | — | comprobado por su control | `ci:internal/evals/sondeo_test.go:TestGuionDelSondeo`, en make ci (verde); `ci:internal/evals/sondeo_test.go:TestSondear`, en make ci (verde) |
| SC-010 | T007(hecha) T008(hecha) | — | comprobado por su control | `ci:internal/evals/definicion_test.go:TestDefinicionDelJob`, en make ci (verde) |
| SC-011 | T010(hecha) T012(hecha) T013(hecha) | — | comprobado por su control | `ci:Makefile:skills-check`, en make ci (verde) |
| SC-012 | T009(hecha) | — | tareas hechas | — |

## 7. Cambios posteriores a la revisión final

Cada commit posterior a lo que juzgó la primera ronda de la revisión final (`885a688`), con la ronda que lo vio y su veredicto, y lo que toca fuera de `gates/` (ADR 0030):

- `4cb52cb` docs(H7.4): veredictos de la revisión final: solo registros de `gates/`.
- `5fcdb6f` fix(H7.4): cierre en la plataforma: lo vio la ronda 2 (juez A: aprobado; juez B: aprobado). Toca `CHANGELOG.md`, `skills/boe-legislacion/SKILL.md`, `cierre.md`, `contracts/skill-boe-legislacion.md`, `plan.md`, `research.md`.
- `44cb806` docs(H7.4): veredictos de la revisión final: solo registros de `gates/`.
- `ad68a70` docs(H7.4): registros del run: solo registros de `gates/`.
- `53dfec5` fix(H7.4): cierre en la plataforma: lo vio la ronda 3 (juez A: rechazado; juez B: rechazado). Toca `CHANGELOG.md`, `skills/boe-legislacion/SKILL.md`, `cierre.md`, `contracts/skill-boe-legislacion.md`, `plan.md`, `research.md`.
- `1a94ddd` fix(H7.4): motivos de la revisión final: lo vio la ronda 4 (juez A: aprobado; juez B: aprobado). Toca `research.md`.
- `9ea37a7` docs(H7.4): veredictos de la revisión final: solo registros de `gates/`.
- `19167ad` docs(H7.4): registros del run: solo registros de `gates/`.

### Cambios que ningún juez vio

Ninguno.

## 8. Cómo comprobarlo y consumo

- Escenarios manuales: `specs/014-h7-4-boe-legislacion-sin/quickstart.md`. Suite de aceptación congelada: `specs/014-h7-4-boe-legislacion-sin/aceptacion/` (activada en `internal/app/testdata/script/`).
- Run `0f3ef83d`: 9 h 9 min de reloj. Coste por paso y por rol: `scripts/coste-run.sh 0f3ef83d`.
