#!/usr/bin/env bash
# Serie de cierres de las skills con juez (ADR 0038, punto 6), sin modelo.
#
#   scripts/workflow/serie.sh fila <skill> <procedencia> [<informe.json>]  # la fila de una medición, en una línea
#   scripts/workflow/serie.sh anotar <serie.jsonl> <fila>                  # la escribe, en lugar de la de esa medición
#   scripts/workflow/serie.sh semilla <fuentes.tsv>                        # las filas previas, de los registros de Actions
#   scripts/workflow/serie.sh mostrar [<skill>]                            # la serie y la tasa acumulada, en Markdown
#   scripts/workflow/serie.sh escalones <total>                            # la tasa más alta de cada escalón de la regla 3
#   scripts/workflow/serie.sh clases <skill> [<commit>]                    # clases.yaml en JSON
#   scripts/workflow/serie.sh por-materia <skill> [<commit>]               # las evals con por_materia: true, en JSON
#
# Cada cierre añade a la serie de cada skill con juez una fila por medición: el paso
# `medir` de cierre.sh, al recoger los informes del job de evals, la escribe en
# specs/<hito>/gates/serie.jsonl. Dentro de gates/ a propósito: es un registro del
# run, y escribirla no cambia el producto que la medición midió (ADR 0030). Cuentan
# todas las mediciones completas del cierre, no solo la que pasó; una incompleta
# —una respuesta medida sin juzgar o una sesión del modelo que decide sin medir— se
# registra con `cuenta: false` y su porqué, y no entra en la tasa. Lo medido fuera de
# un cierre no entra nunca: la serie solo la escribe el cierre.
#
# La serie es la unión de evidencias/adr-0038/serie-previa.jsonl —las mediciones de
# cierre anteriores a este guion, reconstruidas con `semilla` desde los registros de
# GitHub Actions— y de todos los specs/*/gates/serie.jsonl. El formato de cada fila es
# schemas/serie-de-cierres.json, que comprueba `make ci`.
#
# Todo sale de campos del informe del job: las respuestas medidas son las de
# `.evals[]` del modelo que decide, en el modo orden o herramienta, de una eval que
# activa la skill, terminadas y medidas; las marcadas, las de `.juez.respuestas[]`
# con la clase marcada; las dos se cruzan por `.sesion`, que nunca se parte. Lo
# contado se compara con los umbrales del propio informe —el recuento de cada modo y,
# desde H26, el conjunto— y, si no casa, `fila` sale con 1 y lo dice: la fila no se
# escribe.
#
# La tasa acumulada es la de las filas que cuentan con la clave de la cabeza: el
# árbol de skills/<skill> («la skill como está»; el binario no entra), el modelo que
# decide y la huella del juez —su modelo, la versión de Claude Code de sus votos y la
# sha256 de su rúbrica y de sus casos—. `mostrar` no decide nada: enseña la serie, la
# tasa y, con 500 respuestas o más, el escalón de la regla 3 del ADR 0038. Nunca
# propone subir un umbral; si permite estrecharlo, lo dice, y lo baja una persona.
set -euo pipefail
cd "$(dirname "$0")/../.."
. scripts/workflow/comun.sh

# Las respuestas que exige la regla 3 para que la tasa acumulada cuente, y la
# frecuencia de rojos sin motivo que admite el escalón.
minimo_de_respuestas=500
rojos_admitidos=0.02

serie_previa=evidencias/adr-0038/serie-previa.jsonl

# Huella SHA-256 de la entrada estándar.
huella_de_la_entrada() {
  if command -v sha256sum >/dev/null 2>&1; then sha256sum | cut -d' ' -f1
  else shasum -a 256 | cut -d' ' -f1; fi
}

# El contenido de un fichero del repositorio en un commit, o en el árbol de trabajo
# si no se da ninguno.
contenido() { # $1 = ruta, $2 = commit o vacío
  if [ -n "${2:-}" ]; then git show "$2:$1"; else cat "$1"; fi
}

# Las clases de evals/<skill>/juez/clases.yaml, en JSON: [{nombre, decide, umbral}].
# Se leen con awk, sin un lector de YAML: el fichero tiene la forma de
# schemas/juez-clases.yaml.json, una clave por línea, y un test comprueba que lo leído
# aquí es lo que lee el job.
clases() { # $1 = skill, $2 = commit o vacío
  contenido "evals/$1/juez/clases.yaml" "${2:-}" | awk '
    function valor(s) { sub(/^[^:]*:[[:space:]]*/, "", s); sub(/[[:space:]]+#.*$/, "", s); gsub(/^["'\'']|["'\'']$/, "", s); return s }
    /^[[:space:]]*#/ { next }
    /^[[:space:]]*-[[:space:]]*nombre:/ { if (n++) print fila; fila = "{\"nombre\":\"" valor($0) "\"" ; next }
    /^[[:space:]]+decide:/ { fila = fila ",\"decide\":" tolower(valor($0)); next }
    /^[[:space:]]+umbral:/ { fila = fila ",\"umbral\":" valor($0); next }
    END { if (n) print fila }' | sed 's/$/}/' | jq -sc .
}

# Las evals de evals/<skill>/ con `por_materia: true`, por nombre de fichero, en
# JSON. La clave es un booleano de primer nivel (schemas/eval.yaml.json) y se busca
# como línea; un test comprueba que lo leído aquí es lo que lee el job.
por_materia() { # $1 = skill, $2 = commit o vacío
  local f ficheros
  if [ -n "${2:-}" ]; then ficheros=$(git ls-tree --name-only "$2" "evals/$1/")
  else ficheros=$(ls "evals/$1/"*.yaml 2>/dev/null || true); fi
  for f in $ficheros; do
    case "${f##*/}" in [0-9][0-9]-*.yaml) ;; *) continue;; esac
    if contenido "$f" "${2:-}" | grep -qiE '^por_materia:[[:space:]]*true[[:space:]]*(#.*)?$'; then printf '%s\n' "${f##*/}"; fi
  done | jq -Rsc 'split("\n") | map(select(length > 0))'
}

# La fila de una medición de la skill. La procedencia es un objeto JSON con hito,
# medicion, fecha, commit, propuesta, ejecucion y trabajo; si hay informe legible, el
# commit es el suyo, el que midió el job. Sin informe, la fila se registra sin contar.
fila() { # $1 = skill, $2 = procedencia, $3 = informe.json o vacío
  local skill="$1" procedencia="$2" informe="${3:-}" commit arbol rubrica casos clases_json pm pm_de tmp errores
  if [ -n "$informe" ] && jq -e '.skill and (.evals | type == "array")' "$informe" >/dev/null 2>&1; then
    commit=$(jq -r '.commit // ""' "$informe")
    [ -n "$commit" ] || commit=$(jq -r .commit <<<"$procedencia")
  else
    informe=""
    commit=$(jq -r .commit <<<"$procedencia")
  fi
  git cat-file -e "$commit^{commit}" 2>/dev/null || { echo "serie: el commit $commit no está en el repositorio" >&2; return 1; }
  arbol=$(git rev-parse "$commit:skills/$skill")
  rubrica=$(git show "$commit:evals/$skill/juez/rubrica.md" | huella_de_la_entrada)
  casos=$(git show "$commit:evals/$skill/juez/casos.yaml" | huella_de_la_entrada)
  clases_json=$(clases "$skill" "$commit")
  procedencia=$(jq -c --arg c "$commit" '.commit = $c' <<<"$procedencia")

  if [ -z "$informe" ]; then
    jq -nc --arg skill "$skill" --argjson p "$procedencia" --arg arbol "$arbol" '
      {skill: $skill, procedencia: $p, arbol_de_la_skill: $arbol, modelo_que_decide: null, juez: null,
       por_materia_de: null, evals_por_materia: [], sin_juzgar: null, sesiones_sin_medir: null,
       cuenta: false, no_cuenta_porque: "el registro del trabajo no trae un informe.json legible", clases: []}'
    return 0
  fi

  # Los informes anteriores al campo por_materia de cada sesión (los de la semilla) lo
  # toman de los ficheros de la eval del árbol de trabajo, y la fila lo dice.
  if jq -e '(.evals | length) > 0 and all(.evals[]; has("por_materia"))' "$informe" >/dev/null; then
    pm_de=informe; pm='[]'
  else
    pm_de=yaml; pm=$(por_materia "$skill")
  fi

  tmp=$(mktemp)

  jq -c --arg skill "$skill" --argjson p "$procedencia" --arg arbol "$arbol" --arg rubrica "$rubrica" \
    --arg casos "$casos" --argjson clases "$clases_json" --arg pm_de "$pm_de" --argjson pm_yaml "$pm" '
    . as $inf | $inf.modelo_que_decide as $m
    | def en($lista; $x): $lista | any(.[]; . == $x);
    if $inf.juez == null then
      {skill: $skill, procedencia: $p, arbol_de_la_skill: $arbol, modelo_que_decide: $m, juez: null,
       por_materia_de: null, evals_por_materia: [], sin_juzgar: null,
       sesiones_sin_medir: ([($inf.sesiones_sin_medir // [])[] | select(.modelo == $m)] | length),
       cuenta: false, no_cuenta_porque: ("el informe no trae el juez"
         + ([($inf.motivos // [])[] | select(contains("el instrumento no está medido"))][0] // "" | if . == "" then "" else ": " + . end)),
       clases: []}
    else
      [($inf.evals // [])[] | select(.modelo == $m and (.modo == "orden" or .modo == "herramienta") and .activa
                             and .sin_medir == "" and .sesion_terminada)] as $med
      | (if $pm_de == "informe" then [$med[] | select(.por_materia) | .eval] | unique else $pm_yaml end) as $pm
      | [$med[].sesion] as $sesiones
      | ([($inf.juez.sin_juzgar // [])[] | select(en($sesiones; .sesion))] | length) as $sj
      | ([($inf.sesiones_sin_medir // [])[] | select(.modelo == $m)] | length) as $ssm
      | [$clases[] | select(.decide) | .nombre as $c
          | [($inf.juez.respuestas // [])[] | select(any(.clases[]; .clase == $c and .marcada)) | .sesion] as $marcadas
          | [("orden", "herramienta") as $modo | (false, true) as $porm
              | [$med[] | select(.modo == $modo and (en($pm; .eval) == $porm))] as $celda
              | {modo: $modo, por_materia: $porm,
                 marcadas: ([$celda[] | select(en($marcadas; .sesion))] | length), total: ($celda | length)}] as $celdas
          | {clase: $c, marcadas: ($celdas | map(.marcadas) | add), total: ($celdas | map(.total) | add), celdas: $celdas}
        ] as $porclase
      | ([if $sj > 0 then "\($sj) respuesta" + (if $sj > 1 then "s" else "" end) + " sin juzgar" else empty end,
          if $ssm > 0 then "\($ssm) sesi" + (if $ssm > 1 then "ones" else "ón" end) + " del modelo que decide sin medir" else empty end,
          if ($porclase | length) == 0 then "ninguna clase del juez decide" else empty end] | join("; ")) as $porque
      | {skill: $skill, procedencia: $p, arbol_de_la_skill: $arbol, modelo_que_decide: $m,
         juez: {modelo: $inf.juez.modelo, version_de_claude_code: $inf.juez.version_de_claude_code,
                rubrica_sha256: $rubrica, casos_sha256: $casos},
         por_materia_de: $pm_de, evals_por_materia: $pm, sin_juzgar: $sj, sesiones_sin_medir: $ssm,
         cuenta: ($porque == ""), no_cuenta_porque: (if $porque == "" then null else $porque end), clases: $porclase}
    end' "$informe" > "$tmp"

  # El cruce con los umbrales del propio informe: el recuento de cada modo
  # (<clase>:<modelo>:<modo>, desde H24) y el conjunto (<clase>:<modelo>, desde H26).
  errores=$(jq -r --slurpfile f "$tmp" '
    . as $inf | $f[0] as $fila | $inf.modelo_que_decide as $m
    | $fila.clases[] as $c
    | ([($inf.umbrales // [])[] | select(.nombre == $c.clase + ":" + $m)][0]) as $conjunto
    | ([("orden", "herramienta") as $modo
        | ([($inf.umbrales // [])[] | select(.nombre == $c.clase + ":" + $m + ":" + $modo)][0]) as $u
        | ($c.celdas | map(select(.modo == $modo))) as $cel
        | {modo: $modo, u: $u, marcadas: ($cel | map(.marcadas) | add), total: ($cel | map(.total) | add)}]) as $modos
    | (if $conjunto == null and all($modos[]; .u == null) then
         "\($c.clase): el informe no trae ni el umbral conjunto `\($c.clase):\($m)` ni el recuento de ningún modo"
       else empty end),
      ($modos[] | select(.u == null and .total > 0)
        | "\($c.clase): el informe no trae el recuento del modo \(.modo), y la serie cuenta \(.total) respuestas en él"),
      ($modos[] | select(.u != null and (.u.medida != .marcadas or .u.total != .total))
        | "\($c.clase) en el modo \(.modo): la serie cuenta \(.marcadas) de \(.total) y el umbral `\(.u.nombre)` del informe, \(.u.medida) de \(.u.total)"),
      (if $conjunto != null and ($conjunto.medida != $c.marcadas or $conjunto.total != $c.total) then
         "\($c.clase): la serie cuenta \($c.marcadas) de \($c.total) y el umbral conjunto `\($conjunto.nombre)` del informe, \($conjunto.medida) de \($conjunto.total)"
       else empty end)' "$informe")
  if [ -n "$errores" ]; then
    rm -f "$tmp"
    printf 'serie: la fila de %s no casa con los umbrales de su informe (%s, medición %s):\n%s\n' \
      "$skill" "$(jq -r .hito <<<"$procedencia")" "$(jq -r .medicion <<<"$procedencia")" "$errores" | sed '2,$s/^/  - /' >&2
    return 1
  fi
  cat "$tmp"; rm -f "$tmp"
}

# Escribe la fila en la serie: la sustituye si ya hay una de la misma skill, hito y
# medición (recoger otra vez los informes de una medición no la duplica) y la añade
# al final si no.
anotar() { # $1 = serie.jsonl, $2 = fila
  local destino="$1" nueva="$2" tmp
  tmp="$destino.tmp.$$"
  mkdir -p "$(dirname "$destino")"
  { [ -f "$destino" ] && jq -c --argjson n "$nueva" \
      'select((.skill == $n.skill and .procedencia.hito == $n.procedencia.hito and .procedencia.medicion == $n.procedencia.medicion) | not)' \
      "$destino"; printf '%s\n' "$nueva"; } > "$tmp"
  mv "$tmp" "$destino"
}

# Reconstruye las filas de mediciones de cierre anteriores a este guion desde los
# registros de GitHub Actions, que caducan a los 90 días (ADR 0038). Cada línea de
# fuentes.tsv: hito, medición, propuesta, ejecución, trabajo y skill, separados por
# tabuladores; las que empiezan por # se saltan. El informe se extrae del registro
# como lo hace el cierre; si el commit que midió ya no está en el repositorio, se
# trae de refs/pull/<propuesta>/head. Las filas salen por la salida estándar, en el
# orden de fuentes.tsv; lo que no se puede leer, se dice en la fila, nunca se rellena.
semilla() { # $1 = fuentes.tsv
  local hito medicion propuesta ejecucion trabajo skill tmp fecha commit p
  tmp=$(mktemp -d)
  while IFS=$'\t' read -r hito medicion propuesta ejecucion trabajo skill; do
    case "$hito" in ''|'#'*) continue;; esac
    informe_del_trabajo "$ejecucion" "$trabajo" "$tmp/informe.json" "evals ($skill)" "$skill" </dev/null || true
    fecha=$(gh api "repos/{owner}/{repo}/actions/jobs/$trabajo" --jq .started_at </dev/null)
    commit=$(jq -r '.commit // ""' "$tmp/informe.json" 2>/dev/null || true)
    [ -n "$commit" ] || commit=$(gh api "repos/{owner}/{repo}/actions/runs/$ejecucion" --jq .head_sha </dev/null)
    git cat-file -e "$commit^{commit}" 2>/dev/null || git fetch -q origin "refs/pull/$propuesta/head" </dev/null
    p=$(jq -nc --arg h "$hito" --argjson m "$medicion" --arg f "$fecha" --arg c "$commit" --argjson pr "$propuesta" \
      --argjson e "$ejecucion" --argjson t "$trabajo" \
      '{hito: $h, medicion: $m, fecha: $f, commit: $c, propuesta: $pr, ejecucion: $e, trabajo: $t}')
    fila "$skill" "$p" "$( [ -s "$tmp/informe.json" ] && echo "$tmp/informe.json" )"
    rm -f "$tmp/informe.json"
  done < "$1"
  rm -rf "$tmp"
}

# Funciones de jq de la regla 3 del ADR 0038: la probabilidad de que una ejecución de
# n respuestas, a la tasa p, marque más de k; el escalón de una tasa —el umbral más
# estricto con el que esa ejecución sale en rojo sin motivo como mucho 1 de cada 50
# veces—; y la tasa más alta con la que vale un escalón.
jq_regla='
  def cola($n; $p; $k):
    if $p <= 0 then 0 elif $p >= 1 then (if $k < $n then 1 else 0 end)
    else 1 - (reduce range(0; $k + 1) as $i ({s: 0, c: pow(1 - $p; $n)};
                .s += .c | .c = .c * ($n - $i) / ($i + 1) * $p / (1 - $p)) | .s) end;
  def escalon($n; $p; $admitidos): first(range(0; $n + 1) | select(cola($n; $p; .) <= $admitidos));
  def tasa_maxima($n; $k; $admitidos):
    reduce range(0; 60) as $_ ({lo: 0, hi: 1};
      ((.lo + .hi) / 2) as $x | if cola($n; $x; $k) <= $admitidos then .lo = $x else .hi = $x end) | .lo;
  def pct($x; $dec): ($x * 100 * pow(10; $dec) | round) as $r
    | ($r / pow(10; $dec) | floor) as $e | ($r - $e * pow(10; $dec)) as $f
    | "\($e)" + (if $dec > 0 then "," + ("\($f)" | (("0" * ($dec - length)) + .)) else "" end) + " %";
'

# La tasa más alta de cada escalón con un total, de 0 a 5 marcadas, en JSON.
escalones() { # $1 = total
  jq -nc --argjson n "$1" --argjson a "$rojos_admitidos" "$jq_regla"'[range(0; 6) as $k | {escalon: $k, hasta: tasa_maxima($n; $k; $a)}]'
}

# La clave de la cabeza de una skill, en JSON: el árbol de la skill, el modelo que
# decide (MODELO_DE_EVALS del job) y la huella del juez (su medida versionada y la
# sha256 de su rúbrica y de sus casos), todo de HEAD.
clave_de_la_cabeza() { # $1 = skill
  local medida
  medida=$(git show "HEAD:evals/$1/juez/medida.json")
  jq -nc --arg arbol "$(git rev-parse "HEAD:skills/$1")" \
    --arg m "$(git show HEAD:.github/workflows/evals.yml | sed -nE 's/^[[:space:]]*MODELO_DE_EVALS:[[:space:]]*([^[:space:]#]+).*/\1/p' | head -1)" \
    --arg r "$(git show "HEAD:evals/$1/juez/rubrica.md" | huella_de_la_entrada)" \
    --arg c "$(git show "HEAD:evals/$1/juez/casos.yaml" | huella_de_la_entrada)" \
    --argjson medida "$medida" --argjson clases "$(clases "$1" HEAD)" \
    '{arbol_de_la_skill: $arbol, modelo_que_decide: $m,
      juez: {modelo: $medida.modelo_del_juez, version_de_claude_code: $medida.version_de_claude_code,
             rubrica_sha256: $r, casos_sha256: $c}, clases: $clases}'
}

# Todas las filas de la serie, una por línea: la semilla y las de cada hito.
todas_las_filas() {
  local f
  for f in "$serie_previa" specs/*/gates/serie.jsonl; do [ -f "$f" ] && cat "$f"; done
  return 0
}

# La serie de una skill y su tasa acumulada, en Markdown: por cada clase del juez que
# decide en la cabeza, una fila por medición y, debajo, la tasa acumulada de las filas
# que cuentan con la clave de la cabeza y, con 500 respuestas o más, su escalón.
mostrar_skill() { # $1 = skill
  local skill="$1" clave
  clave=$(clave_de_la_cabeza "$skill")
  todas_las_filas | jq -rs --arg skill "$skill" --argjson clave "$clave" --argjson minimo "$minimo_de_respuestas" \
    --argjson admitidos "$rojos_admitidos" "$jq_regla"'
    def corto: if . == null then "—" else .[0:7] end;
    def nombre: "\(.procedencia.hito)-\(.procedencia.medicion)";
    def de_la_clase($cl): [.clases[] | select(.clase == $cl.nombre)][0];
    def celda($c; f): ($c.celdas | map(select(f))) as $x | ($x | map(.total) | add // 0) as $t
      | if $t == 0 then "—" else "\($x | map(.marcadas) | add)/\($t)" end;
    def de_la_clave: .arbol_de_la_skill == $clave.arbol_de_la_skill and .modelo_que_decide == $clave.modelo_que_decide
      and .juez == $clave.juez;
    def en_la_tasa:
      if (.cuenta | not) then "no: " + .no_cuenta_porque
      elif .arbol_de_la_skill != $clave.arbol_de_la_skill then "no: otra skill"
      elif .modelo_que_decide != $clave.modelo_que_decide then "no: otro modelo que decide"
      elif .juez != $clave.juez then "no: otro juez"
      else "sí" end;
    def linea($cl): de_la_clase($cl) as $c
      | "| \(nombre) | \(.procedencia.fecha[0:10]) | `\(.procedencia.commit | corto)` | `\(.arbol_de_la_skill | corto)` | "
        + (if $c == null then "— | — | — | — | —"
           else [celda($c; .modo == "orden"), celda($c; .modo == "herramienta"), celda($c; .por_materia | not),
                 celda($c; .por_materia), celda($c; true)] | join(" | ") end)
        + " | \(en_la_tasa) |";
    def regla($cl; $suman; $m; $n):
      ($m / $n) as $tasa | ($suman[-1].c.total) as $hoy
      | ((($cl.umbral * $hoy) + 1e-9) | floor) as $k_hoy
      | escalon($hoy; $tasa; $admitidos) as $k
      | "Escalón de la regla 3 con \($hoy) respuestas, el total de hoy: como mucho \($k) marcada" + (if $k == 1 then "" else "s" end) + " (vale hasta una tasa del \(pct(tasa_maxima($hoy; $k; $admitidos); 2))). El umbral de hoy, \($cl.umbral) en `evals/\($skill)/juez/clases.yaml`, admite \($k_hoy) de \($hoy).",
        (if $k < $k_hoy then
           "Permite estrechar el umbral a \($k) de \($hoy). Lo baja una persona, con un diff de `evals/\($skill)/juez/clases.yaml` y estas cifras (ADR 0038, punto 4); ningún run lo cambia."
         elif $k > $k_hoy then
           "**Aviso**: la tasa acumulada, \(pct($tasa; 2)), pasa de la que vale para el umbral de hoy (\($k_hoy) de \($hoy), hasta el \(pct(tasa_maxima($hoy; $k_hoy; $admitidos); 2))): con ella, una ejecución sale en rojo sin motivo más de 1 de cada 50 veces. Este informe no propone cambiar el umbral: lo que hay que mirar es la skill (ADR 0038, puntos 4 y 8)."
         else "El umbral de hoy es el de la regla." end);
    def tasa($filas; $cl):
      [$filas[] | select(.cuenta and de_la_clave) | {nombre: nombre, c: de_la_clase($cl)} | select(.c != null)] as $suman
      | ($suman | map(.c.marcadas) | add // 0) as $m | ($suman | map(.c.total) | add // 0) as $n
      | if $n == 0 then
          "Tasa acumulada con la skill como está: ninguna medición completa con la clave de la cabeza. Insuficiente: 0 de \($minimo) respuestas (ADR 0038, punto 3)."
        else
          "Tasa acumulada con la skill como está: \($m) de \($n) (\(pct($m / $n; 2))), de \($suman | map(.nombre) | if length > 1 then (.[0:-1] | join(", ")) + " y " + .[-1] else .[0] end).",
          (if $n < $minimo then
             "Insuficiente: \($n) de \($minimo) respuestas. La regla 3 del ADR 0038 pide la tasa de \($minimo) respuestas al menos: con menos no da escalón."
           else regla($cl; $suman; $m; $n) end)
        end;
    def tabla($filas; $cl):
      "`\($cl.nombre)`:", "",
      "| Medición | Fecha | Commit | Skill | Orden | Herramienta | Nombra | Por materia | Total | En la tasa |",
      "|---|---|---|---|---|---|---|---|---|---|",
      ($filas[] | linea($cl)),
      "",
      tasa($filas; $cl),
      "";
    ([.[] | select(.skill == $skill)] | sort_by(.procedencia.fecha, .procedencia.hito, .procedencia.medicion)) as $filas
    | [$clave.clases[] | select(.decide)] as $deciden
    | "**Serie de cierres de \($skill)** (ADR 0038, punto 6): respuestas de las evals que activan la skill que el juez marca con sus tres votos, por modo y por clase de pregunta.",
      "",
      "Clave de la cabeza: skill `\($clave.arbol_de_la_skill | corto)` (árbol de `skills/\($skill)`), decide `\($clave.modelo_que_decide)`, juez `\($clave.juez.modelo)` con Claude Code \($clave.juez.version_de_claude_code).",
      "",
      (if ($filas | length) == 0 then "Sin mediciones de cierre todavía.", ""
       elif ($deciden | length) == 0 then "Ninguna clase del juez decide.", ""
       else ($deciden[] as $cl | tabla($filas; $cl)) end)'
}

sub="${1:?uso: serie.sh fila|anotar|semilla|mostrar|escalones|clases|por-materia …}"
shift
case "$sub" in
  fila) fila "${1:?skill}" "${2:?procedencia}" "${3:-}";;
  anotar) anotar "${1:?serie.jsonl}" "${2:?fila}";;
  semilla) semilla "${1:?fuentes.tsv}";;
  escalones) escalones "${1:?total}";;
  clases) clases "${1:?skill}" "${2:-}";;
  por-materia) por_materia "${1:?skill}" "${2:-}";;
  mostrar)
    if [ -n "${1:-}" ]; then mostrar_skill "$1"
    else
      for c in evals/*/juez/clases.yaml; do [ -f "$c" ] || continue; c=${c#evals/}; mostrar_skill "${c%%/*}"; done
    fi;;
  *) echo "serie.sh: subcomando desconocido: $sub" >&2; exit 2;;
esac
