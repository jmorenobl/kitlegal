# Evidencia de plataforma de H5 (T030)

Intento 3 de T030, 2026-09-15. Rama `h5-skill-boe-legislacion`, cabeza `857ec465074a273bd0d7c0c175185f830d3112ea`
(`feat(H5): T037`), publicada por avance rápido sobre la del intento 2 (`417635e`, `feat(H5): T035`) con los dos
commits de T036 y T037. Propuesta de cambio: [#27](https://github.com/jmorenobl/kitlegal/pull/27), la misma de los
intentos 1 y 2. Todas las órdenes, tal cual; las que llevan `rtk proxy` delante, con el paso directo (el envoltorio de
la terminal reescribe salidas). Los intentos 1 y 2 están en las versiones anteriores de este fichero (commits `c4c7613`
y `857ec46`).

## 1. Publicación

`git push -u origin h5-skill-boe-legislacion` (el gancho `pre-push` `guardia-push` la deja pasar):

```text
To github.com:jmorenobl/kitlegal.git
   417635e..857ec46  h5-skill-boe-legislacion -> h5-skill-boe-legislacion
branch 'h5-skill-boe-legislacion' set up to track 'origin/h5-skill-boe-legislacion'.
```

`gh pr view h5-skill-boe-legislacion --json number,state,url,headRefOid,baseRefName,mergeable,labels,title`, antes del
push (cabeza `417635e`) y después: la propuesta #27 existe (`OPEN`, base `main`), así que **no se crea ninguna**. Tras
el push:

```json
{"baseRefName":"main","headRefOid":"857ec465074a273bd0d7c0c175185f830d3112ea","labels":[],"mergeable":"UNKNOWN","number":27,"state":"OPEN","title":"feat(H5): skill boe-legislacion y andamiaje de skills","url":"https://github.com/jmorenobl/kitlegal/pull/27"}
```

(`mergeable` `UNKNOWN` es lo que la plataforma devuelve mientras recalcula la fusión en los segundos que siguen a un
push; en los intentos 1 y 2 la lectura fue `MERGEABLE`.) La rama principal remota sigue en `b7eda6e` (workflow 1.9.1,
#26), como en los intentos 1 y 2: no se movió. Al terminar la tarea, el cuerpo de #27 se volvió a sincronizar con
`gh pr edit 27 --body-file specs/006-h5-skill-boe-legislacion/gates/pr-h5.md`, porque T036, T037 y este intento cambian
ese fichero.

## 2. Checks

`gh pr checks h5-skill-boe-legislacion --watch --interval 20` (últimas líneas; el `--watch` terminó con `ci` y sin
ningún estado de Codecov todavía, como advierte la tarea):

```text
ci	pending	0	https://github.com/jmorenobl/kitlegal/actions/runs/34935998623/job/104273902500
ci	pass	4m30s	https://github.com/jmorenobl/kitlegal/actions/runs/34935998623/job/104273902500
```

`gh run list --branch h5-skill-boe-legislacion`, tras el `--watch` y antes de poner ninguna etiqueta (las cuatro filas
antiguas son las de los intentos 1 y 2):

```text
completed	success	feat(H5): skill boe-legislacion y andamiaje de skills	ci	h5-skill-boe-legislacion	pull_request	34935998623	4m34s	2026-09-15T06:13:33Z
completed	failure	feat(H5): skill boe-legislacion y andamiaje de skills	evals	h5-skill-boe-legislacion	pull_request	34930222593	3m52s	2026-09-15T04:48:06Z
completed	success	feat(H5): skill boe-legislacion y andamiaje de skills	ci	h5-skill-boe-legislacion	pull_request	34929854868	4m54s	2026-09-15T04:42:23Z
completed	failure	feat(H5): skill boe-legislacion y andamiaje de skills	evals	h5-skill-boe-legislacion	pull_request	34922606273	1m37s	2026-09-15T02:48:53Z
completed	success	feat(H5): skill boe-legislacion y andamiaje de skills	ci	h5-skill-boe-legislacion	pull_request	34922178932	4m38s	2026-09-15T02:42:06Z
```

`gh run view 34935998623 --json createdAt,conclusion,status,headSha,event,workflowName,jobs`: `createdAt`
`2026-09-15T06:13:33Z`, `event` `pull_request`, `workflowName` `ci`, `headSha` `857ec46…`, job `ci` de 06:13:36 a
06:18:06, `success`.

Check-runs del commit de la cabeza por la API
(`gh api repos/jmorenobl/kitlegal/commits/857ec465074a273bd0d7c0c175185f830d3112ea/check-runs --jq '.check_runs[] | "\(.name)\t\(.status)\t\(.conclusion)\t\(.completed_at)\t\(.output.title)"'`).
En las dos primeras lecturas, nada más terminar el `--watch` y algo después, solo `ci`; en la tercera, tres de Codecov
(`internal/cli` todavía sin publicar); en la cuarta, los cuatro, con `completed_at` entre 06:18:55 y 06:18:58, unos
50 s después del fin de `ci` (06:18:06; en el intento 2 tardaron 15 s). El registro del paso «Publicar el perfil de
cobertura» de la ejecución de `ci` dice `Found 2 coverage files to report` y `Your upload is now queued for processing`
a las 06:18:04:

```text
codecov/project/internal/cli	completed	success	2026-09-15T06:18:58Z	98.09% (target 90.00%)
codecov/project/internal/core	completed	success	2026-09-15T06:18:57Z	90.36% (target 85.00%)
codecov/patch	completed	success	2026-09-15T06:18:56Z	98.42% of diff hit (target 94.70%)
codecov/project	completed	success	2026-09-15T06:18:55Z	96.30% (target 70.00%)
ci	completed	success	2026-09-15T06:18:06Z	null
```

Estados del commit (`gh api repos/jmorenobl/kitlegal/commits/857ec465074a273bd0d7c0c175185f830d3112ea/status --jq '.state, (.statuses[] | "\(.context)\t\(.state)\t\(.description)")'`), en las cuatro lecturas:

```text
pending
```

Ningún estado, como en los intentos 1 y 2: Codecov publica check-runs, no estados de commit, y la API combina una lista
vacía como `pending`. No falta ningún estado de plataforma: los cuatro de Codecov están como check-runs, con cifra y
objetivo.

Tras la prueba de red (§12.2), la lista de check-runs de la misma cabeza añade
`evals	completed	failure	2026-09-15T06:29:56Z` (ejecución 34936425178, `gates/prueba-de-red.md`).

## 3. Lectura

| Estado | Resultado | Objetivo | Veredicto |
|---|---|---|---|
| `ci` | `success` (4 m 30 s) | — | verde |
| `codecov/project` | 96,30 % | 70 % | verde |
| `codecov/project/internal/core` | 90,36 % | 85 % | verde |
| `codecov/project/internal/cli` | 98,09 % | 90 % | verde |
| `codecov/patch` | 98,42 % del diff | 94,70 % (`auto`: la cobertura de la base `90c3637`) | verde |
| `evals` (prueba de red, §12.2) | `failure` | — | **rojo**: `gates/prueba-de-red.md` |

Las cinco cifras son las del intento 2 sobre `417635e`: T036 y T037 cambian solo la definición del job y el guion de la
sesión, ningún fichero Go, y el comentario de `codecov[bot]` (actualizado 2026-09-15T06:18:58Z, «Comparing base
(`90c3637`) to head (`857ec46`)») repite la tabla del intento 2: 33 líneas sin cubrir en los nueve ficheros justificados
en `gates/tarea-T034.md` y `gates/tarea-T035.md`, y `Coverage 94.70% → 96.30% (+1.60%)`, con 4 924 líneas y 4 742
cubiertas. Ningún umbral cambia: `codecov.yml` no está en el diff frente a `main`.

Con `ci` y los cuatro estados de Codecov en verde, la tarea siguió con quickstart §12.1 y §12.2. **La prueba de red
descubre un defecto** del lector de trazas: las trece sesiones arrancan, terminan con código 0 y responden (los dos
motivos del intento 2 están resueltos), pero todas salen con `sesión ilegible` por la línea `vfork()` con relleno de
alineación que `strace` escribe en el runner de x86_64. T030 no se marca: diagnóstico y arreglo en
`gates/tarea-T030.md` (tareas nuevas T038, de datos, y T039).
