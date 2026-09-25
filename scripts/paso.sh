#!/usr/bin/env bash
# Ejecuta a mano un paso `prompt` (juez, corrector, revisor) o `shell` (guardián,
# verificación, commit) del workflow hito con el mismo texto y el mismo modelo
# que usaría el workflow.
#
#   scripts/paso.sh juez_plan H2             # vuelve a juzgar el plan
#   scripts/paso.sh corrector_plan H2        # aplica los motivos del último veredicto
#   scripts/paso.sh revision_juez_a H2       # juez A de la revisión final (item "a" del fan-out)
#   scripts/paso.sh revision_juez_b H2 opus@xhigh   # juez B con otro modelo y esfuerzo
#   scripts/paso.sh verificar H2             # guardián + make ci de la tarea actual, log en gates/ci.log
#   scripts/paso.sh commit_tarea H2          # commit feat(H2): Tnnn de la tarea actual
#
# Desde el workflow 2.0.0 (ADR 0018) ningún gate para el run, así que esto ya no
# es parte de la operación normal: sirve para reproducir un paso concreto al
# investigar un run, o tras una causa mayor antes de reanudar con
# scripts/hito.sh --resume <run_id>. Los pasos shell llaman a scripts/workflow/*.sh,
# que también se pueden ejecutar directamente. No corre con un run vivo sobre el
# árbol (una sola sesión por run).
# Los pasos shell que interpolan salidas de otros pasos no se pueden lanzar sueltos.
#
# Los jueces de la revisión final viven en un fan-out (`revision_jueces`) con una
# única plantilla `revision_juez`; `revision_juez_<id>` selecciona el item con ese
# id ejecutando el paso shell que genera los items (`preparar_jueces`) con los
# inputs por defecto o los KITLEGAL_MODELO_* del entorno.
set -euo pipefail
cd "$(dirname "$0")/.."

# Una sola sesión por run (ADR 0018): con un run vivo sobre este árbol, un paso
# lanzado a mano escribiría a la vez que él.
scripts/workflow/sesion-unica.sh comprobar

paso="${1:?uso: scripts/paso.sh <id_paso> <hito> [modelo]}"
hito="${2:?uso: scripts/paso.sh <id_paso> <hito> [modelo]}"
modelo="${3:-}"

PY=python3
$PY -c 'import yaml' 2>/dev/null || PY="$(ls -d ~/.local/share/uv/tools/specify-cli/bin/python)"

eval "$($PY - "$paso" "$hito" "$modelo" <<'PYEOF'
import json, os, re, shlex, subprocess, sys
import yaml

paso, hito, modelo = sys.argv[1:4]
wf = yaml.safe_load(open('.specify/workflows/hito/workflow.yml'))
inputs = wf['inputs']


def salir(msg):
    print(f"echo {shlex.quote(msg)} >&2; exit 2"); sys.exit(0)


def valor_input(nombre):
    if nombre == 'hito':
        return hito
    return os.environ.get('KITLEGAL_' + nombre.upper()) or str(inputs.get(nombre, {}).get('default', ''))


def render(texto, item=None):
    def rep(m):
        expr = m.group(1).strip()
        if expr.startswith('inputs.'):
            return valor_input(expr[len('inputs.'):])
        if item is not None and expr.startswith('item.'):
            return str(item[expr[len('item.'):]])
        return m.group(0)  # se deja tal cual: lo detecta la comprobación de abajo
    return re.sub(r'\{\{(.+?)\}\}', rep, texto)


pasos = {}
fanouts = []
def walk(steps):
    for s in steps:
        pasos[s.get('id')] = s
        if s.get('type') == 'fan-out':
            fanouts.append(s)
            walk([s['step']])
        for k in ('then', 'else', 'steps'):
            if isinstance(s.get(k), list):
                walk(s[k])
walk(wf['steps'])

item = None
found = pasos.get(paso)
if found is None:
    # revision_juez_a → plantilla revision_juez del fan-out, item con id "a"
    m = re.fullmatch(r'(.+)_([a-z0-9]+)', paso)
    for fo in fanouts if m else []:
        plantilla = fo['step']
        if plantilla.get('id') != m.group(1):
            continue
        ref = re.fullmatch(r'\{\{\s*steps\.(\w+)\.output\.data\.(\w+)\s*\}\}', fo['items'].strip())
        if not ref:
            salir(f"el fan-out {fo['id']} no toma sus items de un paso shell con output json")
        generador = pasos.get(ref.group(1))
        if not generador or generador.get('type') != 'shell':
            salir(f"no encuentro el paso shell {ref.group(1)} que genera los items de {fo['id']}")
        r = subprocess.run(['bash', '-eu'], input=render(generador['run']), text=True, capture_output=True)
        if r.returncode:
            salir(f"{ref.group(1)} falló: {r.stderr.strip()}")
        items = json.loads(r.stdout)[ref.group(2)]
        item = next((i for i in items if str(i.get('id')) == m.group(2)), None)
        if item is None:
            salir(f"el fan-out {fo['id']} no tiene un item con id {m.group(2)!r}")
        found = plantilla
        break
if not found or found.get('type') not in ('prompt', 'shell'):
    salir(f'no existe un paso prompt ni shell con id {paso}')

tipo = found['type']
texto = render(found['prompt'] if tipo == 'prompt' else found['run'], item)
if '{{' in texto:
    salir(f'el paso {paso} interpola salidas de otros pasos; no se puede lanzar suelto')
if not modelo:
    modelo = render(str(found.get('model', '')), item) or 'opus'
print(f"TIPO={shlex.quote(tipo)}")
print(f"TEXTO={shlex.quote(texto)}")
print(f"MODELO={shlex.quote(modelo)}")
PYEOF
)"

if [ "$TIPO" = shell ]; then
  # Igual que el motor: subprocess.run(shell=True) → /bin/sh -c, en la raíz del repo.
  echo "→ paso $paso (shell) · hito $hito" >&2
  exec sh -c "$TEXTO"
fi
echo "→ paso $paso · hito $hito · modelo $MODELO" >&2
exec scripts/claude-modelo.sh -p "$TEXTO" --model "$MODELO" --permission-mode acceptEdits ${SPECKIT_INTEGRATION_CLAUDE_EXTRA_ARGS:-}
