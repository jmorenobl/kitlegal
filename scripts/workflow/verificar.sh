#!/usr/bin/env bash
# Batería determinista del workflow `hito`: guardián de diff y `make ci`.
#
#   scripts/workflow/verificar.sh tarea           # tras implementar o reparar una tarea
#   scripts/workflow/verificar.sh global [<base>] # tras reparar el hito, el barrido, un corrector o el cierre
#
# El log completo queda en <feature_dir>/gates/ci.log y su última línea es
# `kitlegal-verificacion exit=N`, que estado_cierre lee en lugar de fiarse de
# salidas de pasos que quizá no se ejecutaron en la iteración (docs/WORKFLOW.md,
# «Limitaciones conocidas»). Una violación del guardián es un rojo como otro:
# el reparador la ve en el log y la arregla, y el run no se para.
#
# Una tarea [aceptacion] además tiene que dejar su suite en rojo contra el código
# de hoy («rojo primero», scripts/workflow/aceptacion.sh): un guion que ya pasa
# sin implementación no prueba nada. Y una tarea [datos] que deja un manifiesto
# de grabación tiene que poder grabarse (scripts/workflow/grabar-datos.sh comprobar).
set -uo pipefail
cd "$(dirname "$0")/../.."
. scripts/workflow/comun.sh

modo="${1:?uso: verificar.sh tarea | global <base>}"
d=$(feature_dir)
mkdir -p "$d/gates"; log="$d/gates/ci.log"

cerrar() { # $1: código; deja la línea de resultado y el final del log a la vista
  printf '\nkitlegal-verificacion exit=%s\n' "$1" >> "$log"
  if [ "$1" -eq 0 ]; then echo "verificación en verde ($modo)"
  else echo "verificación FALLIDA ($modo, exit $1). Últimas 80 líneas de $log:"; tail -80 "$log"; fi
  exit "$1"
}

: > "$log"
if [ "$modo" = tarea ]; then
  scripts/workflow/guardian-diff.sh tarea >> "$log" 2>&1 || cerrar 1
  if [ "$(jq -r '.aceptacion // false' "$d/gates/tarea-actual.json")" = true ]; then
    scripts/workflow/aceptacion.sh rojo-primero >> "$log" 2>&1 || cerrar 1
  fi
  # Una tarea [datos] con manifiesto de grabación: fuente revisada en main y test de grabación.
  scripts/workflow/grabar-datos.sh comprobar >> "$log" 2>&1 || cerrar 1
else
  # Sin base explícita, la que fijó el último `scripts/workflow/global.sh base`.
  base="${2:-$(cat "$d/gates/base-global" 2>/dev/null || true)}"
  [ -n "$base" ] || { echo "verificar.sh global: sin base (ni argumento ni gates/base-global)" >> "$log"; cerrar 2; }
  scripts/workflow/guardian-diff.sh global "$base" >> "$log" 2>&1 || cerrar 1
fi

if [ -f Makefile ] && grep -qE '^ci:' Makefile; then cmd="make ci"
elif [ -f go.mod ]; then cmd="go build ./... && go vet ./... && go test -race ./..."
else echo "sin Makefile ni go.mod todavía: nada que verificar" >> "$log"; cerrar 0; fi

echo "== $cmd" >> "$log"
sh -c "$cmd" >> "$log" 2>&1
cerrar $?
