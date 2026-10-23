# Contrato: el job de evals (`.github/workflows/evals.yml`) con elección y reutilización

Sustituye en lo que dice a los contratos del job de H7.3, H7.4, H24 y H25 que cita la cabecera de `evals.yml`; lo demás
(instalación, retirada de Python, sesiones, juez, medida del juez, topes) sigue como está. Decisiones en
[../research.md](../research.md) D1, D8, D13, D17 y D18.

## 1. Disparadores y modo

| Disparador | Modo | Qué hace |
|---|---|---|
| Abrir o reabrir una propuesta (`pull_request` `opened`/`reopened`) | `reutilizacion` | El trabajo `cambios` calcula la elección; sin ninguna eval elegida no corre nada más (FR-016). Con alguna, mide lo elegido y reutiliza lo demás |
| Etiqueta `evals` | `reutilizacion` | Igual, aunque la propuesta no elija ninguna (FR-017); es la que pone el cierre del workflow |
| Etiqueta `evals-completa` | `completa` | Las evals de cada skill, todas medidas (FR-018) |
| Despacho manual sin `medir_al_juez` | `completa` | Igual, sobre la rama o etiqueta elegida |
| Etiqueta `evals-prueba-de-red`, o despacho con `prueba_de_red` | `completa` | Igual, con la sesión de prueba de red de la eval 01 de `boe-legislacion`, como hoy |
| Etiqueta `evals-medir-juez`, o despacho con `medir_al_juez` | — | La medida del juez (trabajo `medida`), sin evals, como hoy |

El modo lo fija el `env` del trabajo `tanda`:
`MODO_DE_EVALS: ${{ (github.event_name == 'workflow_dispatch' || github.event.label.name == 'evals-completa' ||
github.event.label.name == 'evals-prueba-de-red') && 'completa' || 'reutilizacion' }}`.

## 2. Trabajos

### `cambios` (solo al abrir o reabrir)

- `timeout-minutes: 15`; `permissions: contents: read, pull-requests: read`; `outputs: elige`.
- Pasos: `actions/checkout@v7` con `ref` la cabeza y `fetch-depth: 0`; `actions/setup-go@v7` como hoy; «Elegir las
  evals que la propuesta ha podido cambiar», con `GH_TOKEN`, `GH_REPO`, `PROPUESTA` y `BASE_DE_LA_PROPUESTA`
  (`github.event.pull_request.base.sha`):
  `go test -tags evals -count=1 -timeout 12m -v -run '^TestEleccionDeLaPropuesta$' ./internal/evals/ -args -propuesta
  "$PROPUESTA" -base "$BASE_DE_LA_PROPUESTA" -salida "$GITHUB_OUTPUT"`.
- Escribe `elige=si` o `elige=no`, y en el registro la elección con su motivo
  ([eleccion-y-huella.md](eleccion-y-huella.md) §3). Un error del cálculo de los conjuntos elige todas (FR-005): `si`.
  Si la orden de `gh` que da los ficheros de la propuesta falla, también elige todas, con `plataforma`: `si`, y el
  registro dice la orden y su salida de error (research D17).

### `tanda`

- `if`: `!cancelled() && needs.cambios.result != 'failure' && ((github.event_name == 'workflow_dispatch' &&
  inputs.medir_al_juez != true) || github.event.label.name == 'evals' || github.event.label.name == 'evals-completa' ||
  github.event.label.name == 'evals-prueba-de-red' || needs.cambios.outputs.elige == 'si')`.
- `outputs: medir, modo`. `TestTandaDelCommit` recibe `-modo "$MODO_DE_EVALS"` y escribe `medir=si|no` y `modo=<modo>`.
- Marcas: «Esta ejecución mide el commit» (si mide) y, nueva, «Esta ejecución mide la suite completa» (si mide y el modo
  es `completa`). Una ejecución no mide si una anterior sin terminar sobre el mismo commit mide con un modo igual o mayor
  (`completa` > `reutilizacion`) (FR-021; data-model §10).

### `evals (<skill>)`

Matriz, concurrencia por skill (5, 4 y 2), grupo de `concurrency`, modelos, repeticiones, umbral y tope (352 min): como
hoy (FR-020). Cambia:

- `permissions: contents: read, actions: read, pull-requests: read`.
- `env` añade `MODO_DE_EVALS: ${{ needs.tanda.outputs.modo }}`, `PROPUESTA: ${{ github.event.pull_request.number }}` y
  `BASE_DE_LA_PROPUESTA: ${{ github.event.pull_request.base.sha }}`.
- El checkout lleva `fetch-depth: 0`.
- Paso nuevo, tras instalar Go: «Decidir qué mide y qué reutiliza», con `GH_TOKEN` y `GH_REPO` solo en él:
  `go test -tags evals -count=1 -timeout 15m -v -run '^TestDecisionDelTrabajo$' ./internal/evals/ -args -skill
  "$SKILL_EVALUADA" -modo "$MODO_DE_EVALS" -commit "$COMMIT_EVALUADO" -propuesta "$PROPUESTA" -base
  "$BASE_DE_LA_PROPUESTA" -modelo-que-decide "$MODELO_DE_EVALS" -version-de-claude-code "$VERSION_DE_CLAUDE_CODE"
  -modelo-del-juez "$MODELO_DEL_JUEZ" -version-del-juez "$VERSION_DE_CLAUDE_CODE_DEL_JUEZ" -salida
  "$RUNNER_TEMP/decision-de-evals.json"`. Escribe el bloque `origen` sin `sesiones_abiertas`
  ([informe-del-job.md](informe-del-job.md) §2) y, en el registro, una línea por eval: medida con su motivo o
  reutilizada con su procedencia. Lo que lee de la plataforma —los ficheros de la propuesta y la búsqueda de medidas de
  `main`— tiene un plazo de 10 minutos. Si `gh` no da los ficheros de la propuesta, elige todas las evals de la skill
  con `plataforma` y no busca; en la búsqueda, la primera orden de `gh` que falla, o el plazo agotado, la para y deja
  sin resultado lo que falte, que se mide (FR-013). En los dos casos, la orden y su salida
  de error (o el plazo agotado, con la orden en curso) van al registro, antes de las líneas de las evals, y a
  `origen.error_de_la_plataforma` del fichero de la decisión, que el informe copia (research D17).
- «Ejecutar las evals»: `DECISION_DE_EVALS="$RUNNER_TEMP/decision-de-evals.json" make evals SKILL="$SKILL_EVALUADA" 2>&1`.
- Paso nuevo, el último: «El informe del trabajo está escrito», con `if: ${{ !cancelled() }}`:
  `test -s "$RUNNER_TEMP/kitlegal-evals-$SKILL_EVALUADA/informe.json"`. Es la marca que lee la búsqueda de medidas de
  `main` (FR-010; research D8).
- El peor caso del tope suma los 15 minutos de la decisión: `boe-legislacion`, 18 977 + 900 = 19 877 s (331,3 min), en
  los 352 de hoy; `TestDefinicionDelJob` lo recalcula.

### `medida`

Sin cambios.

## 3. `scripts/evals.sh`

`DECISION_DE_EVALS` es obligatoria: sin ella, o si no es un fichero legible, termina con 1 antes de la primera sesión
(«evals: falta DECISION_DE_EVALS»). La pasa a `TestEjecucionDelJob` como `-decision`, que abre solo las sesiones de las
evals medidas de la decisión y compone el informe con las reutilizadas.

## 4. Puntos de entrada (etiqueta `evals`, fuera de `make ci`)

| Test | Banderas | Salida |
|---|---|---|
| `TestEleccionDeLaPropuesta` | `-propuesta`, `-base`, `-salida` | `elige=si\|no` en `-salida`; la elección en el registro |
| `TestTandaDelCommit` | las de hoy y `-modo` | `medir=si\|no` y `modo=<modo>` |
| `TestDecisionDelTrabajo` | `-skill`, `-modo`, `-commit`, `-propuesta`, `-base`, `-modelo-que-decide`, `-version-de-claude-code`, `-modelo-del-juez`, `-version-del-juez`, `-salida` | el fichero de la decisión |
| `TestEjecucionDelJob` | las de hoy y `-decision` | `informe.md` e `informe.json` |
| `TestSuiteCompletaDelCommit` | `-commit`, `-etiqueta` | sale bien o falla con el mensaje de [release.md](release.md) §2 |

## 5. Órdenes de la plataforma

Todas de lectura, por `gh`, con texto constante y cada valor en una variable del entorno (research V7, D15):

- `exec gh api --paginate "repos/{owner}/{repo}/pulls/$PROPUESTA/files" --jq '.[] | [.filename, (.previous_filename // "")] | @tsv'`
- `exec gh api "repos/{owner}/{repo}/pulls?state=closed&base=main&sort=updated&direction=desc&per_page=100" --jq '.[] | select(.merged_at != null) | [.number, .head.sha, .merged_at] | @tsv'`
- `exec gh run list --workflow evals.yml --limit 200 --json databaseId,headSha,status,conclusion,createdAt`
- `exec gh run list --workflow evals.yml --commit "$COMMIT_EVALUADO" --limit 100 --json databaseId,headSha,status,conclusion,createdAt` (la release; la tanda sigue con la suya, que pide solo `databaseId,status`)
- `exec gh run view "$EJECUCION" --json jobs` (la de la tanda)
- `exec gh run view "$EJECUCION" --job "$TRABAJO" --log`, del que se quitan trabajo, paso, hora y marca de orden de
  bytes y se toma lo que hay entre las marcas del informe, como `informe_del_trabajo` de `scripts/workflow/comun.sh`.

## 6. Uso

| Salida | Quién la lee y cuántas veces | Tamaño | Cuándo deja de darse |
|---|---|---|---|
| Elección en el registro de `cambios` | quien revisa, al abrir la propuesta | hasta 45 líneas de unos 80 bytes hoy (§3 de [eleccion-y-huella.md](eleccion-y-huella.md)) | al fusionar |
| Decisión en el registro de cada trabajo | quien revisa, una vez por ejecución | una línea por eval de la skill, unos 120 bytes | en cada ejecución nueva |
| Error de la plataforma en el registro de `cambios` o de la decisión | quien revisa, en la ejecución en que falla una orden de `gh` o se agota el plazo | la orden y su salida de error, unas líneas; uno como mucho por trabajo | en la primera ejecución en que la plataforma responde |
| Búsqueda de medidas de `main` | el propio trabajo, una vez por ejecución con reutilización y evals sin elegir | la normal, 1 informe de la skill (como mucho 922 KiB hoy) más su registro; como mucho 10 informes por trabajo, uno por medida recorrida | para en cuanto cada eval no elegida tiene resultado |
