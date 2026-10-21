---
id: H2
fase: "0"
---

# H2 · `internal/httpx`: cliente HTTP responsable + grabación de fixtures

- **Objetivo**: toda la red pasa por aquí. Sin esto no se escribe ningún adaptador.
- **Entrega**: `httpx.New(opts...)` con UA `kitlegal/x.y (+https://ventanillalegal.es/bot)`, timeout por contexto, retries con backoff exponencial y jitter, rate limit por host (`x/time/rate`), `robots.txt` cacheado, y modo `KITLEGAL_RECORD=1` que graba petición/respuesta en `testdata/<fuente>/<nombre>.json`; en tests, `httpx.Replay(dir)`.
- **Alcance**: cadena de decoradores sobre `http.RoundTripper`; errores tipados (`ErrSourceDown` en 5xx/timeout, `ErrRateLimited` en 429). Sin paralelismo por defecto.
- **Controles**: unit con `httptest` (retries, 429, robots deny, UA presente, `--dry-run`); `-race`; `noctx` y `bodyclose` en lint.
- **Aceptación**: un adaptador de prueba no puede hacer una petición sin contexto ni sin UA.
