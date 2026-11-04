# Feature Specification: H40 · Por qué `normativa` no lee sus referencias en modo herramienta

**Feature Branch**: `032-h40-por-qu-e`

**Created**: 2026-11-04

**Status**: Draft

**Input**: el fichero del hito, `docs/roadmap/in-progress/H40-por-que-normativa-no-lee-sus-referencias.md` (entrada de
la persona, ADR 0018).

## Contexto

Lo que el hito quiere saber, medido, es cuál de las tres diferencias entre `normativa` y la skill de laboratorio de H39
hace que, en **modo herramienta** (el agente llama a `<applet>_<verbo>` del servidor `kitlegal mcp serve`), `normativa`
no abra sus referencias, antes de reorganizar las skills del producto (ADR 0050, punto 5). Sin la causa, la
reorganización movería texto sin saber si se va a leer, y el cuerpo de `normativa` no tiene sitio para otro arreglo.

Evidencia de partida, tal como la dan el hito y el ADR 0050 (no medida en esta sesión): en los sondeos de la #214
(evals 05 y 15 de `normativa`; `specs/030-h32-arquitectura-de-skills/research.md` D26 a D30), en modo herramienta
`normas.md` se abre en 8 a 11 de 80 sesiones, frente a 77 a 79 en **modo orden** (el agente ejecuta
`kitlegal <applet> <verbo>`), y `con_herramientas.md` en 2 a 4 de 80. En el sondeo de H39
(`evidencias/sondeos-20261103/`), la skill de laboratorio abre su referencia en 39 o 40 de 40 en modo herramienta. Las
tres diferencias sin medir:

- **el tamaño**: 12 214 caracteres y ocho referencias, frente a 1776 y una;
- **las remisiones condicionales** («si no está», «con ella») frente a una orden;
- **lo que trae la referencia**: lo que el agente cree saber, como la tabla de identificadores del BOE, frente a una
  línea que solo está en la referencia. El sondeo de la eval 15 (`evidencias/sondeos-20261025/`, `brazo-D`) ya
  apuntaba a esta: 0 de 40 sesiones abrieron una referencia de guía que el cuerpo mandaba leer.

Comprobado en esta sesión, en el repositorio: el cuerpo de `skills/normativa/SKILL.md` (desde la línea que sigue al
frontmatter) tiene 10 929 caracteres; la skill tiene ocho referencias, y su cuerpo ocho remisiones, una a cada una. Con
los textos que da el hito, el cuerpo de `remisiones-en-orden` gana unos 150 caracteres (contados en esta sesión sobre
los fragmentos que cambian, en parte con `wc -m` y en parte a mano, sin la variante escrita) y pasa de 11 000; el de
`testigo-en-referencia` gana 44 y queda por debajo.

## Clarifications

### Session 2026-11-04

- Q: En cada sesión del sondeo, ¿qué skills tiene instaladas el agente además de la variante de su fila, que lleva el nombre `normativa`? (FR-008) → A: Solo la variante de su fila, con el nombre `normativa`, y ninguna otra skill, como en el laboratorio de H39 (auto: criterio a; fuente: hito H40 «Objetivo» y «Entrega»; ADR 0050 «Lo medido» y «Decisión» 5; H39 FR-030 y contracts/arnes.md §3.1)
- Q: ¿Puede una misma invocación de `make evals-laboratorio` pedir a la vez variantes del laboratorio de H39 y variantes de `normativa`, y qué hace el arnés si las pide? (FR-010) → A: No: mezclarlas es un error de argumentos; sale con un código distinto de 0, dice qué no vale y no abre ninguna sesión (auto: conservadora; criterio d; fuente: hito H40 «Entrega» y «Fuera»; FR-010, FR-020 y FR-023; H39 contracts/arnes.md §2; constitución, «Criterio de decisión autónoma» 2 y 4)

## User Scenarios & Testing *(mandatory)*

### User Story 1 - Cuatro variantes de `normativa`, cada una con una sola diferencia (Priority: P1)

Para saber cuál de las tres diferencias explica el defecto, hacen falta cuatro skills que sean `normativa` salvo en una
cosa cada una, fuera del producto: la de control, idéntica a la de `skills/`, y una por diferencia, con el texto exacto
que fija el hito.

**Why this priority**: es el objeto de la medida; sin ellas el sondeo no tiene qué comparar.

**Independent Test**: `make ci` compara la variante `control` con `skills/normativa` byte a byte, compara cada una de
las otras con la de control y comprueba que ninguna llega al binario, a la instalación ni al job de evals.

**Acceptance Scenarios**:

1. **Dado** el árbol del hito, **Cuando** corre `make ci`, **Entonces** la variante `control` es idéntica byte a byte a
   `skills/normativa`, fichero a fichero, sin ninguno de más ni de menos (FR-002, SC-001).
2. **Dado** un cambio en `skills/normativa` que no se lleva a la variante `control`, **Cuando** corre `make ci`,
   **Entonces** falla y nombra el fichero que difiere (FR-002, SC-001).
3. **Dado** cada una de `remisiones-en-orden`, `cuerpo-corto` y `testigo-en-referencia`, **Cuando** `make ci` la compara
   con la de control, **Entonces** todo lo que no está en su declaración es idéntico byte a byte y lo declarado es
   exactamente el texto que fija el hito (FR-003 a FR-006, SC-002).
4. **Dado** el binario construido del árbol, **Cuando** se listan las skills que lleva dentro, se ejecuta
   `kitlegal skills install` en un directorio vacío o se piden las evals que elige el job, **Entonces** no aparece
   ningún fichero de ninguna variante (FR-001, SC-003).

---

### User Story 2 - El arnés de H39 lanza las variantes con las evals reales de `normativa` (Priority: P1)

La persona lanza con `make evals-laboratorio` las variantes de `normativa` que pide, en los modos que les corresponden,
con las evals 04, 05, 15 y 22 de `normativa` y sus lecturas grabadas, con el modelo y la versión de Claude Code que
elige.

**Why this priority**: es lo que hace posible la medida de la aceptación; sin ello el sondeo no se puede lanzar.

**Independent Test**: en `make ci`, con un sustituto de `claude` sin modelo, una petición válida abre n/4 sesiones por
eval en cada variante y modo pedidos, y una petición inválida no abre ninguna.

**Acceptance Scenarios**:

1. **Dado** una petición de `control` en los dos modos y de las otras tres en modo herramienta, con 40 sesiones por
   variante y modo, **Cuando** la persona lanza el arnés, **Entonces** abre diez sesiones de cada una de las evals 04,
   05, 15 y 22 por variante y modo, 200 en total, cada una con la variante de su fila (FR-010).
2. **Dado** una petición de `remisiones-en-orden`, `cuerpo-corto` o `testigo-en-referencia` en modo orden, o un número
   de sesiones que no es múltiplo de 4, o una variante que no existe, **Cuando** la persona lanza el arnés, **Entonces**
   sale con un código distinto de 0, dice qué no vale y no abre ninguna sesión (FR-010, SC-006).
3. **Dado** que la versión de Claude Code instalada no es la pedida, o que el modelo no es un identificador completo,
   **Cuando** la persona lanza el arnés, **Entonces** sale con un código distinto de 0, nombra lo que no coincide y no
   abre ninguna sesión (FR-011, SC-006).

---

### User Story 3 - La tabla dice, por variante y modo, qué se activa, qué se abre y qué se cumple (Priority: P1)

Con las transcripciones del sondeo, `make laboratorio-tabla` da una fila por variante y modo con las sesiones que
activan `normativa` antes de la primera herramienta, las que abren cada una de sus ocho referencias y en qué turno, y,
en `testigo-en-referencia`, las que cumplen la regla testigo; cada cifra con su intervalo de Wilson al 95 %.

**Why this priority**: es la medida que decide el ADR.

**Independent Test**: en `make ci`, sin modelo, el guion da las cifras exactas de unas sesiones fijas de los dos modos,
y la regla testigo da el veredicto esperado de unas respuestas fijas.

**Acceptance Scenarios**:

1. **Dado** un juego fijo de transcripciones de los dos modos, **Cuando** el guion las cuenta, **Entonces** da por
   variante y modo los recuentos de FR-020 y sus intervalos de Wilson con z = 1,96: 0 de 40 → 0,0–8,8 %;
   20 de 40 → 35,2–64,8 %; 40 de 40 → 91,2–100,0 % (FR-022, SC-005).
2. **Dado** una sesión que abre `normas.md` dos veces y ninguna otra referencia, **Cuando** el guion la cuenta,
   **Entonces** cuenta una apertura de `normas.md` en el turno de la primera y ninguna de las otras siete (FR-020).
3. **Dado** una respuesta fija a una de las cuatro evals que termina con la línea `Sello de lectura: KL-7Q4`, sola y sin
   formato, **Cuando** la comprueba el guion de la regla testigo, **Entonces** dice que cumple; **Dado** una que no
   termina así —sin la línea, con la línea en negrita, entre comillas o en código, o con la línea antes del final—, o la
   respuesta de `control`, **Entonces** dice que no (FR-030, SC-004).
4. **Dado** las sesiones fijas de un sondeo del laboratorio de H39, **Cuando** el guion las cuenta con este hito,
   **Entonces** da las mismas columnas y las mismas cifras que antes del hito (FR-023, SC-007).

---

### User Story 4 - La persona mide y decide antes de fusionar (Priority: P2)

Con la propuesta de cambio abierta, la persona lanza el sondeo con `claude-sonnet-5-5` y 40 sesiones por variante y
modo, deja su tabla en `evidencias/` y escribe un ADR que dice, con esas cifras, cuál de las tres diferencias explica
el defecto y qué debe cumplir la reorganización.

**Why this priority**: es la aceptación del hito, pero la hace una persona fuera del run (ADR 0032): el run entrega lo
que la hace posible.

**Independent Test**: el informe final del run dice que la propuesta queda pendiente del sondeo y del ADR; antes de
fusionar, `evidencias/` tiene la tabla y `docs/ADR/` el ADR.

**Acceptance Scenarios**:

1. **Dado** el arnés y el guion en la rama, **Cuando** la persona lanza el sondeo con `claude-sonnet-5-5`, 40 sesiones
   por variante y modo, en las cinco filas de FR-040, **Entonces** la tabla, con el modelo y la versión de Claude Code
   anotados, queda en `evidencias/` (FR-040).
2. **Dado** la tabla, **Cuando** la persona escribe el ADR, **Entonces** el ADR dice, con esas cifras, cuál de las tres
   diferencias explica que `normativa` no lea sus referencias en modo herramienta y qué debe cumplir la reorganización;
   si `control` no reproduce el defecto, o ninguna variante lo quita, lo dice y propone qué medir después (FR-041).

---

### Edge Cases

- **`skills/normativa` cambia** (el hito siguiente la reorganiza): `make ci` falla en la comparación de `control` hasta
  que la variante se iguala, y en la de cada otra variante si su diferencia ya no se puede aplicar tal como la fija el
  hito —un párrafo que cambia ya no está, o una línea que se queda ya no existe— (FR-002, FR-006).
- **Sesión sin ninguna herramienta de kitlegal**: si activó `normativa`, cuenta como activada antes de la primera
  herramienta; si no, no (FR-020).
- **Sesión que no termina o termina sin respuesta final**: va en su propia columna y no entra en el denominador
  (FR-021).
- **Referencia abierta varias veces**: cuenta una vez, en el turno de la primera apertura (FR-020).
- **Referencia que ninguna sesión de la fila abre**: «0 de n» con su intervalo, y su turno, vacío (FR-020).
- **Fila sin ninguna sesión terminada** (n = 0): recuentos, porcentajes, intervalos y turnos vacíos (FR-021).
- **Variante sin regla testigo** (`control`, `remisiones-en-orden`, `cuerpo-corto`): su columna de la regla testigo
  sale vacía, no a cero (FR-020).
- **Transcripción que el guion no sabe leer** (manipulada a mano, truncada por algo que no es el arnés): defecto
  `inesperado`, código 1 (ADR 0023), sin más casos.

## Requirements *(mandatory)*

### Functional Requirements

**Las variantes**

- **FR-001**: Las variantes son cuatro —`control`, `remisiones-en-orden`, `cuerpo-corto` y `testigo-en-referencia`—, cada
  una una skill completa con el nombre `normativa`, y viven fuera de `skills/` y de `evals/` y de todo lo que el binario
  empotra o `kitlegal skills install` copia, como el laboratorio de H39: el binario no las contiene, la instalación no
  las pone y el job de evals no las elige ni las mide. `make ci` falla si cualquiera de las tres cosas deja de
  cumplirse.
- **FR-002**: `control` es `skills/normativa` tal como está en el árbol: los mismos ficheros, sin ninguno de más ni de
  menos, idénticos byte a byte. `make ci` falla, y nombra el fichero, si uno difiere en un byte, falta o sobra.
- **FR-003**: `remisiones-en-orden` es `control` con cinco cambios en el cuerpo de `SKILL.md`; cada uno sustituye un
  párrafo o una viñeta entera, y lo demás —el resto del cuerpo, el frontmatter y las referencias— es idéntico byte a
  byte. En el texto sustituido, un salto de línea con la sangría de la línea que sigue cuenta como un espacio:
  1. el párrafo del «Protocolo» pasa a ``Antes de la primera orden o llamada, lee
     [con_herramientas](references/con_herramientas.md). Sigue los pasos en orden, cada operación con `kitlegal …
     --json` desde el `PATH` o, si la tienes, con su herramienta.``;
  2. el del paso 1 pasa a ``Lee [normas](references/normas.md) y toma el `BOE-A-…` de su tabla; si no está, lee
     [buscar_la_norma](references/buscar_la_norma.md) y búscala como dice.``;
  3. la viñeta del paso 2 que empieza por «Sin fecha —con ella» empieza por ``Con una fecha en la pregunta, lee
     [redaccion_en_una_fecha](references/redaccion_en_una_fecha.md) antes de leer el bloque. Sin fecha, lee cada
     bloque``, y sigue como en `control` desde «una vez, con su comprobación detrás»;
  4. la viñeta del paso 2 de varios bloques pasa a ``Para leer varios bloques a la vez, lee antes
     [varios_bloques](references/varios_bloques.md). Si una orden no termina con `0`, lee
     [lecturas_que_fallan](references/lecturas_que_fallan.md) antes de seguir.``;
  5. en el párrafo de «Cómo se cita», lo que sigue a «también del remitido» hasta el punto de la remisión pasa a
     ``. Antes de escribir la respuesta, lee [como_se_cita](references/como_se_cita.md) y repasa con ella las
     citas.``, y lo que viene después, desde «Cada línea `⚠`», sigue como en `control`.

  La remisión de «Redacción modificada» no cambia. El cuerpo conserva las ocho remisiones.
- **FR-004**: `cuerpo-corto` es `control` sin lo que no hace falta para leer y citar las evals 04, 05, 15 y 22. Las
  líneas de su cuerpo son líneas de `control`, en el mismo orden, sin añadir ni reescribir ninguna, salvo el encabezado
  `### 4. Responder`, que pasa a `### 3. Responder`. Contiene, enteros: el encabezado y el párrafo del «Protocolo»; los
  pasos 1 y 2; del paso 4, su encabezado y la viñeta que empieza por «Cada afirmación sobre el contenido de una norma
  lleva su cita»; y las secciones «Cómo se cita», «Redacción en una fecha» y «Redacción modificada». No contiene ninguna
  línea no vacía de los pasos 3, 5 y 6, de las otras viñetas del paso 4, de la sección «Comandos» con su tabla ni de las
  «Reglas». El frontmatter y las referencias son idénticos byte a byte a los de `control`, y el cuerpo conserva las ocho
  remisiones.
- **FR-005**: `testigo-en-referencia` es `control` con dos cambios: (a) `references/con_herramientas.md` es el de
  `control` seguido de una sección `## Cómo termina la respuesta` cuyo texto es el párrafo de la regla de
  `references/cierre.md` de la skill de laboratorio de H39, byte a byte, con su última línea `Sello de lectura: KL-7Q4`; (b) en el párrafo del
  «Protocolo», tras ``[con_herramientas](references/con_herramientas.md)`` y antes del punto final, se inserta
  `, que dice también cómo termina la respuesta`. Lo demás es idéntico byte a byte.
- **FR-006**: `make ci` compara cada una de las tres variantes con `control` según su declaración (FR-003 a FR-005): falla
  si algo fuera de lo declarado difiere en un byte, si un texto declarado no es exactamente el que fija el hito, o si
  la diferencia declarada ya no se puede aplicar sobre `control`.
- **FR-007**: El tope de cuerpo del ADR 0046 no se aplica a las variantes: el texto de `remisiones-en-orden` lo fija el hito y pasa
  de 11 000 caracteres (Contexto). El tope sigue igual para `skills/`.
- **FR-008**: En las sesiones del sondeo, el servidor `kitlegal mcp serve` y el binario son los del árbol, con sus
  instrucciones y las descripciones de sus parámetros de hoy: ninguna variante los cambia. Cada sesión tiene instalada solo la
  variante de su fila, con el nombre `normativa`, y ninguna otra skill: ni las demás del producto (`consulta-juridica`,
  `jurisprudencia`, `plazos-y-territorio`) ni las del laboratorio de H39. Es como prepara hoy sus sesiones el arnés de
  H39: el directorio de skills de la variante es el de la sesión. Las cifras de `control` no salen, por tanto, en el
  mismo entorno que los sondeos de partida de la #214 (que instalan las cuatro skills del producto); si `control` no
  reproduce el defecto con solo la variante, eso señala al entorno y no a una de las tres diferencias, y el ADR lo dice
  y propone qué medir después (FR-041).

**El arnés del sondeo**

- **FR-010**: `make evals-laboratorio` lanza también las variantes de `normativa`, que la persona pide por su nombre sin
  confundirlas con las del laboratorio de H39. Por variante y modo pedidos, abre el número de sesiones que pide la
  persona, repartidas por igual entre las evals 04, 05, 15 y 22 de `normativa` (diez por eval con las 40 de la
  aceptación), servidas por sus lecturas grabadas, sin red a las fuentes (ADR 0043). `control` se mide en los dos modos y
  las otras tres solo en modo herramienta. Pedir una variante que no existe, una variante en un modo que no le
  corresponde, un número de sesiones que no es múltiplo de 4 o una petición que mezcla variantes del laboratorio de H39 y
  variantes de `normativa` es un error de argumentos: sale con un código distinto de 0, dice qué no vale (en la
  mezcla, que se mezclan variantes del laboratorio y de `normativa`), uno por línea y con el nombre del argumento de
  `make`, y no abre ninguna sesión, no construye nada y no crea `DIRECTORIO`, como en H39. Cada invocación es de una
  sola familia, y su tabla tiene las columnas de esa familia (FR-020, FR-023). Cómo se nombran las variantes para no
  confundir las dos `control` lo decide el plan.
- **FR-011**: Antes de abrir ninguna sesión, el arnés comprueba, como en H39, que la versión de Claude Code instalada es
  la pedida y que el modelo es un identificador completo (`claude-<familia>-<versión>`, no un alias); si no, sale con un
  código distinto de 0, nombra lo que no coincide y no abre ninguna. Anota los dos en la tabla.
- **FR-012**: El arnés lo lanza una persona, con la credencial del sondeo de hoy (ADR 0032); ni `make ci`, ni un gancho,
  ni el job de evals, ni un paso del workflow lo ejecutan. Conserva la transcripción de cada sesión, con su variante, su
  modo y su eval, donde el guion la lee.
- **FR-013**: El run no graba nada: no añade ni cambia ninguna grabación de `testdata/` y no escribe en `evidencias/`.

**La tabla**

- **FR-020**: `make laboratorio-tabla` lee las transcripciones de un sondeo de las variantes de `normativa`, sin red ni
  modelo, y da una fila por variante y modo, con el modelo y la versión de Claude Code de FR-011, con:
  - las sesiones que activan `normativa` antes de la primera herramienta de kitlegal;
  - por cada una de las ocho referencias de la variante, las sesiones que la abren y el turno de su primera apertura (en
    la tabla, la mediana —con un decimal si cae entre dos turnos— y los extremos; el de cada sesión, en el detalle);
  - en `testigo-en-referencia`, las sesiones que cumplen la regla testigo según FR-030; en las otras, la columna vacía.

  Cada recuento va como «k de n» con su porcentaje y su intervalo de Wilson al 95 % (z = 1,96), en porcentaje con un
  decimal. Turno, activar una skill, primera herramienta de kitlegal en modo herramienta, primera orden `kitlegal` en
  modo orden y abrir una referencia cuentan exactamente como en la tabla de H39 (`specs/031-h39-c-omo-se/spec.md`, el
  requisito de lo que cuenta su tabla): abrir
  una referencia es un `tool_use` de `Read` con su `file_path`, entero o por tramos, o de `Bash` cuya orden nombra
  `references/<fichero>.md` de la skill; `Grep` y `Glob` no cuentan; `Skill` en el mismo turno que la primera
  herramienta no es antes. Una referencia abierta varias veces cuenta una, en el turno de la primera; una que ninguna
  sesión abre sale «0 de n» con su intervalo y su turno vacío.
- **FR-021**: Las sesiones que no terminan o terminan sin respuesta final van en su propia columna y no entran en `n`;
  con `n` = 0, los recuentos, porcentajes, intervalos y turnos de la fila salen vacíos.
- **FR-022**: `make ci` comprueba el guion con transcripciones fijas de las variantes de `normativa` en los dos modos que
  cubren cada columna de FR-020 y FR-021: una referencia abierta dos veces, una referencia que ninguna sesión abre,
  `Skill` y la primera herramienta en el mismo turno, una sesión sin terminar, una fila sin ninguna sesión terminada, y
  una sesión que cumple la regla testigo y otra que no; y con los intervalos de la historia 3. El test falla si un
  recuento, un turno o un extremo de un intervalo difiere.
- **FR-023**: Lo que H39 ya mide no cambia: con las transcripciones fijas del laboratorio de H39, el guion da las mismas
  columnas y las mismas cifras que antes del hito, y sus tests en `make ci` pasan sin cambiar ningún valor esperado.
  La suite de aceptación del hito, congelada en su primera tarea, pasa al guion las transcripciones fijas de H39 tal
  como están en `main` y compara su salida con la tabla esperada de H39 tal como está en `main`; falla si difiere una
  columna o una cifra, aunque el hito cambie esas transcripciones, esa tabla o los valores de los tests de H39.

**La regla testigo**

- **FR-030**: La regla testigo es la de H39, con el mismo veredicto: cumple la respuesta final cuya última línea es,
  sola y sin ningún formato, `Sello de lectura: KL-7Q4`. `make ci` la comprueba con respuestas fijas a las evals del
  sondeo: al menos una que la cumple y al menos una por cada forma de no cumplirla que la regla distingue (sin la línea;
  la línea con formato; la línea que no es la última), más una respuesta de `control` sin la línea; el test falla si
  alguna sale con el veredicto contrario.

**La aceptación de la persona y las evals del producto**

- **FR-040**: Antes de fusionar, la persona lanza el sondeo con `claude-sonnet-5-5` y 40 sesiones por variante y modo en
  cinco filas —`control` en los dos modos; `remisiones-en-orden`, `cuerpo-corto` y `testigo-en-referencia` en modo
  herramienta—, 200 sesiones, y deja su tabla en `evidencias/`.
- **FR-041**: Antes de fusionar, la persona escribe un ADR que dice, con las cifras de FR-040, cuál de las tres
  diferencias explica que `normativa` no lea sus referencias en modo herramienta y qué debe cumplir la reorganización.
  Si `control` no reproduce el defecto, o ninguna variante lo quita, el ADR lo dice y propone qué medir después.
- **FR-042**: El run no lanza el sondeo ni escribe el ADR; los dos se añaden a la propuesta de cambio antes de
  fusionarla, y la propuesta no se fusiona sin ellos. El informe final del run lo dice.
- **FR-043**: Las evals de las skills del producto salen en verde en la medición del cierre. Si el cierre da un rojo de
  partida en modo herramienta (como la eval 04 de `normativa`, `sin_activar`, en la ejecución 41474010633), la
  reparación no lo arregla tocando `skills/` ni reorganizando una skill, y el informe lo da con sus cifras para que
  decida una persona.

### Key Entities

- **Variante de `normativa`**: una skill completa con el nombre `normativa`, fuera del producto, que es la de `skills/`
  (`control`) o la de control con una diferencia declarada; con los modos en que se mide.
- **Declaración de una variante**: qué partes cambian (cuerpo, una referencia) y con qué texto exacto (FR-003 a FR-005).
- **Eval del sondeo**: una de las evals 04, 05, 15 y 22 de `normativa`, con sus lecturas grabadas.
- **Regla testigo** y su guion: la de H39; da cumple o no cumple sobre una respuesta.
- **Sondeo**: las sesiones de las variantes por modo, con su modelo y su versión de Claude Code.
- **Tabla del sondeo**: una fila por variante y modo con los recuentos de FR-020 y sus intervalos, más el detalle por
  sesión.

### Uso, de fuera adentro

Ninguna skill del producto ni el binario que se distribuye consume nada de lo que entrega el hito; lo consumen el agente
de las sesiones del sondeo y la persona que escribe el ADR.

- **Variantes** (FR-001 a FR-005). Las lee el agente de una sesión del sondeo al activar `normativa`, una vez por
  sesión, y sus referencias cuando las abre: lo que se mide es justo cuántas abre. Tamaño fijo: el de `normativa` hoy
  (10 929 caracteres de cuerpo y ocho referencias), unos 150 más en `remisiones-en-orden`, unos 3300 de cuerpo en
  `cuerpo-corto` y 44 más una sección en `testigo-en-referencia`. No crecen con el uso del producto —cientos de normas y
  miles de bloques consultados no las tocan— y solo cambian cuando cambia `skills/normativa` (FR-002). No son una señal.
- **Veredicto de la regla testigo** (FR-030). Lo pide el guion una vez por sesión de `testigo-en-referencia`: 40
  veredictos de una palabra en el sondeo de la aceptación. Solo existe en el sondeo.
- **Tabla del sondeo** (FR-020, FR-021). La lee la persona una vez, para escribir el ADR: cinco filas, cada una con la
  activación, ocho referencias con su recuento y su turno, la regla testigo y las sesiones sin terminar, más el detalle
  de 200 sesiones. Su tamaño lo fijan las variantes, los modos y las sesiones pedidas, no el uso del producto. No es una
  señal que caduque: deja de leerse cuando el ADR está escrito, y un sondeo nuevo da otra tabla.

## Success Criteria *(mandatory)*

### Measurable Outcomes

- **SC-001**: 0 bytes de diferencia entre la variante `control` y `skills/normativa`, y 0 ficheros de más o de menos.
  Control: el test de FR-002 en `make ci` falla con el primer fichero que difiere, falta o sobra.
- **SC-002**: Las tres variantes difieren de `control` en 0 bytes fuera de lo que declaran, y lo declarado coincide
  carácter a carácter con el texto del hito. Control: el test de FR-006 en `make ci`.
- **SC-003**: 0 ficheros de las variantes en el binario, en lo que instala `kitlegal skills install` y en las evals que
  elige el job. Control: el test de FR-001 en `make ci`.
- **SC-004**: El guion de la regla testigo acierta el 100 % de las respuestas fijas de FR-030. Control: su test en
  `make ci` falla con la primera que no acierta.
- **SC-005**: El guion da los recuentos, los turnos y los extremos de Wilson con un decimal exactos de las
  transcripciones fijas de FR-022. Control: su test en `make ci` falla con el primer valor que difiere.
- **SC-006**: Con una variante inexistente o en un modo que no le corresponde, un número de sesiones que no es múltiplo
  de 4, una petición que mezcla variantes del laboratorio de H39 y de `normativa`, una versión de Claude Code distinta de la pedida o un modelo que no es un id completo, el arnés abre 0 sesiones;
  con una petición válida, abre n/4 por eval en cada variante y modo pedidos. Control: un test de `make ci` con un
  sustituto de `claude` que cuenta sus invocaciones, sin modelo, falla si una petición inválida abre alguna sesión o si
  una válida no abre n/4 por eval en cada variante y modo.
- **SC-007**: 0 columnas y 0 cifras distintas de la tabla esperada de H39 en `main` para sus transcripciones fijas.
  Control: la comparación de la suite de aceptación congelada de FR-023 en `make ci` falla con la primera que difiere,
  también si el hito cambia la tabla esperada o los tests de H39.
- **SC-008**: Las evals de las skills del producto en verde en la medición del cierre. Control: el job de evals del
  cierre sale en rojo si una no pasa su umbral, y el informe final da sus cifras (FR-043).
- **SC-009**: Antes de fusionar, la tabla de las cinco filas de 40 sesiones está en `evidencias/` y el ADR en
  `docs/ADR/`. Lo comprueba la persona (capa 3): ningún control automático lo mide, y el run no lo hace (FR-042).

## Criterios del hito (literales)

**Controles**: `make ci` comprueba que la variante `control` es idéntica byte a byte a `skills/normativa`, para que no
se desvíe cuando cambie la skill (FR-002, SC-001); que cada variante difiere de la de control solo en lo que declara
(FR-006, SC-002); que ninguna se empotra ni se instala (FR-001, SC-003); la regla testigo con respuestas fijas que la
cumplen y que no (FR-030, SC-004); y la tabla con sesiones grabadas (FR-022, FR-023, SC-005, SC-007). Las evals de las
skills del producto salen en verde en la medición del cierre (FR-043, SC-008).

**Aceptación**: el sondeo de la persona, con `claude-sonnet-5-5` y 40 sesiones por variante y modo, deja su tabla en
`evidencias/` (FR-040). Un ADR dice, con esas cifras, cuál de las tres diferencias explica que `normativa` no lea sus
referencias en modo herramienta, y qué debe cumplir la reorganización. Si `control` no reproduce el defecto, o ninguna
variante lo quita, el ADR lo dice y propone qué medir después (FR-041). El run no lanza el sondeo ni escribe el ADR: los
dos se añaden a la propuesta de cambio antes de fusionarla, y la propuesta no se fusiona sin ellos (FR-042).

## Assumptions

- Las cifras de partida (#214, H39, `brazo-D`) son las que citan el hito y el ADR 0050; esta sesión no las ha medido.
- Las evals 04, 05, 15 y 22 de `normativa` y sus lecturas grabadas son las que usa hoy el job de evals; el hito dice que
  ya están grabadas y el run no graba ninguna (FR-013).
- Las variantes no son skills del producto: no se distribuyen ni las mide el job, y el tope de cuerpo del ADR 0046 rige
  `skills/`. Por eso `remisiones-en-orden` puede pasar de 11 000 caracteres sin cambiar el tope (FR-007): su texto lo
  fija el hito, y con él la diferencia de tamaño respecto de `control` es de un 1,4 %, frente al 70 % de `cuerpo-corto`.
- «Antes de la primera herramienta» y el turno se miden como en H39 (FR-020): el mismo turno no es antes.
- El modelo se pasa como identificador completo porque un alias lo resuelve el cliente (ADR 0031).

## Fuera de alcance

- Arreglar o reorganizar `normativa` u otra skill del producto, también para arreglar un rojo de partida: es el hito
  siguiente, que se escribe con el resultado.
- Cambiar ningún umbral, límite ni tope: el de 11 000 caracteres y el de 900 (ADR 0046), los de las instrucciones del
  servidor, los umbrales del job de evals.
- Ampliar el laboratorio de H39: variantes nuevas de su skill, evals de laboratorio nuevas o cambios en lo que su tabla
  mide.
- Medir con el juez (`afirma_lo_no_leido` y demás clases): el sondeo cuenta lecturas y forma, sin modelo.
- Medir `remisiones-en-orden`, `cuerpo-corto` o `testigo-en-referencia` en modo orden; variantes que combinen dos
  diferencias; variantes de otras skills; otras evals que las 04, 05, 15 y 22; otros modelos que `claude-sonnet-5-5` u
  otros hosts que Claude Code.
- Cambiar las instrucciones del servidor o las descripciones de los parámetros en las sesiones del sondeo.
- Lanzar en una invocación variantes del laboratorio de H39 y de `normativa`, en una tabla o en dos (FR-010): se
  consigue con dos invocaciones, una por familia.
- Instalar en las sesiones del sondeo otras skills que la variante de la fila, del producto o del laboratorio (FR-008).
- Columnas de la tabla que FR-020 no enumera (qué herramientas se cargan o se llaman, `ToolSearch`, la lectura del
  `SKILL.md`).
- Que el job de evals mida las variantes, o que `make ci` abra sesiones con modelo.
- Grabar lecturas nuevas o escribir en `evidencias/` desde el run.
- Lanzar el sondeo y escribir el ADR dentro del run (los hace la persona antes de fusionar), y llevar sus cifras a
  `docs/SKILLS.md`.
