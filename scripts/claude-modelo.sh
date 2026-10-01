#!/usr/bin/env bash
# Ejecutable de Claude para los pasos del workflow `hito`
# (SPECKIT_INTEGRATION_CLAUDE_EXECUTABLE, lo exporta scripts/hito.sh).
#
# spec-kit solo pasa `--model <valor>` a `claude -p`. Este wrapper acepta
# `<modelo>@<esfuerzo>` y lo traduce a `--model <modelo> --effort <esfuerzo>`,
# de modo que cada rol del workflow fija modelo y esfuerzo con un único input:
#
#   --model fable@xhigh   →  --model fable --effort xhigh
#   --model opus          →  --model opus               (esfuerzo por defecto)
#
# Es además la única puerta por la que el workflow y scripts/paso.sh abren una
# sesión con modelo, así que aquí se pone lo que esa sesión no puede hacer
# (ADR 0032; al final del fichero).
#
# El valor llega a argv desde los inputs del workflow: se valida con una
# expresión estricta y cualquier otro formato aborta con exit 2. El prompt
# (argumento de -p) se copia sin interpretarlo.
set -euo pipefail

patron='^(fable|opus|sonnet|haiku|claude-[a-z0-9-]+)(@(low|medium|high|xhigh|max))?$'

args=()
while [ $# -gt 0 ]; do
  case "$1" in
    -p|--print)
      args+=("$1"); shift
      if [ $# -gt 0 ]; then args+=("$1"); shift; fi
      ;;
    --model)
      [ $# -ge 2 ] || { echo "claude-modelo: --model sin valor" >&2; exit 2; }
      if [[ ! "$2" =~ $patron ]]; then
        echo "claude-modelo: modelo inválido '$2' (formato: <modelo>[@low|medium|high|xhigh|max])" >&2
        exit 2
      fi
      args+=(--model "${BASH_REMATCH[1]}")
      if [ -n "${BASH_REMATCH[3]}" ]; then args+=(--effort "${BASH_REMATCH[3]}"); fi
      shift 2
      ;;
    *)
      args+=("$1"); shift
      ;;
  esac
done

# Las sesiones headless se reconocen en sus transcripts por el entrypoint
# (scripts/hito.sh «limite_api» y scripts/coste-run.sh). `claude -p` escribe
# `sdk-cli` salvo que CLAUDE_CODE_ENTRYPOINT venga del entorno, y la extensión de
# VS Code exporta `claude-vscode` a todo proceso hijo: un hito lanzado desde ella
# dejaba sesiones que ninguno de los dos scripts reconocía. Se fija aquí para que
# el transcript diga lo mismo se lance el hito desde donde se lance.
export CLAUDE_CODE_ENTRYPOINT=sdk-cli

# Lo que una sesión de un paso no puede hacer lo impide algo que no es su prompt
# (ADR 0032; docs/WORKFLOW.md «Lo que una sesión de un paso no puede hacer»). En H7.4
# `reparar_cierre` leyó el token de la suscripción de quien lanzó el run y dejó en
# segundo plano un sondeo con modelo. Todo se pone aquí, en la única puerta por la
# que el workflow (scripts/hito.sh) y scripts/paso.sh abren una sesión:
#   · la marca con la que el gancho PreToolUse le aplica scripts/workflow/politica-paso.sh;
#   · sin segundo plano: Claude Code no ofrece `run_in_background` ni pasa una orden
#     al fondo cuando vence su plazo;
#   · el `claude` de su PATH, y del de todo lo que ejecute, es el de
#     scripts/workflow/sin-modelo/, que no abre ninguna sesión;
#   · no lee .claude/settings.local.json: los permisos y los directorios adicionales
#     de la sesión interactiva de quien lanza el run no llegan a las del run.
raiz="$(cd "$(dirname "$0")/.." && pwd -P)"
claude="$(command -v "${KITLEGAL_CLAUDE_BIN:-claude}")" \
  || { echo "claude-modelo: no encuentro '${KITLEGAL_CLAUDE_BIN:-claude}' en el PATH" >&2; exit 2; }
export KITLEGAL_PASO_DE_WORKFLOW=1
export CLAUDE_CODE_DISABLE_BACKGROUND_TASKS=1
export PATH="$raiz/scripts/workflow/sin-modelo:$PATH"

exec "$claude" ${args[@]+"${args[@]}"} --setting-sources user,project
