# Evidencia de plataforma de H5 (T030)

Intento 6 de T030 (el tercero que cuenta el workflow, `gates/tareas-intentos.json` `T030: 3`, concedido por la
supervisión del run tras T042 y T043), 2026-09-15. Rama `h5-skill-boe-legislacion`, cabeza
`a99295f50716010b560c1998ac4e71e62cd24da4` (`feat(H5): T043`), publicada por avance rápido sobre la del intento 5
(`a40d16a`, `feat(H5): T041`) con los dos commits de T042 y T043. Propuesta de cambio:
[#27](https://github.com/jmorenobl/kitlegal/pull/27), la misma de los intentos 1 a 5. Todas las órdenes, tal cual; las
que llevan `rtk proxy` delante, con el paso directo (el envoltorio de la terminal reescribe salidas). Los intentos 1 a 5
están en las versiones anteriores de este fichero (historial de git del fichero).

## 1. Publicación

`git ls-remote origin main h5-skill-boe-legislacion` antes del push: la rama remota en `a40d16a` (la cabeza del intento
5) y `main` en `b7eda6e`. `git push -u origin h5-skill-boe-legislacion` (el gancho `pre-push` `guardia-push` la deja
pasar, 0,01 s):

```text
To github.com:jmorenobl/kitlegal.git
   a40d16a..a99295f  h5-skill-boe-legislacion -> h5-skill-boe-legislacion
branch 'h5-skill-boe-legislacion' set up to track 'origin/h5-skill-boe-legislacion'.
```

`gh pr view h5-skill-boe-legislacion --json number,state,url,headRefOid,baseRefName,mergeable,labels,title`, en la
primera lectura tras el push: la propuesta #27 existe (`OPEN`, base `main`), ya con la cabeza nueva, así que **no se
crea ninguna**:

```json
{"baseRefName":"main","headRefOid":"a99295f50716010b560c1998ac4e71e62cd24da4","labels":[],"mergeable":"MERGEABLE","number":27,"state":"OPEN","title":"feat(H5): skill boe-legislacion y andamiaje de skills","url":"https://github.com/jmorenobl/kitlegal/pull/27"}
```

La rama principal remota sigue en `b7eda6e` (workflow 1.9.1, #26; `git ls-remote origin main` antes y después del
push), como en los intentos 1 a 5: no se movió. Al terminar la tarea, el cuerpo de #27 se volvió a sincronizar con
`gh pr edit 27 --body-file specs/006-h5-skill-boe-legislacion/gates/pr-h5.md`, porque T042, T043 y este intento cambian
ese fichero.

## 2. Checks

`gh pr checks h5-skill-boe-legislacion --watch --interval 20` (código 0; el `--watch` terminó con `ci` y sin ningún
estado de Codecov todavía, como advierte la tarea):

```text
ci	pending	0	https://github.com/jmorenobl/kitlegal/actions/runs/34961223411/job/104355014777
ci	pass	4m27s	https://github.com/jmorenobl/kitlegal/actions/runs/34961223411/job/104355014777
```

`gh run list --branch h5-skill-boe-legislacion`, tras el `--watch` y antes de poner ninguna etiqueta (las diez filas
antiguas son las de los intentos 1 a 5):

```text
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

`gh run view 34961223411 --json createdAt,conclusion,status,headSha,event,workflowName,jobs`: `createdAt`
`2026-09-15T11:03:31Z`, `event` `pull_request`, `workflowName` `ci`, `headSha` `a99295f…`, job `ci` de 11:03:34 a
11:08:01, `success`.

Check-runs del commit de la cabeza por la API
(`gh api repos/jmorenobl/kitlegal/commits/a99295f50716010b560c1998ac4e71e62cd24da4/check-runs --jq '.check_runs[] | "\(.name)\t\(.status)\t\(.conclusion)\t\(.completed_at)\t\(.output.title)"'`).
Las lecturas a las 11:08:20Z, 11:08:32Z y 11:08:42Z (19, 31 y 41 s después del fin de `ci`) tenían solo `ci`:

```text
ci	completed	success	2026-09-15T11:08:01Z	null
```

La de las 11:08:53Z (52 s después del fin de `ci`) ya tenía los cinco, con los cuatro de Codecov completados entre
11:08:46 y 11:08:49 (45 a 48 s después del fin de `ci`; en el intento 5 tardaron 30 a 33 s, en el 4, 11 a 14 s, en el
3, 50 s, y en el 2, 15 s):

```text
codecov/project/internal/cli	completed	success	2026-09-15T11:08:49Z	98.09% (target 90.00%)
codecov/project/internal/core	completed	success	2026-09-15T11:08:48Z	90.36% (target 85.00%)
codecov/patch	completed	success	2026-09-15T11:08:47Z	98.42% of diff hit (target 94.70%)
codecov/project	completed	success	2026-09-15T11:08:46Z	96.30% (target 70.00%)
ci	completed	success	2026-09-15T11:08:01Z	null
```

Estados del commit
(`gh api repos/jmorenobl/kitlegal/commits/a99295f50716010b560c1998ac4e71e62cd24da4/status --jq '.state, (.statuses[] | "\(.context)\t\(.state)\t\(.description)")'`):

```text
pending
```

Ningún estado, como en los intentos 1 a 5: Codecov publica check-runs, no estados de commit, y la API combina una lista
vacía como `pending`. No falta ningún estado de plataforma: los cuatro de Codecov están como check-runs, con cifra y
objetivo.

El registro del paso «Publicar el perfil de cobertura» de la ejecución de `ci` muestra la subida de los dos perfiles
(`./codecov upload-coverage … --sha a99295f50716010b560c1998ac4e71e62cd24da4 --file ./coverage.out --file
./coverage-integration.out`) a las 11:07:58. El comentario de `codecov[bot]` (actualizado 2026-09-15T11:08:50Z,
«Comparing base (`90c3637`) to head (`a99295f`)», con el aviso «Report is 1 commits behind head on main», porque la base
comparada es el último commit de `main` con informe de cobertura) repite la tabla de los intentos 2 a 5: 33 líneas sin
cubrir en los nueve ficheros justificados en `gates/tarea-T034.md` y `gates/tarea-T035.md` (`internal/evals/trazas.go`
con 5, 98,56 % del parche) y `Coverage 94.70% → 96.30% (+1.60%)`, con 4 924 líneas y 4 742 cubiertas, las mismas cifras
que sobre `a40d16a`: T042 y T043 no tocan ningún fichero Go. Ningún umbral cambia: `codecov.yml` no está en el diff
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
`evals	completed	failure	2026-09-15T11:17:56Z` (ejecución 34961757559, `gates/prueba-de-red.md`).

Con `ci` y los cuatro estados de Codecov en verde, la tarea siguió con quickstart §12.1 y §12.2. **T042 y T043 hicieron
lo que debían**: las diez positivas leyeron el índice y el bloque nombrado en la pregunta a la primera (las seis de
T043 incluidas; la 02 copió `a1-30` del índice sin pedir ningún vecino) y la 09 citó, sola en su línea tras la
transcripción, en la forma fija exacta que T042 escribió. **Pero el veredicto es `fallo`** por dos motivos nuevos, uno
de ellos del job: la sesión 09 es `sesión ilegible` porque su traza tiene una línea que `LeerTrazas` no admite,
`clone(… <unfinished ...>) = ?`, la forma con la que `strace` escribe una llamada en curso cuando el proceso termina
antes de que tenga resultado (una sesión terminada con código 0, no cortada: el supuesto S4 falla aquí por primera
vez), y la 08 leyó `a20` con 0 pero citó `[art. 20.1 de la LTAIBG, BOE-A-2013-12887, bloque a20]`, con texto delante
del identificador dentro de los corchetes, la tercera sesión distinta en tres intentos que pone la forma legible dentro.
La línea de T030 detiene la tarea sin marcarla. Diagnóstico y arreglo en `gates/tarea-T030.md` (tareas nuevas T044,
T045 y T046).
