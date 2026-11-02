#!/usr/bin/env bash
# Sondeo de la persona antes de fusionar H32 (entrada de H32, «Aceptación»; ADR 0047), repetición de normativa sobre
# v0.4.4: 160 sesiones con claude-sonnet-5-5, concurrencia 5, y sesiones y juez con Claude Code 2.1.289, sobre la cabeza
# de la #214 (854b60a) con el arnés de evidencias/sondeos-20261025/parches/arnes-del-sondeo.diff.
#   normativa, evals 05 y 15, 40 por modo (160): eval 15, afirma_lo_no_leido ≤ 6/80 y cuenta_su_proceso ≤ 8/80
#   (Fisher unilateral p < 0,05 frente a evidencias/sondeos-20261025/, 16 y 17 de 80); eval 05,
#   cuenta_su_proceso ≤ 6/80. El sondeo de plazos-y-territorio sobre ddfb799 sigue valiendo: esa skill no ha cambiado.
# Lee la credencial del llavero una vez, como scripts/evals-sondeo-llavero.sh, y lo bloquea antes de abrir ninguna
# sesión. No se lanza con un job de evals en marcha. Deja la salida en salida-N.txt.
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

echo "$(date '+%H:%M') normativa, evals 05,15, 40 por modo" >&2
if (cd "$S/P" && scripts/evals-sondeo.sh normativa 05,15 claude-sonnet-5-5 40 5) \
	> "$S/salida-N.txt" 2> "$S/registro-N.err"; then
	echo "$(date '+%H:%M') terminada: $S/salida-N.txt" >&2
else
	echo "$(date '+%H:%M') falló: $S/registro-N.err" >&2
	exit 1
fi
