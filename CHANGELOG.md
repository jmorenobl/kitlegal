# Changelog

Todo cambio de comportamiento visible de `kitlegal` se registra aquí.

El formato sigue [Keep a Changelog](https://keepachangelog.com/es-ES/1.1.0/) y el proyecto se adhiere al
[versionado semántico](https://semver.org/lang/es/). Mientras el mayor sea `0` —lo será hasta la primera
release, que es H19 (`v0.1.0`, ADR 0013)— un cambio incompatible sube el **menor**. Este fichero se mantiene **a
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
así que su única entrada, en *Cambiado*, es lo que cambia en `make ci`. **H4 — applet `boe`** trae la
primera fuente legal: la legislación consolidada del BOE, consultable y citable desde el binario
distribuido, con caché, contratos de salida publicados y una verificación nocturna contra la fuente real.
**H5 — skill `boe-legislacion`** trae la primera skill del producto, que consulta y cita cualquier norma
consolidada del BOE con ese binario, y el andamiaje que comparten todas las skills: una tabla de normas como
única fuente de verdad, lo generado comprobado en `make ci`, la instalación con `make install` y el formato
común de eval con su job de evals. El binario distribuido no cambia.

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

*De H4 — el applet `boe`:*

- **Applet `boe`, el primero con fuente**, registrado en el binario distribuido: consulta la API de
  Legislación Consolidada del BOE (`https://www.boe.es/datosabiertos/api/legislacion-consolidada`) y firma
  cada sobre con `fuente` `boe.legislacion-consolidada` y la `url` de la API consultada. `kitlegal boe …` y
  el enlace `boe -> kitlegal` dan la misma salida byte a byte. Tiene **seis verbos**, ninguno por omisión
  —sin verbo termina con `2`— y ninguna bandera propia:
  - `buscar <texto>…` une las palabras y busca por título (`titulo:<p1> AND titulo:<p2> …`), o pasa la
    consulta tal cual si lleva operadores (` AND `, ` OR `, ` NOT `, `titulo:`, `materia:` o comillas); como
    mucho diez resultados, en el orden de la fuente, y `[]` si no hay ninguno.
  - `indice <norma>` devuelve los bloques de la norma en el orden de la fuente, con el tipo de cada uno.
  - `articulo <norma> <bloque>` devuelve el texto vigente del bloque, su `hash_texto`, su dirección pública
    y la `url_eli` de la norma, con los avisos de su vigencia (`codigo`, de un enumerado cerrado de tres
    valores, y `texto`), y nunca emite un texto cuya vigencia no ha podido comprobar. Es el porte de
    `refs/boe.py`: su `data` coincide con lo que calcula `boe.py` en los ocho campos del diff de aceptación,
    para seis artículos de cuatro normas.
  - `articulos <norma> <bloque>…` hace lo mismo con varios bloques, en el orden pedido y con repeticiones,
    pidiendo cada bloque distinto una sola vez y los metadatos como mucho una vez por invocación.
  - `metadatos <norma>` devuelve los datos de la norma y los avisos de su vigencia.
  - `analisis <norma>` devuelve sus materias, sus notas y sus referencias anteriores y posteriores, con el
    texto completo.

  Los identificadores se validan antes de abrir nada (`BOE-A-AAAA-N` para la norma; letras, dígitos, `.` y
  `-` para el bloque) y los fallos salen con los códigos estables: `2` argumentos inválidos; `3` lo que la
  fuente no tiene —un bloque o una norma inexistentes, o una respuesta vacía—; `4` la fuente caída tras los
  reintentos, un estado HTTP no previsto o una respuesta que ya no se sabe interpretar; y `5` el límite de
  peticiones o un `robots.txt` que deniega la ruta. Ninguno termina con `6`. El sobre de un fallo de la
  fuente lleva la `url` de la petición que falló.
- **Caché de las consultas y semántica de `--offline` y `--dry-run`.** `boe` guarda cada respuesta correcta
  en la caché local de H3 (`~/.cache/kitlegal/cache.db`, u otra carpeta con `KITLEGAL_CACHE_DIR`) con una
  vigencia por verbo —300 s para `buscar` y `metadatos`, 7 días para `indice`, `articulo`, `articulos` y
  `analisis`—; los metadatos con los que se comprueba la vigencia se comparten entre `articulo`, `articulos`
  y `metadatos`, y **ningún fallo se guarda**. Una consulta idéntica dentro de la vigencia no emite ninguna
  petición. `--offline` abre la caché en solo lectura y responde con lo vigente, o termina con `4` sin pedir
  nada; `--dry-run` tampoco pide nada ni crea la caché, y describe en la salida de error, una línea por
  petición y sin repetir ninguna, cada `GET` que habría emitido. `--no-graph` y `--asunto` siguen sin
  cambiar la salida.
- **Fila de la fuente en `docs/SOURCES.md`**, que nace con ella: licencia, términos de uso, resultado de la
  revisión del `robots.txt`, ritmo (`1s` entre dos peticiones al mismo sitio), formato y fecha de la revisión
  humana. El adaptador pide con ese ritmo y declara esos términos, y un test falla si la fila y sus
  constantes divergen.
- **Esquemas publicados `schemas/norma.json` y `schemas/bloque.json`**, los primeros contratos de salida
  versionados: `norma.json` describe `buscar`, `indice`, `metadatos` y `analisis`, y `bloque.json`,
  `articulo` y `articulos`, cada verbo bajo `$defs.<verbo>` con su `$id`
  (`https://ventanillalegal.es/schemas/<fichero>/<verbo>`) y la entrada y la salida que emite su
  `--describe`. Los tests validan contra el fichero —no contra lo que emite el binario en ese instante— el
  sobre real de éxito y los de fallo con códigos `2`, `3` y `4` de los seis verbos.
- **`make verify-sources` y la verificación nocturna contra la fuente real.** `make verify-sources` ejecuta
  `scripts/verify-sources.sh`, que pide a la API del BOE `boe articulo BOE-A-2015-10565 a21 --json` y exige
  código `0`, un sobre válido contra la parte `articulo` de `schemas/bloque.json` y un texto no vacío: forma
  y no contenido, porque el texto de un artículo cambia legítimamente con cada reforma. Es el único control
  que pide algo a una fuente real, así que **necesita red y no forma parte de `make ci`**; la misma
  comprobación, sin red, la ejerce `make test` sobre las grabaciones y sobre una respuesta que ya no se
  interpreta. El flujo `nightly` gana el trabajo `fuentes`, independiente del de `make ci` y con permiso para
  escribir incidencias: ejecuta `make verify-sources` y, si falla, comenta la incidencia abierta
  «verify-sources: boe articulo» o la abre. Ningún flujo graba respuestas: las grabaciones contra las que
  corren los tests las hace una persona con `scripts/grabar-fixtures.sh`.

*De H5 — la skill `boe-legislacion` y el andamiaje de skills:*

- **Skill `boe-legislacion`** (`skills/boe-legislacion/`), la primera del producto y genérica para cualquier
  materia: consulta y cita la normativa consolidada del BOE —procedimiento administrativo, contratación pública,
  régimen local, tributos, relaciones laborales…— con el applet `boe`. Su `SKILL.md`, de menos de 300 líneas, fija
  un protocolo de cinco pasos (identificar la norma en su referencia de normas, resolver `BOE-A-…` o buscarlo con
  `boe buscar`, leer el índice y los bloques, evaluar si falta contexto y responder citando), la forma de cita
  `[BOE-A-2015-10565, bloque a21]`, sin nada más dentro de los corchetes, y sus reglas: ningún contenido legal que
  no salga del texto consultado, ley y reglamento distinguidos y la variación autonómica señalada. El id de cada
  bloque se copia de la entrada del índice, nunca se compone del número del artículo, y los bloques se leen de uno
  en uno; si una consulta de varios bloques falla, cada bloque se pide por separado antes de decir qué no se pudo
  consultar. Llama al binario por `scripts/boe`, un enlace al
  binario instalado, y no tiene ningún caso especial de un municipio ni de una comunidad.
- **Tabla de normas `data/normas.yaml`**, única fuente de verdad de las normas que referencian las skills: diez
  normas, cada una con su identificador `BOE-A-…`, su título, su rango, sus materias y, si la tiene, su
  abreviatura. De ella se genera `references/normas.md` de la skill, con la cabecera que prohíbe editarlo.
- **Esquemas de datos `schemas/normas.yaml.json` y `schemas/eval.yaml.json`** (JSON Schema 2020-12), contra los
  que se validan la tabla de normas y cada eval; no son contratos de salida de ningún applet. Los lee un lector
  YAML común que rechaza una clave escrita dos veces en el mismo mapa nombrando sus dos líneas, en lugar de
  quedarse en silencio con el último valor. Para ello entra `go.yaml.in/yaml/v3` como dependencia directa, que
  solo usan las herramientas de desarrollo y que el binario no enlaza.
- **Formato común de eval**: las evals de cada skill en su propio directorio, `evals/<skill>/`, con un fichero
  YAML por eval (`<nn>-<descripción>.yaml`) y los campos `pregunta`, `activa`, `comandos` y `citas` —estos dos,
  obligatorios en una eval que debe activar la skill y prohibidos en una que no— y el opcional `reproduce`.
- **Evals de `boe-legislacion`** (`evals/boe-legislacion/`), doce, escritas en el formato común de eval y antes que
  la skill: diez preguntas de materias distintas que deben activarla, hacer las consultas esperadas y citar el
  bloque esperado —la del IRPF reproduce un uso documentado de `boe-fiscal` y lo declara con
  `reproduce: boe-fiscal`— y dos preguntas ajenas que no deben activarla. Las respuestas del BOE que necesitan las
  graba una persona con `scripts/grabar-evals.sh`.
- **Job de evals**: el flujo `evals` (`.github/workflows/evals.yml`) ejecuta `make evals SKILL=boe-legislacion` a
  mano, cada semana sobre la rama principal y al poner la etiqueta `evals` o `evals-prueba-de-red` en la propuesta
  de cambio de un hito, con un único modelo fijado en su definición (`claude-haiku-4-5-20251001`), Claude Code
  `2.1.270`, el secreto `CLAUDE_CODE_OAUTH_TOKEN` y un runner del que antes retira todo Python. El informe, con el
  veredicto global y lo que se comprobó de cada eval, se imprime en el registro entre marcas.
- **`make skills-sync` real**: deja de anunciar que las skills llegan en H5 y regenera, desde `data/*.yaml` y desde
  `--describe` del binario, `references/`, la tabla de comandos de `SKILL.md` y los enlaces de `scripts/` de cada
  skill; dos ejecuciones seguidas no cambian nada.
- **`make skills-check`**, dentro de `make ci`: regenera todo eso en memoria y lo compara con el árbol sin escribir
  nada, y comprueba el frontmatter y el límite de líneas de cada `SKILL.md`, la tabla de normas contra su esquema y
  cada identificador contra la búsqueda grabada del BOE, y el formato y el conjunto de las evals y que lo que
  necesitan está grabado. Falla nombrando la skill y el fichero o el enlace, sin red y sin modelo.
- **`make evals`** (`scripts/evals.sh <skill>`), fuera de `make ci`: una sesión de Claude Code por eval, bajo
  `strace` y con la red cerrada salvo la del modelo, con la skill instalada por `make install` y el binario
  respondiendo desde una caché preparada con lo grabado; juzga cada sesión sin modelo y escribe el informe. Necesita
  Linux con `strace`, root o `sudo` y ningún Python accesible, y falla antes de la primera sesión si falta algo o si
  una eval está mal formada.

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
  `test-integration` —`go test -race -tags=integration -coverprofile=coverage-integration.out ./...`, la receta de H0 más el perfil de cobertura—
  entre `test` y `vuln`, de modo que un test etiquetado `integration` en rojo, o un fichero etiquetado que
  no compile, hacen fallar el veredicto en local y en la integración continua por igual. Los primeros
  tests con esa etiqueta son los de la caché: los que dependen del entorno —permisos del sistema de
  ficheros y dos procesos— y trabajan solo dentro de directorios temporales. El análisis estático alcanza
  también los ficheros etiquetados (`run.build-tags: [integration]` en `.golangci.yml`), de modo que
  `sqlclosecheck` y `rowserrcheck` los vigilan. Con ello `make ci` encadena nueve controles: los ocho de
  H0 y este. Nada más cambia a la vista: ningún applet, verbo ni bandera nueva, y el binario distribuido
  no enlaza todavía la caché ni el controlador de SQLite.
- **La cobertura que publica CI incluye los tests de integración.** `ci.yml` sube los dos perfiles
  (`coverage.out` y `coverage-integration.out`) y Codecov los une, de modo que las ramas que solo
  ejercitan los tests etiquetados —permisos del sistema de ficheros, dos procesos— cuentan como lo que
  son: código con test. `codecov.yml` declara además el estado `patch` (`target: auto`, bloqueante), que
  hasta ahora regía sin declarar: la regla «no retroceder» aplicada al código nuevo de cada propuesta.

*De H4 — el applet `boe`:*

- **`make schema-check` deja de ser un aviso.** Ejecuta `TestEsquemasPublicados`
  (`go test -count=1 -run '^TestEsquemasPublicados$' ./internal/app/`), que regenera en memoria, desde
  `--describe` de los verbos registrados en el binario distribuido, la forma canónica de
  `schemas/norma.json` y de `schemas/bloque.json` y la compara con los ficheros versionados, sin escribir
  nada; falla nombrando el fichero y el verbo que difieren, o el fichero que no es la serialización canónica
  de sus partes. Sigue dentro de `make ci`, que encadena los mismos nueve controles. Los ficheros solo se
  regeneran a propósito, con la bandera `-actualizar-esquemas` del mismo test.
- **`fecha_consulta` la declara la fuente, también en lo servido desde la caché** (ADR 0015). Hasta H3 el
  kernel fechaba todo sobre con su reloj al montarlo, lo que habría hecho pasar por recién consultado un
  artículo guardado hace seis días. `schema.Procedencia` gana `FechaConsulta`, y el montador fecha el sobre
  con ella cuando la procedencia es válida y la fecha no es cero, en éxito y en fallo; con el valor cero, o
  con una procedencia inválida, lo fecha con su reloj como antes, así que un applet que no declara fecha no
  cambia. `boe` declara el instante en que `internal/httpx` emitió la petición —el de su último intento— o,
  si lo que responde sale de la caché, el instante guardado en la entrada; si `data` se sostiene sobre varias
  consultas, el más antiguo, y en un fallo, el de la petición que falló. La huella no cambia: se sigue
  calculando sobre `data`, que no lleva la fecha.
- **`app.Arrancar` es la raíz de arranque de los dos binarios.** Construye el registro y, si se construye,
  atiende la invocación con `Main`; un registro que no se construye es un defecto de composición y sale como
  sobre del kernel de clase «inesperado», con el mensaje en la salida de error y código `1`, nunca como un
  pánico. El binario distribuido termina con `os.Exit(app.Arrancar(os.Args, app.RegistroDeProduccion, …))`:
  su ayuda enumera `boe` y el fallo de un applet desconocido termina en `applets disponibles: boe`, en lugar
  de decir que el binario no registra ninguno. El binario de extremo a extremo deja su `panic` y arranca por
  el mismo camino con `echo`, `contar` y `boe` —este, sobre la reproducción de sus grabaciones y sin red—, y
  `make test-e2e` recorre con él los seis verbos, el enlace, `--offline`, los códigos y la respuesta servida
  desde la caché por debajo de 200 ms.
- **El binario distribuido enlaza la red y la caché.** `go.mod` no gana ninguna entrada, pero al registrar
  `boe` el binario enlaza `internal/httpx` e `internal/cache` y, con ellos, doce módulos de terceros:
  `github.com/temoto/robotstxt` y `golang.org/x/time` —de la lista cerrada de la constitución, por la red—,
  `modernc.org/sqlite` —ídem, por la caché— y los nueve que ese controlador arrastra (`modernc.org/libc`,
  `modernc.org/mathutil`, `modernc.org/memory`, `github.com/remyoudompheng/bigfft`,
  `github.com/dustin/go-humanize`, `github.com/google/uuid`, `github.com/mattn/go-isatty`,
  `github.com/ncruces/go-strftime` y `golang.org/x/sys`). `TestDependenciasDelBinario` fija la lista de
  dieciocho, cada uno justificado en su línea, y se retiran los tests que exigían que el binario no enlazara
  la red ni la caché, ciertos solo mientras ningún applet las usaba.

*De H5 — la skill `boe-legislacion` y el andamiaje de skills:*

- **`make install` enlaza las skills.** Tras el `go install` de H0, sin cambios, `scripts/instalar-skills.sh` enlaza
  cada skill de `skills/` en `~/.claude/skills/`, el directorio personal de skills de Claude Code, y deja
  `bin/instalado/kitlegal` apuntando al binario instalado, que es lo que alcanza `scripts/boe` de la skill. Repetirla
  deja el mismo estado. Ante una entrada del directorio personal con el nombre de una skill que no es su enlace —un
  directorio, un fichero u otro enlace— escribe una línea de conflicto por cada una y falla sin crear ni cambiar
  nada. `make test-integration` lo prueba (`TestInstalacion`) sobre una copia mínima del árbol, con el directorio
  personal, el de binarios y el `GOPATH` temporales y sin red.
- **`make ci` encadena diez controles**: `skills-check` entra tras `schema-check`, de modo que una skill con una
  deriva o un defecto, una tabla de normas inválida o una eval mal formada hacen fallar el veredicto en local y en la
  integración continua por igual. El análisis estático alcanza también los ficheros con la etiqueta de compilación
  `evals` (`run.build-tags` de `.golangci.yml`), los arneses que usa el job de evals.

Una orden existe ya pero recibe su contenido en un hito posterior y no miente sobre ello: `release`, que falla
con código distinto de `0` hasta H6 porque es una acción con efectos externos. El binario que se publica registra **un solo applet, `boe`**: los de las demás fuentes (`placsp`,
`bdns`…) llegan en los hitos siguientes, en el orden de `docs/ROADMAP.md`.
