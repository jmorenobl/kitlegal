# Feature Specification: H46 · Los cuerpos de las skills de capacidad, con sitio para las materias (ADR 0063)

**Feature Branch**: `044-h46-los-cuerpos-de`

**Created**: 2026-11-25

**Status**: Draft

**Input**: el fichero del hito, `docs/roadmap/in-progress/H46-los-cuerpos-de-las-skills-con-sitio-para-las-materias.md`,
que aplica el ADR 0063 entero
(`docs/ADR/0063-el-cuerpo-de-una-skill-de-capacidad-deja-sitio-para-las-materias.md`), el que sustituye el punto 3 del
ADR 0062. El run no cambia ningún ADR ni escribe otro.

## Contexto

Una materia entra en kitlegal como paquete, sin tocar código ni protocolo (principio X; ADR 0044), y cada paquete añade
una línea al cuerpo de cada skill de capacidad que la recibe (ADR 0062). Hoy no caben dos: el paquete de H42 sin ningún
`.go`, sobre la cabeza de H45.1, deja `make ci` en rojo en el guion `h45-paquete-de-mas`, con el cuerpo de `normativa`
en 11 030 caracteres de 11 000 (cifras del hito y del ADR 0063). El hito adelgaza los cuerpos de `normativa`,
`plazos-y-territorio` y `jurisprudencia` con progressive disclosure —en el cuerpo, solo lo que toda respuesta
necesita—, reserva en cada uno 1400 caracteres para las líneas de los paquetes (unas veinte materias de nombre corto,
a 52 + 3n caracteres cada línea) y hace que `make ci` compruebe esa reserva. Lo que gana cada skill de capacidad es
que la materia siguiente cabe sin recortar su protocolo; sus respuestas no cambian, y lo deciden las evals.

**Comprobado en el repositorio al escribir este spec** (lecturas y órdenes de esta sesión, rama
`044-h46-los-cuerpos-de` desde `main`, `78bb6f76`):

- La rama `cuerpos-prototipo` existe y su cabeza es `f89093dd`; `git diff main f89093dd -- skills/` cambia, y solo
  cambia, `skills/normativa/SKILL.md`, `skills/plazos-y-territorio/SKILL.md`,
  `skills/plazos-y-territorio/references/municipios.md`, `skills/jurisprudencia/SKILL.md`,
  `skills/jurisprudencia/references/consulta_preparada.md` y `skills/jurisprudencia/references/documento_aportado.md`.
  El prototipo no toca `CHANGELOG.md` ni `docs/SKILLS.md`.
- Las tres skills de capacidad llevan el par de marcas de los paquetes `<!-- paquetes: lo genera make skills-sync -->`
  y `<!-- fin de los paquetes -->` (`skills/normativa/SKILL.md:25-26`, `skills/jurisprudencia/SKILL.md:33-34`,
  `skills/plazos-y-territorio/SKILL.md:23`). `consulta-juridica` lleva otro par, el de los paquetes de la base
  (`<!-- paquetes de la base: lo genera make skills-sync -->`), que no es este (`internal/skills/paquetes.go:42-51`).
- `defectosDelCuerpo` (`internal/skills/sincronia.go:606`) da hoy dos defectos, el de 500 líneas o más y el de más de
  11 000 caracteres (`maximoDeCaracteresDelCuerpo`, `:31`), en ese orden; unas marcas que no van en pareja ya dan su
  defecto (`internal/skills/paquetes.go:90`).
- `casosDePresupuestos` (`internal/app/skills_test.go:826`) tiene cuatro casos: 499 y 500 líneas, 11 000 y 11 001
  caracteres. `comprobarElSitioDeOtroPaquete` está en `internal/skills/vertical_de_ensayo_test.go:261`.
- La copia congelada `specs/035-h41-1-plazos-y-territorio/aceptacion/skills-instaladas.txtar` es igual byte a byte a
  `internal/app/testdata/script/h41-1-skills-instaladas.txtar` (`git diff --no-index` vacío); la de H41,
  `specs/034-h41-las-skills-del/aceptacion/skills-instaladas.txtar`, ya difiere hoy de
  `internal/app/testdata/script/h41-skills-instaladas.txtar` (30 líneas).
- Ningún `SKILL.md` ni ningún fichero de `skills/` lleva un número de versión propio; `CHANGELOG.md` registra los
  cambios de una skill por su nombre, bajo *Unreleased* (p. ej. la entrada de H41.1, `CHANGELOG.md:246`).
- **No medido en esta sesión**: los caracteres de los cuerpos. El sandbox de la sesión no dejó correr el guion de
  medida. Las cifras de este spec (10 750, 10 686, 10 842 en `main`; 9498, 9577 —9483 sin la línea de lo
  administrativo— y 9473 con el prototipo; 9568 y 9647 con el paquete fiscal) son las del hito y del ADR 0063, y la
  medida que cuenta es la del gate, en `make ci` (FR-007, SC-002).

## Clarifications

### Session 2026-11-25

Sin preguntas en la redacción: lo que el hito deja a elegir va a «Assumptions».

## User Scenarios & Testing *(mandatory)*

### User Story 1 - La tabla de comandos lleva lo que se necesita antes de la primera llamada (Priority: P1)

El agente que activa `normativa`, `plazos-y-territorio` o `jurisprudencia` lee en el cuerpo, antes de su primera
llamada, qué orden o qué herramienta usar para cada operación y qué hace. Lo que devuelve cada llamada, el sobre y las
banderas comunes los ve en el resultado de cada llamada y con `--describe`; ya no le ocupan el cuerpo.

**Why this priority**: es la mayor parte del sitio que se gana (en `normativa`, 1237 de los 1400, según el hito), y
sin ella no caben las materias.

**Independent Test**: con las skills instaladas por `kitlegal skills install`, la tabla de cada skill de capacidad
tiene tres columnas y ninguna línea del sobre ni de las banderas comunes.

**Acceptance Scenarios**:

1. **Dado** las skills instaladas con `kitlegal skills install` desde el binario de la rama, **Cuando** se lee la
   región de la tabla de comandos del `SKILL.md` de `normativa`, `plazos-y-territorio` y `jurisprudencia`,
   **Entonces** cada cabecera de tabla de la región es `| Orden | Herramienta | Qué hace |` seguida de `|---|---|---|`,
   cada fila tiene tres celdas, y la región no contiene `Qué devuelve en`, ni una línea que empiece por `La orden y la
   herramienta de cada fila devuelven el mismo sobre`, ni una que empiece por `Banderas comunes:`.
2. **Dado** el árbol de la rama, **Cuando** se corre `make skills-sync` y después `make skills-check`, **Entonces**
   `skills-sync` no cambia ningún fichero y `skills-check` sale con 0, sin defectos ni deriva.
3. **Dado** dos verbos de una misma tabla con sobres distintos, o con banderas comunes distintas, **Cuando** se
   renderiza la tabla, **Entonces** sale el mismo error que hoy, aunque la tabla ya no escriba ni el sobre ni las
   banderas.
4. **Dado** un documento de `--describe` con un defecto que hoy da `LeerDescripcionDeVerbo`, **Cuando** se lee,
   **Entonces** da el mismo defecto que en `main`.

---

### User Story 2 - El sitio de las materias lo comprueba `make ci` (Priority: P1)

Quien cambia el protocolo de una skill de capacidad (una revisión final, un hito de capacidad) se encuentra con un rojo
de `make ci` si se come el sitio de las materias, y quien escribe un paquete (H42 y los siguientes) tiene 1400
caracteres para su línea sin que nadie los haya gastado.

**Why this priority**: sin el control, la reserva es una cifra de este hito que el siguiente cambio de protocolo
consume sin aviso.

**Independent Test**: `casosDePresupuestos` en `make ci`, con sus límites, y el registro de mutantes del hito.

**Acceptance Scenarios**:

1. **Dado** una copia del repositorio con el cuerpo de una skill con marcas ensanchado fuera de ellas hasta 9600
   caracteres sin las líneas de los paquetes, **Cuando** se corre la comprobación de las skills, **Entonces** no hay
   ningún defecto.
2. **Dado** la misma copia ensanchada hasta 9601 sin las líneas de los paquetes (y por debajo de 11 001 con ellas),
   **Cuando** se corre la comprobación, **Entonces** hay exactamente un defecto,
   `<skill>: SKILL.md: cuerpo sin las líneas de los paquetes de 9601 caracteres (máximo 9600)`.
3. **Dado** la misma copia ensanchada hasta 11 001 caracteres de cuerpo, **Cuando** se corre la comprobación,
   **Entonces** hay exactamente dos defectos: `<skill>: SKILL.md: cuerpo de 11001 caracteres (máximo 11000)` y el de
   FR-005 con los caracteres del cuerpo sin las líneas de los paquetes.
4. **Dado** una copia con el cuerpo de `consulta-juridica` (sin el par de marcas de FR-005) ensanchado a 11 000 y a
   11 001, **Cuando** se corre la comprobación, **Entonces** da lo que hoy: ningún defecto con 11 000 y solo el de
   11 001 caracteres con 11 001.
5. **Dado** el árbol de la rama, **Cuando** se corre `TestVerticalDeEnsayo`, **Entonces** pasa sin comprobar que cabe
   otra línea de paquete de su misma longitud, y comprueba todo lo demás que comprueba en `main`.
6. **Dado** el árbol de la rama con el mutante del defecto nuevo (la regla que no se aplica, o con otro tope),
   **Cuando** se corre `make ci`, **Entonces** sale en rojo en el subtest `presupuestos`.

---

### User Story 3 - El detalle de una clase de pregunta, en la referencia que esa clase lee sin condición (Priority: P1)

El agente de `plazos-y-territorio` lee `municipios.md` antes de su primera orden o llamada, y el de `jurisprudencia`
lee `consulta_preparada.md` antes de pedir `cita_preparar` y `documento_aportado.md` antes de cada cotejo. Lo que solo
necesita esa clase de pregunta está en esa referencia, no en el cuerpo; lo que necesita toda respuesta —las líneas `⚠`,
las formas fijas, las reglas invariantes y las órdenes de leer— sigue en el cuerpo.

**Why this priority**: es el resto del sitio en `plazos-y-territorio` y `jurisprudencia`, y lo que deja cada cuerpo
bajo 9600.

**Independent Test**: la suite de aceptación sobre las skills instaladas, y `TestSkillsDelProductoDesdeControl` con
cada cambio declarado.

**Acceptance Scenarios** (cada frase citada se busca con cualquier espacio o salto de línea entre sus palabras, como
la parte el texto en sus líneas de 120 columnas):

1. **Dado** las skills instaladas con `kitlegal skills install`, **Cuando** se lee `plazos-y-territorio`,
   **Entonces** el cuerpo contiene «`kitlegal` se invoca desde el `PATH`, y cada herramienta, por el nombre de su
   fila.» y no contiene «Con la herramienta de su fila si entre las tuyas hay una con ese nombre»; y
   `references/municipios.md` termina con un apartado `## Cómo se presenta cada dato` seguido de un apartado
   `## Herramienta u orden`, y este contiene «Con la herramienta de su fila» y los códigos 2 (`argumentos`) y
   3 (`no-encontrado`).
2. **Dado** las skills instaladas, **Cuando** se lee el apartado «Cómo se presenta el territorio» del cuerpo de
   `plazos-y-territorio`, **Entonces** contiene el bloque `text` con las tres líneas `boletin_autonomico: <valor>`,
   `boletin_provincial: <valor>` y `dir3: <valor>`, y no contiene «un boletín, por su `codigo` y su `nombre`
   exactos», que está en `## Cómo se presenta cada dato` de `municipios.md`.
3. **Dado** las skills instaladas, **Cuando** se lee la orden de leer `municipios.md` del cuerpo de
   `plazos-y-territorio`, **Entonces** dice que trae `cómo se pide cada
   operación`, y el primer párrafo de `municipios.md` no remite al cuerpo para la forma de la cobertura.
4. **Dado** las skills instaladas, **Cuando** se lee `jurisprudencia`, **Entonces** el cuerpo no contiene
   `| Cómo se dio | Qué se pide |` ni `lleva el **número de resolución**, no el ROJ`; contiene la orden de leer
   `consulta_preparada.md` con `qué se pide según cómo se dio la referencia` seguida de `Un número sin fecha no se
   prepara de ninguna forma, tampoco como ROJ: lleva la línea y se pide la fecha.`; y
   `references/consulta_preparada.md` termina con un apartado `## Lo que se pide` que contiene la tabla
   `| Cómo se dio | Qué se pide |` y la viñeta del número de resolución.
5. **Dado** las skills instaladas, **Cuando** se lee `jurisprudencia`, **Entonces** el cuerpo no contiene `da además
   la referencia en la forma en que se pidió` ni `Las sentencias que el documento menciona no se citan`, ni el inciso
   `también cuando el documento no es el pedido`; y `references/documento_aportado.md` termina con un apartado
   `## La referencia pedida y las sentencias que menciona` con las dos viñetas, y su apartado «Lo que devuelve»
   contiene `La orden sale con 0 también cuando el documento no es el pedido`.
6. **Dado** las skills instaladas, **Cuando** se lee `normativa`, **Entonces** su cuerpo dice «desde el `PATH` o con
   su herramienta» y no contiene «si la tienes, con su herramienta».
7. **Dado** el árbol de la rama, **Cuando** se compara `skills/` con `f89093dd`, **Entonces** no hay ninguna
   diferencia.

---

### User Story 4 - Lo que fija el texto de hoy sigue fijándolo, con el texto nuevo (Priority: P2)

Quien lleva el proyecto quiere que los controles que hoy fijan el texto de las skills —la declaración de lo que cambia
cada skill respecto a su control, las frases de sus remisiones y los guiones que buscan sus órdenes— sigan comprobando
lo mismo, al día con los textos nuevos.

**Why this priority**: sin ellos, `make ci` sale en rojo con los textos nuevos (por eso van en la misma tarea); su
valor es que el siguiente cambio de texto siga teniendo control.

**Independent Test**: `TestSkillsDelProductoDesdeControl`, `TestRemisionesDelRepositorio` y los guiones `h41-` en
`make ci`.

**Acceptance Scenarios**:

1. **Dado** el árbol de la rama, **Cuando** se corre `TestSkillsDelProductoDesdeControl`, **Entonces** pasa, con un
   cambio declarado por cada texto cambiado o movido (en el cuerpo o en una referencia) con los bytes del prototipo.
2. **Dado** el árbol con uno de esos cambios deshecho en `skills/`, **Cuando** se corre
   `TestSkillsDelProductoDesdeControl`, **Entonces** sale en rojo.
3. **Dado** el árbol de la rama, **Cuando** se corre `TestRemisionesDelRepositorio`, **Entonces** pasa con las dos
   frases nuevas de las órdenes de leer `municipios.md` y `consulta_preparada.md`, y sale en rojo con la frase de
   `main` en cualquiera de las dos.
4. **Dado** el árbol de la rama, **Cuando** corren `h41-skills-instaladas` y `h41-1-skills-instaladas`, **Entonces**
   pasan: el primero encuentra la orden de `consulta_preparada.md` con su frase nueva en el cuerpo y la tabla
   «Cómo se dio / Qué se pide» en la referencia; el segundo, la orden nueva de `municipios.md`; y la copia congelada
   del segundo es igual byte a byte al guion.

---

### User Story 5 - Las respuestas no cambian (Priority: P1)

La persona que pregunta recibe las mismas respuestas que con `main`: lo deciden las evals del cierre, sobre todas las
skills y sin reutilizar nada de `main`.

**Why this priority**: es la condición del ADR 0063 para mover texto a una referencia.

**Independent Test**: el job de evals del cierre.

**Acceptance Scenarios**:

1. **Dado** la cabeza de la propuesta de cambio del hito, **Cuando** corre el cierre de evals de todas las skills sin
   reutilizar nada de `main`, **Entonces** ninguna clase que decide queda por debajo de su umbral y ninguna serie de
   `claude-sonnet-5-5` queda por debajo de 2 de 3 en una eval que decide.

---

### Edge Cases

- **Una skill con marcas y líneas de paquete**: `plazos-y-territorio` lleva hoy la de lo administrativo (94 caracteres
  según el hito); la regla mide el cuerpo sin ellas (9483 según el hito), y el tope de 11 000 mide el cuerpo entero.
- **Exactamente en el límite**: 9600 sin las líneas de los paquetes está en verde; 9601, en rojo (FR-005).
- **Una skill sin el par de marcas de los paquetes** (`consulta-juridica`, o una skill nueva sin ellas): no tiene el
  defecto nuevo, ni con 11 000 caracteres (FR-006).
- **Las marcas sin pareja**: dan el defecto que ya da la sincronía hoy; el defecto nuevo no añade casos.
- **Una región de paquetes vacía** (`normativa` y `jurisprudencia` hoy): el cuerpo sin las líneas de los paquetes es
  el cuerpo entero.

## Requirements *(mandatory)*

### Functional Requirements

**La tabla de comandos** (entrega 1; ADR 0063, punto 2)

- **FR-001**: La tabla de comandos que genera la sincronía en el cuerpo de cada skill lleva, por cada tabla, la
  cabecera `| Orden | Herramienta | Qué hace |`, el separador `|---|---|---|` y una fila por verbo con esas tres
  celdas, con el mismo contenido que hoy en ellas. No lleva la columna «Qué devuelve en `data`», ni la línea del sobre
  («La orden y la herramienta de cada fila devuelven el mismo sobre: …»), ni la de las banderas comunes («Banderas
  comunes: …»). Comprobable en la suite de aceptación (US1, escenario 1) y en los tests de la tabla (FR-003).
- **FR-002**: La lectura de la descripción de un verbo desde `--describe` lee lo mismo que hoy y da los mismos
  defectos; la tabla sigue exigiendo el mismo sobre y las mismas banderas comunes en todas las filas de una tabla y da
  el mismo error que hoy cuando difieren, aunque ya no los escriba. Comprobable con los tests de hoy de esos defectos,
  sin cambiar lo que esperan (US1, escenarios 3 y 4).
- **FR-003**: Los tests de la tabla fijan los bytes nuevos: `TestRenderizarTabla` (`internal/skills/comandos_test.go`),
  `TestRegenerarYComparar` y `TestCuerpoDeLaSkill` con la región de `alfa` (`internal/skills/sincronia_test.go`), y el
  fixture del subtest «control» de `internal/app/skills_test.go`, que copia la cabecera de cuatro columnas, pasa a la
  de tres. Comprobable: cada uno en rojo con la tabla de cuatro columnas.
- **FR-004**: `make skills-sync` rehace las tablas de `normativa`, `plazos-y-territorio` y `jurisprudencia`, y
  `make skills-check` sale con 0, sin defectos y sin deriva (US1, escenario 2).

**El sitio de las materias** (entrega 2; ADR 0063, punto 1)

- **FR-005**: La comprobación del cuerpo de un `SKILL.md` que lleva el par de marcas de los paquetes
  (`<!-- paquetes: lo genera make skills-sync -->` y `<!-- fin de los paquetes -->`) da un defecto más cuando el
  cuerpo, sin lo que hay entre las dos marcas, tiene más de 9600 caracteres:
  `<skill>: SKILL.md: cuerpo sin las líneas de los paquetes de <n> caracteres (máximo 9600)`, con `<n>` esos
  caracteres. Se cuentan como el presupuesto del cuerpo (runas, desde la línea que sigue al cierre del frontmatter, con
  la tabla de comandos generada y con las propias marcas); lo que no cuenta son solo las líneas entre ellas. Con 9600
  no hay defecto. Va detrás del de más de 11 000 caracteres cuando se dan los dos. Comprobable en
  `casosDePresupuestos` (FR-008) y con su mutante (SC-004).
- **FR-006**: Un `SKILL.md` sin ese par de marcas —`consulta-juridica`, con las de los paquetes de la base— no tiene
  el defecto de FR-005; sus presupuestos son los de hoy (US2, escenario 4).
- **FR-007**: Con los textos de FR-010 a FR-012, `normativa`, `plazos-y-territorio` y `jurisprudencia` no dan el
  defecto de FR-005: `make skills-check` sin defectos (SC-002).
- **FR-008**: `casosDePresupuestos` (subtest `presupuestos` de `TestSkillsDelRepositorio`, `internal/app/skills_test.go`)
  cambia para las skills con el par de marcas: el cuerpo ensanchado fuera de ellas hasta 9600 sin las líneas de los
  paquetes, sin defectos; hasta 9601, con solo el defecto de FR-005; y hasta 11 001 caracteres de cuerpo, con el de
  11 001 caracteres y el de FR-005. Para las skills sin marcas, los casos de hoy (11 000 sin defectos; 11 001 con el
  suyo). Los casos de 499 y 500 líneas no cambian. Con ese cambio salen en verde el subtest y el guion
  `h45-paquete-de-mas`, que lo corre en su paso de sincronía.
- **FR-009**: `TestVerticalDeEnsayo` deja de comprobar que cabe «otra línea de paquete de su misma longitud»
  (`comprobarElSitioDeOtroPaquete`); lo demás de ese test no cambia (US2, escenario 5).

**Los textos de `skills/`** (entrega 3; ADR 0063, punto 3)

- **FR-010**: `skills/` es, byte a byte, el de `f89093dd` (`git diff f89093dd -- skills/` vacío). Lo que eso hace en
  `plazos-y-territorio`: el párrafo de «Comandos» sobre la herramienta o la orden y los códigos va a un apartado
  «Herramienta u orden» al final de `references/municipios.md`, y en «Comandos» queda «`kitlegal` se invoca desde el
  `PATH`, y cada herramienta, por el nombre de su fila.»; de «Cómo se presenta el territorio», cómo se escribe cada
  dato y que la cobertura se explica detrás pasan a un apartado «Cómo se presenta cada dato» de la misma referencia,
  delante del anterior, y en el cuerpo quedan el apartado, una frase y las tres líneas de la cobertura; el primer
  párrafo de la referencia deja de remitir al cuerpo y dice lo que trae, y la orden de leerla dice que trae «cómo se
  pide cada operación» (US3, escenarios 1 a 3).
- **FR-011**: En `jurisprudencia`, con los bytes de FR-010: la tabla «Cómo se dio / Qué se pide» y la viñeta del
  número de resolución van a un apartado «Lo que se pide» al final de `references/consulta_preparada.md`, cuyo primer
  párrafo lo dice; la orden de leerla dice que trae «qué se pide según cómo se dio la referencia», y detrás, en el
  cuerpo, «Un número sin fecha no se prepara de ninguna forma, tampoco como ROJ: lleva la línea y se pide la fecha.».
  La viñeta de la referencia en la forma pedida y la de las sentencias que menciona el documento van a un apartado «La
  referencia pedida y las sentencias que menciona» al final de `references/documento_aportado.md`, y el inciso de
  «Comandos» sobre la salida 0 de un documento que no es el pedido, a su apartado «Lo que devuelve» (US3, escenarios
  4 y 5).
- **FR-012**: En `normativa`, con los bytes de FR-010: «desde el `PATH` o, si la tienes, con su herramienta» pasa a
  «desde el `PATH` o con su herramienta», y nada más fuera de la tabla generada (US3, escenario 6).
- **FR-013**: Ninguna línea `⚠`, ninguna forma fija (bloque `text`), ninguna regla invariante ni ninguna orden de leer
  sale de un cuerpo; ninguna descripción cambia; ninguna referencia nueva; ninguna referencia enlaza otra
  (`docs/SKILLS.md` §4); los controles de contenido (`docs/SKILLS.md` §1) no cambian. Comprobable: `make
  skills-check` y `TestSkillsDelProductoDesdeControl`, que fija la descripción y el frontmatter de cada skill.

**Lo que fija el texto en Go y en los guiones** (entregas 4 y 5)

- **FR-014**: Las declaraciones del producto de `TestSkillsDelProductoDesdeControl`
  (`internal/evals/laboratorio_familias_test.go`) llevan un cambio por texto cambiado o movido, en el cuerpo o en una
  referencia, con los bytes de FR-010; el test sigue comprobando lo mismo que hoy (US4, escenarios 1 y 2).
- **FR-015**: `remisionesDelPlan` (`internal/skills/remisiones_test.go`) lleva las dos frases nuevas: la orden de leer
  `municipios.md` y la de leer `consulta_preparada.md`, como las da el cuerpo de FR-010; `TestRemisionesDelRepositorio`
  pasa con ellas (US4, escenario 3).
- **FR-016**: `internal/app/testdata/script/h41-skills-instaladas.txtar` busca la orden de `consulta_preparada.md` con
  su frase nueva en el cuerpo y la tabla «Cómo se dio / Qué se pide» en la referencia; y
  `internal/app/testdata/script/h41-1-skills-instaladas.txtar`, la orden nueva de `municipios.md`, igual en su copia
  congelada `specs/035-h41-1-plazos-y-territorio/aceptacion/skills-instaladas.txtar`, que queda igual byte a byte al
  guion (US4, escenario 4).

**La guía y el registro** (entrega 6)

- **FR-017**: `docs/SKILLS.md` dice en §1 qué lleva la tabla de comandos, con sus tres columnas, como texto nuevo, y,
  en «La vertical de ensayo», ya no el sitio para otra línea de su misma longitud; en §3, el sitio de las materias (9600
  sin las líneas de los paquetes, 1400 para ellas), con su porqué y las cifras de los tres cuerpos al cerrar, medidas
  como el gate; y en §5, el defecto de FR-005 en su tabla con su mensaje y los 9600 en «En los límites, en verde».
- **FR-018**: `CHANGELOG.md` lleva bajo *Unreleased* una entrada del hito (H46; ADR 0063) que nombra cada skill que
  cambia —`normativa`, `plazos-y-territorio` y `jurisprudencia`— y lo que cambia en cada una, y la regla nueva de
  `skills-check`.

**Lo que no cambia**

- **FR-019**: Ningún applet devuelve otra cosa: el `--describe` de cada applet, el esquema de salida de sus
  herramientas, el sobre, los códigos de salida y la forma de `data` de cada orden son los de `main`, y los tests que
  los fijan no cambian. Lo que instala `kitlegal skills install`, y con ello sus huellas, es el contenido de `skills/`
  de FR-010; el applet `skills` lo instala como hoy. No fijan lo que devuelve un applet, sino el texto de las skills, y
  por eso cambian, los tests que nombran FR-003, FR-008, FR-009, FR-014 y FR-015, los guiones
  `internal/app/testdata/script/h41-skills-instaladas.txtar` y `h41-1-skills-instaladas.txtar` y su copia congelada
  (FR-016), y se añade la suite de FR-021. Comprobable: SC-007.

**Orden de las tareas** (del hito)

- **FR-020**: FR-010 a FR-016 van en una sola tarea `[datos]`, como T002 de H41.1
  (`specs/035-h41-1-plazos-y-territorio/tasks.md`). El orden es FR-001 a FR-004, esa tarea y, después, FR-005 a FR-009;
  FR-017 y FR-018 van con lo que describen o al final. La primera tarea escribe la suite de aceptación de FR-021. Cada
  tarea termina con `make ci` en verde (`scripts/workflow/verificar.sh`).

**La suite de aceptación**

- **FR-021**: La suite de aceptación comprueba, sobre las skills instaladas con `kitlegal skills install`, lo de US1
  (escenario 1) y US3 (escenarios 1 a 6): la tabla de cada skill de capacidad con tres columnas y sin las líneas del
  sobre ni de las banderas, las tres líneas de la cobertura en el cuerpo de `plazos-y-territorio`, y lo movido en su
  referencia y no en el cuerpo. El defecto de FR-005 no lo da ninguna orden de `kitlegal`: no es de la suite, sino de
  `casosDePresupuestos` (FR-008) y sus mutantes.

### Key Entities

- **Cuerpo de un `SKILL.md`**: lo que sigue a la línea que cierra el frontmatter, con la tabla de comandos generada y
  las marcas; se cuenta en runas (`docs/SKILLS.md` §3).
- **Región de los paquetes**: lo que hay entre `<!-- paquetes: lo genera make skills-sync -->` y
  `<!-- fin de los paquetes -->`, una línea por paquete con referencia en la skill, que escribe la sincronía; hoy, en
  las tres skills de capacidad.
- **Tabla de comandos**: la región generada desde `--describe` en el cuerpo de las skills con `kitlegal-applets`, una
  tabla por applet, una fila por verbo.
- **Prototipo**: `skills/` de `f89093dd` (rama `cuerpos-prototipo`), la fuente byte a byte de los textos.

### Quién consume cada salida, cuánto y hasta cuándo

El hito no añade ninguna salida del binario: ningún verbo, ningún campo de `data`, ningún hallazgo (FR-019). Lo que
cambia lo consumen tres lectores:

- **El cuerpo de cada skill de capacidad, con su tabla de tres columnas**: lo carga el agente una vez por pregunta en
  que se activa la skill, antes de su primera llamada; usa cada fila para elegir la orden o la herramienta de cada
  operación (US1). Tamaño: como mucho 9600 caracteres sin las líneas de los paquetes (FR-005) y 11 000 con ellas; según
  el hito, 9498, 9577 y 9473 con el prototipo. No depende del uso: con meses de uso, cientos de normas y miles de
  bloques consultados, el cuerpo es el mismo, porque no lleva nada de la caché ni del grafo; solo crece con los
  paquetes, una línea de 52 + 3n caracteres cada uno, acotados a 1400 por FR-005. Lo que deja de dar la tabla (las
  claves de `data`, el sobre, las banderas) el agente lo recibe en el resultado de cada llamada, que es donde lo usa.
- **Las tres referencias que ganan texto** (`municipios.md`, `consulta_preparada.md`, `documento_aportado.md`): las
  lee el agente sin condición, una vez por pregunta antes de la primera orden de `plazos-y-territorio`, una vez antes
  de `cita_preparar`, y una vez antes de cada cotejo de un documento aportado (las órdenes de leer de hoy, ADR 0051 y
  0053). Ganan lo que pierden los cuerpos, los mismos textos con lo mínimo cambiado para leerse fuera del cuerpo (ADR
  0063, punto 3); no dependen del uso ni crecen con él.
- **El defecto de FR-005**: lo lee quien cambia un cuerpo o escribe un paquete, persona o run, una vez por `make ci`
  o `make skills-check`; una línea de unos 90 bytes por skill que se pasa, tres como mucho. No depende del uso. Deja de
  darse cuando el cuerpo vuelve a 9600 o menos sin las líneas de los paquetes; con los textos del hito no se da.

## Success Criteria *(mandatory)*

### Measurable Outcomes

- **SC-001**: `make ci` en verde en la cabeza de la rama, con el guion `h45-paquete-de-mas` y su paquete de más.
  Control: `make ci`, que sale en rojo si cualquiera de sus tests o guiones falla.
- **SC-002**: `make skills-check` sin defectos y sin deriva, con el cuerpo de `normativa`, `plazos-y-territorio` y
  `jurisprudencia` en 9600 caracteres o menos sin las líneas de los paquetes y en 11 000 o menos con ellas. Control: el
  defecto de FR-005 y el de 11 000 de `skills-check`, que sale en rojo dentro de `make ci` si un cuerpo pasa de uno u
  otro tope, y la deriva de `skills-check`, que sale en rojo si `make skills-sync` cambiaría un fichero.
- **SC-003**: El defecto de FR-005, en rojo con su mutante: con la regla quitada, o con el tope cambiado, el subtest
  `presupuestos` sale en rojo; en el límite, 9600 en verde y 9601 en rojo. Control: `casosDePresupuestos` en `make
  ci`, y el registro de mutantes del hito con la orden, lo esperado y lo visto de cada mutante.
- **SC-004**: `TestSkillsDelProductoDesdeControl` en verde con cada cambio declarado, y en rojo con cualquiera de esos
  cambios deshecho en `skills/`. Control: el test, en `make ci`, y el registro de mutantes.
- **SC-005**: `TestRemisionesDelRepositorio` en verde con las frases nuevas, y en rojo con la de `main` en cualquiera
  de las dos. Control: el test, en `make ci`, y el registro de mutantes.
- **SC-006**: La suite de aceptación de FR-021 en verde sobre las skills instaladas, y en rojo sobre las de `main`.
  Control: la suite, dentro de `make ci`.
- **SC-007**: Ningún `kitlegal` que devuelva otra cosa: el diff de la rama contra `main` no cambia ningún test ni
  guion fuera de los que FR-019 nombra como excepción, y todos salen en verde. Control: esos tests en `make ci`, que
  salen en rojo si cambia el `--describe`, el sobre, un código de salida o la forma de `data` de un applet.
- **SC-008**: El cierre de evals de todas las skills, sin reutilizar nada de `main`: cero clases que deciden por debajo
  de su umbral y cero series de `claude-sonnet-5-5` por debajo de 2 de 3 en una eval que decide. Control: el job de
  evals del cierre, que sale en rojo si una clase que decide queda por debajo de su umbral o una serie que decide, por
  debajo de 2 de 3.

## Criterios del hito (literales)

- **Controles** (del hito): «`make skills-check` sin defectos y sin deriva, con los cuerpos bajo los 9600 sin las
  líneas de los paquetes; el defecto nuevo, en rojo con su mutante; `TestSkillsDelProductoDesdeControl` con cada cambio
  declarado; `TestRemisionesDelRepositorio` con las frases nuevas; ningún `kitlegal` que devuelva otra cosa (los tests
  de cada applet sin tocar).» En este spec: SC-002, SC-003, SC-004, SC-005 y SC-007; FR-004, FR-005, FR-007, FR-008,
  FR-014, FR-015 y FR-019.
- **Aceptación** (del hito):
  - «`make ci` en verde, con el guion `h45-paquete-de-mas` y su paquete de más»: SC-001; FR-008.
  - «el cierre de evals (todas las skills, sin reutilizar nada de `main`): ninguna clase que decide por debajo de su
    umbral y ninguna serie de `claude-sonnet-5-5` por debajo de 2 de 3 en una eval que decide. Del criterio de la
    medida previa solo rigen esas dos condiciones: la muestra leída y el segundo despacho son de la medida, no del
    cierre»: SC-008; US5.
  - «la persona comprueba antes de lanzar H42 que su prototipo, rebasado sobre `main` con este hito, da `make ci` en
    verde sin ningún `.go`»: lo hace la persona, fuera de este run (ver «Fuera de alcance»).

## Assumptions

- **«`CHANGELOG.md`, con la versión de cada skill que cambia»**: ninguna skill lleva un número de versión propio en el
  repositorio; se lee como una entrada bajo *Unreleased* que nombra cada skill que cambia y lo que cambia en ella, como
  la de H41.1 (FR-018). La versión del binario la pone la persona al publicar (ADR 0020).
- **El orden de los defectos**: cuando se dan el de 11 000 y el de FR-005, el nuevo va detrás, como los de hoy van en
  el orden en que se comprueban (líneas, caracteres).
- **«El cuerpo sin lo que hay entre ellas»**: se quitan las líneas entre las dos marcas, no las marcas, que cuentan
  como en el presupuesto del cuerpo (ADR 0063, punto 1: «con la tabla de comandos generada y las marcas»).
- **Los mutantes** de SC-003 a SC-005 los elige el plan y los registra una tarea del run en el registro de mutantes
  del hito, con su orden; no los corre `make ci`.
- **La copia congelada de H41** (`specs/034-h41-las-skills-del/aceptacion/skills-instaladas.txtar`) ya difiere hoy de
  su guion y el hito no la nombra: no se toca.
- **Los cambios de Go y de tests del prototipo** son una guía: el plan los rehace con estos requisitos y sus mutantes.

## Dependencias

- H45.1 (fusionado en `main`): los tests del molde sin el estado de hoy, y el guion `h45-paquete-de-mas`.
- H42 depende de este hito.

## Fuera de alcance

- La forma de la línea de un paquete y de las marcas.
- Una referencia índice o cualquier referencia nueva.
- Mover una línea `⚠`, una forma fija, una regla invariante o una orden de leer.
- Cambiar una descripción.
- El tope de 11 000, el de 900 y el de 500 líneas.
- `--describe`, el esquema de salida de las herramientas y lo que devuelve cualquier applet.
- Las evals, sus umbrales y el job.
- Los datos de `data/`.
- `consulta-juridica`, que no tiene tabla de comandos y cuyas regiones de la base crecen con cada paquete sin reserva
  propia (ADR 0063, «En contra»).
- El directorio `laboratorio/`.
- La medida previa del ADR 0063 (punto 4), sobre `cuerpos-prototipo`: es de antes de fusionar la entrada, no del run.
- Comprobar que el prototipo de H42, rebasado sobre `main` con este hito, da `make ci` en verde sin ningún `.go`: lo
  hace la persona antes de lanzar H42.
- Lo de «Antes de lanzar» del hito (`df -h /tmp`, el sondeo del alias, mover el fichero): lo hace la persona.
- La copia congelada de H41 (`specs/034-h41-las-skills-del/aceptacion/skills-instaladas.txtar`).
- Cambiar cualquier ADR o escribir otro.
