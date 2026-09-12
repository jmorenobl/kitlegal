# T018 · intento 1: por qué no queda en verde

**Estado**: T018 queda **sin marcar**. Todo su contenido está hecho y verificado (ver
[`evidencia-cierre.md`](./evidencia-cierre.md)), pero `make ci` **no es verde de forma fiable**: la suite
arrastra un test inestable que no es de esta tarea y que T018 no puede tocar sin salirse de sus rutas.

## El fallo

```
$ rtk proxy make ci
--- FAIL: TestReintentosAgotados (0.00s)
    --- FAIL: TestReintentosAgotados/ocho_intentos_recorren_el_retardo_hasta_su_techo (0.01s)
        assertion_order.go:40:
            	Error Trace:	internal/httpx/reintentos_test.go:331
            	Error:      	"15.107804241s" is not less than "15.049587612s"
            	Test:       	TestReintentosAgotados/ocho_intentos_recorren_el_retardo_hasta_su_techo
            	Messages:   	cada espera arranca donde termina la anterior (SC-005, D9)
FAIL	github.com/jmorenobl/kitlegal/internal/httpx	15.166s
make: *** [test] Error 1
```

**Reproducido**: 3 fallos en 1000 ejecuciones (≈ 0,3 %).

```
$ rtk proxy go test -count=1000 -run '^TestReintentosAgotados$/^ocho_intentos_recorren_el_retardo_hasta_su_techo$' ./internal/httpx/ | grep -c 'is not less than'
3
```

No lo causa T018: esta tarea solo toca markdown (`quickstart.md`, `docs/PENDIENTES.md` y los dos ficheros
de `gates/`). El primer `make ci` del día, antes de tocar nada, sí salió verde: es cuestión de suerte.

## La causa, de raíz

El producto está bien; **la aserción del test es la que no se sostiene en el último tramo**.

`baseDelIntento` (`internal/httpx/reintentos.go:190`) es `min(500 ms · 2^(n−1), 30 s)`, y
`esperaDelIntento` aplica *equal jitter*: la espera cae en `[base/2, base)`. Con `ConIntentos(8)` hay
siete esperas:

| Espera tras el intento | Base | Banda |
|---|---|---|
| 1 | 500 ms | `[250 ms, 500 ms)` |
| 2 | 1 s | `[500 ms, 1 s)` |
| 3 | 2 s | `[1 s, 2 s)` |
| 4 | 4 s | `[2 s, 4 s)` |
| 5 | 8 s | `[4 s, 8 s)` |
| 6 | 16 s | `[8 s, **16 s**)` |
| 7 | **30 s (techo)** | `[**15 s**, 30 s)` |

Mientras la base **dobla**, el suelo de cada banda es el techo de la anterior y las esperas crecen
siempre: es lo que dice el comentario de `reintentos.go:158-163` y lo que SC-005 pide. En cuanto la base
**topa** (el séptimo intento pide 32 s y se queda en 30 s), esa propiedad deja de valer: la banda 6
`[8, 16)` y la banda 7 `[15, 30)` **se solapan en `[15 s, 16 s)`**. Si las dos caen ahí y la séptima sale
menor, `assert.IsIncreasing` (`reintentos_test.go:331`) falla aunque las dos esperas sean correctas —y
`exigeEsperaDelIntento` las da por buenas, porque lo son—. Los dos fallos observados son exactamente eso:
`15,107 s → 15,049 s` y `15,818 s → 15,201 s`.

La ironía es que la subprueba que rompe es justo la que existe para **recorrer el retardo hasta su
techo**: la aserción contradice el propósito del caso que la ejecuta.

## El arreglo

En `internal/httpx/reintentos_test.go`, acotar `IsIncreasing` al tramo en que la base dobla, que es donde
la propiedad es cierta, y dejar el tramo del techo a `exigeEsperaDelIntento`, que ya lo cubre con su
banda. El corte no es un número mágico: es el último intento cuya base sigue por debajo del techo
(`500 ms · 2^(n−1) ≤ 30 s` ⇒ `n ≤ 6`), y conviene derivarlo con la misma ley literal que ya usa
`exigeEsperaDelIntento` para no volver a ser un eco de las constantes del paquete.

Conviene además que el mensaje diga la versión completa de la ley: «cada espera arranca donde termina la
anterior **mientras la base dobla**; al llegar al techo las dos últimas comparten banda».

## Por qué no lo arregla T018

`internal/httpx/reintentos_test.go` **no está en las rutas declaradas** de T018
(`docs/PENDIENTES.md`, `internal/cli`, `internal/core/**`, `quickstart.md`,
`gates/evidencia-cierre.md`), y la tarea tiene prohibido tocar o rehacer el trabajo de otra (es de T008).
El guardián de diff rechazaría la tarea entera. Se anota aquí para que se decida dónde va el arreglo.

## Aviso para el intento siguiente

**Un reintento de T018 saldrá verde el 99,7 % de las veces sin que nadie haya arreglado nada.** Si el
intento 2 pasa, el hito cierra con el test inestable dentro y el punto 1 de la Definition of Done
(«`make ci` en verde») deja de ser cierto de forma duradera: volverá a caer en la integración continua,
en la propuesta de cambio o en cualquier hito posterior, y entonces sin el diagnóstico delante.

Lo que se pide: **no cerrar el hito por el aprobado del reintento**. O se amplían las rutas de T018 con
`internal/httpx/reintentos_test.go`, o se abre una tarea para el arreglo antes de T019.

## Lo que sí está hecho y verificado

Todo lo demás de T018, con su evidencia en `evidencia-cierre.md`:

- escenarios **1, 7, 11 y 12** ejecutados y verdes (SC-008, SC-010, SC-014, SC-015);
- umbrales de cobertura: global **93,3 %** (≥ 70), `internal/core/**` **90,1 %** (≥ 85), `internal/cli`
  **98,6 %** (≥ 90), ninguno rebajado;
- diff de `.golangci.yml` frente a `main`: solo comentarios y las cuatro palabras de `misspell` que
  autoriza el supuesto S2; ninguna regla, ningún linter, ninguna exclusión, ningún `nolint`;
- medida diferencial de `git status --porcelain` idéntica antes y después (escenarios 1 y 7);
- verbos de los dos binarios idénticos a los de H1;
- la orden **entera** de los prerrequisitos imprime «solo direcciones locales» (hizo falta corregir el
  troceado de la propia orden, no ningún test: ver `evidencia-cierre.md`);
- `docs/PENDIENTES.md` actualizado (fixtures → primera tarea `[datos]` de H4; entrada nueva para H4).

El resto de `make ci` —formato, `golangci-lint` con 0 *issues*, `govulncheck`, `gitleaks`, `go mod
verify`, `go mod tidy -diff`— pasa en todas las ejecuciones; lo único rojo es la subprueba de arriba.
