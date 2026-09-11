# Tasks: H1 · Kernel CLI: multicall, flags globales, exit codes, sobre de salida

**Input**: `specs/002-h1-kernel-cli-multicall/` (spec.md, plan.md, research.md, data-model.md, contracts/,
quickstart.md)

**Prerrequisitos**: plan.md y spec.md leídos; `.specify/memory/constitution.md` aplicada; H0 cerrado
(módulo, binario, `Makefile`, `.golangci.yml`, flujos de integración continua y umbrales de cobertura ya
existen).

**Modo**: desatendido. Las tareas se ejecutan y verifican **una a una** por el workflow `hito`, con
guardián de diff por tarea. Por eso cada tarea es una **rebanada vertical** que incluye su comprobación y
deja el veredicto agregado en verde al terminar, y **declara en su propia línea todas las rutas** que va a
crear o modificar.

## Formato: `[ID] [P?] [Story] Descripción con rutas`

- **[P]**: la tarea no comparte fichero ni dependencia con sus vecinas (podrían repartirse si alguien las
  hiciera a mano; en el bucle del workflow van igual en secuencia).
- **[Story]**: historia de usuario del spec — US1 sobre citable · US2 un binario muchos nombres · US3 los
  fallos se distinguen sin leer el mensaje · US4 las mismas banderas en todos los applets · US5 el applet
  se describe a sí mismo · US6 las reglas de arquitectura se hacen cumplir solas.
- **[datos]**: la tarea toca material bajo un directorio `testdata`, lo que provoca **pausa para revisión
  humana**. Solo hay **una**: T014. No mezcla ningún otro trabajo.

## Batería de verificación por tarea

- **`make ci`** desde la primera tarea: formato · lint · tests con detector de carreras · vulnerabilidades
  · comprobación de esquemas · secretos · integridad de módulos · dependencias saneadas. Ya existe desde
  H0, así que ninguna tarea de H1 tiene que crearlo.
- `make ci` no modifica ningún fichero versionado, así que la verificación nunca ensucia el diff que juzga
  el guardián (SC-013).
- **Cero red en el código y en los tests** (FR-058). **`KITLEGAL_RECORD` no se usa en ningún punto del
  hito**: H1 no toca ninguna fuente externa, así que no hay nada que grabar y no nace ningún directorio de
  fixtures de fuente (plan.md, «Fixtures»).

---

## Phase 1: Fundación — el dominio y el supuesto bloqueante

**Objetivo**: fijar el único tipo del que todo depende y cerrar, antes de escribir kernel alguno, lo que
el plan no pudo verificar sin red.

- [X] T001 [US1] Dominio del sobre, puro y sin I/O: `internal/core/schema/sobre.go` con `Sobre` —struct cerrado, sin `omitempty`, con las seis claves `ok`, `fuente`, `url`, `fecha_consulta`, `hash` y `data` y ninguna más—, `Procedencia` (par fuente/url), `Resultado` (procedencia + datos) y la validación que rechaza `fuente` vacía, `url` vacía y `url` que no sea un URI absoluto; `internal/core/schema/huella.go` con la forma canónica de `data` en los tres pasos del contrato —serializar a JSON, releer conservando los números como literales (decodificador con `UseNumber`) y reserializar con las claves de todo objeto ordenadas, sin espacios ni saltos y sin escapar caracteres HTML— y la huella `sha256:` seguida de 64 dígitos hexadecimales en minúscula, con error —no pánico— si `data` no es serializable; `internal/core/schema/contexto.go` con `Contexto` y sus seis campos (`JSON`, `Timeout`, `Offline`, `DryRun`, `SinGrafo`, `Asunto`) y **sin** registrador ni `*slog.Logger`; `internal/core/schema/error.go` con `Clase` y sus seis constantes (`argumentos`, `no-encontrado`, `fuente-no-disponible`, `limite-o-tos`, `identidad-humana`, `inesperado`) y `DatosError` con las dos claves `clase` y `mensaje`; y las tablas de casos con `t.Parallel()` en `internal/core/schema/sobre_test.go` (`TestSobre`), `internal/core/schema/huella_test.go` (**`TestHuella`**, el nombre que invoca el escenario 4 de `quickstart.md`), `internal/core/schema/contexto_test.go` (`TestContexto`) y `internal/core/schema/error_test.go` (`TestClase`), que comprueban que el mismo contenido produce la misma huella con independencia del orden de las claves y del instante de la consulta, que un byte de diferencia produce otra distinta, que la huella no depende de `fecha_consulta`, que `fecha_consulta` se serializa en RFC 3339 con desplazamiento horario explícito y que el paquete importa **solo** `crypto/sha256`, `encoding/hex`, `encoding/json` y `time`; incorpora `github.com/stretchr/testify` al módulo con `go get` (FR-010 … FR-017, SC-005).

- [X] T002 [US3] Supuestos de Kong (bloqueante) y traducción de errores a códigos de salida: descargar `github.com/alecthomas/kong` con `go get` y comprobar los cinco supuestos S1-S5 de research.md D24 con `go doc` y con la prueba `internal/cli/supuestos_kong_test.go`, que fuerza `--help` y una bandera desconocida y comprueba que el análisis devuelve el control **sin terminar el proceso** y sin escribir en los descriptores del sistema —**S3 es bloqueante**: si Kong no lo permite, aplicar antes de seguir la contingencia de D24 (ayuda generada desde el registro y ayuda integrada de Kong desactivada)—, dejando por escrito el resultado de los cinco supuestos, la versión resuelta y la contingencia aplicada si la hubo en `specs/002-h1-kernel-cli-multicall/gates/supuestos-kong.md`; y `internal/cli/errors.go` con los sentinelas `ErrArgumentos`, `ErrNoEncontrado`, `ErrFuenteNoDisponible`, `ErrLimiteOTos` y `ErrIdentidadHumana`, la clasificación con `errors.Is` —envolver con `%w` no cambia la clase— y la **única** traducción `Clase` → código de salida en un `switch` sobre `Clase` **sin rama `default`** (0 correcto · 1 inesperado · 2 argumentos · 3 no encontrado · 4 fuente no disponible · 5 límite o términos de uso · 6 identidad humana), con la tabla de casos de las seis clases más el error desconocido y el error envuelto en `internal/cli/errors_test.go` (FR-029 … FR-032, SC-006 parcial).

**Checkpoint**: el dominio existe y es puro; la traducción de errores es exhaustiva y el análisis de la
línea de órdenes no puede terminar el proceso por su cuenta.

---

## Phase 2: Kernel — banderas, registro de eventos, sobre y autodescripción

**Objetivo**: que un applet no tenga que declarar ni banderas, ni sobre, ni códigos de salida, ni forma de
presentación.

- [X] T003 [P] [US3] Pre-escaneo acotado de los argumentos: `internal/cli/preescaneo.go` con `Preliminar{JSON, Verbose, Ayuda}` y `PreEscanear(args []string) Preliminar`, que reconoce exactamente `--json`, `--verbose` y `--help` en su forma larga y con valor booleano explícito (`--json=false`), se detiene en el terminador `--`, ignora cualquier otro token y **nunca falla ni consume argumentos**, y `internal/cli/preescaneo_test.go` con la tabla de formas reconocidas, ignoradas y posteriores al terminador (FR-045, base de SC-014).

- [X] T004 [P] [US4] Registro de eventos y contrato del presentador: `internal/cli/log.go` con el `*slog.Logger` estructurado que escribe **siempre** en el escritor de error, nivel resuelto con prioridad de `KITLEGAL_LOG` (`debug`, `info`, `warn`, `error`) sobre `--verbose` (equivale a `debug`) y `warn` por omisión, un valor inválido de la variable que **no aborta** la invocación y se avisa por el canal no filtrable, y sin registrar por omisión contenido que identifique a una persona física —applet, verbo, duración y clase de error; los argumentos solo en `debug`—; `internal/cli/presentador.go` con la interfaz `Presentador` que el kernel consume y que el paquete de presentación implementará (`Presentar(sobre, json) error`, `Texto(string) error`, `Aviso(string) error`, `Salida() io.Writer`, `Error() io.Writer`), de modo que el kernel no importa el paquete de presentación; y `internal/cli/log_test.go` con el doble de presentador en memoria que comprueba, en `TestRegistro`, niveles, prioridad de la variable sobre la bandera, aviso del valor inválido y que nada de esto llega a la salida estándar, y en **`TestRegistroPrivacidad`** —el nombre que invoca el escenario 7 de `quickstart.md`— lo que FR-039 exige y hasta ahora no verificaba ningún caso: que con el nivel por omisión (`warn`) y con `info` el registro emitido lleva applet, verbo, duración y clase de error pero **no** los argumentos de la invocación, y que esos argumentos aparecen **solo** cuando el nivel resuelto es `debug`, tanto si lo pide `--verbose` como si lo pide `KITLEGAL_LOG` (FR-025, FR-036 … FR-039, FR-040 parcial, SC-011 parcial).

- [X] T005 [US4] Las ocho banderas globales y la gramática de una invocación: `internal/cli/globales.go` con `Globales` y las ocho banderas declaradas **una sola vez** con sus etiquetas —`--json`, `--timeout` (duración, 30s por omisión), `--offline`, `--dry-run`, `--describe`, `--no-graph`, `--asunto` y `--verbose`— y su conversión a `schema.Contexto`; `internal/cli/parse.go` que construye y analiza la gramática de **una** invocación ya resuelto el applet (globales embebidas más los verbos que recibe como `any`, sin importar el paquete de composición), entrega a Kong los escritores del presentador, traduce todo fallo del análisis a `ErrArgumentos` —bandera desconocida, valor con formato inválido, argumento obligatorio ausente, `--timeout` no positivo o sin unidad— y devuelve la precedencia fija e independiente del orden de escritura: `--help` gana sobre todo (texto para personas en la salida estándar, código 0, sin que `--json` lo altere), después `--describe`, y en otro caso ejecución; y las tablas de casos en `internal/cli/globales_test.go` y `internal/cli/parse_test.go` más el `TestPreescaneo` de coincidencia en `internal/cli/preescaneo_test.go`, que comprueba que para toda invocación bien formada el pre-escaneo y el análisis coinciden en `--json`, `--verbose` y `--help` sobre todas las formas sintácticas que Kong acepta para una bandera booleana larga, declarando por escrito en `specs/002-h1-kernel-cli-multicall/gates/supuestos-kong.md` las que no se soporten (FR-018 … FR-021, FR-023, FR-024, FR-026 … FR-028, FR-042 parcial, FR-049 parcial, SC-010 parcial).

- [X] T006 [US1] Montaje del sobre, en éxito y en fallo, desde un único punto: `internal/cli/sobre.go` con la construcción del sobre de éxito a partir de `schema.Resultado` (procedencia del applet, `fecha_consulta` con reloj inyectable y huella sobre la forma canónica de `data`) y del **sobre de fallo** —mismas seis claves, `ok: false` si y solo si el código de salida no es 0, `data` exactamente `{clase, mensaje}`, procedencia de la fuente consultada cuando se conoce y el espacio reservado del kernel (`kitlegal.cli` / `kitlegal:cli`) cuando no, `fecha_consulta` y huella calculadas igual que en el de éxito—, emitido por el mismo punto que traduce el error a código de salida y nunca por un applet, sin emitir un segundo sobre si la escritura del primero falló; y `internal/cli/sobre_test.go`, cuyo `TestContratoSobre` —el nombre que invoca el escenario 6 de `quickstart.md`— valida los sobres emitidos contra la descripción formal del contrato (`contracts/sobre-de-salida.md` §6) compilada con `github.com/santhosh-tekuri/jsonschema/v6` **con `AssertFormat()` activado** —sin el cual un `url` vacío pasaría—, y cubre las seis clases de error con `--json`, incluidos los dos fallos anteriores a la ejecución del applet (FR-014 … FR-017, FR-045, SC-014, SC-015 parcial).

- [X] T007 [US5] Autodescripción del applet: `internal/cli/describe.go` que emite en la salida estándar **un** esquema JSON válido del borrador 2020-12 generado con `github.com/invopop/jsonschema` a partir de la definición del verbo y de los tipos del sobre —sin ninguna parte mantenida a mano—, con `entrada` (los argumentos del verbo más las ocho banderas globales) y `salida` (el sobre completo con `additionalProperties: false` y `data` **condicionado a `ok`** mediante `if`/`then`/`else` y `$defs/DatosError`), que excluye la ejecución y termina con código 0; y `internal/cli/describe_test.go`, cuyo `TestDescribe` —el nombre que invoca el escenario 6 de `quickstart.md`— compila el esquema emitido con `santhosh-tekuri/jsonschema/v6` y `AssertFormat()`, comprueba que un sobre de éxito y un sobre de fallo de ese mismo verbo validan contra él, y que quitar la rama `else` hace fallar el caso del sobre de fallo (FR-046 … FR-049, SC-007, SC-015).

**Checkpoint**: el kernel sabe analizar, fallar, envolver y describirse; nada de ello lo declara un applet.

---

## Phase 3: Presentación — un único escritor

- [X] T008 [US1] El presentador, único escritor de los dos descriptores: `internal/render/render.go` con `Nuevo(stdout, stderr)`, la implementación de la interfaz que consume el kernel y la regla de que **toda** escritura comprueba y propaga su error (incluido el vaciado final); `internal/render/json.go` con el sobre serializado sin escapar caracteres HTML, con salto de línea final y **nada más** en la salida estándar; `internal/render/tabla.go` con la tabla mínima —las cuatro líneas de procedencia `fuente`, `url`, `fecha_consulta` y `hash` seguidas del contenido de `data` aplanado a pares ruta/valor, donde la ruta concatena claves e índices con `.`—; `internal/render/texto.go` con el texto para personas: las tres líneas de `version` y la ayuda derivada del registro en la salida estándar, y los avisos —mensaje de un fallo, lista de applets disponibles y descripción de `--dry-run`— en la salida de error, sin pasar por el registro de eventos y por tanto sin que ningún nivel pueda ocultarlos; y los tests `internal/render/render_test.go`, `internal/render/json_test.go`, `internal/render/tabla_test.go` y `internal/render/texto_test.go`, incluido `TestEscrituraFallida` con un `io.Writer` que siempre devuelve error, que comprueba que el fallo se propaga, que no se intenta un segundo sobre por el descriptor roto y que no hay pánico (FR-040 … FR-044, SC-011).

**Checkpoint**: todo lo que el binario escribe pasa por un único paquete, y un fallo de escritura es un
error propagado, no un pánico.

---

## Phase 4: Composición — registro, despacho multicall y punto de entrada

- [X] T009 [US2] Contrato del applet y registro: `internal/app/applet.go` con las interfaces `Applet` (`Nombre`, `Descripcion`, `Verbos`) y `Argumentos` (`Ejecutar(ctx, schema.Contexto, *slog.Logger) (schema.Resultado, error)`) y el struct `Verbo` (`Nombre`, `Descripcion`, `Argumentos` como **fábrica** que devuelve un valor nuevo en cada invocación, `Salida` como valor cero solo para reflexión y `PorOmision`); `internal/app/registro.go` con `Registrar`, `Buscar`, `Nombres` ordenados y `RegistroDeProduccion()` **vacío** —el primer applet de producto llega en H4—, rechazando **al construirse** un nombre vacío, con espacios, con prefijo `-`, duplicado o igual a un verbo reservado del binario (`version`), un applet sin verbos, con nombres de verbo repetidos o con **más de un** verbo marcado por omisión, y convirtiendo la raíz de composición ese error en fallo de arranque y nunca en código de salida de usuario; y `internal/app/registro_test.go` con la tabla de las **ocho causas de rechazo** que recogen las cinco reglas de validación de `contracts/registro-y-describe.md` §1 —nombre vacío, con espacios, con prefijo `-`, duplicado, igual a un verbo reservado, applet sin verbos, verbos repetidos y dos verbos por omisión— y la comprobación de que el registro de producción está vacío (FR-001, FR-008, FR-009 parcial, FR-044, SC-010 parcial, SC-012 parcial).

- [X] T010 [US2] Despacho multicall, verbo por omisión y ayuda derivada del registro: `internal/app/despacho.go` con la precedencia fijada —el último componente de `os.Args[0]` sin sufijo `.exe`; si ese nombre está registrado **manda él** y los argumentos se entregan íntegros al applet; si no lo está, incluido `kitlegal`, el applet se toma del primer argumento; sin primer argumento o con un primer argumento que no es applet, código 2 con el nombre desconocido y la lista de applets en la salida de error; los verbos reservados (`version`) se reconocen antes del registro; un nombre de enlace desconocido no inutiliza el binario— y la normalización del verbo **antes** de construir la gramática (si el primer argumento restante no nombra un verbo del applet y hay verbo por omisión, se inserta a la cabeza; si no lo hay, código 2 con la lista de verbos; si se pidió la ayuda del applet, no se inserta nada); `internal/app/ayuda.go` con la ayuda del binario y la de cada applet derivadas **solo** del registro, sin ninguna lista paralela mantenida a mano; y `internal/app/despacho_test.go` y `internal/app/ayuda_test.go` con la tabla de los cinco casos de despacho, la de `TestVerboPorOmision` (con verbo por omisión, sin él, con el argumento que se llama como un verbo y con `--help`) y la comprobación de que registrar un applet lo hace aparecer en la ayuda sin tocar nada más (FR-002 … FR-007, FR-026, SC-003 parcial).

- [X] T011 [US3] Raíz de composición del kernel y tabla completa de códigos de salida: `internal/app/main.go` con `Main(argv []string, registro, stdout, stderr io.Writer, version, commit, fecha string) int`, que monta el presentador del paquete de presentación y lo inyecta en el kernel, resuelve el nivel del registro de eventos, aplica `--timeout` como plazo del `context.Context` de toda la operación —agotarlo es código 4—, atiende el verbo reservado `version` por el presentador sin sobre ni banderas y con el mismo formato de tres líneas de H0, entrega al applet el contexto de ejecución con las seis decisiones globales, honra `--dry-run` sin cortar antes del applet —descripción del applet, el verbo y los argumentos que se habrían ejecutado en la salida de error, siempre visible, salida estándar vacía también con `--json`, código 0—, y traduce en un **único** punto el error a código de salida emitiendo a la vez el sobre de fallo, **sin llamar nunca a `os.Exit`** y sin pánico en ninguna ruta de usuario; y `internal/app/codigos_test.go` con **`TestCodigoSalida`** —cuyos subcasos se nombran por la clase (`correcto`, `inesperado`, `argumentos`, `no-encontrado`, `fuente-no-disponible`, `limite-o-tos`, `identidad-humana`), de modo que `-run 'TestCodigoSalida/inesperado'` del escenario 5 de `quickstart.md` seleccione un caso que existe— y **`TestSobreDeFallo`**, que fuerzan desde applets de prueba declarados en el propio test los códigos 0, 1, 2, 3, 4, 5 y 6 y comprueban para cada uno el código, que el mensaje va a la salida de error y no a la estándar, y que con `--json` la salida estándar lleva un único documento JSON con las seis claves, `ok` falso y la clase y el mensaje dentro de `data` —incluidos los dos fallos anteriores a la ejecución del applet, bandera desconocida y applet no registrado, con procedencia `kitlegal.cli` / `kitlegal:cli`—; y `internal/app/main_test.go` con los dos casos que hasta ahora solo cubría un escenario manual: **`TestDryRun`**, que comprueba las tres reglas de FR-022 —salida estándar vacía **también con `--json`**, descripción del applet, verbo y argumentos en la salida de error y **visible con `KITLEGAL_LOG=error`**, porque no pasa por el registro de eventos, y código 0— y **`TestPlazoAgotado`**, que comprueba la mitad de FR-020 que ningún caso verificaba: con un applet de prueba que espera más que el plazo, **vencer `--timeout`** produce el código 4 y su sobre de fallo, sin depender de que el applet devuelva el error tipado (FR-018, FR-020, FR-022, FR-030, FR-033 … FR-035, FR-045, SC-006, SC-011, SC-014).

- [X] T012 [US6] Punto de entrada reducido y control de escrituras endurecido: `cmd/kitlegal/main.go` reducido a inyectar los descriptores del sistema y el registro de producción —`os.Exit(app.Main(os.Args, app.RegistroDeProduccion(), os.Stdout, os.Stderr, version, commit, fecha))`—, de modo que deja de escribir en la salida estándar y las variables `version`, `commit` y `fecha` siguen inyectándose por `-ldflags`; `cmd/kitlegal/main_test.go` adaptado **conservando exactamente la parte del contrato que no cambia: las tres líneas de `version` y su código 0** (D16, `contracts/cli-version.md` de H0). Lo que H0 resolvía con la línea `uso: kitlegal version` y código 2 —verbo ausente, verbo desconocido, argumento sobrante— **deja de ser contrato de `version`** y pasa a regirse por el despacho que fijó T010: mensaje que nombra el applet desconocido más la lista de applets registrados —**vacía** en el registro de producción— en la salida de error, con código 2 (FR-006, `contracts/registro-y-describe.md` §2 y §3). El test se adapta a ese mensaje; mantener la aserción de H0 contradiría FR-006 y es el atajo que esta tarea no toma; y `.golangci.yml` retirando **entera** la exclusión de `errcheck` para `fmt.Fprint`, `fmt.Fprintf` y `fmt.Fprintln`, que a partir de aquí taparía el único camino que el contrato obliga a vigilar —retirar una exclusión endurece el control y no es una exclusión nueva de las que FR-057 prohíbe— (FR-035, FR-040, FR-052 parcial, FR-057, D16, D27).

**Checkpoint**: el binario distribuido ya es el kernel; su registro está vacío y `kitlegal version` sigue
respondiendo lo mismo que en H0, ahora por el presentador.

---

## Phase 5: La entrega literal del hito — applets de ejemplo y e2e

- [X] T013 [US6] Los applets de ejemplo pasan el mismo listón que el resto del código: `Makefile` con la variable `TESTDATA_PKGS` **derivada con `wildcard`** de los subdirectorios de `internal/app/testdata` que contienen ficheros Go —misma convención que `TOOL_MODULES` de H0, de modo que la orden queda en verde tanto ahora, cuando no existe ninguno, como después, sin volver a tocar el fichero— y añadida explícitamente a los objetivos `lint`, `fmt` y `fmt-check`, porque los comodines de Go no descienden a ese directorio y sin enumerarlos el código que copiará cada applet posterior sería el único que no pasa el listón (FR-050 parcial, D19, obligación 2 del plan).

- [X] T014 [datos] [US1] Material de test del hito, bajo `internal/app/testdata/`: paquete `internal/app/testdata/ejemplo/` con los **dos** applets de ejemplo —`echo`, con un único verbo `repetir` marcado por omisión, `data` derivada de sus argumentos, procedencia `kitlegal.echo` / `kitlegal:applet/echo`, sin red y sin tocar disco; y un segundo applet con **dos** verbos y **ninguno** por omisión, de modo que invocarlo sin nombrar verbo termina en código 2—, ambos sin ninguna excepción de lint y sin `os.Exit`, `os.Stdout`, `os.Stderr` ni `fmt.Print*`; `internal/app/testdata/kitlegal-e2e/main.go`, `package main` que es la raíz de composición del binario de e2e —kernel real más el registro de ejemplo, línea por línea el mismo `main()` que el distribuido salvo qué applets se registran—; y los guiones `internal/app/testdata/script/*.txtar` que describen la entrega del hito: `kitlegal echo hola --json` con las seis claves y código 0, el enlace simbólico `echo -> $KITLEGAL_BIN` con salida y código idénticos a los de la invocación con el applet como primer argumento, `--help` con la lista derivada del registro y código 0, código 2 con argumentos malos, `--describe` con esquema válido y código 0, el segundo applet sin verbo con código 2, y que con `--json` y el registro de eventos al máximo detalle la salida estándar contiene **solo** el documento JSON; verificado construyendo el binario a mano con `go build ./internal/app/testdata/kitlegal-e2e` en un temporal y contrastando salida, descriptores y códigos con lo que afirma cada guion (FR-009, SC-001, SC-003, SC-010). Los guiones son **la definición ejecutable de la entrega** que pide el ritual de `docs/ROADMAP.md` §6.1, pero se escriben aquí y no como primera tarea: eso **no cumple** el ritual, lo invierte, y queda registrado como divergencia consciente en la sección homónima de este documento.

- [X] T015 [US2] El e2e ejecuta la entrega, y el segundo applet demuestra que no hay código duplicado: `internal/app/e2e_test.go`, cuyo `TestMain` construye el binario con `go build` en un directorio temporal, lo antepone al `PATH`, lo expone al guion en `$KITLEGAL_BIN` y **retorna** en lugar de terminar el proceso —para que el borrado del temporal ocurra y el fichero no necesite ninguna excepción de lint—, y que ejecuta con `testscript.Run` y `RequireExplicitExec` los guiones `.txtar` del material de test del paquete de composición, sin red y sin escribir fuera de su temporal; `internal/app/hereda_test.go` con **`TestAppletHereda`** —el test que el escenario 7 de `quickstart.md` invoca por su nombre y que hasta ahora no escribía ninguna tarea: recorre los **dos** applets de ejemplo y comprueba, sobre los mismos casos para ambos, que cada uno acepta las **ocho** banderas globales sin declarar ninguna, que emite el sobre con las seis claves, que su tabla de códigos de salida es la misma, que responde a `--describe` y a `--help` y que las dos formas de presentación —JSON y tabla mínima— salen idénticas en estructura, **sin que el segundo applet declare nada más que nombre, verbos y contenido de `data`** (SC-010); un `-run` sobre un test inexistente termina en 0, así que la evidencia del escenario 7 depende de que este fichero exista—; y `Makefile` con el objetivo `test-e2e` que deja de anunciar el hito ausente y pasa a ejecutar ese paquete; incorpora `github.com/rogpeppe/go-internal` al módulo con `go get` (FR-055, SC-002, SC-003, SC-009, SC-010).

**Checkpoint**: `kitlegal echo hola --json`, el enlace simbólico y `--describe` funcionan y están
demostrados por un test, no por una afirmación.

---

## Phase 6: Las reglas de arquitectura se hacen cumplir solas

- [X] T016 [US6] Las cinco reglas de dependencia, activas y vigiladas en la capa que puede verlas —las tres de importación en dos capas independientes, las dos de símbolo en el lint con análisis de tipos (FR-051, `contracts/reglas-de-arquitectura.md` §2)—: `.golangci.yml` con `depguard` en sus **tres** listas —una por regla de importación—, escritas con patrones sobre la ruta absoluta del fichero —la lista `core`, acotada al árbol del dominio, denegando los ocho paquetes internos **y** los paquetes de entrada y salida de la biblioteca estándar (`log`, `log/slog`, `os`, `io`, `net/http`, `database/sql`); la lista `red`, denegando `net/http` en todo el árbol salvo en el paquete del cliente HTTP, que aún no existe; y la lista `sql`, denegando `database/sql` y el controlador de SQLite salvo en los tres paquetes de almacenamiento —caché, almacén y grafo—, que tampoco existen todavía— y con `forbidigo` en `analyze-types: true` y tres patrones marcados con la regla que vigilan, de modo que el fallo la nombre: `^os\.Exit$` (marca `R4:`), `^fmt\.Print(|f|ln)$` (marca `R5:`, **sin excepción alguna en todo el árbol**, con lo que desaparece la acotación por ruta que H0 dejó como provisional) y `^os\.Stdout$` / `^os\.Stderr$` (marca `R5-descriptores:`), con las excepciones escritas como `path` + `text` sobre la marca y acotadas **solo** a las dos raíces de composición —el punto de entrada del binario distribuido y el `package main` del binario de e2e—, sin alcanzar al kernel ni al paquete hermano de applets de ejemplo; e `internal/arch_test.go`, paquete de solo test cuyo **`TestArquitectura`** —el nombre que invoca el escenario 8 de `quickstart.md`— recorre el grafo **transitivo** real con `go list -deps` sobre los paquetes del módulo más los dos que enumera `TESTDATA_PKGS` y falla nombrando la regla violada, de modo que borrar las tres listas del lint no baste para que una violación de R1, R2 o R3 pase; R4 y R5 **no** entran en este test, porque una llamada a `os.Exit` o una referencia a `os.Stdout` no aparecen en el grafo de dependencias y fingir que las cubre daría una garantía falsa (FR-050 … FR-053, FR-035, FR-040, SC-008).

**Checkpoint**: quien intente saltarse una regla recibe un fallo que la nombra —las tres de importación
en el lint y otra vez en los tests; las dos de símbolo en el lint, que es la única capa que puede verlas—.

---

## Phase 7: Cobertura, documentación y evidencia

- [X] T017 Cobertura de `internal/cli` ≥ 90 %, por el mismo control que ya usa el proyecto: `codecov.yml` con un componente `internal_cli` sobre el árbol del kernel con estado de tipo `project`, objetivo 90 % e `informational: false` —bloqueante, sin exclusiones y sin tocar los umbrales de H0—, y los tests que falten para alcanzarlo en `internal/cli/errors.go`, `internal/cli/preescaneo.go`, `internal/cli/log.go`, `internal/cli/presentador.go`, `internal/cli/globales.go`, `internal/cli/parse.go`, `internal/cli/sobre.go` y `internal/cli/describe.go`, comprobando en local con `go tool cover -func` sobre el perfil que genera `make test` que el paquete llega al umbral; nunca se rebaja el objetivo para que el estado pase (FR-056, SC-004).

- [X] T018 [P] Decisiones de arquitectura registradas: `docs/ADR/0005-contrato-de-applet.md` en formato MADR corto con el contrato `Applet`/`Verbo`/`Argumentos`/`Resultado`, el verbo por omisión y el reparto entre kernel, composición, dominio y presentación con la dirección de dependencia que lo hace posible; y `docs/ADR/0006-sobre-de-salida-y-huella.md` con la forma canónica de `data` y el algoritmo de la huella con prefijo de algoritmo, el espacio de nombres reservado `kitlegal.` / `kitlegal:` —anotando la prohibición que nace en H4: ningún adaptador de fuente puede usar ese prefijo ni ese esquema, y la comprobación mecánica se escribe con el primer adaptador— y la forma del sobre de fallo (Definition of Done §1.7, D23, obligación 6 del plan).

- [X] T019 [P] Registro del cambio visible: `CHANGELOG.md`, sección *Unreleased*, con el kernel CLI —multicall por nombre de invocación o primer argumento, las ocho banderas globales, los códigos de salida estables con el 1 reservado al fallo inesperado, el sobre de salida con huella reproducible, la autodescripción y la tabla mínima—, la retirada de la exclusión de `errcheck` y las cinco dependencias nuevas, siguiendo la convención de changelog que fijó H0 (Definition of Done §1.6).

- [ ] T020 Validación local y demostración de que los controles están vivos, **sin escribir un solo fichero dentro del repositorio salvo la evidencia**: ejecutar los doce escenarios de `quickstart.md` y registrar la evidencia —orden, salida y veredicto— en `specs/002-h1-kernel-cli-multicall/gates/quickstart-h1.md`, **única ruta que esta tarea crea o modifica**. Los escenarios 8 y 9, que son los únicos que necesitan introducir violaciones deliberadas, se ejercen sobre una **copia desechable del árbol fuera del repositorio** (`$(mktemp -d)`, sin `.git`, creada con el `tar` del escenario 8) que se destruye al terminar: así ningún fichero de violación llega a `internal/**` ni a `internal/app/testdata/**` —material protegido por el guardián, que obligaría a etiquetar `[datos]` esta tarea y a mezclar validación con material de test— y `git status` queda limpio **por construcción**. Sobre esa copia: romper **a propósito y una por una** las cinco reglas de dependencia y comprobar que **las dos capas fallan en R1, R2 y R3** y que **el lint falla en R4 y R5** —reglas de símbolo que el grafo de `go list -deps` no ve—, nombrando en los cinco casos la regla violada (escenario 8, obligación 3 del plan); e introducir un defecto deliberado en el paquete de applets de ejemplo **de la copia** para comprobar que el lint no excluye ese material por omisión y que la excepción de la raíz de composición **no** le alcanza —`os.Exit` y los descriptores del sistema fallan ahí y `fmt.Print*` falla en las dos rutas— (escenario 9, obligaciones 2 y 8). Y medir la duración del flujo agregado con el e2e dentro comparándola con el umbral de 3 minutos que fijó H0 (obligación 4); el escenario 12 solo produce `bin/`, que está en `.gitignore` desde H0 (SC-002, SC-004 … SC-013).

- [ ] T021 Evidencia en la propuesta de cambio: comprobar que el estado de cobertura del componente `internal_cli` aparece y **bloquea**, igual que se hizo con el del dominio en H0 —si no apareciera, la causa es la configuración de la plataforma de cobertura, nunca el umbral—, que el flujo de integración continua termina en verde con todos los controles de H0 más los que añade H1 y sin exclusiones nuevas sin justificar, y registrar el resultado, la duración del flujo y el enlace a la propuesta en `specs/002-h1-kernel-cli-multicall/gates/pr-h1.md` (SC-004, SC-013, FR-057, obligación 5 del plan).

---

## Dependencias y orden de ejecución

El orden es **estrictamente secuencial y ejecutable**: ninguna tarea depende de una posterior.

- **T001** no depende de nada: `internal/core/schema` es lo único de lo que todo lo demás depende y no
  depende de nada. `make ci` ya existe desde H0, así que no hay esqueleto que crear.
- **T002** va antes que cualquier código que use Kong (T005 en adelante), como exige la obligación 1 del
  plan; se empareja con `errors.go`, que no depende de Kong pero sí es lo que el resto del kernel usa para
  fallar y lo que la contingencia S5 necesita si el error del análisis no fuera distinguible por tipo.
- **T003** y **T004** son ficheros disjuntos entre sí y solo dependen de T001 y T002.
- **T005** depende de T002 (supuestos cerrados, `ErrArgumentos`) y de T004 (los escritores que se entregan
  a Kong); **T006** depende de T001 y T002; **T007** depende de T006 (el sobre ya fijado) y de T005 (la
  gramática ya fijada), en ese orden y no al revés.
- **T008** depende de T004 (la interfaz que implementa) y de T001 (el sobre que presenta). No retira la
  exclusión de `errcheck`: eso es T012, porque hasta que `cmd/` deje de escribir la retirada dejaría el
  lint en rojo.
- **T009** → **T010** → **T011** es la composición: registro, despacho, raíz. T011 necesita T008 (inyecta
  el presentador) y todo el kernel.
- **T012** cierra el punto de entrada y **solo entonces** endurece `errcheck`.
- **T013** va antes que **T014** a propósito: cuando se creen los paquetes de ejemplo, `make ci` ya los
  lintará, de modo que nacen cumpliendo el listón y ninguna tarea posterior necesita corregirlos —lo que
  el guardián rechazaría, por ser material protegido fuera de una tarea `[datos]`—.
- **T015** necesita T014 (los guiones, los dos applets de ejemplo y el binario del e2e) y es **la primera
  tarea que puede dejar el e2e en verde**; por eso el e2e se ejecuta aquí aunque los guiones que describen
  la entrega se escriban en T014. El ritual §6.1 pide lo contrario —el guion primero, en rojo—: esa
  inversión **no lo cumple**, y está registrada como divergencia consciente más abajo.
- **T016** va al final de la implementación porque las reglas se refieren a paquetes que hasta T015 no
  existen todos, incluidos los dos enumerados en `TESTDATA_PKGS`.
- **T017** a **T019** no alteran ningún control: `make ci` sigue verde. **T020** y **T021** convierten los
  criterios de éxito en evidencia registrada.

### Oportunidades de paralelismo

Marcadas con **[P]**: T003, T004, T018 y T019. En el bucle del workflow se ejecutan igual en secuencia
(una tarea, un guardián, una verificación); la marca indica que no comparten fichero ni dependencia.

---

## Divergencias conscientes de este plan de tareas

> Lo mismo que `plan.md` §*Complexity Tracking* hace con el diseño, aquí con el orden de ejecución: lo que
> se aparta de la lectura literal del plan, del roadmap o de la constitución se registra **como divergencia
> y no como cumplimiento**, con su alternativa rechazada (criterio de decisión autónoma §3). Ninguna afecta
> a alcance, frontera humana, privacidad, términos de uso ni a una decisión cerrada, así que ninguna
> dispara el §4 (escalar).

| Divergencia | Por qué | Alternativa rechazada |
|---|---|---|
| **El guion e2e que describe la entrega se escribe en T014, la decimocuarta de veintiuna tareas, y no la primera.** `docs/ROADMAP.md` §6.1 y la constitución §III piden que el hito **empiece** por el test e2e que describe la entrega, y `plan.md` §*Orden de implementación* lo sitúa como paso 1. Esta inversión **no cumple** ese ritual: lo pospone. | El modo desatendido añade una restricción que el ritual no contempla: **ninguna tarea puede terminar con `make ci` en rojo**, porque el bucle del workflow verifica tarea a tarea y un rojo detiene el hito. Un guion `.txtar` escrito el primer día falla —no hay binario, ni applet, ni registro— y dejaría en rojo las trece tareas siguientes. Además el material vive bajo `internal/app/testdata/`, que solo puede tocarse en una tarea `[datos]` con pausa humana: escribirlo primero y **volver a tocarlo** al final para ajustarlo exigiría dos pausas, o bien una corrección del material protegido desde una tarea que no es `[datos]`, que el guardián rechaza. | **Escribir el guion en T001 y dejarlo en rojo hasta T015**: es la lectura literal del ritual, y se rechaza porque hace indistinguible el rojo esperado del rojo por defecto, que es justo lo que el modo desatendido no puede permitirse. **Escribir el guion primero marcado como saltado** (`t.Skip`, o el guion fuera del directorio que el e2e recorre) y activarlo en T015: mantiene el verde, pero convierte la definición ejecutable de la entrega en un fichero inerte durante trece tareas —un test saltado es exactamente el atajo que la sección *Notas* prohíbe— y sigue exigiendo tocar dos veces material protegido. Lo que **sí** se conserva del ritual: T014 escribe los guiones **antes** de que exista el e2e que los ejecuta (T015), y los contrasta construyendo el binario a mano; la entrega se demuestra ejecutándose, no afirmándose. |
| **La validación local (T020) se ejerce sobre una copia desechable del árbol fuera del repositorio**, en lugar de introducir y revertir las violaciones en el árbol de trabajo, como admitía la lectura más simple de la obligación 3 del plan. | Dos de las siete violaciones cuelgan de `internal/app/testdata/`, protegido por la capa 3 de la constitución: crearlas en el árbol convertiría T020 en una tarea `[datos]` con pausa humana y mezclaría validación, medición y evidencia con material de test, que es lo que el criterio de datos separados prohíbe. Sobre una copia sin `.git`, el árbol de trabajo queda limpio por construcción y no por acordarse de borrar siete ficheros. | **Declarar las siete rutas en T020 y revertirlas al final**: cumple la regla de rutas declaradas, pero mantiene la mezcla con material protegido y hace que la limpieza dependa de que ningún paso falle a mitad. |

---

## Trazabilidad

### Historias de usuario → tareas

| Historia | Prioridad | Tareas | Prueba independiente |
|---|---|---|---|
| US1 · Toda respuesta es un sobre citable | P1 | T001, T006, T008, T014 | Invocar el applet de ejemplo con `--json`: un único JSON con las seis claves, huella reproducible y los cuatro datos de procedencia también en la tabla mínima |
| US2 · Un binario, muchos nombres | P2 | T009, T010, T014, T015 | Enlace simbólico con el nombre de un applet registrado: salida, salida de error y código idénticos a la invocación con el applet como primer argumento |
| US3 · Los fallos se distinguen sin leer el mensaje | P3 | T002, T003, T011 | Forzar cada clase de error y comprobar código, descriptor del mensaje y sobre de fallo con `--json` |
| US4 · Las mismas banderas en todos los applets | P4 | T004, T005, T011, T015 | Dos applets de prueba distintos aceptan las ocho banderas con idéntica sintaxis sin declarar ninguna (`TestAppletHereda`, T015) |
| US5 · El applet se describe a sí mismo | P5 | T007, T014 | `--describe` emite un esquema que un validador acepta y contra el que valida tanto un sobre de éxito como uno de fallo |
| US6 · Las reglas de arquitectura se hacen cumplir solas | P6 | T012, T013, T016, T020 | Romper cada regla en una copia desechable del árbol fuera del repositorio y ver fallar el control que la vigila —las dos capas en R1, R2 y R3; el lint en R4 y R5— nombrando la regla |

### Requisitos → tareas

| Requisitos | Tarea |
|---|---|
| FR-010 … FR-013, FR-015 … FR-017 | T001 |
| FR-029, FR-031, FR-032 | T002 |
| FR-045 (pre-escaneo que lo hace posible) | T003 |
| FR-025, FR-036 … FR-039 | T004 |
| FR-018 … FR-021, FR-023, FR-024, FR-026 … FR-028 | T005 |
| FR-014, FR-045 (montaje del sobre de fallo) | T006 |
| FR-046 … FR-049 | T007 |
| FR-040 … FR-044 | T008 |
| FR-001, FR-008, FR-044 | T009 |
| FR-002 … FR-007, FR-026 (ayuda derivada) | T010 |
| FR-020 (plazo aplicado), FR-022, FR-030, FR-033 … FR-035 | T011 |
| FR-035, FR-040, FR-057 (endurecimiento de `errcheck`) | T012 |
| FR-050 (los paquetes de ejemplo entran en el lint) | T013 |
| FR-009 | T014 |
| FR-055 | T015 (e2e), que además crea `TestAppletHereda` para SC-010 |
| FR-050 … FR-053 | T016 |
| FR-056 | T017 |
| FR-054 | T001 … T011 (cada tarea trae sus unitarios: error→código, sobre, despacho, `--describe`) |
| FR-057 | T012, T016, T020, T021 |
| FR-058 … FR-060 | todas: ninguna tarea hace red ni escribe fuera del directorio de construcción y de los temporales de test, y las cinco dependencias nuevas son las de la lista cerrada (T001 `testify`, T002 `kong`, T006 `santhosh-tekuri/jsonschema/v6`, T007 `invopop/jsonschema`, T015 `rogpeppe/go-internal`) |

### Criterios de éxito → tareas

| Criterio | Tarea |
|---|---|
| SC-001 | T014 (guion) + T015 (ejecución) |
| SC-002 | T014, T015, T020 |
| SC-003 | T010, T014, T015 |
| SC-004 | T017, T020, T021 |
| SC-005 | T001 |
| SC-006 | T002, T011 |
| SC-007 | T007, T014, T015 |
| SC-008 | T016, T020 |
| SC-009 | T015 |
| SC-010 | T005, T009, T014 (los dos applets de ejemplo) + T015 (`TestAppletHereda`, que lo comprueba) |
| SC-011 | T004, T008, T011 |
| SC-012 | T001, T005 … T011 (ningún paquete queda como marcador de posición) |
| SC-013 | T020, T021 |
| SC-014 | T006, T011 |
| SC-015 | T006, T007 |

### Definition of Done (`docs/ROADMAP.md` §1) → tareas

| # | Punto | Cómo se cumple |
|---|---|---|
| 1 | `make ci` en verde | Existe desde H0; **toda** tarea de H1 lo deja en verde, y T020 y T021 lo comprueban en local y en la propuesta de cambio |
| 2 | Tests unitarios offline; fixtures grabados si toca red | T001 … T011 (tablas de casos con `t.Parallel()` y escritores en memoria). **Sin fixtures de fuente**: H1 no toca ninguna red, así que no nace ningún directorio de fixtures de fuente y `KITLEGAL_RECORD` no se usa. El único material bajo un directorio `testdata` es código Go y guiones (T014), y por eso esa tarea va etiquetada `[datos]` |
| 3 | Sin `net/http`, `os.Exit` ni `fmt.Print*` fuera de los paquetes autorizados | T012 (el punto de entrada deja de escribir) y T016 (`depguard` + `forbidigo` + test de arquitectura), demostrado en T020 |
| 4 | Toda salida de applet validada contra su descripción formal | T006 y T007: los sobres de éxito y de fallo se validan con `santhosh-tekuri/jsonschema/v6` y `AssertFormat()` contra el contrato y contra el esquema que emite `--describe`. **El directorio de esquemas versionado no existe en H1**: es alcance de H4 (borrador) y H11, y `make schema-check` sigue anunciando el hito que lo aporta |
| 5 | Errores tipados → código de salida estable; ningún pánico en rutas de usuario | T002 (traducción exhaustiva) y T011 (único punto de traducción, `Main` devuelve `int`) |
| 6 | Comportamiento visible → e2e y `CHANGELOG.md` | T014 y T015 (e2e con `testscript`) y T019 (sección *Unreleased*) |
| 7 | Decisión de arquitectura → ADR | T018 (`0005-contrato-de-applet.md` y `0006-sobre-de-salida-y-huella.md`) |
| 8 | Fuente externa → fila en el inventario de fuentes y caso de verificación | **No aplica**: H1 no toca ninguna fuente. `docs/SOURCES.md` y `scripts/verify-sources.sh` nacen en H4, con el primer adaptador; T018 deja anotada la prohibición que ese hito tendrá que hacer mecánica (el espacio de nombres reservado) |
| 9 | Cobertura de `internal/core/**` ≥ 85 % y global ≥ 70 % | T001 (los tests del dominio) y los de cada tarea; los umbrales de H0 no se tocan y T017 añade el de `internal/cli` ≥ 90 % que exige este hito |

---

## Estrategia de implementación

1. **T001 y T002 son el cimiento y el riesgo**: el dominio que todo importa y el único supuesto bloqueante
   del plan. Si algo va a obligar a corregir el diseño, es S3.
2. **T003 a T008** escriben el kernel y la presentación: al terminar, un applet ya no tiene nada que
   declarar salvo su nombre, sus verbos y su `data`.
3. **T009 a T012** componen el binario y cierran el punto de entrada.
4. **T013 a T015** son la entrega literal del hito y la primera vez que se demuestra ejecutándose.
5. **T016** convierte las reglas en controles, y **T017 a T021** cierran cobertura, documentación y
   evidencia.

**MVP**: T001 … T008 más T009 … T011 entregan US1, US3 y US4 (el sobre, los códigos de salida y las
banderas) sobre applets de prueba; T014 y T015 añaden US2 y US5 y hacen la entrega demostrable.

---

## Notas

- **Todas las rutas van en la línea de la tarea.** El guardián de diff rechaza cualquier fichero fuera de
  ellas. Siempre permitidos sin declarar: `go.mod`, `go.sum`, `CHANGELOG.md` y el propio directorio del
  feature; declarar un fichero `.go` permite además su fichero de test; declarar un directorio permite
  todo lo que cuelga de él.
- **Una sola tarea `[datos]`: T014.** El material bajo `internal/app/testdata/` es **código de test y
  guiones, no fixtures grabados de una fuente**, pero el guardián protege por igual cualquier ruta bajo un
  directorio `testdata`, así que la tarea lleva la etiqueta, provoca su pausa de revisión humana y **no
  mezcla ningún otro trabajo**. Ninguna tarea posterior puede corregir ese material: por eso T013 entra
  antes, para que nazca ya lintado, y T014 contrasta los guiones construyendo el binario a mano.
  **Tampoco T020**, que es la otra tarea que necesitaría tocar ese material: los escenarios 8 y 9 de
  `quickstart.md` se ejercen sobre una copia desechable del árbol fuera del repositorio, de modo que
  ninguna ruta bajo un directorio `testdata` entra en su diff y su única ruta declarada es el fichero de
  evidencia del directorio del feature.
- **Nunca se graba nada contra la red.** `KITLEGAL_RECORD` no aparece en ninguna tarea: H1 no tiene
  fuentes externas ni respuesta que grabar.
- **`internal/httpx`, `internal/cache`, `internal/store`, `internal/graph`, `internal/source/**` y
  `pkg/legalkit` no se crean.** Están fuera de alcance y crearlos vacíos sería un marcador de posición,
  que SC-012 prohíbe. Las reglas de dependencia que los nombran se activan igualmente: activas y vacías.
- **Los objetivos del `Makefile` que H1 toca son exactamente tres** —`test-e2e`, y `lint`/`fmt`/`fmt-check`
  por la variable de paquetes de ejemplo— y ninguno más: `schema-check`, `skills-sync` y `release` siguen
  anunciando el hito que las aporta, y `build`, `install`, `test`, `test-integration`, `vuln`, `secrets`,
  `mod-verify`, `mod-tidy-check`, `check-tools`, `hooks`, `ci` y `help` no cambian.
- **Nada de atajos** al reparar una verificación en rojo: ni desactivar linters, ni `//nolint` sin
  justificar, ni tests saltados, ni errores silenciados, ni umbrales rebajados para que un estado pase. Si
  un test sigue legítimamente en rojo porque su implementación pertenece a una tarea posterior, la tarea
  está mal delimitada y hay que anotarlo, no arreglarlo por la vía rápida.
