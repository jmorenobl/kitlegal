# Data Model: H0 · Esqueleto del repo y gates de CI

**Fecha**: 2026-09-10 · **Spec**: [spec.md](./spec.md) · **Plan**: [plan.md](./plan.md)

H0 **no tiene modelo de dominio**. No hay normas, ni citas, ni plazos, ni grafo: el hito no emite ningún
dato legal y por eso tampoco emite el sobre `{ok, fuente, url, fecha_consulta, hash, data}` (constitución
§II; *Assumptions* del spec). El dominio legal empieza en H1 (`internal/core/schema`) y H4
(`internal/source/boe`).

Lo que sí tiene H0 son **entidades de configuración**: las piezas que componen el veredicto del
repositorio. Se modelan aquí porque sus reglas de validación son las que `/speckit-tasks` tiene que
convertir en tareas comprobables.

---

## 1. `InformacionDeVersion`

La única estructura de datos del código de producto en H0.

| Campo | Tipo Go | Origen | Valor por defecto | Regla |
|---|---|---|---|---|
| `version` | `string` (var de paquete `main`) | `-ldflags -X main.version` | `"dev"` | El binario producido por `make build`/`make install` **nunca** muestra el valor por defecto (FR-004) |
| `commit` | `string` (var de paquete `main`) | `-ldflags -X main.commit` | `"none"` | SHA completo de `git rev-parse HEAD`; debe coincidir con el de la revisión construida (SC-003) |
| `fecha` | `string` (var de paquete `main`) | `-ldflags -X main.fecha` | `"unknown"` | Instante de construcción en UTC, formato RFC 3339 (`2026-09-10T09:12:33Z`) |

**Reglas de validación**

- R1.1 — Los tres campos residen en el paquete del punto de entrada (`cmd/kitlegal`), no en `internal/`
  (FR-002, FR-004).
- R1.2 — Los valores por defecto existen solo para `go run` y `go test`; su presencia en un binario
  construido por el `Makefile` es un defecto.
- R1.3 — Ninguno de los tres se persiste ni se envía a ninguna parte: `version` solo los escribe en
  stdout (FR-037).

**Transiciones de estado**: ninguna. Son constantes de construcción, inmutables en tiempo de ejecución.

---

## 2. `OrdenDelMakefile`

| Atributo | Valores |
|---|---|
| `nombre` | identificador de la orden |
| `clase` | `real` · `marcador-verde` · `marcador-rojo` |
| `en_ci` | si forma parte del veredicto agregado (FR-007) |
| `muta_el_arbol` | si puede modificar ficheros del árbol de trabajo |
| `codigo_de_salida` | 0 en caso de éxito; distinto de 0 al fallar |

**Reglas de validación**

- R2.1 — Ninguna orden es un no-op silencioso (FR-011). Un `marcador-verde` **imprime** el objeto ausente
  y el hito que lo aportará antes de terminar con 0.
- R2.2 — Ninguna orden con `en_ci = sí` tiene `muta_el_arbol = sí` (FR-007, SC-002). En particular, `fmt`
  (mutante) no está en `ci`; `fmt-check` (verificación) sí.
- R2.3 — `release` es el único `marcador-rojo`: termina con código ≠ 0 porque una acción con efectos
  externos no puede simular éxito, y no pertenece ni a `ci` ni a `nightly` (FR-011).
- R2.4 — Toda orden que necesite `go` o `git` depende de `check-tools`, que falla con un mensaje que
  identifica lo que falta y cómo obtenerlo (FR-010). La comprobación **no impone una versión de Go a quien
  contribuye**: el `Makefile` exporta `GOTOOLCHAIN` derivado de la directiva `toolchain` de `go.mod` (§7),
  de modo que toda receta corre ese parche sea cual sea el `go` instalado. Lo que `check-tools` verifica es
  que `go` (≥ 1.21) y `git` están presentes y que **el toolchain fijado es obtenible**
  ([research.md D1 y D18](./research.md)).
- R2.7 — El `GOTOOLCHAIN` que exporta el `Makefile` alcanza **todas** las recetas, incluidas las que operan
  sobre los módulos de herramienta (`go tool -modfile=tools/<n>/go.mod …`), de modo que los analizadores de
  `golangci-lint` se compilan con el mismo parche que compila el producto. Ninguna receta lo redefine ni lo
  desactiva.
- R2.6 — `CGO_ENABLED=0` se aplica **solo** en `build` e `install`. `test` y `test-integration` no lo
  desactivan, porque `-race` (FR-008) necesita cgo, y ningún flujo lo exporta en el entorno del job.
- R2.5 — Las cinco órdenes con objeto ausente están documentadas en `CONTRIBUTING.md` (FR-011).

El inventario completo con su comportamiento exacto está en
[`contracts/make-targets.md`](./contracts/make-targets.md).

---

## 3. `HerramientaDeControl`

| Atributo | Valor |
|---|---|
| `nombre` | nombre del binario tal como lo invoca el `Makefile` |
| `ruta_de_modulo` | ruta del módulo Go que la publica |
| `modfile` | `tools/<nombre>/go.mod` |
| `invocacion` | `go tool -modfile=tools/<nombre>/go.mod <nombre> …` |

| nombre | ruta_de_modulo | versión verificada |
|---|---|---|
| `golangci-lint` | `github.com/golangci/golangci-lint/v2/cmd/golangci-lint` | v2.13.2 |
| `govulncheck` | `golang.org/x/vuln/cmd/govulncheck` | última al crear el módulo |
| `gitleaks` | `github.com/zricethezav/gitleaks/v8` ⚠️ **no** `github.com/gitleaks/…` | v8.30.1 |
| `lefthook` | `github.com/evilmartians/lefthook` | v1.13.6 |

`gofumpt` y `goimports` no aparecen en la tabla: llegan como *formatters* de `golangci-lint` y su versión
queda fijada como dependencia indirecta explícita de `tools/golangci-lint/go.mod`
([research.md D3](./research.md)).

**Reglas de validación** (traducción directa de FR-042)

- R3.1 — Un módulo por herramienta. Compartir módulo entre herramientas está prohibido: rompe la
  compilación por conflicto de Selección de Versión Mínima ([research.md D2](./research.md)).
- R3.2 — La versión está fijada en `tools/<nombre>/go.mod` y es la misma en local y en CI, porque local y
  CI ejecutan la misma orden del `Makefile`.
- R3.3 — Cada `tools/<nombre>/go.sum` existe, no está vacío y `go mod verify -modfile=…` pasa.
- R3.4 — Cada `tools/<nombre>/go.mod` tiene una entrada `gomod` propia en `.github/dependabot.yml`.
- R3.5 — Ninguna de estas dependencias aparece en el `go.mod` del producto.
- R3.6 — Ninguna herramienta requiere instalación manual previa: el único prerrequisito es `go` (y `git`).
- R3.7 — Tras crear cada módulo con `go get -tool`, hay que ejecutar `go mod tidy -modfile=…` una vez, o
  el `go.sum` queda incompleto ([research.md D2.2](./research.md)).

---

## 4. `ControlDeCalidad`

Relaciona un control de `docs/ROADMAP.md` §3 marcado H0 con la herramienta que lo aplica, la orden que lo
ejecuta y **la comprobación que demuestra que sigue vivo**.

| Atributo | Descripción |
|---|---|
| `control` | nombre del control en `docs/ROADMAP.md` §3 |
| `herramienta` | quién lo aplica |
| `orden` | orden del `Makefile` que lo ejecuta (o `—` si no opera sobre un árbol de trabajo) |
| `en_ci` | si forma parte de `make ci` |
| `demostracion` | qué falla si se retira el control |

**Regla de validación**

- R4.1 — Los once controles de §3 marcados H0 tienen los cinco atributos rellenos y una `demostracion`
  concreta (SC-007). El inventario está en la tabla «Controles mecánicos» de [plan.md](./plan.md).
- R4.2 — Todo control con `en_ci = sí` produce el mismo veredicto ejecutado en local y en la PR (SC-004).
  Solo CodeQL y Dependabot quedan fuera de esa equivalencia, por no operar sobre un árbol de trabajo.

---

## 5. `FlujoDeIntegracionContinua`

| nombre | disparador | qué ejecuta | pasos propios permitidos |
|---|---|---|---|
| `ci` | `pull_request`, `push` a `main` | `make ci` | obtener el código, instalar Go, restaurar caché, publicar el perfil de cobertura |
| `codeql` | `schedule` semanal | acción `github/codeql-action` | los de la acción |
| `nightly` | `schedule` diario | `make ci` sobre `main` | los mismos que `ci` |

**Reglas de validación**

- R5.1 — El flujo `ci` **no aplica ningún control de FR-012 a FR-020 por una vía distinta de las órdenes
  del `Makefile`** (FR-025). Ni `golangci-lint-action`, ni `gitleaks-action`, ni `go test` suelto.
- R5.2 — El flujo `ci` conserva entre ejecuciones la caché de dependencias y de herramientas (FR-025), sin
  la cual SC-006 es inalcanzable.
- R5.3 — El flujo `nightly` no ejecuta ninguna verificación contra fuentes reales ni ningún control que no
  esté ya especificado en el hito (FR-027).
- R5.4 — `release` no forma parte de `ci` ni de `nightly` (FR-011).
- R5.5 — Permisos mínimos por job (`contents: read`; `security-events: write` solo en `codeql`).
- R5.6 — Los flujos `ci` y `nightly` instalan Go con `actions/setup-go` y `go-version-file: go.mod`, y
  **no** declaran `go-version` ni `check-latest`: la versión sale de la directiva `toolchain` del `go.mod`
  y el parche efectivo lo fija el `GOTOOLCHAIN` que exporta el `Makefile` (§7). Un flujo que declare la
  versión por su cuenta —o que exporte `GOTOOLCHAIN` en el entorno del job— introduce una segunda fuente
  de verdad y rompe R7.2.
- R5.7 — El paso de publicación de cobertura del flujo `ci` usa `fail_ci_if_error: true` y **requiere el
  secreto `CODECOV_TOKEN`** dado de alta en el repositorio. Es un prerrequisito de plataforma, no un
  artefacto del repositorio: sin él el flujo falla por la subida y no por un control, lo que enmascararía
  el veredicto. No se rebaja la bandera para esquivarlo ([research.md D15](./research.md)).

---

## 6. `DecisionDeArquitectura` (ADR)

| Atributo | Regla |
|---|---|
| `ruta` | `docs/ADR/NNNN-slug.md`, numeración correlativa desde `0001` |
| `formato` | MADR corto: contexto y problema · opciones consideradas · decisión · consecuencias |
| `contenido` | Ninguno queda vacío ni como marcador de posición (SC-008) |

Los cuatro de H0 (`0001-multicall`, `0002-sqlite-sin-cgo`, `0003-no-cendoj-masivo`, `0004-frontera-humana`)
**registran** decisiones ya cerradas en `CLAUDE.md` y en la constitución; no las reabren ni las matizan
(constitución, «Gobernanza»).

---

## 7. `CadenaDeHerramientasGo`

La versión de Go es una entidad de configuración más, con tres campos: dos que la **declaran** y uno que la
**hace efectiva**.

| Campo | Dónde vive | Valor en H0 | Qué significa |
|---|---|---|---|
| directiva `go` | `go.mod` | `1.27.0` | Suelo del lenguaje: la API y las reglas de compilación que el código puede usar |
| directiva `toolchain` | `go.mod` | `go1.27.1` | Parche declarado (la estable actual, verificada contra el proxy de módulos el 2026-09-11). **Es un suelo, no un pin**: con `GOTOOLCHAIN=auto` un Go local más nuevo se impone a la directiva ([research.md D1](./research.md), verificado) |
| `GOTOOLCHAIN` | exportado por el `Makefile`, derivado de la directiva `toolchain` | `go1.27.1` | Parche **efectivo**: el que compila, testea y analiza, idéntico en local y en la integración continua |

La derivación es `GO_TOOLCHAIN := $(shell awk '$$1 == "toolchain" { print $$2; exit }' go.mod)` seguida de
`export GOTOOLCHAIN := $(GO_TOOLCHAIN)`. Hay **una sola fuente de verdad** —la directiva— y el `Makefile`
no la duplica: la lee.

**Reglas de validación**

- R7.1 — Ambas directivas existen. La `toolchain` nombra el **parche estable actual** en el momento de
  implementar el hito (FR-001, *Assumptions* del spec).
- R7.2 — El parche que se ejecuta en local y el que se ejecuta en la CI son **el mismo**, y lo garantiza el
  `GOTOOLCHAIN` que exporta el `Makefile`, no la directiva por sí sola: un `GOTOOLCHAIN` con nombre exacto
  ejecuta ese toolchain sea más nuevo o más viejo que el local, y lo descarga si falta. En la CI el pin
  prevalece sobre el `GOTOOLCHAIN=local` que exporta `actions/setup-go`, porque una asignación del
  `Makefile` gana al entorno heredado (verificado, [research.md D1](./research.md)). Es condición de
  SC-004 y de FR-042 (a), porque `govulncheck` analiza también la biblioteca estándar del toolchain en uso.
- R7.3 — Subir de parche es un cambio versionado y revisable: un commit sobre la directiva `toolchain`, y
  **solo sobre ella** —el `Makefile` no repite el número—. Su disparador es un hallazgo de `make vuln` en
  `ci` o en `nightly`; el procedimiento está en `CONTRIBUTING.md`.
- R7.4 — Ningún flujo de `.github/workflows/` declara la versión de Go por su cuenta (R5.6), y ninguno
  exporta `GOTOOLCHAIN` en el entorno del job: el pin sale del `Makefile` en los dos lados.
- R7.5 — El único prerrequisito externo sigue siendo `go` (≥ 1.21, la primera versión capaz de cambiar de
  toolchain) y `git`. **Ningún parche concreto de Go es prerrequisito**: quien contribuye puede tener
  cualquier `go` ≥ 1.21, y el toolchain fijado se obtiene y se verifica solo (FR-010, FR-042 d, SC-010).
- R7.6 — El `Makefile` falla al leerse si la directiva `toolchain` no existe en `go.mod`
  (`$(error …)`), en lugar de exportar un `GOTOOLCHAIN` vacío y volver silenciosamente al comportamiento
  por defecto. Un pin que se desactiva sin avisar sería un control convertido en adorno.

---

## Inventario de artefactos (SC-008)

Enumeración completa, sin cifra, de lo que la revisión de H0 debe contener. Cada entrada, con contenido
real.

| Artefacto | Requisito |
|---|---|
| `go.mod` (directivas `go 1.27.0` y `toolchain go1.27.1`, §7) | FR-001, FR-042 (a) |
| `Makefile` → `export GOTOOLCHAIN` derivado de la directiva `toolchain` (§7, R7.2) | FR-001, FR-042 (a) |
| `cmd/kitlegal/main.go` | FR-002, FR-003, FR-004 |
| `cmd/kitlegal/main_test.go` | FR-040 |
| `tools/{golangci-lint,govulncheck,gitleaks,lefthook}/go.mod` + `go.sum` | FR-005, FR-042 |
| `Makefile` | FR-006 a FR-011, FR-018 a FR-020, FR-041 |
| `.golangci.yml` | FR-012, FR-013, FR-014, FR-017 |
| `lefthook.yml` | FR-018, FR-022 |
| `codecov.yml` | FR-029 |
| `.github/workflows/ci.yml` | FR-025 |
| `.github/workflows/codeql.yml` | FR-026 |
| `.github/workflows/nightly.yml` | FR-027 |
| `.github/dependabot.yml` | FR-028 |
| `LICENSE` | FR-030 |
| `README.md` | FR-031 |
| `CONTRIBUTING.md` | FR-021, FR-032 |
| `CHANGELOG.md` | FR-023 |
| `docs/ADR/0001-multicall.md` | FR-033 |
| `docs/ADR/0002-sqlite-sin-cgo.md` | FR-034 |
| `docs/ADR/0003-no-cendoj-masivo.md` | FR-035 |
| `docs/ADR/0004-frontera-humana.md` | FR-036 |
| `.gitleaksignore` | FR-018 |

**Ausencias deliberadas**

- No hay `go.sum` en la raíz porque el módulo del producto no tiene dependencias externas. La integridad
  que FR-005 persigue la aportan los cuatro `tools/*/go.sum` más `go mod verify` sobre los módulos
  existentes dentro de `make ci` (ver *Complexity Tracking* en [plan.md](./plan.md) y
  [research.md D5](./research.md)).
- ~~No hay `.gitleaksignore`.~~ **Sí lo hay**, con dos huellas reales: el árbol heredado trae dos líneas
  de ejemplo marcadas `// DON'T` en la documentación vendorizada de `samber/cc-skills-golang` que la regla
  `generic-api-key` detecta. No es un marcador de posición vacío, así que SC-008 se cumple. FR-018 sigue
  cubierto por la **ejecución** del control (`Makefile`, orden `secrets` dentro de `ci`, y `lefthook.yml`
  en pre-commit) y `CONTRIBUTING.md` (FR-021, FR-032) sigue documentando el procedimiento huella a huella
  ([research.md D10](./research.md), [gates/decision-humana-secretos.md](./gates/decision-humana-secretos.md)).
  El párrafo que sigue describe la situación anterior a esa decisión y se conserva como rastro; crear el
  fichero **vacío** seguiría siendo el marcador de posición sin contenido que SC-008
  prohíbe, y contradiría «lo no especificado no se implementa». Coincide con la decisión de
  [tasks.md](./tasks.md) («Notas»), que es la fuente que ejecuta el workflow.
