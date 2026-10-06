# Informe del hito H24 · Las evals juzgan el significado con un modelo: «afirma lo que no ha leído» decide, `boe-legislacion` deja de glosar lo que no ha leído, y la lista de expresiones deja de decidir (adelantado; ADR 0037)

Generado por el workflow `hito` el 2026-10-06T03:26:11Z, sobre `3107ae2` de `017-h24-las-evals-juzgan`.
Lo escribe scripts/workflow/informe.sh sin modelo, desde los artefactos de `specs/017-h24-las-evals-juzgan/`. Fusionar (squash-merge) es una decisión humana:
si algo de lo que sigue no es lo que se quería, se corrige la sección del hito en docs/ROADMAP.md y se relanza.

## 1. Estado

- **make ci local**: verde.
- **CI y evals remotos** sobre `3107ae2` (medición 3): verde; es el producto de la cabeza (lo posterior solo toca `gates/`).
- **Evals por skill**: boe-legislacion aprobado, legal-core aprobado (tasas en la sección 3).
- **Umbrales del job**: boe-legislacion: 12 umbrales, **2 sin cumplir o solo publicados** (`cuenta_su_proceso:claude-sonnet-5-5:orden`, `cuenta_su_proceso:claude-sonnet-5-5:herramienta`); legal-core: sin umbrales (sección 3).
- **Revisión final**: juez A aprobado, juez B aprobado, 6 rondas en 3 ciclos (el primero juzga el hito; cada uno de los siguientes, lo que cambió después del último veredicto).
- **Cambios que ningún juez vio**: ninguno.
- **Tareas**: 18 hechas, 0 en cuarentena, 0 pendientes sin cuarentena.
- **Diff**: 32 commits; 122 files changed, 41244 insertions(+), 2868 deletions(-).

## 2. Supuestos y pendientes

Decisiones que el run tomó sin preguntar, ordenadas por impacto: cada paso que escribe una etiqueta la suya (ADR 0028).

### Cambian el comportamiento visible, el alcance o una skill

**Comportamiento (salida, códigos, ficheros, argumentos)**

- T010: «con el caso y sus frases» no dice qué frases lleva la línea de un defecto que no queda marcado → las de sus votos que cuentan y dicen sí con su frase en la respuesta, que es lo que la regla llama un sí («sí, sí y no: sin marcar, con sus dos frases»): tres en un correcto marcado, y dos, una o ninguna en un defecto sin marcar. Sin ninguna, la línea termina en «sin marcar», sin los dos puntos. La frase de un voto nulo no es una de ellas, tampoco la de su repetición si vuelve a serlo.
- T010: el contrato no fija la línea de un caso sin juzgar, ni si con alguno se nombran también los casos que no dan lo que dice su etiqueta → `<informe> <sesión> [sin <norma> <bloque>]: sin juzgar: <motivo>`, con el motivo de la regla (`voto 2: tope de 35 s agotado`), y solo esos casos (FR-053: «con ese motivo y los casos sin juzgar»): sin medida, los recuentos de los demás no son de nada. El error no lleva cabecera en ninguno de los dos casos: es una línea por caso, y los recuentos ya van en la … (entera en `specs/017-h24-las-evals-juzgan/gates/supuestos.md` o `clarify-respuestas.json`)
- T010: la tarea no dice qué hace `medirAlJuez` con una skill sin juez, sin votante o con menos de un caso a la vez → un error que lo dice, sin ningún voto ni ninguna medida, como hace `EscribirInforme` con su votante y su concurrencia. Tampoco da medida si alguno de sus textos (la skill, el modelo, la versión o el commit) no es UTF-8: se escribe tal cual o no se da.
- T011: contracts/informe-del-job.md §5 fija el veredicto, los motivos, `umbrales`, `juez` y cinco listas del informe del instrumento sin medir, y no dice nada del resto de `informe.json` ni de `informe.md` → la cabecera es la de cualquier informe, con lo recibido tal cual y `sin_python` leído de su fichero; no se leen las evals ni ninguna sesión, así que `ficheros_mal_formados`, `modelos_de_sesion` y `versiones_de_claude_code` van vacías y la duración y los reintentos son 0; `juez` lleva el model… (entera en `specs/017-h24-las-evals-juzgan/gates/supuestos.md` o `clarify-respuestas.json`)
- T012: «imprime `medida.json` entre sus dos marcas si el test lo escribió, y nada entre marcas si no» admite dos lecturas: las dos marcas sin nada dentro, o ninguna marca → ninguna marca. Quien saca la medida del registro toma lo que hay entre ellas, y dos marcas vacías serían una medida vacía donde FR-053 dice que no se imprime ninguna; es además lo que hace `scripts/evals.sh` con un informe que no se escribió. `TestGuionDeLaMedida` lo fija: con el `go` que falla sin escribirla, la salida estánd… (entera en `specs/017-h24-las-evals-juzgan/gates/supuestos.md` o `clarify-respuestas.json`)
- T012: «con la forma del guion del sondeo» no dice si las dos salidas de `go test` se guardan en un registro que solo se imprime si falla, como en el sondeo, o son las del guion → son las del guion, sin guardar, como en `scripts/evals.sh`: el error de `TestMedidaDelJuez` es el que nombra cada caso (contracts/medida-del-juez.md §7), y tiene que estar en el registro del trabajo también cuando la medida se imprime. La medida va detrás, entre sus marcas. Del sondeo toma lo demás: su temporal propio, … (entera en `specs/017-h24-las-evals-juzgan/gates/supuestos.md` o `clarify-respuestas.json`)
- T012: «dice cuál falta y sale con 1» no fija el texto ni qué se mira de cada variable → `evals-medir-juez: falta <variable>`, la línea de `scripts/evals.sh` con el nombre de este guion, una sola: la primera que falta en el orden de la tarea; falta la que no está o está vacía, y `CLAUDE_DEL_JUEZ` también si su ruta no es la de un fichero ejecutable, con la misma línea (como en T011). Sin la skill —sin argumentos, con ella vacía, que es lo que da `make` sin `SKILL`, o con un argumento más—, `evals… (entera en `specs/017-h24-las-evals-juzgan/gates/supuestos.md` o `clarify-respuestas.json`)
- T012: la medida lleva una `fecha` y la tarea no dice de qué instante de una ejecución que puede durar horas → el instante en el que el punto de entrada llama a `medirAlJuez`, antes del primer voto, en el huso del equipo (T010).
- T014: contracts/informe-del-job.md §8 da la línea de la medida que no corresponde solo con el ejemplo de la versión, y la tarea pide «que no corresponde y por qué» → lo que va entre «no corresponde a lo que hay: » y «. Estos recuentos no son los de un juez medido.» son las líneas de `comprobarLaMedida` con el modelo del juez de la definición del job y la versión del equipo, en su orden y separadas por «; »; la de la versión, con las palabras del ejemplo («la versión de Claude Code de sus votos e… (entera en `specs/017-h24-las-evals-juzgan/gates/supuestos.md` o `clarify-respuestas.json`)
- T014: «si ningún transcript declara la versión, la línea lo dice y el sondeo sigue» no fija el texto → `La medida versionada del juez no se puede comparar con lo que hay: ningún transcript de las sesiones declara la versión de Claude Code de este equipo. No se sabe si estos recuentos son los de un juez medido.`, sin llamar a `comprobarLaMedida`, que no tiene versión con la que comparar. No dice «no son los de un juez medido»: sin la versión no se sabe.
- T014: «`Respuestas sin juzgar: ninguna.` o una línea por cada una» no fija la forma de esas líneas → `Respuestas sin juzgar:` y, debajo, `- <sesión>: <motivo>` por cada una, en orden de sesión, como los otros dos apartados de sesiones del sondeo y sin línea en blanco delante: una línea más por respuesta, que es lo que cuenta el contrato.
- T014: T003 dejó para los puntos de entrada qué hace una ejecución interrumpida (`SIGINT`, `SIGTERM`) mientras vota el juez → nada propio, como en T012: los votos abiertos se cortan y los siguientes no llegan a darse, sus respuestas salen en «Respuestas sin juzgar» con su motivo y el sondeo da su salida. La interrupción durante las sesiones sigue siendo el error del repartidor, sin salida.

**Alcance (lo que queda fuera o dentro del hito)**

- plan: dos de los 259 casos etiquetados son de una eval que H7.2 retiró (`19-lpac-articulo-21-redaccion-cambiada`, informe de H7.1), y ni su pregunta ni su grafo previo están en `main`; el spec supone que los 259 se reconstruyen con lo que hay → se restauran sus dos ficheros, byte a byte con los de `c4819d1^`, bajo `testdata/evals/retiradas/`, en una tarea `[datos]`, y la derivada recupera su control (research D15). Rechazado: leerlos de la historia de git al ejecutar, que no existe en el job ni … (entera en `specs/017-h24-las-evals-juzgan/gates/supuestos.md` o `clarify-respuestas.json`)
- T010: FR-051 dice «por cada clase que decide, vota sus casos» y `casos.yaml` lleva una sola `clase` → la medida es la de la clase de los casos, que tiene que ser una clase del juez que decide (research D19); con unos casos de una clase que no es del juez, o de una que solo se publica, la ejecución termina con un error sin pedir ningún voto. No se mide una clase que no decide.
- T017: la tarea pide dejar constancia de que el árbol no lleva ningún `//nolint` nuevo, y lleva uno —el `//nolint:misspell` de `internal/evals/conjunto_test.go:2105`, de T015, con su supuesto más arriba—; no dice qué hacer si lo encuentra → `cierre.md` deja constancia de lo que hay (§5: un `//nolint` nuevo, ningún `t.Skip` ni TODO) y no lo retira: retirarlo es añadir `informativo` a `ignore-rules` de `.golangci.yml` y quitar el comentario del test, dos ficheros que no están entre las rutas de la … (entera en `specs/017-h24-las-evals-juzgan/gates/supuestos.md` o `clarify-respuestas.json`)
- clarify Q1 (criterio d, conservadora): Solo se publica. Las tres respuestas del modelo que decide en la eval sin binario ni servidor (la 21) se juzgan como las demás: el mismo voto, el mismo mensaje —que dice que ninguna herramienta devolvió ningún texto, como en la validación—, la misma comprobación de la frase y la misma regla de los tres votos por orden (FR-001 a FR-011). Sus votos y sus frases van al informe como los de cualquier respuesta con algún voto afirmativo, con si quedó marcada (FR-… (entera en `specs/017-h24-las-evals-juzgan/gates/supuestos.md` o `clarify-respuestas.json`)

**Skill (lo que pide, dice o comprueba una skill)**

- plan: la comprobación de la prosa busca «el sobre», `fecha_vigencia` y `norma_modificadora` en toda la prosa de `SKILL.md`, con su código en línea, así que v0.1.7 quita esos nombres también del paso 4 y de la línea de la redacción modificada, donde las dos fechas pasan a decirse por su orden (research D22, D23; contracts/skill-boe-legislacion.md C5 y C7) → riesgo sin medir sobre las evals 19 y 20, que mide el cierre. Rechazado: buscarlo solo donde el párrafo nombra «la respuesta», que no separa … (entera en `specs/017-h24-las-evals-juzgan/gates/supuestos.md` o `clarify-respuestas.json`)
- plan: la lista de expresiones sale también de `comprobarConsultaRepetida`, la comprobación a mano del quickstart de H7.2, que FR-070 no enumera → se quita, porque es otra respuesta de un modelo juzgada con la lista (research D14). Rechazado: dejarla como está.
- reparar_cierre: la medición del cierre sobre `a5bda45` da 1 de 54 respuestas del modelo que decide marcadas en `afirma_lo_no_leido` en el modo orden (umbral 0; en el modo herramienta, 0 de 54; los demás umbrales que deciden se cumplen, `cuenta_su_proceso`, que solo se publica, da 0 y 3, ninguna respuesta queda sin juzgar y `legal-core` aprueba): la 08-01 parafrasea el art. 20.5 de la Ley 19/2013, que leyó, y cuelga de su remisión al art. 24 el inciso «que es previa a ese recurso», que es la regl… (entera en `specs/017-h24-las-evals-juzgan/gates/supuestos.md` o `clarify-respuestas.json`)
- reparar_cierre: la medición del cierre sobre `b0a7a9c` (ronda 2) da 3 de 54 respuestas del modelo que decide marcadas en `afirma_lo_no_leido` en el modo herramienta (umbral 0; en el modo orden, 0 de 54, y el 1 de la ronda 1 ya no se da; los demás umbrales que deciden se cumplen, `cuenta_su_proceso`, que solo se publica, da 1 y 1, y `legal-core` aprueba): las tres juntan el número de un precepto que no leyeron con su materia —«los artículos que regulan cada impuesto (60, 78, 92, 100 y 104)» (14-0… (entera en `specs/017-h24-las-evals-juzgan/gates/supuestos.md` o `clarify-respuestas.json`)

### Del propio run

- 2026-10-06T00:29:59Z · La ronda 3 de la revisión final (ciclo 1) aprobó con los dos jueces lo que quedaba en revision-pendiente.md; se conserva como specs/017-h24-las-evals-juzgan/gates/revision-pendiente-resuelto-r3.md.
- 2026-10-06T01:15:43Z · La revisión final abre el ciclo 2 para juzgar lo que cambió fuera de gates/ después del veredicto sobre ae250eb: b50e05b «fix(H24): cierre en la plataforma».
- 2026-10-06T02:40:07Z · La revisión final abre el ciclo 3 para juzgar lo que cambió fuera de gates/ después del veredicto sobre 51db15e: d53bc89 «fix(H24): cierre en la plataforma».

- Observaciones de los jueces, por debajo del umbral y sin corregir: 15 en `spec-r2.json`, 14 en `plan-r1.json`, 14 en `tasks-r2.json`, 15 en `revision-a-r6.json`, 7 en `revision-b-r6.json`.

### Internos (74)

Decisiones que no cambian nada observable (técnica, estructura, tests), por autor: T001 (4), T002 (7), T003 (4), T004 (3), T005 (5), T006 (8), T007 (3), T008 (1), T009 (7), T010 (2), T011 (3), T012 (3), T013 (5), T014 (1), T015 (1), T016 (4), T017 (1), T018 (1), clarify Q2 (1), corrector_revision (5), corrector_tasks (1), plan (3), reparar_cierre (1). Enteras en `specs/017-h24-las-evals-juzgan/gates/supuestos.md`.

## 3. Evals sobre la cabeza

**boe-legislacion**: aprobado sobre `3107ae2`; decide `claude-sonnet-5-5` con 2 de 3; 21 evals, 0 nuevas, 8 informativas.

| Eval | claude-sonnet-5-5 | claude-haiku-4-5-20251001 | claude-sonnet-5-5 (herramienta) | claude-haiku-4-5-20251001 (herramienta) | Marca |
|---|---|---|---|---|---|
| `01-lpac-articulo-21.yaml` | 3/3 | 3/3 | 3/3 | 2/3 |  |
| `02-lcsp-contrato-menor.yaml` | 3/3 | 3/3 | 3/3 | 3/3 |  |
| `03-lrbrl-atribuciones-del-pleno.yaml` | 3/3 | 3/3 | 3/3 | 2/3 |  |
| `04-lgt-prescripcion.yaml` | 3/3 | 3/3 | 3/3 | 2/3 |  |
| `05-trlrhl-impuestos-municipales.yaml` | 3/3 | 3/3 | 3/3 | 1/3 |  |
| `06-irpf-rendimientos-del-trabajo.yaml` | 3/3 | 3/3 | 3/3 | 2/3 |  |
| `07-lrjsp-principio-de-legalidad.yaml` | 3/3 | 3/3 | 3/3 | 0/3 |  |
| `08-ltaibg-plazo-de-resolucion.yaml` | 3/3 | 3/3 | 3/3 | 3/3 |  |
| `09-constitucion-articulo-140.yaml` | 3/3 | 3/3 | 3/3 | 3/3 |  |
| `10-et-vacaciones.yaml` | 3/3 | 3/3 | 3/3 | 1/3 |  |
| `11-no-activa-programacion.yaml` | 3/3 | 3/3 | 3/3 | 3/3 |  |
| `12-no-activa-acuerdo-entre-amigos.yaml` | 3/3 | 3/3 | 3/3 | 3/3 |  |
| `13-lrbrl-atribuciones-por-materia.yaml` | 3/3 | — | 3/3 | — | informativa |
| `14-trlrhl-impuestos-por-materia.yaml` | 3/3 | — | 3/3 | — | informativa |
| `15-irpf-rendimientos-por-materia.yaml` | 3/3 | — | 3/3 | — | informativa |
| `16-lrjsp-legalidad-por-materia.yaml` | 3/3 | — | 3/3 | — | informativa |
| `17-ltaibg-plazo-por-materia.yaml` | 3/3 | — | 3/3 | — | informativa |
| `18-lrjpac-norma-derogada.yaml` | 3/3 | — | 3/3 | — | informativa |
| `19-lcsp-contrato-menor-redaccion-cambiada.yaml` | 3/3 | — | 3/3 | — | informativa |
| `20-lcsp-dos-bloques-redaccion-cambiada.yaml` | 3/3 | — | 3/3 | — | informativa |
| `21-sin-binario-ni-servidor.yaml` | 3/3 | 2/3 | — | — |  |

Umbrales que publica el job (ADR 0029):

| Umbral | Medida | Condición | Cumple | Hace fallar el job |
|---|---|---|---|---|
| `sin_activar:claude-sonnet-5-5:orden` | 0 de 54 (0,0 %) | ≤ 0,0 % | sí | sí |
| `afirma_lo_no_leido:claude-sonnet-5-5:orden` | 0 de 54 (0,0 %) | ≤ 0,0 % | sí | sí |
| `cuenta_su_proceso:claude-sonnet-5-5:orden` | 0 de 54 (0,0 %) | ≤ 0,0 % | sí | no: solo se publica, no es un control |
| `sin_activar:claude-sonnet-5-5:herramienta` | 0 de 54 (0,0 %) | ≤ 0,0 % | sí | sí |
| `afirma_lo_no_leido:claude-sonnet-5-5:herramienta` | 0 de 54 (0,0 %) | ≤ 0,0 % | sí | sí |
| `cuenta_su_proceso:claude-sonnet-5-5:herramienta` | 2 de 54 (3,7 %) | ≤ 0,0 % | ✗ **NO** | no: solo se publica, no es un control |
| `medida_del_juez:afirma_lo_no_leido:defectos_sin_marcar` | 0 de 212 (0,0 %) | ≤ 0,0 % | sí | sí |
| `medida_del_juez:afirma_lo_no_leido:correctos_marcados` | 0 de 47 (0,0 %) | ≤ 0,0 % | sí | sí |
| `duracion_de_las_sesiones:orden` | 446 | ≤ 900 | sí | sí |
| `duracion_de_las_sesiones:herramienta` | 454 | ≤ 900 | sí | sí |
| `duracion_del_juez:orden` | 93 | ≤ 900 | sí | sí |
| `duracion_del_juez:herramienta` | 107 | ≤ 900 | sí | sí |

**legal-core**: aprobado sobre `3107ae2`; decide `claude-sonnet-5-5` con 2 de 3; 4 evals, 0 nuevas, 0 informativas.

| Eval | claude-sonnet-5-5 | claude-haiku-4-5-20251001 | claude-sonnet-5-5 (herramienta) | claude-haiku-4-5-20251001 (herramienta) | Marca |
|---|---|---|---|---|---|
| `01-territorio-municipio-cubierto.yaml` | 3/3 | 3/3 | 3/3 | 3/3 |  |
| `02-territorio-municipio-no-cubierto.yaml` | 3/3 | 3/3 | 3/3 | 2/3 |  |
| `03-no-activa-receta-de-cocina.yaml` | 3/3 | 3/3 | 3/3 | 3/3 |  |
| `04-sin-binario-ni-servidor.yaml` | 3/3 | 3/3 | — | — |  |

Umbrales: el job no publica ninguno para esta skill.

Cada celda: sesiones que pasan de las abiertas; ✗, una serie que decide y no llega al umbral. Una eval informativa publica su tasa sin decidir el veredicto (ADR 0016). La regla por serie no hace cumplir un umbral agregado sobre todas las respuestas: eso solo lo hace un umbral del job que lo hace fallar (ADR 0029).

## 4. Revisión que la constitución reserva a la persona

- Fixture o esquema EXISTENTE modificado: `schemas/expresiones-prohibidas.yaml.json`.
- Fixtures nuevos (2), por directorio:
  - `testdata/evals/retiradas/`: `19-lpac-articulo-21-redaccion-cambiada.yaml`
  - `testdata/evals/retiradas/grafo-previo/lpac-a21-version-anterior/`: `GET_https_www.boe.es_datosabiertos_api_legislacion-consolidada_id_BOE-A-2015-10565_texto_bloque_a21.json`
- Esquemas nuevos: `schemas/juez-clases.yaml.json`.

## 5. Tareas en cuarentena

Ninguna.

## 6. Trazabilidad

**Estado**: «tareas hechas» solo dice que las tareas que citan el requisito están marcadas; nada del run lo ha medido. «comprobado por su control» exige su fila en «Controles de umbral» de `plan.md` y que el control esté: un test o una comprobación de `make ci` que existe en la cabeza, con `make ci` en verde, o un umbral que el job de evals publica, cumple y hace fallar el job (ADR 0029). «UMBRAL NO CUMPLIDO» y «CONTROL SIN VERIFICAR» son lo que hay que mirar.

| Requisito | Tareas | Aceptación | Estado | Control de umbral |
|---|---|---|---|---|
| FR-001 | T002(hecha) | — | tareas hechas | — |
| FR-002 | T002(hecha) | — | tareas hechas | — |
| FR-003 | T003(hecha) | — | tareas hechas | — |
| FR-004 | T002(hecha) T003(hecha) | — | tareas hechas | — |
| FR-005 | T002(hecha) | — | tareas hechas | — |
| FR-006 | T002(hecha) | — | tareas hechas | — |
| FR-007 | T002(hecha) T003(hecha) T006(hecha) | — | comprobado por su control | `ci:internal/evals/informe_test.go:TestInformeConElJuez`, en make ci (verde) |
| FR-010 | T002(hecha) | — | tareas hechas | — |
| FR-011 | T002(hecha) | — | tareas hechas | — |
| FR-012 | T006(hecha) | — | tareas hechas | — |
| FR-013 | T006(hecha) T007(hecha) | — | tareas hechas | — |
| FR-014 | T006(hecha) | — | tareas hechas | — |
| FR-020 | T001(hecha) | — | tareas hechas | — |
| FR-021 | T001(hecha) | — | tareas hechas | — |
| FR-022 | T001(hecha) | — | tareas hechas | — |
| FR-023 | T001(hecha) | — | comprobado por su control | `ci:internal/evals/medida_test.go:TestCopiasDelJuez`, en make ci (verde) |
| FR-024 | T008(hecha) T009(hecha) | — | tareas hechas | — |
| FR-030 | T006(hecha) | — | comprobado por su control | `evals:boe-legislacion:afirma_lo_no_leido:claude-sonnet-5-5:orden`: 0 de 54 (0,0 %), ≤ 0,0 %; `evals:boe-legislacion:afirma_lo_no_leido:claude-sonnet-5-5:herramienta`: 0 de 54 (0,0 %), ≤ 0,0 % |
| FR-031 | T006(hecha) | — | tareas hechas | — |
| FR-032 | T004(hecha) T006(hecha) | — | comprobado por su control | `evals:boe-legislacion:medida_del_juez:afirma_lo_no_leido:defectos_sin_marcar`: 0 de 212 (0,0 %), ≤ 0,0 %; `evals:boe-legislacion:medida_del_juez:afirma_lo_no_leido:correctos_marcados`: 0 de 47 (0,0 %), ≤ 0,0 % |
| FR-033 | T006(hecha) | — | comprobado por su control | `evals:boe-legislacion:duracion_del_juez:orden`: 93, ≤ 900; `evals:boe-legislacion:duracion_del_juez:herramienta`: 107, ≤ 900 |
| FR-034 | T005(hecha) | — | comprobado por su control | `evals:boe-legislacion:sin_activar:claude-sonnet-5-5:orden`: 0 de 54 (0,0 %), ≤ 0,0 %; `evals:boe-legislacion:sin_activar:claude-sonnet-5-5:herramienta`: 0 de 54 (0,0 %), ≤ 0,0 %; `evals:boe-legislacion:duracion_de_las_sesiones:orden`: 446, ≤ 900; `evals:boe-legislacion:duracion_de_las_sesiones:herramienta`: 454, ≤ 900 |
| FR-035 | T006(hecha) | — | tareas hechas | — |
| FR-036 | T006(hecha) T017(hecha) | — | tareas hechas | — |
| FR-037 | T006(hecha) | — | tareas hechas | — |
| FR-040 | T001(hecha) | — | tareas hechas | — |
| FR-041 | T004(hecha) | — | tareas hechas | — |
| FR-042 | T004(hecha) | — | comprobado por su control | `ci:internal/evals/medida_test.go:TestMedidaVersionada`, en make ci (verde) |
| FR-043 | T010(hecha) T011(hecha) | — | tareas hechas | — |
| FR-044 | T006(hecha) T011(hecha) | — | tareas hechas | — |
| FR-045 | T004(hecha) T017(hecha) | — | tareas hechas | — |
| FR-050 | T010(hecha) T012(hecha) T013(hecha) | — | tareas hechas | — |
| FR-051 | T009(hecha) T010(hecha) | — | tareas hechas | — |
| FR-052 | T010(hecha) T012(hecha) | — | tareas hechas | — |
| FR-053 | T010(hecha) T012(hecha) | — | tareas hechas | — |
| FR-054 | T010(hecha) T012(hecha) | — | tareas hechas | — |
| FR-060 | T007(hecha) | — | tareas hechas | — |
| FR-061 | T007(hecha) | — | tareas hechas | — |
| FR-062 | T005(hecha) | — | tareas hechas | — |
| FR-070 | T005(hecha) | — | tareas hechas | — |
| FR-071 | T005(hecha) | — | tareas hechas | — |
| FR-075 | T014(hecha) | — | tareas hechas | — |
| FR-076 | T014(hecha) | — | tareas hechas | — |
| FR-080 | T015(hecha) | — | tareas hechas | — |
| FR-081 | T015(hecha) | — | tareas hechas | — |
| FR-082 | T015(hecha) | — | tareas hechas | — |
| FR-083 | T015(hecha) | — | tareas hechas | — |
| FR-084 | T015(hecha) | — | tareas hechas | — |
| FR-085 | T015(hecha) T018(hecha) | — | tareas hechas | — |
| FR-086 | T015(hecha) | — | comprobado por su control | `ci:Makefile:skills-check`, en make ci (verde); `ci:internal/evals/conjunto_test.go:TestProsaDeLaSkill`, en make ci (verde); `ci:internal/evals/conjunto_test.go:TestEvalsDelRepositorio`, en make ci (verde) |
| FR-087 | T016(hecha) | — | tareas hechas | — |
| FR-090 | T004(hecha) T013(hecha) | — | tareas hechas | — |
| FR-091 | T004(hecha) T013(hecha) | — | tareas hechas | — |
| FR-092 | T013(hecha) | — | comprobado por su control | `ci:internal/evals/definicion_test.go:TestDefinicionDelJob`, en make ci (verde) |
| FR-093 | T003(hecha) T011(hecha) T012(hecha) | — | tareas hechas | — |
| FR-095 | T016(hecha) | — | tareas hechas | — |
| FR-096 | T016(hecha) T017(hecha) | — | tareas hechas | — |
| FR-100 | T017(hecha) T018(hecha) | — | tareas hechas | — |
| FR-101 | T017(hecha) | — | tareas hechas | — |
| FR-102 | T002(hecha) | — | comprobado por su control | `ci:internal/evals/juez_test.go:TestVotoDelJuez`, en make ci (verde) |
| FR-103 | T002(hecha) | — | comprobado por su control | `ci:internal/evals/juez_test.go:TestReglaDeLosVotos`, en make ci (verde) |
| FR-104 | T006(hecha) T007(hecha) T011(hecha) | — | comprobado por su control | `ci:internal/evals/informe_test.go:TestInformeConElJuez`, en make ci (verde); `ci:internal/evals/ejecucion_test.go:TestEjecucionSinMedir`, en make ci (verde) |
| FR-105 | T004(hecha) | — | comprobado por su control | `ci:internal/evals/medida_test.go:TestMedidaVersionada`, en make ci (verde) |
| FR-106 | T010(hecha) T012(hecha) | — | comprobado por su control | `ci:internal/evals/medida_test.go:TestEjecucionDeLaMedida`, en make ci (verde) |
| FR-107 | T002(hecha) T003(hecha) | — | comprobado por su control | `ci:internal/evals/juez_test.go:TestMensajeDelVoto`, en make ci (verde); `ci:internal/evals/juez_test.go:TestOrdenDelVoto`, en make ci (verde) |
| FR-108 | T001(hecha) | — | comprobado por su control | `ci:internal/evals/medida_test.go:TestCopiasDelJuez`, en make ci (verde) |
| FR-109 | T008(hecha) T009(hecha) | — | comprobado por su control | `ci:internal/evals/medida_test.go:TestGrabacionesDerivadas`, en make ci (verde) |
| FR-110 | T015(hecha) | — | comprobado por su control | `ci:Makefile:skills-check`, en make ci (verde); `ci:internal/evals/conjunto_test.go:TestProsaDeLaSkill`, en make ci (verde); `ci:internal/evals/conjunto_test.go:TestEvalsDelRepositorio`, en make ci (verde) |
| FR-111 | T005(hecha) | — | comprobado por su control | `ci:internal/evals/juzgar_test.go:TestJuzgarSinLaLista`, en make ci (verde) |
| FR-112 | T013(hecha) | — | comprobado por su control | `ci:internal/evals/definicion_test.go:TestDefinicionDelJob`, en make ci (verde) |
| FR-113 | T016(hecha) | — | tareas hechas | — |
| SC-001 | — | — | SIN TAREA | `evals:boe-legislacion:afirma_lo_no_leido:claude-sonnet-5-5:orden`: 0 de 54 (0,0 %), ≤ 0,0 %; `evals:boe-legislacion:afirma_lo_no_leido:claude-sonnet-5-5:herramienta`: 0 de 54 (0,0 %), ≤ 0,0 %; `evals:boe-legislacion:medida_del_juez:afirma_lo_no_leido:defectos_sin_marcar`: 0 de 212 (0,0 %), ≤ 0,0 %; `evals:boe-legislacion:medida_del_juez:afirma_lo_no_leido:correctos_marcados`: 0 de 47 (0,0 %), ≤ 0,0 %; `evals:boe-legislacion:duracion_del_juez:orden`: 93, ≤ 900; `evals:boe-legislacion:duracion_del_juez:herramienta`: 107, ≤ 900; `evals:boe-legislacion:sin_activar:claude-sonnet-5-5:orden`: 0 de 54 (0,0 %), ≤ 0,0 %; `evals:boe-legislacion:sin_activar:claude-sonnet-5-5:herramienta`: 0 de 54 (0,0 %), ≤ 0,0 %; `evals:boe-legislacion:duracion_de_las_sesiones:orden`: 446, ≤ 900; `evals:boe-legislacion:duracion_de_las_sesiones:herramienta`: 454, ≤ 900; `ci:internal/evals/informe_test.go:TestInformeConElJuez`, en make ci (verde) |
| SC-002 | T002(hecha) | — | comprobado por su control | `ci:internal/evals/juez_test.go:TestVotoDelJuez`, en make ci (verde) |
| SC-003 | T002(hecha) | — | comprobado por su control | `ci:internal/evals/juez_test.go:TestReglaDeLosVotos`, en make ci (verde) |
| SC-004 | T006(hecha) T007(hecha) T011(hecha) T014(hecha) | — | comprobado por su control | `ci:internal/evals/informe_test.go:TestInformeConElJuez`, en make ci (verde); `ci:internal/evals/ejecucion_test.go:TestEjecucionSinMedir`, en make ci (verde) |
| SC-005 | T004(hecha) | — | comprobado por su control | `ci:internal/evals/medida_test.go:TestMedidaVersionada`, en make ci (verde) |
| SC-006 | T010(hecha) | — | comprobado por su control | `ci:internal/evals/medida_test.go:TestEjecucionDeLaMedida`, en make ci (verde) |
| SC-007 | T002(hecha) T003(hecha) | — | comprobado por su control | `ci:internal/evals/juez_test.go:TestMensajeDelVoto`, en make ci (verde); `ci:internal/evals/juez_test.go:TestOrdenDelVoto`, en make ci (verde) |
| SC-008 | T001(hecha) | — | comprobado por su control | `ci:internal/evals/medida_test.go:TestCopiasDelJuez`, en make ci (verde) |
| SC-009 | T008(hecha) T009(hecha) | — | comprobado por su control | `ci:internal/evals/medida_test.go:TestGrabacionesDerivadas`, en make ci (verde) |
| SC-010 | T015(hecha) | — | comprobado por su control | `ci:Makefile:skills-check`, en make ci (verde); `ci:internal/evals/conjunto_test.go:TestProsaDeLaSkill`, en make ci (verde); `ci:internal/evals/conjunto_test.go:TestEvalsDelRepositorio`, en make ci (verde) |
| SC-011 | T005(hecha) | — | comprobado por su control | `ci:internal/evals/juzgar_test.go:TestJuzgarSinLaLista`, en make ci (verde) |
| SC-012 | T013(hecha) | — | comprobado por su control | `ci:internal/evals/definicion_test.go:TestDefinicionDelJob`, en make ci (verde) |
| SC-013 | T016(hecha) T017(hecha) T018(hecha) | — | tareas hechas | — |
| SC-014 | — | — | SIN TAREA | — |

## 7. Cambios posteriores a la revisión final

Cada commit posterior a lo que juzgó la primera ronda de la revisión final (`81d025d`), con la ronda que lo vio y su veredicto, y lo que toca fuera de `gates/` (ADR 0030):

- `10484b5` fix(H24): motivos de la revisión final: lo vio la ronda 2 (juez A: aprobado; juez B: rechazado). Toca `internal/evals/conjunto_test.go`, `internal/evals/definicion.go`, `internal/evals/definicion_test.go`, `internal/evals/informe.go`, `internal/evals/informe_test.go`, `internal/evals/juez.go`, `internal/evals/medida.go`, `internal/evals/medida_test.go`, `internal/evals/sondeo.go`, `cierre.md`, `data-model.md`, `tasks.md`.
- `ae250eb` fix(H24): motivos de la revisión final: lo vio la ronda 3 (juez A: aprobado; juez B: aprobado). Toca `internal/evals/conjunto_test.go`, `internal/evals/medida.go`, `cierre.md`, `tasks.md`.
- `422f339` docs(H24): veredictos de la revisión final: solo registros de `gates/`.
- `a5bda45` docs(H24): registros del run: solo registros de `gates/`.
- `b50e05b` fix(H24): cierre en la plataforma: lo vio la ronda 4 (juez A: rechazado; juez B: rechazado). Toca `CHANGELOG.md`, `skills/boe-legislacion/SKILL.md`, `contracts/skill-boe-legislacion.md`, `plan.md`, `research.md`.
- `51db15e` fix(H24): motivos de la revisión final: lo vio la ronda 5 (juez A: aprobado; juez B: aprobado). Toca `cierre.md`, `contracts/skill-boe-legislacion.md`, `plan.md`.
- `72985b4` docs(H24): veredictos de la revisión final: solo registros de `gates/`.
- `b0a7a9c` docs(H24): registros del run: solo registros de `gates/`.
- `d53bc89` fix(H24): cierre en la plataforma: lo vio la ronda 6 (juez A: aprobado; juez B: aprobado). Toca `CHANGELOG.md`, `skills/boe-legislacion/SKILL.md`, `cierre.md`, `contracts/skill-boe-legislacion.md`, `plan.md`, `research.md`.
- `4290e98` docs(H24): veredictos de la revisión final: solo registros de `gates/`.
- `3107ae2` docs(H24): registros del run: solo registros de `gates/`.

### Cambios que ningún juez vio

Ninguno.

## 8. Cómo comprobarlo y consumo

- Escenarios manuales: `specs/017-h24-las-evals-juzgan/quickstart.md`. Suite de aceptación congelada: `specs/017-h24-las-evals-juzgan/aceptacion/` (activada en `internal/app/testdata/script/`).
- Run `30513ba1`: 16 h 54 min de reloj. Coste por paso y por rol: `scripts/coste-run.sh 30513ba1`.
