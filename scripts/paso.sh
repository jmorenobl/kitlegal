#!/usr/bin/env bash
# Ejecuta a mano un paso `prompt` del workflow hito (juez, corrector, revisor)
# con el mismo texto y el mismo modelo que usaría el workflow.
#
#   scripts/paso.sh juez_plan H0            # vuelve a juzgar el plan
#   scripts/paso.sh corrector_plan H0       # aplica los motivos del último veredicto
#   scripts/paso.sh revision_juez_b H0 opus # con otro modelo
#
# Útil cuando un run se ha parado en check_gate_* y quieres una ronda más
# antes de reanudar con scripts/hito.sh --resume <run_id>.
set -euo pipefail
cd "$(dirname "$0")/.."

paso="${1:?uso: scripts/paso.sh <id_paso> <hito> [modelo]}"
hito="${2:?uso: scripts/paso.sh <id_paso> <hito> [modelo]}"
modelo="${3:-}"

PY=python3
$PY -c 'import yaml' 2>/dev/null || PY="$(ls -d ~/.local/share/uv/tools/specify-cli/bin/python)"

eval "$($PY - "$paso" "$hito" "$modelo" <<'PYEOF'
import sys, yaml, shlex
paso, hito, modelo = sys.argv[1:4]
wf = yaml.safe_load(open('.specify/workflows/hito/workflow.yml'))
found = None
def walk(steps):
    global found
    for s in steps:
        if s.get('id') == paso: found = s
        for k in ('then', 'else', 'steps'):
            if isinstance(s.get(k), list): walk(s[k])
walk(wf['steps'])
if not found or found.get('type') != 'prompt':
    print(f"echo 'no existe un paso prompt con id {paso}' >&2; exit 2"); sys.exit(0)
prompt = found['prompt'].replace('{{ inputs.hito }}', hito)
if '{{' in prompt:
    print(f"echo 'el paso {paso} interpola salidas de otros pasos; no se puede lanzar suelto' >&2; exit 2"); sys.exit(0)
m = found.get('model', '')
if not modelo:
    key = m.replace('{{', '').replace('}}', '').replace('inputs.', '').strip()
    modelo = wf['inputs'].get(key, {}).get('default', 'opus')
print(f"PROMPT={shlex.quote(prompt)}")
print(f"MODELO={shlex.quote(modelo)}")
PYEOF
)"

echo "→ paso $paso · hito $hito · modelo $MODELO" >&2
exec claude -p "$PROMPT" --model "$MODELO" --permission-mode acceptEdits ${SPECKIT_INTEGRATION_CLAUDE_EXTRA_ARGS:-}
