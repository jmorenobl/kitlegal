#!/usr/bin/env bash
# Sondeo local de unas evals de una skill con Claude Code, sin strace ni veredicto (FR-060 a FR-068 de H7.3;
# contracts/sondeo.md §2 de H7.3; research.md D16 de H7.3). Crea su directorio temporal y lo borra al terminar, con el
# código que sea; ejecuta con una sola orden de Go TestSondeo, que comprueba los argumentos y la credencial, construye
# el binario y las skills del árbol de trabajo en el temporal, abre las sesiones de las evals pedidas y las juzga con el
# código del job, sin lo que el job lee de la traza, y deja su salida en salida.txt del temporal; y la imprime. Con un
# argumento que no vale o sin la credencial, TestSondeo deja su mensaje, un error por línea, en uso.txt del temporal, y
# el guion lo imprime solo en la salida de error, sin el registro de go test, y sale con 1 (contracts/sondeo.md §3 de
# H7.4; FR-080 de H7.4). No usa nada posterior a bash 3.2, el de macOS.
#
#   make evals-sondeo SKILL=<skill> EVALS=<nn>[,<nn>…] MODELO=<id> REPETICIONES=<n> [CONCURRENCIA=<n>]
#
# Los cinco argumentos son los de make, en ese orden y tal como los escribe quien lo lanza: CONCURRENCIA vacía es la del
# job para la skill, y los que no valen los nombra TestSondeo (FR-067). Abre sesiones con modelo y consume la
# suscripción de quien lo lanza, con CLAUDE_CODE_OAUTH_TOKEN: lo ejecuta una persona; ni make ci, ni los ganchos, ni el
# job, ni ninguna tarea del workflow (FR-068). No es un veredicto: el veredicto de una skill lo da el job de evals.
set -euo pipefail
cd "$(dirname "$0")/.."

if [[ $# -ne 5 ]]; then
	echo "evals-sondeo: uso: scripts/evals-sondeo.sh <skill> <evals> <modelo> <repeticiones> <concurrencia>" >&2
	exit 1
fi

# El temporal, en TMPDIR o, sin él, en /tmp, el directorio temporal por defecto de POSIX, se borra al salir, también
# cuando el guion termina con otro código (FR-064): todo lo que el sondeo escribe va dentro —el binario, el HOME de las
# sesiones con sus skills, las sesiones, el registro de go test, el error de uso y la salida—, y su tmp/ es el TMPDIR de
# go test, donde escriben el directorio de trabajo de go, el de go install y la preparación de cada sesión, que
# TestSondeo hace en proceso. Fuera quedan solo las cachés de Go. Su ruta es absoluta: TestSondeo la recibe por bandera
# y go test lo ejecuta en el directorio de su paquete.
temporal="$(mktemp -d "${TMPDIR:-/tmp}/kitlegal-sondeo.XXXXXX")"
trap 'rm -rf -- "$temporal"' EXIT
temporal="$(cd "$temporal" && pwd -P)"
mkdir "$temporal/tmp"

# Sin límite de tiempo de go test: el de cada sesión es su tope, y SIGINT y SIGTERM las cierran. Las dos salidas de go
# test van a su registro, que solo se imprime si el sondeo no ha podido abrir y juzgar sus sesiones por otra causa que
# un error de uso: el binario o las skills que no se pudieron construir o instalar, o el repartidor que devolvió un
# error (contracts/sondeo.md §6 de H7.3; FR-081 de H7.4).
codigo=0
TMPDIR="$temporal/tmp" go test -tags evals -count=1 -timeout 0 -run '^TestSondeo$' ./internal/evals/ -args \
	-skill "$1" -evals "$2" -modelo "$3" -repeticiones "$4" -concurrencia "$5" -temporal "$temporal" \
	> "$temporal/go-test.log" 2>&1 || codigo=$?

if [[ $codigo -ne 0 ]]; then
	cat "$temporal/go-test.log" >&2
	exit 1
fi

# Un argumento que no vale o la credencial que falta: TestSondeo ha terminado sin fallar y ha dejado el mensaje en
# uso.txt, que es todo lo que se imprime (contracts/sondeo.md §3 de H7.4; FR-080 de H7.4).
if [[ -s "$temporal/uso.txt" ]]; then
	cat "$temporal/uso.txt" >&2
	exit 1
fi

cat "$temporal/salida.txt"
