#!/usr/bin/env bash
# Entrada de `make evals-sondeo`: toma la credencial de la suscripción de quien lo lanza y ejecuta
# scripts/evals-sondeo.sh con sus cinco argumentos (ADR 0032). El sondeo abre sesiones con modelo: lo lanza una
# persona, o una sesión interactiva a petición suya; nunca un paso del workflow `hito`.
#
#   make evals-sondeo SKILL=<skill> EVALS=<nn>[,<nn>…] MODELO=<id> REPETICIONES=<n> [CONCURRENCIA=<n>]
#
# La credencial no vive en ningún fichero que una sesión pueda leer. En H7.4 estaba en uno, y una sesión del run la
# leyó con `cat` dentro de un guion —las reglas `deny` de Claude Code solo cubren su herramienta Read— y abrió con
# ella 42 sesiones que nadie había pedido. En macOS vive en un llavero propio, que no es el de inicio de sesión:
#
#   ~/Library/Keychains/kitlegal-sondeo.keychain-db   (otra ruta: KITLEGAL_LLAVERO)
#   elemento genérico, servicio «kitlegal-claude-oauth-token», cuenta «kitlegal»
#
# El llavero está bloqueado: leerlo abre el diálogo del sistema, que pide su contraseña a la persona, y este guion lo
# vuelve a bloquear en cuanto tiene el valor, así que cada sondeo la pide. Sin persona delante no hay credencial. Con
# CLAUDE_CODE_OAUTH_TOKEN ya en el entorno (Linux, o quien prefiera ponerla a mano) no se toca el llavero. Sin una ni
# otro, scripts/evals-sondeo.sh da su error de uso. El valor nunca se imprime. Cómo se crea el llavero:
# CONTRIBUTING.md, «Sondeo local». No usa nada posterior a bash 3.2, el de macOS.
set -euo pipefail
cd "$(dirname "$0")/.."

# Un paso del workflow no abre sesiones con modelo, y con un run vivo el árbol de trabajo es suyo. Las dos variables
# las heredan las sesiones del run y todo lo que ejecutan (scripts/hito.sh, scripts/claude-modelo.sh).
if [[ -n "${KITLEGAL_PASO_DE_WORKFLOW:-}" || -n "${KITLEGAL_RUN_TESTIGO:-}" ]]; then
	echo "evals-sondeo: un paso del workflow no lanza el sondeo (ADR 0032): lo que no se pueda medir sin él lo mide el job de cierre" >&2
	exit 1
fi
scripts/workflow/sesion-unica.sh comprobar

llavero="${KITLEGAL_LLAVERO:-$HOME/Library/Keychains/kitlegal-sondeo.keychain-db}"

if [[ -z "${CLAUDE_CODE_OAUTH_TOKEN:-}" && -f "$llavero" ]] && command -v security > /dev/null 2>&1; then
	# El llavero vuelve a quedar bloqueado pase lo que pase, también si la persona cancela el diálogo.
	trap 'security lock-keychain "$llavero" > /dev/null 2>&1 || true' EXIT
	if ! credencial="$(security find-generic-password -a kitlegal -s kitlegal-claude-oauth-token -w "$llavero" 2> /dev/null)"; then
		echo "evals-sondeo: no se ha leído la credencial del llavero $llavero: el diálogo se canceló, la contraseña no es esa o falta el elemento «kitlegal-claude-oauth-token»" >&2
		exit 1
	fi
	if ! security lock-keychain "$llavero"; then
		echo "evals-sondeo: el llavero $llavero no se ha podido bloquear de nuevo; no se abre ninguna sesión" >&2
		exit 1
	fi
	trap - EXIT
	if [[ -z "$credencial" || "$credencial" == *[[:space:]]* ]]; then
		echo "evals-sondeo: el elemento «kitlegal-claude-oauth-token» del llavero $llavero está vacío o lleva espacios" >&2
		exit 1
	fi
	export CLAUDE_CODE_OAUTH_TOKEN="$credencial"
	unset credencial
fi

exec scripts/evals-sondeo.sh "$@"
