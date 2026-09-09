# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Estado del repositorio

**No hay código todavía.** El repo contiene únicamente documentos de diseño (en español) para `kitlegal`, un binario Go multicall + skills agénticas para consultar fuentes legales públicas españolas (BOE, PLACSP, BDNS, BORME, EUR-Lex…). No existe `go.mod`, `Makefile`, tests ni CI.

Skills de agente instaladas (no son parte del producto): el conjunto `samber/cc-skills-golang` vive en `.agents/skills/` (registro en `skills-lock.json`); `.claude/skills/` son symlinks a ese directorio. Al escribir Go, `golang-how-to` orquesta el resto (`golang-cli`, `golang-project-layout`, `golang-testing`, `golang-database`…).

Antes de escribir código, lee los documentos semilla de `refs/` en este orden:

1. `refs/mapa-sistema-legal-skills.md` — sistema legal español, fuentes con semáforo de automatizabilidad (🟢 API / 🟡 scraping / 🔴 requiere identidad humana) y catálogo de skills.
2. `refs/kitlegal-estructura-y-ecosistema.md` — estructura del monorepo, convenciones del multicall, packs por vertical, distribución (skills, plugin, MCP, librería Go).
3. `refs/kitlegal-grafo.md` — grafo legal sobre ELI en SQLite (mundo público + asunto privado), alimentación por uso, `graph check` y reglas de anomalías.

`refs/00-README.md` resume el estado, el primer hito y las decisiones cerradas.

## Primer hito

`kitlegal boe articulo BOE-A-2015-10565 a21` funcionando en Go con caché SQLite, y la skill `boe-fiscal` (hoy Python `scripts/boe.py`, fuera de este repo) migrada para invocar el binario con el mismo comportamiento. Todo lo demás se construye encima. `boe.py` es el patrón a portar a `internal/source/boe`.

## Comandos (planificados, aún no existen)

Cuando se cree el proyecto Go, el `Makefile` debe exponer `build`, `test`, `lint`, `skills-sync` y `release`. Tooling previsto:

- Módulo `github.com/jmorenobl/kitlegal`, CLI con Kong (`github.com/alecthomas/kong`), SQLite sin cgo (`modernc.org/sqlite`), release con goreleaser.
- Tests unitarios offline contra fixtures grabados en `testdata/<fuente>/`; grabar con `KITLEGAL_RECORD=1`. Solo `scripts/verify-sources.sh` (CI nightly) toca la red.
- `scripts/skills-sync.sh` regenera `skills/*/references/*.md` desde `data/*.yaml` y los symlinks de `scripts/`.

## Decisiones ya tomadas (no reabrir)

- **Nombre único** `kitlegal` para repo, módulo, binario y directorios (`.kitlegal/` en cwd para el asunto, `~/.cache/kitlegal/` para caché y grafo del mundo).
- **Multicall**: un solo ejecutable; `os.Args[0]` o el primer argumento elige el applet (`boe`, `placsp`, `bdns`, `cita`, `plazos`, `graph`…). Un symlink `boe -> kitlegal` permite que las skills sigan llamando `scripts/boe articulo …`.
- **Convenciones de agente** en `internal/cli`: flags globales `--json`, `--timeout`, `--offline`, `--dry-run`, `--describe` (emite JSON Schema de entrada/salida, del que se generan las tools MCP y la tabla de comandos de cada SKILL.md), `--no-graph`, `--asunto`. No se extrae `internal/cli` a librería externa hasta que tres applets repitan el patrón.
- **Exit codes estables**: 0 ok · 2 args · 3 no encontrado · 4 fuente no disponible · 5 rate-limited/TOS · 6 requiere identidad humana.
- **Sobre de salida obligatorio** en todo applet: `{ok, fuente, url, fecha_consulta, hash, data}`. Sin `fuente`+`url`+`fecha_consulta`+`hash` no hay cita.
- **Skills sin código**: cada `skills/<nombre>/` lleva `SKILL.md` (< 300 líneas, protocolo de razonamiento + tabla de comandos + reglas), `references/` **generadas** desde `data/*.yaml` (cabecera `<!-- generado desde data/normas.yaml, no editar -->`) y `scripts/` como symlinks al binario. `data/*.yaml` es la única fuente de verdad.
- **Frontera humana**: solo se automatizan fuentes públicas. Presentar escritos, notificaciones y cualquier acción con identidad terminan en "fichero listo para firmar" y exit 6; nunca un POST a una sede.
- **CENDOJ**: solo resolución de ECLI y metadatos; nunca ingesta masiva ni scraping del buscador.
- **HTTP** únicamente vía `internal/httpx` (retries, rate limit por host, `robots.txt`, User-Agent identificable `kitlegal/x.y (+https://ventanillalegal.es/bot)`). Nada de paralelismo agresivo contra PETETE, DYCTEA o HJ del TC.
- **Grafo**: ids naturales (ELI, ECLI, DIR3, NIF, código INE) en `world.db`; el grafo del asunto (`.kitlegal/case.db`) solo guarda ids del mundo y nunca se sincroniza ni exporta por defecto. `Persona` nunca lleva NIF/DNI. Un dato sin `source` no entra.
- Reglas invariantes en todas las skills: nunca inventar contenido legal; cada afirmación con cita resuelta por `cita`; distinguir ley/reglamento; señalar variación autonómica.

## Arquitectura prevista (resumen)

```
cmd/kitlegal/main.go        multicall → internal/app (registro y dispatch de applets)
internal/cli                Kong + flags globales + exit codes
internal/core/{cita,plazos,competencia,schema,ids}
internal/source/<fuente>    un adaptador por fuente: Source{Name, Fetch(ctx,req), TTL, Terms}
internal/{httpx,cache,store,graph,analysis,render,docgen}
pkg/legalkit                API pública pequeña y estable
skills/  packs/  data/  schemas/  testdata/  evals/  mcp/  plugin/  docs/
```

Cada applet implementa además `Emit(ctx) []GraphOp`; `internal/graph` aplica las operaciones tras cada comando. Los packs (`packs/<vertical>/pack.yaml`) agrupan skills por vertical (`legal` base; `fiscal`, `laboral`, `mercantil`, `fiscalizador` extienden `legal`).

## Pendiente de verificar antes de fijar en código o `data/`

- Identificadores BOE de la tabla de leyes vertebrales (algunos están de memoria; comprobar con `kitlegal boe buscar`).
- Endpoints actuales de BDNS (swagger) y nombres de los feeds Atom de PLACSP.
- Términos de uso de PETETE (DGT), DYCTEA (TEAC) y del buscador HJ del TC.

## Convenciones de idioma

Documentación, nombres de applets/verbos (`articulo`, `buscar`, `calcular`) y claves JSON (`fuente`, `fecha_consulta`) van en español. Identificadores Go siguen la convención habitual del lenguaje.
