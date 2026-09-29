#!/usr/bin/env bash
# Comprobaciones mecánicas de los artefactos del hito, antes de cada juez (capa 1).
#
#   scripts/workflow/precheck.sh spec|plan|tasks
#
# Escribe un defecto por línea en <feature_dir>/gates/<fase>-precheck.txt (vacío si
# no hay ninguno) y sale siempre con 0: un defecto mecánico no para el run, entra
# en la ronda como motivo del juez y lo arregla el corrector (scripts/workflow/gate.sh
# fuerza el rechazo mientras quede alguno). Lo que se puede comprobar con un
# script nunca se delega a un modelo (constitución, «Gates»).
set -euo pipefail
cd "$(dirname "$0")/../.."
. scripts/workflow/comun.sh

fase="${1:?uso: precheck.sh spec|plan|tasks}"
d=$(feature_dir)
mkdir -p "$d/gates"
salida="$d/gates/$fase-precheck.txt"
: > "$salida"
defecto() { printf '%s\n' "$*" >> "$salida"; }
lineas_tarea() { grep -E '^[[:space:]]*- \[[ xX]\] T[0-9]+' "$d/tasks.md" || true; }

case "$fase" in
  spec)
    s="$d/spec.md"
    [ -f "$s" ] || { defecto "falta spec.md"; cat "$salida"; exit 0; }
    ! grep -q "NEEDS CLARIFICATION" "$s" || defecto "spec.md conserva [NEEDS CLARIFICATION]: $(grep -c 'NEEDS CLARIFICATION' "$s") marcas"
    grep -qi "^## Fuera de alcance" "$s" || defecto "spec.md no tiene sección '## Fuera de alcance'"
    ls "$d"/checklists/*.md >/dev/null 2>&1 || defecto "no hay checklists/"
    ! grep -qE '^[[:space:]]*- \[ \]' "$d"/checklists/*.md 2>/dev/null || defecto "checklists con ítems sin marcar: $(grep -hE '^[[:space:]]*- \[ \]' "$d"/checklists/*.md | head -5 | tr '\n' ' ')"
    # Fase A del gate de entrada: requisitos con id estable y único, y criterios de aceptación comprobables.
    frs=$(grep -oE '\*\*(FR|SC)-[0-9]+\*\*' "$s" | tr -d '*' || true)
    printf '%s\n' "$frs" | grep -q '^FR-' || defecto "spec.md no tiene ningún requisito con id **FR-nnn**"
    printf '%s\n' "$frs" | grep -q '^SC-' || defecto "spec.md no tiene ningún criterio de éxito con id **SC-nnn**"
    dup=$(printf '%s\n' "$frs" | grep . | sort | uniq -d | tr '\n' ' ')
    [ -z "$dup" ] || defecto "ids de requisito definidos más de una vez: $dup"
    grep -qE '\*\*(Dado|Given)\*\*|(^|[^A-Za-z])(Dado|Given) .*(Cuando|When) .*(Entonces|Then)' "$s" \
      || defecto "spec.md no tiene ningún escenario de aceptación Dado/Cuando/Entonces"
    rotos=$(grep -oE '(FR|SC)-[0-9]+' "$s" | sort -u | while read -r r; do printf '%s\n' "$frs" | grep -qx "$r" || echo "$r"; done | tr '\n' ' ')
    [ -z "$(printf '%s' "$rotos" | tr -d ' ')" ] || defecto "spec.md cita requisitos que no define: $rotos"
    if [ -f "$d/gates/clarify-preguntas.json" ]; then
      n=$(jq '.preguntas | length' "$d/gates/clarify-preguntas.json")
      q=$(grep -cE '^[[:space:]]*- Q:' "$s" || true)
      [ "$q" -ge "$n" ] || defecto "## Clarifications tiene $q entradas y hubo $n preguntas"
    fi;;

  plan)
    [ -f "$d/plan.md" ] || defecto "falta plan.md"
    [ -f "$d/research.md" ] || defecto "falta research.md"
    [ ! -f "$d/plan.md" ] || ! grep -q "NEEDS CLARIFICATION" "$d/plan.md" || defecto "plan.md conserva NEEDS CLARIFICATION"
    [ ! -f "$d/plan.md" ] || grep -q "## Constitution Check" "$d/plan.md" || defecto "plan.md sin '## Constitution Check'"
    [ ! -f "$d/plan.md" ] || grep -qiE 'Aceptación e2e' "$d/plan.md" \
      || defecto "plan.md no dice qué guiones de aceptación e2e describen la entrega (o 'Aceptación e2e: no aplica' con el motivo)"
    # Dónde vive el control de cada umbral (ADR 0029): lo lee el informe final sin
    # modelo. Aquí solo la forma; qué umbrales faltan lo juzga el juez (criterio o).
    if [ -f "$d/plan.md" ]; then
      if ! grep -q '^## Controles de umbral' "$d/plan.md"; then
        defecto "plan.md sin '## Controles de umbral': una fila '| Requisito | Umbral | Control | Dónde |' por cada umbral que se mide en make ci o en el cierre, o 'Ninguno.' con el motivo (ADR 0029)"
      else
        definidos=$(grep -oE '\*\*(FR|SC)-[0-9]+\*\*' "$d/spec.md" 2>/dev/null | tr -d '*' || true)
        filas=$(controles_de_umbral "$d/plan.md")
        if [ -z "$filas" ]; then
          awk '/^## /{p = ($0 ~ /^## Controles de umbral/); next} p && /^Ninguno/ {f = 1} END {exit !f}' "$d/plan.md" \
            || defecto "'## Controles de umbral' no tiene filas ni dice 'Ninguno.'"
        fi
        while IFS="$(printf '\t')" read -r req donde; do
          [ -n "$req$donde" ] || continue
          ids=$(ids_de_linea "$req")
          [ -n "$ids" ] || defecto "Controles de umbral: la fila '$req' no nombra ningún requisito (FR-nnn o SC-nnn)"
          for r in $ids; do printf '%s\n' "$definidos" | grep -qx "$r" || defecto "Controles de umbral: $r no está definido en spec.md"; done
          [ -n "$(controles_de_celda "$donde")" ] \
            || defecto "Controles de umbral: la fila '$req' no dice dónde vive el control (ci:<ruta>, ci:<ruta>:<Test> o evals:<skill>:<nombre>)"
        done <<<"$filas"
      fi
    fi;;

  tasks)
    [ -f "$d/gates/analyze.md" ] || defecto "falta gates/analyze.md"
    [ -f "$d/tasks.md" ] || { defecto "falta tasks.md"; cat "$salida"; exit 0; }
    ! grep -qE '^[[:space:]]*- \[ \]' "$d"/checklists/*.md 2>/dev/null || defecto "checklists con ítems sin marcar"
    grep -qE '^[[:space:]]*- \[ \] T[0-9]+' "$d/tasks.md" || defecto "tasks.md no contiene tareas con el formato '- [ ] Tnnn'"
    dup=$(lineas_tarea | grep -oE '^[[:space:]]*- \[[ xX]\] T[0-9]+' | grep -oE 'T[0-9]+' | sort | uniq -d | tr '\n' ' ')
    [ -z "$dup" ] || defecto "ids de tarea duplicados: $dup"
    lineas_tarea | tr '`,;()' '     ' | grep -vE '(\.?[A-Za-z0-9_-]+/)+[A-Za-z0-9_.*-]*|[A-Za-z0-9_.-]+\.(go|md|yml|yaml|json|sh|txt|mod|sum|toml|txtar)|\.[A-Za-z0-9_-]*ignore|\.editorconfig|Makefile|LICENSE' \
      | cut -c1-160 | sed 's/^/tarea sin rutas declaradas (el guardián la rechazaría): /' >> "$salida" || true
    lineas_tarea | grep -E '(^|[^A-Za-z0-9_])(testdata|schemas)/' | grep -v '\[datos\]' | cut -c1-160 \
      | sed 's/^/tarea que toca testdata\/ o schemas\/ sin [datos]: /' >> "$salida" || true
    lineas_tarea | grep -E '(^|[^A-Za-z0-9_])evidencias/' | cut -c1-160 \
      | sed 's/^/tarea que escribe en evidencias\/ (solo lo hace el paso grabar_datos): /' >> "$salida" || true
    # El cierre en la plataforma (push, propuesta de cambio, CI, evals remotas) lo hace el workflow tras la revisión.
    lineas_tarea | grep -iE '\[plataforma\]|gh pr|gh run|git push|pull request|propuesta de cambio' | cut -c1-160 \
      | sed 's/^/tarea de plataforma (el cierre lo hace el workflow, no una tarea): /' >> "$salida" || true
    # Nada en medio lo hace una persona (ADR 0018).
    lineas_tarea | grep -iE '(la|una) persona (graba|escribe|revisa|aporta|descarga|comprueba|decide|aprueba)|pausa humana|en la pausa' | cut -c1-160 \
      | sed 's/^/tarea que exige a una persona a mitad del run: /' >> "$salida" || true
    lineas_tarea | grep -iE 'KITLEGAL_RECORD' | cut -c1-160 \
      | sed 's/^/tarea que graba con KITLEGAL_RECORD (solo lo hace el paso grabar_datos): /' >> "$salida" || true
    n_acept=$(lineas_tarea | grep -c '\[aceptacion\]' || true)
    if grep -qiE 'Aceptación e2e: no aplica' "$d/plan.md" 2>/dev/null; then
      [ "$n_acept" -eq 0 ] || defecto "el plan dice 'Aceptación e2e: no aplica' y hay $n_acept tareas [aceptacion]"
    else
      [ "$n_acept" -eq 1 ] || defecto "tiene que haber exactamente una tarea [aceptacion] y hay $n_acept"
      primera=$(lineas_tarea | head -1)
      printf '%s' "$primera" | grep -q '\[aceptacion\]' || defecto "la tarea [aceptacion] tiene que ser la primera de tasks.md"
    fi;;

  *) echo "fase desconocida: $fase" >&2; exit 2;;
esac

if [ -s "$salida" ]; then echo "precheck_$fase: $(wc -l < "$salida" | tr -d ' ') defectos mecánicos"; cat "$salida"
else echo "precheck_$fase ok"; fi
