#!/usr/bin/env bash
# Cierre del hito en la plataforma remota (ADR 0018): lo hace el workflow, sin modelo,
# DESPUÉS de la revisión final, sobre la cabeza que se va a fusionar.
#
#   scripts/workflow/cierre.sh iniciar <hito>          # antes del bucle de revisión y cierre: contadores a cero
#   scripts/workflow/cierre.sh publicar <hito>         # empuja la rama y abre la propuesta si no existe
#   scripts/workflow/cierre.sh medir <hito> <max>      # CI y evals sobre la cabeza → JSON {verde, ronda, reparar, …}
#   scripts/workflow/cierre.sh reparado <hito>         # tras reparar_cierre: verifica, commitea y cuenta la reparación → JSON
#   scripts/workflow/cierre.sh evals <hito>            # solo recoge los informes de evals de la última medición
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
# tasa de cada eval (ADR 0028). `evals` hace solo esa recogida, sin empujar ni poner
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
  local e="$d/gates/evals" fila nombre skill link run job destino ajenas
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
    # GitHub guarda el registro en trozos y cada uno empieza por la marca de orden de bytes
    # (EF BB BF), que queda delante de la hora: un informe largo cruza un trozo, y con la
    # marca delante la hora no se quitaba y el informe dejaba de ser JSON (H21: el de
    # boe-legislacion en dos modos, 821 kB, quedó fuera del informe final).
    gh run view "$run" --job "$job" --log 2>/dev/null | cut -f3- | sed $'s/^\xef\xbb\xbf//' | sed -E 's/^[0-9-]+T[0-9:.]+Z //' \
      | awk '$0 == "--- inicio de informe.json ---" {p = 1; next} $0 == "--- fin de informe.json ---" {p = 0} p' > "$destino.bruto" || true
    # Entre las dos marcas puede caer una línea que no es del informe. El registro junta la salida
    # estándar y la de error del paso, que el runner lee por separado, y cuando el trabajo falla el
    # «make: *** [Makefile:…: evals] Error 1» de la salida de error queda en un punto cualquiera de
    # lo que la estándar aún no había entregado (H25, medición 1: cayó dentro del informe de
    # boe-legislacion, que dejó de ser JSON, y la reparación del cierre trabajó sin él). El informe
    # lo escribe el job con sangría: cada línea suya empieza por un espacio o por una llave, y la
    # que no, es de otra salida. Se descarta y se dice cuántas.
    ajenas=$(grep -c -v '^[ {}]' "$destino.bruto" || true)
    grep '^[ {}]' "$destino.bruto" > "$destino" || true
    rm -f "$destino.bruto"
    if jq -e '.skill and (.tasas | type == "array")' "$destino" >/dev/null 2>&1; then
      echo "evals: informe de $skill en $destino" >&2
      [ "${ajenas:-0}" -eq 0 ] || echo "evals: el registro de $nombre traía dentro del informe líneas de otra salida, descartadas: $ajenas" >&2
    else
      rm -f "$destino"; echo "evals: el registro de $nombre no trae un informe.json legible" >&2
    fi
  done
}

case "$sub" in
  evals)
    [ -f "$d/gates/cierre.json" ] || { echo "evals: no hay gates/cierre.json; el cierre no ha medido" >&2; exit 1; }
    command -v gh >/dev/null 2>&1 && gh auth status >/dev/null 2>&1 || { echo "evals: gh no tiene sesión" >&2; exit 1; }
    recoger_evals; exit 0;;
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
    jq -n --arg sha "$sha" --argjson verde "$verde" --argjson rojos "$rojos" --argjson n "$n" --argjson max "$max" \
      '{sha:$sha, verde:$verde, ronda:$n, reutilizada:false, reparar:(($verde | not) and $n < $max), rojos:[$rojos[].name]}';;

  *) echo "subcomando desconocido: $sub" >&2; exit 2;;
esac
