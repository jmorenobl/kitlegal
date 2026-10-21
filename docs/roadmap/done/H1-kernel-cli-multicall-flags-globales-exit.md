---
id: H1
fase: "0"
---

# H1 · Kernel CLI: multicall, flags globales, exit codes, sobre de salida

- **Objetivo**: el patrón que todos los applets repetirán, escrito una vez.
- **Entrega**: `kitlegal echo hola --json` devuelve el sobre `{ok, fuente, url, fecha_consulta, hash, data}`; `ln -s kitlegal echo && ./echo hola` funciona; `--describe` emite JSON Schema del applet.
- **Alcance**: `internal/cli` (Kong, flags `--json --timeout --offline --dry-run --describe --no-graph --asunto --verbose`, `errors.go` con errores tipados → exit codes, `slog` a stderr), `internal/app` (registro de applets, dispatch por `os.Args[0]` o primer argumento), `internal/core/schema` (tipos del sobre, `hash` sha256 del `data` canónico), `internal/render` (json y tabla mínima). El applet `echo` es de ejemplo: vive en `internal/app/ejemplo`, que el binario distribuido no enlaza (ADR 0010).
- **Controles**: unit (mapeo error→exit, sobre, dispatch, `--describe`); e2e testscript (`--help`, symlink, exit 2 con args malos, salida JSON parseable); test de arquitectura; `depguard` activo.
- **Aceptación**: cobertura `internal/cli` ≥ 90 %; el e2e demuestra que stdout solo contiene JSON cuando `--json`.
