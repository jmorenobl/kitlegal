#!/usr/bin/env bash
# Cierre del hito en la plataforma remota (ADR 0018): lo hace el workflow, sin modelo,
# DESPUÉS de la revisión final, sobre la cabeza que se va a fusionar.
#
#   scripts/workflow/cierre.sh publicar <hito>   # empuja la rama y abre la propuesta si no existe
#   scripts/workflow/cierre.sh medir <hito>      # vuelve a medir: CI y evals sobre la cabeza → JSON
#   scripts/workflow/cierre.sh evals <hito>      # solo recoge los informes de evals de la última medición
#
# Tras medir, cada trabajo del flujo `evals` que terminó deja su informe.json en
# gates/evals/<skill>.json: el job lo imprime entero en su registro entre las marcas
# «--- inicio de informe.json ---» y «--- fin de informe.json ---» (scripts/evals.sh),
# y de ahí lo copia este paso, sin modelo. El informe final saca de esos ficheros la
# tasa de cada eval (ADR 0028). `evals` hace solo esa recogida, sin empujar ni poner
# etiquetas, sobre las comprobaciones que ya están en gates/cierre.json: sirve para un
# run que midió con un workflow anterior.
#
# En H5 y H6 la ejecución de cierre de las evals era una tarea [plataforma] dentro
# del bucle, antes de la revisión final: cada corrección de la revisión la dejaba
# sin cubrir la cabeza y hubo que repetirla ronda tras ronda. Aquí va una sola vez
# al final, y si sale en rojo el reparador del cierre arregla y se vuelve a medir.
#
# `medir` pone la etiqueta `evals` (el botón de «vuelve a medir» del job,
# .github/workflows/evals.yml), espera a que terminen todas las comprobaciones de
# GitHub Actions sobre la cabeza y escribe gates/cierre.json y, si algo falla, el
# final del log de cada ejecución en rojo en gates/cierre.log. Solo cuentan las
# comprobaciones con flujo de GitHub Actions: los estados de Codecov son
# informativos (su «:x:» no es un rojo). Nunca fusiona, ni empuja a main, ni
# fuerza: lo impiden los permisos de .claude/settings.json y la protección de main en GitHub (ADR 0021).
set -euo pipefail
cd "$(dirname "$0")/../.."
. scripts/workflow/comun.sh

sub="${1:?uso: cierre.sh publicar|medir|evals <hito>}"
hito="${2:?uso: cierre.sh publicar|medir|evals <hito>}"
d=$(feature_dir)
rama=$(git branch --show-current)
espera_max="${KITLEGAL_CIERRE_ESPERA_MAX:-10800}" # 3 h: el job de evals tiene un tope de 120 min por skill

dossier() {
  printf '# Dossier: %s\n\nPaso de cierre del hito %s (%s).\n\n%s\n\nAl resolverlo, reanudar con scripts/hito.sh --resume <run_id>.\n' \
    "$1" "$hito" "$(date -u +%Y-%m-%dT%H:%M:%SZ)" "$2" > "$d/gates/dossier.md"
  echo "cierre detenido ($1). Dossier en $d/gates/dossier.md" >&2
  exit 3
}

# Copia a gates/evals/<skill>.json el informe.json de cada trabajo de evals de gates/cierre.json.
# Un trabajo cuyo registro no trae un informe legible no deja fichero: el informe final
# lo dice. Nunca para el run: las tasas son información para la persona, no un gate.
recoger_evals() {
  local e="$d/gates/evals" fila nombre skill link run job destino
  rm -rf "$e"; mkdir -p "$e"
  jq -c '.checks[] | select(.workflow == "evals" and (.bucket == "pass" or .bucket == "fail"))' "$d/gates/cierre.json" |
  while IFS= read -r fila; do
    nombre=$(jq -r .name <<<"$fila"); link=$(jq -r .link <<<"$fila")
    skill=$(printf '%s' "$nombre" | sed -nE 's/^evals \(([a-z0-9-]+)\)$/\1/p')
    run=$(printf '%s' "$link" | sed -nE 's#.*/actions/runs/([0-9]+)/job/[0-9]+.*#\1#p')
    job=$(printf '%s' "$link" | sed -nE 's#.*/actions/runs/[0-9]+/job/([0-9]+).*#\1#p')
    [ -n "$skill" ] && [ -n "$run" ] && [ -n "$job" ] || continue
    destino="$e/$skill.json"
    # gh run view --log antepone trabajo, paso y hora a cada línea: se quitan con cut y sed.
    gh run view "$run" --job "$job" --log 2>/dev/null | cut -f3- | sed -E 's/^[0-9-]+T[0-9:.]+Z //' \
      | awk '$0 == "--- inicio de informe.json ---" {p = 1; next} $0 == "--- fin de informe.json ---" {p = 0} p' > "$destino" || true
    if jq -e '.skill and (.tasas | type == "array")' "$destino" >/dev/null 2>&1; then
      echo "evals: informe de $skill en $destino" >&2
    else
      rm -f "$destino"; echo "evals: el registro de $nombre no trae un informe.json legible" >&2
    fi
  done
}

if [ "$sub" = evals ]; then
  [ -f "$d/gates/cierre.json" ] || { echo "evals: no hay gates/cierre.json; el cierre no ha medido" >&2; exit 1; }
  command -v gh >/dev/null 2>&1 && gh auth status >/dev/null 2>&1 || { echo "evals: gh no tiene sesión" >&2; exit 1; }
  recoger_evals; exit 0
fi

[ "$rama" != main ] || dossier "rama equivocada" "La rama actual es main; el cierre nunca empuja main."
command -v gh >/dev/null 2>&1 && gh auth status >/dev/null 2>&1 \
  || dossier "credencial ausente" "gh no está instalado o no tiene sesión (gh auth status). Sin ella no se puede publicar ni medir."

case "$sub" in
  publicar)
    rm -f "$d/gates/cierre-rondas" # la medición cuenta sus rondas desde la publicación
    # Los registros del run (supuestos, cuarentena, veredictos) viven en el directorio del
    # feature y algunos pasos los dejan sin commitear: se versionan antes de publicar.
    git add -A -- "$d"
    git diff --cached --quiet || git commit -q -m "docs($hito): registros del run" -- "$d"
    test -z "$(git status --porcelain)" || { git status --short >&2; dossier "árbol sucio" "Quedan cambios sin commitear fuera de $d antes de publicar."; }
    git push -u origin "$rama" >&2 || dossier "dependencia externa inaccesible" "git push -u origin $rama falló."
    if url=$(gh pr view "$rama" --json url --jq .url 2>/dev/null); then echo "propuesta existente: $url"; exit 0; fi
    titulo=$(awk -v h="$hito" 'index($0, "#### " h " · ") == 1 {print; exit}' docs/ROADMAP.md | sed -E 's/^#### [A-Z0-9.]+ · //')
    cuerpo="$d/gates/informe-final.md"
    if [ ! -f "$cuerpo" ]; then
      cuerpo=$(mktemp)
      printf 'Hito %s · %s\n\nEn curso: el workflow publica la rama para medir el cierre. El informe final sustituirá este cuerpo.\n' "$hito" "$titulo" > "$cuerpo"
    fi
    gh pr create --base main --head "$rama" --title "feat($hito): $titulo" --body-file "$cuerpo" >&2 \
      || dossier "dependencia externa inaccesible" "gh pr create falló."
    echo "propuesta abierta: $(gh pr view "$rama" --json url --jq .url)";;

  medir)
    git add -A -- "$d"
    git diff --cached --quiet || git commit -q -m "docs($hito): registros del run" -- "$d"
    sha=$(git rev-parse HEAD)
    remoto=$(git rev-parse "origin/$rama" 2>/dev/null || echo "")
    [ "$remoto" = "$sha" ] || git push -u origin "$rama" >&2 || dossier "dependencia externa inaccesible" "git push -u origin $rama falló."
    # La etiqueta solo dispara al ponerse: se quita y se vuelve a poner para medir esta cabeza.
    gh pr edit "$rama" --remove-label evals >/dev/null 2>&1 || true
    gh pr edit "$rama" --add-label evals >/dev/null || dossier "dependencia externa inaccesible" "No se pudo poner la etiqueta evals."
    inicio=$(date +%s)
    # Las comprobaciones de la cabeza tardan en registrarse: se espera a verlas, y después a que acaben.
    while :; do
      checks=$(gh pr checks "$rama" --json name,state,bucket,link,workflow,startedAt 2>/dev/null || echo '[]')
      n_evals=$(jq '[.[] | select(.workflow == "evals")] | length' <<<"$checks")
      pendientes=$(jq '[.[] | select(.workflow != "" and (.bucket == "pending"))] | length' <<<"$checks")
      [ "$n_evals" -gt 0 ] && [ "$pendientes" -eq 0 ] && break
      [ $(( $(date +%s) - inicio )) -lt "$espera_max" ] || dossier "dependencia externa inaccesible" \
        "Las comprobaciones de la cabeza $sha no terminaron en ${espera_max}s. Último estado: $checks"
      sleep 60
    done
    rojos=$(jq -c '[.[] | select(.workflow != "" and (.bucket == "fail" or .bucket == "cancel"))]' <<<"$checks")
    verde=true; [ "$(jq length <<<"$rojos")" -eq 0 ] || verde=false
    jq -n --arg sha "$sha" --argjson verde "$verde" --argjson checks "$checks" --arg fecha "$(date -u +%Y-%m-%dT%H:%M:%SZ)" \
      '{sha:$sha, verde:$verde, fecha:$fecha, checks:$checks}' > "$d/gates/cierre.json"
    : > "$d/gates/cierre.log"
    for link in $(jq -r '.[].link' <<<"$rojos"); do
      run=$(printf '%s' "$link" | sed -nE 's#.*/actions/runs/([0-9]+).*#\1#p')
      [ -n "$run" ] || continue
      # gh run view --log antepone trabajo, paso y hora a cada línea: se quita con cut.
      { echo "== $link"; gh run view "$run" --log-failed 2>&1 | cut -f3- | tail -150; } >> "$d/gates/cierre.log"
    done
    recoger_evals
    n=$(( $(cat "$d/gates/cierre-rondas" 2>/dev/null || echo 0) + 1 )); echo "$n" > "$d/gates/cierre-rondas"
    jq -n --arg sha "$sha" --argjson verde "$verde" --argjson rojos "$rojos" --argjson n "$n" '{sha:$sha, verde:$verde, ronda:$n, rojos:[$rojos[].name]}';;

  *) echo "subcomando desconocido: $sub" >&2; exit 2;;
esac
