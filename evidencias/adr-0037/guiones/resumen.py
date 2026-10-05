#!/usr/bin/env python3
"""Resume los votos de la validación del ADR 0037 y escribe lo que una persona tiene que leer.

    resumen.py <casos.jsonl> <votos.jsonl> --grupo ajuste|medida [--para-leer fichero.md]
"""
import argparse
import collections
import json

P1, P2 = "afirma_lo_no_leido", "cuenta_su_proceso"


def marca(voto, pregunta=P1):
    parte = voto.get(pregunta) or {}
    return parte.get("respuesta") == "si" and parte.get("valida")


def main():
    analizador = argparse.ArgumentParser()
    analizador.add_argument("casos")
    analizador.add_argument("votos")
    analizador.add_argument("--grupo", required=True)
    analizador.add_argument("--para-leer", default="")
    opciones = analizador.parse_args()

    casos = [json.loads(l) for l in open(opciones.casos, encoding="utf-8")]
    casos = [c for c in casos if c["grupo"] == opciones.grupo]
    votos, anulados, errores = collections.defaultdict(dict), 0, 0
    entrada = salida = segundos = 0
    coste = 0.0
    for linea in open(opciones.votos, encoding="utf-8"):
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

    filas = collections.defaultdict(collections.Counter)
    leer = []
    for caso in casos:
        suyos = [votos[caso["id"]][n] for n in sorted(votos.get(caso["id"], {}))]
        fila = filas[(caso["informe"], caso["clase"])]
        fila["casos"] += 1
        if not suyos:
            fila["sin votar"] += 1
            continue
        sies = sum(1 for v in suyos if marca(v))
        unanime = len(suyos) == 3 and sies == 3
        # Con la votación por orden, un caso con menos de tres votos tuvo un «no»: no es unánime.
        mayoria = sies >= 2 if len(suyos) == 3 else None
        fila["unanimidad"] += unanime
        if mayoria is not None:
            fila["mayoría"] += mayoria
            fila["con tres votos"] += 1
        fila["primer voto"] += marca(suyos[0])
        fila["proceso (primer voto)"] += marca(suyos[0], P2)
        dudoso = sies > 0 and not unanime
        fila["dudosos"] += dudoso
        if caso["clase"] == "sin_etiqueta" and sies:
            leer.append((caso, suyos, "MARCADA" if unanime else "dudosa"))
        elif caso["etiqueta"] == "defecto" and not unanime:
            leer.append((caso, suyos, "DEFECTO SIN MARCAR"))
        elif caso["clase"] == "sin_etiqueta" and opciones.grupo == "medida" and caso["eval"][:2] in ("19", "20"):
            # Las evals con una redacción cambiada se leen las marque el juez o no (ADR 0037, «Validación»).
            leer.append((caso, suyos, "sin marcar, eval " + caso["eval"][:2]))

    print(f"## Grupo {opciones.grupo}\n")
    columnas = ["casos", "sin votar", "primer voto", "unanimidad", "con tres votos", "mayoría", "dudosos",
                "proceso (primer voto)"]
    print("| informe | clase | " + " | ".join(columnas) + " |")
    print("|---|---|" + "---|" * len(columnas))
    for clave in sorted(filas):
        print(f"| {clave[0]} | {clave[1]} | " + " | ".join(str(filas[clave][c]) for c in columnas) + " |")
    total = sum(len(v) for v in votos.values())
    print(f"\nVotos válidos: {total}. Anulados por una frase que no está: {anulados}. Con error: {errores}.")
    print(f"Tokens de entrada (con caché): {entrada:,}. De salida: {salida:,}. Segundos de sesión: {segundos:,.0f}. "
          f"Coste equivalente: {coste:,.2f} USD.".replace(",", "."))

    if opciones.para_leer:
        with open(opciones.para_leer, "w", encoding="utf-8") as destino:
            destino.write(f"# Para leer: grupo {opciones.grupo}\n\n")
            for caso, suyos, estado in leer:
                destino.write(f"## {estado} · {caso['id']}\n\n")
                if caso.get("quitado"):
                    destino.write(f"Texto quitado: {caso['quitado']['norma']} {caso['quitado']['bloque']}\n\n")
                for voto in suyos:
                    parte = voto.get(P1) or {}
                    destino.write(f"- voto {voto['voto']}: **{parte.get('respuesta')}** · precepto: {parte.get('precepto', '')}\n"
                                  f"  - frase: «{parte.get('frase', '')}»\n  - motivo: {parte.get('motivo', '')}\n")
                destino.write(f"\nÓrdenes: {', '.join(t['orden'] for t in caso['textos'])}\n\n")
                destino.write("Respuesta:\n\n" + "\n".join("> " + l for l in caso["respuesta"].splitlines()) + "\n\n")
        print(f"\nPara leer: {len(leer)} casos en {opciones.para_leer}")


if __name__ == "__main__":
    main()
