#!/usr/bin/env bash
# Lanza el workflow spec-kit `hito` para un hito del roadmap.
#
#   scripts/hito.sh H0                  # desatendido (gates automáticos)
#   scripts/hito.sh H0 supervisado      # además pausa para revisión humana
#   scripts/hito.sh --resume <run_id> [key=value ...]
#
# Claude se ejecuta en modo headless (`claude -p`). Los permisos de
# herramientas vienen de .claude/settings.json; aquí solo se aceptan las
# ediciones de ficheros. Para un entorno aislado (contenedor/VM) puede
# sustituirse por --dangerously-skip-permissions.
set -euo pipefail

cd "$(dirname "$0")/.."

export SPECKIT_INTEGRATION_CLAUDE_EXTRA_ARGS="${SPECKIT_INTEGRATION_CLAUDE_EXTRA_ARGS:---permission-mode acceptEdits}"

if [ "${1:-}" = "--resume" ]; then
  run_id="${2:?run_id}"; shift 2
  args=()
  for kv in "$@"; do args+=(--input "$kv"); done
  exec specify workflow resume "$run_id" "${args[@]}"
fi

hito="${1:?uso: scripts/hito.sh H<n> [desatendido|supervisado]}"
modo="${2:-desatendido}"

if [ -n "$(git status --porcelain)" ]; then
  echo "el árbol de trabajo tiene cambios sin commitear; el hito parte de main limpio" >&2
  exit 1
fi
if [ "$(git branch --show-current)" != "main" ]; then
  echo "sitúate en main antes de lanzar un hito (rama actual: $(git branch --show-current))" >&2
  exit 1
fi

exec specify workflow run hito --input "hito=$hito" --input "modo=$modo"
