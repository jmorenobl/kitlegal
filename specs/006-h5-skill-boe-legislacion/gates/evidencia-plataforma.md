# Evidencia de plataforma de H5 (T030)

Intento 7 de T030 (el tercero que cuenta el workflow, `gates/tareas-intentos.json` `T030: 3`, concedido por la
supervisión del run tras T044, T045 y T046), 2026-09-15. Rama `h5-skill-boe-legislacion`, cabeza
`091facd82d3744dffcfb3b17115c5ea25452d7d5` (`feat(H5): T046`), publicada por avance rápido sobre la del intento 6
(`a99295f`, `feat(H5): T043`) con los tres commits de T044, T045 y T046. Propuesta de cambio:
[#27](https://github.com/jmorenobl/kitlegal/pull/27), la misma de los intentos 1 a 6. Todas las órdenes, tal cual; las
que llevan `rtk proxy` delante, con el paso directo (el envoltorio de la terminal reescribe salidas). Los intentos 1 a 6
están en las versiones anteriores de este fichero (historial de git del fichero).

## 1. Publicación

`git ls-remote origin main h5-skill-boe-legislacion` antes del push: la rama remota en `a99295f` (la cabeza del intento
6) y `main` en `b7eda6e`. `git push -u origin h5-skill-boe-legislacion` (el gancho `pre-push` `guardia-push` la deja
pasar, 0,01 s):

```text
To github.com:jmorenobl/kitlegal.git
   a99295f..091facd  h5-skill-boe-legislacion -> h5-skill-boe-legislacion
branch 'h5-skill-boe-legislacion' set up to track 'origin/h5-skill-boe-legislacion'.
```

`gh pr view h5-skill-boe-legislacion --json number,state,url,headRefOid,baseRefName,mergeable,labels,title`, en la
primera lectura tras el push: la propuesta #27 existe (`OPEN`, base `main`), ya con la cabeza nueva, así que **no se
crea ninguna**:

```json
{"baseRefName":"main","headRefOid":"091facd82d3744dffcfb3b17115c5ea25452d7d5","labels":[],"mergeable":"MERGEABLE","number":27,"state":"OPEN","title":"feat(H5): skill boe-legislacion y andamiaje de skills","url":"https://github.com/jmorenobl/kitlegal/pull/27"}
```

La rama principal remota sigue en `b7eda6e` (workflow 1.9.1, #26; `git ls-remote origin main` antes y después del
push), como en los intentos 1 a 6: no se movió. Al terminar la tarea, el cuerpo de #27 se volvió a sincronizar con
`gh pr edit 27 --body-file specs/006-h5-skill-boe-legislacion/gates/pr-h5.md`, porque T045, T046 y este intento cambian
ese fichero.

## 2. Checks

`gh pr checks h5-skill-boe-legislacion --watch --interval 20` (código 0; el `--watch` terminó con `ci` y sin ningún
estado de Codecov todavía, como advierte la tarea):

```text
ci	pending	0	https://github.com/jmorenobl/kitlegal/actions/runs/34999203695/job/104483080119
ci	pass	4m50s	https://github.com/jmorenobl/kitlegal/actions/runs/34999203695/job/104483080119
```

`gh run list --branch h5-skill-boe-legislacion`, tras el `--watch` y antes de poner ninguna etiqueta (las doce filas
antiguas son las de los intentos 1 a 6):

```text
completed	success	feat(H5): skill boe-legislacion y andamiaje de skills	ci	h5-skill-boe-legislacion	pull_request	34999203695	4m56s	2026-09-15T17:06:46Z
completed	failure	feat(H5): skill boe-legislacion y andamiaje de skills	evals	h5-skill-boe-legislacion	pull_request	34961757559	8m29s	2026-09-15T11:09:28Z
completed	success	feat(H5): skill boe-legislacion y andamiaje de skills	ci	h5-skill-boe-legislacion	pull_request	34961223411	4m31s	2026-09-15T11:03:31Z
completed	failure	feat(H5): skill boe-legislacion y andamiaje de skills	evals	h5-skill-boe-legislacion	pull_request	34956596912	13m39s	2026-09-15T10:11:17Z
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

`gh run view 34999203695 --json createdAt,conclusion,status,headSha,event,workflowName,jobs`: `createdAt`
`2026-09-15T17:06:46Z`, `event` `pull_request`, `workflowName` `ci`, `headSha` `091facd…`, job `ci` de 17:06:51 a
17:11:41, `success` (el paso «Ejecutar los controles», de 17:07:27 a 17:11:34; «Publicar el perfil de cobertura», de
17:11:34 a 17:11:38).

Check-runs del commit de la cabeza por la API
(`gh api repos/jmorenobl/kitlegal/commits/091facd82d3744dffcfb3b17115c5ea25452d7d5/check-runs --jq '.check_runs[] | "\(.name)\t\(.status)\t\(.conclusion)\t\(.completed_at)\t\(.output.title)"'`).
Las lecturas a las 17:11:51Z y 17:12:14Z (10 y 33 s después del fin de `ci`) tenían solo `ci`:

```text
ci	completed	success	2026-09-15T17:11:41Z	null
```

La de las 17:12:24Z (43 s después del fin de `ci`) ya tenía los cinco, con los cuatro de Codecov completados entre
17:12:16 y 17:12:19 (35 a 38 s después del fin de `ci`; en el intento 6 tardaron 45 a 48 s, en el 5, 30 a 33 s, en el
4, 11 a 14 s, en el 3, 50 s, y en el 2, 15 s):

```text
codecov/project/internal/cli	completed	success	2026-09-15T17:12:19Z	98.09% (target 90.00%)
codecov/project/internal/core	completed	success	2026-09-15T17:12:18Z	90.36% (target 85.00%)
codecov/patch	completed	success	2026-09-15T17:12:17Z	98.48% of diff hit (target 94.70%)
codecov/project	completed	success	2026-09-15T17:12:16Z	96.33% (target 70.00%)
ci	completed	success	2026-09-15T17:11:41Z	null
```

Estados del commit
(`gh api repos/jmorenobl/kitlegal/commits/091facd82d3744dffcfb3b17115c5ea25452d7d5/status --jq '.state, (.statuses[] | "\(.context)\t\(.state)\t\(.description)")'`),
leídos a las 17:11:51Z y a las 17:12:32Z, ya con los check-runs de Codecov:

```text
pending
```

Ningún estado, como en los intentos 1 a 6: Codecov publica check-runs, no estados de commit, y la API combina una lista
vacía como `pending`. No falta ningún estado de plataforma: los cuatro de Codecov están como check-runs, con cifra y
objetivo.

El registro del paso «Publicar el perfil de cobertura» de la ejecución de `ci` muestra la subida de los dos perfiles
(`./codecov upload-coverage … --sha 091facd82d3744dffcfb3b17115c5ea25452d7d5 --file ./coverage.out --file
./coverage-integration.out`) a las 17:11:36. El comentario de `codecov[bot]` (actualizado 2026-09-15T17:12:20Z,
«Comparing base (`90c3637`) to head (`091facd`)», con el aviso «Report is 1 commits behind head on main», porque la base
comparada es el último commit de `main` con informe de cobertura) da `Coverage 94.70% → 96.33% (+1.63%)`, con 4 965
líneas y 4 783 cubiertas (frente a 4 924 y 4 742 sobre `a99295f`: las 41 líneas nuevas son las de T045 y T046 en
`internal/evals/trazas.go` y `internal/evals/citas.go`), y `98.48197%` del parche con 32 líneas sin cubrir en ocho
ficheros: `internal/skills/sincronia.go` (12), `internal/evals/preparar.go` (6), `internal/evals/trazas.go` (5, ahora
98,71 % frente a 98,56 % en los intentos 2 a 6, porque T045 añade líneas cubiertas), `internal/skills/esquemas.go` (4),
`internal/evals/informe.go` (2), `internal/evals/conjunto.go` (1), `internal/evals/formato.go` (1) y
`internal/skills/normas.go` (1); las mismas líneas justificadas en `gates/tarea-T034.md` y `gates/tarea-T035.md`, salvo
que `internal/skills/skill.go`, con una en los intentos 2 a 6, ya no aparece en la lista. Ningún umbral cambia:
`codecov.yml` no está en el diff frente a `main`.

## 3. Lectura

| Estado | Resultado | Objetivo | Veredicto |
|---|---|---|---|
| `ci` | `success` (4 m 50 s) | — | verde |
| `codecov/project` | 96,33 % | 70 % | verde |
| `codecov/project/internal/core` | 90,36 % | 85 % | verde |
| `codecov/project/internal/cli` | 98,09 % | 90 % | verde |
| `codecov/patch` | 98,48 % del diff | 94,70 % (`auto`: la cobertura de la base `90c3637`) | verde |
| `evals` (prueba de red, §12.2) | `success` | — | **verde**: `gates/prueba-de-red.md` |

Tras la prueba de red (§12.2), la lista de check-runs de la misma cabeza añade
`evals	completed	success	2026-09-15T17:20:43Z	null` (ejecución 34999845098, `gates/prueba-de-red.md`): los seis
check-runs de la cabeza en verde.

Con `ci` y los cuatro estados de Codecov en verde, la tarea siguió con quickstart §12.1 y §12.2, y **la prueba de red
termina por primera vez con veredicto `aprobado`**: las trece trazas legibles (T044 y T045), las diez positivas y las
dos de no activación en verde, la sesión de prueba de red con el mismo resultado que la 01, las dos filas de `a9998`
con 5 y 4, la conexión `127.0.0.1:9` de clase `local` atribuida a la invocación sin `--offline`, ninguna petición
llegada a la red de una fuente y las citas de diez de las once sesiones con la skill activada extraídas con la forma
legible dentro de los corchetes (T046). Detalle en `gates/prueba-de-red.md`; la tarea se marca.
