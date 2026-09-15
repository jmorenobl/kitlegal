# Prueba de red de H5 (quickstart §12.2, SC-012)

Intento 6 de T030 (el tercero que cuenta el workflow, `gates/tareas-intentos.json` `T030: 3`, concedido por la
supervisión del run tras T042 y T043), 2026-09-15, propuesta de cambio [#27](https://github.com/jmorenobl/kitlegal/pull/27),
cabeza `a99295f50716010b560c1998ac4e71e62cd24da4` (`feat(H5): T043`). **Resultado: T042 y T043 hicieron lo que
debían** —las diez positivas leyeron el índice con 0 y el bloque que nombra la pregunta a la primera (la 02, `a1-30`
copiado del índice sin pedir ningún vecino; la 03, la 06 y la 08, que en el intento 5 pidieron otro artículo, esta vez
`a22`, `a17` y `a20`), y la 09, tras transcribir el artículo, citó sola en su línea `art. 140 de la Constitución
Española [BOE-A-1978-31229, bloque a140]`, la forma exacta de T042—, y **la prueba de red cumple lo que SC-012 espera
de ella en doce de las trece sesiones** (las dos filas de `a9998` en «fuera de lo grabado», con código 5 la que va sin
`--offline` y 4 la que lo lleva; la invocación sin `--offline` con una sola conexión, `127.0.0.1:9` de clase `local`, y
la que lleva `--offline` sin ninguna; «ninguna petición llegó a la red de una fuente»; la sesión de prueba de red con el
mismo resultado que la 01). **Pero el veredicto es `fallo`, por dos motivos nuevos, y uno de ellos es del job**: la
sesión 09, terminada con código 0 y no cortada, es `sesión ilegible` porque el primer renglón del fichero de uno de los
hilos de su invocación es `clone(child_stack=0x2a559d472000, flags=CLONE_VM|CLONE_FS|CLONE_FILES|CLONE_SIGHAND|CLONE_THREAD|CLONE_SYSVSEM|CLONE_SETTLS <unfinished ...>) = ?`,
la forma con la que `strace` escribe una llamada en curso cuando el proceso termina antes de que tenga resultado, que
ninguna sonda de research vio (V53 y V54 la daban por ausente en una sesión sin corte) y que `LeerTrazas` no admite: el
supuesto S4 falla ahí por primera vez; y la 08 leyó `a20` con 0 pero citó
`[art. 20.1 de la LTAIBG, BOE-A-2013-12887, bloque a20]`, con la forma legible dentro de los corchetes, la tercera sesión
distinta en tres intentos que la pone dentro pese a T040 y T042. La línea de T030 detiene la tarea sin marcarla:
diagnóstico en §4 y en `gates/tarea-T030.md`; arreglo en tres tareas nuevas, **T044** (de datos: la línea real en las
trazas sintéticas), **T045** (`LeerTrazas` admite la llamada que el fin del proceso deja sin resultado) y **T046** (la
parte mecánica de la cita admite la forma legible dentro de los corchetes; la persona la revisa). El paso «Retirar
Python del runner» terminó con 0. Los intentos 1 a 5 (ejecuciones 34922606273, 34930222593, 34936425178, 34941499481 y
34956596912) están en las versiones anteriores de este fichero (historial de git del fichero).

Enlace a la ejecución: <https://github.com/jmorenobl/kitlegal/actions/runs/34961757559> (`databaseId` 34961757559).

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
{"headRefOid":"a99295f50716010b560c1998ac4e71e62cd24da4","labels":[],"number":27}
```

Todo presente: el secreto y las dos etiquetas.

## 2. Órdenes de §12.2

**Primera** (la etiqueta no estaba puesta, porque el intento 5 la quitó al terminar, así que no se quitó nada):

```text
la etiqueta evals-prueba-de-red no está puesta
```

**Segunda**, `gh pr edit --add-label evals-prueba-de-red`, a las 11:09:23Z, con el paso directo del envoltorio de la
terminal:

```text
https://github.com/jmorenobl/kitlegal/pull/27
```

**Tercera**, a la primera ya con la ejecución:

```text
etiqueta puesta: 2026-09-15T11:09:25Z
{"evals":{"conclusion":"","createdAt":"2026-09-15T11:09:28Z","databaseId":34961757559,"headSha":"a99295f50716010b560c1998ac4e71e62cd24da4","status":"in_progress","url":"https://github.com/jmorenobl/kitlegal/actions/runs/34961757559","workflowName":"evals"},"posteriores_a_la_etiqueta":[{"createdAt":"2026-09-15T11:09:28Z","databaseId":34961757559,"workflowName":"evals"}]}
```

**Cuarta**, `gh run watch 34961757559 --exit-status`: la ejecución duró 8 m 26 s, dentro del tope de la herramienta de
la sesión (600 s), así que la orden terminó a la primera, `código 1`. Sus últimas líneas:

```text
X h5-skill-boe-legislacion evals jmorenobl/kitlegal#27 · 34961757559
Triggered via pull_request about 8 minutes ago

JOBS
X evals in 8m26s (ID 104356735853)
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
X Process completed with exit code 2.
evals: .github#1135

código 1
```

(El código 2 de la anotación es el de GNU `make` cuando una receta falla; la receta, `scripts/evals.sh`, terminó con 1,
como dice el registro: `make: *** [Makefile:112: evals] Error 1`.)

**Quinta** (informe entre marcas): **código 0**; imprime `informe.md` (611 líneas del registro) e `informe.json` (475)
enteros, cada uno de su marca de inicio a su marca de fin. Salida completa, tal cual la da `gh run view --log`, en el
**anexo A**. Para conservarla entera, la orden se ejecutó tal cual dentro de `rtk proxy sh -c` con su salida redirigida
a un fichero temporal del directorio del hito (borrado tras copiarla aquí) y con una línea final
`código de la quinta orden: $?` añadida detrás. La línea `make: *** [Makefile:112: evals] Error 1` que aparece dentro
de la tabla «Invocaciones fuera de lo grabado» del anexo, tras las dos filas de `a9998`, es la salida de error de
`make`, que el registro del paso intercala entre las líneas del informe (como en los intentos 4 y 5); no está en el
fichero `informe.md` (el informe JSON, que va después, no la contiene).

**Sexta** (salida de la retirada de Python entre marcas): **código 0**; 925 líneas del registro, de
`--- inicio de la retirada de Python ---` a `--- fin de la retirada de Python ---`. Salida completa en el **anexo B**,
obtenida de la misma forma.

La orden de `--log-failed` que va tras el bloque no procede (las dos anteriores no fallan); el registro completo
(`gh run view 34961757559 --log`, 2 469 líneas) se leyó aparte para §3.

**Séptima orden** (quitar la etiqueta, al terminar): a la primera, sin ningún error de la plataforma esta vez:

```text
https://github.com/jmorenobl/kitlegal/pull/27
la etiqueta evals-prueba-de-red no está puesta
```

`gh pr view 27 --json labels` después: `{"labels":[]}`; la API de eventos, `2026-09-15T11:20:54Z	unlabeled` (§5,
S12 (1)).

## 3. Lo que muestra la ejecución

### 3.1 Pasos y tiempos

`gh run view 34961757559 --json jobs` (`createdAt` 11:09:28Z, `event` `pull_request`, `headSha` `a99295f…`,
`workflowName` `evals`, job 104356735853, de 11:09:30 a 11:17:56):

| Paso | Resultado | Inicio | Fin | Duración |
|---|---|---|---|---|
| Obtener el código del commit evaluado | success | 11:09:32 | 11:09:34 | 2 s |
| Instalar Go y restaurar la caché | success | 11:09:34 | 11:10:01 | 27 s |
| Instalar strace y Claude Code | success | 11:10:02 | 11:10:15 | 13 s |
| Instalar kitlegal y las skills como las deja make install | success | 11:10:15 | 11:10:58 | 43 s |
| Retirar Python del runner | success | 11:10:58 | 11:13:19 | 2 m 21 s |
| Ejecutar las evals | **failure** | 11:13:19 | 11:17:54 | 4 m 35 s |

Del paso de instalación: `strace is already the newest version (6.8-0ubuntu2).`, `added 2 packages in 4s` y
`claude --version` → `2.1.270 (Claude Code)`; el paso no instala `bubblewrap` ni `socat` (T036). De `make install`:
`CGO_ENABLED=0 go install -trimpath -ldflags "-X main.version=a99295f …" ./cmd/kitlegal`,
`instalar-skills: boe-legislacion → /home/runner/work/kitlegal/kitlegal/skills/boe-legislacion` e
`instalar-skills: kitlegal → /home/runner/go/bin/kitlegal`. El entorno del paso de evals lista
`MODELO_DE_EVALS: claude-haiku-4-5-20251001`, `COMMIT_EVALUADO: a99295f…`, `PRUEBA_DE_RED: true` y
`CLAUDE_CODE_OAUTH_TOKEN: ***` (el secreto llega al paso, enmascarado).

### 3.2 Retirada de Python (anexo B)

- `búsqueda: find / ( -path /proc -o -path /sys ) -prune -o ( ( -type f -perm /111 ( -iname python* -o -iname pypy* ) ) -o ( -type l ( -iname python* -o -iname pypy* ) ) -o ( ( -type f -o -type l ) ( -iname libpython* -o -iname libpypy* ) ) ) -print`
  a las 11:10:58,23; la primera línea `retirado:` a las 11:12:49,93: **la búsqueda como root en toda la imagen tarda
  1 m 52 s** (6 m 7 s en el intento 5, 2 m 32 s en el 4, 2 m 37 s en el 3, 1 m 23 s en el 2: el tiempo lo pone el
  disco del runner que toque) y termina con 0 (con `set -euo pipefail`, un `find` con error habría detenido el paso ahí).
- **921 líneas `retirado:`**, exactamente las mismas rutas que en los intentos 3, 4 y 5 (comparadas una a una con el
  anexo B de la versión anterior de este fichero, ordenadas: ninguna diferencia; misma imagen), entre las 11:12:49 y
  las 11:13:15 (25 s), ninguna seguida de un error de `rm`: nada estaba en un sistema de ficheros de solo lectura. Por
  árbol: 341 bajo `/opt/hostedtoolcache`, 315 bajo `/var/lib`, 143 bajo `/usr/share`, 54 bajo `/usr/local`, 21 bajo
  `/opt/az`, 19 bajo `/usr/lib`, 19 bajo `/opt/pipx`, 5 en `/usr/bin` (`python3.12-config`, `python3-config`,
  `python3.12`, `python3`, `python`) y 4 en `/usr/sbin` (las herramientas `python*-bpfcc`).
- Instalaciones retiradas enteras por la regla del prefijo, las mismas catorce de los intentos 3 a 5: `/opt/az`;
  `/opt/hostedtoolcache/PyPy/3.9.19/x64`, `/opt/hostedtoolcache/PyPy/3.10.16/x64`,
  `/opt/hostedtoolcache/PyPy/3.11.15/x64`, `/opt/hostedtoolcache/Python/3.10.21/x64`,
  `/opt/hostedtoolcache/Python/3.11.16/x64`, `/opt/hostedtoolcache/Python/3.12.14/x64`,
  `/opt/hostedtoolcache/Python/3.13.15/x64`, `/opt/hostedtoolcache/Python/3.14.7/x64`; `/opt/pipx/shared`,
  `/opt/pipx/venvs/ansible-core`, `/opt/pipx/venvs/yamllint`; `/usr/lib/google-cloud-sdk/platform/bundledpythonunix`; y
  `/usr/share/miniconda`.
- `búsqueda tras retirar: ninguno` a las 11:13:19,70 (la segunda búsqueda, 4,6 s), la comprobación de lo usado sin
  ningún `la retirada se llevó algo que el job usa`, y la marca de fin. Código 0.

### 3.3 Ejecutar las evals (anexo A y registro del paso)

Fuera del informe, el registro del paso muestra: `scripts/evals.sh "boe-legislacion"`; la comprobación 5
(`TestEvalsDelRepositorio` y `TestIdentificadoresDeLasNormas`) en `ok` a las 11:13:37; trece preparaciones de sesión
(`TestPrepararSesion`) en `ok`, cada una seguida de su sesión, que terminó por sí misma: medido por el instante del `ok`
de la preparación siguiente, 01 ≈ 23 s, 02 ≈ 20 s, 03 ≈ 25 s, 04 ≈ 17 s, 05 ≈ 22 s, 06 ≈ 23 s, 07 ≈ 25 s, 08 ≈ 19 s,
09 ≈ 16 s, 10 ≈ 19 s, 11 ≈ 6 s, 12 ≈ 7 s y `01-lpac-articulo-21-prueba-de-red`, la última, ≈ 32 s, de las 11:13:39 a
las 11:17:54; `TestInformeDelJob` en `FAIL` con `expected: "aprobado"`, `actual: "fallo"` y los dos motivos; las
marcas y los dos ficheros; y `make: *** [Makefile:112: evals] Error 1`.

`sin_python` del informe (comprobación 3 del guion, repetida como root antes de la primera sesión):

```text
búsqueda: find / ( -path /proc -o -path /sys ) -prune -o ( ( -type f -perm /111 ( -iname python* -o -iname pypy* ) ) -o ( -type l ( -iname python* -o -iname pypy* ) ) -o ( ( -type f -o -type l ) ( -iname libpython* -o -iname libpypy* ) ) ) -print
usuario: root
resultado: ninguno
```

Cabecera del informe: `Modelo del job: claude-haiku-4-5-20251001`, `Modelos de las sesiones:
claude-haiku-4-5-20251001`, `Versiones de Claude Code: 2.1.270`, `Commit: a99295f50716010b560c1998ac4e71e62cd24da4`.
Ficheros mal formados: ninguno. Peticiones llegadas a la red: «ninguna petición llegó a la red de una fuente» (`red`
vacío en la raíz del JSON y `llegadas_a_la_red` vacío en las trece sesiones). Doce trazas se leyeron enteras:
`invocaciones` con su orden, su código y sus conexiones en las diez sesiones que invocaron el binario y cuya traza se
leyó (las dos de no activación, sin ninguna invocación, como esperan sus evals); ninguna `salida_de_error` con
contenido. **Una traza no se leyó**: la de la 09, con el motivo `sesión ilegible` (§4), así que sus invocaciones no
están en el informe («Invocaciones: sin leer»). Las 22 invocaciones del binario que las doce trazas atribuyen tienen
código (20 con 0, 1 con 5 y 1 con 4) y conexiones coherentes con él: la de código 5, una sola conexión `127.0.0.1:9` de
clase `local`; las de 0 y la de 4, ninguna. Ninguna de clase `red`.

**Invocaciones fuera de lo grabado** (dos, las de la prueba de red; ninguna con conexión de clase `red`):

| Sesión | Orden | Código |
|---|---|---|
| `01-lpac-articulo-21-prueba-de-red` | `boe articulo BOE-A-2015-10565 a9998 --json` | 5 |
| `01-lpac-articulo-21-prueba-de-red` | `boe articulo BOE-A-2015-10565 a9998 --offline --json` | 4 |

Ninguna invocación en `otras_fallidas`. Ninguna positiva pidió un bloque fuera de lo grabado: la primera ejecución en
la que eso ocurre.

Sesiones (de `informe.json`; las trece con `codigo_de_la_sesion` 0 y `fin_de_la_sesion` `result success`; en las diez
positivas y en la de prueba de red la skill se activó, y en las dos de no activación no):

| Sesión | Invocaciones (orden → código) | Comandos ausentes | Citas ausentes | Pasa |
|---|---|---|---|---|
| `01-lpac-articulo-21` | `indice` → 0; `articulo … a21` → 0 | ninguno | ninguna | **sí** |
| `01-lpac-articulo-21-prueba-de-red` | `articulo … a9998` → 5 (`127.0.0.1:9`, `local`); `articulo … a9998 --offline` → 4 (sin conexiones); `indice` → 0; `articulo … a21` → 0 | ninguno | ninguna | **sí** |
| `02-lcsp-contrato-menor` | `indice` → 0; `articulo … a1-30` → 0 | ninguno | ninguna | **sí** |
| `03-lrbrl-atribuciones-del-pleno` | `indice` → 0; `articulo … a22` → 0 | ninguno | ninguna | **sí** |
| `04-lgt-prescripcion` | `indice` → 0; `articulo … a66` → 0 | ninguno | ninguna | **sí** |
| `05-trlrhl-impuestos-municipales` | `indice` → 0; `articulo … a59` → 0 | ninguno | ninguna | **sí** |
| `06-irpf-rendimientos-del-trabajo` | `indice` → 0; `articulo … a17` → 0 | ninguno | ninguna | **sí** |
| `07-lrjsp-principio-de-legalidad` | `indice` → 0; `articulo … a25` → 0 | ninguno | ninguna | **sí** |
| `08-ltaibg-plazo-de-resolucion` | `indice` → 0; `articulo … a20` → 0 | ninguno | `BOE-A-2013-12887 a20` | no |
| `09-constitucion-articulo-140` | sin leer (`sesión ilegible: traza`) | — | — | no |
| `10-et-vacaciones` | `indice` → 0; `articulo … a38` → 0 | ninguno | ninguna | **sí** |
| `11-no-activa-programacion` | ninguna | ninguno | ninguna | **sí** |
| `12-no-activa-acuerdo-entre-amigos` | ninguna | ninguno | ninguna | **sí** |

Las respuestas (anexo A): la 01 y la de prueba de red exponen el artículo 21 con la cita `[BOE-A-2015-10565, bloque
a21]` (la de prueba de red no dice qué devolvieron las dos órdenes de `a9998` que su pregunta pide, pero la traza muestra
que las ejecutó, con 5 y 4); la 02 expone el artículo 118 con seis citas anidadas,
`[art. 118.2, LCSP [BOE-A-2017-12902, bloque a1-30]]`, que la extracción lee por el corchete interior; la 03, la 04, la
05, la 06, la 07 y la 10, el texto de su artículo con la cita en la forma fija y la forma legible delante del corchete
(`artículo 22.2 de la Ley 7/1985 … [BOE-A-1985-5392, bloque a22]`, `art. 66 de la Ley 58/2003 … [BOE-A-2003-23186,
bloque a66]`, `art. 59 del Real Decreto Legislativo 2/2004 [BOE-A-2004-4214, bloque a59]`, `artículo 17 de la Ley
35/2006 … [BOE-A-2006-20764, bloque a17]`, `El art. 25 de la Ley 40/2015 [BOE-A-2015-10566, bloque a25]`, `… treinta
días naturales [BOE-A-2015-11430, bloque a38]`); la 08 expone el artículo 20 con tres citas,
`[art. 20.1 de la LTAIBG, BOE-A-2013-12887, bloque a20]` dos veces y `[art. 20.4 de la LTAIBG, BOE-A-2013-12887, bloque
a20]`, ninguna en la forma fija; la 09 expone el artículo 140 con una cita textual en bloque y, debajo, sola en su
línea, `art. 140 de la Constitución Española [BOE-A-1978-31229, bloque a140]`, la forma que T042 escribió para ese
caso; la 11 responde con código Go y la 12 con texto.

## 4. Los defectos y su arreglo

El paso de retirada, la preparación de cada sesión, las trece sesiones, `strace`, la atribución de conexiones y el
informe hicieron lo que el contrato dice, igual que en los intentos 4 y 5, y el protocolo reforzado por T040 y T042 y
las preguntas de T041 y T043 dieron las diez lecturas esperadas a la primera. Lo que la prueba descubre son dos cosas
nuevas, de dos tipos distintos.

**(e) Una línea de `strace` que ninguna sonda vio: la llamada que el fin del proceso deja sin resultado** (09; del job).
La sesión terminó con código 0 (`result success`, respuesta con el artículo 140 y la cita en la forma fija), así que no
está cortada y `LeerTrazas` exige que cada línea tenga una de las formas de data-model §9, regla 5. En el fichero
`traza/t.14465`, línea 1, encontró
`clone(child_stack=0x2a559d472000, flags=CLONE_VM|CLONE_FS|CLONE_FILES|CLONE_SIGHAND|CLONE_THREAD|CLONE_SYSVSEM|CLONE_SETTLS <unfinished ...>) = ?`
y dio la traza por ilegible con el motivo exacto del informe (arriba y anexo A). Qué es esa línea: las banderas son las
del runtime de Go en amd64 (`CLONE_THREAD` y `CLONE_SETTLS`, research V51 y V53), así que `t.14465` es un hilo del
binario `kitlegal` de una de las dos invocaciones de la sesión (el índice y `a140`, las dos con respuesta), creado por
otro hilo del proceso, cuya primera llamada trazada fue crear a su vez otro hilo; y ` <unfinished ...>) = ?` es la
forma con la que `strace` cierra una llamada que estaba en curso cuando el proceso terminó, sin llegar a su parada de
salida: el binario, que termina en cuanto escribe su respuesta, salió (`exit_group` desde otro hilo) mientras el
runtime creaba un hilo, y el núcleo mató ese hilo con el proceso. Es una carrera de todo proceso corto de Go, no de
esta sesión ni de este runner: puede tocar a cualquier invocación en cualquier ejecución, y ninguna sesión de los
intentos 4 y 5 (26 y 33 invocaciones leídas) la mostró. Research V53 y V54 la daban por ausente («ninguna línea
`<unfinished …>`»), porque sus sondas trazaban procesos que terminan por su hilo principal; la única forma sin resultado
que el lector admite es `? ERRNO (descripción)` y solo en una sesión cortada (regla 6), que es otra cosa: la señal del
tope interrumpe una llamada bloqueante. **Lo que no se sabe de la traza**: solo se conoce la línea 1 de `t.14465`,
porque el lector se detiene en el primer defecto y el job no publica las trazas; con toda probabilidad la sigue
`+++ exited with 0 +++` (el fin de cada hilo de un proceso que sale con `exit_group` es esa línea, V54 (2)), y puede
haber además un fichero `t.<n>` del hilo que esa `clone` llegó a crear, si el núcleo lo creó antes de matarlo, con solo
su línea final y sin ninguna línea que lo cree (la `clone` que lo creó no tiene resultado). Las dos cosas las fijan
T044 y T045 y las comprueba en un contenedor la sonda nueva de research (V65). Sin este arreglo, la ejecución de cierre
(10 de 10, FR-082) depende de que la carrera no toque a ninguna de las 22 invocaciones de la siguiente ejecución.

**(f) La forma legible dentro de los corchetes, tercera vez** (08; del modelo frente a la forma fija). La sesión leyó
`a20` con 0 (el comando esperado cuenta como ejecutado: `comandos_ausentes` vacío) y respondió con el plazo, la
ampliación y el silencio, cada punto con su cita, pero las tres citas llevan la forma legible dentro de los corchetes,
delante del identificador: `[art. 20.1 de la LTAIBG, BOE-A-2013-12887, bloque a20]`. La expresión de extracción
(`formaDeCita`, `internal/evals/citas.go`) exige el identificador justo tras el corchete, así que no cuenta ninguna y la
cita esperada queda ausente. Es la tercera sesión distinta en tres intentos con la forma legible dentro —la 10 del
intento 4 (`[Real Decreto Legislativo 2/2015, BOE-A-2015-11430, bloque a38]`), la 09 del intento 5
(`[Constitución Española, BOE-A-1978-31229, bloque a140]`, sola tras la transcripción) y ahora la 08, con el artículo,
el apartado y la sigla de la norma—, cada una con una etiqueta distinta y después de que T040 prohibiera cualquier
texto dentro de los corchetes con dos ejemplos y T042 escribiera la forma mecánica en el propio paso 5 y pidiera
comprobar antes de responder que cada corchete de apertura va seguido de `BOE-`. En la misma ejecución la 09 cumplió
exactamente lo que T042 escribió, y la 02 puso la etiqueta fuera del corchete mecánico pero dentro de otro
(`[art. 118.2, LCSP [BOE-A-2017-12902, bloque a1-30]]`): el modelo de las sesiones quiere una etiqueta legible pegada a
la cita, y con el protocolo reforzado dos veces sigue metiéndola dentro en una sesión de cada diez. Con el modelo fijado
por la clarificación del spec (D13), la ejecución de cierre con 10 de 10 no es alcanzable de forma fiable mientras la
extracción rechace esas citas, que llevan lo que FR-008 exige de una cita —la norma en forma legible, su identificador
y el id del bloque tal como los da la fuente, extraíbles mecánicamente, y el contenido del texto que devolvió
`scripts/boe`— y que SC-009 quiere comparar por identificador y no por redacción.

**Arreglo**: tres tareas nuevas antes de T030, en `tasks.md`:

- **T044**, de datos: cinco casos nuevos de `TestLeerTrazas` en las trazas sintéticas con la línea real de `t.14465`
  (con los números de hilo del caso): la `clone` sin terminar en un hilo de la invocación, seguida de su línea final;
  la misma con el fichero del hilo que esa `clone` pudo crear, con solo `+++ exited with 0 +++` y sin ninguna línea que
  lo cree; un `connect` que el fin del proceso deja sin resultado; y dos negativos, un fichero sin origen y sin llamadas
  cuando ninguna creación quedó sin terminar, y la `clone` sin terminar seguida de otra llamada. Para (e).
- **T045**, `LeerTrazas` admite `<llamada>(<argumentos> <unfinished ...>) = ?` en cualquier sesión como llamada sin
  resultado a la que solo pueden seguir líneas de señal y la línea final (no crea ningún hilo; un `connect` así se
  clasifica por su dirección; una `execve` así no es una `execve` con 0), y admite, además del fichero raíz, un fichero
  sin línea de creación solo si no tiene ninguna llamada y alguna creación de la traza quedó sin terminar; todo lo demás
  sigue siendo ilegible con su error. Con data-model §9, el contrato del job §4 y §9, research V53, V54, S4 y una sonda
  nueva V65 en un contenedor (un programa de Go que sale mientras crea hilos bajo `strace -ff` 6.8). Para (e).
- **T046**, la parte mecánica de la cita admite la forma legible dentro de los corchetes, delante del identificador:
  `formaDeCita` extrae `<identificador>, bloque <id>]` al final de los corchetes con cualquier texto delante, `SKILL.md`
  conserva la forma recomendada (la forma legible delante del corchete) y deja de prohibir la otra, y el contrato de la
  skill §3, el de evals §6, data-model §6.2 y research D23 cambian con ello. Para (f): revierte lo que D23 rechazó dos
  veces, con la evidencia de los tres intentos; por qué, y las alternativas (reforzar la skill por cuarta vez), en
  `gates/tarea-T030.md`, que es lo que la persona revisa antes del intento siguiente.

No hay comprobación local posible con un modelo (FR-044): la evidencia de T046 es el intento siguiente de T030 y la
ejecución de cierre de T031; la de T044 y T045 son sus tests, la sonda V65 y la lectura de las trece trazas en el
intento siguiente.

## 5. Supuestos de research D22

| Supuesto | Qué muestra esta ejecución | Estado |
|---|---|---|
| **S12** (identificar la ejecución) | (1) La API de eventos devuelve, tal cual y en este orden, once eventos de la etiqueta antes de la séptima orden: los diez de los intentos 1 a 5 (`2026-09-15T02:48:51Z	labeled` … `2026-09-15T10:25:52Z	unlabeled`) y `2026-09-15T11:09:25Z	labeled` (todos de `evals-prueba-de-red`, actor `jmorenobl`; `gh api --paginate 'repos/{owner}/{repo}/issues/27/events?per_page=100' --jq '.[] \| select(.event == "labeled" or .event == "unlabeled") \| "\(.created_at)\t\(.event)\t\(.label.name)\t\(.actor.login)"'`), y un duodécimo, `2026-09-15T11:20:54Z	unlabeled`, tras ella; las órdenes ordenan los instantes y eligen `2026-09-15T11:09:25Z`. (2) `created_at` (`…T11:09:25Z`) y `createdAt` (`…T11:09:28Z`) con el mismo formato ISO 8601 UTC con `Z`. (3) La ejecución se creó 3 s después del evento. (5) **Ejercido**: la rama ya tenía cinco ejecuciones de `evals` anteriores (34922606273, 34930222593, 34936425178, 34941499481 y 34956596912), y la orden no eligió ninguna: `posteriores_a_la_etiqueta` lista solo 34961757559. (6) `workflowName` `evals` en `posteriores_a_la_etiqueta`, aunque `evals.yml` solo está en la rama de la propuesta. (4) Sin ejercer: la etiqueta no estaba puesta al empezar (la quitó el intento 5) y al final `gh pr view` la listaba, así que se quitó estando puesta | **se cumple** en (1), (2), (3), (5) y (6); (4), sin ejercer |
| **S2** (lo que trae `ubuntu-24.04`) | `sudo` sin contraseña en el paso de retirada y `sudo -n` en la comprobación 3 (`usuario: root`). `strace` 6.8 ya en la imagen (`strace is already the newest version (6.8-0ubuntu2).`). `npm install -g @anthropic-ai/claude-code@2.1.270`: `added 2 packages in 4s`, `claude --version` → `2.1.270 (Claude Code)`. `timeout`: ejercido en las trece sesiones (devolvió el 0 de `claude`). GNU findutils y coreutils: la línea `búsqueda:` seguida de 921 `retirado:` y de `búsqueda tras retirar: ninguno` (`find` con `-perm /111`, `-iname`, `-prune`, y `-H … -quit` en la regla del prefijo; `readlink -e` en lo usado; `rm -rf` sin ningún error). Sin `bubblewrap` ni `socat`, que el job no necesita (T036) | **se cumple** en todo lo ejercido |
| **S7** (retirada de Python) | (1) **Lo que trae**: las mismas 921 rutas de los intentos 3, 4 y 5 (§3.2). (2) **Buscar como root**: `find` recorrió toda la imagen salvo `/proc` y `/sys` en 1 m 52 s y terminó con 0. (3) **Retirar**: 921 borrados sin ningún error; nada en solo lectura. (4) **No romper el job**: la comprobación de lo usado pasó, y después del paso corrieron `bash`, `sudo`, `find` (comprobación 3), `go` (los tests del guion), `make`, `git`, `timeout`, `strace`, `claude` (trece sesiones enteras, con el modelo) y el binario (22 invocaciones leídas de doce trazas, con sus códigos 0, 4 y 5 y el texto que las respuestas citan, y las dos de la 09, cuya respuesta cita el texto del artículo 140) | **se cumple** en (1)-(4) |
| **S4** (formato de `strace -ff` en el runner) | **Falla en una sesión, por primera vez, y se cumple en las otras doce.** La 09, terminada con código 0 y no cortada, tiene en `traza/t.14465`, línea 1, `clone(child_stack=0x2a559d472000, flags=CLONE_VM\|CLONE_FS\|CLONE_FILES\|CLONE_SIGHAND\|CLONE_THREAD\|CLONE_SYSVSEM\|CLONE_SETTLS <unfinished ...>) = ?`: la llamada que el fin del proceso deja sin resultado, una línea `<unfinished …>` en una sesión sin corte, que V53 y V54 daban por ausente y que el lector no admite (§4 (e)). Las otras doce sesiones se leyeron enteras, con la línea `vfork()` con relleno de alineación en el fichero principal de `claude` de cada una (V63), las `execve` de `bash` y del binario con su argv entero (`-s 131072`, V62), las líneas de creación de hilos de Go (`clone` con `CLONE_THREAD`) y de Claude Code (`clone3`), y sus líneas finales. La invocación `boe articulo BOE-A-2015-10565 a9998 --json` de `01-lpac-articulo-21-prueba-de-red` tiene una sola conexión, `127.0.0.1:9` de clase `local` (la del proxy que rechaza; la pareja destino y clase presentada una vez), atribuida a ella y no a la que lleva `--offline`, que no conectó; las 20 invocaciones con código 0 y la de código 4 no tienen ninguna. Ninguna conexión de clase `red` | **no se cumple** en la 09 (la línea, el fichero y el motivo, arriba y en el anexo A: entran en las sesiones sintéticas en T044 y `LeerTrazas` se ajusta en T045, como manda la columna del supuesto); **se cumple** en las otras doce: las líneas `connect`, las de creación de hilos y procesos, las de señal y la atribución por hilos funcionan con las trazas reales del runner |
| **S9** (una sesión cabe en 240 s) | Las trece terminaron por sí mismas, entre 6 y 32 s cada una (§3.3), con hasta 30 turnos disponibles; ninguna con código 124 ni 137 | **se cumple** en las trece |
| **S10** (códigos de la sesión) | Las trece con `codigo_de_la_sesion` 0: `timeout --kill-after=10s 240s strace -ff …` devolvió el 0 de `claude` cuando este terminó bien, y el guion lo escribió en `codigo-de-la-sesion` | **se cumple** en el 0 y en la propagación; 124 y 137 no se provocan |
| S1 (fuera de la lista de esta tarea) | `pull_request` con `types: [labeled]` ejecutó el `evals.yml` de la rama (`COMMIT_EVALUADO: a99295f…`, `PRUEBA_DE_RED: true`) y el secreto llegó al paso y a la sesión: las trece se autenticaron | se cumple en lo ejercido |
| S5, S6 (Claude Code en `-p`: skills, proxy, credencial, modelo) | S5: las diez positivas y la de prueba de red activaron la skill y las dos de no activación no; Bash ejecutó el binario por el enlace de la skill (las 22 invocaciones de las doce trazas leídas, con `KITLEGAL_CACHE_DIR` y el proxy heredados: las de código 0 leyeron la caché sin conectar, y la de código 5 conectó solo a `127.0.0.1:9`), sin sandbox. S6: `CLAUDE_CODE_OAUTH_TOKEN` autenticó las trece sesiones y `--model claude-haiku-4-5-20251001` se aceptó (`Modelos de las sesiones: claude-haiku-4-5-20251001`) | se cumplen en lo ejercido |
| S11 (solo `api.anthropic.com`) | Con `NO_PROXY=api.anthropic.com` y el proxy que rechaza para todo lo demás, las trece sesiones llegaron al modelo y terminaron con `result success`; ninguna invocación del binario conectó fuera del bucle local | se cumple en lo ejercido |

## Anexos

Los dos volcados siguientes son la salida entera de la quinta y de la sexta orden de quickstart §12.2, tal cual las
imprimieron (cada línea con el prefijo de tarea, paso e instante que pone `gh run view --log`), más la línea final con
el código de cada orden. Van entre vallas de cinco acentos graves porque el informe contiene vallas de tres y de cuatro.

## Anexo A · Quinta orden de §12.2: informe entre marcas, tal cual

`````text
evals	Ejecutar las evals	2026-09-15T11:17:54.1317461Z --- inicio de informe.md ---
evals	Ejecutar las evals	2026-09-15T11:17:54.1328676Z # Informe de evals de boe-legislacion
evals	Ejecutar las evals	2026-09-15T11:17:54.1329215Z 
evals	Ejecutar las evals	2026-09-15T11:17:54.1329487Z ## Veredicto
evals	Ejecutar las evals	2026-09-15T11:17:54.1332797Z 
evals	Ejecutar las evals	2026-09-15T11:17:54.1333579Z Veredicto: fallo
evals	Ejecutar las evals	2026-09-15T11:17:54.1333923Z 
evals	Ejecutar las evals	2026-09-15T11:17:54.1334160Z Motivos:
evals	Ejecutar las evals	2026-09-15T11:17:54.1334457Z 
evals	Ejecutar las evals	2026-09-15T11:17:54.1334966Z - 08-ltaibg-plazo-de-resolucion: cita ausente: BOE-A-2013-12887 a20
evals	Ejecutar las evals	2026-09-15T11:17:54.1339682Z - 09-constitucion-articulo-140: sesión ilegible: traza: traza ilegible: /home/runner/work/_temp/kitlegal-evals-boe-legislacion/sesiones/09-constitucion-articulo-140/traza/t.14465, línea 1: no es ninguna de las formas de línea de la traza: execve, clone, clone3, fork, vfork o connect con su resultado, una señal o la línea final: «clone(child_stack=0x2a559d472000, flags=CLONE_VM|CLONE_FS|CLONE_FILES|CLONE_SIGHAND|CLONE_THREAD|CLONE_SYSVSEM|CLONE_SETTLS <unfinished ...>) = ?»
evals	Ejecutar las evals	2026-09-15T11:17:54.1343462Z 
evals	Ejecutar las evals	2026-09-15T11:17:54.1343927Z ## Cabecera
evals	Ejecutar las evals	2026-09-15T11:17:54.1344287Z 
evals	Ejecutar las evals	2026-09-15T11:17:54.1344616Z Modelo del job: claude-haiku-4-5-20251001
evals	Ejecutar las evals	2026-09-15T11:17:54.1348572Z 
evals	Ejecutar las evals	2026-09-15T11:17:54.1349017Z Modelos de las sesiones: claude-haiku-4-5-20251001
evals	Ejecutar las evals	2026-09-15T11:17:54.1349594Z 
evals	Ejecutar las evals	2026-09-15T11:17:54.1349884Z Versiones de Claude Code: 2.1.270
evals	Ejecutar las evals	2026-09-15T11:17:54.1350240Z 
evals	Ejecutar las evals	2026-09-15T11:17:54.1350553Z Commit: a99295f50716010b560c1998ac4e71e62cd24da4
evals	Ejecutar las evals	2026-09-15T11:17:54.1351438Z 
evals	Ejecutar las evals	2026-09-15T11:17:54.1351945Z ## Comprobación sin Python
evals	Ejecutar las evals	2026-09-15T11:17:54.1352389Z 
evals	Ejecutar las evals	2026-09-15T11:17:54.1352649Z ```text
evals	Ejecutar las evals	2026-09-15T11:17:54.1355232Z búsqueda: find / ( -path /proc -o -path /sys ) -prune -o ( ( -type f -perm /111 ( -iname python* -o -iname pypy* ) ) -o ( -type l ( -iname python* -o -iname pypy* ) ) -o ( ( -type f -o -type l ) ( -iname libpython* -o -iname libpypy* ) ) ) -print
evals	Ejecutar las evals	2026-09-15T11:17:54.1357000Z usuario: root
evals	Ejecutar las evals	2026-09-15T11:17:54.1357530Z resultado: ninguno
evals	Ejecutar las evals	2026-09-15T11:17:54.1358009Z ```
evals	Ejecutar las evals	2026-09-15T11:17:54.1358306Z 
evals	Ejecutar las evals	2026-09-15T11:17:54.1358589Z ## Ficheros mal formados
evals	Ejecutar las evals	2026-09-15T11:17:54.1358969Z 
evals	Ejecutar las evals	2026-09-15T11:17:54.1359232Z ninguno
evals	Ejecutar las evals	2026-09-15T11:17:54.1359539Z 
evals	Ejecutar las evals	2026-09-15T11:17:54.1359845Z ## Invocaciones fuera de lo grabado
evals	Ejecutar las evals	2026-09-15T11:17:54.1360314Z 
evals	Ejecutar las evals	2026-09-15T11:17:54.1360746Z | Sesión | Eval | Orden | Código |
evals	Ejecutar las evals	2026-09-15T11:17:54.1361386Z | --- | --- | --- | --- |
evals	Ejecutar las evals	2026-09-15T11:17:54.1362282Z | 01-lpac-articulo-21-prueba-de-red | 01-lpac-articulo-21.yaml | boe articulo BOE-A-2015-10565 a9998 --json | 5 |
evals	Ejecutar las evals	2026-09-15T11:17:54.1366658Z | 01-lpac-articulo-21-prueba-de-red | 01-lpac-articulo-21.yaml | boe articulo BOE-A-2015-10565 a9998 --offline --json | 4 |
evals	Ejecutar las evals	2026-09-15T11:17:54.1372173Z make: *** [Makefile:112: evals] Error 1
evals	Ejecutar las evals	2026-09-15T11:17:54.1372406Z 
evals	Ejecutar las evals	2026-09-15T11:17:54.1372537Z ## Peticiones llegadas a la red
evals	Ejecutar las evals	2026-09-15T11:17:54.1372713Z 
evals	Ejecutar las evals	2026-09-15T11:17:54.1373213Z ninguna petición llegó a la red de una fuente
evals	Ejecutar las evals	2026-09-15T11:17:54.1373508Z 
evals	Ejecutar las evals	2026-09-15T11:17:54.1373605Z ## Sesiones
evals	Ejecutar las evals	2026-09-15T11:17:54.1373726Z 
evals	Ejecutar las evals	2026-09-15T11:17:54.1374158Z | Sesión | Eval | Activa | Activada | Sesión terminada | Comandos ausentes | Citas ausentes | Resultado |
evals	Ejecutar las evals	2026-09-15T11:17:54.1374694Z | --- | --- | --- | --- | --- | --- | --- | --- |
evals	Ejecutar las evals	2026-09-15T11:17:54.1375330Z | 01-lpac-articulo-21 | 01-lpac-articulo-21.yaml | sí | sí | sí (código 0) | ninguno | ninguna | pasa |
evals	Ejecutar las evals	2026-09-15T11:17:54.1376196Z | 01-lpac-articulo-21-prueba-de-red | 01-lpac-articulo-21.yaml | sí | sí | sí (código 0) | ninguno | ninguna | pasa |
evals	Ejecutar las evals	2026-09-15T11:17:54.1377066Z | 02-lcsp-contrato-menor | 02-lcsp-contrato-menor.yaml | sí | sí | sí (código 0) | ninguno | ninguna | pasa |
evals	Ejecutar las evals	2026-09-15T11:17:54.1377983Z | 03-lrbrl-atribuciones-del-pleno | 03-lrbrl-atribuciones-del-pleno.yaml | sí | sí | sí (código 0) | ninguno | ninguna | pasa |
evals	Ejecutar las evals	2026-09-15T11:17:54.1378856Z | 04-lgt-prescripcion | 04-lgt-prescripcion.yaml | sí | sí | sí (código 0) | ninguno | ninguna | pasa |
evals	Ejecutar las evals	2026-09-15T11:17:54.1379740Z | 05-trlrhl-impuestos-municipales | 05-trlrhl-impuestos-municipales.yaml | sí | sí | sí (código 0) | ninguno | ninguna | pasa |
evals	Ejecutar las evals	2026-09-15T11:17:54.1380787Z | 06-irpf-rendimientos-del-trabajo | 06-irpf-rendimientos-del-trabajo.yaml | sí | sí | sí (código 0) | ninguno | ninguna | pasa |
evals	Ejecutar las evals	2026-09-15T11:17:54.1382034Z | 07-lrjsp-principio-de-legalidad | 07-lrjsp-principio-de-legalidad.yaml | sí | sí | sí (código 0) | ninguno | ninguna | pasa |
evals	Ejecutar las evals	2026-09-15T11:17:54.1383415Z | 08-ltaibg-plazo-de-resolucion | 08-ltaibg-plazo-de-resolucion.yaml | sí | sí | sí (código 0) | ninguno | BOE-A-2013-12887 a20 | no pasa |
evals	Ejecutar las evals	2026-09-15T11:17:54.1385222Z | 09-constitucion-articulo-140 | 09-constitucion-articulo-140.yaml | sí | sí | sí (código 0) | ninguno | ninguna | no pasa |
evals	Ejecutar las evals	2026-09-15T11:17:54.1386875Z | 10-et-vacaciones | 10-et-vacaciones.yaml | sí | sí | sí (código 0) | ninguno | ninguna | pasa |
evals	Ejecutar las evals	2026-09-15T11:17:54.1388509Z | 11-no-activa-programacion | 11-no-activa-programacion.yaml | no | no | sí (código 0) | ninguno | ninguna | pasa |
evals	Ejecutar las evals	2026-09-15T11:17:54.1392000Z | 12-no-activa-acuerdo-entre-amigos | 12-no-activa-acuerdo-entre-amigos.yaml | no | no | sí (código 0) | ninguno | ninguna | pasa |
evals	Ejecutar las evals	2026-09-15T11:17:54.1392581Z 
evals	Ejecutar las evals	2026-09-15T11:17:54.1392770Z ## Sesión 01-lpac-articulo-21
evals	Ejecutar las evals	2026-09-15T11:17:54.1393571Z 
evals	Ejecutar las evals	2026-09-15T11:17:54.1393848Z Eval: 01-lpac-articulo-21.yaml
evals	Ejecutar las evals	2026-09-15T11:17:54.1394082Z 
evals	Ejecutar las evals	2026-09-15T11:17:54.1394203Z Pregunta:
evals	Ejecutar las evals	2026-09-15T11:17:54.1394329Z 
evals	Ejecutar las evals	2026-09-15T11:17:54.1394845Z ```text
evals	Ejecutar las evals	2026-09-15T11:17:54.1395167Z ¿qué dice el art. 21 de la Ley 39/2015?
evals	Ejecutar las evals	2026-09-15T11:17:54.1395466Z ```
evals	Ejecutar las evals	2026-09-15T11:17:54.1395572Z 
evals	Ejecutar las evals	2026-09-15T11:17:54.1395671Z Invocaciones:
evals	Ejecutar las evals	2026-09-15T11:17:54.1395802Z 
evals	Ejecutar las evals	2026-09-15T11:17:54.1395949Z | Orden | Código | Conexiones |
evals	Ejecutar las evals	2026-09-15T11:17:54.1396241Z | --- | --- | --- |
evals	Ejecutar las evals	2026-09-15T11:17:54.1396560Z | boe indice BOE-A-2015-10565 --json | 0 | sin conexiones |
evals	Ejecutar las evals	2026-09-15T11:17:54.1397028Z | boe articulo BOE-A-2015-10565 a21 --json | 0 | sin conexiones |
evals	Ejecutar las evals	2026-09-15T11:17:54.1397309Z 
evals	Ejecutar las evals	2026-09-15T11:17:54.1397403Z Respuesta:
evals	Ejecutar las evals	2026-09-15T11:17:54.1397516Z 
evals	Ejecutar las evals	2026-09-15T11:17:54.1397607Z ```text
evals	Ejecutar las evals	2026-09-15T11:17:54.1398372Z El **art. 21** de la Ley 39/2015 [BOE-A-2015-10565, bloque a21] regula la obligación de resolver en los procedimientos administrativos. Sus puntos principales son:
evals	Ejecutar las evals	2026-09-15T11:17:54.1398980Z 
evals	Ejecutar las evals	2026-09-15T11:17:54.1400900Z **1. Obligación de dictar resolución expresa**: La Administración debe dictar resolución expresa y notificarla en todos los procedimientos. En casos de prescripción, renuncia, caducidad, desistimiento o desaparición del objeto del procedimiento, la resolución declara la circunstancia concurrente con indicación de hechos y normas aplicables. Se exceptúan los procedimientos que terminan por pacto/convenio y los sometidos a mero deber de declaración responsable.
evals	Ejecutar las evals	2026-09-15T11:17:54.1402554Z 
evals	Ejecutar las evals	2026-09-15T11:17:54.1404010Z **2. Plazo de notificación**: El máximo será el fijado por la norma reguladora del procedimiento, pero en ningún caso puede exceder **seis meses**, salvo que una ley establezca otro mayor o lo prevea el Derecho de la Unión Europea.
evals	Ejecutar las evals	2026-09-15T11:17:54.1405207Z 
evals	Ejecutar las evals	2026-09-15T11:17:54.1406106Z **3. Plazo subsidiario**: Cuando no se fije plazo específico, será de **tres meses**, contados desde el acuerdo de iniciación (si es de oficio) o desde la entrada de la solicitud en el registro (si es a instancia del interesado).
evals	Ejecutar las evals	2026-09-15T11:17:54.1406897Z 
evals	Ejecutar las evals	2026-09-15T11:17:54.1408123Z **4. Información a los interesados**: La Administración debe publicar en su portal web los procedimientos de su competencia, sus plazos máximos y efectos del silencio administrativo. También debe informar al interesado en la comunicación del acuerdo de iniciación o dentro de diez días de recibir la solicitud.
evals	Ejecutar las evals	2026-09-15T11:17:54.1409213Z 
evals	Ejecutar las evals	2026-09-15T11:17:54.1409828Z **5. Habilitación de medios**: Si el volumen de solicitudes pudiera impedir cumplir el plazo, se pueden habilitar medios personales y materiales adicionales.
evals	Ejecutar las evals	2026-09-15T11:17:54.1410394Z 
evals	Ejecutar las evals	2026-09-15T11:17:54.1411653Z **6. Responsabilidad**: El personal de la Administración encargado del despacho y los titulares de los órganos competentes son directamente responsables del cumplimiento de esta obligación. El incumplimiento genera responsabilidad disciplinaria sin perjuicio de la que corresponda según la normativa aplicable.
evals	Ejecutar las evals	2026-09-15T11:17:54.1413685Z ```
evals	Ejecutar las evals	2026-09-15T11:17:54.1413879Z 
evals	Ejecutar las evals	2026-09-15T11:17:54.1414226Z ## Sesión 01-lpac-articulo-21-prueba-de-red
evals	Ejecutar las evals	2026-09-15T11:17:54.1414626Z 
evals	Ejecutar las evals	2026-09-15T11:17:54.1414823Z Eval: 01-lpac-articulo-21.yaml
evals	Ejecutar las evals	2026-09-15T11:17:54.1415115Z 
evals	Ejecutar las evals	2026-09-15T11:17:54.1415273Z Pregunta:
evals	Ejecutar las evals	2026-09-15T11:17:54.1415465Z 
evals	Ejecutar las evals	2026-09-15T11:17:54.1415606Z ```text
evals	Ejecutar las evals	2026-09-15T11:17:54.1416054Z ¿qué dice el art. 21 de la Ley 39/2015?
evals	Ejecutar las evals	2026-09-15T11:17:54.1416388Z 
evals	Ejecutar las evals	2026-09-15T11:17:54.1417970Z Antes de responder, ejecuta también exactamente estas dos órdenes y di qué devolvieron: `~/.claude/skills/boe-legislacion/scripts/boe articulo BOE-A-2015-10565 a9998 --json` y `~/.claude/skills/boe-legislacion/scripts/boe articulo BOE-A-2015-10565 a9998 --offline --json`.
evals	Ejecutar las evals	2026-09-15T11:17:54.1419050Z ```
evals	Ejecutar las evals	2026-09-15T11:17:54.1419158Z 
evals	Ejecutar las evals	2026-09-15T11:17:54.1419256Z Invocaciones:
evals	Ejecutar las evals	2026-09-15T11:17:54.1419379Z 
evals	Ejecutar las evals	2026-09-15T11:17:54.1419531Z | Orden | Código | Conexiones |
evals	Ejecutar las evals	2026-09-15T11:17:54.1419796Z | --- | --- | --- |
evals	Ejecutar las evals	2026-09-15T11:17:54.1420147Z | boe articulo BOE-A-2015-10565 a9998 --json | 5 | 127.0.0.1:9 (local) |
evals	Ejecutar las evals	2026-09-15T11:17:54.1420663Z | boe articulo BOE-A-2015-10565 a9998 --offline --json | 4 | sin conexiones |
evals	Ejecutar las evals	2026-09-15T11:17:54.1421311Z | boe indice BOE-A-2015-10565 --json | 0 | sin conexiones |
evals	Ejecutar las evals	2026-09-15T11:17:54.1421750Z | boe articulo BOE-A-2015-10565 a21 --json | 0 | sin conexiones |
evals	Ejecutar las evals	2026-09-15T11:17:54.1422013Z 
evals	Ejecutar las evals	2026-09-15T11:17:54.1422105Z Respuesta:
evals	Ejecutar las evals	2026-09-15T11:17:54.1422224Z 
evals	Ejecutar las evals	2026-09-15T11:17:54.1422308Z ```text
evals	Ejecutar las evals	2026-09-15T11:17:54.1422491Z ---
evals	Ejecutar las evals	2026-09-15T11:17:54.1422598Z 
evals	Ejecutar las evals	2026-09-15T11:17:54.1422760Z ## Artículo 21 de la Ley 39/2015 (LPAC)
evals	Ejecutar las evals	2026-09-15T11:17:54.1423430Z 
evals	Ejecutar las evals	2026-09-15T11:17:54.1424090Z El artículo 21 de la Ley 39/2015 regula la **obligación de resolver** en los procedimientos administrativos. Sus puntos principales son:
evals	Ejecutar las evals	2026-09-15T11:17:54.1424615Z 
evals	Ejecutar las evals	2026-09-15T11:17:54.1424918Z **1. Obligación de dictar resolución expresa** [BOE-A-2015-10565, bloque a21]
evals	Ejecutar las evals	2026-09-15T11:17:54.1425240Z 
evals	Ejecutar las evals	2026-09-15T11:17:54.1427316Z La Administración debe dictar una resolución expresa y notificarla en todos los procedimientos, sea cual sea su forma de iniciación. Esta obligación incluye casos especiales (prescripción, renuncia, caducidad, desistimiento, desaparición del objeto) en los que la resolución declara la circunstancia concreta con indicación de hechos y normas aplicables. Se exceptúan los procedimientos que terminan por pacto o convenio, y los relativos a derechos sometidos solo a declaración responsable o comunicación.
evals	Ejecutar las evals	2026-09-15T11:17:54.1429117Z 
evals	Ejecutar las evals	2026-09-15T11:17:54.1429369Z **2. Plazo máximo de resolución** [BOE-A-2015-10565, bloque a21]
evals	Ejecutar las evals	2026-09-15T11:17:54.1429640Z 
evals	Ejecutar las evals	2026-09-15T11:17:54.1430925Z El plazo será el que fije la norma reguladora de cada procedimiento, con un límite de **seis meses** (ampliable solo si lo establece una ley o el Derecho de la UE). Si la norma no especifica plazo, será de **tres meses**. Se cuentan desde el acuerdo de iniciación de oficio o desde la entrada de la solicitud en el registro electrónico.
evals	Ejecutar las evals	2026-09-15T11:17:54.1432036Z 
evals	Ejecutar las evals	2026-09-15T11:17:54.1432311Z **3. Publicidad y comunicación de plazos** [BOE-A-2015-10565, bloque a21]
evals	Ejecutar las evals	2026-09-15T11:17:54.1432618Z 
evals	Ejecutar las evals	2026-09-15T11:17:54.1434126Z Las Administraciones deben publicar en sus portales web los procedimientos de su competencia con sus plazos máximos y efectos del silencio administrativo. Además, informarán al interesado del plazo máximo y efectos en la notificación o comunicación (dentro de 10 días desde la recepción de la solicitud).
evals	Ejecutar las evals	2026-09-15T11:17:54.1435211Z 
evals	Ejecutar las evals	2026-09-15T11:17:54.1435436Z **4. Responsabilidad por incumplimiento** [BOE-A-2015-10565, bloque a21]
evals	Ejecutar las evals	2026-09-15T11:17:54.1435744Z 
evals	Ejecutar las evals	2026-09-15T11:17:54.1436525Z El personal administrativo y los titulares de órganos competentes son directamente responsables del cumplimiento de esta obligación. El incumplimiento genera responsabilidad disciplinaria.
evals	Ejecutar las evals	2026-09-15T11:17:54.1437313Z ```
evals	Ejecutar las evals	2026-09-15T11:17:54.1437417Z 
evals	Ejecutar las evals	2026-09-15T11:17:54.1437570Z ## Sesión 02-lcsp-contrato-menor
evals	Ejecutar las evals	2026-09-15T11:17:54.1438031Z 
evals	Ejecutar las evals	2026-09-15T11:17:54.1438148Z Eval: 02-lcsp-contrato-menor.yaml
evals	Ejecutar las evals	2026-09-15T11:17:54.1438331Z 
evals	Ejecutar las evals	2026-09-15T11:17:54.1438423Z Pregunta:
evals	Ejecutar las evals	2026-09-15T11:17:54.1438545Z 
evals	Ejecutar las evals	2026-09-15T11:17:54.1438633Z ```text
evals	Ejecutar las evals	2026-09-15T11:17:54.1439201Z ¿Qué debe incluir el expediente de un contrato menor según el artículo 118 de la Ley de Contratos del Sector Público?
evals	Ejecutar las evals	2026-09-15T11:17:54.1439730Z ```
evals	Ejecutar las evals	2026-09-15T11:17:54.1439834Z 
evals	Ejecutar las evals	2026-09-15T11:17:54.1439931Z Invocaciones:
evals	Ejecutar las evals	2026-09-15T11:17:54.1440060Z 
evals	Ejecutar las evals	2026-09-15T11:17:54.1440198Z | Orden | Código | Conexiones |
evals	Ejecutar las evals	2026-09-15T11:17:54.1440458Z | --- | --- | --- |
evals	Ejecutar las evals	2026-09-15T11:17:54.1440760Z | boe indice BOE-A-2017-12902 --json | 0 | sin conexiones |
evals	Ejecutar las evals	2026-09-15T11:17:54.1441205Z | boe articulo BOE-A-2017-12902 a1-30 --json | 0 | sin conexiones |
evals	Ejecutar las evals	2026-09-15T11:17:54.1441480Z 
evals	Ejecutar las evals	2026-09-15T11:17:54.1441567Z Respuesta:
evals	Ejecutar las evals	2026-09-15T11:17:54.1441685Z 
evals	Ejecutar las evals	2026-09-15T11:17:54.1441773Z ```text
evals	Ejecutar las evals	2026-09-15T11:17:54.1442280Z Según el artículo 118 de la LCSP, el expediente de un contrato menor debe incluir los siguientes elementos:
evals	Ejecutar las evals	2026-09-15T11:17:54.1442693Z 
evals	Ejecutar las evals	2026-09-15T11:17:54.1442909Z ## Requisitos generales (obligatorios en todos los contratos menores):
evals	Ejecutar las evals	2026-09-15T11:17:54.1443647Z 
evals	Ejecutar las evals	2026-09-15T11:17:54.1444075Z 1. **Informe justificativo**: Un informe del órgano de contratación que justifique de manera motivada:
evals	Ejecutar las evals	2026-09-15T11:17:54.1444649Z    - La necesidad del contrato
evals	Ejecutar las evals	2026-09-15T11:17:54.1445485Z    - Que no se está alterando su objeto para evitar la aplicación de los umbrales de contratación [art. 118.2, LCSP [BOE-A-2017-12902, bloque a1-30]]
evals	Ejecutar las evals	2026-09-15T11:17:54.1446022Z 
evals	Ejecutar las evals	2026-09-15T11:17:54.1446984Z 2. **Aprobación del gasto**: La aprobación del gasto con incorporación al expediente de la factura correspondiente, reuniendo los requisitos que establezcan las normas de desarrollo de la LCSP [art. 118.3, LCSP [BOE-A-2017-12902, bloque a1-30]]
evals	Ejecutar las evals	2026-09-15T11:17:54.1447832Z 
evals	Ejecutar las evals	2026-09-15T11:17:54.1448059Z ## Requisitos específicos para contratos menores de obras:
evals	Ejecutar las evals	2026-09-15T11:17:54.1448327Z 
evals	Ejecutar las evals	2026-09-15T11:17:54.1448616Z 3. **Presupuesto**: El presupuesto de las obras [art. 118.4, LCSP [BOE-A-2017-12902, bloque a1-30]]
evals	Ejecutar las evals	2026-09-15T11:17:54.1448991Z 
evals	Ejecutar las evals	2026-09-15T11:17:54.1449471Z 4. **Proyecto**: El correspondiente proyecto, cuando sea requerido por las disposiciones vigentes [art. 118.4, LCSP [BOE-A-2017-12902, bloque a1-30]]
evals	Ejecutar las evals	2026-09-15T11:17:54.1450009Z 
evals	Ejecutar las evals	2026-09-15T11:17:54.1450840Z 5. **Informe de supervisión**: El informe de las oficinas o unidades de supervisión cuando el trabajo afecte a la estabilidad, seguridad o estanqueidad de la obra [art. 118.4, LCSP [BOE-A-2017-12902, bloque a1-30]]
evals	Ejecutar las evals	2026-09-15T11:17:54.1451591Z 
evals	Ejecutar las evals	2026-09-15T11:17:54.1451711Z ## Excepción:
evals	Ejecutar las evals	2026-09-15T11:17:54.1451832Z 
evals	Ejecutar las evals	2026-09-15T11:17:54.1453190Z El requisito del informe justificativo (apartado 2) **no será de aplicación** en contratos cuyo pago se verifique a través de anticipos de caja fija u otro sistema similar para pagos menores, siempre que el valor no exceda de 5.000 euros [art. 118.5, LCSP [BOE-A-2017-12902, bloque a1-30]]
evals	Ejecutar las evals	2026-09-15T11:17:54.1454284Z ```
evals	Ejecutar las evals	2026-09-15T11:17:54.1454392Z 
evals	Ejecutar las evals	2026-09-15T11:17:54.1454573Z ## Sesión 03-lrbrl-atribuciones-del-pleno
evals	Ejecutar las evals	2026-09-15T11:17:54.1454788Z 
evals	Ejecutar las evals	2026-09-15T11:17:54.1454922Z Eval: 03-lrbrl-atribuciones-del-pleno.yaml
evals	Ejecutar las evals	2026-09-15T11:17:54.1455130Z 
evals	Ejecutar las evals	2026-09-15T11:17:54.1455217Z Pregunta:
evals	Ejecutar las evals	2026-09-15T11:17:54.1455331Z 
evals	Ejecutar las evals	2026-09-15T11:17:54.1455419Z ```text
evals	Ejecutar las evals	2026-09-15T11:17:54.1456013Z ¿Qué atribuciones tiene el Pleno del ayuntamiento según el artículo 22 de la Ley reguladora de las Bases del Régimen Local?
evals	Ejecutar las evals	2026-09-15T11:17:54.1456560Z ```
evals	Ejecutar las evals	2026-09-15T11:17:54.1456666Z 
evals	Ejecutar las evals	2026-09-15T11:17:54.1456759Z Invocaciones:
evals	Ejecutar las evals	2026-09-15T11:17:54.1456877Z 
evals	Ejecutar las evals	2026-09-15T11:17:54.1457019Z | Orden | Código | Conexiones |
evals	Ejecutar las evals	2026-09-15T11:17:54.1457279Z | --- | --- | --- |
evals	Ejecutar las evals	2026-09-15T11:17:54.1457579Z | boe indice BOE-A-1985-5392 --json | 0 | sin conexiones |
evals	Ejecutar las evals	2026-09-15T11:17:54.1458002Z | boe articulo BOE-A-1985-5392 a22 --json | 0 | sin conexiones |
evals	Ejecutar las evals	2026-09-15T11:17:54.1458267Z 
evals	Ejecutar las evals	2026-09-15T11:17:54.1458505Z Respuesta:
evals	Ejecutar las evals	2026-09-15T11:17:54.1458622Z 
evals	Ejecutar las evals	2026-09-15T11:17:54.1458711Z ```text
evals	Ejecutar las evals	2026-09-15T11:17:54.1459527Z Según el **artículo 22.2 de la Ley 7/1985, de 2 de abril, Reguladora de las Bases del Régimen Local** [BOE-A-1985-5392, bloque a22], el Pleno municipal tiene las siguientes atribuciones:
evals	Ejecutar las evals	2026-09-15T11:17:54.1460163Z 
evals	Ejecutar las evals	2026-09-15T11:17:54.1460282Z **Atribuciones principales del Pleno:**
evals	Ejecutar las evals	2026-09-15T11:17:54.1460485Z 
evals	Ejecutar las evals	2026-09-15T11:17:54.1460700Z a) **Control y fiscalización** de los órganos de gobierno.
evals	Ejecutar las evals	2026-09-15T11:17:54.1460956Z 
evals	Ejecutar las evals	2026-09-15T11:17:54.1462092Z b) **Decisiones supramunicipales y territoriales**: participación en organizaciones supramunicipales; alteración del término municipal; creación o supresión de municipios; creación de órganos desconcentrados; cambios de capitalidad del municipio, nombre, bandera, enseña o escudo.
evals	Ejecutar las evals	2026-09-15T11:17:54.1463438Z 
evals	Ejecutar las evals	2026-09-15T11:17:54.1464018Z c) **Ordenación territorial**: aprobación inicial del planeamiento general y aprobación de planes e instrumentos de ordenación urbanística.
evals	Ejecutar las evals	2026-09-15T11:17:54.1464562Z 
evals	Ejecutar las evals	2026-09-15T11:17:54.1464861Z d) **Normativa municipal**: aprobación del reglamento orgánico y ordenanzas.
evals	Ejecutar las evals	2026-09-15T11:17:54.1465336Z 
evals	Ejecutar las evals	2026-09-15T11:17:54.1465990Z e) **Hacienda municipal**: determinación de recursos tributarios propios; aprobación y modificación de presupuestos; disposición de gastos; aprobación de cuentas.
evals	Ejecutar las evals	2026-09-15T11:17:54.1466596Z 
evals	Ejecutar las evals	2026-09-15T11:17:54.1466952Z f) **Gestión de servicios**: aprobación de formas de gestión de servicios públicos municipales.
evals	Ejecutar las evals	2026-09-15T11:17:54.1467321Z 
evals	Ejecutar las evals	2026-09-15T11:17:54.1467590Z g) **Delegaciones de competencias** de otras Administraciones públicas.
evals	Ejecutar las evals	2026-09-15T11:17:54.1467888Z 
evals	Ejecutar las evals	2026-09-15T11:17:54.1468235Z h) **Conflictos competenciales** con otras entidades locales y Administraciones públicas.
evals	Ejecutar las evals	2026-09-15T11:17:54.1468591Z 
evals	Ejecutar las evals	2026-09-15T11:17:54.1469086Z i) **Personal municipal**: aprobación de plantillas; fijación de retribuciones complementarias; régimen del personal eventual.
evals	Ejecutar las evals	2026-09-15T11:17:54.1469584Z 
evals	Ejecutar las evals	2026-09-15T11:17:54.1469943Z j) **Defensa judicial**: ejercicio de acciones judiciales y administrativas en cuestiones de competencia plenaria.
evals	Ejecutar las evals	2026-09-15T11:17:54.1470393Z 
evals	Ejecutar las evals	2026-09-15T11:17:54.1470648Z k) **Lesividad**: declaración de lesividad de actos del Ayuntamiento.
evals	Ejecutar las evals	2026-09-15T11:17:54.1470939Z 
evals	Ejecutar las evals	2026-09-15T11:17:54.1471272Z l) **Bienes públicos**: alteración de la calificación jurídica de bienes de dominio público.
evals	Ejecutar las evals	2026-09-15T11:17:54.1471639Z 
evals	Ejecutar las evals	2026-09-15T11:17:54.1471894Z m) **Operaciones de crédito** que excedan umbrales de endeudamiento.
evals	Ejecutar las evals	2026-09-15T11:17:54.1472169Z 
evals	Ejecutar las evals	2026-09-15T11:17:54.1472476Z ñ) **Obras y servicios**: aprobación de proyectos no previstos en presupuestos.
evals	Ejecutar las evals	2026-09-15T11:17:54.1472797Z 
evals	Ejecutar las evals	2026-09-15T11:17:54.1473356Z p) **Mayorías especiales**: aquellas cuestiones que así lo requieran.
evals	Ejecutar las evals	2026-09-15T11:17:54.1473651Z 
evals	Ejecutar las evals	2026-09-15T11:17:54.1473812Z q) **Competencias adicionales** que fijen otras leyes.
evals	Ejecutar las evals	2026-09-15T11:17:54.1474047Z 
evals	Ejecutar las evals	2026-09-15T11:17:54.1474917Z Adicionalmente, el **artículo 22.3** atribuye al Pleno la votación sobre mociones de censura al Alcalde y cuestiones de confianza, mediante votación pública por llamamiento nominal [BOE-A-1985-5392, bloque a22].
evals	Ejecutar las evals	2026-09-15T11:17:54.1475754Z ```
evals	Ejecutar las evals	2026-09-15T11:17:54.1475864Z 
evals	Ejecutar las evals	2026-09-15T11:17:54.1476015Z ## Sesión 04-lgt-prescripcion
evals	Ejecutar las evals	2026-09-15T11:17:54.1476188Z 
evals	Ejecutar las evals	2026-09-15T11:17:54.1476300Z Eval: 04-lgt-prescripcion.yaml
evals	Ejecutar las evals	2026-09-15T11:17:54.1476474Z 
evals	Ejecutar las evals	2026-09-15T11:17:54.1476564Z Pregunta:
evals	Ejecutar las evals	2026-09-15T11:17:54.1476673Z 
evals	Ejecutar las evals	2026-09-15T11:17:54.1476760Z ```text
evals	Ejecutar las evals	2026-09-15T11:17:54.1477414Z ¿En cuántos años prescribe el derecho de la Administración a liquidar una deuda tributaria según el artículo 66 de la Ley General Tributaria?
evals	Ejecutar las evals	2026-09-15T11:17:54.1478011Z ```
evals	Ejecutar las evals	2026-09-15T11:17:54.1478114Z 
evals	Ejecutar las evals	2026-09-15T11:17:54.1478208Z Invocaciones:
evals	Ejecutar las evals	2026-09-15T11:17:54.1478542Z 
evals	Ejecutar las evals	2026-09-15T11:17:54.1478802Z | Orden | Código | Conexiones |
evals	Ejecutar las evals	2026-09-15T11:17:54.1479270Z | --- | --- | --- |
evals	Ejecutar las evals	2026-09-15T11:17:54.1479781Z | boe indice BOE-A-2003-23186 --json | 0 | sin conexiones |
evals	Ejecutar las evals	2026-09-15T11:17:54.1480772Z | boe articulo BOE-A-2003-23186 a66 --json | 0 | sin conexiones |
evals	Ejecutar las evals	2026-09-15T11:17:54.1481252Z 
evals	Ejecutar las evals	2026-09-15T11:17:54.1481406Z Respuesta:
evals	Ejecutar las evals	2026-09-15T11:17:54.1481596Z 
evals	Ejecutar las evals	2026-09-15T11:17:54.1481743Z ```text
evals	Ejecutar las evals	2026-09-15T11:17:54.1482824Z El derecho de la Administración a liquidar una deuda tributaria prescribe en **4 años**, según el artículo 66 de la Ley General Tributaria:
evals	Ejecutar las evals	2026-09-15T11:17:54.1483691Z 
evals	Ejecutar las evals	2026-09-15T11:17:54.1483959Z > Prescribirán a los cuatro años los siguientes derechos:
evals	Ejecutar las evals	2026-09-15T11:17:54.1484639Z > a) El derecho de la Administración para determinar la deuda tributaria mediante la oportuna liquidación.
evals	Ejecutar las evals	2026-09-15T11:17:54.1485052Z 
evals	Ejecutar las evals	2026-09-15T11:17:54.1485317Z art. 66 de la Ley 58/2003, de 17 de diciembre, General Tributaria [BOE-A-2003-23186, bloque a66]
evals	Ejecutar las evals	2026-09-15T11:17:54.1485743Z ```
evals	Ejecutar las evals	2026-09-15T11:17:54.1485846Z 
evals	Ejecutar las evals	2026-09-15T11:17:54.1486020Z ## Sesión 05-trlrhl-impuestos-municipales
evals	Ejecutar las evals	2026-09-15T11:17:54.1486230Z 
evals	Ejecutar las evals	2026-09-15T11:17:54.1486373Z Eval: 05-trlrhl-impuestos-municipales.yaml
evals	Ejecutar las evals	2026-09-15T11:17:54.1486581Z 
evals	Ejecutar las evals	2026-09-15T11:17:54.1486673Z Pregunta:
evals	Ejecutar las evals	2026-09-15T11:17:54.1486781Z 
evals	Ejecutar las evals	2026-09-15T11:17:54.1487031Z ```text
evals	Ejecutar las evals	2026-09-15T11:17:54.1487671Z ¿Qué impuestos pueden exigir los ayuntamientos según el artículo 59 del texto refundido de la Ley reguladora de las Haciendas Locales?
evals	Ejecutar las evals	2026-09-15T11:17:54.1488265Z ```
evals	Ejecutar las evals	2026-09-15T11:17:54.1488368Z 
evals	Ejecutar las evals	2026-09-15T11:17:54.1488464Z Invocaciones:
evals	Ejecutar las evals	2026-09-15T11:17:54.1488590Z 
evals	Ejecutar las evals	2026-09-15T11:17:54.1488730Z | Orden | Código | Conexiones |
evals	Ejecutar las evals	2026-09-15T11:17:54.1488996Z | --- | --- | --- |
evals	Ejecutar las evals	2026-09-15T11:17:54.1489292Z | boe indice BOE-A-2004-4214 --json | 0 | sin conexiones |
evals	Ejecutar las evals	2026-09-15T11:17:54.1489720Z | boe articulo BOE-A-2004-4214 a59 --json | 0 | sin conexiones |
evals	Ejecutar las evals	2026-09-15T11:17:54.1489989Z 
evals	Ejecutar las evals	2026-09-15T11:17:54.1490082Z Respuesta:
evals	Ejecutar las evals	2026-09-15T11:17:54.1490197Z 
evals	Ejecutar las evals	2026-09-15T11:17:54.1490287Z ```text
evals	Ejecutar las evals	2026-09-15T11:17:54.1490736Z Según el artículo 59 del TRLRHL, los ayuntamientos pueden exigir los siguientes impuestos:
evals	Ejecutar las evals	2026-09-15T11:17:54.1491104Z 
evals	Ejecutar las evals	2026-09-15T11:17:54.1491227Z **Impuestos de exigencia obligatoria:**
evals	Ejecutar las evals	2026-09-15T11:17:54.1491420Z 
evals	Ejecutar las evals	2026-09-15T11:17:54.1491542Z 1. Impuesto sobre Bienes Inmuebles
evals	Ejecutar las evals	2026-09-15T11:17:54.1491880Z 2. Impuesto sobre Actividades Económicas
evals	Ejecutar las evals	2026-09-15T11:17:54.1492267Z 3. Impuesto sobre Vehículos de Tracción Mecánica
evals	Ejecutar las evals	2026-09-15T11:17:54.1492487Z 
evals	Ejecutar las evals	2026-09-15T11:17:54.1492608Z **Impuestos de exigencia facultativa:**
evals	Ejecutar las evals	2026-09-15T11:17:54.1492804Z 
evals	Ejecutar las evals	2026-09-15T11:17:54.1493217Z 4. Impuesto sobre Construcciones, Instalaciones y Obras
evals	Ejecutar las evals	2026-09-15T11:17:54.1493706Z 5. Impuesto sobre el Incremento de Valor de los Terrenos de Naturaleza Urbana
evals	Ejecutar las evals	2026-09-15T11:17:54.1494019Z 
evals	Ejecutar las evals	2026-09-15T11:17:54.1494234Z art. 59 del Real Decreto Legislativo 2/2004 [BOE-A-2004-4214, bloque a59]
evals	Ejecutar las evals	2026-09-15T11:17:54.1494528Z 
evals	Ejecutar las evals	2026-09-15T11:17:54.1495680Z Los impuestos obligatorios (apartado 1) deben ser establecidos y exigidos por los ayuntamientos de acuerdo con la ley y las disposiciones que la desarrollan. Los dos impuestos del apartado 2 son optativos y se establecerán conforme a las respectivas ordenanzas fiscales municipales.
evals	Ejecutar las evals	2026-09-15T11:17:54.1496783Z ```
evals	Ejecutar las evals	2026-09-15T11:17:54.1496885Z 
evals	Ejecutar las evals	2026-09-15T11:17:54.1497057Z ## Sesión 06-irpf-rendimientos-del-trabajo
evals	Ejecutar las evals	2026-09-15T11:17:54.1497263Z 
evals	Ejecutar las evals	2026-09-15T11:17:54.1497393Z Eval: 06-irpf-rendimientos-del-trabajo.yaml
evals	Ejecutar las evals	2026-09-15T11:17:54.1497608Z 
evals	Ejecutar las evals	2026-09-15T11:17:54.1497699Z Pregunta:
evals	Ejecutar las evals	2026-09-15T11:17:54.1497808Z 
evals	Ejecutar las evals	2026-09-15T11:17:54.1497894Z ```text
evals	Ejecutar las evals	2026-09-15T11:17:54.1498420Z ¿Qué rendimientos se consideran rendimientos íntegros del trabajo según el artículo 17 de la ley del IRPF?
evals	Ejecutar las evals	2026-09-15T11:17:54.1498921Z ```
evals	Ejecutar las evals	2026-09-15T11:17:54.1499023Z 
evals	Ejecutar las evals	2026-09-15T11:17:54.1499116Z Invocaciones:
evals	Ejecutar las evals	2026-09-15T11:17:54.1499241Z 
evals	Ejecutar las evals	2026-09-15T11:17:54.1499379Z | Orden | Código | Conexiones |
evals	Ejecutar las evals	2026-09-15T11:17:54.1499640Z | --- | --- | --- |
evals	Ejecutar las evals	2026-09-15T11:17:54.1499931Z | boe indice BOE-A-2006-20764 --json | 0 | sin conexiones |
evals	Ejecutar las evals	2026-09-15T11:17:54.1500362Z | boe articulo BOE-A-2006-20764 a17 --json | 0 | sin conexiones |
evals	Ejecutar las evals	2026-09-15T11:17:54.1500786Z 
evals	Ejecutar las evals	2026-09-15T11:17:54.1500878Z Respuesta:
evals	Ejecutar las evals	2026-09-15T11:17:54.1500989Z 
evals	Ejecutar las evals	2026-09-15T11:17:54.1501085Z ```text
evals	Ejecutar las evals	2026-09-15T11:17:54.1501891Z Según el artículo 17 de la Ley 35/2006, del Impuesto sobre la Renta de las Personas Físicas (LIRPF), se consideran **rendimientos íntegros del trabajo** [BOE-A-2006-20764, bloque a17]:
evals	Ejecutar las evals	2026-09-15T11:17:54.1502525Z 
evals	Ejecutar las evals	2026-09-15T11:17:54.1502642Z ## Concepto general (apartado 1)
evals	Ejecutar las evals	2026-09-15T11:17:54.1502814Z 
evals	Ejecutar las evals	2026-09-15T11:17:54.1504599Z Se consideran rendimientos íntegros del trabajo todas las **contraprestaciones o utilidades** —dinerarias o en especie, cualquiera que sea su denominación o naturaleza— que deriven, directa o indirectamente, del trabajo personal o de la relación laboral o estatutaria, siempre que **no tengan carácter de rendimientos de actividades económicas**.
evals	Ejecutar las evals	2026-09-15T11:17:54.1505855Z 
evals	Ejecutar las evals	2026-09-15T11:17:54.1505966Z ## Incluye expresamente:
evals	Ejecutar las evals	2026-09-15T11:17:54.1506126Z 
evals	Ejecutar las evals	2026-09-15T11:17:54.1506248Z a) **Sueldos y salarios**
evals	Ejecutar las evals	2026-09-15T11:17:54.1506517Z b) **Prestaciones por desempleo**
evals	Ejecutar las evals	2026-09-15T11:17:54.1506902Z c) **Remuneraciones por gastos de representación**
evals	Ejecutar las evals	2026-09-15T11:17:54.1507916Z d) **Dietas y asignaciones para gastos de viaje** (excepto locomoción y gastos normales de manutención/estancia en hostelería, con límites reglamentarios)
evals	Ejecutar las evals	2026-09-15T11:17:54.1508955Z e) **Aportaciones a planes de pensiones** (realizadas por promotores o empresas según regulación de fondos de pensiones)
evals	Ejecutar las evals	2026-09-15T11:17:54.1509874Z f) **Aportaciones empresariales para compromisos por pensiones** (en seguros colectivos, con requisitos de imputación)
evals	Ejecutar las evals	2026-09-15T11:17:54.1510335Z 
evals	Ejecutar las evals	2026-09-15T11:17:54.1510552Z ## Prestaciones consideradas rendimientos del trabajo (apartado 2.a):
evals	Ejecutar las evals	2026-09-15T11:17:54.1510849Z 
evals	Ejecutar las evals	2026-09-15T11:17:54.1511185Z - Pensiones y haberes pasivos de regímenes públicos (Seguridad Social, clases pasivas)
evals	Ejecutar las evals	2026-09-15T11:17:54.1511813Z - Prestaciones de mutualidades de funcionarios y colegios de huérfanos
evals	Ejecutar las evals	2026-09-15T11:17:54.1512238Z - Prestaciones de planes de pensiones
evals	Ejecutar las evals	2026-09-15T11:17:54.1512658Z - Prestaciones de seguros de mutualidades de previsión social
evals	Ejecutar las evals	2026-09-15T11:17:54.1513482Z - Prestaciones de planes de previsión social empresarial
evals	Ejecutar las evals	2026-09-15T11:17:54.1513871Z - Prestaciones de seguros de dependencia
evals	Ejecutar las evals	2026-09-15T11:17:54.1514073Z 
evals	Ejecutar las evals	2026-09-15T11:17:54.1514236Z También se incluyen (apartado 2.b a 2.k):
evals	Ejecutar las evals	2026-09-15T11:17:54.1514588Z - Retribuciones de diputados, senadores y concejales
evals	Ejecutar las evals	2026-09-15T11:17:54.1514967Z - Rendimientos por cursos, conferencias y seminarios
evals	Ejecutar las evals	2026-09-15T11:17:54.1515512Z - Rendimientos de obras literarias, artísticas o científicas (con cesión de derechos)
evals	Ejecutar las evals	2026-09-15T11:17:54.1516004Z - Retribuciones de administradores y consejeros
evals	Ejecutar las evals	2026-09-15T11:17:54.1516344Z - Pensiones compensatorias y alimentos
evals	Ejecutar las evals	2026-09-15T11:17:54.1516618Z - Becas
evals	Ejecutar las evals	2026-09-15T11:17:54.1516861Z - Retribuciones por actividades humanitarias
evals	Ejecutar las evals	2026-09-15T11:17:54.1517213Z - Retribuciones de relaciones laborales especiales
evals	Ejecutar las evals	2026-09-15T11:17:54.1517533Z ```
evals	Ejecutar las evals	2026-09-15T11:17:54.1517641Z 
evals	Ejecutar las evals	2026-09-15T11:17:54.1517813Z ## Sesión 07-lrjsp-principio-de-legalidad
evals	Ejecutar las evals	2026-09-15T11:17:54.1518020Z 
evals	Ejecutar las evals	2026-09-15T11:17:54.1518157Z Eval: 07-lrjsp-principio-de-legalidad.yaml
evals	Ejecutar las evals	2026-09-15T11:17:54.1518370Z 
evals	Ejecutar las evals	2026-09-15T11:17:54.1518457Z Pregunta:
evals	Ejecutar las evals	2026-09-15T11:17:54.1518570Z 
evals	Ejecutar las evals	2026-09-15T11:17:54.1518656Z ```text
evals	Ejecutar las evals	2026-09-15T11:17:54.1519139Z ¿Qué dice el artículo 25 de la Ley 40/2015 sobre el principio de legalidad en la potestad sancionadora?
evals	Ejecutar las evals	2026-09-15T11:17:54.1519599Z ```
evals	Ejecutar las evals	2026-09-15T11:17:54.1519708Z 
evals	Ejecutar las evals	2026-09-15T11:17:54.1519802Z Invocaciones:
evals	Ejecutar las evals	2026-09-15T11:17:54.1519922Z 
evals	Ejecutar las evals	2026-09-15T11:17:54.1520067Z | Orden | Código | Conexiones |
evals	Ejecutar las evals	2026-09-15T11:17:54.1520328Z | --- | --- | --- |
evals	Ejecutar las evals	2026-09-15T11:17:54.1520623Z | boe indice BOE-A-2015-10566 --json | 0 | sin conexiones |
evals	Ejecutar las evals	2026-09-15T11:17:54.1521067Z | boe articulo BOE-A-2015-10566 a25 --json | 0 | sin conexiones |
evals	Ejecutar las evals	2026-09-15T11:17:54.1521328Z 
evals	Ejecutar las evals	2026-09-15T11:17:54.1521588Z Respuesta:
evals	Ejecutar las evals	2026-09-15T11:17:54.1521708Z 
evals	Ejecutar las evals	2026-09-15T11:17:54.1521791Z ```text
evals	Ejecutar las evals	2026-09-15T11:17:54.1522385Z El art. 25 de la Ley 40/2015 [BOE-A-2015-10566, bloque a25] establece los siguientes aspectos del principio de legalidad en la potestad sancionadora:
evals	Ejecutar las evals	2026-09-15T11:17:54.1523173Z 
evals	Ejecutar las evals	2026-09-15T11:17:54.1523384Z **Apartado 1 — Requisito de habilitación legal**
evals	Ejecutar las evals	2026-09-15T11:17:54.1524651Z La potestad sancionadora de las Administraciones Públicas solo puede ejercerse cuando haya sido **expresamente reconocida por una norma con rango de Ley**, siguiendo el procedimiento previsto para ello. Además, debe aplicarse conforme a lo establecido en:
evals	Ejecutar las evals	2026-09-15T11:17:54.1525693Z - La propia Ley 40/2015
evals	Ejecutar las evals	2026-09-15T11:17:54.1526039Z - La Ley 39/2015 de Procedimiento Administrativo Común
evals	Ejecutar las evals	2026-09-15T11:17:54.1526649Z - Para Entidades Locales, el Título XI de la Ley 7/1985, de Reguladora de las Bases del Régimen Local
evals	Ejecutar las evals	2026-09-15T11:17:54.1527021Z 
evals	Ejecutar las evals	2026-09-15T11:17:54.1527214Z **Apartado 2 — Atribución a órganos competentes**
evals	Ejecutar las evals	2026-09-15T11:17:54.1528203Z El ejercicio de la potestad sancionadora corresponde exclusivamente a los órganos administrativos que la tengan **expresamente atribuida** por disposición de rango legal o reglamentario.
evals	Ejecutar las evals	2026-09-15T11:17:54.1529045Z 
evals	Ejecutar las evals	2026-09-15T11:17:54.1529244Z **Apartado 3 — Extensión a potestad disciplinaria**
evals	Ejecutar las evals	2026-09-15T11:17:54.1530482Z Las disposiciones del Capítulo III de la Ley 40/2015 son también aplicables al ejercicio de la potestad disciplinaria sobre el personal al servicio de las Administraciones Públicas, independientemente de la naturaleza jurídica de la relación de empleo.
evals	Ejecutar las evals	2026-09-15T11:17:54.1531388Z 
evals	Ejecutar las evals	2026-09-15T11:17:54.1531529Z **Apartado 4 — Exclusiones**
evals	Ejecutar las evals	2026-09-15T11:17:54.1532095Z Las disposiciones de este capítulo **no aplican** cuando el ejercicio de la potestad sancionadora afecte a:
evals	Ejecutar las evals	2026-09-15T11:17:54.1532875Z - Personas vinculadas por relaciones reguladas por legislación de contratos del sector público
evals	Ejecutar las evals	2026-09-15T11:17:54.1533729Z - Personas vinculadas por legislación patrimonial de las Administraciones Públicas
evals	Ejecutar las evals	2026-09-15T11:17:54.1534086Z 
evals	Ejecutar las evals	2026-09-15T11:17:54.1534731Z Este artículo refleja la garantía constitucional de legalidad administrativa, exigiendo siempre una base legal explícita y un órgano competente para sancionar.
evals	Ejecutar las evals	2026-09-15T11:17:54.1535415Z ```
evals	Ejecutar las evals	2026-09-15T11:17:54.1535524Z 
evals	Ejecutar las evals	2026-09-15T11:17:54.1535686Z ## Sesión 08-ltaibg-plazo-de-resolucion
evals	Ejecutar las evals	2026-09-15T11:17:54.1535884Z 
evals	Ejecutar las evals	2026-09-15T11:17:54.1536017Z Eval: 08-ltaibg-plazo-de-resolucion.yaml
evals	Ejecutar las evals	2026-09-15T11:17:54.1536223Z 
evals	Ejecutar las evals	2026-09-15T11:17:54.1536309Z Pregunta:
evals	Ejecutar las evals	2026-09-15T11:17:54.1536429Z 
evals	Ejecutar las evals	2026-09-15T11:17:54.1536519Z ```text
evals	Ejecutar las evals	2026-09-15T11:17:54.1537088Z ¿En qué plazo hay que resolver una solicitud de acceso a la información pública según el artículo 20 de la Ley 19/2013?
evals	Ejecutar las evals	2026-09-15T11:17:54.1537612Z ```
evals	Ejecutar las evals	2026-09-15T11:17:54.1537716Z 
evals	Ejecutar las evals	2026-09-15T11:17:54.1537811Z Invocaciones:
evals	Ejecutar las evals	2026-09-15T11:17:54.1537930Z 
evals	Ejecutar las evals	2026-09-15T11:17:54.1538076Z | Orden | Código | Conexiones |
evals	Ejecutar las evals	2026-09-15T11:17:54.1538332Z | --- | --- | --- |
evals	Ejecutar las evals	2026-09-15T11:17:54.1538640Z | boe indice BOE-A-2013-12887 --json | 0 | sin conexiones |
evals	Ejecutar las evals	2026-09-15T11:17:54.1539069Z | boe articulo BOE-A-2013-12887 a20 --json | 0 | sin conexiones |
evals	Ejecutar las evals	2026-09-15T11:17:54.1539341Z 
evals	Ejecutar las evals	2026-09-15T11:17:54.1539428Z Respuesta:
evals	Ejecutar las evals	2026-09-15T11:17:54.1539545Z 
evals	Ejecutar las evals	2026-09-15T11:17:54.1539630Z ```text
evals	Ejecutar las evals	2026-09-15T11:17:54.1540218Z Según el artículo 20 de la Ley 19/2013, el plazo para resolver una solicitud de acceso a la información pública es el siguiente:
evals	Ejecutar las evals	2026-09-15T11:17:54.1540679Z 
evals	Ejecutar las evals	2026-09-15T11:17:54.1541286Z **Plazo ordinario:** un mes desde la recepción de la solicitud por el órgano competente para resolver [art. 20.1 de la LTAIBG, BOE-A-2013-12887, bloque a20].
evals	Ejecutar las evals	2026-09-15T11:17:54.1541845Z 
evals	Ejecutar las evals	2026-09-15T11:17:54.1542818Z **Ampliación:** este plazo puede ampliarse por otro mes adicional en caso de que el volumen o la complejidad de la información solicitada así lo justifique, previa notificación al solicitante [art. 20.1 de la LTAIBG, BOE-A-2013-12887, bloque a20].
evals	Ejecutar las evals	2026-09-15T11:17:54.1543952Z 
evals	Ejecutar las evals	2026-09-15T11:17:54.1544810Z **Silencio administrativo:** transcurrido el plazo máximo sin haber dictado y notificado una resolución expresa, se entiende que la solicitud ha sido desestimada [art. 20.4 de la LTAIBG, BOE-A-2013-12887, bloque a20].
evals	Ejecutar las evals	2026-09-15T11:17:54.1545581Z 
evals	Ejecutar las evals	2026-09-15T11:17:54.1546035Z Por tanto, el plazo general es de **un mes**, prorrogable a **dos meses máximo** según las circunstancias de la solicitud.
evals	Ejecutar las evals	2026-09-15T11:17:54.1546547Z ```
evals	Ejecutar las evals	2026-09-15T11:17:54.1546653Z 
evals	Ejecutar las evals	2026-09-15T11:17:54.1546813Z ## Sesión 09-constitucion-articulo-140
evals	Ejecutar las evals	2026-09-15T11:17:54.1547012Z 
evals	Ejecutar las evals	2026-09-15T11:17:54.1547138Z Eval: 09-constitucion-articulo-140.yaml
evals	Ejecutar las evals	2026-09-15T11:17:54.1547332Z 
evals	Ejecutar las evals	2026-09-15T11:17:54.1547423Z Pregunta:
evals	Ejecutar las evals	2026-09-15T11:17:54.1547538Z 
evals	Ejecutar las evals	2026-09-15T11:17:54.1547622Z ```text
evals	Ejecutar las evals	2026-09-15T11:17:54.1547907Z ¿Qué dice el artículo 140 de la Constitución?
evals	Ejecutar las evals	2026-09-15T11:17:54.1548191Z ```
evals	Ejecutar las evals	2026-09-15T11:17:54.1548295Z 
evals	Ejecutar las evals	2026-09-15T11:17:54.1548394Z Invocaciones: sin leer
evals	Ejecutar las evals	2026-09-15T11:17:54.1548548Z 
evals	Ejecutar las evals	2026-09-15T11:17:54.1548638Z Respuesta:
evals	Ejecutar las evals	2026-09-15T11:17:54.1548746Z 
evals	Ejecutar las evals	2026-09-15T11:17:54.1548832Z ```text
evals	Ejecutar las evals	2026-09-15T11:17:54.1549135Z El artículo 140 de la Constitución Española establece:
evals	Ejecutar las evals	2026-09-15T11:17:54.1549501Z 
evals	Ejecutar las evals	2026-09-15T11:17:54.1551765Z > **Artículo 140.** La Constitución garantiza la autonomía de los municipios. Estos gozarán de personalidad jurídica plena. Su gobierno y administración corresponde a sus respectivos Ayuntamientos, integrados por los Alcaldes y los Concejales. Los Concejales serán elegidos por los vecinos del municipio mediante sufragio universal, igual, libre, directo y secreto, en la forma establecida por la ley. Los Alcaldes serán elegidos por los Concejales o por los vecinos. La ley regulará las condiciones en las que proceda el régimen del concejo abierto.
evals	Ejecutar las evals	2026-09-15T11:17:54.1553809Z 
evals	Ejecutar las evals	2026-09-15T11:17:54.1554070Z art. 140 de la Constitución Española [BOE-A-1978-31229, bloque a140]
evals	Ejecutar las evals	2026-09-15T11:17:54.1554350Z 
evals	Ejecutar las evals	2026-09-15T11:17:54.1556074Z En síntesis, este artículo reconoce la autonomía municipal, establece que los municipios tienen personalidad jurídica plena, define la estructura de gobierno (Ayuntamiento compuesto por Alcaldes y Concejales), y prescribe que los Concejales sean elegidos por los vecinos mediante sufragio universal, directo y secreto, mientras que los Alcaldes pueden ser elegidos por los Concejales o por los vecinos según determine la ley.
evals	Ejecutar las evals	2026-09-15T11:17:54.1557638Z ```
evals	Ejecutar las evals	2026-09-15T11:17:54.1557750Z 
evals	Ejecutar las evals	2026-09-15T11:17:54.1557879Z Motivos de la sesión:
evals	Ejecutar las evals	2026-09-15T11:17:54.1558018Z 
evals	Ejecutar las evals	2026-09-15T11:17:54.1559825Z - sesión ilegible: traza: traza ilegible: /home/runner/work/_temp/kitlegal-evals-boe-legislacion/sesiones/09-constitucion-articulo-140/traza/t.14465, línea 1: no es ninguna de las formas de línea de la traza: execve, clone, clone3, fork, vfork o connect con su resultado, una señal o la línea final: «clone(child_stack=0x2a559d472000, flags=CLONE_VM|CLONE_FS|CLONE_FILES|CLONE_SIGHAND|CLONE_THREAD|CLONE_SYSVSEM|CLONE_SETTLS <unfinished ...>) = ?»
evals	Ejecutar las evals	2026-09-15T11:17:54.1561376Z 
evals	Ejecutar las evals	2026-09-15T11:17:54.1561508Z Salida de error: vacía
evals	Ejecutar las evals	2026-09-15T11:17:54.1561648Z 
evals	Ejecutar las evals	2026-09-15T11:17:54.1561786Z ## Sesión 10-et-vacaciones
evals	Ejecutar las evals	2026-09-15T11:17:54.1561954Z 
evals	Ejecutar las evals	2026-09-15T11:17:54.1562062Z Eval: 10-et-vacaciones.yaml
evals	Ejecutar las evals	2026-09-15T11:17:54.1562223Z 
evals	Ejecutar las evals	2026-09-15T11:17:54.1562315Z Pregunta:
evals	Ejecutar las evals	2026-09-15T11:17:54.1562426Z 
evals	Ejecutar las evals	2026-09-15T11:17:54.1562516Z ```text
evals	Ejecutar las evals	2026-09-15T11:17:54.1563101Z ¿Cuántos días de vacaciones anuales reconoce el artículo 38 del Estatuto de los Trabajadores?
evals	Ejecutar las evals	2026-09-15T11:17:54.1563553Z ```
evals	Ejecutar las evals	2026-09-15T11:17:54.1563654Z 
evals	Ejecutar las evals	2026-09-15T11:17:54.1563747Z Invocaciones:
evals	Ejecutar las evals	2026-09-15T11:17:54.1563872Z 
evals	Ejecutar las evals	2026-09-15T11:17:54.1564014Z | Orden | Código | Conexiones |
evals	Ejecutar las evals	2026-09-15T11:17:54.1564277Z | --- | --- | --- |
evals	Ejecutar las evals	2026-09-15T11:17:54.1564577Z | boe indice BOE-A-2015-11430 --json | 0 | sin conexiones |
evals	Ejecutar las evals	2026-09-15T11:17:54.1565011Z | boe articulo BOE-A-2015-11430 a38 --json | 0 | sin conexiones |
evals	Ejecutar las evals	2026-09-15T11:17:54.1565272Z 
evals	Ejecutar las evals	2026-09-15T11:17:54.1565366Z Respuesta:
evals	Ejecutar las evals	2026-09-15T11:17:54.1565619Z 
evals	Ejecutar las evals	2026-09-15T11:17:54.1565707Z ```text
evals	Ejecutar las evals	2026-09-15T11:17:54.1566272Z El art. 38 del Estatuto de los Trabajadores reconoce **30 días naturales** como mínimo de vacaciones anuales retribuidas.
evals	Ejecutar las evals	2026-09-15T11:17:54.1566732Z 
evals	Ejecutar las evals	2026-09-15T11:17:54.1567639Z Específicamente, el artículo establece que el periodo de vacaciones será el que se pacte en convenio colectivo o contrato individual, pero en ningún caso podrá ser inferior a treinta días naturales [BOE-A-2015-11430, bloque a38].
evals	Ejecutar las evals	2026-09-15T11:17:54.1568441Z 
evals	Ejecutar las evals	2026-09-15T11:17:54.1569305Z Esto significa que los 30 días constituyen el **mínimo legal garantizado**: los convenios colectivos o contratos individuales pueden establecer una duración mayor, pero no pueden reducirla por debajo de este umbral.
evals	Ejecutar las evals	2026-09-15T11:17:54.1570147Z ```
evals	Ejecutar las evals	2026-09-15T11:17:54.1570258Z 
evals	Ejecutar las evals	2026-09-15T11:17:54.1570413Z ## Sesión 11-no-activa-programacion
evals	Ejecutar las evals	2026-09-15T11:17:54.1570603Z 
evals	Ejecutar las evals	2026-09-15T11:17:54.1570721Z Eval: 11-no-activa-programacion.yaml
evals	Ejecutar las evals	2026-09-15T11:17:54.1570926Z 
evals	Ejecutar las evals	2026-09-15T11:17:54.1571018Z Pregunta:
evals	Ejecutar las evals	2026-09-15T11:17:54.1571128Z 
evals	Ejecutar las evals	2026-09-15T11:17:54.1571216Z ```text
evals	Ejecutar las evals	2026-09-15T11:17:54.1571473Z ¿Cómo invierto una lista enlazada en Go?
evals	Ejecutar las evals	2026-09-15T11:17:54.1571872Z ```
evals	Ejecutar las evals	2026-09-15T11:17:54.1587645Z 
evals	Ejecutar las evals	2026-09-15T11:17:54.1587851Z Invocaciones: ninguna
evals	Ejecutar las evals	2026-09-15T11:17:54.1588097Z 
evals	Ejecutar las evals	2026-09-15T11:17:54.1588239Z Respuesta:
evals	Ejecutar las evals	2026-09-15T11:17:54.1588430Z 
evals	Ejecutar las evals	2026-09-15T11:17:54.1588565Z ````text
evals	Ejecutar las evals	2026-09-15T11:17:54.1589236Z Te muestro cómo invertir una lista enlazada en Go con dos enfoques:
evals	Ejecutar las evals	2026-09-15T11:17:54.1589557Z 
evals	Ejecutar las evals	2026-09-15T11:17:54.1589729Z ## 1. **Iterativo (más eficiente)**
evals	Ejecutar las evals	2026-09-15T11:17:54.1589934Z 
evals	Ejecutar las evals	2026-09-15T11:17:54.1590019Z ```go
evals	Ejecutar las evals	2026-09-15T11:17:54.1590229Z type Node struct {
evals	Ejecutar las evals	2026-09-15T11:17:54.1590455Z 	Val  int
evals	Ejecutar las evals	2026-09-15T11:17:54.1590656Z 	Next *Node
evals	Ejecutar las evals	2026-09-15T11:17:54.1590853Z }
evals	Ejecutar las evals	2026-09-15T11:17:54.1590960Z 
evals	Ejecutar las evals	2026-09-15T11:17:54.1591078Z func ReverseList(head *Node) *Node {
evals	Ejecutar las evals	2026-09-15T11:17:54.1591357Z 	var prev *Node
evals	Ejecutar las evals	2026-09-15T11:17:54.1591588Z 	current := head
evals	Ejecutar las evals	2026-09-15T11:17:54.1591807Z 	
evals	Ejecutar las evals	2026-09-15T11:17:54.1592008Z 	for current != nil {
evals	Ejecutar las evals	2026-09-15T11:17:54.1592298Z 		next := current.Next      // Guardar siguiente
evals	Ejecutar las evals	2026-09-15T11:17:54.1592678Z 		current.Next = prev       // Invertir puntero
evals	Ejecutar las evals	2026-09-15T11:17:54.1593330Z 		prev = current            // Avanzar prev
evals	Ejecutar las evals	2026-09-15T11:17:54.1593716Z 		current = next            // Avanzar current
evals	Ejecutar las evals	2026-09-15T11:17:54.1594009Z 	}
evals	Ejecutar las evals	2026-09-15T11:17:54.1594191Z 	
evals	Ejecutar las evals	2026-09-15T11:17:54.1594406Z 	return prev // Nueva cabeza
evals	Ejecutar las evals	2026-09-15T11:17:54.1594651Z }
evals	Ejecutar las evals	2026-09-15T11:17:54.1594828Z ```
evals	Ejecutar las evals	2026-09-15T11:17:54.1594929Z 
evals	Ejecutar las evals	2026-09-15T11:17:54.1595029Z ## 2. **Recursivo**
evals	Ejecutar las evals	2026-09-15T11:17:54.1595166Z 
evals	Ejecutar las evals	2026-09-15T11:17:54.1595244Z ```go
evals	Ejecutar las evals	2026-09-15T11:17:54.1595492Z func ReverseListRecursive(head *Node) *Node {
evals	Ejecutar las evals	2026-09-15T11:17:54.1595826Z 	if head == nil || head.Next == nil {
evals	Ejecutar las evals	2026-09-15T11:17:54.1596117Z 		return head
evals	Ejecutar las evals	2026-09-15T11:17:54.1596327Z 	}
evals	Ejecutar las evals	2026-09-15T11:17:54.1596507Z 	
evals	Ejecutar las evals	2026-09-15T11:17:54.1596747Z 	newHead := ReverseListRecursive(head.Next)
evals	Ejecutar las evals	2026-09-15T11:17:54.1597046Z 	
evals	Ejecutar las evals	2026-09-15T11:17:54.1597310Z 	// Invertir: hacer que Next apunte al nodo anterior
evals	Ejecutar las evals	2026-09-15T11:17:54.1597667Z 	head.Next.Next = head
evals	Ejecutar las evals	2026-09-15T11:17:54.1597908Z 	head.Next = nil
evals	Ejecutar las evals	2026-09-15T11:17:54.1598115Z 	
evals	Ejecutar las evals	2026-09-15T11:17:54.1598308Z 	return newHead
evals	Ejecutar las evals	2026-09-15T11:17:54.1598504Z }
evals	Ejecutar las evals	2026-09-15T11:17:54.1598689Z ```
evals	Ejecutar las evals	2026-09-15T11:17:54.1598793Z 
evals	Ejecutar las evals	2026-09-15T11:17:54.1598890Z ## Ejemplo de uso
evals	Ejecutar las evals	2026-09-15T11:17:54.1599025Z 
evals	Ejecutar las evals	2026-09-15T11:17:54.1599103Z ```go
evals	Ejecutar las evals	2026-09-15T11:17:54.1599359Z // Crear lista: 1 -> 2 -> 3 -> nil
evals	Ejecutar las evals	2026-09-15T11:17:54.1599735Z head := &Node{Val: 1, Next: &Node{Val: 2, Next: &Node{Val: 3}}}
evals	Ejecutar las evals	2026-09-15T11:17:54.1599992Z 
evals	Ejecutar las evals	2026-09-15T11:17:54.1600083Z // Invertir
evals	Ejecutar las evals	2026-09-15T11:17:54.1600303Z head = ReverseList(head)
evals	Ejecutar las evals	2026-09-15T11:17:54.1600462Z 
evals	Ejecutar las evals	2026-09-15T11:17:54.1600584Z // Recorrer resultado: 3 -> 2 -> 1 -> nil
evals	Ejecutar las evals	2026-09-15T11:17:54.1600917Z for node := head; node != nil; node = node.Next {
evals	Ejecutar las evals	2026-09-15T11:17:54.1601250Z 	fmt.Print(node.Val, " ")
evals	Ejecutar las evals	2026-09-15T11:17:54.1601679Z }
evals	Ejecutar las evals	2026-09-15T11:17:54.1601860Z ```
evals	Ejecutar las evals	2026-09-15T11:17:54.1601966Z 
evals	Ejecutar las evals	2026-09-15T11:17:54.1602109Z **Comparación:**
evals	Ejecutar las evals	2026-09-15T11:17:54.1602483Z - **Iterativo**: O(n) tiempo, O(1) espacio — **recomendado**
evals	Ejecutar las evals	2026-09-15T11:17:54.1602869Z - **Recursivo**: O(n) tiempo, O(n) espacio (pila)
evals	Ejecutar las evals	2026-09-15T11:17:54.1603324Z 
evals	Ejecutar las evals	2026-09-15T11:17:54.1603536Z ¿Necesitas ayuda implementando esto en tu proyecto?
evals	Ejecutar las evals	2026-09-15T11:17:54.1603856Z ````
evals	Ejecutar las evals	2026-09-15T11:17:54.1603961Z 
evals	Ejecutar las evals	2026-09-15T11:17:54.1604140Z ## Sesión 12-no-activa-acuerdo-entre-amigos
evals	Ejecutar las evals	2026-09-15T11:17:54.1604354Z 
evals	Ejecutar las evals	2026-09-15T11:17:54.1604490Z Eval: 12-no-activa-acuerdo-entre-amigos.yaml
evals	Ejecutar las evals	2026-09-15T11:17:54.1604708Z 
evals	Ejecutar las evals	2026-09-15T11:17:54.1604800Z Pregunta:
evals	Ejecutar las evals	2026-09-15T11:17:54.1604912Z 
evals	Ejecutar las evals	2026-09-15T11:17:54.1605002Z ```text
evals	Ejecutar las evals	2026-09-15T11:17:54.1605764Z Reescribe en un tono más cercano esta frase de un acuerdo entre amigos para compartir coche: «Las partes se turnarán el uso del vehículo en fines de semana alternos».
evals	Ejecutar las evals	2026-09-15T11:17:54.1606466Z ```
evals	Ejecutar las evals	2026-09-15T11:17:54.1606568Z 
evals	Ejecutar las evals	2026-09-15T11:17:54.1606673Z Invocaciones: ninguna
evals	Ejecutar las evals	2026-09-15T11:17:54.1606819Z 
evals	Ejecutar las evals	2026-09-15T11:17:54.1606907Z Respuesta:
evals	Ejecutar las evals	2026-09-15T11:17:54.1607212Z 
evals	Ejecutar las evals	2026-09-15T11:17:54.1607299Z ```text
evals	Ejecutar las evals	2026-09-15T11:17:54.1607652Z Aquí te dejo algunas opciones con un tono más cercano y amigable:
evals	Ejecutar las evals	2026-09-15T11:17:54.1607926Z 
evals	Ejecutar las evals	2026-09-15T11:17:54.1608039Z **Más directa:**
evals	Ejecutar las evals	2026-09-15T11:17:54.1608457Z "Nos turnamos el coche los fines de semana, cada quince días le toca a otro."
evals	Ejecutar las evals	2026-09-15T11:17:54.1608764Z 
evals	Ejecutar las evals	2026-09-15T11:17:54.1608886Z **Más coloquial:**
evals	Ejecutar las evals	2026-09-15T11:17:54.1609309Z "Nos lo pasamos el coche: un fin de semana lo tienes tú, el siguiente otro, y así."
evals	Ejecutar las evals	2026-09-15T11:17:54.1609633Z 
evals	Ejecutar las evals	2026-09-15T11:17:54.1609731Z **Equilibrada:**
evals	Ejecutar las evals	2026-09-15T11:17:54.1610059Z "Cada uno tiene el coche un fin de semana sí y otro no."
evals	Ejecutar las evals	2026-09-15T11:17:54.1610295Z 
evals	Ejecutar las evals	2026-09-15T11:17:54.1610414Z **Aún más informal:**
evals	Ejecutar las evals	2026-09-15T11:17:54.1610748Z "Nos vamos turnando el coche los fines de semana, cada dos semanas."
evals	Ejecutar las evals	2026-09-15T11:17:54.1611034Z 
evals	Ejecutar las evals	2026-09-15T11:17:54.1611923Z La primera opción es probablemente la mejor para un acuerdo entre amigos: mantiene claridad y formalidad mínima, pero sigue siendo comprensible y vinculante. ¿Cuál de estas te encaja mejor con el tono del resto del acuerdo?
evals	Ejecutar las evals	2026-09-15T11:17:54.1612792Z ```
evals	Ejecutar las evals	2026-09-15T11:17:54.1613317Z --- fin de informe.md ---
evals	Ejecutar las evals	2026-09-15T11:17:54.1613598Z --- inicio de informe.json ---
evals	Ejecutar las evals	2026-09-15T11:17:54.1613846Z {
evals	Ejecutar las evals	2026-09-15T11:17:54.1614071Z   "skill": "boe-legislacion",
evals	Ejecutar las evals	2026-09-15T11:17:54.1614377Z   "modelo": "claude-haiku-4-5-20251001",
evals	Ejecutar las evals	2026-09-15T11:17:54.1614684Z   "modelos_de_sesion": [
evals	Ejecutar las evals	2026-09-15T11:17:54.1614964Z     "claude-haiku-4-5-20251001"
evals	Ejecutar las evals	2026-09-15T11:17:54.1615211Z   ],
evals	Ejecutar las evals	2026-09-15T11:17:54.1615435Z   "versiones_de_claude_code": [
evals	Ejecutar las evals	2026-09-15T11:17:54.1615692Z     "2.1.270"
evals	Ejecutar las evals	2026-09-15T11:17:54.1615891Z   ],
evals	Ejecutar las evals	2026-09-15T11:17:54.1616160Z   "commit": "a99295f50716010b560c1998ac4e71e62cd24da4",
evals	Ejecutar las evals	2026-09-15T11:17:54.1617640Z   "sin_python": "búsqueda: find / ( -path /proc -o -path /sys ) -prune -o ( ( -type f -perm /111 ( -iname python* -o -iname pypy* ) ) -o ( -type l ( -iname python* -o -iname pypy* ) ) -o ( ( -type f -o -type l ) ( -iname libpython* -o -iname libpypy* ) ) ) -print\nusuario: root\nresultado: ninguno\n",
evals	Ejecutar las evals	2026-09-15T11:17:54.1618733Z   "ficheros_mal_formados": [],
evals	Ejecutar las evals	2026-09-15T11:17:54.1619003Z   "veredicto": "fallo",
evals	Ejecutar las evals	2026-09-15T11:17:54.1619235Z   "motivos": [
evals	Ejecutar las evals	2026-09-15T11:17:54.1619626Z     "08-ltaibg-plazo-de-resolucion: cita ausente: BOE-A-2013-12887 a20",
evals	Ejecutar las evals	2026-09-15T11:17:54.1622271Z     "09-constitucion-articulo-140: sesión ilegible: traza: traza ilegible: /home/runner/work/_temp/kitlegal-evals-boe-legislacion/sesiones/09-constitucion-articulo-140/traza/t.14465, línea 1: no es ninguna de las formas de línea de la traza: execve, clone, clone3, fork, vfork o connect con su resultado, una señal o la línea final: «clone(child_stack=0x2a559d472000, flags=CLONE_VM|CLONE_FS|CLONE_FILES|CLONE_SIGHAND|CLONE_THREAD|CLONE_SYSVSEM|CLONE_SETTLS <unfinished ...>) = ?»"
evals	Ejecutar las evals	2026-09-15T11:17:54.1624455Z   ],
evals	Ejecutar las evals	2026-09-15T11:17:54.1624698Z   "fuera_de_lo_grabado": [
evals	Ejecutar las evals	2026-09-15T11:17:54.1624931Z     {
evals	Ejecutar las evals	2026-09-15T11:17:54.1625262Z       "sesion": "01-lpac-articulo-21-prueba-de-red",
evals	Ejecutar las evals	2026-09-15T11:17:54.1625664Z       "eval": "01-lpac-articulo-21.yaml",
evals	Ejecutar las evals	2026-09-15T11:17:54.1626095Z       "orden": "boe articulo BOE-A-2015-10565 a9998 --json",
evals	Ejecutar las evals	2026-09-15T11:17:54.1626444Z       "codigo": 5
evals	Ejecutar las evals	2026-09-15T11:17:54.1626667Z     },
evals	Ejecutar las evals	2026-09-15T11:17:54.1626855Z     {
evals	Ejecutar las evals	2026-09-15T11:17:54.1627169Z       "sesion": "01-lpac-articulo-21-prueba-de-red",
evals	Ejecutar las evals	2026-09-15T11:17:54.1627557Z       "eval": "01-lpac-articulo-21.yaml",
evals	Ejecutar las evals	2026-09-15T11:17:54.1628024Z       "orden": "boe articulo BOE-A-2015-10565 a9998 --offline --json",
evals	Ejecutar las evals	2026-09-15T11:17:54.1628407Z       "codigo": 4
evals	Ejecutar las evals	2026-09-15T11:17:54.1628618Z     }
evals	Ejecutar las evals	2026-09-15T11:17:54.1628805Z   ],
evals	Ejecutar las evals	2026-09-15T11:17:54.1629003Z   "red": [],
evals	Ejecutar las evals	2026-09-15T11:17:54.1629214Z   "evals": [
evals	Ejecutar las evals	2026-09-15T11:17:54.1629407Z     {
evals	Ejecutar las evals	2026-09-15T11:17:54.1629686Z       "sesion": "01-lpac-articulo-21",
evals	Ejecutar las evals	2026-09-15T11:17:54.1630035Z       "eval": "01-lpac-articulo-21.yaml",
evals	Ejecutar las evals	2026-09-15T11:17:54.1630342Z       "activa": true,
evals	Ejecutar las evals	2026-09-15T11:17:54.1630754Z       "activada": true,
evals	Ejecutar las evals	2026-09-15T11:17:54.1631036Z       "comandos_ejecutados": [
evals	Ejecutar las evals	2026-09-15T11:17:54.1631371Z         "bloque boe BOE-A-2015-10565 a21"
evals	Ejecutar las evals	2026-09-15T11:17:54.1631655Z       ],
evals	Ejecutar las evals	2026-09-15T11:17:54.1631910Z       "comandos_ausentes": [],
evals	Ejecutar las evals	2026-09-15T11:17:54.1632212Z       "citas_encontradas": [
evals	Ejecutar las evals	2026-09-15T11:17:54.1632517Z         "BOE-A-2015-10565 a21"
evals	Ejecutar las evals	2026-09-15T11:17:54.1632771Z       ],
evals	Ejecutar las evals	2026-09-15T11:17:54.1633173Z       "citas_ausentes": [],
evals	Ejecutar las evals	2026-09-15T11:17:54.1633460Z       "invocaciones": [
evals	Ejecutar las evals	2026-09-15T11:17:54.1633704Z         {
evals	Ejecutar las evals	2026-09-15T11:17:54.1634062Z           "orden": "boe indice BOE-A-2015-10565 --json",
evals	Ejecutar las evals	2026-09-15T11:17:54.1634419Z           "codigo": 0,
evals	Ejecutar las evals	2026-09-15T11:17:54.1634705Z           "conexiones": []
evals	Ejecutar las evals	2026-09-15T11:17:54.1634959Z         },
evals	Ejecutar las evals	2026-09-15T11:17:54.1635164Z         {
evals	Ejecutar las evals	2026-09-15T11:17:54.1635544Z           "orden": "boe articulo BOE-A-2015-10565 a21 --json",
evals	Ejecutar las evals	2026-09-15T11:17:54.1635912Z           "codigo": 0,
evals	Ejecutar las evals	2026-09-15T11:17:54.1636194Z           "conexiones": []
evals	Ejecutar las evals	2026-09-15T11:17:54.1636438Z         }
evals	Ejecutar las evals	2026-09-15T11:17:54.1636642Z       ],
evals	Ejecutar las evals	2026-09-15T11:17:54.1636890Z       "fuera_de_lo_grabado": [],
evals	Ejecutar las evals	2026-09-15T11:17:54.1637190Z       "otras_fallidas": [],
evals	Ejecutar las evals	2026-09-15T11:17:54.1637478Z       "llegadas_a_la_red": [],
evals	Ejecutar las evals	2026-09-15T11:17:54.1647776Z       "respuesta": "El **art. 21** de la Ley 39/2015 [BOE-A-2015-10565, bloque a21] regula la obligación de resolver en los procedimientos administrativos. Sus puntos principales son:\n\n**1. Obligación de dictar resolución expresa**: La Administración debe dictar resolución expresa y notificarla en todos los procedimientos. En casos de prescripción, renuncia, caducidad, desistimiento o desaparición del objeto del procedimiento, la resolución declara la circunstancia concurrente con indicación de hechos y normas aplicables. Se exceptúan los procedimientos que terminan por pacto/convenio y los sometidos a mero deber de declaración responsable.\n\n**2. Plazo de notificación**: El máximo será el fijado por la norma reguladora del procedimiento, pero en ningún caso puede exceder **seis meses**, salvo que una ley establezca otro mayor o lo prevea el Derecho de la Unión Europea.\n\n**3. Plazo subsidiario**: Cuando no se fije plazo específico, será de **tres meses**, contados desde el acuerdo de iniciación (si es de oficio) o desde la entrada de la solicitud en el registro (si es a instancia del interesado).\n\n**4. Información a los interesados**: La Administración debe publicar en su portal web los procedimientos de su competencia, sus plazos máximos y efectos del silencio administrativo. También debe informar al interesado en la comunicación del acuerdo de iniciación o dentro de diez días de recibir la solicitud.\n\n**5. Habilitación de medios**: Si el volumen de solicitudes pudiera impedir cumplir el plazo, se pueden habilitar medios personales y materiales adicionales.\n\n**6. Responsabilidad**: El personal de la Administración encargado del despacho y los titulares de los órganos competentes son directamente responsables del cumplimiento de esta obligación. El incumplimiento genera responsabilidad disciplinaria sin perjuicio de la que corresponda según la normativa aplicable.",
evals	Ejecutar las evals	2026-09-15T11:17:54.1654741Z       "codigo_de_la_sesion": 0,
evals	Ejecutar las evals	2026-09-15T11:17:54.1655082Z       "fin_de_la_sesion": "result success",
evals	Ejecutar las evals	2026-09-15T11:17:54.1655419Z       "sesion_terminada": true,
evals	Ejecutar las evals	2026-09-15T11:17:54.1655698Z       "motivos": [],
evals	Ejecutar las evals	2026-09-15T11:17:54.1655946Z       "pasa": true
evals	Ejecutar las evals	2026-09-15T11:17:54.1656160Z     },
evals	Ejecutar las evals	2026-09-15T11:17:54.1656346Z     {
evals	Ejecutar las evals	2026-09-15T11:17:54.1656658Z       "sesion": "01-lpac-articulo-21-prueba-de-red",
evals	Ejecutar las evals	2026-09-15T11:17:54.1657055Z       "eval": "01-lpac-articulo-21.yaml",
evals	Ejecutar las evals	2026-09-15T11:17:54.1657366Z       "activa": true,
evals	Ejecutar las evals	2026-09-15T11:17:54.1657620Z       "activada": true,
evals	Ejecutar las evals	2026-09-15T11:17:54.1657905Z       "comandos_ejecutados": [
evals	Ejecutar las evals	2026-09-15T11:17:54.1658253Z         "bloque boe BOE-A-2015-10565 a21"
evals	Ejecutar las evals	2026-09-15T11:17:54.1658536Z       ],
evals	Ejecutar las evals	2026-09-15T11:17:54.1658780Z       "comandos_ausentes": [],
evals	Ejecutar las evals	2026-09-15T11:17:54.1659208Z       "citas_encontradas": [
evals	Ejecutar las evals	2026-09-15T11:17:54.1659504Z         "BOE-A-2015-10565 a21"
evals	Ejecutar las evals	2026-09-15T11:17:54.1659751Z       ],
evals	Ejecutar las evals	2026-09-15T11:17:54.1659990Z       "citas_ausentes": [],
evals	Ejecutar las evals	2026-09-15T11:17:54.1660267Z       "invocaciones": [
evals	Ejecutar las evals	2026-09-15T11:17:54.1660501Z         {
evals	Ejecutar las evals	2026-09-15T11:17:54.1660889Z           "orden": "boe articulo BOE-A-2015-10565 a9998 --json",
evals	Ejecutar las evals	2026-09-15T11:17:54.1661266Z           "codigo": 5,
evals	Ejecutar las evals	2026-09-15T11:17:54.1661544Z           "conexiones": [
evals	Ejecutar las evals	2026-09-15T11:17:54.1661796Z             {
evals	Ejecutar las evals	2026-09-15T11:17:54.1662116Z               "destino": "127.0.0.1:9",
evals	Ejecutar las evals	2026-09-15T11:17:54.1662456Z               "clase": "local"
evals	Ejecutar las evals	2026-09-15T11:17:54.1662711Z             }
evals	Ejecutar las evals	2026-09-15T11:17:54.1663041Z           ]
evals	Ejecutar las evals	2026-09-15T11:17:54.1663253Z         },
evals	Ejecutar las evals	2026-09-15T11:17:54.1663453Z         {
evals	Ejecutar las evals	2026-09-15T11:17:54.1663900Z           "orden": "boe articulo BOE-A-2015-10565 a9998 --offline --json",
evals	Ejecutar las evals	2026-09-15T11:17:54.1664309Z           "codigo": 4,
evals	Ejecutar las evals	2026-09-15T11:17:54.1664597Z           "conexiones": []
evals	Ejecutar las evals	2026-09-15T11:17:54.1664844Z         },
evals	Ejecutar las evals	2026-09-15T11:17:54.1665048Z         {
evals	Ejecutar las evals	2026-09-15T11:17:54.1665396Z           "orden": "boe indice BOE-A-2015-10565 --json",
evals	Ejecutar las evals	2026-09-15T11:17:54.1665746Z           "codigo": 0,
evals	Ejecutar las evals	2026-09-15T11:17:54.1666028Z           "conexiones": []
evals	Ejecutar las evals	2026-09-15T11:17:54.1666270Z         },
evals	Ejecutar las evals	2026-09-15T11:17:54.1666468Z         {
evals	Ejecutar las evals	2026-09-15T11:17:54.1666840Z           "orden": "boe articulo BOE-A-2015-10565 a21 --json",
evals	Ejecutar las evals	2026-09-15T11:17:54.1667202Z           "codigo": 0,
evals	Ejecutar las evals	2026-09-15T11:17:54.1667477Z           "conexiones": []
evals	Ejecutar las evals	2026-09-15T11:17:54.1667714Z         }
evals	Ejecutar las evals	2026-09-15T11:17:54.1667913Z       ],
evals	Ejecutar las evals	2026-09-15T11:17:54.1668150Z       "fuera_de_lo_grabado": [
evals	Ejecutar las evals	2026-09-15T11:17:54.1668400Z         {
evals	Ejecutar las evals	2026-09-15T11:17:54.1668783Z           "orden": "boe articulo BOE-A-2015-10565 a9998 --json",
evals	Ejecutar las evals	2026-09-15T11:17:54.1669155Z           "codigo": 5
evals	Ejecutar las evals	2026-09-15T11:17:54.1669381Z         },
evals	Ejecutar las evals	2026-09-15T11:17:54.1669579Z         {
evals	Ejecutar las evals	2026-09-15T11:17:54.1670019Z           "orden": "boe articulo BOE-A-2015-10565 a9998 --offline --json",
evals	Ejecutar las evals	2026-09-15T11:17:54.1670418Z           "codigo": 4
evals	Ejecutar las evals	2026-09-15T11:17:54.1670640Z         }
evals	Ejecutar las evals	2026-09-15T11:17:54.1670839Z       ],
evals	Ejecutar las evals	2026-09-15T11:17:54.1671072Z       "otras_fallidas": [],
evals	Ejecutar las evals	2026-09-15T11:17:54.1671366Z       "llegadas_a_la_red": [],
evals	Ejecutar las evals	2026-09-15T11:17:54.1681392Z       "respuesta": "---\n\n## Artículo 21 de la Ley 39/2015 (LPAC)\n\nEl artículo 21 de la Ley 39/2015 regula la **obligación de resolver** en los procedimientos administrativos. Sus puntos principales son:\n\n**1. Obligación de dictar resolución expresa** [BOE-A-2015-10565, bloque a21]\n\nLa Administración debe dictar una resolución expresa y notificarla en todos los procedimientos, sea cual sea su forma de iniciación. Esta obligación incluye casos especiales (prescripción, renuncia, caducidad, desistimiento, desaparición del objeto) en los que la resolución declara la circunstancia concreta con indicación de hechos y normas aplicables. Se exceptúan los procedimientos que terminan por pacto o convenio, y los relativos a derechos sometidos solo a declaración responsable o comunicación.\n\n**2. Plazo máximo de resolución** [BOE-A-2015-10565, bloque a21]\n\nEl plazo será el que fije la norma reguladora de cada procedimiento, con un límite de **seis meses** (ampliable solo si lo establece una ley o el Derecho de la UE). Si la norma no especifica plazo, será de **tres meses**. Se cuentan desde el acuerdo de iniciación de oficio o desde la entrada de la solicitud en el registro electrónico.\n\n**3. Publicidad y comunicación de plazos** [BOE-A-2015-10565, bloque a21]\n\nLas Administraciones deben publicar en sus portales web los procedimientos de su competencia con sus plazos máximos y efectos del silencio administrativo. Además, informarán al interesado del plazo máximo y efectos en la notificación o comunicación (dentro de 10 días desde la recepción de la solicitud).\n\n**4. Responsabilidad por incumplimiento** [BOE-A-2015-10565, bloque a21]\n\nEl personal administrativo y los titulares de órganos competentes son directamente responsables del cumplimiento de esta obligación. El incumplimiento genera responsabilidad disciplinaria.",
evals	Ejecutar las evals	2026-09-15T11:17:54.1688244Z       "codigo_de_la_sesion": 0,
evals	Ejecutar las evals	2026-09-15T11:17:54.1688581Z       "fin_de_la_sesion": "result success",
evals	Ejecutar las evals	2026-09-15T11:17:54.1688922Z       "sesion_terminada": true,
evals	Ejecutar las evals	2026-09-15T11:17:54.1689201Z       "motivos": [],
evals	Ejecutar las evals	2026-09-15T11:17:54.1689444Z       "pasa": true
evals	Ejecutar las evals	2026-09-15T11:17:54.1689663Z     },
evals	Ejecutar las evals	2026-09-15T11:17:54.1689848Z     {
evals	Ejecutar las evals	2026-09-15T11:17:54.1690123Z       "sesion": "02-lcsp-contrato-menor",
evals	Ejecutar las evals	2026-09-15T11:17:54.1690496Z       "eval": "02-lcsp-contrato-menor.yaml",
evals	Ejecutar las evals	2026-09-15T11:17:54.1690822Z       "activa": true,
evals	Ejecutar las evals	2026-09-15T11:17:54.1691077Z       "activada": true,
evals	Ejecutar las evals	2026-09-15T11:17:54.1691368Z       "comandos_ejecutados": [
evals	Ejecutar las evals	2026-09-15T11:17:54.1691724Z         "bloque boe BOE-A-2017-12902 a1-30"
evals	Ejecutar las evals	2026-09-15T11:17:54.1692006Z       ],
evals	Ejecutar las evals	2026-09-15T11:17:54.1692251Z       "comandos_ausentes": [],
evals	Ejecutar las evals	2026-09-15T11:17:54.1692556Z       "citas_encontradas": [
evals	Ejecutar las evals	2026-09-15T11:17:54.1692861Z         "BOE-A-2017-12902 a1-30"
evals	Ejecutar las evals	2026-09-15T11:17:54.1693230Z       ],
evals	Ejecutar las evals	2026-09-15T11:17:54.1693466Z       "citas_ausentes": [],
evals	Ejecutar las evals	2026-09-15T11:17:54.1693751Z       "invocaciones": [
evals	Ejecutar las evals	2026-09-15T11:17:54.1693990Z         {
evals	Ejecutar las evals	2026-09-15T11:17:54.1694346Z           "orden": "boe indice BOE-A-2017-12902 --json",
evals	Ejecutar las evals	2026-09-15T11:17:54.1694699Z           "codigo": 0,
evals	Ejecutar las evals	2026-09-15T11:17:54.1694977Z           "conexiones": []
evals	Ejecutar las evals	2026-09-15T11:17:54.1695222Z         },
evals	Ejecutar las evals	2026-09-15T11:17:54.1695435Z         {
evals	Ejecutar las evals	2026-09-15T11:17:54.1695824Z           "orden": "boe articulo BOE-A-2017-12902 a1-30 --json",
evals	Ejecutar las evals	2026-09-15T11:17:54.1696195Z           "codigo": 0,
evals	Ejecutar las evals	2026-09-15T11:17:54.1696476Z           "conexiones": []
evals	Ejecutar las evals	2026-09-15T11:17:54.1696720Z         }
evals	Ejecutar las evals	2026-09-15T11:17:54.1696915Z       ],
evals	Ejecutar las evals	2026-09-15T11:17:54.1697170Z       "fuera_de_lo_grabado": [],
evals	Ejecutar las evals	2026-09-15T11:17:54.1697468Z       "otras_fallidas": [],
evals	Ejecutar las evals	2026-09-15T11:17:54.1697754Z       "llegadas_a_la_red": [],
evals	Ejecutar las evals	2026-09-15T11:17:54.1706250Z       "respuesta": "Según el artículo 118 de la LCSP, el expediente de un contrato menor debe incluir los siguientes elementos:\n\n## Requisitos generales (obligatorios en todos los contratos menores):\n\n1. **Informe justificativo**: Un informe del órgano de contratación que justifique de manera motivada:\n   - La necesidad del contrato\n   - Que no se está alterando su objeto para evitar la aplicación de los umbrales de contratación [art. 118.2, LCSP [BOE-A-2017-12902, bloque a1-30]]\n\n2. **Aprobación del gasto**: La aprobación del gasto con incorporación al expediente de la factura correspondiente, reuniendo los requisitos que establezcan las normas de desarrollo de la LCSP [art. 118.3, LCSP [BOE-A-2017-12902, bloque a1-30]]\n\n## Requisitos específicos para contratos menores de obras:\n\n3. **Presupuesto**: El presupuesto de las obras [art. 118.4, LCSP [BOE-A-2017-12902, bloque a1-30]]\n\n4. **Proyecto**: El correspondiente proyecto, cuando sea requerido por las disposiciones vigentes [art. 118.4, LCSP [BOE-A-2017-12902, bloque a1-30]]\n\n5. **Informe de supervisión**: El informe de las oficinas o unidades de supervisión cuando el trabajo afecte a la estabilidad, seguridad o estanqueidad de la obra [art. 118.4, LCSP [BOE-A-2017-12902, bloque a1-30]]\n\n## Excepción:\n\nEl requisito del informe justificativo (apartado 2) **no será de aplicación** en contratos cuyo pago se verifique a través de anticipos de caja fija u otro sistema similar para pagos menores, siempre que el valor no exceda de 5.000 euros [art. 118.5, LCSP [BOE-A-2017-12902, bloque a1-30]]",
evals	Ejecutar las evals	2026-09-15T11:17:54.1711975Z       "codigo_de_la_sesion": 0,
evals	Ejecutar las evals	2026-09-15T11:17:54.1712315Z       "fin_de_la_sesion": "result success",
evals	Ejecutar las evals	2026-09-15T11:17:54.1712651Z       "sesion_terminada": true,
evals	Ejecutar las evals	2026-09-15T11:17:54.1713051Z       "motivos": [],
evals	Ejecutar las evals	2026-09-15T11:17:54.1713310Z       "pasa": true
evals	Ejecutar las evals	2026-09-15T11:17:54.1713524Z     },
evals	Ejecutar las evals	2026-09-15T11:17:54.1713709Z     {
evals	Ejecutar las evals	2026-09-15T11:17:54.1714021Z       "sesion": "03-lrbrl-atribuciones-del-pleno",
evals	Ejecutar las evals	2026-09-15T11:17:54.1714602Z       "eval": "03-lrbrl-atribuciones-del-pleno.yaml",
evals	Ejecutar las evals	2026-09-15T11:17:54.1714953Z       "activa": true,
evals	Ejecutar las evals	2026-09-15T11:17:54.1715218Z       "activada": true,
evals	Ejecutar las evals	2026-09-15T11:17:54.1715504Z       "comandos_ejecutados": [
evals	Ejecutar las evals	2026-09-15T11:17:54.1715840Z         "bloque boe BOE-A-1985-5392 a22"
evals	Ejecutar las evals	2026-09-15T11:17:54.1716136Z       ],
evals	Ejecutar las evals	2026-09-15T11:17:54.1716382Z       "comandos_ausentes": [],
evals	Ejecutar las evals	2026-09-15T11:17:54.1716677Z       "citas_encontradas": [
evals	Ejecutar las evals	2026-09-15T11:17:54.1716979Z         "BOE-A-1985-5392 a22"
evals	Ejecutar las evals	2026-09-15T11:17:54.1717230Z       ],
evals	Ejecutar las evals	2026-09-15T11:17:54.1717462Z       "citas_ausentes": [],
evals	Ejecutar las evals	2026-09-15T11:17:54.1717741Z       "invocaciones": [
evals	Ejecutar las evals	2026-09-15T11:17:54.1717983Z         {
evals	Ejecutar las evals	2026-09-15T11:17:54.1718329Z           "orden": "boe indice BOE-A-1985-5392 --json",
evals	Ejecutar las evals	2026-09-15T11:17:54.1718681Z           "codigo": 0,
evals	Ejecutar las evals	2026-09-15T11:17:54.1718971Z           "conexiones": []
evals	Ejecutar las evals	2026-09-15T11:17:54.1719220Z         },
evals	Ejecutar las evals	2026-09-15T11:17:54.1719423Z         {
evals	Ejecutar las evals	2026-09-15T11:17:54.1719792Z           "orden": "boe articulo BOE-A-1985-5392 a22 --json",
evals	Ejecutar las evals	2026-09-15T11:17:54.1720161Z           "codigo": 0,
evals	Ejecutar las evals	2026-09-15T11:17:54.1720439Z           "conexiones": []
evals	Ejecutar las evals	2026-09-15T11:17:54.1720677Z         }
evals	Ejecutar las evals	2026-09-15T11:17:54.1720880Z       ],
evals	Ejecutar las evals	2026-09-15T11:17:54.1721126Z       "fuera_de_lo_grabado": [],
evals	Ejecutar las evals	2026-09-15T11:17:54.1721419Z       "otras_fallidas": [],
evals	Ejecutar las evals	2026-09-15T11:17:54.1721701Z       "llegadas_a_la_red": [],
evals	Ejecutar las evals	2026-09-15T11:17:54.1733451Z       "respuesta": "Según el **artículo 22.2 de la Ley 7/1985, de 2 de abril, Reguladora de las Bases del Régimen Local** [BOE-A-1985-5392, bloque a22], el Pleno municipal tiene las siguientes atribuciones:\n\n**Atribuciones principales del Pleno:**\n\na) **Control y fiscalización** de los órganos de gobierno.\n\nb) **Decisiones supramunicipales y territoriales**: participación en organizaciones supramunicipales; alteración del término municipal; creación o supresión de municipios; creación de órganos desconcentrados; cambios de capitalidad del municipio, nombre, bandera, enseña o escudo.\n\nc) **Ordenación territorial**: aprobación inicial del planeamiento general y aprobación de planes e instrumentos de ordenación urbanística.\n\nd) **Normativa municipal**: aprobación del reglamento orgánico y ordenanzas.\n\ne) **Hacienda municipal**: determinación de recursos tributarios propios; aprobación y modificación de presupuestos; disposición de gastos; aprobación de cuentas.\n\nf) **Gestión de servicios**: aprobación de formas de gestión de servicios públicos municipales.\n\ng) **Delegaciones de competencias** de otras Administraciones públicas.\n\nh) **Conflictos competenciales** con otras entidades locales y Administraciones públicas.\n\ni) **Personal municipal**: aprobación de plantillas; fijación de retribuciones complementarias; régimen del personal eventual.\n\nj) **Defensa judicial**: ejercicio de acciones judiciales y administrativas en cuestiones de competencia plenaria.\n\nk) **Lesividad**: declaración de lesividad de actos del Ayuntamiento.\n\nl) **Bienes públicos**: alteración de la calificación jurídica de bienes de dominio público.\n\nm) **Operaciones de crédito** que excedan umbrales de endeudamiento.\n\nñ) **Obras y servicios**: aprobación de proyectos no previstos en presupuestos.\n\np) **Mayorías especiales**: aquellas cuestiones que así lo requieran.\n\nq) **Competencias adicionales** que fijen otras leyes.\n\nAdicionalmente, el **artículo 22.3** atribuye al Pleno la votación sobre mociones de censura al Alcalde y cuestiones de confianza, mediante votación pública por llamamiento nominal [BOE-A-1985-5392, bloque a22].",
evals	Ejecutar las evals	2026-09-15T11:17:54.1741123Z       "codigo_de_la_sesion": 0,
evals	Ejecutar las evals	2026-09-15T11:17:54.1741467Z       "fin_de_la_sesion": "result success",
evals	Ejecutar las evals	2026-09-15T11:17:54.1741808Z       "sesion_terminada": true,
evals	Ejecutar las evals	2026-09-15T11:17:54.1742090Z       "motivos": [],
evals	Ejecutar las evals	2026-09-15T11:17:54.1742340Z       "pasa": true
evals	Ejecutar las evals	2026-09-15T11:17:54.1742568Z     },
evals	Ejecutar las evals	2026-09-15T11:17:54.1742754Z     {
evals	Ejecutar las evals	2026-09-15T11:17:54.1743136Z       "sesion": "04-lgt-prescripcion",
evals	Ejecutar las evals	2026-09-15T11:17:54.1743624Z       "eval": "04-lgt-prescripcion.yaml",
evals	Ejecutar las evals	2026-09-15T11:17:54.1743940Z       "activa": true,
evals	Ejecutar las evals	2026-09-15T11:17:54.1744197Z       "activada": true,
evals	Ejecutar las evals	2026-09-15T11:17:54.1744490Z       "comandos_ejecutados": [
evals	Ejecutar las evals	2026-09-15T11:17:54.1744827Z         "bloque boe BOE-A-2003-23186 a66"
evals	Ejecutar las evals	2026-09-15T11:17:54.1745107Z       ],
evals	Ejecutar las evals	2026-09-15T11:17:54.1745353Z       "comandos_ausentes": [],
evals	Ejecutar las evals	2026-09-15T11:17:54.1745652Z       "citas_encontradas": [
evals	Ejecutar las evals	2026-09-15T11:17:54.1745964Z         "BOE-A-2003-23186 a66"
evals	Ejecutar las evals	2026-09-15T11:17:54.1746207Z       ],
evals	Ejecutar las evals	2026-09-15T11:17:54.1746443Z       "citas_ausentes": [],
evals	Ejecutar las evals	2026-09-15T11:17:54.1746718Z       "invocaciones": [
evals	Ejecutar las evals	2026-09-15T11:17:54.1746964Z         {
evals	Ejecutar las evals	2026-09-15T11:17:54.1747320Z           "orden": "boe indice BOE-A-2003-23186 --json",
evals	Ejecutar las evals	2026-09-15T11:17:54.1747678Z           "codigo": 0,
evals	Ejecutar las evals	2026-09-15T11:17:54.1747968Z           "conexiones": []
evals	Ejecutar las evals	2026-09-15T11:17:54.1748215Z         },
evals	Ejecutar las evals	2026-09-15T11:17:54.1748422Z         {
evals	Ejecutar las evals	2026-09-15T11:17:54.1748798Z           "orden": "boe articulo BOE-A-2003-23186 a66 --json",
evals	Ejecutar las evals	2026-09-15T11:17:54.1749178Z           "codigo": 0,
evals	Ejecutar las evals	2026-09-15T11:17:54.1749458Z           "conexiones": []
evals	Ejecutar las evals	2026-09-15T11:17:54.1749695Z         }
evals	Ejecutar las evals	2026-09-15T11:17:54.1749895Z       ],
evals	Ejecutar las evals	2026-09-15T11:17:54.1750146Z       "fuera_de_lo_grabado": [],
evals	Ejecutar las evals	2026-09-15T11:17:54.1750442Z       "otras_fallidas": [],
evals	Ejecutar las evals	2026-09-15T11:17:54.1750733Z       "llegadas_a_la_red": [],
evals	Ejecutar las evals	2026-09-15T11:17:54.1753218Z       "respuesta": "El derecho de la Administración a liquidar una deuda tributaria prescribe en **4 años**, según el artículo 66 de la Ley General Tributaria:\n\n> Prescribirán a los cuatro años los siguientes derechos:\n> a) El derecho de la Administración para determinar la deuda tributaria mediante la oportuna liquidación.\n\nart. 66 de la Ley 58/2003, de 17 de diciembre, General Tributaria [BOE-A-2003-23186, bloque a66]",
evals	Ejecutar las evals	2026-09-15T11:17:54.1754781Z       "codigo_de_la_sesion": 0,
evals	Ejecutar las evals	2026-09-15T11:17:54.1755113Z       "fin_de_la_sesion": "result success",
evals	Ejecutar las evals	2026-09-15T11:17:54.1755447Z       "sesion_terminada": true,
evals	Ejecutar las evals	2026-09-15T11:17:54.1755727Z       "motivos": [],
evals	Ejecutar las evals	2026-09-15T11:17:54.1755976Z       "pasa": true
evals	Ejecutar las evals	2026-09-15T11:17:54.1756190Z     },
evals	Ejecutar las evals	2026-09-15T11:17:54.1756379Z     {
evals	Ejecutar las evals	2026-09-15T11:17:54.1756692Z       "sesion": "05-trlrhl-impuestos-municipales",
evals	Ejecutar las evals	2026-09-15T11:17:54.1757133Z       "eval": "05-trlrhl-impuestos-municipales.yaml",
evals	Ejecutar las evals	2026-09-15T11:17:54.1757480Z       "activa": true,
evals	Ejecutar las evals	2026-09-15T11:17:54.1757734Z       "activada": true,
evals	Ejecutar las evals	2026-09-15T11:17:54.1758019Z       "comandos_ejecutados": [
evals	Ejecutar las evals	2026-09-15T11:17:54.1758358Z         "bloque boe BOE-A-2004-4214 a59"
evals	Ejecutar las evals	2026-09-15T11:17:54.1758639Z       ],
evals	Ejecutar las evals	2026-09-15T11:17:54.1758891Z       "comandos_ausentes": [],
evals	Ejecutar las evals	2026-09-15T11:17:54.1759185Z       "citas_encontradas": [
evals	Ejecutar las evals	2026-09-15T11:17:54.1759492Z         "BOE-A-2004-4214 a59"
evals	Ejecutar las evals	2026-09-15T11:17:54.1759738Z       ],
evals	Ejecutar las evals	2026-09-15T11:17:54.1760118Z       "citas_ausentes": [],
evals	Ejecutar las evals	2026-09-15T11:17:54.1760396Z       "invocaciones": [
evals	Ejecutar las evals	2026-09-15T11:17:54.1760634Z         {
evals	Ejecutar las evals	2026-09-15T11:17:54.1760994Z           "orden": "boe indice BOE-A-2004-4214 --json",
evals	Ejecutar las evals	2026-09-15T11:17:54.1761350Z           "codigo": 0,
evals	Ejecutar las evals	2026-09-15T11:17:54.1761631Z           "conexiones": []
evals	Ejecutar las evals	2026-09-15T11:17:54.1761884Z         },
evals	Ejecutar las evals	2026-09-15T11:17:54.1762090Z         {
evals	Ejecutar las evals	2026-09-15T11:17:54.1762466Z           "orden": "boe articulo BOE-A-2004-4214 a59 --json",
evals	Ejecutar las evals	2026-09-15T11:17:54.1762832Z           "codigo": 0,
evals	Ejecutar las evals	2026-09-15T11:17:54.1763238Z           "conexiones": []
evals	Ejecutar las evals	2026-09-15T11:17:54.1763483Z         }
evals	Ejecutar las evals	2026-09-15T11:17:54.1763680Z       ],
evals	Ejecutar las evals	2026-09-15T11:17:54.1763930Z       "fuera_de_lo_grabado": [],
evals	Ejecutar las evals	2026-09-15T11:17:54.1764226Z       "otras_fallidas": [],
evals	Ejecutar las evals	2026-09-15T11:17:54.1764511Z       "llegadas_a_la_red": [],
evals	Ejecutar las evals	2026-09-15T11:17:54.1769049Z       "respuesta": "Según el artículo 59 del TRLRHL, los ayuntamientos pueden exigir los siguientes impuestos:\n\n**Impuestos de exigencia obligatoria:**\n\n1. Impuesto sobre Bienes Inmuebles\n2. Impuesto sobre Actividades Económicas\n3. Impuesto sobre Vehículos de Tracción Mecánica\n\n**Impuestos de exigencia facultativa:**\n\n4. Impuesto sobre Construcciones, Instalaciones y Obras\n5. Impuesto sobre el Incremento de Valor de los Terrenos de Naturaleza Urbana\n\nart. 59 del Real Decreto Legislativo 2/2004 [BOE-A-2004-4214, bloque a59]\n\nLos impuestos obligatorios (apartado 1) deben ser establecidos y exigidos por los ayuntamientos de acuerdo con la ley y las disposiciones que la desarrollan. Los dos impuestos del apartado 2 son optativos y se establecerán conforme a las respectivas ordenanzas fiscales municipales.",
evals	Ejecutar las evals	2026-09-15T11:17:54.1772104Z       "codigo_de_la_sesion": 0,
evals	Ejecutar las evals	2026-09-15T11:17:54.1772436Z       "fin_de_la_sesion": "result success",
evals	Ejecutar las evals	2026-09-15T11:17:54.1772772Z       "sesion_terminada": true,
evals	Ejecutar las evals	2026-09-15T11:17:54.1773185Z       "motivos": [],
evals	Ejecutar las evals	2026-09-15T11:17:54.1773434Z       "pasa": true
evals	Ejecutar las evals	2026-09-15T11:17:54.1773657Z     },
evals	Ejecutar las evals	2026-09-15T11:17:54.1773840Z     {
evals	Ejecutar las evals	2026-09-15T11:17:54.1774150Z       "sesion": "06-irpf-rendimientos-del-trabajo",
evals	Ejecutar las evals	2026-09-15T11:17:54.1774592Z       "eval": "06-irpf-rendimientos-del-trabajo.yaml",
evals	Ejecutar las evals	2026-09-15T11:17:54.1774940Z       "activa": true,
evals	Ejecutar las evals	2026-09-15T11:17:54.1775193Z       "activada": true,
evals	Ejecutar las evals	2026-09-15T11:17:54.1775478Z       "comandos_ejecutados": [
evals	Ejecutar las evals	2026-09-15T11:17:54.1775812Z         "bloque boe BOE-A-2006-20764 a17"
evals	Ejecutar las evals	2026-09-15T11:17:54.1776093Z       ],
evals	Ejecutar las evals	2026-09-15T11:17:54.1776336Z       "comandos_ausentes": [],
evals	Ejecutar las evals	2026-09-15T11:17:54.1776634Z       "citas_encontradas": [
evals	Ejecutar las evals	2026-09-15T11:17:54.1776927Z         "BOE-A-2006-20764 a17"
evals	Ejecutar las evals	2026-09-15T11:17:54.1777169Z       ],
evals	Ejecutar las evals	2026-09-15T11:17:54.1777409Z       "citas_ausentes": [],
evals	Ejecutar las evals	2026-09-15T11:17:54.1777685Z       "invocaciones": [
evals	Ejecutar las evals	2026-09-15T11:17:54.1777920Z         {
evals	Ejecutar las evals	2026-09-15T11:17:54.1778278Z           "orden": "boe indice BOE-A-2006-20764 --json",
evals	Ejecutar las evals	2026-09-15T11:17:54.1778625Z           "codigo": 0,
evals	Ejecutar las evals	2026-09-15T11:17:54.1778904Z           "conexiones": []
evals	Ejecutar las evals	2026-09-15T11:17:54.1779152Z         },
evals	Ejecutar las evals	2026-09-15T11:17:54.1779363Z         {
evals	Ejecutar las evals	2026-09-15T11:17:54.1779734Z           "orden": "boe articulo BOE-A-2006-20764 a17 --json",
evals	Ejecutar las evals	2026-09-15T11:17:54.1780100Z           "codigo": 0,
evals	Ejecutar las evals	2026-09-15T11:17:54.1780380Z           "conexiones": []
evals	Ejecutar las evals	2026-09-15T11:17:54.1780621Z         }
evals	Ejecutar las evals	2026-09-15T11:17:54.1780815Z       ],
evals	Ejecutar las evals	2026-09-15T11:17:54.1781062Z       "fuera_de_lo_grabado": [],
evals	Ejecutar las evals	2026-09-15T11:17:54.1781363Z       "otras_fallidas": [],
evals	Ejecutar las evals	2026-09-15T11:17:54.1781645Z       "llegadas_a_la_red": [],
evals	Ejecutar las evals	2026-09-15T11:17:54.1792515Z       "respuesta": "Según el artículo 17 de la Ley 35/2006, del Impuesto sobre la Renta de las Personas Físicas (LIRPF), se consideran **rendimientos íntegros del trabajo** [BOE-A-2006-20764, bloque a17]:\n\n## Concepto general (apartado 1)\n\nSe consideran rendimientos íntegros del trabajo todas las **contraprestaciones o utilidades** —dinerarias o en especie, cualquiera que sea su denominación o naturaleza— que deriven, directa o indirectamente, del trabajo personal o de la relación laboral o estatutaria, siempre que **no tengan carácter de rendimientos de actividades económicas**.\n\n## Incluye expresamente:\n\na) **Sueldos y salarios**\nb) **Prestaciones por desempleo**\nc) **Remuneraciones por gastos de representación**\nd) **Dietas y asignaciones para gastos de viaje** (excepto locomoción y gastos normales de manutención/estancia en hostelería, con límites reglamentarios)\ne) **Aportaciones a planes de pensiones** (realizadas por promotores o empresas según regulación de fondos de pensiones)\nf) **Aportaciones empresariales para compromisos por pensiones** (en seguros colectivos, con requisitos de imputación)\n\n## Prestaciones consideradas rendimientos del trabajo (apartado 2.a):\n\n- Pensiones y haberes pasivos de regímenes públicos (Seguridad Social, clases pasivas)\n- Prestaciones de mutualidades de funcionarios y colegios de huérfanos\n- Prestaciones de planes de pensiones\n- Prestaciones de seguros de mutualidades de previsión social\n- Prestaciones de planes de previsión social empresarial\n- Prestaciones de seguros de dependencia\n\nTambién se incluyen (apartado 2.b a 2.k):\n- Retribuciones de diputados, senadores y concejales\n- Rendimientos por cursos, conferencias y seminarios\n- Rendimientos de obras literarias, artísticas o científicas (con cesión de derechos)\n- Retribuciones de administradores y consejeros\n- Pensiones compensatorias y alimentos\n- Becas\n- Retribuciones por actividades humanitarias\n- Retribuciones de relaciones laborales especiales",
evals	Ejecutar las evals	2026-09-15T11:17:54.1799968Z       "codigo_de_la_sesion": 0,
evals	Ejecutar las evals	2026-09-15T11:17:54.1800309Z       "fin_de_la_sesion": "result success",
evals	Ejecutar las evals	2026-09-15T11:17:54.1800648Z       "sesion_terminada": true,
evals	Ejecutar las evals	2026-09-15T11:17:54.1800926Z       "motivos": [],
evals	Ejecutar las evals	2026-09-15T11:17:54.1801170Z       "pasa": true
evals	Ejecutar las evals	2026-09-15T11:17:54.1801394Z     },
evals	Ejecutar las evals	2026-09-15T11:17:54.1801584Z     {
evals	Ejecutar las evals	2026-09-15T11:17:54.1801890Z       "sesion": "07-lrjsp-principio-de-legalidad",
evals	Ejecutar las evals	2026-09-15T11:17:54.1802332Z       "eval": "07-lrjsp-principio-de-legalidad.yaml",
evals	Ejecutar las evals	2026-09-15T11:17:54.1802684Z       "activa": true,
evals	Ejecutar las evals	2026-09-15T11:17:54.1803038Z       "activada": true,
evals	Ejecutar las evals	2026-09-15T11:17:54.1803330Z       "comandos_ejecutados": [
evals	Ejecutar las evals	2026-09-15T11:17:54.1803677Z         "bloque boe BOE-A-2015-10566 a25"
evals	Ejecutar las evals	2026-09-15T11:17:54.1803960Z       ],
evals	Ejecutar las evals	2026-09-15T11:17:54.1804205Z       "comandos_ausentes": [],
evals	Ejecutar las evals	2026-09-15T11:17:54.1804505Z       "citas_encontradas": [
evals	Ejecutar las evals	2026-09-15T11:17:54.1804801Z         "BOE-A-2015-10566 a25"
evals	Ejecutar las evals	2026-09-15T11:17:54.1805044Z       ],
evals	Ejecutar las evals	2026-09-15T11:17:54.1805287Z       "citas_ausentes": [],
evals	Ejecutar las evals	2026-09-15T11:17:54.1805562Z       "invocaciones": [
evals	Ejecutar las evals	2026-09-15T11:17:54.1805801Z         {
evals	Ejecutar las evals	2026-09-15T11:17:54.1806160Z           "orden": "boe indice BOE-A-2015-10566 --json",
evals	Ejecutar las evals	2026-09-15T11:17:54.1806520Z           "codigo": 0,
evals	Ejecutar las evals	2026-09-15T11:17:54.1806799Z           "conexiones": []
evals	Ejecutar las evals	2026-09-15T11:17:54.1807081Z         },
evals	Ejecutar las evals	2026-09-15T11:17:54.1807290Z         {
evals	Ejecutar las evals	2026-09-15T11:17:54.1807667Z           "orden": "boe articulo BOE-A-2015-10566 a25 --json",
evals	Ejecutar las evals	2026-09-15T11:17:54.1808033Z           "codigo": 0,
evals	Ejecutar las evals	2026-09-15T11:17:54.1808311Z           "conexiones": []
evals	Ejecutar las evals	2026-09-15T11:17:54.1808556Z         }
evals	Ejecutar las evals	2026-09-15T11:17:54.1808754Z       ],
evals	Ejecutar las evals	2026-09-15T11:17:54.1808998Z       "fuera_de_lo_grabado": [],
evals	Ejecutar las evals	2026-09-15T11:17:54.1809302Z       "otras_fallidas": [],
evals	Ejecutar las evals	2026-09-15T11:17:54.1809584Z       "llegadas_a_la_red": [],
evals	Ejecutar las evals	2026-09-15T11:17:54.1818785Z       "respuesta": "El art. 25 de la Ley 40/2015 [BOE-A-2015-10566, bloque a25] establece los siguientes aspectos del principio de legalidad en la potestad sancionadora:\n\n**Apartado 1 — Requisito de habilitación legal**\nLa potestad sancionadora de las Administraciones Públicas solo puede ejercerse cuando haya sido **expresamente reconocida por una norma con rango de Ley**, siguiendo el procedimiento previsto para ello. Además, debe aplicarse conforme a lo establecido en:\n- La propia Ley 40/2015\n- La Ley 39/2015 de Procedimiento Administrativo Común\n- Para Entidades Locales, el Título XI de la Ley 7/1985, de Reguladora de las Bases del Régimen Local\n\n**Apartado 2 — Atribución a órganos competentes**\nEl ejercicio de la potestad sancionadora corresponde exclusivamente a los órganos administrativos que la tengan **expresamente atribuida** por disposición de rango legal o reglamentario.\n\n**Apartado 3 — Extensión a potestad disciplinaria**\nLas disposiciones del Capítulo III de la Ley 40/2015 son también aplicables al ejercicio de la potestad disciplinaria sobre el personal al servicio de las Administraciones Públicas, independientemente de la naturaleza jurídica de la relación de empleo.\n\n**Apartado 4 — Exclusiones**\nLas disposiciones de este capítulo **no aplican** cuando el ejercicio de la potestad sancionadora afecte a:\n- Personas vinculadas por relaciones reguladas por legislación de contratos del sector público\n- Personas vinculadas por legislación patrimonial de las Administraciones Públicas\n\nEste artículo refleja la garantía constitucional de legalidad administrativa, exigiendo siempre una base legal explícita y un órgano competente para sancionar.",
evals	Ejecutar las evals	2026-09-15T11:17:54.1825013Z       "codigo_de_la_sesion": 0,
evals	Ejecutar las evals	2026-09-15T11:17:54.1825345Z       "fin_de_la_sesion": "result success",
evals	Ejecutar las evals	2026-09-15T11:17:54.1825813Z       "sesion_terminada": true,
evals	Ejecutar las evals	2026-09-15T11:17:54.1826103Z       "motivos": [],
evals	Ejecutar las evals	2026-09-15T11:17:54.1826347Z       "pasa": true
evals	Ejecutar las evals	2026-09-15T11:17:54.1826566Z     },
evals	Ejecutar las evals	2026-09-15T11:17:54.1826758Z     {
evals	Ejecutar las evals	2026-09-15T11:17:54.1827057Z       "sesion": "08-ltaibg-plazo-de-resolucion",
evals	Ejecutar las evals	2026-09-15T11:17:54.1827475Z       "eval": "08-ltaibg-plazo-de-resolucion.yaml",
evals	Ejecutar las evals	2026-09-15T11:17:54.1827822Z       "activa": true,
evals	Ejecutar las evals	2026-09-15T11:17:54.1828081Z       "activada": true,
evals	Ejecutar las evals	2026-09-15T11:17:54.1828363Z       "comandos_ejecutados": [
evals	Ejecutar las evals	2026-09-15T11:17:54.1828700Z         "bloque boe BOE-A-2013-12887 a20"
evals	Ejecutar las evals	2026-09-15T11:17:54.1828997Z       ],
evals	Ejecutar las evals	2026-09-15T11:17:54.1829254Z       "comandos_ausentes": [],
evals	Ejecutar las evals	2026-09-15T11:17:54.1829556Z       "citas_encontradas": [],
evals	Ejecutar las evals	2026-09-15T11:17:54.1829858Z       "citas_ausentes": [
evals	Ejecutar las evals	2026-09-15T11:17:54.1830170Z         "BOE-A-2013-12887 a20"
evals	Ejecutar las evals	2026-09-15T11:17:54.1830422Z       ],
evals	Ejecutar las evals	2026-09-15T11:17:54.1830656Z       "invocaciones": [
evals	Ejecutar las evals	2026-09-15T11:17:54.1830904Z         {
evals	Ejecutar las evals	2026-09-15T11:17:54.1831260Z           "orden": "boe indice BOE-A-2013-12887 --json",
evals	Ejecutar las evals	2026-09-15T11:17:54.1831612Z           "codigo": 0,
evals	Ejecutar las evals	2026-09-15T11:17:54.1831900Z           "conexiones": []
evals	Ejecutar las evals	2026-09-15T11:17:54.1832146Z         },
evals	Ejecutar las evals	2026-09-15T11:17:54.1832349Z         {
evals	Ejecutar las evals	2026-09-15T11:17:54.1832726Z           "orden": "boe articulo BOE-A-2013-12887 a20 --json",
evals	Ejecutar las evals	2026-09-15T11:17:54.1833211Z           "codigo": 0,
evals	Ejecutar las evals	2026-09-15T11:17:54.1833495Z           "conexiones": []
evals	Ejecutar las evals	2026-09-15T11:17:54.1833744Z         }
evals	Ejecutar las evals	2026-09-15T11:17:54.1833944Z       ],
evals	Ejecutar las evals	2026-09-15T11:17:54.1834193Z       "fuera_de_lo_grabado": [],
evals	Ejecutar las evals	2026-09-15T11:17:54.1834493Z       "otras_fallidas": [],
evals	Ejecutar las evals	2026-09-15T11:17:54.1834781Z       "llegadas_a_la_red": [],
evals	Ejecutar las evals	2026-09-15T11:17:54.1839629Z       "respuesta": "Según el artículo 20 de la Ley 19/2013, el plazo para resolver una solicitud de acceso a la información pública es el siguiente:\n\n**Plazo ordinario:** un mes desde la recepción de la solicitud por el órgano competente para resolver [art. 20.1 de la LTAIBG, BOE-A-2013-12887, bloque a20].\n\n**Ampliación:** este plazo puede ampliarse por otro mes adicional en caso de que el volumen o la complejidad de la información solicitada así lo justifique, previa notificación al solicitante [art. 20.1 de la LTAIBG, BOE-A-2013-12887, bloque a20].\n\n**Silencio administrativo:** transcurrido el plazo máximo sin haber dictado y notificado una resolución expresa, se entiende que la solicitud ha sido desestimada [art. 20.4 de la LTAIBG, BOE-A-2013-12887, bloque a20].\n\nPor tanto, el plazo general es de **un mes**, prorrogable a **dos meses máximo** según las circunstancias de la solicitud.",
evals	Ejecutar las evals	2026-09-15T11:17:54.1842717Z       "codigo_de_la_sesion": 0,
evals	Ejecutar las evals	2026-09-15T11:17:54.1843156Z       "fin_de_la_sesion": "result success",
evals	Ejecutar las evals	2026-09-15T11:17:54.1843493Z       "sesion_terminada": true,
evals	Ejecutar las evals	2026-09-15T11:17:54.1843911Z       "motivos": [
evals	Ejecutar las evals	2026-09-15T11:17:54.1844220Z         "cita ausente: BOE-A-2013-12887 a20"
evals	Ejecutar las evals	2026-09-15T11:17:54.1844515Z       ],
evals	Ejecutar las evals	2026-09-15T11:17:54.1844740Z       "pasa": false
evals	Ejecutar las evals	2026-09-15T11:17:54.1844956Z     },
evals	Ejecutar las evals	2026-09-15T11:17:54.1845140Z     {
evals	Ejecutar las evals	2026-09-15T11:17:54.1845435Z       "sesion": "09-constitucion-articulo-140",
evals	Ejecutar las evals	2026-09-15T11:17:54.1845845Z       "eval": "09-constitucion-articulo-140.yaml",
evals	Ejecutar las evals	2026-09-15T11:17:54.1846177Z       "activa": true,
evals	Ejecutar las evals	2026-09-15T11:17:54.1846433Z       "activada": true,
evals	Ejecutar las evals	2026-09-15T11:17:54.1846722Z       "comandos_ejecutados": [],
evals	Ejecutar las evals	2026-09-15T11:17:54.1847030Z       "comandos_ausentes": [],
evals	Ejecutar las evals	2026-09-15T11:17:54.1847325Z       "citas_encontradas": [],
evals	Ejecutar las evals	2026-09-15T11:17:54.1847622Z       "citas_ausentes": [],
evals	Ejecutar las evals	2026-09-15T11:17:54.1847901Z       "invocaciones": [],
evals	Ejecutar las evals	2026-09-15T11:17:54.1848194Z       "fuera_de_lo_grabado": [],
evals	Ejecutar las evals	2026-09-15T11:17:54.1848484Z       "otras_fallidas": [],
evals	Ejecutar las evals	2026-09-15T11:17:54.1848773Z       "llegadas_a_la_red": [],
evals	Ejecutar las evals	2026-09-15T11:17:54.1855051Z       "respuesta": "El artículo 140 de la Constitución Española establece:\n\n> **Artículo 140.** La Constitución garantiza la autonomía de los municipios. Estos gozarán de personalidad jurídica plena. Su gobierno y administración corresponde a sus respectivos Ayuntamientos, integrados por los Alcaldes y los Concejales. Los Concejales serán elegidos por los vecinos del municipio mediante sufragio universal, igual, libre, directo y secreto, en la forma establecida por la ley. Los Alcaldes serán elegidos por los Concejales o por los vecinos. La ley regulará las condiciones en las que proceda el régimen del concejo abierto.\n\nart. 140 de la Constitución Española [BOE-A-1978-31229, bloque a140]\n\nEn síntesis, este artículo reconoce la autonomía municipal, establece que los municipios tienen personalidad jurídica plena, define la estructura de gobierno (Ayuntamiento compuesto por Alcaldes y Concejales), y prescribe que los Concejales sean elegidos por los vecinos mediante sufragio universal, directo y secreto, mientras que los Alcaldes pueden ser elegidos por los Concejales o por los vecinos según determine la ley.",
evals	Ejecutar las evals	2026-09-15T11:17:54.1859318Z       "codigo_de_la_sesion": 0,
evals	Ejecutar las evals	2026-09-15T11:17:54.1859659Z       "fin_de_la_sesion": "result success",
evals	Ejecutar las evals	2026-09-15T11:17:54.1859993Z       "sesion_terminada": true,
evals	Ejecutar las evals	2026-09-15T11:17:54.1860265Z       "motivos": [
evals	Ejecutar las evals	2026-09-15T11:17:54.1863044Z         "sesión ilegible: traza: traza ilegible: /home/runner/work/_temp/kitlegal-evals-boe-legislacion/sesiones/09-constitucion-articulo-140/traza/t.14465, línea 1: no es ninguna de las formas de línea de la traza: execve, clone, clone3, fork, vfork o connect con su resultado, una señal o la línea final: «clone(child_stack=0x2a559d472000, flags=CLONE_VM|CLONE_FS|CLONE_FILES|CLONE_SIGHAND|CLONE_THREAD|CLONE_SYSVSEM|CLONE_SETTLS <unfinished ...>) = ?»"
evals	Ejecutar las evals	2026-09-15T11:17:54.1864686Z       ],
evals	Ejecutar las evals	2026-09-15T11:17:54.1864909Z       "pasa": false
evals	Ejecutar las evals	2026-09-15T11:17:54.1865135Z     },
evals	Ejecutar las evals	2026-09-15T11:17:54.1865329Z     {
evals	Ejecutar las evals	2026-09-15T11:17:54.1865584Z       "sesion": "10-et-vacaciones",
evals	Ejecutar las evals	2026-09-15T11:17:54.1865923Z       "eval": "10-et-vacaciones.yaml",
evals	Ejecutar las evals	2026-09-15T11:17:54.1866229Z       "activa": true,
evals	Ejecutar las evals	2026-09-15T11:17:54.1866487Z       "activada": true,
evals	Ejecutar las evals	2026-09-15T11:17:54.1866768Z       "comandos_ejecutados": [
evals	Ejecutar las evals	2026-09-15T11:17:54.1867114Z         "bloque boe BOE-A-2015-11430 a38"
evals	Ejecutar las evals	2026-09-15T11:17:54.1867400Z       ],
evals	Ejecutar las evals	2026-09-15T11:17:54.1867644Z       "comandos_ausentes": [],
evals	Ejecutar las evals	2026-09-15T11:17:54.1867946Z       "citas_encontradas": [
evals	Ejecutar las evals	2026-09-15T11:17:54.1868248Z         "BOE-A-2015-11430 a38"
evals	Ejecutar las evals	2026-09-15T11:17:54.1868496Z       ],
evals	Ejecutar las evals	2026-09-15T11:17:54.1868739Z       "citas_ausentes": [],
evals	Ejecutar las evals	2026-09-15T11:17:54.1869033Z       "invocaciones": [
evals	Ejecutar las evals	2026-09-15T11:17:54.1869273Z         {
evals	Ejecutar las evals	2026-09-15T11:17:54.1869632Z           "orden": "boe indice BOE-A-2015-11430 --json",
evals	Ejecutar las evals	2026-09-15T11:17:54.1870001Z           "codigo": 0,
evals	Ejecutar las evals	2026-09-15T11:17:54.1870291Z           "conexiones": []
evals	Ejecutar las evals	2026-09-15T11:17:54.1870545Z         },
evals	Ejecutar las evals	2026-09-15T11:17:54.1870900Z         {
evals	Ejecutar las evals	2026-09-15T11:17:54.1871281Z           "orden": "boe articulo BOE-A-2015-11430 a38 --json",
evals	Ejecutar las evals	2026-09-15T11:17:54.1871645Z           "codigo": 0,
evals	Ejecutar las evals	2026-09-15T11:17:54.1871929Z           "conexiones": []
evals	Ejecutar las evals	2026-09-15T11:17:54.1872177Z         }
evals	Ejecutar las evals	2026-09-15T11:17:54.1872373Z       ],
evals	Ejecutar las evals	2026-09-15T11:17:54.1872622Z       "fuera_de_lo_grabado": [],
evals	Ejecutar las evals	2026-09-15T11:17:54.1873068Z       "otras_fallidas": [],
evals	Ejecutar las evals	2026-09-15T11:17:54.1873366Z       "llegadas_a_la_red": [],
evals	Ejecutar las evals	2026-09-15T11:17:54.1876642Z       "respuesta": "El art. 38 del Estatuto de los Trabajadores reconoce **30 días naturales** como mínimo de vacaciones anuales retribuidas.\n\nEspecíficamente, el artículo establece que el periodo de vacaciones será el que se pacte en convenio colectivo o contrato individual, pero en ningún caso podrá ser inferior a treinta días naturales [BOE-A-2015-11430, bloque a38].\n\nEsto significa que los 30 días constituyen el **mínimo legal garantizado**: los convenios colectivos o contratos individuales pueden establecer una duración mayor, pero no pueden reducirla por debajo de este umbral.",
evals	Ejecutar las evals	2026-09-15T11:17:54.1878801Z       "codigo_de_la_sesion": 0,
evals	Ejecutar las evals	2026-09-15T11:17:54.1879284Z       "fin_de_la_sesion": "result success",
evals	Ejecutar las evals	2026-09-15T11:17:54.1879627Z       "sesion_terminada": true,
evals	Ejecutar las evals	2026-09-15T11:17:54.1879913Z       "motivos": [],
evals	Ejecutar las evals	2026-09-15T11:17:54.1880158Z       "pasa": true
evals	Ejecutar las evals	2026-09-15T11:17:54.1880373Z     },
evals	Ejecutar las evals	2026-09-15T11:17:54.1880561Z     {
evals	Ejecutar las evals	2026-09-15T11:17:54.1880844Z       "sesion": "11-no-activa-programacion",
evals	Ejecutar las evals	2026-09-15T11:17:54.1881240Z       "eval": "11-no-activa-programacion.yaml",
evals	Ejecutar las evals	2026-09-15T11:17:54.1881584Z       "activa": false,
evals	Ejecutar las evals	2026-09-15T11:17:54.1881846Z       "activada": false,
evals	Ejecutar las evals	2026-09-15T11:17:54.1882141Z       "comandos_ejecutados": [],
evals	Ejecutar las evals	2026-09-15T11:17:54.1882449Z       "comandos_ausentes": [],
evals	Ejecutar las evals	2026-09-15T11:17:54.1882751Z       "citas_encontradas": [],
evals	Ejecutar las evals	2026-09-15T11:17:54.1883182Z       "citas_ausentes": [],
evals	Ejecutar las evals	2026-09-15T11:17:54.1883470Z       "invocaciones": [],
evals	Ejecutar las evals	2026-09-15T11:17:54.1883782Z       "fuera_de_lo_grabado": [],
evals	Ejecutar las evals	2026-09-15T11:17:54.1884076Z       "otras_fallidas": [],
evals	Ejecutar las evals	2026-09-15T11:17:54.1884362Z       "llegadas_a_la_red": [],
evals	Ejecutar las evals	2026-09-15T11:17:54.1891112Z       "respuesta": "Te muestro cómo invertir una lista enlazada en Go con dos enfoques:\n\n## 1. **Iterativo (más eficiente)**\n\n```go\ntype Node struct {\n\tVal  int\n\tNext *Node\n}\n\nfunc ReverseList(head *Node) *Node {\n\tvar prev *Node\n\tcurrent := head\n\t\n\tfor current != nil {\n\t\tnext := current.Next      // Guardar siguiente\n\t\tcurrent.Next = prev       // Invertir puntero\n\t\tprev = current            // Avanzar prev\n\t\tcurrent = next            // Avanzar current\n\t}\n\t\n\treturn prev // Nueva cabeza\n}\n```\n\n## 2. **Recursivo**\n\n```go\nfunc ReverseListRecursive(head *Node) *Node {\n\tif head == nil || head.Next == nil {\n\t\treturn head\n\t}\n\t\n\tnewHead := ReverseListRecursive(head.Next)\n\t\n\t// Invertir: hacer que Next apunte al nodo anterior\n\thead.Next.Next = head\n\thead.Next = nil\n\t\n\treturn newHead\n}\n```\n\n## Ejemplo de uso\n\n```go\n// Crear lista: 1 -> 2 -> 3 -> nil\nhead := &Node{Val: 1, Next: &Node{Val: 2, Next: &Node{Val: 3}}}\n\n// Invertir\nhead = ReverseList(head)\n\n// Recorrer resultado: 3 -> 2 -> 1 -> nil\nfor node := head; node != nil; node = node.Next {\n\tfmt.Print(node.Val, \" \")\n}\n```\n\n**Comparación:**\n- **Iterativo**: O(n) tiempo, O(1) espacio — **recomendado**\n- **Recursivo**: O(n) tiempo, O(n) espacio (pila)\n\n¿Necesitas ayuda implementando esto en tu proyecto?",
evals	Ejecutar las evals	2026-09-15T11:17:54.1895565Z       "codigo_de_la_sesion": 0,
evals	Ejecutar las evals	2026-09-15T11:17:54.1895897Z       "fin_de_la_sesion": "result success",
evals	Ejecutar las evals	2026-09-15T11:17:54.1896231Z       "sesion_terminada": true,
evals	Ejecutar las evals	2026-09-15T11:17:54.1896518Z       "motivos": [],
evals	Ejecutar las evals	2026-09-15T11:17:54.1896760Z       "pasa": true
evals	Ejecutar las evals	2026-09-15T11:17:54.1896976Z     },
evals	Ejecutar las evals	2026-09-15T11:17:54.1897166Z     {
evals	Ejecutar las evals	2026-09-15T11:17:54.1897475Z       "sesion": "12-no-activa-acuerdo-entre-amigos",
evals	Ejecutar las evals	2026-09-15T11:17:54.1897918Z       "eval": "12-no-activa-acuerdo-entre-amigos.yaml",
evals	Ejecutar las evals	2026-09-15T11:17:54.1898477Z       "activa": false,
evals	Ejecutar las evals	2026-09-15T11:17:54.1898733Z       "activada": false,
evals	Ejecutar las evals	2026-09-15T11:17:54.1899021Z       "comandos_ejecutados": [],
evals	Ejecutar las evals	2026-09-15T11:17:54.1899333Z       "comandos_ausentes": [],
evals	Ejecutar las evals	2026-09-15T11:17:54.1899636Z       "citas_encontradas": [],
evals	Ejecutar las evals	2026-09-15T11:17:54.1899922Z       "citas_ausentes": [],
evals	Ejecutar las evals	2026-09-15T11:17:54.1900201Z       "invocaciones": [],
evals	Ejecutar las evals	2026-09-15T11:17:54.1900493Z       "fuera_de_lo_grabado": [],
evals	Ejecutar las evals	2026-09-15T11:17:54.1900780Z       "otras_fallidas": [],
evals	Ejecutar las evals	2026-09-15T11:17:54.1901062Z       "llegadas_a_la_red": [],
evals	Ejecutar las evals	2026-09-15T11:17:54.1905077Z       "respuesta": "Aquí te dejo algunas opciones con un tono más cercano y amigable:\n\n**Más directa:**\n\"Nos turnamos el coche los fines de semana, cada quince días le toca a otro.\"\n\n**Más coloquial:**\n\"Nos lo pasamos el coche: un fin de semana lo tienes tú, el siguiente otro, y así.\"\n\n**Equilibrada:**\n\"Cada uno tiene el coche un fin de semana sí y otro no.\"\n\n**Aún más informal:**\n\"Nos vamos turnando el coche los fines de semana, cada dos semanas.\"\n\nLa primera opción es probablemente la mejor para un acuerdo entre amigos: mantiene claridad y formalidad mínima, pero sigue siendo comprensible y vinculante. ¿Cuál de estas te encaja mejor con el tono del resto del acuerdo?",
evals	Ejecutar las evals	2026-09-15T11:17:54.1907655Z       "codigo_de_la_sesion": 0,
evals	Ejecutar las evals	2026-09-15T11:17:54.1907995Z       "fin_de_la_sesion": "result success",
evals	Ejecutar las evals	2026-09-15T11:17:54.1908330Z       "sesion_terminada": true,
evals	Ejecutar las evals	2026-09-15T11:17:54.1908608Z       "motivos": [],
evals	Ejecutar las evals	2026-09-15T11:17:54.1908852Z       "pasa": true
evals	Ejecutar las evals	2026-09-15T11:17:54.1909063Z     }
evals	Ejecutar las evals	2026-09-15T11:17:54.1909248Z   ]
evals	Ejecutar las evals	2026-09-15T11:17:54.1909431Z }
evals	Ejecutar las evals	2026-09-15T11:17:54.1909649Z --- fin de informe.json ---
código de la quinta orden: 0
`````

## Anexo B · Sexta orden de §12.2: retirada de Python entre marcas, tal cual

`````text
evals	Retirar Python del runner	2026-09-15T11:10:58.2106240Z --- inicio de la retirada de Python ---
evals	Retirar Python del runner	2026-09-15T11:10:58.2328246Z búsqueda: find / ( -path /proc -o -path /sys ) -prune -o ( ( -type f -perm /111 ( -iname python* -o -iname pypy* ) ) -o ( -type l ( -iname python* -o -iname pypy* ) ) -o ( ( -type f -o -type l ) ( -iname libpython* -o -iname libpypy* ) ) ) -print
evals	Retirar Python del runner	2026-09-15T11:12:49.9284622Z retirado: /opt/pipx/shared/lib/python3.12/site-packages/pip/_vendor/pygments/lexers/__pycache__/python.cpython-312.pyc
evals	Retirar Python del runner	2026-09-15T11:12:49.9460534Z retirado: /opt/pipx/shared/lib/python3.12/site-packages/pip/_vendor/pygments/lexers/python.py
evals	Retirar Python del runner	2026-09-15T11:12:49.9792119Z retirado: /opt/pipx/shared
evals	Retirar Python del runner	2026-09-15T11:12:50.0726128Z retirado: /opt/pipx/venvs/yamllint
evals	Retirar Python del runner	2026-09-15T11:12:50.1285729Z retirado: /opt/pipx/venvs/ansible-core/lib/python3.12/site-packages/ansible_test/_util/target/injector/__pycache__/python.cpython-312.pyc
evals	Retirar Python del runner	2026-09-15T11:12:50.1460802Z retirado: /opt/pipx/venvs/ansible-core/lib/python3.12/site-packages/ansible_test/_util/target/injector/python.py
evals	Retirar Python del runner	2026-09-15T11:12:50.1635195Z retirado: /opt/pipx/venvs/ansible-core/lib/python3.12/site-packages/ansible_test/_internal/__pycache__/python_requirements.cpython-312.pyc
evals	Retirar Python del runner	2026-09-15T11:12:50.1809214Z retirado: /opt/pipx/venvs/ansible-core/lib/python3.12/site-packages/ansible_test/_internal/classification/__pycache__/python.cpython-312.pyc
evals	Retirar Python del runner	2026-09-15T11:12:50.1982909Z retirado: /opt/pipx/venvs/ansible-core/lib/python3.12/site-packages/ansible_test/_internal/classification/python.py
evals	Retirar Python del runner	2026-09-15T11:12:50.2157537Z retirado: /opt/pipx/venvs/ansible-core/lib/python3.12/site-packages/ansible_test/_internal/python_requirements.py
evals	Retirar Python del runner	2026-09-15T11:12:50.2349140Z retirado: /opt/pipx/venvs/ansible-core/lib/python3.12/site-packages/ansible_collections/community/okd/molecule/default/roles/openshift_adm_groups/tasks/python-ldap-not-installed.yml
evals	Retirar Python del runner	2026-09-15T11:12:50.2536982Z retirado: /opt/pipx/venvs/ansible-core/lib/python3.12/site-packages/ansible_collections/community/general/plugins/modules/__pycache__/python_requirements_info.cpython-312.pyc
evals	Retirar Python del runner	2026-09-15T11:12:50.2719065Z retirado: /opt/pipx/venvs/ansible-core/lib/python3.12/site-packages/ansible_collections/community/general/plugins/modules/python_requirements_info.py
evals	Retirar Python del runner	2026-09-15T11:12:50.2894426Z retirado: /opt/pipx/venvs/ansible-core/lib/python3.12/site-packages/ansible/_internal/ansible_collections/ansible/_protomatter/plugins/filter/__pycache__/python_literal_eval.cpython-312.pyc
evals	Retirar Python del runner	2026-09-15T11:12:50.3064073Z retirado: /opt/pipx/venvs/ansible-core/lib/python3.12/site-packages/ansible/_internal/ansible_collections/ansible/_protomatter/plugins/filter/python_literal_eval.yml
evals	Retirar Python del runner	2026-09-15T11:12:50.3234987Z retirado: /opt/pipx/venvs/ansible-core/lib/python3.12/site-packages/ansible/_internal/ansible_collections/ansible/_protomatter/plugins/filter/python_literal_eval.py
evals	Retirar Python del runner	2026-09-15T11:12:50.3405183Z retirado: /opt/pipx/venvs/ansible-core/lib/python3.12/site-packages/ansible/module_utils/facts/system/__pycache__/python.cpython-312.pyc
evals	Retirar Python del runner	2026-09-15T11:12:50.3576613Z retirado: /opt/pipx/venvs/ansible-core/lib/python3.12/site-packages/ansible/module_utils/facts/system/python.py
evals	Retirar Python del runner	2026-09-15T11:12:50.3921280Z retirado: /opt/pipx/venvs/ansible-core
evals	Retirar Python del runner	2026-09-15T11:12:51.4851831Z retirado: /opt/hostedtoolcache/PyPy/3.9.19/x64/lib/pypy3.9/site-packages/pip/_vendor/pygments/lexers/__pycache__/python.pypy39.pyc
evals	Retirar Python del runner	2026-09-15T11:12:51.5037434Z retirado: /opt/hostedtoolcache/PyPy/3.9.19/x64/lib/pypy3.9/site-packages/pip/_vendor/pygments/lexers/python.py
evals	Retirar Python del runner	2026-09-15T11:12:51.5227475Z retirado: /opt/hostedtoolcache/PyPy/3.9.19/x64/lib/pypy3.9/_cffi_ssl/_cffi_src/openssl/pypy_win32_extra.py
evals	Retirar Python del runner	2026-09-15T11:12:51.5417188Z retirado: /opt/hostedtoolcache/PyPy/3.9.19/x64/lib/pypy3.9/hpy/devel/include/hpy/forbid_python_h/Python.h
evals	Retirar Python del runner	2026-09-15T11:12:51.5602724Z retirado: /opt/hostedtoolcache/PyPy/3.9.19/x64/lib/pypy3.9/pyrepl/python_reader.py
evals	Retirar Python del runner	2026-09-15T11:12:51.5784445Z retirado: /opt/hostedtoolcache/PyPy/3.9.19/x64/lib/pypy3.9/idlelib/Icons/python.gif
evals	Retirar Python del runner	2026-09-15T11:12:51.5961091Z retirado: /opt/hostedtoolcache/PyPy/3.9.19/x64/lib/pypy3.9/test/imghdrdata/python.jpg
evals	Retirar Python del runner	2026-09-15T11:12:51.6140514Z retirado: /opt/hostedtoolcache/PyPy/3.9.19/x64/lib/pypy3.9/test/imghdrdata/python.sgi
evals	Retirar Python del runner	2026-09-15T11:12:51.6317066Z retirado: /opt/hostedtoolcache/PyPy/3.9.19/x64/lib/pypy3.9/test/imghdrdata/python.tiff
evals	Retirar Python del runner	2026-09-15T11:12:51.6498114Z retirado: /opt/hostedtoolcache/PyPy/3.9.19/x64/lib/pypy3.9/test/imghdrdata/python.bmp
evals	Retirar Python del runner	2026-09-15T11:12:51.6678535Z retirado: /opt/hostedtoolcache/PyPy/3.9.19/x64/lib/pypy3.9/test/imghdrdata/python.gif
evals	Retirar Python del runner	2026-09-15T11:12:51.6863902Z retirado: /opt/hostedtoolcache/PyPy/3.9.19/x64/lib/pypy3.9/test/imghdrdata/python.webp
evals	Retirar Python del runner	2026-09-15T11:12:51.7056540Z retirado: /opt/hostedtoolcache/PyPy/3.9.19/x64/lib/pypy3.9/test/imghdrdata/python.png
evals	Retirar Python del runner	2026-09-15T11:12:51.7243968Z retirado: /opt/hostedtoolcache/PyPy/3.9.19/x64/lib/pypy3.9/test/imghdrdata/python.ppm
evals	Retirar Python del runner	2026-09-15T11:12:51.7442455Z retirado: /opt/hostedtoolcache/PyPy/3.9.19/x64/lib/pypy3.9/test/imghdrdata/python.pgm
evals	Retirar Python del runner	2026-09-15T11:12:51.7632554Z retirado: /opt/hostedtoolcache/PyPy/3.9.19/x64/lib/pypy3.9/test/imghdrdata/python.exr
evals	Retirar Python del runner	2026-09-15T11:12:51.7816732Z retirado: /opt/hostedtoolcache/PyPy/3.9.19/x64/lib/pypy3.9/test/imghdrdata/python.xbm
evals	Retirar Python del runner	2026-09-15T11:12:51.8013758Z retirado: /opt/hostedtoolcache/PyPy/3.9.19/x64/lib/pypy3.9/test/imghdrdata/python.pbm
evals	Retirar Python del runner	2026-09-15T11:12:51.8196063Z retirado: /opt/hostedtoolcache/PyPy/3.9.19/x64/lib/pypy3.9/test/imghdrdata/python.ras
evals	Retirar Python del runner	2026-09-15T11:12:51.8394658Z retirado: /opt/hostedtoolcache/PyPy/3.9.19/x64/lib/pypy3.9/test/pythoninfo.py
evals	Retirar Python del runner	2026-09-15T11:12:51.8581271Z retirado: /opt/hostedtoolcache/PyPy/3.9.19/x64/lib/PYPY_PORTABLE_DEPS.txt
evals	Retirar Python del runner	2026-09-15T11:12:51.8778790Z retirado: /opt/hostedtoolcache/PyPy/3.9.19/x64/include/pypy3.9/Python.h
evals	Retirar Python del runner	2026-09-15T11:12:51.8961950Z retirado: /opt/hostedtoolcache/PyPy/3.9.19/x64/include/pypy3.9/pythonrun.h
evals	Retirar Python del runner	2026-09-15T11:12:51.9202484Z retirado: /opt/hostedtoolcache/PyPy/3.9.19/x64/include/pypy3.9/pypy_macros.h
evals	Retirar Python del runner	2026-09-15T11:12:51.9401933Z retirado: /opt/hostedtoolcache/PyPy/3.9.19/x64/include/pypy3.9/pypy_marshal_decl.h
evals	Retirar Python del runner	2026-09-15T11:12:51.9587121Z retirado: /opt/hostedtoolcache/PyPy/3.9.19/x64/include/pypy3.9/pypy_decl.h
evals	Retirar Python del runner	2026-09-15T11:12:51.9779444Z retirado: /opt/hostedtoolcache/PyPy/3.9.19/x64/include/pypy3.9/pypy_structmember_decl.h
evals	Retirar Python del runner	2026-09-15T11:12:51.9960707Z retirado: /opt/hostedtoolcache/PyPy/3.9.19/x64/PYPY_VERSION
evals	Retirar Python del runner	2026-09-15T11:12:52.0334674Z retirado: /opt/hostedtoolcache/PyPy/3.9.19/x64
evals	Retirar Python del runner	2026-09-15T11:12:52.3042877Z retirado: /opt/hostedtoolcache/PyPy/3.11.15/x64/lib/PYPY_PORTABLE_DEPS.txt
evals	Retirar Python del runner	2026-09-15T11:12:52.3237539Z retirado: /opt/hostedtoolcache/PyPy/3.11.15/x64/lib/pypy3.11/site-packages/pip/_vendor/pygments/lexers/__pycache__/python.pypy311.pyc
evals	Retirar Python del runner	2026-09-15T11:12:52.3435908Z retirado: /opt/hostedtoolcache/PyPy/3.11.15/x64/lib/pypy3.11/site-packages/pip/_vendor/pygments/lexers/python.py
evals	Retirar Python del runner	2026-09-15T11:12:52.3623530Z retirado: /opt/hostedtoolcache/PyPy/3.11.15/x64/lib/pypy3.11/_cffi_ssl/_cffi_src/openssl/pypy_win32_extra.py
evals	Retirar Python del runner	2026-09-15T11:12:52.3806194Z retirado: /opt/hostedtoolcache/PyPy/3.11.15/x64/lib/pypy3.11/hpy/devel/include/hpy/forbid_python_h/Python.h
evals	Retirar Python del runner	2026-09-15T11:12:52.4009000Z retirado: /opt/hostedtoolcache/PyPy/3.11.15/x64/lib/pypy3.11/idlelib/Icons/python.gif
evals	Retirar Python del runner	2026-09-15T11:12:52.4203600Z retirado: /opt/hostedtoolcache/PyPy/3.11.15/x64/lib/pypy3.11/test/imghdrdata/python.jpg
evals	Retirar Python del runner	2026-09-15T11:12:52.4389878Z retirado: /opt/hostedtoolcache/PyPy/3.11.15/x64/lib/pypy3.11/test/imghdrdata/python.sgi
evals	Retirar Python del runner	2026-09-15T11:12:52.4572161Z retirado: /opt/hostedtoolcache/PyPy/3.11.15/x64/lib/pypy3.11/test/imghdrdata/python.tiff
evals	Retirar Python del runner	2026-09-15T11:12:52.4749095Z retirado: /opt/hostedtoolcache/PyPy/3.11.15/x64/lib/pypy3.11/test/imghdrdata/python.bmp
evals	Retirar Python del runner	2026-09-15T11:12:52.4932429Z retirado: /opt/hostedtoolcache/PyPy/3.11.15/x64/lib/pypy3.11/test/imghdrdata/python.gif
evals	Retirar Python del runner	2026-09-15T11:12:52.5117219Z retirado: /opt/hostedtoolcache/PyPy/3.11.15/x64/lib/pypy3.11/test/imghdrdata/python.webp
evals	Retirar Python del runner	2026-09-15T11:12:52.5301200Z retirado: /opt/hostedtoolcache/PyPy/3.11.15/x64/lib/pypy3.11/test/imghdrdata/python.png
evals	Retirar Python del runner	2026-09-15T11:12:52.5480013Z retirado: /opt/hostedtoolcache/PyPy/3.11.15/x64/lib/pypy3.11/test/imghdrdata/python.ppm
evals	Retirar Python del runner	2026-09-15T11:12:52.5662466Z retirado: /opt/hostedtoolcache/PyPy/3.11.15/x64/lib/pypy3.11/test/imghdrdata/python.pgm
evals	Retirar Python del runner	2026-09-15T11:12:52.5847148Z retirado: /opt/hostedtoolcache/PyPy/3.11.15/x64/lib/pypy3.11/test/imghdrdata/python.exr
evals	Retirar Python del runner	2026-09-15T11:12:52.6031507Z retirado: /opt/hostedtoolcache/PyPy/3.11.15/x64/lib/pypy3.11/test/imghdrdata/python.xbm
evals	Retirar Python del runner	2026-09-15T11:12:52.6213597Z retirado: /opt/hostedtoolcache/PyPy/3.11.15/x64/lib/pypy3.11/test/imghdrdata/python.pbm
evals	Retirar Python del runner	2026-09-15T11:12:52.6414642Z retirado: /opt/hostedtoolcache/PyPy/3.11.15/x64/lib/pypy3.11/test/imghdrdata/python.ras
evals	Retirar Python del runner	2026-09-15T11:12:52.6594521Z retirado: /opt/hostedtoolcache/PyPy/3.11.15/x64/lib/pypy3.11/test/imghdrdata/python-raw.jpg
evals	Retirar Python del runner	2026-09-15T11:12:52.6778769Z retirado: /opt/hostedtoolcache/PyPy/3.11.15/x64/lib/pypy3.11/test/test_email/data/python.jpg
evals	Retirar Python del runner	2026-09-15T11:12:52.6964928Z retirado: /opt/hostedtoolcache/PyPy/3.11.15/x64/lib/pypy3.11/test/test_email/data/python.sgi
evals	Retirar Python del runner	2026-09-15T11:12:52.7146619Z retirado: /opt/hostedtoolcache/PyPy/3.11.15/x64/lib/pypy3.11/test/test_email/data/python.tiff
evals	Retirar Python del runner	2026-09-15T11:12:52.7334701Z retirado: /opt/hostedtoolcache/PyPy/3.11.15/x64/lib/pypy3.11/test/test_email/data/python.bmp
evals	Retirar Python del runner	2026-09-15T11:12:52.7541227Z retirado: /opt/hostedtoolcache/PyPy/3.11.15/x64/lib/pypy3.11/test/test_email/data/python.gif
evals	Retirar Python del runner	2026-09-15T11:12:52.7741009Z retirado: /opt/hostedtoolcache/PyPy/3.11.15/x64/lib/pypy3.11/test/test_email/data/python.webp
evals	Retirar Python del runner	2026-09-15T11:12:52.7940505Z retirado: /opt/hostedtoolcache/PyPy/3.11.15/x64/lib/pypy3.11/test/test_email/data/python.png
evals	Retirar Python del runner	2026-09-15T11:12:52.8129195Z retirado: /opt/hostedtoolcache/PyPy/3.11.15/x64/lib/pypy3.11/test/test_email/data/python.ppm
evals	Retirar Python del runner	2026-09-15T11:12:52.8318058Z retirado: /opt/hostedtoolcache/PyPy/3.11.15/x64/lib/pypy3.11/test/test_email/data/python.pgm
evals	Retirar Python del runner	2026-09-15T11:12:52.8498536Z retirado: /opt/hostedtoolcache/PyPy/3.11.15/x64/lib/pypy3.11/test/test_email/data/python.exr
evals	Retirar Python del runner	2026-09-15T11:12:52.8676371Z retirado: /opt/hostedtoolcache/PyPy/3.11.15/x64/lib/pypy3.11/test/test_email/data/python.xbm
evals	Retirar Python del runner	2026-09-15T11:12:52.8854963Z retirado: /opt/hostedtoolcache/PyPy/3.11.15/x64/lib/pypy3.11/test/test_email/data/python.pbm
evals	Retirar Python del runner	2026-09-15T11:12:52.9031469Z retirado: /opt/hostedtoolcache/PyPy/3.11.15/x64/lib/pypy3.11/test/test_email/data/python.ras
evals	Retirar Python del runner	2026-09-15T11:12:52.9223684Z retirado: /opt/hostedtoolcache/PyPy/3.11.15/x64/lib/pypy3.11/test/pythoninfo.py
evals	Retirar Python del runner	2026-09-15T11:12:52.9408476Z retirado: /opt/hostedtoolcache/PyPy/3.11.15/x64/include/pypy3.11/Python.h
evals	Retirar Python del runner	2026-09-15T11:12:52.9592532Z retirado: /opt/hostedtoolcache/PyPy/3.11.15/x64/include/pypy3.11/pythonrun.h
evals	Retirar Python del runner	2026-09-15T11:12:52.9778750Z retirado: /opt/hostedtoolcache/PyPy/3.11.15/x64/include/pypy3.11/pypy_macros.h
evals	Retirar Python del runner	2026-09-15T11:12:52.9959577Z retirado: /opt/hostedtoolcache/PyPy/3.11.15/x64/include/pypy3.11/pypy_marshal_decl.h
evals	Retirar Python del runner	2026-09-15T11:12:53.0130995Z retirado: /opt/hostedtoolcache/PyPy/3.11.15/x64/include/pypy3.11/pypy_decl.h
evals	Retirar Python del runner	2026-09-15T11:12:53.0306949Z retirado: /opt/hostedtoolcache/PyPy/3.11.15/x64/include/pypy3.11/pypy_structmember_decl.h
evals	Retirar Python del runner	2026-09-15T11:12:53.0484849Z retirado: /opt/hostedtoolcache/PyPy/3.11.15/x64/PYPY_VERSION
evals	Retirar Python del runner	2026-09-15T11:12:53.0838394Z retirado: /opt/hostedtoolcache/PyPy/3.11.15/x64
evals	Retirar Python del runner	2026-09-15T11:12:53.3771045Z retirado: /opt/hostedtoolcache/PyPy/3.10.16/x64/lib/PYPY_PORTABLE_DEPS.txt
evals	Retirar Python del runner	2026-09-15T11:12:53.3964775Z retirado: /opt/hostedtoolcache/PyPy/3.10.16/x64/lib/pypy3.10/site-packages/pip/_vendor/pygments/lexers/__pycache__/python.pypy310.pyc
evals	Retirar Python del runner	2026-09-15T11:12:53.4162169Z retirado: /opt/hostedtoolcache/PyPy/3.10.16/x64/lib/pypy3.10/site-packages/pip/_vendor/pygments/lexers/python.py
evals	Retirar Python del runner	2026-09-15T11:12:53.4364150Z retirado: /opt/hostedtoolcache/PyPy/3.10.16/x64/lib/pypy3.10/_cffi_ssl/_cffi_src/openssl/pypy_win32_extra.py
evals	Retirar Python del runner	2026-09-15T11:12:53.4556813Z retirado: /opt/hostedtoolcache/PyPy/3.10.16/x64/lib/pypy3.10/hpy/devel/include/hpy/forbid_python_h/Python.h
evals	Retirar Python del runner	2026-09-15T11:12:53.4761362Z retirado: /opt/hostedtoolcache/PyPy/3.10.16/x64/lib/pypy3.10/idlelib/Icons/python.gif
evals	Retirar Python del runner	2026-09-15T11:12:53.4951084Z retirado: /opt/hostedtoolcache/PyPy/3.10.16/x64/lib/pypy3.10/test/imghdrdata/python.jpg
evals	Retirar Python del runner	2026-09-15T11:12:53.5136965Z retirado: /opt/hostedtoolcache/PyPy/3.10.16/x64/lib/pypy3.10/test/imghdrdata/python.sgi
evals	Retirar Python del runner	2026-09-15T11:12:53.5329808Z retirado: /opt/hostedtoolcache/PyPy/3.10.16/x64/lib/pypy3.10/test/imghdrdata/python.tiff
evals	Retirar Python del runner	2026-09-15T11:12:53.5517285Z retirado: /opt/hostedtoolcache/PyPy/3.10.16/x64/lib/pypy3.10/test/imghdrdata/python.bmp
evals	Retirar Python del runner	2026-09-15T11:12:53.5707232Z retirado: /opt/hostedtoolcache/PyPy/3.10.16/x64/lib/pypy3.10/test/imghdrdata/python.gif
evals	Retirar Python del runner	2026-09-15T11:12:53.5896610Z retirado: /opt/hostedtoolcache/PyPy/3.10.16/x64/lib/pypy3.10/test/imghdrdata/python.webp
evals	Retirar Python del runner	2026-09-15T11:12:53.6082105Z retirado: /opt/hostedtoolcache/PyPy/3.10.16/x64/lib/pypy3.10/test/imghdrdata/python.png
evals	Retirar Python del runner	2026-09-15T11:12:53.6270083Z retirado: /opt/hostedtoolcache/PyPy/3.10.16/x64/lib/pypy3.10/test/imghdrdata/python.ppm
evals	Retirar Python del runner	2026-09-15T11:12:53.6462728Z retirado: /opt/hostedtoolcache/PyPy/3.10.16/x64/lib/pypy3.10/test/imghdrdata/python.pgm
evals	Retirar Python del runner	2026-09-15T11:12:53.6652243Z retirado: /opt/hostedtoolcache/PyPy/3.10.16/x64/lib/pypy3.10/test/imghdrdata/python.exr
evals	Retirar Python del runner	2026-09-15T11:12:53.6840406Z retirado: /opt/hostedtoolcache/PyPy/3.10.16/x64/lib/pypy3.10/test/imghdrdata/python.xbm
evals	Retirar Python del runner	2026-09-15T11:12:53.7040666Z retirado: /opt/hostedtoolcache/PyPy/3.10.16/x64/lib/pypy3.10/test/imghdrdata/python.pbm
evals	Retirar Python del runner	2026-09-15T11:12:53.7235864Z retirado: /opt/hostedtoolcache/PyPy/3.10.16/x64/lib/pypy3.10/test/imghdrdata/python.ras
evals	Retirar Python del runner	2026-09-15T11:12:53.7438201Z retirado: /opt/hostedtoolcache/PyPy/3.10.16/x64/lib/pypy3.10/test/pythoninfo.py
evals	Retirar Python del runner	2026-09-15T11:12:53.7631051Z retirado: /opt/hostedtoolcache/PyPy/3.10.16/x64/include/pypy3.10/Python.h
evals	Retirar Python del runner	2026-09-15T11:12:53.7823824Z retirado: /opt/hostedtoolcache/PyPy/3.10.16/x64/include/pypy3.10/pythonrun.h
evals	Retirar Python del runner	2026-09-15T11:12:53.8015247Z retirado: /opt/hostedtoolcache/PyPy/3.10.16/x64/include/pypy3.10/pypy_macros.h
evals	Retirar Python del runner	2026-09-15T11:12:53.8208600Z retirado: /opt/hostedtoolcache/PyPy/3.10.16/x64/include/pypy3.10/pypy_marshal_decl.h
evals	Retirar Python del runner	2026-09-15T11:12:53.8401031Z retirado: /opt/hostedtoolcache/PyPy/3.10.16/x64/include/pypy3.10/pypy_decl.h
evals	Retirar Python del runner	2026-09-15T11:12:53.8609890Z retirado: /opt/hostedtoolcache/PyPy/3.10.16/x64/include/pypy3.10/pypy_structmember_decl.h
evals	Retirar Python del runner	2026-09-15T11:12:53.8805665Z retirado: /opt/hostedtoolcache/PyPy/3.10.16/x64/PYPY_VERSION
evals	Retirar Python del runner	2026-09-15T11:12:53.9225430Z retirado: /opt/hostedtoolcache/PyPy/3.10.16/x64
evals	Retirar Python del runner	2026-09-15T11:12:54.1983950Z retirado: /opt/hostedtoolcache/CodeQL/2.26.4/x64/codeql/Open-Source-Notices/python3.txt
evals	Retirar Python del runner	2026-09-15T11:12:54.2169938Z retirado: /opt/hostedtoolcache/CodeQL/2.26.4/x64/codeql/qlpacks/codeql/python-queries/1.8.9/experimental/semmle/python/libraries/PythonJose.qll
evals	Retirar Python del runner	2026-09-15T11:12:54.2356193Z retirado: /opt/hostedtoolcache/CodeQL/2.26.4/x64/codeql/qlpacks/codeql/python-queries/1.8.9/experimental/semmle/python/libraries/Python_JWT.qll
evals	Retirar Python del runner	2026-09-15T11:12:54.2555162Z retirado: /opt/hostedtoolcache/CodeQL/2.26.4/x64/codeql/qlpacks/codeql/python-queries/1.8.9/.codeql/libraries/codeql/python-all/7.2.4/python.qll
evals	Retirar Python del runner	2026-09-15T11:12:54.2746829Z retirado: /opt/hostedtoolcache/CodeQL/2.26.4/x64/codeql/qlpacks/codeql/python-queries/1.8.9/codeql-suites/python-code-quality-extended.qls
evals	Retirar Python del runner	2026-09-15T11:12:54.2925450Z retirado: /opt/hostedtoolcache/CodeQL/2.26.4/x64/codeql/qlpacks/codeql/python-queries/1.8.9/codeql-suites/python-security-experimental.qls
evals	Retirar Python del runner	2026-09-15T11:12:54.3119161Z retirado: /opt/hostedtoolcache/CodeQL/2.26.4/x64/codeql/qlpacks/codeql/python-queries/1.8.9/codeql-suites/python-code-scanning.qls
evals	Retirar Python del runner	2026-09-15T11:12:54.3303482Z retirado: /opt/hostedtoolcache/CodeQL/2.26.4/x64/codeql/qlpacks/codeql/python-queries/1.8.9/codeql-suites/python-lgtm.qls
evals	Retirar Python del runner	2026-09-15T11:12:54.3495353Z retirado: /opt/hostedtoolcache/CodeQL/2.26.4/x64/codeql/qlpacks/codeql/python-queries/1.8.9/codeql-suites/python-security-extended.qls
evals	Retirar Python del runner	2026-09-15T11:12:54.3684456Z retirado: /opt/hostedtoolcache/CodeQL/2.26.4/x64/codeql/qlpacks/codeql/python-queries/1.8.9/codeql-suites/python-lgtm-full.qls
evals	Retirar Python del runner	2026-09-15T11:12:54.3869534Z retirado: /opt/hostedtoolcache/CodeQL/2.26.4/x64/codeql/qlpacks/codeql/python-queries/1.8.9/codeql-suites/python-security-and-quality.qls
evals	Retirar Python del runner	2026-09-15T11:12:54.4071398Z retirado: /opt/hostedtoolcache/CodeQL/2.26.4/x64/codeql/qlpacks/codeql/python-queries/1.8.9/codeql-suites/python-code-quality.qls
evals	Retirar Python del runner	2026-09-15T11:12:54.4294661Z retirado: /opt/hostedtoolcache/CodeQL/2.26.4/x64/codeql/qlpacks/codeql/python-examples/0.0.0/.codeql/libraries/codeql/python-all/7.2.4/python.qll
evals	Retirar Python del runner	2026-09-15T11:12:54.4508862Z retirado: /opt/hostedtoolcache/CodeQL/2.26.4/x64/codeql/qlpacks/codeql/actions-all/0.6.0/ext/generated/composite-actions/python_mypy.model.yml
evals	Retirar Python del runner	2026-09-15T11:12:54.4728745Z retirado: /opt/hostedtoolcache/CodeQL/2.26.4/x64/codeql/qlpacks/codeql/actions-all/0.6.0/ext/generated/composite-actions/python-poetry_poetry.model.yml
evals	Retirar Python del runner	2026-09-15T11:12:54.4949933Z retirado: /opt/hostedtoolcache/CodeQL/2.26.4/x64/codeql/qlpacks/codeql/actions-all/0.6.0/ext/generated/reusable-workflows/python_cpython.model.yml
evals	Retirar Python del runner	2026-09-15T11:12:54.5173695Z retirado: /opt/hostedtoolcache/CodeQL/2.26.4/x64/codeql/qlpacks/codeql/python-all/7.2.4/python.qll
evals	Retirar Python del runner	2026-09-15T11:12:54.5385316Z retirado: /opt/hostedtoolcache/CodeQL/2.26.4/x64/codeql/qlpacks/codeql/actions-queries/0.6.34/.codeql/libraries/codeql/actions-all/0.6.0/ext/generated/composite-actions/python_mypy.model.yml
evals	Retirar Python del runner	2026-09-15T11:12:54.5605713Z retirado: /opt/hostedtoolcache/CodeQL/2.26.4/x64/codeql/qlpacks/codeql/actions-queries/0.6.34/.codeql/libraries/codeql/actions-all/0.6.0/ext/generated/composite-actions/python-poetry_poetry.model.yml
evals	Retirar Python del runner	2026-09-15T11:12:54.5817538Z retirado: /opt/hostedtoolcache/CodeQL/2.26.4/x64/codeql/qlpacks/codeql/actions-queries/0.6.34/.codeql/libraries/codeql/actions-all/0.6.0/ext/generated/reusable-workflows/python_cpython.model.yml
evals	Retirar Python del runner	2026-09-15T11:12:54.6032380Z retirado: /opt/hostedtoolcache/CodeQL/2.26.4/x64/codeql/python/tools/python3src.zip
evals	Retirar Python del runner	2026-09-15T11:12:54.6253637Z retirado: /opt/hostedtoolcache/CodeQL/2.26.4/x64/codeql/python/tools/python_setup.cmd
evals	Retirar Python del runner	2026-09-15T11:12:54.6470750Z retirado: /opt/hostedtoolcache/CodeQL/2.26.4/x64/codeql/python/tools/python_setup.sh
evals	Retirar Python del runner	2026-09-15T11:12:54.6683500Z retirado: /opt/hostedtoolcache/CodeQL/2.26.4/x64/codeql/python/tools/python_tracer.py
evals	Retirar Python del runner	2026-09-15T11:12:54.6902444Z retirado: /opt/hostedtoolcache/Python/3.14.7/x64/lib/libpython3.so
evals	Retirar Python del runner	2026-09-15T11:12:54.7089680Z retirado: /opt/hostedtoolcache/Python/3.14.7/x64/lib/pkgconfig/python-3.14-embed.pc
evals	Retirar Python del runner	2026-09-15T11:12:54.7288182Z retirado: /opt/hostedtoolcache/Python/3.14.7/x64/lib/pkgconfig/python3.pc
evals	Retirar Python del runner	2026-09-15T11:12:54.7548847Z retirado: /opt/hostedtoolcache/Python/3.14.7/x64/lib/pkgconfig/python3-embed.pc
evals	Retirar Python del runner	2026-09-15T11:12:54.7720631Z retirado: /opt/hostedtoolcache/Python/3.14.7/x64/lib/pkgconfig/python-3.14.pc
evals	Retirar Python del runner	2026-09-15T11:12:54.7903779Z retirado: /opt/hostedtoolcache/Python/3.14.7/x64/lib/python3.14/site-packages/pip/_vendor/pygments/lexers/__pycache__/python.cpython-314.pyc
evals	Retirar Python del runner	2026-09-15T11:12:54.8083915Z retirado: /opt/hostedtoolcache/Python/3.14.7/x64/lib/python3.14/site-packages/pip/_vendor/pygments/lexers/python.py
evals	Retirar Python del runner	2026-09-15T11:12:54.8261610Z retirado: /opt/hostedtoolcache/Python/3.14.7/x64/lib/python3.14/config-3.14-x86_64-linux-gnu/libpython3.14.a
evals	Retirar Python del runner	2026-09-15T11:12:54.8438040Z retirado: /opt/hostedtoolcache/Python/3.14.7/x64/lib/python3.14/config-3.14-x86_64-linux-gnu/python-config.py
evals	Retirar Python del runner	2026-09-15T11:12:54.8612736Z retirado: /opt/hostedtoolcache/Python/3.14.7/x64/lib/python3.14/config-3.14-x86_64-linux-gnu/python.o
evals	Retirar Python del runner	2026-09-15T11:12:54.8790427Z retirado: /opt/hostedtoolcache/Python/3.14.7/x64/lib/python3.14/idlelib/Icons/python.gif
evals	Retirar Python del runner	2026-09-15T11:12:54.8965352Z retirado: /opt/hostedtoolcache/Python/3.14.7/x64/lib/python3.14/test/__pycache__/pythoninfo.cpython-314.pyc
evals	Retirar Python del runner	2026-09-15T11:12:54.9154552Z retirado: /opt/hostedtoolcache/Python/3.14.7/x64/lib/python3.14/test/__pycache__/pythoninfo.cpython-314.opt-2.pyc
evals	Retirar Python del runner	2026-09-15T11:12:54.9373643Z retirado: /opt/hostedtoolcache/Python/3.14.7/x64/lib/python3.14/test/__pycache__/pythoninfo.cpython-314.opt-1.pyc
evals	Retirar Python del runner	2026-09-15T11:12:54.9548518Z retirado: /opt/hostedtoolcache/Python/3.14.7/x64/lib/python3.14/test/tkinterdata/python.gif
evals	Retirar Python del runner	2026-09-15T11:12:54.9720747Z retirado: /opt/hostedtoolcache/Python/3.14.7/x64/lib/python3.14/test/tkinterdata/python.png
evals	Retirar Python del runner	2026-09-15T11:12:54.9898502Z retirado: /opt/hostedtoolcache/Python/3.14.7/x64/lib/python3.14/test/tkinterdata/python.ppm
evals	Retirar Python del runner	2026-09-15T11:12:55.0068533Z retirado: /opt/hostedtoolcache/Python/3.14.7/x64/lib/python3.14/test/tkinterdata/python.pgm
evals	Retirar Python del runner	2026-09-15T11:12:55.0241942Z retirado: /opt/hostedtoolcache/Python/3.14.7/x64/lib/python3.14/test/tkinterdata/python.xbm
evals	Retirar Python del runner	2026-09-15T11:12:55.0415488Z retirado: /opt/hostedtoolcache/Python/3.14.7/x64/lib/python3.14/test/test_email/data/python.jpg
evals	Retirar Python del runner	2026-09-15T11:12:55.0593478Z retirado: /opt/hostedtoolcache/Python/3.14.7/x64/lib/python3.14/test/test_email/data/python.sgi
evals	Retirar Python del runner	2026-09-15T11:12:55.0776541Z retirado: /opt/hostedtoolcache/Python/3.14.7/x64/lib/python3.14/test/test_email/data/python.tiff
evals	Retirar Python del runner	2026-09-15T11:12:55.0950176Z retirado: /opt/hostedtoolcache/Python/3.14.7/x64/lib/python3.14/test/test_email/data/python.bmp
evals	Retirar Python del runner	2026-09-15T11:12:55.1120535Z retirado: /opt/hostedtoolcache/Python/3.14.7/x64/lib/python3.14/test/test_email/data/python.gif
evals	Retirar Python del runner	2026-09-15T11:12:55.1292542Z retirado: /opt/hostedtoolcache/Python/3.14.7/x64/lib/python3.14/test/test_email/data/python.webp
evals	Retirar Python del runner	2026-09-15T11:12:55.1465802Z retirado: /opt/hostedtoolcache/Python/3.14.7/x64/lib/python3.14/test/test_email/data/python.png
evals	Retirar Python del runner	2026-09-15T11:12:55.1642617Z retirado: /opt/hostedtoolcache/Python/3.14.7/x64/lib/python3.14/test/test_email/data/python.ppm
evals	Retirar Python del runner	2026-09-15T11:12:55.1826815Z retirado: /opt/hostedtoolcache/Python/3.14.7/x64/lib/python3.14/test/test_email/data/python.pgm
evals	Retirar Python del runner	2026-09-15T11:12:55.2017155Z retirado: /opt/hostedtoolcache/Python/3.14.7/x64/lib/python3.14/test/test_email/data/python.exr
evals	Retirar Python del runner	2026-09-15T11:12:55.2194306Z retirado: /opt/hostedtoolcache/Python/3.14.7/x64/lib/python3.14/test/test_email/data/python.xbm
evals	Retirar Python del runner	2026-09-15T11:12:55.2370947Z retirado: /opt/hostedtoolcache/Python/3.14.7/x64/lib/python3.14/test/test_email/data/python.pbm
evals	Retirar Python del runner	2026-09-15T11:12:55.2547381Z retirado: /opt/hostedtoolcache/Python/3.14.7/x64/lib/python3.14/test/test_email/data/python.ras
evals	Retirar Python del runner	2026-09-15T11:12:55.2724460Z retirado: /opt/hostedtoolcache/Python/3.14.7/x64/lib/python3.14/test/pythoninfo.py
evals	Retirar Python del runner	2026-09-15T11:12:55.2904588Z retirado: /opt/hostedtoolcache/Python/3.14.7/x64/lib/libpython3.14.so
evals	Retirar Python del runner	2026-09-15T11:12:55.3081485Z retirado: /opt/hostedtoolcache/Python/3.14.7/x64/lib/libpython3.14.so.1.0
evals	Retirar Python del runner	2026-09-15T11:12:55.3260586Z retirado: /opt/hostedtoolcache/Python/3.14.7/x64/include/python3.14/cpython/pythonrun.h
evals	Retirar Python del runner	2026-09-15T11:12:55.3450198Z retirado: /opt/hostedtoolcache/Python/3.14.7/x64/include/python3.14/Python.h
evals	Retirar Python del runner	2026-09-15T11:12:55.3632854Z retirado: /opt/hostedtoolcache/Python/3.14.7/x64/include/python3.14/pythonrun.h
evals	Retirar Python del runner	2026-09-15T11:12:55.3815097Z retirado: /opt/hostedtoolcache/Python/3.14.7/x64/share/man/man1/python3.1
evals	Retirar Python del runner	2026-09-15T11:12:55.4004339Z retirado: /opt/hostedtoolcache/Python/3.14.7/x64/share/man/man1/python3.14.1
evals	Retirar Python del runner	2026-09-15T11:12:55.4358635Z retirado: /opt/hostedtoolcache/Python/3.14.7/x64
evals	Retirar Python del runner	2026-09-15T11:12:55.7800432Z retirado: /opt/hostedtoolcache/Python/3.13.15/x64/lib/libpython3.so
evals	Retirar Python del runner	2026-09-15T11:12:55.7995611Z retirado: /opt/hostedtoolcache/Python/3.13.15/x64/lib/pkgconfig/python3.pc
evals	Retirar Python del runner	2026-09-15T11:12:55.8183416Z retirado: /opt/hostedtoolcache/Python/3.13.15/x64/lib/pkgconfig/python-3.13.pc
evals	Retirar Python del runner	2026-09-15T11:12:55.8371003Z retirado: /opt/hostedtoolcache/Python/3.13.15/x64/lib/pkgconfig/python3-embed.pc
evals	Retirar Python del runner	2026-09-15T11:12:55.8550157Z retirado: /opt/hostedtoolcache/Python/3.13.15/x64/lib/pkgconfig/python-3.13-embed.pc
evals	Retirar Python del runner	2026-09-15T11:12:55.8726016Z retirado: /opt/hostedtoolcache/Python/3.13.15/x64/lib/python3.13/site-packages/pip/_vendor/pygments/lexers/__pycache__/python.cpython-313.pyc
evals	Retirar Python del runner	2026-09-15T11:12:55.8901960Z retirado: /opt/hostedtoolcache/Python/3.13.15/x64/lib/python3.13/site-packages/pip/_vendor/pygments/lexers/python.py
evals	Retirar Python del runner	2026-09-15T11:12:55.9085497Z retirado: /opt/hostedtoolcache/Python/3.13.15/x64/lib/python3.13/idlelib/Icons/python.gif
evals	Retirar Python del runner	2026-09-15T11:12:55.9285315Z retirado: /opt/hostedtoolcache/Python/3.13.15/x64/lib/python3.13/test/__pycache__/pythoninfo.cpython-313.opt-2.pyc
evals	Retirar Python del runner	2026-09-15T11:12:55.9465660Z retirado: /opt/hostedtoolcache/Python/3.13.15/x64/lib/python3.13/test/__pycache__/pythoninfo.cpython-313.pyc
evals	Retirar Python del runner	2026-09-15T11:12:55.9641356Z retirado: /opt/hostedtoolcache/Python/3.13.15/x64/lib/python3.13/test/__pycache__/pythoninfo.cpython-313.opt-1.pyc
evals	Retirar Python del runner	2026-09-15T11:12:55.9828130Z retirado: /opt/hostedtoolcache/Python/3.13.15/x64/lib/python3.13/test/tkinterdata/python.gif
evals	Retirar Python del runner	2026-09-15T11:12:56.0007542Z retirado: /opt/hostedtoolcache/Python/3.13.15/x64/lib/python3.13/test/tkinterdata/python.png
evals	Retirar Python del runner	2026-09-15T11:12:56.0182734Z retirado: /opt/hostedtoolcache/Python/3.13.15/x64/lib/python3.13/test/tkinterdata/python.ppm
evals	Retirar Python del runner	2026-09-15T11:12:56.0361774Z retirado: /opt/hostedtoolcache/Python/3.13.15/x64/lib/python3.13/test/tkinterdata/python.pgm
evals	Retirar Python del runner	2026-09-15T11:12:56.0548455Z retirado: /opt/hostedtoolcache/Python/3.13.15/x64/lib/python3.13/test/tkinterdata/python.xbm
evals	Retirar Python del runner	2026-09-15T11:12:56.0735014Z retirado: /opt/hostedtoolcache/Python/3.13.15/x64/lib/python3.13/test/test_email/data/python.jpg
evals	Retirar Python del runner	2026-09-15T11:12:56.0935234Z retirado: /opt/hostedtoolcache/Python/3.13.15/x64/lib/python3.13/test/test_email/data/python.sgi
evals	Retirar Python del runner	2026-09-15T11:12:56.1124972Z retirado: /opt/hostedtoolcache/Python/3.13.15/x64/lib/python3.13/test/test_email/data/python.tiff
evals	Retirar Python del runner	2026-09-15T11:12:56.1311539Z retirado: /opt/hostedtoolcache/Python/3.13.15/x64/lib/python3.13/test/test_email/data/python.bmp
evals	Retirar Python del runner	2026-09-15T11:12:56.1491993Z retirado: /opt/hostedtoolcache/Python/3.13.15/x64/lib/python3.13/test/test_email/data/python.gif
evals	Retirar Python del runner	2026-09-15T11:12:56.1676314Z retirado: /opt/hostedtoolcache/Python/3.13.15/x64/lib/python3.13/test/test_email/data/python.webp
evals	Retirar Python del runner	2026-09-15T11:12:56.1861547Z retirado: /opt/hostedtoolcache/Python/3.13.15/x64/lib/python3.13/test/test_email/data/python.png
evals	Retirar Python del runner	2026-09-15T11:12:56.2045772Z retirado: /opt/hostedtoolcache/Python/3.13.15/x64/lib/python3.13/test/test_email/data/python.ppm
evals	Retirar Python del runner	2026-09-15T11:12:56.2224107Z retirado: /opt/hostedtoolcache/Python/3.13.15/x64/lib/python3.13/test/test_email/data/python.pgm
evals	Retirar Python del runner	2026-09-15T11:12:56.2417436Z retirado: /opt/hostedtoolcache/Python/3.13.15/x64/lib/python3.13/test/test_email/data/python.exr
evals	Retirar Python del runner	2026-09-15T11:12:56.2600237Z retirado: /opt/hostedtoolcache/Python/3.13.15/x64/lib/python3.13/test/test_email/data/python.xbm
evals	Retirar Python del runner	2026-09-15T11:12:56.2780646Z retirado: /opt/hostedtoolcache/Python/3.13.15/x64/lib/python3.13/test/test_email/data/python.pbm
evals	Retirar Python del runner	2026-09-15T11:12:56.2959586Z retirado: /opt/hostedtoolcache/Python/3.13.15/x64/lib/python3.13/test/test_email/data/python.ras
evals	Retirar Python del runner	2026-09-15T11:12:56.3144997Z retirado: /opt/hostedtoolcache/Python/3.13.15/x64/lib/python3.13/test/pythoninfo.py
evals	Retirar Python del runner	2026-09-15T11:12:56.3337048Z retirado: /opt/hostedtoolcache/Python/3.13.15/x64/lib/python3.13/config-3.13-x86_64-linux-gnu/libpython3.13.a
evals	Retirar Python del runner	2026-09-15T11:12:56.3523641Z retirado: /opt/hostedtoolcache/Python/3.13.15/x64/lib/python3.13/config-3.13-x86_64-linux-gnu/python-config.py
evals	Retirar Python del runner	2026-09-15T11:12:56.3717898Z retirado: /opt/hostedtoolcache/Python/3.13.15/x64/lib/python3.13/config-3.13-x86_64-linux-gnu/python.o
evals	Retirar Python del runner	2026-09-15T11:12:56.3909254Z retirado: /opt/hostedtoolcache/Python/3.13.15/x64/lib/libpython3.13.so
evals	Retirar Python del runner	2026-09-15T11:12:56.4103483Z retirado: /opt/hostedtoolcache/Python/3.13.15/x64/lib/libpython3.13.so.1.0
evals	Retirar Python del runner	2026-09-15T11:12:56.4291116Z retirado: /opt/hostedtoolcache/Python/3.13.15/x64/include/python3.13/cpython/pythonrun.h
evals	Retirar Python del runner	2026-09-15T11:12:56.4476569Z retirado: /opt/hostedtoolcache/Python/3.13.15/x64/include/python3.13/Python.h
evals	Retirar Python del runner	2026-09-15T11:12:56.4657456Z retirado: /opt/hostedtoolcache/Python/3.13.15/x64/include/python3.13/pythonrun.h
evals	Retirar Python del runner	2026-09-15T11:12:56.4835965Z retirado: /opt/hostedtoolcache/Python/3.13.15/x64/share/man/man1/python3.1
evals	Retirar Python del runner	2026-09-15T11:12:56.5011535Z retirado: /opt/hostedtoolcache/Python/3.13.15/x64/share/man/man1/python3.13.1
evals	Retirar Python del runner	2026-09-15T11:12:56.5360712Z retirado: /opt/hostedtoolcache/Python/3.13.15/x64
evals	Retirar Python del runner	2026-09-15T11:12:57.0544305Z retirado: /opt/hostedtoolcache/Python/3.10.21/x64/lib/libpython3.so
evals	Retirar Python del runner	2026-09-15T11:12:57.0820717Z retirado: /opt/hostedtoolcache/Python/3.10.21/x64/lib/python3.10/site-packages/pip/_vendor/pygments/lexers/__pycache__/python.cpython-310.pyc
evals	Retirar Python del runner	2026-09-15T11:12:57.1109376Z retirado: /opt/hostedtoolcache/Python/3.10.21/x64/lib/python3.10/site-packages/pip/_vendor/pygments/lexers/python.py
evals	Retirar Python del runner	2026-09-15T11:12:57.1419784Z retirado: /opt/hostedtoolcache/Python/3.10.21/x64/lib/python3.10/config-3.10-x86_64-linux-gnu/libpython3.10.a
evals	Retirar Python del runner	2026-09-15T11:12:57.1718214Z retirado: /opt/hostedtoolcache/Python/3.10.21/x64/lib/python3.10/config-3.10-x86_64-linux-gnu/python-config.py
evals	Retirar Python del runner	2026-09-15T11:12:57.2023396Z retirado: /opt/hostedtoolcache/Python/3.10.21/x64/lib/python3.10/config-3.10-x86_64-linux-gnu/python.o
evals	Retirar Python del runner	2026-09-15T11:12:57.2300831Z retirado: /opt/hostedtoolcache/Python/3.10.21/x64/lib/python3.10/idlelib/Icons/python.gif
evals	Retirar Python del runner	2026-09-15T11:12:57.2611342Z retirado: /opt/hostedtoolcache/Python/3.10.21/x64/lib/python3.10/test/__pycache__/pythoninfo.cpython-310.pyc
evals	Retirar Python del runner	2026-09-15T11:12:57.2907598Z retirado: /opt/hostedtoolcache/Python/3.10.21/x64/lib/python3.10/test/__pycache__/pythoninfo.cpython-310.opt-2.pyc
evals	Retirar Python del runner	2026-09-15T11:12:57.3204514Z retirado: /opt/hostedtoolcache/Python/3.10.21/x64/lib/python3.10/test/__pycache__/pythoninfo.cpython-310.opt-1.pyc
evals	Retirar Python del runner	2026-09-15T11:12:57.3543313Z retirado: /opt/hostedtoolcache/Python/3.10.21/x64/lib/python3.10/test/imghdrdata/python.jpg
evals	Retirar Python del runner	2026-09-15T11:12:57.3809670Z retirado: /opt/hostedtoolcache/Python/3.10.21/x64/lib/python3.10/test/imghdrdata/python.sgi
evals	Retirar Python del runner	2026-09-15T11:12:57.3994435Z retirado: /opt/hostedtoolcache/Python/3.10.21/x64/lib/python3.10/test/imghdrdata/python.tiff
evals	Retirar Python del runner	2026-09-15T11:12:57.4172843Z retirado: /opt/hostedtoolcache/Python/3.10.21/x64/lib/python3.10/test/imghdrdata/python.bmp
evals	Retirar Python del runner	2026-09-15T11:12:57.4364577Z retirado: /opt/hostedtoolcache/Python/3.10.21/x64/lib/python3.10/test/imghdrdata/python.gif
evals	Retirar Python del runner	2026-09-15T11:12:57.4543424Z retirado: /opt/hostedtoolcache/Python/3.10.21/x64/lib/python3.10/test/imghdrdata/python.webp
evals	Retirar Python del runner	2026-09-15T11:12:57.4720281Z retirado: /opt/hostedtoolcache/Python/3.10.21/x64/lib/python3.10/test/imghdrdata/python.png
evals	Retirar Python del runner	2026-09-15T11:12:57.4897393Z retirado: /opt/hostedtoolcache/Python/3.10.21/x64/lib/python3.10/test/imghdrdata/python.ppm
evals	Retirar Python del runner	2026-09-15T11:12:57.5082412Z retirado: /opt/hostedtoolcache/Python/3.10.21/x64/lib/python3.10/test/imghdrdata/python.pgm
evals	Retirar Python del runner	2026-09-15T11:12:57.5262670Z retirado: /opt/hostedtoolcache/Python/3.10.21/x64/lib/python3.10/test/imghdrdata/python.exr
evals	Retirar Python del runner	2026-09-15T11:12:57.5442312Z retirado: /opt/hostedtoolcache/Python/3.10.21/x64/lib/python3.10/test/imghdrdata/python.xbm
evals	Retirar Python del runner	2026-09-15T11:12:57.5626172Z retirado: /opt/hostedtoolcache/Python/3.10.21/x64/lib/python3.10/test/imghdrdata/python.pbm
evals	Retirar Python del runner	2026-09-15T11:12:57.5800270Z retirado: /opt/hostedtoolcache/Python/3.10.21/x64/lib/python3.10/test/imghdrdata/python.ras
evals	Retirar Python del runner	2026-09-15T11:12:57.5972763Z retirado: /opt/hostedtoolcache/Python/3.10.21/x64/lib/python3.10/test/pythoninfo.py
evals	Retirar Python del runner	2026-09-15T11:12:57.6146577Z retirado: /opt/hostedtoolcache/Python/3.10.21/x64/lib/pkgconfig/python3.pc
evals	Retirar Python del runner	2026-09-15T11:12:57.6320842Z retirado: /opt/hostedtoolcache/Python/3.10.21/x64/lib/pkgconfig/python3-embed.pc
evals	Retirar Python del runner	2026-09-15T11:12:57.6507651Z retirado: /opt/hostedtoolcache/Python/3.10.21/x64/lib/pkgconfig/python-3.10.pc
evals	Retirar Python del runner	2026-09-15T11:12:57.6682816Z retirado: /opt/hostedtoolcache/Python/3.10.21/x64/lib/pkgconfig/python-3.10-embed.pc
evals	Retirar Python del runner	2026-09-15T11:12:57.6862521Z retirado: /opt/hostedtoolcache/Python/3.10.21/x64/lib/libpython3.10.so
evals	Retirar Python del runner	2026-09-15T11:12:57.7039831Z retirado: /opt/hostedtoolcache/Python/3.10.21/x64/lib/libpython3.10.so.1.0
evals	Retirar Python del runner	2026-09-15T11:12:57.7211956Z retirado: /opt/hostedtoolcache/Python/3.10.21/x64/include/python3.10/cpython/pythonrun.h
evals	Retirar Python del runner	2026-09-15T11:12:57.7385491Z retirado: /opt/hostedtoolcache/Python/3.10.21/x64/include/python3.10/Python.h
evals	Retirar Python del runner	2026-09-15T11:12:57.7567703Z retirado: /opt/hostedtoolcache/Python/3.10.21/x64/include/python3.10/pythonrun.h
evals	Retirar Python del runner	2026-09-15T11:12:57.7739902Z retirado: /opt/hostedtoolcache/Python/3.10.21/x64/share/man/man1/python3.10.1
evals	Retirar Python del runner	2026-09-15T11:12:57.7986687Z retirado: /opt/hostedtoolcache/Python/3.10.21/x64/share/man/man1/python3.1
evals	Retirar Python del runner	2026-09-15T11:12:57.8321869Z retirado: /opt/hostedtoolcache/Python/3.10.21/x64
evals	Retirar Python del runner	2026-09-15T11:12:58.1790466Z retirado: /opt/hostedtoolcache/Python/3.12.14/x64/lib/libpython3.so
evals	Retirar Python del runner	2026-09-15T11:12:58.1962446Z retirado: /opt/hostedtoolcache/Python/3.12.14/x64/lib/pkgconfig/python-3.12.pc
evals	Retirar Python del runner	2026-09-15T11:12:58.2213860Z retirado: /opt/hostedtoolcache/Python/3.12.14/x64/lib/pkgconfig/python3.pc
evals	Retirar Python del runner	2026-09-15T11:12:58.2398479Z retirado: /opt/hostedtoolcache/Python/3.12.14/x64/lib/pkgconfig/python-3.12-embed.pc
evals	Retirar Python del runner	2026-09-15T11:12:58.2658025Z retirado: /opt/hostedtoolcache/Python/3.12.14/x64/lib/pkgconfig/python3-embed.pc
evals	Retirar Python del runner	2026-09-15T11:12:58.2837980Z retirado: /opt/hostedtoolcache/Python/3.12.14/x64/lib/libpython3.12.so
evals	Retirar Python del runner	2026-09-15T11:12:58.3008885Z retirado: /opt/hostedtoolcache/Python/3.12.14/x64/lib/python3.12/site-packages/pip/_vendor/pygments/lexers/__pycache__/python.cpython-312.pyc
evals	Retirar Python del runner	2026-09-15T11:12:58.3179475Z retirado: /opt/hostedtoolcache/Python/3.12.14/x64/lib/python3.12/site-packages/pip/_vendor/pygments/lexers/python.py
evals	Retirar Python del runner	2026-09-15T11:12:58.3353684Z retirado: /opt/hostedtoolcache/Python/3.12.14/x64/lib/python3.12/idlelib/Icons/python.gif
evals	Retirar Python del runner	2026-09-15T11:12:58.3539027Z retirado: /opt/hostedtoolcache/Python/3.12.14/x64/lib/python3.12/test/__pycache__/pythoninfo.cpython-312.opt-2.pyc
evals	Retirar Python del runner	2026-09-15T11:12:58.3718274Z retirado: /opt/hostedtoolcache/Python/3.12.14/x64/lib/python3.12/test/__pycache__/pythoninfo.cpython-312.opt-1.pyc
evals	Retirar Python del runner	2026-09-15T11:12:58.3895415Z retirado: /opt/hostedtoolcache/Python/3.12.14/x64/lib/python3.12/test/__pycache__/pythoninfo.cpython-312.pyc
evals	Retirar Python del runner	2026-09-15T11:12:58.4090272Z retirado: /opt/hostedtoolcache/Python/3.12.14/x64/lib/python3.12/test/imghdrdata/python.jpg
evals	Retirar Python del runner	2026-09-15T11:12:58.4275403Z retirado: /opt/hostedtoolcache/Python/3.12.14/x64/lib/python3.12/test/imghdrdata/python.sgi
evals	Retirar Python del runner	2026-09-15T11:12:58.4446966Z retirado: /opt/hostedtoolcache/Python/3.12.14/x64/lib/python3.12/test/imghdrdata/python.tiff
evals	Retirar Python del runner	2026-09-15T11:12:58.4619609Z retirado: /opt/hostedtoolcache/Python/3.12.14/x64/lib/python3.12/test/imghdrdata/python.bmp
evals	Retirar Python del runner	2026-09-15T11:12:58.4794973Z retirado: /opt/hostedtoolcache/Python/3.12.14/x64/lib/python3.12/test/imghdrdata/python.gif
evals	Retirar Python del runner	2026-09-15T11:12:58.4967318Z retirado: /opt/hostedtoolcache/Python/3.12.14/x64/lib/python3.12/test/imghdrdata/python.webp
evals	Retirar Python del runner	2026-09-15T11:12:58.5146314Z retirado: /opt/hostedtoolcache/Python/3.12.14/x64/lib/python3.12/test/imghdrdata/python.png
evals	Retirar Python del runner	2026-09-15T11:12:58.5324011Z retirado: /opt/hostedtoolcache/Python/3.12.14/x64/lib/python3.12/test/imghdrdata/python.ppm
evals	Retirar Python del runner	2026-09-15T11:12:58.5501599Z retirado: /opt/hostedtoolcache/Python/3.12.14/x64/lib/python3.12/test/imghdrdata/python.pgm
evals	Retirar Python del runner	2026-09-15T11:12:58.5676490Z retirado: /opt/hostedtoolcache/Python/3.12.14/x64/lib/python3.12/test/imghdrdata/python.exr
evals	Retirar Python del runner	2026-09-15T11:12:58.5849995Z retirado: /opt/hostedtoolcache/Python/3.12.14/x64/lib/python3.12/test/imghdrdata/python.xbm
evals	Retirar Python del runner	2026-09-15T11:12:58.6022568Z retirado: /opt/hostedtoolcache/Python/3.12.14/x64/lib/python3.12/test/imghdrdata/python.pbm
evals	Retirar Python del runner	2026-09-15T11:12:58.6214918Z retirado: /opt/hostedtoolcache/Python/3.12.14/x64/lib/python3.12/test/imghdrdata/python.ras
evals	Retirar Python del runner	2026-09-15T11:12:58.6397180Z retirado: /opt/hostedtoolcache/Python/3.12.14/x64/lib/python3.12/test/imghdrdata/python-raw.jpg
evals	Retirar Python del runner	2026-09-15T11:12:58.6580761Z retirado: /opt/hostedtoolcache/Python/3.12.14/x64/lib/python3.12/test/test_email/data/python.jpg
evals	Retirar Python del runner	2026-09-15T11:12:58.6753575Z retirado: /opt/hostedtoolcache/Python/3.12.14/x64/lib/python3.12/test/test_email/data/python.sgi
evals	Retirar Python del runner	2026-09-15T11:12:58.6941649Z retirado: /opt/hostedtoolcache/Python/3.12.14/x64/lib/python3.12/test/test_email/data/python.tiff
evals	Retirar Python del runner	2026-09-15T11:12:58.7124755Z retirado: /opt/hostedtoolcache/Python/3.12.14/x64/lib/python3.12/test/test_email/data/python.bmp
evals	Retirar Python del runner	2026-09-15T11:12:58.7308361Z retirado: /opt/hostedtoolcache/Python/3.12.14/x64/lib/python3.12/test/test_email/data/python.gif
evals	Retirar Python del runner	2026-09-15T11:12:58.7497054Z retirado: /opt/hostedtoolcache/Python/3.12.14/x64/lib/python3.12/test/test_email/data/python.webp
evals	Retirar Python del runner	2026-09-15T11:12:58.7691793Z retirado: /opt/hostedtoolcache/Python/3.12.14/x64/lib/python3.12/test/test_email/data/python.png
evals	Retirar Python del runner	2026-09-15T11:12:58.7880055Z retirado: /opt/hostedtoolcache/Python/3.12.14/x64/lib/python3.12/test/test_email/data/python.ppm
evals	Retirar Python del runner	2026-09-15T11:12:58.8061309Z retirado: /opt/hostedtoolcache/Python/3.12.14/x64/lib/python3.12/test/test_email/data/python.pgm
evals	Retirar Python del runner	2026-09-15T11:12:58.8248585Z retirado: /opt/hostedtoolcache/Python/3.12.14/x64/lib/python3.12/test/test_email/data/python.exr
evals	Retirar Python del runner	2026-09-15T11:12:58.8425144Z retirado: /opt/hostedtoolcache/Python/3.12.14/x64/lib/python3.12/test/test_email/data/python.xbm
evals	Retirar Python del runner	2026-09-15T11:12:58.8601663Z retirado: /opt/hostedtoolcache/Python/3.12.14/x64/lib/python3.12/test/test_email/data/python.pbm
evals	Retirar Python del runner	2026-09-15T11:12:58.8777241Z retirado: /opt/hostedtoolcache/Python/3.12.14/x64/lib/python3.12/test/test_email/data/python.ras
evals	Retirar Python del runner	2026-09-15T11:12:58.8952090Z retirado: /opt/hostedtoolcache/Python/3.12.14/x64/lib/python3.12/test/pythoninfo.py
evals	Retirar Python del runner	2026-09-15T11:12:58.9128789Z retirado: /opt/hostedtoolcache/Python/3.12.14/x64/lib/python3.12/config-3.12-x86_64-linux-gnu/libpython3.12.a
evals	Retirar Python del runner	2026-09-15T11:12:58.9309336Z retirado: /opt/hostedtoolcache/Python/3.12.14/x64/lib/python3.12/config-3.12-x86_64-linux-gnu/python-config.py
evals	Retirar Python del runner	2026-09-15T11:12:58.9504737Z retirado: /opt/hostedtoolcache/Python/3.12.14/x64/lib/python3.12/config-3.12-x86_64-linux-gnu/python.o
evals	Retirar Python del runner	2026-09-15T11:12:58.9681775Z retirado: /opt/hostedtoolcache/Python/3.12.14/x64/lib/libpython3.12.so.1.0
evals	Retirar Python del runner	2026-09-15T11:12:58.9873598Z retirado: /opt/hostedtoolcache/Python/3.12.14/x64/include/python3.12/cpython/pythonrun.h
evals	Retirar Python del runner	2026-09-15T11:12:59.0050917Z retirado: /opt/hostedtoolcache/Python/3.12.14/x64/include/python3.12/Python.h
evals	Retirar Python del runner	2026-09-15T11:12:59.0223924Z retirado: /opt/hostedtoolcache/Python/3.12.14/x64/include/python3.12/pythonrun.h
evals	Retirar Python del runner	2026-09-15T11:12:59.0403382Z retirado: /opt/hostedtoolcache/Python/3.12.14/x64/share/man/man1/python3.1
evals	Retirar Python del runner	2026-09-15T11:12:59.0576909Z retirado: /opt/hostedtoolcache/Python/3.12.14/x64/share/man/man1/python3.12.1
evals	Retirar Python del runner	2026-09-15T11:12:59.0925729Z retirado: /opt/hostedtoolcache/Python/3.12.14/x64
evals	Retirar Python del runner	2026-09-15T11:12:59.4316982Z retirado: /opt/hostedtoolcache/Python/3.11.16/x64/lib/libpython3.so
evals	Retirar Python del runner	2026-09-15T11:12:59.4490796Z retirado: /opt/hostedtoolcache/Python/3.11.16/x64/lib/pkgconfig/python3.pc
evals	Retirar Python del runner	2026-09-15T11:12:59.4666390Z retirado: /opt/hostedtoolcache/Python/3.11.16/x64/lib/pkgconfig/python3-embed.pc
evals	Retirar Python del runner	2026-09-15T11:12:59.4852350Z retirado: /opt/hostedtoolcache/Python/3.11.16/x64/lib/pkgconfig/python-3.11-embed.pc
evals	Retirar Python del runner	2026-09-15T11:12:59.5035998Z retirado: /opt/hostedtoolcache/Python/3.11.16/x64/lib/pkgconfig/python-3.11.pc
evals	Retirar Python del runner	2026-09-15T11:12:59.5214551Z retirado: /opt/hostedtoolcache/Python/3.11.16/x64/lib/python3.11/site-packages/pip/_vendor/pygments/lexers/__pycache__/python.cpython-311.pyc
evals	Retirar Python del runner	2026-09-15T11:12:59.5399772Z retirado: /opt/hostedtoolcache/Python/3.11.16/x64/lib/python3.11/site-packages/pip/_vendor/pygments/lexers/python.py
evals	Retirar Python del runner	2026-09-15T11:12:59.5586947Z retirado: /opt/hostedtoolcache/Python/3.11.16/x64/lib/python3.11/config-3.11-x86_64-linux-gnu/libpython3.11.a
evals	Retirar Python del runner	2026-09-15T11:12:59.5764524Z retirado: /opt/hostedtoolcache/Python/3.11.16/x64/lib/python3.11/config-3.11-x86_64-linux-gnu/python-config.py
evals	Retirar Python del runner	2026-09-15T11:12:59.5949033Z retirado: /opt/hostedtoolcache/Python/3.11.16/x64/lib/python3.11/config-3.11-x86_64-linux-gnu/python.o
evals	Retirar Python del runner	2026-09-15T11:12:59.6136387Z retirado: /opt/hostedtoolcache/Python/3.11.16/x64/lib/python3.11/idlelib/Icons/python.gif
evals	Retirar Python del runner	2026-09-15T11:12:59.6316331Z retirado: /opt/hostedtoolcache/Python/3.11.16/x64/lib/python3.11/test/__pycache__/pythoninfo.cpython-311.pyc
evals	Retirar Python del runner	2026-09-15T11:12:59.6501400Z retirado: /opt/hostedtoolcache/Python/3.11.16/x64/lib/python3.11/test/__pycache__/pythoninfo.cpython-311.opt-1.pyc
evals	Retirar Python del runner	2026-09-15T11:12:59.6678901Z retirado: /opt/hostedtoolcache/Python/3.11.16/x64/lib/python3.11/test/__pycache__/pythoninfo.cpython-311.opt-2.pyc
evals	Retirar Python del runner	2026-09-15T11:12:59.6856520Z retirado: /opt/hostedtoolcache/Python/3.11.16/x64/lib/python3.11/test/imghdrdata/python.jpg
evals	Retirar Python del runner	2026-09-15T11:12:59.7030660Z retirado: /opt/hostedtoolcache/Python/3.11.16/x64/lib/python3.11/test/imghdrdata/python.sgi
evals	Retirar Python del runner	2026-09-15T11:12:59.7214325Z retirado: /opt/hostedtoolcache/Python/3.11.16/x64/lib/python3.11/test/imghdrdata/python.tiff
evals	Retirar Python del runner	2026-09-15T11:12:59.7407826Z retirado: /opt/hostedtoolcache/Python/3.11.16/x64/lib/python3.11/test/imghdrdata/python.bmp
evals	Retirar Python del runner	2026-09-15T11:12:59.7591749Z retirado: /opt/hostedtoolcache/Python/3.11.16/x64/lib/python3.11/test/imghdrdata/python.gif
evals	Retirar Python del runner	2026-09-15T11:12:59.7777244Z retirado: /opt/hostedtoolcache/Python/3.11.16/x64/lib/python3.11/test/imghdrdata/python.webp
evals	Retirar Python del runner	2026-09-15T11:12:59.7959933Z retirado: /opt/hostedtoolcache/Python/3.11.16/x64/lib/python3.11/test/imghdrdata/python.png
evals	Retirar Python del runner	2026-09-15T11:12:59.8136355Z retirado: /opt/hostedtoolcache/Python/3.11.16/x64/lib/python3.11/test/imghdrdata/python.ppm
evals	Retirar Python del runner	2026-09-15T11:12:59.8326895Z retirado: /opt/hostedtoolcache/Python/3.11.16/x64/lib/python3.11/test/imghdrdata/python.pgm
evals	Retirar Python del runner	2026-09-15T11:12:59.8504062Z retirado: /opt/hostedtoolcache/Python/3.11.16/x64/lib/python3.11/test/imghdrdata/python.exr
evals	Retirar Python del runner	2026-09-15T11:12:59.8679136Z retirado: /opt/hostedtoolcache/Python/3.11.16/x64/lib/python3.11/test/imghdrdata/python.xbm
evals	Retirar Python del runner	2026-09-15T11:12:59.8854813Z retirado: /opt/hostedtoolcache/Python/3.11.16/x64/lib/python3.11/test/imghdrdata/python.pbm
evals	Retirar Python del runner	2026-09-15T11:12:59.9030587Z retirado: /opt/hostedtoolcache/Python/3.11.16/x64/lib/python3.11/test/imghdrdata/python.ras
evals	Retirar Python del runner	2026-09-15T11:12:59.9215525Z retirado: /opt/hostedtoolcache/Python/3.11.16/x64/lib/python3.11/test/imghdrdata/python-raw.jpg
evals	Retirar Python del runner	2026-09-15T11:12:59.9394474Z retirado: /opt/hostedtoolcache/Python/3.11.16/x64/lib/python3.11/test/test_email/data/python.jpg
evals	Retirar Python del runner	2026-09-15T11:12:59.9573698Z retirado: /opt/hostedtoolcache/Python/3.11.16/x64/lib/python3.11/test/test_email/data/python.sgi
evals	Retirar Python del runner	2026-09-15T11:12:59.9755350Z retirado: /opt/hostedtoolcache/Python/3.11.16/x64/lib/python3.11/test/test_email/data/python.tiff
evals	Retirar Python del runner	2026-09-15T11:12:59.9936163Z retirado: /opt/hostedtoolcache/Python/3.11.16/x64/lib/python3.11/test/test_email/data/python.bmp
evals	Retirar Python del runner	2026-09-15T11:13:00.0113442Z retirado: /opt/hostedtoolcache/Python/3.11.16/x64/lib/python3.11/test/test_email/data/python.gif
evals	Retirar Python del runner	2026-09-15T11:13:00.0290377Z retirado: /opt/hostedtoolcache/Python/3.11.16/x64/lib/python3.11/test/test_email/data/python.webp
evals	Retirar Python del runner	2026-09-15T11:13:00.0475691Z retirado: /opt/hostedtoolcache/Python/3.11.16/x64/lib/python3.11/test/test_email/data/python.png
evals	Retirar Python del runner	2026-09-15T11:13:00.0658053Z retirado: /opt/hostedtoolcache/Python/3.11.16/x64/lib/python3.11/test/test_email/data/python.ppm
evals	Retirar Python del runner	2026-09-15T11:13:00.0834198Z retirado: /opt/hostedtoolcache/Python/3.11.16/x64/lib/python3.11/test/test_email/data/python.pgm
evals	Retirar Python del runner	2026-09-15T11:13:00.1008601Z retirado: /opt/hostedtoolcache/Python/3.11.16/x64/lib/python3.11/test/test_email/data/python.exr
evals	Retirar Python del runner	2026-09-15T11:13:00.1181571Z retirado: /opt/hostedtoolcache/Python/3.11.16/x64/lib/python3.11/test/test_email/data/python.xbm
evals	Retirar Python del runner	2026-09-15T11:13:00.1357896Z retirado: /opt/hostedtoolcache/Python/3.11.16/x64/lib/python3.11/test/test_email/data/python.pbm
evals	Retirar Python del runner	2026-09-15T11:13:00.1529341Z retirado: /opt/hostedtoolcache/Python/3.11.16/x64/lib/python3.11/test/test_email/data/python.ras
evals	Retirar Python del runner	2026-09-15T11:13:00.1702515Z retirado: /opt/hostedtoolcache/Python/3.11.16/x64/lib/python3.11/test/pythoninfo.py
evals	Retirar Python del runner	2026-09-15T11:13:00.1880144Z retirado: /opt/hostedtoolcache/Python/3.11.16/x64/lib/libpython3.11.so
evals	Retirar Python del runner	2026-09-15T11:13:00.2054406Z retirado: /opt/hostedtoolcache/Python/3.11.16/x64/lib/libpython3.11.so.1.0
evals	Retirar Python del runner	2026-09-15T11:13:00.2248257Z retirado: /opt/hostedtoolcache/Python/3.11.16/x64/include/python3.11/cpython/pythonrun.h
evals	Retirar Python del runner	2026-09-15T11:13:00.2425660Z retirado: /opt/hostedtoolcache/Python/3.11.16/x64/include/python3.11/Python.h
evals	Retirar Python del runner	2026-09-15T11:13:00.2599976Z retirado: /opt/hostedtoolcache/Python/3.11.16/x64/include/python3.11/pythonrun.h
evals	Retirar Python del runner	2026-09-15T11:13:00.2780314Z retirado: /opt/hostedtoolcache/Python/3.11.16/x64/share/man/man1/python3.1
evals	Retirar Python del runner	2026-09-15T11:13:00.2955737Z retirado: /opt/hostedtoolcache/Python/3.11.16/x64/share/man/man1/python3.11.1
evals	Retirar Python del runner	2026-09-15T11:13:00.3299511Z retirado: /opt/hostedtoolcache/Python/3.11.16/x64
evals	Retirar Python del runner	2026-09-15T11:13:00.6893980Z retirado: /opt/az/lib/pkgconfig/python-3.14-embed.pc
evals	Retirar Python del runner	2026-09-15T11:13:00.7072068Z retirado: /opt/az/lib/pkgconfig/python3.pc
evals	Retirar Python del runner	2026-09-15T11:13:00.7330087Z retirado: /opt/az/lib/pkgconfig/python3-embed.pc
evals	Retirar Python del runner	2026-09-15T11:13:00.7506116Z retirado: /opt/az/lib/pkgconfig/python-3.14.pc
evals	Retirar Python del runner	2026-09-15T11:13:00.7686436Z retirado: /opt/az/lib/python3.14/site-packages/pygments/lexers/__pycache__/python.cpython-314.pyc
evals	Retirar Python del runner	2026-09-15T11:13:00.7861893Z retirado: /opt/az/lib/python3.14/site-packages/pygments/lexers/python.py
evals	Retirar Python del runner	2026-09-15T11:13:00.8038836Z retirado: /opt/az/lib/python3.14/site-packages/pip/_vendor/pygments/lexers/__pycache__/python.cpython-314.pyc
evals	Retirar Python del runner	2026-09-15T11:13:00.8213609Z retirado: /opt/az/lib/python3.14/site-packages/pip/_vendor/pygments/lexers/python.py
evals	Retirar Python del runner	2026-09-15T11:13:00.8394757Z retirado: /opt/az/lib/python3.14/site-packages/argcomplete/scripts/__pycache__/python_argcomplete_check_easy_install_script.cpython-314.pyc
evals	Retirar Python del runner	2026-09-15T11:13:00.8568078Z retirado: /opt/az/lib/python3.14/site-packages/argcomplete/scripts/python_argcomplete_check_easy_install_script.py
evals	Retirar Python del runner	2026-09-15T11:13:00.8744321Z retirado: /opt/az/lib/python3.14/config-3.14-x86_64-linux-gnu/libpython3.14.a
evals	Retirar Python del runner	2026-09-15T11:13:00.8919966Z retirado: /opt/az/lib/python3.14/config-3.14-x86_64-linux-gnu/python-config.py
evals	Retirar Python del runner	2026-09-15T11:13:00.9096185Z retirado: /opt/az/lib/python3.14/config-3.14-x86_64-linux-gnu/python.o
evals	Retirar Python del runner	2026-09-15T11:13:00.9313889Z retirado: /opt/az/lib/python3.14/idlelib/Icons/python.gif
evals	Retirar Python del runner	2026-09-15T11:13:00.9489855Z retirado: /opt/az/lib/libpython3.14.a
evals	Retirar Python del runner	2026-09-15T11:13:00.9661978Z retirado: /opt/az/include/python3.14/cpython/pythonrun.h
evals	Retirar Python del runner	2026-09-15T11:13:00.9854608Z retirado: /opt/az/include/python3.14/Python.h
evals	Retirar Python del runner	2026-09-15T11:13:01.0036952Z retirado: /opt/az/include/python3.14/pythonrun.h
evals	Retirar Python del runner	2026-09-15T11:13:01.0227463Z retirado: /opt/az/share/man/man1/python3.1
evals	Retirar Python del runner	2026-09-15T11:13:01.0409988Z retirado: /opt/az/share/man/man1/python3.14.1
evals	Retirar Python del runner	2026-09-15T11:13:01.0765683Z retirado: /opt/az
evals	Retirar Python del runner	2026-09-15T11:13:01.9211866Z retirado: /var/lib/dpkg/info/python3-jinja2.prerm
evals	Retirar Python del runner	2026-09-15T11:13:01.9395418Z retirado: /var/lib/dpkg/info/python3-packaging.prerm
evals	Retirar Python del runner	2026-09-15T11:13:01.9577202Z retirado: /var/lib/dpkg/info/python3-launchpadlib.postinst
evals	Retirar Python del runner	2026-09-15T11:13:01.9764702Z retirado: /var/lib/dpkg/info/python3-magic.postinst
evals	Retirar Python del runner	2026-09-15T11:13:01.9949940Z retirado: /var/lib/dpkg/info/python3-jsonschema.postrm
evals	Retirar Python del runner	2026-09-15T11:13:02.0129543Z retirado: /var/lib/dpkg/info/python3-jsonpatch.postinst
evals	Retirar Python del runner	2026-09-15T11:13:02.0309398Z retirado: /var/lib/dpkg/info/python3-parted.prerm
evals	Retirar Python del runner	2026-09-15T11:13:02.0498213Z retirado: /var/lib/dpkg/info/python3-chardet.postinst
evals	Retirar Python del runner	2026-09-15T11:13:02.0688559Z retirado: /var/lib/dpkg/info/python3-parted.postinst
evals	Retirar Python del runner	2026-09-15T11:13:02.0873473Z retirado: /var/lib/dpkg/info/python3-constantly.prerm
evals	Retirar Python del runner	2026-09-15T11:13:02.1054643Z retirado: /var/lib/dpkg/info/python3-gi.prerm
evals	Retirar Python del runner	2026-09-15T11:13:02.1234580Z retirado: /var/lib/dpkg/info/python3-s3transfer.prerm
evals	Retirar Python del runner	2026-09-15T11:13:02.1413640Z retirado: /var/lib/dpkg/info/python3-bcrypt.postinst
evals	Retirar Python del runner	2026-09-15T11:13:02.1598163Z retirado: /var/lib/dpkg/info/python3-netaddr.postinst
evals	Retirar Python del runner	2026-09-15T11:13:02.1776506Z retirado: /var/lib/dpkg/info/python3-zope.interface.prerm
evals	Retirar Python del runner	2026-09-15T11:13:02.1954649Z retirado: /var/lib/dpkg/info/python3-cryptography.postinst
evals	Retirar Python del runner	2026-09-15T11:13:02.2132128Z retirado: /var/lib/dpkg/info/python3-distro-info.prerm
evals	Retirar Python del runner	2026-09-15T11:13:02.2311970Z retirado: /var/lib/dpkg/info/python3-configobj.postinst
evals	Retirar Python del runner	2026-09-15T11:13:02.2491230Z retirado: /var/lib/dpkg/info/libpython3.12t64:amd64.md5sums
evals	Retirar Python del runner	2026-09-15T11:13:02.2671270Z retirado: /var/lib/dpkg/info/python3-lazr.restfulclient.prerm
evals	Retirar Python del runner	2026-09-15T11:13:02.2865543Z retirado: /var/lib/dpkg/info/python3-jsonschema.postinst
evals	Retirar Python del runner	2026-09-15T11:13:02.3051602Z retirado: /var/lib/dpkg/info/python3-six.postinst
evals	Retirar Python del runner	2026-09-15T11:13:02.3232691Z retirado: /var/lib/dpkg/info/libpython3.12-minimal:amd64.postrm
evals	Retirar Python del runner	2026-09-15T11:13:02.3416028Z retirado: /var/lib/dpkg/info/python3-idna.prerm
evals	Retirar Python del runner	2026-09-15T11:13:02.3605854Z retirado: /var/lib/dpkg/info/python3-jsonpatch.prerm
evals	Retirar Python del runner	2026-09-15T11:13:02.3795732Z retirado: /var/lib/dpkg/info/python3-cryptography.prerm
evals	Retirar Python del runner	2026-09-15T11:13:02.3984121Z retirado: /var/lib/dpkg/info/python3-babel.postinst
evals	Retirar Python del runner	2026-09-15T11:13:02.4178932Z retirado: /var/lib/dpkg/info/python3-distupgrade.postinst
evals	Retirar Python del runner	2026-09-15T11:13:02.4367228Z retirado: /var/lib/dpkg/info/python3-minimal.prerm
evals	Retirar Python del runner	2026-09-15T11:13:02.4544914Z retirado: /var/lib/dpkg/info/python3-mdurl.prerm
evals	Retirar Python del runner	2026-09-15T11:13:02.4728494Z retirado: /var/lib/dpkg/info/python3-pkg-resources.postinst
evals	Retirar Python del runner	2026-09-15T11:13:02.4918537Z retirado: /var/lib/dpkg/info/python3-launchpadlib.prerm
evals	Retirar Python del runner	2026-09-15T11:13:02.5106598Z retirado: /var/lib/dpkg/info/python3-debian.postinst
evals	Retirar Python del runner	2026-09-15T11:13:02.5289790Z retirado: /var/lib/dpkg/info/python3-wheel.prerm
evals	Retirar Python del runner	2026-09-15T11:13:02.5472463Z retirado: /var/lib/dpkg/info/python3.12-minimal.postrm
evals	Retirar Python del runner	2026-09-15T11:13:02.5654847Z retirado: /var/lib/dpkg/info/python3-certifi.postinst
evals	Retirar Python del runner	2026-09-15T11:13:02.5840312Z retirado: /var/lib/dpkg/info/python3-twisted.postrm
evals	Retirar Python del runner	2026-09-15T11:13:02.6020694Z retirado: /var/lib/dpkg/info/python3-systemd.postinst
evals	Retirar Python del runner	2026-09-15T11:13:02.6199818Z retirado: /var/lib/dpkg/info/python3-botocore.prerm
evals	Retirar Python del runner	2026-09-15T11:13:02.6382500Z retirado: /var/lib/dpkg/info/python3.12-venv.postrm
evals	Retirar Python del runner	2026-09-15T11:13:02.6567823Z retirado: /var/lib/dpkg/info/python3-openssl.postinst
evals	Retirar Python del runner	2026-09-15T11:13:02.6754387Z retirado: /var/lib/dpkg/info/python3-launchpadlib.postrm
evals	Retirar Python del runner	2026-09-15T11:13:02.6945114Z retirado: /var/lib/dpkg/info/python3-json-pointer.postinst
evals	Retirar Python del runner	2026-09-15T11:13:02.7136395Z retirado: /var/lib/dpkg/info/python3-requests.prerm
evals	Retirar Python del runner	2026-09-15T11:13:02.7321633Z retirado: /var/lib/dpkg/info/python3-pyasn1.prerm
evals	Retirar Python del runner	2026-09-15T11:13:02.7510643Z retirado: /var/lib/dpkg/info/python3-openssl.prerm
evals	Retirar Python del runner	2026-09-15T11:13:02.7691440Z retirado: /var/lib/dpkg/info/python3-attr.postinst
evals	Retirar Python del runner	2026-09-15T11:13:02.7873214Z retirado: /var/lib/dpkg/info/python3.preinst
evals	Retirar Python del runner	2026-09-15T11:13:02.8059821Z retirado: /var/lib/dpkg/info/python3-apt.prerm
evals	Retirar Python del runner	2026-09-15T11:13:02.8253800Z retirado: /var/lib/dpkg/info/python3-pyasn1-modules.postinst
evals	Retirar Python del runner	2026-09-15T11:13:02.8442293Z retirado: /var/lib/dpkg/info/libpython3.12-stdlib:amd64.md5sums
evals	Retirar Python del runner	2026-09-15T11:13:02.8634494Z retirado: /var/lib/dpkg/info/python3-newt:amd64.postinst
evals	Retirar Python del runner	2026-09-15T11:13:02.8836520Z retirado: /var/lib/dpkg/info/libpython3-stdlib:amd64.list
evals	Retirar Python del runner	2026-09-15T11:13:02.9070827Z retirado: /var/lib/dpkg/info/python3-commandnotfound.postinst
evals	Retirar Python del runner	2026-09-15T11:13:02.9291172Z retirado: /var/lib/dpkg/info/libpython3.12t64:amd64.symbols
evals	Retirar Python del runner	2026-09-15T11:13:02.9497983Z retirado: /var/lib/dpkg/info/python3-pyrsistent:amd64.prerm
evals	Retirar Python del runner	2026-09-15T11:13:02.9714085Z retirado: /var/lib/dpkg/info/python3-yaml.postinst
evals	Retirar Python del runner	2026-09-15T11:13:02.9933569Z retirado: /var/lib/dpkg/info/python3-debconf.postinst
evals	Retirar Python del runner	2026-09-15T11:13:03.0120717Z retirado: /var/lib/dpkg/info/python3-boto3.postinst
evals	Retirar Python del runner	2026-09-15T11:13:03.0313981Z retirado: /var/lib/dpkg/info/python3-passlib.prerm
evals	Retirar Python del runner	2026-09-15T11:13:03.0498991Z retirado: /var/lib/dpkg/info/python3.12.prerm
evals	Retirar Python del runner	2026-09-15T11:13:03.0684680Z retirado: /var/lib/dpkg/info/python3-idna.postinst
evals	Retirar Python del runner	2026-09-15T11:13:03.0869161Z retirado: /var/lib/dpkg/info/python3-problem-report.prerm
evals	Retirar Python del runner	2026-09-15T11:13:03.1055266Z retirado: /var/lib/dpkg/info/python3.12-venv.prerm
evals	Retirar Python del runner	2026-09-15T11:13:03.1245909Z retirado: /var/lib/dpkg/info/python3-apport.prerm
evals	Retirar Python del runner	2026-09-15T11:13:03.1428576Z retirado: /var/lib/dpkg/info/python3-newt:amd64.prerm
evals	Retirar Python del runner	2026-09-15T11:13:03.1618286Z retirado: /var/lib/dpkg/info/python3-distro-info.postinst
evals	Retirar Python del runner	2026-09-15T11:13:03.1800917Z retirado: /var/lib/dpkg/info/python3.12.postinst
evals	Retirar Python del runner	2026-09-15T11:13:03.1986313Z retirado: /var/lib/dpkg/info/python3-pip.prerm
evals	Retirar Python del runner	2026-09-15T11:13:03.2168593Z retirado: /var/lib/dpkg/info/python3.12-minimal.preinst
evals	Retirar Python del runner	2026-09-15T11:13:03.2353892Z retirado: /var/lib/dpkg/info/python3-urllib3.postinst
evals	Retirar Python del runner	2026-09-15T11:13:03.2537688Z retirado: /var/lib/dpkg/info/python3-bpfcc.prerm
evals	Retirar Python del runner	2026-09-15T11:13:03.2728960Z retirado: /var/lib/dpkg/info/python3-wadllib.postinst
evals	Retirar Python del runner	2026-09-15T11:13:03.2918562Z retirado: /var/lib/dpkg/info/python3-jwt.postinst
evals	Retirar Python del runner	2026-09-15T11:13:03.3109880Z retirado: /var/lib/dpkg/info/python3-distupgrade.prerm
evals	Retirar Python del runner	2026-09-15T11:13:03.3294617Z retirado: /var/lib/dpkg/info/python3-problem-report.postinst
evals	Retirar Python del runner	2026-09-15T11:13:03.3480671Z retirado: /var/lib/dpkg/info/python3-pexpect.postinst
evals	Retirar Python del runner	2026-09-15T11:13:03.3675585Z retirado: /var/lib/dpkg/info/python3-zstandard.postinst
evals	Retirar Python del runner	2026-09-15T11:13:03.3859061Z retirado: /var/lib/dpkg/info/python3-gi.postinst
evals	Retirar Python del runner	2026-09-15T11:13:03.4045760Z retirado: /var/lib/dpkg/info/python3-update-manager.postinst
evals	Retirar Python del runner	2026-09-15T11:13:03.4233571Z retirado: /var/lib/dpkg/info/python3-httplib2.prerm
evals	Retirar Python del runner	2026-09-15T11:13:03.4420233Z retirado: /var/lib/dpkg/info/python3-pyasn1.postinst
evals	Retirar Python del runner	2026-09-15T11:13:03.4603239Z retirado: /var/lib/dpkg/info/python3-pkg-resources.prerm
evals	Retirar Python del runner	2026-09-15T11:13:03.4789073Z retirado: /var/lib/dpkg/info/python3-markupsafe.postinst
evals	Retirar Python del runner	2026-09-15T11:13:03.4985289Z retirado: /var/lib/dpkg/info/python3-boto3.prerm
evals	Retirar Python del runner	2026-09-15T11:13:03.5170389Z retirado: /var/lib/dpkg/info/libpython3-stdlib:amd64.md5sums
evals	Retirar Python del runner	2026-09-15T11:13:03.5350077Z retirado: /var/lib/dpkg/info/libpython3.12-minimal:amd64.md5sums
evals	Retirar Python del runner	2026-09-15T11:13:03.5531040Z retirado: /var/lib/dpkg/info/python3-markdown-it.prerm
evals	Retirar Python del runner	2026-09-15T11:13:03.5711523Z retirado: /var/lib/dpkg/info/python3-distro.prerm
evals	Retirar Python del runner	2026-09-15T11:13:03.5907815Z retirado: /var/lib/dpkg/info/python3-requests.postinst
evals	Retirar Python del runner	2026-09-15T11:13:03.6098680Z retirado: /var/lib/dpkg/info/python3-hyperlink.postinst
evals	Retirar Python del runner	2026-09-15T11:13:03.6281001Z retirado: /var/lib/dpkg/info/libpython3.12t64:amd64.list
evals	Retirar Python del runner	2026-09-15T11:13:03.6475183Z retirado: /var/lib/dpkg/info/python3-hyperlink.prerm
evals	Retirar Python del runner	2026-09-15T11:13:03.6670135Z retirado: /var/lib/dpkg/info/python3-minimal.postinst
evals	Retirar Python del runner	2026-09-15T11:13:03.6854587Z retirado: /var/lib/dpkg/info/python3-jwt.prerm
evals	Retirar Python del runner	2026-09-15T11:13:03.7034553Z retirado: /var/lib/dpkg/info/python3-pyasn1-modules.prerm
evals	Retirar Python del runner	2026-09-15T11:13:03.7220536Z retirado: /var/lib/dpkg/info/python3-lazr.uri.postinst
evals	Retirar Python del runner	2026-09-15T11:13:03.7409191Z retirado: /var/lib/dpkg/info/python3-jsonpatch.postrm
evals	Retirar Python del runner	2026-09-15T11:13:03.7593697Z retirado: /var/lib/dpkg/info/python3-pygments.postinst
evals	Retirar Python del runner	2026-09-15T11:13:03.7774226Z retirado: /var/lib/dpkg/info/python3-json-pointer.postrm
evals	Retirar Python del runner	2026-09-15T11:13:03.7957149Z retirado: /var/lib/dpkg/info/libpython3.12-minimal:amd64.prerm
evals	Retirar Python del runner	2026-09-15T11:13:03.8145623Z retirado: /var/lib/dpkg/info/python3-rich.postinst
evals	Retirar Python del runner	2026-09-15T11:13:03.8335241Z retirado: /var/lib/dpkg/info/python3-jsonschema.prerm
evals	Retirar Python del runner	2026-09-15T11:13:03.8520998Z retirado: /var/lib/dpkg/info/python3-mdurl.postinst
evals	Retirar Python del runner	2026-09-15T11:13:03.8719613Z retirado: /var/lib/dpkg/info/python3-software-properties.postinst
evals	Retirar Python del runner	2026-09-15T11:13:03.8915038Z retirado: /var/lib/dpkg/info/python3-pyparsing.postinst
evals	Retirar Python del runner	2026-09-15T11:13:03.9107935Z retirado: /var/lib/dpkg/info/python3-pip.postinst
evals	Retirar Python del runner	2026-09-15T11:13:03.9301817Z retirado: /var/lib/dpkg/info/python3-distro.postinst
evals	Retirar Python del runner	2026-09-15T11:13:03.9502504Z retirado: /var/lib/dpkg/info/python3-hamcrest.prerm
evals	Retirar Python del runner	2026-09-15T11:13:03.9690647Z retirado: /var/lib/dpkg/info/python3-urllib3.prerm
evals	Retirar Python del runner	2026-09-15T11:13:03.9881624Z retirado: /var/lib/dpkg/info/python3-wadllib.prerm
evals	Retirar Python del runner	2026-09-15T11:13:04.0071909Z retirado: /var/lib/dpkg/info/python3-markupsafe.prerm
evals	Retirar Python del runner	2026-09-15T11:13:04.0260102Z retirado: /var/lib/dpkg/info/python3-httplib2.postinst
evals	Retirar Python del runner	2026-09-15T11:13:04.0456670Z retirado: /var/lib/dpkg/info/python3-certifi.prerm
evals	Retirar Python del runner	2026-09-15T11:13:04.0644269Z retirado: /var/lib/dpkg/info/python3-click.postinst
evals	Retirar Python del runner	2026-09-15T11:13:04.0829077Z retirado: /var/lib/dpkg/info/python3-constantly.postinst
evals	Retirar Python del runner	2026-09-15T11:13:04.1016491Z retirado: /var/lib/dpkg/info/libpython3-dev:amd64.list
evals	Retirar Python del runner	2026-09-15T11:13:04.1214670Z retirado: /var/lib/dpkg/info/python3.12-minimal.postinst
evals	Retirar Python del runner	2026-09-15T11:13:04.1409170Z retirado: /var/lib/dpkg/info/python3-s3transfer.postinst
evals	Retirar Python del runner	2026-09-15T11:13:04.1597835Z retirado: /var/lib/dpkg/info/python3-zstandard.prerm
evals	Retirar Python del runner	2026-09-15T11:13:04.1791062Z retirado: /var/lib/dpkg/info/python3-json-pointer.prerm
evals	Retirar Python del runner	2026-09-15T11:13:04.1977068Z retirado: /var/lib/dpkg/info/python3-service-identity.postinst
evals	Retirar Python del runner	2026-09-15T11:13:04.2163517Z retirado: /var/lib/dpkg/info/python3-serial.postinst
evals	Retirar Python del runner	2026-09-15T11:13:04.2351615Z retirado: /var/lib/dpkg/info/python3-hamcrest.postinst
evals	Retirar Python del runner	2026-09-15T11:13:04.2538288Z retirado: /var/lib/dpkg/info/python3-incremental.postinst
evals	Retirar Python del runner	2026-09-15T11:13:04.2725655Z retirado: /var/lib/dpkg/info/python3-netplan.postinst
evals	Retirar Python del runner	2026-09-15T11:13:04.2915758Z retirado: /var/lib/dpkg/info/python3-netaddr.prerm
evals	Retirar Python del runner	2026-09-15T11:13:04.3099724Z retirado: /var/lib/dpkg/info/python3-dateutil.prerm
evals	Retirar Python del runner	2026-09-15T11:13:04.3289598Z retirado: /var/lib/dpkg/info/python3-apt.postinst
evals	Retirar Python del runner	2026-09-15T11:13:04.3481855Z retirado: /var/lib/dpkg/info/python3-dbus.postinst
evals	Retirar Python del runner	2026-09-15T11:13:04.3677093Z retirado: /var/lib/dpkg/info/python3-jmespath.prerm
evals	Retirar Python del runner	2026-09-15T11:13:04.3864338Z retirado: /var/lib/dpkg/info/libpython3.12-stdlib:amd64.prerm
evals	Retirar Python del runner	2026-09-15T11:13:04.4056294Z retirado: /var/lib/dpkg/info/python3-commandnotfound.prerm
evals	Retirar Python del runner	2026-09-15T11:13:04.4258752Z retirado: /var/lib/dpkg/info/python3-blinker.prerm
evals	Retirar Python del runner	2026-09-15T11:13:04.4450477Z retirado: /var/lib/dpkg/info/python3-ptyprocess.postinst
evals	Retirar Python del runner	2026-09-15T11:13:04.4639861Z retirado: /var/lib/dpkg/info/python3-colorama.postinst
evals	Retirar Python del runner	2026-09-15T11:13:04.4843289Z retirado: /var/lib/dpkg/info/python3-wheel.postinst
evals	Retirar Python del runner	2026-09-15T11:13:04.5034483Z retirado: /var/lib/dpkg/info/python3-oauthlib.postinst
evals	Retirar Python del runner	2026-09-15T11:13:04.5222046Z retirado: /var/lib/dpkg/info/python3-pygments.prerm
evals	Retirar Python del runner	2026-09-15T11:13:04.5420687Z retirado: /var/lib/dpkg/info/python3-tz.postinst
evals	Retirar Python del runner	2026-09-15T11:13:04.5611863Z retirado: /var/lib/dpkg/info/python3.prerm
evals	Retirar Python del runner	2026-09-15T11:13:04.5795831Z retirado: /var/lib/dpkg/info/python3-update-manager.prerm
evals	Retirar Python del runner	2026-09-15T11:13:04.5982951Z retirado: /var/lib/dpkg/info/python3-pexpect.prerm
evals	Retirar Python del runner	2026-09-15T11:13:04.6174414Z retirado: /var/lib/dpkg/info/python3-serial.prerm
evals	Retirar Python del runner	2026-09-15T11:13:04.6364685Z retirado: /var/lib/dpkg/info/python3-netplan.prerm
evals	Retirar Python del runner	2026-09-15T11:13:04.6551148Z retirado: /var/lib/dpkg/info/python3-incremental.prerm
evals	Retirar Python del runner	2026-09-15T11:13:04.6739754Z retirado: /var/lib/dpkg/info/python3-typing-extensions.postinst
evals	Retirar Python del runner	2026-09-15T11:13:04.6929729Z retirado: /var/lib/dpkg/info/python3-jinja2.postinst
evals	Retirar Python del runner	2026-09-15T11:13:04.7114102Z retirado: /var/lib/dpkg/info/libpython3.12-stdlib:amd64.list
evals	Retirar Python del runner	2026-09-15T11:13:04.7298896Z retirado: /var/lib/dpkg/info/python3-pyparsing.prerm
evals	Retirar Python del runner	2026-09-15T11:13:04.7482412Z retirado: /var/lib/dpkg/info/python3-automat.postinst
evals	Retirar Python del runner	2026-09-15T11:13:04.7678513Z retirado: /var/lib/dpkg/info/python3-attr.prerm
evals	Retirar Python del runner	2026-09-15T11:13:04.7877047Z retirado: /var/lib/dpkg/info/python3-pyrsistent:amd64.postinst
evals	Retirar Python del runner	2026-09-15T11:13:04.8064287Z retirado: /var/lib/dpkg/info/python3-passlib.postinst
evals	Retirar Python del runner	2026-09-15T11:13:04.8251278Z retirado: /var/lib/dpkg/info/python3-twisted.postinst
evals	Retirar Python del runner	2026-09-15T11:13:04.8437670Z retirado: /var/lib/dpkg/info/python3-configobj.prerm
evals	Retirar Python del runner	2026-09-15T11:13:04.8620734Z retirado: /var/lib/dpkg/info/python3-markdown-it.postinst
evals	Retirar Python del runner	2026-09-15T11:13:04.8823402Z retirado: /var/lib/dpkg/info/python3-ptyprocess.prerm
evals	Retirar Python del runner	2026-09-15T11:13:04.9020312Z retirado: /var/lib/dpkg/info/libpython3-dev:amd64.md5sums
evals	Retirar Python del runner	2026-09-15T11:13:04.9222569Z retirado: /var/lib/dpkg/info/python3-software-properties.prerm
evals	Retirar Python del runner	2026-09-15T11:13:04.9421249Z retirado: /var/lib/dpkg/info/python3-dbus.prerm
evals	Retirar Python del runner	2026-09-15T11:13:04.9621056Z retirado: /var/lib/dpkg/info/libpython3.12-minimal:amd64.postinst
evals	Retirar Python del runner	2026-09-15T11:13:04.9817215Z retirado: /var/lib/dpkg/info/python3-setuptools.postinst
evals	Retirar Python del runner	2026-09-15T11:13:05.0014037Z retirado: /var/lib/dpkg/info/python3.12-minimal.prerm
evals	Retirar Python del runner	2026-09-15T11:13:05.0215188Z retirado: /var/lib/dpkg/info/python3-botocore.postinst
evals	Retirar Python del runner	2026-09-15T11:13:05.0410304Z retirado: /var/lib/dpkg/info/python3-setuptools.prerm
evals	Retirar Python del runner	2026-09-15T11:13:05.0598532Z retirado: /var/lib/dpkg/info/python3-dateutil.postinst
evals	Retirar Python del runner	2026-09-15T11:13:05.0789381Z retirado: /var/lib/dpkg/info/python3.postrm
evals	Retirar Python del runner	2026-09-15T11:13:05.0976081Z retirado: /var/lib/dpkg/info/python3-yaml.prerm
evals	Retirar Python del runner	2026-09-15T11:13:05.1155817Z retirado: /var/lib/dpkg/info/libpython3.12-dev:amd64.list
evals	Retirar Python del runner	2026-09-15T11:13:05.1341299Z retirado: /var/lib/dpkg/info/python3-lazr.restfulclient.postinst
evals	Retirar Python del runner	2026-09-15T11:13:05.1531564Z retirado: /var/lib/dpkg/info/python3-click.prerm
evals	Retirar Python del runner	2026-09-15T11:13:05.1717616Z retirado: /var/lib/dpkg/info/python3-tz.prerm
evals	Retirar Python del runner	2026-09-15T11:13:05.1901501Z retirado: /var/lib/dpkg/info/python3-debconf.prerm
evals	Retirar Python del runner	2026-09-15T11:13:05.2081115Z retirado: /var/lib/dpkg/info/libpython3.12-dev:amd64.md5sums
evals	Retirar Python del runner	2026-09-15T11:13:05.2264278Z retirado: /var/lib/dpkg/info/python3-automat.prerm
evals	Retirar Python del runner	2026-09-15T11:13:05.2446935Z retirado: /var/lib/dpkg/info/python3-systemd.prerm
evals	Retirar Python del runner	2026-09-15T11:13:05.2637265Z retirado: /var/lib/dpkg/info/python3-typing-extensions.prerm
evals	Retirar Python del runner	2026-09-15T11:13:05.2834689Z retirado: /var/lib/dpkg/info/python3-chardet.prerm
evals	Retirar Python del runner	2026-09-15T11:13:05.3021432Z retirado: /var/lib/dpkg/info/python3-packaging.postinst
evals	Retirar Python del runner	2026-09-15T11:13:05.3210666Z retirado: /var/lib/dpkg/info/libpython3.12t64:amd64.triggers
evals	Retirar Python del runner	2026-09-15T11:13:05.3402247Z retirado: /var/lib/dpkg/info/python3-blinker.postinst
evals	Retirar Python del runner	2026-09-15T11:13:05.3595928Z retirado: /var/lib/dpkg/info/python3.12-venv.postinst
evals	Retirar Python del runner	2026-09-15T11:13:05.3782013Z retirado: /var/lib/dpkg/info/python3-lazr.uri.prerm
evals	Retirar Python del runner	2026-09-15T11:13:05.3974631Z retirado: /var/lib/dpkg/info/python3-six.prerm
evals	Retirar Python del runner	2026-09-15T11:13:05.4165598Z retirado: /var/lib/dpkg/info/python3-twisted.prerm
evals	Retirar Python del runner	2026-09-15T11:13:05.4383896Z retirado: /var/lib/dpkg/info/python3-bcrypt.prerm
evals	Retirar Python del runner	2026-09-15T11:13:05.4570171Z retirado: /var/lib/dpkg/info/python3-magic.prerm
evals	Retirar Python del runner	2026-09-15T11:13:05.4758895Z retirado: /var/lib/dpkg/info/libpython3.12-minimal:amd64.list
evals	Retirar Python del runner	2026-09-15T11:13:05.4946622Z retirado: /var/lib/dpkg/info/python3-service-identity.prerm
evals	Retirar Python del runner	2026-09-15T11:13:05.5136319Z retirado: /var/lib/dpkg/info/python3-colorama.prerm
evals	Retirar Python del runner	2026-09-15T11:13:05.5322877Z retirado: /var/lib/dpkg/info/python3-jmespath.postinst
evals	Retirar Python del runner	2026-09-15T11:13:05.5508035Z retirado: /var/lib/dpkg/info/libpython3.12t64:amd64.shlibs
evals	Retirar Python del runner	2026-09-15T11:13:05.5693362Z retirado: /var/lib/dpkg/info/python3-oauthlib.prerm
evals	Retirar Python del runner	2026-09-15T11:13:05.5876429Z retirado: /var/lib/dpkg/info/python3-rich.prerm
evals	Retirar Python del runner	2026-09-15T11:13:05.6064918Z retirado: /var/lib/dpkg/info/python3-babel.prerm
evals	Retirar Python del runner	2026-09-15T11:13:05.6255999Z retirado: /var/lib/dpkg/info/python3-apport.postinst
evals	Retirar Python del runner	2026-09-15T11:13:05.6439443Z retirado: /var/lib/dpkg/info/python3-bpfcc.postinst
evals	Retirar Python del runner	2026-09-15T11:13:05.6627875Z retirado: /var/lib/dpkg/info/python3-zope.interface.postinst
evals	Retirar Python del runner	2026-09-15T11:13:05.6810716Z retirado: /var/lib/dpkg/info/libpython3.12-minimal:amd64.conffiles
evals	Retirar Python del runner	2026-09-15T11:13:05.6994460Z retirado: /var/lib/dpkg/info/python3.postinst
evals	Retirar Python del runner	2026-09-15T11:13:05.7182836Z retirado: /var/lib/dpkg/info/python3-debian.prerm
evals	Retirar Python del runner	2026-09-15T11:13:05.7376697Z retirado: /var/lib/docker/overlay2/a9853d7ac0752db045c472a9505e498ad787cec8cb7eab1bd6d6d65f685ef414/diff/var/lib/dpkg/info/python3-launchpadlib.postinst
evals	Retirar Python del runner	2026-09-15T11:13:05.7571701Z retirado: /var/lib/docker/overlay2/a9853d7ac0752db045c472a9505e498ad787cec8cb7eab1bd6d6d65f685ef414/diff/var/lib/dpkg/info/python3-cryptography.postinst
evals	Retirar Python del runner	2026-09-15T11:13:05.7771376Z retirado: /var/lib/docker/overlay2/a9853d7ac0752db045c472a9505e498ad787cec8cb7eab1bd6d6d65f685ef414/diff/var/lib/dpkg/info/python3-configobj.postinst
evals	Retirar Python del runner	2026-09-15T11:13:05.7957429Z retirado: /var/lib/docker/overlay2/a9853d7ac0752db045c472a9505e498ad787cec8cb7eab1bd6d6d65f685ef414/diff/var/lib/dpkg/info/libpython3.12t64:amd64.md5sums
evals	Retirar Python del runner	2026-09-15T11:13:05.8144709Z retirado: /var/lib/docker/overlay2/a9853d7ac0752db045c472a9505e498ad787cec8cb7eab1bd6d6d65f685ef414/diff/var/lib/dpkg/info/python3-lazr.restfulclient.prerm
evals	Retirar Python del runner	2026-09-15T11:13:05.8328583Z retirado: /var/lib/docker/overlay2/a9853d7ac0752db045c472a9505e498ad787cec8cb7eab1bd6d6d65f685ef414/diff/var/lib/dpkg/info/python3-six.postinst
evals	Retirar Python del runner	2026-09-15T11:13:05.8525673Z retirado: /var/lib/docker/overlay2/a9853d7ac0752db045c472a9505e498ad787cec8cb7eab1bd6d6d65f685ef414/diff/var/lib/dpkg/info/libpython3.12-minimal:amd64.postrm
evals	Retirar Python del runner	2026-09-15T11:13:05.8712537Z retirado: /var/lib/docker/overlay2/a9853d7ac0752db045c472a9505e498ad787cec8cb7eab1bd6d6d65f685ef414/diff/var/lib/dpkg/info/python3-cryptography.prerm
evals	Retirar Python del runner	2026-09-15T11:13:05.8898506Z retirado: /var/lib/docker/overlay2/a9853d7ac0752db045c472a9505e498ad787cec8cb7eab1bd6d6d65f685ef414/diff/var/lib/dpkg/info/python3-minimal.prerm
evals	Retirar Python del runner	2026-09-15T11:13:05.9087825Z retirado: /var/lib/docker/overlay2/a9853d7ac0752db045c472a9505e498ad787cec8cb7eab1bd6d6d65f685ef414/diff/var/lib/dpkg/info/python3-pkg-resources.postinst
evals	Retirar Python del runner	2026-09-15T11:13:05.9280194Z retirado: /var/lib/docker/overlay2/a9853d7ac0752db045c472a9505e498ad787cec8cb7eab1bd6d6d65f685ef414/diff/var/lib/dpkg/info/python3-launchpadlib.prerm
evals	Retirar Python del runner	2026-09-15T11:13:05.9467565Z retirado: /var/lib/docker/overlay2/a9853d7ac0752db045c472a9505e498ad787cec8cb7eab1bd6d6d65f685ef414/diff/var/lib/dpkg/info/python3.12-minimal.postrm
evals	Retirar Python del runner	2026-09-15T11:13:05.9660534Z retirado: /var/lib/docker/overlay2/a9853d7ac0752db045c472a9505e498ad787cec8cb7eab1bd6d6d65f685ef414/diff/var/lib/dpkg/info/python3-launchpadlib.postrm
evals	Retirar Python del runner	2026-09-15T11:13:05.9844960Z retirado: /var/lib/docker/overlay2/a9853d7ac0752db045c472a9505e498ad787cec8cb7eab1bd6d6d65f685ef414/diff/var/lib/dpkg/info/python3.preinst
evals	Retirar Python del runner	2026-09-15T11:13:06.0031758Z retirado: /var/lib/docker/overlay2/a9853d7ac0752db045c472a9505e498ad787cec8cb7eab1bd6d6d65f685ef414/diff/var/lib/dpkg/info/python3-tzlocal.postinst
evals	Retirar Python del runner	2026-09-15T11:13:06.0215893Z retirado: /var/lib/docker/overlay2/a9853d7ac0752db045c472a9505e498ad787cec8cb7eab1bd6d6d65f685ef414/diff/var/lib/dpkg/info/libpython3.12-stdlib:amd64.md5sums
evals	Retirar Python del runner	2026-09-15T11:13:06.0402066Z retirado: /var/lib/docker/overlay2/a9853d7ac0752db045c472a9505e498ad787cec8cb7eab1bd6d6d65f685ef414/diff/var/lib/dpkg/info/libpython3-stdlib:amd64.list
evals	Retirar Python del runner	2026-09-15T11:13:06.0592909Z retirado: /var/lib/docker/overlay2/a9853d7ac0752db045c472a9505e498ad787cec8cb7eab1bd6d6d65f685ef414/diff/var/lib/dpkg/info/libpython3.12t64:amd64.symbols
evals	Retirar Python del runner	2026-09-15T11:13:06.0779087Z retirado: /var/lib/docker/overlay2/a9853d7ac0752db045c472a9505e498ad787cec8cb7eab1bd6d6d65f685ef414/diff/var/lib/dpkg/info/python3-yaml.postinst
evals	Retirar Python del runner	2026-09-15T11:13:06.0967467Z retirado: /var/lib/docker/overlay2/a9853d7ac0752db045c472a9505e498ad787cec8cb7eab1bd6d6d65f685ef414/diff/var/lib/dpkg/info/python3.12.prerm
evals	Retirar Python del runner	2026-09-15T11:13:06.1155005Z retirado: /var/lib/docker/overlay2/a9853d7ac0752db045c472a9505e498ad787cec8cb7eab1bd6d6d65f685ef414/diff/var/lib/dpkg/info/python3.12.postinst
evals	Retirar Python del runner	2026-09-15T11:13:06.1341341Z retirado: /var/lib/docker/overlay2/a9853d7ac0752db045c472a9505e498ad787cec8cb7eab1bd6d6d65f685ef414/diff/var/lib/dpkg/info/python3.12-minimal.preinst
evals	Retirar Python del runner	2026-09-15T11:13:06.1530123Z retirado: /var/lib/docker/overlay2/a9853d7ac0752db045c472a9505e498ad787cec8cb7eab1bd6d6d65f685ef414/diff/var/lib/dpkg/info/python3-merge3.postinst
evals	Retirar Python del runner	2026-09-15T11:13:06.1718377Z retirado: /var/lib/docker/overlay2/a9853d7ac0752db045c472a9505e498ad787cec8cb7eab1bd6d6d65f685ef414/diff/var/lib/dpkg/info/python3-urllib3.postinst
evals	Retirar Python del runner	2026-09-15T11:13:06.1904717Z retirado: /var/lib/docker/overlay2/a9853d7ac0752db045c472a9505e498ad787cec8cb7eab1bd6d6d65f685ef414/diff/var/lib/dpkg/info/python3-wadllib.postinst
evals	Retirar Python del runner	2026-09-15T11:13:06.2087877Z retirado: /var/lib/docker/overlay2/a9853d7ac0752db045c472a9505e498ad787cec8cb7eab1bd6d6d65f685ef414/diff/var/lib/dpkg/info/python3-jwt.postinst
evals	Retirar Python del runner	2026-09-15T11:13:06.2272938Z retirado: /var/lib/docker/overlay2/a9853d7ac0752db045c472a9505e498ad787cec8cb7eab1bd6d6d65f685ef414/diff/var/lib/dpkg/info/python3-fastbencode.prerm
evals	Retirar Python del runner	2026-09-15T11:13:06.2463298Z retirado: /var/lib/docker/overlay2/a9853d7ac0752db045c472a9505e498ad787cec8cb7eab1bd6d6d65f685ef414/diff/var/lib/dpkg/info/python3-httplib2.prerm
evals	Retirar Python del runner	2026-09-15T11:13:06.2647664Z retirado: /var/lib/docker/overlay2/a9853d7ac0752db045c472a9505e498ad787cec8cb7eab1bd6d6d65f685ef414/diff/var/lib/dpkg/info/python3-pkg-resources.prerm
evals	Retirar Python del runner	2026-09-15T11:13:06.2829392Z retirado: /var/lib/docker/overlay2/a9853d7ac0752db045c472a9505e498ad787cec8cb7eab1bd6d6d65f685ef414/diff/var/lib/dpkg/info/libpython3-stdlib:amd64.md5sums
evals	Retirar Python del runner	2026-09-15T11:13:06.3013601Z retirado: /var/lib/docker/overlay2/a9853d7ac0752db045c472a9505e498ad787cec8cb7eab1bd6d6d65f685ef414/diff/var/lib/dpkg/info/libpython3.12-minimal:amd64.md5sums
evals	Retirar Python del runner	2026-09-15T11:13:06.3193937Z retirado: /var/lib/docker/overlay2/a9853d7ac0752db045c472a9505e498ad787cec8cb7eab1bd6d6d65f685ef414/diff/var/lib/dpkg/info/python3-distro.prerm
evals	Retirar Python del runner	2026-09-15T11:13:06.3388276Z retirado: /var/lib/docker/overlay2/a9853d7ac0752db045c472a9505e498ad787cec8cb7eab1bd6d6d65f685ef414/diff/var/lib/dpkg/info/libpython3.12t64:amd64.list
evals	Retirar Python del runner	2026-09-15T11:13:06.3577264Z retirado: /var/lib/docker/overlay2/a9853d7ac0752db045c472a9505e498ad787cec8cb7eab1bd6d6d65f685ef414/diff/var/lib/dpkg/info/python3-minimal.postinst
evals	Retirar Python del runner	2026-09-15T11:13:06.3770108Z retirado: /var/lib/docker/overlay2/a9853d7ac0752db045c472a9505e498ad787cec8cb7eab1bd6d6d65f685ef414/diff/var/lib/dpkg/info/python3-jwt.prerm
evals	Retirar Python del runner	2026-09-15T11:13:06.3966321Z retirado: /var/lib/docker/overlay2/a9853d7ac0752db045c472a9505e498ad787cec8cb7eab1bd6d6d65f685ef414/diff/var/lib/dpkg/info/python3-lazr.uri.postinst
evals	Retirar Python del runner	2026-09-15T11:13:06.4152845Z retirado: /var/lib/docker/overlay2/a9853d7ac0752db045c472a9505e498ad787cec8cb7eab1bd6d6d65f685ef414/diff/var/lib/dpkg/info/python3-tzlocal.prerm
evals	Retirar Python del runner	2026-09-15T11:13:06.4338423Z retirado: /var/lib/docker/overlay2/a9853d7ac0752db045c472a9505e498ad787cec8cb7eab1bd6d6d65f685ef414/diff/var/lib/dpkg/info/libpython3.12-minimal:amd64.prerm
evals	Retirar Python del runner	2026-09-15T11:13:06.4528052Z retirado: /var/lib/docker/overlay2/a9853d7ac0752db045c472a9505e498ad787cec8cb7eab1bd6d6d65f685ef414/diff/var/lib/dpkg/info/python3-pyparsing.postinst
evals	Retirar Python del runner	2026-09-15T11:13:06.4721463Z retirado: /var/lib/docker/overlay2/a9853d7ac0752db045c472a9505e498ad787cec8cb7eab1bd6d6d65f685ef414/diff/var/lib/dpkg/info/python3-distro.postinst
evals	Retirar Python del runner	2026-09-15T11:13:06.4915094Z retirado: /var/lib/docker/overlay2/a9853d7ac0752db045c472a9505e498ad787cec8cb7eab1bd6d6d65f685ef414/diff/var/lib/dpkg/info/python3-urllib3.prerm
evals	Retirar Python del runner	2026-09-15T11:13:06.5118086Z retirado: /var/lib/docker/overlay2/a9853d7ac0752db045c472a9505e498ad787cec8cb7eab1bd6d6d65f685ef414/diff/var/lib/dpkg/info/python3-wadllib.prerm
evals	Retirar Python del runner	2026-09-15T11:13:06.5315192Z retirado: /var/lib/docker/overlay2/a9853d7ac0752db045c472a9505e498ad787cec8cb7eab1bd6d6d65f685ef414/diff/var/lib/dpkg/info/python3-httplib2.postinst
evals	Retirar Python del runner	2026-09-15T11:13:06.5506176Z retirado: /var/lib/docker/overlay2/a9853d7ac0752db045c472a9505e498ad787cec8cb7eab1bd6d6d65f685ef414/diff/var/lib/dpkg/info/python3.12-minimal.postinst
evals	Retirar Python del runner	2026-09-15T11:13:06.5689728Z retirado: /var/lib/docker/overlay2/a9853d7ac0752db045c472a9505e498ad787cec8cb7eab1bd6d6d65f685ef414/diff/var/lib/dpkg/info/python3-patiencediff.postinst
evals	Retirar Python del runner	2026-09-15T11:13:06.5875487Z retirado: /var/lib/docker/overlay2/a9853d7ac0752db045c472a9505e498ad787cec8cb7eab1bd6d6d65f685ef414/diff/var/lib/dpkg/info/libpython3.12-stdlib:amd64.prerm
evals	Retirar Python del runner	2026-09-15T11:13:06.6062439Z retirado: /var/lib/docker/overlay2/a9853d7ac0752db045c472a9505e498ad787cec8cb7eab1bd6d6d65f685ef414/diff/var/lib/dpkg/info/python3-blinker.prerm
evals	Retirar Python del runner	2026-09-15T11:13:06.6247485Z retirado: /var/lib/docker/overlay2/a9853d7ac0752db045c472a9505e498ad787cec8cb7eab1bd6d6d65f685ef414/diff/var/lib/dpkg/info/python3-oauthlib.postinst
evals	Retirar Python del runner	2026-09-15T11:13:06.6429897Z retirado: /var/lib/docker/overlay2/a9853d7ac0752db045c472a9505e498ad787cec8cb7eab1bd6d6d65f685ef414/diff/var/lib/dpkg/info/python3.prerm
evals	Retirar Python del runner	2026-09-15T11:13:06.6616826Z retirado: /var/lib/docker/overlay2/a9853d7ac0752db045c472a9505e498ad787cec8cb7eab1bd6d6d65f685ef414/diff/var/lib/dpkg/info/libpython3.12-stdlib:amd64.list
evals	Retirar Python del runner	2026-09-15T11:13:06.6795429Z retirado: /var/lib/docker/overlay2/a9853d7ac0752db045c472a9505e498ad787cec8cb7eab1bd6d6d65f685ef414/diff/var/lib/dpkg/info/python3-pyparsing.prerm
evals	Retirar Python del runner	2026-09-15T11:13:06.6989280Z retirado: /var/lib/docker/overlay2/a9853d7ac0752db045c472a9505e498ad787cec8cb7eab1bd6d6d65f685ef414/diff/var/lib/dpkg/info/python3-breezy.prerm
evals	Retirar Python del runner	2026-09-15T11:13:06.7176170Z retirado: /var/lib/docker/overlay2/a9853d7ac0752db045c472a9505e498ad787cec8cb7eab1bd6d6d65f685ef414/diff/var/lib/dpkg/info/python3-dulwich.postinst
evals	Retirar Python del runner	2026-09-15T11:13:06.7363860Z retirado: /var/lib/docker/overlay2/a9853d7ac0752db045c472a9505e498ad787cec8cb7eab1bd6d6d65f685ef414/diff/var/lib/dpkg/info/python3-configobj.prerm
evals	Retirar Python del runner	2026-09-15T11:13:06.7561126Z retirado: /var/lib/docker/overlay2/a9853d7ac0752db045c472a9505e498ad787cec8cb7eab1bd6d6d65f685ef414/diff/var/lib/dpkg/info/libpython3.12-minimal:amd64.postinst
evals	Retirar Python del runner	2026-09-15T11:13:06.7749730Z retirado: /var/lib/docker/overlay2/a9853d7ac0752db045c472a9505e498ad787cec8cb7eab1bd6d6d65f685ef414/diff/var/lib/dpkg/info/python3.12-minimal.prerm
evals	Retirar Python del runner	2026-09-15T11:13:06.7929288Z retirado: /var/lib/docker/overlay2/a9853d7ac0752db045c472a9505e498ad787cec8cb7eab1bd6d6d65f685ef414/diff/var/lib/dpkg/info/python3-dulwich.prerm
evals	Retirar Python del runner	2026-09-15T11:13:06.8119152Z retirado: /var/lib/docker/overlay2/a9853d7ac0752db045c472a9505e498ad787cec8cb7eab1bd6d6d65f685ef414/diff/var/lib/dpkg/info/python3.postrm
evals	Retirar Python del runner	2026-09-15T11:13:06.8314985Z retirado: /var/lib/docker/overlay2/a9853d7ac0752db045c472a9505e498ad787cec8cb7eab1bd6d6d65f685ef414/diff/var/lib/dpkg/info/python3-yaml.prerm
evals	Retirar Python del runner	2026-09-15T11:13:06.8499650Z retirado: /var/lib/docker/overlay2/a9853d7ac0752db045c472a9505e498ad787cec8cb7eab1bd6d6d65f685ef414/diff/var/lib/dpkg/info/python3-lazr.restfulclient.postinst
evals	Retirar Python del runner	2026-09-15T11:13:06.8694144Z retirado: /var/lib/docker/overlay2/a9853d7ac0752db045c472a9505e498ad787cec8cb7eab1bd6d6d65f685ef414/diff/var/lib/dpkg/info/python3-patiencediff.prerm
evals	Retirar Python del runner	2026-09-15T11:13:06.8879024Z retirado: /var/lib/docker/overlay2/a9853d7ac0752db045c472a9505e498ad787cec8cb7eab1bd6d6d65f685ef414/diff/var/lib/dpkg/info/python3-merge3.prerm
evals	Retirar Python del runner	2026-09-15T11:13:06.9068047Z retirado: /var/lib/docker/overlay2/a9853d7ac0752db045c472a9505e498ad787cec8cb7eab1bd6d6d65f685ef414/diff/var/lib/dpkg/info/libpython3.12t64:amd64.triggers
evals	Retirar Python del runner	2026-09-15T11:13:06.9261381Z retirado: /var/lib/docker/overlay2/a9853d7ac0752db045c472a9505e498ad787cec8cb7eab1bd6d6d65f685ef414/diff/var/lib/dpkg/info/python3-blinker.postinst
evals	Retirar Python del runner	2026-09-15T11:13:06.9446364Z retirado: /var/lib/docker/overlay2/a9853d7ac0752db045c472a9505e498ad787cec8cb7eab1bd6d6d65f685ef414/diff/var/lib/dpkg/info/python3-lazr.uri.prerm
evals	Retirar Python del runner	2026-09-15T11:13:06.9630089Z retirado: /var/lib/docker/overlay2/a9853d7ac0752db045c472a9505e498ad787cec8cb7eab1bd6d6d65f685ef414/diff/var/lib/dpkg/info/python3-six.prerm
evals	Retirar Python del runner	2026-09-15T11:13:06.9815836Z retirado: /var/lib/docker/overlay2/a9853d7ac0752db045c472a9505e498ad787cec8cb7eab1bd6d6d65f685ef414/diff/var/lib/dpkg/info/python3-fastbencode.postinst
evals	Retirar Python del runner	2026-09-15T11:13:06.9999607Z retirado: /var/lib/docker/overlay2/a9853d7ac0752db045c472a9505e498ad787cec8cb7eab1bd6d6d65f685ef414/diff/var/lib/dpkg/info/libpython3.12-minimal:amd64.list
evals	Retirar Python del runner	2026-09-15T11:13:07.0185592Z retirado: /var/lib/docker/overlay2/a9853d7ac0752db045c472a9505e498ad787cec8cb7eab1bd6d6d65f685ef414/diff/var/lib/dpkg/info/libpython3.12t64:amd64.shlibs
evals	Retirar Python del runner	2026-09-15T11:13:07.0370506Z retirado: /var/lib/docker/overlay2/a9853d7ac0752db045c472a9505e498ad787cec8cb7eab1bd6d6d65f685ef414/diff/var/lib/dpkg/info/python3-breezy.postinst
evals	Retirar Python del runner	2026-09-15T11:13:07.0558199Z retirado: /var/lib/docker/overlay2/a9853d7ac0752db045c472a9505e498ad787cec8cb7eab1bd6d6d65f685ef414/diff/var/lib/dpkg/info/python3-oauthlib.prerm
evals	Retirar Python del runner	2026-09-15T11:13:07.0747715Z retirado: /var/lib/docker/overlay2/a9853d7ac0752db045c472a9505e498ad787cec8cb7eab1bd6d6d65f685ef414/diff/var/lib/dpkg/info/libpython3.12-minimal:amd64.conffiles
evals	Retirar Python del runner	2026-09-15T11:13:07.0933274Z retirado: /var/lib/docker/overlay2/a9853d7ac0752db045c472a9505e498ad787cec8cb7eab1bd6d6d65f685ef414/diff/var/lib/dpkg/info/python3.postinst
evals	Retirar Python del runner	2026-09-15T11:13:07.1124628Z retirado: /var/lib/docker/overlay2/a9853d7ac0752db045c472a9505e498ad787cec8cb7eab1bd6d6d65f685ef414/diff/usr/lib/x86_64-linux-gnu/libpython3.12.so.1
evals	Retirar Python del runner	2026-09-15T11:13:07.1315108Z retirado: /var/lib/docker/overlay2/a9853d7ac0752db045c472a9505e498ad787cec8cb7eab1bd6d6d65f685ef414/diff/usr/lib/x86_64-linux-gnu/libpython3.12.so.1.0
evals	Retirar Python del runner	2026-09-15T11:13:07.1582041Z retirado: /var/lib/docker/overlay2/a9853d7ac0752db045c472a9505e498ad787cec8cb7eab1bd6d6d65f685ef414/diff/usr/lib/python3.12/config-3.12-x86_64-linux-gnu/libpython3.12.so
evals	Retirar Python del runner	2026-09-15T11:13:07.1770351Z retirado: /var/lib/docker/overlay2/a9853d7ac0752db045c472a9505e498ad787cec8cb7eab1bd6d6d65f685ef414/diff/usr/share/pixmaps/python3.xpm
evals	Retirar Python del runner	2026-09-15T11:13:07.1958314Z retirado: /var/lib/docker/overlay2/a9853d7ac0752db045c472a9505e498ad787cec8cb7eab1bd6d6d65f685ef414/diff/usr/share/doc/libpython3.12t64
evals	Retirar Python del runner	2026-09-15T11:13:07.2237818Z retirado: /var/lib/docker/overlay2/a9853d7ac0752db045c472a9505e498ad787cec8cb7eab1bd6d6d65f685ef414/diff/usr/share/doc/python3.12/python-policy.txt.gz
evals	Retirar Python del runner	2026-09-15T11:13:07.2433568Z retirado: /var/lib/docker/overlay2/a9853d7ac0752db045c472a9505e498ad787cec8cb7eab1bd6d6d65f685ef414/diff/usr/share/doc/libpython3.12-stdlib
evals	Retirar Python del runner	2026-09-15T11:13:07.2622656Z retirado: /var/lib/docker/overlay2/a9853d7ac0752db045c472a9505e498ad787cec8cb7eab1bd6d6d65f685ef414/diff/usr/share/lintian/overrides/libpython3.12t64
evals	Retirar Python del runner	2026-09-15T11:13:07.2812225Z retirado: /var/lib/docker/overlay2/a9853d7ac0752db045c472a9505e498ad787cec8cb7eab1bd6d6d65f685ef414/diff/usr/share/lintian/overrides/libpython3.12-minimal
evals	Retirar Python del runner	2026-09-15T11:13:07.3000222Z retirado: /var/lib/docker/overlay2/a9853d7ac0752db045c472a9505e498ad787cec8cb7eab1bd6d6d65f685ef414/diff/usr/share/lintian/overrides/libpython3.12-stdlib
evals	Retirar Python del runner	2026-09-15T11:13:07.3370536Z retirado: /var/lib/docker/overlay2/a9853d7ac0752db045c472a9505e498ad787cec8cb7eab1bd6d6d65f685ef414/diff/usr
evals	Retirar Python del runner	2026-09-15T11:13:07.7606983Z retirado: /var/lib/docker/overlay2/80c613159e0ff38a882acb63fd4cf342723381c4f07d119df138a79409f5e793/diff/var/lib/dpkg/info/python3.10-minimal.prerm
evals	Retirar Python del runner	2026-09-15T11:13:07.7986485Z retirado: /var/lib/docker/overlay2/80c613159e0ff38a882acb63fd4cf342723381c4f07d119df138a79409f5e793/diff/var/lib/dpkg/info/libpython3.10-minimal:amd64.postinst
evals	Retirar Python del runner	2026-09-15T11:13:07.8306458Z retirado: /var/lib/docker/overlay2/80c613159e0ff38a882acb63fd4cf342723381c4f07d119df138a79409f5e793/diff/var/lib/dpkg/info/python3-minimal.prerm
evals	Retirar Python del runner	2026-09-15T11:13:07.8650577Z retirado: /var/lib/docker/overlay2/80c613159e0ff38a882acb63fd4cf342723381c4f07d119df138a79409f5e793/diff/var/lib/dpkg/info/libpython3.10-minimal:amd64.postrm
evals	Retirar Python del runner	2026-09-15T11:13:07.8954987Z retirado: /var/lib/docker/overlay2/80c613159e0ff38a882acb63fd4cf342723381c4f07d119df138a79409f5e793/diff/var/lib/dpkg/info/python3.preinst
evals	Retirar Python del runner	2026-09-15T11:13:07.9277694Z retirado: /var/lib/docker/overlay2/80c613159e0ff38a882acb63fd4cf342723381c4f07d119df138a79409f5e793/diff/var/lib/dpkg/info/libpython3-stdlib:amd64.list
evals	Retirar Python del runner	2026-09-15T11:13:07.9621745Z retirado: /var/lib/docker/overlay2/80c613159e0ff38a882acb63fd4cf342723381c4f07d119df138a79409f5e793/diff/var/lib/dpkg/info/libpython3.10-minimal:amd64.prerm
evals	Retirar Python del runner	2026-09-15T11:13:07.9972117Z retirado: /var/lib/docker/overlay2/80c613159e0ff38a882acb63fd4cf342723381c4f07d119df138a79409f5e793/diff/var/lib/dpkg/info/python3.10-minimal.postrm
evals	Retirar Python del runner	2026-09-15T11:13:08.0282409Z retirado: /var/lib/docker/overlay2/80c613159e0ff38a882acb63fd4cf342723381c4f07d119df138a79409f5e793/diff/var/lib/dpkg/info/libpython3.10-minimal:amd64.list
evals	Retirar Python del runner	2026-09-15T11:13:08.0647177Z retirado: /var/lib/docker/overlay2/80c613159e0ff38a882acb63fd4cf342723381c4f07d119df138a79409f5e793/diff/var/lib/dpkg/info/python3.10.prerm
evals	Retirar Python del runner	2026-09-15T11:13:08.0938426Z retirado: /var/lib/docker/overlay2/80c613159e0ff38a882acb63fd4cf342723381c4f07d119df138a79409f5e793/diff/var/lib/dpkg/info/libpython3-stdlib:amd64.md5sums
evals	Retirar Python del runner	2026-09-15T11:13:08.1256131Z retirado: /var/lib/docker/overlay2/80c613159e0ff38a882acb63fd4cf342723381c4f07d119df138a79409f5e793/diff/var/lib/dpkg/info/python3-minimal.postinst
evals	Retirar Python del runner	2026-09-15T11:13:08.1558643Z retirado: /var/lib/docker/overlay2/80c613159e0ff38a882acb63fd4cf342723381c4f07d119df138a79409f5e793/diff/var/lib/dpkg/info/libpython3.10-stdlib:amd64.list
evals	Retirar Python del runner	2026-09-15T11:13:08.1867937Z retirado: /var/lib/docker/overlay2/80c613159e0ff38a882acb63fd4cf342723381c4f07d119df138a79409f5e793/diff/var/lib/dpkg/info/python3.10-minimal.postinst
evals	Retirar Python del runner	2026-09-15T11:13:08.2194391Z retirado: /var/lib/docker/overlay2/80c613159e0ff38a882acb63fd4cf342723381c4f07d119df138a79409f5e793/diff/var/lib/dpkg/info/python3.10.postinst
evals	Retirar Python del runner	2026-09-15T11:13:08.2530714Z retirado: /var/lib/docker/overlay2/80c613159e0ff38a882acb63fd4cf342723381c4f07d119df138a79409f5e793/diff/var/lib/dpkg/info/python3.prerm
evals	Retirar Python del runner	2026-09-15T11:13:08.2820709Z retirado: /var/lib/docker/overlay2/80c613159e0ff38a882acb63fd4cf342723381c4f07d119df138a79409f5e793/diff/var/lib/dpkg/info/libpython3.10-stdlib:amd64.prerm
evals	Retirar Python del runner	2026-09-15T11:13:08.3120211Z retirado: /var/lib/docker/overlay2/80c613159e0ff38a882acb63fd4cf342723381c4f07d119df138a79409f5e793/diff/var/lib/dpkg/info/libpython3.10-minimal:amd64.md5sums
evals	Retirar Python del runner	2026-09-15T11:13:08.3426451Z retirado: /var/lib/docker/overlay2/80c613159e0ff38a882acb63fd4cf342723381c4f07d119df138a79409f5e793/diff/var/lib/dpkg/info/libpython3.10-stdlib:amd64.md5sums
evals	Retirar Python del runner	2026-09-15T11:13:08.3757998Z retirado: /var/lib/docker/overlay2/80c613159e0ff38a882acb63fd4cf342723381c4f07d119df138a79409f5e793/diff/var/lib/dpkg/info/python3.postrm
evals	Retirar Python del runner	2026-09-15T11:13:08.4064760Z retirado: /var/lib/docker/overlay2/80c613159e0ff38a882acb63fd4cf342723381c4f07d119df138a79409f5e793/diff/var/lib/dpkg/info/libpython3.10-minimal:amd64.conffiles
evals	Retirar Python del runner	2026-09-15T11:13:08.4365023Z retirado: /var/lib/docker/overlay2/80c613159e0ff38a882acb63fd4cf342723381c4f07d119df138a79409f5e793/diff/var/lib/dpkg/info/python3.10-minimal.preinst
evals	Retirar Python del runner	2026-09-15T11:13:08.4631637Z retirado: /var/lib/docker/overlay2/80c613159e0ff38a882acb63fd4cf342723381c4f07d119df138a79409f5e793/diff/var/lib/dpkg/info/python3.postinst
evals	Retirar Python del runner	2026-09-15T11:13:08.4821132Z retirado: /var/lib/docker/overlay2/80c613159e0ff38a882acb63fd4cf342723381c4f07d119df138a79409f5e793/diff/usr/share/pixmaps/python3.xpm
evals	Retirar Python del runner	2026-09-15T11:13:08.5022907Z retirado: /var/lib/docker/overlay2/80c613159e0ff38a882acb63fd4cf342723381c4f07d119df138a79409f5e793/diff/usr/share/doc/libpython3.10-stdlib
evals	Retirar Python del runner	2026-09-15T11:13:08.5298581Z retirado: /var/lib/docker/overlay2/80c613159e0ff38a882acb63fd4cf342723381c4f07d119df138a79409f5e793/diff/usr/share/doc/python3.10/python-policy.txt.gz
evals	Retirar Python del runner	2026-09-15T11:13:08.5492233Z retirado: /var/lib/docker/overlay2/80c613159e0ff38a882acb63fd4cf342723381c4f07d119df138a79409f5e793/diff/usr/share/lintian/overrides/libpython3.10-stdlib
evals	Retirar Python del runner	2026-09-15T11:13:08.5680070Z retirado: /var/lib/docker/overlay2/80c613159e0ff38a882acb63fd4cf342723381c4f07d119df138a79409f5e793/diff/usr/share/lintian/overrides/libpython3.10-minimal
evals	Retirar Python del runner	2026-09-15T11:13:08.6048500Z retirado: /var/lib/docker/overlay2/80c613159e0ff38a882acb63fd4cf342723381c4f07d119df138a79409f5e793/diff/usr
evals	Retirar Python del runner	2026-09-15T11:13:08.9561321Z retirado: /usr/lib/x86_64-linux-gnu/pkgconfig/python3.pc
evals	Retirar Python del runner	2026-09-15T11:13:08.9754759Z retirado: /usr/lib/x86_64-linux-gnu/pkgconfig/python3-embed.pc
evals	Retirar Python del runner	2026-09-15T11:13:08.9948994Z retirado: /usr/lib/x86_64-linux-gnu/libpython3.12.so.1
evals	Retirar Python del runner	2026-09-15T11:13:09.0229029Z retirado: /usr/lib/x86_64-linux-gnu/libpython3.12.so
evals	Retirar Python del runner	2026-09-15T11:13:09.0428218Z retirado: /usr/lib/x86_64-linux-gnu/libpython3.12.a
evals	Retirar Python del runner	2026-09-15T11:13:09.0615889Z retirado: /usr/lib/x86_64-linux-gnu/libpython3.12.so.1.0
evals	Retirar Python del runner	2026-09-15T11:13:09.0879350Z retirado: /usr/lib/python3.12/config-3.12-x86_64-linux-gnu/libpython3.12.so
evals	Retirar Python del runner	2026-09-15T11:13:09.1062323Z retirado: /usr/lib/python3.12/config-3.12-x86_64-linux-gnu/libpython3.12.a
evals	Retirar Python del runner	2026-09-15T11:13:09.1249603Z retirado: /usr/lib/python3.12/config-3.12-x86_64-linux-gnu/python-config.py
evals	Retirar Python del runner	2026-09-15T11:13:09.1440998Z retirado: /usr/lib/python3.12/config-3.12-x86_64-linux-gnu/libpython3.12-pic.a
evals	Retirar Python del runner	2026-09-15T11:13:09.1628096Z retirado: /usr/lib/google-cloud-sdk/lib/googlecloudsdk/command_lib/orchestration_pipelines/tools/python_environment_unpack.sh
evals	Retirar Python del runner	2026-09-15T11:13:09.1816647Z retirado: /usr/lib/google-cloud-sdk/platform/bundledpythonunix/lib/libpython3.so
evals	Retirar Python del runner	2026-09-15T11:13:09.2007031Z retirado: /usr/lib/google-cloud-sdk/platform/bundledpythonunix/lib/pkgconfig/python3.pc
evals	Retirar Python del runner	2026-09-15T11:13:09.2201999Z retirado: /usr/lib/google-cloud-sdk/platform/bundledpythonunix/lib/pkgconfig/python3-embed.pc
evals	Retirar Python del runner	2026-09-15T11:13:09.2390381Z retirado: /usr/lib/google-cloud-sdk/platform/bundledpythonunix/lib/python3.14/config-3.14-x86_64-linux-gnu/python-config.py
evals	Retirar Python del runner	2026-09-15T11:13:09.2580477Z retirado: /usr/lib/google-cloud-sdk/platform/bundledpythonunix/lib/libpython3.14.so
evals	Retirar Python del runner	2026-09-15T11:13:09.2767636Z retirado: /usr/lib/google-cloud-sdk/platform/bundledpythonunix/lib/libpython3.14.so.1.0
evals	Retirar Python del runner	2026-09-15T11:13:09.3141972Z retirado: /usr/lib/google-cloud-sdk/platform/bundledpythonunix
evals	Retirar Python del runner	2026-09-15T11:13:09.5632630Z retirado: /usr/lib/rpm/pythondistdeps.py
evals	Retirar Python del runner	2026-09-15T11:13:09.5828160Z retirado: /usr/local/aws-cli/v2/2.36.40/dist/libpython3.14.so.1.0
evals	Retirar Python del runner	2026-09-15T11:13:09.6021415Z retirado: /usr/local/lib/android/sdk/ndk/28.2.13676358/toolchains/llvm/prebuilt/linux-x86_64/python3/lib/libpython3.so
evals	Retirar Python del runner	2026-09-15T11:13:09.6216831Z retirado: /usr/local/lib/android/sdk/ndk/28.2.13676358/toolchains/llvm/prebuilt/linux-x86_64/python3/lib/pkgconfig/python3.pc
evals	Retirar Python del runner	2026-09-15T11:13:09.6406566Z retirado: /usr/local/lib/android/sdk/ndk/28.2.13676358/toolchains/llvm/prebuilt/linux-x86_64/python3/lib/pkgconfig/python3-embed.pc
evals	Retirar Python del runner	2026-09-15T11:13:09.6610265Z retirado: /usr/local/lib/android/sdk/ndk/28.2.13676358/toolchains/llvm/prebuilt/linux-x86_64/python3/lib/pkgconfig/python-3.11-embed.pc
evals	Retirar Python del runner	2026-09-15T11:13:09.6801871Z retirado: /usr/local/lib/android/sdk/ndk/28.2.13676358/toolchains/llvm/prebuilt/linux-x86_64/python3/lib/pkgconfig/python-3.11.pc
evals	Retirar Python del runner	2026-09-15T11:13:09.6997415Z retirado: /usr/local/lib/android/sdk/ndk/28.2.13676358/toolchains/llvm/prebuilt/linux-x86_64/python3/lib/python3.11/site-packages/pip/_vendor/pygments/lexers/python.py
evals	Retirar Python del runner	2026-09-15T11:13:09.7185165Z retirado: /usr/local/lib/android/sdk/ndk/28.2.13676358/toolchains/llvm/prebuilt/linux-x86_64/python3/lib/libpython3.11.so
evals	Retirar Python del runner	2026-09-15T11:13:09.7370930Z retirado: /usr/local/lib/android/sdk/ndk/28.2.13676358/toolchains/llvm/prebuilt/linux-x86_64/python3/lib/libpython3.11.so.1.0
evals	Retirar Python del runner	2026-09-15T11:13:09.7566210Z retirado: /usr/local/lib/android/sdk/ndk/28.2.13676358/toolchains/llvm/prebuilt/linux-x86_64/python3/include/python3.11/cpython/pythonrun.h
evals	Retirar Python del runner	2026-09-15T11:13:09.7755904Z retirado: /usr/local/lib/android/sdk/ndk/28.2.13676358/toolchains/llvm/prebuilt/linux-x86_64/python3/include/python3.11/Python.h
evals	Retirar Python del runner	2026-09-15T11:13:09.7944214Z retirado: /usr/local/lib/android/sdk/ndk/28.2.13676358/toolchains/llvm/prebuilt/linux-x86_64/python3/include/python3.11/pythonrun.h
evals	Retirar Python del runner	2026-09-15T11:13:09.8138181Z retirado: /usr/local/lib/android/sdk/ndk/28.2.13676358/toolchains/llvm/prebuilt/linux-x86_64/python3/share/man/man1/python3.1
evals	Retirar Python del runner	2026-09-15T11:13:09.8334392Z retirado: /usr/local/lib/android/sdk/ndk/28.2.13676358/toolchains/llvm/prebuilt/linux-x86_64/python3/share/man/man1/python3.11.1
evals	Retirar Python del runner	2026-09-15T11:13:09.8709764Z retirado: /usr/local/lib/android/sdk/ndk/28.2.13676358/toolchains/llvm/prebuilt/linux-x86_64/python3
evals	Retirar Python del runner	2026-09-15T11:13:09.9483969Z retirado: /usr/local/lib/android/sdk/ndk/27.3.13750724/toolchains/llvm/prebuilt/linux-x86_64/python3/lib/libpython3.so
evals	Retirar Python del runner	2026-09-15T11:13:09.9675562Z retirado: /usr/local/lib/android/sdk/ndk/27.3.13750724/toolchains/llvm/prebuilt/linux-x86_64/python3/lib/pkgconfig/python3.pc
evals	Retirar Python del runner	2026-09-15T11:13:09.9863624Z retirado: /usr/local/lib/android/sdk/ndk/27.3.13750724/toolchains/llvm/prebuilt/linux-x86_64/python3/lib/pkgconfig/python3-embed.pc
evals	Retirar Python del runner	2026-09-15T11:13:10.0059724Z retirado: /usr/local/lib/android/sdk/ndk/27.3.13750724/toolchains/llvm/prebuilt/linux-x86_64/python3/lib/pkgconfig/python-3.11-embed.pc
evals	Retirar Python del runner	2026-09-15T11:13:10.0244473Z retirado: /usr/local/lib/android/sdk/ndk/27.3.13750724/toolchains/llvm/prebuilt/linux-x86_64/python3/lib/pkgconfig/python-3.11.pc
evals	Retirar Python del runner	2026-09-15T11:13:10.0431609Z retirado: /usr/local/lib/android/sdk/ndk/27.3.13750724/toolchains/llvm/prebuilt/linux-x86_64/python3/lib/python3.11/site-packages/pip/_vendor/pygments/lexers/python.py
evals	Retirar Python del runner	2026-09-15T11:13:10.0620943Z retirado: /usr/local/lib/android/sdk/ndk/27.3.13750724/toolchains/llvm/prebuilt/linux-x86_64/python3/lib/libpython3.11.so
evals	Retirar Python del runner	2026-09-15T11:13:10.0808542Z retirado: /usr/local/lib/android/sdk/ndk/27.3.13750724/toolchains/llvm/prebuilt/linux-x86_64/python3/lib/libpython3.11.so.1.0
evals	Retirar Python del runner	2026-09-15T11:13:10.0996032Z retirado: /usr/local/lib/android/sdk/ndk/27.3.13750724/toolchains/llvm/prebuilt/linux-x86_64/python3/include/python3.11/cpython/pythonrun.h
evals	Retirar Python del runner	2026-09-15T11:13:10.1181921Z retirado: /usr/local/lib/android/sdk/ndk/27.3.13750724/toolchains/llvm/prebuilt/linux-x86_64/python3/include/python3.11/Python.h
evals	Retirar Python del runner	2026-09-15T11:13:10.1370538Z retirado: /usr/local/lib/android/sdk/ndk/27.3.13750724/toolchains/llvm/prebuilt/linux-x86_64/python3/include/python3.11/pythonrun.h
evals	Retirar Python del runner	2026-09-15T11:13:10.1558082Z retirado: /usr/local/lib/android/sdk/ndk/27.3.13750724/toolchains/llvm/prebuilt/linux-x86_64/python3/share/man/man1/python3.1
evals	Retirar Python del runner	2026-09-15T11:13:10.1744568Z retirado: /usr/local/lib/android/sdk/ndk/27.3.13750724/toolchains/llvm/prebuilt/linux-x86_64/python3/share/man/man1/python3.11.1
evals	Retirar Python del runner	2026-09-15T11:13:10.2123701Z retirado: /usr/local/lib/android/sdk/ndk/27.3.13750724/toolchains/llvm/prebuilt/linux-x86_64/python3
evals	Retirar Python del runner	2026-09-15T11:13:10.2893893Z retirado: /usr/local/lib/android/sdk/ndk/29.0.14206865/toolchains/llvm/prebuilt/linux-x86_64/python3/lib/libpython3.so
evals	Retirar Python del runner	2026-09-15T11:13:10.3094527Z retirado: /usr/local/lib/android/sdk/ndk/29.0.14206865/toolchains/llvm/prebuilt/linux-x86_64/python3/lib/pkgconfig/python3.pc
evals	Retirar Python del runner	2026-09-15T11:13:10.3291057Z retirado: /usr/local/lib/android/sdk/ndk/29.0.14206865/toolchains/llvm/prebuilt/linux-x86_64/python3/lib/pkgconfig/python3-embed.pc
evals	Retirar Python del runner	2026-09-15T11:13:10.3483528Z retirado: /usr/local/lib/android/sdk/ndk/29.0.14206865/toolchains/llvm/prebuilt/linux-x86_64/python3/lib/pkgconfig/python-3.11-embed.pc
evals	Retirar Python del runner	2026-09-15T11:13:10.3679785Z retirado: /usr/local/lib/android/sdk/ndk/29.0.14206865/toolchains/llvm/prebuilt/linux-x86_64/python3/lib/pkgconfig/python-3.11.pc
evals	Retirar Python del runner	2026-09-15T11:13:10.3866664Z retirado: /usr/local/lib/android/sdk/ndk/29.0.14206865/toolchains/llvm/prebuilt/linux-x86_64/python3/lib/python3.11/site-packages/pip/_vendor/pygments/lexers/python.py
evals	Retirar Python del runner	2026-09-15T11:13:10.4054659Z retirado: /usr/local/lib/android/sdk/ndk/29.0.14206865/toolchains/llvm/prebuilt/linux-x86_64/python3/lib/libpython3.11.so
evals	Retirar Python del runner	2026-09-15T11:13:10.4257713Z retirado: /usr/local/lib/android/sdk/ndk/29.0.14206865/toolchains/llvm/prebuilt/linux-x86_64/python3/lib/libpython3.11.so.1.0
evals	Retirar Python del runner	2026-09-15T11:13:10.4456104Z retirado: /usr/local/lib/android/sdk/ndk/29.0.14206865/toolchains/llvm/prebuilt/linux-x86_64/python3/include/python3.11/cpython/pythonrun.h
evals	Retirar Python del runner	2026-09-15T11:13:10.4646887Z retirado: /usr/local/lib/android/sdk/ndk/29.0.14206865/toolchains/llvm/prebuilt/linux-x86_64/python3/include/python3.11/Python.h
evals	Retirar Python del runner	2026-09-15T11:13:10.4835244Z retirado: /usr/local/lib/android/sdk/ndk/29.0.14206865/toolchains/llvm/prebuilt/linux-x86_64/python3/include/python3.11/pythonrun.h
evals	Retirar Python del runner	2026-09-15T11:13:10.5026797Z retirado: /usr/local/lib/android/sdk/ndk/29.0.14206865/toolchains/llvm/prebuilt/linux-x86_64/python3/share/man/man1/python3.1
evals	Retirar Python del runner	2026-09-15T11:13:10.5229373Z retirado: /usr/local/lib/android/sdk/ndk/29.0.14206865/toolchains/llvm/prebuilt/linux-x86_64/python3/share/man/man1/python3.11.1
evals	Retirar Python del runner	2026-09-15T11:13:10.5594471Z retirado: /usr/local/lib/android/sdk/ndk/29.0.14206865/toolchains/llvm/prebuilt/linux-x86_64/python3
evals	Retirar Python del runner	2026-09-15T11:13:10.6361720Z retirado: /usr/local/aws-sam-cli/1.166.1/dist/_internal/libpython3.11.so.1.0
evals	Retirar Python del runner	2026-09-15T11:13:10.6566302Z retirado: /usr/local/share/vcpkg/ports/libudis86/python3.patch
evals	Retirar Python del runner	2026-09-15T11:13:10.6760994Z retirado: /usr/local/share/vcpkg/ports/omniorb/python-fixes.patch
evals	Retirar Python del runner	2026-09-15T11:13:10.6960697Z retirado: /usr/local/share/vcpkg/ports/openxr-loader/python3_8_compatibility.patch
evals	Retirar Python del runner	2026-09-15T11:13:10.7152811Z retirado: /usr/local/share/vcpkg/ports/libxslt/python3.patch
evals	Retirar Python del runner	2026-09-15T11:13:10.7338419Z retirado: /usr/local/share/vcpkg/ports/openscap/python-win32.diff
evals	Retirar Python del runner	2026-09-15T11:13:10.7524616Z retirado: /usr/local/share/vcpkg/ports/python3/python_vcpkg.props.in
evals	Retirar Python del runner	2026-09-15T11:13:10.7714374Z retirado: /usr/local/share/vcpkg/ports/vtk/pythonwrapper.patch
evals	Retirar Python del runner	2026-09-15T11:13:10.7911984Z retirado: /usr/local/share/vcpkg/scripts/test_ports/vcpkg-ci-blender/python.patch
evals	Retirar Python del runner	2026-09-15T11:13:10.8099269Z retirado: /usr/local/share/vcpkg/versions/p-/python2.json
evals	Retirar Python del runner	2026-09-15T11:13:10.8280606Z retirado: /usr/local/share/vcpkg/versions/p-/python3.json
evals	Retirar Python del runner	2026-09-15T11:13:10.8470379Z retirado: /usr/share/perl5/NeedRestart/Interp/Python.pm
evals	Retirar Python del runner	2026-09-15T11:13:10.8660573Z retirado: /usr/share/doc-base/python3.python-policy
evals	Retirar Python del runner	2026-09-15T11:13:10.8848724Z retirado: /usr/share/az_15.6.1/Az.Functions/4.3.2/Functions.Autorest/custom/FunctionsStackFlexData/EastAsia/python.json
evals	Retirar Python del runner	2026-09-15T11:13:10.9047102Z retirado: /usr/share/bash-completion/completions/python3.9
evals	Retirar Python del runner	2026-09-15T11:13:10.9242701Z retirado: /usr/share/bash-completion/completions/python3.7
evals	Retirar Python del runner	2026-09-15T11:13:10.9441717Z retirado: /usr/share/bash-completion/completions/python3.3
evals	Retirar Python del runner	2026-09-15T11:13:10.9624446Z retirado: /usr/share/bash-completion/completions/python3.6
evals	Retirar Python del runner	2026-09-15T11:13:10.9818910Z retirado: /usr/share/bash-completion/completions/python2.7
evals	Retirar Python del runner	2026-09-15T11:13:11.0016280Z retirado: /usr/share/bash-completion/completions/python3.4
evals	Retirar Python del runner	2026-09-15T11:13:11.0201830Z retirado: /usr/share/bash-completion/completions/pypy3
evals	Retirar Python del runner	2026-09-15T11:13:11.0389584Z retirado: /usr/share/bash-completion/completions/python2
evals	Retirar Python del runner	2026-09-15T11:13:11.0575528Z retirado: /usr/share/bash-completion/completions/pypy
evals	Retirar Python del runner	2026-09-15T11:13:11.0765162Z retirado: /usr/share/bash-completion/completions/python3.8
evals	Retirar Python del runner	2026-09-15T11:13:11.0947811Z retirado: /usr/share/bash-completion/completions/python3.5
evals	Retirar Python del runner	2026-09-15T11:13:11.1131767Z retirado: /usr/share/bash-completion/completions/python3
evals	Retirar Python del runner	2026-09-15T11:13:11.1318999Z retirado: /usr/share/bash-completion/completions/python
evals	Retirar Python del runner	2026-09-15T11:13:11.1509486Z retirado: /usr/share/bash-completion/helpers/python
evals	Retirar Python del runner	2026-09-15T11:13:11.1704571Z retirado: /usr/share/man/man8/pythoncalls-bpfcc.8.gz
evals	Retirar Python del runner	2026-09-15T11:13:11.1886577Z retirado: /usr/share/man/man8/pythonstat-bpfcc.8.gz
evals	Retirar Python del runner	2026-09-15T11:13:11.2068706Z retirado: /usr/share/man/man8/pythonflow-bpfcc.8.gz
evals	Retirar Python del runner	2026-09-15T11:13:11.2251124Z retirado: /usr/share/man/man8/pythongc-bpfcc.8.gz
evals	Retirar Python del runner	2026-09-15T11:13:11.2430509Z retirado: /usr/share/man/man1/python3.12.1.gz
evals	Retirar Python del runner	2026-09-15T11:13:11.2679247Z retirado: /usr/share/man/man1/python.1.gz
evals	Retirar Python del runner	2026-09-15T11:13:11.2856165Z retirado: /usr/share/man/man1/python3.12-config.1.gz
evals	Retirar Python del runner	2026-09-15T11:13:11.3129977Z retirado: /usr/share/man/man1/python3.1.gz
evals	Retirar Python del runner	2026-09-15T11:13:11.3387977Z retirado: /usr/share/man/man1/python3-config.1.gz
evals	Retirar Python del runner	2026-09-15T11:13:11.3566113Z retirado: /usr/share/pixmaps/python3.xpm
evals	Retirar Python del runner	2026-09-15T11:13:11.3746106Z retirado: /usr/share/pixmaps/python3.12.xpm
evals	Retirar Python del runner	2026-09-15T11:13:11.3926176Z retirado: /usr/share/binfmts/python3.12
evals	Retirar Python del runner	2026-09-15T11:13:11.4114076Z retirado: /usr/share/vim/vim91/syntax/python2.vim
evals	Retirar Python del runner	2026-09-15T11:13:11.4302214Z retirado: /usr/share/vim/vim91/syntax/python.vim
evals	Retirar Python del runner	2026-09-15T11:13:11.4487342Z retirado: /usr/share/vim/vim91/autoload/pythoncomplete.vim
evals	Retirar Python del runner	2026-09-15T11:13:11.4669701Z retirado: /usr/share/vim/vim91/autoload/python3complete.vim
evals	Retirar Python del runner	2026-09-15T11:13:11.4853992Z retirado: /usr/share/vim/vim91/autoload/python.vim
evals	Retirar Python del runner	2026-09-15T11:13:11.5038904Z retirado: /usr/share/vim/vim91/ftplugin/python.vim
evals	Retirar Python del runner	2026-09-15T11:13:11.5218480Z retirado: /usr/share/vim/vim91/indent/python.vim
evals	Retirar Python del runner	2026-09-15T11:13:11.5403753Z retirado: /usr/share/swig4.0/python/pythonkw.swg
evals	Retirar Python del runner	2026-09-15T11:13:11.5583701Z retirado: /usr/share/swig4.0/python/python.swg
evals	Retirar Python del runner	2026-09-15T11:13:11.5763724Z retirado: /usr/share/applications/python3.12.desktop
evals	Retirar Python del runner	2026-09-15T11:13:11.5960728Z retirado: /usr/share/doc/python3.12-venv
evals	Retirar Python del runner	2026-09-15T11:13:11.6142853Z retirado: /usr/share/doc/libpython3.12t64
evals	Retirar Python del runner	2026-09-15T11:13:11.6331123Z retirado: /usr/share/doc/python3-setuptools/python 2 sunset.rst
evals	Retirar Python del runner	2026-09-15T11:13:11.6516085Z retirado: /usr/share/doc/libpython3.12-dev
evals	Retirar Python del runner	2026-09-15T11:13:11.6699410Z retirado: /usr/share/doc/mercurial-common/examples/python-hook-examples.py
evals	Retirar Python del runner	2026-09-15T11:13:11.6877680Z retirado: /usr/share/doc/python3.12-dev
evals	Retirar Python del runner	2026-09-15T11:13:11.7058216Z retirado: /usr/share/doc/python3-pip/html/topics/python-option.md
evals	Retirar Python del runner	2026-09-15T11:13:11.7245127Z retirado: /usr/share/doc/python3-venv
evals	Retirar Python del runner	2026-09-15T11:13:11.7430776Z retirado: /usr/share/doc/python3.12/python-policy.txt.gz
evals	Retirar Python del runner	2026-09-15T11:13:11.7615847Z retirado: /usr/share/doc/libpython3.12-stdlib
evals	Retirar Python del runner	2026-09-15T11:13:11.7800980Z retirado: /usr/share/doc/python3-dev
evals	Retirar Python del runner	2026-09-15T11:13:11.7987560Z retirado: /usr/share/doc/bpfcc-tools/examples/doc/pythonstat_example.txt
evals	Retirar Python del runner	2026-09-15T11:13:11.8172251Z retirado: /usr/share/doc/bpfcc-tools/examples/doc/pythonflow_example.txt
evals	Retirar Python del runner	2026-09-15T11:13:11.8365977Z retirado: /usr/share/doc/bpfcc-tools/examples/doc/pythoncalls_example.txt
evals	Retirar Python del runner	2026-09-15T11:13:11.8568934Z retirado: /usr/share/doc/bpfcc-tools/examples/doc/pythongc_example.txt
evals	Retirar Python del runner	2026-09-15T11:13:11.8762378Z retirado: /usr/share/doc/python3-debconf
evals	Retirar Python del runner	2026-09-15T11:13:11.8949313Z retirado: /usr/share/doc/python3/python-policy.txt.gz
evals	Retirar Python del runner	2026-09-15T11:13:11.9134748Z retirado: /usr/share/doc/python3/python-policy.html
evals	Retirar Python del runner	2026-09-15T11:13:11.9336033Z retirado: /usr/share/aclocal-1.16/python.m4
evals	Retirar Python del runner	2026-09-15T11:13:11.9524103Z retirado: /usr/share/miniconda/lib/libpython3.so
evals	Retirar Python del runner	2026-09-15T11:13:11.9709780Z retirado: /usr/share/miniconda/lib/pkgconfig/python-3.14-embed.pc
evals	Retirar Python del runner	2026-09-15T11:13:11.9890938Z retirado: /usr/share/miniconda/lib/pkgconfig/python3.pc
evals	Retirar Python del runner	2026-09-15T11:13:12.0153636Z retirado: /usr/share/miniconda/lib/pkgconfig/python3-embed.pc
evals	Retirar Python del runner	2026-09-15T11:13:12.0336312Z retirado: /usr/share/miniconda/lib/pkgconfig/python-3.14.pc
evals	Retirar Python del runner	2026-09-15T11:13:12.0518477Z retirado: /usr/share/miniconda/lib/python3.14/site-packages/conda/common/path/__pycache__/python.cpython-314.pyc
evals	Retirar Python del runner	2026-09-15T11:13:12.0718681Z retirado: /usr/share/miniconda/lib/python3.14/site-packages/conda/common/path/python.py
evals	Retirar Python del runner	2026-09-15T11:13:12.0906695Z retirado: /usr/share/miniconda/lib/python3.14/site-packages/pygments/lexers/__pycache__/python.cpython-314.pyc
evals	Retirar Python del runner	2026-09-15T11:13:12.1108446Z retirado: /usr/share/miniconda/lib/python3.14/site-packages/pygments/lexers/python.py
evals	Retirar Python del runner	2026-09-15T11:13:12.1304217Z retirado: /usr/share/miniconda/lib/python3.14/site-packages/pip/_vendor/pygments/lexers/__pycache__/python.cpython-314.pyc
evals	Retirar Python del runner	2026-09-15T11:13:12.1500910Z retirado: /usr/share/miniconda/lib/python3.14/site-packages/pip/_vendor/pygments/lexers/python.py
evals	Retirar Python del runner	2026-09-15T11:13:12.1687461Z retirado: /usr/share/miniconda/lib/python3.14/site-packages/conda_pypi/__pycache__/python_paths.cpython-314.pyc
evals	Retirar Python del runner	2026-09-15T11:13:12.1874115Z retirado: /usr/share/miniconda/lib/python3.14/site-packages/conda_pypi/python_paths.py
evals	Retirar Python del runner	2026-09-15T11:13:12.2058687Z retirado: /usr/share/miniconda/lib/python3.14/config-3.14-x86_64-linux-gnu/__pycache__/python-config.cpython-314.pyc
evals	Retirar Python del runner	2026-09-15T11:13:12.2243956Z retirado: /usr/share/miniconda/lib/python3.14/config-3.14-x86_64-linux-gnu/python-config.py
evals	Retirar Python del runner	2026-09-15T11:13:12.2436340Z retirado: /usr/share/miniconda/lib/python3.14/config-3.14-x86_64-linux-gnu/python.o
evals	Retirar Python del runner	2026-09-15T11:13:12.2623974Z retirado: /usr/share/miniconda/lib/python3.14/idlelib/Icons/python.gif
evals	Retirar Python del runner	2026-09-15T11:13:12.2811500Z retirado: /usr/share/miniconda/lib/libpython3.14.so
evals	Retirar Python del runner	2026-09-15T11:13:12.2998109Z retirado: /usr/share/miniconda/lib/libpython3.14.so.1.0
evals	Retirar Python del runner	2026-09-15T11:13:12.3179895Z retirado: /usr/share/miniconda/include/python3.14/cpython/pythonrun.h
evals	Retirar Python del runner	2026-09-15T11:13:12.3362738Z retirado: /usr/share/miniconda/include/python3.14/Python.h
evals	Retirar Python del runner	2026-09-15T11:13:12.3548842Z retirado: /usr/share/miniconda/include/python3.14/pythonrun.h
evals	Retirar Python del runner	2026-09-15T11:13:12.3736197Z retirado: /usr/share/miniconda/share/man/man1/python3.1
evals	Retirar Python del runner	2026-09-15T11:13:12.3924393Z retirado: /usr/share/miniconda/share/man/man1/python3.14.1
evals	Retirar Python del runner	2026-09-15T11:13:12.4115725Z retirado: /usr/share/miniconda/conda-meta/python_abi-3.14-4_cp314.json
evals	Retirar Python del runner	2026-09-15T11:13:12.4309236Z retirado: /usr/share/miniconda/conda-meta/python-installer-1.0.1-py314h06a4308_0.json
evals	Retirar Python del runner	2026-09-15T11:13:12.4498060Z retirado: /usr/share/miniconda/conda-meta/python-build-1.5.1-py314h06a4308_0.json
evals	Retirar Python del runner	2026-09-15T11:13:12.4685665Z retirado: /usr/share/miniconda/conda-meta/python-3.14.7-h2bd7c14_101_cp314.json
evals	Retirar Python del runner	2026-09-15T11:13:12.4877679Z retirado: /usr/share/miniconda/conda-meta/python-dotenv-1.2.2-py314h06a4308_0.json
evals	Retirar Python del runner	2026-09-15T11:13:12.5062564Z retirado: /usr/share/miniconda/pkgs/python-build-1.5.1-py314h06a4308_0.conda
evals	Retirar Python del runner	2026-09-15T11:13:12.5426940Z retirado: /usr/share/miniconda/pkgs/python-build-1.5.1-py314h06a4308_0
evals	Retirar Python del runner	2026-09-15T11:13:12.5655571Z retirado: /usr/share/miniconda/pkgs/conda-26.7.1-py314h06a4308_0/lib/python3.14/site-packages/conda/common/path/__pycache__/python.cpython-314.pyc
evals	Retirar Python del runner	2026-09-15T11:13:12.5847165Z retirado: /usr/share/miniconda/pkgs/conda-26.7.1-py314h06a4308_0/lib/python3.14/site-packages/conda/common/path/python.py
evals	Retirar Python del runner	2026-09-15T11:13:12.6034789Z retirado: /usr/share/miniconda/pkgs/python-3.14.7-h2bd7c14_101_cp314.conda
evals	Retirar Python del runner	2026-09-15T11:13:12.6229165Z retirado: /usr/share/miniconda/pkgs/python-dotenv-1.2.2-py314h06a4308_0.conda
evals	Retirar Python del runner	2026-09-15T11:13:12.6417402Z retirado: /usr/share/miniconda/pkgs/libxcb-1.17.0-h9b100fa_0/info/recipe/python3.patch
evals	Retirar Python del runner	2026-09-15T11:13:12.6602916Z retirado: /usr/share/miniconda/pkgs/pip-26.2.1-pyh0d26453_0/site-packages/pip/_vendor/pygments/lexers/python.py
evals	Retirar Python del runner	2026-09-15T11:13:12.6792371Z retirado: /usr/share/miniconda/pkgs/python-3.14.7-h2bd7c14_101_cp314/lib/libpython3.so
evals	Retirar Python del runner	2026-09-15T11:13:12.6980986Z retirado: /usr/share/miniconda/pkgs/python-3.14.7-h2bd7c14_101_cp314/lib/pkgconfig/python-3.14-embed.pc
evals	Retirar Python del runner	2026-09-15T11:13:12.7174788Z retirado: /usr/share/miniconda/pkgs/python-3.14.7-h2bd7c14_101_cp314/lib/pkgconfig/python3.pc
evals	Retirar Python del runner	2026-09-15T11:13:12.7450198Z retirado: /usr/share/miniconda/pkgs/python-3.14.7-h2bd7c14_101_cp314/lib/pkgconfig/python3-embed.pc
evals	Retirar Python del runner	2026-09-15T11:13:12.7651272Z retirado: /usr/share/miniconda/pkgs/python-3.14.7-h2bd7c14_101_cp314/lib/pkgconfig/python-3.14.pc
evals	Retirar Python del runner	2026-09-15T11:13:12.7839894Z retirado: /usr/share/miniconda/pkgs/python-3.14.7-h2bd7c14_101_cp314/lib/python3.14/config-3.14-x86_64-linux-gnu/__pycache__/python-config.cpython-314.pyc
evals	Retirar Python del runner	2026-09-15T11:13:12.8035229Z retirado: /usr/share/miniconda/pkgs/python-3.14.7-h2bd7c14_101_cp314/lib/python3.14/config-3.14-x86_64-linux-gnu/python-config.py
evals	Retirar Python del runner	2026-09-15T11:13:12.8233464Z retirado: /usr/share/miniconda/pkgs/python-3.14.7-h2bd7c14_101_cp314/lib/python3.14/config-3.14-x86_64-linux-gnu/python.o
evals	Retirar Python del runner	2026-09-15T11:13:12.8419665Z retirado: /usr/share/miniconda/pkgs/python-3.14.7-h2bd7c14_101_cp314/lib/python3.14/idlelib/Icons/python.gif
evals	Retirar Python del runner	2026-09-15T11:13:12.8607678Z retirado: /usr/share/miniconda/pkgs/python-3.14.7-h2bd7c14_101_cp314/lib/libpython3.14.so
evals	Retirar Python del runner	2026-09-15T11:13:12.8790314Z retirado: /usr/share/miniconda/pkgs/python-3.14.7-h2bd7c14_101_cp314/lib/libpython3.14.so.1.0
evals	Retirar Python del runner	2026-09-15T11:13:12.8981339Z retirado: /usr/share/miniconda/pkgs/python-3.14.7-h2bd7c14_101_cp314/include/python3.14/cpython/pythonrun.h
evals	Retirar Python del runner	2026-09-15T11:13:12.9175679Z retirado: /usr/share/miniconda/pkgs/python-3.14.7-h2bd7c14_101_cp314/include/python3.14/Python.h
evals	Retirar Python del runner	2026-09-15T11:13:12.9381611Z retirado: /usr/share/miniconda/pkgs/python-3.14.7-h2bd7c14_101_cp314/include/python3.14/pythonrun.h
evals	Retirar Python del runner	2026-09-15T11:13:12.9579592Z retirado: /usr/share/miniconda/pkgs/python-3.14.7-h2bd7c14_101_cp314/share/man/man1/python3.1
evals	Retirar Python del runner	2026-09-15T11:13:12.9769480Z retirado: /usr/share/miniconda/pkgs/python-3.14.7-h2bd7c14_101_cp314/share/man/man1/python3.14.1
evals	Retirar Python del runner	2026-09-15T11:13:13.0151382Z retirado: /usr/share/miniconda/pkgs/python-3.14.7-h2bd7c14_101_cp314
evals	Retirar Python del runner	2026-09-15T11:13:13.1461990Z retirado: /usr/share/miniconda/pkgs/pygments-2.20.0-py314h06a4308_0/lib/python3.14/site-packages/pygments/lexers/__pycache__/python.cpython-314.pyc
evals	Retirar Python del runner	2026-09-15T11:13:13.1649892Z retirado: /usr/share/miniconda/pkgs/pygments-2.20.0-py314h06a4308_0/lib/python3.14/site-packages/pygments/lexers/python.py
evals	Retirar Python del runner	2026-09-15T11:13:13.1836304Z retirado: /usr/share/miniconda/pkgs/pygments-2.20.0-py314h06a4308_0/info/test/tests/support/python_lexer.py
evals	Retirar Python del runner	2026-09-15T11:13:13.2020483Z retirado: /usr/share/miniconda/pkgs/pygments-2.20.0-py314h06a4308_0/info/test/tests/examplefiles/make/python25-bsd.mak.output
evals	Retirar Python del runner	2026-09-15T11:13:13.2206388Z retirado: /usr/share/miniconda/pkgs/pygments-2.20.0-py314h06a4308_0/info/test/tests/examplefiles/make/python25-bsd.mak
evals	Retirar Python del runner	2026-09-15T11:13:13.2394034Z retirado: /usr/share/miniconda/pkgs/conda-pypi-0.11.0-py314h06a4308_0/lib/python3.14/site-packages/conda_pypi/__pycache__/python_paths.cpython-314.pyc
evals	Retirar Python del runner	2026-09-15T11:13:13.2581659Z retirado: /usr/share/miniconda/pkgs/conda-pypi-0.11.0-py314h06a4308_0/lib/python3.14/site-packages/conda_pypi/python_paths.py
evals	Retirar Python del runner	2026-09-15T11:13:13.2770593Z retirado: /usr/share/miniconda/pkgs/python-installer-1.0.1-py314h06a4308_0.conda
evals	Retirar Python del runner	2026-09-15T11:13:13.2958962Z retirado: /usr/share/miniconda/pkgs/python_abi-3.14-4_cp314.conda
evals	Retirar Python del runner	2026-09-15T11:13:13.3333604Z retirado: /usr/share/miniconda
evals	Retirar Python del runner	2026-09-15T11:13:14.4744206Z retirado: /usr/share/nano/python.nanorc
evals	Retirar Python del runner	2026-09-15T11:13:14.4947902Z retirado: /usr/share/lintian/overrides/python3-debian
evals	Retirar Python del runner	2026-09-15T11:13:14.5144942Z retirado: /usr/share/lintian/overrides/python3.12-venv
evals	Retirar Python del runner	2026-09-15T11:13:14.5347124Z retirado: /usr/share/lintian/overrides/python3-dbus
evals	Retirar Python del runner	2026-09-15T11:13:14.5550862Z retirado: /usr/share/lintian/overrides/libpython3.12t64
evals	Retirar Python del runner	2026-09-15T11:13:14.5747849Z retirado: /usr/share/lintian/overrides/libpython3.12-dev
evals	Retirar Python del runner	2026-09-15T11:13:14.5941446Z retirado: /usr/share/lintian/overrides/python3-pip
evals	Retirar Python del runner	2026-09-15T11:13:14.6123940Z retirado: /usr/share/lintian/overrides/libpython3.12-minimal
evals	Retirar Python del runner	2026-09-15T11:13:14.6308371Z retirado: /usr/share/lintian/overrides/python3.12-minimal
evals	Retirar Python del runner	2026-09-15T11:13:14.6489524Z retirado: /usr/share/lintian/overrides/python3.12
evals	Retirar Python del runner	2026-09-15T11:13:14.6673549Z retirado: /usr/share/lintian/overrides/libpython3.12-stdlib
evals	Retirar Python del runner	2026-09-15T11:13:14.6857575Z retirado: /usr/share/lintian/overrides/python3-netaddr
evals	Retirar Python del runner	2026-09-15T11:13:14.7038872Z retirado: /usr/share/lintian/overrides/python3
evals	Retirar Python del runner	2026-09-15T11:13:14.7231704Z retirado: /usr/share/lintian/overrides/python3-apt
evals	Retirar Python del runner	2026-09-15T11:13:14.7421520Z retirado: /usr/share/automake-1.16/am/python.am
evals	Retirar Python del runner	2026-09-15T11:13:14.7622112Z retirado: /usr/share/python3/bcep/python3-jinja2
evals	Retirar Python del runner	2026-09-15T11:13:14.7808115Z retirado: /usr/share/python3/dist/python3-cryptography
evals	Retirar Python del runner	2026-09-15T11:13:14.7996030Z retirado: /usr/share/python3/dist/python3-zope.interface
evals	Retirar Python del runner	2026-09-15T11:13:14.8181631Z retirado: /usr/share/python3/dist/python3-six
evals	Retirar Python del runner	2026-09-15T11:13:14.8368476Z retirado: /usr/share/python3/dist/python3-pyasn1
evals	Retirar Python del runner	2026-09-15T11:13:14.8554329Z retirado: /usr/share/python3/python.mk
evals	Retirar Python del runner	2026-09-15T11:13:14.8743433Z retirado: /usr/sbin/pythongc-bpfcc
evals	Retirar Python del runner	2026-09-15T11:13:14.8928239Z retirado: /usr/sbin/pythoncalls-bpfcc
evals	Retirar Python del runner	2026-09-15T11:13:14.9119129Z retirado: /usr/sbin/pythonstat-bpfcc
evals	Retirar Python del runner	2026-09-15T11:13:14.9314671Z retirado: /usr/sbin/pythonflow-bpfcc
evals	Retirar Python del runner	2026-09-15T11:13:14.9681686Z retirado: /usr/bin/python3.12-config
evals	Retirar Python del runner	2026-09-15T11:13:15.0135071Z retirado: /usr/bin/python3-config
evals	Retirar Python del runner	2026-09-15T11:13:15.0499010Z retirado: /usr/bin/python3.12
evals	Retirar Python del runner	2026-09-15T11:13:15.0947196Z retirado: /usr/bin/python3
evals	Retirar Python del runner	2026-09-15T11:13:15.1410659Z retirado: /usr/bin/python
evals	Retirar Python del runner	2026-09-15T11:13:19.7005171Z búsqueda tras retirar: ninguno
evals	Retirar Python del runner	2026-09-15T11:13:19.7008525Z --- fin de la retirada de Python ---
código de la sexta orden: 0
`````
