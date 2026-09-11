# Implementation Plan: H0 · Esqueleto del repo y gates de CI

**Branch**: `h0-esqueleto-del-repo` | **Date**: 2026-09-10 | **Spec**: [spec.md](./spec.md)

**Input**: Feature specification from `/specs/001-h0-esqueleto-del-repo/spec.md`

**Modo**: desatendido. Las decisiones técnicas se tomaron con «Criterio de decisión autónoma» de
`.specify/memory/constitution.md` y están registradas con alternativas y motivo en
[research.md](./research.md).

## Summary

H0 convierte un repositorio que hoy solo tiene documentos de diseño en un módulo Go publicable y, sobre
todo, **blindado**: un binario `kitlegal` cuyo único verbo es `version`, y el conjunto completo de
controles automáticos (formato, lint, race, vulnerabilidades, SAST, secretos, cadena de suministro,
dependencias mínimas, pre-commit, cobertura) que toda línea de Go posterior tendrá que atravesar, con el
**mismo veredicto en local y en la PR**.

El enfoque técnico se apoya en tres decisiones que sostienen todo lo demás:

1. **`make ci` es el veredicto**, y contiene los ocho controles que operan sobre un árbol de trabajo
   (FR-007). El flujo de CI no ejecuta ningún control por su cuenta: solo prepara el entorno, invoca
   `make ci` y publica la cobertura. Eso es lo que hace literalmente cierta la equivalencia de SC-004.
2. **Las herramientas de los controles son módulos Go pinados**, uno por herramienta bajo `tools/`,
   invocadas con `go tool -modfile=…`. Versión idéntica en local y en CI, integridad por `go.sum`,
   actualización por Dependabot, y ningún prerrequisito más allá de `go` y `git` (FR-042).
3. **El formato es un gate, no una corrección previa al gate**: `golangci-lint fmt` (mutante, en `make
   fmt` y en el pre-commit) y `golangci-lint fmt --diff` (verificación, en `make ci`) son dos modos de la
   misma herramienta con la misma configuración (FR-012).

Todo el código Go del hito vive en `cmd/kitlegal/` (`main.go` + `main_test.go`). **H0 no crea
`internal/`.**

## Technical Context

**Language/Version**: Go 1.27. `go.mod` lleva **dos** directivas: `go 1.27.0` (suelo del lenguaje) y
`toolchain go1.27.1` (el parche con el que se compila y se analiza; estable actual verificada contra el
proxy de módulos el 2026-09-11: go1.27.1 publicado el 2026-08-28, go1.27.2 inexistente), y es la
única fuente de verdad de la versión de Go. La directiva **por sí sola no fija el parche**: con
`GOTOOLCHAIN=auto` es un suelo, y una máquina con un parche más nuevo ejecutaría ese (verificado). Por eso
el **`Makefile` deriva de la directiva el valor de `GOTOOLCHAIN` y lo exporta a todas sus recetas**
(`GO_TOOLCHAIN := $(shell awk '$$1 == "toolchain" { print $$2; exit }' go.mod)` +
`export GOTOOLCHAIN := $(GO_TOOLCHAIN)`), que sí es un pin exacto en las dos direcciones y que el go
command obtiene solo si falta. Como una asignación del `Makefile` prevalece sobre el entorno heredado, el
pin manda también en la CI, por encima del `GOTOOLCHAIN=local` que exporta `actions/setup-go` (verificado);
la acción sigue instalando Go con `go-version-file: go.mod` y **sin** `go-version` ni `check-latest`, con
lo que instala ya go1.27.1 y el pin no cuesta ninguna descarga. Resultado: local y CI ejecutan el mismo
parche y `make vuln` —que analiza también la biblioteca estándar del toolchain— da el mismo veredicto a
ambos lados (SC-004, FR-042 a). Ver [research.md D1](./research.md).

**Corrección de la revisión final (2026-09-11, motivo [i] del juez B)**: la primera versión de este plan
fijaba `go 1.26.0` + `toolchain go1.26.6` y lo daba por «estable actual verificada» cuando lo verificado
era solo el `go` instalado en la máquina; en el momento de implementar (2026-09-10) ya estaba publicada
go1.27.1. Se sube la directiva `toolchain` a `go1.27.1`, la directiva `go` a `1.27.0` y los cuatro
`tools/*/go.mod` a `go 1.27.1`, y se deja `make ci` en verde con ese toolchain (todas las herramientas
fijadas soportan Go 1.27, verificado). Alternativas rechazadas —quedarse en `go1.26.8`, o subir solo
`toolchain` dejando `go 1.26.0`— y motivo en [research.md D1](./research.md), «Por qué go1.27.1». La
comprobación de «estable actual» contra el proxy de módulos, y no contra `go version`, es el
procedimiento que se deja escrito allí para la próxima vez.

**Primary Dependencies**: **ninguna en el módulo del producto.** El punto de entrada usa solo la
biblioteca estándar (`os`, `io`, `fmt`) y los tests solo `testing`. Kong (`alecthomas/kong`) es alcance de
H1; `testify` y `testscript`, de H1/H3. Ver [research.md D5](./research.md).

**Herramientas de control** (módulos Go pinados, fuera del `go.mod` del producto):
`github.com/golangci/golangci-lint/v2/cmd/golangci-lint` (v2.13.2, aporta también `gofumpt` y `goimports`
como *formatters*), `golang.org/x/vuln/cmd/govulncheck`, `github.com/zricethezav/gitleaks/v8` (v8.30.1),
`github.com/evilmartians/lefthook` (v1.13.6).

**Storage**: N/A en H0. SQLite (`modernc.org/sqlite`) entra en H3/H12.

**Testing**: `go test -race -shuffle=on -coverprofile=coverage.out ./...` con la biblioteca estándar,
table-driven y `t.Parallel()`. Tests de integración declarados (`-tags=integration`) sobre un conjunto
vacío. E2E (`testscript`) es H1.

**Target Platform**: binario multiplataforma sin cgo (`CGO_ENABLED=0`, `-trimpath`). CI en
`ubuntu-latest`. Desarrollo verificado en darwin/arm64. **Alcance de `CGO_ENABLED=0`**: se aplica solo en
las recetas de `build` e `install`, que son las que producen el binario distribuible; `test` y
`test-integration` **no** lo desactivan, porque el detector de carreras que exige FR-008 necesita cgo. Ni
`ci.yml` ni `nightly.yml` exportan `CGO_ENABLED` en el entorno del job: hacerlo dejaría `make test -race`
sin ejecutar el control en la CI mientras sigue ejecutándolo en local, rompiendo SC-004.

**Project Type**: herramienta de línea de órdenes (multicall a partir de H1); en H0, un único `main`.

**Performance Goals**: el flujo `ci` de una PR limpia por debajo de **3 minutos** con la caché de
dependencias y herramientas poblada (SC-006). El tiempo en frío se mide y se registra, no se asume
(ver [research.md D12](./research.md)).

**Constraints**: `make ci` no modifica ningún fichero del árbol de trabajo (SC-002). El código de H0 no
hace peticiones de red ni escribe fuera del directorio de construcción (FR-037), no contiene `panic` en
rutas de usuario (FR-038) y no contradice los códigos de salida estables (FR-039).

**Scale/Scope**: ~18 ficheros nuevos, ~80 líneas de Go de producto y test. El valor del hito está en la
configuración, no en el volumen.

## Constitution Check

*GATE: debe pasar antes de la fase 0 y volver a evaluarse tras la fase 1.*

### Principios

| # | Principio | Cómo lo cumple H0 | Veredicto |
|---|---|---|---|
| **I** | Fuentes públicas y frontera humana | H0 no toca ninguna fuente ni hace ninguna petición de red desde el código del producto (FR-037). No existe `internal/httpx` todavía; el binario no abre sockets. El ADR `0004-frontera-humana.md` deja registrada la regla antes de que exista el primer adaptador, y `0003-no-cendoj-masivo.md` hace lo propio con CENDOJ. | ✅ Cumple |
| **II** | Nada sin cita ni fuente | H0 no emite ningún dato legal, así que el sobre `{ok, fuente, url, fecha_consulta, hash, data}` no aplica: la salida de `version` es texto plano sobre el propio binario, no una afirmación legal. El sobre entra íntegro en H1 (`internal/core/schema`). No se anticipa (criterio §2). | ✅ No aplica, sin deuda |
| **III** | Tests primero y offline | Los tests de `run()` son la única forma de comprobar FR-003 y se escriben antes que el resto del `Makefile` (ver orden de implementación). Son 100 % offline: escritores en memoria, sin red, sin ficheros. `codecov.yml` declara desde ya los dos umbrales (≥ 70 % global, ≥ 85 % en `internal/core/**`) como estado que falla. El e2e `testscript` que exige la constitución para «cada hito» es alcance de H1 según `docs/ROADMAP.md`, y en H0 no habría binario con comportamiento que describir más allá de `version`, que ya cubre el test unitario. | ✅ Cumple |
| **IV** | Arquitectura hexagonal con reglas ejecutables | H0 no crea ningún paquete, así que no hay arquitectura que violar; lo que sí hace es **fijar el andamiaje que la hará ejecutable**: `.golangci.yml` con `forbidigo` acotado a `internal/**` (FR-014) y la regla «solo `cli` y `cmd/` llaman a `os.Exit`» respetada por construcción. `depguard`, `internal/arch_test.go` y el `forbidigo` completo entran en H1, tal y como asignan `docs/ROADMAP.md` §3 y FR-014. Ningún `panic` en rutas de usuario (FR-038): `run()` devuelve `int`, nunca entra en pánico. Códigos de salida: 0 y 2, ambos de la tabla estable. | ✅ Cumple |
| **V** | Simplicidad y dependencias fijadas (YAGNI) | El módulo del producto **no declara ninguna dependencia externa**. Ninguna de las herramientas de control está fuera de la lista de `docs/ROADMAP.md` §3. No hay framework de DI, ni ORM, ni generador de CLI: un `switch` sobre `os.Args` basta para un verbo. Todo lo que el spec declara fuera de alcance queda fuera del plan. | ✅ Cumple |
| **VI** | Un binario, convenciones de agente | Nombre único `kitlegal` para módulo, binario y directorio. El despacho multicall por `os.Args[0]`, las banderas globales y `--describe` son alcance de H1 y no se anticipan; H0 no introduce ninguna convención que H1 tenga que deshacer, porque `run(args, stdout, stderr) int` es exactamente la forma que el kernel CLI absorberá. Skills: ninguna en H0 (H5). | ✅ Cumple |
| **VII** | Grafo y privacidad | H0 no crea ninguna base de datos, ni `world.db`, ni `.kitlegal/case.db`, ni escribe ningún dato personal. `.gitignore` ya excluye `.kitlegal/` para que un asunto local nunca llegue por accidente al repositorio. | ✅ No aplica, sin deuda |

### Reglas de dependencia (`docs/ROADMAP.md` §2, constitución §IV)

| Regla | Situación en H0 | Cómo se hace cumplir | Veredicto |
|---|---|---|---|
| `internal/core/**` no importa `internal/{source,httpx,cache,store,graph,render,cli,app}` | Vacía: no existe `internal/`. | `depguard` + `internal/arch_test.go` en H1 (FR-014, `docs/ROADMAP.md` §3). | ✅ Sin violación posible |
| Solo `internal/httpx` importa `net/http` | Vacía: el código de H0 no importa `net/http` (FR-037). | Verificable hoy con `go list -deps ./...`; regla mecánica en H1. | ✅ Sin violación posible |
| Solo `internal/{cache,store,graph}` importan SQLite y `database/sql` | Vacía: sin dependencias externas ni `database/sql`. | `depguard` en H1. | ✅ Sin violación posible |
| Solo `internal/cli` y `cmd/` llaman a `os.Exit` | **Activa**: `os.Exit` aparece exactamente una vez, en `cmd/kitlegal/main.go`, dentro de `main()`. | Por construcción; `depguard` la hará mecánica en H1. `run()` devuelve `int` en lugar de salir, que es lo que la hace testable. | ✅ Cumple |
| Solo `internal/render` escribe en stdout; logs (`log/slog`) a stderr | **Parcial y deliberada**: `internal/render` no existe, así que la regla se aplica en H0 en su forma acotada —prohibir `fmt.Print*` bajo `internal/**`— que es literalmente el criterio de aceptación del hito. `cmd/` queda autorizado (*Assumptions* del spec). No hay logs en H0 (`log/slog` es H1). | `forbidigo` con `^fmt\.Print(\|f\|ln)$` y `path-except: '^internal/'` en `.golangci.yml`. Demostrable: SC-005. | ✅ Cumple en su forma H0 |

### Gates mecánicos (constitución «Gates», capa 1)

Los que este hito **activa**: `gofumpt`/`goimports` (modo verificación), `golangci-lint` (incluye `govet`
y `gosec`), `go test -race`, `govulncheck`, `gitleaks`, `go mod tidy -diff`, `go mod verify`, cobertura.

Los que **quedan pendientes por no tener objeto todavía**, cada uno con su hito asignado: test de
arquitectura sobre las reglas de dependencia (H1), validación de salidas contra `schemas/*.json` (H4/H11),
fixtures que fuerzan los exit codes 2-6 (H1), tests offline contra `testdata/` (H2/H4), corrección de
citas contra la respuesta grabada del BOE (H4), frontmatter de `SKILL.md` y diff vacío de `references/`
(H5). Ninguno se simula: `make schema-check`, `make test-e2e` y `make skills-sync` anuncian el objeto
ausente y el hito que lo aporta (FR-011).

**Veredicto del gate: PASA.** No hay violación sin justificar. Las seis divergencias conscientes entre
spec y plan —una por fila— están en *Complexity Tracking*.

### Re-evaluación tras la fase 1 (diseño)

El diseño de la fase 1 (`data-model.md`, `contracts/`, `quickstart.md`) no introduce ninguna capacidad
nueva sobre lo evaluado arriba. Los dos únicos elementos que aparecieron al bajar a diseño y que podrían
tocar un principio son:

- **`tools/` con cuatro módulos Go** (§V, simplicidad): estructura adicional que no era la opción por
  defecto. Su necesidad está verificada empíricamente —el módulo compartido no compila— y justificada en
  *Complexity Tracking*. No añade ninguna dependencia al `go.mod` del producto, que sigue vacío, así que
  §V se cumple con más holgura que la lista de dependencias permitidas.
- **Verbo desconocido → exit 2** (§IV, errores tipados que mapean a exit codes estables): usa un código de
  la tabla estable en su sentido literal («args»), sin inventar códigos ni contradecir ninguno (FR-039).
  El mapeo mediante errores tipados (`internal/cli/errors.go`) entra en H1 tal y como asigna
  `docs/ROADMAP.md`; en H0 hay un solo camino de error y una sola comparación.

**Veredicto tras el diseño: PASA**, sin nuevas violaciones ni nuevas entradas en *Complexity Tracking*.

## Project Structure

### Documentation (this feature)

```text
specs/001-h0-esqueleto-del-repo/
├── plan.md              # Este fichero
├── research.md          # Fase 0: 20 decisiones con alternativas y motivo
├── data-model.md        # Fase 1: entidades de configuración del hito
├── quickstart.md        # Fase 1: guía de validación ejecutable
├── contracts/
│   ├── cli-version.md   # Contrato de `kitlegal version` y de los códigos de salida
│   └── make-targets.md  # Contrato de las órdenes del Makefile
├── spec.md
├── checklists/
├── gates/
└── tasks.md             # Fase 2 (`/speckit-tasks`, no lo crea este comando)
```

### Source Code (repository root)

```text
go.mod                          módulo github.com/jmorenobl/kitlegal, go 1.27.0 + toolchain go1.27.1,
                                sin dependencias (no hay go.sum en la raíz: no hay dependencias que
                                sumar — ver D5; el porqué de las dos directivas, en D1). Única fuente
                                de verdad de la versión de Go: el Makefile lee de aquí el GOTOOLCHAIN

cmd/kitlegal/
├── main.go                     main() → os.Exit(run(...)); run(args, stdout, stderr) int
│                               vars version/commit/fecha inyectadas por -ldflags
└── main_test.go                package main, table-driven, t.Parallel()

tools/                          un módulo Go por herramienta de control (D2)
├── golangci-lint/{go.mod,go.sum}
├── govulncheck/{go.mod,go.sum}
├── gitleaks/{go.mod,go.sum}
└── lefthook/{go.mod,go.sum}

Makefile                        build test test-integration test-e2e lint fmt vuln schema-check
                                skills-sync release ci install (+ fmt-check lint-fast secrets
                                mod-verify mod-tidy-check check-tools hooks help). Exporta
                                GOTOOLCHAIN derivado de la directiva `toolchain` de go.mod (D1)
.golangci.yml                   v2, default: standard + los linters de FR-013 + forbidigo (FR-014)
lefthook.yml                    pre-commit: make fmt / lint-fast / secrets / mod-tidy-check
codecov.yml                     global ≥ 70 %, componente internal/core/** ≥ 85 %, no informativos
.gitignore                      ya existe (/bin/, *.out, coverage.*, .kitlegal/)
.gitleaksignore                 2 huellas justificadas de la documentación vendorizada (T001)

.github/
├── workflows/
│   ├── ci.yml                  PR y push a main → setup-go con go-version-file: go.mod (sin go-version
│   │                           ni check-latest; el parche efectivo lo fija el GOTOOLCHAIN que exporta
│   │                           el Makefile) → make ci + cobertura (requiere el secreto CODECOV_TOKEN)
│   ├── codeql.yml              semanal, languages: go, security-extended
│   └── nightly.yml             diario sobre main → make ci (mismo setup-go que ci.yml; sin red contra
│                               fuentes). Es el que hace visible un parche de Go pendiente vía govulncheck
└── dependabot.yml              semanal: gomod en / y en los 4 tools/*, github-actions en /

LICENSE                         Apache-2.0 (canónico, sin editar)
README.md                       qué es, cómo construir, cómo ejecutar los controles
CONTRIBUTING.md                 ritual por hito, controles, Conventional Commits, SemVer,
                                justificación de dependencias, órdenes con objeto ausente,
                                govulncheck sin red, exclusión de falsos positivos de gitleaks
CHANGELOG.md                    sección Unreleased
docs/ADR/
├── 0001-multicall.md
├── 0002-sqlite-sin-cgo.md
├── 0003-no-cendoj-masivo.md
└── 0004-frontera-humana.md
```

**`.gitleaksignore` sí forma parte de H0, con dos huellas reales.** La versión anterior de este plan
afirmaba lo contrario sobre una premisa falsa: que el árbol no tenía ningún falso positivo que excluir. Lo
tiene. La documentación vendorizada de `samber/cc-skills-golang` incluye dos líneas de ejemplo marcadas
`// DON'T` que la regla `generic-api-key` detecta y que no son credenciales; sin exclusión `make secrets`
no queda en verde y el control no queda instalado, que es lo que el hito entrega. El fichero nace con
contenido justificado —huella a huella, con comentario, sin desactivar ninguna regla—, no como marcador de
posición, así que SC-008 se cumple. T001 lo declara entre sus rutas y `CONTRIBUTING.md` (T010) sigue
documentando el procedimiento. FR-018 sigue satisfecho por la *ejecución* del control. Decisión humana del
2026-09-10 en [gates/decision-humana-secretos.md](./gates/decision-humana-secretos.md).

**Structure Decision**: layout `cmd/` + `internal/` estándar de Go para una herramienta de línea de
órdenes, con la particularidad de que **H0 no crea `internal/`** (FR-002, respuesta Q4 del `clarify`): el
directorio nace en H1 con el kernel CLI. La única estructura no obvia es `tools/<herramienta>/`, un módulo
Go por herramienta de control; el motivo —un módulo compartido no compila— está en
[research.md D2](./research.md) y justificado en *Complexity Tracking*.

## Controles mecánicos que este hito añade o toca

Enumerados según la sección «Gates» de la constitución. Cada fila indica qué orden lo ejecuta, si forma
parte del veredicto agregado, y **con qué comprobación concreta se demuestra que el control está vivo**
(SC-007: «falla si se retira el control»).

| # | Control | Herramienta | Orden | ¿En `make ci`? | Demostración de que está activo |
|---|---|---|---|---|---|
| 1 | Formato (verificación) | `golangci-lint fmt --diff` (gofumpt + goimports) | `fmt-check` | Sí | Un fichero `.go` sin formatear hace fallar `make ci` nombrando el fichero, y el árbol queda intacto (US1 esc. 6 y 7) |
| 2 | Formato (corrección) | `golangci-lint fmt` | `fmt` | **No** (mutaría el árbol) | Corrige el fichero anterior; lo usa el pre-commit |
| 3 | Lint | `golangci-lint run` con los 22 linters de FR-013 | `lint` | Sí | Un `err` sin comprobar hace fallar `errcheck` |
| 4 | Regla de arquitectura acotada | `forbidigo` en `internal/**` | `lint` | Sí | **SC-005**: PR de prueba desechable con `fmt.Println` bajo `internal/` falla; el mismo `fmt.Println` en `cmd/` pasa |
| 5 | Race detector | `go test -race` | `test` | Sí | La bandera está en la orden; su ausencia se ve en el diff del `Makefile` |
| 6 | Cobertura | `go test -coverprofile` + `codecov.yml` | `test` (+ publicación en `ci.yml`) | Sí (el perfil); el estado lo emite Codecov | **SC-011**: la revisión de H0 supera el 70 % con `main_test.go`; borrar el test baja del umbral |
| 7 | Vulnerabilidades | `govulncheck ./...` | `vuln` | Sí | Una dependencia con CVE conocido hace fallar la orden (US2 esc. 6). Analiza también la biblioteca estándar del toolchain que fija el `Makefile` a partir de `go.mod`, así que un parche de seguridad de Go pendiente aparece como hallazgo, y aparece igual en local que en la PR (D1, D19) |
| 8 | SAST (lint) | `gosec` dentro de `golangci-lint` | `lint` | Sí | Un `os.WriteFile` con permisos 0777 falla |
| 9 | SAST (semanal) | CodeQL, `security-extended` | — (flujo `codeql.yml`) | No (no opera sobre árbol de trabajo; SC-004 lo excluye) | Los resultados aparecen en la pestaña Security |
| 10 | Secretos | `gitleaks dir . --redact` | `secrets` | Sí | **SC-009**: un token en el diff bloquea la PR (US2 esc. 3) |
| 11 | Cadena de suministro | `go mod verify` (raíz + cada `tools/*/go.mod` **existente**, resuelto por glob y no por lista fija: son tres al terminar T001 y cuatro al terminar T003, y la orden está en verde en los dos momentos) | `mod-verify` | Sí | Alterar un módulo en la caché lo detecta |
| 12 | Cadena de suministro | Dependabot semanal (6 entradas) | — | No | PR automática al publicarse una versión nueva |
| 13 | Dependencias mínimas | `go mod tidy -diff` (módulo raíz) | `mod-tidy-check` | Sí | Un `require` sobrante hace fallar la orden señalando la diferencia (US2 esc. 4) |
| 14 | Esquemas | — (no hay `schemas/` todavía) | `schema-check` | Sí, pasa vacíamente anunciando el hito que lo aporta | **SC-012** |
| 15 | Pre-commit | `lefthook` → `make fmt/lint-fast/secrets/mod-tidy-check` | `hooks` (instalación) | No (la autoridad final es la CI) | `git commit` con un fichero sin formatear lo formatea y lo re-prepara |
| 16 | Prerrequisitos | `command -v go`/`git` + toolchain fijado obtenible (el `Makefile` exporta `GOTOOLCHAIN` desde la directiva) | `check-tools` (dependencia de las demás) | Sí, indirectamente | `make check-tools GO_TOOLCHAIN=go1.26.99` falla con un mensaje accionable que nombra el toolchain pedido, la directiva de la que sale y el error del go command; y `GOTOOLCHAIN=local make check-tools` **pasa**, porque el entorno no puede debilitar el pin (FR-010, D1, D18) |
| 17 | Commits y versiones | Conventional Commits + SemVer + `CHANGELOG.md` | — (documental) | No | `CONTRIBUTING.md` lo exige; la revisión de la PR lo comprueba |
| 18 | Decisiones | 4 ADR en `docs/ADR/` (MADR corto) | — (documental) | No | Existen y registran contexto, opciones y consecuencias |

**Fixtures**: H0 no crea `testdata/`. Los tests de `main_test.go` no necesitan fixtures (entradas y
salidas caben en la tabla de casos), y `testdata/<fuente>/` nace con el primer adaptador (H2/H4). Las
grabaciones con `KITLEGAL_RECORD=1` no existen todavía, así que la regla del modo desatendido
(«`KITLEGAL_RECORD=1` solo en un job separado y revisado») no tiene objeto en este hito.

**Tests de contrato**: los contratos de H0 son el de `kitlegal version` y el de las órdenes del
`Makefile` (ver `contracts/`). El primero se verifica con `main_test.go`; el segundo, con las
comprobaciones de [quickstart.md](./quickstart.md). La validación de salidas contra `schemas/*.json` que
pide la constitución no tiene objeto porque `version` no emite el sobre (*Assumptions* del spec); entra
en H4/H11.

## Orden de implementación

La constitución fija `core → adaptador → applet → skill`; en H0 no hay ninguna de esas capas, así que el
orden es el que hace que cada paso sea verificable por el anterior:

1. `go.mod` + `cmd/kitlegal/main.go` + `cmd/kitlegal/main_test.go` — **el test primero** (principio III):
   describe FR-003 antes de que exista el `Makefile` que lo construye.
2. Los tres módulos de `tools/` que `ci` necesita —`golangci-lint`, `govulncheck` y `gitleaks`— (`go get
   -tool` + `go mod tidy` en cada uno, ver [research.md D2.2](./research.md)). El cuarto,
   `tools/lefthook/`, llega en el paso 5 con `lefthook.yml`, y `mod-verify` lo cubre sin cambios porque
   recorre los modfiles de herramienta por glob (fila 11 de la tabla de controles).
3. `.golangci.yml` + `codecov.yml`.
4. `Makefile` completo, incluida la comprobación de prerrequisitos.
5. `lefthook.yml` + el módulo de herramienta `tools/lefthook/` (igual que los del paso 2) y `make hooks`.
6. Los tres flujos de `.github/workflows/` + `.github/dependabot.yml`.
7. Documentación: `LICENSE`, `README.md`, `CONTRIBUTING.md`, `CHANGELOG.md`, los cuatro ADR.
8. Verificación en PR de prueba desechable: una con `fmt.Println` bajo `internal/` (debe fallar, SC-005) y
   una limpia (debe pasar en < 3 min, SC-006), midiendo y registrando el tiempo en frío y en caliente.

## Complexity Tracking

> Divergencias conscientes respecto de la lectura literal del spec, y estructura añadida que no era la
> opción por defecto. Ninguna afecta a alcance, frontera humana, privacidad, TOS ni a una decisión
> cerrada, por lo que ninguna dispara el criterio §4 (escalar).

| Violación / divergencia | Por qué es necesaria | Alternativa más simple, y por qué se rechaza |
|---|---|---|
| **Cuatro módulos Go bajo `tools/`** en lugar de uno solo, como sugería la respuesta Q5 del `clarify` | Un único módulo compartido **no compila**: verificado que la Selección de Versión Mínima entre `golangci-lint` y `gitleaks` eleva `charmbracelet/x/ansi` a una versión incompatible con el `cellbuf` que gitleaks requiere, y `go tool … gitleaks` falla con una decena de errores de compilación. Aislado, funciona. | Un módulo compartido con `replace`/`exclude` que fuercen versiones compatibles: convierte cada bump de Dependabot en un ejercicio de resolución manual de conflictos y puede quedar sin solución si dos herramientas exigen versiones incompatibles. Un módulo por herramienta hace el choque estructuralmente imposible, a coste de tres directorios y tres entradas de Dependabot más. |
| **No existe `go.sum` en la raíz**, aunque FR-005 y SC-008 lo dan por hecho | El módulo del producto no tiene dependencias externas (Kong es H1, testify es H1/H3), y Go no genera `go.sum` para un módulo sin dependencias. El **fin** de FR-005/FR-019 —integridad verificable de todo lo que se descarga— lo cumplen los cuatro `tools/*/go.sum` y `go mod verify` sobre los cinco módulos dentro de `make ci`. | Añadir `testify` en H0 solo para que exista un `go.sum` en la raíz: es una dependencia sin necesidad funcional introducida para satisfacer la letra de un criterio (criterio §1: «ni atajos ni ñapas») y contradice «Fuera de alcance» del spec. |
| **`gofumpt` no tiene módulo propio en `tools/`**, aunque FR-042 lo enumera entre las herramientas a fijar | Lo aporta `golangci-lint` como *formatter*, con la versión fijada como dependencia indirecta explícita en `tools/golangci-lint/go.{mod,sum}`, donde Dependabot la ve. FR-042 (a)-(e) se cumplen todos por esa vía. | Pinarlo aparte: introduce **dos** versiones de gofumpt que pueden discrepar tras cualquier bump, de modo que `make fmt` y `make ci` podrían dar veredictos contrarios sobre el mismo fichero — exactamente lo que FR-012 y SC-004 existen para impedir. |
| **Ocho órdenes del `Makefile` más allá de las doce de FR-006** (`fmt-check`, `lint-fast`, `secrets`, `mod-verify`, `mod-tidy-check`, `check-tools`, `hooks` y `help`) | FR-006 dice «al menos». Las siete primeras son controles que FR-007/FR-010/FR-018/FR-019/FR-020/FR-022 exigen y que necesitan un nombre propio para poder invocarse desde `ci`, desde el pre-commit y desde la máquina de quien contribuye por separado. `help` no es un control ni entra en `ci`: es el objetivo por defecto que enumera los demás, y sostiene el «sin preguntar a nadie» de SC-010 junto a `README.md`. | Meter los comandos en línea dentro de la receta de `ci`: dejaría a quien contribuye sin forma de ejecutar un control aislado al depurarlo, y obligaría al pre-commit a duplicar la invocación fuera del `Makefile` (segundo camino, riesgo de divergencia). Para `help`, no tenerlo: obliga a leer el `Makefile` para saber qué se puede invocar. |
| **`misspell` corre solo con diccionario inglés**, aunque FR-013 y `docs/ROADMAP.md` §3 escriben «misspell (español e inglés)» | La herramienta **no tiene diccionario español**: su ajuste `locale` solo admite `US` y `UK` ([research.md D9](./research.md)). Se ejecuta con `locale: US` sobre todo el repositorio —comentarios y literales en español incluidos, que es donde aporta al detectar anglicismos mal escritos— y los falsos positivos que genere el español se neutralizan uno a uno en `misspell.ignore-rules`, cada uno con su comentario. En H0 no hay ninguno. Es un límite conocido de la herramienta, declarado, no un control debilitado. | Sustituir `misspell` por un corrector ortográfico con diccionario español (`cspell`, `hunspell`, `typos`): ninguno está en la lista de herramientas de `docs/ROADMAP.md` §3, todos rompen FR-042 (d) —dejan de obtenerse con la sola cadena Go— y `cspell` añadiría Node como prerrequisito externo, contra SC-010. Añadir uno de ellos **junto a** misspell es alcance nuevo que el spec no pide. |
| **`go mod tidy -diff` solo sobre el módulo raíz**, no sobre los de herramientas | En un módulo de herramientas lo que determina qué binario se ejecuta es la directiva `tool` más el `go.sum`; la «saneidad» de su `go.mod` es cosmética, y comprobarla obliga a descargar el grafo de dependencias *de test* de cada herramienta (verificado) encareciendo la CI. El riesgo real —una versión alterada a mano— lo cubre `go mod verify`. | Extenderlo a los cinco módulos: más tiempo de CI y más red, sin cubrir ningún riesgo que `go mod verify` no cubra ya. |

## Obligaciones que este plan traslada a `tasks.md`

1. **Medir el flujo `ci` en frío y en caliente** y dejar ambos números en la PR del hito. Lo exige
   *Assumptions* del spec de forma explícita. Si el escenario en caliente superara los 3 minutos, la
   contingencia es partir `make ci` en jobs paralelos, cada uno invocando una orden del `Makefile` —se
   corrige el diseño del flujo, nunca el criterio ([research.md D12](./research.md)).
2. **Comprobar `stage_fixed` de lefthook v1.13.6** cuando la orden no recibe `{staged_files}`; si no
   re-preparara los ficheros corregidos, pasar a `git add {staged_files}` explícito
   ([research.md D16](./research.md)).
3. **PR de prueba desechable** con `fmt.Println` bajo `internal/` para demostrar SC-005, y su gemela
   limpia para SC-006. Desechables: no se integran en `main`.
4. **Verificar si Dependabot (`gomod`) propone la subida de la directiva `toolchain`** de `go.mod`. Si la
   propone, la actualización del parche de Go queda cubierta por FR-028; si no, dejar escrito en
   `CONTRIBUTING.md` el procedimiento manual (subir `toolchain`, `make ci`, commit `chore(deps)`), cuyo
   disparador es el hallazgo de `govulncheck` en `ci`/`nightly` ([research.md D1](./research.md)).
   No se asume ninguna de las dos respuestas.
5. **Comprobar el estado del componente `internal/core` de Codecov** en la PR de prueba limpia: en H0 su
   `paths` no casa con ningún fichero y no debe aparecer ni como fallo ni como pendiente que bloquee la
   fusión. Si bloqueara, la corrección es `informational: true` **solo hasta H7**, con el motivo escrito
   en `codecov.yml`; nunca retirar el umbral ([research.md D15](./research.md)).
6. **Dar de alta el secreto `CODECOV_TOKEN`** en el repositorio y comprobar en la PR de prueba limpia que
   el paso de subida de cobertura termina bien. Las versiones actuales de `codecov/codecov-action` lo
   exigen incluso para ramas del propio repositorio, y con `fail_ci_if_error: true` su ausencia haría
   fallar el flujo `ci` **por la subida y no por un control**. La configuración de secretos de la
   plataforma queda fuera de lo que el repositorio puede contener, así que se declara aquí como
   prerrequisito; no se rebaja `fail_ci_if_error` para esquivarlo ([research.md D15](./research.md)).
