# Changelog

Todo cambio de comportamiento visible de `kitlegal` se registra aquí.

El formato sigue [Keep a Changelog](https://keepachangelog.com/es-ES/1.1.0/) y el proyecto se adhiere al
[versionado semántico](https://semver.org/lang/es/). Mientras el mayor sea `0` —la primera release, `v0.1.0`,
es la de H19 (ADR 0019)— un cambio incompatible sube el **menor**. Este fichero se mantiene **a mano**, también
desde esa release: cada propuesta de cambio añade su entrada bajo *Unreleased* en el mismo cambio que
introduce el comportamiento, y al publicar una versión esa sección se cierra bajo su número y su fecha y
se abre una nueva vacía. Las notas de cada release las genera goreleaser desde los Conventional Commits, y no
sustituyen a este fichero.

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
común de eval con su job de evals. El binario distribuido no cambia. **H6 — applet `territorio` y skill
`legal-core`** trae el segundo applet, el primero sin fuente que consultar en red: resuelve cualquier municipio de
España a su territorio desde datos congelados que viajan dentro del binario, y declara lo que no está configurado en
lugar de inventarlo; y la skill madre `legal-core`, que empieza toda pregunta por el territorio. **H19 — instalar
sin clonar** trae la distribución (ADR 0019): el binario se instala con el gestor de paquetes de cada plataforma o con
`install.sh`, lleva las skills dentro y las instala con el applet `skills`, las skills invocan `kitlegal` desde el
`PATH` y `make install` pasa a ser el bucle de desarrollo.

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
  `art. 21 de la Ley 39/2015 [BOE-A-2015-10565, bloque a21]`, en la que lo que hace cita es que los corchetes terminen
  en el identificador y el bloque —con la forma legible delante del corchete o, dentro, delante del identificador—,
  también cuando la cita va sola, y sus reglas: ningún contenido legal que
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
  obligatorios en una eval que debe activar la skill y prohibidos en una que no— y los opcionales `reproduce` e
  `informativa`, que marca la eval que se mide y se publica sin decidir el veredicto.
- **Evals de `boe-legislacion`** (`evals/boe-legislacion/`), diecisiete, escritas en el formato común de eval y antes
  que la skill: diez preguntas de materias distintas que deben activarla, hacer las consultas esperadas y citar el
  bloque esperado —la del IRPF reproduce un uso documentado de `boe-fiscal` y lo declara con
  `reproduce: boe-fiscal`—, dos preguntas ajenas que no deben activarla y cinco preguntas por materia, sin el número
  del artículo, marcadas `informativa: true`: se ejecutan y su tasa se publica, pero no deciden el veredicto mientras
  siga en el backlog la herramienta que busca dentro de una norma el artículo que trata una materia (ADR 0016). Las
  respuestas del BOE que necesitan las graba una persona con `scripts/grabar-evals.sh`.
- **Job de evals**: el flujo `evals` (`.github/workflows/evals.yml`) ejecuta `make evals SKILL=boe-legislacion` a
  mano, al abrirse o reabrirse una propuesta de cambio que toque la skill, sus datos, sus evals, el applet `boe`, el
  kernel, el arnés de las evals, el `Makefile` o el propio job, y al poner la etiqueta `evals` o `evals-prueba-de-red`
  en cualquiera —sin ejecución programada y sin relanzarse al empujar a una propuesta abierta—, con Claude Code
  `2.1.270`, el secreto `CLAUDE_CODE_OAUTH_TOKEN` y un runner del que antes retira todo Python. Decide con
  `claude-sonnet-5`, el modelo del uso real de la skill, y ejecuta además `claude-haiku-4-5-20251001` como límite
  inferior que se publica sin decidir; cada eval se abre tres veces con `claude-sonnet-5` y, si no es informativa,
  otras tres con `claude-haiku-4-5-20251001`, y cada una de esas series pasa con dos (ADR 0016). El
  informe, con el veredicto global, la tasa de cada eval y lo que se comprobó de cada sesión, se imprime en el registro
  entre marcas.
- **`make skills-sync` real**: deja de anunciar que las skills llegan en H5 y regenera, desde `data/*.yaml` y desde
  `--describe` del binario, `references/`, la tabla de comandos de `SKILL.md` y los enlaces de `scripts/` de cada
  skill; dos ejecuciones seguidas no cambian nada.
- **`make skills-check`**, dentro de `make ci`: regenera todo eso en memoria y lo compara con el árbol sin escribir
  nada, y comprueba el frontmatter y el límite de líneas de cada `SKILL.md`, la tabla de normas contra su esquema y
  cada identificador contra la búsqueda grabada del BOE, y el formato y el conjunto de las evals y que lo que
  necesitan está grabado. Falla nombrando la skill y el fichero o el enlace, sin red y sin modelo.
- **`make evals`** (`scripts/evals.sh <skill>`), fuera de `make ci`: una sesión de Claude Code por cada eval, cada
  modelo y cada repetición que pide el plan, bajo `strace` y con la red cerrada salvo la del modelo, con la skill
  instalada por `make install` y el binario respondiendo desde una caché preparada con lo grabado; juzga cada sesión sin
  modelo, agrupa las de cada eval y cada modelo en una serie con su tasa y escribe el informe. Necesita Linux con
  `strace`, root o `sudo` y ningún Python accesible, y falla antes de la primera sesión si falta algo o si una eval está
  mal formada.

*De H5.1 — los avisos de vigencia en las evals:*

- **Forma fija de los avisos de vigencia en `boe-legislacion`**: su `SKILL.md` fija cómo traslada la respuesta cada
  aviso del sobre, con `⚠`, la etiqueta del aviso tal como la da el binario y dos puntos, seguidos de la frase del
  binario o de una explicación —`⚠ NORMA DEROGADA:` para `derogada`, `⚠ VIGENCIA AGOTADA:` para `vigencia-agotada` y
  `⚠ TEXTO POSIBLEMENTE DESACTUALIZADO:` para `consolidacion-no-finalizada`—, con la etiqueta entera y en la misma
  línea; decir con otras palabras que la norma está derogada no traslada el aviso. La etiqueta de cada código tiene una
  sola fuente de verdad, el applet `boe`, cuya salida no cambia en un byte, y `make skills-check` falla nombrando el
  código si el `SKILL.md` pierde la forma fija de alguno.
- **Campo `avisos` en el formato común de eval**: opcional y solo en una eval que activa la skill, lista los códigos
  de aviso de vigencia del binario —`consolidacion-no-finalizada`, `derogada` o `vigencia-agotada`— cuya forma fija
  tiene que llevar la respuesta. `schemas/eval.yaml.json` rechaza un código desconocido, una lista vacía y `avisos` en
  una eval de no activación, y `make skills-check` comprueba que el esquema admite exactamente los códigos del binario.
  Una eval con `avisos` solo pasa si la respuesta lleva la forma fija de cada uno: se juzga sin modelo, con el selector
  de variante tras `⚠`, el énfasis de Markdown, otros blancos y las minúsculas tolerados, y otra redacción o la negación
  dejan el aviso ausente. Las evals sin `avisos` se juzgan igual que antes.
- **Reparto de avisos en el informe**: cada sesión publica en `informe.json` `avisos_encontrados` y `avisos_ausentes`,
  en el orden de la eval —`[]` si no espera ninguno—, y cada aviso ausente da el motivo `aviso ausente: <código>`,
  detrás de los de las citas ausentes; la tabla de sesiones de `informe.md` gana las columnas «Avisos encontrados» y
  «Avisos ausentes». Un rojo por un aviso ausente se distingue así de uno por una cita ausente.
- **Eval informativa de una norma derogada** (`evals/boe-legislacion/18-lrjpac-norma-derogada.yaml`): pregunta por el
  artículo 42 de la Ley 30/1992 sin decir nada de su vigencia y exige consultar y citar su bloque `a42` y trasladar los
  avisos `derogada` y `vigencia-agotada` con su forma fija. Nace `informativa: true` (ADR 0016): se ejecuta y su tasa
  se publica sin decidir el veredicto. La Ley 30/1992 entra en `data/normas.yaml`, sin ninguna marca de derogación, y
  su índice lo graba una persona con `scripts/grabar-evals.sh`.

*De H6 — el applet `territorio` y la skill `legal-core`:*

- **Applet `territorio`, con un solo verbo, `resolver`**, registrado en el binario distribuido junto a `boe`:
  `kitlegal territorio resolver <consulta>` —o `territorio resolver <consulta>` por el enlace— resuelve un municipio
  de España, por su nombre o por su código INE de cinco cifras o de seis con el dígito de control, a su territorio.
  `data` lleva siempre las mismas ocho claves —`municipio`, `codigo_ine`, `provincia`, `comunidad`, `dir3`,
  `regimen`, `boletines` y `cobertura`—, ninguna omitida y cada dato con su `source`: la fila de `docs/SOURCES.md` o
  el fichero de `data/territorio/` del que sale. `boletines` trae siempre el BOE y, además, solo los boletines que la
  comunidad tiene configurados; `cobertura` dice si el boletín autonómico y el provincial están `configurado` o
  `no-configurado` y si el DIR3 del ayuntamiento está `verificado` o `no-verificado`, y ninguno de sus valores
  significa «no existe». Fuera del territorio configurado la salida no nombra ningún boletín que no tenga, y un DIR3
  sin verificar va vacío, nunca calculado. `regimen` marca `comun` o `foral` en todas las comunidades, estén
  configuradas o no. No pide nada a la red ni toca la caché: el sobre lleva `fuente` `kitlegal.territorio`, `url`
  `kitlegal:applet/territorio` y, como `fecha_consulta`, la fecha más antigua de los ficheros que sostienen la
  respuesta, de modo que la misma consulta da la misma salida byte a byte, con `--offline` o sin él y desde cualquier
  directorio. Códigos: `0` resuelto; `2` una entrada que no llega a ser un código —provincia fuera de `01`-`52`,
  municipio `000`— o un dígito de control que no es el oficial, y un nombre que corresponde a más de un municipio,
  cuyo mensaje enumera todos los candidatos como `<código INE> <nombre> (<provincia>)`; `3` un nombre, o un código
  bien formado, que no está en la relación. El applet no decide nunca `4`, `5` ni `6`; solo el plazo de `--timeout`,
  que pone el kernel para todo applet, termina en `4`. Su contrato se publica en `schemas/municipio.json`
  (`$defs.resolver`), generado desde `--describe` y comprobado por `make schema-check` como los de `boe`.
- **Datos congelados de territorio en `data/territorio/`**, versionados y embebidos en el binario, que no los pide
  en red en ningún momento (ADR 0017): `municipios.yaml`, la relación de municipios del INE con su dígito de control,
  su provincia y su comunidad (`source` `ine.municipios`); `dir3.yaml`, el DIR3 del ayuntamiento de cada municipio,
  `L` más su número de inscripción en el Registro de Entidades Locales, solo con los municipios cuyo número es coherente
  con su código INE y su dígito de control (`source` `mpt.rel`); la regla se verificó contra el directorio DIR3 del
  Punto de Acceso General en una muestra con municipios fusionados, forales y con entidades locales menores; `estado.yaml`, el boletín estatal; y `comunidades/`, un fichero por cada una de las 19 comunidades y
  ciudades autónomas con su régimen y sus provincias, cuyos nombres salen de las tablas de códigos del INE (`source`
  `ine.codigos-territoriales`). Solo la Comunidad de Madrid trae `boletines`: el BOCM, a la vez
  autonómico y provincial, con el motivo escrito en su fichero. Añadir un territorio es rellenar los boletines de su
  fichero, sin tocar código ni skills. Una persona los genera fuera del repositorio desde las descargas del INE y del
  REL; `docs/SOURCES.md` lleva la fila de cada origen con la fecha del fichero y `scripts/verify-sources.sh` no gana
  ningún caso, porque no se consultan. Se validan contra `schemas/territorio-municipios.yaml.json`,
  `schemas/territorio-dir3.yaml.json`, `schemas/territorio-estado.yaml.json` y
  `schemas/territorio-comunidad.yaml.json`, y `make skills-check` comprueba además su integridad —toda provincia
  declarada por una sola comunidad, todo DIR3 coherente con el código y el dígito de su municipio—, la procedencia de
  cada dato y que todo municipio se alcance por su nombre oficial, resuelto o entre los candidatos de un nombre
  ambiguo.
- **Jerarquía normativa como dato**, `data/jerarquia.yaml`, validada contra `schemas/jerarquia.yaml.json` dentro de
  `make skills-check`: los cinco niveles en su orden —Unión Europea, Estado, comunidad autónoma, provincia y
  municipio—, con la clase de boletín que publica las normas de cada uno y sus tipos de norma de mayor a menor rango,
  y las cuatro reglas de interpretación: competencia antes que jerarquía, ley posterior, ley especial y reglamento
  nunca contra ley. No lleva el identificador ni el texto de ninguna norma.
- **Las leyes vertebrales, con su identificador verificado**: `data/normas.yaml` gana siete normas, con el
  identificador, el título y el rango copiados de su búsqueda grabada del BOE, y la marca opcional `vertebral: true`
  en las quince de la tabla de leyes vertebrales del mapa del sistema legal, y en ninguna más. `schemas/normas.yaml.json`
  gana esa propiedad y, en el enumerado de rangos, los que traen las búsquedas nuevas. `references/normas.md` de
  `boe-legislacion` se regenera con ellas; su `SKILL.md` y sus evals no cambian.
- **Skill `legal-core` v0** (`skills/legal-core/`), la skill madre y el punto de partida de las preguntas de derecho
  público que dependen de un municipio. Su `SKILL.md`, de menos de 300 líneas, fija un protocolo que empieza por
  identificar el territorio —y pregunta el municipio si la conversación no lo dice, sin suponerlo—, lo resuelve con
  `scripts/territorio resolver … --json`, traslada la `cobertura` a la respuesta sin nombrar ningún boletín que el
  applet no haya devuelto, ofrece todos los candidatos de un nombre ambiguo, no concluye «no existe» de lo que no
  encuentra, razona con sus dos referencias y delega en `boe-legislacion`, en un solo sentido, el texto de cualquier
  artículo. Sus dos referencias se generan con `make skills-sync` y llevan la cabecera que prohíbe editarlas:
  `references/leyes_vertebrales.md`, desde las normas marcadas `vertebral` de `data/normas.yaml`, y
  `references/jerarquia_normativa.md`, desde `data/jerarquia.yaml`. Llama al binario por `scripts/territorio`, un
  enlace al binario instalado, y `make install` la enlaza junto a `boe-legislacion`.
- **Identificadores código INE y DIR3** (`internal/core/ids`): el código INE se analiza con cinco cifras —provincia
  `01`-`52` y municipio `001`-`999`— o con seis, la última el dígito de control, que se compara con el oficial
  nombrando los dos; el DIR3 del ayuntamiento, con la forma `L01PPMMMD`, se analiza con la letra en mayúscula o en
  minúscula y se normaliza a mayúscula, y se compone desde el código y su dígito. Cada error nombra la entrada y dice
  qué tiene de malo, con la clase de argumentos inválidos y nunca con un pánico; `make test` ejecuta sus dos objetivos
  de fuzz sobre un corpus semilla versionado. Los patrones de identificador de los esquemas de territorio aceptan
  exactamente lo mismo que los analizadores, y `make skills-check` lo comprueba.
- **Variante de territorio del formato común de eval**: `comandos` admite una cuarta forma, `applet`, `verbo`
  `resolver` y `municipio`, y la eval gana el esperado opcional `territorio` —`comunidad`, `provincia`, los códigos de
  `boletines` y los aspectos de `cobertura` en la forma `<aspecto>: <valor>` del vocabulario del applet—; una eval que
  activa la skill lleva `citas`, `territorio` o los dos. Se juzga sin modelo: la comunidad y la provincia sin
  distinguir mayúsculas ni tildes, el código de cada boletín como palabra exacta y cada aspecto de cobertura en su
  forma fija, con los blancos y las mayúsculas tolerados; un comando de territorio no necesita nada grabado. El
  informe gana las columnas «Territorio encontrado» y «Territorio ausente», y `make skills-check` comprueba que el
  enumerado de cobertura del esquema y el vocabulario del applet dicen lo mismo. Las evals de `boe-legislacion` siguen
  validando sin cambiar un byte.
- **Evals de `legal-core`** (`evals/legal-core/`), tres, escritas antes que la skill y que deciden todas: qué
  comunidad, provincia y boletines corresponden a un ayuntamiento, sobre un municipio de la Comunidad de Madrid —con
  su comunidad, su provincia y el BOCM en lo esperado— y sobre uno de una comunidad sin configuración —con su
  comunidad, su provincia y los dos boletines declarados no configurados—, y una pregunta ajena que no debe activarla.
  Ninguna lleva citas, porque la skill no afirma el contenido de ninguna norma. `make skills-check` exige de su
  conjunto al menos tres evals, la del municipio configurado, la del no configurado, una de no activación y un
  esperado verificable en toda eval que activa la skill.

*De H19 — instalar sin clonar:*

- **Las skills viajan dentro del binario**: `skills.go`, en la raíz del módulo, empotra el `SKILL.md` y todo
  `references/` de cada skill del repositorio —hoy `boe-legislacion` y `legal-core`—, y es lo único que se instala.
  También un `go install ./cmd/kitlegal` sin inyecciones las lleva y las instala igual. El binario no enlaza ningún
  módulo nuevo.
- **Applet `skills`, con tres verbos y ninguno por omisión**, registrado en el binario distribuido junto a `boe` y
  `territorio`; sin verbo termina con `2`. **`kitlegal skills install [skill…]`** instala las skills empotradas —todas,
  o solo las nombradas— en el directorio neutro del ámbito: `.agents/skills/` del directorio de trabajo por omisión,
  `~/.agents/skills/` con `-g` o la ruta de `--dir`. En los dos primeros, si existe `.claude/` o se pide
  `--host claude`, enlaza cada skill en `.claude/skills/<skill>` con el destino literal relativo
  `../../.agents/skills/<skill>`; donde el sistema no deja crear enlaces, la entrada de host es una copia anotada
  `copia`. Escribe el manifiesto del ámbito, `kitlegal.json`, determinista y sin rutas absolutas, fechas ni usuarios:
  la versión del binario y, por skill, su versión, la huella SHA-256 de cada fichero y sus entradas de host con su
  modo. Solo toca lo que ese manifiesto declara: antes de escribir nada busca los conflictos —carpeta ajena, fichero,
  enlace a otro sitio, enlace roto, fichero editado, fichero ajeno, ruta que no es directorio, manifiesto ilegible y
  manifiesto con entradas de host— y, con uno solo, no crea ni cambia nada y termina con `1`, con la cabecera
  `skills install: nada se ha creado ni cambiado; conflictos:` y una línea `<clase>: <ruta>` por conflicto, en el
  `mensaje` del sobre y en la salida de error. Nunca lee, escribe ni retira a través de un enlace simbólico por debajo
  del ámbito, ni abre lo que no es un fichero regular. Cada skill sale `instalada`, `actualizada` o `sin cambios`:
  repetir la orden con el mismo binario deja el disco byte a byte igual, y con otro sustituye lo instalado y retira lo
  que el binario nuevo ya no empotra; una skill que el manifiesto declara y el binario no empotra se conserva sin
  tocar. Si la escritura falla a mitad, repetir la orden la completa. `--dry-run` no toca el disco y termina con el
  mismo código que la orden real, con una línea por skill en la salida de error.
- **`kitlegal skills list` y `kitlegal skills doctor`**, que no cambian nada en disco. `list` enumera lo que declara el
  manifiesto del ámbito —el directorio, su versión y, por skill, su versión, si el binario la empotra y sus
  enlaces—. `doctor` comprueba lo instalado contra el manifiesto y da hallazgos de cinco clases —fichero editado,
  enlace colgando, enlace a otro sitio, copia y versión distinta—, cada uno como `<clase>: <ruta>: <orden>`, con una
  orden de shell POSIX de una línea que lo arregla (`rm -- '<ruta>' && kitlegal skills install <skill> …`, con `-r`
  solo sobre un directorio real y nunca `-f`), en un orden en que ejecutarlas una tras otra deja el siguiente `doctor`
  sin hallazgos; con alguno termina con `1`. Sin manifiesto, los dos terminan con `0`, con versión nula y lista vacía;
  ante un manifiesto ilegible o una ruta del ámbito que no es un directorio, con `1`.
- **Validación de la invocación antes de mirar el disco**: `-g` con `--dir`, `--host` con `--dir`, `--host` distinto
  de `claude` y una skill que el binario no lleva —con las disponibles en el mensaje— terminan con `2`, en ese orden de
  precedencia; `-g` sin `HOME` termina con `1`. El applet no pide nada a la red, no emite operaciones de grafo y firma
  su sobre con `fuente` `kitlegal.skills` y `url` `kitlegal:applet/skills`. Su contrato se publica en
  `schemas/instalacion.json` (`$defs.install`, `$defs.list` y `$defs.doctor`), generado desde `--describe` y
  comprobado por `make schema-check`.
- **Aviso de versión, sin red.** Toda orden de otro applet resuelta con un verbo —también si termina con un error de
  argumentos del verbo, con `--dry-run` o con `--describe`— busca, sin seguir enlaces, el manifiesto del directorio de
  trabajo y, si no lo hay, el de la cuenta. Si su versión, o la de alguna skill declarada que el binario empotra, es
  distinta de la del binario —quitando a cada una una `v` inicial—, escribe exactamente una línea en la salida de error:
  `aviso: las skills instaladas son de kitlegal <instalada> y este binario es kitlegal <binario>; ejecuta: kitlegal
  skills install`, con ` -g` si el manifiesto es el de la cuenta. No cambia la salida estándar ni el código de salida,
  y un fallo al escribirla no se propaga. No avisan `skills`, `version`, la ayuda, los fallos anteriores a resolver el
  applet con su verbo ni un binario de desarrollo, cuya versión no tiene forma SemVer.
- **La release**: `.goreleaser.yaml`, con goreleaser v2.18.1 como módulo de herramienta (`tools/goreleaser/`), construye
  darwin, linux y windows en amd64 y arm64, sin cgo, con `-trimpath` y con las mismas cuatro inyecciones que el
  `Makefile` —en una etiqueta, la versión es la etiqueta tal cual—; archivos `kitlegal_<os>_<arch>.tar.gz` (`.zip` en
  Windows), `checksums.txt`, que incluye `install.sh`, un SBOM por archivo (syft), la firma sin clave de
  `checksums.txt` (cosign), notas de release agrupadas por Conventional Commits, el cask del tap
  `jmorenobl/homebrew-tap` para `brew install jmorenobl/tap/kitlegal` —con el gancho que retira la cuarentena de
  macOS—, el manifiesto del bucket `jmorenobl/scoop-bucket` para `scoop install kitlegal`, y paquetes `.deb` y `.rpm`.
  `make goreleaser-check` la valida dentro de `make ci`; `make release` construye el snapshot en `dist/` sin publicar,
  firmar ni generar SBOM; y `make snapshot-check` lo comprueba (`TestSnapshot`) y ejecuta contra él los guiones del
  instalador. `TestConfiguracionDeLaRelease` fija la configuración y los flujos.
- **Flujo `release`** (`.github/workflows/release.yml`), solo al empujar una etiqueta `v*`: publica con
  `PUBLISHER_TOKEN` —el único secreto, visible solo en el paso que publica—, atesta la procedencia de los seis
  archivos y de `checksums.txt` y comprueba lo publicado en un trabajo de humo: la huella, `gh attestation verify`,
  que `version` imprime la etiqueta, que `boe articulo … --offline` con la caché vacía termina con `4` y que
  `skills install` deja `boe-legislacion` en un directorio vacío. Y el trabajo **`snapshot`** del flujo `ci`, en cada
  propuesta de cambio y en cada push a `main`, sin `id-token` ni secretos: `make release` y `make snapshot-check`.
- **`scripts/install.sh`**, el instalador POSIX `sh` de macOS y Linux, sin Go, git ni gestor de paquetes:
  `curl -fsSL https://raw.githubusercontent.com/jmorenobl/kitlegal/main/scripts/install.sh | sh`, o con una versión
  (`sh -s -- 0.1.0`, con o sin `v`). Detecta el sistema y la arquitectura, descarga el archivo y `checksums.txt` de la
  release, verifica la huella en la línea cuyo segundo campo es exactamente el archivo, deja solo el binario en
  `$KITLEGAL_INSTALL_DIR` o `~/.local/bin` renombrándolo encima del anterior, no toca ningún fichero de arranque del
  shell, imprime la línea `export PATH=…` si el directorio no está en el `PATH` y termina con
  `kitlegal skills install`. Todo error sale por la salida de error con el prefijo `install.sh: `, un código distinto
  de `0` y nada instalado. Se sirve desde `main` y va adjunto a cada release; los guiones `instalador-` del e2e lo
  prueban contra un origen local y contra el snapshot, sin red.

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
  deja el mismo estado. Antes del `go install`, el mismo guion con `--comprobar` busca las entradas del directorio
  personal con el nombre de una skill que no son su enlace —un directorio, un fichero, un enlace a otro sitio o un
  enlace roto—: escribe una línea de conflicto por cada una y falla sin crear ni cambiar nada, tampoco el binario.
  `make test-integration` lo prueba (`TestInstalacion`) sobre una copia mínima del árbol, con el directorio personal,
  el de binarios y el `GOPATH` temporales y sin red.
- **`make ci` encadena diez controles**: `skills-check` entra tras `schema-check`, de modo que una skill con una
  deriva o un defecto, una tabla de normas inválida o una eval mal formada hacen fallar el veredicto en local y en la
  integración continua por igual. El análisis estático alcanza también los ficheros con la etiqueta de compilación
  `evals` (`run.build-tags` de `.golangci.yml`), los arneses que usa el job de evals.

*De H6 — el applet `territorio` y la skill `legal-core`:*

- **El job de evals mide las dos skills en la misma ejecución.** El flujo `evals` deja de evaluar solo
  `boe-legislacion` y pasa a una matriz de skills, `boe-legislacion` y `legal-core`: un trabajo por skill, cada uno
  con su informe, y el rojo de uno no cancela el otro (`fail-fast: false`). La prueba de red sigue siendo solo de
  `boe-legislacion`, cuyo texto invoca el applet `boe`: la entrada `prueba_de_red` y la etiqueta
  `evals-prueba-de-red` la añaden a su trabajo y no al de `legal-core`, porque `territorio` no puede pedir nada a la
  red. Al abrirse o reabrirse una propuesta de cambio, el job arranca también si toca los paquetes de dominio
  (`internal/core/`); los ficheros congelados de `data/territorio/` ya los cubría `data/`. El umbral no cambia: cada
  serie de tres sesiones pasa con dos, y el informe lleva el commit evaluado y el identificador del modelo.

*De H19 — instalar sin clonar:*

- **Las skills invocan `kitlegal` desde el `PATH`.** La tabla de comandos de cada `SKILL.md`, que regenera
  `make skills-sync`, titula cada applet `kitlegal <applet>` y escribe cada orden `kitlegal <applet> <verbo> …`; en el
  texto libre de `boe-legislacion` y `legal-core`, `scripts/boe` y `scripts/territorio` pasan a `kitlegal boe` y
  `kitlegal territorio`, y las frases que decían de dónde sale el binario dicen ahora que se invoca desde el `PATH`.
  Protocolo, reglas, forma de la cita y de los avisos, y evals, sin cambios.
- **`make install` es el bucle de desarrollo**, no la forma de instalar: `go install` con las inyecciones del
  `Makefile` y, con ese binario, `kitlegal skills install -g --host claude`, que deja las skills en `~/.agents/skills/`,
  con su manifiesto, y enlazadas en `~/.claude/skills/<skill>` con destino `../../.agents/skills/<skill>`. Ya no
  comprueba nada antes de `go install`: un conflicto lo da `skills install`, que termina con `1` sin cambiar nada,
  con el binario ya instalado. Los enlaces absolutos que dejaba el `make install` anterior son un conflicto («enlace a
  otro sitio»), y `CONTRIBUTING.md` da el paso único que los retira. `TestInstalacion` y sus cuatro guiones lo
  comprueban con `HOME`, `GOBIN` y `GOPATH` temporales.
- **`make skills-sync` y `make skills-check` dejan los enlaces**: ni los generan ni los comprueban; una skill con
  `scripts/` es un defecto que hace fallar a los dos (ADR 0019), y `make skills-check` comprueba además que cada orden
  de la tabla de comandos de cada skill empotrada nombra un applet y un verbo del binario
  (`TestOrdenesDeLasSkillsEmpotradas`).
- **`make ci` encadena once controles**: `goreleaser-check` entra tras `skills-check`.
- **El job de evals instala con `make install`** y añade al `PATH` de cada trabajo el directorio donde `go install`
  deja el binario; `scripts/evals.sh` comprueba también que `kitlegal` está en el `PATH`, y la sesión de la prueba de
  red pide `kitlegal boe articulo BOE-A-2015-10565 a9998 --json` y la misma con `--offline`. Ninguna eval cambia.
- **La ayuda y los errores del binario enumeran `skills`**: `kitlegal --help` lo lista y un applet desconocido termina
  en `applets disponibles: boe, skills, territorio`.
- **`make test-e2e` construye además tres binarios de extremo a extremo** —con la versión `v0.1.0`, con `v0.2.0` y con
  `v0.1.0` y un creador de enlaces que siempre falla— y un origen de release local, que usan los guiones del aviso, del
  recurso de copia y del instalador.

### Eliminado

*De H19 — instalar sin clonar:*

- **La instalación por enlaces**: `skills/boe-legislacion/scripts/boe`, `skills/legal-core/scripts/territorio`,
  `scripts/instalar-skills.sh` con la comprobación previa de `make install`, `bin/instalado/`,
  `internal/skills/enlaces.go` con sus tests y las derivas de enlaces de `skills-sync` y `skills-check`
  (`enlace-ausente`, `enlace-sobrante` y `enlace-con-otro-destino`). `TestSinInstalacionPorEnlaces` comprueba que no
  vuelven y que ni los `SKILL.md`, ni el `Makefile`, ni los flujos nombran `scripts/boe`, `scripts/territorio` ni
  `bin/instalado`.

### Corregido

*De H5.1 — los avisos de vigencia en las evals:*

- **La traza de una sesión con la llamada desconocida cerrada se lee entera.** Un hilo que muere en la parada de
  entrada de una llamada que strace no llega a identificar deja `???( <unfinished ...>`, que la lectura de la traza ya
  admitía, o la misma llamada cerrada con el resultado de la llamada sin terminar y el relleno de alineación,
  `???()` seguido de espacios y `= ?`, que declaraba la sesión ilegible y hacía fallar el veredicto del job aunque
  todas las series pasaran. Así ocurrió con una sesión del modelo informativo en la ejecución `35156339496`. La
  segunda forma se lee ahora como la primera; con argumentos, con otro resultado o con la marca dentro sigue siendo
  ilegible.

Ninguna orden del `Makefile` espera ya su contenido de un hito posterior: `release`, que fallaba hasta H19, construye
el snapshot de la release. El binario que se publica registra **tres applets, `boe`, `skills` y `territorio`**: los de
las demás fuentes (`placsp`, `bdns`…) llegan en los hitos siguientes, en el orden de `docs/ROADMAP.md`.
