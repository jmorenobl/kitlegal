# Feature Specification: H32 · Arquitectura de skills: una entrada genérica, skills por capacidad y las materias como referencias

**Feature Branch**: `030-h32-arquitectura-de-skills`

**Created**: 2026-10-29

**Status**: Draft

**Input**: el fichero del hito, `docs/roadmap/in-progress/H32-arquitectura-de-skills-una-entrada-generica.md`, la única
entrada humana del run (ADR 0018), con su evidencia en `evidencias/sondeos-20261028/` (`brazo-P80`) y en los jobs de
evals de la #207, que el run usa y no escribe. Parte de `main` en `c991808`.

## Contexto

El hito reparte las tres skills de hoy —`boe-legislacion`, `legal-core` y `jurisprudencia`— en una skill de entrada y
skills por capacidad, saca lo administrativo a un paquete de materia y fija el molde con el que entra una materia nueva
(ADR 0044, puntos 3, 5 y 7). Al repartir, arregla los tres defectos que quedan tras H31.1:

- **`cuenta_su_proceso`, por la línea de la comprobación.** En `brazo-P80`, las 4 marcadas de 80 dan el resultado de
  `kitlegal graph check`; en el job de `cd21c7b` (ejecución 40700995075), 17 de las 21 marcadas de 120 de
  `claude-sonnet-5-5`, y en el de `0f9ef2f` (40682099061), las 16 de 120. Se concentra en la eval 05: 6 de sus 12
  respuestas en las dos ejecuciones. Antes de H31 eran el 1,8 % y el 2,8 %, con un cuerpo que decía que un bloque sin
  `version-obsoleta` no da nada a la respuesta (`e1bfb77`, `skills/boe-legislacion/SKILL.md`, líneas 215 a 220).
- **`afirma_lo_no_leido`, por reglas de memoria y materia pegada a una remisión.** Las 2 marcadas de `brazo-P80`
  explican «íntegros» con una regla del rendimiento neto y describen el art. 7 de la Ley 35/2006 sin su número; el paso 5
  de hoy lo permite («o el número, o la materia»; «Puedes avisar de que una materia se regula en otra parte»).
- **Dos formas para la misma cita.** `boe-legislacion` cita `[BOE-A-2015-10565, bloque daprimera]`; la línea
  `⚠ NO CUBIERTO:` de `legal-core` pide `(BOE-A-2015-10565)`, la cobertura por materia escribe
  `(BOE-A-2015-10565, bloque daprimera)` y el paso 6 de `legal-core` manda citar «como ella cita». Una sesión mezcló las
  dos y dejó fuera la norma que lo rige: `cobertura_sin_declarar` dio 1 de 36 respuestas en las evals 07, 08 y 12, y tumbó
  el job.

Medido en esta sesión con `wc -m` sobre el cuerpo (desde la línea que sigue al `---` que cierra el frontmatter): 10 995
caracteres en `boe-legislacion`, 10 944 en `legal-core` y 10 859 en `jurisprudencia`; el máximo es 11 000 (ADR 0046).

Nombres fijados para que cada requisito tenga una sola lectura (los nombres de las skills los fija el plan, FR-007):

- **La skill de entrada**: la skill nueva por la que entra una pregunta de cualquier materia; decide qué capacidad la
  resuelve y qué cubre la base.
- **Una skill de capacidad**: cada skill de la entrega que tiene herramienta: la de **normativa** (`kitlegal boe` y
  `kitlegal graph`), la de **jurisprudencia** (`kitlegal cita`), la de **plazos** (`kitlegal plazos`) y la de
  **territorio** (`kitlegal territorio`). Si el plan junta dos capacidades en una skill, «la de plazos» y «la de
  territorio» nombran esa misma skill.
- **Un paquete de materia**: lo que una materia trae a la base —normas, una referencia por skill de capacidad que la
  necesite, filas de cobertura y evals— y nada más (ADR 0044, punto 5). **El paquete de lo administrativo** es el primero;
  **la vertical de ensayo**, el de `testdata/` que prueba el molde.
- **Un nombre retirado**: el de una skill de `main` en `c991808` que la entrega ya no empotra.
- **La forma única de cita**: `<forma legible> [<identificador>, bloque <id>]`, la de `boe-legislacion` hoy; dentro de
  los corchetes, el identificador, una coma, un espacio, la palabra `bloque`, un espacio y el id, y nada más. **Una cita
  de un bloque** es toda aparición de un identificador `BOE-A-…` o del marcador `<identificador>` seguido de
  `, bloque `.
- **Las evals por su número**: las de hoy, por su fichero en `c991808`: la 05 y la 15 de normativa son
  `evals/boe-legislacion/05-trlrhl-impuestos-municipales.yaml` y `15-irpf-rendimientos-por-materia.yaml`; la 14 y la
  17, `14-trlrhl-impuestos-por-materia.yaml` y `17-ltaibg-plazo-por-materia.yaml`; las 07, 08 y 12 de cobertura, las de
  `evals/legal-core/` con ese prefijo. **Las evals por materia** son las de `por_materia: true`.

## Clarifications

### Session 2026-10-29

- Q: Cuando la pregunta es de una materia que la base no cubre (la póliza de docs/USO.md, 2026-10-21) o pide documentos o destinatarios, ¿con qué forma lo declara la skill de entrada? → A: Sin forma fija nueva: lo dice en prosa, se apoya en la norma leída con kitlegal citada en la forma única, no da nada de memoria ni pide `plazos calcular`; `⚠ NO CUBIERTO:` sigue siendo la de lo que la ley regula aparte, con su precepto de la LPAC; la eval de FR-035 no exige línea (auto: conservadora; criterio d; fuente: hito H32 «Una sola forma de cita» y «Fuera de alcance», entrada de H33, ADR 0044 punto 4)
- Q: ¿Qué normas cuentan como «de un paquete» para FR-021, de modo que ningún cuerpo de una skill de capacidad las nombre? → A: Las que `data/normas.yaml` marca con la vertical de un paquete (en la entrega, lo administrativo), por identificador, título o abreviatura; salen los ejemplos con la Ley 39/2015 y los de la LJCA si el plan la marca; la suite toma la lista de los datos de la rama y mantiene el mínimo de las evals por materia (auto: criterio a; fuente: hito H32 «Las materias como referencias», ADR 0044 puntos 4 y 5, FR-021 y FR-022)
- Q: Con `kitlegal skills install <nombres>` que no pide todas las skills, ¿se retira una skill de nombre retirado cuyas huellas coinciden aunque el ámbito no quede con todas las skills que heredan sus preguntas? → A: Sí: se retira siempre que las huellas coincidan, con independencia de lo que pida la invocación o haya en el ámbito; no se avisa de herederas ausentes (auto: criterio a; fuente: hito H32 «Las instalaciones existentes», FR-026, FR-025, FR-027, constitución «Criterio de decisión autónoma» punto 2)

## User Scenarios & Testing *(mandatory)*

### User Story 1 - Una pregunta de cualquier materia entra por una skill que sabe qué la resuelve (Priority: P1)

Quien pregunta por una situación legal —una póliza de seguro, un recurso, un artículo— no sabe qué herramienta la
resuelve. La skill de entrada clasifica la pregunta, dice qué capacidad la resuelve y qué cubre la base, y lo dice si no
lo cubre, sin dar nada de memoria.

**Why this priority**: es el objetivo del hito y lo que pidió el uso (`docs/USO.md`, 2026-10-21): la póliza se leyó
bien y nada dijo lo que faltaba alrededor de la norma.

**Independent Test**: las evals de activación de la skill de entrada en el job (SC-015) y la revisión final del cuerpo
contra la tabla del plan (FR-001 a FR-004).

**Acceptance Scenarios**:

1. **Dado** la pregunta de la póliza de `docs/USO.md` (2026-10-21), **Cuando** se mide su eval en el job, **Entonces**
   se activa la skill de entrada, la sesión lee la Ley 50/1980 (`BOE-A-1980-22501`) con kitlegal y la cita, y no pide
   `plazos calcular` (FR-035).
2. **Dado** una pregunta que resuelve una sola capacidad, **Cuando** se mide su eval, **Entonces** se activan la skill
   de entrada y la de esa capacidad, y ninguna skill de capacidad que la eval declare en `no_se_activan` (FR-036).
3. **Dado** el frontmatter de cada skill de la entrega, **Cuando** se lee su descripción, **Entonces** ninguna dice
   «derecho público» (FR-004).

---

### User Story 2 - Cada skill de capacidad responde sola y cita de una sola forma (Priority: P1)

Una respuesta que usa dos skills no puede recibir dos instrucciones que choquen. Cada skill de capacidad lleva en su
cuerpo su protocolo, las reglas invariantes y las formas fijas de sus respuestas, no remite a otra skill para citar, y
todas citan un bloque igual.

**Why this priority**: es el defecto que tumbó el job de la #207 (`cobertura_sin_declarar`, 1 de 36).

**Independent Test**: los dos controles nuevos de `make ci`, cada uno con su mutante en rojo (SC-003, SC-004), y el
cierre (SC-011).

**Acceptance Scenarios**:

1. **Dado** los `SKILL.md` y las referencias de la entrega, **Cuando** corre el control de la forma única de cita,
   **Entonces** sale en verde; **y Dado** una referencia con `(BOE-A-2015-10565, bloque daprimera)` o con
   `[art. 46.1 de la LJCA, BOE-A-1998-16718, bloque a46]`, **Entonces** sale en rojo y nombra el fichero y la línea.
2. **Dado** dos `SKILL.md` que llevan una línea `⚠` con la misma etiqueta, **Cuando** corre el control de las formas
   compartidas, **Entonces** las líneas de esa etiqueta son las mismas en los dos; con una cambiada en uno, sale en rojo.
3. **Dado** la cobertura por materia generada, **Cuando** se lee la fila de la aplicación de los tributos, **Entonces**
   lleva la línea `⚠ NO CUBIERTO:` entera, con la Ley 58/2003 por su nombre y `BOE-A-2003-23186` y la DA 1ª.2.a citada
   `[BOE-A-2015-10565, bloque daprimera]`, y es la forma del cuerpo con sus marcadores sustituidos (FR-012).
4. **Dado** el informe del job de cierre, **Cuando** se lee `cobertura_sin_declarar` de cada modo, **Entonces** es 0,
   contado con la medida que exige la norma esperada y la cita del precepto de su fila en la misma línea (FR-014).

---

### User Story 3 - La skill de normativa no cuenta la comprobación ni pega materia a una remisión (Priority: P1)

Quien pregunta no ve `kitlegal graph check`: la respuesta no dice nada de ella si no hay `version-obsoleta`. Y de un
precepto remitido que no se ha leído, la respuesta repite las palabras del texto leído, sin materia.

**Why this priority**: son las marcas que quedan de `cuenta_su_proceso` y de `afirma_lo_no_leido`.

**Independent Test**: la suite de aceptación contra el texto que fija el plan (FR-015 a FR-017), el cierre (SC-010) y
el sondeo de la persona (SC-016).

**Acceptance Scenarios**:

1. **Dado** el cuerpo de la skill de normativa, **Cuando** se lee el paso que pide `kitlegal graph check`, **Entonces**
   lleva, en ese paso, que un bloque sin entrada `version-obsoleta` no da nada a la respuesta: ni una línea, ni una frase
   junto a los avisos o a la fecha de vigencia.
2. **Dado** su paso 5, **Cuando** se lee, **Entonces** no contiene «o el número, o la materia» ni «Puedes avisar de que
   una materia se regula en otra parte», y dice que una remisión no leída se repite con las palabras del texto leído,
   sin materia, y que explicar una palabra del texto leído con lo que regula otro precepto es afirmar ese precepto, cada
   cosa con su porqué.
3. **Dado** sus pasos 3 y 6 y la primera viñeta de su paso 4, **Cuando** se comparan con los de `main`, **Entonces** son
   iguales, salvo lo que la tabla del plan muestre que pide la comprobación.

---

### User Story 4 - Una materia nueva es un paquete de ficheros (Priority: P1)

Quien añade una materia escribe sus normas, una referencia por skill de capacidad, sus filas de cobertura y sus evals, y
`make skills-sync` enlaza cada referencia desde su skill; no escribe Go ni edita ningún `SKILL.md`. Lo administrativo de
hoy es el primer paquete.

**Why this priority**: es la segunda mitad del objetivo (ADR 0044, punto 5) y lo que el principio X exige.

**Independent Test**: el test de la vertical de ensayo en `make ci` (SC-005) y la suite de aceptación (SC-017).

**Acceptance Scenarios**:

1. **Dado** la vertical de ensayo de `testdata/`, **Cuando** el test la instala sobre una copia de `skills/` y corre la
   sincronía, **Entonces** pasa `skills-check`, sus evals se cargan, y el diff de la copia no tiene ningún `.go` ni
   ningún cambio en un `SKILL.md` fuera de las marcas que genera `make skills-sync`.
2. **Dado** el cuerpo de cada skill de capacidad, **Cuando** se busca una norma de las evals por materia, **Entonces** no
   está (FR-021).
3. **Dado** el paquete de lo administrativo, **Cuando** se comparan sus referencias generadas con las de `legal-core` en
   `c991808`, **Entonces** cada fila lleva los mismos datos, con la cita en la forma única.

---

### User Story 5 - Quien ya instaló las skills no se queda con dos protocolos (Priority: P2)

Quien instaló `boe-legislacion` o `legal-core` y vuelve a instalar tiene en su `.agents/skills/` las del nombre anterior;
dos skills con el mismo protocolo se activarían a la vez. La instalación retira lo que instaló con el nombre anterior si
nadie lo ha tocado, y avisa si no.

**Why this priority**: sin ello, renombrar crea el defecto que el hito arregla, en el equipo de quien ya usa kitlegal.

**Independent Test**: el e2e de la instalación con un nombre retirado (SC-006).

**Acceptance Scenarios**:

1. **Dado** un ámbito cuyo manifiesto declara `boe-legislacion` con sus ficheros intactos y su enlace de host, **Cuando**
   se ejecuta `kitlegal skills install` con el binario de la entrega, **Entonces** sale con 0, el directorio, el enlace y
   la entrada del manifiesto de `boe-legislacion` ya no están y `data` lo dice.
2. **Dado** el mismo ámbito con un fichero de `boe-legislacion` editado, **Cuando** se instala, **Entonces** sale con 0,
   nada de `boe-legislacion` cambia, y `data` y la salida legible llevan el aviso con su ruta.
3. **Dado** el ámbito del escenario 1, **Cuando** se instala con `--dry-run`, **Entonces** dice lo que retiraría y no
   toca nada.

---

### User Story 6 - Las evals siguen a su pregunta y el cierre mide todas las skills (Priority: P1)

Cada eval de hoy pasa a la skill que hereda su pregunta, con su historia, y el job mide todas las skills de la entrega
con los umbrales de hoy.

**Why this priority**: es la aceptación sin persona del hito.

**Independent Test**: el informe del job de cierre (SC-010 a SC-015) y `make ci` (SC-001, SC-008).

**Acceptance Scenarios**:

1. **Dado** el informe del job de cierre, **Cuando** se leen sus umbrales, **Entonces** los que deciden se cumplen en
   todas las skills, `afirma_lo_no_leido:claude-sonnet-5-5` como mucho 0.028 en la de normativa y 0.017 en la que hereda
   las evals de `jurisprudencia`, y `cuenta_su_proceso` se publica por modo.
2. **Dado** una eval de hoy, **Cuando** se sigue su fichero en la rama, **Entonces** está en la carpeta de evals de la
   skill que hereda su pregunta, renombrada en git, y pasa.
3. **Dado** la definición del job, **Cuando** corre `TestDefinicionDelJob` en `make ci`, **Entonces** la matriz tiene
   todas las skills de la entrega y ningún nombre retirado, y el tope de cada trabajo cubre su peor caso.

---

### User Story 7 - La guía, el ADR y lo que nombra las skills dicen el reparto (Priority: P2)

Quien escribe una skill o un paquete lee en `docs/SKILLS.md` el reparto, el molde y lo medido; el ADR del hito registra
los nombres; y nada de lo que se instala, se documenta para instalar o se ejecuta nombra una skill retirada.

**Why this priority**: H33 a H38 y las verticales trabajan sobre este reparto.

**Independent Test**: la revisión final (FR-041, FR-042) y la búsqueda de la suite de aceptación (SC-018).

**Acceptance Scenarios**:

1. **Dado** `docs/SKILLS.md` de la rama, **Cuando** se lee, **Entonces** dice qué skills hay y qué cubre cada una, el
   molde de un paquete y lo medido, con sus cifras y sus fuentes.
2. **Dado** la búsqueda de FR-032, con sus directorios y sus excepciones, **Cuando** corre en la suite de aceptación,
   **Entonces** no encuentra ningún nombre retirado.

---

### User Story 8 - Una persona decide la fusión con el sondeo (Priority: P2)

El cierre mide pocas respuestas por eval; antes de fusionar, una persona lanza un sondeo de 280 respuestas con la skill
de la propuesta y decide con sus cifras (ADR 0047).

**Why this priority**: es la medida con potencia del hito, y la lanza una persona fuera del run (ADR 0032).

**Independent Test**: el sondeo de la persona (SC-016).

**Acceptance Scenarios**:

1. **Dado** el sondeo con la skill de la propuesta, **Cuando** cada parte da como mucho su límite de la tabla de SC-016,
   **Entonces** la propuesta se puede fusionar.
2. **Dado** el mismo sondeo, **Cuando** alguna parte no lo cumple, **Entonces** la propuesta no se fusiona y se decide
   con sus cifras y sus respuestas marcadas; el sondeo no se repite.

---

### Edge Cases

- **Una pregunta que pide dos capacidades** (un plazo que cuenta los festivos de un municipio): la tabla del plan dice
  qué skills se activan; cada una responde sola y la forma fija que comparten es la misma línea (FR-005, FR-013).
- **Una pregunta de una materia sin paquete** (la póliza): la lee la de normativa; ninguna skill da de memoria un plazo,
  un órgano ni una vía (FR-001, FR-035).
- **Un paquete cuyo enlace no cabe**: si una línea generada lleva un cuerpo a más de 11 000 caracteres, `make
  skills-sync` y `make ci` dan el defecto; el paquete espera a un hito de la base (FR-009, FR-024).
- **Un nombre retirado pedido por su nombre** (`kitlegal skills install boe-legislacion`): sale con 2, como hoy toda
  skill que el binario no empotra (FR-030).
- **La skill del nombre retirado en `-g` y en el proyecto**: cada ámbito se trata con su manifiesto, por separado.
- **Una skill no empotrada que no tiene un nombre retirado** (la instaló el binario de otra rama): se queda como hoy
  (FR-030).
- **Ficheros del kit, de la caché o del manifiesto manipulados por algo que no es el binario**: regla genérica, defecto
  `inesperado`, código 1 (ADR 0023), sin casos. Lo único que el hito trata es lo que pide: los ficheros de una skill de
  nombre retirado que no tienen la huella del manifiesto (FR-028).

## Requirements *(mandatory)*

### Functional Requirements

#### La skill de entrada

- **FR-001**: La entrega añade la skill de entrada, empotrada en el binario e instalada por `kitlegal skills install` con
  las demás. Su cuerpo lleva el método jurídico genérico para cualquier materia: (1) qué tipo de pregunta es; (2) qué
  capacidad la resuelve —normativa, jurisprudencia, plazos o territorio— y qué skill la tiene; (3) qué cubre la base,
  por capacidad y por paquete; y (4) si no lo cubre, cómo lo dice, sin dar de memoria ningún plazo, órgano ni vía, ni
  aplicar los de otra materia. Para una pregunta de una materia que ninguna capacidad ni paquete cubre (la póliza de
  `docs/USO.md`, 2026-10-21), la skill de entrada lo dice en prosa, sin forma fija nueva: la base no la cubre, se apoya
  en la norma que la sesión haya leído con kitlegal, citada en la forma única, y no da de memoria ningún plazo, órgano ni
  vía, ni aplica los de otra materia, ni pide `plazos calcular`. La línea `⚠ NO CUBIERTO:` sigue siendo la de lo que la
  ley regula aparte, con su precepto de la LPAC (FR-011); la forma fija única de todas las materias es de H33. Criterio:
  el plan da el texto de cada punto y la revisión final lo coteja; las evals de FR-035 y FR-036 pasan en el job.
- **FR-002**: La skill de entrada declara no cubiertos, en prosa y del mismo modo que FR-001, los documentos y los
  destinatarios, que no tienen herramienta genérica hasta H35 y H37; el cotejo de una sentencia sigue en la de jurisprudencia. Criterio: el plan da el texto y la
  revisión final lo coteja.
- **FR-003**: La skill de entrada no lleva ninguna regla ni forma fija que una skill de capacidad necesite para
  responder: cada una está en el cuerpo de la skill de capacidad que la usa (FR-006). Criterio: los controles de FR-006;
  una forma fija que la de entrada lleve también es la misma línea que en la de capacidad (FR-013).
- **FR-004**: La descripción de la skill de entrada habla de cualquier materia, y ninguna descripción de la entrega
  presenta su skill como de «derecho público». Criterio: la suite de aceptación sale en rojo si el frontmatter de un
  `SKILL.md` de `skills/` contiene «derecho público»; la longitud y la tercera persona, SC-007.

#### Las skills de capacidad

- **FR-005**: Cada capacidad con herramienta —normativa (`kitlegal boe` y `kitlegal graph`), jurisprudencia
  (`kitlegal cita`), plazos (`kitlegal plazos`) y territorio (`kitlegal territorio`)— está en exactamente una skill de
  capacidad, y no hay skill de capacidad sin herramienta (documentos, destinatarios). El plan fija qué skills hay y cómo
  se reparten las tres de hoy, por unidades coherentes («Design coherent units»): ninguna pregunta habitual necesita
  cargar otra skill con instrucciones que choquen, ni una skill es tan ancha que no se active con precisión. Lo dice por
  cada eval de hoy: qué skills se activan con su pregunta y qué formas fijas da cada una. Criterio: la suite de
  aceptación sale en rojo si uno de los applets `boe`, `graph`, `cita`, `plazos` y `territorio` no está en el
  `kitlegal-applets` de exactamente una skill de capacidad; la tabla del plan la coteja la revisión final.
- **FR-006**: Cada skill de capacidad responde sola: su cuerpo lleva su protocolo, las reglas invariantes de `CLAUDE.md`
  —no inventar contenido legal, cada afirmación con su cita, distinguir ley y reglamento, señalar la variación
  autonómica, declarar no cubierto lo no declarado sin plazos, órganos ni vías de memoria, no aplicar el procedimiento
  común a lo que la ley regula aparte, no concluir que algo no existe, no actuar en nombre de nadie, y sin herramienta
  ni binario la respuesta lo dice— y las formas fijas que dan sus respuestas, aunque otra skill las repita. Ni su cuerpo
  ni sus referencias remiten a otra skill de la entrega: una remisión es el nombre de otra skill como código en línea
  solo (`` `<nombre>` ``) o detrás de la palabra «skill». Criterio: el plan da, por skill, la frase de cada invariante y
  las formas fijas que comprueban sus evals, y la revisión final lo coteja; la suite de aceptación sale en rojo si una de
  esas formas no está en un bloque `text` del `SKILL.md` de su skill, o si el cuerpo o una referencia de una skill de
  capacidad lleva una remisión a otra skill de la entrega.
- **FR-007**: Los nombres de las skills los fija el plan y los registra el ADR del hito (FR-042): de capacidad y no de
  fuente —H34 añade EUR-Lex a la misma de normativa—, con el mismo patrón en todas, en español y como sustantivo
  («Noun phrases»). Criterio: la validación del frontmatter de `make ci` (sin `anthropic` ni `claude`, ADR 0046, punto 3)
  y la suite de aceptación, que sale en rojo si un nombre de la entrega contiene `boe`, `cendoj` o `eur`.
- **FR-008**: La descripción de la skill de plazos dice qué plazos cuenta hoy: los de la LPAC y la interposición del
  recurso contencioso-administrativo, hasta H36. Criterio: el plan da la descripción y la revisión final la coteja.
- **FR-009**: Cada cuerpo de la entrega tiene como mucho 11 000 caracteres (ADR 0046), con lo que genera `make
  skills-sync`. El sitio para los arreglos de FR-015 y FR-016 sale del reparto; si no basta, como en H31.1: lo que solo
  pide una clase de pregunta va a la referencia de esa clase, con el paso que la pide y su condición, y lo demás se
  reescribe más corto. No salen del cuerpo las formas fijas, la tabla de comandos ni lo que toda respuesta necesita.
  Criterio: `make skills-sync` y la comprobación de las skills de `make ci` salen en rojo con 11 001 o más (gate de H31,
  sin cambios); el plan da, por cada texto movido o reescrito, el de `c991808` y el de la entrega, y la revisión final lo
  coteja.

#### Una sola forma de cita

- **FR-010**: Toda cita de un bloque en un `SKILL.md` o en una referencia de cualquier skill de la entrega, escrita a mano
  o generada, también en las formas fijas (`⚠ PLAZO:`, `⚠ PLAZO SIN CALCULAR:`, `⚠ REDACCIÓN A FECHA:`,
  `⚠ REDACCIÓN DE HOY:`, `⚠ SIN REDACCIÓN VIGENTE:`, `⚠ REDACCIÓN MODIFICADA:`, `⚠ NO CUBIERTO:`) y en las tablas de las
  referencias generadas, va en la forma única de cita, con el marcador `<id>` para el id. Criterio: un control nuevo de
  `make ci` sale en rojo, con el fichero y la línea, en cada cita de un bloque que no está entre corchetes que solo la
  contienen; su mutante, una referencia con `(BOE-A-2015-10565, bloque daprimera)` y otra con
  `[art. 46.1 de la LJCA, BOE-A-1998-16718, bloque a46]`, lo pone en rojo.
- **FR-011**: La línea `⚠ NO CUBIERTO:` cita el precepto de la LPAC que lo regula aparte en la forma única y nombra la
  norma que lo rige por su nombre y su identificador, como hoy; el cuerpo que la lleva conserva la forma con sus
  marcadores. Criterio: los controles de FR-010 y FR-013.
- **FR-012**: La cobertura por materia generada da, por cada fila, la línea `⚠ NO CUBIERTO:` entera, lista para copiar,
  con los datos de la fila («Templates for output format»): la forma del cuerpo que le corresponde —con norma que lo rige
  o sin ella— con cada marcador sustituido. Criterio:
  `make skills-check` sale en rojo si la referencia no es la que genera `make skills-sync`, y un test de `make ci` sale
  en rojo si una línea generada no es la forma del cuerpo con sus marcadores sustituidos (mutante: la forma del cuerpo
  cambiada sin regenerar).
- **FR-013**: Cada forma fija que dos skills de la entrega comparten es la misma línea en las dos: para cada etiqueta
  `⚠ <ETIQUETA>:` de una línea de un bloque `text` de un `SKILL.md`, las líneas con esa etiqueta son las mismas en cada
  `SKILL.md` que la lleva. Criterio: un control nuevo de `make ci` sale en rojo, con las dos skills y la etiqueta, si
  difieren; su mutante, una palabra cambiada en la línea `⚠ SIN CONSULTA AL BOE:` de una sola skill, lo pone en rojo.
- **FR-014**: La medida de `cobertura_sin_declarar` cuenta como declarada una respuesta solo si una misma línea
  `⚠ NO CUBIERTO:` lleva la norma esperada de su eval y la cita, en la forma única, del precepto de la LPAC de su fila de
  la cobertura. Su umbral no cambia: 0 en cada modo, y decide. Criterio: un test de `make ci` sale en rojo si una
  respuesta con la norma esperada y sin la cita del precepto cuenta como declarada; en el cierre, el job sale en rojo si
  la medida pasa de 0 (SC-011).

#### La skill de normativa

- **FR-015**: El paso del cuerpo de la skill de normativa que pide `kitlegal graph check` lleva, como *gotcha* («Keep
  gotchas in `SKILL.md` where the agent reads them before encountering the situation»), que un bloque sin entrada
  `version-obsoleta` no da nada a la respuesta: ni una línea, ni una frase junto a los avisos o a la fecha de vigencia.
  La redacción la fija el plan partiendo de `e1bfb77` (`skills/boe-legislacion/SKILL.md`, líneas 215 a 220); la juzga
  el sondeo (SC-016). Criterio: la suite de aceptación sale en rojo si ese paso no contiene el texto que fija el plan.
- **FR-016**: El paso 5 de la skill de normativa cambia en dos cosas, cada una con su porqué: una remisión no leída se
  repite con las palabras del texto leído, sin materia, y sale el permiso de avisar de que una materia se regula en otra
  parte; y explicar una palabra del texto leído con lo que regula otro precepto es afirmar ese precepto. Sus ejemplos no
  nombran ninguna norma. Criterio: la suite de aceptación sale en rojo si el paso contiene «o el número, o la materia» o
  «Puedes avisar de que una materia se regula en otra parte», o si no contiene el texto que fija el plan.
- **FR-017**: Los pasos 3 y 6 y la primera viñeta del paso 4 de la skill de normativa son los de
  `skills/boe-legislacion/SKILL.md` en `c991808`, con su encabezado, salvo lo que la tabla del plan muestre que pide la
  comprobación (FR-015). Criterio: la suite de aceptación sale en rojo con una diferencia que la tabla del plan no
  recoge.
- **FR-018**: Las formas fijas de la skill de normativa —las tres de la redacción en una fecha, la de la redacción
  modificada, la de `⚠ SIN CONSULTA AL BOE:` y las etiquetas de aviso de la regla 3— siguen en su `SKILL.md`, con lo que
  cambie FR-010 y FR-013; y su tabla de comandos es la que genera `make skills-sync`. Criterio:
  `linea-sin-consulta`, `formas-de-la-redaccion-fechada`, `avisos-de-la-skill`, `hallazgos-de-la-skill` y `make
  skills-check`, en verde sobre la skill de normativa.

#### Las materias y el molde

- **FR-019**: Lo que hoy es de lo administrativo en el cuerpo y las referencias de `legal-core` —la cobertura por
  materia, los recursos y plazos, las leyes vertebrales y lo que solo pide el procedimiento común— pasa al paquete de lo
  administrativo. Sus referencias generadas siguen generadas desde los mismos datos de `data/` (`cobertura.yaml`,
  `plazos.yaml`, `normas.yaml`), con su cabecera «generado desde». Criterio: `make skills-check`; y la suite de
  aceptación sale en rojo si una fila de `cobertura_por_materia.md`, `recursos_y_plazos.md` o `leyes_vertebrales.md` de
  `c991808` no tiene sus datos en una referencia del paquete, con la cita en la forma única.
- **FR-020**: Lo que vale para cualquier materia —la jerarquía normativa, el método— queda en una skill de capacidad o en
  la de entrada, no en un paquete. Criterio: el plan dice dónde queda cada cosa; la revisión final lo coteja.
- **FR-021**: Ningún cuerpo de una skill de capacidad nombra una norma de un paquete (principio X). Una
  norma «de un paquete» es la que `data/normas.yaml` marca con la vertical de un paquete; en la entrega, el de lo
  administrativo (la vertical de ensayo vive en `testdata/`). Ningún cuerpo nombra una de ellas por su identificador, su
  título (rango, número y año) ni su abreviatura: salen los ejemplos con la Ley 39/2015 (`kitlegal boe indice
  BOE-A-2015-10565`, `art. 21 de la Ley 39/2015 [BOE-A-2015-10565, bloque a21]`) y los de la LJCA (`BOE-A-1998-16718`)
  si el plan la marca en el paquete (FR-019), y el cuerpo lleva marcadores (`<identificador>`, `<id>`, `<norma>`) o una
  norma que ningún paquete marca; las formas fijas (`⚠ PLAZO:`, `⚠ NO CUBIERTO:`) se quedan con sus marcadores, y la
  línea con la norma concreta la da la referencia generada del paquete (FR-012). Como mínimo, ninguno nombra una norma de
  las evals por materia, aunque alguna no se marque. Criterio: la suite de aceptación toma la lista de
  `data/normas.yaml` de la rama, y sale en rojo si el cuerpo de una skill de capacidad nombra una norma con vertical de
  paquete o contiene `BOE-A-1985-5392`, `BOE-A-2004-4214`,
  `BOE-A-2006-20764`, `BOE-A-2013-12887`, `BOE-A-2015-10566`, «Ley 7/1985», «Real Decreto Legislativo 2/2004»,
  «Ley 35/2006», «Ley 19/2013», «Ley 40/2015», «LRBRL», «TRLRHL», «IRPF», «LTAIBG» o «LRJSP».
- **FR-022**: El molde de vertical: un paquete de materia trae normas —marcadas con su vertical en los datos de normas
  (ADR 0044, punto 5)—, una referencia por skill de capacidad que la necesite, sus filas de cobertura y sus evals, y nada
  más. El plan fija dónde vive un paquete y cómo llega cada referencia a su skill, con `references/` plano (ADR 0046).
  Criterio: el plan y `docs/SKILLS.md` (FR-041); el test de FR-024.
- **FR-023**: El enlace a cada referencia de un paquete, con la condición de su clase («si la pregunta es de
  <materia>»), lo genera `make skills-sync` entre marcas en el `SKILL.md` de la skill que la recibe, como la tabla de
  comandos; un paquete no edita a mano ningún `SKILL.md`, ni su cuerpo ni su frontmatter. Criterio: `make skills-check`
  sale en rojo si el texto entre las marcas no es el que genera `make skills-sync`, o si una referencia de un paquete no
  está enlazada.
- **FR-024**: La vertical de ensayo vive en `testdata/` y trae dos normas, una referencia por skill de capacidad, una
  fila de cobertura y una eval. Un test de `make ci` la instala sobre una copia de `skills/` y de los datos que necesita,
  corre la sincronía y comprueba que pasa `skills-check`, que su eval se carga y que, para entrar, no escribe ningún
  `.go` ni cambia ningún `SKILL.md` fuera de lo que genera `make skills-sync`; ni `skills/` ni `data/` del repositorio
  cambian. Criterio: el test sale en rojo si alguna de esas comprobaciones falla; su mutante, un molde que pide editar un
  `SKILL.md` fuera de las marcas, lo pone en rojo.

#### Las instalaciones existentes

- **FR-025**: El binario lleva la lista de los nombres retirados y, por cada uno, las skills de la entrega que heredan
  sus preguntas. Criterio: un test de `make ci` sale en rojo si un nombre de `skills/` de `c991808` que la entrega no
  empotra falta en la lista, o si la lista nombra una skill que el binario empotra.
- **FR-026**: `kitlegal skills install`, sin `--dry-run` y en cualquier ámbito, retira cada skill que el manifiesto del
  ámbito declara con un nombre retirado y cuyas huellas coinciden —su directorio tiene exactamente los ficheros que el
  manifiesto le declara, cada uno con su huella, y cada entrada de host es la que declara—:
  quita su directorio, sus entradas de host y su entrada del manifiesto, y `data` lo dice. Sale con 0. No mira qué
  skills pide la invocación (`<nombres>` o ninguno) ni cuáles quedan en el ámbito: la lista de FR-025 solo nombra, en el
  aviso de FR-027, la skill que hereda las preguntas, y que el ámbito quede sin alguna heredera no se comprueba ni se
  avisa (quien la quiera la instala; sin nombres se instalan todas). Criterio: el e2e
  de FR-031.
- **FR-027**: Si las huellas de una skill de nombre retirado no coinciden, la instalación no toca nada de ella, conserva
  su entrada del manifiesto y lleva en `data` y en la salida legible un aviso con su ruta, que dice que su protocolo
  puede activarse junto al de la skill que la hereda y que retirarla es cosa de la persona; sale con 0 (ADR 0023). Cuando
  su directorio y sus entradas de host ya no están, la instalación siguiente quita su entrada del manifiesto y no avisa.
  Criterio: el e2e de FR-031.
- **FR-028**: Con `--dry-run`, la instalación dice lo que retiraría y lo que avisaría, y no toca nada. Criterio: el e2e
  de FR-031.
- **FR-029**: Una skill nombrada en la invocación que es un nombre retirado sale con 2 (`argumentos`), como hoy toda
  skill que el binario no empotra. Criterio: el e2e de FR-031.
- **FR-030**: Una skill que el manifiesto declara, que el binario no empotra y cuyo nombre no es un nombre retirado
  sigue como hoy (H19): no se retira, no se cambia y no se nombra. Criterio:
  `internal/app/testdata/script/h19-skills-no-empotrada.txtar` sigue en verde.
- **FR-031**: Un e2e (`testscript`) de la instalación con un nombre retirado: con las huellas que coinciden, en el ámbito
  del proyecto y en `-g`; con un fichero editado; con `--dry-run`; y nombrado en la invocación. Criterio: sale en rojo si
  alguno no da lo de FR-026 a FR-029.
- **FR-032**: Lo que nombra una skill para operar con ella usa los nombres de la entrega: las skills y sus evals, el
  plugin y el servidor MCP que el binario empaqueta, el texto de la página `/instalar/` de la web, `README.md`,
  `.github/workflows/`, `Makefile` e `internal/`. Criterio: la suite de aceptación sale en rojo si `git grep` encuentra,
  en `skills/`, `evals/`, `internal/`, `cmd/`, `.github/workflows/`, `Makefile`, `README.md` o `web/src/pages/`, un
  nombre retirado como ruta que opera con la skill —`skills/<nombre>` o `evals/<nombre>` del repositorio, o
  `.agents/skills/<nombre>` de un ámbito, sin otro segmento de ruta delante—, como código en línea solo o detrás de la
  palabra «skill». No cuentan la lista de FR-025 y sus tests, el e2e de FR-031, el texto alternativo de una captura que
  enseña la skill de hoy, los ficheros de `evals/<skill>/juez/` (`casos.yaml`, `medida.json`, `rubrica.md`), lo grabado
  bajo `internal/evals/testdata/` ni una ruta a `specs/`.

#### Las evals

- **FR-033**: Cada eval de hoy pasa a la carpeta de evals de la skill que hereda su pregunta, renombrada en git para que
  conserve su historia. Su contenido no cambia salvo los nombres de skill (`no_se_activan`, `reproduce`) y lo que pida
  FR-014 en las de cobertura. Criterio: `git diff --find-renames c991808` da cada fichero de `evals/` de `c991808` como
  renombrado o sin cambios, y `TestEvalsDelRepositorio` carga todas.
- **FR-034**: El plan da, por cada eval de hoy, la skill que la hereda, las skills que su pregunta activa y las formas
  fijas que da cada una (FR-005). Criterio: la revisión final lo coteja.
- **FR-035**: Una eval nueva de la skill de entrada, con la pregunta de la póliza de `docs/USO.md` (2026-10-21), activa
  la skill de entrada, lee la Ley 50/1980 (`BOE-A-1980-22501`) con kitlegal y la cita, y prohíbe `plazos calcular`: con
  una materia que no cubre, la sesión no pide un plazo de otra. No exige ninguna línea fija (FR-001). Sus bloques los añade una tarea `[datos]` al manifiesto de grabaciones de las evals y los
  graba el paso `grabar_datos` desde `boe.legislacion-consolidada`, fuente con fila revisada en `docs/SOURCES.md`; la
  caché de sus sesiones los sirve sin red. Criterio: `TestEvalsDelRepositorio` la carga; los tests de la caché de las
  sesiones (H26.1) salen en rojo si no sirve uno de sus bloques; pasa en el job (SC-015).
- **FR-036**: Una eval nueva de la skill de entrada con una pregunta que resuelve una sola capacidad, que activa la de
  entrada y la de esa capacidad. Criterio: pasa en el job (SC-015).
- **FR-037**: Cada skill de capacidad tiene al menos una eval de otra skill que la declara en `no_se_activan`, con una
  pregunta de otra capacidad. Las reglas del conjunto de evals de cada skill siguen a la skill que hereda sus evals, y la
  de entrada tiene las suyas, con las dos de FR-035 y FR-036. Criterio: `TestEvalsDelRepositorio` sale en rojo si una
  skill de capacidad no está en el `no_se_activan` de ninguna eval de otra skill; pasan en el job (SC-015).
- **FR-038**: Ningún umbral cambia: los `clases.yaml` de los jueces van con la skill que hereda las evals que juzgan, con
  los mismos valores; cada eval que hoy juzga un juez la sigue juzgando, con las mismas clases; y los umbrales del informe
  son los de hoy para las evals de cada skill. El plan dice cómo siguen la serie de cierres y la reutilización de las
  medidas de `main` con un nombre nuevo (ADR 0045), y la elección de las evals del job elige las de una skill cuando
  cambia un fichero que la skill carga, también la referencia de un paquete que enlaza. Criterio: los tests de los
  umbrales del informe y `TestLimiteDeLaTasa` en `make ci`, con los nombres nuevos y los valores de hoy, salen en rojo
  con un umbral distinto; `TestEleccionDeLosMutantes` sale en rojo si tocar una referencia de un paquete no elige las
  evals de la skill que la enlaza.
- **FR-039**: La matriz del job de evals tiene todas las skills de la entrega y ningún nombre retirado, con como mucho
  once sesiones a la vez entre todos sus trabajos (H30). Antes de cerrarse, el plan recalcula el peor caso de cada
  trabajo con la suite completa, que es lo que el job mide con skills nuevas y renombradas, y su tope (`timeout-minutes` de
  `.github/workflows/evals.yml`). Criterio: `TestDefinicionDelJob` sale en rojo en `make ci` si falta una skill de la
  entrega, si está un nombre retirado, si la matriz admite más de once sesiones o si el tope de un trabajo es menor que
  su peor caso.
- **FR-040**: La entrega retira los tres guiones `internal/app/testdata/script/h31-1-h31-1-*.txtar`, que fijan palabra por
  palabra el cuerpo y las referencias de `boe-legislacion` frente a `brazo-W`, y su suite de aceptación lleva lo que
  sigue valiendo: las formas fijas (FR-018), la tabla de comandos (FR-018), los pasos 3 y 6 y la primera viñeta del paso
  4 (FR-017) y ningún cuerpo de una skill de capacidad que nombre una norma de las evals por materia (FR-021).
  `h19-skills-invocacion` sigue exigiendo `PATH` en cada `SKILL.md`, con los nombres nuevos. Criterio: `make ci`.

#### La guía, el ADR y el cierre

- **FR-041**: `docs/SKILLS.md` dice el reparto —qué skills hay, qué capacidad cubre cada una y cuál es la de entrada—, el
  molde de un paquete (FR-022 a FR-024) y lo medido: los tres defectos, con sus cifras y sus fuentes, tal como las da el
  Contexto; y la tabla de controles de §1, con los nombres nuevos y los controles de FR-010 y FR-013. Las prácticas de
  las guías de autoría se citan como las atribuye el hito, con los enlaces que ya registra `docs/SKILLS.md`. Criterio:
  la revisión final coteja las cifras con el hito y que cada URL que la rama añade está en un fichero de `main`.
- **FR-042**: Un ADR nuevo registra los nombres de las skills por capacidad y de la de entrada, y el reparto (ADR 0044,
  punto 7). Criterio: el fichero en `docs/ADR/`, con la tabla de nombres de `c991808` a la entrega.
- **FR-043**: `CHANGELOG.md` lleva en *Unreleased* la entrada del hito: la skill de entrada, las skills por capacidad y
  sus nombres, el paquete de lo administrativo, la forma única de cita, los dos arreglos de la skill de normativa y la
  retirada de las skills con nombre retirado al instalar. Criterio: la entrada está en *Unreleased*.

### Key Entities

- **Skill de entrada**: `skills/<nombre>/`; método, capacidades y cobertura; ninguna regla que otra skill necesite.
- **Skill de capacidad**: `skills/<nombre>/` con su protocolo, invariantes, formas fijas, tabla de comandos y los enlaces
  generados a las referencias de los paquetes.
- **Paquete de materia**: normas marcadas con su vertical, referencias por skill de capacidad, filas de cobertura y
  evals; el de lo administrativo y la vertical de ensayo.
- **Línea `⚠ NO CUBIERTO:` de una fila**: la forma del cuerpo con los datos de una fila de la cobertura; una por fila.
- **Nombre retirado**: un nombre de skill de `c991808` que la entrega no empotra, con las skills que heredan sus
  preguntas.
- **Manifiesto de la instalación** (`kitlegal.json`, H19): por skill declarada, sus ficheros con su huella y sus
  entradas de host; lo que decide si una skill de nombre retirado se retira o se avisa.

## Uso, de fuera adentro

- **El cuerpo de la skill de entrada (FR-001 a FR-004)**: lo carga el agente entero una vez por pregunta que la activa,
  como mucho 11 000 caracteres, además del cuerpo de cada skill de capacidad que la pregunta active, cada uno con como
  mucho 11 000 (FR-009). No crece con el uso: lo escribe una persona y el gate lo acota. Lo gana cada respuesta: la
  pregunta llega a la capacidad que la resuelve y lo que no cubre la base se dice (US1).
- **Los enlaces generados a las referencias de los paquetes (FR-023)**: una línea por paquete en cada `SKILL.md` que
  recibe una referencia suya; con el paquete de lo administrativo, una por skill de capacidad que la necesita. Crecen
  con los paquetes que se escriben, no con las normas consultadas, y cuentan en los 11 000 caracteres del cuerpo: el
  paquete que no cabe pone en rojo `make ci` (Edge Cases).
- **Las referencias de un paquete (FR-019)**: las lee el agente solo si la pregunta es de su materia, cuando el enlace de
  su skill lo pide, una vez por pregunta. Medido en esta sesión con `wc -m`, las de `legal-core` que pasan al paquete de
  lo administrativo tienen hoy 1 448 (`cobertura_por_materia.md`), 1 364 (`recursos_y_plazos.md`) y 2 585
  (`leyes_vertebrales.md`) caracteres; la cobertura crece con la línea entera de cada una de sus 9 filas (FR-012), unos
  250 caracteres cada una. Crecen con las filas que una persona escribe en `data/`, nunca con lo consultado: cientos de
  normas y miles de bloques leídos no añaden ninguna fila.
- **La línea `⚠ NO CUBIERTO:` (FR-011, FR-012)**: la escribe la respuesta una vez por pregunta cuyo procedimiento está
  en una fila de la cobertura, copiada de la referencia; una línea. Deja de darse para un procedimiento cuando su fila
  sale de la cobertura porque una materia o H33 lo cubren; no se repite sin información nueva, porque solo va en la
  respuesta a la pregunta que cae en su fila.
- **La comprobación de la skill de normativa (FR-015)**: `kitlegal graph check` se pide una vez por bloque leído, detrás
  de su lectura, y devuelve como mucho 50 hallazgos, con cualquier volumen de lo consultado. Sin `version-obsoleta`, no
  da nada a la respuesta; con ella, una línea `⚠ REDACCIÓN MODIFICADA:` por bloque, que deja de darse al leerlo de nuevo.
  Ninguna señal nueva: el hito quita frases de la respuesta.
- **Lo que la instalación dice de un nombre retirado (FR-026 a FR-028)**: una entrada en `data` y una línea en la salida
  legible por skill de nombre retirado que el manifiesto declara, cada una con una ruta: una por nombre retirado, sea
  cual sea lo consultado. La retirada se dice una vez: el manifiesto deja de declararla. El aviso se repite
  en cada instalación mientras sus ficheros no tengan la huella del manifiesto, y deja de darse cuando la persona quita
  su directorio y sus entradas de host (FR-027).
- **`docs/SKILLS.md` y el ADR (FR-041, FR-042)**: los lee quien escribe una skill o un paquete, una vez por skill o por
  paquete; dejan de valer cuando otro ADR cambie el reparto.
- **La entrada de `CHANGELOG.md` (FR-043)**: la lee quien actualiza kitlegal, una vez por versión.

## Success Criteria *(mandatory)*

### Measurable Outcomes

- **SC-001**: `make ci` en verde, con `make skills-check`, `ordenes-de-ejemplo`, `normas-nombradas` y los demás controles
  de contenido de `docs/SKILLS.md` §1, sobre las skills nuevas: cada control que hoy se aplica a `boe-legislacion` se
  aplica a la skill de normativa, el que se aplica a `legal-core`, a cada skill que hereda sus preguntas, y el que se
  aplica a todas, a todas las de la entrega. Lo hace cumplir el propio `make ci`, que sale en rojo con cualquiera de
  ellos.
- **SC-002**: Cada cuerpo de la entrega tiene como mucho 11 000 caracteres; la comprobación de las skills de `make ci` y
  `make skills-sync` salen en rojo con 11 001 o más (FR-009).
- **SC-003**: El control nuevo de una sola forma de cita (FR-010) sale en verde sobre todos los `SKILL.md` y todas las
  referencias de la entrega, y en rojo con su mutante, en `make ci`.
- **SC-004**: El control nuevo de las formas compartidas (FR-013) sale en verde sobre la entrega y en rojo con su
  mutante, en `make ci`.
- **SC-005**: El test de la vertical de ensayo (FR-024) sale en verde en `make ci`, y en rojo con su mutante.
- **SC-006**: El e2e de la instalación con un nombre retirado (FR-031) sale en verde en `make ci`, y en rojo si la
  instalación no retira, retira lo editado o no avisa.
- **SC-007**: Cada descripción de la entrega tiene como mucho 900 caracteres —la validación del frontmatter de `make ci`
  sale en rojo con 901— y la juzga en tercera persona el juez de las descripciones del job (ADR 0046), con su umbral de
  0, que decide: una descripción marcada pone el veredicto en fallo y el job en rojo.
- **SC-008**: El tope de cada trabajo de evals, recalculado antes de cerrar el plan, cubre su peor caso;
  `TestDefinicionDelJob` sale en rojo si no lo cubre (FR-039).
- **SC-009**: `CHANGELOG.md` (*Unreleased*) (FR-043); y el job de evals en la propuesta de cambio, con todas las skills,
  que el workflow mide en el cierre sobre la cabeza de la rama.
- **SC-010**: En el informe del job de cierre, `afirma_lo_no_leido:claude-sonnet-5-5` se cumple con el umbral de hoy,
  0.028, en la skill de normativa, la que hereda `evals/boe-legislacion/juez/clases.yaml`; en la que hereda
  `evals/jurisprudencia/juez/clases.yaml` sigue en 0.017 (SC-013, FR-038). El umbral decide: si la tasa pasa del límite
  del ADR 0045, el veredicto es fallo y el job sale en rojo.
- **SC-011**: En el informe del job de cierre, `cobertura_sin_declarar` es 0 en los dos modos, con la medida de FR-014;
  decide: con una respuesta sin declarar, el veredicto es fallo y el job sale en rojo.
- **SC-012**: En el informe del job de cierre, ninguna sesión de `claude-sonnet-5-5` de las evals 14, 15 y 17 tiene en
  `fuera_de_lo_grabado` un bloque de sus `lecturas` (umbral: 0 sesiones), contado como lo contaba el spec de H31.1. Lo hacen
  cumplir `TestCacheDeLasSesiones` y `TestLecturaDeLaEvalSinGrabar` (`internal/evals/preparar_test.go`), que salen en
  rojo en `make ci` si la caché de la skill que hereda esas evals no sirve un bloque de sus `lecturas` sin fecha y con
  `--fecha 9999-12-31`, y la preparación de la sesión, que da la falta y no la abre. `fuera_de_lo_grabado` sigue
  publicándose sin decidir.
- **SC-013**: Los demás umbrales que deciden del informe del job de cierre se cumplen en todas las skills; el que no se
  cumple pone su veredicto en fallo y el job en rojo.
- **SC-014**: `cuenta_su_proceso` se publica por modo, sin cifra de aceptación en el cierre: no decide.
- **SC-015**: Las evals de activación nuevas (FR-035 a FR-037) pasan, y las de hoy siguen pasando en la skill que hereda
  su pregunta: cada eval que decide pasa con la regla de su serie (ADR 0016), y una que no pasa pone el veredicto de su
  skill en fallo y el job en rojo.
- **SC-016**: Lo que la entrada deja a una persona antes de fusionar: un sondeo con la skill de la propuesta, como los de
  `evidencias/sondeos-20261028/` —el mismo arnés, `claude-sonnet-5-5`, Claude Code 2.1.289, concurrencia 5—, de 280
  sesiones, que no se lanza con un job de evals en marcha, con los límites del ADR 0047:

  | Parte | Skill y evals | Respuestas | Pasa con | A la tasa objetivo | A una peor |
  |---|---|---|---|---|---|
  | `afirma_lo_no_leido` | normativa, eval 15 | 80, 40 por modo | como mucho 6 | 5 %: 89 % | 12,5 %: 11 % |
  | `cuenta_su_proceso` | normativa, eval 15 | las mismas 80 | como mucho 8 | 5 %: 98 % | 12,5 %: 32 % |
  | `cuenta_su_proceso` | normativa, eval 05 | 80, 40 por modo | como mucho 6 | 5 %: 89 % | 15 %: 3 % |
  | la línea `⚠ NO CUBIERTO:` con la norma esperada y la cita del precepto | la que la lleve, evals 07, 08 y 12 | 120, 20 por eval y modo | como mucho 1 sin ella | 0,5 %: 88 % | 2,8 %: 15 % |

  Calculadas en esta sesión con la binomial, las probabilidades de la tabla dan 0,895, 0,114, 0,982, 0,318, 0,895,
  0,035, 0,878 y 0,148; y con Fisher unilateral, 6 de 80 frente a 21 de 120 da p = 0,032, 6 de 80 frente a 16 de 80,
  p = 0,019, y 8 de 80 frente a 17 de 80, p = 0,040. Lo mide una persona fuera del run (ADR 0032): el informe final lo
  da como pendiente. Si alguna parte no se cumple, la propuesta no se fusiona y se decide con sus cifras y sus respuestas
  marcadas; el sondeo no se repite para ver si sale en verde.
- **SC-017**: La suite de aceptación sale en rojo si no se cumplen FR-004, FR-005, FR-006, FR-007, FR-015, FR-016,
  FR-017, FR-019 o FR-021, cada uno con su criterio.
- **SC-018**: Nada de lo que busca FR-032 nombra una skill retirada; la suite de aceptación lo comprueba con `git grep`,
  con las excepciones de FR-032, y sale en rojo con una aparición que cuenta.

### Criterios del hito, tal como los escribe

**Controles** (SC-001, SC-003 a SC-009): «`make ci` en verde, con `make skills-check`, `ordenes-de-ejemplo`,
`normas-nombradas` y los demás controles de contenido de `docs/SKILLS.md` §1, sobre las skills nuevas; un control nuevo
de una sola forma de cita, con su mutante en rojo: ningún `SKILL.md` ni ninguna referencia de ninguna skill cita un
bloque con otra forma que `[<identificador>, bloque <id>]`, y cada forma fija que dos skills comparten es la misma línea
en las dos; el test de la vertical de ensayo y el e2e de la instalación con un nombre retirado; las descripciones nuevas,
juzgadas en tercera persona en el job (ADR 0046), de como mucho 900 caracteres; el tope de los trabajos de evals,
recalculado; `CHANGELOG.md` (*Unreleased*); y el job de evals en la propuesta de cambio, con todas las skills.»

**Aceptación** (SC-010 a SC-016): «en el informe del job de cierre, los umbrales que deciden se cumplen en todas las
skills, con los de hoy: `afirma_lo_no_leido:claude-sonnet-5-5` como mucho 0.028, y `cobertura_sin_declarar` 0 en los dos
modos, con la medida que exige también la cita del precepto; ninguna sesión de las evals 14, 15 y 17 tiene en
`fuera_de_lo_grabado` un bloque de sus `lecturas`. `cuenta_su_proceso` se publica por modo, sin cifra de aceptación en el
cierre; las evals de activación nuevas pasan, y las de hoy siguen pasando en la skill que hereda su pregunta; lo que la
entrada deja a una persona antes de fusionar: un sondeo con la skill de la propuesta, como los de
`evidencias/sondeos-20261028/` —el mismo arnés, `claude-sonnet-5-5`, Claude Code 2.1.289, concurrencia 5—, con el límite
de cada parte calculado como pide el ADR 0047: el objetivo, el número de respuestas y la probabilidad de pasar a la tasa
objetivo y a una claramente peor. Son 280 sesiones, unas dos horas, y no se lanza con un job de evals en marcha. Si
alguna parte no se cumple, la propuesta no se fusiona y se decide con sus cifras y sus respuestas marcadas; el sondeo no
se repite para ver si sale en verde.» Con la tabla de SC-016 y sus notas:

- «**La eval 15** repite el criterio del ADR 0047, con Fisher unilateral p < 0,05 frente al control de
  `evidencias/sondeos-20261025/`, 16 y 17 de 80. Protege que el reparto no pierda lo que ganó H31.1: con 80 respuestas
  no se distingue el 2,5 % de `brazo-P80` de una tasa menor, y el arreglo del paso 5 lo deciden las marcas, no la
  cifra.»
- «**La eval 05** es la que concentra la línea de la comprobación en el job. El objetivo, el 5 %, es la tasa de la eval
  15 en `brazo-P80` y está por encima de la de antes de H31; la peor, el 15 %, está por debajo de la del job de
  `cd21c7b`, 21 de 120 (17,5 %). No tiene control en un sondeo: el límite es absoluto, y 6 de 80 frente a 21 de 120 da
  p = 0,03 (Fisher unilateral).»
- «**Las evals 07, 08 y 12** miden la línea sin el juez: el sondeo la juzga sin la traza, con el motivo de la
  declaración ausente. El objetivo, el 0,5 %, es el de una forma que la sesión copia de la referencia; la peor, el
  2,8 %, la de hoy, 1 de 36.»

«Las probabilidades salen de la binomial, con la misma función que `evidencias/sondeos-20261026/guiones/analizar.py`.»

## Assumptions

- Los nombres de las skills no los fija el spec: los fija el plan y los registra el ADR (FR-007, FR-042). El spec las
  nombra por su capacidad.
- La serie de cierres de una skill nueva o renombrada empieza de cero, porque su clave es el árbol de su carpeta en
  `skills/` (ADR 0038), y `scripts/workflow/serie.sh` no cambia: es la consecuencia de cambiar el árbol, no algo que haga
  el hito.
- Las sesiones del run no tienen red fuera de las herramientas de Go: lo que `docs/SKILLS.md` diga de las guías de
  autoría es lo que el hito les atribuye, con los enlaces que el repositorio ya registra (§1 de `docs/SKILLS.md`).
- Si `grabar_datos` no puede grabar un bloque de la Ley 50/1980 porque la fuente no responde, es una causa mayor del run
  (dependencia externa inaccesible), no un caso del producto.
- El sondeo de SC-016 lo prepara y lo lanza la persona con el arnés de `evidencias/sondeos-20261028/`; si sus guiones
  nombran una skill por su nombre de hoy, los adapta ella: el run no escribe en `evidencias/`.
- `CLAUDE.md`, `CONTRIBUTING.md` y los ADR anteriores nombran las skills de hoy; actualizarlos lo decide una persona y
  el informe final lo da como pendiente.
- La versión de cada skill solo se registra en `CHANGELOG.md`.

## Fuera de alcance

- Cualquier umbral del job, la rúbrica y los casos de los jueces, y que `cuenta_su_proceso` decida en el job: subir un
  umbral pide un ADR, y hacer decidir una clase que solo se publica, otro (ADR 0037 y 0038).
- Que `kitlegal boe articulo` dé la comprobación en su propio sobre, en lugar de que la sesión pida `graph check`.
- Que `kitlegal boe` devuelva la rúbrica o el id de los preceptos remitidos (backlog; `evidencias/adr-0038/`).
- Las demás marcas de `cuenta_su_proceso` del job, que no son de la comprobación: las dos de la eval 23, la de «no hace
  falta leer nada más» y la de «el BOE no dice aquí qué norma la sustituye».
- La cobertura en positivo y la línea `⚠ NO CUBIERTO:` única de todas las materias (H33), el motor de cómputo (H36), los
  ejes de territorio y los destinatarios (H37), los documentos (H35) y la actualización de las skills instaladas con el
  binario (H38).
- Las skills de documentos y destinatarios, y los packs (`packs/`), que llegan con la segunda vertical (ADR 0044, punto 5).
- Un juez con modelo de la skill de entrada, y juzgar con modelo que la respuesta a la póliza no da órganos ni vías de
  memoria: un juez solo decide medido contra casos etiquetados (ADR 0037), y la rúbrica y los casos quedan fuera. Lo que
  tiene forma lo comprueba la eval sin modelo (FR-035).
- Un paquete de otra materia que lo administrativo, la migración de `boe-fiscal` y cualquier vertical: la de ensayo solo
  vive en `testdata/` (FR-024).
- Que la skill de plazos cuente otros plazos que los de hoy (H36); cambios en los applets `boe`, `graph`, `cita`,
  `plazos` y `territorio` y en `internal/source`. Del applet `skills`, solo lo de FR-025 a FR-031.
- `kitlegal skills doctor` y `kitlegal skills list` con un nombre retirado: siguen como hoy.
- `CLAUDE.md`, `CONTRIBUTING.md`, la constitución, `docs/ROADMAP.md` y los ADR anteriores.
- Una línea `⚠ NO CUBIERTO:` sin precepto de la LPAC, o una etiqueta fija nueva, para lo que la base no cubre (la póliza,
  los documentos, los destinatarios): lo dice la skill de entrada en prosa hasta que H33 dé la forma única.
- Un aviso, en `data` o en la salida legible, de que el ámbito se queda sin alguna skill que hereda las preguntas de una
  retirada, y condicionar la retirada a que el ámbito las tenga todas.
- Las capturas de la página `/instalar/` que enseñan una skill de hoy: las rehace una persona; la entrega cambia el texto
  de la página.
- Cambios en `scripts/workflow/`, `.specify/` e `internal/workflow/`, y en `scripts/workflow/serie.sh`.
- Cambiar el tope de una sesión de las evals (`topeDeUnaSesion`, 240 s).
- Lanzar el sondeo de SC-016 y versionar su evidencia: los hace una persona fuera del run.
- La release y su etiqueta, y la fusión: las decide una persona.
