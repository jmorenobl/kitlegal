#!/usr/bin/env bash
# Pasos shell del bucle de tareas del workflow `hito` (docs/WORKFLOW.md «Implementación tarea a tarea»).
#
#   scripts/workflow/tarea.sh siguiente <hito>   # elige la tarea de esta iteración → JSON
#   scripts/workflow/tarea.sh redelimitada       # ¿el ejecutor redelimitó la tarea en este intento? → JSON
#   scripts/workflow/tarea.sh fallo              # anota en gates/tarea-Tnnn.md por qué el intento sigue en rojo
#   scripts/workflow/tarea.sh estado             # ¿se commitea esta iteración? → JSON
#   scripts/workflow/tarea.sh commit <hito>      # commit feat(<hito>): Tnnn
#
# Ninguno de estos pasos detiene el run por una tarea que no sale: tras tres
# intentos (o dos iteraciones de cierre que no commitean) la tarea entra en
# CUARENTENA —su trabajo se guarda como parche en gates/cuarentena/Tnnn.patch,
# el árbol vuelve al último commit en verde y el bucle sigue con la siguiente—.
# La única parada es una causa mayor determinista: tres cuarentenas seguidas sin
# ninguna tarea en verde entre medias, que es la señal de que el resto de
# tasks.md depende de lo que está en cuarentena (ADR 0018, «DAG bloqueado»).
set -euo pipefail
cd "$(dirname "$0")/../.."
. scripts/workflow/comun.sh

d=$(feature_dir)
mkdir -p "$d/gates"
intentos="$d/gates/tareas-intentos.json"; [ -f "$intentos" ] || echo '{}' > "$intentos"
cuarentena="$d/gates/cuarentena.json"; [ -f "$cuarentena" ] || echo '{"tareas":{}, "consecutivas":0}' > "$cuarentena"
actual="$d/gates/tarea-actual.json"
max_intentos=3
max_cierres=2
max_consecutivas=3

nota_de() { printf '%s/gates/tarea-%s.md' "$d" "$1"; }
huella_nota() { if [ -f "$(nota_de "$1")" ]; then huella "$(nota_de "$1")"; else echo -; fi; }
pendiente() { grep -qE "^[[:space:]]*- \[ \] $1([^0-9]|\$)" "$d/tasks.md"; }
marcada() { grep -qE "^[[:space:]]*- \[[Xx]\] $1([^0-9]|\$)" "$d/tasks.md"; }
en_cuarentena() { jq -e --arg id "$1" '.tareas | has($id)' "$cuarentena" >/dev/null; }
contar() { # $1 clave en tareas-intentos.json → imprime el valor incrementado
  local n
  n=$(jq -r --arg k "$1" '.[$k] // 0' "$intentos"); n=$((n + 1))
  jq --arg k "$1" --argjson n "$n" '.[$k] = $n' "$intentos" > "$intentos.tmp" && mv "$intentos.tmp" "$intentos"
  echo "$n"
}

# Guarda el trabajo sin commitear de la tarea como parche, devuelve el árbol al
# último commit (fuera del directorio del feature, que conserva tasks.md, notas y
# veredictos) y la registra. Sin commits de por medio, HEAD es la base de la tarea.
poner_en_cuarentena() {
  local id="$1" motivo="$2" parche k
  parche="$d/gates/cuarentena/$id.patch"
  aparcar_desde "$(git rev-parse HEAD)" "$parche"
  jq --arg id "$id" --arg m "$motivo" --arg p "$parche" --arg b "$(git rev-parse HEAD)" \
     --arg fecha "$(date -u +%Y-%m-%dT%H:%M:%SZ)" --argjson n "$(jq -r --arg id "$id" --argjson m "$max_intentos" '[(.[$id] // 0), $m] | min' "$intentos")" \
     '.tareas[$id] = {motivo:$m, intentos:$n, parche:$p, base:$b, fecha:$fecha} | .consecutivas += 1' \
     "$cuarentena" > "$cuarentena.tmp" && mv "$cuarentena.tmp" "$cuarentena"
  anotar_supuesto "Tarea $id en cuarentena: $motivo. Su trabajo está en $parche y la causa en $(nota_de "$id")."
  echo "tarea $id en CUARENTENA: $motivo; el bucle sigue con la siguiente" >&2
  k=$(jq -r .consecutivas "$cuarentena")
  if [ "$k" -ge "$max_consecutivas" ]; then
    jq '.bloqueado = true' "$cuarentena" > "$cuarentena.tmp" && mv "$cuarentena.tmp" "$cuarentena"
    dossier_dag_bloqueado; exit 3
  fi
}

dossier_dag_bloqueado() {
  local f="$d/gates/dossier.md"
  {
    echo "# Dossier: DAG bloqueado"
    echo
    echo "Causa mayor (ADR 0018): $max_consecutivas tareas seguidas en cuarentena sin ninguna en verde entre medias."
    echo "Lo más probable es que las tareas pendientes dependan de lo que está en cuarentena."
    echo
    echo "## Tareas en cuarentena"
    jq -r '.tareas | to_entries[] | "- \(.key): \(.value.motivo) (\(.value.intentos) intentos; parche \(.value.parche))"' "$cuarentena"
    echo
    echo "## Tareas pendientes"
    grep -E '^[[:space:]]*- \[ \] T[0-9]+' "$d/tasks.md" | cut -c1-200 || true
    echo
    echo "## Qué hace falta para seguir"
    echo "Leer la nota gates/tarea-Tnnn.md de cada tarea en cuarentena. Si la causa está en el spec o en el roadmap,"
    echo "se corrige ahí y se relanza el hito. Si es del entorno (herramienta, red, permisos), se arregla y se reanuda"
    echo "con scripts/hito.sh --resume <run_id>. Al reanudar, el contador de cuarentenas seguidas vuelve a cero; las"
    echo "tareas que sigan en gates/cuarentena.json no se reintentan: para reintentar una, se borra su entrada."
  } > "$f"
  echo "DAG bloqueado: $max_consecutivas cuarentenas seguidas. Dossier en $f" >&2
}

sin_tarea() { # $1: converger (true|false)
  jq -n --argjson c "$1" '{id:"", texto:"", intento:0, datos:false, aceptacion:false, base:"", rutas:[], cierre:false, converger:$c, nota:"-"}' | tee "$actual"
}

case "${1:?uso: tarea.sh siguiente <hito> | redelimitada | fallo | estado | commit <hito>}" in
  siguiente)
    # Reanudar tras un DAG bloqueado es la decisión de seguir: el contador vuelve a cero.
    if jq -e '.bloqueado == true' "$cuarentena" >/dev/null; then
      jq '.consecutivas = 0 | .bloqueado = false' "$cuarentena" > "$cuarentena.tmp" && mv "$cuarentena.tmp" "$cuarentena"
      echo "reanudación tras DAG bloqueado: cuarentenas seguidas a cero" >&2
    fi
    # Reanudación dentro del bucle: el motor no retoma el paso interno que falló sino
    # que reejecuta el paso padre entero, es decir, arranca una iteración nueva
    # (engine.py: «resume will re-run the parent step and its nested body»). Si la
    # tarea anterior quedó marcada [X] con su trabajo sin commitear, esta iteración la
    # cierra —verificación y commit con su base original, sin implementar nada— en vez
    # de atribuir su diff a la siguiente. Dos cierres que no commitean bastan para saber
    # que no avanza: sin implementador de por medio, nada cambia entre uno y otro.
    if [ -f "$actual" ]; then
      prev=$(jq -r '.id // ""' "$actual")
      if [ -n "$prev" ] && marcada "$prev" && [ -n "$(git status --porcelain | grep -v " $d/" || true)" ]; then
        c=$(contar "$prev:cierre")
        if [ "$c" -gt "$max_cierres" ]; then
          # La marca [X] no vale sin commit: se deshace para que la tarea no cuente como hecha.
          sed -i.bak -E "s/^([[:space:]]*- )\[[Xx]\] $prev([^0-9]|\$)/\1[ ] $prev\2/" "$d/tasks.md" && rm -f "$d/tasks.md.bak"
          poner_en_cuarentena "$prev" "marcada [X] pero $max_cierres iteraciones de cierre no dejaron la verificación en verde"
        else
          echo "la tarea $prev está marcada [X] con cambios sin commitear: esta iteración la cierra (cierre $c de $max_cierres)" >&2
          jq '. + {cierre: true}' "$actual" > "$actual.tmp" && mv "$actual.tmp" "$actual"
          cat "$actual"; exit 0
        fi
      fi
    fi
    hito="${2:?uso: tarea.sh siguiente <hito>}"
    while :; do
      linea=""
      while IFS= read -r l; do
        i=$(printf '%s' "$l" | grep -oE 'T[0-9]+' | head -1)
        en_cuarentena "$i" || { linea="$l"; break; }
      done < <(grep -E '^[[:space:]]*- \[ \] T[0-9]+' "$d/tasks.md" || true)
      if [ -z "$linea" ]; then
        if [ -f "$d/gates/converge-hecho" ]; then sin_tarea false; else sin_tarea true; fi
        exit 0
      fi
      id=$(printf '%s' "$linea" | grep -oE 'T[0-9]+' | head -1)
      n=$(contar "$id")
      if [ "$n" -gt "$max_intentos" ]; then
        poner_en_cuarentena "$id" "$max_intentos intentos sin quedar en verde"
        continue
      fi
      break
    done
    texto=$(printf '%s' "$linea" | sed -E 's/^[[:space:]]*- \[ \] //')
    datos=false; printf '%s' "$texto" | grep -q '\[datos\]' && datos=true
    aceptacion=false; printf '%s' "$texto" | grep -q '\[aceptacion\]' && aceptacion=true
    base=$(git rev-parse HEAD)
    rutas=$(printf '%s' "$texto" | tr '`,;()' '     ' | grep -oE '(\.?[A-Za-z0-9_-]+/)+[A-Za-z0-9_.*-]*|[A-Za-z0-9_.-]+\.(go|md|yml|yaml|json|sh|txt|mod|sum|toml|txtar)|\.[A-Za-z0-9_-]*ignore|\.editorconfig|Makefile|LICENSE' | grep -vE '^https?:' | sort -u | jq -R . | jq -s .)
    jq -n --arg id "$id" --arg texto "$texto" --argjson n "$n" --argjson datos "$datos" --argjson aceptacion "$aceptacion" \
      --arg base "$base" --argjson rutas "$rutas" --arg nota "$(huella_nota "$id")" \
      '{id:$id, texto:$texto, intento:$n, datos:$datos, aceptacion:$aceptacion, base:$base, rutas:$rutas, cierre:false, converger:false, nota:$nota}' | tee "$actual";;

  redelimitada)
    # Redelimitada = sigue [ ] y su nota cambió en ESTE intento (huella distinta de la
    # que tenía al elegirse la tarea). El intento siguiente relee la línea de tasks.md,
    # con sus rutas nuevas, y hereda el árbol; no se repara ni se commitea.
    id=$(jq -r .id "$actual"); r=false
    if pendiente "$id" && [ "$(huella_nota "$id")" != "$(jq -r '.nota // "-"' "$actual")" ]; then r=true; fi
    [ "$r" = true ] && echo "la tarea $id queda [ ] con $(nota_de "$id") escrito en este intento: el bucle la reintenta con la línea actual de tasks.md" >&2
    jq -n --argjson r "$r" '{redelimitada: $r}';;

  fallo)
    # El intento termina en rojo tras reparar: el diagnóstico va a la nota de la tarea,
    # que el intento siguiente (con modelo_escalada) lee antes de empezar.
    id=$(jq -r .id "$actual"); n=$(jq -r .intento "$actual"); nota=$(nota_de "$id")
    que="Intento $n"; [ "$(jq -r .cierre "$actual")" = true ] && que="Cierre (marcada [X] sin commitear) tras el intento $n"
    {
      [ -f "$nota" ] || printf '# %s: por qué no quedó en verde\n' "$id"
      printf '\n## %s: verificación en rojo tras reparar (%s)\n\n```\n' "$que" "$(date -u +%Y-%m-%dT%H:%M:%SZ)"
      grep -v '^kitlegal-verificacion exit=' "$d/gates/ci.log" | tail -60
      printf '```\n'
    } >> "$nota"
    # La nota escrita aquí no es una redelimitación del ejecutor: se actualiza la huella.
    jq --arg h "$(huella_nota "$id")" '.nota = $h' "$actual" > "$actual.tmp" && mv "$actual.tmp" "$actual"
    echo "tarea $id: intento $n en rojo; diagnóstico en $nota";;

  estado)
    # Decide el commit con el estado real de ESTA iteración —última verificación del log
    # y tasks.md en el árbol—, nunca con salidas de pasos que quizá no se ejecutaron: el
    # motor conserva la última salida de cada id entre iteraciones (H4, T003: 145 vueltas).
    id=$(jq -r .id "$actual")
    rc=$(tail -n 1 "$d/gates/ci.log" 2>/dev/null | sed -n 's/^kitlegal-verificacion exit=\([0-9][0-9]*\)$/\1/p')
    [ -n "$rc" ] || { echo "sin resultado de verificación en $d/gates/ci.log: no se commitea" >&2; rc=1; }
    c=false; [ "$rc" -eq 0 ] && marcada "$id" && c=true
    echo "tarea $id: verificación exit=$rc, marcada=$(marcada "$id" && echo sí || echo no), commitear=$c" >&2
    jq -n --argjson c "$c" --argjson rc "$rc" '{commitear:$c, verificacion:$rc}';;

  commit)
    hito="${2:?uso: tarea.sh commit <hito>}"
    id=$(jq -r .id "$actual")
    jq '.consecutivas = 0' "$cuarentena" > "$cuarentena.tmp" && mv "$cuarentena.tmp" "$cuarentena"
    git add -A
    git diff --cached --quiet && { echo "tarea $id: nada que commitear"; exit 0; }
    git commit -q -m "feat($hito): $id" -m "$(jq -r .texto "$actual")"
    git log --oneline -1;;

  *) echo "subcomando desconocido: $1" >&2; exit 2;;
esac
