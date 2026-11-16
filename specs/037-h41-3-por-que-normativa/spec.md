# Feature Specification: H41.3 · Por qué `normativa` no se activa con todas las skills instaladas, y si unas descripciones cortas lo arreglan

**Feature Branch**: `037-h41-3-por-que-normativa`

**Created**: 2026-11-16

**Status**: Draft

**Input**: el fichero del hito, `docs/roadmap/in-progress/H41.3-normativa-se-activa-con-todas-las-skills-instaladas.md`
(entrada de la persona, ADR 0018).

## Contexto

En el job de evals, la eval 04 de `normativa` (`04-lgt-prescripcion`) cae a veces en **modo herramienta**: la sesión
pide `boe_articulo` sin ningún `tool_use` de `Skill` y cuenta en `sin_activar`, cuyo umbral es cero. El ADR 0054
atribuye la caída, como hipótesis de trabajo, a la instalación completa (las cuatro skills del producto, como las deja
`kitlegal skills install`), y encarga a H41.3 confirmarla o descartarla, medir si la causa es el presupuesto del
listado de skills de Claude Code y medir un remedio que solo toca las descripciones de las skills del producto
(ADR 0054, punto 4). El remedio lo aplica un hito siguiente, con el ADR que lea el sondeo.

Evidencia de partida, tal como la dan el hito y el ADR 0054 (`evidencias/sondeos-20261112/`; no medida en esta sesión):

- en el sondeo de H41.2, con la eval 04 en modo herramienta, 0 de 80 sesiones con `normativa` sola dejan de activar la
  skill, y 6 de 80 [3,5–15,4 %] con las cuatro del producto (2 con Claude Code 2.1.284, 4 con 2.1.322; Fisher
  unilateral, p ≈ 0,014). Las seis cargan `boe_articulo` con `ToolSearch` y responden sin ningún `tool_use` de
  `Skill`, como las cinco caídas del job;
- con una sola skill instalada, otras 100 sesiones de los sondeos de H40 y H41 activaron antes: 0 de 180 con las de
  H41.2, una agregación hecha después de ver los datos que no sustituye a una medida hecha para esto;
- el listado de skills de Claude Code tiene un presupuesto del 1 % de la ventana del modelo, y al pasarlo quita
  descripciones, empezando por las de las skills que menos se usan; los nombres quedan siempre
  (https://code.claude.com/docs/en/skills, «Skill descriptions are cut short»). El presupuesto se fija en caracteres
  con `SLASH_COMMAND_TOOL_CHAR_BUDGET`, y la descripción de cada entrada se trunca a 1536 caracteres. Las sesiones de
  H41.2 listan 20 skills con 2.1.284 y 21 con 2.1.322, y las cuatro descripciones del producto suman 3313 caracteres.
  Si el listado pasa del presupuesto no se ha medido.

Comprobado en esta sesión, en el repositorio:

- el arnés del laboratorio (`make evals-laboratorio`) acepta hoy `INSTALACION=producto` y `EVAL=<nn>` (H41.2,
  `specs/036-h41-2-por-qu-e/contracts/arnes.md` §1 y §2), y su `sondeo.json` de una familia del producto anota
  `skills_instaladas`, los nombres de las skills de cada sesión (§3);
- la tabla de una familia del producto (`make laboratorio-tabla`) da una fila por variante, modo, modelo, versión de
  Claude Code y skills instaladas, con «Activan antes», «Activan otra skill», «Otras skills» y «No activan ninguna»,
  cada recuento con su intervalo de Wilson al 95 %; las columnas de las skills instaladas y de las activaciones solo
  salen si algún sondeo de la tabla anota `skills_instaladas` (`internal/evals/laboratorio_tabla.go`,
  `conLasActivaciones`), y la suite congelada de H41.2 (`specs/036-h41-2-por-qu-e/aceptacion/laboratorio-tabla.txtar`)
  compara esa tabla byte a byte;
- cada sesión del arnés recibe un entorno de lista cerrada —las variables de `variablesQueVenLasSesionesDelSondeo`
  (`internal/evals/sondeo.go:74`), más `PATH` y `HOME`—, así que hoy `SLASH_COMMAND_TOOL_CHAR_BUDGET` ya no le llega
  desde el entorno de quien lanza el arnés. Lo que el hito pide sin `PRESUPUESTO` es que siga sin llegarle, y su
  control lo comprueba con la variable puesta en el entorno de quien lanza;
- los límites de una descripción del producto se cuentan en runas, y «etiqueta XML» es la de la validación de la
  especificación (ADR 0046, puntos 2 y 3; `internal/skills/frontmatter.go`).

Para leer la tabla del sondeo (calculado en esta sesión con `awk` y la fórmula de Wilson, z = 1,96): 0 de 160 da
0,0–2,3 %; 7 de 160, 2,1–8,8 %; 8 de 160, 2,6–9,6 %; 12 de 160 (7,5 %), 4,3–12,7 %. Contra una fila con 0 de 160, otra
fila de 160 deja de tocar su intervalo desde 8 de 160.

## Clarifications

### Session 2026-11-16

- Q: ¿Con qué valores de VARIANTES acepta el arnés INSTALACION=producto-descripciones-cortas sin dar un error de argumentos (FR-010, FR-016)? → A: Solo con `VARIANTES=normativa-producto`; cualquier otra variante, sola o junto a ella (versionadas, `<skill>-producto` de otras familias, H39), es un error de argumentos de FR-016, con la regla declarada junto a la familia de `normativa` y el test de FR-018 cubriendo `normativa-control`, `jurisprudencia-producto` y una de H39 (auto: conservadora; criterio d; fuente: hito H41.3, «Entrega» y «El sondeo de H41.3»; constitución, «Criterio de decisión autónoma», puntos 2 y 4)
- Q: ¿El sondeo.json de una invocación sin INSTALACION=producto-descripciones-cortas ni PRESUPUESTO (filas sola y control) queda como hoy byte a byte, o anota también el nombre de su instalación y la ausencia de presupuesto (FR-012, FR-015, FR-022)? → A: Queda como hoy, byte a byte: solo las invocaciones con alguna de las dos opciones anotan la constancia de FR-012; la tabla deduce la instalación de `skills_instaladas` (FR-022) y el presupuesto vacío, y los tests de hoy no cambian (auto: criterio a; fuente: hito H41.3, «Entrega» y «Controles»; `skillsInstaladasDelSondeo`, `internal/evals/laboratorio_sondeo.go`)

## User Scenarios & Testing *(mandatory)*

### User Story 1 - Una variante del producto con las cuatro descripciones cortas (Priority: P1)

La persona y el sondeo disponen, en `laboratorio/` y fuera del producto, de la variante de instalación
`producto-descripciones-cortas`: solo cuatro `description`, una por skill de `skills/`, más cortas que las de hoy.
`make ci` dice en rojo si la variante deja de tener una por skill o pasa de sus límites.

**Why this priority**: es el remedio candidato que mide la cuarta fila del sondeo; sin él, el sondeo no puede decir
si unas descripciones cortas quitan la caída.

**Independent Test**: en `make ci`, sin red ni modelo, la comprobación de la variante sale en verde con la de la
rama y en rojo con una variante a la que le falta una skill, le sobra una, tiene una descripción de 451 runas, suma
1701 o lleva una etiqueta XML.

**Acceptance Scenarios**:

1. **Dado** la variante de la rama, **Cuando** corre `make ci`, **Entonces** la comprobación sale en verde: una
   `description` por cada skill de `skills/`, ninguna de más ni repetida, cada una de como mucho 450 runas, las cuatro
   de como mucho 1700 y ninguna con una etiqueta XML (FR-001, FR-002, SC-001).
2. **Dado** una variante a la que le falta la `description` de una skill de `skills/`, o que tiene una de una skill
   que no está en `skills/`, **Cuando** corre la comprobación, **Entonces** sale en rojo y nombra la skill (FR-002).
3. **Dado** una variante con una `description` de 451 runas, o cuatro que suman 1701, o una con `<ejemplo>`,
   **Cuando** corre la comprobación, **Entonces** sale en rojo con la medida o la etiqueta y el límite (FR-002, SC-001).
4. **Dado** que un hito posterior añade una quinta skill a `skills/` sin tocar la variante, **Cuando** corre
   `make ci`, **Entonces** la comprobación sale en rojo con el nombre de la skill que falta (FR-002).

---

### User Story 2 - El arnés instala la variante compuesta y fija el presupuesto del listado (Priority: P1)

La persona lanza `make evals-laboratorio` con `INSTALACION=producto-descripciones-cortas` y cada sesión tiene las
cuatro skills de `skills/` de la cabeza, cada una con la `description` de la variante y todo lo demás byte a byte; o
con `PRESUPUESTO=<caracteres>`, y cada sesión corre con `SLASH_COMMAND_TOOL_CHAR_BUDGET` a ese valor. Sin las dos
opciones, el arnés hace lo que hace hoy.

**Why this priority**: son las filas de diagnóstico y de remedio del sondeo.

**Independent Test**: en `make ci`, con el sustituto de `claude` sin modelo ni red que anota sus invocaciones, se
comprueban las skills, las descripciones y el entorno de cada sesión con cada opción y sin ellas.

**Acceptance Scenarios**:

1. **Dado** una petición de `normativa-producto` con `INSTALACION=producto-descripciones-cortas`, **Cuando** el arnés
   prepara cada sesión, **Entonces** la sesión tiene instaladas `consulta-juridica`, `jurisprudencia`, `normativa` y
   `plazos-y-territorio`, ninguna más ni repetida; la `description` del frontmatter de cada `SKILL.md`, leída, es
   exactamente la de la variante para esa skill, y cada otro byte de cada skill —el resto de su `SKILL.md`, sus
   referencias y cualquier otro fichero— es el de `skills/` del árbol que lanza el sondeo (FR-010, FR-011, SC-002).
2. **Dado** una petición válida con `PRESUPUESTO=1000000`, **Cuando** el arnés abre cada sesión, **Entonces** el
   entorno de cada sesión lleva `SLASH_COMMAND_TOOL_CHAR_BUDGET=1000000` (FR-013, SC-003).
3. **Dado** una petición válida sin `PRESUPUESTO` y con `SLASH_COMMAND_TOOL_CHAR_BUDGET=500` en el entorno de quien
   lanza el arnés, **Cuando** el arnés abre cada sesión, **Entonces** el entorno de ninguna sesión lleva esa variable
   (FR-014, SC-003).
4. **Dado** cualquier petición válida sin las dos opciones, **Cuando** el arnés prepara y abre sus sesiones,
   **Entonces** las skills instaladas, las sesiones que abre, sus invocaciones de `claude`, su entorno, su
   `sondeo.json` (sin ninguna constancia nueva, FR-012) y su salida son los de hoy, byte a byte, con los valores esperados de los
   tests de hoy sin cambiar (FR-015, SC-004).
5. **Dado** una petición con `PRESUPUESTO` sin la forma o el rango de FR-013, con `PRESUPUESTO` y variantes del
   laboratorio de H39, con `INSTALACION=producto-descripciones-cortas` y cualquier `VARIANTES` que no sea solo
   `normativa-producto`, o con un valor de `INSTALACION` que
   no vale, **Cuando** la persona lanza el arnés, **Entonces** sale con un código distinto de 0, dice qué no vale con
   el nombre del argumento de `make` y no abre ninguna sesión, no construye nada y no crea `DIRECTORIO` (FR-016,
   SC-005).

---

### User Story 3 - La tabla dice la instalación y el presupuesto de cada fila (Priority: P1)

Con las transcripciones de los cuatro sondeos de H41.3, `make laboratorio-tabla` da una fila por instalación y
presupuesto, con el nombre de la instalación y el presupuesto en una columna cada uno, y los recuentos que lee el ADR.

**Why this priority**: sin esas columnas, la fila de control y la de diagnóstico (mismas skills, mismas variantes) y
la de remedio (mismos nombres de skills) saldrían juntas.

**Independent Test**: en `make ci`, sin red ni modelo, el guion da las filas, columnas y recuentos exactos de unos
sondeos fijos de las cuatro filas, y la tabla de los sondeos fijos de las suites congeladas sale igual byte a byte.

**Acceptance Scenarios**:

1. **Dado** cuatro sondeos fijos de `normativa-producto` en modo herramienta, con el mismo modelo y la misma versión
   de Claude Code —sin `INSTALACION`, con `INSTALACION=producto`, con `INSTALACION=producto` y `PRESUPUESTO=1000000`, y
   con `INSTALACION=producto-descripciones-cortas`—, **Cuando** el guion los cuenta, **Entonces** da cuatro filas, con
   «Instalación» `normativa`, `producto`, `producto` y `producto-descripciones-cortas`, y «Presupuesto» vacío, vacío,
   `1000000` y vacío (FR-020, FR-021).
2. **Dado** esos sondeos con sesiones que activan `normativa` antes de la primera herramienta, solo otra skill y
   ninguna, **Cuando** el guion los cuenta, **Entonces** cada «Activan antes», «Activan otra skill» y «No activan
   ninguna» y cada intervalo de Wilson es el esperado, con las reglas de H41.2 (FR-020, FR-024, SC-006).
3. **Dado** los sondeos fijos de las suites congeladas de H40, H41 y H41.2, ninguno con la constancia de FR-012,
   **Cuando** el guion los cuenta, **Entonces** la tabla y el detalle salen como hoy, byte a byte (FR-022, SC-007).

---

### User Story 4 - La persona mide y lleva la tabla a un ADR (Priority: P2)

Con la propuesta de cambio abierta, la persona lanza el sondeo de H41.3 —cuatro filas de 160 sesiones de la eval 04,
modo herramienta, `claude-sonnet-5-5`, Claude Code 2.1.284—, deja su tabla en `evidencias/`, añade la tabla y su
lectura a la propuesta de cambio y lleva la tabla a un ADR que dice la instalación, el presupuesto y el remedio.

**Why this priority**: es la aceptación del hito, pero la hace una persona fuera del run (ADR 0032): el run entrega lo
que la hace posible, y la fusión la deciden los controles.

**Independent Test**: el informe final del run dice que la propuesta queda pendiente del sondeo y de su lectura; antes
de fusionar, `evidencias/` tiene la tabla y la propuesta su lectura.

**Acceptance Scenarios**:

1. **Dado** el arnés y el guion en la rama, **Cuando** la persona lanza las cuatro invocaciones de FR-030, **Entonces**
   la tabla, con el modelo, la versión de Claude Code, la instalación y el presupuesto de cada fila, queda en
   `evidencias/` con, por fila, las sesiones que activan `normativa` antes de la primera herramienta, las que activan
   otra skill y las que no activan ninguna, y sus intervalos de Wilson al 95 % (FR-030).
2. **Dado** la tabla, **Cuando** el intervalo de «No activan ninguna» del control queda por encima del de `sola` sin
   tocarse, **Entonces** el ADR dice que la caída es de la instalación completa; **Cuando** se tocan, el ADR dice que
   no se demuestra, que las otras dos comparaciones no deciden nada y qué se mide después (FR-031).
3. **Dado** que el control reproduce la caída, **Cuando** el intervalo del diagnóstico queda por debajo del del
   control sin tocarse, **Entonces** el ADR dice que la causa es el presupuesto del listado; **Cuando** se tocan, que
   no se demuestra que lo sea (FR-031).
4. **Dado** que el control reproduce la caída, **Cuando** el intervalo del remedio queda por debajo del del control
   sin tocarse y su «Activan otra skill» no queda por encima de la del control, **Entonces** el ADR dice que el remedio
   lo aplica el hito siguiente y lo valida con las evals de las cuatro skills en el job; **Cuando** no, el ADR dice
   qué se mide después (FR-031).

---

### Edge Cases

- **`INSTALACION=producto-descripciones-cortas` con `PRESUPUESTO`**: vale; las sesiones tienen la variante compuesta y
  la variable a ese valor (FR-010, FR-013).
- **`PRESUPUESTO` sin `INSTALACION`** (la skill de la familia sola): vale; las sesiones tienen la instalación de hoy y
  la variable (FR-013).
- **Una `description` de la variante con caracteres que el YAML del frontmatter tiene que citar** (dos puntos,
  comillas, `#`): la `description` leída del `SKILL.md` compuesto es la de la variante, exacta, y la skill compuesta
  pasa la validación de la especificación como las de `skills/` (FR-011).
- **Una skill de `skills/` sin `description` en la variante, o una `description` de una skill que no está**: `make ci`
  sale en rojo (FR-002), y el arnés con `INSTALACION=producto-descripciones-cortas` no abre ninguna sesión (FR-016).
- **Una tabla que junta sondeos de H41.2 (con `skills_instaladas` y sin la constancia de FR-012) y de H41.3**: las
  filas de H41.2 dan como instalación `normativa` (la skill de la familia sola) o `producto` (las cuatro), según sus
  `skills_instaladas`, y el presupuesto vacío (FR-022).
- **Fila con `n` = 0** (todas sus sesiones sin terminar): recuentos e intervalos vacíos, como hoy; su instalación y su
  presupuesto salen igual (FR-020).
- **Transcripción, `sondeo.json` o variante que el guion o el arnés no sabe leer** (manipulado a mano, truncado por
  algo que no es el arnés): defecto `inesperado`, código 1 (ADR 0023), sin más casos.

## Requirements *(mandatory)*

### Functional Requirements

**La variante `producto-descripciones-cortas`**

- **FR-001**: `laboratorio/` guarda la variante de instalación `producto-descripciones-cortas`, fuera del producto: solo
  cuatro textos de `description`, uno por skill de `skills/` (`consulta-juridica`, `jurisprudencia`, `normativa` y
  `plazos-y-territorio`), y nada más de ninguna skill (ni cuerpos, ni referencias, ni otras claves del frontmatter).
  Cada texto:
  - tiene como mucho 450 runas, y los cuatro juntos, como mucho 1700;
  - va en tercera persona y no tiene ninguna etiqueta XML (ADR 0046);
  - dice qué hace la skill y cuándo se usa, con los disparadores de la `description` de hoy que quepan, empezando por
    los que nombran su capacidad.

  Los cuatro textos los fija el plan del hito, que dice por cada uno qué disparadores de la `description` de hoy
  conserva y cuáles deja. La tercera persona y lo que dice cada texto no los juzga nadie en este hito: la tercera
  persona la juzgará el job de evals cuando el hito siguiente aplique los textos al producto (ADR 0046, punto 4).
- **FR-002**: `make ci` comprueba la variante sin red ni modelo: la comprobación sale en rojo si a una skill de
  `skills/` le falta su `description` en la variante, si la variante tiene una de una skill que no está en `skills/`
  o repetida, si una tiene más de 450 runas, si las cuatro suman más de 1700 o si una tiene una etiqueta XML (la
  definición de ADR 0046, punto 3), y en cada caso dice la skill, la medida o la etiqueta y el límite.

**El arnés del laboratorio**

- **FR-010**: `make evals-laboratorio` acepta `INSTALACION=producto-descripciones-cortas`. Con ella, cada sesión tiene
  instaladas las cuatro skills de la variante compuesta (FR-011), también la de la familia, que ocupa el lugar de la
  de la variante pedida (`VARIANTES=normativa-producto`), y ningún nombre repetido. El resto —las sesiones que abre,
  sus invocaciones, la comprobación de la versión de Claude Code y del modelo, y la constancia de `skills_instaladas`
  en `sondeo.json`— es como con `INSTALACION=producto`.
  `INSTALACION=producto-descripciones-cortas` solo vale con `VARIANTES=normativa-producto`, la variante que nombra el
  hito. Cualquier otra variante en `VARIANTES`, sola o junto a `normativa-producto`, es un error de argumentos de
  FR-016: las versionadas de `normativa` (`normativa-control`…), las `<skill>-producto` y las versionadas de las otras
  dos familias del producto, y las de H39. El mensaje nombra `INSTALACION` y dice que `producto-descripciones-cortas`
  solo vale con `normativa-producto`; con las variantes de H39 conserva el de FR-016, que dice que no son de ninguna
  familia del producto.
- **FR-011**: El arnés compone la variante al preparar el sondeo: copia `skills/` del árbol que lanza el sondeo y pone
  en el frontmatter de cada `SKILL.md` la `description` de la variante para esa skill, y nada más. La `description`
  leída de cada `SKILL.md` compuesto es exactamente el texto de la variante; cada otro byte de cada skill es el de
  `skills/` de ese árbol; y cada `SKILL.md` compuesto pasa la validación del frontmatter que pasan las de `skills/`
  (ADR 0046, punto 3). La copia compuesta no se guarda en `DIRECTORIO`, como hoy no se guarda la de `skills/`.
- **FR-012**: El `sondeo.json` de una invocación con `INSTALACION=producto-descripciones-cortas` anota el nombre de la
  instalación, `producto-descripciones-cortas`, y el de una invocación con `PRESUPUESTO`, su valor; los dos, donde los
  lee el guion de la tabla.
  Una invocación sin las dos opciones (las filas `sola`, con `INSTALACION` vacía, y `control`, con
  `INSTALACION=producto`) deja su `sondeo.json` como hoy, byte a byte: no anota ni el nombre de su instalación ni la
  ausencia de presupuesto. La tabla deduce su instalación de `skills_instaladas` (FR-022), y no distingue un sondeo de
  H41.3 sin opciones de uno de H41.2: miden lo mismo, la misma instalación y ningún presupuesto en la sesión.
- **FR-013**: `make evals-laboratorio` acepta `PRESUPUESTO=<caracteres>`, un entero mayor que 0 escrito solo con cifras,
  sin signo ni ceros delante, que cabe en un `int`: la forma y el rango de `SESIONES`. Con él,
  cada sesión que abre el arnés corre con `SLASH_COMMAND_TOOL_CHAR_BUDGET=<caracteres>` en su entorno, con el valor
  tal como se pidió, y nada más de su entorno cambia. Vale con cualquier `INSTALACION` válida de una familia del
  producto, también vacía.
- **FR-014**: Sin `PRESUPUESTO`, o vacío, ninguna sesión del arnés recibe `SLASH_COMMAND_TOOL_CHAR_BUDGET`, aunque
  esté en el entorno de quien lanza el arnés.
- **FR-015**: Una petición válida sin las dos opciones —`INSTALACION` vacía o `producto`, y `PRESUPUESTO` vacío— hace lo
  que hace hoy, byte a byte: las skills que instala en cada sesión, las sesiones que abre, sus invocaciones de
  `claude`, su entorno, su `sondeo.json` (sin ninguna constancia nueva, FR-012) y su salida. La única diferencia posible con hoy
  es la de FR-014, que con el entorno de las sesiones de hoy no cambia ningún byte. Una petición sin las dos opciones
  que no vale da los mensajes de error de hoy, byte a byte.
- **FR-016**: Son errores de argumentos como los de hoy —junto a los demás, uno por línea, cada uno con el nombre del
  argumento de `make`; salida con un código distinto de 0; ninguna sesión abierta, nada construido y `DIRECTORIO` sin
  crear—: `PRESUPUESTO` que no tiene la forma o el rango de FR-013 (`0`, `0100`, `+1000000` o `99999999999999999999`, que no cabe en un `int`, no valen);
  `PRESUPUESTO` o `INSTALACION=producto-descripciones-cortas` con variantes del laboratorio de H39, que no son de
  ninguna familia del producto; `INSTALACION` con un valor que no es vacío, `producto` ni
  `producto-descripciones-cortas`, con el mensaje de hoy, byte a byte;
  `INSTALACION=producto-descripciones-cortas` con cualquier `VARIANTES` que no sea solo `normativa-producto` (FR-010); y
  `INSTALACION=producto-descripciones-cortas` cuando la variante de instalación no tiene una `description` por
  skill de `skills/` o que pasa de sus límites (FR-002). Los mensajes de hoy de los demás argumentos no cambian.
- **FR-017**: Lo demás del arnés no cambia: comprueba antes de abrir ninguna sesión que la versión de Claude Code
  instalada es la de `CLAUDE_CODE` y que el modelo es un id completo, y los anota; lo lanza una persona con la
  credencial del sondeo (ADR 0032), nunca `make ci`, un gancho, el job de evals ni un paso del workflow. El run no lo
  lanza, no graba nada en `testdata/` y no escribe en `evidencias/`.
- **FR-018**: `make ci` comprueba el arnés sin red ni modelo, con el sustituto de `claude` que anota sus invocaciones
  y su entorno:
  - con `INSTALACION=producto-descripciones-cortas`, qué skills deja instaladas en cada sesión —las cuatro de
    `skills/`, ninguna repetida— y con qué `description` —la de la variante, exacta—, y que lo que no es la
    `description` es `skills/` de la cabeza, byte a byte, fichero a fichero;
  - con `PRESUPUESTO`, que `SLASH_COMMAND_TOOL_CHAR_BUDGET` llega al entorno de cada sesión con el valor pedido; y sin
    él, con la variable puesta en el entorno de quien lanza el arnés, que no llega a ninguna;
  - sin las dos opciones, lo que hace hoy, byte a byte, también los mensajes de una petición que no vale, con los
    valores esperados de los tests de hoy sin cambiar;
  - con cada petición inválida de FR-016, 0 sesiones abiertas; para `INSTALACION=producto-descripciones-cortas`, al
    menos con una variante versionada (`normativa-control`), con la `-producto` de otra familia
    (`jurisprudencia-producto`) y con una de H39.

  El test falla con la primera sesión cuyas skills, descripción, bytes o entorno difieren de los esperados, o con una
  petición inválida que abre alguna sesión.

**La tabla del sondeo**

- **FR-020**: En la tabla de una familia del producto con algún sondeo que lleva la constancia de FR-012, cada fila
  da, además de lo de hoy, el nombre de su instalación y su presupuesto, en una columna cada uno, y hay una fila por
  variante, modo, modelo, versión de Claude Code, skills instaladas, instalación y presupuesto. La instalación es
  `producto-descripciones-cortas` en los sondeos con esa opción, `producto` en los de las cuatro skills sin ella, y el
  nombre de la skill de la familia (`normativa`) en los de la skill sola. El presupuesto es el valor pedido, o la celda
  vacía sin él. Los recuentos —«Activan antes», «Activan otra skill», «Otras skills», «No activan ninguna»—, sus
  intervalos de Wilson al 95 % con un decimal, `n` y el detalle son los de H41.2, sin cambiar cómo se cuentan.
- **FR-021**: En el sondeo de H41.3, la fila `sola` da `normativa` y presupuesto vacío; `control`, `producto` y vacío;
  `diagnóstico`, `producto` y `1000000`; `remedio`, `producto-descripciones-cortas` y vacío.
- **FR-022**: Una tabla sin ningún sondeo con la constancia de FR-012 —todos los anteriores a este hito— sale igual
  que hoy, byte a byte, tabla y detalle. En una tabla que junta sondeos con la constancia y sin ella, las filas de los
  que no la tienen dan como instalación `producto` si sus `skills_instaladas` son las cuatro del producto, o la skill
  de la familia si son solo ella o no las anotan, y el presupuesto vacío.
- **FR-023**: La tabla de la familia del laboratorio de H39 no cambia.
- **FR-024**: `make ci` comprueba el guion con sondeos fijos de `normativa-producto` en modo herramienta, con el mismo
  modelo y la misma versión, uno por cada fila de FR-021, con sesiones que activan `normativa` antes de la primera
  herramienta, solo otra skill y ninguna, y con uno de H41.2 junto a ellos: el test falla si una fila, una instalación,
  un presupuesto, un recuento o un intervalo difiere del esperado.

**La aceptación de la persona y lo que no cambia**

- **FR-030**: Con la propuesta de cambio abierta, la persona lanza el sondeo de H41.3 con el arnés de este hito:
  `VARIANTES=normativa-producto`, solo la eval 04 (`EVAL=04`), solo el modo herramienta (`MODOS=herramienta`),
  `MODELO=claude-sonnet-5-5` y `CLAUDE_CODE=2.1.284`, la versión del job (ADR 0031), con `SESIONES=160` en cuatro
  invocaciones, una fila de la tabla cada una, 640 sesiones en total:
  - **sola**: sin `INSTALACION`;
  - **control**: `INSTALACION=producto`;
  - **diagnóstico**: `INSTALACION=producto` y `PRESUPUESTO=1000000`, que no puede recortar el listado aunque el
    presupuesto cuente también las órdenes: las sesiones de H41.2 declaran como mucho 21 skills y 49 órdenes en su
    mensaje `system/init`, y 70 entradas con la descripción en su tope de 1536 caracteres ocupan 107 520;
  - **remedio**: `INSTALACION=producto-descripciones-cortas`.

  Su tabla queda en `evidencias/` con, por fila, las sesiones que activan `normativa` antes de la primera herramienta,
  las que activan otra skill y las que no activan ninguna, y sus intervalos de Wilson al 95 %.
- **FR-031**: La tabla y su lectura se añaden a la propuesta de cambio antes de fusionarla, y la propuesta no se
  fusiona sin ellas; no deciden la fusión, que la deciden los controles. Cada comparación es entre los intervalos de
  dos filas en «No activan ninguna», salvo donde se dice otra cosa; dos intervalos «no se tocan» si el extremo superior
  del de abajo es menor que el inferior del de arriba, con los extremos que imprime la tabla, y si son iguales se tocan.
  La tabla va a un ADR que dice:
  - **la instalación**: si el control queda por encima de `sola`, con intervalos que no se tocan, la caída es de la
    instalación completa; si se tocan, no se demuestra. Si el control no reproduce la caída —sus intervalos se tocan
    con los de `sola`, en torno a 1 de cada 12 sondeos a una tasa del 7,5 %—, las otras dos comparaciones no deciden
    nada, y el ADR dice qué se mide después;
  - **el presupuesto**: si el diagnóstico queda por debajo del control, con intervalos que no se tocan, la causa es el
    presupuesto del listado; si se tocan, no se demuestra que lo sea;
  - **el remedio**: si queda por debajo del control, con intervalos que no se tocan, y su «Activan otra skill» no queda
    por encima de la del control (sus intervalos se tocan o el del remedio queda por debajo), lo aplica el hito
    siguiente, que lo valida con las evals de las cuatro skills en el job, que las mide todas cuando cambia una
    descripción (ADR 0045). Si no, el ADR dice qué se mide después.

  El run no lanza el sondeo ni escribe el ADR, y su informe final dice que la propuesta queda pendiente de la tabla y
  su lectura.
- **FR-032**: Las suites de aceptación congeladas de H39 a H41.2 (`specs/031-h39-c-omo-se/aceptacion/`,
  `specs/032-h40-por-qu-e/aceptacion/`, `specs/034-h41-las-skills-del/aceptacion/`,
  `specs/035-h41-1-plazos-y-territorio/aceptacion/` y `specs/036-h41-2-por-qu-e/aceptacion/`) siguen en verde en
  `make ci` sin cambiar ningún fichero suyo.
- **FR-033**: Ninguna skill del producto cambia: ni `skills/`, ni sus cuerpos, ni sus referencias, ni sus
  descripciones, ni lo que genera `data/`; la variante vive solo en `laboratorio/` y solo la usa el arnés.
- **FR-034**: Las evals de las skills del producto salen en verde en la medición del cierre, en los dos modos. Si caen
  solo por la caída que este hito mide (`sin_activar` en la eval 04 de `normativa` en modo herramienta), la reparación
  no la arregla ni la esquiva —no toca una skill, una eval, un umbral, el job ni sus versiones—: el informe final da el
  umbral sin cumplir con sus cifras, como hoy, y la deja dicha para la persona (ADR 0054, «Consecuencias»).

### Key Entities

- **Variante de instalación `producto-descripciones-cortas`**: cuatro textos de `description`, uno por skill de
  `skills/`, en `laboratorio/`.
- **Instalación compuesta**: `skills/` de la cabeza con la `description` de cada `SKILL.md` sustituida por la de la
  variante.
- **Instalación de una fila**: `normativa` (la skill de la familia sola), `producto` (las cuatro) o
  `producto-descripciones-cortas` (las cuatro compuestas).
- **Presupuesto de una fila**: el valor de `SLASH_COMMAND_TOOL_CHAR_BUDGET` de sus sesiones, o ninguno.
- **Sondeo de H41.3**: las 640 sesiones de FR-030 y su tabla en `evidencias/`.

### Uso, de fuera adentro

Ninguna skill del producto ni el binario que se distribuye consume nada de lo que entrega el hito, que vive en
`laboratorio/` y en el arnés (FR-033). Lo que gana la skill es indirecto: saber si la caída de `normativa` es de la
instalación, si la causa es el presupuesto del listado y si unas descripciones cortas la quitan, antes de que el hito
siguiente cambie las descripciones del producto.

- **Las cuatro descripciones de la variante** (FR-001, FR-011). Las lee el modelo de cada sesión de la fila de remedio
  una vez, en el listado de skills del arranque, al decidir qué skill activar —es lo que se mide—: 160 sesiones en el
  sondeo. Tamaño fijo, como mucho 1700 runas las cuatro frente a las 3313 de hoy. No crece con el uso del producto
  —cientos de normas y miles de bloques consultados no la cambian—. Deja de usarse cuando el ADR que lea el sondeo
  decide; si el hito siguiente aplica los textos al producto, la variante coincide con `skills/` y su comprobación
  (FR-002) sigue protegiendo que tenga una por skill.
- **La comprobación de la variante en `make ci`** (FR-002). La lee la persona o el agente que hace `make ci`, solo
  cuando sale en rojo: una línea por defecto, como mucho una por skill y una por la suma. Se da mientras la variante no
  cumple, y deja de darse al corregirla; un hito que añade una skill a `skills/` la vuelve a dar hasta que la variante
  la tenga o se retire.
- **`sondeo.json` y la salida del arnés** (FR-012, FR-016). El guion de la tabla lee `sondeo.json` una vez por sondeo;
  la constancia nueva añade a lo de hoy el nombre de la instalación y el presupuesto, unas decenas de bytes. La salida
  de error la lee la persona cuando la orden no vale, una línea por argumento, y deja de darse cuando la orden vale.
  No crecen con el uso del producto, solo con las opciones pedidas.
- **Las columnas nuevas de la tabla** (FR-020). Las lee la persona una vez, para la lectura y el ADR: cuatro filas en
  el sondeo de H41.3, cada una con su instalación (un nombre) y su presupuesto (un número o nada), más el detalle de
  640 sesiones como hoy. Su tamaño lo fijan las filas y las sesiones pedidas, no el uso del producto. Deja de leerse
  cuando el ADR está escrito; un sondeo nuevo da otra tabla.

## Success Criteria *(mandatory)*

### Measurable Outcomes

- **SC-001**: La variante tiene 1 `description` por cada skill de `skills/` y 0 de más, cada una de ≤ 450 runas, las
  cuatro de ≤ 1700 en total y 0 etiquetas XML. Control: la comprobación de FR-002 en `make ci` sale en rojo si a una
  skill le falta, si sobra una, si una pasa de 450, si la suma pasa de 1700 o si hay una etiqueta.
- **SC-002**: Con `INSTALACION=producto-descripciones-cortas`, 4 skills por sesión, 0 repetidas, el 100 % de las
  `description` iguales a las de la variante y 0 bytes de diferencia con `skills/` de la cabeza en todo lo demás.
  Control: el test de FR-018 en `make ci` falla con la primera sesión que difiere.
- **SC-003**: Con `PRESUPUESTO=N`, el 100 % de las sesiones con `SLASH_COMMAND_TOOL_CHAR_BUDGET=N`; sin él, 0 sesiones
  con la variable aunque esté en el entorno de quien lanza el arnés. Control: el test de FR-018 falla con la primera
  sesión que difiere.
- **SC-004**: Sin las dos opciones, 0 bytes de diferencia con lo que hace el arnés hoy en lo que enumera FR-015,
  también en los mensajes de una petición que no vale.
  Control: el test de FR-018, con los valores esperados de hoy sin cambiar, falla con la primera diferencia.
- **SC-005**: 0 sesiones abiertas con cada petición inválida de FR-016. Control: el test de FR-018 falla si una abre
  alguna.
- **SC-006**: El guion da las instalaciones, los presupuestos, los recuentos y los extremos de Wilson con un decimal
  exactos de los sondeos fijos de FR-024. Control: su test en `make ci` falla con el primer valor que difiere.
- **SC-007**: 0 bytes de diferencia en la tabla y el detalle de los sondeos fijos de las suites congeladas, y las
  suites congeladas de FR-032 en verde. Control: esas suites en `make ci` fallan con la primera diferencia.
- **SC-008**: Las evals de las skills del producto en verde en la medición del cierre, en los dos modos, con
  `sin_activar` ≤ 0 como hoy. Control: el job de evals del cierre sale en rojo si una no pasa su umbral, y el informe
  final da sus cifras (FR-034).
- **SC-009**: Antes de fusionar, la tabla de las cuatro filas de 160 sesiones del sondeo de H41.3 está en
  `evidencias/` y la propuesta de cambio lleva su lectura. Lo comprueba la persona (capa 3): ningún control automático
  lo mide, y el run no lo hace (FR-031).

## Criterios del hito (literales)

**Controles**:

- `make ci` comprueba, sin red ni modelo, que la variante guarda una `description` por skill de `skills/` y que
  cumplen los límites de caracteres y de etiquetas (FR-002, SC-001);
- el arnés con las opciones nuevas, sin red ni modelo, comprueba (FR-018):
  - qué skills deja instaladas en cada sesión y con qué descripción, y que lo que no es la `description` es `skills/`
    de la cabeza, byte a byte (SC-002);
  - que `SLASH_COMMAND_TOOL_CHAR_BUDGET` llega al entorno de la sesión con el valor pedido, y que sin la opción no
    llega aunque esté en el entorno de quien lanza el arnés (SC-003);
  - que sin las dos opciones hace lo que hace hoy, byte a byte, salvo eso (SC-004);
- las suites congeladas de H39 a H41.2 siguen en verde (FR-032, SC-007).

Las evals de las skills del producto salen en verde en la medición del cierre, en los dos modos. Si caen solo por la
caída de este hito, la reparación no la arregla ni la esquiva: la deja dicha para la persona (ADR 0054,
«Consecuencias») (FR-034, SC-008).

**Aceptación**: el sondeo de la persona deja su tabla en `evidencias/`, con las sesiones de cada fila que activan
`normativa` antes de la primera herramienta, las que activan otra skill y las que no activan ninguna, y sus intervalos
de Wilson al 95 % (FR-030). No decide la fusión, que valen los controles: la tabla y su lectura se añaden a la
propuesta de cambio antes de fusionarla, y la propuesta no se fusiona sin ellas. Cada comparación es entre los
intervalos de dos filas en «No activan ninguna», salvo donde se dice otra cosa. La tabla va a un ADR que dice
(FR-031, SC-009):

- **la instalación**: si el control queda por encima de `sola`, con intervalos que no se tocan, la caída es de la
  instalación completa; si se tocan, no se demuestra. Si el control no reproduce la caída (en torno a 1 de cada 12
  sondeos a una tasa del 7,5 %), las otras dos comparaciones no deciden nada, y el ADR dice qué se mide después;
- **el presupuesto**: si el diagnóstico queda por debajo del control, con intervalos que no se tocan, la causa es el
  presupuesto del listado; si se tocan, no se demuestra que lo sea;
- **el remedio**: si queda por debajo del control, con intervalos que no se tocan, y su «Activan otra skill» no queda
  por encima de la del control, lo aplica el hito siguiente. Ese hito lo valida con las evals de las cuatro skills en
  el job, que las mide todas cuando cambia una descripción (ADR 0045). Si no, el ADR dice qué se mide después.

El run no lanza el sondeo (FR-017, FR-031).

## Assumptions

- Las cifras de partida (las 6 de 80, las 0 de 180, los 3313 caracteres, las 20 y 21 skills del listado, las 49
  órdenes) son las que dan el hito y el ADR 0054; esta sesión no las ha medido.
- «Caracteres» en los límites de la variante son runas, como en los límites de la descripción del producto (ADR 0046,
  punto 2), y «etiqueta XML» es la de su validación (punto 3).
- «Sin las dos opciones» es una petición válida con `INSTALACION` vacía o `producto` y `PRESUPUESTO` vacío. Una
  petición que no vale y no lleva ninguna de las dos da los mensajes de error de hoy, byte a byte: el de un
  `INSTALACION` que no vale sigue siendo el de hoy («no es producto»), que sigue siendo cierto (FR-015, FR-016).
- `INSTALACION=producto-descripciones-cortas` solo con `normativa-producto` es la lectura conservadora: el hito no pide
  otra variante y ningún sondeo la usa; con una versionada, la variante compuesta sustituiría por entero su contenido y
  una fila etiquetada `normativa-control` mediría la cabeza compuesta, un dato mal etiquetado (FR-010).
- `PRESUPUESTO` con variantes del laboratorio de H39 es un error de argumentos: la tabla de H39 no cambia (FR-023) y no
  podría decir el presupuesto de su fila; es la lectura que menos comportamiento añade, la misma que H41.2 dio a
  `INSTALACION=producto` con esas variantes.
- Que hoy la variable no llegue a las sesiones (su entorno es una lista cerrada) no hace innecesario FR-014: el hito
  pide que no llegue, y su control lo fija para que un cambio del entorno de las sesiones no la deje pasar.
- La variante compuesta se rehace en cada invocación desde `skills/` del árbol que lanza el sondeo, así que no se queda
  atrás cuando otro hito cambia un cuerpo o una referencia; lo único que puede quedarse atrás son las cuatro
  descripciones, y eso lo dice la comprobación de FR-002.
- El modelo se pasa como id completo porque un alias lo resuelve el cliente (ADR 0031).

## Fuera de alcance

- Cambiar ninguna skill del producto, sus cuerpos, sus referencias o sus descripciones: el remedio lo aplica el hito
  siguiente (FR-033).
- Cambiar el umbral de `sin_activar`, cómo se cuenta, o ningún otro umbral, límite ni tope (ADR 0038, punto 1).
- Cambiar el job de evals, lo que instala, `VERSION_DE_CLAUDE_CODE` o `MODELO_DE_EVALS` (ADR 0031).
- Pedir la activación en las instrucciones del servidor o en `consulta-juridica` (ADR 0050, punto 3).
- Quitar del producto `consulta-juridica` u otra skill (ADR 0048).
- Fijar `SLASH_COMMAND_TOOL_CHAR_BUDGET`, o cualquier ajuste de Claude Code, en el equipo de quien instala kitlegal o
  en `kitlegal skills install`.
- Evals nuevas, y cambiar las que ya existen.
- Medir con el juez en el sondeo, que cuenta activación, sin modelo; y juzgar en este hito la tercera persona o el
  contenido de las cuatro descripciones de la variante.
- `INSTALACION=producto-descripciones-cortas` con otra variante que `normativa-producto`: las versionadas
  (`normativa-control`…), las `<skill>-producto` de las otras dos familias y las de H39 (FR-010).
- Anotar en `sondeo.json` el nombre de la instalación y la ausencia de presupuesto en las invocaciones sin las dos
  opciones, y distinguir en la tabla un sondeo de H41.3 sin opciones de uno de H41.2 (FR-012).
- Otras variantes de instalación que `producto-descripciones-cortas`; variantes que cambien algo más que la
  `description` (el `name`, otras claves del frontmatter, cuerpos o referencias); y `INSTALACION` o `PRESUPUESTO` con
  las variantes del laboratorio de H39.
- Medir el listado de skills de las sesiones (cuántos caracteres ocupa, qué descripciones recorta Claude Code) o
  anotarlo en la tabla.
- Columnas nuevas en la tabla de la familia del laboratorio de H39, o cambiar cómo se cuentan las activaciones.
- Otras versiones de Claude Code o modelos en el sondeo de H41.3.
- Que el job de evals mida las variantes o lance el arnés, o que `make ci` abra sesiones con modelo.
- Lanzar el sondeo y escribir el ADR dentro del run, grabar lecturas nuevas o escribir en `evidencias/` desde el run.
