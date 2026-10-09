#!/usr/bin/env python3
"""Construye los casos de la validación del juez de jurisprudencia (ADR 0037).

    casos.py <repositorio> <binario kitlegal> <carpeta de la evidencia> <salida.jsonl>

Sin modelo y sin nada escrito a mano. Tres orígenes, cada uno con su informe:
  - informe-h23:      las respuestas del informe del cierre de H23. La pregunta se lee de la eval en el commit del
                      informe, y los textos de las herramientas se reconstruyen repitiendo las órdenes de la sesión
                      con el binario, que en el applet cita no pide nada a la red. En el modo herramienta, cotejar se
                      repite con el documento que la orden lleva, que es la ficha. En el modo orden, lo que leyó por
                      la entrada estándar no está en el informe: se reconstruye con el texto pegado en la pregunta.
  - sondeo-con-skill: las respuestas de sondeo-con-skill.json. La pregunta la compone preguntas.py, y los textos son
                      los que devolvió cada orden de kitlegal en la sesión, tal como están en el informe del sondeo.
  - sondeo-sin-skill: las de sondeo-sin-skill.json, que no tienen herramientas ni textos.

Dos clases de caso:
  - leido:    la respuesta tal como está. Su etiqueta es la de etiquetas.json, la lectura del 2026-10-09.
  - derivado: la misma respuesta con una parte del texto pegado quitada de la pregunta. `quitado` dice cuál:
              `documento` (todo el texto, y con él las órdenes cotejar), `fallo` (desde «F A L L O»; queda la ficha)
              o `apartado-2` (solo el 2.º del fallo). Solo se deriva de una respuesta leída como correcta. Su
              etiqueta sale de una cuenta sin modelo: `palabras` son las de la respuesta que están en lo quitado y no
              en lo que queda delante.
                · con tres o más que solo están en el fallo, la respuesta cuenta el fallo que ya no tiene: defecto;
                · sin el documento, con el nombre del ponente, da datos de una ficha que ya no tiene: defecto;
                · sin el fallo o sin su apartado 2.º, con menos de tres, solo daba la ficha: correcto;
                · lo demás no se deriva: no hay regla que lo etiquete sin leerlo.
"""
import collections, json, os, re, subprocess, sys, tempfile, unicodedata

sys.path.insert(0, os.path.dirname(os.path.abspath(__file__)))
from preguntas import COMMIT_DEL_INFORME, MARCA_DEL_FALLO, pregunta_de, preguntas  # noqa: E402

INFORME_H23 = "specs/019-h23-skill-jurisprudencia-ninguna/gates/evals/jurisprudencia.json"
MINIMO_DEL_FALLO = 3


def palabras(texto):
    plano = unicodedata.normalize("NFD", texto.lower())
    return set(re.findall(r"[a-zñ]{5,}", "".join(c for c in plano if unicodedata.category(c) != "Mn")))


def partir(pregunta):
    """(entrada, documento) de una pregunta con texto pegado; documento vacío si no lo lleva."""
    if "\n\nRoj:" not in pregunta:
        return pregunta, ""
    entrada, documento = pregunta.split("\n\n", 1)
    return entrada, documento


def recortes(documento):
    """Lo que queda del documento con cada cosa quitada."""
    quedan = {"documento": ""}
    if MARCA_DEL_FALLO in documento:
        corte = documento.index(MARCA_DEL_FALLO)
        apartado = re.search(r"^2\.º-.*\n\n", documento, re.M).group(0)
        quedan["fallo"] = documento[:corte].rstrip("\n") + "\n"
        quedan["apartado-2"] = documento.replace(apartado, "")
    return quedan


def argumentos(orden):
    """(verbo, {bandera: valor}, referencia) de una orden del informe del job, en sus dos modos."""
    orden = orden.replace("cita_", "cita ", 1)
    cabeza, *resto = re.split(r"(?:^|\s)--(?=[a-z])", orden)
    banderas = {}
    for trozo in resto:
        nombre, _, valor = re.match(r"([a-z]+)(=|\s|$)(.*)", trozo, re.S).groups()
        banderas[nombre] = valor.strip() if nombre != "documento" else valor
    return cabeza.split()[1], banderas, cabeza.split()[2:]


def repetir(binario, casa, orden, documento):
    """La salida de una orden del informe del job, repetida. None si no es una orden del applet cita."""
    if not orden.startswith("cita"):
        return None
    verbo, banderas, referencia = argumentos(orden)
    llamada = [binario, "cita", verbo, *referencia]
    for nombre, valor in banderas.items():
        if nombre != "json":
            llamada += [f"--{nombre}", valor]
    entrada = documento if verbo == "cotejar" and "documento" not in banderas else None
    hecho = subprocess.run(llamada + ["--json", "--no-graph"], input=entrada, capture_output=True, text=True,
                           env={"HOME": casa, "KITLEGAL_CACHE_DIR": os.path.join(casa, "cache"), "PATH": os.environ["PATH"]})
    return {"verbo": verbo, "codigo": hecho.returncode, "salida": hecho.stdout or hecho.stderr}


def textos_del_job(binario, casa, sesion, documento):
    textos = []
    for invocacion in sesion["invocaciones"]:
        salida = repetir(binario, casa, invocacion["orden"], documento)
        if salida is None or (salida["verbo"] == "cotejar" and not documento):
            continue  # sin documento no hay nada que cotejar
        if salida["codigo"] != invocacion["codigo"]:
            raise SystemExit(f"{sesion['sesion']}: {invocacion['orden'][:60]} da {salida['codigo']} y el informe dice {invocacion['codigo']}")
        textos.append({"orden": invocacion["orden"], "codigo": salida["codigo"], "salida": salida["salida"]})
    return textos


def textos_del_sondeo(sesion, documento):
    textos = []
    for invocacion in sesion["invocaciones"]:
        orden = (invocacion["entrada"].get("command") or "") if invocacion["herramienta"] == "Bash" else ""
        if not orden.startswith("kitlegal ") or ("cita cotejar" in orden and not documento):
            continue
        textos.append({"orden": orden, "codigo": 1 if invocacion.get("error") else 0, "salida": invocacion.get("salida", "")})
    return textos


def leidas(repo, binario, evidencia, casa):
    """(base del caso, pregunta, función que da los textos con lo que quede del documento) de cada respuesta."""
    informe = json.load(open(os.path.join(repo, INFORME_H23), encoding="utf-8"))
    if informe["commit"] != COMMIT_DEL_INFORME:
        raise SystemExit("el informe de H23 no es el del commit de las preguntas")
    for sesion in informe["evals"]:
        base = {"origen": "informe-h23", "informe": INFORME_H23, "sesion": sesion["sesion"], "modelo": sesion["modelo"],
                "modo": sesion["modo"] or "orden", "respuesta": sesion["respuesta"]}
        yield base, pregunta_de(repo, informe["commit"], sesion["eval"]), \
            (lambda documento, sesion=sesion: textos_del_job(binario, casa, sesion, documento))
    for nombre in ("sondeo-con-skill", "sondeo-sin-skill"):
        sondeo = json.load(open(os.path.join(evidencia, nombre + ".json"), encoding="utf-8"))
        textos = {p["id"]: p["pregunta"] for p in preguntas(repo)}
        for sesion in sondeo["sesiones"]:
            base = {"origen": nombre, "informe": nombre + ".json", "sesion": sesion["sesion"], "modelo": sesion["modelo"],
                    "modo": "orden" if nombre == "sondeo-con-skill" else "sin-herramientas", "respuesta": sesion["respuesta"]}
            yield base, textos[sesion["pregunta"]], (lambda documento, sesion=sesion: textos_del_sondeo(sesion, documento))


def main():
    repo, binario, evidencia, destino = sys.argv[1:5]
    etiquetas = json.load(open(os.path.join(evidencia, "etiquetas.json"), encoding="utf-8"))
    casa = tempfile.mkdtemp(prefix="kitlegal-casos-")
    casos, vistas = [], set()
    for base, pregunta, textos_con in leidas(repo, binario, evidencia, casa):
        nombre = base["sesion"]
        vistas.add(nombre)
        if nombre in etiquetas["excluida"]:
            continue
        if not base["respuesta"].strip():
            raise SystemExit(f"{nombre}: sin respuesta y sin motivo en etiquetas.json")
        entrada, documento = partir(pregunta)
        etiqueta = "defecto" if nombre in etiquetas["defecto"] else "correcto"
        caso = {**base, "id": nombre, "clase": "leido", "etiqueta": etiqueta, "procedencia": "lectura",
                "pregunta": pregunta, "textos": textos_con(documento)}
        if nombre in etiquetas["frontera"]:
            if etiquetas["frontera"][nombre] not in base["respuesta"]:
                raise SystemExit(f"{nombre}: la frase de la frontera no está en la respuesta")
            caso["frase"] = etiquetas["frontera"][nombre]
        casos.append(caso)
        if not documento or etiqueta != "correcto":
            continue
        solo_del_fallo = set()
        if MARCA_DEL_FALLO in documento:
            corte = documento.index(MARCA_DEL_FALLO)
            solo_del_fallo = palabras(documento[corte:]) - palabras(documento[:corte])
        for quitado, queda in recortes(documento).items():
            nueva = entrada + ("\n\n" + queda if queda else "")
            textos = textos_con(queda)
            delante = palabras(nueva) | palabras(" ".join(t["orden"] + " " + t["salida"] for t in textos))
            suyas = sorted((palabras(base["respuesta"]) & palabras(documento)) - delante)
            del_fallo = sorted(set(suyas) & solo_del_fallo)
            if len(del_fallo) >= MINIMO_DEL_FALLO:
                derivada, regla = "defecto", "cuenta-el-fallo"
            elif quitado == "documento" and {"torres"} <= set(suyas):
                derivada, regla = "defecto", "da-la-ficha"
            elif quitado != "documento":
                derivada, regla = "correcto", "solo-daba-la-ficha"
            else:
                continue
            casos.append({**base, "id": f"{nombre}/sin-{quitado}", "clase": "derivado", "quitado": quitado,
                          "etiqueta": derivada, "procedencia": "derivado", "regla": regla, "pregunta": nueva,
                          "textos": textos, "palabras": suyas})
    sobran = (set(etiquetas["defecto"]) | set(etiquetas["excluida"]) | set(etiquetas["frontera"])) - vistas
    if sobran:
        raise SystemExit(f"etiquetas de sesiones que no existen: {sorted(sobran)}")
    with open(destino, "w", encoding="utf-8") as salida:
        for caso in casos:
            salida.write(json.dumps(caso, ensure_ascii=False) + "\n")
    cuenta = collections.Counter()
    for caso in casos:
        cuenta[(caso["etiqueta"], caso["origen"], caso["clase"] + (" sin-" + caso["quitado"] + " " + caso["regla"] if caso.get("quitado") else ""))] += 1
    for clave in sorted(cuenta):
        print(f"{cuenta[clave]:4}  " + " · ".join(clave))
    defectos = sum(1 for c in casos if c["etiqueta"] == "defecto")
    print(f"total {len(casos)}: {defectos} defectos y {len(casos) - defectos} correctos; "
          f"votos de una pasada sin fallos: {3 * defectos + len(casos) - defectos}")


if __name__ == "__main__":
    main()
