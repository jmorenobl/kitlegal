# kitlegal

`kitlegal` es un conjunto de **skills agénticas** para consultar fuentes legales públicas españolas
—BOE, PLACSP, BDNS, BORME, EUR-Lex…— y actuar en tu municipio, sea cual sea: qué contrata y subvenciona
tu ayuntamiento, qué dicen sus ordenanzas, cuándo vence un plazo, qué escrito presentar. Las skills
razonan; las herramientas deterministas que usan las da un binario Go **multicall**, que también se
puede usar desde la línea de órdenes. Un solo ejecutable: el applet lo elige `os.Args[0]` o el primer
argumento, de modo que `kitlegal boe articulo …` y un symlink `boe -> kitlegal` son la misma cosa.
Toda respuesta va envuelta en `{ok, fuente, url, fecha_consulta, hash, data}`: sin fuente, URL, fecha
de consulta y hash no hay cita, y sin cita no hay respuesta.

Solo se automatizan fuentes públicas. Cualquier acción que exija identidad —presentar un escrito,
recoger una notificación— termina en un fichero listo para firmar, nunca en un envío a una sede.

## Qué entrega este hito (H6)

H6 trae el **segundo applet**, `territorio`, y la **skill madre**, `legal-core`: toda pregunta que depende de un
municipio empieza por su territorio, que el binario resuelve desde datos congelados sin inventar lo que no está
configurado.

```console
$ ./bin/kitlegal territorio --help
uso: territorio <verbo> [banderas]

Resuelve un municipio de España a su provincia, su comunidad, su régimen, el DIR3 de su ayuntamiento y sus boletines.

verbos:
  resolver  Devuelve el territorio de un municipio, por su nombre o por su código INE, con la cobertura de lo que está configurado y verificado.

Las banderas de cada verbo, en «territorio <verbo> --help».
```

- **`kitlegal territorio resolver Leganés`** —o por su código INE, `28074`, o con su dígito de control, `280745`—
  devuelve el municipio, su código INE, su provincia, su comunidad, el DIR3 de su ayuntamiento, su régimen (`comun` o
  `foral`), sus boletines y la `cobertura` de lo que está configurado y verificado, cada dato con su `source`. Fuera
  del territorio configurado —hoy, la Comunidad de Madrid— trae solo el BOE y declara `no-configurado` el boletín
  autonómico y el provincial, sin nombrar ninguno: lo que no está configurado nunca se presenta como inexistente. No
  pide nada a la red, así que responde igual con `--offline`. Los fallos salen con los códigos estables: `2` una
  entrada que no es un código bien formado, un dígito de control que no es el oficial o un nombre de varios
  municipios, con todos los candidatos en el mensaje; `3` un municipio que no está en la relación; nunca `4`, `5` ni
  `6`. Su contrato está publicado en `schemas/municipio.json`.
- **Datos congelados** en `data/territorio/`: la relación de municipios del INE, el DIR3 del ayuntamiento de cada
  municipio —derivado del número de inscripción del Registro de Entidades Locales, con la regla verificada contra el
  directorio DIR3 oficial—, el boletín estatal y un fichero por comunidad y ciudad autónoma con su régimen y, si está
  configurada, sus boletines. Los genera una persona fuera del repositorio (ADR
  0017), viajan dentro del binario y `make skills-check` los valida.
- **Skill `legal-core`**: identifica el territorio antes de razonar —y pregunta el municipio si no se dice—, lo
  resuelve con el applet, traslada la cobertura sin nombrar ningún boletín que el applet no haya devuelto y razona con
  dos referencias generadas desde los datos: las leyes vertebrales, con su identificador `BOE-A-…` comprobado contra
  la búsqueda grabada del BOE, y la jerarquía normativa de `data/jerarquia.yaml`. El texto de un artículo lo delega
  en `boe-legislacion`.
- **Identificadores** código INE y DIR3, con su gramática y su dígito de control, y **la variante de territorio del
  formato común de eval**, con la que se miden las tres evals de `legal-core`.

H5 trajo la **primera skill del producto**: `boe-legislacion`, que consulta y cita cualquier norma consolidada del
BOE —procedimiento administrativo, contratación pública, régimen local, tributos, relaciones laborales…— con el
applet `boe` del binario, y el andamiaje que comparten todas las skills: los datos como única fuente de verdad, lo
generado comprobado en `make ci`, la instalación con una orden y unas evals comparables entre ejecuciones.

- **Skill `boe-legislacion`**: un protocolo de cinco pasos —identificar la norma, resolver su identificador
  `BOE-A-…`, leer el índice y los bloques con el binario, evaluar si falta contexto y responder citando— y una forma
  de cita de la que se extraen mecánicamente la norma y el bloque: la forma legible seguida de corchetes que terminan
  en el identificador y el id del bloque, como `art. 21 de la Ley 39/2015 [BOE-A-2015-10565, bloque a21]`, con la
  forma legible delante del corchete o, dentro, delante del identificador. Lo que dice de una norma sale del texto que
  el binario devuelve en la misma conversación, distinguiendo ley y reglamento y señalando la variación autonómica.
- **Tabla de normas** `data/normas.yaml`: once normas —las diez de materias distintas que citan las evals que deciden
  y, desde H5.1, la Ley 30/1992, derogada—, validada contra `schemas/normas.yaml.json` y con cada identificador
  comprobado contra la búsqueda grabada del BOE. De ella se generan las referencias de la skill.
- **`make install`** instala el binario y enlaza la skill en el directorio personal de skills de Claude Code;
  **`make skills-sync`** regenera lo que se deriva de los datos y del binario; **`make skills-check`**, dentro de
  `make ci`, falla si algo diverge; y **`make evals`** mide la skill con Claude Code desde el job de evals.
- **El formato común de eval** (`schemas/eval.yaml.json`) y dieciocho evals de `boe-legislacion`: diez preguntas de
  materias distintas que deben activar la skill, hacer las consultas esperadas y citar lo esperado, dos que no
  deben activarla y seis informativas, que se miden sin decidir el veredicto (ADR 0016).

El detalle está en [Skills](#skills).

H4 trajo el **primer applet con fuente**: `boe`, que consulta la API de Legislación Consolidada del BOE
—cualquier norma consolidada, estatal o autonómica, por su identificador `BOE-A-…`— y devuelve cada
respuesta lista para citar, con `fuente` `boe.legislacion-consolidada`, la `url` de la API consultada y
la `fecha_consulta` en que se obtuvo lo que `data` contiene.

```console
$ ./bin/kitlegal boe --help
uso: boe <verbo> [banderas]

Consulta la legislación consolidada del BOE y la devuelve lista para citar.

verbos:
  buscar     Busca normas consolidadas por las palabras de su título o con una consulta de la fuente.
  indice     Devuelve los bloques de una norma consolidada, en el orden de la fuente.
  articulo   Devuelve el texto vigente de un bloque de una norma, con los avisos de su vigencia.
  articulos  Devuelve el texto vigente de varios bloques de una norma, en el orden pedido.
  metadatos  Devuelve los datos de una norma y los avisos de su vigencia.
  analisis   Devuelve las materias, las notas y las referencias de una norma.

Las banderas de cada verbo, en «boe <verbo> --help».
```

- **`kitlegal boe articulo BOE-A-2015-10565 a21`** devuelve el texto vigente del artículo 21 de la Ley
  39/2015, su huella, su dirección pública y la ELI de la norma, con los avisos de su vigencia. Es el
  porte de `refs/boe.py`, la skill `boe-fiscal` en Python: nunca emite un texto cuya vigencia no ha
  podido comprobar.
- **Caché local**: la segunda consulta idéntica no sale a la red. Vive en `~/.cache/kitlegal/cache.db`
  (en otra carpeta con `KITLEGAL_CACHE_DIR`), con una vigencia por verbo —5 minutos para `buscar` y
  `metadatos`, 7 días para los demás—, y nunca guarda un fallo. Lo servido desde la caché conserva la
  `fecha_consulta` de cuando se pidió, no la de la invocación. `--offline` responde solo con lo guardado
  y vigente —si falta, código `4` sin pedir nada— y `--dry-run` describe en la salida de error cada
  petición que habría emitido, sin emitir ninguna.
- **Una fuente pública, pedida con cuidado**: cada petición se identifica, respeta el `robots.txt`, deja
  al menos un segundo con la anterior al mismo sitio y se reintenta ante fallos transitorios. Licencia,
  términos de uso y ritmo están en la fila de la fuente de [`docs/SOURCES.md`](docs/SOURCES.md). Los
  fallos salen con los códigos estables: `2` una norma o un bloque mal escritos, `3` lo que la fuente no
  tiene, `4` la fuente caída o una respuesta que ya no se sabe interpretar, `5` límite de peticiones o
  `robots.txt`; nunca `6`.
- **Contratos publicados**: `schemas/norma.json` (`buscar`, `indice`, `metadatos`, `analisis`) y
  `schemas/bloque.json` (`articulo`, `articulos`) describen la entrada y la salida de cada verbo. Se
  generan desde `--describe`, toda salida se valida contra ellos en los tests y `make schema-check`
  falla si divergen.

Debajo están los hitos de fundación, que `boe` hereda sin escribir nada de ellos. H1 es el **kernel de
la línea de órdenes**, que fija la forma de invocarse, de fallar y de citar de todos los applets:

- **Despacho multicall**: el applet lo elige el nombre de invocación —un enlace `boe -> kitlegal`
  ejecuta el applet `boe`— y, si ese nombre no está registrado, el primer argumento.
- **Ocho banderas globales** que ningún applet escribe: `--json`, `--timeout`, `--offline`,
  `--dry-run`, `--describe`, `--no-graph`, `--asunto` y `--verbose`.
- **Sobre de salida** `{ok, fuente, url, fecha_consulta, hash, data}` y **códigos de salida estables**
  (`0` ok · `2` args · `3` no encontrado · `4` fuente no disponible · `5` límite o TOS · `6` requiere
  identidad humana; `1` queda para el fallo inesperado), traducidos en un único punto.
- **`--describe`**: esquema JSON 2020-12 de la entrada y la salida de cada verbo, derivado por
  reflexión y sin nada escrito a mano, para que un agente descubra el contrato sin leer código.

H2 dejó `internal/httpx`, la única puerta a la red, y H3 la caché local en SQLite. H0 dejó el
**esqueleto del repositorio y sus controles**: el `Makefile` como única superficie de invocación, y los
controles de formato, análisis estático, tests con detector de carreras, vulnerabilidades conocidas,
secretos e integridad de los módulos.

El binario distribuido registra **dos applets, `boe` y `territorio`**: los de ejemplo (`echo`, `contar`) viven solo
en los tests. Los applets de las demás fuentes (`placsp`, `bdns`…) llegan en hitos posteriores; el orden
está en [`docs/ROADMAP.md`](docs/ROADMAP.md).

## Prerrequisitos

Exactamente dos, y nada más:

| Prerrequisito | Comprobación |
|---|---|
| Go 1.21 o superior | `go version` |
| `git` | `git --version` |

**No hay que instalar ninguna herramienta de control.** `golangci-lint`, `govulncheck`, `gitleaks` y
`lefthook` se construyen solos, con la versión fijada en `tools/<herramienta>/go.mod`, la primera vez
que se invoca la orden que los usa.

**Tampoco hay que instalar un parche concreto de Go, ni importa cuál tengas.** `go.mod` declara la
directiva `toolchain` y el `Makefile` exporta `GOTOOLCHAIN` con ese valor, así que todas las órdenes
se ejecutan con ese parche exacto —el mismo que ejecuta la integración continua— y el go command lo
descarga y lo verifica solo si falta. Cualquier `go` ≥ 1.21 sirve.

```bash
git clone https://github.com/jmorenobl/kitlegal.git
cd kitlegal
make check-tools     # comprueba go, git y que el toolchain fijado es obtenible
```

> La **primera** ejecución compila las herramientas desde fuente y, si el parche fijado no está en la
> caché, lo descarga: requiere red y tarda varios minutos. Las siguientes las sirve la caché de
> construcción de Go en segundos.

## Construir e instalar

```bash
make build       # deja el ejecutable en bin/kitlegal
make install     # instala el binario en el directorio de binarios de Go ($GOBIN, o $HOME/go/bin) y enlaza las skills en el directorio personal de skills de Claude Code (~/.claude/skills)
```

Las dos inyectan los mismos datos de construcción —versión, commit y fecha— y compilan sin cgo y con
`-trimpath`. Qué enlaza `make install`, y cómo, está en la sección siguiente.

## Skills

Las skills son el producto y el binario, su herramienta. Una skill es un directorio sin código: un `SKILL.md` con
el protocolo de razonamiento, la tabla de comandos y las reglas; `references/`, generadas desde `data/*.yaml`; y
`scripts/`, con un enlace por applet que llega al binario instalado. Hoy hay dos: `boe-legislacion`, que consulta y
cita la normativa consolidada del BOE, y `legal-core`, que identifica el territorio y razona con la jerarquía normativa
y las leyes vertebrales.

### Tres directorios llamados `skills`

| Directorio | Qué es | ¿Es kitlegal? |
|---|---|---|
| `skills/` | **El producto que se distribuye**: las skills de kitlegal, hoy `boe-legislacion` y `legal-core` | sí |
| `.agents/skills/` | Skills de agente vendorizadas para trabajar en este repositorio: las de Go de `samber/cc-skills-golang`, registradas con su origen y su huella en el registro de bloqueo `skills-lock.json`, y las `speckit-*` que genera la integración `agy` de spec-kit para Antigravity (registradas en `.specify/integrations/agy.manifest.json`). Se versionan tal cual y no se editan; `.agents/.gitattributes` las marca como vendorizadas y generadas, para que no cuenten en las estadísticas de lenguaje del repositorio ni se desplieguen en los diffs de las propuestas de cambio | no |
| `.claude/skills/` | Lo que carga Claude Code al trabajar en el repositorio: un enlace a cada skill de `.agents/skills/` más las skills de spec-kit, con las que se prepara cada hito | no |

### Instalar las skills

`make install` —la orden de [Construir e instalar](#construir-e-instalar)— enlaza cada skill de `skills/` en
`~/.claude/skills/`, el directorio personal de skills de Claude Code, y deja `bin/instalado/kitlegal` apuntando al
binario recién instalado. Es lo que alcanzan los enlaces de `scripts/` de cada skill —`scripts/boe` en
`boe-legislacion` y `scripts/territorio` en `legal-core`—, y como cada uno se invoca con el nombre de su applet, el
despacho multicall ejecuta ese applet. Repetir la instalación deja el mismo estado. Si en
`~/.claude/skills/` ya hay una entrada con el nombre de una skill que no es el enlace que crearía —un directorio, un
fichero, un enlace a otro sitio o un enlace roto—, la nombra, falla y no crea ni cambia nada: los conflictos se buscan
antes de `go install`, así que tampoco se instala el binario.

Con las skills enlazadas basta preguntar a Claude Code por una norma —«¿qué dice el art. 21 de la Ley 39/2015?»—:
`boe-legislacion` se activa, consulta el BOE con `scripts/boe … --json` y responde citando
`art. 21 de la Ley 39/2015 [BOE-A-2015-10565, bloque a21]`. Y por el territorio de un ayuntamiento —«¿qué comunidad,
provincia y boletines corresponden a mi ayuntamiento, el de Leganés?»—: `legal-core` lo resuelve con
`scripts/territorio resolver Leganés --json` y responde con lo que devuelve el applet, la Comunidad de Madrid, la
provincia de Madrid y el BOCM; para un municipio de una comunidad sin configuración, dice que ni el boletín autonómico
ni el provincial están configurados, sin nombrar ninguno.

### Lo generado: `make skills-sync` y `make skills-check`

`data/*.yaml` es la única fuente de verdad: hoy `data/normas.yaml`, la tabla de normas, y `data/jerarquia.yaml`, la
jerarquía normativa, validadas contra `schemas/normas.yaml.json` y `schemas/jerarquia.yaml.json`. Los ficheros
congelados de `data/territorio/` no generan nada: viajan dentro del binario y los lee `territorio resolver`. De los
datos y de `--describe` del binario se derivan tres cosas de cada skill, que no se editan a mano: `references/*.md`,
con la cabecera `<!-- generado desde data/<fichero>.yaml, no editar -->`, que nombra el fichero del que sale cada
una; la tabla de comandos de `SKILL.md`, entre sus marcas de inicio y de fin; y los enlaces de `scripts/`.

- **`make skills-sync`** las regenera y escribe en el árbol; dos ejecuciones seguidas no cambian nada. Se ejecuta
  tras cambiar `data/` o un verbo, y lo que regenera va en el mismo cambio, porque sin ello `make ci` falla.
- **`make skills-check`** las regenera en memoria y las compara con el árbol sin escribir nada; comprueba además el
  frontmatter de cada `SKILL.md` y que tenga menos de 300 líneas, la tabla de normas contra su esquema y sus
  identificadores contra la búsqueda grabada del BOE, los ficheros congelados de `data/territorio/` contra sus
  esquemas y su integridad, la jerarquía normativa contra el suyo, y el formato y el conjunto de las evals y lo grabado
  que necesitan. Está dentro de `make ci`: una referencia editada a mano, un dato sin regenerar o un `--describe` que
  ha cambiado lo hacen fallar nombrando la skill y el fichero o el enlace.

### Formato común de eval

Una eval es una pregunta y lo que se espera de la sesión que la responde. Las evals de cada skill viven en un
directorio propio, `evals/<skill>/`, con un fichero YAML por eval llamado `<nn>-<descripción>.yaml`, y todas siguen
el formato común de eval:

```yaml
pregunta: "¿qué dice el art. 21 de la Ley 39/2015?"
activa: true
comandos:
  - applet: boe
    norma: BOE-A-2015-10565
    bloque: a21
citas:
  - norma: BOE-A-2015-10565
    bloque: a21
```

| Campo | Qué fija |
|---|---|
| `pregunta` | La pregunta con la que se abre la sesión |
| `activa` | Si la pregunta debe activar la skill. Una eval de no activación (`false`) no lleva `comandos` ni `citas` |
| `comandos` | Obligatorio si `activa` es `true`: las consultas que la sesión debe hacer con éxito, en una de cuatro formas —un bloque (`applet`, `norma`, `bloque`), una consulta de norma (`applet`, `verbo` `indice`, `metadatos` o `analisis`, `norma`), una búsqueda (`applet`, `verbo` `buscar`, `terminos`) o un municipio (`applet`, `verbo` `resolver`, `municipio`)— |
| `citas` | Si `activa` es `true`, obligatorio salvo que la eval declare `territorio`: cada `norma` y `bloque` que la respuesta debe citar |
| `territorio` | Solo si `activa` es `true`, y obligatorio si no hay `citas`: lo que la respuesta debe declarar del territorio que devuelve `territorio resolver` —`comunidad`, `provincia`, los códigos de `boletines` y los aspectos de `cobertura` en la forma `<aspecto>: <valor>`, como `boletin_autonomico: no-configurado`—, con al menos una de esas claves |
| `avisos` | Opcional, solo si `activa` es `true`: los códigos de aviso de vigencia del binario (`consolidacion-no-finalizada`, `derogada`, `vigencia-agotada`) cuya forma fija —`⚠`, la etiqueta del aviso y dos puntos— debe llevar la respuesta |
| `informativa` | Opcional: con `true`, la eval se ejecuta solo con el modelo que decide y su tasa se publica, pero no decide el veredicto (ADR 0016) |
| `reproduce` | Opcional: la skill cuyo uso documentado reproduce la eval, p. ej. `boe-fiscal` |

Cada fichero de cada directorio `evals/<skill>/`, sea de la skill que sea, se valida contra el esquema
`schemas/eval.yaml.json` dentro de `make ci` (`make skills-check`): una
clave desconocida o repetida, un identificador mal escrito o una eval positiva sin citas ni territorio fallan
nombrando el fichero. `evals/boe-legislacion/` tiene dieciocho. Deciden el veredicto doce: diez preguntas de materias distintas que
deben activar la skill y dos ajenas que no deben activarla. Las otras seis activan la skill y son informativas: cinco
preguntas por materia, que no nombran el artículo y esperan a la herramienta del backlog que lo busca dentro de la
norma, y una sobre la Ley 30/1992, derogada, que exige trasladar sus dos avisos de vigencia y cuya promoción a
decisoria se decidirá con los datos de varias ejecuciones. `evals/legal-core/` tiene tres, y deciden todas: qué
comunidad, provincia y boletines corresponden a un ayuntamiento, sobre un municipio de la Comunidad de Madrid y sobre
uno de una comunidad sin configuración, y una pregunta ajena que no debe activar la skill.

### Job de evals

`make evals SKILL=<skill>` ejecuta las evals de una skill con Claude Code, con la skill instalada por `make install`.
Abre cada eval varias veces con el modelo que decide y, si no es informativa, otras tantas con cada modelo
informativo, y escribe un informe que juzga cada sesión sin modelo —si la abrió el modelo pedido, si terminó, si activó
la skill, si hizo con éxito las consultas esperadas, si citó lo esperado, si declaró el territorio esperado y si la
respuesta lleva la forma fija de cada aviso de `avisos`— y publica la tasa de cada eval con cada modelo: cuántas de sus sesiones pasan. El veredicto global
falla si una eval que decide no llega al umbral de sesiones que pasan con el modelo que decide, si una eval no tiene
exactamente las sesiones que pide el plan, si una sesión es ilegible, si un fichero de eval está mal formado o si una
petición llega a la red; la tasa de las evals informativas y la de los modelos informativos se publican sin decidirlo.
Las sesiones no piden nada a las fuentes: el binario responde desde una caché preparada con lo grabado y la red solo
alcanza el modelo. Necesita Linux con `strace`, root o `sudo`, ningún Python accesible y la credencial de Claude Code;
cuesta y no es determinista, así que no forma parte de `make ci`. Lo lanza el job de evals, el flujo `evals` de la
plataforma, que fija en su definición el modelo que decide, los informativos, las repeticiones de cada eval con cada
modelo y el umbral (ADR 0016): decide `claude-sonnet-5`, el modelo del uso real de la skill;
`claude-haiku-4-5-20251001` se ejecuta como límite inferior sin decidir; y cada eval se abre tres veces con cada uno
—las informativas, solo con el que decide— y su serie pasa si pasan al menos dos sesiones. Cada skill —hoy
`boe-legislacion` y `legal-core`— se evalúa en un trabajo propio de la misma ejecución, con su informe, y el rojo de
uno no cancela el otro. Se lanza:

- **a mano**, sobre la rama que se elija;
- **al abrir o reabrir una propuesta de cambio** que toque lo que las evals miden —las skills y sus datos, las evals y
  su esquema, el applet `boe`, los paquetes de dominio, el kernel que los invoca, el arnés del job, el `Makefile` o el
  propio job—;
- **por etiqueta**, sobre la rama de cualquier propuesta de cambio: poner la etiqueta `evals` lanza las evals antes
  de fusionar, y `evals-prueba-de-red` añade además la sesión de la prueba de red al trabajo de `boe-legislacion`.

No hay ejecución programada, y un empujón a una propuesta ya abierta no la relanza. El informe se imprime en el registro de la ejecución. Cómo se lanza en un hito y qué hay que ver en él, en
[`CONTRIBUTING.md`](CONTRIBUTING.md).

## `kitlegal version`

Verbo reservado del kernel, que se reconoce antes que el registro de applets. Sin banderas y sin
subverbos:

```console
$ ./bin/kitlegal version
kitlegal a3dee64-dirty
commit: a3dee64269ee98da45b2f0a96202899fdaac9354
fecha:  2026-09-10T19:27:47Z
```

Tres líneas en la salida estándar, salida de error vacía y código de salida `0`. La versión sale de
`git describe --tags --always --dirty`: mientras no haya ninguna etiqueta es el commit abreviado, con
el sufijo `-dirty` si el árbol tiene cambios sin confirmar; en cuanto exista `v0.1.0` pasará a leerse
`v0.1.0-3-ga3dee64`. El commit coincide carácter a carácter con `git rev-parse HEAD` de la revisión
construida y la fecha es el instante de construcción en UTC. Un binario hecho con `make build` o
`make install` nunca imprime los valores por defecto del código (`dev`, `none`, `unknown`); si los
ves, lo estás ejecutando con `go run`.

Cualquier otra invocación —sin applet, con un applet desconocido o con un argumento sobrante tras
`version`— escribe el fallo en la salida de error, nombrando lo que no ha reconocido y enumerando los
applets que existen, y termina con código `2`, que es «args» en la tabla de códigos de salida
estables del proyecto:

```console
$ ./bin/kitlegal inventado
argumentos inválidos: "inventado" no es ningún applet de kitlegal; applets disponibles: boe, territorio
$ echo $?
2
```

`kitlegal --help` describe el uso y enumera los applets registrados, con código `0`.

## Ejecutar los controles

`make ci` es **el veredicto del repositorio**: encadena los diez controles, falla nombrando el que
falla y no modifica ningún fichero versionado del árbol de trabajo.

```bash
make ci
```

La integración continua ejecuta esta misma orden, así que el veredicto local predice el de la
propuesta de cambio. Cada control se puede invocar por separado mientras se depura:

| Orden | Qué hace | ¿En `ci`? |
|---|---|---|
| `make fmt-check` | Comprueba el formato sin tocar ningún fichero | sí |
| `make lint` | Análisis estático completo, `gosec` incluido | sí |
| `make test` | Tests unitarios con detector de carreras; deja `coverage.out` | sí |
| `make test-integration` | Tests con la etiqueta de compilación `integration`, con detector de carreras: los que dependen del entorno (permisos del sistema de ficheros, dos procesos, la instalación de las skills con `make install` en un directorio personal temporal), siempre dentro de directorios temporales | sí |
| `make vuln` | Vulnerabilidades conocidas (consulta la base de datos de Go: requiere red) | sí |
| `make schema-check` | Comprueba que los esquemas publicados en `schemas/` (`norma.json` y `bloque.json`, de `boe`, y `municipio.json`, de `territorio`) coinciden con lo que emite `--describe` de cada verbo, sin escribir nada | sí |
| `make skills-check` | Comprueba sin red, sin modelo y sin escribir nada el frontmatter y el límite de líneas de cada `SKILL.md`; las derivas de las referencias, de la tabla de comandos y de los enlaces de `scripts/`; la tabla de normas contra su esquema y sus identificadores; los ficheros congelados de `data/territorio/` contra sus esquemas y su integridad; la jerarquía normativa contra su esquema; y el formato y el conjunto de las evals y lo grabado que necesitan | sí |
| `make secrets` | Detección de secretos en todo el árbol | sí |
| `make mod-verify` | Integridad del módulo raíz y de cada módulo de herramienta | sí |
| `make mod-tidy-check` | Comprueba que `go.mod` y `go.sum` están saneados | sí |
| `make fmt` | **Corrige** el formato; por eso no forma parte de `ci` | no |
| `make skills-sync` | **Regenera** las referencias, la tabla de comandos de `SKILL.md` y los enlaces de `scripts/` de cada skill: escribe en el árbol, y por eso no forma parte de `ci` | no |
| `make lint-fast` | Análisis estático rápido, el del gancho de pre-commit | no |
| `make test-e2e` | Tests de extremo a extremo con `testscript`, contra un binario que registra los applets de ejemplo, `boe`, que responde desde sus grabaciones sin red, y `territorio` | no |
| `make verify-sources` | Comprueba contra la fuente real que sus respuestas se siguen interpretando —hoy, `boe articulo` contra la API del BOE—: **requiere red** y lo ejecuta el flujo nocturno | no |
| `make evals` | Ejecuta las evals de una skill (`SKILL=<skill>`) en sesiones con modelo de Claude Code; lo lanza el job de evals | no |

`make verify-sources` es el único control que pide algo a una fuente real, y por eso no está en `ci`:
todos los tests de `make ci` corren sin red, contra respuestas grabadas. Cada noche lo ejecuta el
trabajo `fuentes` del flujo `nightly`, que abre o comenta una incidencia si falla; el detalle está en
[`CONTRIBUTING.md`](CONTRIBUTING.md). `make evals` tampoco está en `ci`, aunque no pide nada a ninguna
fuente: sus sesiones usan un modelo, que cuesta y no es determinista ([Job de evals](#job-de-evals)).

`make help` —el objetivo por defecto— enumera todas las órdenes, incluida la que existe pero recibe su
contenido en un hito posterior (`release`).

### Ganchos de pre-commit

```bash
make hooks     # instala los ganchos con lefthook
```

Una vez instalados, cada commit corrige el formato —y vuelve a preparar lo que haya corregido, de
modo que se confirma ya formateado— y ejecuta `lint-fast`, `secrets` y `mod-tidy-check`. El gancho
**no es la autoridad final**: la integración continua vuelve a ejecutar
los mismos controles, y `make ci` es lo que hay que ver en verde antes de abrir una propuesta de
cambio.

## Cómo contribuir

El ritual por hito, el catálogo completo de controles, la convención de commits y las reglas sobre
dependencias nuevas están en [`CONTRIBUTING.md`](CONTRIBUTING.md). Las decisiones de arquitectura ya
cerradas están registradas en [`docs/ADR/`](docs/ADR/).

## Licencia

Apache-2.0. Ver [`LICENSE`](LICENSE).
