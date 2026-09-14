#!/usr/bin/env bash
# Verifica contra la fuente real que sus respuestas se siguen interpretando. El
# único caso es «boe articulo»: pide a la API de Legislación Consolidada del BOE
# el artículo 21 de la Ley 39/2015 y exige código 0, un sobre válido contra la
# parte articulo de schemas/bloque.json y texto no vacío (FR-115; contrato
# specs/005-h4-applet-boe-puerto/contracts/esquemas-fixtures-y-controles.md §7).
#
#   make verify-sources
#
# Toca la red: es el único control que pide algo a una fuente real. Lo ejecutan
# make verify-sources y el trabajo fuentes del flujo nocturno, que abre o
# actualiza una incidencia si falla; ni make ci, ni los ganchos, ni ninguna tarea
# del workflow lo ejecutan. No graba nada ni escribe en la caché de la cuenta: la
# de la verificación vive en una carpeta temporal del test.
#
# Ejecuta solo TestVerificarFuentes, que únicamente compila con la etiqueta
# fuentes. La misma comprobación, sin red, la ejerce make test sobre las
# grabaciones y sobre una respuesta que ya no se interpreta.
set -euo pipefail
cd "$(dirname "$0")/.."

go test -tags=fuentes -count=1 -run '^TestVerificarFuentes$' ./internal/app/
