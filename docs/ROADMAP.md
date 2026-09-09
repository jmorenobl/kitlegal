# kitlegal: roadmap de hitos

Roadmap para implementar lo descrito en `refs/`. Enfoque lean: hitos de 0,5 a 3 días, cada uno termina en algo que se puede usar (o que protege lo que ya se usa). No hay fechas; hay orden y dependencias. Lo que no está aquí no se construye hasta que un hito lo necesite (YAGNI).

Las decisiones cerradas de `CLAUDE.md` y `refs/00-README.md` no se repiten: este documento las asume.

---

## 0. Cómo leer este roadmap

- **Fase** = épica. Al final de cada fase hay un usuario real que gana algo: tú en Claude Code, luego un pack instalable, luego terceros.
- **Hito** = PR única, rama corta desde `main` (trunk-based), squash-merge con CI verde. Un hito no se cierra sin cumplir la *Definition of Done* (§1).
- **Ahora / Siguiente / Después**: solo la fase en curso se planifica al detalle. Las fases lejanas son intencionadamente esquemáticas y se refinan al llegar.
- Cada hito indica: objetivo, entrega usable, alcance, controles (tests y calidad) y criterio de aceptación.

---

## 1. Definition of Done (aplica a todos los hitos)

1. `make ci` en verde: `fmt`, `lint`, `test` (con `-race`), `vuln`, `schema-check`.
2. Tests unitarios offline para el código nuevo; fixtures grabados en `testdata/<fuente>/` si toca red.
3. Sin `net/http`, `os.Exit`, `fmt.Print*` fuera de los paquetes autorizados (lo vigila `depguard`/`forbidigo`, ver §3).
4. Toda salida de applet pasa la validación contra `schemas/*.json` en test.
5. Errores tipados que mapean a un exit code estable; ningún `panic` en rutas de usuario.
6. Si cambia un comportamiento visible: test e2e (`testscript`) actualizado y `CHANGELOG.md` (sección *Unreleased*).
7. Si cambia una decisión de arquitectura: ADR nuevo en `docs/ADR/`.
8. Si toca una fuente externa: fila de `docs/SOURCES.md` actualizada (licencia, TOS, rate limit, fecha de verificación) y caso en `scripts/verify-sources.sh`.
9. Cobertura de `internal/core/**` ≥ 85 % y global ≥ 70 % (umbrales en `codecov.yml`; no son objetivo, son red de seguridad).

---

## 2. Arquitectura objetivo (resumen operativo)

Arquitectura hexagonal ligera. Sin framework de inyección: composición manual en `internal/app` con functional options.

```
                ┌────────────────────────── cmd/kitlegal (main, multicall)
                │
        internal/app ── composition root, registro de applets (Command + Registry)
                │
        internal/cli ── Kong, flags globales, mapeo error→exit code, sobre de salida
                │
  ┌─────────────┼─────────────────────────────────────────────┐
  │ internal/core (dominio puro, sin I/O)                     │
  │   ids · cita · plazos · competencia · schema              │
  │   define los PUERTOS: Source, Fetcher, Cache, GraphStore, │
  │   Renderer                                                │
  └─────────────┬─────────────────────────────────────────────┘
                │ implementan
   ADAPTADORES: internal/source/<fuente>  (Adapter, uno por fuente)
                internal/httpx            (Decorator chain: UA → robots → ratelimit → retry → record/replay)
                internal/cache, store, graph (Repository sobre SQLite)
                internal/render           (Strategy: json | table | markdown)
```

Patrones que se usan y por qué:

| Patrón | Dónde | Motivo |
|---|---|---|
| Adapter | `internal/source/<x>` | Cada fuente expone la misma interfaz `Source`; cambiar BOE no toca PLACSP |
| Command + Registry | `internal/app` | Un applet = un comando registrado; el multicall solo despacha |
| Decorator / middleware | `internal/httpx` | Retries, rate limit, robots y grabación componen sin acoplarse |
| Strategy | `internal/render` | `--json` manda; tabla y markdown son intercambiables |
| Repository | `cache`, `store`, `graph` | El dominio no sabe que hay SQLite |
| Null Object | `graph` con `--no-graph` | Los applets siempre llaman a `Emit`; el store nulo descarta |
| Errores tipados (sentinel + `errors.As`) | `internal/cli/errors.go` | `ErrNotFound`→3, `ErrSourceDown`→4, `ErrRateLimited`→5, `ErrHumanRequired`→6 |
| Functional options | constructores de `httpx`, `cache`, applets | Configuración explícita y testeable sin globals |
| Facade | `pkg/legalkit` (fase 4) | API pública pequeña sobre los internos |

Reglas de dependencia (se hacen cumplir con `depguard`, no con buena voluntad):

- `internal/core/**` no importa nada de `internal/{source,httpx,cache,store,graph,render,cli,app}`.
- Solo `internal/httpx` importa `net/http`.
- Solo `internal/{cache,store,graph}` importan `modernc.org/sqlite` y `database/sql`.
- Solo `internal/cli` y `cmd/` llaman a `os.Exit`.
- Solo `internal/render` escribe en stdout; logs (`log/slog`) siempre a stderr.

---

## 3. Estrategia de calidad y controles

Qué control, con qué herramienta, y en qué hito entra. Todo lo que entra se queda como gate de CI.

| Control | Herramienta | Entra en |
|---|---|---|
| Formato | `gofumpt` + `goimports` (vía golangci-lint formatters) | H0 |
| Lint | `golangci-lint` con `.golangci.yml` estricto: `errcheck`, `govet`, `staticcheck`, `revive`, `gocritic`, `errorlint`, `exhaustive`, `nilerr`, `bodyclose`, `noctx`, `contextcheck`, `sqlclosecheck`, `rowserrcheck`, `nolintlint`, `misspell` (es/en), `gocyclo`, `dupl`, `testifylint`, `thelper`, `paralleltest`, `tparallel`, `gosec` | H0 |
| Reglas de arquitectura | `depguard` + `forbidigo` (§2) y un test `internal/arch_test.go` que recorre `go list -deps` | H1 |
| Tests unitarios | stdlib + `testify`; table-driven; `t.Parallel()`; fixtures en `testdata/`; golden files para JSON de salida | H1 |
| Grabación/replay HTTP | `httpx` con `KITLEGAL_RECORD=1` (graba a `testdata/<fuente>/`); en test, replay determinista | H2 |
| Race detector | `go test -race ./...` en CI | H0 |
| Fuzzing | `go test -fuzz` sobre parsers: `ids`, `cita`, parseo XML/Atom; corpus mínimo versionado en `testdata/fuzz/` | H7, H8, H16 |
| Property tests | `testing/quick` o `pgregory.net/rapid` en `plazos` (idempotencia, monotonía, ida y vuelta) | H9 |
| Tests de integración | `//go:build integration`, SQLite en `t.TempDir()`, `httptest.Server` como fuente | H3 |
| Tests e2e | `github.com/rogpeppe/go-internal/testscript` contra el binario compilado: `kitlegal …`, `boe …` vía symlink, exit codes, `--json`, `--offline` | H1 |
| Tests de contrato | Cada salida validada contra `schemas/*.json` (`santhosh-tekuri/jsonschema`); `--describe` genera el schema y CI falla si `schemas/` está desactualizado | H11 (borrador en H4) |
| Smoke contra red | `scripts/verify-sources.sh` en CI nightly; abre issue si una fuente cambia | H4 |
| Evals de skills | `evals/<skill>/*.yaml` (pregunta, comandos esperados, citas esperadas); job semanal con modelo barato | H5 |
| Vulnerabilidades | `govulncheck` en cada PR | H0 |
| SAST | `gosec` (en golangci) + CodeQL semanal | H0 |
| Secretos | `gitleaks` en pre-commit y CI | H0 |
| Cadena de suministro | `go mod verify`, `go.sum` obligatorio, Dependabot semanal, SBOM (syft) + firma (cosign) + provenance en release | H0 / H6 |
| Dependencias mínimas | `go mod tidy -diff` en CI; una dependencia nueva requiere justificación en la PR | H0 |
| Observabilidad local | `log/slog` a stderr, `--verbose`, `KITLEGAL_LOG=debug`; `--dry-run` imprime la petición sin ejecutarla | H1 |
| Pre-commit | `lefthook`: `gofumpt`, `golangci-lint run --fast`, `gitleaks`, `go mod tidy -diff` | H0 |
| Commits / versiones | Conventional Commits, SemVer, `CHANGELOG.md` generado por goreleaser | H0 / H6 |
| Docs de decisiones | `docs/ADR/NNNN-*.md` (formato MADR corto) | H0 |

Dependencias fijadas (fuera de esta lista, justificar): `alecthomas/kong`, `modernc.org/sqlite`, `stretchr/testify`, `rogpeppe/go-internal` (testscript), `golang.org/x/time/rate`, `temoto/robotstxt`, `gopkg.in/yaml.v3`, `invopop/jsonschema` (generación), `santhosh-tekuri/jsonschema` (validación en tests). En fase 4: `modelcontextprotocol/go-sdk`.

---

## 4. Hitos

### Fase 0 — Fundación y primer hito (Ahora)

Al terminar: `kitlegal boe articulo BOE-A-2015-10565 a21` funciona con caché, la skill `boe-fiscal` usa el binario en Claude Code y hay un release v0.1.0 firmado.

#### H0 · Esqueleto del repo y gates de CI
- **Objetivo**: un repo vacío pero blindado. Cualquier línea de Go que entre después pasa por los controles.
- **Entrega**: `make build test lint vuln` en verde; `kitlegal version` imprime versión, commit y fecha.
- **Alcance**: `go.mod` (módulo `github.com/jmorenobl/kitlegal`, Go estable actual), `cmd/kitlegal/main.go` mínimo, `Makefile` (`build test test-integration test-e2e lint fmt vuln schema-check skills-sync release ci install`), `.golangci.yml`, `lefthook.yml`, `.github/workflows/{ci,codeql,nightly}.yml`, `.github/dependabot.yml`, `codecov.yml`, `LICENSE` (decisión open-core: Apache-2.0 en el núcleo), `README.md`, `CONTRIBUTING.md`, `CHANGELOG.md`, `docs/ADR/0001-multicall.md`, `0002-sqlite-sin-cgo.md`, `0003-no-cendoj-masivo.md`, `0004-frontera-humana.md`.
- **Controles**: todos los de §3 marcados H0.
- **Aceptación**: PR de prueba con un `fmt.Println` en `internal/` falla en lint; PR limpia pasa en < 3 min.

#### H1 · Kernel CLI: multicall, flags globales, exit codes, sobre de salida
- **Objetivo**: el patrón que todos los applets repetirán, escrito una vez.
- **Entrega**: `kitlegal echo hola --json` devuelve el sobre `{ok, fuente, url, fecha_consulta, hash, data}`; `ln -s kitlegal echo && ./echo hola` funciona; `--describe` emite JSON Schema del applet.
- **Alcance**: `internal/cli` (Kong, flags `--json --timeout --offline --dry-run --describe --no-graph --asunto --verbose`, `errors.go` con errores tipados → exit codes, `slog` a stderr), `internal/app` (registro de applets, dispatch por `os.Args[0]` o primer argumento), `internal/core/schema` (tipos del sobre, `hash` sha256 del `data` canónico), `internal/render` (json y tabla mínima). El applet `echo` es de ejemplo y vive solo en tests (`internal/app/testdata`).
- **Controles**: unit (mapeo error→exit, sobre, dispatch, `--describe`); e2e testscript (`--help`, symlink, exit 2 con args malos, salida JSON parseable); test de arquitectura; `depguard` activo.
- **Aceptación**: cobertura `internal/cli` ≥ 90 %; el e2e demuestra que stdout solo contiene JSON cuando `--json`.

#### H2 · `internal/httpx`: cliente HTTP responsable + grabación de fixtures
- **Objetivo**: toda la red pasa por aquí. Sin esto no se escribe ningún adaptador.
- **Entrega**: `httpx.New(opts...)` con UA `kitlegal/x.y (+https://ventanillalegal.es/bot)`, timeout por contexto, retries con backoff exponencial y jitter, rate limit por host (`x/time/rate`), `robots.txt` cacheado, y modo `KITLEGAL_RECORD=1` que graba petición/respuesta en `testdata/<fuente>/<nombre>.json`; en tests, `httpx.Replay(dir)`.
- **Alcance**: cadena de decoradores sobre `http.RoundTripper`; errores tipados (`ErrSourceDown` en 5xx/timeout, `ErrRateLimited` en 429). Sin paralelismo por defecto.
- **Controles**: unit con `httptest` (retries, 429, robots deny, UA presente, `--dry-run`); `-race`; `noctx` y `bodyclose` en lint.
- **Aceptación**: un adaptador de prueba no puede hacer una petición sin contexto ni sin UA.

#### H3 · `internal/cache`: SQLite con TTL y `--offline`
- **Objetivo**: las normas cambian poco; no volver a pedir lo que ya tenemos.
- **Entrega**: `cache.Get/Put(key, ttl)` en `~/.cache/kitlegal/cache.db` (ruta configurable con `KITLEGAL_CACHE_DIR`); con `--offline` solo lee, y devuelve exit 4 si falta.
- **Alcance**: `modernc.org/sqlite`, migraciones embebidas (`embed` + tabla `schema_version`), WAL, `PRAGMA` seguros. Interfaz `Cache` definida en `core`.
- **Controles**: integración (`//go:build integration`) con `t.TempDir()`; unit de expiración de TTL con reloj inyectado; `sqlclosecheck`.
- **Aceptación**: segunda llamada idéntica no toca la red (se verifica con `Replay` en modo estricto que falla ante cualquier petición).

#### H4 · Applet `boe`: puerto de `boe.py` — PRIMER HITO
- **Objetivo**: `kitlegal boe articulo BOE-A-2015-10565 a21` con el mismo comportamiento que `boe.py`.
- **Entrega**: verbos `buscar`, `indice`, `articulo`, `articulos`, `metadatos`, `analisis` contra la API de Legislación Consolidada, con caché y sobre de salida. `fuente: "boe.legislacion-consolidada"`.
- **Alcance**: `internal/source/boe` implementando `Source`; cliente, tipos, mapeo a `data`; borrador de `schemas/norma.json` y `bloque.json`. Antes de portar, leer `boe.py` línea a línea y anotar comportamientos no obvios en `internal/source/boe/doc.go`.
- **Controles**: fixtures grabados para cada verbo; golden files de la salida JSON; e2e testscript de los seis verbos en modo replay; `scripts/verify-sources.sh` con el caso `boe articulo` (nightly); fuzz ligero sobre el parser de ids de bloque (`a21`, `da3`, `dt1`).
- **Aceptación**: diff entre la salida de `boe.py articulo` y `kitlegal boe articulo` (campo `data`) vacío para 5 artículos de 3 leyes; tiempo < 200 ms en caché.

#### H5 · Skill `boe-fiscal` migrada + `data/normas.yaml` + `skills-sync`
- **Objetivo**: usarlo desde ya en Claude Code sin cambiar cómo razona la skill.
- **Entrega**: `skills/boe-fiscal/SKILL.md` idéntico salvo `scripts/boe.py` → `scripts/boe` (symlink al binario); `references/normas_fiscales.md` generado desde `data/normas.yaml` (`vertical: fiscal`); `make install` copia el binario a `$GOBIN` y enlaza `skills/*` en `~/.claude/skills/`.
- **Alcance**: `data/normas.yaml` (schema propio en `schemas/normas.yaml.json`), `scripts/skills-sync.sh`, cabecera "generado, no editar", `evals/boe-fiscal/` con 10 preguntas.
- **Controles**: CI comprueba que `references/` regenerado no difiere del commiteado (drift); validación del YAML contra su schema; eval de la skill en job manual/semanal.
- **Aceptación**: en una sesión de Claude Code, la skill responde una consulta fiscal citando `BOE-A-…` y artículo, con el binario Go y sin Python instalado.

#### H6 · Release v0.1.0
- **Objetivo**: instalación reproducible y firmada; a partir de aquí cada hito puede publicar.
- **Entrega**: `.goreleaser.yaml` (darwin/linux/windows, arm64/amd64, `CGO_ENABLED=0`, `-trimpath`, ldflags de versión), checksums, SBOM, firma cosign keyless, changelog automático, tag `v0.1.0`; script `install.sh` y tap Homebrew opcional.
- **Controles**: `goreleaser check` en CI; job de release solo en tag; smoke post-release (descarga el binario publicado y ejecuta `boe articulo` en `--offline` con caché vacía → exit 4 esperado, y `version`).
- **Aceptación**: `brew install` o `curl | sh` en una máquina limpia y `kitlegal boe articulo …` funciona.

### Fase 1 — Núcleo legal (Siguiente)

Al terminar: `legal-core`, `cita-verificada` y `boe-legislacion` existen; una cita en lenguaje natural se resuelve a texto vigente con URL ELI.

#### H7 · `internal/core/ids`
- Parse y validación de `BOE-A-…`, `BOE-B-…`, ELI, ECLI, CELEX, código INE, DIR3, NIF (solo validación de formato y dígito de control; nunca se persiste NIF de persona física).
- Controles: fuzz por cada tipo, table-driven exhaustivos, benchmark de referencia. Cobertura ≥ 95 %.

#### H8 · `internal/core/cita` + applet `cita` + skill `cita-verificada`
- Parser de `"art. 21.1 Ley 39/2015"`, `"artículo 118 LCSP"`, `"DA 3ª LGT"` → `{norma, bloque}`; resolución a ELI vía `data/normas.yaml` y a texto vía `boe articulo`. Verbos `resolver`, `validar`. Exit 3 si el bloque no existe en el índice.
- Controles: fuzz del parser; corpus de 100 citas reales en `testdata/cita/corpus.txt` con resultado esperado; e2e.

#### H9 · `internal/core/plazos` + applet `plazos` + skill `legal-core`
- Días hábiles/naturales, arts. 30-31 LPAC, agosto inhábil judicial, festivos desde `data/festivos/AAAA.json` (con `source`). Tipos: alzada, reposición, contencioso, LTAIBG, silencio.
- Controles: property tests (añadir N días hábiles nunca cae en festivo; inverso consistente), tabla de casos con fecha esperada verificada a mano. `legal-core` con `references/` generadas (`leyes_vertebrales.md`, `recursos_y_plazos.md`, `reglas_de_cita.md`).

#### H10 · `boe sumario | vigilar | eli` + skill `boe-legislacion` + verificación de ids
- Sumario diario, detección de normas actualizadas por rango de fechas, construcción de ELI. Cerrar el pendiente: verificar con `boe buscar` los ids de la tabla de leyes vertebrales y fijarlos en `data/normas.yaml`.
- Controles: fixtures, golden, verify-sources para sumario.

#### H11 · `--describe` completo, `schemas/` generado y `render` markdown
- JSON Schema de entrada/salida por applet desde los tipos Go; `make schema-check` falla si `schemas/` no coincide; tabla de comandos de cada `SKILL.md` generada por `skills-sync` desde `--describe`.
- Controles: test de contrato en todos los applets existentes; drift check en CI. Aquí hay ya 4 applets (`boe`, `cita`, `plazos`, `echo`): momento de revisar si `internal/cli` pide refactor, no extracción.

### Fase 2 — Grafo mínimo

Al terminar: el uso alimenta `world.db`; un asunto en `.kitlegal/` recuerda qué se consultó y `graph check` avisa de versiones obsoletas.

#### H12 · `internal/graph` + emisión desde `boe` + `graph show|stats`
- Tablas `nodes/edges/texts` (+ FTS5), `Apply([]GraphOp)` transaccional, `--no-graph` (Null Object). `boe articulo` emite `Norma`, `Bloque`, `BloqueVersion`.
- Controles: integración con DB temporal; idempotencia (aplicar dos veces = un nodo); test de que ningún nodo/arista entra sin `source`.

#### H13 · `boe analisis` → aristas ELI + `graph history`
- `eli:amends/repeals/amended_by/repealed_by`; nodos stub; diff entre `BloqueVersion`.
- Controles: fixtures de una ley con modificaciones; golden del diff.

#### H14 · Grafo del asunto + `graph check` inicial
- `.kitlegal/case.db` y `config.yaml`, nodo `Consulta`, `lb:en_asunto`, `--asunto`; `ATTACH` de `world.db`. Reglas: `version-obsoleta`, `fuente-caducada`, `cita-inexistente`. Exit 0/1/2.
- Controles: e2e de un flujo completo (consultar, simular nueva versión en fixture, `check` marca warning); test de que `case.db` nunca contiene texto ni props del mundo, solo ids; `graph export` de asunto pide confirmación.

### Fase 3 — Pack `fiscalizador` sobre Leganés

Al terminar: PLACSP, BDNS y BORME alimentan el grafo y `graph anomalies` devuelve señales verificables con sus URLs.

#### H15 · Spike de fuentes (timebox: 1 día)
- Verificar endpoints BDNS (swagger), nombres de feeds Atom de PLACSP, API de BORME. Resultado: `docs/SOURCES.md` con filas completas y fixtures grabados. Sin código de producto.

#### H16 · `placsp sync | licitaciones | organo | adjudicatario` + `internal/store`
- Sync incremental de Atom + CODICE XML a `store`; emisión de `Organo`, `Licitacion`, `Contrato`, `Entidad` y aristas.
- Controles: fuzz del parser CODICE; integración de sync incremental (segunda ejecución no duplica); e2e con replay.

#### H17 · `bdns convocatorias | concesiones | beneficiario`
- REST; emisión de `Convocatoria`, `Concesion`, `Entidad`. Mismos controles que H16.

#### H18 · `borme sumario | empresa | administrador`
- Vía API BOE; `Persona` sin NIF, `lb:administra` con `valid_from/to`.
- Controles: test explícito de que `Persona` rechaza cualquier prop que parezca NIF/DNI (regex) en `graph.Apply`.

#### H19 · `graph neighbors | path` + reglas de anomalías de contratación
- `data/anomalias/*.yaml` (SQL + umbrales), `graph anomalies --organo`: `fraccionamiento`, `licitador-unico`, `concentracion`, `plazo-publicacion`. Salida con `nodos`, `aristas`, `verificar_en`.
- Controles: cada regla con un grafo sintético en fixture que la dispara y otro que no; validación del YAML contra schema; comprobación de que el SQL de las reglas es de solo lectura (rechazar todo lo que no empiece por `SELECT`/`WITH`).

#### H20 · Cruces entre fuentes + skills + pack
- `administrador-comun`, `beneficiario-adjudicatario`, `empresa-recien-creada`. Skills `contratacion-publica`, `subvenciones`, `entidades-y-registros`; `packs/fiscalizador/pack.yaml`; `data/entidades/leganes.yaml`.
- Controles: evals de las tres skills; validación de casos conocidos de Leganés (documentar como fixtures anonimizados de aristas, no de personas).

#### H21 · `vigilar run` sobre delta + `anomalies mark`
- Orquesta `boe vigilar`, `placsp sync`, `bdns`, `borme sumario` sobre `config.yaml`; anomalías solo sobre aristas nuevas; `mark --confirmada|--descartada` en el asunto.
- Controles: e2e de dos ejecuciones consecutivas con fixtures distintos; solo emite la diferencia. Release **v0.5.0**.

### Fase 4 — Distribución

#### H22 · `kitlegal mcp serve --stdio`
- Tools generadas desde `--describe`; mismo sobre de salida. `modelcontextprotocol/go-sdk`.
- Controles: test de conformidad (listar tools == applets registrados); e2e con cliente MCP de prueba; el servidor no expone `graph export` del asunto sin flag explícita.

#### H23 · `kitlegal skills install|list|doctor` + `plugin/`
- Instala packs en el directorio del cliente; `doctor` comprueba versión mínima, hash de skills y lanza `verify-sources` para las fuentes del pack. `plugin/.claude-plugin/plugin.json` y `marketplace.json` generados desde `packs/`.
- Controles: e2e en `t.TempDir()` simulando `~/.claude`; drift check de `plugin/`.

#### H24 · `pkg/legalkit` (solo ahora: hay > 3 applets)
- Facade mínima: `Resolve(cita)`, `Articulo(id, bloque)`, tipos de dominio. Compromiso de compatibilidad documentado. Release **v1.0.0**.
- Controles: `apidiff` en CI para detectar roturas de API pública; ejemplo compilable en `pkg/legalkit/example_test.go`.

### Fase 5 — Más fuentes (un hito por fuente, independientes, en el orden de `refs/mapa` §5)

Cada uno sigue el molde de H16: spike de TOS → fixtures → adaptador → emisión al grafo → skill → verify-sources. Orden sugerido: `boletin` (motor YAML, BOCM primero) → `edictos` (TEU) → `catastro` → `ine` + `datosgob` → `eurlex` → `congreso` → `ecli` → `tc` → `presupuesto` → `sede` → `dgt` y `teac` **solo tras revisar TOS** (pendiente de `refs/00-README.md`).

### Fase 6 — Acción (frontera humana)

#### H25 · `escrito generar` + `Afirmacion` + `fundamenta` + `check` bloqueante
- Plantillas en `internal/docgen` (markdown y DOCX), cada afirmación con arista `fundamenta` o marca `[SIN FUNDAMENTO]` y exit 3; `check` completo (`cita-huerfana`, `norma-derogada`, `plazo-vencido`…). Termina siempre en fichero + URL de la sede + exit 6.
- Controles: test de que no existe ninguna llamada HTTP con método distinto de GET/HEAD en todo el módulo (test de arquitectura sobre `httpx`); golden de escritos; e2e de `check` bloqueando sin `--force`.

#### H26 · `expediente listar|add|vencimientos` + exportaciones
- Registro en el asunto; `graph export --format md|jsonld|dot|csv` con anonimización por defecto.

---

## 5. Lo que deliberadamente NO se hace (hasta que un hito lo pida)

- Framework de DI, ORM, generador de CLI: Kong + composición manual bastan.
- Extraer `internal/cli` a librería (regla de los tres applets se cumple en H11; aun así, solo refactor interno).
- MCP remoto con auth, cuotas o caché compartida.
- Triple store, SPARQL local, embeddings, sincronización entre máquinas.
- Paralelismo en descargas: primero correcto y respetuoso, luego rápido.
- Ingesta de CENDOJ, scraping de PETETE/DYCTEA/HJ sin TOS revisados.
- Cobertura como objetivo: el umbral existe para no retroceder, no para perseguirlo.

---

## 6. Ritual por hito (para que "lean" no degenere en "sin control")

1. Abrir rama `hNN-nombre-corto`; escribir primero el test e2e (`testscript`) que describe la entrega.
2. Implementar de dentro afuera: `core` → adaptador → applet → skill.
3. `make ci` local; PR con plantilla (objetivo, alcance, controles añadidos, decisiones, pendientes).
4. Revisión: `/code-review` y `/security-review` sobre la PR; si toca fuentes, `docs/SOURCES.md`.
5. Squash-merge; si el hito cierra una fase, tag y release.
6. Actualizar este roadmap solo si cambia el orden o el alcance; el detalle vive en las PR y los ADR.
