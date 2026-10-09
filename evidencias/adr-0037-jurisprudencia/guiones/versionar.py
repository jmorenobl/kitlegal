#!/usr/bin/env python3
"""Escribe lo que se versiona de los casos y de los votos de la validación del juez de jurisprudencia.

    versionar.py <casos.jsonl> <votos.jsonl> <carpeta de la evidencia en el repositorio> <destino>

  - casos.yaml:  los casos por referencia a su informe y a su sesión, como los de evidencias/adr-0037/: nada de la
                 respuesta ni de los textos, que se reconstruyen con casos.py;
  - lectura.md:  las etiquetas de la lectura, con la frase que decide las de la frontera, para revisarlas;
  - votos.jsonl: los votos sin los datos de consumo.
"""
import collections, json, os, sys

CABECERA = """# Casos etiquetados de la clase afirma_lo_no_leido de jurisprudencia (ADR 0037; validación del 2026-10-09).
# Cada caso es la respuesta de una sesión de un informe versionado: el del job de evals del cierre de H23, o el de
# uno de los dos sondeos de esta carpeta. Nada está escrito a mano: la respuesta se lee del informe, y los textos de
# sus herramientas se reconstruyen repitiendo sus órdenes o se leen del informe del sondeo (guiones/casos.py).
# Un caso con `quitado` es un derivado: la misma respuesta con esa parte del texto pegado fuera de la pregunta
# —`documento`, todo, y con él las órdenes cotejar; `fallo`, desde «F A L L O»; `apartado-2`, el 2.º del fallo—.
# Procedencias: `lectura` es la del 2026-10-09, hecha por el agente antes del primer voto y adoptada por Jorge;
# `derivado`, por construcción, con la regla sin modelo de guiones/casos.py que dice `regla`.
# `frase` es la que tiene cerca de la raya una respuesta leída como correcta.
clase: afirma_lo_no_leido
casos:
"""


def comillas(texto):
    return json.dumps(texto, ensure_ascii=False)


def main():
    casos_de, votos_de, evidencia, destino = sys.argv[1:5]
    casos = [json.loads(l) for l in open(casos_de, encoding="utf-8")]
    with open(os.path.join(destino, "casos.yaml"), "w", encoding="utf-8") as salida:
        salida.write(CABECERA)
        for caso in casos:
            informe = caso["informe"] if caso["origen"] == "informe-h23" else f"{evidencia}/{caso['informe']}"
            salida.write(f"  - informe: {informe}\n    sesion: {caso['sesion']}\n")
            if caso.get("quitado"):
                salida.write(f"    quitado: {{texto: {caso['quitado']}}}\n")
            salida.write(f"    grupo: medida\n    etiqueta: {caso['etiqueta']}\n    procedencia: {caso['procedencia']}\n")
            if caso.get("regla"):
                salida.write(f"    regla: {caso['regla']}\n")
            if caso.get("frase"):
                salida.write(f"    frase: {comillas(caso['frase'])}\n")
    with open(os.path.join(destino, "lectura.md"), "w", encoding="utf-8") as salida:
        leidos = [c for c in casos if c["clase"] == "leido"]
        salida.write("# Lectura de las respuestas\n\nLas etiquetas de las respuestas leídas, del 2026-10-09: las del informe "
                     "del cierre de H23 y las de los dos sondeos. Se fijaron antes del primer voto. De una respuesta correcta "
                     "que está cerca de la raya se da la frase que lo decide; las demás correctas no tienen ninguna.\n")
        por_origen = collections.defaultdict(list)
        for caso in leidos:
            por_origen[caso["origen"]].append(caso)
        for origen, suyos in por_origen.items():
            salida.write(f"\n## {origen}\n\n| Sesión | Etiqueta | Frase |\n|---|---|---|\n")
            for caso in suyos:
                frase = "«" + caso["frase"].replace("|", "\\|") + "»" if caso.get("frase") else ""
                salida.write(f"| `{caso['sesion']}` | {caso['etiqueta']} | {frase} |\n")
    quitar = ("uso", "coste_usd", "duracion_ms", "segundos", "modelos")
    with open(os.path.join(destino, "votos.jsonl"), "w", encoding="utf-8") as salida:
        for linea in open(votos_de, encoding="utf-8"):
            voto = json.loads(linea)
            salida.write(json.dumps({k: v for k, v in voto.items() if k not in quitar}, ensure_ascii=False) + "\n")
    print(len(casos), "casos")


if __name__ == "__main__":
    main()
