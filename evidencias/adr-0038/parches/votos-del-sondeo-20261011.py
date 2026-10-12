"""Vota con el juez, con la orden de scripts/evals-voto.sh, los mensajes volcados de un sondeo: un voto, y dos más si el primero es «si»."""
import json, subprocess, sys, pathlib, concurrent.futures as cf
mensajes, destino, claude, juez = map(pathlib.Path, sys.argv[1:5])
rubrica = (juez / "rubrica.md").read_text(encoding="utf-8")
esquema = (juez / "esquema.json").read_text(encoding="utf-8").rstrip("\n")
def voto(mensaje, n):
    r = subprocess.run([str(claude), "-p", "--model", "claude-opus-5-5", "--tools", "", "--strict-mcp-config",
        "--disable-slash-commands", "--no-session-persistence", "--setting-sources", "", "--system-prompt", rubrica,
        "--output-format", "json", "--json-schema", esquema], input=mensaje, capture_output=True, text=True, timeout=120, cwd=str(destino))
    try:
        d = json.loads(r.stdout); return {"voto": n, "error": d.get("is_error"), "salida": d.get("structured_output")}
    except Exception as e:
        return {"voto": n, "error": True, "salida": None, "crudo": r.stdout[:300], "stderr": r.stderr[:300]}
def juzgar(f):
    nombre = f.name.removesuffix(".mensaje.txt"); m = f.read_text(encoding="utf-8")
    votos = [voto(m, 1)]
    s = votos[0]["salida"] or {}
    if (s.get("afirma_lo_no_leido") or {}).get("respuesta") == "si":
        votos += [voto(m, 2), voto(m, 3)]
    (destino / (nombre + ".votos.json")).write_text(json.dumps(votos, ensure_ascii=False, indent=1), encoding="utf-8")
    return nombre, votos
with cf.ThreadPoolExecutor(4) as ex:
    for nombre, votos in ex.map(juzgar, sorted(mensajes.glob("*.mensaje.txt"))):
        r = [((v["salida"] or {}).get("afirma_lo_no_leido") or {}).get("respuesta", "ERROR") for v in votos]
        print(nombre.replace("15-irpf-rendimientos-por-materia-", "").replace("claude-sonnet-5-5-", ""), "/".join(r), flush=True)
