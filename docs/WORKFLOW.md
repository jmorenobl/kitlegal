# Workflow por hito con spec-kit

Cada hito de `ROADMAP.md` se implementa con una pasada del workflow `hito` de spec-kit (`.specify/workflows/hito/workflow.yml`). El workflow encadena los comandos `speckit.*` en modo headless (`claude -p`) e intercala **gates automáticos** que evalúa Claude con la sección «Criterio de decisión autónoma» de `.specify/memory/constitution.md`.

## Piezas

| Pieza | Dónde | Para qué |
|---|---|---|
| Constitución | `.specify/memory/constitution.md` | Principios, restricciones, DoD y el criterio con el que se decide sin humano |
| Workflow | `.specify/workflows/hito/workflow.yml` | Secuencia, gates, bucle de reparación de CI |
| Lanzador | `scripts/hito.sh` | Comprueba `main` limpio, exporta flags de Claude y lanza/reanuda |
| Extensión git | `.specify/extensions/git/` | Rama `NNN-hN-slug` por hito y auto-commit (Conventional Commits) tras cada fase |
| Skills | `.claude/skills/speckit-*` | Los comandos `/speckit-*` (generados por `specify init`, no editar a mano) |
| Artefactos | `specs/NNN-hN-slug/` | `spec.md`, `plan.md`, `research.md`, `tasks.md`, `checklists/`, `gates/*.json` |

## Secuencia

```
extraer_hito (shell: sección del hito en ROADMAP.md → JSON)
→ specify (modo desatendido: alcance = sección; ambigüedades como [NEEDS CLARIFICATION], sin resolver)
→ clarify_preguntas (command: solo formula preguntas, sin recomendación → gates/clarify-preguntas.json)
→ resolver_clarify (prompt en PROCESO NUEVO, contexto limpio → gates/clarify-respuestas.json)
→ check_clarify (shell: todas respondidas; ninguna escalada) → clarify_integrar (command: integra respuestas en spec.md)
→ precheck_spec (shell) → ronda_spec ×2 [juez_spec → leer_gate_spec → corrector_spec si rechazado y corregible] → check_gate_spec
→ [supervisado] gate humano
→ plan → precheck_plan → ronda_plan ×2 [juez → leer → corrector] → check_gate_plan → [supervisado] gate humano
→ tasks → analyze → precheck_tasks (formato, ids, rutas declaradas, [datos]) → ronda_tasks ×2 → check_gate_tasks
→ bucle por tarea (do-while):
     siguiente_tarea (shell: primera "- [ ] Tnnn", intentos, base git, rutas declaradas, [datos] → gates/tarea-actual.json)
     → implementar_tarea (command implement, solo esa tarea)
     → guardian_diff (shell: rutas declaradas; testdata/ y schemas/ solo con [datos])
     → verificar (shell: make ci | go build+vet+test; log en gates/ci.log)
     → si falla: reparar (prompt con las últimas 80 líneas del log) → guardian_diff_reparacion → verificar_reparacion
     → si la tarea es [datos]: gate humano (independiente de `modo`)
→ ci (make ci del hito) → si falla: do-while (reparar_hito + ci_reintento) ×3 → ci_tras_reparacion
→ converge → implement_restante
→ ronda_revision ×2 [juez A (DoD) + juez B (adversarial) → leer_revision → corrector si ambos rechazan y es corregible]
→ ci_final (make ci; ambos jueces aprobado; sin tareas pendientes; árbol limpio)
→ rutas_sensibles → gate humano forzado si el diff toca docs/SOURCES.md, data/anomalias/ o una fuente nueva
→ [supervisado] gate humano final
```

La fusión a `main` (PR + squash-merge) y el release son siempre acciones humanas. El workflow deja la rama del hito commiteada y en verde; nunca hace `push` ni `merge`.

## Tres capas de gate

Sigue la sección «Gates» de la constitución: cada comprobación vive en la capa más baja que pueda verificarla.

1. **Mecánica (shell, sin LLM).** `precheck_*` corre antes de cualquier juez: marcadores pendientes, checklists, sección "Fuera de alcance", formato e ids de tareas, rutas declaradas, etiqueta `[datos]`. `verificar`/`ci` ejecutan `make ci`, que debe encadenar lint, `-race`, schema-check, drift de `references/`, test de arquitectura y golden files de citas. `leer_gate_*` valida el JSON del juez y su coherencia (aprobado ⇔ todos los criterios cumplen).
2. **Juez LLM con rúbrica.** `juez_spec`, `juez_plan`, `juez_tasks`, `revision_juez_a` y `revision_juez_b` reciben criterios fijos (`a`…`i`) y escriben `{"veredicto","corregible","criterios":[{id,criterio,cumple,evidencia}],"motivos"}`. **No corrigen nada.** Si rechazan y el motivo es corregible, un `corrector_*` (proceso distinto, modelo de redacción o implementación) aplica los motivos y se vuelve a juzgar, con tope de dos rondas. Si el motivo requiere decisión humana (`corregible: false`), el `check_gate_*` para el run.
3. **Humano.** Pausas que no dependen de `modo` ni admiten pre-aprobación por input: tras cada tarea `[datos]` (tocó `testdata/` o `schemas/`) y al final si el diff toca `docs/SOURCES.md`, `data/anomalias/` o crea un directorio nuevo bajo `internal/source/`. Se reanudan con `specify workflow resume <run_id>` desde un terminal, que pregunta approve/reject.

## Clarificación sin sesgo

`clarify` se divide en tres pasos para que quien formula las preguntas no sea quien las responde:

1. **`clarify_preguntas`** ejecuta `/speckit-clarify` con la orden de solo escribir las preguntas y sus opciones en `gates/clarify-preguntas.json`, sin recomendación ni opción preferida, y sin tocar el spec.
2. **`resolver_clarify`** es un `prompt` que spec-kit lanza como un proceso `claude -p` nuevo: no comparte conversación con el paso anterior y se le dice que decida solo con documentos (hito, `CLAUDE.md`, `refs/`, constitución). Aplica el criterio en este orden: (a) lo determinan las fuentes → esa es la respuesta; (b) no está especificado → "fuera de alcance: no se implementa"; (c) varias opciones válidas → la de mayor calidad y mejores prácticas, con la alternativa rechazada; (d) alcance, frontera humana, privacidad, TOS, anomalías o decisión cerrada → `escalar: true`.
3. **`check_clarify`** detiene el run si alguna respuesta está escalada. Se edita `gates/clarify-respuestas.json` a mano y se reanuda. **`clarify_integrar`** vuelve a llamar a `/speckit-clarify` solo para integrar los pares Q/A en `spec.md` con marca `(auto: criterio, fuente)`.

## Implementación tarea a tarea y guardián de diff

Con `granularidad=tarea` (valor por defecto) el bucle toma la primera línea `- [ ] Tnnn` de `tasks.md`, anota el commit base y las rutas que la tarea declara, lanza `/speckit-implement` restringido a esa tarea y después:

- **`guardian_diff`** compara `git diff --name-only <base>` más los ficheros nuevos con las rutas declaradas. Falla, y el run se para, si un fichero queda fuera (scope creep) o si toca `testdata/` o `schemas/` sin etiqueta `[datos]` (arreglar el test en vez del código, grabar fixtures en el bucle). Siempre permitidos: `go.mod`, `go.sum`, `CHANGELOG.md`, el directorio del feature, y `x_test.go` cuando se declara `x.go`. Declarar un directorio (`internal/cli/`) permite todo lo que cuelga de él.
- **`verificar`** ejecuta la batería determinista. Si falla, `reparar` recibe **las últimas 80 líneas de la salida** en el prompt y la ruta del log completo, con la orden de arreglar la causa y no el control; después vuelven a pasar el guardián y la verificación. Si sigue en rojo, el run se detiene.
- Una tarea que no queda marcada `[X]` tras 3 intentos detiene el run (`gates/tareas-intentos.json`, `gates/tarea-Tnnn.md`).
- Una tarea `[datos]` termina siempre en un gate humano.

Para que la verificación por tarea tenga sentido, `tasks` recibe la regla de que cada tarea es una rebanada vertical (test + implementación) que deja `make ci` en verde por sí sola, con todas sus rutas declaradas, y `precheck_tasks` + `juez_tasks` lo comprueban. Mientras no exista `Makefile` con objetivo `ci`, la batería es `go build ./... && go vet ./... && go test -race ./...`; sin `go.mod` no hay nada que verificar (primeras tareas de H0).

`granularidad=hito` conserva la pasada única de `implement` sin guardián por tarea (más barata, menos control).

## Revisión final con dos jueces

`revision_juez_a` evalúa la Definition of Done con rúbrica; `revision_juez_b` parte de la hipótesis contraria y busca evidencia de atajos, fixtures retocados, tests vacíos, alcance excedido y violaciones que el linter no ve. Ninguno modifica ficheros. `leer_revision` combina: ambos aprueban → sigue; ambos rechazan y es corregible → `corrector_revision` y nueva ronda (máximo dos); cualquier desacuerdo o motivo no corregible → `ci_final` para el run para un humano.

## Modelo por paso

Cada paso `command` y `prompt` lleva `model: "{{ inputs.modelo_<rol> }}"`, que spec-kit traduce en `claude -p … --model <valor>`. Los valores admitidos son los alias `fable`, `opus`, `sonnet`, `haiku` o el nombre completo del modelo.

| Input | Pasos | Por defecto | Razón |
|---|---|---|---|
| `modelo_juez` | `resolver_clarify`, `juez_spec`, `juez_plan`, `juez_tasks`, `revision_juez_a`, `revision_juez_b` | `fable` | Decisiones con el criterio de la constitución; es donde un error cuesta más |
| `modelo_redaccion` | `specify`, `clarify_preguntas`, `clarify_integrar`, `plan`, `tasks`, `corrector_spec`, `corrector_plan`, `corrector_tasks` | `opus` | Artefactos largos con muchas reglas que respetar |
| `modelo_implementacion` | `implementar_tarea`, `implement`, `implement_restante`, `reparar`, `reparar_hito`, `corrector_revision` | `opus` | Código y depuración |
| `modelo_analisis` | `analyze`, `converge` | `sonnet` | Lectura y contraste de artefactos; barato y suficiente |

Sobrescritura por run:

```bash
specify workflow run hito -i hito=H0 -i modelo_implementacion=sonnet -i modelo_analisis=haiku
KITLEGAL_MODELO_IMPLEMENTACION=sonnet KITLEGAL_MODELO_ANALISIS=haiku scripts/hito.sh H0
```

Los pasos `shell` no usan modelo. Para fijar un modelo distinto en un solo paso sin tocar los inputs, edita su `model:` en el YAML o usa un overlay (`specify workflow overlay add …`).

## Uso

```bash
scripts/hito.sh H0                     # desatendido, tarea a tarea
scripts/hito.sh H0 supervisado         # con pausas humanas
specify workflow run hito -i hito=H0 -i granularidad=hito   # implement en una pasada
specify workflow status                # runs y estado
specify workflow status <run_id>
scripts/hito.sh --resume <run_id>      # tras corregir a mano un artefacto
scripts/hito.sh --resume <run_id> veredicto_plan=approve
specify workflow resolve hito          # ver el workflow compuesto con overlays
```

Estado de cada run en `.specify/workflows/runs/<run_id>/` (`state.json`, `inputs.json`, `log.jsonl`).

## Permisos de Claude en modo headless

`claude -p` respeta `.claude/settings.json`. Para que `implement` pueda compilar y testear sin prompts, el proyecto necesita una lista de permisos como esta (añadir a `.claude/settings.json`; `deny` gana a `allow`):

```json
{
  "permissions": {
    "allow": [
      "Read", "Edit", "Write", "Glob", "Grep",
      "Bash(rtk:*)", "Bash(go:*)", "Bash(gofmt:*)", "Bash(gofumpt:*)", "Bash(goimports:*)",
      "Bash(golangci-lint:*)", "Bash(govulncheck:*)", "Bash(goreleaser:*)", "Bash(lefthook:*)",
      "Bash(gitleaks:*)", "Bash(make:*)", "Bash(git:*)", "Bash(jq:*)",
      "Bash(ls:*)", "Bash(cat:*)", "Bash(head:*)", "Bash(tail:*)", "Bash(wc:*)", "Bash(find:*)",
      "Bash(grep:*)", "Bash(rg:*)", "Bash(sed:*)", "Bash(awk:*)", "Bash(mkdir:*)", "Bash(ln:*)",
      "Bash(chmod:*)", "Bash(cp:*)", "Bash(mv:*)", "Bash(touch:*)", "Bash(date:*)",
      "Bash(specify:*)", "Bash(.specify/scripts/bash/*)", "Bash(.specify/extensions/git/scripts/bash/*)",
      "Bash(./kitlegal:*)", "Bash(./bin/kitlegal:*)"
    ],
    "deny": [
      "Bash(git push:*)", "Bash(git merge:*)", "Bash(git reset --hard:*)",
      "Bash(rm -rf:*)", "Bash(curl:*)", "Bash(wget:*)", "Bash(sudo:*)"
    ]
  }
}
```

`scripts/hito.sh` exporta `SPECKIT_INTEGRATION_CLAUDE_EXTRA_ARGS="--permission-mode acceptEdits"`. En un entorno aislado (contenedor o VM) puede sustituirse por `--dangerously-skip-permissions`.

## Limitaciones conocidas

- Los pasos `command` transmiten la salida al terminal y no la capturan; por eso los jueces escriben ficheros en `gates/` y `analyze` recibe la instrucción de guardar su informe.
- Los pasos `prompt` y `shell` tienen `timeout` explícito (1800 s); `make ci` debe caber en ese margen.
- Un `shell` que falla detiene el run salvo `continue_on_error: true`; solo lo llevan los pasos cuyo fallo se enruta a una reparación (`verificar`, `ci`, `ci_reintento`). Los `precheck_*`, `check_*`, `guardian_diff*` y `leer_*` fallan a propósito para parar.
- En las rondas juez → corrector, la segunda corrección no vuelve a juzgarse: `check_gate_*` lee el último veredicto y, si sigue rechazado, para. Es el tope de dos rondas de la constitución.
- El guardián de diff extrae rutas de la línea de la tarea (tokens con `/` o con extensión conocida, `Makefile`, `LICENSE`). Una tarea que toque muchos ficheros debe declarar directorios.
- La batería por tarea ejecuta `make ci` completo tras cada tarea; en hitos grandes es lento pero determinista. Si hace falta, añadir un objetivo `make check` más rápido y usarlo en `verificar`.
- `inputs.hito` se interpola en un `shell`; está restringido por `enum`. No añadir inputs libres a pasos `shell`.
- Si `speckit init` se actualiza (`specify integration upgrade`), regenera `.claude/skills/speckit-*`; el workflow y la constitución no se tocan.
