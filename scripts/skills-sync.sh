#!/usr/bin/env bash
# Regenera references/, la tabla de comandos de SKILL.md y los enlaces de scripts/ de cada skill de skills/,
# desde data/*.yaml y desde --describe del binario (contracts/sincronizacion-y-comprobacion.md de H5).
set -euo pipefail
cd "$(dirname "$0")/.."

go test -count=1 -run '^TestSkillsDelRepositorio$' ./internal/app/ -args -regenerar-skills
