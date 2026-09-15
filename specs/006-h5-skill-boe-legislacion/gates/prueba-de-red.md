# Prueba de red de H5 (quickstart §12.2, SC-012)

Intento 2 de T030, 2026-09-15, propuesta de cambio [#27](https://github.com/jmorenobl/kitlegal/pull/27), cabeza
`417635e39abf5807f8600b79debd6026fc320ce8` (`feat(H5): T035`). **Resultado: el job llega al informe, y el veredicto es
`fallo` porque las trece sesiones terminan con código 1 al arrancar Claude Code**, sin ningún mensaje en el transcript:
`CLAUDE_CODE_SUBPROCESS_ENV_SCRUB=1` exige `bubblewrap` en Linux y la imagen `ubuntu-24.04` no lo trae. El paso «Retirar
Python del runner» (T033) terminó con 0. La prueba de red descubre un defecto del job y T030 se detiene sin marcarse
(`gates/tarea-T030.md`; arreglo en T036). El intento 1 (ejecución 34922606273, detenida en la purga de paquetes) está
en la versión anterior de este fichero (commit `c4c7613` y anteriores).

Enlace a la ejecución: <https://github.com/jmorenobl/kitlegal/actions/runs/34930222593> (`databaseId` 34930222593).

## 1. Prerrequisitos (§12.1)

`rtk proxy gh secret list`:

```text
CLAUDE_CODE_OAUTH_TOKEN	2026-09-14T09:17:03Z
CODECOV_TOKEN	2026-09-10T21:18:02Z
```

`rtk proxy gh label list --search evals`:

```text
evals	Ejecuta el job de evals de skills sobre la propuesta de cambio	#1D76DB
evals-prueba-de-red	Ejecuta el job de evals con la prueba de red (SC-012 de H5)	#B60205
```

`gh pr view --json number,headRefOid,labels`:

```json
{"headRefOid":"417635e39abf5807f8600b79debd6026fc320ce8","labels":[],"number":27}
```

Todo presente: el secreto y las dos etiquetas.

## 2. Órdenes de §12.2

**Primera** (la etiqueta no estaba puesta, porque el intento 1 la quitó al terminar, así que no se quitó nada):

```text
la etiqueta evals-prueba-de-red no está puesta
```

**Segunda**, `gh pr edit --add-label evals-prueba-de-red`: la orden va sin `rtk proxy` en el quickstart, y el
envoltorio de la terminal reescribió su salida como `ok edited #evals-prueba-de-red` (la orden imprime
`https://github.com/jmorenobl/kitlegal/pull/27`, como en el intento 1); el evento `labeled` que creó está en la
tercera orden.

**Tercera**, a la primera ya con la ejecución:

```text
etiqueta puesta: 2026-09-15T04:48:02Z
{"evals":{"conclusion":"","createdAt":"2026-09-15T04:48:06Z","databaseId":34930222593,"headSha":"417635e39abf5807f8600b79debd6026fc320ce8","status":"in_progress","url":"https://github.com/jmorenobl/kitlegal/actions/runs/34930222593","workflowName":"evals"},"posteriores_a_la_etiqueta":[{"createdAt":"2026-09-15T04:48:06Z","databaseId":34930222593,"workflowName":"evals"}]}
```

**Cuarta**, `gh run watch 34930222593 --exit-status` (en primer plano hasta el final; terminó con `código 1`; su
último bloque son las anotaciones de abajo). `gh run view 34930222593`, al terminar, con los pasos:

```text
X h5-skill-boe-legislacion evals jmorenobl/kitlegal#27 · 34930222593
Triggered via pull_request about 4 minutes ago

JOBS
X evals in 3m49s (ID 104256702587)
  ✓ Set up job
  ✓ Obtener el código del commit evaluado
  ✓ Instalar Go y restaurar la caché
  ✓ Instalar strace y Claude Code
  ✓ Instalar kitlegal y las skills como las deja make install
  ✓ Retirar Python del runner
  X Ejecutar las evals
  - Post Instalar Go y restaurar la caché
  ✓ Post Obtener el código del commit evaluado
  ✓ Complete job

ANNOTATIONS
X 
      at nBn (/$bunfs/root/chunk-y58z5vzc.js:11:25510)
      at A (/$bunfs/root/chunk-xgzajagg.js:11:3535)
      at async <anonymous> (/$bunfs/root/chunk-mpjnf6jh.js:88:4824)
evals: ../../../../../$bunfs/root/chunk-y58z5vzc.js#11
```

(la misma anotación, siete veces). Esas anotaciones no las escribe el job: son el *problem matcher* de `actions/setup-go`,
que toma por un error de Go las líneas `at … (/$bunfs/root/chunk-….js:11:25510)` de las trazas de pila de Claude Code que
el informe copia al registro; no son un estado ni un aviso de la plataforma.

**Quinta** (informe entre marcas): **código 0**; imprime `informe.md` (647 líneas del registro) e `informe.json` (450)
enteros, cada uno de su marca de inicio a su marca de fin. Salida completa, tal cual la da `gh run view --log`, en el
**anexo A**.

**Sexta** (salida de la retirada de Python entre marcas): **código 0**; 925 líneas del registro, de
`--- inicio de la retirada de Python ---` a `--- fin de la retirada de Python ---`. Salida completa en el **anexo B**.

La orden de `--log-failed` que va tras el bloque no procede (las dos anteriores no fallan); se ejecutó de todos modos
para leer el paso «Ejecutar las evals», y su parte fuera del informe es lo que resume §3.3.

**Séptima orden** (quitar la etiqueta, al terminar):

```text
https://github.com/jmorenobl/kitlegal/pull/27
la etiqueta evals-prueba-de-red no está puesta
```

## 3. Lo que muestra la ejecución

### 3.1 Pasos y tiempos

`gh run view 34930222593 --json jobs` (`createdAt` 04:48:06Z, `event` `pull_request`, `headSha` `417635e…`):

| Paso | Resultado | Inicio | Fin | Duración |
|---|---|---|---|---|
| Obtener el código del commit evaluado | success | 04:48:11 | 04:48:13 | 2 s |
| Instalar Go y restaurar la caché | success | 04:48:13 | 04:48:59 | 46 s |
| Instalar strace y Claude Code | success | 04:48:59 | 04:49:12 | 13 s |
| Instalar kitlegal y las skills como las deja make install | success | 04:49:12 | 04:49:42 | 30 s |
| Retirar Python del runner | success | 04:49:42 | 04:51:29 | 1 m 47 s |
| Ejecutar las evals | **failure** | 04:51:29 | 04:51:54 | 25 s |

Del paso de instalación: `strace is already the newest version (6.8-0ubuntu2).`, `added 2 packages in 3s` y
`claude --version` → `2.1.270 (Claude Code)`. De `make install`: `CGO_ENABLED=0 go install -trimpath -ldflags "-X
main.version=417635e …" ./cmd/kitlegal`, `instalar-skills: boe-legislacion → /home/runner/work/kitlegal/kitlegal/skills/boe-legislacion`
e `instalar-skills: kitlegal → /home/runner/go/bin/kitlegal`. La variable de entorno del paso de evals lista
`CLAUDE_CODE_OAUTH_TOKEN: ***` (el secreto llega al paso, enmascarado).

### 3.2 Retirada de Python (anexo B)

- `búsqueda: find / ( -path /proc -o -path /sys ) -prune -o ( ( -type f -perm /111 ( -iname python* -o -iname pypy* ) ) -o ( -type l ( -iname python* -o -iname pypy* ) ) -o ( ( -type f -o -type l ) ( -iname libpython* -o -iname libpypy* ) ) ) -print`
  a las 04:49:42,47; la primera línea `retirado:` a las 04:51:05,12: **la búsqueda como root en toda la imagen tarda
  1 m 23 s** y termina con 0 (con `set -euo pipefail`, un `find` con error habría detenido el paso ahí).
- **921 líneas `retirado:`**, entre las 04:51:05 y las 04:51:25 (20 s), ninguna seguida de un error de `rm`: nada
  estaba en un sistema de ficheros de solo lectura. Por árbol: 341 bajo `/opt/hostedtoolcache` (las ocho instalaciones
  de `Python` y `PyPy` de la caché de herramientas, retiradas enteras, y 24 ficheros de `CodeQL/2.26.4`), 315 bajo
  `/var/lib` (202 de `/var/lib/dpkg/info/`: los guiones de mantenimiento ejecutables `python3*.{preinst,postinst,prerm,postrm}`,
  que se llaman como su paquete, y las listas `.list`, `.md5sums`, `.conffiles`, `.shlibs`, `.symbols` y `.triggers`
  de los paquetes `libpython*`, que casan por nombre aunque no sean ejecutables, como anota V60 (1); 113 en capas de
  imágenes de contenedor bajo `/var/lib/docker/overlay2/…`; y 6 más), 143 bajo `/usr/share` (64 de
  `/usr/share/miniconda`, retirado entero; documentación, páginas de manual, completados y `lintian` con nombre
  `python*`), 54 bajo `/usr/local`, 21 bajo `/opt/az` (la CLI de Azure, retirada entera), 19 bajo `/usr/lib`
  (`libpython3.12.so*`, `libpython3.12.a`, `libpython3.12-pic.a`, los `.pc`, `python-config.py` y el Python empaquetado
  del SDK de Google Cloud), 19 bajo `/opt/pipx`, 5 en `/usr/bin` (`python3.12-config`, `python3-config`,
  `python3.12`, `python3`, `python`) y 4 en `/usr/sbin` (las herramientas `python*-bpfcc`).
- Instalaciones retiradas enteras por la regla del prefijo (`<prefijo>/bin/python*` con `<prefijo>/lib/python*`): la
  lista completa, en §3.2.1.
- `búsqueda tras retirar: ninguno` a las 04:51:29,38 (la segunda búsqueda, con la caché de disco caliente, 4 s), la
  comprobación de lo usado sin ningún `la retirada se llevó algo que el job usa`, y la marca de fin. Código 0.

#### 3.2.1 Prefijos retirados enteros

Líneas `retirado:` cuyo objetivo es un prefijo (la ruta encontrada estaba en `<prefijo>/bin` y `<prefijo>/lib` tenía
un `python*` o un `pypy*`, sin nada de lo que el job usa dentro): `/opt/az`; `/opt/hostedtoolcache/PyPy/3.9.19/x64`,
`/opt/hostedtoolcache/PyPy/3.10.16/x64`, `/opt/hostedtoolcache/PyPy/3.11.15/x64`,
`/opt/hostedtoolcache/Python/3.10.21/x64`, `/opt/hostedtoolcache/Python/3.11.16/x64`,
`/opt/hostedtoolcache/Python/3.12.14/x64`, `/opt/hostedtoolcache/Python/3.13.15/x64`,
`/opt/hostedtoolcache/Python/3.14.7/x64`; `/opt/pipx/shared`, `/opt/pipx/venvs/ansible-core`,
`/opt/pipx/venvs/yamllint`; `/usr/lib/google-cloud-sdk/platform/bundledpythonunix`; y `/usr/share/miniconda`. El
intérprete del sistema no se retira por prefijo (`/usr` contiene `bash`, `sudo`, `find`… de lo usado): van sus
ficheros uno a uno (`/usr/bin/python3.12`, `/usr/bin/python3`, `/usr/bin/python`, los dos `*-config` y las bibliotecas
de `/usr/lib`). Las líneas `retirado:` de ficheros dentro de un prefijo que después se retira entero (p. ej. las de
`/opt/pipx/venvs/ansible-core/lib/python3.12/site-packages/…`) salen antes que la del prefijo porque la búsqueda las
listó antes y el bucle comprueba que cada ruta sigue existiendo (`sudo test -e`), como prevé el paso.

### 3.3 Ejecutar las evals (anexo A y registro del paso)

Fuera del informe, el registro del paso muestra: `scripts/evals.sh "boe-legislacion"`; la comprobación 5
(`TestEvalsDelRepositorio` y `TestIdentificadoresDeLasNormas`) en `ok` a las 04:51:42; trece preparaciones de sesión
(`TestPrepararSesion`) en `ok`, una cada 0,75 s aproximadamente entre las 04:51:43 y las 04:51:53, **cada una seguida
de su sesión, que duró menos de un segundo**; `TestInformeDelJob` en `FAIL` con `expected: "aprobado"`, `actual: "fallo"`
y los 54 motivos; las marcas y los dos ficheros; y `make: *** [Makefile:112: evals] Error 1`.

`sin_python` del informe (comprobación 3 del guion, repetida como root antes de la primera sesión):

```text
búsqueda: find / ( -path /proc -o -path /sys ) -prune -o ( ( -type f -perm /111 ( -iname python* -o -iname pypy* ) ) -o ( -type l ( -iname python* -o -iname pypy* ) ) -o ( ( -type f -o -type l ) ( -iname libpython* -o -iname libpypy* ) ) ) -print
usuario: root
resultado: ninguno
```

Cabecera del informe: `Modelo del job: claude-haiku-4-5-20251001`, `Modelos de las sesiones: ninguno`, `Versiones de
Claude Code: ninguna`, `Commit: 417635e39abf5807f8600b79debd6026fc320ce8`. Ficheros mal formados: ninguno.
Invocaciones fuera de lo grabado: **ninguna** (se esperaban las dos de `a9998`). Peticiones llegadas a la red:
«ninguna petición llegó a la red de una fuente». `red` y `fuera_de_lo_grabado` vacíos en el JSON.

Sesiones (de `informe.json`, las trece iguales en lo que importa):

| Sesión | Activa | Activada | `codigo_de_la_sesion` | `fin_de_la_sesion` | Invocaciones | Motivos | Pasa |
|---|---|---|---|---|---|---|---|
| `01-lpac-articulo-21` … `10-et-vacaciones` (diez positivas) | sí | no | 1 | `sin mensajes` | ninguna | `la sesión no terminó: código 1`; `la activación no coincide …`; `comando ausente: …` (uno o dos); `cita ausente: …` | no |
| `01-lpac-articulo-21-prueba-de-red` | sí | no | 1 | `sin mensajes` | ninguna | los mismos cuatro de la eval 01 | no |
| `11-no-activa-programacion`, `12-no-activa-acuerdo-entre-amigos` | no | no | 1 | `sin mensajes` | ninguna | `la sesión no terminó: código 1` | no |

Respuesta vacía en todas. Salida de error, idéntica en las trece (tras la cabecera de licencia del binario y una línea
`import{…}from"/$bunfs/root/chunk-….js"…` de unos 1 500 caracteres, que el informe copia entera):

```text
error: bubblewrap is required for subprocess env scrubbing and isolation. Install with: sudo apt-get install -y bubblewrap, set sandbox.bwrapPath in managed settings, or set CLAUDE_CODE_SUBPROCESS_ENV_SCRUB=0 to disable (loses subprocess isolation).
      at nBn (/$bunfs/root/chunk-y58z5vzc.js:11:25510)
      at A (/$bunfs/root/chunk-xgzajagg.js:11:3535)
      at async <anonymous> (/$bunfs/root/chunk-mpjnf6jh.js:88:4824)
Bun v1.4.3 (Linux x64)
```

Ninguna sesión tiene el motivo `sesión ilegible`: `LeerTrazas` leyó las trece trazas reales del runner (el arranque
de `claude`, un ejecutable de Bun para Linux x86_64, bajo `strace -ff`, con sus hilos y su línea final `+++ exited with
1 +++`) sin encontrar ninguna línea que no admita.

## 4. El defecto y su arreglo

Claude Code 2.1.270 en Linux exige `bwrap` cuando `CLAUDE_CODE_SUBPROCESS_ENV_SCRUB` está activo, y con esa variable
activa fuerza además el modo de permisos a `default` e ignora `--permission-mode bypassPermissions` (leído del binario:
`gates/tarea-T030.md`). Arreglo en **T036**: `bubblewrap` en el paso de instalación del job y `--allowedTools Skill Bash`
en la orden de la sesión, con las decisiones y la comprobación en contenedores desechables que la tarea describe. Lo que
esta ejecución sí confirma del job, y lo que solo dirá el intento 3, está en `gates/tarea-T030.md`.

## 5. Supuestos de research D22

| Supuesto | Qué muestra esta ejecución | Estado |
|---|---|---|
| **S12** (identificar la ejecución) | (1) La API de eventos devuelve, tal cual y en este orden, tres eventos de la etiqueta: `2026-09-15T02:48:51Z	labeled	evals-prueba-de-red	jmorenobl`, `2026-09-15T02:58:51Z	unlabeled	evals-prueba-de-red	jmorenobl` y `2026-09-15T04:48:02Z	labeled	evals-prueba-de-red	jmorenobl` (`gh api --paginate 'repos/{owner}/{repo}/issues/27/events?per_page=100' --jq '.[] \| select(.event == "labeled" or .event == "unlabeled") \| "\(.created_at)\t\(.event)\t\(.label.name)\t\(.actor.login)"'`); las órdenes ordenan los instantes y eligen `2026-09-15T04:48:02Z`. (2) `created_at` (`…T04:48:02Z`) y `createdAt` (`…T04:48:06Z`) con el mismo formato ISO 8601 UTC con `Z`. (3) La ejecución se creó 4 s después del evento. (5) **Ejercido**: la rama ya tenía una ejecución de `evals` anterior (34922606273, del intento 1, creada tras el primer `labeled`), y la orden no la eligió: `posteriores_a_la_etiqueta` lista solo 34930222593. (6) `workflowName` `evals` en `posteriores_a_la_etiqueta`, aunque `evals.yml` solo está en la rama de la propuesta. (4) Sin ejercer: la etiqueta no estaba puesta al empezar (la quitó el intento 1) y al final `gh pr view` la listaba, así que se quitó estando puesta | **se cumple** en (1), (2), (3), (5) y (6); (4), sin ejercer |
| **S2** (lo que trae `ubuntu-24.04`) | `sudo` sin contraseña en el paso de retirada y `sudo -n` en la comprobación 3 (`usuario: root`). `strace` 6.8 ya en la imagen (`strace is already the newest version (6.8-0ubuntu2).`). `npm install -g @anthropic-ai/claude-code@2.1.270`: `added 2 packages in 3s`, `claude --version` → `2.1.270 (Claude Code)`. `timeout`: ejercido en las trece sesiones (devolvió el 1 de `claude`). GNU findutils y coreutils: la línea `búsqueda:` seguida de 921 `retirado:` y de `búsqueda tras retirar: ninguno` (`find` con `-perm /111`, `-iname`, `-prune`, y `-H … -quit` en la regla del prefijo, que se aplicó a varias instalaciones; `readlink -e` en lo usado; `rm -rf` sin ningún error) | **se cumple** en todo lo ejercido. **Nuevo**: la imagen no trae `bubblewrap`, que el `ENV_SCRUB` de Claude Code exige; que `apt` lo instale y que `bwrap` cree espacios de nombres de usuario sin privilegios en el runner lo comprueba el intento 3 (T036 lo añade a S2) |
| **S7** (retirada de Python) | (1) **Lo que trae**: 921 rutas con nombre `python*`, `pypy*`, `libpython*` o `libpypy*` (§3.2): el intérprete del sistema y sus enlaces en `/usr/bin`, sus bibliotecas en `/usr/lib`, la caché de herramientas (`/opt/hostedtoolcache/Python` y `/opt/hostedtoolcache/PyPy`), los entornos de `pipx`, el Python empaquetado del SDK de Google Cloud y de la CLI de Azure (`/opt/az`), los guiones de mantenimiento de dpkg y capas de imágenes de contenedor; ninguna ruta con otro nombre puede verse. (2) **Buscar como root**: `find` recorrió toda la imagen salvo `/proc` y `/sys` en 1 m 23 s y terminó con 0. (3) **Retirar**: 921 borrados sin ningún error; nada en solo lectura. (4) **No romper el job**: la comprobación de lo usado pasó, y después del paso corrieron `bash`, `sudo`, `find` (comprobación 3), `go` (cinco tests), `make`, `git`, `timeout`, `strace` y `claude` (que arrancó y falló por `bubblewrap`, no por nada retirado: el mensaje nombra la causa, y `bwrap` no casa con la búsqueda) | **se cumple** en (1)-(4) |
| **S4** (formato de `strace -ff` en el runner) | Ninguna invocación de applet ni conexión: la sesión murió antes. **Parcial**: las trece trazas reales de x86_64 (el arranque de `claude`, un ejecutable de Bun con sus hilos, hasta `+++ exited with 1 +++`) las leyó `LeerTrazas` sin `sesión ilegible` con `cortada` falso (código 1), así que las líneas de creación de hilos de Claude Code en amd64 y las líneas finales tienen las formas de data-model §9 | **sin evidencia** de `connect` ni de la atribución por invocación; se cumple en las líneas de hilos y finales de Claude Code |
| **S9** (una sesión cabe en 240 s) | Ninguna sesión pasó del arranque (menos de un segundo cada una); ninguna con código 124 ni 137 | sin evidencia |
| **S10** (códigos de la sesión) | Las trece con `codigo_de_la_sesion` 1: `timeout --kill-after=10s 240s strace -ff …` devolvió el código de `claude` cuando este terminó por sí mismo, y el guion lo escribió en `codigo-de-la-sesion` | **se cumple** en lo ejercido (el código de `claude` se propaga); 0 sin ejercer |
| S1 (fuera de la lista de esta tarea) | `pull_request` con `types: [labeled]` ejecutó el `evals.yml` de la rama (`COMMIT_EVALUADO: 417635e…`, `PRUEBA_DE_RED: true`) y el secreto llegó al paso (`CLAUDE_CODE_OAUTH_TOKEN: ***` en su entorno); que llegue a la sesión, sin ejercer | se cumple en lo ejercido |
| S5, S6 (Claude Code en `-p`: skills, proxy, `ENV_SCRUB`, credencial, modelo) | **Difieren en lo que dan por hecho del `ENV_SCRUB`**: con la variable activa, Claude Code 2.1.270 en Linux no arranca sin `bubblewrap` y, según el binario, fuerza el modo de permisos a `default` (§4); lo demás (skills cargadas, proxy heredado, autenticación, modelo aceptado, token retirado del entorno de las órdenes) sin ejercer | **difieren** en el `ENV_SCRUB`; el resto, sin evidencia |
| S11 (solo `api.anthropic.com`) | Ninguna petición al modelo | sin evidencia |

## Anexos

Los dos volcados siguientes son la salida entera de la quinta y de la sexta orden de quickstart §12.2, tal cual las
imprimieron (cada línea con el prefijo de tarea, paso e instante que pone `gh run view --log`).

## Anexo A · Quinta orden de §12.2: informe entre marcas, tal cual

```text
evals	Ejecutar las evals	2026-09-15T04:51:54.1049437Z --- inicio de informe.md ---
evals	Ejecutar las evals	2026-09-15T04:51:54.1049864Z # Informe de evals de boe-legislacion
evals	Ejecutar las evals	2026-09-15T04:51:54.1050182Z 
evals	Ejecutar las evals	2026-09-15T04:51:54.1050375Z ## Veredicto
evals	Ejecutar las evals	2026-09-15T04:51:54.1050607Z 
evals	Ejecutar las evals	2026-09-15T04:51:54.1050792Z Veredicto: fallo
evals	Ejecutar las evals	2026-09-15T04:51:54.1051032Z 
evals	Ejecutar las evals	2026-09-15T04:51:54.1051217Z Motivos:
evals	Ejecutar las evals	2026-09-15T04:51:54.1051523Z 
evals	Ejecutar las evals	2026-09-15T04:51:54.1052095Z - 01-lpac-articulo-21: la sesión no terminó: código 1
evals	Ejecutar las evals	2026-09-15T04:51:54.1053120Z - 01-lpac-articulo-21: la activación no coincide: se esperaba que la skill boe-legislacion se activara y no se activó
evals	Ejecutar las evals	2026-09-15T04:51:54.1054042Z - 01-lpac-articulo-21: comando ausente: bloque boe BOE-A-2015-10565 a21
evals	Ejecutar las evals	2026-09-15T04:51:54.1054693Z - 01-lpac-articulo-21: cita ausente: BOE-A-2015-10565 a21
evals	Ejecutar las evals	2026-09-15T04:51:54.1055469Z - 01-lpac-articulo-21-prueba-de-red: la sesión no terminó: código 1
evals	Ejecutar las evals	2026-09-15T04:51:54.1056678Z - 01-lpac-articulo-21-prueba-de-red: la activación no coincide: se esperaba que la skill boe-legislacion se activara y no se activó
evals	Ejecutar las evals	2026-09-15T04:51:54.1057764Z - 01-lpac-articulo-21-prueba-de-red: comando ausente: bloque boe BOE-A-2015-10565 a21
evals	Ejecutar las evals	2026-09-15T04:51:54.1058558Z - 01-lpac-articulo-21-prueba-de-red: cita ausente: BOE-A-2015-10565 a21
evals	Ejecutar las evals	2026-09-15T04:51:54.1059616Z - 02-lcsp-contrato-menor: la sesión no terminó: código 1
evals	Ejecutar las evals	2026-09-15T04:51:54.1060589Z - 02-lcsp-contrato-menor: la activación no coincide: se esperaba que la skill boe-legislacion se activara y no se activó
evals	Ejecutar las evals	2026-09-15T04:51:54.1061540Z - 02-lcsp-contrato-menor: comando ausente: boe indice BOE-A-2017-12902
evals	Ejecutar las evals	2026-09-15T04:51:54.1062263Z - 02-lcsp-contrato-menor: comando ausente: bloque boe BOE-A-2017-12902 a1-30
evals	Ejecutar las evals	2026-09-15T04:51:54.1062946Z - 02-lcsp-contrato-menor: cita ausente: BOE-A-2017-12902 a1-30
evals	Ejecutar las evals	2026-09-15T04:51:54.1063935Z - 03-lrbrl-atribuciones-del-pleno: la sesión no terminó: código 1
evals	Ejecutar las evals	2026-09-15T04:51:54.1065178Z - 03-lrbrl-atribuciones-del-pleno: la activación no coincide: se esperaba que la skill boe-legislacion se activara y no se activó
evals	Ejecutar las evals	2026-09-15T04:51:54.1066370Z - 03-lrbrl-atribuciones-del-pleno: comando ausente: boe indice BOE-A-1985-5392
evals	Ejecutar las evals	2026-09-15T04:51:54.1067193Z - 03-lrbrl-atribuciones-del-pleno: comando ausente: bloque boe BOE-A-1985-5392 a22
evals	Ejecutar las evals	2026-09-15T04:51:54.1067959Z - 03-lrbrl-atribuciones-del-pleno: cita ausente: BOE-A-1985-5392 a22
evals	Ejecutar las evals	2026-09-15T04:51:54.1068781Z - 04-lgt-prescripcion: la sesión no terminó: código 1
evals	Ejecutar las evals	2026-09-15T04:51:54.1069990Z - 04-lgt-prescripcion: la activación no coincide: se esperaba que la skill boe-legislacion se activara y no se activó
evals	Ejecutar las evals	2026-09-15T04:51:54.1070894Z - 04-lgt-prescripcion: comando ausente: boe indice BOE-A-2003-23186
evals	Ejecutar las evals	2026-09-15T04:51:54.1071599Z - 04-lgt-prescripcion: comando ausente: bloque boe BOE-A-2003-23186 a66
evals	Ejecutar las evals	2026-09-15T04:51:54.1072234Z - 04-lgt-prescripcion: cita ausente: BOE-A-2003-23186 a66
evals	Ejecutar las evals	2026-09-15T04:51:54.1073144Z - 05-trlrhl-impuestos-municipales: la sesión no terminó: código 1
evals	Ejecutar las evals	2026-09-15T04:51:54.1074216Z - 05-trlrhl-impuestos-municipales: la activación no coincide: se esperaba que la skill boe-legislacion se activara y no se activó
evals	Ejecutar las evals	2026-09-15T04:51:54.1075284Z - 05-trlrhl-impuestos-municipales: comando ausente: boe indice BOE-A-2004-4214
evals	Ejecutar las evals	2026-09-15T04:51:54.1076138Z - 05-trlrhl-impuestos-municipales: comando ausente: bloque boe BOE-A-2004-4214 a59
evals	Ejecutar las evals	2026-09-15T04:51:54.1076927Z - 05-trlrhl-impuestos-municipales: cita ausente: BOE-A-2004-4214 a59
evals	Ejecutar las evals	2026-09-15T04:51:54.1077742Z - 06-irpf-rendimientos-del-trabajo: la sesión no terminó: código 1
evals	Ejecutar las evals	2026-09-15T04:51:54.1080463Z - 06-irpf-rendimientos-del-trabajo: la activación no coincide: se esperaba que la skill boe-legislacion se activara y no se activó
evals	Ejecutar las evals	2026-09-15T04:51:54.1081424Z - 06-irpf-rendimientos-del-trabajo: comando ausente: boe indice BOE-A-2006-20764
evals	Ejecutar las evals	2026-09-15T04:51:54.1082201Z - 06-irpf-rendimientos-del-trabajo: comando ausente: bloque boe BOE-A-2006-20764 a17
evals	Ejecutar las evals	2026-09-15T04:51:54.1082863Z - 06-irpf-rendimientos-del-trabajo: cita ausente: BOE-A-2006-20764 a17
evals	Ejecutar las evals	2026-09-15T04:51:54.1083573Z - 07-lrjsp-principio-de-legalidad: la sesión no terminó: código 1
evals	Ejecutar las evals	2026-09-15T04:51:54.1084561Z - 07-lrjsp-principio-de-legalidad: la activación no coincide: se esperaba que la skill boe-legislacion se activara y no se activó
evals	Ejecutar las evals	2026-09-15T04:51:54.1085390Z - 07-lrjsp-principio-de-legalidad: comando ausente: boe indice BOE-A-2015-10566
evals	Ejecutar las evals	2026-09-15T04:51:54.1085960Z - 07-lrjsp-principio-de-legalidad: comando ausente: bloque boe BOE-A-2015-10566 a25
evals	Ejecutar las evals	2026-09-15T04:51:54.1086402Z - 07-lrjsp-principio-de-legalidad: cita ausente: BOE-A-2015-10566 a25
evals	Ejecutar las evals	2026-09-15T04:51:54.1086838Z - 08-ltaibg-plazo-de-resolucion: la sesión no terminó: código 1
evals	Ejecutar las evals	2026-09-15T04:51:54.1087427Z - 08-ltaibg-plazo-de-resolucion: la activación no coincide: se esperaba que la skill boe-legislacion se activara y no se activó
evals	Ejecutar las evals	2026-09-15T04:51:54.1087971Z - 08-ltaibg-plazo-de-resolucion: comando ausente: boe indice BOE-A-2013-12887
evals	Ejecutar las evals	2026-09-15T04:51:54.1088424Z - 08-ltaibg-plazo-de-resolucion: comando ausente: bloque boe BOE-A-2013-12887 a20
evals	Ejecutar las evals	2026-09-15T04:51:54.1088827Z - 08-ltaibg-plazo-de-resolucion: cita ausente: BOE-A-2013-12887 a20
evals	Ejecutar las evals	2026-09-15T04:51:54.1089411Z - 09-constitucion-articulo-140: la sesión no terminó: código 1
evals	Ejecutar las evals	2026-09-15T04:51:54.1089986Z - 09-constitucion-articulo-140: la activación no coincide: se esperaba que la skill boe-legislacion se activara y no se activó
evals	Ejecutar las evals	2026-09-15T04:51:54.1090520Z - 09-constitucion-articulo-140: comando ausente: bloque boe BOE-A-1978-31229 a140
evals	Ejecutar las evals	2026-09-15T04:51:54.1090909Z - 09-constitucion-articulo-140: cita ausente: BOE-A-1978-31229 a140
evals	Ejecutar las evals	2026-09-15T04:51:54.1091523Z - 10-et-vacaciones: la sesión no terminó: código 1
evals	Ejecutar las evals	2026-09-15T04:51:54.1092259Z - 10-et-vacaciones: la activación no coincide: se esperaba que la skill boe-legislacion se activara y no se activó
evals	Ejecutar las evals	2026-09-15T04:51:54.1092893Z - 10-et-vacaciones: comando ausente: boe indice BOE-A-2015-11430
evals	Ejecutar las evals	2026-09-15T04:51:54.1093436Z - 10-et-vacaciones: comando ausente: bloque boe BOE-A-2015-11430 a38
evals	Ejecutar las evals	2026-09-15T04:51:54.1094040Z - 10-et-vacaciones: cita ausente: BOE-A-2015-11430 a38
evals	Ejecutar las evals	2026-09-15T04:51:54.1094776Z - 11-no-activa-programacion: la sesión no terminó: código 1
evals	Ejecutar las evals	2026-09-15T04:51:54.1095343Z - 12-no-activa-acuerdo-entre-amigos: la sesión no terminó: código 1
evals	Ejecutar las evals	2026-09-15T04:51:54.1095621Z 
evals	Ejecutar las evals	2026-09-15T04:51:54.1107028Z make: *** [Makefile:112: evals] Error 1
evals	Ejecutar las evals	2026-09-15T04:51:54.1107391Z ## Cabecera
evals	Ejecutar las evals	2026-09-15T04:51:54.1107564Z 
evals	Ejecutar las evals	2026-09-15T04:51:54.1107728Z Modelo del job: claude-haiku-4-5-20251001
evals	Ejecutar las evals	2026-09-15T04:51:54.1107987Z 
evals	Ejecutar las evals	2026-09-15T04:51:54.1108132Z Modelos de las sesiones: ninguno
evals	Ejecutar las evals	2026-09-15T04:51:54.1108351Z 
evals	Ejecutar las evals	2026-09-15T04:51:54.1108486Z Versiones de Claude Code: ninguna
evals	Ejecutar las evals	2026-09-15T04:51:54.1108710Z 
evals	Ejecutar las evals	2026-09-15T04:51:54.1109115Z Commit: 417635e39abf5807f8600b79debd6026fc320ce8
evals	Ejecutar las evals	2026-09-15T04:51:54.1109385Z 
evals	Ejecutar las evals	2026-09-15T04:51:54.1109613Z ## Comprobación sin Python
evals	Ejecutar las evals	2026-09-15T04:51:54.1109822Z 
evals	Ejecutar las evals	2026-09-15T04:51:54.1110099Z ```text
evals	Ejecutar las evals	2026-09-15T04:51:54.1111379Z búsqueda: find / ( -path /proc -o -path /sys ) -prune -o ( ( -type f -perm /111 ( -iname python* -o -iname pypy* ) ) -o ( -type l ( -iname python* -o -iname pypy* ) ) -o ( ( -type f -o -type l ) ( -iname libpython* -o -iname libpypy* ) ) ) -print
evals	Ejecutar las evals	2026-09-15T04:51:54.1112458Z usuario: root
evals	Ejecutar las evals	2026-09-15T04:51:54.1112734Z resultado: ninguno
evals	Ejecutar las evals	2026-09-15T04:51:54.1112977Z ```
evals	Ejecutar las evals	2026-09-15T04:51:54.1113099Z 
evals	Ejecutar las evals	2026-09-15T04:51:54.1113221Z ## Ficheros mal formados
evals	Ejecutar las evals	2026-09-15T04:51:54.1113414Z 
evals	Ejecutar las evals	2026-09-15T04:51:54.1113515Z ninguno
evals	Ejecutar las evals	2026-09-15T04:51:54.1113644Z 
evals	Ejecutar las evals	2026-09-15T04:51:54.1113781Z ## Invocaciones fuera de lo grabado
evals	Ejecutar las evals	2026-09-15T04:51:54.1113964Z 
evals	Ejecutar las evals	2026-09-15T04:51:54.1114030Z ninguna
evals	Ejecutar las evals	2026-09-15T04:51:54.1114114Z 
evals	Ejecutar las evals	2026-09-15T04:51:54.1114201Z ## Peticiones llegadas a la red
evals	Ejecutar las evals	2026-09-15T04:51:54.1114341Z 
evals	Ejecutar las evals	2026-09-15T04:51:54.1114494Z ninguna petición llegó a la red de una fuente
evals	Ejecutar las evals	2026-09-15T04:51:54.1114661Z 
evals	Ejecutar las evals	2026-09-15T04:51:54.1114730Z ## Sesiones
evals	Ejecutar las evals	2026-09-15T04:51:54.1114835Z 
evals	Ejecutar las evals	2026-09-15T04:51:54.1115147Z | Sesión | Eval | Activa | Activada | Sesión terminada | Comandos ausentes | Citas ausentes | Resultado |
evals	Ejecutar las evals	2026-09-15T04:51:54.1115562Z | --- | --- | --- | --- | --- | --- | --- | --- |
evals	Ejecutar las evals	2026-09-15T04:51:54.1116149Z | 01-lpac-articulo-21 | 01-lpac-articulo-21.yaml | sí | no | no (código 1) | bloque boe BOE-A-2015-10565 a21 | BOE-A-2015-10565 a21 | no pasa |
evals	Ejecutar las evals	2026-09-15T04:51:54.1116987Z | 01-lpac-articulo-21-prueba-de-red | 01-lpac-articulo-21.yaml | sí | no | no (código 1) | bloque boe BOE-A-2015-10565 a21 | BOE-A-2015-10565 a21 | no pasa |
evals	Ejecutar las evals	2026-09-15T04:51:54.1117902Z | 02-lcsp-contrato-menor | 02-lcsp-contrato-menor.yaml | sí | no | no (código 1) | boe indice BOE-A-2017-12902, bloque boe BOE-A-2017-12902 a1-30 | BOE-A-2017-12902 a1-30 | no pasa |
evals	Ejecutar las evals	2026-09-15T04:51:54.1119151Z | 03-lrbrl-atribuciones-del-pleno | 03-lrbrl-atribuciones-del-pleno.yaml | sí | no | no (código 1) | boe indice BOE-A-1985-5392, bloque boe BOE-A-1985-5392 a22 | BOE-A-1985-5392 a22 | no pasa |
evals	Ejecutar las evals	2026-09-15T04:51:54.1120224Z | 04-lgt-prescripcion | 04-lgt-prescripcion.yaml | sí | no | no (código 1) | boe indice BOE-A-2003-23186, bloque boe BOE-A-2003-23186 a66 | BOE-A-2003-23186 a66 | no pasa |
evals	Ejecutar las evals	2026-09-15T04:51:54.1121234Z | 05-trlrhl-impuestos-municipales | 05-trlrhl-impuestos-municipales.yaml | sí | no | no (código 1) | boe indice BOE-A-2004-4214, bloque boe BOE-A-2004-4214 a59 | BOE-A-2004-4214 a59 | no pasa |
evals	Ejecutar las evals	2026-09-15T04:51:54.1122302Z | 06-irpf-rendimientos-del-trabajo | 06-irpf-rendimientos-del-trabajo.yaml | sí | no | no (código 1) | boe indice BOE-A-2006-20764, bloque boe BOE-A-2006-20764 a17 | BOE-A-2006-20764 a17 | no pasa |
evals	Ejecutar las evals	2026-09-15T04:51:54.1123362Z | 07-lrjsp-principio-de-legalidad | 07-lrjsp-principio-de-legalidad.yaml | sí | no | no (código 1) | boe indice BOE-A-2015-10566, bloque boe BOE-A-2015-10566 a25 | BOE-A-2015-10566 a25 | no pasa |
evals	Ejecutar las evals	2026-09-15T04:51:54.1124393Z | 08-ltaibg-plazo-de-resolucion | 08-ltaibg-plazo-de-resolucion.yaml | sí | no | no (código 1) | boe indice BOE-A-2013-12887, bloque boe BOE-A-2013-12887 a20 | BOE-A-2013-12887 a20 | no pasa |
evals	Ejecutar las evals	2026-09-15T04:51:54.1125355Z | 09-constitucion-articulo-140 | 09-constitucion-articulo-140.yaml | sí | no | no (código 1) | bloque boe BOE-A-1978-31229 a140 | BOE-A-1978-31229 a140 | no pasa |
evals	Ejecutar las evals	2026-09-15T04:51:54.1126378Z | 10-et-vacaciones | 10-et-vacaciones.yaml | sí | no | no (código 1) | boe indice BOE-A-2015-11430, bloque boe BOE-A-2015-11430 a38 | BOE-A-2015-11430 a38 | no pasa |
evals	Ejecutar las evals	2026-09-15T04:51:54.1127132Z | 11-no-activa-programacion | 11-no-activa-programacion.yaml | no | no | no (código 1) | ninguno | ninguna | no pasa |
evals	Ejecutar las evals	2026-09-15T04:51:54.1127878Z | 12-no-activa-acuerdo-entre-amigos | 12-no-activa-acuerdo-entre-amigos.yaml | no | no | no (código 1) | ninguno | ninguna | no pasa |
evals	Ejecutar las evals	2026-09-15T04:51:54.1128262Z 
evals	Ejecutar las evals	2026-09-15T04:51:54.1128385Z ## Sesión 01-lpac-articulo-21
evals	Ejecutar las evals	2026-09-15T04:51:54.1128525Z 
evals	Ejecutar las evals	2026-09-15T04:51:54.1128618Z Eval: 01-lpac-articulo-21.yaml
evals	Ejecutar las evals	2026-09-15T04:51:54.1128748Z 
evals	Ejecutar las evals	2026-09-15T04:51:54.1128824Z Pregunta:
evals	Ejecutar las evals	2026-09-15T04:51:54.1129367Z 
evals	Ejecutar las evals	2026-09-15T04:51:54.1129445Z ```text
evals	Ejecutar las evals	2026-09-15T04:51:54.1129670Z ¿qué dice el art. 21 de la Ley 39/2015?
evals	Ejecutar las evals	2026-09-15T04:51:54.1129883Z ```
evals	Ejecutar las evals	2026-09-15T04:51:54.1129978Z 
evals	Ejecutar las evals	2026-09-15T04:51:54.1130056Z Invocaciones: ninguna
evals	Ejecutar las evals	2026-09-15T04:51:54.1130170Z 
evals	Ejecutar las evals	2026-09-15T04:51:54.1130268Z Respuesta: vacía
evals	Ejecutar las evals	2026-09-15T04:51:54.1130368Z 
evals	Ejecutar las evals	2026-09-15T04:51:54.1130468Z Motivos de la sesión:
evals	Ejecutar las evals	2026-09-15T04:51:54.1130574Z 
evals	Ejecutar las evals	2026-09-15T04:51:54.1130689Z - la sesión no terminó: código 1
evals	Ejecutar las evals	2026-09-15T04:51:54.1130825Z 
evals	Ejecutar las evals	2026-09-15T04:51:54.1130900Z Salida de error:
evals	Ejecutar las evals	2026-09-15T04:51:54.1130991Z 
evals	Ejecutar las evals	2026-09-15T04:51:54.1131061Z ```text
evals	Ejecutar las evals	2026-09-15T04:51:54.1131338Z  6 | // and may be used to improve Anthropic's products, including training models.
evals	Ejecutar las evals	2026-09-15T04:51:54.1131757Z  7 | // You are responsible for reviewing any code suggestions before use.
evals	Ejecutar las evals	2026-09-15T04:51:54.1132036Z  8 | 
evals	Ejecutar las evals	2026-09-15T04:51:54.1132638Z  9 | // (c) Anthropic PBC. All rights reserved. Use is subject to the Legal Agreements outlined here: https://code.claude.com/docs/en/legal-and-compliance.
evals	Ejecutar las evals	2026-09-15T04:51:54.1133133Z 10 | 
evals	Ejecutar las evals	2026-09-15T04:51:54.1135761Z 11 | import{W,B,he,fN}from"/$bunfs/root/chunk-265sathk.js";import{He,fo,$e}from"/$bunfs/root/chunk-7khdpcz9.js";import{Ude,Yt}from"/$bunfs/root/chunk-4a0graca.js";import{Wi,Dne,Oct,iwe,qct,Lne,Jq}from"/$bunfs/root/chunk-hk7hcqnf.js";import{f}from"/$bunfs/root/chunk-xpzkpaw3.js";import{hx,a}from"/$bunfs/root/chunk-crze9c5c.js";import{Va,l,Mt,hT}from"/$bunfs/root/chunk-a382p5z3.js";import{Mu,qc,K,le,t}from"/$bunfs/root/chunk-0y8dccaw.js";import{lMe,n2n,act,r2n,cMe,Jon,vAr,VSe,PVe,lct,g6,cct,GLt,EAr,uct,IS,zc}from"/$bunfs/root/chunk-ebh8j0r9.js";import{Br}from"/$bunfs/root/chunk-jrz35ppm.js";import{Lue}from"/$bunfs/root/chunk-z0zff54w.js";import{ron}from"/$bunfs/root/chunk-armd3k2m.js";import{gn}from"/$bunfs/root/chunk-aj7ffynp.js";import{aBn}from"/$bunfs/root/chunk-vb37gsz0.js";import{Bo}from"/$bunfs/root/chunk-gk39j4sy.js";import{ie,u,V,R}from"/$bunfs/root/chunk-9kr1d91j.js";import{Y}from"/$bunfs/root/chunk-4112m93q.js";import{homedir as be}from"os";import{dirname as Q,posix as D}from"path";var L=(e)=>JSON.stringif
evals	Ejecutar las evals	2026-09-15T04:51:54.1138331Z 
evals	Ejecutar las evals	2026-09-15T04:51:54.1139186Z error: bubblewrap is required for subprocess env scrubbing and isolation. Install with: sudo apt-get install -y bubblewrap, set sandbox.bwrapPath in managed settings, or set CLAUDE_CODE_SUBPROCESS_ENV_SCRUB=0 to disable (loses subprocess isolation).
evals	Ejecutar las evals	2026-09-15T04:51:54.1140033Z       at nBn (/$bunfs/root/chunk-y58z5vzc.js:11:25510)
evals	Ejecutar las evals	2026-09-15T04:51:54.1140358Z       at A (/$bunfs/root/chunk-xgzajagg.js:11:3535)
evals	Ejecutar las evals	2026-09-15T04:51:54.1140725Z       at async <anonymous> (/$bunfs/root/chunk-mpjnf6jh.js:88:4824)
evals	Ejecutar las evals	2026-09-15T04:51:54.1140919Z 
evals	Ejecutar las evals	2026-09-15T04:51:54.1163334Z ##[error]
evals	Ejecutar las evals	      at nBn (/$bunfs/root/chunk-y58z5vzc.js:11:25510)
evals	Ejecutar las evals	      at A (/$bunfs/root/chunk-xgzajagg.js:11:3535)
evals	Ejecutar las evals	      at async <anonymous> (/$bunfs/root/chunk-mpjnf6jh.js:88:4824)
evals	Ejecutar las evals	2026-09-15T04:51:54.1169237Z 
evals	Ejecutar las evals	2026-09-15T04:51:54.1169363Z Bun v1.4.3 (Linux x64)
evals	Ejecutar las evals	2026-09-15T04:51:54.1169578Z ```
evals	Ejecutar las evals	2026-09-15T04:51:54.1169671Z 
evals	Ejecutar las evals	2026-09-15T04:51:54.1169881Z ## Sesión 01-lpac-articulo-21-prueba-de-red
evals	Ejecutar las evals	2026-09-15T04:51:54.1170063Z 
evals	Ejecutar las evals	2026-09-15T04:51:54.1170326Z Eval: 01-lpac-articulo-21.yaml
evals	Ejecutar las evals	2026-09-15T04:51:54.1170471Z 
evals	Ejecutar las evals	2026-09-15T04:51:54.1170544Z Pregunta:
evals	Ejecutar las evals	2026-09-15T04:51:54.1170640Z 
evals	Ejecutar las evals	2026-09-15T04:51:54.1170711Z ```text
evals	Ejecutar las evals	2026-09-15T04:51:54.1170924Z ¿qué dice el art. 21 de la Ley 39/2015?
evals	Ejecutar las evals	2026-09-15T04:51:54.1171084Z 
evals	Ejecutar las evals	2026-09-15T04:51:54.1171957Z Antes de responder, ejecuta también exactamente estas dos órdenes y di qué devolvieron: `~/.claude/skills/boe-legislacion/scripts/boe articulo BOE-A-2015-10565 a9998 --json` y `~/.claude/skills/boe-legislacion/scripts/boe articulo BOE-A-2015-10565 a9998 --offline --json`.
evals	Ejecutar las evals	2026-09-15T04:51:54.1172784Z ```
evals	Ejecutar las evals	2026-09-15T04:51:54.1172873Z 
evals	Ejecutar las evals	2026-09-15T04:51:54.1172954Z Invocaciones: ninguna
evals	Ejecutar las evals	2026-09-15T04:51:54.1173073Z 
evals	Ejecutar las evals	2026-09-15T04:51:54.1173177Z Respuesta: vacía
evals	Ejecutar las evals	2026-09-15T04:51:54.1173279Z 
evals	Ejecutar las evals	2026-09-15T04:51:54.1173378Z Motivos de la sesión:
evals	Ejecutar las evals	2026-09-15T04:51:54.1173634Z 
evals	Ejecutar las evals	2026-09-15T04:51:54.1173754Z - la sesión no terminó: código 1
evals	Ejecutar las evals	2026-09-15T04:51:54.1173898Z 
evals	Ejecutar las evals	2026-09-15T04:51:54.1173976Z Salida de error:
evals	Ejecutar las evals	2026-09-15T04:51:54.1174081Z 
evals	Ejecutar las evals	2026-09-15T04:51:54.1174152Z ```text
evals	Ejecutar las evals	2026-09-15T04:51:54.1174440Z  6 | // and may be used to improve Anthropic's products, including training models.
evals	Ejecutar las evals	2026-09-15T04:51:54.1174858Z  7 | // You are responsible for reviewing any code suggestions before use.
evals	Ejecutar las evals	2026-09-15T04:51:54.1175137Z  8 | 
evals	Ejecutar las evals	2026-09-15T04:51:54.1175654Z  9 | // (c) Anthropic PBC. All rights reserved. Use is subject to the Legal Agreements outlined here: https://code.claude.com/docs/en/legal-and-compliance.
evals	Ejecutar las evals	2026-09-15T04:51:54.1176149Z 10 | 
evals	Ejecutar las evals	2026-09-15T04:51:54.1178752Z 11 | import{W,B,he,fN}from"/$bunfs/root/chunk-265sathk.js";import{He,fo,$e}from"/$bunfs/root/chunk-7khdpcz9.js";import{Ude,Yt}from"/$bunfs/root/chunk-4a0graca.js";import{Wi,Dne,Oct,iwe,qct,Lne,Jq}from"/$bunfs/root/chunk-hk7hcqnf.js";import{f}from"/$bunfs/root/chunk-xpzkpaw3.js";import{hx,a}from"/$bunfs/root/chunk-crze9c5c.js";import{Va,l,Mt,hT}from"/$bunfs/root/chunk-a382p5z3.js";import{Mu,qc,K,le,t}from"/$bunfs/root/chunk-0y8dccaw.js";import{lMe,n2n,act,r2n,cMe,Jon,vAr,VSe,PVe,lct,g6,cct,GLt,EAr,uct,IS,zc}from"/$bunfs/root/chunk-ebh8j0r9.js";import{Br}from"/$bunfs/root/chunk-jrz35ppm.js";import{Lue}from"/$bunfs/root/chunk-z0zff54w.js";import{ron}from"/$bunfs/root/chunk-armd3k2m.js";import{gn}from"/$bunfs/root/chunk-aj7ffynp.js";import{aBn}from"/$bunfs/root/chunk-vb37gsz0.js";import{Bo}from"/$bunfs/root/chunk-gk39j4sy.js";import{ie,u,V,R}from"/$bunfs/root/chunk-9kr1d91j.js";import{Y}from"/$bunfs/root/chunk-4112m93q.js";import{homedir as be}from"os";import{dirname as Q,posix as D}from"path";var L=(e)=>JSON.stringif
evals	Ejecutar las evals	2026-09-15T04:51:54.1181503Z 
evals	Ejecutar las evals	2026-09-15T04:51:54.1182157Z error: bubblewrap is required for subprocess env scrubbing and isolation. Install with: sudo apt-get install -y bubblewrap, set sandbox.bwrapPath in managed settings, or set CLAUDE_CODE_SUBPROCESS_ENV_SCRUB=0 to disable (loses subprocess isolation).
evals	Ejecutar las evals	2026-09-15T04:51:54.1182983Z       at nBn (/$bunfs/root/chunk-y58z5vzc.js:11:25510)
evals	Ejecutar las evals	2026-09-15T04:51:54.1183308Z       at A (/$bunfs/root/chunk-xgzajagg.js:11:3535)
evals	Ejecutar las evals	2026-09-15T04:51:54.1183671Z       at async <anonymous> (/$bunfs/root/chunk-mpjnf6jh.js:88:4824)
evals	Ejecutar las evals	2026-09-15T04:51:54.1183869Z 
evals	Ejecutar las evals	2026-09-15T04:51:54.1184997Z ##[error]
evals	Ejecutar las evals	      at nBn (/$bunfs/root/chunk-y58z5vzc.js:11:25510)
evals	Ejecutar las evals	      at A (/$bunfs/root/chunk-xgzajagg.js:11:3535)
evals	Ejecutar las evals	      at async <anonymous> (/$bunfs/root/chunk-mpjnf6jh.js:88:4824)
evals	Ejecutar las evals	2026-09-15T04:51:54.1185869Z 
evals	Ejecutar las evals	2026-09-15T04:51:54.1185953Z Bun v1.4.3 (Linux x64)
evals	Ejecutar las evals	2026-09-15T04:51:54.1186132Z ```
evals	Ejecutar las evals	2026-09-15T04:51:54.1186217Z 
evals	Ejecutar las evals	2026-09-15T04:51:54.1186356Z ## Sesión 02-lcsp-contrato-menor
evals	Ejecutar las evals	2026-09-15T04:51:54.1186503Z 
evals	Ejecutar las evals	2026-09-15T04:51:54.1186601Z Eval: 02-lcsp-contrato-menor.yaml
evals	Ejecutar las evals	2026-09-15T04:51:54.1186737Z 
evals	Ejecutar las evals	2026-09-15T04:51:54.1186809Z Pregunta:
evals	Ejecutar las evals	2026-09-15T04:51:54.1186904Z 
evals	Ejecutar las evals	2026-09-15T04:51:54.1186973Z ```text
evals	Ejecutar las evals	2026-09-15T04:51:54.1187361Z ¿Qué debe incluir el expediente de un contrato menor según la Ley de Contratos del Sector Público?
evals	Ejecutar las evals	2026-09-15T04:51:54.1187731Z ```
evals	Ejecutar las evals	2026-09-15T04:51:54.1187818Z 
evals	Ejecutar las evals	2026-09-15T04:51:54.1188033Z Invocaciones: ninguna
evals	Ejecutar las evals	2026-09-15T04:51:54.1188143Z 
evals	Ejecutar las evals	2026-09-15T04:51:54.1188241Z Respuesta: vacía
evals	Ejecutar las evals	2026-09-15T04:51:54.1188343Z 
evals	Ejecutar las evals	2026-09-15T04:51:54.1188439Z Motivos de la sesión:
evals	Ejecutar las evals	2026-09-15T04:51:54.1188545Z 
evals	Ejecutar las evals	2026-09-15T04:51:54.1188660Z - la sesión no terminó: código 1
evals	Ejecutar las evals	2026-09-15T04:51:54.1188791Z 
evals	Ejecutar las evals	2026-09-15T04:51:54.1189091Z Salida de error:
evals	Ejecutar las evals	2026-09-15T04:51:54.1189224Z 
evals	Ejecutar las evals	2026-09-15T04:51:54.1189292Z ```text
evals	Ejecutar las evals	2026-09-15T04:51:54.1189575Z  6 | // and may be used to improve Anthropic's products, including training models.
evals	Ejecutar las evals	2026-09-15T04:51:54.1189984Z  7 | // You are responsible for reviewing any code suggestions before use.
evals	Ejecutar las evals	2026-09-15T04:51:54.1190260Z  8 | 
evals	Ejecutar las evals	2026-09-15T04:51:54.1190754Z  9 | // (c) Anthropic PBC. All rights reserved. Use is subject to the Legal Agreements outlined here: https://code.claude.com/docs/en/legal-and-compliance.
evals	Ejecutar las evals	2026-09-15T04:51:54.1191347Z 10 | 
evals	Ejecutar las evals	2026-09-15T04:51:54.1193964Z 11 | import{W,B,he,fN}from"/$bunfs/root/chunk-265sathk.js";import{He,fo,$e}from"/$bunfs/root/chunk-7khdpcz9.js";import{Ude,Yt}from"/$bunfs/root/chunk-4a0graca.js";import{Wi,Dne,Oct,iwe,qct,Lne,Jq}from"/$bunfs/root/chunk-hk7hcqnf.js";import{f}from"/$bunfs/root/chunk-xpzkpaw3.js";import{hx,a}from"/$bunfs/root/chunk-crze9c5c.js";import{Va,l,Mt,hT}from"/$bunfs/root/chunk-a382p5z3.js";import{Mu,qc,K,le,t}from"/$bunfs/root/chunk-0y8dccaw.js";import{lMe,n2n,act,r2n,cMe,Jon,vAr,VSe,PVe,lct,g6,cct,GLt,EAr,uct,IS,zc}from"/$bunfs/root/chunk-ebh8j0r9.js";import{Br}from"/$bunfs/root/chunk-jrz35ppm.js";import{Lue}from"/$bunfs/root/chunk-z0zff54w.js";import{ron}from"/$bunfs/root/chunk-armd3k2m.js";import{gn}from"/$bunfs/root/chunk-aj7ffynp.js";import{aBn}from"/$bunfs/root/chunk-vb37gsz0.js";import{Bo}from"/$bunfs/root/chunk-gk39j4sy.js";import{ie,u,V,R}from"/$bunfs/root/chunk-9kr1d91j.js";import{Y}from"/$bunfs/root/chunk-4112m93q.js";import{homedir as be}from"os";import{dirname as Q,posix as D}from"path";var L=(e)=>JSON.stringif
evals	Ejecutar las evals	2026-09-15T04:51:54.1196531Z 
evals	Ejecutar las evals	2026-09-15T04:51:54.1197176Z error: bubblewrap is required for subprocess env scrubbing and isolation. Install with: sudo apt-get install -y bubblewrap, set sandbox.bwrapPath in managed settings, or set CLAUDE_CODE_SUBPROCESS_ENV_SCRUB=0 to disable (loses subprocess isolation).
evals	Ejecutar las evals	2026-09-15T04:51:54.1197993Z       at nBn (/$bunfs/root/chunk-y58z5vzc.js:11:25510)
evals	Ejecutar las evals	2026-09-15T04:51:54.1198308Z       at A (/$bunfs/root/chunk-xgzajagg.js:11:3535)
evals	Ejecutar las evals	2026-09-15T04:51:54.1198665Z       at async <anonymous> (/$bunfs/root/chunk-mpjnf6jh.js:88:4824)
evals	Ejecutar las evals	2026-09-15T04:51:54.1199010Z 
evals	Ejecutar las evals	2026-09-15T04:51:54.1200010Z ##[error]
evals	Ejecutar las evals	      at nBn (/$bunfs/root/chunk-y58z5vzc.js:11:25510)
evals	Ejecutar las evals	      at A (/$bunfs/root/chunk-xgzajagg.js:11:3535)
evals	Ejecutar las evals	      at async <anonymous> (/$bunfs/root/chunk-mpjnf6jh.js:88:4824)
evals	Ejecutar las evals	2026-09-15T04:51:54.1200797Z 
evals	Ejecutar las evals	2026-09-15T04:51:54.1200877Z Bun v1.4.3 (Linux x64)
evals	Ejecutar las evals	2026-09-15T04:51:54.1201046Z ```
evals	Ejecutar las evals	2026-09-15T04:51:54.1201141Z 
evals	Ejecutar las evals	2026-09-15T04:51:54.1201296Z ## Sesión 03-lrbrl-atribuciones-del-pleno
evals	Ejecutar las evals	2026-09-15T04:51:54.1201449Z 
evals	Ejecutar las evals	2026-09-15T04:51:54.1201556Z Eval: 03-lrbrl-atribuciones-del-pleno.yaml
evals	Ejecutar las evals	2026-09-15T04:51:54.1201722Z 
evals	Ejecutar las evals	2026-09-15T04:51:54.1201792Z Pregunta:
evals	Ejecutar las evals	2026-09-15T04:51:54.1201881Z 
evals	Ejecutar las evals	2026-09-15T04:51:54.1201952Z ```text
evals	Ejecutar las evals	2026-09-15T04:51:54.1202354Z ¿Qué atribuciones tiene el Pleno del ayuntamiento según la Ley reguladora de las Bases del Régimen Local?
evals	Ejecutar las evals	2026-09-15T04:51:54.1202721Z ```
evals	Ejecutar las evals	2026-09-15T04:51:54.1202803Z 
evals	Ejecutar las evals	2026-09-15T04:51:54.1202884Z Invocaciones: ninguna
evals	Ejecutar las evals	2026-09-15T04:51:54.1202992Z 
evals	Ejecutar las evals	2026-09-15T04:51:54.1203096Z Respuesta: vacía
evals	Ejecutar las evals	2026-09-15T04:51:54.1203200Z 
evals	Ejecutar las evals	2026-09-15T04:51:54.1203294Z Motivos de la sesión:
evals	Ejecutar las evals	2026-09-15T04:51:54.1203401Z 
evals	Ejecutar las evals	2026-09-15T04:51:54.1203517Z - la sesión no terminó: código 1
evals	Ejecutar las evals	2026-09-15T04:51:54.1203647Z 
evals	Ejecutar las evals	2026-09-15T04:51:54.1203723Z Salida de error:
evals	Ejecutar las evals	2026-09-15T04:51:54.1203820Z 
evals	Ejecutar las evals	2026-09-15T04:51:54.1203884Z ```text
evals	Ejecutar las evals	2026-09-15T04:51:54.1204156Z  6 | // and may be used to improve Anthropic's products, including training models.
evals	Ejecutar las evals	2026-09-15T04:51:54.1204571Z  7 | // You are responsible for reviewing any code suggestions before use.
evals	Ejecutar las evals	2026-09-15T04:51:54.1204989Z  8 | 
evals	Ejecutar las evals	2026-09-15T04:51:54.1205517Z  9 | // (c) Anthropic PBC. All rights reserved. Use is subject to the Legal Agreements outlined here: https://code.claude.com/docs/en/legal-and-compliance.
evals	Ejecutar las evals	2026-09-15T04:51:54.1205992Z 10 | 
evals	Ejecutar las evals	2026-09-15T04:51:54.1208584Z 11 | import{W,B,he,fN}from"/$bunfs/root/chunk-265sathk.js";import{He,fo,$e}from"/$bunfs/root/chunk-7khdpcz9.js";import{Ude,Yt}from"/$bunfs/root/chunk-4a0graca.js";import{Wi,Dne,Oct,iwe,qct,Lne,Jq}from"/$bunfs/root/chunk-hk7hcqnf.js";import{f}from"/$bunfs/root/chunk-xpzkpaw3.js";import{hx,a}from"/$bunfs/root/chunk-crze9c5c.js";import{Va,l,Mt,hT}from"/$bunfs/root/chunk-a382p5z3.js";import{Mu,qc,K,le,t}from"/$bunfs/root/chunk-0y8dccaw.js";import{lMe,n2n,act,r2n,cMe,Jon,vAr,VSe,PVe,lct,g6,cct,GLt,EAr,uct,IS,zc}from"/$bunfs/root/chunk-ebh8j0r9.js";import{Br}from"/$bunfs/root/chunk-jrz35ppm.js";import{Lue}from"/$bunfs/root/chunk-z0zff54w.js";import{ron}from"/$bunfs/root/chunk-armd3k2m.js";import{gn}from"/$bunfs/root/chunk-aj7ffynp.js";import{aBn}from"/$bunfs/root/chunk-vb37gsz0.js";import{Bo}from"/$bunfs/root/chunk-gk39j4sy.js";import{ie,u,V,R}from"/$bunfs/root/chunk-9kr1d91j.js";import{Y}from"/$bunfs/root/chunk-4112m93q.js";import{homedir as be}from"os";import{dirname as Q,posix as D}from"path";var L=(e)=>JSON.stringif
evals	Ejecutar las evals	2026-09-15T04:51:54.1211457Z 
evals	Ejecutar las evals	2026-09-15T04:51:54.1212096Z error: bubblewrap is required for subprocess env scrubbing and isolation. Install with: sudo apt-get install -y bubblewrap, set sandbox.bwrapPath in managed settings, or set CLAUDE_CODE_SUBPROCESS_ENV_SCRUB=0 to disable (loses subprocess isolation).
evals	Ejecutar las evals	2026-09-15T04:51:54.1212912Z       at nBn (/$bunfs/root/chunk-y58z5vzc.js:11:25510)
evals	Ejecutar las evals	2026-09-15T04:51:54.1213228Z       at A (/$bunfs/root/chunk-xgzajagg.js:11:3535)
evals	Ejecutar las evals	2026-09-15T04:51:54.1213585Z       at async <anonymous> (/$bunfs/root/chunk-mpjnf6jh.js:88:4824)
evals	Ejecutar las evals	2026-09-15T04:51:54.1213789Z 
evals	Ejecutar las evals	2026-09-15T04:51:54.1214815Z ##[error]
evals	Ejecutar las evals	      at nBn (/$bunfs/root/chunk-y58z5vzc.js:11:25510)
evals	Ejecutar las evals	      at A (/$bunfs/root/chunk-xgzajagg.js:11:3535)
evals	Ejecutar las evals	      at async <anonymous> (/$bunfs/root/chunk-mpjnf6jh.js:88:4824)
evals	Ejecutar las evals	2026-09-15T04:51:54.1215589Z 
evals	Ejecutar las evals	2026-09-15T04:51:54.1215671Z Bun v1.4.3 (Linux x64)
evals	Ejecutar las evals	2026-09-15T04:51:54.1215843Z ```
evals	Ejecutar las evals	2026-09-15T04:51:54.1215928Z 
evals	Ejecutar las evals	2026-09-15T04:51:54.1216062Z ## Sesión 04-lgt-prescripcion
evals	Ejecutar las evals	2026-09-15T04:51:54.1216190Z 
evals	Ejecutar las evals	2026-09-15T04:51:54.1216282Z Eval: 04-lgt-prescripcion.yaml
evals	Ejecutar las evals	2026-09-15T04:51:54.1216412Z 
evals	Ejecutar las evals	2026-09-15T04:51:54.1216478Z Pregunta:
evals	Ejecutar las evals	2026-09-15T04:51:54.1216568Z 
evals	Ejecutar las evals	2026-09-15T04:51:54.1216644Z ```text
evals	Ejecutar las evals	2026-09-15T04:51:54.1217101Z ¿En cuántos años prescribe el derecho de la Administración a liquidar una deuda tributaria según la Ley General Tributaria?
evals	Ejecutar las evals	2026-09-15T04:51:54.1217506Z ```
evals	Ejecutar las evals	2026-09-15T04:51:54.1217588Z 
evals	Ejecutar las evals	2026-09-15T04:51:54.1217667Z Invocaciones: ninguna
evals	Ejecutar las evals	2026-09-15T04:51:54.1217774Z 
evals	Ejecutar las evals	2026-09-15T04:51:54.1217868Z Respuesta: vacía
evals	Ejecutar las evals	2026-09-15T04:51:54.1217982Z 
evals	Ejecutar las evals	2026-09-15T04:51:54.1218073Z Motivos de la sesión:
evals	Ejecutar las evals	2026-09-15T04:51:54.1218186Z 
evals	Ejecutar las evals	2026-09-15T04:51:54.1218304Z - la sesión no terminó: código 1
evals	Ejecutar las evals	2026-09-15T04:51:54.1218446Z 
evals	Ejecutar las evals	2026-09-15T04:51:54.1218517Z Salida de error:
evals	Ejecutar las evals	2026-09-15T04:51:54.1218615Z 
evals	Ejecutar las evals	2026-09-15T04:51:54.1218684Z ```text
evals	Ejecutar las evals	2026-09-15T04:51:54.1219135Z  6 | // and may be used to improve Anthropic's products, including training models.
evals	Ejecutar las evals	2026-09-15T04:51:54.1219557Z  7 | // You are responsible for reviewing any code suggestions before use.
evals	Ejecutar las evals	2026-09-15T04:51:54.1219843Z  8 | 
evals	Ejecutar las evals	2026-09-15T04:51:54.1220335Z  9 | // (c) Anthropic PBC. All rights reserved. Use is subject to the Legal Agreements outlined here: https://code.claude.com/docs/en/legal-and-compliance.
evals	Ejecutar las evals	2026-09-15T04:51:54.1220802Z 10 | 
evals	Ejecutar las evals	2026-09-15T04:51:54.1223403Z 11 | import{W,B,he,fN}from"/$bunfs/root/chunk-265sathk.js";import{He,fo,$e}from"/$bunfs/root/chunk-7khdpcz9.js";import{Ude,Yt}from"/$bunfs/root/chunk-4a0graca.js";import{Wi,Dne,Oct,iwe,qct,Lne,Jq}from"/$bunfs/root/chunk-hk7hcqnf.js";import{f}from"/$bunfs/root/chunk-xpzkpaw3.js";import{hx,a}from"/$bunfs/root/chunk-crze9c5c.js";import{Va,l,Mt,hT}from"/$bunfs/root/chunk-a382p5z3.js";import{Mu,qc,K,le,t}from"/$bunfs/root/chunk-0y8dccaw.js";import{lMe,n2n,act,r2n,cMe,Jon,vAr,VSe,PVe,lct,g6,cct,GLt,EAr,uct,IS,zc}from"/$bunfs/root/chunk-ebh8j0r9.js";import{Br}from"/$bunfs/root/chunk-jrz35ppm.js";import{Lue}from"/$bunfs/root/chunk-z0zff54w.js";import{ron}from"/$bunfs/root/chunk-armd3k2m.js";import{gn}from"/$bunfs/root/chunk-aj7ffynp.js";import{aBn}from"/$bunfs/root/chunk-vb37gsz0.js";import{Bo}from"/$bunfs/root/chunk-gk39j4sy.js";import{ie,u,V,R}from"/$bunfs/root/chunk-9kr1d91j.js";import{Y}from"/$bunfs/root/chunk-4112m93q.js";import{homedir as be}from"os";import{dirname as Q,posix as D}from"path";var L=(e)=>JSON.stringif
evals	Ejecutar las evals	2026-09-15T04:51:54.1226113Z 
evals	Ejecutar las evals	2026-09-15T04:51:54.1226774Z error: bubblewrap is required for subprocess env scrubbing and isolation. Install with: sudo apt-get install -y bubblewrap, set sandbox.bwrapPath in managed settings, or set CLAUDE_CODE_SUBPROCESS_ENV_SCRUB=0 to disable (loses subprocess isolation).
evals	Ejecutar las evals	2026-09-15T04:51:54.1227707Z       at nBn (/$bunfs/root/chunk-y58z5vzc.js:11:25510)
evals	Ejecutar las evals	2026-09-15T04:51:54.1228033Z       at A (/$bunfs/root/chunk-xgzajagg.js:11:3535)
evals	Ejecutar las evals	2026-09-15T04:51:54.1228391Z       at async <anonymous> (/$bunfs/root/chunk-mpjnf6jh.js:88:4824)
evals	Ejecutar las evals	2026-09-15T04:51:54.1228586Z 
evals	Ejecutar las evals	2026-09-15T04:51:54.1229746Z ##[error]
evals	Ejecutar las evals	      at nBn (/$bunfs/root/chunk-y58z5vzc.js:11:25510)
evals	Ejecutar las evals	      at A (/$bunfs/root/chunk-xgzajagg.js:11:3535)
evals	Ejecutar las evals	      at async <anonymous> (/$bunfs/root/chunk-mpjnf6jh.js:88:4824)
evals	Ejecutar las evals	2026-09-15T04:51:54.1230538Z 
evals	Ejecutar las evals	2026-09-15T04:51:54.1230626Z Bun v1.4.3 (Linux x64)
evals	Ejecutar las evals	2026-09-15T04:51:54.1230795Z ```
evals	Ejecutar las evals	2026-09-15T04:51:54.1230877Z 
evals	Ejecutar las evals	2026-09-15T04:51:54.1231031Z ## Sesión 05-trlrhl-impuestos-municipales
evals	Ejecutar las evals	2026-09-15T04:51:54.1231189Z 
evals	Ejecutar las evals	2026-09-15T04:51:54.1231292Z Eval: 05-trlrhl-impuestos-municipales.yaml
evals	Ejecutar las evals	2026-09-15T04:51:54.1231462Z 
evals	Ejecutar las evals	2026-09-15T04:51:54.1231535Z Pregunta:
evals	Ejecutar las evals	2026-09-15T04:51:54.1231621Z 
evals	Ejecutar las evals	2026-09-15T04:51:54.1231697Z ```text
evals	Ejecutar las evals	2026-09-15T04:51:54.1232139Z ¿Qué impuestos pueden exigir los ayuntamientos según el texto refundido de la Ley reguladora de las Haciendas Locales?
evals	Ejecutar las evals	2026-09-15T04:51:54.1232552Z ```
evals	Ejecutar las evals	2026-09-15T04:51:54.1232630Z 
evals	Ejecutar las evals	2026-09-15T04:51:54.1232711Z Invocaciones: ninguna
evals	Ejecutar las evals	2026-09-15T04:51:54.1232821Z 
evals	Ejecutar las evals	2026-09-15T04:51:54.1232912Z Respuesta: vacía
evals	Ejecutar las evals	2026-09-15T04:51:54.1233016Z 
evals	Ejecutar las evals	2026-09-15T04:51:54.1233114Z Motivos de la sesión:
evals	Ejecutar las evals	2026-09-15T04:51:54.1233213Z 
evals	Ejecutar las evals	2026-09-15T04:51:54.1233328Z - la sesión no terminó: código 1
evals	Ejecutar las evals	2026-09-15T04:51:54.1233462Z 
evals	Ejecutar las evals	2026-09-15T04:51:54.1233532Z Salida de error:
evals	Ejecutar las evals	2026-09-15T04:51:54.1233627Z 
evals	Ejecutar las evals	2026-09-15T04:51:54.1233695Z ```text
evals	Ejecutar las evals	2026-09-15T04:51:54.1233964Z  6 | // and may be used to improve Anthropic's products, including training models.
evals	Ejecutar las evals	2026-09-15T04:51:54.1234362Z  7 | // You are responsible for reviewing any code suggestions before use.
evals	Ejecutar las evals	2026-09-15T04:51:54.1234641Z  8 | 
evals	Ejecutar las evals	2026-09-15T04:51:54.1235123Z  9 | // (c) Anthropic PBC. All rights reserved. Use is subject to the Legal Agreements outlined here: https://code.claude.com/docs/en/legal-and-compliance.
evals	Ejecutar las evals	2026-09-15T04:51:54.1235604Z 10 | 
evals	Ejecutar las evals	2026-09-15T04:51:54.1238199Z 11 | import{W,B,he,fN}from"/$bunfs/root/chunk-265sathk.js";import{He,fo,$e}from"/$bunfs/root/chunk-7khdpcz9.js";import{Ude,Yt}from"/$bunfs/root/chunk-4a0graca.js";import{Wi,Dne,Oct,iwe,qct,Lne,Jq}from"/$bunfs/root/chunk-hk7hcqnf.js";import{f}from"/$bunfs/root/chunk-xpzkpaw3.js";import{hx,a}from"/$bunfs/root/chunk-crze9c5c.js";import{Va,l,Mt,hT}from"/$bunfs/root/chunk-a382p5z3.js";import{Mu,qc,K,le,t}from"/$bunfs/root/chunk-0y8dccaw.js";import{lMe,n2n,act,r2n,cMe,Jon,vAr,VSe,PVe,lct,g6,cct,GLt,EAr,uct,IS,zc}from"/$bunfs/root/chunk-ebh8j0r9.js";import{Br}from"/$bunfs/root/chunk-jrz35ppm.js";import{Lue}from"/$bunfs/root/chunk-z0zff54w.js";import{ron}from"/$bunfs/root/chunk-armd3k2m.js";import{gn}from"/$bunfs/root/chunk-aj7ffynp.js";import{aBn}from"/$bunfs/root/chunk-vb37gsz0.js";import{Bo}from"/$bunfs/root/chunk-gk39j4sy.js";import{ie,u,V,R}from"/$bunfs/root/chunk-9kr1d91j.js";import{Y}from"/$bunfs/root/chunk-4112m93q.js";import{homedir as be}from"os";import{dirname as Q,posix as D}from"path";var L=(e)=>JSON.stringif
evals	Ejecutar las evals	2026-09-15T04:51:54.1240975Z 
evals	Ejecutar las evals	2026-09-15T04:51:54.1241623Z error: bubblewrap is required for subprocess env scrubbing and isolation. Install with: sudo apt-get install -y bubblewrap, set sandbox.bwrapPath in managed settings, or set CLAUDE_CODE_SUBPROCESS_ENV_SCRUB=0 to disable (loses subprocess isolation).
evals	Ejecutar las evals	2026-09-15T04:51:54.1242425Z       at nBn (/$bunfs/root/chunk-y58z5vzc.js:11:25510)
evals	Ejecutar las evals	2026-09-15T04:51:54.1242742Z       at A (/$bunfs/root/chunk-xgzajagg.js:11:3535)
evals	Ejecutar las evals	2026-09-15T04:51:54.1243100Z       at async <anonymous> (/$bunfs/root/chunk-mpjnf6jh.js:88:4824)
evals	Ejecutar las evals	2026-09-15T04:51:54.1243295Z 
evals	Ejecutar las evals	2026-09-15T04:51:54.1244275Z ##[error]
evals	Ejecutar las evals	      at nBn (/$bunfs/root/chunk-y58z5vzc.js:11:25510)
evals	Ejecutar las evals	      at A (/$bunfs/root/chunk-xgzajagg.js:11:3535)
evals	Ejecutar las evals	      at async <anonymous> (/$bunfs/root/chunk-mpjnf6jh.js:88:4824)
evals	Ejecutar las evals	2026-09-15T04:51:54.1245045Z 
evals	Ejecutar las evals	2026-09-15T04:51:54.1245227Z Bun v1.4.3 (Linux x64)
evals	Ejecutar las evals	2026-09-15T04:51:54.1245398Z ```
evals	Ejecutar las evals	2026-09-15T04:51:54.1245483Z 
evals	Ejecutar las evals	2026-09-15T04:51:54.1245628Z ## Sesión 06-irpf-rendimientos-del-trabajo
evals	Ejecutar las evals	2026-09-15T04:51:54.1245793Z 
evals	Ejecutar las evals	2026-09-15T04:51:54.1245900Z Eval: 06-irpf-rendimientos-del-trabajo.yaml
evals	Ejecutar las evals	2026-09-15T04:51:54.1246056Z 
evals	Ejecutar las evals	2026-09-15T04:51:54.1246129Z Pregunta:
evals	Ejecutar las evals	2026-09-15T04:51:54.1246219Z 
evals	Ejecutar las evals	2026-09-15T04:51:54.1246284Z ```text
evals	Ejecutar las evals	2026-09-15T04:51:54.1246621Z ¿Qué rendimientos se consideran rendimientos íntegros del trabajo en la ley del IRPF?
evals	Ejecutar las evals	2026-09-15T04:51:54.1246944Z ```
evals	Ejecutar las evals	2026-09-15T04:51:54.1247028Z 
evals	Ejecutar las evals	2026-09-15T04:51:54.1247104Z Invocaciones: ninguna
evals	Ejecutar las evals	2026-09-15T04:51:54.1247215Z 
evals	Ejecutar las evals	2026-09-15T04:51:54.1247310Z Respuesta: vacía
evals	Ejecutar las evals	2026-09-15T04:51:54.1247409Z 
evals	Ejecutar las evals	2026-09-15T04:51:54.1247506Z Motivos de la sesión:
evals	Ejecutar las evals	2026-09-15T04:51:54.1247610Z 
evals	Ejecutar las evals	2026-09-15T04:51:54.1247722Z - la sesión no terminó: código 1
evals	Ejecutar las evals	2026-09-15T04:51:54.1247853Z 
evals	Ejecutar las evals	2026-09-15T04:51:54.1247927Z Salida de error:
evals	Ejecutar las evals	2026-09-15T04:51:54.1248019Z 
evals	Ejecutar las evals	2026-09-15T04:51:54.1248087Z ```text
evals	Ejecutar las evals	2026-09-15T04:51:54.1248372Z  6 | // and may be used to improve Anthropic's products, including training models.
evals	Ejecutar las evals	2026-09-15T04:51:54.1248779Z  7 | // You are responsible for reviewing any code suggestions before use.
evals	Ejecutar las evals	2026-09-15T04:51:54.1249151Z  8 | 
evals	Ejecutar las evals	2026-09-15T04:51:54.1249650Z  9 | // (c) Anthropic PBC. All rights reserved. Use is subject to the Legal Agreements outlined here: https://code.claude.com/docs/en/legal-and-compliance.
evals	Ejecutar las evals	2026-09-15T04:51:54.1250121Z 10 | 
evals	Ejecutar las evals	2026-09-15T04:51:54.1252739Z 11 | import{W,B,he,fN}from"/$bunfs/root/chunk-265sathk.js";import{He,fo,$e}from"/$bunfs/root/chunk-7khdpcz9.js";import{Ude,Yt}from"/$bunfs/root/chunk-4a0graca.js";import{Wi,Dne,Oct,iwe,qct,Lne,Jq}from"/$bunfs/root/chunk-hk7hcqnf.js";import{f}from"/$bunfs/root/chunk-xpzkpaw3.js";import{hx,a}from"/$bunfs/root/chunk-crze9c5c.js";import{Va,l,Mt,hT}from"/$bunfs/root/chunk-a382p5z3.js";import{Mu,qc,K,le,t}from"/$bunfs/root/chunk-0y8dccaw.js";import{lMe,n2n,act,r2n,cMe,Jon,vAr,VSe,PVe,lct,g6,cct,GLt,EAr,uct,IS,zc}from"/$bunfs/root/chunk-ebh8j0r9.js";import{Br}from"/$bunfs/root/chunk-jrz35ppm.js";import{Lue}from"/$bunfs/root/chunk-z0zff54w.js";import{ron}from"/$bunfs/root/chunk-armd3k2m.js";import{gn}from"/$bunfs/root/chunk-aj7ffynp.js";import{aBn}from"/$bunfs/root/chunk-vb37gsz0.js";import{Bo}from"/$bunfs/root/chunk-gk39j4sy.js";import{ie,u,V,R}from"/$bunfs/root/chunk-9kr1d91j.js";import{Y}from"/$bunfs/root/chunk-4112m93q.js";import{homedir as be}from"os";import{dirname as Q,posix as D}from"path";var L=(e)=>JSON.stringif
evals	Ejecutar las evals	2026-09-15T04:51:54.1255266Z 
evals	Ejecutar las evals	2026-09-15T04:51:54.1255910Z error: bubblewrap is required for subprocess env scrubbing and isolation. Install with: sudo apt-get install -y bubblewrap, set sandbox.bwrapPath in managed settings, or set CLAUDE_CODE_SUBPROCESS_ENV_SCRUB=0 to disable (loses subprocess isolation).
evals	Ejecutar las evals	2026-09-15T04:51:54.1256722Z       at nBn (/$bunfs/root/chunk-y58z5vzc.js:11:25510)
evals	Ejecutar las evals	2026-09-15T04:51:54.1257035Z       at A (/$bunfs/root/chunk-xgzajagg.js:11:3535)
evals	Ejecutar las evals	2026-09-15T04:51:54.1257391Z       at async <anonymous> (/$bunfs/root/chunk-mpjnf6jh.js:88:4824)
evals	Ejecutar las evals	2026-09-15T04:51:54.1257595Z 
evals	Ejecutar las evals	2026-09-15T04:51:54.1258559Z ##[error]
evals	Ejecutar las evals	      at nBn (/$bunfs/root/chunk-y58z5vzc.js:11:25510)
evals	Ejecutar las evals	      at A (/$bunfs/root/chunk-xgzajagg.js:11:3535)
evals	Ejecutar las evals	      at async <anonymous> (/$bunfs/root/chunk-mpjnf6jh.js:88:4824)
evals	Ejecutar las evals	2026-09-15T04:51:54.1259717Z 
evals	Ejecutar las evals	2026-09-15T04:51:54.1259798Z Bun v1.4.3 (Linux x64)
evals	Ejecutar las evals	2026-09-15T04:51:54.1259967Z ```
evals	Ejecutar las evals	2026-09-15T04:51:54.1260051Z 
evals	Ejecutar las evals	2026-09-15T04:51:54.1260211Z ## Sesión 07-lrjsp-principio-de-legalidad
evals	Ejecutar las evals	2026-09-15T04:51:54.1260365Z 
evals	Ejecutar las evals	2026-09-15T04:51:54.1260474Z Eval: 07-lrjsp-principio-de-legalidad.yaml
evals	Ejecutar las evals	2026-09-15T04:51:54.1260633Z 
evals	Ejecutar las evals	2026-09-15T04:51:54.1260700Z Pregunta:
evals	Ejecutar las evals	2026-09-15T04:51:54.1260789Z 
evals	Ejecutar las evals	2026-09-15T04:51:54.1260859Z ```text
evals	Ejecutar las evals	2026-09-15T04:51:54.1261184Z ¿Qué dice la Ley 40/2015 sobre el principio de legalidad en la potestad sancionadora?
evals	Ejecutar las evals	2026-09-15T04:51:54.1261490Z ```
evals	Ejecutar las evals	2026-09-15T04:51:54.1261574Z 
evals	Ejecutar las evals	2026-09-15T04:51:54.1261655Z Invocaciones: ninguna
evals	Ejecutar las evals	2026-09-15T04:51:54.1261760Z 
evals	Ejecutar las evals	2026-09-15T04:51:54.1261973Z Respuesta: vacía
evals	Ejecutar las evals	2026-09-15T04:51:54.1262080Z 
evals	Ejecutar las evals	2026-09-15T04:51:54.1262173Z Motivos de la sesión:
evals	Ejecutar las evals	2026-09-15T04:51:54.1262285Z 
evals	Ejecutar las evals	2026-09-15T04:51:54.1262408Z - la sesión no terminó: código 1
evals	Ejecutar las evals	2026-09-15T04:51:54.1262537Z 
evals	Ejecutar las evals	2026-09-15T04:51:54.1262614Z Salida de error:
evals	Ejecutar las evals	2026-09-15T04:51:54.1262712Z 
evals	Ejecutar las evals	2026-09-15T04:51:54.1262778Z ```text
evals	Ejecutar las evals	2026-09-15T04:51:54.1263055Z  6 | // and may be used to improve Anthropic's products, including training models.
evals	Ejecutar las evals	2026-09-15T04:51:54.1263459Z  7 | // You are responsible for reviewing any code suggestions before use.
evals	Ejecutar las evals	2026-09-15T04:51:54.1263731Z  8 | 
evals	Ejecutar las evals	2026-09-15T04:51:54.1264217Z  9 | // (c) Anthropic PBC. All rights reserved. Use is subject to the Legal Agreements outlined here: https://code.claude.com/docs/en/legal-and-compliance.
evals	Ejecutar las evals	2026-09-15T04:51:54.1264687Z 10 | 
evals	Ejecutar las evals	2026-09-15T04:51:54.1267278Z 11 | import{W,B,he,fN}from"/$bunfs/root/chunk-265sathk.js";import{He,fo,$e}from"/$bunfs/root/chunk-7khdpcz9.js";import{Ude,Yt}from"/$bunfs/root/chunk-4a0graca.js";import{Wi,Dne,Oct,iwe,qct,Lne,Jq}from"/$bunfs/root/chunk-hk7hcqnf.js";import{f}from"/$bunfs/root/chunk-xpzkpaw3.js";import{hx,a}from"/$bunfs/root/chunk-crze9c5c.js";import{Va,l,Mt,hT}from"/$bunfs/root/chunk-a382p5z3.js";import{Mu,qc,K,le,t}from"/$bunfs/root/chunk-0y8dccaw.js";import{lMe,n2n,act,r2n,cMe,Jon,vAr,VSe,PVe,lct,g6,cct,GLt,EAr,uct,IS,zc}from"/$bunfs/root/chunk-ebh8j0r9.js";import{Br}from"/$bunfs/root/chunk-jrz35ppm.js";import{Lue}from"/$bunfs/root/chunk-z0zff54w.js";import{ron}from"/$bunfs/root/chunk-armd3k2m.js";import{gn}from"/$bunfs/root/chunk-aj7ffynp.js";import{aBn}from"/$bunfs/root/chunk-vb37gsz0.js";import{Bo}from"/$bunfs/root/chunk-gk39j4sy.js";import{ie,u,V,R}from"/$bunfs/root/chunk-9kr1d91j.js";import{Y}from"/$bunfs/root/chunk-4112m93q.js";import{homedir as be}from"os";import{dirname as Q,posix as D}from"path";var L=(e)=>JSON.stringif
evals	Ejecutar las evals	2026-09-15T04:51:54.1269899Z 
evals	Ejecutar las evals	2026-09-15T04:51:54.1270539Z error: bubblewrap is required for subprocess env scrubbing and isolation. Install with: sudo apt-get install -y bubblewrap, set sandbox.bwrapPath in managed settings, or set CLAUDE_CODE_SUBPROCESS_ENV_SCRUB=0 to disable (loses subprocess isolation).
evals	Ejecutar las evals	2026-09-15T04:51:54.1271375Z       at nBn (/$bunfs/root/chunk-y58z5vzc.js:11:25510)
evals	Ejecutar las evals	2026-09-15T04:51:54.1271690Z       at A (/$bunfs/root/chunk-xgzajagg.js:11:3535)
evals	Ejecutar las evals	2026-09-15T04:51:54.1272045Z       at async <anonymous> (/$bunfs/root/chunk-mpjnf6jh.js:88:4824)
evals	Ejecutar las evals	2026-09-15T04:51:54.1272239Z 
evals	Ejecutar las evals	2026-09-15T04:51:54.1273237Z ##[error]
evals	Ejecutar las evals	      at nBn (/$bunfs/root/chunk-y58z5vzc.js:11:25510)
evals	Ejecutar las evals	      at A (/$bunfs/root/chunk-xgzajagg.js:11:3535)
evals	Ejecutar las evals	      at async <anonymous> (/$bunfs/root/chunk-mpjnf6jh.js:88:4824)
evals	Ejecutar las evals	2026-09-15T04:51:54.1274008Z 
evals	Ejecutar las evals	2026-09-15T04:51:54.1274085Z Bun v1.4.3 (Linux x64)
evals	Ejecutar las evals	2026-09-15T04:51:54.1274261Z ```
evals	Ejecutar las evals	2026-09-15T04:51:54.1274341Z 
evals	Ejecutar las evals	2026-09-15T04:51:54.1274481Z ## Sesión 08-ltaibg-plazo-de-resolucion
evals	Ejecutar las evals	2026-09-15T04:51:54.1274633Z 
evals	Ejecutar las evals	2026-09-15T04:51:54.1274731Z Eval: 08-ltaibg-plazo-de-resolucion.yaml
evals	Ejecutar las evals	2026-09-15T04:51:54.1274887Z 
evals	Ejecutar las evals	2026-09-15T04:51:54.1274960Z Pregunta:
evals	Ejecutar las evals	2026-09-15T04:51:54.1275053Z 
evals	Ejecutar las evals	2026-09-15T04:51:54.1275121Z ```text
evals	Ejecutar las evals	2026-09-15T04:51:54.1275495Z ¿En qué plazo hay que resolver una solicitud de acceso a la información pública según la Ley 19/2013?
evals	Ejecutar las evals	2026-09-15T04:51:54.1275963Z ```
evals	Ejecutar las evals	2026-09-15T04:51:54.1276043Z 
evals	Ejecutar las evals	2026-09-15T04:51:54.1276122Z Invocaciones: ninguna
evals	Ejecutar las evals	2026-09-15T04:51:54.1276234Z 
evals	Ejecutar las evals	2026-09-15T04:51:54.1276330Z Respuesta: vacía
evals	Ejecutar las evals	2026-09-15T04:51:54.1276435Z 
evals	Ejecutar las evals	2026-09-15T04:51:54.1276531Z Motivos de la sesión:
evals	Ejecutar las evals	2026-09-15T04:51:54.1276635Z 
evals	Ejecutar las evals	2026-09-15T04:51:54.1276750Z - la sesión no terminó: código 1
evals	Ejecutar las evals	2026-09-15T04:51:54.1276885Z 
evals	Ejecutar las evals	2026-09-15T04:51:54.1276956Z Salida de error:
evals	Ejecutar las evals	2026-09-15T04:51:54.1277053Z 
evals	Ejecutar las evals	2026-09-15T04:51:54.1277121Z ```text
evals	Ejecutar las evals	2026-09-15T04:51:54.1277390Z  6 | // and may be used to improve Anthropic's products, including training models.
evals	Ejecutar las evals	2026-09-15T04:51:54.1277787Z  7 | // You are responsible for reviewing any code suggestions before use.
evals	Ejecutar las evals	2026-09-15T04:51:54.1278063Z  8 | 
evals	Ejecutar las evals	2026-09-15T04:51:54.1278545Z  9 | // (c) Anthropic PBC. All rights reserved. Use is subject to the Legal Agreements outlined here: https://code.claude.com/docs/en/legal-and-compliance.
evals	Ejecutar las evals	2026-09-15T04:51:54.1279236Z 10 | 
evals	Ejecutar las evals	2026-09-15T04:51:54.1281852Z 11 | import{W,B,he,fN}from"/$bunfs/root/chunk-265sathk.js";import{He,fo,$e}from"/$bunfs/root/chunk-7khdpcz9.js";import{Ude,Yt}from"/$bunfs/root/chunk-4a0graca.js";import{Wi,Dne,Oct,iwe,qct,Lne,Jq}from"/$bunfs/root/chunk-hk7hcqnf.js";import{f}from"/$bunfs/root/chunk-xpzkpaw3.js";import{hx,a}from"/$bunfs/root/chunk-crze9c5c.js";import{Va,l,Mt,hT}from"/$bunfs/root/chunk-a382p5z3.js";import{Mu,qc,K,le,t}from"/$bunfs/root/chunk-0y8dccaw.js";import{lMe,n2n,act,r2n,cMe,Jon,vAr,VSe,PVe,lct,g6,cct,GLt,EAr,uct,IS,zc}from"/$bunfs/root/chunk-ebh8j0r9.js";import{Br}from"/$bunfs/root/chunk-jrz35ppm.js";import{Lue}from"/$bunfs/root/chunk-z0zff54w.js";import{ron}from"/$bunfs/root/chunk-armd3k2m.js";import{gn}from"/$bunfs/root/chunk-aj7ffynp.js";import{aBn}from"/$bunfs/root/chunk-vb37gsz0.js";import{Bo}from"/$bunfs/root/chunk-gk39j4sy.js";import{ie,u,V,R}from"/$bunfs/root/chunk-9kr1d91j.js";import{Y}from"/$bunfs/root/chunk-4112m93q.js";import{homedir as be}from"os";import{dirname as Q,posix as D}from"path";var L=(e)=>JSON.stringif
evals	Ejecutar las evals	2026-09-15T04:51:54.1284414Z 
evals	Ejecutar las evals	2026-09-15T04:51:54.1285054Z error: bubblewrap is required for subprocess env scrubbing and isolation. Install with: sudo apt-get install -y bubblewrap, set sandbox.bwrapPath in managed settings, or set CLAUDE_CODE_SUBPROCESS_ENV_SCRUB=0 to disable (loses subprocess isolation).
evals	Ejecutar las evals	2026-09-15T04:51:54.1285861Z       at nBn (/$bunfs/root/chunk-y58z5vzc.js:11:25510)
evals	Ejecutar las evals	2026-09-15T04:51:54.1286174Z       at A (/$bunfs/root/chunk-xgzajagg.js:11:3535)
evals	Ejecutar las evals	2026-09-15T04:51:54.1286533Z       at async <anonymous> (/$bunfs/root/chunk-mpjnf6jh.js:88:4824)
evals	Ejecutar las evals	2026-09-15T04:51:54.1286726Z 
evals	Ejecutar las evals	2026-09-15T04:51:54.1287699Z ##[error]
evals	Ejecutar las evals	      at nBn (/$bunfs/root/chunk-y58z5vzc.js:11:25510)
evals	Ejecutar las evals	      at A (/$bunfs/root/chunk-xgzajagg.js:11:3535)
evals	Ejecutar las evals	      at async <anonymous> (/$bunfs/root/chunk-mpjnf6jh.js:88:4824)
evals	Ejecutar las evals	2026-09-15T04:51:54.1288500Z 
evals	Ejecutar las evals	2026-09-15T04:51:54.1288581Z Bun v1.4.3 (Linux x64)
evals	Ejecutar las evals	2026-09-15T04:51:54.1288755Z ```
evals	Ejecutar las evals	2026-09-15T04:51:54.1288925Z 
evals	Ejecutar las evals	2026-09-15T04:51:54.1289061Z ## Sesión 09-constitucion-articulo-140
evals	Ejecutar las evals	2026-09-15T04:51:54.1289219Z 
evals	Ejecutar las evals	2026-09-15T04:51:54.1289320Z Eval: 09-constitucion-articulo-140.yaml
evals	Ejecutar las evals	2026-09-15T04:51:54.1289468Z 
evals	Ejecutar las evals	2026-09-15T04:51:54.1289539Z Pregunta:
evals	Ejecutar las evals	2026-09-15T04:51:54.1289629Z 
evals	Ejecutar las evals	2026-09-15T04:51:54.1289692Z ```text
evals	Ejecutar las evals	2026-09-15T04:51:54.1289918Z ¿Qué dice el artículo 140 de la Constitución?
evals	Ejecutar las evals	2026-09-15T04:51:54.1290144Z ```
evals	Ejecutar las evals	2026-09-15T04:51:54.1290228Z 
evals	Ejecutar las evals	2026-09-15T04:51:54.1290304Z Invocaciones: ninguna
evals	Ejecutar las evals	2026-09-15T04:51:54.1290419Z 
evals	Ejecutar las evals	2026-09-15T04:51:54.1290520Z Respuesta: vacía
evals	Ejecutar las evals	2026-09-15T04:51:54.1290619Z 
evals	Ejecutar las evals	2026-09-15T04:51:54.1290715Z Motivos de la sesión:
evals	Ejecutar las evals	2026-09-15T04:51:54.1290819Z 
evals	Ejecutar las evals	2026-09-15T04:51:54.1290929Z - la sesión no terminó: código 1
evals	Ejecutar las evals	2026-09-15T04:51:54.1291062Z 
evals	Ejecutar las evals	2026-09-15T04:51:54.1291136Z Salida de error:
evals	Ejecutar las evals	2026-09-15T04:51:54.1291227Z 
evals	Ejecutar las evals	2026-09-15T04:51:54.1291296Z ```text
evals	Ejecutar las evals	2026-09-15T04:51:54.1291576Z  6 | // and may be used to improve Anthropic's products, including training models.
evals	Ejecutar las evals	2026-09-15T04:51:54.1291986Z  7 | // You are responsible for reviewing any code suggestions before use.
evals	Ejecutar las evals	2026-09-15T04:51:54.1292396Z  8 | 
evals	Ejecutar las evals	2026-09-15T04:51:54.1292883Z  9 | // (c) Anthropic PBC. All rights reserved. Use is subject to the Legal Agreements outlined here: https://code.claude.com/docs/en/legal-and-compliance.
evals	Ejecutar las evals	2026-09-15T04:51:54.1293355Z 10 | 
evals	Ejecutar las evals	2026-09-15T04:51:54.1295956Z 11 | import{W,B,he,fN}from"/$bunfs/root/chunk-265sathk.js";import{He,fo,$e}from"/$bunfs/root/chunk-7khdpcz9.js";import{Ude,Yt}from"/$bunfs/root/chunk-4a0graca.js";import{Wi,Dne,Oct,iwe,qct,Lne,Jq}from"/$bunfs/root/chunk-hk7hcqnf.js";import{f}from"/$bunfs/root/chunk-xpzkpaw3.js";import{hx,a}from"/$bunfs/root/chunk-crze9c5c.js";import{Va,l,Mt,hT}from"/$bunfs/root/chunk-a382p5z3.js";import{Mu,qc,K,le,t}from"/$bunfs/root/chunk-0y8dccaw.js";import{lMe,n2n,act,r2n,cMe,Jon,vAr,VSe,PVe,lct,g6,cct,GLt,EAr,uct,IS,zc}from"/$bunfs/root/chunk-ebh8j0r9.js";import{Br}from"/$bunfs/root/chunk-jrz35ppm.js";import{Lue}from"/$bunfs/root/chunk-z0zff54w.js";import{ron}from"/$bunfs/root/chunk-armd3k2m.js";import{gn}from"/$bunfs/root/chunk-aj7ffynp.js";import{aBn}from"/$bunfs/root/chunk-vb37gsz0.js";import{Bo}from"/$bunfs/root/chunk-gk39j4sy.js";import{ie,u,V,R}from"/$bunfs/root/chunk-9kr1d91j.js";import{Y}from"/$bunfs/root/chunk-4112m93q.js";import{homedir as be}from"os";import{dirname as Q,posix as D}from"path";var L=(e)=>JSON.stringif
evals	Ejecutar las evals	2026-09-15T04:51:54.1298624Z 
evals	Ejecutar las evals	2026-09-15T04:51:54.1299357Z error: bubblewrap is required for subprocess env scrubbing and isolation. Install with: sudo apt-get install -y bubblewrap, set sandbox.bwrapPath in managed settings, or set CLAUDE_CODE_SUBPROCESS_ENV_SCRUB=0 to disable (loses subprocess isolation).
evals	Ejecutar las evals	2026-09-15T04:51:54.1300209Z       at nBn (/$bunfs/root/chunk-y58z5vzc.js:11:25510)
evals	Ejecutar las evals	2026-09-15T04:51:54.1300531Z       at A (/$bunfs/root/chunk-xgzajagg.js:11:3535)
evals	Ejecutar las evals	2026-09-15T04:51:54.1300885Z       at async <anonymous> (/$bunfs/root/chunk-mpjnf6jh.js:88:4824)
evals	Ejecutar las evals	2026-09-15T04:51:54.1301088Z 
evals	Ejecutar las evals	2026-09-15T04:51:54.1302063Z ##[error]
evals	Ejecutar las evals	      at nBn (/$bunfs/root/chunk-y58z5vzc.js:11:25510)
evals	Ejecutar las evals	      at A (/$bunfs/root/chunk-xgzajagg.js:11:3535)
evals	Ejecutar las evals	      at async <anonymous> (/$bunfs/root/chunk-mpjnf6jh.js:88:4824)
evals	Ejecutar las evals	2026-09-15T04:51:54.1302916Z 
evals	Ejecutar las evals	2026-09-15T04:51:54.1302991Z Bun v1.4.3 (Linux x64)
evals	Ejecutar las evals	2026-09-15T04:51:54.1303160Z ```
evals	Ejecutar las evals	2026-09-15T04:51:54.1303246Z 
evals	Ejecutar las evals	2026-09-15T04:51:54.1303370Z ## Sesión 10-et-vacaciones
evals	Ejecutar las evals	2026-09-15T04:51:54.1303492Z 
evals	Ejecutar las evals	2026-09-15T04:51:54.1303575Z Eval: 10-et-vacaciones.yaml
evals	Ejecutar las evals	2026-09-15T04:51:54.1303697Z 
evals	Ejecutar las evals	2026-09-15T04:51:54.1303767Z Pregunta:
evals	Ejecutar las evals	2026-09-15T04:51:54.1303856Z 
evals	Ejecutar las evals	2026-09-15T04:51:54.1303925Z ```text
evals	Ejecutar las evals	2026-09-15T04:51:54.1304236Z ¿Cuántos días de vacaciones anuales reconoce el Estatuto de los Trabajadores?
evals	Ejecutar las evals	2026-09-15T04:51:54.1304531Z ```
evals	Ejecutar las evals	2026-09-15T04:51:54.1304612Z 
evals	Ejecutar las evals	2026-09-15T04:51:54.1304692Z Invocaciones: ninguna
evals	Ejecutar las evals	2026-09-15T04:51:54.1304798Z 
evals	Ejecutar las evals	2026-09-15T04:51:54.1304896Z Respuesta: vacía
evals	Ejecutar las evals	2026-09-15T04:51:54.1305006Z 
evals	Ejecutar las evals	2026-09-15T04:51:54.1305098Z Motivos de la sesión:
evals	Ejecutar las evals	2026-09-15T04:51:54.1305203Z 
evals	Ejecutar las evals	2026-09-15T04:51:54.1305320Z - la sesión no terminó: código 1
evals	Ejecutar las evals	2026-09-15T04:51:54.1305502Z 
evals	Ejecutar las evals	2026-09-15T04:51:54.1305578Z Salida de error:
evals	Ejecutar las evals	2026-09-15T04:51:54.1305673Z 
evals	Ejecutar las evals	2026-09-15T04:51:54.1305737Z ```text
evals	Ejecutar las evals	2026-09-15T04:51:54.1306010Z  6 | // and may be used to improve Anthropic's products, including training models.
evals	Ejecutar las evals	2026-09-15T04:51:54.1306414Z  7 | // You are responsible for reviewing any code suggestions before use.
evals	Ejecutar las evals	2026-09-15T04:51:54.1306689Z  8 | 
evals	Ejecutar las evals	2026-09-15T04:51:54.1307173Z  9 | // (c) Anthropic PBC. All rights reserved. Use is subject to the Legal Agreements outlined here: https://code.claude.com/docs/en/legal-and-compliance.
evals	Ejecutar las evals	2026-09-15T04:51:54.1307637Z 10 | 
evals	Ejecutar las evals	2026-09-15T04:51:54.1310316Z 11 | import{W,B,he,fN}from"/$bunfs/root/chunk-265sathk.js";import{He,fo,$e}from"/$bunfs/root/chunk-7khdpcz9.js";import{Ude,Yt}from"/$bunfs/root/chunk-4a0graca.js";import{Wi,Dne,Oct,iwe,qct,Lne,Jq}from"/$bunfs/root/chunk-hk7hcqnf.js";import{f}from"/$bunfs/root/chunk-xpzkpaw3.js";import{hx,a}from"/$bunfs/root/chunk-crze9c5c.js";import{Va,l,Mt,hT}from"/$bunfs/root/chunk-a382p5z3.js";import{Mu,qc,K,le,t}from"/$bunfs/root/chunk-0y8dccaw.js";import{lMe,n2n,act,r2n,cMe,Jon,vAr,VSe,PVe,lct,g6,cct,GLt,EAr,uct,IS,zc}from"/$bunfs/root/chunk-ebh8j0r9.js";import{Br}from"/$bunfs/root/chunk-jrz35ppm.js";import{Lue}from"/$bunfs/root/chunk-z0zff54w.js";import{ron}from"/$bunfs/root/chunk-armd3k2m.js";import{gn}from"/$bunfs/root/chunk-aj7ffynp.js";import{aBn}from"/$bunfs/root/chunk-vb37gsz0.js";import{Bo}from"/$bunfs/root/chunk-gk39j4sy.js";import{ie,u,V,R}from"/$bunfs/root/chunk-9kr1d91j.js";import{Y}from"/$bunfs/root/chunk-4112m93q.js";import{homedir as be}from"os";import{dirname as Q,posix as D}from"path";var L=(e)=>JSON.stringif
evals	Ejecutar las evals	2026-09-15T04:51:54.1313023Z 
evals	Ejecutar las evals	2026-09-15T04:51:54.1313669Z error: bubblewrap is required for subprocess env scrubbing and isolation. Install with: sudo apt-get install -y bubblewrap, set sandbox.bwrapPath in managed settings, or set CLAUDE_CODE_SUBPROCESS_ENV_SCRUB=0 to disable (loses subprocess isolation).
evals	Ejecutar las evals	2026-09-15T04:51:54.1314597Z       at nBn (/$bunfs/root/chunk-y58z5vzc.js:11:25510)
evals	Ejecutar las evals	2026-09-15T04:51:54.1314918Z       at A (/$bunfs/root/chunk-xgzajagg.js:11:3535)
evals	Ejecutar las evals	2026-09-15T04:51:54.1315282Z       at async <anonymous> (/$bunfs/root/chunk-mpjnf6jh.js:88:4824)
evals	Ejecutar las evals	2026-09-15T04:51:54.1315474Z 
evals	Ejecutar las evals	2026-09-15T04:51:54.1316454Z ##[error]
evals	Ejecutar las evals	      at nBn (/$bunfs/root/chunk-y58z5vzc.js:11:25510)
evals	Ejecutar las evals	      at A (/$bunfs/root/chunk-xgzajagg.js:11:3535)
evals	Ejecutar las evals	      at async <anonymous> (/$bunfs/root/chunk-mpjnf6jh.js:88:4824)
evals	Ejecutar las evals	2026-09-15T04:51:54.1317228Z 
evals	Ejecutar las evals	2026-09-15T04:51:54.1317307Z Bun v1.4.3 (Linux x64)
evals	Ejecutar las evals	2026-09-15T04:51:54.1317477Z ```
evals	Ejecutar las evals	2026-09-15T04:51:54.1317561Z 
evals	Ejecutar las evals	2026-09-15T04:51:54.1317691Z ## Sesión 11-no-activa-programacion
evals	Ejecutar las evals	2026-09-15T04:51:54.1317837Z 
evals	Ejecutar las evals	2026-09-15T04:51:54.1317933Z Eval: 11-no-activa-programacion.yaml
evals	Ejecutar las evals	2026-09-15T04:51:54.1318075Z 
evals	Ejecutar las evals	2026-09-15T04:51:54.1318148Z Pregunta:
evals	Ejecutar las evals	2026-09-15T04:51:54.1318241Z 
evals	Ejecutar las evals	2026-09-15T04:51:54.1318316Z ```text
evals	Ejecutar las evals	2026-09-15T04:51:54.1318526Z ¿Cómo invierto una lista enlazada en Go?
evals	Ejecutar las evals	2026-09-15T04:51:54.1318734Z ```
evals	Ejecutar las evals	2026-09-15T04:51:54.1318816Z 
evals	Ejecutar las evals	2026-09-15T04:51:54.1318987Z Invocaciones: ninguna
evals	Ejecutar las evals	2026-09-15T04:51:54.1319096Z 
evals	Ejecutar las evals	2026-09-15T04:51:54.1319192Z Respuesta: vacía
evals	Ejecutar las evals	2026-09-15T04:51:54.1319289Z 
evals	Ejecutar las evals	2026-09-15T04:51:54.1319388Z Motivos de la sesión:
evals	Ejecutar las evals	2026-09-15T04:51:54.1319493Z 
evals	Ejecutar las evals	2026-09-15T04:51:54.1319603Z - la sesión no terminó: código 1
evals	Ejecutar las evals	2026-09-15T04:51:54.1319738Z 
evals	Ejecutar las evals	2026-09-15T04:51:54.1319814Z Salida de error:
evals	Ejecutar las evals	2026-09-15T04:51:54.1319905Z 
evals	Ejecutar las evals	2026-09-15T04:51:54.1319974Z ```text
evals	Ejecutar las evals	2026-09-15T04:51:54.1320245Z  6 | // and may be used to improve Anthropic's products, including training models.
evals	Ejecutar las evals	2026-09-15T04:51:54.1320658Z  7 | // You are responsible for reviewing any code suggestions before use.
evals	Ejecutar las evals	2026-09-15T04:51:54.1320932Z  8 | 
evals	Ejecutar las evals	2026-09-15T04:51:54.1321418Z  9 | // (c) Anthropic PBC. All rights reserved. Use is subject to the Legal Agreements outlined here: https://code.claude.com/docs/en/legal-and-compliance.
evals	Ejecutar las evals	2026-09-15T04:51:54.1321898Z 10 | 
evals	Ejecutar las evals	2026-09-15T04:51:54.1324476Z 11 | import{W,B,he,fN}from"/$bunfs/root/chunk-265sathk.js";import{He,fo,$e}from"/$bunfs/root/chunk-7khdpcz9.js";import{Ude,Yt}from"/$bunfs/root/chunk-4a0graca.js";import{Wi,Dne,Oct,iwe,qct,Lne,Jq}from"/$bunfs/root/chunk-hk7hcqnf.js";import{f}from"/$bunfs/root/chunk-xpzkpaw3.js";import{hx,a}from"/$bunfs/root/chunk-crze9c5c.js";import{Va,l,Mt,hT}from"/$bunfs/root/chunk-a382p5z3.js";import{Mu,qc,K,le,t}from"/$bunfs/root/chunk-0y8dccaw.js";import{lMe,n2n,act,r2n,cMe,Jon,vAr,VSe,PVe,lct,g6,cct,GLt,EAr,uct,IS,zc}from"/$bunfs/root/chunk-ebh8j0r9.js";import{Br}from"/$bunfs/root/chunk-jrz35ppm.js";import{Lue}from"/$bunfs/root/chunk-z0zff54w.js";import{ron}from"/$bunfs/root/chunk-armd3k2m.js";import{gn}from"/$bunfs/root/chunk-aj7ffynp.js";import{aBn}from"/$bunfs/root/chunk-vb37gsz0.js";import{Bo}from"/$bunfs/root/chunk-gk39j4sy.js";import{ie,u,V,R}from"/$bunfs/root/chunk-9kr1d91j.js";import{Y}from"/$bunfs/root/chunk-4112m93q.js";import{homedir as be}from"os";import{dirname as Q,posix as D}from"path";var L=(e)=>JSON.stringif
evals	Ejecutar las evals	2026-09-15T04:51:54.1326992Z 
evals	Ejecutar las evals	2026-09-15T04:51:54.1327633Z error: bubblewrap is required for subprocess env scrubbing and isolation. Install with: sudo apt-get install -y bubblewrap, set sandbox.bwrapPath in managed settings, or set CLAUDE_CODE_SUBPROCESS_ENV_SCRUB=0 to disable (loses subprocess isolation).
evals	Ejecutar las evals	2026-09-15T04:51:54.1328571Z       at nBn (/$bunfs/root/chunk-y58z5vzc.js:11:25510)
evals	Ejecutar las evals	2026-09-15T04:51:54.1328970Z       at A (/$bunfs/root/chunk-xgzajagg.js:11:3535)
evals	Ejecutar las evals	2026-09-15T04:51:54.1329338Z       at async <anonymous> (/$bunfs/root/chunk-mpjnf6jh.js:88:4824)
evals	Ejecutar las evals	2026-09-15T04:51:54.1329536Z 
evals	Ejecutar las evals	2026-09-15T04:51:54.1330503Z ##[error]
evals	Ejecutar las evals	      at nBn (/$bunfs/root/chunk-y58z5vzc.js:11:25510)
evals	Ejecutar las evals	      at A (/$bunfs/root/chunk-xgzajagg.js:11:3535)
evals	Ejecutar las evals	      at async <anonymous> (/$bunfs/root/chunk-mpjnf6jh.js:88:4824)
evals	Ejecutar las evals	2026-09-15T04:51:54.1331350Z 
evals	Ejecutar las evals	2026-09-15T04:51:54.1331434Z Bun v1.4.3 (Linux x64)
evals	Ejecutar las evals	2026-09-15T04:51:54.1331618Z ```
evals	Ejecutar las evals	2026-09-15T04:51:54.1331807Z 
evals	Ejecutar las evals	2026-09-15T04:51:54.1331955Z ## Sesión 12-no-activa-acuerdo-entre-amigos
evals	Ejecutar las evals	2026-09-15T04:51:54.1332112Z 
evals	Ejecutar las evals	2026-09-15T04:51:54.1332226Z Eval: 12-no-activa-acuerdo-entre-amigos.yaml
evals	Ejecutar las evals	2026-09-15T04:51:54.1332396Z 
evals	Ejecutar las evals	2026-09-15T04:51:54.1332463Z Pregunta:
evals	Ejecutar las evals	2026-09-15T04:51:54.1332552Z 
evals	Ejecutar las evals	2026-09-15T04:51:54.1332623Z ```text
evals	Ejecutar las evals	2026-09-15T04:51:54.1333204Z Reescribe en un tono más cercano esta frase de un acuerdo entre amigos para compartir coche: «Las partes se turnarán el uso del vehículo en fines de semana alternos».
evals	Ejecutar las evals	2026-09-15T04:51:54.1333702Z ```
evals	Ejecutar las evals	2026-09-15T04:51:54.1333784Z 
evals	Ejecutar las evals	2026-09-15T04:51:54.1333866Z Invocaciones: ninguna
evals	Ejecutar las evals	2026-09-15T04:51:54.1333971Z 
evals	Ejecutar las evals	2026-09-15T04:51:54.1334068Z Respuesta: vacía
evals	Ejecutar las evals	2026-09-15T04:51:54.1334170Z 
evals	Ejecutar las evals	2026-09-15T04:51:54.1334263Z Motivos de la sesión:
evals	Ejecutar las evals	2026-09-15T04:51:54.1334369Z 
evals	Ejecutar las evals	2026-09-15T04:51:54.1334483Z - la sesión no terminó: código 1
evals	Ejecutar las evals	2026-09-15T04:51:54.1334612Z 
evals	Ejecutar las evals	2026-09-15T04:51:54.1334689Z Salida de error:
evals	Ejecutar las evals	2026-09-15T04:51:54.1334786Z 
evals	Ejecutar las evals	2026-09-15T04:51:54.1334851Z ```text
evals	Ejecutar las evals	2026-09-15T04:51:54.1335145Z  6 | // and may be used to improve Anthropic's products, including training models.
evals	Ejecutar las evals	2026-09-15T04:51:54.1335553Z  7 | // You are responsible for reviewing any code suggestions before use.
evals	Ejecutar las evals	2026-09-15T04:51:54.1335841Z  8 | 
evals	Ejecutar las evals	2026-09-15T04:51:54.1336328Z  9 | // (c) Anthropic PBC. All rights reserved. Use is subject to the Legal Agreements outlined here: https://code.claude.com/docs/en/legal-and-compliance.
evals	Ejecutar las evals	2026-09-15T04:51:54.1336799Z 10 | 
evals	Ejecutar las evals	2026-09-15T04:51:54.1339535Z 11 | import{W,B,he,fN}from"/$bunfs/root/chunk-265sathk.js";import{He,fo,$e}from"/$bunfs/root/chunk-7khdpcz9.js";import{Ude,Yt}from"/$bunfs/root/chunk-4a0graca.js";import{Wi,Dne,Oct,iwe,qct,Lne,Jq}from"/$bunfs/root/chunk-hk7hcqnf.js";import{f}from"/$bunfs/root/chunk-xpzkpaw3.js";import{hx,a}from"/$bunfs/root/chunk-crze9c5c.js";import{Va,l,Mt,hT}from"/$bunfs/root/chunk-a382p5z3.js";import{Mu,qc,K,le,t}from"/$bunfs/root/chunk-0y8dccaw.js";import{lMe,n2n,act,r2n,cMe,Jon,vAr,VSe,PVe,lct,g6,cct,GLt,EAr,uct,IS,zc}from"/$bunfs/root/chunk-ebh8j0r9.js";import{Br}from"/$bunfs/root/chunk-jrz35ppm.js";import{Lue}from"/$bunfs/root/chunk-z0zff54w.js";import{ron}from"/$bunfs/root/chunk-armd3k2m.js";import{gn}from"/$bunfs/root/chunk-aj7ffynp.js";import{aBn}from"/$bunfs/root/chunk-vb37gsz0.js";import{Bo}from"/$bunfs/root/chunk-gk39j4sy.js";import{ie,u,V,R}from"/$bunfs/root/chunk-9kr1d91j.js";import{Y}from"/$bunfs/root/chunk-4112m93q.js";import{homedir as be}from"os";import{dirname as Q,posix as D}from"path";var L=(e)=>JSON.stringif
evals	Ejecutar las evals	2026-09-15T04:51:54.1342083Z 
evals	Ejecutar las evals	2026-09-15T04:51:54.1342726Z error: bubblewrap is required for subprocess env scrubbing and isolation. Install with: sudo apt-get install -y bubblewrap, set sandbox.bwrapPath in managed settings, or set CLAUDE_CODE_SUBPROCESS_ENV_SCRUB=0 to disable (loses subprocess isolation).
evals	Ejecutar las evals	2026-09-15T04:51:54.1343534Z       at nBn (/$bunfs/root/chunk-y58z5vzc.js:11:25510)
evals	Ejecutar las evals	2026-09-15T04:51:54.1343859Z       at A (/$bunfs/root/chunk-xgzajagg.js:11:3535)
evals	Ejecutar las evals	2026-09-15T04:51:54.1344226Z       at async <anonymous> (/$bunfs/root/chunk-mpjnf6jh.js:88:4824)
evals	Ejecutar las evals	2026-09-15T04:51:54.1344422Z 
evals	Ejecutar las evals	2026-09-15T04:51:54.1345404Z ##[error]
evals	Ejecutar las evals	      at nBn (/$bunfs/root/chunk-y58z5vzc.js:11:25510)
evals	Ejecutar las evals	      at A (/$bunfs/root/chunk-xgzajagg.js:11:3535)
evals	Ejecutar las evals	      at async <anonymous> (/$bunfs/root/chunk-mpjnf6jh.js:88:4824)
evals	Ejecutar las evals	2026-09-15T04:51:54.1346289Z 
evals	Ejecutar las evals	2026-09-15T04:51:54.1346369Z Bun v1.4.3 (Linux x64)
evals	Ejecutar las evals	2026-09-15T04:51:54.1346538Z ```
evals	Ejecutar las evals	2026-09-15T04:51:54.1346699Z --- fin de informe.md ---
evals	Ejecutar las evals	2026-09-15T04:51:54.1346907Z --- inicio de informe.json ---
evals	Ejecutar las evals	2026-09-15T04:51:54.1347105Z {
evals	Ejecutar las evals	2026-09-15T04:51:54.1347282Z   "skill": "boe-legislacion",
evals	Ejecutar las evals	2026-09-15T04:51:54.1347516Z   "modelo": "claude-haiku-4-5-20251001",
evals	Ejecutar las evals	2026-09-15T04:51:54.1347755Z   "modelos_de_sesion": [],
evals	Ejecutar las evals	2026-09-15T04:51:54.1347970Z   "versiones_de_claude_code": [],
evals	Ejecutar las evals	2026-09-15T04:51:54.1348240Z   "commit": "417635e39abf5807f8600b79debd6026fc320ce8",
evals	Ejecutar las evals	2026-09-15T04:51:54.1349485Z   "sin_python": "búsqueda: find / ( -path /proc -o -path /sys ) -prune -o ( ( -type f -perm /111 ( -iname python* -o -iname pypy* ) ) -o ( -type l ( -iname python* -o -iname pypy* ) ) -o ( ( -type f -o -type l ) ( -iname libpython* -o -iname libpypy* ) ) ) -print\nusuario: root\nresultado: ninguno\n",
evals	Ejecutar las evals	2026-09-15T04:51:54.1350403Z   "ficheros_mal_formados": [],
evals	Ejecutar las evals	2026-09-15T04:51:54.1350614Z   "veredicto": "fallo",
evals	Ejecutar las evals	2026-09-15T04:51:54.1350803Z   "motivos": [
evals	Ejecutar las evals	2026-09-15T04:51:54.1351115Z     "01-lpac-articulo-21: la sesión no terminó: código 1",
evals	Ejecutar las evals	2026-09-15T04:51:54.1351728Z     "01-lpac-articulo-21: la activación no coincide: se esperaba que la skill boe-legislacion se activara y no se activó",
evals	Ejecutar las evals	2026-09-15T04:51:54.1352279Z     "01-lpac-articulo-21: comando ausente: bloque boe BOE-A-2015-10565 a21",
evals	Ejecutar las evals	2026-09-15T04:51:54.1352668Z     "01-lpac-articulo-21: cita ausente: BOE-A-2015-10565 a21",
evals	Ejecutar las evals	2026-09-15T04:51:54.1353117Z     "01-lpac-articulo-21-prueba-de-red: la sesión no terminó: código 1",
evals	Ejecutar las evals	2026-09-15T04:51:54.1353818Z     "01-lpac-articulo-21-prueba-de-red: la activación no coincide: se esperaba que la skill boe-legislacion se activara y no se activó",
evals	Ejecutar las evals	2026-09-15T04:51:54.1354459Z     "01-lpac-articulo-21-prueba-de-red: comando ausente: bloque boe BOE-A-2015-10565 a21",
evals	Ejecutar las evals	2026-09-15T04:51:54.1354935Z     "01-lpac-articulo-21-prueba-de-red: cita ausente: BOE-A-2015-10565 a21",
evals	Ejecutar las evals	2026-09-15T04:51:54.1355379Z     "02-lcsp-contrato-menor: la sesión no terminó: código 1",
evals	Ejecutar las evals	2026-09-15T04:51:54.1356008Z     "02-lcsp-contrato-menor: la activación no coincide: se esperaba que la skill boe-legislacion se activara y no se activó",
evals	Ejecutar las evals	2026-09-15T04:51:54.1356562Z     "02-lcsp-contrato-menor: comando ausente: boe indice BOE-A-2017-12902",
evals	Ejecutar las evals	2026-09-15T04:51:54.1357014Z     "02-lcsp-contrato-menor: comando ausente: bloque boe BOE-A-2017-12902 a1-30",
evals	Ejecutar las evals	2026-09-15T04:51:54.1357430Z     "02-lcsp-contrato-menor: cita ausente: BOE-A-2017-12902 a1-30",
evals	Ejecutar las evals	2026-09-15T04:51:54.1357884Z     "03-lrbrl-atribuciones-del-pleno: la sesión no terminó: código 1",
evals	Ejecutar las evals	2026-09-15T04:51:54.1358580Z     "03-lrbrl-atribuciones-del-pleno: la activación no coincide: se esperaba que la skill boe-legislacion se activara y no se activó",
evals	Ejecutar las evals	2026-09-15T04:51:54.1359287Z     "03-lrbrl-atribuciones-del-pleno: comando ausente: boe indice BOE-A-1985-5392",
evals	Ejecutar las evals	2026-09-15T04:51:54.1359789Z     "03-lrbrl-atribuciones-del-pleno: comando ausente: bloque boe BOE-A-1985-5392 a22",
evals	Ejecutar las evals	2026-09-15T04:51:54.1360260Z     "03-lrbrl-atribuciones-del-pleno: cita ausente: BOE-A-1985-5392 a22",
evals	Ejecutar las evals	2026-09-15T04:51:54.1360680Z     "04-lgt-prescripcion: la sesión no terminó: código 1",
evals	Ejecutar las evals	2026-09-15T04:51:54.1361278Z     "04-lgt-prescripcion: la activación no coincide: se esperaba que la skill boe-legislacion se activara y no se activó",
evals	Ejecutar las evals	2026-09-15T04:51:54.1361811Z     "04-lgt-prescripcion: comando ausente: boe indice BOE-A-2003-23186",
evals	Ejecutar las evals	2026-09-15T04:51:54.1362226Z     "04-lgt-prescripcion: comando ausente: bloque boe BOE-A-2003-23186 a66",
evals	Ejecutar las evals	2026-09-15T04:51:54.1362615Z     "04-lgt-prescripcion: cita ausente: BOE-A-2003-23186 a66",
evals	Ejecutar las evals	2026-09-15T04:51:54.1363047Z     "05-trlrhl-impuestos-municipales: la sesión no terminó: código 1",
evals	Ejecutar las evals	2026-09-15T04:51:54.1363740Z     "05-trlrhl-impuestos-municipales: la activación no coincide: se esperaba que la skill boe-legislacion se activara y no se activó",
evals	Ejecutar las evals	2026-09-15T04:51:54.1364465Z     "05-trlrhl-impuestos-municipales: comando ausente: boe indice BOE-A-2004-4214",
evals	Ejecutar las evals	2026-09-15T04:51:54.1364961Z     "05-trlrhl-impuestos-municipales: comando ausente: bloque boe BOE-A-2004-4214 a59",
evals	Ejecutar las evals	2026-09-15T04:51:54.1365417Z     "05-trlrhl-impuestos-municipales: cita ausente: BOE-A-2004-4214 a59",
evals	Ejecutar las evals	2026-09-15T04:51:54.1365894Z     "06-irpf-rendimientos-del-trabajo: la sesión no terminó: código 1",
evals	Ejecutar las evals	2026-09-15T04:51:54.1366601Z     "06-irpf-rendimientos-del-trabajo: la activación no coincide: se esperaba que la skill boe-legislacion se activara y no se activó",
evals	Ejecutar las evals	2026-09-15T04:51:54.1367210Z     "06-irpf-rendimientos-del-trabajo: comando ausente: boe indice BOE-A-2006-20764",
evals	Ejecutar las evals	2026-09-15T04:51:54.1367716Z     "06-irpf-rendimientos-del-trabajo: comando ausente: bloque boe BOE-A-2006-20764 a17",
evals	Ejecutar las evals	2026-09-15T04:51:54.1368304Z     "06-irpf-rendimientos-del-trabajo: cita ausente: BOE-A-2006-20764 a17",
evals	Ejecutar las evals	2026-09-15T04:51:54.1368786Z     "07-lrjsp-principio-de-legalidad: la sesión no terminó: código 1",
evals	Ejecutar las evals	2026-09-15T04:51:54.1369571Z     "07-lrjsp-principio-de-legalidad: la activación no coincide: se esperaba que la skill boe-legislacion se activara y no se activó",
evals	Ejecutar las evals	2026-09-15T04:51:54.1370179Z     "07-lrjsp-principio-de-legalidad: comando ausente: boe indice BOE-A-2015-10566",
evals	Ejecutar las evals	2026-09-15T04:51:54.1370670Z     "07-lrjsp-principio-de-legalidad: comando ausente: bloque boe BOE-A-2015-10566 a25",
evals	Ejecutar las evals	2026-09-15T04:51:54.1371133Z     "07-lrjsp-principio-de-legalidad: cita ausente: BOE-A-2015-10566 a25",
evals	Ejecutar las evals	2026-09-15T04:51:54.1371593Z     "08-ltaibg-plazo-de-resolucion: la sesión no terminó: código 1",
evals	Ejecutar las evals	2026-09-15T04:51:54.1372262Z     "08-ltaibg-plazo-de-resolucion: la activación no coincide: se esperaba que la skill boe-legislacion se activara y no se activó",
evals	Ejecutar las evals	2026-09-15T04:51:54.1372856Z     "08-ltaibg-plazo-de-resolucion: comando ausente: boe indice BOE-A-2013-12887",
evals	Ejecutar las evals	2026-09-15T04:51:54.1373347Z     "08-ltaibg-plazo-de-resolucion: comando ausente: bloque boe BOE-A-2013-12887 a20",
evals	Ejecutar las evals	2026-09-15T04:51:54.1373794Z     "08-ltaibg-plazo-de-resolucion: cita ausente: BOE-A-2013-12887 a20",
evals	Ejecutar las evals	2026-09-15T04:51:54.1374252Z     "09-constitucion-articulo-140: la sesión no terminó: código 1",
evals	Ejecutar las evals	2026-09-15T04:51:54.1374918Z     "09-constitucion-articulo-140: la activación no coincide: se esperaba que la skill boe-legislacion se activara y no se activó",
evals	Ejecutar las evals	2026-09-15T04:51:54.1375508Z     "09-constitucion-articulo-140: comando ausente: bloque boe BOE-A-1978-31229 a140",
evals	Ejecutar las evals	2026-09-15T04:51:54.1375947Z     "09-constitucion-articulo-140: cita ausente: BOE-A-1978-31229 a140",
evals	Ejecutar las evals	2026-09-15T04:51:54.1376343Z     "10-et-vacaciones: la sesión no terminó: código 1",
evals	Ejecutar las evals	2026-09-15T04:51:54.1376941Z     "10-et-vacaciones: la activación no coincide: se esperaba que la skill boe-legislacion se activara y no se activó",
evals	Ejecutar las evals	2026-09-15T04:51:54.1377462Z     "10-et-vacaciones: comando ausente: boe indice BOE-A-2015-11430",
evals	Ejecutar las evals	2026-09-15T04:51:54.1377869Z     "10-et-vacaciones: comando ausente: bloque boe BOE-A-2015-11430 a38",
evals	Ejecutar las evals	2026-09-15T04:51:54.1378237Z     "10-et-vacaciones: cita ausente: BOE-A-2015-11430 a38",
evals	Ejecutar las evals	2026-09-15T04:51:54.1378652Z     "11-no-activa-programacion: la sesión no terminó: código 1",
evals	Ejecutar las evals	2026-09-15T04:51:54.1379193Z     "12-no-activa-acuerdo-entre-amigos: la sesión no terminó: código 1"
evals	Ejecutar las evals	2026-09-15T04:51:54.1379470Z   ],
evals	Ejecutar las evals	2026-09-15T04:51:54.1379659Z   "fuera_de_lo_grabado": [],
evals	Ejecutar las evals	2026-09-15T04:51:54.1379859Z   "red": [],
evals	Ejecutar las evals	2026-09-15T04:51:54.1380027Z   "evals": [
evals	Ejecutar las evals	2026-09-15T04:51:54.1380187Z     {
evals	Ejecutar las evals	2026-09-15T04:51:54.1380410Z       "sesion": "01-lpac-articulo-21",
evals	Ejecutar las evals	2026-09-15T04:51:54.1380968Z       "eval": "01-lpac-articulo-21.yaml",
evals	Ejecutar las evals	2026-09-15T04:51:54.1381306Z       "activa": true,
evals	Ejecutar las evals	2026-09-15T04:51:54.1381670Z       "activada": false,
evals	Ejecutar las evals	2026-09-15T04:51:54.1382054Z       "comandos_ejecutados": [],
evals	Ejecutar las evals	2026-09-15T04:51:54.1382379Z       "comandos_ausentes": [
evals	Ejecutar las evals	2026-09-15T04:51:54.1382812Z         "bloque boe BOE-A-2015-10565 a21"
evals	Ejecutar las evals	2026-09-15T04:51:54.1393083Z       ],
evals	Ejecutar las evals	2026-09-15T04:51:54.1393352Z       "citas_encontradas": [],
evals	Ejecutar las evals	2026-09-15T04:51:54.1393773Z       "citas_ausentes": [
evals	Ejecutar las evals	2026-09-15T04:51:54.1394002Z         "BOE-A-2015-10565 a21"
evals	Ejecutar las evals	2026-09-15T04:51:54.1394193Z       ],
evals	Ejecutar las evals	2026-09-15T04:51:54.1394372Z       "invocaciones": [],
evals	Ejecutar las evals	2026-09-15T04:51:54.1394602Z       "fuera_de_lo_grabado": [],
evals	Ejecutar las evals	2026-09-15T04:51:54.1394831Z       "otras_fallidas": [],
evals	Ejecutar las evals	2026-09-15T04:51:54.1395051Z       "llegadas_a_la_red": [],
evals	Ejecutar las evals	2026-09-15T04:51:54.1395267Z       "respuesta": "",
evals	Ejecutar las evals	2026-09-15T04:51:54.1395484Z       "codigo_de_la_sesion": 1,
evals	Ejecutar las evals	2026-09-15T04:51:54.1395733Z       "fin_de_la_sesion": "sin mensajes",
evals	Ejecutar las evals	2026-09-15T04:51:54.1395990Z       "sesion_terminada": false,
evals	Ejecutar las evals	2026-09-15T04:51:54.1396199Z       "motivos": [
evals	Ejecutar las evals	2026-09-15T04:51:54.1396520Z         "la sesión no terminó: código 1",
evals	Ejecutar las evals	2026-09-15T04:51:54.1397120Z         "la activación no coincide: se esperaba que la skill boe-legislacion se activara y no se activó",
evals	Ejecutar las evals	2026-09-15T04:51:54.1397729Z         "comando ausente: bloque boe BOE-A-2015-10565 a21",
evals	Ejecutar las evals	2026-09-15T04:51:54.1398042Z         "cita ausente: BOE-A-2015-10565 a21"
evals	Ejecutar las evals	2026-09-15T04:51:54.1398265Z       ],
evals	Ejecutar las evals	2026-09-15T04:51:54.1398441Z       "pasa": false
evals	Ejecutar las evals	2026-09-15T04:51:54.1398604Z     },
evals	Ejecutar las evals	2026-09-15T04:51:54.1398742Z     {
evals	Ejecutar las evals	2026-09-15T04:51:54.1399093Z       "sesion": "01-lpac-articulo-21-prueba-de-red",
evals	Ejecutar las evals	2026-09-15T04:51:54.1399396Z       "eval": "01-lpac-articulo-21.yaml",
evals	Ejecutar las evals	2026-09-15T04:51:54.1399626Z       "activa": true,
evals	Ejecutar las evals	2026-09-15T04:51:54.1399902Z       "activada": false,
evals	Ejecutar las evals	2026-09-15T04:51:54.1400181Z       "comandos_ejecutados": [],
evals	Ejecutar las evals	2026-09-15T04:51:54.1400404Z       "comandos_ausentes": [
evals	Ejecutar las evals	2026-09-15T04:51:54.1400650Z         "bloque boe BOE-A-2015-10565 a21"
evals	Ejecutar las evals	2026-09-15T04:51:54.1400854Z       ],
evals	Ejecutar las evals	2026-09-15T04:51:54.1401035Z       "citas_encontradas": [],
evals	Ejecutar las evals	2026-09-15T04:51:54.1401246Z       "citas_ausentes": [
evals	Ejecutar las evals	2026-09-15T04:51:54.1401457Z         "BOE-A-2015-10565 a21"
evals	Ejecutar las evals	2026-09-15T04:51:54.1401643Z       ],
evals	Ejecutar las evals	2026-09-15T04:51:54.1401815Z       "invocaciones": [],
evals	Ejecutar las evals	2026-09-15T04:51:54.1402023Z       "fuera_de_lo_grabado": [],
evals	Ejecutar las evals	2026-09-15T04:51:54.1402245Z       "otras_fallidas": [],
evals	Ejecutar las evals	2026-09-15T04:51:54.1402464Z       "llegadas_a_la_red": [],
evals	Ejecutar las evals	2026-09-15T04:51:54.1402666Z       "respuesta": "",
evals	Ejecutar las evals	2026-09-15T04:51:54.1402876Z       "codigo_de_la_sesion": 1,
evals	Ejecutar las evals	2026-09-15T04:51:54.1403117Z       "fin_de_la_sesion": "sin mensajes",
evals	Ejecutar las evals	2026-09-15T04:51:54.1403354Z       "sesion_terminada": false,
evals	Ejecutar las evals	2026-09-15T04:51:54.1403553Z       "motivos": [
evals	Ejecutar las evals	2026-09-15T04:51:54.1403834Z         "la sesión no terminó: código 1",
evals	Ejecutar las evals	2026-09-15T04:51:54.1404389Z         "la activación no coincide: se esperaba que la skill boe-legislacion se activara y no se activó",
evals	Ejecutar las evals	2026-09-15T04:51:54.1404880Z         "comando ausente: bloque boe BOE-A-2015-10565 a21",
evals	Ejecutar las evals	2026-09-15T04:51:54.1405208Z         "cita ausente: BOE-A-2015-10565 a21"
evals	Ejecutar las evals	2026-09-15T04:51:54.1405425Z       ],
evals	Ejecutar las evals	2026-09-15T04:51:54.1405626Z       "pasa": false
evals	Ejecutar las evals	2026-09-15T04:51:54.1405799Z     },
evals	Ejecutar las evals	2026-09-15T04:51:54.1405944Z     {
evals	Ejecutar las evals	2026-09-15T04:51:54.1406154Z       "sesion": "02-lcsp-contrato-menor",
evals	Ejecutar las evals	2026-09-15T04:51:54.1406432Z       "eval": "02-lcsp-contrato-menor.yaml",
evals	Ejecutar las evals	2026-09-15T04:51:54.1406674Z       "activa": true,
evals	Ejecutar las evals	2026-09-15T04:51:54.1406869Z       "activada": false,
evals	Ejecutar las evals	2026-09-15T04:51:54.1407088Z       "comandos_ejecutados": [],
evals	Ejecutar las evals	2026-09-15T04:51:54.1407319Z       "comandos_ausentes": [
evals	Ejecutar las evals	2026-09-15T04:51:54.1407563Z         "boe indice BOE-A-2017-12902",
evals	Ejecutar las evals	2026-09-15T04:51:54.1407833Z         "bloque boe BOE-A-2017-12902 a1-30"
evals	Ejecutar las evals	2026-09-15T04:51:54.1408041Z       ],
evals	Ejecutar las evals	2026-09-15T04:51:54.1408231Z       "citas_encontradas": [],
evals	Ejecutar las evals	2026-09-15T04:51:54.1408445Z       "citas_ausentes": [
evals	Ejecutar las evals	2026-09-15T04:51:54.1408668Z         "BOE-A-2017-12902 a1-30"
evals	Ejecutar las evals	2026-09-15T04:51:54.1408976Z       ],
evals	Ejecutar las evals	2026-09-15T04:51:54.1409158Z       "invocaciones": [],
evals	Ejecutar las evals	2026-09-15T04:51:54.1409369Z       "fuera_de_lo_grabado": [],
evals	Ejecutar las evals	2026-09-15T04:51:54.1409586Z       "otras_fallidas": [],
evals	Ejecutar las evals	2026-09-15T04:51:54.1409805Z       "llegadas_a_la_red": [],
evals	Ejecutar las evals	2026-09-15T04:51:54.1410019Z       "respuesta": "",
evals	Ejecutar las evals	2026-09-15T04:51:54.1410229Z       "codigo_de_la_sesion": 1,
evals	Ejecutar las evals	2026-09-15T04:51:54.1410475Z       "fin_de_la_sesion": "sin mensajes",
evals	Ejecutar las evals	2026-09-15T04:51:54.1410847Z       "sesion_terminada": false,
evals	Ejecutar las evals	2026-09-15T04:51:54.1411051Z       "motivos": [
evals	Ejecutar las evals	2026-09-15T04:51:54.1411328Z         "la sesión no terminó: código 1",
evals	Ejecutar las evals	2026-09-15T04:51:54.1411899Z         "la activación no coincide: se esperaba que la skill boe-legislacion se activara y no se activó",
evals	Ejecutar las evals	2026-09-15T04:51:54.1412356Z         "comando ausente: boe indice BOE-A-2017-12902",
evals	Ejecutar las evals	2026-09-15T04:51:54.1412719Z         "comando ausente: bloque boe BOE-A-2017-12902 a1-30",
evals	Ejecutar las evals	2026-09-15T04:51:54.1413040Z         "cita ausente: BOE-A-2017-12902 a1-30"
evals	Ejecutar las evals	2026-09-15T04:51:54.1413254Z       ],
evals	Ejecutar las evals	2026-09-15T04:51:54.1413418Z       "pasa": false
evals	Ejecutar las evals	2026-09-15T04:51:54.1413582Z     },
evals	Ejecutar las evals	2026-09-15T04:51:54.1413717Z     {
evals	Ejecutar las evals	2026-09-15T04:51:54.1413949Z       "sesion": "03-lrbrl-atribuciones-del-pleno",
evals	Ejecutar las evals	2026-09-15T04:51:54.1414419Z       "eval": "03-lrbrl-atribuciones-del-pleno.yaml",
evals	Ejecutar las evals	2026-09-15T04:51:54.1414673Z       "activa": true,
evals	Ejecutar las evals	2026-09-15T04:51:54.1414862Z       "activada": false,
evals	Ejecutar las evals	2026-09-15T04:51:54.1415091Z       "comandos_ejecutados": [],
evals	Ejecutar las evals	2026-09-15T04:51:54.1415314Z       "comandos_ausentes": [
evals	Ejecutar las evals	2026-09-15T04:51:54.1415555Z         "boe indice BOE-A-1985-5392",
evals	Ejecutar las evals	2026-09-15T04:51:54.1415814Z         "bloque boe BOE-A-1985-5392 a22"
evals	Ejecutar las evals	2026-09-15T04:51:54.1416016Z       ],
evals	Ejecutar las evals	2026-09-15T04:51:54.1416197Z       "citas_encontradas": [],
evals	Ejecutar las evals	2026-09-15T04:51:54.1416407Z       "citas_ausentes": [
evals	Ejecutar las evals	2026-09-15T04:51:54.1416627Z         "BOE-A-1985-5392 a22"
evals	Ejecutar las evals	2026-09-15T04:51:54.1416806Z       ],
evals	Ejecutar las evals	2026-09-15T04:51:54.1416975Z       "invocaciones": [],
evals	Ejecutar las evals	2026-09-15T04:51:54.1417184Z       "fuera_de_lo_grabado": [],
evals	Ejecutar las evals	2026-09-15T04:51:54.1417403Z       "otras_fallidas": [],
evals	Ejecutar las evals	2026-09-15T04:51:54.1417617Z       "llegadas_a_la_red": [],
evals	Ejecutar las evals	2026-09-15T04:51:54.1417819Z       "respuesta": "",
evals	Ejecutar las evals	2026-09-15T04:51:54.1418031Z       "codigo_de_la_sesion": 1,
evals	Ejecutar las evals	2026-09-15T04:51:54.1418276Z       "fin_de_la_sesion": "sin mensajes",
evals	Ejecutar las evals	2026-09-15T04:51:54.1418518Z       "sesion_terminada": false,
evals	Ejecutar las evals	2026-09-15T04:51:54.1418722Z       "motivos": [
evals	Ejecutar las evals	2026-09-15T04:51:54.1419086Z         "la sesión no terminó: código 1",
evals	Ejecutar las evals	2026-09-15T04:51:54.1419663Z         "la activación no coincide: se esperaba que la skill boe-legislacion se activara y no se activó",
evals	Ejecutar las evals	2026-09-15T04:51:54.1420116Z         "comando ausente: boe indice BOE-A-1985-5392",
evals	Ejecutar las evals	2026-09-15T04:51:54.1420464Z         "comando ausente: bloque boe BOE-A-1985-5392 a22",
evals	Ejecutar las evals	2026-09-15T04:51:54.1420763Z         "cita ausente: BOE-A-1985-5392 a22"
evals	Ejecutar las evals	2026-09-15T04:51:54.1420976Z       ],
evals	Ejecutar las evals	2026-09-15T04:51:54.1421145Z       "pasa": false
evals	Ejecutar las evals	2026-09-15T04:51:54.1421305Z     },
evals	Ejecutar las evals	2026-09-15T04:51:54.1421445Z     {
evals	Ejecutar las evals	2026-09-15T04:51:54.1421644Z       "sesion": "04-lgt-prescripcion",
evals	Ejecutar las evals	2026-09-15T04:51:54.1421904Z       "eval": "04-lgt-prescripcion.yaml",
evals	Ejecutar las evals	2026-09-15T04:51:54.1422129Z       "activa": true,
evals	Ejecutar las evals	2026-09-15T04:51:54.1422317Z       "activada": false,
evals	Ejecutar las evals	2026-09-15T04:51:54.1422547Z       "comandos_ejecutados": [],
evals	Ejecutar las evals	2026-09-15T04:51:54.1422775Z       "comandos_ausentes": [
evals	Ejecutar las evals	2026-09-15T04:51:54.1423018Z         "boe indice BOE-A-2003-23186",
evals	Ejecutar las evals	2026-09-15T04:51:54.1423285Z         "bloque boe BOE-A-2003-23186 a66"
evals	Ejecutar las evals	2026-09-15T04:51:54.1423483Z       ],
evals	Ejecutar las evals	2026-09-15T04:51:54.1423662Z       "citas_encontradas": [],
evals	Ejecutar las evals	2026-09-15T04:51:54.1423874Z       "citas_ausentes": [
evals	Ejecutar las evals	2026-09-15T04:51:54.1424094Z         "BOE-A-2003-23186 a66"
evals	Ejecutar las evals	2026-09-15T04:51:54.1424272Z       ],
evals	Ejecutar las evals	2026-09-15T04:51:54.1424443Z       "invocaciones": [],
evals	Ejecutar las evals	2026-09-15T04:51:54.1424650Z       "fuera_de_lo_grabado": [],
evals	Ejecutar las evals	2026-09-15T04:51:54.1424870Z       "otras_fallidas": [],
evals	Ejecutar las evals	2026-09-15T04:51:54.1425080Z       "llegadas_a_la_red": [],
evals	Ejecutar las evals	2026-09-15T04:51:54.1425278Z       "respuesta": "",
evals	Ejecutar las evals	2026-09-15T04:51:54.1425485Z       "codigo_de_la_sesion": 1,
evals	Ejecutar las evals	2026-09-15T04:51:54.1425724Z       "fin_de_la_sesion": "sin mensajes",
evals	Ejecutar las evals	2026-09-15T04:51:54.1425960Z       "sesion_terminada": false,
evals	Ejecutar las evals	2026-09-15T04:51:54.1426166Z       "motivos": [
evals	Ejecutar las evals	2026-09-15T04:51:54.1426431Z         "la sesión no terminó: código 1",
evals	Ejecutar las evals	2026-09-15T04:51:54.1426977Z         "la activación no coincide: se esperaba que la skill boe-legislacion se activara y no se activó",
evals	Ejecutar las evals	2026-09-15T04:51:54.1427733Z         "comando ausente: boe indice BOE-A-2003-23186",
evals	Ejecutar las evals	2026-09-15T04:51:54.1428126Z         "comando ausente: bloque boe BOE-A-2003-23186 a66",
evals	Ejecutar las evals	2026-09-15T04:51:54.1428448Z         "cita ausente: BOE-A-2003-23186 a66"
evals	Ejecutar las evals	2026-09-15T04:51:54.1428661Z       ],
evals	Ejecutar las evals	2026-09-15T04:51:54.1428829Z       "pasa": false
evals	Ejecutar las evals	2026-09-15T04:51:54.1429149Z     },
evals	Ejecutar las evals	2026-09-15T04:51:54.1429291Z     {
evals	Ejecutar las evals	2026-09-15T04:51:54.1429530Z       "sesion": "05-trlrhl-impuestos-municipales",
evals	Ejecutar las evals	2026-09-15T04:51:54.1429855Z       "eval": "05-trlrhl-impuestos-municipales.yaml",
evals	Ejecutar las evals	2026-09-15T04:51:54.1430109Z       "activa": true,
evals	Ejecutar las evals	2026-09-15T04:51:54.1430303Z       "activada": false,
evals	Ejecutar las evals	2026-09-15T04:51:54.1430521Z       "comandos_ejecutados": [],
evals	Ejecutar las evals	2026-09-15T04:51:54.1430746Z       "comandos_ausentes": [
evals	Ejecutar las evals	2026-09-15T04:51:54.1431120Z         "boe indice BOE-A-2004-4214",
evals	Ejecutar las evals	2026-09-15T04:51:54.1431789Z         "bloque boe BOE-A-2004-4214 a59"
evals	Ejecutar las evals	2026-09-15T04:51:54.1432127Z       ],
evals	Ejecutar las evals	2026-09-15T04:51:54.1432420Z       "citas_encontradas": [],
evals	Ejecutar las evals	2026-09-15T04:51:54.1432768Z       "citas_ausentes": [
evals	Ejecutar las evals	2026-09-15T04:51:54.1433112Z         "BOE-A-2004-4214 a59"
evals	Ejecutar las evals	2026-09-15T04:51:54.1433396Z       ],
evals	Ejecutar las evals	2026-09-15T04:51:54.1433670Z       "invocaciones": [],
evals	Ejecutar las evals	2026-09-15T04:51:54.1434011Z       "fuera_de_lo_grabado": [],
evals	Ejecutar las evals	2026-09-15T04:51:54.1434371Z       "otras_fallidas": [],
evals	Ejecutar las evals	2026-09-15T04:51:54.1434711Z       "llegadas_a_la_red": [],
evals	Ejecutar las evals	2026-09-15T04:51:54.1435032Z       "respuesta": "",
evals	Ejecutar las evals	2026-09-15T04:51:54.1435248Z       "codigo_de_la_sesion": 1,
evals	Ejecutar las evals	2026-09-15T04:51:54.1435496Z       "fin_de_la_sesion": "sin mensajes",
evals	Ejecutar las evals	2026-09-15T04:51:54.1435751Z       "sesion_terminada": false,
evals	Ejecutar las evals	2026-09-15T04:51:54.1435957Z       "motivos": [
evals	Ejecutar las evals	2026-09-15T04:51:54.1436249Z         "la sesión no terminó: código 1",
evals	Ejecutar las evals	2026-09-15T04:51:54.1436816Z         "la activación no coincide: se esperaba que la skill boe-legislacion se activara y no se activó",
evals	Ejecutar las evals	2026-09-15T04:51:54.1437288Z         "comando ausente: boe indice BOE-A-2004-4214",
evals	Ejecutar las evals	2026-09-15T04:51:54.1437638Z         "comando ausente: bloque boe BOE-A-2004-4214 a59",
evals	Ejecutar las evals	2026-09-15T04:51:54.1437942Z         "cita ausente: BOE-A-2004-4214 a59"
evals	Ejecutar las evals	2026-09-15T04:51:54.1438155Z       ],
evals	Ejecutar las evals	2026-09-15T04:51:54.1438396Z       "pasa": false
evals	Ejecutar las evals	2026-09-15T04:51:54.1438634Z     },
evals	Ejecutar las evals	2026-09-15T04:51:54.1438834Z     {
evals	Ejecutar las evals	2026-09-15T04:51:54.1439231Z       "sesion": "06-irpf-rendimientos-del-trabajo",
evals	Ejecutar las evals	2026-09-15T04:51:54.1439560Z       "eval": "06-irpf-rendimientos-del-trabajo.yaml",
evals	Ejecutar las evals	2026-09-15T04:51:54.1439813Z       "activa": true,
evals	Ejecutar las evals	2026-09-15T04:51:54.1440010Z       "activada": false,
evals	Ejecutar las evals	2026-09-15T04:51:54.1440227Z       "comandos_ejecutados": [],
evals	Ejecutar las evals	2026-09-15T04:51:54.1440454Z       "comandos_ausentes": [
evals	Ejecutar las evals	2026-09-15T04:51:54.1440708Z         "boe indice BOE-A-2006-20764",
evals	Ejecutar las evals	2026-09-15T04:51:54.1440976Z         "bloque boe BOE-A-2006-20764 a17"
evals	Ejecutar las evals	2026-09-15T04:51:54.1441188Z       ],
evals	Ejecutar las evals	2026-09-15T04:51:54.1441374Z       "citas_encontradas": [],
evals	Ejecutar las evals	2026-09-15T04:51:54.1441588Z       "citas_ausentes": [
evals	Ejecutar las evals	2026-09-15T04:51:54.1441808Z         "BOE-A-2006-20764 a17"
evals	Ejecutar las evals	2026-09-15T04:51:54.1441984Z       ],
evals	Ejecutar las evals	2026-09-15T04:51:54.1442160Z       "invocaciones": [],
evals	Ejecutar las evals	2026-09-15T04:51:54.1442369Z       "fuera_de_lo_grabado": [],
evals	Ejecutar las evals	2026-09-15T04:51:54.1442586Z       "otras_fallidas": [],
evals	Ejecutar las evals	2026-09-15T04:51:54.1442804Z       "llegadas_a_la_red": [],
evals	Ejecutar las evals	2026-09-15T04:51:54.1443007Z       "respuesta": "",
evals	Ejecutar las evals	2026-09-15T04:51:54.1443213Z       "codigo_de_la_sesion": 1,
evals	Ejecutar las evals	2026-09-15T04:51:54.1443454Z       "fin_de_la_sesion": "sin mensajes",
evals	Ejecutar las evals	2026-09-15T04:51:54.1443698Z       "sesion_terminada": false,
evals	Ejecutar las evals	2026-09-15T04:51:54.1443901Z       "motivos": [
evals	Ejecutar las evals	2026-09-15T04:51:54.1444179Z         "la sesión no terminó: código 1",
evals	Ejecutar las evals	2026-09-15T04:51:54.1444735Z         "la activación no coincide: se esperaba que la skill boe-legislacion se activara y no se activó",
evals	Ejecutar las evals	2026-09-15T04:51:54.1445188Z         "comando ausente: boe indice BOE-A-2006-20764",
evals	Ejecutar las evals	2026-09-15T04:51:54.1445535Z         "comando ausente: bloque boe BOE-A-2006-20764 a17",
evals	Ejecutar las evals	2026-09-15T04:51:54.1445995Z         "cita ausente: BOE-A-2006-20764 a17"
evals	Ejecutar las evals	2026-09-15T04:51:54.1446201Z       ],
evals	Ejecutar las evals	2026-09-15T04:51:54.1446361Z       "pasa": false
evals	Ejecutar las evals	2026-09-15T04:51:54.1446529Z     },
evals	Ejecutar las evals	2026-09-15T04:51:54.1446672Z     {
evals	Ejecutar las evals	2026-09-15T04:51:54.1446912Z       "sesion": "07-lrjsp-principio-de-legalidad",
evals	Ejecutar las evals	2026-09-15T04:51:54.1447240Z       "eval": "07-lrjsp-principio-de-legalidad.yaml",
evals	Ejecutar las evals	2026-09-15T04:51:54.1447496Z       "activa": true,
evals	Ejecutar las evals	2026-09-15T04:51:54.1447688Z       "activada": false,
evals	Ejecutar las evals	2026-09-15T04:51:54.1447906Z       "comandos_ejecutados": [],
evals	Ejecutar las evals	2026-09-15T04:51:54.1448135Z       "comandos_ausentes": [
evals	Ejecutar las evals	2026-09-15T04:51:54.1448380Z         "boe indice BOE-A-2015-10566",
evals	Ejecutar las evals	2026-09-15T04:51:54.1448645Z         "bloque boe BOE-A-2015-10566 a25"
evals	Ejecutar las evals	2026-09-15T04:51:54.1449030Z       ],
evals	Ejecutar las evals	2026-09-15T04:51:54.1449222Z       "citas_encontradas": [],
evals	Ejecutar las evals	2026-09-15T04:51:54.1449546Z       "citas_ausentes": [
evals	Ejecutar las evals	2026-09-15T04:51:54.1449765Z         "BOE-A-2015-10566 a25"
evals	Ejecutar las evals	2026-09-15T04:51:54.1449947Z       ],
evals	Ejecutar las evals	2026-09-15T04:51:54.1450119Z       "invocaciones": [],
evals	Ejecutar las evals	2026-09-15T04:51:54.1450338Z       "fuera_de_lo_grabado": [],
evals	Ejecutar las evals	2026-09-15T04:51:54.1450563Z       "otras_fallidas": [],
evals	Ejecutar las evals	2026-09-15T04:51:54.1450779Z       "llegadas_a_la_red": [],
evals	Ejecutar las evals	2026-09-15T04:51:54.1450984Z       "respuesta": "",
evals	Ejecutar las evals	2026-09-15T04:51:54.1451196Z       "codigo_de_la_sesion": 1,
evals	Ejecutar las evals	2026-09-15T04:51:54.1451442Z       "fin_de_la_sesion": "sin mensajes",
evals	Ejecutar las evals	2026-09-15T04:51:54.1451685Z       "sesion_terminada": false,
evals	Ejecutar las evals	2026-09-15T04:51:54.1451889Z       "motivos": [
evals	Ejecutar las evals	2026-09-15T04:51:54.1452170Z         "la sesión no terminó: código 1",
evals	Ejecutar las evals	2026-09-15T04:51:54.1452724Z         "la activación no coincide: se esperaba que la skill boe-legislacion se activara y no se activó",
evals	Ejecutar las evals	2026-09-15T04:51:54.1453193Z         "comando ausente: boe indice BOE-A-2015-10566",
evals	Ejecutar las evals	2026-09-15T04:51:54.1453541Z         "comando ausente: bloque boe BOE-A-2015-10566 a25",
evals	Ejecutar las evals	2026-09-15T04:51:54.1453862Z         "cita ausente: BOE-A-2015-10566 a25"
evals	Ejecutar las evals	2026-09-15T04:51:54.1454071Z       ],
evals	Ejecutar las evals	2026-09-15T04:51:54.1454240Z       "pasa": false
evals	Ejecutar las evals	2026-09-15T04:51:54.1454407Z     },
evals	Ejecutar las evals	2026-09-15T04:51:54.1454547Z     {
evals	Ejecutar las evals	2026-09-15T04:51:54.1454774Z       "sesion": "08-ltaibg-plazo-de-resolucion",
evals	Ejecutar las evals	2026-09-15T04:51:54.1455085Z       "eval": "08-ltaibg-plazo-de-resolucion.yaml",
evals	Ejecutar las evals	2026-09-15T04:51:54.1455336Z       "activa": true,
evals	Ejecutar las evals	2026-09-15T04:51:54.1455529Z       "activada": false,
evals	Ejecutar las evals	2026-09-15T04:51:54.1455749Z       "comandos_ejecutados": [],
evals	Ejecutar las evals	2026-09-15T04:51:54.1455975Z       "comandos_ausentes": [
evals	Ejecutar las evals	2026-09-15T04:51:54.1456216Z         "boe indice BOE-A-2013-12887",
evals	Ejecutar las evals	2026-09-15T04:51:54.1456485Z         "bloque boe BOE-A-2013-12887 a20"
evals	Ejecutar las evals	2026-09-15T04:51:54.1456690Z       ],
evals	Ejecutar las evals	2026-09-15T04:51:54.1456872Z       "citas_encontradas": [],
evals	Ejecutar las evals	2026-09-15T04:51:54.1457088Z       "citas_ausentes": [
evals	Ejecutar las evals	2026-09-15T04:51:54.1457304Z         "BOE-A-2013-12887 a20"
evals	Ejecutar las evals	2026-09-15T04:51:54.1457481Z       ],
evals	Ejecutar las evals	2026-09-15T04:51:54.1457658Z       "invocaciones": [],
evals	Ejecutar las evals	2026-09-15T04:51:54.1457873Z       "fuera_de_lo_grabado": [],
evals	Ejecutar las evals	2026-09-15T04:51:54.1458090Z       "otras_fallidas": [],
evals	Ejecutar las evals	2026-09-15T04:51:54.1458309Z       "llegadas_a_la_red": [],
evals	Ejecutar las evals	2026-09-15T04:51:54.1458515Z       "respuesta": "",
evals	Ejecutar las evals	2026-09-15T04:51:54.1458725Z       "codigo_de_la_sesion": 1,
evals	Ejecutar las evals	2026-09-15T04:51:54.1459058Z       "fin_de_la_sesion": "sin mensajes",
evals	Ejecutar las evals	2026-09-15T04:51:54.1459301Z       "sesion_terminada": false,
evals	Ejecutar las evals	2026-09-15T04:51:54.1459505Z       "motivos": [
evals	Ejecutar las evals	2026-09-15T04:51:54.1459775Z         "la sesión no terminó: código 1",
evals	Ejecutar las evals	2026-09-15T04:51:54.1460323Z         "la activación no coincide: se esperaba que la skill boe-legislacion se activara y no se activó",
evals	Ejecutar las evals	2026-09-15T04:51:54.1460792Z         "comando ausente: boe indice BOE-A-2013-12887",
evals	Ejecutar las evals	2026-09-15T04:51:54.1461138Z         "comando ausente: bloque boe BOE-A-2013-12887 a20",
evals	Ejecutar las evals	2026-09-15T04:51:54.1461446Z         "cita ausente: BOE-A-2013-12887 a20"
evals	Ejecutar las evals	2026-09-15T04:51:54.1461656Z       ],
evals	Ejecutar las evals	2026-09-15T04:51:54.1461832Z       "pasa": false
evals	Ejecutar las evals	2026-09-15T04:51:54.1461993Z     },
evals	Ejecutar las evals	2026-09-15T04:51:54.1462131Z     {
evals	Ejecutar las evals	2026-09-15T04:51:54.1462361Z       "sesion": "09-constitucion-articulo-140",
evals	Ejecutar las evals	2026-09-15T04:51:54.1462784Z       "eval": "09-constitucion-articulo-140.yaml",
evals	Ejecutar las evals	2026-09-15T04:51:54.1463032Z       "activa": true,
evals	Ejecutar las evals	2026-09-15T04:51:54.1463230Z       "activada": false,
evals	Ejecutar las evals	2026-09-15T04:51:54.1463447Z       "comandos_ejecutados": [],
evals	Ejecutar las evals	2026-09-15T04:51:54.1463668Z       "comandos_ausentes": [
evals	Ejecutar las evals	2026-09-15T04:51:54.1463922Z         "bloque boe BOE-A-1978-31229 a140"
evals	Ejecutar las evals	2026-09-15T04:51:54.1464127Z       ],
evals	Ejecutar las evals	2026-09-15T04:51:54.1464310Z       "citas_encontradas": [],
evals	Ejecutar las evals	2026-09-15T04:51:54.1464522Z       "citas_ausentes": [
evals	Ejecutar las evals	2026-09-15T04:51:54.1464746Z         "BOE-A-1978-31229 a140"
evals	Ejecutar las evals	2026-09-15T04:51:54.1464936Z       ],
evals	Ejecutar las evals	2026-09-15T04:51:54.1465109Z       "invocaciones": [],
evals	Ejecutar las evals	2026-09-15T04:51:54.1465326Z       "fuera_de_lo_grabado": [],
evals	Ejecutar las evals	2026-09-15T04:51:54.1465545Z       "otras_fallidas": [],
evals	Ejecutar las evals	2026-09-15T04:51:54.1465758Z       "llegadas_a_la_red": [],
evals	Ejecutar las evals	2026-09-15T04:51:54.1466070Z       "respuesta": "",
evals	Ejecutar las evals	2026-09-15T04:51:54.1466286Z       "codigo_de_la_sesion": 1,
evals	Ejecutar las evals	2026-09-15T04:51:54.1466540Z       "fin_de_la_sesion": "sin mensajes",
evals	Ejecutar las evals	2026-09-15T04:51:54.1466786Z       "sesion_terminada": false,
evals	Ejecutar las evals	2026-09-15T04:51:54.1466992Z       "motivos": [
evals	Ejecutar las evals	2026-09-15T04:51:54.1467268Z         "la sesión no terminó: código 1",
evals	Ejecutar las evals	2026-09-15T04:51:54.1467838Z         "la activación no coincide: se esperaba que la skill boe-legislacion se activara y no se activó",
evals	Ejecutar las evals	2026-09-15T04:51:54.1468315Z         "comando ausente: bloque boe BOE-A-1978-31229 a140",
evals	Ejecutar las evals	2026-09-15T04:51:54.1468643Z         "cita ausente: BOE-A-1978-31229 a140"
evals	Ejecutar las evals	2026-09-15T04:51:54.1468968Z       ],
evals	Ejecutar las evals	2026-09-15T04:51:54.1469141Z       "pasa": false
evals	Ejecutar las evals	2026-09-15T04:51:54.1469307Z     },
evals	Ejecutar las evals	2026-09-15T04:51:54.1469452Z     {
evals	Ejecutar las evals	2026-09-15T04:51:54.1469648Z       "sesion": "10-et-vacaciones",
evals	Ejecutar las evals	2026-09-15T04:51:54.1469909Z       "eval": "10-et-vacaciones.yaml",
evals	Ejecutar las evals	2026-09-15T04:51:54.1470136Z       "activa": true,
evals	Ejecutar las evals	2026-09-15T04:51:54.1470341Z       "activada": false,
evals	Ejecutar las evals	2026-09-15T04:51:54.1470560Z       "comandos_ejecutados": [],
evals	Ejecutar las evals	2026-09-15T04:51:54.1470784Z       "comandos_ausentes": [
evals	Ejecutar las evals	2026-09-15T04:51:54.1471037Z         "boe indice BOE-A-2015-11430",
evals	Ejecutar las evals	2026-09-15T04:51:54.1471300Z         "bloque boe BOE-A-2015-11430 a38"
evals	Ejecutar las evals	2026-09-15T04:51:54.1471508Z       ],
evals	Ejecutar las evals	2026-09-15T04:51:54.1471690Z       "citas_encontradas": [],
evals	Ejecutar las evals	2026-09-15T04:51:54.1471901Z       "citas_ausentes": [
evals	Ejecutar las evals	2026-09-15T04:51:54.1472120Z         "BOE-A-2015-11430 a38"
evals	Ejecutar las evals	2026-09-15T04:51:54.1472300Z       ],
evals	Ejecutar las evals	2026-09-15T04:51:54.1472474Z       "invocaciones": [],
evals	Ejecutar las evals	2026-09-15T04:51:54.1472691Z       "fuera_de_lo_grabado": [],
evals	Ejecutar las evals	2026-09-15T04:51:54.1472911Z       "otras_fallidas": [],
evals	Ejecutar las evals	2026-09-15T04:51:54.1473122Z       "llegadas_a_la_red": [],
evals	Ejecutar las evals	2026-09-15T04:51:54.1473327Z       "respuesta": "",
evals	Ejecutar las evals	2026-09-15T04:51:54.1473537Z       "codigo_de_la_sesion": 1,
evals	Ejecutar las evals	2026-09-15T04:51:54.1473774Z       "fin_de_la_sesion": "sin mensajes",
evals	Ejecutar las evals	2026-09-15T04:51:54.1474016Z       "sesion_terminada": false,
evals	Ejecutar las evals	2026-09-15T04:51:54.1474225Z       "motivos": [
evals	Ejecutar las evals	2026-09-15T04:51:54.1474494Z         "la sesión no terminó: código 1",
evals	Ejecutar las evals	2026-09-15T04:51:54.1475057Z         "la activación no coincide: se esperaba que la skill boe-legislacion se activara y no se activó",
evals	Ejecutar las evals	2026-09-15T04:51:54.1475520Z         "comando ausente: boe indice BOE-A-2015-11430",
evals	Ejecutar las evals	2026-09-15T04:51:54.1475863Z         "comando ausente: bloque boe BOE-A-2015-11430 a38",
evals	Ejecutar las evals	2026-09-15T04:51:54.1476173Z         "cita ausente: BOE-A-2015-11430 a38"
evals	Ejecutar las evals	2026-09-15T04:51:54.1476385Z       ],
evals	Ejecutar las evals	2026-09-15T04:51:54.1476549Z       "pasa": false
evals	Ejecutar las evals	2026-09-15T04:51:54.1476712Z     },
evals	Ejecutar las evals	2026-09-15T04:51:54.1476902Z     {
evals	Ejecutar las evals	2026-09-15T04:51:54.1477121Z       "sesion": "11-no-activa-programacion",
evals	Ejecutar las evals	2026-09-15T04:51:54.1477411Z       "eval": "11-no-activa-programacion.yaml",
evals	Ejecutar las evals	2026-09-15T04:51:54.1477651Z       "activa": false,
evals	Ejecutar las evals	2026-09-15T04:51:54.1477849Z       "activada": false,
evals	Ejecutar las evals	2026-09-15T04:51:54.1478068Z       "comandos_ejecutados": [],
evals	Ejecutar las evals	2026-09-15T04:51:54.1478303Z       "comandos_ausentes": [],
evals	Ejecutar las evals	2026-09-15T04:51:54.1478524Z       "citas_encontradas": [],
evals	Ejecutar las evals	2026-09-15T04:51:54.1478741Z       "citas_ausentes": [],
evals	Ejecutar las evals	2026-09-15T04:51:54.1479171Z       "invocaciones": [],
evals	Ejecutar las evals	2026-09-15T04:51:54.1479388Z       "fuera_de_lo_grabado": [],
evals	Ejecutar las evals	2026-09-15T04:51:54.1479607Z       "otras_fallidas": [],
evals	Ejecutar las evals	2026-09-15T04:51:54.1479816Z       "llegadas_a_la_red": [],
evals	Ejecutar las evals	2026-09-15T04:51:54.1480019Z       "respuesta": "",
evals	Ejecutar las evals	2026-09-15T04:51:54.1480226Z       "codigo_de_la_sesion": 1,
evals	Ejecutar las evals	2026-09-15T04:51:54.1480463Z       "fin_de_la_sesion": "sin mensajes",
evals	Ejecutar las evals	2026-09-15T04:51:54.1480706Z       "sesion_terminada": false,
evals	Ejecutar las evals	2026-09-15T04:51:54.1480906Z       "motivos": [
evals	Ejecutar las evals	2026-09-15T04:51:54.1481174Z         "la sesión no terminó: código 1"
evals	Ejecutar las evals	2026-09-15T04:51:54.1481384Z       ],
evals	Ejecutar las evals	2026-09-15T04:51:54.1481548Z       "pasa": false
evals	Ejecutar las evals	2026-09-15T04:51:54.1481711Z     },
evals	Ejecutar las evals	2026-09-15T04:51:54.1481848Z     {
evals	Ejecutar las evals	2026-09-15T04:51:54.1482091Z       "sesion": "12-no-activa-acuerdo-entre-amigos",
evals	Ejecutar las evals	2026-09-15T04:51:54.1482535Z       "eval": "12-no-activa-acuerdo-entre-amigos.yaml",
evals	Ejecutar las evals	2026-09-15T04:51:54.1482796Z       "activa": false,
evals	Ejecutar las evals	2026-09-15T04:51:54.1482986Z       "activada": false,
evals	Ejecutar las evals	2026-09-15T04:51:54.1483204Z       "comandos_ejecutados": [],
evals	Ejecutar las evals	2026-09-15T04:51:54.1483433Z       "comandos_ausentes": [],
evals	Ejecutar las evals	2026-09-15T04:51:54.1483652Z       "citas_encontradas": [],
evals	Ejecutar las evals	2026-09-15T04:51:54.1483864Z       "citas_ausentes": [],
evals	Ejecutar las evals	2026-09-15T04:51:54.1484073Z       "invocaciones": [],
evals	Ejecutar las evals	2026-09-15T04:51:54.1484287Z       "fuera_de_lo_grabado": [],
evals	Ejecutar las evals	2026-09-15T04:51:54.1484502Z       "otras_fallidas": [],
evals	Ejecutar las evals	2026-09-15T04:51:54.1484711Z       "llegadas_a_la_red": [],
evals	Ejecutar las evals	2026-09-15T04:51:54.1484919Z       "respuesta": "",
evals	Ejecutar las evals	2026-09-15T04:51:54.1485133Z       "codigo_de_la_sesion": 1,
evals	Ejecutar las evals	2026-09-15T04:51:54.1485381Z       "fin_de_la_sesion": "sin mensajes",
evals	Ejecutar las evals	2026-09-15T04:51:54.1485637Z       "sesion_terminada": false,
evals	Ejecutar las evals	2026-09-15T04:51:54.1485846Z       "motivos": [
evals	Ejecutar las evals	2026-09-15T04:51:54.1486119Z         "la sesión no terminó: código 1"
evals	Ejecutar las evals	2026-09-15T04:51:54.1486338Z       ],
evals	Ejecutar las evals	2026-09-15T04:51:54.1486503Z       "pasa": false
evals	Ejecutar las evals	2026-09-15T04:51:54.1486658Z     }
evals	Ejecutar las evals	2026-09-15T04:51:54.1486798Z   ]
evals	Ejecutar las evals	2026-09-15T04:51:54.1486946Z }
evals	Ejecutar las evals	2026-09-15T04:51:54.1487121Z --- fin de informe.json ---
```

## Anexo B · Sexta orden de §12.2: retirada de Python entre marcas, tal cual

```text
evals	Retirar Python del runner	2026-09-15T04:49:42.4508460Z --- inicio de la retirada de Python ---
evals	Retirar Python del runner	2026-09-15T04:49:42.4675524Z búsqueda: find / ( -path /proc -o -path /sys ) -prune -o ( ( -type f -perm /111 ( -iname python* -o -iname pypy* ) ) -o ( -type l ( -iname python* -o -iname pypy* ) ) -o ( ( -type f -o -type l ) ( -iname libpython* -o -iname libpypy* ) ) ) -print
evals	Retirar Python del runner	2026-09-15T04:51:05.1242583Z retirado: /opt/pipx/shared/lib/python3.12/site-packages/pip/_vendor/pygments/lexers/__pycache__/python.cpython-312.pyc
evals	Retirar Python del runner	2026-09-15T04:51:05.1383431Z retirado: /opt/pipx/shared/lib/python3.12/site-packages/pip/_vendor/pygments/lexers/python.py
evals	Retirar Python del runner	2026-09-15T04:51:05.1654175Z retirado: /opt/pipx/shared
evals	Retirar Python del runner	2026-09-15T04:51:05.2353031Z retirado: /opt/pipx/venvs/yamllint
evals	Retirar Python del runner	2026-09-15T04:51:05.2773458Z retirado: /opt/pipx/venvs/ansible-core/lib/python3.12/site-packages/ansible_test/_util/target/injector/__pycache__/python.cpython-312.pyc
evals	Retirar Python del runner	2026-09-15T04:51:05.2912663Z retirado: /opt/pipx/venvs/ansible-core/lib/python3.12/site-packages/ansible_test/_util/target/injector/python.py
evals	Retirar Python del runner	2026-09-15T04:51:05.3050311Z retirado: /opt/pipx/venvs/ansible-core/lib/python3.12/site-packages/ansible_test/_internal/__pycache__/python_requirements.cpython-312.pyc
evals	Retirar Python del runner	2026-09-15T04:51:05.3187163Z retirado: /opt/pipx/venvs/ansible-core/lib/python3.12/site-packages/ansible_test/_internal/classification/__pycache__/python.cpython-312.pyc
evals	Retirar Python del runner	2026-09-15T04:51:05.3325844Z retirado: /opt/pipx/venvs/ansible-core/lib/python3.12/site-packages/ansible_test/_internal/classification/python.py
evals	Retirar Python del runner	2026-09-15T04:51:05.3463637Z retirado: /opt/pipx/venvs/ansible-core/lib/python3.12/site-packages/ansible_test/_internal/python_requirements.py
evals	Retirar Python del runner	2026-09-15T04:51:05.3604127Z retirado: /opt/pipx/venvs/ansible-core/lib/python3.12/site-packages/ansible_collections/community/okd/molecule/default/roles/openshift_adm_groups/tasks/python-ldap-not-installed.yml
evals	Retirar Python del runner	2026-09-15T04:51:05.3745398Z retirado: /opt/pipx/venvs/ansible-core/lib/python3.12/site-packages/ansible_collections/community/general/plugins/modules/__pycache__/python_requirements_info.cpython-312.pyc
evals	Retirar Python del runner	2026-09-15T04:51:05.3887508Z retirado: /opt/pipx/venvs/ansible-core/lib/python3.12/site-packages/ansible_collections/community/general/plugins/modules/python_requirements_info.py
evals	Retirar Python del runner	2026-09-15T04:51:05.4025006Z retirado: /opt/pipx/venvs/ansible-core/lib/python3.12/site-packages/ansible/_internal/ansible_collections/ansible/_protomatter/plugins/filter/__pycache__/python_literal_eval.cpython-312.pyc
evals	Retirar Python del runner	2026-09-15T04:51:05.4163148Z retirado: /opt/pipx/venvs/ansible-core/lib/python3.12/site-packages/ansible/_internal/ansible_collections/ansible/_protomatter/plugins/filter/python_literal_eval.yml
evals	Retirar Python del runner	2026-09-15T04:51:05.4301564Z retirado: /opt/pipx/venvs/ansible-core/lib/python3.12/site-packages/ansible/_internal/ansible_collections/ansible/_protomatter/plugins/filter/python_literal_eval.py
evals	Retirar Python del runner	2026-09-15T04:51:05.4440405Z retirado: /opt/pipx/venvs/ansible-core/lib/python3.12/site-packages/ansible/module_utils/facts/system/__pycache__/python.cpython-312.pyc
evals	Retirar Python del runner	2026-09-15T04:51:05.4577586Z retirado: /opt/pipx/venvs/ansible-core/lib/python3.12/site-packages/ansible/module_utils/facts/system/python.py
evals	Retirar Python del runner	2026-09-15T04:51:05.4845706Z retirado: /opt/pipx/venvs/ansible-core
evals	Retirar Python del runner	2026-09-15T04:51:06.3060892Z retirado: /opt/hostedtoolcache/PyPy/3.9.19/x64/lib/pypy3.9/site-packages/pip/_vendor/pygments/lexers/__pycache__/python.pypy39.pyc
evals	Retirar Python del runner	2026-09-15T04:51:06.3190624Z retirado: /opt/hostedtoolcache/PyPy/3.9.19/x64/lib/pypy3.9/site-packages/pip/_vendor/pygments/lexers/python.py
evals	Retirar Python del runner	2026-09-15T04:51:06.3327387Z retirado: /opt/hostedtoolcache/PyPy/3.9.19/x64/lib/pypy3.9/_cffi_ssl/_cffi_src/openssl/pypy_win32_extra.py
evals	Retirar Python del runner	2026-09-15T04:51:06.3465007Z retirado: /opt/hostedtoolcache/PyPy/3.9.19/x64/lib/pypy3.9/hpy/devel/include/hpy/forbid_python_h/Python.h
evals	Retirar Python del runner	2026-09-15T04:51:06.3602149Z retirado: /opt/hostedtoolcache/PyPy/3.9.19/x64/lib/pypy3.9/pyrepl/python_reader.py
evals	Retirar Python del runner	2026-09-15T04:51:06.3738767Z retirado: /opt/hostedtoolcache/PyPy/3.9.19/x64/lib/pypy3.9/idlelib/Icons/python.gif
evals	Retirar Python del runner	2026-09-15T04:51:06.3876380Z retirado: /opt/hostedtoolcache/PyPy/3.9.19/x64/lib/pypy3.9/test/imghdrdata/python.jpg
evals	Retirar Python del runner	2026-09-15T04:51:06.4014921Z retirado: /opt/hostedtoolcache/PyPy/3.9.19/x64/lib/pypy3.9/test/imghdrdata/python.sgi
evals	Retirar Python del runner	2026-09-15T04:51:06.4153094Z retirado: /opt/hostedtoolcache/PyPy/3.9.19/x64/lib/pypy3.9/test/imghdrdata/python.tiff
evals	Retirar Python del runner	2026-09-15T04:51:06.4291289Z retirado: /opt/hostedtoolcache/PyPy/3.9.19/x64/lib/pypy3.9/test/imghdrdata/python.bmp
evals	Retirar Python del runner	2026-09-15T04:51:06.4429777Z retirado: /opt/hostedtoolcache/PyPy/3.9.19/x64/lib/pypy3.9/test/imghdrdata/python.gif
evals	Retirar Python del runner	2026-09-15T04:51:06.4567859Z retirado: /opt/hostedtoolcache/PyPy/3.9.19/x64/lib/pypy3.9/test/imghdrdata/python.webp
evals	Retirar Python del runner	2026-09-15T04:51:06.4706573Z retirado: /opt/hostedtoolcache/PyPy/3.9.19/x64/lib/pypy3.9/test/imghdrdata/python.png
evals	Retirar Python del runner	2026-09-15T04:51:06.4845662Z retirado: /opt/hostedtoolcache/PyPy/3.9.19/x64/lib/pypy3.9/test/imghdrdata/python.ppm
evals	Retirar Python del runner	2026-09-15T04:51:06.4984774Z retirado: /opt/hostedtoolcache/PyPy/3.9.19/x64/lib/pypy3.9/test/imghdrdata/python.pgm
evals	Retirar Python del runner	2026-09-15T04:51:06.5124788Z retirado: /opt/hostedtoolcache/PyPy/3.9.19/x64/lib/pypy3.9/test/imghdrdata/python.exr
evals	Retirar Python del runner	2026-09-15T04:51:06.5263496Z retirado: /opt/hostedtoolcache/PyPy/3.9.19/x64/lib/pypy3.9/test/imghdrdata/python.xbm
evals	Retirar Python del runner	2026-09-15T04:51:06.5402159Z retirado: /opt/hostedtoolcache/PyPy/3.9.19/x64/lib/pypy3.9/test/imghdrdata/python.pbm
evals	Retirar Python del runner	2026-09-15T04:51:06.5540729Z retirado: /opt/hostedtoolcache/PyPy/3.9.19/x64/lib/pypy3.9/test/imghdrdata/python.ras
evals	Retirar Python del runner	2026-09-15T04:51:06.5679835Z retirado: /opt/hostedtoolcache/PyPy/3.9.19/x64/lib/pypy3.9/test/pythoninfo.py
evals	Retirar Python del runner	2026-09-15T04:51:06.5817958Z retirado: /opt/hostedtoolcache/PyPy/3.9.19/x64/lib/PYPY_PORTABLE_DEPS.txt
evals	Retirar Python del runner	2026-09-15T04:51:06.5956408Z retirado: /opt/hostedtoolcache/PyPy/3.9.19/x64/include/pypy3.9/Python.h
evals	Retirar Python del runner	2026-09-15T04:51:06.6094843Z retirado: /opt/hostedtoolcache/PyPy/3.9.19/x64/include/pypy3.9/pythonrun.h
evals	Retirar Python del runner	2026-09-15T04:51:06.6233750Z retirado: /opt/hostedtoolcache/PyPy/3.9.19/x64/include/pypy3.9/pypy_macros.h
evals	Retirar Python del runner	2026-09-15T04:51:06.6375376Z retirado: /opt/hostedtoolcache/PyPy/3.9.19/x64/include/pypy3.9/pypy_marshal_decl.h
evals	Retirar Python del runner	2026-09-15T04:51:06.6514733Z retirado: /opt/hostedtoolcache/PyPy/3.9.19/x64/include/pypy3.9/pypy_decl.h
evals	Retirar Python del runner	2026-09-15T04:51:06.6652525Z retirado: /opt/hostedtoolcache/PyPy/3.9.19/x64/include/pypy3.9/pypy_structmember_decl.h
evals	Retirar Python del runner	2026-09-15T04:51:06.6791051Z retirado: /opt/hostedtoolcache/PyPy/3.9.19/x64/PYPY_VERSION
evals	Retirar Python del runner	2026-09-15T04:51:06.7063950Z retirado: /opt/hostedtoolcache/PyPy/3.9.19/x64
evals	Retirar Python del runner	2026-09-15T04:51:06.9128098Z retirado: /opt/hostedtoolcache/PyPy/3.11.15/x64/lib/PYPY_PORTABLE_DEPS.txt
evals	Retirar Python del runner	2026-09-15T04:51:06.9271607Z retirado: /opt/hostedtoolcache/PyPy/3.11.15/x64/lib/pypy3.11/site-packages/pip/_vendor/pygments/lexers/__pycache__/python.pypy311.pyc
evals	Retirar Python del runner	2026-09-15T04:51:06.9411927Z retirado: /opt/hostedtoolcache/PyPy/3.11.15/x64/lib/pypy3.11/site-packages/pip/_vendor/pygments/lexers/python.py
evals	Retirar Python del runner	2026-09-15T04:51:06.9553474Z retirado: /opt/hostedtoolcache/PyPy/3.11.15/x64/lib/pypy3.11/_cffi_ssl/_cffi_src/openssl/pypy_win32_extra.py
evals	Retirar Python del runner	2026-09-15T04:51:06.9693916Z retirado: /opt/hostedtoolcache/PyPy/3.11.15/x64/lib/pypy3.11/hpy/devel/include/hpy/forbid_python_h/Python.h
evals	Retirar Python del runner	2026-09-15T04:51:06.9833249Z retirado: /opt/hostedtoolcache/PyPy/3.11.15/x64/lib/pypy3.11/idlelib/Icons/python.gif
evals	Retirar Python del runner	2026-09-15T04:51:06.9973182Z retirado: /opt/hostedtoolcache/PyPy/3.11.15/x64/lib/pypy3.11/test/imghdrdata/python.jpg
evals	Retirar Python del runner	2026-09-15T04:51:07.0111328Z retirado: /opt/hostedtoolcache/PyPy/3.11.15/x64/lib/pypy3.11/test/imghdrdata/python.sgi
evals	Retirar Python del runner	2026-09-15T04:51:07.0249968Z retirado: /opt/hostedtoolcache/PyPy/3.11.15/x64/lib/pypy3.11/test/imghdrdata/python.tiff
evals	Retirar Python del runner	2026-09-15T04:51:07.0389952Z retirado: /opt/hostedtoolcache/PyPy/3.11.15/x64/lib/pypy3.11/test/imghdrdata/python.bmp
evals	Retirar Python del runner	2026-09-15T04:51:07.0530258Z retirado: /opt/hostedtoolcache/PyPy/3.11.15/x64/lib/pypy3.11/test/imghdrdata/python.gif
evals	Retirar Python del runner	2026-09-15T04:51:07.0669525Z retirado: /opt/hostedtoolcache/PyPy/3.11.15/x64/lib/pypy3.11/test/imghdrdata/python.webp
evals	Retirar Python del runner	2026-09-15T04:51:07.0808649Z retirado: /opt/hostedtoolcache/PyPy/3.11.15/x64/lib/pypy3.11/test/imghdrdata/python.png
evals	Retirar Python del runner	2026-09-15T04:51:07.0949637Z retirado: /opt/hostedtoolcache/PyPy/3.11.15/x64/lib/pypy3.11/test/imghdrdata/python.ppm
evals	Retirar Python del runner	2026-09-15T04:51:07.1090009Z retirado: /opt/hostedtoolcache/PyPy/3.11.15/x64/lib/pypy3.11/test/imghdrdata/python.pgm
evals	Retirar Python del runner	2026-09-15T04:51:07.1232033Z retirado: /opt/hostedtoolcache/PyPy/3.11.15/x64/lib/pypy3.11/test/imghdrdata/python.exr
evals	Retirar Python del runner	2026-09-15T04:51:07.1375801Z retirado: /opt/hostedtoolcache/PyPy/3.11.15/x64/lib/pypy3.11/test/imghdrdata/python.xbm
evals	Retirar Python del runner	2026-09-15T04:51:07.1530068Z retirado: /opt/hostedtoolcache/PyPy/3.11.15/x64/lib/pypy3.11/test/imghdrdata/python.pbm
evals	Retirar Python del runner	2026-09-15T04:51:07.1663318Z retirado: /opt/hostedtoolcache/PyPy/3.11.15/x64/lib/pypy3.11/test/imghdrdata/python.ras
evals	Retirar Python del runner	2026-09-15T04:51:07.1802291Z retirado: /opt/hostedtoolcache/PyPy/3.11.15/x64/lib/pypy3.11/test/imghdrdata/python-raw.jpg
evals	Retirar Python del runner	2026-09-15T04:51:07.1943727Z retirado: /opt/hostedtoolcache/PyPy/3.11.15/x64/lib/pypy3.11/test/test_email/data/python.jpg
evals	Retirar Python del runner	2026-09-15T04:51:07.2084011Z retirado: /opt/hostedtoolcache/PyPy/3.11.15/x64/lib/pypy3.11/test/test_email/data/python.sgi
evals	Retirar Python del runner	2026-09-15T04:51:07.2223426Z retirado: /opt/hostedtoolcache/PyPy/3.11.15/x64/lib/pypy3.11/test/test_email/data/python.tiff
evals	Retirar Python del runner	2026-09-15T04:51:07.2360552Z retirado: /opt/hostedtoolcache/PyPy/3.11.15/x64/lib/pypy3.11/test/test_email/data/python.bmp
evals	Retirar Python del runner	2026-09-15T04:51:07.2503480Z retirado: /opt/hostedtoolcache/PyPy/3.11.15/x64/lib/pypy3.11/test/test_email/data/python.gif
evals	Retirar Python del runner	2026-09-15T04:51:07.2643291Z retirado: /opt/hostedtoolcache/PyPy/3.11.15/x64/lib/pypy3.11/test/test_email/data/python.webp
evals	Retirar Python del runner	2026-09-15T04:51:07.2782198Z retirado: /opt/hostedtoolcache/PyPy/3.11.15/x64/lib/pypy3.11/test/test_email/data/python.png
evals	Retirar Python del runner	2026-09-15T04:51:07.2921010Z retirado: /opt/hostedtoolcache/PyPy/3.11.15/x64/lib/pypy3.11/test/test_email/data/python.ppm
evals	Retirar Python del runner	2026-09-15T04:51:07.3059740Z retirado: /opt/hostedtoolcache/PyPy/3.11.15/x64/lib/pypy3.11/test/test_email/data/python.pgm
evals	Retirar Python del runner	2026-09-15T04:51:07.3196601Z retirado: /opt/hostedtoolcache/PyPy/3.11.15/x64/lib/pypy3.11/test/test_email/data/python.exr
evals	Retirar Python del runner	2026-09-15T04:51:07.3335971Z retirado: /opt/hostedtoolcache/PyPy/3.11.15/x64/lib/pypy3.11/test/test_email/data/python.xbm
evals	Retirar Python del runner	2026-09-15T04:51:07.3475910Z retirado: /opt/hostedtoolcache/PyPy/3.11.15/x64/lib/pypy3.11/test/test_email/data/python.pbm
evals	Retirar Python del runner	2026-09-15T04:51:07.3615644Z retirado: /opt/hostedtoolcache/PyPy/3.11.15/x64/lib/pypy3.11/test/test_email/data/python.ras
evals	Retirar Python del runner	2026-09-15T04:51:07.3755736Z retirado: /opt/hostedtoolcache/PyPy/3.11.15/x64/lib/pypy3.11/test/pythoninfo.py
evals	Retirar Python del runner	2026-09-15T04:51:07.3896762Z retirado: /opt/hostedtoolcache/PyPy/3.11.15/x64/include/pypy3.11/Python.h
evals	Retirar Python del runner	2026-09-15T04:51:07.4034977Z retirado: /opt/hostedtoolcache/PyPy/3.11.15/x64/include/pypy3.11/pythonrun.h
evals	Retirar Python del runner	2026-09-15T04:51:07.4174949Z retirado: /opt/hostedtoolcache/PyPy/3.11.15/x64/include/pypy3.11/pypy_macros.h
evals	Retirar Python del runner	2026-09-15T04:51:07.4316017Z retirado: /opt/hostedtoolcache/PyPy/3.11.15/x64/include/pypy3.11/pypy_marshal_decl.h
evals	Retirar Python del runner	2026-09-15T04:51:07.4455503Z retirado: /opt/hostedtoolcache/PyPy/3.11.15/x64/include/pypy3.11/pypy_decl.h
evals	Retirar Python del runner	2026-09-15T04:51:07.4598714Z retirado: /opt/hostedtoolcache/PyPy/3.11.15/x64/include/pypy3.11/pypy_structmember_decl.h
evals	Retirar Python del runner	2026-09-15T04:51:07.4735570Z retirado: /opt/hostedtoolcache/PyPy/3.11.15/x64/PYPY_VERSION
evals	Retirar Python del runner	2026-09-15T04:51:07.5008606Z retirado: /opt/hostedtoolcache/PyPy/3.11.15/x64
evals	Retirar Python del runner	2026-09-15T04:51:07.7233857Z retirado: /opt/hostedtoolcache/PyPy/3.10.16/x64/lib/PYPY_PORTABLE_DEPS.txt
evals	Retirar Python del runner	2026-09-15T04:51:07.7371205Z retirado: /opt/hostedtoolcache/PyPy/3.10.16/x64/lib/pypy3.10/site-packages/pip/_vendor/pygments/lexers/__pycache__/python.pypy310.pyc
evals	Retirar Python del runner	2026-09-15T04:51:07.7507764Z retirado: /opt/hostedtoolcache/PyPy/3.10.16/x64/lib/pypy3.10/site-packages/pip/_vendor/pygments/lexers/python.py
evals	Retirar Python del runner	2026-09-15T04:51:07.7643124Z retirado: /opt/hostedtoolcache/PyPy/3.10.16/x64/lib/pypy3.10/_cffi_ssl/_cffi_src/openssl/pypy_win32_extra.py
evals	Retirar Python del runner	2026-09-15T04:51:07.7779612Z retirado: /opt/hostedtoolcache/PyPy/3.10.16/x64/lib/pypy3.10/hpy/devel/include/hpy/forbid_python_h/Python.h
evals	Retirar Python del runner	2026-09-15T04:51:07.7916051Z retirado: /opt/hostedtoolcache/PyPy/3.10.16/x64/lib/pypy3.10/idlelib/Icons/python.gif
evals	Retirar Python del runner	2026-09-15T04:51:07.8072017Z retirado: /opt/hostedtoolcache/PyPy/3.10.16/x64/lib/pypy3.10/test/imghdrdata/python.jpg
evals	Retirar Python del runner	2026-09-15T04:51:07.8210865Z retirado: /opt/hostedtoolcache/PyPy/3.10.16/x64/lib/pypy3.10/test/imghdrdata/python.sgi
evals	Retirar Python del runner	2026-09-15T04:51:07.8350683Z retirado: /opt/hostedtoolcache/PyPy/3.10.16/x64/lib/pypy3.10/test/imghdrdata/python.tiff
evals	Retirar Python del runner	2026-09-15T04:51:07.8492611Z retirado: /opt/hostedtoolcache/PyPy/3.10.16/x64/lib/pypy3.10/test/imghdrdata/python.bmp
evals	Retirar Python del runner	2026-09-15T04:51:07.8632111Z retirado: /opt/hostedtoolcache/PyPy/3.10.16/x64/lib/pypy3.10/test/imghdrdata/python.gif
evals	Retirar Python del runner	2026-09-15T04:51:07.8773922Z retirado: /opt/hostedtoolcache/PyPy/3.10.16/x64/lib/pypy3.10/test/imghdrdata/python.webp
evals	Retirar Python del runner	2026-09-15T04:51:07.8912289Z retirado: /opt/hostedtoolcache/PyPy/3.10.16/x64/lib/pypy3.10/test/imghdrdata/python.png
evals	Retirar Python del runner	2026-09-15T04:51:07.9072423Z retirado: /opt/hostedtoolcache/PyPy/3.10.16/x64/lib/pypy3.10/test/imghdrdata/python.ppm
evals	Retirar Python del runner	2026-09-15T04:51:07.9211376Z retirado: /opt/hostedtoolcache/PyPy/3.10.16/x64/lib/pypy3.10/test/imghdrdata/python.pgm
evals	Retirar Python del runner	2026-09-15T04:51:07.9349472Z retirado: /opt/hostedtoolcache/PyPy/3.10.16/x64/lib/pypy3.10/test/imghdrdata/python.exr
evals	Retirar Python del runner	2026-09-15T04:51:07.9507398Z retirado: /opt/hostedtoolcache/PyPy/3.10.16/x64/lib/pypy3.10/test/imghdrdata/python.xbm
evals	Retirar Python del runner	2026-09-15T04:51:07.9648620Z retirado: /opt/hostedtoolcache/PyPy/3.10.16/x64/lib/pypy3.10/test/imghdrdata/python.pbm
evals	Retirar Python del runner	2026-09-15T04:51:07.9787044Z retirado: /opt/hostedtoolcache/PyPy/3.10.16/x64/lib/pypy3.10/test/imghdrdata/python.ras
evals	Retirar Python del runner	2026-09-15T04:51:07.9928296Z retirado: /opt/hostedtoolcache/PyPy/3.10.16/x64/lib/pypy3.10/test/pythoninfo.py
evals	Retirar Python del runner	2026-09-15T04:51:08.0067971Z retirado: /opt/hostedtoolcache/PyPy/3.10.16/x64/include/pypy3.10/Python.h
evals	Retirar Python del runner	2026-09-15T04:51:08.0208099Z retirado: /opt/hostedtoolcache/PyPy/3.10.16/x64/include/pypy3.10/pythonrun.h
evals	Retirar Python del runner	2026-09-15T04:51:08.0346059Z retirado: /opt/hostedtoolcache/PyPy/3.10.16/x64/include/pypy3.10/pypy_macros.h
evals	Retirar Python del runner	2026-09-15T04:51:08.0489326Z retirado: /opt/hostedtoolcache/PyPy/3.10.16/x64/include/pypy3.10/pypy_marshal_decl.h
evals	Retirar Python del runner	2026-09-15T04:51:08.0628615Z retirado: /opt/hostedtoolcache/PyPy/3.10.16/x64/include/pypy3.10/pypy_decl.h
evals	Retirar Python del runner	2026-09-15T04:51:08.0767371Z retirado: /opt/hostedtoolcache/PyPy/3.10.16/x64/include/pypy3.10/pypy_structmember_decl.h
evals	Retirar Python del runner	2026-09-15T04:51:08.0909286Z retirado: /opt/hostedtoolcache/PyPy/3.10.16/x64/PYPY_VERSION
evals	Retirar Python del runner	2026-09-15T04:51:08.1181386Z retirado: /opt/hostedtoolcache/PyPy/3.10.16/x64
evals	Retirar Python del runner	2026-09-15T04:51:08.3327552Z retirado: /opt/hostedtoolcache/CodeQL/2.26.4/x64/codeql/Open-Source-Notices/python3.txt
evals	Retirar Python del runner	2026-09-15T04:51:08.3471657Z retirado: /opt/hostedtoolcache/CodeQL/2.26.4/x64/codeql/qlpacks/codeql/python-queries/1.8.9/experimental/semmle/python/libraries/PythonJose.qll
evals	Retirar Python del runner	2026-09-15T04:51:08.3617822Z retirado: /opt/hostedtoolcache/CodeQL/2.26.4/x64/codeql/qlpacks/codeql/python-queries/1.8.9/experimental/semmle/python/libraries/Python_JWT.qll
evals	Retirar Python del runner	2026-09-15T04:51:08.3757616Z retirado: /opt/hostedtoolcache/CodeQL/2.26.4/x64/codeql/qlpacks/codeql/python-queries/1.8.9/.codeql/libraries/codeql/python-all/7.2.4/python.qll
evals	Retirar Python del runner	2026-09-15T04:51:08.3898802Z retirado: /opt/hostedtoolcache/CodeQL/2.26.4/x64/codeql/qlpacks/codeql/python-queries/1.8.9/codeql-suites/python-code-quality-extended.qls
evals	Retirar Python del runner	2026-09-15T04:51:08.4038406Z retirado: /opt/hostedtoolcache/CodeQL/2.26.4/x64/codeql/qlpacks/codeql/python-queries/1.8.9/codeql-suites/python-security-experimental.qls
evals	Retirar Python del runner	2026-09-15T04:51:08.4179638Z retirado: /opt/hostedtoolcache/CodeQL/2.26.4/x64/codeql/qlpacks/codeql/python-queries/1.8.9/codeql-suites/python-code-scanning.qls
evals	Retirar Python del runner	2026-09-15T04:51:08.4317976Z retirado: /opt/hostedtoolcache/CodeQL/2.26.4/x64/codeql/qlpacks/codeql/python-queries/1.8.9/codeql-suites/python-lgtm.qls
evals	Retirar Python del runner	2026-09-15T04:51:08.4459203Z retirado: /opt/hostedtoolcache/CodeQL/2.26.4/x64/codeql/qlpacks/codeql/python-queries/1.8.9/codeql-suites/python-security-extended.qls
evals	Retirar Python del runner	2026-09-15T04:51:08.4601600Z retirado: /opt/hostedtoolcache/CodeQL/2.26.4/x64/codeql/qlpacks/codeql/python-queries/1.8.9/codeql-suites/python-lgtm-full.qls
evals	Retirar Python del runner	2026-09-15T04:51:08.4740183Z retirado: /opt/hostedtoolcache/CodeQL/2.26.4/x64/codeql/qlpacks/codeql/python-queries/1.8.9/codeql-suites/python-security-and-quality.qls
evals	Retirar Python del runner	2026-09-15T04:51:08.4881117Z retirado: /opt/hostedtoolcache/CodeQL/2.26.4/x64/codeql/qlpacks/codeql/python-queries/1.8.9/codeql-suites/python-code-quality.qls
evals	Retirar Python del runner	2026-09-15T04:51:08.5018510Z retirado: /opt/hostedtoolcache/CodeQL/2.26.4/x64/codeql/qlpacks/codeql/python-examples/0.0.0/.codeql/libraries/codeql/python-all/7.2.4/python.qll
evals	Retirar Python del runner	2026-09-15T04:51:08.5158175Z retirado: /opt/hostedtoolcache/CodeQL/2.26.4/x64/codeql/qlpacks/codeql/actions-all/0.6.0/ext/generated/composite-actions/python_mypy.model.yml
evals	Retirar Python del runner	2026-09-15T04:51:08.5303509Z retirado: /opt/hostedtoolcache/CodeQL/2.26.4/x64/codeql/qlpacks/codeql/actions-all/0.6.0/ext/generated/composite-actions/python-poetry_poetry.model.yml
evals	Retirar Python del runner	2026-09-15T04:51:08.5442080Z retirado: /opt/hostedtoolcache/CodeQL/2.26.4/x64/codeql/qlpacks/codeql/actions-all/0.6.0/ext/generated/reusable-workflows/python_cpython.model.yml
evals	Retirar Python del runner	2026-09-15T04:51:08.5583996Z retirado: /opt/hostedtoolcache/CodeQL/2.26.4/x64/codeql/qlpacks/codeql/python-all/7.2.4/python.qll
evals	Retirar Python del runner	2026-09-15T04:51:08.5723760Z retirado: /opt/hostedtoolcache/CodeQL/2.26.4/x64/codeql/qlpacks/codeql/actions-queries/0.6.34/.codeql/libraries/codeql/actions-all/0.6.0/ext/generated/composite-actions/python_mypy.model.yml
evals	Retirar Python del runner	2026-09-15T04:51:08.5865821Z retirado: /opt/hostedtoolcache/CodeQL/2.26.4/x64/codeql/qlpacks/codeql/actions-queries/0.6.34/.codeql/libraries/codeql/actions-all/0.6.0/ext/generated/composite-actions/python-poetry_poetry.model.yml
evals	Retirar Python del runner	2026-09-15T04:51:08.6008015Z retirado: /opt/hostedtoolcache/CodeQL/2.26.4/x64/codeql/qlpacks/codeql/actions-queries/0.6.34/.codeql/libraries/codeql/actions-all/0.6.0/ext/generated/reusable-workflows/python_cpython.model.yml
evals	Retirar Python del runner	2026-09-15T04:51:08.6147749Z retirado: /opt/hostedtoolcache/CodeQL/2.26.4/x64/codeql/python/tools/python3src.zip
evals	Retirar Python del runner	2026-09-15T04:51:08.6286912Z retirado: /opt/hostedtoolcache/CodeQL/2.26.4/x64/codeql/python/tools/python_setup.cmd
evals	Retirar Python del runner	2026-09-15T04:51:08.6428001Z retirado: /opt/hostedtoolcache/CodeQL/2.26.4/x64/codeql/python/tools/python_setup.sh
evals	Retirar Python del runner	2026-09-15T04:51:08.6570074Z retirado: /opt/hostedtoolcache/CodeQL/2.26.4/x64/codeql/python/tools/python_tracer.py
evals	Retirar Python del runner	2026-09-15T04:51:08.6707744Z retirado: /opt/hostedtoolcache/Python/3.14.7/x64/lib/libpython3.so
evals	Retirar Python del runner	2026-09-15T04:51:08.6846147Z retirado: /opt/hostedtoolcache/Python/3.14.7/x64/lib/pkgconfig/python-3.14-embed.pc
evals	Retirar Python del runner	2026-09-15T04:51:08.6986983Z retirado: /opt/hostedtoolcache/Python/3.14.7/x64/lib/pkgconfig/python3.pc
evals	Retirar Python del runner	2026-09-15T04:51:08.7183305Z retirado: /opt/hostedtoolcache/Python/3.14.7/x64/lib/pkgconfig/python3-embed.pc
evals	Retirar Python del runner	2026-09-15T04:51:08.7324296Z retirado: /opt/hostedtoolcache/Python/3.14.7/x64/lib/pkgconfig/python-3.14.pc
evals	Retirar Python del runner	2026-09-15T04:51:08.7463460Z retirado: /opt/hostedtoolcache/Python/3.14.7/x64/lib/python3.14/site-packages/pip/_vendor/pygments/lexers/__pycache__/python.cpython-314.pyc
evals	Retirar Python del runner	2026-09-15T04:51:08.7602970Z retirado: /opt/hostedtoolcache/Python/3.14.7/x64/lib/python3.14/site-packages/pip/_vendor/pygments/lexers/python.py
evals	Retirar Python del runner	2026-09-15T04:51:08.7742170Z retirado: /opt/hostedtoolcache/Python/3.14.7/x64/lib/python3.14/config-3.14-x86_64-linux-gnu/libpython3.14.a
evals	Retirar Python del runner	2026-09-15T04:51:08.7882674Z retirado: /opt/hostedtoolcache/Python/3.14.7/x64/lib/python3.14/config-3.14-x86_64-linux-gnu/python-config.py
evals	Retirar Python del runner	2026-09-15T04:51:08.8021807Z retirado: /opt/hostedtoolcache/Python/3.14.7/x64/lib/python3.14/config-3.14-x86_64-linux-gnu/python.o
evals	Retirar Python del runner	2026-09-15T04:51:08.8161239Z retirado: /opt/hostedtoolcache/Python/3.14.7/x64/lib/python3.14/idlelib/Icons/python.gif
evals	Retirar Python del runner	2026-09-15T04:51:08.8300271Z retirado: /opt/hostedtoolcache/Python/3.14.7/x64/lib/python3.14/test/__pycache__/pythoninfo.cpython-314.pyc
evals	Retirar Python del runner	2026-09-15T04:51:08.8439178Z retirado: /opt/hostedtoolcache/Python/3.14.7/x64/lib/python3.14/test/__pycache__/pythoninfo.cpython-314.opt-2.pyc
evals	Retirar Python del runner	2026-09-15T04:51:08.8580104Z retirado: /opt/hostedtoolcache/Python/3.14.7/x64/lib/python3.14/test/__pycache__/pythoninfo.cpython-314.opt-1.pyc
evals	Retirar Python del runner	2026-09-15T04:51:08.8718650Z retirado: /opt/hostedtoolcache/Python/3.14.7/x64/lib/python3.14/test/tkinterdata/python.gif
evals	Retirar Python del runner	2026-09-15T04:51:08.8858423Z retirado: /opt/hostedtoolcache/Python/3.14.7/x64/lib/python3.14/test/tkinterdata/python.png
evals	Retirar Python del runner	2026-09-15T04:51:08.8997913Z retirado: /opt/hostedtoolcache/Python/3.14.7/x64/lib/python3.14/test/tkinterdata/python.ppm
evals	Retirar Python del runner	2026-09-15T04:51:08.9138211Z retirado: /opt/hostedtoolcache/Python/3.14.7/x64/lib/python3.14/test/tkinterdata/python.pgm
evals	Retirar Python del runner	2026-09-15T04:51:08.9277441Z retirado: /opt/hostedtoolcache/Python/3.14.7/x64/lib/python3.14/test/tkinterdata/python.xbm
evals	Retirar Python del runner	2026-09-15T04:51:08.9418231Z retirado: /opt/hostedtoolcache/Python/3.14.7/x64/lib/python3.14/test/test_email/data/python.jpg
evals	Retirar Python del runner	2026-09-15T04:51:08.9560417Z retirado: /opt/hostedtoolcache/Python/3.14.7/x64/lib/python3.14/test/test_email/data/python.sgi
evals	Retirar Python del runner	2026-09-15T04:51:08.9701854Z retirado: /opt/hostedtoolcache/Python/3.14.7/x64/lib/python3.14/test/test_email/data/python.tiff
evals	Retirar Python del runner	2026-09-15T04:51:08.9841902Z retirado: /opt/hostedtoolcache/Python/3.14.7/x64/lib/python3.14/test/test_email/data/python.bmp
evals	Retirar Python del runner	2026-09-15T04:51:08.9981279Z retirado: /opt/hostedtoolcache/Python/3.14.7/x64/lib/python3.14/test/test_email/data/python.gif
evals	Retirar Python del runner	2026-09-15T04:51:09.0120407Z retirado: /opt/hostedtoolcache/Python/3.14.7/x64/lib/python3.14/test/test_email/data/python.webp
evals	Retirar Python del runner	2026-09-15T04:51:09.0259513Z retirado: /opt/hostedtoolcache/Python/3.14.7/x64/lib/python3.14/test/test_email/data/python.png
evals	Retirar Python del runner	2026-09-15T04:51:09.0398500Z retirado: /opt/hostedtoolcache/Python/3.14.7/x64/lib/python3.14/test/test_email/data/python.ppm
evals	Retirar Python del runner	2026-09-15T04:51:09.0538298Z retirado: /opt/hostedtoolcache/Python/3.14.7/x64/lib/python3.14/test/test_email/data/python.pgm
evals	Retirar Python del runner	2026-09-15T04:51:09.0679306Z retirado: /opt/hostedtoolcache/Python/3.14.7/x64/lib/python3.14/test/test_email/data/python.exr
evals	Retirar Python del runner	2026-09-15T04:51:09.0817678Z retirado: /opt/hostedtoolcache/Python/3.14.7/x64/lib/python3.14/test/test_email/data/python.xbm
evals	Retirar Python del runner	2026-09-15T04:51:09.0956823Z retirado: /opt/hostedtoolcache/Python/3.14.7/x64/lib/python3.14/test/test_email/data/python.pbm
evals	Retirar Python del runner	2026-09-15T04:51:09.1095450Z retirado: /opt/hostedtoolcache/Python/3.14.7/x64/lib/python3.14/test/test_email/data/python.ras
evals	Retirar Python del runner	2026-09-15T04:51:09.1235404Z retirado: /opt/hostedtoolcache/Python/3.14.7/x64/lib/python3.14/test/pythoninfo.py
evals	Retirar Python del runner	2026-09-15T04:51:09.1377167Z retirado: /opt/hostedtoolcache/Python/3.14.7/x64/lib/libpython3.14.so
evals	Retirar Python del runner	2026-09-15T04:51:09.1515058Z retirado: /opt/hostedtoolcache/Python/3.14.7/x64/lib/libpython3.14.so.1.0
evals	Retirar Python del runner	2026-09-15T04:51:09.1657476Z retirado: /opt/hostedtoolcache/Python/3.14.7/x64/include/python3.14/cpython/pythonrun.h
evals	Retirar Python del runner	2026-09-15T04:51:09.1802383Z retirado: /opt/hostedtoolcache/Python/3.14.7/x64/include/python3.14/Python.h
evals	Retirar Python del runner	2026-09-15T04:51:09.1943331Z retirado: /opt/hostedtoolcache/Python/3.14.7/x64/include/python3.14/pythonrun.h
evals	Retirar Python del runner	2026-09-15T04:51:09.2078814Z retirado: /opt/hostedtoolcache/Python/3.14.7/x64/share/man/man1/python3.1
evals	Retirar Python del runner	2026-09-15T04:51:09.2215767Z retirado: /opt/hostedtoolcache/Python/3.14.7/x64/share/man/man1/python3.14.1
evals	Retirar Python del runner	2026-09-15T04:51:09.2482198Z retirado: /opt/hostedtoolcache/Python/3.14.7/x64
evals	Retirar Python del runner	2026-09-15T04:51:09.5056563Z retirado: /opt/hostedtoolcache/Python/3.13.15/x64/lib/libpython3.so
evals	Retirar Python del runner	2026-09-15T04:51:09.5191767Z retirado: /opt/hostedtoolcache/Python/3.13.15/x64/lib/pkgconfig/python3.pc
evals	Retirar Python del runner	2026-09-15T04:51:09.5325666Z retirado: /opt/hostedtoolcache/Python/3.13.15/x64/lib/pkgconfig/python-3.13.pc
evals	Retirar Python del runner	2026-09-15T04:51:09.5460401Z retirado: /opt/hostedtoolcache/Python/3.13.15/x64/lib/pkgconfig/python3-embed.pc
evals	Retirar Python del runner	2026-09-15T04:51:09.5619885Z retirado: /opt/hostedtoolcache/Python/3.13.15/x64/lib/pkgconfig/python-3.13-embed.pc
evals	Retirar Python del runner	2026-09-15T04:51:09.5752896Z retirado: /opt/hostedtoolcache/Python/3.13.15/x64/lib/python3.13/site-packages/pip/_vendor/pygments/lexers/__pycache__/python.cpython-313.pyc
evals	Retirar Python del runner	2026-09-15T04:51:09.5889918Z retirado: /opt/hostedtoolcache/Python/3.13.15/x64/lib/python3.13/site-packages/pip/_vendor/pygments/lexers/python.py
evals	Retirar Python del runner	2026-09-15T04:51:09.6023232Z retirado: /opt/hostedtoolcache/Python/3.13.15/x64/lib/python3.13/idlelib/Icons/python.gif
evals	Retirar Python del runner	2026-09-15T04:51:09.6158439Z retirado: /opt/hostedtoolcache/Python/3.13.15/x64/lib/python3.13/test/__pycache__/pythoninfo.cpython-313.opt-2.pyc
evals	Retirar Python del runner	2026-09-15T04:51:09.6296870Z retirado: /opt/hostedtoolcache/Python/3.13.15/x64/lib/python3.13/test/__pycache__/pythoninfo.cpython-313.pyc
evals	Retirar Python del runner	2026-09-15T04:51:09.6435017Z retirado: /opt/hostedtoolcache/Python/3.13.15/x64/lib/python3.13/test/__pycache__/pythoninfo.cpython-313.opt-1.pyc
evals	Retirar Python del runner	2026-09-15T04:51:09.6571581Z retirado: /opt/hostedtoolcache/Python/3.13.15/x64/lib/python3.13/test/tkinterdata/python.gif
evals	Retirar Python del runner	2026-09-15T04:51:09.6708068Z retirado: /opt/hostedtoolcache/Python/3.13.15/x64/lib/python3.13/test/tkinterdata/python.png
evals	Retirar Python del runner	2026-09-15T04:51:09.6843875Z retirado: /opt/hostedtoolcache/Python/3.13.15/x64/lib/python3.13/test/tkinterdata/python.ppm
evals	Retirar Python del runner	2026-09-15T04:51:09.6979710Z retirado: /opt/hostedtoolcache/Python/3.13.15/x64/lib/python3.13/test/tkinterdata/python.pgm
evals	Retirar Python del runner	2026-09-15T04:51:09.7116256Z retirado: /opt/hostedtoolcache/Python/3.13.15/x64/lib/python3.13/test/tkinterdata/python.xbm
evals	Retirar Python del runner	2026-09-15T04:51:09.7255346Z retirado: /opt/hostedtoolcache/Python/3.13.15/x64/lib/python3.13/test/test_email/data/python.jpg
evals	Retirar Python del runner	2026-09-15T04:51:09.7395490Z retirado: /opt/hostedtoolcache/Python/3.13.15/x64/lib/python3.13/test/test_email/data/python.sgi
evals	Retirar Python del runner	2026-09-15T04:51:09.7561997Z retirado: /opt/hostedtoolcache/Python/3.13.15/x64/lib/python3.13/test/test_email/data/python.tiff
evals	Retirar Python del runner	2026-09-15T04:51:09.7700291Z retirado: /opt/hostedtoolcache/Python/3.13.15/x64/lib/python3.13/test/test_email/data/python.bmp
evals	Retirar Python del runner	2026-09-15T04:51:09.7841021Z retirado: /opt/hostedtoolcache/Python/3.13.15/x64/lib/python3.13/test/test_email/data/python.gif
evals	Retirar Python del runner	2026-09-15T04:51:09.7981195Z retirado: /opt/hostedtoolcache/Python/3.13.15/x64/lib/python3.13/test/test_email/data/python.webp
evals	Retirar Python del runner	2026-09-15T04:51:09.8121734Z retirado: /opt/hostedtoolcache/Python/3.13.15/x64/lib/python3.13/test/test_email/data/python.png
evals	Retirar Python del runner	2026-09-15T04:51:09.8259189Z retirado: /opt/hostedtoolcache/Python/3.13.15/x64/lib/python3.13/test/test_email/data/python.ppm
evals	Retirar Python del runner	2026-09-15T04:51:09.8398475Z retirado: /opt/hostedtoolcache/Python/3.13.15/x64/lib/python3.13/test/test_email/data/python.pgm
evals	Retirar Python del runner	2026-09-15T04:51:09.8538000Z retirado: /opt/hostedtoolcache/Python/3.13.15/x64/lib/python3.13/test/test_email/data/python.exr
evals	Retirar Python del runner	2026-09-15T04:51:09.8677191Z retirado: /opt/hostedtoolcache/Python/3.13.15/x64/lib/python3.13/test/test_email/data/python.xbm
evals	Retirar Python del runner	2026-09-15T04:51:09.8817166Z retirado: /opt/hostedtoolcache/Python/3.13.15/x64/lib/python3.13/test/test_email/data/python.pbm
evals	Retirar Python del runner	2026-09-15T04:51:09.8957439Z retirado: /opt/hostedtoolcache/Python/3.13.15/x64/lib/python3.13/test/test_email/data/python.ras
evals	Retirar Python del runner	2026-09-15T04:51:09.9096822Z retirado: /opt/hostedtoolcache/Python/3.13.15/x64/lib/python3.13/test/pythoninfo.py
evals	Retirar Python del runner	2026-09-15T04:51:09.9236391Z retirado: /opt/hostedtoolcache/Python/3.13.15/x64/lib/python3.13/config-3.13-x86_64-linux-gnu/libpython3.13.a
evals	Retirar Python del runner	2026-09-15T04:51:09.9375498Z retirado: /opt/hostedtoolcache/Python/3.13.15/x64/lib/python3.13/config-3.13-x86_64-linux-gnu/python-config.py
evals	Retirar Python del runner	2026-09-15T04:51:09.9515686Z retirado: /opt/hostedtoolcache/Python/3.13.15/x64/lib/python3.13/config-3.13-x86_64-linux-gnu/python.o
evals	Retirar Python del runner	2026-09-15T04:51:09.9654324Z retirado: /opt/hostedtoolcache/Python/3.13.15/x64/lib/libpython3.13.so
evals	Retirar Python del runner	2026-09-15T04:51:09.9794184Z retirado: /opt/hostedtoolcache/Python/3.13.15/x64/lib/libpython3.13.so.1.0
evals	Retirar Python del runner	2026-09-15T04:51:09.9933551Z retirado: /opt/hostedtoolcache/Python/3.13.15/x64/include/python3.13/cpython/pythonrun.h
evals	Retirar Python del runner	2026-09-15T04:51:10.0074441Z retirado: /opt/hostedtoolcache/Python/3.13.15/x64/include/python3.13/Python.h
evals	Retirar Python del runner	2026-09-15T04:51:10.0220012Z retirado: /opt/hostedtoolcache/Python/3.13.15/x64/include/python3.13/pythonrun.h
evals	Retirar Python del runner	2026-09-15T04:51:10.0349258Z retirado: /opt/hostedtoolcache/Python/3.13.15/x64/share/man/man1/python3.1
evals	Retirar Python del runner	2026-09-15T04:51:10.0486313Z retirado: /opt/hostedtoolcache/Python/3.13.15/x64/share/man/man1/python3.13.1
evals	Retirar Python del runner	2026-09-15T04:51:10.0755760Z retirado: /opt/hostedtoolcache/Python/3.13.15/x64
evals	Retirar Python del runner	2026-09-15T04:51:10.3389740Z retirado: /opt/hostedtoolcache/Python/3.10.21/x64/lib/libpython3.so
evals	Retirar Python del runner	2026-09-15T04:51:10.3531321Z retirado: /opt/hostedtoolcache/Python/3.10.21/x64/lib/python3.10/site-packages/pip/_vendor/pygments/lexers/__pycache__/python.cpython-310.pyc
evals	Retirar Python del runner	2026-09-15T04:51:10.3669606Z retirado: /opt/hostedtoolcache/Python/3.10.21/x64/lib/python3.10/site-packages/pip/_vendor/pygments/lexers/python.py
evals	Retirar Python del runner	2026-09-15T04:51:10.3811015Z retirado: /opt/hostedtoolcache/Python/3.10.21/x64/lib/python3.10/config-3.10-x86_64-linux-gnu/libpython3.10.a
evals	Retirar Python del runner	2026-09-15T04:51:10.3950847Z retirado: /opt/hostedtoolcache/Python/3.10.21/x64/lib/python3.10/config-3.10-x86_64-linux-gnu/python-config.py
evals	Retirar Python del runner	2026-09-15T04:51:10.4092233Z retirado: /opt/hostedtoolcache/Python/3.10.21/x64/lib/python3.10/config-3.10-x86_64-linux-gnu/python.o
evals	Retirar Python del runner	2026-09-15T04:51:10.4230577Z retirado: /opt/hostedtoolcache/Python/3.10.21/x64/lib/python3.10/idlelib/Icons/python.gif
evals	Retirar Python del runner	2026-09-15T04:51:10.4370200Z retirado: /opt/hostedtoolcache/Python/3.10.21/x64/lib/python3.10/test/__pycache__/pythoninfo.cpython-310.pyc
evals	Retirar Python del runner	2026-09-15T04:51:10.4513095Z retirado: /opt/hostedtoolcache/Python/3.10.21/x64/lib/python3.10/test/__pycache__/pythoninfo.cpython-310.opt-2.pyc
evals	Retirar Python del runner	2026-09-15T04:51:10.4652016Z retirado: /opt/hostedtoolcache/Python/3.10.21/x64/lib/python3.10/test/__pycache__/pythoninfo.cpython-310.opt-1.pyc
evals	Retirar Python del runner	2026-09-15T04:51:10.4793440Z retirado: /opt/hostedtoolcache/Python/3.10.21/x64/lib/python3.10/test/imghdrdata/python.jpg
evals	Retirar Python del runner	2026-09-15T04:51:10.4933674Z retirado: /opt/hostedtoolcache/Python/3.10.21/x64/lib/python3.10/test/imghdrdata/python.sgi
evals	Retirar Python del runner	2026-09-15T04:51:10.5072336Z retirado: /opt/hostedtoolcache/Python/3.10.21/x64/lib/python3.10/test/imghdrdata/python.tiff
evals	Retirar Python del runner	2026-09-15T04:51:10.5210646Z retirado: /opt/hostedtoolcache/Python/3.10.21/x64/lib/python3.10/test/imghdrdata/python.bmp
evals	Retirar Python del runner	2026-09-15T04:51:10.5350352Z retirado: /opt/hostedtoolcache/Python/3.10.21/x64/lib/python3.10/test/imghdrdata/python.gif
evals	Retirar Python del runner	2026-09-15T04:51:10.5489069Z retirado: /opt/hostedtoolcache/Python/3.10.21/x64/lib/python3.10/test/imghdrdata/python.webp
evals	Retirar Python del runner	2026-09-15T04:51:10.5628592Z retirado: /opt/hostedtoolcache/Python/3.10.21/x64/lib/python3.10/test/imghdrdata/python.png
evals	Retirar Python del runner	2026-09-15T04:51:10.5767329Z retirado: /opt/hostedtoolcache/Python/3.10.21/x64/lib/python3.10/test/imghdrdata/python.ppm
evals	Retirar Python del runner	2026-09-15T04:51:10.5908071Z retirado: /opt/hostedtoolcache/Python/3.10.21/x64/lib/python3.10/test/imghdrdata/python.pgm
evals	Retirar Python del runner	2026-09-15T04:51:10.6046459Z retirado: /opt/hostedtoolcache/Python/3.10.21/x64/lib/python3.10/test/imghdrdata/python.exr
evals	Retirar Python del runner	2026-09-15T04:51:10.6189483Z retirado: /opt/hostedtoolcache/Python/3.10.21/x64/lib/python3.10/test/imghdrdata/python.xbm
evals	Retirar Python del runner	2026-09-15T04:51:10.6331189Z retirado: /opt/hostedtoolcache/Python/3.10.21/x64/lib/python3.10/test/imghdrdata/python.pbm
evals	Retirar Python del runner	2026-09-15T04:51:10.6471092Z retirado: /opt/hostedtoolcache/Python/3.10.21/x64/lib/python3.10/test/imghdrdata/python.ras
evals	Retirar Python del runner	2026-09-15T04:51:10.6611315Z retirado: /opt/hostedtoolcache/Python/3.10.21/x64/lib/python3.10/test/pythoninfo.py
evals	Retirar Python del runner	2026-09-15T04:51:10.6750959Z retirado: /opt/hostedtoolcache/Python/3.10.21/x64/lib/pkgconfig/python3.pc
evals	Retirar Python del runner	2026-09-15T04:51:10.6890829Z retirado: /opt/hostedtoolcache/Python/3.10.21/x64/lib/pkgconfig/python3-embed.pc
evals	Retirar Python del runner	2026-09-15T04:51:10.7029180Z retirado: /opt/hostedtoolcache/Python/3.10.21/x64/lib/pkgconfig/python-3.10.pc
evals	Retirar Python del runner	2026-09-15T04:51:10.7169510Z retirado: /opt/hostedtoolcache/Python/3.10.21/x64/lib/pkgconfig/python-3.10-embed.pc
evals	Retirar Python del runner	2026-09-15T04:51:10.7308711Z retirado: /opt/hostedtoolcache/Python/3.10.21/x64/lib/libpython3.10.so
evals	Retirar Python del runner	2026-09-15T04:51:10.7448175Z retirado: /opt/hostedtoolcache/Python/3.10.21/x64/lib/libpython3.10.so.1.0
evals	Retirar Python del runner	2026-09-15T04:51:10.7586012Z retirado: /opt/hostedtoolcache/Python/3.10.21/x64/include/python3.10/cpython/pythonrun.h
evals	Retirar Python del runner	2026-09-15T04:51:10.7726657Z retirado: /opt/hostedtoolcache/Python/3.10.21/x64/include/python3.10/Python.h
evals	Retirar Python del runner	2026-09-15T04:51:10.7864890Z retirado: /opt/hostedtoolcache/Python/3.10.21/x64/include/python3.10/pythonrun.h
evals	Retirar Python del runner	2026-09-15T04:51:10.8005194Z retirado: /opt/hostedtoolcache/Python/3.10.21/x64/share/man/man1/python3.10.1
evals	Retirar Python del runner	2026-09-15T04:51:10.8202854Z retirado: /opt/hostedtoolcache/Python/3.10.21/x64/share/man/man1/python3.1
evals	Retirar Python del runner	2026-09-15T04:51:10.8477354Z retirado: /opt/hostedtoolcache/Python/3.10.21/x64
evals	Retirar Python del runner	2026-09-15T04:51:11.3015831Z retirado: /opt/hostedtoolcache/Python/3.12.14/x64/lib/libpython3.so
evals	Retirar Python del runner	2026-09-15T04:51:11.3166622Z retirado: /opt/hostedtoolcache/Python/3.12.14/x64/lib/pkgconfig/python-3.12.pc
evals	Retirar Python del runner	2026-09-15T04:51:11.3384349Z retirado: /opt/hostedtoolcache/Python/3.12.14/x64/lib/pkgconfig/python3.pc
evals	Retirar Python del runner	2026-09-15T04:51:11.3537234Z retirado: /opt/hostedtoolcache/Python/3.12.14/x64/lib/pkgconfig/python-3.12-embed.pc
evals	Retirar Python del runner	2026-09-15T04:51:11.3754485Z retirado: /opt/hostedtoolcache/Python/3.12.14/x64/lib/pkgconfig/python3-embed.pc
evals	Retirar Python del runner	2026-09-15T04:51:11.3907557Z retirado: /opt/hostedtoolcache/Python/3.12.14/x64/lib/libpython3.12.so
evals	Retirar Python del runner	2026-09-15T04:51:11.4063254Z retirado: /opt/hostedtoolcache/Python/3.12.14/x64/lib/python3.12/site-packages/pip/_vendor/pygments/lexers/__pycache__/python.cpython-312.pyc
evals	Retirar Python del runner	2026-09-15T04:51:11.4214013Z retirado: /opt/hostedtoolcache/Python/3.12.14/x64/lib/python3.12/site-packages/pip/_vendor/pygments/lexers/python.py
evals	Retirar Python del runner	2026-09-15T04:51:11.4363432Z retirado: /opt/hostedtoolcache/Python/3.12.14/x64/lib/python3.12/idlelib/Icons/python.gif
evals	Retirar Python del runner	2026-09-15T04:51:11.4515754Z retirado: /opt/hostedtoolcache/Python/3.12.14/x64/lib/python3.12/test/__pycache__/pythoninfo.cpython-312.opt-2.pyc
evals	Retirar Python del runner	2026-09-15T04:51:11.4667782Z retirado: /opt/hostedtoolcache/Python/3.12.14/x64/lib/python3.12/test/__pycache__/pythoninfo.cpython-312.opt-1.pyc
evals	Retirar Python del runner	2026-09-15T04:51:11.4821157Z retirado: /opt/hostedtoolcache/Python/3.12.14/x64/lib/python3.12/test/__pycache__/pythoninfo.cpython-312.pyc
evals	Retirar Python del runner	2026-09-15T04:51:11.4973204Z retirado: /opt/hostedtoolcache/Python/3.12.14/x64/lib/python3.12/test/imghdrdata/python.jpg
evals	Retirar Python del runner	2026-09-15T04:51:11.5123661Z retirado: /opt/hostedtoolcache/Python/3.12.14/x64/lib/python3.12/test/imghdrdata/python.sgi
evals	Retirar Python del runner	2026-09-15T04:51:11.5274114Z retirado: /opt/hostedtoolcache/Python/3.12.14/x64/lib/python3.12/test/imghdrdata/python.tiff
evals	Retirar Python del runner	2026-09-15T04:51:11.5425967Z retirado: /opt/hostedtoolcache/Python/3.12.14/x64/lib/python3.12/test/imghdrdata/python.bmp
evals	Retirar Python del runner	2026-09-15T04:51:11.5576828Z retirado: /opt/hostedtoolcache/Python/3.12.14/x64/lib/python3.12/test/imghdrdata/python.gif
evals	Retirar Python del runner	2026-09-15T04:51:11.5726239Z retirado: /opt/hostedtoolcache/Python/3.12.14/x64/lib/python3.12/test/imghdrdata/python.webp
evals	Retirar Python del runner	2026-09-15T04:51:11.5880497Z retirado: /opt/hostedtoolcache/Python/3.12.14/x64/lib/python3.12/test/imghdrdata/python.png
evals	Retirar Python del runner	2026-09-15T04:51:11.6032937Z retirado: /opt/hostedtoolcache/Python/3.12.14/x64/lib/python3.12/test/imghdrdata/python.ppm
evals	Retirar Python del runner	2026-09-15T04:51:11.6183984Z retirado: /opt/hostedtoolcache/Python/3.12.14/x64/lib/python3.12/test/imghdrdata/python.pgm
evals	Retirar Python del runner	2026-09-15T04:51:11.6339521Z retirado: /opt/hostedtoolcache/Python/3.12.14/x64/lib/python3.12/test/imghdrdata/python.exr
evals	Retirar Python del runner	2026-09-15T04:51:11.6492444Z retirado: /opt/hostedtoolcache/Python/3.12.14/x64/lib/python3.12/test/imghdrdata/python.xbm
evals	Retirar Python del runner	2026-09-15T04:51:11.6645274Z retirado: /opt/hostedtoolcache/Python/3.12.14/x64/lib/python3.12/test/imghdrdata/python.pbm
evals	Retirar Python del runner	2026-09-15T04:51:11.6797300Z retirado: /opt/hostedtoolcache/Python/3.12.14/x64/lib/python3.12/test/imghdrdata/python.ras
evals	Retirar Python del runner	2026-09-15T04:51:11.6949728Z retirado: /opt/hostedtoolcache/Python/3.12.14/x64/lib/python3.12/test/imghdrdata/python-raw.jpg
evals	Retirar Python del runner	2026-09-15T04:51:11.7102389Z retirado: /opt/hostedtoolcache/Python/3.12.14/x64/lib/python3.12/test/test_email/data/python.jpg
evals	Retirar Python del runner	2026-09-15T04:51:11.7255653Z retirado: /opt/hostedtoolcache/Python/3.12.14/x64/lib/python3.12/test/test_email/data/python.sgi
evals	Retirar Python del runner	2026-09-15T04:51:11.7407122Z retirado: /opt/hostedtoolcache/Python/3.12.14/x64/lib/python3.12/test/test_email/data/python.tiff
evals	Retirar Python del runner	2026-09-15T04:51:11.7560772Z retirado: /opt/hostedtoolcache/Python/3.12.14/x64/lib/python3.12/test/test_email/data/python.bmp
evals	Retirar Python del runner	2026-09-15T04:51:11.7713983Z retirado: /opt/hostedtoolcache/Python/3.12.14/x64/lib/python3.12/test/test_email/data/python.gif
evals	Retirar Python del runner	2026-09-15T04:51:11.7863582Z retirado: /opt/hostedtoolcache/Python/3.12.14/x64/lib/python3.12/test/test_email/data/python.webp
evals	Retirar Python del runner	2026-09-15T04:51:11.8014982Z retirado: /opt/hostedtoolcache/Python/3.12.14/x64/lib/python3.12/test/test_email/data/python.png
evals	Retirar Python del runner	2026-09-15T04:51:11.8166362Z retirado: /opt/hostedtoolcache/Python/3.12.14/x64/lib/python3.12/test/test_email/data/python.ppm
evals	Retirar Python del runner	2026-09-15T04:51:11.8318774Z retirado: /opt/hostedtoolcache/Python/3.12.14/x64/lib/python3.12/test/test_email/data/python.pgm
evals	Retirar Python del runner	2026-09-15T04:51:11.8470480Z retirado: /opt/hostedtoolcache/Python/3.12.14/x64/lib/python3.12/test/test_email/data/python.exr
evals	Retirar Python del runner	2026-09-15T04:51:11.8621905Z retirado: /opt/hostedtoolcache/Python/3.12.14/x64/lib/python3.12/test/test_email/data/python.xbm
evals	Retirar Python del runner	2026-09-15T04:51:11.8775484Z retirado: /opt/hostedtoolcache/Python/3.12.14/x64/lib/python3.12/test/test_email/data/python.pbm
evals	Retirar Python del runner	2026-09-15T04:51:11.8927972Z retirado: /opt/hostedtoolcache/Python/3.12.14/x64/lib/python3.12/test/test_email/data/python.ras
evals	Retirar Python del runner	2026-09-15T04:51:11.9080568Z retirado: /opt/hostedtoolcache/Python/3.12.14/x64/lib/python3.12/test/pythoninfo.py
evals	Retirar Python del runner	2026-09-15T04:51:11.9243300Z retirado: /opt/hostedtoolcache/Python/3.12.14/x64/lib/python3.12/config-3.12-x86_64-linux-gnu/libpython3.12.a
evals	Retirar Python del runner	2026-09-15T04:51:11.9396820Z retirado: /opt/hostedtoolcache/Python/3.12.14/x64/lib/python3.12/config-3.12-x86_64-linux-gnu/python-config.py
evals	Retirar Python del runner	2026-09-15T04:51:11.9549465Z retirado: /opt/hostedtoolcache/Python/3.12.14/x64/lib/python3.12/config-3.12-x86_64-linux-gnu/python.o
evals	Retirar Python del runner	2026-09-15T04:51:11.9701806Z retirado: /opt/hostedtoolcache/Python/3.12.14/x64/lib/libpython3.12.so.1.0
evals	Retirar Python del runner	2026-09-15T04:51:11.9854579Z retirado: /opt/hostedtoolcache/Python/3.12.14/x64/include/python3.12/cpython/pythonrun.h
evals	Retirar Python del runner	2026-09-15T04:51:12.0007978Z retirado: /opt/hostedtoolcache/Python/3.12.14/x64/include/python3.12/Python.h
evals	Retirar Python del runner	2026-09-15T04:51:12.0170154Z retirado: /opt/hostedtoolcache/Python/3.12.14/x64/include/python3.12/pythonrun.h
evals	Retirar Python del runner	2026-09-15T04:51:12.0316723Z retirado: /opt/hostedtoolcache/Python/3.12.14/x64/share/man/man1/python3.1
evals	Retirar Python del runner	2026-09-15T04:51:12.0468459Z retirado: /opt/hostedtoolcache/Python/3.12.14/x64/share/man/man1/python3.12.1
evals	Retirar Python del runner	2026-09-15T04:51:12.0769793Z retirado: /opt/hostedtoolcache/Python/3.12.14/x64
evals	Retirar Python del runner	2026-09-15T04:51:12.3672663Z retirado: /opt/hostedtoolcache/Python/3.11.16/x64/lib/libpython3.so
evals	Retirar Python del runner	2026-09-15T04:51:12.3825077Z retirado: /opt/hostedtoolcache/Python/3.11.16/x64/lib/pkgconfig/python3.pc
evals	Retirar Python del runner	2026-09-15T04:51:12.3978509Z retirado: /opt/hostedtoolcache/Python/3.11.16/x64/lib/pkgconfig/python3-embed.pc
evals	Retirar Python del runner	2026-09-15T04:51:12.4131785Z retirado: /opt/hostedtoolcache/Python/3.11.16/x64/lib/pkgconfig/python-3.11-embed.pc
evals	Retirar Python del runner	2026-09-15T04:51:12.4282074Z retirado: /opt/hostedtoolcache/Python/3.11.16/x64/lib/pkgconfig/python-3.11.pc
evals	Retirar Python del runner	2026-09-15T04:51:12.4433688Z retirado: /opt/hostedtoolcache/Python/3.11.16/x64/lib/python3.11/site-packages/pip/_vendor/pygments/lexers/__pycache__/python.cpython-311.pyc
evals	Retirar Python del runner	2026-09-15T04:51:12.4584366Z retirado: /opt/hostedtoolcache/Python/3.11.16/x64/lib/python3.11/site-packages/pip/_vendor/pygments/lexers/python.py
evals	Retirar Python del runner	2026-09-15T04:51:12.4736322Z retirado: /opt/hostedtoolcache/Python/3.11.16/x64/lib/python3.11/config-3.11-x86_64-linux-gnu/libpython3.11.a
evals	Retirar Python del runner	2026-09-15T04:51:12.4888763Z retirado: /opt/hostedtoolcache/Python/3.11.16/x64/lib/python3.11/config-3.11-x86_64-linux-gnu/python-config.py
evals	Retirar Python del runner	2026-09-15T04:51:12.5041080Z retirado: /opt/hostedtoolcache/Python/3.11.16/x64/lib/python3.11/config-3.11-x86_64-linux-gnu/python.o
evals	Retirar Python del runner	2026-09-15T04:51:12.5194725Z retirado: /opt/hostedtoolcache/Python/3.11.16/x64/lib/python3.11/idlelib/Icons/python.gif
evals	Retirar Python del runner	2026-09-15T04:51:12.5347370Z retirado: /opt/hostedtoolcache/Python/3.11.16/x64/lib/python3.11/test/__pycache__/pythoninfo.cpython-311.pyc
evals	Retirar Python del runner	2026-09-15T04:51:12.5499435Z retirado: /opt/hostedtoolcache/Python/3.11.16/x64/lib/python3.11/test/__pycache__/pythoninfo.cpython-311.opt-1.pyc
evals	Retirar Python del runner	2026-09-15T04:51:12.5651514Z retirado: /opt/hostedtoolcache/Python/3.11.16/x64/lib/python3.11/test/__pycache__/pythoninfo.cpython-311.opt-2.pyc
evals	Retirar Python del runner	2026-09-15T04:51:12.5804810Z retirado: /opt/hostedtoolcache/Python/3.11.16/x64/lib/python3.11/test/imghdrdata/python.jpg
evals	Retirar Python del runner	2026-09-15T04:51:12.5955579Z retirado: /opt/hostedtoolcache/Python/3.11.16/x64/lib/python3.11/test/imghdrdata/python.sgi
evals	Retirar Python del runner	2026-09-15T04:51:12.6109384Z retirado: /opt/hostedtoolcache/Python/3.11.16/x64/lib/python3.11/test/imghdrdata/python.tiff
evals	Retirar Python del runner	2026-09-15T04:51:12.6262202Z retirado: /opt/hostedtoolcache/Python/3.11.16/x64/lib/python3.11/test/imghdrdata/python.bmp
evals	Retirar Python del runner	2026-09-15T04:51:12.6427286Z retirado: /opt/hostedtoolcache/Python/3.11.16/x64/lib/python3.11/test/imghdrdata/python.gif
evals	Retirar Python del runner	2026-09-15T04:51:12.6579841Z retirado: /opt/hostedtoolcache/Python/3.11.16/x64/lib/python3.11/test/imghdrdata/python.webp
evals	Retirar Python del runner	2026-09-15T04:51:12.6733319Z retirado: /opt/hostedtoolcache/Python/3.11.16/x64/lib/python3.11/test/imghdrdata/python.png
evals	Retirar Python del runner	2026-09-15T04:51:12.6885699Z retirado: /opt/hostedtoolcache/Python/3.11.16/x64/lib/python3.11/test/imghdrdata/python.ppm
evals	Retirar Python del runner	2026-09-15T04:51:12.7039896Z retirado: /opt/hostedtoolcache/Python/3.11.16/x64/lib/python3.11/test/imghdrdata/python.pgm
evals	Retirar Python del runner	2026-09-15T04:51:12.7191545Z retirado: /opt/hostedtoolcache/Python/3.11.16/x64/lib/python3.11/test/imghdrdata/python.exr
evals	Retirar Python del runner	2026-09-15T04:51:12.7342800Z retirado: /opt/hostedtoolcache/Python/3.11.16/x64/lib/python3.11/test/imghdrdata/python.xbm
evals	Retirar Python del runner	2026-09-15T04:51:12.7496049Z retirado: /opt/hostedtoolcache/Python/3.11.16/x64/lib/python3.11/test/imghdrdata/python.pbm
evals	Retirar Python del runner	2026-09-15T04:51:12.7649636Z retirado: /opt/hostedtoolcache/Python/3.11.16/x64/lib/python3.11/test/imghdrdata/python.ras
evals	Retirar Python del runner	2026-09-15T04:51:12.7793916Z retirado: /opt/hostedtoolcache/Python/3.11.16/x64/lib/python3.11/test/imghdrdata/python-raw.jpg
evals	Retirar Python del runner	2026-09-15T04:51:12.7934372Z retirado: /opt/hostedtoolcache/Python/3.11.16/x64/lib/python3.11/test/test_email/data/python.jpg
evals	Retirar Python del runner	2026-09-15T04:51:12.8072616Z retirado: /opt/hostedtoolcache/Python/3.11.16/x64/lib/python3.11/test/test_email/data/python.sgi
evals	Retirar Python del runner	2026-09-15T04:51:12.8211883Z retirado: /opt/hostedtoolcache/Python/3.11.16/x64/lib/python3.11/test/test_email/data/python.tiff
evals	Retirar Python del runner	2026-09-15T04:51:12.8350017Z retirado: /opt/hostedtoolcache/Python/3.11.16/x64/lib/python3.11/test/test_email/data/python.bmp
evals	Retirar Python del runner	2026-09-15T04:51:12.8488559Z retirado: /opt/hostedtoolcache/Python/3.11.16/x64/lib/python3.11/test/test_email/data/python.gif
evals	Retirar Python del runner	2026-09-15T04:51:12.8630735Z retirado: /opt/hostedtoolcache/Python/3.11.16/x64/lib/python3.11/test/test_email/data/python.webp
evals	Retirar Python del runner	2026-09-15T04:51:12.8768688Z retirado: /opt/hostedtoolcache/Python/3.11.16/x64/lib/python3.11/test/test_email/data/python.png
evals	Retirar Python del runner	2026-09-15T04:51:12.8909045Z retirado: /opt/hostedtoolcache/Python/3.11.16/x64/lib/python3.11/test/test_email/data/python.ppm
evals	Retirar Python del runner	2026-09-15T04:51:12.9051637Z retirado: /opt/hostedtoolcache/Python/3.11.16/x64/lib/python3.11/test/test_email/data/python.pgm
evals	Retirar Python del runner	2026-09-15T04:51:12.9191586Z retirado: /opt/hostedtoolcache/Python/3.11.16/x64/lib/python3.11/test/test_email/data/python.exr
evals	Retirar Python del runner	2026-09-15T04:51:12.9329451Z retirado: /opt/hostedtoolcache/Python/3.11.16/x64/lib/python3.11/test/test_email/data/python.xbm
evals	Retirar Python del runner	2026-09-15T04:51:12.9471640Z retirado: /opt/hostedtoolcache/Python/3.11.16/x64/lib/python3.11/test/test_email/data/python.pbm
evals	Retirar Python del runner	2026-09-15T04:51:12.9609989Z retirado: /opt/hostedtoolcache/Python/3.11.16/x64/lib/python3.11/test/test_email/data/python.ras
evals	Retirar Python del runner	2026-09-15T04:51:12.9755797Z retirado: /opt/hostedtoolcache/Python/3.11.16/x64/lib/python3.11/test/pythoninfo.py
evals	Retirar Python del runner	2026-09-15T04:51:12.9896260Z retirado: /opt/hostedtoolcache/Python/3.11.16/x64/lib/libpython3.11.so
evals	Retirar Python del runner	2026-09-15T04:51:13.0034734Z retirado: /opt/hostedtoolcache/Python/3.11.16/x64/lib/libpython3.11.so.1.0
evals	Retirar Python del runner	2026-09-15T04:51:13.0173942Z retirado: /opt/hostedtoolcache/Python/3.11.16/x64/include/python3.11/cpython/pythonrun.h
evals	Retirar Python del runner	2026-09-15T04:51:13.0314012Z retirado: /opt/hostedtoolcache/Python/3.11.16/x64/include/python3.11/Python.h
evals	Retirar Python del runner	2026-09-15T04:51:13.0455863Z retirado: /opt/hostedtoolcache/Python/3.11.16/x64/include/python3.11/pythonrun.h
evals	Retirar Python del runner	2026-09-15T04:51:13.0596334Z retirado: /opt/hostedtoolcache/Python/3.11.16/x64/share/man/man1/python3.1
evals	Retirar Python del runner	2026-09-15T04:51:13.0736390Z retirado: /opt/hostedtoolcache/Python/3.11.16/x64/share/man/man1/python3.11.1
evals	Retirar Python del runner	2026-09-15T04:51:13.1010261Z retirado: /opt/hostedtoolcache/Python/3.11.16/x64
evals	Retirar Python del runner	2026-09-15T04:51:13.3794615Z retirado: /opt/az/lib/pkgconfig/python-3.14-embed.pc
evals	Retirar Python del runner	2026-09-15T04:51:13.3934283Z retirado: /opt/az/lib/pkgconfig/python3.pc
evals	Retirar Python del runner	2026-09-15T04:51:13.4130615Z retirado: /opt/az/lib/pkgconfig/python3-embed.pc
evals	Retirar Python del runner	2026-09-15T04:51:13.4269854Z retirado: /opt/az/lib/pkgconfig/python-3.14.pc
evals	Retirar Python del runner	2026-09-15T04:51:13.4410205Z retirado: /opt/az/lib/python3.14/site-packages/pygments/lexers/__pycache__/python.cpython-314.pyc
evals	Retirar Python del runner	2026-09-15T04:51:13.4548590Z retirado: /opt/az/lib/python3.14/site-packages/pygments/lexers/python.py
evals	Retirar Python del runner	2026-09-15T04:51:13.4688504Z retirado: /opt/az/lib/python3.14/site-packages/pip/_vendor/pygments/lexers/__pycache__/python.cpython-314.pyc
evals	Retirar Python del runner	2026-09-15T04:51:13.4830133Z retirado: /opt/az/lib/python3.14/site-packages/pip/_vendor/pygments/lexers/python.py
evals	Retirar Python del runner	2026-09-15T04:51:13.4970510Z retirado: /opt/az/lib/python3.14/site-packages/argcomplete/scripts/__pycache__/python_argcomplete_check_easy_install_script.cpython-314.pyc
evals	Retirar Python del runner	2026-09-15T04:51:13.5109931Z retirado: /opt/az/lib/python3.14/site-packages/argcomplete/scripts/python_argcomplete_check_easy_install_script.py
evals	Retirar Python del runner	2026-09-15T04:51:13.5249703Z retirado: /opt/az/lib/python3.14/config-3.14-x86_64-linux-gnu/libpython3.14.a
evals	Retirar Python del runner	2026-09-15T04:51:13.5389749Z retirado: /opt/az/lib/python3.14/config-3.14-x86_64-linux-gnu/python-config.py
evals	Retirar Python del runner	2026-09-15T04:51:13.5529912Z retirado: /opt/az/lib/python3.14/config-3.14-x86_64-linux-gnu/python.o
evals	Retirar Python del runner	2026-09-15T04:51:13.5669762Z retirado: /opt/az/lib/python3.14/idlelib/Icons/python.gif
evals	Retirar Python del runner	2026-09-15T04:51:13.5809541Z retirado: /opt/az/lib/libpython3.14.a
evals	Retirar Python del runner	2026-09-15T04:51:13.5949637Z retirado: /opt/az/include/python3.14/cpython/pythonrun.h
evals	Retirar Python del runner	2026-09-15T04:51:13.6089857Z retirado: /opt/az/include/python3.14/Python.h
evals	Retirar Python del runner	2026-09-15T04:51:13.6228682Z retirado: /opt/az/include/python3.14/pythonrun.h
evals	Retirar Python del runner	2026-09-15T04:51:13.6373193Z retirado: /opt/az/share/man/man1/python3.1
evals	Retirar Python del runner	2026-09-15T04:51:13.6516977Z retirado: /opt/az/share/man/man1/python3.14.1
evals	Retirar Python del runner	2026-09-15T04:51:13.6791661Z retirado: /opt/az
evals	Retirar Python del runner	2026-09-15T04:51:14.3324752Z retirado: /var/lib/dpkg/info/python3-jinja2.prerm
evals	Retirar Python del runner	2026-09-15T04:51:14.3461889Z retirado: /var/lib/dpkg/info/python3-packaging.prerm
evals	Retirar Python del runner	2026-09-15T04:51:14.3597553Z retirado: /var/lib/dpkg/info/python3-launchpadlib.postinst
evals	Retirar Python del runner	2026-09-15T04:51:14.3733782Z retirado: /var/lib/dpkg/info/python3-magic.postinst
evals	Retirar Python del runner	2026-09-15T04:51:14.3868321Z retirado: /var/lib/dpkg/info/python3-jsonschema.postrm
evals	Retirar Python del runner	2026-09-15T04:51:14.4004026Z retirado: /var/lib/dpkg/info/python3-jsonpatch.postinst
evals	Retirar Python del runner	2026-09-15T04:51:14.4141208Z retirado: /var/lib/dpkg/info/python3-parted.prerm
evals	Retirar Python del runner	2026-09-15T04:51:14.4278266Z retirado: /var/lib/dpkg/info/python3-chardet.postinst
evals	Retirar Python del runner	2026-09-15T04:51:14.4416901Z retirado: /var/lib/dpkg/info/python3-parted.postinst
evals	Retirar Python del runner	2026-09-15T04:51:14.4552141Z retirado: /var/lib/dpkg/info/python3-constantly.prerm
evals	Retirar Python del runner	2026-09-15T04:51:14.4689374Z retirado: /var/lib/dpkg/info/python3-gi.prerm
evals	Retirar Python del runner	2026-09-15T04:51:14.4826892Z retirado: /var/lib/dpkg/info/python3-s3transfer.prerm
evals	Retirar Python del runner	2026-09-15T04:51:14.4962650Z retirado: /var/lib/dpkg/info/python3-bcrypt.postinst
evals	Retirar Python del runner	2026-09-15T04:51:14.5099970Z retirado: /var/lib/dpkg/info/python3-netaddr.postinst
evals	Retirar Python del runner	2026-09-15T04:51:14.5236365Z retirado: /var/lib/dpkg/info/python3-zope.interface.prerm
evals	Retirar Python del runner	2026-09-15T04:51:14.5376797Z retirado: /var/lib/dpkg/info/python3-cryptography.postinst
evals	Retirar Python del runner	2026-09-15T04:51:14.5512776Z retirado: /var/lib/dpkg/info/python3-distro-info.prerm
evals	Retirar Python del runner	2026-09-15T04:51:14.5649299Z retirado: /var/lib/dpkg/info/python3-configobj.postinst
evals	Retirar Python del runner	2026-09-15T04:51:14.5785211Z retirado: /var/lib/dpkg/info/libpython3.12t64:amd64.md5sums
evals	Retirar Python del runner	2026-09-15T04:51:14.5923110Z retirado: /var/lib/dpkg/info/python3-lazr.restfulclient.prerm
evals	Retirar Python del runner	2026-09-15T04:51:14.6057910Z retirado: /var/lib/dpkg/info/python3-jsonschema.postinst
evals	Retirar Python del runner	2026-09-15T04:51:14.6194933Z retirado: /var/lib/dpkg/info/python3-six.postinst
evals	Retirar Python del runner	2026-09-15T04:51:14.6334799Z retirado: /var/lib/dpkg/info/libpython3.12-minimal:amd64.postrm
evals	Retirar Python del runner	2026-09-15T04:51:14.6473444Z retirado: /var/lib/dpkg/info/python3-idna.prerm
evals	Retirar Python del runner	2026-09-15T04:51:14.6610073Z retirado: /var/lib/dpkg/info/python3-jsonpatch.prerm
evals	Retirar Python del runner	2026-09-15T04:51:14.6745310Z retirado: /var/lib/dpkg/info/python3-cryptography.prerm
evals	Retirar Python del runner	2026-09-15T04:51:14.6881248Z retirado: /var/lib/dpkg/info/python3-babel.postinst
evals	Retirar Python del runner	2026-09-15T04:51:14.7019852Z retirado: /var/lib/dpkg/info/python3-distupgrade.postinst
evals	Retirar Python del runner	2026-09-15T04:51:14.7158120Z retirado: /var/lib/dpkg/info/python3-minimal.prerm
evals	Retirar Python del runner	2026-09-15T04:51:14.7296093Z retirado: /var/lib/dpkg/info/python3-mdurl.prerm
evals	Retirar Python del runner	2026-09-15T04:51:14.7432448Z retirado: /var/lib/dpkg/info/python3-pkg-resources.postinst
evals	Retirar Python del runner	2026-09-15T04:51:14.7572539Z retirado: /var/lib/dpkg/info/python3-launchpadlib.prerm
evals	Retirar Python del runner	2026-09-15T04:51:14.7714304Z retirado: /var/lib/dpkg/info/python3-debian.postinst
evals	Retirar Python del runner	2026-09-15T04:51:14.7854027Z retirado: /var/lib/dpkg/info/python3-wheel.prerm
evals	Retirar Python del runner	2026-09-15T04:51:14.7994087Z retirado: /var/lib/dpkg/info/python3.12-minimal.postrm
evals	Retirar Python del runner	2026-09-15T04:51:14.8133526Z retirado: /var/lib/dpkg/info/python3-certifi.postinst
evals	Retirar Python del runner	2026-09-15T04:51:14.8272894Z retirado: /var/lib/dpkg/info/python3-twisted.postrm
evals	Retirar Python del runner	2026-09-15T04:51:14.8412216Z retirado: /var/lib/dpkg/info/python3-systemd.postinst
evals	Retirar Python del runner	2026-09-15T04:51:14.8551840Z retirado: /var/lib/dpkg/info/python3-botocore.prerm
evals	Retirar Python del runner	2026-09-15T04:51:14.8690142Z retirado: /var/lib/dpkg/info/python3.12-venv.postrm
evals	Retirar Python del runner	2026-09-15T04:51:14.8829638Z retirado: /var/lib/dpkg/info/python3-openssl.postinst
evals	Retirar Python del runner	2026-09-15T04:51:14.8967741Z retirado: /var/lib/dpkg/info/python3-launchpadlib.postrm
evals	Retirar Python del runner	2026-09-15T04:51:14.9108356Z retirado: /var/lib/dpkg/info/python3-json-pointer.postinst
evals	Retirar Python del runner	2026-09-15T04:51:14.9246119Z retirado: /var/lib/dpkg/info/python3-requests.prerm
evals	Retirar Python del runner	2026-09-15T04:51:14.9385640Z retirado: /var/lib/dpkg/info/python3-pyasn1.prerm
evals	Retirar Python del runner	2026-09-15T04:51:14.9525201Z retirado: /var/lib/dpkg/info/python3-openssl.prerm
evals	Retirar Python del runner	2026-09-15T04:51:14.9666094Z retirado: /var/lib/dpkg/info/python3-attr.postinst
evals	Retirar Python del runner	2026-09-15T04:51:14.9806385Z retirado: /var/lib/dpkg/info/python3.preinst
evals	Retirar Python del runner	2026-09-15T04:51:14.9944552Z retirado: /var/lib/dpkg/info/python3-apt.prerm
evals	Retirar Python del runner	2026-09-15T04:51:15.0083427Z retirado: /var/lib/dpkg/info/python3-pyasn1-modules.postinst
evals	Retirar Python del runner	2026-09-15T04:51:15.0221963Z retirado: /var/lib/dpkg/info/libpython3.12-stdlib:amd64.md5sums
evals	Retirar Python del runner	2026-09-15T04:51:15.0361397Z retirado: /var/lib/dpkg/info/python3-newt:amd64.postinst
evals	Retirar Python del runner	2026-09-15T04:51:15.0499970Z retirado: /var/lib/dpkg/info/libpython3-stdlib:amd64.list
evals	Retirar Python del runner	2026-09-15T04:51:15.0638750Z retirado: /var/lib/dpkg/info/python3-commandnotfound.postinst
evals	Retirar Python del runner	2026-09-15T04:51:15.0779703Z retirado: /var/lib/dpkg/info/libpython3.12t64:amd64.symbols
evals	Retirar Python del runner	2026-09-15T04:51:15.0918580Z retirado: /var/lib/dpkg/info/python3-pyrsistent:amd64.prerm
evals	Retirar Python del runner	2026-09-15T04:51:15.1057851Z retirado: /var/lib/dpkg/info/python3-yaml.postinst
evals	Retirar Python del runner	2026-09-15T04:51:15.1196365Z retirado: /var/lib/dpkg/info/python3-debconf.postinst
evals	Retirar Python del runner	2026-09-15T04:51:15.1341453Z retirado: /var/lib/dpkg/info/python3-boto3.postinst
evals	Retirar Python del runner	2026-09-15T04:51:15.1479688Z retirado: /var/lib/dpkg/info/python3-passlib.prerm
evals	Retirar Python del runner	2026-09-15T04:51:15.1620132Z retirado: /var/lib/dpkg/info/python3.12.prerm
evals	Retirar Python del runner	2026-09-15T04:51:15.1758417Z retirado: /var/lib/dpkg/info/python3-idna.postinst
evals	Retirar Python del runner	2026-09-15T04:51:15.1897952Z retirado: /var/lib/dpkg/info/python3-problem-report.prerm
evals	Retirar Python del runner	2026-09-15T04:51:15.2033483Z retirado: /var/lib/dpkg/info/python3.12-venv.prerm
evals	Retirar Python del runner	2026-09-15T04:51:15.2168320Z retirado: /var/lib/dpkg/info/python3-apport.prerm
evals	Retirar Python del runner	2026-09-15T04:51:15.2303316Z retirado: /var/lib/dpkg/info/python3-newt:amd64.prerm
evals	Retirar Python del runner	2026-09-15T04:51:15.2439869Z retirado: /var/lib/dpkg/info/python3-distro-info.postinst
evals	Retirar Python del runner	2026-09-15T04:51:15.2574539Z retirado: /var/lib/dpkg/info/python3.12.postinst
evals	Retirar Python del runner	2026-09-15T04:51:15.2710225Z retirado: /var/lib/dpkg/info/python3-pip.prerm
evals	Retirar Python del runner	2026-09-15T04:51:15.2846854Z retirado: /var/lib/dpkg/info/python3.12-minimal.preinst
evals	Retirar Python del runner	2026-09-15T04:51:15.2985113Z retirado: /var/lib/dpkg/info/python3-urllib3.postinst
evals	Retirar Python del runner	2026-09-15T04:51:15.3120064Z retirado: /var/lib/dpkg/info/python3-bpfcc.prerm
evals	Retirar Python del runner	2026-09-15T04:51:15.3254678Z retirado: /var/lib/dpkg/info/python3-wadllib.postinst
evals	Retirar Python del runner	2026-09-15T04:51:15.3389966Z retirado: /var/lib/dpkg/info/python3-jwt.postinst
evals	Retirar Python del runner	2026-09-15T04:51:15.3522912Z retirado: /var/lib/dpkg/info/python3-distupgrade.prerm
evals	Retirar Python del runner	2026-09-15T04:51:15.3659675Z retirado: /var/lib/dpkg/info/python3-problem-report.postinst
evals	Retirar Python del runner	2026-09-15T04:51:15.3794103Z retirado: /var/lib/dpkg/info/python3-pexpect.postinst
evals	Retirar Python del runner	2026-09-15T04:51:15.3931096Z retirado: /var/lib/dpkg/info/python3-zstandard.postinst
evals	Retirar Python del runner	2026-09-15T04:51:15.4064392Z retirado: /var/lib/dpkg/info/python3-gi.postinst
evals	Retirar Python del runner	2026-09-15T04:51:15.4202401Z retirado: /var/lib/dpkg/info/python3-update-manager.postinst
evals	Retirar Python del runner	2026-09-15T04:51:15.4337192Z retirado: /var/lib/dpkg/info/python3-httplib2.prerm
evals	Retirar Python del runner	2026-09-15T04:51:15.4471435Z retirado: /var/lib/dpkg/info/python3-pyasn1.postinst
evals	Retirar Python del runner	2026-09-15T04:51:15.4607368Z retirado: /var/lib/dpkg/info/python3-pkg-resources.prerm
evals	Retirar Python del runner	2026-09-15T04:51:15.4743684Z retirado: /var/lib/dpkg/info/python3-markupsafe.postinst
evals	Retirar Python del runner	2026-09-15T04:51:15.4877894Z retirado: /var/lib/dpkg/info/python3-boto3.prerm
evals	Retirar Python del runner	2026-09-15T04:51:15.5012288Z retirado: /var/lib/dpkg/info/libpython3-stdlib:amd64.md5sums
evals	Retirar Python del runner	2026-09-15T04:51:15.5147938Z retirado: /var/lib/dpkg/info/libpython3.12-minimal:amd64.md5sums
evals	Retirar Python del runner	2026-09-15T04:51:15.5283441Z retirado: /var/lib/dpkg/info/python3-markdown-it.prerm
evals	Retirar Python del runner	2026-09-15T04:51:15.5417152Z retirado: /var/lib/dpkg/info/python3-distro.prerm
evals	Retirar Python del runner	2026-09-15T04:51:15.5552266Z retirado: /var/lib/dpkg/info/python3-requests.postinst
evals	Retirar Python del runner	2026-09-15T04:51:15.5691075Z retirado: /var/lib/dpkg/info/python3-hyperlink.postinst
evals	Retirar Python del runner	2026-09-15T04:51:15.5830292Z retirado: /var/lib/dpkg/info/libpython3.12t64:amd64.list
evals	Retirar Python del runner	2026-09-15T04:51:15.5978464Z retirado: /var/lib/dpkg/info/python3-hyperlink.prerm
evals	Retirar Python del runner	2026-09-15T04:51:15.6115774Z retirado: /var/lib/dpkg/info/python3-minimal.postinst
evals	Retirar Python del runner	2026-09-15T04:51:15.6251488Z retirado: /var/lib/dpkg/info/python3-jwt.prerm
evals	Retirar Python del runner	2026-09-15T04:51:15.6394026Z retirado: /var/lib/dpkg/info/python3-pyasn1-modules.prerm
evals	Retirar Python del runner	2026-09-15T04:51:15.6530844Z retirado: /var/lib/dpkg/info/python3-lazr.uri.postinst
evals	Retirar Python del runner	2026-09-15T04:51:15.6667950Z retirado: /var/lib/dpkg/info/python3-jsonpatch.postrm
evals	Retirar Python del runner	2026-09-15T04:51:15.6805807Z retirado: /var/lib/dpkg/info/python3-pygments.postinst
evals	Retirar Python del runner	2026-09-15T04:51:15.6960006Z retirado: /var/lib/dpkg/info/python3-json-pointer.postrm
evals	Retirar Python del runner	2026-09-15T04:51:15.7088206Z retirado: /var/lib/dpkg/info/libpython3.12-minimal:amd64.prerm
evals	Retirar Python del runner	2026-09-15T04:51:15.7227640Z retirado: /var/lib/dpkg/info/python3-rich.postinst
evals	Retirar Python del runner	2026-09-15T04:51:15.7365342Z retirado: /var/lib/dpkg/info/python3-jsonschema.prerm
evals	Retirar Python del runner	2026-09-15T04:51:15.7506288Z retirado: /var/lib/dpkg/info/python3-mdurl.postinst
evals	Retirar Python del runner	2026-09-15T04:51:15.7645505Z retirado: /var/lib/dpkg/info/python3-software-properties.postinst
evals	Retirar Python del runner	2026-09-15T04:51:15.7787084Z retirado: /var/lib/dpkg/info/python3-pyparsing.postinst
evals	Retirar Python del runner	2026-09-15T04:51:15.7928147Z retirado: /var/lib/dpkg/info/python3-pip.postinst
evals	Retirar Python del runner	2026-09-15T04:51:15.8069364Z retirado: /var/lib/dpkg/info/python3-distro.postinst
evals	Retirar Python del runner	2026-09-15T04:51:15.8205715Z retirado: /var/lib/dpkg/info/python3-hamcrest.prerm
evals	Retirar Python del runner	2026-09-15T04:51:15.8347047Z retirado: /var/lib/dpkg/info/python3-urllib3.prerm
evals	Retirar Python del runner	2026-09-15T04:51:15.8489869Z retirado: /var/lib/dpkg/info/python3-wadllib.prerm
evals	Retirar Python del runner	2026-09-15T04:51:15.8629031Z retirado: /var/lib/dpkg/info/python3-markupsafe.prerm
evals	Retirar Python del runner	2026-09-15T04:51:15.8767439Z retirado: /var/lib/dpkg/info/python3-httplib2.postinst
evals	Retirar Python del runner	2026-09-15T04:51:15.8910203Z retirado: /var/lib/dpkg/info/python3-certifi.prerm
evals	Retirar Python del runner	2026-09-15T04:51:15.9053211Z retirado: /var/lib/dpkg/info/python3-click.postinst
evals	Retirar Python del runner	2026-09-15T04:51:15.9192405Z retirado: /var/lib/dpkg/info/python3-constantly.postinst
evals	Retirar Python del runner	2026-09-15T04:51:15.9331597Z retirado: /var/lib/dpkg/info/libpython3-dev:amd64.list
evals	Retirar Python del runner	2026-09-15T04:51:15.9471494Z retirado: /var/lib/dpkg/info/python3.12-minimal.postinst
evals	Retirar Python del runner	2026-09-15T04:51:15.9611427Z retirado: /var/lib/dpkg/info/python3-s3transfer.postinst
evals	Retirar Python del runner	2026-09-15T04:51:15.9753140Z retirado: /var/lib/dpkg/info/python3-zstandard.prerm
evals	Retirar Python del runner	2026-09-15T04:51:15.9889945Z retirado: /var/lib/dpkg/info/python3-json-pointer.prerm
evals	Retirar Python del runner	2026-09-15T04:51:16.0029226Z retirado: /var/lib/dpkg/info/python3-service-identity.postinst
evals	Retirar Python del runner	2026-09-15T04:51:16.0169858Z retirado: /var/lib/dpkg/info/python3-serial.postinst
evals	Retirar Python del runner	2026-09-15T04:51:16.0308100Z retirado: /var/lib/dpkg/info/python3-hamcrest.postinst
evals	Retirar Python del runner	2026-09-15T04:51:16.0448424Z retirado: /var/lib/dpkg/info/python3-incremental.postinst
evals	Retirar Python del runner	2026-09-15T04:51:16.0588626Z retirado: /var/lib/dpkg/info/python3-netplan.postinst
evals	Retirar Python del runner	2026-09-15T04:51:16.0729629Z retirado: /var/lib/dpkg/info/python3-netaddr.prerm
evals	Retirar Python del runner	2026-09-15T04:51:16.0868627Z retirado: /var/lib/dpkg/info/python3-dateutil.prerm
evals	Retirar Python del runner	2026-09-15T04:51:16.1009606Z retirado: /var/lib/dpkg/info/python3-apt.postinst
evals	Retirar Python del runner	2026-09-15T04:51:16.1151054Z retirado: /var/lib/dpkg/info/python3-dbus.postinst
evals	Retirar Python del runner	2026-09-15T04:51:16.1291978Z retirado: /var/lib/dpkg/info/python3-jmespath.prerm
evals	Retirar Python del runner	2026-09-15T04:51:16.1437776Z retirado: /var/lib/dpkg/info/libpython3.12-stdlib:amd64.prerm
evals	Retirar Python del runner	2026-09-15T04:51:16.1575142Z retirado: /var/lib/dpkg/info/python3-commandnotfound.prerm
evals	Retirar Python del runner	2026-09-15T04:51:16.1717241Z retirado: /var/lib/dpkg/info/python3-blinker.prerm
evals	Retirar Python del runner	2026-09-15T04:51:16.1858112Z retirado: /var/lib/dpkg/info/python3-ptyprocess.postinst
evals	Retirar Python del runner	2026-09-15T04:51:16.1998667Z retirado: /var/lib/dpkg/info/python3-colorama.postinst
evals	Retirar Python del runner	2026-09-15T04:51:16.2138031Z retirado: /var/lib/dpkg/info/python3-wheel.postinst
evals	Retirar Python del runner	2026-09-15T04:51:16.2278732Z retirado: /var/lib/dpkg/info/python3-oauthlib.postinst
evals	Retirar Python del runner	2026-09-15T04:51:16.2418505Z retirado: /var/lib/dpkg/info/python3-pygments.prerm
evals	Retirar Python del runner	2026-09-15T04:51:16.2558175Z retirado: /var/lib/dpkg/info/python3-tz.postinst
evals	Retirar Python del runner	2026-09-15T04:51:16.2700576Z retirado: /var/lib/dpkg/info/python3.prerm
evals	Retirar Python del runner	2026-09-15T04:51:16.2842357Z retirado: /var/lib/dpkg/info/python3-update-manager.prerm
evals	Retirar Python del runner	2026-09-15T04:51:16.2980218Z retirado: /var/lib/dpkg/info/python3-pexpect.prerm
evals	Retirar Python del runner	2026-09-15T04:51:16.3121407Z retirado: /var/lib/dpkg/info/python3-serial.prerm
evals	Retirar Python del runner	2026-09-15T04:51:16.3258680Z retirado: /var/lib/dpkg/info/python3-netplan.prerm
evals	Retirar Python del runner	2026-09-15T04:51:16.3398078Z retirado: /var/lib/dpkg/info/python3-incremental.prerm
evals	Retirar Python del runner	2026-09-15T04:51:16.3537130Z retirado: /var/lib/dpkg/info/python3-typing-extensions.postinst
evals	Retirar Python del runner	2026-09-15T04:51:16.3677161Z retirado: /var/lib/dpkg/info/python3-jinja2.postinst
evals	Retirar Python del runner	2026-09-15T04:51:16.3816418Z retirado: /var/lib/dpkg/info/libpython3.12-stdlib:amd64.list
evals	Retirar Python del runner	2026-09-15T04:51:16.3956855Z retirado: /var/lib/dpkg/info/python3-pyparsing.prerm
evals	Retirar Python del runner	2026-09-15T04:51:16.4089865Z retirado: /var/lib/dpkg/info/python3-automat.postinst
evals	Retirar Python del runner	2026-09-15T04:51:16.4225276Z retirado: /var/lib/dpkg/info/python3-attr.prerm
evals	Retirar Python del runner	2026-09-15T04:51:16.4362366Z retirado: /var/lib/dpkg/info/python3-pyrsistent:amd64.postinst
evals	Retirar Python del runner	2026-09-15T04:51:16.4501534Z retirado: /var/lib/dpkg/info/python3-passlib.postinst
evals	Retirar Python del runner	2026-09-15T04:51:16.4639896Z retirado: /var/lib/dpkg/info/python3-twisted.postinst
evals	Retirar Python del runner	2026-09-15T04:51:16.4774697Z retirado: /var/lib/dpkg/info/python3-configobj.prerm
evals	Retirar Python del runner	2026-09-15T04:51:16.4911503Z retirado: /var/lib/dpkg/info/python3-markdown-it.postinst
evals	Retirar Python del runner	2026-09-15T04:51:16.5047017Z retirado: /var/lib/dpkg/info/python3-ptyprocess.prerm
evals	Retirar Python del runner	2026-09-15T04:51:16.5183788Z retirado: /var/lib/dpkg/info/libpython3-dev:amd64.md5sums
evals	Retirar Python del runner	2026-09-15T04:51:16.5323373Z retirado: /var/lib/dpkg/info/python3-software-properties.prerm
evals	Retirar Python del runner	2026-09-15T04:51:16.5458540Z retirado: /var/lib/dpkg/info/python3-dbus.prerm
evals	Retirar Python del runner	2026-09-15T04:51:16.5596447Z retirado: /var/lib/dpkg/info/libpython3.12-minimal:amd64.postinst
evals	Retirar Python del runner	2026-09-15T04:51:16.5731494Z retirado: /var/lib/dpkg/info/python3-setuptools.postinst
evals	Retirar Python del runner	2026-09-15T04:51:16.5869828Z retirado: /var/lib/dpkg/info/python3.12-minimal.prerm
evals	Retirar Python del runner	2026-09-15T04:51:16.6012411Z retirado: /var/lib/dpkg/info/python3-botocore.postinst
evals	Retirar Python del runner	2026-09-15T04:51:16.6150537Z retirado: /var/lib/dpkg/info/python3-setuptools.prerm
evals	Retirar Python del runner	2026-09-15T04:51:16.6288365Z retirado: /var/lib/dpkg/info/python3-dateutil.postinst
evals	Retirar Python del runner	2026-09-15T04:51:16.6430449Z retirado: /var/lib/dpkg/info/python3.postrm
evals	Retirar Python del runner	2026-09-15T04:51:16.6569583Z retirado: /var/lib/dpkg/info/python3-yaml.prerm
evals	Retirar Python del runner	2026-09-15T04:51:16.6706821Z retirado: /var/lib/dpkg/info/libpython3.12-dev:amd64.list
evals	Retirar Python del runner	2026-09-15T04:51:16.6846009Z retirado: /var/lib/dpkg/info/python3-lazr.restfulclient.postinst
evals	Retirar Python del runner	2026-09-15T04:51:16.6983906Z retirado: /var/lib/dpkg/info/python3-click.prerm
evals	Retirar Python del runner	2026-09-15T04:51:16.7123128Z retirado: /var/lib/dpkg/info/python3-tz.prerm
evals	Retirar Python del runner	2026-09-15T04:51:16.7262067Z retirado: /var/lib/dpkg/info/python3-debconf.prerm
evals	Retirar Python del runner	2026-09-15T04:51:16.7399214Z retirado: /var/lib/dpkg/info/libpython3.12-dev:amd64.md5sums
evals	Retirar Python del runner	2026-09-15T04:51:16.7537795Z retirado: /var/lib/dpkg/info/python3-automat.prerm
evals	Retirar Python del runner	2026-09-15T04:51:16.7675886Z retirado: /var/lib/dpkg/info/python3-systemd.prerm
evals	Retirar Python del runner	2026-09-15T04:51:16.7814590Z retirado: /var/lib/dpkg/info/python3-typing-extensions.prerm
evals	Retirar Python del runner	2026-09-15T04:51:16.7953619Z retirado: /var/lib/dpkg/info/python3-chardet.prerm
evals	Retirar Python del runner	2026-09-15T04:51:16.8092981Z retirado: /var/lib/dpkg/info/python3-packaging.postinst
evals	Retirar Python del runner	2026-09-15T04:51:16.8231460Z retirado: /var/lib/dpkg/info/libpython3.12t64:amd64.triggers
evals	Retirar Python del runner	2026-09-15T04:51:16.8369621Z retirado: /var/lib/dpkg/info/python3-blinker.postinst
evals	Retirar Python del runner	2026-09-15T04:51:16.8509011Z retirado: /var/lib/dpkg/info/python3.12-venv.postinst
evals	Retirar Python del runner	2026-09-15T04:51:16.8648572Z retirado: /var/lib/dpkg/info/python3-lazr.uri.prerm
evals	Retirar Python del runner	2026-09-15T04:51:16.8788375Z retirado: /var/lib/dpkg/info/python3-six.prerm
evals	Retirar Python del runner	2026-09-15T04:51:16.8927450Z retirado: /var/lib/dpkg/info/python3-twisted.prerm
evals	Retirar Python del runner	2026-09-15T04:51:16.9066588Z retirado: /var/lib/dpkg/info/python3-bcrypt.prerm
evals	Retirar Python del runner	2026-09-15T04:51:16.9205752Z retirado: /var/lib/dpkg/info/python3-magic.prerm
evals	Retirar Python del runner	2026-09-15T04:51:16.9344234Z retirado: /var/lib/dpkg/info/libpython3.12-minimal:amd64.list
evals	Retirar Python del runner	2026-09-15T04:51:16.9483840Z retirado: /var/lib/dpkg/info/python3-service-identity.prerm
evals	Retirar Python del runner	2026-09-15T04:51:16.9622355Z retirado: /var/lib/dpkg/info/python3-colorama.prerm
evals	Retirar Python del runner	2026-09-15T04:51:16.9763160Z retirado: /var/lib/dpkg/info/python3-jmespath.postinst
evals	Retirar Python del runner	2026-09-15T04:51:16.9903127Z retirado: /var/lib/dpkg/info/libpython3.12t64:amd64.shlibs
evals	Retirar Python del runner	2026-09-15T04:51:17.0042635Z retirado: /var/lib/dpkg/info/python3-oauthlib.prerm
evals	Retirar Python del runner	2026-09-15T04:51:17.0182616Z retirado: /var/lib/dpkg/info/python3-rich.prerm
evals	Retirar Python del runner	2026-09-15T04:51:17.0321828Z retirado: /var/lib/dpkg/info/python3-babel.prerm
evals	Retirar Python del runner	2026-09-15T04:51:17.0462596Z retirado: /var/lib/dpkg/info/python3-apport.postinst
evals	Retirar Python del runner	2026-09-15T04:51:17.0602174Z retirado: /var/lib/dpkg/info/python3-bpfcc.postinst
evals	Retirar Python del runner	2026-09-15T04:51:17.0740731Z retirado: /var/lib/dpkg/info/python3-zope.interface.postinst
evals	Retirar Python del runner	2026-09-15T04:51:17.0880275Z retirado: /var/lib/dpkg/info/libpython3.12-minimal:amd64.conffiles
evals	Retirar Python del runner	2026-09-15T04:51:17.1019549Z retirado: /var/lib/dpkg/info/python3.postinst
evals	Retirar Python del runner	2026-09-15T04:51:17.1159333Z retirado: /var/lib/dpkg/info/python3-debian.prerm
evals	Retirar Python del runner	2026-09-15T04:51:17.1301217Z retirado: /var/lib/docker/overlay2/a9853d7ac0752db045c472a9505e498ad787cec8cb7eab1bd6d6d65f685ef414/diff/var/lib/dpkg/info/python3-launchpadlib.postinst
evals	Retirar Python del runner	2026-09-15T04:51:17.1457748Z retirado: /var/lib/docker/overlay2/a9853d7ac0752db045c472a9505e498ad787cec8cb7eab1bd6d6d65f685ef414/diff/var/lib/dpkg/info/python3-cryptography.postinst
evals	Retirar Python del runner	2026-09-15T04:51:17.1597729Z retirado: /var/lib/docker/overlay2/a9853d7ac0752db045c472a9505e498ad787cec8cb7eab1bd6d6d65f685ef414/diff/var/lib/dpkg/info/python3-configobj.postinst
evals	Retirar Python del runner	2026-09-15T04:51:17.1737365Z retirado: /var/lib/docker/overlay2/a9853d7ac0752db045c472a9505e498ad787cec8cb7eab1bd6d6d65f685ef414/diff/var/lib/dpkg/info/libpython3.12t64:amd64.md5sums
evals	Retirar Python del runner	2026-09-15T04:51:17.1877204Z retirado: /var/lib/docker/overlay2/a9853d7ac0752db045c472a9505e498ad787cec8cb7eab1bd6d6d65f685ef414/diff/var/lib/dpkg/info/python3-lazr.restfulclient.prerm
evals	Retirar Python del runner	2026-09-15T04:51:17.2018234Z retirado: /var/lib/docker/overlay2/a9853d7ac0752db045c472a9505e498ad787cec8cb7eab1bd6d6d65f685ef414/diff/var/lib/dpkg/info/python3-six.postinst
evals	Retirar Python del runner	2026-09-15T04:51:17.2158346Z retirado: /var/lib/docker/overlay2/a9853d7ac0752db045c472a9505e498ad787cec8cb7eab1bd6d6d65f685ef414/diff/var/lib/dpkg/info/libpython3.12-minimal:amd64.postrm
evals	Retirar Python del runner	2026-09-15T04:51:17.2299345Z retirado: /var/lib/docker/overlay2/a9853d7ac0752db045c472a9505e498ad787cec8cb7eab1bd6d6d65f685ef414/diff/var/lib/dpkg/info/python3-cryptography.prerm
evals	Retirar Python del runner	2026-09-15T04:51:17.2438503Z retirado: /var/lib/docker/overlay2/a9853d7ac0752db045c472a9505e498ad787cec8cb7eab1bd6d6d65f685ef414/diff/var/lib/dpkg/info/python3-minimal.prerm
evals	Retirar Python del runner	2026-09-15T04:51:17.2584458Z retirado: /var/lib/docker/overlay2/a9853d7ac0752db045c472a9505e498ad787cec8cb7eab1bd6d6d65f685ef414/diff/var/lib/dpkg/info/python3-pkg-resources.postinst
evals	Retirar Python del runner	2026-09-15T04:51:17.2723529Z retirado: /var/lib/docker/overlay2/a9853d7ac0752db045c472a9505e498ad787cec8cb7eab1bd6d6d65f685ef414/diff/var/lib/dpkg/info/python3-launchpadlib.prerm
evals	Retirar Python del runner	2026-09-15T04:51:17.2863885Z retirado: /var/lib/docker/overlay2/a9853d7ac0752db045c472a9505e498ad787cec8cb7eab1bd6d6d65f685ef414/diff/var/lib/dpkg/info/python3.12-minimal.postrm
evals	Retirar Python del runner	2026-09-15T04:51:17.3003178Z retirado: /var/lib/docker/overlay2/a9853d7ac0752db045c472a9505e498ad787cec8cb7eab1bd6d6d65f685ef414/diff/var/lib/dpkg/info/python3-launchpadlib.postrm
evals	Retirar Python del runner	2026-09-15T04:51:17.3144104Z retirado: /var/lib/docker/overlay2/a9853d7ac0752db045c472a9505e498ad787cec8cb7eab1bd6d6d65f685ef414/diff/var/lib/dpkg/info/python3.preinst
evals	Retirar Python del runner	2026-09-15T04:51:17.3282498Z retirado: /var/lib/docker/overlay2/a9853d7ac0752db045c472a9505e498ad787cec8cb7eab1bd6d6d65f685ef414/diff/var/lib/dpkg/info/python3-tzlocal.postinst
evals	Retirar Python del runner	2026-09-15T04:51:17.3422561Z retirado: /var/lib/docker/overlay2/a9853d7ac0752db045c472a9505e498ad787cec8cb7eab1bd6d6d65f685ef414/diff/var/lib/dpkg/info/libpython3.12-stdlib:amd64.md5sums
evals	Retirar Python del runner	2026-09-15T04:51:17.3561736Z retirado: /var/lib/docker/overlay2/a9853d7ac0752db045c472a9505e498ad787cec8cb7eab1bd6d6d65f685ef414/diff/var/lib/dpkg/info/libpython3-stdlib:amd64.list
evals	Retirar Python del runner	2026-09-15T04:51:17.3700943Z retirado: /var/lib/docker/overlay2/a9853d7ac0752db045c472a9505e498ad787cec8cb7eab1bd6d6d65f685ef414/diff/var/lib/dpkg/info/libpython3.12t64:amd64.symbols
evals	Retirar Python del runner	2026-09-15T04:51:17.3843702Z retirado: /var/lib/docker/overlay2/a9853d7ac0752db045c472a9505e498ad787cec8cb7eab1bd6d6d65f685ef414/diff/var/lib/dpkg/info/python3-yaml.postinst
evals	Retirar Python del runner	2026-09-15T04:51:17.3981070Z retirado: /var/lib/docker/overlay2/a9853d7ac0752db045c472a9505e498ad787cec8cb7eab1bd6d6d65f685ef414/diff/var/lib/dpkg/info/python3.12.prerm
evals	Retirar Python del runner	2026-09-15T04:51:17.4120254Z retirado: /var/lib/docker/overlay2/a9853d7ac0752db045c472a9505e498ad787cec8cb7eab1bd6d6d65f685ef414/diff/var/lib/dpkg/info/python3.12.postinst
evals	Retirar Python del runner	2026-09-15T04:51:17.4259104Z retirado: /var/lib/docker/overlay2/a9853d7ac0752db045c472a9505e498ad787cec8cb7eab1bd6d6d65f685ef414/diff/var/lib/dpkg/info/python3.12-minimal.preinst
evals	Retirar Python del runner	2026-09-15T04:51:17.4400140Z retirado: /var/lib/docker/overlay2/a9853d7ac0752db045c472a9505e498ad787cec8cb7eab1bd6d6d65f685ef414/diff/var/lib/dpkg/info/python3-merge3.postinst
evals	Retirar Python del runner	2026-09-15T04:51:17.4539288Z retirado: /var/lib/docker/overlay2/a9853d7ac0752db045c472a9505e498ad787cec8cb7eab1bd6d6d65f685ef414/diff/var/lib/dpkg/info/python3-urllib3.postinst
evals	Retirar Python del runner	2026-09-15T04:51:17.4680323Z retirado: /var/lib/docker/overlay2/a9853d7ac0752db045c472a9505e498ad787cec8cb7eab1bd6d6d65f685ef414/diff/var/lib/dpkg/info/python3-wadllib.postinst
evals	Retirar Python del runner	2026-09-15T04:51:17.4822010Z retirado: /var/lib/docker/overlay2/a9853d7ac0752db045c472a9505e498ad787cec8cb7eab1bd6d6d65f685ef414/diff/var/lib/dpkg/info/python3-jwt.postinst
evals	Retirar Python del runner	2026-09-15T04:51:17.4962148Z retirado: /var/lib/docker/overlay2/a9853d7ac0752db045c472a9505e498ad787cec8cb7eab1bd6d6d65f685ef414/diff/var/lib/dpkg/info/python3-fastbencode.prerm
evals	Retirar Python del runner	2026-09-15T04:51:17.5102381Z retirado: /var/lib/docker/overlay2/a9853d7ac0752db045c472a9505e498ad787cec8cb7eab1bd6d6d65f685ef414/diff/var/lib/dpkg/info/python3-httplib2.prerm
evals	Retirar Python del runner	2026-09-15T04:51:17.5242185Z retirado: /var/lib/docker/overlay2/a9853d7ac0752db045c472a9505e498ad787cec8cb7eab1bd6d6d65f685ef414/diff/var/lib/dpkg/info/python3-pkg-resources.prerm
evals	Retirar Python del runner	2026-09-15T04:51:17.5383196Z retirado: /var/lib/docker/overlay2/a9853d7ac0752db045c472a9505e498ad787cec8cb7eab1bd6d6d65f685ef414/diff/var/lib/dpkg/info/libpython3-stdlib:amd64.md5sums
evals	Retirar Python del runner	2026-09-15T04:51:17.5522950Z retirado: /var/lib/docker/overlay2/a9853d7ac0752db045c472a9505e498ad787cec8cb7eab1bd6d6d65f685ef414/diff/var/lib/dpkg/info/libpython3.12-minimal:amd64.md5sums
evals	Retirar Python del runner	2026-09-15T04:51:17.5662965Z retirado: /var/lib/docker/overlay2/a9853d7ac0752db045c472a9505e498ad787cec8cb7eab1bd6d6d65f685ef414/diff/var/lib/dpkg/info/python3-distro.prerm
evals	Retirar Python del runner	2026-09-15T04:51:17.5802104Z retirado: /var/lib/docker/overlay2/a9853d7ac0752db045c472a9505e498ad787cec8cb7eab1bd6d6d65f685ef414/diff/var/lib/dpkg/info/libpython3.12t64:amd64.list
evals	Retirar Python del runner	2026-09-15T04:51:17.5939186Z retirado: /var/lib/docker/overlay2/a9853d7ac0752db045c472a9505e498ad787cec8cb7eab1bd6d6d65f685ef414/diff/var/lib/dpkg/info/python3-minimal.postinst
evals	Retirar Python del runner	2026-09-15T04:51:17.6078723Z retirado: /var/lib/docker/overlay2/a9853d7ac0752db045c472a9505e498ad787cec8cb7eab1bd6d6d65f685ef414/diff/var/lib/dpkg/info/python3-jwt.prerm
evals	Retirar Python del runner	2026-09-15T04:51:17.6217640Z retirado: /var/lib/docker/overlay2/a9853d7ac0752db045c472a9505e498ad787cec8cb7eab1bd6d6d65f685ef414/diff/var/lib/dpkg/info/python3-lazr.uri.postinst
evals	Retirar Python del runner	2026-09-15T04:51:17.6365384Z retirado: /var/lib/docker/overlay2/a9853d7ac0752db045c472a9505e498ad787cec8cb7eab1bd6d6d65f685ef414/diff/var/lib/dpkg/info/python3-tzlocal.prerm
evals	Retirar Python del runner	2026-09-15T04:51:17.6507905Z retirado: /var/lib/docker/overlay2/a9853d7ac0752db045c472a9505e498ad787cec8cb7eab1bd6d6d65f685ef414/diff/var/lib/dpkg/info/libpython3.12-minimal:amd64.prerm
evals	Retirar Python del runner	2026-09-15T04:51:17.6646894Z retirado: /var/lib/docker/overlay2/a9853d7ac0752db045c472a9505e498ad787cec8cb7eab1bd6d6d65f685ef414/diff/var/lib/dpkg/info/python3-pyparsing.postinst
evals	Retirar Python del runner	2026-09-15T04:51:17.6787379Z retirado: /var/lib/docker/overlay2/a9853d7ac0752db045c472a9505e498ad787cec8cb7eab1bd6d6d65f685ef414/diff/var/lib/dpkg/info/python3-distro.postinst
evals	Retirar Python del runner	2026-09-15T04:51:17.6926249Z retirado: /var/lib/docker/overlay2/a9853d7ac0752db045c472a9505e498ad787cec8cb7eab1bd6d6d65f685ef414/diff/var/lib/dpkg/info/python3-urllib3.prerm
evals	Retirar Python del runner	2026-09-15T04:51:17.7067025Z retirado: /var/lib/docker/overlay2/a9853d7ac0752db045c472a9505e498ad787cec8cb7eab1bd6d6d65f685ef414/diff/var/lib/dpkg/info/python3-wadllib.prerm
evals	Retirar Python del runner	2026-09-15T04:51:17.7205065Z retirado: /var/lib/docker/overlay2/a9853d7ac0752db045c472a9505e498ad787cec8cb7eab1bd6d6d65f685ef414/diff/var/lib/dpkg/info/python3-httplib2.postinst
evals	Retirar Python del runner	2026-09-15T04:51:17.7345394Z retirado: /var/lib/docker/overlay2/a9853d7ac0752db045c472a9505e498ad787cec8cb7eab1bd6d6d65f685ef414/diff/var/lib/dpkg/info/python3.12-minimal.postinst
evals	Retirar Python del runner	2026-09-15T04:51:17.7485822Z retirado: /var/lib/docker/overlay2/a9853d7ac0752db045c472a9505e498ad787cec8cb7eab1bd6d6d65f685ef414/diff/var/lib/dpkg/info/python3-patiencediff.postinst
evals	Retirar Python del runner	2026-09-15T04:51:17.7625138Z retirado: /var/lib/docker/overlay2/a9853d7ac0752db045c472a9505e498ad787cec8cb7eab1bd6d6d65f685ef414/diff/var/lib/dpkg/info/libpython3.12-stdlib:amd64.prerm
evals	Retirar Python del runner	2026-09-15T04:51:17.7765444Z retirado: /var/lib/docker/overlay2/a9853d7ac0752db045c472a9505e498ad787cec8cb7eab1bd6d6d65f685ef414/diff/var/lib/dpkg/info/python3-blinker.prerm
evals	Retirar Python del runner	2026-09-15T04:51:17.7907629Z retirado: /var/lib/docker/overlay2/a9853d7ac0752db045c472a9505e498ad787cec8cb7eab1bd6d6d65f685ef414/diff/var/lib/dpkg/info/python3-oauthlib.postinst
evals	Retirar Python del runner	2026-09-15T04:51:17.8047562Z retirado: /var/lib/docker/overlay2/a9853d7ac0752db045c472a9505e498ad787cec8cb7eab1bd6d6d65f685ef414/diff/var/lib/dpkg/info/python3.prerm
evals	Retirar Python del runner	2026-09-15T04:51:17.8186803Z retirado: /var/lib/docker/overlay2/a9853d7ac0752db045c472a9505e498ad787cec8cb7eab1bd6d6d65f685ef414/diff/var/lib/dpkg/info/libpython3.12-stdlib:amd64.list
evals	Retirar Python del runner	2026-09-15T04:51:17.8326940Z retirado: /var/lib/docker/overlay2/a9853d7ac0752db045c472a9505e498ad787cec8cb7eab1bd6d6d65f685ef414/diff/var/lib/dpkg/info/python3-pyparsing.prerm
evals	Retirar Python del runner	2026-09-15T04:51:17.8465031Z retirado: /var/lib/docker/overlay2/a9853d7ac0752db045c472a9505e498ad787cec8cb7eab1bd6d6d65f685ef414/diff/var/lib/dpkg/info/python3-breezy.prerm
evals	Retirar Python del runner	2026-09-15T04:51:17.8604720Z retirado: /var/lib/docker/overlay2/a9853d7ac0752db045c472a9505e498ad787cec8cb7eab1bd6d6d65f685ef414/diff/var/lib/dpkg/info/python3-dulwich.postinst
evals	Retirar Python del runner	2026-09-15T04:51:17.8742861Z retirado: /var/lib/docker/overlay2/a9853d7ac0752db045c472a9505e498ad787cec8cb7eab1bd6d6d65f685ef414/diff/var/lib/dpkg/info/python3-configobj.prerm
evals	Retirar Python del runner	2026-09-15T04:51:17.8884532Z retirado: /var/lib/docker/overlay2/a9853d7ac0752db045c472a9505e498ad787cec8cb7eab1bd6d6d65f685ef414/diff/var/lib/dpkg/info/libpython3.12-minimal:amd64.postinst
evals	Retirar Python del runner	2026-09-15T04:51:17.9025082Z retirado: /var/lib/docker/overlay2/a9853d7ac0752db045c472a9505e498ad787cec8cb7eab1bd6d6d65f685ef414/diff/var/lib/dpkg/info/python3.12-minimal.prerm
evals	Retirar Python del runner	2026-09-15T04:51:17.9162841Z retirado: /var/lib/docker/overlay2/a9853d7ac0752db045c472a9505e498ad787cec8cb7eab1bd6d6d65f685ef414/diff/var/lib/dpkg/info/python3-dulwich.prerm
evals	Retirar Python del runner	2026-09-15T04:51:17.9303264Z retirado: /var/lib/docker/overlay2/a9853d7ac0752db045c472a9505e498ad787cec8cb7eab1bd6d6d65f685ef414/diff/var/lib/dpkg/info/python3.postrm
evals	Retirar Python del runner	2026-09-15T04:51:17.9441791Z retirado: /var/lib/docker/overlay2/a9853d7ac0752db045c472a9505e498ad787cec8cb7eab1bd6d6d65f685ef414/diff/var/lib/dpkg/info/python3-yaml.prerm
evals	Retirar Python del runner	2026-09-15T04:51:17.9586175Z retirado: /var/lib/docker/overlay2/a9853d7ac0752db045c472a9505e498ad787cec8cb7eab1bd6d6d65f685ef414/diff/var/lib/dpkg/info/python3-lazr.restfulclient.postinst
evals	Retirar Python del runner	2026-09-15T04:51:17.9728682Z retirado: /var/lib/docker/overlay2/a9853d7ac0752db045c472a9505e498ad787cec8cb7eab1bd6d6d65f685ef414/diff/var/lib/dpkg/info/python3-patiencediff.prerm
evals	Retirar Python del runner	2026-09-15T04:51:17.9866816Z retirado: /var/lib/docker/overlay2/a9853d7ac0752db045c472a9505e498ad787cec8cb7eab1bd6d6d65f685ef414/diff/var/lib/dpkg/info/python3-merge3.prerm
evals	Retirar Python del runner	2026-09-15T04:51:18.0006250Z retirado: /var/lib/docker/overlay2/a9853d7ac0752db045c472a9505e498ad787cec8cb7eab1bd6d6d65f685ef414/diff/var/lib/dpkg/info/libpython3.12t64:amd64.triggers
evals	Retirar Python del runner	2026-09-15T04:51:18.0144816Z retirado: /var/lib/docker/overlay2/a9853d7ac0752db045c472a9505e498ad787cec8cb7eab1bd6d6d65f685ef414/diff/var/lib/dpkg/info/python3-blinker.postinst
evals	Retirar Python del runner	2026-09-15T04:51:18.0285682Z retirado: /var/lib/docker/overlay2/a9853d7ac0752db045c472a9505e498ad787cec8cb7eab1bd6d6d65f685ef414/diff/var/lib/dpkg/info/python3-lazr.uri.prerm
evals	Retirar Python del runner	2026-09-15T04:51:18.0425656Z retirado: /var/lib/docker/overlay2/a9853d7ac0752db045c472a9505e498ad787cec8cb7eab1bd6d6d65f685ef414/diff/var/lib/dpkg/info/python3-six.prerm
evals	Retirar Python del runner	2026-09-15T04:51:18.0566524Z retirado: /var/lib/docker/overlay2/a9853d7ac0752db045c472a9505e498ad787cec8cb7eab1bd6d6d65f685ef414/diff/var/lib/dpkg/info/python3-fastbencode.postinst
evals	Retirar Python del runner	2026-09-15T04:51:18.0706765Z retirado: /var/lib/docker/overlay2/a9853d7ac0752db045c472a9505e498ad787cec8cb7eab1bd6d6d65f685ef414/diff/var/lib/dpkg/info/libpython3.12-minimal:amd64.list
evals	Retirar Python del runner	2026-09-15T04:51:18.0847161Z retirado: /var/lib/docker/overlay2/a9853d7ac0752db045c472a9505e498ad787cec8cb7eab1bd6d6d65f685ef414/diff/var/lib/dpkg/info/libpython3.12t64:amd64.shlibs
evals	Retirar Python del runner	2026-09-15T04:51:18.0987182Z retirado: /var/lib/docker/overlay2/a9853d7ac0752db045c472a9505e498ad787cec8cb7eab1bd6d6d65f685ef414/diff/var/lib/dpkg/info/python3-breezy.postinst
evals	Retirar Python del runner	2026-09-15T04:51:18.1127937Z retirado: /var/lib/docker/overlay2/a9853d7ac0752db045c472a9505e498ad787cec8cb7eab1bd6d6d65f685ef414/diff/var/lib/dpkg/info/python3-oauthlib.prerm
evals	Retirar Python del runner	2026-09-15T04:51:18.1267656Z retirado: /var/lib/docker/overlay2/a9853d7ac0752db045c472a9505e498ad787cec8cb7eab1bd6d6d65f685ef414/diff/var/lib/dpkg/info/libpython3.12-minimal:amd64.conffiles
evals	Retirar Python del runner	2026-09-15T04:51:18.1410770Z retirado: /var/lib/docker/overlay2/a9853d7ac0752db045c472a9505e498ad787cec8cb7eab1bd6d6d65f685ef414/diff/var/lib/dpkg/info/python3.postinst
evals	Retirar Python del runner	2026-09-15T04:51:18.1550728Z retirado: /var/lib/docker/overlay2/a9853d7ac0752db045c472a9505e498ad787cec8cb7eab1bd6d6d65f685ef414/diff/usr/lib/x86_64-linux-gnu/libpython3.12.so.1
evals	Retirar Python del runner	2026-09-15T04:51:18.1692629Z retirado: /var/lib/docker/overlay2/a9853d7ac0752db045c472a9505e498ad787cec8cb7eab1bd6d6d65f685ef414/diff/usr/lib/x86_64-linux-gnu/libpython3.12.so.1.0
evals	Retirar Python del runner	2026-09-15T04:51:18.1890543Z retirado: /var/lib/docker/overlay2/a9853d7ac0752db045c472a9505e498ad787cec8cb7eab1bd6d6d65f685ef414/diff/usr/lib/python3.12/config-3.12-x86_64-linux-gnu/libpython3.12.so
evals	Retirar Python del runner	2026-09-15T04:51:18.2029604Z retirado: /var/lib/docker/overlay2/a9853d7ac0752db045c472a9505e498ad787cec8cb7eab1bd6d6d65f685ef414/diff/usr/share/pixmaps/python3.xpm
evals	Retirar Python del runner	2026-09-15T04:51:18.2167573Z retirado: /var/lib/docker/overlay2/a9853d7ac0752db045c472a9505e498ad787cec8cb7eab1bd6d6d65f685ef414/diff/usr/share/doc/libpython3.12t64
evals	Retirar Python del runner	2026-09-15T04:51:18.2368343Z retirado: /var/lib/docker/overlay2/a9853d7ac0752db045c472a9505e498ad787cec8cb7eab1bd6d6d65f685ef414/diff/usr/share/doc/python3.12/python-policy.txt.gz
evals	Retirar Python del runner	2026-09-15T04:51:18.2506157Z retirado: /var/lib/docker/overlay2/a9853d7ac0752db045c472a9505e498ad787cec8cb7eab1bd6d6d65f685ef414/diff/usr/share/doc/libpython3.12-stdlib
evals	Retirar Python del runner	2026-09-15T04:51:18.2647103Z retirado: /var/lib/docker/overlay2/a9853d7ac0752db045c472a9505e498ad787cec8cb7eab1bd6d6d65f685ef414/diff/usr/share/lintian/overrides/libpython3.12t64
evals	Retirar Python del runner	2026-09-15T04:51:18.2791146Z retirado: /var/lib/docker/overlay2/a9853d7ac0752db045c472a9505e498ad787cec8cb7eab1bd6d6d65f685ef414/diff/usr/share/lintian/overrides/libpython3.12-minimal
evals	Retirar Python del runner	2026-09-15T04:51:18.2929677Z retirado: /var/lib/docker/overlay2/a9853d7ac0752db045c472a9505e498ad787cec8cb7eab1bd6d6d65f685ef414/diff/usr/share/lintian/overrides/libpython3.12-stdlib
evals	Retirar Python del runner	2026-09-15T04:51:18.3203596Z retirado: /var/lib/docker/overlay2/a9853d7ac0752db045c472a9505e498ad787cec8cb7eab1bd6d6d65f685ef414/diff/usr
evals	Retirar Python del runner	2026-09-15T04:51:18.5808732Z retirado: /var/lib/docker/overlay2/80c613159e0ff38a882acb63fd4cf342723381c4f07d119df138a79409f5e793/diff/var/lib/dpkg/info/python3.10-minimal.prerm
evals	Retirar Python del runner	2026-09-15T04:51:18.5949343Z retirado: /var/lib/docker/overlay2/80c613159e0ff38a882acb63fd4cf342723381c4f07d119df138a79409f5e793/diff/var/lib/dpkg/info/libpython3.10-minimal:amd64.postinst
evals	Retirar Python del runner	2026-09-15T04:51:18.6085506Z retirado: /var/lib/docker/overlay2/80c613159e0ff38a882acb63fd4cf342723381c4f07d119df138a79409f5e793/diff/var/lib/dpkg/info/python3-minimal.prerm
evals	Retirar Python del runner	2026-09-15T04:51:18.6225815Z retirado: /var/lib/docker/overlay2/80c613159e0ff38a882acb63fd4cf342723381c4f07d119df138a79409f5e793/diff/var/lib/dpkg/info/libpython3.10-minimal:amd64.postrm
evals	Retirar Python del runner	2026-09-15T04:51:18.6370567Z retirado: /var/lib/docker/overlay2/80c613159e0ff38a882acb63fd4cf342723381c4f07d119df138a79409f5e793/diff/var/lib/dpkg/info/python3.preinst
evals	Retirar Python del runner	2026-09-15T04:51:18.6509881Z retirado: /var/lib/docker/overlay2/80c613159e0ff38a882acb63fd4cf342723381c4f07d119df138a79409f5e793/diff/var/lib/dpkg/info/libpython3-stdlib:amd64.list
evals	Retirar Python del runner	2026-09-15T04:51:18.6649782Z retirado: /var/lib/docker/overlay2/80c613159e0ff38a882acb63fd4cf342723381c4f07d119df138a79409f5e793/diff/var/lib/dpkg/info/libpython3.10-minimal:amd64.prerm
evals	Retirar Python del runner	2026-09-15T04:51:18.6788721Z retirado: /var/lib/docker/overlay2/80c613159e0ff38a882acb63fd4cf342723381c4f07d119df138a79409f5e793/diff/var/lib/dpkg/info/python3.10-minimal.postrm
evals	Retirar Python del runner	2026-09-15T04:51:18.6927974Z retirado: /var/lib/docker/overlay2/80c613159e0ff38a882acb63fd4cf342723381c4f07d119df138a79409f5e793/diff/var/lib/dpkg/info/libpython3.10-minimal:amd64.list
evals	Retirar Python del runner	2026-09-15T04:51:18.7066116Z retirado: /var/lib/docker/overlay2/80c613159e0ff38a882acb63fd4cf342723381c4f07d119df138a79409f5e793/diff/var/lib/dpkg/info/python3.10.prerm
evals	Retirar Python del runner	2026-09-15T04:51:18.7203081Z retirado: /var/lib/docker/overlay2/80c613159e0ff38a882acb63fd4cf342723381c4f07d119df138a79409f5e793/diff/var/lib/dpkg/info/libpython3-stdlib:amd64.md5sums
evals	Retirar Python del runner	2026-09-15T04:51:18.7340277Z retirado: /var/lib/docker/overlay2/80c613159e0ff38a882acb63fd4cf342723381c4f07d119df138a79409f5e793/diff/var/lib/dpkg/info/python3-minimal.postinst
evals	Retirar Python del runner	2026-09-15T04:51:18.7478324Z retirado: /var/lib/docker/overlay2/80c613159e0ff38a882acb63fd4cf342723381c4f07d119df138a79409f5e793/diff/var/lib/dpkg/info/libpython3.10-stdlib:amd64.list
evals	Retirar Python del runner	2026-09-15T04:51:18.7618474Z retirado: /var/lib/docker/overlay2/80c613159e0ff38a882acb63fd4cf342723381c4f07d119df138a79409f5e793/diff/var/lib/dpkg/info/python3.10-minimal.postinst
evals	Retirar Python del runner	2026-09-15T04:51:18.7755650Z retirado: /var/lib/docker/overlay2/80c613159e0ff38a882acb63fd4cf342723381c4f07d119df138a79409f5e793/diff/var/lib/dpkg/info/python3.10.postinst
evals	Retirar Python del runner	2026-09-15T04:51:18.7893375Z retirado: /var/lib/docker/overlay2/80c613159e0ff38a882acb63fd4cf342723381c4f07d119df138a79409f5e793/diff/var/lib/dpkg/info/python3.prerm
evals	Retirar Python del runner	2026-09-15T04:51:18.8032792Z retirado: /var/lib/docker/overlay2/80c613159e0ff38a882acb63fd4cf342723381c4f07d119df138a79409f5e793/diff/var/lib/dpkg/info/libpython3.10-stdlib:amd64.prerm
evals	Retirar Python del runner	2026-09-15T04:51:18.8169355Z retirado: /var/lib/docker/overlay2/80c613159e0ff38a882acb63fd4cf342723381c4f07d119df138a79409f5e793/diff/var/lib/dpkg/info/libpython3.10-minimal:amd64.md5sums
evals	Retirar Python del runner	2026-09-15T04:51:18.8308023Z retirado: /var/lib/docker/overlay2/80c613159e0ff38a882acb63fd4cf342723381c4f07d119df138a79409f5e793/diff/var/lib/dpkg/info/libpython3.10-stdlib:amd64.md5sums
evals	Retirar Python del runner	2026-09-15T04:51:18.8447798Z retirado: /var/lib/docker/overlay2/80c613159e0ff38a882acb63fd4cf342723381c4f07d119df138a79409f5e793/diff/var/lib/dpkg/info/python3.postrm
evals	Retirar Python del runner	2026-09-15T04:51:18.8586151Z retirado: /var/lib/docker/overlay2/80c613159e0ff38a882acb63fd4cf342723381c4f07d119df138a79409f5e793/diff/var/lib/dpkg/info/libpython3.10-minimal:amd64.conffiles
evals	Retirar Python del runner	2026-09-15T04:51:18.8724481Z retirado: /var/lib/docker/overlay2/80c613159e0ff38a882acb63fd4cf342723381c4f07d119df138a79409f5e793/diff/var/lib/dpkg/info/python3.10-minimal.preinst
evals	Retirar Python del runner	2026-09-15T04:51:18.8863445Z retirado: /var/lib/docker/overlay2/80c613159e0ff38a882acb63fd4cf342723381c4f07d119df138a79409f5e793/diff/var/lib/dpkg/info/python3.postinst
evals	Retirar Python del runner	2026-09-15T04:51:18.9000559Z retirado: /var/lib/docker/overlay2/80c613159e0ff38a882acb63fd4cf342723381c4f07d119df138a79409f5e793/diff/usr/share/pixmaps/python3.xpm
evals	Retirar Python del runner	2026-09-15T04:51:18.9139703Z retirado: /var/lib/docker/overlay2/80c613159e0ff38a882acb63fd4cf342723381c4f07d119df138a79409f5e793/diff/usr/share/doc/libpython3.10-stdlib
evals	Retirar Python del runner	2026-09-15T04:51:18.9335888Z retirado: /var/lib/docker/overlay2/80c613159e0ff38a882acb63fd4cf342723381c4f07d119df138a79409f5e793/diff/usr/share/doc/python3.10/python-policy.txt.gz
evals	Retirar Python del runner	2026-09-15T04:51:18.9474272Z retirado: /var/lib/docker/overlay2/80c613159e0ff38a882acb63fd4cf342723381c4f07d119df138a79409f5e793/diff/usr/share/lintian/overrides/libpython3.10-stdlib
evals	Retirar Python del runner	2026-09-15T04:51:18.9614205Z retirado: /var/lib/docker/overlay2/80c613159e0ff38a882acb63fd4cf342723381c4f07d119df138a79409f5e793/diff/usr/share/lintian/overrides/libpython3.10-minimal
evals	Retirar Python del runner	2026-09-15T04:51:18.9885913Z retirado: /var/lib/docker/overlay2/80c613159e0ff38a882acb63fd4cf342723381c4f07d119df138a79409f5e793/diff/usr
evals	Retirar Python del runner	2026-09-15T04:51:20.7181211Z retirado: /usr/lib/x86_64-linux-gnu/pkgconfig/python3.pc
evals	Retirar Python del runner	2026-09-15T04:51:20.7334325Z retirado: /usr/lib/x86_64-linux-gnu/pkgconfig/python3-embed.pc
evals	Retirar Python del runner	2026-09-15T04:51:20.7487348Z retirado: /usr/lib/x86_64-linux-gnu/libpython3.12.so.1
evals	Retirar Python del runner	2026-09-15T04:51:20.7705524Z retirado: /usr/lib/x86_64-linux-gnu/libpython3.12.so
evals	Retirar Python del runner	2026-09-15T04:51:20.7858426Z retirado: /usr/lib/x86_64-linux-gnu/libpython3.12.a
evals	Retirar Python del runner	2026-09-15T04:51:20.8010086Z retirado: /usr/lib/x86_64-linux-gnu/libpython3.12.so.1.0
evals	Retirar Python del runner	2026-09-15T04:51:20.8226690Z retirado: /usr/lib/python3.12/config-3.12-x86_64-linux-gnu/libpython3.12.so
evals	Retirar Python del runner	2026-09-15T04:51:20.8377579Z retirado: /usr/lib/python3.12/config-3.12-x86_64-linux-gnu/libpython3.12.a
evals	Retirar Python del runner	2026-09-15T04:51:20.8531209Z retirado: /usr/lib/python3.12/config-3.12-x86_64-linux-gnu/python-config.py
evals	Retirar Python del runner	2026-09-15T04:51:20.8684804Z retirado: /usr/lib/python3.12/config-3.12-x86_64-linux-gnu/libpython3.12-pic.a
evals	Retirar Python del runner	2026-09-15T04:51:20.8839021Z retirado: /usr/lib/google-cloud-sdk/lib/googlecloudsdk/command_lib/orchestration_pipelines/tools/python_environment_unpack.sh
evals	Retirar Python del runner	2026-09-15T04:51:20.8992439Z retirado: /usr/lib/google-cloud-sdk/platform/bundledpythonunix/lib/libpython3.so
evals	Retirar Python del runner	2026-09-15T04:51:20.9146589Z retirado: /usr/lib/google-cloud-sdk/platform/bundledpythonunix/lib/pkgconfig/python3.pc
evals	Retirar Python del runner	2026-09-15T04:51:20.9300301Z retirado: /usr/lib/google-cloud-sdk/platform/bundledpythonunix/lib/pkgconfig/python3-embed.pc
evals	Retirar Python del runner	2026-09-15T04:51:20.9454005Z retirado: /usr/lib/google-cloud-sdk/platform/bundledpythonunix/lib/python3.14/config-3.14-x86_64-linux-gnu/python-config.py
evals	Retirar Python del runner	2026-09-15T04:51:20.9603555Z retirado: /usr/lib/google-cloud-sdk/platform/bundledpythonunix/lib/libpython3.14.so
evals	Retirar Python del runner	2026-09-15T04:51:20.9744673Z retirado: /usr/lib/google-cloud-sdk/platform/bundledpythonunix/lib/libpython3.14.so.1.0
evals	Retirar Python del runner	2026-09-15T04:51:21.0014275Z retirado: /usr/lib/google-cloud-sdk/platform/bundledpythonunix
evals	Retirar Python del runner	2026-09-15T04:51:21.1918377Z retirado: /usr/lib/rpm/pythondistdeps.py
evals	Retirar Python del runner	2026-09-15T04:51:21.2058113Z retirado: /usr/local/aws-cli/v2/2.36.40/dist/libpython3.14.so.1.0
evals	Retirar Python del runner	2026-09-15T04:51:21.2201081Z retirado: /usr/local/lib/android/sdk/ndk/28.2.13676358/toolchains/llvm/prebuilt/linux-x86_64/python3/lib/libpython3.so
evals	Retirar Python del runner	2026-09-15T04:51:21.2341906Z retirado: /usr/local/lib/android/sdk/ndk/28.2.13676358/toolchains/llvm/prebuilt/linux-x86_64/python3/lib/pkgconfig/python3.pc
evals	Retirar Python del runner	2026-09-15T04:51:21.2481095Z retirado: /usr/local/lib/android/sdk/ndk/28.2.13676358/toolchains/llvm/prebuilt/linux-x86_64/python3/lib/pkgconfig/python3-embed.pc
evals	Retirar Python del runner	2026-09-15T04:51:21.2620971Z retirado: /usr/local/lib/android/sdk/ndk/28.2.13676358/toolchains/llvm/prebuilt/linux-x86_64/python3/lib/pkgconfig/python-3.11-embed.pc
evals	Retirar Python del runner	2026-09-15T04:51:21.2760244Z retirado: /usr/local/lib/android/sdk/ndk/28.2.13676358/toolchains/llvm/prebuilt/linux-x86_64/python3/lib/pkgconfig/python-3.11.pc
evals	Retirar Python del runner	2026-09-15T04:51:21.2901731Z retirado: /usr/local/lib/android/sdk/ndk/28.2.13676358/toolchains/llvm/prebuilt/linux-x86_64/python3/lib/python3.11/site-packages/pip/_vendor/pygments/lexers/python.py
evals	Retirar Python del runner	2026-09-15T04:51:21.3039638Z retirado: /usr/local/lib/android/sdk/ndk/28.2.13676358/toolchains/llvm/prebuilt/linux-x86_64/python3/lib/libpython3.11.so
evals	Retirar Python del runner	2026-09-15T04:51:21.3177521Z retirado: /usr/local/lib/android/sdk/ndk/28.2.13676358/toolchains/llvm/prebuilt/linux-x86_64/python3/lib/libpython3.11.so.1.0
evals	Retirar Python del runner	2026-09-15T04:51:21.3318239Z retirado: /usr/local/lib/android/sdk/ndk/28.2.13676358/toolchains/llvm/prebuilt/linux-x86_64/python3/include/python3.11/cpython/pythonrun.h
evals	Retirar Python del runner	2026-09-15T04:51:21.3460220Z retirado: /usr/local/lib/android/sdk/ndk/28.2.13676358/toolchains/llvm/prebuilt/linux-x86_64/python3/include/python3.11/Python.h
evals	Retirar Python del runner	2026-09-15T04:51:21.3598561Z retirado: /usr/local/lib/android/sdk/ndk/28.2.13676358/toolchains/llvm/prebuilt/linux-x86_64/python3/include/python3.11/pythonrun.h
evals	Retirar Python del runner	2026-09-15T04:51:21.3737375Z retirado: /usr/local/lib/android/sdk/ndk/28.2.13676358/toolchains/llvm/prebuilt/linux-x86_64/python3/share/man/man1/python3.1
evals	Retirar Python del runner	2026-09-15T04:51:21.3876004Z retirado: /usr/local/lib/android/sdk/ndk/28.2.13676358/toolchains/llvm/prebuilt/linux-x86_64/python3/share/man/man1/python3.11.1
evals	Retirar Python del runner	2026-09-15T04:51:21.4148652Z retirado: /usr/local/lib/android/sdk/ndk/28.2.13676358/toolchains/llvm/prebuilt/linux-x86_64/python3
evals	Retirar Python del runner	2026-09-15T04:51:21.4739324Z retirado: /usr/local/lib/android/sdk/ndk/27.3.13750724/toolchains/llvm/prebuilt/linux-x86_64/python3/lib/libpython3.so
evals	Retirar Python del runner	2026-09-15T04:51:21.4883642Z retirado: /usr/local/lib/android/sdk/ndk/27.3.13750724/toolchains/llvm/prebuilt/linux-x86_64/python3/lib/pkgconfig/python3.pc
evals	Retirar Python del runner	2026-09-15T04:51:21.5024362Z retirado: /usr/local/lib/android/sdk/ndk/27.3.13750724/toolchains/llvm/prebuilt/linux-x86_64/python3/lib/pkgconfig/python3-embed.pc
evals	Retirar Python del runner	2026-09-15T04:51:21.5162348Z retirado: /usr/local/lib/android/sdk/ndk/27.3.13750724/toolchains/llvm/prebuilt/linux-x86_64/python3/lib/pkgconfig/python-3.11-embed.pc
evals	Retirar Python del runner	2026-09-15T04:51:21.5303812Z retirado: /usr/local/lib/android/sdk/ndk/27.3.13750724/toolchains/llvm/prebuilt/linux-x86_64/python3/lib/pkgconfig/python-3.11.pc
evals	Retirar Python del runner	2026-09-15T04:51:21.5442972Z retirado: /usr/local/lib/android/sdk/ndk/27.3.13750724/toolchains/llvm/prebuilt/linux-x86_64/python3/lib/python3.11/site-packages/pip/_vendor/pygments/lexers/python.py
evals	Retirar Python del runner	2026-09-15T04:51:21.5583153Z retirado: /usr/local/lib/android/sdk/ndk/27.3.13750724/toolchains/llvm/prebuilt/linux-x86_64/python3/lib/libpython3.11.so
evals	Retirar Python del runner	2026-09-15T04:51:21.5722976Z retirado: /usr/local/lib/android/sdk/ndk/27.3.13750724/toolchains/llvm/prebuilt/linux-x86_64/python3/lib/libpython3.11.so.1.0
evals	Retirar Python del runner	2026-09-15T04:51:21.5864338Z retirado: /usr/local/lib/android/sdk/ndk/27.3.13750724/toolchains/llvm/prebuilt/linux-x86_64/python3/include/python3.11/cpython/pythonrun.h
evals	Retirar Python del runner	2026-09-15T04:51:21.6006453Z retirado: /usr/local/lib/android/sdk/ndk/27.3.13750724/toolchains/llvm/prebuilt/linux-x86_64/python3/include/python3.11/Python.h
evals	Retirar Python del runner	2026-09-15T04:51:21.6144803Z retirado: /usr/local/lib/android/sdk/ndk/27.3.13750724/toolchains/llvm/prebuilt/linux-x86_64/python3/include/python3.11/pythonrun.h
evals	Retirar Python del runner	2026-09-15T04:51:21.6286026Z retirado: /usr/local/lib/android/sdk/ndk/27.3.13750724/toolchains/llvm/prebuilt/linux-x86_64/python3/share/man/man1/python3.1
evals	Retirar Python del runner	2026-09-15T04:51:21.6426200Z retirado: /usr/local/lib/android/sdk/ndk/27.3.13750724/toolchains/llvm/prebuilt/linux-x86_64/python3/share/man/man1/python3.11.1
evals	Retirar Python del runner	2026-09-15T04:51:21.6702987Z retirado: /usr/local/lib/android/sdk/ndk/27.3.13750724/toolchains/llvm/prebuilt/linux-x86_64/python3
evals	Retirar Python del runner	2026-09-15T04:51:21.7299845Z retirado: /usr/local/lib/android/sdk/ndk/29.0.14206865/toolchains/llvm/prebuilt/linux-x86_64/python3/lib/libpython3.so
evals	Retirar Python del runner	2026-09-15T04:51:21.7442554Z retirado: /usr/local/lib/android/sdk/ndk/29.0.14206865/toolchains/llvm/prebuilt/linux-x86_64/python3/lib/pkgconfig/python3.pc
evals	Retirar Python del runner	2026-09-15T04:51:21.7586052Z retirado: /usr/local/lib/android/sdk/ndk/29.0.14206865/toolchains/llvm/prebuilt/linux-x86_64/python3/lib/pkgconfig/python3-embed.pc
evals	Retirar Python del runner	2026-09-15T04:51:21.7725366Z retirado: /usr/local/lib/android/sdk/ndk/29.0.14206865/toolchains/llvm/prebuilt/linux-x86_64/python3/lib/pkgconfig/python-3.11-embed.pc
evals	Retirar Python del runner	2026-09-15T04:51:21.7866050Z retirado: /usr/local/lib/android/sdk/ndk/29.0.14206865/toolchains/llvm/prebuilt/linux-x86_64/python3/lib/pkgconfig/python-3.11.pc
evals	Retirar Python del runner	2026-09-15T04:51:21.8008332Z retirado: /usr/local/lib/android/sdk/ndk/29.0.14206865/toolchains/llvm/prebuilt/linux-x86_64/python3/lib/python3.11/site-packages/pip/_vendor/pygments/lexers/python.py
evals	Retirar Python del runner	2026-09-15T04:51:21.8147368Z retirado: /usr/local/lib/android/sdk/ndk/29.0.14206865/toolchains/llvm/prebuilt/linux-x86_64/python3/lib/libpython3.11.so
evals	Retirar Python del runner	2026-09-15T04:51:21.8287789Z retirado: /usr/local/lib/android/sdk/ndk/29.0.14206865/toolchains/llvm/prebuilt/linux-x86_64/python3/lib/libpython3.11.so.1.0
evals	Retirar Python del runner	2026-09-15T04:51:21.8427670Z retirado: /usr/local/lib/android/sdk/ndk/29.0.14206865/toolchains/llvm/prebuilt/linux-x86_64/python3/include/python3.11/cpython/pythonrun.h
evals	Retirar Python del runner	2026-09-15T04:51:21.8568584Z retirado: /usr/local/lib/android/sdk/ndk/29.0.14206865/toolchains/llvm/prebuilt/linux-x86_64/python3/include/python3.11/Python.h
evals	Retirar Python del runner	2026-09-15T04:51:21.8706747Z retirado: /usr/local/lib/android/sdk/ndk/29.0.14206865/toolchains/llvm/prebuilt/linux-x86_64/python3/include/python3.11/pythonrun.h
evals	Retirar Python del runner	2026-09-15T04:51:21.8847948Z retirado: /usr/local/lib/android/sdk/ndk/29.0.14206865/toolchains/llvm/prebuilt/linux-x86_64/python3/share/man/man1/python3.1
evals	Retirar Python del runner	2026-09-15T04:51:21.8988797Z retirado: /usr/local/lib/android/sdk/ndk/29.0.14206865/toolchains/llvm/prebuilt/linux-x86_64/python3/share/man/man1/python3.11.1
evals	Retirar Python del runner	2026-09-15T04:51:21.9260649Z retirado: /usr/local/lib/android/sdk/ndk/29.0.14206865/toolchains/llvm/prebuilt/linux-x86_64/python3
evals	Retirar Python del runner	2026-09-15T04:51:21.9857903Z retirado: /usr/local/aws-sam-cli/1.166.1/dist/_internal/libpython3.11.so.1.0
evals	Retirar Python del runner	2026-09-15T04:51:22.0001634Z retirado: /usr/local/share/vcpkg/ports/libudis86/python3.patch
evals	Retirar Python del runner	2026-09-15T04:51:22.0145084Z retirado: /usr/local/share/vcpkg/ports/omniorb/python-fixes.patch
evals	Retirar Python del runner	2026-09-15T04:51:22.0285654Z retirado: /usr/local/share/vcpkg/ports/openxr-loader/python3_8_compatibility.patch
evals	Retirar Python del runner	2026-09-15T04:51:22.0427375Z retirado: /usr/local/share/vcpkg/ports/libxslt/python3.patch
evals	Retirar Python del runner	2026-09-15T04:51:22.0567202Z retirado: /usr/local/share/vcpkg/ports/openscap/python-win32.diff
evals	Retirar Python del runner	2026-09-15T04:51:22.0707491Z retirado: /usr/local/share/vcpkg/ports/python3/python_vcpkg.props.in
evals	Retirar Python del runner	2026-09-15T04:51:22.0849052Z retirado: /usr/local/share/vcpkg/ports/vtk/pythonwrapper.patch
evals	Retirar Python del runner	2026-09-15T04:51:22.0986886Z retirado: /usr/local/share/vcpkg/scripts/test_ports/vcpkg-ci-blender/python.patch
evals	Retirar Python del runner	2026-09-15T04:51:22.1126374Z retirado: /usr/local/share/vcpkg/versions/p-/python2.json
evals	Retirar Python del runner	2026-09-15T04:51:22.1265921Z retirado: /usr/local/share/vcpkg/versions/p-/python3.json
evals	Retirar Python del runner	2026-09-15T04:51:22.1406316Z retirado: /usr/share/perl5/NeedRestart/Interp/Python.pm
evals	Retirar Python del runner	2026-09-15T04:51:22.1544556Z retirado: /usr/share/doc-base/python3.python-policy
evals	Retirar Python del runner	2026-09-15T04:51:22.1682726Z retirado: /usr/share/az_15.6.1/Az.Functions/4.3.2/Functions.Autorest/custom/FunctionsStackFlexData/EastAsia/python.json
evals	Retirar Python del runner	2026-09-15T04:51:22.1820504Z retirado: /usr/share/bash-completion/completions/python3.9
evals	Retirar Python del runner	2026-09-15T04:51:22.1958575Z retirado: /usr/share/bash-completion/completions/python3.7
evals	Retirar Python del runner	2026-09-15T04:51:22.2098145Z retirado: /usr/share/bash-completion/completions/python3.3
evals	Retirar Python del runner	2026-09-15T04:51:22.2238260Z retirado: /usr/share/bash-completion/completions/python3.6
evals	Retirar Python del runner	2026-09-15T04:51:22.2378044Z retirado: /usr/share/bash-completion/completions/python2.7
evals	Retirar Python del runner	2026-09-15T04:51:22.2517794Z retirado: /usr/share/bash-completion/completions/python3.4
evals	Retirar Python del runner	2026-09-15T04:51:22.2677639Z retirado: /usr/share/bash-completion/completions/pypy3
evals	Retirar Python del runner	2026-09-15T04:51:22.2817570Z retirado: /usr/share/bash-completion/completions/python2
evals	Retirar Python del runner	2026-09-15T04:51:22.2957712Z retirado: /usr/share/bash-completion/completions/pypy
evals	Retirar Python del runner	2026-09-15T04:51:22.3111462Z retirado: /usr/share/bash-completion/completions/python3.8
evals	Retirar Python del runner	2026-09-15T04:51:22.3259184Z retirado: /usr/share/bash-completion/completions/python3.5
evals	Retirar Python del runner	2026-09-15T04:51:22.3398302Z retirado: /usr/share/bash-completion/completions/python3
evals	Retirar Python del runner	2026-09-15T04:51:22.3538770Z retirado: /usr/share/bash-completion/completions/python
evals	Retirar Python del runner	2026-09-15T04:51:22.3679079Z retirado: /usr/share/bash-completion/helpers/python
evals	Retirar Python del runner	2026-09-15T04:51:22.3819575Z retirado: /usr/share/man/man8/pythoncalls-bpfcc.8.gz
evals	Retirar Python del runner	2026-09-15T04:51:22.3962613Z retirado: /usr/share/man/man8/pythonstat-bpfcc.8.gz
evals	Retirar Python del runner	2026-09-15T04:51:22.4103683Z retirado: /usr/share/man/man8/pythonflow-bpfcc.8.gz
evals	Retirar Python del runner	2026-09-15T04:51:22.4255993Z retirado: /usr/share/man/man8/pythongc-bpfcc.8.gz
evals	Retirar Python del runner	2026-09-15T04:51:22.4396248Z retirado: /usr/share/man/man1/python3.12.1.gz
evals	Retirar Python del runner	2026-09-15T04:51:22.4938790Z retirado: /usr/share/man/man1/python.1.gz
evals	Retirar Python del runner	2026-09-15T04:51:22.5079218Z retirado: /usr/share/man/man1/python3.12-config.1.gz
evals	Retirar Python del runner	2026-09-15T04:51:22.5277982Z retirado: /usr/share/man/man1/python3.1.gz
evals	Retirar Python del runner	2026-09-15T04:51:22.5520120Z retirado: /usr/share/man/man1/python3-config.1.gz
evals	Retirar Python del runner	2026-09-15T04:51:22.5628639Z retirado: /usr/share/pixmaps/python3.xpm
evals	Retirar Python del runner	2026-09-15T04:51:22.5774147Z retirado: /usr/share/pixmaps/python3.12.xpm
evals	Retirar Python del runner	2026-09-15T04:51:22.5917342Z retirado: /usr/share/binfmts/python3.12
evals	Retirar Python del runner	2026-09-15T04:51:22.6056710Z retirado: /usr/share/vim/vim91/syntax/python2.vim
evals	Retirar Python del runner	2026-09-15T04:51:22.6193504Z retirado: /usr/share/vim/vim91/syntax/python.vim
evals	Retirar Python del runner	2026-09-15T04:51:22.6331768Z retirado: /usr/share/vim/vim91/autoload/pythoncomplete.vim
evals	Retirar Python del runner	2026-09-15T04:51:22.6472526Z retirado: /usr/share/vim/vim91/autoload/python3complete.vim
evals	Retirar Python del runner	2026-09-15T04:51:22.6610384Z retirado: /usr/share/vim/vim91/autoload/python.vim
evals	Retirar Python del runner	2026-09-15T04:51:22.6745273Z retirado: /usr/share/vim/vim91/ftplugin/python.vim
evals	Retirar Python del runner	2026-09-15T04:51:22.6882314Z retirado: /usr/share/vim/vim91/indent/python.vim
evals	Retirar Python del runner	2026-09-15T04:51:22.7018407Z retirado: /usr/share/swig4.0/python/pythonkw.swg
evals	Retirar Python del runner	2026-09-15T04:51:22.7155981Z retirado: /usr/share/swig4.0/python/python.swg
evals	Retirar Python del runner	2026-09-15T04:51:22.7290933Z retirado: /usr/share/applications/python3.12.desktop
evals	Retirar Python del runner	2026-09-15T04:51:22.7426837Z retirado: /usr/share/doc/python3.12-venv
evals	Retirar Python del runner	2026-09-15T04:51:22.7562488Z retirado: /usr/share/doc/libpython3.12t64
evals	Retirar Python del runner	2026-09-15T04:51:22.7697127Z retirado: /usr/share/doc/python3-setuptools/python 2 sunset.rst
evals	Retirar Python del runner	2026-09-15T04:51:22.7832066Z retirado: /usr/share/doc/libpython3.12-dev
evals	Retirar Python del runner	2026-09-15T04:51:22.7967192Z retirado: /usr/share/doc/mercurial-common/examples/python-hook-examples.py
evals	Retirar Python del runner	2026-09-15T04:51:22.8105855Z retirado: /usr/share/doc/python3.12-dev
evals	Retirar Python del runner	2026-09-15T04:51:22.8242025Z retirado: /usr/share/doc/python3-pip/html/topics/python-option.md
evals	Retirar Python del runner	2026-09-15T04:51:22.8375849Z retirado: /usr/share/doc/python3-venv
evals	Retirar Python del runner	2026-09-15T04:51:22.8517138Z retirado: /usr/share/doc/python3.12/python-policy.txt.gz
evals	Retirar Python del runner	2026-09-15T04:51:22.8653149Z retirado: /usr/share/doc/libpython3.12-stdlib
evals	Retirar Python del runner	2026-09-15T04:51:22.8789831Z retirado: /usr/share/doc/python3-dev
evals	Retirar Python del runner	2026-09-15T04:51:22.8929032Z retirado: /usr/share/doc/bpfcc-tools/examples/doc/pythonstat_example.txt
evals	Retirar Python del runner	2026-09-15T04:51:22.9082167Z retirado: /usr/share/doc/bpfcc-tools/examples/doc/pythonflow_example.txt
evals	Retirar Python del runner	2026-09-15T04:51:22.9223370Z retirado: /usr/share/doc/bpfcc-tools/examples/doc/pythoncalls_example.txt
evals	Retirar Python del runner	2026-09-15T04:51:22.9364001Z retirado: /usr/share/doc/bpfcc-tools/examples/doc/pythongc_example.txt
evals	Retirar Python del runner	2026-09-15T04:51:22.9503553Z retirado: /usr/share/doc/python3-debconf
evals	Retirar Python del runner	2026-09-15T04:51:22.9643308Z retirado: /usr/share/doc/python3/python-policy.txt.gz
evals	Retirar Python del runner	2026-09-15T04:51:22.9785016Z retirado: /usr/share/doc/python3/python-policy.html
evals	Retirar Python del runner	2026-09-15T04:51:22.9923773Z retirado: /usr/share/aclocal-1.16/python.m4
evals	Retirar Python del runner	2026-09-15T04:51:23.0061352Z retirado: /usr/share/miniconda/lib/libpython3.so
evals	Retirar Python del runner	2026-09-15T04:51:23.0197687Z retirado: /usr/share/miniconda/lib/pkgconfig/python-3.14-embed.pc
evals	Retirar Python del runner	2026-09-15T04:51:23.0337009Z retirado: /usr/share/miniconda/lib/pkgconfig/python3.pc
evals	Retirar Python del runner	2026-09-15T04:51:23.0531538Z retirado: /usr/share/miniconda/lib/pkgconfig/python3-embed.pc
evals	Retirar Python del runner	2026-09-15T04:51:23.0670707Z retirado: /usr/share/miniconda/lib/pkgconfig/python-3.14.pc
evals	Retirar Python del runner	2026-09-15T04:51:23.0807162Z retirado: /usr/share/miniconda/lib/python3.14/site-packages/conda/common/path/__pycache__/python.cpython-314.pyc
evals	Retirar Python del runner	2026-09-15T04:51:23.0944587Z retirado: /usr/share/miniconda/lib/python3.14/site-packages/conda/common/path/python.py
evals	Retirar Python del runner	2026-09-15T04:51:23.1081697Z retirado: /usr/share/miniconda/lib/python3.14/site-packages/pygments/lexers/__pycache__/python.cpython-314.pyc
evals	Retirar Python del runner	2026-09-15T04:51:23.1220099Z retirado: /usr/share/miniconda/lib/python3.14/site-packages/pygments/lexers/python.py
evals	Retirar Python del runner	2026-09-15T04:51:23.1361453Z retirado: /usr/share/miniconda/lib/python3.14/site-packages/pip/_vendor/pygments/lexers/__pycache__/python.cpython-314.pyc
evals	Retirar Python del runner	2026-09-15T04:51:23.1499741Z retirado: /usr/share/miniconda/lib/python3.14/site-packages/pip/_vendor/pygments/lexers/python.py
evals	Retirar Python del runner	2026-09-15T04:51:23.1638437Z retirado: /usr/share/miniconda/lib/python3.14/site-packages/conda_pypi/__pycache__/python_paths.cpython-314.pyc
evals	Retirar Python del runner	2026-09-15T04:51:23.1773319Z retirado: /usr/share/miniconda/lib/python3.14/site-packages/conda_pypi/python_paths.py
evals	Retirar Python del runner	2026-09-15T04:51:23.1912311Z retirado: /usr/share/miniconda/lib/python3.14/config-3.14-x86_64-linux-gnu/__pycache__/python-config.cpython-314.pyc
evals	Retirar Python del runner	2026-09-15T04:51:23.2047369Z retirado: /usr/share/miniconda/lib/python3.14/config-3.14-x86_64-linux-gnu/python-config.py
evals	Retirar Python del runner	2026-09-15T04:51:23.2186581Z retirado: /usr/share/miniconda/lib/python3.14/config-3.14-x86_64-linux-gnu/python.o
evals	Retirar Python del runner	2026-09-15T04:51:23.2322959Z retirado: /usr/share/miniconda/lib/python3.14/idlelib/Icons/python.gif
evals	Retirar Python del runner	2026-09-15T04:51:23.2462147Z retirado: /usr/share/miniconda/lib/libpython3.14.so
evals	Retirar Python del runner	2026-09-15T04:51:23.2602282Z retirado: /usr/share/miniconda/lib/libpython3.14.so.1.0
evals	Retirar Python del runner	2026-09-15T04:51:23.2739840Z retirado: /usr/share/miniconda/include/python3.14/cpython/pythonrun.h
evals	Retirar Python del runner	2026-09-15T04:51:23.2877417Z retirado: /usr/share/miniconda/include/python3.14/Python.h
evals	Retirar Python del runner	2026-09-15T04:51:23.3017893Z retirado: /usr/share/miniconda/include/python3.14/pythonrun.h
evals	Retirar Python del runner	2026-09-15T04:51:23.3155939Z retirado: /usr/share/miniconda/share/man/man1/python3.1
evals	Retirar Python del runner	2026-09-15T04:51:23.3296366Z retirado: /usr/share/miniconda/share/man/man1/python3.14.1
evals	Retirar Python del runner	2026-09-15T04:51:23.3433882Z retirado: /usr/share/miniconda/conda-meta/python_abi-3.14-4_cp314.json
evals	Retirar Python del runner	2026-09-15T04:51:23.3572157Z retirado: /usr/share/miniconda/conda-meta/python-installer-1.0.1-py314h06a4308_0.json
evals	Retirar Python del runner	2026-09-15T04:51:23.3711151Z retirado: /usr/share/miniconda/conda-meta/python-build-1.5.1-py314h06a4308_0.json
evals	Retirar Python del runner	2026-09-15T04:51:23.3849451Z retirado: /usr/share/miniconda/conda-meta/python-3.14.7-h2bd7c14_101_cp314.json
evals	Retirar Python del runner	2026-09-15T04:51:23.3989414Z retirado: /usr/share/miniconda/conda-meta/python-dotenv-1.2.2-py314h06a4308_0.json
evals	Retirar Python del runner	2026-09-15T04:51:23.4128664Z retirado: /usr/share/miniconda/pkgs/python-build-1.5.1-py314h06a4308_0.conda
evals	Retirar Python del runner	2026-09-15T04:51:23.4398604Z retirado: /usr/share/miniconda/pkgs/python-build-1.5.1-py314h06a4308_0
evals	Retirar Python del runner	2026-09-15T04:51:23.4573387Z retirado: /usr/share/miniconda/pkgs/conda-26.7.1-py314h06a4308_0/lib/python3.14/site-packages/conda/common/path/__pycache__/python.cpython-314.pyc
evals	Retirar Python del runner	2026-09-15T04:51:23.4712298Z retirado: /usr/share/miniconda/pkgs/conda-26.7.1-py314h06a4308_0/lib/python3.14/site-packages/conda/common/path/python.py
evals	Retirar Python del runner	2026-09-15T04:51:23.4852116Z retirado: /usr/share/miniconda/pkgs/python-3.14.7-h2bd7c14_101_cp314.conda
evals	Retirar Python del runner	2026-09-15T04:51:23.4992228Z retirado: /usr/share/miniconda/pkgs/python-dotenv-1.2.2-py314h06a4308_0.conda
evals	Retirar Python del runner	2026-09-15T04:51:23.5130753Z retirado: /usr/share/miniconda/pkgs/libxcb-1.17.0-h9b100fa_0/info/recipe/python3.patch
evals	Retirar Python del runner	2026-09-15T04:51:23.5268119Z retirado: /usr/share/miniconda/pkgs/pip-26.2.1-pyh0d26453_0/site-packages/pip/_vendor/pygments/lexers/python.py
evals	Retirar Python del runner	2026-09-15T04:51:23.5408202Z retirado: /usr/share/miniconda/pkgs/python-3.14.7-h2bd7c14_101_cp314/lib/libpython3.so
evals	Retirar Python del runner	2026-09-15T04:51:23.5546291Z retirado: /usr/share/miniconda/pkgs/python-3.14.7-h2bd7c14_101_cp314/lib/pkgconfig/python-3.14-embed.pc
evals	Retirar Python del runner	2026-09-15T04:51:23.5685215Z retirado: /usr/share/miniconda/pkgs/python-3.14.7-h2bd7c14_101_cp314/lib/pkgconfig/python3.pc
evals	Retirar Python del runner	2026-09-15T04:51:23.5884179Z retirado: /usr/share/miniconda/pkgs/python-3.14.7-h2bd7c14_101_cp314/lib/pkgconfig/python3-embed.pc
evals	Retirar Python del runner	2026-09-15T04:51:23.6023240Z retirado: /usr/share/miniconda/pkgs/python-3.14.7-h2bd7c14_101_cp314/lib/pkgconfig/python-3.14.pc
evals	Retirar Python del runner	2026-09-15T04:51:23.6163866Z retirado: /usr/share/miniconda/pkgs/python-3.14.7-h2bd7c14_101_cp314/lib/python3.14/config-3.14-x86_64-linux-gnu/__pycache__/python-config.cpython-314.pyc
evals	Retirar Python del runner	2026-09-15T04:51:23.6304273Z retirado: /usr/share/miniconda/pkgs/python-3.14.7-h2bd7c14_101_cp314/lib/python3.14/config-3.14-x86_64-linux-gnu/python-config.py
evals	Retirar Python del runner	2026-09-15T04:51:23.6447299Z retirado: /usr/share/miniconda/pkgs/python-3.14.7-h2bd7c14_101_cp314/lib/python3.14/config-3.14-x86_64-linux-gnu/python.o
evals	Retirar Python del runner	2026-09-15T04:51:23.6589190Z retirado: /usr/share/miniconda/pkgs/python-3.14.7-h2bd7c14_101_cp314/lib/python3.14/idlelib/Icons/python.gif
evals	Retirar Python del runner	2026-09-15T04:51:23.6728562Z retirado: /usr/share/miniconda/pkgs/python-3.14.7-h2bd7c14_101_cp314/lib/libpython3.14.so
evals	Retirar Python del runner	2026-09-15T04:51:23.6871204Z retirado: /usr/share/miniconda/pkgs/python-3.14.7-h2bd7c14_101_cp314/lib/libpython3.14.so.1.0
evals	Retirar Python del runner	2026-09-15T04:51:23.7017043Z retirado: /usr/share/miniconda/pkgs/python-3.14.7-h2bd7c14_101_cp314/include/python3.14/cpython/pythonrun.h
evals	Retirar Python del runner	2026-09-15T04:51:23.7155409Z retirado: /usr/share/miniconda/pkgs/python-3.14.7-h2bd7c14_101_cp314/include/python3.14/Python.h
evals	Retirar Python del runner	2026-09-15T04:51:23.7297677Z retirado: /usr/share/miniconda/pkgs/python-3.14.7-h2bd7c14_101_cp314/include/python3.14/pythonrun.h
evals	Retirar Python del runner	2026-09-15T04:51:23.7436457Z retirado: /usr/share/miniconda/pkgs/python-3.14.7-h2bd7c14_101_cp314/share/man/man1/python3.1
evals	Retirar Python del runner	2026-09-15T04:51:23.7576858Z retirado: /usr/share/miniconda/pkgs/python-3.14.7-h2bd7c14_101_cp314/share/man/man1/python3.14.1
evals	Retirar Python del runner	2026-09-15T04:51:23.7850598Z retirado: /usr/share/miniconda/pkgs/python-3.14.7-h2bd7c14_101_cp314
evals	Retirar Python del runner	2026-09-15T04:51:23.8861612Z retirado: /usr/share/miniconda/pkgs/pygments-2.20.0-py314h06a4308_0/lib/python3.14/site-packages/pygments/lexers/__pycache__/python.cpython-314.pyc
evals	Retirar Python del runner	2026-09-15T04:51:23.9000786Z retirado: /usr/share/miniconda/pkgs/pygments-2.20.0-py314h06a4308_0/lib/python3.14/site-packages/pygments/lexers/python.py
evals	Retirar Python del runner	2026-09-15T04:51:23.9142366Z retirado: /usr/share/miniconda/pkgs/pygments-2.20.0-py314h06a4308_0/info/test/tests/support/python_lexer.py
evals	Retirar Python del runner	2026-09-15T04:51:23.9282986Z retirado: /usr/share/miniconda/pkgs/pygments-2.20.0-py314h06a4308_0/info/test/tests/examplefiles/make/python25-bsd.mak.output
evals	Retirar Python del runner	2026-09-15T04:51:23.9422069Z retirado: /usr/share/miniconda/pkgs/pygments-2.20.0-py314h06a4308_0/info/test/tests/examplefiles/make/python25-bsd.mak
evals	Retirar Python del runner	2026-09-15T04:51:23.9562959Z retirado: /usr/share/miniconda/pkgs/conda-pypi-0.11.0-py314h06a4308_0/lib/python3.14/site-packages/conda_pypi/__pycache__/python_paths.cpython-314.pyc
evals	Retirar Python del runner	2026-09-15T04:51:23.9703638Z retirado: /usr/share/miniconda/pkgs/conda-pypi-0.11.0-py314h06a4308_0/lib/python3.14/site-packages/conda_pypi/python_paths.py
evals	Retirar Python del runner	2026-09-15T04:51:23.9842523Z retirado: /usr/share/miniconda/pkgs/python-installer-1.0.1-py314h06a4308_0.conda
evals	Retirar Python del runner	2026-09-15T04:51:23.9984280Z retirado: /usr/share/miniconda/pkgs/python_abi-3.14-4_cp314.conda
evals	Retirar Python del runner	2026-09-15T04:51:24.0261054Z retirado: /usr/share/miniconda
evals	Retirar Python del runner	2026-09-15T04:51:24.9015505Z retirado: /usr/share/nano/python.nanorc
evals	Retirar Python del runner	2026-09-15T04:51:24.9157927Z retirado: /usr/share/lintian/overrides/python3-debian
evals	Retirar Python del runner	2026-09-15T04:51:24.9298250Z retirado: /usr/share/lintian/overrides/python3.12-venv
evals	Retirar Python del runner	2026-09-15T04:51:24.9437488Z retirado: /usr/share/lintian/overrides/python3-dbus
evals	Retirar Python del runner	2026-09-15T04:51:24.9579751Z retirado: /usr/share/lintian/overrides/libpython3.12t64
evals	Retirar Python del runner	2026-09-15T04:51:24.9717406Z retirado: /usr/share/lintian/overrides/libpython3.12-dev
evals	Retirar Python del runner	2026-09-15T04:51:24.9856417Z retirado: /usr/share/lintian/overrides/python3-pip
evals	Retirar Python del runner	2026-09-15T04:51:24.9996548Z retirado: /usr/share/lintian/overrides/libpython3.12-minimal
evals	Retirar Python del runner	2026-09-15T04:51:25.0134893Z retirado: /usr/share/lintian/overrides/python3.12-minimal
evals	Retirar Python del runner	2026-09-15T04:51:25.0275244Z retirado: /usr/share/lintian/overrides/python3.12
evals	Retirar Python del runner	2026-09-15T04:51:25.0417189Z retirado: /usr/share/lintian/overrides/libpython3.12-stdlib
evals	Retirar Python del runner	2026-09-15T04:51:25.0557480Z retirado: /usr/share/lintian/overrides/python3-netaddr
evals	Retirar Python del runner	2026-09-15T04:51:25.0696774Z retirado: /usr/share/lintian/overrides/python3
evals	Retirar Python del runner	2026-09-15T04:51:25.0835560Z retirado: /usr/share/lintian/overrides/python3-apt
evals	Retirar Python del runner	2026-09-15T04:51:25.0972873Z retirado: /usr/share/automake-1.16/am/python.am
evals	Retirar Python del runner	2026-09-15T04:51:25.1113533Z retirado: /usr/share/python3/bcep/python3-jinja2
evals	Retirar Python del runner	2026-09-15T04:51:25.1254520Z retirado: /usr/share/python3/dist/python3-cryptography
evals	Retirar Python del runner	2026-09-15T04:51:25.1397439Z retirado: /usr/share/python3/dist/python3-zope.interface
evals	Retirar Python del runner	2026-09-15T04:51:25.1537757Z retirado: /usr/share/python3/dist/python3-six
evals	Retirar Python del runner	2026-09-15T04:51:25.1674815Z retirado: /usr/share/python3/dist/python3-pyasn1
evals	Retirar Python del runner	2026-09-15T04:51:25.1814131Z retirado: /usr/share/python3/python.mk
evals	Retirar Python del runner	2026-09-15T04:51:25.1953952Z retirado: /usr/sbin/pythongc-bpfcc
evals	Retirar Python del runner	2026-09-15T04:51:25.2094710Z retirado: /usr/sbin/pythoncalls-bpfcc
evals	Retirar Python del runner	2026-09-15T04:51:25.2234707Z retirado: /usr/sbin/pythonstat-bpfcc
evals	Retirar Python del runner	2026-09-15T04:51:25.2371818Z retirado: /usr/sbin/pythonflow-bpfcc
evals	Retirar Python del runner	2026-09-15T04:51:25.2644906Z retirado: /usr/bin/python3.12-config
evals	Retirar Python del runner	2026-09-15T04:51:25.2975888Z retirado: /usr/bin/python3-config
evals	Retirar Python del runner	2026-09-15T04:51:25.3249213Z retirado: /usr/bin/python3.12
evals	Retirar Python del runner	2026-09-15T04:51:25.3580317Z retirado: /usr/bin/python3
evals	Retirar Python del runner	2026-09-15T04:51:25.3909468Z retirado: /usr/bin/python
evals	Retirar Python del runner	2026-09-15T04:51:29.3811164Z búsqueda tras retirar: ninguno
evals	Retirar Python del runner	2026-09-15T04:51:29.3811837Z --- fin de la retirada de Python ---
```
