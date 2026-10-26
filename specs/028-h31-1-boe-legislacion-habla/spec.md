# Feature Specification: H31.1 · `boe-legislacion` habla de la norma, no de la consulta, y lee lo que nombra

**Feature Branch**: `028-h31-1-boe-legislacion-habla`

**Created**: 2026-10-25

**Status**: Draft

**Input**: el fichero del hito, `docs/roadmap/in-progress/H31.1-boe-legislacion-habla-de-la-norma-y-lee-lo-que-nombra.md`
(la rama lo pasa a `docs/roadmap/done/` en `3dec6e3`), la única entrada humana del run (ADR 0018), con su evidencia en
`evidencias/sondeos-20261025/`, que está en `main` desde la propuesta de la entrada (#199): el run la usa, no la
escribe.

## Contexto

El hito no entrega skill nueva: protege `boe-legislacion` (principio VIII) de dos defectos medidos antes de que H32
reorganice las skills sobre ella. Lo medido está en `evidencias/sondeos-20261025/LEEME.md`:

- **`cuenta_su_proceso`**: las respuestas de `claude-sonnet-5-5` que cuentan lo que hizo el agente pasaron, en el job, del
  1,8 % (modo orden) y el 2,8 % (modo herramienta) antes de H31 al 26,5 % y el 34,0 % desde H31; en el sondeo de la eval
  15, la skill de antes de H31 (`brazo-E`, el árbol `27aeccc` de `d14bb86^`) da 0 de 40 y la de `main` (`brazo-A` y
  `brazo-A2`), 17 de 80.
- **`afirma_lo_no_leido`** en las evals por materia: casi todas las respuestas de la eval 15 nombran el art. 7 de la Ley
  35/2006, al que remite el art. 17 y que ninguna sesión lee, y la marca llega cuando le pegan su materia. La skill de
  antes de H31 marca lo mismo que la de hoy (8 de 40 frente a 16 de 80), y ninguna de tres redacciones de la regla lo
  baja (`brazo-B` a `brazo-D`).
- **Lo que va a una referencia como guía no se lee**: ninguna de las 40 sesiones de `brazo-D` abrió la referencia que el
  cuerpo mandaba leer antes de responder; en los seis brazos, las sesiones abren `normas` (de 23 a 40 por brazo) y, con
  la skill de H31, `con_herramientas` (de 6 a 9), las referencias de las que necesitan un dato para seguir.

Dos nombres del hito, fijados aquí para que cada requisito tenga una sola lectura:

- **«La regla 5»** del hito es el paso 5 del protocolo del cuerpo de hoy, «De un precepto no leído, nada», el que dice
  que lo remitido se lee y se cita o se nombra con las palabras leídas; no la regla 5 del apartado «Reglas» de hoy
  («Ningún caso especial para un territorio»), que no cambia. **«El paso 3»** es el de hoy, «Completar».
- **«El paso 5, «Responder citando», y la última viñeta del paso 4»** son los del cuerpo de antes de H31, el
  `skills/boe-legislacion/SKILL.md` del árbol `27aeccc` (`d14bb86^`): su paso 5 entero y la viñeta del paso 4 «Lo que no
  puedas leer, no lo suplas (paso 5); lo que decidas aquí no va en la respuesta: quien pregunta no lo ve». En la skill
  de hoy su sitio es el paso 4, «Responder», y el paso 5.

Medido en esta sesión con `wc -m` sobre el cuerpo (desde la línea que sigue al `---` que cierra el frontmatter): el de
hoy tiene 10 992 caracteres, 8 por debajo del máximo de 11 000 (ADR 0046).

## Clarifications

### Session 2026-10-25

- Q: Si el plan concluye que las sesiones de las evals 14, 15 y 17, con las lecturas que añade la regla de leer lo que se nombra, no caben en el tope de una sesión (topeDeUnaSesion = 240 s en internal/evals/definicion.go, común a las tres skills), ¿qué hace este hito con ese tope? (FR-013) → A: Opción C. El hito no cambia el tope de una sesión: queda fuera de alcance y se cierra con 240 s; FR-013 recalcula el peor caso del trabajo de boe-legislacion y su `timeout-minutes` con ese tope (tandas de 272 s), y si una sesión no cabe, el informe del cierre lo muestra con el motivo «tope de 240 s agotado (código 124)»; la persona decide con esas cifras. No hay medida de que las sesiones no quepan (la evidencia de los sondeos no guarda la duración de cada sesión). (auto: conservadora; criterio d; fuente: hito H31.1 «Alcance», «La duración», «Controles» y «Fuera de alcance»; ADR 0016, línea 96; internal/evals/definicion.go:72-89; .github/workflows/evals.yml:250-276; constitución, «Criterio de decisión autónoma», puntos 2 y 4)
- Q: ¿Qué lecturas del paso 3 son el único salto cuyas remisiones ya no se leen y se nombran solo con las palabras del texto leído: solo las de un precepto remitido por un bloque del paso 2, también las de un id que solo salió en un índice, y también lo remitido que el paso 3 lee hoy porque cambia la respuesta? (FR-002, FR-003) → A: Opción B. El salto son las lecturas por remisión de un bloque del paso 2 y las de un id salido de un índice; lo que remiten no se lee y se nombra solo con las palabras del texto leído, también si la lectura falla. La regla del paso 3 de hoy sobre lo remitido que cambia la respuesta no cambia y se sigue leyendo sin límite de saltos; el plan cuenta para FR-013 las lecturas de la regla nueva, con un solo salto, además de las de hoy. (auto: criterio c; fuente: hito H31.1 «Alcance», «Lo que la respuesta nombra, se lee antes» y «Objetivo»; skills/boe-legislacion/SKILL.md:51; constitución, «Criterio de decisión autónoma», punto 2; spec, «Edge Cases» y «Uso, de fuera adentro»)

## User Scenarios & Testing *(mandatory)*

### User Story 1 - La respuesta empieza por lo que se pregunta y habla de la norma (Priority: P1)

Quien pregunta por una norma lee solo la respuesta: no ve las órdenes ni las llamadas, ni lo que devuelven, ni lo que el
agente decide. Una respuesta que cuenta la consulta («la comprobación no detectó cambios…») le habla de algo que no ve.
La skill vuelve a llevar en el cuerpo lo que el de antes de H31 decía de cómo empieza la respuesta y de qué habla, con
sus ejemplos.

**Why this priority**: es el defecto que H31 introdujo y el que el sondeo atribuye al texto de la skill (0 de 40 con la
de antes de H31 frente a 17 de 80 con la de hoy).

**Independent Test**: se compara el cuerpo de v0.3.1 con los dos pasajes de `27aeccc` —su paso 5 entero y la última viñeta
de su paso 4— (FR-001) y se buscan
en él sus seis ejemplos; y lo mide una persona con el sondeo de la eval 15 (SC-012).

**Acceptance Scenarios**:

1. **Dado** el cuerpo de `skills/boe-legislacion/SKILL.md` de v0.3.1, **Cuando** se buscan en él los ejemplos
   «el artículo no lleva avisos de vigencia», «la lectura no trajo avisos», «el art. N, que es…», «el art. N sobre
   <materia>», «los artículos que regulan <materia> (N y M)» y «no puedo decir qué <regla> fija», **Entonces** están los
   seis, fuera de cualquier referencia.
2. **Dado** el cuerpo de v0.3.1, **Cuando** se lee su paso «Responder», **Entonces** dice que la respuesta es todo lo que
   se escribe tras la última orden o llamada, desde su primera palabra; que quien pregunta la lee entera y no ve las
   órdenes, las llamadas ni lo que devuelven; que lo que se decide al completar la lectura no va en la respuesta; y que
   lo que el agente ha hecho, va a hacer o ha devuelto una orden no le sirve.
3. **Dado** el sondeo de la eval 15 que lanza una persona con la skill de la propuesta (40 respuestas), **Cuando** se
   cuentan las marcas de `cuenta_su_proceso`, **Entonces** son como mucho 2.

---

### User Story 2 - Lo que la respuesta nombra, se ha leído antes (Priority: P1)

La respuesta solo nombra un artículo, un apartado o una disposición cuyo texto ha devuelto `kitlegal boe articulo` o
`articulos` en la conversación. Si el bloque leído remite a otro precepto y la respuesta lo va a nombrar, el agente lo
lee antes de escribirla, y entonces puede decir de qué trata, con su cita. No sigue la cadena: lo que remite esa segunda
lectura se nombra solo con las palabras del texto leído.

**Why this priority**: es el único candidato sin probar contra `afirma_lo_no_leido`, que con la tasa de hoy pone en rojo
el umbral de la clase (0.028) más de 1 de cada 50 ejecuciones completas.

**Independent Test**: se lee el paso 3 y el paso 5 del cuerpo de v0.3.1 (FR-002 a FR-004); y lo mide el cierre
(SC-007) y, con más respuestas, el sondeo de una persona (SC-012).

**Acceptance Scenarios**:

1. **Dado** un bloque leído en el paso 2 que remite a otro precepto, **Cuando** la respuesta lo va a nombrar, **Entonces**
   el paso 3 del cuerpo manda leerlo con los pasos 1 y 2 —su norma identificada, su lectura y su comprobación— antes de
   escribir la respuesta, y la respuesta puede decir de qué trata con su cita.
2. **Dado** un precepto leído por esa remisión que remite a su vez a un tercero, **Cuando** la respuesta nombra el
   tercero, **Entonces** el cuerpo dice que el tercero no se lee y se nombra solo con las palabras del texto leído, sin
   su materia.
3. **Dado** un precepto remitido cuya lectura termina con un código distinto de `0`, **Cuando** la respuesta lo nombra,
   **Entonces** el cuerpo dice que se nombra solo con las palabras del texto leído, como dice hoy el paso 5, y la regla 2
   sigue diciendo qué bloques no se pudieron consultar y por qué.
4. **Dado** un precepto cuyo id solo ha salido en un índice, **Cuando** la respuesta lo va a nombrar, **Entonces** el
   cuerpo dice que el índice no basta: se lee antes con los pasos 1 y 2 o no se nombra.
5. **Dado** el job de cierre de `boe-legislacion`, **Cuando** el juez mide las respuestas de `claude-sonnet-5-5`,
   **Entonces** `afirma_lo_no_leido:claude-sonnet-5-5` se cumple con 0.028.
6. **Dado** el sondeo de la eval 15 de una persona con la skill de la propuesta, **Cuando** se cuentan las marcas de
   `afirma_lo_no_leido`, **Entonces** son como mucho 2 de 40.

---

### User Story 3 - Las evals por materia sirven lo que sus respuestas nombran (Priority: P1)

Las evals 14, 15 y 17 llevan en `lecturas` los preceptos que sus respuestas nombran sin haberlos leído, para que la
regla de leer lo que se nombra se pueda cumplir en las sesiones sin red; sin ellos, la lectura sale con 5.

**Why this priority**: sin las grabaciones, la regla de la US2 pone las lecturas de las evals por materia fuera de lo
grabado y el cierre no mide la skill, sino la caché.

**Independent Test**: los tests de `lecturas` de H26.1 con los bloques nuevos en `make ci` (SC-004).

**Acceptance Scenarios**:

1. **Dado** el manifiesto `internal/evals/testdata/grabaciones.json` y sus grabaciones tras la tarea `[datos]`, **Cuando**
   se prepara la caché de una sesión de la eval 15, **Entonces** `kitlegal boe articulo BOE-A-2006-20764 a7 --offline`
   sale con 0 sin fecha y con `--fecha 9999-12-31`, y lo mismo `a19`, `a20` y `dadecimoctava`.
2. **Dado** lo mismo, **Cuando** se prepara la caché de una sesión de la eval 14, **Entonces** salen con 0, sin fecha y con
   `--fecha 9999-12-31`, `a60`, `a92`, `a100` y `a104` de `BOE-A-2004-4214`; y de la eval 17, `a24` de
   `BOE-A-2013-12887`.
3. **Dado** un bloque de `lecturas` sin grabar, **Cuando** se prepara la sesión, **Entonces** la preparación da la falta y
   no la abre, como hoy (H26.1).

---

### User Story 4 - El cierre mide la skill con los umbrales de hoy (Priority: P2)

El job de evals de la propuesta de cambio mide las tres skills; los umbrales que deciden, con sus valores de hoy, dan el
veredicto.

**Why this priority**: es la aceptación sin persona del hito; las cifras finas las da el sondeo (US6).

**Independent Test**: el informe del job de cierre sobre la cabeza de la rama (SC-007 a SC-011).

**Acceptance Scenarios**:

1. **Dado** el informe del job de cierre de `boe-legislacion`, **Cuando** se leen sus umbrales, **Entonces**
   `afirma_lo_no_leido:claude-sonnet-5-5` se cumple con 0.028 y los demás umbrales que deciden se cumplen; y
   `cuenta_su_proceso` se publica por modo, sin decidir.
2. **Dado** ese informe, **Cuando** se leen las sesiones de `claude-sonnet-5-5` de las evals 14, 15 y 17, **Entonces**
   ninguna tiene en `fuera_de_lo_grabado` una orden o llamada que sale con 4 o 5 porque la caché de la sesión no sirve
   un bloque de las `lecturas` de su eval, contada como dice SC-008.
3. **Dado** los informes del job de cierre de `jurisprudencia` y `legal-core`, **Cuando** se lee su veredicto, **Entonces**
   es aprobado.
4. **Dado** la definición del job, **Cuando** corre `TestDefinicionDelJob` en `make ci`, **Entonces** el tope del trabajo de
   `boe-legislacion` cubre su peor caso, recalculado por el plan.

---

### User Story 5 - La guía dice qué referencias se leen (Priority: P2)

Quien escribe o reparte una skill —H32 la primera— sabe por `docs/SKILLS.md` §1 que las sesiones abren las referencias
de las que necesitan un dato para seguir y no las que solo traen guía, aunque el cuerpo mande leerlas.

**Why this priority**: H32 lo necesita escrito antes de repartir skills y referencias.

**Independent Test**: se lee `docs/SKILLS.md` §1 (FR-012).

**Acceptance Scenarios**:

1. **Dado** `docs/SKILLS.md` §1 de la rama, **Cuando** se lee, **Entonces** dice lo medido, con sus cifras y la ruta
   `evidencias/sondeos-20261025/`, y sus dos consecuencias: lo que toda respuesta necesita va en el cuerpo, y una
   referencia nueva lleva el paso del protocolo que la pide.

---

### User Story 6 - Una persona decide la fusión con el sondeo de la eval 15 (Priority: P2)

El cierre mide 6 respuestas de la eval 15 y no distingue un 20 % de un 10 %; el sondeo, de 40, sí. Antes de fusionar,
una persona lo lanza con la skill de la propuesta, como los de `evidencias/sondeos-20261025/`, y decide con sus cifras.

**Why this priority**: es la medida con potencia del hito, pero la lanza una persona, fuera del run (ADR 0032).

**Independent Test**: el sondeo de la persona (SC-012).

**Acceptance Scenarios**:

1. **Dado** el sondeo de la eval 15 con la skill de la propuesta —el arnés de `evidencias/sondeos-20261025/`, 20
   repeticiones por modo, Claude Code 2.1.289—, **Cuando** da como mucho 2 de 40 en `afirma_lo_no_leido` y como mucho 2
   de 40 en `cuenta_su_proceso`, **Entonces** la propuesta se puede fusionar.
2. **Dado** el mismo sondeo, **Cuando** solo `cuenta_su_proceso` da como mucho 2 de 40, **Entonces** lo del cuerpo de la
   respuesta (US1) puede ir solo en su propia propuesta, y leer lo que se nombra (US2) vuelve al análisis.
3. **Dado** el mismo sondeo, **Cuando** ninguna de las dos cifras se cumple, **Entonces** la propuesta no se fusiona y se
   decide con sus cifras.

---

### Edge Cases

- **Lo remitido es de otra norma que no está en la tabla de normas**: se identifica con el paso 1 (búsqueda por su
  título); si no se encuentra o su lectura falla, se nombra solo con las palabras del texto leído (US2, escenario 3).
- **Lo remitido no se nombra en la respuesta**: no hay lectura que añadir; sigue valiendo lo de hoy, que se lee lo
  remitido que cambia la respuesta.
- **Una pregunta con fecha**: la lectura del precepto remitido es la del paso 2 con esa fecha («Redacción en una fecha»),
  como toda lectura de la skill; las evals 14, 15 y 17 no llevan fecha.
- **El sitio no da**: si con lo que vuelve y la regla nueva el cuerpo pasa de 11 000 caracteres, `make skills-sync` y
  `make ci` dan el defecto y no se cierra hasta sacar del cuerpo lo que solo pide una clase de pregunta (FR-005, FR-006).
- **Un ejemplo con una expresión de `evals/boe-legislacion/expresiones-prohibidas.yaml`**, o una orden concreta que la
  caché de las sesiones no sirve sin red, en el cuerpo o en una referencia a mano: `prosa-de-la-skill` u
  `ordenes-de-ejemplo` salen en rojo (FR-007).
- **Ficheros del kit, de la caché o de las grabaciones modificados a mano** por algo que no es el binario ni el paso
  `grabar_datos`: regla genérica, defecto `inesperado`, código 1 (ADR 0023), sin casos.

## Requirements *(mandatory)*

### Functional Requirements

#### El cuerpo: cómo empieza la respuesta y de qué habla

- **FR-001**: El cuerpo de `skills/boe-legislacion/SKILL.md` de v0.3.1 lleva cada regla del paso 5, «Responder citando»,
  y de la última viñeta del paso 4 del cuerpo de `27aeccc`, con sus ejemplos: lo que H31 comprimió o llevó a una
  referencia vuelve al cuerpo, en el paso «Responder» o en el paso 5. Todas son de toda respuesta. Una regla que el
  cuerpo de hoy ya dice con el mismo contenido se queda como está, salvo las de lo remitido y del precepto no leído: esas
  vuelven como las dejan FR-002 y FR-003, que prevalecen. El plan trae cada regla desde la tabla de reglas de
  `specs/027-h31-las-skills-siguen/research.md` y dice, por cada una, la frase del cuerpo de v0.3.1 que la lleva; la de
  esas dos, con el cambio de FR-002 y FR-003.
  Criterio: el cuerpo de v0.3.1 contiene los seis ejemplos de esos pasajes —«el artículo no lleva avisos de vigencia»,
  «la lectura no trajo avisos», «el art. N, que es…», «el art. N sobre <materia>», «los artículos que regulan <materia>
  (N y M)» y «no puedo decir qué <regla> fija»—, y cada regla de los dos pasajes de `27aeccc` —su paso 5 entero y la
  última viñeta de su paso 4— tiene su fila en la tabla del plan con su frase en el cuerpo, que la revisión final coteja.

#### El cuerpo: lo que la respuesta nombra, se lee antes

- **FR-002**: El paso 3, «Completar», dice que la respuesta nombra un artículo, un apartado o una disposición solo si
  `kitlegal boe articulo` o `articulos` ha devuelto su texto en la conversación —su id en un índice no basta—, salvo en
  los dos casos de FR-003; que si un bloque leído en el paso 2 remite a él, o su id solo salió en un índice, y la
  respuesta lo va a nombrar, se lee con los pasos 1 y 2 antes de escribirla, o no se nombra; y que, leído, la respuesta
  puede decir de qué trata, con su cita. Criterio: el paso 3 del cuerpo lo dice; la revisión final lo coteja.
- **FR-003**: El cuerpo dice que la lectura de FR-002 es un salto: aquello a lo que remite el precepto leído por esa
  remisión no se lee, y se nombra solo con las palabras del texto leído; y que si una lectura de FR-002 falla, el
  precepto también se nombra solo con las palabras del texto leído, como dice hoy el paso 5. El salto son solo las dos
  lecturas de FR-002 —la de un precepto al que remite un bloque leído en el paso 2 y la de un precepto cuyo id solo salió
  en un índice—; la regla del paso 3 de hoy, «Lee lo remitido que cambia la respuesta, de esta norma o de otra, resuelta
  antes con los pasos 1 y 2», no cambia: lo remitido que cambia la respuesta se sigue leyendo y citando como hoy, sin
  límite de saltos. El plan cuenta para FR-013 las lecturas de FR-002, con un solo salto, además de las de hoy.
  Criterio: el paso 5 del cuerpo lo dice, junto a la prohibición de juntar número y materia de lo no leído, y el paso 3
  conserva su regla de hoy; la revisión final lo coteja.
- **FR-004**: Lo de FR-002 y FR-003 va en el cuerpo, en el paso 3 y en el paso 5, y en ninguna referencia como única
  fuente. Es genérico (principio X): no nombra una materia, y sus ejemplos, si los lleva, no usan ninguna norma de las
  evals 14, 15 y 17 (`BOE-A-2004-4214`, `BOE-A-2006-20764`, `BOE-A-2013-12887`) ni sus abreviaturas o números. Criterio:
  `sin-instrucciones-de-evals` sale en verde, y las líneas que la rama añade al cuerpo no contienen ninguno de esos tres
  identificadores, ni «Ley 35/2006», ni «Real Decreto Legislativo 2/2004», ni «Ley 19/2013», ni «IRPF», ni «TRLRHL».

#### El sitio en el cuerpo

- **FR-005**: El cuerpo de v0.3.1 tiene como mucho 11 000 caracteres (ADR 0046). Criterio: `make skills-sync` y la
  comprobación de las skills de `make ci` salen en rojo con 11 001 o más (gate de H31, sin cambios).
- **FR-006**: El sitio sale de textos que solo pide una clase de pregunta. Cada texto que sale del cuerpo va a una
  referencia a mano, existente o nueva, y el cuerpo lleva, en el paso del protocolo que la necesita, la frase con su
  enlace que dice cuándo leerla; el plan lo dice por cada texto. No sale del cuerpo nada que toda respuesta necesita —el
  protocolo, las reglas invariantes, lo de FR-001 a FR-004— ni ninguna forma fija. La tabla de comandos no cambia: es la
  que genera `make skills-sync`. Criterio: los controles `linea-sin-consulta`, `formas-de-la-redaccion-fechada`,
  `avisos-de-la-skill`, `hallazgos-de-la-skill` y `lineas-de-aviso-de-las-referencias` salen en verde; `skills-check` da
  en rojo la referencia sin enlace desde `SKILL.md` (gate de H31, sin cambios); y el plan lista cada texto que sale con
  el paso que enlaza su referencia y la condición de leerla, que la revisión final coteja con el cuerpo.
- **FR-007**: Los controles de contenido de `docs/SKILLS.md` §1 siguen en verde sobre el `SKILL.md` y cada referencia a
  mano: `prosa-de-la-skill`, `ordenes-para-powershell`, `ordenes-de-ejemplo`, `normas-nombradas` y
  `sin-instrucciones-de-evals`, además de los de FR-006. Una orden concreta nueva de `kitlegal boe` en el cuerpo o en una
  referencia sale con 0 con `--offline` sobre la caché de una sesión de la skill. Criterio: `make ci`.
- **FR-008**: La descripción de la skill no cambia. Criterio: el `description` del frontmatter de la rama es igual al de
  `main`.

#### Las lecturas grabadas

- **FR-009**: Las `lecturas` de las evals llevan, además de las de hoy (`a15` en la 14 y `a18` en la 15), estos bloques,
  cada uno con su `motivo`, que dice que las respuestas de los sondeos y de los jobs lo nombran sin haberlo leído y que la
  regla de leer lo que se nombra lo pide: en `evals/boe-legislacion/15-irpf-rendimientos-por-materia.yaml`, `a7`, `a19`,
  `a20` y `dadecimoctava` de `BOE-A-2006-20764`; en `14-trlrhl-impuestos-por-materia.yaml`, `a60`, `a92`, `a100` y
  `a104` de `BOE-A-2004-4214`; y en `17-ltaibg-plazo-por-materia.yaml`, `a24` de `BOE-A-2013-12887`. Nada más de esas
  evals cambia —pregunta, `comandos`, `citas`, `informativa`, `por_materia`— ni ninguna otra eval; y una lectura sigue
  sin contar en ningún umbral (H26.1). Criterio: la lectura del conjunto de evals las acepta, y el diff de `evals/` de la
  rama solo añade esas nueve entradas.
- **FR-010**: Una tarea `[datos]` añade al manifiesto `internal/evals/testdata/grabaciones.json`, en la entrada de su
  norma, los bloques de FR-009 que aún no tiene —todos menos `a24` de `BOE-A-2013-12887`, que ya está grabado para
  `data/plazos.yaml`—, con su `para`, y el paso `grabar_datos` del workflow los graba en
  `internal/evals/testdata/boe.legislacion-consolidada/` desde la fuente `boe.legislacion-consolidada`, con fila revisada
  en `docs/SOURCES.md`. Cada id es el que da el índice grabado de su norma, nunca compuesto ni de memoria. Criterio: tras
  `grabar_datos`, cada bloque de FR-009 tiene su grabación, y la preparación de la caché no da ninguna falta.
- **FR-011**: La caché de las sesiones de `boe-legislacion` sirve cada bloque de FR-009 en las sesiones de su eval, sin
  fecha y con `--fecha 9999-12-31`, sin red. Criterio: los tests de `lecturas` de H26.1 —el de la caché de las sesiones y
  el de la lectura sin grabar— con los bloques nuevos en `make ci` (SC-004).

#### La guía de autoría

- **FR-012**: `docs/SKILLS.md` §1 dice lo medido en `evidencias/sondeos-20261025/`: las sesiones abren las referencias de
  las que necesitan un dato para seguir y no las que solo traen guía, aunque el cuerpo mande leerlas —ninguna de las 40
  sesiones de `brazo-D` abrió la que el cuerpo mandaba leer antes de responder, y en los seis brazos se abrieron
  `normas` (de 23 a 40 sesiones por brazo) y, en los brazos con la skill de H31 (todos menos `brazo-E`, que no tiene esa
  referencia), `con_herramientas` (de 6 a 9)—; y sus consecuencias: lo que toda respuesta
  necesita va en el cuerpo, y una referencia nueva lleva el paso del protocolo que la pide. Criterio: el apartado lo
  dice con esas cifras y la ruta de la evidencia; la revisión final lo coteja con `LEEME.md`.

#### El cierre

- **FR-013**: Antes de cerrarse, el plan recalcula el peor caso del trabajo de evals de `boe-legislacion` con lo que añade
  leer lo que se nombra en las evals por materia, y su tope (`timeout-minutes` de `.github/workflows/evals.yml`): si no
  lo cubre, el tope sube. El tope de una sesión (`topeDeUnaSesion`, 240 s, común a las tres skills) no cambia: el hito se
  cierra con 240 s. El peor caso cuenta cada sesión en su tope (tandas de 272 s: los 240 s, los 10 s de margen hasta KILL
  y los 22 s de preparación), así que las lecturas nuevas no lo suben mientras no cambien ni ese tope ni el número de
  sesiones, y el plan lo deja escrito al recalcularlo. Si una sesión de las evals 14, 15 o 17 no cabe en 240 s, termina
  con el motivo «tope de 240 s agotado (código 124)» y el informe del cierre lo muestra, junto con su efecto en los
  umbrales que deciden; la persona decide con esas cifras al leer el informe final. Criterio: `TestDefinicionDelJob`
  sale en rojo si el tope del trabajo es menor que su peor caso (SC-005).
- **FR-014**: `CHANGELOG.md` lleva en *Unreleased* la entrada de `boe-legislacion` v0.3.1, que sustituye a v0.3: el cuerpo
  vuelve a decir cómo empieza la respuesta y de qué habla, y la regla de leer lo que se nombra. Criterio: la entrada está
  en la sección *Unreleased*.

### Key Entities

- **Cuerpo de `SKILL.md`**: lo que el agente carga entero al activar la skill; protocolo, reglas, formas fijas y tabla de
  comandos generada; como mucho 11 000 caracteres.
- **Referencia a mano**: `skills/boe-legislacion/references/<nombre>.md` sin cabecera «generado desde»; el agente la lee
  solo si un paso del cuerpo la pide y la sesión la necesita.
- **Lectura de una eval** (`lecturas`, H26.1): `{norma, bloque, motivo}`; un bloque que la caché de las sesiones de la
  skill sirve porque la eval lo provoca; no es comando, ni cita, ni cuenta en ningún umbral.
- **Entrada del manifiesto de grabaciones**: por norma, su búsqueda, sus `bloques` y su `para`; lo que graba
  `grabar_datos`.

## Uso, de fuera adentro

- **El cuerpo (FR-001 a FR-006)**: lo carga el agente una vez por pregunta que activa la skill, entero: como mucho 11 000
  caracteres, unos 4800 tokens, los mismos que hoy (10 992). No crece con el uso: lo escribe una persona y el gate lo
  acota. Lo gana cada respuesta: empieza por lo que se pregunta (US1) y no da materia de lo no leído (US2).
- **Las lecturas que añade la regla (FR-002, FR-003)**: las pide el agente en el paso 3, una vez por precepto que la
  respuesta va a nombrar y no ha leído, con `articulo`, o todas en una invocación de `articulos`, cada una con su
  comprobación detrás. Las acota la respuesta y el salto: solo lo que la respuesta nombra, remitido por un bloque leído en
  el paso 2 o salido de su índice, y nada de lo que remiten esas lecturas. En la eval 15, el art. 17 de la Ley 35/2006
  lleva a cinco bloques (`a7`, `a18`, `a19`, `a20`, `dadecimoctava`): en esa pregunta, cinco sobres más, cada uno del
  tamaño de un bloque (los cinco sobres suman 35 908 bytes, medidos tras grabarlos; research V11). Con cientos de normas
  y miles de bloques consultados no cambia: cada lectura se pide por su id y devuelve un bloque, no lo acumulado. Sus
  señales son las de toda lectura: los avisos de vigencia del bloque y, de la comprobación, `⚠ REDACCIÓN MODIFICADA:`
  solo si el bloque cambió desde una lectura anterior, que deja de darse al leerlo (lo que ya hace hoy el paso 2).
  Ninguna señal nueva.
- **Las `lecturas` de las evals y sus grabaciones (FR-009 a FR-011)**: las pide la preparación de la caché de cada sesión
  de `boe-legislacion`, una vez por sesión: nueve bloques más, ocho grabaciones nuevas de 1,7 KB (`a60`) a 393 KB (`a7`,
  con sus veinte redacciones). No crecen con el uso de kitlegal: solo con lo que se escribe en `evals/`.
- **El apartado de `docs/SKILLS.md` §1 (FR-012)**: lo lee quien escribe o reparte una skill, una vez por skill; un
  párrafo. Deja de valer cuando otra medida, en otra evidencia, diga otra cosa.
- **La entrada de `CHANGELOG.md` (FR-014)**: la lee quien actualiza kitlegal, una vez por versión.

## Success Criteria *(mandatory)*

### Measurable Outcomes

- **SC-001**: `make ci` en verde, con `make skills-check`, `ordenes-de-ejemplo` y los demás controles de contenido de
  `docs/SKILLS.md` §1 (FR-005 a FR-007). Lo hace cumplir el propio `make ci`, que sale en rojo con cualquiera de ellos.
- **SC-002**: El cuerpo de `SKILL.md` tiene como mucho 11 000 caracteres; la comprobación de las skills de `make ci` y
  `make skills-sync` salen en rojo con 11 001 o más (FR-005).
- **SC-003**: El cuerpo contiene los seis ejemplos de FR-001 y las reglas de FR-002 y FR-003 en sus pasos, y la tabla del
  plan da cada regla de los dos pasajes de `27aeccc` —su paso 5 entero y la última viñeta de su paso 4— con su frase,
  la de lo remitido y del precepto no leído con el cambio de FR-002 y FR-003; lo coteja la revisión final.
- **SC-004**: Los tests de `lecturas` de H26.1 con los bloques nuevos: cada uno se sirve en la caché de su eval, sin
  fecha y con `9999-12-31`, y la caché de las sesiones lo da sin red; esos tests salen en rojo en `make ci` si la caché
  no sirve uno de ellos (FR-011).
- **SC-005**: El tope del trabajo de evals de `boe-legislacion`, recalculado antes de cerrar el plan, cubre su peor caso;
  `TestDefinicionDelJob` sale en rojo si no lo cubre (FR-013).
- **SC-006**: `CHANGELOG.md` (*Unreleased*); y el job de evals en la propuesta de cambio, con las tres skills, que el
  workflow mide en el cierre sobre la cabeza de la rama (FR-014).
- **SC-007**: En el informe del job de cierre de `boe-legislacion`, `afirma_lo_no_leido:claude-sonnet-5-5` se cumple con
  el umbral de hoy, 0.028. El umbral decide: si la tasa pasa de 0.028, el veredicto del job es fallo y el job sale en
  rojo.
- **SC-008**: En el informe del job de cierre de `boe-legislacion`, ninguna sesión de `claude-sonnet-5-5` de las evals 14,
  15 y 17 tiene en `fuera_de_lo_grabado` una orden o llamada que sale con 4 o 5 porque la caché de la sesión no sirve un
  bloque de sus `lecturas` (umbral: 0 sesiones). No cuenta la orden o llamada `articulos` que sale con 5 porque pide,
  con bloques que la caché sirve, otro de la misma norma sin grabar: la salida se atribuye a ese bloque (en el cierre de
  H31, `boe_articulos BOE-A-2004-4214 a57 a59` salió entera con 5 por `a57`, con `a59` grabado). Lo hacen cumplir los
  tests de SC-004, que salen en rojo en `make ci` si la caché de la skill no sirve un bloque de `lecturas` sin fecha y
  con `9999-12-31`, y la preparación de la sesión, que da la falta y no la abre si no está grabado (FR-010, FR-011).
- **SC-009**: Los demás umbrales que deciden del informe del job de cierre de `boe-legislacion` se cumplen; el que no se
  cumple pone su veredicto en fallo.
- **SC-010**: `cuenta_su_proceso` se publica por modo, como hoy, sin cifra de aceptación en el cierre: no decide, y
  hacerla decidir pide medir antes su juez (ADR 0037).
- **SC-011**: `jurisprudencia` y `legal-core` siguen aprobadas: el veredicto de su job de cierre es aprobado, y si no lo
  es el job sale en rojo.
- **SC-012**: Lo que la entrada deja a una persona antes de fusionar: un sondeo de la eval 15 con la skill de la
  propuesta, como los de `evidencias/sondeos-20261025/` —el mismo arnés, 20 repeticiones por modo, Claude Code 2.1.289—,
  de 40 respuestas, en el que `afirma_lo_no_leido` dé como mucho 2 de 40, por debajo del control de esa evidencia, 16 de
  80, con Fisher unilateral p < 0,05 (2 de 40 frente a 16 de 80 da p ≈ 0,023, calculado en esta sesión), y
  `cuenta_su_proceso` dé como mucho 2 de 40. Lo mide una persona fuera del run (ADR 0032): el informe final lo da como
  pendiente. Si no se cumple, la propuesta no se fusiona y se decide con sus cifras; si se cumple solo
  `cuenta_su_proceso`, lo del cuerpo de la respuesta puede ir solo en su propia propuesta, y leer lo que se nombra vuelve
  al análisis.
- **SC-013**: `docs/SKILLS.md` §1 dice lo de FR-012 con sus cifras y su evidencia; lo coteja la revisión final.

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
`evidencias/sondeos-20261025/` —el mismo arnés, 20 repeticiones por modo, Claude Code 2.1.289—, de 40 respuestas, en el
que `afirma_lo_no_leido` dé como mucho 2 de 40, por debajo del control de esa evidencia, 16 de 80, con Fisher unilateral
p < 0,05, y `cuenta_su_proceso` dé como mucho 2 de 40, cerca de los 0 de 40 de la skill de antes de H31. El cierre mide 6
respuestas de la eval 15 y no distingue un 20 % de un 10 %; el sondeo, sí. Si no se cumple, la propuesta no se fusiona y
se decide con sus cifras; si se cumple solo `cuenta_su_proceso`, lo del cuerpo de la respuesta puede ir solo en su propia
propuesta, y leer lo que se nombra vuelve al análisis.»

## Assumptions

- Los ids de FR-009 están en los índices grabados de sus normas, como dice el hito; el plan lo comprueba contra lo
  grabado. Si `grabar_datos` no puede grabar uno porque la fuente no responde, es una causa mayor del run (dependencia
  externa inaccesible), no un caso del producto.
- `a24` de `BOE-A-2013-12887` ya está en el manifiesto y grabado (lo pide `data/plazos.yaml`, H9): la eval 17 lo lleva en
  `lecturas` igualmente, con su motivo, como pide el hito.
- La serie de cierres de `boe-legislacion` empieza de cero, porque su clave es el árbol de `skills/boe-legislacion` (ADR
  0038): es la consecuencia de cambiar el árbol, no algo que haga el hito.
- Las sesiones del job son las de la versión de Claude Code fijada en `.github/workflows/evals.yml`; el sondeo de SC-012,
  el de 2.1.289, como los de la evidencia.
- La versión de la skill solo se registra en `CHANGELOG.md`.

## Fuera de alcance

- Cualquier umbral, la rúbrica y los casos del juez, y que `cuenta_su_proceso` decida: subir un umbral pide un ADR, y
  hacer decidir una clase que solo se publica, otro.
- Que `kitlegal boe` devuelva la rúbrica de los preceptos a los que remite un bloque (backlog; `evidencias/adr-0038/`).
- `jurisprudencia`, `legal-core` y la tabla de comandos del cuerpo.
- La reorganización de las skills por capacidad (H32).
- La descripción de `boe-legislacion` (FR-008).
- Evals nuevas, y cambios en las evals de `boe-legislacion` que no sean las `lecturas` de FR-009.
- Cambios en el binario `kitlegal` (applets, `internal/source`, `internal/core`) y en el arnés de evals, salvo el tope del
  trabajo de FR-013 y lo que sus tests esperan.
- Subir el tope de una sesión de las evals (`topeDeUnaSesion`, 240 s), común o propio de `boe-legislacion`: no se
  implementa; sería otra entrada (el ADR 0016 lo ata a la vigencia de 300 s de `buscar` y `metadatos` en la caché).
- Limitar el paso 3 de hoy: la lectura de lo remitido que cambia la respuesta no entra en el salto de FR-003, ni se
  añade un segundo nivel de lecturas (índice → bloque → lo que remite).
- Lanzar el sondeo de SC-012 y versionar su evidencia: los hace una persona fuera del run.
- `docs/SKILLS.md` fuera de §1; el ADR 0046, la constitución, `CLAUDE.md` y `docs/ROADMAP.md`.
- Cambios en `scripts/workflow/`, `.specify/` e `internal/workflow/`, y en `web/`.
- La release y su etiqueta, y la fusión: las decide una persona.
