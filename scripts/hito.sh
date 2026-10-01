#!/usr/bin/env bash
# Lanza y supervisa el workflow spec-kit `hito` para un hito del roadmap.
#
#   scripts/hito.sh H7                      # lanza el hito y lo lleva hasta el informe final
#   scripts/hito.sh --resume <run_id> [key=value ...]
#   scripts/hito.sh --clasificar <run_id>   # qué haría el supervisor con ese run, sin hacerlo
#   KITLEGAL_MODELO_JUEZ=fable@max KITLEGAL_MODELO_IMPLEMENTACION=sonnet@xhigh scripts/hito.sh H7
#
# La persona está en los extremos (ADR 0018): escribe la sección del hito en
# docs/ROADMAP.md antes de lanzar y lee el informe final (cuerpo de la propuesta
# de cambio) antes de fusionar. En medio no hay ninguna pausa humana ni ninguna
# parada que decida un modelo.
#
# Claude se ejecuta en modo headless (`claude -p`). Los permisos de herramientas
# vienen de .claude/settings.json; aquí solo se aceptan las ediciones de
# ficheros. Lo que una sesión del run no puede hacer —leer credenciales, abrir
# sesiones con modelo, dejar nada en segundo plano, escribir fuera del repositorio
# y del directorio temporal— lo impiden el sandbox con el que las abre
# scripts/claude-modelo.sh, el gancho de scripts/workflow/politica-paso.sh y las
# comprobaciones de `comprobar_entorno` (ADR 0032), no los prompts.
#
# Una sola sesión por run: al arrancar toma el candado del árbol de trabajo
# (scripts/workflow/sesion-unica.sh) y exporta su testigo; mientras vive, ninguna
# otra sesión de Claude Code puede editar ni cambiar el historial de este árbol.
#
# Supervisor. `specify workflow run` termina en cuanto un paso falla; el motor no
# reintenta nada. Este script clasifica cada parada leyendo state.json, log.jsonl
# y el transcript de la sesión headless, y actúa con una lista cerrada de
# acciones (docs/WORKFLOW.md «Supervisor»):
#
#   limite       la sesión de Claude murió por límite de uso o error de la API →
#                cambia de familia de modelo (KITLEGAL_MODELO_FALLBACK, por
#                defecto fable=opus,opus=sonnet) en todos los roles que usaban la
#                familia que falló, espera KITLEGAL_ESPERA_LIMITE segundos (60) y
#                reanuda; sin repuesto, espera KITLEGAL_ESPERA_SIN_REPUESTO (1800)
#                y reanuda con los mismos modelos. No consume reanudaciones: lo
#                acota el presupuesto de tiempo.
#   transitorio  cualquier otro fallo de un paso → reanuda; el mismo paso no se
#                reanuda más de dos veces.
#   rechazo      el juez de entrada rechazó el spec (check_gate_spec): el único
#                camino de vuelta a la persona antes del final; informe de
#                rechazo en gates/informe-rechazo.md.
#   causa_mayor  un paso shell salió con 3 y dejó gates/dossier.md: DAG
#                bloqueado, fuente sin revisión humana, dependencia externa
#                inaccesible o credencial ausente.
#   presupuesto  el run lleva más de KITLEGAL_TIEMPO_MAXIMO segundos (48 h) de
#                reloj desde su primer paso.
#
# Tope: KITLEGAL_MAX_REANUDACIONES (8) reanudaciones por fallo transitorio por
# invocación. El supervisor nunca edita artefactos, veredictos ni permisos: solo
# relanza `specify workflow resume`.
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
espera_sin_repuesto="${KITLEGAL_ESPERA_SIN_REPUESTO:-1800}"
tiempo_maximo="${KITLEGAL_TIEMPO_MAXIMO:-172800}"
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

salida_codigo() { jq -r --arg s "$2" '.step_results[$s].output.exit_code // ""' "$runs/$1/state.json"; }

# Segundos de reloj desde el primer paso del run.
edad_run() {
  local t0
  t0=$(head -1 "$runs/$1/log.jsonl" | jq -r .timestamp); t0=${t0%%.*}
  echo $(( $(date -u +%s) - $(date -j -u -f '%Y-%m-%dT%H:%M:%S' "$t0" +%s 2>/dev/null || date -u -d "$t0" +%s) ))
}

# clase<TAB>paso<TAB>detalle
clasificar() {
  local run="$1" st fallo paso err tipo nombre
  st=$(estado_run "$run")
  case "$st" in
    completed) printf 'completado\t\t\n'; return 0;;
    # Ya no hay gates humanos (ADR 0018); una pausa solo puede venir de un run anterior.
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
      if [ "$nombre" = check_gate_spec ]; then
        printf 'rechazo\t%s\t%s\n' "$paso" "$(salida_paso "$run" "$paso")"
      elif [ "$nombre" = extraer_hito ]; then
        printf 'entrada\t%s\t%s\n' "$paso" "$(salida_paso "$run" "$paso")"
      elif [ "$(salida_codigo "$run" "$paso")" = 3 ]; then
        printf 'causa_mayor\t%s\t%s\n' "$paso" "$(salida_paso "$run" "$paso")"
      else
        printf 'transitorio\t%s\t%s\n' "$paso" "$(salida_paso "$run" "$paso")"
      fi;;
    *) printf 'transitorio\t%s\t%s\n' "$paso" "$err";;
  esac
}

# ---------------------------------------------------------------- bucle del supervisor
# $1: run_id ("" para lanzar uno nuevo); el resto, argumentos de `specify workflow run|resume`.
supervisar() {
  local run="$1"; shift
  local n=0 clase paso detalle fam nuevo reanudados=" " extra veces d
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
    d=$(jq -r .feature_directory .specify/feature.json 2>/dev/null || echo "")
    case "$clase" in
      completado)
        log "run $run completado: el informe final es el cuerpo de la propuesta de cambio ($d/gates/informe-final.md)"
        avisar "Hito completado: lee el informe ($run)"; return 0;;
      rechazo)
        log "el juez de entrada rechazó el spec: informe de rechazo en $d/gates/informe-rechazo.md"
        log "corrige la sección del hito en docs/ROADMAP.md y relanza el hito desde main"
        avisar "Spec rechazado en la entrada ($run)"; return 4;;
      entrada)
        log "no se pudo leer la entrada del hito: $detalle"; avisar "Entrada inválida ($run)"; return 2;;
      causa_mayor)
        log "run $run detenido por causa mayor en '$paso': $detalle"
        log "dossier en $d/gates/dossier.md; al resolverlo: scripts/hito.sh --resume $run"
        avisar "Causa mayor en $paso ($run)"; return 3;;
      gate|desconocido)
        log "run $run en '$paso' con estado $clase: $detalle. Reanuda con: scripts/hito.sh --resume $run"
        avisar "Run parado en $paso ($run)"; return 1;;
    esac
    if [ "$(edad_run "$run")" -gt "$tiempo_maximo" ]; then
      log "presupuesto de tiempo agotado: el run $run supera ${tiempo_maximo}s de reloj (KITLEGAL_TIEMPO_MAXIMO); última parada: $clase en '$paso'"
      log "para seguir: KITLEGAL_TIEMPO_MAXIMO=<segundos> scripts/hito.sh --resume $run"
      avisar "Presupuesto de tiempo agotado ($run)"; return 3
    fi
    case "$clase" in
      limite)
        fam=$(familia "$detalle"); nuevo=$(sustituto "$fam")
        if [ -z "$nuevo" ]; then
          log "límite de uso en '$paso' con $detalle y sin repuesto para '$fam': espera ${espera_sin_repuesto}s y reanuda con los mismos modelos"
          sleep "$espera_sin_repuesto"
        else
          extra=()
          while IFS= read -r l; do extra+=("$l"); done < <(inputs_fallback "$run" "$fam" "$nuevo")
          log "límite de uso en '$paso' ($detalle): $fam → $nuevo en los roles afectados; espera ${espera_limite}s y reanuda"
          sleep "$espera_limite"
          set -- ${extra[@]+"${extra[@]}"}
        fi;;
      transitorio)
        n=$((n + 1))
        if [ "$n" -gt "$max_reanudaciones" ]; then
          log "agotadas $max_reanudaciones reanudaciones por fallo transitorio (última en '$paso': $detalle); reanuda a mano: scripts/hito.sh --resume $run"
          avisar "Supervisor agotado en $paso ($run)"; return 1
        fi
        veces=$(printf '%s' "$reanudados" | tr ' ' '\n' | grep -cx "$(nombre_paso "$paso")" || true)
        if [ "$veces" -ge 2 ]; then
          log "'$paso' falló por tercera vez: $detalle"
          log "reanuda a mano: scripts/hito.sh --resume $run"; avisar "Fallo repetido en $paso ($run)"; return 1
        fi
        reanudados="$reanudados$(nombre_paso "$paso") "
        log "fallo transitorio en '$paso' ($detalle); reanuda ($n/$max_reanudaciones)";;
    esac
  done
}

# Una sola sesión por run: el candado vive lo que viva este proceso.
tomar_candado() {
  local testigo
  testigo=$(KITLEGAL_PID_RUN=$$ scripts/workflow/sesion-unica.sh tomar) || exit 1
  export KITLEGAL_RUN_TESTIGO="$testigo"
  trap 'scripts/workflow/sesion-unica.sh soltar "$KITLEGAL_RUN_TESTIGO"' EXIT
}

# Credenciales que el run necesita antes del final: sin ellas, fallaría a las
# tantas horas en el cierre. Faltar una es causa mayor, y mejor al principio.
comprobar_credenciales() {
  command -v claude >/dev/null 2>&1 || { log "no encuentro 'claude' en el PATH"; exit 3; }
  command -v gh >/dev/null 2>&1 && gh auth status >/dev/null 2>&1 \
    || { log "gh no está instalado o no tiene sesión (gh auth status): el cierre no podría publicar ni medir"; exit 3; }
  git ls-remote --exit-code origin HEAD >/dev/null 2>&1 || { log "no se puede leer origin: el cierre no podría empujar la rama"; exit 3; }
}

# Lo que dejaría a una sesión del run usar una credencial que no es suya o escribir
# fuera de lo que el workflow declara (ADR 0032): en H7.4 `reparar_cierre` leyó de un
# fichero el token de la suscripción de quien lanzó el run y abrió con él 42 sesiones
# con modelo. Se comprueba antes de tomar el candado, y también al reanudar:
#   · ninguna credencial de Claude en el entorno, que heredarían las sesiones y todo
#     lo que ejecutan: el run usa la sesión de `claude` de quien lo lanza;
#   · el token del sondeo no está en un fichero: vive en su llavero, que pide su
#     contraseña a la persona cada vez (scripts/evals-sondeo-llavero.sh);
#   · ningún directorio adicional en los settings que leen las sesiones del run
#     (el del proyecto y el de la cuenta; .claude/settings.local.json no lo leen):
#     el sandbox de las sesiones los abriría a la escritura;
#   · el sandbox está disponible, y la política de las sesiones deniega lo que
#     tiene que denegar.
comprobar_entorno() {
  local v f
  for v in CLAUDE_CODE_OAUTH_TOKEN ANTHROPIC_API_KEY ANTHROPIC_AUTH_TOKEN; do
    if [ -n "${!v:-}" ]; then
      log "$v está en el entorno y la heredarían las sesiones del run: lanza el hito sin ella (unset $v)"; exit 1
    fi
  done
  if [ -e "$HOME/.config/kitlegal/claude-oauth-token" ]; then
    log "el token del sondeo está en un fichero que cualquier sesión puede leer ($HOME/.config/kitlegal/claude-oauth-token): guárdalo en su llavero y borra el fichero (docs/WORKFLOW.md, «Lo que una sesión de un paso no puede hacer»)"; exit 1
  fi
  for f in .claude/settings.json "$HOME/.claude/settings.json"; do
    [ -f "$f" ] || continue
    if [ "$(jq -r '(.permissions.additionalDirectories // []) | length' "$f")" != 0 ]; then
      log "$f da a las sesiones del run directorios adicionales (permissions.additionalDirectories): el run solo escribe en el repositorio y en el directorio temporal; quítalos antes de lanzar"; exit 1
    fi
  done
  case "$(uname -s)" in
    Darwin) command -v sandbox-exec >/dev/null 2>&1 \
      || { log "no encuentro sandbox-exec: las sesiones del run corren en el sandbox de Claude Code y no arrancarían"; exit 1; };;
    *) { command -v bwrap >/dev/null 2>&1 && command -v socat >/dev/null 2>&1; } \
      || { log "faltan bwrap o socat: las sesiones del run corren en el sandbox de Claude Code y no arrancarían"; exit 1; };;
  esac
  scripts/workflow/politica-paso.sh prueba >/dev/null \
    || { log "la política de las sesiones del run no deniega lo que debe (scripts/workflow/politica-paso.sh prueba)"; exit 1; }
}

# ---------------------------------------------------------------- entrada
case "${1:-}" in
  --clasificar)
    clasificar "${2:?run_id}"; exit 0;;
  --resume)
    run_id="${2:?run_id}"; shift 2
    args=()
    for kv in "$@"; do args+=(--input "$kv"); done
    comprobar_credenciales
    comprobar_entorno
    tomar_candado
    supervisar "$run_id" ${args[@]+"${args[@]}"}; exit $?;;
esac

hito="${1:?uso: scripts/hito.sh H<n> | --resume <run_id> [k=v …] | --clasificar <run_id>}"
if [ -n "${2:-}" ]; then
  echo "scripts/hito.sh ya no admite modo ('$2'): el workflow no tiene pausas humanas (ADR 0018)" >&2
  exit 2
fi

if [ -n "$(git status --porcelain)" ]; then
  echo "el árbol de trabajo tiene cambios sin commitear; el hito parte de main limpio" >&2
  exit 1
fi
if [ "$(git branch --show-current)" != "main" ]; then
  echo "sitúate en main antes de lanzar un hito (rama actual: $(git branch --show-current))" >&2
  exit 1
fi
comprobar_credenciales
comprobar_entorno
tomar_candado

# Modelo y esfuerzo por rol (<alias o nombre completo>[@esfuerzo]; roles en
# docs/WORKFLOW.md «Modelo por paso»). Ejemplo:
#   KITLEGAL_MODELO_IMPLEMENTACION=opus@max scripts/hito.sh H7
modelos=()
for rol in DECISION JUEZ REVISOR ADVERSARIO REDACCION IMPLEMENTACION ESCALADA ANALISIS; do
  var="KITLEGAL_MODELO_$rol"
  if [ -n "${!var:-}" ]; then
    modelos+=(--input "modelo_$(printf '%s' "$rol" | tr '[:upper:]' '[:lower:]')=${!var}")
  fi
done

supervisar "" --input "hito=$hito" ${modelos[@]+"${modelos[@]}"}
