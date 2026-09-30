# Cómo contribuir a kitlegal

Esta guía es el contrato de trabajo del repositorio: cómo se organiza un hito, qué forma tiene una
propuesta de cambio, qué controles hay que ver en verde antes de abrirla y qué se espera de quien
añade una dependencia. El [`README.md`](README.md) es para quien usa kitlegal y no dice nada de esto; las decisiones
ya cerradas están en [`docs/ADR/`](docs/ADR/).

## Trabajar desde el clon

Usar kitlegal no necesita nada de lo que sigue: basta la instalación del README. Para trabajar en el propio kitlegal
hacen falta dos cosas que hay que instalar y, para `make test`, las utilidades de sistema que ya traen macOS y
cualquier Linux, y nada más:

| Prerrequisito | Comprobación |
|---|---|
| Go 1.21 o superior | `go version` |
| `git` | `git --version` |
| `sh` con las utilidades POSIX, `curl`, `tar`, `mktemp` y `sha256sum` o `shasum`: con ellas los tests ejecutan `scripts/install.sh` contra un origen local, sin red | `command -v sh curl tar mktemp` |

**No hay que instalar ninguna herramienta de control.** `golangci-lint`, `govulncheck`, `gitleaks`,
`lefthook` y `goreleaser` se construyen solos, con la versión fijada en `tools/<herramienta>/go.mod`, la
primera vez que se invoca la orden que los usa.

**Tampoco hay que instalar un parche concreto de Go, ni importa cuál tengas.** `go.mod` declara la
directiva `toolchain` y el `Makefile` exporta `GOTOOLCHAIN` con ese valor, así que todas las órdenes
se ejecutan con ese parche exacto —el mismo que ejecuta la integración continua— y el go command lo
descarga y lo verifica solo si falta. Cualquier `go` ≥ 1.21 sirve.

```bash
git clone https://github.com/jmorenobl/kitlegal.git
cd kitlegal
make check-tools     # comprueba go, git y que el toolchain fijado es obtenible
make build           # deja el ejecutable en bin/kitlegal
make install         # bucle de desarrollo: go install del binario y, con él, kitlegal skills install -g --host claude
```

> La **primera** ejecución compila las herramientas desde fuente y, si el parche fijado no está en la
> caché, lo descarga: requiere red y tarda varios minutos. Las siguientes las sirve la caché de
> construcción de Go en segundos.

`make build` y `make install` inyectan los mismos datos de construcción —versión, commit y fecha— y compilan sin cgo y
con `-trimpath`. Qué instala `make install`, y dónde, está en
[`make install` y los enlaces del anterior](#make-install-y-los-enlaces-del-anterior).

### Tres directorios llamados `skills`

| Directorio | Qué es | ¿Es kitlegal? |
|---|---|---|
| `skills/` | **El producto que se distribuye**: las skills de kitlegal, hoy `boe-legislacion` y `legal-core`, empotradas en el binario, que las instala con `kitlegal skills install` | sí |
| `.agents/skills/` | Skills de agente vendorizadas para trabajar en este repositorio: las de Go de `samber/cc-skills-golang`, registradas con su origen y su huella en el registro de bloqueo `skills-lock.json`, y las `speckit-*` que genera la integración `agy` de spec-kit para Antigravity (registradas en `.specify/integrations/agy.manifest.json`). Se versionan tal cual y no se editan; `.agents/.gitattributes` las marca como vendorizadas y generadas, para que no cuenten en las estadísticas de lenguaje del repositorio ni se desplieguen en los diffs de las propuestas de cambio | no |
| `.claude/skills/` | Lo que carga Claude Code al trabajar en el repositorio: un enlace a cada skill de `.agents/skills/` más las skills de spec-kit, con las que se prepara cada hito | no |

En el proyecto de quien usa kitlegal, `.agents/skills/` es donde `kitlegal skills install` deja las skills. En este
repositorio ese directorio es el de las vendorizadas, y el bucle de desarrollo no instala en el clon sino en la cuenta.

### `kitlegal version`

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
construida y la fecha es el instante de construcción en UTC. Un binario de la release imprime su
etiqueta tal cual (`kitlegal v0.1.0`). Un binario hecho con `make build` o `make install` nunca
imprime los valores por defecto del código (`dev`, `none`, `unknown`); si los ves, lo estás
ejecutando con `go run` o lo construyó un `go install` sin inyecciones, que instala las mismas skills
pero es un binario de desarrollo: no da el aviso de versión.

Cualquier otra invocación —sin applet, con un applet desconocido o con un argumento sobrante tras
`version`— escribe el fallo en la salida de error, nombrando lo que no ha reconocido y enumerando los
applets que existen, y termina con código `2`, que es «args» en la tabla de códigos de salida
estables del proyecto:

```console
$ ./bin/kitlegal inventado
argumentos inválidos: "inventado" no es ningún applet de kitlegal; applets disponibles: boe, graph, skills, territorio; la versión, con «kitlegal version»
$ echo $?
2
```

`kitlegal --help` describe el uso y enumera los applets registrados, con código `0`.

## El ritual por hito

El trabajo avanza **hito a hito**, en el orden de [`docs/ROADMAP.md`](docs/ROADMAP.md). Un hito es una
rama, una propuesta de cambio y un *squash-merge* con la integración continua en verde.

1. **Rama** `hNN-nombre-corto` desde `main`. Se escriben primero las evals de la skill que el hito
   entrega o mejora (desde H5, que aporta su formato) y el test de extremo a extremo (`testscript`) que
   describe la entrega. En H0 todavía no existe el ejecutor de esos tests —lo aporta H1—, así que el
   hito arranca por los tests unitarios del código que introduce.
2. **Planificar de fuera adentro** (qué debe resolver la skill → qué herramientas necesita) e
   **implementar de dentro afuera**: `core` → adaptador → applet → skill. Una herramienta que ninguna
   skill usa no se construye, y nada se particulariza para un municipio (constitución, principios VIII
   y IX).
3. **`make ci` en verde en local** y propuesta de cambio con la estructura de la sección siguiente.
4. **Revisión** de código y de seguridad sobre la propuesta. Si el hito toca una fuente externa, su fila
   de [`docs/SOURCES.md`](docs/SOURCES.md) y su caso de `make verify-sources` entran en el mismo cambio; si la
   fuente no se consulta en red y sus datos entran congelados en `data/`, como los de `data/territorio/` (ADR 0017),
   solo su fila, con la fecha del fichero generado.
5. **Squash-merge**. Si el hito cierra una fase, etiqueta y release ([La release](#la-release)).
6. **Actualizar el roadmap solo si cambia el orden o el alcance**; el detalle vive en las propuestas de
   cambio y en los ADR.

Un hito no se cierra sin la *Definition of Done* completa (`docs/ROADMAP.md` §1). La fusión a `main` y
el release son siempre acciones humanas.

Los hitos se preparan con el workflow `hito` de spec-kit (`scripts/hito.sh H<n>`, documentado en
[`docs/WORKFLOW.md`](docs/WORKFLOW.md)), que deja sus artefactos en `specs/NNN-hN-slug/`. La persona está en
los extremos (ADR 0018): escribe la sección del hito en `docs/ROADMAP.md`, que es la entrada del run, y lee
el informe final, que llega como cuerpo de la propuesta de cambio; entre medias el workflow no pausa. Los
principios que rigen las decisiones automáticas están en
[`.specify/memory/constitution.md`](.specify/memory/constitution.md).

## Estructura de la propuesta de cambio

**No hay fichero de plantilla**: la estructura se copia a mano en la descripción. Cinco apartados, en
este orden y ninguno vacío:

| Apartado | Qué contiene |
|---|---|
| **Objetivo** | Qué entrega el cambio, en una o dos frases. Si es un hito, el nombre del hito |
| **Alcance** | Qué ficheros y qué comportamiento toca; y qué queda deliberadamente fuera |
| **Controles añadidos** | Qué control nuevo entra o qué control existente pasa a cubrir algo más. «Ninguno» es una respuesta válida y hay que escribirla |
| **Decisiones** | Cada decisión tomada sin instrucción previa, con la alternativa rechazada y por qué. Aquí va también la justificación de toda dependencia nueva |
| **Pendientes** | Lo que queda abierto: lo que no se ha implementado por no estar especificado, lo que hay que verificar en un hito posterior y cualquier deuda asumida |

Un cambio que altera una decisión de arquitectura añade además un ADR en `docs/ADR/` (formato MADR
corto). Un cambio de comportamiento visible añade su entrada en `CHANGELOG.md`.

## Commits

**Conventional Commits**, sin excepciones:

```text
<tipo>(<ámbito>): <resumen en imperativo y en minúscula>
```

- **Tipos**: `feat`, `fix`, `docs`, `refactor`, `test`, `perf`, `build`, `ci`, `chore`, `revert`.
- **Ámbito**: el hito (`H0`) mientras se implementa uno, o el componente al que afecta (`boe`, `cli`,
  `deps`, `graph`…).
- **Cambio incompatible**: `!` tras el ámbito (`feat(cli)!: …`) y un pie `BREAKING CHANGE: …` que
  explique la migración.
- El cuerpo del mensaje explica el **porqué**, no el qué: el qué ya está en el diff.

Ejemplos del propio repositorio: `feat(H0): T009`, `chore(deps): go1.27.1`, `docs(adr): 0002 sqlite sin cgo`.

Como los hitos se integran con *squash-merge*, el mensaje que acaba en `main` es el título de la
propuesta de cambio: también tiene que cumplir la convención.

## Versionado y `CHANGELOG.md`

**Versionado semántico.** El proyecto está en `0.y.z` hasta la primera release, que es H19 (`v0.1.0`);
mientras el mayor sea `0`, un cambio incompatible sube el **menor**. A partir de `1.0.0`, mayor para lo
incompatible, menor para funcionalidad nueva compatible y parche para correcciones. Las etiquetas son
`vX.Y.Z` y las pone una persona, nunca la integración continua.

**`CHANGELOG.md`** sigue el formato *Keep a Changelog*: una sección `## [Unreleased]` siempre presente en
cabeza, y bajo ella los apartados `Añadido`, `Cambiado`, `Obsoleto`, `Eliminado`, `Corregido` y
`Seguridad`, solo los que tengan contenido. Todo cambio de comportamiento visible entra en *Unreleased*
en la misma propuesta que lo introduce; al publicar una versión, esa sección se cierra bajo su número y
su fecha y se abre una nueva vacía. El changelog se mantiene **a mano**, también desde la primera
release: las notas de cada release las genera goreleaser desde los Conventional Commits
(`changelog` de `.goreleaser.yaml`), y no sustituyen a este fichero.

## Los controles

`make ci` es **el veredicto del repositorio**: si está en verde en local, la propuesta de cambio pasa,
porque la integración continua ejecuta esa misma orden y no aplica ningún control por otra vía; lo único
que añade es el trabajo `snapshot`, que ejecuta otras dos órdenes del mismo `Makefile`, `make release` y
`make snapshot-check` ([La release](#la-release)). `make ci` no modifica ningún fichero versionado, así
que se puede ejecutar con el árbol sucio sin miedo.

| Control | Orden | ¿En `make ci`? |
|---|---|---|
| Formato (`gofumpt` + `goimports`), modo verificación | `make fmt-check` | sí |
| Formato, modo corrección | `make fmt` | no — **corrige**, y corregir el árbol para aprobarlo no es un gate |
| Análisis estático (`golangci-lint`, `gosec` y `govet` incluidos) | `make lint` | sí |
| Análisis estático rápido | `make lint-fast` | no — es el del gancho de pre-commit |
| Tests unitarios con detector de carreras y perfil de cobertura | `make test` | sí |
| Tests con la etiqueta `integration` (dependen del entorno: permisos, dos procesos, la instalación de las skills con `make install` en un directorio personal temporal, la matriz de `world.db` de `internal/graph`) | `make test-integration` | sí |
| Medidas con el reloj de pared, las dos de `MEDIDAS_DE_TIEMPO`: las cotas de tiempo de los guiones e2e (`TestMedidasDeTiempo`: `boe articulo` desde la caché y `territorio resolver` por debajo de 200 ms, ahora entregando al grafo del mundo) y el coste del grafo (`TestCosteDelGrafo`: con un `world.db` que ya existe, la entrega no añade más de 150 ms a la mediana de 20 `boe articulo` desde la caché respecto de los mismos con `--no-graph`, y, en medianas de cinco sobre 10 000 nodos y 10 000 aristas `graph check` tarda menos de 3 s y `graph stats` menos de 1 s). Solas y sin la caché de resultados de `go test`, después de las dos anteriores, que las saltan: medir con el reloj mientras corren todos los paquetes mide la carga de la máquina, no el programa | `make test-tiempos` | sí |
| Vulnerabilidades conocidas (`govulncheck`) | `make vuln` | sí |
| Esquemas publicados en `schemas/` iguales a lo que emite `--describe` de cada verbo, sin escribir nada | `make schema-check` | sí |
| Skills, datos y evals, sin red, sin modelo y sin escribir nada: frontmatter y límite de líneas de cada `SKILL.md`; ninguna skill con `scripts/`; derivas de las referencias y de la tabla de comandos; cada orden de la tabla de comandos de cada skill empotrada nombra un applet y un verbo del binario; tabla de normas contra su esquema y sus identificadores; ficheros congelados de `data/territorio/` contra sus esquemas y su integridad; jerarquía normativa contra su esquema; formato y conjunto de evals, lo grabado que necesitan y el grafo previo que nombran | `make skills-check` | sí |
| Regeneración de lo que se deriva de cada skill (referencias y tabla de comandos de `SKILL.md`); una skill con `scripts/` la hace fallar | `make skills-sync` | no — escribe en el árbol |
| Configuración de la release válida (`goreleaser check`, sin construir nada; una propiedad obsoleta en la versión fijada falla) | `make goreleaser-check` | sí |
| Detección de secretos (`gitleaks`) | `make secrets` | sí |
| Integridad de los módulos (`go mod verify`, raíz y herramientas) | `make mod-verify` | sí |
| Dependencias saneadas (`go mod tidy -diff`) | `make mod-tidy-check` | sí |
| Prerrequisitos (`go`, `git`, toolchain fijado obtenible) | `make check-tools` | sí, como dependencia de las demás |
| Verificación contra la fuente real (`scripts/verify-sources.sh`; requiere red) | `make verify-sources` | no — toca la red; lo ejecuta el trabajo `fuentes` del flujo nocturno, que abre o comenta una incidencia si falla |
| Evals de una skill con Claude Code (`scripts/evals.sh`; Linux con `strace`, como root o con `sudo`) | `make evals` | no — sesiones con modelo y credencial, fuera de `make ci`; las lanza el job de evals |
| Snapshot de la release en `dist/` (`goreleaser release --snapshot --clean --skip=publish,sign,sbom`): seis archivos, los cuatro paquetes `.deb` y `.rpm` y `checksums.txt`, que lista también `install.sh`, sin publicar, firmar ni SBOM | `make release` | no — construye seis plataformas; lo ejecuta el trabajo `snapshot` de CI |
| Comprobación del snapshot (`TestSnapshot`) y guiones `instalador-` de `scripts/install.sh` contra él, sin red | `make snapshot-check` | no — necesita el `dist/` de `make release`; lo ejecuta el trabajo `snapshot` de CI |
| La web: tipos, cada cita contra su sobre y construcción en `web/dist` ([La web](#la-web)) | `make web` | no — necesita Node y pnpm; lo ejecuta el flujo `web` |
| La web en local, con recarga al guardar | `make web-dev` | no — servidor de desarrollo |
| Sobres de las citas de la web, regenerados con el binario de `make build` | `make web-citas` | no — toca la red y escribe en el árbol |
| Bucle de desarrollo: `go install` del binario y, con él, `kitlegal skills install -g --host claude` | `make install` | no — instala en la cuenta ([`make install` y los enlaces del anterior](#make-install-y-los-enlaces-del-anterior)) |
| Cobertura: global ≥ 70 % y `internal/core/**` ≥ 85 % | `make test` genera el perfil; el umbral lo aplica Codecov sobre la propuesta | no como orden |
| Análisis de seguridad semanal (CodeQL) | — (flujo `.github/workflows/codeql.yml`) | no |
| Actualización semanal de dependencias | — (Dependabot, `.github/dependabot.yml`) | no |
| Ganchos de pre-commit | `make hooks` los instala | no |

`make help` —el objetivo por defecto— enumera todas las órdenes.

### El gancho de pre-commit no es la autoridad final

`make hooks` instala los ganchos: cada `git commit` corrige el formato y vuelve a preparar lo corregido,
y ejecuta `lint-fast`, `secrets` y `mod-tidy-check`. Es una red para no confirmar lo obvio, **no un
veredicto**: ejecuta un subconjunto de los controles y el rápido en lugar del completo. La autoridad
final es la integración continua, y lo que hay que ver en verde antes de abrir una propuesta de cambio
es `make ci`. Saltarse el gancho con `--no-verify` no adelanta nada: la propuesta ejecuta los controles
enteros de todos modos.

## Dependencias nuevas

**Toda dependencia nueva exige justificación explícita.** No basta con que compile y con que el árbol
quede saneado.

Las dependencias permitidas están enumeradas en la constitución (§V): `alecthomas/kong`,
`modernc.org/sqlite`, `stretchr/testify`, `rogpeppe/go-internal`, `golang.org/x/time/rate`,
`temoto/robotstxt`, `gopkg.in/yaml.v3`, `invopop/jsonschema`, `santhosh-tekuri/jsonschema` y, en la fase de
distribución, `modelcontextprotocol/go-sdk`. **Cualquier otra** se justifica en dos sitios: en el `plan.md` del hito
(sección *Complexity Tracking*) y en el apartado **Decisiones** de la propuesta de cambio.

La justificación responde a cuatro preguntas, y ninguna se responde con «es lo estándar»:

1. **Qué problema resuelve** y por qué no lo resuelve la biblioteca estándar. Lo que no pide un hito no
   se construye, y lo que no hace falta no se importa.
2. **Qué arrastra**: dependencias transitivas, tamaño y si toca cgo (el binario se compila con
   `CGO_ENABLED=0`).
3. **Licencia y mantenimiento**: licencia compatible, actividad reciente, versión estable.
4. **Cómo se sale**: qué costaría sustituirla si deja de mantenerse.

Sin framework de inyección de dependencias, sin ORM y sin generador de CLI: esas tres puertas están
cerradas por la constitución.

Lo mecánico lo cubren los controles —`make mod-tidy-check` falla si el fichero de dependencias no está
saneado, `make mod-verify` comprueba la integridad de los módulos y Dependabot propone las
actualizaciones semanales—, pero ninguno de ellos juzga si la dependencia debía entrar. Eso lo juzga la
revisión, y para eso necesita la justificación escrita.

Las herramientas de los controles no son dependencias del producto: viven en su propio módulo
(`tools/<herramienta>/go.mod`) precisamente para que el `go.mod` de `kitlegal` no las arrastre.

## La release

Desde H19 ninguna orden del `Makefile` espera su contenido de un hito posterior: la última, `make release`,
construye ya la release. La define `.goreleaser.yaml` y la construye goreleaser, un módulo de herramienta
más (`tools/goreleaser/go.mod`), que se invoca como los demás y que `make mod-verify` y Dependabot cubren.
Tres órdenes:

- **`make goreleaser-check`**, dentro de `make ci`: `goreleaser check` valida la configuración sin construir
  nada, y falla si usa una propiedad que la versión fijada de goreleaser declara obsoleta.
- **`make release`** construye el snapshot en `dist/` con `goreleaser release --snapshot --clean
  --skip=publish,sign,sbom`: los seis archivos (`kitlegal_{darwin,linux}_{amd64,arm64}.tar.gz` y
  `kitlegal_windows_{amd64,arm64}.zip`), `checksums.txt` y los paquetes `.deb` y `.rpm`, sin publicar, sin firmar y
  sin SBOM, así que no necesita syft, cosign ni ningún secreto. Es la única definición del snapshot. `dist/` está
  en `.gitignore`.
- **`make snapshot-check`**, sobre el `dist/` de `make release`: `TestSnapshot` comprueba que están los seis
  archivos con su huella en `checksums.txt`, que no hay SBOM ni firmas y que el binario de la plataforma que lo
  ejecuta imprime en `version` la versión y el commit del snapshot; y los guiones `instalador-` del e2e ejecutan
  `scripts/install.sh` contra ese `dist/`, sin red. Sin ningún guion `instalador-` que ejecutar, falla en lugar de
  pasar en vacío.

Las dos últimas no están en `make ci` porque construyen seis plataformas; las ejecuta, en cada propuesta de
cambio y en cada push a `main`, el trabajo `snapshot` del flujo `ci`, sin `id-token` y sin ningún secreto.

**Publicar es humano.** Tras fusionar, una persona empuja la etiqueta `vX.Y.Z`, y solo eso dispara el flujo
`release` (`.github/workflows/release.yml`). Su trabajo `publicar` construye con el mismo goreleaser, genera los
SBOM con syft, firma `checksums.txt` con cosign sin clave, publica la release con `install.sh` adjunto, el cask
del tap `jmorenobl/homebrew-tap` y el manifiesto del bucket `jmorenobl/scoop-bucket`, y atesta la procedencia de
los seis archivos y de `checksums.txt`. Su trabajo `humo` comprueba lo publicado como lo recibe quien lo instala,
sin el código del repositorio: la huella del archivo linux/amd64, `gh attestation verify` contra el repositorio,
que `kitlegal version` imprime la etiqueta, que `kitlegal boe articulo BOE-A-2015-10565 a21 --offline` con una
caché vacía sale con `4` y que `kitlegal skills install` en un directorio vacío deja
`.agents/skills/boe-legislacion/SKILL.md`. El único secreto de la publicación es `PUBLISHER_TOKEN`, con permiso
de escritura en el tap y en el bucket, y solo lo ve el paso que publica. Los dos repositorios ya existen, públicos y
vacíos; antes de la primera etiqueta, una persona da de alta el secreto y hace público este, desde cuya rama `main`
se sirve `install.sh` y sin el cual no se puede verificar la atestación. Al etiquetar, la sección *Unreleased* de `CHANGELOG.md` se cierra bajo su número y su fecha
([Versionado y `CHANGELOG.md`](#versionado-y-changelogmd)).

## `make schema-check` y `make verify-sources`

`make schema-check` regenera en memoria, desde `--describe` de cada verbo que registra el binario
distribuido, los esquemas publicados en `schemas/` —hoy `norma.json` y `bloque.json`, los de `boe`,
`municipio.json`, el de `territorio`, `instalacion.json`, el de `skills`, y `grafo.json`, el de `graph`— y los
compara con los ficheros versionados sin escribir nada. Si falla, nombra el fichero y el verbo: la salida
de ese verbo ha cambiado y el contrato publicado no. Eso es un cambio de contrato, así que los ficheros se
regeneran a propósito, con la bandera del mismo test, y el diff se revisa en la propuesta de cambio:

```bash
go test -count=1 -run '^TestEsquemasPublicados$' ./internal/app/ -args -actualizar-esquemas
```

`make verify-sources` es el **único control que pide algo a una fuente real**: ejecuta
`scripts/verify-sources.sh`, que comprueba que las respuestas de cada fuente se siguen interpretando —hoy,
`boe articulo BOE-A-2015-10565 a21` contra la API del BOE: código `0`, sobre válido contra su esquema y
texto no vacío—. Por eso **necesita red y no está en `make ci`**, cuyos tests corren siempre sin red,
contra respuestas grabadas. Lo ejecuta cada noche el trabajo `fuentes` del flujo `nightly`, que, si falla,
comenta la incidencia abierta con el título del caso o la abre. Un hito que añade una fuente añade su caso
a esta verificación (*Definition of Done*, punto 8), salvo que la fuente no se consulte en red y sus datos entren
congelados en `data/` (ADR 0017): no hay respuesta que verificar, y lleva solo su fila de `docs/SOURCES.md`. Ningún control ni flujo graba respuestas: las
grabaciones contra las que corren los tests las hace una persona con `scripts/grabar-fixtures.sh`.

## El grafo del mundo (`internal/graph`)

Desde H7 el binario recuerda lo que observa (ADR 0014) en `world.db`, una base SQLite en la carpeta de la caché y con
su misma regla de ubicación (`cache.Directorio`: `KITLEGAL_CACHE_DIR` o `~/.cache/kitlegal`). Quien usa kitlegal lo lee
con el applet `graph` —`show`, `stats` y `check`, que no escriben nada—, y lo escribe el kernel, nunca un applet:

- **Un applet declara lo que observa en `Resultado.Grafo`** (`schema.Observado`): la vigencia de la consulta y sus
  operaciones, `schema.Nodo`, `schema.Arista` y `schema.Texto`, sin fuente, url ni fecha. El kernel las entrega
  después de presentar la salida, y solo si la invocación termina con `0`, al `core.GraphStore` del registro
  (`Registro.EntregarAlGrafo`), con la `fuente`, la `url` y la `fecha_consulta` del sobre presentado; con `--no-graph`
  entrega a `graph.Nulo`, que no abre nada. Si la entrega falla, la invocación escribe una línea de aviso en la salida
  de error y conserva su salida y su código. Un applet que no observa identidades del mundo deja el campo vacío y no
  implementa nada; uno nuevo que las observa nace emitiendo. Hoy emiten `boe articulo`, `boe articulos` y
  `territorio resolver`, con los ids naturales (ELI, `ine:<código>`, DIR3) y los tipos y relaciones de
  `internal/core/grafo/vocabulario.go`.
- **La lógica es dominio**, `internal/core/grafo`: la validación del lote —con el rechazo de un documento de
  identidad en un nodo `Persona`—, la fusión de observaciones, las reglas de `check` y sus explicaciones, sin E/S y
  dentro del umbral de cobertura de `internal/core/**`. **El almacén es un adaptador**, `internal/graph`: SQLite con
  sus migraciones embebidas (`internal/graph/migraciones/`), modo WAL, esperas en tramos de 100 ms que miran el
  contexto, una lectura que sin `world.db-wal` no cambia ni un byte y una primera escritura que crea `world.db` en su
  sitio, sin temporal ni enlace. Un `world.db` que el binario no puede usar sigue la regla genérica (constitución,
  «Gates»): código 1 con la ruta y la causa, sin caso ni test propios más allá de un fichero que no es una base
  SQLite. La regla R6 —lista `grafo` de `depguard` en `.golangci.yml` y `TestArquitectura`— le impide importar
  `internal/source` e `internal/render`; con `internal/cache`, es el único paquete que importa `database/sql` y SQLite
  (R3), y su superficie exportada no nombra ninguno de los dos (`TestSuperficieExportada`).

Sus tests, y la orden que los ejecuta:

| Qué comprueban | Dónde | Orden |
|---|---|---|
| Validación, `Persona`, fusión, lecturas, reglas, orden y cota de `check` y explicaciones | `internal/core/grafo/*_test.go` | `make test` |
| Ruta, errores, esperas, migraciones, modo de apertura según haya o no `world.db-wal`, lectura con ámbito y sin él, escritura con sus lecturas y creación en su sitio, sobre `t.TempDir()` | `internal/graph/*_test.go` | `make test` |
| La matriz por la API pública: esquema, idempotencia, orden de llegada, rechazos, ocho entregas a la vez, un `world.db` que no es una base de datos y uno de una versión posterior, el plazo y el bloqueo, el `-wal` de una escritura propia interrumpida, un `world.db` escrito por H7 y sin residuos | `internal/graph/integracion_test.go` (`//go:build integration`) | `make test-integration` |
| El applet, su salida contra `schemas/grafo.json`, ningún texto legal en su salida, la salida de `boe` igual con grafo y sin él, y la procedencia de cada operación | `internal/app/grafo_test.go` | `make test` |
| La salida legible de `stats`, `show` y `check` sin `--json` | `internal/app/grafo_legible_test.go` | `make test` |
| La medida de `graph check` sobre el grafo sembrado de la bitácora (`TestMedidaDelGrafo`: 50 hallazgos en 40 000 bytes como mucho sin argumentos, y solo los del ámbito con la norma o con la norma y un bloque) y lo que lee la skill (`TestLoQueLeeLaSkill`: cinco `version-obsoleta` en 3 800 bytes como mucho) | `internal/app/medida_test.go` (`//go:build integration`) | `make test-integration` |
| El coste de la entrega y de `graph check` y `graph stats` sobre un grafo grande; las medianas medidas salen en el mensaje de toda cota incumplida y con `go test -v -count=1 -run '^TestCosteDelGrafo$' ./internal/app/` | `internal/app/coste_test.go` (`TestCosteDelGrafo`) | `make test-tiempos` |
| Los guiones de extremo a extremo, con tres binarios de reloj fijo (`KITLEGAL_T0_BIN`, `KITLEGAL_T1_BIN` y `KITLEGAL_T8_BIN`) y las respuestas derivadas de `internal/app/testdata/derivadas/` | `internal/app/testdata/script/` | `make test-e2e`, y `make test` con todo lo demás |

**Ningún test escribe en `~/.cache/kitlegal`**: el que entrega lo hace a `graph.ConDirectorio(t.TempDir())` o monta un
registro sin almacén, y un guion usa el `KITLEGAL_CACHE_DIR` del arnés. De los bytes de `world.db` y de sus auxiliares
solo se promete lo que dice [CHANGELOG.md](CHANGELOG.md), y solo de lo que deja el propio binario: leer sin `-wal` no
cambia ni un byte (`TestLeerSinRastro`), leer junto al `-wal` de una escritura propia interrumpida deja `world.db` y el
`-wal` como estaban (`TestIntegracionLecturaConWAL`), y una entrega junto a ese `-wal` lo recupera
(`TestIntegracionRecuperacionDeclarada`). Un estado al que el binario no llega —un enlace, permisos cambiados, un
diario de rollback, una base de otra aplicación— no lleva caso ni test propios: lo cubre la regla genérica.

## Skills y evals

Una skill es un directorio sin código bajo `skills/`: `SKILL.md` y `references/`, sin `scripts/` (ADR 0019): cada
orden de la skill invoca `kitlegal <applet> <verbo> …` desde el `PATH`, y el binario la lleva empotrada y la instala con
`kitlegal skills install`. Qué son los tres directorios llamados `skills` está en
[Tres directorios llamados `skills`](#tres-directorios-llamados-skills) y cómo se instala, en el
[`README.md`](README.md#instalar); esta sección es lo que hace falta para cambiar una skill, sus datos o sus evals.

**Lo generado no se edita.** `references/*.md` y la tabla de comandos de `SKILL.md` —entre sus marcas— se derivan de
`data/*.yaml` y de `--describe` del binario. Tras cambiar `data/`, añadir un verbo o cambiar su entrada o su salida,
se ejecuta `make skills-sync` y lo regenerado va en el mismo cambio: `make skills-check`, dentro de `make ci`, lo
regenera en memoria y falla nombrando la skill y el fichero que difieren. Comprueba además el frontmatter de cada
`SKILL.md` y que tenga menos de 300 líneas, que ninguna skill tiene `scripts/` —una entrada `skills/<skill>/scripts`
hace fallar también `make skills-sync`, que no la retira—, que cada orden de la tabla de comandos de cada skill
empotrada nombra un applet y un verbo que el binario registra, la tabla de normas contra `schemas/normas.yaml.json`,
que cada identificador está en la búsqueda grabada del BOE, la jerarquía normativa de `data/jerarquia.yaml` contra
`schemas/jerarquia.yaml.json`, los ficheros congelados de `data/territorio/` contra sus esquemas y su integridad, y el
formato y el conjunto de las evals, que lo que necesitan está grabado y que el grafo previo que nombran se prepara.
Una norma nueva, o una eval que consulta algo que no está grabado, llega con su grabación, que hace una persona con
`scripts/grabar-evals.sh`: ningún control ni flujo graba respuestas.

### `make install` y los enlaces del anterior

`make install` es el bucle de desarrollo, no la forma de instalar kitlegal: `CGO_ENABLED=0 go install -trimpath` del
binario con los datos de construcción del `Makefile` y, con ese binario —por la ruta que da
`go list -f '{{.Target}}' ./cmd/kitlegal`, no por el `PATH`, donde podría ir antes otro `kitlegal`—,
`kitlegal skills install -g --host claude`. Deja las skills del árbol en `~/.agents/skills/`, con su manifiesto
`kitlegal.json`, y enlazadas en `~/.claude/skills/<skill>` con destino `../../.agents/skills/<skill>` y, si existe
`~/.gemini/config/`, en `~/.gemini/config/skills/<skill>` con destino `../../../.agents/skills/<skill>`, donde las ve
Antigravity (ADR 0025); tras cambiar una skill, repetirla la deja `actualizada`. Las skills invocan `kitlegal` desde el `PATH`, así que el directorio de
binarios de Go (`$GOBIN` o, sin él, `$GOPATH/bin`) tiene que estar en él. No crea `bin/instalado/` ni comprueba nada
antes del `go install`: ante un conflicto, `skills install` sale con `7` sin cambiar nada (ADR 0023), y `make install` falla con
él —`make` termina con su propio código, `2`—, con el binario ya instalado. `make test-integration` lo prueba (`TestInstalacion`) sobre una copia mínima del árbol,
con `HOME`, `GOBIN` y `GOPATH` temporales y sin red.

El `make install` anterior a H19 dejaba `~/.claude/skills/boe-legislacion` y `~/.claude/skills/legal-core` como
enlaces absolutos a `skills/<skill>` del clon, y `bin/instalado/kitlegal` apuntando al binario instalado. El nuevo
nombra esos enlaces como «enlace a otro sitio» y falla sin tocar `~/.agents` ni `~/.claude`. Se retiran una sola vez,
desde la raíz del clon:

```sh
for s in boe-legislacion legal-core; do
  if [ "$(readlink "$HOME/.claude/skills/$s")" = "$(pwd -P)/skills/$s" ]; then rm -- "$HOME/.claude/skills/$s"; fi
done
rm -r -- bin/instalado
make install
```

Solo se retira un enlace cuyo destino literal es la skill de este clon —`pwd -P`, porque el guion anterior enlazaba la
ruta física—; cualquier otra entrada con ese nombre queda como está, y `make install` la sigue nombrando como
conflicto hasta que se decida qué hacer con ella. `rm -r` sobre `bin/instalado` retira el directorio y el enlace que
contiene, no el binario al que apuntaba.

**Los datos de territorio no se regeneran con `make skills-sync`**: `data/territorio/` no deriva de nada del
repositorio, sino de descargas públicas que no se consultan en red (ADR 0017). Los refresca una persona, fuera del
repositorio, desde la relación de municipios del INE, las tablas de códigos de comunidad y provincia del INE (los
nombres) y el volcado del Registro de Entidades Locales, con la fila de cada origen en `docs/SOURCES.md` puesta a la
fecha del fichero. Un refresco de `dir3.yaml` vuelve a verificar la derivación contra DIR3 real en una muestra con
un municipio fusionado o renombrado, uno foral y uno con entidades locales menores, y la registra con el código
derivado, el real y su procedencia (ADR 0017; como en
`specs/008-h6-territorio-skill-legal/gates/verificacion-dir3.md`): sin ella, `cobertura.dir3: verificado` no sería
cierto. Añadir un territorio es rellenar `boletines` en el
fichero de su comunidad, `data/territorio/comunidades/<código>.yaml`, sin tocar código ni skills. Los ficheros viajan
dentro del binario, así que un cambio en ellos llega a `territorio resolver` al volver a construirlo.

### Formato común de eval

Las evals de una skill se escriben **antes** que la skill o que el cambio que la mejora (ritual, paso 1), en su propio
directorio, `evals/<skill>/`, con un fichero YAML por eval llamado `<nn>-<descripción>.yaml` —dos cifras y una
descripción en minúsculas con guiones— y, si la skill la tiene, su lista de expresiones prohibidas (abajo). Todas las
evals siguen el formato común de eval:

| Campo | ¿Obligatorio? | Qué fija |
|---|---|---|
| `pregunta` | sí | La pregunta con la que se abre la sesión; no vacía |
| `activa` | sí | Si la pregunta debe activar la skill |
| `comandos` | sí si `activa` es `true`; prohibido si es `false` | Las consultas que la sesión debe hacer con éxito, cada una en una de cinco formas: un bloque (`applet`, `norma`, `bloque`), una consulta de norma (`applet`, `verbo` —`indice`, `metadatos` o `analisis`—, `norma`), una búsqueda (`applet`, `verbo` `buscar`, `terminos`), un municipio (`applet`, `verbo` `resolver`, `municipio`) o una comprobación (`applet`, `verbo` `check` y, si se quiere, `norma`) |
| `citas` | sí si `activa` es `true` y no hay `territorio`; prohibido si es `false` | Cada `norma` y `bloque` que la respuesta debe citar |
| `territorio` | sí si `activa` es `true` y no hay `citas`; prohibido si es `false` | Lo que la respuesta debe declarar del territorio que devuelve `territorio resolver`, con al menos una de estas claves: `comunidad`, `provincia`, los códigos de `boletines` y los aspectos de `cobertura` en la forma `<aspecto>: <valor>` del vocabulario del applet (`boletin_autonomico: no-configurado`…) |
| `avisos` | no; solo si `activa` es `true`, prohibido si es `false` | Los códigos de aviso de vigencia del binario (`consolidacion-no-finalizada`, `derogada`, `vigencia-agotada`) cuya forma fija —`⚠`, la etiqueta del aviso y dos puntos— debe llevar la respuesta |
| `hallazgos` | no; solo si `activa` es `true`, prohibido si es `false` | Las clases de hallazgo de `graph check` cuya forma fija —`⚠`, la etiqueta que da el binario y dos puntos— debe llevar la respuesta; hoy solo `version-obsoleta` (`⚠ REDACCIÓN MODIFICADA:`), la única que el binario etiqueta |
| `prohibidos` | no; solo si `activa` es `true`, prohibido si es `false` | Los `applet` y `verbo` que la sesión no puede invocar (`graph` `show`…); al menos uno |
| `grafo_previo` | no; solo si `activa` es `true`, prohibido si es `false` | Lo que el grafo del mundo de la sesión ya ha observado al empezar: `grabaciones`, un directorio de `testdata/evals/grafo-previo/` con respuestas del BOE que se ponen encima de las grabadas, y `comandos`, los bloques (`applet`, `norma`, `bloque`) que se consultan contra ellas antes de la sesión, entregando lo observado al grafo de la sesión; la caché de la sesión se prepara después, como siempre y sin entregar nada |
| `informativa` | no | Con `true`, la eval se ejecuta solo con el modelo que decide y su tasa se publica, pero no decide el veredicto (ADR 0016). En `boe-legislacion`, solo en una eval que activa la skill |
| `reproduce` | no | La skill cuyo uso documentado reproduce la eval (p. ej. `boe-fiscal`) |

Cada fichero de eval de cada directorio `evals/<skill>/`, sea de la skill que sea, se valida contra
`schemas/eval.yaml.json` dentro de `make ci`. Una entrada del directorio que no es un fichero con esa forma de nombre
ni la lista de expresiones prohibidas, una clave desconocida o repetida, un identificador o un bloque mal escritos, una
eval positiva sin citas ni territorio o una de no activación con comandos fallan nombrando el fichero; ninguna se
salta. Para `boe-legislacion`, `make ci` exige además las reglas de su conjunto: exactamente diez positivas que
deciden, de materias distintas, al menos una de no activación y al menos una informativa, y ninguna informativa de no
activación, entre otras. Para `legal-core`, al menos tres evals: una positiva que resuelve un municipio del territorio
configurado y declara sus boletines, otra que resuelve uno de una comunidad sin configuración y declara no
configurados el boletín autonómico y el provincial, al menos una de no activación, y citas o territorio en toda
positiva. El directorio de cada `grafo_previo` tiene que existir, y su preparación, en temporales, deja en el grafo un
`BloqueVersion` por comando sin ninguna falta; después, leyendo como la sesión cada bloque de los `comandos` de la eval,
`graph check` termina con `0` con exactamente las clases de `hallazgos` de la eval, y cada `version-obsoleta` con la
fecha de vigencia de la redacción que dejó el grafo previo y la de la que acaba de leer (`TestEvalsDelRepositorio`,
subprueba `grafo-previo`). Cada respuesta de un grafo previo es una derivada de la grabación de H4 que sustituye —la
misma respuesta sin sus redacciones posteriores a una fecha de vigencia, y nada más—: se declara en
`derivadasDelGrafoPrevio()` de `internal/app/grafo_test.go`, la escribe
`go test -count=1 -run '^TestGrabacionesDerivadas$' ./internal/app/ -args -actualizar-derivadas`, nunca una persona, y
`TestGrabacionesDerivadas` exige que sea, byte a byte, esa derivación y que, servida en lugar de la grabación, `boe` dé
exactamente una de las redacciones que trae la grabada, la de esa fecha.

**La lista de expresiones prohibidas** es `evals/<skill>/expresiones-prohibidas.yaml`, opcional y una por skill: las
expresiones que no lleva la respuesta de una eval que activa la skill. `make ci` la reconoce por ese nombre exacto —no
es un fichero de eval— y la valida contra su propio esquema, `schemas/expresiones-prohibidas.yaml.json`: tres claves
obligatorias, una por familia, cada una con una lista no vacía de expresiones de una o más palabras separadas por un
espacio, sin blancos en los extremos ni `*` o `_`:

| Familia | Qué recoge |
|---|---|
| `maquinaria` | Lo que la respuesta no cuenta de cómo trabaja la skill: la memoria de consultas, `kitlegal graph` y sus verbos, los códigos de salida, los hallazgos y sus clases, el JSON y el sobre, en formas que no chocan con el castellano corriente ni con el texto de las normas (`memoria de consultas`, `graph check`, `código 0`, `sobre de salida`…) |
| `otra_conversacion` | Lo que atribuye a la skill algo dicho a quien pregunta en otra conversación: un verbo de decir con «te» en pretérito o en condicional compuesto, y `conversación anterior` (`te dije`, `te habría confirmado`…) |
| `anuncio` | El anuncio de la respuesta o del estado de lo comprobado: que va a responder, que ya tiene lo que necesita o que no hay nada que trasladar (`ya puedo responder`, `tengo todo lo necesario`, `redacto la respuesta`, `que trasladar`…). Ninguna dice a quien lee que la norma no está derogada o que no tiene avisos: eso es derecho, no maquinaria |

Cada expresión se compara con la respuesta por la forma, sin ningún modelo y con la tolerancia de las formas fijas de
los avisos (H5.1): sin distinguir mayúsculas y con blancos y énfasis de Markdown de más entre las palabras y
alrededor; delimitada como palabra —`hallazgo` no se encuentra dentro de `hallazgos`—, sin plegar tildes y sin admitir
un salto de línea entre dos palabras. **Lo mismo dicho con otras palabras no se detecta**: es una limitación declarada,
y el informe publica cada respuesta para verlo. Una lista mal formada —clave desconocida o repetida, familia que falta
o vacía, expresión con blancos en un extremo o con `*` o `_`, o una entrada con ese nombre que no es un fichero— es un
fichero mal formado, como una eval: `make ci` falla nombrándola y las evals de la carpeta se leen sin lista. Hoy solo
la tiene `boe-legislacion`, y `TestEvalsDelRepositorio` comprueba además la suya: marca exactamente 35 de las 93
respuestas del informe de evals de H7.1 y 10 de las 93 del de H7.2, los informes versionados en los directorios de
esos hitos en `specs/`, eval por eval y familia por familia, y ninguna otra (subprueba `expresiones-calibradas`), así
que una expresión nueva que marca una respuesta más, o una retirada que deja de marcar una, hace fallar `make ci`
nombrando el informe, la eval y la familia; ninguna expresión casa con el texto de los bloques que leen sus evals y
sus grafos previos (`expresiones-en-los-bloques`), ni con las formas fijas de los avisos y de los hallazgos o los
bloques `text` de su `SKILL.md`, que son lo que la skill enseña a escribir (`expresiones-de-la-skill`); y **la prosa
de su `SKILL.md` no enseña lo que la respuesta no puede decir** (`prosa-de-la-skill`): sobre el fichero entero,
frontmatter incluido, sin los bloques delimitados, los tramos de código en línea ni la región generada de la tabla de
comandos, y párrafo a párrafo —partido por las líneas en blanco y por cada elemento de lista—, ninguna expresión de la
lista, y en todo el fichero ninguna fecha `AAAAMMDD` escrita con cifras, que un ejemplo daría a copiar. Falla
nombrando la primera línea de cada párrafo que lleva alguna, con sus expresiones, y cada línea con una fecha. Lo que la
skill ejecuta y lo que lee en la salida del binario va en código (`version-obsoleta`, `data.hallazgos`, `3`), y no
cuenta; los ejemplos de una forma que lleva datos del binario usan marcadores como `AAAAMMDD`.

Una sesión de una eval pasa si la abre el modelo pedido, activa la skill cuando debe y no la activa cuando no debe,
termina, hace con éxito cada consulta de `comandos` —una comprobación la cumple una invocación de ese applet con
`check`, y con su `norma` si la lleva, que termina con `0`—, no invoca nada de `prohibidos` —cuenta toda invocación
que consulta de ese applet y verbo, termine como termine; la ayuda, `--describe` y `--dry-run`, no— y responde
citando cada `norma` y `bloque` de `citas`, declarando lo que espera `territorio` —la comunidad y la provincia sin
distinguir mayúsculas ni tildes, el código de cada boletín como palabra exacta y cada aspecto de cobertura en su forma
fija `<aspecto>: <valor>`—, con la forma fija de cada aviso de `avisos` y de cada clase de `hallazgos` y, si la eval
activa la skill y la skill tiene lista, sin ninguna expresión prohibida —cada una encontrada es un motivo
`expresión prohibida: <expresión>`—; una eval de no activación, o una de una skill sin lista, no la mira. Lo juzga el
informe sin modelo, con lo que deja la sesión: su transcript y su traza. Cada eval se abre varias veces con un mismo
modelo, y esa serie pasa si las sesiones que pasan llegan al umbral ([Job de evals](#job-de-evals)).

### Job de evals

`make evals SKILL=<skill>` ejecuta `scripts/evals.sh` y no forma parte de `make ci`: sus sesiones usan un modelo,
necesitan la credencial de Claude Code, cuestan y no son deterministas. Necesita Linux con `strace`, root o `sudo` y
ningún Python accesible. Antes de la primera sesión comprueba todo eso, las variables del job —también que
`CONCURRENCIA_DE_EVALS` es un entero mayor o igual que 1—, que ninguna eval ni la lista de expresiones prohibidas
están mal formadas, que lo que necesitan está grabado, que la skill está instalada y que `kitlegal` está en el `PATH`,
y termina con `1` si algo falla. Después, una sola orden de Go, `TestEjecucionDelJob` (etiqueta `evals`, en
`internal/evals/job_test.go`), compone el plan —cada eval, tantas veces como repeticiones, con el modelo que decide y,
si no es informativa, otras tantas con cada modelo informativo— y abre sus sesiones de Claude Code en ese orden,
**como mucho `CONCURRENCIA_DE_EVALS` a la vez**. Cada sesión se prepara justo antes de abrirla, en su propio
directorio: el de trabajo, su caché y su grafo, el estado de Claude Code (`CLAUDE_CONFIG_DIR`), con un enlace a cada
skill tal como la deja `make install`, y su temporal (`TMPDIR`); ninguna escribe en nada de otra. La abre
`scripts/evals-sesion.sh` con la orden de `claude` de siempre, bajo `strace` y con la red cerrada salvo la del modelo,
en su propio grupo de procesos, y el tope lo pone el repartidor de `internal/evals`: a los 240 s envía `TERM` al grupo
y, 10 s después, `KILL` (la sesión queda con el código 124 o 137). Ninguna sesión se reintenta ni se abre dos veces, y
`SIGINT` o `SIGTERM` cierran las abiertas y dejan la orden sin informe. Juzga cada sesión y agrupa las de cada eval
con cada modelo en una serie con su tasa, cuántas de sus sesiones pasan; sin límites de uso, el informe es el mismo
que abriéndolas una tras otra, salvo los tiempos. El informe publica la tasa de cada serie y da un veredicto global
que falla si una serie que decide —la de una eval que no es informativa con el modelo que decide— no llega al umbral,
una serie no tiene exactamente las sesiones que pide el plan, una sesión es ilegible, un fichero está mal formado, no
hay ninguna eval bien formada que juzgar o una petición llega a la red (ADR 0016); y también si un umbral que decide
no se cumple, si alguna sesión queda sin medir por un límite de uso de la cuenta o si la duración de las sesiones
pasa de su objetivo (abajo). Con `fallo`, la orden y el trabajo terminan en rojo.

Lo ejecuta el job de evals, el flujo `evals` (`.github/workflows/evals.yml`), con un trabajo por skill en la misma
ejecución —hoy `boe-legislacion` y `legal-core`—, cada uno con su informe y sin que el rojo de uno cancele el otro; el
de cada skill se llama `evals (<skill>)`, que es por donde el cierre del workflow lee su informe.
Cada trabajo instala con `make install` y añade al `PATH` el directorio donde `go install` deja el binario, de modo que
las sesiones invocan `kitlegal` igual que quien lo usa.
Fija en su definición, cada uno en su variable, el modelo que decide (`MODELO_DE_EVALS`, el del uso real de la skill), los modelos informativos
(`MODELOS_INFORMATIVOS_DE_EVALS`, separados por comas, que se publican como límite inferior sin decidir), las
repeticiones de cada eval con cada modelo (`REPETICIONES_DE_EVALS`) y el umbral de sesiones que pasan
(`UMBRAL_DE_EVALS`); y, para cada skill en la entrada `include` de su matriz, cuántas sesiones abre a la vez
(`CONCURRENCIA_DE_EVALS`) y el objetivo de duración de sus sesiones, en segundos (`OBJETIVO_DE_DURACION_DE_EVALS`; `0`
es no tener objetivo). Los modelos van por su identificador completo: cambiar de modelo es un cambio de ese fichero, y
subir o bajar la concurrencia, también, con los reintentos por límite de ritmo del informe como dato. Usa
el secreto de repositorio `CLAUDE_CODE_OAUTH_TOKEN`, el token de la suscripción de Claude que da `claude setup-token`
(el proyecto no usa una clave de API de pago por uso). Hoy fija:

| Variable | Valor |
|---|---|
| `MODELO_DE_EVALS` | `claude-sonnet-5` |
| `MODELOS_INFORMATIVOS_DE_EVALS` | `claude-haiku-4-5-20251001` |
| `REPETICIONES_DE_EVALS` | `3` |
| `UMBRAL_DE_EVALS` | `2` |
| `CONCURRENCIA_DE_EVALS` | `4` en `boe-legislacion`; `1` en `legal-core` |
| `OBJETIVO_DE_DURACION_DE_EVALS` | `900` en `boe-legislacion`; `0` en `legal-core` |

Con las diecinueve evals de `boe-legislacion`, su trabajo abre 93 sesiones, cuatro a la vez: 36 de `claude-sonnet-5`
sobre las doce que deciden, 21 sobre las siete informativas y 36 de `claude-haiku-4-5-20251001` sobre las doce que
deciden. El tope de 120 minutos de cada trabajo (`timeout-minutes`) corta un cuelgue; no es el control de la duración,
que es un umbral del informe, porque un trabajo cancelado no escribe informe. Cubre el peor caso de cada skill —todas
las sesiones de su plan, con la de la prueba de red, llegando a su tope, con su concurrencia, más la preparación de
cada una, la del runner y el informe—: `485 s + ⌈N / C⌉ × (22 s + 240 s + 10 s)`, con `N` esas sesiones y `C` su
concurrencia. `TestDefinicionDelJob` comprueba en `make ci` la definición del job —el nombre del trabajo, el grupo de
`concurrency` y `cancel-in-progress`, que no hay `concurrency` de nivel de flujo, la concurrencia y el objetivo de cada
skill y que el tope cubre su peor caso— y falla nombrando la clave, el valor encontrado y el esperado; una eval nueva
que deja el peor caso por encima del tope la hace fallar. Se lanza de tres formas:

| Lanzamiento | Sobre qué rama | Cómo |
|---|---|---|
| Manual | La que se elija | Desde la plataforma, con la entrada `prueba_de_red` si se quiere también la prueba de red |
| Apertura | La de una propuesta de cambio, antes de fusionar | Al abrirla o reabrirla, si toca lo que las evals miden: `skills/`, `evals/`, `data/`, el código del binario que las skills invocan (`cmd/`, `internal/` —también `internal/graph/`— y `skills.go`), `scripts/evals.sh`, `scripts/evals-sesion.sh`, `.github/workflows/evals.yml`, `schemas/eval.yaml.json` o el `Makefile` |
| Por etiqueta | La de cualquier propuesta de cambio, antes de fusionar | Poniendo la etiqueta `evals` en su propuesta de cambio; `evals-prueba-de-red` añade la prueba de red |

No hay ejecución programada: la semanal sobre `main`, con el modelo, la versión de Claude Code y las respuestas del
BOE fijados, no medía ningún cambio. Tampoco reacciona a cada empujón (`synchronize`): cada ejecución abre decenas de
sesiones con modelo y un hito empuja muchas veces, así que cada informe mide el commit que había cuando se abrió o se
reabrió la propuesta, o cuando se puso la etiqueta. **Para volver a medir, la etiqueta**: una que ya está puesta no
lanza nada, así que se quita y se vuelve a poner. Una propuesta que solo toca documentación no arranca el job, y es lo
esperado. **Una sola tanda por commit**: el trabajo de cada skill lleva una `concurrency` por commit y skill con
`cancel-in-progress: false`, así que un segundo disparo sobre el mismo commit mientras el primero sigue —la apertura y
la etiqueta, por ejemplo— espera a que termine, sin cancelarse ni saltarse, y abre su tanda después; no hay
`concurrency` de nivel de flujo, que detendría también el trabajo de la otra skill. El informe se imprime en el registro de la ejecución, entre las marcas `--- inicio de informe.md ---` y
`--- fin de informe.md ---` (y las mismas de `informe.json`), y en el resumen de la ejecución. La *Definition of Done* (punto 10) pide las evals de la skill en verde sobre un commit
del que la cabeza solo difiere en el directorio del hito en `specs/`: vale la ejecución de apertura si después no
cambia nada fuera de ese directorio y, si cambia, se repiten por etiqueta tras el último cambio, porque un cambio
posterior obliga a repetirlas.

Cómo se lee el informe:

- **El veredicto** es `aprobado` o `fallo`, y los motivos son exactamente las causas del fallo: una serie que decide y
  no llega al umbral —`<eval> con <modelo>: pasan 1 de 3, y el umbral es 2`, seguido de los motivos de sus sesiones
  que no pasan—, una serie planificada con más o menos sesiones de las que pide el plan, una sesión ilegible, un
  fichero mal formado del directorio de evals —una eval o la lista de expresiones prohibidas—, una petición llegada
  a la red o un umbral que decide y no se cumple
  (`umbral expresiones_prohibidas:claude-sonnet-5: 3 de 51 (5,9 %), y tiene que ser ≤ 5,0 %`). Detrás van, con el
  prefijo fijo `de la ejecución, no de la skill: `, que los distingue sin modelo de los de la skill, el de las sesiones
  sin medir por un límite de uso de la cuenta y el de la duración de las sesiones por encima de su objetivo: dicen que
  el job no pudo medir, o no midió a tiempo, y no piden cambiar la skill.
- **La tabla «Tasas por eval»** tiene una fila por serie, con si decide, si la pide el plan, la tasa
  (`<pasan> de <sesiones>`) y si llega al umbral. Se publica también la de las series que pasan: un `2 de 3` es verde,
  pero es la degradación que conviene ver antes de que se vuelva roja. Un fallo aislado no se ve en el veredicto; se
  ve aquí y en la tabla de sesiones.
- **Una serie informativa** —la de un modelo informativo, o la de una eval `informativa: true` con el modelo que
  decide— se ejecuta y se publica con la columna «Decide» en `no`: no llegar al umbral no da ningún motivo, así que su
  tasa se mira, pero no bloquea. Lo que no depende de la tasa cuenta en cualquier serie: una sesión que falta, una
  ilegible, una sin medir o una petición llegada a la red hacen fallar el veredicto igual.
- **Las expresiones prohibidas**: cada sesión publica en `informe.json` las de la lista que lleva su respuesta,
  `expresiones_prohibidas` —`[]` si ninguna o si no se le aplica la lista—, y la tabla «Sesiones» de `informe.md`, la
  columna «Expresiones prohibidas» (`ninguna` si no lleva ninguna). En la raíz de `informe.json`,
  `expresiones_prohibidas_por_modelo` da un elemento por modelo del job —el que decide y después los informativos, en
  su orden— con `modelo`, `respuestas`, las sesiones medidas (ni las ilegibles ni las sin medir) de las series
  planificadas cuya eval activa la skill, las informativas incluidas, y `con_alguna`, las de ellas que llevan alguna
  expresión; `informe.md` lo da en la sección «Expresiones prohibidas por modelo», detrás de «Tasas por eval». Una
  skill sin lista da `[]` y el párrafo «la skill no tiene lista de expresiones prohibidas»: un recuento de cero diría
  que se buscó. Una sesión con alguna expresión no pasa, y su serie decide con el umbral de siempre si es de las que
  deciden; además, el recuento de cada modelo es un umbral.
- **Los umbrales**: `umbrales`, siempre en la raíz de `informe.json`, con el contrato del ADR 0029 —`nombre`,
  `descripcion`, `medida`, `total` (solo si se compara una proporción), `comparacion` (`"<="`), `umbral`, `cumple` y
  `decide`—; `cumple` es la comparación en coma flotante, sin redondeos, de `medida` entre `total` (0 si `total` es 0)
  o de la propia `medida`, y se puede rehacer con los otros campos. Una skill con lista tiene uno por modelo del job,
  `expresiones_prohibidas:<modelo>`, con el recuento de ese modelo y `umbral` `0.05`, que solo decide en el modelo que
  decide: con las 51 respuestas de `claude-sonnet-5` en las evals de `boe-legislacion` que activan la skill, como
  mucho 2 con alguna expresión; el de `claude-haiku-4-5-20251001` se publica con `decide: false`. Una skill con
  objetivo de duración tiene además `duracion_de_las_sesiones`, sin `total` y con `decide: true`; `legal-core` no
  tiene ninguno, `[]`. Uno que decide y no se cumple da su motivo y `fallo`; uno que se cumple, o que no decide, no
  cambia nada. `informe.md` los da en la sección «Umbrales», detrás de «Expresiones prohibidas por modelo»: la tabla
  `Umbral | Medida | Condición | Cumple | Hace fallar el veredicto` (`no: solo se publica` en los que no deciden), o
  `ninguno`.
- **Las sesiones sin medir**: una sesión que un límite de la cuenta no dejó terminar no es una eval fallida. El informe
  la reconoce sin modelo por su transcript —(a) su resultado es el mensaje del límite de uso de Claude Code, el que
  empieza por `You've hit your`, `You've reached your` o `You're out of`; (b) agotó los reintentos por `rate_limit`;
  (c) la cortó el tope durante reintentos por `rate_limit`— y la deja sin medir: no pasa ni falla, lleva como único
  motivo `sin medir por límite de uso: <clase>` y no cuenta en `respuestas` ni en `con_alguna`, así que tampoco en los
  umbrales. Su serie queda sin medir —`sin medir (<n>)` en la columna «Resultado» de «Tasas por eval», sin pasar y sin
  el motivo «pasan N de M»—. Los reintentos por otra causa, como la sobrecarga, no son un límite de la cuenta, y una
  sesión que se recupera de sus reintentos se mide como cualquier otra. Tras una de la clase (a), que no se repone
  dentro del job, el job no abre ninguna sesión más: las abiertas terminan y se juzgan, y las que faltaban del plan
  cuentan en su serie como sin medir, `sin abrir tras el límite de uso`; tras (b) o (c), sigue. `informe.json` las
  lista en `sesiones_sin_medir` (`sesion`, `eval`, `modelo` y `motivo`), y `informe.md`, en la sección «Sesiones sin
  medir», detrás de «Umbrales», y en la columna «Sin medir» de «Sesiones». Con alguna, el veredicto es `fallo` por la
  ejecución, no por la skill: no pide cambiarla. El motivo de una sesión que no terminó por otra causa lleva el texto
  del último `result` con `is_error` de su transcript (`código 1: result con is_error: Failed to authenticate. …`, el
  de una credencial que no sirve).
- **Los reintentos y la duración**: `reintentos_por_limite_de_ritmo`, en cada sesión y sumados en la raíz, cuenta los
  reintentos por `rate_limit` de su transcript; es el dato para ajustar la concurrencia. `duracion_de_las_sesiones`
  son los segundos desde que se prepara la primera sesión hasta que termina la última, sin la preparación del runner.
  La cabecera de `informe.md` lleva `Duración de las sesiones: <s> s` y `Reintentos por límite de ritmo: <n>`, y la
  tabla «Sesiones», la columna «Reintentos por límite de ritmo». En `boe-legislacion`, más de 900 s da `fallo` con el
  motivo `de la ejecución, no de la skill: duracion_de_las_sesiones: <s> s, y tiene que ser ≤ 900 s`.

La **prueba de red** añade al trabajo de `boe-legislacion` —`territorio` no puede pedir nada a la red, así que el de
`legal-core` no la lleva—, con el modelo que decide, una sesión con la pregunta de la primera eval y dos consultas a
un bloque que no está grabado, sin y con `--offline`: comprueba que el binario no alcanza la fuente —termina con `5` y
con `4` sin pedirle nada— y que el informe registra las dos como consultas fuera de lo grabado. No se repite ni decide:
su fila de tasas lleva «(pregunta ampliada)» y «Planificada» en `no`. Se juzga con la lista de expresiones prohibidas
y publica las suyas, pero no entra en `expresiones_prohibidas_por_modelo` ni en los umbrales de las expresiones: el
recuento es el mismo con la prueba de red y sin ella. Sí ocupa su sitio entre las sesiones que se abren a la vez y
cuenta en la duración y en los reintentos.

## `make vuln` necesita red

`govulncheck` consulta la base de datos de vulnerabilidades de Go. **Sin red, `make vuln` falla**, y ese
fallo **no significa «no hay vulnerabilidades»**: significa que el control no se ha podido ejecutar.

Ese error **no se captura**. Devolver 0 con un aviso convertiría un gate en un adorno: quien trabajara
sin conexión vería el veredicto en verde sin que nadie hubiera mirado nada, y el fallo aparecería más
tarde, en la propuesta de cambio, cuando ya no es evidente de dónde viene. Si estás sin conexión,
`make vuln` te lo dirá y el resto de `make ci` seguirá siendo ejecutable por separado; el veredicto
completo espera a que haya red.

La primera ejecución de cualquier orden también necesita red: descarga y compila la herramienta que usa
y, si hace falta, el parche de Go que fija la directiva `toolchain`. A partir de ahí todo lo sirve la
caché de Go sin conexión.

## Excluir un falso positivo de la detección de secretos

`make secrets` es `gitleaks dir . --redact --no-banner` sobre **todo** el árbol. Cuando marca algo que
está comprobado que no es una credencial, se excluye **ese hallazgo concreto**, por su huella, dejando
por escrito por qué. El control no se toca: no se desactiva ninguna regla, no se excluye ninguna ruta y
no se añade un `.gitleaks.toml` con exclusiones por patrón —un patrón silenciaría también los hallazgos
futuros que nadie ha visto todavía, y eso ya no es excluir un falso positivo, es apagar el control.

Procedimiento:

1. **Comprobar que de verdad es un falso positivo.** Abrir el fichero y la línea. Si hay la más mínima
   duda de que sea una credencial real, no se excluye: se rota y se retira del árbol.
2. **Copiar la huella de la salida de `make secrets`.** Cada hallazgo se imprime con estas líneas, y la
   última es la que hace falta —tal cual, sin reescribirla a mano—:

   ```text
   Finding:     ...
   Secret:      REDACTED
   RuleID:      generic-api-key
   Entropy:     3.923538
   File:        ruta/al/fichero.md
   Line:        31
   Fingerprint: ruta/al/fichero.md:generic-api-key:31
   ```

   El formato de la huella es `fichero:regla:línea`. Con `--redact`, el secreto sale como `REDACTED`, así
   que copiar esa línea nunca copia una credencial.
3. **Añadirla a `.gitleaksignore`**, en la raíz del repositorio, **precedida de un comentario que la
   justifique**: qué es ese texto, por qué no es una credencial de este proyecto y qué lo acredita. Una
   línea de huella, un comentario. Una huella sin comentario es indistinguible de un descuido.
4. **Volver a ejecutar `make secrets`** —debe quedar en verde— y `make ci`.
5. **Explicarlo en la propuesta de cambio**, en el apartado *Decisiones*. Añadir una huella es una
   decisión revisable, no un detalle de configuración.

Una ruta desnuda en `.gitleaksignore` no excluye nada: en el modo `dir` solo funciona la huella
completa. Y como la huella lleva el número de línea, deja de casar si el fichero se desplaza: cuando eso
pase, el hallazgo reaparecerá y habrá que actualizar la línea —y volver a comprobar que sigue siendo un
falso positivo—, que es exactamente el comportamiento que se quiere.

El fichero nace en H0 con dos huellas reales, ambas de la documentación vendorizada de las skills de
agente: sirven de ejemplo del formato.

## Subir el parche de Go (directiva `toolchain`)

La versión de Go aparece en **un solo sitio**: la directiva `toolchain` de `go.mod`. El `Makefile` la lee
y exporta `GOTOOLCHAIN`, de modo que todas las órdenes —en local y en la integración continua— compilan y
analizan con ese parche exacto. Ningún flujo de `.github/workflows/` declara la versión por su cuenta.

**El disparador de la subida es un hallazgo de `make vuln`**, no un calendario. `govulncheck` analiza
también la biblioteca estándar del toolchain en uso: cuando se publique un parche de Go que corrija un
fallo de la stdlib, `make vuln` empieza a fallar sobre el parche fijado, en la propuesta de cambio y en
el flujo nocturno sobre `main` —que existe precisamente para que un hallazgo nuevo aparezca sin esperar a
la siguiente propuesta—. El arreglo no es rebajar el control, es subir el parche.

Procedimiento manual, cuatro pasos:

1. Editar **solo** la directiva `toolchain` de `go.mod`, al parche que corrige el hallazgo
   (`toolchain go1.27.2`). El `Makefile` no se toca: no repite el número. La directiva `go` se queda como
   está —marca la versión mínima del lenguaje, no el parche—.
2. `make ci`. Tiene que quedar en verde, y `vuln` en particular: si el hallazgo persiste, el parche
   elegido no lo corrige.
3. Commit de una línea: `chore(deps): go1.27.2`, con el identificador de la vulnerabilidad en el cuerpo.
4. Propuesta de cambio como cualquier otra. La subida de parche es un cambio versionado y revisable, no
   una actualización silenciosa.

Si Dependabot llegara a proponer esta subida por sí solo —está por comprobar sobre el repositorio real—,
su propuesta se revisa igual que cualquier otra y este procedimiento queda como camino manual.

## La web

La web, https://kitlegal.es, vive en `web/` (ADR 0024): Astro, HTML estático y ningún script en la página. Necesita
Node 22.12 o superior y pnpm (`corepack enable pnpm`); nada más del repositorio los necesita, y por eso no entra en
`make ci`. El flujo `.github/workflows/web.yml` la construye con `make web` en cada propuesta de cambio que toca
`web/`, `data/` o `schemas/`, y la publica en GitHub Pages desde `main` y al terminar cada release.

**Ninguna cita se escribe a mano.** Para citar un artículo, se declara en `web/src/data/citas.yaml` —norma, bloque y
los fragmentos, copiados del texto— y se ejecuta `make web-citas`, que pide al binario el sobre de cada cita y lo
guarda en `web/src/data/sobres/`. Los sobres se versionan con el cambio. `make web` falla si un fragmento no está
literal en su sobre o si los fragmentos no siguen el orden del texto, y la página une con « […] » los que no son
contiguos. Las cifras (municipios, versión) salen de `data/` y de la última release, nunca del texto de la página.

En `/llms.txt`, el índice de la web para agentes de IA ([llmstxt.org](https://llmstxt.org/)), solo el resumen está
escrito en `web/src/pages/llms.txt.ts`. El resto lo toma al construir: el título y la descripción de cada página
(`web/src/data/`), las órdenes de instalación, las cifras, las skills de la última versión publicada —leídas del
árbol de su etiqueta, que el flujo de la web trae— y los esquemas de `schemas/`. Para cambiar un enlace o su
descripción, se cambia su origen.

Las dependencias de la web van en `web/package.json` y `web/pnpm-lock.yaml`. pnpm no instala una versión hasta que
lleva una semana publicada (`web/pnpm-workspace.yaml`), y Dependabot las propone agrupadas con la misma espera.

## Dónde está escrito lo demás

- [`README.md`](README.md) — para quien usa kitlegal: qué sabe hacer su agente con él, cómo instalarlo y qué le
  garantiza. Construirlo y ejecutar los controles está aquí.
- [`docs/ROADMAP.md`](docs/ROADMAP.md) — los hitos, su orden y la *Definition of Done*.
- [`docs/ADR/`](docs/ADR/) — las decisiones de arquitectura, con contexto y consecuencias.
- [`docs/SOURCES.md`](docs/SOURCES.md) — las fuentes externas: licencia, términos de uso, `robots.txt`,
  ritmo y fecha de la revisión humana.
- [`CLAUDE.md`](CLAUDE.md) — convenciones del repositorio, incluida la de idioma: documentación, verbos de
  applet y claves JSON en español; identificadores Go según la convención del lenguaje.
- [`.specify/memory/constitution.md`](.specify/memory/constitution.md) — principios, restricciones y el
  criterio de decisión autónoma.
