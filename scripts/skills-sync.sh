#!/usr/bin/env bash
# Regenera references/ y la tabla de comandos de SKILL.md de cada skill de skills/, desde data/*.yaml y desde
# --describe del binario (contracts/sincronizacion-y-comprobacion.md de H5; contracts/skills-e-invocacion.md §3 de H19).
set -euo pipefail
cd "$(dirname "$0")/.."

go test -count=1 -run '^TestSkillsDelRepositorio$' ./internal/app/ -args -regenerar-skills
