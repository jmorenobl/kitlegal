#!/usr/bin/env bash
# Lanza el workflow spec-kit `hito` para un hito del roadmap.
#
#   scripts/hito.sh H0                  # desatendido (gates automáticos)
#   scripts/hito.sh H0 supervisado      # además pausa para revisión humana
#   scripts/hito.sh --resume <run_id> [key=value ...]
#   KITLEGAL_MODELO_JUEZ=fable@max KITLEGAL_MODELO_IMPLEMENTACION=sonnet@xhigh scripts/hito.sh H0
#
# Claude se ejecuta en modo headless (`claude -p`). Los permisos de
# herramientas vienen de .claude/settings.json; aquí solo se aceptan las
# ediciones de ficheros. Para un entorno aislado (contenedor/VM) puede
# sustituirse por --dangerously-skip-permissions.
set -euo pipefail

cd "$(dirname "$0")/.."

export SPECKIT_INTEGRATION_CLAUDE_EXTRA_ARGS="${SPECKIT_INTEGRATION_CLAUDE_EXTRA_ARGS:---permission-mode acceptEdits}"
# Los inputs modelo_* admiten <modelo>@<esfuerzo>; el wrapper lo traduce a
# --model/--effort. Sin él, `claude` rechazaría el valor.
export SPECKIT_INTEGRATION_CLAUDE_EXECUTABLE="${SPECKIT_INTEGRATION_CLAUDE_EXECUTABLE:-$PWD/scripts/claude-modelo.sh}"

if [ "${1:-}" = "--resume" ]; then
  run_id="${2:?run_id}"; shift 2
  args=()
  for kv in "$@"; do args+=(--input "$kv"); done
  exec specify workflow resume "$run_id" ${args[@]+"${args[@]}"}
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

# Modelo y esfuerzo por rol (<alias o nombre completo>[@esfuerzo]; roles en
# docs/WORKFLOW.md «Modelo por paso»). Ejemplo:
#   KITLEGAL_MODELO_IMPLEMENTACION=opus@max scripts/hito.sh H0
modelos=()
for rol in DECISION JUEZ REVISOR REDACCION IMPLEMENTACION ESCALADA ANALISIS; do
  var="KITLEGAL_MODELO_$rol"
  if [ -n "${!var:-}" ]; then
    modelos+=(--input "modelo_$(printf '%s' "$rol" | tr '[:upper:]' '[:lower:]')=${!var}")
  fi
done

exec specify workflow run hito --input "hito=$hito" --input "modo=$modo" ${modelos[@]+"${modelos[@]}"}
