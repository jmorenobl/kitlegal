#!/usr/bin/env bash
# Informe final del hito (ADR 0018): lo único que la persona lee antes de decidir.
#
#   scripts/workflow/informe.sh <hito>             # escribe, commitea, empuja y lo pone de cuerpo de la propuesta
#   scripts/workflow/informe.sh <hito> --solo-ver  # lo escribe en stdout, sin tocar nada
#
# Se construye sin modelo, desde los artefactos y los gates del run, en el orden en que
# la persona tiene que leerlo (ADR 0028): primero lo que puede cambiar su decisión.
#   1. estado: make ci local, CI y evals remotos sobre la cabeza, revisión final;
#   2. supuestos por impacto: cada escritor etiqueta su línea de gates/supuestos.md
#      ([comportamiento], [alcance], [skill], [interno], [proceso]) y clarify da el
#      impacto de cada respuesta; se listan enteros los que cambian lo observable, el
#      alcance o una skill, después lo del propio run (gates agotados, cuarentena,
#      observaciones de los jueces), después lo que no lleva etiqueta —runs anteriores
#      a la 2.1.0: se enseña entero, nunca se esconde— y lo interno solo contado;
#   3. evals: la tasa de cada eval del job sobre la cabeza (gates/evals/<skill>.json,
#      que recoge scripts/workflow/cierre.sh), con las nuevas, las cambiadas y las
#      informativas marcadas; y debajo, los recuentos que publica el job
#      (expresiones_prohibidas_por_modelo: con alguna, total y porcentaje por modelo) y
#      sus `umbrales` (ADR 0029), con los que no se cumplen o solo se publican marcados;
#   4. lo que la constitución (capa 3) reserva a la persona: fuentes, anomalías,
#      adaptadores nuevos, fixtures y esquemas nuevos o modificados (nombrados),
#      grabaciones y evidencias;
#   5. tareas en cuarentena, con su parche y su nota;
#   6. trazabilidad: cada FR/SC del spec → tareas (y su estado) → guiones de aceptación,
#      y su control de umbral (la sección «Controles de umbral» de plan.md): «tareas
#      hechas» no es «comprobado por su control» (ADR 0029);
#   7. commits posteriores a la revisión final, que ningún juez vio, con lo que toca
#      cada uno fuera de gates/;
#   8. cómo comprobarlo y duración del run.
set -euo pipefail
cd "$(dirname "$0")/../.."
. scripts/workflow/comun.sh

hito="${1:?uso: informe.sh <hito> [--solo-ver]}"
solo_ver="${2:-}"
d=$(feature_dir)
g="$d/gates"
rama=$(git branch --show-current)
titulo=$(awk -v h="$hito" 'index($0, "#### " h " · ") == 1 {print; exit}' docs/ROADMAP.md | sed -E 's/^#### [A-Z0-9.]+ · //')

estado_tarea() { # X, cuarentena o pendiente
  if jq -e --arg id "$1" '.tareas | has($id)' "$g/cuarentena.json" >/dev/null 2>&1; then echo cuarentena
  elif grep -qE "^[[:space:]]*- \[[Xx]\] $1([^0-9]|\$)" "$d/tasks.md"; then echo hecha
  else echo pendiente; fi
}

veredicto() { [ -f "$1" ] && jq -r .veredicto "$1" || echo "sin veredicto"; }

# Supuestos de gates/supuestos.md y de clarify, clasificados por impacto, en JSON:
# [{impacto, autor, texto}]. El impacto de una línea es su etiqueta inicial; sin
# etiqueta es "sin". El de una respuesta de clarify es su campo impacto; sin él
# (runs anteriores a la 2.1.0) lo que se decidió dejar fuera —criterio b o lectura
# conservadora— es alcance, y lo demás queda sin clasificar.
supuestos_json() {
  local s="$g/supuestos.md" c="$g/clarify-respuestas.json"
  {
    if [ -f "$s" ]; then
      jq -R -s '[split("\n")[] | select(startswith("- ")) | .[2:]
        | (capture("^\\[(?<impacto>comportamiento|alcance|skill|interno|proceso)\\] (?<texto>.*)$") // {impacto: "sin", texto: .})
        | .autor = ((.texto | capture("^(?<a>[^:·]{1,60}?)(:| ·)") | .a) // "—")]' "$s"
    else echo '[]'; fi
    if [ -f "$c" ]; then
      jq '[.respuestas[] | select(.criterio != "a")
        | {impacto: (.impacto // (if .conservadora == true or .criterio == "b" then "alcance" else "sin" end)),
           autor: ("clarify " + .id),
           texto: ("clarify " + .id + " (criterio " + .criterio + (if .conservadora then ", conservadora" else "" end) + "): " + .respuesta)}]' "$c"
    else echo '[]'; fi
  } | jq -s 'add'
}

# Una línea de supuesto en el informe: entera hasta 500 caracteres; si es más larga, la
# línea completa está en su fichero, que es lo que se lee si el resumen no convence.
lista_supuestos() { # $1 = impacto
  jq -r --arg i "$1" --arg f "$g/supuestos.md" '.[] | select(.impacto == $i)
    | "- " + (.texto | if length > 500 then .[0:500] + "… (entera en `" + $f + "` o `clarify-respuestas.json`)" else . end)'
}

seccion_supuestos() {
  local todos n i nombre p obs
  todos=$(supuestos_json)
  echo "Decisiones que el run tomó sin preguntar, ordenadas por impacto: cada paso que escribe una etiqueta la suya (ADR 0028)."
  echo
  echo "### Cambian el comportamiento visible, el alcance o una skill"
  echo
  n=$(jq '[.[] | select(.impacto == "comportamiento" or .impacto == "alcance" or .impacto == "skill")] | length' <<<"$todos")
  if [ "$n" -eq 0 ]; then echo "Ninguno etiquetado así."; echo; fi
  for i in comportamiento alcance skill; do
    [ "$(jq --arg i "$i" '[.[] | select(.impacto == $i)] | length' <<<"$todos")" -gt 0 ] || continue
    case "$i" in
      comportamiento) nombre="Comportamiento (salida, códigos, ficheros, argumentos)";;
      alcance) nombre="Alcance (lo que queda fuera o dentro del hito)";;
      skill) nombre="Skill (lo que pide, dice o comprueba una skill)";;
    esac
    echo "**$nombre**"; echo
    lista_supuestos "$i" <<<"$todos"
    echo
  done

  echo "### Del propio run"
  echo
  lista_supuestos proceso <<<"$todos"
  for p in "$g"/*-pendiente.md; do
    [ -f "$p" ] || continue
    echo; echo "**$(basename "$p")**:"; echo; sed 's/^/> /' "$p"
  done
  # Lo que los jueces vieron por debajo del umbral de materialidad (ADR 0028): no fue a
  # ningún corrector; se cuenta por fase, con el veredicto donde está cada una.
  obs=$(for p in spec plan tasks revision-a revision-b; do
          f=$(ls "$g/$p"-r*.json 2>/dev/null | sort | tail -1 || true) # como mucho cuatro rondas: el orden alfabético basta
          [ -n "$f" ] || continue
          n=$(jq '(.observaciones // []) | length' "$f" 2>/dev/null || echo 0)
          [ "$n" -gt 0 ] && printf '%s en `%s`, ' "$n" "$(basename "$f")"
        done || true)
  [ -z "$obs" ] || { echo; echo "- Observaciones de los jueces, por debajo del umbral y sin corregir: ${obs%, }."; }
  echo

  n=$(jq '[.[] | select(.impacto == "sin")] | length' <<<"$todos")
  if [ "$n" -gt 0 ]; then
    echo "### Sin clasificar ($n)"
    echo
    echo "Líneas sin etiqueta de impacto: las de un run anterior al workflow 2.1.0, o de un paso que no la puso. Van enteras, para no esconder ninguna."
    echo
    lista_supuestos sin <<<"$todos"
    echo
  fi

  n=$(jq '[.[] | select(.impacto == "interno")] | length' <<<"$todos")
  echo "### Internos ($n)"
  echo
  if [ "$n" -gt 0 ]; then
    echo "Decisiones que no cambian nada observable (técnica, estructura, tests), por autor: $(jq -r '[.[] | select(.impacto == "interno") | .autor]
      | group_by(.) | map("\(.[0]) (\(length))") | join(", ")' <<<"$todos"). Enteras en \`$g/supuestos.md\`."
  else
    echo "Ninguno."
  fi
  echo
}

# Tasa de cada eval del job sobre la cabeza, una tabla por skill (gates/evals/<skill>.json).
seccion_evals() {
  local f cambios cabeza
  if ! ls "$g"/evals/*.json >/dev/null 2>&1; then
    if [ -f "$g/cierre.json" ]; then
      echo "El cierre no dejó el informe de ningún trabajo de evals. En un run anterior al workflow 2.1.0 se recoge con \`scripts/workflow/cierre.sh evals $hito\`."
    else
      echo "Sin medir: el cierre no llegó a ejecutarse."
    fi
    echo; return
  fi
  # Evals que la rama añade (A) o cambia (M), por ruta.
  cambios=$(git diff --no-renames --name-status main...HEAD -- evals/ | jq -R -s '[split("\n")[] | select(length > 0) | split("\t") | {(.[1]): .[0]}] | add // {}')
  cabeza=$(jq -r '.sha // ""' "$g/cierre.json" 2>/dev/null || true)
  for f in "$g"/evals/*.json; do
    jq -r --argjson cambios "$cambios" --arg cabeza "$cabeza" '
      . as $inf
      | ([$inf.modelo_que_decide] + $inf.modelos_informativos + [$inf.tasas[].modelo]) | reduce .[] as $m ([]; if index([$m]) then . else . + [$m] end) | . as $modelos
      | [$inf.tasas[] | .fila = (.eval + (if .pregunta_ampliada then " (prueba de red)" elif (.planificada | not) then " (fuera del plan)" else "" end))]
        | group_by(.fila) | map({fila: .[0].fila, eval: .[0].eval, series: .}) as $filas
      | [$filas[] | . + {
          cambio: ($cambios["evals/" + $inf.skill + "/" + .eval] // ""),
          informativa: ([.series[] | select(.modelo == $inf.modelo_que_decide and .planificada and (.decide | not))] | length > 0)}] as $filas
      | "**\($inf.skill)**: \($inf.veredicto) sobre `\($inf.commit[0:7])`"
        + (if $cabeza != "" and ($inf.commit | startswith($cabeza[0:7]) | not) then " (NO es la cabeza medida, `\($cabeza[0:7])`)" else "" end)
        + "; decide `\($inf.modelo_que_decide)` con \($inf.umbral) de \($inf.repeticiones); "
        + "\($filas | length) evals, \([$filas[] | select(.cambio == "A")] | length) nuevas, \([$filas[] | select(.informativa)] | length) informativas.",
        "",
        (if ($inf.motivos | length) > 0 then ($inf.motivos[] | "- Motivo del fallo: " + .), "" else empty end),
        "| Eval | " + ($modelos | join(" | ")) + " | Marca |",
        "|---|" + ($modelos | map("---") | join("|")) + "|---|",
        ($filas | sort_by([(if .cambio == "A" then 0 elif .cambio == "M" then 1 else 2 end), .fila])[]
          | . as $r
          | "| `\($r.fila)` | "
            + ([$modelos[] as $m | ([$r.series[] | select(.modelo == $m)][0]
                | if . == null then "—" else "\(.pasan)/\(.sesiones)" + (if .decide and (.pasa | not) then " ✗" else "" end) end)] | join(" | "))
            + " | " + ([(if $r.cambio == "A" then "**nueva**" elif $r.cambio == "M" then "cambiada" else empty end),
                        (if $r.informativa then "informativa" else empty end)] | join(" · ")) + " |"),
        ""' "$f"
    recuentos_y_umbrales "$f"
  done
  echo "Cada celda: sesiones que pasan de las abiertas; ✗, una serie que decide y no llega al umbral. Una eval informativa publica su tasa sin decidir el veredicto (ADR 0016). La regla por serie no hace cumplir un umbral agregado sobre todas las respuestas: eso solo lo hace un umbral del job que lo hace fallar (ADR 0029)."
  echo
}

# Umbrales del informe del job (contrato del ADR 0029). Cada elemento de `umbrales` es
#   {nombre, descripcion, medida, total?, comparacion, umbral, cumple, decide}
# Con `total`, lo que se compara es la proporción medida/total (0 si total es 0) y
# `umbral` es una proporción; sin él, la propia medida. La comparación se rehace aquí,
# en coma flotante como la hace el job, para señalar un `cumple` que no casa con la
# medida. `decide` dice si incumplirlo pone el veredicto del job en fallo: un umbral
# que solo se publica no es un control.
jq_umbral='
  def pct($a; $b): if $b > 0 then (($a * 1000 / $b) | round) as $p | "\($p / 10 | floor),\($p % 10) %" else "—" end;
  def valor: if (.total // null) != null then (if .total > 0 then .medida / .total else 0 end) else .medida end;
  def cumple_calculado: valor as $v | .umbral as $u | .comparacion as $o
    | if $o == "<=" then $v <= $u elif $o == "<" then $v < $u elif $o == ">=" then $v >= $u elif $o == ">" then $v > $u else null end;
  def condicion: ((.comparacion as $o | {"<=": "≤", "<": "<", ">=": "≥", ">": ">"}[$o]) // .comparacion) + " "
    + (if (.total // null) != null then ((.umbral * 1000) | round) as $p | "\($p / 10 | floor),\($p % 10) %" else (.umbral | tostring) end);
  def medida_txt: if (.total // null) != null then "\(.medida) de \(.total) (\(pct(.medida; .total)))" else (.medida | tostring) end;
'

# Recuentos y umbrales de una skill, bajo su tabla de tasas (sección 3).
recuentos_y_umbrales() { # $1 = gates/evals/<skill>.json
  jq -r "$jq_umbral"'
    . as $inf
    | (if (.expresiones_prohibidas_por_modelo // []) | length > 0 then
        "Respuestas con alguna expresión prohibida, en las evals que activan la skill (`expresiones_prohibidas_por_modelo`):", "",
        "| Modelo | Con alguna | Respuestas | Porcentaje |", "|---|---|---|---|",
        (.expresiones_prohibidas_por_modelo[] | "| `\(.modelo)` | \(.con_alguna) | \(.respuestas) | \(pct(.con_alguna; .respuestas)) |"), ""
       else empty end),
      (if has("umbrales") | not then
         "Umbrales: el informe del job no trae `umbrales` (es anterior al contrato del ADR 0029)"
           + (if (.expresiones_prohibidas_por_modelo // []) | length > 0 then ". Ningún recuento de esta skill hace fallar el job: lo de arriba solo se publica." else "." end), ""
       elif (.umbrales | length) == 0 then "Umbrales: el job no publica ninguno para esta skill.", ""
       else
         "Umbrales que publica el job (ADR 0029):", "",
         "| Umbral | Medida | Condición | Cumple | Hace fallar el job |", "|---|---|---|---|---|",
         (.umbrales[] | cumple_calculado as $c
           | "| `\(.nombre)` | \(medida_txt) | \(condicion) | "
             + (if .cumple then "sí" else "✗ **NO**" end)
             + (if $c != null and $c != .cumple then " (INCOHERENTE: la medida da \(if $c then "sí" else "no" end))" else "" end)
             + " | " + (if .decide then "sí" else "no: solo se publica, no es un control" end) + " |"),
         "",
         ([.umbrales[] | select(.decide and (.cumple | not))] | select(length > 0 and $inf.veredicto == "aprobado")
           | "- **INCOHERENTE**: el veredicto es aprobado con " + (map("`" + .nombre + "`") | join(", ")) + " sin cumplir y haciendo fallar el job.", "")
       end)' "$1"
}

# Una línea por skill con sus umbrales, para la sección 1.
resumen_umbrales() {
  ls "$g"/evals/*.json >/dev/null 2>&1 || return 0
  jq -rs "$jq_umbral"'map(
      .skill + ": "
      + (if has("umbrales") | not then "el job no publica umbrales"
         elif (.umbrales | length) == 0 then "sin umbrales"
         else ([.umbrales[] | select((.cumple | not) or (.decide | not))]) as $mal
           | "\(.umbrales | length) umbrales"
             + (if ($mal | length) == 0 then ", todos cumplen y hacen fallar el job"
                else ", **\($mal | length) sin cumplir o solo publicados** (" + ($mal | map("`" + .nombre + "`") | join(", ")) + ")" end)
         end)
      + (if ((.expresiones_prohibidas_por_modelo // []) | length > 0) and ((has("umbrales") | not) or (.umbrales | length) == 0)
         then "; publica recuentos que ningún control hace cumplir (expresiones prohibidas: " + ([.expresiones_prohibidas_por_modelo[] | "\(.con_alguna)/\(.respuestas) `\(.modelo)`"] | join(", ")) + ")"
         else "" end))
    | join("; ")' "$g"/evals/*.json
}

# Estado de un control de «Controles de umbral» (plan.md): «ok|…», «no|…» (medido y sin
# cumplir) o «sin|…» (declarado y no verificable: no existe, no se mide o solo se publica).
estado_control() { # $1 = ci:<ruta>[:<Test>] | evals:<skill>:<nombre>
  local c="$1" ruta t skill nombre f
  case "$c" in
    ci:*)
      ruta=${c#ci:}; t=""
      case "$ruta" in *:*) t=${ruta##*:}; ruta=${ruta%%:*};; esac
      if ! git cat-file -e "HEAD:$ruta" 2>/dev/null; then echo "sin|\`$c\`: \`$ruta\` no existe en la cabeza"; return; fi
      if [ -n "$t" ] && ! git grep -qE "^(func $t\\(|$t:)" HEAD -- "$ruta"; then echo "sin|\`$c\`: \`$t\` no está en \`$ruta\`"; return; fi
      if [ "${ci_local:-x}" = 0 ]; then echo "ok|\`$c\`, en make ci (verde)"; else echo "sin|\`$c\`: make ci no está en verde"; fi;;
    evals:*)
      skill=$(printf '%s' "$c" | cut -d: -f2); nombre=$(printf '%s' "$c" | cut -d: -f3-)
      f="$g/evals/$skill.json"
      [ -f "$f" ] || { echo "sin|\`$c\`: no hay informe del job de evals de $skill"; return; }
      jq -r --arg n "$nombre" --arg c "$c" "$jq_umbral"'
        ([(.umbrales // [])[] | select(.nombre == $n)][0]) as $u
        | if $u == null then "sin|`\($c)`: el job no publica ese umbral"
          elif ($u | cumple_calculado) as $k | $k != null and $k != $u.cumple
            then "sin|`\($c)`: \($u | medida_txt), \($u | condicion); INCOHERENTE: el job dice que \(if $u.cumple then "cumple" else "no cumple" end)"
          elif ($u.decide | not) then "sin|`\($c)`: \($u | medida_txt), \($u | condicion); se publica sin hacer fallar el job"
          elif $u.cumple then "ok|`\($c)`: \($u | medida_txt), \($u | condicion)"
          else "no|`\($c)`: \($u | medida_txt), NO cumple \($u | condicion)" end' "$f";;
    *) echo "sin|\`$c\`: control con forma desconocida";;
  esac
}

# Una línea de estado por skill evaluada, para la sección 1.
resumen_evals() {
  local f
  ls "$g"/evals/*.json >/dev/null 2>&1 || return 0
  for f in "$g"/evals/*.json; do
    jq -r '"\(.skill) \(.veredicto)"' "$f"
  done | paste -sd ',' - | sed 's/,/, /g'
}

# Lo que la capa 3 de la constitución reserva a la persona, una línea por asunto, con los
# ficheros nombrados.
capa3() {
  local cambiados nuevos_src f activados nuevos_fix nuevos_esq
  cambiados=$(git diff --no-renames --name-status main...HEAD)
  printf '%s\n' "$cambiados" | awk '$2 == "docs/SOURCES.md" {print "- `docs/SOURCES.md` cambia: responsabilidad sobre qué fuentes se tocan y cómo."}'
  printf '%s\n' "$cambiados" | awk '$2 ~ /^data\/anomalias\// {print "- Regla de anomalías: `" $2 "` (" $1 ")."}'
  nuevos_src=$(printf '%s\n' "$cambiados" | awk '{print $NF}' | grep -oE '^internal/source/[^/]+' | sort -u || true)
  for f in $nuevos_src; do git cat-file -e "main:$f" 2>/dev/null || echo "- Adaptador de fuente nuevo: \`$f/\`."; done
  printf '%s\n' "$cambiados" | awk '$1 == "M" && ($2 ~ /(^|\/)testdata\// || $2 ~ /^schemas\//) {print "- Fixture o esquema EXISTENTE modificado: `" $2 "`."}'
  # La suite de aceptación activada vive bajo testdata/ pero no es un fixture: la escribió
  # la primera tarea desde el spec y está congelada. Se nombra aparte.
  activados=$(jq -r '.ficheros // {} | keys[] | select(startswith("internal/app/testdata/script/"))' "$g/aceptacion-congelada.json" 2>/dev/null || true)
  nuevos_fix=$(printf '%s\n' "$cambiados" | awk '$1 == "A" && $2 ~ /(^|\/)testdata\// {print $2}' | grep -vxF -- "$activados" || true)
  if [ -n "$nuevos_fix" ]; then
    echo "- Fixtures nuevos ($(printf '%s\n' "$nuevos_fix" | grep -c .)), por directorio:"
    printf '%s\n' "$nuevos_fix" | awk '{d = $0; sub(/\/[^\/]*$/, "/", d); f = substr($0, length(d) + 1)
      if (d != prev) { if (prev != "") print linea; linea = "  - `" d "`: `" f "`"; prev = d } else linea = linea ", `" f "`" }
      END { if (prev != "") print linea }'
  fi
  nuevos_esq=$(printf '%s\n' "$cambiados" | awk '$1 == "A" && $2 ~ /^schemas\// {print "`" $2 "`"}' | paste -sd ',' - | sed 's/,/, /g')
  [ -z "$nuevos_esq" ] || echo "- Esquemas nuevos: $nuevos_esq."
  [ -z "$activados" ] || echo "- Suite de aceptación activada y congelada ($(printf '%s\n' "$activados" | grep -c .) guiones, escritos desde el spec en la primera tarea): $(printf '%s\n' "$activados" | sed 's#.*/##; s/.*/`&`/' | paste -sd ',' - | sed 's/,/, /g')."
  printf '%s\n' "$cambiados" | awk '$2 ~ /^evidencias\// {n++} END {if (n) print "- " n " ficheros de evidencia en evidencias/ (manifiesto con huellas)."}'
  if [ -f "$g/grabaciones.md" ]; then echo "- Grabaciones del paso grabar_datos: \`$g/grabaciones.md\`."; fi
}

ci_local=$(tail -n 1 "$g/ci.log" 2>/dev/null | sed -n 's/^kitlegal-verificacion exit=//p' || true) # ci.log no se versiona

informe() {
  echo "# Informe del hito $hito · $titulo"
  echo
  echo "Generado por el workflow \`hito\` el $(date -u +%Y-%m-%dT%H:%M:%SZ), sobre \`$(git rev-parse --short HEAD)\` de \`$rama\`."
  echo "Lo escribe scripts/workflow/informe.sh sin modelo, desde los artefactos de \`$d/\`. Fusionar (squash-merge) es una decisión humana:"
  echo "si algo de lo que sigue no es lo que se quería, se corrige la sección del hito en docs/ROADMAP.md y se relanza."
  echo

  echo "## 1. Estado"
  echo
  local remoto ev um
  echo "- **make ci local**: $( [ "${ci_local:-x}" = 0 ] && echo "verde" || echo "ROJO (exit ${ci_local:-desconocido}; gates/ci.log)")."
  if [ -f "$g/cierre.json" ]; then
    remoto=$(jq -r 'if .verde then "verde" else "ROJO: " + ([.checks[] | select(.workflow != "" and (.bucket == "fail" or .bucket == "cancel")) | .name] | join(", ")) end' "$g/cierre.json")
    echo "- **CI y evals remotos** sobre \`$(jq -r '.sha[0:7]' "$g/cierre.json")\`: $remoto."
  else
    echo "- **CI y evals remotos**: sin medir."
  fi
  ev=$(resumen_evals)
  [ -z "$ev" ] || echo "- **Evals por skill**: $ev (tasas en la sección 3)."
  um=$(resumen_umbrales)
  [ -z "$um" ] || echo "- **Umbrales del job**: $um (sección 3)."
  echo "- **Revisión final**: juez A $(veredicto "$g/revision-a.json"), juez B $(veredicto "$g/revision-b.json"), $(cat "$g/revision-rondas" 2>/dev/null || echo 0) rondas."
  local hechas cuar sin
  hechas=$(grep -cE '^[[:space:]]*- \[[Xx]\] T[0-9]+' "$d/tasks.md" || true)
  cuar=$(jq '.tareas | length' "$g/cuarentena.json" 2>/dev/null || echo 0)
  sin=$(( $(grep -cE '^[[:space:]]*- \[ \] T[0-9]+' "$d/tasks.md" || true) - cuar ))
  echo "- **Tareas**: $hechas hechas, $cuar en cuarentena, $sin pendientes sin cuarentena."
  echo "- **Diff**: $(git rev-list --count main..HEAD) commits; $(git diff --shortstat main...HEAD | sed 's/^ //')."
  echo

  echo "## 2. Supuestos y pendientes"
  echo
  seccion_supuestos

  echo "## 3. Evals sobre la cabeza"
  echo
  seccion_evals

  echo "## 4. Revisión que la constitución reserva a la persona"
  echo
  local capa3
  capa3=$(capa3)
  if [ -n "$capa3" ]; then printf '%s\n' "$capa3"; else echo "Nada: el hito no toca fuentes, anomalías, adaptadores, fixtures, esquemas, grabaciones ni evidencias."; fi
  echo

  echo "## 5. Tareas en cuarentena"
  echo
  if [ "$(jq '.tareas | length' "$g/cuarentena.json" 2>/dev/null || echo 0)" -gt 0 ]; then
    jq -r --arg g "$g" '.tareas | to_entries[] | "- **\(.key)**: \(.value.motivo). Parche `\(.value.parche)`; causa en `\($g)/tarea-\(.key).md`."' "$g/cuarentena.json"
  else
    echo "Ninguna."
  fi
  echo

  # «tareas hechas» y «comprobado» son cosas distintas (ADR 0029): que las tareas que
  # citan un requisito estén marcadas no mide nada. Solo figura como comprobado un
  # requisito con su fila en «Controles de umbral» de plan.md cuyo control está: un
  # test o una comprobación de make ci que existe en la cabeza, con make ci en verde,
  # o un umbral del job de evals publicado, cumplido y que hace fallar el job.
  echo "## 6. Trazabilidad"
  echo
  local filas_ctl req tareas acept est t e toks tok r ctl no sin
  filas_ctl=$(controles_de_umbral "$d/plan.md")
  echo "**Estado**: «tareas hechas» solo dice que las tareas que citan el requisito están marcadas; nada del run lo ha medido. «comprobado por su control» exige su fila en «Controles de umbral» de \`plan.md\` y que el control esté: un test o una comprobación de \`make ci\` que existe en la cabeza, con \`make ci\` en verde, o un umbral que el job de evals publica, cumple y hace fallar el job (ADR 0029). «UMBRAL NO CUMPLIDO» y «CONTROL SIN VERIFICAR» son lo que hay que mirar."
  if ! grep -q '^## Controles de umbral' "$d/plan.md" 2>/dev/null; then
    echo
    echo "\`plan.md\` no tiene «Controles de umbral» (runs anteriores al workflow 2.2.0): ningún requisito figura como comprobado por un control."
  fi
  echo
  echo "| Requisito | Tareas | Aceptación | Estado | Control de umbral |"
  echo "|---|---|---|---|---|"
  for req in $(grep -oE '\*\*(FR|SC)-[0-9]+\*\*' "$d/spec.md" | tr -d '*' | awk '!v[$0]++'); do
    tareas=""; est="tareas hechas"
    while IFS= read -r l; do
      t=$(printf '%s' "$l" | grep -oE 'T[0-9]+' | head -1)
      ids_de_linea "$l" | grep -qx "$req" || continue
      e=$(estado_tarea "$t"); tareas="$tareas $t($e)"
      case "$e" in cuarentena) est="EN CUARENTENA";; pendiente) [ "$est" = "EN CUARENTENA" ] || est="PENDIENTE";; esac
    done < <(grep -E '^[[:space:]]*- \[[ xX]\] T[0-9]+' "$d/tasks.md")
    [ -n "$tareas" ] || est="SIN TAREA"
    acept=$(grep -lwF "$req" "$d"/aceptacion/* 2>/dev/null | xargs -n1 basename 2>/dev/null | tr '\n' ' ' || true)
    # Controles de las filas de «Controles de umbral» que nombran este requisito.
    toks=""; ctl=""; no=0; sin=0
    while IFS="$(printf '\t')" read -r celda donde; do
      [ -n "$celda" ] || continue
      ids_de_linea "$celda" | grep -qx "$req" || continue
      toks="$toks $(controles_de_celda "$donde" | tr '\n' ' ')"
      [ -n "$(controles_de_celda "$donde")" ] || { sin=1; ctl="$ctl; la fila no dice dónde vive el control"; }
    done <<<"$filas_ctl"
    for tok in $toks; do
      r=$(estado_control "$tok")
      case "${r%%|*}" in no) no=1;; sin) sin=1;; esac
      ctl="$ctl; ${r#*|}"
    done
    ctl=${ctl#; }
    if [ "$est" = "tareas hechas" ] && [ -n "$ctl" ]; then
      if [ "$no" = 1 ]; then est="UMBRAL NO CUMPLIDO"
      elif [ "$sin" = 1 ]; then est="CONTROL SIN VERIFICAR"
      else est="comprobado por su control"; fi
    fi
    echo "| $req |${tareas:- —} | ${acept:-—} | $est | ${ctl:-—} |"
  done
  echo

  echo "## 7. Cambios posteriores a la revisión final"
  echo
  local ult h hc s fuera
  ult=$(git log --format=%H --grep="^docs($hito): veredictos de la revisión final" -n 1 main..HEAD || true)
  if [ -n "$ult" ] && [ -n "$(git log --oneline "$ult"..HEAD)" ]; then
    echo "Commits posteriores a los veredictos, que ningún juez juzgó (correcciones del cierre, registros y este informe), con lo que cada uno toca fuera de \`gates/\`:"; echo
    git log --format='%H %h %s' "$ult"..HEAD | while read -r h hc s; do
      fuera=$(git diff-tree --no-commit-id --name-only -r "$h" | grep -v "^$d/gates/" || true)
      if [ -z "$fuera" ]; then echo "- \`$hc\` $s: solo registros de \`gates/\`."
      else echo "- \`$hc\` $s: $(printf '%s\n' "$fuera" | sed "s#^$d/##" | sed 's/.*/`&`/' | paste -sd ',' - | sed 's/,/, /g')."; fi
    done
  else
    echo "Ninguno."
  fi
  echo

  echo "## 8. Cómo comprobarlo y consumo"
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
