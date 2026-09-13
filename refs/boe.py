#!/usr/bin/env python3
"""
boe.py — CLI para consultar la API de Datos Abiertos del BOE.

Uso:
    python boe.py buscar "impuesto sociedades"
    python boe.py buscar-materia 4107
    python boe.py indice BOE-A-2006-20764
    python boe.py articulo BOE-A-2006-20764 a17
    python boe.py articulos BOE-A-2006-20764 a28 a29 a30
    python boe.py metadatos BOE-A-2006-20764
    python boe.py analisis BOE-A-2006-20764
    python boe.py sumario 20260313
    python boe.py sumario 20260313 --fiscal
    python boe.py materias "tributario"

Todos los comandos imprimen texto plano legible a stdout.
Cache en disco (~/.cache/boe/) para evitar llamadas repetidas.
"""

import re
import sys
import json
import hashlib
import time
import urllib.request
import urllib.parse
import urllib.error
from pathlib import Path
from xml.etree import ElementTree as ET

from network_utils import robust_request

# ── Cache ────────────────────────────────────────────────────────────

CACHE_DIR = Path.home() / ".cache" / "boe"
CACHE_TTL = {
    "buscar": 300,
    "buscar-materia": 300,
    "indice": 604800,       # 7 días — estructura de normas raramente cambia
    "articulo": 604800,     # 7 días — textos consolidados son prácticamente inmutables
    "metadatos": 300,
    "analisis": 604800,     # 7 días — referencias cruzadas cambian con reformas
    "sumario": 86400,
    "materias": 86400,
}


def cache_get(cmd: str, *args) -> str | None:
    CACHE_DIR.mkdir(parents=True, exist_ok=True)
    key = hashlib.md5(f"{cmd}:{'|'.join(args)}".encode()).hexdigest()
    path = CACHE_DIR / f"{key}.json"
    if not path.exists():
        return None
    data = json.loads(path.read_text())
    if time.time() - data["ts"] > CACHE_TTL.get(cmd, 300):
        path.unlink()
        return None
    return data["val"]


def cache_set(cmd: str, val: str, *args):
    CACHE_DIR.mkdir(parents=True, exist_ok=True)
    key = hashlib.md5(f"{cmd}:{'|'.join(args)}".encode()).hexdigest()
    path = CACHE_DIR / f"{key}.json"
    path.write_text(json.dumps({"ts": time.time(), "val": val}))


# ── HTTP ─────────────────────────────────────────────────────────────

BOE_BASE = "https://www.boe.es"
BOE = f"{BOE_BASE}/datosabiertos/api/legislacion-consolidada"
BOE_SUMARIO = f"{BOE_BASE}/datosabiertos/api/boe/sumario"
BOE_AUX = f"{BOE_BASE}/datosabiertos/api/datos-auxiliares"


def get(url: str, accept: str = "application/json") -> str:
    """GET a la API del BOE. Retorna el body como string."""
    req = urllib.request.Request(url, headers={"Accept": accept})
    try:
        data = robust_request(req, timeout=30)
        return data.decode("utf-8")
    except urllib.error.HTTPError as e:
        return json.dumps({"error": f"HTTP {e.code}", "url": url})
    except Exception as e:
        return json.dumps({"error": str(e), "url": url})


def get_consolidada(path: str, accept: str = "application/json") -> str:
    return get(f"{BOE}{path}", accept)


# ── Parseo XML ───────────────────────────────────────────────────────

def xml_bloque_to_text(xml_str: str) -> dict:
    """Extrae texto limpio de la respuesta XML de un bloque del BOE."""
    try:
        root = ET.fromstring(xml_str)
    except Exception:
        return _fallback_parse(xml_str)

    bloque = root.find(".//bloque")
    if bloque is None:
        return {"titulo": "?", "texto": xml_str[:2000], "fecha": "?"}

    titulo = bloque.get("titulo", "")
    tipo = bloque.get("tipo", "")

    versiones = list(bloque.findall("version"))
    if not versiones:
        texto = _elem_all_text(bloque)
        return {"titulo": titulo, "tipo": tipo, "texto": texto, "fecha": "original"}

    # La última versión es la vigente
    ultima = versiones[-1]
    fecha = ultima.get("fecha_publicacion", "original")
    fecha_vigencia = ultima.get("fecha_vigencia", "")
    norma_modificadora = ultima.get("id_norma", "")

    texto = _elem_all_text(ultima)

    return {
        "titulo": titulo,
        "tipo": tipo,
        "fecha": fecha,
        "fecha_vigencia": fecha_vigencia,
        "norma_modificadora": norma_modificadora,
        "texto": texto,
    }


def _elem_all_text(elem) -> str:
    """Extrae todo el texto de un elemento XML recursivamente."""
    raw = ET.tostring(elem, encoding="unicode", method="text")
    lines = [line.strip() for line in raw.split("\n") if line.strip()]
    return "\n".join(lines)


def _fallback_parse(xml_str: str) -> dict:
    """Parseo de emergencia con regex cuando ET falla."""
    titulo = ""
    m = re.search(r'titulo="([^"]*)"', xml_str)
    if m:
        titulo = m.group(1)

    textos = re.findall(r"<p[^>]*>(.*?)</p>", xml_str, re.DOTALL)
    texto_limpio = "\n".join(
        re.sub(r"<[^>]+>", "", t).strip() for t in textos if t.strip()
    )
    return {"titulo": titulo, "texto": texto_limpio or xml_str[:2000], "fecha": "?"}


# ── Inferir tipo de bloque desde su ID ───────────────────────────────

def _tipo_from_id(bloque_id: str) -> str:
    """Infiere el tipo de bloque a partir de su ID."""
    bid = bloque_id.lower()
    if re.match(r"^a\d", bid):
        return "articulo"
    if bid.startswith("t"):
        return "titulo"
    if bid.startswith("ci") or bid.startswith("cv") or re.match(r"^c[ivxlcdm]", bid):
        return "capitulo"
    if bid.startswith("s") and (bid[1:2].isdigit() or bid[1:2] == "e"):
        return "seccion"
    if bid == "preambulo":
        return "preambulo"
    if bid.startswith("da"):
        return "disposicion_adicional"
    if bid.startswith("dt"):
        return "disposicion_transitoria"
    if bid.startswith("dd"):
        return "disposicion_derogatoria"
    if bid.startswith("df"):
        return "disposicion_final"
    return ""


# ── Verificación de vigencia ─────────────────────────────────────────

def _check_vigencia(norma_id: str) -> str | None:
    """Comprueba el estado de consolidación de una norma.

    Devuelve un aviso si la norma está desactualizada o derogada, None si todo OK.
    """
    raw = get_consolidada(f"/id/{norma_id}/metadatos")
    try:
        data = json.loads(raw).get("data", [])
    except (json.JSONDecodeError, AttributeError):
        return None

    if not data:
        return None

    item = data[0] if isinstance(data, list) and data else data

    avisos = []

    estado = item.get("estado_consolidacion", {})
    estado_cod = estado.get("codigo", "") if isinstance(estado, dict) else ""
    if estado_cod == "4":
        avisos.append(
            "⚠ TEXTO POSIBLEMENTE DESACTUALIZADO: la consolidación de esta norma "
            "no está finalizada. Puede haber modificaciones recientes aún no integradas."
        )

    if item.get("estatus_derogacion") == "S":
        avisos.append("⚠ NORMA DEROGADA: esta norma ha sido derogada.")

    if item.get("vigencia_agotada") == "S":
        avisos.append("⚠ VIGENCIA AGOTADA: esta norma ya no está en vigor.")

    return "\n".join(avisos) if avisos else None


# ── Normalizar listas de la API ──────────────────────────────────────

def _ensure_list(val) -> list:
    """La API del BOE a veces devuelve un objeto suelto en vez de un array
    de un solo elemento. Esta función normaliza a lista."""
    if val is None:
        return []
    if isinstance(val, list):
        return val
    return [val]


# ── Códigos de materia fiscal ────────────────────────────────────────

MATERIAS_FISCALES = {
    "6658", "4107", "4113", "4102", "4101", "4121", "4114", "4116",
    "4108", "4096", "4093", "4087", "7859", "4105", "4075", "4081",
    "3926", "4281", "5876", "5881", "2496", "4129", "2485", "5940",
    "5937", "826", "6924", "6839", "6841", "3985", "3986", "7878",
    "119", "2990", "4584", "5686",
}

# Palabras clave para detectar contenido fiscal en títulos del sumario
_FISCAL_KEYWORDS = re.compile(
    r"(?i)(tribut|fiscal|impuesto|impost|IVA|IRPF|contribu|recauda|"
    r"hacienda|catastro|arancel|aduane|liquidaci[oó]n.*tribut|"
    r"deuda\s+tributar|desgravaci|deducci[oó]n.*fiscal|"
    r"tipo\s+imposit|hecho\s+imponible|base\s+imponible|"
    r"exenci[oó]n.*tribut|sanci[oó]n.*tribut|autoliquidaci|"
    r"Agencia\s+(Estatal\s+de\s+)?Administraci[oó]n\s+Tributaria|"
    r"AEAT|Direcci[oó]n\s+General\s+de\s+Tributos)"
)


# ── Comandos ─────────────────────────────────────────────────────────

def cmd_buscar(texto: str, limit: int = 10) -> str:
    """Busca normas consolidadas por título."""
    cached = cache_get("buscar", texto)
    if cached:
        return cached

    # Si el texto ya contiene operadores Elasticsearch o prefijos de campo,
    # usarlo directamente. Si no, construir una query por título palabra a palabra.
    _ES_OPERATORS = (" AND ", " OR ", " NOT ", "titulo:", "materia:", '"')
    if any(op in texto for op in _ES_OPERATORS):
        query_str = texto
    else:
        palabras = texto.split()
        if len(palabras) > 1:
            query_str = " AND ".join(f"titulo:{p}" for p in palabras)
        else:
            query_str = f"titulo:{texto}"

    query = json.dumps({"query": {"query_string": {"query": query_str}}})
    path = f"?limit={limit}&query={urllib.parse.quote(query)}"
    raw = get_consolidada(path)

    try:
        parsed = json.loads(raw)
        if "error" in parsed:
            return f"Error al buscar en el BOE (API no disponible): {parsed['error']}. Usa otras fuentes como dgt_buscar."
        data = parsed.get("data", [])
    except (json.JSONDecodeError, AttributeError):
        return f"Error parseando respuesta: {raw[:500]}"

    if not data:
        return "No se encontraron normas."

    lines = []
    for item in data:
        if isinstance(item, dict):
            estado = item.get("estado_consolidacion", {})
            estado_txt = estado.get("texto", "?") if isinstance(estado, dict) else "?"
            rango = item.get("rango", {})
            rango_txt = rango.get("texto", "?") if isinstance(rango, dict) else "?"
            lines.append(
                f"  {item.get('identificador', '?')}\n"
                f"    {item.get('titulo', '?')}\n"
                f"    Rango: {rango_txt} | "
                f"Vigencia agotada: {item.get('vigencia_agotada', '?')} | "
                f"Estado: {estado_txt}\n"
                f"    URL: https://www.boe.es/buscar/act.php?id={item.get('identificador', '')}"
            )

    result = f"Resultados ({len(data)}):\n\n" + "\n\n".join(lines)
    cache_set("buscar", result, texto)
    return result


def cmd_buscar_materia(codigo_materia: str, limit: int = 10) -> str:
    """Busca normas consolidadas por código de materia (más preciso que por título)."""
    cached = cache_get("buscar-materia", codigo_materia)
    if cached:
        return cached

    query_str = f"materia@codigo:{codigo_materia}"
    query = json.dumps({"query": {"query_string": {"query": query_str}}})
    path = f"?limit={limit}&query={urllib.parse.quote(query)}"
    raw = get_consolidada(path)

    try:
        parsed = json.loads(raw)
        if "error" in parsed:
            return f"Error al buscar en el BOE (API no disponible): {parsed['error']}."
        data = parsed.get("data", [])
    except (json.JSONDecodeError, AttributeError):
        return f"Error parseando respuesta: {raw[:500]}"

    if not data:
        return f"No se encontraron normas para materia {codigo_materia}."

    lines = []
    for item in data:
        if isinstance(item, dict):
            estado = item.get("estado_consolidacion", {})
            estado_txt = estado.get("texto", "?") if isinstance(estado, dict) else "?"
            rango = item.get("rango", {})
            rango_txt = rango.get("texto", "?") if isinstance(rango, dict) else "?"
            lines.append(
                f"  {item.get('identificador', '?')}\n"
                f"    {item.get('titulo', '?')}\n"
                f"    Rango: {rango_txt} | "
                f"Vigencia agotada: {item.get('vigencia_agotada', '?')} | "
                f"Estado: {estado_txt}\n"
                f"    URL: https://www.boe.es/buscar/act.php?id={item.get('identificador', '')}"
            )

    result = f"Resultados materia {codigo_materia} ({len(data)}):\n\n" + "\n\n".join(lines)
    cache_set("buscar-materia", result, codigo_materia)
    return result


def cmd_indice(norma_id: str) -> str:
    """Obtiene el índice de bloques de una norma."""
    cached = cache_get("indice", norma_id)
    if cached:
        return cached

    raw = get_consolidada(f"/id/{norma_id}/texto/indice")

    try:
        data = json.loads(raw).get("data", [])
    except (json.JSONDecodeError, AttributeError):
        return f"Error parseando respuesta: {raw[:500]}"

    if not data:
        return f"No se encontró índice para {norma_id}."

    # La API devuelve data como array; los bloques están en data[0]["bloque"]
    bloques = data
    if isinstance(data, list) and len(data) > 0 and isinstance(data[0], dict):
        if "bloque" in data[0]:
            bloques = data[0]["bloque"]

    INDENT = {
        "titulo": 0,
        "capitulo": 1,
        "seccion": 2,
        "subseccion": 3,
        "articulo": 2,
        "preambulo": 0,
        "disposicion_adicional": 1,
        "disposicion_transitoria": 1,
        "disposicion_derogatoria": 1,
        "disposicion_final": 1,
    }

    lines = [f"Índice de {norma_id}:", f"URL: https://www.boe.es/buscar/act.php?id={norma_id}", ""]
    for item in bloques:
        if isinstance(item, dict):
            bloque_id = item.get("id", "?")
            titulo = item.get("titulo", "")
            tipo = _tipo_from_id(bloque_id)
            indent = "  " * INDENT.get(tipo, 0)
            label = titulo if titulo else bloque_id
            lines.append(f"{indent}[{bloque_id}] {label}")

    result = "\n".join(lines)
    cache_set("indice", result, norma_id)
    return result


def cmd_articulo(norma_id: str, id_bloque: str) -> str:
    """Obtiene el texto vigente de un artículo, con verificación de vigencia."""
    cached = cache_get("articulo", norma_id, id_bloque)
    if cached:
        return cached

    raw = get_consolidada(
        f"/id/{norma_id}/texto/bloque/{id_bloque}", accept="application/xml"
    )
    parsed = xml_bloque_to_text(raw)

    # Verificar estado de consolidación
    aviso_vigencia = _check_vigencia(norma_id)

    lines = []
    if aviso_vigencia:
        lines.append(aviso_vigencia)
        lines.append("")

    lines.append(f"{'═' * 60}")
    lines.append(f"{parsed.get('titulo', id_bloque)}")
    lines.append(f"Versión vigente: {parsed.get('fecha', '?')}")

    if parsed.get("norma_modificadora"):
        lines.append(f"Modificado por: {parsed['norma_modificadora']}")
    if parsed.get("fecha_vigencia"):
        lines.append(f"En vigor desde: {parsed['fecha_vigencia']}")
    lines.append(f"{'═' * 60}")
    lines.append("")
    lines.append(parsed.get("texto", "(sin texto)"))
    lines.append("")
    lines.append(f"URL: https://www.boe.es/buscar/act.php?id={norma_id}#{id_bloque}")

    result = "\n".join(lines)
    cache_set("articulo", result, norma_id, id_bloque)
    return result


def cmd_articulos(norma_id: str, ids_bloque: list[str]) -> str:
    """Obtiene varios artículos de la misma norma."""
    parts = []
    for bid in ids_bloque:
        parts.append(cmd_articulo(norma_id, bid))
        parts.append("")
    return "\n".join(parts)


def cmd_metadatos(norma_id: str) -> str:
    """Obtiene metadatos de una norma."""
    cached = cache_get("metadatos", norma_id)
    if cached:
        return cached

    raw = get_consolidada(f"/id/{norma_id}/metadatos")

    try:
        data = json.loads(raw).get("data", [])
    except (json.JSONDecodeError, AttributeError):
        return f"Error: {raw[:500]}"

    if not data:
        return f"No se encontraron metadatos para {norma_id}."

    # data es un array, tomar el primer elemento
    item = data[0] if isinstance(data, list) and data else data

    estado = item.get("estado_consolidacion", {})
    estado_txt = estado.get("texto", "?") if isinstance(estado, dict) else str(estado)
    estado_cod = estado.get("codigo", "") if isinstance(estado, dict) else ""
    rango = item.get("rango", {})
    rango_txt = rango.get("texto", "?") if isinstance(rango, dict) else str(rango)

    lines = [
        f"Metadatos de {norma_id}:",
        f"  Título:        {item.get('titulo', '?')}",
        f"  Rango:         {rango_txt}",
        f"  Nº oficial:    {item.get('numero_oficial', '?')}",
        f"  Fecha disp.:   {item.get('fecha_disposicion', '?')}",
        f"  Publicación:   {item.get('fecha_publicacion', '?')}",
        f"  Vigencia:      {item.get('fecha_vigencia', '?')}",
        f"  Derogada:      {item.get('estatus_derogacion', '?')}",
        f"  Vig. agotada:  {item.get('vigencia_agotada', '?')}",
        f"  Consolidación: {estado_txt}",
        f"  URL ELI:       {item.get('url_eli', '?')}",
    ]

    if estado_cod == "4":
        lines.append("")
        lines.append(
            "  ⚠ TEXTO POSIBLEMENTE DESACTUALIZADO: la consolidación no está "
            "finalizada. Puede haber modificaciones recientes aún no integradas."
        )

    if item.get("estatus_derogacion") == "S":
        lines.append("")
        lines.append("  ⚠ NORMA DEROGADA")

    if item.get("vigencia_agotada") == "S":
        lines.append("")
        lines.append("  ⚠ VIGENCIA AGOTADA")

    result = "\n".join(lines)
    cache_set("metadatos", result, norma_id)
    return result


def cmd_analisis(norma_id: str) -> str:
    """Obtiene análisis jurídico (materias, referencias cruzadas)."""
    cached = cache_get("analisis", norma_id)
    if cached:
        return cached

    raw = get_consolidada(f"/id/{norma_id}/analisis")

    try:
        data = json.loads(raw).get("data", [])
    except (json.JSONDecodeError, AttributeError):
        return f"Error: {raw[:500]}"

    if not data:
        return f"No se encontró análisis para {norma_id}."

    # data es un array, tomar el primer elemento
    item = data[0] if isinstance(data, list) and data else data

    lines = [f"Análisis de {norma_id}:", ""]

    # Materias
    materias = item.get("materias", [])
    if materias:
        lines.append("Materias:")
        for m in materias:
            mat = m.get("materia", m) if isinstance(m, dict) else m
            if isinstance(mat, dict):
                lines.append(f"  - {mat.get('texto', '?')} ({mat.get('codigo', '?')})")
            else:
                lines.append(f"  - {mat}")

    # Notas
    notas = item.get("notas", {})
    if notas:
        nota_txt = notas.get("nota", notas) if isinstance(notas, dict) else notas
        lines.append(f"\nNotas: {nota_txt}")

    # Referencias
    refs = item.get("referencias", {})
    if isinstance(refs, dict):
        anteriores = refs.get("anteriores", [])
        if anteriores:
            lines.append("\nReferencias anteriores (normas que esta modifica/deroga):")
            for ref_group in anteriores:
                ref_list = (
                    ref_group.get("anterior", [ref_group])
                    if isinstance(ref_group, dict)
                    else [ref_group]
                )
                for r in ref_list:
                    if isinstance(r, dict):
                        rel = r.get("relacion", {})
                        rel_txt = (
                            rel.get("texto", "?") if isinstance(rel, dict) else str(rel)
                        )
                        lines.append(f"  {rel_txt} → {r.get('id_norma', '?')}")
                        texto_ref = r.get("texto", "")
                        if texto_ref:
                            lines.append(f"    {texto_ref[:200]}")

        posteriores = refs.get("posteriores", [])
        if posteriores:
            lines.append("\nReferencias posteriores (normas que la modifican):")
            for ref_group in posteriores:
                ref_list = (
                    ref_group.get("posterior", [ref_group])
                    if isinstance(ref_group, dict)
                    else [ref_group]
                )
                for r in ref_list:
                    if isinstance(r, dict):
                        rel = r.get("relacion", {})
                        rel_txt = (
                            rel.get("texto", "?") if isinstance(rel, dict) else str(rel)
                        )
                        lines.append(f"  {rel_txt} → {r.get('id_norma', '?')}")

    result = "\n".join(lines)
    cache_set("analisis", result, norma_id)
    return result


def cmd_sumario(fecha: str, solo_fiscal: bool = False) -> str:
    """Obtiene el sumario del BOE de una fecha (YYYYMMDD).

    Si solo_fiscal=True, filtra solo las entradas relacionadas con fiscalidad.
    """
    cache_key = f"{fecha}:{'fiscal' if solo_fiscal else 'todo'}"
    cached = cache_get("sumario", cache_key)
    if cached:
        return cached

    raw = get(f"{BOE_SUMARIO}/{fecha}")

    try:
        resp = json.loads(raw)
    except (json.JSONDecodeError, AttributeError):
        if "404" in raw[:200]:
            return f"No hay BOE publicado para la fecha {fecha} (¿domingo o festivo?)."
        if "400" in raw[:200]:
            return f"Fecha no válida: {fecha}. Formato esperado: YYYYMMDD."
        return f"Error: {raw[:500]}"

    if "error" in resp:
        error_msg = resp.get("error", "")
        if "404" in error_msg:
            return f"No hay BOE publicado para la fecha {fecha} (¿domingo o festivo?)."
        return f"Error consultando sumario: {error_msg}"

    sumario = resp.get("data", {}).get("sumario", {})
    if not sumario:
        return f"No se encontró sumario para {fecha}."

    meta = sumario.get("metadatos", {})
    diarios = _ensure_list(sumario.get("diario", []))

    lines = [
        f"Sumario del BOE — {meta.get('fecha_publicacion', fecha)}",
    ]

    if solo_fiscal:
        lines[0] += " (solo contenido fiscal)"

    for diario in diarios:
        num = diario.get("numero", "?")
        lines.append(f"Número: {num}")
        lines.append("")

        secciones = _ensure_list(diario.get("seccion", []))
        for seccion in secciones:
            codigo_sec = seccion.get("codigo", "?")
            nombre_sec = seccion.get("nombre", "?")

            items_seccion = []
            departamentos = _ensure_list(seccion.get("departamento", []))

            for depto in departamentos:
                nombre_depto = depto.get("nombre", "?")

                # Secciones 1-4: departamento → epigrafe → item
                # Secciones 5A/5B: departamento → item (sin epigrafe)
                items_depto = []

                epigrafes = _ensure_list(depto.get("epigrafe", []))
                for ep in epigrafes:
                    items_raw = _ensure_list(ep.get("item", []))
                    for it in items_raw:
                        if isinstance(it, dict):
                            items_depto.append(it)

                # Items directos en departamento (sección 5)
                items_directos = _ensure_list(depto.get("item", []))
                for it in items_directos:
                    if isinstance(it, dict):
                        items_depto.append(it)

                if solo_fiscal:
                    items_depto = [
                        it for it in items_depto
                        if _FISCAL_KEYWORDS.search(it.get("titulo", ""))
                    ]

                if items_depto:
                    items_seccion.append((nombre_depto, items_depto))

            if items_seccion:
                lines.append(f"{'─' * 50}")
                lines.append(f"  {nombre_sec}")
                lines.append(f"{'─' * 50}")

                for nombre_depto, items in items_seccion:
                    lines.append(f"  {nombre_depto}")
                    for it in items:
                        ident = it.get("identificador", "?")
                        titulo = it.get("titulo", "?")
                        lines.append(f"    [{ident}] {titulo}")
                    lines.append("")

    if solo_fiscal and len(lines) <= 3:
        lines.append("No se encontraron entradas fiscales en este BOE.")

    result = "\n".join(lines)
    cache_set("sumario", result, cache_key)
    return result


def cmd_materias(filtro: str = "") -> str:
    """Lista materias disponibles en el BOE, opcionalmente filtradas por texto."""
    cached = cache_get("materias", filtro)
    if cached:
        return cached

    raw = get(f"{BOE_AUX}/materias")

    try:
        data = json.loads(raw).get("data", {})
    except (json.JSONDecodeError, AttributeError):
        return f"Error: {raw[:500]}"

    if not data:
        return "No se encontraron materias."

    if filtro:
        filtro_lower = filtro.lower()
        items = [
            (cod, txt) for cod, txt in data.items()
            if filtro_lower in txt.lower()
        ]
    else:
        # Sin filtro, mostrar solo las fiscales conocidas
        items = [
            (cod, txt) for cod, txt in data.items()
            if cod in MATERIAS_FISCALES
        ]

    items.sort(key=lambda x: x[1])

    if not items:
        return f"No se encontraron materias que contengan '{filtro}'."

    lines = [f"Materias ({len(items)} resultados):", ""]
    for cod, txt in items:
        lines.append(f"  [{cod}] {txt}")

    result = "\n".join(lines)
    cache_set("materias", result, filtro)
    return result


# ── Main ─────────────────────────────────────────────────────────────

USAGE = """
Uso: python boe.py <comando> <args...>

Comandos — Legislación consolidada:
  buscar <texto>                        Buscar normas por título
  buscar-materia <codigo>               Buscar normas por código de materia
  indice <norma_id>                     Tabla de contenidos de una norma
  articulo <norma_id> <id_bloque>       Texto vigente de un artículo
  articulos <norma_id> <id1> <id2>...   Varios artículos a la vez
  metadatos <norma_id>                  Vigencia, estado, fechas
  analisis <norma_id>                   Referencias cruzadas, materias

Comandos — Sumario diario:
  sumario <YYYYMMDD>                    Sumario completo del BOE de esa fecha
  sumario <YYYYMMDD> --fiscal           Solo entradas fiscales/tributarias

Comandos — Datos auxiliares:
  materias                              Lista materias fiscales conocidas
  materias <texto>                      Buscar materias por nombre

Ejemplos:
  python boe.py buscar "impuesto sociedades"
  python boe.py buscar-materia 4107
  python boe.py indice BOE-A-2006-20764
  python boe.py articulo BOE-A-2006-20764 a17
  python boe.py articulos BOE-A-2006-20764 a28 a29 a30
  python boe.py metadatos BOE-A-2006-20764
  python boe.py sumario 20260313 --fiscal
  python boe.py materias "tributario"

Normas frecuentes:
  IRPF:             BOE-A-2006-20764
  Sociedades:       BOE-A-2014-12328
  IVA:              BOE-A-1992-28740
  LGT:              BOE-A-2003-23186
  Haciendas Locales: BOE-A-2004-4214
""".strip()


def main():
    if len(sys.argv) < 2:
        print(USAGE)
        sys.exit(1)

    cmd = sys.argv[1]

    if cmd == "buscar" and len(sys.argv) >= 3:
        print(cmd_buscar(" ".join(sys.argv[2:])))
    elif cmd == "buscar-materia" and len(sys.argv) >= 3:
        print(cmd_buscar_materia(sys.argv[2]))
    elif cmd == "indice" and len(sys.argv) >= 3:
        print(cmd_indice(sys.argv[2]))
    elif cmd == "articulo" and len(sys.argv) >= 4:
        print(cmd_articulo(sys.argv[2], sys.argv[3]))
    elif cmd == "articulos" and len(sys.argv) >= 4:
        print(cmd_articulos(sys.argv[2], sys.argv[3:]))
    elif cmd == "metadatos" and len(sys.argv) >= 3:
        print(cmd_metadatos(sys.argv[2]))
    elif cmd == "analisis" and len(sys.argv) >= 3:
        print(cmd_analisis(sys.argv[2]))
    elif cmd == "sumario" and len(sys.argv) >= 3:
        fiscal = "--fiscal" in sys.argv
        fecha = sys.argv[2]
        print(cmd_sumario(fecha, solo_fiscal=fiscal))
    elif cmd == "materias":
        filtro = " ".join(sys.argv[2:]) if len(sys.argv) >= 3 else ""
        print(cmd_materias(filtro))
    else:
        print(USAGE)
        sys.exit(1)


if __name__ == "__main__":
    main()
