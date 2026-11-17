# Feature Specification: H41.4 · Las skills del producto, con las descripciones cortas que quitan la caída de la eval 04

**Feature Branch**: `038-h41-4-las-skills-del`

**Created**: 2026-11-17

**Status**: Draft

**Input**: el fichero del hito, `docs/roadmap/in-progress/H41.4-las-skills-del-producto-con-descripciones-cortas.md`
(entrada de la persona, ADR 0018).

## Contexto

En el job de evals, la eval 04 de `normativa` (`04-lgt-prescripcion`) cae a veces en **modo herramienta**: la sesión
pide `boe_articulo` y responde sin ningún `tool_use` de `Skill`, y cuenta en `sin_activar`, cuyo umbral es cero. El
ADR 0055 concluye que la caída es de la instalación completa (las cuatro skills del producto), no demuestra que la
cause el presupuesto del listado de skills, y encarga a H41.4 aplicar en `skills/` las cuatro `description` de la
variante `laboratorio/producto-descripciones-cortas/` y nada más, con un criterio de aceptación fijado antes de medir
(punto 4).

Evidencia de partida, tal como la dan el hito y el ADR 0055 (`evidencias/sondeos-20261117/`; no medida en esta
sesión): 640 sesiones de la eval 04 en modo herramienta, `claude-sonnet-5-5` y Claude Code 2.1.326. En «No activan
ninguna»: `sola` 1 de 160 [0,1–3,5 %], `control` 14 de 160 [5,3–14,2 %], `diagnostico` 7 de 160 [2,1–8,8 %] y
`remedio` (la variante `producto-descripciones-cortas`) 0 de 160 [0,0–2,3 %]. En `remedio`, 18 sesiones activan
también `consulta-juridica`, frente a 0 en `control`, y en las 18 `normativa` se activa antes de la primera
herramienta. Las que no activan `normativa` antes de la primera herramienta son 14 de 160 en `control` y 0 de 160 en
`remedio`.

Comprobado en esta sesión, en el repositorio (base de la rama `ab951c227e1fc8acac8bd6b23534e275e4a36a76`):

- las cuatro skills de `skills/` (`consulta-juridica`, `jurisprudencia`, `normativa`, `plazos-y-territorio`) llevan
  hoy su `description` en un bloque `>-` partido en líneas, justo detrás de `name:`;
- la variante tiene un fichero `<skill>.txt` por skill, de una línea y su salto final; sin el salto, sus textos miden
  417 (`consulta-juridica`), 420 (`jurisprudencia`), 426 (`normativa`) y 427 (`plazos-y-territorio`) caracteres,
  1690 entre los cuatro (`wc -m` en UTF-8, restando el salto de línea de cada fichero);
- la suite congelada de H41.3 fija esos cuatro textos (`internal/evals/laboratorio_descripciones_test.go`,
  `TestDescripcionesCortas`), y la composición del arnés con `INSTALACION=producto-descripciones-cortas` escribe la
  `description` como escalar de una línea, no como bloque `>-` (`skillConLaDescripcion`,
  `internal/evals/laboratorio_descripciones.go`);
- `TestSkillsDelProductoDesdeControl` (`internal/evals/laboratorio_familias_test.go`) compara hoy `skills/<skill>` de
  `normativa`, `jurisprudencia` y `plazos-y-territorio` con su base y su declaración, ruta a ruta y byte a byte, el
  `SKILL.md` incluido («desde control»), y su frontmatter con el de su control («frontmatter»);
- fuera de `skills/`, una frase de la descripción de hoy de `normativa` («dónde se regula una») solo aparece en las
  variantes de `laboratorio/normativa/`, en `specs/` y en una entrada pasada de `CHANGELOG.md` (`grep -rl`, sin
  `specs/` ni `evidencias/` en el recuento de lo vivo); el hito afirma que nada generado del repositorio copia las
  descripciones, y FR-007 lo hace comprobable.

## Clarifications

Ninguna en esta sesión: el fichero del hito y el ADR 0055 fijan el texto, la forma, los controles y la aceptación.

## User Scenarios & Testing *(mandatory)*

### User Story 1 - Las cuatro skills del producto llevan las descripciones medidas (Priority: P1)

Quien instala kitlegal (persona o despacho, con `kitlegal skills install`, el plugin o el `.mcpb`) recibe las cuatro
skills con las `description` cortas que en el sondeo de H41.3 quitaron la caída de la eval 04, y nada más cambiado:
el agente que lee el listado de skills al empezar cada sesión ve esos cuatro textos y activa `normativa` antes de la
primera herramienta `boe_…`.

**Why this priority**: es la entrega del hito; sin ella la caída sigue y el cierre de cualquier hito puede salir en
rojo por ella (ADR 0055, «Consecuencias»).

**Independent Test**: en `make ci`, sin red ni modelo, la comprobación de las descripciones sale en verde con la rama
y en rojo con un carácter cambiado en una de las cuatro `description` de `skills/`.

**Acceptance Scenarios**:

1. **Dado** la cabeza de la rama, **Cuando** corre `make ci`, **Entonces** la `description` que lee
   `skills.LeerFrontmatter` de cada `skills/<skill>/SKILL.md` es, byte a byte, el texto de
   `laboratorio/producto-descripciones-cortas/<skill>.txt` sin su salto de línea final (FR-001, SC-001).
2. **Dado** una copia de `skills/` con un carácter añadido, quitado o cambiado en la `description` de una skill,
   **Cuando** corre la comprobación de FR-001, **Entonces** sale en rojo y nombra la skill (FR-001).
3. **Dado** la cabeza de la rama, **Cuando** se lee el frontmatter de cada `skills/<skill>/SKILL.md`, **Entonces** la
   clave es `description: >-` y su texto ocupa más de una línea sangrada (FR-002).
4. **Dado** la cabeza de la rama, **Cuando** corre `make ci`, **Entonces** cada `description` de `skills/` tiene como
   mucho 450 caracteres, las cuatro suman como mucho 1700, ninguna pasa de 900 y ninguna lleva una etiqueta XML
   (FR-004, SC-003).

---

### User Story 2 - Nada más de las skills cambia (Priority: P1)

Quien usa kitlegal recibe los mismos protocolos, referencias y nombres que en `main`: el sondeo de aceptación compara
solo las descripciones, y la persona que lo lee puede atribuir a ellas cualquier diferencia.

**Why this priority**: el ADR 0055 (opción 3) rechaza cualquier remedio sin medir; un cambio más en las skills haría
que el sondeo midiese otra cosa.

**Independent Test**: en `make ci`, la comprobación de FR-003 sale en verde con la rama y en rojo con un carácter
cambiado en el cuerpo de un `SKILL.md`, en una referencia o en `name`, en cualquiera de las cuatro skills.

**Acceptance Scenarios**:

1. **Dado** la cabeza de la rama, **Cuando** corre `TestSkillsDelProductoDesdeControl`, **Entonces** `normativa`,
   `jurisprudencia` y `plazos-y-territorio` son lo que dan su base y su declaración, y su frontmatter el de su
   control, en los dos con la `description` de la variante en lugar de la suya, y sale en verde (FR-003, FR-005).
2. **Dado** la cabeza de la rama, **Cuando** corre la comprobación de `consulta-juridica`, **Entonces** cada fichero
   de `skills/consulta-juridica/` es, ruta a ruta y byte a byte, el de la base de la rama, salvo las líneas de la
   clave `description` de su `SKILL.md`, y no falta ni sobra ninguno (FR-003, SC-002).
3. **Dado** una copia de `skills/` con un carácter cambiado en el cuerpo, en una referencia o en `name` de una de las
   cuatro skills, o con un fichero de más o de menos, **Cuando** corre la comprobación de FR-003, **Entonces** sale en
   rojo con la ruta del primer fichero que difiere, falta o sobra (FR-003).

---

### User Story 3 - Las suites congeladas, el laboratorio y lo generado siguen como estaban (Priority: P1)

El mantenedor ve `go test ./...` y `make ci` en verde con un único cambio declarado en una suite congelada, la
variante de laboratorio intacta para comparar y nada generado que regenerar.

**Why this priority**: las suites de H39 a H41.3 protegen lo que ya funciona; el hito declara exactamente qué cambia
en ellas.

**Independent Test**: `make ci` en verde en la cabeza de la rama, y el diff de la rama contra la base no toca
`laboratorio/`, `web/`, `evals/`, `.github/workflows/evals.yml` ni ningún test congelado salvo
`TestSkillsDelProductoDesdeControl`.

**Acceptance Scenarios**:

1. **Dado** la cabeza de la rama, **Cuando** corre `go test ./...`, **Entonces** sale en verde, y
   `TestDescripcionesCortas`, `TestConLaDescripcion`, `TestVariantesDeControl`, `TestSkillsDelRepositorio` y los
   mutantes de `eleccion_test.go` pasan sin cambiar ningún valor esperado (FR-006, FR-007, FR-009).
2. **Dado** la cabeza de la rama, **Cuando** corre `make skills-sync`, **Entonces** no cambia ningún fichero (FR-007).
3. **Dado** la cabeza de la rama, **Cuando** se lee `CHANGELOG.md`, **Entonces** «[Unreleased]» → «Cambiado» lleva una
   entrada que nombra las cuatro skills y dice que su `description` pasa a ser la de la variante
   `producto-descripciones-cortas` (FR-008).

---

### User Story 4 - La persona decide con el sondeo y con las evals del cierre (Priority: P2)

La persona lanza fuera del run el sondeo de aceptación en las condiciones del ADR 0055 (punto 4), añade la tabla y su
lectura a la propuesta de cambio y fusiona solo si se cumple el criterio y las evals y el juicio de la tercera persona
salen en verde en el cierre.

**Why this priority**: es la aceptación del hito, pero la ejecuta una persona y no el run.

**Independent Test**: la propuesta de cambio lleva la medición del cierre con las evals de las cuatro skills en los
dos modos y el juicio de la tercera persona; la tabla del sondeo la añade la persona.

**Acceptance Scenarios**:

1. **Dado** la propuesta de cambio del hito, **Cuando** el cierre mide las evals, **Entonces** el job mide las cuatro
   skills del producto en los dos modos, porque cambia una descripción (ADR 0045), y juzga la tercera persona de las
   cuatro descripciones (ADR 0046, punto 4); el hito solo se acepta si todo sale en verde (FR-010, SC-005).
2. **Dado** que una eval o el juicio de la tercera persona caen en el cierre, **Cuando** corre la reparación,
   **Entonces** no esquiva la caída ni cambia ninguna descripción para que pase, y deja dicha la caída para la persona
   (FR-010).
3. **Dado** dos filas de 320 sesiones de la eval 04 en modo herramienta lanzadas el mismo día, una detrás de otra,
   `control` desde un árbol de trabajo en la base de la rama y `producto` desde la cabeza, **Cuando** la persona lee la
   tabla, **Entonces** el hito solo se fusiona si «No activan `normativa` antes de la primera herramienta» queda en
   `producto` por debajo de `control` con intervalos de Wilson al 95 % que no se tocan (FR-011, SC-006).

---

### Edge Cases

- **Una quinta skill en `skills/` sin su texto en la variante** (un hito posterior): la comprobación de FR-001 sale en
  rojo y nombra la skill, igual que hoy `TestDescripcionesCortas` (H41.3) con la variante.
- **La variante cambia en `main` después de la base**: `TestDescripcionesCortas` sale en rojo con la primera línea que
  difiere; el hito no la cambia (FR-006).
- **Lo que llega de fuera manipulado a mano** (un `SKILL.md` sin frontmatter, una variante ilegible): lo cubre la regla
  genérica, defecto `inesperado` con código 1 (ADR 0023), sin casos propios.

## Requirements *(mandatory)*

### Functional Requirements

- **FR-001**: En cada una de las cuatro skills de `skills/`, la `description` que da `skills.LeerFrontmatter` es,
  byte a byte, el contenido de `laboratorio/producto-descripciones-cortas/<skill>.txt` sin su salto de línea final
  (los cuatro textos de `TestDescripcionesCortas`). Control: una comprobación nueva de `make ci`, sin red ni modelo,
  que lee las descripciones de `skills/` y los textos de la variante; sale en rojo nombrando la skill si una difiere en
  un byte y si una skill de `skills/` no tiene texto en la variante.
- **FR-002**: En cada `skills/<skill>/SKILL.md`, la clave sigue siendo `description: >-` en su sitio de hoy, detrás de
  `name:`, con el texto partido en más de una línea sangrada, como en la base. Control: la comprobación de FR-001 sale
  en rojo, nombrando la skill, si la clave de una no es un bloque `>-` de más de una línea.
- **FR-003**: Fuera de las líneas de la clave `description` de su `SKILL.md`, ninguna de las cuatro skills cambia:
  ni el cuerpo del `SKILL.md`, ni las demás claves del frontmatter (`name`, `metadata`), ni sus referencias, escritas a
  mano o generadas, ni sus nombres, y no se añade ni se quita ningún fichero. Controles:
  - para `normativa`, `jurisprudencia` y `plazos-y-territorio`, `TestSkillsDelProductoDesdeControl` con el cambio de
    FR-005;
  - para `consulta-juridica`, que no tiene variante de control, una comprobación nueva de `make ci`, sin red, que
    compara `skills/consulta-juridica/` con su contenido en la base de la rama (`ab951c22`), fijado en el repositorio:
    ruta a ruta y byte a byte, y el `SKILL.md` salvo las líneas de la clave `description`. Sale en rojo con la ruta
    del primer fichero que difiere, falta o sobra.
- **FR-004**: Cada `description` de `skills/` cumple los límites del ADR 0046 —como mucho 900 caracteres y ninguna
  etiqueta XML— y los de la variante —como mucho 450 cada una y 1700 entre las cuatro—, contados en caracteres (runas)
  como en H41.3. Las de la rama miden 417, 420, 426 y 427, y suman 1690. Controles: `TestSkillsDelRepositorio`
  (`ValidarFrontmatter`: en rojo con la skill, la medida y el límite si una pasa de 900 o lleva una etiqueta XML) y la
  comprobación de FR-001, que mide sobre `skills/` los 450 por descripción y los 1700 entre las cuatro y sale en rojo
  con la skill o la suma, la medida y el límite. La tercera persona no se comprueba en `make ci`: la juzga el job de
  evals (FR-010; ADR 0046, punto 4).
- **FR-005**: El único cambio en las suites congeladas de H39 a H41.3 es este, en `TestSkillsDelProductoDesdeControl`,
  para `normativa`, `jurisprudencia` y `plazos-y-territorio`, y solo en lo esperado de la `description`:
  - en el subtest «desde control», lo esperado del `SKILL.md` sigue siendo el de su base con los cambios de su
    declaración —la base es su control en `normativa` y `jurisprudencia`, y `plazos-y-territorio-sin-condicion` en
    `plazos-y-territorio`; los cambios del cuerpo declarados no cambian—, con la `description` de la variante en lugar
    de la suya;
  - en el subtest «frontmatter», lo esperado es el frontmatter de su control con la `description` de la variante en
    lugar de la suya.

  En los dos, fuera de las líneas de la clave `description`, byte a byte, y la `description` que lee
  `skills.LeerFrontmatter`, el texto de la variante. Lo demás del test no cambia: los otros
  ficheros, los movimientos y los mensajes de rojo. Sigue saliendo en rojo con la ruta del primer fichero que difiere,
  falta o sobra, también con un carácter cambiado en el cuerpo, en `name` o en `metadata` de un `SKILL.md`. Es lo que
  hizo H41.1 con el subtest de H41 que comparaba `plazos-y-territorio` con su control.
- **FR-006**: `laboratorio/` no cambia un byte: la variante `producto-descripciones-cortas`, las variantes de control y
  el arnés. Controles: `TestDescripcionesCortas` y `TestVariantesDeControl` (huellas de los ficheros de control), sin
  cambiar ningún valor esperado. `INSTALACION=producto-descripciones-cortas` sigue aceptándose y pasa a instalar las
  mismas descripciones que `INSTALACION=producto`.
- **FR-007**: Nada generado del repositorio cambia: `make skills-sync` no modifica ningún fichero (`TestSkillsDelRepositorio`
  sale en rojo si lo generado no es lo que da), y el hito no toca `web/`, `plugin/`, `mcp/` ni el empaquetado. El
  plugin y el `.mcpb` copian `skills/` al empaquetar, y el `llms.txt` de la web lee las descripciones de la versión
  publicada, así que cambian con la release siguiente.
- **FR-008**: `CHANGELOG.md` lleva, en «[Unreleased]» → «Cambiado», una entrada que nombra las cuatro skills, dice que
  su `description` pasa a ser la de la variante `producto-descripciones-cortas` medida en H41.3, que nada más de las
  skills cambia y que el cambio es para quitar la caída de la eval 04 de `normativa` en modo herramienta (ADR 0055).
  Se verifica en el quickstart y en la revisión, sin control en `make ci`: la release siguiente cierra «[Unreleased]».
- **FR-009**: Las suites congeladas de H39 a H41.3 siguen en verde con el único cambio de FR-005; `go test ./...` y
  `make ci` salen en verde en la cabeza de la rama. `TestConLaDescripcion` y el mutante de `eleccion_test.go` que
  cambia la descripción de `plazos-y-territorio` siguen en verde sin cambios, porque la clave sigue en un bloque `>-`
  (FR-002).
- **FR-010**: En la medición del cierre, el job de evals mide las evals de las cuatro skills del producto en los dos
  modos, porque cambia una descripción (ADR 0045), y juzga la tercera persona de las cuatro descripciones (ADR 0046,
  punto 4); el hito pide que todo salga en verde. Control: el job sale en rojo si cae una eval o el juicio de la
  tercera persona de una descripción. Si cae, la reparación no esquiva la caída ni cambia ninguna descripción para que
  pase: lo deja dicho para la persona en el informe.
- **FR-011**: La aceptación es el sondeo de la persona del ADR 0055 (punto 4), que el run no lanza: `normativa-producto`,
  solo la eval 04, solo el modo herramienta, `claude-sonnet-5-5` y la versión de Claude Code del job, en dos filas de
  320 sesiones lanzadas el mismo día, una detrás de otra —**control**, `INSTALACION=producto` desde un árbol de trabajo
  en la base de la rama, fuera del repositorio; **producto**, `INSTALACION=producto` desde la cabeza de la rama—, con
  su tabla en `evidencias/`. El criterio es «No activan `normativa` antes de la primera herramienta» (*n* menos
  «Activan antes»): en `producto`, por debajo de `control`, con intervalos de Wilson al 95 % que no se tocan. «Activan
  otra skill» se da en la tabla y no decide. El quickstart del hito da las dos órdenes del sondeo con las herramientas
  que ya existen (`make evals-laboratorio`, `make laboratorio-tabla`).

### Key Entities

- **Descripción de una skill**: el valor de la clave `description` del frontmatter de `skills/<skill>/SKILL.md`, tal
  como lo lee `skills.LeerFrontmatter`; en el fichero, un bloque `>-` partido en líneas.
- **Variante `producto-descripciones-cortas`**: `laboratorio/producto-descripciones-cortas/<skill>.txt`, un texto por
  skill de `skills/`, lo medido en H41.3; no cambia.
- **Base de la rama**: `ab951c227e1fc8acac8bd6b23534e275e4a36a76`, el `main` del que sale la rama; de ella salen lo
  esperado de `consulta-juridica` y la fila `control` del sondeo.
- **Control de una familia**: la variante `<skill>-control` de `laboratorio/` de `normativa`, `jurisprudencia` y
  `plazos-y-territorio`, con la que compara `TestSkillsDelProductoDesdeControl`.

### Uso, de fuera adentro

- **Las cuatro descripciones** las consume el agente (Claude Code u otro host) en el listado de skills que lee al
  empezar cada sesión: una vez por sesión, no por consulta ni por norma. Tamaño: 1690 caracteres entre las cuatro
  (como mucho 1700, FR-004), frente a los 3313 de hoy que da H41.3 (`specs/037-h41-3-por-que-normativa/spec.md`,
  «Contexto»; no medido en esta sesión). No crece con el uso: con meses de uso diario, cientos de normas y miles de
  bloques consultados, el listado sigue siendo de cuatro textos fijos, porque ni la caché ni el grafo entran en él. Lo
  que gana la skill: `normativa` se activa antes de `boe_articulo` u otra herramienta `boe_…`, la entrada de su paso 1.
  No hay señales que caduquen: el texto cambia solo con una release.
- **El `CHANGELOG.md`** lo lee la persona que actualiza kitlegal, una vez por release; una entrada de un párrafo.
- **La tabla del sondeo** la produce `make laboratorio-tabla`, que no cambia, y la lee la persona una vez, para decidir
  la fusión: dos filas y su detalle.
- **Las comprobaciones de `make ci`** (FR-001 a FR-005) las lee el mantenedor cuando salen en rojo: una línea por
  defecto, con la skill o la ruta.

## Success Criteria *(mandatory)*

### Measurable Outcomes

- **SC-001**: 4 de 4 descripciones de `skills/` son, byte a byte, las de la variante, con 0 bytes de diferencia.
  Control: la comprobación de FR-001 en `make ci`, en rojo con un byte de diferencia.
- **SC-002**: 0 bytes cambian en las cuatro skills fuera de las líneas de la clave `description`, respecto a lo que
  dan su base y su declaración con la descripción de la variante (`normativa`, `jurisprudencia`,
  `plazos-y-territorio`) o a la base de la rama (`consulta-juridica`), y 0 ficheros se añaden o se quitan. Control: `TestSkillsDelProductoDesdeControl` (FR-005) y
  la comprobación de FR-003 en `make ci`, en rojo con el primer fichero que difiere, falta o sobra.
- **SC-003**: cada descripción de `skills/` tiene como mucho 450 caracteres (y por tanto menos de 900), las cuatro
  suman como mucho 1700 y llevan 0 etiquetas XML; en la rama, 417, 420, 426 y 427, y 1690. Control: la comprobación de
  FR-001 y `TestSkillsDelRepositorio` en `make ci`, en rojo si una medida pasa de su límite.
- **SC-004**: `go test ./...` y `make ci` en verde en la cabeza de la rama, sin red ni modelo, con
  `TestSkillsDelProductoDesdeControl` como única suite congelada cambiada (FR-005, FR-009).
- **SC-005**: en la medición del cierre, las evals de las cuatro skills salen en verde en los dos modos y el juicio de
  la tercera persona sale en verde en las cuatro descripciones. Control: el job de evals sale en rojo con la eval o la
  descripción que cae (FR-010).
- **SC-006**: en el sondeo de la persona, con 320 sesiones por fila, «No activan `normativa` antes de la primera
  herramienta» en `producto` queda por debajo de `control`, con intervalos de Wilson al 95 % que no se tocan. Lo mide y
  lo decide la persona, no `make ci` ni el job (FR-011); si no se cumple, la propuesta no se fusiona.

## Criterios del hito (literales)

**Controles** (del fichero del hito):

- `make ci` comprueba, sin red ni modelo:
  - que la `description` de cada skill de `skills/` es la de la variante, byte a byte;
  - que lo demás de `normativa`, `jurisprudencia` y `plazos-y-territorio` es lo que dan su base y su declaración
    (`TestSkillsDelProductoDesdeControl`, de H41 y H41.1), y que lo demás de `consulta-juridica`, que no tiene
    variante de control, es lo de la base de la rama;
  - que las descripciones cumplen los límites del ADR 0046 (tercera persona sin etiquetas XML, como mucho 900
    caracteres) y los de la variante (450 cada una y 1700 entre las cuatro);
- las suites congeladas de H39 a H41.3 siguen en verde, con un solo cambio, declarado aquí. En
  `TestSkillsDelProductoDesdeControl`, el `SKILL.md` de esas tres skills deja de compararse con el de su control tal
  cual, en el frontmatter y en el fichero entero, y pasa a compararse con el de su control con la `description` de
  la variante. Es lo que hizo H41.1 con el subtest de H41 que comparaba `plazos-y-territorio` con su control. Con las
  descripciones nuevas en un bloque `>-`, es el único de `go test ./...` que se pone en rojo.

Las evals de las cuatro skills del producto salen en verde en la medición del cierre, en los dos modos. El job las
mide todas porque cambia una descripción (ADR 0045), y juzga la tercera persona de las cuatro (ADR 0046, punto 4).
Si caen, la reparación no esquiva la caída ni cambia ninguna descripción para que pase: lo deja dicho para la persona.

Se cubren con FR-001 a FR-005, FR-009 y FR-010, y SC-001 a SC-005. «Tercera persona sin etiquetas XML»: las etiquetas
XML y los 900 caracteres los comprueba `make ci` (ADR 0046, puntos 2 y 3); la tercera persona, el juez del job
(punto 4), como dice el propio hito en el párrafo de las evals.

**Aceptación** (del fichero del hito; ADR 0055, punto 4): el sondeo de la persona deja su tabla en `evidencias/`. Las
condiciones son `normativa-producto`, solo la eval 04, solo el modo herramienta, `claude-sonnet-5-5` y la versión de
Claude Code del job, en dos filas de 320 sesiones lanzadas el mismo día, una detrás de otra:

- **control**: `INSTALACION=producto`, lanzado desde un árbol de trabajo en la base de la rama, fuera del
  repositorio, para que instale `skills/` de `main`;
- **producto**: `INSTALACION=producto`, lanzado desde la cabeza de la rama.

El criterio es **«No activan `normativa` antes de la primera herramienta»**, que es *n* menos «Activan antes». En
`producto` tiene que quedar por debajo de `control`, con intervalos de Wilson al 95 % que no se tocan. «Activan otra
skill» se da en la tabla y no decide; si la que se activa es otra que `consulta-juridica`, el ADR que lea el sondeo
lo dice. La tabla y su lectura se añaden a la propuesta de cambio antes de fusionarla.

**Si el criterio no se cumple, o si fallan las evals o el juicio de la tercera persona, la propuesta no se fusiona.**
El sondeo no se repite para ver si sale en verde (ADR 0047, punto 4), y el ADR que lo lea dice qué se mide después.
Con 320 por fila, un remedio que funcione como `sola` pasa en torno a 49 de cada 50 sondeos, y uno sin efecto, en
torno a 3 de cada 1000 (ADR 0055, «Validación»).

El run no lanza el sondeo.

Se cubre con FR-011 y SC-006.

## Assumptions

- **La base de la rama es `ab951c22`** (`git merge-base main HEAD` en esta sesión). Lo esperado de `consulta-juridica`
  (FR-003) se fija desde ese commit en el repositorio, porque `make ci` no tiene red ni garantiza el historial de git.
- **«Byte a byte» es el texto que se lee de la clave**, el de `skills.LeerFrontmatter`, como dice el hito; los bytes
  de las líneas del bloque `>-` no son los del `.txt`, porque el bloque parte el texto en líneas.
- **Cómo se parte el bloque `>-`** (dónde se corta cada línea) no lo fija el hito; basta con que sea un bloque `>-` de
  más de una línea sangrada que se lea como el texto exacto (FR-002). Lo decide el plan.
- **La tercera persona no tiene control sin modelo**: es una cuestión de significado que juzga un modelo (ADR 0037 y
  0046, punto 4); `make ci` comprueba lo que tiene forma.
- **El job de evals y su modelo no cambian** (ADR 0031): `MODELO_DE_EVALS` y `VERSION_DE_CLAUDE_CODE` son los de
  `main`, Claude Code 2.1.326.

## Fuera de alcance

- Cambiar los cuerpos, las referencias o los nombres de las skills, o una descripción en algo distinto de la variante
  medida, aunque sea una palabra (ADR 0055, opción 3).
- Cambiar el umbral de `sin_activar`, cómo se cuenta, o ningún otro umbral, límite ni tope (ADR 0038, punto 1).
- Cambiar el job de evals, lo que instala, `VERSION_DE_CLAUDE_CODE` o `MODELO_DE_EVALS` (ADR 0031).
- Fijar `SLASH_COMMAND_TOOL_CHAR_BUDGET`, o cualquier ajuste de Claude Code, en el equipo de quien instala kitlegal o
  en `kitlegal skills install` (ADR 0055, punto 2).
- Pedir la activación en las instrucciones del servidor o en `consulta-juridica` (ADR 0050, punto 3).
- Quitar del producto `consulta-juridica` u otra skill (ADR 0048).
- Cambiar el arnés del laboratorio, la variante `producto-descripciones-cortas` o las variantes de control.
- Evals nuevas, y cambiar las que ya existen.
- Cambiar otra suite congelada que no sea `TestSkillsDelProductoDesdeControl`, o en ella algo distinto de FR-005.
- Tocar `web/`, el plugin, el `.mcpb` o el empaquetado: cambian con la release siguiente.
- Lanzar el sondeo de aceptación, escribir su tabla en `evidencias/` o el ADR que lo lea: los hace la persona, fuera
  del run.
- Etiquetar, publicar una release o fusionar (ADR 0018, 0020 y 0021).
- Comprobar `CHANGELOG.md` en `make ci` o en la suite de aceptación.
- Medir o proteger aparte la activación de `jurisprudencia`, `plazos-y-territorio` o `consulta-juridica` más allá de
  lo que ya mide el job: el ADR 0055 lo acepta como riesgo.
