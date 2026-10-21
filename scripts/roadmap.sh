#!/usr/bin/env bash
# Enseña el estado del roadmap desde docs/roadmap/: lo que está en curso, la cola en orden y cuántos hitos están
# hechos y cuántos ficheros tiene el backlog (docs/roadmap/README.md).
set -euo pipefail
cd "$(dirname "$0")/.."
. scripts/workflow/comun.sh

titulos() { # $@ = ficheros de hito
  local f id
  for f in "$@"; do
    [ -f "$f" ] || continue
    id=$(basename "$f" .md | sed -E 's/^[0-9]{3}-//; s/-.*$//')
    printf '  %-7s %s\n' "$id" "$(titulo_del_hito "$id")"
  done
}

echo "En curso:"
titulos docs/roadmap/en-curso/H*.md
echo "Siguientes, en orden:"
titulos docs/roadmap/siguientes/[0-9][0-9][0-9]-H*.md
printf 'Hechos: %s hitos. Backlog: %s ficheros.\n' \
  "$(find docs/roadmap/hechos -name 'H*.md' | wc -l | tr -d ' ')" \
  "$(find docs/roadmap/backlog -name '*.md' ! -name README.md | wc -l | tr -d ' ')"
