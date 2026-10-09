#!/usr/bin/env python3
"""Vota los casos de la validación del juez de jurisprudencia (ADR 0037) con el juez con modelo.

Es guiones/juez.py de evidencias/adr-0037/ con las preguntas de esta rúbrica, el claude del juez por su ruta y sin grupos.

    juez.py <casos.jsonl> <votos.jsonl> --claude <ruta> [--votos 1|3] [--por-orden] [--clase c[,c]]
            [--concurrencia n] [--limite n] [--humo] [--simulado] [--modelo id]

Cada voto es una sesión nueva de Claude Code sin herramientas, sin configuración de la cuenta ni del proyecto y
con la rúbrica como prompt de sistema. Abre sesiones con modelo: lo lanza una persona, con CLAUDE_CODE_OAUTH_TOKEN
en el entorno, o con --credencial sesion y el inicio de sesión de su cuenta. Se puede interrumpir y relanzar: no repite los votos ya escritos.

--por-orden pide el segundo y el tercer voto solo si los anteriores marcaron la pregunta 1 con su frase
comprobada, que es como decide la unanimidad. Un «si» cuya frase no está en la respuesta es nulo y se repite una
vez. --simulado no abre ninguna sesión: prueba el guion con votos inventados.
"""
import argparse
import concurrent.futures
import hashlib
import json
import os
import re
import subprocess
import sys
import tempfile
import threading
import time

AQUI = os.path.dirname(os.path.abspath(__file__))
PREGUNTAS = ("afirma_lo_no_leido", "afirma_que_existe")
LIMITE = r"limit|límite|rate|overloaded|credit|quota|429|529"
ESPERA = 600
escritura = threading.Lock()
parar = threading.Event()


def normal(texto):
    """Sin énfasis de Markdown y con los blancos colapsados: la tolerancia de las formas fijas del job."""
    return re.sub(r"\s+", " ", re.sub(r"[*_`]", "", texto)).strip()


def frase_esta(frase, respuesta):
    return bool(normal(frase)) and normal(frase) in normal(respuesta)


def prompt_de(caso):
    partes = ["<textos_de_las_herramientas>"]
    if not caso["textos"]:
        partes.append("(ninguna herramienta devolvió ningún texto)")
    for texto in caso["textos"]:
        salida = texto["salida"].strip() or "(sin salida)"
        partes.append(f'<texto orden="{texto["orden"]}">\n{salida}\n</texto>')
    partes.append("</textos_de_las_herramientas>")
    partes.append(f"<pregunta>\n{caso['pregunta']}\n</pregunta>")
    partes.append(f"<respuesta>\n{caso['respuesta']}\n</respuesta>")
    partes.append("Responde a las dos preguntas de la rúbrica sobre esta respuesta.")
    return "\n\n".join(partes)


def voto_simulado(caso, numero):
    semilla = int(hashlib.sha256(f"{caso['id']}/{numero}".encode()).hexdigest(), 16)
    marca = caso["etiqueta"] == "defecto" and semilla % 10 < 8 or semilla % 50 == 0
    frase = caso["respuesta"].strip()[:80] if marca else ""
    if marca and semilla % 13 == 0:
        frase = "una frase que no está en la respuesta"
    return {"afirma_lo_no_leido": {"motivo": "simulado", "respuesta": "si" if marca else "no", "frase": frase,
                                   "sentencia": "simulado" if marca else ""},
            "afirma_que_existe": {"motivo": "simulado", "respuesta": "no", "frase": ""}}, {}, ""


def voto_real(caso, opciones, aislado):
    orden = [opciones.claude, "-p", "--model", opciones.modelo, "--tools", "", "--strict-mcp-config",
             "--disable-slash-commands", "--no-session-persistence", "--system-prompt", opciones.rubrica,
             "--output-format", "json", "--json-schema", opciones.esquema]
    if opciones.credencial == "token":
        entorno = {"PATH": os.environ["PATH"], "HOME": aislado, "CLAUDE_CONFIG_DIR": os.path.join(aislado, "config"),
                   "CLAUDE_CODE_OAUTH_TOKEN": os.environ["CLAUDE_CODE_OAUTH_TOKEN"]}
    else:
        # Con el inicio de sesión de la cuenta: sin ninguna fuente de ajustes, y sin nada del entorno de la sesión
        # que lo lanza. El directorio de trabajo sigue siendo uno vacío.
        orden += ["--setting-sources", ""]
        entorno = {k: os.environ[k] for k in ("PATH", "HOME", "USER", "LANG", "TMPDIR") if k in os.environ}
    hecho = subprocess.run(orden, input=prompt_de(caso), text=True, capture_output=True, env=entorno,
                           cwd=os.path.join(aislado, "cwd"), timeout=900, check=False)
    crudo = hecho.stdout.strip()
    try:
        sobre = json.loads(crudo)
    except json.JSONDecodeError:
        return None, {}, f"salida que no es JSON (código {hecho.returncode}): {crudo[:300]} {hecho.stderr[:300]}"
    uso = {"duracion_ms": sobre.get("duration_ms"), "uso": sobre.get("usage"),
           "modelos": sorted((sobre.get("modelUsage") or {}).keys()), "coste_usd": sobre.get("total_cost_usd")}
    if sobre.get("is_error"):
        return None, uso, f"sesión con error: {str(sobre.get('result'))[:300]}"
    juicio = sobre.get("structured_output")
    if juicio is None:
        texto = re.sub(r"^```(?:json)?\s*|\s*```$", "", str(sobre.get("result", "")).strip())
        try:
            juicio = json.loads(texto)
        except json.JSONDecodeError:
            return None, uso, f"respuesta sin la forma pedida: {texto[:300]}"
    return juicio, uso, ""


def votar(caso, numero, intento, opciones, aislado):
    comienzo = time.time()
    if opciones.simulado:
        juicio, uso, error = voto_simulado(caso, f"{numero}/{intento}")
    else:
        juicio, uso, error = voto_real(caso, opciones, aislado)
    # Un límite de uso no es un voto: se espera y se repite, hasta el tope de espera.
    esperado = 0
    while error and re.search(LIMITE, error, re.I) and esperado < opciones.espera_maxima and not opciones.simulado:
        print(f"  límite de uso en {caso['id']}: espero {ESPERA // 60} min ({error[:120]})", file=sys.stderr)
        time.sleep(ESPERA)
        esperado += ESPERA
        juicio, uso, error = voto_real(caso, opciones, aislado)
    voto = {"id": caso["id"], "voto": numero, "intento": intento, "segundos": round(time.time() - comienzo, 1),
            "error": error, **uso}
    if juicio is not None:
        for pregunta in PREGUNTAS:
            parte = dict(juicio.get(pregunta) or {})
            parte["valida"] = parte.get("respuesta") != "si" or frase_esta(parte.get("frase", ""), caso["respuesta"])
            voto[pregunta] = parte
    return voto


def marca(voto):
    parte = voto.get("afirma_lo_no_leido") or {}
    return parte.get("respuesta") == "si" and parte.get("valida")


def nulo(voto):
    return any((voto.get(p) or {}).get("valida") is False for p in PREGUNTAS)


def juzgar_caso(caso, hechos, opciones, aislado, destino):
    """Los votos que faltan de un caso, seguidos, para que compartan la caché del prompt."""
    votos = dict(hechos)
    for numero in range(1, opciones.votos + 1):
        if parar.is_set():
            return
        if numero not in votos:
            voto = votar(caso, numero, 1, opciones, aislado)
            if not voto["error"] and nulo(voto):
                with escritura:
                    destino.write(json.dumps({**voto, "anulado": True}, ensure_ascii=False) + "\n")
                    destino.flush()
                voto = votar(caso, numero, 2, opciones, aislado)
            with escritura:
                destino.write(json.dumps(voto, ensure_ascii=False) + "\n")
                destino.flush()
            if voto["error"]:
                print(f"  {caso['id']} voto {numero}: {voto['error']}", file=sys.stderr)
                if re.search(LIMITE, voto["error"], re.I):
                    parar.set()
                return
            votos[numero] = voto
        if opciones.por_orden and not marca(votos[numero]):
            return


def main():
    analizador = argparse.ArgumentParser()
    analizador.add_argument("casos")
    analizador.add_argument("votos_escritos")
    analizador.add_argument("--claude", default="claude")
    analizador.add_argument("--votos", type=int, default=3, choices=[1, 2, 3])
    analizador.add_argument("--por-orden", action="store_true")
    analizador.add_argument("--clase", default="")
    analizador.add_argument("--concurrencia", type=int, default=3)
    analizador.add_argument("--limite", type=int, default=0)
    analizador.add_argument("--humo", action="store_true")
    analizador.add_argument("--simulado", action="store_true")
    analizador.add_argument("--modelo", default="claude-opus-5-5")
    analizador.add_argument("--credencial", default="token", choices=["token", "sesion"])
    analizador.add_argument("--espera-maxima", type=int, default=6 * 3600)
    opciones = analizador.parse_args()
    opciones.rubrica = open(os.path.join(os.path.dirname(AQUI), "rubrica.md"), encoding="utf-8").read()
    opciones.esquema = open(os.path.join(os.path.dirname(AQUI), "esquema.json"), encoding="utf-8").read().strip()

    if not opciones.simulado and opciones.credencial == "token" and not os.environ.get("CLAUDE_CODE_OAUTH_TOKEN"):
        sys.exit("juez: falta CLAUDE_CODE_OAUTH_TOKEN en el entorno, o --credencial sesion para usar el inicio de sesión de la cuenta")
    if os.environ.get("KITLEGAL_PASO_DE_WORKFLOW") or os.environ.get("KITLEGAL_RUN_TESTIGO"):
        sys.exit("juez: un paso del workflow no abre sesiones con modelo (ADR 0032)")

    clases = {c for c in opciones.clase.split(",") if c}
    casos = [json.loads(linea) for linea in open(opciones.casos, encoding="utf-8")]
    casos = [c for c in casos if not clases or c["clase"] in clases]
    if opciones.limite:
        casos = casos[:opciones.limite]

    hechos = {}
    if os.path.exists(opciones.votos_escritos):
        for linea in open(opciones.votos_escritos, encoding="utf-8"):
            voto = json.loads(linea)
            if not voto.get("error") and not voto.get("anulado"):
                hechos.setdefault(voto["id"], {})[voto["voto"]] = voto

    aislado = tempfile.mkdtemp(prefix="kitlegal-juez-")
    os.makedirs(os.path.join(aislado, "config"))
    os.makedirs(os.path.join(aislado, "cwd"))

    if opciones.humo:
        voto = votar(casos[0], 1, 1, opciones, aislado)
        print(json.dumps(voto, ensure_ascii=False, indent=2))
        return

    print(f"juez: {len(casos)} casos, hasta {opciones.votos} votos cada uno, "
          f"{'por orden' if opciones.por_orden else 'todos'}, modelo {opciones.modelo}", file=sys.stderr)
    with open(opciones.votos_escritos, "a", encoding="utf-8") as destino, \
            concurrent.futures.ThreadPoolExecutor(opciones.concurrencia) as hilos:
        pendientes = [hilos.submit(juzgar_caso, caso, hechos.get(caso["id"], {}), opciones, aislado, destino)
                      for caso in casos]
        for numero, pendiente in enumerate(concurrent.futures.as_completed(pendientes), 1):
            pendiente.result()
            if numero % 20 == 0:
                print(f"  {numero}/{len(casos)} casos", file=sys.stderr)
    if parar.is_set():
        sys.exit("juez: parado por un límite de uso o un error de la cuenta; relánzalo después: sigue donde lo dejó")


if __name__ == "__main__":
    main()
