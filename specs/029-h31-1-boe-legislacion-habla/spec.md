# Feature Specification: H31.1 · `boe-legislacion` habla de la norma, no de la consulta, y lee lo que nombra

**Feature Branch**: `029-h31-1-boe-legislacion-habla`

**Created**: 2026-10-27

**Status**: Draft

**Input**: el fichero del hito, `docs/roadmap/in-progress/H31.1-boe-legislacion-habla-de-la-norma-y-lee-lo-que-nombra.md`,
la única entrada humana del run (ADR 0018), con su evidencia en `evidencias/sondeos-20261025/` y
`evidencias/sondeos-20261026/`, que están en `main`: el run las usa, no las escribe. Es el relanzamiento de H31.1: la
rama del primer run (`028-h31-1-boe-legislacion-habla`, la #200, cabeza `47c3c24`) se conserva sin fusionar porque la
evidencia remite a ella.

## Contexto

El hito no entrega skill nueva: protege `boe-legislacion` (principio VIII) de dos defectos medidos antes de que H32
reorganice las skills sobre ella.

- **`cuenta_su_proceso`** (respuestas que cuentan lo que hizo el agente): en el job, del 1,8 % y el 2,8 % antes de H31 al
  26,5 % y el 34,0 % desde H31; en el sondeo de la eval 15, 17 de 80 con la skill de `main` (el control,
  `evidencias/sondeos-20261025/`).
- **`afirma_lo_no_leido`** en las evals por materia: 16 de 80 en el control; casi todas las respuestas de la eval 15
  nombran el art. 7 de la Ley 35/2006, al que remite el art. 17 y que ninguna sesión lee.
- **El primer run no bastó** (`evidencias/sondeos-20261026/LEEME.md`): la skill de la #200 (`brazo-H`) da 11 de 40 y 10
  de 40. La misma con el paso 6, «Repasar antes de enviar», y un paso 3 que lee por defecto lo remitido, con su porqué
  (`brazo-W`), da 1 de 40 en `afirma_lo_no_leido` y 2 de 40 en `cuenta_su_proceso`; sin las prohibiciones concretas del
  paso 5 (`brazo-X`), vuelve a 6 de 40.

**La skill medida** es `evidencias/sondeos-20261026/brazo-W/skill/` (`SKILL.md` y `references/`): la de `47c3c24` con
`parches/brazo-W.diff`. Medido en esta sesión con `wc -m` sobre el cuerpo (desde la línea que sigue al `---` que cierra
el frontmatter): el suyo tiene **12 081** caracteres y el de `main`, 10 992; el máximo es 11 000 (ADR 0046;
`maximoDeCaracteresDelCuerpo` en `internal/skills/sincronia.go`). Su frontmatter es igual al de `main`, y de sus
referencias solo difieren de las de `main` `buscar_la_norma` y `redaccion_modificada`, lo que cambió la #200.

Nombres fijados para que cada requisito tenga una sola lectura. **Los pasajes medidos** son, en el `SKILL.md` de la skill
medida, con su encabezado: el paso 3, «Completar» (líneas 48 a 59); la primera viñeta del paso 4, «Responder», la que
empieza por «**La respuesta empieza por lo que se pregunta…**» (líneas 63 a 68); el paso 5, «De un precepto no leído,
nada» (líneas 78 a 86); y el paso 6, «Repasar antes de enviar» (líneas 88 a 100). **La regla de leer lo que se nombra**
es la del paso 3 y la del paso 6, punto 1. **Un texto de una clase de pregunta** es el que solo se aplica cuando se da
una condición de la pregunta o de una lectura —una fecha, una orden que no termina con `0`, las herramientas en vez de
las órdenes, varios bloques a la vez, una norma fuera de la tabla, una entrada `version-obsoleta`, una pregunta por lo
que cambió…—; lo demás es **lo que toda respuesta necesita**.

## Clarifications

### Session 2026-10-27

- Q: ¿Qué control hace cumplir SC-008 (0 sesiones de `claude-sonnet-5-5` de las evals 14, 15 y 17 con un bloque de sus `lecturas` en `fuera_de_lo_grabado`), dado que `fuera_de_lo_grabado:<modelo>:<modo>` se publica y no decide nunca? → A: SC-008 sigue con umbral 0 y lo hacen cumplir `TestCacheDeLasSesiones` y `TestLecturaDeLaEvalSinGrabar` (`internal/evals/preparar_test.go`) con los nueve bloques de FR-011, más la preparación de la sesión, que da la falta y no la abre si un bloque no está grabado; el hito no toca el arnés y `fuera_de_lo_grabado` sigue con `decide: false`. Una lectura con otra fecha no llega al apartado: `--fecha 9999-12-31` deja guardada la entrada de las versiones del bloque, la caché preparada sirve su lectura en cualquier otra fecha (research D18 de H20) y esa lectura sale con 0 o 3, no con 4 o 5 (auto: criterio a; fuente: hito H31.1 «Controles» y «Fuera de alcance»; `docs/roadmap/done/H26.1-las-evals-sirven.md`, punto 5; `specs/024-h20-boe-articulo-fecha/research.md`, D18; `internal/evals/juzgar.go`, `FueraDeLoGrabado`)
- Q: ¿Un texto de una clase de pregunta que sale del cuerpo a una referencia (FR-007) llega a ella con las palabras de la skill medida, o se puede reescribir al moverlo con los criterios de FR-006? → A: Llega palabra por palabra, como en el `SKILL.md` de la skill medida; el diff de la referencia solo añade ese texto, la frase que lo presenta y, si procede, el cambio de la remisión. Reescribir más corto (FR-006) vale para lo que se queda en el cuerpo, no para lo que sale; un texto que no se entiende en la referencia sin cambiarlo no sale, se queda y se reescribe allí si no es un pasaje medido, una forma fija ni la tabla (auto: criterio a; fuente: hito H31.1 «El sitio en el cuerpo»; spec FR-001, FR-006, FR-007 y «Fuera de alcance»)
- Q: Cuando un texto de una clase de pregunta solo lo necesita un pasaje medido o una parte del cuerpo que no es un paso numerado, ¿dónde va el enlace con la condición de leer su referencia que exige FR-007? → A: Solo en un paso numerado del protocolo que no sea un pasaje medido (el paso 1, el 2 o el paso 4 fuera de su primera viñeta), en la frase que dice cuándo se da la clase. Un texto cuya única necesidad está en un pasaje medido o fuera de los pasos numerados («Protocolo», «Cómo se cita», «Redacción en una fecha», «Redacción modificada», «Reglas») no sale del cuerpo. Los enlaces que la skill medida ya tiene fuera de los pasos (`con_herramientas`, `como_se_cita`, `redaccion_modificada`) siguen, y un texto que sale a una de esas referencias lleva además su enlace y condición en un paso numerado no medido. «El paso que la pide» es el que enlaza la referencia con la condición de su clase, no con un «antes de responder» que vale para toda pregunta (auto: criterio a; fuente: hito H31.1 «El sitio en el cuerpo», punto 1; spec FR-002, FR-006, FR-007 y FR-014; `evidencias/sondeos-20261025/LEEME.md`)

## User Scenarios & Testing *(mandatory)*

### User Story 1 - La respuesta empieza por lo que se pregunta y habla de la norma (Priority: P1)

Quien pregunta por una norma lee solo la respuesta: no ve las órdenes, las llamadas ni lo que devuelven. Una respuesta
que cuenta la consulta («la comprobación no detectó cambios…») le habla de algo que no ve. La skill lleva la primera
viñeta del paso 4 y el paso 6 de la skill medida, que con ella bajaron `cuenta_su_proceso` de 17 de 80 a 2 de 40.

**Why this priority**: es el defecto que introdujo H31, y el texto que lo arregla está medido.

**Independent Test**: se comparan los pasajes medidos del cuerpo de v0.3.1 con los de la skill medida (FR-002); y lo
mide una persona con el sondeo de la eval 15 (SC-012).

**Acceptance Scenarios**:

1. **Dado** el `SKILL.md` de v0.3.1, **Cuando** se compara la primera viñeta de su paso 4 y su paso 6 con las líneas 63 a
   68 y 88 a 100 del `SKILL.md` de la skill medida, **Entonces** son idénticas.
2. **Dado** el sondeo de la eval 15 que lanza una persona con la skill de la propuesta (40 respuestas), **Cuando** se
   cuentan las marcas de `cuenta_su_proceso`, **Entonces** son como mucho 4.

---

### User Story 2 - Lo que la respuesta nombra, se ha leído antes o se nombra con las palabras leídas (Priority: P1)

Si un bloque leído remite a otro precepto, o su id solo salió en un índice, y la respuesta va a hablar de él, el agente
lo lee antes de escribirla; es un salto, y lo que remite esa segunda lectura no se lee y se nombra solo con las palabras
del texto leído. Antes de enviar, repasa cada cosa que dice de un precepto. De lo no leído, número y materia no van
juntos.

**Why this priority**: con la tasa de hoy, `afirma_lo_no_leido` pone en rojo su umbral (0.028) más de 1 de cada 50
ejecuciones completas; la skill medida lo baja a 1 de 40.

**Independent Test**: se comparan los pasos 3, 5 y 6 del cuerpo de v0.3.1 con los de la skill medida (FR-002); lo mide el
cierre (SC-007) y, con más respuestas, el sondeo de una persona (SC-012).

**Acceptance Scenarios**:

1. **Dado** el `SKILL.md` de v0.3.1, **Cuando** se comparan sus pasos 3 y 5 con las líneas 48 a 59 y 78 a 86 del
   `SKILL.md` de la skill medida, **Entonces** son idénticos, y ninguna referencia es la única que dice la regla de leer
   lo que se nombra.
2. **Dado** el job de cierre de `boe-legislacion`, **Cuando** el juez mide las respuestas de `claude-sonnet-5-5`,
   **Entonces** `afirma_lo_no_leido:claude-sonnet-5-5` se cumple con 0.028.
3. **Dado** el sondeo de la eval 15 de una persona con la skill de la propuesta, **Cuando** se cuentan las marcas de
   `afirma_lo_no_leido`, **Entonces** son como mucho 2 de 40.

---

### User Story 3 - El cuerpo cabe sin perder lo que toda respuesta necesita (Priority: P1)

El cuerpo de la skill medida pasa en 1 081 caracteres del máximo. El sitio sale del resto del cuerpo: lo que solo pide una
clase de pregunta va a la referencia de esa clase, con el paso del protocolo que dice cuándo leerla, porque una
referencia que ningún paso pide no se lee; lo demás se dice más corto. Los pasajes medidos, las formas fijas y la tabla
de comandos no se tocan.

**Why this priority**: sin el sitio, la skill medida no pasa `make ci`; y si lo que toda respuesta necesita se va a una
referencia, el agente no lo lee (`evidencias/sondeos-20261025/`).

**Independent Test**: `make ci` (FR-005, FR-008) y la tabla de textos del plan, que coteja la revisión final (FR-006,
FR-007).

**Acceptance Scenarios**:

1. **Dado** el cuerpo de v0.3.1, **Cuando** se cuenta con `wc -m`, **Entonces** tiene como mucho 11 000 caracteres.
2. **Dado** un texto de la skill medida que la entrega saca del cuerpo, **Cuando** se busca en `references/`, **Entonces**
   está en la referencia de su clase, y el paso del protocolo que lo necesita la enlaza y dice cuándo leerla.
3. **Dado** un texto del cuerpo que la entrega reescribe, **Cuando** se coteja la tabla del plan, **Entonces** da el texto
   de la skill medida y el de la entrega, y lo que dice el primero que el agente haría mal sin ello sigue en el segundo.

---

### User Story 4 - Las evals por materia sirven lo que sus respuestas nombran (Priority: P1)

Las evals 14, 15 y 17 llevan en `lecturas` los preceptos que sus respuestas nombran sin haberlos leído, para que la regla
de leer lo que se nombra se pueda cumplir en las sesiones sin red; sin ellos, la lectura sale con 5.

**Why this priority**: sin las grabaciones, cumplir la regla pone la lectura fuera de lo grabado, y el cierre mide la
caché, no la skill.

**Independent Test**: los tests de `lecturas` de H26.1 con los bloques nuevos, en `make ci` (SC-004).

**Acceptance Scenarios**:

1. **Dado** el manifiesto `internal/evals/testdata/grabaciones.json` y sus grabaciones tras la tarea `[datos]`, **Cuando**
   se prepara la caché de una sesión de la eval 15, **Entonces** `kitlegal boe articulo BOE-A-2006-20764 a7 --offline`
   sale con 0, sin fecha y con `--fecha 9999-12-31`, y lo mismo `a19`, `a20` y `dadecimoctava`.
2. **Dado** lo mismo, **Cuando** se prepara la caché de una sesión de la eval 14, **Entonces** salen con 0, sin fecha y con
   `--fecha 9999-12-31`, `a60`, `a92`, `a100` y `a104` de `BOE-A-2004-4214`; y de la eval 17, `a24` de
   `BOE-A-2013-12887`.
3. **Dado** un bloque de `lecturas` sin grabar, **Cuando** se prepara la sesión, **Entonces** la preparación da la falta y
   no la abre, como hoy (H26.1).

---

### User Story 5 - El cierre mide la skill con los umbrales de hoy (Priority: P2)

El job de evals de la propuesta de cambio mide las tres skills; los umbrales que deciden, con sus valores de hoy, dan el
veredicto.

**Why this priority**: es la aceptación sin persona del hito; las cifras finas las da el sondeo (US7).

**Independent Test**: el informe del job de cierre sobre la cabeza de la rama (SC-007 a SC-011).

**Acceptance Scenarios**:

1. **Dado** el informe del job de cierre de `boe-legislacion`, **Cuando** se leen sus umbrales, **Entonces**
   `afirma_lo_no_leido:claude-sonnet-5-5` se cumple con 0.028, los demás umbrales que deciden se cumplen, y
   `cuenta_su_proceso` se publica por modo, sin decidir.
2. **Dado** ese informe, **Cuando** se leen las sesiones de `claude-sonnet-5-5` de las evals 14, 15 y 17, **Entonces**
   ninguna tiene en `fuera_de_lo_grabado` un bloque de las `lecturas` de su eval, contado como dice SC-008.
3. **Dado** los informes del job de cierre de `jurisprudencia` y `legal-core`, **Cuando** se lee su veredicto, **Entonces**
   es aprobado.
4. **Dado** la definición del job, **Cuando** corre `TestDefinicionDelJob` en `make ci`, **Entonces** el tope del trabajo de
   `boe-legislacion` cubre su peor caso, recalculado por el plan.

---

### User Story 6 - La guía dice qué referencias se leen (Priority: P2)

Quien escribe o reparte una skill —H32 la primera— sabe por `docs/SKILLS.md` §1 que las sesiones abren las referencias
de las que necesitan un dato para seguir y no las que solo traen guía, aunque el cuerpo mande leerlas, y qué prácticas de
autoría lo respaldan.

**Why this priority**: H32 lo necesita escrito antes de repartir skills y referencias.

**Independent Test**: se lee `docs/SKILLS.md` §1 (FR-014).

**Acceptance Scenarios**:

1. **Dado** `docs/SKILLS.md` §1 de la rama, **Cuando** se lee, **Entonces** dice lo medido, con sus cifras y la ruta
   `evidencias/sondeos-20261025/`; sus dos consecuencias —lo que toda respuesta necesita va en el cuerpo, y una
   referencia nueva lleva el paso que la pide—; y las tres prácticas de las guías de autoría de agentskills.io y de
   Anthropic, con `https://agentskills.io/` y el enlace a «Las dos fuentes» (FR-014).

---

### User Story 7 - Una persona decide la fusión con el sondeo de la eval 15 (Priority: P2)

El cierre mide 6 respuestas de la eval 15 y no distingue un 20 % de un 10 %; el sondeo, de 40, sí, y dice además si la
skill reescrita conserva lo que ganó `brazo-W`. Antes de fusionar, una persona lo lanza con la skill de la propuesta y
decide con sus cifras.

**Why this priority**: es la medida con potencia del hito, pero la lanza una persona, fuera del run (ADR 0032).

**Independent Test**: el sondeo de la persona (SC-012).

**Acceptance Scenarios**:

1. **Dado** el sondeo de la eval 15 con la skill de la propuesta —el arnés de `evidencias/sondeos-20261025/` y
   `evidencias/sondeos-20261026/`, 20 repeticiones por modo, Claude Code 2.1.289—, **Cuando** da como mucho 2 de 40 en
   `afirma_lo_no_leido` y como mucho 4 de 40 en `cuenta_su_proceso`, **Entonces** la propuesta se puede fusionar.
2. **Dado** el mismo sondeo, **Cuando** no se cumple alguna de las dos cifras, **Entonces** la propuesta no se fusiona y se
   decide con sus cifras.

---

### Edge Cases

- **Lo remitido es de otra norma que no está en la tabla de normas**: se identifica con el paso 1; si no se encuentra o su
  lectura no termina con `0`, se nombra solo con las palabras del texto leído (paso 5 de la skill medida).
- **Una pregunta con fecha**: la lectura de lo remitido es la del paso 2 con esa fecha, como toda lectura de la skill;
  las evals 14, 15 y 17 no llevan fecha.
- **El sitio no da**: si el cuerpo pasa de 11 000 caracteres, `make skills-sync` y `make ci` dan el defecto y no se cierra
  (FR-005).
- **Un texto movido con una forma fija**: la forma sigue en el `SKILL.md` (FR-003), y su copia en la referencia, si la
  lleva, es una línea del `SKILL.md` (`lineas-de-aviso-de-las-referencias`).
- **Un texto movido con una orden concreta** que la caché de las sesiones no sirve sin red, o con una expresión de
  `evals/boe-legislacion/expresiones-prohibidas.yaml`: `ordenes-de-ejemplo` o `prosa-de-la-skill` salen en rojo (FR-008).
- **Ficheros del kit, de la caché o de las grabaciones modificados a mano** por algo que no es el binario ni el paso
  `grabar_datos`: regla genérica, defecto `inesperado`, código 1 (ADR 0023), sin casos.

## Requirements *(mandatory)*

### Functional Requirements

#### La skill de la entrega parte de la medida

- **FR-001**: `skills/boe-legislacion/` de v0.3.1 parte de la skill medida: su `SKILL.md` y sus referencias, con lo que la
  #200 cambió en `buscar_la_norma` y `redaccion_modificada`. Sobre ella, la entrega solo cambia lo que piden FR-005 a
  FR-007: el cuerpo de `SKILL.md`, y las referencias que reciben un texto que sale del cuerpo; una referencia nueva solo
  existe para recibir uno. Criterio: cada fichero de `references/` de la rama es igual al de la skill medida, salvo los
  que reciben un texto de la tabla del plan (FR-007), cuyo diff con la skill medida solo añade ese texto, la frase que
  lo presenta y, si la referencia remitía a ese texto en el cuerpo, el cambio de esa remisión. El texto que sale llega a la
  referencia palabra por palabra, tal como está en el `SKILL.md` de la skill medida: no se reescribe al moverlo.
- **FR-002**: Los pasajes medidos van en el cuerpo de v0.3.1 palabra por palabra, con su encabezado y su número de paso, y
  en ninguna referencia como única fuente: el paso 3, la primera viñeta del paso 4, el paso 5 y el paso 6. Criterio: cada
  uno es idéntico, línea a línea, a las líneas 48 a 59, 63 a 68, 78 a 86 y 88 a 100 del `SKILL.md` de la skill medida.
- **FR-003**: Cada forma fija de la skill medida —las tres de «Redacción en una fecha», la de «Redacción modificada», la
  de la regla 7 (`⚠ SIN CONSULTA AL BOE:`) y las etiquetas de aviso de la regla 3— sigue en el `SKILL.md` tal cual.
  Criterio: `linea-sin-consulta`, `formas-de-la-redaccion-fechada`, `avisos-de-la-skill` y `hallazgos-de-la-skill` salen
  en verde, y cada línea de un bloque `text` del `SKILL.md` de la skill medida es, sin los blancos de los extremos, una
  línea del `SKILL.md` de la rama.
- **FR-004**: La tabla de comandos no cambia: es la que genera `make skills-sync` entre sus marcas. Criterio: el texto
  entre sus marcas es igual al de la skill medida y `make skills-check` sale en verde.

#### El sitio en el cuerpo

- **FR-005**: El cuerpo de `SKILL.md` de v0.3.1 tiene como mucho 11 000 caracteres (ADR 0046). Criterio: `make
  skills-sync` y la comprobación de las skills de `make ci` salen en rojo con 11 001 o más (gate de H31, sin cambios).
- **FR-006**: Lo que toda respuesta necesita sigue en el cuerpo: el protocolo, con sus seis pasos en orden; «Cómo se
  cita»; y las reglas, entre ellas las invariantes de `CLAUDE.md` —no inventar contenido legal, cada afirmación con su
  cita, distinguir ley y reglamento, señalar la variación autonómica, no concluir que algo no existe, no actuar en nombre
  de nadie, el texto sale de `kitlegal boe` y no del grafo, y sin herramienta ni binario la respuesta lo dice—. Fuera de
  los pasajes medidos, las formas fijas y la tabla, un texto suyo se puede reescribir más corto, como piden las guías de
  autoría de agentskills.io y de Anthropic: quitar lo que el agente no haría mal sin ello y lo que el cuerpo ya dice en
  otro sitio, y decir el porqué en vez de enumerar prohibiciones. Reescribir no mueve lo que se remite: las referencias
  que FR-001 deja igual remiten al cuerpo por número y título («regla 1», «regla 2», «regla 3», «paso 2», «Cómo se
  cita», «Redacción en una fecha», «Redacción modificada», «Comandos»), y el cuerpo a su regla 7; por eso cada encabezado
  `##` y `###` del cuerpo conserva su texto y su orden, y cada una de las siete reglas su número y su encabezado en
  negrita. Criterio: el plan da, por cada texto reescrito, el de la skill medida y el de la entrega, y por qué lo
  quitado no hace falta; cada regla invariante de esta lista tiene su frase en el cuerpo de v0.3.1; la revisión final
  lo coteja; y la suite de aceptación sale en rojo si un encabezado o el número o el encabezado en negrita de una regla
  difieren de los de la skill medida.
- **FR-007**: Un texto de una clase de pregunta puede salir del cuerpo a la referencia a mano de esa clase, existente o
  nueva, y el paso del protocolo que lo necesita lleva, en la frase, el enlace a la referencia y la condición de leerla.
  Ningún texto que toda respuesta necesita sale del cuerpo. El enlace va solo en un paso numerado del protocolo que no sea
  un pasaje medido (el paso 1, el 2 o el paso 4 fuera de su primera viñeta), en la frase que dice cuándo se da la clase,
  con la condición de su clase y no con un «antes de responder» que valga para toda pregunta. Un texto cuya única
  necesidad está en un pasaje medido (FR-002) o fuera de los pasos numerados («Protocolo», «Cómo se cita», «Redacción en
  una fecha», «Redacción modificada», «Reglas») no sale: se queda en el cuerpo y, si no es un pasaje medido, una forma
  fija ni la tabla, se reescribe allí más corto (FR-006). Los enlaces que la skill medida ya tiene fuera de los pasos
  (`con_herramientas`, `como_se_cita`, `redaccion_modificada`) siguen como están, y un texto que sale a una de esas
  referencias lleva además su enlace y su condición en un paso numerado que no es un pasaje medido. Criterio: el plan da,
  por cada texto que sale, el de la skill medida (idéntico al que recibe la referencia), su clase, la referencia, el paso
  que la enlaza y la condición; `skills-check` sale en rojo con una referencia que el `SKILL.md` no enlaza (gate de H31,
  sin cambios); la suite de aceptación y la revisión final cotejan la referencia con el texto de la skill medida palabra
  por palabra y con el cuerpo.
- **FR-008**: Los controles de contenido de `docs/SKILLS.md` §1 siguen en verde sobre el `SKILL.md` y cada referencia a
  mano: `prosa-de-la-skill`, `ordenes-para-powershell`, `linea-sin-consulta`, `formas-de-la-redaccion-fechada`,
  `avisos-de-la-skill`, `hallazgos-de-la-skill`, `lineas-de-aviso-de-las-referencias`, `ordenes-de-ejemplo`,
  `normas-nombradas` y `sin-instrucciones-de-evals`. Criterio: `make ci`.
- **FR-009**: Lo que la rama añade al cuerpo y a las referencias es genérico (principio X): ninguna línea que añade
  respecto de `main` contiene `BOE-A-2004-4214`, `BOE-A-2006-20764`, `BOE-A-2013-12887`, «Ley 35/2006», «Real Decreto
  Legislativo 2/2004», «Ley 19/2013», «IRPF», «TRLRHL» ni «LTAIBG». Criterio: el diff de `skills/boe-legislacion/` de la
  rama frente a `main`, sin el frontmatter, que no cambia.
- **FR-010**: El frontmatter de `SKILL.md`, con la descripción, no cambia. Criterio: es igual al de `main`.

#### Las lecturas grabadas

- **FR-011**: Las `lecturas` de las evals llevan, además de las de hoy (`a15` en la 14 y `a18` en la 15), estos bloques,
  cada uno con su `motivo`, que dice que las respuestas de los sondeos y de los jobs lo nombran sin haberlo leído y que la
  regla de leer lo que se nombra lo pide: en `evals/boe-legislacion/15-irpf-rendimientos-por-materia.yaml`, `a7`, `a19`,
  `a20` y `dadecimoctava` de `BOE-A-2006-20764`; en `14-trlrhl-impuestos-por-materia.yaml`, `a60`, `a92`, `a100` y
  `a104` de `BOE-A-2004-4214`; y en `17-ltaibg-plazo-por-materia.yaml`, `a24` de `BOE-A-2013-12887`. Nada más de esas
  evals cambia ni de ninguna otra, y una lectura sigue sin contar en ningún umbral (H26.1). Criterio: la lectura del
  conjunto de evals las acepta, y el diff de `evals/` de la rama solo añade esas nueve entradas.
- **FR-012**: Una tarea `[datos]` añade al manifiesto `internal/evals/testdata/grabaciones.json`, en la entrada de su
  norma, los bloques de FR-011 que aún no tiene —todos menos `a24` de `BOE-A-2013-12887`, que ya está para
  `data/plazos.yaml`—, con su `para`, y el paso `grabar_datos` los graba desde `boe.legislacion-consolidada`, fuente con
  fila revisada en `docs/SOURCES.md`. Cada id es el del índice grabado de su norma (en esta sesión, los nueve están en
  `testdata/evals/boe.legislacion-consolidada/…_texto_indice.json`). Criterio: tras `grabar_datos`, cada bloque de
  FR-011 tiene su grabación, y la preparación de la caché no da ninguna falta.
- **FR-013**: La caché de las sesiones de `boe-legislacion` sirve cada bloque de FR-011 en las sesiones de su eval, sin
  fecha y con `--fecha 9999-12-31`, sin red. Criterio: los tests de `lecturas` de H26.1 —el de la caché de las sesiones y
  el de la lectura sin grabar— con los bloques nuevos, en `make ci` (SC-004).

#### La guía de autoría

- **FR-014**: `docs/SKILLS.md` §1 dice lo medido en `evidencias/sondeos-20261025/`: las sesiones abren las referencias de
  las que necesitan un dato para seguir y no las que solo traen guía, aunque el cuerpo mande leerlas —ninguna de las 40
  sesiones de `brazo-D` abrió la que el cuerpo mandaba leer antes de responder, y en los seis brazos se abrieron
  `normas`, de 23 a 40 sesiones por brazo, y, en los brazos con la skill de H31, `con_herramientas`, de 6 a 9—; sus
  consecuencias: lo que toda respuesta necesita va en el cuerpo, y una referencia nueva
  lleva el paso del protocolo que la pide; y las prácticas de las guías de autoría de agentskills.io y de Anthropic que lo
  respaldan: decir cuándo leer cada referencia, no solo que existe; explicar el porqué en vez de encadenar
  prohibiciones; y quitar del cuerpo lo que el agente no haría mal sin ello. El enlace de cada guía es el que registra
  el repositorio: `https://agentskills.io/` (`web/src/components/Pie.astro`) para la especificación; la URL de la guía
  de Anthropic no la registra ningún fichero, así que §1 no la escribe de memoria, enlaza a «Las dos fuentes»
  (`#las-dos-fuentes`) y el informe final la da como pendiente de la persona. Nada de §1 que no sea eso cambia.
  Criterio: el apartado lo dice con esas cifras, la ruta de la evidencia, `https://agentskills.io/` y el enlace a «Las
  dos fuentes»; la revisión final coteja las cifras con `LEEME.md` y que cada URL que la rama añade a §1 está en un
  fichero de `main`.

#### El cierre

- **FR-015**: Antes de cerrarse, el plan recalcula el peor caso del trabajo de evals de `boe-legislacion` con lo que
  añaden leer lo que se nombra y repasar en las evals por materia, y su tope (`timeout-minutes` de
  `.github/workflows/evals.yml`): si no lo cubre, el tope sube. Criterio: `TestDefinicionDelJob` sale en rojo en `make ci`
  si el tope del trabajo es menor que su peor caso (SC-005).
- **FR-016**: `CHANGELOG.md` lleva en *Unreleased* la entrada de `boe-legislacion` v0.3.1, que sustituye a v0.3: la
  respuesta empieza por lo que se pregunta, lo que nombra se lee antes, se repasa antes de enviar, y el cuerpo cabe en
  11 000 caracteres. Criterio: la entrada está en la sección *Unreleased*.

### Key Entities

- **Cuerpo de `SKILL.md`**: lo que el agente carga entero al activar la skill; protocolo, reglas, formas fijas y tabla de
  comandos generada; como mucho 11 000 caracteres.
- **Referencia a mano**: `skills/boe-legislacion/references/<nombre>.md` sin cabecera «generado desde»; el agente la lee
  si un paso del cuerpo la pide y la sesión la necesita.
- **Lectura de una eval** (`lecturas`, H26.1): `{norma, bloque, motivo}`; un bloque que la caché de las sesiones sirve
  porque la eval lo provoca; no es comando ni cita, ni cuenta en ningún umbral.
- **Entrada del manifiesto de grabaciones**: por norma, su búsqueda, sus `bloques` y su `para`; lo que graba
  `grabar_datos`.

## Uso, de fuera adentro

- **El cuerpo (FR-001 a FR-010)**: lo carga el agente una vez por pregunta que activa la skill, entero: como mucho 11 000
  caracteres: al menos 1 081 menos que la skill medida, y como mucho 8 más que el de `main`. No crece con el uso: lo escribe una
  persona y el gate lo acota. Lo gana cada respuesta: empieza por lo que se pregunta (US1) y no da materia de lo no leído
  (US2). Una referencia que recibe un texto se lee solo en la pregunta de su clase, cuando su paso lo pide (FR-007).
- **Las lecturas de la regla de leer lo que se nombra (FR-002)**: las pide el agente en el paso 3, y en el paso 6 la que
  faltó, una vez por precepto que la respuesta va a nombrar y no ha leído, con `articulo` o todas en una invocación de
  `articulos`, cada una con su comprobación detrás. Las acotan la respuesta y el salto: solo lo que la respuesta nombra,
  remitido por un bloque leído o salido de un índice, y nada de lo que remiten esas lecturas. Cada lectura devuelve un
  bloque, el que se pide por su id: con cientos de normas y miles de bloques consultados, el tamaño no cambia, porque no
  devuelve lo acumulado. En la eval 15, como mucho los cuatro bloques de FR-011 además de `a18`. Sus señales son las de
  toda lectura: los avisos de vigencia del bloque y, de la comprobación, `⚠ REDACCIÓN MODIFICADA:` solo si el bloque
  cambió desde una lectura anterior, que deja de darse al leerlo. Ninguna señal nueva. La mediana de las sesiones de la
  eval 15 es de 19 s con la skill de la #200 (`brazo-H`) y de 23 s con la medida (`evidencias/sondeos-20261026/LEEME.md`).
- **Las `lecturas` de las evals y sus grabaciones (FR-011 a FR-013)**: las pide la preparación de la caché de cada sesión
  de `boe-legislacion`, una vez por sesión: nueve bloques más, ocho grabaciones nuevas; las mismas, grabadas en la rama
  del primer run, van de 1 766 bytes (`a60`) a 402 526 bytes (`a7`), medidas en esta sesión con `git ls-tree -l`. No
  crecen con el uso de kitlegal: solo con lo que se escribe en `evals/`.
- **El apartado de `docs/SKILLS.md` §1 (FR-014)**: lo lee quien escribe o reparte una skill, una vez por skill; unos
  párrafos. Deja de valer cuando otra medida, en otra evidencia, diga otra cosa.
- **La entrada de `CHANGELOG.md` (FR-016)**: la lee quien actualiza kitlegal, una vez por versión.

## Success Criteria *(mandatory)*

### Measurable Outcomes

- **SC-001**: `make ci` en verde, con `make skills-check`, `ordenes-de-ejemplo` y los demás controles de contenido de
  `docs/SKILLS.md` §1 (FR-003, FR-004, FR-005, FR-008). Lo hace cumplir el propio `make ci`, que sale en rojo con
  cualquiera de ellos.
- **SC-002**: El cuerpo de `SKILL.md` tiene como mucho 11 000 caracteres; la comprobación de las skills de `make ci` y
  `make skills-sync` salen en rojo con 11 001 o más (FR-005).
- **SC-003**: Los pasajes medidos son idénticos a los de la skill medida (FR-002), las formas fijas siguen tal cual
  (FR-003), el frontmatter no cambia (FR-010) y los encabezados y las reglas conservan lo que fija FR-006; la suite de
  aceptación lo comprueba contra
  `evidencias/sondeos-20261026/brazo-W/skill/SKILL.md` y sale en rojo con una diferencia. La tabla de textos del plan
  (FR-006, FR-007) la coteja la revisión final.
- **SC-004**: Los tests de `lecturas` de H26.1 con los bloques nuevos: cada uno se sirve en la caché de su eval, sin
  fecha y con `9999-12-31`, y la caché de las sesiones lo da sin red; esos tests salen en rojo en `make ci` si la caché
  no sirve uno de ellos (FR-013).
- **SC-005**: El tope del trabajo de evals de `boe-legislacion`, recalculado antes de cerrar el plan, cubre su peor caso;
  `TestDefinicionDelJob` sale en rojo si no lo cubre (FR-015).
- **SC-006**: `CHANGELOG.md` (*Unreleased*) (FR-016); y el job de evals en la propuesta de cambio, con las tres skills,
  que el workflow mide en el cierre sobre la cabeza de la rama.
- **SC-007**: En el informe del job de cierre de `boe-legislacion`, `afirma_lo_no_leido:claude-sonnet-5-5` se cumple con
  el umbral de hoy, 0.028. El umbral decide: si la tasa pasa de 0.028, el veredicto del job es fallo y el job sale en
  rojo.
- **SC-008**: En el informe del job de cierre de `boe-legislacion`, ninguna sesión de `claude-sonnet-5-5` de las evals 14,
  15 y 17 tiene en `fuera_de_lo_grabado` una orden o llamada que sale con 4 o 5 porque la caché de la sesión no sirve un
  bloque de sus `lecturas` (umbral: 0 sesiones). No cuenta la orden o llamada `articulos` que sale con 5 porque pide, con
  bloques que la caché sirve, otro de la misma norma sin grabar: la salida se atribuye a ese bloque. Lo hacen cumplir los
  tests de SC-004, que salen en rojo en `make ci` si la caché de la skill no sirve un bloque de `lecturas` sin fecha y
  con `9999-12-31` —`TestCacheDeLasSesiones` y `TestLecturaDeLaEvalSinGrabar`, de `internal/evals/preparar_test.go`, con
  los nueve bloques de FR-011—, y la preparación de la sesión, que da la falta y no la abre si no está grabado (FR-012,
  FR-013). Una lectura con otra fecha no llega a `fuera_de_lo_grabado`: la lectura con `9999-12-31` deja guardada la
  entrada de las versiones del bloque, la caché preparada sirve su lectura en cualquier fecha (research D18 de H20) y sale
  con 0 o 3, no con 4 o 5. `fuera_de_lo_grabado` sigue publicándose con `decide: false`: el hito no toca el arnés para
  que decida.
- **SC-009**: Los demás umbrales que deciden del informe del job de cierre de `boe-legislacion` se cumplen; el que no se
  cumple pone su veredicto en fallo y el job en rojo.
- **SC-010**: `cuenta_su_proceso` se publica por modo, como hoy, sin cifra de aceptación en el cierre: no decide, y
  hacerla decidir pide medir antes su juez (ADR 0037).
- **SC-011**: `jurisprudencia` y `legal-core` siguen aprobadas: el veredicto de su job de cierre es aprobado, y si no lo
  es el job sale en rojo.
- **SC-012**: Lo que la entrada deja a una persona antes de fusionar: un sondeo de la eval 15 con la skill de la
  propuesta, como los de `evidencias/sondeos-20261025/` y `evidencias/sondeos-20261026/` —el mismo arnés, 20
  repeticiones por modo, Claude Code 2.1.289—, de 40 respuestas, en el que `afirma_lo_no_leido` dé como mucho 2 de 40,
  por debajo del control de `evidencias/sondeos-20261025/`, 16 de 80, con Fisher unilateral p < 0,05 (2 de 40 frente a
  16 de 80 da p ≈ 0,023; 3 de 40, p ≈ 0,062; calculado en esta sesión), y `cuenta_su_proceso` dé como mucho 4 de 40, por
  debajo de los 17 de 80 del control (el hito pide la cifra, no la significación: 4 de 40 frente a 17 de 80 da p ≈
  0,099). Lo mide una persona fuera del run (ADR 0032): el informe final lo da como pendiente. Si no se cumple, la
  propuesta no se fusiona y se decide con sus cifras.
- **SC-013**: `docs/SKILLS.md` §1 dice lo de FR-014 con sus cifras, su evidencia, `https://agentskills.io/` y el enlace
  a «Las dos fuentes», sin ninguna URL que no esté en un fichero de `main`; lo coteja la revisión final.

### Criterios del hito, tal como los escribe

**Controles** (SC-001, SC-004, SC-005, SC-006): «`make ci` en verde, con `make skills-check`, `ordenes-de-ejemplo` y los
demás controles de contenido de `docs/SKILLS.md` §1; los tests de `lecturas` de H26.1 con los bloques nuevos: cada uno se
sirve en la caché de su eval, sin fecha y con `9999-12-31`, y la caché de las sesiones lo da sin red; el tope del trabajo
de evals de `boe-legislacion`, recalculado antes de cerrar el plan; `CHANGELOG.md` (*Unreleased*); y el job de evals en
la propuesta de cambio, con las tres skills.»

**Aceptación** (SC-007 a SC-012): «en el informe del job de cierre de `boe-legislacion`,
`afirma_lo_no_leido:claude-sonnet-5-5` se cumple con el umbral de hoy, 0.028; ninguna sesión de `claude-sonnet-5-5` de
las evals 14, 15 y 17 tiene en `fuera_de_lo_grabado` un bloque de sus `lecturas`; y los demás umbrales que deciden se
cumplen. `cuenta_su_proceso` se publica por modo, como hoy, sin cifra de aceptación en el cierre: no decide, y hacerla
decidir pide medir antes su juez (ADR 0037); `jurisprudencia` y `legal-core` siguen aprobadas; lo que la entrada deja a
una persona antes de fusionar: un sondeo de la eval 15 con la skill de la propuesta, como los de
`evidencias/sondeos-20261025/` y `evidencias/sondeos-20261026/` —el mismo arnés, 20 repeticiones por modo, Claude Code
2.1.289—, de 40 respuestas, en el que `afirma_lo_no_leido` dé como mucho 2 de 40, por debajo del control de
`evidencias/sondeos-20261025/`, 16 de 80, con Fisher unilateral p < 0,05, y `cuenta_su_proceso` dé como mucho 4 de 40,
por debajo de los 17 de 80 del control. El cierre mide 6 respuestas de la eval 15 y no distingue un 20 % de un 10 %; el
sondeo, sí, y es además el que dice si la skill reescrita conserva lo que ganó `brazo-W`. Si no se cumple, la propuesta no
se fusiona y se decide con sus cifras.»

## Assumptions

- Lo reescrito no se midió: la skill de la entrega ya no es la réplica de `brazo-W`, y por eso decide el sondeo de la
  persona (SC-012), no el cierre.
- Las sesiones del run no tienen red fuera de las herramientas de Go: lo que `docs/SKILLS.md` §1 dice de las guías de
  autoría es lo que el hito les atribuye y lo que ya registra «Las dos fuentes», que no da ninguna URL; no lleva una
  cita literal ni una URL que no se hayan consultado. La única URL de una guía que registra el repositorio es
  `https://agentskills.io/`, en `web/src/components/Pie.astro` (FR-014).
- `a24` de `BOE-A-2013-12887` ya está en el manifiesto y grabado (lo pide `data/plazos.yaml`, H9): la eval 17 lo lleva en
  `lecturas` igualmente, con su motivo, como pide el hito. Si `grabar_datos` no puede grabar un bloque porque la fuente
  no responde, es una causa mayor del run (dependencia externa inaccesible), no un caso del producto.
- La serie de cierres de `boe-legislacion` empieza de cero, porque su clave es el árbol de `skills/boe-legislacion` (ADR
  0038): es la consecuencia de cambiar el árbol, no algo que haga el hito.
- La versión de la skill solo se registra en `CHANGELOG.md`.

## Fuera de alcance

- Cualquier umbral, la rúbrica y los casos del juez, y que `cuenta_su_proceso` decida: subir un umbral pide un ADR, y
  hacer decidir una clase que solo se publica, otro.
- Que `kitlegal boe` devuelva la rúbrica de los preceptos a los que remite un bloque (backlog; `evidencias/adr-0038/`).
- Otra redacción de los pasos 3, 5 y 6 y de la primera viñeta del paso 4 que la medida.
- `jurisprudencia`, `legal-core` y la tabla de comandos del cuerpo.
- La reorganización de las skills por capacidad (H32).
- La descripción y el resto del frontmatter de `boe-legislacion` (FR-010).
- Reescribir las referencias más allá de recibir los textos que salen del cuerpo (FR-001), y reescribir al moverlo un
  texto que sale: llega palabra por palabra.
- Renumerar o juntar las reglas del cuerpo y cambiar un encabezado suyo: las referencias remiten a ellos (FR-006).
- Escribir en `docs/SKILLS.md` §1 una URL que no registra ningún fichero del repositorio (FR-014).
- Sacar del cuerpo un texto cuya única necesidad está en un pasaje medido o fuera de los pasos numerados («Protocolo»,
  «Cómo se cita», «Redacción en una fecha», «Redacción modificada», «Reglas»): se queda en el cuerpo (FR-007).
- Que `fuera_de_lo_grabado` decida en el cierre de `boe-legislacion`, y cambiar el arnés de evals para ello (SC-008).
- Evals nuevas, y cambios en las evals de `boe-legislacion` que no sean las `lecturas` de FR-011.
- Cambios en el binario `kitlegal` (applets, `internal/source`, `internal/core`) y en el arnés de evals, salvo el tope del
  trabajo de FR-015 y lo que sus tests esperan.
- Cambiar el tope de una sesión de las evals (`topeDeUnaSesion`, 240 s en `internal/evals/definicion.go`), común a las
  tres skills: el hito pide recalcular el tope del trabajo, no ese.
- Lanzar el sondeo de SC-012 y versionar su evidencia: los hace una persona fuera del run.
- `docs/SKILLS.md` fuera de §1; el ADR 0046, la constitución, `CLAUDE.md` y `docs/ROADMAP.md`.
- Cambios en `scripts/workflow/`, `.specify/` e `internal/workflow/`, y en `web/`.
- La release y su etiqueta, y la fusión: las decide una persona.
