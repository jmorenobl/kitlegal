# Feature Specification: H41 · Las skills del producto mandan leer sus referencias

**Feature Branch**: `033-h41-las-skills-del`

**Created**: 2026-11-05

**Status**: Draft

**Input**: el fichero del hito, `docs/roadmap/in-progress/H41-las-skills-mandan-leer-sus-referencias.md` (entrada de la
persona, ADR 0018).

## Contexto

El ADR 0051 decide que una remisión a una referencia es una orden de leerla: el verbo «lee», el enlace, cuándo y lo que
trae (punto 1); la condición que pone el caso se queda y la que pone el agente se quita (punto 2); y la orden va antes
de la primera llamada o en el paso que usa la referencia, no al final (punto 3). Este hito lleva esa decisión a las tres
skills del producto que tienen referencias —`normativa`, `plazos-y-territorio` y `jurisprudencia`— y la mide con el
arnés de H39 y H40 antes de darla por buena (ADR 0051, punto 4). Lo que gana la skill: que en **modo herramienta** (el
agente llama a `<applet>_<verbo>` del servidor `kitlegal mcp serve`) `normativa` abra las referencias que necesita como
las abre en **modo orden** (el agente ejecuta `kitlegal <applet> <verbo>`), y que ninguna de las tres abra menos que hoy.

Evidencia de partida, tal como la dan el hito y el ADR 0051 (`evidencias/sondeos-20261104/`; no medida en esta sesión):
con `normativa` tal como está, en modo herramienta `con_herramientas.md` se abre en 3 de 40 [2,6–19,9 %] y `normas.md`
en 7 [8,7–32,0 %], frente a 21 y 40 de 40 en modo orden; con `normativa-remisiones-en-orden`, en 38 [83,5–98,6 %] y 39
[87,1–99,6 %]; la orden de leer `como_se_cita.md` «antes de escribir la respuesta» se sigue en 20 de 40; y esa variante
mide 11 084 caracteres de cuerpo, 84 por encima del tope de 11 000 (ADR 0046, punto 2).

Comprobado en esta sesión, en el repositorio: `skills/` no ha cambiado desde la etiqueta `v0.9.0` (`git diff v0.9.0
HEAD -- skills/` vacío); `laboratorio/normativa/normativa-control/skills/normativa` es idéntica a `skills/normativa`
(`diff -r` vacío); las remisiones del cuerpo de cada skill son las que describe el hito:

- `normativa`: ocho referencias y ocho remisiones —«si la tienes, con su herramienta, tras leer [con_herramientas]»,
  «Toma el `BOE-A-…` de la tabla de [normas]; si no está, búscala con [buscar_la_norma]», «Sin fecha —con ella,
  [redaccion_en_una_fecha]—», «Varios bloques a la vez: [varios_bloques]. Si una orden no termina con `0`, antes de
  seguir: [lecturas_que_fallan]», «repásalas, y las dudas, en [como_se_cita]» y «Por cada entrada `version-obsoleta` …,
  lee [redaccion_modificada]»—;
- `plazos-y-territorio`: una sección «Referencias», antes del «Protocolo», con la orden generada de `administrativo.md`
  y una lista de `jerarquia_normativa.md`, `municipios.md` y `calculo_de_plazos.md` sin el verbo que manda leer;
- `jurisprudencia`: «Antes de dar la consulta, lee [references/consulta_preparada.md]» al final del paso 2, «Con cada
  documento, antes del cotejo, lee [references/documento_aportado.md]» al final del paso 3, y «lee «La búsqueda por
  materia» en [references/consulta_preparada.md]» en el paso 6.

## Clarifications

### Session 2026-11-05

- Q: ¿La regla ampliada de FR-041 (contar como apertura la orden de Bash que hace `cd` a `<skill>/references` y nombra `<f>` por su nombre solo) se aplica también a la tabla de la familia del laboratorio de H39 (`articulos-del-boe`), o solo a las tablas de las familias de las skills del producto? → A: B. Solo a las familias de las skills del producto (`normativa`, `plazos-y-territorio`, `jurisprudencia`); la de H39 se sigue contando con su contrato §2, byte a byte (auto: conservadora; criterio d; fuente: hito H41 Entrega; ADR 0051, Consecuencias; `specs/032-h40-por-qu-e/contracts/tabla.md` §1; constitución, «Criterio de decisión autónoma», puntos 2 y 4)
- Q: ¿De dónde toma `make laboratorio-tabla` la lista de referencias de `<skill>-producto` «tal como se instalaron» en el sondeo, si `skills/<skill>` puede haber cambiado entre el sondeo y el recuento? → A: A, con la lista de nombres y no con la copia. El arnés deja en el directorio del sondeo los nombres de los ficheros de `references/` instalados, en orden de nombre; la tabla toma de ahí las columnas; sin lista legible, defecto `inesperado`, código 1 (auto: criterio c; fuente: FR-040, FR-022 y Edge Cases del spec; hito H41, Entrega; `specs/032-h40-por-qu-e/contracts/arnes.md` §3; constitución, «Criterio de decisión autónoma», punto 1)
- Q: En la regla de FR-041, ¿qué cuenta como nombrar `<f>` «por su nombre solo» tras el `cd`? → A: B. Una palabra de una orden simple posterior de la misma orden que sea exactamente `<f>` o `./<f>` (partido como en sh, con comillas); no cuentan otro directorio delante ni subcadenas (`mis-normas.md`, `normas.md.bak`) (auto: criterio c; fuente: hito H41, Entrega; FR-041 y Assumptions; ADR 0051, Lo medido y Consecuencias; `evidencias/sondeos-20261104/guiones/recuento.py` l. 61; `specs/031-h39-c-omo-se/contracts/tabla.md` §2; constitución, «Criterio de decisión autónoma», punto 1)

## User Scenarios & Testing *(mandatory)*

### User Story 1 - Cada remisión de las tres skills es una orden de leer (Priority: P1)

El agente que usa `normativa`, `plazos-y-territorio` o `jurisprudencia` encuentra, para cada referencia, una orden que
le dice que la lea, cuándo y qué trae, antes de la primera orden o llamada o en el paso que la usa, también en modo
herramienta.

**Why this priority**: es el objeto del hito; sin ello el defecto de `normativa` en modo herramienta sigue abierto (ADR
0051, punto 4).

**Independent Test**: un test de `make ci` recorre el cuerpo de cada una de las tres skills y comprueba, por cada
enlace a `references/<f>`, la orden y su sitio según la lista de remisiones del plan.

**Acceptance Scenarios**:

1. **Dado** el árbol del hito, **Cuando** corre `make ci`, **Entonces** cada referencia de las tres skills tiene en el
   cuerpo al menos una remisión cuya frase contiene «lee» o «Lee» antes del enlace, y ninguna frase con un enlace a una
   referencia contiene «si la tienes», «antes de escribir la respuesta» ni «antes de dar la consulta» (FR-001 a FR-003,
   SC-001).
2. **Dado** `skills/normativa` tras el hito, **Cuando** se lee su cuerpo, **Entonces** las órdenes de leer
   `con_herramientas.md` y `normas.md` van sin condición antes de la primera orden o llamada, y las de
   `redaccion_en_una_fecha.md`, `varios_bloques.md`, `lecturas_que_fallan.md`, `buscar_la_norma.md` y
   `redaccion_modificada.md` conservan la condición del caso (FR-004).
3. **Dado** `skills/jurisprudencia` tras el hito, **Cuando** se lee su paso 2, **Entonces** la orden de leer
   `consulta_preparada.md` va antes de pedir `cita_preparar` o su orden (FR-006).
4. **Dado** las tres skills tras el hito, **Cuando** corre `make ci`, **Entonces** el cuerpo de cada skill del producto
   tiene 11 000 caracteres como mucho y su frontmatter es idéntico byte a byte al de su `<skill>-control` (FR-008,
   FR-010, SC-002, SC-003).

---

### User Story 2 - Lo que sale del cuerpo para hacer sitio está entero en su referencia (Priority: P1)

Para que las órdenes quepan sin pasar de 11 000 caracteres, lo que no necesita toda respuesta sale del cuerpo a una
referencia que el cuerpo manda leer en el paso que lo usa, sin quitar ni cambiar de sentido ninguna regla.

**Why this priority**: sin sitio, `normativa` no puede llevar las órdenes que la medida pide (ADR 0051, consecuencias).

**Independent Test**: un test de `make ci` comprueba, con la lista de movimientos del plan, que cada texto movido está
en su referencia y ya no en el cuerpo.

**Acceptance Scenarios**:

1. **Dado** un texto que el plan mueve del cuerpo de una skill a una referencia, **Cuando** corre `make ci`,
   **Entonces** el texto está en esa referencia palabra por palabra y no está en el cuerpo, y el cuerpo manda leer esa
   referencia en el paso que lo usa (FR-009, SC-004).
2. **Dado** el cuerpo de una skill tras el hito, **Cuando** se compara con el de su `<skill>-control`, **Entonces**
   todo lo que cambia es una remisión reescrita o un texto movido de la lista del plan (FR-007, FR-009, SC-004).

---

### User Story 3 - El arnés mide `control` frente a `producto` en las tres familias (Priority: P1)

La persona lanza con `make evals-laboratorio` las variantes de una familia —`normativa`, `plazos-y-territorio` o
`jurisprudencia`—, cada una con su `control` congelada y su `producto` instalada desde `skills/`, con cuatro evals de la
skill y sus lecturas grabadas.

**Why this priority**: es lo que hace posible la medida de la aceptación.

**Independent Test**: en `make ci`, con un sustituto de `claude` sin modelo, una petición válida de cada familia abre
n/4 sesiones por eval en cada variante y modo pedidos, cada una con la skill de su variante, y una inválida no abre
ninguna.

**Acceptance Scenarios**:

1. **Dado** dos peticiones, como en H40, con 40 sesiones por variante y modo: `normativa-control` y
   `normativa-producto` en los dos modos, y `normativa-remisiones-en-orden` en modo herramienta, **Cuando** la persona
   lanza el arnés con cada una, **Entonces** abre diez sesiones de cada una de las evals 04, 05, 15 y 22 de `normativa`
   por variante y modo, 160 con la primera y 40 con la segunda, 200 en total, cada una con solo la skill de su variante
   instalada (FR-030, FR-032).
2. **Dado** una petición de `plazos-y-territorio-control` y `plazos-y-territorio-producto` en los dos modos, con 40
   sesiones por variante y modo, **Cuando** la persona lanza el arnés, **Entonces** abre diez sesiones de cada una de las
   cuatro evals de la familia por variante y modo, 160 en total; y lo mismo con `jurisprudencia` (FR-030, FR-032).
3. **Dado** una sesión de `<skill>-producto`, **Cuando** el sustituto de `claude` lista su skill instalada, **Entonces**
   sus ficheros son idénticos byte a byte a los de `skills/<skill>` del árbol en ese momento (FR-022, SC-006).
4. **Dado** una variante que no existe, una variante en un modo que no le corresponde, un número de sesiones que no es
   un múltiplo de 4 mayor que 0 o una petición que mezcla variantes de dos familias, **Cuando** la persona lanza el arnés, **Entonces**
   sale con un código distinto de 0, dice qué no vale y no abre ninguna sesión (FR-030, SC-007).

---

### User Story 4 - La tabla cuenta la lectura con `cd` al directorio de las referencias (Priority: P1)

Con las transcripciones de un sondeo, `make laboratorio-tabla` da una fila por variante y modo de la familia, y cuenta
como apertura de una referencia la orden de `Bash` que hace `cd` al directorio de las referencias de la skill y la lee
por su nombre solo, la lectura que no contaba en 9 de las 200 sesiones de H40.

**Why this priority**: sin ella, la tabla de la aceptación cuenta de menos en las celdas que deciden.

**Independent Test**: en `make ci`, sin modelo, el guion da los recuentos exactos de transcripciones fijas que leen con
`cd … && cat <f>` y que no leen.

**Acceptance Scenarios**:

1. **Dado** una sesión cuya orden de `Bash` es `cd <ruta>/normativa/references && cat normas.md`, **Cuando** el guion la
   cuenta, **Entonces** cuenta una apertura de `normas.md` en ese turno (FR-041, SC-005).
2. **Dado** una sesión con `cd <ruta>/normativa/references` y, en otra orden posterior, `cat normas.md`; o con
   `cd <ruta>/jurisprudencia/references && cat normas.md` en la familia de `normativa`; o con
   `cd <ruta>/normativa && cat normas.md`; o con `cat normas.md && cd <ruta>/normativa/references`, **Cuando** el guion
   la cuenta, **Entonces** no cuenta ninguna apertura de `normas.md` (FR-041, SC-005).
   Tras el `cd`, también cuentan `cat "normas.md"` y `head -n 40 ./normas.md`; no cuentan `cat otra/normas.md`,
   `cat mis-normas.md` ni `cat normas.md.bak` (FR-041, SC-005).
3. **Dado** las suites de aceptación congeladas de H39 y H40, **Cuando** corre `make ci`, **Entonces** pasan sin cambiar
   ningún fichero suyo (FR-042, SC-008).
4. **Dado** un sondeo de `plazos-y-territorio` o de `jurisprudencia`, **Cuando** el guion lo cuenta, **Entonces** da una
   fila por variante y modo con las sesiones que activan la skill de la familia antes de la primera herramienta y, por
   cada referencia de la familia, las que la abren y su turno, con su intervalo de Wilson al 95 % (FR-040, SC-005).

---

### User Story 5 - La persona mide y decide antes de fusionar (Priority: P2)

Con la propuesta de cambio abierta, la persona lanza el sondeo de las trece filas con `claude-sonnet-5-5`, deja sus
tablas, una por familia, en `evidencias/` y las añade, con su lectura, a la propuesta; solo la fusiona si cumplen la
aceptación.

**Why this priority**: es la aceptación del hito, pero la hace una persona fuera del run (ADR 0032).

**Independent Test**: el informe final del run dice que la propuesta queda pendiente del sondeo; antes de fusionar,
`evidencias/` tiene las tablas y la propuesta su lectura.

**Acceptance Scenarios**:

1. **Dado** la rama con el arnés y las tres skills reorganizadas, **Cuando** la persona lanza el sondeo con
   `claude-sonnet-5-5` y 40 sesiones por variante y modo en las trece filas de FR-032, **Entonces** las tablas de las
   tres familias, con trece filas entre las tres, quedan en `evidencias/` con el modelo y la versión de Claude Code
   anotados (FR-040, FR-052).
2. **Dado** las tablas, **Cuando** en modo herramienta el intervalo de `normativa-producto` en `con_herramientas.md` y en
   `normas.md` no se toca con el de `normativa-control` y sí con el de `normativa-remisiones-en-orden`, y en ninguna
   familia ni modo `producto` queda por debajo de `control` en las sesiones que activan la skill antes de la primera
   herramienta ni en ninguna referencia a cuya condición llegan las evals, **Entonces** la reorganización vale y la
   propuesta puede fusionarse (FR-052, SC-010).
3. **Dado** unas tablas que no cumplen alguna de esas condiciones, **Cuando** la persona las lee, **Entonces** la propuesta
   no se fusiona y lo medido va a un ADR que dice qué medir después (FR-052, SC-010).

---

### Edge Cases

- **El cuerpo de una skill pasa de 11 000 caracteres** con las órdenes nuevas: `make ci` falla (FR-008).
- **Un cambio en `laboratorio/<skill>/<skill>-control`**, a mano o por una tarea: `make ci` falla y nombra el fichero
  que difiere, falta o sobra (FR-020).
- **`skills/<skill>` cambia después de lanzar el hito**: `<skill>-control` no cambia con ella (FR-020); `<skill>-producto`
  la sigue, porque se instala desde `skills/` en cada sondeo (FR-022); y si cambia entre el sondeo y la tabla, la tabla
  conserva las columnas de lo medido, porque las toma de la lista que el sondeo dejó (FR-033).
- **Una referencia nueva** en `<skill>-producto`, a la que se mueve texto del cuerpo: la tabla le da su columna; en la
  fila de una variante que no la tiene, la celda va vacía, no a cero (FR-040).
- **Una referencia a cuya condición no llega ninguna eval del sondeo** (en `normativa`, `buscar_la_norma.md`,
  `lecturas_que_fallan.md`, `redaccion_modificada.md` y `varios_bloques.md`, según el ADR 0051): sale «0 de n» con su
  intervalo, como en H40, y no entra en la comparación de la aceptación (FR-040, FR-052).
- **Sesión sin terminar o fila sin ninguna sesión terminada**: como en H40 (FR-040).
- **`DIRECTORIO` con sondeos de más de una familia** (p. ej., el padre de los lanzamientos de la aceptación): la tabla
  es de la familia de la primera variante que lee; un directorio de variante de otra familia, también de H39, es un
  defecto que nombra su ruta, con código 1 y sin tabla (FR-040). Cada familia se cuenta con su directorio.
- **Transcripción que el guion no sabe leer** (manipulada a mano, truncada por algo que no es el arnés): defecto
  `inesperado`, código 1 (ADR 0023), sin más casos.

## Requirements *(mandatory)*

### Functional Requirements

**Las remisiones de las tres skills** (`skills/normativa`, `skills/plazos-y-territorio`, `skills/jurisprudencia`)

- **FR-001**: Cada referencia de las tres skills tiene en el cuerpo de su `SKILL.md` al menos una remisión que es una
  orden de leerla (ADR 0051, punto 1): en la misma frase, el verbo «lee» (o «Lee») antes del enlace, el enlace a
  `references/<f>`, cuándo leerla (la condición del caso, o ninguna) y lo que trae. La frase de un enlace va desde el
  final de la anterior —un punto seguido de espacio o de salto de línea, una línea en blanco, o la marca de una viñeta o
  de un paso— hasta el final de la suya; el punto de `.md` dentro de un enlace no la cierra. Ningún enlace a una referencia queda
  como mención dentro de otra orden, tras una condición sobre el agente, entre rayas o como sitio de dudas. El plan da
  la lista de remisiones de cada skill —texto, sitio y condición— y un test de `make ci` falla si un enlace del cuerpo a
  una referencia no está en una frase con «lee» o «Lee» antes del enlace, o si una referencia no tiene ninguna remisión.
- **FR-002**: La condición que pone el caso se queda —una fecha en la pregunta, varios bloques, una orden que no termina
  con `0`, una entrada `version-obsoleta`, un municipio que no sale con `0`, una norma que no está en la tabla, un
  documento delante, una pregunta por materia—; la que pone el agente («si la tienes») se quita, y lo que siempre hace
  falta se manda sin condición (ADR 0051, punto 2). El test de FR-001 falla si una frase con un enlace a una referencia
  contiene «si la tienes».
- **FR-003**: Cada orden de leer va antes de la primera orden o llamada de la skill, o en el paso que usa la referencia
  y antes de la orden o llamada de ese paso; nunca como orden de leer antes de escribir o dar la respuesta (ADR 0051,
  punto 3). Lo que toda respuesta necesita sigue en el cuerpo (ADR 0046, punto 1; ADR 0050, punto 1). El test de FR-001
  falla si una frase con un enlace a una referencia contiene «antes de escribir la respuesta» o «antes de dar la
  consulta».
- **FR-004**: En `normativa`, el punto de partida son las cinco remisiones de `normativa-remisiones-en-orden`
  (`con_herramientas.md`; `normas.md` con `buscar_la_norma.md`; `redaccion_en_una_fecha.md`; `varios_bloques.md` con
  `lecturas_que_fallan.md`), completadas con lo que trae cada referencia (FR-001):
  - `con_herramientas.md` y `normas.md` se mandan leer sin condición antes de la primera orden o llamada;
  - `buscar_la_norma.md`, con la norma que no está en la tabla de `normas.md`; `redaccion_en_una_fecha.md`, con una fecha
    en la pregunta y antes de leer el bloque; `varios_bloques.md`, antes de leer varios bloques a la vez;
    `lecturas_que_fallan.md`, si una orden no termina con `0` y antes de seguir; `redaccion_modificada.md`, por cada
    entrada `version-obsoleta`;
  - `como_se_cita.md` tiene una orden de leer que cumple FR-001 a FR-003 y no es la de «antes de escribir la respuesta»
    de `normativa-remisiones-en-orden`; su sitio, y por qué, lo dice el plan. La forma de cita que toda respuesta usa
    sigue en el cuerpo.
- **FR-005**: En `plazos-y-territorio`, las remisiones a `jerarquia_normativa.md`, `municipios.md` y
  `calculo_de_plazos.md` pasan a ser órdenes de leer con FR-001 a FR-003, con la condición del caso que hoy da la lista
  (el nivel que regula algo y la norma que prevalece; presentar un territorio, un municipio que falta o no sale con `0` y
  su cobertura; un plazo sin calcular en los municipios que cuentan, una lectura fallida o con `redaccion_posterior`).
  La orden de `administrativo.md` y el bloque generado que la contiene no cambian.
- **FR-006**: En `jurisprudencia`, la orden de leer `consulta_preparada.md` va en el paso 2 antes de pedir
  `cita_preparar` o su orden, y no «antes de dar la consulta»; las de `documento_aportado.md` (antes del cotejo) y la del
  paso 6 (la búsqueda por materia) siguen siendo órdenes y cumplen FR-001 a FR-003.
- **FR-007**: Fuera de las remisiones (FR-001 a FR-006) y de lo que se mueve (FR-009), las tres skills no cambian:
  ninguna regla se quita ni cambia de sentido, una referencia de `<skill>-control` solo cambia por lo que se le mueve, y
  `consulta-juridica` no se toca. El plan declara cada cambio del cuerpo como remisión o como movimiento, y un test de
  `make ci` falla si el cuerpo de una skill, deshechos los cambios declarados, difiere del de su `<skill>-control`, o si
  una referencia de `<skill>-control`, quitado lo que se le mueve, difiere de la suya.
- **FR-008**: El cuerpo de `SKILL.md` de cada skill de `skills/` —desde la línea que sigue a la que cierra el
  frontmatter— tiene 11 000 caracteres como mucho (ADR 0046, punto 2). El test de `make ci` que hoy lo comprueba sale en
  rojo, con el nombre de la skill y su recuento, si uno pasa de 11 000.
- **FR-009**: Para hacer sitio sin pasar de 11 000 caracteres, lo que no necesita toda respuesta sale del cuerpo a una
  referencia —existente o nueva— que el cuerpo manda leer con una orden (FR-001 a FR-003) en el paso que lo usa (ADR
  0050 y 0051, consecuencias). Lo movido está entero en su referencia, palabra por palabra: solo pueden cambiar los
  espacios y saltos de línea y la ruta de un enlace, que se escribe relativa a la referencia. El plan lista cada
  movimiento —qué texto, de dónde, a qué referencia— con su porqué, y un test de `make ci` falla si un texto de la lista
  no está en su referencia o sigue en el cuerpo.
- **FR-010**: El frontmatter de cada una de las tres skills —también su `description`— es idéntico byte a byte al de su
  `<skill>-control`, y la tabla de comandos y las referencias generadas desde `data/` son las que da `make skills-sync`.
  `make ci` falla si el frontmatter difiere o si `make skills-sync` deja diferencias.

**Las variantes del laboratorio**

- **FR-020**: Cada familia —`normativa`, `plazos-y-territorio` y `jurisprudencia`— tiene una variante
  `<skill>-control`: `skills/<skill>` tal como está en `main` al lanzar el hito, la de `v0.9.0`, congelada en
  `laboratorio/<skill>/`. La de `normativa` ya existe (`laboratorio/normativa/normativa-control`) y deja de compararse
  con `skills/normativa`. `make ci` falla, y nombra el fichero, si un fichero de una `<skill>-control` difiere en un byte
  de ese estado, falta o sobra.
- **FR-021**: `normativa-remisiones-en-orden`, `normativa-cuerpo-corto` y `normativa-testigo-en-referencia` siguen
  comparándose con `normativa-control`, como hasta ahora: `make ci` falla si una difiere de ella fuera de lo que declara.
- **FR-022**: Cada familia tiene una variante `<skill>-producto`, que es `skills/<skill>` en el momento del sondeo: el
  arnés la instala desde `skills/`, y no hay copia versionada suya en `laboratorio/` ni en ningún otro sitio. Un test de
  `make ci` falla si la skill que ve una sesión de `<skill>-producto` difiere en un byte de `skills/<skill>` del árbol, o
  si el repositorio tiene un directorio de variante `<skill>-producto`.
- **FR-023**: Ninguna variante versionada (`<skill>-control` y las de H40) se empotra en el binario, la pone
  `kitlegal skills install` ni la elige o mide el job de evals; `make ci` falla si una de las tres cosas deja de
  cumplirse. El tope de 11 000 caracteres no se aplica a las variantes versionadas: rige `skills/`.

**El arnés del sondeo**

- **FR-030**: `make evals-laboratorio` lanza las variantes de las familias `normativa`, `plazos-y-territorio` y
  `jurisprudencia`, una familia por invocación. `<skill>-control` y `<skill>-producto` se miden en los dos modos;
  `normativa-remisiones-en-orden`, `normativa-cuerpo-corto` y `normativa-testigo-en-referencia`, solo en modo
  herramienta, como en H40. Por variante y modo pedidos, abre el número de sesiones que pide la persona, repartidas por
  igual entre cuatro evals de la familia (diez por eval con las 40 de la aceptación), servidas por sus lecturas
  grabadas, sin red a las fuentes (ADR 0043); cada sesión tiene instalada solo la skill de su variante, con el nombre de
  la skill, y ninguna otra. Las de `normativa` son sus evals 04, 05, 15 y 22. Las de `plazos-y-territorio` y
  `jurisprudencia` son cuatro evals de la skill que ya tienen sus lecturas grabadas y que entre las cuatro llegan a la
  condición del mayor número de remisiones; el research del hito da la elección, su porqué y, por familia, las
  referencias a cuya condición llega cada eval. Pedir una variante que no existe, una variante en un modo que no le
  corresponde, un número de sesiones que no es un múltiplo de 4 mayor que 0 o variantes de dos familias es un error de
  argumentos: sale con un código distinto de 0, dice qué no vale, uno por línea y con el nombre del argumento de `make`,
  y no abre ninguna sesión. Los demás argumentos (`MODOS`, `MODELO`, `CLAUDE_CODE`, `CONCURRENCIA`, `DIRECTORIO`) y la
  versión de Claude Code se comprueban como en H40 (`specs/032-h40-por-qu-e/contracts/arnes.md` §2).
- **FR-031**: El arnés lo lanza una persona, con la credencial del sondeo (ADR 0032); ni `make ci`, ni un gancho, ni el
  job de evals, ni un paso del workflow lo ejecutan. El run no graba nada: no añade ni cambia ninguna grabación de
  `testdata/` y no escribe en `evidencias/`.
- **FR-033**: Al instalar `<skill>-producto` desde `skills/`, el arnés escribe en el directorio del sondeo los nombres de
  los ficheros de `references/` de la skill que instaló, en orden de nombre; solo los nombres, ni la copia de la skill
  ni una huella de su contenido. La forma del fichero la fija el plan, sin cambiar lo que los guiones de H39 y H40 leen
  de `sondeo.json`, para que sus suites congeladas sigan en verde (FR-042). `make ci` comprueba que el arnés deja la
  lista y que la tabla toma de ella las columnas de `<skill>-producto`.
- **FR-032**: Las filas del sondeo de la aceptación son trece: `<skill>-control` y `<skill>-producto` de cada familia, en
  los dos modos (doce), y `normativa-remisiones-en-orden` en modo herramienta (una); 40 sesiones por fila, 520 en total.

**La tabla**

- **FR-040**: `make laboratorio-tabla` cuenta los sondeos de las tres familias con las reglas de la tabla de H40
  (`specs/032-h40-por-qu-e/contracts/tabla.md` §2 y §3), con la skill y las referencias de la familia: una fila por
  variante y modo, con el modelo y la versión de Claude Code; las sesiones que activan la skill antes de la primera
  herramienta de kitlegal; por cada referencia, las sesiones que la abren y el turno de su primera apertura; y las
  sesiones sin terminar, fuera de `n`. Cada recuento con su intervalo de Wilson al 95 %. Las referencias de la tabla son
  todas las de las variantes del sondeo, las de `<skill>-producto` tal como se instalaron en él, en orden de nombre; en
  la fila de una variante que no tiene una referencia, su celda va vacía. Las de `<skill>-producto` salen de la lista
  que el arnés deja en el directorio del sondeo (FR-033); las de las variantes versionadas, de su declaración. Un sondeo
  de `<skill>-producto` sin esa lista, o con una que el guion no sabe leer, es un directorio manipulado: defecto
  `inesperado`, código 1 y sin tabla (ADR 0023). Una tabla cuenta una sola familia, la de la primera variante que lee
  en orden de nombre, como en H40 (`specs/032-h40-por-qu-e/contracts/tabla.md` §1): un directorio de variante de otra
  familia, también de H39, es un defecto que nombra su ruta, con código 1 y sin tabla. La columna «Cumplen la regla» solo va en la
  tabla de `normativa`, como en H40. `make ci` comprueba el guion con transcripciones fijas de cada familia y falla si un
  recuento, un turno o un extremo de un intervalo difiere.
- **FR-041**: Una orden de `Bash` también abre la referencia `<f>` de la skill de la familia si hace `cd` a un directorio
  cuya ruta termina en `<skill>/references` (con o sin barra final) y después, en la misma orden, nombra `<f>` por su
  nombre solo, sin directorio delante. «Nombra `<f>` por su nombre solo» es: una orden simple posterior de la misma
  `input.command` tiene una palabra que es exactamente `<f>` o `./<f>`. Las palabras se parten como en sh (comillas
  simples y dobles, barra invertida) y las órdenes simples por `|`, `||`, `&&`, `;`, `&` y saltos de línea, partidas
  como en `specs/031-h39-c-omo-se/contracts/tabla.md` §2.
  Cuentan `cat normas.md`, `cat "normas.md"`, `cat 'normas.md'` y `head -n 40 ./normas.md`; no cuentan
  `cat otra/normas.md`, `cat /x/normas.md`, `cat mis-normas.md` ni `cat normas.md.bak`. No la abre si el `cd` va a otro
  directorio (el de la skill, el de las referencias de otra skill), si `<f>` va antes del `cd` o si el `cd` y el nombre
  van en órdenes distintas. Lo demás de la regla de H40 no cambia, y la regla de H40 que cuenta una orden que contiene
  `references/<f>` se aplica aparte. La regla ampliada rige solo en las familias de las skills del producto
  (`normativa`, `plazos-y-territorio` y `jurisprudencia`); la tabla de la familia del laboratorio de H39
  (`articulos-del-boe`) se cuenta exactamente con su contrato (`specs/031-h39-c-omo-se/contracts/tabla.md` §2), como
  hoy. `make ci` lo comprueba con sesiones fijas con las formas que cuentan y con las que no, y falla si un
  recuento difiere.
- **FR-042**: Las suites de aceptación congeladas de H39 (`specs/031-h39-c-omo-se/aceptacion/`) y de H40
  (`specs/032-h40-por-qu-e/aceptacion/`) siguen en verde en `make ci` sin cambiar ningún fichero suyo.

**La documentación, el cierre y la aceptación de la persona**

- **FR-050**: `CHANGELOG.md` (*Unreleased*) da el cambio de las tres skills. `docs/SKILLS.md` recoge lo que la
  reorganización dé y la guía no recoja; si no da nada, el plan lo dice y por qué.
- **FR-051**: Las evals de las skills del producto salen en verde en la medición del cierre, en los dos modos.
- **FR-052**: Antes de fusionar, la persona lanza el sondeo de FR-032 con `claude-sonnet-5-5` y 40 sesiones por variante
  y modo y deja en `evidencias/` sus tablas, una por familia (FR-040). La reorganización vale si cumple todo esto: en modo herramienta,
  `normativa-producto` abre `con_herramientas.md` y `normas.md` con un intervalo que no se toca con el de
  `normativa-control` y sí con el de `normativa-remisiones-en-orden`; y en cada familia y modo, `producto` no queda por
  debajo de `control` —con un intervalo que no se toca con el de `control`— en las sesiones que activan la skill antes de
  la primera herramienta, ni en ninguna referencia a cuya condición llegan las evals (las que da el research, FR-030). El
  run no lanza el sondeo; las tablas y su lectura se añaden a la propuesta de cambio antes de fusionarla, y la propuesta no
  se fusiona sin ellas. Si el sondeo no cumple, la propuesta no se fusiona, y lo que se mida va a un ADR que dice qué
  medir después. El informe final del run lo dice.

### Key Entities

- **Remisión**: la frase del cuerpo de un `SKILL.md` que enlaza una referencia; con su verbo, su condición, su sitio y
  lo que trae (FR-001 a FR-006).
- **Movimiento**: un texto que sale del cuerpo a una referencia, con su porqué (FR-009).
- **Familia del laboratorio**: las variantes de una skill del producto, con sus evals del sondeo y sus referencias.
- **`<skill>-control`**: la skill de `v0.9.0`, congelada en `laboratorio/<skill>/` (FR-020).
- **`<skill>-producto`**: `skills/<skill>` en el momento del sondeo, sin copia versionada (FR-022).
- **Sondeo y tabla**: las sesiones de las variantes por modo, y su recuento por fila (FR-030 a FR-041).

### Uso, de fuera adentro

- **El cuerpo de las tres skills** (FR-001 a FR-009). Lo lee el agente una vez por pregunta, al activar la skill; tiene
  11 000 caracteres como mucho (FR-008), y no crece con el uso: cientos de normas y miles de bloques consultados no lo
  tocan.
- **Las referencias que se mandan leer** (FR-001 a FR-006, FR-009). El agente lee cada una una vez por pregunta, cuando
  llega su condición; en `normativa`, `con_herramientas.md` (2,4 KB hoy) y `normas.md` (3,4 KB) en toda pregunta, y las
  demás (0,9 a 2,7 KB) solo con su condición: por pregunta, el cuerpo (11 000 caracteres como mucho) más unos 5,8 KB de
  referencias que hoy, en modo herramienta, casi nunca se leen, y hasta otros 2,7 KB por cada condición que llega. Lo
  que gana la skill es justo esto: que en modo herramienta se lean como en
  modo orden. `normas.md` crece con las normas declaradas en `data/normas.yaml` al añadir un paquete, no con las
  consultadas; las demás, solo cuando un hito las cambia. Lo movido (FR-009) se lee solo en el paso que lo usa, en vez
  de en toda pregunta con el cuerpo. No llevan señales que caduquen: su contenido no depende de la consulta.
- **Las tablas del sondeo** (FR-040, FR-041). Las lee la persona una vez, para decidir la fusión: las de las tres
  familias, con trece filas entre las tres, cada una con la activación y una celda por referencia (ocho o más en `normativa`, cuatro en `plazos-y-territorio`, dos en
  `jurisprudencia`), más el detalle de 520 sesiones. Su tamaño lo fijan las variantes, los modos y las sesiones pedidas,
  no el uso del producto; dejan de leerse cuando la persona decide, y un sondeo nuevo da otras tablas.

## Success Criteria *(mandatory)*

### Measurable Outcomes

- **SC-001**: El 100 % de las referencias de las tres skills con al menos una remisión que es una orden, y 0 frases con
  un enlace a una referencia sin «lee» o «Lee» antes del enlace o con «si la tienes», «antes de escribir la respuesta» o
  «antes de dar la consulta». Control: el test de FR-001 en `make ci` falla con la primera.
- **SC-002**: 0 skills de `skills/` con un cuerpo de más de 11 000 caracteres. Control: el test de `make ci` de FR-008.
- **SC-003**: 0 bytes de diferencia entre el frontmatter de cada una de las tres skills y el de su `<skill>-control`, y 0
  diferencias tras `make skills-sync`. Control: los tests de FR-010 en `make ci`.
- **SC-004**: El 100 % de los textos de la lista de movimientos del plan, en su referencia y fuera del cuerpo, y 0
  diferencias con `<skill>-control`, en el cuerpo y en las referencias, fuera de lo declarado. Control: los tests de
  FR-007 y FR-009 en `make ci` fallan con el primer texto o la primera diferencia.
- **SC-005**: Los recuentos, turnos y extremos de Wilson de las transcripciones fijas de las tres familias, exactos,
  también con las lecturas por `cd` y las que no lo son. Control: los tests de FR-040 y FR-041 en `make ci` fallan con
  el primer valor que difiere.
- **SC-006**: 0 bytes de diferencia entre la skill de una sesión de `<skill>-producto` y `skills/<skill>`, y 0
  directorios de variante `<skill>-producto` en el repositorio. Control: el test de FR-022 en `make ci`.
- **SC-007**: Con una petición inválida (FR-030), 0 sesiones abiertas; con una válida, n/4 por eval en cada variante y
  modo pedidos de cada familia. Control: un test de `make ci` con un sustituto de `claude` que cuenta sus invocaciones,
  sin modelo, falla si una inválida abre alguna o una válida no abre n/4 por eval.
- **SC-008**: 0 bytes de diferencia en cada `<skill>-control` respecto de `skills/<skill>` de `v0.9.0`; 0 bytes fuera de
  lo declarado entre cada variante de H40 y `normativa-control`; 0 ficheros de variantes versionadas en el binario, en
  la instalación y en las evals del job; las suites congeladas de H39 y H40 en verde. Control: los tests de FR-020,
  FR-021, FR-023 y FR-042 en `make ci`.
- **SC-009**: Las evals de las skills del producto en verde en la medición del cierre, en los dos modos. Control: el job
  de evals del cierre sale en rojo si una no pasa su umbral en un modo, y el informe final da sus cifras (FR-051).
- **SC-010**: Antes de fusionar, las tablas de las tres familias, con trece filas de 40 sesiones entre las tres, están
  en `evidencias/` y cumplen las
  condiciones de FR-052. Lo mide la persona (capa 3): el sondeo abre sesiones con modelo, que ni `make ci` ni el run
  abren (ADR 0032), y ningún control automático lo hace cumplir; la propuesta no se fusiona sin las tablas y su lectura.

## Criterios del hito (literales)

**Controles**:

- `make ci` comprueba que cada `<skill>-control` no cambia (FR-020, SC-008);
- que cada variante de H40 difiere de `normativa-control` solo en lo que declara (FR-021, SC-008);
- que ninguna variante se empotra ni se instala (FR-023, SC-008);
- el cuerpo de cada skill, como hasta ahora, con 11 000 caracteres como mucho (FR-008, SC-002);
- la regla de lectura ampliada, con sesiones grabadas que leen con `cd … && cat <f>` y que no leen (FR-041, SC-005).

Las evals de las skills del producto salen en verde en la medición del cierre, en los dos modos (FR-051, SC-009).

**Aceptación**: el sondeo de la persona, con `claude-sonnet-5-5` y 40 sesiones por variante y modo, deja su tabla en
`evidencias/`, y la reorganización vale si cumple todo esto (FR-052, SC-010):

- en modo herramienta, `normativa-producto` abre `con_herramientas.md` y `normas.md` con un intervalo que no se toca
  con el de `normativa-control` y sí con el de `normativa-remisiones-en-orden` (ADR 0051, punto 4);
- en cada familia y modo, `producto` no queda por debajo de `control` en las sesiones que activan la skill antes de
  la primera herramienta, ni en ninguna referencia a cuya condición llegan las evals. «Por debajo» quiere decir con
  un intervalo que no se toca con el de `control`.

El run no lanza el sondeo. La tabla y su lectura se añaden a la propuesta de cambio antes de fusionarla, y la propuesta
no se fusiona sin ellas. Si el sondeo no cumple, la propuesta no se fusiona, y lo que se mida va a un ADR que dice qué
medir después.

## Assumptions

- Las cifras de partida (H40, `evidencias/sondeos-20261104/`) son las que citan el hito y el ADR 0051; esta sesión no
  las ha medido.
- «Lee `<f>` por su nombre solo» (FR-041) se cuenta como la regla de H40 cuenta `references/<f>` en una orden de
  `Bash`: cualquier orden simple que tenga `<f>` o `./<f>` como palabra tras el `cd` cuenta, no solo `cat`; así la
  ampliación no cambia qué es abrir, solo cómo se nombra el fichero. Esta sesión no ha visto qué forma usaron las 9
  sesiones de H40: sus transcripciones se guardan fuera del repositorio.
- `skills/` en `main` al lanzar el hito es la de `v0.9.0` (comprobado: `git diff v0.9.0 HEAD -- skills/` vacío), así
  que `<skill>-control` es las dos cosas a la vez.
- Las evals de `plazos-y-territorio` y `jurisprudencia` del sondeo son de las que ya existen y ya tienen sus lecturas
  grabadas (el hito lo dice); el run no graba ninguna (FR-031). `jurisprudencia` no consulta el CENDOJ: sus evals no
  leen fuentes por red (ADR 0036).
- En cada sesión del sondeo solo está instalada la skill de su variante, como en H40; por eso las cifras de
  `plazos-y-territorio` en modo herramienta no se comparan con las 58 de 60 de la #214, que instalaba las cuatro skills,
  sino con su `control` en el mismo sondeo.
- El modelo se pasa como identificador completo porque un alias lo resuelve el cliente (ADR 0031).

## Fuera de alcance

- Cambiar ningún umbral, límite ni tope, tampoco los 11 000 caracteres del cuerpo (ADR 0046, punto 2), los 900 de la
  descripción ni los umbrales del job de evals.
- Cambiar la descripción de una skill: la activación ya se mide en 40 de 40 en los dos modos (ADR 0050 y 0051).
- Cambiar el generador de la tabla de comandos o de las referencias generadas desde `data/`, incluida la orden generada
  de `administrativo.md` en `plazos-y-territorio`.
- `consulta-juridica`, que no tiene referencias.
- Cambiar las instrucciones del servidor `kitlegal mcp serve` o las descripciones de sus herramientas, que en modo
  herramienta comparten `control` y `producto` porque el arnés arranca el `kitlegal` del árbol
  (`specs/032-h40-por-qu-e/contracts/arnes.md` §3).
- Cambiar las evals que ya existen (sus preguntas o expectativas), también las cuatro de cada familia que elige el
  sondeo (FR-030).
- Evals nuevas, también las que llegarían a la condición de `buscar_la_norma.md`, `lecturas_que_fallan.md`,
  `redaccion_modificada.md` y `varios_bloques.md`, y otras evals de `normativa` en el sondeo que las 04, 05, 15 y 22.
- Medir con el juez en el sondeo, que cuenta lecturas y activación, sin modelo.
- Guardar las sesiones del job de evals como artefacto, que sigue en `backlog/medida-de-las-skills.md`.
- Medir en este sondeo `normativa-cuerpo-corto` o `normativa-testigo-en-referencia`, o `normativa-remisiones-en-orden`
  en modo orden; otros modelos que `claude-sonnet-5-5` u otros hosts que Claude Code.
- Instalar en las sesiones del sondeo otras skills que la de la variante de la fila.
- Que la tabla calcule si dos intervalos se tocan o dé un veredicto de aceptación: la lectura la hace la persona.
- Anotar en la tabla una huella de la skill `<skill>-producto` medida, o guardar en el directorio del sondeo una copia de
  la skill instalada: solo la lista de nombres de sus referencias (FR-033).
- Ampliar la regla de lectura con `cd` (FR-041) a la tabla de la familia del laboratorio de H39 (`articulos-del-boe`):
  se cuenta como hoy, byte a byte.
- Columnas de la tabla que FR-040 no enumera.
- Que el job de evals mida las variantes, o que `make ci` abra sesiones con modelo.
- Grabar lecturas nuevas o escribir en `evidencias/` desde el run, y lanzar el sondeo o escribir el ADR de un sondeo
  que no cumple dentro del run.
