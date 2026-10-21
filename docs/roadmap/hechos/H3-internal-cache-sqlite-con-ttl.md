---
id: H3
fase: "0"
---

# H3 · `internal/cache`: SQLite con TTL y `--offline`

- **Objetivo**: las normas cambian poco; no volver a pedir lo que ya tenemos.
- **Entrega**: `cache.Get/Put(key, ttl)` en `~/.cache/kitlegal/cache.db` (ruta configurable con `KITLEGAL_CACHE_DIR`); con `--offline` solo lee, y devuelve exit 4 si falta.
- **Alcance**: `modernc.org/sqlite`, migraciones embebidas (`embed` + tabla `schema_version`), WAL, `PRAGMA` seguros. Interfaz `Cache` definida en `core`.
- **Controles**: integración (`//go:build integration`) con `t.TempDir()`; unit de expiración de TTL con reloj inyectado; `sqlclosecheck`.
- **Aceptación**: segunda llamada idéntica no toca la red (se verifica con `Replay` en modo estricto que falla ante cualquier petición).
