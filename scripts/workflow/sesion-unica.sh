#!/usr/bin/env bash
# Una sola sesión por run (ADR 0018). En H6 una copia duplicada de la conversación
# estuvo trabajando en paralelo sobre el mismo árbol que el run.
#
#   scripts/workflow/sesion-unica.sh tomar        # hito.sh, al arrancar → imprime el testigo
#   scripts/workflow/sesion-unica.sh soltar <t>   # hito.sh, al salir
#   scripts/workflow/sesion-unica.sh comprobar    # paso.sh y cualquiera: sale con 1 si hay un run vivo
#   scripts/workflow/sesion-unica.sh gancho       # PreToolUse de Claude Code (.claude/settings.json)
#
# El candado es un directorio en el git-dir del árbol de trabajo (cada worktree
# tiene el suyo) con el PID de hito.sh y un testigo aleatorio. hito.sh exporta el
# testigo en KITLEGAL_RUN_TESTIGO y lo heredan specify, los pasos shell y las
# sesiones `claude -p` del run. Mientras el PID vive, el gancho rechaza en
# cualquier otra sesión de Claude Code las ediciones y las órdenes que cambian el
# árbol o el historial; leer sigue permitido. Un candado cuyo PID ya no existe
# (hito.sh murió sin soltarlo) se ignora y se recupera.
set -euo pipefail
cd "$(dirname "$0")/../.."

candado="$(git rev-parse --git-dir 2>/dev/null || echo .git)/kitlegal-run.lock"

vivo() { # ¿hay un run vivo con candado?
  [ -f "$candado/pid" ] && kill -0 "$(cat "$candado/pid")" 2>/dev/null
}

case "${1:?uso: sesion-unica.sh tomar | soltar <testigo> | comprobar | gancho}" in
  tomar)
    if vivo; then
      echo "ya hay un run del workflow sobre este árbol (PID $(cat "$candado/pid"), desde $(cat "$candado/desde" 2>/dev/null)); una sola sesión por run" >&2
      exit 1
    fi
    rm -rf "$candado"
    mkdir "$candado" 2>/dev/null || { echo "otro proceso acaba de tomar el candado $candado" >&2; exit 1; }
    testigo=$(od -An -N16 -tx1 /dev/urandom | tr -d ' \n')
    printf '%s\n' "${KITLEGAL_PID_RUN:-$PPID}" > "$candado/pid"
    printf '%s\n' "$testigo" > "$candado/testigo"
    date -u +%Y-%m-%dT%H:%M:%SZ > "$candado/desde"
    echo "$testigo";;

  soltar)
    [ -f "$candado/testigo" ] && [ "$(cat "$candado/testigo")" = "${2:-}" ] && rm -rf "$candado"
    exit 0;;

  comprobar)
    if vivo; then
      echo "hay un run del workflow vivo sobre este árbol (PID $(cat "$candado/pid")); espera a que termine o páralo antes" >&2
      exit 1
    fi;;

  gancho)
    vivo || exit 0
    [ "${KITLEGAL_RUN_TESTIGO:-}" = "$(cat "$candado/testigo" 2>/dev/null)" ] && exit 0
    entrada=$(cat)
    herramienta=$(jq -r '.tool_name // ""' <<<"$entrada")
    motivo="Hay un run del workflow hito vivo sobre este árbol (PID $(cat "$candado/pid"), desde $(cat "$candado/desde" 2>/dev/null)). Una sola sesión por run (ADR 0018): esta sesión puede leer, pero no editar ni cambiar el historial hasta que el run termine o se pare."
    case "$herramienta" in
      Edit|Write|NotebookEdit|MultiEdit) echo "$motivo" >&2; exit 2;;
      Bash)
        orden=$(jq -r '.tool_input.command // ""' <<<"$entrada")
        # Mirar cómo va el run es leer: scripts/hito.sh --clasificar no cambia nada.
        if printf '%s' "$orden" | grep -qE '^[[:space:]]*scripts/hito\.sh[[:space:]]+--clasificar[[:space:]]+[0-9a-f]+[[:space:]]*$'; then exit 0; fi
        if printf '%s' "$orden" | grep -qE '(^|[;&|[:space:]])(rtk[[:space:]]+)?git[[:space:]]+(add|commit|checkout|switch|stash|reset|restore|rebase|cherry-pick|merge|push|clean|rm|mv|apply|am|revert)([[:space:]]|$)|specify[[:space:]]+workflow[[:space:]]+(run|resume)|scripts/(hito|paso)\.sh|make[[:space:]]+(skills-sync|hooks|release)|(^|[[:space:]])(rm|mv|cp|sed[[:space:]]+-i|perl[[:space:]]+-[a-z]*i)[[:space:]]'; then
          echo "$motivo" >&2; exit 2
        fi;;
    esac
    exit 0;;

  *) echo "subcomando desconocido: $1" >&2; exit 2;;
esac
