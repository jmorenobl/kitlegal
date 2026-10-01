#!/usr/bin/env bash
# Lo que una sesión con modelo de un paso del workflow no hace (ADR 0032). En H7.4
# `reparar_cierre` leyó el token de la suscripción de quien lanzó el run, escribió un
# guion fuera del repositorio y lo dejó en segundo plano abriendo sesiones con modelo
# (`rtk proxy sh <guion>` con `run_in_background`), y citó una verificación que no
# llegó a producirse.
#
#   scripts/workflow/politica-paso.sh gancho   # PreToolUse: lo llama sesion-unica.sh
#   scripts/workflow/politica-paso.sh prueba   # sus casos; lo ejecuta hito.sh al arrancar
#
# Solo se aplica a las sesiones que lanza scripts/claude-modelo.sh, que exporta
# KITLEGAL_PASO_DE_WORKFLOW: las de un run y las de scripts/paso.sh. Deniega, con el
# motivo en la salida de error y código 2:
#
#   · una orden en segundo plano: `run_in_background`, `nohup`, `disown`, `setsid`, `&`;
#   · abrir sesiones con modelo: `claude`, `make evals`, `make evals-sondeo`,
#     `scripts/evals*.sh` y `TestSondeo`;
#   · pedir una credencial: `security` sobre un llavero, o dar valor a
#     CLAUDE_CODE_OAUTH_TOKEN, ANTHROPIC_API_KEY o ANTHROPIC_AUTH_TOKEN;
#   · escribir, con Write, Edit, MultiEdit o NotebookEdit, fuera del repositorio, del
#     directorio temporal y de la memoria del proyecto.
#
# Mira el texto de la orden, así que no ve lo que hace un guion ni un programa que la
# sesión escriba y ejecute (en la reproducción del incidente, la sesión escribió el
# guion dentro del repositorio, lo copió fuera con `cp` y lo lanzó con «&» desde otro).
# Lo que hace cualquier orden de Bash, y todo lo que ejecuta, lo restringe el sandbox
# con el que scripts/claude-modelo.sh abre estas sesiones: la red, la escritura y la
# lectura de secretos. De este gancho dependen Write y Edit, que no pasan por el
# sandbox; en lo demás da el motivo antes y en la propia sesión. La credencial del
# sondeo no está en ningún fichero (scripts/evals-sondeo-llavero.sh). Compatible con
# bash 3.2 (macOS).
set -euo pipefail
cd "$(dirname "$0")/../.."

regla="Regla de las sesiones de un paso del workflow (constitución, «Reglas del modo desatendido»; ADR 0032)"

# Imprime el motivo y devuelve 1 si la orden de Bash no se admite. $1: la orden; $2: true si va en segundo plano.
orden_vedada() {
  local orden="$1" fondo="${2:-false}"
  if [ "$fondo" = true ]; then
    echo "$regla: nada en segundo plano. Ejecuta la orden en primer plano y espera a que termine; si no cabe en el plazo de la herramienta, pártela."
    return 1
  fi
  if printf '%s' "$orden" | grep -qE '(^|[^[:alnum:]_./-])(nohup|disown|setsid)([[:space:]]|$)|[[:space:])]&([[:space:];)]|$)'; then
    echo "$regla: nada en segundo plano (nohup, disown, setsid ni «&»). Ejecuta la orden en primer plano y espera a que termine."
    return 1
  fi
  # En posición de orden: al principio de una línea, tras un separador y un espacio (sin él es la barra de
  # una expresión regular: grep -E 'a|claude -p'), dentro de $( ) o de `sh -c '…'`, o tras rtk proxy, exec o
  # time; con las asignaciones de variables que lleve delante.
  local pos="(^|[;|&(][[:space:]]+|\\\$\\(|\`|-c[[:space:]]+['\"]|rtk[[:space:]]+proxy[[:space:]]|exec[[:space:]]|time[[:space:]])[[:space:]]*([A-Za-z_][A-Za-z0-9_]*=[^[:space:]]*[[:space:]]+)*"
  if printf '%s' "$orden" | grep -qE "$pos([^[:space:];|&()'\"]*/)?claude([[:space:]]|\$)|npx[^;|&]*claude-code|${pos}make[[:space:]]+([^;|&]*[[:space:]])?evals(-sondeo)?([[:space:]]|\$)|$pos((sh|bash|zsh)[[:space:]]+(-[eux]+[[:space:]]+)*)?(\\./)?scripts/evals[a-z-]*\\.sh|go[[:space:]].*test.*TestSondeo([^[:alnum:]_]|\$)"; then
    echo "$regla: un paso del workflow no abre sesiones con modelo (claude, make evals, make evals-sondeo, scripts/evals*.sh, TestSondeo). Lo que no puedas medir sin ellas, dilo: lo mide el job de cierre. Si solo buscabas ese texto, usa la herramienta Grep."
    return 1
  fi
  if printf '%s' "$orden" | grep -qE '(^|[^[:alnum:]_./-])security[[:space:]]+[a-z-]*(password|keychain|identity|certificate)|(CLAUDE_CODE_OAUTH_TOKEN|ANTHROPIC_API_KEY|ANTHROPIC_AUTH_TOKEN)=|setup-token'; then
    echo "$regla: un paso del workflow no pide ni usa credenciales que no le ha dado el workflow (llaveros, tokens de Claude)."
    return 1
  fi
  return 0
}

# Imprime el motivo y devuelve 1 si la ruta que se va a escribir cae fuera del repositorio,
# del directorio temporal y de la memoria del proyecto. $1: la ruta.
ruta_vedada() {
  local ruta="$1" raiz dir resto="" fisica permitida
  raiz=$(pwd -P)
  case "$ruta" in
    /*) ;;
    *) ruta="$raiz/$ruta";;
  esac
  case "$ruta/" in
    */../*|*/./*)
      echo "$regla: escribe la ruta sin «.» ni «..» ($1)."
      return 1;;
  esac
  # El directorio más profundo que ya existe, sin enlaces; lo que falta por crear va detrás.
  dir=$(dirname "$ruta")
  while [ ! -d "$dir" ]; do
    resto="/$(basename "$dir")$resto"
    dir=$(dirname "$dir")
  done
  fisica="$(cd "$dir" && pwd -P)$resto/"
  for permitida in "$raiz" /tmp /private/tmp /var/folders /private/var/folders "${TMPDIR:-/tmp}" \
    "$HOME/.claude/projects/$(printf '%s' "$raiz" | sed -E 's/[^A-Za-z0-9]/-/g')/memory"; do
    permitida="${permitida%/}"
    [ -d "$permitida" ] && permitida=$(cd "$permitida" && pwd -P)
    case "$fisica" in
      "$permitida"/*) return 0;;
    esac
  done
  echo "$regla: un paso del workflow solo escribe en el repositorio, en el directorio temporal y en la memoria del proyecto, no en $1."
  return 1
}

case "${1:?uso: politica-paso.sh gancho | prueba}" in
  gancho)
    entrada=$(cat)
    herramienta=$(jq -r '.tool_name // ""' <<<"$entrada")
    case "$herramienta" in
      Bash)
        orden=$(jq -r '.tool_input.command // ""' <<<"$entrada")
        fondo=$(jq -r '.tool_input.run_in_background // false' <<<"$entrada")
        motivo=$(orden_vedada "$orden" "$fondo") || { echo "$motivo" >&2; exit 2; };;
      Edit|Write|MultiEdit|NotebookEdit)
        ruta=$(jq -r '.tool_input.file_path // .tool_input.notebook_path // ""' <<<"$entrada")
        if [ -n "$ruta" ]; then
          motivo=$(ruta_vedada "$ruta") || { echo "$motivo" >&2; exit 2; }
        fi;;
    esac
    exit 0;;

  prueba)
    fallos=0
    # espera <admite|deniega> <orden|ruta> <valor> [fondo]
    espera() {
      local esperado="$1" clase="$2" valor="$3" fondo="${4:-false}" resultado=admite
      if [ "$clase" = orden ]; then
        orden_vedada "$valor" "$fondo" >/dev/null || resultado=deniega
      else
        ruta_vedada "$valor" >/dev/null || resultado=deniega
      fi
      if [ "$resultado" != "$esperado" ]; then
        echo "politica-paso: se esperaba «${esperado}» y da «${resultado}»: ${clase} ${valor}" >&2
        fallos=$((fallos + 1))
      fi
    }
    # La vía de H7.4, orden a orden (transcripts 82050c28 y 388e428c del run 0f3ef83d).
    espera deniega orden "rtk proxy sh /Users/x/Library/Logs/kitlegal/sondeo-c12.sh > /Users/x/Library/Logs/kitlegal/sondeo-c12.log 2>&1; echo terminado" true
    espera deniega orden "rtk proxy sh -c 'CLAUDE_CODE_OAUTH_TOKEN=\"\$(tr -d \"\\n\" < \"\$HOME/.config/kitlegal/claude-oauth-token\")\" make evals-sondeo SKILL=boe-legislacion EVALS=01 MODELO=claude-sonnet-5-5 REPETICIONES=3 > /tmp/s.txt 2>&1'"
    espera deniega orden "rtk proxy go test -tags evals -count=1 -timeout 0 -run '^TestSondeo\$' ./internal/evals/ -args -skill boe-legislacion"
    espera deniega orden "make evals-sondeo SKILL=boe-legislacion EVALS=18 MODELO=claude-sonnet-5-5 REPETICIONES=3"
    espera deniega orden "rtk proxy make evals SKILL=boe-legislacion"
    espera deniega orden "scripts/evals.sh --help"
    espera deniega orden "rtk proxy sh scripts/evals-sondeo.sh boe-legislacion 18 claude-sonnet-5-5 3 ''"
    espera deniega orden "claude -p --model sonnet --output-format json 'Responde solo: ok'"
    espera deniega orden "rtk proxy /Users/x/.local/bin/claude -p hola"
    espera deniega orden "npx -y @anthropic-ai/claude-code@2.1.284 -p hola"
    espera deniega orden "v=\$(claude --version); echo \$v"
    espera deniega orden "cd /tmp/x && CLAUDE_CONFIG_DIR=/tmp/c claude -p hola"
    espera deniega orden "security find-generic-password -s kitlegal-claude-oauth-token -w"
    espera deniega orden "rtk proxy security unlock-keychain /Users/x/Library/Keychains/kitlegal.keychain-db"
    espera deniega orden "nohup scripts/hito.sh H8"
    espera deniega orden "rtk proxy sh -c '(sleep 5; make ci) > /tmp/ci.log 2>&1 & echo lanzado'"
    espera deniega orden "make ci &"
    espera deniega orden "rtk proxy make ci 2>&1 | tail -40" true
    # Lo que un paso hace a diario.
    espera admite orden "rtk proxy make ci 2>&1 | tail -40"
    espera admite orden "rtk proxy sh -c 'make ci > specs/014/gates/ci.log 2>&1; echo \"código \$?\"'"
    espera admite orden "go build ./... && go vet ./... && go test -race ./..."
    espera admite orden "rtk proxy go test -count=1 -run 'TestEvalsDelRepositorio' ./internal/evals/ 2>&1 | tail -15"
    espera admite orden "rtk proxy go vet -tags evals ./internal/evals/"
    espera admite orden "cat scripts/evals-sondeo.sh; ls .claude/ scripts/claude-modelo.sh"
    espera admite orden "grep -n 'evals-sondeo' -A 6 Makefile | head -30"
    espera admite orden "jq '.umbrales' specs/014/gates/evals/boe-legislacion.json >&2"
    espera admite orden "make skills-check && make evals-check 2>&1"
    espera admite orden "rtk proxy git log --oneline -5 -- .claude/settings.json"
    espera admite orden "rtk proxy bash -n scripts/evals-sondeo.sh && echo ok"
    espera admite orden "ps aux | grep -E 'evals-sondeo|sondeo-c12|claude -p' | grep -v grep"
    espera admite orden "ls -la \"\$(command -v claude)\"; go test -count=1 -run 'TestGuionDelSondeo|TestSondear' ./internal/evals/"
    espera admite ruta "specs/014-h7-4/research.md"
    espera admite ruta "$(pwd -P)/internal/nuevo/paquete/fichero.go"
    espera admite ruta "/tmp/kitlegal-prototipo/main.go"
    espera admite ruta "${TMPDIR:-/tmp}/copia/x.go"
    espera admite ruta "$HOME/.claude/projects/$(pwd -P | sed -E 's/[^A-Za-z0-9]/-/g')/memory/nota.md"
    espera deniega ruta "/Users/x/Library/Logs/kitlegal/sondeo-c12.sh"
    espera deniega ruta "/Users/x/.config/kitlegal/claude-oauth-token"
    espera deniega ruta "/Users/x/.claude/settings.json"
    espera deniega ruta "/tmp/../Users/x/x.sh"
    if [ "$fallos" -ne 0 ]; then
      echo "politica-paso: $fallos casos no dan lo esperado" >&2
      exit 1
    fi
    echo "politica-paso: todos los casos dan lo esperado";;

  *) echo "subcomando desconocido: $1" >&2; exit 2;;
esac
