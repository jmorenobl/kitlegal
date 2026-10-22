#!/usr/bin/env bash
# Cierre del hito en la plataforma remota (ADR 0018): lo hace el workflow, sin modelo,
# DESPUÉS de la revisión final, sobre la cabeza que se va a fusionar.
#
#   scripts/workflow/cierre.sh iniciar <hito>          # antes del bucle de revisión y cierre: contadores a cero
#   scripts/workflow/cierre.sh publicar <hito>         # empuja la rama y abre la propuesta si no existe
#   scripts/workflow/cierre.sh medir <hito> <max>      # CI y evals sobre la cabeza → JSON {verde, ronda, reparar, …}
#   scripts/workflow/cierre.sh reparado <hito>         # tras reparar_cierre: verifica, commitea y cuenta la reparación → JSON
#   scripts/workflow/cierre.sh evals <hito>            # solo recoge los informes de evals de la última medición y su fila de la serie
#
# Una medición vale solo para el producto que midió (ADR 0030). `medir` la repite en
# cada vuelta del bucle, salvo en un caso: si desde la última nada ha cambiado fuera
# de gates/ y ninguna reparación ha terminado (gates/cierre-reparaciones, que cuenta
# `reparado`), la medición sigue siendo la de este producto y se reutiliza sin contar
# otra. Es lo que pasa al reanudar un run interrumpido entre la medición y el final de
# su reparación: se retoma la reparación en lugar de gastar una de las mediciones en
# medir otra vez lo mismo. Tras una reparación terminada se mide siempre, aunque no
# cambiara nada: un rojo de la plataforma (una ejecución cancelada, un runner caído)
# solo se aclara volviendo a medir. `reparar` dice si queda medición para la vuelta
# siguiente (ronda < <max>) y la medición es roja: es la condición del bucle.
#
# Tras medir, cada trabajo del flujo `evals` que terminó deja su informe.json en
# gates/evals/<skill>.json: el job lo imprime entero en su registro entre las marcas
# «--- inicio de informe.json ---» y «--- fin de informe.json ---» (scripts/evals.sh),
# y de ahí lo copia este paso, sin modelo. El informe final saca de esos ficheros la
# tasa de cada eval (ADR 0028). Con cada informe, si la skill tiene juez, anota la
# fila de la medición en gates/serie.jsonl, la serie de cierres del ADR 0038
# (scripts/workflow/serie.sh): cuentan todas las mediciones completas del cierre, no
# solo la última, y una fila que no casa con los umbrales de su informe detiene el
# cierre con un dossier. `evals` hace solo esa recogida y esa fila, sin empujar ni poner
# etiquetas, sobre las comprobaciones que ya están en gates/cierre.json: sirve para un
# run que midió con un workflow anterior.
#
# En H5 y H6 la ejecución de cierre de las evals era una tarea [plataforma] dentro
# del bucle, antes de la revisión final: cada corrección de la revisión la dejaba
# sin cubrir la cabeza y hubo que repetirla ronda tras ronda. Aquí va al final, en el
# mismo bucle que la revisión final (bucle_final): si sale en rojo, el reparador del
# cierre arregla, los dos jueces juzgan lo que cambió fuera de gates/ y se vuelve a
# medir (ADR 0030).
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

sub="${1:?uso: cierre.sh iniciar|publicar|medir|reparado|evals <hito>}"
hito="${2:?uso: cierre.sh iniciar|publicar|medir|reparado|evals <hito>}"
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
  local e="$d/gates/evals" fila nombre skill link run job
  rm -rf "$e"; mkdir -p "$e"
  jq -c '.checks[] | select(.workflow == "evals" and (.bucket == "pass" or .bucket == "fail"))' "$d/gates/cierre.json" |
  while IFS= read -r fila; do
    nombre=$(jq -r .name <<<"$fila"); link=$(jq -r .link <<<"$fila")
    skill=$(printf '%s' "$nombre" | sed -nE 's/^evals \(([a-z0-9-]+)\)$/\1/p')
    run=$(printf '%s' "$link" | sed -nE 's#.*/actions/runs/([0-9]+)/job/[0-9]+.*#\1#p')
    job=$(printf '%s' "$link" | sed -nE 's#.*/actions/runs/[0-9]+/job/([0-9]+).*#\1#p')
    [ -n "$skill" ] && [ -n "$run" ] && [ -n "$job" ] || continue
    informe_del_trabajo "$run" "$job" "$e/$skill.json" "$nombre" "$skill" </dev/null || true
  done
}

# Añade a gates/serie.jsonl, por cada trabajo de evals de gates/cierre.json cuya skill
# tiene juez en el commit medido, la fila de esta medición (ADR 0038, punto 6;
# scripts/workflow/serie.sh): también si la medición está incompleta o el trabajo no
# dejó informe, que la fila lo dice y no cuenta. Recoger otra vez la misma medición
# sustituye su fila. Sale con 1 si una fila no casa con los umbrales de su informe:
# entonces el recuento de la serie no es el del job, y no se escribe.
anotar_serie() {
  local fila nombre skill link run job sha ronda fecha propuesta p nueva informe
  sha=$(jq -r .sha "$d/gates/cierre.json"); ronda=$(jq -r .ronda "$d/gates/cierre.json"); fecha=$(jq -r '.fecha // ""' "$d/gates/cierre.json")
  propuesta=$(gh pr view "$rama" --json number --jq .number 2>/dev/null </dev/null || true)
  printf '%s' "$propuesta" | grep -qE '^[0-9]+$' || propuesta=null
  while IFS= read -r fila; do
    nombre=$(jq -r .name <<<"$fila"); link=$(jq -r .link <<<"$fila")
    skill=$(printf '%s' "$nombre" | sed -nE 's/^evals \(([a-z0-9-]+)\)$/\1/p')
    run=$(printf '%s' "$link" | sed -nE 's#.*/actions/runs/([0-9]+)/job/[0-9]+.*#\1#p')
    job=$(printf '%s' "$link" | sed -nE 's#.*/actions/runs/[0-9]+/job/([0-9]+).*#\1#p')
    [ -n "$skill" ] && [ -n "$run" ] && [ -n "$job" ] || continue
    git cat-file -e "$sha:evals/$skill/juez/clases.yaml" 2>/dev/null || continue
    p=$(jq -nc --arg h "$hito" --argjson m "$ronda" --arg f "$(jq -r --arg f "$fecha" '.startedAt // $f' <<<"$fila")" \
      --arg c "$sha" --argjson pr "$propuesta" --argjson e "$run" --argjson t "$job" \
      '{hito: $h, medicion: $m, fecha: $f, commit: $c, propuesta: $pr, ejecucion: $e, trabajo: $t}')
    informe="$d/gates/evals/$skill.json"; [ -f "$informe" ] || informe=""
    nueva=$(scripts/workflow/serie.sh fila "$skill" "$p" "$informe") || return 1
    scripts/workflow/serie.sh anotar "$d/gates/serie.jsonl" "$nueva" || return 1
    echo "evals: fila de $skill en $d/gates/serie.jsonl ($(jq -r 'if .cuenta then "cuenta" else "no cuenta: " + .no_cuenta_porque end' <<<"$nueva"))" >&2
  done < <(jq -c '.checks[] | select(.workflow == "evals" and (.bucket == "pass" or .bucket == "fail"))' "$d/gates/cierre.json")
}

case "$sub" in
  evals)
    [ -f "$d/gates/cierre.json" ] || { echo "evals: no hay gates/cierre.json; el cierre no ha medido" >&2; exit 1; }
    command -v gh >/dev/null 2>&1 && gh auth status >/dev/null 2>&1 || { echo "evals: gh no tiene sesión" >&2; exit 1; }
    recoger_evals
    anotar_serie || exit 1
    exit 0;;
  iniciar)
    rm -f "$d/gates/cierre-rondas" "$d/gates/cierre-reparaciones" "$d/gates/cierre.json"
    echo "cierre: mediciones y reparaciones a cero"; exit 0;;
  reparado)
    # Una reparación terminada, se haya commiteado, apartado o no haya cambiado nada:
    # la medición siguiente vuelve a medir aunque el producto sea el mismo.
    r=$(scripts/workflow/global.sh cerrar "$hito")
    echo $(( $(cat "$d/gates/cierre-reparaciones" 2>/dev/null || echo 0) + 1 )) > "$d/gates/cierre-reparaciones"
    printf '%s\n' "$r"; exit 0;;
esac

[ "$rama" != main ] || dossier "rama equivocada" "La rama actual es main; el cierre nunca empuja main."
command -v gh >/dev/null 2>&1 && gh auth status >/dev/null 2>&1 \
  || dossier "credencial ausente" "gh no está instalado o no tiene sesión (gh auth status). Sin ella no se puede publicar ni medir."

case "$sub" in
  publicar)
    # Los registros del run (supuestos, cuarentena, veredictos) viven en el directorio del
    # feature y algunos pasos los dejan sin commitear: se versionan antes de publicar.
    git add -A -- "$d"
    git diff --cached --quiet || git commit -q -m "docs($hito): registros del run" -- "$d"
    test -z "$(git status --porcelain)" || { git status --short >&2; dossier "árbol sucio" "Quedan cambios sin commitear fuera de $d antes de publicar."; }
    git push -u origin "$rama" >&2 || dossier "dependencia externa inaccesible" "git push -u origin $rama falló."
    if url=$(gh pr view "$rama" --json url --jq .url 2>/dev/null); then echo "propuesta existente: $url"; exit 0; fi
    titulo=$(titulo_del_hito "$hito")
    cuerpo="$d/gates/informe-final.md"
    if [ ! -f "$cuerpo" ]; then
      cuerpo=$(mktemp)
      printf 'Hito %s · %s\n\nEn curso: el workflow publica la rama para medir el cierre. El informe final sustituirá este cuerpo.\n' "$hito" "$titulo" > "$cuerpo"
    fi
    gh pr create --base main --head "$rama" --title "feat($hito): $titulo" --body-file "$cuerpo" >&2 \
      || dossier "dependencia externa inaccesible" "gh pr create falló."
    echo "propuesta abierta: $(gh pr view "$rama" --json url --jq .url)";;

  medir)
    max="${3:?uso: cierre.sh medir <hito> <max>}"
    reparaciones=$(cat "$d/gates/cierre-reparaciones" 2>/dev/null || echo 0)
    n=$(cat "$d/gates/cierre-rondas" 2>/dev/null || echo 0)
    if [ -f "$d/gates/cierre.json" ] && [ "$(jq -r '.reparaciones // -1' "$d/gates/cierre.json")" = "$reparaciones" ] \
       && producto_igual "$(jq -r '.sha // ""' "$d/gates/cierre.json")"; then
      echo "medir: el producto no ha cambiado desde la medición de $(jq -r '.sha[0:7]' "$d/gates/cierre.json") y ninguna reparación ha terminado; se reutiliza" >&2
      jq -c --argjson n "$n" --argjson max "$max" \
        '{sha, verde, ronda:$n, reutilizada:true, reparar:((.verde | not) and $n < $max),
          rojos:[.checks[] | select(.workflow != "" and (.bucket == "fail" or .bucket == "cancel")) | .name]}' "$d/gates/cierre.json"
      exit 0
    fi
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
    n=$((n + 1)); echo "$n" > "$d/gates/cierre-rondas"
    jq -n --arg sha "$sha" --argjson verde "$verde" --argjson checks "$checks" --arg fecha "$(date -u +%Y-%m-%dT%H:%M:%SZ)" \
      --argjson n "$n" --argjson rep "$reparaciones" \
      '{sha:$sha, verde:$verde, fecha:$fecha, ronda:$n, reparaciones:$rep, checks:$checks}' > "$d/gates/cierre.json"
    : > "$d/gates/cierre.log"
    for link in $(jq -r '.[].link' <<<"$rojos"); do
      run=$(printf '%s' "$link" | sed -nE 's#.*/actions/runs/([0-9]+).*#\1#p')
      [ -n "$run" ] || continue
      # gh run view --log antepone trabajo, paso y hora a cada línea: se quita con cut.
      { echo "== $link"; gh run view "$run" --log-failed 2>&1 | cut -f3- | tail -150; } >> "$d/gates/cierre.log"
    done
    recoger_evals
    anotar_serie || dossier "serie de cierres" "La fila de la serie de cierres de una skill no casa con los umbrales de su informe (scripts/workflow/serie.sh fila; la salida de error del paso dice cuál y en qué). El recuento de la serie no es el del job: hay que arreglar serie.sh antes de seguir."
    jq -n --arg sha "$sha" --argjson verde "$verde" --argjson rojos "$rojos" --argjson n "$n" --argjson max "$max" \
      '{sha:$sha, verde:$verde, ronda:$n, reutilizada:false, reparar:(($verde | not) and $n < $max), rojos:[$rojos[].name]}';;

  *) echo "subcomando desconocido: $sub" >&2; exit 2;;
esac
