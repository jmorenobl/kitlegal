# Tasks: H3 · `internal/cache`: SQLite con TTL y `--offline`

**Input**: `specs/004-h3-internal-cache-sqlite/` (spec.md, plan.md, research.md, data-model.md,
contracts/, quickstart.md)

**Prerrequisitos**: spec.md y plan.md leídos; `.specify/memory/constitution.md` aplicada; H0, H1 y H2
cerrados (módulo, binario, `Makefile` con `ci` y `test-integration`, `.golangci.yml` con `depguard`,
`forbidigo`, `gosec`, `misspell`, `sqlclosecheck` y `rowserrcheck` ya activos, flujos de integración
continua, umbrales de cobertura, kernel de la línea de órdenes con `schema.Contexto.Offline`,
`schema.ConClase`, `cli.Clasificar`, `internal/httpx` con `Replay` estricto y `internal/arch_test.go` ya
existen). **Ninguna tarea de H3 crea el esqueleto ni los gates: `make ci` existe y pasa desde H0**, y
cada tarea lo deja en verde al terminar.

**Modo**: desatendido. Las tareas se ejecutan y verifican **una a una** por el workflow `hito`, con
guardián de diff por tarea. Por eso cada tarea es una **rebanada vertical** que incluye su comprobación y
deja `make ci` en verde al terminar, y **declara en su propia línea todas las rutas** que va a crear o
modificar.

## Formato: `[ID] [P?] [Story] Descripción con rutas`

- **[P]**: la tarea no comparte fichero con sus vecinas (podrían repartirse si alguien las hiciera a mano;
  en el bucle del workflow van igual en secuencia).
- **[Story]**: historia de usuario del spec — US1 no volver a pedir lo que ya tenemos · US2 lo caducado no
  se sirve · US3 trabajar sin red con lo que ya hay · US4 la caché vive donde la persona usuaria decide ·
  US5 la base sobrevive a las versiones y a las invocaciones simultáneas · US6 el dominio no sabe que hay
  SQLite.
- **[datos]**: la tarea toca material bajo un directorio `testdata`. Solo hay **una**: T007, que no mezcla
  ningún otro trabajo. Es material de test escrito a mano contra el host ficticio `fuente.prueba`, **no**
  una grabación de ninguna fuente real y nunca obtenida de la red (plan.md §«Fixtures», obligación 4).
- **[plataforma]**: la tarea necesita la plataforma remota. Solo hay **una**: T014, la última, único punto
  del hito autorizado a empujar la rama y a abrir la propuesta de cambio. Fusionar nunca es del workflow.

## Batería de verificación por tarea

- **`make ci`** en cada tarea: formato · lint (`depguard`, `forbidigo`, `gosec`, `misspell`,
  `sqlclosecheck`, `rowserrcheck`, `paralleltest`) · tests con detector de carreras · vulnerabilidades ·
  comprobación de esquemas · secretos · integridad de módulos · `go mod tidy -diff`. Desde T010 incluye
  además `test-integration`.
- **Cero red en todo el hito**: ningún test abre una conexión. La única consulta que el adaptador de
  prueba resuelve pasa por `httpx.Replay` sobre un directorio de grabaciones; la segunda, por `Replay`
  sobre un directorio **vacío** (estricto: cualquier petición emitida hace fallar el test nombrándola).
- **`KITLEGAL_RECORD` no aparece en ninguna tarea** (constitución, «Reglas del modo desatendido»): no se
  graba ni un fixture. El único material de reproducción del hito se escribe a mano (T007).
- **Ningún test toca la caché real de la cuenta**: toda base vive en `t.TempDir()`, y donde se ejercita la
  ruta por omisión se redirige `HOME`/`USERPROFILE` con `t.Setenv` (SC-004, SC-009, FR-040).
- **Solo el host ficticio `fuente.prueba`** en tests y fixture (control 18 del plan, obligación 12). La
  única dirección externa admitida es la de la identificación, `https://ventanillalegal.es/bot`, que viaja
  en una cabecera grabada y no se consulta.
- **Sin atajos** al reparar un rojo: ni `//nolint`, ni linter desactivado, ni exclusión nueva, ni test
  saltado, ni error silenciado, ni umbral rebajado (SC-008). El único `t.Skip` admitido en todo el hito es
  la precondición de entorno de los tests de permisos (T010) **fuera de la integración continua**, declarada
  en *Complexity Tracking* del plan; dentro de ella la misma precondición es `t.Fatalf`, de modo que el
  entorno inadecuado se ve en el gate y no en un salto que el registro no imprime.
  Si un test queda legítimamente en rojo porque su implementación pertenece a una tarea posterior, la
  tarea está mal delimitada: se anota, no se parchea.

**Rebanadas verticales, y las ocho excepciones declaradas.** T002, T003, T004, T005, T008 y T010 son
rebanadas completas: cada una escribe su test antes que el código que lo hace pasar y deja `make ci` en
verde por sí sola. En **T010** ese código son los gates que su propio test necesita para valer
(`run.build-tags` en `.golangci.yml` y `test-integration` en la lista de prerrequisitos de `ci`): el
comportamiento que sus tests miden —la reapertura `immutable=1` y el fallo con `-wal` sin `-shm`, filas 12
y 13 del contrato de errores— lo implementa **T004**, que fija la rama entera, de modo que T010 no
necesita tocar ningún fichero de producto fuera de las rutas que declara.

Las otras **ocho** no añaden producto nuevo, por una razón distinta cada una, y también lo dejan en
verde: **T001** crea el paquete `core` y su puerto, que **no contiene ninguna sentencia** (una
interfaz y dos comentarios) y por eso no tiene fichero de test propio —su comprobación es mecánica y ya
existe: `compruebaDominioPuro` de `internal/arch_test.go` y la lista `core` de `depguard` cubren
`internal/core/**` y pasan a vigilar el paquete nuevo desde el primer `make ci`, más la compilación—;
**T006** y **T009** son tests **sobre código ya existente** de las tareas anteriores, que es el único caso
en que test e implementación pueden ir separados; **T007** solo puede tocar material bajo un directorio
`testdata` (regla `[datos]`, que prohíbe mezclarlo con otro trabajo) y añade un fichero que ningún test lee
hasta T008; **T011** son controles y documentación sobre el árbol ya terminado; **T012** y **T013** son
validaciones que se ejecutan sobre una copia desechable o sobre el árbol ya terminado y cuyo único fichero
escrito es la evidencia; **T014** solo publica. Ninguna deja nada a medias para una tarea posterior.

---

## Phase 1: El dominio — el puerto antes que el adaptador

**Objetivo**: que el puerto de caché exista en el dominio antes de que exista nada que lo implemente, que
es el orden que fija la constitución (`core` → adaptador) y lo que hace que el dominio no llegue nunca a
saber que hay SQLite detrás.

- [X] T001 [US6] El puerto de caché en el dominio, sin saber que hay una base de datos detrás: `internal/core/doc.go` con el comentario del paquete `core` («los puertos del dominio»; nace aquí el paquete raíz que el §2 del roadmap reserva a los puertos, que hasta hoy solo tenía el subpaquete `schema`) e `internal/core/cache.go` con la interfaz `Cache` —`Get(ctx context.Context, clave string) (contenido []byte, presente bool, err error)` con el idioma «coma ok» (ausencia, incluida la expirada, es `nil, false, nil`; el fallo declara su clase por `schema.ConClase`) y `Put(ctx context.Context, clave string, contenido []byte, vigencia time.Duration) error`, sin `Close` y con los comentarios que fijan esas garantías—, importando **solo** `context` y `time` y sin nombrar SQLite, `database/sql`, fichero alguno ni el paquete adaptador de la caché; ninguno de los dos ficheros contiene sentencias, y por eso su comprobación es la mecánica que ya existe y que desde este `make ci` pasa a cubrir el paquete nuevo: `compruebaDominioPuro` de `TestArquitectura` (regla R1, cuyo `paquetesBajo` incluye el propio prefijo del paquete vigilado) y la lista `core` de `depguard`, más la compilación; la cobertura del dominio no se altera porque no hay sentencias nuevas que cubrir (FR-001, FR-038, SC-013, US6 escenario 1, D1).

**Checkpoint**: el dominio declara qué es una caché. Nada la implementa todavía y el binario no cambia.

---

## Phase 2: El adaptador, de dentro afuera — errores, ruta, apertura, entradas

**Objetivo**: construir `internal/cache` en el orden en que sus piezas dependen unas de otras, con el test
de tabla antes que el código en cada paso: primero el error que declara su clase, luego dónde vive la base
de datos, luego cómo se abre y se migra, y por último qué se guarda y qué se lee.

- [X] T002 [P] [US6] El paquete y el único error que produce, con su clase y sus mensajes: `internal/cache/doc.go` con el comentario de paquete que declara qué garantiza y qué **no** hace (no lee ninguna bandera —el modo lo pone quien construye, a partir de `schema.Contexto.Offline`—, no borra la base por su cuenta, no cifra, no interpreta ni la clave ni el contenido, no fija ninguna vigencia por omisión); `internal/cache/errores.go` con `Error{Operacion, Ruta, Origen, Clave, Causa}` más su clase privada y los constructores internos por clase, `Error() string` con mensaje en español que nombra lo que la tabla del contrato de errores §3 exige en cada situación (ruta y origen —opción `ConDirectorio` o variable `KITLEGAL_CACHE_DIR`—, versión esperada y encontrada, la clave y que la invocación es de solo lectura, la operación tras cerrar, y la ruta con `cache.db-wal` y `cache.db-shm` en el fallo de la fila 13, cuyo constructor interno nace aquí y usa T004), `Unwrap()` que expone `Causa` y `Clase() schema.Clase` que lo hace `schema.ConClase`, sin `panic` en ninguna ruta y sin decidir ningún código de salida por su cuenta; e `internal/cache/errores_test.go` con `TestErrorMensajes` (una fila por forma de mensaje del contrato de errores **§6**, la fila 13 incluida: el fallo de WAL sin memoria compartida nombra la ruta, `cache.db-wal` y `cache.db-shm`), `TestErrorEnvueltoConservaLaClase` (envolver con `fmt.Errorf("%w")` no cambia lo que `cli.Clasificar` devuelve) y `TestErrorNuloOCero` (`Error()` no vacío y `Clase()` «inesperado» sobre el valor cero); comprobando además al primer `make lint` el supuesto S2, para lo que declara también `.golangci.yml` **acotado a las palabras nuevas bajo `misspell.ignore-rules`** que el español dispare como falso positivo —ninguna regla, ningún linter, ninguna exclusión y nunca una supresión—, y sin tocarlo si no hace falta (FR-033, FR-034, FR-035, obligaciones 3 y 14 del plan, D9, D15).

- [X] T003 [US4] Dónde vive la base de datos y con qué opciones se construye el cliente, antes de abrir nada: `internal/cache/ruta.go` con la resolución de la ruta efectiva en el orden opción > `KITLEGAL_CACHE_DIR` > `~/.cache/kitlegal` (la variable se lee **una sola vez**, al construir), el origen de la ruta para los mensajes, la validación del valor de la ruta que vale en **cualquier** modo (cadena vacía → «argumentos» 2; ruta que existe y no es un directorio → 2, nombrando origen y ruta; nunca una caída silenciosa a la ruta por omisión) y la **regla «inexistente»** —`esInexistente(err)` es verdadero si y solo si `errors.Is(err, fs.ErrNotExist) || errors.Is(err, syscall.ENOTDIR)`, definida una vez aquí y única razón por la que el paquete importa `syscall`—; `internal/cache/cliente.go` con `const VariableDirectorio = "KITLEGAL_CACHE_DIR"`, el tipo `Cliente`, `type Opcion func(*configuracion) error` y las cuatro opciones `ConDirectorio`, `SoloLectura`, `ConReloj` y `ConRegistrador` (reloj por omisión `time.Now`, registrador por omisión que descarta), `New(ctx context.Context, opciones ...Opcion) (*Cliente, error)` que valida las opciones en orden —la primera inválida es «argumentos» (2)—, resuelve y valida la ruta, respeta `ctx` (cancelado o vencido → «fuente no disponible», 4) y **todavía no abre ninguna base de datos**, y `Close()` idempotente sobre ese cliente; e `internal/cache/ruta_test.go` con `TestEsInexistente` (tabla del predicado sobre errores reales de `os.Stat` en `t.TempDir()`: `fs.ErrNotExist` → true, `syscall.ENOTDIR` → true, `fs.ErrPermission` → false, `nil` → false, error ajeno → false), `TestRutaEfectivaPrecedencia` (`t.Setenv`: opción > variable > omisión, y la variable leída una sola vez) y `TestRutaInservible` (tabla × modo: vacía por opción, vacía por variable y fichero como directorio, los tres → 2 nombrando origen y ruta, igual en modo normal y en solo lectura), más `TestNewRechazaOpcionesInvalidas` en `internal/cache/cliente_test.go` (`ConDirectorio("")`, `ConReloj(nil)`, `ConRegistrador(nil)` → 2 nombrando la opción) (FR-003, FR-004, FR-009, FR-015, FR-019, FR-020, FR-022, FR-023, US4 escenarios 3 y 4, D2, D3).

- [X] T004 [US5] Abrir, crear y migrar la base de datos —y la única dependencia nueva del hito—: `internal/cache/abrir.go` con la apertura normal (`MkdirAll(dir, 0o700)` cuando el directorio no existe, con el fallo de creación o de escritura como «argumentos» 2 nombrando origen y ruta; `OpenFile(cache.db, O_RDWR|O_CREATE, 0o600)` y cierre, para que el fichero nazca con permisos reservados y no con los que SQLite le pondría; DSN `file:<filepath.Clean(ruta)>?_pragma=busy_timeout(5000)&_pragma=journal_mode(WAL)&_pragma=synchronous(FULL)&_txlock=immediate` y `SetMaxOpenConns(1)`) y la apertura de solo lectura (`Stat` de `cache.db` con la regla «inexistente» de T003 → cliente **sin base**, cualquier otro fallo → «inesperado» 1 y nunca una ausencia falsa; DSN `file:<ruta>?mode=ro&_pragma=busy_timeout(5000)&_pragma=query_only(1)` sin `immutable`, para ver lo que otra invocación ya confirmó en el WAL, y la rama del directorio que no admite crear `-shm`, que **es una sola y se dispara ante los dos códigos que produce ese caso**: si la primera consulta de solo lectura falla con `*sqlite.Error` de código `SQLITE_READONLY_DIRECTORY` (1544) **o** `SQLITE_CANTOPEN` (14) —1544 sin `-wal`, 14 cuando `-wal` existe y falta `-shm`, research D5 y sonda 3 B, D y E—, se comprueba `<ruta>-wal`: **ausente** → cerrar y reabrir con `&immutable=1`, y si esa reapertura o su primera consulta vuelven a fallar, «inesperado» (1) nombrando la ruta; **presente** → «inesperado» (1) nombrando la ruta, `cache.db-wal` y `cache.db-shm`, que es la forma de mensaje que `TestErrorMensajes` de T002 fija y la fila 13 del contrato de errores exige; `SQLITE_NOTADB` (26) o cualquier otro código → «inesperado» (1) sin degradar a ausencia), sin exponer la conexión ni ninguna forma de ejecutar SQL desde fuera del paquete; `internal/cache/migraciones/0001_entradas.sql` con el esquema v1 (`schema_version(version INTEGER PRIMARY KEY, aplicada_en TEXT NOT NULL) STRICT` y `entradas(clave TEXT PRIMARY KEY NOT NULL, contenido BLOB NOT NULL, expira_en INTEGER NOT NULL) STRICT`, sin índices adicionales); `internal/cache/migraciones.go` con el `embed.FS` (`//go:embed migraciones/*.sql`), la versión conocida derivada del número de ficheros, la versión registrada (`0` si no existe la tabla, `COALESCE(MAX(version), 0)` si existe), la aplicación de cada migración pendiente en una **transacción inmediata** que relee la versión dentro de la transacción y la registra antes de confirmar (atómica: una interrupción deja la versión anterior), el rechazo con «inesperado» (1) de una versión mayor que la conocida diciendo esperada y encontrada **sin modificar ni borrar el fichero**, el mismo trato para un fichero que no es una base utilizable, y la comprobación sin aplicar nada en modo de solo lectura (`0` → sin esquema, `== conocida` → lectura, otra → 1); `internal/cache/cliente.go`, que T003 dejó sin base de datos, pasa a usar lo anterior: el tipo `Cliente` gana los campos `db *sql.DB` (nulo cuando el modo de solo lectura no encuentra `cache.db`) y `versionEsquema`, el `New` que «todavía no abre ninguna base de datos» llama tras resolver la ruta a la apertura de `internal/cache/abrir.go` —que crea, abre y migra en modo normal y solo comprueba en solo lectura—, y `Close` cierra la conexión si la hay y sigue siendo idempotente y sin error al cerrar dos veces o un cliente sin base; `go.mod` y `go.sum` con la **única** dependencia nueva `modernc.org/sqlite v1.58.0` y los indirectos que `go mod tidy` escriba, comprobando aquí el supuesto S1 (obligación 1: `go mod tidy -diff` limpio, `go mod graph` sin módulos más allá de los que declara el `go.mod` del driver, `TestDependenciasDelBinario` en verde **sin tocar su lista**, y la diferencia frente a la predicción de S1 anotada para el cierre del hito); `internal/cache/abrir_test.go` con `TestAbrirCreaDirectorioYFicheroConPermisosReservados` (`0700` y `0600`), `TestAbrirAplicaLosPragma` (`journal_mode = wal`, `synchronous = 2`, `busy_timeout = 5000`, y `query_only = 1` en solo lectura) y `TestFicheroInutilizable` (normal y solo lectura → 1, SHA-256 idéntico, no se borra); `internal/cache/migraciones_test.go` con `TestMigracionesEmbebidasBienFormadas` (numeración contigua desde `0001`, versión conocida = número de ficheros), `TestMigracionesIdempotentes` (dos aperturas: una fila en `schema_version`, versión 1), `TestVersionMayorQueLaConocida` (`INSERT INTO schema_version VALUES (99, …)`; normal y solo lectura → 1 con esperada y encontrada; SHA-256 igual), `TestMigracionAtomica` (tabla `entradas` ajena preexistente: `New` → 1, sin `schema_version` y con la tabla ajena intacta), `TestSoloLecturaNoMigra` (base en versión 0 sigue en 0) y `TestDosClientesMigranUnaVez` (dos `New` concurrentes sobre un directorio vacío: cero errores, una sola versión); y en `internal/cache/cliente_test.go` `TestCloseEsIdempotente` (cerrar dos veces no falla) y `TestNewPorOmisionUsaElDirectorioDeLaCuenta` (`t.Setenv` de `HOME` y `USERPROFILE` a un temporal: `cache.db` aparece en `<HOME>/.cache/kitlegal/` y en ningún otro sitio), todos con la base en `t.TempDir()` y sin ninguna supresión de `sqlclosecheck` ni `rowserrcheck` (FR-002, FR-021, FR-024 a FR-031, FR-043, SC-005, SC-006, US4 escenario 1, US5 escenarios 1 a 3, D4, D5, D6, D7).

- [X] T005 [US2] Guardar y recuperar, con la vigencia decidida por un reloj que el test adelanta: `internal/cache/entradas.go` con `Get` (clave vacía → 2; cliente cerrado → 1; contexto cancelado o vencido → 4; `QueryRowContext(...).Scan` de una sola fila; vigente ⇔ `reloj().Before(time.Unix(0, expira_en))`, de modo que en el instante exacto de expiración y después el resultado es ausencia; en modo normal la ausencia —incluida la expirada— es `nil, false, nil` y la fila se conserva; en modo de solo lectura la ausencia, la base sin esquema y el cliente sin base son «fuente no disponible» (4) nombrando la clave; presente devuelve el contenido byte a byte, y un BLOB vacío se normaliza a `[]byte{}` no nulo para distinguirlo de la ausencia) y `Put` (clave vacía o vigencia `<= 0` → 2 **sin escribir nada**; en modo de solo lectura → «inesperado» 1 nombrando la clave y sin escribir; en modo normal, upsert en una sola sentencia `ExecContext` que sustituye por completo contenido y vigencia, con `expira_en` calculado como `reloj().Add(vigencia).UnixNano()` y `nil` guardado como BLOB vacío), más `var _ core.Cache = (*Cliente)(nil)` que fija que el adaptador implementa el puerto; `internal/cache/entradas_test.go` con `TestPutYGetIntegros` (tabla: texto, binario con bytes altos, cero bytes, 1 MiB y clave larga), `TestGetAusente`, `TestPutSustituyeLaEntrada`, `TestClavesIndependientes`, `TestClaveVacia`, `TestVigenciaInvalida` (0 y negativa → 2, nada escrito), `TestExpiracionConRelojInyectado` (las **tres** posiciones del borde —antes: presente; instante exacto: ausencia; después: ausencia— con `ConReloj` y sin esperar tiempo real: la duración del test no depende de la vigencia usada), `TestPutSobreEntradaCaducada` (la nueva es vigente y la caducada no reaparece), `TestContextoCancelado` (`New`, `Get` y `Put` → 4), `TestSoloLecturaAusenciaEsFuenteNoDisponible`, `TestSoloLecturaExpiradaEsFuenteNoDisponible` y `TestSoloLecturaPutFalla` (→ 1, SHA-256 de `cache.db` idéntico); y las comprobaciones de solo lectura que solo se pueden hacer leyendo, en `internal/cache/abrir_test.go` —`TestSoloLecturaSinBaseNoCreaNada` (tabla: directorio vacío, directorio inexistente y directorio cuyo padre es un fichero; listado del padre idéntico antes y después, `Get` → 4), `TestSoloLecturaBaseVacia` (fichero de 0 bytes: ausencia, sin migrar, fichero intacto), `TestSoloLecturaNoModificaElFichero` (SHA-256 antes y después de un `Get` presente y de uno ausente, excluyendo `-wal` y `-shm`) y `TestSoloLecturaVeLoConfirmadoEnElWAL` (un escritor confirma sin punto de control; el lector `ro` ve lo confirmado y no lo pendiente)—, en `internal/cache/ruta_test.go` `TestDirectorioNoCreable` (directorio `<fichero>/sub` cuyo padre es un fichero: subtest `normal` → 2 nombrando origen y ruta; subtest `solo-lectura` → `New` sin error, cliente sin base, `Get` → **4 y nunca 2**, con el fichero padre intacto) y en `internal/cache/cliente_test.go` `TestOperacionTrasCierre` (`Get` y `Put` tras `Close` → 1, sin tocar el disco) (FR-003 a FR-018, FR-042, SC-002, SC-003, SC-010, US1 escenarios 4 y 5, US2 completa, US3 escenarios 4 y 6, D8).

- [X] T006 [US5] Dos clientes a la vez y la superficie cerrada, sobre el código ya existente: `internal/cache/concurrencia_test.go` con `TestDosClientesEnElMismoProceso` (dos subtests de un solo nivel, sin anidar —el escenario 12 del quickstart cuenta sus líneas—, `normal-normal` y `normal-solo-lectura`: el escritor confirma N entradas y avisa por canal; el lector las encuentra **todas**, ninguna a medias, cero fallos por bloqueo y cero ausencias falsas) y `TestMismaClaveDosEscritores` (la última escritura completa gana y el contenido leído es uno de los dos, nunca una mezcla); `internal/cache/superficie_test.go` con `TestSuperficieExportada`, que recorre con `go/parser` y `go/ast` todas las declaraciones exportadas del paquete y falla si alguna nombra un selector de `sql` o de `sqlite` o si aparece una declaración exportada fuera de la lista del contrato (`VariableDirectorio`, `Cliente`, `New`, `Opcion`, las cuatro opciones, `Get`, `Put`, `Close`, `Error` y sus tres métodos), de modo que la imposibilidad de ejecutar SQL arbitrario desde fuera sea **por construcción** y no solo por control mecánico; y `TestClienteDesdeVariasGoroutines` en `internal/cache/cliente_test.go` (un solo cliente, N goroutines alternando `Get` y `Put`), todo bajo el detector de carreras que `make test` ya activa; la tarea **no añade producto**: comprueba el de T003 a T005 (FR-005, FR-007, FR-032, SC-007, SC-013, US5 escenario 4, US6 escenario 1, D2, D10).

**Checkpoint**: la caché está completa y comprobada como paquete: guarda, sirve lo vigente, ignora lo
caducado, no escribe en solo lectura, sobrevive a dos clientes a la vez y no deja ver la base de datos.
Nada de esto es todavía observable desde ninguna invocación.

---

## Phase 3: El criterio de aceptación — la caché con el kernel y sin red

**Objetivo**: demostrar lo que el hito promete —dos consultas idénticas, una sola petición— y que
`--offline` significa por fin algo, con el kernel invocado en proceso y sin que ningún binario conozca
nada de esto.

- [X] T007 [P] [datos] [US1] La única grabación del hito, **escrita a mano** contra el host ficticio `http://fuente.prueba` —material de test, **no** una grabación de ninguna fuente real y nunca obtenida de la red (constitución, «Reglas del modo desatendido»; `KITLEGAL_RECORD` no se ejecuta en ninguna tarea)—, con el nombre que produce la regla del contrato de grabación de H2 §2 y el contenido de su §1: `internal/cache/testdata/reproduccion/prueba/GET_http_fuente.prueba_norma.json` con `"formato": 1`, `grabado_en` fijo `2026-09-12T00:00:00Z`, `peticion` `GET http://fuente.prueba/norma` con `cabeceras` que llevan **solo** la identificación `kitlegal/dev (+https://ventanillalegal.es/bot)`, respuesta `200` con `Content-Length: 24`, `Content-Type: text/plain; charset=utf-8` y una cabecera `Date` fija, y `cuerpo` `<norma>contenido</norma>`; ninguna otra ruta y ningún otro fichero —la lista es cerrada (plan.md §«Fixtures»)—, ningún fichero fuera de ese árbol de material de test del propio paquete, y ninguna tarea de H3 crea el árbol de fixtures por fuente de la raíz del repositorio ni toca el directorio de esquemas de la raíz (punto 2 de la Definition of Done, FR-046, obligación 4 del plan, D12).

- [X] T008 [US1] El adaptador de prueba con el kernel en proceso: la segunda consulta no toca la red y `--offline` adquiere su significado: `internal/cache/adaptador_test.go`, en `package cache_test` y por tanto **material de test —no se crea ningún adaptador en el directorio reservado a los adaptadores de fuente, que el hito no abre, ni se registra el applet en ningún binario—**, con los tipos `adaptadorDePrueba` (applet `prueba`), `argumentosConsultar` (verbo `consultar`), `argumentosGuardar` (verbo `guardar`) y `cuerpoDeLaFuente`, que aplican el patrón del contrato del puerto §8 —leer de `schema.Contexto.Offline` el modo y pasarlo como `cache.SoloLectura()`, mirar la caché antes de pedir, pedir por `httpx.Replay` (nunca `net/http`, que `depguard` deniega por prefijo también en los tests) y guardar con la vigencia— y `TestAdaptadorDePruebaConElKernel`, invocando `app.Main` en proceso sobre un registro de pruebas, con los subtests `primera-consulta` (la consulta llega a la reproducción sobre el directorio de T007, el resultado sale al sobre y queda guardado con su vigencia), `segunda-consulta-sin-red` (misma clave dentro de la vigencia, **ejecutada contra un `httpx.Replay` estricto construido sobre un directorio vacío**, de modo que cualquier petición emitida termine en un fallo ruidoso que la nombra: código **0**, `origen: cache` y cuerpo **idéntico byte a byte** al de la primera, y el subtest pasa precisamente porque no se emitió ninguna petición), `sin-cache-la-reproduccion-falla` (**el control negativo**: caché **vacía** y el mismo `httpx.Replay` estricto sobre un directorio vacío, de modo que la consulta sí se emite y el subtest termina con código **1** y un mensaje que nombra `GET http://fuente.prueba/norma`, que es lo que demuestra que el anterior pasa *porque* no pide), `offline-presente` (código 0 y cero peticiones), `offline-ausente` y `offline-expirada` (sembrada con `ConReloj`, sin esperar; las dos → código **4** y cero peticiones), `offline-sin-base` y `offline-directorio-inexistente` (código **4** y nunca 2, con el listado del directorio y el SHA-256 de `cache.db` —o su ausencia— idénticos antes y después, excluyendo `cache.db-wal` y `cache.db-shm`), `offline-guardar` (un intento de escribir sobre el cliente de solo lectura → código **1**, nada escrito) —**nueve** subtests, los mismos que el inventario del plan y D12, **todos de un solo nivel** (nueve `t.Run` sin anidar, porque los escenarios 2 y 4 del quickstart cuentan líneas de subprueba), de los que **seis y solo seis** empiezan por `offline-`, que es lo que filtra el escenario 4 del quickstart: ningún subtest nuevo puede llevar ese prefijo sin entrar antes en el inventario y en el esperado de ese escenario—; US3 escenario 5 y FR-014 (sin `--offline`, la ausencia no es un fallo: la consulta sigue su curso hacia la fuente y el código no cambia) los acredita `primera-consulta`, que parte de una caché vacía, llega a la reproducción y termina en 0 con `origen: fuente`, de modo que no se añade ningún subtest que lo repita; y `TestAdaptadorConVariableDeEntorno` (`t.Setenv`: la base aparece bajo `KITLEGAL_CACHE_DIR` y no bajo `HOME`; valor vacío → 2; ruta que es un fichero → 2); y declara también `.golangci.yml` **acotado a una palabra nueva en la lista `ignore-rules` de `misspell`** —`reproduccion`, que el diccionario inglés lee como «reproduction» y que la tarea no puede reescribir porque la fijan dos contratos: el nombre del subtest `sin-cache-la-reproduccion-falla` del inventario de tests del plan y del escenario 2 del quickstart, y el tramo del directorio de grabaciones que el contrato de grabación de H2 da a la reproducción de cada fuente—, con su comentario y ninguna otra línea: ninguna regla, ningún linter, ninguna exclusión y nunca una supresión (contingencia S2, precedente de H2 T010) (FR-014, FR-016, FR-017, FR-044 a FR-047, SC-001, SC-003, SC-004, US1 escenarios 1 a 3, US3 escenarios 1 a 5, D12).

- [X] T009 [US6] Ninguna ruta de fallo sin clase, comprobada sobre el error real: `internal/cache/errores_test.go` añade `TestClasesDeError` con las **trece** subpruebas deterministas de la tabla cerrada del contrato de errores §3 —filas 1 a 11, 14 y 15: clave vacía, vigencia `<= 0`, opción inválida, variable presente y vacía, ruta que existe y no es un directorio, directorio que no se puede crear en modo normal, ruta por omisión indeterminable, ausencia o expirada en solo lectura (incluidos `cache.db` y directorio inexistentes), versión de esquema desconocida, fichero inutilizable, escritura en solo lectura, operación tras `Close` y contexto cancelado o vencido—, provocando cada situación con un `t.TempDir()` y comprobando la clase que `cli.Clasificar` devuelve y el código que `cli.CodigoSalida` produce sobre el error **tal como sale del paquete** y también envuelto con `%w` —las dos comprobaciones van **dentro de la misma subprueba**: trece `t.Run` y **ninguna subprueba anidada**, porque el escenario 8 del quickstart cuenta exactamente trece líneas de primer nivel—, y comprobando además que ninguna situación produce las clases «no encontrado» (3), «límite o TOS» (5) ni «requiere identidad humana» (6) y que ninguna termina en `panic`; las filas 12 y 13, que dependen de los permisos del sistema de ficheros, las cubren los tests de T010; la tarea cubre código ya existente de T002 a T008 y **no añade producto** (FR-033, FR-034, SC-011, punto 5 de la Definition of Done, D9).

**Checkpoint**: el criterio de aceptación del hito está demostrado —dos consultas idénticas, una sola
petición— y las quince situaciones de fallo tienen código de salida comprobado salvo las dos que exigen
permisos reales.

---

## Phase 4: Integración, controles y cierre

**Objetivo**: que lo que depende del entorno se pruebe donde el entorno lo hace valer y entre como gate de
integración continua, que la regla R3 pase a tener dueño y que el hito se cierre con la medida hecha.

- [X] T010 [US5] Lo que solo se puede comprobar con permisos reales y con dos procesos, y su puerta en integración continua: `internal/cache/integracion_test.go`, en `package cache_test` con `//go:build integration` y con toda base en `t.TempDir()`, con `TestIntegracionDosProcesos` (subtests `escritor-padre-lector-solo-lectura-hijo` y `escritor-hijo-lector-padre`: el segundo proceso es el **propio binario de test** relanzado con `exec.CommandContext(ctx, os.Args[0], "-test.run=^TestProcesoAuxiliar$")` y una variable de entorno que fija el papel, sin añadir ningún `package main`; el lector encuentra **todas** las entradas que el escritor confirmó, ninguna a medias y sin fallar por bloqueo), `TestIntegracionDirectorioNoEscribible` (`0500`: subtest `solo-lectura-lee` → código 0 gracias a la reapertura con `immutable=1` que T004 implementa ante `SQLITE_READONLY_DIRECTORY` (1544) sin `-wal`; subtest `normal-argumentos` → 2), `TestIntegracionDirectorioDenegado` (`0000`: solo lectura → **1** y nunca una ausencia falsa; normal → 2), `TestIntegracionWALSinMemoriaCompartida` (`cache.db` y `cache.db-wal` copiados a un directorio `0500`: la primera consulta falla con `SQLITE_CANTOPEN` (14) y, por la rama única de T004, el resultado es 1 nombrando la ruta y los dos auxiliares) y `TestProcesoAuxiliar` (retorna sin hacer nada cuando falta la variable de papel), todos con la precondición comprobada de que el sistema de ficheros hace valer los permisos (crear un directorio `0000` y exigir que un `os.Stat` dentro devuelva `fs.ErrPermission`), **resuelta según dónde corra el proceso**: en integración continua —variable de entorno `CI` no vacía, que GitHub Actions exporta como `CI=true`— la precondición incumplida es `t.Fatalf` nombrando la causa (un ejecutor privilegiado deja el trabajo `ci` en rojo, que es lo que hace valer el supuesto S3 **mecánicamente**, sin depender de que el registro muestre un salto que la receta no imprime); fuera de ella es `t.Skip` nombrando la misma causa —único `t.Skip` del hito, que nunca se produce donde la aserción tendría sentido—; y con la restauración de permisos registrada en `t.Cleanup` **después** de `t.TempDir()`, para que su `RemoveAll` no falle; y **sin ninguna supresión**: el ejecutable llega a `exec.CommandContext` como **parámetro** de una función y el único argumento es un literal (G204), toda ruta pasa por `filepath.Clean` antes de `os.OpenFile`, `os.ReadFile` y `os.WriteFile` (G304) y los ficheros de test se escriben con `0o600` (G306); `.golangci.yml` añade `run.build-tags: [integration]` para que el análisis estático —`sqlclosecheck` y `rowserrcheck` incluidos— alcance el fichero etiquetado; y `Makefile` añade `test-integration` a la lista de prerrequisitos de `ci` **inmediatamente después de `test` y antes de `vuln`**, de modo que la línea quede exactamente `ci: fmt-check lint test test-integration vuln schema-check secrets mod-verify mod-tidy-check` —que es lo que el escenario 11 del quickstart compara literalmente y lo que T013 deja como evidencia—, **sin cambiar ninguna receta**, para que un test de integración en rojo o un fichero etiquetado que no compile hagan fallar el gate (FR-015, FR-022, FR-032, FR-039 a FR-041, SC-004, SC-007, SC-009, SC-011 filas 12 y 13, US5 escenario 5, obligaciones 2, 6 y 7 del plan, D5, D11).

- [ ] T011 [US6] La regla R3 estrena dueño y el binario se queda como estaba, con el pendiente anotado para H4: `internal/arch_test.go` actualiza el comentario de la subprueba R3 para dejar escrito que el paquete de la caché es su **primer dueño real** y por qué `duenoObligatorio` sigue en `false` —la exigencia pide los tres dueños y `store` y `graph` llegan en H12 y H17 (FR-037)—, y añade `TestElBinarioNoEnlazaCache`, que comprueba con `go list -deps` sobre el paquete `main` del binario (la misma orden del escenario 10 del quickstart) que el binario distribuido **no** enlaza el paquete de la caché ni ningún módulo del driver de SQLite, con el comentario que declara que es temporal y que H4 lo retirará al enlazarlo, dejando `TestDependenciasDelBinario` y su lista intactos; `.golangci.yml` actualiza **solo los comentarios** de la lista `sql` para dejar constancia de que la regla ya tiene dueño, sin cambiar ninguna regla, sin activar ningún linter y sin añadir ninguna exclusión ni `//nolint`; y `docs/PENDIENTES.md` anota para H4 que hay que retirar `TestElBinarioNoEnlazaCache` y ampliar `modulosDelBinario` con los módulos que esa misma orden muestre cuando el binario enlace por fin el paquete de la caché, **justificados uno a uno** por escrito y **sin copiar ninguna lista escrita a mano** (el `go.mod` del driver declara módulos que `go mod tidy` no incorpora al grafo), y que la clave de caché por fuente y su vigencia salen de la fuente (`Source.TTL()`) (FR-036, FR-037, FR-044, SC-012, SC-013, punto 3 de la Definition of Done, obligación 10 del plan, D13).

- [ ] T012 Demostrar que los controles fallan ante cada intento de saltarse una regla, **sobre una copia desechable del árbol fuera del repositorio** y nunca sobre el árbol de trabajo: ejecutar el escenario 9 de `quickstart.md` en sus cuatro variantes —`database/sql` importado desde un paquete que no es ninguno de los tres de almacenamiento (caché, almacén y grafo) que la regla R3 autoriza (falla en `depguard` **y** en la subprueba R3, nombrando la regla), el paquete de la caché importado desde un subpaquete del dominio (falla en `depguard` **y** en `compruebaDominioPuro`, regla R1), un `*sql.Rows` sin cerrar en el fichero etiquetado `integration` (que `sqlclosecheck` solo alcanza gracias a `run.build-tags`) y `net/http` importado en un fichero de test del paquete de la caché (regla R2, por prefijo)—, comprobando que cada una hace fallar `make ci` nombrando la regla incumplida, y dejar la salida literal de las cuatro como evidencia en `specs/004-h3-internal-cache-sqlite/gates/evidencia-sc013.md`; la tarea no modifica ningún fichero del árbol (FR-038, FR-039, SC-008, SC-013, obligación 8 del plan).

- [ ] T013 Cierre del hito: validación agregada, medidas diferenciales y umbrales: ejecutar los escenarios 1, 4, 7, 10, 11 y 12 de `quickstart.md` y dejar su resultado en `specs/004-h3-internal-cache-sqlite/gates/evidencia-cierre.md` —`make ci` en verde de principio a fin con `-race`, con `test-integration` dentro y **sin ninguna supresión nueva** (el diff de la configuración del lint frente a `main` añade `run.build-tags`, comentarios y, como mucho, las palabras que el supuesto S2 obligara a ignorar en `misspell`: ninguna regla, ningún linter y ninguna exclusión); umbrales de cobertura respetados sin rebajar ninguno (global ≥ 70 %, el del dominio ≥ 85 %, con los tests de integración fuera de `coverage.out`); medida **diferencial** de `ls -laR ~/.cache/kitlegal` idéntica antes y después de la suite y un `TMPDIR` desechable vacío al terminar los tests de integración, de modo que ningún fichero quede fuera de `t.TempDir()`; conjunto de verbos del binario distribuido idéntico al de H2 y `go list -deps` sobre el paquete `main` del binario sin el paquete de la caché ni ningún módulo del driver de SQLite; y la orden **entera** de los prerrequisitos del quickstart que extrae toda dirección `http(s)://` de los ficheros de test y del material de reproducción del paquete de la caché, confirmando «solo direcciones ficticias» (`fuente.prueba`, más la identificación que se envía y no se consulta)— (SC-003, SC-008, SC-009, SC-012, SC-014, puntos 1, 2 y 9 de la Definition of Done, obligaciones 9, 11 y 12 del plan).

- [ ] T014 [plataforma] Publicar la rama del hito y abrir la propuesta de cambio: `git push -u origin h3-internal-cache-sqlite` y `gh pr create --base main --head h3-internal-cache-sqlite` con el cuerpo tomado de `specs/004-h3-internal-cache-sqlite/gates/pr-h3.md`, que la propia tarea escribe con la plantilla del ritual (objetivo, alcance, controles añadidos, decisiones, pendientes y la justificación de la dependencia nueva que exige la constitución §V), adjuntando como evidencia `specs/004-h3-internal-cache-sqlite/gates/evidencia-sc013.md` y `specs/004-h3-internal-cache-sqlite/gates/evidencia-cierre.md` y anotando la diferencia, si la hay, entre los indirectos que `go mod tidy` escribió y la predicción del supuesto S1; y leer después los estados de la integración continua y de Codecov con `gh pr checks`, comprobando el supuesto S3 —**el trabajo `ci` de la propuesta de cambio está en verde con `test-integration` dentro**, que es lo que `gh pr checks` y `gh run list` sí muestran: la precondición de permisos de T010 termina en `t.Fatalf` cuando el proceso corre en integración continua, de modo que un ejecutor privilegiado dejaría el trabajo en rojo y el verde es la comprobación; no se busca ningún `PASS` ni `SKIP` en el registro, porque la receta de `test-integration` es contrato de H0, corre sin `-v` y no los imprime— y el supuesto S5 (duración de `ci` de la propuesta de cambio, con `test-integration` dentro), y dejar su resultado en `specs/004-h3-internal-cache-sqlite/gates/evidencia-plataforma.md`; **no** fusiona, **no** empuja a `main`, **no** fuerza y **no** crea etiquetas: la fusión (squash-merge) es siempre humana y el gancho `pre-push` la rechazaría igualmente (punto 1 de la Definition of Done, §6 del roadmap, ADR 0007, obligación 13 del plan).

---

## Definition of Done: qué punto cubre qué tarea

| # | Punto | Dónde se cumple |
|---|---|---|
| 1 | `make ci` en verde (`fmt`, `lint`, `test` con `-race`, `vuln`, `schema-check`) | Todas las tareas; T010 lo amplía con `test-integration`; T013 lo cierra medido |
| 2 | Tests unitarios offline para el código nuevo; fixtures | T002 a T010; el único material de reproducción, en T007 `[datos]`, escrito a mano y sin red |
| 3 | Sin `net/http`, `os.Exit`, `fmt.Print*` fuera de los paquetes autorizados | T011 (comentarios de la lista `sql`), T012 (demostración de las cuatro violaciones) |
| 4 | Salida de applet validada contra `schemas/*.json` | **No aplica**: H3 no añade ningún applet ni ninguna salida de applet (spec, *Fuera de alcance*; SC-012). Ninguna tarea toca `schemas/` |
| 5 | Errores tipados con exit code estable; ningún `panic` | T002 (el tipo y su clase), T009 (las trece situaciones deterministas), T010 (las dos que exigen permisos) |
| 6 | Si cambia un comportamiento visible: e2e (`testscript`) y `CHANGELOG.md` | **No aplica**: H3 no cambia ningún comportamiento visible del binario (FR-044, SC-012), y `testscript` solo observa binarios. Lo que `--offline` gana se comprueba con el kernel **en proceso** en T008, que es la primera tarea que puede dejarlo en verde. Si alguna tarea detectara un cambio visible, la entrada en `CHANGELOG.md` y el guion pasarían a ser obligatorios y se anotaría en `gates/tarea-Tnnn.md` |
| 7 | Si cambia una decisión de arquitectura: ADR nuevo | **No aplica**: SQLite sin cgo ya está en `docs/ADR/0002-sqlite-sin-cgo.md` y el patrón *Repository* para `cache`, `store` y `graph` en `docs/ROADMAP.md` §2; H3 no se aparta de ninguno (spec, *Fuera de alcance*) |
| 8 | Si toca una fuente externa: `docs/SOURCES.md` y `scripts/verify-sources.sh` | **No aplica**: H3 no consulta ninguna fuente real (cero red en todo el hito); no hay fuente que declarar ni verificar, y editar esa tabla es decisión humana (constitución, capa 3) |
| 9 | Cobertura `internal/core/**` ≥ 85 % y global ≥ 70 % | T013 (medida y anotada, sin rebajar ningún umbral); T001 no altera `internal/core` porque no añade sentencias |
| 10 | Evals de skill (desde H5) | **No aplica**: H3 no entrega ni modifica ninguna skill; es un hito de fundación (principio VIII) |
| 11 | Dimensión territorial | **No aplica**: H3 no la tiene; el host del fixture es `fuente.prueba` |

## Trazabilidad: requisito → tarea

| Requisito | Tarea |
|---|---|
| FR-001, FR-038 (puerto en el dominio; `core` no importa el adaptador) | T001, y T012 lo demuestra |
| FR-002 (`internal/cache` sobre SQLite, único que abre y migra) | T004 |
| FR-003 (contexto obligatorio en toda operación) | T003 (`New`), T005 (`Get`/`Put`) |
| FR-004 (`Close` idempotente; nada sin cerrar) | T003, T004, T005 |
| FR-005 (superficie sin conexión ni SQL arbitrario) | T006 |
| FR-006, FR-012 (escritura con vigencia; contenido opaco e íntegro) | T005 |
| FR-007 (sustitución completa de la entrada) | T005, T006 |
| FR-008, FR-042 (vigencia y su borde; reloj inyectado) | T005 |
| FR-009 (reloj inyectable al construir) | T003 (opción), T005 (test) |
| FR-010, FR-011 (vigencia `<= 0` y clave vacía → 2) | T005 |
| FR-013, FR-014 (tres resultados; la ausencia no altera la invocación) | T005, T008 |
| FR-015 (modo de solo lectura: no crea, no migra, no escribe, ve lo confirmado) | T003, T004, T005, T008, T010 |
| FR-016, FR-018 (ausencia o expirada en solo lectura → 4) | T005, T008 |
| FR-017 (escritura en solo lectura → 1) | T005, T008 |
| FR-019, FR-020, FR-023 (ruta por omisión, variable y opción; precedencia) | T003, T004 |
| FR-021 (directorio `0700` y fichero `0600`) | T004 |
| FR-022 (ruta inservible → 2, cualificada por modo) | T003, T005, T008, T010 |
| FR-024 a FR-029 (migraciones embebidas, `schema_version`, atomicidad, versión ajena, fichero inutilizable, esquema v1) | T004 |
| FR-030, FR-031 (WAL y `PRAGMA` seguros) | T004 |
| FR-032 (dos clientes, mismo proceso y dos procesos) | T006, T010 |
| FR-033, FR-035 (clase declarada y mensajes que nombran lo necesario) | T002, T009, T010 |
| FR-034 (ningún `panic`) | T009 |
| FR-036, FR-037 (R3 con su primer dueño; `duenoObligatorio` sigue en `false`) | T011 |
| FR-039 (`sqlclosecheck` y `rowserrcheck` en verde sin supresión) | T010, T012 |
| FR-040, FR-041 (tests etiquetados en `t.TempDir()`, como gate y alcanzados por el lint) | T010 |
| FR-043 (única dependencia nueva) | T004 |
| FR-044 (la superficie del binario no cambia) | T011, T013 |
| FR-045, FR-046 (adaptador de prueba; reproducción estricta) | T007, T008 |
| FR-047 (`--offline` por el kernel, con el applet mínimo) | T008 |
| SC-001 | T007, T008 |
| SC-002 | T005 |
| SC-003 | T005, T008, T013 |
| SC-004 | T003, T004, T008, T010 |
| SC-005, SC-006 | T004 |
| SC-007 | T006, T010 |
| SC-008 | T010, T012, T013 |
| SC-009 | T010, T013 |
| SC-010 | T005 |
| SC-011 | T009 (trece), T010 (dos) |
| SC-012 | T011, T013 |
| SC-013 | T001, T006, T011, T012 |
| SC-014 | T013 |

## Dependencias y orden

El orden es estrictamente secuencial: **ninguna tarea depende de una posterior**.

- **T001** (el puerto) va primero: la constitución fija `core` → adaptador, y `var _ core.Cache` (T005) no
  compilaría sin él.
- **T002 → T003 → T004 → T005** es el adaptador de dentro afuera: el error existe antes que quien lo
  produce; la ruta antes que la apertura que la usa; la apertura y las migraciones antes que las entradas
  que necesitan una base migrada. La regla «inexistente» se define en T003 porque los dos `Stat` de T004
  la usan.
- **T005** cierra el paquete: hasta que existen `Get` y `Put` no se pueden comprobar la ausencia en solo
  lectura, la expiración ni el estado «sin base»; por eso recoge las comprobaciones de T003 y T004 que
  exigen leer, declarando sus ficheros de test.
- **T006** y **T009** son tests sobre el código ya existente y no podrían ir antes.
- **T007** (`[datos]`) precede a **T008**, único consumidor de la grabación.
- **T010** necesita el paquete completo y el adaptador de prueba; entra con `.golangci.yml` y `Makefile` en
  la misma tarea porque un fichero etiquetado no debe existir sin que el lint lo alcance ni sin que la
  integración continua lo ejecute (obligación 7).
- **T011 → T012 → T013 → T014**: los controles, su demostración sobre copia desechable, la medida agregada
  y, al final, la única tarea que habla con la plataforma.

La dependencia nueva entra en **T004**, con el código que la importa: añadirla antes dejaría
`go mod tidy -diff` en rojo.

## Notas

- **Lo que este hito no crea, y no por olvido**: `internal/source/`, `internal/store/`, `internal/graph/`,
  `pkg/`, `testdata/<fuente>/` en la raíz, `schemas/`, `docs/SOURCES.md`, `scripts/verify-sources.sh`,
  `evals/`, `skills/`, ningún applet, verbo ni bandera nueva, ningún guion `testscript` y ninguna entrada en
  `CHANGELOG.md`. Están fuera de alcance (spec, *Fuera de alcance*) y crearlos vacíos sería un marcador de
  posición.
- **Los ficheros de H0/H1/H2 que se tocan son exactamente estos y ninguno más**: `internal/arch_test.go`
  (T011), `Makefile` (T010, solo la lista de prerrequisitos de `ci`; ninguna receta), `.golangci.yml`
  (T002 solo `misspell.ignore-rules` si el español lo obliga; T008 solo `misspell.ignore-rules`, la palabra
  `reproduccion`; T010 solo `run.build-tags`; T011 solo comentarios), `docs/PENDIENTES.md` (T011), más `go.mod` y `go.sum` (T004). Ninguna tarea toca
  `internal/core/schema`, `internal/cli`, `internal/app`, `internal/httpx`, `internal/render` ni `cmd/`.
- **`KITLEGAL_RECORD` no aparece en ninguna tarea** y no se graba ningún fixture contra ninguna fuente: el
  único material de reproducción se escribe a mano en T007 (constitución, «Reglas del modo desatendido»).
- **La contingencia de `misspell`** (supuesto S2) está asignada a T002, la primera tarea que crea ficheros
  de `internal/cache`, y acotada a `misspell.ignore-rules`. Si el falso positivo apareciera en otra tarea,
  esa tarea **no** edita un fichero que no declara: reescribe el término en español —no es un atajo, porque
  no toca ninguna regla ni suprime ningún hallazgo— y, si el término lo fija un contrato y no puede
  cambiarse, se anota en `gates/tarea-Tnnn.md`, se redelimita la tarea declarando `.golangci.yml` acotado a
  esas palabras y se deja sin marcar para que el intento siguiente la cierre (precedente de H2 T010). No se
  usan como palabra suelta `directorios`, `transaccion` ni `configuracion` (D15). **T008 es ese caso**
  (intento 1): `reproduccion` la fijan el nombre del subtest `sin-cache-la-reproduccion-falla` y el
  directorio de grabaciones de H2, y la línea de T008 declara ya `.golangci.yml` acotado a esa palabra.
- **Ninguna supresión en todo el hito** (SC-008): `gosec` G204 se resuelve pasando el ejecutable como
  parámetro y un literal como único argumento; G304 pasando toda ruta por `filepath.Clean`; G301, G302 y
  G306 con `0o700`, `0o600` y `0o600`. El único `t.Skip` es la precondición de entorno de los tests de
  permisos (T010), que nunca se produce donde la aserción tendría sentido y que **no puede producirse en
  integración continua**: allí la misma precondición incumplida es `t.Fatalf` (variable `CI` no vacía), de
  modo que el gate lo hace valer por sí solo y T014 no tiene que leerlo en ningún registro.
- **Los tests de permisos restauran los permisos en `t.Cleanup` registrado después de `t.TempDir()`**
  (obligación 6): la limpieza de `TempDir` es un `RemoveAll` y fallaría el test si no pudiera borrar.
- **El criterio de aceptación se mide dos veces** en T008: por el resultado (idéntico byte a byte) y por la
  imposibilidad de pedir (reproducción estricta sobre un directorio vacío, que nombra la petición si
  alguien la emite).

## Comprobación contra la rúbrica del juez (`juez_tasks`, criterios a-g)

| Criterio | Dónde se cumple |
|---|---|
| a. analisis_critico | `gates/analyze.md` ya está escrito: 0 CRITICAL, 0 HIGH, cobertura 100 % (61/61) y dos hallazgos MEDIUM de recuento en prosa —F1 sobre el número de excepciones y la fila `c` de esta tabla, F2 sobre la frase introductoria de la tabla del contrato de errores—, **ambos corregidos** en este fichero y en `contracts/errores-y-codigos.md`, y con ellos el residuo de la misma clase que quedaba fuera de los ficheros que el análisis citó (la frase de recuento de D9 en `research.md`, ahora «nueve filas ★ y seis restantes» como el contrato); las obligaciones que el plan trasladó a `tasks.md` (1 a 14) están todas asignadas a una tarea concreta y nombradas en su línea |
| b. trazabilidad | Tabla «Trazabilidad: requisito → tarea» con los 47 requisitos funcionales y los 14 criterios de éxito; ninguna tarea añade nada que el spec no pida |
| c. rebanadas_verdes | Cada tarea lleva su test y la implementación mínima que lo hace pasar; las **ocho** excepciones (T001, T006, T007, T009, T011, T012, T013, T014) están declaradas con su razón y ninguna deja un test en rojo; ninguna tarea de solo test mide un comportamiento que otra no haya implementado ya (las filas 12 y 13 del contrato de errores, que T010 mide, las fija T004); el orden es secuencial y ninguna tarea depende de una posterior (§Dependencias). **Toda comprobación que una tarea promete es observable con las órdenes que esa tarea ejecuta**: T013 espera de `make ci` la línea `ci:` exacta que T010 escribe, y T014 comprueba S3 por el color del trabajo `ci` (`gh pr checks`, `gh run list`) y no por un `PASS`/`SKIP` que la receta de `test-integration`, contrato de H0 y sin `-v`, no imprime; lo que el registro no puede mostrar lo hace valer el propio gate, con la precondición de permisos de T010 en `t.Fatalf` bajo `CI` |
| d. rutas_declaradas | Cada línea nombra todos los ficheros que crea o modifica, sin rutas genéricas; declarar `x.go` habilita `x_test.go`, pero **no al revés**: declarar `x_test.go` no habilita `x.go`. El principio vale igual para los ficheros de test y para los de producto: **toda tarea que amplíe o cambie el cuerpo de un fichero creado por una tarea anterior lo declara** —de test (T004 y T005 declaran `cliente_test.go`; T005 declara `abrir_test.go` y `ruta_test.go`; T006 declara `cliente_test.go`; T009 declara `errores_test.go`) y de producto (T004 declara `cliente.go`, que crea T003, porque allí `New` pasa a abrir y migrar, `Close` a cerrar la conexión y `Cliente` gana `db` y `versionEsquema`)—. Llamar a una función o leer un campo de un fichero anterior **no** obliga a declararlo: por eso T004 no declara `ruta.go` (solo usa `esInexistente`) ni `errores.go` (solo usa constructores que nacen en T002), y T005 no declara `cliente.go` (`Get`, `Put` y `var _ core.Cache` van en `entradas.go` y solo leen `cerrado`, `db`, `versionEsquema` y `soloLectura`). **Y ninguna línea nombra en forma de ruta algo que la tarea no vaya a tocar**: las advertencias sobre lo que el hito no abre van en prosa —el directorio reservado a los adaptadores de fuente, los tres paquetes de almacenamiento, el árbol de fixtures y el de esquemas de la raíz, el paquete `main` del binario, el paquete de la caché visto desde fuera, la configuración del lint y el roadmap—, comprobado reproduciendo la extracción del guardián sobre cada línea `- [ ] Tnnn` |
| e. datos_separados | T007 es la única `[datos]`, toca solo `internal/cache/testdata/` y no mezcla otro trabajo; T014 es la única `[plataforma]`, va la última y solo publica y lee estados, escribiendo únicamente en el directorio del feature (el cuerpo de la propuesta, `gates/pr-h3.md`, y su evidencia) |
| f. dod | Tabla «Definition of Done» con los once puntos: los **cinco** aplicables (1, 2, 3, 5, 9) con su tarea y los **seis** no aplicables (4, 6, 7, 8, 10, 11) con la razón del spec (sin applet, sin comportamiento visible, sin ADR nuevo, sin fuente externa, sin skill, sin territorio) |
| g. checklist_veraz | `checklists/requirements.md` tiene sus dieciséis ítems marcados y ninguna tarea de este fichero los modifica |
