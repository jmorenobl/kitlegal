# Feature Specification: H33 · Cobertura en positivo por capacidad y materia

**Feature Branch**: `039-h33-cobertura-en-positivo`

**Created**: 2026-11-20

**Status**: Draft

**Input**: el fichero del hito, `docs/roadmap/in-progress/H33-cobertura-en-positivo-por-capacidad.md` (entrada de la
persona, ADR 0018), que aplica el ADR 0058 (`docs/ADR/0058-un-solo-no-cubierto-con-plantilla-y-lo-que-precisa.md`).

## Contexto

Hoy `data/cobertura.yaml` solo lista lo que la LPAC deja fuera del procedimiento común (`regulados_aparte`, nueve
filas). Una pregunta civil, de seguros o laboral no cae en ninguna fila, y `consulta-juridica` dice en prosa lo que no
cubre (§4, «Lo que no cubre, en prosa»), sin forma que un programa compruebe. En la bitácora (`docs/USO.md`,
2026-10-21), kitlegal leyó y citó bien la Ley 50/1980 en la pregunta de la póliza, pero nada dijo a la familia que los
plazos, a quién reclamar y qué vía seguir no estaban cubiertos. H33 declara en positivo lo que kitlegal cubre, por
capacidad y materia, y hace que todo lo demás se declare con una sola línea, la misma plantilla en todas las materias
(ADR 0044, punto 4; ADR 0058; principio II).

Comprobado en esta sesión, en el repositorio (base de la rama `7a7526e8`, `main`):

- `data/cobertura.yaml` tiene 9 filas con `procedimiento:` y ninguna lista `cubierto` ni clave
  `motivo_de_lo_no_declarado`;
- el cuerpo de `skills/consulta-juridica/SKILL.md` mide 3810 caracteres y el de `skills/plazos-y-territorio/SKILL.md`,
  10 882 (`wc -m` de lo que sigue al frontmatter);
- `consulta-juridica` §3 lleva cuatro viñetas escritas a mano (Normativa, Jurisprudencia, Plazos, Territorio), la de
  los paquetes con su región generada (`<!-- paquetes de la base: … -->`) y la de lo que no tiene herramienta; §4 se
  titula «Lo que no cubre, en prosa»;
- `plazos-y-territorio` lleva las dos formas `⚠ NO CUBIERTO:` de hoy en un bloque `text` de «Cómo se presenta un
  plazo», y la tabla `cobertura_por_materia` de `references/administrativo.md` tiene 9 filas con la forma vieja
  («… no por el procedimiento común: kitlegal no calcula sus plazos ni sus recursos (…)»);
- en `data/normas.yaml`, `BOE-A-2015-11430` (ET) y `BOE-A-1980-22501` (Ley 50/1980) no llevan `vertical`;
- las evals con `no_cubierto` son hoy la 07, la 08 y la 12 de `plazos-y-territorio`; `consulta-juridica` tiene 2 evals
  (01 la póliza, con `lecturas` del art. 4; 02 las vacaciones del ET); ninguna eval lleva `tambien_la_resuelven`;
- el comentario del tope de `.github/workflows/evals.yml` (línea 267) dice «8485 s en consulta-juridica … con sus dos
  evals»;
- la rama `h33-prototipo` (local y en `origin`) es el prototipo de la persona; no se fusiona y no es base de nada.
  `TestActivadaPor` y `TestSkillsNombradasEnTambienLaResuelven` no existen en `main`.

La medida previa con modelo («Antes de lanzar» del hito) la hizo la persona; el run no la repite. Su resultado, según
el hito y no medido en esta sesión: el despacho 44595560885, sobre `27413343`, con los cuatro trabajos en verde y
`cobertura_sin_declarar` de las dos skills y `sin_activar` de `consulta-juridica` en 0 en los dos modos.

## Clarifications

Ninguna pregunta en esta sesión: el fichero del hito y el ADR 0058 fijan los textos, los datos, el contrato del
generador, las evals, los controles y la aceptación. Una corrección del gate del spec:

- Q: El hito («El tope» de `plazos-y-territorio`) admite que, «si la implementación necesita más texto», el cuerpo
  crezca hasta dejar 140 caracteres libres y que lo que sobre vaya a `administrativo.md` o a `calculo_de_plazos.md`;
  ¿entra esa excepción, que choca con FR-005, FR-009, FR-020 y «Fuera de alcance»? → A: No. El cuerpo que entrega H33
  es exactamente el de FR-009 (10 686), sin texto añadido ni movido; `calculo_de_plazos.md` no cambia y
  `administrativo.md`, solo en lo que fija FR-005; 10 860 queda como la cota del guion de FR-025. Alternativa
  descartada: abrir esa excepción en FR-005, FR-009, FR-020 y «Fuera de alcance», con un `SKILL.md` y una referencia
  distintos de los textos medidos antes de lanzar y una declaración escrita por el mismo run (auto: conservadora;
  impacto: skill; fuente: hito H33, «El tope»; constitución, «Criterio de decisión autónoma», puntos 2 y 4).

## User Scenarios & Testing *(mandatory)*

### User Story 1 - Quien pregunta por una materia sin paquete sabe qué no está cubierto (Priority: P1)

Una persona pregunta por su póliza o por su despido (seguros, laboral: materias sin paquete). `consulta-juridica` lee y
cita la norma con la skill de normativa, que es lo cubierto, y declara con una línea `⚠ NO CUBIERTO:` lo que la
pregunta pide y nada en kitlegal resuelve —el plazo de esa materia, a quién reclamar, qué vía seguir, un escrito—, con
el motivo «fuera de lo que kitlegal cubre», sin dar de memoria plazos, órganos ni vías y sin pedir
`kitlegal plazos calcular`.

**Why this priority**: es el objetivo del hito y la evidencia de la bitácora: la familia de la póliza no supo que eso no
estaba cubierto.

**Independent Test**: las evals 01 y 04 de `consulta-juridica`, con `no_cubierto: {fuera_de_lo_declarado: true}`,
juzgadas sin modelo por `cobertura_sin_declarar` sobre respuestas escritas como constantes en los tests.

**Acceptance Scenarios**:

1. **Dado** la eval 01 (la póliza, «¿Qué dice la ley y qué podemos hacer?»), **Cuando** la respuesta cita el art. 22
   de la Ley 50/1980 y lleva una línea que empieza por `⚠ NO CUBIERTO:`, **Entonces** no cuenta en
   `cobertura_sin_declarar`.
2. **Dado** la eval 04 (el despido del 2 de noviembre de 2026), **Cuando** la respuesta no lleva ninguna línea
   `⚠ NO CUBIERTO:`, **Entonces** cuenta en `cobertura_sin_declarar` con el motivo
   `declaración de no cubierto ausente: fuera de lo declarado`.
3. **Dado** una eval fuera de lo declarado, **Cuando** la línea es
   `⚠ NO CUBIERTO: el día en que vence el plazo: es un plazo laboral. kitlegal no da sus plazos, órganos, vías ni escritos.`
   o `⚠ NO CUBIERTO: a quién reclamar y qué escrito presentar. Está fuera de lo que kitlegal cubre, que no da plazos, órganos, vías ni escritos para esto.`,
   **Entonces** no cuenta como sin declarar: la cola y el motivo no se comprueban palabra por palabra.
4. **Dado** la eval 04, **Cuando** la sesión pide `plazos calcular`, **Entonces** la eval cae por un comando prohibido.
5. **Dado** una eval fuera de lo declarado, **Cuando** la respuesta solo lleva una línea `⚠ PLAZO:`, **Entonces** cuenta
   como sin declarar.

---

### User Story 2 - Lo cubierto se declara en datos y llega generado a la skill de entrada (Priority: P1)

Quien mantiene kitlegal o trae un paquete de materia declara lo cubierto en `data/cobertura.yaml` (`cubierto`, una fila
por capacidad y materia) y `make skills-sync` lo escribe en la región de lo cubierto de `consulta-juridica`; el agente
lee esa lista en §3 en cada sesión y decide con ella la frontera.

**Why this priority**: sin la lista en datos, la frontera vive en prosa escrita a mano y un paquete no puede declarar lo
que cubre sin tocar el protocolo (ADR 0058, punto 4; principio X).

**Independent Test**: `make skills-sync` sobre el repositorio da diff vacío; la vertical de ensayo de
`TestVerticalDeEnsayo` añade su fila y aparece en la región.

**Acceptance Scenarios**:

1. **Dado** `data/cobertura.yaml` con las cuatro filas de `cubierto` del hito, **Cuando** se ejecuta `make skills-sync`,
   **Entonces** la región entre `<!-- lo que cubre la base: lo genera make skills-sync -->` y
   `<!-- fin de lo que cubre la base -->` tiene 4 líneas `- **<Capacidad>**: <alcance>`, en el orden del fichero, y el
   repositorio no cambia.
2. **Dado** la vertical de ensayo con su fila `{capacidad: plazos, alcance: los de ensayo del molde de los paquetes.,
   materia: ensayo}`, **Cuando** se sincroniza, **Entonces** la región de `consulta-juridica` gana la línea
   `- **Plazos**: los de ensayo del molde de los paquetes.` sin un `.go` nuevo ni un cambio en un `SKILL.md` fuera de
   las marcas.
3. **Dado** una fila de `cubierto` con `materia: civil` y ninguna norma de `data/normas.yaml` con `vertical: civil`,
   **Cuando** se sincroniza, **Entonces** la sincronía da su defecto y `make ci` sale en rojo.
4. **Dado** una fila de `cubierto` con `normas:` o con `capacidad: documentos`, **Cuando** se valida contra
   `schemas/cobertura.yaml.json`, **Entonces** el esquema la rechaza con su defecto.

---

### User Story 3 - Lo regulado aparte usa la misma plantilla, con su motivo (Priority: P1)

Quien pregunta por un recurso contra una liquidación tributaria o una multa de tráfico recibe de
`plazos-y-territorio` la línea de su fila, que ahora es la plantilla única con el motivo de lo regulado aparte, y nada
más tras ella sobre la norma que lo rige (ADR 0043, punto 7).

**Why this priority**: una sola plantilla en todas las materias es la decisión del ADR 0058 (punto 1), y las 9 filas de
`administrativo.md` salen de ella.

**Independent Test**: `TestRegionDeLaCobertura` y el guion `h32-paquetes.txtar`, con la fila de los tributos byte a
byte.

**Acceptance Scenarios**:

1. **Dado** la fila de la aplicación de los tributos, **Cuando** se genera la tabla, **Entonces** su última columna es,
   byte a byte, `⚠ NO CUBIERTO: aplicación de los tributos en materia tributaria y aduanera y su revisión en vía administrativa: se rige por Ley 58/2003, General Tributaria (BOE-A-2003-23186), no por el procedimiento común (DA 1ª.2.a LPAC [BOE-A-2015-10565, bloque daprimera]). kitlegal no da sus plazos, órganos, vías ni escritos.`
2. **Dado** la eval 07 (lo regulado aparte con norma), **Cuando** la línea `⚠ NO CUBIERTO:` no lleva `BOE-A-2003-23186`
   o no lleva `[BOE-A-2015-10565, bloque daprimera]`, **Entonces** cuenta como sin declarar con el motivo
   `declaración de no cubierto ausente: ` y la norma esperada.
3. **Dado** una eval de lo regulado aparte, **Cuando** la etiqueta va en una línea y la cita en otra, **Entonces**
   cuenta como sin declarar.
4. **Dado** una eval de lo regulado aparte, **Cuando** la respuesta lleva la línea con el motivo fijo «fuera de lo que
   kitlegal cubre» y sin la norma y la cita de su precepto, **Entonces** cuenta como sin declarar.

---

### User Story 4 - Las evals miden la frontera sin modelo y el cierre la hace cumplir (Priority: P1)

El job de evals mide la frontera en dos materias: una eval a cada lado en seguros y en laboral. Los umbrales
`cobertura_sin_declarar` y `sin_activar` de `consulta-juridica` deciden el cierre con 0 de tope. La eval 17 de
`normativa` cuenta como resuelta también si la resuelve `plazos-y-territorio` (`tambien_la_resuelven`).

**Why this priority**: sin evals con forma, la plantilla no se protege; con tope 0, un rojo es un defecto que ve la
persona (ADR 0038, punto 1; ADR 0058, puntos 6 y 7).

**Independent Test**: `TestEvalsDelRepositorio`, `TestUmbralesDeLasSkills`, `TestActivadaPor`,
`TestSkillsNombradasEnTambienLaResuelven` y `TestDefinicionDelJob` en `make ci`.

**Acceptance Scenarios**:

1. **Dado** el conjunto de la rama, **Cuando** se comprueba la regla «frontera en dos materias», **Entonces** se cumple;
   **Dado** un mutante sin la eval 03, **Cuando** se comprueba, **Entonces** da su defecto.
2. **Dado** la eval 17 de `normativa` con `tambien_la_resuelven: [plazos-y-territorio]`, **Cuando** una sesión activa
   `plazos-y-territorio` y no `normativa`, **Entonces** la eval cuenta como activada y no suma en `sin_activar`.
3. **Dado** una eval con `tambien_la_resuelven: [no-existe]`, **Cuando** se comprueba el conjunto, **Entonces** la regla
   «skills nombradas» da su defecto.
4. **Dado** una eval con `no_cubierto` que lleva a la vez `fuera_de_lo_declarado: true` y `norma`, **Cuando** se valida
   contra `schemas/eval.yaml.json`, **Entonces** el esquema la rechaza; **Dado** una eval `{norma, precepto}` mal
   escrita, **Cuando** se valida, **Entonces** los mensajes son los de hoy (`TestFormatoDePlazos` sin cambios).
5. **Dado** los umbrales de `consulta-juridica`, **Cuando** se calculan, **Entonces** llevan `sin_activar` y
   `cobertura_sin_declarar` del modelo que decide en cada modo, los cuatro con `Decide: true`, delante de sus cuatro
   `fuera_de_lo_grabado`.

---

### User Story 5 - Lo instalado lleva la plantilla y nada más cambia (Priority: P2)

Quien instala kitlegal con `kitlegal skills install` recibe `consulta-juridica` y `plazos-y-territorio` con la misma
plantilla, la marca solo en `consulta-juridica`, la región de lo cubierto con cuatro líneas y la tabla de
`administrativo.md` generada desde la plantilla; las suites congeladas siguen en verde con los cuatro cambios
declarados.

**Why this priority**: es la aceptación e2e del hito; protege lo que llega al equipo de quien lo usa.

**Independent Test**: un guion `testscript` sobre `kitlegal skills install`.

**Acceptance Scenarios**:

1. **Dado** un directorio vacío, **Cuando** se ejecuta `kitlegal skills install`, **Entonces** el `SKILL.md` instalado
   de `consulta-juridica` y el de `plazos-y-territorio` llevan la misma línea de plantilla, y solo el de
   `consulta-juridica` lleva la línea `⚠ NO VERIFICADO:`.
2. **Dado** lo instalado, **Cuando** se lee la región de lo cubierto de `consulta-juridica`, **Entonces** tiene 4
   líneas, una por fila de `cubierto`.
3. **Dado** lo instalado, **Cuando** se lee la tabla `cobertura_por_materia` de `administrativo.md`, **Entonces** cada
   una de sus 9 filas es la plantilla con el motivo de su caso, y la de los tributos es la del escenario 1 de la
   historia 3, byte a byte.
4. **Dado** lo instalado, **Cuando** se lee la línea de la plantilla de cada `SKILL.md`, **Entonces** no contiene
   «procedimiento común».

### Edge Cases

- **Sin `data/cobertura.yaml`, o con defectos**: la región de lo cubierto y la tabla quedan como están, y el defecto es
  el del fichero, como hoy en la región de los paquetes (FR-007). La comprobación del motivo fijo en la prosa da el
  mismo defecto del fichero (FR-003).
- **Dos filas de una capacidad** (la de lo administrativo y la de ensayo): dos líneas en la región, en el orden del
  fichero (FR-007).
- **Una skill con cero o con dos líneas de plantilla**: no tiene plantilla; si lleva la región, defecto (FR-006).
- **Una pregunta que no pide lo no cubierto** (eval 02, eval 03): ninguna línea se exige; una línea de más no la mide
  ningún umbral (límite conocido; «Fuera de alcance»).
- **Una sesión de la 03 o la 04 que lee fuera de lo grabado**: `fuera_de_lo_grabado` lo da y no decide; las `lecturas`
  las declara la persona después (FR-016).
- **Lo que llega de fuera manipulado a mano** (un `data/cobertura.yaml` o una eval ilegibles, una referencia truncada):
  lo cubre la regla genérica, defecto `inesperado` con código 1 (ADR 0023), sin casos propios.

## Requirements *(mandatory)*

### Functional Requirements

**Los datos de la cobertura**

- **FR-001**: `data/cobertura.yaml`, validado contra `schemas/cobertura.yaml.json` y leído por `skills.LeerCobertura`,
  tiene dos listas y un texto:
  - `motivo_de_lo_no_declarado: fuera de lo que kitlegal cubre`, la única fuente del motivo de la línea fuera de lo
    declarado;
  - `cubierto`, con estas cuatro filas, en este orden, con estos textos y ninguna más:
    1. `capacidad: normativa`, `alcance: leer y citar cualquier norma consolidada del BOE, estatal o autonómica, de
       cualquier materia y en cualquier fecha.`;
    2. `capacidad: jurisprudencia`, `alcance: preparar la consulta exacta de una sentencia, que la persona busca con su
       navegador, y cotejar el documento que trae. No consulta ningún buscador ni dice si una sentencia existe.`;
    3. `capacidad: plazos`, `alcance: solo los del procedimiento administrativo común, los del acceso a la información
       pública y la interposición del recurso contencioso-administrativo, con los festivos de los municipios que
       cuentan. Lo que la ley regula aparte de ese procedimiento lo declara `plazos-y-territorio` con su línea.`,
       `materia: administrativo`;
    4. `capacidad: territorio`, `alcance: la provincia, la comunidad, el régimen, el DIR3 del ayuntamiento y los
       boletines de un municipio.`, sin materia (ADR 0058, punto 5);
  - `regulados_aparte`, con sus 9 filas de hoy sin cambiar un byte de sus datos.

  Una fila de `cubierto` lleva `capacidad` (una de `normativa`, `jurisprudencia`, `plazos`, `territorio`) y `alcance`;
  `materia`, si la lleva, es el `vertical` de un paquete de `data/normas.yaml`; la fila sin `materia` sirve a todas las
  materias. No lleva `normas`. Control: `TestLeerCobertura` en `make ci`, que lee el fichero del repositorio y sale en
  rojo si las filas o el motivo difieren.
- **FR-002**: El esquema de la cobertura rechaza, cada caso con un defecto que nombra el campo:
  - una fila de `cubierto` sin `capacidad` o sin `alcance`;
  - una `capacidad` fuera del vocabulario de FR-001;
  - una clave que no declara, como `normas`;
  - un fichero sin `motivo_de_lo_no_declarado` o con él vacío.

  Lo que rechaza hoy en `regulados_aparte` lo sigue rechazando igual. Control: un mutante por caso en
  `TestLeerCobertura` (`internal/skills/cobertura_test.go`, en `make ci`), en rojo si un caso se acepta.
- **FR-003**: La sincronía de las skills (`internal/skills`, en `make skills-sync` y en `make ci`) da su defecto:
  - si una `materia` de `cubierto` no es el `vertical` de ninguna norma de `data/normas.yaml`;
  - si una skill con plantilla (FR-004) no dice en su prosa el motivo fijo tal cual, entre «», tras juntar los blancos
    seguidos, saltos de línea incluidos;
  - sin `data/cobertura.yaml`, con el defecto del fichero, como en la región.

  Control: un mutante por caso en `internal/skills`, en rojo si la sincronía no da el defecto.

**La línea**

- **FR-004**: La plantilla única es, en una línea de un bloque `text`:

  ```text
  ⚠ NO CUBIERTO: <asunto>: <motivo>. kitlegal no da sus plazos, órganos, vías ni escritos.
  ```

  y la marca de lo leído, que solo lleva `consulta-juridica`, en la línea siguiente del mismo bloque:

  ```text
  ⚠ NO VERIFICADO: <lo encontrado>: <forma legible> [<identificador>, bloque <id>]; kitlegal no ha comprobado que se aplique a este asunto.
  ```

  La plantilla es la misma línea, byte a byte, en `consulta-juridica` y en `plazos-y-territorio`, y no contiene
  «procedimiento común». Controles: el subtest `formas-compartidas` de `TestSkillsDelRepositorio`
  (`internal/app/skills_test.go`), en rojo si las dos líneas difieren; `forma-de-cita` y `ordenes-de-ejemplo` siguen en verde; el guion de aceptación de FR-025.
- **FR-005**: El `<motivo>`, según el caso:
  - fuera de lo declarado: `fuera de lo que kitlegal cubre`, el de `data/cobertura.yaml`;
  - lo regulado aparte con norma: `se rige por <norma> (<identificador de la norma>), no por el procedimiento común
    (<precepto> [<identificador>, bloque <id>])`;
  - lo regulado aparte sin norma: `lo regula aparte la ley, no el procedimiento común (<precepto> [<identificador>,
    bloque <id>])`.

  Los dos motivos de lo regulado aparte van en `references/administrativo.md`, en «Lo que la ley regula aparte»,
  delante de la región, tras la frase «Los motivos de la línea; el primero es el de la fila con norma:», en un bloque
  `text`, cada uno en una línea, el de la fila con norma primero. Lo demás escrito a mano de `administrativo.md` no
  cambia. Control: el subtest de `plazos-y-territorio` de `TestSkillsDelProductoDesdeControl` con la declaración de
  FR-020.
- **FR-006**: El generador de la región `cobertura_por_materia` (`internal/skills/regiones.go`) cumple este contrato:
  - la plantilla es la única línea de un bloque `text` del `SKILL.md` que empieza por `⚠ NO CUBIERTO: ` y lleva
    `<motivo>`; con otra cantidad, la skill no tiene plantilla;
  - los motivos son las líneas de los bloques `text` de la referencia del paquete que tiene la región que llevan
    `<precepto>` y no empiezan por `⚠`: una con `<norma>` y otra sin ella;
  - cada fila es la plantilla con `<asunto>` en lugar del procedimiento de la fila y `<motivo>` en lugar del motivo de
    su caso, y los marcadores de hoy sustituidos por sus datos; ningún texto de la línea vive en Go;
  - los marcadores conocidos son los de hoy más `<asunto>` y `<motivo>`;
  - defectos, cada uno con el prefijo `<skill>: references/<ref>.md: ` del fichero de su referencia:
    - región en una skill sin plantilla: `región cobertura_por_materia en una skill sin la plantilla ⚠ NO CUBIERTO:
      en SKILL.md`;
    - región en una referencia sin sus dos motivos, uno de cada: `región cobertura_por_materia en una referencia sin
      sus dos motivos`;
    - un marcador desconocido, en la plantilla o en un motivo: el mensaje nombra el marcador y termina en
      `un marcador que la tabla no conoce`, como hoy;
  - una fila de `data/cobertura.yaml` que ninguna región genera sigue dando su defecto de hoy en cada skill con
    plantilla, que son `consulta-juridica` y `plazos-y-territorio`.

  Los mensajes de `docs/SKILLS.md` pasan a decir «la plantilla» donde dicen «las dos formas». Controles:
  `TestRegiones` y `TestRegionDeLaCobertura`, con la fila de los tributos byte a byte, y un mutante por defecto, en
  rojo si el generador no da la fila o el defecto esperados.
- **FR-007**: La región de lo cubierto, nueva, va en el `SKILL.md` que lleva sus marcas, hoy solo `consulta-juridica`:
  - marcas, cada una en su línea: `<!-- lo que cubre la base: lo genera make skills-sync -->` y
    `<!-- fin de lo que cubre la base -->`;
  - entre ellas, una línea por fila de `cubierto`, en el orden del fichero: `- **<Capacidad>**: <alcance>`, con la
    primera letra de la capacidad en mayúscula y sin escribir `materia`; dos filas de una capacidad dan dos líneas;
  - sin `data/cobertura.yaml`, o con defectos, la región queda como está y el defecto es el del fichero;
  - `partirElCuerpo` (`internal/app/skills_test.go`) no parte sus líneas, como no parte las de la tabla de comandos.

  Con las filas de FR-001, la región repite las cuatro viñetas de hoy, cada una en una sola línea, y solo cambia la de
  los plazos, que nombra también los del acceso a la información pública (`acceso-resolver`, `acceso-reclamar` de
  `data/plazos.yaml`). Control: `TestSkillsDelRepositorio` en `make ci`, en rojo si `make skills-sync` cambia algún
  fichero.

**Las skills**

- **FR-008**: En `consulta-juridica`:
  - el párrafo de entrada, la viñeta «Paquetes de materia» de §3 con su región, la de «Sin herramienta, no cubiertos»
    y el frontmatter no cambian;
  - las cuatro viñetas de las capacidades de §3 pasan a ser la región de lo cubierto (FR-007);
  - §4 pasa a ser, palabra por palabra:

    > ### 4. Lo que no cubre, con su línea
    >
    > Lo que la pregunta pide y no resuelve ninguna capacidad de la lista del 3 —el plazo de una materia que no es la
    > del procedimiento administrativo común, a quién reclamar, qué vía seguir, un escrito— lleva su línea, con el
    > motivo «fuera de lo que kitlegal cubre». Lo que no pide no lleva ninguna, tampoco el cómputo de una fecha que
    > no ha pedido. Esta línea no es para lo que la ley regula aparte del procedimiento común: eso lo declara, con su
    > propia línea, la skill de la capacidad que lo pide, y tras ella no leas su norma. Lo que leas con la skill de
    > normativa para completarla va detrás, citado, con su marca: el texto citado es el de la norma; lo que no está
    > verificado es que se aplique a este asunto. No des de memoria ningún plazo, órgano ni vía, ni apliques los de
    > otra materia o los del procedimiento común, ni pidas `kitlegal plazos calcular` para un asunto de otra materia:
    > ese plazo no es el suyo.

    seguido de la plantilla y la marca en un bloque `text` (FR-004) y del párrafo «Por qué» de hoy, sin cambios;
  - el cuerpo pasa de 3810 a 4481 caracteres.

  Control: `TestConsultaJuridicaDesdeLaBase` con la huella de FR-020, en rojo si el `SKILL.md`, con sus dos regiones
  reducidas a sus marcas, difiere en un byte.
- **FR-009**: En `plazos-y-territorio`, tres cambios del cuerpo y nada más:
  - paso 7: tras «**ni `plazos calcular`** ni ninguna fecha (regla 9)» se añade «; si es de otra materia, igual, con
    el motivo de lo no declarado»;
  - «Cómo se presenta un plazo»: las dos formas de hoy y su párrafo pasan a ser la plantilla, sola en su bloque `text` y
    sin la marca, detrás del párrafo «Lo no cubierto (paso 7 y regla 9) va con esta forma: en un asunto de la cobertura
    del paquete, la línea de su fila, entera; en otra materia, con el motivo «fuera de lo que kitlegal cubre»:»;
  - regla 9: su última frase pasa a ser «Un plazo, un órgano o una vía de otra materia tampoco: su línea, sin darlos de
    memoria ni pedir `plazos calcular`.».

  Las órdenes de leer (ADR 0052 y 0053), el frontmatter y lo demás del cuerpo no cambian. El cuerpo baja de 10 882 a
  10 686 caracteres. Control: el subtest de `plazos-y-territorio` de `TestSkillsDelProductoDesdeControl` (FR-020), en
  rojo con el primer fichero que difiere de su base con la declaración.
- **FR-010**: Tope del cuerpo. Cada `SKILL.md` de `skills/` tiene un cuerpo de como mucho 11 000 caracteres (ADR 0046),
  también con la línea de la vertical de ensayo; con ella, el de `plazos-y-territorio` mide 10 756. El cuerpo de
  `plazos-y-territorio` que entrega H33 es exactamente el de FR-009 (10 686), sin texto añadido ni movido a una
  referencia, y no se acorta otro texto del cuerpo; 10 860 (140 caracteres libres sin la vertical de ensayo) es la cota
  que hace cumplir el guion de FR-025 (ver «Clarifications»). Controles:
  `TestSkillsDelRepositorio` y `TestVerticalDeEnsayo` en `make ci`, en rojo si un cuerpo pasa de 11 000; el guion de
  aceptación de FR-025, en rojo si el cuerpo instalado de `plazos-y-territorio` pasa de 10 860.
- **FR-011**: El paquete de materia incluye lo cubierto (ADR 0058, punto 4): trae sus filas de `cubierto` y, si trae
  región de cobertura, sus dos motivos en su referencia. La vertical de ensayo de `TestVerticalDeEnsayo`, en una tarea
  `[datos]` porque es `testdata/`:
  - gana en su `data/cobertura.yaml` la fila `{capacidad: plazos, alcance: los de ensayo del molde de los paquetes.,
    materia: ensayo}`, delante de `regulados_aparte`;
  - gana en su `references/ensayo.md` de `plazos-y-territorio` la frase y los dos motivos de FR-005.

  El test cambia en tres sitios: `anadirEntradas` copia las dos listas, cada una bajo su clave; `partirPorLasMarcas`
  conoce las marcas de lo cubierto; `comprobarLoSincronizado` comprueba que `consulta-juridica` lista la fila de la
  vertical. La vertical sigue entrando sin un `.go` ni un cambio en un `SKILL.md` fuera de las marcas. Control:
  `TestVerticalDeEnsayo` en `make ci`.

**Las evals**

- **FR-012**: El formato de eval (`schemas/eval.yaml.json`, en una tarea `[datos]`) admite en `no_cubierto`, además de
  `{norma, precepto}`, la variante `{fuera_de_lo_declarado: true}`, nunca las dos a la vez, con un `if`/`then` sobre
  `fuera_de_lo_declarado` y no con `oneOf`. Una eval con cualquiera de las dos es de cobertura. Los mensajes que da hoy a
  una eval `{norma, precepto}` mal escrita no cambian. Controles: los tests del formato de `internal/evals`, con un
  mutante que lleva las dos variantes a la vez (en rojo si se acepta), y `TestFormatoDePlazos` sin cambios.
- **FR-013**: El juicio de la declaración (`internal/evals`):
  - la respuesta declarada lleva una línea `⚠ NO CUBIERTO:`, encontrada por su etiqueta, con la tolerancia de hoy
    (énfasis, `⚠️`);
  - en lo regulado aparte, la misma línea lleva el identificador de la norma esperada y la cita de su precepto, como
    hoy; fuera de lo declarado basta la línea;
  - la línea `⚠ PLAZO:` no cuenta como declaración;
  - solo se comprueban literales la etiqueta, el identificador y la cita, nunca la cola ni el motivo; ningún texto de la
    línea vive en `internal/evals`;
  - el motivo de una respuesta sin declarar es `declaración de no cubierto ausente: ` seguido de la norma esperada, o
    de `fuera de lo declarado` en una eval `fuera_de_lo_declarado`.

  Control: mutantes de respuestas escritas como constantes en los tests de `make ci`. Cuentan como sin declarar, cada
  uno con su motivo: sin la línea, en una eval fuera de lo declarado y en una de lo regulado aparte; en lo regulado
  aparte, sin la norma o sin la cita; la línea partida en dos; la línea `⚠ PLAZO:`. No cuentan como sin declarar,
  fuera de lo declarado: la línea con negritas, con `⚠️`, y las dos líneas literales del despacho 44568601898 (historia
  1, escenario 3). El test sale en rojo si un mutante se clasifica al revés.
- **FR-014**: Los umbrales (`internal/evals/umbrales.go`) se calculan en toda skill con alguna eval de cobertura:
  - `cobertura_sin_declarar:<modelo>:<modo>` en `consulta-juridica` (evals 01 y 04) y `plazos-y-territorio` (07, 08 y
    12), con 0 de tope (ADR 0038, punto 1). Su descripción, una constante que no copia ningún texto de la línea, es:
    «Respuestas de %s en el modo %s sin una línea ⚠ NO CUBIERTO: —en lo regulado aparte, con la norma que lo rige y la
    cita de su precepto—, o con una línea ⚠ PLAZO:, sobre sus respuestas medidas en las evals que esperan una
    declaración de no cubierto»;
  - `sin_activar:<modelo>:<modo>` nace en `consulta-juridica`, con 0 de tope, sobre las respuestas de sus cuatro evals,
    con el código de las demás skills, sin caso especial (ADR 0058, punto 6).

  `⚠ NO VERIFICADO:` no entra en ningún umbral. Controles: `TestUmbralesDeLasSkills` en `make ci`, en rojo si
  `consulta-juridica` no lleva `sin_activar` y `cobertura_sin_declarar` del modelo que decide en cada modo, los cuatro
  con `Decide: true` y delante de sus cuatro `fuera_de_lo_grabado`; en el cierre, el job de evals sale en rojo si
  cualquiera de esos umbrales pasa de 0.
- **FR-015**: `tambien_la_resuelven`, campo nuevo del formato de eval (`schemas/eval.yaml.json`, en una tarea `[datos]`,
  e `internal/evals/formato.go`): lista de nombres de otras skills del producto cuya activación cuenta como la de la
  skill de la carpeta, para `Activada`, para el juicio de la eval y para `sin_activar`, cuyo tope sigue en 0 (ADR 0058,
  punto 7). Cada nombre es una skill de `skills/`, con la regla «skills nombradas» de hoy. Lo lleva solo la eval 17 de
  `normativa`, con `[plazos-y-territorio]` y un comentario que dice por qué; sus comandos y sus citas no cambian.
  Controles: `TestActivadaPor` y `TestSkillsNombradasEnTambienLaResuelven` en `make ci` (los de la rama
  `h33-prototipo`), en rojo si la activación de `plazos-y-territorio` no cuenta en la 17 o si un nombre que no es una
  skill de `skills/` se acepta.
- **FR-016**: Las evals de `consulta-juridica` son cuatro, una a cada lado de la frontera en dos materias sin paquete:
  - **01** (la póliza): gana `no_cubierto: {fuera_de_lo_declarado: true}`; su comentario deja de decir que lo no
    cubierto se dice en prosa; lo demás no cambia;
  - **02** (vacaciones, art. 38 ET): sin cambios;
  - **03**, nueva: pregunta «¿Qué dice la Ley de Contrato de Seguro sobre la prórroga de la póliza?», comando y cita de
    `BOE-A-1980-22501`, bloque `aveintidos`, sin fecha; `se_activan: [normativa]`;
    `no_se_activan: [jurisprudencia, plazos-y-territorio]`;
  - **04**, nueva: pregunta «Me despidieron el 2 de noviembre de 2026. ¿Qué dice el Estatuto de los Trabajadores del
    plazo para reclamar y hasta qué día puedo hacerlo?»; comando `{applet: boe, norma: BOE-A-2015-11430, bloque: a59,
    fecha: "2026-11-02"}`; cita de `a59` sin `redaccion_fechada`; `prohibidos: [plazos calcular]`;
    `no_cubierto: {fuera_de_lo_declarado: true}`; `no_se_activan: [jurisprudencia]`, y ninguna otra restricción de
    activación.

  Ni la 03 ni la 04 declaran `lecturas` (ADR 0043, punto 2; ADR 0032). La 07, la 08 y la 12 de `plazos-y-territorio`
  no cambian de fichero. Control: `TestEvalsDelRepositorio` en `make ci` (subtest «formato» con las dos evals nuevas, y
  el mutante de «no activación de cada capacidad», que quita `plazos-y-territorio` de `no_se_activan` en la 02 y en la
  03).
- **FR-017**: Regla nueva del conjunto de `consulta-juridica`, «frontera en dos materias»: hay al menos dos normas de
  `data/normas.yaml` sin `vertical` que tienen cada una una eval activa que la cita y no es de cobertura, y otra activa
  que la cita, es de cobertura con `fuera_de_lo_declarado` y prohíbe `plazos calcular`. La cumplen las cuatro evals de
  FR-016 (Ley 50/1980 y ET). Control: `TestEvalsDelRepositorio` en `make ci`, en rojo con el conjunto que no la cumple;
  un mutante sin la eval de un lado da su defecto.
- **FR-018**: Lo grabado: el art. 59 del ET (`BOE-A-2015-11430`, `a59`) en la fecha 2026-11-02 lo graba una tarea
  `[datos]` con el paso `grabar_datos` (sin modelo, `KITLEGAL_RECORD=1`); el art. 22 de la Ley 50/1980 ya está grabado.
  Control: `TestCacheDeLasSesiones` en `make ci`, en rojo si falta una consulta de las evals de la skill.
- **FR-019**: El trabajo de evals: `consulta-juridica` pasa de 2 a 4 evals y sigue a concurrencia 1; su peor caso pasa
  de 8485 s a 15 013 s, dentro del tope de 352 minutos (21 120 s). La concurrencia de las demás skills no cambia. En
  `.github/workflows/evals.yml` solo cambia el comentario del tope (la cifra y «sus dos evals»), no sus valores.
  Controles: `TestDefinicionDelJob` en `make ci`, en rojo si el peor caso de un trabajo pasa de su tope;
  `TestMatrizDeLasPropuestas` con la matriz de cada ejecución.

**Suites congeladas, tests, documentación y cierre**

- **FR-020**: Las suites congeladas siguen en verde con estos cuatro cambios y ninguno más; la tarea `[datos]` que
  cambia los ficheros de los guiones de H32 de `internal/app/testdata/script/` es la misma que cambia la skill; las
  copias de `specs/030-h32-*/aceptacion/` no cambian:
  - `h32-paquetes.txtar`: su `tributos.txt` pasa a ser la línea de la historia 3, escenario 1, en su fila de la tabla;
    lo demás no cambia (las 33 filas de `c991808` siguen con sus datos en la referencia);
  - `h32-capacidades.txtar`: en `formas.txt`, la fila `plazos-y-territorio|⚠ NO CUBIERTO:|2` pasa a `|1`, y el
    comentario de cabecera dice una, la plantilla; el total sigue en `formas: 16 de 16`;
  - `TestSkillsDelProductoDesdeControl`, subtest de `plazos-y-territorio`: la declaración gana los tres cambios de
    FR-009 y el cambio declarado en lo escrito a mano de `administrativo.md` (FR-005); la declaración se amplía para
    admitir un cambio declarado en una referencia, y la comparación quita también las regiones generadas de las
    referencias escritas a mano; los subtests de `normativa` y `jurisprudencia` no cambian;
  - `TestConsultaJuridicaDesdeLaBase`: la huella sha256 pasa a ser la del `SKILL.md` que entrega H33 con sus dos
    regiones generadas reducidas a sus marcas; los mensajes dicen «de la entrega de H33» donde hoy dicen «de la base de
    la rama (ab951c22)»; sus mutantes siguen dando su defecto exacto.

  Control: `make ci` y `go test ./...` en verde en la cabeza de la rama.
- **FR-021**: Los tests de unidad que el prototipo puso en rojo se actualizan con los datos y la forma nuevos, con
  `go test ./...` y `-tags=integration` en `internal/{evals,skills,app,workflow}` y `data`, y son estos:
  `TestLeerCobertura`, `TestRegiones`, `TestRegionDeLaCobertura`, `TestVerticalDeEnsayo`, `TestSkillsDelRepositorio`,
  `TestJuzgarPlazos`, `TestJuzgarLasLecturasSinServir`, `TestUmbralesDePlazosYTerritorio`, los de `informe_test.go`,
  `TestUmbralesDeLasSkills`, `TestDefinicionDelJob`, `TestMatrizDeLasPropuestas`, `TestCacheDeLasSesiones` y los dos
  subtests de `TestEvalsDelRepositorio` de FR-016. Los de juicio de lo regulado aparte cambian solo sus constantes con
  la línea vieja, que pasan a la plantilla; lo que comprueban no cambia. `TestFormatoDePlazos` no cambia. Control:
  `make ci` en verde.
- **FR-022**: Documentación:
  - `docs/SKILLS.md`: la plantilla, los motivos en la referencia del paquete, la región de lo cubierto, los mensajes de
    la tabla de defectos de la región con «la plantilla» y el de los motivos, y la fila de `consulta-juridica` de la
    tabla de las skills sin «en prosa» ni «sin… líneas `⚠`»;
  - `CONTRIBUTING.md`: la fila de `no_cubierto` con `fuera_de_lo_declarado`, la fila nueva de `tambien_la_resuelven`,
    la fila de `cobertura_sin_declarar` con la descripción de FR-014 y los umbrales de `consulta-juridica` con los dos
    que nacen;
  - el comentario del tope de `.github/workflows/evals.yml` (FR-019);
  - `docs/HISTORIAL.md`: la viñeta del hito;
  - `docs/USO.md`: una entrada de la bitácora con lo que enseñó la medida previa: un criterio de sondeo o de umbral es
    un defecto que vería quien usa kitlegal, no la ruta entre skills, y los `⚠ NO CUBIERTO:` falsos los cazó leer las
    respuestas, no un umbral;
  - `CHANGELOG.md`: «[Unreleased]» → «Cambiado»;
  - `web/src/pages/como-se-evalua/index.astro`: la frase que anuncia H33 pasa a decir lo hecho.

  El run no escribe ningún ADR: el 0058 ya está. Se verifica en la revisión y en el quickstart; no lleva umbral
  numérico.
- **FR-023**: Lo generado se regenera con `make skills-sync` y queda sin diff. Control: `TestSkillsDelRepositorio`.
- **FR-024**: En la medición del cierre, las evals de las cuatro skills del producto salen en verde en los dos modos,
  con `cobertura_sin_declarar` en 0 en `consulta-juridica` y en `plazos-y-territorio` y `sin_activar` en 0 en
  `consulta-juridica`. Control: el job de evals sale en rojo si cae una eval o uno de esos umbrales pasa de 0. Si caen,
  la reparación no toca ningún texto de las skills ni de `administrativo.md`, ni las evals ni los umbrales, para que
  pasen: lo deja dicho para la persona.
- **FR-025**: Aceptación e2e, un guion `testscript` sobre `kitlegal skills install`, en `make ci`, que sale en rojo si:
  - el `SKILL.md` instalado de `consulta-juridica` o el de `plazos-y-territorio` no lleva la plantilla, o las dos
    líneas difieren, o `plazos-y-territorio` lleva la marca, o `consulta-juridica` no la lleva;
  - la región de lo cubierto de `consulta-juridica` no tiene exactamente una línea por cada una de las cuatro filas de
    `cubierto`;
  - una fila de `cobertura_por_materia` de `administrativo.md` no es la plantilla con el motivo de su caso, o la de los
    tributos difiere en un byte de la de la historia 3, escenario 1;
  - la línea de la plantilla de un `SKILL.md` contiene «procedimiento común»;
  - el cuerpo instalado de `plazos-y-territorio` pasa de 10 860 caracteres (FR-010).
- **FR-026**: El run no lanza ningún sondeo ni abre sesiones con modelo (ADR 0032).

### Key Entities

- **Fila de lo cubierto**: `{capacidad, alcance, materia?}` de `cubierto` en `data/cobertura.yaml`; `capacidad` del
  vocabulario de FR-001; `materia`, un `vertical` de `data/normas.yaml`, o ninguna (todas las materias).
- **Fila de lo regulado aparte**: la de hoy de `regulados_aparte` (procedimiento, precepto, norma si se cita de lo
  grabado).
- **Motivo de lo no declarado**: el texto `fuera de lo que kitlegal cubre` de `motivo_de_lo_no_declarado`.
- **Plantilla**: la única línea `⚠ NO CUBIERTO: … <motivo> …` de un bloque `text` de un `SKILL.md`.
- **Marca de lo leído**: la línea `⚠ NO VERIFICADO:` de `consulta-juridica`.
- **Motivos de lo regulado aparte**: las dos líneas del bloque `text` de la referencia del paquete (con y sin
  `<norma>`).
- **Región de lo cubierto**: las líneas entre las marcas de FR-007 de un `SKILL.md`.
- **Eval de cobertura**: la que lleva `no_cubierto`, `{norma, precepto}` o `{fuera_de_lo_declarado: true}`.

### Uso, de fuera adentro

- **La región de lo cubierto** la lee el agente al cargar `consulta-juridica`, una vez por sesión, en el paso 3, y con
  ella decide en §4 si lo que pide la pregunta lleva la línea. Tamaño: una línea por fila de `cubierto`, hoy 4 líneas
  de unos 120 a 330 caracteres (las cuatro viñetas de hoy miden 723 caracteres; con el acceso a la información, unos
  770); crece una línea por paquete que traiga una fila, no con el uso: con
  meses de uso diario, cientos de normas y miles de bloques consultados, ni la caché ni el grafo entran en ella. Lo
  que gana la skill: la frontera en datos, igual para toda materia. No es una señal que caduque: cambia cuando cambian
  los datos (una release o un paquete nuevo).
- **La línea `⚠ NO CUBIERTO:` de una respuesta** la escribe el agente y la lee la persona: como mucho una por asunto
  no cubierto que la pregunta pide, ninguna por lo que no pide (§4, segunda frase); una frase de una línea. Se deja de
  dar en una pregunta cuando lo que pide entra en lo cubierto: un paquete que añade su fila (FR-011) o una capacidad
  nueva de otro hito. No se repite en una pregunta que no lo pide.
- **La marca `⚠ NO VERIFICADO:`** la escribe el agente de `consulta-juridica` solo tras su línea fuera de lo declarado,
  una por lectura hecha con la skill de normativa para completar; ninguna en lo regulado aparte.
- **La tabla `cobertura_por_materia`** la lee el agente de `plazos-y-territorio` en el paso 7, una vez por pregunta de
  plazo de un asunto de su paquete, y copia una fila: 9 filas (3128 bytes entre las nueve; antes de H33, 3047), que crecen
  con las filas de `regulados_aparte` de los paquetes, no con el uso.
- **Los defectos de la sincronía** (FR-002, FR-003, FR-006) los lee quien mantiene kitlegal cuando `make skills-sync` o
  `make ci` salen en rojo: una línea por defecto, con la skill y el fichero.
- **Los umbrales y sus descripciones** (FR-014) los lee la persona en el informe del job, una vez por medición: dos
  umbrales más por modo, del modelo que decide (cuatro), en `consulta-juridica`.

## Success Criteria *(mandatory)*

### Measurable Outcomes

- **SC-001**: 4 de 4 filas de `cubierto` salen en la región de lo cubierto de `consulta-juridica`, una línea cada una,
  y `make skills-sync` deja 0 ficheros cambiados. Control: `TestSkillsDelRepositorio` y el guion de FR-025, en rojo si
  falta o sobra una línea o hay diff.
- **SC-002**: 9 de 9 filas de `cobertura_por_materia` son la plantilla con el motivo de su caso, y la de los tributos
  difiere en 0 bytes de la del hito. Control: `TestRegionDeLaCobertura`, `h32-paquetes.txtar` y el guion de FR-025, en
  rojo con un byte de diferencia.
- **SC-003**: la plantilla es la misma línea en 2 de 2 skills con plantilla, contiene 0 veces «procedimiento común», y
  la marca está en 1 skill (`consulta-juridica`). Control: el subtest `formas-compartidas` de
  `TestSkillsDelRepositorio` y el guion de FR-025, en rojo si una de las tres cuentas difiere.
- **SC-004**: en la medición del cierre, `cobertura_sin_declarar` es 0 en `consulta-juridica` y en
  `plazos-y-territorio`, y `sin_activar` es 0 en `consulta-juridica`, en los dos modos. Control: el job de evals sale
  en rojo si alguno pasa de 0 (FR-014, FR-024).
- **SC-005**: el cuerpo de cada `SKILL.md` mide como mucho 11 000 caracteres, también con la línea de la vertical de
  ensayo (`plazos-y-territorio`: 10 686 sin ella y 10 756 con ella; `consulta-juridica`: 4481), y el de
  `plazos-y-territorio` entregado, como mucho 10 860. Control: `TestSkillsDelRepositorio`, `TestVerticalDeEnsayo` y el
  guion de FR-025, en rojo si una medida pasa de su tope.
- **SC-006**: el peor caso del trabajo de `consulta-juridica` es de 15 013 s, como mucho 21 120 s. Control:
  `TestDefinicionDelJob`, en rojo si pasa del tope.
- **SC-007**: al menos 2 normas sin `vertical` cumplen la regla «frontera en dos materias». Control:
  `TestEvalsDelRepositorio`, en rojo con menos de 2.
- **SC-008**: cada mutante de FR-002, FR-003, FR-006, FR-012, FR-013 y FR-017 da su defecto o su clasificación (0
  mutantes aceptados). Control: los tests de `make ci` que los llevan.
- **SC-009**: `make ci` y `go test ./...` (y `-tags=integration` en `internal/{evals,skills,app,workflow}` y `data`) en
  verde en la cabeza de la rama, con 4 suites congeladas cambiadas, las de FR-020, y ninguna más.
- **SC-010**: el run lanza 0 sondeos y abre 0 sesiones con modelo. Control: el sandbox de los pasos del run, sin
  credenciales (ADR 0032).

## Criterios del hito (literales)

**Controles** (del fichero del hito):

- `make ci` comprueba, sin red ni modelo:
  - **El esquema de la cobertura** rechaza, cada caso con su defecto:
    - una fila de `cubierto` sin `capacidad` o `alcance`, con una capacidad fuera del vocabulario, o con una clave
      que no declara, como `normas`;
    - un fichero sin `motivo_de_lo_no_declarado` o con él vacío.

    La sincronía rechaza una `materia` que no es el `vertical` de ninguna norma de `data/normas.yaml`, y una skill con
    plantilla cuya prosa no dice el motivo fijo. Con `regulados_aparte`, igual que hoy.
  - **Las regiones.** La de lo cubierto de `consulta-juridica` sale de `cubierto` con diff vacío. La tabla de
    `administrativo.md` sale de la plantilla y de los motivos de la referencia, entera y en una línea por fila. Cada
    defecto del contrato del generador tiene un mutante que lo da.
  - **`formas-compartidas`**: la plantilla es la misma línea en `consulta-juridica` y en `plazos-y-territorio`.
    `forma-de-cita` y `ordenes-de-ejemplo` siguen en verde (comprobado en el prototipo).
  - **`cobertura_sin_declarar`, con mutantes** de respuestas escritas como constantes en los tests. Cuentan como sin
    declarar, cada una con su motivo:
    - una respuesta sin la línea, en una eval fuera de lo declarado y en una de lo regulado aparte;
    - en lo regulado aparte, sin la norma o sin la cita;
    - la línea partida en dos, con la etiqueta en una y la cita en otra;
    - la línea con `⚠ PLAZO:`.

    No cuentan como sin declarar, fuera de lo declarado, la línea con negritas o con `⚠️`, ni la línea con la cola
    o el motivo dichos con otras palabras. De estas, las dos del despacho 44568601898, literales:

    ```text
    ⚠ NO CUBIERTO: a quién reclamar y qué escrito presentar. Está fuera de lo que kitlegal cubre, que no da plazos, órganos, vías ni escritos para esto.
    ⚠ NO CUBIERTO: el día en que vence el plazo: es un plazo laboral. kitlegal no da sus plazos, órganos, vías ni escritos.
    ```
  - **La regla «frontera en dos materias»**: el conjunto de la rama la cumple, y un mutante sin la eval de un lado da
    su defecto.
  - **La vertical de ensayo** entra con su fila de `cubierto` y sus motivos, y su fila sale en la región de lo cubierto
    de `consulta-juridica`.
  - **El tope** del cuerpo de cada skill, también con la línea de la vertical de ensayo (`TestVerticalDeEnsayo`), y el
    peor caso del trabajo de evals bajo su tope (`TestDefinicionDelJob`).
- **Tests de unidad que el prototipo puso en rojo**, con `go test ./...` y con `-tags=integration` en
  `internal/{evals,skills,app,workflow}` y `data`. La lista es completa. Se actualizan con los datos y la forma nuevos;
  lo que cambia lo que comprueban va dicho:
  - en `internal/skills`:
    - `TestLeerCobertura`, `TestRegiones` y `TestRegionDeLaCobertura`, que lleva la fila de los tributos byte a byte;
    - `TestVerticalDeEnsayo`, con los tres cambios de la vertical de arriba;
  - en `internal/app`, `TestSkillsDelRepositorio` (sus presupuestos de líneas de `consulta-juridica`): `partirElCuerpo`
    no parte la región de lo cubierto;
  - en `internal/evals`:
    - `TestJuzgarPlazos`, `TestJuzgarLasLecturasSinServir`, `TestUmbralesDePlazosYTerritorio` y las de
      `informe_test.go` siguen en verde, porque en lo regulado aparte la comprobación es la de hoy; sus constantes con
      la línea vieja pasan a la plantilla;
    - `TestUmbralesDeLasSkills`, que cambia lo que comprueba. Su comentario fija que `consulta-juridica` va «sin
      `sin_activar`». Pasa a llevar `sin_activar` y `cobertura_sin_declarar` del modelo que decide en cada modo, los
      cuatro con `Decide: true` y delante de sus cuatro `fuera_de_lo_grabado` (ADR 0058, punto 6);
    - con los datos de las evals nuevas: `TestDefinicionDelJob` (el peor caso de `consulta-juridica`, de 8485 s a
      15 013 s), `TestMatrizDeLasPropuestas` (la matriz de cada ejecución) y `TestCacheDeLasSesiones` (el art. 59, que
      se graba);
    - en `conjunto_test.go` (`TestEvalsDelRepositorio`), dos subtests: «formato», con las dos evals de la entrega de
      `consulta-juridica`, y el mutante de «no activación de cada capacidad». La eval 03 también lleva
      `plazos-y-territorio` en `no_se_activan`, y el mutante tiene que quitarlo de las dos.

    `TestFormatoDePlazos` no tiene que cambiar (ver el `if`/`then` de arriba).
- **Las suites congeladas siguen en verde, con cuatro cambios declarados aquí.** Son las únicas que se pusieron en rojo
  en el prototipo. Los dos guiones de H32 están en `internal/app/testdata/script/`: la tarea `[datos]` que cambia sus
  ficheros es la misma que cambia la skill, para que `make ci` quede en verde en cada tarea. Sus copias de
  `specs/030-h32-*/aceptacion/` son el registro de H32, no se ejecutan y no cambian.
  - **`h32-paquetes.txtar`** (de H32). Su `tributos.txt`, la fila de la aplicación de los tributos byte a byte, pasa a
    ser la línea de arriba en su fila de la tabla. Lo demás del guion no cambia: las 33 filas de `c991808` siguen con
    sus datos en la referencia.
  - **`h32-capacidades.txtar`** (de H32). En su `formas.txt`, la fila `plazos-y-territorio|⚠ NO CUBIERTO:|2` pasa a
    `|1`, y su comentario de cabecera, que dice «dos en ⚠ NO CUBIERTO:», pasa a decir una, la plantilla. El total
    sigue en `formas: 16 de 16`: no mira `consulta-juridica`, y `plazos-y-territorio` no lleva la marca.
  - **`TestSkillsDelProductoDesdeControl`**, el subtest de `plazos-y-territorio` (de H41, H41.1 y H41.4):
    - la declaración de `plazos-y-territorio` gana los tres cambios del cuerpo de arriba;
    - y gana un cambio en lo escrito a mano de `administrativo.md`, los motivos. Hoy la declaración solo admite cambios
      del `SKILL.md` y añadidos al final de una referencia, así que se amplía para admitir un cambio declarado en una
      referencia;
    - la comparación quita también las regiones generadas de las referencias escritas a mano, como ya hace con las
      del `SKILL.md`.

    Es lo que hizo H41.1 con el subtest de H41. Los de `normativa` y `jurisprudencia` no cambian.
  - **`TestConsultaJuridicaDesdeLaBase`** (de H41.4). Guarda la huella sha256 del `SKILL.md` de `consulta-juridica` en
    `ab951c22`, y no hay en el repositorio texto de esa base sobre el que declarar cambios.
    - La huella pasa a ser la del `SKILL.md` que entrega H33, con sus dos regiones generadas reducidas a sus marcas,
      para que regenerar la lista de lo cubierto o la de los paquetes no la mueva.
    - El nombre de la base en los mensajes pasa a ser «la entrega de H33»: donde hoy dicen «de la base de la rama
      (ab951c22)», dicen «de la entrega de H33».
    - Sus mutantes siguen dando su defecto exacto.

    Lo que cambia del `SKILL.md` lo ve la persona en el diff de la propuesta.

Las evals salen en verde en la medición del cierre, en los dos modos, con `cobertura_sin_declarar` en 0 en
`consulta-juridica` y en `plazos-y-territorio` y `sin_activar` en 0 en `consulta-juridica`. Las primeras medidas son
las de «Antes de lanzar». Si en el cierre caen, la reparación no toca para que pasen ningún texto de las skills ni de
`administrativo.md`, ni las evals ni los umbrales: lo deja dicho para la persona.

Se cubren con FR-002 a FR-006, FR-010, FR-011, FR-013, FR-014, FR-016, FR-017, FR-019 a FR-021 y FR-024, y SC-001 a
SC-009.

**Aceptación** (del fichero del hito):

- **e2e** (`testscript`) sobre `kitlegal skills install`:
  - el `SKILL.md` instalado de `consulta-juridica` y el de `plazos-y-territorio` llevan la plantilla, la misma línea
    en los dos, y solo el de `consulta-juridica` lleva la marca;
  - la región de lo cubierto de `consulta-juridica` tiene una línea por cada una de las cuatro filas de `cubierto`;
  - cada fila de `cobertura_por_materia` de `administrativo.md` es la plantilla con el motivo de su caso, con la de los
    tributos byte a byte;
  - y la línea de la plantilla de cada `SKILL.md` no contiene «procedimiento común»: eso solo lo dicen los motivos de
    la referencia.
- **Las evals** de las cuatro skills, en verde en el cierre.
- El run no lanza ningún sondeo.

Se cubren con FR-025 (y la historia 5), FR-024 (SC-004) y FR-026 (SC-010).

## Assumptions

- **La base de la rama es `7a7526e8`** (`main` al crear la rama en esta sesión). La rama `h33-prototipo` es material de
  consulta de la persona; los textos que fija este spec son los del hito, y donde coinciden con el prototipo es porque
  el hito los copia de él.
- **Los 140 caracteres libres** del cuerpo de `plazos-y-territorio` (auto; impacto: skill) se miden sin la línea de la
  vertical de ensayo, sobre el cuerpo que entrega H33, que es exactamente el de FR-009 (10 686, frente a la cota de
  10 860; ver «Clarifications»): es la lectura con la que el hito cuenta «desde ahí caben tres más de 70». El control
  vive en el guion de aceptación de H33 (FR-025), que un hito posterior que añada un paquete cambia como cambio
  declarado, como hace este con los de H32.
- **El mensaje del marcador desconocido en un motivo** no lo fija el hito palabra por palabra; FR-006 fija su prefijo,
  que nombra el marcador y su final, `un marcador que la tabla no conoce`; el resto lo fija el plan.
- **Las cifras de la medida previa** (despachos, sondeos, tasas) son las del hito y no se han medido en esta sesión.
- **El job de evals y su modelo no cambian** (ADR 0031): `MODELO_DE_EVALS` y `VERSION_DE_CLAUDE_CODE` son los de
  `main`.

## Fuera de alcance

- Cubrir algo que hoy no se cubre, o dejar de cubrir algo: ninguna fila de `cubierto` más allá de las cuatro de FR-001,
  ninguna capacidad nueva (destinatarios y documentos son H35 y H37) y ningún cambio de texto en sus alcances salvo el
  de los plazos, que nombra el acceso a la información.
- Una clase del juez que mida que, tras la línea `⚠ NO CUBIERTO:`, la respuesta no da de memoria un plazo, un órgano ni
  una vía (ADR 0037 y 0038); va a la bitácora, para otro hito.
- Cambiar las filas de `regulados_aparte` (sus procedimientos, preceptos y normas) o quitarlas.
- Exponer o marcar con `⚠ NO VERIFICADO:` la norma que rige lo regulado aparte (ADR 0043, punto 7; ADR 0058, punto 2).
- Cambiar `skills/normativa` o `skills/jurisprudencia`, o una `description` de cualquier skill (ADR 0055 y 0057); de
  sus evals, solo la 17 de `normativa` cambia, y solo para ganar `tambien_la_resuelven`.
- Cambiar en `consulta-juridica`, `plazos-y-territorio` o `administrativo.md` algo distinto de FR-005 y FR-007 a FR-009.
- Un umbral o una clase del juez para `⚠ NO VERIFICADO:` (ADR 0037).
- Medir dos límites conocidos, que si el uso los muestra van a la bitácora: una línea `⚠ NO CUBIERTO:` de más en una
  eval de dentro de lo cubierto (las 02 y 03 solo comprueban comando y cita), y una fecha contada a mano en la prosa de
  una respuesta que lleva la línea (`juzgarElPlazo` solo falla por la línea ausente o por una `⚠ PLAZO:`).
- Cambiar el tope de `cobertura_sin_declarar` o de `sin_activar`, cómo cuentan los demás, o ningún otro umbral, límite
  ni tope, más allá de los dos que nacen en `consulta-juridica` y de las activaciones que declara
  `tambien_la_resuelven` (ADR 0038, punto 1; ADR 0058, punto 7).
- Cambiar el job de evals más allá de lo que piden sus tests con las evals nuevas y del comentario de su tope: ni
  `VERSION_DE_CLAUDE_CODE`, ni `MODELO_DE_EVALS`, ni lo que instala (ADR 0031).
- Que el binario o un servidor detecte la materia de una pregunta o lea una respuesta: la frontera la decide la skill
  con `cubierto` y la comprueban las evals (principio X).
- Evals de más materias; la aceptación de la fase 1, con cinco o seis materias, no es de este hito.
- Declarar `lecturas` en las evals 03 y 04: las declara la persona después de ver una sesión (ADR 0043, punto 2).
- Escribir o cambiar un ADR: el 0058 ya está en el repositorio.
- Repetir la medida previa con modelo o lanzar sondeos.
