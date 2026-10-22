#!/usr/bin/env bash
# Vuelve a instalar el workflow hito en el registro de spec-kit (.specify/workflows/workflow-registry.json), que es lo
# que enseña `specify workflow list`, desde su propio directorio: así el nombre, la versión y la descripción del
# registro son los de la cabecera de .specify/workflows/hito/workflow.yml, y la fuente queda relativa a la raíz en lugar
# de un directorio temporal que deja de existir. No cambia workflow.yml. Hay que ejecutarlo cada vez que cambia esa
# cabecera: TestElRegistroDeSpecKitEsElDelWorkflow (internal/workflow) falla en make ci si no, y se ejecuta al final.
set -euo pipefail
cd "$(dirname "$0")/.."

command -v specify >/dev/null 2>&1 || {
  echo "workflow-registro: falta specify (spec-kit), el que ejecuta el workflow hito (docs/WORKFLOW.md)" >&2
  exit 1
}

specify workflow add --dev .specify/workflows/hito
go test -count=1 -run '^TestElRegistroDeSpecKitEsElDelWorkflow$' ./internal/workflow/
