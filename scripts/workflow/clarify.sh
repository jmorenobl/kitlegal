#!/usr/bin/env bash
# Clarificación sin persona del workflow `hito` (docs/WORKFLOW.md «Clarificación sin sesgo»).
#
#   scripts/workflow/clarify.sh iniciar      # antes de la ronda: pone a cero los intentos
#   scripts/workflow/clarify.sh comprobar    # tras resolver_clarify → JSON {completo, n, ronda}
#   scripts/workflow/clarify.sh completar    # tras la ronda: cierra de forma conservadora lo que falte
#
# Las preguntas las formula un proceso y las responde otro, con contexto limpio y
# el criterio de decisión autónoma de la constitución. Nada se escala a una
# persona: una ambigüedad que toca alcance, frontera humana, privacidad, términos
# de uso, anomalías o una decisión cerrada se cierra con la lectura conservadora
# (no se implementa) y queda marcada `conservadora` para el informe final
# (ADR 0018). Si tras los intentos alguna pregunta sigue sin respuesta válida,
# `completar` la cierra igual: fuera de alcance.
set -euo pipefail
cd "$(dirname "$0")/../.."
. scripts/workflow/comun.sh

d=$(feature_dir); g="$d/gates"; mkdir -p "$g"
p="$g/clarify-preguntas.json"; r="$g/clarify-respuestas.json"; c="$g/clarify-rondas"

case "${1:?uso: clarify.sh iniciar | comprobar | completar}" in
  iniciar) rm -f "$c"; echo "intentos de clarify a cero";;

  comprobar)
    n=$(( $(cat "$c" 2>/dev/null || echo 0) + 1 )); echo "$n" > "$c"
    [ -f "$p" ] || echo '{"preguntas":[]}' > "$p"
    total=$(jq '.preguntas | length' "$p")
    faltan=""
    if [ -f "$r" ] && jq -e '.respuestas | type == "array"' "$r" >/dev/null 2>&1; then
      faltan=$(jq -r --slurpfile r "$r" '([.preguntas[].id] - [$r[0].respuestas[] | select((.respuesta // "") != "") | .id]) | join(" ")' "$p")
    elif [ "$total" -gt 0 ]; then
      faltan=$(jq -r '[.preguntas[].id] | join(" ")' "$p")
    fi
    [ -z "$faltan" ] || echo "preguntas sin respuesta válida: $faltan" >&2
    jq -n --argjson n "$total" --argjson ronda "$n" --arg faltan "$faltan" '{completo:($faltan == ""), n:$n, ronda:$ronda, faltan:$faltan}';;

  completar)
    [ -f "$p" ] || echo '{"preguntas":[]}' > "$p"
    { [ -f "$r" ] && jq -e '.respuestas | type == "array"' "$r" >/dev/null 2>&1; } || echo '{"respuestas":[]}' > "$r"
    # Una respuesta marcada escalar (formato anterior a ADR 0018) se trata como conservadora.
    jq --slurpfile p "$p" '
      .respuestas |= map(if .escalar == true then . + {conservadora:true} | del(.escalar) else del(.escalar) end)
      | ([.respuestas[] | select((.respuesta // "") != "") | .id]) as $ok
      | .respuestas = ([.respuestas[] | select((.respuesta // "") != "")]
          + [$p[0].preguntas[] | select(.id as $i | $ok | index($i) | not)
             | {id, opcion:"", respuesta:"fuera de alcance: no se implementa (sin respuesta válida del proceso de decisión)",
                criterio:"b", fuente:"workflow hito (clarify.sh completar)", alternativa_rechazada:"", conservadora:true}])' \
      "$r" > "$r.tmp" && mv "$r.tmp" "$r"
    # El informe final lista estas respuestas desde el propio JSON (sección «Supuestos»).
    jq '{n: (.respuestas | length), conservadoras: ([.respuestas[] | select(.conservadora == true)] | length)}' "$r";;

  *) echo "subcomando desconocido: $1" >&2; exit 2;;
esac
