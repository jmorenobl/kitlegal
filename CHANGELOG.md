# Changelog

Todo cambio de comportamiento visible de `kitlegal` se registra aquí.

El formato sigue [Keep a Changelog](https://keepachangelog.com/es-ES/1.1.0/) y el proyecto se adhiere al
[versionado semántico](https://semver.org/lang/es/). Mientras el mayor sea `0` —lo será hasta la primera
release, que es H6 (`v0.1.0`)— un cambio incompatible sube el **menor**. Este fichero se mantiene **a
mano** hasta ese hito: cada propuesta de cambio añade su entrada bajo *Unreleased* en el mismo cambio que
introduce el comportamiento, y al publicar una versión esa sección se cierra bajo su número y su fecha y
se abre una nueva vacía.

## [Unreleased]

Todavía no hay ninguna versión publicada. Lo que sigue es lo que aportan los hitos cerrados hasta hoy, en
orden de llegada: **H0 — esqueleto del repositorio y sus controles**, un repositorio sin fuentes legales
todavía, pero blindado, para que cualquier línea de Go que entre después atraviese los mismos gates; y
**H1 — kernel de la línea de órdenes**, lo que un applet **no** tiene que declarar, de modo que cada
fuente legal de los hitos siguientes herede la misma forma de invocarse, de fallar y de citar sin volver
a escribirla. **H3 — caché local en SQLite** es un hito de fundación que no cambia nada en el binario,
así que su única entrada, en *Cambiado*, es lo que cambia en `make ci`.

### Añadido

*De H0 — el repositorio y sus controles:*

- **Binario `kitlegal`** con un único verbo, `version`, que imprime en tres líneas de la salida estándar
  la versión, el commit y la fecha de construcción, deja vacía la salida de error y termina con código
  `0`. La versión sale de `git describe --tags --always --dirty` y el commit coincide carácter a carácter
  con la revisión construida. Cualquier otra invocación —sin verbo, con un verbo desconocido o con un
  argumento sobrante— escribe una línea de uso en la salida de error y termina con código `2` («args» en
  la tabla de códigos de salida estables del proyecto).
- **`Makefile` como única superficie de invocación de los controles**, con `make ci` como veredicto del
  repositorio: la misma orden que se ejecuta en local es la que ejecutan el gancho de pre-commit y la
  integración continua, y ningún control se aplica por otra vía. `make ci` no modifica ningún fichero
  versionado. `make build` y `make install` compilan sin cgo, con `-trimpath` e inyectando los datos de
  construcción; `make help` es el objetivo por defecto.
- **Ocho controles activos dentro de `make ci`**: formato en modo verificación (`gofumpt` y `goimports`),
  análisis estático (`golangci-lint`, con `gosec` y `govet` incluidos), tests unitarios con detector de
  carreras y perfil de cobertura, vulnerabilidades conocidas (`govulncheck`), validación contra esquemas,
  detección de secretos (`gitleaks`), integridad de los módulos (`go mod verify`, raíz y herramientas) y
  dependencias saneadas (`go mod tidy -diff`). `make check-tools` comprueba los prerrequisitos como
  dependencia de las demás órdenes.
- **Cadena de herramientas reproducible y sin instalación previa**: los cuatro controles con binario
  propio —`golangci-lint`, `govulncheck`, `gitleaks` y `lefthook`— se construyen solos desde la versión
  fijada en `tools/<herramienta>/go.mod`, y `go.mod` declara la directiva `toolchain` (`go1.27.1`, la
  estable actual al cerrar el hito) que el `Makefile` exporta como `GOTOOLCHAIN`, de modo que todas las
  órdenes usan el mismo parche de Go que la integración continua. Los dos únicos prerrequisitos son una
  cadena Go 1.21 o superior y `git`.
- **Ganchos de pre-commit** (`make hooks`, con `lefthook`): cada commit corrige el formato y vuelve a
  preparar lo corregido, y ejecuta `lint-fast`, `secrets` y `mod-tidy-check`. Es un subconjunto rápido, no
  el veredicto: la autoridad final sigue siendo la integración continua.
- **Umbrales de cobertura bloqueantes** (`codecov.yml`): ≥ 70 % global y ≥ 85 % en el dominio interno,
  ambos declarados como estado que falla y no como información.
- **Flujos de la plataforma**: `ci` en cada propuesta de cambio y en cada push a la rama principal,
  `nightly` a diario sobre la rama principal —ambos se limitan a preparar el entorno y ejecutar `make
  ci`—, `codeql` semanal como segundo análisis de seguridad y Dependabot semanal sobre el módulo raíz, los
  cuatro módulos de herramienta y las acciones de los flujos.
- **Documentación fundacional**: `LICENSE` (Apache-2.0), `README.md`, `CONTRIBUTING.md` —ritual por hito,
  estructura de propuesta de cambio, Conventional Commits, versionado semántico, catálogo de controles y
  justificación obligatoria de toda dependencia nueva—, este `CHANGELOG.md` y los cuatro ADR fundacionales
  en `docs/ADR/`: multicall, SQLite sin cgo, CENDOJ no masivo y frontera humana.

*De H1 — el kernel de la línea de órdenes:*

- **Despacho multicall**: un solo ejecutable atiende a todos los applets. Manda el **nombre con el que se
  invoca**, de modo que un enlace simbólico `echo -> kitlegal` ejecuta el applet `echo` y le entrega
  íntegros los argumentos —`./echo boe` ejecuta `echo` con el argumento `boe`, y no el applet `boe`—; solo
  cuando ese nombre no está registrado —`kitlegal` entre ellos— el applet sale del **primer argumento**.
  Los verbos reservados del binario se reconocen antes que el registro, y por eso el registro rechaza al
  construirse un applet que se llame como uno de ellos. Una invocación que no resuelve ningún applet
  termina con código `2` nombrando lo que no reconoció y enumerando lo que existe; nunca con un pánico.
- **El registro es la única lista**: registrar un applet basta para que aparezca en `--help` y declarar un
  verbo basta para que se enumere, sin ninguna lista paralela mantenida a mano. Si el applet declara un
  verbo por omisión, invocarlo sin nombrar verbo lo usa; si no lo declara, nombrarlo es obligatorio y
  omitirlo termina con código `2`. La ayuda del binario, la de un applet y la de un verbo son texto para
  una persona y `--json` no las altera.
- **Ocho banderas globales idénticas en todos los applets**, declaradas una sola vez en el kernel y
  heredadas sin escribir ninguna: `--json`, `--timeout` (30 s por omisión, plazo de **toda** la operación
  y no de una petición suelta), `--offline`, `--dry-run`, `--describe`, `--no-graph`, `--asunto` y
  `--verbose`. Un applet declara su nombre, sus verbos y el contenido de `data`, y las recibe ya
  interpretadas. Tres —`--offline`, `--no-graph` y `--asunto`— fijan hoy solo su sintaxis y se propagan
  tal cual: su semántica llega con la caché (H3), con el grafo (H17) y con el asunto (H18), y no se
  inventa antes.
- **Códigos de salida estables**: `0` correcto, `2` argumentos inválidos, `3` no encontrado, `4` fuente no
  disponible, `5` límite de peticiones o términos de uso, `6` requiere identidad humana, y **`1` reservado
  al fallo inesperado** —lo que nadie declaró—, que es la convención de Unix para el error general y el
  único valor que un consumidor interpreta sin documentación. Un applet nombra la clase de su fallo con un
  sentinela y **nunca** un número; la traducción de clase a código ocurre en un único punto por binario,
  cubre las seis clases sin rama por defecto y el linter falla si alguien añade una clase y se olvida de
  darle código. Ninguna ruta de usuario termina en pánico.
- **Sobre de salida con huella reproducible**: toda invocación emite las mismas **seis** claves —`ok`,
  `fuente`, `url`, `fecha_consulta`, `hash` y `data`—, en éxito y en fallo, sin omitir ninguna aunque su
  valor sea el cero. `fecha_consulta` va en RFC 3339 con desplazamiento horario explícito; `hash` lleva
  delante el algoritmo que lo produjo —`sha256:`— seguido de los 64 dígitos hexadecimales del SHA-256 de
  la forma canónica de `data`, de modo que el mismo contenido dé siempre la misma huella con
  independencia del orden de las claves y del instante de la consulta, y que un solo byte distinto dé
  otra. Sin `fuente` y sin una `url` absoluta no se emite sobre, porque sin ellas no hay cita. Cuando `ok`
  es falso, `data` lleva exactamente `clase` y `mensaje`: el detalle técnico va al registro de eventos y
  no al sobre.
- **Dos formas de presentación y una sola salida estándar**: con `--json`, un único documento en una línea
  y nada más, ni siquiera con el registro de eventos al máximo detalle; sin `--json`, la **tabla mínima**,
  que escribe las cuatro líneas de procedencia —`fuente`, `url`, `fecha_consulta` y `hash`— y a
  continuación el contenido de `data` aplanado a pares ruta/valor, de modo que elegir la forma legible por
  una persona no pierda la cita. El registro de eventos va **siempre** a la salida de error, y su nivel se
  fija con `--verbose` o con la variable de entorno `KITLEGAL_LOG` (`debug`, `info`, `warn`, `error`); un
  valor fuera de esa lista se ignora, se avisa y no aborta la invocación.
- **Autodescripción**: `--describe` emite un esquema JSON con la entrada y la salida del verbo —las dos en
  un mismo documento, bajo `entrada` y `salida`— y excluye la ejecución, así que el applet nunca llega a
  ver esa invocación. Es la contraparte legible por máquina de `--help`, y de ella saldrán las tools MCP y
  la tabla de órdenes de cada `SKILL.md`.
- **Cinco dependencias directas nuevas**, todas de la lista cerrada de la constitución (§V): en el
  binario, `github.com/alecthomas/kong` (análisis de la línea de órdenes) e `github.com/invopop/jsonschema`
  (generación del esquema de `--describe`); solo en los tests, `github.com/stretchr/testify` (aserciones),
  `github.com/rogpeppe/go-internal` (los guiones `testscript` del e2e) y
  `github.com/santhosh-tekuri/jsonschema/v6` (validación del sobre contra su descripción formal). Con
  ellas aparece por primera vez un `go.sum` en la raíz del módulo. **El binario distribuido enlaza además
  los cuatro módulos que `invopop/jsonschema` arrastra** y que ninguna versión de esa biblioteca deja de
  traer: `github.com/pb33f/ordered-map/v2` (las propiedades del esquema conservan el orden de
  declaración), `github.com/bahlo/generic-list-go` y `github.com/buger/jsonparser` (dependencias de ese
  mapa) y `go.yaml.in/yaml/v4` (en versión candidata, `v4.0.0-rc.2`, fijada por aquel mapa; Dependabot la
  sigue). Están justificados en el plan del hito (*Complexity Tracking*) y en la propuesta de cambio, y
  un test de arquitectura (`TestDependenciasDelBinario`) fija la lista exacta de módulos que el binario
  enlaza, de modo que uno nuevo no entra sin justificarse por escrito. Los módulos que solo usan los
  tests —`go.yaml.in/yaml/v3` por `testify`, `golang.org/x/sys` y `golang.org/x/tools` por `testscript`,
  `golang.org/x/text` por el validador— no se enlazan en el binario.
- **La forma corta `-h` pide la ayuda en las mismas posiciones que `--help`**: en el binario, en un
  applet y en un verbo. La ayuda del verbo, que escribe el analizador, la anunciaba ya como `-h, --help`;
  ahora el kernel la reconoce también donde todavía no hay gramática que la lea.

### Cambiado

*De H1 — el kernel de la línea de órdenes:*

- **El contrato observable del binario deja de ser el de H0.** De aquel se conservan las tres líneas de
  `version` y su código `0` —ahora escritas por el presentador, y un fallo al escribirlas se propaga en
  lugar de descartarse— y el código `2` de cualquier otra invocación; lo que cambia es el mensaje y lo
  que hay detrás. `version` sigue sin admitir argumentos ni banderas: `kitlegal version extra` o
  `kitlegal version --json` terminan con `2`, como en H0, pero el mensaje nombra lo que sobra en lugar de
  la línea `uso: kitlegal version`. Un applet desconocido lo resuelve el despacho multicall, que nombra lo
  que no reconoció y enumera lo registrado. `--help` responde la ayuda derivada del registro y termina
  con `0`, donde H0 terminaba con `2` por no ser `version`; y los códigos de salida posibles ya no son
  solo `0` y `2`, sino los siete de la tabla estable.
- **Un descriptor de salida roto termina siempre con `1`**, sea lo que sea lo que se estaba escribiendo:
  el sobre, la tabla, el esquema de `--describe`, `version` o cualquiera de las tres ayudas —también la
  del verbo, que escribe el analizador y que sin vigilar el escritor habría salido con el `2` de
  argumentos inválidos—. Vale también para una tubería del sistema cuyo lector ha terminado
  (`kitlegal … | head -1`): el binario ignora la señal `SIGPIPE` con la que el runtime de Go mataría el
  proceso sin código ni mensaje, de modo que la escritura fallida se traduce como cualquier otro fallo,
  con el mensaje en la salida de error y el código `1`, y nunca con la muerte por señal (`141` en el
  intérprete de órdenes).
- **La exclusión de `errcheck` desaparece.** H0 eximía `fmt.Fprint`, `fmt.Fprintf` y `fmt.Fprintln` porque
  el punto de entrada escribía él mismo y su contrato de códigos de salida no tenía dónde poner un fallo
  de escritura. Ya no escribe —inyecta los descriptores del sistema, el registro de producción y los datos
  de construcción, y termina con el código que le devuelve la raíz de composición—, así que ninguna
  función queda exenta y todo error de escritura se comprueba.
- **`make test-e2e` deja de anunciar el hito ausente** y ejecuta los guiones `testscript` que describen la
  entrega, contra el binario que el propio test construye.

*De H3 — la caché local en SQLite:*

- **`make ci` ejecuta también los tests de integración.** La lista de prerrequisitos de `ci` gana
  `test-integration` —`go test -race -tags=integration ./...`, la receta que H0 fijó y que no cambia—
  entre `test` y `vuln`, de modo que un test etiquetado `integration` en rojo, o un fichero etiquetado que
  no compile, hacen fallar el veredicto en local y en la integración continua por igual. Los primeros
  tests con esa etiqueta son los de la caché: los que dependen del entorno —permisos del sistema de
  ficheros y dos procesos— y trabajan solo dentro de directorios temporales. El análisis estático alcanza
  también los ficheros etiquetados (`run.build-tags: [integration]` en `.golangci.yml`), de modo que
  `sqlclosecheck` y `rowserrcheck` los vigilan. Con ello `make ci` encadena nueve controles: los ocho de
  H0 y este. Nada más cambia a la vista: ningún applet, verbo ni bandera nueva, y el binario distribuido
  no enlaza todavía la caché ni el controlador de SQLite.

Tres órdenes existen ya pero reciben su contenido en un hito posterior y ninguna miente sobre ello:
`schema-check` (H4 y H10), `skills-sync` (H5) y `release`, que falla con código distinto de `0` hasta H6
por ser la única con efectos externos. El binario que se publica **no registra
todavía ningún applet** y su ayuda lo dice en lugar de enumerar una lista vacía: los de fuentes (`boe`,
`placsp`, `bdns`…) llegan en los hitos siguientes, en el orden de `docs/ROADMAP.md`, y el primero es el de
H4.
