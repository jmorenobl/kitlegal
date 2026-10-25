#!/usr/bin/env bash
# Sondeo de la regresión de boe-legislacion desde H31: la eval 15 en los dos modos, 20 repeticiones por modo, con
# claude-sonnet-5-5 y el juez medido (Claude Code 2.1.289), en cuatro brazos que solo cambian la regla 5 de SKILL.md:
#   A la skill de main (control) · B la regla 5 de antes de H31 · C forma fija con ejemplos en el cuerpo
#   D la regla 5 corta y los ejemplos en references/responder.md, que el cuerpo manda leer antes de responder.
# Lee la credencial del llavero una vez, como scripts/evals-sondeo-llavero.sh, y lo vuelve a bloquear antes de abrir
# ninguna sesión. Cada brazo deja su salida en salida-<brazo>.txt y su temporal, con sesiones y votos, en la ruta que
# escribe registro-<brazo>.err.
set -euo pipefail

S="$(cd "$(dirname "$0")" && pwd -P)"
llavero="${KITLEGAL_LLAVERO:-$HOME/Library/Keychains/kitlegal-sondeo.keychain-db}"

"$S/A/scripts/workflow/sesion-unica.sh" comprobar

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

for brazo in A B C D; do
	echo "$(date '+%H:%M') brazo $brazo" >&2
	if (cd "$S/$brazo" && scripts/evals-sondeo.sh boe-legislacion 15 claude-sonnet-5-5 20 5) \
		> "$S/salida-$brazo.txt" 2> "$S/registro-$brazo.err"; then
		echo "$(date '+%H:%M') brazo $brazo terminado" >&2
	else
		echo "$(date '+%H:%M') brazo $brazo falló: $S/registro-$brazo.err" >&2
	fi
done
