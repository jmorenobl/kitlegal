# Contrato: reglas de arquitectura en H3

Estado de las cinco reglas de `docs/ROADMAP.md` §2 al cerrar H3, y los dos analizadores que este hito
estrena. Continúa `specs/003-h2-internal-httpx-cliente/contracts/reglas-de-arquitectura.md`.

## 1. Las cinco reglas

| # | Regla | Estado en H2 | Estado en H3 |
|---|---|---|---|
| R1 | `internal/core/**` no importa `internal/{source,httpx,cache,store,graph,render,cli,app}` ni I/O de la biblioteca estándar | Activa | **Sin cambio de regla**: nace el paquete `core` (`internal/core/{doc,cache}.go`) con el puerto `Cache`, que importa solo `context` y `time`. `internal/cache` importa `internal/core` (por el puerto) y `internal/core/schema` (por `Clase` y `ConClase`), nunca al revés (FR-038) |
| R2 | Solo `internal/httpx` importa `net/http` | Activa y con dueño | Sin cambio: **ningún fichero** de `internal/cache`, tests incluidos, importa `net/http` ni `net/http/httptest` (la lista `red` de `depguard` los deniega por prefijo fuera de `internal/httpx/**`, con `run.tests: true`). El adaptador de prueba usa `httpx.Replay` en las dos consultas y el literal `"GET"` en `httpx.Peticion` (research D12) |
| R3 | Solo `internal/{cache,store,graph}` importan SQLite y `database/sql` | Activa y vacía | **Activa y con su primer dueño real**: `internal/cache` importa `database/sql`, `modernc.org/sqlite` y `modernc.org/sqlite/lib`. `duenoObligatorio` sigue en `false` con la razón escrita en `arch_test.go` (FR-037): exigirlo requiere `store` (H12) y `graph` (H17) |
| R4 | Solo `internal/cli` y `cmd/` llaman a `os.Exit` | Activa, forma estricta | Sin cambio: `internal/cache` no llama a `os.Exit`; el proceso auxiliar de los tests de integración termina devolviendo de la función de test, y el resultado viaja en la salida estándar y el código que `testing` pone |
| R5 | Solo `internal/render` escribe en stdout; logs a stderr | Activa | Sin cambio: `internal/cache` escribe solo por el `*slog.Logger` recibido. El proceso auxiliar de test escribe con `t.Log`/`t.Errorf`, no con `fmt.Print*` |

## 2. Cómo se comprueba cada una en H3

| Regla | `depguard` | `forbidigo` | `internal/arch_test.go` | Cambio en H3 |
|---|---|---|---|---|
| R1 | lista `core` (cubre `internal/core/*.go` por `**/internal/core/**`) | — | `compruebaDominioPuro` (`paquetesBajo` incluye el propio `internal/core`) | ninguno |
| R2 | lista `red` | — | subprueba «R2» | ninguno |
| R3 | lista `sql` (`**` salvo `!**/internal/{cache,store,graph}/**`) | — | subprueba «R3»; sin `duenoObligatorio` | comentario de `.golangci.yml` y de `arch_test.go`: «R3 tiene dueño desde H3» |
| R4 | — | `^os\.Exit$` | — | ninguno |
| R5 | — | `^fmt\.Print(\|f\|ln)$`, `^os\.Stdout$`, `^os\.Stderr$` | — | ninguno |

`.golangci.yml` gana `run.build-tags: [integration]` para que las cinco reglas y todos los analizadores
alcancen `integracion_test.go` (FR-041). Ninguna exclusión nueva ni `//nolint` (SC-008).

## 3. Los dos analizadores que H3 estrena

Activos desde H0; por primera vez tienen código que vigilar. Verificado en su código (research.md,
tabla de verificación):

| Analizador | Qué exige | Cómo lo cumple `internal/cache` |
|---|---|---|
| `sqlclosecheck` v0.6.0 | Que todo `*sql.Rows` y `*sql.Stmt` se cierre | El paquete no crea ninguno: `Get` usa `QueryRowContext(...).Scan`, `Put` y las migraciones `ExecContext`, la versión `QueryRowContext`. Sin sentencias preparadas |
| `rowserrcheck` | Que se compruebe `Rows.Err` tras iterar | No hay iteración de filas; ninguna consulta devuelve más de una |

Que **vigilan de verdad** se demuestra en el quickstart (escenario 9): un `QueryContext` sin `Close` en
`integracion_test.go` sobre una copia desechable hace fallar `make lint` con `sqlclosecheck`, lo que
prueba a la vez que `run.build-tags` alcanza el fichero etiquetado.

## 4. Importaciones de `internal/cache` y de sus tests

| Fichero(s) | Importa del módulo | Importa de terceros | Nota |
|---|---|---|---|
| producto (`cliente.go`, `ruta.go`, `abrir.go`, `migraciones.go`, `entradas.go`, `errores.go`) | `internal/core`, `internal/core/schema` | `modernc.org/sqlite` (registro del driver), `modernc.org/sqlite/lib` (constantes de código) | más `database/sql`, `embed`, `os`, `path/filepath`, `log/slog`, `context`, `time`, `errors`, `fmt`, `sync`, `io/fs` y, solo en `ruta.go` y solo por la constante `ENOTDIR` de la regla «inexistente», `syscall` (ninguna lista de `depguard` ni regla R1-R5 lo restringe fuera de `internal/core`) |
| tests de caja blanca (`package cache`) | `internal/cli` (para `Clasificar`/`CodigoSalida` en `TestClasesDeError`), `internal/core/schema` | `stretchr/testify` | `database/sql` para manipular la base en los tests de esquema (permitido por R3) |
| `adaptador_test.go`, `integracion_test.go` (`package cache_test`) | `internal/app`, `internal/core/schema`, `internal/httpx`, `internal/cache` | `stretchr/testify` | **Ni `net/http` ni `net/http/httptest`**: la lista `red` de `depguard` (`**` salvo `!**/internal/httpx/**`, por prefijo) los denegaría. Las dos consultas del adaptador van por `httpx.Replay`: la primera sobre el único fixture del hito (`internal/cache/testdata/reproduccion/prueba/GET_http_fuente.prueba_norma.json`, escrito a mano) y la segunda sobre un directorio vacío (research D12). `os/exec` para el proceso auxiliar |

Un `import "net/http"` en cualquier fichero de `internal/cache` —también en un `_test.go`— falla en
`make lint` con «R2: la red se usa a través de internal/httpx …». Es la razón por la que el adaptador de
prueba de H3 no levanta un servidor local como el de H2 (que sí podía, por vivir bajo `internal/httpx`).

## 5. Otras comprobaciones mecánicas de arquitectura de este hito

| Comprobación | Dónde | Qué protege |
|---|---|---|
| El binario distribuido no enlaza `internal/cache` ni `modernc.org/*` | `TestElBinarioNoEnlazaCache` (nuevo, temporal hasta H4) y `TestDependenciasDelBinario` (lista sin cambios), sobre `go list -deps ./cmd/kitlegal` | SC-012, FR-044 |
| Ningún tipo de `database/sql` ni `modernc.org/sqlite` en la superficie exportada; ninguna declaración exportada fuera del contrato | `TestSuperficieExportada` (`internal/cache`, `go/parser` + `go/ast`) | FR-005 |
| El adaptador de prueba compila solo contra lo exportado | `adaptador_test.go` es `package cache_test` | FR-045 |
| El adaptador de prueba no está en ningún registro de binario | `TestElBinarioNoEnlazaLosEjemplos` (H1) y ausencia de `adaptadorDePrueba`, `"consultar"` y `"guardar"` fuera de `_test.go` en `internal/app`, `cmd` e `internal/cache` (quickstart) | FR-044 |
| `Cliente` implementa el puerto | `var _ core.Cache = (*Cliente)(nil)` | FR-001, FR-002 |

## 6. Qué falla, y cómo, ante una violación (SC-013)

| Intento | Quién lo detecta | Mensaje |
|---|---|---|
| `import "database/sql"` o `import "modernc.org/sqlite"` en `internal/app` (o en cualquier paquete que no sea `internal/{cache,store,graph}`) | `make lint` (`depguard`, lista `sql`) **y** `go test ./internal/ -run 'TestArquitectura/R3'` | «R3: el acceso a base de datos vive en internal/{cache,store,graph} …» / «R3 · importación reservada: … importa "database/sql"» |
| `import ".../internal/cache"` en `internal/core` | `make lint` (`depguard`, lista `core`) **y** `go test ./internal/ -run 'TestArquitectura/R1'` | «R1: el dominio no importa adaptadores; la caché se usa desde internal/source» / «R1 · el dominio es puro: … importa …/internal/cache» |
| `rows, _ := db.QueryContext(…)` sin `rows.Close()` en `internal/cache` (también en un fichero `//go:build integration`) | `make lint` (`sqlclosecheck`) | «Rows/Stmt/NamedStmt was not closed» |
| `for rows.Next() {…}` sin `rows.Err()` | `make lint` (`rowserrcheck`) | «rows.Err must be checked» |
| Un tipo exportado con un campo `*sql.DB` | `TestSuperficieExportada` | nombra la declaración |

Se demuestra sobre una copia desechable del árbol ([quickstart.md](../quickstart.md), escenario 9),
nunca sobre el árbol de trabajo.
