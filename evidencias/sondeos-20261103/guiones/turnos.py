"""Cuenta, por variante y modo, las sesiones del sondeo de H39 en las que ToolSearch va en el mismo turno que Skill.

Uso: python3 -I turnos.py <directorio del sondeo>, el que recibió make laboratorio-tabla (con a/ y b/ dentro).
Un turno es un mensaje del agente (su message.id en sesion.jsonl); la cifra se compara con la de normativa en los
sondeos de la #214, 240 de 240 en modo herramienta (specs/030-h32-arquitectura-de-skills/research.md D26 a D30).
"""
import collections
import glob
import json
import sys

cuentas = collections.defaultdict(lambda: [0, 0, 0])
for fichero in glob.glob(sys.argv[1] + "/*/*/*/*/sesion.jsonl"):
    variante, modo = fichero.split("/")[-4:-2]
    turnos = collections.OrderedDict()
    with open(fichero, encoding="utf-8") as f:
        for linea in f:
            d = json.loads(linea)
            if d.get("type") != "assistant":
                continue
            for c in d["message"]["content"]:
                if c.get("type") == "tool_use":
                    turnos.setdefault(d["message"]["id"], []).append(c["name"])
    turnos = list(turnos.values())
    busca = [i for i, t in enumerate(turnos) if "ToolSearch" in t]
    activa = [i for i, t in enumerate(turnos) if "Skill" in t]
    c = cuentas[(variante, modo)]
    c[0] += 1
    c[1] += bool(busca and activa and busca[0] == activa[0])
    c[2] += bool(busca)

print("| Variante | Modo | ToolSearch en el turno de Skill | Con ToolSearch | n |")
print("|---|---|---|---|---|")
for (variante, modo), (n, mismo, con) in sorted(cuentas.items()):
    print(f"| {variante} | {modo} | {mismo} | {con} | {n} |")
