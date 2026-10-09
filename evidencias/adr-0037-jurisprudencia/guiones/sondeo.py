#!/usr/bin/env python3
"""Abre las sesiones de un sondeo de la validación del juez de jurisprudencia y escribe su informe.

    sondeo.py <repositorio> <informe.json> --modo sin-skill|con-skill --claude <ruta> --kitlegal <ruta>
              [--preguntas id[,id]] [--repeticiones n] [--concurrencia n] [--modelo id]

No es el sondeo del producto (make evals-sondeo) ni da un veredicto: reúne respuestas para etiquetarlas. Cada sesión
es la orden de scripts/evals-sesion.sh, con tres diferencias que se declaran: usa el inicio de sesión de la cuenta de
quien lo lanza, así que el HOME es el suyo y las fuentes de ajustes se limitan (ninguna sin la skill; solo el
proyecto con ella); no declara ningún servidor MCP (--strict-mcp-config); y sin la skill no tiene herramientas. Con
la skill, las tres skills se instalan en un proyecto vacío con `kitlegal skills install --host claude` y el binario
va delante en el PATH: es el modo orden. Las variables de la sesión son las del job: el proxy que rechaza toda
petición salvo la del modelo, la caché en el temporal de la sesión y sin tráfico no esencial.

El informe no lleva el texto de las preguntas, que compone preguntas.py: tres llevan el fragmento de una sentencia.

Abre sesiones con modelo y consume la suscripción: lo lanza una persona, o una sesión interactiva a petición suya.
"""
import argparse, concurrent.futures, json, os, shutil, subprocess, sys, tempfile, time

sys.path.insert(0, os.path.dirname(os.path.abspath(__file__)))
from preguntas import preguntas as todas_las_preguntas  # noqa: E402

TOPE = 240


def entorno_de(opciones, dir_sesion):
    temporal = os.path.join(dir_sesion, "tmp")
    rutas = [os.path.dirname(opciones.claude), os.path.dirname(shutil.which("node")), "/usr/bin", "/bin", "/usr/sbin", "/sbin"]
    if opciones.modo == "con-skill":
        rutas.insert(0, os.path.dirname(opciones.kitlegal))
    entorno = {k: os.environ[k] for k in ("HOME", "USER", "LOGNAME", "LANG", "SHELL", "TERM") if k in os.environ}
    entorno.update({
        "PATH": ":".join(rutas),
        "KITLEGAL_CACHE_DIR": os.path.join(dir_sesion, "cache"),
        "HTTP_PROXY": "http://127.0.0.1:9", "HTTPS_PROXY": "http://127.0.0.1:9",
        "http_proxy": "http://127.0.0.1:9", "https_proxy": "http://127.0.0.1:9",
        "NO_PROXY": "api.anthropic.com", "no_proxy": "api.anthropic.com",
        "CLAUDE_CODE_DISABLE_NONESSENTIAL_TRAFFIC": "1", "CLAUDE_CODE_SUBPROCESS_ENV_SCRUB": "0",
        "TMPDIR": temporal, "CLAUDE_CODE_TMPDIR": temporal,
    })
    return entorno


def abrir(opciones, pregunta, repeticion, raiz):
    sesion = f"{pregunta['id']}-{opciones.modo}-{repeticion:02d}"
    dir_sesion = os.path.join(raiz, sesion)
    trabajo = os.path.join(dir_sesion, "trabajo")
    for d in (trabajo, os.path.join(dir_sesion, "tmp"), os.path.join(dir_sesion, "cache")):
        os.makedirs(d)
    entorno = entorno_de(opciones, dir_sesion)
    orden = [opciones.claude, "-p", pregunta["pregunta"], "--model", opciones.modelo, "--output-format", "stream-json",
             "--verbose", "--max-turns", "30", "--no-session-persistence", "--strict-mcp-config"]
    if opciones.modo == "con-skill":
        subprocess.run([opciones.kitlegal, "skills", "install", "--host", "claude"], cwd=trabajo, env=entorno,
                       capture_output=True, check=True)
        orden += ["--setting-sources", "project", "--settings", '{"sandbox":{"enabled":false}}',
                  "--permission-mode", "bypassPermissions", "--disallowedTools", "WebFetch", "WebSearch"]
    else:
        orden += ["--setting-sources", "", "--tools", ""]
    comienzo = time.time()
    try:
        hecho = subprocess.run(orden, cwd=trabajo, env=entorno, capture_output=True, text=True, timeout=TOPE, stdin=subprocess.DEVNULL)
        salida, codigo, fin = hecho.stdout, hecho.returncode, ""
        if not salida.strip():
            fin = "sin salida: " + hecho.stderr[:300]
    except subprocess.TimeoutExpired as agotado:
        salida, codigo, fin = (agotado.stdout or b"").decode("utf-8", "replace") if isinstance(agotado.stdout, bytes) else (agotado.stdout or ""), -1, "tope de tiempo"
    with open(os.path.join(raiz, sesion + ".jsonl"), "w", encoding="utf-8") as transcrito:
        transcrito.write(salida)
    llamadas, inicio, resultado = {}, {}, {}
    invocaciones = []
    for linea in salida.splitlines():
        try:
            evento = json.loads(linea)
        except json.JSONDecodeError:
            continue
        if evento.get("type") == "system" and evento.get("subtype") == "init":
            inicio = evento
        elif evento.get("type") == "assistant":
            for parte in evento["message"].get("content", []):
                if parte.get("type") == "tool_use":
                    llamadas[parte["id"]] = {"herramienta": parte["name"], "entrada": parte["input"]}
                    invocaciones.append(llamadas[parte["id"]])
        elif evento.get("type") == "user":
            contenido = evento["message"].get("content", [])
            for parte in contenido if isinstance(contenido, list) else []:
                if parte.get("type") == "tool_result" and parte.get("tool_use_id") in llamadas:
                    texto = parte.get("content")
                    if isinstance(texto, list):
                        texto = "\n".join(p.get("text", "") for p in texto if isinstance(p, dict))
                    llamadas[parte["tool_use_id"]].update({"salida": texto or "", "error": bool(parte.get("is_error"))})
        elif evento.get("type") == "result":
            resultado = evento
    return {
        "sesion": sesion, "pregunta": pregunta["id"], "modo": opciones.modo, "modelo": opciones.modelo,
        "modelos_de_la_sesion": sorted((resultado.get("modelUsage") or {}).keys()),
        "version_de_claude_code": inicio.get("claude_code_version"),
        "respuesta": resultado.get("result") or "", "invocaciones": invocaciones,
        "activada": [i["entrada"].get("skill") for i in invocaciones if i["herramienta"] == "Skill"],
        "codigo_de_la_sesion": codigo, "fin_de_la_sesion": fin or f"{resultado.get('type')} {resultado.get('subtype')}",
        "error": bool(resultado.get("is_error")), "turnos": resultado.get("num_turns"), "segundos": round(time.time() - comienzo, 1),
        "inicio": {k: inicio.get(k) for k in ("tools", "mcp_servers", "skills", "plugins", "agents", "slash_commands", "permissionMode", "model")},
    }


def main():
    a = argparse.ArgumentParser()
    a.add_argument("repositorio"); a.add_argument("informe")
    a.add_argument("--modo", required=True, choices=["sin-skill", "con-skill"])
    a.add_argument("--claude", required=True); a.add_argument("--kitlegal", required=True)
    a.add_argument("--preguntas", dest="ids", default=""); a.add_argument("--repeticiones", type=int, default=3)
    a.add_argument("--concurrencia", type=int, default=3); a.add_argument("--modelo", default="claude-sonnet-5-5")
    a.add_argument("--transcritos", default="")
    opciones = a.parse_args()
    if os.environ.get("KITLEGAL_PASO_DE_WORKFLOW") or os.environ.get("KITLEGAL_RUN_TESTIGO"):
        sys.exit("sondeo: un paso del workflow no abre sesiones con modelo (ADR 0032)")
    ids = {i for i in opciones.ids.split(",") if i}
    preguntas = [p for p in todas_las_preguntas(opciones.repositorio) if not ids or p["id"][:2] in ids]
    raiz = opciones.transcritos or tempfile.mkdtemp(prefix="kitlegal-sondeo-juez-")
    os.makedirs(raiz, exist_ok=True)
    tareas = [(p, r) for p in preguntas for r in range(1, opciones.repeticiones + 1)]
    with concurrent.futures.ThreadPoolExecutor(opciones.concurrencia) as hilos:
        sesiones = list(hilos.map(lambda t: abrir(opciones, t[0], t[1], raiz), tareas))
    comun = sesiones[0]["inicio"] if sesiones else {}
    json.dump({"sondeo": opciones.modo, "modelo": opciones.modelo, "repeticiones": opciones.repeticiones,
               "preguntas": [{"id": p["id"], "origen": p["origen"]} for p in preguntas], "inicio": comun,
               "sesiones": [{k: v for k, v in s.items() if k != "inicio"} for s in sesiones]},
              open(opciones.informe, "w", encoding="utf-8"), ensure_ascii=False, indent=2)
    for s in sesiones:
        print(s["sesion"], s["fin_de_la_sesion"], s["segundos"], "s", len(s["respuesta"]), "car.", s["modelos_de_la_sesion"], s["activada"], file=sys.stderr)
    print("transcritos en", raiz, file=sys.stderr)


if __name__ == "__main__":
    main()
