# T002, intento 1: verificación en rojo por `TestTopeDeLaSesion`; tarea redelimitada

## Qué falló

`make ci` (`go test -race -shuffle=on … ./...`, semilla `1792971470553360000`) salió en rojo solo en
`internal/evals`, por `TestTopeDeLaSesion/termina-antes` y `TestTopeDeLaSesion/muere-por-otra-senal`:
`codigo-de-la-sesion` con `124` en lugar de `3` y de `129`, `Cortada` a `true`, tras 14,07 s y 15,22 s. Los otros dos
casos de la tabla, `basta-term` y `hace-falta-kill`, en verde. El paquete tardó 463,6 s en total. Lint, fmt y el resto de
paquetes, en verde; el guardián de diff dio por buenos los 9 ficheros del intento.

Aislado, en esta sesión, el test pasa:

```
go test -race -count=1 -run '^TestTopeDeLaSesion$' ./internal/evals/
Go test: 5 passed in 1 packages
```

## Causa raíz

Los cuatro casos de `TestTopeDeLaSesion` abren la sesión con `topeDeLosTests` (1 s) y `margenDeLosTests` (1 s), los que
fija `sesionesConRepeticiones`. Solo `basta-term` y `hace-falta-kill` miran el corte; `termina-antes` y
`muere-por-otra-senal` prueban el código que deja un sustituto de `claude` que termina solo (`exit 3` y
`kill -s HUP $$`), y para ellos el tope es una cota sin holgura: `scripts/evals-sesion.sh` y el sustituto
(`sustitutoDeClaude`) lanzan una decena de procesos antes de llegar a su `exit`, y con la carga de `make ci` —toda la
suite con `-race` en paralelo y, en el mismo paquete y a la vez, las tres cachés de sesión y los dieciocho
`boe articulo --offline` más que T002 añade legítimamente a `probarLasLecturasDeLasEvals`— tardan más de 1 s. El
contexto del tope vence, `exec` llama a `Cancel`, `conTERM` se pone a `true` y `esperarConElTope` devuelve
`codigoDelTope` (124) sea cual sea el código del proceso.

Las aserciones lo confirman: en `termina-antes` el sustituto llegó (`llego` a `true`) pero no cumplió su espera ni anotó
el TERM, así que TERM llegó antes de que la subshell instalara su manejador; en `muere-por-otra-senal` cumplió la
espera y tampoco anotó el TERM, así que llegó entre `espera-cumplida` y el `exit`.

El propio fichero fija la regla que estos dos casos incumplen (`sesiones_test.go`, comentario de las constantes):
`topeSinCorte` «es el de los tests que no miran el tope: ninguna sesión suya se acerca a él, así que no depende de la
carga de la máquina», y lo usan los demás tests del repartidor (`ejecucion.Tope = topeSinCorte`).

## Por qué no se arregla en este intento

El arreglo es de `internal/evals/sesiones_test.go`, que la línea de T002 no declara, y las rutas del intento están
congeladas en `gates/tarea-actual.json`. No es un rojo legítimo de lo que T002 cambia ni pertenece a una tarea posterior:
es un test ajeno con una cota sin holgura que la carga que la tarea añade al paquete deja al descubierto. Repetir
`make ci` hasta verde dejaría la causa en su sitio y volvería en las verificaciones de T003 y T004 y en el cierre.

## Qué se ha hecho

- La línea de T002 en `tasks.md` declara `internal/evals/sesiones_test.go` y el arreglo: un campo `tope` por caso en la
  tabla de `TestTopeDeLaSesion`, `topeDeLosTests` en `basta-term` y `hace-falta-kill` y `topeSinCorte` en
  `termina-antes` y `muere-por-otra-senal`, con `ejecucion.Tope = caso.tope` y el comentario del test diciéndolo; y
  `TestTopeDeLaSesion` en verde entre sus tests. `topeDeLosTests`, `margenDeLosTests`, `topeSinCorte`, el tope y el
  margen de verdad (240 s y 10 s) y lo demás del fichero no cambian; ni código de producción, ni
  `.github/workflows/evals.yml`, ni `internal/evals/definicion.go` (research D10).
- La tarea queda como `[ ]` para que el bucle la reintente con la línea nueva.
- Los cambios del intento 1 (las tres evals, `conjunto_test.go` y `preparar_test.go`) quedan en el árbol de trabajo tal
  cual: son los que la tarea pide y nada del log los señala.

## Intento 1: verificación en rojo tras reparar (2026-10-26T00:01:40Z)

```
created by testing.(*T).Run in goroutine 60309
	/Users/jorge/go/pkg/mod/golang.org/toolchain@v0.0.1-go1.27.3.darwin-arm64/src/testing/testing.go:2260 +0x7c0

goroutine 98639 [chan receive]:
testing.(*testState).waitParallel(0xc00061eff0)
	/Users/jorge/go/pkg/mod/golang.org/toolchain@v0.0.1-go1.27.3.darwin-arm64/src/testing/testing.go:2379 +0x108
testing.(*T).Parallel(0xc0025e0d88)
	/Users/jorge/go/pkg/mod/golang.org/toolchain@v0.0.1-go1.27.3.darwin-arm64/src/testing/testing.go:1960 +0x35c
github.com/jmorenobl/kitlegal/internal/evals.TestGramaticasCoincidenConBoe.func1.16.1(0xc0025e0d88)
	/Users/jorge/Projects/kitlegal/internal/evals/formato_test.go:2787 +0x38
testing.tRunner(0xc0025e0d88, 0xc003d133b0)
	/Users/jorge/go/pkg/mod/golang.org/toolchain@v0.0.1-go1.27.3.darwin-arm64/src/testing/testing.go:2195 +0x168
created by testing.(*T).Run in goroutine 60308
	/Users/jorge/go/pkg/mod/golang.org/toolchain@v0.0.1-go1.27.3.darwin-arm64/src/testing/testing.go:2260 +0x7c0

goroutine 42572 [chan receive]:
testing.tRunner.func1()
	/Users/jorge/go/pkg/mod/golang.org/toolchain@v0.0.1-go1.27.3.darwin-arm64/src/testing/testing.go:2144 +0x660
testing.tRunner(0xc0063a8908, 0x104b469b0)
	/Users/jorge/go/pkg/mod/golang.org/toolchain@v0.0.1-go1.27.3.darwin-arm64/src/testing/testing.go:2201 +0x190
created by testing.(*T).Run in goroutine 812
	/Users/jorge/go/pkg/mod/golang.org/toolchain@v0.0.1-go1.27.3.darwin-arm64/src/testing/testing.go:2260 +0x7c0

goroutine 98757 [chan receive]:
testing.(*testState).waitParallel(0xc00061eff0)
	/Users/jorge/go/pkg/mod/golang.org/toolchain@v0.0.1-go1.27.3.darwin-arm64/src/testing/testing.go:2379 +0x108
testing.(*T).Parallel(0xc002356908)
	/Users/jorge/go/pkg/mod/golang.org/toolchain@v0.0.1-go1.27.3.darwin-arm64/src/testing/testing.go:1960 +0x35c
github.com/jmorenobl/kitlegal/internal/evals.TestGramaticasCoincidenConBoe.func1.16.1(0xc002356908)
	/Users/jorge/Projects/kitlegal/internal/evals/formato_test.go:2787 +0x38
testing.tRunner(0xc002356908, 0xc001e7f9e0)
	/Users/jorge/go/pkg/mod/golang.org/toolchain@v0.0.1-go1.27.3.darwin-arm64/src/testing/testing.go:2195 +0x168
created by testing.(*T).Run in goroutine 60309
	/Users/jorge/go/pkg/mod/golang.org/toolchain@v0.0.1-go1.27.3.darwin-arm64/src/testing/testing.go:2260 +0x7c0

goroutine 98755 [chan receive]:
testing.(*testState).waitParallel(0xc00061eff0)
	/Users/jorge/go/pkg/mod/golang.org/toolchain@v0.0.1-go1.27.3.darwin-arm64/src/testing/testing.go:2379 +0x108
testing.(*T).Parallel(0xc002356488)
	/Users/jorge/go/pkg/mod/golang.org/toolchain@v0.0.1-go1.27.3.darwin-arm64/src/testing/testing.go:1960 +0x35c
github.com/jmorenobl/kitlegal/internal/evals.TestGramaticasCoincidenConBoe.func1.16.1(0xc002356488)
	/Users/jorge/Projects/kitlegal/internal/evals/formato_test.go:2787 +0x38
testing.tRunner(0xc002356488, 0xc001e7f8c0)
	/Users/jorge/go/pkg/mod/golang.org/toolchain@v0.0.1-go1.27.3.darwin-arm64/src/testing/testing.go:2195 +0x168
created by testing.(*T).Run in goroutine 60309
	/Users/jorge/go/pkg/mod/golang.org/toolchain@v0.0.1-go1.27.3.darwin-arm64/src/testing/testing.go:2260 +0x7c0
FAIL	github.com/jmorenobl/kitlegal/internal/evals	601.856s
ok  	github.com/jmorenobl/kitlegal/internal/graph	8.943s	coverage: 88.4% of statements
ok  	github.com/jmorenobl/kitlegal/internal/httpx	19.334s	coverage: 96.4% of statements
ok  	github.com/jmorenobl/kitlegal/internal/mcp	1.842s	coverage: 100.0% of statements
ok  	github.com/jmorenobl/kitlegal/internal/mcp/mcptest	2.666s	coverage: 96.0% of statements
ok  	github.com/jmorenobl/kitlegal/internal/render	2.497s	coverage: 95.1% of statements
ok  	github.com/jmorenobl/kitlegal/internal/skills	209.618s	coverage: 97.9% of statements
ok  	github.com/jmorenobl/kitlegal/internal/source/bocm	45.244s	coverage: 98.6% of statements
ok  	github.com/jmorenobl/kitlegal/internal/source/boe	15.961s	coverage: 99.3% of statements
ok  	github.com/jmorenobl/kitlegal/internal/source/boe/fiestas	11.619s	coverage: 98.5% of statements
ok  	github.com/jmorenobl/kitlegal/internal/workflow	157.894s	coverage: [no statements]
FAIL
make: *** [test] Error 1

```

## Intento 2: en verde (2026-10-26)

- Los dos tests de las lecturas, vistos en rojo con las evals de hoy antes de restaurar el diff de `evals/` del intento 1:
  `TestEvalsDelRepositorio/lecturas` falla por las nueve entradas que el mapa espera y las evals no llevan, y
  `TestCacheDeLasSesiones/lecturas` sale con 4 (`fuente-no-disponible`) en cada bloque nuevo con `--offline`, sin fecha y
  con `--fecha 9999-12-31`, desde la caché de la sesión de su eval.
- `internal/evals/sesiones_test.go`: la tabla de `TestTopeDeLaSesion` lleva el campo `tope` por caso y la sesión se abre
  con `ejecucion.Tope = caso.tope`: `topeDeLosTests` en `basta-term` y `hace-falta-kill`, `topeSinCorte` en
  `termina-antes` y `muere-por-otra-senal`; el comentario del test lo dice. Las constantes y lo demás del fichero no
  cambian. `go test -race -count=1 -run '^TestTopeDeLaSesion$'` pasa (5,8 s; `termina-antes` 2,31 s,
  `muere-por-otra-senal` 2,18 s).
- Quickstart escenario 6: `jq` da `Ley 35/2006, a18 a7 a19 a20 dadecimoctava` y `Real Decreto Legislativo 2/2004, a15
  a60 a92 a100 a104`; `git ls-files` da `8`; los tres `go test` en `ok`. Escenario 7: `TestDefinicionDelJob` en `ok` y
  `timeout-minutes: 352` en la línea 276 de `.github/workflows/evals.yml`.
- `make ci`, por objetivos y en primer plano porque entero supera el tope de 10 min por orden de la sesión: `fmt-check`,
  `lint` (0 issues), `schema-check`, `skills-check`, `goreleaser-check`, `secrets`, `mod-verify`, `mod-tidy-check`,
  `vuln`, `test` (`internal/evals` en 130,7 s), `test-integration` (177,8 s) y `test-tiempos`, todos en verde.
- Ningún supuesto nuevo: `gates/supuestos.md` no cambia.
