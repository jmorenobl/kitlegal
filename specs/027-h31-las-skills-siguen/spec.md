# Feature Specification: H31 · Las skills siguen la especificación Agent Skills: guía de autoría y gates

**Feature Branch**: `027-h31-las-skills-siguen`

**Created**: 2026-10-24

**Status**: Draft

**Input**: el fichero del hito, `docs/roadmap/in-progress/H31-las-skills-siguen-la-especificacion-agent.md`, la única
entrada humana del run (ADR 0018). El ADR 0046, la constitución 2.15.0 (principio VIII), `CLAUDE.md`, la DoD §1.10 y la
fila de §3 de `docs/ROADMAP.md` y `evidencias/adr-0046/` están en `main` desde la propuesta de la entrada: el run los
cumple, no los decide.

## Contexto

Las tres skills del producto (`boe-legislacion`, `legal-core` y `jurisprudencia`) llevan hoy el protocolo entero en el
cuerpo de `SKILL.md`, porque el principio VIII decía que todo `references/` se genera desde `data/` y el único límite era
el de menos de 300 líneas, que se cumple con líneas largas. Medido en `main` el 2026-10-24, sin el frontmatter, los
cuerpos tienen 283, 283 y 193 líneas y unos 24 400, 19 200 y 13 600 caracteres; el ADR 0046 los midió en unos 10 760,
8520 y 6000 tokens con `claude-sonnet-5-5`, a 2,26 caracteres por token, por encima de los 5000 que recomienda la
especificación Agent Skills. Las tres descripciones hablan al agente en segunda persona («Úsala», «Actívala», «creas
conocer»), y la de `boe-legislacion` tiene 1023 caracteres de los 1024 que admite la especificación.

Lo que gana cada skill: el agente carga al activarla un cuerpo de como mucho 11 000 caracteres (unos 4870 tokens) con lo
que toda respuesta necesita, y lee una referencia solo cuando la clase de pregunta la pide; la descripción, que el agente
lee en todas las sesiones para decidir si activa la skill, está en tercera persona, como pide la guía de autoría de
Anthropic, y una materia nueva (ADR 0044) puede entrar como referencia sin agrandar el cuerpo. Todo ello sin perder
ninguna regla ni empeorar la activación ni las respuestas, que deciden las evals.

## Clarifications

### Session 2026-10-24

- Q: ¿El nombre de otra skill que aparece en una referencia a mano cuenta para el cierre por nombre de la parte (b) del conjunto de dependencias de una eval, igual que si apareciera en `SKILL.md`, o el cierre por nombre sigue leyendo solo `SKILL.md`, como hoy (FR-014)? → A: Opción A. El cierre por nombre de la parte (b) lee `SKILL.md` y sus referencias a mano: una skill nombrada solo en una referencia a mano de una skill de la eval (la regla `nombra` de hoy) entra en (b) con su `SKILL.md` y sus `references/`, y el cierre sigue desde ella. Las referencias generadas no se leen para nombres (salen de `data/`, que entra por (c)). (auto: criterio c; fuente: hito H31, entrega 4 y «Alcance»; ADR 0046, decisión 5; ADR 0045, «En contra»; H30, FR-001(b); `internal/evals/dependencias.go`, `skillsDeLaEval` y `nombra`)
- Q: Cuando la elección del job no incluye el cambio de ninguna descripción —la etiqueta `evals` o la suite completa sin descripciones cambiadas, o un cambio solo de la rúbrica, los casos, el modelo o la versión del juez—, ¿el job juzga o reutiliza igualmente las descripciones, o no corre el juez de las descripciones (FR-030)? → A: Opción B. El juez solo corre cuando la elección incluye el cambio de una descripción; en las demás ejecuciones no abre ninguna sesión del juez ni publica el juicio de ninguna descripción, y su umbral no aparece. «Incluye el cambio de una descripción» se decide por las rutas de la propuesta (la regla `cambiaLaDescripcion` de `eligeLaRuta`), no por el motivo de cada eval. (auto: criterio a; fuente: hito H31, entrega 3; ADR 0046, decisión 4; FR-025 y FR-026; `internal/evals/eleccion.go`)
- Q: ¿Una descripción que queda sin juicio porque un voto no llega a darse tras su repetición hace fallar el job, o se publica sin juicio y sin decidir, como una respuesta sin juzgar del juez de las evals (FR-007 de H24) (FR-032)? → A: Opción A. Una descripción sin juicio hace fallar el job, con un motivo propio que nombra la skill y el motivo del voto, distinguible sin modelo de una descripción marcada; queda «sin juzgar», no «sin marcar», y no deja juicio que reutilizar. (auto: criterio a; fuente: ADR 0037, decisión 2; ADR 0046, decisión 4; H24, FR-007; `internal/evals/informe.go`, `motivoDeLasSinJuzgar`; constitución, «Umbrales que se cumplen»)
- Q: ¿Este hito cambia los generadores de las referencias generadas para que escriban su índice cuando su salida pase de 100 líneas, o solo el gate de FR-009 las pone en rojo y el generador se cambia cuando `data/` llegue a ese tamaño (FR-024)? → A: Opción B; no material: los generadores no cambian en este hito (la referencia generada más larga, `normas.md`, tiene 24 líneas). Si un cambio posterior de `data/` lleva una referencia generada por encima de 100 líneas sin índice, el gate de FR-009 la da como defecto y ese cambio adapta su generador. (auto: criterio e; fuente: constitución, «Gates», umbral de materialidad, y «Criterio de decisión autónoma», punto 2; hito H31, «Entrega»; FR-024; `wc -l skills/*/references/*.md`)
- Q: ¿Una referencia de más de 100 líneas sin ningún encabezado de nivel 2 cumple FR-009, que pide un índice que nombre cada encabezado de nivel 2, o es un defecto (FR-009)? → A: Opción B. Es un defecto: una referencia de más de 100 líneas tiene al menos un encabezado de nivel 2 y abre con el índice que los nombra; con 100 líneas o menos no se exige nada, con o sin encabezados. (auto: criterio a; fuente: hito H31, entrega 2; ADR 0046, contexto y decisión 1; spec, «Uso, de fuera adentro», FR-009 y FR-011)

## User Scenarios & Testing *(mandatory)*

### User Story 1 - La skill carga un cuerpo corto y lee el detalle cuando la pregunta lo pide (Priority: P1)

Una persona pregunta algo que activa `boe-legislacion`, `legal-core` o `jurisprudencia`. El agente carga el cuerpo de
`SKILL.md`, con el protocolo, las reglas invariantes, las formas fijas y la tabla de comandos, y, si la clase de
pregunta lo pide, lee la referencia que el cuerpo le dice que lea en ese caso. La respuesta sigue las mismas reglas que
con la versión de `main`.

**Why this priority**: es el objetivo del hito; sin las tres skills reescritas no hay entrega.

**Independent Test**: el job de evals sobre la cabeza de la propuesta mide las 45 evals en los dos modos y sale en verde
con los umbrales vigentes; `make ci` sale en verde con los gates de US2 sobre las tres skills.

**Acceptance Scenarios**:

1. **Dado** la propuesta de H31 con las tres skills reescritas y sus tres descripciones cambiadas, **Cuando** el job de
   evals corre sobre su cabeza, **Entonces** elige las 45 evals en los dos modos, no reutiliza ninguna, juzga las tres
   descripciones y sale en verde con los umbrales vigentes (ADR 0029, 0038 y 0045), sin ninguna descripción marcada
   (SC-010).
2. **Dado** el `SKILL.md` de cada una de las tres skills reescritas, **Cuando** se le quita el frontmatter y se cuenta el
   cuerpo con la tabla de comandos como la deja `make skills-sync`, **Entonces** tiene menos de 500 líneas y como mucho
   11 000 caracteres, y su descripción como mucho 900 (FR-004, FR-005).
3. **Dado** la tabla de `research.md` con cada regla de la versión de `main` de cada skill, **Cuando** la revisión la
   contrasta con la versión nueva, **Entonces** cada regla tiene su sitio en la nueva (fichero y apartado) y la nueva no
   tiene ninguna regla que no esté en la tabla (FR-021).

---

### User Story 2 - `make ci` hace cumplir la forma de la especificación y de la guía (Priority: P1)

Quien cambia una skill —una persona o una tarea de un run— ejecuta `make ci`. Si la skill se sale de la forma
(frontmatter, presupuestos, `references/`), `skills-check` sale en rojo con un defecto que nombra la skill, el fichero y
la regla, con la medida y el límite.

**Why this priority**: sin los gates, el presupuesto y la forma se perderían en el primer cambio; sustituyen al de menos
de 300 líneas.

**Independent Test**: un mutante por gate sobre una copia de una skill hace salir `skills-check` en rojo; las tres skills
del repositorio lo dejan en verde.

**Acceptance Scenarios**:

1. **Dado** una copia de una skill con `name: claude-legal` o `name: x-anthropic`, **Cuando** corre `skills-check`,
   **Entonces** sale en rojo con el defecto del nombre (FR-002).
2. **Dado** una copia con `<ejemplo>` en la descripción, **Cuando** corre `skills-check`, **Entonces** sale en rojo con
   el defecto de la etiqueta XML; **y dado** «menos de 5000 tokens (< 5000)» en la descripción, **Entonces** no es una
   etiqueta (FR-002, FR-003).
3. **Dado** una copia cuyo cuerpo tiene 11 001 caracteres, con caracteres de varios bytes, **Cuando** corre
   `skills-check`, **Entonces** sale en rojo con un defecto que da los 11 001 caracteres y el máximo de 11 000; con
   11 000 caracteres, y más de 11 000 bytes, sale en verde (FR-004, FR-010).
4. **Dado** una copia cuyo cuerpo tiene 500 líneas, **Cuando** corre `skills-check`, **Entonces** sale en rojo; con 499,
   en verde (FR-004).
5. **Dado** una copia con una descripción de 901 caracteres, **Cuando** corre `skills-check`, **Entonces** sale en rojo;
   con 900, en verde (FR-005).
6. **Dado** una copia con un subdirectorio o un fichero que no es `.md` en `references/`, una referencia que `SKILL.md`
   no enlaza, una referencia enlazada desde otra, o una referencia de 101 líneas sin índice, **Cuando** corre
   `skills-check`, **Entonces** sale en rojo con el defecto de cada una; una de 100 líneas sin índice sale en verde
   (FR-006 a FR-009).

---

### User Story 3 - El job de evals juzga que la descripción está en tercera persona (Priority: P1)

Quien abre una propuesta que cambia la descripción de una skill ve en el job de evals el juicio de cada descripción: si
alguna habla en primera o en segunda persona, el job sale en rojo y el informe da la frase y el motivo.

**Why this priority**: la tercera persona es una cuestión de significado que una lista de palabras no distingue (ADR
0046), y sin juez medido la entrega 3 del hito no se cumple.

**Independent Test**: los tests de `make ci` con votos simulados (sin abrir ninguna sesión) deciden marcada, no marcada
y nulo; el control de identidad compara con `evidencias/adr-0046/`; el job sobre la cabeza de la propuesta juzga las
tres descripciones.

**Acceptance Scenarios**:

1. **Dado** tres votos «si», cada uno con una frase que está en la descripción, **Cuando** el juicio decide, **Entonces**
   la descripción queda marcada con esa frase y su motivo (FR-029).
2. **Dado** un primer y un segundo voto «si» con su frase y un tercero «no», **Cuando** el juicio decide, **Entonces** la
   descripción no queda marcada (FR-029).
3. **Dado** un primer voto «no», **Cuando** el juicio pide los votos por orden, **Entonces** no pide el segundo ni el
   tercero y la descripción no queda marcada (FR-029).
4. **Dado** un voto «si» cuya frase no está en la descripción, **Cuando** el juicio lo comprueba sin modelo, **Entonces**
   el voto es nulo y se pide otra vez, una sola (FR-028).
5. **Dado** un byte distinto en la rúbrica, el esquema, los casos o la medida que usa el job respecto de
   `evidencias/adr-0046/`, **Cuando** corre `make ci`, **Entonces** el control de identidad sale en rojo (FR-025).
6. **Dado** una medida que no corresponde (otra huella de la rúbrica o de los casos, otro modelo del juez u otra versión
   de Claude Code) o que no se cumple (62 de 63 defectos marcados, o 1 de 36 correctos marcado), **Cuando** el job de
   evals va a juzgar descripciones, **Entonces** sale en rojo antes de abrir ninguna sesión (FR-026).
7. **Dado** una descripción marcada en una ejecución que la juzga, **Cuando** el job publica sus umbrales, **Entonces**
   el umbral de la clase sale en rojo, el job falla y el informe da la skill, la frase y el motivo (FR-031).
8. **Dado** una descripción con un voto que no llega a darse (FR-028, FR-032), **Cuando** el job la juzga, **Entonces**
   queda sin juzgar, el job falla con un motivo propio que nombra la skill y el motivo del voto, y no deja juicio que
   reutilizar (FR-032).
9. **Dado** una ejecución cuya elección no incluye el cambio de ninguna descripción (la etiqueta `evals` o la suite
   completa sin descripciones cambiadas, o un cambio solo del juez), **Cuando** el job decide qué corre, **Entonces** no
   abre ninguna sesión del juez de las descripciones ni publica su umbral (FR-030).

---

### User Story 4 - El protocolo se escribe a mano en una referencia y la sincronización no lo toca (Priority: P2)

Quien mantiene una skill escribe el detalle del protocolo en `references/<fichero>.md`, sin cabecera. `make skills-sync`
lo deja como está y sigue regenerando lo que sale de `data/`; el job de evals cuenta la referencia a mano entre las
dependencias de las evals de su skill.

**Why this priority**: sin ella no hay sitio donde poner el detalle que sale del cuerpo; hoy `skills-sync` retira
cualquier fichero de `references/` que no esté declarado.

**Independent Test**: un test de `make ci` sobre una copia con una referencia a mano y una generada alterada; un test del
conjunto de dependencias con una referencia a mano.

**Acceptance Scenarios**:

1. **Dado** una skill con una referencia a mano y una referencia generada con un byte cambiado, **Cuando** corre
   `make skills-sync`, **Entonces** la referencia a mano queda idéntica byte a byte y la generada vuelve a ser lo que
   genera su fila (FR-012, SC-003).
2. **Dado** un fichero de `references/` declarado en `kitlegal-referencias` que no es lo que genera su fila, con o sin
   la cabecera «generado desde», o uno no declarado que abre con ella, **Cuando** corre `skills-check`, **Entonces** sale
   en rojo por deriva, como hoy (FR-012).
3. **Dado** una skill con una referencia a mano, **Cuando** se calcula el conjunto de dependencias de una de sus evals
   (H30), **Entonces** la referencia entra en la parte (b), sin parte (c), sin error; y una propuesta que solo la cambia
   elige las evals de esa skill (FR-014, SC-004).

---

### User Story 5 - Mover un texto a una referencia no lo saca de ningún control (Priority: P2)

Los controles de contenido que hoy leen el texto de `SKILL.md` leen la skill entera: `SKILL.md` y sus referencias a
mano. Una expresión prohibida, una orden sin su forma para PowerShell o una variante mal escrita de una forma fija en
una referencia ponen el control en rojo igual que en `SKILL.md`; y una forma fija que las evals comprueban sin modelo
que sale del cuerpo a una referencia también.

**Why this priority**: la reescritura mueve texto a referencias; sin esto, el traslado apagaría controles sin que nadie
lo viera.

**Independent Test**: un mutante por control que lleva su regla a una referencia a mano de una copia de la skill.

**Acceptance Scenarios**:

1. **Dado** una expresión de `evals/boe-legislacion/expresiones-prohibidas.yaml` o una fecha AAAAMMDD con cifras en la
   prosa de una referencia a mano de `boe-legislacion`, **Cuando** corre `make ci`, **Entonces** el control de la prosa
   sale en rojo (FR-017).
2. **Dado** una orden de lectura o de comprobación sin su forma para PowerShell en una referencia a mano, **Cuando** corre
   `make ci`, **Entonces** el control de las órdenes para PowerShell sale en rojo (FR-018).
3. **Dado** la línea sin consulta, una forma de aviso, de hallazgo o de la redacción fechada llevada del cuerpo de
   `SKILL.md` a una referencia, **Cuando** corre `make ci`, **Entonces** su control sale en rojo (FR-018).
4. **Dado** una orden de ejemplo de una referencia a mano que lee del BOE algo que la caché de una sesión de evals no
   sirve sin red, **Cuando** corre `make ci`, **Entonces** el control de las órdenes de ejemplo sale en rojo (FR-019).

---

### User Story 6 - La guía de autoría dice cómo se escribe una skill (Priority: P3)

Quien escribe o revisa una skill lee `docs/SKILLS.md` y sabe qué va en el cuerpo y qué en una referencia, cómo se escribe
la descripción, qué presupuestos hay y por qué, y la regla de `references/`; `CONTRIBUTING.md` y `CHANGELOG.md` dicen lo
mismo.

**Why this priority**: sin la guía, los gates dicen qué falla pero no cómo arreglarlo; no cambia lo que mide el job.

**Independent Test**: la revisión final contrasta la guía con el ADR 0046 y con los gates.

**Acceptance Scenarios**:

1. **Dado** `docs/SKILLS.md`, **Cuando** se lee, **Entonces** tiene las dos fuentes con su fecha de consulta, el reparto
   entre cuerpo y referencia, la forma de la descripción, los presupuestos con su porqué y la regla de `references/`
   (FR-001).

---

### Edge Cases

- Una skill sin `references/` (`jurisprudencia` en `main`) es válida; tras la reescritura puede tener referencias a mano
  sin ninguna generada (FR-006, FR-012).
- Un fichero con el nombre de una referencia declarada en `kitlegal-referencias` es esa referencia generada aunque no
  abra con la cabecera: su contenido distinto es deriva y `skills-sync` lo rehace, como hoy (FR-012).
- Un enlace a `references/<fichero>.md#apartado` enlaza ese fichero; un texto con forma de enlace dentro de un bloque de
  código no es un enlace (FR-007, FR-008).
- Una referencia generada que pasa de 100 líneas (`normas.md` con cientos de normas en `data/normas.yaml`) también abre
  con su índice: si su generador no lo escribe, `skills-check` sale en rojo (FR-009).
- Un `<` que no abre una etiqueta («< 5000», «a<b») no es una etiqueta XML (FR-003).
- Una descripción con una pregunta citada en primera o segunda persona, o con un verbo que se escribe como un imperativo
  de tú con la skill de sujeto, es correcta: es lo que distingue la medida de `evidencias/adr-0046/` (FR-022).
- Un fichero del kit modificado a mano fuera del binario, truncado o con permisos cambiados lo cubre la regla genérica:
  defecto `inesperado`, código 1 (ADR 0023), sin casos propios.

## Requirements *(mandatory)*

### Functional Requirements

#### La guía de autoría

- **FR-001**: `docs/SKILLS.md` es la guía de autoría del proyecto sobre las dos fuentes, la especificación Agent Skills
  (agentskills.io, «Specification») y la guía de Anthropic («Skill authoring best practices»), con la fecha de consulta
  que registra el ADR 0046 (2026-10-23) y las citas que recoge; el run no tiene red y no las vuelve a consultar. Dice:
  (a) qué va en el cuerpo de `SKILL.md` —el protocolo, las reglas invariantes, las formas fijas que las evals comprueban
  sin modelo (las líneas `⚠`), la tabla de comandos y, por cada referencia, cuándo leerla— y qué en una referencia —el
  detalle que solo pide una clase de pregunta—; (b) la forma de la descripción: en tercera persona, qué hace la skill y
  cuándo se usa, con los dos casos correctos que una lista no distingue (la pregunta citada y el verbo con la skill de
  sujeto); (c) los presupuestos con su porqué: menos de 500 líneas (la especificación), como mucho 11 000 caracteres
  (2,26 caracteres por token, unos 4870 tokens, por debajo de 5000), como mucho 900 en la descripción (el 88 % de 1024),
  contados como runas, y que la razón se vuelve a medir cuando cambia el modelo que decide (ADR 0031); (d) la regla de
  `references/` de FR-006 a FR-009 y FR-012. Criterio: la revisión final contrasta la guía con el ADR 0046 y con las
  cifras de los gates (SC-013).

#### Los gates de `skills-check`

- **FR-002**: La validación de la especificación, en Go y sin red, aplica las reglas de hoy de `ValidarFrontmatter`
  (nombre, descripción con como mucho 1024 caracteres, compatibilidad, claves admitidas, applets y referencias
  declaradas) y, además: ninguna etiqueta XML (FR-003) en `name` ni en `description`, y `name` no contiene `anthropic`
  ni `claude`. Cada incumplimiento es un defecto de la skill y `skills-check` sale en rojo. Criterio: SC-002.
- **FR-003**: Una etiqueta XML es un texto de la forma `<nombre…>`, `</nombre>` o `<nombre/>`, donde `nombre` empieza por
  una letra; un `<` que no abre una así no lo es. Criterio: SC-002.
- **FR-004**: El cuerpo de `SKILL.md` —el fichero sin el frontmatter y con la región de la tabla de comandos como la deja
  `make skills-sync`— tiene menos de 500 líneas y como mucho 11 000 caracteres; si no, es un defecto que dice la medida
  y el límite, y `skills-check` sale en rojo. Sustituye al límite de menos de 300 líneas del fichero entero, que
  desaparece. Criterio: SC-002.
- **FR-005**: La descripción tiene como mucho 900 caracteres; si no, es un defecto con la medida y el límite y
  `skills-check` sale en rojo. Criterio: SC-002.
- **FR-006**: `references/` es plano: solo ficheros regulares `.md`, sin subdirectorios ni ficheros de otra extensión;
  cada entrada que no lo cumple es un defecto y `skills-check` sale en rojo. Criterio: SC-002.
- **FR-007**: Cada referencia, a mano o generada, está enlazada desde `SKILL.md` con un enlace de Markdown, fuera de los
  bloques de código, cuyo destino es su ruta relativa `references/<fichero>.md` (con o sin `#apartado`); la que no lo
  está es un defecto y `skills-check` sale en rojo. Criterio: SC-002.
- **FR-008**: Ninguna referencia enlaza otra referencia de su skill (un enlace de Markdown, fuera de los bloques de
  código, cuyo destino es un fichero de `references/`); el que lo hace es un defecto y `skills-check` sale en rojo.
  Criterio: SC-002.
- **FR-009**: Una referencia, a mano o generada, de más de 100 líneas abre con un índice de sus apartados: una lista,
  antes de su primer encabezado de nivel 2, que nombra cada encabezado de nivel 2 del fichero, en su orden. Sin ella es
  un defecto y `skills-check` sale en rojo. Una de más de 100 líneas sin ningún encabezado de nivel 2 no abre con ningún
  índice y es también un defecto, que nombra la skill, el fichero y la regla; con 100 líneas o menos no se exige nada,
  con o sin encabezados. Los generadores no cambian en este hito: si un cambio posterior de `data/` lleva una
  referencia generada por encima de 100 líneas sin índice, este gate la da como defecto y ese cambio adapta su
  generador. Criterio: SC-002.
- **FR-010**: Los caracteres se cuentan como runas, no como bytes; las líneas, como cuenta hoy el límite de líneas.
  Criterio: SC-002 (el mutante de 11 001 caracteres pasa de 11 000 bytes y el de 11 000 caracteres también).
- **FR-011**: Cada gate de FR-002 a FR-009 sale en rojo con un mutante en `make ci`, y no sale en rojo en el límite
  (499 líneas, 11 000 caracteres, 900 caracteres, 100 líneas sin índice). El defecto nombra la skill, el fichero y la
  regla, y `make skills-sync` lo da también, como da hoy los de la skill. Criterio: SC-002.

#### La regla de `references/`

- **FR-012**: Un fichero de `references/` se clasifica primero por la declaración: (1) el que declara
  `kitlegal-referencias` es la referencia generada, con o sin la cabecera «generado desde» (la línea que hoy abre cada
  referencia generada); si difiere de lo que genera su fila es deriva y `skills-sync` lo rehace; (2) el que no está
  declarado y abre con la cabecera es sobrante: deriva, y `skills-sync` lo retira; (3) solo el `.md` no declarado que no
  abre con la cabecera es una referencia a mano. Lo que sale de `data/` sigue generado, con su cabecera y declarado.
  `make skills-sync` no escribe, no rehace ni retira una referencia a mano, y `skills-check` no la da como deriva.
  Criterio: SC-003.
- **FR-013**: Una referencia a mano se escribe sin cabecera y su contenido es el protocolo de detalle de su skill; no
  repite lo que sale de `data/`, que tiene su referencia generada. Criterio: la revisión final.
- **FR-014**: El conjunto de dependencias de cada eval (H30) cuenta cada referencia a mano de las skills de la eval en la
  parte (b), sin parte (c), y no da error por la referencia sin cabecera «generado desde»; la huella de la eval cambia
  cuando cambia una de ellas, y una propuesta que solo la cambia elige las evals de esa skill y de las que la nombran,
  como con cualquier fichero de (b). El cierre por nombre de (b) lee `SKILL.md` y las referencias a mano de cada skill de
  la eval: una skill nombrada solo en una referencia a mano (el nombre de otro directorio de `skills/` entre límites que
  no son `[a-z0-9-]`, la regla `nombra` de hoy) entra en (b) con su `SKILL.md` y sus `references/`, y el cierre sigue
  desde ella como desde cualquier skill nombrada. Las referencias generadas no se leen para nombres: salen de `data/`,
  que entra por (c), y hoy ninguna nombra una skill. Con las skills de `main` la elección no cambia. Criterio: SC-004
  (que incluye el caso de una skill nombrada solo desde una referencia a mano, que tiene que entrar en el conjunto).
- **FR-015**: Las referencias a mano viajan dentro del binario y `kitlegal skills install` las instala con su skill, como
  cualquier fichero de `references/`: el agente de quien usa el kit las tiene donde el cuerpo las enlaza. Criterio: un
  test de `make ci` sobre lo empotrado comprueba que cada referencia de cada skill del repositorio, a mano o generada,
  está en el binario con sus bytes, y sale en rojo si falta una.

#### Los controles de contenido

- **FR-016**: Cada control de contenido de `make ci` que hoy lee el texto de `SKILL.md` —entre ellos la prosa sin
  expresiones prohibidas ni fechas AAAAMMDD con cifras, la línea sin consulta, las órdenes para PowerShell y las formas
  de la redacción fechada, y los demás que leen ese texto (las formas de aviso y de hallazgo, las normas nombradas, sin
  instrucciones de las evals)— lee la skill entera, `SKILL.md` y sus referencias a mano, en las skills a las que hoy se
  aplica. Criterio: SC-005.
- **FR-017**: Un texto que pone en rojo un control cuando está en `SKILL.md` lo pone en rojo igual en una referencia a
  mano. Criterio: SC-005.
- **FR-018**: Una forma fija que un control exige y las evals comprueban sin modelo (la línea sin consulta, las formas de
  aviso, de hallazgo y de la redacción fechada) está en el cuerpo de `SKILL.md`: llevarla solo a una referencia pone su
  control en rojo, y una variante mal escrita de ella en una referencia a mano también. Cada orden de lectura y de
  comprobación de la skill, en `SKILL.md` o en una referencia a mano, lleva detrás su forma para PowerShell; la que no la
  lleva pone su control en rojo. Criterio: SC-005.
- **FR-019**: Las órdenes de ejemplo de la skill que leen del BOE, en `SKILL.md` y en sus referencias a mano, son de lo
  que la caché de las sesiones de evals sirve sin red (ADR 0043): un control de `make ci` ejecuta cada una con
  `--offline` sobre la caché preparada de una sesión de su skill y sale en rojo si una no sale con 0. Criterio: SC-005.

#### Las tres skills reescritas

- **FR-020**: `boe-legislacion`, `legal-core` y `jurisprudencia` tienen la forma de FR-001: el cuerpo con lo que toda
  respuesta necesita y, por cada referencia, cuándo leerla; el detalle que solo pide una clase de pregunta, en
  referencias a mano. Las tres pasan los gates de FR-002 a FR-009 y los controles de FR-016 a FR-019. Criterio: SC-001,
  SC-009.
- **FR-021**: La reescritura cambia la forma, no las reglas: ninguna regla de la versión de `main` de cada skill
  desaparece y ninguna aparece. `research.md` lleva la tabla de cada regla de la versión de `main` (fichero y apartado)
  con su sitio en la nueva (fichero y apartado), y la revisión final la contrasta con las dos versiones. Criterio: SC-012.
- **FR-022**: La descripción de cada una de las tres está en tercera persona, dice qué hace la skill y cuándo se usa, y
  tiene como mucho 900 caracteres; la juzga el juez de FR-025 a FR-033, y su activación la deciden las evals de
  activación del job. Criterio: SC-010.
- **FR-023**: `CHANGELOG.md` sube la versión de cada una de las tres skills por encima de la de `main`, con lo que
  cambia, y dice los gates nuevos y el juez de las descripciones; `CONTRIBUTING.md` deja de pedir menos de 300 líneas y
  dice los gates, la guía y la regla de las referencias a mano. Criterio: SC-013.
- **FR-024**: La reescritura no cambia el `name` de ninguna skill, ni sus `kitlegal-applets` ni sus
  `kitlegal-referencias`, ni el contenido de las referencias generadas (H32 decide los nombres). Criterio: la revisión
  final contrasta el frontmatter y `references/` generado con los de `main`; `skills-check` da deriva si una generada
  difiere de lo que genera su fila.

#### El juez de las descripciones

- **FR-025**: La clase `habla_en_primera_o_segunda_persona` del juez con modelo de las evals (ADR 0037 y 0046) usa la
  rúbrica, el esquema y los casos de `evidencias/adr-0046/`, y comprueba la medida `evidencias/adr-0046/medida.json`,
  idénticos byte a byte: un control de identidad de `make ci` sale en rojo con un byte distinto en cualquiera de los
  cuatro. Criterio: SC-006.
- **FR-026**: En cada ejecución del job que va a juzgar descripciones, antes de abrir ninguna sesión, el job comprueba sin
  modelo que la medida corresponde —la huella de la rúbrica y de los casos que da `medida.json` es la de los ficheros que
  usa, y su modelo del juez y su versión de Claude Code son los fijados para el juez de las evals— y que se cumple —0
  defectos sin marcar de 63 y 0 correctos marcados de 36—; si no, sale en rojo sin abrir ninguna sesión. `make ci` hace
  la misma comprobación sin modelo. Criterio: SC-008.
- **FR-027**: Cada voto es una sesión con el modelo y la versión de Claude Code fijados para el juez de las evals (hoy
  `claude-opus-5-5` y 2.1.289), abierta con la orden de `scripts/evals-voto.sh`, con la rúbrica y el esquema de FR-025 y
  el mensaje de `mensaje_de` de `evidencias/adr-0046/guiones/juez.py`: la descripción, sin espacios al principio ni al
  final, entre `<descripcion skill="<skill>">` y `</descripcion>`, cada marca en su línea, una línea en blanco y
  «Responde a la pregunta de la rúbrica sobre esta descripción.». Un test de `make ci` compara el mensaje con el de un
  caso de `casos.yaml` construido como lo hace `mensaje_de`. Criterio: SC-007.
- **FR-028**: La frase citada se comprueba sin modelo, con la comparación de `frase_esta` de
  `evidencias/adr-0046/guiones/juez.py`: un «si» cuya frase no está en la descripción es un voto nulo y se pide otra
  vez, una sola; cuenta la repetición y, si vuelve a ser nulo, su «si» no cuenta como «si». El voto que corta el tope de
  tiempo de un voto también se pide otra vez, una sola, como en el juez de las evals; la sesión que termina con error,
  la que no devuelve una respuesta con la forma del esquema y la que corta un límite de uso no se repiten (FR-032).
  Criterio: SC-007.
- **FR-029**: Una descripción queda marcada con tres votos «si» unánimes, cada uno con su frase comprobada, pedidos por
  orden: el segundo solo si el primero marcó y el tercero solo si el segundo marcó. Con dos «si» y un «no», o con un
  primer «no», no queda marcada. Como mucho seis votos por descripción, contadas las repeticiones. Criterio: SC-007.
- **FR-030**: El juez juzga la descripción de cada skill en una ejecución del job cuya elección (ADR 0045; H30) incluye el
  cambio de una descripción, y reutiliza por huella el juicio de una descripción que no ha cambiado: el de una medida de
  `main` con la misma descripción, la misma skill y el mismo juez (rúbrica, esquema, casos, medida, modelo y versión de
  Claude Code), buscado como H30 busca los resultados de las evals; la descripción sin juicio reutilizable se juzga. El
  tope del trabajo que lo corre cubre su peor caso (las descripciones de las skills × 6 votos × el tope de un voto).
  Cuando la elección no incluye el cambio de ninguna descripción —la etiqueta `evals` o la suite completa sin
  descripciones cambiadas, el despacho manual sin propuesta, o un cambio solo de la rúbrica, el esquema, los casos, la
  medida, el modelo o la versión del juez—, el job no corre el juez de las descripciones: no abre ninguna sesión del
  juez, no publica el juicio de ninguna descripción y su umbral no aparece. «La elección incluye el cambio de una
  descripción» se decide por las rutas de la propuesta (alguna es un `SKILL.md` cuyo `description` cambia: la regla
  `cambiaLaDescripcion` de `eligeLaRuta`), no por el motivo de cada eval: en una propuesta que cambia también lo común,
  como esta, el motivo de cada eval es `comun` y el juez corre igual. Criterio: SC-010, SC-015.
- **FR-031**: El umbral de la clase es 0 descripciones marcadas de las juzgadas en la ejecución y hace fallar el job:
  una sola marcada pone en rojo el umbral, publicado con `decide: true` (contrato del ADR 0029). El informe da, de cada
  descripción marcada, la skill, la frase y el motivo, y de cada una, si se juzgó o se reutilizó y con qué procedencia;
  lo reutilizado no entra en el umbral (ADR 0045, punto 1). Criterio: SC-007, SC-010.
- **FR-032**: Una descripción que queda sin juzgar porque un voto no llega a darse (el tope de tiempo lo corta también en
  su repetición, o, sin repetición, la sesión termina con error, no devuelve una respuesta con la forma del esquema o la
  corta un límite de uso; FR-028) se publica sin juicio, con el motivo del voto, y hace fallar el job, como una respuesta sin juzgar del juez
  de las evals (FR-007 de H24; ADR 0037, decisión 2): queda «sin juzgar», no «sin marcar», y no cuenta como marcada. El
  job falla con un motivo propio, que nombra la skill y el motivo del voto y se distingue sin modelo de una descripción
  marcada. No deja juicio que reutilizar: la siguiente ejecución que juzga descripciones la vuelve a juzgar (FR-030).
  Criterio: SC-007.
- **FR-033**: El juez de las descripciones corre solo en el job de evals y en la medida del juez (FR-034): nunca en
  `make ci` ni en un paso del run (ADR 0032). Los tests de `make ci` lo ejercen con votos simulados, con un `claude`
  sustituto que no abre ninguna sesión. Criterio: SC-007, SC-014.
- **FR-034**: La medida del juez del job (etiqueta `evals-medir-juez` o entrada `medir_al_juez`) mide también la clase
  `habla_en_primera_o_segunda_persona` contra los 99 casos de `casos.yaml` con el código del job e imprime su medida,
  con la forma de `medida.json`, sin escribirla en el repositorio; la lanza una persona, nunca un paso del run.
  Criterio: SC-011.

### Key Entities

- **Cuerpo de `SKILL.md`**: el fichero sin el frontmatter, con la región de la tabla de comandos regenerada; lo que el
  agente carga al activar la skill. Medidas: líneas y caracteres (runas).
- **Descripción**: el `description` del frontmatter; lo que el agente lee siempre. Medidas: caracteres; juicio de tercera
  persona.
- **Referencia generada**: fichero de `references/` declarado en `kitlegal-referencias`, regenerado desde `data/` con
  la cabecera «generado desde» (FR-012).
- **Referencia a mano**: fichero `.md` de `references/` no declarado que no abre con esa cabecera; protocolo de detalle
  que `skills-sync` no toca (FR-012).
- **Juicio de una descripción**: de una skill y su descripción, los votos, si queda marcada, la frase y el motivo, si se
  juzgó o se reutilizó, su huella y su procedencia.
- **Medida del juez de las descripciones**: `evidencias/adr-0046/medida.json`: huellas de la rúbrica y de los casos,
  modelo, versión de Claude Code y los dos recuentos.

## Uso, de fuera adentro

Ninguna salida de este hito crece con los meses de uso de kitlegal: las normas y los bloques consultados no entran en
el texto de las skills ni en el juicio de las descripciones. Crecen con lo que se escribe en el repositorio (skills,
referencias, normas de `data/`), y los gates lo acotan.

- **La descripción (FR-005, FR-022)**: la lee el agente en cada sesión, de cada skill instalada, una vez, para decidir si
  la activa. Como mucho 900 caracteres por skill, unos 400 tokens: con las tres, como mucho 2700 caracteres frente a los
  2847 de hoy (1023, 993 y 831), y ninguna ya al borde del límite de la especificación. No depende de las preguntas
  hechas; cambia solo cuando una propuesta cambia una descripción.
- **El cuerpo (FR-004, FR-020)**: lo carga el agente una vez por pregunta por cada skill que activa: con `legal-core` y
  `boe-legislacion` juntas, la pregunta de un plazo con su artículo, como mucho 2 × 11 000 caracteres, unos 9700 tokens,
  frente a unos 19 300 de hoy (10 760 + 8520, ADR 0046). Con cientos de normas y miles de bloques consultados no cambia:
  el texto lo escribe una persona y el gate lo acota.
- **Las referencias a mano (FR-012, FR-020)**: las lee el agente solo cuando la clase de pregunta lo pide, como dice el
  cuerpo, de 0 a pocas por pregunta; cada una se lee entera o, con más de 100 líneas, por el apartado que da su índice
  (FR-009). Su tamaño lo fija quien la escribe y no crece con el uso.
- **Las referencias generadas (FR-009, FR-012)**: crecen con `data/`: `normas.md` tiene hoy 24 líneas, y con cientos de
  normas en `data/normas.yaml` (unos 150 bytes por fila, unos 45 KB con 300) pasa de 100 líneas y abre con su índice,
  para que el agente lea el apartado que necesita y no el fichero entero.
- **Los defectos de los gates (FR-011)**: los lee quien ejecuta `make ci` o `make skills-sync` —una persona o una tarea
  de un run—, una línea de unos 100 bytes por regla incumplida y skill. Dejan de darse cuando la skill se corrige.
- **El juicio de las descripciones en el informe (FR-031)**: lo lee quien revisa la propuesta, una vez por ejecución que
  juzga descripciones: una entrada por skill (tres hoy), de unos 300 bytes con la frase y el motivo si está marcada. La
  marca deja de darse cuando la propuesta cambia esa descripción y el juicio nuevo no la marca; la marca «reutilizado»,
  cuando cambia la descripción o el juez (otra huella) o cuando su medida de `main` sale de la ventana de búsqueda de
  H30.
- **La comprobación de la medida (FR-026)**: una línea en el registro del job por ejecución que juzga descripciones; deja
  de fallar cuando la medida vuelve a corresponder y a cumplirse, lo que solo cambia una persona fuera de un run.

## Success Criteria *(mandatory)*

### Measurable Outcomes

- **SC-001**: `make ci` en verde con los gates nuevos sobre las tres skills. Lo hace cumplir el propio `make ci`, que el
  workflow mide antes de la revisión y en el cierre.
- **SC-002**: Cada gate sale en rojo con su mutante y en verde en su límite, en un test de `make ci` (`skills-check`):
  `name` con `claude` y con `anthropic`; una etiqueta XML en `name` y en `description` (y «< 5000» no lo es); un cuerpo
  de 500 líneas (499, en verde); un cuerpo de 11 001 caracteres con caracteres de varios bytes (11 000, en verde); una
  descripción de 901 caracteres (900, en verde); un subdirectorio y un fichero que no es `.md` en `references/`; una
  referencia sin enlace desde `SKILL.md`; una referencia enlazada desde otra; una de 101 líneas sin índice (100, en
  verde) y una de 101 líneas sin ningún encabezado de nivel 2 (100, en verde). Con cualquiera de ellos sin detectar, el
  test sale en rojo.
- **SC-003**: Un test de `make ci` comprueba que `make skills-sync`, sobre una copia con una referencia a mano y una
  generada alterada (un byte cambiado, y otra sin su cabecera), deja la referencia a mano idéntica byte a byte y rehace
  las generadas; sale en rojo si toca la primera o no rehace una de las otras.
- **SC-004**: Un test de `make ci` comprueba que el conjunto de dependencias de una eval incluye la referencia a mano de
  su skill en la parte (b), sin parte (c), y que una propuesta que solo la cambia elige las evals de esa skill; y que una
  skill nombrada solo desde una referencia a mano entra en el conjunto con su `SKILL.md` y sus `references/`; sale en
  rojo si no la incluye o si da error.
- **SC-005**: Cada control de contenido de FR-016 sale en rojo en `make ci` con su regla movida a una referencia a mano de
  una copia de la skill: un mutante por cada control de FR-016, FR-018 y FR-019, entre ellos la expresión prohibida, la
  fecha AAAAMMDD, la orden sin su forma para PowerShell, la línea sin consulta, las formas de aviso, de hallazgo y de la
  redacción fechada, la norma nombrada que no está en `data/normas.yaml`, la instrucción de las evals y la orden de
  ejemplo que la caché no sirve sin red.
- **SC-006**: El control de identidad de `make ci` sale en rojo con un byte distinto en la rúbrica, el esquema, los casos
  o la medida respecto de `evidencias/adr-0046/`.
- **SC-007**: Tests de `make ci` del juicio de una descripción con votos simulados, sin abrir ninguna sesión: marca con
  tres «si» y su frase; no marca con dos «si» y un «no» ni con un primer «no», y entonces no pide los votos siguientes;
  un «si» con una frase que no está es nulo y se repite una vez; un voto cortado por el tope se repite una vez y, si
  vuelve a cortarse, la descripción queda sin juzgar y el job falla; una sesión con error no se repite y la deja sin
  juzgar; no se piden más de seis votos por descripción; el
  mensaje es el de `mensaje_de`; una descripción marcada pone en rojo el umbral 0 de la clase y su informe lleva la
  frase y el motivo. Con cualquiera de ellos distinto, el test sale en rojo.
- **SC-008**: Un test de `make ci` de la decisión del job comprueba que, con la medida que no corresponde (otra huella de
  la rúbrica o de los casos, otro modelo, otra versión de Claude Code) o que no se cumple (62 de 63, 1 de 36), el job
  falla antes de abrir ninguna sesión; sale en rojo si planifica una sola.
- **SC-009**: El cuerpo de cada una de las tres skills tiene menos de 500 líneas y como mucho 11 000 caracteres, y su
  descripción como mucho 900: lo hacen cumplir los gates de SC-002 sobre el repositorio, en `make ci`.
- **SC-010**: El job de evals sobre la cabeza de la propuesta, con las tres descripciones cambiadas, mide las 45 evals en
  los dos modos, juzga las tres descripciones y sale en verde con los umbrales vigentes (ADR 0029, 0038 y 0045), sin
  nada reutilizado y sin ninguna descripción marcada. Lo hace cumplir el propio job, cuyos umbrales con `decide: true`
  lo hacen fallar, en la medición del cierre; que una descripción cambiada elige las 45 evals, sin reutilizar ninguna,
  lo hace cumplir en `make ci` el test de la elección de H30 (su mutante de una descripción, que sale en rojo si la
  elección difiere en una sola eval).
- **SC-011**: Antes de fusionar, una persona lanza la medida del juez de las descripciones con el código del job
  (etiqueta `evals-medir-juez`) y da 63 de 63 y 0 de 36. La mide una persona, no el run.
- **SC-012**: La tabla de `research.md` tiene una fila por cada regla de la versión de `main` de cada skill con su sitio
  en la nueva; la revisión final la contrasta y no encuentra ninguna regla perdida ni añadida.
- **SC-013**: `docs/SKILLS.md` existe con lo de FR-001, sus cifras son las de los gates, y `CHANGELOG.md` y
  `CONTRIBUTING.md` dicen lo de FR-023; lo contrasta la revisión final.
- **SC-014**: Ninguna orden de `make ci` ni de un paso del run abre una sesión del juez de las descripciones: sus tests
  usan un `claude` sustituto (FR-033).
- **SC-015**: El tope del trabajo del job que juzga las descripciones cubre su peor caso (FR-030): un test de `make ci`
  sobre la definición del job sale en rojo si el tope es menor que lo que suman el peor caso de sus evals y el de las
  descripciones (las descripciones de las skills × 6 votos × el tope de un voto).

### Criterios del hito, tal como los escribe

**Controles** (SC-002 a SC-008): «un mutante de `make ci` por gate; un test de que `make skills-sync` conserva una
referencia a mano y rehace una generada; un test de que el conjunto de dependencias incluye la referencia a mano; cada
control de contenido falla con su regla movida a una referencia; el control de identidad del juez de las descripciones
con `evidencias/adr-0046/`; tests del juicio de la descripción con votos simulados (marca con tres «si» y su frase, no
con dos, nulo con una frase que no está), sin abrir ninguna sesión.»

**Aceptación** (SC-001, SC-010 y SC-011): «`make ci` en verde con los gates nuevos sobre las tres skills; y el job de
evals sobre la cabeza de la propuesta, que con las tres descripciones cambiadas mide las 45 evals en los dos modos y
juzga las tres descripciones, en verde con los umbrales vigentes (ADR 0029, 0038 y 0045), sin nada reutilizado y sin
ninguna descripción marcada. Antes de fusionar, una persona lanza la medida del juez de las descripciones con el código
del job (etiqueta `evals-medir-juez`) y da 63 de 63 y 0 de 36.»

**Alcance** (FR-018, FR-019, FR-021): «La reescritura cambia la forma, no las reglas: ninguna regla de la skill
desaparece ni aparece»; «mover un texto a una referencia no lo saca de ningún control»; «Las órdenes de ejemplo de una
referencia son de lo que la caché de las sesiones sirve sin red (ADR 0043), como las de `SKILL.md`».

## Assumptions

- Las dos fuentes de la guía se citan con la fecha y las frases que registra el ADR 0046 (2026-10-23): el run no tiene
  red y no las consulta.
- «La elección incluye el cambio de una descripción» es la regla de FR-030: alguna ruta de la propuesta es un
  `SKILL.md` cuyo `description` difiere entre la base y la cabeza (la regla `cambiaLaDescripcion` de H30), sea cual sea
  el motivo de cada eval. Con las tres descripciones cambiadas, la elección son las 45 evals (ADR 0046, «En contra»).
- El modelo y la versión de Claude Code del juez de las descripciones son los del juez de las evals,
  `MODELO_DEL_JUEZ` y `VERSION_DE_CLAUDE_CODE_DEL_JUEZ` de `.github/workflows/evals.yml` (hoy `claude-opus-5-5` y
  2.1.289), los mismos con que se tomó la medida de `evidencias/adr-0046/`.
- El binario empotra e instala todo fichero de `references/` (`internal/core/instalacion`), así que las referencias a
  mano viajan e instalan como las generadas sin cambiar la instalación.
- Las referencias generadas de hoy no cambian de contenido; `kitlegal-referencias` sigue declarando las mismas.
- La serie de cierres de cada skill empieza de cero, porque su clave es el árbol de `skills/<skill>` (ADR 0038); no lo
  hace este hito, es la consecuencia de cambiar el árbol.
- La primera release tras H31 necesita la suite completa sobre el commit etiquetado (ADR 0045): la decide y la lanza una
  persona, fuera del run.

## Fuera de alcance

- Los nombres de las skills y las skills por capacidad (H32).
- La cobertura en positivo y `⚠ NO CUBIERTO:` (H33).
- Reglas nuevas en el protocolo de ninguna skill, y quitar alguna.
- Evals o casos nuevos de las skills.
- El umbral, la rúbrica y los casos de `afirma_lo_no_leido` y de las demás clases de hoy del juez de las evals.
- La rúbrica, el esquema, los casos y la medida del juez de las descripciones, que son los de `evidencias/adr-0046/` y
  solo los cambia una persona fuera de un run; volver a medir la razón de caracteres por token.
- El ADR 0046, la constitución, `CLAUDE.md`, la DoD §1.10 y la fila de §3 de `docs/ROADMAP.md`, ya en `main`.
- `skills-ref` u otra validación en Python en `make ci` (ADR 0046, punto 3), y un presupuesto en tokens.
- El juez de las descripciones en `make evals-sondeo`, en `make ci` o en un paso del run.
- Las skills de agente que no son del producto (`.agents/skills/`, `seo-semanal`).
- Cambios en `scripts/workflow/`, `.specify/` e `internal/workflow/`.
- Cambiar los generadores de las referencias generadas para que escriban su índice o sus encabezados de nivel 2
  (Clarificación Q4): si un cambio posterior de `data/` lleva una por encima de 100 líneas, ese cambio adapta su
  generador.
- La release y su etiqueta, y lanzar la medida del juez: los hace una persona.
