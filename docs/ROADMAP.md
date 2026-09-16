# kitlegal: roadmap de hitos

Roadmap para implementar lo descrito en `refs/`. Enfoque lean: hitos de 0,5 a 3 días, cada uno termina en algo que se puede usar (o que protege lo que ya se usa). No hay fechas; hay orden y dependencias. Lo que no está aquí no se construye hasta que un hito lo necesite (YAGNI).

Las decisiones cerradas de `CLAUDE.md` y `refs/00-README.md` no se repiten: este documento las asume.

**Qué se construye y en qué orden.** El producto son las **skills**; el binario Go es la capa de herramientas deterministas que usan (constitución, principio VIII). El orden lo fija el **tiempo hasta el uso** (ADR 0013): primero consultar y citar normativa (fase 0), luego el terreno común y la memoria determinista —territorio, grafo del mundo, citas, plazos— (fase 1), luego **actuar en tu municipio** —expedientes y escritos— (fase 2) y después consultar lo que hace —contratación, subvenciones, boletines, ordenanzas, sociedades, presupuestos— (fase 3). Todo lo demás —distribución, verticales, otros territorios, más fuentes, profundidad del BOE, fiscalización y vigilancia— es un **backlog sin número** que se reprioriza con el uso real. Todo es **genérico para cualquier municipio de España** y se valida primero en el territorio de referencia, la Comunidad de Madrid con Leganés (principio IX); los demás territorios llegan como datos y adaptadores, sin tocar las skills. Las skills nacen **genéricas** y las **verticales** (fiscal, laboral, mercantil…) llegan después como especializaciones (ADR 0012). El grafo llega en tres piezas: memoria pronto, asunto con los expedientes, instrumento tarde (ADR 0014).

---

## 0. Cómo leer este roadmap

- **Fase** = épica. Al final de cada fase hay un usuario real que gana algo: tú en Claude Code con tu municipio, luego un pack instalable, luego terceros.
- **Hito** = PR única, rama corta desde `main` (trunk-based), squash-merge con CI verde. Un hito no se cierra sin cumplir la *Definition of Done* (§1).
- **Numerados y backlog.** Solo las fases 0 a 3 tienen hitos numerados y detallados. Lo posterior es un backlog de candidatos sin número, agrupados por tema y sin orden entre grupos (§4, «Backlog»). Un candidato recibe número cuando entra en la fase en curso, y lo que decide que entre es la bitácora de uso `docs/USO.md` (§6).
- Cada hito indica: objetivo, entrega usable, alcance, controles (tests y calidad) y criterio de aceptación. A partir de H5, el objetivo se formula como lo que una skill pasa a poder resolver.
- Historia de la numeración: rehecha el 2026-09-11 (ADR 0008 y 0009); el 2026-09-12 la fase 0 pasó a entregar `boe-legislacion` en H5 (ADR 0012); el 2026-09-13 se reordenó por tiempo hasta el uso desde H6 y el grafo pasó a tres piezas (ADR 0013 y 0014). La correspondencia con la numeración anterior está en `docs/ADR/0013-roadmap-por-tiempo-hasta-el-uso.md`.

---

## 1. Definition of Done (aplica a todos los hitos)

1. `make ci` en verde: `fmt`, `lint`, `test` (con `-race`), `test-integration`, `vuln`, `schema-check`.
2. Tests unitarios offline para el código nuevo; fixtures grabados en `testdata/<fuente>/` si toca red.
3. Sin `net/http`, `os.Exit`, `fmt.Print*` fuera de los paquetes autorizados (lo vigila `depguard`/`forbidigo`, ver §3).
4. Toda salida de applet pasa la validación contra `schemas/*.json` en test, y `make schema-check` falla si `schemas/` no coincide con lo que `--describe` emite.
5. Errores tipados que mapean a un exit code estable; ningún `panic` en rutas de usuario.
6. Si cambia un comportamiento visible: test e2e (`testscript`) actualizado y `CHANGELOG.md` (sección *Unreleased*).
7. Si cambia una decisión de arquitectura: ADR nuevo en `docs/ADR/`.
8. Si toca una fuente externa: fila de `docs/SOURCES.md` actualizada (licencia, TOS, rate limit, fecha de verificación) y caso en `scripts/verify-sources.sh`.
9. Cobertura de `internal/core/**` ≥ 85 % y global ≥ 70 % (umbrales en `codecov.yml`; no son objetivo, son red de seguridad).
10. Si entrega o cambia una skill (desde H5): evals en `evals/<skill>/` (activación y respuesta con citas esperadas) escritas antes que el código y en verde; `SKILL.md` < 300 líneas con frontmatter válido; `references/` y tabla de comandos regeneradas sin drift.
11. Si tiene dimensión territorial: e2e con un municipio cubierto y otro no cubierto, con la cobertura declarada en `data`; ningún caso especial para un municipio en código ni en `data/` (los municipios concretos solo aparecen en fixtures y evals).
12. Desde H7: todo applet que observe identidades devuelve sus operaciones de grafo en el `Resultado` (ADR 0014); un test comprueba que ninguna entra sin `source` y que aplicar dos veces el mismo resultado no duplica nada. Ningún verbo devuelve texto legal leído del grafo.

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
        internal/cli ── Kong, flags globales, mapeo error→exit code, sobre de salida,
                │       entrega de las operaciones de grafo del Resultado al GraphStore
  ┌─────────────┼─────────────────────────────────────────────┐
  │ internal/core (dominio puro, sin I/O)                     │
  │   ids · territorio · cita · plazos · competencia · schema │
  │   define los PUERTOS: Source, Fetcher, Cache, GraphStore, │
  │   Renderer, y los tipos de operación de grafo             │
  └─────────────┬─────────────────────────────────────────────┘
                │ implementan
   ADAPTADORES: internal/source/<fuente>  (Adapter, uno por fuente)
                internal/httpx            (Decorator chain: UA → robots → ratelimit → retry → record/replay)
                internal/cache, store, graph (Repository sobre SQLite)
                internal/render           (Strategy: json | table | markdown)
```

**Qué va a Go y qué no.** Va al binario lo que exige determinismo o verificabilidad: acceso a fuentes, parseo, caché, fechas y plazos, identificadores, hashes y grafo. El razonamiento, la interpretación, la elección de fuentes y la redacción de la respuesta viven en `SKILL.md` y `references/`. Si una capacidad puede resolverla el modelo con fiabilidad a partir de una referencia generada, no se convierte en applet.

**Territorio.** Los datos de cada municipio (INE, provincia, comunidad, DIR3) salen de registros nacionales; lo que varía por territorio se configura por comunidad o por boletín en `data/territorio/` y `data/boletines/`. El municipio de la persona usuaria vive en `.kitlegal/config.yaml`. Fuera del territorio cubierto, cada applet declara la cobertura dentro de `data`.

**Grafo.** Es memoria y validador, nunca fuente del texto que se cita (ADR 0014): dice qué hay que volver a comprobar, no qué dice el artículo. El texto se cita siempre del sobre de una consulta al applet de la fuente, que la caché sirve sin red cuando está vigente.

Patrones que se usan y por qué:

| Patrón | Dónde | Motivo |
|---|---|---|
| Adapter | `internal/source/<x>` | Cada fuente expone la misma interfaz `Source`; cambiar BOE no toca PLACSP |
| Command + Registry | `internal/app` | Un applet = un comando registrado; el multicall solo despacha |
| Decorator / middleware | `internal/httpx` | Retries, rate limit, robots y grabación componen sin acoplarse |
| Strategy | `internal/render` | `--json` manda; tabla y markdown son intercambiables |
| Repository | `cache`, `store`, `graph` | El dominio no sabe que hay SQLite |
| Null Object | `graph` con `--no-graph` | Las operaciones del `Resultado` siempre se entregan; el store nulo descarta |
| Errores tipados (sentinel + `errors.As`) | `internal/cli/errors.go` | `ErrNotFound`→3, `ErrSourceDown`→4, `ErrRateLimited`→5, `ErrHumanRequired`→6 |
| Functional options | constructores de `httpx`, `cache`, `graph`, applets | Configuración explícita y testeable sin globals |
| Configuración por datos | `data/territorio/`, `data/boletines/` | Un territorio o un boletín nuevo es un YAML, no código ni una skill nueva |
| Facade | `pkg/legalkit` (backlog · distribución) | API pública pequeña sobre los internos |

Reglas de dependencia (se hacen cumplir con `depguard`, no con buena voluntad):

- `internal/core/**` no importa nada de `internal/{source,httpx,cache,store,graph,render,cli,app}`.
- Solo `internal/httpx` importa `net/http`.
- Solo `internal/{cache,store,graph}` importan `modernc.org/sqlite` y `database/sql`.
- Solo `internal/cli` y `cmd/` llaman a `os.Exit`.
- Solo `internal/render` escribe en stdout; logs (`log/slog`) siempre a stderr.
- `internal/graph` no importa `internal/source/*` ni `internal/render`: recibe operaciones, no consulta fuentes ni presenta.

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
| Fuzzing | `go test -fuzz` sobre parsers: `ids`, `cita`, parseo XML/Atom; corpus mínimo versionado en `testdata/fuzz/` | H6, H8, H13 |
| Property tests | `testing/quick` o `pgregory.net/rapid` en `plazos` (idempotencia, monotonía, ida y vuelta) | H9 |
| Tests de integración | `//go:build integration`, SQLite en `t.TempDir()`, `httptest.Server` como fuente | H3 |
| Tests e2e | `github.com/rogpeppe/go-internal/testscript` contra el binario compilado: `kitlegal …`, `boe …` vía symlink, exit codes, `--json`, `--offline` | H1 |
| Matriz territorial | e2e por applet con dimensión territorial: municipio cubierto (Leganés), no cubierto (Tordesillas) y, cuando aplique, foral | H6 |
| Tests de contrato | Cada salida validada contra `schemas/*.json` (`santhosh-tekuri/jsonschema`); `--describe` genera el schema desde H1 y `make schema-check` falla si `schemas/` está desactualizado | H4 (primer applet con fuente) |
| Grafo | Idempotencia de `Apply`; ninguna operación sin `source`; `case.db` solo con ids del mundo; ningún verbo de `graph` devuelve texto legal; `Persona` rechaza NIF/DNI por regex | H7 / H10 |
| Smoke contra red | `scripts/verify-sources.sh` en CI nightly; abre issue si una fuente cambia | H4 |
| Comprobación mecánica de skills | En `make ci`: frontmatter válido, `SKILL.md` < 300 líneas, `references/` y tabla de comandos idénticas a lo generado | H5 |
| Evals de skills | `evals/<skill>/*.yaml` (pregunta, si debe activar la skill, comandos esperados, citas esperadas por identificador); ~~job semanal con modelo barato~~ job disparado por los cambios que las evals miden, con el modelo del uso real de la skill, repeticiones y umbral, y modelos informativos que se publican sin decidir [enmienda 2026-09-16, ADR 0016: «modelo barato» no era un criterio de calidad —el proyecto usa suscripción, no API de pago por uso— y medía el peor caso; la ejecución semanal sobre la rama principal no medía ningún cambio] | H5 |
| Vulnerabilidades | `govulncheck` en cada PR | H0 |
| SAST | `gosec` (en golangci) + CodeQL semanal | H0 |
| Secretos | `gitleaks` en pre-commit y CI | H0 |
| Cadena de suministro | `go mod verify`, `go.sum` obligatorio, Dependabot semanal, SBOM (syft) + firma (cosign) + provenance en release | H0 / H19 |
| Dependencias mínimas | `go mod tidy -diff` en CI; una dependencia nueva requiere justificación en la PR | H0 |
| Observabilidad local | `log/slog` a stderr, `--verbose`, `KITLEGAL_LOG=debug`; `--dry-run` imprime la petición sin ejecutarla | H1 |
| Pre-commit | `lefthook`: `gofumpt`, `golangci-lint run --fast`, `gitleaks`, `go mod tidy -diff` | H0 |
| Commits / versiones | Conventional Commits, SemVer, `CHANGELOG.md` generado por goreleaser | H0 / H19 |
| Docs de decisiones | `docs/ADR/NNNN-*.md` (formato MADR corto) | H0 |

Dependencias fijadas (fuera de esta lista, justificar): `alecthomas/kong`, `modernc.org/sqlite`, `stretchr/testify`, `rogpeppe/go-internal` (testscript), `golang.org/x/time/rate`, `temoto/robotstxt`, `gopkg.in/yaml.v3`, `invopop/jsonschema` (generación), `santhosh-tekuri/jsonschema` (validación en tests). En el hito de distribución (backlog): `modelcontextprotocol/go-sdk`.

---

## 4. Hitos

### Fase 0 — Fundación: la primera skill con su herramienta (Ahora)

Al terminar: la skill `boe-legislacion` responde en Claude Code a consultas sobre cualquier norma consolidada del BOE con el binario Go, medida con evals; `kitlegal boe articulo BOE-A-2015-10565 a21` funciona con caché y `make install` deja binario y skill en su sitio. No hay release todavía: llega al cerrar la fase 3 (H19), o antes si alguien distinto de quien desarrolla tiene que instalarlo (ADR 0013). La skill `boe-fiscal` (Python, fuera de este repo) sigue funcionando como hasta ahora y se migra como primera vertical (backlog) cuando se quiera, sin bloquear nada.

#### H0 · Esqueleto del repo y gates de CI
- **Objetivo**: un repo vacío pero blindado. Cualquier línea de Go que entre después pasa por los controles.
- **Entrega**: `make build test lint vuln` en verde; `kitlegal version` imprime versión, commit y fecha.
- **Alcance**: `go.mod` (módulo `github.com/jmorenobl/kitlegal`, Go estable actual), `cmd/kitlegal/main.go` mínimo, `Makefile` (`build test test-integration test-e2e lint fmt vuln schema-check skills-sync release ci install`), `.golangci.yml`, `lefthook.yml`, `.github/workflows/{ci,codeql,nightly}.yml`, `.github/dependabot.yml`, `codecov.yml`, `LICENSE` (decisión open-core: Apache-2.0 en el núcleo), `README.md`, `CONTRIBUTING.md`, `CHANGELOG.md`, `docs/ADR/0001-multicall.md`, `0002-sqlite-sin-cgo.md`, `0003-no-cendoj-masivo.md`, `0004-frontera-humana.md`.
- **Controles**: todos los de §3 marcados H0.
- **Aceptación**: PR de prueba con un `fmt.Println` en `internal/` falla en lint; PR limpia pasa en < 3 min.

#### H1 · Kernel CLI: multicall, flags globales, exit codes, sobre de salida
- **Objetivo**: el patrón que todos los applets repetirán, escrito una vez.
- **Entrega**: `kitlegal echo hola --json` devuelve el sobre `{ok, fuente, url, fecha_consulta, hash, data}`; `ln -s kitlegal echo && ./echo hola` funciona; `--describe` emite JSON Schema del applet.
- **Alcance**: `internal/cli` (Kong, flags `--json --timeout --offline --dry-run --describe --no-graph --asunto --verbose`, `errors.go` con errores tipados → exit codes, `slog` a stderr), `internal/app` (registro de applets, dispatch por `os.Args[0]` o primer argumento), `internal/core/schema` (tipos del sobre, `hash` sha256 del `data` canónico), `internal/render` (json y tabla mínima). El applet `echo` es de ejemplo: vive en `internal/app/ejemplo`, que el binario distribuido no enlaza (ADR 0010).
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
- **Alcance**: `internal/source/boe` implementando `Source`; cliente, tipos, mapeo a `data`; `schemas/norma.json` y `bloque.json` generados desde `--describe` y vigilados por `make schema-check`. Antes de portar, leer **`refs/boe.py`** (copia congelada del 2026-05-14 de la skill `boe-fiscal`, el único fuente de referencia del porte) línea a línea y anotar comportamientos no obvios en `internal/source/boe/doc.go`. Su dependencia `network_utils.robust_request` no se porta: reintentos, ritmo y `robots.txt` ya son de `internal/httpx`; los TTL por verbo de `boe.py` sí se portan como `TTL` del `Source`. Los ids naturales que el applet observa (`BOE-A-…`, ELI, id de bloque, hash del texto) se conservan en los tipos: en H7 se emiten al grafo (ADR 0014).
- **Controles**: fixtures grabados para cada verbo; golden files de la salida JSON; e2e testscript de los seis verbos en modo replay; `scripts/verify-sources.sh` con el caso `boe articulo` (nightly); fuzz ligero sobre el parser de ids de bloque (`a21`, `da3`, `dt1`).
- **Aceptación**: diff entre la salida de `boe.py articulo` y `kitlegal boe articulo` (campo `data`) vacío para 5 artículos de 3 leyes; tiempo < 200 ms en caché.

#### H5 · Skill `boe-legislacion` (genérica) + andamiaje de skills (`skills-sync`, evals)
- **Objetivo**: que un agente en Claude Code sepa consultar y citar cualquier norma consolidada del BOE (administrativa, contratación, local, fiscal…), y dejar el andamiaje con el que se miden todas las skills siguientes. Es la primera skill del repo y la base de la que las verticales (backlog) se especializan.
- **Entrega**: `skills/boe-legislacion/` con `SKILL.md` (protocolo: identificar la norma, resolver `BOE-A-…`, leer índice y bloques con `scripts/boe`, citar por identificador y bloque, distinguir ley y reglamento, señalar variación autonómica), `scripts/boe` (symlink al binario) y `references/normas.md` generado desde `data/normas.yaml`; `make install` copia el binario a `$GOBIN` y enlaza `skills/*` en `~/.claude/skills/`; formato común de evals y `evals/boe-legislacion/` con 10 preguntas de materias distintas (LPAC, LCSP, LRBRL, LGT, TRLRHL…), al menos una de ellas reproduciendo una consulta que hoy resuelve `boe-fiscal`, para no perder comportamiento cuando se migre.
- **Alcance**: `data/normas.yaml` con las leyes vertebrales que la skill referencia, sin campo de vertical (el campo entra con la primera vertical) y con los ids `BOE-A-…` de las normas incluidas verificados con `boe buscar` (la tabla completa se cierra en H6); schema propio en `schemas/normas.yaml.json`; `scripts/skills-sync.sh` (genera `references/` y la tabla de comandos de `SKILL.md` desde `--describe`); cabecera "generado, no editar"; formato de eval (pregunta, si debe activar la skill, comandos esperados, citas esperadas por identificador); comprobación mecánica de skills dentro de `make ci` (frontmatter, líneas, drift de `references/` y de la tabla de comandos). El protocolo de `SKILL.md` parte del de `boe-fiscal` generalizado a cualquier materia.
- **Controles**: CI comprueba que `references/` y la tabla de comandos regenerados no difieren de lo commiteado (drift); validación del YAML contra su schema; ~~eval de la skill en job manual/semanal~~ eval de la skill en job manual, por etiqueta y por los cambios que las evals miden [enmienda 2026-09-16, ADR 0016: se quita la programación semanal, que sobre la rama principal no medía ningún cambio; el job decide con el modelo del uso real, repite cada eval con umbral y publica la tasa, y los modelos y las evals informativos se publican sin decidir]; las citas esperadas se comparan por identificador (`BOE-A-…` + bloque).
- **Aceptación**: en una sesión de Claude Code, la skill responde una consulta sobre una norma no fiscal (p. ej. «¿qué dice el art. 21 de la Ley 39/2015?») y otra fiscal citando `BOE-A-…` y artículo, con el binario Go y sin Python instalado; las 10 evals pasan.

#### H5.1 · Avisos de vigencia en las evals
- **Objetivo**: que las evals midan lo que hoy no comprueba nadie: que la skill **traslada a su respuesta** los avisos de vigencia que emite el binario. No entrega skill nueva; protege la que ya existe (constitución, principio VIII). Entra desde la bitácora `docs/USO.md`, entrada del 2026-09-16 «Ninguna eval comprueba qué hace la skill ante una norma derogada» (ADR 0013, §6). Va antes de H6 para que el formato de eval deje de crecer justo cuando un hito empieza a apoyarse en él.
- **Entrega**: `SKILL.md` de `boe-legislacion` fija la forma con la que la respuesta traslada cada aviso de vigencia, como ya fija la de la cita; el formato común de eval gana `avisos`, la lista de códigos de aviso que la respuesta tiene que llevar; `Juzgar` los reparte en encontrados y ausentes como hace con las citas, un aviso ausente impide que la eval pase, y el informe los publica; y `evals/boe-legislacion/` gana una eval sobre una norma derogada que los exige.
- **Decisión del mecanismo** [2026-09-16, Jorge, al revisar el plan del primer run del hito]: el aviso se compara **por su forma fija, como la cita por su identificador**, y no leyendo la redacción libre de la respuesta. El primer run llegó a un plan que juzgaba la respuesta con listas cerradas de expresiones por aviso y reglas de atribución de cada expresión a una norma: cumplía el alcance, pero no escala —cada aviso, skill o redacción nueva pide listas y reglas nuevas, y el propio plan declaraba frases que sus reglas leen mal—. Un juicio semántico (modelo o similitud) queda descartado: lo prohíbe este alcance y la constitución (capa 1), y la similitud no separa «ha sido derogada» de «no ha sido derogada». Con la forma fija, lo que se mide es una regla escrita de la skill, la negación deja de importar —una respuesta que afirma lo contrario no lleva la forma— y un aviso nuevo del binario es una etiqueta más, sin reglas nuevas; además, el aviso llega siempre visible y con la misma forma a quien usa la skill.
- **Alcance**: `skills/boe-legislacion/SKILL.md` (regla 3 y paso 5, y donde explica cómo se cita): cada aviso de vigencia del sobre se traslada con su forma fija, `⚠` seguido de la etiqueta del aviso tal como la da el binario (`NORMA DEROGADA`, `VIGENCIA AGOTADA`, `TEXTO POSIBLEMENTE DESACTUALIZADO`) y dos puntos, con la frase del binario o una explicación detrás; sin nombrar evals, el job ni modelos (FR-077 de H5), con la región generada intacta y menos de 300 líneas. `internal/source/boe`: exportar la etiqueta de cada código de aviso, única fuente de verdad de la forma, sin cambiar códigos, frases, condiciones ni salida del applet. `schemas/eval.yaml.json` e `internal/evals/formato.go` (campo `avisos`, valores del enumerado de `internal/source/boe.CodigosDeAviso()`, permitido solo con `activa: true`). La comparación en `internal/evals/juzgar.go`, **mecánica y sin juicio de ningún modelo**: un aviso está en la respuesta si la respuesta lleva su forma fija, con tolerancia solo a lo que no cambia qué aviso es (la variante de presentación del emoji, el énfasis de Markdown, los espacios y las mayúsculas) y exacta en la etiqueta, como la extracción de la cita tras T046 de H5 es tolerante con lo que rodea al identificador y exacta en él; el plan fija esa tolerancia al detalle. Así **distingue una respuesta que traslada el aviso de otra que afirma lo contrario**: la que dice que la norma sigue en vigor, o traslada el aviso con otra redacción, no lleva la forma y el aviso queda ausente. `avisos_encontrados`, `avisos_ausentes` y sus motivos en `internal/evals/informe.go`. La norma de la eval es `BOE-A-1992-26318` (Ley 30/1992): sus metadatos, ya grabados en H4, dan `estatus_derogacion` y `vigencia_agotada` afirmativos —los dos avisos `derogada` y `vigencia-agotada`—, su bloque `a42` también está grabado y su `buscar` lo resuelve la búsqueda grabada en H4 `procedimiento administrativo común` (`internal/source/boe/testdata/golden/buscar-procedimiento-administrativo-comun.json`); falta grabar su `indice`, que FR-074 exige del índice y los metadatos de toda norma de una eval. Eso entra en `testdata/evals/grabaciones.json` y lo graba una persona en una tarea `[datos]` con pausa, nunca dentro de un job. La norma entra en `data/normas.yaml`, que obliga a regenerar `references/normas.md` con `make skills-sync`. La eval nace `informativa: true` (ADR 0016): así no toca la regla «positivas», que pide exactamente 10 que deciden, ni «materias distintas».
- **Fuera de alcance**: promover la eval a decisoria, que obligaría a enmendar FR-062 y su regla, y que se decide con los datos de varias ejecuciones; marcar la derogación en `data/normas.yaml` o en `references/`, porque el aviso lo emite el binario en el momento de consultar y no la tabla; extender los avisos a otras skills o fuentes; juzgar la redacción libre de la respuesta con listas de expresiones, reglas de atribución a normas o cualquier forma de similitud (ver *Decisión del mecanismo*), incluido el caso de una respuesta que lleva la forma fija y además dice lo contrario, que queda como limitación declarada y visible en el informe, que publica la respuesta; cambiar en `SKILL.md` algo distinto de la forma de trasladar los avisos; y tocar `internal/source/boe` más allá de exportar las etiquetas de aviso.
- **Controles**: `make ci` en verde, con el drift de lo generado y las reglas del conjunto de evals; tests de formato (un `avisos` con un código desconocido o en una eval de no activación es un fichero mal formado); de `Juzgar` (aviso esperado con su forma fija, también con las variantes toleradas; ausente; trasladado con otra redacción, que queda ausente; y una respuesta que niega el aviso, que no puede pasar); del informe; una comprobación mecánica en `make ci` de que los códigos del esquema coinciden con `CodigosDeAviso()` y de que `SKILL.md` lleva la etiqueta de cada uno, que falla nombrando el código; y la ejecución del job de evals que abre la propuesta de cambio.
- **Aceptación**: en el informe de esa ejecución, la eval de la norma derogada declara sus dos avisos esperados y su tasa dice en cuántas de sus tres sesiones la skill los trasladó; el veredicto sigue aprobado, porque el hito cambia `SKILL.md` (Definition of Done §1.10); y `red` sigue vacío.

### Fase 1 — Terreno y memoria: territorio, grafo del mundo, citas y plazos (Siguiente)

Al terminar: `legal-core` y `cita-verificada` funcionan para cualquier municipio de España. Una cita en lenguaje natural se resuelve a texto vigente con URL ELI, y el plazo para recurrir un acto de tu ayuntamiento se calcula con los festivos nacionales, autonómicos y locales cuando el territorio está cubierto (Comunidad de Madrid primero), avisando explícitamente cuando no lo está. Y el binario recuerda en `~/.cache/kitlegal/world.db` cada norma, bloque, versión y órgano que ha visto, con su fuente: `graph check` dice si algo de lo consultado ha cambiado o caducado.

#### H6 · `territorio` + skill `legal-core` v0
- **Objetivo**: que toda skill sepa, antes de razonar, en qué territorio está la pregunta y qué normas vertebrales aplican. Es la pieza que hace genéricas a las demás.
- **Entrega**: `kitlegal territorio resolver <nombre|código INE>` devuelve municipio, código INE, provincia, comunidad autónoma, DIR3 del ayuntamiento, régimen (común o foral), boletines aplicables y `cobertura` (qué del territorio está configurado y qué no). Skill `legal-core` v0: protocolo que empieza por identificar el territorio, `references/leyes_vertebrales.md` y `jerarquia_normativa.md`.
- **Alcance**: registro de municipios desde la relación oficial del INE y DIR3 desde el inventario público (formato, licencia y forma de actualización a verificar en tarea `[datos]`, con fila en `docs/SOURCES.md`); `data/territorio/` con configuración por comunidad, solo la Comunidad de Madrid rellena (su boletín, el BOCM, hace de provincial por ser uniprovincial); el resto de comunidades responde con los datos nacionales y cobertura parcial. Nombre ambiguo: lista de candidatos y exit 2. Ids que entran aquí: código INE y DIR3 (`internal/core/ids`, formato y dígito de control), que en H7 son los ids naturales de `Municipio` y `Organo`. Verificar con `boe buscar` los ids de la tabla de leyes vertebrales y fijarlos en `data/normas.yaml`.
- **Controles**: matriz territorial en e2e: Leganés (cubierto), Tordesillas (datos nacionales completos, boletines autonómico y provincial declarados no cubiertos, nada inventado), un municipio de Navarra o del País Vasco (régimen foral marcado), un nombre ambiguo; fuzz del parser de INE y DIR3; evals de `legal-core` que exigen identificar el territorio.
- **Aceptación**: en Claude Code, «¿qué comunidad, provincia y boletines corresponden a mi ayuntamiento?» se responde para cualquier municipio, con la cobertura explícita cuando no es de la Comunidad de Madrid.

#### H7 · `internal/graph`: grafo del mundo, operaciones en el `Resultado` y `graph check` mínimo
- **Objetivo**: que el binario recuerde lo que ha visto, con su fuente, desde el primer applet con fuente y no catorce hitos después: el grafo es una serie temporal y solo vale si lleva grabando (ADR 0014, pieza G0). Lo que una skill gana: `boe-legislacion` v0.1 comprueba con `graph check`, antes de responder sobre una norma ya consultada, si su versión ha cambiado o su consulta ha caducado, y lo dice.
- **Entrega**: `~/.cache/kitlegal/world.db` (ruta con `KITLEGAL_CACHE_DIR`, como la caché) con `nodes/edges/texts`; `Apply` transaccional e idempotente; las operaciones de grafo viajan como campo de `schema.Resultado` y el kernel las entrega al `GraphStore` tras presentar la salida, con la `Procedencia` del propio resultado como `source` de cada nodo y arista (el contrato `Applet` del ADR 0005 no cambia); `--no-graph` es un `GraphStore` nulo; `graph show <id>`, `graph stats` y `graph check` con `version-obsoleta` y `fuente-caducada`, exit 0/1/2. `boe` emite `Norma`, `Bloque`, `BloqueVersion` y `eli:has_part`, `eli:has_version`; `territorio` emite `Municipio`, `Organo` y `lb:pertenece_a`.
- **Alcance**: `modernc.org/sqlite` con el andamiaje de `internal/cache` (migraciones embebidas, WAL, espera ante bloqueo que respeta el contexto); puerto `GraphStore` y tipos de operación en `internal/core`; rechazo por regex de NIF/DNI en `Persona` dentro de `Apply` (constitución VII), aunque el primer emisor de `Persona` sea H17. La tabla `texts` guarda el cuerpo por hash para `check` y para el futuro `graph history`, **nunca para responder**: ningún verbo de `graph` devuelve texto legal. **Fuera de alcance**: `eli:cites` por regex sobre el texto (entra con `boe analisis`, backlog), FTS5, `graph query` con SQL libre, `neighbors`, `path`, exportaciones, promoción de props a columnas, grafo del asunto (H10).
- **Controles**: integración con DB temporal; idempotencia (aplicar dos veces el mismo `Resultado` = un nodo, una arista); test de que ninguna operación entra sin `source` y de que el `source` coincide con la procedencia del sobre; e2e: consultar un bloque con fixture A, sustituir por fixture B con `fecha_vigencia` posterior, `graph check` marca `version-obsoleta`; e2e: la salida de `boe articulo` es byte a byte la misma con y sin `--no-graph`; test de arquitectura (`internal/graph` no importa `source/*` ni `render`); evals de `boe-legislacion` con una consulta repetida tras cambio de versión, en las que la respuesta vuelve a pasar por `boe articulo` y no por `graph show`.
- **Aceptación**: tras `boe articulo` y `territorio resolver`, `graph stats` muestra los nodos y aristas con su fuente; con `--no-graph` `world.db` queda intacta; `graph check` sobre un bloque cuya versión cambió en fixture devuelve exit 1 con explicación citable.

#### H8 · `internal/core/cita` + applet `cita` + skill `cita-verificada`
- Parser de `"art. 21.1 Ley 39/2015"`, `"artículo 118 LCSP"`, `"DA 3ª LGT"` → `{norma, bloque}`; resolución a ELI vía `data/normas.yaml` y a texto vía `boe articulo`. Verbos `resolver`, `validar`. Exit 3 si el bloque no existe en el índice. Las normas autonómicas se resuelven igual (el BOE consolida legislación de las comunidades); `data/normas.yaml` empieza con las estatales vertebrales y las autonómicas que citen las skills del territorio de validación. Ids que entran aquí: `BOE-A-…` y ELI; como `Norma` ya usa ELI como id de nodo (H7), `cita` resuelve directamente al id del grafo y nace emitiendo. `cita-verificada` apoya `validar` en `graph check`: una cita a una versión que el mundo sabe superada se devuelve con aviso.
- Controles: fuzz del parser y de los ids; corpus de 100 citas reales en `testdata/cita/corpus.txt` con resultado esperado (incluye citas autonómicas); e2e; evals de `cita-verificada`.

#### H9 · `internal/core/plazos` + applet `plazos` + skill `legal-core` v1
- Días hábiles/naturales, arts. 30-31 LPAC (incluido el art. 30.6: inhábil en el municipio del interesado o en la sede del órgano), agosto inhábil judicial. Tipos: alzada, reposición, contencioso, LTAIBG, silencio. Festivos desde `data/festivos/` (con `source`): nacionales y autonómicos de todas las comunidades (se publican cada año en el BOE) y locales de los municipios del territorio cubierto (la Comunidad de Madrid publica los de todos sus municipios; formato a verificar en tarea `[datos]`). Fuera de cobertura, el resultado avisa en `data` de que faltan los festivos locales y nunca los da por inexistentes.
- `legal-core` v1: `references/recursos_y_plazos.md` y `reglas_de_cita.md` generadas; la skill usa `territorio` y `plazos`.
- Controles: property tests (añadir N días hábiles nunca cae en festivo; inverso consistente); tabla de casos con fecha esperada verificada a mano, incluido un plazo en Leganés que cruza un festivo local y el mismo plazo en Tordesillas con aviso de cobertura; evals.
- Al cerrar: hay cinco applets (`boe`, `territorio`, `graph`, `cita`, `plazos`). Momento de revisar si `internal/cli` pide refactor, no extracción. Es una nota del cierre, no un hito.

### Fase 2 — Actuar en mi municipio (Después)

Al terminar: registras lo que has pedido a tu ayuntamiento y el kit te avisa del silencio; una solicitud de acceso a información o un recurso de reposición salen listos para firmar, con la fundamentación resuelta por `cita`, el órgano por `territorio` y el plazo por `plazos`. Es el primer bucle completo —consultar, actuar, seguir— y llega antes que las fuentes municipales porque no depende de ninguna (ADR 0013). Su primera versión fundamenta solo con normativa estatal; los escritos que se apoyen en ordenanzas esperan a la fase 3.

#### H10 · Grafo del asunto + `expediente` + skill `seguimiento-expedientes`
- `.kitlegal/case.db` y `config.yaml` (con el municipio de la persona usuaria), nodo `Consulta`, `lb:en_asunto`, `--asunto`; `ATTACH` de `world.db` (ADR 0014, pieza G1). `expediente add | listar | vencimientos`: solicitudes y recursos presentados, con los silencios calculados por `plazos` según la norma aplicable (la estatal por defecto; la autonómica cuando el territorio la configure). `graph check` gana `plazo-vencido` y `plazo-sin-base`. La skill responde «¿qué tengo pendiente y qué vence?» y avisa del silencio que habilita reclamación.
- Controles: e2e de un flujo completo (registrar una solicitud, avanzar el reloj, `vencimientos` avisa; simular nueva versión en fixture y `check` marca warning); test de que `case.db` nunca contiene texto ni props del mundo, solo ids; ningún comando saca el asunto de `.kitlegal/` (la exportación es backlog y pedirá confirmación); evals.

#### H11 · `escrito generar` (mínimo) + skill `redaccion-escritos` + `render` markdown
- Tipos iniciales: solicitud de acceso a información (LTAIBG art. 17) y recurso de reposición contra un acto municipal (LPAC arts. 123-124). Salida en markdown desde plantillas en `internal/docgen`; `internal/render` gana la forma markdown (Strategy), que hasta aquí nadie pedía; órgano destinatario desde `territorio`; fundamentación resuelta por `cita` (una cita que no se resuelve sale como `[SIN FUNDAMENTO]` y exit 3); plazo por `plazos`. Dónde presentarlo: el Registro Electrónico General de la AGE, que admite escritos dirigidos a cualquier administración (art. 16.4 LPAC), y la sede del ayuntamiento cuando conste. Termina siempre en fichero + URL + exit 6. Emite `Escrito` (hash del fichero) y `lb:en_asunto` al asunto.
- Controles: test de arquitectura de que no existe ninguna llamada HTTP con método distinto de GET/HEAD en todo el módulo; golden de escritos; evals. Fuera de alcance (backlog · profundidad): DOCX, `Afirmacion`/`fundamenta` en el grafo, `check` bloqueante y más tipos de escrito.

### Fase 3 — Consultar mi municipio

Al terminar: para cualquier municipio, las skills responden qué contrata y qué subvenciona su ayuntamiento, qué se publica sobre él en boletines y edictos, qué dicen sus ordenanzas, qué sociedades y administradores aparecen en BORME y cómo van sus presupuestos. Validado en Leganés. Fuera de la Comunidad de Madrid, contratación, subvenciones, BORME y presupuestos funcionan igual (son fuentes nacionales) y boletines y ordenanzas declaran su cobertura. Todos los applets nacen emitiendo al grafo, y la fase cierra con el release `v0.1.0`.

#### H12 · Spike de fuentes municipales (timebox: 1 día)
- Verificar los feeds Atom de PLACSP (qué órganos publican directamente, qué llega agregado desde plataformas autonómicas y si los contratos menores agregados están incluidos), los endpoints de BDNS (swagger), el BOCM (sumario, buscador, formatos, TOS) y el Tablón Edictal Único. Resultado: `docs/SOURCES.md` con filas completas y fixtures grabados. Sin código de producto. BORME y los datos abiertos de Hacienda se verifican en sus hitos (H17, H18).

#### H13 · `placsp sync | licitaciones | organo | adjudicatario` + `internal/store` + skill `contratacion-publica` v0
- Sync incremental de Atom + CODICE XML a `store`; el órgano se identifica por su DIR3, obtenido con `territorio` a partir del municipio. Skill `contratacion-publica` v0: qué ha licitado y adjudicado el ayuntamiento, a quién y por qué importe, incluidos los contratos menores. Si el órgano publica en una plataforma autonómica cuyos datos no llegan a PLACSP, la salida lo declara en la cobertura. Emite `Organo`, `Licitacion`, `Contrato`, `Entidad` y `lb:convoca`, `lb:licita`, `lb:adjudica`, `lb:modifica_contrato` con las props que las anomalías (backlog) necesitarán (importes, procedimiento, número de licitadores, fechas). Sin anomalías.
- Es el hito más pesado del roadmap. Si el plan lo desborda, se parte en dos sin renumerar: sync + `store` + `licitaciones` con la skill v0 primero; `organo`, `adjudicatario` y el resto después.
- Controles: fuzz del parser CODICE; integración de sync incremental (segunda ejecución no duplica, ni en `store` ni en el grafo); e2e con replay para Leganés y para un municipio de otra comunidad; evals.

#### H14 · `bdns convocatorias | concesiones | beneficiario` + skill `subvenciones` v0
- REST; órgano por DIR3 vía `territorio`. La skill responde qué subvenciones convoca y concede el ayuntamiento y a quién. Emite `Convocatoria`, `Concesion`, `Entidad`, `lb:convoca`, `lb:concede`. Mismos controles que H13.

#### H15 · `boletin` (motor genérico, BOCM) + `edictos` (TEU) + skill `bop-y-edictos`
- `boletin sumario | buscar` con un motor genérico configurado por YAML (`data/boletines/bocm.yaml` es la primera y única configuración) y `edictos buscar` sobre el Tablón Edictal Único, que es nacional. El filtro por municipio usa `territorio`. Añadir otro boletín es un YAML más, y un adaptador solo si el motor no basta (backlog · territorios). Emite `Publicacion` y `lb:publica`.
- Controles: fixtures de BOCM y TEU; e2e con Leganés (BOCM + TEU) y Tordesillas (TEU, más la cobertura que declara no configurados los boletines autonómico y provincial); `verify-sources`; evals.

#### H16 · Ordenanzas vía boletín + skill `ordenanzas-locales`
- Las ordenanzas se publican íntegras en el boletín provincial o, en comunidades uniprovinciales, en el autonómico (art. 70.2 LRBRL; ordenanzas fiscales, art. 17.4 TRLRHL). `ordenanzas listar | ver --municipio` busca con `boletin` los anuncios de aprobación definitiva y de modificación del territorio, sin crawler por ayuntamiento. El texto sale del formato que publique el boletín; si solo hay PDF, la dependencia de extracción se justifica en el plan. La skill presenta el texto aprobado y las modificaciones que encuentre, citando cada anuncio, sin afirmar que es un texto consolidado.
- Controles: fixtures; e2e con la matriz territorial; evals («¿qué dice la ordenanza de terrazas de mi municipio?»). El crawler de sedes municipales queda para el backlog de fuentes.

#### H17 · `borme sumario | empresa | administrador` + skill `entidades-y-registros`
- Vía API BOE; `Persona` sin NIF, `lb:administra` con `valid_from/to`, `lb:participa`. La skill responde qué sociedades y administradores aparecen en BORME para una entidad o un nombre, con el histórico de nombramientos y ceses.
- Controles: test explícito de que `Persona` rechaza cualquier prop que parezca NIF/DNI (regex) en `graph.Apply` (la regla existe desde H7; aquí se prueba con el primer emisor real); fixtures; evals.

#### H18 · `presupuesto descargar | comparar` + skill `presupuestos-y-cuentas`
- Presupuestos y liquidaciones de entidades locales desde los datos abiertos del Ministerio de Hacienda, por código INE: fuente nacional, sirve para cualquier municipio (endpoint y formato a verificar en tarea `[datos]`). `comparar --aprobado --liquidado` da desviaciones por capítulo y programa.
- Controles: fixtures; e2e con la matriz territorial; evals.

#### H19 · Release `v0.1.0`
- **Objetivo**: instalación reproducible y firmada; a partir de aquí cada hito puede publicar. Se adelanta a cualquier punto anterior en cuanto alguien distinto de quien desarrolla tenga que instalarlo.
- **Entrega**: `.goreleaser.yaml` (darwin/linux/windows, arm64/amd64, `CGO_ENABLED=0`, `-trimpath`, ldflags de versión), checksums, SBOM, firma cosign keyless, changelog automático, tag `v0.1.0`; script `install.sh` y tap Homebrew opcional.
- **Controles**: `goreleaser check` en CI; job de release solo en tag; smoke post-release (descarga el binario publicado y ejecuta `boe articulo` en `--offline` con caché vacía → exit 4 esperado, y `version`).
- **Aceptación**: `brew install` o `curl | sh` en una máquina limpia y `kitlegal boe articulo …` funciona.

### Backlog (sin número; sin orden entre grupos)

Candidatos agrupados por tema. Ninguno tiene número hasta que entra en la fase en curso, y lo que decide que entre es `docs/USO.md` (§6). Entre grupos no hay dependencias salvo una: «fiscalizar y vigilar» necesita las fuentes de la fase 3 y el grafo, y es el único que no puede adelantarse. Cada candidato conserva los controles que tenía cuando era hito numerado (ADR 0013, tabla de correspondencia).

**Distribución** (hojas: solo necesitan applets y `--describe`; entran cuando alguien más tenga que usar esto o quieras usarlo fuera de Claude Code)
- `kitlegal mcp serve --stdio`: tools generadas desde `--describe`, mismo sobre; `modelcontextprotocol/go-sdk`. Test de conformidad (tools == applets registrados); e2e con cliente MCP de prueba; el servidor no expone nada del asunto sin flag explícita.
- `kitlegal skills install | list | doctor` + `plugin/` + mecanismo de packs (`packs/<vertical>/pack.yaml`, `legal` base): instala packs en el directorio del cliente; `doctor` comprueba versión mínima, hash de skills y lanza `verify-sources`. `plugin/.claude-plugin/plugin.json` y `marketplace.json` generados desde `packs/`. E2e en `t.TempDir()` simulando `~/.claude`; drift check de `plugin/`.
- `pkg/legalkit`: facade mínima (`Resolve(cita)`, `Articulo(id, bloque)`, tipos de dominio), compromiso de compatibilidad documentado, `apidiff` en CI, ejemplo compilable. Release `v1.0.0`.

**Verticales** (un hito por vertical, en cualquier momento tras H5; ADR 0012)
- Molde: normas con `vertical: <nombre>` en `data/normas.yaml` (el campo entra con la primera) → skill `<fuente>-<vertical>` que reutiliza el protocolo de la base y añade `references/` generadas, calendario y reglas → evals → `packs/<vertical>/pack.yaml` cuando exista el mecanismo de packs (distribución). Si una vertical necesita una fuente propia (PETETE, DYCTEA, AEAT…), la fuente entra por «fuentes» con sus TOS revisados.
- **Primera: `boe-fiscal`**, migración de la skill Python existente sobre `boe-legislacion`: `references/normas_fiscales.md` generado, mismo comportamiento medido con evals que reproducen sus consultas actuales; las haciendas forales llegan cuando «territorios» cubra el régimen foral.
- Después, con el mismo molde: `laboral`, `mercantil`… en el orden en que se necesiten.

**Territorios** (un hito por territorio, bajo demanda)
- Molde: fila en `docs/SOURCES.md` → configuración en `data/territorio/` y `data/boletines/` → adaptador solo si el motor genérico no basta → fixtures → la fila de la matriz territorial pasa a «cubierto» → evals de las skills afectadas. Las skills no se tocan.
- Candidatos: el boletín autonómico y los BOP de una comunidad multiprovincial (p. ej. BOCYL y BOP de Valladolid, que cubren Tordesillas), DOGC, BOJA…; festivos locales de cada comunidad; plataformas autonómicas de contratación que no lleguen agregadas a PLACSP; consejos autonómicos de transparencia; régimen foral (normativa foral en `boe-legislacion`, haciendas forales en la vertical fiscal, catastros forales).

**Fuentes** (un hito por fuente, independientes; molde de H13: spike de TOS → fixtures → adaptador → emisión al grafo → skill → `verify-sources`)
- Orden sugerido (`refs/mapa` §5): `catastro` → `ine` + `datosgob` → `transparencia` → skill `boletines-autonomicos` (normas autonómicas, sobre el motor de H15) → `eurlex` → `congreso` → `ecli` → `tc` (vías verificadas y hito propuesto en `docs/JURISPRUDENCIA.md`: el TC se resuelve por ECLI contra el sumario del BOE; el TS no tiene vía sin el buscador del CENDOJ) → `sede` (crawler de sedes municipales, solo donde el boletín no baste) → `dgt` y `teac` **solo tras revisar TOS** (pendiente de `refs/00-README.md`).

**Profundidad del BOE y de los escritos**
- `boe sumario | vigilar | eli` + `boe-legislacion` v1: sumario diario, detección de normas actualizadas por rango de fechas, construcción de ELI. Fixtures, golden, `verify-sources` para sumario; evals.
- `boe analisis` → aristas ELI (`eli:amends/repeals/amended_by/repealed_by`), nodos stub, `eli:cites` por regex sobre el texto de los bloques (precisión antes que cobertura), `graph history` con diff entre `BloqueVersion`. Fixtures de una ley con modificaciones; golden del diff. Es lo que hace útiles `norma-derogada` y `remision-no-seguida` en `check`.
- Escritos fundamentados en el grafo: `Afirmacion` + `lb:fundamenta` + `check` completo y bloqueante (`cita-huerfana`, `cita-inexistente`, `norma-derogada`, `norma-no-vigente`, `reglamento-contra-ley`, `competencia-dudosa`…) sin `--force`; salida DOCX; más tipos (alzada, reclamación ante el consejo de transparencia competente según el territorio, alegaciones, recurso especial en contratación). Golden de escritos; e2e de `check` bloqueando.

**Fiscalizar y vigilar** (ADR 0014, pieza G2; necesita la fase 3 entera)
- `graph neighbors | path` + reglas de anomalías de contratación en `data/anomalias/*.yaml` (SQL + umbrales): `fraccionamiento`, `licitador-unico`, `concentracion`, `plazo-publicacion`. Las reglas con umbral legal lo toman de la ley; las relativas comparan con municipios de tamaño parecido (tramos de población del padrón del INE). Salida con `nodos`, `aristas`, `verificar_en`. `contratacion-publica` v1. Cada regla con un grafo sintético que la dispara y otro que no; validación del YAML contra schema; SQL de solo lectura (`SELECT`/`WITH`); ninguna regla referencia un municipio u órgano concreto. Todo cambio en `data/anomalias/` es gate humano (constitución, capa 3).
- Cruces entre fuentes + pack `fiscalizador`: `administrador-comun`, `beneficiario-adjudicatario`, `empresa-recien-creada`. `subvenciones` v1. Validación con casos conocidos de Leganés (fixtures anonimizados de aristas, no de personas) más un municipio de otra comunidad.
- `vigilar run` sobre delta + `anomalies mark --confirmada|--descartada`: orquesta `placsp sync`, `bdns`, `borme sumario`, `boletin`, `edictos` y `boe vigilar` sobre el municipio y las entidades de `config.yaml`; anomalías solo sobre aristas nuevas. E2e de dos ejecuciones consecutivas con fixtures distintos; solo emite la diferencia.
- `graph export --format md|jsonld|dot|csv` con anonimización por defecto; la exportación del asunto pide confirmación.

---

## 5. Lo que deliberadamente NO se hace (hasta que un hito lo pida)

- Framework de DI, ORM, generador de CLI: Kong + composición manual bastan.
- Extraer `internal/cli` a librería (la regla de los tres applets se cumple en H8; aun así, solo refactor interno, revisado al cerrar H9).
- Applets sin una skill que los use, o capacidades en Go que el modelo resuelve con una referencia generada.
- Casos especiales, YAML o código para un municipio concreto; configuración municipal curada a mano.
- Territorios distintos del de validación (BOCYL, BOP de otras provincias, régimen foral…) hasta que el uso los pida (backlog · territorios).
- Skills verticales (fiscal, laboral, mercantil…) que no sean especialización de una skill base ya existente; ni campo `vertical` en `data/normas.yaml` ni packs por vertical antes de la primera vertical.
- Anomalías, cruces, vigilancia y exportaciones del grafo antes de que existan las fuentes de la fase 3 (backlog · fiscalizar y vigilar). El grafo de las fases 1 y 2 es memoria y validador, no instrumento.
- `eli:cites` por regex, FTS5, `graph query` con SQL libre, promoción de props a columnas: sin consumidor hasta `boe analisis` o las anomalías.
- Crawler de sedes municipales mientras el boletín baste.
- Release firmado antes de que alguien distinto de quien desarrolla tenga que instalar (H19 o antes bajo demanda); mientras tanto, `make install`.
- MCP remoto con auth, cuotas o caché compartida.
- Triple store, SPARQL local, embeddings, sincronización entre máquinas.
- Paralelismo en descargas: primero correcto y respetuoso, luego rápido.
- Ingesta de CENDOJ, scraping de PETETE/DYCTEA/HJ sin TOS revisados.
- Cobertura como objetivo: el umbral existe para no retroceder, no para perseguirlo.
- Numerar el backlog: un candidato recibe número cuando entra en la fase en curso, no antes.

---

## 6. Ritual por hito (para que "lean" no degenere en "sin control")

1. Abrir rama `hNN-nombre-corto`; escribir primero las evals de la skill que el hito entrega o mejora (desde H5) y el test e2e (`testscript`) que describe la entrega.
2. Planificar de fuera adentro (qué debe resolver la skill → qué herramientas necesita) e implementar de dentro afuera: `core` → adaptador → applet → skill.
3. `make ci` local; PR con plantilla (objetivo, alcance, controles añadidos, decisiones, pendientes).
4. Revisión: `/code-review` y `/security-review` sobre la PR; si toca fuentes, `docs/SOURCES.md`.
5. Squash-merge; si el hito cierra una fase y existe el release (H19), tag y release.
6. Actualizar este roadmap solo si cambia el orden o el alcance; el detalle vive en las PR y los ADR.
7. **Bitácora y repriorización.** Quien usa el kit apunta en `docs/USO.md` qué pidió, qué falló y qué faltó. Cada tres o cuatro hitos, antes de abrir el siguiente, se relee la bitácora: el siguiente hito sale de ahí (un fallo, una carencia, una fuente que el uso pidió) o del backlog; ante discrepancia gana la bitácora y este roadmap se actualiza (ADR 0013). Un candidato del backlog que entra recibe el siguiente número libre.
