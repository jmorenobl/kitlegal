---
id: H4
fase: "0"
---

# H4 · Applet `boe`: puerto de `boe.py` — PRIMER HITO

- **Objetivo**: `kitlegal boe articulo BOE-A-2015-10565 a21` con el mismo comportamiento que `boe.py`.
- **Entrega**: verbos `buscar`, `indice`, `articulo`, `articulos`, `metadatos`, `analisis` contra la API de Legislación Consolidada, con caché y sobre de salida. `fuente: "boe.legislacion-consolidada"`.
- **Alcance**: `internal/source/boe` implementando `Source`; cliente, tipos, mapeo a `data`; `schemas/norma.json` y `bloque.json` generados desde `--describe` y vigilados por `make schema-check`. Antes de portar, leer **`refs/boe.py`** (copia congelada del 2026-05-14 de la skill `boe-fiscal`, el único fuente de referencia del porte) línea a línea y anotar comportamientos no obvios en `internal/source/boe/doc.go`. Su dependencia `network_utils.robust_request` no se porta: reintentos, ritmo y `robots.txt` ya son de `internal/httpx`; los TTL por verbo de `boe.py` sí se portan como `TTL` del `Source`. Los ids naturales que el applet observa (`BOE-A-…`, ELI, id de bloque, hash del texto) se conservan en los tipos: en H7 se emiten al grafo (ADR 0014).
- **Controles**: fixtures grabados para cada verbo; golden files de la salida JSON; e2e testscript de los seis verbos en modo replay; `scripts/verify-sources.sh` con el caso `boe articulo` (nightly); fuzz ligero sobre el parser de ids de bloque (`a21`, `da3`, `dt1`).
- **Aceptación**: diff entre la salida de `boe.py articulo` y `kitlegal boe articulo` (campo `data`) vacío para 5 artículos de 3 leyes; tiempo < 200 ms en caché.
