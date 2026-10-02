# Specification Analysis Report — H21 (`015-h21-kitlegal-mcp-serve`)

Análisis de solo lectura de `spec.md`, `plan.md` y `tasks.md` contra la constitución 2.10.0 (modo autónomo, sesión de un paso, ADR 0032).

**Qué se ejecutó y qué se leyó en esta sesión.** Ejecutado: `.specify/scripts/bash/check-prerequisites.sh --json --require-spec --require-tasks --include-tasks` (devolvió `research.md`, `data-model.md`, `contracts/`, `quickstart.md` y `tasks.md`) y búsquedas de solo lectura (`grep`, `ls`, `wc`, `git log`) sobre el repositorio. Leídos enteros: `spec.md`, `plan.md`, `tasks.md` y la constitución. Leídos en parte, por búsqueda: `research.md` (V21, V33, V36, V41 a V45, S4 a S11, D1, D17 a D20), `contracts/servidor-mcp.md` (solo las líneas que nombran `timeout` y plazo), `contracts/evals-en-dos-modos.md` (solo las líneas que nombran el registro), `gates/supuestos.md`, `checklists/requirements.md`, y del código `scripts/workflow/guardian-diff.sh`, `scripts/workflow/informe.sh` (por búsqueda), `scripts/evals-sesion.sh`, `internal/evals/conjunto.go` (reglas de tamaño), `internal/skills/sincronia_test.go` (el límite de líneas), `internal/app/testdata/script/ayuda.txtar` y la lista de `internal/evals/testdata/`. **No leídos:** `data-model.md`, `quickstart.md`, `contracts/skills.md`, `contracts/arnes-e2e.md`, `contracts/skills-prototipo.diff` y el resto de `research.md` y de los contratos. **No ejecutado:** `make ci`, ningún test, ningún guion, ningún mutante. Toda afirmación sobre el comportamiento del SDK de MCP, de Claude Code o de lo que da el job de evals es la de `research.md` (tabla V y supuestos S) y no se ha vuelto a medir aquí.

## Hallazgos

| ID | Categoría | Severidad | Ubicación | Resumen | Recomendación |
|----|-----------|-----------|-----------|---------|---------------|
| E1 | Cobertura | MEDIUM | spec.md FR-020 (l. 249) y «Edge Cases» (l. 212); contracts/servidor-mcp.md l. 14, 105 y 119; tasks.md T001 (`mcp-proceso`), T007 | `--timeout` en `mcp serve` no tiene ningún control: el spec dice que es el plazo de cada llamada «y nunca el de la vida del servidor», y que una llamada que lo pasa da un error de herramienta de clase `fuente-no-disponible`, pero FR-020 solo declara comprobables `--offline` y `--no-graph` (FR-071) y la ausencia de banderas en los esquemas (FR-070). Ni los cinco guiones de T001 ni los tests de T007 (`TestHerramientasDelServidor`, `TestLlamadasSimultaneas`, `TestServirSinEntrada`, `TestServirEnEnsayo`) ejercen el plazo. La vía es real: lo pasa quien declara `kitlegal mcp serve --timeout=…` en su agente, y un servidor al que el kernel le diera plazo de vida se pararía a los segundos, con un resultado observable (el agente pierde sus herramientas). | Antes de escribir T001, añadir a `mcp-proceso` un caso con `--timeout` corto en el que el servidor sigue respondiendo a una llamada hecha después de vencer ese plazo; y a T007 un test con un verbo local que bloquee hasta su plazo, cuya llamada devuelva el sobre de fallo de la clase de la orden y no pare las siguientes. Citar FR-020 y el borde del plazo en sus líneas de requisitos. |
| I1 | Inconsistencia | LOW | spec.md FR-006 y SC-010 («512 caracteres»); plan.md Constraints y tasks.md T002 («512 bytes»); research V37 (485 bytes) | El spec mide las `instructions` en caracteres y el plan, el test y las tareas en bytes. Con texto español con tildes, 512 bytes es más estricto que 512 caracteres, así que el control no deja pasar nada que el spec rechace; es solo deriva de vocabulario. | Dejar «bytes» en el control (es lo que mide el test) y decirlo en el plan, o pedir «caracteres» en el test. No bloquea. |
| I2 | Inconsistencia | LOW | plan.md Constraints («el tope efectivo de `boe-legislacion` es 298»); spec.md FR-036 («menos de 300») y Assumptions («298 líneas hoy») | El plan habla de un tope efectivo de 298; el repositorio da 299 como máximo (`internal/skills/sincronia_test.go:502`: «tiene 300 líneas (máximo 299)»), y T012 usa el mutante de 297 + 3 = 300. No he encontrado de dónde sale el 298 (no he leído el contrato de las skills). | Quitar la frase o decir de dónde sale. Con 297 líneas previstas no afecta a ninguna tarea. |
| I3 | Inconsistencia | LOW | plan.md «Obligaciones para `tasks.md`» («las tareas `[datos]` son las de los pasos 5, 6, 8 y 9»); tasks.md etiqueta `[datos]` solo en T008, T010 y T014 | El plan reparte la etiqueta entre cuatro pasos y las tareas la ponen en tres: T013 (paso 8) y T015 (paso 9) no la llevan. tasks.md lo explica en «Dependencias y orden», precisiones 4 y 5. En el repositorio, `internal/evals/testdata/` no tiene ningún fichero de resultado esperado ni `schemas/` ningún esquema del informe, así que T013, T015 y T016 no tocan nada que el guardián proteja (`guardian-diff.sh:75`). | Alinear la frase del plan con tasks.md. No bloquea. |
| C1 | Constitución | LOW | constitución, principio II («Todo applet emite el sobre»); spec.md Assumptions («`mcp serve` no imprime un sobre propio»); plan.md, fila del principio II | `mcp serve` no escribe un sobre propio: lo escribe cada llamada, dentro del protocolo. El spec y el plan registran esa lectura y la cumplen en cada llamada (FR-010, FR-011); la salida estándar tiene que ser solo protocolo (FR-023), así que otra lectura contradiría el hito. No es una violación; lo dejo anotado para que el juez de la revisión final la vea juzgada. | Ninguna. |
| O1 | Proporcionalidad | LOW | tasks.md T007 | T007 junta el cambio de `ejecutarVerbo`, `emitir`, la llamada como invocación, `AppletMCP`, `herramientasDe`, `modulosDelBinario` y cinco tests, con unos 25 requisitos. Está justificada como rebanada vertical y no hay otra forma de dejar `make ci` en verde, pero es la tarea con más riesgo de entrar en cuarentena. | Ninguna acción ahora. Si entra en cuarentena, el corte natural es el de la propia tarea: el cuerpo de `herramientasDe` y la llamada, y después `AppletMCP` y `ejecutarVerbo`. |
| G1 | Cobertura | LOW | plan.md «Puntos de entrada fuera de `make ci`»; tasks.md T013 y T016 (`TestEjecucionDelJob`) | El código que lleva las duraciones de cada tanda al informe vive en `job_test.go`, con la etiqueta `evals`: `make ci` lo compila con `go vet` pero no lo ejecuta, y ninguna tarea lo ejecuta. El control de `duracion_de_las_sesiones:<modo>` queda ejercido con duraciones sintéticas en `TestUmbralesDelInforme` (901 s en un modo, 900 s en el otro). Es el mismo reparto de H7.3 y H7.4. | Ninguna. SC-001 solo se mide en el job del cierre; lo digo para que el informe final no lo presente como medido por una tarea. |

Total: 7 hallazgos (0 CRITICAL, 0 HIGH, 1 MEDIUM, 6 LOW). Ninguno llega al tope de 50.

## Cobertura de requisitos

Fuente: la tabla «Trazabilidad: requisitos → tareas» de tasks.md, contrastada con el texto de cada tarea.

| Requisito | ¿Tarea? | Tareas | Notas |
|-----------|---------|--------|-------|
| FR-001 | Sí | T001, T002, T007, T008 | |
| FR-002, FR-004 | Sí | T001, T007, T008 | |
| FR-003 | Sí | T004, T007 | |
| FR-005, FR-008 | Sí | T001, T002, T007 | |
| FR-006 | Sí | T001, T002 | `TestInstrucciones`; ver I1 |
| FR-007 | Sí | T001, T002, T003, T009 | |
| FR-010 | Sí | T001, T002, T004, T005, T006, T007, T009 | |
| FR-011 | Sí | T001, T002, T004, T007, T009 | |
| FR-012, FR-013, FR-015 | Sí | T001, T007, T009 | |
| FR-014 | Sí | T005, T006, T007 | `TestRitmoCompartido`, `TestLlamadasSimultaneas` |
| FR-020 | Parcial | T001, T004, T007, T009 | `--offline`, `--no-graph` y ausencia de banderas en los esquemas, sí; `--timeout`, no (E1) |
| FR-021 a FR-025 | Sí | T001, T007, T009 (y T002, T008) | |
| FR-026 | Sí | T002, T007 | |
| FR-030 a FR-034, FR-036 | Sí | T012 | |
| FR-035 | Sí | T012, T015 | |
| FR-040, FR-041, FR-042 | Sí | T013, T014, T015, T016 | |
| FR-043, FR-044, FR-045 | Sí | T016 (y T013, la duración por tanda) | |
| FR-046, FR-047 | Sí | T010, T012, T013, T015, T016 | |
| FR-048 | Sí | T010, T013, T016, T018 | |
| FR-049 | Sí | T016, T018 y la batería | |
| FR-050, FR-084 | Sí | T011 | |
| FR-051 | Sí | T008, T010, T014 | |
| FR-060, FR-061 | Sí | T012 (tablas), T017 | |
| FR-070 | Sí | T007, T008 | |
| FR-071 a FR-076 | Sí | T001, T003, T007, T009 | |
| FR-077 | Sí | T002, T012 | |
| FR-078 | No (por diseño) | el cierre del workflow | no es una tarea; T013 a T016 dejan el job listo |
| FR-079 | Sí | T002, T008, T018 | |
| FR-080, FR-081 | Sí | T016, T015 | |
| FR-082 | Sí | plan.md, T016, T018 | las ocho filas de «Controles de umbral» existen en plan.md |
| FR-083 | Sí | T010, T013 | |
| FR-090 | Sí | T018 y la batería | |
| SC-001 | Sí (control) | T010, T013, T015, T016; se mide en el job | |
| SC-002 | No (por diseño) | — | la mide una persona tras fusionar |
| SC-003 a SC-013 | Sí | ver tasks.md | |

**Controles de umbral (ADR 0029).** Las ocho filas que pide FR-082 están en plan.md, cada una con su tarea y su test que la ve fallar en tasks.md («Controles de umbral → tareas»), más las siete de `make ci` (herramientas, líneas, nombres, instrucciones, importaciones, tope del trabajo, sondeo). Los nombres de umbral cumplen `^[a-z0-9_.:-]+$` (ADR 0029, línea 145). `scripts/workflow/informe.sh:264` los resuelve por `.nombre`, así que `evals:boe-legislacion:expresiones_prohibidas:claude-sonnet-5-5:orden` casa con el umbral `expresiones_prohibidas:claude-sonnet-5-5:orden`. Ninguna tarea rebaja un umbral (FR-049 y la batería).

## Problemas de alineación con la constitución

Ninguno CRITICAL. Revisado: I (fuentes y frontera humana: sin red nueva, solo lectura, un cliente HTTP por llamada con ritmo compartido); II (ver C1); III (la primera tarea escribe la suite, `[datos]` para `schemas/` y `testdata/`); IV (R1 a R7 con la regla nueva R7; `net/http` enlazado por el SDK declarado en Complexity Tracking); V (una dependencia directa de la lista del principio y seis indirectas justificadas); VI; VII (`--asunto` es un error); VIII (las dos skills con la tabla en dos formas); IX (`territorio_resolver` con Leganés y Tordesillas). Reglas del modo desatendido: ninguna tarea usa la red salvo `go get` en T002, graba, abre sesiones con modelo ni ejecuta `make evals`, `make evals-sondeo`, `TestEjecucionDelJob` o `TestSondeo`. Guardián de diff: `.golangci.yml` (T002) y `.github/workflows/evals.yml` (T010) van declarados en la línea de su tarea, y el modo global que los rechaza solo se aplica a reparaciones y correctores (`guardian-diff.sh:76-82`); H7.3 y H7.4 los cambiaron por el mismo camino.

## Tareas sin requisito

Ninguna. T018 (cierre) se mapea a FR-048, FR-049, FR-051, FR-079, FR-082 y FR-090.

## Comprobaciones hechas contra el repositorio y que no dan hallazgo

- Las cuatro listas literales de applets de V41 están en T008. `internal/app/testdata/script/ayuda.txtar` no necesita cambio: el ancho de columna lo fija `territorio` y `mcp` es más corto.
- `internal/evals` ya importa `internal/app` (`preparar.go`, `trazas.go`): T014 y T015 no abren una dependencia nueva.
- `internal/evals/conjunto.go`: el mínimo de `legal-core` es 3, así que 4 evals no lo rompen; el máximo de `boe-legislacion` es 20 y T010 lo sube.
- Las cifras del peor caso cuadran: 485 + (25 + 24 + 2) × 272 = 14 357 s (≤ 240 min = 14 400 s) y 485 + (19 + 18 + 6) × 272 = 12 181 s.
- Las sesiones: 96 + 96 + 6 = 198 (199 con la prueba de red); 66 series de `tasas` = 2 × (20 + 12) + 2.
- Los nombres de test que las tareas dan por existentes (`TestPlan`, `TestAbrirUnaSesion`, `TestEjecutarSesiones…`, `TestDefinicionDelJob`, `TestEjecucionDelJob`, `TestSondeo`) existen en `internal/evals/`.
- Checklist `checklists/requirements.md`: 16 de 16 marcados. Sin marcadores pendientes (`TODO`, `TKTK`, `NEEDS CLARIFICATION`) en spec, plan, tasks, research, data-model, quickstart ni contratos.

## Supuestos del plan que ninguna tarea puede medir

Lo dice el propio plan (research S4 a S11) y aquí no se ha medido nada: que `claude -p --mcp-config` deje llamar a las herramientas con `bypassPermissions` (S4), la forma del `tool_use` MCP en el transcript (S5), que el cliente de Claude Code acepte el esquema de salida (S6), que el servidor no reciba el entorno entero (S7), el `PATH` del shell del runner (S8), la duración del modo herramienta (S9, no medido), el límite de `timeout-minutes: 240` (S10) y lo que el ADR 0035 dice de cada agente (S11). Los mide el job del cierre o la prueba humana de SC-002.

## Métricas

- Requisitos: 58 funcionales (FR-001 a FR-090) y 13 criterios de éxito: 71.
- Tareas: 18 (T001 a T018).
- Cobertura por al menos una tarea: 57 de 58 requisitos funcionales (98 %; el que falta es FR-078, que es del cierre del workflow por diseño) y 11 de 13 criterios de éxito con tarea o control propio de `make ci`; SC-001 se mide en el job y SC-002 la mide una persona. FR-020 está cubierto solo en parte (E1).
- Ambigüedades: 1 (I1). Duplicados: 0. Problemas CRITICAL: 0.

## Siguientes pasos

- Sin CRITICAL ni HIGH: se puede pasar a `/speckit-implement`.
- E1 (MEDIUM) es el único que cambia lo que se comprueba: conviene incorporarlo a T001 (`mcp-proceso`) y a los tests de T007 antes de que T001 congele la suite. Los demás son de redacción y no cambian ninguna tarea.
- Este informe no modifica `spec.md`, `plan.md` ni `tasks.md`. Los dos hooks de `.specify/extensions.yml` (`speckit.git.commit` en `before_analyze` y `after_analyze`) son opcionales y no se ejecutaron: la sesión es de un paso y no commitea.
