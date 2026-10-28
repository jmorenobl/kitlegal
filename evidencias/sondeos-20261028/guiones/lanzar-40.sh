#!/usr/bin/env bash
# Sondeo de la persona antes de fusionar H31.1 (2026-10-27): la eval 15 en los dos modos, 20 repeticiones por modo,
# concurrencia 5, con claude-sonnet-5-5 y el juez medido (Claude Code 2.1.289), sobre la cabeza de la #204 (1d7944f)
# con el arnés de evidencias/sondeos-20261025/parches/arnes-del-sondeo.diff, como evidencias/sondeos-20261026/.
# Criterio de la entrada: afirma_lo_no_leido como mucho 2 de 40 y cuenta_su_proceso como mucho 4 de 40.
# Lee la credencial del llavero una vez, como scripts/evals-sondeo-llavero.sh, y lo vuelve a bloquear antes de abrir
# ninguna sesión. Deja la salida en salida-P.txt y su temporal, con sesiones y votos, en la ruta que escribe
# registro-P.err.
set -euo pipefail

S="$(cd "$(dirname "$0")" && pwd -P)"
llavero="${KITLEGAL_LLAVERO:-$HOME/Library/Keychains/kitlegal-sondeo.keychain-db}"

"$S/P/scripts/workflow/sesion-unica.sh" comprobar

trap 'security lock-keychain "$llavero" > /dev/null 2>&1 || true' EXIT
credencial="$(security find-generic-password -a kitlegal -s kitlegal-claude-oauth-token -w "$llavero")"
security lock-keychain "$llavero"
trap - EXIT
if [[ -z "$credencial" || "$credencial" == *[[:space:]]* ]]; then
	echo "sondeo: la credencial del llavero está vacía o lleva espacios" >&2
	exit 1
fi
export CLAUDE_CODE_OAUTH_TOKEN="$credencial"
unset credencial

export PATH="$S/claude-2.1.289/node_modules/.bin:$PATH"
[[ "$(claude --version)" == 2.1.289* ]] || { echo "sondeo: el claude del PATH no es el 2.1.289" >&2; exit 1; }

echo "$(date '+%H:%M') brazo P (la #204)" >&2
if (cd "$S/P" && scripts/evals-sondeo.sh boe-legislacion 15 claude-sonnet-5-5 20 5) \
	> "$S/salida-P.txt" 2> "$S/registro-P.err"; then
	echo "$(date '+%H:%M') brazo P terminado: $S/salida-P.txt" >&2
else
	echo "$(date '+%H:%M') brazo P falló: $S/registro-P.err" >&2
	exit 1
fi
