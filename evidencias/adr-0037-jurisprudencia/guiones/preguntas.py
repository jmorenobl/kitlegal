#!/usr/bin/env python3
"""Compone las preguntas de los dos sondeos de la validación del juez de jurisprudencia (ADR 0037).

    preguntas.py <repositorio>

Las lee de preguntas.json, que dice de dónde sale cada una: de una eval de jurisprudencia en el commit del informe
del cierre de H23, o nueva. El texto de la sentencia que llevan tres de las nuevas es el fragmento de
evidencias/adr-0036/, o su ficha, que se pone aquí byte a byte y nunca se teclea.
"""
import json
import os
import re
import subprocess
import sys

AQUI = os.path.dirname(os.path.abspath(__file__))
DATOS = json.load(open(os.path.join(os.path.dirname(AQUI), "preguntas.json"), encoding="utf-8"))
COMMIT_DEL_INFORME = DATOS["commit"]
MARCA_DEL_FALLO = "F A L L O"


def pregunta_de(repo, commit, fichero, vistas={}):
    """La pregunta de una eval de jurisprudencia en un commit."""
    if fichero not in vistas:
        texto = subprocess.run(["git", "-C", repo, "show", f"{commit}:evals/jurisprudencia/{fichero}"],
                               capture_output=True, text=True, check=True).stdout
        casa = re.search(r'^pregunta:\s*"(.*)"\s*$', texto, re.M)
        if casa:
            vistas[fichero] = casa.group(1).replace('\\"', '"')
        else:  # escalar de bloque: las líneas con dos espacios de sangría, hasta la siguiente clave
            cuerpo = []
            for linea in texto.split("pregunta: |\n", 1)[1].split("\n"):
                if linea and not linea.startswith("  "):
                    break
                cuerpo.append(linea[2:])
            vistas[fichero] = "\n".join(cuerpo).rstrip("\n") + "\n"
    return vistas[fichero]


def preguntas(repo):
    """Las doce preguntas, cada una con su id, su origen y su texto."""
    fragmento = open(os.path.join(repo, DATOS["fragmento"]), encoding="utf-8").read()
    ficha = fragmento[:fragmento.index("\n\n")] + "\n"
    if not (ficha.startswith("Roj:") and ficha.rstrip().endswith("Tipo de Resolución: Sentencia") and MARCA_DEL_FALLO in fragmento):
        raise SystemExit(f"{DATOS['fragmento']} no tiene la forma esperada")
    lista = []
    for una in DATOS["preguntas"]:
        if "eval" in una:
            origen, texto = f"evals/jurisprudencia/{una['eval']}@{COMMIT_DEL_INFORME[:7]}", pregunta_de(repo, COMMIT_DEL_INFORME, una["eval"])
        else:
            origen, texto = "nueva", una["pregunta"].replace("{fragmento}", fragmento).replace("{ficha}", ficha)
        lista.append({"id": una["id"], "origen": origen, "pregunta": texto})
    return lista


if __name__ == "__main__":
    for una in preguntas(sys.argv[1]):
        print(una["id"], una["origen"], len(una["pregunta"]))
