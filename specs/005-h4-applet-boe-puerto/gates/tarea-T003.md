# T003 · intento 1 · sin marcar: tarea mal delimitada

## Causa

La implementación y los tests de la tarea están hechos y en verde dentro de sus rutas, pero `make ci` queda en
rojo por dos tests de `internal/httpx/reproducir_test.go`, un fichero que la tarea no declaraba (no es el
`_test.go` de ninguno de sus `.go`) y que el guardián de diff rechazaría:

1. **`TestReplayGarantiasVigentes/identifica la petición`** (línea 308) exige que bajo el decorador de
   identificación esté directamente `*transporteDeReproduccion`. El contrato httpx-acepta-e-instante §4 fija
   `Replay: identificación → marca de emisión → transporte de reproducción`, así que ahí está ahora
   `*decoradorDeMarcaDeEmision`. La aserción tiene que bajar un escalón más, sin perder lo que protege
   (FR-046, FR-049: ni robots.txt, ni ritmo, ni reintentos, ni transporte de red).
2. **`TestReplayEsDeterminista`** (líneas 157 y 161) compara dos `Respuesta` enteras con `assert.Equal`. Desde
   T003, `Respuesta.Instante` sale de la hora del cliente (`time.Now` por omisión, contrato §2: «Reproducción →
   hora de `ConHora` al servir la grabación»), y dos peticiones no pueden dar el mismo instante. El determinismo
   de FR-048 vale para lo que guarda la grabación; el instante hay que fijarlo con `ConHora`.

Además, el comentario de `decoradorDeGrabacion` en `internal/httpx/grabar.go` («el escalón de más abajo de la
cadena, el único por debajo del transporte») deja de ser cierto con la marca entre la grabación y el transporte.
Se añade a la tarea para no dejar un comentario falso.

No hay otra salida que respete el contrato: la marca es un decorador propio en las dos cadenas (§4), y meterla
dentro del transporte de reproducción o de la identificación exigiría tocar `reproducir.go` o `identificar.go`,
que tampoco están declarados y además contradicen el contrato.

## Redelimitación

La línea de T003 en `tasks.md` declara ahora también `internal/httpx/reproducir_test.go` y `internal/httpx/grabar.go`
(solo ese comentario). El trabajo queda en el árbol para el intento siguiente.

## Lo que queda en el árbol (dentro de las rutas congeladas)

- `internal/httpx/instante.go` (nuevo): `marcaDeEmision` (`atomic.Pointer[time.Time]`; `anotar`, `reiniciar`,
  `instanteO`), clave de contexto sin exportar, `contextoConMarca`, `contextoSinMarca`, `marcaDelContexto` y el
  decorador `conMarcaDeEmision`, que anota `hora()` justo antes del transporte si la petición lleva marca.
- `internal/httpx/peticion.go`: `Respuesta.Instante`. `internal/httpx/errores.go`: `Error.Instante` y
  `(*Error).fechado`, que devuelve una **copia** fechada (la denegación del robots.txt se cachea y se comparte entre
  goroutines: fecharla en su sitio sería una carrera de datos).
- `internal/httpx/cliente.go`: `ConHora` (nula → «argumentos»; admitida por `New` y `Replay`), la marca en las dos
  cadenas en el orden del §4, `Pedir` que la pone en el contexto, `seguirLaCadena` que la reinicia antes de cada
  salto, y la copia al volver (vacía → `hora()`); argumentos inválidos → `hora()`; ensayo → cero. Las funciones
  internas del camino de `Pedir` (`comprobarPeticion`, `comprobarFormato`, `seguirLaCadena`, `emitir`,
  `falloAlEmitir`, `siguienteDestino`, `entregar`) devuelven `*Error` en vez de `error`, de modo que ningún fallo
  puede llegar a `Pedir` sin poder fecharse y no hace falta ninguna rama inalcanzable.
- `internal/httpx/robots.go`: la obtención del robots.txt usa `contextoSinMarca`.
- `internal/httpx/instante_test.go`: `TestInstanteDeEmision` con los nueve subtests (sin anidar, para que la salida
  detallada tenga exactamente nueve líneas); `internal/httpx/cliente_test.go`: `TestConHoraRechazaNula`.

## Cambios verificados para el intento siguiente

Aplicados y comprobados sobre una copia desechable (`rsync` del árbol con estos cambios a `/tmp`):

- `reproducir_test.go`, subtest «identifica la petición»:
  `escalonDeLaMarca, esLaMarca := espia.siguiente.(*decoradorDeMarcaDeEmision)` con `require.True`, y después
  `_, esDeReproduccion := escalonDeLaMarca.siguiente.(*transporteDeReproduccion)` con la aserción de siempre.
- `reproducir_test.go`, `TestReplayEsDeterminista`: `horaFija := ConHora(func() time.Time { return instanteDePrueba })`
  en los dos `clienteDeReproduccion`, y el comentario del test explicando que el instante es lo único que la grabación
  no guarda. La parte de los fallos compara `Error()`, que no incluye el instante, y no cambia.
- `grabar.go`: «decoradorDeGrabacion es el escalón que va justo encima de la marca de emisión y, por tanto, del
  transporte: …».

## Evidencia

- Rojo antes del código: `go vet ./internal/httpx/` → `undefined: ConHora`, `respuesta.Instante undefined`.
- En el repositorio: `TestInstanteDeEmision` (9 subtests) y `TestConHoraRechazaNula` en verde con `-race -count=5`;
  `golangci-lint run ./internal/httpx/...` → 0 issues; `golangci-lint fmt --diff` sin diferencias.
- Sondas de mutación sobre copias, cada una en rojo:
  sin `contextoSinMarca` en robots.go → `acierto`, `acierto-tras-reintentos`, `fallo-tras-reintentos`,
  `denegada-por-robots`, `redireccion`, `sin-turno`; sin `marca.reiniciar()` → `redireccion`; `Replay` sin marca →
  `reproduccion`; marca por encima del ritmo → `sin-turno`; marca por encima del robots.txt →
  `acierto-tras-reintentos`, `fallo-tras-reintentos`, `denegada-por-robots`, `sin-turno`.
- `go clean -testcache && make -C <copia> ci` con los cambios de arriba → código 0, «ci: todos los controles en verde»
  (`internal/httpx` 97,2 % de cobertura).
- `make ci` en el repositorio → código 2: los dos tests de `reproducir_test.go` descritos y, además,
  `internal/cache` › `TestContextoCanceladoDuranteLaMigracion`, ajeno a esta tarea (no toca la caché): pasó 200 de 200
  veces en aislamiento (`go test -race -count=200 -run '^TestContextoCanceladoDuranteLaMigracion$' ./internal/cache/`).
  Es la segunda aparición del fallo intermitente con SQLITE_BUSY ya visto en T001; merece una tarea propia que lo
  arregle de raíz en `internal/cache`, nunca un reintento en el test.

# T003 · intento 2 · en verde y marcada

Con las dos rutas añadidas en la redelimitación, el intento 2 aplica exactamente los cambios verificados arriba y
nada más fuera de ellas:

- `internal/httpx/reproducir_test.go`: «identifica la petición» exige `*decoradorDeMarcaDeEmision` bajo la
  identificación (`require.True`) y `*transporteDeReproduccion` bajo la marca (la aserción de siempre, FR-046,
  FR-049); `TestReplayEsDeterminista` construye los dos clientes de reproducción con la misma `ConHora` fija
  (`instanteDePrueba`) y explica en su comentario que el instante es lo único que la grabación no guarda.
- `internal/httpx/grabar.go`: solo los comentarios de `decoradorDeGrabacion` (el del tipo y el del campo
  `siguiente`, que también decía «es el transporte»); ninguna línea de código.

## Evidencia

- Rojo antes de tocar los tests: `TestReplayEsDeterminista` (líneas 157 y 161, `Instante` distinto en cada
  respuesta) y `TestReplayGarantiasVigentes/identifica la petición` (línea 309, bajo la identificación está la marca).
- Sonda de mutación tras el ajuste: `Replay` sin `conMarcaDeEmision` → rojo en «identifica la petición» y en
  `TestInstanteDeEmision/reproduccion`; revertida.
- `misspell` casa «clientes» con *clients*: el comentario del test usa el singular («uno y otro cliente»).
- `go test -race -count=1 -shuffle=on ./internal/httpx/` → ok; `golangci-lint run ./internal/httpx/...` → 0 issues;
  `golangci-lint fmt --diff` sin diferencias.
- `make ci` → código 0, «ci: todos los controles en verde» (`internal/httpx` 97,2 % de cobertura en los dos
  perfiles); `TestContextoCanceladoDuranteLaMigracion` no ha vuelto a fallar en este run.

# T003 · intento 2 · la verificación del workflow sale en rojo: redelimitada otra vez

## Causa

El verificador (`make ci` del paso `verificar`, tras marcar la tarea) falló en `internal/cache` ›
`TestContextoCanceladoDuranteLaMigracion` con `database is locked (5) (SQLITE_BUSY)` en `consultaEntero`
(`migraciones_test.go` 338 ← 217). Es la **tercera** vez en H4 (T001; T003 intento 1; ahora) y el diff de T003 no toca
la caché. Arreglarlo exige tocar `internal/cache`, que la tarea no declaraba: por la regla del paso `reparar`, no se
arregla en este intento; se redelimita y queda `[ ]`. Con la línea nueva, el intento 3 aplica lo que sigue, todo
verificado sobre una copia desechable (`rsync` del árbol a `/tmp/kitlegal-t003`).

## Causa raíz, confirmada

`aplica` (`internal/cache/migraciones.go`) abre la transacción de la migración con `base.BeginTx(ctx, nil)` y el
contexto de quien llama. Cuando ese contexto termina —en el test lo cancela el reloj inyectado desde dentro de la
transacción—, `database/sql` la deshace **él mismo, en su propia goroutine** (`Tx.awaitDone` → `rollback(true)`:
`ROLLBACK` y descarte de la conexión, `sql.go` 2214-2227 y 2331-2360 de go1.27.1). Esa goroutine gana la carrera al
`Rollback` diferido de `aplica` (que recibe `sql.ErrTxDone` y no espera a nadie), `cierraTrasElFallo` cierra una base
sin conexiones libres —la única está en uso en la otra goroutine— y `New` vuelve con la conexión todavía abierta. La
lectura directa del test (`paraLeer`, `mode=ro`, sin `busy_timeout`) llega mientras esa conexión se cierra y SQLite
retira el registro de escritura con el bloqueo exclusivo: SQLITE_BUSY. Incumple el contrato de H3
`specs/004-h3-internal-cache-sqlite/contracts/esquema-y-apertura.md` §5: «Ninguno deja conexión abierta».

Sonda sobre la copia (500 vueltas de `New` cancelado; se mira si `cache.db-wal` existe justo al volver y se lee con
`mode=ro` sin `busy_timeout`): **sin arreglo, `-wal` vivo al volver en 492 de 500 y una lectura fallida con
SQLITE_BUSY; con el arreglo, 0 y 0.**

## Arreglo (para el intento 3, exactamente esto y nada más)

1. `internal/cache/migraciones.go`, solo `empiezaLaTransaccion` (comentario y cuerpo; ninguna otra línea del
   fichero):

   ```go
   // empiezaLaTransaccion abre la transacción inmediata de una migración esperando
   // el bloqueo de escritura por tramos que miran el contexto (FR-003, FR-031).
   //
   // La transacción nace con el contexto sin su cancelación, a propósito: si
   // llevara la de quien llama, database/sql la desharía él mismo en cuanto el
   // contexto terminase, desde su propia goroutine (Tx.awaitDone), y New podría
   // volver con esa conexión todavía abierta —el bloqueo de escritura y los
   // auxiliares del registro de escritura se soltarían después de haber contestado,
   // que es lo que el contrato de apertura prohíbe (§5: ningún fallo deja conexión
   // abierta)—. Sin esa cancelación, deshacerla es cosa de aplica y ocurre antes de
   // volver. El contexto de quien llama sigue gobernando cada sentencia de dentro,
   // que es donde una cancelación se atiende (FR-003); empezar la transacción la
   // atiende entre tramo y tramo, como cualquier otra espera ante bloqueo, y
   // confirmarla ya no la mira: una migración cuya última sentencia entró se
   // confirma entera, que es lo que la siguiente invocación quiere encontrar.
   func (c *Cliente) empiezaLaTransaccion(ctx context.Context, base *sql.DB) (*sql.Tx, error) {
   	var tx *sql.Tx

   	sinCancelacion := context.WithoutCancel(ctx)

   	err := c.reintentaMientrasBloqueada(ctx, func() (err error) {
   		tx, err = base.BeginTx(sinCancelacion, nil)

   		return err
   	})

   	return tx, err
   }
   ```

   Por qué esta forma y no otra: el `BEGIN IMMEDIATE` del controlador (`modernc.org/sqlite@v1.58.0/tx.go` 22-27)
   deja de interrumpirse con el contexto de quien llama, pero cada intento dura como mucho `tramoDeEspera` (100 ms,
   `busy_timeout`) y `reintentaMientrasBloqueada` mira el contexto entre tramos, que es lo que `espera.go` ya documenta
   como retardo máximo de una cancelación; `TestNewConLaBaseBloqueadaRespetaElContexto` (plazo 300 ms, margen 2,5 s)
   sigue en verde. `Commit` y `Rollback` del controlador ya usaban `context.Background()`. Las alternativas
   descartadas: esperar en `cierraTrasElFallo` a que `Stats().OpenConnections` llegue a 0 (sondeo, y sigue habiendo
   una goroutine ajena cerrando); sustituir `sql.Tx` por `sql.Conn` con `BEGIN`/`COMMIT` a mano (deja muerto el
   `_txlock=immediate` del DSN y reescribe `aplica` entera); poner `busy_timeout` o reintentos en `paraLeer` (tapa el
   síntoma y deja la conexión viva al volver, contra el contrato §5).

2. `internal/cache/migraciones_test.go`: importar `io/fs` y, en `TestContextoCanceladoDuranteLaMigracion`, entre la
   aserción `assert.Contains(t, err.Error(), ruta, …)` y la primera `consultaEntero`:

   ```go
   	for _, sufijo := range []string{sufijoRegistroDeEscritura, sufijoMemoriaCompartida} {
   		_, err := os.Stat(ruta + sufijo)
   		require.ErrorIs(t, err, fs.ErrNotExist,
   			"al volver no queda ninguna conexión: la transacción se deshizo antes de contestar, no en otra goroutine")
   	}
   ```

   Es la misma comprobación que ya hace `TestNewConLaBaseBloqueadaRespetaElContexto` para el otro camino de fallo.
   Rojo sin el arreglo (falla en esa `require`, no en la lectura directa: 5 de 5 en la copia), verde con él. Nada más
   cambia en el test: la lectura directa sigue como está, sin `busy_timeout` ni reintentos, y con la conexión ya
   cerrada al volver no tiene con qué chocar.

## Evidencia sobre la copia (todo con la línea nueva de T003, sin tocar el repositorio)

- Sonda de 500 vueltas: 492 → 0 con `-wal` vivo al volver; 1 → 0 lecturas fallidas.
- `go test -race -count=20 -run '^TestContextoCanceladoDuranteLaMigracion$' ./internal/cache/` → ok.
- Mutación (`BeginTx(ctx, nil)` de nuevo): `-count=5` → 5 fallos en `migraciones_test.go` 220, el mensaje de arriba;
  restaurado.
- `go test -race -count=3 -shuffle=on ./internal/cache/` → ok; `go test -race -tags=integration ./internal/cache/` → ok.
- `golangci-lint run ./internal/cache/...` → 0 issues; `golangci-lint fmt --diff` sin diferencias.
- `go clean -testcache && make -C /tmp/kitlegal-t003 ci` con los dos cambios → código 0, «ci: todos los controles en
  verde» (`internal/cache` 90,7 % en el perfil unitario y 96,2 % en el de integración; `internal/httpx` 97,2 % en los
  dos).

## Lo que hereda el intento 3

- El trabajo de `httpx` del intento 2 (arriba) está entero; según cómo salga `verificar_reparacion` estará commiteado
  como `feat(H4): T003` o seguirá en el árbol. En los dos casos solo falta la parte de la caché: aplicar los dos
  cambios tal cual, comprobar rojo → verde, `make ci` en verde y marcar `[X]`.
- `tasks.md`: la línea de T003 declara ahora `internal/cache/migraciones.go` e `internal/cache/migraciones_test.go`
  (sin nombrar el directorio entero, para que el guardián no lo abra), y la nota «Ficheros de H0-H3 que se tocan» los
  lista.
- Es el último intento antes de la parada para revisión humana (`siguiente_tarea` para con más de tres).
- Hueco del workflow, para después de H4: si `verificar_reparacion` sale en verde con la tarea redelimitada
  (`[ ]`), `cerrar_tarea` la commitea igual, porque `redelimitada_reparacion` solo se calcula en rojo y
  `redelimitada` se calculó antes de `reparar`.

# T003 · intento 3 · en verde y marcada

El trabajo de `httpx` del intento 2 estaba entero y commiteado como `feat(H4): T003` (c5f7f3d, base de este intento;
`internal/` sin diferencias frente a ella). El intento 3 aplica solo la parte de la caché, exactamente como la dejó
verificada el intento 2 y nada más:

- `internal/cache/migraciones_test.go`: importa `io/fs` y, en `TestContextoCanceladoDuranteLaMigracion`, entre la
  aserción que exige la ruta en el mensaje y la primera lectura directa, exige con `require.ErrorIs(…, fs.ErrNotExist)`
  que no existan `-wal` ni `-shm` (`sufijoRegistroDeEscritura`, `sufijoMemoriaCompartida`). La lectura directa sigue
  como estaba: sin `busy_timeout` ni reintentos.
- `internal/cache/migraciones.go`: solo `empiezaLaTransaccion` (comentario y cuerpo). `BeginTx` recibe
  `context.WithoutCancel(ctx)`; `reintentaMientrasBloqueada` sigue mirando el contexto de quien llama entre tramos, y
  cada sentencia de `aplica` sigue con ese contexto.

## Evidencia

- Rojo antes del arreglo, en el repositorio: `go test -race -count=5 -run '^TestContextoCanceladoDuranteLaMigracion$'
  ./internal/cache/` → 5 de 5 fallos en `migraciones_test.go` 220 («al volver no queda ninguna conexión…»): el `-wal`
  sigue vivo al volver `New`.
- Verde tras el arreglo: el mismo test con `-count=20` → ok; `go test -race -count=3 -shuffle=on ./internal/cache/`
  → ok; `go test -race -tags=integration ./internal/cache/` → ok; `TestNewConLaBaseBloqueadaRespetaElContexto` con
  `-count=3` → ok (el plazo entre tramos sigue respetándose).
- `golangci-lint run ./internal/cache/...` (el fijado por el repo) → 0 issues; `golangci-lint fmt --diff` sin
  diferencias.
- `go clean -testcache && make ci` → código 0, «ci: todos los controles en verde» (`internal/cache` 90,7 % en el perfil
  unitario y 96,2 % en el de integración; `internal/httpx` 97,2 % en los dos). `TestContextoCanceladoDuranteLaMigracion`
  no ha fallado en ninguna ejecución de este intento.
