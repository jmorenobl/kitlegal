#!/usr/bin/env bash
# Rondas juez → corrector del workflow `hito` (docs/WORKFLOW.md «Tres capas de gate»).
#
#   scripts/workflow/gate.sh iniciar spec|plan|tasks|revision   # antes del bucle: pone a cero las rondas
#   scripts/workflow/gate.sh cambios spec|plan|tasks             # al empezar cada ronda: diff de la corrección anterior
#   scripts/workflow/gate.sh leer spec|plan|tasks <max>          # tras el juez → JSON
#   scripts/workflow/gate.sh leer revision <max>                 # tras los dos jueces finales → JSON
#   scripts/workflow/gate.sh cerrar spec|plan|tasks|revision <hito>
#
# leer valida el veredicto, le suma los defectos del precheck mecánico (un defecto
# mecánico fuerza el rechazo), archiva la ronda en gates/<fase>-r<n>.json y decide
# `seguir`: otra ronda mientras no esté aprobado y queden rondas. Un veredicto
# ausente o mal formado no para nada: la ronda se repite con un juez nuevo, sin
# corrector (`invalido`). Las `observaciones` del veredicto (lo que el juez vio y no
# pasa el umbral de materialidad ni el criterio de uso, ADR 0028) no cuentan para
# nada: se archivan con la ronda y no llegan al corrector.
#
# Convergencia (ADR 0028): desde la segunda ronda, un motivo nuevo solo cuenta si lo
# introdujo la corrección anterior o si es material. Para que el juez lo sepa sin
# fiarse de nadie, leer guarda la instantánea de los artefactos que juzgó (fuera del
# historial, en el git-dir) y cambios, al empezar la ronda siguiente, escribe en
# gates/<fase>-correccion-r<n>.diff lo que el corrector cambió desde ella. En la
# revisión final no hace falta: cada corrección es un commit.
#
# Ningún juez para el run (ADR 0018). La única excepción es el juez del spec, que
# puede RECHAZAR LA ENTRADA con una lista cerrada de motivos (`entrada`:
# contradiccion, sin_criterio_comprobable, objetivo_vacio): el spec no permite
# derivar tests y adivinar sería peor que devolverlo. cerrar escribe entonces
# gates/informe-rechazo.md y sale con 1; el supervisor lo reconoce como rechazo.
# En cualquier otro caso, si las rondas se agotan sin aprobar, los motivos que
# quedan van a gates/<fase>-pendiente.md y al informe final, y el run sigue.
set -euo pipefail
cd "$(dirname "$0")/../.."
. scripts/workflow/comun.sh

accion="${1:?uso: gate.sh iniciar|leer|cerrar <fase> …}"
fase="${2:?uso: gate.sh iniciar|leer|cerrar <fase> …}"
d=$(feature_dir); g="$d/gates"; mkdir -p "$g"
rondas="$g/$fase-rondas"
tipos_entrada='["contradiccion","sin_criterio_comprobable","objetivo_vacio"]'

# Valida un veredicto; imprime el motivo si no vale.
invalido() {
  [ -f "$1" ] || { echo "falta $1"; return 0; }
  jq -e 'has("veredicto") and (.criterios | type == "array" and length > 0) and (.motivos | type == "array")
         and ((has("observaciones") | not) or (.observaciones | type == "array"))' "$1" >/dev/null 2>&1 \
    || { echo "formato inválido en $1"; return 0; }
  [ "$(jq -r '(.criterios | all(.cumple == true)) == (.veredicto == "aprobado")' "$1")" = true ] \
    || { echo "veredicto incoherente con sus criterios en $1"; return 0; }
  return 1
}

siguiente_ronda() { local n; n=$(cat "$rondas" 2>/dev/null || echo 0); n=$((n + 1)); echo "$n" > "$rondas"; echo "$n"; }

# Instantánea de lo que juzgó la ronda $1: el directorio del feature sin gates/ ni
# aceptacion/, que ningún corrector de artefactos toca.
instantanea() {
  local snap
  snap="$(dir_instantaneas)/$fase-r$1"
  rm -rf "$snap"; mkdir -p "$snap"
  cp -R "$d/." "$snap/"
  rm -rf "$snap/gates" "$snap/aceptacion"
}

case "$accion" in
  iniciar)
    rm -f "$rondas" "$g/$fase"-correccion-r*.diff
    rm -rf "$(dir_instantaneas)/$fase"-r*
    echo "rondas de $fase a cero";;

  cambios)
    # Lo que cambió la corrección de la última ronda juzgada: diff entre su instantánea
    # y el directorio del feature de ahora. Sin ronda anterior (la primera), nada.
    n=$(cat "$rondas" 2>/dev/null || echo 0)
    snap="$(dir_instantaneas)/$fase-r$n"
    if [ "$n" -eq 0 ] || [ ! -d "$snap" ]; then echo "$fase: primera ronda, sin corrección anterior"; exit 0; fi
    salida="$g/$fase-correccion-r$n.diff"
    # diff sale con 1 cuando hay diferencias: no es un fallo. Las rutas se reescriben
    # para que el fichero no dependa del git-dir ni del directorio del feature.
    { diff -ruN -x gates -x aceptacion "$snap" "$d" || [ $? -eq 1 ]; } \
      | sed -e "s#^\(--- \)$snap/#\1antes/#" -e "s#^\(+++ \)$d/#\1ahora/#" -e "s#^diff -ruN -x gates -x aceptacion $snap/\([^ ]*\) .*#diff antes/\1 ahora/\1#" > "$salida"
    echo "$fase: corrección de la ronda $n en $salida ($(awk '/^[+-]/ && !/^(\+\+\+ ahora|--- antes)\// {n++} END {print n + 0}' "$salida") líneas cambiadas)";;

  leer)
    max="${3:?uso: gate.sh leer <fase> <max>}"
    n=$(siguiente_ronda)
    if [ "$fase" = revision ]; then
      a="$g/revision-a.json"; b="$g/revision-b.json"
      ma=$(invalido "$a" || true); mb=$(invalido "$b" || true)
      [ -z "$ma" ] && cp "$a" "$g/revision-a-r$n.json"
      [ -z "$mb" ] && cp "$b" "$g/revision-b-r$n.json"
      if [ -n "$ma$mb" ]; then
        jq -n --argjson n "$n" --argjson max "$max" --arg m "$ma $mb" \
          '{estado:"invalido", ronda:$n, invalido:true, seguir:($n < $max), corregir:false, motivos:["veredicto inválido: " + $m]}'
        exit 0
      fi
      va=$(jq -r .veredicto "$a"); vb=$(jq -r .veredicto "$b")
      # aprobado: los dos aprueban. Si no, la unión de sus motivos va al corrector: el
      # desacuerdo entre jueces mide a los jueces, no al código (ADR 0007).
      estado=corregir; [ "$va" = aprobado ] && [ "$vb" = aprobado ] && estado=aprobado
      jq -n --arg estado "$estado" --arg va "$va" --arg vb "$vb" --argjson n "$n" --argjson max "$max" \
        --slurpfile a "$a" --slurpfile b "$b" \
        '{estado:$estado, juez_a:$va, juez_b:$vb, ronda:$n, invalido:false,
          seguir:($estado != "aprobado" and $n < $max), corregir:($estado != "aprobado" and $n < $max),
          motivos:(($a[0].motivos // []) + ($b[0].motivos // []))}'
      exit 0
    fi
    v="$g/$fase.json"
    instantanea "$n" # también con un veredicto inválido: la ronda siguiente verá un diff vacío, que es la verdad
    m=$(invalido "$v" || true)
    if [ -n "$m" ]; then
      jq -n --argjson n "$n" --argjson max "$max" --arg m "$m" \
        '{veredicto:"invalido", ronda:$n, invalido:true, rechazo_entrada:false, seguir:($n < $max), corregir:false, motivos:["veredicto inválido: " + $m]}'
      exit 0
    fi
    cp "$v" "$g/$fase-r$n.json" # el juez sobrescribe <fase>.json en cada ronda; la copia conserva el historial
    mecanicos=$(cat "$g/$fase-precheck.txt" 2>/dev/null || true)
    entrada=false
    if [ "$fase" = spec ] && jq -e --argjson t "$tipos_entrada" '(.entrada // []) | length > 0 and all(.[]; (.tipo as $x | $t | index($x)) != null)' "$v" >/dev/null 2>&1; then
      entrada=true
    fi
    jq --argjson n "$n" --argjson max "$max" --arg mec "$mecanicos" --argjson entrada "$entrada" '
      ([$mec | split("\n")[] | select(length > 0) | "[mecánico] " + .]) as $m
      | (if ($m | length) > 0 then "rechazado" else .veredicto end) as $ver
      | {veredicto:$ver, ronda:$n, invalido:false, rechazo_entrada:$entrada,
         seguir:($ver != "aprobado" and ($entrada | not) and $n < $max),
         corregir:($ver != "aprobado" and ($entrada | not) and $n < $max),
         motivos:(.motivos + $m)}' "$v";;

  cerrar)
    hito="${3:?uso: gate.sh cerrar <fase> <hito>}"
    if [ "$fase" = revision ]; then
      # Los jueces dejan sus veredictos sin commitear; se versionan aquí con su historial.
      vs=""; for f in "$g"/revision-*.json "$rondas"; do [ -f "$f" ] && vs="$vs $f"; done
      if [ -n "$vs" ]; then
        git add -- $vs
        git diff --cached --quiet -- $vs || git commit -q -m "docs($hito): veredictos de la revisión final" -- $vs
      fi
      va=$(jq -r .veredicto "$g/revision-a.json" 2>/dev/null || echo "sin veredicto")
      vb=$(jq -r .veredicto "$g/revision-b.json" 2>/dev/null || echo "sin veredicto")
      if [ "$va" = aprobado ] && [ "$vb" = aprobado ]; then echo "revisión final aprobada por los dos jueces"; exit 0; fi
      { printf '# Motivos de la revisión final sin resolver al agotar las rondas\n\nJuez A: %s. Juez B: %s.\n\n' "$va" "$vb"
        jq -r '.motivos[]? | "- " + .' "$g/revision-a.json" "$g/revision-b.json" 2>/dev/null; } > "$g/revision-pendiente.md"
      anotar_supuesto "La revisión final agotó sus rondas sin la aprobación de los dos jueces (A: $va, B: $vb); motivos en $g/revision-pendiente.md."
      echo "revisión final sin aprobar tras $(cat "$rondas" 2>/dev/null) rondas: pendientes en $g/revision-pendiente.md; el run sigue"
      exit 0
    fi
    v="$g/$fase.json"
    if [ "$fase" = spec ] && jq -e --argjson t "$tipos_entrada" '(.entrada // []) | length > 0 and all(.[]; (.tipo as $x | $t | index($x)) != null)' "$v" >/dev/null 2>&1; then
      {
        printf '# Informe de rechazo del spec del hito %s\n\n' "$hito"
        printf 'El juez de entrada no puede aprobar el spec: con lo que dice la sección del hito no se pueden derivar\n'
        printf 'tests verificables, y adivinar sería peor que devolverlo. Corrige la sección `#### %s ·` de\n' "$hito"
        printf 'docs/ROADMAP.md respondiendo a cada pregunta y relanza el hito desde main.\n\n'
        jq -r '.entrada[] | "## \(.tipo) · \(.requisito // "sin requisito")\n\n- **Fragmento**: \(.fragmento // "—")\n- **Por qué impide derivar un test**: \(.por_que // "—")\n- **Pregunta**: \(.pregunta // "—")\n"' "$v"
      } > "$g/informe-rechazo.md"
      echo "SPEC RECHAZADO EN LA ENTRADA: informe en $g/informe-rechazo.md" >&2
      exit 1
    fi
    if [ -f "$v" ] && jq -e '.veredicto == "aprobado"' "$v" >/dev/null && [ ! -s "$g/$fase-precheck.txt" ]; then
      echo "gate_$fase aprobado en $(cat "$rondas" 2>/dev/null || echo '?') rondas"; exit 0
    fi
    { printf '# Motivos del gate de %s sin resolver al agotar las rondas\n\n' "$fase"
      jq -r '.motivos[]? | "- " + .' "$v" 2>/dev/null || true
      sed 's/^/- [mecánico] /' "$g/$fase-precheck.txt" 2>/dev/null || true; } > "$g/$fase-pendiente.md"
    anotar_supuesto "El gate de $fase agotó sus rondas sin aprobar; el run sigue con la última versión y los motivos en $g/$fase-pendiente.md."
    echo "gate_$fase sin aprobar tras $(cat "$rondas" 2>/dev/null) rondas: pendientes en $g/$fase-pendiente.md; el run sigue";;

  *) echo "acción desconocida: $accion" >&2; exit 2;;
esac
