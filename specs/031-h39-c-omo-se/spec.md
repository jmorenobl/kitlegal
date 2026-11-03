# Feature Specification: H39 · Cómo se organiza una skill para que se active y se lea en los dos modos

**Feature Branch**: `031-h39-c-omo-se`

**Created**: 2026-11-03

**Status**: Draft

**Input**: el fichero del hito, `docs/roadmap/in-progress/H39-como-se-organiza-una-skill.md` (entrada de la persona, ADR 0018).

## Contexto

Lo que el hito quiere saber, medido, es qué organización de una skill y del servidor MCP hace que el agente active la
skill antes de la primera herramienta de kitlegal y lea y cumpla lo que va fuera del cuerpo, en **modo orden** (el
agente ejecuta `kitlegal <applet> <verbo>`) y en **modo herramienta** (el agente llama a `<applet>_<verbo>` del
servidor `kitlegal mcp serve`), antes de reorganizar las skills del producto. Los cuerpos están en el techo del ADR 0046
(`normativa`, 10 927 caracteres de un tope real de 10 930) porque cada defecto medido se ha arreglado con una frase
más en el cuerpo, y lo que va a una referencia, en modo herramienta, no se lee.

Evidencia de partida, tal como la da el hito (no medida en esta sesión): las 440 sesiones de los sondeos de la #214
(`specs/030-h32-arquitectura-de-skills/research.md` D26 a D30; `evidencias/sondeos-20261101/`). En `normativa`, en modo
herramienta, `ToolSearch` va en el mismo turno que `Skill` en 240 de 240 sesiones; `normas.md` se abre en 11, 11 y 8
de 80 (en modo orden, 79, 77 y 78) y `con_herramientas.md` en 4, 3 y 2 de 80 (39, 26 y 41). En `plazos-y-territorio`,
cuya descripción no nombra ninguna herramienta, `administrativo.md` se abre en 58 de 60 en modo herramienta. La suite
completa de la release v0.9.0 (ejecución 41474010633, sobre `cd61442`) sale en rojo solo en modo herramienta: eval 04
de `normativa`, `sin_activar` 1 de 60, y eval 02 de `consulta-juridica`, 1 de 3 con umbral 2.

Comprobado en esta sesión, en el repositorio: los argumentos de los verbos declaran su ayuda (28 textos en
`internal/app/{boe,cita,grafo,instalacion,plazos,territorio}.go`, 2373 caracteres en total, el más largo de 146), y, antes de
este hito, las propiedades de `entrada` de `schemas/bloque.json`, generado desde `--describe`, no llevaban
`description`.

## Clarifications

### Session 2026-11-03

- Q: En modo herramienta, ¿cuenta una llamada a ToolSearch que carga el esquema de una herramienta de kitlegal como «la primera herramienta de kitlegal» a efectos de FR-041? → A: No. La primera herramienta de kitlegal es el primer `tool_use` que llama a una herramienta del registro `<applet>_<verbo>`; `ToolSearch` no la llama y no cuenta; no se añade ninguna columna (auto: criterio a; fuente: hito H39, Objetivo y Entrega; spec.md Contexto y FR-041; internal/evals/sesion.go; internal/app/herramientas.go)
- Q: En modo orden, ¿qué orden de Bash cuenta como «la primera orden kitlegal» de FR-041? → A: La primera que invoca el binario `kitlegal` con un verbo de un applet que da herramienta del servidor (no `mcp` ni `skills`) y que consulta (sin ayuda, `--describe` ni `--dry-run`), termine como termine; `which kitlegal`, `--help`, `--version`, `--describe`, `skills list` y un `grep` no cuentan (auto: criterio c; fuente: hito H39; internal/evals/trazas.go `Invocacion.Consulta`; internal/app/herramientas.go; internal/evals/juzgar.go)
- Q: ¿Qué cuenta como «activar la skill» de laboratorio (y la de entrada) en FR-041: solo una llamada a la herramienta Skill con su nombre, o también leer su SKILL.md con otra herramienta? → A: Solo un `tool_use` de `Skill` cuyo `input.skill` es su nombre, como `Sesion.Activada`; leer el `SKILL.md` con `Read` o `Bash` no es activarla (auto: criterio a; fuente: internal/evals/sesion.go `Activada`; hito H39, Evidencia de partida, Controles y Objetivo; spec.md FR-041)
- Q: ¿Qué cuenta como «abrir la referencia» en las variantes con referencia (FR-041), y con ello cuál es el turno de su primera apertura? → A: Un `tool_use` de `Read` con el `file_path` de la referencia, entero o por tramos, o de `Bash` cuya orden nombra su fichero; `Grep` y `Glob` no cuentan; el turno es el del primer bloque que cumple y abierta varias veces cuenta una (auto: criterio c; fuente: hito H39, Objetivo y Aceptación; spec.md FR-041 y Edge Cases; evidencias/sondeos-20261025/guiones/analizar.py)
- Q: ¿Cuántas evals de laboratorio tiene el sondeo entre las que FR-030 reparte por igual las 40 sesiones de cada variante y modo? → A: Cuatro, todas sobre la misma herramienta (la que lleva la regla en la descripción de su parámetro en `parametro`) y con lecturas que ya sirve `testdata/evals/`: 10 sesiones de cada una por variante y modo; el número de sesiones pedido ha de ser múltiplo de 4 o el arnés sale con un código distinto de 0 sin abrir ninguna sesión (auto: criterio c; fuente: hito H39, Entrega; spec.md FR-012, FR-024, FR-028 y FR-030; testdata/evals/boe.legislacion-consolidada/)

## User Scenarios & Testing *(mandatory)*

### User Story 1 - La ayuda de cada argumento llega a su herramienta (Priority: P1)

En modo herramienta, el agente lee el esquema de entrada de una herramienta de kitlegal cuando la carga, y hoy no
encuentra en él lo que la ayuda de cada argumento dice (por ejemplo, que el bloque de `boe_articulo` sale del campo `id`
de `boe_indice` y no se compone con el número del artículo). Con este hito, la ayuda de cada argumento de un verbo es
la descripción de su parámetro en `--describe` y en la herramienta MCP, para todas las skills del producto y para la
variante del sondeo que pone la regla en un parámetro.

**Why this priority**: es un defecto de la base conocido desde H32, común a todas las herramientas, y la variante
`parametro` (FR-024) no se puede medir sin él.

**Independent Test**: `kitlegal boe articulo BOE-A-2015-10565 a21 --describe` y la lista de herramientas de `kitlegal mcp serve` llevan, en
cada parámetro con ayuda, esa ayuda como descripción; ningún `SKILL.md` cambia.

**Acceptance Scenarios**:

1. **Dado** el verbo `boe articulo`, cuyo argumento `bloque` declara la ayuda «Id del bloque, del campo id de kitlegal
   boe indice (a21, da-3, a1-30…): no se compone con el número del artículo, que no siempre es su id.», **Cuando** se
   pide `kitlegal boe articulo BOE-A-2015-10565 a21 --describe` (sin sus dos argumentos sale con 2: el análisis va
   antes que la descripción), **Entonces** la propiedad `bloque` del esquema de entrada lleva como
   `description` exactamente ese texto (FR-001).
2. **Dado** el servidor `kitlegal mcp serve`, **Cuando** un cliente pide la lista de herramientas, **Entonces** el
   `inputSchema` de `boe_articulo` lleva en `bloque` la misma `description`, y lo mismo cada parámetro con ayuda de
   cada herramienta (FR-002).
3. **Dado** un argumento de un verbo sin ayuda declarada, **Cuando** se describe su verbo, **Entonces** su propiedad no
   lleva `description` (FR-003).
4. **Dado** el árbol con el cambio, **Cuando** corre `make ci`, **Entonces** `schemas/` coincide con lo que emite
   `--describe`, las tablas de comandos de `skills/*/SKILL.md` salen idénticas byte a byte y el diff del hito no toca
   `skills/` (FR-004).

---

### User Story 2 - Una skill de laboratorio con su regla testigo y sus variantes (Priority: P1)

Para medir dónde hay que poner algo para que el agente lo cumpla, hace falta una skill que no sea del producto, con una
regla que el modelo no pueda cumplir de memoria y que un guion compruebe sin modelo, y un juego de variantes que
difieran de la de control en una sola cosa cada una, con las mismas evals.

**Why this priority**: es el objeto de la medida; sin ella el sondeo no tiene qué comparar.

**Independent Test**: `make ci` comprueba la regla testigo contra respuestas fijas que la cumplen y que no, que cada
variante difiere de la de control solo en lo que declara, y que nada de la skill de laboratorio llega al binario ni a
`kitlegal skills install`.

**Acceptance Scenarios**:

1. **Dado** una respuesta fija que cumple la regla testigo, **Cuando** la comprueba el guion de la regla, **Entonces**
   dice que cumple; **Dado** una que no la cumple —entre ellas la respuesta de una skill del producto a la misma
   pregunta—, **Entonces** dice que no (FR-013, SC-003).
2. **Dado** cada variante de FR-020 a FR-027, **Cuando** `make ci` la compara con la de control, **Entonces** todo lo
   que no está en la declaración de la variante es idéntico byte a byte, y lo declarado difiere (FR-028, SC-004).
3. **Dado** el binario construido del árbol, **Cuando** se listan las skills que lleva dentro o se ejecuta
   `kitlegal skills install` en un directorio vacío, **Entonces** no aparece ningún fichero de la skill de laboratorio
   ni de su skill de entrada (FR-010, SC-005).

---

### User Story 3 - El arnés del sondeo y el guion que lo cuenta (Priority: P1)

La persona lanza el sondeo de las variantes por modo, con el modelo y la versión de Claude Code que elige, y obtiene una
tabla por variante y modo con las sesiones que activan la skill antes de la primera herramienta, las que abren la
referencia y en qué turno, y las que cumplen la regla testigo, cada cifra con su intervalo de Wilson al 95 %.

**Why this priority**: es la medida que decide el ADR; sin el arnés y el guion, la aceptación del hito no se puede hacer.

**Independent Test**: en `make ci`, sin modelo, el guion da las cifras exactas de unas sesiones fijas de los dos modos,
y el arnés con una versión de Claude Code que no es la pedida sale sin abrir ninguna sesión.

**Acceptance Scenarios**:

1. **Dado** un juego fijo de transcripciones de sesiones de los dos modos, **Cuando** el guion las cuenta, **Entonces**
   da por variante y modo los recuentos de FR-041 y sus intervalos de Wilson con z = 1,96: 0 de 40 → 0,0–8,8 %;
   20 de 40 → 35,2–64,8 %; 40 de 40 → 91,2–100,0 % (SC-006).
2. **Dado** una sesión en la que `Skill` de la skill de laboratorio y la primera herramienta de kitlegal van en el mismo
   turno, **Cuando** el guion la cuenta, **Entonces** no cuenta como activada antes de la primera herramienta (FR-041).
3. **Dado** que la versión de Claude Code instalada no es la que se pasa al arnés, o que el modelo pasado no es un
   identificador completo, **Cuando** la persona lo lanza, **Entonces** sale con un código distinto de 0, dice qué no
   coincide y no abre ninguna sesión (FR-031, SC-007).
4. **Dado** que se pide una variante que solo se mide en modo herramienta (`instrucciones`, `parametro`) en modo
   orden, o un número de sesiones que no es múltiplo de 4, **Cuando** la persona lanza el arnés, **Entonces** sale con
   un código distinto de 0 sin abrir ninguna sesión (FR-030, SC-007).

---

### User Story 4 - La persona mide y decide antes de fusionar (Priority: P2)

Con la propuesta de cambio abierta, la persona lanza el sondeo con `claude-sonnet-5-5`, 40 sesiones por variante y modo,
deja su tabla en `evidencias/` y escribe un ADR que precisa el 0046 con esas cifras. La propuesta no se fusiona sin los
dos.

**Why this priority**: es la aceptación del hito, pero la hace una persona fuera del run (ADR 0032): el run entrega lo
que la hace posible.

**Independent Test**: el informe final del run dice que la propuesta queda pendiente del sondeo y del ADR; antes de
fusionar, `evidencias/` tiene la tabla y `docs/ADR/` el ADR.

**Acceptance Scenarios**:

1. **Dado** el arnés y el guion fusionados en la rama, **Cuando** la persona lanza el sondeo con `claude-sonnet-5-5` y
   40 sesiones por variante y modo, **Entonces** la tabla, con el modelo y la versión de Claude Code anotados, queda
   en `evidencias/` (FR-050).
2. **Dado** la tabla, **Cuando** la persona escribe el ADR, **Entonces** el ADR precisa el 0046 y dice, con esas cifras,
   dónde va en una skill, y en el servidor, lo que hace que el agente la active y cumpla lo que pide en cada modo; si
   ninguna referencia se lee de forma fiable en modo herramienta, lo dice y propone dónde va lo imprescindible (FR-051).

---

### Edge Cases

- **Sesión sin ninguna herramienta de kitlegal**: si activó la skill, cuenta como activada antes de la primera
  herramienta; si no la activó, no (FR-041).
- **Sesión que no termina o termina sin respuesta** (límite de tiempo, error del cliente): el guion la da en su propia
  columna y no entra en el denominador de ninguna cifra (FR-042).
- **Referencia abierta varias veces**: cuenta el turno de la primera apertura (FR-041).
- **Variante sin referencia** (`control`, `instrucciones`, `entrada`, `parametro`, `descripcion-sin-herramientas`):
  las columnas de la referencia salen vacías, no a cero (FR-041).
- **Variante con referencia que ninguna sesión abre**: las aperturas salen «0 de n» con su intervalo, y las columnas
  del turno de primera apertura, vacías (FR-041, FR-043).
- **Fila sin ninguna sesión terminada** (n = 0): la fila sale con su columna de sesiones sin terminar y los recuentos,
  porcentajes e intervalos vacíos (FR-042, FR-043).
- **Transcripción que el guion no sabe leer** (manipulada a mano, truncada por algo que no es el arnés): defecto
  `inesperado`, código 1 (ADR 0023), sin más casos.

## Requirements *(mandatory)*

### Functional Requirements

**La descripción de los parámetros (la base)**

- **FR-001**: Todo argumento o bandera que un verbo de un applet declara con una ayuda aparece en el esquema de entrada
  que emite `kitlegal <applet> <verbo> … --describe` (con sus argumentos obligatorios) con una `description` igual, carácter a carácter, a esa ayuda.
- **FR-002**: El `inputSchema` de cada herramienta `<applet>_<verbo>` que anuncia `kitlegal mcp serve` lleva, en cada
  parámetro que corresponde a un argumento con ayuda, la misma `description` que FR-001.
- **FR-003**: Un argumento sin ayuda declarada no gana `description`; ninguna ayuda se reescribe, se acorta ni se
  completa al pasar al esquema; y nada más del esquema de entrada ni del de salida cambia.
- **FR-004**: `schemas/*.json` se regenera desde `--describe` y `make schema-check` sale en verde; las tablas de
  comandos de `skills/*/SKILL.md`, que salen del mismo `--describe`, no ganan columnas ni texto: la comprobación de
  deriva de `make ci` falla si el generador las cambia, y el diff del hito no toca `skills/`.
- **FR-005**: `CHANGELOG.md` (*Unreleased*) da el cambio de los esquemas: los parámetros de `--describe` y de las
  herramientas MCP llevan la ayuda de su argumento.

**La skill de laboratorio**

- **FR-010**: La skill de laboratorio, su skill de entrada (FR-022), sus variantes y sus evals viven fuera de `skills/`
  y de `evals/` y de todo lo que el binario empotra o `kitlegal skills install` copia: el binario no los contiene, la
  instalación no los pone y el job de evals no los elige ni los mide. `make ci` falla si cualquiera de las tres cosas
  deja de cumplirse.
- **FR-011**: La skill de laboratorio y su skill de entrada cumplen la forma que `make ci` exige a las del producto
  (frontmatter de la especificación Agent Skills, `references/` plano y enlazado desde `SKILL.md`, cuerpo de como mucho
  11 000 caracteres y descripción de como mucho 900, ADR 0046); `make ci` falla si una variante no la cumple.
- **FR-012**: Sus evals son cuatro, todas sobre la misma herramienta de kitlegal y sobre lecturas que ya están grabadas
  para las evals de hoy (`testdata/evals/`), las mismas para todas las variantes, y ninguna pide una lectura que la
  grabación no sirve: `make ci` falla si una eval de laboratorio espera una orden que la grabación no tiene. El run no graba nada: no añade ni
  cambia ninguna grabación y no escribe en `evidencias/`.
- **FR-013**: La skill lleva una **regla testigo**: una regla de forma sobre la respuesta, inventada para el
  laboratorio —no aparece en ninguna skill, referencia ni instrucción del producto—, que un guion decide sin modelo,
  cumple o no cumple, sobre el texto de la respuesta final de una sesión. La respuesta de una skill del producto a una
  pregunta de las evals de laboratorio no la cumple (FR-014).
- **FR-014**: `make ci` comprueba el guion de la regla testigo con respuestas fijas: al menos una que la cumple y al
  menos una por cada forma de no cumplirla que la regla distingue, más una respuesta de una skill del producto a una
  pregunta de las evals de laboratorio; el test falla si alguna sale con el veredicto contrario.

**Las variantes**

Cada variante es la de control con una sola diferencia, que declara: qué partes cambian (cuerpo, una referencia, la
descripción, una skill más, las instrucciones del servidor de la sesión, la descripción de un parámetro de la sesión).
Las ocho, y en qué modos se miden:

- **FR-020**: `control` — la regla testigo va en el cuerpo de `SKILL.md`; la descripción nombra las herramientas, como
  hoy la de `normativa`. Ambos modos.
- **FR-021**: `instrucciones` — la skill es la de control; las instrucciones del servidor MCP de las sesiones del sondeo
  piden activar la skill antes de la primera herramienta. Solo modo herramienta.
- **FR-022**: `entrada` — la skill es la de control, y se instala además una skill de entrada que remite a la de
  laboratorio, como `consulta-juridica`. Ambos modos.
- **FR-023**: `referencia-secundaria` — la regla sale del cuerpo y va a una referencia que el cuerpo manda leer en una
  frase secundaria, como hoy. Ambos modos.
- **FR-024**: `parametro` — la regla sale del cuerpo y va a la descripción del parámetro de la herramienta que piden
  las evals (la misma en las cuatro, FR-012), en el servidor de las sesiones del sondeo. Solo modo herramienta.
- **FR-025**: `referencia-paso` — la regla sale del cuerpo y va a una referencia que se lee como paso propio antes de la
  primera llamada. Ambos modos.
- **FR-026**: `referencia-en-el-punto` — la regla sale del cuerpo y va a una referencia que el cuerpo manda leer en el
  punto en que hace falta, como `administrativo.md`. Ambos modos.
- **FR-027**: `descripcion-sin-herramientas` — la descripción de la skill no nombra ninguna herramienta; lo demás, el
  control. Ambos modos.
- **FR-028**: `make ci` compara cada variante con la de control: falla si algo fuera de lo que la variante declara
  difiere en un byte, o si lo declarado no difiere.
- **FR-029**: Las instrucciones y las descripciones de parámetros distintas de FR-021 y FR-024 existen solo en las
  sesiones del sondeo. El binario que se distribuye anuncia las instrucciones de hoy (las que fija `TestInstrucciones`)
  y las ayudas de hoy como descripciones, y ninguna bandera, variable de entorno ni orden suya las cambia.

**El arnés del sondeo**

- **FR-030**: El arnés lanza, para las variantes y los modos que la persona pide, el número de sesiones por variante y
  modo que pide, repartidas por igual entre las cuatro evals de laboratorio (diez por eval con las 40 de la aceptación), en los dos modos de las evals del producto y con
  las sesiones servidas por la grabación, sin red a las fuentes (ADR 0043). Cada sesión tiene instalada solo la
  skill de laboratorio de su variante y, en `entrada`, su skill de entrada: ninguna skill del producto. Pedir una variante en un modo que FR-020 a
  FR-027 no le dan, o un número de sesiones que no es múltiplo de 4, es un error de argumentos: sale con un código
  distinto de 0 sin abrir ninguna sesión.
- **FR-031**: El arnés recibe el modelo y la versión de Claude Code y, antes de abrir ninguna sesión, comprueba que la
  versión instalada es la pedida y que el modelo es un identificador completo (`claude-<familia>-<versión>`, como
  `claude-sonnet-5-5`), no un alias como `sonnet`; si no, sale con un código
  distinto de 0, nombra lo que no coincide y no abre ninguna. Abre cada sesión con ese modelo y anota los dos en la
  tabla.
- **FR-032**: El arnés lo lanza una persona, con la credencial del sondeo de hoy (ADR 0032); ni `make ci`, ni un gancho,
  ni el job de evals, ni un paso del workflow lo ejecutan. Ni el arnés ni el guion van en `evidencias/`.
- **FR-033**: El arnés conserva la transcripción de cada sesión, con su variante, su modo y su eval, donde el guion la
  lee.

**El guion que cuenta el sondeo**

- **FR-040**: El guion lee las transcripciones de un sondeo sin red ni modelo y da una tabla con una fila por variante y
  modo, con el modelo y la versión de Claude Code de FR-031.
- **FR-041**: Por fila: las sesiones que activan la skill de laboratorio antes de la primera herramienta de kitlegal
  (en un turno anterior al de la primera llamada a una herramienta de kitlegal en modo herramienta, o a la primera orden
  `kitlegal` en modo orden; en el mismo turno no cuenta); en `entrada`, además, las que activan la skill de entrada
  antes de la primera herramienta y las que activan las dos antes de la primera herramienta; en las variantes con referencia, las que la abren, y el
  turno de su primera apertura (en la tabla, la mediana —con un decimal si cae entre dos turnos— y los extremos; el de
  cada sesión, en el detalle); y las que cumplen la
  regla testigo según el guion de FR-013. Cada recuento va como «k de n» con su porcentaje y su intervalo de Wilson al
  95 % (z = 1,96), en porcentaje con un decimal. Qué cuenta en cada caso, igual en todas las filas:
  - **Turno**: una respuesta del modelo, es decir, los bloques que emite antes de recibir el resultado de ninguna
    herramienta, vengan en uno o en varios mensajes `assistant` de la transcripción; se numeran desde 1 en el orden de
    la sesión. «El turno de un bloque» es el de la respuesta que lo emite.
  - **Activar una skill** (la de laboratorio y la de entrada): un bloque `tool_use` de la herramienta `Skill` cuyo
    `input.skill` es su nombre. Leer su `SKILL.md`, con `Read` o con `Bash`, no la activa.
  - **Primera herramienta de kitlegal, en modo herramienta**: el primer bloque `tool_use` que llama a una herramienta
    del registro `<applet>_<verbo>` (por su nombre o por lo que sigue a su último «__»), termine como termine.
    `ToolSearch` carga el esquema de una herramienta y no la llama: no cuenta. `Skill` y `ToolSearch` en el mismo turno
    y la primera llamada al registro en uno posterior cuenta como activada antes; `Skill` y esa llamada en el mismo
    turno, no.
  - **Primera orden `kitlegal`, en modo orden**: el primer bloque `tool_use` de `Bash` cuya orden invoca el binario
    `kitlegal` (como orden, o dentro de una tubería o una secuencia) con un verbo de un applet que da herramienta del
    servidor (no `mcp` ni `skills`) y que consulta: sin ayuda, sin `--describe` y sin `--dry-run`. Termine como
    termine. No cuentan `which kitlegal`, `kitlegal --help`, `kitlegal --version`, `kitlegal boe articulo --describe`,
    `kitlegal skills list` ni un `grep` que nombra la palabra. El turno es el del bloque de `Bash` en la transcripción.
  - **Abrir la referencia**: un bloque `tool_use` de `Read` cuyo `file_path` es el de la referencia, entero o por
    tramos (`offset` y `limit`), o de `Bash` cuya orden nombra su fichero (`references/<fichero>.md` de la skill de
    laboratorio). `Grep` y `Glob` no cuentan. El turno de la primera apertura es el del primer bloque que cumple.
  - **Las sesiones de una fila** son las de las evals de laboratorio de FR-012, repartidas por igual (FR-030).
- **FR-042**: Las sesiones que no terminan o terminan sin respuesta final van en su propia columna y no entran en `n`;
  con `n` = 0, los recuentos, porcentajes e intervalos de la fila salen vacíos.
- **FR-043**: `make ci` comprueba el guion con transcripciones fijas de los dos modos que cubren cada columna, incluidos
  `Skill` y la primera herramienta en el mismo turno (también en mensajes separados de la transcripción), una
  referencia abierta dos veces, una variante con referencia que ninguna sesión abre, una sesión sin terminar y una fila
  sin ninguna sesión terminada, y con
  los intervalos de la historia 3; el test falla si un recuento o un extremo de un intervalo difiere.

**La aceptación de la persona y las evals del producto**

- **FR-050**: Antes de fusionar, la persona lanza el sondeo con `claude-sonnet-5-5`, 40 sesiones por variante y modo
  (14 filas: seis variantes en los dos modos y dos solo en modo herramienta; 560 sesiones), y deja su tabla en
  `evidencias/`.
- **FR-051**: Antes de fusionar, la persona escribe un ADR que precisa el 0046 y dice, con las cifras de FR-050, dónde
  va en una skill, y en el servidor, lo que hace que el agente la active y cumpla lo que pide en cada modo; si ninguna
  referencia se lee de forma fiable en modo herramienta, lo dice y propone dónde va lo imprescindible.
- **FR-052**: El run no lanza el sondeo ni escribe el ADR; los dos se añaden a la propuesta de cambio antes de
  fusionarla, y la propuesta no se fusiona sin ellos. El informe final del run lo dice.
- **FR-053**: Las evals de las skills del producto salen en verde en la medición del cierre; el cambio de los esquemas
  es común, así que el job las elige todas. Si el cierre da los rojos de partida en modo herramienta (la eval 04 de
  `normativa`, `sin_activar`, y la eval 02 de `consulta-juridica`, como en la ejecución 41474010633), la reparación no
  los arregla con texto nuevo en las skills del producto ni reorganizándolas, y el informe los da con sus cifras para
  que decida una persona.

### Key Entities

- **Skill de laboratorio**: skill pequeña con la forma de las del producto, fuera de él, que lleva la regla testigo;
  una por variante.
- **Skill de entrada de laboratorio**: la de FR-022, que remite a la de laboratorio.
- **Variante**: la skill de control con una diferencia declarada (partes que cambian) y los modos en que se mide.
- **Eval de laboratorio**: una pregunta que activa la skill y una o varias órdenes servidas por la grabación de hoy;
  son cuatro, sobre la misma herramienta (FR-012).
- **Regla testigo** y su guion: regla de forma inventada; el guion da cumple o no cumple sobre una respuesta.
- **Sondeo**: las sesiones de las variantes por modo, con su modelo y su versión de Claude Code.
- **Tabla del sondeo**: una fila por variante y modo con los recuentos de FR-041 y sus intervalos.

### Uso, de fuera adentro

- **Descripción de los parámetros** (FR-001, FR-002). La pide el agente en modo herramienta al cargar una herramienta
  (`ToolSearch`), una vez por herramienta y sesión, y la lee en el punto en que va a llamarla: es donde la regla del
  bloque de `boe_articulo` («no se compone con el número del artículo») llega a la llamada que la eval 04 hizo sin el
  índice. Tamaño: las 28 ayudas de todos los verbos suman 2373 caracteres y la mayor tiene 146, así que lo que añaden
  los esquemas de las herramientas que carga una pregunta está acotado por 2373 caracteres aunque las cargue todas.
  Es texto fijo del binario: no crece con cientos de normas ni miles de bloques consultados, y solo
  cambia cuando cambia la ayuda de un argumento. No es una señal: no avisa de nada que caduque.
- **Tablas de comandos de las skills** (FR-004). Las lee el agente al activar una skill, una vez por sesión; no cambian.
- **Veredicto de la regla testigo** (FR-013). Lo pide el guion de FR-041, una vez por sesión del sondeo: 560 veredictos
  de una palabra. Solo existe en el sondeo.
- **Tabla del sondeo** (FR-040, FR-041). La lee la persona una vez, para escribir el ADR: 14 filas de unas diez
  columnas, más el detalle de 560 sesiones. No crece con el uso del producto; deja de leerse cuando el ADR está escrito,
  y un sondeo nuevo da otra tabla.
- **Skill de laboratorio**. Solo la lee el agente en las sesiones del sondeo; ninguna persona ni skill del producto la
  ve.

## Success Criteria *(mandatory)*

### Measurable Outcomes

- **SC-001**: El 100 % de los argumentos con ayuda de todos los verbos (28 hoy) llevan esa ayuda como descripción en
  `--describe` y en la herramienta MCP, y el 0 % de los que no tienen ayuda gana una. Control: un test de `make ci`
  que recorre todos los verbos del registro y falla con el primer argumento que no cumple.
- **SC-002**: 0 bytes de diferencia en `skills/` y en las tablas de comandos regeneradas. Control: la comprobación de
  deriva de las skills en `make ci` falla si el generador cambia una tabla; el guardián de diff del workflow, si una
  tarea toca `skills/`.
- **SC-003**: El guion de la regla testigo acierta el 100 % de las respuestas fijas de FR-014. Control: su test en
  `make ci` falla con la primera que no acierta.
- **SC-004**: Las siete variantes difieren de la de control en 0 bytes fuera de lo que declaran, y en más de 0 dentro.
  Control: el test de FR-028 en `make ci`.
- **SC-005**: 0 ficheros de laboratorio en el binario, en lo que instala `kitlegal skills install` y en las evals que
  elige el job. Control: el test de FR-010 en `make ci`.
- **SC-006**: El guion da los recuentos exactos y los extremos de Wilson con un decimal de las transcripciones fijas de
  FR-043. Control: su test en `make ci`.
- **SC-007**: Con la versión de Claude Code distinta de la pedida, un modelo que no es un id completo, una variante en
  un modo que no le corresponde o un número de sesiones que no es múltiplo de 4, el arnés abre 0 sesiones; con una
  petición válida, abre n/4 por eval en cada variante y modo pedidos. Control: un test de `make ci` con un sustituto de
  `claude` que cuenta sus invocaciones, sin modelo, falla si una petición inválida abre alguna sesión o si una válida no
  abre n/4 por eval en cada variante y modo.
- **SC-008**: Las evals de las skills del producto en verde en la medición del cierre. Control: el job de evals del
  cierre sale en rojo si una no pasa su umbral, y el informe final da sus cifras (FR-053).
- **SC-009**: Antes de fusionar, la tabla de las 14 filas de 40 sesiones está en `evidencias/` y el ADR en `docs/ADR/`.
  Lo comprueba la persona (capa 3): ningún control automático lo mide, y el run no lo hace (FR-052).

## Criterios del hito (literales)

**Controles**: `make ci` comprueba la regla testigo con respuestas fijas que la cumplen y que no (FR-014, SC-003), que
las variantes difieren de la de control solo en lo declarado (FR-028, SC-004), que la skill de laboratorio no se
empotra ni se instala (FR-010, SC-005), y la descripción de los parámetros en `--describe` y en las herramientas MCP
(FR-001, FR-002, SC-001). Las evals de las skills del producto en verde: el cambio de los esquemas es común y las mide
todas (FR-053, SC-008). Los rojos de partida en modo herramienta (la eval 04 de `normativa`, `sin_activar`, y la eval 02
de `consulta-juridica`, en la ejecución 41474010633) son el defecto que este hito mide: si el cierre los da, la
reparación no los arregla con texto nuevo en las skills del producto ni reorganizándolas, que queda fuera, y el informe
los da con sus cifras para que decida una persona (FR-053).

**Aceptación**: el sondeo de la persona, con `claude-sonnet-5-5`, 40 sesiones por variante y modo, deja su tabla en
`evidencias/` (FR-050), y un ADR que precisa el 0046 dice, con esas cifras, dónde va en una skill, y en el servidor, lo
que hace que el agente la active y cumpla lo que pide en cada modo. Si ninguna referencia se lee de forma fiable en
modo herramienta, el ADR lo dice y propone dónde va lo imprescindible (FR-051). El run no lanza el sondeo ni escribe el
ADR, que no tiene cifras hasta el sondeo: los dos se añaden a la propuesta de cambio antes de fusionarla, como el
sondeo de H32, y la propuesta no se fusiona sin ellos (FR-052).

## Assumptions

- Las 440 sesiones de partida y la ejecución 41474010633 son las que cita el hito; esta sesión no las ha medido.
- «Argumento de un applet» son los argumentos y banderas que declara un verbo; las banderas globales del kernel quedan
  como hoy (Fuera de alcance).
- «Antes de la primera herramienta» se mide por turnos del agente, cada uno una respuesta del modelo (FR-041): el mismo
  turno no es antes, porque las llamadas de un turno se emiten sin haber leído el resultado de las otras (la evidencia de partida: `ToolSearch` en el mismo turno
  que `Skill`).
- El modelo se pasa como identificador completo porque un alias lo resuelve el cliente (ADR 0031): solo así la tabla
  dice qué modelo midió.
- `schemas/` cambia (FR-004): como toda modificación de `schemas/`, la revisa una persona al leer el informe (capa 3).

## Fuera de alcance

- Reorganizar las skills del producto o cambiar su texto, también para arreglar los rojos de partida: es un hito aparte
  cuya entrada se escribe con el resultado del sondeo.
- Cambiar ningún umbral, límite ni tope: el de 11 000 caracteres y el de 900 (ADR 0046), los 512 bytes de las
  instrucciones del servidor, los umbrales del job de evals.
- Cambiar las instrucciones del servidor del producto o el texto de las ayudas de los argumentos.
- Dar descripción a las banderas globales del kernel en los esquemas.
- Guardar las sesiones del job de evals como artefacto (sigue en `backlog/medida-de-las-skills.md`).
- Que el job de evals mida la skill de laboratorio, o que `make ci` abra sesiones con modelo.
- Grabar lecturas nuevas o escribir en `evidencias/` desde el run.
- Lanzar el sondeo y escribir el ADR dentro del run (los hace la persona antes de fusionar).
- Una columna aparte para los `ToolSearch` que cargan una herramienta de kitlegal, o contar como activación la lectura
  del `SKILL.md` de la skill: la tabla mide lo que enumera FR-041.
- Evals de laboratorio sobre varias herramientas, con la regla de `parametro` repartida en varios parámetros.
- Variantes que combinen dos diferencias, otros modelos que `claude-sonnet-5-5` u otros hosts que Claude Code.
