# Contrato: reglas de arquitectura en H2

Estado de las cinco reglas de `docs/ROADMAP.md` §2 al cerrar H2, y los dos controles de lint que este
hito estrena. Continúa `specs/002-h1-kernel-cli-multicall/contracts/reglas-de-arquitectura.md`.

## 1. Las cinco reglas

| # | Regla | Estado en H1 | Estado en H2 |
|---|---|---|---|
| R1 | `internal/core/**` no importa `internal/{source,httpx,cache,store,graph,render,cli,app}` ni I/O de la biblioteca estándar | Activa | **Sin cambio**: `internal/core/schema` gana una interfaz (`ConClase`) y un campo (`Resultado.Ensayo`) sin importar nada nuevo. `internal/httpx` importa `schema` (por `Contexto` y `Clase`), nunca al revés (FR-055) |
| R2 | Solo `internal/httpx` importa `net/http` | Activa y **vacía** (nadie lo importaba) | **Activa y con dueño**: `internal/httpx` es el único paquete del módulo que importa `net/http`, `net/http/httptest` (en sus tests) y `golang.org/x/time/rate`, `github.com/temoto/robotstxt`. El test de arquitectura exige que el dueño exista (FR-053) |
| R3 | Solo `internal/{cache,store,graph}` importan SQLite y `database/sql` | Activa y vacía | Sin cambio (H3) |
| R4 | Solo `internal/cli` y `cmd/` llaman a `os.Exit` | Activa, forma estricta | Sin cambio: `internal/httpx` no llama a `os.Exit` |
| R5 | Solo `internal/render` escribe en stdout; logs a stderr | Activa | Sin cambio: `internal/httpx` escribe solo por el `*slog.Logger` recibido (FR-035); la descripción de `--dry-run` la presenta el kernel (contrato de errores §5) |

## 2. Cómo se comprueba cada una en H2

| Regla | `depguard` | `forbidigo` | `internal/arch_test.go` | Cambio en H2 |
|---|---|---|---|---|
| R1 | lista `core` | — | `compruebaDominioPuro` | ninguno |
| R2 | lista `red` (`**` salvo `!**/internal/httpx/**`) | — | subprueba «R2»; **gana** `require.Contains(g.importa, modulo+"/internal/httpx")` para no pasar en vacío | comentario de `.golangci.yml` («el paquete todavía no existe») actualizado; test endurecido |
| R3 | lista `sql` | — | subprueba «R3» | ninguno |
| R4 | — | `^os\.Exit$` | — | ninguno |
| R5 | — | `^fmt\.Print(\|f\|ln)$`, `^os\.Stdout$`, `^os\.Stderr$` | — | ninguno |

Ninguna exclusión nueva en `.golangci.yml` (SC-010).

## 3. Los dos analizadores que H2 estrena

Activos desde H0; por primera vez tienen código que vigilar. Verificado en su código
(research.md, tabla de verificación):

| Analizador | Qué prohíbe | Cómo lo cumple `internal/httpx` |
|---|---|---|
| `noctx` v0.5.1 | `net/http.Get/Head/Post/PostForm`, `(*net/http.Client).Get/Head/Post/PostForm`, `net/http.NewRequest`, `net/http/httptest.NewRequest` | Toda petición se crea con `http.NewRequestWithContext(ctx, …)` y se emite con `(*http.Client).Do` o directamente por el `RoundTripper` |
| `bodyclose` | Un `*http.Response` cuyo `Body` no se cierra | El único `*http.Response` del módulo se lee entero y se cierra dentro de `internal/httpx` en cada punto que lo recibe: intento descartado, salto de redirección, respuesta final, `robots.txt` |

## 4. Otras comprobaciones mecánicas de arquitectura de este hito

| Comprobación | Dónde | Qué protege |
|---|---|---|
| El binario distribuido no enlaza `internal/httpx` ni sus dos módulos | `TestElBinarioNoEnlazaHTTPX` y `TestDependenciasDelBinario` (`internal/arch_test.go`), sobre `go list -deps ./cmd/kitlegal` | SC-014: la superficie del binario no cambia; se retira en H4 con la actualización justificada de la lista |
| Ningún tipo de `net/http` en la superficie exportada de `internal/httpx` | `TestSuperficieExportada` (`internal/httpx`): analiza el paquete con `go/parser` y recorre con `go/ast` las declaraciones exportadas (funciones, métodos, campos y tipos); falla si en alguna firma o campo exportado aparece un selector del paquete `http` | FR-002 |
| El adaptador de prueba compila solo contra lo exportado | `internal/httpx/adaptador_test.go` es `package httpx_test` | FR-061 |
| El adaptador de prueba no está en ningún registro de binario | `TestElBinarioNoEnlazaLosEjemplos` (H1, sin cambios) y ausencia del tipo `adaptadorDePrueba` y del literal `"consultar"` fuera de `_test.go` en `internal/app`, `cmd` e `internal/httpx` (`grep -rln --include='*.go' --exclude='*_test.go'`, quickstart escenario 9; research D16) | FR-062 |

## 5. Qué falla, y cómo, ante una violación (SC-012)

| Intento | Quién lo detecta | Mensaje |
|---|---|---|
| `import "net/http"` en `internal/app` (o cualquier paquete que no sea `internal/httpx`) | `make lint` (`depguard`, lista `red`) **y** `go test ./internal/ -run 'TestArquitectura/R2'` | «R2: la red se usa a través de internal/httpx …» / «R2 · importación reservada: … importa "net/http"» |
| `http.Get(url)` o `http.NewRequest(…)` en `internal/httpx` | `make lint` (`noctx`) | «net/http.Get must not be called. use net/http.NewRequestWithContext …» |
| Llamar a `Pedir` sin `ctx` o sin `ejecucion` | el compilador | «not enough arguments in call to cliente.Pedir» |
| Un `*http.Response` sin cerrar en `internal/httpx` | `make lint` (`bodyclose`) | «response body must be closed» |

Se demuestra sobre una copia desechable del árbol ([quickstart.md](../quickstart.md), escenario 10),
nunca sobre el árbol de trabajo.
