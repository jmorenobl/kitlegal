# Evidencia de plataforma de H5 (T030)

Intento 2 de T030, 2026-09-15. Rama `h5-skill-boe-legislacion`, cabeza `417635e39abf5807f8600b79debd6026fc320ce8`
(`feat(H5): T035`), publicada por avance rápido sobre la del intento 1 (`6d68c21`, `feat(H5): T029`) con los tres
commits de T033, T034 y T035. Propuesta de cambio: [#27](https://github.com/jmorenobl/kitlegal/pull/27), la misma del
intento 1. Todas las órdenes, tal cual; las que llevan `rtk proxy` delante, con el paso directo (el envoltorio de la
terminal reescribe salidas).

## 1. Publicación

`git push -u origin h5-skill-boe-legislacion` (el gancho `pre-push` `guardia-push` la deja pasar):

```text
To github.com:jmorenobl/kitlegal.git
   6d68c21..417635e  h5-skill-boe-legislacion -> h5-skill-boe-legislacion
branch 'h5-skill-boe-legislacion' set up to track 'origin/h5-skill-boe-legislacion'.
```

`gh pr view h5-skill-boe-legislacion`, antes del push y después: la propuesta #27 existe (`OPEN`, base `main`,
`mergeable: MERGEABLE`), así que **no se crea ninguna**. `--json number,headRefOid,labels` tras el push:

```json
{"headRefOid":"417635e39abf5807f8600b79debd6026fc320ce8","labels":[],"number":27}
```

La rama principal remota sigue en `b7eda6e` (workflow 1.9.1, #26), como en el intento 1: no se movió.

## 2. Checks

`gh pr checks h5-skill-boe-legislacion --watch --interval 20` (últimas líneas; el `--watch` terminó con `ci` y sin
ningún estado de Codecov todavía, como advierte la tarea):

```text
ci	pending	0	https://github.com/jmorenobl/kitlegal/actions/runs/34929854868/job/104255634243
ci	pass	4m49s	https://github.com/jmorenobl/kitlegal/actions/runs/34929854868/job/104255634243
código 0
```

`gh run list --branch h5-skill-boe-legislacion`, antes de poner ninguna etiqueta (las dos filas antiguas son las del
intento 1):

```text
completed	success	feat(H5): skill boe-legislacion y andamiaje de skills	ci	h5-skill-boe-legislacion	pull_request	34929854868	4m54s	2026-09-15T04:42:23Z
completed	failure	feat(H5): skill boe-legislacion y andamiaje de skills	evals	h5-skill-boe-legislacion	pull_request	34922606273	1m37s	2026-09-15T02:48:53Z
completed	success	feat(H5): skill boe-legislacion y andamiaje de skills	ci	h5-skill-boe-legislacion	pull_request	34922178932	4m38s	2026-09-15T02:42:06Z
```

Check-runs del commit de la cabeza por la API
(`gh api repos/jmorenobl/kitlegal/commits/417635e39abf5807f8600b79debd6026fc320ce8/check-runs --jq '.check_runs[] | "\(.name)\t\(.status)\t\(.conclusion)\t\(.completed_at)\t\(.output.title)"'`).
Primera lectura, nada más terminar el `--watch`: `codecov/patch` todavía `in_progress` y solo `codecov/project`
publicado; segunda lectura, unos segundos después, ya con los cuatro:

```text
codecov/project/internal/cli	completed	success	2026-09-15T04:47:34Z	98.09% (target 90.00%)
codecov/project/internal/core	completed	success	2026-09-15T04:47:33Z	90.36% (target 85.00%)
codecov/patch	completed	success	2026-09-15T04:47:32Z	98.42% of diff hit (target 94.70%)
codecov/project	completed	success	2026-09-15T04:47:31Z	96.30% (target 70.00%)
ci	completed	success	2026-09-15T04:47:16Z	null
```

Estados del commit (`gh api repos/jmorenobl/kitlegal/commits/417635e39abf5807f8600b79debd6026fc320ce8/status --jq '.state, (.statuses[] | "\(.context)\t\(.state)\t\(.description)")'`):

```text
pending
```

Ningún estado, como en el intento 1: Codecov publica check-runs, no estados de commit, y la API combina una lista vacía
como `pending`. No falta ningún estado de plataforma: los cuatro de Codecov están como check-runs, con cifra y objetivo.

Tras la prueba de red (§12.2), la lista de check-runs de la misma cabeza añade `evals	completed	failure`
(ejecución 34930222593, `gates/prueba-de-red.md`).

## 3. Lectura

| Estado | Resultado | Objetivo | Veredicto |
|---|---|---|---|
| `ci` | `success` (4 m 49 s) | — | verde |
| `codecov/project` | 96,30 % | 70 % | verde |
| `codecov/project/internal/core` | 90,36 % | 85 % | verde |
| `codecov/project/internal/cli` | 98,09 % | 90 % | verde |
| `codecov/patch` | 98,42 % del diff | 94,70 % (`auto`: la cobertura de la base `90c3637`) | **verde** (en el intento 1, 93,54 %: rojo) |
| `evals` (prueba de red, §12.2) | `failure` | — | **rojo**: `gates/prueba-de-red.md` |

**`codecov/patch` pasa de rojo a verde con T034 y T035, sin tocar ningún umbral** (`codecov.yml` no está en el diff
frente a `main`). El comentario de `codecov[bot]` (actualizado 2026-09-15T04:47:34Z; el `:x:` inicial es el icono que
el comentario pone siempre que queda alguna línea sin cubrir, no un estado en rojo: el check-run es `success`):

```text
:x: Patch coverage is `98.42332%` with `33 lines` in your changes missing coverage. Please review.
:white_check_mark: Project coverage is 96.30%. Comparing base (`90c3637`) to head (`417635e`).

| Files with missing lines | Patch % | Lines |
| internal/skills/sincronia.go | 94.97% | 12 Missing |
| internal/evals/preparar.go   | 94.05% | 6 Missing  |
| internal/evals/trazas.go     | 98.56% | 5 Missing  |
| internal/skills/esquemas.go  | 97.67% | 4 Missing  |
| internal/evals/informe.go    | 99.20% | 2 Missing  |
| internal/evals/conjunto.go   | 99.34% | 1 Missing  |
| internal/evals/formato.go    | 93.75% | 1 Missing  |
| internal/skills/normas.go    | 98.93% | 1 Missing  |
| internal/skills/skill.go     | 96.15% | 1 Missing  |

@@            Coverage Diff             @@
##             main      #27      +/-   ##
+ Coverage   94.70%   96.30%   +1.60%
  Files          62       80      +18
  Lines        2831     4924    +2093
+ Hits         2681     4742    +2061
- Misses        150      182      +32
```

(Tabla abreviada: enlaces retirados.) Las 33 líneas que quedan son los bloques justificados uno a uno en
`gates/tarea-T034.md` (16, de una línea) y `gates/tarea-T035.md` (18, de una sentencia; Codecov cuenta 17 líneas en
`internal/skills` porque dos sentencias comparten línea): errores del registro de applets, de las banderas globales, de
los esquemas incrustados y de la codificación de tipos fijos, y lecturas o retiradas de lo que la propia función acaba
de listar o crear. La cobertura global sube del 94,22 % del intento 1 al 96,30 %, por encima de la base.

Con `ci` y los cuatro estados de Codecov en verde, la tarea sigue con quickstart §12.1 y §12.2. **La prueba de red
descubre un defecto del job** (todas las sesiones terminan con código 1 al arrancar Claude Code), así que T030 no se
marca: diagnóstico y arreglo en `gates/tarea-T030.md` (tarea nueva T036).
