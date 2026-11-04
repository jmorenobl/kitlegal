#!/usr/bin/env bash
# Sondeo de la persona antes de fusionar H40 (entrada de H40, «Aceptación»; FR-040): 200 sesiones con
# claude-sonnet-5-5, 40 por variante y modo, concurrencia 5 y Claude Code 2.1.317 (la del sondeo de H39, ADR 0050),
# sobre la cabeza de la #228 (be1dac4a) con el arnés que entrega el hito (make evals-laboratorio, familia normativa).
#   A. normativa-control en los dos modos: 2 filas, 80 sesiones.
#   B. normativa-remisiones-en-orden, normativa-cuerpo-corto y normativa-testigo-en-referencia, solo en modo
#      herramienta: 3 filas, 120 sesiones.
#   C. make laboratorio-tabla junta A y B en la tabla de 5 filas.
# Lee la credencial del llavero una vez, como scripts/evals-sondeo-llavero.sh, y lo bloquea antes de abrir ninguna
# sesión; con CLAUDE_CODE_OAUTH_TOKEN en el entorno, el guion del llavero no lo vuelve a pedir. No se lanza con un job
# de evals en marcha. Deja las salidas en salida-a.txt, salida-b.txt y tabla.md.
set -euo pipefail

S="$(cd "$(dirname "$0")" && pwd -P)"
llavero="${KITLEGAL_LLAVERO:-$HOME/Library/Keychains/kitlegal-sondeo.keychain-db}"

"$S/P/scripts/workflow/sesion-unica.sh" comprobar
[[ ! -e "$S/h40" ]] || { echo "sondeo: $S/h40 ya existe: el arnés no escribe sobre un sondeo anterior" >&2; exit 1; }

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

export PATH="$S/claude-2.1.317/node_modules/.bin:$PATH"
[[ "$(claude --version)" == 2.1.317* ]] || { echo "sondeo: el claude del PATH no es el 2.1.317" >&2; exit 1; }

parte() { # $1 nombre, $2 variantes, $3 modos
	echo "$(date '+%H:%M') parte $1: $2 en $3" >&2
	if make --no-print-directory -C "$S/P" evals-laboratorio VARIANTES="$2" MODOS="$3" SESIONES=40 MODELO=claude-sonnet-5-5 \
		CLAUDE_CODE=2.1.317 CONCURRENCIA=5 DIRECTORIO="$S/h40/$1" > "$S/salida-$1.txt" 2> "$S/registro-$1.err"; then
		echo "$(date '+%H:%M') parte $1 terminada: $S/salida-$1.txt" >&2
	else
		echo "$(date '+%H:%M') parte $1 falló: $S/registro-$1.err" >&2
		return 1
	fi
}

parte a normativa-control orden,herramienta
parte b normativa-remisiones-en-orden,normativa-cuerpo-corto,normativa-testigo-en-referencia herramienta
make --no-print-directory -C "$S/P" laboratorio-tabla DIRECTORIO="$S/h40" > "$S/tabla.md"
echo "$(date '+%H:%M') tabla: $S/tabla.md" >&2
