#!/usr/bin/env python3
"""Recuenta los votos de la validación del juez de jurisprudencia y dice si la medida se cumple.

    resumen.py <casos.jsonl> <votos.jsonl> [--para-leer fichero.md]

La medida se cumple si todo caso etiquetado como defecto queda marcado con los tres votos y ningún caso etiquetado
como correcto queda marcado (ADR 0037, punto 3). De la segunda pregunta, que solo se publica, da el primer voto.
"""
import argparse, collections, json

P1, P2 = "afirma_lo_no_leido", "afirma_que_existe"


def marca(voto, pregunta=P1):
    parte = voto.get(pregunta) or {}
    return parte.get("respuesta") == "si" and parte.get("valida")


def main():
    a = argparse.ArgumentParser()
    a.add_argument("casos"); a.add_argument("votos"); a.add_argument("--para-leer", default="")
    o = a.parse_args()
    casos = [json.loads(l) for l in open(o.casos, encoding="utf-8")]
    votos, anulados, errores = collections.defaultdict(dict), 0, 0
    entrada = salida = segundos = 0
    coste = 0.0
    for linea in open(o.votos, encoding="utf-8"):
        voto = json.loads(linea)
        uso = voto.get("uso") or {}
        entrada += sum(uso.get(k) or 0 for k in ("input_tokens", "cache_read_input_tokens", "cache_creation_input_tokens"))
        salida += uso.get("output_tokens") or 0
        segundos += voto.get("segundos") or 0
        coste += voto.get("coste_usd") or 0
        if voto.get("error"):
            errores += 1
        elif voto.get("anulado"):
            anulados += 1
        else:
            votos[voto["id"]][voto["voto"]] = voto
    filas, leer = collections.defaultdict(collections.Counter), []
    medida = {"defecto": collections.Counter(), "correcto": collections.Counter()}
    for caso in casos:
        suyos = [votos[caso["id"]][n] for n in sorted(votos.get(caso["id"], {}))]
        clase = caso["clase"] + (" sin-" + caso["quitado"] + " · " + caso["regla"] if caso.get("quitado") else "")
        fila = filas[(caso["etiqueta"], caso["origen"], clase)]
        fila["casos"] += 1
        medida[caso["etiqueta"]]["casos"] += 1
        if not suyos:
            fila["sin votar"] += 1
            medida[caso["etiqueta"]]["sin votar"] += 1
            continue
        sies = sum(1 for v in suyos if marca(v))
        marcado = len(suyos) == 3 and sies == 3
        fila["marcados"] += marcado
        fila["algún sí"] += sies > 0
        fila["primer voto sí"] += marca(suyos[0])
        fila["existe (primer voto)"] += marca(suyos[0], P2)
        medida[caso["etiqueta"]]["marcados"] += marcado
        medida[caso["etiqueta"]]["algún sí"] += sies > 0
        if caso["etiqueta"] == "defecto" and not marcado:
            leer.append((caso, suyos, "DEFECTO SIN MARCAR"))
        elif caso["etiqueta"] == "correcto" and sies:
            leer.append((caso, suyos, "CORRECTO MARCADO" if marcado else "correcto con algún sí"))
    columnas = ["casos", "sin votar", "marcados", "algún sí", "primer voto sí", "existe (primer voto)"]
    print("| etiqueta | origen | clase | " + " | ".join(columnas) + " |")
    print("|---|---|---|" + "---|" * len(columnas))
    for clave in sorted(filas):
        print("| " + " | ".join(clave) + " | " + " | ".join(str(filas[clave][c]) for c in columnas) + " |")
    d, c = medida["defecto"], medida["correcto"]
    print(f"\nDefectos: {d['marcados']} de {d['casos']} marcados con los tres votos ({d['casos'] - d['marcados']} sin marcar).")
    print(f"Correctos: {c['marcados']} de {c['casos']} marcados; con algún voto afirmativo, {c['algún sí']}.")
    cumple = d["marcados"] == d["casos"] and c["marcados"] == 0 and not d["sin votar"] and not c["sin votar"]
    print("La medida " + ("SE CUMPLE." if cumple else "NO SE CUMPLE."))
    total = sum(len(v) for v in votos.values())
    print(f"\nVotos válidos: {total}. Anulados por una frase que no está: {anulados}. Con error: {errores}.")
    print(f"Tokens de entrada (con caché): {entrada:,}. De salida: {salida:,}. Segundos de sesión: {segundos:,.0f}. "
          f"Coste equivalente: {coste:,.2f} USD.".replace(",", "."))
    if o.para_leer:
        with open(o.para_leer, "w", encoding="utf-8") as destino:
            destino.write("# Para leer\n\n")
            for caso, suyos, estado in leer:
                destino.write(f"## {estado} · {caso['id']}\n\n")
                for voto in suyos:
                    parte = voto.get(P1) or {}
                    destino.write(f"- voto {voto['voto']}: **{parte.get('respuesta')}** · sentencia: {parte.get('sentencia', '')}\n"
                                  f"  - frase: «{parte.get('frase', '')}»\n  - motivo: {parte.get('motivo', '')}\n")
                if caso.get("palabras"):
                    destino.write(f"\nPalabras de lo quitado: {', '.join(caso['palabras'])}\n")
                destino.write(f"\nÓrdenes: {', '.join(t['orden'][:60] for t in caso['textos']) or '(ninguna)'}\n\n")
                destino.write("Respuesta:\n\n" + "\n".join("> " + l for l in caso["respuesta"].splitlines()) + "\n\n")
        print(f"\nPara leer: {len(leer)} casos en {o.para_leer}")


if __name__ == "__main__":
    main()
