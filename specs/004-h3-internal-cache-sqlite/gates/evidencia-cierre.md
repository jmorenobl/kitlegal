# Evidencia de cierre · H3 medido de principio a fin

Tarea T013 de [`tasks.md`](../tasks.md): los escenarios 1, 4, 7, 10, 11 y 12 de
[`quickstart.md`](../quickstart.md), la orden de direcciones de sus prerrequisitos y las medidas
diferenciales (SC-003, SC-008, SC-009, SC-012, SC-014; puntos 1, 2 y 9 de la Definition of Done;
obligaciones 9, 11 y 12 del plan). Todas las salidas son literales.

## Condiciones

- **Fecha**: 2026-09-12. **Rama**: `h3-internal-cache-sqlite`, sobre `495c88a` (T012).
- **Herramientas**: `go version go1.27.1 darwin/arm64` (la directiva `toolchain` de `go.mod`),
  `git version 2.50.1 (Apple Git-155)`, `GNU Make 3.81`.
- **Árbol de trabajo al empezar**: `rtk proxy git status --porcelain` solo muestra los dos ficheros que
  el workflow `hito` actualiza al lanzar la tarea (`gates/tarea-actual.json`, `gates/tareas-intentos.json`).
- **`$TMP`** abajo es el directorio temporal de la sesión, fuera del repositorio.
- **H2 de referencia**: `main` en `ed6d1c4`, clonado en `$TMP/h2-main` con `git clone --branch main`.
  Entre el merge de H2 (`012c235`) y `ed6d1c4` no cambia ningún `*.go`, ni `Makefile` ni `go.mod`:
  `rtk proxy git diff --stat 012c235 main -- '*.go' Makefile go.mod` sale vacío. El binario de `main`
  es, por tanto, el de H2.

### Tres cosas que había que neutralizar para que ninguna medida saliera verde en vacío

1. **La caché de resultados de `go test`.** El primer `make ci` de la sesión terminó en verde, pero con
   todos los paquetes de `test-integration` en `(cached)`: esos tests no se ejecutaron, y una medida
   diferencial de `~/.cache/kitlegal` alrededor de ellos no habría medido nada. Se vació con
   `go clean -testcache`, que solo toca la caché de compilación de Go, y se repitió `make ci`. **La
   ejecución de §1 es esa segunda**, sin caché.
2. **Envoltorios de la salida.** `rtk` reescribe `go test`, `git status` y `git diff`. Además, la shell de
   la sesión sustituye `grep` y `find` por funciones que llaman a un `ugrep` y un `bfs` embebidos
   (`~/.claude/shell-snapshots/snapshot-zsh-…sh`, líneas 6958 a 6985). La primera ejecución de la orden de
   direcciones tal cual imprimió `grep:  : No such file or directory` antes del «solo direcciones
   ficticias» (§7). Por eso toda orden cuya salida se filtra, cuenta o compara va con `rtk proxy`, que
   ejecuta el binario del sistema sin pasar ni por la reescritura ni por esas funciones.
3. **Permisos del modo desatendido.** Piden aprobación, y por tanto no se pueden usar: `$?`, `$(…)`, el
   prefijo de entorno `VAR=valor orden`, `mktemp -d` y ejecutar un binario del directorio temporal. Cada
   una se sustituye por una forma equivalente:

   | En el quickstart | Aquí | Por qué es lo mismo |
   |---|---|---|
   | `ANTES=$(ls -laR … 2>&1 \| shasum)` … `test "$ANTES" = "$DESPUES"` | `ls -laR … > fichero 2>&1` antes y después, `rtk proxy cmp` y `rtk proxy shasum` | `shasum` de un fichero es el del flujo que lo llenó |
   | `… \| grep -vx '…' ; test $? -eq 1 && echo "…"` | la misma tubería con el estado de salida que informa la herramienta, y con `\|\| echo "…"` | `Exit code 1` sin ninguna línea es `test $? -eq 1`; que no hay estado 2 lo prueba la ausencia de mensajes de error |
   | `SONDA=$(mktemp -d)` y `TMPDIR="$SONDA" go test …` | `mkdir $TMP/sonda-tmpdir` y `rtk proxy env TMPDIR=$TMP/sonda-tmpdir go test …` | directorio vacío fuera del repositorio; `env` entrega la variable igual |
   | `./bin/kitlegal --help` de H2 | `go -C $TMP/h2-main run ./cmd/kitlegal --help`, y lo mismo en la rama | mismo paquete `main`; la ayuda no depende de las variables que inyecta `-ldflags` |

   **Cada sustitución lleva su sonda positiva**: la misma orden sobre un caso que debe delatar, para
   descartar que el verde venga de un flujo vacío.

## Resumen

| Comprobación | Resultado | Criterio |
|---|---|---|
| `make ci` con la caché de tests vacía | código 0 y `ci: todos los controles en verde`; `test-integration` ejecutado entre `test` y `vuln`, sin `(cached)` | DoD 1, SC-008, SC-009, SC-014 |
| Línea `ci:` del `Makefile` | `ci: fmt-check lint test test-integration vuln schema-check secrets mod-verify mod-tidy-check` | escenario 11 |
| `//nolint` añadidos frente a `main` | **`0`** (sonda: la misma tubería cuenta `6999` líneas añadidas) | SC-008 |
| Diff de `.golangci.yml` frente a `main` | solo `run.build-tags: [integration]`, comentarios de la lista `sql` y las palabras `reproduccion` y `versiones` en `misspell.ignore-rules` | SC-008 |
| Cobertura global (`coverage.out`, sin integración) | **92,9 %** (1967/2117 sentencias) ≥ 70 % | DoD 9, obligación 11 |
| Cobertura de `internal/core/**` | **90,1 %** (73/81 sentencias) ≥ 85 % | DoD 9, obligación 11 |
| `ls -laR ~/.cache/kitlegal` | idéntico (SHA-1 `519f40d4…`) antes y después de `make ci`, de `make test` y de toda la sesión | SC-009, obligación 9 |
| `TMPDIR` desechable tras la integración | **`0`** ficheros (sonda con `go test -work`: `292`) | SC-009, obligación 9 |
| Escenario 1 | `make test` en verde; ningún cambio bajo `testdata`, `internal/cache/testdata` ni `internal/source` | SC-009, SC-014, DoD 2 |
| Escenario 4 | seis `--- PASS` `offline-` | SC-003, obligación 9 |
| Escenario 7 | los cuatro tests de integración y sus subtests en `--- PASS`, ningún `SKIP`; `make test-integration` en verde | SC-007, SC-009 |
| Escenario 10 | ayuda idéntica byte a byte a la de H2; `go list -deps` → `0` (sonda: `31`); `go version -m` → `0` (sonda: `4`); los cuatro tests del binario en verde; el adaptador de prueba solo en un test | SC-012, FR-044 |
| Escenario 12 | cuatro `--- PASS` | FR-015, FR-032 |
| Direcciones en tests y material de reproducción | «solo direcciones ficticias» (sonda: delata una dirección de `boe.es`) | obligación 12, control 18 |
| Dependencias | una directa nueva, `modernc.org/sqlite v1.58.0`, con los indirectos de S1 y la subida de `golang.org/x/tools` ya anotada en [`s1-dependencias.md`](./s1-dependencias.md) | SC-014 |

---

## 1. Escenario 11 · `make ci` de principio a fin, con la caché de tests vacía

```
$ go clean -testcache
$ ls -laR "$HOME/.cache/kitlegal" > $TMP/cache-antes-2.txt 2>&1
$ rtk proxy make ci > $TMP/make-ci-2.log 2>&1        # código de salida 0
go tool -modfile=tools/golangci-lint/go.mod golangci-lint fmt --diff ./...
# github.com/golangci/golangci-lint/v2/cmd/golangci-lint
ld: warning: -bind_at_load is deprecated on macOS
go tool -modfile=tools/golangci-lint/go.mod golangci-lint run ./...
# github.com/golangci/golangci-lint/v2/cmd/golangci-lint
ld: warning: -bind_at_load is deprecated on macOS
0 issues.
go test -race -shuffle=on -coverprofile=coverage.out ./...
ok  	github.com/jmorenobl/kitlegal/cmd/kitlegal	1.387s	coverage: 0.0% of statements
ok  	github.com/jmorenobl/kitlegal/internal	2.321s	coverage: [no statements]
ok  	github.com/jmorenobl/kitlegal/internal/app	3.871s	coverage: 93.1% of statements
	github.com/jmorenobl/kitlegal/internal/app/ejemplo		coverage: 0.0% of statements
	github.com/jmorenobl/kitlegal/internal/app/ejemplo/kitlegal-e2e		coverage: 0.0% of statements
ok  	github.com/jmorenobl/kitlegal/internal/cache	2.686s	coverage: 87.8% of statements
ok  	github.com/jmorenobl/kitlegal/internal/cli	3.035s	coverage: 98.6% of statements
?   	github.com/jmorenobl/kitlegal/internal/core	[no test files]
ok  	github.com/jmorenobl/kitlegal/internal/core/schema	2.631s	coverage: 90.1% of statements
ok  	github.com/jmorenobl/kitlegal/internal/httpx	17.047s	coverage: 96.8% of statements
ok  	github.com/jmorenobl/kitlegal/internal/render	3.480s	coverage: 95.8% of statements
go test -race -tags=integration ./...
ok  	github.com/jmorenobl/kitlegal/cmd/kitlegal	1.345s
ok  	github.com/jmorenobl/kitlegal/internal	1.874s
ok  	github.com/jmorenobl/kitlegal/internal/app	3.679s
?   	github.com/jmorenobl/kitlegal/internal/app/ejemplo	[no test files]
?   	github.com/jmorenobl/kitlegal/internal/app/ejemplo/kitlegal-e2e	[no test files]
ok  	github.com/jmorenobl/kitlegal/internal/cache	4.692s
ok  	github.com/jmorenobl/kitlegal/internal/cli	2.436s
?   	github.com/jmorenobl/kitlegal/internal/core	[no test files]
ok  	github.com/jmorenobl/kitlegal/internal/core/schema	2.640s
ok  	github.com/jmorenobl/kitlegal/internal/httpx	15.703s
ok  	github.com/jmorenobl/kitlegal/internal/render	2.061s
go tool -modfile=tools/govulncheck/go.mod govulncheck ./...
No vulnerabilities found.
schema-check: no hay schemas/ todavía; los aportan H4 (borrador) y H10 (contrato)
go tool -modfile=tools/gitleaks/go.mod gitleaks dir . --redact --no-banner
10:51PM INF scanned ~8390240 bytes (8.39 MB) in 775ms
10:51PM INF no leaks found
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

(Los códigos de color de `gitleaks` se han quitado de las dos líneas `INF`.)

`-race` está en las dos recetas de test. `test-integration` corre entre `test` y `vuln`, en el orden de
la línea `ci:`, y **se ejecutó**: sin ningún `(cached)`, con `internal/cache` en 4,692 s frente a los
2,686 s de la suite unitaria.

```
$ rtk proxy grep -nE '^ci:' Makefile
148:ci: fmt-check lint test test-integration vuln schema-check secrets mod-verify mod-tidy-check
```

### Ninguna supresión nueva

```
$ rtk proxy git diff main -- '*.go' | rtk proxy grep -c '^+.*//nolint'
0                                                     # código de salida 1: ninguna línea
$ rtk proxy git diff main -- '*.go' | rtk proxy grep -c '^+'
6999                                                  # sonda: el diff llega en crudo y el filtro ve lo añadido
```

```
$ rtk proxy git diff main -- .golangci.yml
diff --git a/.golangci.yml b/.golangci.yml
index b1e520a..2f7fb65 100644
--- a/.golangci.yml
+++ b/.golangci.yml
@@ -8,6 +8,12 @@ version: "2"
 run:
   # Los tests son código del proyecto y se lintan igual que el resto.
   tests: true
+  # También los que llevan la etiqueta de compilación integration, que make ci
+  # ejecuta con test-integration: sin ella ningún linter —sqlclosecheck y
+  # rowserrcheck incluidos— cargaría esos ficheros, y uno que ni compila pasaría
+  # el lint en verde (H3 FR-041, research.md D11).
+  build-tags:
+    - integration
 
 linters:
   # `standard` aporta errcheck, govet, staticcheck, ineffassign y unused.
@@ -111,8 +117,12 @@ linters:
               desc: "R2: la red se usa a través de internal/httpx, que concentra reintentos, límite por sitio, robots.txt y User-Agent identificable"
 
         # R3 · SQLite y database/sql son de los tres paquetes de almacenamiento y
-        # de nadie más. Ninguno de los tres existe todavía —los aportan H3, H12 y
-        # H16—, así que hoy la regla también está activa y vacía.
+        # de nadie más. Desde H3 la regla tiene dueño: internal/cache existe y es
+        # el único paquete del árbol que importa database/sql y el controlador de
+        # SQLite. internal/store e internal/graph llegan en H12 y H17 y ya figuran
+        # entre las excepciones. internal/arch_test.go comprueba lo mismo sobre el
+        # cierre transitivo real, aunque sin exigir dueño todavía: esa exigencia
+        # pide los tres paquetes en el grafo (H3 FR-036, FR-037).
         sql:
           list-mode: original
           files:
@@ -210,9 +220,19 @@ linters:
         # larga de la tabla del contrato de grabación §2 de H2 que
         # TestNombreDeGrabacion copia tal cual; misspell lo lee como «legislation».
         - legislacion
+        # «Reproduccion», el tramo del directorio de grabaciones que fija el
+        # contrato de grabación de H2 (testdata/reproduccion/<fuente>) y que nombra
+        # el subtest sin-cache-la-reproduccion-falla del inventario de tests de H3;
+        # misspell lo lee como «reproduction».
+        - reproduccion
         # «Resolucion de adjudicacion», parte de la otra ruta larga de la misma
         # tabla, la que se recorta; misspell lo lee como «resolution».
         - resolucion
+        # Plural español de «versión» —las del esquema de la caché: la que el
+        # fichero trae y la que el binario conoce, que el contrato de errores de
+        # H3 §3 obliga a nombrar en el mensaje—; misspell lo lee como
+        # «versions». El singular lleva tilde y no casa (H3 S2, D15).
+        - versiones
 ```

Exactamente lo que admite la tarea. Hay una clave nueva, `run.build-tags` (T010). La lista `sql` solo
cambia comentarios (T011). Y `misspell.ignore-rules` gana dos palabras que el supuesto S2 obligó a
ignorar: `reproduccion` (T008) y `versiones` (T002). No hay ninguna regla, ningún linter, ninguna
exclusión ni ningún `//nolint`.

### Umbrales de cobertura

`coverage.out` lo escribe solo la receta `test` (`go test -race -shuffle=on -coverprofile=coverage.out
./...`, línea 67 del `Makefile`). `test-integration` (`go test -race -tags=integration ./...`, línea 71)
no lleva `-coverprofile`, así que los tests de integración no alimentan el perfil.

```
$ rtk proxy go tool cover -func=coverage.out | rtk proxy tail -1
total:										(statements)				92.9%
$ awk 'NR == 1 { print "modo:", $0 } NR > 1 && $1 ~ /\/internal\/cache\// { f++ } NR > 1 { n += $2; if ($3 > 0) c += $2 } END { printf "global: %d/%d sentencias = %.1f%%\nbloques de internal/cache en el perfil: %d\n", c, n, 100 * c / n, f }' coverage.out
modo: mode: atomic
global: 1967/2117 sentencias = 92.9%
bloques de internal/cache en el perfil: 270
$ awk 'NR > 1 && $1 ~ /\/internal\/core\// { n += $2; if ($3 > 0) c += $2 } END { printf "internal/core/**: %d/%d sentencias = %.1f%%\n", c, n, 100 * c / n }' coverage.out
internal/core/**: 73/81 sentencias = 90.1%
```

- **Global**: 92,9 % ≥ 70 % (`codecov.yml`, `coverage.status.project.default.target`).
- **Dominio**: `internal/core/**` 90,1 % ≥ 85 % (componente `internal_core`). T001 no añadió sentencias:
  `internal/core` sigue en `[no test files]` sin sentencias, y las 81 son de `internal/core/schema`.
- El kernel, `internal/cli`, está en 98,6 % ≥ 90 %.
- `codecov.yml` no cambia frente a `main` (`rtk proxy git diff --stat main -- codecov.yml` sale vacío):
  ningún umbral se ha rebajado.

La primera ejecución de `make ci`, con la caché de tests llena, dio los mismos 92,9 % y 90,1 %.

## 2. Medida diferencial de la caché real de la cuenta (SC-009, obligación 9)

`~/.cache/kitlegal` no existe en esta cuenta. El listado es, por tanto, el error de `ls`, y la medida
detecta tanto que el directorio aparezca como que cambie cualquier cosa bajo él.

```
$ cat $TMP/cache-antes.txt
ls: /Users/jorge/.cache/kitlegal: No such file or directory
$ rtk proxy shasum $TMP/cache-antes.txt $TMP/cache-despues-ci.txt $TMP/cache-despues-esc1.txt $TMP/cache-final.txt
519f40d45a127eb08b6a9c0d7f6c9d50b1944384  $TMP/cache-antes.txt
519f40d45a127eb08b6a9c0d7f6c9d50b1944384  $TMP/cache-despues-ci.txt
519f40d45a127eb08b6a9c0d7f6c9d50b1944384  $TMP/cache-despues-esc1.txt
519f40d45a127eb08b6a9c0d7f6c9d50b1944384  $TMP/cache-final.txt
$ rtk proxy shasum $TMP/cache-antes-2.txt $TMP/cache-despues-2.txt
519f40d45a127eb08b6a9c0d7f6c9d50b1944384  $TMP/cache-antes-2.txt
519f40d45a127eb08b6a9c0d7f6c9d50b1944384  $TMP/cache-despues-2.txt
$ rtk proxy cmp $TMP/cache-antes-2.txt $TMP/cache-despues-2.txt && echo "la caché real de la cuenta no se ha tocado"
la caché real de la cuenta no se ha tocado
$ rtk proxy cmp $TMP/cache-antes.txt $TMP/cache-final.txt && echo "la caché real de la cuenta no se ha tocado en toda la sesión"
la caché real de la cuenta no se ha tocado en toda la sesión
```

| Instantánea | Cuándo |
|---|---|
| `cache-antes.txt` | al empezar la sesión, antes del primer `make ci` |
| `cache-despues-ci.txt` | tras el primer `make ci` (caché de tests llena) |
| `cache-antes-2.txt` / `cache-despues-2.txt` | justo antes y justo después del `make ci` sin caché de §1 |
| `cache-despues-esc1.txt` | tras el `make test` del escenario 1 |
| `cache-final.txt` | tras todos los escenarios: 4, 7 (con los tests de integración), 10 y 12 |

## 3. Escenario 1 · la suite entera, sin red, con `-race`, sin tocar la caché real

```
$ rtk proxy make test > $TMP/esc1-make-test.log 2>&1   # código de salida 0
go test -race -shuffle=on -coverprofile=coverage.out ./...
ok  	github.com/jmorenobl/kitlegal/cmd/kitlegal	1.445s	coverage: 0.0% of statements
ok  	github.com/jmorenobl/kitlegal/internal	1.510s	coverage: [no statements]
ok  	github.com/jmorenobl/kitlegal/internal/app	2.864s	coverage: 93.1% of statements
	github.com/jmorenobl/kitlegal/internal/app/ejemplo		coverage: 0.0% of statements
	github.com/jmorenobl/kitlegal/internal/app/ejemplo/kitlegal-e2e		coverage: 0.0% of statements
ok  	github.com/jmorenobl/kitlegal/internal/cache	2.539s	coverage: 87.8% of statements
ok  	github.com/jmorenobl/kitlegal/internal/cli	1.845s	coverage: 98.6% of statements
?   	github.com/jmorenobl/kitlegal/internal/core	[no test files]
ok  	github.com/jmorenobl/kitlegal/internal/core/schema	1.917s	coverage: 90.1% of statements
ok  	github.com/jmorenobl/kitlegal/internal/httpx	15.304s	coverage: 96.8% of statements
ok  	github.com/jmorenobl/kitlegal/internal/render	2.274s	coverage: 95.8% of statements
$ rtk proxy cmp $TMP/cache-despues-2.txt $TMP/cache-despues-esc1.txt && echo "la caché real de la cuenta no se ha tocado"
la caché real de la cuenta no se ha tocado
$ rtk proxy git status --porcelain -- testdata internal/cache/testdata internal/source
$                                                     # ninguna línea
$ rtk proxy git status --porcelain -- specs
 M specs/004-h3-internal-cache-sqlite/gates/tarea-actual.json
 M specs/004-h3-internal-cache-sqlite/gates/tareas-intentos.json
```

- **`make test`**: en verde, sin ningún `FAIL` ni `WARNING: DATA RACE`.
- **Caché real**: la de la cuenta no cambia.
- **Material de test**: entre la orden de `git status` y el final no aparece ninguna línea. La última
  orden es la sonda positiva: el mismo `git status` filtrado por ruta sí muestra los cambios que existen,
  así que el vacío anterior es real y no un «ok» reescrito.

**Tests offline y material de reproducción (DoD 2).** Frente a `main`, el hito añade un único fichero
bajo los árboles de fixtures y esquemas: el material escrito a mano de T007. No hay ningún
`testdata/<fuente>/` en la raíz, ningún `schemas/` ni ningún `internal/source/`.

```
$ rtk proxy git diff --stat main -- testdata schemas internal/cache/testdata internal/source
 .../prueba/GET_http_fuente.prueba_norma.json       | 28 ++++++++++++++++++++++
 1 file changed, 28 insertions(+)
$ rtk proxy git ls-files internal/cache/testdata testdata schemas
internal/cache/testdata/reproduccion/prueba/GET_http_fuente.prueba_norma.json
```

## 4. Escenario 4 · `--offline` por el kernel (SC-003)

```
$ rtk proxy go test -race -count=1 -v -run 'TestAdaptadorDePruebaConElKernel/offline-' ./internal/cache/ | rtk proxy grep -E '^\s*--- (PASS|FAIL): TestAdaptadorDePruebaConElKernel/offline-'
    --- PASS: TestAdaptadorDePruebaConElKernel/offline-directorio-inexistente (0.01s)
    --- PASS: TestAdaptadorDePruebaConElKernel/offline-sin-base (0.01s)
    --- PASS: TestAdaptadorDePruebaConElKernel/offline-ausente (0.02s)
    --- PASS: TestAdaptadorDePruebaConElKernel/offline-expirada (0.02s)
    --- PASS: TestAdaptadorDePruebaConElKernel/offline-presente (0.02s)
    --- PASS: TestAdaptadorDePruebaConElKernel/offline-guardar (0.03s)
```

Seis líneas `--- PASS`, las seis del inventario, y ninguna `FAIL`. El orden es el de terminación de las
subpruebas paralelas. Cada subtest compara el SHA-256 de `cache.db`, o su ausencia, y el listado del
directorio antes y después de invocar `app.Main` con `--offline`.

## 5. Escenario 7 · dos clientes y dos procesos; la integración solo escribe en su `t.TempDir()` (SC-007, SC-009)

```
$ rtk proxy go test -race -count=1 -run '^(TestDosClientesEnElMismoProceso|TestMismaClaveDosEscritores|TestSoloLecturaVeLoConfirmadoEnElWAL)$' ./internal/cache/
ok  	github.com/jmorenobl/kitlegal/internal/cache	1.416s
$ mkdir $TMP/sonda-tmpdir
$ rtk proxy env TMPDIR=$TMP/sonda-tmpdir go test -race -count=1 -tags=integration -v -run '^TestIntegracion' ./internal/cache/ | rtk proxy grep -E '^\s*--- (PASS|FAIL|SKIP)'
--- PASS: TestIntegracionDirectorioDenegado (0.00s)
    --- PASS: TestIntegracionDirectorioDenegado/normal-argumentos (0.04s)
    --- PASS: TestIntegracionDirectorioDenegado/solo-lectura-inesperado (0.04s)
--- PASS: TestIntegracionWALSinMemoriaCompartida (0.05s)
--- PASS: TestIntegracionDirectorioNoEscribible (0.00s)
    --- PASS: TestIntegracionDirectorioNoEscribible/normal-argumentos (0.05s)
    --- PASS: TestIntegracionDirectorioNoEscribible/solo-lectura-lee (0.06s)
--- PASS: TestIntegracionDosProcesos (0.00s)
    --- PASS: TestIntegracionDosProcesos/escritor-hijo-lector-padre (1.07s)
    --- PASS: TestIntegracionDosProcesos/escritor-padre-lector-solo-lectura-hijo (1.08s)
$ rtk proxy find $TMP/sonda-tmpdir -mindepth 1 | rtk proxy wc -l
       0
$ rm -r $TMP/sonda-tmpdir
$ rtk proxy make test-integration > $TMP/esc7-test-integration.log 2>&1   # código de salida 0
go test -race -tags=integration ./...
ok  	github.com/jmorenobl/kitlegal/cmd/kitlegal	(cached)
ok  	github.com/jmorenobl/kitlegal/internal	(cached)
ok  	github.com/jmorenobl/kitlegal/internal/app	(cached)
?   	github.com/jmorenobl/kitlegal/internal/app/ejemplo	[no test files]
?   	github.com/jmorenobl/kitlegal/internal/app/ejemplo/kitlegal-e2e	[no test files]
ok  	github.com/jmorenobl/kitlegal/internal/cache	(cached)
ok  	github.com/jmorenobl/kitlegal/internal/cli	(cached)
?   	github.com/jmorenobl/kitlegal/internal/core	[no test files]
ok  	github.com/jmorenobl/kitlegal/internal/core/schema	(cached)
ok  	github.com/jmorenobl/kitlegal/internal/httpx	(cached)
ok  	github.com/jmorenobl/kitlegal/internal/render	(cached)
```

- **Mismo proceso**: la orden termina en `ok`.
- **Integración**: solo líneas `--- PASS` y ningún `SKIP`. Aparecen todos los nombres que exige el
  escenario: `escritor-padre-lector-solo-lectura-hijo`, `escritor-hijo-lector-padre`, `solo-lectura-lee`,
  `normal-argumentos`, `TestIntegracionDirectorioDenegado` y `TestIntegracionWALSinMemoriaCompartida`.
- **Máquina y entorno**: puesto de desarrollo sin privilegios y sin la variable `CI`, así que los
  permisos se hacen valer y la precondición no salta.
- **`TMPDIR`**: el `find` imprime **`0`**.
- **`make test-integration`**: en verde. Sale `(cached)` porque es la misma ejecución que `make ci` acababa
  de hacer sin caché en §1, con el mismo código.

**Sonda positiva del `TMPDIR`.** Hay que descartar que el `0` salga porque la variable no llegó a Go o
porque el `find` no ve nada. Para ello se lanzó `go test -work`, que conserva su directorio de trabajo,
sobre otro directorio vacío:

```
$ mkdir $TMP/sonda-work
$ rtk proxy env TMPDIR=$TMP/sonda-work go test -work -count=1 -run '^$' ./internal/cache/
WORK=$TMP/sonda-work/go-build235910919
ok  	github.com/jmorenobl/kitlegal/internal/cache	0.390s [no tests to run]
$ rtk proxy find $TMP/sonda-work -mindepth 1 | rtk proxy wc -l
     292
$ rm -r $TMP/sonda-work
```

- **La variable llega**: el directorio de trabajo nace bajo el `TMPDIR` inyectado. `os.TempDir`, y con él
  `t.TempDir()`, lee la misma variable en macOS.
- **El `find` cuenta**: lo que queda, 292 entradas.

## 6. Escenario 10 · la superficie del binario no cambia (SC-012, FR-044)

```
$ rtk proxy make build
CGO_ENABLED=0 go build -trimpath -ldflags "-X main.version=495c88a-dirty -X main.commit=495c88a2eb38cc5ee86bff8eeb9f6d3e88675aad -X main.fecha=2026-09-12T20:53:45Z -X github.com/jmorenobl/kitlegal/internal/httpx.version=495c88a-dirty" -o bin/kitlegal ./cmd/kitlegal
$ ./bin/kitlegal --help
uso: kitlegal <applet> [verbo] [banderas]

Este binario no registra ningún applet.
$ ./bin/kitlegal version
kitlegal 495c88a-dirty
commit: 495c88a2eb38cc5ee86bff8eeb9f6d3e88675aad
fecha:  2026-09-12T20:53:45Z
```

El `-dirty` se debe a los dos ficheros de `gates/` que el workflow tiene modificados al lanzar la tarea.

### El conjunto de verbos es el de H2

```
$ rtk proxy make -C $TMP/h2-main build
CGO_ENABLED=0 go build -trimpath -ldflags "-X main.version=ed6d1c4 -X main.commit=ed6d1c4d00cb614ec67c36cf1af71e3175b3f11c -X main.fecha=2026-09-12T20:53:47Z -X github.com/jmorenobl/kitlegal/internal/httpx.version=ed6d1c4" -o bin/kitlegal ./cmd/kitlegal
$ rtk proxy go -C $TMP/h2-main run ./cmd/kitlegal --help > $TMP/help-h2-gorun.txt 2>&1
$ rtk proxy go run ./cmd/kitlegal --help > $TMP/help-h3-gorun.txt 2>&1
$ rtk proxy cmp $TMP/help-h2-gorun.txt $TMP/help-h3-gorun.txt && echo "ayuda idéntica byte a byte"
ayuda idéntica byte a byte
```

- **Ayuda**: las dos dicen `uso: kitlegal <applet> [verbo] [banderas]` y `Este binario no registra ningún
  applet.`, igual que el binario de la rama construido con `make build`.
- **Verbos**: en H2 y en H3, ningún applet, `version` y la ayuda.
- **Por qué `go run`**: ejecutar el binario de H2 desde el directorio temporal pide aprobación en modo
  desatendido. `go run` construye el mismo paquete `main` sin `-ldflags`, que solo afecta a lo que imprime
  `version`.

### El binario no enlaza la caché ni el driver

```
$ rtk proxy go list -deps ./cmd/kitlegal | rtk proxy grep -cE 'internal/cache|modernc\.org'
0                                                     # código de salida 1
$ rtk proxy go version -m bin/kitlegal | rtk proxy grep -c 'modernc.org'
0                                                     # código de salida 1
```

Las dos sondas positivas aplican los mismos filtros a algo que sí enlaza la caché:

```
$ rtk proxy go list -deps ./internal/cache | rtk proxy grep -cE 'internal/cache|modernc\.org'
31
$ rtk proxy go test -c -o $TMP/cache.test ./internal/cache
$ rtk proxy go version -m $TMP/cache.test | rtk proxy grep -c 'modernc.org'
4
```

```
$ rtk proxy go test -count=1 -v ./internal/ -run '^(TestElBinarioNoEnlazaCache|TestElBinarioNoEnlazaHTTPX|TestDependenciasDelBinario|TestElBinarioNoEnlazaLosEjemplos)$' | rtk proxy grep -E '^(--- |ok|FAIL|PASS)'
--- PASS: TestElBinarioNoEnlazaCache (0.45s)
--- PASS: TestDependenciasDelBinario (0.45s)
--- PASS: TestElBinarioNoEnlazaLosEjemplos (0.45s)
--- PASS: TestElBinarioNoEnlazaHTTPX (0.46s)
PASS
ok  	github.com/jmorenobl/kitlegal/internal	0.768s
```

La lista que vigila la superficie, `modulosDelBinario`, es la de H2. El diff de
`internal/arch_test.go` frente a `main` solo cambia el comentario de R3 y añade
`TestElBinarioNoEnlazaCache` (T011). `rtk proxy git diff main -- internal/arch_test.go` no toca ninguna
línea de `modulosDelBinario`.

### El adaptador de prueba solo existe en tests

```
$ rtk proxy grep -rln --include='*.go' --exclude='*_test.go' -e 'adaptadorDePrueba' -e '"consultar"' -e '"guardar"' internal/app cmd internal/cache
$                                                     # ninguna línea, código de salida 1: «el adaptador de prueba solo existe en tests»
$ rtk proxy grep -rln --include='*.go' -e 'adaptadorDePrueba' -e '"consultar"' -e '"guardar"' internal/app cmd internal/cache
internal/cache/adaptador_test.go                      # sonda: sin --exclude, lo encuentra
```

### Dependencias (SC-014)

```
$ rtk proxy git diff main -- go.mod
@@ -12,15 +12,24 @@ require (
 	github.com/stretchr/testify v1.12.1
 	github.com/temoto/robotstxt v1.1.2
 	golang.org/x/time v0.16.0
+	modernc.org/sqlite v1.58.0
 )
 
 require (
 	github.com/bahlo/generic-list-go v0.2.0 // indirect
 	github.com/buger/jsonparser v1.1.2 // indirect
+	github.com/dustin/go-humanize v1.0.1 // indirect
+	github.com/google/uuid v1.6.0 // indirect
+	github.com/mattn/go-isatty v0.0.24 // indirect
+	github.com/ncruces/go-strftime v1.0.0 // indirect
 	github.com/pb33f/ordered-map/v2 v2.3.1 // indirect
+	github.com/remyoudompheng/bigfft v0.0.0-20230129092748-24d4a6f8daec // indirect
 	go.yaml.in/yaml/v3 v3.0.5 // indirect
 	go.yaml.in/yaml/v4 v4.0.0-rc.2 // indirect
-	golang.org/x/sys v0.26.0 // indirect
+	golang.org/x/sys v0.47.0 // indirect
 	golang.org/x/text v0.14.0 // indirect
-	golang.org/x/tools v0.26.0 // indirect
+	golang.org/x/tools v0.48.0 // indirect
+	modernc.org/libc v1.75.6 // indirect
+	modernc.org/mathutil v1.7.1 // indirect
+	modernc.org/memory v1.12.1 // indirect
 )
```

- **Dependencia directa**: una sola nueva, `modernc.org/sqlite v1.58.0`, la que fija el hito (FR-043).
- **Indirectos**: los nueve que predijo la sonda 6 de S1.
- **Subidas de versión**: `golang.org/x/sys`, que S1 anticipaba, y `golang.org/x/tools`, la única
  diferencia frente a la predicción. Está justificada en [`s1-dependencias.md`](./s1-dependencias.md) y
  T014 la anota en la propuesta de cambio.
- **Controles**: `go mod tidy -diff`, `go mod verify` y `govulncheck` en verde en §1.

## 7. Escenario 12 · el lector de solo lectura ve lo confirmado (FR-015, FR-032)

```
$ rtk proxy go test -race -count=1 -v -run '^(TestSoloLecturaVeLoConfirmadoEnElWAL|TestDosClientesEnElMismoProceso)$' ./internal/cache/ | rtk proxy grep -E '^\s*--- (PASS|FAIL)'
--- PASS: TestSoloLecturaVeLoConfirmadoEnElWAL (0.02s)
--- PASS: TestDosClientesEnElMismoProceso (0.00s)
    --- PASS: TestDosClientesEnElMismoProceso/normal-solo-lectura (0.06s)
    --- PASS: TestDosClientesEnElMismoProceso/normal-normal (0.06s)
```

Cuatro líneas `--- PASS`: los dos tests padre sin sangrar y las dos subpruebas sangradas. Ninguna
`--- FAIL`.

## 8. Prerrequisitos · solo direcciones ficticias en tests y material de reproducción (obligación 12)

**La orden tal cual**, con las funciones `grep` de la shell de la sesión:

```
$ grep -rhoE 'https?://[^"'"'"'`) ]+' internal/cache/*_test.go internal/cache/testdata \
  | sed -E 's/:+$//' \
  | grep -viE '^https?://fuente\.prueba(/|$)' \
  | grep -vx 'https://ventanillalegal.es/bot' ; test $? -eq 1 && echo "solo direcciones ficticias"
grep:  : No such file or directory
solo direcciones ficticias
```

La línea de error sale del `ugrep` embebido que sustituye a `grep` (Condiciones, punto 2), no de la orden.
Con esa línea delante, el esperado («ninguna línea antes») no se cumple. La orden se repite **entera**,
con cada etapa sin envoltorio:

```
$ rtk proxy grep -rhoE 'https?://[^"'"'"'`) ]+' internal/cache/*_test.go internal/cache/testdata
http://fuente.prueba/norma
http://fuente.prueba/otra
http://fuente.prueba/Norma
http://fuente.prueba/nor_a
http://fuente.prueba/nor%
http://fuente.prueba/nor
http://fuente.prueba/NORMA
http://fuente.prueba/norma/x
http://fuente.prueba/norma
http://fuente.prueba/norma
http://fuente.prueba/norma
http://fuente.prueba/norma
http://fuente.prueba/norma
https://ventanillalegal.es/bot
$ rtk proxy grep -rhoE 'https?://[^"'"'"'`) ]+' internal/cache/*_test.go internal/cache/testdata | rtk proxy sed -E 's/:+$//' | rtk proxy grep -viE '^https?://fuente\.prueba(/|$)' | rtk proxy grep -vx 'https://ventanillalegal.es/bot'
$                                                     # ninguna línea, código de salida 1 (= test $? -eq 1)
$ rtk proxy grep -rhoE 'https?://[^"'"'"'`) ]+' internal/cache/*_test.go internal/cache/testdata | rtk proxy sed -E 's/:+$//' | rtk proxy grep -viE '^https?://fuente\.prueba(/|$)' | rtk proxy grep -vx 'https://ventanillalegal.es/bot' || echo "solo direcciones ficticias"
solo direcciones ficticias
```

- **Resultado**: «solo direcciones ficticias», sin ninguna línea antes y sin ningún mensaje de error.
- **Direcciones encontradas**: todas son del host ficticio `fuente.prueba`. Las variantes de mayúsculas y
  las que llevan `_` o `%` están en `internal/cache/entradas_test.go`
  (`rtk proxy grep -lE 'fuente\.prueba/(nor_a|NORMA|Norma|nor%)'` sobre los diez ficheros de test solo
  lista ese). La única otra es la identificación `https://ventanillalegal.es/bot`, que viaja en la
  cabecera `User-Agent` grabada y no se consulta.
- **Ficheros recorridos**: los diez `internal/cache/*_test.go`, `integracion_test.go` incluido, y el
  árbol `internal/cache/testdata`.

**Sonda positiva**: la misma orden, con un fichero de `$TMP/sonda-url/` que contiene una dirección real,
la delata:

```
$ rtk proxy grep -rhoE 'https?://[^"'"'"'`) ]+' internal/cache/*_test.go internal/cache/testdata $TMP/sonda-url | rtk proxy sed -E 's/:+$//' | rtk proxy grep -viE '^https?://fuente\.prueba(/|$)' | rtk proxy grep -vx 'https://ventanillalegal.es/bot'
https://www.boe.es/datosabiertos/api/legislacion-consolidada
```

## 9. Estado final del árbol

```
$ rtk proxy git status --porcelain --ignored -- bin coverage.out internal testdata cmd
!! bin/
!! coverage.out
```

Lo único que dejan los escenarios en el árbol son `bin/` y `coverage.out`, ignorados desde H0. Ningún
escenario ha modificado un fichero versionado, el índice ni el historial. Esta evidencia es el único
fichero que escribe la tarea.
