# Cómo contribuir a kitlegal

Esta guía es el contrato de trabajo del repositorio: cómo se organiza un hito, qué forma tiene una
propuesta de cambio, qué controles hay que ver en verde antes de abrirla y qué se espera de quien
añade una dependencia. Lo que hace falta para construir y ejecutar el binario está en
[`README.md`](README.md); las decisiones ya cerradas, en [`docs/ADR/`](docs/ADR/).

Prerrequisitos: `go` (1.21 o superior) y `git`. Nada más —ni las herramientas de los controles ni un
parche concreto de Go—; el porqué está en el README.

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
5. **Squash-merge**. Si el hito cierra una fase, etiqueta y release.
6. **Actualizar el roadmap solo si cambia el orden o el alcance**; el detalle vive en las propuestas de
   cambio y en los ADR.

Un hito no se cierra sin la *Definition of Done* completa (`docs/ROADMAP.md` §1). La fusión a `main` y
el release son siempre acciones humanas.

Los hitos se preparan con el workflow `hito` de spec-kit (`scripts/hito.sh H<n>`, documentado en
[`docs/WORKFLOW.md`](docs/WORKFLOW.md)), que deja sus artefactos en `specs/NNN-hN-slug/`. Los principios
que rigen las decisiones automáticas están en [`.specify/memory/constitution.md`](.specify/memory/constitution.md).

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

**Versionado semántico.** El proyecto está en `0.y.z` hasta la primera release, que es H6 (`v0.1.0`);
mientras el mayor sea `0`, un cambio incompatible sube el **menor**. A partir de `1.0.0`, mayor para lo
incompatible, menor para funcionalidad nueva compatible y parche para correcciones. Las etiquetas son
`vX.Y.Z` y las pone una persona, nunca la integración continua.

**`CHANGELOG.md`** sigue el formato *Keep a Changelog*: una sección `## [Unreleased]` siempre presente en
cabeza, y bajo ella los apartados `Añadido`, `Cambiado`, `Obsoleto`, `Eliminado`, `Corregido` y
`Seguridad`, solo los que tengan contenido. Todo cambio de comportamiento visible entra en *Unreleased*
en la misma propuesta que lo introduce; al publicar una versión, esa sección se cierra bajo su número y
su fecha y se abre una nueva vacía. En H0 el changelog se mantiene **a mano**; su generación automática
llega con el release de H6.

## Los controles

`make ci` es **el veredicto del repositorio**: si está en verde en local, la propuesta de cambio pasa,
porque la integración continua ejecuta esa misma orden y no aplica ningún control por otra vía.
`make ci` no modifica ningún fichero versionado, así que se puede ejecutar con el árbol sucio sin miedo.

| Control | Orden | ¿En `make ci`? |
|---|---|---|
| Formato (`gofumpt` + `goimports`), modo verificación | `make fmt-check` | sí |
| Formato, modo corrección | `make fmt` | no — **corrige**, y corregir el árbol para aprobarlo no es un gate |
| Análisis estático (`golangci-lint`, `gosec` y `govet` incluidos) | `make lint` | sí |
| Análisis estático rápido | `make lint-fast` | no — es el del gancho de pre-commit |
| Tests unitarios con detector de carreras y perfil de cobertura | `make test` | sí |
| Tests con la etiqueta `integration` (dependen del entorno: permisos, dos procesos, la instalación de las skills con `make install` en un directorio personal temporal) | `make test-integration` | sí |
| Vulnerabilidades conocidas (`govulncheck`) | `make vuln` | sí |
| Esquemas publicados en `schemas/` iguales a lo que emite `--describe` de cada verbo, sin escribir nada | `make schema-check` | sí |
| Skills, datos y evals, sin red, sin modelo y sin escribir nada: frontmatter y límite de líneas de cada `SKILL.md`; derivas de las referencias, de la tabla de comandos y de los enlaces de `scripts/`; tabla de normas contra su esquema y sus identificadores; ficheros congelados de `data/territorio/` contra sus esquemas y su integridad; jerarquía normativa contra su esquema; formato y conjunto de evals y lo grabado que necesitan | `make skills-check` | sí |
| Regeneración de lo que se deriva de cada skill (referencias, tabla de comandos de `SKILL.md`, enlaces de `scripts/`) | `make skills-sync` | no — escribe en el árbol |
| Detección de secretos (`gitleaks`) | `make secrets` | sí |
| Integridad de los módulos (`go mod verify`, raíz y herramientas) | `make mod-verify` | sí |
| Dependencias saneadas (`go mod tidy -diff`) | `make mod-tidy-check` | sí |
| Prerrequisitos (`go`, `git`, toolchain fijado obtenible) | `make check-tools` | sí, como dependencia de las demás |
| Verificación contra la fuente real (`scripts/verify-sources.sh`; requiere red) | `make verify-sources` | no — toca la red; lo ejecuta el trabajo `fuentes` del flujo nocturno, que abre o comenta una incidencia si falla |
| Evals de una skill con Claude Code (`scripts/evals.sh`; Linux con `strace`, como root o con `sudo`) | `make evals` | no — sesiones con modelo y credencial, fuera de `make ci`; las lanza el job de evals |
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

## Órdenes que existen pero reciben su contenido en un hito posterior

Ninguna miente ni pasa en silencio: cada una nombra el objeto ausente y el hito que lo aporta.

| Orden | Qué hace hoy | Hito |
|---|---|---|
| `make release` | **Falla** con código distinto de 0 | H6 (`.goreleaser.yaml`) |

`release` falla en lugar de anunciar lo que le falta y terminar con éxito porque es una acción con
efectos externos: no puede simular éxito. No forma parte de `ci` ni del flujo nocturno.

`make test-e2e`, `make test-integration`, `make schema-check` y `make skills-sync` ya no están en esta
tabla: desde H1 la primera ejecuta los guiones `testscript` contra el binario que el propio test
construye; desde H3 la segunda ejecuta los tests etiquetados `integration` —los que dependen del entorno:
los de la caché y, a partir de H5, el de la instalación de las skills— y forma parte de `make ci`; desde
H4 la tercera compara `schemas/` con lo que emite `--describe` (sección siguiente); y desde H5, `make skills-sync`
regenera, desde `data/*.yaml` y desde `--describe` del binario, las referencias, la tabla de comandos de
`SKILL.md` y los enlaces de `scripts/` de cada skill (sección [«Skills y evals»](#skills-y-evals)).

## `make schema-check` y `make verify-sources`

`make schema-check` regenera en memoria, desde `--describe` de cada verbo que registra el binario
distribuido, los esquemas publicados en `schemas/` —hoy `norma.json` y `bloque.json`, los de `boe`, y
`municipio.json`, el de `territorio`— y los
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

## Skills y evals

Una skill es un directorio sin código bajo `skills/`: `SKILL.md`, `references/` y `scripts/`. Qué son los tres
directorios llamados `skills`, qué hace `make install` y qué se genera está en el [`README.md`](README.md#skills);
esta sección es lo que hace falta para cambiar una skill, sus datos o sus evals.

**Lo generado no se edita.** `references/*.md`, la tabla de comandos de `SKILL.md` —entre sus marcas— y los enlaces
de `scripts/` se derivan de `data/*.yaml` y de `--describe` del binario. Tras cambiar `data/`, añadir un verbo o
cambiar su entrada o su salida, se ejecuta `make skills-sync` y lo regenerado va en el mismo cambio:
`make skills-check`, dentro de `make ci`, lo regenera en memoria y falla nombrando la skill y el fichero o el enlace
que difieren. Comprueba además el frontmatter de cada `SKILL.md` y que tenga menos de 300 líneas, la tabla de normas
contra `schemas/normas.yaml.json`, que cada identificador está en la búsqueda grabada del BOE, la jerarquía normativa
de `data/jerarquia.yaml` contra `schemas/jerarquia.yaml.json`, los ficheros congelados de `data/territorio/` contra
sus esquemas y su integridad, y el formato y el conjunto de las evals y que lo que necesitan está grabado. Una norma
nueva, o una eval que consulta algo que no está grabado, llega con su grabación, que hace una persona con
`scripts/grabar-evals.sh`: ningún control ni flujo graba respuestas.

**Los datos de territorio no se regeneran con `make skills-sync`**: `data/territorio/` no deriva de nada del
repositorio, sino de descargas públicas que no se consultan en red (ADR 0017). Los refresca una persona, fuera del
repositorio, desde la relación de municipios del INE y el volcado del Registro de Entidades Locales, con la fila de
cada origen en `docs/SOURCES.md` puesta a la fecha del fichero. Añadir un territorio es rellenar `boletines` en el
fichero de su comunidad, `data/territorio/comunidades/<código>.yaml`, sin tocar código ni skills. Los ficheros viajan
dentro del binario, así que un cambio en ellos llega a `territorio resolver` al volver a construirlo.

### Formato común de eval

Las evals de una skill se escriben **antes** que la skill o que el cambio que la mejora (ritual, paso 1), en su propio
directorio, `evals/<skill>/`, con un fichero YAML por eval llamado `<nn>-<descripción>.yaml` —dos cifras y una
descripción en minúsculas con guiones—. Todas siguen el formato común de eval:

| Campo | ¿Obligatorio? | Qué fija |
|---|---|---|
| `pregunta` | sí | La pregunta con la que se abre la sesión; no vacía |
| `activa` | sí | Si la pregunta debe activar la skill |
| `comandos` | sí si `activa` es `true`; prohibido si es `false` | Las consultas que la sesión debe hacer con éxito, cada una en una de cuatro formas: un bloque (`applet`, `norma`, `bloque`), una consulta de norma (`applet`, `verbo` —`indice`, `metadatos` o `analisis`—, `norma`), una búsqueda (`applet`, `verbo` `buscar`, `terminos`) o un municipio (`applet`, `verbo` `resolver`, `municipio`) |
| `citas` | sí si `activa` es `true` y no hay `territorio`; prohibido si es `false` | Cada `norma` y `bloque` que la respuesta debe citar |
| `territorio` | sí si `activa` es `true` y no hay `citas`; prohibido si es `false` | Lo que la respuesta debe declarar del territorio que devuelve `territorio resolver`, con al menos una de estas claves: `comunidad`, `provincia`, los códigos de `boletines` y los aspectos de `cobertura` en la forma `<aspecto>: <valor>` del vocabulario del applet (`boletin_autonomico: no-configurado`…) |
| `avisos` | no; solo si `activa` es `true`, prohibido si es `false` | Los códigos de aviso de vigencia del binario (`consolidacion-no-finalizada`, `derogada`, `vigencia-agotada`) cuya forma fija —`⚠`, la etiqueta del aviso y dos puntos— debe llevar la respuesta |
| `informativa` | no | Con `true`, la eval se ejecuta solo con el modelo que decide y su tasa se publica, pero no decide el veredicto (ADR 0016). En `boe-legislacion`, solo en una eval que activa la skill |
| `reproduce` | no | La skill cuyo uso documentado reproduce la eval (p. ej. `boe-fiscal`) |

Cada fichero de cada directorio `evals/<skill>/`, sea de la skill que sea, se valida contra `schemas/eval.yaml.json`
dentro de `make ci`. Una entrada del directorio que no es un
fichero con esa forma de nombre, una clave desconocida o repetida, un identificador o un bloque mal escritos, una
eval positiva sin citas ni territorio o una de no activación con comandos fallan nombrando el fichero; ninguna se
salta. Para `boe-legislacion`, `make ci` exige además las reglas de su conjunto: exactamente diez positivas que
deciden, de materias distintas, al menos una de no activación y al menos una informativa, y ninguna informativa de no
activación, entre otras. Para `legal-core`, al menos tres evals: una positiva que resuelve un municipio del territorio
configurado y declara sus boletines, otra que resuelve uno de una comunidad sin configuración y declara no
configurados el boletín autonómico y el provincial, al menos una de no activación, y citas o territorio en toda
positiva.

Una sesión de una eval pasa si la abre el modelo pedido, activa la skill cuando debe y no la activa cuando no debe,
termina, hace con éxito cada consulta de `comandos` y responde citando cada `norma` y `bloque` de `citas`, declarando lo
que espera `territorio` —la comunidad y la provincia sin distinguir mayúsculas ni tildes, el código de cada boletín
como palabra exacta y cada aspecto de cobertura en su forma fija `<aspecto>: <valor>`— y con la forma fija de cada
aviso de `avisos`. Lo juzga el informe sin modelo, con lo que deja la sesión: su transcript y su
traza. Cada eval se abre varias veces con un mismo modelo, y esa serie pasa si las sesiones que pasan llegan al umbral
([Job de evals](#job-de-evals)).

### Job de evals

`make evals SKILL=<skill>` ejecuta `scripts/evals.sh` y no forma parte de `make ci`: sus sesiones usan un modelo,
necesitan la credencial de Claude Code, cuestan y no son deterministas. Necesita Linux con `strace`, root o `sudo` y
ningún Python accesible. Antes de la primera sesión comprueba todo eso, que ninguna eval está mal formada, que lo que
necesitan está grabado y que la skill está instalada, y termina con `1` si algo falla. Después abre las sesiones de
Claude Code del plan, cada una bajo `strace` y con la red cerrada salvo la del modelo: cada eval, tantas veces como
repeticiones, con el modelo que decide y, si no es informativa, otras tantas con cada modelo informativo. Juzga cada
sesión y agrupa las de cada eval con cada modelo en una serie con su tasa, cuántas de sus sesiones pasan. El informe
publica la tasa de cada serie y da un veredicto global que falla si una serie que decide —la de una eval que no es
informativa con el modelo que decide— no llega al umbral, una serie no tiene exactamente las sesiones que pide el
plan, una sesión es ilegible, un fichero está mal formado, no hay ninguna eval bien formada que juzgar o una petición
llega a la red (ADR 0016).

Lo ejecuta el job de evals, el flujo `evals` (`.github/workflows/evals.yml`), con un trabajo por skill en la misma
ejecución —hoy `boe-legislacion` y `legal-core`—, cada uno con su informe y sin que el rojo de uno cancele el otro.
Fija en su definición, cada uno en su variable, el modelo que decide (`MODELO_DE_EVALS`, el del uso real de la skill), los modelos informativos
(`MODELOS_INFORMATIVOS_DE_EVALS`, separados por comas, que se publican como límite inferior sin decidir), las
repeticiones de cada eval con cada modelo (`REPETICIONES_DE_EVALS`) y el umbral de sesiones que pasan
(`UMBRAL_DE_EVALS`). Los modelos van por su identificador completo: cambiar de modelo es un cambio de ese fichero. Usa
el secreto de repositorio `CLAUDE_CODE_OAUTH_TOKEN`, el token de la suscripción de Claude que da `claude setup-token`
(el proyecto no usa una clave de API de pago por uso). Hoy fija:

| Variable | Valor |
|---|---|
| `MODELO_DE_EVALS` | `claude-sonnet-5` |
| `MODELOS_INFORMATIVOS_DE_EVALS` | `claude-haiku-4-5-20251001` |
| `REPETICIONES_DE_EVALS` | `3` |
| `UMBRAL_DE_EVALS` | `2` |

Con las dieciocho evals de `boe-legislacion`, su trabajo abre 90 sesiones: 36 de `claude-sonnet-5` sobre las doce
que deciden, 18 sobre las seis informativas y 36 de `claude-haiku-4-5-20251001` sobre las doce que deciden, dentro
del tope de 120 minutos del job (`timeout-minutes`). Se lanza de tres formas:

| Lanzamiento | Sobre qué rama | Cómo |
|---|---|---|
| Manual | La que se elija | Desde la plataforma, con la entrada `prueba_de_red` si se quiere también la prueba de red |
| Apertura | La de una propuesta de cambio, antes de fusionar | Al abrirla o reabrirla, si toca lo que las evals miden: `skills/`, `evals/`, `data/`, `internal/source/boe/`, `internal/core/`, `internal/cli/`, `internal/evals/`, `scripts/evals.sh`, `.github/workflows/evals.yml`, `schemas/eval.yaml.json` o el `Makefile` |
| Por etiqueta | La de cualquier propuesta de cambio, antes de fusionar | Poniendo la etiqueta `evals` en su propuesta de cambio; `evals-prueba-de-red` añade la prueba de red |

No hay ejecución programada: la semanal sobre `main`, con el modelo, la versión de Claude Code y las respuestas del
BOE fijados, no medía ningún cambio. Tampoco reacciona a cada empujón (`synchronize`): cada ejecución abre decenas de
sesiones con modelo y un hito empuja muchas veces, así que cada informe mide el commit que había cuando se abrió o se
reabrió la propuesta, o cuando se puso la etiqueta. **Para volver a medir, la etiqueta**: una que ya está puesta no
lanza nada, así que se quita y se vuelve a poner. Una propuesta que solo toca documentación no arranca el job, y es lo
esperado. El informe se imprime en el registro de la ejecución, entre las marcas `--- inicio de informe.md ---` y
`--- fin de informe.md ---` (y las mismas de `informe.json`), y en el resumen de la ejecución. La *Definition of Done* (punto 10) pide las evals de la skill en verde sobre un commit
del que la cabeza solo difiere en el directorio del hito en `specs/`: vale la ejecución de apertura si después no
cambia nada fuera de ese directorio y, si cambia, se repiten por etiqueta tras el último cambio, porque un cambio
posterior obliga a repetirlas.

Cómo se lee el informe:

- **El veredicto** es `aprobado` o `fallo`, y los motivos son exactamente las causas del fallo: una serie que decide y
  no llega al umbral —`<eval> con <modelo>: pasan 1 de 3, y el umbral es 2`, seguido de los motivos de sus sesiones
  que no pasan—, una serie planificada con más o menos sesiones de las que pide el plan, una sesión ilegible, un
  fichero de eval mal formado o una petición llegada a la red.
- **La tabla «Tasas por eval»** tiene una fila por serie, con si decide, si la pide el plan, la tasa
  (`<pasan> de <sesiones>`) y si llega al umbral. Se publica también la de las series que pasan: un `2 de 3` es verde,
  pero es la degradación que conviene ver antes de que se vuelva roja. Un fallo aislado no se ve en el veredicto; se
  ve aquí y en la tabla de sesiones.
- **Una serie informativa** —la de un modelo informativo, o la de una eval `informativa: true` con el modelo que
  decide— se ejecuta y se publica con la columna «Decide» en `no`: no llegar al umbral no da ningún motivo, así que su
  tasa se mira, pero no bloquea. Lo que no depende de la tasa cuenta en cualquier serie: una sesión que falta, una
  ilegible o una petición llegada a la red hacen fallar el veredicto igual.

La **prueba de red** añade al trabajo de `boe-legislacion` —`territorio` no puede pedir nada a la red, así que el de
`legal-core` no la lleva—, con el modelo que decide, una sesión con la pregunta de la primera eval y dos consultas a
un bloque que no está grabado, sin y con `--offline`: comprueba que el binario no alcanza la fuente —termina con `5` y
con `4` sin pedirle nada— y que el informe registra las dos como consultas fuera de lo grabado. No se repite ni decide:
su fila de tasas lleva «(pregunta ampliada)» y «Planificada» en `no`.

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

## Dónde está escrito lo demás

- [`README.md`](README.md) — qué es kitlegal, cómo construirlo y cómo ejecutar los controles.
- [`docs/ROADMAP.md`](docs/ROADMAP.md) — los hitos, su orden y la *Definition of Done*.
- [`docs/ADR/`](docs/ADR/) — las decisiones de arquitectura, con contexto y consecuencias.
- [`docs/SOURCES.md`](docs/SOURCES.md) — las fuentes externas: licencia, términos de uso, `robots.txt`,
  ritmo y fecha de la revisión humana.
- [`CLAUDE.md`](CLAUDE.md) — convenciones del repositorio, incluida la de idioma: documentación, verbos de
  applet y claves JSON en español; identificadores Go según la convención del lenguaje.
- [`.specify/memory/constitution.md`](.specify/memory/constitution.md) — principios, restricciones y el
  criterio de decisión autónoma.
