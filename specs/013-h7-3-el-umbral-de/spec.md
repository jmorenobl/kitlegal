# Feature Specification: H7.3 · El umbral de expresiones prohibidas decide, `boe-legislacion` sin el vocabulario que provoca el ruido, y evals en paralelo con un sondeo local

**Feature Branch**: `013-h7-3-el-umbral-de`

**Created**: 2026-09-29

**Status**: Draft

**Input**: Sección «#### H7.3 · El umbral de expresiones prohibidas decide, `boe-legislacion` sin el vocabulario que provoca el ruido, y evals en paralelo con un sondeo local» de `docs/ROADMAP.md`, en modo desatendido (ADR 0018).

## Resumen

H7.2 (PR #82) no cumple su criterio de éxito. En su cierre (informe del job sobre `fbdab2e`, `specs/012-h7-2-la-consulta-repetida/gates/evals/boe-legislacion.json`, clave `expresiones_prohibidas_por_modelo`), 10 de las 51 respuestas de Sonnet 5 en las evals que activan `boe-legislacion` (19,6 %) llevan una expresión prohibida, frente a un umbral de ≤ 5 % (2 de 51); Haiku 4.5, 0 de 30. Por eval: 03 (1), 06 (1), 13 (2), 14 (2), 15 (3) y 19 (1). Bitácora `docs/USO.md`, entrada del 2026-09-29 «H7.2 no cumple su umbral: el ruido cambia de palabras, y medirlo cuesta 40 minutos».

1. **El ruido tiene una forma, y es el vocabulario de `SKILL.md`.** Ocho de las diez respuestas empiezan por «Sin hallazgos que trasladar.»; otra, por «No hay hallazgos, así que respondo con el texto vigente.». La regla 7 de v0.1.2 empareja «con hallazgos o sin ellos» con «trasládalos», y el paso 5 enumera «los hallazgos» entre lo que la respuesta no nombra: el modelo repite lo que se le pide no decir.
2. **Un ejemplo copiado como dato.** La eval 19, sesión 02 de Sonnet 5, copió la fecha 20250101 del ejemplo de `⚠ REDACCIÓN MODIFICADA:` de `SKILL.md`, una redacción que el BOE no tiene.
3. **Nadie lo paró.** El job dio «aprobado»: la lista decide por serie (2 de 3, ADR 0016), ocho de las diez respuestas son de evals informativas (13, 14, 15 y 19) y el 5 % solo se publicaba.
4. **Medir cuesta 40 minutos.** El trabajo de `boe-legislacion` abre sus 93 sesiones una tras otra (33 min 22 s y 34 min 28 s de sesiones en las dos ejecuciones del cierre, unos 22 s cada una; 36 min 37 s y 42 min 33 s el trabajo entero) y exige Linux con strace y sudo.

H7.3 no entrega skill nueva: arregla la que existe y protege el instrumento que la mide (constitución, principio VIII). Seis piezas:

1. el informe del job publica `umbrales` (contrato del ADR 0029) y el de expresiones prohibidas de Sonnet 5 decide;
2. `boe-legislacion` v0.1.3, cuya prosa no enseña el vocabulario que la respuesta no puede decir;
3. una tercera familia en la lista de expresiones prohibidas, el anuncio de la respuesta;
4. el job abre varias sesiones a la vez, con un objetivo de duración que decide, sin confundir un límite de uso de la suscripción con una eval fallida y con una sola tanda de sesiones por commit;
5. un sondeo local para personas, que no es un veredicto;
6. un escenario de quickstart que comprueba que el sondeo reproduce lo que mide el job.

Va antes de H20 y de cualquier release: ni H7, ni H7.1 ni H7.2 han salido en ninguna.

**Cómo se citan los requisitos de otros hitos.** Los requisitos y criterios de H5, H7.1 y H7.2 se citan sin guion —«H7.2 SC 001», «H7.2 FR 054», «H5 FR 077»—, para no confundirlos con los de este spec. Los ids con guion (FR-001, SC-001…) son siempre de este spec.

## Clarifications

### Session 2026-09-29

- Q: En el sondeo, que no tiene traza de strace, ¿cómo cuentan en la tasa de cada serie los criterios que el juicio del job saca de la traza —los comandos esperados y los prohibidos de cada eval—, dado que `Juzgar` hoy exige que no falte ningún comando esperado para que una sesión pase? → A: El sondeo no juzga los criterios que salen de la traza: los comandos esperados de cada eval (incluida la comprobación de la redacción) y los comandos prohibidos. Cada sesión pasa o falla por el resto de lo que juzga `Juzgar`: que terminó, la activación, las citas, los avisos, las formas de hallazgo, el territorio y las expresiones prohibidas. Todo eso sale del transcript y de `codigo-de-la-sesion` (`internal/evals/sesion.go`: `SkillsActivadas`, `Respuesta`, `Terminada`), y se juzga con el mismo código, sin un juez propio del sondeo. La salida nombra estos criterios entre lo que no comprueba (qué órdenes se ejecutaron: los comandos esperados y los prohibidos de cada eval), junto a las llegadas a la red y la ausencia de Python (FR-065). Por eso, sobre las mismas sesiones, la tasa de cada serie nunca es menor que la del job, y no es la del job. El test de FR-096 compara eval a eval todo el resultado del juicio salvo lo que sale de la traza: comandos ejecutados y ausentes, prohibidos ejecutados, invocaciones, fuera de lo grabado, otras fallidas y llegadas a la red, además de los motivos y el `pasa` que dependen de ellos. El plan fija cómo se aplica `Juzgar` sin esos criterios, por ejemplo con la eval sin comandos ni prohibidos, que según `Juzgar` deja su juicio «el de antes». (auto: criterio a; fuente: docs/ROADMAP.md, H7.3, Alcance, «Sondeo local para personas»: «publica la tasa de cada serie» y, en «Lo que no comprueba, lo dice», «Lo que el job lee de la traza de strace —qué órdenes se ejecutaron y las llegadas a la red— y la ausencia de Python quedan sin comprobar, y su salida los nombra: por eso sus tasas no son las del job»; spec FR-061, FR-065 y FR-096; internal/evals/juzgar.go: `repartirComandos` y `anotarProhibidos` leen `sesion.Invocaciones`; internal/evals/sesion.go: las invocaciones solo vienen de la traza (`LeerTrazas`))
- Q: ¿Qué credencial acepta el sondeo y cómo decide que «sirve» antes de abrir ninguna sesión, si no puede cargar la configuración de Claude Code de quien lo lanza? → A: El sondeo acepta una sola credencial: `CLAUDE_CODE_OAUTH_TOKEN` en el entorno, como el job, el token de la suscripción que da `claude setup-token`. Es la única que pasa a sus sesiones: no usa ninguna clave de API de pago por uso ni lee la sesión iniciada de Claude Code de la persona, es decir, ni su llavero ni sus ficheros de credenciales. Antes de abrir ninguna sesión basta con comprobar que la variable está y no está vacía. Si falta o está vacía, el sondeo dice que falta la credencial, nombra la variable y sale con un código distinto de 0 sin abrir ninguna sesión (FR-063, US4-5). Esa comprobación no consume ninguna llamada al modelo. Una credencial caducada o revocada se ve en la primera sesión: la sesión no termina y la salida la publica con su motivo, como el job. No hay comprobación previa contra el servicio. (auto: conservadora; criterio d; fuente: docs/ROADMAP.md, H7.3, «Sondeo local para personas»: «la misma preparación […] que el job», «usa su credencial, y sin una que sirva lo dice antes de abrir ninguna sesión»; spec FR-063 y US4-5; CONTRIBUTING.md, job de evals: «el secreto de repositorio `CLAUDE_CODE_OAUTH_TOKEN`, el token de la suscripción de Claude que da `claude setup-token` (el proyecto no usa una clave de API de pago por uso)»; ADR 0016; .github/workflows/evals.yml y scripts/evals.sh; constitución, «Criterio de decisión autónoma», punto 4)
- Q: ¿Dice el sondeo, sesión a sesión, qué expresiones prohibidas encontró, y conserva al terminar su directorio temporal con los transcripts y las respuestas? → A: El sondeo publica solo lo agregado que pide FR-065: la tasa de cada serie, el recuento de respuestas con alguna expresión prohibida sobre las respuestas en evals que activan la skill (con el 5 % como referencia), lo que no comprueba y las sesiones sin medir por límite de uso. No lista las expresiones encontradas en cada sesión. Su directorio temporal (sesiones, cachés preparadas, transcripts, respuestas y el estado de Claude Code de cada sesión) se borra al terminar, también cuando termina con un código distinto de 0. No deja nada en disco, salvo las cachés de compilación de Go (FR-064). Quien necesite cada respuesta y sus expresiones las tiene en el informe del job, que publica las expresiones por sesión (H7.2 FR 053). (auto: criterio c; fuente: spec FR-064, FR-065 y FR-066; spec, «Uso, de fuera adentro», fila «La salida del sondeo»; spec, «Fuera de alcance»: «Que el sondeo escriba un informe en otro formato, guarde un historial de sondeos o compare dos sondeos»; constitución, «Gates», criterio de uso, y «Criterio de decisión autónoma», puntos 1 y 2)
- Q: Cuando una sesión del sondeo termina con el mensaje del límite de uso de la cuenta, ¿qué hace el sondeo con las sesiones que aún no ha abierto y con qué código sale? → A: Como el job (FR-044). Tras una sesión cuyo resultado lleva el mensaje del límite de uso (el de la ventana, el semanal o el de la familia del modelo, que no se reponen mientras dura el sondeo), el sondeo no abre ninguna sesión más. Las que ya estaban abiertas terminan y se juzgan, y las que faltaban quedan sin medir por límite de uso. Su serie queda sin medir, y la salida publica esas sesiones como sin medir por límite de uso (FR-065). Sale con 0: ha abierto y juzgado lo que la cuenta le dejaba, y clasificado el resto con la misma regla que el job. Su código de salida no dice nada de la skill (FR-066). Tras reintentos por `rate_limit` agotados, o cortados por el tope (FR-040, b y c), sigue abriendo las demás, como el job. (auto: criterio c; fuente: spec FR-040, FR-044, FR-061, FR-065 y FR-066; spec, «Edge Cases»: «Una sesión del sondeo que la cuenta no deja terminar»; docs/ROADMAP.md, H7.3: «la misma preparación y el mismo juez» y «Tras un límite que no se repone […] el job no abre más sesiones, y las que faltan quedan también sin medir»)
- Q: ¿Acepta el sondeo un solo modelo por ejecución, como dicen FR-060 y FR-067, o varios, como sugiere la tabla de uso («≤ 19 evals × los modelos pedidos»)? → A: Un solo modelo por sondeo, en un único argumento que no puede estar vacío (FR-060, FR-067). La salida da las series de ese modelo y su recuento de expresiones prohibidas. La fila «La salida del sondeo» de «Uso, de fuera adentro» se corrige a «una línea por serie (≤ 19 evals)», sin «× los modelos pedidos». Para medir otro modelo se lanza otro sondeo. (auto: criterio a; fuente: docs/ROADMAP.md, H7.3, «Sondeo local para personas»: «`make evals-sondeo SKILL=<skill> EVALS=<números> MODELO=<id> REPETICIONES=<n>` […] abre en paralelo las sesiones de ese subconjunto de evals con ese modelo»; spec FR-060 y FR-067; aceptación del hito y FR-070: el quickstart usa solo Sonnet 5)

### Cierre 2026-09-30

- Q: El job de cierre sobre `6ab3add` no cumple SC-001 (5 de 51 respuestas de Sonnet 5 con una expresión de la lista, 8 de 51 con el párrafo de transición; la eval 06, 1 de 3). ¿Cambia de sitio la comprobación de la redacción (FR-015), y qué se mantiene de lo que decidió H7.1? → A: Sí. Con la prosa ya sin el vocabulario, el párrafo sigue tras la última orden, que es siempre la comprobación, y no aparece donde la última orden trae lo que la respuesta cita (`legal-core`, 0 de 18 en tres cierres; evals 18 y 19, 0 de 6): la causa es el sitio, y «hallazgos» sale ahora de la salida del binario (FR-014, su segunda rama). La comprobación pasa a ir en la misma orden que la lectura, detrás de ella y con su misma norma y sus mismos bloques (`kitlegal boe articulo … --json && kitlegal graph check … --json`). Se mantiene de H7.1: después de leer, antes de responder, con la norma y los bloques leídos, nunca sin argumentos ni antes de leer, y cada bloque una sola lectura. Cambia: «una vez por norma citada» pasa a «una vez por cada orden que lee bloques», que es lo mismo en la mayoría de las preguntas (un bloque, o varios pedidos juntos) y una comprobación más cuando una remisión lleva a leer después otro bloque de la misma norma; cada bloque se comprueba una sola vez y la cota de H7.1 (≤ k señales con k bloques leídos) no cambia. Donde FR-015 y «Uso, de fuera adentro» dicen «una vez por norma citada» o «una comprobación por norma citada», vale esto. Además, la respuesta se define por su posición (todo lo escrito tras la última orden) y salen de la prosa «Antes de responder», «antes de la respuesta» y «ni anuncies que vas a responder». Ni la lista ni el umbral cambian (FR-008). (auto: criterio a; fuente: `gates/evals/boe-legislacion.json` y `gates/evals/legal-core.json` de este hito y los de H7.1 y H7.2; docs/ROADMAP.md, H7.3, «La comprobación, en su sitio» y «La prosa no enseña el vocabulario prohibido»: «Si el research concluye […] que el modelo repite la palabra de la salida y no la de la prosa, lo dice, y el cambio va a cómo la skill pide y lee esa salida, sin tocar el binario»; spec FR-014 y FR-015; research, «Cierre (2026-09-30)…»; contracts/skill-boe-legislacion.md §8; alternativa descartada: la comprobación solo en la orden del último bloque de cada norma, que conserva «una vez por norma» al pie de la letra pero puede dejar un bloque sin comprobar)

## Criterios del hito, literales

Transcripción literal de `docs/ROADMAP.md` §4, H7.3. El Objetivo y el Alcance se reflejan en cada requisito.

- **Entrega**: «(1) el informe del job publica `umbrales` (ADR 0029) y el de expresiones prohibidas de Sonnet 5 decide: por encima del 5 %, veredicto `fallo` y job en rojo; (2) `boe-legislacion` v0.1.3, cuya prosa no enseña el vocabulario que la respuesta no puede decir, con la regla 7 reescrita y ejemplos sin datos que copiar; (3) la lista de expresiones prohibidas recoge también el anuncio de la respuesta, para que el ruido no pase el umbral cambiando de palabras; (4) el job abre varias sesiones a la vez, con un objetivo de duración que decide, sin confundir un límite de uso de la suscripción con una eval fallida y con una sola tanda de sesiones por commit; (5) un sondeo local para personas, un objetivo de `make` que funciona en macOS sin strace ni sudo, con la misma preparación y el mismo juez que el job, y que dice que no es un veredicto; (6) un escenario de quickstart que comprueba que el sondeo reproduce lo que mide el job.»
- **Controles**:
  - «`make ci` en verde, con `skills-check` y `schema-check` sin drift y las reglas del conjunto;»
  - «la sección «Controles de umbral» del plan (ADR 0029) con dos filas: el 5 % de Sonnet 5, en `evals:boe-legislacion:expresiones_prohibidas:claude-sonnet-5`, y los 900 s, en `evals:boe-legislacion:duracion_de_las_sesiones`. El de Haiku 4.5 no tiene fila, porque no es un umbral del hito;»
  - «tests del informe con sesiones sintéticas: con 3 de 51 el umbral de Sonnet 5 no se cumple y el veredicto es `fallo` con su motivo, y con 2 de 51 se cumple y el veredicto no cambia; `cumple` es la comparación rehecha; el de Haiku 4.5 incumplido no cambia el veredicto; `umbrales` es `[]` en una skill sin lista; y la duración, por encima y por debajo de 900 s;»
  - «tests de los límites con transcripts sintéticos: el mensaje de límite de uso, reintentos por `rate_limit` hasta agotarlos y hasta el tope, y un 429 del que la sesión se recupera, que se mide y publica sus reintentos; la serie sin medir; el motivo propio; y que tras un límite que no se repone no se abren más sesiones;»
  - «un test de la ejecución en paralelo sin modelo, con el `claude` sustituto, que corre en `make ci` también en macOS, sin strace ni sudo (lo que reparte las sesiones se prueba aparte de las comprobaciones previas del job, que exigen Linux): nunca más de `CONCURRENCIA_DE_EVALS` sesiones a la vez, cada una en sus directorios, y el mismo informe que en serie; y una comprobación de la definición del job de que un segundo disparo sobre el mismo commit no abre otra tanda a la vez;»
  - «una comprobación en `make ci` de que la prosa de `SKILL.md`, fuera del código y de la región generada, no lleva ninguna expresión de la lista, y de que `SKILL.md` no lleva ninguna fecha `AAAAMMDD` escrita con cifras;»
  - «el calibrado de la lista con la familia nueva sobre los informes de H7.1 y H7.2 (35 y 10, eval por eval), y las comprobaciones de H7.2 sobre los bloques grabados y las formas de la skill, extendidas a ella;»
  - «tests del sondeo con el `claude` sustituto: sobre las mismas sesiones da, eval a eval, lo mismo que el juicio del job en todo lo que no sale de la traza; su salida dice que no es un veredicto y qué no comprueba; sale con 0 con tasas en rojo; y no escribe fuera de su directorio temporal;»
  - «`CHANGELOG.md` (*Unreleased*);»
  - «y el job de evals en la propuesta de cambio.»
- **Aceptación**:
  - «en el informe del job de cierre de `boe-legislacion`, `umbrales` lleva `expresiones_prohibidas:claude-sonnet-5` con `decide: true` y `cumple: true` (como mucho 2 de 51, frente a 10 de 51 en H7.2), el de Haiku 4.5 con `decide: false`, y `duracion_de_las_sesiones` con `decide: true` y `cumple: true` (≤ 900 s, frente a unos 2 000 s en H7.2); ninguna sesión queda sin medir por un límite; el veredicto es aprobado (el hito cambia `SKILL.md`, Definition of Done §1.10) y `red` está vacío; y el informe final da los dos umbrales como «comprobado por su control»;»
  - «el escenario del quickstart, que ejecuta la persona al leer el informe final (necesita modelo, así que es una medida sin control; ADR 0029): el sondeo, con Sonnet 5, tres repeticiones y las evals 03, 06, 13, 14 y 15 —las de nueve de las diez respuestas con ruido de H7.2; 15 respuestas—, en un Mac sin strace. Con la `SKILL.md` de `main` (v0.1.2), en una copia de trabajo desechable de la rama, marca al menos 5 de 15 (el job de cierre de H7.2 dio 9 de 15 en esas evals); con la del hito, como mucho 2 de 15, compatible con el job de cierre de H7.3. El quickstart anota cuánto tarda cada sondeo.»

Trazabilidad resumida (el detalle, en cada requisito):

| Criterio del hito | Dónde se cumple |
|---|---|
| Entrega (1), `umbrales` y el de Sonnet 5 que decide | FR-001 a FR-008, US2 |
| Entrega (2), `boe-legislacion` v0.1.3 | FR-010 a FR-018, US1 |
| Entrega (3), la tercera familia de la lista | FR-020 a FR-024, US1 |
| Entrega (4), el job en paralelo, los límites, la duración y una tanda por commit | FR-030 a FR-037, FR-040 a FR-044, FR-050 a FR-052, US3 |
| Entrega (5), el sondeo local | FR-060 a FR-068, US4 |
| Entrega (6), el quickstart | FR-070, US5, SC-002 |
| Relación con H7.2 | FR-080 |
| Controles | FR-090 a FR-099, SC-003 a SC-010 |
| Aceptación | SC-001, SC-002 |

## Relación con H7.2 y H5

`specs/012-h7-2-la-consulta-repetida/` no se edita: es el registro de su run. No cambia ninguna decisión de arquitectura —el veredicto sigue siendo del job, con Sonnet 5 en tres repeticiones y Haiku 4.5 informativo (ADR 0016), y el contrato de `umbrales` es el del ADR 0029—, así que no hay ADR nuevo (FR-080).

**Sustituye** (lo que dice H7.2 deja de valer y vale lo de este spec):

- H7.2 SC 001, en lo que dice de Haiku 4.5 (≤ 1 de 30): FR-004. En lo que dice de Sonnet 5, lo hace cumplir FR-002 y FR-003, y lo mide SC-001.
- La viñeta de «Fuera de alcance» de H7.2 «Que el umbral del 5 % por modelo entre en el veredicto del job […] el umbral se comprueba en el recuento que publica el informe (SC-001)»: FR-002 y FR-003.
- H7.2 FR 040, en cuanto pedía que `SKILL.md` enumerase lo que la respuesta no nombra (la memoria de consultas, `graph check`, los códigos de salida, los hallazgos…): FR-010 y FR-011. Lo que la respuesta no lleva sigue igual y lo mide la lista.
- H7.2 FR 043 (la regla 7): FR-012.
- H7.2 FR 042, en su ejemplo con fechas: FR-013. La forma y las fechas tal como las da el binario se quedan.
- H7.2 FR 047 y FR 056, que acotaban lo que cambiaban `SKILL.md` e `internal/evals` en aquel hito: los acota este spec (FR-016, FR-017 y «Fuera de alcance»).
- H7.2 FR 084 y FR 085 (calibrado y bloques grabados): los extienden FR-021 y FR-022.
- Del contrato del job de H5 (`specs/006-h5-skill-boe-legislacion/contracts/job-de-evals.md` §3.2), que las sesiones se abran una tras otra: FR-030. Todo lo que garantiza cada sesión se queda (FR-031).

**Se queda**: H7.2 FR 054 (la lista decide en las 10 positivas con 2 de 3 por serie): mide otra cosa que el umbral agregado y hacen falta los dos (ADR 0029, opción 2); H7.2 FR 053 (expresiones por sesión y recuento por modelo, sin la sesión de la prueba de red); y H5 FR 077 (la sesión ve la skill tal como la deja `make install`; `SKILL.md` no nombra evals, el job ni modelos). Todo lo demás de H7.2 y H5 sigue en vigor.

## User Scenarios & Testing *(mandatory)*

### User Story 1 - Quien pregunta con Sonnet 5 lee la norma, sin el estado de la comprobación delante (Priority: P1)

Una persona pregunta a `boe-legislacion`, con Sonnet 5, por un artículo. La respuesta empieza por lo que pregunta: el texto citado, su cita, sus avisos de vigencia y, solo si la redacción cambió desde una lectura anterior, la línea `⚠ REDACCIÓN MODIFICADA:` con las dos fechas que da el binario. No abre con un párrafo que cuente que no hay nada que contar, ni con otras palabras.

**Why this priority**: es lo que el hito dice que gana la skill, y el defecto medido que llegaría a una de cada cinco respuestas de la release.

**Independent Test**: el umbral de Sonnet 5 en el job (SC-001); en `make ci`, la comprobación de la prosa de `SKILL.md` (SC-005) y el calibrado de la lista (SC-003).

**Acceptance Scenarios**:

1. **Dado** `SKILL.md` v0.1.3, **Cuando** se comprueba su texto fuera del código y de la región generada, **Entonces** no lleva ninguna expresión de ninguna de las tres familias de la lista (FR-010).
2. **Dado** `SKILL.md` v0.1.3, **Cuando** se busca en todo el fichero una fecha de ocho cifras, **Entonces** no hay ninguna, y el ejemplo de `⚠ REDACCIÓN MODIFICADA:` lleva dos marcadores `AAAAMMDD` (FR-013).
3. **Dado** una respuesta que empieza por «Nada que trasladar. Ya puedo responder.», **Cuando** se juzga, **Entonces** se encuentra una expresión de la familia del anuncio y la sesión no pasa (FR-020, FR-024).
4. **Dado** una respuesta con «No hay avisos de vigencia sobre este bloque» y nada más de la lista, **Cuando** se juzga, **Entonces** no se encuentra ninguna expresión (FR-020).
5. **Dado** los informes versionados de H7.1 y H7.2, **Cuando** se les aplica la lista con la tercera familia, **Entonces** marca exactamente 35 y 10 respuestas, con el reparto por eval de FR-021, y ninguna otra.

---

### User Story 2 - El umbral del 5 % lo hace cumplir el job, no la lectura de su informe (Priority: P1)

El informe del job publica sus umbrales con el contrato del ADR 0029. Si más del 5 % de las respuestas de Sonnet 5 en las evals que activan `boe-legislacion` llevan una expresión prohibida, el veredicto es `fallo` y el job sale en rojo, aunque todas las series pasen con 2 de 3 y aunque las respuestas sean de evals informativas.

**Why this priority**: sin él, el defecto de H7.2 vuelve a salir en verde en cualquier propuesta de cambio.

**Independent Test**: tests del informe con sesiones sintéticas (SC-006); el job de cierre (SC-001).

**Acceptance Scenarios**:

1. **Dado** un informe de `boe-legislacion` con 3 de 51 respuestas de Sonnet 5 con alguna expresión y todas las series que deciden pasando, **Cuando** se escribe el informe, **Entonces** `expresiones_prohibidas:claude-sonnet-5` lleva `medida` 3, `total` 51, `comparacion` `"<="`, `umbral` 0.05, `cumple: false` y `decide: true`, el veredicto es `fallo` con un motivo que nombra el umbral, la medida y el total, y el job sale en rojo (FR-002, FR-003).
2. **Dado** lo mismo con 2 de 51, **Cuando** se escribe el informe, **Entonces** el umbral lleva `cumple: true` y el veredicto es el mismo que sin el umbral (FR-003).
3. **Dado** un informe con 2 de 30 respuestas de Haiku 4.5 con alguna expresión, **Cuando** se escribe, **Entonces** `expresiones_prohibidas:claude-haiku-4-5-20251001` lleva `cumple: false` y `decide: false`, y el veredicto no cambia (FR-004).
4. **Dado** cualquier informe, **Cuando** se rehace la comparación de cada elemento de `umbrales` con su `medida`, su `total`, su `comparacion` y su `umbral`, **Entonces** da su `cumple` (FR-001).
5. **Dado** el informe de `legal-core`, **Cuando** se escribe, **Entonces** `umbrales` es `[]` (FR-006).
6. **Dado** cualquier informe, **Cuando** se lee `informe.md`, **Entonces** muestra la tabla de `umbrales` junto al recuento de expresiones prohibidas por modelo (FR-001).

---

### User Story 3 - El job mide en minutos, con las mismas garantías, sin confundir un límite con una eval fallida (Priority: P1)

El job abre las sesiones de `boe-legislacion` de cuatro en cuatro y termina sus sesiones en unos 9 minutos. Cada sesión sigue preparada y aislada como antes. Si la cuenta alcanza un límite, el informe lo dice como tal, con su propio motivo, y no como un defecto de la skill. Un segundo disparo sobre el mismo commit no dobla el consumo.

**Why this priority**: cada ajuste de la skill cuesta hoy 40 minutos de job, y un job que va más rápido sin distinguir los límites daría fallos de la skill que no lo son.

**Independent Test**: el test de la ejecución en paralelo con el `claude` sustituto y la comprobación de la definición del job en `make ci` (SC-008); los tests de los límites (SC-007) y de la duración (SC-006); el job de cierre (SC-001).

**Acceptance Scenarios**:

1. **Dado** un plan de 93 sesiones y `CONCURRENCIA_DE_EVALS` 4, con el `claude` sustituto y sin límites de uso, **Cuando** se ejecutan, **Entonces** en ningún momento hay más de 4 sesiones abiertas, ninguna escribe en un fichero o directorio que escriba otra, y el informe es el mismo que con `CONCURRENCIA_DE_EVALS` 1, salvo los tiempos (FR-030, FR-032).
2. **Dado** una sesión cuyo transcript lleva en su resultado el mensaje del límite de uso de la cuenta, **Cuando** se juzga, **Entonces** queda sin medir por límite de uso, su serie queda sin medir, el veredicto es `fallo` con un motivo propio que nombra el límite y las sesiones, y el job no abre ninguna sesión más; las que faltaban del plan quedan sin medir (FR-040, FR-042, FR-043, FR-044).
3. **Dado** una sesión que reintenta por `rate_limit` hasta agotar los reintentos, o hasta que el tope de 240 s la corta, **Cuando** se juzga, **Entonces** queda sin medir por límite de uso, con el mismo motivo propio, y el job sigue abriendo las demás sesiones (FR-040, FR-044).
4. **Dado** una sesión que recibe un 429 por `rate_limit`, reintenta y termina con su resultado, **Cuando** se juzga, **Entonces** se mide como cualquier otra y el informe publica sus reintentos, en la sesión y en el total (FR-033, FR-041).
5. **Dado** un informe de `boe-legislacion` con `duracion_de_las_sesiones` de 901 s, **Cuando** se escribe, **Entonces** el umbral lleva `cumple: false` y `decide: true` y el veredicto es `fallo` con un motivo propio que no es de la skill; con 900 s, `cumple: true` y el veredicto no cambia (FR-050, FR-051).
6. **Dado** la definición del job, **Cuando** se comprueba en `make ci`, **Entonces** un segundo disparo del flujo sobre el mismo commit mientras el primero sigue no abre otra tanda de sesiones a la vez ni deja una comprobación —cancelada o saltada— que el cierre lea en lugar de la de la tanda que corre o que le haga dejar de esperarla (FR-034).

---

### User Story 4 - Una persona prueba un cambio de `SKILL.md` en su Mac en minutos, sabiendo que no es el veredicto (Priority: P2)

Quien ajusta la skill lanza un objetivo de `make` con unas pocas evals, un solo modelo y unas repeticiones. En macOS o en Linux, sin strace ni sudo, abre las sesiones en paralelo con el binario y la skill de su árbol de trabajo, sin su configuración de Claude Code y con el `CLAUDE_CODE_OAUTH_TOKEN` de su entorno, y le dice la tasa de cada serie y cuántas respuestas llevan una expresión prohibida. La salida empieza diciendo que no es un veredicto y qué no comprueba.

**Why this priority**: es lo que hace que medir un cambio cueste minutos y no cuarenta; el veredicto sigue siendo del job.

**Independent Test**: los tests del sondeo con el `claude` sustituto en `make ci` (SC-009); el escenario del quickstart (SC-002).

**Acceptance Scenarios**:

1. **Dado** unas sesiones fijadas del `claude` sustituto, **Cuando** las juzgan el sondeo y el juicio del job, **Entonces** dan, eval a eval, lo mismo en todo lo que no sale de la traza (FR-061).
2. **Dado** cualquier sondeo, **Cuando** se lee su salida, **Entonces** su primera línea dice que no es un veredicto, y la salida nombra lo que no comprueba: qué órdenes se ejecutaron (los comandos esperados y los prohibidos de cada eval), las llegadas a la red y la ausencia de Python (FR-065).
3. **Dado** un sondeo en el que todas las series quedan por debajo de su umbral y el recuento pasa del 5 %, **Cuando** termina, **Entonces** sale con 0 y no escribe `informe.json` ni ningún veredicto (FR-066).
4. **Dado** un sondeo, **Cuando** termina, **Entonces** no ha escrito nada fuera de su directorio temporal, salvo las cachés de compilación de Go (FR-064).
5. **Dado** una persona sin `CLAUDE_CODE_OAUTH_TOKEN` en el entorno, o con la variable vacía, **Cuando** lanza el sondeo, **Entonces** dice que falta la credencial, nombra la variable y sale con un código distinto de 0 sin abrir ninguna sesión (FR-063).
6. **Dado** una sesión del sondeo cuyo resultado lleva el mensaje del límite de uso de la cuenta, **Cuando** el sondeo la juzga, **Entonces** no abre ninguna sesión más, las que faltaban quedan sin medir por límite de uso, las ya abiertas terminan y se juzgan, y sale con 0 (FR-066).

---

### User Story 5 - Cualquiera comprueba que el sondeo reproduce lo que mide el job (Priority: P3)

Al leer el informe final, la persona ejecuta el sondeo en un Mac sin strace sobre las evals en que H7.2 tuvo el ruido, primero con la `SKILL.md` de `main` y después con la del hito.

**Why this priority**: da confianza en el sondeo antes de usarlo para iterar; no bloquea nada del run.

**Independent Test**: el escenario del quickstart (FR-070, SC-002), fuera de `make ci` y del run.

**Acceptance Scenarios**:

1. **Dado** una copia de trabajo desechable de la rama con la `SKILL.md` de `main` (v0.1.2), **Cuando** se lanza el sondeo con Sonnet 5, tres repeticiones y las evals 03, 06, 13, 14 y 15, **Entonces** marca al menos 5 de las 15 respuestas, y el quickstart anota cuánto tardó (SC-002).
2. **Dado** la rama del hito, **Cuando** se lanza el mismo sondeo, **Entonces** marca como mucho 2 de 15, y el quickstart anota cuánto tardó (SC-002).

---

### Edge Cases

- **3 de 51 y 2 de 51**: 3/51 ≈ 0,0588 > 0,05, no cumple; 2/51 ≈ 0,0392, cumple (FR-002, FR-003).
- **La sesión de la prueba de red** (etiqueta `evals-prueba-de-red` o entrada `prueba_de_red`): no entra en la medida ni en el total del umbral de expresiones prohibidas (H7.2 FR 053), y sí en la concurrencia y en la duración (FR-030, FR-050). La ejecución de cierre no la lleva.
- **Sesiones sin medir y el umbral de expresiones**: no entran ni en la medida ni en el total (FR-002); con todas las de Sonnet 5 sin medir, el total es 0 y la proporción 0 (contrato del ADR 0029), pero el veredicto ya es `fallo` por las sesiones sin medir (FR-043).
- **Un límite a mitad del job**: las sesiones ya abiertas terminan y se juzgan; tras un límite de uso (FR-040, a) no se abren más, y tras reintentos agotados por `rate_limit` (FR-040, b y c) sí (FR-044).
- **Un segundo disparo después de que el primero ha terminado**: abre su tanda, que ya no es simultánea (FR-034).
- **Unas evals que no activan la skill en el sondeo**: se juzgan como en el job; no entran en el recuento de expresiones prohibidas (FR-065).
- **Una sesión del sondeo que la cuenta no deja terminar**: se publica sin medir por límite de uso, como en el job; tras el mensaje del límite de uso no se abren más y las que faltan quedan sin medir, y el sondeo sigue saliendo con 0 si ha podido abrir y juzgar sus sesiones (FR-061, FR-066).
- **Una credencial caducada o revocada en el sondeo**: no se detecta antes; la primera sesión no termina y se publica con su motivo, como en el job (FR-063, FR-065).
- **Un valor de `CONCURRENCIA_DE_EVALS` o un argumento del sondeo que no es válido**: salen con un código distinto de 0 antes de abrir ninguna sesión (FR-030, FR-067).
- **Lo dicho con otras palabras** que ninguna familia recoge: no se detecta (limitación declarada de H7.2); el informe publica cada respuesta.

## Requirements *(mandatory)*

### Functional Requirements

#### El umbral decide

- **FR-001**: `informe.json` MUST llevar siempre la clave `umbrales`, con los campos, las invariantes y la comparación del «Contrato de umbrales del informe del job de evals» del ADR 0029: `nombre`, `descripcion`, `medida`, `total` (opcional), `comparacion`, `umbral`, `cumple` (la comparación calculada en coma flotante de doble precisión, sin redondeos) y `decide`. `informe.md` MUST mostrar la tabla de `umbrales` junto al recuento de expresiones prohibidas por modelo. Comprobable: en todo informe que escribe el job, rehacer la comparación de cada elemento da su `cumple`.
- **FR-002**: En `boe-legislacion`, `umbrales` MUST llevar `expresiones_prohibidas:claude-sonnet-5`: `medida`, las respuestas de Sonnet 5 con alguna expresión prohibida; `total`, las respuestas de Sonnet 5 en las evals que activan la skill, las informativas incluidas; los dos sobre las sesiones medidas de las series que pide el plan, sin la de la prueba de red (H7.2 FR 053) y sin las sesiones sin medir (FR-040); `comparacion` `"<="`, `umbral` 0.05 y `decide: true`. Con el conjunto actual (17 evals que activan la skill por 3 repeticiones), `total` es 51.
- **FR-003**: Un umbral con `decide: true` y `cumple: false` MUST poner el veredicto en `fallo`, con un motivo que nombra el umbral, la medida y el total, y el job MUST salir en rojo. Uno que se cumple MUST NOT cambiar el veredicto. Comprobable: con 3 de 51, `fallo` y el job en rojo aunque todas las series pasen; con 2 de 51, el veredicto de siempre.
- **FR-004**: En `boe-legislacion`, `umbrales` MUST llevar `expresiones_prohibidas:claude-haiku-4-5-20251001`, con la medida y el umbral de FR-002 sobre las respuestas de Haiku 4.5 (con el conjunto actual, 30) y `decide: false`: incumplido, no cambia el veredicto. Haiku 4.5 es el modelo informativo del ADR 0016; este umbral no es un umbral del hito, no tiene fila en «Controles de umbral» y el informe final lo muestra como «solo se publica».
- **FR-005**: En `boe-legislacion`, `umbrales` MUST llevar `duracion_de_las_sesiones` (FR-050, FR-051).
- **FR-006**: Una skill sin lista de expresiones prohibidas no tiene umbrales de expresiones, y una sin objetivo de duración no tiene el de duración. `legal-core` no tiene ninguno de los dos: su `umbrales` MUST ser `[]`.
- **FR-007**: La regla por serie del ADR 0016 y H7.2 FR 054 MUST seguir como están: las evals informativas no deciden por su serie; lo que entra en el umbral es cada respuesta.
- **FR-008**: Ninguna corrección del run MUST cumplir un umbral de FR-002 o FR-051 rebajándolo, retirándolo, dejándolo en `decide: false`, sacando las evals informativas del total o recortando la lista para que case menos (ADR 0029).

#### `boe-legislacion` v0.1.3

- **FR-010**: El texto de `skills/boe-legislacion/SKILL.md` —frontmatter incluido— MUST NOT llevar ninguna expresión de ninguna familia de la lista (`hallazgo`, `memoria de consultas`, `graph check`, `código 3`, `que trasladar`…), comparada como en H7.2 FR 051, fuera de tres sitios: las órdenes que el agente ejecuta, los nombres y valores que lee en la salida del binario (en código: `version-obsoleta`, `data.hallazgos`, un código de salida como `3`…) y la región generada de la tabla de comandos. Comprobable: FR-091.
- **FR-011**: El protocolo MUST describir lo que la skill comprueba y lo que dice a quien pregunta con palabras que la respuesta puede usar, y MUST NOT nombrar el caso en que la comprobación de la redacción no da nada que decir: ninguna frase lo trata como un resultado que haya que trasladar, callar o anunciar. Lo que v0.1.2 pide para ese caso (H7.2 FR 041: no afirmar que lo consultado antes cambió ni que no) MUST decirse desde el caso en que la respuesta lleva `⚠ REDACCIÓN MODIFICADA:` —fuera de él, nada sobre lo consultado antes—. Comprobable: ninguna frase de `SKILL.md` se refiere a ese caso; su efecto lo mide SC-001.
- **FR-012**: La regla 7 MUST reescribirse, título incluido, para decir: (a) que cuando la comprobación de la redacción termina con un código distinto de 0, la skill responde con el texto leído y la frase fija de v0.1.2, «No se ha podido comprobar si la redacción ha cambiado desde una consulta anterior.», sin afirmar que ha cambiado ni que no; y (b) cuándo lleva la respuesta `⚠ REDACCIÓN MODIFICADA:`. MUST NOT nombrar el resultado vacío ni emparejarlo con ninguna acción («con … o sin ellos», «trasládalos»).
- **FR-013**: Los ejemplos de formas que llevan datos del binario MUST usar marcadores (`AAAAMMDD`), no fechas: el de `⚠ REDACCIÓN MODIFICADA:` MUST llevar sus dos fechas como dos marcadores `AAAAMMDD`. `SKILL.md` MUST NOT llevar ninguna fecha `AAAAMMDD` escrita con cifras (ocho cifras seguidas, sin otra cifra delante ni detrás, con un mes de 01 a 12 y un día de 01 a 31). Los ejemplos de cita pueden seguir siendo citas reales.
- **FR-014**: El research MUST partir del de H7.2 («Causa de raíz del ruido») y del cierre de H7.2 —la forma dominante y dónde la enseña `SKILL.md` v0.1.2, línea a línea— y trazar cada cambio de `SKILL.md` a su causa. Si concluye, con la evidencia de los informes, que el modelo repite la palabra de la salida del binario y no la de la prosa, MUST decirlo, y el cambio va a cómo la skill pide y lee esa salida, sin tocar el binario. Una prohibición más no satisface este requisito. Comprobable: el research da la causa con esas medidas y el plan traza a ella cada cambio.
- **FR-015**: Si el research encuentra que la causa está en que la comprobación de la redacción es la última orden antes de responder, el plan MUST decidir con esa evidencia si cambia su sitio en el protocolo y registrarlo en «Decisiones», manteniendo lo que decidió H7.1: una vez por norma citada, después de leer sus bloques y antes de responder.
- **FR-016**: MUST quedarse de v0.1.2: la forma `⚠ REDACCIÓN MODIFICADA:` con sus dos fechas, tal como las da el binario, en la misma línea; la forma de la cita y la de los avisos de vigencia; el protocolo de cinco pasos; lo que dicen las reglas 1 a 6 y «Nada de otra conversación», cuya redacción cambia solo donde lleva vocabulario de la lista (la regla 2 dice «código 3», la 6 «la memoria de consultas»); y lo que la respuesta no dice (H7.2 FR 040, FR 041 y FR 044), que lo mide la lista.
- **FR-017**: `SKILL.md` MUST seguir por debajo de 300 líneas, con frontmatter válido, la región generada intacta (`make skills-check` sin drift) y sin nombrar evals, el job ni modelos (H5 FR 077).
- **FR-018**: `CHANGELOG.md` (*Unreleased*) MUST registrar `boe-legislacion` v0.1.3 con lo que cambia para quien la usa.

#### La lista, contra el desplazamiento

- **FR-020**: `evals/boe-legislacion/expresiones-prohibidas.yaml` MUST ganar una tercera familia, el anuncio de la respuesta y del estado de lo comprobado, con las formas medidas en los cierres de H7.1 y H7.2 («ya puedo responder», «ya tengo todo lo necesario», «redacto la respuesta», «que trasladar»…); el plan fija la lista. MUST NOT entrar «No hay avisos de vigencia sobre este bloque» ni frases que digan a quien lee que la norma no está derogada o que no tiene avisos: son derecho, no maquinaria.
- **FR-021**: Con las tres familias, la lista MUST marcar, comparada como en H7.2 FR 051, exactamente 35 de las 93 respuestas del informe de H7.1 (`specs/011-h7-1-graph-check-acotado/gates/evals/boe-legislacion.json`) —02 (1), 03 (3), 04 (3), 05 (3), 06 (3), 07 (3), 08 (2), 09 (2), 13 (3), 14 (3), 15 (3), 16 (2), 17 (3) y 19 (1)— y exactamente 10 de las 93 del de H7.2 (`specs/012-h7-2-la-consulta-repetida/gates/evals/boe-legislacion.json`) —03 (1), 06 (1), 13 (2), 14 (2), 15 (3) y 19 (1)—, eval por eval, y ninguna otra.
- **FR-022**: Ninguna expresión de la tercera familia MUST aparecer, comparada como en H7.2 FR 051, en el texto de los bloques grabados para las evals de `boe-legislacion` ni en las formas que la skill enseña a escribir (la cita, los avisos de vigencia, `⚠ REDACCIÓN MODIFICADA:` y la frase fija de la regla 7): las comprobaciones de H7.2 FR 085 y de las formas de la skill se extienden a ella.
- **FR-023**: La tercera familia MUST entrar en `schemas/expresiones-prohibidas.yaml.json` por una tarea `[datos]`; una lista que no cumple el esquema sigue siendo un fichero mal formado (H7.2 FR 055).
- **FR-024**: La tercera familia MUST aplicarse como las otras dos: en `Juzgar`, a cada sesión de cada eval que activa la skill (H7.2 FR 052), en la regla por serie (H7.2 FR 054), en las expresiones por sesión y el recuento por modelo (H7.2 FR 053) y en los umbrales (FR-002, FR-004).

#### El job, en paralelo y con las mismas garantías

- **FR-030**: `scripts/evals.sh` MUST abrir las sesiones de una skill varias a la vez, como mucho `CONCURRENCIA_DE_EVALS`, fijada en la definición del job como las repeticiones: 4 en `boe-legislacion` y 1 en `legal-core`. Un valor que no es un entero ≥ 1 MUST terminar el guion con código 1 antes de la primera sesión, como las repeticiones. Comprobable: FR-094.
- **FR-031**: MUST NOT cambiar nada de lo que garantiza cada sesión (contrato del job de H5 y ADR 0016): se prepara justo antes de abrirla, con su directorio de trabajo, su caché y su grafo; corre bajo strace, con el proxy que rechaza toda petición salvo la del modelo, sin Python accesible y con el tope de 240 s; y el informe exige todas las sesiones del plan, contando como tales las que quedan sin medir (FR-040, FR-044).
- **FR-032**: Ninguna sesión MUST escribir en un fichero o directorio que escriba otra, tampoco en el estado de Claude Code; cada una ve la skill tal como la deja `make install` (H5 FR 077). El juicio de cada sesión depende solo de su transcript y de su traza, y el orden en que terminan no cambia el informe, salvo los tiempos y, tras un límite de uso (FR-044), qué sesiones se llegaron a abrir.
- **FR-033**: El informe MUST publicar, por sesión y en total, los reintentos por límite de ritmo que registra su transcript (el evento `system/api_retry` de `stream-json` con `error` `rate_limit`).
- **FR-034**: Un segundo disparo del flujo de evals sobre el mismo commit mientras el primero sigue MUST NOT abrir otra tanda de sesiones a la vez; el plan elige si el segundo espera o se salta. Elija lo que elija, el segundo disparo MUST NOT dejar ninguna comprobación —cancelada, saltada ni de ningún otro estado— que el cierre del workflow lea en lugar de la de la tanda que corre, ni que le haga dejar de esperarla: `scripts/workflow/cierre.sh` (`medir`) espera mientras haya comprobaciones pendientes y cuenta como roja una cancelada, y `recoger_evals` copia el informe de las de `evals` que acaban en `pass` o `fail`. El cierre espera a la tanda que corre y recoge su informe. Comprobable: FR-094.
- **FR-035**: El `timeout-minutes` de cada trabajo del job MUST seguir siendo el tope que corta un cuelgue y MUST cubrir el peor caso de su skill —todas las sesiones del plan, con la de la prueba de red, en su tope de 240 s y el margen de cierre forzoso, con su concurrencia, más la preparación de cada sesión, la del runner (hasta 8 min en el cierre de H7.2) y el informe—, para que el informe se escriba siempre; el plan fija el cálculo y los tiempos de preparación y de informe con que lo hace, a partir de lo medido en el cierre de H7.2. No es el control del objetivo de duración (FR-051). Comprobable: la comprobación de la definición del job (FR-094) falla si el tope de un trabajo queda por debajo de ese peor caso.
- **FR-036**: Lo que reparte las sesiones MUST poder probarse sin strace, sin sudo, sin modelo y también en macOS, aparte de las comprobaciones previas del job, que exigen Linux.
- **FR-037**: El consumo de la suscripción MUST seguir siendo el número de sesiones del plan: ninguna sesión se reintenta ni se duplica.

#### Un límite no es una eval fallida

- **FR-040**: Una sesión MUST quedar **sin medir por límite de uso**, y no como sesión fallida de su eval, cuando su transcript registra que no terminó por un límite de la cuenta: (a) el mensaje del límite de uso en su resultado; (b) reintentos por `rate_limit` hasta agotarlos; o (c) reintentos por `rate_limit` hasta que la corta el tope de 240 s.
- **FR-041**: Una sesión con reintentos por `rate_limit` que termina con su resultado MUST medirse como cualquier otra y publicar sus reintentos (FR-033).
- **FR-042**: Una serie con alguna sesión sin medir MUST quedar sin medir: ni pasa ni falla, y el informe lo dice en su tasa.
- **FR-043**: Con alguna sesión sin medir, el veredicto MUST ser `fallo`, con un motivo propio que nombra el límite y las sesiones, que se distingue sin modelo de los motivos de la skill; el job sale en rojo.
- **FR-044**: Tras una sesión con el mensaje del límite de uso (FR-040, a: el de la ventana, el semanal o el de la familia del modelo, que no se reponen dentro del job), el job MUST NOT abrir ninguna sesión más; las que ya estaban abiertas terminan y se juzgan, y las que faltaban del plan quedan sin medir por límite de uso. Tras reintentos agotados (FR-040, b y c), el job sigue abriendo las demás.

#### Duración objetivo, con su control

- **FR-050**: El informe MUST publicar `duracion_de_las_sesiones`: los segundos desde que se prepara la primera sesión hasta que termina la última, sin la preparación del runner (instalar, retirar Python).
- **FR-051**: En `boe-legislacion`, `duracion_de_las_sesiones` MUST ser un elemento de `umbrales` con `umbral` 900, `comparacion` `"<="` y `decide: true`, sin `total`; incumplido, el veredicto es `fallo` con un motivo propio, como el de FR-043, que nombra el umbral y la medida y se distingue sin modelo de los motivos de la skill, y el job sale en rojo. `legal-core` no tiene objetivo de duración.
- **FR-052**: El `timeout-minutes` del trabajo MUST NOT ser el control del objetivo: un trabajo cancelado no escribe informe (FR-035).

#### Sondeo local para personas

- **FR-060**: Un objetivo de `make`, que aparece en `make help`, MUST abrir en paralelo las sesiones de un subconjunto de evals de una skill con un solo modelo, dado en un único argumento, y unas repeticiones (el plan fija los nombres de los argumentos y sus valores por defecto), con la concurrencia del job para esa skill por defecto, en macOS y en Linux, sin strace ni sudo. Para medir otro modelo se lanza otro sondeo.
- **FR-061**: Cada sesión MUST prepararse con `TestPrepararSesion`, como en el job, y juzgarse con el mismo código que el job (`Juzgar`, la lista, el recuento y la clasificación de los límites de FR-040), sin un juez propio del sondeo. Como no tiene traza de strace, MUST NOT juzgar los criterios que salen de ella —los comandos esperados de cada eval, incluida la comprobación de la redacción, y los comandos prohibidos—: cada sesión pasa o falla por el resto de lo que juzga `Juzgar` (que terminó, la activación, las citas, los avisos, las formas de hallazgo, el territorio y las expresiones prohibidas), que sale del transcript y del código de la sesión (`SkillsActivadas`, `Respuesta`, `Terminada`). Por eso, sobre las mismas sesiones, la tasa de cada serie nunca es menor que la del job y no es la del job. El plan fija cómo se aplica `Juzgar` sin esos criterios. Comprobable: sobre las mismas sesiones da, eval a eval, lo mismo que el juicio del job en todo lo que no sale de la traza —el resultado entero salvo los comandos ejecutados y ausentes, los prohibidos ejecutados, las invocaciones, lo fuera de lo grabado, las otras fallidas y las llegadas a la red, y los motivos y el `pasa` que dependen de ellos— (FR-096).
- **FR-062**: MUST ejecutar el binario y la skill del árbol de trabajo, no los que la persona tenga instalados, con la caché preparada, las peticiones a las fuentes rechazadas como en el job y el tope de 240 s por sesión.
- **FR-063**: MUST NOT cargar la configuración de Claude Code de quien lo lanza —su `CLAUDE.md`, sus hooks, sus plugins ni sus skills—. Acepta una sola credencial, `CLAUDE_CODE_OAUTH_TOKEN` en el entorno, como el job (el token de la suscripción que da `claude setup-token`), y es la única que pasa a sus sesiones: MUST NOT usar una clave de API de pago por uso ni leer la sesión iniciada de Claude Code de la persona (su llavero o sus ficheros de credenciales). Antes de abrir ninguna sesión basta con que la variable esté y no esté vacía; si falta o está vacía, MUST decir que falta la credencial, nombrar la variable y salir con un código distinto de 0 sin abrir ninguna sesión. Esa comprobación MUST NOT consumir ninguna llamada al modelo, y MUST NOT haber otra previa contra el servicio: una credencial caducada o revocada se ve en la primera sesión, que no termina y se publica con su motivo, como en el job.
- **FR-064**: MUST NOT escribir nada fuera de un directorio temporal, salvo las cachés de compilación de Go. Ese directorio —sesiones, cachés preparadas, transcripts, respuestas y el estado de Claude Code de cada sesión— MUST borrarse al terminar, también cuando termina con un código distinto de 0.
- **FR-065**: Su salida MUST empezar diciendo que no es un veredicto; MUST nombrar lo que no comprueba —lo que el job lee de la traza de strace (qué órdenes se ejecutaron: los comandos esperados y los prohibidos de cada eval, y las llegadas a la red) y la ausencia de Python— y que por eso sus tasas no son las del job; y MUST publicar solo lo agregado: la tasa de cada serie y el recuento de respuestas con alguna expresión prohibida sobre las respuestas en evals que activan la skill, con el 5 % como referencia, las sesiones sin medir por límite de uso y cada sesión que no terminó por otra causa, con su motivo, como el job (una credencial caducada o revocada, un modelo que no existe). MUST NOT listar las expresiones encontradas en cada sesión (las publica el informe del job, H7.2 FR 053). La salida da las series de su modelo y su recuento.
- **FR-066**: MUST NOT escribir `informe.json` ni ningún veredicto, y su código de salida MUST NOT decir si la skill pasa: 0 si ha podido abrir y juzgar las sesiones, sean cuales sean sus tasas. Tras una sesión con el mensaje del límite de uso, MUST NOT abrir ninguna más, como el job (FR-044): las ya abiertas terminan y se juzgan, las que faltaban quedan sin medir por límite de uso, su serie queda sin medir y sale con 0; tras reintentos por `rate_limit` agotados o cortados por el tope (FR-040, b y c) sigue abriendo las demás.
- **FR-067**: Un argumento que no designa evals existentes de la skill, un modelo vacío o unas repeticiones que no son un entero ≥ 1 MUST hacerlo salir con un código distinto de 0 antes de abrir ninguna sesión, nombrando el argumento.
- **FR-068**: Ninguna tarea ni ningún paso del workflow MUST ejecutarlo: abre sesiones con modelo y consume la suscripción de quien lo lanza. En `make ci` se prueba con un `claude` sustituto que escribe transcripts fijados sin abrir ninguna sesión (FR-096).

#### Escenario de quickstart

- **FR-070**: `quickstart.md` MUST tener un escenario que ejecuta el sondeo con Sonnet 5, tres repeticiones y las evals 03, 06, 13, 14 y 15 (15 respuestas), en un Mac sin strace: primero con la `SKILL.md` de `main` (v0.1.2), en una copia de trabajo desechable de la rama, y después con la del hito; anota cuánto tarda cada sondeo. Queda fuera de `make ci` y del run —necesita modelo— y lo ejecuta la persona al leer el informe final. Mide SC-002.

#### Relación con otros hitos

- **FR-080**: `specs/012-h7-2-la-consulta-repetida/` MUST NOT editarse; este spec nombra lo que sustituye de H7.2 y de H5 («Relación con H7.2 y H5»). No hay ADR nuevo.

#### Controles y Definition of Done

- **FR-090**: `make ci` MUST quedar en verde, con `skills-check` y `schema-check` sin drift y las reglas del conjunto.
- **FR-091**: Una comprobación en `make ci` MUST fallar si el texto de `SKILL.md` de `boe-legislacion`, fuera del código (spans en línea y bloques delimitados) y de la región generada, lleva alguna expresión de la lista (FR-010), o si `SKILL.md` lleva alguna fecha `AAAAMMDD` escrita con cifras (FR-013).
- **FR-092**: Los tests del informe con sesiones sintéticas MUST cubrir: 3 de 51 con Sonnet 5, que da `cumple: false` y el veredicto `fallo` con su motivo; 2 de 51, con `cumple: true` y el veredicto sin cambiar; que `cumple` es la comparación rehecha; que el de Haiku 4.5 incumplido no cambia el veredicto; `umbrales` `[]` en una skill sin lista; y la duración por encima y por debajo de 900 s.
- **FR-093**: Los tests de los límites con transcripts sintéticos MUST cubrir: el mensaje de límite de uso; reintentos por `rate_limit` hasta agotarlos y hasta el tope; un 429 del que la sesión se recupera, que se mide y publica sus reintentos; la serie sin medir; el motivo propio; y que tras un límite que no se repone no se abren más sesiones.
- **FR-094**: Un test de la ejecución en paralelo sin modelo, con el `claude` sustituto, que corre en `make ci` también en macOS, sin strace ni sudo, MUST comprobar que nunca hay más de `CONCURRENCIA_DE_EVALS` sesiones a la vez, que cada una usa sus directorios y que, sin límites de uso, el informe es el mismo que en serie, salvo los tiempos. Una comprobación de la definición del job en `make ci` MUST fallar si un segundo disparo sobre el mismo commit puede abrir otra tanda a la vez o dejar una comprobación que el cierre lea en lugar de la de la tanda que corre, o que le haga dejar de esperarla (FR-034), si la concurrencia de cada skill no es la de FR-030, o si el tope de un trabajo no cubre su peor caso (FR-035).
- **FR-095**: El calibrado de FR-021 y las comprobaciones de FR-022 MUST correr en `make ci` y fallar si la lista marca otra cosa.
- **FR-096**: Los tests del sondeo con el `claude` sustituto MUST cubrir: que sobre las mismas sesiones da, eval a eval, lo mismo que el juicio del job en todo lo que no sale de la traza; que su salida dice que no es un veredicto y qué no comprueba (los comandos esperados y prohibidos, las llegadas a la red y Python), y no lista las expresiones por sesión; que sale con 0 con tasas en rojo; que tras el mensaje del límite de uso no abre más sesiones, deja las que faltan sin medir y sale con 0; que sin `CLAUDE_CODE_OAUTH_TOKEN`, o con la variable vacía, sale con un código distinto de 0 sin abrir ninguna sesión; que un modelo vacío lo hace salir sin abrirlas; y que no escribe fuera de su directorio temporal y lo borra al terminar, también con un código distinto de 0.
- **FR-097**: `CHANGELOG.md` (*Unreleased*) MUST registrar el umbral que decide, `boe-legislacion` v0.1.3 (FR-018), la tercera familia de la lista, el job en paralelo con los límites y la duración, y el sondeo.
- **FR-098**: El job de evals MUST ejecutarse en la propuesta de cambio (lo hace el workflow tras la revisión final), y su informe da lo que pide SC-001.
- **FR-099**: La sección «Controles de umbral» del plan MUST tener dos filas: FR-002 (el 5 % de Sonnet 5), en `evals:boe-legislacion:expresiones_prohibidas:claude-sonnet-5`, y FR-051 (los 900 s), en `evals:boe-legislacion:duracion_de_las_sesiones`. El de Haiku 4.5 (FR-004) no tiene fila.

### Key Entities

- **Umbral del informe**: un elemento de `umbrales` con el contrato del ADR 0029 (nombre, descripción, medida, total opcional, comparación, umbral, cumple, decide).
- **Sesión sin medir por límite de uso**: una sesión que no terminó por un límite de la cuenta (FR-040); no pasa ni falla, deja su serie sin medir y pone el veredicto en `fallo` con un motivo propio.
- **Reintentos por límite de ritmo**: los eventos de reintento por `rate_limit` del transcript de una sesión; se publican por sesión y en total.
- **Duración de las sesiones**: los segundos entre la preparación de la primera sesión y el final de la última.
- **Concurrencia de evals**: cuántas sesiones de una skill se abren a la vez; fijada por skill en la definición del job.
- **Familia del anuncio**: la tercera familia de la lista de expresiones prohibidas de `boe-legislacion`.
- **Sondeo**: una ejecución local de un subconjunto de evals, con la preparación y el juez del job, sin traza ni veredicto.

## Success Criteria *(mandatory)*

### Measurable Outcomes

- **SC-001**: En el informe del job de cierre de `boe-legislacion`, `umbrales` lleva `expresiones_prohibidas:claude-sonnet-5` con `decide: true` y `cumple: true` (como mucho 2 de 51, frente a 10 de 51 en H7.2), el de Haiku 4.5 con `decide: false`, y `duracion_de_las_sesiones` con `decide: true` y `cumple: true` (≤ 900 s, frente a unos 2 000 s en H7.2); ninguna sesión queda sin medir por un límite; el veredicto es aprobado (el hito cambia `SKILL.md`, Definition of Done §1.10) y `red` está vacío; y el informe final da los dos umbrales como «comprobado por su control». **Control**: el job de evals de `boe-legislacion` sale en rojo, con el veredicto `fallo`, si pasan más de 2 de 51 (FR-003), si la duración pasa de 900 s (FR-051) o si alguna sesión queda sin medir (FR-043); el cierre del workflow lo cuenta como rojo.
- **SC-002**: En el escenario del quickstart, el sondeo con Sonnet 5, tres repeticiones y las evals 03, 06, 13, 14 y 15 marca al menos 5 de 15 respuestas con la `SKILL.md` de `main` (v0.1.2) y como mucho 2 de 15 con la del hito, y el quickstart anota cuánto tarda cada sondeo. Lo mide la persona: es una medida sin control (ADR 0029).
- **SC-003**: Sobre los informes versionados, la lista con las tres familias marca exactamente 35 de las 93 respuestas de H7.1 y 10 de las 93 de H7.2, con el reparto por eval de FR-021, y 0 de las demás. **Control**: el calibrado falla en `make ci` (FR-095).
- **SC-004**: 0 expresiones de la lista en el texto de los bloques grabados para las evals de `boe-legislacion` y en las formas que la skill enseña a escribir. **Control**: la comprobación de FR-022 falla en `make ci` (FR-095).
- **SC-005**: 0 expresiones de la lista en el texto de `SKILL.md` fuera del código y de la región generada, y 0 fechas `AAAAMMDD` escritas con cifras en `SKILL.md`. **Control**: la comprobación de FR-091 falla en `make ci`.
- **SC-006**: Los tests del informe distinguen los casos de FR-092: con 3 de 51, `fallo`; con 2 de 51, el veredicto sin cambiar; con 901 s, `fallo`; con 900 s, sin cambiar. **Control**: fallan en `make ci`.
- **SC-007**: Los tests de los límites distinguen los casos de FR-093. **Control**: fallan en `make ci`.
- **SC-008**: Con el `claude` sustituto, 0 instantes con más de `CONCURRENCIA_DE_EVALS` sesiones abiertas, 0 ficheros o directorios escritos por dos sesiones, y, sin límites de uso, el mismo informe que en serie salvo los tiempos; la definición del job no permite una segunda tanda simultánea sobre el mismo commit ni una comprobación que el cierre lea en lugar de la de la tanda que corre o que le haga dejar de esperarla, fija 4 y 1, y su tope cubre el peor caso. **Control**: el test y la comprobación de FR-094 fallan en `make ci`.
- **SC-009**: Los tests del sondeo distinguen los casos de FR-096, y el sondeo escribe 0 ficheros fuera de su directorio temporal salvo las cachés de compilación de Go, y deja 0 ficheros suyos al terminar. **Control**: fallan en `make ci`.
- **SC-010**: `make ci` en verde, con `skills-check` y `schema-check` sin drift y con las reglas del conjunto; `SKILL.md` de `boe-legislacion` por debajo de 300 líneas; y la entrada de `CHANGELOG.md` (*Unreleased*). **Control**: `make ci` (`skills-check` falla con 300 líneas o más).

## Uso, de fuera adentro

Cada salida del hito, desde quien la consume (criterio de uso, ADR 0028). El hito no cambia el binario: lo que la skill lee de la comprobación de la redacción sigue acotado como en H7.1 (una comprobación por norma citada, ≤ k señales con k bloques leídos, ≤ 3 800 bytes con cinco bloques cambiados y ≈ 300 sin cambios, con cualquier volumen). Volumen de referencia: meses de uso diario, cientos de normas y miles de bloques consultados, la mayoría hace más de una semana. Nada de lo que entrega el job o el sondeo depende de ese volumen: cada sesión empieza de un estado preparado.

| Salida | Quién la pide, cuántas veces y qué hace con ella | Tamaño | Cuándo deja de darse cada señal |
|---|---|---|---|
| La respuesta de `boe-legislacion` | La persona que pregunta; una por pregunta. Lee el texto citado, sus avisos y, si cambió, la línea `⚠ REDACCIÓN MODIFICADA:` (FR-011, FR-016). | Lo que ocupa la norma, y 0 líneas sobre la comprobación; una línea de ≈ 160 bytes por bloque cuya redacción cambió —0 en la mayoría de las preguntas, como mucho k × 160 con k bloques leídos—. Con cientos de normas y miles de bloques consultados es la misma: no cuenta lo acumulado. | Como en H7.1 y H7.2: la línea sale en la respuesta de la lectura que ve la redacción nueva y no en la siguiente sobre ese artículo; al cabo de un mes, solo si el BOE publica otra redacción. |
| `SKILL.md` v0.1.3 | El modelo, una vez por conversación en que se activa la skill. | < 300 líneas (FR-017). | No da señales. |
| `umbrales` del informe | El job, que decide con ellos el veredicto (FR-003, FR-051); el informe final del workflow, que los lee sin modelo (ADR 0029); la persona. Una vez por job y skill. | 3 elementos de ≈ 250 bytes en `boe-legislacion`, `[]` en `legal-core`; fijo, no crece con el uso. | Cada job los mide de nuevo sobre su commit; un umbral incumplido deja de darse en el primer job que lo cumple. |
| Sesiones sin medir y su motivo | La persona que lee el informe y el cierre del workflow, que no deben cambiar la skill por él (FR-043); una vez por job. | 0 lo habitual; como mucho las 94 sesiones del plan con la prueba de red, ≈ 60 bytes cada una, ≤ 6 KB. | Solo en ese job; el siguiente lo vuelve a medir. |
| Reintentos por `rate_limit` | Quien ajusta `CONCURRENCIA_DE_EVALS` en la definición del job; una vez por job. | Un entero por sesión (≤ 94) y un total, ≤ 2 KB. | Cada job los mide de nuevo. |
| `duracion_de_las_sesiones` | El job (umbral que decide en `boe-legislacion`) y la persona; una vez por job. | Un número. | Cada job la mide de nuevo. |
| La salida del sondeo | Quien ajusta `SKILL.md`, antes de pedir el job; una vez por sondeo. | Una línea por serie (≤ 19 evals; un solo modelo por sondeo), el recuento, lo que no comprueba y una línea por sesión sin medir o que no terminó, con su motivo (a lo sumo una por sesión pedida): con 5 evals y todas terminadas, unas 15 líneas, < 3 KB. | Cada sondeo mide de nuevo; no deja veredicto ni informe (FR-066). |
| La lista con la tercera familia | El job, al juzgar cada sesión de las evals que activan la skill; el sondeo; el calibrado y las comprobaciones de `make ci`. | Fija, del orden de decenas de expresiones; no crece con el uso. | No da señales. |

## Fuera de alcance

Del hito, literal:

- «cambiar el binario de `kitlegal` (applets, la salida de `graph check`, la descripción de `--describe`): cambian `internal/evals`, `scripts/evals.sh`, `.github/workflows/evals.yml` y el `Makefile`, además de la skill y sus evals;»
- «que el umbral de Haiku 4.5 decida, o cambiar el modelo que decide, las repeticiones o la regla por serie (ADR 0016);»
- «promover a decisoria ninguna eval informativa, o añadir evals;»
- «extender la lista o el umbral a `legal-core` (ninguna de sus 18 respuestas lleva ruido en los cierres de H7, H7.1 y H7.2);»
- «juzgar la redacción libre con un modelo o por similitud (H5.1, *Decisión del mecanismo*), incluidas las paráfrasis que la lista no recoge, que siguen siendo la limitación declarada de H7.2;»
- «que la eval de la consulta repetida compruebe las fechas de su línea `⚠ REDACCIÓN MODIFICADA:`: la causa del dato copiado, un ejemplo con una fecha inventada, se retira y se vigila en `make ci`, y las fechas las comprueba el escenario del quickstart de H7.2;»
- «que el sondeo decida algo, se ejecute en CI o en el workflow, o compruebe lo que exige strace;»
- «más runners o uno propio, y acelerar la preparación del runner;»
- «la cobertura de Codecov (pendiente del ADR 0029);»
- «editar el spec, el plan, los guiones o las grabaciones de H7, H7.1 y H7.2.»

De lo que el hito no especifica (constitución, «Criterio de decisión autónoma», punto 2):

- Cambiar el workflow `hito`, `scripts/workflow/` (también `cierre.sh`) o el informe final: una tanda por commit (FR-034) se resuelve en la definición del job.
- Ajustar la concurrencia sola, según los reintentos, o subirla por encima de 4 y 1: cambiarla es un cambio de una línea de la definición del job, con los reintentos publicados como dato (FR-033).
- Reintentar dentro del job una sesión sin medir, o esperar a que se reponga un límite.
- Tratar como límite de la cuenta otros reintentos (sobrecarga, 529) o fallos de la sesión que el transcript no registra como límite (FR-040).
- Un umbral de duración para `legal-core` o para el sondeo.
- Que el sondeo escriba un informe en otro formato, guarde un historial de sondeos o compare dos sondeos.
- Que el sondeo juzgue los comandos esperados y los prohibidos con las órdenes que el agente pide en el transcript, o publique las series con comandos sin tasa: esos criterios quedan sin juzgar y la salida lo dice (FR-061, FR-065).
- Que el sondeo acepte otra credencial que `CLAUDE_CODE_OAUTH_TOKEN` (una clave de API de pago por uso, la sesión iniciada de Claude Code de la persona, su llavero o sus ficheros de credenciales) o haga una comprobación previa de la credencial contra el servicio (FR-063).
- Que el sondeo liste las expresiones prohibidas de cada sesión, o conserve su directorio temporal con los transcripts y las respuestas y su ruta (FR-064, FR-065).
- Que el sondeo siga abriendo sesiones tras el mensaje del límite de uso, o salga con un código distinto de 0 por él (FR-066).
- Que el sondeo acepte varios modelos en una ejecución, con series y recuento por modelo (FR-060, FR-067).
- Documentar el sondeo fuera de `make help`, del quickstart y de `CHANGELOG.md`.
- Cambiar las formas de la cita, de los avisos de vigencia o de `⚠ REDACCIÓN MODIFICADA:`, o la frase fija de la regla 7.
- Grabaciones nuevas: el hito no las necesita.

## Assumptions

- La evidencia es la de los informes versionados de H7.1 y H7.2 (93 respuestas cada uno) y la de la entrada de `docs/USO.md` del 2026-09-29; las duraciones de H7.2 salen de los registros de sus dos ejecuciones de cierre.
- Con el conjunto actual, `boe-legislacion` abre 93 sesiones (19 evals × 3 con Sonnet 5 y las 12 no informativas × 3 con Haiku 4.5), 94 con la prueba de red; las evals que activan la skill dan 51 respuestas de Sonnet 5 y 30 de Haiku 4.5. `legal-core` abre 18.
- La suscripción no publica un tope de sesiones simultáneas; lo medido con esta cuenta son dos ejecuciones del job a la vez en los cierres de H7, H7.1 y H7.2 (dos sesiones de `boe-legislacion` durante más de media hora y cuatro durante los minutos de `legal-core`), con las 111 sesiones de cada ejecución de H7.2 terminadas con `result success`. Con 4 y 1 el pico es de 5 sesiones y las de `boe-legislacion` bajan de unos 34 min a unos 9.
- La versión de Claude Code fijada en el job registra en `stream-json` los reintentos por límite de ritmo (`system/api_retry` con `error` `rate_limit`) y el mensaje del límite de uso en el resultado; el research lo comprueba y fija sus formas.
- Quien lanza el sondeo tiene Go, Claude Code y `CLAUDE_CODE_OAUTH_TOKEN` de su suscripción en el entorno (`claude setup-token`).
- Los nombres técnicos que aparecen (`Juzgar`, `TestPrepararSesion`, `CONCURRENCIA_DE_EVALS`, `scripts/evals.sh`, `.github/workflows/evals.yml`, `schemas/expresiones-prohibidas.yaml.json`, los nombres de los umbrales) los fija el hito o existen en el repositorio; las expresiones exactas de la tercera familia, el nombre del objetivo del sondeo y sus argumentos, la forma de los motivos propios, cómo se aísla el estado de Claude Code de cada sesión y cómo evita el flujo la segunda tanda son del plan.
