# Evidencia SC-012 — los controles fallan ante cada intento de saltárselos

**Tarea**: T017 · **Hito**: H2 (`internal/httpx`) · **Fecha**: 2026-09-12
**Rama**: `h2-internal-httpx-cliente` · **Commit del árbol copiado**: `ced329a` (T016)
**Cubre**: FR-061, SC-012, obligación 6 del plan, escenario 10 de `quickstart.md`

---

## 0. Cómo se ejecutó

Sobre una **copia desechable del árbol fuera del repositorio**, nunca sobre el árbol de trabajo. El
escenario 10 del quickstart escribe la copia con `COPIA=$(mktemp -d)` y `tar … | tar …`; aquí se usó
`rsync` con las mismas exclusiones y un directorio desechable del espacio temporal de la sesión, porque
el entorno de ejecución no autoriza `tar` ni la sustitución de órdenes. La copia es la misma: el árbol
completo **sin** `.git`, `bin` ni `coverage.out`.

```bash
rsync -a --exclude=.git --exclude=bin --exclude=coverage.out \
  /Users/jorge/Projects/kitlegal/ "$COPIA"/
```

La salida literal se captura con `rtk proxy`, que ejecuta la orden sin el filtro de resumen del hook
(sin él, `go test` llega reescrito y `--- FAIL` no aparece).

Versiones: `go version go1.27.1 darwin/arm64`; `golangci-lint v2.13.2` con `noctx v0.5.1`,
`depguard v2.2.1`, `bodyclose` y `gofumpt v0.11.0`, todos fijados en `tools/golangci-lint/go.mod`.

**Línea base — la copia intacta está en verde** (`rtk proxy make -C "$COPIA" ci`, rc=0). Se repitió
después de retirar las variantes a, b y c, para que cada rojo sea atribuible solo a la violación
inyectada:

```
go tool -modfile=tools/golangci-lint/go.mod golangci-lint fmt --diff ./...
go tool -modfile=tools/golangci-lint/go.mod golangci-lint run ./...
0 issues.
go test -race -shuffle=on -coverprofile=coverage.out ./...
ok  	github.com/jmorenobl/kitlegal/cmd/kitlegal	1.256s	coverage: 0.0% of statements
ok  	github.com/jmorenobl/kitlegal/internal	1.506s	coverage: [no statements]
ok  	github.com/jmorenobl/kitlegal/internal/app	2.516s	coverage: 92.8% of statements
	github.com/jmorenobl/kitlegal/internal/app/ejemplo		coverage: 0.0% of statements
	github.com/jmorenobl/kitlegal/internal/app/ejemplo/kitlegal-e2e		coverage: 0.0% of statements
ok  	github.com/jmorenobl/kitlegal/internal/cli	2.007s	coverage: 98.6% of statements
ok  	github.com/jmorenobl/kitlegal/internal/core/schema	1.750s	coverage: 90.1% of statements
ok  	github.com/jmorenobl/kitlegal/internal/httpx	13.354s	coverage: 94.8% of statements
ok  	github.com/jmorenobl/kitlegal/internal/render	2.155s	coverage: 95.8% of statements
go tool -modfile=tools/govulncheck/go.mod govulncheck ./...
No vulnerabilities found.
schema-check: no hay schemas/ todavía; los aportan H4 (borrador) y H10 (contrato)
go tool -modfile=tools/gitleaks/go.mod gitleaks dir . --redact --no-banner
INF scanned ~7062242 bytes (7.06 MB) in 651ms
INF no leaks found
go mod verify
all modules verified
== tools/gitleaks
all modules verified
== tools/golangci-lint
all modules verified
== tools/govulncheck
all modules verified
== tools/lefthook
all modules verified
go mod tidy -diff
ci: todos los controles en verde
```

Ninguna variante llega a `vuln`: `make ci` se detiene en el primer objetivo rojo. Ninguna orden de esta
evidencia consultó una fuente legal ni usó `KITLEGAL_RECORD`.

---

## 1. Variante a — `*http.Response` sin cerrar, en un paquete del kernel (`bodyclose` + `depguard` + R2)

Fichero inyectado, tal como lo escribe el escenario 10.a, en `internal/app/violacion.go` de la copia:

```go
package app

import "net/http"

func violacion() (int, error) {
	respuesta, err := http.Get("http://127.0.0.1:1/")
	if err != nil {
		return 0, err
	}

	return respuesta.StatusCode, nil
}
```

`rtk proxy make -C "$COPIA" ci` → **rc=2**, se detiene en `lint`:

```
go tool -modfile=tools/golangci-lint/go.mod golangci-lint fmt --diff ./...
go tool -modfile=tools/golangci-lint/go.mod golangci-lint run ./...
internal/app/violacion.go:6:28: response body must be closed (bodyclose)
	respuesta, err := http.Get("http://127.0.0.1:1/")
	                          ^
internal/app/violacion.go:3:8: import 'net/http' is not allowed from list 'red': R2: la red se usa a través de internal/httpx, que concentra reintentos, límite por sitio, robots.txt y User-Agent identificable (depguard)
import "net/http"
       ^
internal/app/violacion.go:5:6: func violacion is unused (unused)
func violacion() (int, error) {
     ^
3 issues:
* bodyclose: 1
* depguard: 1
* unused: 1
make: *** [lint] Error 1
```

Y la subprueba de arquitectura, `rtk proxy go test -C "$COPIA" -count=1 ./internal/ -run TestArquitectura/R2`
→ **rc=1**:

```
--- FAIL: TestArquitectura (0.07s)
    --- FAIL: TestArquitectura/R2_·_solo_internal/httpx_importa_net/http (0.00s)
        arch_test.go:68: R2 · importación reservada: github.com/jmorenobl/kitlegal/internal/app importa "net/http" (lo prohíbe "net/http"). Está reservada a github.com/jmorenobl/kitlegal/internal/httpx: la biblioteca HTTP se usa a través de internal/httpx, que es lo que concentra los reintentos, el límite de peticiones por sitio, robots.txt y el User-Agent identificable (contracts/reglas-de-arquitectura.md R2).
```

Esta variante es también la que el enunciado de T017 nombra en cuarto lugar —**la importación de la
biblioteca HTTP en un paquete del kernel, que falla en `depguard` y en la subprueba R2**—: las dos
salidas de arriba lo demuestran, la primera nombrando `depguard` con el texto de R2 y la segunda
nombrando «R2 · importación reservada».

**Hallazgo 1 — `noctx` queda tapado aquí, y no por no dispararse.** El escenario esperaba además una
línea `must not be called`. No aparece porque `golangci-lint` trae `--uniq-by-line` activo por omisión
(`.golangci.yml` no lo cambia) y `bodyclose` y `noctx` señalan **la misma posición**, `6:28`: solo se
imprime la primera. Con el mismo fichero y la misma configuración, desactivando únicamente esa
deduplicación, `noctx` sí nombra la regla:

```
$ go tool -modfile=tools/golangci-lint/go.mod golangci-lint run --uniq-by-line=false ./internal/app/
internal/app/violacion.go:6:28: response body must be closed (bodyclose)
	respuesta, err := http.Get("http://127.0.0.1:1/")
	                          ^
internal/app/violacion.go:3:8: import 'net/http' is not allowed from list 'red': R2: la red se usa a través de internal/httpx, que concentra reintentos, límite por sitio, robots.txt y User-Agent identificable (depguard)
import "net/http"
       ^
internal/app/violacion.go:6:28: net/http.Get must not be called. use net/http.NewRequestWithContext and (*net/http.Client).Do(*http.Request) (noctx)
	respuesta, err := http.Get("http://127.0.0.1:1/")
	                          ^
internal/app/violacion.go:5:6: func violacion is unused (unused)
func violacion() (int, error) {
     ^
4 issues:
* bodyclose: 1
* depguard: 1
* noctx: 1
* unused: 1
```

No es un agujero del control: el intento se rechaza igual, y lo rechazan tres reglas distintas. Es una
imprecisión del texto del escenario, corregida en `quickstart.md` (ver §6).

---

## 2. Variante b — `http.NewRequest` sin contexto (`noctx`)

### 2.1 Tal como lo escribe el escenario, en una sola línea

`internal/httpx/violacion.go` de la copia:

```go
package httpx

import "net/http"

func sinContexto() (*http.Request, error) { return http.NewRequest(http.MethodGet, "http://127.0.0.1:1/", nil) }
```

`rtk proxy make -C "$COPIA" ci` → **rc=2**, y se detiene **antes de `lint`**, en `fmt-check`:

```
go tool -modfile=tools/golangci-lint/go.mod golangci-lint fmt --diff ./...
diff internal/httpx/violacion.go.orig internal/httpx/violacion.go
--- internal/httpx/violacion.go.orig
+++ internal/httpx/violacion.go
@@ -2,4 +2,6 @@
 
 import "net/http"
 
-func sinContexto() (*http.Request, error) { return http.NewRequest(http.MethodGet, "http://127.0.0.1:1/", nil) }
+func sinContexto() (*http.Request, error) {
+	return http.NewRequest(http.MethodGet, "http://127.0.0.1:1/", nil)
+}
make: *** [fmt-check] Error 1
```

Y el `make lint` del escenario, que sí llega al analizador, tampoco nombra `noctx`:

```
go tool -modfile=tools/golangci-lint/go.mod golangci-lint run ./...
internal/httpx/violacion.go:5:6: func sinContexto is unused (unused)
func sinContexto() (*http.Request, error) { return http.NewRequest(http.MethodGet, "http://127.0.0.1:1/", nil) }
     ^
1 issues:
* unused: 1
make: *** [lint] Error 1
```

**Hallazgo 2 — el mismo `--uniq-by-line`.** Con todo en una línea, `unused` (5:6), `gofumpt` (5:1),
`goimports` (5:1) y `noctx` (5:67) caen en la línea 5 y solo sobrevive el primero. Desactivando solo la
deduplicación se ve que `noctx` está ahí:

```
$ go tool -modfile=tools/golangci-lint/go.mod golangci-lint run --uniq-by-line=false ./internal/httpx/
internal/httpx/violacion.go:5:1: File is not properly formatted (gofumpt)
func sinContexto() (*http.Request, error) { return http.NewRequest(http.MethodGet, "http://127.0.0.1:1/", nil) }
^
internal/httpx/violacion.go:5:1: File is not properly formatted (goimports)
func sinContexto() (*http.Request, error) { return http.NewRequest(http.MethodGet, "http://127.0.0.1:1/", nil) }
^
internal/httpx/violacion.go:5:67: net/http.NewRequest must not be called. use net/http.NewRequestWithContext (noctx)
func sinContexto() (*http.Request, error) { return http.NewRequest(http.MethodGet, "http://127.0.0.1:1/", nil) }
                                                                  ^
internal/httpx/violacion.go:5:6: func sinContexto is unused (unused)
func sinContexto() (*http.Request, error) { return http.NewRequest(http.MethodGet, "http://127.0.0.1:1/", nil) }
     ^
4 issues:
* gofumpt: 1
* goimports: 1
* noctx: 1
* unused: 1
```

### 2.2 La misma violación escrita como el propio control de formato exige

Sin ningún cambio de configuración: el cuerpo en su línea, que es lo que `fmt-check` pide en 2.1.

```go
package httpx

import "net/http"

func sinContexto() (*http.Request, error) {
	return http.NewRequest(http.MethodGet, "http://127.0.0.1:1/", nil)
}
```

`rtk proxy make -C "$COPIA" ci` → **rc=2**, ahora sí en `lint` y nombrando `noctx`:

```
go tool -modfile=tools/golangci-lint/go.mod golangci-lint fmt --diff ./...
go tool -modfile=tools/golangci-lint/go.mod golangci-lint run ./...
internal/httpx/violacion.go:6:24: net/http.NewRequest must not be called. use net/http.NewRequestWithContext (noctx)
	return http.NewRequest(http.MethodGet, "http://127.0.0.1:1/", nil)
	                      ^
internal/httpx/violacion.go:5:6: func sinContexto is unused (unused)
func sinContexto() (*http.Request, error) {
     ^
2 issues:
* noctx: 1
* unused: 1
make: *** [lint] Error 1
```

Las dos formas del intento acaban en rojo; la segunda, además, con el nombre de la regla que la prohíbe.
El escenario 10.b de `quickstart.md` pasa a escribirse así (ver §6).

---

## 3. Variante c — llamar a `Pedir` sin contexto: no compila

`internal/httpx/violacion_test.go` de la copia, tal como lo escribe el escenario 10.c:

```go
package httpx_test

import (
	"testing"

	"github.com/jmorenobl/kitlegal/internal/core/schema"
	"github.com/jmorenobl/kitlegal/internal/httpx"
)

func TestSinContexto(t *testing.T) {
	c, _ := httpx.New()
	_, _ = c.Pedir(schema.Contexto{}, httpx.Peticion{Metodo: "GET", URL: "http://127.0.0.1:1/"})
}
```

`rtk proxy go vet -C "$COPIA" ./internal/httpx/` → **rc=1**:

```
# github.com/jmorenobl/kitlegal/internal/httpx_test
# [github.com/jmorenobl/kitlegal/internal/httpx_test]
vet: internal/httpx/violacion_test.go:12:93: not enough arguments in call to c.Pedir
	have (schema.Contexto, httpx.Peticion)
	want (context.Context, schema.Contexto, httpx.Peticion)
```

`rtk proxy make -C "$COPIA" ci` → **rc=2**, se detiene en `lint` con el mismo rechazo del compilador:

```
go tool -modfile=tools/golangci-lint/go.mod golangci-lint fmt --diff ./...
go tool -modfile=tools/golangci-lint/go.mod golangci-lint run ./...
internal/httpx/adaptador_test.go:1: : # github.com/jmorenobl/kitlegal/internal/httpx_test [github.com/jmorenobl/kitlegal/internal/httpx.test]
internal/httpx/violacion_test.go:12:36: not enough arguments in call to c.Pedir
	have (schema.Contexto, httpx.Peticion)
	want (context.Context, schema.Contexto, httpx.Peticion) (typecheck)
// adaptador_test.go es el adaptador de prueba del hito y la demostración de que
1 issues:
* typecheck: 1
make: *** [lint] Error 1
```

La garantía «contexto obligatorio» no depende de ningún analizador: la firma de `Pedir` la hace
imposible de esquivar, y el error nombra exactamente lo que falta, `context.Context` en primera
posición (FR-061, SC-013).

---

## 4. Variante d — R2 ya no pasa en vacío: sin dueño, el test de arquitectura falla

Tal como lo escribe el escenario 10.d, sobre la copia:

```bash
rm -r "$COPIA/internal/httpx" "$COPIA/cmd/kitlegal/main_test.go"
```

`rtk proxy go test -C "$COPIA" -count=1 ./internal/ -run TestArquitectura/R2` → **rc=1**:

```
--- FAIL: TestArquitectura (0.06s)
    --- FAIL: TestArquitectura/R2_·_solo_internal/httpx_importa_net/http (0.00s)
        arch_test.go:68: 
            	Error Trace:	…/copia-t017/internal/arch_test.go:353
            	            				…/copia-t017/internal/arch_test.go:317
            	            				…/copia-t017/internal/arch_test.go:68
            	Error:      	Should NOT be empty, but was []
            	Test:       	TestArquitectura/R2_·_solo_internal/httpx_importa_net/http
            	Messages:   	R2 · el grafo no contiene github.com/jmorenobl/kitlegal/internal/httpx, que es quien tiene que concentrar net/http: la regla quedaría activa y vacía, cumpliéndose porque no hay nada que vigilar
FAIL
FAIL	github.com/jmorenobl/kitlegal/internal	0.279s
FAIL
```

`rtk proxy make -C "$COPIA" ci` → **rc=2**. Aquí el analizador está limpio —no hay ninguna importación
prohibida, justamente porque el paquete que la concentraba ya no está— y el rojo lo pone `test`:

```
go tool -modfile=tools/golangci-lint/go.mod golangci-lint fmt --diff ./...
go tool -modfile=tools/golangci-lint/go.mod golangci-lint run ./...
0 issues.
go test -race -shuffle=on -coverprofile=coverage.out ./...
	github.com/jmorenobl/kitlegal/cmd/kitlegal		coverage: 0.0% of statements
-test.shuffle 1789200191488769000
--- FAIL: TestArquitectura (0.08s)
    --- FAIL: TestArquitectura/R2_·_solo_internal/httpx_importa_net/http (0.00s)
        arch_test.go:68: 
            	Error Trace:	…/copia-t017/internal/arch_test.go:353
            	            				…/copia-t017/internal/arch_test.go:317
            	            				…/copia-t017/internal/arch_test.go:68
            	Error:      	Should NOT be empty, but was []
            	Test:       	TestArquitectura/R2_·_solo_internal/httpx_importa_net/http
            	Messages:   	R2 · el grafo no contiene github.com/jmorenobl/kitlegal/internal/httpx, que es quien tiene que concentrar net/http: la regla quedaría activa y vacía, cumpliéndose porque no hay nada que vigilar
FAIL
coverage: [no statements]
FAIL	github.com/jmorenobl/kitlegal/internal	0.378s
ok  	github.com/jmorenobl/kitlegal/internal/app	2.793s	coverage: 92.8% of statements
	github.com/jmorenobl/kitlegal/internal/app/ejemplo		coverage: 0.0% of statements
	github.com/jmorenobl/kitlegal/internal/app/ejemplo/kitlegal-e2e		coverage: 0.0% of statements
ok  	github.com/jmorenobl/kitlegal/internal/cli	1.672s	coverage: 98.6% of statements
ok  	github.com/jmorenobl/kitlegal/internal/core/schema	2.450s	coverage: 90.1% of statements
ok  	github.com/jmorenobl/kitlegal/internal/render	2.153s	coverage: 95.8% of statements
FAIL
make: *** [test] Error 1
```

Esto es exactamente lo que T016 endureció: la subprueba deja de cumplirse por no tener nada que vigilar
(FR-053, SC-010, D18).

---

## 5. El árbol de trabajo no se tocó

`git status --porcelain` antes y después del escenario, idéntico —solo los dos ficheros de estado que
el propio workflow escribe—, y la copia desechable borrada al terminar:

```
 M specs/003-h2-internal-httpx-cliente/gates/tarea-actual.json
 M specs/003-h2-internal-httpx-cliente/gates/tareas-intentos.json
```

**Árbol de trabajo intacto.**

---

## 6. Resumen y corrección aplicada al escenario

| # | Intento | Regla que lo rechaza | Dónde se nombra | `make ci` |
|---|---|---|---|---|
| a | `*http.Response` sin cerrar | `bodyclose` | `lint` | rc=2 |
| a | `net/http` en un paquete del kernel | `depguard` (lista `red`, R2) y `TestArquitectura/R2` | `lint` y `test` | rc=2 |
| b | `http.NewRequest` sin contexto | `gofumpt`/`goimports` en una línea; `noctx` ya formateado | `fmt-check` y `lint` | rc=2 |
| c | `Pedir` sin contexto | el compilador (`typecheck`) | `lint` | rc=2 |
| d | R2 sin dueño | `TestArquitectura/R2` (`exigeDueno`) | `test` | rc=2 |

Los cuatro intentos terminan en rojo y ninguno necesita disciplina humana para detectarse: **SC-012
cumplido**.

Dos correcciones a `quickstart.md`, escenario 10, que este intento obliga a hacer y que **no tocan
ninguna regla, ningún linter ni ninguna exclusión** —el `.golangci.yml` no cambia—:

1. **10.b se escribe con el cuerpo en su propia línea.** La versión de una sola línea muere antes, en
   `fmt-check`, y cuando se llega a `lint` la deduplicación por línea deja solo `unused`: el `grep -F
   'net/http.NewRequest must not be called'` del escenario no encontraba nada y el escenario no
   demostraba lo que dice demostrar. Con el cuerpo en su línea, `noctx` nombra la regla y el escenario
   vuelve a ser una demostración.
2. **El «Esperado» de 10.a deja de prometer una línea de `noctx`**, que `--uniq-by-line` tapa tras
   `bodyclose` por compartir posición. El intento se rechaza igual, y por partida triple.

`--uniq-by-line` es un asunto de presentación del informe, no del veredicto: golangci-lint sale con 1
cuente los hallazgos que cuente. Por eso no se cambia la configuración.
