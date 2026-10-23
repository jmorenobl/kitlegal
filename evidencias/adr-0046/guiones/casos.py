#!/usr/bin/env python3
"""Construye los casos etiquetados de la validación del juez de las descripciones (ADR 0046), sin modelo.

    casos.py <historicas.json> <casos.jsonl>

Cada descripción que ha tenido una skill del repositorio es un defecto leído: todas dicen «Úsala». De cada una sale
una base en tercera persona, con unas sustituciones fijas, y de cada base, por construcción, los derivados: los que
meten una forma de primera o de segunda persona son defectos, y los que la meten dentro de una cita o cambian la
forma por otra impersonal o de tercera persona son correctos. La etiqueta la pone la construcción, antes del primer
voto, y cada defecto lleva la forma que lo hace serlo, para comprobar sin modelo que la frase del juez la contiene.
"""
import json
import sys

A_TERCERA = [
    ("Úsala siempre que", "Se usa siempre que"),
    ("Úsala cuando", "Se usa cuando"),
    ("también cuando creas conocer la respuesta", "también cuando el agente crea conocer la respuesta"),
    ("también cuando creas conocerla", "también cuando el agente crea conocerla"),
    ("Actívala antes de llamar a", "Se activa antes de llamar a"),
]
SEGUNDA_EN_LA_LEIDA = ("Úsala", "Actívala", "creas")


def sustituir(texto, viejo, nuevo):
    if viejo not in texto:
        return None
    return texto.replace(viejo, nuevo, 1)


def anadir(texto, frase):
    return texto.rstrip() + " " + frase


def derivados(base):
    """(clase, etiqueta, forma, descripción) de cada derivado que la base admite."""
    defectos = [
        ("imperativo de tú", "Úsala", sustituir(base, "Se usa", "Úsala")),
        ("imperativo de usted", "Úsela", sustituir(base, "Se usa", "Úsela")),
        ("imperativo de tú: activar", "Actívala", sustituir(base, "Se activa antes", "Actívala antes")),
        ("subjuntivo de tú", "creas", sustituir(base, "el agente crea conocer", "creas conocer")),
        ("pronombre de tú", "te preguntan", anadir(base, "Si te preguntan por una norma, puedes leerla con ella.")),
        ("primera del singular", "Te ayudo", anadir(base, "Te ayudo a encontrar el precepto y a citarlo.")),
        ("primera del plural", "nuestras", anadir(base, "Con nuestras herramientas, la cita sale del texto oficial.")),
        ("posesivo de tú", "tu pregunta", sustituir(base, "la pregunta", "tu pregunta")),
        ("posesivo de tú", "tu respuesta",
         None if "la pregunta" in base else sustituir(base, "la respuesta", "tu respuesta")),
        ("pronombre de vosotros", "os piden", anadir(base, "Si os piden un plazo, se calcula con kitlegal.")),
    ]
    correctos = [
        ("base en tercera persona", base),
        ("primera y segunda dentro de una cita",
         anadir(base, "Por ejemplo: «¿hasta cuándo puedo recurrir?» o «¿puedes decirme qué dice el artículo 21?».")),
        ("impersonal", sustituir(base, "Se usa", "Conviene activarla")),
        ("el agente en tercera persona",
         sustituir(base, "Se activa antes", "El agente la activa antes") or sustituir(base, "Se usa", "El agente la usa")),
        ("verbo que se escribe como un imperativo, con la skill de sujeto",
         None if "Lee " in base else anadir(base, "Lee la norma con kitlegal y responde citando identificador y bloque.")),
    ]
    for clase, forma, texto in defectos:
        if texto is not None:
            yield clase, "defecto", forma, texto
    for clase, texto in correctos:
        if texto is not None:
            yield clase, "correcto", "", texto


def main():
    historicas = json.load(open(sys.argv[1], encoding="utf-8"))
    casos, vistos = [], set()

    def anotar(caso):
        if caso["descripcion"] in vistos:
            return
        vistos.add(caso["descripcion"])
        caso["id"] = f"{len(casos) + 1:03d}"
        casos.append(caso)

    for leida in historicas:
        origen = f"{leida['skill']}@{leida['commit']}"
        texto = leida["descripcion"]
        assert "Úsala" in texto, origen
        anotar({"skill": leida["skill"], "origen": origen, "clase": "leída", "etiqueta": "defecto",
                "forma": "Úsala", "descripcion": texto})
        base = texto
        for viejo, nuevo in A_TERCERA:
            base = base.replace(viejo, nuevo)
        restos = [forma for forma in SEGUNDA_EN_LA_LEIDA if forma in base]
        assert not restos, f"{origen}: la base conserva {restos}"
        for clase, etiqueta, forma, derivado in derivados(base):
            anotar({"skill": leida["skill"], "origen": origen, "clase": clase, "etiqueta": etiqueta, "forma": forma,
                    "descripcion": derivado})

    with open(sys.argv[2], "w", encoding="utf-8") as destino:
        for caso in casos:
            destino.write(json.dumps(caso, ensure_ascii=False) + "\n")

    defectos = sum(caso["etiqueta"] == "defecto" for caso in casos)
    print(f"{len(casos)} casos: {defectos} defectos y {len(casos) - defectos} correctos")


if __name__ == "__main__":
    main()
