# Feature Specification: H21 · `kitlegal mcp serve`: las herramientas del binario por MCP, con las skills y las evals en los dos modos

**Feature Branch**: `015-h21-kitlegal-mcp-serve`

**Created**: 2026-10-01

**Status**: Draft

**Input**: Sección «#### H21 · `kitlegal mcp serve`: las herramientas del binario por MCP, con las skills y las evals en los dos modos (adelantado; ADR 0035)» de `docs/ROADMAP.md`, en modo desatendido (ADR 0018).

## Resumen

Hoy `boe-legislacion` y `legal-core` solo funcionan donde el agente puede ejecutar `kitlegal` en un shell del equipo. Sus dos `SKILL.md` dicen «ejecuta `kitlegal boe articulo …`», encadenan órdenes con `&&` y leen códigos de salida, y las evals solo miden órdenes. Eso deja fuera a quien más las necesita (`docs/USO.md`, entradas del 2026-09-27 «Un botón de instalar», del 2026-10-01 «ninguna de sus dos audiencias usa una terminal» y del 2026-10-01 «Usar kitlegal desde Claude Cowork»):

- **Un agente que no puede ejecutar el binario.** Claude Cowork corre su shell en una máquina virtual donde el binario del equipo no está.
- **Un agente que lo ejecuta con trabas.** El sandbox de Codex le corta la red a la orden.
- **Un agente que tiene la skill y nada más.** Hoy la skill dice «falta instalar kitlegal», sin forma fija y sin que ninguna eval compruebe que no responde de memoria.

H21 da a las dos skills una segunda forma de pedir cada operación: una herramienta de un servidor MCP local, `kitlegal mcp serve`, que corre en el equipo de la persona y habla por la entrada y la salida estándar. La herramienta devuelve el mismo sobre que la orden, así que la respuesta lleva la misma cita, los mismos avisos y las mismas reglas. Las skills usan la herramienta si el agente la tiene y la orden si no. Si no tienen ninguna de las dos, lo dicen con una línea de forma fija y no afirman nada de memoria. El job de evals mide las dos skills en los dos modos, con los umbrales de hoy en cada uno.

H21 no entrega skill nueva: protege las dos que existen haciéndolas usables donde está quien las necesita (constitución, principio VIII). Es la mitad de dentro: la instalación sin terminal es H22 y necesita este servidor (ADR 0035).

**Cómo se citan los requisitos de otros hitos.** Los de H5 a H7.4 se citan sin guion: «H7.4 FR 040», «H7.1 FR 013». Los ids con guion (FR-001, SC-001…) son siempre de este spec.

**Los dos modos.** En este spec, **modo orden** es el de hoy: la sesión tiene `kitlegal` en el `PATH` de su shell y ningún servidor declarado. **Modo herramienta** es el nuevo: el binario está fuera del `PATH` del shell de la sesión y el servidor está declarado al agente. `<modelo>` es el id del modelo que decide (ADR 0031), hoy `claude-sonnet-5-5`.

## Clarifications

### Session 2026-10-01

- Q: ¿Qué mide `make evals-sondeo` tras este hito: solo el modo orden, como hoy, o también el modo herramienta y las dos evals sin binario ni servidor, y qué hace cuando se le pide una de esas dos evals? → A: Opción A. Sigue midiendo solo el modo orden, con sus cinco argumentos, su preparación, su juicio, su salida y sus códigos sin cambios; el modo herramienta y las dos evals de FR-046 se miden solo en el job. Pedir en `EVALS` una eval sin binario ni servidor es un error de uso: una línea en la salida de error que nombra `EVALS` y la eval, salida estándar vacía, ninguna sesión abierta ni árbol preparado y salida 1, también si la lista lleva evals que sí se miden. CONTRIBUTING lo dice; los tests del sondeo de `make ci` lo comprueban. Sondear el modo herramienta, esas dos evals y cualquier argumento nuevo van a «Fuera de alcance». (auto: conservadora; criterio d; fuente: docs/ROADMAP.md, H21, Alcance, «Evals en dos modos» y Controles, que no nombran el sondeo; ADR 0035, Consecuencias; constitución, «Criterio de decisión autónoma», puntos 2 y 4; H7.3, «Sondeo local para personas» y ADR 0032; specs/013-h7-3-el-umbral-de/contracts/sondeo.md §3.1 y §6; specs/014-h7-4-boe-legislacion-sin/contracts/sondeo.md §1 a §3; `internal/evals/sondeo.go`, leído y no ejecutado en esta sesión)

## Criterios del hito, literales

Transcripción literal de `docs/ROADMAP.md` §4, H21. El Objetivo y el Alcance se reflejan en los requisitos.

- **Entrega**: «`kitlegal mcp serve` atiende el protocolo MCP por la entrada y la salida estándar y ofrece una herramienta por cada verbo de consulta del registro —hoy `boe_buscar`, `boe_indice`, `boe_articulo`, `boe_articulos`, `boe_metadatos`, `boe_analisis`, `territorio_resolver`, `graph_check`, `graph_show` y `graph_stats`—, con los esquemas de entrada y de salida que da `--describe` y, como resultado, el mismo sobre que la orden con `--json`. Con el binario instalado, quien usa la app de escritorio de ChatGPT (Codex), Antigravity o Claude Code declara el servidor con una línea de configuración y pregunta. Las dos skills nombran cada operación de las dos formas: usan la herramienta si el agente la tiene y la orden si no; y si no tienen ninguna de las dos, lo dicen y no responden de memoria.»
- **Controles**:
  - «conformidad: el conjunto de herramientas es exactamente el de los verbos del registro menos los excluidos, y los dos esquemas de cada una son los de `--describe` de su verbo;»
  - «e2e con un cliente MCP real contra el binario por stdio, con replay y `HOME` temporal: para cada herramienta, el sobre es el de su orden con `--json` para la misma entrada, salvo `fecha_consulta`; un bloque inexistente, una fuente no disponible (`--offline` con la caché vacía) y unos argumentos inválidos devuelven el sobre de su clase como error de herramienta; `graph_check` con hallazgos no es un error; tras `boe_articulo`, `graph_show` ve el bloque, y con `mcp serve --no-graph` no;»
  - «con un manifiesto de skills de otra versión y `--verbose`, cada línea de stdout es un mensaje del protocolo y el aviso está en stderr;»
  - «el mismo binario responde a un cliente de la especificación 2026-07-28 y a uno de la anterior;»
  - «llamadas simultáneas a la misma herramienta y a herramientas distintas, con `-race`: cada una recibe su resultado;»
  - «al cerrarse la entrada estándar, el servidor termina con 0;»
  - «lanzado con `/` como directorio de trabajo y desde una ruta con espacios, responde igual;»
  - «`skills-check` sin drift; toda herramienta que nombra la tabla de una skill existe en el servidor y toda orden, en el registro; `SKILL.md` < 300 líneas;»
  - «el job de evals en los dos modos, con sus umbrales por modo, y las dos evals sin binario ni servidor (Definition of Done §1.10);»
  - «`depguard`: el SDK de MCP solo se importa desde el paquete del servidor; `make ci`.»
- **Aceptación**: «en el run, el informe del job de evals da cada umbral cumplido en el modo orden y en el modo herramienta para las dos skills, y las dos evals sin binario ni servidor pasan. Después de fusionar, y fuera del run porque es humano: en la app de escritorio de ChatGPT, con `kitlegal` instalado, las skills en `~/.agents/skills/` y el servidor añadido en *Settings > MCP servers*, «¿qué dice el art. 21 de la Ley 39/2015?» se responde con `art. 21 de la Ley 39/2015 [BOE-A-2015-10565, bloque a21]` sin que el agente ejecute ninguna orden; y la misma pregunta en la app de escritorio de Claude, sin carpeta, con ese binario dentro de un `.mcpb` hecho a mano como el de la prueba del ADR 0035 e instalado con doble clic.»

Trazabilidad resumida (el detalle, en cada requisito):

| Criterio del hito | Dónde se cumple |
|---|---|
| Applet `mcp`, verbo `serve`; las herramientas, del registro | FR-001 a FR-008, US4, US6 |
| Una llamada es una invocación del kernel | FR-010 a FR-015, US1 |
| Banderas; la salida estándar es solo protocolo; dónde arranca | FR-020 a FR-025, US4 |
| Solo lectura; `instructions`; protocolo | FR-005, FR-006, FR-007 |
| Dependencia | FR-026 |
| Skills: las dos formas y la regla nueva | FR-030 a FR-036, US1, US2, US3 |
| Evals en dos modos | FR-040 a FR-051, US5 |
| Documentación | FR-060, FR-061, US7 |
| Controles | FR-070 a FR-083, SC-003 a SC-013 |
| Aceptación | SC-001 (en el run), SC-002 (humana, tras fusionar) |

## Relación con H5, H6, H7.4 y H19

No se edita ningún spec anterior: son el registro de sus runs. No cambia ninguna decisión de arquitectura que no haya cambiado ya el ADR 0035, así que no hay ADR nuevo (FR-090).

**Sustituye** (lo que dice el hito citado deja de valer, y vale lo de este spec):

- La última frase de la regla 2 de `boe-legislacion` («Si `kitlegal` no está en el `PATH`, di que falta instalar kitlegal») y la última viñeta del paso 2 de `legal-core` («Si `kitlegal` no está en el `PATH`, di que falta instalar kitlegal y no suplas los datos»): las sustituye la regla de FR-035, con su línea de forma fija.
- Los umbrales de H7.4 (H7.4 FR 040 a FR 043), en lo que dicen de que hay uno por skill: ahora hay uno por modo, con el mismo valor (FR-043). El total de cada uno no cambia: 54 respuestas del modelo que decide por modo.
- H7.4 FR 054, en el tamaño del conjunto de `boe-legislacion` (20 evals): pasa a 21 con la eval sin binario ni servidor (FR-046). `legal-core` pasa de 3 a 4.
- La tabla de comandos generada de H5 (una columna «Orden»): cada fila nombra ahora la orden y la herramienta (FR-030; constitución 2.10.0, principio VIII).

**Se queda**: el protocolo de las dos skills, sus reglas, la forma de la cita, de los avisos de vigencia, de `⚠ REDACCIÓN MODIFICADA:` y de la frase de la regla 7 (H7.4 FR 023 y FR 025); las órdenes para PowerShell (H7.4 FR 024); la lista de expresiones prohibidas, su calibrado y las comprobaciones sobre la prosa de `SKILL.md` (H7.4 FR 021, FR 032 y FR 033); el juicio sobre la respuesta a la pregunta (H7.4 FR 060); una tanda por commit y skill (H7.4 FR 070); las repeticiones y la regla por serie (ADR 0016); el contrato de `umbrales` (ADR 0029); la salida de `graph check`, acotada (H7.1); el aviso de versión distinta (H19); y los verbos de hoy, sus argumentos, sus sobres y sus códigos de salida, que no cambian ni un byte.

## User Scenarios & Testing *(mandatory)*

### User Story 1 - Quien pregunta desde un agente que no ejecuta órdenes recibe la misma respuesta, con su cita (Priority: P1)

Una persona tiene `kitlegal` instalado, las skills en `~/.agents/skills/` y el servidor declarado en su agente (la app de escritorio de ChatGPT, Antigravity o Claude Code). Pregunta «¿qué dice el art. 21 de la Ley 39/2015?». La skill pide el bloque con la herramienta `boe_articulo` y, si no falló, comprueba la redacción con `graph_check`. La respuesta lleva `art. 21 de la Ley 39/2015 [BOE-A-2015-10565, bloque a21]`, sus avisos de vigencia y, si la redacción cambió, su línea `⚠ REDACCIÓN MODIFICADA:`, igual que con la orden. El agente no ejecuta ninguna orden.

**Why this priority**: es el objetivo del hito. Sin la herramienta, las dos skills no existen para quien usa Cowork ni para quien usa Codex con su sandbox.

**Independent Test**: el e2e con un cliente MCP contra el binario (SC-004) y la conformidad (SC-003), en `make ci`; las evals de las dos skills en el modo herramienta, en el job de cierre (SC-001); y, tras fusionar, la prueba humana (SC-002).

**Acceptance Scenarios**:

1. **Dado** el servidor arrancado con las respuestas grabadas del BOE y un `HOME` temporal, **Cuando** un cliente MCP llama a `boe_articulo` con `norma` `BOE-A-2015-10565` y `bloque` `a21`, **Entonces** el resultado no es un error y lleva, en `structuredContent` y serializado en un bloque de texto, el mismo sobre que `kitlegal boe articulo BOE-A-2015-10565 a21 --json`, salvo `fecha_consulta` (FR-010).
2. **Dado** el mismo servidor, **Cuando** el cliente llama a `boe_articulo` con un bloque que no existe en la norma, **Entonces** el resultado está marcado como error de herramienta y lleva el sobre de fallo con `clase` `no-encontrado` (FR-011).
3. **Dado** el servidor arrancado con `--offline` y la caché vacía, **Cuando** el cliente llama a `boe_articulo`, **Entonces** el resultado es un error de herramienta con el sobre de `clase` `fuente-no-disponible` (FR-011, FR-020).
4. **Dado** el servidor, **Cuando** el cliente llama a `boe_articulo` sin `bloque`, o con una propiedad que su esquema no declara, **Entonces** el resultado es un error de herramienta con el sobre de `clase` `argumentos` (FR-011).
5. **Dado** un grafo con una lectura anterior de un bloque que vio otra redacción, **Cuando** el cliente llama a `boe_articulo` y, recibido su resultado, a `graph_check` con esa norma y ese bloque, **Entonces** `graph_check` devuelve un sobre con `ok` verdadero y un hallazgo `version-obsoleta`, y el resultado no está marcado como error (FR-012, FR-013).
6. **Dado** el servidor sin `--no-graph`, **Cuando** el cliente llama a `boe_articulo` y, recibido su resultado, a `graph_show` con el id del bloque, **Entonces** `graph_show` devuelve el nodo; **Dado** el servidor arrancado con `mcp serve --no-graph` y un grafo vacío, **Cuando** hace lo mismo, **Entonces** `graph_show` devuelve un error de herramienta con `clase` `no-encontrado` (FR-013, FR-020).
7. **Dado** el job de cierre en el modo herramienta, **Cuando** se juzgan las evals de `boe-legislacion` y de `legal-core`, **Entonces** las series que deciden pasan y cada umbral que decide se cumple (FR-040, FR-043, SC-001).

---

### User Story 2 - Las dos skills dicen cada paso de las dos formas y eligen la que el agente tiene (Priority: P1)

Quien lee `SKILL.md` —el modelo, una vez por conversación en que la skill se activa— encuentra cada operación nombrada como herramienta y como orden. Si el agente tiene la herramienta, la usa; si no, ejecuta la orden. Donde la orden encadena la lectura y `kitlegal graph check` con `&&`, con herramientas son dos llamadas seguidas, y la segunda solo si la primera no falló. Donde la orden lee un código de salida, con herramientas lee la `clase` del sobre. Lo demás del protocolo no cambia.

**Why this priority**: una herramienta que la skill no sabe pedir no sirve a nadie, y una skill que solo supiera pedir herramientas dejaría de funcionar donde hoy funciona.

**Independent Test**: `skills-check` y la comprobación de nombres (SC-009), en `make ci`; las evals de las dos skills en los dos modos (SC-001).

**Acceptance Scenarios**:

1. **Dado** las dos skills tras `make skills-sync`, **Cuando** se lee su tabla de comandos, **Entonces** cada fila nombra la orden (`kitlegal <applet> <verbo> …`) y la herramienta (`<applet>_<verbo>`) de la misma operación, y `make skills-check` no encuentra drift (FR-030).
2. **Dado** la tabla de comandos de cada skill, **Cuando** se comprueba en `make ci`, **Entonces** toda herramienta que nombra existe en el servidor y toda orden, en el registro (FR-031).
3. **Dado** una sesión del modo orden, **Cuando** responde una eval que activa la skill, **Entonces** pasa como hoy, con las órdenes (FR-032, FR-040).
4. **Dado** una sesión del modo herramienta de la eval 20 (dos bloques de la LCSP con la redacción cambiada), **Cuando** responde, **Entonces** pasa solo si lee `a1-30` y `da-3` y comprueba con `graph_check` y la norma, no pide `graph_show`, cita los dos bloques y lleva sus dos líneas `⚠ REDACCIÓN MODIFICADA:` (FR-033, FR-042).
5. **Dado** una llamada a `boe_articulo` que devuelve un error con `clase` `no-encontrado`, **Cuando** la skill responde, **Entonces** no llama a `graph_check` para ese bloque y aplica lo que hoy aplica al código `3` (FR-033, FR-034).
6. **Dado** los dos `SKILL.md` del hito, **Cuando** se cuentan sus líneas, **Entonces** cada uno tiene menos de 300 (FR-036).

---

### User Story 3 - Sin herramienta y sin binario, la respuesta lo dice y no afirma nada de memoria (Priority: P1)

Una persona tiene la skill y nada más: el agente no tiene la herramienta, y la orden falla porque `kitlegal` no está. Pregunta por un artículo. La respuesta no dice qué dice la norma ni lleva ninguna cita. Lleva una línea de forma fija, `⚠ SIN CONSULTA AL BOE:`, con la causa y `https://kitlegal.es/instalar/`.

**Why this priority**: es el caso de quien instala solo la skill, y el de la web y el móvil. Sin la regla, la respuesta de memoria cita un precepto que nadie leyó (constitución, principio II).

**Independent Test**: los tests de `Juzgar` (SC-012), en `make ci`; las dos evals sin binario ni servidor, en el job de cierre (SC-001).

**Acceptance Scenarios**:

1. **Dado** una sesión sin `kitlegal` en el `PATH` y sin servidor declarado, **Cuando** se le pregunta a `boe-legislacion` qué dice un artículo, **Entonces** la eval pasa solo si la respuesta lleva una línea que empieza por `⚠ SIN CONSULTA AL BOE:` y tiene en esa misma línea `https://kitlegal.es/instalar/`, y no lleva ninguna cita (FR-035, FR-046, FR-047).
2. **Dado** una sesión sin binario ni servidor, **Cuando** se le pregunta a `legal-core` por el territorio de un municipio, **Entonces** la eval pasa solo si la respuesta lleva esa misma línea y ninguna cita (FR-035, FR-046, FR-047).
3. **Dado** una respuesta con la línea y, además, una cita con la forma `[BOE-A-…, bloque …]`, **Cuando** se juzga en una de esas dos evals, **Entonces** la sesión no pasa (FR-047).
4. **Dado** una respuesta sin la línea, o con la etiqueta y sin la dirección en su misma línea, **Cuando** se juzga, **Entonces** la sesión no pasa (FR-047).

---

### User Story 4 - Quien declara el servidor en su agente lo hace con una línea, y el servidor se comporta como un servidor (Priority: P2)

Una persona con `kitlegal` instalado declara `kitlegal mcp serve` en su agente con una línea de configuración. El agente lo arranca desde donde quiera —la app de Claude, con `/` como directorio de trabajo y desde una ruta suya con espacios— y con la versión del protocolo que hable. El servidor anuncia sus herramientas como de solo lectura, da unas instrucciones que se bastan, escribe en la salida estándar solo mensajes del protocolo y termina cuando el agente cierra su entrada.

**Why this priority**: sin esto el servidor no arranca o el agente lo descarta, pero no cambia lo que dice la respuesta.

**Independent Test**: los e2e de FR-072 a FR-076 (SC-005 a SC-008) y el test de las instrucciones (SC-010), en `make ci`.

**Acceptance Scenarios**:

1. **Dado** un manifiesto de skills instaladas de otra versión, **Cuando** se arranca `kitlegal mcp serve --verbose` y un cliente lista las herramientas y llama a una, **Entonces** cada línea de la salida estándar es un mensaje del protocolo, y el aviso de versión y el registro de eventos están en la salida de error (FR-023).
2. **Dado** el mismo binario, **Cuando** se conecta un cliente de la especificación 2026-07-28 y, en otro arranque, uno de la anterior (2025-11-25), **Entonces** los dos listan las mismas herramientas y reciben el mismo sobre de una misma llamada, salvo `fecha_consulta` (FR-007).
3. **Dado** el servidor en marcha, **Cuando** el cliente cierra la entrada estándar, **Entonces** el servidor termina con el código 0 (FR-024).
4. **Dado** el binario copiado a una ruta con espacios y lanzado con `/` como directorio de trabajo, **Cuando** un cliente llama a una herramienta, **Entonces** recibe el mismo sobre que desde cualquier otro sitio, y la caché y el grafo están donde siempre (FR-025).
5. **Dado** `kitlegal mcp serve --asunto x`, **Cuando** se ejecuta, **Entonces** sale con el código 2 y la clase `argumentos`, sin atender ningún mensaje (FR-021).
6. **Dado** el servidor, **Cuando** un cliente lista sus herramientas, **Entonces** cada una se anuncia de solo lectura, y las instrucciones del servidor llevan sus cinco contenidos en sus primeros 512 caracteres (FR-005, FR-006).
7. **Dado** `kitlegal mcp serve --dry-run` con un mensaje del protocolo en su entrada, **Cuando** se ejecuta, **Entonces** termina sin responderlo, con la salida estándar vacía y la descripción de la operación en la salida de error (FR-022).
8. **Dado** el servidor, **Cuando** un cliente lanza a la vez varias llamadas a la misma herramienta y a herramientas distintas, **Entonces** cada llamada recibe su resultado, y ninguna el de otra (FR-014).

---

### User Story 5 - El job de evals mide las dos skills en los dos modos, y cada umbral decide en cada uno (Priority: P1)

El informe del job de cada skill da, por modo, las tasas de sus evals y sus umbrales. En `boe-legislacion`, los cuatro umbrales que hoy deciden deciden en el modo orden y en el modo herramienta, con el mismo valor. Si uno no se cumple en un modo, el veredicto es `fallo` y el job sale en rojo, aunque el otro modo lo cumpla. Las dos evals sin binario ni servidor deciden con su serie.

**Why this priority**: es la Aceptación del hito. Sin medir el modo herramienta, lo ganado de H7.2 a H7.4 valdría solo donde hay shell.

**Independent Test**: los tests del informe con sesiones sintéticas (SC-011) y los de `Juzgar` (SC-012), en `make ci`; el job de cierre (SC-001).

**Acceptance Scenarios**:

1. **Dado** el job de cierre de `boe-legislacion`, **Cuando** se lee `umbrales` de su informe, **Entonces** lleva, para cada modo, `expresiones_prohibidas:<modelo>`, `sin_activar:<modelo>`, `redaccion_no_leida:<modelo>` y `duracion_de_las_sesiones`, cada uno con `decide: true` y `cumple: true`, y su nombre dice de qué modo es (FR-043).
2. **Dado** un informe sintético en el que 3 de las 54 respuestas del modelo que decide llevan una expresión prohibida en el modo herramienta y 0 de 54 en el modo orden, **Cuando** se escribe, **Entonces** el umbral del modo herramienta lleva `cumple: false`, el veredicto es `fallo` y el motivo nombra ese umbral con su modo; con 2 de 54 en los dos modos, los dos se cumplen y el veredicto no cambia (FR-043, FR-044).
3. **Dado** una eval con un `comando` de bloque (`applet: boe`, norma y bloque), **Cuando** la sesión llama a la herramienta `boe_articulo` con esa norma y ese bloque y su resultado no es un error de herramienta (el sobre lleva `ok` verdadero), **Entonces** el comando queda satisfecho igual que con la orden `kitlegal boe articulo` que termina con 0; si el resultado es un error de herramienta, no lo satisface, como la orden que termina con otro código (FR-042).
4. **Dado** una eval con `graph show` en `prohibidos`, **Cuando** la sesión llama a la herramienta `graph_show`, **Entonces** la sesión no pasa (FR-042).
5. **Dado** el job de cierre de `legal-core`, **Cuando** se lee su informe, **Entonces** `umbrales` es `[]` y sus series que deciden pasan en los dos modos (FR-045).
6. **Dado** una serie de una eval que decide que pasa en el modo orden y no llega a 2 de 3 en el modo herramienta, **Cuando** se escribe el informe, **Entonces** el veredicto es `fallo` (FR-044).

---

### User Story 6 - Un applet nuevo aparece como herramienta sin tocar `mcp` (Priority: P2)

Quien añade un applet o un verbo al registro en un hito posterior (H20, H8, H9) no escribe nada en `mcp`: el servidor lo anuncia como herramienta, con el nombre, la descripción y los esquemas de su `--describe`.

**Why this priority**: es lo que hace que los hitos siguientes nazcan con su herramienta (ADR 0035, Consecuencias). No cambia nada para quien pregunta hoy.

**Independent Test**: el control de conformidad (SC-003), en `make ci`.

**Acceptance Scenarios**:

1. **Dado** el registro de producción, **Cuando** un cliente lista las herramientas, **Entonces** son exactamente `boe_buscar`, `boe_indice`, `boe_articulo`, `boe_articulos`, `boe_metadatos`, `boe_analisis`, `territorio_resolver`, `graph_check`, `graph_show` y `graph_stats`: los verbos del registro menos los de `skills` y `mcp` (FR-002).
2. **Dado** un registro con un applet más, **Cuando** un cliente lista las herramientas, **Entonces** aparecen las de sus verbos, sin ningún cambio en `mcp` (FR-004).
3. **Dado** cada herramienta, **Cuando** se comparan sus dos esquemas con `--describe` de su verbo, **Entonces** el de entrada es la entrada de `--describe` sin las banderas globales y el de salida es su salida (FR-003).

---

### User Story 7 - Quien lee el README sabe cómo declarar el servidor y dónde no es compatible (Priority: P3)

**Why this priority**: sin la línea de configuración nadie llega a la primera pregunta, pero es documentación.

**Independent Test**: lectura del README, de CONTRIBUTING y de `CHANGELOG.md` contra FR-060 y FR-061 (SC-013).

**Acceptance Scenarios**:

1. **Dado** el README, **Cuando** se busca cómo declarar el servidor, **Entonces** lo dice para Claude Code (`claude mcp add`), para la app de escritorio de ChatGPT y Codex (`codex mcp add kitlegal -- kitlegal mcp serve`, o *Settings > MCP servers*) y para Antigravity (`mcp_config.json`), y dice que ChatGPT y Claude en la web y en el móvil solo admiten servidores remotos y no son compatibles (FR-060).

---

### Edge Cases

- **Una llamada a una herramienta que el servidor no anuncia** (`skills_install`, `mcp_serve`, `version`): no hay verbo que invocar, así que no hay sobre. La responde el protocolo como herramienta desconocida.
- **Una bandera global como argumento de una herramienta** (`offline: true` en `boe_articulo`): el esquema de entrada no la declara, así que es un argumento inválido: error de herramienta con `clase` `argumentos` (FR-011, FR-020).
- **`boe_articulo` y `graph_check` lanzadas a la vez**, y no una detrás de otra: cada una recibe su resultado (FR-014), pero `graph_check` puede no ver esa lectura. La skill las pide seguidas, y la segunda después de recibir el resultado de la primera (FR-033).
- **`boe_articulos` con un bloque que falla**: falla entera, como la orden, y la skill pide cada bloque por separado, como hoy.
- **Una llamada que falla** —por la clase que sea, también `inesperado`—: el servidor sigue atendiendo las siguientes (FR-015).
- **Una llamada que pasa del plazo de `--timeout`**: devuelve como error de herramienta el sobre de la clase que da la orden cuando agota su plazo (FR-020).
- **`territorio_resolver` con el servidor en `--offline` y la caché vacía**: responde, porque no consulta ninguna fuente (FR-010).
- **La entrada estándar se cierra con llamadas en curso**: el servidor termina con 0 (FR-024). No se promete resultado a un cliente que ya no escucha, y `world.db` no queda a medias: la entrega es transaccional desde H7.
- **Dos servidores a la vez, o un servidor y una orden**, sobre la misma caché y el mismo grafo: como varias invocaciones de hoy (H7: una entrega espera a la otra dentro del plazo).
- **El agente tiene la herramienta y también el binario en el `PATH`**: la skill usa la herramienta (FR-032).
- **El agente tiene la herramienta y la llamada falla**: no es el caso de `⚠ SIN CONSULTA AL BOE:`. La skill lee la `clase` y aplica la regla que hoy aplica a ese código (FR-034, FR-035).
- **En el modo herramienta, la sesión ejecuta la orden**: el binario no está en el `PATH` de su shell y la orden falla. Una orden que no llega a ejecutar `kitlegal` no satisface ningún `comando` (FR-042).
- **Un nombre de herramienta con el prefijo que le pone el agente** (Claude Code las presenta como `mcp__<servidor>__<herramienta>`): cuenta como la herramienta `<applet>_<verbo>` del servidor (FR-042).
- **3 de 54 en un modo y 0 de 54 en el otro**: 3/54 ≈ 0,0556 > 0,05, así que el umbral de ese modo no se cumple y el veredicto es `fallo`. Sumados (3 de 108 ≈ 0,028) se cumpliría: por eso los modos no se suman (FR-043).
- **Lo que llega de fuera manipulado** —un mensaje que no es del protocolo, una caché o un `world.db` tocados a mano—: lo responde el protocolo o lo cubre la regla genérica (defecto `inesperado`, ADR 0023); este spec no lo especifica caso a caso.

## Requirements *(mandatory)*

### Functional Requirements

#### El applet `mcp` y sus herramientas

- **FR-001**: El binario MUST tener un applet `mcp` con un verbo `serve`. `kitlegal mcp serve` atiende el protocolo MCP por la entrada y la salida estándar, que es su único transporte: MUST NOT abrir ningún puerto ni atender ninguna conexión de red. Comprobable: FR-071.
- **FR-002**: Las herramientas MUST derivarse del registro, nunca de una lista escrita a mano: una por cada verbo de cada applet registrado, salvo los verbos de los applets `skills` y `mcp` y los verbos reservados del binario (`version`). Con el registro de hoy son exactamente diez: `boe_buscar`, `boe_indice`, `boe_articulo`, `boe_articulos`, `boe_metadatos`, `boe_analisis`, `territorio_resolver`, `graph_check`, `graph_show` y `graph_stats`. Comprobable: FR-070.
- **FR-003**: Cada herramienta MUST llamarse `<applet>_<verbo>`, llevar como descripción la del verbo, como esquema de entrada (`inputSchema`) la entrada de `--describe` de ese verbo sin las ocho banderas globales, y como esquema de salida (`outputSchema`) la salida de `--describe` de ese verbo: el sobre, con `data` condicionado a `ok`. Comprobable: FR-070.
- **FR-004**: Un applet o un verbo nuevo en el registro MUST aparecer como herramienta sin ningún cambio en `mcp`. Comprobable: el control de conformidad lo ejerce con un registro que lleva un applet más (FR-070).
- **FR-005**: Las diez herramientas de hoy MUST anunciarse de solo lectura (`readOnlyHint`): sus verbos no cambian nada fuera de la caché y del grafo del mundo del propio kitlegal. Es lo que Codex mira para no pedir aprobación en cada llamada. Comprobable: FR-070.
- **FR-006**: El servidor MUST dar unas `instructions` de texto fijo. Sus primeros 512 caracteres MUST decir cinco cosas, porque es la parte que la documentación de Codex pide que se baste a sí misma: (1) qué son las herramientas; (2) que ningún contenido legal se afirma si no viene del texto devuelto; (3) que cada afirmación lleva norma y bloque; (4) que los avisos del sobre se trasladan; y (5) que el protocolo completo son las skills de kitlegal. El plan fija el texto. Comprobable: un test de `make ci` falla si alguno de los cinco contenidos queda fuera de los primeros 512 caracteres (FR-077).
- **FR-007**: El mismo proceso MUST atender a un cliente de la especificación vigente del protocolo (2026-07-28, sin `initialize`) y a uno de las anteriores. El servidor MUST NOT estrechar la lista de versiones que atiende el SDK: la app de escritorio de Claude negocia hoy la 2025-11-25 (ADR 0035). Comprobable: FR-073.
- **FR-008**: El servidor MUST ofrecer solo herramientas. MUST NOT anunciar `prompts` ni `resources`, ni servir por el protocolo las skills empotradas. Comprobable: FR-070.

#### Una llamada es una invocación del kernel

- **FR-010**: Una llamada a una herramienta MUST recorrer el mismo camino que la orden de su verbo: los mismos argumentos, la misma caché, `internal/httpx`, el mismo sobre, la misma huella y su fecha de consulta. Su resultado MUST llevar el sobre en `structuredContent` y, serializado, en un bloque de texto. Para la misma entrada y el mismo estado, ese sobre MUST ser el de la orden con `--json`, salvo `fecha_consulta`. Comprobable: FR-071.
- **FR-011**: Una llamada que falla MUST devolver el sobre de fallo —`ok` falso, con `clase` y `mensaje` en `data`— marcado como error de herramienta. La `clase` es la del ADR 0023 y hace el papel del código de salida de la orden: `argumentos` (2), `no-encontrado` (3), `fuente-no-disponible` (4), `limite-o-tos` (5), `identidad-humana` (6), `conflicto` (7) e `inesperado` (1). Unos argumentos que el esquema de entrada no admite —falta uno obligatorio, sobra una propiedad, un valor no es del tipo declarado— o que la orden rechazaría con el código 2 MUST dar la clase `argumentos`. Comprobable: FR-071.
- **FR-012**: Los hallazgos de `graph_check` MUST ser un resultado y no un error: el sobre lleva `ok` verdadero con los hallazgos en `data`, y el resultado no está marcado como error de herramienta (ADR 0023). Comprobable: FR-071.
- **FR-013**: Lo que observa una llamada que termina bien MUST entregarse al grafo del mundo como lo entrega la orden: después de tener el resultado, sin cambiarlo —el sobre es el mismo si la entrega falla, y el fallo va a la salida de error, como en H7—, y nunca en una llamada que falla. Una llamada que el cliente envía después de recibir el resultado de otra MUST ver lo que esa entregó: es lo que hace que `graph_check`, pedida detrás de `boe_articulo`, compare la lectura que se acaba de hacer. Con `mcp serve --no-graph` no se entrega nada. Comprobable: FR-071.
- **FR-014**: Las llamadas simultáneas, a la misma herramienta o a herramientas distintas, MUST recibir cada una su resultado, sin mezclar el de una con el de otra. Las que piden a una misma fuente MUST respetar entre todas el ritmo por sitio de `internal/httpx` (constitución, principio I). Comprobable: FR-074.
- **FR-015**: Una llamada que falla, por la clase que sea, MUST NOT terminar el servidor: sigue atendiendo las siguientes. Comprobable: en el e2e de FR-071, las llamadas que fallan y las que no van contra el mismo proceso.

#### El proceso del servidor

- **FR-020**: `--offline`, `--no-graph` y `--timeout` MUST darse a `mcp serve` y valer para todas sus llamadas, cada una con el significado que tiene en la orden; `--timeout` es el plazo de cada llamada, con el valor por omisión de la orden si no se da, y nunca el de la vida del servidor, que no tiene plazo. MUST NOT ser parámetros de ninguna herramienta. Comprobable: FR-071 (`--offline`, `--no-graph`) y FR-070 (ningún esquema de entrada lleva una bandera global).
- **FR-021**: `kitlegal mcp serve --asunto <valor>` MUST ser un error de argumentos: sale con el código 2 y la clase `argumentos`, como cualquier error de argumentos del kernel, sin atender ningún mensaje. El servidor MUST NOT exponer nada del asunto: ninguna herramienta lo recibe ni lo lee. Comprobable: FR-075.
- **FR-022**: Las demás banderas globales MUST valer en `mcp serve` lo que valen en cualquier verbo: `--verbose` sube el registro de eventos, que va a la salida de error; `--describe` da el esquema del verbo sin ejecutarlo, y lo resuelve el kernel antes de llegar al applet; `--json` no cambia nada del protocolo. `--dry-run` no lo corta el kernel, que llama al applet con la bandera: MUST honrarlo el propio applet. `kitlegal mcp serve --dry-run` no atiende ningún mensaje, deja la salida estándar vacía y termina sin esperar a que se cierre la entrada, como termina cualquier verbo con `--dry-run`, con la descripción del kernel en la salida de error. Comprobable: FR-072 (`--verbose`), FR-075 (`--dry-run`) y el control del kernel que ya recorre cada verbo del registro con `--describe`.
- **FR-023**: Mientras sirve, la salida estándar MUST llevar solo mensajes del protocolo. El registro de eventos y el aviso de versión distinta (H19) MUST ir a la salida de error. El aviso MUST darse una vez por arranque del servidor, no una por llamada. Comprobable: FR-072.
- **FR-024**: Al cerrarse la entrada estándar, el servidor MUST terminar con el código 0. Comprobable: FR-075.
- **FR-025**: El servidor MUST NOT depender de su directorio de trabajo ni de la ruta de su ejecutable: lanzado con `/` como directorio de trabajo y desde una ruta con espacios, responde igual. La caché y el grafo siguen en `~/.cache/kitlegal/`, o donde diga `KITLEGAL_CACHE_DIR`, como en cualquier orden. Comprobable: FR-076.
- **FR-026**: La dependencia que entra es `modelcontextprotocol/go-sdk`, la que el principio V prevé para la distribución. MUST importarse solo desde el paquete del servidor, y las reglas de arquitectura de hoy MUST seguir cumpliéndose. Comprobable: FR-079.

#### Las dos skills

- **FR-030**: `make skills-sync` MUST generar la tabla de comandos de cada skill con la orden y la herramienta de cada operación: cada fila nombra `kitlegal <applet> <verbo> …` y `<applet>_<verbo>`, y sale de `--describe` y del registro. La región generada MUST decir además que la herramienta devuelve el mismo sobre que la orden. `make skills-check` MUST fallar con cualquier diferencia entre lo generado y lo que hay. Comprobable: FR-077.
- **FR-031**: Toda herramienta que nombra la tabla de una skill MUST existir en el servidor, y toda orden, en el registro. Comprobable: una comprobación de `make ci` falla si no (FR-077).
- **FR-032**: El texto de los dos `SKILL.md` MUST decir que cada operación se pide con la herramienta si el agente la tiene y con la orden si no, y MUST decir cada paso de las dos formas. Comprobable: las evals de las dos skills pasan en los dos modos (SC-001).
- **FR-033**: Donde `SKILL.md` encadena hoy la lectura y la comprobación en una orden (`kitlegal boe articulo … && kitlegal graph check …`, y la de `articulos`), MUST decir que con herramientas son dos llamadas seguidas —la de lectura y, recibido su resultado, `graph_check` con la misma norma y los mismos bloques— y que la segunda solo se hace si la primera no falló. Comprobable: la eval 20 y la 19 en el modo herramienta (SC-001).
- **FR-034**: Donde `SKILL.md` lee hoy un código de salida, MUST decir que con herramientas se lee la `clase` del sobre, con su correspondencia (FR-011): en `boe-legislacion`, `3`, `4` y `5`, y el «otro código que `0`» de la regla 7; en `legal-core`, `2`, `3` y «cualquier otro código». Lo que la skill hace con cada caso no cambia. Comprobable: las evals en el modo herramienta (SC-001).
- **FR-035**: **Regla nueva**, en las dos skills. Si el agente no tiene la herramienta y la orden falla porque el binario no está, la respuesta MUST NOT afirmar nada del contenido de la norma —en `legal-core`, tampoco ningún dato de territorio— y MUST llevar una línea con forma fija: empieza por `⚠ SIN CONSULTA AL BOE:` y lleva detrás, en la misma línea, la causa y `https://kitlegal.es/instalar/`. La respuesta MUST NOT llevar ninguna cita. La regla sustituye a lo que hoy dicen las dos skills cuando `kitlegal` no está en el `PATH`. El plan fija la forma exacta, con un marcador en lugar de la causa. Comprobable: FR-046 y FR-081.
- **FR-036**: MUST NOT haber más cambios de protocolo, de reglas, ni de la forma de la cita y de los avisos. Cada `SKILL.md` MUST seguir por debajo de 300 líneas (hoy, 298 el de `boe-legislacion` y 170 el de `legal-core`), con frontmatter válido, sin `scripts/` y sin nombrar evals, el job ni modelos (H5 FR 077). Las comprobaciones de H7.4 sobre `boe-legislacion` MUST seguir en verde sin recortar la lista de expresiones prohibidas: el calibrado sobre los tres informes versionados (36, 11 y 9), la prosa sin expresiones de la lista y las órdenes para PowerShell. Si la línea nueva choca con la lista, se declara como forma fija, como la frase de la regla 7 en H7.4. Comprobable: FR-077.

#### Las evals en dos modos

- **FR-040**: El job MUST medir cada eval de las dos skills en el modo orden, el de hoy, y en el modo herramienta: el binario fuera del `PATH` del shell de la sesión y el servidor declarado al agente. En cada modo valen los mismos modelos —el que decide y el informativo—, las mismas repeticiones y la misma regla por serie (ADR 0016) que hoy. La sesión de la prueba de red sigue siendo solo del modo orden. Comprobable: SC-001.
- **FR-041**: Una sesión del modo herramienta MUST tener las mismas garantías que una del modo orden: la skill tal como la deja la instalación, sin más red que la del modelo, y con la caché y el grafo preparados para esa sesión, que son los que usa el servidor. Una llamada que pide algo fuera de lo grabado se publica en el informe como hoy la orden que lo pide. Comprobable: el informe del job de cierre publica las invocaciones de cada sesión (FR-042) y lo que queda fuera de lo grabado.
- **FR-042**: Un `comando` de una eval MUST quedar satisfecho igual por una orden que por una llamada a la herramienta del mismo applet —y del mismo verbo, si el comando lo nombra— con la misma norma y el mismo bloque, o con los mismos términos o el mismo municipio en los comandos que los llevan. La llamada lo satisface solo si su resultado no es un error de herramienta —el sobre con `ok` verdadero—, como hoy la orden solo si termina con el código 0. Un comando de `prohibidos` MUST contar con la herramienta de ese applet y ese verbo con la regla con que hoy cuenta la orden: termine como termine. La herramienta se reconoce por su nombre `<applet>_<verbo>`, lleve o no el prefijo que le pone el agente. El informe MUST publicar, por sesión, las llamadas a herramientas junto a las órdenes. Comprobable: FR-081.
- **FR-043**: Cada umbral que hoy decide (ADR 0029 y H7.4) MUST decidir en cada modo, con su mismo valor, y el informe MUST darlos por modo. En `boe-legislacion` son, por cada modo:
  - `expresiones_prohibidas:<modelo>`: `"<="` 0.05 sobre las respuestas del modelo que decide en las evals que activan la skill en ese modo (54 con todas terminadas: admite 2);
  - `sin_activar:<modelo>`: `"<="` 0 sobre esas mismas respuestas;
  - `redaccion_no_leida:<modelo>`: `"<="` 0 sobre esas mismas respuestas;
  - `duracion_de_las_sesiones`: `"<="` 900 s, de la preparación de la primera sesión de ese modo al final de la última.

  Los cuatro llevan `decide: true`. El del modelo informativo sigue con `decide: false`, también por modo. Cada elemento de `umbrales` MUST decir su modo en el nombre, único en el informe y con el patrón del contrato del ADR 0029; el plan fija la forma. Las medidas de un modo MUST NOT sumarse a las del otro. Lo que cuenta en cada medida y en cada total es lo que cuenta hoy (H7.4 FR 045).
- **FR-044**: Un umbral de FR-043 que no se cumple en un modo MUST poner el veredicto de la skill en `fallo`, con un motivo que nombra el umbral, su modo, la medida y, si lo tiene, el total, y el job MUST salir en rojo (H7.3 FR 003). Una serie de una eval que decide que no llega a su umbral en un modo MUST poner también el veredicto en `fallo`, aunque pase en el otro. Comprobable: FR-080.
- **FR-045**: `legal-core` MUST seguir sin umbrales: su `umbrales` es `[]` (H7.3 FR 006). Lo que decide en cada modo son sus series.
- **FR-046**: Cada skill MUST ganar una eval nueva, sin binario ni servidor: la sesión no tiene `kitlegal` en el `PATH` de su shell ni el servidor declarado. La de `boe-legislacion` pregunta qué dice un artículo de una norma; la de `legal-core`, por el territorio de un municipio. Las dos deciden: su serie tiene que pasar con el modelo que decide (ADR 0016). `boe-legislacion` pasa a 21 evals y `legal-core` a 4. El plan fija las preguntas y cómo declara el formato de eval una eval así.
- **FR-047**: Una sesión de una eval de FR-046 MUST pasar solo si la respuesta lleva una línea que empieza por `⚠ SIN CONSULTA AL BOE:` y tiene en esa misma línea `https://kitlegal.es/instalar/`, y la respuesta no lleva ninguna cita: ningún texto con la forma `[… BOE-A-…, bloque <id>]`, sea de la norma que sea. Todo se compara sin modelo. Estas evals no son de ninguno de los dos modos: se miden una vez, y sus sesiones MUST NOT entrar en la medida, en el total ni en la duración de ningún umbral de FR-043. Comprobable: FR-081.
- **FR-048**: El informe de cada skill MUST seguir siendo uno, el que hoy recoge el cierre del workflow: el de la comprobación `evals (<skill>)`, con los dos modos y las evals de FR-046 dentro. MUST conseguirse sin cambiar `scripts/workflow/`, que es el proceso del workflow y no del hito. El informe MUST poder leerse con el `scripts/workflow/informe.sh` de hoy, que escribe con él la sección de evals del informe final, lo único que lee la persona (ADR 0018). Ese guion da una celda por eval y modelo de `tasas` y una fila de recuento de expresiones prohibidas por modelo, así que ninguna pareja de fila y modelo de `tasas` MUST repetirse entre modos —cada serie de cada modo tiene su celda, con su ✗ si no pasa— y cada recuento de expresiones MUST decir de qué modo es. El plan fija cómo. Comprobable: FR-080. El tope del trabajo MUST seguir cubriendo el peor caso, recalculado con las evals del repositorio y los dos modos (H7.3 FR 035). Si no se pudiera sin tocar `scripts/workflow/`, queda como supuesto en `gates/supuestos.md` y ese cambio va en su propia rama.
- **FR-049**: Ninguna corrección del run MUST cumplir un umbral de FR-043 rebajándolo, dejándolo en `decide: false`, sacando evals o un modo del total, sumando los modos o recortando la lista (ADR 0029).
- **FR-050**: `make evals-sondeo` MUST seguir midiendo solo el modo orden. Sus cinco argumentos (`SKILL`, `EVALS`, `MODELO`, `REPETICIONES`, `CONCURRENCIA`), la preparación de cada sesión —el binario y la skill del árbol de trabajo en el `PATH`, ningún servidor declarado—, su juicio, su salida y sus códigos de salida MUST NOT cambiar para las evals de los dos modos. El modo herramienta y las dos evals de FR-046 se miden solo en el job. Pedir en `EVALS` una eval que el formato de eval declara sin binario ni servidor MUST ser un error de uso, como hoy un argumento que no vale: una línea en la salida de error que nombra el argumento `EVALS` y el número de la eval y dice que esa eval solo la mide el job —una línea por eval así, junto a los demás errores de argumentos y en su orden—, la salida estándar vacía, ninguna sesión abierta ni árbol preparado, y el guion termina con 1 (`make`, con 2 detrás). Vale también cuando la lista lleva además evals que el sondeo sí mide: con un error de argumentos no se abre ninguna. El plan fija el texto de la línea. Comprobable: FR-084.
- **FR-051**: Lo que cambie bajo `schemas/` o `testdata/` —el formato de eval, los transcripts sintéticos de los tests— MUST ir en tareas `[datos]`. El e2e se sirve de las respuestas del BOE grabadas desde H4; si necesita una que no está grabada, MUST grabarla el paso `grabar_datos` (fila `boe.legislacion-consolidada` revisada en `docs/SOURCES.md`).

#### Documentación

- **FR-060**: El README MUST decir cómo se declara el servidor en Claude Code (`claude mcp add`), en la app de escritorio de ChatGPT y en Codex (`codex mcp add kitlegal -- kitlegal mcp serve`, o *Settings > MCP servers*) y en Antigravity (`mcp_config.json`), y que ChatGPT y Claude en la web y en el móvil solo admiten servidores remotos y no son compatibles. MUST dejar de anunciar el servidor MCP como algo por venir.
- **FR-061**: CONTRIBUTING MUST explicar el applet `mcp`, de dónde salen sus herramientas y los dos modos del job de evals, y MUST decir en «Sondeo local» que el sondeo mide solo el modo orden y que el modo herramienta y las evals sin binario ni servidor son del job (FR-050). `CHANGELOG.md` (*Unreleased*) MUST registrar `kitlegal mcp serve` con sus herramientas, las dos formas de cada operación en las dos skills, la regla `⚠ SIN CONSULTA AL BOE:` y las evals en dos modos. Las tablas de comandos de las dos skills MUST quedar regeneradas (FR-030).

#### Controles y Definition of Done

- **FR-070**: Un control de conformidad en `make ci` MUST fallar si el conjunto de herramientas del servidor no es exactamente el de los verbos del registro menos los excluidos (FR-002), si alguna no lleva el nombre y la descripción de su verbo, si alguno de sus dos esquemas no es el de `--describe` de su verbo —el de entrada, sin las banderas globales— (FR-003), si alguna de las de hoy no se anuncia de solo lectura (FR-005), o si el servidor anuncia `prompts` o `resources` (FR-008). MUST ejercerse también con un registro que lleva un applet más (FR-004).
- **FR-071**: Un e2e en `make ci`, con un cliente MCP real contra el binario por stdio, con replay y `HOME` temporal, MUST comprobar: para cada herramienta, que el sobre es el de su orden con `--json` para la misma entrada, salvo `fecha_consulta`, en `structuredContent` y en el bloque de texto; que un bloque inexistente, una fuente no disponible (`--offline` con la caché vacía) y unos argumentos inválidos devuelven el sobre de su clase como error de herramienta; que `graph_check` con hallazgos no es un error; y que tras `boe_articulo`, `graph_show` ve el bloque, y con `mcp serve --no-graph` no.
- **FR-072**: Un e2e MUST comprobar que, con un manifiesto de skills de otra versión y `--verbose`, cada línea de la salida estándar es un mensaje del protocolo y el aviso está en la salida de error, una sola vez.
- **FR-073**: Un e2e MUST comprobar que el mismo binario responde a un cliente de la especificación 2026-07-28 y a uno de la anterior.
- **FR-074**: Un test con `-race` MUST lanzar llamadas simultáneas a la misma herramienta y a herramientas distintas y comprobar que cada una recibe su resultado.
- **FR-075**: Un e2e MUST comprobar que, al cerrarse la entrada estándar, el servidor termina con 0; que `mcp serve --asunto` sale con 2 sin atender ningún mensaje; y que `mcp serve --dry-run`, con la entrada abierta y un mensaje del protocolo en ella, termina sin responderlo, con la salida estándar vacía y la descripción en la salida de error (FR-022).
- **FR-076**: Un e2e MUST comprobar que el servidor, lanzado con `/` como directorio de trabajo y desde una ruta con espacios, responde igual.
- **FR-077**: `make ci` MUST comprobar: `skills-check` sin drift; que toda herramienta que nombra la tabla de una skill existe en el servidor y toda orden, en el registro; que cada `SKILL.md` tiene menos de 300 líneas; que las comprobaciones de H7.4 sobre `boe-legislacion` siguen en verde (FR-036); y que los cinco contenidos de las `instructions` están en sus primeros 512 caracteres (FR-006).
- **FR-078**: El job de evals MUST ejecutarse en la propuesta de cambio (lo hace el workflow tras la revisión final), en los dos modos, con sus umbrales por modo y con las dos evals sin binario ni servidor (Definition of Done §1.10). Su informe da lo que pide SC-001.
- **FR-079**: `depguard` MUST fallar si el SDK de MCP se importa desde otro paquete que el del servidor, y `make ci` MUST quedar en verde, con `schema-check` sin drift.
- **FR-080**: Los tests del informe con sesiones sintéticas MUST cubrir, en `make ci`: cada umbral de FR-043 incumplido en un solo modo, que da `fallo` con un motivo que nombra el umbral y su modo; los dos modos cumplidos, que no cambia el veredicto; 3 de 54 en un modo y 0 de 54 en el otro, que no se cumple; una serie que decide y falla solo en un modo, que da `fallo`; y, leído un informe con los dos modos por el `scripts/workflow/informe.sh` de hoy, una celda por cada serie de cada modo —la que no pasa en un solo modo, con su ✗— y una fila de recuento de expresiones por modelo y modo (FR-048).
- **FR-081**: Los tests de `Juzgar` MUST cubrir, en `make ci`: un `comando` satisfecho por la llamada a su herramienta y no satisfecho por la de otra norma u otro bloque, ni por la llamada a la herramienta correcta, con la norma y el bloque de la eval, cuyo resultado es un error de herramienta; un nombre de herramienta con prefijo del agente; un `prohibido` por herramienta, que no pasa; y, en una eval sin binario ni servidor, la respuesta con la línea y sin citas, que pasa, y las que no pasan: sin la línea, con la etiqueta y sin la dirección en su línea, y con la línea y una cita.
- **FR-082**: La sección «Controles de umbral» del plan MUST tener una fila por cada umbral de FR-043 y por cada modo —ocho de `evals:boe-legislacion:…`, con el id del modelo que decide y el nombre que fije el plan—, y las de `make ci`: las 300 líneas de cada `SKILL.md` (FR-036), los 512 caracteres de las `instructions` (FR-006) y el conjunto exacto de herramientas (FR-002).
- **FR-083**: La comprobación de la definición del job (`TestDefinicionDelJob`, H7.3 y H7.4) MUST seguir en verde con los dos modos: el tope del trabajo cubre el peor caso y sigue habiendo una sola tanda por commit y skill.
- **FR-084**: Los tests del sondeo de `make ci`, con el `claude` sustituto, MUST cubrir: pedir en `EVALS` una eval sin binario ni servidor da error de uso, con una línea que nombra `EVALS` y la eval, salida estándar vacía, ninguna sesión abierta ni árbol preparado, y salida 1; lo mismo con una lista que lleva además evals que el sondeo mide; una línea por cada eval así y en su orden entre los demás errores de argumentos; y que pedir solo evals de los dos modos sigue dando la misma preparación, la misma salida y los mismos códigos que antes del hito.

#### Relación con otros hitos

- **FR-090**: Los specs de los hitos anteriores MUST NOT editarse. Este spec nombra lo que sustituye de ellos («Relación con H5, H6, H7.4 y H19»). No hay ADR nuevo: la decisión es la del ADR 0035.

### Key Entities

- **Herramienta**: una operación del servidor, una por verbo de consulta del registro. Tiene nombre (`<applet>_<verbo>`), descripción, esquema de entrada, esquema de salida y la marca de solo lectura.
- **Llamada**: una petición de un cliente a una herramienta. Es una invocación del kernel y devuelve un sobre.
- **Error de herramienta**: el resultado de una llamada que falla: el sobre de fallo, con su `clase`, marcado como error.
- **Modo orden / modo herramienta**: las dos formas en que una sesión de eval tiene la operación a mano: `kitlegal` en el `PATH` y ningún servidor, o el servidor declarado y el binario fuera del `PATH`.
- **Eval sin binario ni servidor**: una eval cuya sesión no tiene ninguna de las dos formas. Juzga la línea `⚠ SIN CONSULTA AL BOE:` y que no hay citas.
- **Línea `⚠ SIN CONSULTA AL BOE:`**: forma fija de la respuesta cuando la skill no ha podido consultar: la etiqueta, la causa y `https://kitlegal.es/instalar/`, en una línea.
- **Umbral por modo**: un elemento de `umbrales` del informe, con el contrato del ADR 0029, medido sobre las sesiones de un solo modo.
- **`instructions`**: el texto fijo que el servidor da al agente al conectarse.

## Success Criteria *(mandatory)*

### Measurable Outcomes

- **SC-001**: En el run, el informe del job de evals da cada umbral cumplido en el modo orden y en el modo herramienta para las dos skills, y las dos evals sin binario ni servidor pasan. En concreto: en `boe-legislacion`, por cada modo, `expresiones_prohibidas:<modelo>` (como mucho 2 de 54), `sin_activar:<modelo>` (0), `redaccion_no_leida:<modelo>` (0) y `duracion_de_las_sesiones` (≤ 900 s), los ocho con `decide: true` y `cumple: true`; en `legal-core`, `umbrales` es `[]`; en las dos, toda serie que decide pasa en cada modo; y el veredicto de las dos es aprobado. **Control**: el job de evals de la skill sale en rojo, con el veredicto `fallo`, si en cualquiera de los dos modos pasan de 2 de 54 las respuestas con alguna expresión, si alguna respuesta queda sin la skill activada o con una redacción no leída, si la duración pasa de 900 s o si una serie que decide no pasa; y también si no pasa la serie de la eval sin binario ni servidor (FR-044, FR-046). El cierre del workflow cuenta ese rojo.
- **SC-002**: Después de fusionar, y fuera del run porque es humano: en la app de escritorio de ChatGPT, con `kitlegal` instalado, las skills en `~/.agents/skills/` y el servidor añadido en *Settings > MCP servers*, «¿qué dice el art. 21 de la Ley 39/2015?» se responde con `art. 21 de la Ley 39/2015 [BOE-A-2015-10565, bloque a21]` sin que el agente ejecute ninguna orden; y la misma pregunta en la app de escritorio de Claude, sin carpeta, con ese binario dentro de un `.mcpb` hecho a mano como el de la prueba del ADR 0035 e instalado con doble clic. Lo mide una persona: no tiene control en el run.
- **SC-003**: El servidor anuncia exactamente las herramientas de los verbos del registro menos los excluidos —10 con el registro de hoy—, 10 de 10 con el nombre, la descripción y los dos esquemas de `--describe` de su verbo y la marca de solo lectura, y 0 `prompts` y 0 `resources`. **Control**: el control de conformidad falla en `make ci` (FR-070).
- **SC-004**: Para 10 de 10 herramientas, el sobre de la llamada es el de su orden con `--json` para la misma entrada, salvo `fecha_consulta`; 3 de 3 fallos (bloque inexistente, fuente no disponible, argumentos inválidos) devuelven el sobre de su clase como error de herramienta; `graph_check` con hallazgos no es un error; y `graph_show` ve el bloque tras `boe_articulo`, y con `--no-graph` no. **Control**: el e2e falla en `make ci` (FR-071).
- **SC-005**: Con un manifiesto de otra versión y `--verbose`, 0 líneas de la salida estándar que no sean un mensaje del protocolo, y 1 aviso de versión en la salida de error. **Control**: el e2e falla en `make ci` (FR-072).
- **SC-006**: 2 de 2 clientes —el de la especificación 2026-07-28 y el de la anterior— listan las herramientas y reciben el sobre de una llamada del mismo binario. **Control**: el e2e falla en `make ci` (FR-073).
- **SC-007**: De N llamadas simultáneas a la misma herramienta y a herramientas distintas, N reciben su propio resultado, y el detector de carreras no encuentra ninguna. **Control**: el test falla en `make ci`, que corre con `-race` (FR-074).
- **SC-008**: Al cerrarse la entrada estándar, el servidor termina con 0; `mcp serve --asunto` sale con 2; `mcp serve --dry-run` termina con 0 mensajes respondidos y 0 bytes en la salida estándar; y, lanzado con `/` como directorio de trabajo desde una ruta con espacios, da el mismo sobre. **Control**: los e2e fallan en `make ci` (FR-075, FR-076).
- **SC-009**: `skills-check` sin drift; 0 herramientas nombradas en la tabla de una skill que no estén en el servidor y 0 órdenes que no estén en el registro; cada `SKILL.md` con menos de 300 líneas; y las comprobaciones de H7.4 en verde, con el calibrado en 36, 11 y 9. **Control**: `make ci` falla (`skills-check` falla con 300 líneas o más; FR-077).
- **SC-010**: Los 5 contenidos de las `instructions` están en sus primeros 512 caracteres. **Control**: un test falla en `make ci` (FR-077).
- **SC-011**: Los tests del informe distinguen los casos de FR-080: un umbral incumplido en un solo modo da `fallo`; 3 de 54 en un modo con 0 de 54 en el otro no se cumple; una serie que falla en un modo da `fallo`; y 0 series de `tasas` de los dos modos sin celda propia en la tabla que escribe `scripts/workflow/informe.sh`. **Control**: fallan en `make ci`.
- **SC-012**: Los tests de `Juzgar` distinguen los casos de FR-081: la herramienta satisface el `comando`, la de otro bloque no, la que falla con la norma y el bloque de la eval tampoco, la prohibida no pasa, y de las cuatro respuestas de la eval sin binario ni servidor pasa solo la que lleva la línea y ninguna cita. **Control**: fallan en `make ci`.
- **SC-013**: 0 importaciones del SDK de MCP fuera del paquete del servidor, y `make ci` en verde. El README dice las tres formas de declarar el servidor y lo que no es compatible; CONTRIBUTING y `CHANGELOG.md` (*Unreleased*) llevan lo de FR-061. **Control**: `depguard` falla en `make ci` (FR-079); la documentación la comprueba la revisión final contra FR-060 y FR-061.

## Uso, de fuera adentro

Cada salida del hito, desde quien la consume (criterio de uso, ADR 0028). Lo que gana la skill es pedir cada operación como herramienta y recibir el mismo sobre. Volumen de referencia: meses de uso diario, con cientos de normas y miles de bloques consultados, la mayoría hace más de una semana. Ninguna salida nueva crece con ese volumen: el servidor no añade ningún verbo ni cambia el `data` de ninguno, y cada sobre sigue acotado por lo que se pide.

Medidas de esta sesión, sobre `main` (`22b5bda`): los diez documentos de `--describe` de los verbos de consulta suman 46 184 bytes, con sangría, con las ocho banderas globales en cada entrada y con sus definiciones (de 3 391 bytes el de `boe buscar` a 7 779 el de `territorio resolver`); el sobre de `territorio resolver Leganés --json` ocupa 1 356 bytes; el de `boe indice BOE-A-2017-12902 --json`, leído sin red de la caché local, 34 720. Los sobres de `boe articulo` versionados en `web/src/data/sobres/` van de 1 463 a 8 210 bytes, y el del art. 21 de la Ley 39/2015 ocupa 4 433. No se ha medido en esta sesión qué parte del resultado —`structuredContent`, el bloque de texto o los dos— entrega cada agente a su modelo.

| Salida | Quién la pide, cuántas veces y qué hace con ella | Tamaño | Cuándo deja de darse cada señal |
|---|---|---|---|
| La lista de herramientas (FR-002, FR-003) | El agente, una vez por conexión con el servidor, para saber qué puede llamar. La skill no la pide. | 10 herramientas; como mucho los 46 184 bytes de sus documentos de `--describe`, menos las banderas globales. No crece con el uso: crece con los verbos, unos 4,6 KB por verbo nuevo. | No da señales. |
| El resultado de una llamada (FR-010) | La skill, las mismas veces que hoy la orden: en `boe-legislacion`, `boe_buscar` si la norma no está en su tabla, `boe_indice` si no conoce el id del bloque y, por cada bloque, `boe_articulo` y `graph_check`; en `legal-core`, `territorio_resolver` una vez por municipio. Lee `data`, los avisos y la procedencia, como hoy. | El sobre de la orden, dos veces en el mensaje (`structuredContent` y el bloque de texto). Para una pregunta por el art. 118 de la LCSP que necesita el índice: 34 720 del índice + 2 625 del bloque (`web/src/data/sobres/BOE-A-2017-12902-a1-30.json`) + ≈ 325 de `graph_check` sin cambios = ≈ 37 670 bytes de sobres en tres llamadas, ≈ 75 KB en los mensajes. Con la orden son los mismos 37 670. Con cientos de normas consultadas es lo mismo: lo acotan la norma y el bloque pedidos, no lo acumulado. | Cada llamada es independiente de la anterior. |
| `graph_check` con norma y bloques (FR-012), sin cambios en el verbo | La skill, una vez por cada lectura de bloques, detrás de ella (FR-033). Traslada cada `version-obsoleta` como una línea `⚠ REDACCIÓN MODIFICADA:`. | ≈ 325 bytes sin cambios y ≈ 1 000 por bloque cambiado; ≤ 3 800 con cinco (H7.1), dos veces en el mensaje. Sin argumentos, que la skill no pide nunca, ≤ 50 hallazgos y ≤ 40 000 bytes (H7.1 FR 013). | Cada señal se da una vez y la apaga la siguiente lectura del bloque (H7.1). |
| El error de herramienta con su `clase` (FR-011) | La skill, cuando una llamada falla: lee la `clase` y aplica lo que hoy aplica al código (FR-034). | El sobre de fallo: `clase` y `mensaje`, menos de 1 KB. | Deja de darse en la llamada siguiente que no falla. |
| La línea `⚠ SIN CONSULTA AL BOE:` (FR-035) | La persona que pregunta; una por respuesta cuando no hay herramienta ni binario. Le dice por qué no hay respuesta y dónde se arregla. | Una línea: la etiqueta, la causa y la dirección, del orden de 200 bytes. Una por respuesta, no una por norma. | Deja de darse en la primera pregunta tras instalar kitlegal o declarar el servidor. |
| Las `instructions` (FR-006) | El agente, una vez por conexión. Valen sobre todo donde hay servidor y no hay skill. | Texto fijo, con lo imprescindible en sus primeros 512 caracteres; no crece con el uso. | No da señales. |
| El aviso de versión distinta, en la salida de error (FR-023) | La persona que mira el registro de su agente; una vez por arranque del servidor. | Una línea, la de H19. | Deja de darse al reinstalar las skills con `kitlegal skills install` de esa versión. |
| El registro de eventos, en la salida de error (FR-023) | Quien depura; una línea por llamada, como hoy una por invocación. | Del orden de 150 bytes por llamada; no se guarda en ningún sitio del kit. | No es una señal para la skill ni para quien pregunta. |
| La tabla de comandos de cada `SKILL.md` (FR-030) | El modelo, una vez por conversación en que se activa la skill, para saber cómo se pide cada operación. | Una fila por operación con sus dos formas: 9 en `boe-legislacion` y 1 en `legal-core`. `SKILL.md` < 300 líneas. | No da señales. |
| `umbrales` del informe, por modo (FR-043) | El job, que decide con ellos el veredicto; el informe final del workflow, que los lee sin modelo; y la persona. Una vez por job y skill. | En `boe-legislacion`, 10 elementos de ≈ 250 bytes (cuatro que deciden y el del informativo, por dos modos); `[]` en `legal-core`. Tamaño fijo. | Cada job los mide de nuevo sobre su commit. Uno incumplido deja de darse en el primer job que lo cumple. |
| Las tasas y las sesiones del informe, por modo (FR-040, FR-042, FR-048) | `scripts/workflow/informe.sh`, una vez por run: con `tasas` y los recuentos de expresiones escribe la sección de evals del informe final. Y la persona, que lee ese informe final y, si quiere el detalle, el del job: ve si cada serie pasa en cada modo y qué pidió cada sesión, con órdenes o con herramientas. | En `boe-legislacion`, 66 series en `tasas`: por modo, 20 del modelo que decide y 12 del informativo, más 2 de la eval sin binario ni servidor; cada una con su celda en el informe final (FR-048). El doble de sesiones que hoy, más las de las evals sin binario ni servidor: lo fija el conjunto de evals, no el uso. | Cada job las mide de nuevo. |

## Fuera de alcance

Del hito, literal:

- «cualquier transporte de red y cualquier servidor de kitlegal (ADR 0027 y 0035);»
- «el `.mcpb`, el plugin y el marketplace (H22);»
- «que `skills install` escriba la configuración MCP de cada agente;»
- «`prompts` y `resources` MCP —servir las skills empotradas por el protocolo— mientras ningún agente verificado los cargue sin que la persona los elija;»
- «herramientas para `skills`;»
- «figurar en directorios de servidores MCP (§0);»
- «y la web, que la cambia la persona (ADR 0024).»

De lo que el hito no especifica (constitución, «Criterio de decisión autónoma», punto 2):

- Cambiar los verbos de hoy: sus argumentos, su `data`, sus sobres, sus códigos de salida o su `--describe`. El servidor los usa como están.
- Herramientas que escriban, y cómo anunciaría el servidor un verbo que no fuera de solo lectura: llega con el primer verbo que lo sea.
- Parámetros de herramienta para `--offline`, `--no-graph`, `--timeout` o cualquier otra bandera global, y cualquier acceso al asunto desde el servidor.
- Otras piezas del protocolo que las herramientas y las `instructions`: notificaciones de cambio de la lista, progreso, registro por el protocolo, peticiones del servidor al cliente.
- Otro cambio en las skills que nombrar cada operación de las dos formas y la regla de FR-035: su protocolo, sus reglas, sus `description`, la cita, los avisos, `⚠ REDACCIÓN MODIFICADA:` y la frase de la regla 7 se quedan.
- Comprobar en la eval sin binario ni servidor algo más que la línea y que no hay citas: por ejemplo, que la respuesta de `legal-core` no da datos de territorio de memoria, o juzgar con un modelo si la respuesta afirma contenido de la norma.
- Umbrales nuevos, umbrales para `legal-core`, o que decida el del modelo informativo; cambiar el modelo que decide, las repeticiones o la regla por serie (ADR 0016 y 0031).
- La prueba de red en el modo herramienta, y medir cualquier modo en Windows o con otro agente que Claude Code.
- Cambiar el workflow `hito`, `scripts/workflow/` o el informe final (FR-048).
- Sondear el modo herramienta con `make evals-sondeo`, sondear las dos evals sin binario ni servidor de FR-046 y cualquier argumento nuevo del sondeo, incluido uno de modo (FR-050). Esas evals y ese modo se miden solo en el job.
- Evals nuevas distintas de las dos de FR-046, y cambiar las preguntas, los comandos esperados o el juicio de las que hay, más allá de que una herramienta satisfaga un `comando` (FR-042).
- Editar el spec, el plan, los guiones o las grabaciones de los hitos anteriores, la release y un ADR nuevo.

## Assumptions

- **Qué es «no tener la herramienta»**. La skill no puede preguntar al servidor si existe: lo que ve es si entre las herramientas del agente hay una con el nombre `<applet>_<verbo>` de kitlegal. Cada agente las presenta a su manera —Claude Code, como `mcp__<servidor>__<herramienta>`—, y por eso la skill y el juicio de las evals las reconocen por ese nombre, con prefijo o sin él (FR-042). El plan fija cómo lo dice `SKILL.md`.
- **La regla de FR-035 cubre también al agente sin shell.** El hito la enuncia como «la orden falla porque el binario no está», y la Entrega, como «si no tienen ninguna de las dos». Un agente que no puede ejecutar ninguna orden —el chat de Claude en la web— está en el mismo caso: no tiene la herramienta ni puede usar la orden. El ADR 0035 lo dice así («donde hay skill y no hay servidor»). Las evals miden el caso del hito: sin binario ni servidor, con shell.
- **La misma etiqueta en las dos skills.** El hito pide la línea `⚠ SIN CONSULTA AL BOE:` en «los dos `SKILL.md`» y «una eval nueva por skill» que la comprueba, así que `legal-core` lleva esa etiqueta tal cual, aunque lo que no ha podido pedir sea `territorio resolver`.
- **Las evals sin binario ni servidor no son de ningún modo.** El hito define dos modos y, aparte, «una eval nueva por skill, sin binario ni servidor». Los umbrales «que hoy deciden» se miden «en cada modo», así que las sesiones de esas dos evals no entran en ellos (FR-047), y los totales por modo siguen siendo los de H7.4: 54 respuestas del modelo que decide y 30 del informativo. Sus series las mide también el modelo informativo, como las de cualquier eval que decide (ADR 0016).
- **La entrega al grafo y la llamada siguiente.** El hito dice «entrega al grafo después del resultado», que en la orden es «tras presentar la salida» y antes de que el proceso termine, de modo que `&&` ve lo entregado. En un servidor no hay fin de proceso entre dos llamadas, así que FR-013 pide lo observable: el resultado no cambia por la entrega, y la llamada que el cliente envía tras recibir un resultado ve lo que esa llamada entregó. Cómo se consigue es del plan.
- **`mcp serve` no imprime un sobre propio al terminar.** El principio II («todo applet emite el sobre») se cumple en cada llamada, dentro del protocolo: la salida estándar es solo protocolo (FR-023). Un fallo anterior a servir —el de `--asunto`— se presenta como cualquier fallo del kernel.
- **Las banderas que el hito no nombra** (`--verbose`, `--describe`, `--dry-run`, `--json`) valen lo que en cualquier verbo (FR-022). Para `--verbose`, `--describe` y `--json` es lo que el kernel ya hace con todo verbo del registro, sin comportamiento nuevo. Para `--dry-run` no: el kernel llama al applet con la bandera y es cada capa con efectos la que la honra, así que no servir con `--dry-run` lo hace el applet `mcp` y lo comprueba FR-075.
- **El informe y el cierre.** El cierre del workflow recoge un informe por skill, por el nombre de comprobación `evals (<skill>)`, y espera como mucho 3 h (`scripts/workflow/cierre.sh`). Cómo caben los dos modos en eso —en un trabajo o en varios, con qué concurrencia— es del plan, sin tocar `scripts/workflow/` (FR-048). El informe lo lee también `scripts/workflow/informe.sh`, que agrupa `tasas` por eval, enseña de cada fila la primera serie de cada modelo y da una fila de expresiones prohibidas por modelo: con dos series de la misma eval y el mismo modelo enseñaría un modo y callaría el otro, y por eso FR-048 pide que el informe del job distinga los modos en lo que ese guion lee. En el cierre de H7.4, el modo orden de `boe-legislacion` midió 452 s con 96 sesiones.
- **`boe-legislacion` tiene hoy 298 líneas** (`wc -l` en esta sesión), a dos del límite. Decir cada paso de las dos formas exige reescribir, no añadir. La tabla generada gana una columna, no filas.
- **Las versiones de las skills.** El hito no fija números. `CHANGELOG.md` nombra el cambio de cada skill por lo que cambia para quien la usa (FR-061); si lleva número, es el siguiente de su serie, como en cada hito desde H7.1.
- **La dependencia** está prevista en el principio V «en la fase de distribución» y la adelanta el ADR 0035. La lista de módulos que enlaza el binario (`internal/arch_test.go`) gana el SDK y lo que arrastre, con su justificación en el plan (Complexity Tracking), como pide la constitución.
- **La dirección `https://kitlegal.es/instalar/`** existe en la web (`web/src/pages/instalar/`). Su contenido lo cambia la persona (ADR 0024).
- **Lo que dice el ADR 0035 de cada agente** —qué admite Codex, Antigravity y la app de Claude, qué versión del protocolo negocia, desde dónde arranca el servidor— se toma como está: es lo leído y lo probado a mano el 2026-10-01, y este hito no lo vuelve a verificar. La prueba humana de SC-002 lo confirma después de fusionar.
- **Los nombres técnicos que aparecen** (`structuredContent`, `inputSchema`, `outputSchema`, `readOnlyHint`, `instructions`, `Juzgar`, `TestDefinicionDelJob`, `skills-check`, `depguard`) son los del hito o existen en el repositorio. Quedan para el plan: el paquete del servidor, el texto de las `instructions`, la forma exacta de la línea `⚠ SIN CONSULTA AL BOE:`, la forma de la tabla generada, el nombre de cada umbral con su modo, cómo declara el formato de eval una eval sin binario ni servidor, cómo se declara el servidor a la sesión y cómo se lee de ella una llamada a una herramienta.
