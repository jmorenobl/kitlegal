# kitlegal: roadmap de hitos

Roadmap para implementar lo descrito en `refs/`. Enfoque lean: hitos de 0,5 a 3 días, cada uno termina en algo que se puede usar (o que protege lo que ya se usa). No hay fechas; hay orden y dependencias. Lo que no está aquí no se construye hasta que un hito lo necesite (YAGNI).

Las decisiones cerradas de `CLAUDE.md` y `refs/00-README.md` no se repiten: este documento las asume.

**Qué se construye y en qué orden.** El producto son las **skills**; el binario Go es la capa de herramientas deterministas que usan (constitución, principio VIII). La prioridad, tras la fundación, es **hacer cosas en tu municipio**: contratación, subvenciones, boletines, ordenanzas, plazos, escritos y su seguimiento. Todo es **genérico para cualquier municipio de España** y se valida primero en el territorio de referencia, la Comunidad de Madrid con Leganés (principio IX). Los demás territorios (BOCYL, BOP de comunidades multiprovinciales, régimen foral…) llegan después como datos y adaptadores, sin tocar las skills.

---

## 0. Cómo leer este roadmap

- **Fase** = épica. Al final de cada fase hay un usuario real que gana algo: tú en Claude Code con tu municipio, luego un pack instalable, luego terceros.
- **Hito** = PR única, rama corta desde `main` (trunk-based), squash-merge con CI verde. Un hito no se cierra sin cumplir la *Definition of Done* (§1).
- **Ahora / Siguiente / Después**: solo la fase en curso se planifica al detalle. Las fases lejanas son intencionadamente esquemáticas y se refinan al llegar.
- Cada hito indica: objetivo, entrega usable, alcance, controles (tests y calidad) y criterio de aceptación. A partir de H5, el objetivo se formula como lo que una skill pasa a poder resolver.
- La numeración de hitos se rehízo el 2026-09-11 (ADR 0005 y 0006); la correspondencia con la anterior está en `docs/ADR/0005-skills-primero.md`.

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
10. Si entrega o cambia una skill (desde H5): evals en `evals/<skill>/` (activación y respuesta con citas esperadas) escritas antes que el código y en verde; `SKILL.md` < 300 líneas con frontmatter válido; `references/` y tabla de comandos regeneradas sin drift.
11. Si tiene dimensión territorial: e2e con un municipio cubierto y otro no cubierto, con la cobertura declarada en `data`; ningún caso especial para un municipio en código ni en `data/` (los municipios concretos solo aparecen en fixtures y evals).

---

## 2. Arquitectura objetivo (resumen operativo)

Dos capas: las **skills** (el producto: protocolo de razonamiento, conocimiento y comandos) y el **binario** (herramientas deterministas que las skills invocan). `data/` es la única fuente de verdad de ambas.

```
  skills/<skill>/  SKILL.md (protocolo) · references/ (generadas) · scripts/<applet> → kitlegal   ← EL PRODUCTO
  packs/<vertical>/pack.yaml                                    (agrupan skills por vertical)
  data/  normas · territorio · boletines · festivos · anomalias  (fuente de verdad; skills-sync genera references/)
                │  el agente ejecuta scripts/<applet> <verbo> … y recibe el sobre JSON
                ▼
                ┌────────────────────────── cmd/kitlegal (main, multicall)
                │
        internal/app ── composition root, registro de applets (Command + Registry)
                │
        internal/cli ── Kong, flags globales, mapeo error→exit code, sobre de salida
                │
  ┌─────────────┼─────────────────────────────────────────────┐
  │ internal/core (dominio puro, sin I/O)                     │
  │   ids · territorio · cita · plazos · competencia · schema │
  │   define los PUERTOS: Source, Fetcher, Cache, GraphStore, │
  │   Renderer                                                │
  └─────────────┬─────────────────────────────────────────────┘
                │ implementan
   ADAPTADORES: internal/source/<fuente>  (Adapter, uno por fuente)
                internal/httpx            (Decorator chain: UA → robots → ratelimit → retry → record/replay)
                internal/cache, store, graph (Repository sobre SQLite)
                internal/render           (Strategy: json | table | markdown)
```

**Qué va a Go y qué no.** Va al binario lo que exige determinismo o verificabilidad: acceso a fuentes, parseo, caché, fechas y plazos, identificadores, hashes y grafo. El razonamiento, la interpretación, la elección de fuentes y la redacción de la respuesta viven en `SKILL.md` y `references/`. Si una capacidad puede resolverla el modelo con fiabilidad a partir de una referencia generada, no se convierte en applet.

**Territorio.** Los datos de cada municipio (INE, provincia, comunidad, DIR3) salen de registros nacionales; lo que varía por territorio se configura por comunidad o por boletín en `data/territorio/` y `data/boletines/`. El municipio de la persona usuaria vive en `.kitlegal/config.yaml`. Fuera del territorio cubierto, cada applet declara la cobertura dentro de `data`.

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
| Configuración por datos | `data/territorio/`, `data/boletines/` | Un territorio o un boletín nuevo es un YAML, no código ni una skill nueva |
| Facade | `pkg/legalkit` (H30) | API pública pequeña sobre los internos |

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
| Fuzzing | `go test -fuzz` sobre parsers: `ids`, `cita`, parseo XML/Atom; corpus mínimo versionado en `testdata/fuzz/` | H7, H8, H12 |
| Property tests | `testing/quick` o `pgregory.net/rapid` en `plazos` (idempotencia, monotonía, ida y vuelta) | H9 |
| Tests de integración | `//go:build integration`, SQLite en `t.TempDir()`, `httptest.Server` como fuente | H3 |
| Tests e2e | `github.com/rogpeppe/go-internal/testscript` contra el binario compilado: `kitlegal …`, `boe …` vía symlink, exit codes, `--json`, `--offline` | H1 |
| Matriz territorial | e2e por applet con dimensión territorial: municipio cubierto (Leganés), no cubierto (Tordesillas) y, cuando aplique, foral | H7 |
| Tests de contrato | Cada salida validada contra `schemas/*.json` (`santhosh-tekuri/jsonschema`); `--describe` genera el schema y CI falla si `schemas/` está desactualizado | H10 (borrador en H4) |
| Smoke contra red | `scripts/verify-sources.sh` en CI nightly; abre issue si una fuente cambia | H4 |
| Comprobación mecánica de skills | En `make ci`: frontmatter válido, `SKILL.md` < 300 líneas, `references/` y tabla de comandos idénticas a lo generado | H5 |
| Evals de skills | `evals/<skill>/*.yaml` (pregunta, si debe activar la skill, comandos esperados, citas esperadas por identificador); job semanal con modelo barato | H5 |
| Vulnerabilidades | `govulncheck` en cada PR | H0 |
| SAST | `gosec` (en golangci) + CodeQL semanal | H0 |
| Secretos | `gitleaks` en pre-commit y CI | H0 |
| Cadena de suministro | `go mod verify`, `go.sum` obligatorio, Dependabot semanal, SBOM (syft) + firma (cosign) + provenance en release | H0 / H6 |
| Dependencias mínimas | `go mod tidy -diff` en CI; una dependencia nueva requiere justificación en la PR | H0 |
| Observabilidad local | `log/slog` a stderr, `--verbose`, `KITLEGAL_LOG=debug`; `--dry-run` imprime la petición sin ejecutarla | H1 |
| Pre-commit | `lefthook`: `gofumpt`, `golangci-lint run --fast`, `gitleaks`, `go mod tidy -diff` | H0 |
| Commits / versiones | Conventional Commits, SemVer, `CHANGELOG.md` generado por goreleaser | H0 / H6 |
| Docs de decisiones | `docs/ADR/NNNN-*.md` (formato MADR corto) | H0 |

Dependencias fijadas (fuera de esta lista, justificar): `alecthomas/kong`, `modernc.org/sqlite`, `stretchr/testify`, `rogpeppe/go-internal` (testscript), `golang.org/x/time/rate`, `temoto/robotstxt`, `gopkg.in/yaml.v3`, `invopop/jsonschema` (generación), `santhosh-tekuri/jsonschema` (validación en tests). En la fase de distribución (H28): `modelcontextprotocol/go-sdk`.

---

## 4. Hitos

### Fase 0 — Fundación: la primera skill con su herramienta (Ahora)

Al terminar: la skill `boe-fiscal` responde en Claude Code con el binario Go y sin Python, medida con evals; `kitlegal boe articulo BOE-A-2015-10565 a21` funciona con caché y hay un release v0.1.0 firmado.

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

#### H5 · Skill `boe-fiscal` migrada + andamiaje de skills (`skills-sync`, evals)
- **Objetivo**: usar `boe-fiscal` desde ya en Claude Code sin cambiar cómo razona, y dejar el andamiaje con el que se miden todas las skills siguientes.
- **Entrega**: `skills/boe-fiscal/SKILL.md` idéntico salvo `scripts/boe.py` → `scripts/boe` (symlink al binario); `references/normas_fiscales.md` generado desde `data/normas.yaml` (`vertical: fiscal`); `make install` copia el binario a `$GOBIN` y enlaza `skills/*` en `~/.claude/skills/`; formato común de evals y `evals/boe-fiscal/` con 10 preguntas.
- **Alcance**: `data/normas.yaml` (schema propio en `schemas/normas.yaml.json`), `scripts/skills-sync.sh`, cabecera "generado, no editar", formato de eval (pregunta, si debe activar la skill, comandos esperados, citas esperadas por identificador), comprobación mecánica de skills dentro de `make ci` (frontmatter, líneas, drift de `references/`).
- **Controles**: CI comprueba que `references/` regenerado no difiere del commiteado (drift); validación del YAML contra su schema; eval de la skill en job manual/semanal; las citas esperadas se comparan por identificador (`BOE-A-…` + bloque).
- **Aceptación**: en una sesión de Claude Code, la skill responde una consulta fiscal citando `BOE-A-…` y artículo, con el binario Go y sin Python instalado; las 10 evals pasan.

#### H6 · Release v0.1.0
- **Objetivo**: instalación reproducible y firmada; a partir de aquí cada hito puede publicar.
- **Entrega**: `.goreleaser.yaml` (darwin/linux/windows, arm64/amd64, `CGO_ENABLED=0`, `-trimpath`, ldflags de versión), checksums, SBOM, firma cosign keyless, changelog automático, tag `v0.1.0`; script `install.sh` y tap Homebrew opcional.
- **Controles**: `goreleaser check` en CI; job de release solo en tag; smoke post-release (descarga el binario publicado y ejecuta `boe articulo` en `--offline` con caché vacía → exit 4 esperado, y `version`).
- **Aceptación**: `brew install` o `curl | sh` en una máquina limpia y `kitlegal boe articulo …` funciona.

### Fase 1 — Núcleo común: territorio, citas y plazos (Siguiente)

Al terminar: `legal-core` y `cita-verificada` funcionan para cualquier municipio de España. Una cita en lenguaje natural se resuelve a texto vigente con URL ELI, y el plazo para recurrir un acto de tu ayuntamiento se calcula con los festivos nacionales, autonómicos y locales cuando el territorio está cubierto (Comunidad de Madrid primero), avisando explícitamente cuando no lo está.

#### H7 · `territorio` + skill `legal-core` v0
- **Objetivo**: que toda skill sepa, antes de razonar, en qué territorio está la pregunta y qué normas vertebrales aplican. Es la pieza que hace genéricas a las demás.
- **Entrega**: `kitlegal territorio resolver <nombre|código INE>` devuelve municipio, código INE, provincia, comunidad autónoma, DIR3 del ayuntamiento, régimen (común o foral), boletines aplicables y `cobertura` (qué del territorio está configurado y qué no). Skill `legal-core` v0: protocolo que empieza por identificar el territorio, `references/leyes_vertebrales.md` y `jerarquia_normativa.md`.
- **Alcance**: registro de municipios desde la relación oficial del INE y DIR3 desde el inventario público (formato, licencia y forma de actualización a verificar en tarea `[datos]`, con fila en `docs/SOURCES.md`); `data/territorio/` con configuración por comunidad, solo la Comunidad de Madrid rellena (su boletín, el BOCM, hace de provincial por ser uniprovincial); el resto de comunidades responde con los datos nacionales y cobertura parcial. Nombre ambiguo: lista de candidatos y exit 2. Ids que entran aquí: código INE y DIR3 (`internal/core/ids`, formato y dígito de control). Verificar con `boe buscar` los ids de la tabla de leyes vertebrales y fijarlos en `data/normas.yaml`.
- **Controles**: matriz territorial en e2e: Leganés (cubierto), Tordesillas (datos nacionales completos, boletines autonómico y provincial declarados no cubiertos, nada inventado), un municipio de Navarra o del País Vasco (régimen foral marcado), un nombre ambiguo; fuzz del parser de INE y DIR3; evals de `legal-core` que exigen identificar el territorio.
- **Aceptación**: en Claude Code, «¿qué comunidad, provincia y boletines corresponden a mi ayuntamiento?» se responde para cualquier municipio, con la cobertura explícita cuando no es de la Comunidad de Madrid.

#### H8 · `internal/core/cita` + applet `cita` + skill `cita-verificada`
- Parser de `"art. 21.1 Ley 39/2015"`, `"artículo 118 LCSP"`, `"DA 3ª LGT"` → `{norma, bloque}`; resolución a ELI vía `data/normas.yaml` y a texto vía `boe articulo`. Verbos `resolver`, `validar`. Exit 3 si el bloque no existe en el índice. Las normas autonómicas se resuelven igual (el BOE consolida legislación de las comunidades); `data/normas.yaml` empieza con las estatales vertebrales y las autonómicas que citen las skills del territorio de validación. Ids que entran aquí: `BOE-A-…` y ELI.
- Controles: fuzz del parser y de los ids; corpus de 100 citas reales en `testdata/cita/corpus.txt` con resultado esperado (incluye citas autonómicas); e2e; evals de `cita-verificada`.

#### H9 · `internal/core/plazos` + applet `plazos` + skill `legal-core` v1
- Días hábiles/naturales, arts. 30-31 LPAC (incluido el art. 30.6: inhábil en el municipio del interesado o en la sede del órgano), agosto inhábil judicial. Tipos: alzada, reposición, contencioso, LTAIBG, silencio. Festivos desde `data/festivos/` (con `source`): nacionales y autonómicos de todas las comunidades (se publican cada año en el BOE) y locales de los municipios del territorio cubierto (la Comunidad de Madrid publica los de todos sus municipios; formato a verificar en tarea `[datos]`). Fuera de cobertura, el resultado avisa en `data` de que faltan los festivos locales y nunca los da por inexistentes.
- `legal-core` v1: `references/recursos_y_plazos.md` y `reglas_de_cita.md` generadas; la skill usa `territorio` y `plazos`.
- Controles: property tests (añadir N días hábiles nunca cae en festivo; inverso consistente); tabla de casos con fecha esperada verificada a mano, incluido un plazo en Leganés que cruza un festivo local y el mismo plazo en Tordesillas con aviso de cobertura; evals.

#### H10 · `--describe` completo, `schemas/` generado y `render` markdown
- JSON Schema de entrada/salida por applet desde los tipos Go; `make schema-check` falla si `schemas/` no coincide; tabla de comandos de cada `SKILL.md` generada por `skills-sync` desde `--describe`.
- Controles: test de contrato en todos los applets existentes; drift check en CI. Aquí hay ya 5 applets (`boe`, `territorio`, `cita`, `plazos`, `echo`): momento de revisar si `internal/cli` pide refactor, no extracción.

### Fase 2 — Mi municipio: consultar y actuar (Después)

Al terminar: para cualquier municipio, las skills responden qué contrata y qué subvenciona su ayuntamiento, qué se publica sobre él en boletines y edictos y qué dicen sus ordenanzas, y preparan una solicitud de acceso a información o un recurso de reposición listos para firmar. Validado en Leganés. Fuera de la Comunidad de Madrid, contratación y subvenciones funcionan igual (son fuentes nacionales) y boletines y ordenanzas declaran su cobertura.

#### H11 · Spike de fuentes municipales (timebox: 1 día)
- Verificar los feeds Atom de PLACSP (qué órganos publican directamente, qué llega agregado desde plataformas autonómicas y si los contratos menores agregados están incluidos), los endpoints de BDNS (swagger), el BOCM (sumario, buscador, formatos, TOS) y el Tablón Edictal Único. Resultado: `docs/SOURCES.md` con filas completas y fixtures grabados. Sin código de producto.

#### H12 · `placsp sync | licitaciones | organo | adjudicatario` + `internal/store` + skill `contratacion-publica` v0
- Sync incremental de Atom + CODICE XML a `store`; el órgano se identifica por su DIR3, obtenido con `territorio` a partir del municipio. Skill `contratacion-publica` v0: qué ha licitado y adjudicado el ayuntamiento, a quién y por qué importe, incluidos los contratos menores. Si el órgano publica en una plataforma autonómica cuyos datos no llegan a PLACSP, la salida lo declara en la cobertura. Sin grafo ni anomalías (fase 3).
- Controles: fuzz del parser CODICE; integración de sync incremental (segunda ejecución no duplica); e2e con replay para Leganés y para un municipio de otra comunidad; evals.

#### H13 · `bdns convocatorias | concesiones | beneficiario` + skill `subvenciones` v0
- REST; órgano por DIR3 vía `territorio`. La skill responde qué subvenciones convoca y concede el ayuntamiento y a quién. Mismos controles que H12.

#### H14 · `boletin` (motor genérico, BOCM) + `edictos` (TEU) + skill `bop-y-edictos`
- `boletin sumario | buscar` con un motor genérico configurado por YAML (`data/boletines/bocm.yaml` es la primera y única configuración) y `edictos buscar` sobre el Tablón Edictal Único, que es nacional. El filtro por municipio usa `territorio`. Añadir otro boletín es un YAML más, y un adaptador solo si el motor no basta (fase 6).
- Controles: fixtures de BOCM y TEU; e2e con Leganés (BOCM + TEU) y Tordesillas (TEU, más la cobertura que declara no configurados los boletines autonómico y provincial); `verify-sources`; evals.

#### H15 · Ordenanzas vía boletín + skill `ordenanzas-locales`
- Las ordenanzas se publican íntegras en el boletín provincial o, en comunidades uniprovinciales, en el autonómico (art. 70.2 LRBRL; ordenanzas fiscales, art. 17.4 TRLRHL). `ordenanzas listar | ver --municipio` busca con `boletin` los anuncios de aprobación definitiva y de modificación del territorio, sin crawler por ayuntamiento. El texto sale del formato que publique el boletín; si solo hay PDF, la dependencia de extracción se justifica en el plan. La skill presenta el texto aprobado y las modificaciones que encuentre, citando cada anuncio, sin afirmar que es un texto consolidado.
- Controles: fixtures; e2e con la matriz territorial; evals («¿qué dice la ordenanza de terrazas de mi municipio?»). El crawler de sedes municipales queda para la fase 7.

#### H16 · `escrito generar` (mínimo) + skill `redaccion-escritos`
- Tipos iniciales: solicitud de acceso a información (LTAIBG art. 17) y recurso de reposición contra un acto municipal (LPAC arts. 123-124). Salida en markdown desde plantillas en `internal/docgen`; órgano destinatario desde `territorio`; fundamentación resuelta por `cita` (una cita que no se resuelve sale como `[SIN FUNDAMENTO]` y exit 3); plazo por `plazos`. Dónde presentarlo: el Registro Electrónico General de la AGE, que admite escritos dirigidos a cualquier administración (art. 16.4 LPAC), y la sede del ayuntamiento cuando conste. Termina siempre en fichero + URL + exit 6.
- Controles: test de arquitectura de que no existe ninguna llamada HTTP con método distinto de GET/HEAD en todo el módulo; golden de escritos; evals. Fuera de alcance hasta H26: DOCX, `Afirmacion`/`fundamenta` en el grafo y `check` bloqueante.

### Fase 3 — Mi municipio: seguir y fiscalizar

Al terminar: el uso alimenta el grafo, tus expedientes avisan de silencios y vencimientos, y PLACSP, BDNS, BORME y presupuestos se cruzan para que `graph anomalies` devuelva señales verificables sobre el órgano de cualquier municipio, validadas primero en Leganés.

#### H17 · `internal/graph` + emisión desde los applets existentes + `graph show|stats`
- Tablas `nodes/edges/texts` (+ FTS5), `Apply([]GraphOp)` transaccional, `--no-graph` (Null Object). `Emit` en `boe` (`Norma`, `Bloque`, `BloqueVersion`), `territorio` (`Municipio`, `Organo`, `lb:pertenece_a`), `placsp` y `bdns`: los applets anteriores al grafo lo incorporan aquí.
- Controles: integración con DB temporal; idempotencia (aplicar dos veces = un nodo); test de que ningún nodo/arista entra sin `source`.

#### H18 · Grafo del asunto + `expediente` + skill `seguimiento-expedientes`
- `.kitlegal/case.db` y `config.yaml` (con el municipio de la persona usuaria), nodo `Consulta`, `lb:en_asunto`, `--asunto`; `ATTACH` de `world.db`. `expediente add | listar | vencimientos`: solicitudes y recursos presentados, con los silencios calculados por `plazos` según la norma aplicable (la estatal por defecto; la autonómica cuando el territorio la configure). `graph check` inicial: `version-obsoleta`, `fuente-caducada`, `cita-inexistente`, `plazo-vencido`. Exit 0/1/2.
- Controles: e2e de un flujo completo (registrar una solicitud, avanzar el reloj, `vencimientos` avisa; simular nueva versión en fixture y `check` marca warning); test de que `case.db` nunca contiene texto ni props del mundo, solo ids; `graph export` de asunto pide confirmación.

#### H19 · `borme sumario | empresa | administrador` + skill `entidades-y-registros`
- Vía API BOE; `Persona` sin NIF, `lb:administra` con `valid_from/to`.
- Controles: test explícito de que `Persona` rechaza cualquier prop que parezca NIF/DNI (regex) en `graph.Apply`; evals.

#### H20 · `presupuesto descargar | comparar` + skill `presupuestos-y-cuentas`
- Presupuestos y liquidaciones de entidades locales desde los datos abiertos del Ministerio de Hacienda, por código INE: fuente nacional, sirve para cualquier municipio (endpoint y formato a verificar en tarea `[datos]`). `comparar --aprobado --liquidado` da desviaciones por capítulo y programa.
- Controles: fixtures; e2e con la matriz territorial; evals.

#### H21 · `graph neighbors | path` + reglas de anomalías de contratación
- `data/anomalias/*.yaml` (SQL + umbrales), `graph anomalies --organo`: `fraccionamiento`, `licitador-unico`, `concentracion`, `plazo-publicacion`. Las reglas con umbral legal (`fraccionamiento`, `plazo-publicacion`) lo toman de la ley; las relativas (`licitador-unico`, `concentracion`) comparan con municipios de tamaño parecido (tramos de población del padrón del INE, fuente nacional). Salida con `nodos`, `aristas`, `verificar_en`. `contratacion-publica` v1 incorpora la fiscalización.
- Controles: cada regla con un grafo sintético en fixture que la dispara y otro que no; validación del YAML contra schema; comprobación de que el SQL de las reglas es de solo lectura (rechazar todo lo que no empiece por `SELECT`/`WITH`); comprobación de que ninguna regla referencia un municipio u órgano concreto.

#### H22 · Cruces entre fuentes + pack `fiscalizador`
- `administrador-comun`, `beneficiario-adjudicatario`, `empresa-recien-creada`. `subvenciones` v1 incorpora los cruces; `packs/fiscalizador/pack.yaml` sin lista de municipios (el municipio lo pone cada persona en su `.kitlegal/config.yaml`).
- Controles: evals de las skills del pack; validación con casos conocidos de Leganés, documentados como fixtures anonimizados de aristas, no de personas, más un municipio de otra comunidad para comprobar que las reglas no dependen del de referencia.

#### H23 · `vigilar run` sobre delta + `anomalies mark`
- Orquesta `placsp sync`, `bdns`, `borme sumario`, `boletin` y `edictos` sobre el municipio y las entidades de `config.yaml`; anomalías solo sobre aristas nuevas; `mark --confirmada|--descartada` en el asunto.
- Controles: e2e de dos ejecuciones consecutivas con fixtures distintos; solo emite la diferencia. Release **v0.5.0**.

### Fase 4 — Profundidad legal

#### H24 · `boe sumario | vigilar | eli` + skill `boe-legislacion`
- Sumario diario, detección de normas actualizadas por rango de fechas, construcción de ELI; `boe-fiscal` pasa a ser una especialización de `boe-legislacion`. `vigilar run` incorpora `boe vigilar`.
- Controles: fixtures, golden, `verify-sources` para sumario; evals.

#### H25 · `boe analisis` → aristas ELI + `graph history`
- `eli:amends/repeals/amended_by/repealed_by`; nodos stub; diff entre `BloqueVersion`.
- Controles: fixtures de una ley con modificaciones; golden del diff.

#### H26 · Escritos fundamentados en el grafo: `Afirmacion` + `fundamenta` + `check` bloqueante
- Cada afirmación de un escrito con arista `fundamenta` o marca `[SIN FUNDAMENTO]` y exit 3; `check` completo (`cita-huerfana`, `norma-derogada`, `plazo-vencido`…); salida DOCX además de markdown; más tipos de escrito (alzada, reclamación ante el consejo de transparencia competente según el territorio, alegaciones, recurso especial en contratación).
- Controles: golden de escritos; e2e de `check` bloqueando sin `--force`.

#### H27 · Exportaciones del grafo
- `graph export --format md|jsonld|dot|csv` con anonimización por defecto.

### Fase 5 — Distribución

#### H28 · `kitlegal mcp serve --stdio`
- Tools generadas desde `--describe`; mismo sobre de salida. `modelcontextprotocol/go-sdk`.
- Controles: test de conformidad (listar tools == applets registrados); e2e con cliente MCP de prueba; el servidor no expone `graph export` del asunto sin flag explícita.

#### H29 · `kitlegal skills install|list|doctor` + `plugin/`
- Instala packs en el directorio del cliente; `doctor` comprueba versión mínima, hash de skills y lanza `verify-sources` para las fuentes del pack. `plugin/.claude-plugin/plugin.json` y `marketplace.json` generados desde `packs/`.
- Controles: e2e en `t.TempDir()` simulando `~/.claude`; drift check de `plugin/`.

#### H30 · `pkg/legalkit` (solo ahora: hay > 3 applets)
- Facade mínima: `Resolve(cita)`, `Articulo(id, bloque)`, tipos de dominio. Compromiso de compatibilidad documentado. Release **v1.0.0**.
- Controles: `apidiff` en CI para detectar roturas de API pública; ejemplo compilable en `pkg/legalkit/example_test.go`.

### Fase 6 — Más territorios (un hito por territorio, bajo demanda)

Molde: fila en `docs/SOURCES.md` → configuración en `data/territorio/` y `data/boletines/` → adaptador solo si el motor genérico no basta → fixtures → la fila de la matriz territorial pasa a «cubierto» → evals de las skills afectadas. Las skills no se tocan. Candidatos, sin orden fijado: el boletín autonómico y los BOP de una comunidad multiprovincial (p. ej. BOCYL y BOP de Valladolid, que cubren Tordesillas), DOGC, BOJA…; festivos locales de cada comunidad; plataformas autonómicas de contratación que no lleguen agregadas a PLACSP; consejos autonómicos de transparencia; régimen foral (normativa y haciendas forales en `boe-fiscal`, catastros forales).

### Fase 7 — Más fuentes (un hito por fuente, independientes, en el orden de `refs/mapa` §5)

Cada uno sigue el molde de H12: spike de TOS → fixtures → adaptador → emisión al grafo → skill → verify-sources. Orden sugerido: `catastro` → `ine` + `datosgob` → `transparencia` → skill `boletines-autonomicos` (normas autonómicas, sobre el motor de H14) → `eurlex` → `congreso` → `ecli` → `tc` → `sede` (crawler de sedes municipales, solo donde el boletín no baste) → `dgt` y `teac` **solo tras revisar TOS** (pendiente de `refs/00-README.md`).

---

## 5. Lo que deliberadamente NO se hace (hasta que un hito lo pida)

- Framework de DI, ORM, generador de CLI: Kong + composición manual bastan.
- Extraer `internal/cli` a librería (regla de los tres applets se cumple en H10; aun así, solo refactor interno).
- Applets sin una skill que los use, o capacidades en Go que el modelo resuelve con una referencia generada.
- Casos especiales, YAML o código para un municipio concreto; configuración municipal curada a mano.
- Territorios distintos del de validación (BOCYL, BOP de otras provincias, régimen foral…) antes de la fase 6.
- Crawler de sedes municipales mientras el boletín baste.
- MCP remoto con auth, cuotas o caché compartida.
- Triple store, SPARQL local, embeddings, sincronización entre máquinas.
- Paralelismo en descargas: primero correcto y respetuoso, luego rápido.
- Ingesta de CENDOJ, scraping de PETETE/DYCTEA/HJ sin TOS revisados.
- Cobertura como objetivo: el umbral existe para no retroceder, no para perseguirlo.

---

## 6. Ritual por hito (para que "lean" no degenere en "sin control")

1. Abrir rama `hNN-nombre-corto`; escribir primero las evals de la skill que el hito entrega o mejora (desde H5) y el test e2e (`testscript`) que describe la entrega.
2. Planificar de fuera adentro (qué debe resolver la skill → qué herramientas necesita) e implementar de dentro afuera: `core` → adaptador → applet → skill.
3. `make ci` local; PR con plantilla (objetivo, alcance, controles añadidos, decisiones, pendientes).
4. Revisión: `/code-review` y `/security-review` sobre la PR; si toca fuentes, `docs/SOURCES.md`.
5. Squash-merge; si el hito cierra una fase, tag y release.
6. Actualizar este roadmap solo si cambia el orden o el alcance; el detalle vive en las PR y los ADR.
