#!/usr/bin/env python3
"""Recuenta por caso los votos de la validación del juez de las descripciones (ADR 0046), sin modelo.

    resumen.py <casos.jsonl> <votos.jsonl>

Un caso está marcado si sus tres votos válidos dicen «si» con su frase en la descripción, que es como decide la
unanimidad. De cada defecto marcado comprueba además, sin modelo, que alguna frase citada contiene su forma.
"""
import collections
import json
import sys

PREGUNTA = "habla_en_primera_o_segunda_persona"


def main():
    casos = {c["id"]: c for c in map(json.loads, open(sys.argv[1], encoding="utf-8"))}
    votos = collections.defaultdict(dict)
    errores = nulos = 0
    for voto in map(json.loads, open(sys.argv[2], encoding="utf-8")):
        if voto.get("anulado"):
            nulos += 1
            continue
        if voto.get("error"):
            errores += 1
            continue
        votos[voto["id"]][voto["voto"]] = voto

    filas = collections.Counter()
    fallos, sin_forma = [], []
    for ident, caso in casos.items():
        suyos = votos.get(ident, {})
        si = [v for v in suyos.values() if (v.get(PREGUNTA) or {}).get("respuesta") == "si" and v[PREGUNTA]["valida"]]
        marcado = len(si) == 3
        alguno = bool(si)
        filas[(caso["etiqueta"], caso["clase"], "casos")] += 1
        filas[(caso["etiqueta"], caso["clase"], "tres")] += marcado
        filas[(caso["etiqueta"], caso["clase"], "alguno")] += alguno
        if (caso["etiqueta"] == "defecto") != marcado:
            fallos.append((ident, caso["etiqueta"], caso["clase"], [v[PREGUNTA] for v in suyos.values()]))
        if marcado and caso["forma"] and not any(caso["forma"] in v[PREGUNTA]["frase"] for v in si):
            sin_forma.append((ident, caso["forma"], [v[PREGUNTA]["frase"] for v in si]))

    print("| Etiqueta | Clase | Casos | Marcados con los tres votos | Con algún voto afirmativo |")
    print("|---|---|---|---|---|")
    for etiqueta, clase in sorted({(e, c) for e, c, _ in filas}):
        print(f"| {etiqueta} | {clase} | {filas[(etiqueta, clase, 'casos')]} | {filas[(etiqueta, clase, 'tres')]} "
              f"| {filas[(etiqueta, clase, 'alguno')]} |")
    defectos = [c for c in casos.values() if c["etiqueta"] == "defecto"]
    correctos = [c for c in casos.values() if c["etiqueta"] == "correcto"]
    marcados = lambda grupo: sum(len([v for v in votos.get(c["id"], {}).values()
                                      if (v.get(PREGUNTA) or {}).get("respuesta") == "si" and v[PREGUNTA]["valida"]]) == 3
                                 for c in grupo)
    print(f"\ndefectos: {marcados(defectos)} de {len(defectos)} marcados; correctos: {marcados(correctos)} de "
          f"{len(correctos)} marcados")
    total = sum(len(v) for v in votos.values())
    print(f"votos válidos: {total}; nulos repetidos: {nulos}; con error: {errores}")
    for fallo in fallos:
        print("FALLO", json.dumps(fallo, ensure_ascii=False))
    for falta in sin_forma:
        print("SIN LA FORMA", json.dumps(falta, ensure_ascii=False))


if __name__ == "__main__":
    main()
