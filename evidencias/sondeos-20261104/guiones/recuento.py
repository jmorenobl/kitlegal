"""Cuenta, por variante, modo y eval, las sesiones del sondeo de H40 que abren con_herramientas.md, normas.md,
redaccion_en_una_fecha.md y como_se_cita.md, y las que cumplen la regla testigo.

Uso: python3 -I recuento.py <directorio del sondeo> <tabla.md>, el directorio que recibió make laboratorio-tabla (con
a/ y b/ dentro) y la tabla que imprimió.

Dos cuentas por referencia:
- «tabla»: la de make laboratorio-tabla, leída del detalle de tabla.md (specs/032-h40-por-qu-e/contracts/tabla.md §2:
  Read de la referencia, o Bash cuya orden contiene references/<f>);
- «ampliada»: además, Bash cuya orden nombra <f> sin la ruta, tras un cd al directorio references/ («cd …/references
  && cat con_herramientas.md»), que el contrato no cuenta.
«Cumplen» sale del detalle de tabla.md (reglaTestigo). Los intervalos son de Wilson al 95 %.
"""
import collections
import glob
import json
import math
import re
import sys

REFERENCIAS = ["con_herramientas.md", "normas.md", "redaccion_en_una_fecha.md", "como_se_cita.md"]


def wilson(k, n, z=1.959963984540054):
    if n == 0:
        return "—"
    p = k / n
    d = 1 + z * z / n
    c = (p + z * z / (2 * n)) / d
    h = z * math.sqrt(p * (1 - p) / n + z * z / (4 * n * n)) / d
    return f"[{max(0, c - h) * 100:.1f}–{min(1, c + h) * 100:.1f} %]".replace(".", ",")


def detalle(tabla):
    """Sesión → {columna: valor} del detalle de tabla.md."""
    texto = open(tabla, encoding="utf-8").read().split("## Detalle", 1)[1]
    filas = [l for l in texto.splitlines() if l.startswith("| ")]
    cabecera = [c.strip() for c in filas[0].strip("|").split("|")]
    out = {}
    for l in filas[1:]:
        celdas = [c.strip() for c in l.strip("|").split("|")]
        out[celdas[0]] = dict(zip(cabecera, celdas))
    return out


def ampliada(fichero):
    """Referencias que la sesión lee con Bash tras un cd a references/, sin la ruta en la orden."""
    leidas = set()
    with open(fichero, encoding="utf-8") as f:
        for linea in f:
            d = json.loads(linea)
            if d.get("type") != "assistant":
                continue
            for c in d["message"]["content"]:
                if c.get("type") != "tool_use" or c["name"] != "Bash":
                    continue
                orden = c["input"].get("command", "")
                if not re.search(r"\bcd\s+\S*references/?(\s|&|;|$)", orden):
                    continue
                for r in REFERENCIAS:
                    if re.search(r"(^|[\s/])" + re.escape(r) + r"\b", orden):
                        leidas.add(r)
    return leidas


raiz, tabla = sys.argv[1], sys.argv[2]
filas = detalle(tabla)
cuentas = collections.defaultdict(collections.Counter)
for fichero in sorted(glob.glob(raiz + "/*/*/*/*/sesion.jsonl")):
    parte, variante, modo, sesion = fichero.split("/")[-5:-1]
    fila = filas[f"{parte}/{variante}/{modo}/{sesion}"]
    extra = ampliada(fichero)
    for clave in [(variante, modo, sesion[:2]), (variante, modo, "todas")]:
        c = cuentas[clave]
        c["n"] += 1
        for r in REFERENCIAS:
            abierta = fila[r] != "no"
            c[r + " tabla"] += abierta
            c[r + " ampliada"] += abierta or r in extra
        c["cumplen"] += fila["Cumple"] == "sí"

print("| Variante | Modo | Eval | n | " + " | ".join(r + " (tabla · ampliada)" for r in REFERENCIAS) + " | Cumplen |")
print("|---|---|---|---|" + "---|" * len(REFERENCIAS) + "---|")
for (variante, modo, ev), c in sorted(cuentas.items(), key=lambda kv: (kv[0][0], kv[0][1], kv[0][2] == "todas", kv[0][2])):
    n = c["n"]
    celdas = []
    for r in REFERENCIAS:
        t, a = c[r + " tabla"], c[r + " ampliada"]
        celda = f"{t} · {a}"
        if ev == "todas":
            celda += f" {wilson(a, n)}"
        celdas.append(celda)
    cumplen = str(c["cumplen"]) if variante.endswith("testigo-en-referencia") else ""
    print(f"| {variante} | {modo} | {ev} | {n} | " + " | ".join(celdas) + f" | {cumplen} |")
