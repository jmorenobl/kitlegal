# T001, intento 1: sin verde por un test que la tarea no toca y una ruta que no declaraba

## Qué quedó hecho (en el árbol de trabajo)

- `specs/037-h41-3-por-que-normativa/aceptacion/laboratorio-tabla.txtar`, el guion de la tarea. La verificación lo dio
  por bueno: `guardián de diff ok (T001, 6 ficheros)` y `aceptación: 1 guiones en rojo por una aserción; evals nuevas: 0`
  (gates/ci.log). No se ha tocado en la reparación.
- Nada más: el árbol de `internal/` está como en la base (`git status --short internal/` no da nada).

## Por qué la verificación quedó en rojo

`make ci` falló en un solo test, que ningún fichero de la tarea toca:
`TestTopeDeLaSesion/muere-por-otra-senal` (internal/evals/sesiones_test.go:121 y :125), con `expected: "129\n"`,
`actual: "0\n"`.

El caso lanza el sustituto de `claude` con `KITLEGAL_SUSTITUTO_SENAL=HUP`; el sustituto, un guion de `/bin/sh`, se envía
`kill -s HUP` a sí mismo y el repartidor tiene que anotar `128 + SIGHUP`. Una señal que llega ignorada a una shell no
interactiva no se puede restaurar desde la shell, y las señales ignoradas se heredan de proceso en proceso: si el proceso
de test llega con SIGHUP ignorada, el `kill` no mata al sustituto, que sigue hasta su `exit 0`.

Medido en esta sesión:

- Tal cual, `go test -race -count=1 -run TestTopeDeLaSesion ./internal/evals` pasa (5 de 5). Las órdenes de esta
  sesión llegan con SIGHUP sin ignorar (`SigIgn: 0000000000000000` en `/proc/self/status`).
- Con un `TestMain` de un solo uso que llama a `signal.Ignore(syscall.SIGHUP)` (fichero sin versionar, ya borrado), el
  mismo `go test` falla solo en ese subtest y con el mismo mensaje del log: `expected: "129\n"`, `actual: "0\n"`.
- Con ese `TestMain` y, además, `signal.Notify` de `syscall.SIGHUP` en el subtest antes de `abrirSesion` (con su
  `signal.Stop` en `t.Cleanup`), `go test -race -count=3 -run TestTopeDeLaSesion ./internal/evals` pasa (15 de 15). El
  experimento está deshecho.

No medido: la disposición de SIGHUP en el proceso del verificador. Desde el sandbox de la sesión no se ve; que llega
ignorada (por ejemplo, un run lanzado con `nohup`) es lo que se deduce de que el fallo sea el mismo, byte a byte, que el
reproducido. El código del repartidor (`codigoDelProceso`) hace lo que el test pide: el defecto es la condición de
partida que el test da por supuesta.

## Por qué no se arregla en este intento

El remedio toca internal/evals/sesiones_test.go, que la tarea no declaraba: el guardián de diff lo rechazaría. La línea
de T001 en tasks.md queda redelimitada con esa ruta y con el cambio, y la tarea vuelve a `[ ]`.

## Qué tiene que hacer el intento siguiente

1. No tocar el guion de aceptación: está escrito y la verificación lo aceptó.
2. En `TestTopeDeLaSesion`, en cada subtest y antes de `abrirSesion`: un canal `chan os.Signal` con hueco para una
   señal, `signal.Notify(canal, syscall.SIGHUP)` y `t.Cleanup(func() { signal.Stop(canal) })`, con un comentario que
   diga por qué; hace falta importar `os/signal`. Ningún caso, valor esperado ni aserción cambia, ni ningún otro fichero
   de Go.
3. Comprobarlo como arriba: con un `TestMain` de un solo uso que ignore SIGHUP, el test pasa; el fichero se borra antes
   de terminar (el guardián no lo admite y no es parte de la entrega).
4. `make ci` en primer plano hasta verde y la tarea a `[X]`.

## Alternativas descartadas

- Cambiar el repartidor para que las sesiones no hereden las señales ignoradas: es comportamiento del producto que
  ningún requisito del hito pide (constitución, criterio 2).
- Cambiar la señal del caso (por ejemplo, a USR1): cambia el dato y el valor esperado del test y deja la misma
  suposición sobre el entorno, con otra señal.
- Dejar el test en rojo o saltarlo: con SIGHUP ignorada en el verificador, el rojo se repite en cada tarea del hito.
- Normalizar las señales en scripts/workflow/verificar.sh o en scripts/hito.sh: es el workflow, que cambia una persona
  con su versión; queda como pendiente para quien lea el informe.

## Intento 1: verificación en rojo tras reparar (2026-11-16T11:57:37Z)

```
ok  	github.com/jmorenobl/kitlegal	1.468s	coverage: 100.0% of statements
	github.com/jmorenobl/kitlegal/cmd/empaquetar		coverage: 0.0% of statements
ok  	github.com/jmorenobl/kitlegal/cmd/kitlegal	1.014s	coverage: 0.0% of statements
ok  	github.com/jmorenobl/kitlegal/cmd/laboratorio	1.041s	coverage: 87.5% of statements
ok  	github.com/jmorenobl/kitlegal/data	1.009s	coverage: 100.0% of statements
ok  	github.com/jmorenobl/kitlegal/internal	1.615s	coverage: [no statements]
ok  	github.com/jmorenobl/kitlegal/internal/app	9.333s	coverage: 96.2% of statements
	github.com/jmorenobl/kitlegal/internal/app/ejemplo		coverage: 0.0% of statements
ok  	github.com/jmorenobl/kitlegal/internal/app/ejemplo/kitlegal-e2e	1.015s	coverage: 81.8% of statements
ok  	github.com/jmorenobl/kitlegal/internal/cache	2.928s	coverage: 87.9% of statements
ok  	github.com/jmorenobl/kitlegal/internal/cli	1.314s	coverage: 98.1% of statements
?   	github.com/jmorenobl/kitlegal/internal/core	[no test files]
ok  	github.com/jmorenobl/kitlegal/internal/core/cita	1.035s	coverage: 100.0% of statements
ok  	github.com/jmorenobl/kitlegal/internal/core/grafo	1.106s	coverage: 92.5% of statements
ok  	github.com/jmorenobl/kitlegal/internal/core/ids	1.037s	coverage: 100.0% of statements
ok  	github.com/jmorenobl/kitlegal/internal/core/instalacion	1.385s	coverage: 99.7% of statements
ok  	github.com/jmorenobl/kitlegal/internal/core/plazos	20.920s	coverage: 99.3% of statements
ok  	github.com/jmorenobl/kitlegal/internal/core/schema	1.025s	coverage: 89.6% of statements
ok  	github.com/jmorenobl/kitlegal/internal/core/territorio	4.751s	coverage: 100.0% of statements
ok  	github.com/jmorenobl/kitlegal/internal/disco	1.024s	coverage: 95.6% of statements
ok  	github.com/jmorenobl/kitlegal/internal/empaquetado	1.256s	coverage: 94.2% of statements
-test.shuffle 1794830167268147909
--- FAIL: TestTopeDeLaSesion (0.00s)
    --- FAIL: TestTopeDeLaSesion/muere-por-otra-senal (4.38s)
        sesiones_test.go:121: 
            	Error Trace:	/home/jorge/code/kitlegal/internal/evals/sesiones_test.go:121
            	Error:      	Not equal: 
            	            	expected: "129\n"
            	            	actual  : "0\n"
            	            	
            	            	Diff:
            	            	--- Expected
            	            	+++ Actual
            	            	@@ -1,2 +1,2 @@
            	            	-129
            	            	+0
            	            	 
            	Test:       	TestTopeDeLaSesion/muere-por-otra-senal
        sesiones_test.go:125: 
            	Error Trace:	/home/jorge/code/kitlegal/internal/evals/sesiones_test.go:125
            	Error:      	Not equal: 
            	            	expected: 129
            	            	actual  : 0
            	Test:       	TestTopeDeLaSesion/muere-por-otra-senal
FAIL
coverage: 95.9% of statements
FAIL	github.com/jmorenobl/kitlegal/internal/evals	90.101s
ok  	github.com/jmorenobl/kitlegal/internal/graph	6.355s	coverage: 88.4% of statements
ok  	github.com/jmorenobl/kitlegal/internal/httpx	12.252s	coverage: 96.4% of statements
ok  	github.com/jmorenobl/kitlegal/internal/mcp	1.042s	coverage: 100.0% of statements
ok  	github.com/jmorenobl/kitlegal/internal/mcp/mcptest	1.111s	coverage: 97.0% of statements
ok  	github.com/jmorenobl/kitlegal/internal/render	1.017s	coverage: 95.1% of statements
ok  	github.com/jmorenobl/kitlegal/internal/skills	9.298s	coverage: 97.0% of statements
ok  	github.com/jmorenobl/kitlegal/internal/source/bocm	3.383s	coverage: 98.6% of statements
ok  	github.com/jmorenobl/kitlegal/internal/source/boe	2.159s	coverage: 99.2% of statements
ok  	github.com/jmorenobl/kitlegal/internal/source/boe/fiestas	1.707s	coverage: 98.5% of statements
ok  	github.com/jmorenobl/kitlegal/internal/workflow	4.236s	coverage: [no statements]
FAIL
make: *** [Makefile:97: test] Error 1

```
