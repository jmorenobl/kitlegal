# Feature Specification: H3 · `internal/cache`: SQLite con TTL y `--offline`

**Feature Branch**: `h3-internal-cache-sqlite`

**Created**: 2026-09-12

**Status**: Draft

**Input**: Sección `#### H3 · internal/cache: SQLite con TTL y --offline` de `docs/ROADMAP.md` (modo desatendido).

## Resumen

H2 dejó la única puerta del módulo hacia la red: un cliente que se identifica, lee `robots.txt`, respeta el ritmo de cada sitio y no se puede usar mal. Lo que todavía no existe es **la memoria de lo que ya se pidió**. Las normas consolidadas del BOE cambian poco —una ley se modifica unas cuantas veces al año, no cada minuto—, y sin memoria cada consulta de una skill volvería a molestar a un servidor público para recibir exactamente el mismo texto. H3 construye esa memoria: una caché local en SQLite con vigencia por entrada.

El hito tiene tres caras, y las tres se pueden comprobar sin red:

1. **No volver a pedir lo que ya tenemos.** Quien consulta guarda lo que la fuente respondió, con una vigencia (TTL), y la segunda consulta idéntica se sirve de lo guardado. El criterio de aceptación del hito es exactamente eso, y se comprueba de la forma más dura posible: con un cliente de reproducción **estricto**, construido sobre un directorio que no tiene la grabación de esa petición, de modo que cualquier petición que se emitiera haría fallar el test en lugar de pasar inadvertida.

2. **Lo caducado no se sirve.** La vigencia no es un adorno: una entrada cuyo TTL ha pasado se trata como si no estuviera, y eso se comprueba con un reloj inyectado —el test no espera tiempo real— porque una caché que sirve texto legal caducado sin decirlo es peor que no tener caché (constitución §II: nada sin cita ni fuente vigente).

3. **Trabajar sin red con lo que ya hay.** `--offline` es una bandera que H1 declaró y propagó y que H2 dejó deliberadamente sin significado. Aquí lo adquiere: quien construye la caché convierte esa bandera en un cliente que **solo lee** —no escribe ninguna entrada, no crea la base de datos, no aplica ninguna migración; la caché misma no mira ninguna bandera— y lo que falta no se inventa: la invocación termina con el código de salida 4, que es lo que un agente en un entorno aislado necesita para distinguir «no lo tengo» de «no existe».

Por debajo, todo lo que toca el disco se hace con el cuidado que exige un fichero que va a sobrevivir a muchas versiones del binario: `modernc.org/sqlite` sin cgo (ADR 0002), migraciones embebidas en el propio binario con una tabla `schema_version`, WAL para que leer y escribir a la vez no sea un problema, y `PRAGMA` que no sacrifican integridad por velocidad. El puerto `Cache` se define en `internal/core`, así que el dominio sigue sin saber que hay SQLite detrás (constitución §IV, regla R3), y la caché es el **primer dueño real** de esa regla, hoy activa y sin ninguno.

H3 no entrega ningún comando nuevo ni mejora ninguna skill: es, como H0, H1 y H2, un hito de **fundación** de los que el principio VIII admite porque protegen a las skills que vendrán. Al cerrarlo, la superficie visible del binario es exactamente la misma que al abrirlo; el primer applet que use la caché es `boe`, en H4.

## Criterios del hito, literales

Transcripción literal de `docs/ROADMAP.md` §4, H3, para que la trazabilidad del spec se pueda comprobar sin volver al roadmap.

- **Objetivo**: «las normas cambian poco; no volver a pedir lo que ya tenemos.»
- **Entrega**: «`cache.Get/Put(key, ttl)` en `~/.cache/kitlegal/cache.db` (ruta configurable con `KITLEGAL_CACHE_DIR`); con `--offline` solo lee, y devuelve exit 4 si falta.»
- **Alcance**: «`modernc.org/sqlite`, migraciones embebidas (`embed` + tabla `schema_version`), WAL, `PRAGMA` seguros. Interfaz `Cache` definida en `core`.»
- **Controles**: «integración (`//go:build integration`) con `t.TempDir()`; unit de expiración de TTL con reloj inyectado; `sqlclosecheck`.»
- **Aceptación**: «segunda llamada idéntica no toca la red (se verifica con `Replay` en modo estricto que falla ante cualquier petición).»

Trazabilidad resumida (el detalle, en cada requisito y criterio):

| Criterio literal | Dónde se cumple |
|---|---|
| `cache.Get/Put(key, ttl)` | FR-001, FR-006, FR-008, FR-013, US1 escenarios 1 y 2 |
| `~/.cache/kitlegal/cache.db` | FR-019, FR-021, SC-004, US4 escenario 1 |
| ruta configurable con `KITLEGAL_CACHE_DIR` | FR-020, FR-022, FR-023, SC-004, US4 escenarios 2 y 3 |
| con `--offline` solo lee | FR-015, FR-017, FR-018, SC-003, US3 escenarios 1, 3 y 6 |
| devuelve exit 4 si falta | FR-016, FR-022, FR-033, SC-003, SC-011, US3 escenarios 2 y 3 |
| `modernc.org/sqlite` | FR-002, FR-036, FR-043, Assumptions |
| migraciones embebidas (`embed` + tabla `schema_version`) | FR-024 a FR-029, SC-005, US5 escenarios 1 a 3 |
| WAL | FR-030, SC-006, US5 escenario 4, US3 escenario 6 |
| `PRAGMA` seguros | FR-031, FR-032, SC-006, SC-007, US5 escenario 4, US3 escenario 6 |
| Interfaz `Cache` definida en `core` | FR-001, FR-005, FR-038, SC-013, US6 escenarios 1 y 2 |
| integración (`//go:build integration`) con `t.TempDir()` | FR-040, FR-041, SC-009, US5 escenario 5 |
| unit de expiración de TTL con reloj inyectado | FR-009, FR-042, SC-002, US2 |
| `sqlclosecheck` | FR-004, FR-039, SC-008, US6 escenario 3 |
| segunda llamada idéntica no toca la red | FR-045, FR-046, SC-001, US1 |
| se verifica con `Replay` en modo estricto que falla ante cualquier petición | FR-046, SC-001, US1 escenario 3 |

## Clarifications

### Session 2026-09-12

- Q: ¿Cómo recibe la caché el modo de solo lectura que impone `--offline` y qué forma tiene entonces la escritura: una misma superficie que falla al escribir en tiempo de ejecución, o una superficie que hace la escritura imposible? (FR-015, FR-017, SC-011) → A: opción funcional del constructor (p. ej. `cache.SoloLectura()`; el nombre lo fija el plan); un único tipo `Cache` con `Get`/`Put`; **la caché no lee ninguna bandera**: quien la construye —en H3 el applet mínimo y el adaptador de prueba, desde H4 el adaptador de fuente o la composición— toma el modo de `schema.Contexto.Offline`, que H1 entrega ya interpretado al applet, y lo pasa como opción del constructor, porque el modo debe conocerse al abrir (crear la base y migrar ocurre al abrir, no al operar); en modo de solo lectura `Put` no escribe nada y falla en ejecución con clase «inesperado» (código 1), acreditado con un test que llama a `Put` a través del kernel; impedirlo además por construcción es refuerzo opcional del plan, nunca sustituto del fallo en ejecución. (auto: criterio c; fuente: FR-015, FR-017, FR-033, FR-047, SC-011; Key Entities «Cliente de caché»; internal/core/schema/contexto.go; docs/ROADMAP.md §2 y §4)
- Q: En modo de solo lectura, ¿qué cuenta como que «el directorio de caché queda exactamente como estaba» respecto de los ficheros auxiliares que SQLite necesita para leer una base en WAL (`cache.db-wal`, `cache.db-shm`), y cómo debe abrirse la base para respetarlo? (FR-015, FR-030, SC-003) → A: «idéntico» se refiere al contenido de `cache.db`, no a los auxiliares `-wal`/`-shm`, que SC-003 excluye de la comparación; si `cache.db` no existe no se abre nada; si existe, se abre en modo de solo lectura **sin desactivar WAL y sin marcar la base como inmutable**, para que la lectura vea las transacciones confirmadas que aún viven en el WAL y respete los bloqueos de una invocación normal que esté escribiendo a la vez (comportamiento observable en FR-015, FR-032, US3 escenario 6 y SC-007); si existe pero no se puede leer, el fallo es «inesperado» (código 1) y nunca se degrada a ausencia (código 4). Cómo se abre exactamente, y si se añade `query_only`, es detalle del plan. (auto: criterio c; fuente: FR-015, FR-016, FR-030, FR-032, FR-033, SC-003, SC-007; US3 escenarios 1, 3, 4 y 6; docs/ADR/0002)
- Q: ¿Cómo distingue el puerto del dominio los tres resultados de una lectura —presente y vigente, ausente (incluida la expirada) y fallo— sin convertir la ausencia en un error? (FR-001, FR-013) → A: `Get(ctx, clave) (contenido []byte, presente bool, err error)`, idioma «coma ok» de Go; ausencia es `presente=false, err=nil`; fallo es `err != nil` con `schema.ConClase`; ningún error centinela para la ausencia. (auto: criterio c; fuente: FR-001, FR-012, FR-013, FR-033; internal/core/schema/error.go; docs/ROADMAP.md §2)
- Q: ¿Dónde vive el adaptador de prueba que demuestra el criterio de aceptación, junto con el applet que el kernel invoca en proceso para el camino de `--offline`, y queda disponible para que H4 lo reutilice? (FR-044, FR-045, FR-047) → A: en ficheros `_test.go` del paquete externo `cache_test`, dentro de `internal/cache`; nada queda disponible fuera de ese paquete; H4 escribe su propio adaptador real y sus propios fixtures. (auto: criterio a; fuente: FR-044, FR-045, FR-047, SC-012; internal/httpx/adaptador_test.go, specs/003 FR-060 a FR-062; docs/ADR/0010)
- Q: ¿Cómo se acredita el escenario de dos procesos distintos trabajando a la vez sobre la misma base de datos, dado que este hito no añade ningún binario ni registra el adaptador de prueba en ninguno? (FR-032, FR-044, SC-007) → A: el propio binario de test se relanza a sí mismo como segundo proceso (`exec.CommandContext(ctx, os.Args[0], ...)` con una variable de entorno que fija el papel de escritor o lector), sin añadir ningún `package main` nuevo. (auto: criterio c; fuente: FR-032, FR-040, FR-041, FR-044, SC-007, SC-012; constitución §V (`rogpeppe/go-internal`); docs/ROADMAP.md §3)

## User Scenarios & Testing *(mandatory)*

Como en H2, las personas y procesos a los que sirve este hito no son las usuarias finales del producto: son quien escribirá el primer adaptador de fuente (H4 y siguientes), quien mantiene la suite de tests, quien ejecuta un agente en un entorno sin salida a Internet y —al otro lado del cable— el servidor público al que no volvemos a pedir lo que ya nos dio.

### User Story 1 - No volver a pedir lo que ya tenemos (Priority: P1)

Quien escribe un adaptador consulta un recurso de una fuente pública y guarda lo que le respondieron con una vigencia. La segunda consulta idéntica, dentro de esa vigencia, se responde con lo guardado: la fuente no recibe ninguna petición y quien llama recibe lo mismo que la primera vez.

**Why this priority**: es el objetivo literal del hito («no volver a pedir lo que ya tenemos») y su criterio de aceptación. Sin esta historia no hay hito.

**Independent Test**: con un adaptador de prueba que combina la caché con el cliente de H2 en modo reproducción y un directorio de caché temporal: se pide dos veces lo mismo y se cuenta cuántas peticiones se emitieron.

**Acceptance Scenarios**:

1. **Given** una caché vacía en un directorio temporal y un recurso cuya respuesta está disponible para la primera consulta, **When** se pide ese recurso por primera vez, **Then** la consulta llega a la fuente, el resultado se entrega a quien llamó y queda guardado con su vigencia.
2. **Given** la misma caché justo después, **When** se pide **el mismo** recurso con la misma clave y dentro de la vigencia, **Then** el resultado se entrega desde lo guardado, íntegro e idéntico byte a byte al de la primera consulta, y **no** se emite ninguna petición.
3. **Given** ese segundo intento, **When** se ejecuta contra un cliente de reproducción **estricto** —construido sobre un directorio que no contiene la grabación de esa petición, de modo que cualquier petición emitida termina en un fallo ruidoso—, **Then** el test pasa, y pasa precisamente porque no se emitió ninguna petición; si la caché no sirviera, el test fallaría nombrando la petición que se intentó.
4. **Given** dos claves distintas, **When** se piden las dos, **Then** cada una se guarda y se sirve por separado: guardar una no responde por la otra y borrar o caducar una no afecta a la otra.
5. **Given** una clave ya guardada, **When** se vuelve a guardar la misma clave con otro valor y otra vigencia, **Then** la entrada anterior queda sustituida por completo —valor y vigencia— y la lectura siguiente devuelve la nueva.

---

### User Story 2 - Lo caducado no se sirve (Priority: P1)

Quien consulta una norma no recibe nunca texto guardado cuya vigencia ha pasado. La caché no se lo entrega «por si acaso» ni con un aviso: para quien pregunta, una entrada caducada es una entrada que no está, y la consulta vuelve a la fuente.

**Why this priority**: es la otra mitad del objetivo del hito. Una caché sin vigencia comprobable convertiría al proyecto en una fuente de citas obsoletas, que es lo que la constitución §II prohíbe de raíz. El control literal del hito («unit de expiración de TTL con reloj inyectado») es de esta historia.

**Independent Test**: con un reloj inyectado que el test adelanta a voluntad, sin esperar tiempo real: se guarda una entrada con una vigencia corta, se adelanta el reloj más allá de ella y se lee.

**Acceptance Scenarios**:

1. **Given** una entrada guardada con una vigencia y un reloj que no ha avanzado, **When** se lee, **Then** se entrega la entrada.
2. **Given** la misma entrada y el reloj adelantado **justo hasta** el instante en que la vigencia termina, **When** se lee, **Then** el resultado es ausencia: el borde lo fija FR-008 —vigente mientras el reloj es anterior al instante de expiración— y no depende de la implementación ni del azar.
3. **Given** la misma entrada y el reloj adelantado más allá de la vigencia, **When** se lee, **Then** el resultado es ausencia: no se entrega ningún valor, ni acompañado de aviso, y quien llama queda libre de volver a la fuente.
4. **Given** una entrada caducada, **When** se vuelve a guardar la misma clave, **Then** la nueva entrada es vigente y la caducada no reaparece nunca.
5. **Given** el test completo de expiración, **When** se ejecuta, **Then** no espera tiempo real: su duración no depende de la vigencia usada y la suite no se alarga por comprobar caducidades.

---

### User Story 3 - Trabajar sin red con lo que ya hay (Priority: P1)

Quien ejecuta un agente en un entorno sin salida a Internet —un contenedor aislado, una máquina en un tren— invoca con `--offline`: lo que esté guardado y vigente se responde, y lo que falte se declara con un código de salida estable en lugar de intentar salir a la red o devolver un vacío que parezca una respuesta.

**Why this priority**: es la mitad de la entrega literal del hito («con `--offline` solo lee, y devuelve exit 4 si falta») y la primera vez que esa bandera significa algo. Comparte prioridad con las dos anteriores porque es la razón por la que la caché existe también fuera del camino feliz.

**Independent Test**: con una caché preparada en un directorio temporal y el kernel invocado en proceso sobre un adaptador de prueba, comprobando el resultado, el código de salida y que el fichero de la caché no se modifica.

**Acceptance Scenarios**:

1. **Given** una caché que contiene la entrada pedida y vigente, y la invocación con `--offline`, **When** se consulta, **Then** se responde desde lo guardado, no se emite ninguna petición y el código de salida es 0.
2. **Given** una caché que **no** contiene la entrada pedida —o la contiene caducada—, y la invocación con `--offline`, **When** se consulta, **Then** no se emite ninguna petición, el fallo es de la clase «fuente no disponible» y el código de salida es **4**.
3. **Given** la misma invocación con `--offline` y un directorio de caché **sin** base de datos —o sin el directorio de caché siquiera, o con el directorio montado de solo lectura—, **When** se consulta, **Then** el resultado es el del escenario anterior (código 4, nunca el 2 de una ruta inservible) **y** el sitio queda como estaba: no se crea el directorio, no se crea la base de datos, no se aplica ninguna migración y no se escribe ninguna entrada (FR-015, FR-022).
4. **Given** una invocación con `--offline` y una caché existente, **When** termina, **Then** el fichero de la base de datos es idéntico al de antes de la invocación: ni una entrada nueva, ni una vigencia refrescada, ni un cambio de esquema.
5. **Given** una invocación **sin** `--offline`, **When** consulta algo que no está guardado, **Then** la ausencia no es un fallo: la consulta sigue su curso hacia la fuente y el código de salida no cambia por el hecho de que la caché no tuviera nada.
6. **Given** una invocación normal que ya ha **confirmado** una entrada sobre la misma base de datos y sigue trabajando, con otra escritura en curso, y una invocación con `--offline` que consulta a la vez, **When** se consulta la entrada ya confirmada, **Then** se responde con código 0 aunque todavía no esté consolidada en el fichero principal —leer en solo lectura nunca la declara ausente (código 4) por serlo—, y mientras tanto la lectura respeta los bloqueos de quien escribe: no falla por bloqueo y de la escritura en curso ve la entrada completa o ninguna, nunca una a medias (FR-015, FR-032).

---

### User Story 4 - La caché vive donde la persona usuaria decide (Priority: P2)

Quien usa el binario tiene su caché en el directorio de caché de su cuenta, sin haber configurado nada, y quien necesita otro sitio —un test, un contenedor con el disco montado en otro lugar, dos asuntos separados— lo declara con una variable de entorno y el binario la respeta o falla diciendo por qué.

**Why this priority**: la entrega del hito nombra la ruta y la variable. Va después de las tres primeras porque es cómo se encuentra la caché, no qué hace.

**Independent Test**: con la variable de entorno apuntando a un directorio temporal y, en el caso por omisión, con un directorio base también temporal, comprobando dónde aparece el fichero y que no aparece en ningún otro sitio.

**Acceptance Scenarios**:

1. **Given** ninguna configuración, **When** se construye la caché por primera vez, **Then** la base de datos aparece en la ruta por omisión que fija FR-019, con el nombre `cache.db`, creando el directorio si no existía y con acceso reservado a la cuenta que la crea.
2. **Given** `KITLEGAL_CACHE_DIR` apuntando a un directorio, **When** se construye la caché, **Then** la base de datos aparece dentro de ese directorio y **en ningún otro sitio**: la ruta por omisión no se toca.
3. **Given** `KITLEGAL_CACHE_DIR` con un valor inservible —cadena vacía o una ruta que es un fichero y no un directorio en cualquier modo; un directorio que no se puede crear ni escribir, fuera del modo de solo lectura—, **When** se construye la caché, **Then** la construcción falla de forma explícita nombrando la variable y la ruta, con la clase «argumentos» (código 2), y **no** se cae en silencio a la ruta por omisión (FR-022).
4. **Given** un test, **When** necesita una caché propia, **Then** puede declarar el directorio como opción explícita del constructor, sin tocar el entorno del proceso, y esa opción tiene precedencia sobre la variable (FR-023).

---

### User Story 5 - La base de datos sobrevive a las versiones y a las invocaciones simultáneas (Priority: P2)

Quien actualiza el binario no pierde su caché ni se queda con una base de datos a medias, y dos invocaciones que coinciden en el tiempo —un agente que lanza dos consultas— no se estorban ni corrompen el fichero.

**Why this priority**: es el alcance literal del hito (migraciones embebidas, `schema_version`, WAL, `PRAGMA` seguros). Va después de las historias de comportamiento porque es lo que las hace duraderas.

**Independent Test**: tests de integración con la base de datos en `t.TempDir()`: abrir dos veces, abrir una base con una versión de esquema desconocida, abrir un fichero que no es una base de datos, y leer y escribir desde varios clientes a la vez.

**Acceptance Scenarios**:

1. **Given** un directorio sin base de datos, **When** se construye la caché, **Then** la base se crea con el esquema al día y la versión aplicada queda registrada en `schema_version`.
2. **Given** una base de datos ya migrada, **When** se vuelve a construir la caché sobre ella, **Then** no se aplica nada de nuevo, no se produce ningún fallo y la versión registrada es la misma.
3. **Given** una base de datos cuya versión de esquema es **mayor** que la que el binario conoce —la dejó un binario más nuevo—, **When** se construye la caché, **Then** falla de forma explícita diciendo qué versión esperaba y cuál encontró, y el fichero **no** se modifica ni se borra.
4. **Given** dos clientes de caché sobre el mismo fichero, uno escribiendo y otro leyendo a la vez, **When** trabajan, **Then** ninguno falla por bloqueo, la base no se corrompe y lo que el lector ve es una entrada completa o ninguna, nunca a medias.
5. **Given** los tests de integración de este hito, **When** se ejecutan, **Then** trabajan exclusivamente dentro del directorio temporal que el propio test crea, no leen ni escriben la caché real de la cuenta, y al terminar no queda ningún fichero fuera de ese directorio.

---

### User Story 6 - El dominio no sabe que hay SQLite (Priority: P2)

Quien escriba mañana un adaptador de fuente, o el propio dominio, trabaja con la idea de «caché» y no con una base de datos: la interfaz vive en el dominio, la implementación en el adaptador, y ningún paquete fuera de los tres de almacenamiento puede abrir una base de datos aunque quiera.

**Why this priority**: es el alcance literal («Interfaz `Cache` definida en `core`») y lo que convierte la regla R3, hoy activa y sin dueño, en una regla con dueño real. Sin esto, la caché sería un atajo que se propaga.

**Independent Test**: los controles mecánicos de `make ci` —lint con `depguard` y `sqlclosecheck`, y el test de arquitectura sobre el grafo transitivo real— ejecutados sobre un intento de saltarse la regla.

**Acceptance Scenarios**:

1. **Given** el puerto de caché definido en el dominio, **When** se revisa lo que el dominio importa, **Then** no importa la base de datos, ni el controlador de SQLite, ni el adaptador: la dependencia sigue yendo hacia dentro.
2. **Given** un intento de importar la base de datos o el controlador de SQLite desde cualquier paquete que no sea `internal/{cache,store,graph}`, **When** se ejecutan los controles de `make ci`, **Then** fallan tanto en el lint como en el test de arquitectura, nombrando la regla R3.
3. **Given** el código nuevo del adaptador, **When** se ejecuta el lint, **Then** `sqlclosecheck` y `rowserrcheck` quedan en verde sin ninguna supresión: no hay filas, sentencias ni conexiones sin cerrar.
4. **Given** el binario distribuido, **When** se revisa qué enlaza al cerrar el hito, **Then** no enlaza ni el adaptador de caché ni el controlador de SQLite, porque ningún applet los usa todavía, y su conjunto de verbos es el mismo que al cerrar H2.

---

### Edge Cases

- **Lectura de una clave que nunca se guardó**: ausencia, no fallo; con `--offline`, la ausencia se convierte en el código 4 (FR-013, FR-016).
- **Lectura de una entrada caducada**: ausencia, sin servir el valor ni con aviso, también con `--offline` (FR-008, FR-018).
- **Valor guardado vacío**: se distingue de la ausencia; un valor de cero bytes guardado se devuelve como valor presente y vacío (FR-012).
- **Clave vacía o TTL menor o igual que cero**: valor inválido, clase «argumentos» (código 2), y no se escribe nada. «No cachear» no se dice con un TTL de cero: se dice no llamando a la escritura (FR-010, FR-011).
- **Escritura con `--offline`**: no se escribe nada y el fallo es explícito, de la clase «inesperado» (código 1), porque lo que está mal no es la invocación de la persona sino el código que intenta escribir en un modo que solo lee (FR-017, FR-033).
- **`--offline` con el directorio de caché vacío**: ni se crea la base de datos ni se aplica ninguna migración; todas las lecturas son ausencia y el código es 4 (FR-015, FR-016).
- **`KITLEGAL_CACHE_DIR` u opción de ruta con un valor inservible en sí mismo** (cadena vacía, ruta que existe y no es un directorio): clase «argumentos» (código 2) en cualquier modo, nombrando de dónde vino la ruta y cuál era, nunca una caída silenciosa a la ruta por omisión (FR-022).
- **El directorio de caché existe pero no se puede escribir** (permisos, disco de solo lectura), **fuera del modo de solo lectura**: clase «argumentos» (código 2) si se detecta al construir con la ruta declarada; un fallo de escritura sobrevenido (disco lleno, error de E/S) es de la clase «inesperado» (código 1) (FR-022, FR-033).
- **`--offline` con el directorio de caché inexistente o no escribible**: en ese modo la existencia y la escribibilidad del directorio no son condición, porque no se crea ni se escribe nada. Un directorio inexistente se trata como un `cache.db` inexistente —ausencia, código 4, nunca el 2 de una ruta inservible— y uno no escribible se lee con normalidad. Solo si el directorio existe y deniega el acceso, de modo que no se puede leer `cache.db` ni saber si está, el fallo es «inesperado» (código 1), nunca una ausencia falsa (FR-015, FR-022, FR-033).
- **La base de datos tiene una versión de esquema mayor que la conocida**: fallo explícito, clase «inesperado» (código 1), sin modificar ni borrar el fichero: un binario antiguo no destruye la caché de uno nuevo (FR-027).
- **El fichero existe y no es una base de datos utilizable** (corrupta, otro formato, truncada): fallo explícito de la clase «inesperado» (código 1), sin borrarla y sin rehacerla por su cuenta (FR-028).
- **Una migración interrumpida a mitad** (el proceso muere): la base queda en la versión anterior, nunca a medias, y la invocación siguiente la aplica entera (FR-026).
- **Dos invocaciones simultáneas sobre la misma base**, una escribiendo y otra leyendo —también cuando la que lee es de solo lectura (`--offline`)—: ninguna falla por bloqueo, la base no se corrompe y la lectura ve toda entrada que la otra ya confirmó, nunca una ausencia falsa (FR-015, FR-030, FR-031, FR-032).
- **Dos clientes de caché en el mismo proceso**: se comportan como dos invocaciones y la suite pasa con el detector de carreras (FR-032).
- **Clave muy larga o valor grande**: se guardan y se devuelven íntegros; H3 no fija ningún límite de tamaño ni desalojo, y no lo simula (FR-012, *Fuera de alcance*).
- **La misma clave escrita a la vez por dos clientes**: la última escritura completa gana y la entrada nunca queda mezclada entre las dos (FR-007, FR-032).
- **Cierre del cliente de caché**: cerrar dos veces no rompe nada y ninguna operación posterior al cierre escribe ni lee (FR-004).

## Requirements *(mandatory)*

### Functional Requirements

**Puerto y superficie**

- **FR-001**: `internal/core` DEBE definir el puerto `Cache` con las dos operaciones que nombra el hito —lectura por clave y escritura con vigencia (`Get`/`Put(key, ttl)`)—, expresadas con tipos del dominio y de la biblioteca estándar. La lectura DEBE devolver el contenido, un indicador booleano de presencia y un error —`Get(ctx, clave) ([]byte, bool, error)`—, siguiendo el idioma «coma ok» de Go; la escritura devuelve solo un error. El dominio NO DEBE importar la base de datos, el controlador de SQLite ni el adaptador (reglas R1 y R3, constitución §IV).
- **FR-002**: `internal/cache` DEBE implementar ese puerto sobre SQLite (`modernc.org/sqlite`, sin cgo, ADR 0002) y DEBE ser el único lugar donde se abre y se migra la base de datos de la caché.
- **FR-003**: Toda operación de la caché DEBE exigir un `context.Context` proporcionado por quien llama, y su cancelación o vencimiento DEBE cortar la operación. El paquete NO DEBE crear un contexto de fondo por su cuenta.
- **FR-004**: El cliente de caché DEBE poder cerrarse explícitamente por quien lo construyó, cerrar dos veces NO DEBE fallar, y ninguna operación DEBE dejar abierta una conexión, una sentencia preparada o un conjunto de filas.
- **FR-005**: La superficie pública del adaptador NO DEBE exponer la conexión a la base de datos ni ninguna forma de ejecutar SQL arbitrario desde fuera del paquete: quien usa la caché solo puede hacer lo que el puerto declara (precedente de H2, specs/003 FR-002).

**Entrada, vigencia y claves**

- **FR-006**: La escritura DEBE recibir la clave, el contenido a guardar y la vigencia (TTL), y DEBE registrar el instante de expiración calculado con el reloj del cliente en el momento de escribir. El contenido DEBE tratarse como **opaco**: una secuencia de bytes que la caché guarda y devuelve sin interpretar, sin saber qué hay dentro y sin añadirle metadatos propios. Lo que el sobre de salida necesita para citar —dirección final, estado, fecha de consulta— lo mete dentro del contenido quien llama, cuando exista quien llame (H4, *Assumptions*).
- **FR-007**: Escribir una clave que ya existe DEBE sustituir la entrada anterior por completo —contenido y vigencia—, sin dejar rastro de la anterior y sin duplicar la clave.
- **FR-008**: La lectura DEBE entregar la entrada **solo** si está vigente según el reloj del cliente. Una entrada expirada DEBE tratarse exactamente como una ausencia. El borde se declara aquí para que no dependa de la implementación: una entrada es vigente mientras el instante del reloj es **anterior** al instante de expiración; en el instante exacto de expiración, y después, es ausencia.
- **FR-009**: El reloj que decide la vigencia DEBE poder inyectarse al construir el cliente, de modo que un test pueda adelantarlo sin esperar tiempo real (control literal del hito).
- **FR-010**: Una vigencia menor o igual que cero DEBE rechazarse como valor inválido (clase «argumentos», código 2) sin escribir nada: «no cachear» no se expresa con una vigencia nula, sino no escribiendo.
- **FR-011**: La clave DEBE tratarse como opaca: la caché no la interpreta, no la deriva y no la normaliza. Quien llama es responsable de que la clave distinga todo lo que deba distinguir; el esquema de claves por fuente llega con el primer adaptador real (H4, *Fuera de alcance*). Una clave vacía es valor inválido (clase «argumentos», código 2).
- **FR-012**: Lo guardado DEBE devolverse íntegro y byte a byte igual que se escribió, sin truncar, recodificar ni interpretar. Un contenido de cero bytes DEBE poder guardarse y DEBE distinguirse de la ausencia.

**Lectura, ausencia y escritura**

- **FR-013**: La lectura DEBE distinguir tres resultados: presente y vigente, ausente (incluida la expirada) y fallo. La ausencia NO DEBE ser un error de la caché, sino el resultado normal que lleva a quien llama a pedir a la fuente. Estos tres resultados se expresan con el indicador de presencia y el error de FR-001: presente y vigente es `presente=true, err=nil`; ausente es `presente=false, err=nil`; fallo es `err != nil`. Quien llama NO DEBE comparar errores para saber si había entrada.
- **FR-014**: Fuera del modo de solo lectura, la ausencia NO DEBE producir por sí sola ningún código de salida ni alterar el resultado de la invocación.

**`--offline`: solo lectura**

- **FR-015**: El modo de operación lo fija **quien construye** la caché, no la caché: el cliente NO DEBE leer ninguna bandera ni inspeccionar el contexto de ejecución de cada llamada. Quien lo construye —en H3 el applet mínimo y el adaptador de prueba (FR-047); desde H4 el adaptador de fuente o la composición— toma el modo de `schema.Contexto.Offline`, que H1 entrega ya interpretado al applet, y lo declara como opción del constructor, nunca como parámetro de cada operación ni por el contexto de cada llamada; DEBE conocerse al construir porque crear la base de datos y aplicar migraciones ocurre al abrir, no al operar. Construido en modo de solo lectura —lo que ocurre cuando la invocación trae `--offline`—, el cliente NO DEBE escribir ninguna entrada, NO DEBE aplicar ninguna migración y NO DEBE crear la base de datos si no existe. El fichero de la base de datos DEBE quedar, al terminar la invocación, exactamente como estaba. «El fichero de la base de datos queda exactamente como estaba» se refiere a `cache.db`: los ficheros auxiliares que SQLite usa para leer una base en WAL (`cache.db-wal`, `cache.db-shm`) pueden aparecer o desaparecer durante la invocación sin que eso cuente como cambio (SC-003 los excluye de la comparación). Si `cache.db` no existe, la caché NO DEBE abrir nada y el directorio queda intacto también en los auxiliares. Que el directorio de caché exista, se pueda crear o se pueda escribir NO DEBE ser condición en este modo, porque en él no se crea ni se escribe nada: un directorio de caché **inexistente** DEBE tratarse exactamente como un `cache.db` inexistente —no se crea nada y toda lectura es ausencia, que FR-016 convierte en el código 4— y un directorio existente que **no se puede escribir** —el contenedor aislado o el disco de solo lectura de US3— DEBE leerse con normalidad. La exigencia de que el directorio se pueda crear y escribir de FR-022 queda acotada, por tanto, al modo normal; lo que sigue siendo inservible en cualquier modo es el valor de la ruta en sí —cadena vacía o ruta que existe y no es un directorio—, que es «argumentos» (código 2). Si `cache.db` existe pero no se puede abrir o leer, o si el directorio existe y deniega el acceso de modo que ni siquiera se puede determinar si `cache.db` está, el fallo DEBE ser explícito de la clase «inesperado» (código 1) y NO DEBE degradarse a ausencia (FR-016, FR-033). Abrir en modo de solo lectura NO DEBE recortar lo que la lectura ve ni cómo convive con quien escribe: una entrada **ya confirmada** por otra invocación DEBE verse aunque todavía no esté consolidada en `cache.db`, y los bloqueos de una invocación normal que escriba a la vez DEBEN respetarse (FR-030, FR-032). Declarar ausente —y, por FR-016, no disponible (código 4)— una entrada que otra invocación ya confirmó incumple este requisito. El significado que `--offline` adquiere en este hito es **solo** este: el hito no le pide nada al cliente HTTP de H2 y H3 no lo toca (*Fuera de alcance*, *Assumptions*).
- **FR-016**: En ese modo, una lectura que no encuentra entrada vigente —ausente o expirada— DEBE producir un fallo de la clase «fuente no disponible», que el kernel traduce al código de salida **4** (literal del hito: «devuelve exit 4 si falta»).
- **FR-017**: En ese modo, un intento de escritura NO DEBE escribir nada y DEBE fallar de forma explícita con la clase «inesperado» (código 1): lo que está mal no es la invocación de la persona, sino el código que intenta escribir donde solo se lee. El plan puede además impedirlo por construcción, como refuerzo y nunca en sustitución del fallo en ejecución que este requisito y SC-011 miden, y sin partir el puerto ni el cliente en dos tipos (FR-001, FR-015).
- **FR-018**: En ese modo la vigencia se aplica igual: una entrada expirada NO DEBE servirse por el hecho de no haber red, ni con aviso (FR-008).

**Ubicación de la base de datos**

- **FR-019**: Por omisión, la base de datos DEBE vivir en `~/.cache/kitlegal/cache.db`, la ruta que fijan `CLAUDE.md`, `refs/00-README.md` y el propio hito. Esa ruta se toma **al pie de la letra** —el directorio `.cache/kitlegal/` dentro del directorio de la cuenta, en cualquier plataforma— y NO DEBE reinterpretarse con el directorio de caché de cada sistema ni con variables del entorno de escritorio: `~/.cache/kitlegal/` es una decisión cerrada de `CLAUDE.md` que este spec no reabre, y quien necesite otra ubicación la declara con `KITLEGAL_CACHE_DIR` (FR-020, *Assumptions*).
- **FR-020**: `KITLEGAL_CACHE_DIR` DEBE tener precedencia sobre la ruta por omisión: la base de datos se llama `cache.db` dentro del directorio que la variable declara y no se escribe nada bajo la ruta por omisión.
- **FR-021**: Fuera del modo de solo lectura, el directorio DEBE crearse si no existe y la base de datos DEBE crearse y migrarse en la primera invocación. El directorio y el fichero DEBEN crearse con acceso reservado a la cuenta que los crea, que es lo que el lint del proyecto ya exige a cualquier código que cree ficheros o directorios (`gosec`, activo en `.golangci.yml`, reglas G301 y G302 sobre permisos) y lo que pide la prudencia por omisión con lo que se guarda en el disco de la persona.
- **FR-022**: Una ruta de caché inservible —venga de `KITLEGAL_CACHE_DIR` o de la opción del constructor (FR-023)— DEBE fallar de forma explícita con la clase «argumentos» (código 2), nombrando de dónde vino la ruta y cuál era, y NO DEBE caer en silencio a la ruta por omisión. Son inservibles **en cualquier modo** la cadena vacía y una ruta que existe y no es un directorio. Que el directorio se pueda crear y escribir DEBE exigirse **solo fuera del modo de solo lectura** —donde la caché tiene que crear la base de datos, migrarla y escribir entradas para servir de algo (FR-021)—: ahí, un directorio que no existe y no se puede crear, o que existe y no se puede escribir, es también «argumentos» (código 2). En modo de solo lectura esa exigencia NO DEBE aplicarse y rige FR-015: no se crea nada, no se escribe nada, un directorio inexistente es ausencia (código 4) y no un fallo de argumentos, y uno no escribible se lee con normalidad (*Assumptions*).
- **FR-023**: La ruta efectiva DEBE poder declararse como opción explícita del constructor, para que un test trabaje sobre `t.TempDir()` sin tocar el entorno del proceso. El orden de precedencia DEBE ser: opción del constructor, variable de entorno, ruta por omisión. La variable DEBE leerse una sola vez, al construir.

**Esquema y migraciones**

- **FR-024**: El esquema DEBE crearse y actualizarse con migraciones **embebidas en el binario**, sin ficheros externos y sin que ningún consumidor escriba SQL.
- **FR-025**: Una tabla `schema_version` DEBE registrar la versión de esquema aplicada. Abrir dos veces la misma base de datos NO DEBE volver a aplicar lo aplicado ni fallar: la migración es idempotente.
- **FR-026**: Cada migración DEBE aplicarse de forma atómica, de modo que una interrupción deje la base en la versión anterior y nunca a medias.
- **FR-027**: Una base de datos cuya versión de esquema es **mayor** que la que el binario conoce DEBE terminar en un fallo explícito (clase «inesperado», código 1) que diga qué versión se esperaba y cuál se encontró, y el fichero NO DEBE modificarse ni borrarse.
- **FR-028**: Un fichero que existe y no es una base de datos utilizable DEBE terminar en un fallo explícito (clase «inesperado», código 1) sin borrarlo ni rehacerlo por su cuenta.
- **FR-029**: H3 aporta la **primera** versión del esquema. Cuántas migraciones la componen, qué tablas e índices crea y cómo se representan la clave, el contenido y la expiración los fija el plan (*Assumptions*).

**Concurrencia y `PRAGMA`**

- **FR-030**: La base de datos DEBE abrirse con el registro de escritura anticipada (WAL) activo, que es lo que permite leer mientras otra invocación escribe.
- **FR-031**: Los `PRAGMA` DEBEN ser seguros: ninguno DEBE sacrificar la integridad de lo escrito por velocidad, y la espera ante un bloqueo DEBE estar configurada para que una invocación simultánea no falle de inmediato. Los valores concretos los fija el plan.
- **FR-032**: Dos clientes de caché sobre la misma base de datos —en el mismo proceso y en dos procesos, y tanto entre dos clientes normales como entre un cliente de solo lectura (FR-015) y uno normal— DEBEN poder leer y escribir sin corromperla y sin fallar por bloqueo; lo que un lector ve DEBE ser una entrada completa o ninguna, y DEBE incluir toda entrada que el escritor ya haya confirmado. La suite DEBE pasar con el detector de carreras activado e incluir al menos un test concurrente.

**Clases de fallo y códigos de salida**

- **FR-033**: Todo fallo de la caché DEBE declarar su clase por el puerto que el dominio ya define para eso (`schema.ConClase`, H1) y NO DEBE decidir ningún código de salida por su cuenta. La lista es **cerrada** y ningún camino de fallo queda sin clase:
  - clave vacía, vigencia menor o igual que cero → «argumentos» (código 2);
  - `KITLEGAL_CACHE_DIR` u opción de ruta inservibles —cadena vacía o ruta que existe y no es un directorio, en cualquier modo; directorio que no se puede crear ni escribir, solo fuera del modo de solo lectura (FR-022)— → «argumentos» (código 2);
  - ausencia o entrada expirada en modo de solo lectura, incluida la de un directorio de caché o un `cache.db` inexistentes (FR-015) → «fuente no disponible» (código 4);
  - versión de esquema desconocida, fichero inutilizable, escritura en modo de solo lectura, acceso denegado que impide leer o saber si `cache.db` está (FR-015), fallo de E/S o de la base de datos → «inesperado» (código 1).

  Ningún fallo de la caché DEBE producir los códigos 3, 5 ni 6: no encontrar algo en la caché no es no encontrarlo en la fuente, la caché no tiene límite de peticiones y no realiza ninguna acción con identidad.
- **FR-034**: Ningún camino de la caché DEBE terminar en `panic` (Definition of Done 5).
- **FR-035**: Los mensajes de fallo DEBEN nombrar lo necesario para actuar —la ruta del fichero, la variable o la opción implicada y, cuando hay una petición o clave implicada, cuál era—, sin traza técnica en el sobre de salida, que ya acota H1.

**Arquitectura y controles**

- **FR-036**: `internal/cache` DEBE quedar como dueño de la regla R3 junto con `internal/store` e `internal/graph` cuando existan: ningún otro paquete del módulo DEBE importar la base de datos ni el controlador de SQLite, comprobado en las dos capas (lint y test de arquitectura sobre el grafo transitivo real).
- **FR-037**: R3 DEBE seguir vigilando después del hito. H3 le da su **primer** dueño real; la exigencia de «dueño obligatorio» del test de arquitectura NO DEBE activarse todavía, porque exige que los tres dueños existan y dos llegan en hitos posteriores (H12 y H17), y la razón DEBE quedar escrita donde hoy está escrita la del caso anterior.
- **FR-038**: `internal/core/**` NO DEBE importar `internal/cache` (regla R1).
- **FR-039**: `sqlclosecheck` y `rowserrcheck` DEBEN quedar en verde sobre el código nuevo —que es donde por primera vez tienen algo que vigilar— sin ninguna supresión (`//nolint`).
- **FR-040**: DEBEN existir tests de integración marcados con `//go:build integration` que trabajen sobre una base de datos en `t.TempDir()`, no toquen la caché real de la cuenta y no dejen ningún fichero fuera de ese directorio.
- **FR-041**: Esos tests DEBEN ejecutarse como gate de integración continua, porque «todo lo que entra se queda como gate de CI» (`docs/ROADMAP.md` §3): un test de integración que falla, o un fichero etiquetado que no compila, DEBEN hacer fallar CI. El análisis estático DEBE alcanzar también el código etiquetado, de modo que `sqlclosecheck` lo vigile.
- **FR-042**: DEBE existir un test unitario de expiración de vigencia con el reloj inyectado, que no espere tiempo real.
- **FR-043**: La única dependencia nueva DEBE ser `modernc.org/sqlite`, ya fijada por la constitución §V y por `docs/ROADMAP.md` §3, más los módulos que ella arrastre, que DEBEN quedar registrados en `go.mod`. Cualquier otra dependencia requiere justificación explícita.
- **FR-044**: La superficie visible del binario distribuido NO DEBE cambiar por este hito: ningún applet, verbo ni bandera nueva, y el adaptador de prueba NO DEBE registrarse en ningún binario, ni en el distribuido ni en el de extremo a extremo (precedente de H2, specs/003 FR-062). Mientras ningún applet use la caché, el binario distribuido NO DEBE enlazar `internal/cache` ni el controlador de SQLite, y la lista de módulos de terceros que vigila su superficie NO DEBE cambiar.

**Adaptador de prueba (criterio de aceptación del hito)**

- **FR-045**: DEBE existir un adaptador de prueba que combine la caché con el cliente HTTP de H2 y demuestre el criterio de aceptación: dos consultas idénticas dentro de la vigencia producen **una sola** petición. DEBE vivir en material de test y NO DEBE crearse bajo `internal/source/`, reservado a los adaptadores de fuente real, cuya aparición es decisión humana (constitución, capa 3 de los gates). DEBE vivir en ficheros `_test.go` del paquete externo `cache_test`, dentro de `internal/cache`, siguiendo el precedente de `internal/httpx/adaptador_test.go` (H2, specs/003 FR-060 a FR-062): no se crea ningún paquete nuevo compartido entre paquetes de test, y H4 no lo reutiliza —escribe su propio adaptador real con sus propios fixtures.
- **FR-046**: La verificación DEBE hacerse con un cliente de reproducción en **modo estricto**: construido sobre un directorio que no contiene la grabación de la petición en cuestión, de modo que cualquier petición emitida termine en un fallo ruidoso que hace fallar el test. No hace falta ninguna capacidad nueva en `internal/httpx` para eso: su reproducción ya falla así ante una petición sin grabación (specs/003 FR-047 y FR-063); si el plan concluyera que hace falta, sería un cambio de este requisito y no un detalle de diseño.
- **FR-047**: El mismo adaptador de prueba DEBE demostrar el camino de `--offline` con el kernel invocado en proceso: con la entrada guardada y vigente, responde sin red y con código 0; sin ella, la invocación termina con el código **4**. Ese applet mínimo es, en H3, **el único sitio donde `--offline` se convierte en un cliente de solo lectura**: recibe de H1 el contexto de ejecución ya interpretado, lee `schema.Contexto.Offline` y construye la caché con la opción de solo lectura (FR-015); la caché no ve la bandera. Por ese mismo camino DEBE acreditarse la escritura en modo de solo lectura (FR-017): un intento de escribir sobre ese cliente falla con la clase «inesperado» y el kernel lo traduce al código 1 (SC-011). El applet vive en el mismo paquete `cache_test` y su registro de applets existe solo dentro del test: ningún binario lo conoce (FR-044).

### Key Entities

- **Puerto de caché (`Cache`)**: la idea de caché tal como la ve el dominio y la usará un adaptador de fuente: guardar algo bajo una clave con una vigencia, y recuperarlo mientras esté vigente. Vive en `internal/core` y no menciona SQLite ni fichero alguno.
- **Cliente de caché**: lo que devuelve el constructor del adaptador. Reúne la ruta efectiva de la base de datos, el reloj que decide la vigencia y el modo de operación (normal o solo lectura). El modo se fija al construir, con una opción funcional del constructor (p. ej. `cache.SoloLectura()`), nunca por parámetro de cada operación ni por el contexto de ejecución: FR-015 exige conocerlo antes de abrir la base. El cliente no lee ninguna bandera; quien lo construye toma el modo de `schema.Contexto.Offline`, que H1 entrega ya interpretado (FR-015, FR-047). Es, con `store` y `graph` cuando existan, lo único del módulo que puede abrir una base de datos.
- **Entrada de caché**: lo que se guarda bajo una clave: el contenido —una secuencia de bytes **opaca**, que la caché guarda y devuelve sin interpretar— y el instante en que deja de ser vigente. Nada más: los metadatos que el sobre de salida necesita para citar —dirección final, estado, fecha de consulta— no son campos de la entrada, sino parte del contenido que serializa quien llama, desde H4 el adaptador de fuente (FR-006, FR-012, *Assumptions*). Cómo se representan la clave, el contenido y el instante de expiración dentro del esquema lo fija el plan (FR-029).
- **Clave**: la cadena opaca con la que quien llama identifica lo que guarda. La caché no la interpreta; quien la construye responde de que distinga lo que debe distinguir.
- **Vigencia (TTL)**: cuánto tiempo vale lo guardado, declarado al escribir y convertido en un instante de expiración. Cada fuente tendrá la suya (`Source.TTL()`, H4); H3 no fija ningún valor por omisión.
- **Reloj**: de dónde sale «ahora» para decidir si una entrada está vigente. Se inyecta al construir, lo que permite comprobar la caducidad sin esperar.
- **Base de datos de la caché**: el fichero `cache.db` del directorio de caché, con su esquema versionado en `schema_version`. Sobrevive a las invocaciones y a las versiones del binario; es desechable para la persona usuaria, pero el binario nunca la borra por su cuenta.
- **Modo de solo lectura**: el modo en que quien construye la caché la pone cuando la invocación trae `--offline`. No escribe, no crea —ni la base de datos ni el directorio de caché—, no migra, y convierte la ausencia en un fallo de la clase «fuente no disponible». Que el directorio exista o se pueda escribir no es condición suya: lo que no está es ausencia (código 4), no un fallo de argumentos (FR-015, FR-022). Leer menos que un cliente normal tampoco forma parte del modo: ve todo lo que otra invocación ya confirmó y respeta sus bloqueos (FR-015).
- **Clase de fallo**: el vocabulario que el kernel ya conoce desde H1. Este hito produce «argumentos», «fuente no disponible» e «inesperado», y nunca «no encontrado», «límite o TOS» ni «identidad humana» (FR-033).

## Success Criteria *(mandatory)*

### Measurable Outcomes

- **SC-001**: Dos consultas idénticas dentro de la vigencia producen **exactamente una** petición: la segunda se ejecuta contra un cliente de reproducción estricto sobre un directorio sin la grabación de esa petición, de modo que una petición emitida haría fallar el test nombrándola, y el test pasa.
- **SC-002**: La caducidad se comprueba con el reloj inyectado en las tres posiciones del borde —antes del instante de expiración (presente), en el instante exacto (ausencia) y después (ausencia)— y el test no espera tiempo real: su duración no depende de la vigencia usada.
- **SC-003**: Con `--offline`: una entrada guardada y vigente se responde con código de salida 0 y cero peticiones; una ausente o expirada termina con código **4** y cero peticiones; y `cache.db` queda idéntico antes y después —sin directorio de caché creado, sin base de datos creada, sin migración aplicada, sin entrada escrita—, medido byte a byte sobre ese fichero (o su ausencia, si no existía) y excluyendo los ficheros auxiliares transitorios de SQLite (`cache.db-wal`, `cache.db-shm`), que una conexión de solo lectura no controla.
- **SC-004**: La base de datos aparece donde debe y solo ahí: con `KITLEGAL_CACHE_DIR` apuntando a un directorio temporal, el fichero `cache.db` está en él y no hay ningún fichero nuevo bajo la ruta por omisión; con un valor de ruta inservible en sí mismo —cadena vacía o ruta que existe y no es un directorio—, la invocación termina con el código **2** nombrando la variable y la ruta, y no se escribe nada en ninguna parte, en cualquier modo. La escribibilidad se mide cualificada por el modo: un directorio que no se puede crear ni escribir da el código **2** fuera del modo de solo lectura, mientras que con `--offline` un directorio de caché inexistente termina con el código **4** sin crear nada y uno no escribible se lee con normalidad, nunca con el código 2 (FR-015, FR-022, SC-011). Las variantes que dependen de permisos se miden donde el sistema de ficheros los hace valer; las demás son deterministas en cualquier entorno. Ningún test escribe en la caché real de la cuenta.
- **SC-005**: Abrir dos veces la misma base de datos deja la misma versión en `schema_version` y cero errores; una base con versión mayor que la conocida termina con código **1**, dice qué versión esperaba y cuál encontró, y deja el fichero idéntico byte a byte; un fichero que no es una base de datos utilizable termina igual y tampoco se borra.
- **SC-006**: La base de datos creada tiene el registro de escritura anticipada (WAL) activo, comprobado consultándolo en la propia base, y ningún `PRAGMA` de los aplicados sacrifica la integridad de lo escrito.
- **SC-007**: Con dos clientes trabajando a la vez sobre la misma base —uno escribiendo y otro leyendo, en el mismo proceso y en dos procesos—, cero fallos por bloqueo, cero entradas a medias y `go test -race` en verde. La pareja se mide dos veces: con dos clientes normales y con un cliente de **solo lectura** leyendo frente a un escritor normal; en el segundo caso, además, el lector de solo lectura encuentra **todas** las entradas que el escritor confirmó antes de la lectura —cero ausencias falsas, cero códigos 4 indebidos—, también en el mismo proceso y en dos procesos.
- **SC-008**: `make ci` queda en verde con `sqlclosecheck` y `rowserrcheck` activos y sin ninguna supresión (`//nolint`) añadida por este hito.
- **SC-009**: Los tests de integración etiquetados se ejecutan en integración continua y su fallo hace fallar el gate; trabajan solo dentro de `t.TempDir()` —al terminar no queda ningún fichero fuera de él— y el análisis estático alcanza también los ficheros etiquetados.
- **SC-010**: Lo guardado y lo leído son idénticos byte a byte para un contenido cualquiera y para un contenido de cero bytes, que se distingue de la ausencia; una clave distinta nunca devuelve el contenido de otra.
- **SC-011**: Las **ocho** situaciones de fallo del hito se traducen a un código de salida estable y comprobado, ninguna al azar: clave vacía → 2, vigencia menor o igual que cero → 2, ruta o variable inservibles → 2 (cadena vacía y ruta que no es un directorio en cualquier modo; directorio que no se puede crear ni escribir, solo fuera del modo de solo lectura), ausencia o entrada expirada en solo lectura → 4, **directorio de caché inexistente en modo de solo lectura → 4 y nunca 2**, versión de esquema desconocida → 1, fichero inutilizable → 1, escritura en modo de solo lectura → 1. Ninguna produce 3, 5 ni 6.
- **SC-012**: La superficie visible del binario distribuido no cambia: el conjunto de verbos que atiende al cerrar H3 es el mismo que al cerrar H2, el binario no enlaza el adaptador de caché ni el controlador de SQLite, y la lista de módulos de terceros que vigila su superficie es la misma.
- **SC-013**: Las reglas de dependencia se comprueban mecánicamente: un intento de importar la base de datos o el controlador de SQLite desde un paquete que no sea `internal/{cache,store,graph}`, y un intento de que el dominio importe el adaptador de caché, hacen fallar `make ci` nombrando la regla (R3 y R1).
- **SC-014**: La Definition of Done aplicable queda cumplida: `make ci` en verde (`fmt`, `lint`, `test` con `-race`, `vuln`, `schema-check`), tests offline para todo el código nuevo, cobertura global ≥ 70 % y `internal/core/**` ≥ 85 %, y ninguna dependencia nueva fuera de las fijadas.

## Fuera de alcance

Lo que sigue **no** se construye en H3, por no estar en la sección del hito, en `CLAUDE.md`, en `refs/` ni en la constitución. Cada línea indica dónde le corresponde llegar.

- **Cualquier adaptador de fuente real** bajo `internal/source/` —BOE incluido—, el puerto `Source` (`Name`, `Fetch`, `TTL`, `Terms`) y el puerto `Fetcher`: es H4. El adaptador de este hito es de prueba y vive en material de test (FR-045).
- **El esquema de claves de caché y el valor del TTL de cada fuente**: los decide la fuente cuando exista (`Source.TTL()`, H4). H3 trata la clave como opaca y no fija ninguna vigencia por omisión (FR-011).
- **Peticiones condicionales (`ETag`, `If-Modified-Since`) y revalidación de lo caducado contra la fuente**: no las pide el hito. Aquí una entrada caducada es simplemente una ausencia (FR-008); revalidar en lugar de volver a pedir llega, si llega, con la fuente que lo permita (H4 o posterior).
- **Trasladar a disco la caché de `robots.txt`**: H2 la dejó en memoria y mientras vive el cliente (specs/003 FR-015), y el hito no pide moverla.
- **`internal/store` e `internal/graph`**, `world.db`, `.kitlegal/case.db`, FTS5 y cualquier cosa del grafo: H12, H17 y H18. H3 solo aporta el primer dueño de la regla R3 (FR-036, FR-037).
- **Activar la exigencia de «dueño obligatorio» de R3 en el test de arquitectura**: exige que los tres paquetes de almacenamiento existan, y dos llegan después (H12 y H17). Hasta entonces la regla sigue como está, con la razón escrita (FR-037).
- **Verbos o banderas de mantenimiento de la caché** (`purgar`, `estadísticas`, compactar, invalidar por clave o por fuente): no los pide el hito y no hay applet que los exponga; el binario no cambia (FR-044).
- **Política de tamaño máximo, desalojo (LRU), compresión de lo guardado y borrado de lo caducado en segundo plano**: nada de eso está en el hito. Una entrada caducada deja de servirse (FR-008); si el plan la borra al leerla o la deja, es un detalle de implementación sin efecto observable.
- **Cifrado de la base de datos**: la caché guarda respuestas de fuentes públicas, no datos personales; el cuidado con los datos del asunto vive en el grafo del asunto (constitución §VII, H18).
- **Caché compartida entre máquinas, entre cuentas o entre procesos remotos, y MCP con caché compartida**: excluido por `docs/ROADMAP.md` §5.
- **Métricas o contadores de aciertos y fallos de la caché**: el registro de eventos que H1 dejó (`log/slog` a la salida de error, `--verbose`) basta, y el hito no pide ninguna medida.
- **Reglas del grafo que dependen de la vigencia** (`fuente-caducada`, `version-obsoleta`): son de `graph check`, H18 y H21.
- **Dar significado a `--no-graph` o a `--asunto`, y emitir al grafo (`Emit`)**: el grafo es H17, y ahí se incorpora a lo anterior.
- **Que el cliente HTTP rechace emitir peticiones bajo `--offline`**: el hito da a la bandera un significado acotado a la caché («con `--offline` solo lee, y devuelve exit 4 si falta») y no pide nada a `internal/httpx`, que H2 cerró sin darle ninguno. H3 no lo toca. Quien tiene que no salir a la red bajo esa bandera es quien decide si pedir, y eso es el adaptador de fuente: en H3 lo demuestra el adaptador de prueba (FR-047) y en H4 lo hace el primero real. Si algún hito quisiera además la garantía por construcción dentro del cliente, sería un cambio de alcance suyo y no un detalle de este plan.
- **Reinterpretar el directorio de caché con las convenciones de cada plataforma** (`os.UserCacheDir`, `XDG_CACHE_HOME`, `~/Library/Caches`): no lo pide el hito y `CLAUDE.md` fija `~/.cache/kitlegal/` como directorio de la caché y del grafo del mundo (FR-019). `KITLEGAL_CACHE_DIR` cubre cualquier otra ubicación que alguien necesite, incluida la que su plataforma prefiera.
- **Semántica de `--dry-run` en la caché**: la bandera describe en lugar de ejecutar y ya tiene su contrato en H1 y su efecto en la red en H2; el hito no le pide nada a la caché y no se le inventa ninguno.
- **Validación de salida contra `schemas/*.json` y `--describe`**: H3 no añade ningún applet ni ninguna salida de applet; el punto 4 de la Definition of Done no tiene sujeto aquí.
- **Evals de skill**: H3 no entrega ni modifica ninguna skill; es un hito de fundación de los que el principio VIII admite porque protegen a las que vendrán. El andamiaje de evals es H5.
- **Dimensión territorial**: H3 no la tiene; el punto 11 de la Definition of Done no aplica.
- **Guion de extremo a extremo (`testscript`) nuevo**: no tiene sujeto. La entrega es un paquete interno que ningún binario expone (FR-044, SC-012), y `testscript` solo observa binarios; lo que este hito añade a `--offline` se comprueba con el kernel invocado en proceso (US3, FR-047). El primer extremo a extremo que ejercita la caché llega en H4, donde el roadmap ya lo pide.
- **Fila en `docs/SOURCES.md` y caso en `scripts/verify-sources.sh`**: H3 no consulta ninguna fuente real, así que no hay fuente que declarar ni verificar; y editar esa tabla es, en todo caso, decisión humana (constitución, capa 3).
- **Grabación de fixtures contra una fuente real**: prohibida dentro del bucle de implementación (constitución, «Reglas del modo desatendido»). El material de reproducción que este hito necesite se escribe a mano, como el de H2.
- **Entrada en `CHANGELOG.md` por el binario**: el punto 6 de la Definition of Done la exige cuando cambia un comportamiento visible, y H3 no cambia ninguno del binario (FR-044, SC-012), así que no hay guion de extremo a extremo. Lo que sí cambia a la vista es la composición de `make ci`, que gana `test-integration` (FR-041; el mecanismo lo fijó el plan, *Assumptions*), y ese cambio se registra bajo *Cambiado*, como H1 registró el de `make test-e2e` (revisión final).
- **ADR nuevo**: SQLite sin cgo ya está decidido en `docs/ADR/0002-sqlite-sin-cgo.md` y el patrón *Repository* para `cache`, `store` y `graph` está en `docs/ROADMAP.md` §2; solo haría falta un ADR si este hito se apartara de lo ya decidido.
- **Decidir dónde viven los fixtures grabados** (`testdata/` de la raíz frente a `internal/source/<fuente>/testdata/`): `docs/PENDIENTES.md` lo sitúa en la primera tarea `[datos]` de H4 y H3 no lo necesita, porque no graba ningún fixture.
- **Extraer la caché a `pkg/` o a una librería externa**: la API pública es H30.
- **Umbral de cobertura propio para `internal/cache`**: el hito no fija ninguno; rigen los umbrales generales de la Definition of Done (SC-014).

## Assumptions

Decisiones tomadas con la información disponible, registradas aquí para que un cambio en cualquiera de ellas no obligue a tocar ningún requisito. Las tres primeras son las que el modo desatendido habría dejado como preguntas abiertas si no las determinara el «Criterio de decisión autónoma» de la constitución: se resuelven con su punto 2 («lo no especificado no se implementa») y su punto 3 (elegir con criterio y dejar rastro, con la alternativa rechazada), y ninguna reabre una decisión cerrada.

- **El contenido guardado es opaco.** El hito nombra `Get/Put(key, ttl)` y no dice nada del contenido, así que la caché no sabe qué guarda: lo devuelve tal cual (FR-006, FR-012). Es además lo único compatible con que el puerto viva en el dominio y con las reglas de dependencia: una entrada que conservara los metadatos de una respuesta HTTP obligaría al dominio a nombrar tipos del cliente (R1) o al adaptador de caché a importar el de red, y la caché pasaría a sostener parte del sobre de salida sin que ningún hito lo pida. Lo que la cita necesita —dirección final, estado, fecha de consulta— lo serializa dentro del contenido quien llama, que desde H4 es el adaptador de fuente. Alternativa descartada: una entrada estructurada con esos metadatos, que se puede añadir después sin romper nada si una fuente lo pide, mientras lo contrario —quitarlos— sería un cambio de contrato.
- **`--offline` significa aquí solo lo que el hito dice.** La bandera queda acotada a la caché: solo lectura y código 4 si falta (FR-015, FR-016). Extenderla al cliente HTTP sería implementar lo no especificado y modificar un paquete que el hito no nombra. Alternativa descartada: hacer que `internal/httpx` rechace toda petición bajo esa bandera, que es una garantía atractiva pero de otro hito: en H3 no hay ningún applet que salga a la red, y quien decide si pedir es el adaptador de fuente, que llega en H4 (*Fuera de alcance*).
- **La ruta por omisión se lee al pie de la letra.** `~/.cache/kitlegal/` está en `CLAUDE.md` («decisiones ya tomadas: no reabrir»), en `refs/00-README.md` y en el enunciado del hito, y el mismo directorio alojará el grafo del mundo (H17): un solo sitio para lo que el binario cachea, igual en todas las plataformas y fácil de nombrar en la documentación (FR-019). Alternativa descartada: el directorio de caché de cada plataforma (`os.UserCacheDir`, `XDG_CACHE_HOME`, `~/Library/Caches` en macOS), que es la convención del sistema pero contradice una decisión cerrada, partiría la caché y el grafo entre dos rutas según la máquina y no lo pide ningún hito; quien la quiera la obtiene con `KITLEGAL_CACHE_DIR`.
- **En modo de solo lectura, el directorio de caché no tiene que existir ni ser escribible.** El hito pide que con `--offline` la caché «solo lea, y devuelva exit 4 si falta», y la respuesta a Q2 fija que una base que no existe no se abre y que solo un fallo real de lectura (código 1) rompe esa promesa. Exigir además, como hace FR-022 fuera de ese modo, un directorio creable y escribible convertiría el caso más común del modo —el contenedor aislado o el disco de solo lectura que describe US3— en un fallo de argumentos (código 2) por una escritura que nunca se iba a intentar, y dejaría la misma situación con dos códigos según qué requisito se leyera primero. Así que la creación y la escribibilidad del directorio son condición **solo fuera** del modo de solo lectura (FR-021, FR-022); en él, un directorio inexistente se trata como un `cache.db` inexistente (ausencia → código 4, FR-015 y FR-016) y lo único inservible en cualquier modo es el valor de la ruta en sí —cadena vacía o ruta que existe y no es un directorio—. Alternativa descartada: dejar FR-022 sin cualificar el modo, más corto de enunciar pero incoherente con FR-015 y con la respuesta a Q2, y que haría fallar con código 2 invocaciones con `--offline` que el hito promete terminar en 4.
- **La caché no decide cuándo pedir; lo decide quien la usa.** El hito entrega la herramienta —guardar y recuperar con vigencia— y el patrón de uso «mira antes de pedir» lo aplica el adaptador de fuente, que llega en H4. En H3 ese patrón solo existe en el adaptador de prueba, que es lo que hace comprobable el criterio de aceptación (FR-045).
- **La vigencia se declara al escribir, no al leer.** Es lo que dice el enunciado del hito (`Put(key, ttl)`) y lo que encaja con que cada fuente declare su propio TTL (`Source.TTL()`, `refs/kitlegal-estructura-y-ecosistema.md` §214). Alternativa descartada: pasar la vigencia en la lectura, que permitiría a dos llamantes tener criterios distintos sobre la misma entrada y haría el resultado dependiente de quién pregunta.
- **Una entrada caducada es una ausencia, también sin red.** Nada en el hito pide servir lo caducado, y hacerlo contradiría la constitución §II: una cita se sostiene sobre texto vigente y sobre una fecha de consulta, no sobre lo último que hubiera por ahí. Alternativa descartada: servir lo caducado con un aviso bajo `--offline`, que dejaría a la skill decidiendo sobre material que no puede fechar.
- **La reproducción estricta ya existe.** El criterio de aceptación se comprueba con `httpx.Replay` sobre un directorio sin la grabación de la petición: la reproducción de H2 falla de forma ruidosa y con clase «inesperado» ante cualquier petición no grabada (specs/003 FR-047, FR-063). Este hito no toca `internal/httpx`; si el plan concluyera que hace falta, sería un cambio de FR-046.
- **El reloj inyectado repite el patrón de H2.** Los reintentos del cliente HTTP ya se prueban con un reloj inyectado para no esperar de verdad; la caché adopta el mismo enfoque, y su forma exacta (tipo, firma, opción del constructor) es diseño.
- **La forma de la base de datos la fija el plan.** El número de migraciones, las tablas e índices, cómo se representa el instante de expiración, qué `PRAGMA` concretos se aplican, con cuánta espera ante bloqueo, con qué cadena de conexión se abre en solo lectura —siempre que la lectura vea lo que FR-015 exige ver y respete los bloqueos que exige respetar— y si lo caducado se borra al leerlo: todo eso es diseño, y ninguno de esos puntos cambia el enunciado de un requisito.
- **`modernc.org/sqlite` ya estaba elegido.** Está en la lista cerrada de la constitución §V, en `docs/ROADMAP.md` §3 y en `docs/ADR/0002-sqlite-sin-cgo.md`: el spec lo registra para acotar FR-043, no lo elige. Arrastra módulos indirectos que quedan en `go.mod`; mientras ningún applet use la caché, no entran en el binario distribuido (FR-044).
- **La base de datos es desechable para la persona usuaria, pero el binario no la borra.** Si se pierde, lo único que se pierde es velocidad. Esa es la razón de que un fichero inutilizable o de versión desconocida termine en un fallo explícito y no en un borrado automático (FR-027, FR-028): borrar por iniciativa propia un fichero del disco de otra persona no es cosa de una herramienta, y la caché convivirá en el mismo directorio con el grafo del mundo (`world.db`, H17).
- **Los tests de integración necesitan puerta en CI.** `docs/ROADMAP.md` §3 sitúa los tests de integración en H3 y dice que «todo lo que entra se queda como gate de CI», pero `make ci` hoy no ejecuta el objetivo que los lanza. FR-041 lo exige sin fijar el mecanismo —dentro de `make ci` o como paso propio del flujo de integración continua—, que es decisión del plan.
- **El escenario de dos procesos se acredita relanzando el propio binario de test.** FR-032 y SC-007 piden «el mismo proceso y en dos procesos»: el segundo proceso es el propio ejecutable de test relanzado como subproceso (`exec.CommandContext(ctx, os.Args[0], ...)` con una variable de entorno que fija el papel de escritor o lector), sin añadir ningún `package main` nuevo (FR-044). Alternativa descartada: compilar un programa auxiliar aparte bajo material de test, que añadiría un binario que el hito dice no añadir y correría sin `-race` salvo pasarlo a mano.
- **Vale nombrar `internal/cache`, `internal/core`, `cache.Get/Put`, `~/.cache/kitlegal/cache.db`, `KITLEGAL_CACHE_DIR`, `schema_version`, WAL, `//go:build integration` y `sqlclosecheck`.** No son decisiones de diseño de este spec: son el objeto literal del hito y de la regla de arquitectura R3, que ya vive en `.golangci.yml` y en `internal/arch_test.go`. Omitirlos haría el spec no trazable contra el roadmap.
- **El kernel ya sabe traducir las clases a códigos de salida.** H1 dejó el vocabulario de clases en el dominio y un único punto de traducción a código; H3 produce tres de esas clases y no añade ninguna (FR-033).
