# Contrato: órdenes del `Makefile`

**Fecha**: 2026-09-10 · **Requisitos**: FR-006 a FR-011, FR-041, FR-042 · **Criterios**: SC-001, SC-002,
SC-004, SC-012

El `Makefile` es la **única superficie de invocación de los controles**: lo que ejecuta quien contribuye,
lo que ejecuta el gancho de pre-commit y lo que ejecuta la integración continua son la misma orden. Esa
identidad es lo que hace cierta la equivalencia local ↔ CI de SC-004; romperla en cualquier punto
invalida el criterio.

Notación: `$(TOOL,<n>)` abrevia `go tool -modfile=tools/<n>/go.mod`.

---

## Órdenes reales

| Orden | Comando | Muta el árbol | En `ci` | Requisito |
|---|---|---|---|---|
| `build` | `CGO_ENABLED=0 go build -trimpath -ldflags "$(LDFLAGS)" -o bin/kitlegal ./cmd/kitlegal` | Sí (`bin/`, ignorado por git) | No | FR-009, FR-004 |
| `install` | `CGO_ENABLED=0 go install -trimpath -ldflags "$(LDFLAGS)" ./cmd/kitlegal` | No (escribe en `$GOBIN`) | No | FR-041 |
| `test` | `go test -race -shuffle=on -coverprofile=coverage.out ./...` | Sí (`coverage.out`, ignorado por git) | **Sí** | FR-008, FR-015 |
| `test-integration` | `go test -race -tags=integration ./...` | No | No | FR-011 |
| `lint` | `$(TOOL,golangci-lint) golangci-lint run ./...` | No | **Sí** | FR-013, FR-014, FR-017 |
| `lint-fast` | `$(TOOL,golangci-lint) golangci-lint run --fast-only ./...` | No | No (pre-commit) | FR-022 |
| `fmt` | `$(TOOL,golangci-lint) golangci-lint fmt ./...` | **Sí** | **No** | FR-012 |
| `fmt-check` | `$(TOOL,golangci-lint) golangci-lint fmt --diff ./...` | No | **Sí** | FR-012, FR-007 |
| `vuln` | `$(TOOL,govulncheck) govulncheck ./...` | No | **Sí** | FR-016 |
| `secrets` | `$(TOOL,gitleaks) gitleaks dir . --redact --no-banner` | No | **Sí** | FR-018 |
| `mod-verify` | `go mod verify` en la raíz y en cada `tools/*/go.mod` **existente** (glob, no lista fija) | No | **Sí** | FR-019 |
| `mod-tidy-check` | `go mod tidy -diff` (módulo raíz) | No | **Sí** | FR-020 |
| `check-tools` | Comprueba `go` (≥ 1.21), `git` y que el toolchain fijado es obtenible | No | Sí (dependencia) | FR-010 |
| `hooks` | `$(TOOL,lefthook) lefthook install` | Sí (`.git/hooks`) | No | FR-022 |
| `ci` | `fmt-check lint test vuln schema-check secrets mod-verify mod-tidy-check` | **No** | — | FR-007 |
| `help` | Enumera las órdenes de este contrato con su descripción; objetivo por defecto | No | No | SC-010 |

**Alcance de `mod-verify`**: la lista de modfiles se **descubre**, no se escribe. Con una lista fija la
orden fallaría en el momento en que un módulo de herramienta todavía no exista —los tres primeros llegan
con T001 y `tools/lefthook/` con T003— y también quedaría muda si alguien añadiera un módulo nuevo sin
tocar el `Makefile`, que es justo el caso en que un control debe seguir vivo:

```make
TOOL_MODULES := $(patsubst %/go.mod,%,$(wildcard tools/*/go.mod))

mod-verify: check-tools
	go mod verify
	@for d in $(TOOL_MODULES); do (cd "$$d" && go mod verify) || exit 1; done
```

**Verificado**: `go mod verify` **no acepta** `-modfile` (`go help mod verify` no declara ninguna
bandera), a diferencia de `go tool`, así que cada módulo de herramienta se verifica entrando en su
directorio. El `GOTOOLCHAIN` que exporta el `Makefile` se hereda en el subshell, de modo que los cinco
módulos se verifican con el mismo toolchain fijado.

**`LDFLAGS`**: `-X main.version=$(VERSION) -X main.commit=$(COMMIT) -X main.fecha=$(FECHA)`, con

```make
VERSION ?= $(shell git describe --tags --always --dirty 2>/dev/null || echo 0.0.0-dev)
COMMIT  ?= $(shell git rev-parse HEAD 2>/dev/null || echo desconocido)
FECHA   ?= $(shell date -u +%Y-%m-%dT%H:%M:%SZ)
```

`build` e `install` aplican **los mismos** `LDFLAGS` (FR-041).

**Alcance de `CGO_ENABLED=0`**: la variable se fija **solo** en las recetas de `build` e `install` —las que
producen el binario distribuible, según *Restricciones técnicas* de la constitución—, y precede al comando
en la propia receta, no se exporta a nivel de `Makefile`. `test` y `test-integration` **no** la desactivan:
`go test -race` requiere cgo y con `CGO_ENABLED=0` termina en error, de modo que desactivarla globalmente
dejaría sin control el detector de carreras que exige FR-008. Por el mismo motivo, `ci.yml` y `nightly.yml`
**no** exportan `CGO_ENABLED` en el entorno del job: se limitan a invocar `make ci`, y cada receta decide.

**Versión de Go**: el `Makefile` **exporta `GOTOOLCHAIN`** con el valor de la directiva `toolchain` de
`go.mod`, leído del propio fichero:

```make
GO_TOOLCHAIN := $(shell awk '$$1 == "toolchain" { print $$2; exit }' go.mod)

ifeq ($(GO_TOOLCHAIN),)
$(error go.mod no declara una directiva `toolchain`)
endif

export GOTOOLCHAIN := $(GO_TOOLCHAIN)
```

Con eso, **todas** las órdenes de este contrato —incluidas las que operan sobre los módulos de
herramienta— se ejecutan con `go1.27.1` exacto, en local y en la integración continua. La directiva sola
no bastaría: con `GOTOOLCHAIN=auto` es un suelo, y una máquina con un parche más nuevo ejecutaría ese. Un
`GOTOOLCHAIN` con nombre exacto, en cambio, ejecuta ese toolchain sea más nuevo o más viejo que el local, y
lo descarga y verifica si falta. En la CI el pin prevalece sobre el `GOTOOLCHAIN=local` que exporta
`actions/setup-go`, porque una asignación del `Makefile` gana al entorno heredado. Es lo que hace que
`make vuln`, que analiza también la biblioteca estándar del toolchain, dé el mismo veredicto en los dos
sitios (SC-004, FR-042 a; [research.md D1](../research.md)).

**Fuente de verdad única**: el número de versión aparece **solo** en la directiva `toolchain` de `go.mod`.
Subir de parche es editar esa línea; el `Makefile` no se toca. Para probar el camino de fallo de
`check-tools` sin tocar `go.mod` se puede sobrescribir la variable derivada en la línea de órdenes
(`make check-tools GO_TOOLCHAIN=go1.26.99`), que es lo que hace el escenario 7 de
[`quickstart.md`](../quickstart.md).

---

## Órdenes con el objeto todavía ausente

Ninguna es un no-op silencioso (FR-011).

| Orden | Comportamiento | Salida exacta | Código |
|---|---|---|---|
| `test-e2e` | Anuncia y termina bien | `test-e2e: sin tests e2e todavía; los aporta H1 (testscript)` | 0 |
| `schema-check` | Anuncia y termina bien | `schema-check: no hay schemas/ todavía; los aportan H4 (borrador) y H11 (contrato)` | 0 |
| `skills-sync` | Anuncia y termina bien | `skills-sync: no hay skills/ ni data/*.yaml todavía; los aporta H5` | 0 |
| `release` | **Falla** | `release: sin configurar hasta H6 (.goreleaser.yaml)` | ≠ 0 |

`release` es la excepción porque es una acción con efectos externos: no puede simular éxito. **No** forma
parte de `ci` ni de `nightly`.

Las cuatro, más `test-integration`, se documentan en `CONTRIBUTING.md` indicando que existen pero reciben
su contenido en un hito posterior (FR-011).

---

## Contrato de `ci`

`make ci` es el veredicto del repositorio.

**Contenido** (FR-007): comprobación de formato · lint (incluye `gosec` y `govet`) · tests con detector de
carreras · análisis de vulnerabilidades · comprobación de esquemas · detección de secretos · verificación
de la integridad de los módulos · comprobación de dependencias saneadas.

**Invariantes**

- M1 — **`ci` no modifica ningún fichero versionado del árbol de trabajo.** `git status --porcelain` antes
  y después produce la misma salida (SC-002, US1 esc. 7). Los ficheros que genera (`coverage.out`) están
  ya cubiertos por `.gitignore`.
- M2 — `ci` falla si falla cualquiera de sus ocho controles, y el mensaje identifica cuál.
- M3 — `ci` ejecuta el modo de **verificación** del formato, nunca el mutante (FR-012). Corregir el árbol
  para poder aprobarlo no es un gate.
- M4 — `ci` no invoca `release`.
- M5 — La integración continua ejecuta **esta misma orden**, sin aplicar ningún control por otra vía
  (FR-025). Los pasos propios del flujo se limitan a preparar el entorno y publicar la cobertura.

---

## Contrato de prerrequisitos

`check-tools` es dependencia de toda orden que necesite la cadena de herramientas.

| Situación | Comportamiento |
|---|---|
| Falta `go` | Falla nombrando `go`, la versión mínima (**1.21**, la primera capaz de cambiar de toolchain) y `https://go.dev/dl/` |
| Falta `git` | Falla nombrando `git` y cómo instalarlo |
| El toolchain fijado no se puede obtener | Falla nombrando el toolchain pedido, la directiva `toolchain` de `go.mod` de la que sale y el error literal del go command (p. ej. `toolchain not available`, o `go.mod requires go >= …`) |

**Lo que `check-tools` NO hace: exigir que quien contribuye tenga instalado un parche concreto de Go.**
Como el `Makefile` exporta `GOTOOLCHAIN`, el parche correcto lo garantiza la propia orden; cualquier `go`
≥ 1.21 sirve. La tercera fila no es una precondición impuesta al contribuidor sino la **postcondición** del
pin: lo único que puede fallar es que el toolchain fijado no esté disponible —sin red, `GOPROXY=off`, un
`go` anterior a 1.21, o una directiva que nombre una versión inexistente— y entonces conviene que el
diagnóstico salga de aquí y no del primer control que se tropiece. Con esto FR-010 se cumple en su letra
(«presencia **y versión**»): presencia en las dos primeras filas, versión en la tercera
([research.md D1 y D18](../research.md)).

Un `GOTOOLCHAIN` heredado del entorno **no** es un modo de fallo: la asignación del `Makefile` prevalece
sobre él, así que `GOTOOLCHAIN=local make check-tools` pasa igual.

Tras la decisión de construir las herramientas con `go tool`
([research.md D2](../research.md)), **`go` y `git` son los únicos prerrequisitos externos del proyecto**
(FR-010, FR-042 d, SC-010). No hay que instalar `golangci-lint`, `govulncheck`, `gitleaks` ni `lefthook`,
ni tampoco el parche concreto de Go: el `GOTOOLCHAIN` exportado hace que se descargue y se verifique solo.

---

## Comportamiento sin red

`vuln` consulta la base de datos de vulnerabilidades de Go. **Sin red, `make vuln` falla**, y ese fallo
no debe confundirse con «no hay vulnerabilidades» (FR-016, «Edge Cases» del spec). El comportamiento está
documentado en `CONTRIBUTING.md`. No se captura el error para devolver 0 con un aviso: eso convertiría un
control en un adorno.

La primera ejecución de cualquier orden que use una herramienta descarga y compila esa herramienta, lo que
también requiere red. A partir de ahí, la caché de construcción de Go la sirve sin conexión. Lo mismo vale
para el toolchain que fija `GOTOOLCHAIN`: si el `go` local no es ya ese parche y el toolchain no está
descargado, la primera orden lo obtiene —`check-tools` falla con el mensaje del cuadro anterior si eso no
es posible—, y después queda en la caché de toolchains de Go. Cuando el `go` que se ejecuta **ya es** el
parche fijado, el go command no descarga nada: es el caso de la integración continua, donde
`actions/setup-go` instala precisamente go1.27.1, y el del desarrollo local en cuanto el parche coincide.
