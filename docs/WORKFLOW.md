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
→ gate_spec (juez → gates/spec.json) → check_gate_spec (shell) → [supervisado] gate humano
→ plan (Constitution Check estricto) → gate_plan → check_gate_plan → [supervisado] gate humano
→ tasks (cada tarea = rebanada vertical que deja `make ci` en verde) → analyze → gate_tasks → check_gate_tasks
→ bucle por tarea (do-while):
     siguiente_tarea (shell: primera "- [ ] Tnnn" de tasks.md, cuenta intentos)
     → implementar_tarea (command implement, solo esa tarea)
     → verificar (shell: make ci | go build+vet+test; log en gates/ci.log)
     → si falla: reparar (prompt con las últimas 80 líneas del log) → verificar_reparacion (si sigue rojo, parada)
→ ci (make ci del hito) → si falla: do-while (reparar_hito + ci_reintento) ×3 → ci_tras_reparacion
→ converge → implement_restante
→ revision (diff completo contra DoD → gates/revision.md/.json) → ci_final
→ [supervisado] gate humano final
```

La fusión a `main` (PR + squash-merge) y el release son siempre acciones humanas. El workflow deja la rama del hito commiteada y en verde; nunca hace `push` ni `merge`.

## Clarificación sin sesgo

`clarify` se divide en tres pasos para que quien formula las preguntas no sea quien las responde:

1. **`clarify_preguntas`** ejecuta `/speckit-clarify` con la orden de solo escribir las preguntas y sus opciones en `gates/clarify-preguntas.json`, sin recomendación ni opción preferida, y sin tocar el spec.
2. **`resolver_clarify`** es un `prompt` que spec-kit lanza como un proceso `claude -p` nuevo: no comparte conversación con el paso anterior y se le dice que decida solo con documentos (hito, `CLAUDE.md`, `refs/`, constitución). Aplica el criterio en este orden: (a) lo determinan las fuentes → esa es la respuesta; (b) no está especificado → "fuera de alcance: no se implementa"; (c) varias opciones válidas → la de mayor calidad y mejores prácticas, con la alternativa rechazada; (d) alcance, frontera humana, privacidad, TOS o decisión cerrada → `escalar: true`.
3. **`check_clarify`** detiene el run si alguna respuesta está escalada. Se edita `gates/clarify-respuestas.json` a mano y se reanuda. **`clarify_integrar`** vuelve a llamar a `/speckit-clarify` solo para integrar los pares Q/A en `spec.md` con marca `(auto: criterio, fuente)`.

Los jueces de cada gate y el revisor final son también procesos nuevos, por la misma razón.

## Implementación tarea a tarea

Con `granularidad=tarea` (valor por defecto) el bucle toma la primera línea `- [ ] Tnnn` de `tasks.md`, lanza `/speckit-implement` restringido a esa tarea y ejecuta la batería determinista. Si falla:

- el paso `reparar` recibe **las últimas 80 líneas de la salida** interpoladas en el prompt y la ruta del log completo (`gates/ci.log`), con la orden de arreglar la causa y no el control;
- `verificar_reparacion` vuelve a ejecutar la batería; si sigue en rojo, el run se detiene para revisión humana;
- una tarea que no queda marcada `[X]` tras 3 intentos detiene el run (`gates/tareas-intentos.json`, `gates/tarea-Tnnn.md`).

Para que la verificación por tarea tenga sentido, `tasks` recibe la regla de que cada tarea es una rebanada vertical (test + implementación) que deja `make ci` en verde por sí sola; el gate de tareas lo comprueba. Mientras no exista `Makefile` con objetivo `ci`, la batería es `go build ./... && go vet ./... && go test -race ./...`; sin `go.mod` no hay nada que verificar (primeras tareas de H0).

`granularidad=hito` conserva la pasada única de `implement` (más barata, menos control).

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
- Un `shell` que falla detiene el run salvo `continue_on_error: true`; solo lo llevan los pasos cuyo fallo se enruta a una reparación (`verificar`, `ci`, `ci_reintento`). Los `check_*` fallan a propósito para parar.
- La batería por tarea ejecuta `make ci` completo tras cada tarea; en hitos grandes es lento pero determinista. Si hace falta, añadir un objetivo `make check` más rápido y usarlo en `verificar`.
- `inputs.hito` se interpola en un `shell`; está restringido por `enum`. No añadir inputs libres a pasos `shell`.
- Si `speckit init` se actualiza (`specify integration upgrade`), regenera `.claude/skills/speckit-*`; el workflow y la constitución no se tocan.
