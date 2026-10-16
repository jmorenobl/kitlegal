#!/usr/bin/env bash
# El modelo con el que corrió de verdad cada sesión con modelo del workflow `hito`.
#
#   scripts/workflow/modelos.sh empezar            # extraer_hito, el primer paso de un run: olvida el
#                                                  #   feature del run anterior
#   scripts/workflow/modelos.sh registrar <sesion> <paso> <rol> <pedido> <esfuerzo>
#                                                  # scripts/claude-modelo.sh, antes de abrir la sesión
#   scripts/workflow/modelos.sh recoger            # recoger_sesiones, tras specify: lleva al hito las
#                                                  #   sesiones apuntadas antes de que existiera
#   scripts/workflow/modelos.sh resolver [fichero] # rellena el modelo real de cada sesión registrada
#                                                  #   (en el registro o, con fichero, en una copia)
#   scripts/workflow/modelos.sh juez <paso> [desde] # → JSON {sesion, pedido, real, sin_modelo_real} de la
#                                                  #   última sesión del paso iniciada después de <desde>
#   scripts/workflow/modelos.sh asignado <rol>     # el modelo que workflow.yml da al rol, sin esfuerzo
#   scripts/workflow/modelos.sh distintos [fichero] # → JSON, una entrada por sesión cuyo modelo real no
#                                                  #   es el de su rol o no se sabe, con la causa
#
# Que un paso corrió con el modelo que workflow.yml asigna a su rol (ADR 0033) tiene
# que poder comprobarse. En H26 el supervisor bajó todos los roles de modelo por el
# límite de uso (fable → opus → sonnet) y el informe final lo calló: los dos jueces
# de la revisión final fueron Sonnet.
#
# El registro vive en <feature_dir>/gates/modelos.jsonl, una línea por sesión. La
# escribe scripts/claude-modelo.sh, la única puerta por la que el workflow y
# scripts/paso.sh abren una sesión, antes del `exec` de Claude Code: el paso y su rol
# (del sufijo #<paso>:<rol> que cada paso prompt de workflow.yml pone a su modelo,
# porque spec-kit no le dice al ejecutable qué paso es y los dos jueces del fan-out
# corren a la vez), el modelo y el esfuerzo pedidos, el id de la sesión (que el
# wrapper le fija con --session-id), de dónde viene (KITLEGAL_SESION_ORIGEN: el
# workflow o paso.sh, y KITLEGAL_MODELO_A_MANO si se pidió otro modelo a mano) y el
# repuesto que el supervisor aplicó por el límite de uso (KITLEGAL_MODELO_REPUESTO).
#
# El modelo real no sale de la configuración sino de lo que Claude Code escribe de la
# sesión: el transcript <config>/projects/*/<sesion>.jsonl, donde cada respuesta lleva
# en .message.model el id con el que contestó la API. Los mensajes «<synthetic>» los
# escribe Claude Code sin llamar al modelo (un error de la API, el límite de uso) y no
# cuentan; los subagentes van en su propio fichero. No se lee de la salida de la
# sesión porque spec-kit pide la de texto y `modelUsage` solo viene en la JSON, y el
# wrapper tiene que terminar en `exec`: spec-kit mata con SIGKILL lo que lanzó cuando
# vence el plazo de un paso, y una sesión hija del wrapper quedaría huérfana. Por eso
# resolver va aparte, cuando la sesión ya ha terminado: lo llaman `gate.sh leer
# revision`, tras los jueces, e informe.sh. Si no hay transcript o no trae ninguna
# respuesta del modelo, `real` queda null y `sin_modelo_real` dice por qué: nunca se
# deduce.
#
# La sesión de specify se abre antes de que exista el directorio del hito: es ella
# quien lo crea y escribe .specify/feature.json. Ese fichero no se versiona y, entre
# un run y el siguiente, sigue apuntando al feature anterior; en H23.1 la línea de
# specify fue a parar al registro de H26 y entró en la rama del hito. Por eso
# extraer_hito, el primer paso de un run, lo borra con `empezar`; mientras no hay
# feature.json, el registro es .specify/sesiones-sin-feature.jsonl (sin versionar),
# y recoger_sesiones, el paso que sigue a specify, lleva esas líneas al registro
# del hito. Un clon sin feature.json tampoco pierde la sesión.
#
# Nada de esto lee credenciales, abre sesiones ni escribe fuera del repositorio
# (ADR 0032): lee el transcript de la propia sesión y escribe en gates/ y en .specify/.
set -euo pipefail
cd "$(dirname "$0")/../.."
. scripts/workflow/comun.sh

sub="${1:?uso: modelos.sh empezar|registrar|recoger|resolver|juez|asignado|distintos …}"
feature=.specify/feature.json
sin_feature=.specify/sesiones-sin-feature.jsonl
registro() {
  if [ -f "$feature" ]; then printf '%s/gates/modelos.jsonl\n' "$(feature_dir)"
  else printf '%s\n' "$sin_feature"; fi
}

# El modelo que workflow.yml da a un rol (input modelo_<rol>), sin el esfuerzo.
asignado() { # $1 = rol
  awk -v rol="$1" '
    $0 ~ "^  " rol ":[[:space:]]*$" { dentro = 1; next }
    dentro && /^  [a-z_]+:/ { exit }
    dentro && /^    default:/ { v = $0; sub(/^    default:[[:space:]]*/, "", v); gsub(/"/, "", v); sub(/@.*/, "", v); print v; exit }
  ' .specify/workflows/hito/workflow.yml
}

# Lo que dice el transcript de una sesión: {real, sin_modelo_real}.
del_transcript() { # $1 = sesion
  local f
  f=$(ls "${CLAUDE_CONFIG_DIR:-$HOME/.claude}"/projects/*/"$1".jsonl 2>/dev/null | head -1 || true)
  if [ -z "$f" ]; then
    jq -n --arg s "$1" '{real: null, sin_modelo_real: ("no está el transcript de la sesión " + $s)}'
    return
  fi
  jq -R 'fromjson? | select(.type == "assistant") | .message.model // empty' "$f" | jq -s '
    [.[] | select(. != "<synthetic>")] as $m
    | if ($m | length) > 0 then {real: ($m | group_by(.) | map({(.[0]): length}) | add), sin_modelo_real: null}
      else {real: null, sin_modelo_real: (if length > 0 then "la sesión no trae ninguna respuesta del modelo, solo \(length) mensajes de Claude Code sin modelo (un error de la API o el límite de uso)"
                                          else "la sesión no trae ninguna respuesta del modelo" end)} end'
}

case "$sub" in
  empezar)
    rm -f -- "$feature" "$sin_feature";;

  recoger)
    [ -f "$feature" ] || { echo "modelos.sh recoger: no hay $feature: specify no ha creado el directorio del hito" >&2; exit 1; }
    [ -f "$sin_feature" ] || exit 0
    r=$(registro); [ -d "$(dirname "$r")" ] || mkdir -p "$(dirname "$r")"
    cat "$sin_feature" >> "$r"; rm -f -- "$sin_feature";;

  registrar)
    sesion="${2-}"; paso="${3-}"; rol="${4-}"; pedido="${5:?uso: modelos.sh registrar <sesion> <paso> <rol> <pedido> <esfuerzo>}"; esfuerzo="${6-}"
    r=$(registro); [ -d "$(dirname "$r")" ] || mkdir -p "$(dirname "$r")"
    jq -cn --arg sesion "$sesion" --arg paso "$paso" --arg rol "$rol" --arg pedido "$pedido" --arg esfuerzo "$esfuerzo" \
      --arg origen "${KITLEGAL_SESION_ORIGEN:-workflow}" --arg a_mano "${KITLEGAL_MODELO_A_MANO:-}" \
      --arg repuesto "${KITLEGAL_MODELO_REPUESTO:-}" --arg inicio "$(date -u +%Y-%m-%dT%H:%M:%SZ)" '
      def n: if . == "" then null else . end;
      {sesion: ($sesion | n), paso: ($paso | n), rol: ($rol | n), pedido: $pedido, esfuerzo: ($esfuerzo | n),
       origen: $origen, a_mano: ($a_mano != ""), repuesto: ($repuesto | n), inicio: $inicio,
       real: null, sin_modelo_real: (if $sesion == "" then "la sesión no tiene un id propio (se reanudó otra)" else "sin resolver" end)}' >> "$r";;

  resolver)
    r=$(registro); [ -f "$r" ] || exit 0
    destino="${2:-$r}"; tmp="$destino.tmp"; : > "$tmp"
    while IFS= read -r l; do
      [ -n "$l" ] || continue
      s=$(jq -r '.sesion // ""' <<<"$l")
      if [ -n "$s" ]; then
        t=$(del_transcript "$s")
        # Un transcript que ya no está (Claude Code los borra con el tiempo) no deshace lo
        # que se resolvió cuando estaba.
        if [ "$(jq -r '.real == null' <<<"$t")" = true ] && [ "$(jq -r '.real != null' <<<"$l")" = true ]; then
          printf '%s\n' "$l" >> "$tmp"
        else
          jq -c --argjson t "$t" '. + $t' <<<"$l" >> "$tmp"
        fi
      else
        printf '%s\n' "$l" >> "$tmp"
      fi
    done < "$r"
    mv "$tmp" "$destino";;

  juez)
    paso="${2:?uso: modelos.sh juez <paso> [desde]}"; desde="${3-}"; r=$(registro)
    if [ -f "$r" ]; then cat "$r"; fi | jq -s -c --arg p "$paso" --arg d "$desde" '
      [.[] | select(.paso == $p and .inicio > $d)] | last
      | if . == null then {sesion: null, pedido: null, real: null,
          sin_modelo_real: ("ninguna sesión de " + $p + " registrada" + (if $d == "" then "" else " desde " + $d end) + ": el veredicto no lo escribió una sesión del workflow ni de scripts/paso.sh")}
        else {sesion, pedido, real, sin_modelo_real} end';;

  asignado)
    asignado "${2:?uso: modelos.sh asignado <rol>}";;

  distintos)
    r="${2:-$(registro)}"
    [ -f "$r" ] || { echo '[]'; exit 0; }
    asignados=$(for rol in $(jq -r '.rol // empty' "$r" | sort -u); do jq -n --arg r "$rol" --arg m "$(asignado "$rol")" '{($r): $m}'; done | jq -s 'add // {}')
    jq -s -c --argjson asig "$asignados" '
      [.[] | ($asig[.rol // ""] // "") as $a | . + {asignado: (if $a == "" then null else $a end)}
       | select(.asignado == null or .real == null or ((.real | keys) - [.asignado] | length) > 0)
       | . + {causa: (
           if .asignado == null then "la sesión no dice su paso ni su rol (se abrió sin el sufijo #<paso>:<rol> en el modelo)"
           elif .real == null then "no se sabe con qué modelo corrió: " + .sin_modelo_real
           elif .repuesto != null then "el supervisor había cambiado de modelo por el límite de uso (" + .repuesto + ")"
           elif .origen == "paso.sh" and .a_mano then "se pidió a mano con scripts/paso.sh (" + .pedido + ")"
           elif (.pedido | sub("@.*"; "")) != .asignado and .origen != "paso.sh" then "el run se lanzó o se reanudó con otro modelo para el rol (" + .pedido + "; KITLEGAL_MODELO_<ROL> o --resume " + .rol + "=…)"
           else "Claude Code respondió con otro modelo que el pedido (" + .pedido + ")" end)}]' "$r";;

  *) echo "subcomando desconocido: $sub" >&2; exit 2;;
esac
