#!/usr/bin/env bash
# Informe final del hito (ADR 0018): lo único que la persona lee antes de decidir.
#
#   scripts/workflow/informe.sh <hito>             # escribe, commitea, empuja y lo pone de cuerpo de la propuesta
#   scripts/workflow/informe.sh <hito> --solo-ver  # lo escribe en stdout, sin tocar nada
#
# Se construye sin modelo, desde los artefactos y los gates del run:
#   1. estado: make ci local, CI y evals remotos sobre la cabeza, revisión final;
#   2. trazabilidad: cada FR/SC del spec → tareas (y su estado) → guiones de aceptación;
#   3. supuestos: decisiones conservadoras de clarify y de los correctores, y pendientes
#      de las rondas de juez agotadas;
#   4. tareas en cuarentena, con su parche y su nota;
#   5. lo que la constitución (capa 3) reserva a la persona: fuentes, anomalías, fuentes
#      nuevas, fixtures y esquemas, grabaciones y evidencias;
#   6. commits posteriores a la revisión final, que ningún juez vio;
#   7. duración del run.
set -euo pipefail
cd "$(dirname "$0")/../.."
. scripts/workflow/comun.sh

hito="${1:?uso: informe.sh <hito> [--solo-ver]}"
solo_ver="${2:-}"
d=$(feature_dir)
g="$d/gates"
rama=$(git branch --show-current)
titulo=$(awk -v h="$hito" 'index($0, "#### " h " · ") == 1 {print; exit}' docs/ROADMAP.md | sed -E 's/^#### [A-Z0-9.]+ · //')

# Ids de requisito que menciona una línea, con los rangos «FR-001 a FR-008» expandidos.
ids_de_linea() {
  printf '%s\n' "$1" | awk '{
    s = $0
    while (match(s, /(FR|SC)-[0-9]+ (a|al|–|-) (FR|SC)-[0-9]+/)) {
      r = substr(s, RSTART, RLENGTH); split(r, p, /[^A-Z0-9-]+/)
      pre = substr(p[1], 1, 3); a = substr(p[1], 4) + 0; b = substr(p[length(p)], 4) + 0
      if (b >= a && b - a < 200) for (i = a; i <= b; i++) printf "%s%03d\n", pre, i
      s = substr(s, RSTART + RLENGTH)
    }
    t = $0
    while (match(t, /(FR|SC)-[0-9]+/)) { print substr(t, RSTART, RLENGTH); t = substr(t, RSTART + RLENGTH) }
  }' | sort -u
}

estado_tarea() { # X, cuarentena o pendiente
  if jq -e --arg id "$1" '.tareas | has($id)' "$g/cuarentena.json" >/dev/null 2>&1; then echo cuarentena
  elif grep -qE "^[[:space:]]*- \[[Xx]\] $1([^0-9]|\$)" "$d/tasks.md"; then echo hecha
  else echo pendiente; fi
}

veredicto() { [ -f "$1" ] && jq -r .veredicto "$1" || echo "sin veredicto"; }

# Lo que la capa 3 de la constitución reserva a la persona, una línea por asunto.
capa3() {
  local cambiados nuevos_src f
  cambiados=$(git diff --name-status main...HEAD)
  printf '%s\n' "$cambiados" | awk '$2 == "docs/SOURCES.md" {print "- `docs/SOURCES.md` cambia: responsabilidad sobre qué fuentes se tocan y cómo."}'
  printf '%s\n' "$cambiados" | awk '$2 ~ /^data\/anomalias\// {print "- Regla de anomalías: `" $2 "` (" $1 ")."}'
  nuevos_src=$(printf '%s\n' "$cambiados" | awk '{print $NF}' | grep -oE '^internal/source/[^/]+' | sort -u || true)
  for f in $nuevos_src; do git cat-file -e "main:$f" 2>/dev/null || echo "- Adaptador de fuente nuevo: \`$f/\`."; done
  printf '%s\n' "$cambiados" | awk '$1 ~ /^M/ && ($2 ~ /(^|\/)testdata\// || $2 ~ /^schemas\//) {print "- Fixture o esquema EXISTENTE modificado: `" $2 "`."}'
  printf '%s\n' "$cambiados" | awk '$1 == "A" && ($2 ~ /^testdata\// || $2 ~ /^internal\/source\/.*\/testdata\// || $2 ~ /^schemas\//) {n++} END {if (n) print "- " n " fixtures o esquemas nuevos bajo testdata/ o schemas/."}'
  printf '%s\n' "$cambiados" | awk '$2 ~ /^evidencias\// {n++} END {if (n) print "- " n " ficheros de evidencia en evidencias/ (manifiesto con huellas)."}'
  if [ -f "$g/grabaciones.md" ]; then echo "- Grabaciones del paso grabar_datos: \`$g/grabaciones.md\`."; fi
}

informe() {
  echo "# Informe del hito $hito · $titulo"
  echo
  echo "Generado por el workflow \`hito\` el $(date -u +%Y-%m-%dT%H:%M:%SZ), sobre \`$(git rev-parse --short HEAD)\` de \`$rama\`."
  echo "Lo escribe scripts/workflow/informe.sh sin modelo, desde los artefactos de \`$d/\`. Fusionar (squash-merge) es una decisión humana:"
  echo "si algo de lo que sigue no es lo que se quería, se corrige la sección del hito en docs/ROADMAP.md y se relanza."
  echo

  echo "## 1. Estado"
  echo
  local ci remoto
  ci=$(tail -n 1 "$g/ci.log" 2>/dev/null | sed -n 's/^kitlegal-verificacion exit=//p')
  echo "- **make ci local**: $( [ "${ci:-x}" = 0 ] && echo "verde" || echo "ROJO (exit ${ci:-desconocido}; gates/ci.log)")."
  if [ -f "$g/cierre.json" ]; then
    remoto=$(jq -r 'if .verde then "verde" else "ROJO: " + ([.checks[] | select(.workflow != "" and (.bucket == "fail" or .bucket == "cancel")) | .name] | join(", ")) end' "$g/cierre.json")
    echo "- **CI y evals remotos** sobre \`$(jq -r '.sha[0:7]' "$g/cierre.json")\`: $remoto."
  else
    echo "- **CI y evals remotos**: sin medir."
  fi
  echo "- **Revisión final**: juez A $(veredicto "$g/revision-a.json"), juez B $(veredicto "$g/revision-b.json"), $(cat "$g/revision-rondas" 2>/dev/null || echo 0) rondas."
  local hechas cuar sin
  hechas=$(grep -cE '^[[:space:]]*- \[[Xx]\] T[0-9]+' "$d/tasks.md" || true)
  cuar=$(jq '.tareas | length' "$g/cuarentena.json" 2>/dev/null || echo 0)
  sin=$(( $(grep -cE '^[[:space:]]*- \[ \] T[0-9]+' "$d/tasks.md" || true) - cuar ))
  echo "- **Tareas**: $hechas hechas, $cuar en cuarentena, $sin pendientes sin cuarentena."
  echo "- **Diff**: $(git rev-list --count main..HEAD) commits; $(git diff --shortstat main...HEAD | sed 's/^ //')."
  echo

  echo "## 2. Trazabilidad"
  echo
  echo "| Requisito | Tareas | Aceptación | Estado |"
  echo "|---|---|---|---|"
  local req lineas tareas acept est t e
  for req in $(grep -oE '\*\*(FR|SC)-[0-9]+\*\*' "$d/spec.md" | tr -d '*' | awk '!v[$0]++'); do
    tareas=""; est=hecho
    while IFS= read -r l; do
      t=$(printf '%s' "$l" | grep -oE 'T[0-9]+' | head -1)
      ids_de_linea "$l" | grep -qx "$req" || continue
      e=$(estado_tarea "$t"); tareas="$tareas $t($e)"
      case "$e" in cuarentena) est="EN CUARENTENA";; pendiente) [ "$est" = "EN CUARENTENA" ] || est="PENDIENTE";; esac
    done < <(grep -E '^[[:space:]]*- \[[ xX]\] T[0-9]+' "$d/tasks.md")
    [ -n "$tareas" ] || est="SIN TAREA"
    acept=$(grep -lwF "$req" "$d"/aceptacion/* 2>/dev/null | xargs -n1 basename 2>/dev/null | tr '\n' ' ' || true)
    echo "| $req |${tareas:- —} | ${acept:-—} | $est |"
  done
  echo

  echo "## 3. Supuestos y pendientes"
  echo
  echo "Decisiones que el run tomó sin preguntar. Son las primeras que hay que leer."
  echo
  if [ -f "$g/clarify-respuestas.json" ]; then
    # Una línea por decisión que no salió de las fuentes; el texto entero y la alternativa
    # descartada están en el JSON, que es lo que se lee si la línea no convence.
    jq -r --arg f "$g/clarify-respuestas.json" '.respuestas[] | select(.criterio != "a")
      | "- **clarify \(.id)** (criterio \(.criterio)" + (if .conservadora then ", conservadora" else "" end) + "): "
        + (.respuesta | if length > 300 then .[0:300] + "… (completo en `\($f)`)" else . end)' "$g/clarify-respuestas.json"
  fi
  [ -f "$g/supuestos.md" ] && grep -E '^- ' "$g/supuestos.md"
  local p
  for p in "$g"/*-pendiente.md; do
    [ -f "$p" ] || continue
    echo; echo "**$(basename "$p")**:"; echo; sed 's/^/> /' "$p"
  done
  echo

  echo "## 4. Tareas en cuarentena"
  echo
  if [ "$(jq '.tareas | length' "$g/cuarentena.json" 2>/dev/null || echo 0)" -gt 0 ]; then
    jq -r --arg g "$g" '.tareas | to_entries[] | "- **\(.key)**: \(.value.motivo). Parche `\(.value.parche)`; causa en `\($g)/tarea-\(.key).md`."' "$g/cuarentena.json"
  else
    echo "Ninguna."
  fi
  echo

  echo "## 5. Revisión que la constitución reserva a la persona"
  echo
  local capa3
  capa3=$(capa3)
  if [ -n "$capa3" ]; then printf '%s\n' "$capa3"; else echo "Nada: el hito no toca fuentes, anomalías, adaptadores, fixtures, esquemas, grabaciones ni evidencias."; fi
  echo

  echo "## 6. Cambios posteriores a la revisión final"
  echo
  local ult
  ult=$(git log --format=%H --grep="^docs($hito): veredictos de la revisión final" -n 1 main..HEAD || true)
  if [ -n "$ult" ] && [ -n "$(git log --oneline "$ult"..HEAD)" ]; then
    echo "Commits que ningún juez vio (correcciones del cierre y este informe):"; echo
    git log --format='- `%h` %s' "$ult"..HEAD
  else
    echo "Ninguno."
  fi
  echo

  echo "## 7. Cómo comprobarlo y consumo"
  echo
  echo "- Escenarios manuales: \`$d/quickstart.md\`. Suite de aceptación congelada: \`$d/aceptacion/\` (activada en \`internal/app/testdata/script/\`)."
  local run
  run=$(for r in $(ls -t .specify/workflows/runs 2>/dev/null); do
          [ "$(jq -r '.inputs.hito // ""' ".specify/workflows/runs/$r/inputs.json" 2>/dev/null)" = "$hito" ] && { echo "$r"; break; }; done || true)
  if [ -n "$run" ]; then
    local t0 s
    t0=$(head -1 ".specify/workflows/runs/$run/log.jsonl" | jq -r .timestamp)
    s=$(( $(date +%s) - $(date -j -u -f '%Y-%m-%dT%H:%M:%S' "${t0%%.*}" +%s 2>/dev/null || date -u -d "${t0%%.*}" +%s) ))
    echo "- Run \`$run\`: $((s / 3600)) h $(((s % 3600) / 60)) min de reloj. Coste por paso y por rol: \`scripts/coste-run.sh $run\`."
  fi
}

if [ "$solo_ver" = --solo-ver ]; then informe; exit 0; fi

informe > "$g/informe-final.md"
git add -A -- "$d"
git diff --cached --quiet || git commit -q -m "docs($hito): informe final del run" -- "$d"
if command -v gh >/dev/null 2>&1 && gh auth status >/dev/null 2>&1; then
  git push -q -u origin "$rama" || echo "informe: no se pudo empujar; el cuerpo de la propuesta queda sin actualizar" >&2
  cuerpo="$g/informe-final.md"
  if [ "$(wc -c < "$cuerpo")" -gt 60000 ]; then # tope de GitHub para el cuerpo: 65 536 caracteres
    cuerpo=$(mktemp); head -c 59000 "$g/informe-final.md" > "$cuerpo"
    printf '\n\n…(recortado; el informe completo está en `%s`)\n' "$g/informe-final.md" >> "$cuerpo"
  fi
  gh pr edit "$rama" --body-file "$cuerpo" >/dev/null && echo "informe publicado como cuerpo de $(gh pr view "$rama" --json url --jq .url)" \
    || echo "informe: no se pudo actualizar el cuerpo de la propuesta" >&2
fi
echo "informe final: $g/informe-final.md"
