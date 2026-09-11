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
4. **Revisión** de código y de seguridad sobre la propuesta. Si el hito toca una fuente externa,
   `docs/SOURCES.md` se actualiza en el mismo cambio.
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
| Tests con la etiqueta `integration` | `make test-integration` | no |
| Vulnerabilidades conocidas (`govulncheck`) | `make vuln` | sí |
| Validación contra esquemas | `make schema-check` | sí |
| Detección de secretos (`gitleaks`) | `make secrets` | sí |
| Integridad de los módulos (`go mod verify`, raíz y herramientas) | `make mod-verify` | sí |
| Dependencias saneadas (`go mod tidy -diff`) | `make mod-tidy-check` | sí |
| Prerrequisitos (`go`, `git`, toolchain fijado obtenible) | `make check-tools` | sí, como dependencia de las demás |
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
| `make test-integration` | Ejecuta ya el comando real (`go test -race -tags=integration ./...`), que hoy pasa sobre un conjunto vacío de tests | según vaya habiendo tests de integración |
| `make test-e2e` | Anuncia que no hay tests de extremo a extremo todavía y termina con éxito | H1 (`testscript`) |
| `make schema-check` | Anuncia que no hay `schemas/` todavía y termina con éxito; `ci` lo invoca y sigue en verde | H4 (borrador) y H10 (contrato) |
| `make skills-sync` | Anuncia que no hay `skills/` ni `data/*.yaml` todavía y termina con éxito | H5 |
| `make release` | **Falla** con código distinto de 0 | H6 (`.goreleaser.yaml`) |

`release` es la excepción porque es una acción con efectos externos: no puede simular éxito. No forma
parte de `ci` ni del flujo nocturno.

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
- [`CLAUDE.md`](CLAUDE.md) — convenciones del repositorio, incluida la de idioma: documentación, verbos de
  applet y claves JSON en español; identificadores Go según la convención del lenguaje.
- [`.specify/memory/constitution.md`](.specify/memory/constitution.md) — principios, restricciones y el
  criterio de decisión autónoma.
