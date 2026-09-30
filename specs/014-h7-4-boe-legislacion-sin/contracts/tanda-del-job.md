# Contrato: una tanda por commit y skill

`.github/workflows/evals.yml`, `internal/evals/tanda.go` (nuevo), el punto de entrada `TestTandaDelCommit` y
`TestDefinicionDelJob` (research D15, D16, D17). `scripts/workflow/` no cambia (FR-072).

## 1. La definición del job

Lo que cambia; lo demás (`on`, `cambios`, la matriz, `env`, los pasos de `evals`) queda como hoy.

```yaml
  # Una tanda por commit (FR-070 y FR-071 de H7.4): solo corre en una ejecución que quiere medir —el despacho, la
  # etiqueta evals o evals-prueba-de-red, o la apertura que toca lo que las evals miden— y decide si mide: no mide si una
  # ejecución anterior del flujo sobre el mismo commit, sin terminar, ya mide. Espera, como mucho unos segundos, a que
  # decidan las anteriores; nunca a su tanda. Su último paso es la marca que leen las posteriores. Sin concurrency: ni
  # queda en espera ni se cancela, y no deja nunca una comprobación roja por decidir no medir.
  tanda:
    needs: [cambios]
    if: >-
      !cancelled() && needs.cambios.result != 'failure' && (
        github.event_name == 'workflow_dispatch' ||
        github.event.label.name == 'evals' ||
        github.event.label.name == 'evals-prueba-de-red' ||
        needs.cambios.outputs.coincide == 'si'
      )
    runs-on: ubuntu-24.04
    timeout-minutes: 15
    permissions:
      contents: read
      actions: read
    outputs:
      medir: ${{ steps.decidir.outputs.medir }}
    steps:
      - name: Obtener el código del commit evaluado
        uses: actions/checkout@v7
        with:
          ref: ${{ github.event.pull_request.head.sha || github.sha }}

      - name: Instalar Go y restaurar la caché
        uses: actions/setup-go@v7
        with:
          go-version-file: go.mod
          cache: true
          cache-dependency-path: |
            go.sum
            tools/*/go.sum

      # -timeout 12m: la decisión espera como mucho 10 min, más sus consultas, y con los 10 min de go test por omisión
      # el test acabaría en pánico antes de medir tras la espera; con la preparación, cabe en los 15 del trabajo.
      - name: Mirar si otra tanda mide este commit
        id: decidir
        env:
          GH_TOKEN: ${{ secrets.GITHUB_TOKEN }}
          GH_REPO: ${{ github.repository }}
          COMMIT_EVALUADO: ${{ github.event.pull_request.head.sha || github.sha }}
          EJECUCION: ${{ github.run_id }}
        run: >-
          go test -tags evals -count=1 -timeout 12m -v -run '^TestTandaDelCommit$' ./internal/evals/
          -args -commit "$COMMIT_EVALUADO" -ejecucion "$EJECUCION" -salida "$GITHUB_OUTPUT"

      - name: Esta ejecución mide el commit
        if: steps.decidir.outputs.medir == 'si'
        run: echo "La tanda de este commit la mide esta ejecución."

  evals:
    name: evals (${{ matrix.skill }})
    needs: [tanda]
    # !cancelled(): sin una función de estado, el if lleva un success() implícito que podría saltar el trabajo por
    # cambios, saltado en la etiqueta y en el despacho (research S9; el mismo motivo que el if de hoy, H5).
    if: ${{ !cancelled() && needs.tanda.outputs.medir == 'si' }}
    # concurrency, strategy, runs-on, permissions, env y steps: como hoy.
    timeout-minutes: 122
```

- El `if` de `tanda` es el que hoy lleva `evals`: una ejecución que no quiere medir (otra etiqueta, una apertura que no
  toca lo que las evals miden) salta `tanda`, y con ella `evals`.
- `evals` salta entero sin `medir=si` (también si `tanda` se saltó o falló): deja la comprobación
  `evals (${{ matrix.skill }})` en `skipping` (research S3, O2), que el cierre ni cuenta como roja ni lee como informe.
- La `concurrency` de `evals` por commit y skill con `cancel-in-progress: false` se queda (SC-010: «como mucho una a la
  vez»).
- `timeout-minutes: 122`: el peor caso de `boe-legislacion` con 97 sesiones es 7 285 s (research V13).
- `-timeout 12m` en el paso `decidir` (T008): la decisión espera como mucho 10 min más sus consultas, y con los 10 min
  de `go test` por omisión el test acabaría en pánico antes de medir tras la espera, con `tanda` en rojo (FR-070); con la
  preparación, cabe en los 15 del trabajo.

## 2. La decisión (`internal/evals/tanda.go`)

```go
// ejecucionDelCommit es una ejecución del flujo evals sobre el commit evaluado.
type ejecucionDelCommit struct {
	id        int64 // databaseId
	terminada bool  // status == "completed"
	tanda     estadoDeLaTanda
}

type estadoDeLaTanda int // tandaSinDecidir, tandaQueMide, tandaQueNoMide

type decisionDeLaTanda struct {
	mide       bool
	pendientes []int64 // anteriores sin terminar que aún no han decidido
}

// decisionTrasLaEspera es la de la última consulta del bucle; agotada, si mide por agotar la espera.
type decisionTrasLaEspera struct {
	decisionDeLaTanda
	agotada bool
}

func (e ejecucionDelCommit) cuentaPara(propia int64) bool // e.id < propia && !e.terminada

func decidirLaTanda(propia int64, ejecuciones []ejecucionDelCommit) decisionDeLaTanda

func esperarLaDecision(ctx context.Context, propia int64,
	consultar func(context.Context) ([]ejecucionDelCommit, error), ahora func() time.Time,
	esperar func(context.Context, time.Duration) error,
) (decisionTrasLaEspera, error)

func consultarLasEjecuciones(ctx context.Context, propia int64, listar func(context.Context) ([]byte, error),
	verLosTrabajos func(context.Context, int64) ([]byte, error),
) ([]ejecucionDelCommit, error)
```

Todo sin exportar: solo lo usan `TestTandaDelCommit` y los tests del paquete.

- Solo cuentan (`cuentaPara`) las ejecuciones con `id < propia` y sin terminar.
- Si alguna tiene `tandaQueMide`: no mide, sin pendientes.
- Si no, y alguna tiene `tandaSinDecidir`: mide, con esas como pendientes (el bucle vuelve a consultar).
- Si no: mide, sin pendientes.

**El estado de la tanda de una ejecución**, de su `gh run view <id> --json jobs`: `tandaQueMide` si su trabajo `tanda`
está `completed` y su paso «Esta ejecución mide el commit» tiene `conclusion` `success`; `tandaSinDecidir` si no tiene
trabajo `tanda` o no está `completed`; `tandaQueNoMide` en cualquier otro caso (saltado, fallido, cancelado, o la marca
saltada). La `conclusion` del trabajo no se lee: saltado, fallido o cancelado, su marca no está en `success`. El nombre
del trabajo y el del paso son constantes del paquete (`trabajoDeLaTanda` y `marcaDeLaTanda`), las mismas que
`TestDefinicionDelJob` exige en la definición (§4).

**La consulta** (`consultarLasEjecuciones`): lee la lista de `gh run list` y, solo de las ejecuciones que cuentan
(`cuentaPara`), el estado de su tanda de `gh run view`; la de las demás no se lee. `TestTandaDelCommit` le da las dos
órdenes de abajo, y los tests, el JSON sintético que daría `gh`.

**El bucle** (`esperarLaDecision`): consulta, decide y, con pendientes, espera 10 s (`esperaEntreConsultas`) y
vuelve a consultar; pasados 10 min (`esperaMaximaDeLaDecision`, contados desde la primera consulta e inclusive: la
consulta del minuto 10 ya no espera) con alguna pendiente, mide (`agotada`) y lo dice en el registro (la `concurrency`
la pondría detrás de la otra, como antes de H7.4). Un error de una consulta es un error que la nombra (`la consulta <n>
de las ejecuciones del commit`): el de `gh` (la orden termina con otro código que `0`) nombra la orden, su variable,
cómo terminó y lo que escribió en sus dos salidas; el de unos trabajos, la ejecución; y un JSON de `gh` que no se puede
leer también es un error. Con cualquiera, el paso falla (defecto `inesperado`).

**Las órdenes** (research D16), con `exec.CommandContext(ctx, "sh", "-c", <constante>)` y los valores en el entorno de
la orden (el de la ejecución, `os.Environ()`, más la variable de la orden, que sustituye a la del mismo nombre):

```sh
exec gh run list --workflow evals.yml --commit "$COMMIT_EVALUADO" --limit 100 --json databaseId,status
exec gh run view "$EJECUCION_ANTERIOR" --json jobs
```

## 3. El punto de entrada `TestTandaDelCommit` (etiqueta `evals`, `internal/evals/job_test.go`)

Banderas: `-commit` (la de hoy), `-ejecucion` y `-salida`, las tres obligatorias. Consulta con las órdenes de §2,
decide con el bucle, registra con `t.Log` qué ejecuciones anteriores miró en cada consulta y por qué mide o no, y añade
a `-salida` una línea `medir=si` o `medir=no` (detrás de lo que ya tenga, creándolo si no existe). Falla solo con un
error: una bandera que falta o un `-ejecucion` que no es un entero, el de una consulta (§2) o el de escribir en
`-salida`, que la nombra. No abre sesiones, no escribe fuera de `-salida` y no llega a la red salvo por `gh`.

## 4. `TestDefinicionDelJob` (FR-100; SC-010)

`leerDefinicionDelJob` lee además `jobs.tanda` (`name`, `if`, `permissions`, `concurrency`, `outputs`, `timeout-minutes` y los
pasos con su `id`, su `name`, su `if` y su `run`) y `needs` e `if` de `jobs.evals`. Subpruebas:

- **`del-repositorio`** (la de hoy, más): una línea por clave que no es la del contrato:
  `jobs.tanda` está; `jobs.tanda.name` no está o es `tanda` (gh da el trabajo por su `name` y, sin él, por su id, y
  la decisión lo busca por `tanda`); `jobs.tanda.if` es el de §1; `jobs.tanda.concurrency` no está; `jobs.tanda.permissions.actions` es `read`;
  `jobs.tanda.outputs.medir` es `${{ steps.decidir.outputs.medir }}`; el paso `decidir` ejecuta
  `-run '^TestTandaDelCommit$'` con `-commit`, `-ejecucion` y `-salida "$GITHUB_OUTPUT"` (su `run` entero, el de §1:
  sin `-tags evals`, `go test` no encontraría el test y terminaría con 0 sin escribir `medir`); el último paso se llama como
  la marca de §2 y su `if` es `steps.decidir.outputs.medir == 'si'`; `jobs.evals.needs` es `[tanda]` y `jobs.evals.if`,
  `${{ !cancelled() && needs.tanda.outputs.medir == 'si' }}`; ni `concurrency` de flujo ni `cancel-in-progress: true` (hoy); y el tope cubre el
  peor caso (hoy).
- **`sinteticas`** (la de hoy, más): cada definición que se aparta en una sola de esas claves da la línea que la nombra
  —sin `tanda` (`sin-tanda`), con un `name` en `tanda` que no es su id (`tanda-con-nombre`), con otro `if` en `tanda`
  (`tanda-sin-la-prueba-de-red`), con `concurrency` en `tanda` (`tanda-con-concurrency`), sin `actions: read`
  (`tanda-sin-actions-read`), sin la salida `medir` (`tanda-sin-la-salida-medir`), con otro `run` en `decidir` o sin el
  paso de id `decidir` (`decidir-sin-la-ejecucion`, `decidir-sin-su-id`), con la marca con otro nombre o sin su `if`
  (`marca-con-otro-nombre`, `marca-sin-su-if`), con `evals` sin `needs: [tanda]` o con otro `if`
  (`evals-sin-needs-tanda`, `evals-con-otro-if`), con `cancel-in-progress: true` y con `concurrency` de flujo (las de
  hoy)—; y `tanda-con-su-id-como-nombre` (`name: tanda`) no da ninguna.
- **`segundo-disparo`** (nueva), `esperarLaDecision` con `consultarLasEjecuciones` sobre consultas sintéticas —el JSON
  que darían `gh run list` y `gh run view`, que recorre también sus dos lectores— y un reloj que no duerme; cada caso
  fija la decisión, las consultas y una espera de 10 s entre cada dos:

| Caso | Ejecuciones (propia = 20) | Esperado |
|---|---|---|
| `anterior-que-mide-corre` | 10 sin terminar, `tandaQueMide` | no mide |
| `anterior-que-mide-espera` | 10 sin terminar, `tandaQueMide` (su `evals` esperando) | no mide |
| `anterior-sin-decidir-que-mide` | 10 sin terminar, `tandaSinDecidir` y, en la consulta siguiente, `tandaQueMide` | consulta dos veces; no mide |
| `anterior-sin-decidir-que-no-mide` | 10 sin terminar, `tandaSinDecidir` y después `tandaQueNoMide` | mide |
| `anterior-que-no-mide` | 10 sin terminar, `tandaQueNoMide` | mide |
| `anterior-terminada` | 10 terminada, `tandaQueMide` | mide (FR-071: la etiqueta vuelve a medir) |
| `posterior-que-mide` | 30 sin terminar, `tandaQueMide` | mide (solo cuentan las anteriores) |
| `sola` | ninguna otra | mide |
| `espera-agotada` | 10 sin terminar, siempre `tandaSinDecidir` | mide al agotar la espera: 61 consultas, con 10 pendiente |

`anterior-que-mide-espera` se distingue de `anterior-que-mide-corre` por lo que los distingue en `gh`: sus trabajos
`evals` en `pending` y no en `in_progress`. Tras la tabla, `consulta-que-falla`: una consulta que falla da un error que
la nombra (`la consulta 1 de las ejecuciones del commit`) y envuelve el suyo.

Y **`estado-de-la-tanda`**: el estado leído de un `gh run view --json jobs` sintético —con la marca en `success`; con la
marca `skipped`; con `tanda` `in_progress`; sin `tanda`; con `tanda` `skipped`—.

## 5. Uso, de fuera adentro

| Salida | Quién, cuántas veces | Tamaño | Cuándo se apaga |
|---|---|---|---|
| La comprobación `tanda` | el cierre (`medir` la ve `pending` mientras decide y `pass` después); una por disparo que quiere medir | una comprobación | la de cada disparo termina en segundos; no se repite |
| Las consultas a la API de `gh` | `TestTandaDelCommit`, una vez por disparo: 1 `run list` y 1 `run view` por anterior sin terminar, por consulta | ≈ 100 B por ejecución listada y ≈ 2-5 KB por `run view` (los pasos de sus trabajos); acotado por las ejecuciones del commit, no por el historial del repositorio (`--commit`, `--limit 100`) | termina al decidir: en la mayoría de los disparos, una consulta; con una anterior aún en `cambios`, dos o tres |
| El tiempo antes de las sesiones | la tanda que mide | + la preparación del trabajo `tanda` (checkout, Go con la caché y la compilación del paquete de evals), fuera de `duracion_de_las_sesiones`, que cuenta desde la primera sesión; no se mide aquí: lo da el registro del job de cierre | — |
