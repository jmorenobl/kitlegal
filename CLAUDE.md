# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Estado del repositorio

`kitlegal` es un conjunto de **skills agénticas** para consultar fuentes legales públicas españolas (BOE, PLACSP, BDNS, BORME, EUR-Lex…) y actuar en el propio municipio, apoyadas en un binario Go multicall que les da **herramientas deterministas**. El producto son las skills; el binario es su herramienta (constitución, principio VIII; ADR 0008). Todo es genérico para cualquier municipio de España y se valida primero en la Comunidad de Madrid con Leganés (principio IX; ADR 0009). En `main` están H0 (esqueleto y gates de CI: `go.mod`, `Makefile`, CI, lint, `codecov.yml`), H1 (kernel de la CLI: `internal/cli`, `internal/app`, `internal/core/schema`, `internal/render`; applets de ejemplo solo en tests) y H2 (`internal/httpx`: la única puerta a la red; `httpx.New` da un cliente con contexto obligatorio, identificación, `robots.txt`, ritmo por sitio, reintentos y errores con clase; `KITLEGAL_RECORD=1` graba y `httpx.Replay(dir)` reproduce sin abrir ninguna conexión; `schema.ConClase` y `Resultado.Ensayo` amplían el kernel) y H3 (`internal/cache`: caché SQLite con TTL en `~/.cache/kitlegal/cache.db`, ruta con `KITLEGAL_CACHE_DIR`; con `--offline` solo lee y falta = exit 4; la clave y la vigencia las fija cada fuente). Todavía no hay ninguna fuente legal: el siguiente hito es H4 (applet `boe`).

**Orden del roadmap: por tiempo hasta el uso** (ADR 0013). Fases numeradas 0 a 3 —fundación (H0–H5), terreno y memoria (H6–H9: `territorio`, grafo del mundo, `cita`, `plazos`), actuar en el municipio (H10–H11: expedientes y escritos) y consultar el municipio (H12–H19: PLACSP, BDNS, boletines, ordenanzas, BORME, presupuestos, release)—; todo lo demás (distribución, verticales, territorios, fuentes, profundidad del BOE, fiscalización y vigilancia) es backlog sin número que se reprioriza con `docs/USO.md`. El grafo llega en tres piezas (ADR 0014): memoria en H7, asunto en H10, instrumento en el backlog.

**Método de trabajo: spec-kit por hito.** El proyecto se implementa hito a hito (`docs/ROADMAP.md`) con el workflow `hito` de spec-kit (`.specify/workflows/hito/workflow.yml`, documentado en `docs/WORKFLOW.md`). La constitución `.specify/memory/constitution.md` recoge principios, restricciones y el «Criterio de decisión autónoma» que rige `clarify` y los gates automáticos: siempre la mejor solución sin atajos; lo no especificado no se implementa; escalar (parar) ante alcance, frontera humana, privacidad, TOS o decisiones cerradas. Artefactos por hito en `specs/NNN-hN-slug/`. Lanzar con `scripts/hito.sh H<n>`, que además supervisa el run: reanuda ante límites de uso o fallos transitorios y se detiene ante gates y paradas deliberadas. Las tareas que necesitan la plataforma remota llevan `[plataforma]`; el workflow empuja la rama del hito y abre la propuesta de cambio, y fusionar es siempre humano (gancho `pre-push`; ADR 0007).

Skills de agente instaladas (no son parte del producto): el conjunto `samber/cc-skills-golang` vive en `.agents/skills/` (registro en `skills-lock.json`); `.claude/skills/` son symlinks a ese directorio. Al escribir Go, `golang-how-to` orquesta el resto (`golang-cli`, `golang-project-layout`, `golang-testing`, `golang-database`…).

Antes de escribir código, lee los documentos semilla de `refs/` en este orden:

1. `refs/mapa-sistema-legal-skills.md` — sistema legal español, fuentes con semáforo de automatizabilidad (🟢 API / 🟡 scraping / 🔴 requiere identidad humana) y catálogo de skills.
2. `refs/kitlegal-estructura-y-ecosistema.md` — estructura del monorepo, convenciones del multicall, packs por vertical, distribución (skills, plugin, MCP, librería Go).
3. `refs/kitlegal-grafo.md` — grafo legal sobre ELI en SQLite (mundo público + asunto privado), alimentación por uso, `graph check` y reglas de anomalías.

`refs/00-README.md` resume el estado, el primer hito y las decisiones cerradas. Ante conflicto, la constitución y `docs/ROADMAP.md` prevalecen sobre `refs/`.

## Primer hito con fuente

`kitlegal boe articulo BOE-A-2015-10565 a21` funcionando en Go con caché SQLite (H4), y la skill genérica `boe-legislacion` (H5): consultar y citar cualquier norma consolidada del BOE con el binario. H2 (`internal/httpx`) y H3 (`internal/cache`) ya están en `main`. La skill `boe-fiscal` (hoy Python `scripts/boe.py`, fuera de este repo) es el patrón: `boe.py` se porta a `internal/source/boe` y su protocolo, generalizado a cualquier materia, es el de `boe-legislacion`. La propia `boe-fiscal` se migra después como primera vertical, sobre la base (ADR 0012).

## Comandos

El `Makefile` es la única superficie de invocación; `make help` lista los objetivos (`build`, `test`, `lint`, `ci`, `hooks`, `skills-sync`, `release`…). Tooling:

- Módulo `github.com/jmorenobl/kitlegal`, CLI con Kong (`github.com/alecthomas/kong`), SQLite sin cgo (`modernc.org/sqlite`), release con goreleaser.
- Tests unitarios offline contra fixtures grabados en `testdata/<fuente>/`; grabar con `KITLEGAL_RECORD=1`. Solo `scripts/verify-sources.sh` (CI nightly) toca la red.
- `scripts/skills-sync.sh` regenera `skills/*/references/*.md` desde `data/*.yaml` y los symlinks de `scripts/`.

## Decisiones ya tomadas (no reabrir)

- **Skills primero**: cada hito entrega o mejora una skill medible con evals, o protege las existentes. Solo va a Go lo que exige determinismo o verificabilidad (fuentes, parseo, caché, fechas y plazos, ids, hashes, grafo); una herramienta que ninguna skill usa no se construye. Se planifica de fuera adentro (skill → herramientas) y se implementa de dentro afuera.
- **Base antes que vertical** (ADR 0012): las skills nacen genéricas (`boe-legislacion`, `legal-core`, `cita-verificada`…). Las verticales (fiscal, laboral, mercantil…) son especializaciones que reutilizan el protocolo de una skill base y añaden normas con `vertical:` en `data/normas.yaml`, `references/` propias y evals; no añaden herramientas. Llegan bajo demanda desde el backlog del roadmap (grupo «verticales»), en cualquier momento tras H5; `boe-fiscal` es la primera.
- **Genericidad territorial**: ningún caso especial para un municipio en código, `data/` ni skills. Datos de municipio desde registros nacionales (INE, DIR3); lo territorial se configura por comunidad o boletín en `data/territorio/` y `data/boletines/`; el municipio del usuario va en `.kitlegal/config.yaml`. Territorio de validación: Comunidad de Madrid (Leganés); otros territorios (BOCYL, BOP multiprovinciales, forales…) desde el backlog («territorios»), bajo demanda. Fuera de cobertura, la salida la declara dentro de `data` y ninguna skill concluye «no existe» sin cobertura completa.
- **Nombre único** `kitlegal` para repo, módulo, binario y directorios (`.kitlegal/` en cwd para el asunto, `~/.cache/kitlegal/` para caché y grafo del mundo).
- **Multicall**: un solo ejecutable; `os.Args[0]` o el primer argumento elige el applet (`boe`, `placsp`, `bdns`, `cita`, `plazos`, `graph`…). Un symlink `boe -> kitlegal` permite que las skills sigan llamando `scripts/boe articulo …`.
- **Convenciones de agente** en `internal/cli`: flags globales `--json`, `--timeout`, `--offline`, `--dry-run`, `--describe` (emite JSON Schema de entrada/salida, del que se generan las tools MCP y la tabla de comandos de cada SKILL.md), `--no-graph`, `--asunto`. No se extrae `internal/cli` a librería externa hasta que tres applets repitan el patrón.
- **Exit codes estables**: 0 ok · 2 args · 3 no encontrado · 4 fuente no disponible · 5 rate-limited/TOS · 6 requiere identidad humana.
- **Sobre de salida obligatorio** en todo applet: `{ok, fuente, url, fecha_consulta, hash, data}`. Sin `fuente`+`url`+`fecha_consulta`+`hash` no hay cita.
- **Skills sin código**: cada `skills/<nombre>/` lleva `SKILL.md` (< 300 líneas, protocolo de razonamiento + tabla de comandos + reglas), `references/` **generadas** desde `data/*.yaml` (cabecera `<!-- generado desde data/normas.yaml, no editar -->`) y `scripts/` como symlinks al binario. `data/*.yaml` es la única fuente de verdad.
- **Frontera humana**: solo se automatizan fuentes públicas. Presentar escritos, notificaciones y cualquier acción con identidad terminan en "fichero listo para firmar" y exit 6; nunca un POST a una sede.
- **CENDOJ**: solo resolución de ECLI y metadatos; nunca ingesta masiva ni scraping del buscador.
- **HTTP** únicamente vía `internal/httpx` (retries, rate limit por host, `robots.txt`, User-Agent identificable `kitlegal/x.y (+https://ventanillalegal.es/bot)`). Nada de paralelismo agresivo contra PETETE, DYCTEA o HJ del TC.
- **Grafo** (ADR 0014): ids naturales (ELI, ECLI, DIR3, NIF, código INE) en `world.db`; el grafo del asunto (`.kitlegal/case.db`) solo guarda ids del mundo y nunca se sincroniza ni exporta por defecto. `Persona` nunca lleva NIF/DNI. Un dato sin `source` no entra. Las operaciones de grafo viajan en el `Resultado` del applet y el kernel las aplica con la `Procedencia` del propio resultado como `source`; el contrato `Applet` (ADR 0005) no cambia. El grafo es índice y validador, nunca fuente del texto que se cita: dice qué hay que volver a comprobar, no qué dice el artículo.
- Reglas invariantes en todas las skills: nunca inventar contenido legal; cada afirmación con cita resuelta por `cita`; distinguir ley/reglamento; señalar variación autonómica.

## Arquitectura prevista (resumen)

```
skills/<skill>/             EL PRODUCTO: SKILL.md + references/ (generadas) + scripts/ → kitlegal
packs/  data/  evals/       packs por vertical · fuente de verdad (normas, territorio, boletines, festivos, anomalías) · evals de skills
cmd/kitlegal/main.go        multicall → internal/app (registro y dispatch de applets)
internal/cli                Kong + flags globales + exit codes
internal/core/{territorio,cita,plazos,competencia,schema,ids}
internal/source/<fuente>    un adaptador por fuente: Source{Name, Fetch(ctx,req), TTL, Terms}
internal/{httpx,cache,store,graph,analysis,render,docgen}
pkg/legalkit                API pública pequeña y estable
schemas/  testdata/  mcp/  plugin/  docs/
```

Desde H7 (`internal/graph`), todo applet que observe identidades devuelve sus operaciones de grafo en el `Resultado` y el kernel las entrega al `GraphStore` tras presentar la salida (`--no-graph` = store nulo); `boe` y `territorio`, anteriores, lo incorporan en H7 y los demás nacen emitiendo (ADR 0014). Los packs (`packs/<vertical>/pack.yaml`) agrupan skills por vertical (`legal` base; `fiscal`, `laboral`, `mercantil`, `fiscalizador` extienden `legal`) y llegan con la distribución (backlog).

## Pendientes

Los pendientes de **estructura del repositorio** (dónde van los fixtures, los tres directorios `skills`, los symlinks de las skills al binario, `refs/`…) están en `docs/PENDIENTES.md`, cada uno con el hito en que se decide.

## Pendiente de verificar antes de fijar en código o `data/`

- Identificadores BOE de la tabla de leyes vertebrales (algunos están de memoria; comprobar con `kitlegal boe buscar`).
- Endpoints actuales de BDNS (swagger) y nombres de los feeds Atom de PLACSP (y qué llega agregado desde plataformas autonómicas).
- Formato, licencia y actualización de la relación de municipios del INE, del inventario DIR3 y de los festivos locales publicados por la Comunidad de Madrid.
- Términos de uso de PETETE (DGT), DYCTEA (TEAC) y del buscador HJ del TC.

## Convenciones de idioma

Documentación, nombres de applets/verbos (`articulo`, `buscar`, `calcular`) y claves JSON (`fuente`, `fecha_consulta`) van en español. Identificadores Go siguen la convención habitual del lenguaje.
