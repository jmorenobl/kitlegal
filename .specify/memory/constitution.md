# Constitución de kitlegal

Documento normativo que gobierna toda especificación, plan, tarea e implementación de `kitlegal`. Complementa, no sustituye, a `CLAUDE.md` (decisiones cerradas) y a `docs/ROADMAP.md` (Definition of Done, arquitectura objetivo y controles). Ante conflicto, prevalece el orden: esta constitución → `CLAUDE.md` → `docs/ROADMAP.md` → `refs/`.

## Principios

### I. Fuentes públicas y frontera humana (INNEGOCIABLE)
Solo se automatizan fuentes públicas. Presentar escritos, notificaciones y cualquier acción con identidad terminan en "fichero listo para firmar" y exit 6; nunca un POST a una sede. CENDOJ: solo resolución de ECLI y metadatos. Toda petición HTTP pasa por `internal/httpx` (User-Agent identificable, `robots.txt`, rate limit por host, sin paralelismo agresivo). No existe ninguna llamada HTTP con método distinto de GET/HEAD en el módulo.

### II. Nada sin cita ni fuente
Todo applet emite el sobre `{ok, fuente, url, fecha_consulta, hash, data}`. Sin `fuente`+`url`+`fecha_consulta`+`hash` no hay cita. Un dato sin `source` no entra en el grafo. Nunca se inventa contenido legal; se distingue ley de reglamento y se señala variación autonómica.

### III. Tests primero y offline
Cada hito empieza por el test e2e (`testscript`) que describe la entrega. Tests unitarios offline contra fixtures en `testdata/<fuente>/`; solo `scripts/verify-sources.sh` toca la red. Toda salida de applet se valida contra `schemas/*.json` en test. Cobertura mínima: `internal/core/**` ≥ 85 %, global ≥ 70 %.

### IV. Arquitectura hexagonal con reglas ejecutables
Dominio puro en `internal/core` (define los puertos); adaptadores en `internal/source/<fuente>`, `httpx`, `cache`, `store`, `graph`, `render`; composición manual en `internal/app`. Reglas de dependencia (vigiladas por `depguard`/`forbidigo` y `internal/arch_test.go`): `core` no importa adaptadores; solo `httpx` importa `net/http`; solo `cache|store|graph` importan SQLite; solo `cli` y `cmd/` llaman a `os.Exit`; solo `render` escribe en stdout; logs con `log/slog` a stderr. Errores tipados que mapean a exit codes estables (0 ok · 2 args · 3 no encontrado · 4 fuente no disponible · 5 rate-limited/TOS · 6 requiere identidad humana). Ningún `panic` en rutas de usuario.

### V. Simplicidad y dependencias fijadas (YAGNI)
Lo que no pide un hito no se construye. Sin framework de DI, ORM ni generador de CLI. Dependencias permitidas: `alecthomas/kong`, `modernc.org/sqlite`, `stretchr/testify`, `rogpeppe/go-internal`, `golang.org/x/time/rate`, `temoto/robotstxt`, `gopkg.in/yaml.v3`, `invopop/jsonschema`, `santhosh-tekuri/jsonschema` y, en fase 4, `modelcontextprotocol/go-sdk`. Cualquier otra requiere justificación explícita en el plan (sección Complexity Tracking) y en la PR. `internal/cli` no se extrae a librería.

### VI. Un binario, convenciones de agente
Nombre único `kitlegal`; multicall por `os.Args[0]` o primer argumento; flags globales `--json --timeout --offline --dry-run --describe --no-graph --asunto --verbose`. `--describe` emite JSON Schema, del que se generan tools MCP y tablas de comandos. Skills sin código: `SKILL.md` < 300 líneas, `references/` generadas desde `data/*.yaml` (única fuente de verdad), `scripts/` como symlinks al binario.

### VII. Grafo y privacidad
Ids naturales (ELI, ECLI, DIR3, NIF de entidad, INE) en `world.db`; el grafo del asunto (`.kitlegal/case.db`) solo guarda ids del mundo y nunca se sincroniza ni exporta por defecto. `Persona` nunca lleva NIF/DNI (rechazo por regex en `graph.Apply`).

## Restricciones técnicas

- Go estable actual, módulo `github.com/jmorenobl/kitlegal`, `CGO_ENABLED=0`, `-trimpath`.
- Idioma: documentación, verbos de applet y claves JSON en español; identificadores Go según convención del lenguaje.
- Formato y lint: `gofumpt` + `goimports`, `.golangci.yml` estricto (lista en `docs/ROADMAP.md` §3). `go test -race`. `govulncheck`, `gosec`, `gitleaks`, `go mod tidy -diff` en CI.
- Conventional Commits, SemVer, `CHANGELOG.md` (sección *Unreleased*), ADR en `docs/ADR/` (MADR corto) cuando cambia una decisión de arquitectura, fila en `docs/SOURCES.md` cuando se toca una fuente externa.

## Flujo de trabajo y gates

1. Un hito = una rama corta desde `main` + una PR + squash-merge con CI verde. Un hito no se cierra sin la Definition of Done completa (`docs/ROADMAP.md` §1).
2. Orden de implementación: `core` → adaptador → applet → skill.
3. Gates automáticos: Constitution Check en el plan (violaciones no justificadas = ERROR), checklists completas antes de implementar, `make ci` verde antes de revisión, revisión de código y seguridad antes de la PR.
4. La fusión a `main` y el release son siempre acciones humanas.

## Criterio de decisión autónoma

Se aplica en `clarify`, en los gates automáticos del workflow `hito` y en cualquier punto en que un agente deba elegir sin intervención humana:

1. **Siempre la mejor solución.** Entre alternativas, elegir la que mejor cumpla calidad, escalabilidad, mantenibilidad, seguridad y buenas prácticas del ecosistema Go y de este proyecto. Ni atajos ni ñapas: nada de `//nolint` sin justificación, nada de tests desactivados, nada de "TODO: arreglar luego", nada de capturar errores para silenciarlos.
2. **Lo no especificado no se implementa.** Si una funcionalidad, comportamiento o parámetro no está definido en el hito del roadmap, en `CLAUDE.md`, en `refs/` o en esta constitución, se deja fuera de alcance y se anota como tal (sección *Fuera de alcance* del spec o *Pendientes* de la PR). No se inventan requisitos ni se anticipan fases futuras.
3. **Cuando hay que elegir, elegir con criterio y dejar rastro.** Cada decisión automática se registra en el artefacto correspondiente (`spec.md` → `## Clarifications` con prefijo `(auto)`, `plan.md` → *Complexity Tracking* o *Decisiones*, `gates/*.json` → `motivos`) con la alternativa rechazada y por qué.
4. **Escalar en lugar de adivinar.** Si la duda afecta a alcance, a la frontera humana, a privacidad, a términos de uso de una fuente o a una decisión ya cerrada, el agente no decide: marca el gate como `rechazado` con el motivo y el workflow se detiene para que un humano resuelva.
5. **Verificar antes de aprobar.** Un gate solo se marca `aprobado` tras comprobar el artefacto contra el hito, los principios anteriores y la Definition of Done; corregir el artefacto es preferible a rechazarlo, salvo en el caso 4.

## Gobernanza

Esta constitución prevalece sobre cualquier otra práctica. Toda PR se revisa contra ella. Enmiendas: cambio en este fichero + ADR que lo motive + actualización de `CLAUDE.md` si afecta a una decisión cerrada. Las decisiones listadas como cerradas en `CLAUDE.md` no se reabren en spec, plan ni clarify.

**Version**: 1.0.0 | **Ratified**: 2026-09-09 | **Last Amended**: 2026-09-09
