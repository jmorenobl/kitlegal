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
→ specify (args: modo desatendido, alcance = sección, no especificado = fuera de alcance)
→ clarify (args: se autorresponde con el criterio; [ESCALAR] si toca alcance/frontera/privacidad/TOS/decisión cerrada)
→ gate_spec (prompt juez → gates/spec.json) → check_gate_spec (shell: veredicto, marcadores, checklists)
→ [supervisado] gate humano
→ plan (Constitution Check estricto) → gate_plan → check_gate_plan → [supervisado] gate humano
→ tasks → analyze (informe a gates/analyze.md) → gate_tasks → check_gate_tasks
→ implement → make ci → si falla: do-while (reparar + make ci) ×3
→ converge → implement_restante
→ revision (diff completo contra DoD → gates/revision.md/.json) → ci_final
→ [supervisado] gate humano final
```

La fusión a `main` (PR + squash-merge) y el release son siempre acciones humanas. El workflow deja la rama del hito commiteada y en verde; nunca hace `push` ni `merge`.

## Cómo funcionan los gates desatendidos

1. **Los comandos no preguntan.** Cada `command` recibe en `args` la instrucción "MODO DESATENDIDO" con la regla de decisión. En `clarify`, Claude genera sus preguntas y las responde él mismo; deja rastro en `## Clarifications` con `(auto: <criterio>)`.
2. **Un juez por fase.** `gate_spec`, `gate_plan`, `gate_tasks` y `revision` son pasos `prompt`: leen los artefactos, corrigen lo corregible y escriben `gates/<fase>.json` con `{"veredicto","motivos","correcciones"}`.
3. **Un `shell` determinista decide.** `check_gate_*` lee el JSON con `jq` y además comprueba invariantes mecánicos (sin `[NEEDS CLARIFICATION]`, sin `[ESCALAR]`, checklists completas). Si falla, el run queda en estado `failed`.
4. **Escalado = parada.** Cuando el criterio dice "no decidas" (alcance, frontera humana, privacidad, TOS, decisión cerrada), el juez responde `rechazado` y el workflow se detiene. Se corrige a mano el artefacto y se reanuda desde ese paso.
5. **Modo supervisado.** `modo=supervisado` añade gates humanos tras spec, plan y revisión final; se resuelven con `resume … veredicto_spec=approve`, etc.

## Uso

```bash
scripts/hito.sh H0                     # desatendido
scripts/hito.sh H0 supervisado         # con pausas humanas
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
- `inputs.hito` se interpola en un `shell`; está restringido por `enum`. No añadir inputs libres a pasos `shell`.
- Si `speckit init` se actualiza (`specify integration upgrade`), regenera `.claude/skills/speckit-*`; el workflow y la constitución no se tocan.
