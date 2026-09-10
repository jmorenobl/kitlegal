# Validación local de H0 — escenarios 1 a 10 y 12

**Tarea**: T013 · **Fecha**: 2026-09-10 · **Guía**: [`../quickstart.md`](../quickstart.md)

Registro de la ejecución local de los escenarios 1 a 10 y 12 de `quickstart.md`. El escenario 11 no se
ejecuta aquí: es de plataforma y lo cubre T014 en `verificacion-pr.md`.

## Entorno

| Dato | Valor |
|---|---|
| Máquina | darwin/arm64 |
| `go version` | `go version go1.26.6 darwin/arm64` |
| `git --version` | `git version 2.50.1 (Apple Git-155)` |
| Directiva `toolchain` de `go.mod` | `go1.26.6` |
| Revisión de partida (`git rev-parse HEAD`) | `615ae60352b14f317d53f13252598dd1b367ce31` |

La caché de construcción de Go ya tenía las cuatro herramientas de `tools/`, así que ninguna orden
pagó la compilación inicial descrita en los prerrequisitos de la guía.

## Veredicto por escenario

| Escenario | Qué comprueba | Criterios | Resultado |
|---|---|---|---|
| 1 | Veredicto reproducible en local | SC-001 | ✅ |
| 2 | `make ci` no modifica el árbol | SC-002 | ✅ |
| 3 | El formato es un gate, no una corrección | SC-002, SC-007 | ✅ |
| 4 | `kitlegal version` y códigos de salida | SC-003 | ✅ |
| 5 | Cobertura por encima del umbral | SC-011 | ✅ |
| 6 | Órdenes con el objeto ausente | SC-012 | ✅ |
| 7 | Prerrequisitos ausentes y pin de toolchain | SC-010 | ✅ |
| 8 | Ganchos de pre-commit | SC-007 | ✅ |
| 9 | Detección de secretos | SC-004, SC-009 | ✅ |
| 10 | Dependencias saneadas | SC-004, SC-009 | ✅ |
| 12 | Documentación fundacional | SC-008, SC-010 | ✅ |

---

## Escenario 1 — Veredicto reproducible en local (SC-001)

```bash
make build ; make test ; make lint ; make vuln
```

Salida relevante y códigos de salida:

```text
CGO_ENABLED=0 go build -trimpath -ldflags "-X main.version=615ae60-dirty -X main.commit=615ae60352b14f317d53f13252598dd1b367ce31 -X main.fecha=2026-09-10T19:41:04Z" -o bin/kitlegal ./cmd/kitlegal
EXIT_build=0
go test -race -shuffle=on -coverprofile=coverage.out ./...
ok  	github.com/jmorenobl/kitlegal/cmd/kitlegal	1.218s	coverage: 83.3% of statements
EXIT_test=0
go tool -modfile=tools/golangci-lint/go.mod golangci-lint run ./...
0 issues.
EXIT_lint=0
go tool -modfile=tools/govulncheck/go.mod govulncheck ./...
No vulnerabilities found.
EXIT_vuln=0
```

**✅** Las cuatro terminan con éxito, sin hallazgos. `make build` deja `bin/kitlegal`; `make test`
ejecuta con detector de carreras y deja `coverage.out`. El sufijo `-dirty` de la versión es esperado:
bajo el workflow `hito` el directorio del feature siempre tiene cambios sin confirmar.

El único aviso que aparece en la salida de las órdenes que usan `golangci-lint`
(`ld: warning: -bind_at_load is deprecated on macOS`) lo emite el enlazador del sistema al construir
la herramienta desde su módulo, no el control: no es un hallazgo y no altera ningún código de salida.

## Escenario 2 — La orden agregada (SC-002)

```bash
git status --porcelain > /tmp/antes.txt
make ci
git status --porcelain > /tmp/despues.txt
diff /tmp/antes.txt /tmp/despues.txt
```

```text
go tool -modfile=tools/golangci-lint/go.mod golangci-lint fmt --diff ./...
go tool -modfile=tools/golangci-lint/go.mod golangci-lint run ./...
0 issues.
go test -race -shuffle=on -coverprofile=coverage.out ./...
ok  	github.com/jmorenobl/kitlegal/cmd/kitlegal	1.186s	coverage: 83.3% of statements
go tool -modfile=tools/govulncheck/go.mod govulncheck ./...
No vulnerabilities found.
schema-check: no hay schemas/ todavía; los aportan H4 (borrador) y H11 (contrato)
go tool -modfile=tools/gitleaks/go.mod gitleaks dir . --redact --no-banner
	INF	no leaks found
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
EXIT_ci=0
EXIT_diff=0     # ficheros idénticos
```

**✅** Los ocho controles pasan y el `diff` de los dos `git status --porcelain` es vacío: la orden que
emite el veredicto no ha modificado nada (M1 del contrato de `make-targets.md`).

## Escenario 3 — El formato es un gate, no una corrección (FR-012)

Rojo → verde dentro del propio escenario. Mutación de espaciado sobre una línea existente, como exige
la guía:

```bash
sed -i.bak 's/^func main() {/func  main( )  {/' cmd/kitlegal/main.go && rm cmd/kitlegal/main.go.bak
git diff cmd/kitlegal/main.go
```

```diff
-func main() { os.Exit(run(os.Args[1:], os.Stdout, os.Stderr)) }
+func  main( )  { os.Exit(run(os.Args[1:], os.Stdout, os.Stderr)) }
```

**Rojo** — `make ci` falla en el primer control, `fmt-check`, nombrando el fichero:

```text
go tool -modfile=tools/golangci-lint/go.mod golangci-lint fmt --diff ./...
diff cmd/kitlegal/main.go.orig cmd/kitlegal/main.go
--- cmd/kitlegal/main.go.orig
+++ cmd/kitlegal/main.go
@@ -20,7 +20,7 @@
-func  main( )  { os.Exit(run(os.Args[1:], os.Stdout, os.Stderr)) }
+func main() { os.Exit(run(os.Args[1:], os.Stdout, os.Stderr)) }
make: *** [fmt-check] Error 1
EXIT_ci_1=2
```

Y **no** lo modifica — `git diff --stat cmd/kitlegal/main.go` sigue acusando la mutación:

```text
 cmd/kitlegal/main.go | 2 +-
 1 file changed, 1 insertion(+), 1 deletion(-)
```

**Verde** — `make fmt` sí lo corrige (`git diff cmd/kitlegal/main.go` queda vacío) y el segundo
`make ci` vuelve a pasar:

```text
go tool -modfile=tools/golangci-lint/go.mod golangci-lint fmt ./...
EXIT_fmt=0
...
ci: todos los controles en verde
EXIT_ci_2=0
```

**✅** Restaurado con `git checkout -- cmd/kitlegal/main.go` (no-op a esas alturas);
`git diff --stat -- cmd/kitlegal/main.go` vacío.

## Escenario 4 — `kitlegal version` (SC-003)

```bash
make build && ./bin/kitlegal version ; echo "código de salida: $?" ; git rev-parse HEAD
```

```text
kitlegal 615ae60-dirty
commit: 615ae60352b14f317d53f13252598dd1b367ce31
fecha:  2026-09-10T19:41:39Z
código de salida: 0
```

```text
git rev-parse HEAD → 615ae60352b14f317d53f13252598dd1b367ce31
```

Tres líneas —versión, commit y fecha—, código de salida `0` y el commit impreso **idéntico** carácter
a carácter al de la revisión construida. Ninguno de los tres valores es `dev`, `none` ni `unknown`
(FR-004).

Invocaciones no reconocidas:

```text
$ ./bin/kitlegal
uso: kitlegal version
código de salida: 2
$ ./bin/kitlegal inventado
uso: kitlegal version
código de salida: 2
```

**✅**

## Escenario 5 — Cobertura por encima del umbral (SC-011)

```bash
make test
go tool cover -func=coverage.out | tail -1
```

```text
ok  	github.com/jmorenobl/kitlegal/cmd/kitlegal	1.200s	coverage: 83.3% of statements
total:							(statements)	83.3%
```

**✅** 83,3 % frente al 70 % declarado en `codecov.yml`, sin exclusiones ni excepciones. La distancia
al 100 % es la línea de `main()`, que solo llama a `run` y propaga su código.

## Escenario 6 — Órdenes con el objeto ausente (SC-012 · FR-011)

```bash
make test-integration ; make test-e2e ; make schema-check ; make skills-sync ; make release
```

| Orden | Salida | Código |
|---|---|---|
| `make test-integration` | `go test -race -tags=integration ./...` → `ok  github.com/jmorenobl/kitlegal/cmd/kitlegal  (cached)` | `0` |
| `make test-e2e` | `test-e2e: sin tests e2e todavía; los aporta H1 (testscript)` | `0` |
| `make schema-check` | `schema-check: no hay schemas/ todavía; los aportan H4 (borrador) y H11 (contrato)` | `0` |
| `make skills-sync` | `skills-sync: no hay skills/ ni data/*.yaml todavía; los aporta H5` | `0` |
| `make release` | `release: sin configurar hasta H6 (.goreleaser.yaml)` (a stderr) + `make: *** [release] Error 1` | `2` |

**✅** Ninguna es silenciosa; `test-integration` ejecuta el comando real sobre 0 tests; las tres
siguientes nombran el objeto ausente y el hito que lo aportará; `release` falla explícitamente.

## Escenario 7 — Prerrequisitos ausentes y pin de toolchain (FR-010 · SC-010)

**a) Binario ausente.**

```bash
env PATH=/usr/bin:/bin make build
```

```text
kitlegal: falta 'go'. Se necesita Go 1.21 o superior; instálalo desde https://go.dev/dl/
make: *** [check-tools] Error 1
-> 2
```

**✅** Falla en `check-tools`, nombrando `go`, la versión mínima y de dónde obtenerlo. No es un error
opaco del intérprete de órdenes.

**b) Toolchain fijado inexistente**, provocado sobrescribiendo la variable en la línea de órdenes, sin
tocar `go.mod` ni instalar otro Go:

```bash
make check-tools GO_TOOLCHAIN=go1.26.99
```

```text
kitlegal: no se pudo usar el toolchain go1.26.99, que sale de la directiva 'toolchain' de go.mod.
kitlegal: el go command responde: go: downloading go1.26.99 (darwin/arm64)
go: download go1.26.99 for darwin/arm64: toolchain not available
kitlegal: comprueba que hay red para descargarlo, que el 'go' instalado es 1.21 o superior y que esa directiva nombra un parche existente.
make: *** [check-tools] Error 1
-> 2
```

**✅** Nombra el toolchain pedido, la directiva de `go.mod` de la que sale y el error literal del go
command (`toolchain not available`). Falla en `check-tools`, no a mitad de un control posterior.

**c) El entorno no puede debilitar el pin.**

```bash
make check-tools                        # -> 0
GOTOOLCHAIN=local make check-tools      # -> 0
GOTOOLCHAIN=auto  make check-tools      # -> 0
go version                              # go version go1.26.6 darwin/arm64
make build && go version -m bin/kitlegal | head -2
```

```text
bin/kitlegal: go1.26.6
	path	github.com/jmorenobl/kitlegal/cmd/kitlegal
```

**✅** Las tres invocaciones pasan: la asignación del `Makefile` prevalece sobre la variable heredada
del entorno, y el binario producido lleva `go1.26.6`.

> **Alcance real de esta comprobación en esta máquina.** El `go` instalado *es* `go1.26.6`, el mismo
> parche que fija la directiva. Eso hace que el caso `GOTOOLCHAIN=local` sea aquí degenerado: pasaría
> igual sin pin, porque el toolchain local ya coincide. Lo que este subescenario demuestra en esta
> ejecución es que el `Makefile` neutraliza la variable del entorno (la asignación gana a la herencia,
> research.md D1); lo que **no** puede demostrar sin un segundo parche instalado es el caso en que
> difieren. Ese caso queda cubierto por la integración continua, que ejecuta las mismas órdenes sobre
> el `go` que traiga el runner.

## Escenario 8 — Ganchos de pre-commit (FR-022)

Los dos bloques se ejecutaron en la **misma sesión de intérprete**, con el deshacer anclado a `$antes`.

```bash
make hooks
antes=$(git rev-parse HEAD)                     # 615ae60352b14f317d53f13252598dd1b367ce31
printf '\n// Comprobación del gancho de pre-commit.\n' >> cmd/kitlegal/main.go
sed -i.bak 's/^func main() {/func  main( )  {/' cmd/kitlegal/main.go && rm cmd/kitlegal/main.go.bak
git add cmd/kitlegal/main.go
git commit -m "test: comprobar el gancho"
```

```text
go tool -modfile=tools/lefthook/go.mod lefthook install
sync hooks: ✔️ (pre-commit)
...
🥊 lefthook v1.13.6  hook: pre-commit
summary: (done in 1.43 seconds)
✔️ fmt (0.37 seconds)
✔️ lint-fast (0.42 seconds)
✔️ secrets (0.55 seconds)
✔️ mod-tidy-check (0.07 seconds)
[h0-esqueleto-del-repo 2a34bdc] test: comprobar el gancho
 1 file changed, 2 insertions(+)
EXIT_commit=0
```

`git show -p HEAD -- cmd/kitlegal/main.go` — el commit contiene **solo** la línea de comentario, ya
formateada; el espaciado roto no aparece:

```diff
@@ -38,3 +38,5 @@ func run(args []string, stdout, stderr io.Writer) int {
 	return 0
 }
+
+// Comprobación del gancho de pre-commit.
```

`git diff HEAD -- cmd/kitlegal/main.go` vacío: lo confirmado es lo que hay en el árbol, es decir, el
gancho devolvió `func main() {` a su forma canónica y volvió a preparar el fichero (`stage_fixed`).

**Deshacer** —anclado a `$antes`, nunca a `HEAD~1`, y quirúrgico, nunca `--hard`:

```bash
git reset --soft "$antes"
git restore --source="$antes" --staged --worktree -- cmd/kitlegal/main.go
```

```text
git rev-parse HEAD: 615ae60352b14f317d53f13252598dd1b367ce31
antes             : 615ae60352b14f317d53f13252598dd1b367ce31
git diff --stat        "$antes" -- cmd/kitlegal/main.go → vacío
git diff --cached --stat "$antes" -- cmd/kitlegal/main.go → vacío
```

**✅** El gancho corrige y confirma lo formateado; `HEAD` y el fichero vuelven al estado de partida.
El gancho no es la autoridad final: la CI ejecuta los mismos controles.

## Escenario 9 — Detección de secretos (SC-009 · SC-004)

Rojo → verde. El token se generó al vuelo con 36 caracteres alfanuméricos aleatorios de
`/dev/urandom`, de forma que casara la regla de gitleaks (patrón `ghp_[0-9a-zA-Z]{36}`), superara su
entropía mínima y no pudiera contener una *stopword* de la lista global. Se añadió como línea de
comentario al final de `cmd/kitlegal/main.go` —inerte para el formato y para el lint— y se descartó la
variable acto seguido.

> **El valor del token no se transcribe en este registro**: ni en el comando, ni en la salida, ni en
> ningún extracto del fichero mutado. `make secrets` es `gitleaks dir . --redact --no-banner` y
> recorre **todo** el árbol, `specs/` incluido; un literal aquí dejaría `make ci` en rojo de forma
> permanente por un fichero de `gates/`, no por un control. Por eso abajo se anotan los veredictos y
> los códigos de salida, no el secreto. Por lo mismo, el comando queda registrado en su forma
> generadora —sin valor— y la salida de gitleaks se reproduce solo en su línea de resumen.

```bash
tok=$(LC_ALL=C tr -dc 'A-Za-z0-9' </dev/urandom | head -c 36)
printf '\n// Token de prueba del escenario 9, se restaura al final: <PREFIJO>%s\n' "$tok" >> cmd/kitlegal/main.go
unset tok
make secrets      # debe FALLAR
make ci           # debe FALLAR también
```

**Rojo** — `make secrets`:

```text
go tool -modfile=tools/gitleaks/go.mod gitleaks dir . --redact --no-banner
	WRN	leaks found: 1
make: *** [secrets] Error 1
-> 2
```

**Rojo** — `make ci`, y **en `secrets`**, después de haber pasado `fmt-check`, `lint`, `test`, `vuln`
y `schema-check`:

```text
go tool -modfile=tools/golangci-lint/go.mod golangci-lint fmt --diff ./...
go tool -modfile=tools/golangci-lint/go.mod golangci-lint run ./...
0 issues.
go test -race -shuffle=on -coverprofile=coverage.out ./...
ok  	github.com/jmorenobl/kitlegal/cmd/kitlegal	1.308s	coverage: 83.3% of statements
go tool -modfile=tools/govulncheck/go.mod govulncheck ./...
No vulnerabilities found.
schema-check: no hay schemas/ todavía; los aportan H4 (borrador) y H11 (contrato)
go tool -modfile=tools/gitleaks/go.mod gitleaks dir . --redact --no-banner
	WRN	leaks found: 1
make: *** [secrets] Error 1
-> 2
```

Que falle también `make ci`, y en el control que corresponde, es lo que hace cierta SC-004: el
veredicto local predice el de la propuesta de cambio, porque es literalmente la misma orden.

**Verde** — restaurado con `git restore -- cmd/kitlegal/main.go`;
`git diff --stat -- cmd/kitlegal/main.go` vacío. Comprobación adicional de que el token no sobrevive
en ningún sitio del árbol: `git status --porcelain -- cmd/` vacío y búsqueda del prefijo de la regla
bajo `cmd/` sin resultados.

**✅** Con `--redact`, gitleaks no llegó a imprimir el valor detectado en la salida; solo la línea de
resumen. No hubo, por tanto, nada que redactar a mano en este registro.

## Escenario 10 — Dependencias saneadas (SC-004 · SC-009)

Rojo → verde.

```bash
go mod edit -require=github.com/stretchr/testify@v1.11.1
make mod-tidy-check   # debe FALLAR señalando la diferencia
make ci               # debe FALLAR también
go mod edit -droprequire=github.com/stretchr/testify
make mod-tidy-check   # vuelve a pasar
```

**Rojo** — `make mod-tidy-check` señala exactamente la diferencia:

```text
go mod tidy -diff
diff current/go.mod tidy/go.mod
--- current/go.mod
+++ tidy/go.mod
@@ -3,5 +3,3 @@
 go 1.26.0

 toolchain go1.26.6
-
-require github.com/stretchr/testify v1.11.1

make: *** [mod-tidy-check] Error 1
-> 2
```

**Rojo** — `make ci` falla igualmente, en `mod-tidy-check`, que es su último control: los siete
anteriores (`fmt-check`, `lint`, `test`, `vuln`, `schema-check`, `secrets`, `mod-verify`) pasan y el
fallo llega con el mismo diff. Código de salida `2`.

**Verde** — tras `go mod edit -droprequire`, `make mod-tidy-check` devuelve `0` y
`git diff --stat -- go.mod go.sum` es vacío.

**✅**

## Escenario 12 — Documentación fundacional (SC-008 · SC-010)

```bash
head -2 LICENSE
grep -c "Unreleased" CHANGELOG.md
ls docs/ADR/
```

```text
                                 Apache License
                           Version 2.0, January 2004

2

0001-multicall.md  0002-sqlite-sin-cgo.md  0003-no-cendoj-masivo.md  0004-frontera-humana.md
```

**✅** `LICENSE` es Apache-2.0; `CHANGELOG.md` tiene su sección *Unreleased*; los cuatro ADR
fundacionales están.

**La lectura, que no es un comando.** Todo lo que hizo falta para ejecutar esta validación —los dos
únicos prerrequisitos y cómo comprobarlos, que las herramientas de control no se instalan a mano, que
el parche concreto de Go es indiferente, `make build`, `make install`, qué imprime `kitlegal version`
y con qué código de salida, el catálogo de controles con la orden de cada uno, `make hooks` y la
advertencia de que el gancho no es la autoridad final— está en `README.md`. `CONTRIBUTING.md` añade
el ritual por hito, el comportamiento de `make vuln` sin red, el procedimiento de exclusión trazable
de un falso positivo de secretos y el de subida de la directiva `toolchain`. No hubo que consultar
nada fuera de esos dos ficheros.

Única anotación: la variable `GO_TOOLCHAIN`, que el escenario 7 b sobrescribe en la línea de órdenes,
está documentada en el comentario del propio `Makefile` y no en `README.md`. No se considera una
carencia de SC-010: es una palanca para ejercer el camino de fallo de `check-tools` desde esta guía de
validación, no algo que alguien que llega por primera vez necesite para construir el proyecto o
ejecutar sus controles.

---

## Restauración: cada ruta mutada, ruta a ruta

Comprobación **acotada por ruta**, que es la que corresponde bajo el workflow `hito`: el árbol nunca
está globalmente limpio durante una tarea, porque `gates/tarea-actual.json`, `gates/tareas-intentos.json`
y este mismo registro son cambios versionados sin confirmar. Lo que hay que demostrar es que ningún
fichero mutado por un escenario sobrevive alterado.

| Ruta | Escenarios que la mutaron | `git diff --stat -- <ruta>` | `git diff --cached --stat -- <ruta>` |
|---|---|---|---|
| `cmd/kitlegal/main.go` | 3, 8, 9 | vacío | vacío |
| `go.mod` | 10 | vacío | vacío |
| `go.sum` | 10 | vacío | vacío |
| `lefthook.yml` | 8 (`make hooks`) | vacío | vacío |
| `Makefile` | — (control) | vacío | vacío |

`HEAD` vuelve a la revisión de partida: `615ae60352b14f317d53f13252598dd1b367ce31`, con lo que el
commit de prueba del escenario 8 no deja rastro en el historial.

`git status --porcelain` completo al terminar los escenarios:

```text
 M specs/001-h0-esqueleto-del-repo/gates/tarea-actual.json
 M specs/001-h0-esqueleto-del-repo/gates/tareas-intentos.json
```

Es decir: **el único cambio que T013 aporta al árbol es este registro**. Los dos ficheros de `gates/`
que aparecen modificados los escribe el paso del workflow que abre cada tarea, no ningún escenario.

Ficheros que sí cambiaron y no se restauran porque no están versionados —`.gitignore` los excluye o
viven fuera del árbol de trabajo—: `bin/kitlegal` y `coverage.out` (regenerados por `make build` y
`make test`) y `.git/hooks/pre-commit` (instalado por `make hooks`, que es precisamente lo que el
escenario 8 comprueba y lo que `README.md` pide hacer una vez).

## Cierre

Los once escenarios ejecutables de esta ronda —1 a 10 y 12— dan el resultado que la guía espera. Los
tres que son rojo → verde (3, formato; 9, secretos; 10, dependencias) fallan cuando deben, en el
control que deben y sin modificar nada, y vuelven a verde al deshacer la mutación. SC-004 queda
demostrada de forma acumulada por esos tres: el control que fallaría en la propuesta de cambio falla
también en local, porque es la misma orden.

Queda fuera de este registro el escenario 11 (SC-005 y SC-006), que solo puede medirse en la
plataforma; lo ejecuta T014 y se anota en `verificacion-pr.md`.
