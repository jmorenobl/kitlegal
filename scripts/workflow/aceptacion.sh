#!/usr/bin/env bash
# Suite de aceptación congelada del workflow `hito` (ADR 0018).
#
#   scripts/workflow/aceptacion.sh rojo-primero   # verificación de la tarea [aceptacion]
#   scripts/workflow/aceptacion.sh congelar       # tras el commit de la tarea [aceptacion]
#   scripts/workflow/aceptacion.sh activar <hito> # tras el bucle de tareas, antes del make ci del hito
#
# La primera tarea del hito, etiquetada [aceptacion], escribe a partir del spec
# los guiones e2e que describen la entrega (<feature_dir>/aceptacion/*.txtar) y,
# si el hito entrega o cambia una skill, sus evals (evals/<skill>/). Los guiones
# viven en el directorio del feature, donde `make ci` no los ejecuta, para que
# cada tarea pueda seguir dejando `make ci` en verde mientras el hito avanza.
#
# rojo-primero  cada guion, copiado un momento al directorio de guiones de
#               internal/app, tiene que FALLAR contra el código de hoy, y por
#               una aserción: un guion que pasa sin implementación, o que falla
#               porque no se puede leer o no compila, no prueba nada.
# congelar      guarda la huella de cada fichero de la suite en
#               gates/aceptacion-congelada.json; desde ahí el guardián de diff
#               rechaza cualquier cambio en ellos.
# activar       copia los guiones al directorio de guiones de internal/app, los
#               añade a la huella y lo commitea: desde ahí `make ci` los ejecuta
#               y el código tiene que hacerlos pasar sin tocarlos.
set -euo pipefail
cd "$(dirname "$0")/../.."
. scripts/workflow/comun.sh

d=$(feature_dir)
suite="$d/aceptacion"
guiones_e2e=internal/app/testdata/script
congelada="$d/gates/aceptacion-congelada.json"

guiones() { ls "$suite"/*.txtar 2>/dev/null || true; }

case "${1:?uso: aceptacion.sh rojo-primero | congelar | activar <hito>}" in
  rojo-primero)
    t="$d/gates/tarea-actual.json"; base=$(jq -r .base "$t")
    evals_nuevas=$(git diff --name-only --diff-filter=A "$base" -- evals/; git ls-files --others --exclude-standard -- evals/)
    if [ -z "$(guiones)" ] && [ -z "$evals_nuevas" ]; then
      if grep -qiE '^[^|]*Aceptación e2e: no aplica' "$d/plan.md" 2>/dev/null; then
        echo "aceptación: el plan declara que el hito no tiene aceptación e2e"; exit 0
      fi
      echo "aceptación: la tarea [aceptacion] no dejó guiones en $suite/ ni evals nuevas en evals/"; exit 1
    fi
    verdes=""; invalidos=""
    for g in $(guiones); do
      nombre="zz-rojo-$(basename "$g" .txtar)"
      copia="$guiones_e2e/$nombre.txtar"
      cp "$g" "$copia"
      rc=0; salida=$(go test -count=1 -v -run "^TestEntregaDelHito\$/^$nombre\$" ./internal/app/ 2>&1) || rc=$?
      rm -f "$copia"
      if ! printf '%s' "$salida" | grep -qF -- "--- " || ! printf '%s' "$salida" | grep -qF "TestEntregaDelHito/$nombre"; then
        invalidos="$invalidos $g"; printf '%s\n' "$salida" | tail -20      # ni siquiera llegó a ejecutarse
      elif [ "$rc" -eq 0 ]; then verdes="$verdes $g"
      elif printf '%s' "$salida" | grep -qE 'unknown command|cannot parse|unexpected command|usage: '; then
        invalidos="$invalidos $g"; printf '%s\n' "$salida" | tail -20
      fi
    done
    rc=0
    [ -z "$verdes" ] || { echo "aceptación: estos guiones pasan sin implementación y no prueban nada:$verdes"; rc=1; }
    [ -z "$invalidos" ] || { echo "aceptación: estos guiones fallan por no poder leerse o compilarse, no por una aserción:$invalidos"; rc=1; }
    [ "$rc" -eq 0 ] && echo "aceptación: $(guiones | wc -l | tr -d ' ') guiones en rojo por una aserción; evals nuevas: $(printf '%s\n' "$evals_nuevas" | grep -c . || true)"
    exit "$rc";;

  congelar)
    t="$d/gates/tarea-actual.json"; base=$(jq -r .base "$t"); id=$(jq -r .id "$t")
    ficheros=$( { guiones; ls "$suite"/* 2>/dev/null; git diff --name-only --diff-filter=A "$base" HEAD -- evals/; } | sort -u | grep . || true)
    objeto='{}'
    for f in $ficheros; do objeto=$(jq --arg f "$f" --arg h "$(huella "$f")" '. + {($f): $h}' <<<"$objeto"); done
    jq -n --arg tarea "$id" --arg commit "$(git rev-parse HEAD)" --argjson ficheros "$objeto" \
      '{tarea:$tarea, commit:$commit, ficheros:$ficheros}' > "$congelada"
    echo "aceptación congelada: $(jq '.ficheros | length' "$congelada") ficheros de $id en $congelada";;

  activar)
    hito="${2:?uso: aceptacion.sh activar <hito>}"
    [ -f "$congelada" ] || { echo "aceptación: no hay suite congelada; nada que activar"; exit 0; }
    prefijo=$(printf '%s' "$hito" | tr '[:upper:]' '[:lower:]' | tr . -)
    activados=""
    for g in $(guiones); do
      destino="$guiones_e2e/$prefijo-$(basename "$g")"
      if [ -f "$destino" ] && cmp -s "$g" "$destino"; then continue; fi
      [ ! -f "$destino" ] || { echo "aceptación: $destino ya existe con otro contenido; no se sobrescribe" >&2; exit 1; }
      cp "$g" "$destino"; activados="$activados $destino"
      jq --arg f "$destino" --arg h "$(huella "$destino")" '.ficheros += {($f): $h}' "$congelada" > "$congelada.tmp" && mv "$congelada.tmp" "$congelada"
    done
    [ -n "$activados" ] || { echo "aceptación: suite ya activada"; exit 0; }
    git add -- $activados "$congelada"
    git commit -q -m "test($hito): activa la suite de aceptación congelada" -- $activados "$congelada"
    echo "aceptación activada:$activados"; git log --oneline -1;;

  *) echo "subcomando desconocido: $1" >&2; exit 2;;
esac
