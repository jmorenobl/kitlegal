#!/usr/bin/env bash
# Coste y consumo por paso y por rol de un run del workflow `hito`.
#
#   scripts/coste-run.sh              # el run más reciente
#   scripts/coste-run.sh 5f3e014a
#
# spec-kit no conserva la salida de `claude -p`, así que el consumo se
# reconstruye desde los transcripts de Claude Code (~/.claude/projects/<ruta>):
# cada sesión headless (entrypoint sdk-cli) se asigna al paso command/prompt
# del log del run en cuyo intervalo empezó, y el rol sale del `model:` de ese
# paso en la copia congelada del workflow del run.
#
# El coste es una estimación a precio de lista de la API (tabla PRECIOS); con
# suscripción, léase como proporción entre pasos. Transcripts en otra ruta:
# KITLEGAL_TRANSCRIPTS=<dir>.
set -euo pipefail
cd "$(dirname "$0")/.."

runs=.specify/workflows/runs
run_id="${1:-$(ls -t "$runs" 2>/dev/null | head -1)}"
[ -n "$run_id" ] && [ -f "$runs/$run_id/log.jsonl" ] || { echo "no existe el run '$run_id' en $runs" >&2; exit 2; }

PY=python3
$PY -c 'import yaml' 2>/dev/null || PY="$(ls -d ~/.local/share/uv/tools/specify-cli/bin/python)"

exec "$PY" - "$runs/$run_id" "$run_id" <<'PYEOF'
import collections, datetime, glob, json, os, re, sys
import yaml

run_dir, run_id = sys.argv[1:3]

# USD por millón de tokens: entrada, salida, lectura de caché. La escritura de
# caché cuesta 1,25× la entrada (TTL 5 min) o 2× (TTL 1 h, el de Claude Code).
PRECIOS = {
    "claude-fable-5-1": (10.0, 50.0, 0.25),
    "claude-opus-5": (5.0, 25.0, 0.50),
    "claude-sonnet-5": (2.0, 10.0, 0.20),
    "claude-haiku-4-5": (1.0, 5.0, 0.10),
}

def precio(modelo):
    for k, p in PRECIOS.items():
        if modelo.startswith(k):
            return p
    return None

def ts(s):
    return datetime.datetime.fromisoformat(s.replace("Z", "+00:00"))

# Plantillas de fan-out: el modelo es "{{ item.modelo }}" y el rol depende del item.
# Orden de los items tal y como los genera preparar_jueces en el workflow.
FAN_OUT_ROLES = {"revision_juez": ["modelo_revisor", "modelo_juez"]}


def nombre(step_id):
    # "ronda_plan:juez_plan:1" → "juez_plan"; "ronda_revision:revision_jueces:revision_juez:0" → "revision_juez:0"
    partes = step_id.split(":")
    if len(partes) >= 2 and partes[-1].isdigit() and partes[-2] in FAN_OUT_ROLES:
        return f"{partes[-2]}:{partes[-1]}"
    return [p for p in partes if not p.isdigit()][-1]

# ---- rol de cada paso según la copia congelada del workflow del run
rol_de = {}
def walk(steps):
    for s in steps:
        m = re.search(r"inputs\.(modelo_\w+)", str(s.get("model", "")))
        if m:
            rol_de[s["id"]] = m.group(1)
        for k in ("then", "else", "steps"):
            if isinstance(s.get(k), list):
                walk(s[k])
        if isinstance(s.get("step"), dict):
            walk([s["step"]])
walk(yaml.safe_load(open(os.path.join(run_dir, "workflow.yml")))["steps"])
for base, roles in FAN_OUT_ROLES.items():
    for i, rol in enumerate(roles):
        rol_de[f"{base}:{i}"] = rol
inputs = json.load(open(os.path.join(run_dir, "inputs.json"))).get("inputs", {})

# ---- intervalos de los pasos que invocan a Claude
intervalos, abiertos = [], {}
eventos = [json.loads(l) for l in open(os.path.join(run_dir, "log.jsonl")) if l.strip()]
for e in eventos:
    sid = e.get("step_id")
    if not sid:
        continue
    if e["event"] == "step_started" and e.get("type") in ("command", "prompt"):
        abiertos[sid] = ts(e["timestamp"])
    elif sid in abiertos and e["event"] != "step_started":
        intervalos.append((abiertos.pop(sid), ts(e["timestamp"]), nombre(sid)))
fin_abierto = datetime.datetime.max.replace(tzinfo=datetime.timezone.utc)
intervalos += [(t0, fin_abierto, nombre(sid)) for sid, t0 in abiertos.items()]
if not intervalos:
    sys.exit(f"el run {run_id} aún no ha ejecutado ningún paso con modelo")
ini_run = min(i[0] for i in intervalos)

# ---- sesiones headless y su consumo
proyecto = os.path.abspath(".")
tdir = os.environ.get("KITLEGAL_TRANSCRIPTS") or os.path.expanduser(
    "~/.claude/projects/" + re.sub(r"[^A-Za-z0-9]", "-", proyecto))

def consumo(ficheros):
    ultimo = {}
    for f in ficheros:
        for l in open(f):
            try:
                e = json.loads(l)
            except ValueError:
                continue
            if e.get("type") == "assistant" and e.get("message", {}).get("usage"):
                m = e["message"]
                ultimo[m.get("id") or id(m)] = m
    return list(ultimo.values())

filas = collections.defaultdict(lambda: {"ses": 0, "turnos": 0, "out": 0, "coste": 0.0, "modelos": set(), "sin_precio": False})
for f in glob.glob(os.path.join(tdir, "*.jsonl")):
    inicio = headless = None
    with open(f) as fh:
        for l in fh:
            try:
                e = json.loads(l)
            except ValueError:
                continue
            if e.get("type") == "user" and e.get("timestamp"):
                inicio, headless = ts(e["timestamp"]), e.get("entrypoint") == "sdk-cli"
                break
    if not inicio or not headless or inicio < ini_run:
        continue
    candidatos = [n for a, b, n in intervalos if a <= inicio <= b]
    if not candidatos:
        continue
    sesion = os.path.splitext(os.path.basename(f))[0]
    mensajes = consumo([f] + glob.glob(os.path.join(tdir, sesion, "**", "*.jsonl"), recursive=True))
    paso = candidatos[0]
    if len(candidatos) > 1 and mensajes:
        # dos pasos a la vez (jueces en paralelo): el que declare la misma familia de modelo que la sesión
        familia = re.sub(r"^claude-|-[0-9].*$", "", mensajes[-1].get("model", ""))
        for n in candidatos:
            valor = str(inputs.get(rol_de.get(n, ""), ""))
            if re.sub(r"^claude-|@.*$|-[0-9].*$", "", valor) == familia:
                paso = n
                break
    fila = filas[paso]
    fila["ses"] += 1
    for m in mensajes:
        u, modelo = m["usage"], m.get("model", "")
        fila["turnos"] += 1
        fila["out"] += u.get("output_tokens", 0)
        fila["modelos"].add(modelo)
        p = precio(modelo)
        if not p:
            fila["sin_precio"] = True
            continue
        cc = u.get("cache_creation") or {}
        w1h = cc.get("ephemeral_1h_input_tokens", 0)
        w5m = cc.get("ephemeral_5m_input_tokens", u.get("cache_creation_input_tokens", 0) - w1h)
        fila["coste"] += (u.get("input_tokens", 0) * p[0] + u.get("output_tokens", 0) * p[1]
                          + u.get("cache_read_input_tokens", 0) * p[2]
                          + w5m * p[0] * 1.25 + w1h * p[0] * 2) / 1e6

if not filas:
    sys.exit(f"no hay sesiones headless del run {run_id} en {tdir}")

total = sum(f["coste"] for f in filas.values())
print(f"run {run_id} · hito {inputs.get('hito', '?')} · {sum(f['ses'] for f in filas.values())} sesiones · "
      f"${total:.2f} estimado a precio de lista\n")
print(f"{'paso':28} {'rol':22} {'modelo':18} {'ses':>3} {'turnos':>6} {'salida':>7} {'coste':>8} {'%':>5}")
por_rol = collections.Counter()
for paso, f in sorted(filas.items(), key=lambda kv: -kv[1]["coste"]):
    rol = rol_de.get(paso, "?")
    por_rol[rol] += f["coste"]
    modelo = ",".join(sorted(m.replace("claude-", "") for m in f["modelos"]))
    aviso = " (sin precio)" if f["sin_precio"] else ""
    print(f"{paso:28} {rol:22} {modelo:18} {f['ses']:3d} {f['turnos']:6d} {f['out'] // 1000:6d}k "
          f"{f['coste']:8.2f} {100 * f['coste'] / total:5.1f}{aviso}")
print(f"\n{'rol':22} {'valor en el run':18} {'coste':>8} {'%':>5}")
for rol, c in por_rol.most_common():
    print(f"{rol:22} {str(inputs.get(rol, '?')):18} {c:8.2f} {100 * c / total:5.1f}")
PYEOF
