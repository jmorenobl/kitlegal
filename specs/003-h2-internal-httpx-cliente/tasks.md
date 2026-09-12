# Tasks: H2 · `internal/httpx`: cliente HTTP responsable + grabación de fixtures

**Input**: `specs/003-h2-internal-httpx-cliente/` (spec.md, plan.md, research.md, data-model.md,
contracts/, quickstart.md)

**Prerrequisitos**: spec.md y plan.md leídos; `.specify/memory/constitution.md` aplicada; H0 y H1 cerrados
(módulo, binario, `Makefile`, `.golangci.yml` con `noctx`, `bodyclose`, `depguard` y `forbidigo`, flujos de
integración continua, umbrales de cobertura, kernel de la línea de órdenes, sobre de salida, códigos de
salida y `internal/arch_test.go` ya existen).

**Modo**: desatendido. Las tareas se ejecutan y verifican **una a una** por el workflow `hito`, con
guardián de diff por tarea. Por eso cada tarea es una **rebanada vertical** que incluye su comprobación y
deja `make ci` en verde al terminar, y **declara en su propia línea todas las rutas** que va a crear o
modificar.

## Formato: `[ID] [P?] [Story] Descripción con rutas`

- **[P]**: la tarea no comparte fichero con sus vecinas (podrían repartirse si alguien las hiciera a mano;
  en el bucle del workflow van igual en secuencia).
- **[Story]**: historia de usuario del spec — US1 consultar sin poder faltar al respeto · US2 saber por qué
  falló sin leer el mensaje · US3 la suite entera sin red · US4 grabar lo que la fuente respondió · US5 no
  poder saltarse las garantías · US6 ver qué se pediría sin pedirlo.
- **[datos]**: la tarea toca material bajo un directorio `testdata`. Solo hay **una**: T012, que no mezcla
  ningún otro trabajo. Es material de test escrito a mano contra un host ficticio, **no** una grabación de
  ninguna fuente real (plan.md §«Fixtures», FR-044).
- **[plataforma]**: la tarea necesita la plataforma remota. Solo hay **una**: T019, la última, único punto
  del hito autorizado a empujar la rama. Fusionar nunca es del workflow.

## Batería de verificación por tarea

- **`make ci`** desde la primera tarea: formato · lint (`noctx`, `bodyclose`, `depguard`, `forbidigo`,
  `gosec`, `misspell`) · tests con detector de carreras · vulnerabilidades · comprobación de esquemas ·
  secretos · integridad de módulos · `go mod tidy -diff`. Existe desde H0: ninguna tarea de H2 lo crea.
- **Cero red en todo el hito**: cada test levanta su propio servidor de bucle local (`httptest`) o lee un
  directorio de grabaciones. Ningún test conoce una dirección externa.
- **`KITLEGAL_RECORD` solo dentro de `t.Setenv`**, en un test, contra un servidor local y un `t.TempDir()`
  (FR-044). Ninguna tarea lo exporta en el entorno ni graba contra una fuente real.
- **Solo hosts admitidos** en tests y fixtures: `127.0.0.1`, `localhost`, `fuente.prueba` y `otra.prueba`
  (plan.md, control 18 y obligación 11). La única dirección externa que puede aparecer es la de la
  identificación, `https://ventanillalegal.es/bot`, que se envía y no se consulta.
- **Sin atajos** al reparar un rojo: ni `//nolint`, ni linter desactivado, ni test saltado, ni error
  silenciado, ni umbral rebajado (SC-010). Si un test queda legítimamente en rojo porque su implementación
  pertenece a una tarea posterior, la tarea está mal delimitada: se anota, no se parchea.

**Rebanadas verticales, y las cinco excepciones declaradas.** T001 a T011, T013 y T015 son rebanadas
completas: cada una escribe su test de tabla antes que el código que lo hace pasar y deja `make ci` en
verde por sí sola. Las otras cinco no añaden producto nuevo, por una razón distinta cada una, y también lo
dejan en verde: **T012** solo puede tocar material bajo un directorio `testdata` (regla `[datos]`, que
prohíbe mezclarlo con otro trabajo) y añade ficheros que ningún test lee todavía; **T014** y **T016** son
tests y controles **sobre código ya existente** de las tareas anteriores, que es el único caso en que test e
implementación pueden ir separados; **T017** y **T018** son validaciones que se ejecutan sobre una copia
desechable o sobre el árbol ya terminado y cuyo único fichero escrito es la evidencia; **T019** solo
publica. Ninguna deja nada a medias para una tarea posterior.

---

## Phase 1: Fundación — las dos ampliaciones del kernel de H1

**Objetivo**: que el dominio sepa transportar una clase de error y una descripción de ensayo antes de que
exista nada que las produzca. Sin esto, el cliente no podría declarar su clase sin importar `internal/cli`
ni hacer visible la descripción de `--dry-run`.

- [X] T001 [US2] Clase de error declarada por el propio error y reconocida por el kernel: `internal/core/schema/error.go` gana la interfaz `ConClase` (`error` embebido más `Clase() Clase`), sin importar nada nuevo y sin tocar `Clase`, sus seis constantes ni `DatosError`; `internal/cli/errors.go` amplía `Clasificar` con una comprobación `errors.As` sobre `schema.ConClase` **después** de los cinco sentinelas de H1 y antes de la salida «inesperado», dejando intacta la tabla `codigoDeClase` y su `switch` sin rama `default`, de modo que una clase declarada fuera del vocabulario caiga en «inesperado» (código 1) y que envolver con `%w` siga sin cambiar la clase; y los casos nuevos en `internal/core/schema/error_test.go` y en `TestClasificar` de `internal/cli/errors_test.go` con un tipo local de test que implementa `ConClase` para las seis clases, para una clase desconocida y para el error envuelto, manteniendo `internal/cli` ≥ 90 % y `internal/core` ≥ 85 % (FR-031, FR-063, D4).

- [ ] T002 [P] [US6] La descripción del ensayo viaja por el dominio y la presenta el kernel: `internal/core/schema/sobre.go` añade a `Resultado` el campo `Ensayo []string` —sin I/O, sin importaciones nuevas y sin alterar el `Sobre` ni sus seis claves—; `internal/app/main.go` presenta esas líneas por el presentador hacia la salida de error, **siempre visibles** y sin pasar por el registro de eventos, conservando el código de salida 0 que H1 fijó para `--dry-run`; `internal/app/main_test.go` añade `TestDryRunPresentaElEnsayo` con un applet de test que devuelve un `Resultado` con `Ensayo`, comprobando que la línea aparece en la salida de error, que la salida estándar no la lleva y que el código es 0; y `docs/ADR/0011-ensayo-por-capas.md` registra en formato MADR corto la decisión y sus alternativas rechazadas (`slog` filtrable, un `io.Writer` en el cliente, cambiar la firma de `Ejecutar`), porque cambia el contrato del applet de ADR 0005 (FR-051, FR-065, punto 7 de la Definition of Done, D6).

**Checkpoint**: un error puede declarar su clase sin importar el kernel y una descripción de ensayo puede
llegar a la salida de error. Nada de esto es todavía observable desde el binario, cuya superficie no cambia.

---

## Phase 2: Los tipos del paquete, sin red

**Objetivo**: fijar la forma de lo que el cliente pide, devuelve y falla, con tests puros, antes de abrir
ninguna conexión. `internal/httpx` nace aquí.

- [ ] T003 [US1] El paquete y sus tipos de datos: `internal/httpx/doc.go` con el comentario de paquete que declara qué garantiza y por qué no se puede usar mal (única excepción declarada a «un fichero de test por fichero de código»: no tiene declaraciones); `internal/httpx/peticion.go` con `Peticion{Metodo, URL}`, `Respuesta{Peticion, URL, Estado, Cabeceras, Cuerpo, Ensayo}` —cuerpo **ya leído** en `[]byte`, nunca un `*http.Response`—, `Cabeceras map[string][]string` con `Get` que canonicaliza el nombre y devuelve `""` si no está, y `Respuesta.Descripcion()` que produce `"<Metodo> <URL>"`; `internal/httpx/sitio.go` con la clave de sitio (esquema + host en minúsculas + puerto, con 80/443 por omisión del esquema, sin ruta ni consulta) y el mapa de sitios con exclusión mutua para el uso concurrente; y las tablas con `t.Parallel()` de `internal/httpx/peticion_test.go` (`TestCabecerasGet`, `TestRespuestaDescripcion`) y `internal/httpx/sitio_test.go` (`TestClaveDeSitio` con las ocho filas del plan, escritas **solo** contra `fuente.prueba`, `otra.prueba` y dos puertos de `127.0.0.1`); y comprobando al primer `make lint` el supuesto S2 (`misspell` sobre el español nuevo), para lo que declara también `.golangci.yml`, acotado a **una palabra nueva bajo `misspell.ignore-rules`** si el español dispara un falso positivo —ninguna regla, ningún linter, ninguna exclusión y nunca una supresión—, y sin tocarlo si no hace falta (FR-002, FR-013, FR-019, D2, D7, D14, obligaciones 3 y 11 del plan).

- [ ] T004 [P] [US2] El único error que el paquete produce: `internal/httpx/errores.go` con `Error{Peticion, Estado, Espera, EsperaConocida, Causa}` más su clase privada, los constructores internos por clase, `Error() string` con mensaje en español que nombra sitio y ruta cuando hay petición implicada y la opción que falta o sobra cuando no la hay, `Unwrap()` que expone `Causa` y `Clase() schema.Clase` que lo hace `schema.ConClase`, sin `panic` en ninguna ruta; y `internal/httpx/errores_test.go` con `TestErrorEnvueltoConservaLaClase` (envolver con `fmt.Errorf("%w")` no cambia lo que `cli.Clasificar` devuelve), `TestErrorRetryAfter` (`Espera` y `EsperaConocida` a partir de la cabecera, y su ausencia sin cambiar la clase) y `TestErrorMensajesEnEspanol` (FR-030, FR-031, FR-033, FR-034, FR-063, D4, D20).

- [ ] T005 [P] [US1] La identificación y su versión, coherentes con el binario: `internal/httpx/agente.go` con la variable de paquete `version` (por omisión `dev`, inyectable por `-ldflags`) y `AgenteDeUsuario()` que devuelve exactamente `kitlegal/<versión> (+https://ventanillalegal.es/bot)`, sin literal de versión escrito a mano y sin forma de vaciarla desde fuera; `internal/httpx/agente_test.go` con `TestAgenteDeUsuario`, que construye la expresión regular **sobre la variable** `version` (no sobre el literal `dev`) y pasa el sufijo por `regexp.QuoteMeta`, exige `version` no vacía y registra la identificación con `t.Log` para que la inyección sea observable; `Makefile` añade a `LDFLAGS` la `-X <módulo>/internal/httpx.version=$(VERSION)` junto a las tres que ya existen, sin tocar ningún objetivo; y `cmd/kitlegal/main_test.go` añade `TestVersionDeLaIdentificacion`, que importa `internal/httpx` **solo en test** y comprueba que el valor por omisión de `main.version` y el de la identificación coinciden, de modo que no puedan divergir en silencio (FR-006, FR-007, FR-008, D5).

**Checkpoint**: existen la petición, la respuesta, el sitio, el error y la identificación; ninguno importa
todavía nada de red y todos tienen su tabla de casos.

---

## Phase 3: El cliente contra la red, de dentro afuera por la cadena

**Objetivo**: `Pedir` como única operación, y cada decorador añadiendo su garantía sin que se pueda
desactivar. Cada tarea deja la cadena completa hasta donde llega y su control literal del hito en verde.

- [ ] T006 [US1] La única operación de red, con identificación, plazo, ensayo y redirecciones: `internal/httpx/transporte.go` con el clon de `http.DefaultTransport` (verificación de certificados no desactivable, `Client.Timeout = 0` porque el plazo lo pone el contexto, `CheckRedirect` que impide las redirecciones automáticas de la biblioteca y sin almacén de cookies); `internal/httpx/identificar.go` con el decorador que pone la cabecera de identificación en **toda** petición que baja por la cadena y con el **único constructor de peticiones del paquete**, `nuevaPeticionIdentificada(ctx, metodo, url)`, que llama a `http.NewRequestWithContext` y fija la cabecera con `AgenteDeUsuario()` antes de devolver la petición, de modo que también nazcan identificadas las que el propio paquete origina por debajo del decorador —el bucle de redirecciones de `cliente.go` y, en T009, la de `robots.txt`— (FR-009, D3); `internal/httpx/cliente.go` con `Cliente` (campos privados), `New(opciones ...Opcion) (*Cliente, error)` que valida en la construcción, `Opcion` y las opciones `ConFuente` y `ConRegistrador` con sus reglas de validez, y `Pedir(ctx context.Context, ejecucion schema.Contexto, p Peticion) (Respuesta, error)` como **única** superficie de red —método restringido a GET/HEAD y dirección absoluta `http`/`https` comprobados antes de abrir nada (clase 2), salida temprana de ensayo, bucle propio de redirecciones con tope 10 y detección de bucle (clase 4), lectura íntegra del cuerpo y cierre del `*http.Response` dentro del paquete, y clasificación del resultado (5xx → clase 4, 429 con `Retry-After` → clase 5, todo lo demás entregado como `Respuesta`)—, con la cadena de un solo escalón `identificar → transporte` y registro de eventos únicamente por el `*slog.Logger` recibido, a nivel `debug` y nunca a la salida estándar; y los tests `internal/httpx/transporte_test.go` (`TestTransporteConfiguracion`), `internal/httpx/identificar_test.go` (`TestIdentificacionEnTodaPeticion` con las subpruebas que la cadena de esta tarea ya permite —`recurso` y `saltos`—, donde el servidor cuenta en cero las peticiones sin la cabecera exacta y que T008 y T009 amplían con las suyas; y `TestSoloIdentificarConstruyePeticiones`, que recorre los ficheros del paquete con `go/parser` y falla si alguno que no sea `identificar.go` llama a `http.NewRequestWithContext` o construye un `http.Request`), `internal/httpx/cliente_test.go` (`TestNewSinOpciones`, `TestOpcionesInvalidas`, `TestPedirRechazaMetodo`, `TestPedirRechazaDireccion`, `TestPedirRespetaElPlazo`, `TestPedirSigueRedirecciones`, `TestPedirCortaCadenasDeRedirecciones`) e `internal/httpx/ensayo_test.go` (`TestEnsayoNoAbreConexion`, `TestEnsayoDevuelveRespuestaDeclarada`, `TestEnsayoCompruebaArgumentos`: cero accesos al servidor, `Respuesta{Ensayo: true}` sin estado ni cuerpo y **sin error**, y método o configuración inválidos comprobados igualmente) (FR-001 a FR-006, FR-009 a FR-012, FR-029, FR-030, FR-032, FR-035, FR-050, FR-052, FR-065, SC-001, SC-002, SC-013, US6 escenarios 1-4, D1, D2, D3, D10, D11).

- [ ] T007 [US1] Ritmo por sitio, y solo por sitio: `internal/httpx/sitio.go` añade a la struct `sitio` el campo `limitador *rate.Limiter`, creado al crear el sitio (data-model §4: un solo mapa y una sola exclusión mutua por clave de sitio, nunca un mapa por decorador); `internal/httpx/ritmo.go` con el decorador que usa ese limitador por clave de sitio —un token cada intervalo, ráfaga 1, sin paralelismo propio— y espera turno con `Wait(ctx)`, de modo que el vencimiento del contexto mientras se espera impida emitir la petición y produzca clase 4; `internal/httpx/cliente.go` añade la opción `ConIntervalo(d time.Duration)` (1 s por omisión, `d <= 0` es clase 2, un solo valor para todos los sitios de un cliente) e inserta el decorador en la cadena por debajo de la identificación; `internal/httpx/ritmo_test.go` con `TestRitmoSeparaPeticionesDelMismoSitio`, `TestRitmoNoRetrasaOtroSitio` (dos servidores locales que solo difieren en el puerto y por tanto son dos sitios) y `TestRitmoRespetaElContexto`; e incorpora `golang.org/x/time` v0.16.0 al módulo con `go get`, comprobando el supuesto S1 —`go mod tidy -diff` limpio, `go mod graph` sin módulos nuevos más allá del añadido y `TestDependenciasDelBinario` en verde **sin tocar su lista**, porque ningún paquete de producción importa `internal/httpx` (FR-019 a FR-022, FR-058, SC-004, D8, obligación 1 del plan).

- [ ] T008 [US1] Reintentos con retardo creciente y aleatorio, interrumpibles: `internal/httpx/reintentos.go` con el decorador que reintenta **solo** los 5xx y los fallos de transporte —nunca un 4xx, 429 incluido—, hasta el número de intentos configurado, con espera base 500 ms, factor 2, techo 30 s y *equal jitter* cuya parte aleatoria sale de `crypto/rand.Read` sobre ocho bytes reducidos con `encoding/binary` (ningún fichero del paquete importa `math/rand`, `math/rand/v2` ni `math/big`, porque G404 marca los dos primeros; el resultado de `Read` no se comprueba porque la función no devuelve nunca error, y la conversión a `int64` lleva el `>> 1` que G115 reconoce, todo ello sin una sola supresión), con el reloj inyectable desde el propio paquete para que los tests no esperen de verdad, cierre y descarte del cuerpo de cada respuesta desechada y ningún intento posterior a la cancelación del contexto; `internal/httpx/cliente.go` añade la opción `ConIntentos(n int)` (3 por omisión, `n < 1` es clase 2) y coloca el decorador por encima del ritmo, de modo que cada reintento espere su turno; `internal/httpx/reintentos_test.go` con `TestReintentosDosErroresYUnAcierto` (tres peticiones contadas en el servidor —las tres con la cabecera de identificación exacta, contadas en cero sin ella—, esperas crecientes y distintas entre dos ejecuciones), `TestReintentosNoRepite4xx`, `TestReintentosCancelacionGana`, `TestReintentosCierraCuerposDescartados` y `TestReintentosAgotados`; y `internal/httpx/identificar_test.go` gana la subprueba `reintentos` de `TestIdentificacionEnTodaPeticion`, ahora que la cadena tiene ese escalón, sin tocar `internal/httpx/identificar.go` (FR-009, FR-023 a FR-029, SC-001, SC-005, D9, obligaciones 2 y 12 del plan).

- [ ] T009 [US1] `robots.txt` antes de la primera petición a cada sitio, con su lista cerrada de casos: `internal/httpx/sitio.go` añade a la struct `sitio` el campo `reglas *reglasDelSitio` (`nil` hasta la primera evaluación, después inmutable), bajo la misma exclusión mutua que ya protege al sitio (data-model §4); `internal/httpx/robots.go` con el decorador que obtiene y evalúa el `robots.txt` del sitio para la ruta pedida y el agente `kitlegal`, **construyendo la petición de `robots.txt` con `nuevaPeticionIdentificada` de `identificar.go`** —la fabrica él, por debajo del decorador `identificar`, así que sin ese constructor saldría sin cabecera y rompería FR-009— antes de entregarla a su propio siguiente, honrando **solo** las reglas de permiso y denegación de rutas (ninguna otra directiva, `Crawl-delay` incluida), interpretándolo con `temoto/robotstxt`, cacheando el resultado por sitio en memoria y sin caducidad —una sola obtención por sitio y cliente, nunca en disco—, sin evaluarse a sí mismo ni a ningún salto de su propia cadena de redirecciones, y aplicando la lista cerrada de FR-015: 4xx distinta de 429 o cuerpo vacío permiten; 5xx o fallo de transporte agotados los intentos, 2xx ilegible y estado no previsto deniegan con clase 5; 429 deniega con clase 5 sin reintento y con `Retry-After`; la cadena de redirecciones propia se sigue con el mismo tope 10 y su exceso deniega con clase 5, no con la clase 4 del recurso; el vencimiento del contexto termina la operación con clase 4 sin evaluar ninguna regla; `internal/httpx/cliente.go` inserta el decorador en la cadena por encima de los reintentos, de modo que cada salto del recurso se someta al `robots.txt` y al ritmo de su sitio de destino y ningún destino herede el permiso del origen; `internal/httpx/robots_test.go` con `TestRobotsDeniegaLaRuta` (cero accesos a la ruta en el servidor), `TestRobotsSePideUnaVezPorSitio` (diez rutas, una sola obtención, y esa obtención llega al servidor con la cabecera de identificación exacta), `TestRobotsCasosDeObtencion` (la tabla completa de FR-015), `TestRobotsRedirigido` (con la cabecera comprobada en cada salto de la cadena del `robots.txt`), `TestRobotsNoSeEvaluaASiMismo` y `TestRobotsPorSitio`; `internal/httpx/identificar_test.go` gana la subprueba `robots.txt` de `TestIdentificacionEnTodaPeticion` —cerrando las cuatro procedencias que el control 6 del plan y el escenario 2 del quickstart dan por cubiertas— sin tocar `internal/httpx/identificar.go`; `internal/httpx/cliente_test.go` añade `TestPedirDesdeVariasGoroutines`, que usa el mismo cliente desde varias goroutines contra el mismo sitio bajo el detector de carreras y comprueba que el `robots.txt` se pidió una sola vez; e incorpora `github.com/temoto/robotstxt` v1.1.2 al módulo con `go get`, comprobando de nuevo el supuesto S1 (FR-009, FR-013 a FR-018, FR-021, FR-030, FR-057, FR-058, SC-001, SC-003, SC-011, US1 escenarios 4 y 8, US2 escenario 4, D3, D7, obligación 12 del plan).

**Checkpoint**: el cliente contra la red está completo y ninguna de sus garantías se puede desactivar. Los
cinco controles literales del hito —reintentos, 429, `robots.txt` que deniega, identificación presente y
`--dry-run`— tienen ya su test unitario sin red (SC-009). La identificación queda cubierta **en las cuatro
procedencias** porque T006 abrió `TestIdentificacionEnTodaPeticion` con `recurso` y `saltos`, T008 le
añadió `reintentos` y T009 `robots.txt`; y `TestSoloIdentificarConstruyePeticiones` (T006) impide que
aparezca una quinta procedencia sin cabecera.

---

## Phase 4: Grabación y reproducción — la suite sin red

**Objetivo**: las dos mitades del mecanismo del principio III. El formato existe para ser reproducido, así
que el nombre y la escritura van antes que la lectura, y los fixtures antes que los tests que los leen.

- [ ] T010 [P] [US4] El nombre de fichero derivado, determinista y portable: `internal/httpx/nombre.go` con la derivación del nombre a partir del método y de la dirección completa —esquema, sitio, ruta y parámetros de consulta—, saneada a un nombre portable a Windows (sin `<>:"/\|?*`), legible y estable, que nadie fuera del paquete declara porque buena parte de las peticiones las origina el propio cliente; y `internal/httpx/nombre_test.go` con `TestNombreDeGrabacion`, la tabla del contrato de grabación §2 más los diez pares petición → fichero de plan.md §«Fixtures», incluido el par que colisiona (`/a,b` y `/a_b` producen el mismo nombre), escrita **solo** contra `fuente.prueba` y `otra.prueba` (contrato de grabación §2, FR-039, D12, obligación 11 del plan).

- [ ] T011 [US4] Grabar lo que la fuente respondió, solo con la variable y solo donde se declara: `internal/httpx/grabar.go` con el decorador que, con `KITLEGAL_RECORD=1`, escribe `<raíz>/<fuente>/<nombre>.json` con el objeto del contrato §1 —claves `formato`, `grabado_en`, `peticion` y `respuesta` en ese orden, indentado con dos espacios, sin escape de HTML, cabeceras como `nombre canónico → lista de valores` con las claves ordenadas, cuerpo como texto si es UTF-8 válido y `cuerpo_base64` si no—, mediante fichero temporal en el mismo directorio y `rename`, sin alterar ni consumir lo que recibe quien llamó, detectando la colisión por contenido (si el fichero existente guarda otra petición, clase 1 nombrando las dos; la misma petición sí sustituye) y distinguiendo el fallo de E/S sobrevenido (clase 1) del valor de opción inválido (clase 2); `internal/httpx/cliente.go` añade la opción `ConRaizDeGrabacion(dir string)`, lee `KITLEGAL_RECORD` una sola vez en `New` con la constante exportada `VariableGrabacion`, y aplica las reglas de construcción del contrato §3 —grabación activa sin `ConFuente` o sin `ConRaizDeGrabacion` es clase 2 nombrando la opción que falta, raíz inexistente, no directorio o no creable es clase 2 nombrando la ruta, otro valor de la variable es clase 2 nombrando el valor, y el cliente **no** deduce nunca la raíz del directorio de trabajo ni de la raíz del módulo—; y `internal/httpx/grabar_test.go` con `TestGrabarEscribeElFichero`, `TestGrabarNoAlteraLaRespuesta`, `TestGrabarFormatoEstable` (dos grabaciones idénticas byte a byte tras neutralizar solo `grabado_en` y las cabeceras que el servidor regenera, `Date` entre ellas), `TestGrabarSinVariableNoEscribe`, `TestGrabarConfiguracionIncompleta`, `TestGrabarRaizInvalida`, `TestGrabarColision`, `TestGrabarValorDeVariableInvalido` y `TestGrabarCuerpoBinario`, todos con `t.Setenv` y `t.TempDir()` contra un servidor local —nunca contra una fuente real (FR-044)— y midiendo al cerrar la tarea la garantía SC-008 de forma **diferencial**: la salida de `git status --porcelain` restringida a los directorios de fixtures del repositorio es la misma antes y después de `make ci` (FR-036 a FR-043, FR-064, SC-008, US4, D12, obligación 7 del plan).

- [ ] T012 [datos] [US3] Las diez grabaciones escritas a mano del directorio de reproducción —material de test contra el host ficticio `http://fuente.prueba`, **no** una grabación de ninguna fuente real y nunca obtenida de la red (FR-044)—, cada una con el nombre exacto que `TestNombreDeGrabacion` ya fija, `"formato": 1`, `grabado_en` fijo `2026-09-12T00:00:00Z`, `peticion.cabeceras` con solo la identificación `kitlegal/dev (+https://ventanillalegal.es/bot)` y una cabecera `Date` fija en la respuesta: `internal/httpx/testdata/reproduccion/prueba/GET_http_fuente.prueba_norma_q_id_BOE-A-2015-10565.json` (200, XML de una línea con `<`, `>` y `&` sin escapar), `internal/httpx/testdata/reproduccion/prueba/HEAD_http_fuente.prueba_norma_q_id_BOE-A-2015-10565.json` (200, mismas cabeceras, `"cuerpo": ""`), `internal/httpx/testdata/reproduccion/prueba/GET_http_fuente.prueba_documento.pdf.json` (200, `cuerpo_base64` de bytes que no son UTF-8 válido), `internal/httpx/testdata/reproduccion/prueba/GET_http_fuente.prueba_antigua.json` (302 con `Location` relativa al primero), `internal/httpx/testdata/reproduccion/prueba/GET_http_fuente.prueba_movida.json` (302 hacia una dirección sin grabación), `internal/httpx/testdata/reproduccion/prueba/GET_http_fuente.prueba_bucle.json` (302 hacia sí misma), `internal/httpx/testdata/reproduccion/prueba/GET_http_fuente.prueba_caida.json` (503), `internal/httpx/testdata/reproduccion/prueba/GET_http_fuente.prueba_limitada.json` (429 con `Retry-After: 120`), `internal/httpx/testdata/reproduccion/prueba/GET_http_fuente.prueba_a_b.json` (200, la que colisiona con `/a_b`) e `internal/httpx/testdata/reproduccion/prueba/GET_http_fuente.prueba_robots.txt.json` (200 con `User-agent: *` y `Disallow: /`, que debe quedar **sin usar**); ninguna otra ruta y ningún otro fichero (punto 2 de la Definition of Done, plan.md §«Fixtures», obligaciones 4 y 11 del plan).

- [ ] T013 [US3] Reproducir sin abrir una sola conexión y sin depender del reloj: `internal/httpx/reproducir.go` con el transporte de reproducción que localiza la grabación por el nombre derivado, comprueba que el método y la dirección guardados coinciden exactamente con los de la petición entrante —las cabeceras **no** participan en el emparejamiento—, rechaza un `formato` distinto de 1, y falla de forma ruidosa con clase 1 nombrando la petición ausente y, en la colisión, también la que encontró en su lugar y el fichero, sin devolver nunca una respuesta vacía ni caer a la red; `internal/httpx/cliente.go` añade `Replay(dir string, opciones ...Opcion) (*Cliente, error)`, que valida el directorio (vacío, inexistente o no directorio es clase 2 nombrando la ruta), rechaza con clase 2 la grabación y la reproducción a la vez y las opciones sin sentido en reproducción (`ConRaizDeGrabacion`, `ConIntervalo`, `ConIntentos`), y monta la cadena reducida a la **lista cerrada** de garantías vigentes —contexto obligatorio, identificación, método GET/HEAD, redirecciones grabadas con su tope resueltas dentro del directorio y clasificación por el estado grabado—, **sin** `robots.txt`, **sin** ritmo y **sin** reintentos; y `internal/httpx/reproducir_test.go` con `TestReplaySirveLasGrabaciones` (texto, HEAD y binario), `TestReplayEsDeterminista` (mismo resultado con `-count=2 -shuffle=on`), `TestReplayPeticionSinGrabacion`, `TestReplayGrabacionDeOtraPeticion`, `TestReplayGarantiasVigentes`, `TestReplaySinRobots` (subprueba `con-grabacion`, donde la grabación del `robots.txt` que lo deniega todo queda sin usar y no es un error, y subprueba `sin-grabacion`, que copia el directorio a un `t.TempDir()` sin ese fichero), `TestReplayRedirecciones` (seguida, destino ausente y bucle), `TestReplayEstadoDeErrorGrabado` (503 → clase 4 con una sola búsqueda y sin espera; 429 → clase 5 con `Espera` de 120 s) y `TestReplayRechazaOpciones`, todos sobre las diez grabaciones de T012 (FR-043, FR-045 a FR-049, SC-007, US3, D13).

**Checkpoint**: la suite entera pasa en una máquina sin salida a Internet y da el mismo resultado dos veces
seguidas; el modo de grabación existe y no se ha ejecutado ni una vez contra una fuente real.

---

## Phase 5: El criterio de aceptación del hito y los controles

**Objetivo**: convertir en control mecánico lo que hasta aquí es disciplina, y demostrar la frase del
roadmap: «un adaptador de prueba no puede hacer una petición sin contexto ni sin UA».

- [ ] T014 [US2] Ninguna ruta de fallo sin clase, comprobada sobre el error real: `internal/httpx/errores_test.go` añade `TestClasesDeError` con las **dieciséis** subpruebas de la tabla cerrada del contrato de errores §3 —una por fila, provocando cada situación con el servidor local, un directorio temporal o un cliente de reproducción y comprobando la clase que `cli.Clasificar` devuelve y el código que `cli.CodigoSalida` produce sobre el error tal como sale del paquete y también envuelto—, cubriendo las nueve situaciones de red (5xx del recurso agotado → 4, contexto vencido en cualquier punto → 4, cadena del recurso en bucle o excedida → 4, 429 del recurso → 5, `robots.txt` que deniega → 5, 5xx o transporte de `robots.txt` agotado → 5, 429 de `robots.txt` → 5, cadena de `robots.txt` excedida → 5, método no permitido → 2) y las siete restantes (dirección inválida → 2, configuración de grabación incompleta o valor inválido → 2, raíz o directorio inválidos → 2, grabación y reproducción a la vez u opción sin sentido → 2, E/S sobrevenida al grabar → 1, colisión de nombres → 1, grabación ausente o ajena → 1), y comprobando además que el cliente **no** produce nunca las clases «no encontrado» (3) ni «requiere identidad humana» (6) y que ninguna de las dieciséis termina en `panic`; el test cubre código ya existente de T004 a T013 y no añade producto (FR-029 a FR-033, FR-063, SC-006, SC-016, punto 5 de la Definition of Done).

- [ ] T015 [US5] El adaptador de prueba, ejecutado con el kernel en proceso: `internal/httpx/adaptador_test.go`, en `package httpx_test` y por tanto **material de test: no se crea como adaptador de fuente en el árbol de producción —el hito no abre ese directorio, ver Notas— ni se registra como applet en ningún binario**, con el tipo `adaptadorDePrueba` (applet `prueba`) y `argumentosConsultar` (verbo `consultar`) que copian el patrón del contrato del cliente §6 —recibe el contexto de quien le llama, lo propaga a `Pedir` junto con el contexto de ejecución, no puede compilar de otra forma porque `Pedir` es la única operación y exige ambos—, y `TestAdaptadorDePruebaConElKernel` con las subpruebas `ensayo` (invocando `app.Main` en proceso sobre un registro de pruebas y con `--dry-run`: cero accesos al servidor local, ni al recurso ni a `robots.txt`, la descripción `GET http://…` en la salida de error, salida estándar sin cuerpo inventado y **código de salida 0**), `consulta` (la respuesta llega íntegra y sale identificada) y `plazo` (el vencimiento del contexto corta la operación y el kernel la traduce al código 4) (FR-059 a FR-062, SC-013, US5 escenario 1, US6, D16).

- [ ] T016 [US5] Las reglas de arquitectura pasan a tener dueño y la superficie queda cerrada: `internal/arch_test.go` endurece la subprueba R2 para que **exija** que el grafo transitivo real contenga `internal/httpx` —no puede volver a pasar en vacío— y siga comprobando que ningún otro paquete importa `net/http`, y añade `TestElBinarioNoEnlazaHTTPX`, que comprueba con `go list -deps` que el binario distribuido no enlaza el paquete y deja anotado en su comentario que H4 lo retirará al enlazarlo; `internal/httpx/superficie_test.go` con `TestSuperficieExportada`, que recorre con `go/parser` y `go/ast` todas las declaraciones exportadas del paquete y falla si alguna nombra `*http.Client`, `http.RoundTripper`, `*http.Request` o `*http.Response`, de modo que la imposibilidad de esquivar las garantías sea **por construcción** y no solo por control mecánico; y `.golangci.yml` actualiza **solo los comentarios** de la lista `red` para dejar constancia de que la regla ya tiene dueño, sin cambiar ninguna regla, sin activar ningún linter y sin añadir ninguna exclusión ni `//nolint` (FR-002, FR-053 a FR-055, SC-010, SC-014, punto 3 de la Definition of Done, D18).

- [ ] T017 [US5] Demostrar que los controles fallan ante cada intento de saltárselos, **sobre una copia desechable del árbol fuera del repositorio** y nunca sobre el árbol de trabajo: ejecutar el escenario 10 de `quickstart.md` en sus cuatro variantes —un `*http.Response` sin cerrar (`bodyclose`), un `http.NewRequest` sin contexto (`noctx`), una llamada a `Pedir` sin contexto (que no compila) y la importación de la biblioteca HTTP en un paquete del kernel, tal como la escribe el escenario 10.d del quickstart (que falla en `depguard` **y** en la subprueba R2)—, comprobando que cada una hace fallar `make ci` nombrando la regla incumplida, y dejar la salida literal de las cuatro como evidencia en `specs/003-h2-internal-httpx-cliente/gates/evidencia-sc012.md`; la tarea no modifica ningún fichero del árbol (FR-061, SC-012, obligación 6 del plan).

- [ ] T018 Cierre del hito: validación agregada y pendientes anotados: ejecutar los escenarios 1, 7, 11 y 12 de `quickstart.md` y dejar su resultado en `specs/003-h2-internal-httpx-cliente/gates/evidencia-cierre.md` —`make ci` en verde con `-race` y sin ninguna supresión nueva (el diff del fichero de configuración del linter frente a `main` solo añade comentarios y, como mucho, las palabras que el supuesto S2 obligara a ignorar en `misspell`: ninguna regla, ningún linter y ninguna exclusión), umbrales de cobertura respetados (global ≥ 70 %, `internal/core/**` ≥ 85 %, `internal/cli` ≥ 90 %) sin rebajar ninguno, medida **diferencial** de `git status --porcelain` sobre los directorios de fixtures idéntica antes y después de la ejecución, conjunto de verbos del binario distribuido y del binario de comprobación de extremo a extremo idéntico al de H1, y la orden **entera** de los prerrequisitos del quickstart que extrae toda dirección `http(s)://` de los tests y del material de test del paquete y confirma «solo direcciones locales»—; y actualizar `docs/PENDIENTES.md` moviendo «Antes de H2 · Dónde viven los fixtures grabados» a «Antes de la primera tarea `[datos]` de H4» con la misma recomendación, y anotando para H4 la retirada de `TestElBinarioNoEnlazaHTTPX` con la ampliación justificada de `modulosDelBinario` y que el ritmo por fuente (`ConIntervalo`) saldrá de la tabla de fuentes que ese hito crea (plan, obligación 8; H2 no la toca) (SC-008, SC-010, SC-011, SC-014, SC-015, puntos 1, 2 y 9 de la Definition of Done, obligaciones 7, 8, 10 y 11 del plan).

- [ ] T019 [plataforma] Publicar la rama del hito y abrir la propuesta de cambio: `git push -u origin h2-internal-httpx-cliente` y `gh pr create` con la plantilla del ritual (objetivo, alcance, controles añadidos, decisiones y pendientes), adjuntando como evidencia `specs/003-h2-internal-httpx-cliente/gates/evidencia-sc012.md` y `specs/003-h2-internal-httpx-cliente/gates/evidencia-cierre.md`, y leer después los estados de la integración continua y de Codecov con `gh pr checks`, dejando su resultado en `specs/003-h2-internal-httpx-cliente/gates/evidencia-plataforma.md`; **no** fusiona, **no** empuja a `main`, **no** fuerza y **no** crea etiquetas: la fusión (squash-merge) es siempre humana y el gancho `pre-push` la rechazaría igualmente (punto 1 de la Definition of Done, §6 del roadmap, ADR 0007).

---

## Definition of Done: qué punto cubre qué tarea

| # | Punto (`docs/ROADMAP.md` §1) | Tareas | Nota |
|---|---|---|---|
| 1 | `make ci` en verde | T001 … T018, verificado en T018 | Batería por tarea; el estado remoto lo lee T019 |
| 2 | Tests offline y fixtures | T003 … T015; fixtures en **T012 `[datos]`** | Cero red; ningún fixture grabado contra una fuente real (FR-044) |
| 3 | Sin `net/http`, `os.Exit`, `fmt.Print*` fuera de los paquetes autorizados | T016 | R2 pasa de «activa y vacía» a **activa con dueño** |
| 4 | Salida de applet validada contra `schemas/*.json` | — | **No aplica**: H2 no añade applet ni salida de applet (spec, *Fuera de alcance*). Ninguna tarea toca `schemas/` |
| 5 | Errores tipados con código de salida estable; sin `panic` | T001, T004, T014 | Tabla cerrada de dieciséis situaciones → {1, 2, 4, 5} |
| 6 | e2e (`testscript`) y `CHANGELOG.md` si cambia un comportamiento visible | — | **No aplica**: ningún comportamiento visible cambia (FR-062, SC-014), verificado en T016 y T018. Sin guion e2e nuevo: `testscript` solo observa binarios y el adaptador de prueba no se registra en ninguno |
| 7 | ADR si cambia una decisión de arquitectura | T002 | `docs/ADR/0011-ensayo-por-capas.md`: `Resultado.Ensayo` cambia el contrato del applet de ADR 0005 |
| 8 | Fila de `docs/SOURCES.md` y caso en `scripts/verify-sources.sh` | — | **No aplica**: H2 no consulta ninguna fuente real; la primera fila llega en H4. Ninguna tarea toca esos ficheros |
| 9 | Cobertura `internal/core/**` ≥ 85 % y global ≥ 70 % | T001, T018 | `internal/cli` ≥ 90 % se mantiene en T001; ningún umbral se rebaja |
| 10 | Evals de skill | — | **No aplica**: H2 no entrega ni cambia ninguna skill; es un hito de fundación (principio VIII) |
| 11 | Dimensión territorial | — | **No aplica**: H2 no la tiene; los hosts de los tests son ficticios (`fuente.prueba`, `otra.prueba`) |

---

## Trazabilidad: requisito → tarea

| Requisitos | Tareas |
|---|---|
| FR-001, FR-002 | T006, T016 |
| FR-003, FR-004, FR-005 | T006 |
| FR-006, FR-007, FR-008 | T005, T006 |
| FR-009 | T006 (`identificar.go`: decorador y constructor único de peticiones), T008 (cabecera en cada reintento), T009 (la petición de `robots.txt`, construida identificada y asertada en `robots_test.go`) |
| FR-010, FR-011, FR-012 | T006 |
| FR-013 … FR-018 | T003 (clave de sitio), T009 (reglas del sitio y decorador) |
| FR-019 … FR-022 | T003, T007 (limitador del sitio y decorador) |
| FR-023 … FR-028 | T008 |
| FR-029 … FR-032 | T004, T006, T009, T014 |
| FR-033, FR-034 | T004, T014 |
| FR-035 | T006 |
| FR-036 … FR-043 | T010, T011, T013 |
| FR-044 | T011, T012 (ninguna grabación real en todo el hito) |
| FR-045 … FR-049 | T012, T013 |
| FR-050, FR-051, FR-052 | T002, T006 |
| FR-053, FR-054 | T016 |
| FR-055 | T001, T002 (R1: el dominio gana `ConClase` y `Resultado.Ensayo` sin importar nada nuevo), T016 (R1 sigue sin cambio, vigilada por `depguard` y `compruebaDominioPuro`; plan.md, «Reglas de dependencia») |
| FR-056 | T006, T008, T009, T011 |
| FR-057 | T009, T018 |
| FR-058 | T007, T009 |
| FR-059 … FR-062 | T015, T016, T018 |
| FR-063 | T001, T004, T014 |
| FR-064 | T011 |
| FR-065 | T002, T006 |
| SC-001 T006 (recurso y saltos), T008 (reintentos), T009 (`robots.txt`) · SC-002 T006 · SC-003 T009 · SC-004 T007 · SC-005 T008 · SC-006 T014 · SC-007 T013 · SC-008 T011, T018 | |
| SC-009 T006-T011 · SC-010 T016, T018 · SC-011 T009, T018 · SC-012 T017 · SC-013 T006, T015 · SC-014 T016, T018 · SC-015 T018 · SC-016 T014 | |

Historias de usuario: **US1** T003, T005, T006, T007, T008, T009 · **US2** T001, T004, T014 · **US3** T012,
T013 · **US4** T010, T011 · **US5** T015, T016, T017 · **US6** T002, T006, T015.

---

## Dependencias y orden

El orden es **estrictamente secuencial**: ninguna tarea depende de una posterior.

1. **T001-T002** amplían el kernel de H1. Sin `schema.ConClase` el error del cliente no podría declarar su
   clase sin importar `internal/cli`; sin `Resultado.Ensayo` la descripción de `--dry-run` no tendría canal.
2. **T003-T005** fijan los tipos del paquete sin abrir ninguna conexión. `errores.go` (T004) necesita
   `Peticion` (T003) y `schema.ConClase` (T001); `agente.go` (T005) no necesita nada del paquete.
3. **T006-T009** montan la cadena de dentro afuera: primero la operación y el transporte, después el ritmo,
   después los reintentos y por último `robots.txt`, que se apoya en los dos anteriores. Cada tarea inserta
   su decorador y su opción en `cliente.go`, que por eso aparece declarado en las cuatro.
4. **T010-T013** añaden grabación y reproducción: el nombre antes que la escritura, la escritura antes que
   los fixtures y los fixtures antes que los tests que los leen.
5. **T014-T018** cierran: la tabla completa de clases de error cuando ya existen las dieciséis rutas, el
   adaptador de prueba, los controles de arquitectura y la validación agregada.
6. **T019** es la única que toca la plataforma remota y va la última.

**Paralelismo real**: escaso por diseño. Solo T002, T004, T005 y T010 no comparten fichero con su vecina
anterior y llevan `[P]`; T006 a T009 y T011 a T013 comparten `internal/httpx/cliente.go` y van en secuencia
obligada.

**MVP**: T001 … T009 entregan US1, US2 y US6 —el cliente responsable completo, sus errores tipados y el
ensayo—, que es el objetivo literal del hito («toda la red pasa por aquí»); T010 … T013 añaden US3 y US4
(la suite sin red) y T015 … T017 hacen demostrable US5.

---

## Notas

- **Todas las rutas van en la línea de la tarea.** El guardián de diff rechaza cualquier fichero fuera de
  ellas. Siempre permitidos sin declarar: `go.mod`, `go.sum`, `CHANGELOG.md` y el propio directorio del
  feature; declarar un fichero `.go` permite además su fichero de test; declarar un directorio permite todo
  lo que cuelga de él. Por eso `internal/httpx/ensayo_test.go` (T006), `internal/httpx/superficie_test.go`
  (T016) y `internal/httpx/adaptador_test.go` (T015) se declaran uno a uno: no tienen fichero de producto
  homónimo. `internal/httpx/identificar_test.go` se declara además en **T008** y **T009**, que amplían
  `TestIdentificacionEnTodaPeticion` con la subprueba de su escalón de la cadena sin tocar
  `identificar.go`, e `internal/httpx/sitio.go` se declara en **T003** (la struct y el mapa), **T007**
  (campo `limitador`) y **T009** (campo `reglas`), porque cada uno de esos tipos llega con su dependencia.
  Ninguna tarea declara un directorio genérico como `internal/httpx/`.
- **Una sola tarea `[datos]`: T012.** El material bajo `internal/httpx/testdata/` es material de test
  escrito a mano contra un host ficticio, **no** una grabación de una fuente real, así que el guardián
  exige la etiqueta pero el workflow no pausa (`docs/WORKFLOW.md`: material nuevo bajo
  `internal/<pkg>/testdata/`). La tarea no mezcla ningún otro trabajo y su lista de diez ficheros es
  cerrada. Ninguna tarea posterior puede corregir ese material: por eso T010 entra antes, para que los
  nombres estén ya fijados por `TestNombreDeGrabacion`, y T013 los ejercita sin tocarlos.
- **`KITLEGAL_RECORD` nunca se ejecuta fuera de un test.** Aparece solo en T011 y T013, siempre con
  `t.Setenv`, un `t.TempDir()` y un servidor local. En H2 no se graba ni un fixture real (FR-044); grabar
  contra una fuente es una tarea `[datos]` revisada de los hitos de fuente.
- **Ninguna tarea crea `testdata/<fuente>/` en la raíz, `schemas/`, `docs/SOURCES.md`,
  `scripts/verify-sources.sh`, `internal/source/`, `internal/cache`, `internal/store`, `internal/graph` ni
  `pkg/legalkit`.** Están fuera de alcance y crearlos vacíos sería un marcador de posición.
- **Los ficheros de H1 que se tocan son exactamente estos y ninguno más**:
  `internal/core/schema/error.go`, `internal/core/schema/sobre.go`, `internal/cli/errors.go`,
  `internal/app/main.go`, `cmd/kitlegal/main_test.go`, `internal/arch_test.go`, `Makefile` (solo la
  variable `LDFLAGS`, ningún objetivo), `.golangci.yml` (en T016 solo comentarios; en T003 solo una
  palabra bajo `misspell.ignore-rules` si S2 se cumple, y nada si no) y `docs/PENDIENTES.md`, más sus
  ficheros de test.
- **Las dos dependencias nuevas entran con el código que las usa**, no antes: `golang.org/x/time` en T007 y
  `github.com/temoto/robotstxt` en T009. Añadirlas sin importarlas dejaría `go mod tidy -diff` en rojo. No
  se añade ninguna otra, ni directa ni transitiva (FR-058).
- **Sin `CHANGELOG.md` y sin guion `testscript` nuevo**, por la razón que el spec fija en *Fuera de
  alcance*: H2 no cambia ningún comportamiento visible. El guion de H1 sobre `--dry-run`
  (`internal/app/testdata/script/verbo-obligatorio.txtar`) sigue vigente y ninguna tarea lo toca.
- **Nada de atajos** al reparar una verificación en rojo: ni desactivar linters, ni `//nolint`, ni tests
  saltados, ni errores silenciados, ni umbrales rebajados. La única supresión admitida en todo el hito es
  una entrada en `ignore-rules` de `misspell` si el español dispara un falso positivo, como en H1, y la
  añade **T003**, la única tarea que declara `.golangci.yml` por esa razón (T016 lo declara solo para
  comentarios). Si el falso positivo apareciera en una tarea posterior, esa tarea **no** edita un fichero
  que no declara: reescribe el término en español —no es un atajo, porque no toca ninguna regla ni
  suprime ningún hallazgo— y, si el término lo fija un contrato y no puede cambiarse, se anota en
  `gates/tarea-Tnnn.md` y la tarea se deja sin marcar (research D22 S2).
