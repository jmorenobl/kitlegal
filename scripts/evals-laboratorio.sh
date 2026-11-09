#!/usr/bin/env bash
# Sondeo del laboratorio de H39, o de las variantes de una de las tres familias de las skills del producto, normativa
# (H40 y H41), plazos-y-territorio o jurisprudencia (H41), una familia por invocación, con Claude Code: abre las
# sesiones de las cuatro evals de la familia con cada variante y modo pedidos y las deja en DIRECTORIO, sin juzgarlas,
# para la tabla de make laboratorio-tabla (FR-030 a FR-033 de H39; FR-010 de H40; FR-030 a FR-033 de H41;
# specs/031-h39-c-omo-se/contracts/arnes.md §1 y §2; specs/032-h40-por-qu-e/contracts/arnes.md §1 a §3;
# specs/034-h41-las-skills-del/contracts/arnes.md §1 a §3; research.md D8 de H39). Como scripts/evals-sondeo.sh: crea su
# directorio temporal y lo borra al terminar, con el código que sea; ejecuta con una sola orden de Go
# TestSondeoDelLaboratorio, que comprueba los argumentos, la credencial y la versión de Claude Code antes de construir
# nada, construye kitlegal y laboratorio del árbol de trabajo en el temporal, abre las sesiones y deja su salida en
# salida.txt del temporal; y la imprime. Con un argumento que no vale, sin la credencial o con otra versión de Claude
# Code, TestSondeoDelLaboratorio deja su mensaje, un error por línea, en uso.txt del temporal, y el guion lo imprime
# solo en la salida de error, sin el registro de go test, y sale con 1. No usa nada posterior a bash 3.2, el de macOS.
#
#   make evals-laboratorio VARIANTES=<v>[,<v>…] MODOS=<m>[,<m>] SESIONES=<n> MODELO=<id> CLAUDE_CODE=<versión> \
#     CONCURRENCIA=<c> DIRECTORIO=<ruta>
#
# Los siete argumentos son los de make, en ese orden y tal como los escribe quien lo lanza: ninguno tiene valor por
# defecto, y los que no valen los nombra TestSondeoDelLaboratorio. Abre sesiones con modelo y consume la suscripción de
# quien lo lanza, con CLAUDE_CODE_OAUTH_TOKEN: lo ejecuta una persona; ni make ci, ni los ganchos, ni el job de evals,
# ni ningún paso del workflow (FR-032 de H39). La credencial se la pone en el entorno scripts/evals-sondeo-llavero.sh,
# que es lo que ejecuta make (ADR 0032).
set -euo pipefail
cd "$(dirname "$0")/.."

if [[ $# -ne 7 ]]; then
	echo "evals-laboratorio: uso: scripts/evals-laboratorio.sh <variantes> <modos> <sesiones> <modelo> <claude-code> <concurrencia> <directorio>" >&2
	exit 1
fi

# DIRECTORIO relativo es desde la raíz del repositorio, donde lo escribe quien lanza make; go test ejecuta
# TestSondeoDelLaboratorio en el directorio de su paquete, así que lo recibe con la raíz delante. Vacío o absoluto pasa
# tal cual: el error de uno vacío o que ya existe lo da TestSondeoDelLaboratorio (contracts/arnes.md §2 de H39).
directorio="$7"
if [[ -n "$directorio" && "$directorio" != /* ]]; then
	directorio="$(pwd -P)/$directorio"
fi

# El temporal, en TMPDIR o, sin él, en /tmp, el directorio temporal por defecto de POSIX, se borra al salir, también
# cuando el guion termina con otro código: todo lo que el arnés prepara va dentro —kitlegal, laboratorio, el HOME de
# las sesiones, la copia de las skills de cada variante, el registro de go test, el error de uso y la salida—, y su
# tmp/ es el TMPDIR de go test, donde escriben el directorio de trabajo de go, el de go install y la preparación de cada
# sesión. Fuera quedan las cachés de Go y las sesiones, en DIRECTORIO. Su ruta es absoluta: TestSondeoDelLaboratorio la
# recibe por bandera y go test lo ejecuta en el directorio de su paquete.
temporal="$(mktemp -d "${TMPDIR:-/tmp}/kitlegal-laboratorio.XXXXXX")"
trap 'rm -rf -- "$temporal"' EXIT
temporal="$(cd "$temporal" && pwd -P)"
mkdir "$temporal/tmp"

# Sin límite de tiempo de go test: el de cada sesión es su tope, y SIGINT y SIGTERM las cierran. Las dos salidas de go
# test van a su registro, que solo se imprime si el arnés no ha podido abrir sus sesiones por otra causa que un error de
# uso: kitlegal o laboratorio que no se pudieron construir, las skills de una variante que no se pudieron copiar,
# DIRECTORIO que no se pudo crear o el repartidor que devolvió un error (contracts/arnes.md §1 de H39).
codigo=0
TMPDIR="$temporal/tmp" go test -tags evals -count=1 -timeout 0 -run '^TestSondeoDelLaboratorio$' ./internal/evals/ -args \
	-variantes "$1" -modos "$2" -sesiones "$3" -modelo "$4" -claude-code "$5" -concurrencia "$6" -directorio "$directorio" \
	-temporal "$temporal" > "$temporal/go-test.log" 2>&1 || codigo=$?

if [[ $codigo -ne 0 ]]; then
	cat "$temporal/go-test.log" >&2
	exit 1
fi

# Un argumento que no vale, la credencial que falta u otra versión de Claude Code: TestSondeoDelLaboratorio ha terminado
# sin fallar y ha dejado el mensaje en uso.txt, que es todo lo que se imprime (contracts/arnes.md §1 y §2 de H39).
if [[ -s "$temporal/uso.txt" ]]; then
	cat "$temporal/uso.txt" >&2
	exit 1
fi

cat "$temporal/salida.txt"
