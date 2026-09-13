#!/usr/bin/env bash
# Graba las respuestas reales de la API de Legislación Consolidada del BOE que
# lista el manifiesto internal/source/boe/testdata/grabaciones.json, en
# internal/source/boe/testdata/boe.legislacion-consolidada/.
#
#   scripts/grabar-fixtures.sh
#
# Toca la red. Lo ejecuta una persona, a mano, en la pausa de la tarea [datos]
# del manifiesto y en el orden del procedimiento de esa pausa: después de revisar
# los términos de uso y el robots.txt del BOE, de dejar la fila de la fuente en
# docs/SOURCES.md con la fecha de la revisión y de alinear con ella las
# constantes de internal/source/boe/terminos.go, de modo que la grabación pide
# con el ritmo revisado (specs/005-h4-applet-boe-puerto/contracts/
# esquemas-fixtures-y-controles.md §3.2). Ningún objetivo de make, ningún
# gancho, ninguna tarea del workflow ni la integración continua lo ejecutan.
#
# Ejecuta solo TestGrabarFixtures, que únicamente compila con la etiqueta
# grabacion y que falla sin la variable de grabación antes de pedir nada.
set -euo pipefail
cd "$(dirname "$0")/.."

KITLEGAL_RECORD=1 go test -tags=grabacion -count=1 -run '^TestGrabarFixtures$' ./internal/source/boe/
