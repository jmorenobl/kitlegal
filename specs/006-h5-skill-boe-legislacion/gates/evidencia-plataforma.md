# Evidencia de plataforma de H5 (T030)

Intento 5 de T030 (el tercero que cuenta el workflow, `gates/tareas-intentos.json` `T030: 3`, concedido por la
supervisión del run tras T040 y T041), 2026-09-15. Rama `h5-skill-boe-legislacion`, cabeza
`a40d16a6d2bd4b4159bc6bf2788f77f20f3c5f4a` (`feat(H5): T041`), publicada por avance rápido sobre la del intento 4
(`537e5d6`, `feat(H5): T039`) con los dos commits de T040 y T041. Propuesta de cambio:
[#27](https://github.com/jmorenobl/kitlegal/pull/27), la misma de los intentos 1 a 4. Todas las órdenes, tal cual; las
que llevan `rtk proxy` delante, con el paso directo (el envoltorio de la terminal reescribe salidas). Los intentos 1 a 4
están en las versiones anteriores de este fichero (historial de git del fichero).

## 1. Publicación

`git push -u origin h5-skill-boe-legislacion` (el gancho `pre-push` `guardia-push` la deja pasar, 0,01 s):

```text
To github.com:jmorenobl/kitlegal.git
   537e5d6..a40d16a  h5-skill-boe-legislacion -> h5-skill-boe-legislacion
branch 'h5-skill-boe-legislacion' set up to track 'origin/h5-skill-boe-legislacion'.
```

`gh pr view h5-skill-boe-legislacion --json number,state,url,headRefOid,baseRefName,mergeable,labels,title`, antes del
push (cabeza `537e5d6`) y después: la propuesta #27 existe (`OPEN`, base `main`), así que **no se crea ninguna**. Antes
del push:

```json
{"baseRefName":"main","headRefOid":"537e5d6f7ed5ff48f17f343323f0cef026cb9e33","labels":[],"mergeable":"MERGEABLE","number":27,"state":"OPEN","title":"feat(H5): skill boe-legislacion y andamiaje de skills","url":"https://github.com/jmorenobl/kitlegal/pull/27"}
```

y, en la primera lectura tras el push, ya la cabeza nueva:

```json
{"baseRefName":"main","headRefOid":"a40d16a6d2bd4b4159bc6bf2788f77f20f3c5f4a","labels":[],"mergeable":"MERGEABLE","number":27,"state":"OPEN","title":"feat(H5): skill boe-legislacion y andamiaje de skills","url":"https://github.com/jmorenobl/kitlegal/pull/27"}
```

La rama principal remota sigue en `b7eda6e` (workflow 1.9.1, #26; `git ls-remote origin main`), como en los intentos 1
a 4: no se movió. Al terminar la tarea, el cuerpo de #27 se volvió a sincronizar con
`gh pr edit 27 --body-file specs/006-h5-skill-boe-legislacion/gates/pr-h5.md`, porque T040, T041 y este intento cambian
ese fichero.

## 2. Checks

`gh pr checks h5-skill-boe-legislacion --watch --interval 20` (código 0; el `--watch` terminó con `ci` y sin ningún
estado de Codecov todavía, como advierte la tarea):

```text
ci	pending	0	https://github.com/jmorenobl/kitlegal/actions/runs/34956091100/job/104338379873
ci	pass	4m27s	https://github.com/jmorenobl/kitlegal/actions/runs/34956091100/job/104338379873
```

`gh run list --branch h5-skill-boe-legislacion`, tras el `--watch` y antes de poner ninguna etiqueta (las ocho filas
antiguas son las de los intentos 1 a 4):

```text
completed	success	feat(H5): skill boe-legislacion y andamiaje de skills	ci	h5-skill-boe-legislacion	pull_request	34956091100	4m31s	2026-09-15T10:05:46Z
completed	failure	feat(H5): skill boe-legislacion y andamiaje de skills	evals	h5-skill-boe-legislacion	pull_request	34941499481	10m11s	2026-09-15T07:23:55Z
completed	success	feat(H5): skill boe-legislacion y andamiaje de skills	ci	h5-skill-boe-legislacion	pull_request	34941036417	4m49s	2026-09-15T07:18:27Z
completed	failure	feat(H5): skill boe-legislacion y andamiaje de skills	evals	h5-skill-boe-legislacion	pull_request	34936425178	10m27s	2026-09-15T06:19:30Z
completed	success	feat(H5): skill boe-legislacion y andamiaje de skills	ci	h5-skill-boe-legislacion	pull_request	34935998623	4m34s	2026-09-15T06:13:33Z
completed	failure	feat(H5): skill boe-legislacion y andamiaje de skills	evals	h5-skill-boe-legislacion	pull_request	34930222593	3m52s	2026-09-15T04:48:06Z
completed	success	feat(H5): skill boe-legislacion y andamiaje de skills	ci	h5-skill-boe-legislacion	pull_request	34929854868	4m54s	2026-09-15T04:42:23Z
completed	failure	feat(H5): skill boe-legislacion y andamiaje de skills	evals	h5-skill-boe-legislacion	pull_request	34922606273	1m37s	2026-09-15T02:48:53Z
completed	success	feat(H5): skill boe-legislacion y andamiaje de skills	ci	h5-skill-boe-legislacion	pull_request	34922178932	4m38s	2026-09-15T02:42:06Z
```

`gh run view 34956091100 --json createdAt,conclusion,status,headSha,event,workflowName,jobs`: `createdAt`
`2026-09-15T10:05:46Z`, `event` `pull_request`, `workflowName` `ci`, `headSha` `a40d16a…`, job `ci` de 10:05:49 a
10:10:16, `success`.

Check-runs del commit de la cabeza por la API
(`gh api repos/jmorenobl/kitlegal/commits/a40d16a6d2bd4b4159bc6bf2788f77f20f3c5f4a/check-runs --jq '.check_runs[] | "\(.name)\t\(.status)\t\(.conclusion)\t\(.completed_at)\t\(.output.title)"'`).
La primera lectura, a las 10:10:45Z (29 s después del fin de `ci`), tenía `ci` completado y un solo check-run de Codecov,
`codecov/project`, todavía `in_progress`:

```text
codecov/project	in_progress	null	null	null
ci	completed	success	2026-09-15T10:10:16Z	null
```

La segunda, a las 10:10:55Z (39 s después del fin de `ci`), ya tenía los cinco, con los cuatro de Codecov completados
entre 10:10:46 y 10:10:49 (30 a 33 s después del fin de `ci`; en el intento 4 tardaron 11 a 14 s, en el 3, 50 s, y en
el 2, 15 s):

```text
codecov/project/internal/cli	completed	success	2026-09-15T10:10:49Z	98.09% (target 90.00%)
codecov/project/internal/core	completed	success	2026-09-15T10:10:48Z	90.36% (target 85.00%)
codecov/patch	completed	success	2026-09-15T10:10:47Z	98.42% of diff hit (target 94.70%)
codecov/project	completed	success	2026-09-15T10:10:46Z	96.30% (target 70.00%)
ci	completed	success	2026-09-15T10:10:16Z	null
```

Estados del commit
(`gh api repos/jmorenobl/kitlegal/commits/a40d16a6d2bd4b4159bc6bf2788f77f20f3c5f4a/status --jq '.state, (.statuses[] | "\(.context)\t\(.state)\t\(.description)")'`):

```text
pending
```

Ningún estado, como en los intentos 1 a 4: Codecov publica check-runs, no estados de commit, y la API combina una lista
vacía como `pending`. No falta ningún estado de plataforma: los cuatro de Codecov están como check-runs, con cifra y
objetivo.

El registro del paso «Publicar el perfil de cobertura» de la ejecución de `ci` muestra la subida de los dos perfiles
(`./codecov upload-coverage … --sha a40d16a6d2bd4b4159bc6bf2788f77f20f3c5f4a --file ./coverage.out --file
./coverage-integration.out`) a las 10:10:12. El comentario de `codecov[bot]` (actualizado 2026-09-15T10:10:50Z,
«Comparing base (`90c3637`) to head (`a40d16a`)», con el aviso «Report is 1 commits behind head on main», porque la base
comparada es el último commit de `main` con informe de cobertura) repite la tabla de los intentos 2 a 4: 33 líneas sin
cubrir en los nueve ficheros justificados en `gates/tarea-T034.md` y `gates/tarea-T035.md` (`internal/evals/trazas.go`
con 5, 98,56 % del parche) y `Coverage 94.70% → 96.30% (+1.60%)`, con 4 924 líneas y 4 742 cubiertas, las mismas cifras
que sobre `537e5d6`: T040 y T041 no tocan ningún fichero Go. Ningún umbral cambia: `codecov.yml` no está en el diff
frente a `main`.

## 3. Lectura

| Estado | Resultado | Objetivo | Veredicto |
|---|---|---|---|
| `ci` | `success` (4 m 27 s) | — | verde |
| `codecov/project` | 96,30 % | 70 % | verde |
| `codecov/project/internal/core` | 90,36 % | 85 % | verde |
| `codecov/project/internal/cli` | 98,09 % | 90 % | verde |
| `codecov/patch` | 98,42 % del diff | 94,70 % (`auto`: la cobertura de la base `90c3637`) | verde |
| `evals` (prueba de red, §12.2) | `failure` | — | **rojo**: `gates/prueba-de-red.md` |

Tras la prueba de red (§12.2), la lista de check-runs de la misma cabeza añade
`evals	completed	failure	2026-09-15T10:24:53Z` (ejecución 34956596912, `gates/prueba-de-red.md`).

Con `ci` y los cuatro estados de Codecov en verde, la tarea siguió con quickstart §12.1 y §12.2. **La prueba de red
cumple todo lo que SC-012 espera de ella** (las trece trazas se leen enteras, la conexión `local` de `a9998` se atribuye
a su invocación, ninguna petición llegó a la red de una fuente y la sesión de prueba de red pasa como la 01), y las
cuatro positivas que fallaron en el intento 4 pasan con T040 y T041 (02, 05, 07 y 10), pero **el veredicto es `fallo`**
porque otras cuatro no pasan: 03, 06 y 08 por `comando ausente` y `cita ausente` (el modelo pidió otro artículo, la
misma causa por la que la persona decidió que la 05 y la 07 nombren el artículo) y 09 por `cita ausente` (el nombre de
la norma dentro de los corchetes, la forma que T040 declaró inválida). Es lo que el modelo de las sesiones hizo frente
a lo grabado, no un defecto del job, y la línea de T030 detiene la tarea sin marcarla. Diagnóstico y arreglo en
`gates/tarea-T030.md` (tareas nuevas T042 y T043).
