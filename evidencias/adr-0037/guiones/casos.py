#!/usr/bin/env python3
"""Construye los casos de la validación del ADR 0037 desde los textos reconstruidos.

    casos.py <dir de textos> <repositorio> <salida.jsonl>

Tres clases de caso, sin nada escrito a mano:
  - leido:     la respuesta de una sesión con la etiqueta «defecto» de la bitácora (docs/USO.md, 2026-09-30);
  - derivado:  la respuesta de una sesión que pasó, con el texto de un bloque que cita quitado: defecto por
               construcción;
  - sin_etiqueta: las demás respuestas, tal como están en su informe.
El grupo es «ajuste» para los informes de H7.1 a H7.3 y «medida» para los de H7.4, H21 y H22.
"""
import json
import re
import subprocess
import sys

DEFECTOS_LEIDOS = {  # docs/USO.md, 2026-09-30, clase B
    ("h7.1", "19", "01"), ("h7.2", "19", "01"), ("h7.2", "19", "02"), ("h7.3", "19", "01"), ("h7.3", "19", "02"),
}
GRUPO = {"h7.1": "ajuste", "h7.2": "ajuste", "h7.3": "ajuste", "h7.4": "medida", "h21": "medida", "h22": "medida"}
ORDEN = ["h7.1", "h7.2", "h7.3", "h7.4", "h21", "h22"]


def pregunta_de(repo, commit, fichero, vistas={}):
    clave = (commit, fichero)
    if clave not in vistas:
        texto = subprocess.run(["git", "-C", repo, "show", f"{commit}:evals/boe-legislacion/{fichero}"],
                               capture_output=True, text=True, check=True).stdout
        casa = re.search(r'^pregunta:\s*"(.*)"\s*$', texto, re.M)
        if not casa:
            raise SystemExit(f"sin pregunta: {commit}:{fichero}")
        vistas[clave] = casa.group(1).replace('\\"', '"')
    return vistas[clave]


def bloques_de(orden):
    """(norma, [bloques]) de una orden que lee bloques, o None."""
    partes = orden.replace("boe_articulos", "boe articulos").replace("boe_articulo", "boe articulo").split()
    if len(partes) < 4 or partes[0] != "boe" or partes[1] not in ("articulo", "articulos"):
        return None
    return partes[2], [p for p in partes[3:] if not p.startswith("--")]


def quitar_bloque(textos, norma, bloque):
    """Los textos sin el del bloque: fuera la salida de articulo, y el elemento de la de articulos."""
    quedan, quitados = [], 0
    for texto in textos:
        leidos = bloques_de(texto["orden"])
        if not leidos or leidos[0] != norma or bloque not in leidos[1] or texto["codigo"] != 0:
            quedan.append(texto)
            continue
        sobre = json.loads(texto["salida"])
        if isinstance(sobre["data"], list):
            resto = [b for b in sobre["data"] if b.get("bloque") != bloque]
            quitados += len(sobre["data"]) - len(resto)
            if resto:
                sobre["data"] = resto
                quedan.append({**texto, "salida": json.dumps(sobre, ensure_ascii=False)})
        else:
            quitados += 1
    return quedan, quitados


def main():
    textos_dir, repo, salida = sys.argv[1:4]
    casos, derivados_vistos = [], set()
    for informe in ORDEN:
        for linea in open(f"{textos_dir}/{informe}.jsonl", encoding="utf-8"):
            s = json.loads(linea)
            numero, repeticion = s["eval"][:2], s["sesion"][-2:]
            modo = s["modo"] or "orden"
            base = {
                "grupo": GRUPO[informe], "informe": informe, "sesion": s["sesion"], "eval": s["eval"], "modo": modo,
                "pasa": s["pasa"], "pregunta": pregunta_de(repo, s["commit"], s["eval"]), "respuesta": s["respuesta"],
            }
            textos = [{"orden": t["orden"], "codigo": t["codigo"], "salida": t["salida"]} for t in s["textos"]]
            leido = (informe, numero, repeticion) in DEFECTOS_LEIDOS
            casos.append({**base, "id": f"{informe}/{s['sesion']}/{modo}", "clase": "leido" if leido else "sin_etiqueta",
                          "etiqueta": "defecto" if leido else None, "textos": textos})
            # Un derivado por informe, eval y modo: la primera sesión que pasó, sin su primer bloque citado.
            clave = (informe, numero, modo)
            if s["pasa"] and s["citas_encontradas"] and clave not in derivados_vistos and s["respuesta"].strip():
                norma, bloque = s["citas_encontradas"][0].split()
                quedan, quitados = quitar_bloque(textos, norma, bloque)
                if quitados:
                    derivados_vistos.add(clave)
                    casos.append({**base, "id": f"{informe}/{s['sesion']}/{modo}/sin-{norma}-{bloque}", "clase": "derivado",
                                  "etiqueta": "defecto", "quitado": {"norma": norma, "bloque": bloque}, "textos": quedan})
    with open(salida, "w", encoding="utf-8") as destino:
        for caso in casos:
            destino.write(json.dumps(caso, ensure_ascii=False) + "\n")
    cuenta = {}
    for caso in casos:
        clave = (caso["grupo"], caso["clase"])
        cuenta[clave] = cuenta.get(clave, 0) + 1
    for clave in sorted(cuenta):
        print(*clave, cuenta[clave])
    print("total", len(casos))


if __name__ == "__main__":
    main()
