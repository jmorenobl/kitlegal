# T012: por qué no quedó en verde

## Intento 1: tarea mal delimitada (redelimitada, queda `[ ]`)

La emisión está hecha y probada dentro de las rutas declaradas, y el árbol la conserva para el intento siguiente:

- `internal/source/boe/grafo.go`: `observadoDeLosArticulos`, `operacionesDelArticulo` y `eliDeLaNorma`
  (contracts/emision.md §1; research D20).
- `internal/source/boe/articulo.go`: `resultadoDeLosArticulos` pone lo observado en el resultado de éxito de
  `articulo` y `articulos`; con las líneas de un ensayo no emite nada.
- `internal/source/boe/grafo_test.go`: `TestObservadoDeBoe` (rojo por aserción antes de implementar, verde
  después; la sonda `Path` en lugar de `EscapedPath` la pone en rojo en `url-eli-barra-escapada`).
- `internal/source/boe/articulo_test.go`: las comparaciones de `Resultado` entero de éxito llevan el `Observado`
  esperado (`resultadoDelArticulo`, `resultadoDeArticulos`).
- Supuestos en `gates/supuestos.md` (cuatro líneas «T012»).

`make ci` falla en **un único test, fuera de las rutas declaradas**: el subtest
`articulo-con-los-metadatos-en-cache` de `TestCacheDeLosSeisVerbos`, en `internal/source/boe/fuente_test.go`,
compara con `resultadoResuelto` un `Resultado` entero de éxito de `articulo`, que ahora lleva lo observado. Es
exactamente el caso que la tarea prevé para `articulo_test.go` («si algún test de éxito compara un `Resultado`
entero, con el `Observado` esperado»), pero en un fichero que no declaraba. La línea de T012 en `tasks.md` gana
`internal/source/boe/fuente_test.go` (comprobado con el filtro de rutas del workflow: las mismas de antes más esa).

Salida de `make ci` (formato y lint en verde; el resto de paquetes, en verde):

```
go test -race -shuffle=on -coverprofile=coverage.out -skip '^TestMedidasDeTiempo$' ./...
…
--- FAIL: TestCacheDeLosSeisVerbos (0.00s)
    --- FAIL: TestCacheDeLosSeisVerbos/articulo-con-los-metadatos-en-cache (0.06s)
        fuente_test.go:677:
            	Error Trace:	internal/source/boe/metadatos_test.go:456
            	            	internal/source/boe/fuente_test.go:677
            	Error:      	Not equal:
            	            	expected: schema.Resultado{…, Grafo:schema.Observado{Vigencia:0, Operaciones:[]schema.Operacion(nil)}}
            	            	actual  : schema.Resultado{…, Grafo:schema.Observado{Vigencia:604800000000000, Operaciones:[…las seis de a21…]}}
FAIL	github.com/jmorenobl/kitlegal/internal/source/boe
make: *** [test] Error 1
```

### Lo que tiene que hacer el intento 2

Solo esto, en `internal/source/boe/fuente_test.go`, línea 677:

```go
		compruebaResuelta(t, resultado, err,
			resultadoDelArticulo(direccionDelArticulo21, t0, eliDeLaLey39, articuloDelArticulo21(t)))
```

en lugar de `resultadoResuelto(direccionDelArticulo21, t0, articuloDelArticulo21(t))`. Verificado sobre una copia
desechable del árbol con ese cambio: `go test -race ./internal/source/boe/` en verde, y `make fmt-check lint test
test-integration test-tiempos schema-check skills-check` en verde; `make secrets` en verde en el repositorio. Después,
`make ci` en primer plano y marcar `[X]`.

Ningún otro test del árbol compara un `Resultado` de `articulo` o `articulos` entero: `go test -tags=integration
./...` sobre el árbol de este intento solo da ese rojo. `TestOfflineDeLosSeisVerbos` (`--offline --dry-run` igual
que `--offline`) sigue en verde porque la fuente pone lo observado también cuando, bajo `--dry-run`, los artículos
salen de sus entradas: no entregarlo es cosa del kernel (FR-034; research D4; supuesto «T012» correspondiente).

Observado una vez en la copia, bajo la carga de `make test test-integration` seguidos: `TestTuberiaCerrada/echo
repetir --help` (`internal/app/tuberia_test.go`, applet de ejemplo `echo`, ajeno a `boe` y al grafo) salió con 0 en
lugar de 1. Aislado, 20 de 20 en verde con `-race -tags=integration`; en el repositorio, en verde en el objetivo
`test` de `make ci` y en `go test -tags=integration ./...`. No se tocó; si reaparece en la verificación, es una
intermitencia previa a T012.

## Intento 2: en verde (2026-09-28)

Con `internal/source/boe/fuente_test.go` ya declarado, el intento hizo solo lo previsto:

1. Rojo reproducido sobre el árbol del intento 1 con `go test -race -count=1 ./internal/source/boe/`: un único
   fallo, `TestCacheDeLosSeisVerbos/articulo-con-los-metadatos-en-cache`, con el `Grafo` actual lleno (las seis
   operaciones de a21, vigencia 604 800 s) y el esperado vacío.
2. La comparación de ese subtest (`fuente_test.go`, línea 677) pasa a `resultadoDelArticulo(direccionDelArticulo21,
   t0, eliDeLaLey39, articuloDelArticulo21(t))`, el mismo ayudante que ya usan las comparaciones de `Resultado` entero
   de `articulo_test.go`. Ningún byte más de código ni de tests.
3. `go test -race -count=1 -tags=integration ./...` y `go vet ./...` en verde en `boe` y en todos los paquetes salvo la
   intermitencia de abajo; `golangci-lint` del repositorio sobre `./internal/source/boe/...`: 0 hallazgos.
4. `make ci` en primer plano: `ci: todos los controles en verde`; `internal/source/boe` al 99,5 % de cobertura en
   `test` y en `test-integration`; `test-tiempos`, `vuln`, `schema-check`, `skills-check`, `goreleaser-check`,
   `secrets` y `mod-verify` en verde. Sin ninguna línea `FAIL` en el registro.

### Intermitencia ajena a T012, no tocada: `TestReintentar` en `internal/graph`

En la pasada previa de `go test -race -count=1 -tags=integration ./...` (no en `make ci`, que salió en verde) cayó una
vez `TestReintentar/el_contexto_que_termina_corta_la_espera` (`internal/graph/espera_test.go:156`, de T007):

```
Error: "299.999709ms" is not greater than or equal to "300ms"
Messages: cada tramo se espera entero
```

Causa, leída en el test: el plazo del contexto se arma en la línea 127 (`context.WithTimeout(ctx, caso.plazo)`) y el
origen de la medida, `inicio := time.Now()`, se toma en la 135, después. El contexto vence 300 ms tras la 127; lo que
se mide desde la 135 puede quedar por debajo de 300 ms exactamente por lo que tarde el hilo entre las dos líneas
(aquí, 0,3 µs). Es el patrón ya visto en `httpx` (H5, T032): el origen debe tomarse antes de armar lo que ancla los
instantes. El arreglo es de una línea —tomar `inicio` antes de `context.WithTimeout`, o medir contra
`ctx.Deadline()`— y no está aplicado porque la ruta no es de T012 y la tarea no está mal delimitada por ello: le
corresponde a la revisión final o a una tarea previa al cierre que declare `internal/graph/espera_test.go`. Aislado,
`go test -race -count=10 -run TestReintentar ./internal/graph/` dio 10 de 10 en verde, y `make ci` lo pasó.
