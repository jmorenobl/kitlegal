#!/usr/bin/env bash
# Regenera la lista de hitos que admite el input hito de .specify/workflows/hito/workflow.yml desde docs/roadmap/
# (done/, in-progress/ y todo/), y comprueba después el roadmap entero (docs/roadmap/README.md).
set -euo pipefail
cd "$(dirname "$0")/.."

go test -count=1 -run '^TestRoadmap$' ./internal/workflow/ -args -regenerar-hitos
