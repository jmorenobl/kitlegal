# Informe del hito H21 · `kitlegal mcp serve`: las herramientas del binario por MCP, con las skills y las evals en los dos modos (adelantado; ADR 0035)

Generado por el workflow `hito` el 2026-10-02T05:57:17Z, sobre `c58edeb` de `015-h21-kitlegal-mcp-serve`.
Lo escribe scripts/workflow/informe.sh sin modelo, desde los artefactos de `specs/015-h21-kitlegal-mcp-serve/`. Fusionar (squash-merge) es una decisión humana:
si algo de lo que sigue no es lo que se quería, se corrige la sección del hito en docs/ROADMAP.md y se relanza.

## 1. Estado

- **make ci local**: verde.
- **CI y evals remotos** sobre `faae2e6` (medición 1): verde; es el producto de la cabeza (lo posterior solo toca `gates/`).
- **Evals por skill**: boe-legislacion aprobado, legal-core aprobado (tasas en la sección 3).
- **Umbrales del job**: boe-legislacion: 10 umbrales, **2 sin cumplir o solo publicados** (`expresiones_prohibidas:claude-haiku-4-5-20251001:orden`, `expresiones_prohibidas:claude-haiku-4-5-20251001:herramienta`); legal-core: sin umbrales (sección 3).
- **Revisión final**: juez A aprobado, juez B aprobado, 1 ronda.
- **Cambios que ningún juez vio**: ninguno.
- **Tareas**: 18 hechas, 0 en cuarentena, 0 pendientes sin cuarentena.
- **Diff**: 24 commits; 171 files changed, 25110 insertions(+), 1123 deletions(-).

## 2. Supuestos y pendientes

Decisiones que el run tomó sin preguntar, ordenadas por impacto: cada paso que escribe una etiqueta la suya (ADR 0028).

### Cambian el comportamiento visible, el alcance o una skill

**Comportamiento (salida, códigos, ficheros, argumentos)**

- corrector_spec: [i] las tasas por modo, vistas desde `scripts/workflow/informe.sh`, que agrupa `tasas` por eval y enseña la primera serie de cada modelo → FR-048 pide que el informe del job se lea con el `informe.sh` de hoy, sin cambiarlo: ninguna pareja de fila y modelo de `tasas` repetida entre modos y cada recuento de expresiones con su modo; FR-080 y SC-011 ganan el caso. Se acota a lo que ese guion ya lee y no se añade salida nueva; cómo distingue el informe los modos lo fija el plan. Queda… (entera en `specs/015-h21-kitlegal-mcp-serve/gates/supuestos.md` o `clarify-respuestas.json`)
- corrector_spec: [h][f] `mcp serve --dry-run`, que el spec daba por hecho en el kernel y el kernel no corta → FR-022 separa las banderas: `--describe` lo resuelve el kernel; `--dry-run` lo honra el applet `mcp`, que no atiende ningún mensaje, deja la salida estándar vacía y termina con la descripción del kernel en la salida de error; FR-075, SC-008 y US4 lo comprueban. Es la lectura que menos comportamiento añade: la otra servía protocolo con una bandera que pide no ejecutar.
- plan: qué versión del SDK de MCP entra → `modelcontextprotocol/go-sdk` v1.7.0, con `golang.org/x/oauth2` v0.37.0 y `github.com/segmentio/asm` v1.2.1 como indirectas: son las que la caché de módulos tiene verificadas. La v1.8.0 que leyó el ADR 0035 no se pudo descargar ni verificar: el sandbox de las sesiones del run no deja escribir en `sumdb`. Ya atiende la especificación 2026-07-28 y las anteriores; la subida queda para Dependabot (research D1, V1, V2).
- corrector_plan: [h][m] con un cliente HTTP por proceso, lo que `internal/httpx` recuerda de `robots.txt` —las reglas y la denegación por no haberlo podido obtener— viviría lo que el servidor: un fallo pasajero dejaría toda llamada a `boe_*` fuera de la caché en `limite-o-tos` hasta reiniciar, cuando la orden lee en cuanto vuelve la red → cada llamada que llega a la red construye su cliente, como la orden, y lo que obtiene de `robots.txt` vale solo para ella (H2 FR 015 sigue como está: no se le d… (entera en `specs/015-h21-kitlegal-mcp-serve/gates/supuestos.md` o `clarify-respuestas.json`)
- plan: qué hace el servidor con la cancelación que envía el cliente → no interrumpe la llamada, que termina con su plazo de `--timeout`; su resultado se descarta. El spec no pide cancelación y no se añade un contexto al kernel para ella (research D9).
- plan: el tope del trabajo de evals con tres tandas → `timeout-minutes: 240`, que cubre el peor caso (14 357 s en `boe-legislacion`), por encima de las 3 h que espera el cierre del workflow: solo si casi todas las sesiones agotaran su tope el cierre se detendría antes de leer un informe que sería `fallo`. No se sube la concurrencia, que es donde H7.3 midió la suscripción (research D20).
- T004: el ejemplo de `outputSchema` de contracts/servidor-mcp.md §2 escribe la clave `$defs` al final, y el serializador del generador de `--describe` la escribe al principio (el orden de los campos de `invopop.Schema`) → se deja el orden del generador: es el mismo documento JSON para cualquier cliente y el orden de las claves no cambia su tamaño; moverla pediría recomponer el esquema ya serializado, que es la alternativa que research D6 rechaza. El ejemplo de `inputSchema`, que no lleva `$defs`,… (entera en `specs/015-h21-kitlegal-mcp-serve/gates/supuestos.md` o `clarify-respuestas.json`)
- T004: contracts/servidor-mcp.md §3 no dice qué rechazo se da cuando una llamada incumple varias cosas, ni qué propiedad se nombra cuando sobran varias → primero que los argumentos no sean un objeto; después la propiedad de más, y de varias la primera por orden alfabético; después, campo a campo en el orden de la orden, el tipo de cada valor antes que el argumento de posición sin el que le precede. El mensaje va en el sobre: la misma llamada tiene que dar siempre el mismo, y el orden de un objeto… (entera en `specs/015-h21-kitlegal-mcp-serve/gates/supuestos.md` o `clarify-respuestas.json`)
- T004: qué es «un número entero» en los argumentos de una llamada (research D5) → todo número JSON sin parte fraccionaria, se escriba `3`, `3.0` o `3e0`: es lo que admite el `integer` del esquema de entrada que la herramienta anuncia, y lo que se exige de un valor sale de ese esquema. Llega a la línea de órdenes con todas sus cifras y nada más (`3`). Hoy ningún verbo declara un entero (research V25).
- T007: ningún contrato fija la descripción del applet `mcp` ni la de su verbo `serve`, que salen en la ayuda y en `--describe` → «Sirve las herramientas de kitlegal a un agente por el protocolo MCP.» para el applet y, para el verbo, la frase con la que contracts/servidor-mcp.md §1 dice qué hace la orden: «Atiende el protocolo MCP por la entrada y la salida estándar hasta que la entrada se cierra.». `schemas/servidor.json` (T008) saldrá con la segunda.
- T007: contracts/servidor-mcp.md §4 pide un evento por llamada con la forma `applet=… verbo=…`, y la orden registra `verbo=""` cuando su análisis falla (research V39) → el evento de una llamada lleva siempre el applet y el verbo de su herramienta, también cuando sus argumentos no llegan a analizarse (los rechaza la conversión, que no tiene orden equivalente, o el analizador): en una llamada el verbo se conoce desde el principio, y así lo registra un solo sitio. Va antes del mensaje del fallo, com… (entera en `specs/015-h21-kitlegal-mcp-serve/gates/supuestos.md` o `clarify-respuestas.json`)
- T007: contracts/servidor-mcp.md §1 dice que un final que no es el cierre de la entrada es `inesperado`, código 1, y no fija su mensaje → `mcp: el servidor ha terminado sin que se cierre su entrada: <causa que da el SDK>`. Y unas dependencias sin entrada, que es un defecto de composición como el reloj de `graph`, dan `mcp: el applet se compuso sin entrada y no puede atender ningún mensaje`, también `inesperado`, en lugar de un pánico al leer.

**Alcance (lo que queda fuera o dentro del hito)**

- plan: «editar los guiones de los hitos anteriores» está fuera de alcance, y registrar `mcp` cambia la lista de applets que fija `internal/app/testdata/script/argumentos.txtar` (H1) → cambia solo esa lista, en sus líneas 24 y 30, en la tarea `[datos]` que registra el applet, como en H6, H19 y H7; nada más de ningún guion anterior (research D26).
- plan: las dos evals sin binario ni servidor y la tarea `[aceptacion]` → no las escribe la primera tarea: su clave no existe en el formato de eval hasta la tarea que cambia `schemas/eval.yaml.json`, y antes dejarían `make ci` en rojo. Entran con esa tarea, antes que el texto de las dos `SKILL.md` (plan, «Aceptación e2e» y paso 6).
- T008: la tarea pide añadir a sus rutas cualquier otra lista de los applets del registro que aparezca fuera de ellas, y `git grep` de `applets disponibles` da tres más: el ejemplo de `CONTRIBUTING.md:85`, las entradas ya publicadas de `CHANGELOG.md` (líneas 399, 1078 y 1143) y los textos de prueba de `internal/render`, que no nombran el registro (`(ninguno)`, `boe, placsp`) → no se añade ninguna ruta: el ejemplo de `CONTRIBUTING.md` y la entrada nueva de `CHANGELOG.md` son de T017, que los declar… (entera en `specs/015-h21-kitlegal-mcp-serve/gates/supuestos.md` o `clarify-respuestas.json`)
- T017: FR-060 nombra la orden de Claude Code solo como `claude mcp add`, sin el nombre del servidor ni lo que arranca, y de Antigravity, solo el fichero `mcp_config.json` → el README da la orden entera, `claude mcp add kitlegal -- kitlegal mcp serve`, con la forma de la de Codex, que FR-060 sí fija carácter a carácter; y para Antigravity, las dos rutas del ADR 0035 (`~/.gemini/config/mcp_config.json` y `.agents/mcp_config.json`) con un bloque `mcpServers` de `command` `kitlegal` y `args` `mcp`, `… (entera en `specs/015-h21-kitlegal-mcp-serve/gates/supuestos.md` o `clarify-respuestas.json`)
- T017: «sin prometer el `.mcpb` ni el plugin, que son de H22» no dice si hay que retirar las menciones al plugin que el README ya traía —«Llegarán con un plugin», en Cowork y el chat de Claude, y «un plugin para Claude», en «Lo que viene»— → lo que se añade no nombra ni el `.mcpb` ni el plugin, y de la frase de «otras formas de usarlo» sale solo el servidor MCP. Las dos menciones anteriores se quedan como estaban: H21 no las hace falsas, y quitar la segunda dejaría sin destino el enlace de la pri… (entera en `specs/015-h21-kitlegal-mcp-serve/gates/supuestos.md` o `clarify-respuestas.json`)
- T017: la tarea enumera lo que cambia en CONTRIBUTING (el applet `mcp`, «Job de evals», «Sondeo local» y el ejemplo de la lista de applets) bajo el título «la documentación que el hito deja falsa o incompleta», y el hito deja falsos otros cuatro sitios del mismo fichero → se corrigen también, sin salir de él: la lista de esquemas de `make schema-check` gana `servidor.json`; la fila de `make skills-check` y el párrafo de «Skills y evals», las dos formas de cada operación, la columna «Herramienta» … (entera en `specs/015-h21-kitlegal-mcp-serve/gates/supuestos.md` o `clarify-respuestas.json`)
- T018: la tarea pide tests «si una cifra de cobertura queda bajo su umbral» y nombra el global y el dominio, pero Codecov mide además el diff en la propuesta de cambio (`patch`, `target: auto`), y 33 líneas añadidas de producto no las ejecuta ningún perfil → no se añade ningún test: el global (96,9 % y 97,4 %), el dominio (98,6 %) e `internal/cli` (98,5 %) cumplen su umbral, y la estimación local del diff (97,4 %, 1 221 de 1 254 líneas) queda por encima de la de `main` con el mismo criterio (95,5… (entera en `specs/015-h21-kitlegal-mcp-serve/gates/supuestos.md` o `clarify-respuestas.json`)
- clarify Q1 (criterio d, conservadora): `make evals-sondeo` sigue midiendo solo el modo orden. Sus cinco argumentos (`SKILL`, `EVALS`, `MODELO`, `REPETICIONES`, `CONCURRENCIA`), la preparación de cada sesión —el binario y la skill del árbol de trabajo en el `PATH`, ningún servidor declarado—, su juicio, su salida y sus códigos de salida no cambian para las evals de los dos modos. El modo herramienta y las dos evals sin binario ni servidor (FR-046) se miden solo en el job. Pedir en `EVALS` una eva… (entera en `specs/015-h21-kitlegal-mcp-serve/gates/supuestos.md` o `clarify-respuestas.json`)

### Del propio run


- Observaciones de los jueces, por debajo del umbral y sin corregir: 15 en `spec-r2.json`, 15 en `plan-r2.json`, 11 en `tasks-r2.json`, 10 en `revision-a-r1.json`, 12 en `revision-b-r1.json`.

### Internos (60)

Decisiones que no cambian nada observable (técnica, estructura, tests), por autor: T001 (5), T002 (1), T003 (2), T005 (2), T006 (2), T007 (3), T008 (1), T009 (7), T010 (5), T012 (1), T013 (5), T014 (5), T015 (6), T016 (7), T017 (2), T018 (3), corrector_spec (1), plan (2). Enteras en `specs/015-h21-kitlegal-mcp-serve/gates/supuestos.md`.

## 3. Evals sobre la cabeza

**boe-legislacion**: aprobado sobre `faae2e6`; decide `claude-sonnet-5-5` con 2 de 3; 21 evals, 1 nuevas, 8 informativas.

| Eval | claude-sonnet-5-5 | claude-haiku-4-5-20251001 | claude-sonnet-5-5 (herramienta) | claude-haiku-4-5-20251001 (herramienta) | Marca |
|---|---|---|---|---|---|
| `21-sin-binario-ni-servidor.yaml` | 2/3 | 3/3 | — | — | **nueva** |
| `01-lpac-articulo-21.yaml` | 3/3 | 3/3 | 3/3 | 3/3 |  |
| `02-lcsp-contrato-menor.yaml` | 3/3 | 2/3 | 3/3 | 3/3 |  |
| `03-lrbrl-atribuciones-del-pleno.yaml` | 3/3 | 2/3 | 3/3 | 2/3 |  |
| `04-lgt-prescripcion.yaml` | 2/3 | 2/3 | 3/3 | 2/3 |  |
| `05-trlrhl-impuestos-municipales.yaml` | 3/3 | 1/3 | 3/3 | 1/3 |  |
| `06-irpf-rendimientos-del-trabajo.yaml` | 3/3 | 1/3 | 3/3 | 0/3 |  |
| `07-lrjsp-principio-de-legalidad.yaml` | 3/3 | 2/3 | 3/3 | 3/3 |  |
| `08-ltaibg-plazo-de-resolucion.yaml` | 3/3 | 3/3 | 3/3 | 3/3 |  |
| `09-constitucion-articulo-140.yaml` | 3/3 | 3/3 | 3/3 | 3/3 |  |
| `10-et-vacaciones.yaml` | 3/3 | 3/3 | 3/3 | 2/3 |  |
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

Respuestas con alguna expresión prohibida, en las evals que activan la skill (`expresiones_prohibidas_por_modelo`):

| Modelo | Con alguna | Respuestas | Porcentaje |
|---|---|---|---|
| `claude-sonnet-5-5 (orden)` | 0 | 54 | 0,0 % |
| `claude-haiku-4-5-20251001 (orden)` | 0 | 30 | 0,0 % |
| `claude-sonnet-5-5 (herramienta)` | 0 | 54 | 0,0 % |
| `claude-haiku-4-5-20251001 (herramienta)` | 0 | 30 | 0,0 % |

Umbrales que publica el job (ADR 0029):

| Umbral | Medida | Condición | Cumple | Hace fallar el job |
|---|---|---|---|---|
| `expresiones_prohibidas:claude-sonnet-5-5:orden` | 0 de 54 (0,0 %) | ≤ 5,0 % | sí | sí |
| `sin_activar:claude-sonnet-5-5:orden` | 0 de 54 (0,0 %) | ≤ 0,0 % | sí | sí |
| `redaccion_no_leida:claude-sonnet-5-5:orden` | 0 de 54 (0,0 %) | ≤ 0,0 % | sí | sí |
| `expresiones_prohibidas:claude-haiku-4-5-20251001:orden` | 0 de 30 (0,0 %) | ≤ 5,0 % | sí | no: solo se publica, no es un control |
| `expresiones_prohibidas:claude-sonnet-5-5:herramienta` | 0 de 54 (0,0 %) | ≤ 5,0 % | sí | sí |
| `sin_activar:claude-sonnet-5-5:herramienta` | 0 de 54 (0,0 %) | ≤ 0,0 % | sí | sí |
| `redaccion_no_leida:claude-sonnet-5-5:herramienta` | 0 de 54 (0,0 %) | ≤ 0,0 % | sí | sí |
| `expresiones_prohibidas:claude-haiku-4-5-20251001:herramienta` | 0 de 30 (0,0 %) | ≤ 5,0 % | sí | no: solo se publica, no es un control |
| `duracion_de_las_sesiones:orden` | 471 | ≤ 900 | sí | sí |
| `duracion_de_las_sesiones:herramienta` | 479 | ≤ 900 | sí | sí |

**legal-core**: aprobado sobre `faae2e6`; decide `claude-sonnet-5-5` con 2 de 3; 4 evals, 1 nuevas, 0 informativas.

| Eval | claude-sonnet-5-5 | claude-haiku-4-5-20251001 | claude-sonnet-5-5 (herramienta) | claude-haiku-4-5-20251001 (herramienta) | Marca |
|---|---|---|---|---|---|
| `04-sin-binario-ni-servidor.yaml` | 3/3 | 3/3 | — | — | **nueva** |
| `01-territorio-municipio-cubierto.yaml` | 3/3 | 3/3 | 3/3 | 3/3 |  |
| `02-territorio-municipio-no-cubierto.yaml` | 3/3 | 3/3 | 3/3 | 3/3 |  |
| `03-no-activa-receta-de-cocina.yaml` | 3/3 | 3/3 | 3/3 | 3/3 |  |

Umbrales: el job no publica ninguno para esta skill.

Cada celda: sesiones que pasan de las abiertas; ✗, una serie que decide y no llega al umbral. Una eval informativa publica su tasa sin decidir el veredicto (ADR 0016). La regla por serie no hace cumplir un umbral agregado sobre todas las respuestas: eso solo lo hace un umbral del job que lo hace fallar (ADR 0029).

## 4. Revisión que la constitución reserva a la persona

- Fixture o esquema EXISTENTE modificado: `internal/app/testdata/script/argumentos.txtar`.
- Fixture o esquema EXISTENTE modificado: `schemas/eval.yaml.json`.
- Fixtures nuevos (23), por directorio:
  - `internal/evals/testdata/sesiones/leer-llamadas/con-prefijo/`: `codigo-de-la-sesion`, `servidor.json`, `sesion.err`, `sesion.jsonl`
  - `internal/evals/testdata/sesiones/leer-llamadas/error-por-is-error/`: `codigo-de-la-sesion`, `sesion.err`, `sesion.jsonl`
  - `internal/evals/testdata/sesiones/leer-llamadas/error-por-ok-falso/`: `codigo-de-la-sesion`, `sesion.err`, `sesion.jsonl`
  - `internal/evals/testdata/sesiones/leer-llamadas/herramienta-ajena/`: `codigo-de-la-sesion`, `sesion.err`, `sesion.jsonl`
  - `internal/evals/testdata/sesiones/leer-llamadas/modo-orden/`: `codigo-de-la-sesion`, `sesion.err`, `sesion.jsonl`
  - `internal/evals/testdata/sesiones/leer-llamadas/sin-prefijo/`: `codigo-de-la-sesion`, `servidor.json`, `sesion.err`, `sesion.jsonl`
  - `internal/evals/testdata/sesiones/leer-llamadas/sin-resultado/`: `codigo-de-la-sesion`, `sesion.err`, `sesion.jsonl`
- Esquemas nuevos: `schemas/servidor.json`.
- Suite de aceptación activada y congelada (5 guiones, escritos desde el spec en la primera tarea): `h21-mcp-errores.txtar`, `h21-mcp-herramientas.txtar`, `h21-mcp-llamadas.txtar`, `h21-mcp-proceso.txtar`, `h21-mcp-protocolo.txtar`.

## 5. Tareas en cuarentena

Ninguna.

## 6. Trazabilidad

**Estado**: «tareas hechas» solo dice que las tareas que citan el requisito están marcadas; nada del run lo ha medido. «comprobado por su control» exige su fila en «Controles de umbral» de `plan.md` y que el control esté: un test o una comprobación de `make ci` que existe en la cabeza, con `make ci` en verde, o un umbral que el job de evals publica, cumple y hace fallar el job (ADR 0029). «UMBRAL NO CUMPLIDO» y «CONTROL SIN VERIFICAR» son lo que hay que mirar.

| Requisito | Tareas | Aceptación | Estado | Control de umbral |
|---|---|---|---|---|
| FR-001 | T001(hecha) T002(hecha) T007(hecha) T008(hecha) | mcp-errores.txtar mcp-herramientas.txtar mcp-llamadas.txtar mcp-proceso.txtar mcp-protocolo.txtar  | tareas hechas | — |
| FR-002 | T001(hecha) T007(hecha) T008(hecha) | mcp-herramientas.txtar  | comprobado por su control | `ci:internal/app/herramientas_test.go:TestHerramientasDelServidor`, en make ci (verde) |
| FR-003 | T001(hecha) T004(hecha) T007(hecha) | mcp-herramientas.txtar  | tareas hechas | — |
| FR-004 | T001(hecha) T007(hecha) T008(hecha) | mcp-herramientas.txtar  | tareas hechas | — |
| FR-005 | T001(hecha) T002(hecha) T007(hecha) | mcp-herramientas.txtar  | tareas hechas | — |
| FR-006 | T001(hecha) T002(hecha) | mcp-herramientas.txtar  | comprobado por su control | `ci:internal/mcp/instrucciones_test.go:TestInstrucciones`, en make ci (verde) |
| FR-007 | T001(hecha) T002(hecha) T003(hecha) T009(hecha) | mcp-protocolo.txtar  | tareas hechas | — |
| FR-008 | T001(hecha) T002(hecha) T007(hecha) | mcp-herramientas.txtar  | tareas hechas | — |
| FR-010 | T001(hecha) T002(hecha) T003(hecha) T004(hecha) T005(hecha) T006(hecha) T007(hecha) T009(hecha) | mcp-llamadas.txtar  | tareas hechas | — |
| FR-011 | T001(hecha) T002(hecha) T003(hecha) T004(hecha) T007(hecha) T009(hecha) | mcp-errores.txtar  | tareas hechas | — |
| FR-012 | T001(hecha) T007(hecha) T009(hecha) | mcp-llamadas.txtar  | tareas hechas | — |
| FR-013 | T001(hecha) T007(hecha) T009(hecha) | mcp-llamadas.txtar  | tareas hechas | — |
| FR-014 | T005(hecha) T006(hecha) T007(hecha) | — | tareas hechas | — |
| FR-015 | T001(hecha) T007(hecha) T009(hecha) | mcp-errores.txtar  | tareas hechas | — |
| FR-020 | T001(hecha) T004(hecha) T007(hecha) T009(hecha) | mcp-errores.txtar mcp-llamadas.txtar  | tareas hechas | — |
| FR-021 | T001(hecha) T007(hecha) T009(hecha) | mcp-proceso.txtar  | tareas hechas | — |
| FR-022 | T001(hecha) T007(hecha) T008(hecha) T009(hecha) | mcp-proceso.txtar  | tareas hechas | — |
| FR-023 | T001(hecha) T003(hecha) T007(hecha) T009(hecha) | mcp-protocolo.txtar  | tareas hechas | — |
| FR-024 | T001(hecha) T002(hecha) T007(hecha) T009(hecha) | mcp-proceso.txtar  | tareas hechas | — |
| FR-025 | T001(hecha) T007(hecha) T009(hecha) | mcp-proceso.txtar  | tareas hechas | — |
| FR-026 | T002(hecha) T007(hecha) | — | tareas hechas | — |
| FR-030 | T012(hecha) | — | tareas hechas | — |
| FR-031 | T012(hecha) | — | comprobado por su control | `ci:internal/app/skills_test.go:TestOrdenesDeLasSkillsEmpotradas`, en make ci (verde) |
| FR-032 | T012(hecha) | — | tareas hechas | — |
| FR-033 | T012(hecha) | — | tareas hechas | — |
| FR-034 | T012(hecha) | — | tareas hechas | — |
| FR-035 | T012(hecha) | — | tareas hechas | — |
| FR-036 | T012(hecha) | — | comprobado por su control | `ci:Makefile:skills-check`, en make ci (verde); `ci:internal/evals/conjunto_test.go:TestEvalsDelRepositorio`, en make ci (verde) |
| FR-040 | T013(hecha) T015(hecha) T016(hecha) | — | tareas hechas | — |
| FR-041 | T013(hecha) T014(hecha) T015(hecha) | — | tareas hechas | — |
| FR-042 | T014(hecha) T015(hecha) | — | tareas hechas | — |
| FR-043 | T013(hecha) T016(hecha) T018(hecha) | — | comprobado por su control | `evals:boe-legislacion:expresiones_prohibidas:claude-sonnet-5-5:orden`: 0 de 54 (0,0 %), ≤ 5,0 %; `evals:boe-legislacion:expresiones_prohibidas:claude-sonnet-5-5:herramienta`: 0 de 54 (0,0 %), ≤ 5,0 %; `evals:boe-legislacion:sin_activar:claude-sonnet-5-5:orden`: 0 de 54 (0,0 %), ≤ 0,0 %; `evals:boe-legislacion:sin_activar:claude-sonnet-5-5:herramienta`: 0 de 54 (0,0 %), ≤ 0,0 %; `evals:boe-legislacion:redaccion_no_leida:claude-sonnet-5-5:orden`: 0 de 54 (0,0 %), ≤ 0,0 %; `evals:boe-legislacion:redaccion_no_leida:claude-sonnet-5-5:herramienta`: 0 de 54 (0,0 %), ≤ 0,0 %; `evals:boe-legislacion:duracion_de_las_sesiones:orden`: 471, ≤ 900; `evals:boe-legislacion:duracion_de_las_sesiones:herramienta`: 479, ≤ 900 |
| FR-044 | T016(hecha) | — | tareas hechas | — |
| FR-045 | T016(hecha) | — | tareas hechas | — |
| FR-046 | T010(hecha) T013(hecha) T015(hecha) | — | tareas hechas | — |
| FR-047 | T012(hecha) T015(hecha) T016(hecha) | — | tareas hechas | — |
| FR-048 | T010(hecha) T013(hecha) T016(hecha) T018(hecha) | — | tareas hechas | — |
| FR-049 | T016(hecha) T018(hecha) | — | tareas hechas | — |
| FR-050 | T011(hecha) | — | tareas hechas | — |
| FR-051 | T008(hecha) T010(hecha) T014(hecha) T018(hecha) | — | tareas hechas | — |
| FR-060 | T017(hecha) | — | tareas hechas | — |
| FR-061 | T012(hecha) T017(hecha) | — | tareas hechas | — |
| FR-070 | T007(hecha) T008(hecha) | mcp-herramientas.txtar  | comprobado por su control | `ci:internal/app/herramientas_test.go:TestHerramientasDelServidor`, en make ci (verde) |
| FR-071 | T001(hecha) T003(hecha) T009(hecha) | — | comprobado por su control | `ci:internal/app/e2e_test.go:TestEntregaDelHito`, en make ci (verde) |
| FR-072 | T001(hecha) T003(hecha) T009(hecha) | — | comprobado por su control | `ci:internal/app/e2e_test.go:TestEntregaDelHito`, en make ci (verde) |
| FR-073 | T001(hecha) T003(hecha) T009(hecha) | — | comprobado por su control | `ci:internal/app/e2e_test.go:TestEntregaDelHito`, en make ci (verde) |
| FR-074 | T007(hecha) | — | comprobado por su control | `ci:internal/app/mcp_test.go:TestLlamadasSimultaneas`, en make ci (verde) |
| FR-075 | T001(hecha) T009(hecha) | — | comprobado por su control | `ci:internal/app/e2e_test.go:TestEntregaDelHito`, en make ci (verde) |
| FR-076 | T001(hecha) T009(hecha) | — | comprobado por su control | `ci:internal/app/e2e_test.go:TestEntregaDelHito`, en make ci (verde) |
| FR-077 | T002(hecha) T012(hecha) | — | comprobado por su control | `ci:Makefile:skills-check`, en make ci (verde); `ci:internal/app/skills_test.go:TestOrdenesDeLasSkillsEmpotradas`, en make ci (verde); `ci:internal/mcp/instrucciones_test.go:TestInstrucciones`, en make ci (verde) |
| FR-078 | — | — | SIN TAREA | — |
| FR-079 | T002(hecha) T008(hecha) T018(hecha) | — | comprobado por su control | `ci:.golangci.yml`, en make ci (verde); `ci:internal/arch_test.go:TestArquitectura`, en make ci (verde) |
| FR-080 | T016(hecha) | — | comprobado por su control | `ci:internal/evals/umbrales_test.go:TestUmbralesDelInforme`, en make ci (verde); `ci:internal/evals/informe_test.go:TestInformeEnDosModos`, en make ci (verde) |
| FR-081 | T015(hecha) | — | comprobado por su control | `ci:internal/evals/juzgar_test.go:TestJuzgarLasLlamadas`, en make ci (verde); `ci:internal/evals/juzgar_test.go:TestJuzgarSinBinarioNiServidor`, en make ci (verde) |
| FR-082 | T016(hecha) T018(hecha) | — | tareas hechas | — |
| FR-083 | T010(hecha) T013(hecha) | — | comprobado por su control | `ci:internal/evals/definicion_test.go:TestDefinicionDelJob`, en make ci (verde) |
| FR-084 | T011(hecha) | — | comprobado por su control | `ci:internal/evals/sondeo_test.go:TestComprobarElSondeo`, en make ci (verde); `ci:internal/evals/sondeo_test.go:TestSondear`, en make ci (verde); `ci:internal/evals/sondeo_test.go:TestGuionDelSondeo`, en make ci (verde) |
| FR-090 | T018(hecha) | — | tareas hechas | — |
| SC-001 | T010(hecha) T016(hecha) | — | comprobado por su control | `evals:boe-legislacion:expresiones_prohibidas:claude-sonnet-5-5:orden`: 0 de 54 (0,0 %), ≤ 5,0 %; `evals:boe-legislacion:expresiones_prohibidas:claude-sonnet-5-5:herramienta`: 0 de 54 (0,0 %), ≤ 5,0 %; `evals:boe-legislacion:sin_activar:claude-sonnet-5-5:orden`: 0 de 54 (0,0 %), ≤ 0,0 %; `evals:boe-legislacion:sin_activar:claude-sonnet-5-5:herramienta`: 0 de 54 (0,0 %), ≤ 0,0 %; `evals:boe-legislacion:redaccion_no_leida:claude-sonnet-5-5:orden`: 0 de 54 (0,0 %), ≤ 0,0 %; `evals:boe-legislacion:redaccion_no_leida:claude-sonnet-5-5:herramienta`: 0 de 54 (0,0 %), ≤ 0,0 %; `evals:boe-legislacion:duracion_de_las_sesiones:orden`: 471, ≤ 900; `evals:boe-legislacion:duracion_de_las_sesiones:herramienta`: 479, ≤ 900 |
| SC-002 | — | — | SIN TAREA | — |
| SC-003 | T001(hecha) T004(hecha) T007(hecha) T008(hecha) | mcp-herramientas.txtar  | comprobado por su control | `ci:internal/app/herramientas_test.go:TestHerramientasDelServidor`, en make ci (verde) |
| SC-004 | T001(hecha) T009(hecha) | mcp-errores.txtar mcp-llamadas.txtar  | comprobado por su control | `ci:internal/app/e2e_test.go:TestEntregaDelHito`, en make ci (verde) |
| SC-005 | T001(hecha) T003(hecha) T009(hecha) | mcp-protocolo.txtar  | comprobado por su control | `ci:internal/app/e2e_test.go:TestEntregaDelHito`, en make ci (verde) |
| SC-006 | T001(hecha) T003(hecha) T009(hecha) | mcp-protocolo.txtar  | comprobado por su control | `ci:internal/app/e2e_test.go:TestEntregaDelHito`, en make ci (verde) |
| SC-007 | T005(hecha) T007(hecha) | — | comprobado por su control | `ci:internal/app/mcp_test.go:TestLlamadasSimultaneas`, en make ci (verde) |
| SC-008 | T001(hecha) T009(hecha) | mcp-proceso.txtar  | comprobado por su control | `ci:internal/app/e2e_test.go:TestEntregaDelHito`, en make ci (verde) |
| SC-009 | T012(hecha) | — | comprobado por su control | `ci:Makefile:skills-check`, en make ci (verde); `ci:internal/app/skills_test.go:TestOrdenesDeLasSkillsEmpotradas`, en make ci (verde); `ci:internal/evals/conjunto_test.go:TestEvalsDelRepositorio`, en make ci (verde) |
| SC-010 | T002(hecha) | — | comprobado por su control | `ci:internal/mcp/instrucciones_test.go:TestInstrucciones`, en make ci (verde) |
| SC-011 | T016(hecha) | — | comprobado por su control | `ci:internal/evals/umbrales_test.go:TestUmbralesDelInforme`, en make ci (verde); `ci:internal/evals/informe_test.go:TestInformeEnDosModos`, en make ci (verde) |
| SC-012 | T015(hecha) | — | comprobado por su control | `ci:internal/evals/juzgar_test.go:TestJuzgarLasLlamadas`, en make ci (verde); `ci:internal/evals/juzgar_test.go:TestJuzgarSinBinarioNiServidor`, en make ci (verde) |
| SC-013 | T002(hecha) T017(hecha) T018(hecha) | — | comprobado por su control | `ci:.golangci.yml`, en make ci (verde); `ci:internal/arch_test.go:TestArquitectura`, en make ci (verde) |

## 7. Cambios posteriores a la revisión final

Cada commit posterior a lo que juzgó la primera ronda de la revisión final (`be4a12d`), con la ronda que lo vio y su veredicto, y lo que toca fuera de `gates/` (ADR 0030):

- `faae2e6` docs(H21): veredictos de la revisión final: solo registros de `gates/`.
- `c58edeb` docs(H21): informe final del run: solo registros de `gates/`.

### Cambios que ningún juez vio

Ninguno.

## 8. Cómo comprobarlo y consumo

- Escenarios manuales: `specs/015-h21-kitlegal-mcp-serve/quickstart.md`. Suite de aceptación congelada: `specs/015-h21-kitlegal-mcp-serve/aceptacion/` (activada en `internal/app/testdata/script/`).
- Run `af13b17c`: 14 h 16 min de reloj. Coste por paso y por rol: `scripts/coste-run.sh af13b17c`.
