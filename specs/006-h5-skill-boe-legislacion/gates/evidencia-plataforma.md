# Evidencia de plataforma de H5 (T030)

Intento 1 de T030, 2026-09-15. Rama `h5-skill-boe-legislacion`, cabeza `6d68c21be163d2f860eefbe1c7e1871f33ba701c`
(`feat(H5): T029`). Propuesta de cambio: [#27](https://github.com/jmorenobl/kitlegal/pull/27). Todas las órdenes, tal
cual; las que llevan `rtk proxy` delante, con el paso directo (el envoltorio de la terminal reescribe salidas).

## 1. Publicación

`git push -u origin h5-skill-boe-legislacion` (el gancho `pre-push` `guardia-push` la deja pasar):

```text
To github.com:jmorenobl/kitlegal.git
 * [new branch]      h5-skill-boe-legislacion -> h5-skill-boe-legislacion
branch 'h5-skill-boe-legislacion' set up to track 'origin/h5-skill-boe-legislacion'.
```

`gh pr view h5-skill-boe-legislacion`, antes de crear nada (código 1):

```text
no pull requests found for branch "h5-skill-boe-legislacion"
```

`gh pr create --base main --head h5-skill-boe-legislacion --title "feat(H5): skill boe-legislacion y andamiaje de skills" --body-file specs/006-h5-skill-boe-legislacion/gates/pr-h5.md`:

```text
https://github.com/jmorenobl/kitlegal/pull/27
```

La rama principal remota (`b7eda6e`, workflow 1.9.1, #26) lleva un commit que la rama no tiene; solo cambia
`.specify/workflows/hito/workflow.yml` y `docs/WORKFLOW.md`, que `ci` no compila ni prueba. La propuesta de cambio se
evalúa sobre su commit de fusión, que lo incluye.

## 2. Checks

`gh pr checks h5-skill-boe-legislacion --watch --interval 20` (últimas líneas; cuando el `--watch` terminó ya estaban
los cuatro estados de Codecov):

```text
codecov/patch	fail	0	https://app.codecov.io/gh/jmorenobl/kitlegal/pull/27
ci	pass	4m35s	https://github.com/jmorenobl/kitlegal/actions/runs/34922178932/job/104232536074
codecov/project	pass	0	https://app.codecov.io/gh/jmorenobl/kitlegal/pull/27
codecov/project/internal/cli	pass	1s	https://app.codecov.io/gh/jmorenobl/kitlegal/pull/27
codecov/project/internal/core	pass	1s	https://app.codecov.io/gh/jmorenobl/kitlegal/pull/27
```

`gh run list --branch h5-skill-boe-legislacion`, antes de poner ninguna etiqueta:

```text
completed	success	feat(H5): skill boe-legislacion y andamiaje de skills	ci	h5-skill-boe-legislacion	pull_request	34922178932	4m38s	2026-09-15T02:42:06Z
```

Check-runs del commit de la cabeza por la API
(`gh api repos/jmorenobl/kitlegal/commits/6d68c21be163d2f860eefbe1c7e1871f33ba701c/check-runs --jq '.check_runs[] | "\(.name)\t\(.status)\t\(.conclusion)\t\(.completed_at)\t\(.output.title)"'`),
leídos al terminar la prueba de red, cuando ya estaba también el job `evals` que lanzó su etiqueta:

```text
evals	completed	failure	2026-09-15T02:50:29Z	null
codecov/project/internal/cli	completed	success	2026-09-15T02:46:53Z	98.09% (target 90.00%)
codecov/project/internal/core	completed	success	2026-09-15T02:46:52Z	90.36% (target 85.00%)
codecov/patch	completed	failure	2026-09-15T02:46:51Z	93.54% of diff hit (target 94.70%)
codecov/project	completed	success	2026-09-15T02:46:50Z	94.22% (target 70.00%)
ci	completed	success	2026-09-15T02:46:44Z	null
```

La primera lectura, antes de la prueba de red, dio las mismas cinco filas sin la de `evals`.

Estados del commit (`gh api repos/jmorenobl/kitlegal/commits/6d68c21be163d2f860eefbe1c7e1871f33ba701c/status --jq '.state, (.statuses[] | "\(.context)\t\(.state)\t\(.description)")'`):

```text
pending
```

Ningún estado: Codecov publica check-runs, no estados de commit, y la API combina una lista vacía como `pending`. No
falta ningún estado de plataforma: los cuatro de Codecov están como check-runs, con cifra y objetivo.

## 3. Lectura

| Estado | Resultado | Objetivo | Veredicto |
|---|---|---|---|
| `ci` | `success` (4 m 35 s) | — | verde |
| `codecov/project` | 94,22 % | 70 % | verde |
| `codecov/project/internal/core` | 90,36 % | 85 % | verde |
| `codecov/project/internal/cli` | 98,09 % | 90 % | verde |
| `codecov/patch` | 93,54 % del diff | 94,70 % (`auto`: la cobertura de la base `90c3637`) | **rojo** |
| `evals` (prueba de red, §12.2) | `failure` | — | **rojo**: `gates/prueba-de-red.md` |

**`codecov/patch` no es un verde vacío ni un fallo de configuración**: mide 20 ficheros y 2 090 líneas del diff, y es
bloqueante (`informational: false`). El comentario de `codecov[bot]` (actualizado 2026-09-15T02:46:53Z) da:

```text
:x: Patch coverage is `93.54067%` with `135 lines` in your changes missing coverage. Please review.
:white_check_mark: Project coverage is 94.22%. Comparing base (`90c3637`) to head (`6d68c21`).

| Files with missing lines | Patch % | Lines |
| internal/evals/trazas.go      | 86.99% | 45 Missing |
| internal/skills/sincronia.go  | 87.44% | 30 Missing |
| internal/evals/preparar.go    | 81.18% | 19 Missing |
| internal/skills/esquemas.go   | 91.86% | 14 Missing |
| internal/evals/sesion.go      | 90.08% | 12 Missing |
| internal/skills/normas.go     | 94.73% | 5 Missing  |
| internal/evals/formato.go     | 80.00% | 3 Missing  |
| internal/evals/informe.go     | 99.20% | 2 Missing  |
| internal/skills/skill.go      | 92.30% | 2 Missing  |
| internal/evals/conjunto.go    | 99.34% | 1 Missing  |
| ... and 2 more

@@            Coverage Diff             @@
##             main      #27      +/-   ##
- Coverage   94.70%   94.22%   -0.48%
  Files          62       80      +18
  Lines        2831     4921    +2090
+ Hits         2681     4637    +1956
- Misses        150      284     +134
```

(Tabla abreviada: enlaces retirados.) Las dos filas de «2 more» son `internal/skills/comandos.go` e
`internal/skills/frontmatter.go`, una línea cada una: la unión local de los dos perfiles de `make ci` da las mismas
ramas sin cubrir, listadas por fichero y línea en `gates/tarea-T030.md`, que es donde está el diagnóstico y el arreglo
(T034 y T035). Ningún umbral se toca.

Con `codecov/patch` y la prueba de red en rojo, T030 no se marca.
