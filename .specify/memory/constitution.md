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

## Gates: mecánico antes que juez, juez antes que humano

Tres capas, y cada comprobación vive en la capa más baja que pueda verificarla. Lo que se puede comprobar con un script nunca se delega a un modelo; lo que exige criterio se juzga con rúbrica cerrada; lo que compromete responsabilidad legal lo decide siempre una persona.

**Capa 1 · Mecánica (sin LLM, forma parte de `make ci` y del workflow):** `gofumpt`/`goimports`, `golangci-lint`, `go vet`, `go test -race`, `govulncheck`, `gitleaks`, `go mod tidy -diff`; test de arquitectura sobre las reglas de dependencia; validación de toda salida de applet contra `schemas/*.json`; exit codes con fixtures que fuerzan 2, 3, 4, 5 y 6; tests offline contra `testdata/` grabado; **corrección de citas contra la respuesta grabada del BOE** (la verdad terreno existe; nunca la sustituye la opinión de un modelo); frontmatter válido en cada `SKILL.md` y `references/` idénticas a lo generado desde `data/*.yaml` (diff vacío); en el workflow, guardián de diff por tarea (rutas declaradas, `testdata/` y `schemas/` protegidos), ausencia de marcadores pendientes y checklists completas.

**Capa 2 · Juez LLM (rúbrica cerrada):** protocolo de razonamiento de cada `SKILL.md`; que la `description` de una skill active con las preguntas que debe y no con las que no debe (evals); claridad de las explicaciones de `graph check` y `anomalies`; que un adaptador nuevo respeta las reglas de fuentes que el linter no ve (rate limit, User-Agent, TOS); coherencia spec ↔ plan ↔ tasks. Reglas: criterios fijos con `cumple`/`no cumple`, evidencia citada y umbral (todos cumplen); el juez no corrige, solo emite veredicto; las correcciones las aplica un proceso distinto y se vuelve a juzgar, con un tope de dos rondas; cada juez es un proceso nuevo sin acceso al razonamiento anterior; el gate que bloquea el merge lo emiten dos jueces con prompts distintos (uno de ellos adversarial) y el desacuerdo escala a humano.

**Capa 3 · Humano (nunca se automatiza):** cualquier cambio en `docs/SOURCES.md` (qué fuentes se tocan y cómo: responsabilidad legal); cualquier cambio en `data/anomalias/` (una regla mal calibrada produce acusaciones implícitas a escala); un adaptador de fuente nuevo bajo `internal/source/`; toda modificación de ficheros existentes en `testdata/` y `schemas/` y toda grabación de fixtures (solo en tareas etiquetadas `[datos]`, con pausa obligatoria); la fusión a `main` y el release.

**Reglas del modo desatendido:** el ejecutor arregla el código, nunca el test ni el fixture; `KITLEGAL_RECORD=1` solo se ejecuta en un job separado y revisado, jamás dentro del bucle de implementación; el diff de cada tarea se limita a las rutas que la tarea declara (más `go.mod`, `go.sum`, `CHANGELOG.md` y el directorio del feature); una tarea sin rutas declaradas no se ejecuta.

## Criterio de decisión autónoma

Se aplica en `clarify`, en los gates automáticos del workflow `hito` y en cualquier punto en que un agente deba elegir sin intervención humana:

1. **Siempre la mejor solución.** Entre alternativas, elegir la que mejor cumpla calidad, escalabilidad, mantenibilidad, seguridad y buenas prácticas del ecosistema Go y de este proyecto. Ni atajos ni ñapas: nada de `//nolint` sin justificación, nada de tests desactivados, nada de "TODO: arreglar luego", nada de capturar errores para silenciarlos.
2. **Lo no especificado no se implementa.** Si una funcionalidad, comportamiento o parámetro no está definido en el hito del roadmap, en `CLAUDE.md`, en `refs/` o en esta constitución, se deja fuera de alcance y se anota como tal (sección *Fuera de alcance* del spec o *Pendientes* de la PR). No se inventan requisitos ni se anticipan fases futuras.
3. **Cuando hay que elegir, elegir con criterio y dejar rastro.** Cada decisión automática se registra en el artefacto correspondiente (`spec.md` → `## Clarifications` con prefijo `(auto)`, `plan.md` → *Complexity Tracking* o *Decisiones*, `gates/*.json` → `motivos`) con la alternativa rechazada y por qué.
4. **Escalar en lugar de adivinar.** Si la duda afecta a alcance, a la frontera humana, a privacidad, a términos de uso de una fuente, a las reglas de anomalías o a una decisión ya cerrada, el agente no decide: marca el gate como `rechazado` con el motivo y el workflow se detiene para que un humano resuelva.
5. **Verificar antes de aprobar.** Un gate solo se marca `aprobado` tras comprobar el artefacto contra el hito, los principios anteriores y la Definition of Done; corregir el artefacto es preferible a rechazarlo, salvo en el caso 4.

## Gobernanza

Esta constitución prevalece sobre cualquier otra práctica. Toda PR se revisa contra ella. Enmiendas: cambio en este fichero + ADR que lo motive + actualización de `CLAUDE.md` si afecta a una decisión cerrada. Las decisiones listadas como cerradas en `CLAUDE.md` no se reabren en spec, plan ni clarify.

**Version**: 1.1.0 | **Ratified**: 2026-09-09 | **Last Amended**: 2026-09-09
