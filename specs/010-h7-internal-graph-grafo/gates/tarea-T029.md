# T029: por qué no quedó en verde

## Intento 1: tarea mal delimitada (redelimitada, queda `[ ]`)

El trabajo de la tarea está hecho dentro de sus rutas y el árbol lo conserva para el intento siguiente:

- `specs/010-h7-internal-graph-grafo/datos-tocados.md` (nuevo): la lista de esquemas y datos de prueba con su tarea y
  su motivo, los guiones que añadirá la activación de la suite, los tres puntos de la Definition of Done que no
  aplican, la cobertura medida (`internal/core/**` 99,6 %, global 97,4 %) y los escenarios 2 a 12 del quickstart con su
  resultado.
- `specs/010-h7-internal-graph-grafo/quickstart.md`: los dos desajustes del propio texto (escenario 10, la frase de
  `--no-graph` partida en dos líneas por la ayuda; escenario 11, `t.Logf` solo visible con `-v`).
- `gates/supuestos.md`: cuatro líneas «T029».

`make ci` falla en **un único test, fuera de las rutas declaradas** y ajeno a lo que la tarea toca:
`TestReintentar/el_contexto_que_termina_corta_la_espera`, en `internal/graph/espera_test.go` (línea 156, de T007).
Formato, lint y el resto de paquetes, en verde:

```
--- FAIL: TestReintentar (0.02s)
    --- FAIL: TestReintentar/el_contexto_que_termina_corta_la_espera (0.30s)
        espera_test.go:156:
            	Error:      	"297.048542ms" is not greater than or equal to "300ms"
            	Messages:   	cada tramo se espera entero
FAIL	github.com/jmorenobl/kitlegal/internal/graph	5.950s
```

### Causa

Es la intermitencia de T007 ya anotada en `gates/tarea-T012.md` («Intermitencia ajena a T012, no tocada») y vista
después en T017 (294,64 ms) y T024 (299,83 ms): el plazo del contexto se arma en la línea 127
(`context.WithTimeout(ctx, caso.plazo)`) y el origen de la medida, `inicio := time.Now()`, se toma en la 135, después.
El contexto vence 300 ms tras la 127; medido desde la 135 queda por debajo exactamente por lo que tarde el hilo entre
las dos líneas, que bajo `-race` con el árbol entero en paralelo llega a varios milisegundos. `reintentar` solo vuelve
por `<-ctx.Done()` (o por `ctx.Err()` tras el intento), es decir, nunca antes del plazo: con el origen tomado antes de
armarlo, `tardo ≥ 300 ms` queda garantizado por el reloj monotónico, sin ninguna holgura en la cota.
`TestEsperaAnteUnBloqueoDeVerdad` tiene el mismo patrón latente (plazo en la 324, origen en la 335).

No se arregló en T012, T017 ni T024 porque ninguna declaraba `internal/graph/espera_test.go` y sus verificaciones
acabaron en verde al repetirse. T029 es la última tarea del hito: si no lo declara ella, el rojo llega a la revisión
final y a la CI de la propuesta de cambio con la misma probabilidad. Repetir `make ci` hasta un verde probabilístico
sería cerrar sin arreglar la raíz (constitución, «Criterio de decisión autónoma»).

### Comprobado sobre copias desechables del árbol (fuera del repositorio)

- **Mutante rojo** (el árbol tal cual más `time.Sleep(3 * time.Millisecond)` entre armar el plazo y tomar el origen,
  el hueco que abre el planificador hecho fijo): `go test -race -count=3 -run 'TestReintentar$' ./internal/graph/`
  falla 3 de 3 en ese subtest con 297,74, 298,05 y 297,69 ms, el mismo defecto y la misma magnitud que el rojo de la
  verificación.
- **Arreglo** (el origen antes del plazo en los dos tests): `-count=30` sobre `TestReintentar` y
  `TestEsperaAnteUnBloqueoDeVerdad`, 30 de 30 en verde; el paquete entero con `-race -shuffle=on -tags=integration`,
  en verde; `golangci-lint run ./internal/graph/...` del repositorio, 0 hallazgos; `golangci-lint fmt --diff`, sin
  diferencias.
- **Arreglo más el mismo hueco fijo**: `-count=10`, 10 de 10 en verde. El mecanismo queda cerrado, no tapado.

### Redelimitación

La línea de T029 en `tasks.md` gana `internal/graph/espera_test.go` y describe la corrección (comprobado con el filtro
de rutas de `scripts/workflow/tarea.sh`: las rutas de antes más esa). La tarea vuelve a `[ ]`. La excepción declarada
de T029 en la cabecera de `tasks.md` dice ahora que la tarea lleva esa corrección de test y ningún código de producto.

### Lo que tiene que hacer el intento 2

Solo esto, en `internal/graph/espera_test.go`, en los dos tests que arman un plazo:

1. En `TestReintentar` (subtest, tras `t.Parallel()`): `inicio := time.Now()` pasa de la línea 135 a justo antes de
   `ctx := t.Context()`, con su comentario de por qué (el origen va antes de lo que ancla el instante en que termina la
   espera). Queda:

   ```go
   		t.Parallel()

   		// El origen de la medida va antes del plazo, que es lo que ancla el
   		// instante en que termina la espera: medido después, lo que tarda el
   		// hilo entre las dos líneas se descuenta y la cota inferior falla.
   		inicio := time.Now()
   		ctx := t.Context()

   		if caso.plazo > 0 {
   			…
   		}

   		var llamadas int

   		err := caso.espera.reintentar(ctx, func() error {
   ```

2. En `TestEsperaAnteUnBloqueoDeVerdad` (subtest, tras `abrirBaseDePrueba`): igual, `inicio := time.Now()` pasa de la
   línea 335 a justo antes de `ctx := t.Context()`, antes del plazo y del `time.AfterFunc` que suelta el bloqueo; el
   `var tx *sql.Tx` se queda donde está.

Ningún byte más de código ni de tests; ninguna holgura en las cotas (`tarda`, `sobra`, `tiempoDeSobra`). Después:

3. En `datos-tocados.md`, sección «Cobertura»: la frase «T029 no toca código» pasa a decir que la única modificación de
   código de T029 está en un fichero `_test.go` de `internal/graph` y no cambia ninguna sentencia de producto, así que
   las cifras siguen valiendo; comprobarlo con `go tool cover -func coverage.out` sobre el perfil del `make ci` del
   intento y, si alguna cifra cambia, escribir la nueva.
4. `make ci` en primer plano y marcar `[X]`. Los escenarios del quickstart, la lista de lo tocado y los supuestos ya
   están hechos y no hay que repetirlos.

## Intento 2: en verde (queda `[X]`)

Aplicado exactamente lo que pide la sección anterior, en `internal/graph/espera_test.go`: `inicio := time.Now()` pasa
a estar antes de `ctx := t.Context()` en los dos tests que arman un plazo, con su comentario; ningún otro byte de código
y ninguna holgura en las cotas (`tarda`, `sobra`, `tiempoDeSobra` no cambian). Los pasos 3 y 4, hechos.

- **Rojo primero, en el árbol**: con el hueco fijo de 3 ms entre el plazo y el origen (un `time.Sleep`, retirado
  después; `grep MUTANTE` da 0), `go test -race -count=3 -run 'TestReintentar$' ./internal/graph/` falla 3 de 3 en
  el mismo subtest y con la misma magnitud que la verificación del intento 1: 297,70, 298,39 y 296,84 ms < 300 ms.
- **Verde con el arreglo y el hueco aún puesto**: `-count=5`, 5 de 5.
- **Sin el hueco**: `-count=30` sobre `TestReintentar` y `TestEsperaAnteUnBloqueoDeVerdad`, 30 de 30;
  `golangci-lint fmt --diff ./internal/graph/` sin diferencias y `golangci-lint run ./internal/graph/...` con
  0 hallazgos (los dos, el `golangci-lint` de `tools/golangci-lint/go.mod`).
- **`make ci`** en primer plano, en verde (dos minutos), sin ningún `FAIL` en su salida; `internal/graph` en verde en
  `test` y en `test-integration`.
- **Cobertura** sobre el `coverage.out` de ese `make ci`: global 97,4 %; `internal/core/**` 99,6 % (2531 de 2542
  sentencias); `internal/core/grafo` 100,0 %; `internal/graph` 92,2 %. Las mismas cifras que el intento 1, como
  corresponde a un cambio que no toca ninguna sentencia de producto; «Cobertura» de `datos-tocados.md` lo dice así.
- Los escenarios 2 a 12 del quickstart no se repiten: el código que ejercen es el mismo de `6d01b81` sobre el que los
  ejecutó el intento 1 (el diff de este intento fuera de `specs/` es solo el `_test.go`), y su resultado sigue en
  «Quickstart» de `datos-tocados.md`.
