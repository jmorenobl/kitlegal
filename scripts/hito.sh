#!/usr/bin/env bash
# Lanza y supervisa el workflow spec-kit `hito` para un hito del roadmap.
#
#   scripts/hito.sh H2                      # desatendido (gates automáticos), con supervisor
#   scripts/hito.sh H2 supervisado          # además pausa para revisión humana
#   scripts/hito.sh --resume <run_id> [key=value ...]
#   scripts/hito.sh --clasificar <run_id>   # qué haría el supervisor con ese run, sin hacerlo
#   KITLEGAL_MODELO_JUEZ=fable@max KITLEGAL_MODELO_IMPLEMENTACION=sonnet@xhigh scripts/hito.sh H2
#
# Claude se ejecuta en modo headless (`claude -p`). Los permisos de herramientas
# vienen de .claude/settings.json; aquí solo se aceptan las ediciones de
# ficheros. Para un entorno aislado (contenedor/VM) puede sustituirse por
# --dangerously-skip-permissions.
#
# Supervisor. `specify workflow run` termina en cuanto un paso falla o un gate
# pausa; el motor no reintenta nada. Este script clasifica cada parada leyendo
# state.json, log.jsonl y el transcript de la sesión headless, y actúa con una
# lista cerrada de acciones (docs/WORKFLOW.md «Supervisor»):
#
#   limite       la sesión de Claude murió por límite de uso o error de la API →
#                cambia de familia de modelo (KITLEGAL_MODELO_FALLBACK, por
#                defecto fable=opus,opus=sonnet) en todos los roles que usaban la
#                familia que falló, espera KITLEGAL_ESPERA_LIMITE segundos (60) y
#                reanuda.
#   transitorio  un paso shell de commit, o una sesión de Claude que terminó con
#                error sin ser límite de uso → reanuda una vez; el mismo paso no
#                se reanuda dos veces.
#   deliberado   precheck_*, check_*, leer_*, guardian_*, siguiente_tarea
#                (intentos agotados), redelimitada_reparacion, ci_final,
#                publicar_rama… fallan a propósito → se detiene y dice por qué.
#   gate         pausa humana ([datos] con material existente o fixtures, rutas
#                sensibles, modo supervisado) → se detiene.
#
# Tope: KITLEGAL_MAX_REANUDACIONES (8) por invocación. El supervisor nunca edita
# artefactos, veredictos ni permisos: solo relanza `specify workflow resume`.
# Compatible con bash 3.2 (macOS): sin mapfile ni arrays asociativos.
set -euo pipefail

cd "$(dirname "$0")/.."

export SPECKIT_INTEGRATION_CLAUDE_EXTRA_ARGS="${SPECKIT_INTEGRATION_CLAUDE_EXTRA_ARGS:---permission-mode acceptEdits}"
# Los inputs modelo_* admiten <modelo>@<esfuerzo>; el wrapper lo traduce a
# --model/--effort. Sin él, `claude` rechazaría el valor.
export SPECKIT_INTEGRATION_CLAUDE_EXECUTABLE="${SPECKIT_INTEGRATION_CLAUDE_EXECUTABLE:-$PWD/scripts/claude-modelo.sh}"

runs=.specify/workflows/runs
max_reanudaciones="${KITLEGAL_MAX_REANUDACIONES:-8}"
fallback="${KITLEGAL_MODELO_FALLBACK:-fable=opus,opus=sonnet}"
espera_limite="${KITLEGAL_ESPERA_LIMITE:-60}"
transcripts="${KITLEGAL_TRANSCRIPTS:-$HOME/.claude/projects/$(printf '%s' "$PWD" | sed -E 's/[^A-Za-z0-9]/-/g')}"
# Entrypoints con los que se reconoce una sesión headless en los transcripts.
# scripts/claude-modelo.sh fija sdk-cli; para runs grabados antes de esa
# corrección desde la extensión de VS Code: KITLEGAL_ENTRYPOINTS_HEADLESS=sdk-cli,claude-vscode.
entrypoints_headless="${KITLEGAL_ENTRYPOINTS_HEADLESS:-sdk-cli}"

log() { printf 'hito.sh: %s\n' "$*" >&2; }
avisar() { # notificación de escritorio si existe; nunca falla
  if command -v osascript >/dev/null 2>&1; then
    osascript -e "display notification \"$1\" with title \"kitlegal · hito\"" >/dev/null 2>&1 || true
  fi
}

# ---------------------------------------------------------------- lectura de un run
estado_run() { jq -r '.status // ""' "$runs/$1/state.json"; }
paso_actual() { jq -r '.current_step_id // ""' "$runs/$1/state.json"; }
ultimo_fallo() { jq -c 'select(.event == "step_failed")' "$runs/$1/log.jsonl" | tail -1; }
tipo_paso() { jq -r --arg s "$2" '.step_results[$s].type // ""' "$runs/$1/state.json"; }
modelo_paso() { jq -r --arg s "$2" '.step_results[$s].model // ""' "$runs/$1/state.json"; }
salida_paso() { jq -r --arg s "$2" '.step_results[$s].output | ((.stdout // "") + "\n" + (.stderr // ""))' "$runs/$1/state.json" | grep -v '^$' | tail -6 | tr '\n' ' '; }
inicio_paso() { jq -r --arg s "$2" 'select(.event == "step_started" and .step_id == $s) | .timestamp' "$runs/$1/log.jsonl" | tail -1; }
# "ronda_spec:juez_spec:1" → "juez_spec"; "bucle_tareas:siguiente_tarea:23" → "siguiente_tarea"
nombre_paso() { printf '%s' "$1" | tr ':' '\n' | grep -vE '^[0-9]+$' | tail -1; }

# ¿Alguna sesión headless iniciada después de $1 (ISO 8601) terminó por límite de
# uso o error de la API? Se lee del transcript de Claude Code, porque spec-kit no
# conserva la salida de `claude -p`.
limite_api() {
  [ -n "$1" ] && [ -d "$transcripts" ] || return 1
  python3 - "$transcripts" "$1" "$entrypoints_headless" <<'PYEOF'
import datetime, glob, json, os, re, sys
tdir, desde, entrypoints = sys.argv[1:4]
t0 = datetime.datetime.fromisoformat(desde.replace("Z", "+00:00")).timestamp()
patron = re.compile(r"rate_limit|spend limit|usage limit|rate limit|overloaded|api_error|internal server error|API Error:? ?\d{3}\b", re.I)
for f in glob.glob(os.path.join(tdir, "*.jsonl")):
    if os.path.getmtime(f) < t0:
        continue
    lineas = open(f).readlines()
    # la sesión headless se reconoce por el entrypoint del primer mensaje de usuario
    # (las primeras líneas pueden ser operaciones de cola)
    cabeza = "".join(lineas[:10]).replace(" ", "")
    # Una sesión interactiva de Claude Code lleva origin.kind = human en su primer
    # mensaje; las headless no traen origin. Importa cuando el entrypoint admitido
    # es el de la extensión de VS Code, que comparten unas y otras.
    if not any('"entrypoint":"%s"' % ep in cabeza for ep in entrypoints.split(",")) \
            or '"origin":{"kind":"human"}' in cabeza:
        continue
    cola = lineas[-40:]
    for l in cola:
        try:
            e = json.loads(l)
        except ValueError:
            continue
        if e.get("type") != "assistant":
            continue
        textos = [str(e.get("error", ""))]
        for c in (e.get("message") or {}).get("content") or []:
            if isinstance(c, dict):
                textos.append(str(c.get("text", "")))
        if any(patron.search(t) for t in textos):
            sys.exit(0)
sys.exit(1)
PYEOF
}

# fable@xhigh → fable · claude-opus-5 → opus · claude-haiku-4-5 → haiku
familia() { printf '%s' "$1" | sed -E 's/@.*$//; s/^claude-//; s/-[0-9].*$//'; }
sustituto() { printf '%s' "$fallback" | tr ',' '\n' | awk -F= -v f="$1" '$1 == f {print $2}' | head -1; }

# Imprime, una por línea, las parejas "--input" / "modelo_<rol>=<nuevo>[@esfuerzo]"
# para cada rol de inputs.json cuya familia sea la que falló.
inputs_fallback() {
  local run="$1" fam="$2" nuevo="$3" rol val esf
  jq -r '.inputs | to_entries[] | select(.key | startswith("modelo_")) | "\(.key)=\(.value)"' "$runs/$run/inputs.json" |
    while IFS='=' read -r rol val; do
      [ "$(familia "$val")" = "$fam" ] || continue
      esf=""; case "$val" in *@*) esf="@${val#*@}";; esac
      [ "$nuevo" = haiku ] && esf=""
      printf -- '--input\n%s=%s%s\n' "$rol" "$nuevo" "$esf"
    done
}

# clase<TAB>paso<TAB>detalle
clasificar() {
  local run="$1" st fallo paso err tipo nombre
  st=$(estado_run "$run")
  case "$st" in
    completed) printf 'completado\t\t\n'; return 0;;
    paused) printf 'gate\t%s\t\n' "$(paso_actual "$run")"; return 0;;
    failed) ;;
    *) printf 'desconocido\t%s\testado %s\n' "$(paso_actual "$run")" "$st"; return 0;;
  esac
  fallo=$(ultimo_fallo "$run")
  paso=$(jq -r '.step_id // ""' <<<"$fallo"); err=$(jq -r '.error // ""' <<<"$fallo")
  tipo=$(tipo_paso "$run" "$paso"); nombre=$(nombre_paso "$paso")
  case "$tipo" in
    prompt|command)
      if limite_api "$(inicio_paso "$run" "$paso")"; then
        printf 'limite\t%s\t%s\n' "$paso" "$(modelo_paso "$run" "$paso")"
      else
        printf 'transitorio\t%s\t%s\n' "$paso" "$err"
      fi;;
    shell)
      case "$nombre" in
        commit_*) printf 'transitorio\t%s\t%s\n' "$paso" "$(salida_paso "$run" "$paso")";;
        *) printf 'deliberado\t%s\t%s\n' "$paso" "$(salida_paso "$run" "$paso")";;
      esac;;
    *) printf 'deliberado\t%s\t%s\n' "$paso" "$err";;
  esac
}

# ---------------------------------------------------------------- bucle del supervisor
# $1: run_id ("" para lanzar uno nuevo); el resto, argumentos de `specify workflow run|resume`.
supervisar() {
  local run="$1"; shift
  local n=0 clase paso detalle fam nuevo reanudados=" " extra
  # macOS: el run de H3 perdió 103 min con el Mac dormido a mitad de `plan`
  # («Connection lost while your computer was asleep»). caffeinate -i impide el
  # reposo por inactividad mientras viva este proceso (-w lo liga a este PID);
  # la pantalla sí puede apagarse.
  if command -v caffeinate >/dev/null 2>&1; then caffeinate -i -w $$ & fi
  while :; do
    set +e
    if [ -z "$run" ]; then
      specify workflow run hito "$@"
      run=$(ls -t "$runs" | head -1)
    else
      specify workflow resume "$run" "$@"
    fi
    set -e
    set -- # los inputs solo se aplican una vez
    [ -n "$run" ] && [ -f "$runs/$run/state.json" ] || { log "no encuentro el estado del run"; return 1; }
    IFS=$'\t' read -r clase paso detalle < <(clasificar "$run")
    case "$clase" in
      completado)
        log "run $run completado"; avisar "Hito completado ($run)"; return 0;;
      gate)
        log "run $run en pausa humana en '$paso'. Revisa y reanuda con: scripts/hito.sh --resume $run"
        avisar "Pausa humana en $paso ($run)"; return 3;;
      deliberado|desconocido)
        log "run $run detenido a propósito en '$paso': $detalle"
        log "corrige y reanuda con: scripts/hito.sh --resume $run"
        avisar "Parada en $paso ($run)"; return 1;;
    esac
    n=$((n + 1))
    if [ "$n" -gt "$max_reanudaciones" ]; then
      log "agotadas $max_reanudaciones reanudaciones automáticas (última: $clase en '$paso'); reanuda a mano: scripts/hito.sh --resume $run"
      avisar "Supervisor agotado en $paso ($run)"; return 1
    fi
    case "$clase" in
      limite)
        fam=$(familia "$detalle"); nuevo=$(sustituto "$fam")
        if [ -z "$nuevo" ]; then
          log "límite de uso en '$paso' con $detalle y sin repuesto para '$fam' (KITLEGAL_MODELO_FALLBACK=$fallback)"
          avisar "Límite de uso sin repuesto ($run)"; return 1
        fi
        extra=()
        while IFS= read -r l; do extra+=("$l"); done < <(inputs_fallback "$run" "$fam" "$nuevo")
        log "límite de uso en '$paso' ($detalle): $fam → $nuevo en los roles afectados; espera ${espera_limite}s y reanuda ($n/$max_reanudaciones)"
        sleep "$espera_limite"
        set -- ${extra[@]+"${extra[@]}"};;
      transitorio)
        case "$reanudados" in
          *" $paso "*)
            log "'$paso' volvió a fallar tras reanudarlo: $detalle"
            log "reanuda a mano: scripts/hito.sh --resume $run"; avisar "Fallo repetido en $paso ($run)"; return 1;;
        esac
        reanudados="$reanudados$paso "
        log "fallo transitorio en '$paso' ($detalle); reanuda ($n/$max_reanudaciones)";;
    esac
  done
}

# ---------------------------------------------------------------- entrada
case "${1:-}" in
  --clasificar)
    clasificar "${2:?run_id}"; exit 0;;
  --resume)
    run_id="${2:?run_id}"; shift 2
    args=()
    for kv in "$@"; do args+=(--input "$kv"); done
    supervisar "$run_id" ${args[@]+"${args[@]}"}; exit $?;;
esac

hito="${1:?uso: scripts/hito.sh H<n> [desatendido|supervisado] | --resume <run_id> [k=v …] | --clasificar <run_id>}"
modo="${2:-desatendido}"

if [ -n "$(git status --porcelain)" ]; then
  echo "el árbol de trabajo tiene cambios sin commitear; el hito parte de main limpio" >&2
  exit 1
fi
if [ "$(git branch --show-current)" != "main" ]; then
  echo "sitúate en main antes de lanzar un hito (rama actual: $(git branch --show-current))" >&2
  exit 1
fi

# Modelo y esfuerzo por rol (<alias o nombre completo>[@esfuerzo]; roles en
# docs/WORKFLOW.md «Modelo por paso»). Ejemplo:
#   KITLEGAL_MODELO_IMPLEMENTACION=opus@max scripts/hito.sh H2
modelos=()
for rol in DECISION JUEZ REVISOR REDACCION IMPLEMENTACION ESCALADA ANALISIS; do
  var="KITLEGAL_MODELO_$rol"
  if [ -n "${!var:-}" ]; then
    modelos+=(--input "modelo_$(printf '%s' "$rol" | tr '[:upper:]' '[:lower:]')=${!var}")
  fi
done

supervisar "" --input "hito=$hito" --input "modo=$modo" ${modelos[@]+"${modelos[@]}"}
