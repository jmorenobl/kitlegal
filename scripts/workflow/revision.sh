#!/usr/bin/env bash
# Ciclos de la revisión final dentro del bucle de revisión y cierre (ADR 0030).
#
#   scripts/workflow/revision.sh pendiente <hito>   # al empezar cada vuelta del bucle → JSON {revisar, …}
#   scripts/workflow/revision.sh cambios <hito>     # al preparar a los jueces: qué juzga esta ronda
#   scripts/workflow/revision.sh rejuzgar <hito> "<motivo>"   # a mano: volver a juzgar el hito entero
#
# El cierre puede cambiar el producto: desde el ADR 0029 un umbral del job de evals
# decide, y `reparar_cierre` toca código, la skill o los artefactos para ponerlo en
# verde. Por eso la revisión final vuelve a juzgar todo lo que cambia fuera de gates/
# después de su último veredicto, antes de volver a medir (el caso que lo motivó, en
# el ADR 0030).
#
# pendiente decide si esta vuelta del bucle juzga algo, solo con el estado del disco
# (así una reanudación retoma en su sitio, porque el motor vuelve a ejecutar el bucle
# entero, docs/WORKFLOW.md «Limitaciones conocidas»):
#   1. cierra el paso con modelo que una interrupción dejó a medias (global.sh retomar);
#   2. con un ciclo abierto en gates/revision-ciclo.json (una reanudación a mitad de una
#      ronda), lo continúa con las rondas que le quedan;
#   3. sin ningún veredicto todavía, abre el primer ciclo: la revisión final del hito;
#   4. si algo cambió fuera de gates/ desde la cabeza que juzgó el último veredicto
#      (commits o cambios sin commitear), abre un ciclo nuevo para juzgarlo;
#   5. si no, no hay nada que juzgar.
#
# cambios escribe gates/revision-cambios.md —lo que juzga la ronda: el diff entero en
# la primera, o el rango desde el último veredicto, con cada commit, el paso que lo hizo
# y sus ficheros— y el diff de ese rango fuera de gates/ en
# gates/revision-cambios-r<n>.diff. Lo lee cada juez para su regla de convergencia, y
# el corrector de la revisión. Nunca para el run.
#
# rejuzgar abre, con el ciclo anterior cerrado, un ciclo `entero` que vuelve a juzgar el
# hito entero, sin veredicto anterior, por el motivo que da la persona. En H26 los dos
# asientos de la revisión final los juzgó Sonnet porque el supervisor había bajado de
# modelo por el límite de uso; para repetirla con los modelos del workflow, `cambios` daba
# un rango vacío desde el último veredicto y cada juez leía como suyo el del otro modelo, y
# hubo que hacerlo desde un clon. Lo que rejuzgar hace:
#   · se niega con un run vivo, con un ciclo abierto (pendiente ya lo continúa) y sin ningún
#     veredicto (eso es la revisión normal);
#   · aparta revision-a.json, revision-b.json y revision-pendiente.md, para que la regla de
#     CONVERGENCIA de los jueces, que se aplica si su veredicto existe, no se aplique. No se
#     pierde ninguno: si no hay ya una copia idéntica revision-<a|b>-r<n>.json, se archiva
#     como revision-<a|b>-sustituido-c<ciclo>.json (y revision-pendiente-sustituido-c<ciclo>.md),
#     que el informe final enseña como sustituidos;
#   · abre el ciclo en revision-ciclo.json con entero: true, el motivo y desde: null, y lo
#     anota como supuesto.
# Después se juzga como cualquier ciclo abierto: reanudando el run (pendiente lo continúa,
# con los modelos de su inputs.json, y la medición del cierre se reutiliza si el producto no
# cambia) o, con un run ya completado, que spec-kit no reanuda, con scripts/paso.sh
# revision_juez_a|b, leer_revision, cerrar_revision e informe_final (docs/WORKFLOW.md).
set -euo pipefail
cd "$(dirname "$0")/../.."
. scripts/workflow/comun.sh

sub="${1:?uso: revision.sh pendiente|cambios|rejuzgar <hito> [motivo]}"
hito="${2:?uso: revision.sh pendiente|cambios|rejuzgar <hito> [motivo]}"
d=$(feature_dir); g="$d/gates"; mkdir -p "$g"
f_ciclo="$g/revision-ciclo.json"
rondas_hechas=$(cat "$g/revision-rondas" 2>/dev/null || echo 0)

# El paso del workflow que hizo un commit, por su mensaje (los que escribe el propio workflow).
paso_de() {
  case "$1" in
    "fix($hito): motivos de la revisión final"*) echo "paso \`corrector_revision\`";;
    "fix($hito): cierre en la plataforma"*) echo "paso \`reparar_cierre\`, tras una medición del cierre en rojo";;
    "fix($hito): make ci del hito completo"*) echo "paso \`reparar_hito\`";;
    "docs($hito): barrido"*) echo "paso \`barrido\`";;
    "docs($hito): registros del run"*) echo "pasos \`publicar\` o \`medir_cierre\` (registros del run)";;
    "docs($hito): informe final"*) echo "paso \`informe_final\`";;
    *) echo "fuera de los pasos del workflow";;
  esac
}

# El ciclo que sigue al último conocido: el del fichero o, en un run anterior al workflow
# 2.3.0, el de las rondas ya juzgadas (todas del primer ciclo).
ciclo_siguiente() {
  local c
  c=$( { jq -r '.ciclo // 0' "$f_ciclo" 2>/dev/null || echo 0; rondas_juzgadas "$hito" | cut -d' ' -f2; } | sort -n | tail -1)
  echo $((c + 1))
}

case "$sub" in
  rejuzgar)
    motivo="${3:-}"
    [ -n "$motivo" ] || { echo "uso: revision.sh rejuzgar <hito> \"<motivo>\": el motivo va al informe final" >&2; exit 2; }
    scripts/workflow/sesion-unica.sh comprobar
    if [ -f "$f_ciclo" ] && [ "$(jq -r '.cerrado // false' "$f_ciclo")" != true ]; then
      echo "rejuzgar: el ciclo $(jq -r '.ciclo' "$f_ciclo") de la revisión final sigue abierto ($f_ciclo): no se abre otro; reanuda el run o lanza los jueces con scripts/paso.sh, que lo continúan" >&2
      exit 1
    fi
    if [ -z "$(rondas_juzgadas "$hito")" ]; then
      echo "rejuzgar: la revisión final de $hito no tiene ningún veredicto todavía; no hay nada que volver a juzgar (la primera revisión ya juzga el hito entero)" >&2
      exit 1
    fi
    ciclo=$(ciclo_siguiente)
    apartados=""
    for f in "$g/revision-a.json" "$g/revision-b.json" "$g/revision-pendiente.md"; do
      [ -f "$f" ] || continue
      base=$(basename "$f"); copia=""
      case "$base" in
        revision-pendiente.md) destino="$g/revision-pendiente-sustituido-c$ciclo.md";;
        *) destino="$g/${base%.json}-sustituido-c$ciclo.json"
           for r in "$g/${base%.json}"-r*.json; do
             [ -f "$r" ] && cmp -s "$f" "$r" && { copia="$r"; break; }
           done;;
      esac
      if [ -n "$copia" ]; then
        rm -f "$f"; apartados="$apartados, $base (ya archivado como $(basename "$copia"))"
      else
        mv "$f" "$destino"; apartados="$apartados, $base → $(basename "$destino")"
      fi
    done
    jq -n --argjson c "$ciclo" --argjson desde_ronda "$((rondas_hechas + 1))" --arg motivo "$motivo" \
      '{ciclo:$c, desde_ronda:$desde_ronda, motivo:$motivo, entero:true, desde:null, commits:[], cerrado:false}' > "$f_ciclo"
    anotar_supuesto "La revisión final abre el ciclo $ciclo, desde la ronda $((rondas_hechas + 1)), para volver a juzgar el hito entero sin veredicto anterior (revision.sh rejuzgar). Motivo: $motivo. Veredictos apartados:${apartados#,}$([ -n "$apartados" ] || echo " ninguno")."
    echo "rejuzgar: ciclo $ciclo abierto para juzgar el hito entero desde la ronda $((rondas_hechas + 1)); veredictos apartados:${apartados#,}$([ -n "$apartados" ] || echo " ninguno")" >&2
    jq -c '{ciclo, desde_ronda, motivo, entero}' "$f_ciclo";;

  pendiente)
    retomado=$(scripts/workflow/global.sh retomar "$hito")
    if [ -f "$f_ciclo" ] && [ "$(jq -r '.cerrado // false' "$f_ciclo")" != true ]; then
      jq -c --argjson r "$retomado" '{revisar:true, continua:true, ciclo, motivo, desde, retomado:$r}' "$f_ciclo"
      exit 0
    fi
    ult=$(ultimo_juzgado "$hito")
    ciclo=$(ciclo_siguiente)
    if [ -z "$ult" ]; then
      motivo="revisión final del hito"; commits='[]'
    else
      fuera=$( { fuera_de_gates_entre "$ult"; fuera_de_gates_sin_commitear; } | sort -u)
      if [ -z "$fuera" ]; then
        jq -n --arg ult "$ult" --argjson r "$retomado" '{revisar:false, desde:$ult, retomado:$r}'
        exit 0
      fi
      motivo="cambios fuera de gates/ posteriores al último veredicto"
      commits=$(git log --reverse --format='%H%x09%s' "$ult"..HEAD | while IFS="$(printf '\t')" read -r h s; do
          if [ -n "$(fuera_de_gates_en "$h")" ]; then jq -n --arg h "$h" --arg s "$s" '{sha:$h, asunto:$s}'; fi
        done | jq -s '.')
      anotar_supuesto "La revisión final abre el ciclo $ciclo para juzgar lo que cambió fuera de gates/ después del veredicto sobre $(printf '%s' "$ult" | cut -c1-7): $(jq -r 'map(.sha[0:7] + " «" + .asunto + "»") | join(", ")' <<<"$commits")$([ -z "$(fuera_de_gates_sin_commitear)" ] || echo ", y cambios sin commitear")."
    fi
    jq -n --argjson c "$ciclo" --argjson desde_ronda "$((rondas_hechas + 1))" --arg motivo "$motivo" --arg desde "$ult" --argjson commits "$commits" \
      '{ciclo:$c, desde_ronda:$desde_ronda, motivo:$motivo, desde:(if $desde == "" then null else $desde end), commits:$commits, cerrado:false}' > "$f_ciclo"
    jq -c --argjson r "$retomado" '{revisar:true, continua:false, ciclo, motivo, desde, retomado:$r}' "$f_ciclo";;

  cambios)
    n=$((rondas_hechas + 1))
    salida="$g/revision-cambios.md"; diff="$g/revision-cambios-r$n.diff"
    ciclo=$(jq -r '.ciclo // 1' "$f_ciclo" 2>/dev/null || echo 1)
    motivo=$(jq -r '.motivo // "revisión final del hito"' "$f_ciclo" 2>/dev/null || echo "revisión final del hito")
    ult=$(rondas_juzgadas "$hito" | tail -1)
    # La primera ronda de un ciclo `entero` (revision.sh rejuzgar) no tiene veredicto
    # anterior: los de los ciclos anteriores están apartados. Sus rondas siguientes, si el
    # corrector cambió algo, vuelven al rango desde la última ronda del ciclo.
    entero_sin_ronda=""
    if [ "$(jq -r '.entero // false' "$f_ciclo" 2>/dev/null)" = true ] \
       && [ "$(printf '%s\n' "$ult" | cut -d' ' -f1)" -lt "$(jq -r '.desde_ronda' "$f_ciclo")" ] 2>/dev/null; then
      entero_sin_ronda=1; ult=""
    fi
    sucio=$(fuera_de_gates_sin_commitear)
    {
      printf '# Qué juzga la ronda %s de la revisión final\n\n' "$n"
      printf 'Ciclo %s: %s. Lo escribe `scripts/workflow/revision.sh cambios` sin modelo, al preparar a los jueces (ADR 0030).\n\n' "$ciclo" "$motivo"
      if [ -n "$entero_sin_ronda" ]; then
        printf 'Rejuicio del hito entero: este ciclo lo abrió `scripts/workflow/revision.sh rejuzgar` para volver a juzgar el hito desde cero. Su motivo: %s.\n\n' "${motivo%.}"
        printf 'No hay ningún veredicto anterior tuyo: los de los ciclos anteriores están apartados y no cuentan para tu regla de CONVERGENCIA. Juzgas el hito entero, `git diff main...HEAD`.\n'
      elif [ -z "$ult" ]; then
        printf 'Primera ronda: no hay ningún veredicto anterior. Juzgas el hito entero, `git diff main...HEAD`.\n'
      else
        set -- $ult # ronda ciclo sha juez_a juez_b
        printf 'Tu último veredicto es el de la ronda %s, sobre `%s` (juez A: %s; juez B: %s). ' "$1" "$(printf '%s' "$3" | cut -c1-7)" "$4" "$5"
        printf 'Lo que ha cambiado fuera de `gates/` desde entonces es el rango `%s..HEAD` (`%s`):\n\n' "$(printf '%s' "$3" | cut -c1-7)" "$(git rev-parse --short HEAD)"
        n_commits=0
        while IFS="$(printf '\t')" read -r h hc s; do
          [ -n "$h" ] || continue
          f=$(fuera_de_gates_en "$h")
          [ -n "$f" ] || continue
          n_commits=$((n_commits + 1))
          printf -- '- `%s` %s — %s: %s.\n' "$hc" "$s" "$(paso_de "$s")" "$(printf '%s\n' "$f" | sed 's/.*/`&`/' | paste -sd ',' - | sed 's/,/, /g')"
        done < <(git log --reverse --format='%H%x09%h%x09%s' "$3"..HEAD)
        [ "$n_commits" -gt 0 ] || printf -- '- Ningún commit cambia nada fuera de `gates/` desde ese veredicto.\n'
        git diff "$3" HEAD -- . ":(exclude)$g" > "$diff"
        printf '\nEl diff completo de ese rango, fuera de `gates/`, está en `%s` (%s líneas cambiadas). ' "$diff" "$(awk '/^[+-]/ && !/^(\+\+\+|---) / {n++} END {print n + 0}' "$diff")"
        printf 'Para tu regla de CONVERGENCIA, todo el rango es «la corrección», sea del corrector de la revisión o de otro paso posterior a tu veredicto.\n'
        if grep -q "cierre en la plataforma" <(git log --format=%s "$3"..HEAD); then
          printf '\nEl rango trae una reparación del cierre: la medición de CI y evals que la motivó está en `gates/cierre.json`, `gates/cierre.log` y `gates/evals/*.json` (con sus `umbrales`), y sus decisiones, en las líneas `reparar_cierre` de `gates/supuestos.md`. Tras esta revisión, el workflow vuelve a medir la cabeza.\n'
        fi
      fi
      if [ -n "$sucio" ]; then
        printf '\nAviso: el árbol tiene cambios sin commitear fuera de `gates/` que no entran en el rango: %s.\n' "$(printf '%s\n' "$sucio" | sed 's/.*/`&`/' | paste -sd ',' - | sed 's/,/, /g')"
      fi
    } > "$salida"
    echo "revisión: la ronda $n juzga $( [ -z "$ult" ] && echo "el hito entero$([ -z "$entero_sin_ronda" ] || echo ", en un rejuicio")" || echo "el rango desde $(printf '%s' "$ult" | cut -d' ' -f3 | cut -c1-7)" ) ($salida)" >&2;;

  *) echo "subcomando desconocido: $sub" >&2; exit 2;;
esac
