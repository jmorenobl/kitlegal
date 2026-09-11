# Feature Specification: H1 · Kernel CLI: multicall, flags globales, exit codes, sobre de salida

**Feature Branch**: `h1-kernel-cli-multicall`

**Created**: 2026-09-11

**Status**: Draft

**Input**: Sección `#### H1 · Kernel CLI: multicall, flags globales, exit codes, sobre de salida` de `docs/ROADMAP.md` (modo desatendido).

## Resumen

H0 dejó un repositorio blindado pero sin producto: el binario solo sabe decir su versión. H1 escribe **una vez** el patrón que todos los applets del proyecto repetirán después —`boe`, `cita`, `plazos`, `placsp`, `bdns`…— para que ninguno tenga que inventarlo ni desviarse de él.

Ese patrón tiene cuatro piezas y ninguna es opcional:

1. **Un solo ejecutable con muchos nombres.** El binario decide qué applet ejecuta a partir del nombre con que se le invoca o del primer argumento, de modo que un enlace simbólico `boe -> kitlegal` permita a una skill seguir escribiendo `scripts/boe articulo …`.
2. **Un sobre de salida idéntico en todos los applets**: `{ok, fuente, url, fecha_consulta, hash, data}`. Sin `fuente`, `url`, `fecha_consulta` y `hash` no hay cita, y sin cita el proyecto no sirve para lo que existe.
3. **Un juego de banderas globales y unos códigos de salida estables**, para que un agente pueda invocar cualquier applet sin leer su documentación y distinguir «no existe» de «la fuente está caída» sin analizar un mensaje de texto.
4. **Unas reglas de arquitectura que se hacen cumplir solas**: quien escriba el applet número doce no puede saltarse el patrón aunque quiera, porque el lint y un test de arquitectura lo impiden.

El valor de H1 no es responder una consulta legal —eso llega en H4—, sino que a partir de aquí **añadir un applet sea declarar un comando y devolver un `data`**, y que todo lo demás (banderas, sobre, `hash`, renderizado, códigos de salida, registro de eventos) venga dado. El applet `echo` es el ejemplo mínimo que demuestra el patrón, no una funcionalidad del producto.

## Clarifications

### Session 2026-09-11

- Q: ¿El binario `kitlegal` que se distribuye registra el applet de ejemplo `echo`, o la entrega literal del hito (`kitlegal echo hola --json` y `ln -s kitlegal echo && ./echo hola`) se demuestra sobre un binario compilado por el test e2e a partir del applet de ejemplo de `internal/app/testdata`? (FR-009) → A: El binario distribuido NO registra `echo`; el applet de ejemplo vive únicamente en material de test (`internal/app/testdata`) y la entrega literal del hito se demuestra en el test e2e con `testscript`, que construye un binario a partir del kernel real (`cmd`/`internal/app`) más el registro del applet de ejemplo; la superficie del binario distribuido al cerrar H1 es la de H0 (`version`) más el kernel, y el primer applet de producto llega en H4 (`boe`) (auto: criterio a; fuente: docs/ROADMAP.md §4 H1 «Alcance» y «Entrega»; spec.md «Assumptions»; constitución §V y «Criterio de decisión autónoma» §2).
- Q: ¿Qué valores toman `fuente` y `url` en un applet que no consulta ninguna fuente externa —`echo` hoy, `cita` y `plazos` en H8 y H9? (FR-016) → A: `fuente` y `url` nunca van vacías; el esquema del sobre las exige no vacías y `url` con formato URI. Un applet sin fuente externa usa el espacio de nombres reservado: `fuente` con el prefijo `kitlegal.` (p. ej. `kitlegal.echo`) y `url` con el esquema de URI `kitlegal:` (p. ej. `kitlegal:applet/echo`); ningún adaptador de `internal/source/*` puede usar ese prefijo ni ese esquema, y un sobre con procedencia `kitlegal:` es un resultado calculado, no una cita de fuente pública (auto: criterio c; fuente: Constitución §II; CLAUDE.md «Sobre de salida obligatorio»; refs/kitlegal-estructura-y-ecosistema.md §2; «Criterio de decisión autónoma» §1 y §2).
- Q: Cuando un applet falla y se pidió `--json`, ¿qué se escribe exactamente en la salida estándar y dónde viaja la causa del fallo, dado que el sobre no tiene campo de error? (FR-045) → A: La salida estándar lleva el mismo sobre de seis claves con `ok: false`, y la causa (clase de error y mensaje para la persona) va dentro de `data`; las seis claves del nivel superior no cambian y no se añade ninguna. Reglas: (1) lo emite el kernel, no el applet, desde el único punto que traduce el error a código de salida, de modo que la representación es idéntica para todos los applets y para los fallos anteriores a la ejecución del applet; (2) `ok: false` si y solo si el código de salida no es 0; (3) `fuente` y `url` son las de la fuente que se estaba consultando cuando se conocen (p. ej. la URL que devolvió «no encontrado», útil como cita negativa) y, en caso contrario, el espacio de nombres reservado de Q2 referido al kernel (p. ej. `kitlegal.cli` / `kitlegal:cli`); `fecha_consulta` y `hash` se calculan como siempre, con la huella sobre la forma canónica del `data` de error; (4) el mensaje para la persona sigue yendo además a la salida de error y el código de salida sigue siendo la vía primaria de clasificación: el sobre de fallo la duplica en forma estructurada, no la sustituye; (5) el esquema de salida que emite `--describe` describe `data` condicionado a `ok`: el `data` propio del applet cuando `ok` es verdadero y la forma común de error del kernel cuando es falso. Los nombres exactos de las claves dentro del `data` de error y la técnica de esquema condicional los fija el plan (auto: criterio c; fuente: CLAUDE.md «Sobre de salida obligatorio»; spec.md FR-014, FR-016, FR-017, FR-045, FR-047; docs/ROADMAP.md H1 «Aceptación» y H22; constitución §II y §IV).
- Q: ¿Qué escribe en la salida estándar un applet invocado con `--dry-run` y `--json`? (FR-022, FR-010) → A: La salida estándar queda vacía, también con `--json`; la descripción de la operación que se habría realizado va a la salida de error (siempre visible, no depende de `--verbose`) y el código de salida es 0, porque el sobre es una cita de un contenido consultado y una operación no realizada no tiene nada que citar; el kernel no corta antes del applet, la bandera viaja en el Contexto de ejecución y cada capa con efectos la honra (auto: criterio c; fuente: docs/ROADMAP.md §3 fila «Observabilidad local» H1 y §2 reglas de dependencia; H2 «Controles»; constitución §II; «Criterio de decisión autónoma» §2).
- Q: ¿Qué ocurre cuando se combinan `--json` y `--help`, es decir, cuando se pide salida legible por máquina para una salida que no es el sobre de un applet? (FR-028, FR-026, FR-042) → A: La ayuda se emite siempre como texto para personas en la salida estándar, derivada del registro de applets, con código de salida 0; `--json` no la altera y la combinación se comporta igual que `--help` solo, en cualquier orden. FR-042 y SC-002 quedan acotados a la presentación del resultado de un applet, no a la ayuda ni a `--dry-run`; la contraparte legible por máquina de la ayuda es `--describe` (auto: criterio c; fuente: spec.md FR-026, FR-046; docs/ROADMAP.md H1 «Controles»; refs/kitlegal-estructura-y-ecosistema.md §3; «Criterio de decisión autónoma» §1 y §2).

## User Scenarios & Testing *(mandatory)*

### User Story 1 - Toda respuesta es un sobre citable (Priority: P1)

Un agente (o una persona) invoca un applet pidiendo salida legible por máquina y recibe siempre la misma estructura: si la respuesta fue correcta, de qué fuente procede, en qué URL puede comprobarse, cuándo se consultó, una huella del contenido que permite detectar si cambió, y el contenido mismo. No necesita saber qué applet es: el envoltorio es el mismo para todos.

**Why this priority**: Es el principio II de la constitución («Nada sin cita ni fuente») convertido en contrato ejecutable y la primera mitad de la entrega literal del hito. Todo lo demás del proyecto —el grafo, las skills, `cita`, los escritos— asume que este sobre existe y es fiable. Sin él no hay nada que citar.

**Independent Test**: Se puede probar por completo invocando el applet de ejemplo con la bandera de salida legible por máquina y comprobando que lo que se escribe en la salida estándar es un único documento JSON con exactamente esas seis claves, que la huella corresponde al contenido y que dos invocaciones con el mismo contenido producen la misma huella.

**Acceptance Scenarios**:

1. **Given** el binario que compila el test e2e, con el applet de ejemplo registrado (FR-009), **When** se invoca `kitlegal echo hola --json`, **Then** la salida estándar contiene un único documento JSON con exactamente las claves `ok`, `fuente`, `url`, `fecha_consulta`, `hash` y `data`, y el comando termina con código 0. *(entrega literal del hito)*
2. **Given** el binario construido, **When** se invoca un applet con salida legible por máquina, **Then** la salida estándar contiene **solo** ese JSON: ni una línea de registro, ni un aviso, ni una cabecera. *(criterio de aceptación literal del hito)*
3. **Given** dos invocaciones que producen el mismo contenido en `data`, **When** se comparan sus huellas, **Then** son idénticas, con independencia del orden en que se serializaron las claves y del momento de la consulta.
4. **Given** dos invocaciones cuyo contenido en `data` difiere en un solo byte, **When** se comparan sus huellas, **Then** son distintas.
5. **Given** el sobre emitido por cualquier applet, **When** se inspecciona `fecha_consulta`, **Then** es una marca temporal con zona horaria explícita y no un texto libre.
6. **Given** una salida sin la bandera de salida legible por máquina, **When** se observa la salida estándar, **Then** se presenta el mismo contenido en una tabla mínima legible por una persona, y los cuatro datos de procedencia (`fuente`, `url`, `fecha_consulta`, `hash`) siguen estando presentes.

---

### User Story 2 - Un binario, muchos nombres (Priority: P2)

Una skill instalada en un cliente de agentes llama a `scripts/boe articulo …`. Ese `scripts/boe` es un enlace simbólico al único binario del proyecto. El binario reconoce con qué nombre lo han llamado y ejecuta el applet correspondiente, sin que la skill tenga que saber que por debajo hay un solo ejecutable. La misma invocación funciona escribiendo el applet como primer argumento.

**Why this priority**: Es la segunda mitad de la entrega literal del hito, la decisión cerrada «multicall» de `CLAUDE.md` y el mecanismo que permite migrar la skill `boe-fiscal` en H5 sin tocar su `SKILL.md`. Depende de que exista al menos un applet registrado (P1), pero es independiente del contenido del sobre.

**Independent Test**: Se puede probar de forma aislada creando un enlace simbólico con el nombre de un applet registrado y comprobando que invocarlo por ese nombre produce exactamente el mismo resultado que invocar el binario con el applet como primer argumento.

**Acceptance Scenarios**:

1. **Given** el binario que compila el test e2e, con el applet de ejemplo registrado (FR-009), colocado en un directorio, **When** se crea un enlace simbólico `echo` que apunta a él y se ejecuta `./echo hola`, **Then** el resultado es idéntico —salida y código de salida— al de `kitlegal echo hola`. *(entrega literal del hito)*
2. **Given** el binario invocado por su nombre propio sin ningún argumento, **When** termina, **Then** describe los applets disponibles y devuelve el código de salida reservado a los errores de argumentos.
3. **Given** el binario invocado con un primer argumento que no corresponde a ningún applet registrado, **When** termina, **Then** el mensaje de error nombra el applet desconocido, la lista de applets disponibles va a la salida de error y el código de salida es el reservado a los errores de argumentos.
4. **Given** un enlace simbólico cuyo nombre no corresponde a ningún applet registrado, **When** se invoca, **Then** el binario se comporta como si se le hubiera invocado por su nombre propio y trata el primer argumento como el applet.
5. **Given** el binario o cualquiera de sus enlaces, **When** se pide la ayuda, **Then** la lista de applets y verbos que se muestra procede del registro de applets, sin ninguna lista duplicada mantenida a mano.

---

### User Story 3 - Los fallos se distinguen sin leer el mensaje (Priority: P3)

Un agente invoca un applet y algo falla. En lugar de tener que interpretar un texto en español, recibe un código de salida que clasifica el fallo: argumentos mal escritos, resultado inexistente, fuente no disponible, límite de peticiones o términos de uso, o acción que requiere identidad humana. Con eso decide solo: reintentar, cambiar de estrategia, o detenerse y pedir a una persona que actúe.

**Why this priority**: Los códigos de salida estables son una decisión cerrada (`CLAUDE.md`) y la vía por la que la frontera humana se hace observable (código 6). Son también la mitad de los controles unitarios exigidos por el hito. No bloquean la entrega del sobre, pero sin ellos ningún applet posterior puede fallar correctamente.

**Independent Test**: Se puede probar de forma aislada provocando desde un applet de prueba cada clase de error y comprobando, para cada una, el código de salida que produce, que el mensaje va a la salida de error y no a la estándar, y que con la bandera de salida legible por máquina la salida estándar lleva el sobre de seis claves con `ok` falso y la causa dentro de `data`.

**Acceptance Scenarios**:

1. **Given** una invocación con argumentos inválidos (bandera desconocida, verbo ausente, valor con formato incorrecto), **When** el binario termina, **Then** el código de salida es 2. *(control literal del hito: «exit 2 con args malos»)*
2. **Given** un applet que señala que lo pedido no existe, **When** el binario termina, **Then** el código de salida es 3.
3. **Given** un applet que señala que la fuente no está disponible, **When** el binario termina, **Then** el código de salida es 4.
4. **Given** un applet que señala un límite de peticiones o una restricción de términos de uso, **When** el binario termina, **Then** el código de salida es 5.
5. **Given** un applet que señala que la acción requiere identidad humana, **When** el binario termina, **Then** el código de salida es 6 y no se ha realizado ninguna acción con efectos externos.
6. **Given** una invocación correcta, **When** el binario termina, **Then** el código de salida es 0.
7. **Given** cualquier fallo, **When** se observa la salida, **Then** el mensaje dirigido a la persona va a la salida de error y nunca se mezcla con la salida estándar.
8. **Given** cualquier fallo en una ruta de usuario, **When** el binario termina, **Then** lo hace por su camino normal de salida y nunca por un pánico.
9. **Given** cualquiera de las clases de error y una invocación con salida legible por máquina, **When** el binario termina, **Then** la salida estándar contiene un único documento JSON con exactamente las seis claves del sobre, `ok` falso, y la clase del error y el mensaje para la persona dentro de `data`; `fuente`, `url`, `fecha_consulta` y `hash` no van vacías y el documento valida contra la misma descripción formal del sobre que una respuesta correcta.
10. **Given** un fallo que ocurre antes de llegar al applet —una bandera desconocida o un applet no registrado— y una invocación con salida legible por máquina, **When** el binario termina, **Then** el sobre emitido lleva la procedencia reservada del kernel (`kitlegal.` / `kitlegal:`), `ok` falso, la clase «argumentos» en `data`, y el código de salida es 2.
11. **Given** un fallo con salida legible por máquina, **When** se comparan el sobre y el código de salida, **Then** la clase declarada en `data` corresponde al código de salida emitido, y `ok` es falso si y solo si el código no es 0.

---

### User Story 4 - Las mismas banderas en todos los applets (Priority: P4)

Quien escribe una skill —o el agente que la ejecuta— aprende un único juego de banderas y lo usa con cualquier applet: pedir salida legible por máquina, limitar el tiempo, trabajar sin red, ver qué haría sin hacerlo, no tocar el grafo, trabajar sobre un asunto concreto y ver qué está ocurriendo por dentro. Ningún applet inventa las suyas ni cambia el significado de las comunes.

**Why this priority**: Es el «alcance» literal del hito y la convención de agente de `CLAUDE.md`. Su valor se materializa cuando hay applets reales (H4 en adelante), por eso va después del sobre, del despacho y de los códigos de salida; pero fijarlo ahora es exactamente lo que evita que cada applet las reinvente.

**Independent Test**: Se puede probar de forma aislada registrando dos applets de prueba distintos y comprobando que ambos aceptan las ocho banderas globales con idéntica sintaxis y semántica sin que ninguno de los dos las declare.

**Acceptance Scenarios**:

1. **Given** un applet cualquiera registrado, **When** se le pasan `--json`, `--timeout`, `--offline`, `--dry-run`, `--describe`, `--no-graph`, `--asunto` y `--verbose`, **Then** todas son aceptadas sin que el applet las haya declarado.
2. **Given** la bandera de tiempo máximo, **When** se invoca un applet, **Then** el límite se aplica a toda la ejecución de la operación y, si se agota, el binario termina con el código reservado a fuente no disponible.
3. **Given** la bandera de ejecución simulada, **When** se invoca un applet sin pedir detalle adicional ni con la bandera ni con la variable de entorno, **Then** la descripción de la operación que habría realizado aparece en la salida de error, la operación no se realiza, la salida estándar queda vacía —también si se añade la bandera de salida legible por máquina— y el código de salida es 0.
4. **Given** la bandera de detalle, **When** se invoca un applet, **Then** el registro de eventos internos aparece en la salida de error con mayor nivel de detalle y la salida estándar no cambia en absoluto.
5. **Given** una variable de entorno que fija el nivel de registro, **When** se invoca un applet, **Then** el nivel de registro cambia sin necesidad de pasar ninguna bandera.
6. **Given** un applet nuevo que solo declara su nombre, sus verbos y su resultado, **When** se registra, **Then** hereda las ocho banderas, el sobre, los códigos de salida y el renderizado sin escribir código para ninguno de ellos.
7. **Given** la bandera de ejecución simulada, **When** se invoca un applet, **Then** el kernel no corta la ejecución antes de entregarle el control: la bandera le llega en el contexto de ejecución, igual que las de trabajo sin red, uso del grafo y asunto, y es el applet —y en hitos posteriores cada capa con efectos— quien describe en lugar de ejecutar.

---

### User Story 5 - El applet se describe a sí mismo (Priority: P5)

Un consumidor automático —el generador de herramientas MCP, el generador de la tabla de comandos de una skill, o un agente que descubre el binario por primera vez— pregunta al applet qué entrada acepta y qué salida produce, y recibe una descripción formal en lugar de tener que leer documentación escrita para personas.

**Why this priority**: Es la tercera parte de la entrega literal del hito y la pieza que en fases posteriores evita que existan tres descripciones distintas del mismo applet (binario, skill y servidor MCP). Depende de que el sobre y las banderas estén fijados.

**Independent Test**: Se puede probar de forma aislada invocando el applet de ejemplo con la bandera de autodescripción y comprobando que lo emitido es un esquema JSON válido que describe la entrada y la salida de ese applet, y que los datos de salida corresponden a la estructura del sobre.

**Acceptance Scenarios**:

1. **Given** un applet registrado —en la entrega literal del hito, el de ejemplo en el binario que compila el e2e (FR-009)—, **When** se le invoca con la bandera de autodescripción, **Then** emite en la salida estándar un esquema JSON válido que describe su entrada y su salida, y termina con código 0. *(entrega literal del hito)*
2. **Given** la autodescripción de cualquier applet, **When** se inspecciona la parte de salida, **Then** describe el sobre `{ok, fuente, url, fecha_consulta, hash, data}` con `data` condicionado a `ok`: el contenido propio de ese applet cuando `ok` es verdadero y la forma común de error del kernel cuando es falso.
3. **Given** la bandera de autodescripción, **When** se invoca, **Then** el applet no ejecuta su operación: describirse y actuar son excluyentes.
4. **Given** un applet cuyos verbos o argumentos cambian, **When** se vuelve a pedir su autodescripción, **Then** el esquema refleja el cambio sin que nadie lo haya editado a mano.
5. **Given** la autodescripción de un applet y una ejecución de ese mismo applet que termina en error con `--json`, **When** se valida el sobre de fallo emitido contra esa autodescripción, **Then** lo cumple: el esquema describe tanto la salida correcta como la fallida del applet.

---

### User Story 6 - Las reglas de arquitectura se hacen cumplir solas (Priority: P6)

Quien añada código al proyecto —persona o agente— no puede saltarse las reglas de dependencia aunque lo intente: el dominio no puede importar adaptadores, solo un paquete puede hablar HTTP, solo unos pocos pueden tocar la base de datos, solo el kernel y el punto de entrada pueden terminar el proceso, y solo el presentador puede escribir en la salida estándar. Quien lo intenta recibe un fallo, no un aviso en una revisión.

**Why this priority**: Es un control literal del hito («test de arquitectura; `depguard` activo») y lo que hace que el patrón escrito en H1 sobreviva a los veinticinco hitos siguientes. Va al final porque necesita que existan los paquetes a los que las reglas se refieren.

**Independent Test**: Se puede probar de forma aislada introduciendo, en una copia desechable del árbol, una violación de cada regla y comprobando que el control correspondiente falla y nombra la regla violada.

**Acceptance Scenarios**:

1. **Given** el proyecto con los controles de H1 activos, **When** un paquete del dominio importa un adaptador, **Then** el control de reglas de dependencia falla y nombra la regla.
2. **Given** el proyecto con los controles activos, **When** un paquete distinto del cliente HTTP importa la biblioteca HTTP estándar, **Then** el control falla.
3. **Given** el proyecto con los controles activos, **When** un paquete distinto del kernel o del punto de entrada termina el proceso, **Then** el control falla.
4. **Given** el proyecto con los controles activos, **When** un paquete distinto del presentador escribe directamente en la salida estándar, **Then** el control falla.
5. **Given** el proyecto tal y como queda al cerrar H1, **When** se ejecutan todos los controles, **Then** ninguna regla está violada y la orden agregada de integración continua termina en verde.

---

### Edge Cases

- **El nombre del ejecutable y el primer argumento compiten**: el binario se invoca por un enlace llamado `echo` y además se le pasa `boe` como primer argumento. La regla de precedencia debe estar fijada y ser comprobable (FR-004).
- **Enlace simbólico con un nombre que no es un applet** (p. ej. `kitlegal-dev -> kitlegal`): no puede hacer inutilizable el binario (FR-005).
- **Invocación sin argumentos y sin applet deducible del nombre**: debe orientar, no fallar en silencio (US2 escenario 2).
- **Bandera de tiempo máximo con valor cero, negativo o sin unidad**: es un error de argumentos (código 2), no un tiempo infinito accidental.
- **Banderas mutuamente excluyentes**: autodescripción junto a una operación real (FR-049: `--describe` excluye la ejecución); salida legible por máquina junto a una petición de ayuda (`--json --help`: la ayuda se emite siempre como texto para personas en la salida estándar, con código 0, y `--json` no la altera; ver FR-042 y `## Clarifications`). Ambas combinaciones tienen un comportamiento definido, idéntico con independencia del orden en que se escriban las banderas.
- **Banderas cuyo objeto todavía no existe** (`--offline`, `--no-graph`, `--asunto`): en H1 no hay caché ni grafo ni asunto. Su comportamiento en este hito debe ser explícito y no engañoso (FR-021, FR-023, FR-024).
- **Contenido de `data` no serializable o con claves duplicadas**: no puede producir un sobre a medias ya escrito en la salida estándar.
- **La salida estándar falla al escribirse** (tubería cerrada): debe traducirse a un código de salida, nunca a un pánico.
- **Un applet que devuelve a la vez un error y contenido parcial**: solo puede haber una respuesta, y el código de salida manda; lo que se emite es el sobre de fallo de FR-045 (`ok: false`, causa en `data`), nunca el contenido parcial con `ok` verdadero.
- **Un applet registrado dos veces o con un nombre que colisiona con una bandera**: debe detectarse al construir el registro, no en la invocación de un usuario.
- **Registro de eventos con la salida de error redirigida a la estándar por quien invoca**: el binario no puede garantizar la separación si quien llama las une; la garantía es que el binario escribe cada cosa en su descriptor.

## Requirements *(mandatory)*

### Requisitos funcionales

#### Despacho multicall (`internal/app`)

- **FR-001**: El proyecto DEBE mantener un **registro de applets** en el que cada applet se declara una sola vez con su nombre, sus verbos y su descripción, y del que se derivan el despacho, la ayuda y la autodescripción, sin ninguna lista paralela mantenida a mano.
- **FR-002**: El binario DEBE seleccionar el applet a ejecutar a partir del nombre con el que se le invoca (`os.Args[0]`) cuando ese nombre corresponde a un applet registrado.
- **FR-003**: El binario DEBE seleccionar el applet a ejecutar a partir de su primer argumento cuando el nombre de invocación es el del propio binario o no corresponde a ningún applet registrado.
- **FR-004**: La precedencia entre ambos mecanismos DEBE estar fijada de forma explícita y ser comprobable con un test: cuando el nombre de invocación identifica un applet, ese applet manda y el primer argumento se entrega íntegro al applet.
- **FR-005**: Un nombre de invocación que no corresponde a ningún applet registrado NO DEBE impedir el uso del binario: se trata como invocación por nombre propio.
- **FR-006**: Una invocación cuyo applet no puede determinarse (sin argumentos y sin nombre reconocible) o cuyo applet no está registrado DEBE terminar con el código de salida 2 y escribir en la salida de error la lista de applets disponibles.
- **FR-007**: El resultado de invocar un applet por enlace simbólico DEBE ser indistinguible del de invocarlo con el applet como primer argumento: misma salida estándar, misma salida de error y mismo código de salida.
- **FR-008**: El registro DEBE rechazar, en el momento de construirse, nombres de applet duplicados o que colisionen con el juego de banderas globales; ese fallo NO DEBE manifestarse como un error de usuario en tiempo de invocación.
- **FR-009**: El applet `echo` es un **ejemplo del patrón, no funcionalidad del producto**, y su definición vive en `internal/app/testdata`. El binario `kitlegal` que se distribuye NO registra `echo`: la entrega literal del hito (`kitlegal echo hola --json`, `ln -s kitlegal echo && ./echo hola`, `--describe`) se demuestra contra un binario compilado por el test e2e (`testscript`) a partir del kernel real (`cmd`/`internal/app`) más el registro del applet de ejemplo. El registro de applets del binario distribuido y el del e2e usan exactamente el mismo mecanismo de registro (FR-001); lo único que cambia es qué applets se registran.

#### Sobre de salida (`internal/core/schema`)

- **FR-010**: Todo applet DEBE emitir su resultado dentro del sobre `{ok, fuente, url, fecha_consulta, hash, data}`, con esas claves en español y sin claves adicionales en el nivel superior.
- **FR-011**: `hash` DEBE ser la huella SHA-256 del contenido de `data` en su forma canónica, de modo que la misma información produzca siempre la misma huella con independencia del orden de las claves, de los espacios de la serialización y del momento de la consulta.
- **FR-012**: La huella DEBE ir precedida del algoritmo que la produjo, de forma que sea posible cambiar de algoritmo en el futuro sin romper a quien la lea.
- **FR-013**: `fecha_consulta` DEBE ser una marca temporal con zona horaria explícita, en formato legible por máquina.
- **FR-014**: `ok` DEBE indicar si la operación tuvo éxito, y su valor DEBE ser coherente con el código de salida del proceso.
- **FR-015**: El sobre DEBE quedar definido en el dominio, en un único lugar del que dependan todos los applets, de modo que ningún applet pueda emitir una forma distinta.
- **FR-016**: `fuente` y `url` DEBEN identificar la procedencia comprobable del contenido y NUNCA ir vacías. Un applet que no consulta ninguna fuente externa (p. ej. `echo` en H1; previsiblemente `cita` en H8 y `plazos` en H9) DEBE usar el espacio de nombres reservado: `fuente` con el prefijo `kitlegal.` (p. ej. `kitlegal.echo`) y `url` con el esquema de URI `kitlegal:` (p. ej. `kitlegal:applet/echo`). El propio kernel usa ese espacio de nombres (`kitlegal.cli` / `kitlegal:cli`) cuando emite un sobre sin haber llegado a consultar nada (FR-045). Ningún adaptador de `internal/source/<fuente>` DEBE usar ese prefijo ni ese esquema; sus valores identifican la fuente pública y una URL http(s) comprobable. Un sobre con procedencia `kitlegal:` es un resultado calculado, no una cita de fuente pública, y así debe tratarlo cualquier consumidor.
- **FR-017**: El sobre emitido DEBE ser validable contra una descripción formal, de manera que un test pueda comprobar que una salida cumple el contrato sin inspeccionar campo a campo; esa validación formal DEBE rechazar un sobre cuyo `url` esté vacío o no sea un URI.

#### Banderas globales (`internal/cli`)

- **FR-018**: El kernel DEBE ofrecer a **todos** los applets, sin que ninguno las declare, las banderas `--json`, `--timeout`, `--offline`, `--dry-run`, `--describe`, `--no-graph`, `--asunto` y `--verbose`, con idéntica sintaxis y semántica en todos ellos.
- **FR-019**: `--json` DEBE seleccionar la salida legible por máquina; en su ausencia se presenta la salida legible por una persona.
- **FR-020**: `--timeout` DEBE limitar la duración total de la operación; agotarlo DEBE traducirse en el código de salida reservado a fuente no disponible. Un valor no positivo o con formato inválido es un error de argumentos.
- **FR-021**: `--offline` DEBE declarar que la operación no puede acceder a la red. En H1 ningún applet accede a la red, por lo que la bandera se acepta y se propaga al applet, sin que su semántica completa (responder solo desde caché) exista hasta H3.
- **FR-022**: `--dry-run` DEBE describir la operación que se habría realizado sin realizarla, y terminar con código 0. La descripción DEBE ir a la salida de error y ser **siempre visible**, con el nivel de registro por omisión y sin depender de `--verbose` ni de `KITLEGAL_LOG`; también cuando se combina con `--json`. La salida estándar queda vacía porque el sobre es una cita de un contenido consultado y una operación no realizada no tiene nada que citar. El kernel NO DEBE cortar la ejecución antes del applet: la bandera viaja en el Contexto de ejecución (FR-018) y cada capa con efectos la honra describiendo en lugar de ejecutar —igual que `--offline`, `--no-graph` y `--asunto` (FR-021, FR-023, FR-024)—; en H1, donde ninguna capa tiene efectos, la descripción (applet, verbo y argumentos que se habrían ejecutado) la escribe el kernel y el presentador no se invoca.
- **FR-023**: `--no-graph` DEBE declarar que la ejecución no altera el grafo. En H1 no existe grafo, por lo que la bandera se acepta y se propaga al applet, sin efecto observable hasta H12.
- **FR-024**: `--asunto` DEBE declarar sobre qué asunto se trabaja. En H1 no existe el grafo del asunto, por lo que la bandera se acepta y se propaga al applet, sin efecto observable hasta H14.
- **FR-025**: `--verbose` DEBE aumentar el detalle del registro de eventos en la salida de error, sin alterar en absoluto la salida estándar.
- **FR-026**: El binario DEBE ofrecer ayuda (`--help`) en el binario y en cada applet, con la lista de applets y verbos derivada del registro, terminando con código 0.
- **FR-027**: Un argumento o bandera desconocido, un valor con formato inválido o la ausencia de un argumento obligatorio DEBEN producir el código de salida 2 y un mensaje en la salida de error que nombre el problema concreto.
- **FR-028**: Las combinaciones de banderas mutuamente excluyentes DEBEN tener un comportamiento definido e independiente del orden de escritura.

#### Errores tipados y códigos de salida (`internal/cli/errors.go`)

- **FR-029**: El kernel DEBE definir un juego de **errores tipados** que los applets devuelven y que el kernel traduce a los códigos de salida estables del proyecto: 0 ok · 2 args · 3 no encontrado · 4 fuente no disponible · 5 rate-limited/TOS · 6 requiere identidad humana.
- **FR-030**: La traducción de error a código de salida DEBE estar en un único lugar y ser exhaustiva: un error de una clase que no se contemple NO DEBE pasar inadvertido.
- **FR-031**: Un error que no corresponda a ninguna clase prevista DEBE traducirse a un código de salida distinto de los reservados y de 0, de modo que un fallo inesperado nunca se confunda con un éxito ni con un fallo clasificado.
- **FR-032**: Un applet DEBE poder añadir contexto a un error sin perder su clase: envolver un error no cambia el código de salida resultante.
- **FR-033**: Ningún camino de usuario DEBE terminar en pánico; los fallos se propagan como error y salen por el único punto que termina el proceso.
- **FR-034**: El código de salida 6 DEBE ser el que señala que la acción requiere identidad humana, y al producirse NO DEBE haberse realizado ninguna acción con efectos externos.
- **FR-035**: Solo el kernel y el punto de entrada DEBEN terminar el proceso; ningún otro paquete lo hace.

#### Observabilidad (registro de eventos)

- **FR-036**: El registro de eventos DEBE escribirse siempre en la salida de error y nunca en la estándar.
- **FR-037**: El nivel de registro DEBE poder fijarse tanto con `--verbose` como con la variable de entorno `KITLEGAL_LOG`.
- **FR-038**: El registro DEBE ser estructurado, de modo que un consumidor automático pueda filtrarlo.
- **FR-039**: El registro NO DEBE incluir, por omisión, contenido que identifique a una persona física.

#### Presentación (`internal/render`)

- **FR-040**: `internal/render` DEBE ser el **único** paquete que escribe en la salida estándar.
- **FR-041**: El presentador DEBE ofrecer en H1 dos formas: la salida legible por máquina (el sobre en JSON) y una tabla mínima legible por una persona.
- **FR-042**: Con `--json`, la salida estándar DEBE contener exactamente un documento JSON y nada más. Esta regla queda acotada a la presentación del resultado de un applet: no rige sobre la ayuda (`--help`, que se emite siempre como texto para personas en la salida estándar, sin que `--json` la altere) ni sobre `--dry-run` (donde no hay resultado que presentar y la salida estándar queda vacía).
- **FR-043**: La tabla mínima DEBE mostrar los cuatro datos de procedencia (`fuente`, `url`, `fecha_consulta`, `hash`) además del contenido, para que la cita no se pierda por elegir el formato legible.
- **FR-044**: Añadir una forma de presentación nueva NO DEBE obligar a modificar ningún applet.
- **FR-045**: La representación de un fallo bajo `--json` DEBE estar definida por el contrato del sobre y no dejarse a cada applet: la salida estándar lleva el mismo sobre de seis claves con `ok: false` —sin claves adicionales ni ausentes en el nivel superior (FR-010)— y la causa (clase de error y mensaje para la persona) va dentro de `data`. Lo emite el kernel, no el applet, desde el único punto que traduce el error a código de salida (FR-030), de modo que la representación es idéntica para todos los applets y para los fallos que ocurren antes de llegar al applet. `ok` es falso si y solo si el código de salida no es 0 (FR-014).

  La **procedencia del sobre de fallo** DEBE cumplir FR-016 y NUNCA ir vacía: `fuente` y `url` son las de la fuente que se estaba consultando cuando se conocen (p. ej. la URL que devolvió «no encontrado», útil como cita negativa) y, cuando no se conocen —fallos anteriores a la ejecución del applet: bandera o argumento desconocido y applet no registrado (FR-006, FR-027)—, las del espacio de nombres reservado de FR-016 referido al propio kernel (p. ej. `fuente: kitlegal.cli`, `url: kitlegal:cli`). `fecha_consulta` y `hash` se calculan como en cualquier otro sobre, con la huella sobre la forma canónica del `data` de error, de modo que el sobre de fallo es validable contra la misma descripción formal que el de éxito (FR-017).

  El mensaje para la persona sigue yendo además a la salida de error; el código de salida sigue siendo la vía primaria de clasificación y el sobre de fallo la duplica en forma estructurada, no la sustituye.

#### Autodescripción (`--describe`)

- **FR-046**: `--describe` DEBE emitir un esquema JSON válido que describa la entrada y la salida del applet invocado.
- **FR-047**: La parte de salida del esquema DEBE describir el sobre completo, con `data` **condicionado a `ok`**: el `data` propio del applet cuando `ok` es verdadero y la forma común de error del kernel (FR-045) cuando `ok` es falso. Una ejecución fallida del applet, emitida según FR-045, DEBE por tanto validar contra el esquema que ese mismo applet emite con `--describe`; un test de contrato sobre una ejecución fallida no puede rechazar una salida conforme.
- **FR-048**: El esquema DEBE derivarse de la definición del applet y de los tipos del sobre, no mantenerse a mano: un cambio en el applet se refleja en su esquema sin edición manual.
- **FR-049**: `--describe` DEBE excluir la ejecución de la operación y terminar con código 0.

#### Reglas de arquitectura ejecutables

- **FR-050**: El proyecto DEBE activar `depguard` con las reglas de dependencia de `docs/ROADMAP.md` §2: el dominio (`internal/core/**`) no importa adaptadores ni kernel; solo `internal/httpx` importa la biblioteca HTTP estándar; solo `internal/{cache,store,graph}` importan SQLite y el acceso a base de datos; solo `internal/cli` y `cmd/` terminan el proceso; solo `internal/render` escribe en la salida estándar.
- **FR-051**: El proyecto DEBE incluir un **test de arquitectura** que recorra el grafo de dependencias real del módulo y falle si alguna de esas reglas se viola, de modo que la garantía no dependa solo de la configuración del lint.
- **FR-052**: Las reglas que en H0 se acotaron por ruta como medida provisional DEBEN quedar sustituidas por las reglas completas de §2, ahora que existen los paquetes a los que se refieren.
- **FR-053**: Un intento de violar cualquiera de las reglas DEBE producir un fallo que nombre la regla violada, no un aviso.

#### Controles y cobertura

- **FR-054**: El hito DEBE incluir tests unitarios de: la traducción de error a código de salida, el sobre (incluida la reproducibilidad de la huella y la forma del sobre de fallo de FR-045), el despacho multicall y la autodescripción.
- **FR-055**: El hito DEBE incluir un test e2e ejecutado con `testscript` contra el binario compilado que cubra: la ayuda, la invocación por enlace simbólico, el código de salida 2 con argumentos inválidos y que la salida legible por máquina es JSON analizable.
- **FR-056**: La cobertura de `internal/cli` DEBE alcanzar al menos el 90 %, y las de `internal/core/**` (≥ 85 %) y global (≥ 70 %) DEBEN seguir cumpliéndose.
- **FR-057**: Todos los controles heredados de H0 —formato, lint, tests con detector de carreras, vulnerabilidades, SAST, secretos, verificación de módulos y dependencias saneadas— DEBEN seguir en verde sin excepciones ni exclusiones nuevas sin justificar.

#### Invariantes del proyecto que H1 no puede violar

- **FR-058**: El código introducido en H1 NO DEBE realizar ninguna petición de red.
- **FR-059**: El código introducido en H1 NO DEBE escribir en disco fuera del directorio de construcción y de los directorios temporales de los tests.
- **FR-060**: Las dependencias nuevas DEBEN limitarse a las ya fijadas por la constitución para este hito; cualquier otra exige justificación explícita.

### Criterios del hito, literales

Se transcriben aquí, sin alterar, los apartados «Controles» y «Aceptación» de `#### H1 · Kernel CLI: multicall, flags globales, exit codes, sobre de salida` en `docs/ROADMAP.md`:

- **Controles**: unit (mapeo error→exit, sobre, dispatch, `--describe`); e2e testscript (`--help`, symlink, exit 2 con args malos, salida JSON parseable); test de arquitectura; `depguard` activo.
- **Aceptación**: cobertura `internal/cli` ≥ 90 %; el e2e demuestra que stdout solo contiene JSON cuando `--json`.

La transcripción de «Controles» se desarrolla en FR-050 a FR-055; la de «Aceptación», en SC-004 y SC-002.

### Key Entities

- **Applet**: una capacidad invocable del binario, identificada por un nombre (`echo`, y más adelante `boe`, `cita`, `plazos`…) y un conjunto de verbos. Declara qué entrada acepta y qué contenido devuelve; no declara banderas, ni sobre, ni códigos de salida, ni forma de presentación.
- **Registro de applets**: la colección de applets disponibles en el binario. Única fuente de la que se derivan el despacho, la ayuda y la autodescripción.
- **Sobre de salida**: la estructura `{ok, fuente, url, fecha_consulta, hash, data}` que envuelve el resultado de cualquier applet, tanto si la operación tuvo éxito como si falló (FR-045). `data` es lo único que varía: el contenido propio del applet cuando `ok` es verdadero y la forma común de error del kernel cuando es falso.
- **Error tipado**: un fallo clasificado en una de las clases que el proyecto reconoce (argumentos, no encontrado, fuente no disponible, límite o términos de uso, identidad humana), más la clase residual «inesperado» de FR-031 para lo que no encaja en ninguna. El kernel traduce esa clase a un código de salida estable y, bajo `--json`, la declara dentro del `data` del sobre de fallo (FR-045).
- **Contexto de ejecución**: el conjunto de decisiones globales que el kernel entrega al applet —formato de salida, tiempo máximo, trabajo sin red, ejecución simulada, uso del grafo, asunto y nivel de detalle— sin que el applet tenga que interpretarlas.

## Success Criteria *(mandatory)*

### Resultados medibles

- **SC-001**: `kitlegal echo hola --json`, ejecutado contra el binario que compila el test e2e —el kernel real más el registro del applet de ejemplo, no el artefacto de release, que no registra `echo` (FR-009)—, emite en la salida estándar un único documento JSON con exactamente las seis claves del sobre, `ok` verdadero y `data` conteniendo lo pedido, y termina con código de salida 0.
- **SC-002**: El test e2e demuestra que, cuando se pide `--json`, la salida estándar contiene **solo** JSON: el documento capturado se analiza por completo y no queda ningún texto fuera de él, ni antes ni después, con el registro de eventos activo al máximo detalle. *(criterio de aceptación literal del hito)*
- **SC-003**: Un enlace simbólico con el nombre de un applet registrado produce salida estándar, salida de error y código de salida byte a byte idénticos a los de la invocación con el applet como primer argumento (salvo los datos que dependen del instante de la consulta). Se verifica —igual que SC-001— contra el binario que compila el test e2e, único que registra el applet de ejemplo (FR-009); sobre el binario distribuido, `ln -s kitlegal echo && ./echo hola` termina en el código de salida 2 por applet desconocido, como cualquier otro nombre no registrado (FR-005, FR-006).
- **SC-004**: La cobertura de `internal/cli` es ≥ 90 %, medida por el mismo control de cobertura que ya usa el proyecto y sin exclusiones añadidas. *(criterio de aceptación literal del hito)*
- **SC-005**: Para el mismo contenido de `data`, invocaciones repetidas producen huellas idénticas; para contenidos que difieren en un solo byte, huellas distintas. Ambas propiedades están cubiertas por tests.
- **SC-006**: Existe, para cada uno de los cinco códigos de salida de error (2, 3, 4, 5, 6) y para el 0, al menos un caso de prueba que lo fuerza y lo comprueba; ninguna clase de error queda sin caso.
- **SC-007**: `--describe` sobre un applet registrado —el de ejemplo, contra el binario que compila el e2e (FR-009)— emite un esquema JSON que un validador de esquemas acepta como válido y que describe entrada y salida, incluido el sobre con `data` condicionado a `ok` (FR-047).
- **SC-008**: Las cinco reglas de dependencia de `docs/ROADMAP.md` §2 están activas y son demostrables: para cada una existe una comprobación que falla si la regla se retira o se viola, tanto en el lint como en el test de arquitectura.
- **SC-009**: El test e2e cubre los cuatro casos que exige el hito —ayuda, enlace simbólico, código 2 con argumentos inválidos y salida JSON analizable— ejecutándose contra el binario compilado, sin red y sin escribir fuera de su directorio temporal.
- **SC-010**: Registrar un applet nuevo que solo declara nombre, verbos y contenido de `data` basta para que herede las ocho banderas globales, el sobre, la traducción de errores a códigos de salida y las dos formas de presentación; se demuestra con dos applets de ejemplo distintos, sin código duplicado entre ellos.
- **SC-011**: Ejecutando cualquier applet con la salida estándar y la de error capturadas por separado, la estándar contiene exclusivamente lo que emite el presentador y la de error exclusivamente el registro de eventos y los mensajes dirigidos a la persona.
- **SC-012**: Todos los paquetes enumerados en el «Alcance» de H1 (`internal/cli`, `internal/app`, `internal/core/schema`, `internal/render`) existen, cumplen sus requisitos funcionales y ninguno queda como marcador de posición sin contenido.
- **SC-013**: La orden agregada de integración continua del proyecto termina en verde con todos los controles de H0 más los que añade H1, sin haber modificado el árbol de trabajo.
- **SC-014**: Para cada una de las cinco clases de error (códigos 2, 3, 4, 5 y 6) y para el fallo inesperado de FR-031, existe un caso de prueba que fuerza el fallo con la bandera de salida legible por máquina y comprueba que la salida estándar contiene un único documento JSON con exactamente las seis claves del sobre, `ok` falso, la clase del error y el mensaje para la persona dentro de `data`, `fuente` y `url` no vacías —las de la fuente consultada cuando se conocen y las del espacio de nombres reservado del kernel cuando no—, y el código de salida correspondiente a esa clase. Al menos dos de esos casos son fallos anteriores a la ejecución del applet (bandera desconocida y applet no registrado).
- **SC-015**: El sobre de fallo de SC-014 valida contra la descripción formal del sobre (FR-017) y contra el esquema que el propio applet emite con `--describe` (FR-047); ninguna de las dos validaciones rechaza una salida conforme a FR-045.

## Assumptions

- **Los «usuarios» de este hito siguen siendo quienes contribuyen** (personas y agentes) y, por primera vez, quien escribirá los applets posteriores. H1 no entrega valor a un usuario final del producto legal; ese valor llega en H4.
- **El alcance de `--describe` en H1 es el literal del hito**: que cada applet emita el esquema JSON de su entrada y su salida. La generación del directorio `schemas/`, la comprobación de deriva en integración continua y la generación de las tablas de comandos de las skills y de las herramientas MCP son alcance de H11 y H22 y no se anticipan (constitución, *Criterio de decisión autónoma* §2).
- **`--offline`, `--no-graph` y `--asunto` se declaran y se propagan en H1 pero no tienen objeto todavía**: la caché entra en H3, el grafo en H12 y el grafo del asunto en H14. H1 fija su sintaxis y su presencia en el juego común; no inventa una semántica que ningún hito ha definido.
- **El applet `echo` es el ejemplo mínimo del patrón**, no una funcionalidad del producto: no accede a la red, no toca disco y su `data` se deriva de sus argumentos. Su ubicación (`internal/app/testdata`) la fija el hito; no se registra en el binario distribuido (FR-009).
- **Un segundo applet de ejemplo** es necesario para demostrar SC-010 (que el patrón no está cableado a un caso único). También vive solo en material de test y tampoco es funcionalidad del producto.
- **Los códigos de salida 0, 2, 3, 4, 5 y 6 son una decisión cerrada** (`CLAUDE.md`) y no se reabren. El valor concreto que FR-031 reserva a los fallos inesperados es una decisión del plan, acotada a «ni 0 ni ninguno de los reservados».
- **El formato de la tabla mínima** (qué columnas, cómo se aplanan los contenidos anidados) lo decide el plan: el hito solo exige que exista, que sea legible y que conserve los cuatro datos de procedencia.
- **El valor por omisión de `--timeout`** lo decide el plan; el hito exige que exista un límite aplicado a toda la operación y que agotarlo produzca el código 4.
- **La forma canónica de `data` para calcular la huella** (ordenación de claves, normalización de espacios y de números) la decide el plan; el hito exige que la huella sea reproducible y que dependa solo del contenido.
- **`cmd/kitlegal` deja de escribir directamente en la salida estándar** en H1: la excepción que H0 documentó para `cmd/` desaparece cuando existe `internal/render`, que pasa a ser el único escritor. El verbo `version` de H0 se mantiene y se adapta al patrón en la medida en que lo exijan las reglas de arquitectura; no se le añade sobre ni banderas si el hito no lo pide.
- **Las dependencias que H1 incorpora** son las ya fijadas por la constitución para este hito: el analizador de línea de órdenes (`alecthomas/kong`), la biblioteca de aserciones de test (`stretchr/testify`), el ejecutor de tests e2e (`rogpeppe/go-internal`), el generador de esquemas (`invopop/jsonschema`) y el validador de esquemas en tests (`santhosh-tekuri/jsonschema`). Ninguna otra entra sin justificación explícita.
- **No hay red en ningún test de H1**: no hay fuentes externas todavía, de modo que no hay fixtures que grabar ni filas que añadir a `docs/SOURCES.md`.
- **El hito cambia decisiones de arquitectura ya registradas** solo en la medida en que las materializa; si al implementar aparece una decisión nueva de arquitectura, la *Definition of Done* §1.7 obliga a un ADR, que decide el plan.

## Dependencias

- **Depende de H0** y solo de él: el módulo, el binario, el `Makefile`, la configuración de lint, los flujos de integración continua y los umbrales de cobertura ya existen y H1 los amplía, no los reemplaza.
- **No depende de ninguna fuente externa** ni de ninguna red: H1 es íntegramente local y offline.
- **Bloquea a todos los hitos posteriores**: H2 (cliente HTTP) necesita los errores tipados; H3 (caché) necesita `--offline`; H4 (applet `boe`) necesita el registro, el sobre y el renderizado; H5 (skill migrada) necesita el enlace simbólico; H11, H12 y H22 necesitan `--describe`, `--no-graph` y `--asunto`.

## Fuera de alcance

Todo lo siguiente está definido en `docs/ROADMAP.md`, `CLAUDE.md`, `refs/` o la constitución, pero **no** pertenece al alcance de H1 y **no se implementa** en este hito:

- **Cliente HTTP (H2)**: `internal/httpx`, User-Agent identificable, `robots.txt`, límite de peticiones por host, reintentos con retroceso exponencial, grabación y reproducción de fixtures (`KITLEGAL_RECORD=1`). H1 define los errores tipados «fuente no disponible» y «límite o términos de uso», pero ningún código que los produzca a partir de una respuesta real.
- **Caché y almacenamiento (H3, H12, H16)**: `internal/cache`, `internal/store`, `internal/graph`, SQLite, migraciones, TTL, `KITLEGAL_CACHE_DIR`. `--offline` se declara pero no consulta caché alguna; `--no-graph` se declara pero no hay grafo que desactivar, ni objeto nulo que lo descarte, ni `Emit(ctx) []GraphOp` en los applets.
- **Grafo del asunto (H14)**: `.kitlegal/case.db`, `config.yaml` del asunto, nodo `Consulta`, `graph check`. `--asunto` se declara pero no abre ni crea nada.
- **Adaptadores de fuentes (H4 en adelante)**: `internal/source/<fuente>`, la interfaz `Source` con `Name`, `Fetch`, `TTL` y `Terms`, el applet `boe` y cualquier otro applet de producto, el enlace `boe -> kitlegal`, `docs/SOURCES.md` y `scripts/verify-sources.sh`.
- **El applet `echo` en el binario distribuido**: `echo` no se registra en el `kitlegal` que se publica; vive solo en `internal/app/testdata` y se registra únicamente en el binario que compila el test e2e (FR-009).
- **Dominio legal (H7, H8, H9)**: `internal/core/ids`, `internal/core/cita`, `internal/core/plazos`, `internal/core/competencia`. De `internal/core` H1 solo crea `schema`.
- **Esquemas versionados y contratos (H4 borrador, H11)**: el directorio `schemas/*.json`, su generación desde `--describe`, la comprobación de deriva en integración continua y la orden `schema-check` con contenido real. H1 emite el esquema por applet; no lo persiste ni lo compara con nada commiteado.
- **Skills, packs y datos (H5)**: `skills/`, `packs/`, `data/*.yaml`, `scripts/skills-sync.sh`, las `references/` generadas, la tabla de comandos de cada `SKILL.md` derivada de `--describe` y `evals/`.
- **Servidor MCP (H22)**: `kitlegal mcp serve --stdio` y la generación de herramientas MCP desde `--describe`.
- **API pública (H24)**: `pkg/legalkit` y el control de compatibilidad de API.
- **Extracción de `internal/cli` a una librería**: prohibida explícitamente por la constitución y por `docs/ROADMAP.md` §5; la revisión de si `internal/cli` pide refactor interno —solo refactor, nunca extracción— llega cuando se cumple la regla de los tres applets (`docs/ROADMAP.md` §5), es decir, cuando existan `boe`, `cita` y `plazos`; `echo` no cuenta, porque no se registra en el binario distribuido (FR-009).
- **Presentación en markdown (H11)**: H1 entrega solo las dos formas que el hito nombra, salida legible por máquina y tabla mínima.
- **Release firmado (H6)**: `.goreleaser.yaml`, artefactos multiplataforma, SBOM, firma, `install.sh`.
- **Fuzzing, property tests y benchmarks** (H7, H8, H9, H16): H1 no tiene parsers de dominio que fuzzear.
- **Tests de integración con base de datos** (H3): no hay base de datos en H1.
- **Capa de acción y generación de escritos (H25)**: `internal/docgen`, `Afirmacion`, `fundamenta`. H1 define el código de salida 6, no lo que lo produce.
- **Cualquier semántica de `--offline`, `--no-graph` o `--asunto` más allá de aceptarlas y propagarlas**, y cualquier bandera global fuera de las ocho que el hito enumera.
- **Rediseño del verbo `version` de H0** (formato de salida, sobre, banderas) más allá de lo que exijan las reglas de arquitectura que H1 activa.
- **Protección de rama, revisores obligatorios y cualquier otro ajuste de la plataforma de alojamiento**: siguen fuera de alcance, como en H0.
