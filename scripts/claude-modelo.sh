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

exec "${KITLEGAL_CLAUDE_BIN:-claude}" ${args[@]+"${args[@]}"}
