# Feature Specification: H41 · Las skills del producto mandan leer sus referencias

**Feature Branch**: `034-h41-las-skills-del`

**Created**: 2026-11-06

**Status**: Draft

**Input**: el fichero del hito, `docs/roadmap/in-progress/H41-las-skills-mandan-leer-sus-referencias.md` (entrada de la
persona, ADR 0018), corregido con el ADR 0052 tras el primer run de H41 (`0f548c7d`, #245, cerrada sin fusionar).

## Contexto

El ADR 0051 decide que una remisión a una referencia es una orden de leerla: el verbo «lee», el enlace, cuándo y lo que
trae (punto 1); la condición que trae el caso se queda y la que pone el agente se quita (punto 2); la orden va antes de
la primera llamada o en el paso que usa la referencia, no al final (punto 3). El ADR 0052 precisa el punto 2: una
condición se queda si la trae el caso y el agente puede saber que se da sin juzgar lo que va a hacer, y una lista de
casos unidos con «y» no es una condición, sino varias.

Este hito relanza la reorganización con lo que midió la #245: lleva las órdenes a `normativa` y `jurisprudencia` como en
la #245, deja `plazos-y-territorio` como está (la de `v0.9.0`) y mide en el laboratorio tres formas de sus órdenes de
`municipios.md` y `calculo_de_plazos.md`, que aplicará H41.1. Lo que gana la skill: que en **modo herramienta** (el
agente llama a `<applet>_<verbo>` del servidor `kitlegal mcp serve`) `normativa` abra las referencias que necesita como
las abre en **modo orden** (el agente ejecuta `kitlegal <applet> <verbo>`), que ni `normativa` ni `jurisprudencia` abran
menos que hoy, y saber qué forma de orden se sigue en `plazos-y-territorio` antes de cambiarla.

Evidencia de partida, tal como la dan el hito y los ADR 0051 y 0052 (`evidencias/sondeos-20261104/` y
`evidencias/sondeos-20261106/`; no medida en esta sesión): en modo herramienta, `normativa` tal como está abre
`con_herramientas.md` en 3 de 40 y `normas.md` en 7 (H40), y en 0 y 9 de 40 en el sondeo de la #245, donde las
remisiones de la #245 las abren en 38 y 38 de 40; en `jurisprudencia`, las de la #245 se siguen como en
`jurisprudencia-control` en los dos modos; en `plazos-y-territorio`, la lista con la condición delante del verbo abre
`municipios.md` en 1 de 40 en modo orden, frente a 17 de `plazos-y-territorio-control`, y `calculo_de_plazos.md` cae de
10 a 3 de 10 en la eval 05 en modo orden.

Comprobado en esta sesión, en el repositorio: `git diff --quiet v0.9.0 HEAD -- skills/` sale con 0 (`skills/` es la de
`v0.9.0`); `diff -r laboratorio/normativa/normativa-control/skills/normativa skills/normativa` no da ninguna diferencia;
el cuerpo de `SKILL.md` (desde la línea que sigue a la que cierra el frontmatter) mide 10 929 caracteres en `normativa`,
10 842 en `jurisprudencia` y 10 930 en `plazos-y-territorio`; y la rama `033-h41-las-skills-del` conserva el spec, el
plan, los contratos y el código de la #245 (`git show 033-h41-las-skills-del:specs/033-h41-las-skills-del/…`).

## User Scenarios & Testing *(mandatory)*

### User Story 1 - Cada remisión de `normativa` y `jurisprudencia` es una orden de leer (Priority: P1)

El agente que usa `normativa` o `jurisprudencia` encuentra, para cada referencia, una orden que le dice que la lea,
cuándo y qué trae, antes de la primera orden o llamada o en el paso que la usa, también en modo herramienta.

**Why this priority**: es el objeto del hito; sin ello el defecto de `normativa` en modo herramienta sigue abierto (ADR
0051, punto 4).

**Independent Test**: un test de `make ci` recorre el cuerpo de las dos skills y comprueba, por cada enlace a
`references/<f>`, la frase de su remisión contra la lista de remisiones del plan.

**Acceptance Scenarios**:

1. **Dado** el árbol del hito, **Cuando** corre `make ci`, **Entonces** cada referencia de `normativa` y de
   `jurisprudencia` tiene en el cuerpo al menos una remisión cuya frase lleva «lee» o «Lee» antes del enlace, ninguna
   frase con un enlace a una referencia lleva «si la tienes», «antes de escribir la respuesta» ni «antes de dar la
   consulta», y la frase de cada remisión es la de la lista del plan (FR-001 a FR-003, SC-001).
2. **Dado** `skills/normativa` tras el hito, **Cuando** se lee su cuerpo, **Entonces** las órdenes de leer
   `con_herramientas.md` y `normas.md` van sin condición antes de la primera orden o llamada, y las de
   `buscar_la_norma.md`, `redaccion_en_una_fecha.md`, `redaccion_modificada.md`, `como_se_cita.md`, `varios_bloques.md`
   y `lecturas_que_fallan.md` llevan una condición del caso (FR-004).
3. **Dado** `skills/jurisprudencia` tras el hito, **Cuando** se lee su paso 2, **Entonces** la orden de leer
   `consulta_preparada.md` va antes de pedir `cita_preparar` o su orden; y en el paso 3, la de `documento_aportado.md`
   va antes de la viñeta que pide el cotejo (FR-005).
4. **Dado** el árbol del hito, **Cuando** corre `make ci`, **Entonces** el cuerpo de cada skill de `skills/` tiene
   11 000 caracteres como mucho, y el frontmatter de `normativa` y de `jurisprudencia` es idéntico byte a byte al de su
   `<skill>-control` (FR-008, FR-009, SC-002, SC-003).

---

### User Story 2 - Lo que sale del cuerpo de `normativa` está entero en su referencia (Priority: P1)

Para que las órdenes quepan en `normativa` sin pasar de 11 000 caracteres, dos textos que solo necesita un caso salen
del cuerpo a la referencia que el cuerpo manda leer en ese caso, enteros y palabra por palabra, como en la #245.

**Why this priority**: sin sitio, `normativa` no puede llevar las órdenes que la medida pide.

**Independent Test**: un test de `make ci` comprueba que cada texto movido está en su referencia y ya no está en el
cuerpo, y que nada más cambia respecto de `normativa-control`.

**Acceptance Scenarios**:

1. **Dado** el árbol del hito, **Cuando** corre `make ci`, **Entonces** «Nada de otra conversación…» está en
   `references/redaccion_modificada.md` y «Si una lectura falla, di qué bloques faltan y por qué.» en
   `references/lecturas_que_fallan.md`, palabra por palabra, y ninguno de los dos está en el cuerpo (FR-006, SC-004).
2. **Dado** el cuerpo y las referencias de `normativa` y de `jurisprudencia` tras el hito, **Cuando** se comparan con
   los de su `<skill>-control`, **Entonces** todo lo que cambia es una remisión de la lista del plan o uno de los dos
   movimientos (FR-007, SC-004).

---

### User Story 3 - `plazos-y-territorio` no cambia y sus tres formas se miden en el laboratorio (Priority: P1)

La skill del producto sigue siendo la de `v0.9.0`, con «Si no lo dice, **pregúntalo** y espera» en el paso 1. En
`laboratorio/plazos-y-territorio/` hay, junto a `plazos-y-territorio-control`, tres variantes que solo cambian la forma
de las órdenes de `municipios.md` y `calculo_de_plazos.md`, para que el sondeo diga cuál aplica H41.1.

**Why this priority**: la #245 no se fusionó por `plazos-y-territorio` (ADR 0052); sin las variantes, H41.1 no tiene
qué aplicar.

**Independent Test**: un test de `make ci` compara `skills/plazos-y-territorio` con `plazos-y-territorio-control` byte
a byte y cada variante con `plazos-y-territorio-control` fuera de lo que declara.

**Acceptance Scenarios**:

1. **Dado** el árbol del hito, **Cuando** corre `make ci`, **Entonces** cada fichero de `skills/plazos-y-territorio` es
   idéntico byte a byte al de `plazos-y-territorio-control`, y no sobra ni falta ninguno (FR-010, SC-005).
2. **Dado** cada una de las tres variantes, **Cuando** corre `make ci`, **Entonces** difiere de
   `plazos-y-territorio-control` solo en las órdenes de `municipios.md` y `calculo_de_plazos.md` que declara; su línea
   de `jerarquia_normativa.md`, la del paquete, «Si no lo dice, **pregúntalo** y espera» en el paso 1, el resto del
   cuerpo y sus referencias son los de control (FR-023, SC-009).
3. **Dado** el test de las remisiones (FR-001), **Cuando** corre `make ci`, **Entonces** no recorre
   `plazos-y-territorio`, y el test declara que queda fuera hasta H41.1 (FR-010).
4. **Dado** el árbol del hito, **Cuando** se construye el binario, se instalan las skills con `kitlegal skills install`
   o corre el job de evals, **Entonces** ninguna variante versionada del laboratorio está en el binario, en la
   instalación ni entre las evals medidas (FR-025, SC-009).

---

### User Story 4 - El arnés mide `control`, `producto` y las variantes de las tres familias (Priority: P1)

La persona lanza con `make evals-laboratorio` las variantes de una familia —`normativa`, `plazos-y-territorio` o
`jurisprudencia`—, cada una con su `control` congelada y su `producto` instalada desde `skills/`, y en
`plazos-y-territorio` también las tres variantes, con cuatro evals de la skill y sus lecturas grabadas.

**Why this priority**: es lo que hace posible la medida de la aceptación y la de H41.1.

**Independent Test**: en `make ci`, con un sustituto de `claude` sin modelo, una petición válida de cada familia abre
n/4 sesiones por eval en cada variante y modo pedidos, cada una con la skill de su variante, y una inválida no abre
ninguna.

**Acceptance Scenarios**:

1. **Dado** dos peticiones de `normativa`, una de `normativa-control` y `normativa-producto` en los dos modos y otra de
   `normativa-remisiones-en-orden` en modo herramienta, con 40 sesiones por variante y modo, **Cuando** la persona lanza
   el arnés con cada una, **Entonces** abre diez sesiones de cada una de las evals 04, 05, 15 y 22 de `normativa` por
   variante y modo, 160 en la primera y 40 en la segunda, 200 en total, cada una con solo la skill de su variante
   instalada (FR-030, FR-033).
2. **Dado** una petición de `plazos-y-territorio-control`, `plazos-y-territorio-producto` y las tres variantes, en los
   dos modos, con 40 sesiones por variante y modo, **Cuando** la persona lanza el arnés, **Entonces** abre diez sesiones
   de cada una de las evals 02, 05, 10 y 11 de `plazos-y-territorio` por variante y modo, 400 en total; y con
   `jurisprudencia-control` y `jurisprudencia-producto` en los dos modos, diez de cada una de sus evals 01, 03, 04 y 06,
   160 en total (FR-030, FR-033).
3. **Dado** una sesión de `<skill>-producto`, **Cuando** el sustituto de `claude` lista su skill instalada, **Entonces**
   sus ficheros son idénticos byte a byte a los de `skills/<skill>` del árbol en ese momento; en
   `plazos-y-territorio-producto`, también a los de `plazos-y-territorio-control` (FR-022, SC-007).
4. **Dado** una variante que no existe, una variante en un modo que no le corresponde, un número de sesiones que no es
   un múltiplo de 4 mayor que 0 o una petición que mezcla variantes de dos familias, **Cuando** la persona lanza el
   arnés, **Entonces** sale con un código distinto de 0, dice qué no vale y no abre ninguna sesión (FR-030, SC-008).

---

### User Story 5 - La tabla cuenta la lectura con `cd` al directorio de las referencias (Priority: P1)

Con las transcripciones de un sondeo, `make laboratorio-tabla` da una fila por variante y modo de la familia, y cuenta
como apertura de una referencia la orden de `Bash` que hace `cd` al directorio de las referencias de la skill y la lee
por su nombre solo, como en la #245.

**Why this priority**: sin ella, la tabla de la aceptación cuenta de menos en las celdas que deciden.

**Independent Test**: en `make ci`, sin modelo, el guion da los recuentos exactos de transcripciones fijas que leen con
`cd … && cat <f>` y que no leen.

**Acceptance Scenarios**:

1. **Dado** una sesión cuya orden de `Bash` es `cd <ruta>/normativa/references && cat normas.md`, **Cuando** el guion la
   cuenta, **Entonces** cuenta una apertura de `normas.md` en ese turno; y también con `cat "normas.md"` o
   `head -n 40 ./normas.md` tras el `cd` (FR-041, SC-006).
2. **Dado** una sesión con `cd <ruta>/normativa/references` y, en otra orden posterior, `cat normas.md`; o con
   `cd <ruta>/jurisprudencia/references && cat normas.md` en la familia de `normativa`; o con
   `cd <ruta>/normativa && cat normas.md`; o con `cat normas.md && cd <ruta>/normativa/references`; o, tras el `cd`, con
   `cat otra/normas.md`, `cat mis-normas.md` o `cat normas.md.bak`, **Cuando** el guion la cuenta, **Entonces** no cuenta
   ninguna apertura de `normas.md` (FR-041, SC-006).
3. **Dado** las suites de aceptación congeladas de H39 y H40, **Cuando** corre `make ci`, **Entonces** pasan sin cambiar
   ningún fichero suyo (FR-042, SC-009).
4. **Dado** un sondeo de `plazos-y-territorio` con sus cinco variantes, **Cuando** el guion lo cuenta, **Entonces** da
   una fila por variante y modo con las sesiones que activan la skill antes de la primera herramienta y, por cada
   referencia de la familia, las que la abren y el turno de la primera apertura, con su intervalo de Wilson al 95 %, y
   un detalle con una línea por sesión que nombra su eval (FR-040, SC-006).

---

### User Story 6 - La persona mide y decide antes de fusionar (Priority: P2)

Con la propuesta de cambio abierta, la persona lanza el sondeo de las diecinueve filas con `claude-sonnet-5-5`, deja sus
tablas, una por familia, en `evidencias/` y las añade, con su lectura, a la propuesta; solo la fusiona si `normativa` y
`jurisprudencia` cumplen la aceptación. La tabla de `plazos-y-territorio` va a un ADR que dice qué forma aplica H41.1.

**Why this priority**: es la aceptación del hito, pero la hace una persona fuera del run (ADR 0032).

**Independent Test**: el informe final del run dice que la propuesta queda pendiente del sondeo; antes de fusionar,
`evidencias/` tiene las tablas y la propuesta su lectura.

**Acceptance Scenarios**:

1. **Dado** la rama con el arnés, las dos skills reorganizadas y las tres variantes, **Cuando** la persona lanza el
   sondeo con `claude-sonnet-5-5` y 40 sesiones por variante y modo en las diecinueve filas de FR-033, **Entonces** las
   tablas de las tres familias, con diecinueve filas y 760 sesiones entre las tres, quedan en `evidencias/` con el
   modelo y la versión de Claude Code anotados (FR-040, FR-052).
2. **Dado** las tablas, **Cuando** en modo herramienta el intervalo de `normativa-producto` en `con_herramientas.md` y en
   `normas.md` no se toca con el de `normativa-control` y sí con el de `normativa-remisiones-en-orden`, y en `normativa`
   y `jurisprudencia`, en cada modo, `producto` no queda por debajo de `control` en las sesiones que activan la skill
   antes de la primera herramienta ni en ninguna referencia a cuya condición llegan las evals, **Entonces** la propuesta
   puede fusionarse (FR-052, SC-011).
3. **Dado** unas tablas en que `normativa` o `jurisprudencia` no cumplen alguna de esas condiciones, **Cuando** la
   persona las lee, **Entonces** la propuesta no se fusiona y lo medido va a un ADR que dice qué medir después (FR-052,
   SC-011).
4. **Dado** la tabla de `plazos-y-territorio`, **Cuando** la persona la lee con la misma regla frente a
   `plazos-y-territorio-control` y con el desglose por eval, **Entonces** escribe un ADR que dice qué forma aplica
   H41.1 o, si ninguna cumple, que no aplica ninguna y qué medir después; y la fila de `plazos-y-territorio-producto`
   dice cuánto se mueve el control de una medida a otra (FR-053, SC-012).

---

### Edge Cases

- **El cuerpo de una skill de `skills/` pasa de 11 000 caracteres** con las órdenes nuevas: `make ci` falla (FR-008).
- **El cuerpo de una variante de `plazos-y-territorio` pasa de 10 930 caracteres**: no falla nada, porque las variantes
  no las limita el tope; el plan da lo que movería H41.1 para que quepa (FR-024).
- **Un cambio en `laboratorio/<skill>/<skill>-control` o en `skills/plazos-y-territorio`**, a mano o por una tarea:
  `make ci` falla y nombra el fichero que difiere, falta o sobra (FR-010, FR-020).
- **`skills/<skill>` cambia después de lanzar el hito**: `<skill>-control` no cambia con ella (FR-020);
  `<skill>-producto` la sigue, porque se instala desde `skills/` en cada sondeo (FR-022); y si cambia entre el sondeo y
  la tabla, la tabla conserva las columnas de lo medido, porque las toma de la lista que el sondeo dejó (FR-032).
- **Una referencia a cuya condición no llega ninguna eval del sondeo** (en `normativa`, `buscar_la_norma.md`,
  `como_se_cita.md`, `lecturas_que_fallan.md`, `redaccion_modificada.md` y `varios_bloques.md`; en
  `plazos-y-territorio`, `jerarquia_normativa.md`): sale «0 de n» con su intervalo y no entra en la comparación de la
  aceptación (FR-040, FR-052, FR-053).
- **Sesión sin terminar o fila sin ninguna sesión terminada**: como en H40; la sesión sin terminar queda fuera de `n`
  (FR-040).
- **`DIRECTORIO` con sondeos de más de una familia**: la tabla es de la familia de la primera variante que lee en orden
  de nombre; un directorio de variante de otra familia, también de H39, es un defecto que nombra su ruta, con código 1 y
  sin tabla (FR-040).
- **Transcripción o directorio de sondeo que el guion no sabe leer** (manipulado a mano, truncado por algo que no es el
  arnés): defecto `inesperado`, código 1 (ADR 0023), sin más casos.

## Requirements *(mandatory)*

### Functional Requirements

**Las remisiones de `normativa` y `jurisprudencia`** (`skills/normativa`, `skills/jurisprudencia`)

- **FR-001**: Cada referencia de `normativa` y de `jurisprudencia` tiene en el cuerpo de su `SKILL.md` al menos una
  remisión que es una orden de leerla (ADR 0051, punto 1): en la misma frase, el verbo «lee» (o «Lee») antes del
  enlace, el enlace a `references/<f>`, cuándo leerla (una condición del caso, o ninguna) y lo que trae. La frase de un
  enlace va desde el final de la anterior —un punto seguido de espacio o de salto de línea, una línea en blanco, o la
  marca de una viñeta o de un paso— hasta el final de la suya; el punto de `.md` dentro de un enlace no la cierra.
  Ningún enlace a una referencia queda como mención dentro de otra orden, tras una condición sobre el agente, entre
  rayas o como sitio de dudas. El plan da la lista de remisiones de cada skill —frase, sitio y condición—, que parte de
  la de la #245 (`specs/033-h41-las-skills-del/contracts/remisiones.md` §2 y §5, en la rama `033-h41-las-skills-del`) y
  dice con su porqué lo que cambia respecto a ella. Un test de `make ci` sale en rojo, con la skill y la frase, si un
  enlace del cuerpo a una referencia no está en una frase con «lee» o «Lee» antes del enlace, si una referencia no
  tiene ninguna remisión o si la frase de una remisión no es la de la lista del plan.
- **FR-002**: La condición de una orden es del caso y el agente puede saber que se da sin juzgar lo que va a hacer
  (ADR 0051, punto 2, precisado por el ADR 0052, punto 2): una fecha en la pregunta, varios bloques, una orden que no
  termina con `0`, una entrada `version-obsoleta`, una norma que no está en la tabla, un documento delante, una pregunta
  por materia, una sentencia que no está delante. Se quitan las que pone el agente («si la tienes», «al presentar…»,
  «para explicar…»), lo que siempre hace falta se manda sin condición, y casos que se dan por separado no se juntan con
  «y» en una condición: cada uno va en su orden o, si llevan a la misma referencia, unidos con «o». El test de FR-001
  sale en rojo si una frase con un enlace a una referencia lleva «si la tienes».
- **FR-003**: Cada orden de leer va antes de la primera orden o llamada de la skill, o en el paso que usa la referencia:
  antes de la orden o llamada de ese paso o, si la condición la trae lo que devuelve una llamada, detrás de ella; nunca
  antes de escribir o dar la respuesta (ADR 0051, punto 3). Lo que toda respuesta necesita sigue en el cuerpo (ADR
  0046, punto 1; ADR 0050, punto 1). El test de FR-001 sale en rojo si una frase con un enlace a una referencia lleva
  «antes de escribir la respuesta» o «antes de dar la consulta».
- **FR-004**: En `normativa`, como en la #245 (§2 de su contrato): `con_herramientas.md` y `normas.md` se mandan leer
  sin condición antes de la primera orden o llamada; `buscar_la_norma.md`, con la norma que no está en la tabla de
  `normas.md`; `redaccion_en_una_fecha.md`, con una fecha en la pregunta y antes de leer el bloque;
  `redaccion_modificada.md`, por cada entrada `version-obsoleta`, detrás de la comprobación que la devuelve;
  `como_se_cita.md`, con la condición del caso que da el plan y no «antes de escribir la respuesta»; `varios_bloques.md`,
  antes de leer varios bloques a la vez; `lecturas_que_fallan.md`, si una orden no termina con `0` y antes de seguir.
  La forma de cita que toda respuesta usa sigue en el cuerpo.
- **FR-005**: En `jurisprudencia`, como en la #245 (§5 de su contrato): la orden de leer `consulta_preparada.md` va en
  el paso 2, antes de pedir `cita_preparar` o su orden, y no «antes de dar la consulta»; la de `documento_aportado.md`,
  en el paso 3, con cada documento y antes de la viñeta que pide el cotejo; y la del paso 6 (la búsqueda por materia)
  sigue siendo una orden con FR-001 a FR-003.
- **FR-006**: Para hacer sitio en `normativa`, salen del cuerpo exactamente dos textos, como en la #245 (§3 de su
  contrato): la viñeta «**Nada de otra conversación**, ni para afirmar, confirmar o desmentir: de ellas solo va la línea
  `⚠ REDACCIÓN MODIFICADA:` de cada bloque que la trae.», al final de `references/redaccion_modificada.md`; y «Si una
  lectura falla, di qué bloques faltan y por qué.», al final de `references/lecturas_que_fallan.md`. Llegan enteros y
  palabra por palabra: solo pueden cambiar los espacios y los saltos de línea. Ningún otro texto sale del cuerpo de
  `normativa` ni del de `jurisprudencia`. Un test de `make ci` sale en rojo si uno de los dos textos no está en su
  referencia o sigue en el cuerpo.
- **FR-007**: Fuera de las remisiones (FR-001 a FR-005) y de los dos movimientos (FR-006), `normativa` y
  `jurisprudencia` no cambian: ninguna regla se quita ni cambia de sentido, y una referencia de `<skill>-control` solo
  cambia por lo que se le mueve. El plan declara cada cambio del cuerpo como remisión o como movimiento, y un test de
  `make ci` sale en rojo si el cuerpo de una de las dos, deshechos los cambios declarados, difiere del de su
  `<skill>-control`, o si una referencia de `<skill>-control`, quitado lo que se le mueve, difiere de la suya.
- **FR-008**: El cuerpo de `SKILL.md` de cada skill de `skills/` —desde la línea que sigue a la que cierra el
  frontmatter— tiene 11 000 caracteres como mucho (ADR 0046, punto 2). Los tests de `make ci` que hoy lo comprueban
  salen en rojo, con el nombre de la skill y su recuento, si uno pasa de 11 000; `TestVerticalDeEnsayo` sigue en verde
  sin cambios.
- **FR-009**: El frontmatter de `normativa` y de `jurisprudencia` —también su `description`— es idéntico byte a byte al
  de su `<skill>-control`, y la tabla de comandos y las referencias generadas desde `data/` son las que da
  `make skills-sync`. `make ci` sale en rojo si el frontmatter difiere o si `make skills-sync` deja diferencias.
- **FR-010**: `skills/plazos-y-territorio` no cambia: su `SKILL.md` y sus referencias son idénticos byte a byte a los
  de `plazos-y-territorio-control` (la de `v0.9.0`), con «Si no lo dice, **pregúntalo** y espera» en el paso 1 (ADR
  0052, punto 4). Un test de `make ci` sale en rojo, y nombra el fichero, si uno difiere en un byte, falta o sobra. El
  test de las remisiones (FR-001) deja fuera `plazos-y-territorio` y lo declara en su código, con H41.1 como el hito que
  la incluye. `consulta-juridica` tampoco cambia.

**Las variantes del laboratorio**

- **FR-020**: Cada familia —`normativa`, `plazos-y-territorio` y `jurisprudencia`— tiene una variante
  `<skill>-control`: `skills/<skill>` de `v0.9.0`, congelada en `laboratorio/<skill>/`. La de `normativa` ya existe
  (`laboratorio/normativa/normativa-control`) y deja de compararse con `skills/normativa`. `make ci` sale en rojo, y
  nombra el fichero, si un fichero de una `<skill>-control` difiere en un byte de ese estado, falta o sobra.
- **FR-021**: `normativa-remisiones-en-orden`, `normativa-cuerpo-corto` y `normativa-testigo-en-referencia` siguen
  comparándose con `normativa-control`, como hasta ahora: `make ci` sale en rojo si una difiere de ella fuera de lo que
  declara.
- **FR-022**: Cada familia tiene una variante `<skill>-producto`, que es `skills/<skill>` en el momento del sondeo: el
  arnés la instala desde `skills/`, y no hay copia versionada suya en `laboratorio/` ni en ningún otro sitio. En
  `plazos-y-territorio`, por FR-010, `plazos-y-territorio-producto` es igual que `plazos-y-territorio-control` byte a
  byte, y el arnés y la tabla la tratan como a las otras dos `<skill>-producto`: es la fila de control contra control.
  Un test de `make ci` sale en rojo si la skill que ve una sesión de `<skill>-producto` difiere en un byte de
  `skills/<skill>` del árbol, o si el repositorio tiene un directorio de variante `<skill>-producto`.
- **FR-023**: `laboratorio/plazos-y-territorio/` tiene, junto a `plazos-y-territorio-control`, tres variantes (ADR 0052,
  punto 3), cada una con su nombre `plazos-y-territorio-<forma>`, que da el plan. Cada una difiere de control solo en la
  forma de las órdenes de `municipios.md` y `calculo_de_plazos.md`; la línea de `jerarquia_normativa.md`, la del
  paquete, «Si no lo dice, **pregúntalo** y espera» en el paso 1, el resto del cuerpo, el frontmatter y las referencias
  quedan como en control:
  - **sin condición**: las líneas de `municipios.md` y `calculo_de_plazos.md` salen de la lista de «Referencias»; «lee
    [municipios.md]…», con lo que trae, va antes de la primera orden o llamada, como `con_herramientas.md` en
    `normativa`, y «lee [calculo_de_plazos.md]…», con lo que trae, antes del cálculo;
  - **una condición del caso por orden**: las mismas dos líneas salen de la lista; «si falta el municipio o
    `territorio resolver` no sale con `0`, lee [municipios.md]…», «si alguno de los municipios cuenta, lee
    [calculo_de_plazos.md]…», y una orden aparte de leer `calculo_de_plazos.md` por cada uno de sus otros casos —un
    plazo sin calcular; una lectura que sale con 4 o 5; una lectura con `redaccion_posterior`—, cada una con lo que trae
    y en el paso cuya orden o llamada trae su condición, detrás de ella. «Al presentar un territorio» y «para explicar
    su cobertura» no se escriben;
  - **la lista de control con el verbo delante del enlace**: las dos líneas siguen en la lista, en su sitio y con sus
    condiciones de control, con «Lee » delante del enlace: «- Lee [municipios.md](references/municipios.md): al
    presentar un territorio…», y lo mismo `calculo_de_plazos.md`; nada más cambia.

  El plan da el texto exacto de cada variante, como diferencia frente a control. `make ci` sale en rojo si una variante
  difiere de `plazos-y-territorio-control` fuera de lo que declara, o si lo que declara toca otra cosa que las órdenes
  de `municipios.md` y `calculo_de_plazos.md`.
- **FR-024**: El tope de 11 000 caracteres del cuerpo no se aplica a las variantes versionadas del laboratorio, como a
  `normativa-remisiones-en-orden` en H40: rige `skills/`. El plan da el número de caracteres del cuerpo de cada
  variante de `plazos-y-territorio` y, para la que pase de 10 930, lo que movería H41.1 para que quepa, con la regla
  del ADR 0052, punto 4: solo a una referencia a cuya condición llega alguna eval del sondeo (como la frase del paso 10
  «Si sale con 4 o 5, **`⚠ PLAZO:` se mantiene**…», de 105 caracteres, a `calculo_de_plazos.md`, a la que llega la eval
  11), y ninguna regla que el cuerpo da como orden sin una eval que llegue a su caso. El run no aplica ese movimiento.
  Los 10 930 caracteres no son aquí un límite, sino lo que decide si el plan da un movimiento: ninguna comprobación
  sale en rojo con una variante por encima, y que el plan dé el recuento y el movimiento lo juzga el juez del plan.
- **FR-025**: Ninguna variante versionada (`<skill>-control`, las de H40 y las tres de `plazos-y-territorio`) se
  empotra en el binario, la pone `kitlegal skills install` ni la elige o mide el job de evals; `make ci` sale en rojo si
  una de las tres cosas deja de cumplirse.

**El arnés del sondeo**

- **FR-030**: `make evals-laboratorio` lanza las variantes de las familias `normativa`, `plazos-y-territorio` y
  `jurisprudencia`, una familia por invocación, además de la de H39 como hoy. `<skill>-control`, `<skill>-producto` y
  las tres variantes de `plazos-y-territorio` se miden en los dos modos; `normativa-remisiones-en-orden`,
  `normativa-cuerpo-corto` y `normativa-testigo-en-referencia`, solo en modo herramienta, como en H40. Por variante y
  modo pedidos, abre el número de sesiones que pide la persona, repartidas por igual entre las cuatro evals de la
  familia, las de la #245 (research D12): 04, 05, 15 y 22 de `normativa`; 02, 05, 10 y 11 de `plazos-y-territorio`; 01,
  03, 04 y 06 de `jurisprudencia`. Las sirven sus lecturas grabadas, sin red a las fuentes (ADR 0043); cada sesión tiene
  instalada solo la skill de su variante, con el nombre de la skill, y ninguna otra. Las referencias a cuya condición
  llegan esas evals son las de la research D12 de la #245: en `normativa`, `con_herramientas.md`, `normas.md` y
  `redaccion_en_una_fecha.md`; en `plazos-y-territorio`, `administrativo.md`, `municipios.md` y `calculo_de_plazos.md`;
  en `jurisprudencia`, `consulta_preparada.md` y `documento_aportado.md`; el research del plan lo confirma con las
  remisiones del hito. Pedir una variante que no existe, una variante en un modo que no le corresponde, un número de
  sesiones que no es un múltiplo de 4 mayor que 0 o variantes de dos familias es un error de argumentos: sale con un
  código distinto de 0, dice qué no vale, uno por línea y con el nombre del argumento de `make`, y no abre ninguna
  sesión. Los demás argumentos (`MODOS`, `MODELO`, `CLAUDE_CODE`, `CONCURRENCIA`, `DIRECTORIO`) y la versión de Claude
  Code se comprueban como en H40 (`specs/032-h40-por-qu-e/contracts/arnes.md` §2).
- **FR-031**: El arnés lo lanza una persona, con la credencial del sondeo (ADR 0032); ni `make ci`, ni un gancho, ni el
  job de evals, ni un paso del workflow lo ejecutan. El run no graba nada: no añade ni cambia ninguna grabación de
  `testdata/` y no escribe en `evidencias/`.
- **FR-032**: Al instalar `<skill>-producto` desde `skills/`, el arnés escribe en el directorio del sondeo los nombres de
  los ficheros de `references/` de la skill que instaló, en orden de nombre; solo los nombres, ni la copia de la skill
  ni una huella de su contenido. La forma del fichero la fija el plan, sin cambiar lo que los guiones de H39 y H40 leen
  de `sondeo.json`, para que sus suites congeladas sigan en verde (FR-042). `make ci` sale en rojo si el arnés no deja
  la lista o si la tabla no toma de ella las columnas de `<skill>-producto`.
- **FR-033**: Las filas del sondeo de la aceptación son diecinueve, de 40 sesiones cada una, 760 en total:
  `<skill>-control` y `<skill>-producto` de cada familia en los dos modos (doce), las tres variantes de
  `plazos-y-territorio` en los dos modos (seis) y `normativa-remisiones-en-orden` en modo herramienta (una). Como
  `MODOS` vale para todas las variantes de una invocación (FR-030), el arnés las acepta en cuatro invocaciones, como en
  H40 y en la #245: dos de `normativa`, una de `normativa-control,normativa-producto` con `MODOS=orden,herramienta`
  (cuatro filas, 160 sesiones) y otra de `normativa-remisiones-en-orden` con `MODOS=herramienta` (una fila, 40); una de
  `plazos-y-territorio` (diez filas, 400) y una de `jurisprudencia` (cuatro, 160). La tabla de `normativa` cuenta el
  directorio que junta las dos de esa familia (FR-040).

**La tabla**

- **FR-040**: `make laboratorio-tabla` cuenta los sondeos de las tres familias con las reglas de la tabla de H40
  (`specs/032-h40-por-qu-e/contracts/tabla.md` §2 y §3), con la skill y las referencias de la familia: una fila por
  variante y modo, con el modelo y la versión de Claude Code; las sesiones que activan la skill antes de la primera
  herramienta de kitlegal; por cada referencia, las sesiones que la abren y el turno de su primera apertura; y las
  sesiones sin terminar, fuera de `n`. Cada recuento con su intervalo de Wilson al 95 %. Las referencias de la tabla son
  todas las de las variantes del sondeo, las de `<skill>-producto` tal como se instalaron en él, en orden de nombre; en
  la fila de una variante que no tiene una referencia, su celda va vacía. Las de `<skill>-producto` salen de la lista
  que el arnés deja en el directorio del sondeo (FR-032); las de las variantes versionadas, de su declaración. Su
  detalle da una línea por sesión, cuyo nombre lleva la variante, el modo y la eval, como en la #245: de ahí sale el
  desglose por eval que lee la persona. Una tabla cuenta una sola familia, la de la primera variante que lee en orden de
  nombre (`specs/032-h40-por-qu-e/contracts/tabla.md` §1): un directorio de variante de otra familia, también de H39,
  es un defecto que nombra su ruta, con código 1 y sin tabla. La columna «Cumplen la regla» solo va en la tabla de
  `normativa`, como en H40. `make ci` comprueba el guion con transcripciones fijas de cada familia, también de las tres
  variantes de `plazos-y-territorio`, y sale en rojo si un recuento, un turno o un extremo de un intervalo difiere.
- **FR-041**: Una orden de `Bash` también abre la referencia `<f>` de la skill de la familia si hace `cd` a un directorio
  cuya ruta termina en `<skill>/references` (con o sin barra final) y después, en la misma orden, nombra `<f>` por su
  nombre solo (research D13 de la #245). «Nombra `<f>` por su nombre solo» es: una orden simple posterior de la misma
  `input.command` tiene una palabra que es exactamente `<f>` o `./<f>`. Las palabras se parten como en sh (comillas
  simples y dobles, barra invertida) y las órdenes simples por `|`, `||`, `&&`, `;`, `&` y saltos de línea, como en
  `specs/031-h39-c-omo-se/contracts/tabla.md` §2. Cuentan `cat normas.md`, `cat "normas.md"`, `cat 'normas.md'` y
  `head -n 40 ./normas.md`; no cuentan `cat otra/normas.md`, `cat /x/normas.md`, `cat mis-normas.md` ni
  `cat normas.md.bak`. No la abre si el `cd` va a otro directorio (el de la skill, el de las referencias de otra skill),
  si `<f>` va antes del `cd` o si el `cd` y el nombre van en órdenes distintas. La regla de H40 que cuenta una orden que
  contiene `references/<f>` se aplica aparte, y lo demás de la regla de H40 no cambia. La regla ampliada rige solo en
  las familias de las skills del producto; la de H39 (`articulos-del-boe`) se cuenta exactamente con su contrato
  (`specs/031-h39-c-omo-se/contracts/tabla.md` §2). `make ci` lo comprueba con sesiones fijas con las formas que cuentan
  y con las que no, y sale en rojo si un recuento difiere.
- **FR-042**: Las suites de aceptación congeladas de H39 (`specs/031-h39-c-omo-se/aceptacion/`) y de H40
  (`specs/032-h40-por-qu-e/aceptacion/`) siguen en verde en `make ci` sin cambiar ningún fichero suyo.

**La documentación, el cierre y la aceptación de la persona**

- **FR-050**: `CHANGELOG.md` (*Unreleased*) da el cambio de `normativa` y `jurisprudencia`. `docs/SKILLS.md` §1 («Cómo
  se remite») recoge lo que da la reorganización de esas dos skills y la guía no recoja; el punto 2 del ADR 0052 no
  entra en el run: entra cuando lo mida el sondeo de este hito, en el ADR que lo lea.
- **FR-051**: Las evals de las skills del producto salen en verde en la medición del cierre, en los dos modos.
- **FR-052**: Antes de fusionar, la persona lanza el sondeo de FR-033 con `claude-sonnet-5-5` y 40 sesiones por
  variante y modo y deja en `evidencias/` sus tablas, una por familia (FR-040). El hito vale si cumple todo esto: en
  modo herramienta, `normativa-producto` abre `con_herramientas.md` y `normas.md` con un intervalo que no se toca con el
  de `normativa-control` y sí con el de `normativa-remisiones-en-orden` (ADR 0051, punto 4); y en `normativa` y
  `jurisprudencia`, en cada modo, `producto` no queda por debajo de `control` en las sesiones que activan la skill antes
  de la primera herramienta, ni en ninguna referencia a cuya condición llegan las evals (FR-030). «Por debajo» quiere
  decir con un intervalo que no se toca con el de `control`. El run no lanza el sondeo; las tablas y su lectura se
  añaden a la propuesta de cambio antes de fusionarla, y la propuesta no se fusiona sin ellas. Si `normativa` o
  `jurisprudencia` no cumplen, la propuesta no se fusiona, y lo que se mida va a un ADR que dice qué medir después. El
  informe final del run lo dice.
- **FR-053**: Las variantes de `plazos-y-territorio` no deciden la fusión: su tabla, con la misma regla frente a
  `plazos-y-territorio-control` (ADR 0052, punto 3) y el desglose por eval, va a un ADR que dice qué forma aplica
  H41.1 o, si ninguna cumple, que no aplica ninguna y qué medir después. La fila de `plazos-y-territorio-producto` dice
  cuánto se mueve el control de una medida a otra. El ADR lo escribe la persona tras el sondeo, fuera del run.

### Key Entities

- **Remisión**: la frase del cuerpo de un `SKILL.md` que enlaza una referencia; con su verbo, su condición, su sitio y
  lo que trae (FR-001 a FR-005).
- **Movimiento**: uno de los dos textos que salen del cuerpo de `normativa` a una referencia (FR-006).
- **Familia del laboratorio**: las variantes de una skill del producto, con sus cuatro evals del sondeo y sus
  referencias (FR-030).
- **`<skill>-control`**: la skill de `v0.9.0`, congelada en `laboratorio/<skill>/` (FR-020).
- **`<skill>-producto`**: `skills/<skill>` en el momento del sondeo, sin copia versionada (FR-022).
- **Variante de forma de `plazos-y-territorio`**: una de las tres de FR-023, que difiere de control solo en la forma
  de dos órdenes.
- **Sondeo y tabla**: las sesiones de las variantes por modo, y su recuento por fila y por sesión (FR-030 a FR-041).

### Uso, de fuera adentro

- **El cuerpo de `normativa` y `jurisprudencia`** (FR-001 a FR-008). Lo lee el agente una vez por pregunta, al activar
  la skill; tiene 11 000 caracteres como mucho (FR-008) y no crece con el uso: cientos de normas y miles de bloques
  consultados no lo tocan.
- **Las referencias que se mandan leer** (FR-001 a FR-006). El agente lee cada una una vez por pregunta, cuando llega su
  condición. En `normativa`, `con_herramientas.md` (2 442 bytes hoy) y `normas.md` (3 510) en toda pregunta, unos
  5,9 KB que hoy, en modo herramienta, casi nunca se leen —lo que gana la skill es justo esto—; las demás (909 a 2 809
  bytes; `redaccion_modificada.md` y `lecturas_que_fallan.md` crecen unos 160 y 60 bytes con lo movido) solo con su
  condición: por pregunta, el cuerpo más unos 5,9 KB, y hasta otros 2,9 KB por cada condición que llega. En
  `jurisprudencia`, `consulta_preparada.md` (2 038 bytes) con cada sentencia que no está delante y
  `documento_aportado.md` (2 805) con cada documento, una vez por pregunta aunque lleguen varios. `normas.md` crece con
  las normas declaradas en `data/normas.yaml` al añadir un paquete, no con las consultadas; las demás, solo cuando un
  hito las cambia. Lo movido (FR-006) se lee solo en el caso que lo usa, en vez de en toda pregunta con el cuerpo. No
  llevan señales que caduquen: su contenido no depende de la consulta.
- **`plazos-y-territorio` y sus variantes** (FR-010, FR-023). La skill del producto no cambia: lo que lee el agente por
  pregunta es lo de hoy. Las variantes solo las ve una sesión del sondeo, una vez por sesión; no se instalan (FR-025).
- **La lista de referencias del sondeo** (FR-032). La escribe el arnés una vez por sondeo de `<skill>-producto` y la lee
  la tabla una vez: de dos a ocho nombres, menos de 300 bytes. No crece con el uso del producto.
- **Las tablas del sondeo** (FR-040, FR-041). Las lee la persona una vez, para decidir la fusión y escribir el ADR de
  FR-053: diecinueve filas entre las tres familias —cinco en `normativa`, diez en `plazos-y-territorio`, cuatro en
  `jurisprudencia`—, cada una con la activación y una celda por referencia (ocho en `normativa`, cuatro en
  `plazos-y-territorio`, dos en `jurisprudencia`), más el detalle de 760 sesiones, una línea de 95 a 140 bytes por
  sesión: unos 28, 45 y 16 KB por familia, a partir de las de `evidencias/sondeos-20261106/` (28,6, 20,5 y 15,6 KB con
  cinco, cuatro y cuatro filas). Su tamaño lo fijan las variantes, los modos y las sesiones pedidas, no el uso del
  producto; dejan de leerse cuando la persona decide y escribe el ADR, y un sondeo nuevo da otras tablas.

## Success Criteria *(mandatory)*

### Measurable Outcomes

- **SC-001**: El 100 % de las referencias de `normativa` y `jurisprudencia` con al menos una remisión que es una orden;
  0 frases con un enlace a una referencia sin «lee» o «Lee» antes del enlace o con «si la tienes», «antes de escribir la
  respuesta» o «antes de dar la consulta»; y 0 remisiones cuya frase no es la de la lista del plan. Control: el test de
  FR-001 en `make ci` sale en rojo con la primera.
- **SC-002**: 0 skills de `skills/` con un cuerpo de más de 11 000 caracteres. Control: los tests de `make ci` de
  FR-008, que salen en rojo con la skill y su recuento.
- **SC-003**: 0 bytes de diferencia entre el frontmatter de `normativa` y de `jurisprudencia` y el de su
  `<skill>-control`, y 0 diferencias tras `make skills-sync`. Control: los tests de FR-009 en `make ci`.
- **SC-004**: Los dos textos de FR-006 en su referencia y fuera del cuerpo, y 0 diferencias de `normativa` y
  `jurisprudencia` con su `<skill>-control`, en el cuerpo y en las referencias, fuera de lo declarado. Control: los
  tests de FR-006 y FR-007 en `make ci` salen en rojo con el primer texto o la primera diferencia.
- **SC-005**: 0 bytes de diferencia entre `skills/plazos-y-territorio` y `plazos-y-territorio-control`, y 0 ficheros que
  falten o sobren. Control: el test de FR-010 en `make ci`, que sale en rojo y nombra el fichero.
- **SC-006**: Los recuentos, turnos y extremos de Wilson de las transcripciones fijas de las tres familias, también de
  las variantes de `plazos-y-territorio`, exactos, con las lecturas por `cd` y las que no lo son. Control: los tests de
  FR-040 y FR-041 en `make ci` salen en rojo con el primer valor que difiere.
- **SC-007**: 0 bytes de diferencia entre la skill de una sesión de `<skill>-producto` y `skills/<skill>`, y 0
  directorios de variante `<skill>-producto` en el repositorio. Control: el test de FR-022 en `make ci`.
- **SC-008**: Con una petición inválida (FR-030), 0 sesiones abiertas; con una válida, n/4 por eval en cada variante y
  modo pedidos de cada familia, también en las cuatro invocaciones de FR-033. Control: un test de `make ci` con un
  sustituto de `claude` que cuenta sus invocaciones, sin modelo, sale en rojo si una inválida abre alguna o una válida
  no abre n/4 por eval.
- **SC-009**: 0 bytes de diferencia en cada `<skill>-control` respecto de `skills/<skill>` de `v0.9.0`; 0 bytes fuera de
  lo declarado entre cada variante de H40 y `normativa-control` y entre cada variante de FR-023 y
  `plazos-y-territorio-control`; 0 ficheros de variantes versionadas en el binario, en la instalación y en las evals
  del job; las suites congeladas de H39 y H40 en verde. Control: los tests de FR-020, FR-021, FR-023, FR-025 y FR-042 en
  `make ci`.
- **SC-010**: Las evals de las skills del producto en verde en la medición del cierre, en los dos modos. Control: el job
  de evals del cierre sale en rojo si una no pasa su umbral en un modo, y el informe final da sus cifras (FR-051).
- **SC-011**: Antes de fusionar, las tablas de las tres familias, con diecinueve filas de 40 sesiones entre las tres,
  están en `evidencias/`, y `normativa` y `jurisprudencia` cumplen las condiciones de FR-052. Lo mide la persona (capa
  3): el sondeo abre sesiones con modelo, que ni `make ci` ni el run abren (ADR 0032), y ningún control automático lo
  hace cumplir; la propuesta no se fusiona sin las tablas y su lectura.
- **SC-012**: Tras el sondeo, un ADR lee la tabla de `plazos-y-territorio` con la regla de FR-053 y dice qué forma
  aplica H41.1, o que ninguna y qué medir después. Lo escribe la persona fuera del run (capa 3); no decide la fusión.

## Criterios del hito (literales)

**Controles**:

- `make ci` comprueba que cada `<skill>-control` no cambia y que `skills/plazos-y-territorio` es igual que
  `plazos-y-territorio-control` (FR-010, FR-020, SC-005, SC-009);
- que cada variante de H40 difiere de `normativa-control`, y cada variante de `plazos-y-territorio` de
  `plazos-y-territorio-control`, solo en lo que declara (FR-021, FR-023, SC-009);
- que ninguna variante se empotra ni se instala (FR-025, SC-009);
- la frase de cada remisión de `normativa` y `jurisprudencia` (`TestRemisionesDelRepositorio` de la #245), con
  `plazos-y-territorio` fuera hasta H41.1, declarado en el test (FR-001, FR-010, SC-001);
- el cuerpo de cada skill del producto, con 11 000 caracteres como mucho (FR-008, SC-002);
- la regla de lectura ampliada, con sesiones grabadas que leen con `cd … && cat <f>` y que no leen (FR-041, SC-006).

Las evals de las skills del producto salen en verde en la medición del cierre, en los dos modos (FR-051, SC-010).

**Aceptación**: el sondeo de la persona, con `claude-sonnet-5-5` y 40 sesiones por variante y modo, deja una tabla por
familia en `evidencias/`, y el hito vale si cumple todo esto (FR-052, SC-011):

- en modo herramienta, `normativa-producto` abre `con_herramientas.md` y `normas.md` con un intervalo que no se toca
  con el de `normativa-control` y sí con el de `normativa-remisiones-en-orden` (ADR 0051, punto 4);
- en `normativa` y `jurisprudencia`, en cada modo, `producto` no queda por debajo de `control` en las sesiones que
  activan la skill antes de la primera herramienta, ni en ninguna referencia a cuya condición llegan las evals. «Por
  debajo» quiere decir con un intervalo que no se toca con el de `control`.

Las variantes de `plazos-y-territorio` no deciden la fusión: su tabla, con la misma regla frente a
`plazos-y-territorio-control` (ADR 0052, punto 3) y el desglose por eval, va a un ADR que dice qué forma aplica H41.1,
o, si ninguna cumple, que no aplica ninguna y qué medir después. La fila de `plazos-y-territorio-producto` dice cuánto
se mueve el control de una medida a otra (FR-053, SC-012).

El run no lanza el sondeo. Las tablas y su lectura se añaden a la propuesta de cambio antes de fusionarla, y la
propuesta no se fusiona sin ellas. Si `normativa` o `jurisprudencia` no cumplen, la propuesta no se fusiona, y lo que se
mida va a un ADR que dice qué medir después.

## Assumptions

- Las cifras de partida (H40 y la #245; `evidencias/sondeos-20261104/` y `evidencias/sondeos-20261106/`) son las que
  citan el hito y los ADR 0051 y 0052; esta sesión no las ha medido.
- «Como en la #245» se lee contra lo que conserva la rama `033-h41-las-skills-del` (spec, plan, contratos, research y
  código), que el hito cita y esta sesión ha leído con `git show`; lo que el plan cambie respecto a ella lo dice con su
  porqué (hito, Entrega).
- La lista de las referencias a cuya condición llegan las evals (FR-030) es la de la research D12 de la #245, hecha con
  sus remisiones; en `normativa` y `jurisprudencia` las de este hito parten de las mismas, y en `plazos-y-territorio`
  las variantes no cambian ninguna condición de control salvo la forma, así que la lista vale para comparar, y el
  research del plan la confirma.
- El desglose por eval que pide la aceptación de `plazos-y-territorio` lo da el detalle de la tabla, una línea por
  sesión con su eval en el nombre, como en la #245 y en el `recuento.md` del ADR 0052; la tabla no añade otra (lectura
  conservadora: el hito pide el arnés «como en la #245» y no una columna nueva).
- `skills/` en `main` al lanzar el hito es la de `v0.9.0` (comprobado), así que `<skill>-control` es las dos cosas a la
  vez, y `plazos-y-territorio-producto` es igual que `plazos-y-territorio-control` mientras FR-010 se cumpla.
- Las evals del sondeo ya existen y ya tienen sus lecturas grabadas; el run no graba ninguna (FR-031). `jurisprudencia`
  no consulta el CENDOJ: sus evals no leen fuentes por red (ADR 0036).
- En cada sesión del sondeo solo está instalada la skill de su variante, como en H40 y en la #245.
- El modelo se pasa como identificador completo porque un alias lo resuelve el cliente (ADR 0031).

## Fuera de alcance

- Cambiar `skills/plazos-y-territorio`, su `SKILL.md` o sus referencias, aplicarle la forma que gane o mover texto de
  su cuerpo para que quepa: es H41.1, con la forma que gane.
- Cambiar ningún umbral, límite ni tope, tampoco los 11 000 caracteres del cuerpo (ADR 0046, punto 2), los 900 de la
  descripción ni los umbrales del job de evals.
- Cambiar la descripción de una skill: la activación ya se mide en 40 de 40 en los dos modos (ADR 0050 y 0051).
- Cambiar el generador de la tabla de comandos o de las referencias generadas desde `data/`, que es común a todas las
  skills, incluida la orden generada de `administrativo.md` en `plazos-y-territorio`.
- `consulta-juridica`, que no tiene referencias.
- Evals nuevas, y cambiar las que ya existen (sus preguntas o expectativas): «pregúntalo y espera» no sale del cuerpo, y
  no hace falta la eval que deja el municipio sin decir (ADR 0052, punto 4). Tampoco otras evals en el sondeo que las
  doce de FR-030.
- Medir con el juez en el sondeo, que cuenta lecturas y activación, sin modelo.
- Guardar las sesiones del job de evals como artefacto, que sigue en `backlog/medida-de-las-skills.md`.
- Llevar el punto 2 del ADR 0052 a `docs/SKILLS.md` en el run: entra cuando lo mida el sondeo, en el ADR que lo lea.
- Mover otros textos del cuerpo de `normativa` o `jurisprudencia` que los dos de FR-006.
- Cambiar las instrucciones del servidor `kitlegal mcp serve` o las descripciones de sus herramientas, que en modo
  herramienta comparten todas las variantes porque el arnés arranca el `kitlegal` del árbol
  (`specs/032-h40-por-qu-e/contracts/arnes.md` §3).
- Medir en este sondeo `normativa-cuerpo-corto` o `normativa-testigo-en-referencia`, `normativa-remisiones-en-orden` en
  modo orden, otros modelos que `claude-sonnet-5-5` u otros hosts que Claude Code.
- Instalar en las sesiones del sondeo otras skills que la de la variante de la fila.
- Que la tabla calcule si dos intervalos se tocan, dé un veredicto de aceptación o un desglose por eval agregado: la
  lectura la hace la persona, con el detalle por sesión.
- Anotar en la tabla una huella de la skill `<skill>-producto` medida, o guardar en el directorio del sondeo una copia
  de la skill instalada: solo la lista de nombres de sus referencias (FR-032).
- Ampliar la regla de lectura con `cd` (FR-041) a la tabla de la familia del laboratorio de H39 (`articulos-del-boe`).
- Columnas de la tabla que FR-040 no enumera.
- Que el job de evals mida las variantes, o que `make ci` abra sesiones con modelo.
- Grabar lecturas nuevas o escribir en `evidencias/` desde el run, lanzar el sondeo o escribir dentro del run el ADR de
  FR-052 o el de FR-053.
