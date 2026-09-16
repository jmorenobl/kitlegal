#!/usr/bin/env bash
# Graba las respuestas del BOE que necesitan las evals de las skills y la verificación de data/normas.yaml, según
# testdata/evals/grabaciones.json, en testdata/evals/boe.legislacion-consolidada/. Toca la red: lo ejecuta una persona
# en la pausa de la tarea [datos] del manifiesto (contracts/evals-y-grabaciones.md §3.3 de H5). Ningún objetivo de
# make, gancho, tarea ni flujo lo ejecuta.
set -euo pipefail
cd "$(dirname "$0")/.."

KITLEGAL_RECORD=1 go test -tags=grabacion -count=1 -v -run '^TestGrabarEvals$' ./internal/evals/
