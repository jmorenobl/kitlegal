#!/usr/bin/env bash
# Sondeo de la persona antes de fusionar H32 (entrada de H32, «Aceptación»; ADR 0047): 280 sesiones con
# claude-sonnet-5-5, concurrencia 5, y sesiones y juez con Claude Code 2.1.289, sobre la cabeza de la #214 con el
# arnés de evidencias/sondeos-20261025/parches/arnes-del-sondeo.diff, como evidencias/sondeos-20261028/.
#   1. normativa, evals 05 y 15, 40 por modo (160): eval 15, afirma_lo_no_leido ≤ 6/80 y cuenta_su_proceso ≤ 8/80
#      (Fisher unilateral p < 0,05 frente a evidencias/sondeos-20261025/, 16 y 17 de 80); eval 05,
#      cuenta_su_proceso ≤ 6/80.
#   2. plazos-y-territorio, evals 07, 08 y 12, 20 por modo (120): como mucho 1 de 120 sin la línea ⚠ NO CUBIERTO:
#      con la norma esperada y la cita del precepto.
# Lee la credencial del llavero una vez, como scripts/evals-sondeo-llavero.sh, y lo bloquea antes de abrir ninguna
# sesión. No se lanza con un job de evals en marcha. Deja las salidas en salida-N.txt y salida-T.txt.
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

parte() { # $1 nombre, $2 skill, $3 evals, $4 repeticiones por modo
	echo "$(date '+%H:%M') parte $1: $2, evals $3, $4 por modo" >&2
	if (cd "$S/P" && scripts/evals-sondeo.sh "$2" "$3" claude-sonnet-5-5 "$4" 5) \
		> "$S/salida-$1.txt" 2> "$S/registro-$1.err"; then
		echo "$(date '+%H:%M') parte $1 terminada: $S/salida-$1.txt" >&2
	else
		echo "$(date '+%H:%M') parte $1 falló: $S/registro-$1.err" >&2
		return 1
	fi
}

parte N normativa 05,15 40
parte T plazos-y-territorio 07,08,12 20
