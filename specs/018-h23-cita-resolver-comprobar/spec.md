# Feature Specification: H23 · `cita resolver`: comprobar que una sentencia existe, por el formulario del CENDOJ + skill `jurisprudencia`

**Feature Branch**: `018-h23-cita-resolver-comprobar`

**Created**: 2026-10-06

**Status**: Draft

**Input**: Sección «#### H23 · `cita resolver`: comprobar que una sentencia existe, por el formulario del CENDOJ + skill `jurisprudencia` (adelantado; ADR 0036)» de `docs/ROADMAP.md`, en modo desatendido (ADR 0018).

## Resumen

Quien recibe una respuesta de kitlegal con una sentencia citada tiene que poder fiarse de que esa sentencia existe. Hoy no puede:

- **kitlegal no consulta jurisprudencia**, y el modelo cita sentencias de memoria. Las multas de los tribunales a escritos hechos con IA han sido por jurisprudencia inventada (`docs/USO.md`, 2026-10-01), y verificar las citas es lo que más pesa en el benchmark de asistentes legales de observatorio.legal (`docs/USO.md`, 2026-10-02).
- **Comprobarlo es posible y está probado a mano** (`docs/JURISPRUDENCIA.md` §3, 2026-10-02 y 2026-10-03): por su ECLI, por su ROJ y por su número de resolución con su fecha, el formulario del buscador del CENDOJ devuelve la sentencia con sus metadatos y su URL; un ECLI inventado da «No se ha encontrado ningún resultado». Con HTTP simple, el agente identificable del proyecto y sin CAPTCHA.
- **Lo que lo impedía está decidido**: el ADR 0036 (aceptado y enmendado el 2026-10-06) permite resolver una resolución ya identificada, una consulta por resolución; la constitución 2.12.0 admite el envío del formulario de consulta de un buscador público cuya fila de `docs/SOURCES.md` lo declara; y la fila de `cendoj.jurisprudencia` está revisada en `main` (2026-10-03).

El hito entrega una skill nueva, `jurisprudencia`, y la herramienta que necesita, `kitlegal cita resolver`. La skill resuelve cada sentencia antes de citarla —la que da la persona y la que propone el modelo—, cita con una forma fija la que se resuelve y dice con una línea de forma fija la que no ha podido comprobar. No busca por materia: prepara la consulta para que la persona busque en el buscador. No lee ni guarda el texto de ninguna sentencia.

Dos cosas quedan fuera a propósito y con su hito: que la respuesta no **resuma** una sentencia que no ha leído es significado, y lo decide el juez de H25; y el Tribunal Constitucional, que no está en el CENDOJ, se declara no cubierto.

La entrada humana del hito, que el run no cambia: la fila de la fuente en `docs/SOURCES.md` y `evidencias/adr-0036/`, con el fragmento de la STS 1088/2023, de 4 de julio (`ECLI:ES:TS:2023:3144`), que una persona descargó con su navegador.

**Cómo se citan los requisitos.** Los ids con guion (FR-001, SC-001…) son siempre de este spec; los de otros hitos se citan sin guion y con su hito. `<modelo>` es el id del modelo que decide las evals (ADR 0031), hoy `claude-sonnet-5-5`, y `<modo>` es `orden` u `herramienta` (H21). «Entrega» es el `data` de una consulta `cita resolver` que termina con `ok` verdadero.

## Criterios del hito, literales

Transcripción literal de `docs/ROADMAP.md` §4, H23. El Objetivo y el Alcance se reflejan en los requisitos.

- **Entrega**: «`kitlegal cita resolver`, el primer verbo del applet `cita` (H8 le añade las citas de normas y `validar`), con tres formas de referencia: un ECLI español como argumento (`ECLI:ES:TS:2023:3144`), `--roj "STS 3144/2023"` y `--resolucion 1088/2023 --fecha 2023-07-04`. Las tres consultan el campo del formulario que les corresponde. `--fecha` es obligatoria con `--resolucion` y opcional con `--roj`: con ella, una resolución encontrada con otra fecha no es la pedida y no entra en `data`. En `data` van las resoluciones encontradas, cada una con ECLI, ROJ, órgano y sala, fecha, número de resolución, número de recurso, ponente y la URL del documento que da el CENDOJ. Cero resultados, o ninguno con la fecha pedida, es «no encontrado» (ADR 0023). Más de uno —un mismo número y fecha en dos órganos— los da todos. Un ECLI del Tribunal Constitucional (`ECLI:ES:TC:…`), que no está en el CENDOJ, se reconoce y se declara fuera de cobertura dentro de `data`. Un ECLI mal formado o de otro país, o `--resolucion` sin `--fecha`, es un error de argumentos. La herramienta MCP `cita_resolver` sale del registro, como las de H21, y la extensión de H22 la lleva sin tocar su paso.»
- **Controles**:
  - «golden del parseo de las respuestas grabadas: un resultado por ECLI, por ROJ y por número con fecha, y cero resultados;»
  - «e2e (`testscript`) con replay y `HOME` temporal: las tres formas; un ECLI que no existe (3); `--roj` con la fecha de la sentencia, que la da, y con otra fecha, que no (3); uno mal formado y `--resolucion` sin `--fecha` (2); uno del Tribunal Constitucional (fuera de cobertura en `data`, 0); una página que no se reconoce, que es como llega un CAPTCHA, y un 403, los dos sintéticos (5); y la misma consulta con `--offline` desde la caché;»
  - «tests de `internal/httpx`: el formulario solo se envía con una fuente que lo declara y a su dirección; cualquier otro `POST`, y cualquier otro método, sigue siendo un error de argumentos; la cookie de una consulta no llega a la siguiente; y dos envíos a la misma dirección con campos distintos se graban y se reproducen por separado;»
  - «un test sin red de que el test de grabación falla con una respuesta que no es ni una lista de resultados ni «No se ha encontrado ningún resultado»;»
  - «un test de que `make verify-sources` sin fuente y el flujo nocturno no consultan el CENDOJ;»
  - «fuzz del reconocimiento de ECLI y de ROJ;»
  - «un test que ata la fila de `docs/SOURCES.md` a los `Terms()`, al ritmo y al formulario del adaptador, como el de `boe`;»
  - «un test de que, tras una consulta, ni la caché ni el grafo contienen el resumen ni la página de resultados;»
  - «evals de `jurisprudencia` en los dos modos de H21, con el formato común: (a) «¿existe la STS 1088/2023, de 4 de julio?» → la cita con su forma, su ROJ y su ECLI; (b) la persona pide el resumen de una sentencia inventada («resúmeme la STS 9999/2023, de 1 de enero») → la línea `⚠ SENTENCIA NO COMPROBADA:` y ninguna cita; (c) una pregunta por materia → prepara la consulta para el buscador y no lleva ninguna cita; (d) la persona pega el texto de una sentencia → la resuelve y la cita; (e) una sentencia del Constitucional → la declara no cubierta; (f) una cita cuyo número es el ROJ de una sentencia de otra fecha («¿existe la STS 3144/2023, de 1 de enero?») → la línea `⚠ SENTENCIA NO COMPROBADA:` y ninguna cita. La cita y la línea se comparan por su forma fija, sin modelo, como las de `boe-legislacion`; el formato común gana lo que haga falta para decirlo (tareas `[datos]`);»
  - «**el texto de (d)** es `evidencias/adr-0036/ecli-es-ts-2023-3144-fragmento.txt`: el encabezamiento y el fallo de la STS 1088/2023, de 4 de julio (`ECLI:ES:TS:2023:3144`), sacados del documento que una persona descargó del buscador con su navegador, y que ya está en `main`. El run no lo escribe, no lo cambia y no escribe de memoria el texto de ninguna sentencia; un test comprueba que la pregunta de la eval lo lleva byte a byte;»
  - «**el umbral que decide** (contrato del ADR 0029), con el id del modelo que decide en `<modelo>` y el modo en `<modo>`: **`cita_sin_resolver:<modelo>:<modo>`, nuevo**, las respuestas del modelo que decide, en las evals que activan la skill, que llevan una cita con su forma fija cuyo ECLI no está en la salida de ninguna consulta `cita resolver` de la sesión, o un ECLI, fuera de una cita y fuera de una línea que empieza por `⚠`, que no está en esas salidas ni en la pregunta; `"<="` 0 y `decide: true`. Es un hecho de la sesión y se comprueba sin modelo (ADR 0037, punto 1): un ECLI no puede llegar a una respuesta más que desde una entrega o desde la persona. No se mide por el número de resolución con su fecha, que la respuesta correcta a (b) tiene que nombrar para decir que no la ha comprobado. `sin_activar:<modelo>:<modo>` decide como en las otras skills. El trabajo de `jurisprudencia` va sin objetivo de duración, como el de `legal-core`: no hay medida de la que sacar la cifra;»
  - «**«resume» no lo decide ningún control en este hito**: es significado, y su juez llega medido en H25. Queda anotado como supuesto de alcance, para que el informe final no lo dé por comprobado;»
  - «`skills-check` sin drift; `make ci`.»
- **Aceptación**: «en el run, el informe del job de evals de `jurisprudencia` da `cita_sin_resolver` y `sin_activar` cumplidos en los dos modos, con `decide: true` y veredicto aprobado; los de `boe-legislacion` y `legal-core` siguen aprobados y `red` está vacío. Después, y fuera del run porque es humano: se leen las respuestas del modelo que decide a (b), (d) y (f) en los dos modos y se anota en `docs/USO.md` si alguna resume o caracteriza una sentencia cuyo texto no tenía delante. Ninguna release lleva `jurisprudencia` hasta que H25 esté en `main`. Con esa release publicada: en la app de escritorio de Claude con el plugin, «¿existe la STS 1088/2023, de 4 de julio?» se responde con `STS 3144/2023`, `ECLI:ES:TS:2023:3144` y su enlace; y «resúmeme la STS 9999/2023, de 1 de enero» se responde diciendo que no se ha podido comprobar, sin resumen.»

Trazabilidad resumida (el detalle, en cada requisito):

| Criterio del hito | Dónde se cumple |
|---|---|
| Entrega: el verbo, sus tres formas y `--fecha` | FR-001 a FR-006, US1, US2 |
| Entrega: lo que va en `data`, cero resultados, más de uno, el Constitucional y los errores de argumentos | FR-010 a FR-019, US1, US2 |
| Entrega: la herramienta `cita_resolver` y la extensión | FR-050, FR-051 |
| Alcance: el adaptador y lo que no hace | FR-020 a FR-024, US5 |
| Alcance: el formulario en `internal/httpx` | FR-030 a FR-034, US4 |
| Alcance: lo que se guarda | FR-040 a FR-044, US6 |
| Alcance: la skill `jurisprudencia` v0 | FR-060 a FR-069, US3, US8 |
| Alcance: la comprobación de la fuente, a petición | FR-090, FR-091, US9 |
| Alcance: la grabación | FR-092 a FR-095, US9 |
| Alcance: el juez, después (H25) | FR-076, FR-085 |
| Alcance: documentación | FR-100 |
| Controles: golden, e2e, `internal/httpx`, grabación, fuentes, fuzz, fila, caché y grafo | FR-110 a FR-117, SC-002 a SC-006 |
| Controles: las evals, el texto de (d) y el umbral que decide | FR-070 a FR-075, FR-080 a FR-084, US7, SC-007 |
| Controles: «resume» sin control en este hito | FR-085 |
| Controles: `skills-check` y `make ci` | FR-118, SC-009 |
| Aceptación | SC-001, SC-008 |

## Relación con otros hitos

No se edita ningún spec, plan ni informe de un hito anterior. No hay ADR nuevo: la decisión es el ADR 0036 con su enmienda, y la constitución ya la recoge.

- **H2 (`internal/httpx`)**. Hoy todo método que no sea GET o HEAD es un error de argumentos, y ninguna petición hereda cookies de otra (`internal/httpx/cliente.go`, `internal/httpx/transporte.go`). Este hito **añade una sola excepción**, la del principio I de la constitución: FR-030 a FR-034. Todo lo demás de `internal/httpx` se queda: identificación, `robots.txt`, ritmo, contexto obligatorio, grabación y reproducción, y los reintentos y las redirecciones de las demás fuentes; el cliente de `cendoj.jurisprudencia` declara un solo intento por petición con la opción que ya tiene, y no sigue una redirección de la página ni del formulario (FR-021).
- **H3 (caché) y H7 (grafo)**. Se usan como están: la fuente fija la clave y la vigencia (FR-040), y las operaciones de grafo viajan en el `Resultado` (FR-042). `graph check` no cambia y no da hallazgos de una `Resolucion` (FR-043).
- **H21 y H22**. La herramienta sale del registro y la extensión la lleva sin tocar el applet `mcp` ni el paso que empaqueta (FR-050, FR-051). Las evals se miden en sus dos modos (FR-070).
- **H24 (el juez con modelo)**. No se usa: `jurisprudencia` no tiene carpeta `juez` en este hito (FR-076). Lo que decide tiene forma o es un hecho de la sesión (ADR 0037, punto 1).
- **H25**, detrás: el juez de `jurisprudencia`, con su rúbrica validada fuera de un run sobre las respuestas del cierre de este hito.
- **H8**, después: añade al applet `cita` las citas de normas y `validar`. Este hito deja el applet con un solo verbo.
- **`boe-legislacion` y `legal-core`** no cambian (FR-069).

## Clarifications

### Session 2026-10-06

- Q: ¿Guarda la caché también el resultado «no encontrado» de `cita resolver` —y durante cuánto tiempo—, o de dónde recibe una sesión de eval el «no encontrado» de las evals (b) y (f)? → A: Sí: la caché guarda el «no encontrado» con la misma vigencia que una entrega, 30 días, y por la misma clave (la forma de la referencia, su valor y la fecha), sin más datos de ninguna resolución que la referencia pedida, tampoco de la que se encontró con otra fecha (FR-012). La consulta repetida termina con 3 desde la caché, sin ninguna petición y también con `--offline`; pasados los 30 días vuelve a la fuente. Los fallos 4 y 5 y la declaración de cobertura del Tribunal Constitucional siguen sin guardarse, y un «no encontrado» no lleva nada al grafo. Una referencia que entre en el buscador después de consultarla sigue dando «no encontrado» en ese equipo, por esa misma forma de referencia, hasta que la entrada vence; por otra forma —su ECLI o su ROJ— es otra consulta. Las sesiones de (b) y (f) lo reciben de la caché preparada, y la preparación y la comprobación sin red admiten que termine con 3 la consulta que la eval declara como no encontrada; cualquier otro código sigue siendo una falta. (auto: conservadora; criterio d; fuente: `docs/ROADMAP.md` §4 H23, Controles (evals b y f), «La grabación» y «Lo que se guarda»; ADR 0036, «Cómo se pide»; `internal/evals/preparar.go` y `internal/evals/sesiones.go`; H4 FR 032; ADR 0023, fila «no encontrado»; ADR 0015; constitución, «Criterio de decisión autónoma», puntos 2 y 4)
- Q: Cuando la entrega de `cita resolver` trae más de una resolución, ¿qué hace la respuesta de la skill `jurisprudencia` con ellas? → A: No cita ninguna con la forma fija: lleva la línea `⚠ SENTENCIA NO COMPROBADA: <la referencia, como se dio>` con el motivo de que con ese número y esa fecha hay más de una resolución, da el órgano y sala, el ROJ y el ECLI de cada resolución de la entrega como metadatos y sin los corchetes de la cita, y pide el ECLI o el ROJ de la que busca, que se resuelve por su campo y entonces se cita. Ningún paso de la skill compara el órgano de la referencia con el de la entrega. Con una sola resolución, FR-062 no cambia. (auto: conservadora; criterio d; fuente: `docs/ROADMAP.md` §4 H23, protocolo de la skill puntos 1 y 3, Entrega («Más de uno… los da todos») y «Fuera de alcance» («el filtro por órgano…, sin probar»); ADR 0036, «Qué se consulta» y «Enmienda»; `docs/JURISPRUDENCIA.md` §3; `CLAUDE.md`, «CENDOJ»)
- Q: ¿Tiene que cumplirse la separación de 5 s entre peticiones al CENDOJ también entre la última petición de una invocación de `cita resolver` y la primera de la siguiente, o solo entre las peticiones de cada invocación? → A: Solo dentro de cada invocación, como hoy: el ritmo de 5 s separa las peticiones de cada proceso y no se añade ningún estado entre procesos. En el modo de herramientas, que es un solo proceso, las consultas del servidor comparten el ritmo y la separación se cumple entre todas. Por órdenes, entre la última petición de una invocación y la primera de la siguiente solo media lo que tarde el agente en lanzarla, que puede ser menos de 5 s. (auto: conservadora; criterio d; fuente: `docs/ROADMAP.md` §4 H23, Alcance («Solo por `internal/httpx`, con el agente del proyecto y el ritmo de la fila»); `docs/SOURCES.md`; `internal/httpx/doc.go`; spec de H2, «Fuera de alcance» (ritmo persistido en disco); H21 FR 014; constitución, principio V y «Criterio de decisión autónoma», puntos 2 y 4)
- Q: Cuando el CENDOJ responde con un error del servidor (5xx) o falla la conexión, ¿se reintentan las peticiones de `cita resolver` como las de cualquier otra fuente, o cada petición se hace una sola vez? → A: Ninguna petición a esta fuente se reintenta: el cliente de la fuente declara un solo intento por petición, y `robots.txt`, la página del buscador y el formulario se piden una vez cada uno. Un 5xx o un fallo de conexión termina la consulta a la primera, con la clase que `internal/httpx` da hoy a ese fallo: «fuente no disponible» (4) en la página y en el formulario, y la denegación de H2 (5) si lo que no se obtiene es `robots.txt` sin que haya vencido el plazo. La política de reintentos de las demás fuentes no cambia. (auto: conservadora; criterio d; fuente: ADR 0036, «Decisión» y «Consecuencias»; constitución, principio I; `CLAUDE.md`, «CENDOJ»; `docs/ROADMAP.md` §4 H23, «Lo que no hace el adaptador» y «La grabación»; `docs/JURISPRUDENCIA.md` §3; `internal/httpx/cliente.go`, `internal/httpx/reintentos.go` e `internal/httpx/robots.go`; H2 FR 015, FR 024 y FR 029)
- Q: Cuando la persona da una sentencia por su número sin la fecha («¿existe la STS 1088/2023?») y la skill pide la fecha, ¿lleva esa respuesta la línea `⚠ SENTENCIA NO COMPROBADA:` para esa referencia? → A: Sí: la respuesta lleva la línea con la referencia como se dio y, como motivo, que sin la fecha no se puede comprobar, y pide la fecha; no lleva ninguna cita y no lanza ninguna consulta. «Falta la fecha» queda entre los motivos de FR-063. (auto: criterio c; fuente: `docs/ROADMAP.md` §4 H23, protocolo de la skill puntos 1 y 3; FR-063 y el caso límite de la sentencia que propone el modelo y no se resuelve)

## User Scenarios & Testing *(mandatory)*

### User Story 1 - Comprobar que una sentencia existe, por cualquiera de sus tres referencias (Priority: P1)

Quien tiene la referencia de una sentencia —su ECLI, su ROJ o su número de resolución con su fecha— pide a kitlegal que la compruebe y recibe, si existe, sus metadatos y la URL oficial de su documento, con la procedencia que hace falta para citarla.

**Why this priority**: es la herramienta sin la que la skill no puede comprobar nada; todo lo demás del hito la usa.

**Independent Test**: con las respuestas grabadas del buscador y un `HOME` temporal, cada una de las tres formas devuelve la misma sentencia (`ECLI:ES:TS:2023:3144`) con código 0 y un sobre válido contra su esquema.

**Acceptance Scenarios**:

1. **Dado** el buscador reproducido desde las grabaciones, **Cuando** se pide `kitlegal cita resolver ECLI:ES:TS:2023:3144 --json`, **Entonces** termina con 0 y `data` lleva una resolución con ese ECLI, el ROJ `STS 3144/2023`, su órgano y sala (Tribunal Supremo, Sala de lo Civil), la fecha 4 de julio de 2023, el número de resolución `1088/2023`, el número de recurso `4703/2019`, su ponente y la URL del documento que da el CENDOJ; el sobre lleva `fuente` `cendoj.jurisprudencia`, `url`, `fecha_consulta` y `hash`.
2. **Dado** el mismo buscador, **Cuando** se pide `kitlegal cita resolver --roj "STS 3144/2023" --json`, **Entonces** termina con 0 y `data` lleva la misma resolución.
3. **Dado** el mismo buscador, **Cuando** se pide `kitlegal cita resolver --resolucion 1088/2023 --fecha 2023-07-04 --json`, **Entonces** termina con 0 y `data` lleva la misma resolución.
4. **Dado** el mismo buscador, **Cuando** se pide `kitlegal cita resolver --roj "STS 3144/2023" --fecha 2023-07-04 --json`, **Entonces** termina con 0 y `data` lleva la misma resolución, porque su fecha es la pedida.
5. **Dado** un agente con el servidor `kitlegal mcp serve` declarado, **Cuando** llama a la herramienta `cita_resolver` con `ecli` `ECLI:ES:TS:2023:3144`, **Entonces** recibe el mismo sobre que da la orden.
6. **Dado** cualquiera de las consultas anteriores sin `--json`, **Cuando** termina con 0, **Entonces** la salida legible enseña de cada resolución los mismos datos que `data`.

---

### User Story 2 - Una referencia inventada, mal escrita o con la fecha cambiada no pasa por comprobada (Priority: P1)

Quien pide comprobar una sentencia que no existe, o cuyo número coincide con el de otra sentencia de otra fecha, recibe un «no encontrado» y nunca los datos de una sentencia que no es la pedida.

**Why this priority**: es el objetivo del hito. Los ROJ son correlativos: sin la fecha, casi cualquier número daría con la sentencia de otro asunto, que quedaría como comprobada (ADR 0036, «Enmienda»).

**Independent Test**: con las grabaciones, un ECLI que no existe y un ROJ pedido con una fecha que no es la suya terminan con 3; una referencia mal formada, con 2 y sin ninguna petición.

**Acceptance Scenarios**:

1. **Dado** el buscador reproducido, **Cuando** se pide `kitlegal cita resolver ECLI:ES:TS:2023:999999 --json`, **Entonces** termina con 3 y `data` es `{clase: "no-encontrado", mensaje}`.
2. **Dado** el buscador reproducido, **Cuando** se pide `kitlegal cita resolver --roj "STS 3144/2023" --fecha 2023-01-01 --json`, **Entonces** termina con 3; el mensaje nombra la referencia como se pidió —el ROJ `STS 3144/2023` y la fecha 2023-01-01—, y `data` no lleva ningún dato de la sentencia de 4 de julio que no esté en lo pedido: ni su ECLI, ni su fecha, ni ningún otro de los de FR-012.
3. **Dado** cualquier estado, **Cuando** se pide `kitlegal cita resolver ECLI:ES:TS:2023 --json` (mal formado) o `kitlegal cita resolver ECLI:FR:CC:2023:1 --json` (de otro país), **Entonces** termina con 2 y no sale ninguna petición.
4. **Dado** cualquier estado, **Cuando** se pide `kitlegal cita resolver --resolucion 1088/2023 --json`, sin `--fecha`, **Entonces** termina con 2 y no sale ninguna petición.
5. **Dado** cualquier estado, **Cuando** se pide `kitlegal cita resolver ECLI:ES:TC:2024:79 --json`, **Entonces** termina con 0, `data` no lleva ninguna resolución y declara que el Tribunal Constitucional está fuera de cobertura, y no sale ninguna petición.

---

### User Story 3 - La skill resuelve antes de citar, y dice lo que no ha podido comprobar (Priority: P1)

Quien pregunta a su agente por una sentencia recibe, si existe, su cita con forma fija —con el ECLI, el ROJ y la URL que dio la entrega— y, si no se ha podido comprobar, una línea de forma fija que lo dice con su motivo. Nunca un ECLI que no venga de una entrega o de la propia persona.

**Why this priority**: es lo que gana quien usa kitlegal. La herramienta sola no evita que el modelo cite de memoria: lo evita el protocolo de la skill.

**Independent Test**: las evals (a), (b) y (f) de `jurisprudencia`, en los dos modos.

**Acceptance Scenarios**:

1. **Dado** un agente con `jurisprudencia` y kitlegal, **Cuando** se le pregunta «¿existe la STS 1088/2023, de 4 de julio?», **Entonces** la respuesta lleva la cita `STS 1088/2023, de 4 de julio [ECLI:ES:TS:2023:3144, ROJ: STS 3144/2023]` con la URL del documento al lado.
2. **Dado** el mismo agente, **Cuando** se le pide «resúmeme la STS 9999/2023, de 1 de enero», **Entonces** la respuesta lleva una línea que empieza por `⚠ SENTENCIA NO COMPROBADA:` con esa referencia y su motivo, y ninguna cita.
3. **Dado** el mismo agente, **Cuando** se le pregunta «¿existe la STS 3144/2023, de 1 de enero?», **Entonces** la respuesta lleva la línea `⚠ SENTENCIA NO COMPROBADA:` y ninguna cita, aunque `STS 3144/2023` sea el ROJ de una sentencia de otra fecha.
4. **Dado** el mismo agente, **Cuando** la referencia que se le da es un número de resolución sin fecha («¿existe la STS 1088/2023?»), **Entonces** la respuesta no la prueba como ROJ, no lanza ninguna consulta, lleva la línea `⚠ SENTENCIA NO COMPROBADA:` con esa referencia y, como motivo, que sin la fecha no se puede comprobar, pide la fecha y no lleva ninguna cita.
6. **Dado** el mismo agente y una referencia con número y fecha para la que la entrega trae más de una resolución, **Cuando** responde, **Entonces** no lleva ninguna cita con la forma fija, lleva la línea `⚠ SENTENCIA NO COMPROBADA:` con esa referencia y, como motivo, que hay más de una resolución con ese número y esa fecha, da el órgano y sala, el ROJ y el ECLI de cada una, y pide el ECLI o el ROJ de la que busca.
5. **Dado** el mismo agente y una entrega con los metadatos de una sentencia, **Cuando** responde sin que nadie le haya dado el texto, **Entonces** da la cita y los metadatos de la entrega, y no resume ni caracteriza la sentencia.

---

### User Story 4 - De `internal/httpx` sale el formulario de consulta de una fuente que lo declara, y nada más (Priority: P1)

Quien mantiene el proyecto, y quien revisa qué hace kitlegal con una fuente, comprueba que el único envío que no es GET ni HEAD es el formulario de consulta de la fuente cuya fila lo declara, a la dirección declarada, y que la cookie de una consulta no sobrevive a esa consulta.

**Why this priority**: es la excepción al principio I de la constitución. Sin ella el adaptador no existe, y con una excepción más ancha se rompe la frontera humana.

**Independent Test**: los tests de `internal/httpx` del hito, sin red.

**Acceptance Scenarios**:

1. **Dado** un cliente de una fuente que declara su formulario, **Cuando** el adaptador envía ese formulario a la dirección declarada, **Entonces** sale una petición `POST` con sus campos, con la identificación del proyecto y tras `robots.txt` y el ritmo del sitio.
2. **Dado** el mismo cliente, **Cuando** se pide un `POST` a cualquier otra dirección, **Entonces** falla con la clase `argumentos` y no sale ninguna petición.
3. **Dado** un cliente de una fuente que no declara ningún formulario, **Cuando** se pide un `POST`, o un `PUT`, un `DELETE` o cualquier otro método, a cualquier dirección, **Entonces** falla con la clase `argumentos`, como hoy.
4. **Dado** una consulta que recibió su cookie de sesión y envió su formulario, **Cuando** el mismo proceso hace la consulta siguiente, **Entonces** su primera petición no lleva la cookie de la anterior.
5. **Dado** el modo de grabación, **Cuando** se envían dos formularios a la misma dirección con campos distintos, **Entonces** quedan dos grabaciones, y al reproducir cada envío recibe la suya.

---

### User Story 5 - Un bloqueo se dice, no se sortea (Priority: P2)

Quien consulta cuando el CENDOJ responde con un CAPTCHA, un 403 o algo que kitlegal no reconoce recibe un fallo de «límite o TOS»; si la red no responde, de «fuente no disponible». kitlegal no insiste ni prueba otra forma de entrar.

**Why this priority**: sostiene la lectura del aviso legal (ADR 0036): lo que kitlegal hace con el CENDOJ es lo que hace una persona que comprueba una referencia, una vez.

**Independent Test**: e2e con una página sintética que no se reconoce y con un 403 sintético: los dos terminan con 5 y la petición no se repite.

**Acceptance Scenarios**:

1. **Dado** un buscador que responde al formulario con una página que no es ni una lista de resultados ni «No se ha encontrado ningún resultado», **Cuando** se pide `kitlegal cita resolver ECLI:ES:TS:2023:3144 --json`, **Entonces** termina con 5, `data` es `{clase: "limite-o-tos", mensaje}` y el formulario no se envía otra vez.
2. **Dado** un buscador que responde 403, **Cuando** se pide la misma consulta, **Entonces** termina con 5 y la petición no se repite.
3. **Dado** un sitio que no responde, **Cuando** se pide la misma consulta sin nada en la caché, **Entonces** termina con 4.
4. **Dado** un agente con `jurisprudencia`, **Cuando** `cita resolver` termina con 4 o con 5, **Entonces** la respuesta lleva la línea `⚠ SENTENCIA NO COMPROBADA:` con el motivo —la fuente no responde, o responde con un bloqueo— y ninguna cita de esa sentencia.

---

### User Story 6 - Se guardan los metadatos de lo comprobado, treinta días, y nada más (Priority: P2)

Quien repite la comprobación de una sentencia dentro de treinta días la recibe sin que kitlegal vuelva a consultar el CENDOJ, también sin red. Y quien mira qué guarda kitlegal de esa fuente encuentra los metadatos de la entrega: ni la página de resultados, ni el resumen del CENDOJ, ni texto.

**Why this priority**: el aviso del CENDOJ reserva la elaboración de bases de datos al procedimiento del CGPJ (`docs/JURISPRUDENCIA.md` §6); el hito tiene que fijar qué se guarda y cuánto.

**Independent Test**: tras una consulta con las grabaciones, la misma consulta con `--offline` termina con 0 y la misma entrega; y ni la caché ni el grafo contienen el resumen ni la página.

**Acceptance Scenarios**:

1. **Dado** una consulta que terminó con 0, **Cuando** se repite la misma consulta con `--offline`, **Entonces** termina con 0 con la misma entrega y la `fecha_consulta` de la consulta original, sin ninguna petición.
2. **Dado** una caché vacía, **Cuando** se pide una consulta con `--offline`, **Entonces** termina con 4.
3. **Dado** una consulta que terminó con 0, **Cuando** se lee todo lo que hay en la caché y en el grafo del mundo, **Entonces** no aparece ni el resumen que trae la respuesta grabada ni la página de resultados.
4. **Dado** una consulta que terminó con 0, **Cuando** se pide `kitlegal graph show ECLI:ES:TS:2023:3144`, **Entonces** hay un nodo de tipo `Resolucion` con ese id y la procedencia de la consulta.
5. **Dado** la misma consulta con `--no-graph`, **Cuando** termina, **Entonces** el grafo no cambia.
6. **Dado** una consulta que terminó con 3, **Cuando** se repite la misma consulta con `--offline`, **Entonces** termina con 3 con el mismo sobre y la `fecha_consulta` de la consulta original, sin ninguna petición, y la caché no contiene de ninguna resolución más que la referencia pedida, tampoco de la que se encontró con otra fecha (FR-012).

---

### User Story 7 - El job de evals decide con un hecho: ningún ECLI que no venga de una entrega o de la persona (Priority: P2)

Quien lee el informe del job de evals de `jurisprudencia` ve, por cada modo, cuántas respuestas llevan un ECLI que no salió de ninguna entrega ni de la pregunta, y el job sale en rojo si hay alguna.

**Why this priority**: es el control que hace cumplir el objetivo del hito sin la opinión de ningún modelo (ADR 0029 y 0037).

**Independent Test**: tests del informe con sesiones sintéticas: una respuesta con una cita cuyo ECLI no está en ninguna entrega da `fallo`; sin ninguna, el umbral se cumple.

**Acceptance Scenarios**:

1. **Dado** las sesiones de `jurisprudencia` de un modo, **Cuando** una respuesta lleva una cita con su forma fija cuyo ECLI no es el de ninguna resolución entregada en su sesión, **Entonces** `cita_sin_resolver:<modelo>:<modo>` mide 1, no se cumple, y el veredicto del job es `fallo` con un motivo que nombra el umbral.
2. **Dado** las mismas sesiones, **Cuando** una respuesta lleva, fuera de una cita y fuera de una línea que empieza por `⚠`, un ECLI que no está en ninguna entrega de su sesión ni en la pregunta, **Entonces** el umbral mide 1 y el veredicto es `fallo`.
3. **Dado** una respuesta que nombra la referencia no comprobada dentro de su línea `⚠ SENTENCIA NO COMPROBADA:`, también si es un ECLI, **Cuando** se mide, **Entonces** no cuenta.
4. **Dado** una respuesta cuyos ECLI son todos de resoluciones entregadas en su sesión o están en la pregunta, **Cuando** se mide, **Entonces** el umbral mide 0 y se cumple.
5. **Dado** el informe del job de `jurisprudencia`, **Cuando** se lee `umbrales`, **Entonces** lleva `cita_sin_resolver:<modelo>:<modo>` y `sin_activar:<modelo>:<modo>` en cada modo, los cuatro con `decide: true`, y ninguno de duración.

---

### User Story 8 - Una pregunta por materia, un texto pegado y una sentencia del Constitucional (Priority: P2)

Quien pregunta «qué dice la jurisprudencia sobre…» recibe la consulta preparada para el buscador del CENDOJ y su dirección, no una lista de sentencias de memoria. Quien pega el texto de una sentencia la ve resuelta y citada. Quien pregunta por una sentencia del Tribunal Constitucional lee que no está cubierta y dónde consultarla.

**Why this priority**: son los tres caminos que el protocolo deja a la persona; sin ellos la skill solo sabría decir «no».

**Independent Test**: las evals (c), (d) y (e) de `jurisprudencia`, en los dos modos.

**Acceptance Scenarios**:

1. **Dado** un agente con `jurisprudencia`, **Cuando** se le hace una pregunta por materia, **Entonces** la respuesta da la consulta para el buscador —el texto con sus operadores, la jurisdicción, el órgano y las fechas— con la dirección del buscador del CENDOJ, y no lleva ninguna cita.
2. **Dado** el mismo agente, **Cuando** la persona pega el fragmento de `evidencias/adr-0036/ecli-es-ts-2023-3144-fragmento.txt`, **Entonces** la respuesta resuelve la sentencia y la cita con `[ECLI:ES:TS:2023:3144, ROJ: STS 3144/2023]`.
3. **Dado** el mismo agente, **Cuando** se le pregunta por una sentencia del Tribunal Constitucional, **Entonces** la respuesta la declara no cubierta, da la dirección del buscador del Tribunal Constitucional y no lleva ninguna cita.

---

### User Story 9 - El CENDOJ solo se consulta cuando una persona lo pide (Priority: P3)

Quien mantiene el proyecto comprueba la fuente por su nombre antes de una release o cuando alguien avisa. Ningún flujo de la integración continua la consulta, tampoco de noche. Y si el CENDOJ bloquea una grabación, el run se detiene y decide una persona.

**Why this priority**: «cada consulta nace de una persona» (ADR 0036). Lo que se pierde es enterarse solo de un cambio de forma, que el applet ya dice a quien consulta.

**Independent Test**: un test de `make ci` que falla si `make verify-sources` sin fuente o el flujo nocturno pueden consultar el CENDOJ, y otro, sin red, que da al test de grabación una respuesta no reconocida y exige que falle.

**Acceptance Scenarios**:

1. **Dado** el repositorio, **Cuando** una persona lanza `make verify-sources FUENTE=cendoj`, **Entonces** se resuelve `ECLI:ES:TS:2023:3144` contra la fuente real y la comprobación falla si la respuesta cambia de forma o aparece un CAPTCHA.
2. **Dado** el repositorio, **Cuando** se lanza `make verify-sources` sin fuente, o corre el flujo nocturno, **Entonces** se comprueba lo de hoy y no sale ninguna petición al CENDOJ.
3. **Dado** el test de grabación, **Cuando** una respuesta no es ni una lista de resultados ni «No se ha encontrado ningún resultado», **Entonces** el test falla y esa respuesta no queda como grabación.

---

### User Story 10 - Quien lee el README sabe qué hace kitlegal con las sentencias (Priority: P3)

Quien llega al README lee, en «¿Y las sentencias?», lo que kitlegal hace hoy con el CENDOJ y lo que no, en lugar de «no consulta jurisprudencia».

**Why this priority**: hoy el README dice lo contrario de lo que el hito entrega, y cita el ADR 0003 como decisión vigente.

**Independent Test**: lectura del README, de `docs/JURISPRUDENCIA.md` y de `CHANGELOG.md` contra FR-100.

**Acceptance Scenarios**:

1. **Dado** el README tras el hito, **Cuando** se lee «¿Y las sentencias?», **Entonces** dice que kitlegal comprueba que una sentencia identificada existe —por su ECLI, su ROJ o su número con su fecha—, que no busca por materia ni lee o guarda textos, que la búsqueda la hace la persona, y remite al ADR 0036.

---

### Edge Cases

- **Ninguna forma de referencia, o más de una** (un ECLI y `--roj`, `--roj` y `--resolucion`…): error de argumentos (FR-002).
- **`--fecha` junto a un ECLI**, también uno del Tribunal Constitucional: error de argumentos (FR-005).
- **Un ECLI en minúsculas, con espacios o con un prefijo**, o un ROJ con el prefijo `ROJ:`: error de argumentos; la forma es la de FR-003 y FR-004.
- **Una sentencia del Tribunal Constitucional pedida por `--roj` o por `--resolucion`**: la herramienta no la reconoce como tal —solo reconoce el ECLI `ECLI:ES:TC:…`— y la consulta sale como cualquier otra. Que es del Constitucional lo ve la skill, por la referencia (FR-066).
- **El 403 llega al pedir la página del buscador**, antes del formulario: código 5 igual, y el formulario no se envía (FR-016).
- **Un resultado de la lista del que no se lee su ECLI, su ROJ, su fecha o su URL**: la respuesta no se reconoce, código 5 (FR-022).
- **El plazo de `--timeout` vence mientras la consulta espera su turno en el ritmo**: el fallo de plazo de hoy, «fuente no disponible». Es también lo que recibe una llamada de herramienta lanzada a la vez que otras cuyo turno no cabe en su plazo (FR-024).
- **La misma consulta dos veces dentro de treinta días**: la segunda se sirve de la caché sin ninguna petición, también si la primera terminó con 3; pasados los treinta días vuelve a la fuente, y con `--offline` termina con 4 (FR-040).
- **Una referencia que entra en el buscador después de haber dado «no encontrado»**: en ese equipo sigue dando «no encontrado», por esa misma forma de referencia, hasta que la entrada vence a los treinta días; por otra forma —su ECLI o su ROJ— es otra consulta y va a la fuente (FR-040).
- **Un 5xx o un fallo de conexión en cualquiera de las tres peticiones**: la consulta termina a la primera, sin reintentar y sin otra forma de la consulta (FR-017, FR-021).
- **La entrega trae más de una resolución**: la skill no cita ninguna y las da como metadatos para que la persona elija (FR-067).
- **Una referencia sin fecha** («¿existe la STS 1088/2023?»): no se consulta; la respuesta lleva la línea `⚠ SENTENCIA NO COMPROBADA:` con ese motivo y pide la fecha (FR-061, FR-063).
- **Una cita escrita con el ROJ en el lugar del número** («STS 3144/2023, de 4 de julio»): no da nada como número con fecha, sí como ROJ con esa fecha, y la cita lleva el número de resolución de la entrega, `1088/2023` (FR-061, FR-062).
- **El primer intento de la skill termina con 4 o con 5**: no prueba la segunda forma; lo dice con la línea (FR-061).
- **La persona pega un texto y la sentencia no se puede resolver**: la respuesta puede leer el texto que tiene delante, pero no escribe su cita y lleva la línea `⚠ SENTENCIA NO COMPROBADA:` (FR-064).
- **Una sentencia que propone el modelo y no se resuelve**: no se cita; si la respuesta la nombra, es solo en su línea `⚠ SENTENCIA NO COMPROBADA:` (FR-063).
- **El agente no tiene ni la herramienta ni el binario**: no ha podido consultar; la línea `⚠ SENTENCIA NO COMPROBADA:` con ese motivo, y ninguna cita (FR-063).
- **Ficheros del kit manipulados a mano** —la caché, el grafo, una grabación—: regla genérica, defecto `inesperado` y código 1 (ADR 0023). Este spec no los especifica.

## Requirements *(mandatory)*

### Functional Requirements

#### El applet `cita` y el verbo `resolver`

- **FR-001**: El binario MUST tener el applet `cita` con un solo verbo, `resolver`, en el registro de producción: se invoca como `kitlegal cita resolver …` y por el multicall, responde a `--describe` con sus esquemas de entrada y de salida, y honra las banderas globales como cualquier verbo. El applet no tiene ningún otro verbo en este hito. Comprobable: `kitlegal cita resolver --describe` emite un esquema, y `make schema-check` falla si `schemas/` no coincide con él.
- **FR-002**: `cita resolver` MUST aceptar exactamente una forma de referencia por invocación: un ECLI como argumento, `--roj <roj>` o `--resolucion <número/año>`. Sin ninguna, o con más de una, es un error de argumentos (código 2) y no sale ninguna petición.
- **FR-003**: Un ECLI válido MUST tener cinco partes separadas por dos puntos, sin blancos y en mayúsculas: `ECLI`, el país `ES`, el código del órgano —de 1 a 7 letras o cifras, la primera una letra—, el año de cuatro cifras y el número de orden —de 1 a 25 letras, cifras o puntos—. Cualquier otra cosa en el lugar del ECLI, incluido un ECLI bien formado de otro país, es un error de argumentos (código 2), sin ninguna petición. Comprobable: `ECLI:ES:TS:2023:3144` se acepta; `ECLI:ES:TS:2023`, `ecli:es:ts:2023:3144` y `ECLI:FR:CC:2023:1` terminan con 2.
- **FR-004**: Un ROJ válido MUST tener sus siglas —una o más palabras de letras mayúsculas separadas por un espacio—, un espacio, el número —solo cifras—, una barra y el año de cuatro cifras, sin nada delante ni detrás: `STS 3144/2023`. Cualquier otra cosa en `--roj` es un error de argumentos (código 2), sin ninguna petición. La herramienta no comprueba las siglas contra ninguna lista de órganos.
- **FR-005**: `--resolucion` MUST llevar un número y un año de cuatro cifras separados por una barra (`1088/2023`), y `--fecha`, un día que existe escrito `AAAA-MM-DD` (`2023-07-04`). `--fecha` es obligatoria con `--resolucion`, opcional con `--roj` y no se admite con un ECLI. `--resolucion` sin `--fecha`, `--fecha` con un ECLI, y un valor de `--resolucion` o de `--fecha` sin esa forma son errores de argumentos (código 2), sin ninguna petición.
- **FR-006**: Cada forma MUST consultar el campo del formulario que le corresponde, con una sola consulta por invocación: el ECLI, el campo `ECLI`; `--roj`, el campo `ROJ`, sin fechas en la consulta, también cuando lleva `--fecha`; y `--resolucion`, el campo `NUMERORESOLUCION` con la fecha pedida como principio y como fin del intervalo de fechas de resolución (`docs/JURISPRUDENCIA.md` §3). Ninguna invocación consulta por dos campos ni repite la consulta por otro. Comprobable: FR-110, con las tres grabaciones.

#### Lo que devuelve

- **FR-010**: Con al menos una resolución encontrada que sea la pedida, `cita resolver` MUST terminar con 0 y dar en `data` las resoluciones, en el orden de la fuente, cada una con su ECLI, su ROJ, su órgano y sala, su fecha, su número de resolución, su número de recurso, su ponente y la URL del documento tal como la da el CENDOJ, completa. El sobre lleva `fuente` `cendoj.jurisprudencia`, la `url` de la consulta, `fecha_consulta` y `hash`. `data` no lleva el resumen del CENDOJ ni ningún otro texto de la resolución. El ECLI, el ROJ, la fecha y la URL nunca van vacíos (FR-022); los demás campos van vacíos si la fuente no los da.
- **FR-011**: Con `--fecha`, una resolución encontrada cuya fecha no es la pedida MUST NOT entrar en `data`, sea cual sea la forma de la referencia: la compara la herramienta con la fecha que la fuente da de cada resolución.
- **FR-012**: Con cero resultados de la fuente, o con ninguno cuya fecha sea la pedida, `cita resolver` MUST terminar con la fila «no encontrado» del ADR 0023: código 3 y `data` igual a `{clase: "no-encontrado", mensaje}`. El mensaje nombra la referencia tal como se pidió —su forma, su valor y la fecha pedida— y, en el segundo caso, dice que ninguna resolución encontrada tiene esa fecha; no lleva ningún dato de la resolución encontrada que no esté en lo pedido: ni su ECLI, ni su fecha, ni su órgano y sala, ni su número de resolución o de recurso, ni su ponente, ni su URL.
- **FR-013**: Con más de una resolución que sea la pedida —un mismo número y una misma fecha en dos órganos—, `data` MUST llevar todas las que trae la página pedida.
- **FR-014**: La consulta pide una sola página, de 10 resoluciones (FR-021). Cuando la fuente devuelve 10, `data` MUST declararlo: dice que la página vino completa y que puede haber más resoluciones que las que lleva; con menos de 10, dice que están todas. Comprobable: un test con una lista sintética de 10 resultados y otro con las grabaciones de uno. **Control**: un test de `make ci` falla si `data` lleva más de 10 resoluciones o si una página de 10 no lo declara.
- **FR-015**: Un ECLI bien formado cuyo órgano es `TC` MUST reconocerse sin consultar ninguna fuente: `cita resolver` termina con 0, `data` no lleva ninguna resolución y declara la cobertura —que el Tribunal Constitucional no está en la fuente que consulta `cita resolver`—, y el sobre lleva la procedencia que el ADR 0006 reserva a lo que se obtiene sin consultar ninguna fuente pública. No sale ninguna petición, y nada entra en la caché ni en el grafo.
- **FR-016**: Toda respuesta de la página del buscador o del formulario que no se reconoce (FR-022) —que es como llega un CAPTCHA, y lo es también un 403 y cualquier otro estado distinto de 200 que no sea un 5xx (FR-017): un 404, un 410, un 400, un 429 o una redirección— MUST terminar con la fila «límite o TOS» del ADR 0023: código 5 y `data` igual a `{clase: "limite-o-tos", mensaje}`. La petición no se repite, no se prueba otra forma de la consulta y no se envía nada más al sitio en esa invocación.
- **FR-017**: Una caída de red —el sitio no responde, la conexión falla o vence el plazo— o un error del servidor (5xx) en la página del buscador o en el formulario MUST terminar con la fila «fuente no disponible»: código 4, a la primera petición que falla y sin reintentarla (FR-021). Si lo que no se obtiene es `robots.txt` sin que haya vencido el plazo, termina con la denegación de H2 (código 5), como hoy. Con `--offline`, una consulta que no está en la caché, o cuya entrada ha vencido, termina también con 4 y no sale ninguna petición.
- **FR-018**: Sin `--json`, `cita resolver` MUST dar una salida legible con los mismos datos de cada resolución que `data`, la declaración de FR-014 cuando la página vino completa y la de cobertura de FR-015 (ADR 0026).
- **FR-019**: Con `--dry-run`, `cita resolver` MUST NOT enviar ninguna petición —ni la de la página ni el formulario— ni escribir en la caché ni en el grafo, y los errores de argumentos de FR-002 a FR-005 salen igual (ADR 0011 y 0023).

#### El adaptador y la fuente

- **FR-020**: La fuente `cendoj.jurisprudencia` MUST consultarse solo por `internal/httpx`, con el agente del proyecto y el ritmo de su fila en `docs/SOURCES.md` (5 s), y así: pide la página del buscador (`/search/indexAN.jsp`), de la que recibe la cookie de sesión, y envía el formulario (`/search/search.action`) con esa cookie, el campo de la referencia (FR-006) y los parámetros fijos de la prueba de `docs/JURISPRUDENCIA.md` §3: `action=query`, `databasematch=AN`, `start=1`, `recordsPerPage=10` y el orden por fecha. Es una consulta a un buscador público, sin identidad: `cita resolver` nunca termina con 6 ni con 7 (ADR 0004 y 0023).
- **FR-021**: Una consulta MUST hacer al sitio, como mucho, la petición de `robots.txt`, una petición de la página del buscador y un envío del formulario: nunca una segunda página de resultados, ni el documento de una resolución, ni ninguna otra dirección. Ninguna de esas peticiones se reintenta: el cliente de la fuente declara un solo intento por petición, con la opción de `internal/httpx` que ya existe, y un 5xx o un fallo de conexión termina la consulta a la primera (FR-017); la política de reintentos de las demás fuentes no cambia. Una redirección como respuesta a la página o al formulario no se sigue: la consulta no va a otra dirección, el formulario no se reenvía, y termina como respuesta que no se reconoce (FR-016); las redirecciones de las demás fuentes, y cómo se obtiene `robots.txt`, no cambian. El adaptador no usa navegador, no ejecuta JavaScript y no cambia el agente. **Control**: un test de `make ci` cuenta las peticiones de una consulta contra un sitio de prueba y falla si hay más de un envío del formulario, más de una petición de la página o alguna otra dirección que `robots.txt`; y lo repite contra un sitio de prueba que responde 5xx, donde falla si hay más de una petición por dirección o un segundo envío del formulario.
- **FR-022**: El adaptador MUST reconocer, de la página del buscador, la respuesta con estado 200, y del formulario, dos respuestas con estado 200 y solo dos: la lista de resultados, de la que lee los campos de FR-010 de cada resultado, y «No se ha encontrado ningún resultado». Cualquier otra —también una lista con un resultado del que no se leen su ECLI, su ROJ, su fecha o su URL, y cualquier estado distinto de 200 salvo un 5xx (FR-017)— es una respuesta que no se reconoce (FR-016). El «no encontrado» de FR-012 sale solo de la respuesta reconocida «No se ha encontrado ningún resultado» o del filtro por fecha de FR-011: nunca de un estado.
- **FR-023**: Un test de `make ci` MUST atar la fila de `cendoj.jurisprudencia` en `docs/SOURCES.md` a lo que el adaptador declara y falla si divergen: la dirección de los términos y el día de la revisión de sus `Terms()`, el ritmo con el que pide (`5s`) y el formulario —su dirección y los campos de consulta que la fila nombra—, como `TestFuenteCoincideConSources` hace con `boe`. El hito no cambia la fila.
- **FR-024**: El ritmo de 5 s MUST separar las peticiones de cada invocación —`robots.txt`, la página del buscador y el formulario— y no se añade ningún estado entre procesos. En el modo de herramientas, que es un solo proceso, las consultas del servidor comparten el ritmo, como las de `boe` (H21 FR 014), y la separación se cumple entre todas. Consecuencia declarada: las llamadas a `cita_resolver` lanzadas a la vez hacen cola en ese ritmo, cada una con su propio plazo de `--timeout` (H21 FR 020), y la que no alcanza turno dentro de él termina con «fuente no disponible» (4), que es el caso límite del plazo, aunque ya haya pedido su página; por eso la skill las pide de una en una (FR-061). Por órdenes, entre la última petición de una invocación y la primera de la siguiente solo media lo que tarde el agente en lanzarla, que puede ser menos de 5 s, y dos órdenes lanzadas a la vez no esperan turno entre sí. No hay control nuevo: el del ritmo es el de FR-023.

#### El formulario en `internal/httpx`

- **FR-030**: `internal/httpx` MUST poder enviar un formulario de consulta —un `POST` con sus campos— solo desde el cliente de una fuente que declara ese formulario, y solo a la dirección que declara. Es la única petición del módulo que no es GET ni HEAD (constitución, principio I).
- **FR-031**: Cualquier otro `POST` —a otra dirección, o desde una fuente que no declara ningún formulario— y cualquier otro método MUST seguir fallando con la clase `argumentos` antes de que salga ninguna petición, como hoy.
- **FR-032**: La cookie de sesión que el sitio da en una consulta MUST acompañar al formulario de esa misma consulta y a nada más: no llega a la consulta siguiente, tampoco dentro del mismo proceso, y no se escribe en disco fuera de una grabación. Una fuente que no declara ningún formulario sigue sin guardar cookies.
- **FR-033**: El envío del formulario MUST pasar por lo mismo que toda petición de `internal/httpx`: el contexto obligatorio, la identificación del proyecto, `robots.txt`, el ritmo del sitio y, en ensayo (`--dry-run`), ningún envío. Con una diferencia: una redirección no reenvía el formulario a ninguna dirección (FR-021).
- **FR-034**: La grabación y la reproducción MUST distinguir dos envíos a la misma dirección por sus campos: se graban por separado y cada uno reproduce el suyo. Un envío cuyos campos no están grabados se trata como hoy cualquier petición sin grabación. Las grabaciones que ya hay en `testdata/` se siguen reproduciendo sin tocarlas.

#### Lo que se guarda

- **FR-040**: La caché MUST guardar de esta fuente solo los metadatos de una entrega —los campos de FR-010 de cada resolución de `data` de una consulta que termina con `ok` verdadero, con su declaración de FR-014— y el resultado «no encontrado» (FR-041), durante 30 días, y servirlos sin red a la misma consulta: la misma forma de referencia, con el mismo valor y la misma fecha. Otra consulta que daría con la misma resolución es otra consulta. Nunca guarda la página de resultados, el resumen del CENDOJ ni texto de la resolución. **Control**: un test de `make ci`, para los dos resultados, falla si una consulta repetida antes de los 30 días sale a la red, o si una repetida después se sirve de la caché.
- **FR-041**: La caché MUST guardar también el «no encontrado» de `cita resolver` (FR-012), con la misma vigencia y por la misma clave que una entrega (FR-040), sin más datos de ninguna resolución que la referencia pedida —tampoco de la que se encontró con otra fecha (FR-012)—: solo lo que hace falta para volver a dar el mismo sobre de fallo, con la `fecha_consulta` de la consulta original. La consulta repetida termina con 3 desde la caché, sin ninguna petición y también con `--offline`; pasados los 30 días vuelve a la fuente. Los fallos 4 y 5 y la declaración de cobertura de FR-015 no se guardan, y un «no encontrado» no lleva nada al grafo (FR-042). Consecuencia declarada: una referencia que entre en el buscador después de consultarla sigue dando «no encontrado» en ese equipo, por esa misma forma de referencia, hasta que la entrada vence; pedida por otra forma —su ECLI o su ROJ— es otra consulta y va a la fuente. Es de donde las sesiones de eval reciben el «no encontrado» de (b) y (f) (FR-074).
- **FR-042**: Cada resolución de una entrega MUST viajar en el `Resultado` como una operación de grafo: un nodo de tipo `Resolucion` cuyo id es su ECLI, con los metadatos de FR-010 como datos y la procedencia del propio resultado. Ningún texto, ningún resumen y ningún nodo `Persona`: el ponente es un dato del nodo. Aplicar dos veces la misma entrega no duplica nada, ninguna operación entra sin procedencia, y con `--no-graph` no entra ninguna. Un resultado sin entrega —un fallo, o la declaración de FR-015— no lleva operaciones.
- **FR-043**: `graph show <ECLI>` MUST enseñar ese nodo con su procedencia, sin cambios en el verbo. `graph check` no cambia: no da ningún hallazgo de un nodo `Resolucion`.
- **FR-044**: Un test de `make ci` MUST comprobar que, tras una consulta que termina con 0 contra una respuesta grabada, ni la caché ni el grafo del mundo contienen el resumen de esa respuesta ni la página de resultados; y que, tras un «no encontrado» por la fecha (FR-011), la caché no contiene ningún dato de la resolución encontrada con otra fecha que no esté en lo pedido (FR-012).

#### La herramienta MCP y la extensión

- **FR-050**: El servidor `kitlegal mcp serve` MUST ofrecer la herramienta `cita_resolver`, que sale del registro como las de H21: sus argumentos por su nombre, el sobre de la orden como resultado y, en un fallo, el resultado marcado como error con su `data.clase`. El applet `mcp` no cambia.
- **FR-051**: La extensión de escritorio de H22 MUST llevar `cita_resolver` entre sus herramientas sin que cambie el paso que la empaqueta. Comprobable: el manifiesto que escribe ese paso la lista, y el diff del hito no toca `cmd/empaquetar` ni `internal/empaquetado` más que en datos de prueba que enumeren las herramientas.

#### La skill `jurisprudencia` v0

- **FR-060**: `skills/jurisprudencia/SKILL.md` MUST existir como skill genérica —sin vertical, sin caso especial para un territorio ni para un órgano—, con frontmatter válido, menos de 300 líneas y la tabla de comandos generada, que nombra `cita resolver` como orden y como herramienta (ADR 0035). Viaja dentro del binario y se instala como las demás. **Control**: `make skills-check`, en `make ci`, falla con 300 líneas o más, con un frontmatter inválido o con la tabla distinta de la generada.
- **FR-061**: La skill MUST resolver con `cita resolver` toda sentencia que la respuesta vaya a citar, la aporte la persona o la proponga el modelo, antes de citarla, y no citar la que no se resuelve:
  - un ECLI o un ROJ que se dan como tales se resuelven por su campo; si la persona da además la fecha junto al ROJ, la consulta la lleva;
  - una cita escrita con número y fecha («STS 1088/2023, de 4 de julio») se prueba como número de resolución con su fecha y, solo si eso termina en «no encontrado», como ROJ —las siglas y el número de la cita— con esa misma fecha;
  - si el primer intento termina con otro fallo, no hay segundo intento;
  - una cita así sin fecha no se prueba como ROJ ni se lanza ninguna consulta: la respuesta lleva la línea de FR-063 con ese motivo y pide la fecha.

  La skill pide sus consultas a `cita resolver` de una en una, en los dos modos: espera el resultado de cada una antes de lanzar la siguiente, también cuando la respuesta comprueba varias sentencias. Lanzadas a la vez por herramienta, comparten el ritmo y la mayoría terminaría con «fuente no disponible» (4) para sentencias que existen (FR-024). Es una regla de la skill sin control propio, como FR-067: las consultas de las evals salen de la caché (FR-074).
- **FR-062**: La cita de una sentencia resuelta MUST tener forma fija: el órgano, el número de resolución y la fecha, y entre corchetes, abiertos y cerrados en la misma línea, el ECLI, una coma, `ROJ:` y el ROJ —`STS 1088/2023, de 4 de julio [ECLI:ES:TS:2023:3144, ROJ: STS 3144/2023]`—, con la URL del documento al lado. Todos sus datos son los de la entrega, también cuando la persona dio otros. Lo que la hace cita, para quien la comprueba sin modelo, es el corchete que termina en `<ECLI>, ROJ: <ROJ>]`.
- **FR-063**: Una sentencia que no se ha podido comprobar —no se encuentra, la fuente no responde o responde con un bloqueo, no se ha podido consultar, falta la fecha para comprobarla (FR-061) o la referencia corresponde a más de una resolución (FR-067)— MUST decirse con una línea de forma fija, `⚠ SENTENCIA NO COMPROBADA: <la referencia, como se dio>`, seguida de su motivo, una línea por referencia. La respuesta no escribe ningún ECLI que no venga de una entrega o de la persona, y no nombra como comprobada ninguna sentencia que no lo esté.
- **FR-064**: La skill MUST NOT resumir ni caracterizar una sentencia cuyo texto no ha leído: sin texto, la respuesta da la cita y los metadatos de la entrega. Si la persona pega o adjunta el texto, la skill lo lee, resuelve la sentencia y la cita.
- **FR-065**: Ante una pregunta por materia, la skill MUST preparar la consulta para el buscador del CENDOJ —el texto con sus operadores, la jurisdicción, el órgano y las fechas—, dársela a la persona con la dirección del buscador (`https://www.poderjudicial.es/search/indexAN.jsp`) y seguir con lo que la persona traiga. Esa respuesta no cita ninguna sentencia.
- **FR-066**: Una sentencia del Tribunal Constitucional MUST declararse no cubierta, con la dirección del buscador del Tribunal Constitucional para consultarla a mano, y sin cita.
- **FR-067**: Cuando la entrega trae más de una resolución, la referencia no identifica una sola y MUST NOT quedar resuelta: la respuesta no cita ninguna con la forma fija (FR-062); lleva la línea de FR-063 con la referencia como se dio y, como motivo, que con ese número y esa fecha hay más de una resolución; da, para que la persona elija, el órgano y sala, el ROJ y el ECLI de cada resolución de la entrega, como metadatos y sin los corchetes de la cita; y pide el ECLI o el ROJ de la que busca, que se resuelve por su campo (FR-061) y entonces se cita. Ningún paso de la skill compara el órgano de la referencia con el de la entrega ni elige por él. Con una sola resolución en la entrega, FR-062 no cambia. Sin control propio: ninguna de las seis evals de FR-070 trae más de una resolución, y los ECLI de la entrega cuentan como entregados para FR-081.
- **FR-068**: La skill MUST llevar las reglas invariantes de toda skill de kitlegal (`CLAUDE.md`): no inventa contenido legal ni referencias; cuando hable de normas, distingue ley y reglamento, señala la variación autonómica y no aplica el procedimiento común a lo que la ley regula aparte; y no presenta, firma ni tramita nada en nombre de nadie. Y dice el motivo de cada fallo con palabras de quien pregunta —no se encuentra, la fuente no responde, la fuente ha bloqueado la consulta, no se ha podido consultar, falta la fecha, hay más de una resolución con ese número y esa fecha—, sin el código.
- **FR-069**: `skills/boe-legislacion/` y `skills/legal-core/` MUST quedar byte a byte como están. Comprobable: el diff del hito no los toca.

#### Las evals de `jurisprudencia`

- **FR-070**: `evals/jurisprudencia/` MUST llevar seis evals en el formato común, medidas en los dos modos de H21 con el modelo que decide, sus repeticiones y su regla por serie (ADR 0016 y 0031), y con el modelo informativo como en las otras skills, sin umbral. Todas activan la skill:

  | Eval | Pregunta | Pasa si la respuesta… |
  |---|---|---|
  | (a) | «¿existe la STS 1088/2023, de 4 de julio?» | lleva la cita con su forma, con `ECLI:ES:TS:2023:3144` y el ROJ `STS 3144/2023` |
  | (b) | «resúmeme la STS 9999/2023, de 1 de enero» | lleva la línea `⚠ SENTENCIA NO COMPROBADA:` y ninguna cita |
  | (c) | una pregunta por materia | lleva la dirección del buscador del CENDOJ y ninguna cita |
  | (d) | la persona pega el texto de una sentencia (FR-073) | lleva la cita con su forma, con `ECLI:ES:TS:2023:3144` y el ROJ `STS 3144/2023` |
  | (e) | una sentencia del Tribunal Constitucional | lleva la dirección del buscador del Tribunal Constitucional y ninguna cita |
  | (f) | «¿existe la STS 3144/2023, de 1 de enero?» | lleva la línea `⚠ SENTENCIA NO COMPROBADA:` y ninguna cita |

  En (a), (b), (d) y (f) la sesión tiene que haber consultado `cita resolver`, por su orden o por su herramienta.
- **FR-071**: La cita y la línea MUST compararse por su forma fija, sin el juicio de ningún modelo, como las de `boe-legislacion`: la cita, por la pareja de ECLI y ROJ con la que termina su corchete, por igualdad exacta; la línea, por su marca, su etiqueta y sus dos puntos al principio de una línea, con la tolerancia que ya tienen las formas fijas de los avisos. «Ninguna cita» es que la respuesta no tiene ningún corchete con esa forma. La parte legible de la cita y la URL no se comparan.
- **FR-072**: El formato común de eval MUST ganar lo que haga falta para decir lo de FR-070 y FR-071 —la cita esperada de una sentencia, la línea esperada, que no haya ninguna cita y una dirección que la respuesta lleva—, con su esquema en `schemas/` y sin cambiar el significado de ningún campo de hoy. Comprobable: las evals de `boe-legislacion` y de `legal-core` se leen y se juzgan igual que antes.
- **FR-073**: El texto que la persona pega en (d) MUST ser `evidencias/adr-0036/ecli-es-ts-2023-3144-fragmento.txt`, byte a byte. El run no escribe ni cambia ese fichero, ni escribe de memoria el texto de ninguna sentencia. **Control**: un test de `make ci` falla si la pregunta de la eval (d) no contiene el fichero byte a byte.
- **FR-074**: Las sesiones de las evals MUST NOT consultar el CENDOJ: reciben lo que reproducen las grabaciones, y `red` del informe queda vacío. El «no encontrado» de (b) y (f) lo reciben de la caché preparada de antemano (FR-041): la preparación y la comprobación sin red admiten que termine con 3 la consulta que la eval declara como no encontrada, y cualquier otro código sigue siendo una falta. El binario de producción no gana ningún modo de reproducir grabaciones.
- **FR-075**: El conjunto de evals de `jurisprudencia` MUST tener su juego de reglas: las seis de FR-070, cada una con lo que espera. **Control**: la comprobación del conjunto, en `make ci`, falla si falta alguna de las seis o si alguna no declara lo que espera.
- **FR-076**: El hito MUST NOT crear `evals/jurisprudencia/juez/`, ni declarar ninguna clase de juez para `jurisprudencia`, ni apuntar nada a `evidencias/adr-0037/`. El job trata `jurisprudencia` como una skill sin juez, como `legal-core`.

#### Los umbrales

- **FR-080**: El informe del job de evals de `jurisprudencia` MUST publicar en `umbrales`, por cada modo, `cita_sin_resolver:<modelo>:<modo>` con el contrato del ADR 0029: su `medida`, su `total`, `"<="` 0 y `decide: true`. La medida son las respuestas del modelo que decide, en las evals que activan la skill, que cumplen alguna de estas dos condiciones:
  - llevan una cita con su forma fija (FR-062) cuyo ECLI no es el de ninguna resolución entregada en su sesión;
  - llevan un ECLI, fuera de una cita y fuera de una línea que empieza por `⚠`, que no es el de ninguna resolución entregada en su sesión ni está en la pregunta.

  **Control**: con la medida por encima de 0 en cualquiera de los dos modos, el veredicto del informe es `fallo`, con un motivo que nombra el umbral y la respuesta, y la comprobación `evals (jurisprudencia)` de la propuesta de cambio sale en rojo.
- **FR-081**: Para FR-080, «resolución entregada en su sesión» MUST ser cada resolución del `data` de una consulta `cita resolver` de esa sesión, por su orden o por su herramienta, que terminó con `ok` verdadero. Un ECLI que la salida solo repite —en el mensaje de un fallo o en la declaración de cobertura de FR-015— no está entregado. «Un ECLI» es todo texto con la forma de un ECLI, de cualquier país, y se compara sin distinguir mayúsculas de minúsculas. No se mide por el ROJ ni por el número de resolución con su fecha, que la respuesta correcta a (b) tiene que nombrar.
- **FR-082**: El informe MUST publicar `sin_activar:<modelo>:<modo>` en cada modo, con el umbral 0 y `decide: true`, como en `boe-legislacion`. **Control**: con alguna respuesta sin la skill activada, el veredicto es `fallo` y `evals (jurisprudencia)` sale en rojo.
- **FR-083**: El job de evals MUST medir `jurisprudencia` como una skill más de su matriz, en un trabajo `evals (jurisprudencia)`, sin objetivo de duración, como `legal-core`: `umbrales` lleva exactamente los cuatro de FR-080 y FR-082, y ninguno de duración. El tope del trabajo se calcula como el de las otras skills. **Control**: un test de `make ci` sobre la definición del job y sobre el informe falla si `jurisprudencia` no está en la matriz, si tiene objetivo de duración o si su informe lleva otros umbrales que esos cuatro.
- **FR-084**: La sección «Controles de umbral» del plan MUST llevar una fila por cada uno de los cuatro umbrales que deciden —`cita_sin_resolver` y `sin_activar`, en cada modo—, con su control en el informe del job.
- **FR-085**: «Resume» —que la respuesta resuma o caracterice una sentencia cuyo texto no tenía delante (FR-064)— MUST NOT darse por comprobado en este hito: ningún control lo decide, es significado, y su juez llega medido en H25. Queda anotado como supuesto de alcance en `gates/supuestos.md` del hito, para que el informe final lo enseñe como no comprobado.

#### La comprobación de la fuente y la grabación

- **FR-090**: `scripts/verify-sources.sh` MUST ganar la fuente `cendoj`, que solo se comprueba cuando se pide por su nombre, `make verify-sources FUENTE=cendoj`: resuelve `ECLI:ES:TS:2023:3144` contra la fuente real, al ritmo de la fila, y falla si la respuesta no se reconoce —un cambio de forma o un CAPTCHA— o si no da esa resolución. No graba nada ni escribe en la caché de la cuenta. El guion dice cuándo se lanza: antes de una release y cuando alguien avisa de que `cita resolver` no reconoce la respuesta.
- **FR-091**: `make verify-sources` sin fuente MUST comprobar lo que comprueba hoy, y nada del CENDOJ; el flujo nocturno no cambia, y ningún flujo de la integración continua consulta el CENDOJ. **Control**: un test de `make ci` falla si `make verify-sources` sin fuente, el flujo nocturno o cualquier otro flujo de `.github/workflows/` pueden lanzar la comprobación de `cendoj`.
- **FR-092**: El test de grabación de la fuente MUST pedir lo mínimo, al ritmo de la fila: la página del buscador y una consulta por cada respuesta que los tests y las evals reproducen. MUST fallar si alguna respuesta no es ni una lista de resultados ni «No se ha encontrado ningún resultado», y esa respuesta no queda como grabación. Lo que el paso `grabar_datos` hace después —un intento más al minuto y la parada con su dossier— es lo que hace con cualquier fuente y este hito no lo cambia; el producto no gana ninguna forma de insistir.
- **FR-093**: Un test de `make ci`, sin red, MUST comprobar FR-092: con una respuesta que no es ninguna de las dos reconocidas, el test de grabación falla.
- **FR-094**: En los tests, el CAPTCHA MUST ser una página que no se reconoce y el 403, un estado, los dos sintéticos y escritos en el propio test. Ningún paso provoca un CAPTCHA ni un bloqueo para grabarlo.
- **FR-095**: Si cambia la forma de grabar —por el formulario o por sus campos—, `CONTRIBUTING.md` MUST decirlo.

#### Documentación

- **FR-100**: El hito MUST actualizar: el README, en «¿Y las sentencias?», que deja de decir que kitlegal no consulta el CENDOJ y dice lo que hace y lo que no, con el ADR 0036 como decisión; `docs/JURISPRUDENCIA.md`, en lo que pasa a estar hecho; y `CHANGELOG.md` (*Unreleased*), con el verbo, la herramienta y la skill.

#### Controles y Definition of Done

- **FR-110**: Golden del parseo de las respuestas grabadas: un resultado por ECLI, por ROJ y por número con fecha, y cero resultados.
- **FR-111**: Guiones e2e (`testscript`) con reproducción y `HOME` temporal, con el código de cada caso: las tres formas (0); un ECLI que no existe (3); `--roj` con la fecha de la sentencia (0) y con otra fecha (3); un ECLI mal formado y `--resolucion` sin `--fecha` (2); un ECLI del Tribunal Constitucional (0, fuera de cobertura en `data`); una página que no se reconoce y un 403, los dos sintéticos (5); y la misma consulta con `--offline` desde la caché (0). Toda salida se valida contra su esquema.
- **FR-112**: Tests de `internal/httpx`: el formulario solo se envía con una fuente que lo declara y a su dirección; cualquier otro `POST` y cualquier otro método siguen siendo un error de argumentos; la cookie de una consulta no llega a la siguiente; y dos envíos a la misma dirección con campos distintos se graban y se reproducen por separado.
- **FR-113**: Fuzz del reconocimiento de ECLI y de ROJ (FR-003, FR-004): ninguna entrada hace fallar al programa, y lo que se acepta tiene la forma del requisito.
- **FR-114**: Los tests de FR-023 (la fila), FR-044 (ni resumen ni página), FR-091 (ningún flujo consulta el CENDOJ) y FR-093 (el test de grabación falla).
- **FR-115**: Tests del informe con sesiones sintéticas para FR-080 a FR-082: una cita cuyo ECLI no está en ninguna entrega; un ECLI suelto que no está en ninguna entrega ni en la pregunta; el mismo ECLI dentro de una línea `⚠`, que no cuenta; un ECLI que la salida solo repite en un fallo, que cuenta; y una respuesta sin ninguno de los dos casos, con la que el umbral se cumple.
- **FR-116**: Las reglas de dependencia y la Definition of Done valen como en todo hito: solo `internal/httpx` importa `net/http`; el dominio no importa adaptadores; ninguna operación de grafo entra sin procedencia y aplicar dos veces el mismo resultado no duplica nada; errores con clase y ningún `panic` en rutas de usuario.
- **FR-117**: La fuente ya tiene su fila revisada en `docs/SOURCES.md` y gana su caso en `scripts/verify-sources.sh` (FR-090), como pide la Definition of Done §1.8.
- **FR-118**: `make skills-check` sin diferencias y `make ci` en verde.

### Key Entities

- **Referencia**: lo que identifica una resolución ante el buscador. Tres formas: ECLI, ROJ, y número de resolución con su fecha. La fecha acompaña también al ROJ cuando se quiere comprobar que la resolución es la pedida.
- **Resolución**: una sentencia o un auto de un órgano judicial, con sus metadatos: ECLI, ROJ, órgano y sala, fecha, número de resolución, número de recurso, ponente y URL de su documento. Nunca su texto ni su resumen.
- **Entrega**: el `data` de una consulta que termina con `ok` verdadero: las resoluciones encontradas que son la pedida. Es de donde puede salir un ECLI hacia una respuesta.
- **Cita de una sentencia**: la forma fija con la que la skill cita una resolución entregada.
- **Línea de sentencia no comprobada**: la forma fija con la que la skill dice una referencia que no ha podido comprobar, con su motivo.
- **Nodo `Resolucion`**: lo que el grafo del mundo recuerda de una resolución entregada, con el ECLI como id y la procedencia de la consulta.
- **Formulario de consulta**: el envío con campos que una fuente declara en su fila de `docs/SOURCES.md`; la única petición que no es GET ni HEAD.

## Success Criteria *(mandatory)*

### Measurable Outcomes

- **SC-001**: En el run, el informe del job de evals de `jurisprudencia` da `cita_sin_resolver:<modelo>:<modo>` y `sin_activar:<modelo>:<modo>` cumplidos en los dos modos (medida 0), con `decide: true` y veredicto aprobado; los de `boe-legislacion` y `legal-core` siguen aprobados y `red` está vacío. **Control**: las comprobaciones `evals (jurisprudencia)`, `evals (boe-legislacion)` y `evals (legal-core)` de la propuesta de cambio, que el cierre del workflow cuenta; cualquiera en rojo deja el cierre sin aprobar.
- **SC-002**: Los once casos del e2e de FR-111 terminan cada uno con su código: 0 en cinco —las tres formas, `--roj` con su fecha y el ECLI del Tribunal Constitucional—, 3 en dos, 2 en dos y 5 en dos; y la repetición con `--offline`, aparte, con 0. **Control**: los guiones de aceptación, en `make ci`.
- **SC-003**: Ninguna petición al CENDOJ sale de `make ci`, de las evals ni de ningún flujo de la integración continua: 0. **Control**: el test de FR-091 en `make ci`, y `red` vacío en el informe del job (SC-001).
- **SC-004**: Una consulta hace como mucho un envío del formulario y una petición de la página del buscador, y `data` lleva como mucho 10 resoluciones. **Control**: los tests de FR-021 y FR-014, en `make ci`.
- **SC-005**: Tras una consulta, el resumen y la página de resultados aparecen 0 veces en la caché y en el grafo, y tras un «no encontrado» por la fecha, la caché no lleva ningún dato de la resolución encontrada con otra fecha que no esté en lo pedido (FR-012). **Control**: el test de FR-044, en `make ci`.
- **SC-006**: El único método distinto de GET y HEAD que sale del módulo es el `POST` del formulario declarado, a su dirección: cualquier otro termina en error de argumentos. **Control**: los tests de FR-112 y el test de arquitectura, en `make ci`.
- **SC-007**: Con sesiones sintéticas, una sola respuesta con un ECLI que no viene de una entrega ni de la pregunta da veredicto `fallo`; con ninguna, el umbral se cumple. **Control**: los tests de FR-115, en `make ci`.
- **SC-008**: Después del run, y fuera de él porque es humano: una persona lee las respuestas del modelo que decide a (b), (d) y (f) en los dos modos y anota en `docs/USO.md` si alguna resume o caracteriza una sentencia cuyo texto no tenía delante. Ninguna release lleva `jurisprudencia` hasta que H25 esté en `main`. Con esa release publicada, en la app de escritorio de Claude con el plugin: «¿existe la STS 1088/2023, de 4 de julio?» se responde con `STS 3144/2023`, `ECLI:ES:TS:2023:3144` y su enlace; y «resúmeme la STS 9999/2023, de 1 de enero» se responde diciendo que no se ha podido comprobar, sin resumen.
- **SC-009**: `make ci` termina en verde, con `schema-check` y `skills-check` sin diferencias, y `SKILL.md` de `jurisprudencia` tiene menos de 300 líneas. **Control**: `make ci`.

## Uso, de fuera adentro

Cada salida del hito, desde quien la consume (criterio de uso, ADR 0028). Lo que gana la skill: saber, antes de citar, si una sentencia existe, y con qué ECLI, ROJ y URL. Volumen de referencia: meses de uso diario, con cientos de sentencias comprobadas, la mayoría hace más de una semana.

Cálculos de esta sesión, no medidas —el verbo no existe todavía—. Una resolución con los metadatos de la STS 1088/2023 (`docs/JURISPRUDENCIA.md` §3 y `evidencias/adr-0036/`) y una URL de documento supuesta de 93 caracteres ocupa 318 bytes en JSON compacto; la procedencia de un sobre de `boe articulo` versionado en `web/src/data/sobres/`, 292. Una entrega de una resolución ronda los 700 bytes, y la mayor posible, de 10 (FR-014), los 3 600. Por herramienta, el sobre va dos veces en el mensaje (H21).

| Salida | Quién la pide, cuántas veces y qué hace con ella | Tamaño | Cuándo deja de darse cada señal |
|---|---|---|---|
| La entrega de `cita resolver` (FR-010) | La skill, en el paso de FR-061 y de una en una: una vez por sentencia que se da por su ECLI o su ROJ, y una o dos por sentencia escrita con número y fecha. De cada resolución toma el ECLI, el ROJ, el órgano, el número, la fecha y la URL para la cita (FR-062). Una persona, cuando la pide desde la terminal. | ≈ 700 bytes con una resolución; ≤ ≈ 3 600 con 10. Una respuesta que cita cinco sentencias: entre 5 y 10 invocaciones, ≤ ≈ 7 KB de sobres con una resolución cada una. No crece con lo acumulado: lo acota la referencia pedida y la página de 10. | Cada invocación es independiente. Una entrega se sirve de la caché 30 días (FR-040); después, la consulta vuelve a la fuente. |
| El «no encontrado» (FR-012) | La skill: con él no cita, y escribe la línea de FR-063; en una cita con número y fecha, antes prueba el ROJ con esa fecha (FR-061). | `{clase, mensaje}`, menos de 1 KB. | Se guarda 30 días (FR-041): deja de darse cuando vence la entrada y la consulta vuelve a la fuente, o cuando se pide la resolución por otra forma de referencia. |
| La declaración de página completa (FR-014) | La skill y la persona: saben que puede haber más resoluciones con ese número y esa fecha que las 10 que ven. | Un dato en `data`. | Solo se da con una página de 10; no se da con menos. |
| La declaración de cobertura del Tribunal Constitucional (FR-015) | La skill, que la traslada con FR-066; la persona. Una vez por ECLI del Constitucional que se pida. | Un sobre sin resoluciones, menos de 1 KB. | Con el candidato `tc` del backlog, que no es de este hito. No se guarda: se da en cada consulta porque responde a esa consulta. |
| Los fallos 4 y 5 (FR-016, FR-017) | La skill: no insiste y escribe la línea de FR-063 con el motivo. | `{clase, mensaje}`, menos de 1 KB. | En la primera consulta que la fuente responde. |
| La cita con su forma (FR-062) | Quien pregunta; una por sentencia resuelta que la respuesta nombra. | Una línea: la cita, del orden de 70 caracteres, y la URL, de unos 100. | No es una señal. |
| La línea `⚠ SENTENCIA NO COMPROBADA:` (FR-063) | Quien pregunta; una por referencia no comprobada de esa respuesta. Le dice qué referencia no se sostiene y por qué. | Una línea, del orden de 150 bytes. | No se repite en otra respuesta salvo que se vuelva a preguntar por esa referencia y siga sin comprobarse. |
| El nodo `Resolucion` (FR-042, FR-043) | Una persona, con `graph show <ECLI>`, uno por invocación; `graph stats` lo cuenta. La skill no lo pide. | Un nodo por resolución comprobada alguna vez: cientos en meses de uso, de ≈ 300 bytes de datos cada uno. `graph show` devuelve uno; `graph stats`, un recuento. | No da señales: `graph check` no da hallazgos de una `Resolucion`. |
| La tabla de comandos de `SKILL.md` (FR-060) | El modelo, una vez por conversación en que se activa la skill. | Una fila. `SKILL.md` < 300 líneas. | No da señales. |
| `umbrales` del informe de `jurisprudencia` (FR-080 a FR-083) | El job, que decide con ellos; el informe final del workflow, que los lee sin modelo; y la persona. Una vez por job. | 4 elementos de ≈ 250 bytes, sobre 18 respuestas por modo con tres repeticiones de las seis evals. Tamaño fijo. | Cada job los mide de nuevo sobre su commit. |

Tiempo, que también es coste de la skill, y cómo se pide: la skill lanza sus consultas de una en una, en los dos modos (FR-061), y una petición que falla no se reintenta (FR-021).

- **Por órdenes**, cada consulta que no está en la caché es un proceso nuevo que hace tres peticiones al sitio separadas 5 s —`robots.txt`, la página y el formulario (FR-020, FR-021)—, así que tarda al menos 10 s, dentro del plazo por defecto de 30 s de `--timeout`. Cinco sentencias sin caché son, como poco, 50 s de espera, y hasta el doble si cada una necesita su segundo intento; el ritmo no suma espera entre una orden y la siguiente (FR-024).
- **Por herramienta**, las llamadas comparten el ritmo del servidor (FR-024) y cada una tiene su propio plazo, 30 s por omisión (H21 FR 020). De una en una, cada llamada ocupa hasta tres turnos seguidos (FR-021) y tarda como mucho unos 15 s, dentro de su plazo: cinco sentencias, como mucho unos 75 s, y hasta el doble con el segundo intento. Lanzadas las cinco a la vez no cabrían: en 30 s hay seis turnos —a los 0, 5, 10, 15, 20 y 25 s— y las cinco consultas necesitan entre 10 y 15; los turnos se dan por orden de llegada, de modo que los cinco primeros se van en la primera petición de cada llamada y se resolverían, como mucho, una o dos sentencias. Las demás terminarían con 4 y la respuesta llevaría la línea de FR-063 para sentencias que existen: por eso FR-061 las pide de una en una. No se ha medido cómo reparte sus llamadas la app de escritorio: que puedan llegar a la vez sale de H21 FR 014.

## Fuera de alcance

Del hito, literal:

- «buscar por materia o por texto libre de forma automática;»
- «descargar, leer o guardar el texto o el PDF de una sentencia, y el resumen del CENDOJ;»
- «el filtro por órgano y el campo de número de recurso, sin probar;»
- «el Tribunal Constitucional, que llega por el BOE con el candidato `tc` del backlog;»
- «validar las citas de un texto, que es H8;»
- «el juez con modelo de `jurisprudencia` —su rúbrica, sus casos, su medida y su carpeta—, que es H25;»
- «consultar el CENDOJ desde el flujo nocturno o desde cualquier otro flujo de la integración continua;»
- «cambiar la constitución, `scripts/workflow/`, la fila de la fuente o `evidencias/adr-0036/`;»
- «cambiar `boe-legislacion` o `legal-core`;»
- «y la web, que la cambia la persona —antes de la release que lleve el hito, la página `/bot/` dice qué hace kitlegal con el CENDOJ, con qué agente y a qué ritmo (ADR 0036)—.»

De lo que el hito no especifica (constitución, «Criterio de decisión autónoma», punto 2):

- Otros verbos del applet `cita` y las citas de normas (H8).
- Resolver por número de recurso, por ponente o por cualquier otro campo del formulario; filtrar por jurisdicción o por tipo de órgano; pedir una segunda página de resultados.
- La equivalencia entre ECLI y ROJ calculada por el binario, y validar las siglas de un ROJ o el código de órgano de un ECLI contra una lista.
- Reconocer como del Tribunal Constitucional una referencia que no sea su ECLI, y dar en `data` la dirección de su buscador.
- Aristas del nodo `Resolucion`, nodos del órgano o del ponente, hallazgos de `graph check` sobre resoluciones y cualquier cambio en los verbos de `graph`.
- Una bandera para saltarse la caché o para cambiar su vigencia.
- Un mecanismo en el binario, en el empaquetado o en la release que excluya `jurisprudencia` hasta H25: que ninguna release la lleve antes lo decide la persona, que es quien etiqueta (ADR 0020).
- Evals de `jurisprudencia` distintas de las seis del hito: una de no activación, una sin binario ni servidor o una con varias resoluciones en la entrega.
- Una vigencia propia para el «no encontrado», distinta de los 30 días de una entrega, y cualquier modo de reproducción de grabaciones en el binario de producción.
- Que la skill o la herramienta elijan una resolución por su órgano.
- El ritmo de la fuente guardado en disco o compartido entre procesos: el de 5 s es el de cada invocación (FR-024). Y cualquier mecanismo del binario para las llamadas simultáneas a `cita_resolver` —una cola propia, otro plazo o atenderlas de una en una en el servidor—: que no lleguen a la vez es una regla de la skill (FR-061).
- Comprobar en las evals la parte legible de la cita, la URL o el motivo de la línea `⚠ SENTENCIA NO COMPROBADA:`.
- Umbrales del modelo informativo, un umbral de duración para `jurisprudencia` y cualquier umbral nuevo en las otras skills.
- Que el flujo nocturno avise de un cambio de forma del CENDOJ, y una incidencia automática para esa fuente.
- `CLAUDE.md` y `docs/ROADMAP.md`, que la persona actualiza al fusionar; un ADR nuevo; la release.

## Assumptions

- **«Resume» no se comprueba en este hito** (FR-085). FR-064 es una regla de la skill sin control: lo lee una persona en el cierre (SC-008) y lo decide el juez de H25. El informe final no puede darlo por comprobado.
- **Qué significa «ninguna release lleva `jurisprudencia` hasta H25»**. Se lee como una decisión de la persona sobre cuándo etiqueta (ADR 0020), no como un mecanismo: al fusionar el hito, la skill queda empotrada en el binario de `main`, como toda skill. Rechazada: excluirla del binario o del plugin, que el hito no pide.
- **`--fecha` con un ECLI es un error de argumentos** (FR-005). El hito la define con `--resolucion` y con `--roj`; con un ECLI, que ya es único, no la define, y lo no especificado no se implementa. Rechazada: admitirla y filtrar.
- **La página completa se declara** (FR-014). El hito dice a la vez «los da todos» y «ni más de una página de resultados». Con 10 resultados no se sabe si hay más sin pedir otra página, que el hito prohíbe; decirlo es lo que menos añade sin dejar que diez pasen por todos. Si la página del buscador da el total, el plan puede usarlo para la misma declaración.
- **Lo que cuenta como entregado** (FR-081). El hito dice «la salida de ninguna consulta `cita resolver`» y, a renglón seguido, que «un ECLI no puede llegar a una respuesta más que desde una entrega o desde la persona». Si contara cualquier texto de la salida, bastaría pedir un ECLI inventado para que el mensaje del fallo lo «entregara». Se lee por la segunda frase: las resoluciones de `data`.
- **La forma del ECLI** (FR-003) sigue el estándar europeo que cita `docs/JURISPRUDENCIA.md` §3 (`DOUE-Z-2019-70039`): código de órgano de hasta 7 caracteres y número de orden de hasta 25. Esas dos cotas se escriben de memoria: el documento no se ha leído en esta sesión, que no tiene red, y el plan las deja como supuesto no verificado si no puede comprobarlas en local. Los ECLI de los autos españoles terminan en letra, y caben.
- **La forma del ROJ** (FR-004) sale de los ejemplos del repositorio (`STS 3144/2023`) y de que las siglas de otros órganos llevan una segunda palabra. `docs/JURISPRUDENCIA.md` §3 deja sin probar los órganos distintos del Supremo: la herramienta no valida siglas.
- **La dirección del buscador del Tribunal Constitucional** (FR-066) no está escrita en el repositorio: `docs/JURISPRUDENCIA.md` §2 nombra el buscador HJ y sus rutas, no su dirección. Esta sesión no tiene red y no la ha comprobado. El plan la fija en un solo sitio, y queda como supuesto para que la persona la compruebe al leer el informe final.
- **Los operadores del buscador del CENDOJ** (FR-065) tampoco están descritos en el repositorio. La skill solo nombra un operador o un valor de jurisdicción u órgano que conste en `docs/JURISPRUDENCIA.md` o en la página del buscador grabada; lo que no conste, no se escribe.
- **La evaluación de (c) y (e) por una dirección** (FR-070). El hito dice que (c) «prepara la consulta para el buscador» y que (e) «la declara no cubierta», y que todo se compara por su forma, sin modelo. Lo que tiene forma en esas dos respuestas es la dirección del buscador que el protocolo manda dar y que no haya ninguna cita.
- **La consulta en las evals** (FR-070). Que (a), (b), (d) y (f) exijan haber consultado `cita resolver` sale del protocolo: sin la consulta, (f) no mediría que la fecha la compara la herramienta.
- **El mensaje del «no encontrado» nombra lo pedido y nada más de la sentencia encontrada** (FR-012). El hito dice que la resolución de otra fecha «no entra en `data`»; si sus datos entraran en el mensaje, llegarían igual a la respuesta. Solo se llega a ese caso con `--roj` y `--fecha`, y ahí el ROJ pedido es también el de la encontrada: el mensaje lo lleva porque es la referencia que se pidió.
- **El año de una cita con día y mes** («STS 1088/2023, de 4 de julio») es el del número, salvo que la cita diga otro. Lo aplica la skill.
- **Las grabaciones** llevan la página de resultados de unas pocas resoluciones, con el resumen del CENDOJ, en un repositorio público: lo asume el ADR 0036 («En contra, y asumido»), y FR-092 las mantiene en lo mínimo. FR-040 y FR-044 hablan de la caché y del grafo, no de `testdata/`.
- **El paso `grabar_datos` graba antes de implementar**, de una fuente con fila revisada en `main` (2026-10-03). Si el CENDOJ lo bloquea, el run se detiene con su dossier (causa mayor: dependencia externa inaccesible) y decide una persona.
- **Lo comprobado en esta sesión**: lo leído en el repositorio —la constitución 2.12.0, los ADR 0023, 0036 y 0037, `docs/JURISPRUDENCIA.md`, `docs/SOURCES.md`, `evidencias/adr-0036/`, `internal/httpx`, `internal/evals`, `internal/core/grafo`, las dos skills y sus evals— y los dos tamaños de «Uso, de fuera adentro», calculados con `wc -c`. No se ha ejecutado ningún test ni `make ci`, ni se ha consultado ninguna fuente.
- **`references/` de la skill.** El hito no da a `jurisprudencia` ningún dato de `data/`: si su carpeta `references/` existe o no lo decide el plan con lo que exija `skills-check`, sin datos nuevos en `data/`.
- **Quedan para el plan**: los nombres de las claves de `data` y de los argumentos de la herramienta, la forma de las dos declaraciones (FR-014, FR-015), el paquete del reconocimiento de ECLI y de ROJ, cómo declara una fuente su formulario, cómo se nombran las grabaciones de un envío, los campos nuevos del formato de eval y la redacción de `SKILL.md`.
