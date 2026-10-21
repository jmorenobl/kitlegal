# Feature Specification: H26.1 · Las evals sirven lo que la skill lleva a leer y hacen fallar a propósito la lectura que quieren probar

**Feature Branch**: `025-h26-1-las-evals-sirven`

**Created**: 2026-10-21

**Status**: Draft

**Input**: la sección «H26.1 · Las evals sirven lo que la skill lleva a leer y hacen fallar a propósito la lectura que
quieren probar (ADR 0043)» de `docs/ROADMAP.md`, la única entrada humana del run (ADR 0018).

## Contexto

Las sesiones de evals corren sin red, con una caché preparada que hoy solo sirve las consultas que nombran las evals de
su skill: sus comandos esperados, sus citas y el índice y los metadatos de sus normas. Cualquier otra lectura sale con 5,
como con el BOE caído, aunque esté grabada. En el job de la #175, `cobertura_sin_declarar:claude-sonnet-5-5:herramienta`
de `legal-core`, que decide con 0, dio 1 de 6 porque el bloque `daprimera` de `BOE-A-2015-10565`, grabado desde H9, no se
sirvió; y en `boe-legislacion` una de las tres respuestas marcadas por `afirma_lo_no_leido:claude-sonnet-5-5` (3 de 120,
con 3,36 de tope) salió de los arts. 26 y 27 de la Ley 40/2015, sin grabar, en la eval 16. Y al revés: si la caché sirve
`daprimera`, ninguna eval comprueba ya que la línea `⚠ NO CUBIERTO:` aguanta una lectura fallida, que `legal-core`
promete en su paso 7 y en su regla 9. La clasificación de cada lectura fuera de lo grabado de los jobs de la #175 y de la
#178 está en el ADR 0043.

El hito no entrega skill nueva: protege `legal-core` y `boe-legislacion` (principio VIII). Lo que gana cada una: que el
verde y el rojo de sus evals dejen de depender de lo que la caché preparada deja leer; que la robustez de `legal-core`
ante una lectura fallida (pasos 7 y 10, regla 9) se mida a propósito y no por azar; y que el informe del job enseñe sin
buscarlo en el registro un hueco nuevo de lo grabado y el motivo de un código 1 que no es un defecto de kitlegal. Va
antes de H9.1, que cierra midiendo `legal-core` con umbrales de 0 y añade a `data/plazos.yaml` preceptos que, con este
hito, se sirven solos.

## Clarifications

### Session 2026-10-21

- Q: En la eval 11 de `legal-core` (FR-017), ¿qué pasa con el juicio de una sesión cuya respuesta lleva la línea `⚠ REDACCIÓN A FECHA:` del bloque `a124` de `BOE-A-2015-10565`, que la caché de esa eval no sirve? → A: Nada: «sin `⚠ REDACCIÓN A FECHA:`» solo dice que la eval no lleva `redaccion_fechada`; ningún control del job mira esa línea en la 11 (ni campo ni regla nueva ni motivo propio), y FR-029 se mide en el job por la línea `⚠ PLAZO:` con `2026-10-13` y por la lectura de `a124` intentada. El resto de FR-029 queda en el texto del paso 10 de la skill, que juzga la revisión final. (auto: conservadora; criterio d; fuente: `docs/ROADMAP.md`, H26.1, «Las dos evals de `legal-core`»; ADR 0043, punto 6; `internal/evals/plazos.go` y `informe.go`)
- Q: ¿Qué comprobación del job sale en rojo si alguna de las seis sesiones de `claude-sonnet-5-5` de la eval 11 no intenta la lectura de `a124` que su `sin_servir` hace fallar (SC-013)? → A: Un umbral sin modelo nuevo, `lectura_sin_intentar:<modelo>:<modo>`, con 0, que decide, publicado en la forma de `plazo_sin_fecha` (FR-033); cuenta las sesiones medidas del modelo que decide de las evals con un comando esperado de bloque cuya lectura está en su `sin_servir` (el conjunto de FR-016, hoy solo la 11) que no tienen en `lecturas_sin_servir` ninguna invocación de esa lectura. `sesiones_con_lectura_sin_servir` sigue publicado sin decidir. (auto: criterio a; fuente: `docs/ROADMAP.md`, H26.1, Aceptación; constitución, «Gates», «Umbrales que se cumplen», y «Criterio de decisión autónoma», punto 2; ADR 0043, punto 4)
- Q: ¿Qué comprueba el job de lo que FR-029 y FR-030 piden al texto de la respuesta de `legal-core`, si `legal-core` no tiene juez con modelo y el binario no analiza el texto de una respuesta? → A: Solo lo que tiene forma y ya está en el spec: en la 11, la línea `⚠ PLAZO:` con `2026-10-13` y la lectura de `a124` intentada; en la 07, la 08 y la 12, la línea `⚠ NO CUBIERTO:` con su norma y ninguna línea del plazo, con `cobertura_sin_declarar` y `plazos calcular` en `prohibidos`. Que la respuesta diga aparte que no pudo leer la redacción y no afirme el precepto (FR-029), y que tras `⚠ NO CUBIERTO:` no exponga los plazos, los recursos ni el procedimiento de la norma que lo rige (FR-030), quedan en el texto de los pasos 10 y 7 de la skill, que comprueban la validación de las skills de `make ci` y la revisión final (capa 2). (auto: conservadora; criterio d; fuente: `docs/ROADMAP.md`, H26.1; ADR 0043, punto 7; ADR 0037; `internal/evals/formato.go`, `ComandoProhibido`; constitución, «Gates», capa 2)

## User Scenarios & Testing *(mandatory)*

### User Story 1 - La caché de la sesión sirve lo que la skill lleva a leer (Priority: P1)

Quien lee el informe del job de evals (la persona, al leer el informe final del run) necesita que el rojo de una eval
diga algo de la skill y no de la caché: una lectura que la skill pide en sus referencias o en sus reglas —un precepto de
`data/plazos.yaml` o de `data/cobertura.yaml`, el índice o los metadatos de una norma de `data/`— se sirve en toda
sesión, de cualquier skill, igual que en el uso real.

**Why this priority**: es el defecto que puso en rojo `legal-core` en la #175, y H9.1 añade a `data/plazos.yaml`
preceptos que sin esto tendría que nombrar a mano en alguna eval.

**Independent Test**: se prepara la caché de una sesión de cualquier eval y se ejecuta desde ella, sin red, cada lectura
derivada de `data/`; todas salen con 0. Se añade a una copia de `data/plazos.yaml` un precepto sin grabar y la
preparación da una falta que lo nombra.

**Acceptance Scenarios**:

1. **Dado** `data/cobertura.yaml`, cuyas entradas nombran el bloque `daprimera` de `BOE-A-2015-10565`, y una eval de
   `legal-core` que no lo nombra (la 07), **Cuando** se prepara la caché de una sesión de esa eval, **Entonces** la
   lectura del bloque `daprimera` de `BOE-A-2015-10565` sale con 0 sin fecha, con `--fecha 9999-12-31` y con
   `--fecha 2026-09-09`.
2. **Dado** cada norma de `data/normas.yaml`, de `data/plazos.yaml` y de `data/cobertura.yaml`, **Cuando** se prepara la
   caché de una sesión de una eval de `boe-legislacion`, de `legal-core` o de `jurisprudencia`, **Entonces** su índice y
   sus metadatos salen con 0 desde ella.
3. **Dado** un precepto añadido a `data/plazos.yaml` o a `data/cobertura.yaml` cuyo bloque no está en lo grabado,
   **Cuando** se prepara la caché, **Entonces** la preparación da una falta que nombra esa lectura, y el test de `make ci`
   que prepara la caché de las evals del repositorio sale en rojo.
4. **Dado** la sesión con la que se mide el juez de una skill, **Cuando** se prepara su caché, **Entonces** sirve lo
   mismo que la de una sesión del job.

---

### User Story 2 - Una eval hace fallar a propósito la lectura que quiere probar (Priority: P1)

`legal-core` promete que la línea `⚠ NO CUBIERTO:` va aunque falle la lectura de `daprimera` (paso 7, regla 9) y, con
este hito, que la línea `⚠ PLAZO:` se mantiene aunque falle la lectura del precepto en la fecha de inicio (paso 10). Una
eval puede declarar en `sin_servir` lo que la caché de sus sesiones no sirve, aunque otra eval, `lecturas` o `data/` lo
necesiten; la lectura sale entonces como con el BOE inalcanzable, que es la avería real. Dos evals nuevas de
`legal-core` lo usan, y el informe dice cuántas sesiones de cada una intentaron la lectura que no se sirve.

**Why this priority**: sin ella, la US1 deja sin medir la robustez que arregló la #178.

**Independent Test**: se juzgan sesiones sintéticas de las evals 11 y 12 con la lectura intentada y fallida, y se
prepara la caché de una sesión de la 11 y otra de la 05.

**Acceptance Scenarios**:

1. **Dado** la eval 11, con el bloque `a124` de `BOE-A-2015-10565` en `sin_servir`, y la 05, que lo pide como comando,
   **Cuando** se prepara la caché de una sesión de cada una, **Entonces** desde la de la 11 la lectura de ese bloque sale
   con 4 o 5 sin fecha, con `--fecha 2026-09-09` y con `--fecha 9999-12-31`, y desde la de la 05 sale con 0 en las tres.
2. **Dado** una sesión de la eval 11 que resuelve Leganés, pide `plazos calcular reposicion`, intenta leer `a124` en
   la fecha 2026-09-09, la lectura sale con 5, y la respuesta lleva `⚠ PLAZO: 2026-10-13` con la cita de cada precepto
   de `data.preceptos` y dice aparte que no pudo leer esa redacción, **Cuando** se juzga, **Entonces** pasa; la lectura
   está en `lecturas_sin_servir` de la sesión y no en `fuera_de_lo_grabado` ni en `otras_fallidas`; y la sesión no cuenta
   en `plazo_sin_fecha`.
3. **Dado** una sesión de la eval 12 (cuotas de la Seguridad Social, `daprimera` en `sin_servir`) cuya respuesta lleva
   la línea `⚠ NO CUBIERTO:` con `BOE-A-2015-10565` y ninguna línea del plazo, y que no ejecuta `plazos calcular`,
   **Cuando** se juzga, **Entonces** pasa y no cuenta en `cobertura_sin_declarar`; si ejecuta `plazos calcular`, no pasa.
4. **Dado** una eval con un bloque en `sin_servir` que es también una de sus `citas`, **Cuando** se lee el conjunto,
   **Entonces** el conjunto se rechaza como mal formado, con el fichero de la eval y el bloque.
5. **Dado** el informe del job de `legal-core`, **Cuando** se escribe, **Entonces** cada serie de las evals 11 y 12 lleva
   cuántas de sus sesiones intentaron alguna lectura de su `sin_servir`, también si son 0, y las series de las demás
   evals no lo llevan.

---

### User Story 3 - Una eval sirve la exploración razonable que la provoca (Priority: P2)

En las evals por materia de `boe-legislacion`, el modelo que decide, con la skill activada, lee junto al artículo de la
pregunta los que lo rodean en la misma norma (clase (b) del ADR 0043). La eval puede llevar esas lecturas en `lecturas`,
cada una con su motivo: la caché de la skill las sirve y el juicio no las exige.

**Why this priority**: quita una de las vías por las que `afirma_lo_no_leido` marca una respuesta por un hueco de lo
grabado; es menor que la de la US1 porque esa clase no decide con 0.

**Independent Test**: se lee un conjunto con `lecturas` y se prepara la caché; se juzga una sesión que no las lee.

**Acceptance Scenarios**:

1. **Dado** la eval 16 de `boe-legislacion` con los bloques `a26` y `a27` de `BOE-A-2015-10566` en `lecturas`,
   **Cuando** se prepara la caché de una sesión de cualquier eval de `boe-legislacion`, **Entonces** los dos bloques salen
   con 0 sin fecha y con `--fecha 9999-12-31`.
2. **Dado** una sesión de la eval 16 que no lee ni `a26` ni `a27`, **Cuando** se juzga, **Entonces** su juicio es el
   mismo que sin `lecturas`: ningún motivo, comando ni cita nuevos.
3. **Dado** una entrada de `lecturas` con el motivo vacío, o con un bloque que ya es un comando o una cita de la misma
   eval, **Cuando** se lee el conjunto, **Entonces** se rechaza como mal formado, con el fichero y la entrada.

---

### User Story 4 - El informe cuenta las lecturas fuera de lo grabado (Priority: P2)

Quien lee el informe del job ve en sus umbrales publicados, por cada modelo y modo, cuántas lecturas de las sesiones
salieron con 4 o 5 sin estar en la caché —sin las que provoca `sin_servir`— sobre cuántas sesiones medidas: un hueco
nuevo se ve en el resumen sin buscarlo en el registro.

**Why this priority**: es la señal que habría enseñado el hueco de la #175 antes de que pusiera nada en rojo.

**Independent Test**: un informe de sesiones sintéticas con lecturas de 4 y 5, unas de `sin_servir` y otras no.

**Acceptance Scenarios**:

1. **Dado** tres sesiones medidas de `claude-sonnet-5-5` en el modo orden, dos lecturas con 5 fuera de la caché en una
   y una lectura de `sin_servir` con 5 en otra, **Cuando** se escribe el informe, **Entonces** lleva el umbral
   `fuera_de_lo_grabado:claude-sonnet-5-5:orden` con medida 2, total 3 y `decide: false`, y su veredicto no cambia.
2. **Dado** un modelo y un modo sin ninguna lectura fuera de lo grabado, **Cuando** se escribe el informe, **Entonces**
   el umbral de ese modelo y ese modo está, con medida 0.

---

### User Story 5 - El motivo de un 1 que no es de kitlegal (Priority: P3)

En `otras_fallidas` de `boe-legislacion` aparecen unas diez invocaciones de `boe indice` por job con código 1, el de un
defecto inesperado, y ninguna es un defecto del binario: unas son llamadas a una herramienta en sesiones del modo orden,
que no tienen servidor, cuyo resultado no trae sobre de kitlegal; las otras, con toda probabilidad, órdenes cuya salida
estándar se cerró (un `| python3 …` en un runner sin Python) y con las que kitlegal termina con 1 como promete el
contrato de H1. Quien lee el informe necesita distinguirlas de un 1 que sí es un defecto.

**Why this priority**: no cambia ningún veredicto; aclara una señal que hoy se lee como un defecto.

**Independent Test**: se juzgan sesiones sintéticas con esas tres formas de fallo.

**Acceptance Scenarios**:

1. **Dado** una sesión del modo orden con una llamada a `boe_indice` cuyo resultado es un error sin sobre de kitlegal,
   **Cuando** se juzga, **Entonces** la llamada se publica en `invocaciones` y en `otras_fallidas` con `codigo: null` y
   `motivo: "sin_sobre"`, y no con 1.
2. **Dado** una sesión con una orden de Bash `kitlegal boe indice BOE-A-2015-10565 --json | python3 -c …` cuyo resultado
   trae la línea con la que el binario declara que no pudo escribir su salida estándar, y en su traza la invocación
   `boe indice BOE-A-2015-10565` terminada con 1, **Cuando** se juzga, **Entonces** esa invocación está en
   `otras_fallidas` con `codigo: 1`, `motivo: "salida_no_escrita"` y `bash` con la orden tal cual.
3. **Dado** una invocación de la traza que terminó con 1 y cuya orden de Bash trae otra salida de error, **Cuando** se
   juzga, **Entonces** queda como hoy: `codigo: 1`, sin `motivo` ni `bash`.
4. **Dado** dos usos de Bash de la misma sesión que cumplen las condiciones del enlace para la misma invocación,
   **Cuando** se juzga, **Entonces** la invocación queda como hoy.

---

### User Story 6 - `legal-core` v1.2.3 dice qué hacer con una lectura fallida (Priority: P2)

Quien pregunta hasta cuándo puede recurrir sigue necesitando la fecha aunque el BOE no responda: `plazos` la da sin red y
ya comprobó contra el BOE grabado que su apartado regía el día de inicio (ADR 0042). Y tras `⚠ NO CUBIERTO:`, la
respuesta no expone el procedimiento de la norma que lo rige, que llega con su vertical (ADR 0027).

**Why this priority**: es lo que la eval 11 mide; la eval 12 mide lo que ya promete la v1.2.2.

**Independent Test**: `make ci` valida la skill (frontmatter, referencias, longitud), la revisión final juzga su
protocolo, y las evals 11 y 12 miden en el job lo que tiene forma (la línea `⚠ PLAZO:`, la lectura intentada, la línea
`⚠ NO CUBIERTO:`, `plazos calcular` prohibido).

**Acceptance Scenarios**:

1. **Dado** una lectura de un precepto de `data.preceptos` en la fecha de inicio que sale con 4 o 5, **Cuando** la skill
   responde, **Entonces** lleva la línea `⚠ PLAZO:` con la fecha de `plazos` y la cita de cada precepto de
   `data.preceptos`, sin la línea `⚠ REDACCIÓN A FECHA:` de ese precepto, dice aparte que no pudo leer esa redacción y no
   afirma lo que dice el precepto (el texto del paso 10 lo pide; el job mide la línea `⚠ PLAZO:` y la lectura intentada).
2. **Dado** una materia de `references/cobertura_por_materia.md`, **Cuando** la skill da la línea `⚠ NO CUBIERTO:`,
   **Entonces** no expone los plazos, los recursos ni el procedimiento de la norma que lo rige: la nombra la línea con lo
   que trae la referencia (el texto del paso 7 lo pide; el job mide la línea `⚠ NO CUBIERTO:` y `plazos calcular`
   prohibido).

---

### Edge Cases

- Un bloque que está a la vez en `sin_servir` de una eval y en `lecturas`, en otra eval o en `data/`: no se sirve en las
  sesiones de la eval de `sin_servir`, y sí en las demás (FR-013).
- Un bloque de `sin_servir` que es un comando esperado de la misma eval: la lectura intentada en la fecha del comando
  que sale con 4 o 5 lo satisface, y la intentada solo en otra fecha o sin ella, no (FR-016).
- Una invocación que consulta una lectura de `sin_servir` y es una llamada sin sobre: va a `otras_fallidas` con su
  motivo y no es una lectura intentada (FR-015, FR-023, FR-033); si es una orden de salida no escrita, va a
  `lecturas_sin_servir` con su motivo y es una lectura intentada (FR-015, FR-024, FR-033).
- Una serie de una eval con `sin_servir` cuyas sesiones no intentan ninguna lectura de él: el informe lo enseña con 0
  (FR-015).
- Una invocación con código 1 que no se puede enlazar sin ambigüedad con un uso de Bash —ninguno o más de uno cumple—:
  queda como hoy (FR-027).
- Una sesión de una eval sin binario ni servidor: no cuenta en `fuera_de_lo_grabado:<modelo>:<modo>` (FR-021).

## Requirements *(mandatory)*

### Functional Requirements

**Lo que lleva a leer la skill (ADR 0043, punto 1)**

- **FR-001**: La caché preparada de toda sesión de evals, de cualquier skill, sirve, además de las consultas necesarias
  de las evals de su skill: cada bloque de `data/plazos.yaml` —los de los tipos, de las unidades y de los cómputos— y de
  `data/cobertura.yaml` —el precepto de cada entrada y el aplicable de su norma—, sin fecha y con `--fecha 9999-12-31`,
  con la que se sirve la lectura del bloque en cualquier fecha; y el índice y los metadatos de cada norma de esos dos
  ficheros y de `data/normas.yaml`. Criterio: desde la caché preparada de una sesión de cada skill, cada una de esas
  lecturas sale con 0 sin red.
- **FR-002**: Lo de FR-001 se deriva de los ficheros de `data/` leídos con los lectores que ya tienen (`plazos.LeerTipos`,
  `skills.LeerCobertura` y el de las normas), sin copiar ningún identificador de norma ni de bloque en el arnés de evals.
  Criterio: un precepto añadido a `data/plazos.yaml` se sirve sin tocar ninguna eval ni el arnés; y el mutante que retira
  la derivación hace fallar algún test de `make ci`.
- **FR-003**: Si una lectura de FR-001 no está en lo grabado, la preparación de la caché da una falta que nombra esa
  lectura, como hoy con las consultas necesarias, y la sesión no se abre. Criterio: con un precepto de `data/` sin
  grabar, el test de `make ci` que prepara la caché de las evals del repositorio sale en rojo.
- **FR-004**: La sesión con la que se mide el juez de una skill (`internal/evals/medida.go`) sirve lo mismo que FR-001.
  Criterio: un test prepara su caché y lee desde ella una lectura de FR-001 que no nombra ninguna eval, con 0.
- **FR-005**: FR-001 no graba nada: todo lo que deriva hoy de `data/` está en los tres conjuntos de lo grabado que se
  unen para las sesiones (comprobado el 2026-10-20 según el hito). Criterio: el test de FR-003 pasa con el `data/` de
  `main` sin ninguna grabación nueva.

**`lecturas` (punto 2)**

- **FR-006**: Una eval puede llevar `lecturas`, una lista opcional de `{norma, bloque, motivo}`. La caché de las
  sesiones de su skill sirve cada bloque sin fecha y con `--fecha 9999-12-31`. Criterio: desde la caché de una sesión de
  otra eval de la misma skill, la lectura sale con 0.
- **FR-007**: El juicio no comprueba ni exige que se lea un bloque de `lecturas`: no es un comando, ni una cita, ni da
  motivo, ni cuenta en ningún umbral. Criterio: una sesión que no lo lee tiene el mismo juicio que con la eval sin
  `lecturas`.
- **FR-008**: La lectura del conjunto de evals (`LeerConjunto`) rechaza como mal formada la eval con una entrada de
  `lecturas` cuyo motivo está vacío o es solo de blancos, o cuyo bloque —la misma norma y el mismo bloque— ya es un
  comando esperado de bloque, con o sin fecha, o una cita de la misma eval; el error nombra el fichero y la entrada.
  Criterio: un test por cada uno de los tres rechazos.
- **FR-009**: El esquema del formato de eval admite `lecturas` y `sin_servir` con su forma y rechaza otra; `schema-check`
  valida todas las evals del repositorio contra él. Criterio: `make ci` en verde con las evals del hito y en rojo con una
  `lecturas` sin `motivo`.
- **FR-010**: Una tarea `[datos]` añade al manifiesto de `internal/evals/testdata/grabaciones.json` —el único de ese
  paquete en el que escribe `grabar_datos`— los bloques `a26` y `a27` de la Ley 40/2015 (`BOE-A-2015-10566`), `a18` de
  la Ley 35/2006 (`BOE-A-2006-20764`) y `a15` del texto refundido de la Ley reguladora de las Haciendas Locales
  (`BOE-A-2004-4214`), cuyos ids están en los índices grabados; los graba `grabar_datos`, sin modelo, de la fuente con
  fila revisada en `docs/SOURCES.md` de `main`. Ninguna otra lectura de la clasificación del ADR 0043 se graba.
  Criterio: tras la grabación, la preparación de la caché de `boe-legislacion` no da ninguna falta.
- **FR-011**: Las evals de `boe-legislacion` llevan en `lecturas`, cada bloque con su motivo: la 16
  (`16-lrjsp-legalidad-por-materia.yaml`), `a26` y `a27` de `BOE-A-2015-10566`; la 15
  (`15-irpf-rendimientos-por-materia.yaml`), `a18` de `BOE-A-2006-20764`; y la 14
  (`14-trlrhl-impuestos-por-materia.yaml`), `a15` de `BOE-A-2004-4214`. Nada más cambia en ellas. Criterio: el test de
  las evals del repositorio lee esas tres con sus `lecturas` y la caché de `boe-legislacion` sirve los cuatro bloques.

**`sin_servir` (punto 4)**

- **FR-012**: Una eval puede llevar `sin_servir`, una lista opcional de bloques `{norma, bloque}` y de consultas de norma
  `{norma, verbo}` con `verbo` `indice` o `metadatos`. Criterio: el esquema y la lectura del conjunto admiten las dos
  formas y rechazan otro verbo.
- **FR-013**: La caché de las sesiones de esa eval no sirve lo que lleva su `sin_servir`, con ninguna fecha, aunque lo
  necesiten otra eval, `lecturas` o `data/`; en las sesiones de las demás evals se sirve como sin él. Criterio: desde la
  caché de una sesión de la eval, la lectura del bloque sale con 4 o 5 sin fecha y con `--fecha 9999-12-31`, y desde la
  de otra eval que lo necesita, con 0; y el mutante que aplica `sin_servir` a todas las evals de la skill hace fallar
  algún test.
- **FR-014**: La lectura del conjunto rechaza como mal formada la eval con un bloque de `sin_servir` que es una de sus
  `citas`; el error nombra el fichero y el bloque. Criterio: un test.
- **FR-015**: Cada invocación de una sesión —orden de la traza o llamada a una herramienta— que consulta una lectura de
  `sin_servir` de su eval (`articulo` de esa norma con ese bloque, con o sin `--fecha`, o `articulos` con ese bloque
  entre los que pide; o el verbo `indice` o `metadatos` de esa norma), termine como termine, va a `lecturas_sin_servir`
  de la sesión con su orden y su código, y no a `fuera_de_lo_grabado` ni a `otras_fallidas`, y es una lectura intentada;
  también la orden de salida no escrita (FR-024), que sí llegó a la lectura y lleva ahí su `codigo: 1`, su `motivo` y su
  `bash`. Solo la llamada sin sobre (FR-023), en la que el binario no corrió, va a `otras_fallidas` como dice ese
  requisito y no es una lectura intentada, ni en `sesiones_con_lectura_sin_servir` ni en FR-033. Cada serie de una eval
  con `sin_servir` lleva en `sesiones_con_lectura_sin_servir` cuántas de sus sesiones leídas tienen alguna en
  `lecturas_sin_servir`, también 0; en las series de las demás evals es `null`. Ni la lista ni el recuento deciden por
  sí mismos. Criterio: un test con sesiones sintéticas que intentan y que no intentan la
  lectura, con una llamada sin sobre que la consulta, que queda en `otras_fallidas` y no cuenta como intentada, y con una
  orden de salida no escrita que la consulta, que queda en `lecturas_sin_servir` con su motivo y cuenta como intentada.
- **FR-016**: Un comando esperado de bloque de una eval cuya lectura está en su `sin_servir` lo satisface, además de lo
  que hoy lo satisface, la invocación que lee ese bloque como hoy lo pide el comando —en su `fecha`, si la lleva— y
  termina con 4 o 5 (ADR 0043, punto 6: la lectura fallida es lo que la eval provoca): solo cambia el código admitido.
  Ningún otro comando cambia. Criterio: un test con la sesión que la intenta en la fecha del comando con 5, que satisface
  el comando; otra que la intenta solo sin `--fecha` o con otra fecha, y otra que no la intenta, que llevan el motivo del
  comando ausente.

**Las dos evals de `legal-core`**

- **FR-017**: `evals/legal-core/11-plazo-reposicion-sin-leer-el-precepto.yaml` lleva la pregunta de la 05, con
  `activa: true`; los comandos de la 05 —`territorio resolver Leganés`, `plazos calcular reposicion` y el bloque `a124`
  de `BOE-A-2015-10565` con `fecha: "2026-09-09"`—; ese bloque en `sin_servir`; el plazo de la 05 —`fecha: "2026-10-13"`
  y `avisos: []`—; y ni `citas` ni `redaccion_fechada`. La cita de los preceptos de `data.preceptos` en su línea
  `⚠ PLAZO:` no es un motivo ni cuenta en ningún umbral, y `citas_sin_consulta` sigue siendo solo de las evals sin
  binario ni servidor, así que en la 11 queda vacía (ADR 0043, punto 6). «Sin `⚠ REDACCIÓN A FECHA:`» solo dice que la
  eval no lleva `redaccion_fechada`, a diferencia de la 05: la presencia de esa línea en una respuesta de la 11 no cambia
  su juicio, y el formato de eval y el juicio no ganan campo, regla ni motivo por ella. Criterio: el test de las evals
  del repositorio la lee, y sesiones sintéticas con y sin la línea `⚠ PLAZO: 2026-10-13` pasan y no pasan.
- **FR-018**: `evals/legal-core/12-cobertura-seguridad-social-sin-leer-la-da-primera.yaml` pregunta, con el municipio y
  la fecha de la notificación, como la 07, hasta cuándo se puede recurrir una reclamación de cuotas de la Seguridad
  Social, el ejemplo del paso 7; lleva `activa: true`, el bloque `daprimera` de `BOE-A-2015-10565` en `sin_servir`,
  `prohibidos` con `plazos calcular` y `no_cubierto` con `norma: BOE-A-2015-10565`, la entrada de
  `data/cobertura.yaml` sin norma; y ni `citas` ni `comandos`. Criterio: el test de las evals del repositorio la lee, y
  una sesión sintética con la línea `⚠ NO CUBIERTO:` de `BOE-A-2015-10565` pasa y otra que ejecuta `plazos calcular` no.
- **FR-019**: Las evals 11 y 12 corren en los dos modos, con las repeticiones y los modelos de hoy, y deciden con los
  umbrales sin modelo que ya tiene la skill: la 11 cuenta en `plazo_sin_fecha:<modelo>:<modo>` y la 12 en
  `cobertura_sin_declarar:<modelo>:<modo>`, los dos con 0, y las dos en `sin_activar:<modelo>:<modo>`; sus series
  deciden el veredicto como las de las demás evals que no son informativas. Ningún umbral nuevo decide por ellas, salvo
  `lectura_sin_intentar:<modelo>:<modo>`, el control de SC-013 (FR-033). Criterio: el plan del job las incluye en los dos
  modos, y la sección «Controles de umbral» del plan tiene una fila por cada umbral en el que cuentan.
- **FR-020**: El tope del trabajo de evals de `legal-core`, hoy de 352 minutos, cubre su peor caso recalculado con las
  dos evals nuevas; el plan lo recalcula antes de cerrarse. Criterio: `TestDefinicionDelJob`, en `make ci`, sale en rojo
  si el peor caso no cabe en el tope del job.

**`lectura_sin_intentar:<modelo>:<modo>` (punto 4)**

- **FR-033**: El informe del job de una skill con al menos una eval con un comando esperado de bloque cuya lectura está
  en su `sin_servir` —el conjunto de FR-016; hoy solo `legal-core`, por la 11, con `a124`— publica, por cada modelo y
  cada modo del plan, el umbral sin modelo `lectura_sin_intentar:<modelo>:<modo>`, en la forma de `plazo_sin_fecha`:
  medida, las sesiones medidas de ese modelo en ese modo que no tienen en `lecturas_sin_servir` (FR-015) ninguna
  invocación que consulte esa lectura, con o sin `--fecha` y termine como termine, también con salida no escrita
  (FR-024); la llamada sin sobre no está en esa lista; total, las sesiones medidas de ese modelo y modo de esas evals; la
  comparación `<=` con 0; y `decide: true` solo para el modelo que decide de las evals (ADR 0031), de modo que el que no
  se cumple pone el veredicto del job en fallo. Lo lleva también con medida 0. El informe de una skill sin una eval así
  no lo lleva, como hoy el de `boe-legislacion` no lleva `plazo_sin_fecha`. La 12, que no tiene comandos, no entra: el paso
  7 de `legal-core` no obliga a leer `daprimera`. `sesiones_con_lectura_sin_servir` (FR-015) sigue publicado sin decidir,
  y el motivo del comando ausente (FR-016) con la regla de 2 de 3 de su serie sigue como está, sin ser el control de
  SC-013. Criterio: un test con sesiones sintéticas que intentan la lectura y que no, otro que no lo encuentra en el
  informe de `boe-legislacion` ni en el de `jurisprudencia` con el plan del job, un mutante que retire el umbral o no
  lo deje decidir hace fallar algún test, y la sección «Controles de umbral» del plan tiene su fila.

**`fuera_de_lo_grabado:<modelo>:<modo>` (punto 5)**

- **FR-021**: El informe del job de cada skill publica, por cada modelo y cada modo del plan, el umbral
  `fuera_de_lo_grabado:<modelo>:<modo>`: medida, las lecturas de las sesiones medidas de ese modelo en ese modo que
  salieron con 4 o 5 —las de `fuera_de_lo_grabado` de cada sesión, sin las de `sin_servir` (FR-015)—; total, esas
  sesiones medidas; la comparación `<=` con 0, en la forma de los demás umbrales publicados; y `decide: false`, de modo
  que no cambia el veredicto: es una señal para quien lee el informe, no un umbral que el hito pida cumplir. Lo lleva
  también con medida 0. Las sesiones de las evals sin binario ni servidor, que no tienen modo del plan, no cuentan.
  Criterio: un test del recuento, sin contar `sin_servir` y con `decide: false`.
- **FR-022**: El `fuera_de_lo_grabado` de la raíz del informe, que junta el de cada sesión, tampoco lleva las lecturas de
  `sin_servir`. Criterio: el test de FR-021.

**El motivo de un 1 que no es de kitlegal**

- **FR-023**: Una llamada a una herramienta cuyo resultado no trae sobre de kitlegal (un texto que no es un objeto JSON
  con `ok`, como el aviso del agente de que la herramienta no existe) se publica en `invocaciones` y en
  `otras_fallidas` con `codigo: null` y `motivo: "sin_sobre"`, y no con 1, también si consulta una lectura de
  `sin_servir` (FR-015). Una llamada sin resultado sigue como hoy. Criterio: un test con una sesión sintética.
- **FR-024**: Una invocación de la traza que terminó con 1 porque el binario no pudo escribir su salida estándar, y que
  se enlaza con el uso de Bash del que salió (FR-026), lleva en `otras_fallidas` `codigo: 1`,
  `motivo: "salida_no_escrita"` y `bash` con la orden de ese uso tal cual; si consulta una lectura de `sin_servir`, los
  lleva en `lecturas_sin_servir` y no en `otras_fallidas` (FR-015). Criterio: un test con una orden con `| python3` cuya
  salida trae la línea de error del binario.
- **FR-025**: La línea de error se reconoce con la misma pieza del binario que la escribe —su error, o una constante que
  el binario y el arnés compartan, con un test que los ate—, nunca con una copia literal del mensaje ni con texto que
  escribe el modelo; y se lee del resultado de cada uso del transcript, que ya se lee para el juez. Criterio: el mutante
  que cambia el mensaje de la escritura fallida del binario hace fallar algún test.
- **FR-026**: Una invocación de la traza que terminó con 1 se enlaza con un uso de Bash de la misma sesión si ese uso es
  una orden que nombra `kitlegal` y contiene como palabras —con la regla con la que el juicio ya decide si una orden
  nombra `kitlegal`— el applet, el verbo y cada argumento de la invocación, y su
  resultado trae la línea de FR-025; y si es el único uso de la sesión que lo cumple para esa invocación. Criterio: un
  test con un uso que cumple y otro con dos.
- **FR-027**: Una invocación con código 1 que no se enlaza sin ambigüedad —ningún uso cumple FR-026, o más de uno—, o
  cuyo uso trae otra salida de error, queda como hoy: `codigo: 1`, sin `motivo` ni `bash`. Criterio: un test de cada
  caso.
- **FR-028**: Nada de FR-023 a FR-027 decide: ningún juicio, serie, umbral ni veredicto cambia por ellos; tampoco cambia
  el código con que termina el binario. Criterio: los tests de FR-023 y FR-024 comparan el juicio de la sesión con el de
  antes, salvo los campos publicados.

**`legal-core` v1.2.3**

- **FR-029**: El paso 10 de `skills/legal-core/SKILL.md` dice qué hacer si falla la lectura de un precepto en la fecha
  de inicio (ADR 0043, punto 6): la línea `⚠ PLAZO:` se mantiene con la fecha de `plazos` y la cita de cada precepto de
  `data.preceptos`; en lugar de su línea `⚠ REDACCIÓN A FECHA:`, la respuesta dice aparte que no pudo leer esa
  redacción, como `boe-legislacion` dice cualquier lectura fallida, y no afirma lo que dice el precepto. Criterio: el
  texto del paso 10, que comprueban la validación de las skills de `make ci` y la revisión final (capa 2); y la eval 11
  en el job, que solo mide lo que tiene forma: la línea `⚠ PLAZO:` con `2026-10-13` (`plazo_sin_fecha`) y la lectura de
  `a124` intentada (FR-016, FR-033). Que la respuesta diga aparte que no pudo leer la redacción y no afirme el precepto
  no lo comprueba ningún control del job.
- **FR-030**: El paso 7 dice que, tras la línea `⚠ NO CUBIERTO:`, la respuesta no expone la norma que rige el
  procedimiento —sus plazos, sus recursos—: la nombra la línea con lo que trae `references/cobertura_por_materia.md`
  (ADR 0043, punto 7). Criterio: el texto del paso 7, que comprueban la validación de las skills de `make ci` y la
  revisión final (capa 2); y las evals 07, 08 y 12 en el job, que solo miden lo que tiene forma: la línea
  `⚠ NO CUBIERTO:` con su norma y ninguna línea del plazo (`cobertura_sin_declarar`) y `plazos calcular` en `prohibidos`.
  Que no exponga los plazos, los recursos ni el procedimiento de la norma que lo rige no lo comprueba ningún control del
  job.
- **FR-031**: `skills/legal-core/SKILL.md` sigue por debajo de 300 líneas y la versión es la 1.2.3; la 1.3 es la de
  H9.1. Criterio: la comprobación de las skills de `make ci`, que sale en rojo con 300 líneas o más.

**Documentación**

- **FR-032**: `CHANGELOG.md` (*Unreleased*) recoge la entrega; `CONTRIBUTING.md` documenta `lecturas` y `sin_servir`
  en la tabla del formato de las evals, con qué sirve la caché de una sesión; y lo que `CLAUDE.md` y `docs/WORKFLOW.md`
  digan de la caché de las sesiones queda alineado con FR-001, FR-006 y FR-013. Criterio: la revisión final.

### Key Entities

- **Lectura derivada de `data/`**: un bloque de `data/plazos.yaml` o de `data/cobertura.yaml`, o el índice o los
  metadatos de una norma de esos dos ficheros o de `data/normas.yaml`, que la caché de toda sesión sirve.
- **`lecturas` de una eval**: `{norma, bloque, motivo}`; la exploración razonable que provoca esa eval (clase (b)).
- **`sin_servir` de una eval**: `{norma, bloque}` o `{norma, verbo}`; lo que la caché de sus sesiones no sirve.
- **`lecturas_sin_servir` de una sesión**: las invocaciones que consultan una lectura de `sin_servir`, con orden y
  código (la orden de salida no escrita, también con su `motivo` y su `bash`), salvo la llamada sin sobre.
- **`sesiones_con_lectura_sin_servir` de una serie**: un entero en las series de las evals con `sin_servir`; `null` en
  las demás.
- **`fuera_de_lo_grabado:<modelo>:<modo>`**: un umbral publicado del informe, con medida, total y `decide: false`.
- **`motivo` y `bash` de una invocación fallida**: `sin_sobre` (con `codigo: null`) o `salida_no_escrita` (con
  `codigo: 1` y la orden de Bash).

## Uso, de fuera adentro

- **La caché derivada de `data/` (FR-001)** la pide la skill en la sesión: `legal-core` lee `daprimera` una vez por
  pregunta en el paso 7 y cada precepto de `data.preceptos` (de uno a tres) en el paso 10; `boe-legislacion` lee el
  índice de la norma de la pregunta. La prepara el arnés una vez por sesión. Tamaño: hoy 17 bloques × 2 lecturas y el
  índice y los metadatos de unas veinte normas, del orden de 80 consultas, que crecen con `data/` y no con el uso: los
  meses de uso de kitlegal, sus cientos de normas y miles de bloques consultados, no están en ninguna sesión de evals,
  que arranca de cero. Con los preceptos que añada H9.1, unas decenas más. Deja de darse cuando el precepto sale de
  `data/`.
- **`lecturas` (FR-006)**: cuatro bloques en tres evals de hoy, servidos en todas las sesiones de `boe-legislacion`.
  Crece solo cuando un hito añade una lectura de la clase (b) con su motivo y su grabación; deja de darse cuando se
  retira de la eval.
- **`lecturas_sin_servir` y `sesiones_con_lectura_sin_servir` (FR-015)** los lee la persona en el informe del job, una
  vez por job: una lista por sesión, acotada por las invocaciones de esa sesión (de una a tres lecturas en la 11 y la 12,
  unos 150 bytes cada una), y un entero por serie de las dos evals con `sin_servir`: 2 evals × 2 modelos × 2 modos = 8
  enteros. Cada job los recalcula desde cero: ninguna señal de un job pasa al siguiente.
- **`lectura_sin_intentar:<modelo>:<modo>` (FR-033)**: lo lee la persona en los umbrales del informe del job de
  `legal-core`, la única skill con una eval con un comando de bloque en su `sin_servir`, una vez por job, y decide su
  veredicto: 2 modelos × 2 modos = 4 umbrales en ese job y ninguno en el de las demás skills, de unos 300 bytes cada
  uno; si el del modelo que decide no se cumple, el motivo del veredicto, en la forma del de `plazo_sin_fecha`, nombra
  las sesiones que cuentan, acotadas por las de las series de las evals con un comando de bloque en su `sin_servir`
  (hoy, como mucho las tres de la 11 en ese modo), sin depender de lo acumulado. Cada job lo recalcula desde cero y
  ninguna señal pasa al siguiente. Deja de darse en el job en que las sesiones intentan la lectura, o cuando la skill
  deja de tener una eval que lleve en `sin_servir` la lectura de un comando, y entonces sale del informe.
- **`fuera_de_lo_grabado:<modelo>:<modo>` (FR-021)**: lo lee la persona en los umbrales del informe, una vez por job: 2
  modelos × 2 modos = 4 umbrales por skill, unos 300 bytes cada uno, sin depender del número de sesiones. La medida de un
  hueco deja de darse en el job siguiente al que lo cierra —la lectura entra en `data/` o en `lecturas` y se graba, o la
  skill deja de pedirla— y no se arrastra.
- **`motivo` y `bash` (FR-023, FR-024)**: en las invocaciones fallidas de cada sesión, que la persona lee en el informe;
  con lo observado en la #175 y la #178, unas diez por job en `boe-legislacion`, de unos 200 bytes cada una. Dejan de
  darse cuando el modelo deja de llamar a una herramienta que su sesión no tiene o de filtrar la salida con un programa
  que el runner no tiene.

## Success Criteria *(mandatory)*

### Measurable Outcomes

- **SC-001**: `make ci` en verde, con `schema-check` y las reglas del conjunto. Lo hace cumplir el propio `make ci`, que
  el workflow mide antes de la revisión y en el cierre.
- **SC-002**: La sección «Controles de umbral» del plan tiene una fila por cada umbral de las evals 11 y 12 —también
  `lectura_sin_intentar`, FR-033—, y las de
  los que ya decidían. La comprueba `scripts/workflow/precheck.sh plan`, que da un defecto mecánico por cada fila sin
  requisito definido o sin control.
- **SC-003**: Un test de `make ci` comprueba que la caché preparada sirve cada bloque de `data/plazos.yaml` y de
  `data/cobertura.yaml`, con y sin fecha, y el índice y los metadatos de cada norma de los tres ficheros, y que un
  precepto añadido a `data/` sin grabar da una falta (FR-001 a FR-005).
- **SC-004**: Tests de `lecturas`: se sirve sin exigirse; motivo vacío o bloque que ya es comando o cita, rechazados
  (FR-006 a FR-008).
- **SC-005**: Tests de `sin_servir`: el bloque no se sirve en las sesiones de esa eval, ni sin fecha ni con
  `9999-12-31`, y sí en las de otra eval que lo necesite; una cita en `sin_servir`, rechazada; el recuento de sesiones que
  lo intentaron (FR-013 a FR-015).
- **SC-006**: Un test del recuento `fuera_de_lo_grabado:<modelo>:<modo>`, sin contar `sin_servir` y con `decide: false`
  (FR-021, FR-022).
- **SC-007**: Tests del motivo con sesiones sintéticas: una llamada sin sobre de kitlegal sale como «sin sobre» y no con
  1; una orden con `| python3` que termina con 1 y la línea de error del binario lleva la orden y el motivo; un 1 con
  otra salida de error queda como hoy, sin motivo; y un mutante que cambie el mensaje de la escritura fallida del binario
  hace fallar algún test (FR-023 a FR-027).
- **SC-008**: Un mutante por pieza: sin la derivación de `data/`, o con `sin_servir` aplicado a todas las evals de la
  skill, algún test falla (FR-002, FR-013).
- **SC-009**: El tope del trabajo de evals de `legal-core`, recalculado con las dos evals nuevas antes de cerrar el plan,
  cubre su peor caso; `TestDefinicionDelJob` sale en rojo si no lo cubre (FR-020).
- **SC-010**: `CHANGELOG.md` (*Unreleased*); y el job de evals en la propuesta de cambio, que el workflow mide en el
  cierre sobre la cabeza de la rama (FR-032).
- **SC-011**: En el informe del job de cierre de `legal-core`, ninguna sesión de `claude-sonnet-5-5` tiene en
  `fuera_de_lo_grabado` un bloque de `data/` (umbral: 0 sesiones). Lo hace cumplir el test de SC-003, que sale en rojo
  en `make ci` si la caché no sirve un bloque de `data/` sin fecha y con `9999-12-31`, con la que se sirve en cualquier
  fecha; y la preparación de la sesión, que da la falta y no la abre si no está grabado. Las lecturas de `sin_servir` van
  aparte (FR-015).
- **SC-012**: Las evals 11 y 12 pasan en los dos modos: cada serie de `claude-sonnet-5-5` de cada una llega al umbral de
  su serie (ADR 0016), y si no llega el veredicto del job de `legal-core` es fallo (FR-019).
- **SC-013**: La eval 11 da las seis sesiones de `claude-sonnet-5-5` (tres por modo) con la lectura de `a124`
  intentada: `sesiones_con_lectura_sin_servir` es 3 en cada una de sus dos series. Lo hace cumplir el umbral sin modelo
  `lectura_sin_intentar:claude-sonnet-5-5:<modo>` con 0 (FR-033), que decide: una sesión de la 11 que no intenta la
  lectura de `a124` pone el veredicto del job de `legal-core` en fallo, sin depender de la regla de 2 de 3 de la serie.
- **SC-014**: `cobertura_sin_declarar`, `plazo_sin_fecha`, `vigencia_sin_declarar`, `redaccion_sin_fechar`,
  `lectura_sin_intentar` y `sin_activar` de `legal-core` se cumplen en el informe del job de cierre, con 0 en cada modo;
  los seis deciden, y el que no se cumple pone el veredicto del job en fallo.
- **SC-015**: `fuera_de_lo_grabado:<modelo>:<modo>` se publica con `decide: false` en el informe del job de cierre de
  cada skill, por cada modelo y modo del plan (FR-021).
- **SC-016**: `red` está vacío en el informe del job de cierre de `legal-core`; una petición llegada a la red es un
  motivo del informe y pone su veredicto en fallo.
- **SC-017**: En el informe del job de cierre de `boe-legislacion`, ninguna invocación de `otras_fallidas` con código 1
  es una llamada sin sobre (umbral: 0), y las órdenes que terminaron con 1 por no poder escribir su salida, enlazadas
  con su uso de Bash según FR-026, llevan su motivo. Lo hacen cumplir los tests de SC-007 en `make ci`: una llamada sin
  sobre no puede tener código 1 (FR-023). El informe de este hito confirma o descarta la causa del ADR 0043 con lo que
  ese informe del job enseñe de esas invocaciones.
- **SC-018**: En el informe del job de cierre de `boe-legislacion`, ninguna sesión tiene en `fuera_de_lo_grabado` un
  bloque de las `lecturas` de sus evals (umbral: 0). Lo hacen cumplir los tests de SC-004 —la caché de la skill sirve
  cada bloque de `lecturas` sin fecha y con `9999-12-31`— y la falta de la preparación si no está grabado (FR-003,
  FR-006).
- **SC-019**: `afirma_lo_no_leido:<modelo>` de `boe-legislacion` se cumple, con su tasa de hoy (ADR 0038): el umbral
  decide, y si no se cumple el veredicto del job es fallo.
- **SC-020**: `jurisprudencia` sigue aprobado: el veredicto de su job de cierre es aprobado.
- **SC-021**: `skills/legal-core/SKILL.md` tiene menos de 300 líneas; la comprobación de las skills de `make ci` sale en
  rojo con 300 o más (FR-031).

### Criterios del hito, tal como los escribe

**Controles** (SC-001 a SC-010):

- `make ci` en verde, con `schema-check` y las reglas del conjunto;
- la sección «Controles de umbral» del plan con una fila por cada umbral de las evals 11 y 12, y las de los que ya
  decidían;
- un test de que la caché preparada sirve cada bloque de `data/plazos.yaml` y de `data/cobertura.yaml`, con y sin fecha,
  y el índice y los metadatos de cada norma de los tres ficheros, y de que un precepto añadido a `data/` sin grabar da
  una falta;
- tests de `lecturas`: se sirve sin exigirse; motivo vacío o bloque que ya es comando o cita, rechazados;
- tests de `sin_servir`: el bloque no se sirve en las sesiones de esa eval, ni sin fecha ni con `9999-12-31`, y sí en
  las de otra eval que lo necesite; una cita en `sin_servir`, rechazada; el recuento de sesiones que lo intentaron;
- un test del recuento `fuera_de_lo_grabado:<modelo>:<modo>`, sin contar `sin_servir` y con `decide: false`;
- tests del motivo con sesiones sintéticas: una llamada sin sobre de kitlegal sale como «sin sobre» y no con 1; una
  orden con `| python3` que termina con 1 y la línea de error del binario lleva la orden y el motivo; un 1 con otra
  salida de error queda como hoy, sin motivo; y un mutante que cambie el mensaje de la escritura fallida del binario hace
  fallar algún test;
- un mutante por pieza: sin la derivación de `data/`, o con `sin_servir` aplicado a todas las evals de la skill, algún
  test falla;
- el tope del trabajo de evals de `legal-core`, recalculado con las dos evals nuevas antes de cerrar el plan;
- `CHANGELOG.md` (*Unreleased*); y el job de evals en la propuesta de cambio.

**Aceptación** (SC-011 a SC-020): en el informe del job de cierre de `legal-core`, ninguna sesión de
`claude-sonnet-5-5` tiene en `fuera_de_lo_grabado` un bloque de `data/`; las evals 11 y 12 pasan en los dos modos, y la
11 da las seis sesiones con la lectura de `a124` intentada; `cobertura_sin_declarar`, `plazo_sin_fecha`,
`vigencia_sin_declarar`, `redaccion_sin_fechar` y `sin_activar` se cumplen; `fuera_de_lo_grabado:<modelo>:<modo>` se
publica con `decide: false`; y `red` está vacío. En el de `boe-legislacion`, ninguna invocación de `otras_fallidas` con
código 1 es una llamada sin sobre, y las órdenes que terminaron con 1 por no poder escribir su salida llevan su motivo;
ninguna sesión tiene en `fuera_de_lo_grabado` un bloque de las `lecturas` de sus evals, y `afirma_lo_no_leido:<modelo>`
se cumple. `jurisprudencia` sigue aprobado.

## Assumptions

- Una sesión de evals no tiene red: una lectura que su caché no sirve sale con 5, porque no alcanza el `robots.txt` del
  BOE, o con 4; FR-013 y FR-016 cuentan los dos (ADR 0043).
- La lectura de un bloque con `--fecha 9999-12-31` deja en la caché sus versiones, de las que se sirve después la lectura
  en cualquier fecha (H20); por eso FR-001, FR-006 y FR-013 hablan solo de sin fecha y de `9999-12-31`.
- Todo lo que FR-001 deriva del `data/` de hoy está grabado (ADR 0043, comprobado el 2026-10-20: 17 bloques y el índice y
  los metadatos de cada norma); si no lo estuviera, FR-003 lo hace visible en `make ci`.
- El BOE ya tiene fila revisada en `docs/SOURCES.md` de `main`: la grabación de FR-010 no pide ninguna revisión humana
  nueva.
- Las repeticiones, los modelos, los modos y la concurrencia del job no cambian; el peor caso de `legal-core` sube con
  las 24 sesiones de las dos evals nuevas (2 evals × 3 repeticiones × 2 modelos × 2 modos).
- El transcript de cada sesión trae el resultado de cada uso de Bash que nombra `kitlegal`, con su salida de error, como
  ya lo lee el juez (H24); la traza da el código y los argumentos de cada invocación.

## Fuera de alcance

- Grabar lo que la clasificación del ADR 0043 pone en las clases (c) o la (b) que no se graba: búsquedas de texto libre,
  lecturas de sesiones sin la skill activada, los arts. 223 y 235 de la Ley General Tributaria.
- Servir lo grabado durante la sesión en lugar de prepararlo, o cualquier modo de reproducción en el binario publicado
  (ADR 0043, opción 2).
- Que `claude-haiku-4-5` active `legal-core` en el job.
- Cambiar el código con que termina una orden que no puede escribir su salida (contrato de H1; ADR 0023), o grabar las
  escrituras en la traza de `strace`.
- `boe-legislacion` y `jurisprudencia`, salvo las `lecturas` de sus evals 14, 15 y 16.
- Cambiar el modelo que decide, las repeticiones, los umbrales del juez o la regla de 2 de 3 por eval (ADR 0016, 0031 y
  0038).
- Guardar los transcripts de las sesiones en el job o cambiar el workflow `hito` para confirmar la causa del código 1:
  se confirma con lo que publica el informe del job (SC-017).
- Que `lecturas` o `sin_servir` sirvan o retiren búsquedas (`boe buscar`) o lecturas de otro applet que `boe`.
- Un umbral que decida sobre `fuera_de_lo_grabado` o sobre el motivo de una invocación fallida: el hito los publica sin
  decidir.
- Que el juicio de una eval con `sin_servir` dé un motivo por una línea `⚠ REDACCIÓN A FECHA:` de un bloque que no se
  sirvió. El informe final lo enseña como límite conocido: una respuesta de la 11 que ponga esa línea de memoria pasa el
  juicio del job.
- Una clase del juez con modelo para `legal-core`, y una prohibición en las evals 07, 08 y 12 de leer la norma que rige el
  procedimiento de su materia (`boe articulo`): `prohibidos` es un applet y un verbo con cualquier argumento, y el ADR 0043,
  punto 7, prohíbe exponer esa norma, no leerla. El informe final lo enseña como límite conocido: la exposición de la
  norma del procedimiento tras `⚠ NO CUBIERTO:` (los arts. 223 y 235 de la LGT en la 07) sigue sin medirse en el job.
- Un umbral que cuente en `lectura_sin_intentar` las evals con `sin_servir` que no tienen un comando esperado de bloque
  con esa lectura, como la 12.
