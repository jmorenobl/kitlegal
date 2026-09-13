# Evidencia SC-013 · los controles fallan ante cada intento de saltarse una regla

Tarea T012 de [`tasks.md`](../tasks.md) (FR-038, FR-039, SC-008, SC-013, obligación 8 del plan): el
escenario 9 de [`quickstart.md`](../quickstart.md) en sus cuatro variantes, sobre copias desechables del
árbol fuera del repositorio. Todas las salidas son literales.

## Condiciones

- **Fecha**: 2026-09-12. **Rama**: `h3-internal-cache-sqlite`, sobre `833ba00` (T011).
- **Toolchain**: `go version go1.27.1 darwin/arm64`, el que fija la directiva `toolchain` de `go.mod`.
- **Dónde**: una copia del árbol **por variante**, fuera del repositorio, en el directorio temporal de la
  sesión (`$TMP` abajo): `rsync -a --exclude .git <repositorio>/ "$TMP/sc013-<variante>/"`, que copia lo
  mismo que el `tar --exclude=.git` del escenario. Una copia por variante en vez de una sola en la que se
  pone y se quita cada violación: así corren en paralelo y ninguna hereda restos de la anterior.
- **Cómo**: `make -C <copia> …` y `go -C <copia> …`, que equivalen al `( cd "$COPIA" && … )` del
  escenario. Todas las órdenes van con `rtk proxy` para que la salida llegue en crudo.
- **Árbol de trabajo**: ninguna orden escribió en él. `rtk proxy git status --porcelain`, ejecutado tras
  las cuatro variantes, muestra solo los dos ficheros que el workflow `hito` actualiza al lanzar la tarea:

  ```
   M specs/004-h3-internal-cache-sqlite/gates/tarea-actual.json
   M specs/004-h3-internal-cache-sqlite/gates/tareas-intentos.json
  ```

## Resumen

| Variante | Violación | `make ci` | Qué nombra | Recuentos del escenario | Test de arquitectura |
|---|---|---|---|---|---|
| a | `database/sql` importado en `internal/app` | código 2, en `lint` | `depguard`, lista `sql`: **R3** | `1` y `1` | `R3 · importación reservada` |
| b | `internal/core/violacion` importa `internal/cache` | código 2, en `lint` | `depguard`, lista `core`: **R1** | `1` y `1` | `R1 · el dominio es puro` (`compruebaDominioPuro`) |
| c | `*sql.Rows` sin cerrar en un fichero `//go:build integration` | código 2, en `lint` | **`sqlclosecheck`** | `2` (hallazgo y resumen); `0` sin `run.build-tags` | — |
| d | `net/http` importado en un test de `internal/cache` | código 2, en `lint` | `depguard`, lista `red`: **R2** | `1`; también `1` con `net/http/httptest` (prefijo) | — |

`make ci` se detiene en su primer prerrequisito en rojo. En las cuatro variantes es `lint`, tras
`fmt-check` en verde, y el mensaje de la regla es lo último que imprime antes de
`make: *** [lint] Error 1`. En a y b la segunda capa, `TestArquitectura`, que `make test` ejecuta dentro de
`make ci`, falla además por su cuenta nombrando la misma regla.

**Las sondas de a, b y c no son las que el escenario traía al empezar la tarea.** Con aquellas, `make ci`
fallaba pero no nombraba R3, R1 ni `sqlclosecheck`, porque otro linter ocupaba la misma línea (§1). Se
corrigió el escenario 9 de `quickstart.md` y no ningún control; las sondas corregidas son las de §2 a §4.

---

## 0. Control positivo: la copia limpia está en verde

Sin ninguna violación, la copia pasa los dos prerrequisitos de `make ci` que las variantes hacen fallar y
el test de arquitectura. Cada rojo de abajo lo produce, por tanto, la violación y no la copia.

```
$ rtk proxy make -C "$TMP/sc013-base" fmt-check lint
go tool -modfile=tools/golangci-lint/go.mod golangci-lint fmt --diff ./...
# github.com/golangci/golangci-lint/v2/cmd/golangci-lint
ld: warning: -bind_at_load is deprecated on macOS
go tool -modfile=tools/golangci-lint/go.mod golangci-lint run ./...
# github.com/golangci/golangci-lint/v2/cmd/golangci-lint
ld: warning: -bind_at_load is deprecated on macOS
0 issues.
```

Código de salida 0.

```
$ rtk proxy go -C "$TMP/sc013-base" test -count=1 ./internal/ -run 'TestArquitectura' -v
=== RUN   TestArquitectura
=== PAUSE TestArquitectura
=== CONT  TestArquitectura
=== RUN   TestArquitectura/R1_·_el_dominio_no_importa_el_kernel,_los_adaptadores_ni_entrada_y_salida
=== PAUSE TestArquitectura/R1_·_el_dominio_no_importa_el_kernel,_los_adaptadores_ni_entrada_y_salida
=== RUN   TestArquitectura/R2_·_solo_internal/httpx_importa_net/http
=== PAUSE TestArquitectura/R2_·_solo_internal/httpx_importa_net/http
=== RUN   TestArquitectura/R3_·_solo_internal/{cache,store,graph}_importan_SQLite_y_database/sql
=== PAUSE TestArquitectura/R3_·_solo_internal/{cache,store,graph}_importan_SQLite_y_database/sql
=== CONT  TestArquitectura/R1_·_el_dominio_no_importa_el_kernel,_los_adaptadores_ni_entrada_y_salida
=== CONT  TestArquitectura/R3_·_solo_internal/{cache,store,graph}_importan_SQLite_y_database/sql
=== CONT  TestArquitectura/R2_·_solo_internal/httpx_importa_net/http
--- PASS: TestArquitectura (0.08s)
    --- PASS: TestArquitectura/R1_·_el_dominio_no_importa_el_kernel,_los_adaptadores_ni_entrada_y_salida (0.00s)
    --- PASS: TestArquitectura/R3_·_solo_internal/{cache,store,graph}_importan_SQLite_y_database/sql (0.00s)
    --- PASS: TestArquitectura/R2_·_solo_internal/httpx_importa_net/http (0.00s)
PASS
ok  	github.com/jmorenobl/kitlegal/internal	0.434s
```

Código de salida 0.

---

## 1. Hallazgo: con las sondas originales, `make ci` falla sin nombrar la regla

Primera ejecución, con las sondas tal como el escenario 9 las traía: `import _ "database/sql"` sin
comentario (a), el subpaquete sin comentario de paquete ni de importación (b) y `_ = rows; return nil`
(c). Las tres copias fallan `make ci`, pero el hallazgo que se imprime es de otro linter y los recuentos del
escenario habrían salido `0`.

### a original

```
$ rtk proxy make -C "$TMP/sc013-a" ci
go tool -modfile=tools/golangci-lint/go.mod golangci-lint fmt --diff ./...
# github.com/golangci/golangci-lint/v2/cmd/golangci-lint
ld: warning: -bind_at_load is deprecated on macOS
go tool -modfile=tools/golangci-lint/go.mod golangci-lint run ./...
# github.com/golangci/golangci-lint/v2/cmd/golangci-lint
ld: warning: -bind_at_load is deprecated on macOS
internal/app/violacion_r3.go:3:8: blank-imports: a blank import should be only in a main or test package, or have a comment justifying it (revive)
import _ "database/sql"
       ^
1 issues:
* revive: 1
make: *** [lint] Error 1
```

Código de salida 2.

### b original

```
$ rtk proxy make -C "$TMP/sc013-b" ci
go tool -modfile=tools/golangci-lint/go.mod golangci-lint fmt --diff ./...
# github.com/golangci/golangci-lint/v2/cmd/golangci-lint
ld: warning: -bind_at_load is deprecated on macOS
go tool -modfile=tools/golangci-lint/go.mod golangci-lint run ./...
# github.com/golangci/golangci-lint/v2/cmd/golangci-lint
ld: warning: -bind_at_load is deprecated on macOS
internal/core/violacion/violacion.go:1:1: package-comments: should have a package comment (revive)
package violacion
^
internal/core/violacion/violacion.go:3:8: blank-imports: a blank import should be only in a main or test package, or have a comment justifying it (revive)
import _ "github.com/jmorenobl/kitlegal/internal/cache"
       ^
2 issues:
* revive: 2
make: *** [lint] Error 1
```

Código de salida 2.

### c original

```
$ rtk proxy make -C "$TMP/sc013-c" ci
go tool -modfile=tools/golangci-lint/go.mod golangci-lint fmt --diff ./...
# github.com/golangci/golangci-lint/v2/cmd/golangci-lint
ld: warning: -bind_at_load is deprecated on macOS
go tool -modfile=tools/golangci-lint/go.mod golangci-lint run ./...
# github.com/golangci/golangci-lint/v2/cmd/golangci-lint
ld: warning: -bind_at_load is deprecated on macOS
internal/cache/violacion_sqlclose_test.go:11:30: rows.Err must be checked (rowserrcheck)
	rows, err := db.QueryContext(ctx, "SELECT 1")
	                            ^
internal/cache/violacion_sqlclose_test.go:10:6: func filasSinCerrar is unused (unused)
func filasSinCerrar(ctx context.Context, db *sql.DB) error {
     ^
2 issues:
* rowserrcheck: 1
* unused: 1
make: *** [lint] Error 1
```

Código de salida 2.

### Causa: un solo hallazgo por línea

golangci-lint deduplica los hallazgos por línea (`issues.uniq-by-line`, activo por omisión en la v2 y sin
tocar en `.golangci.yml`). Las reglas sí detectaban las tres violaciones: con la deduplicación desactivada
**solo en la línea de órdenes de la copia**, la misma línea lleva los dos hallazgos.

```
$ rtk proxy go -C "$TMP/sc013-a" tool -modfile=tools/golangci-lint/go.mod golangci-lint run --uniq-by-line=false ./...
# github.com/golangci/golangci-lint/v2/cmd/golangci-lint
ld: warning: -bind_at_load is deprecated on macOS
internal/app/violacion_r3.go:3:8: import 'database/sql' is not allowed from list 'sql': R3: el acceso a base de datos vive en internal/{cache,store,graph}; el resto del árbol los usa por su interfaz (depguard)
import _ "database/sql"
       ^
internal/app/violacion_r3.go:3:8: blank-imports: a blank import should be only in a main or test package, or have a comment justifying it (revive)
import _ "database/sql"
       ^
2 issues:
* depguard: 1
* revive: 1
```

Código de salida 1.

```
$ rtk proxy go -C "$TMP/sc013-b" tool -modfile=tools/golangci-lint/go.mod golangci-lint run --uniq-by-line=false ./...
# github.com/golangci/golangci-lint/v2/cmd/golangci-lint
ld: warning: -bind_at_load is deprecated on macOS
internal/core/violacion/violacion.go:3:8: import 'github.com/jmorenobl/kitlegal/internal/cache' is not allowed from list 'core': R1: el dominio no importa adaptadores; la caché se usa desde internal/source (depguard)
import _ "github.com/jmorenobl/kitlegal/internal/cache"
       ^
internal/core/violacion/violacion.go:1:1: package-comments: should have a package comment (revive)
package violacion
^
internal/core/violacion/violacion.go:3:8: blank-imports: a blank import should be only in a main or test package, or have a comment justifying it (revive)
import _ "github.com/jmorenobl/kitlegal/internal/cache"
       ^
3 issues:
* depguard: 1
* revive: 2
```

Código de salida 1.

```
$ rtk proxy go -C "$TMP/sc013-c" tool -modfile=tools/golangci-lint/go.mod golangci-lint run --uniq-by-line=false ./...
# github.com/golangci/golangci-lint/v2/cmd/golangci-lint
ld: warning: -bind_at_load is deprecated on macOS
internal/cache/violacion_sqlclose_test.go:11:30: rows.Err must be checked (rowserrcheck)
	rows, err := db.QueryContext(ctx, "SELECT 1")
	                            ^
internal/cache/violacion_sqlclose_test.go:11:30: Rows/Stmt/NamedStmt was not closed (sqlclosecheck)
	rows, err := db.QueryContext(ctx, "SELECT 1")
	                            ^
internal/cache/violacion_sqlclose_test.go:10:6: func filasSinCerrar is unused (unused)
func filasSinCerrar(ctx context.Context, db *sql.DB) error {
     ^
3 issues:
* rowserrcheck: 1
* sqlclosecheck: 1
* unused: 1
```

Código de salida 1.

### Remedio: sondas que aíslan la regla, sin tocar ningún control

El defecto estaba en las sondas, que cometían dos infracciones en la misma línea, y no en los controles.
Se corrigió el escenario 9 de `quickstart.md`, con una nota que explica por qué cada violación tiene que
ocupar sola su línea:

- **a**: la importación en blanco lleva el comentario que `revive` (`blank-imports`) pide fuera de un
  `package main` o de test.
- **b**: igual, y además el comentario de paquete, para que ningún otro hallazgo acompañe al de R1.
- **c**: el `*sql.Rows` sigue sin cerrarse, pero devuelve `rows.Err()`, de modo que `rowserrcheck` no ocupa
  la línea que marca `sqlclosecheck`.

Ni `.golangci.yml`, ni `internal/arch_test.go`, ni el `Makefile` cambian. La variante d no necesitaba
corrección: en un fichero de test `revive` no marca la importación en blanco.

**Queda para la revisión humana, fuera de esta tarea.** Fijar `issues.uniq-by-line: false` haría que el
primer `make ci` nombrara todas las reglas incumplidas de una línea aunque otro linter la marque también.
El caso realista es importar en blanco un controlador SQL, sin comentario, fuera de los paquetes de
almacenamiento: hoy ese primer rojo nombra a `revive`, R3 aparece en cuanto se añade el comentario, y
`TestArquitectura` lo nombra en cualquier caso. No se aplica aquí porque ningún requisito lo pide, T012 no
declara `.golangci.yml` y el diff de esa configuración en H3 es una lista cerrada que el escenario 11
comprueba (T013).

---

## 2. Variante a · `database/sql` fuera de `internal/{cache,store,graph}` (R3)

Sonda, `$TMP/sc013-a2/internal/app/violacion_r3.go`:

```go
package app

import _ "database/sql" // Violación deliberada de R3 (escenario 9 de quickstart.md).
```

```
$ rtk proxy make -C "$TMP/sc013-a2" ci
go tool -modfile=tools/golangci-lint/go.mod golangci-lint fmt --diff ./...
# github.com/golangci/golangci-lint/v2/cmd/golangci-lint
ld: warning: -bind_at_load is deprecated on macOS
go tool -modfile=tools/golangci-lint/go.mod golangci-lint run ./...
# github.com/golangci/golangci-lint/v2/cmd/golangci-lint
ld: warning: -bind_at_load is deprecated on macOS
internal/app/violacion_r3.go:3:8: import 'database/sql' is not allowed from list 'sql': R3: el acceso a base de datos vive en internal/{cache,store,graph}; el resto del árbol los usa por su interfaz (depguard)
import _ "database/sql" // Violación deliberada de R3 (escenario 9 de quickstart.md).
       ^
1 issues:
* depguard: 1
make: *** [lint] Error 1
```

Código de salida 2.

```
$ rtk proxy go -C "$TMP/sc013-a2" test -count=1 ./internal/ -run 'TestArquitectura/R3'
--- FAIL: TestArquitectura (0.08s)
    --- FAIL: TestArquitectura/R3_·_solo_internal/{cache,store,graph}_importan_SQLite_y_database/sql (0.00s)
        arch_test.go:82: R3 · importación reservada: github.com/jmorenobl/kitlegal/internal/app importa "database/sql" (lo prohíbe "database/sql"). Está reservada a github.com/jmorenobl/kitlegal/internal/cache, github.com/jmorenobl/kitlegal/internal/store, github.com/jmorenobl/kitlegal/internal/graph: el acceso a SQLite vive en los tres paquetes de almacenamiento —caché, almacén y grafo—; el resto del árbol los usa a través de su interfaz (contracts/reglas-de-arquitectura.md R3).
FAIL
FAIL	github.com/jmorenobl/kitlegal/internal	0.426s
FAIL
```

Código de salida 1.

Recuentos del escenario (esperado: `1` y `1`):

```
$ rtk proxy make -C "$TMP/sc013-a2" lint 2>&1 | grep -c 'R3: el acceso a base de datos vive en internal/{cache,store,graph}'
1
$ rtk proxy go -C "$TMP/sc013-a2" test ./internal/ -run 'TestArquitectura/R3' 2>&1 | grep -c 'R3 · importación reservada'
1
```

---

## 3. Variante b · el dominio importa el adaptador de caché (R1)

Sonda, `$TMP/sc013-b2/internal/core/violacion/violacion.go`: un subpaquete nuevo del dominio, porque un
import en el propio `internal/core` sería un ciclo que el compilador rechazaría antes de nombrar la regla.

```go
// Package violacion incumple R1 a propósito (escenario 9 de quickstart.md).
package violacion

import _ "github.com/jmorenobl/kitlegal/internal/cache" // Violación deliberada de R1.
```

```
$ rtk proxy make -C "$TMP/sc013-b2" ci
go tool -modfile=tools/golangci-lint/go.mod golangci-lint fmt --diff ./...
# github.com/golangci/golangci-lint/v2/cmd/golangci-lint
ld: warning: -bind_at_load is deprecated on macOS
go tool -modfile=tools/golangci-lint/go.mod golangci-lint run ./...
# github.com/golangci/golangci-lint/v2/cmd/golangci-lint
ld: warning: -bind_at_load is deprecated on macOS
internal/core/violacion/violacion.go:4:8: import 'github.com/jmorenobl/kitlegal/internal/cache' is not allowed from list 'core': R1: el dominio no importa adaptadores; la caché se usa desde internal/source (depguard)
import _ "github.com/jmorenobl/kitlegal/internal/cache" // Violación deliberada de R1.
       ^
1 issues:
* depguard: 1
make: *** [lint] Error 1
```

Código de salida 2.

```
$ rtk proxy go -C "$TMP/sc013-b2" test -count=1 ./internal/ -run 'TestArquitectura/R1'
--- FAIL: TestArquitectura (0.07s)
    --- FAIL: TestArquitectura/R1_·_el_dominio_no_importa_el_kernel,_los_adaptadores_ni_entrada_y_salida (0.00s)
        arch_test.go:62: R1 · el dominio es puro: github.com/jmorenobl/kitlegal/internal/core/violacion importa "github.com/jmorenobl/kitlegal/internal/cache" (lo prohíbe "github.com/jmorenobl/kitlegal/internal/cache"). internal/core no depende del kernel, de los adaptadores ni de la entrada y salida de la biblioteca estándar: el registro llega como *slog.Logger al método Ejecutar y la presentación se inyecta desde la raíz de composición (contracts/reglas-de-arquitectura.md R1).
FAIL
FAIL	github.com/jmorenobl/kitlegal/internal	0.383s
FAIL
```

Código de salida 1. El fallo sale de `compruebaDominioPuro` (línea 62 es su llamada en la subprueba R1).

Recuentos del escenario (esperado: `1` y `1`):

```
$ rtk proxy make -C "$TMP/sc013-b2" lint 2>&1 | grep -c 'R1: el dominio no importa adaptadores; la caché se usa desde internal/source'
1
$ rtk proxy go -C "$TMP/sc013-b2" test ./internal/ -run 'TestArquitectura/R1' 2>&1 | grep -c 'R1 · el dominio es puro'
1
```

---

## 4. Variante c · un `*sql.Rows` sin cerrar en un fichero etiquetado `integration`

Sonda, `$TMP/sc013-c2/internal/cache/violacion_sqlclose_test.go`:

```go
//go:build integration

package cache_test

import (
	"context"
	"database/sql"
)

func filasSinCerrar(ctx context.Context, db *sql.DB) error {
	rows, err := db.QueryContext(ctx, "SELECT 1")
	if err != nil {
		return err
	}

	return rows.Err()
}
```

```
$ rtk proxy make -C "$TMP/sc013-c2" ci
go tool -modfile=tools/golangci-lint/go.mod golangci-lint fmt --diff ./...
# github.com/golangci/golangci-lint/v2/cmd/golangci-lint
ld: warning: -bind_at_load is deprecated on macOS
go tool -modfile=tools/golangci-lint/go.mod golangci-lint run ./...
# github.com/golangci/golangci-lint/v2/cmd/golangci-lint
ld: warning: -bind_at_load is deprecated on macOS
internal/cache/violacion_sqlclose_test.go:11:30: Rows/Stmt/NamedStmt was not closed (sqlclosecheck)
	rows, err := db.QueryContext(ctx, "SELECT 1")
	                            ^
internal/cache/violacion_sqlclose_test.go:10:6: func filasSinCerrar is unused (unused)
func filasSinCerrar(ctx context.Context, db *sql.DB) error {
     ^
2 issues:
* sqlclosecheck: 1
* unused: 1
make: *** [lint] Error 1
```

Código de salida 2.

Recuento del escenario (esperado: `2`, la línea del hallazgo y la del resumen `* sqlclosecheck: 1`):

```
$ rtk proxy make -C "$TMP/sc013-c2" lint 2>&1 | grep -c 'sqlclosecheck'
2
```

**Contraprueba: `sqlclosecheck` solo alcanza el fichero gracias a `run.build-tags`.** Se hizo en otra copia
(`$TMP/sc013-c-sin-tags`) con la misma sonda y con `.golangci.yml` sin el bloque `build-tags: [integration]`
bajo `run` ni su comentario. El fichero etiquetado queda fuera del análisis y el lint pasa en verde:

```
$ rtk proxy make -C "$TMP/sc013-c-sin-tags" lint
go tool -modfile=tools/golangci-lint/go.mod golangci-lint run ./...
# github.com/golangci/golangci-lint/v2/cmd/golangci-lint
ld: warning: -bind_at_load is deprecated on macOS
0 issues.
$ rtk proxy make -C "$TMP/sc013-c-sin-tags" lint 2>&1 | grep -c 'sqlclosecheck'
0
```

Código de salida 0 en el lint.

---

## 5. Variante d · `net/http` en un test de `internal/cache` (R2)

Sonda, `$TMP/sc013-d/internal/cache/violacion_r2_test.go`, sin cambios respecto al escenario original:

```go
package cache_test

import _ "net/http"
```

```
$ rtk proxy make -C "$TMP/sc013-d" ci
go tool -modfile=tools/golangci-lint/go.mod golangci-lint fmt --diff ./...
# github.com/golangci/golangci-lint/v2/cmd/golangci-lint
ld: warning: -bind_at_load is deprecated on macOS
go tool -modfile=tools/golangci-lint/go.mod golangci-lint run ./...
# github.com/golangci/golangci-lint/v2/cmd/golangci-lint
ld: warning: -bind_at_load is deprecated on macOS
internal/cache/violacion_r2_test.go:3:8: import 'net/http' is not allowed from list 'red': R2: la red se usa a través de internal/httpx, que concentra reintentos, límite por sitio, robots.txt y User-Agent identificable (depguard)
import _ "net/http"
       ^
1 issues:
* depguard: 1
make: *** [lint] Error 1
```

Código de salida 2.

Recuento del escenario (esperado: `1`):

```
$ rtk proxy make -C "$TMP/sc013-d" lint 2>&1 | grep -c 'R2: la red se usa a través de internal/httpx'
1
```

**Por prefijo.** En otra copia (`$TMP/sc013-d-prefijo`), el fichero
`internal/cache/violacion_r2_prefijo_test.go` importa en blanco un subpaquete, `net/http/httptest`, en
lugar de `net/http`. La lista `red` lo deniega igual, porque `depguard` compara `pkg` por prefijo:

```
$ rtk proxy make -C "$TMP/sc013-d-prefijo" ci
go tool -modfile=tools/golangci-lint/go.mod golangci-lint fmt --diff ./...
# github.com/golangci/golangci-lint/v2/cmd/golangci-lint
ld: warning: -bind_at_load is deprecated on macOS
go tool -modfile=tools/golangci-lint/go.mod golangci-lint run ./...
# github.com/golangci/golangci-lint/v2/cmd/golangci-lint
ld: warning: -bind_at_load is deprecated on macOS
internal/cache/violacion_r2_prefijo_test.go:3:8: import 'net/http/httptest' is not allowed from list 'red': R2: la red se usa a través de internal/httpx, que concentra reintentos, límite por sitio, robots.txt y User-Agent identificable (depguard)
import _ "net/http/httptest"
       ^
1 issues:
* depguard: 1
make: *** [lint] Error 1
$ rtk proxy make -C "$TMP/sc013-d-prefijo" lint 2>&1 | grep -c 'R2: la red se usa a través de internal/httpx'
1
```

Código de salida 2 en `make ci`.
