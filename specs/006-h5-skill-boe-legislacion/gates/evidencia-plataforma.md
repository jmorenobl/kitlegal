# Evidencia de plataforma de H5 (T030)

Intento 4 de T030 (el tercero que cuenta el workflow, `gates/tareas-intentos.json` `T030: 3`, concedido por la
supervisión del run tras T038 y T039), 2026-09-15. Rama `h5-skill-boe-legislacion`, cabeza
`537e5d6f7ed5ff48f17f343323f0cef026cb9e33` (`feat(H5): T039`), publicada por avance rápido sobre la del intento 3
(`857ec46`, `feat(H5): T037`) con los dos commits de T038 y T039. Propuesta de cambio:
[#27](https://github.com/jmorenobl/kitlegal/pull/27), la misma de los intentos 1 a 3. Todas las órdenes, tal cual; las
que llevan `rtk proxy` delante, con el paso directo (el envoltorio de la terminal reescribe salidas). Los intentos 1 a 3
están en las versiones anteriores de este fichero (historial de git del fichero).

## 1. Publicación

`git push -u origin h5-skill-boe-legislacion` (el gancho `pre-push` `guardia-push` la deja pasar, 0,01 s):

```text
To github.com:jmorenobl/kitlegal.git
   857ec46..537e5d6  h5-skill-boe-legislacion -> h5-skill-boe-legislacion
branch 'h5-skill-boe-legislacion' set up to track 'origin/h5-skill-boe-legislacion'.
```

`gh pr view h5-skill-boe-legislacion --json number,state,url,headRefOid,baseRefName,mergeable,labels,title`, antes del
push (cabeza `857ec46`) y después: la propuesta #27 existe (`OPEN`, base `main`), así que **no se crea ninguna**.
Inmediatamente después del push la plataforma todavía devolvía la cabeza anterior:

```json
{"baseRefName":"main","headRefOid":"857ec465074a273bd0d7c0c175185f830d3112ea","labels":[],"mergeable":"MERGEABLE","number":27,"state":"OPEN","title":"feat(H5): skill boe-legislacion y andamiaje de skills","url":"https://github.com/jmorenobl/kitlegal/pull/27"}
```

y unos segundos después, la nueva: `{"headRefOid":"537e5d6f7ed5ff48f17f343323f0cef026cb9e33","mergeable":"MERGEABLE"}`
(y, con la orden de §12.1, `{"headRefOid":"537e5d6…","labels":[],"number":27}`). La rama principal remota sigue en
`b7eda6e` (workflow 1.9.1, #26), como en los intentos 1 a 3: no se movió. Al terminar la tarea, el cuerpo de #27 se
volvió a sincronizar con `gh pr edit 27 --body-file specs/006-h5-skill-boe-legislacion/gates/pr-h5.md`, porque T038,
T039 y este intento cambian ese fichero.

## 2. Checks

`gh pr checks h5-skill-boe-legislacion --watch --interval 20` (código 0; el `--watch` terminó con `ci` y sin ningún
estado de Codecov todavía, como advierte la tarea):

```text
ci	pending	0	https://github.com/jmorenobl/kitlegal/actions/runs/34941036417/job/104289549188
ci	pass	4m44s	https://github.com/jmorenobl/kitlegal/actions/runs/34941036417/job/104289549188
```

`gh run list --branch h5-skill-boe-legislacion`, tras el `--watch` y antes de poner ninguna etiqueta (las seis filas
antiguas son las de los intentos 1 a 3):

```text
completed	success	feat(H5): skill boe-legislacion y andamiaje de skills	ci	h5-skill-boe-legislacion	pull_request	34941036417	4m49s	2026-09-15T07:18:27Z
completed	failure	feat(H5): skill boe-legislacion y andamiaje de skills	evals	h5-skill-boe-legislacion	pull_request	34936425178	10m27s	2026-09-15T06:19:30Z
completed	success	feat(H5): skill boe-legislacion y andamiaje de skills	ci	h5-skill-boe-legislacion	pull_request	34935998623	4m34s	2026-09-15T06:13:33Z
completed	failure	feat(H5): skill boe-legislacion y andamiaje de skills	evals	h5-skill-boe-legislacion	pull_request	34930222593	3m52s	2026-09-15T04:48:06Z
completed	success	feat(H5): skill boe-legislacion y andamiaje de skills	ci	h5-skill-boe-legislacion	pull_request	34929854868	4m54s	2026-09-15T04:42:23Z
completed	failure	feat(H5): skill boe-legislacion y andamiaje de skills	evals	h5-skill-boe-legislacion	pull_request	34922606273	1m37s	2026-09-15T02:48:53Z
completed	success	feat(H5): skill boe-legislacion y andamiaje de skills	ci	h5-skill-boe-legislacion	pull_request	34922178932	4m38s	2026-09-15T02:42:06Z
```

`gh run view 34941036417 --json createdAt,conclusion,status,headSha,event,workflowName,jobs`: `createdAt`
`2026-09-15T07:18:27Z`, `event` `pull_request`, `workflowName` `ci`, `headSha` `537e5d6…`, job `ci` de 07:18:32 a
07:23:16, `success`.

Check-runs del commit de la cabeza por la API
(`gh api repos/jmorenobl/kitlegal/commits/537e5d6f7ed5ff48f17f343323f0cef026cb9e33/check-runs --jq '.check_runs[] | "\(.name)\t\(.status)\t\(.conclusion)\t\(.completed_at)\t\(.output.title)"'`).
La primera lectura, a las 07:23:34Z (18 s después del fin de `ci`), ya tenía los cinco, con los cuatro de Codecov
completados entre 07:23:27 y 07:23:30 (11 a 14 s después del fin de `ci`; en el intento 3 tardaron 50 s y en el 2,
15 s):

```text
codecov/project/internal/cli	completed	success	2026-09-15T07:23:30Z	98.09% (target 90.00%)
codecov/project/internal/core	completed	success	2026-09-15T07:23:29Z	90.36% (target 85.00%)
codecov/patch	completed	success	2026-09-15T07:23:28Z	98.42% of diff hit (target 94.70%)
codecov/project	completed	success	2026-09-15T07:23:27Z	96.30% (target 70.00%)
ci	completed	success	2026-09-15T07:23:16Z	null
```

Estados del commit
(`gh api repos/jmorenobl/kitlegal/commits/537e5d6f7ed5ff48f17f343323f0cef026cb9e33/status --jq '.state, (.statuses[] | "\(.context)\t\(.state)\t\(.description)")'`):

```text
pending
```

Ningún estado, como en los intentos 1 a 3: Codecov publica check-runs, no estados de commit, y la API combina una lista
vacía como `pending`. No falta ningún estado de plataforma: los cuatro de Codecov están como check-runs, con cifra y
objetivo.

El registro del paso «Publicar el perfil de cobertura» de la ejecución de `ci` muestra la subida de los dos perfiles con
`--sha 537e5d6f7ed5ff48f17f343323f0cef026cb9e33` a las 07:23:10. El comentario de `codecov[bot]` (actualizado
2026-09-15T07:23:31Z, «Comparing base (`90c3637`) to head (`537e5d6`)», con el aviso «Report is 1 commits behind head on
main», porque la base comparada es el último commit de `main` con informe de cobertura) repite la tabla de los intentos 2
y 3: 33 líneas sin cubrir en los nueve ficheros justificados en `gates/tarea-T034.md` y `gates/tarea-T035.md`
(`internal/evals/trazas.go` con 5, 98,56 % del parche) y `Coverage 94.70% → 96.30% (+1.60%)`, con 4 924 líneas y 4 742
cubiertas, las mismas cifras que sobre `857ec46`: el cambio de T039 en `formaDeLlamada` no añade ninguna línea sin
cubrir. Ningún umbral cambia: `codecov.yml` no está en el diff frente a `main`.

Tras la prueba de red (§12.2), la lista de check-runs de la misma cabeza añade
`evals	completed	failure	2026-09-15T07:34:05Z` (ejecución 34941499481, `gates/prueba-de-red.md`).

## 3. Lectura

| Estado | Resultado | Objetivo | Veredicto |
|---|---|---|---|
| `ci` | `success` (4 m 44 s) | — | verde |
| `codecov/project` | 96,30 % | 70 % | verde |
| `codecov/project/internal/core` | 90,36 % | 85 % | verde |
| `codecov/project/internal/cli` | 98,09 % | 90 % | verde |
| `codecov/patch` | 98,42 % del diff | 94,70 % (`auto`: la cobertura de la base `90c3637`) | verde |
| `evals` (prueba de red, §12.2) | `failure` | — | **rojo**: `gates/prueba-de-red.md` |

Con `ci` y los cuatro estados de Codecov en verde, la tarea siguió con quickstart §12.1 y §12.2. **La prueba de red
cumple todo lo que SC-012 espera de ella** (las trece trazas se leen enteras, la conexión `local` de `a9998` se atribuye
a su invocación, ninguna petición llegó a la red de una fuente y la sesión de prueba de red pasa como la 01), pero **el
veredicto es `fallo`** porque seis positivas no pasan por `comando ausente` y `cita ausente`: es un defecto que la
prueba de red descubre en el protocolo de la skill frente a lo grabado, no en el job, y la línea de T030 detiene la
tarea sin marcarla. Diagnóstico y arreglo en `gates/tarea-T030.md` (tarea nueva T040 y una decisión para la persona).
