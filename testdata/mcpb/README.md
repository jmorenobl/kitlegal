# Esquema del manifiesto de MCP Bundle

`mcpb-manifest-v0.3.schema.json` es el esquema JSON oficial de la versión `0.3` del manifiesto de un MCP Bundle
(`.mcpb`), copiado byte a byte de su fuente. Contra él se valida el manifiesto de `kitlegal.mcpb` (ADR 0035; H22,
FR-017): el que escribe el paso, en `make ci` (`internal/empaquetado`, `TestPiezas`), y el del snapshot, en
`make snapshot-check` (`TestSnapshot`, `manifiesto-de-la-extension`).

| | |
| --- | --- |
| Fuente | `github.com/modelcontextprotocol/mcpb`, `schemas/mcpb-manifest-v0.3.schema.json` |
| Commit | `70fe3b34cd6dff1b3bba046638edc72a6467a4fb` (2026-04-22); el fichero no cambia desde `257af30` (2025-10-30) |
| Copiado | 2026-10-02 |
| sha256 | `3a0ac9d845711a1b9b17dfa5a52f8b60628239d6a86a9db417206a9efc78592d` |
| Licencia | la del proyecto MCP: Apache-2.0, o MIT en las contribuciones anteriores a su cambio de licencia |

No es una fuente del producto: el binario no lo pide ni lo lleva, y por eso no tiene fila en `docs/SOURCES.md`. No se
edita a mano. Para actualizarlo se sustituye el fichero entero por el de la fuente y se cambian el commit y la huella
aquí y en `internal/empaquetado/esquema_test.go`, que falla si el fichero no es el de su huella. Una versión nueva del
manifiesto es otro fichero, y cambiar la que declara `kitlegal.mcpb` es una decisión de su hito.
