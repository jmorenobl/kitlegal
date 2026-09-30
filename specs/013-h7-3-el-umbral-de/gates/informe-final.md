# Informe del hito H7.3 · El umbral de expresiones prohibidas decide, `boe-legislacion` sin el vocabulario que provoca el ruido, y evals en paralelo con un sondeo local

Generado por el workflow `hito` el 2026-09-30T08:49:39Z, sobre `8f85eb9` de `013-h7-3-el-umbral-de`.
Lo escribe scripts/workflow/informe.sh sin modelo, desde los artefactos de `specs/013-h7-3-el-umbral-de/`. Fusionar (squash-merge) es una decisión humana:
si algo de lo que sigue no es lo que se quería, se corrige la sección del hito en docs/ROADMAP.md y se relanza.

## 1. Estado

- **make ci local**: verde.
- **CI y evals remotos** sobre `196ee05`: verde.
- **Evals por skill**: boe-legislacion aprobado, legal-core aprobado (tasas en la sección 3).
- **Umbrales del job**: boe-legislacion: 3 umbrales, **1 sin cumplir o solo publicados** (`expresiones_prohibidas:claude-haiku-4-5-20251001`); legal-core: sin umbrales (sección 3).
- **Revisión final**: juez A aprobado, juez B aprobado, 3 rondas.
- **Tareas**: 15 hechas, 0 en cuarentena, 0 pendientes sin cuarentena.
- **Diff**: 26 commits; 93 files changed, 19818 insertions(+), 614 deletions(-).

## 2. Supuestos y pendientes

Decisiones que el run tomó sin preguntar, ordenadas por impacto: cada paso que escribe una etiqueta la suya (ADR 0028).

### Cambian el comportamiento visible, el alcance o una skill

**Comportamiento (salida, códigos, ficheros, argumentos)**

- corrector_spec: FR-032 («el orden no cambia el informe») contradecía FR-044 (tras un límite de uso, las abiertas se juzgan y las demás quedan sin medir) → FR-032 se acota a que el juicio de cada sesión depende solo de su transcript y su traza, y el orden no cambia el informe salvo los tiempos y, tras un límite de uso, qué sesiones se llegaron a abrir; el test de FR-094, SC-008 y US3-1 comparan con la ejecución en serie sin límites de uso.
- corrector_spec: FR-065 (solo lo agregado) no dejaba publicar la sesión que no termina con su motivo que pide FR-063 → la salida del sondeo publica también cada sesión que no terminó por otra causa, con su motivo, como el job, y sigue sin listar las expresiones por sesión; la fila «La salida del sondeo» de «Uso» lo acota a una línea por sesión pedida como mucho.
- corrector_spec: FR-034 no decía qué deja el segundo disparo al cierre, que lo consume (`medir` y `recoger_evals`) → elija el plan esperar o saltarse, el segundo disparo no deja ninguna comprobación (cancelada, saltada ni de otro estado) que el cierre lea en lugar de la de la tanda que corre ni que le haga dejar de esperarla; el cierre espera a esa tanda y recoge su informe, y la comprobación de la definición del job (FR-094, SC-008, US3-6) falla si eso puede ocurrir. Se resuelve en la definición… (entera en `specs/013-h7-3-el-umbral-de/gates/supuestos.md` o `clarify-respuestas.json`)
- corrector_plan: [i] el texto del `result` con `is_error` no llegaba al motivo de una sesión sin terminar cuando `claude -p` sale con código 1, que es lo que da una credencial que no sirve (H5 V61) → el motivo lleva el texto del último `result` con `is_error` sea cual sea el código (`código <n>: result con is_error: <texto>`; con 0, `result con is_error: <texto>`; los del tope no cambian), en el informe del job y en la salida del sondeo; V18 lo verifica (también el mensaje de un modelo que no exi… (entera en `specs/013-h7-3-el-umbral-de/gates/supuestos.md` o `clarify-respuestas.json`)
- T003: el spec no decía qué es `ErrorDelResultado` cuando el último mensaje es un `result` con `is_error` sin la clave `result` (p. ej. `error_during_execution`) → vacío, y la sesión sigue siendo legible con el motivo de hoy (`código <n>` o `result con subtype …`): exigir el texto haría ilegible una sesión que hoy se lee, que es más comportamiento; lo fija el caso `result-con-is-error-sin-result` de `TestLeerSesionConReintentos`.
- T004: el contrato fija que una sesión sin medir «no pasa, lleva como único motivo `sin medir por límite de uso: <clase>`», pero no dice qué publican sus demás campos ni la celda «Resultado» de la tabla «Sesiones» → se publican tal como los da `Juzgar` (comandos, citas, avisos y territorio repartidos, invocaciones, llegadas a la red) y la celda dice «no pasa», con la clase en la columna «Sin medir»: vaciar los repartos o cambiar la celda serían reglas que el spec no pide, y lo observado de la ses… (entera en `specs/013-h7-3-el-umbral-de/gates/supuestos.md` o `clarify-respuestas.json`)
- T005: contracts/informe-del-job.md §4 escribe las filas de la tabla «Umbrales» fuera de un bloque de código, con el nombre del umbral entre comillas invertidas, y no dice si esas comillas son de informe.md o solo del formato del contrato → informe.md escribe el nombre entre comillas invertidas, byte a byte como las filas del contrato (`TestInformeMarkdownDeLosUmbrales` las compara fila a fila): es la lectura literal y no añade nada más; el nombre no lleva `|` que escapar ni sale en los motivos c… (entera en `specs/013-h7-3-el-umbral-de/gates/supuestos.md` o `clarify-respuestas.json`)
- T007: la tarea dice que «un error de preparación, de lectura o de escritura» cierra las abiertas y devuelve el error justo después de pedir que el repartidor lea cada sesión con `LeerSesion`, y no dice si una sesión que `LeerSesion` no puede leer (un transcript con una línea que no es un mensaje de `stream-json`, p. ej. la que el tope corta a medias) es ese error de lectura; el contrato §3.6 habla de «un error de preparación o de E/S» → no lo es: la sesión ilegible no detiene el reparto ni cierr… (entera en `specs/013-h7-3-el-umbral-de/gates/supuestos.md` o `clarify-respuestas.json`)
- T010: contracts/sondeo.md §4 da a la salida del sondeo dos apartados de sesiones —las sin medir por límite de uso y las que no terminaron por otra causa, con `la sesión no terminó: ` y `MotivoSinTerminar`— y no dice dónde va una sesión que `LeerSesion` no puede leer (p. ej., un transcript con la última línea cortada a medias), que no es ninguna de las dos: no se sabe si terminó → cuenta en su serie como no pasada, como en el job (motivo `sesión ilegible: …` en su resultado), y la salida no la li… (entera en `specs/013-h7-3-el-umbral-de/gates/supuestos.md` o `clarify-respuestas.json`)
- T011: contracts/sondeo.md §3.1 no dice qué es `EVALS` con un número repetido (`03,14,03`) → no vale, con el error de su forma, antes de construir nada: abrirlo tal cual haría fallar el repartidor en el directorio repetido tras consumir sesiones, y quitar el repetido cambiaría en silencio lo que se mide; el contrato §3.1 lo dice y `TestComprobarElSondeo` lo fija.
- T011: contracts/sondeo.md §3.1 no dice si los errores de varios argumentos salen a la vez ni qué forma tiene «un entero mayor o igual que 1» → salen juntos, uno por línea y en el orden de la tabla (con un `SKILL` que no vale, de `EVALS` solo se comprueba la forma), como `exigirBanderas` con las banderas del job; y el entero tiene la forma `^[1-9][0-9]*$`, la que `scripts/evals.sh` exige a `CONCURRENCIA_DE_EVALS` (T009), así que `03` o `+3` no valen; el contrato §3.1 lo dice.
- T012: contracts/sondeo.md §2 dice «uso y código 1» sin fijar el texto del uso de `scripts/evals-sondeo.sh` con otro número de argumentos → `evals-sondeo: uso: scripts/evals-sondeo.sh <skill> <evals> <modelo> <repeticiones> <concurrencia>` en la salida de error, sin crear el temporal ni ejecutar go, la forma del de `scripts/evals.sh`; `make evals-sondeo` le pasa siempre cinco, así que solo lo ve quien invoca el guion a mano.
- corrector_revision: [b][f][l] el sondeo escribía fuera de su temporal —la preparación de cada sesión, en proceso, copia las grabaciones con `os.MkdirTemp("")` en el `TMPDIR` de quien lo lanza, y ahí quedaban si el proceso moría sin llegar a retirarlas— y el control de la fila FR-064/SC-009 no lo veía (`TestSondear` mira el `TMPDIR` de la base que recibe `sondear`, no el del proceso; el `go` sustituto de `TestGuionDelSondeo` no escribía nada) → `scripts/evals-sondeo.sh` crea `<temporal>/tmp` y ej… (entera en `specs/013-h7-3-el-umbral-de/gates/supuestos.md` o `clarify-respuestas.json`)
- clarify Q2 (criterio d, conservadora): El sondeo acepta una sola credencial: `CLAUDE_CODE_OAUTH_TOKEN` en el entorno, como el job, el token de la suscripción que da `claude setup-token`. Es la única que pasa a sus sesiones: no usa ninguna clave de API de pago por uso ni lee la sesión iniciada de Claude Code de la persona, es decir, ni su llavero ni sus ficheros de credenciales. Antes de abrir ninguna sesión basta con comprobar que la variable está y no está vacía. Si falta o está vacía, el sonde… (entera en `specs/013-h7-3-el-umbral-de/gates/supuestos.md` o `clarify-respuestas.json`)
- clarify Q3 (criterio c): El sondeo publica solo lo agregado que pide FR-065: la tasa de cada serie, el recuento de respuestas con alguna expresión prohibida sobre las respuestas en evals que activan la skill (con el 5 % como referencia), lo que no comprueba y las sesiones sin medir por límite de uso. No lista las expresiones encontradas en cada sesión. Su directorio temporal (sesiones, cachés preparadas, transcripts, respuestas y el estado de Claude Code de cada sesión) se borra al terminar, tam… (entera en `specs/013-h7-3-el-umbral-de/gates/supuestos.md` o `clarify-respuestas.json`)
- clarify Q4 (criterio c): Como el job (FR-044). Tras una sesión cuyo resultado lleva el mensaje del límite de uso (el de la ventana, el semanal o el de la familia del modelo, que no se reponen mientras dura el sondeo), el sondeo no abre ninguna sesión más. Las que ya estaban abiertas terminan y se juzgan, y las que faltaban quedan sin medir por límite de uso. Su serie queda sin medir, y la salida publica esas sesiones como sin medir por límite de uso (FR-065). Sale con 0: ha abierto y juzgado lo … (entera en `specs/013-h7-3-el-umbral-de/gates/supuestos.md` o `clarify-respuestas.json`)

**Alcance (lo que queda fuera o dentro del hito)**

- T011: el `go install` escribe la telemetría del go command en el directorio de configuración de Go del `HOME` de quien lo lanza (`go env GOTELEMETRYDIR`), y no se apaga desde el entorno (comprobado: con `GOTELEMETRY=off` en el entorno la escribe igual), así que «el `HOME` de la base no cambia» de contracts/sondeo.md §7 no se cumple al pie de la letra → se trata como las cachés de compilación de Go que exceptúa FR-064: es estado del propio go command, que el `go test` del guion del sondeo escribe… (entera en `specs/013-h7-3-el-umbral-de/gates/supuestos.md` o `clarify-respuestas.json`)
- T015: la tarea solo pide tests si la cobertura global (≥ 70 %) o la del dominio (≥ 85 %) quedan bajo su umbral, y no lo hacen (97,4 % y 98,6 % en la unión de los dos perfiles), pero el estado `patch` de `codecov.yml` (`target: auto`, bloqueante) compara el diff con la base, y la estimación local por líneas da 93,98 % (890 de 947), por debajo del 97,5 % que midió el cierre de H7.2 → no se añade ningún test, como en H7.1 y H7.2: las 57 líneas sin ejecutar son ramas de error de la regla genérica en… (entera en `specs/013-h7-3-el-umbral-de/gates/supuestos.md` o `clarify-respuestas.json`)
- reparar_cierre: el paso pide reproducir el fallo en local si se puede, y lo que lo reproduce es el sondeo (`make evals-sondeo`), que abre sesiones con modelo → no se ha ejecutado: FR-068 («ninguna tarea ni ningún paso del workflow lo ejecuta») y el entorno del run no tiene `CLAUDE_CODE_OAUTH_TOKEN`. La corrección se apoya en las 51 respuestas del informe del job y en lo comprobado sin modelo con el binario del árbol (la orden encadenada sobre la sesión preparada de la eval 19, con la lectura fal… (entera en `specs/013-h7-3-el-umbral-de/gates/supuestos.md` o `clarify-respuestas.json`)

**Skill (lo que pide, dice o comprueba una skill)**

- reparar_cierre: el job de cierre sobre `6ab3add` no cumple SC-001 —`expresiones_prohibidas:claude-sonnet-5`, 5 de 51 (9,8 %; el umbral es ≤ 5 %), y la eval 06 con 1 de 3—, y FR-015 pedía mantener «una vez por norma citada» si la comprobación cambiaba de sitio → la comprobación de la redacción pasa a ir en la misma orden que la lectura (`kitlegal boe articulo <norma> <bloque> --json && kitlegal graph check <norma> <bloque> --json`), para que la última orden antes de la respuesta traiga el texto q… (entera en `specs/013-h7-3-el-umbral-de/gates/supuestos.md` o `clarify-respuestas.json`)
- reparar_cierre: de las 8 respuestas del cierre con el párrafo de transición, 3 (03-03, 13-03 y 14-03: «Sin cambios de redacción respecto a lecturas anteriores. Aquí está la respuesta.», «Sin cambios desde una lectura anterior. Respondo.», «No hay avisos de vigencia ni cambios de redacción. Con esto tengo la respuesta completa.») no llevan ninguna expresión de la lista, así que el umbral las cuenta como limpias → la lista no se amplía en el cierre: el arreglo de una eval en rojo va a la skill, no… (entera en `specs/013-h7-3-el-umbral-de/gates/supuestos.md` o `clarify-respuestas.json`)

### Del propio run

- sesión posterior al run (2026-09-30): tercera ronda de la revisión final, a mano, sobre `3dda87f`: los dos jueces aprueban sin motivos el cambio de protocolo de `eb6b4c8` («una vez por cada orden que lee bloques» en lugar de «una vez por norma citada»: la mejor solución sin tocar el binario, registrada como decisión) y el arreglo del rojo de `ci` sobre `e3bae8d` (`0222c39`: ETXTBSY en `TestSondear`, sustitutos escritos con `syscall.ForkLock`; 50 de 50 con `-race`). Detalle: el cambio de protocol… (entera en `specs/013-h7-3-el-umbral-de/gates/supuestos.md` o `clarify-respuestas.json`)

- Observaciones de los jueces, por debajo del umbral y sin corregir: 13 en `spec-r2.json`, 6 en `plan-r2.json`, 4 en `tasks-r3.json`, 8 en `revision-a-r3.json`, 10 en `revision-b-r3.json`.

### Internos (16)

Decisiones que no cambian nada observable (técnica, estructura, tests), por autor: T001 (1), T006 (3), T007 (1), T008 (1), T011 (2), T012 (2), T013 (1), T014 (2), T015 (1), barrido (1), sesión posterior al run (2026-09-30) (1). Enteras en `specs/013-h7-3-el-umbral-de/gates/supuestos.md`.

## 3. Evals sobre la cabeza

**boe-legislacion**: aprobado sobre `196ee05`; decide `claude-sonnet-5` con 2 de 3; 19 evals, 0 nuevas, 7 informativas.

| Eval | claude-sonnet-5 | claude-haiku-4-5-20251001 | Marca |
|---|---|---|---|
| `01-lpac-articulo-21.yaml` | 3/3 | 3/3 |  |
| `02-lcsp-contrato-menor.yaml` | 3/3 | 3/3 |  |
| `03-lrbrl-atribuciones-del-pleno.yaml` | 3/3 | 3/3 |  |
| `04-lgt-prescripcion.yaml` | 3/3 | 3/3 |  |
| `05-trlrhl-impuestos-municipales.yaml` | 3/3 | 3/3 |  |
| `06-irpf-rendimientos-del-trabajo.yaml` | 3/3 | 2/3 |  |
| `07-lrjsp-principio-de-legalidad.yaml` | 3/3 | 3/3 |  |
| `08-ltaibg-plazo-de-resolucion.yaml` | 3/3 | 3/3 |  |
| `09-constitucion-articulo-140.yaml` | 3/3 | 3/3 |  |
| `10-et-vacaciones.yaml` | 3/3 | 3/3 |  |
| `11-no-activa-programacion.yaml` | 3/3 | 3/3 |  |
| `12-no-activa-acuerdo-entre-amigos.yaml` | 3/3 | 3/3 |  |
| `13-lrbrl-atribuciones-por-materia.yaml` | 3/3 | — | informativa |
| `14-trlrhl-impuestos-por-materia.yaml` | 2/3 | — | informativa |
| `15-irpf-rendimientos-por-materia.yaml` | 3/3 | — | informativa |
| `16-lrjsp-legalidad-por-materia.yaml` | 3/3 | — | informativa |
| `17-ltaibg-plazo-por-materia.yaml` | 3/3 | — | informativa |
| `18-lrjpac-norma-derogada.yaml` | 3/3 | — | informativa |
| `19-lcsp-contrato-menor-redaccion-cambiada.yaml` | 3/3 | — | informativa |

Respuestas con alguna expresión prohibida, en las evals que activan la skill (`expresiones_prohibidas_por_modelo`):

| Modelo | Con alguna | Respuestas | Porcentaje |
|---|---|---|---|
| `claude-sonnet-5` | 1 | 51 | 2,0 % |
| `claude-haiku-4-5-20251001` | 0 | 30 | 0,0 % |

Umbrales que publica el job (ADR 0029):

| Umbral | Medida | Condición | Cumple | Hace fallar el job |
|---|---|---|---|---|
| `expresiones_prohibidas:claude-sonnet-5` | 1 de 51 (2,0 %) | ≤ 5,0 % | sí | sí |
| `expresiones_prohibidas:claude-haiku-4-5-20251001` | 0 de 30 (0,0 %) | ≤ 5,0 % | sí | no: solo se publica, no es un control |
| `duracion_de_las_sesiones` | 492 | ≤ 900 | sí | sí |

**legal-core**: aprobado sobre `196ee05`; decide `claude-sonnet-5` con 2 de 3; 3 evals, 0 nuevas, 0 informativas.

| Eval | claude-sonnet-5 | claude-haiku-4-5-20251001 | Marca |
|---|---|---|---|
| `01-territorio-municipio-cubierto.yaml` | 3/3 | 3/3 |  |
| `02-territorio-municipio-no-cubierto.yaml` | 3/3 | 3/3 |  |
| `03-no-activa-receta-de-cocina.yaml` | 3/3 | 3/3 |  |

Umbrales: el job no publica ninguno para esta skill.

Cada celda: sesiones que pasan de las abiertas; ✗, una serie que decide y no llega al umbral. Una eval informativa publica su tasa sin decidir el veredicto (ADR 0016). La regla por serie no hace cumplir un umbral agregado sobre todas las respuestas: eso solo lo hace un umbral del job que lo hace fallar (ADR 0029).

## 4. Revisión que la constitución reserva a la persona

- Fixture o esquema EXISTENTE modificado: `schemas/expresiones-prohibidas.yaml.json`.

## 5. Tareas en cuarentena

Ninguna.

## 6. Trazabilidad

**Estado**: «tareas hechas» solo dice que las tareas que citan el requisito están marcadas; nada del run lo ha medido. «comprobado por su control» exige su fila en «Controles de umbral» de `plan.md` y que el control esté: un test o una comprobación de `make ci` que existe en la cabeza, con `make ci` en verde, o un umbral que el job de evals publica, cumple y hace fallar el job (ADR 0029). «UMBRAL NO CUMPLIDO» y «CONTROL SIN VERIFICAR» son lo que hay que mirar.

| Requisito | Tareas | Aceptación | Estado | Control de umbral |
|---|---|---|---|---|
| FR-001 | T005(hecha) | — | tareas hechas | — |
| FR-002 | T005(hecha) T015(hecha) | — | comprobado por su control | `evals:boe-legislacion:expresiones_prohibidas:claude-sonnet-5`: 1 de 51 (2,0 %), ≤ 5,0 % |
| FR-003 | T005(hecha) | — | comprobado por su control | `evals:boe-legislacion:expresiones_prohibidas:claude-sonnet-5`: 1 de 51 (2,0 %), ≤ 5,0 % |
| FR-004 | T005(hecha) | — | tareas hechas | — |
| FR-005 | T005(hecha) | — | tareas hechas | — |
| FR-006 | T005(hecha) | — | tareas hechas | — |
| FR-007 | T002(hecha) T004(hecha) T005(hecha) | — | tareas hechas | — |
| FR-008 | T005(hecha) T015(hecha) | — | tareas hechas | — |
| FR-010 | T013(hecha) | — | tareas hechas | — |
| FR-011 | T013(hecha) | — | tareas hechas | — |
| FR-012 | T013(hecha) | — | tareas hechas | — |
| FR-013 | T013(hecha) | — | comprobado por su control | `ci:internal/evals/conjunto_test.go:TestEvalsDelRepositorio`, en make ci (verde) |
| FR-014 | T013(hecha) | — | tareas hechas | — |
| FR-015 | T013(hecha) | — | tareas hechas | — |
| FR-016 | T013(hecha) | — | tareas hechas | — |
| FR-017 | T013(hecha) | — | comprobado por su control | `ci:Makefile:skills-check`, en make ci (verde) |
| FR-018 | T014(hecha) | — | tareas hechas | — |
| FR-020 | T001(hecha) T002(hecha) | — | tareas hechas | — |
| FR-021 | T002(hecha) | — | comprobado por su control | `ci:internal/evals/conjunto_test.go:TestEvalsDelRepositorio`, en make ci (verde) |
| FR-022 | T002(hecha) | — | comprobado por su control | `ci:internal/evals/conjunto_test.go:TestEvalsDelRepositorio`, en make ci (verde) |
| FR-023 | T001(hecha) | — | tareas hechas | — |
| FR-024 | T001(hecha) T002(hecha) | — | tareas hechas | — |
| FR-030 | T007(hecha) T008(hecha) T009(hecha) | — | comprobado por su control | `ci:internal/evals/sesiones_test.go:TestEjecutarSesionesEnParalelo`, en make ci (verde); `ci:internal/evals/definicion_test.go:TestDefinicionDelJob`, en make ci (verde) |
| FR-031 | T006(hecha) T009(hecha) | — | tareas hechas | — |
| FR-032 | T006(hecha) T007(hecha) | — | tareas hechas | — |
| FR-033 | T003(hecha) T004(hecha) | — | tareas hechas | — |
| FR-034 | T008(hecha) | — | comprobado por su control | `ci:internal/evals/definicion_test.go:TestDefinicionDelJob`, en make ci (verde) |
| FR-035 | T008(hecha) | — | comprobado por su control | `ci:internal/evals/definicion_test.go:TestDefinicionDelJob`, en make ci (verde) |
| FR-036 | T006(hecha) T007(hecha) | — | tareas hechas | — |
| FR-037 | T007(hecha) T009(hecha) | — | tareas hechas | — |
| FR-040 | T003(hecha) T004(hecha) | — | tareas hechas | — |
| FR-041 | T003(hecha) T004(hecha) | — | tareas hechas | — |
| FR-042 | T004(hecha) | — | tareas hechas | — |
| FR-043 | T004(hecha) T009(hecha) | — | comprobado por su control | `ci:internal/evals/informe_test.go:TestInformeConSesionesSinMedir`, en make ci (verde) |
| FR-044 | T004(hecha) T007(hecha) | — | tareas hechas | — |
| FR-050 | T005(hecha) T007(hecha) T009(hecha) | — | tareas hechas | — |
| FR-051 | T005(hecha) T008(hecha) T009(hecha) T015(hecha) | — | comprobado por su control | `evals:boe-legislacion:duracion_de_las_sesiones`: 492, ≤ 900 |
| FR-052 | T008(hecha) | — | tareas hechas | — |
| FR-060 | T008(hecha) T011(hecha) T012(hecha) | — | tareas hechas | — |
| FR-061 | T006(hecha) T009(hecha) T010(hecha) T011(hecha) | — | tareas hechas | — |
| FR-062 | T006(hecha) T011(hecha) | — | tareas hechas | — |
| FR-063 | T003(hecha) T011(hecha) T012(hecha) | — | tareas hechas | — |
| FR-064 | T011(hecha) T012(hecha) | — | comprobado por su control | `ci:internal/evals/sondeo_test.go:TestSondear`, en make ci (verde); `ci:internal/evals/sondeo_test.go:TestGuionDelSondeo`, en make ci (verde); `ci:internal/evals/sondeo_integracion_test.go:TestPrepararElArbolDelSondeo`, en make ci (verde) |
| FR-065 | T003(hecha) T010(hecha) | — | tareas hechas | — |
| FR-066 | T010(hecha) T011(hecha) T012(hecha) | — | tareas hechas | — |
| FR-067 | T011(hecha) T012(hecha) | — | tareas hechas | — |
| FR-068 | T006(hecha) T012(hecha) T015(hecha) | — | tareas hechas | — |
| FR-070 | T015(hecha) | — | tareas hechas | — |
| FR-080 | T002(hecha) T015(hecha) | — | tareas hechas | — |
| FR-090 | T015(hecha) | — | tareas hechas | — |
| FR-091 | T013(hecha) | — | comprobado por su control | `ci:internal/evals/conjunto_test.go:TestEvalsDelRepositorio`, en make ci (verde) |
| FR-092 | T005(hecha) | — | tareas hechas | — |
| FR-093 | T003(hecha) T004(hecha) T007(hecha) | — | tareas hechas | — |
| FR-094 | T007(hecha) T008(hecha) | — | comprobado por su control | `ci:internal/evals/sesiones_test.go:TestEjecutarSesionesEnParalelo`, en make ci (verde) |
| FR-095 | T002(hecha) | — | comprobado por su control | `ci:internal/evals/conjunto_test.go:TestEvalsDelRepositorio`, en make ci (verde); `ci:internal/evals/conjunto_test.go:TestEvalsDelRepositorio`, en make ci (verde) |
| FR-096 | T010(hecha) T011(hecha) T012(hecha) | — | tareas hechas | — |
| FR-097 | T014(hecha) | — | tareas hechas | — |
| FR-098 | — | — | SIN TAREA | — |
| FR-099 | T015(hecha) | — | tareas hechas | — |
| SC-001 | T004(hecha) T005(hecha) T009(hecha) | — | comprobado por su control | `evals:boe-legislacion:expresiones_prohibidas:claude-sonnet-5`: 1 de 51 (2,0 %), ≤ 5,0 %; `evals:boe-legislacion:duracion_de_las_sesiones`: 492, ≤ 900; `ci:internal/evals/informe_test.go:TestInformeConSesionesSinMedir`, en make ci (verde) |
| SC-002 | — | — | SIN TAREA | — |
| SC-003 | T001(hecha) T002(hecha) | — | comprobado por su control | `ci:internal/evals/conjunto_test.go:TestEvalsDelRepositorio`, en make ci (verde) |
| SC-004 | T002(hecha) | — | comprobado por su control | `ci:internal/evals/conjunto_test.go:TestEvalsDelRepositorio`, en make ci (verde) |
| SC-005 | T013(hecha) | — | comprobado por su control | `ci:internal/evals/conjunto_test.go:TestEvalsDelRepositorio`, en make ci (verde) |
| SC-006 | T005(hecha) | — | tareas hechas | — |
| SC-007 | T003(hecha) T004(hecha) T007(hecha) | — | tareas hechas | — |
| SC-008 | T006(hecha) T007(hecha) T008(hecha) | — | comprobado por su control | `ci:internal/evals/sesiones_test.go:TestEjecutarSesionesEnParalelo`, en make ci (verde); `ci:internal/evals/definicion_test.go:TestDefinicionDelJob`, en make ci (verde) |
| SC-009 | T010(hecha) T011(hecha) T012(hecha) | — | comprobado por su control | `ci:internal/evals/sondeo_test.go:TestSondear`, en make ci (verde); `ci:internal/evals/sondeo_test.go:TestGuionDelSondeo`, en make ci (verde); `ci:internal/evals/sondeo_integracion_test.go:TestPrepararElArbolDelSondeo`, en make ci (verde) |
| SC-010 | T013(hecha) T014(hecha) T015(hecha) | — | comprobado por su control | `ci:Makefile:skills-check`, en make ci (verde) |

## 7. Cambios posteriores a la revisión final

Commits posteriores a los veredictos, que ningún juez juzgó (correcciones del cierre, registros y este informe), con lo que cada uno toca fuera de `gates/`:

- `8f85eb9` docs(H7.3): informe final del run: solo registros de `gates/`.

## 8. Cómo comprobarlo y consumo

- Escenarios manuales: `specs/013-h7-3-el-umbral-de/quickstart.md`. Suite de aceptación congelada: `specs/013-h7-3-el-umbral-de/aceptacion/` (activada en `internal/app/testdata/script/`).
- Run `f41e85d1`: 12 h 3 min de reloj. Coste por paso y por rol: `scripts/coste-run.sh f41e85d1`.
