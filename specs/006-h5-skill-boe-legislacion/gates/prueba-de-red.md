# Prueba de red de H5 (quickstart §12.2, SC-012)

Intento 7 de T030 (el tercero que cuenta el workflow, `gates/tareas-intentos.json` `T030: 3`, concedido por la
supervisión del run tras T044, T045 y T046), 2026-09-15, propuesta de cambio [#27](https://github.com/jmorenobl/kitlegal/pull/27),
cabeza `091facd82d3744dffcfb3b17115c5ea25452d7d5` (`feat(H5): T046`). **Resultado: veredicto `aprobado`, por primera
vez** —las trece sesiones terminadas con código 0 y las trece trazas leídas enteras, las diez positivas y las dos de
no activación en verde, «ninguna petición llegó a la red de una fuente»—, y **la prueba de red cumple todo lo que
SC-012 espera de ella**: las dos filas de `a9998` en «fuera de lo grabado», con código 5 la que va sin `--offline` y 4
la que lo lleva; la invocación sin `--offline` con una sola conexión, `127.0.0.1:9` de clase `local`, y la que lleva
`--offline` sin ninguna; ninguna de las dos cuenta como ejecutada; la sesión de prueba de red con el mismo resultado que
la 01. **T044, T045 y T046 hicieron lo que debían**: ninguna sesión es `sesión ilegible` (el lector admite desde T045
las llamadas que el fin del proceso deja sin terminar, y ninguna de las trece trazas quedó fuera), y las citas de diez
de las once sesiones con la skill activada llevan la forma legible dentro de los corchetes, delante del identificador,
y la extracción de T046 las cuenta todas (con la expresión anterior, diez sesiones habrían fallado por `cita ausente`).
El paso «Retirar Python del runner» terminó con 0. La línea de T030 marca la tarea: `ci` en verde
(`gates/evidencia-plataforma.md`) y la prueba de red sin ningún defecto. Los intentos 1 a 6 (ejecuciones 34922606273,
34930222593, 34936425178, 34941499481, 34956596912 y 34961757559) están en las versiones anteriores de este fichero
(historial de git del fichero).

Enlace a la ejecución: <https://github.com/jmorenobl/kitlegal/actions/runs/34999845098> (`databaseId` 34999845098).

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
{"headRefOid":"091facd82d3744dffcfb3b17115c5ea25452d7d5","labels":[],"number":27}
```

Todo presente: el secreto y las dos etiquetas.

## 2. Órdenes de §12.2

**Primera** (la etiqueta no estaba puesta, porque el intento 6 la quitó al terminar, así que no se quitó nada):

```text
la etiqueta evals-prueba-de-red no está puesta
```

**Segunda**, `gh pr edit --add-label evals-prueba-de-red`, a las 17:12:50Z, con el paso directo del envoltorio de la
terminal:

```text
https://github.com/jmorenobl/kitlegal/pull/27
```

**Tercera**, a la primera ya con la ejecución:

```text
etiqueta puesta: 2026-09-15T17:12:52Z
{"evals":{"conclusion":"","createdAt":"2026-09-15T17:12:54Z","databaseId":34999845098,"headSha":"091facd82d3744dffcfb3b17115c5ea25452d7d5","status":"in_progress","url":"https://github.com/jmorenobl/kitlegal/actions/runs/34999845098","workflowName":"evals"},"posteriores_a_la_etiqueta":[{"createdAt":"2026-09-15T17:12:54Z","databaseId":34999845098,"workflowName":"evals"}]}
```

**Cuarta**, `gh run watch 34999845098 --exit-status`: la ejecución duró 7 m 46 s, dentro del tope de la herramienta de
la sesión (600 s), así que la orden terminó a la primera, **`código 0`**. Sus últimas líneas:

```text
✓ h5-skill-boe-legislacion evals jmorenobl/kitlegal#27 · 34999845098
Triggered via pull_request about 7 minutes ago

JOBS
✓ evals in 7m46s (ID 104485233565)
  ✓ Set up job
  ✓ Obtener el código del commit evaluado
  ✓ Instalar Go y restaurar la caché
  ✓ Instalar strace y Claude Code
  ✓ Instalar kitlegal y las skills como las deja make install
  ✓ Retirar Python del runner
  ✓ Ejecutar las evals
  ✓ Post Instalar Go y restaurar la caché
  ✓ Post Obtener el código del commit evaluado
  ✓ Complete job
código 0
```

**Quinta** (informe entre marcas): **código 0**; imprime `informe.md` (580 líneas del registro) e `informe.json` (504)
enteros, cada uno de su marca de inicio a su marca de fin. Salida completa, tal cual la da `gh run view --log`, en el
**anexo A**. Para conservarla entera, la orden se ejecutó tal cual dentro de `rtk proxy sh -c` con su salida redirigida
a un fichero temporal del directorio del hito (borrado tras copiarla aquí) y con una línea final
`código de la quinta orden: $?` añadida detrás. Esta vez ninguna línea de `make` se intercala en el informe: la receta
terminó con 0.

**Sexta** (salida de la retirada de Python entre marcas): **código 0**; 925 líneas del registro, de
`--- inicio de la retirada de Python ---` a `--- fin de la retirada de Python ---`. Salida completa en el **anexo B**,
obtenida de la misma forma.

La orden de `--log-failed` que va tras el bloque no procede (las dos anteriores no fallan); el registro completo
(`gh run view 34999845098 --log`, 2 455 líneas) se leyó aparte para §3.

**Séptima orden** (quitar la etiqueta, al terminar): a la primera, sin ningún error de la plataforma:

```text
https://github.com/jmorenobl/kitlegal/pull/27
la etiqueta evals-prueba-de-red no está puesta
```

`gh pr view 27 --json labels` después: `{"labels":[]}`; la API de eventos, `2026-09-15T17:22:25Z	unlabeled` (§5,
S12 (1)).

## 3. Lo que muestra la ejecución

### 3.1 Pasos y tiempos

`gh run view 34999845098 --json jobs` (`createdAt` 17:12:54Z, `event` `pull_request`, `headSha` `091facd…`,
`workflowName` `evals`, job 104485233565, de 17:12:57 a 17:20:43):

| Paso | Resultado | Inicio | Fin | Duración |
|---|---|---|---|---|
| Obtener el código del commit evaluado | success | 17:12:59 | 17:13:02 | 3 s |
| Instalar Go y restaurar la caché | success | 17:13:02 | 17:13:26 | 24 s |
| Instalar strace y Claude Code | success | 17:13:26 | 17:13:38 | 12 s |
| Instalar kitlegal y las skills como las deja make install | success | 17:13:38 | 17:14:19 | 41 s |
| Retirar Python del runner | success | 17:14:19 | 17:16:04 | 1 m 45 s |
| Ejecutar las evals | **success** | 17:16:04 | 17:20:39 | 4 m 35 s |

Del paso de instalación: `strace is already the newest version (6.8-0ubuntu2).`, `added 2 packages in 4s` y
`claude --version` → `2.1.270 (Claude Code)`; el paso no instala `bubblewrap` ni `socat` (T036). De `make install`:
`CGO_ENABLED=0 go install -trimpath -ldflags "-X main.version=091facd …" ./cmd/kitlegal`,
`instalar-skills: boe-legislacion → /home/runner/work/kitlegal/kitlegal/skills/boe-legislacion` e
`instalar-skills: kitlegal → /home/runner/go/bin/kitlegal`. El entorno del paso de evals lista
`MODELO_DE_EVALS: claude-haiku-4-5-20251001`, `COMMIT_EVALUADO: 091facd…`, `PRUEBA_DE_RED: true` y
`CLAUDE_CODE_OAUTH_TOKEN: ***` (el secreto llega al paso, enmascarado).

### 3.2 Retirada de Python (anexo B)

- `búsqueda: find / ( -path /proc -o -path /sys ) -prune -o ( ( -type f -perm /111 ( -iname python* -o -iname pypy* ) ) -o ( -type l ( -iname python* -o -iname pypy* ) ) -o ( ( -type f -o -type l ) ( -iname libpython* -o -iname libpypy* ) ) ) -print`
  a las 17:14:19,20; la primera línea `retirado:` a las 17:15:44,77: **la búsqueda como root en toda la imagen tarda
  1 m 25 s** (1 m 52 s en el intento 6, 6 m 7 s en el 5, 2 m 32 s en el 4, 2 m 37 s en el 3, 1 m 23 s en el 2: el
  tiempo lo pone el disco del runner que toque) y termina con 0 (con `set -euo pipefail`, un `find` con error habría
  detenido el paso ahí).
- **921 líneas `retirado:`**, exactamente las mismas rutas que en los intentos 3 a 6 (comparadas una a una con el
  anexo B de la versión anterior de este fichero, ordenadas: ninguna diferencia; misma imagen), entre las 17:15:44 y
  las 17:16:01 (17 s), ninguna seguida de un error de `rm`: nada estaba en un sistema de ficheros de solo lectura. Por
  árbol: 341 bajo `/opt/hostedtoolcache`, 315 bajo `/var/lib`, 143 bajo `/usr/share`, 54 bajo `/usr/local`, 21 bajo
  `/opt/az`, 19 bajo `/usr/lib`, 19 bajo `/opt/pipx`, 5 en `/usr/bin` (`python3.12-config`, `python3-config`,
  `python3.12`, `python3`, `python`) y 4 en `/usr/sbin` (las herramientas `python*-bpfcc`).
- Instalaciones retiradas enteras por la regla del prefijo, las mismas catorce de los intentos 3 a 6: `/opt/az`;
  `/opt/hostedtoolcache/PyPy/3.9.19/x64`, `/opt/hostedtoolcache/PyPy/3.10.16/x64`,
  `/opt/hostedtoolcache/PyPy/3.11.15/x64`, `/opt/hostedtoolcache/Python/3.10.21/x64`,
  `/opt/hostedtoolcache/Python/3.11.16/x64`, `/opt/hostedtoolcache/Python/3.12.14/x64`,
  `/opt/hostedtoolcache/Python/3.13.15/x64`, `/opt/hostedtoolcache/Python/3.14.7/x64`; `/opt/pipx/shared`,
  `/opt/pipx/venvs/ansible-core`, `/opt/pipx/venvs/yamllint`; `/usr/lib/google-cloud-sdk/platform/bundledpythonunix`; y
  `/usr/share/miniconda`.
- `búsqueda tras retirar: ninguno` a las 17:16:04,45 (la segunda búsqueda, 3,1 s), la comprobación de lo usado sin
  ningún `la retirada se llevó algo que el job usa`, y la marca de fin. Código 0.

### 3.3 Ejecutar las evals (anexo A y registro del paso)

Fuera del informe, el registro del paso muestra: `scripts/evals.sh "boe-legislacion"` a las 17:16:04; la comprobación
5 (`TestEvalsDelRepositorio` y `TestIdentificadoresDeLasNormas`) en `ok` a las 17:16:19; trece preparaciones de sesión
(`TestPrepararSesion`) en `ok`, cada una seguida de su sesión, que terminó por sí misma, en el orden del guion (las doce
evals por su nombre de fichero y la de prueba de red la última): medido por el instante del `ok` de la preparación
siguiente, 01 ≈ 27 s, 02 ≈ 26 s, 03 ≈ 23 s, 04 ≈ 17 s, 05 ≈ 21 s, 06 ≈ 23 s, 07 ≈ 25 s, 08 ≈ 20 s, 09 ≈ 17 s,
10 ≈ 17 s, 11 ≈ 8 s, 12 ≈ 7 s y `01-lpac-articulo-21-prueba-de-red` ≈ 28 s, de las 17:16:20 a las 17:20:39;
`TestInformeDelJob` en `ok` a las 17:20:39; las marcas y los dos ficheros; y ninguna línea de error de `make`
(la receta terminó con 0; el paso, con `success`).

`sin_python` del informe (comprobación 3 del guion, repetida como root antes de la primera sesión):

```text
búsqueda: find / ( -path /proc -o -path /sys ) -prune -o ( ( -type f -perm /111 ( -iname python* -o -iname pypy* ) ) -o ( -type l ( -iname python* -o -iname pypy* ) ) -o ( ( -type f -o -type l ) ( -iname libpython* -o -iname libpypy* ) ) ) -print
usuario: root
resultado: ninguno
```

Cabecera del informe: `Veredicto: aprobado`, `Motivos: ninguno`, `Modelo del job: claude-haiku-4-5-20251001`, `Modelos
de las sesiones: claude-haiku-4-5-20251001`, `Versiones de Claude Code: 2.1.270`, `Commit:
091facd82d3744dffcfb3b17115c5ea25452d7d5`. Ficheros mal formados: ninguno. Peticiones llegadas a la red: «ninguna
petición llegó a la red de una fuente» (`red` vacío en la raíz del JSON y `llegadas_a_la_red` vacío en las trece
sesiones). **Las trece trazas se leyeron enteras**: `invocaciones` con su orden, su código y sus conexiones en las once
sesiones que invocaron el binario (las dos de no activación, sin ninguna invocación, como esperan sus evals); ninguna
`salida_de_error` con contenido; ningún motivo en ninguna sesión. Las 25 invocaciones del binario tienen código (22 con
0, 2 con 5 y 1 con 4) y conexiones coherentes con él: las dos de código 5, una sola conexión `127.0.0.1:9` de clase
`local` cada una; las de 0 y la de 4, ninguna. Ninguna de clase `red`.

**Invocaciones fuera de lo grabado** (tres: las dos de la prueba de red y una búsqueda de la 05; ninguna con conexión de
clase `red`):

| Sesión | Orden | Código |
|---|---|---|
| `01-lpac-articulo-21-prueba-de-red` | `boe articulo BOE-A-2015-10565 a9998 --json` | 5 |
| `01-lpac-articulo-21-prueba-de-red` | `boe articulo BOE-A-2015-10565 a9998 --offline --json` | 4 |
| `05-trlrhl-impuestos-municipales` | `boe buscar texto refundido ley reguladora haciendas locales --json` | 5 |

Ninguna invocación en `otras_fallidas`. La tercera fila es nueva: la 05 empezó por buscar la norma con sus propios
términos (paso 1 del protocolo; FR-005), la búsqueda no está grabada (lo grabado de esa norma es la búsqueda del
manifiesto, con otros términos), el proxy la rechazó (código 5, conexión `127.0.0.1:9` `local`) y la sesión siguió
por `references/normas.md` hasta `BOE-A-2004-4214`, leyó el índice y `a59` con 0 y citó: es el caso que FR-076 (1)
describe —«un paso legítimo del protocolo que sale de lo grabado (…) buscar con términos propios, FR-005»—, la
invocación queda nombrada con su eval y no hace fallar la eval ni el veredicto, ejercido en la plataforma por primera
vez.

Sesiones (de `informe.json`; las trece con `codigo_de_la_sesion` 0 y `fin_de_la_sesion` `result success`; en las diez
positivas y en la de prueba de red la skill se activó, y en las dos de no activación no):

| Sesión | Invocaciones (orden → código) | Comandos ausentes | Citas ausentes | Pasa |
|---|---|---|---|---|
| `01-lpac-articulo-21` | `indice` → 0; `articulo … a21` → 0 | ninguno | ninguna | **sí** |
| `01-lpac-articulo-21-prueba-de-red` | `articulo … a9998` → 5 (`127.0.0.1:9`, `local`); `articulo … a9998 --offline` → 4 (sin conexiones); `indice` → 0; `articulo … a21` → 0 | ninguno | ninguna | **sí** |
| `02-lcsp-contrato-menor` | `indice` → 0; `articulo … a1-30` → 0 | ninguno | ninguna | **sí** |
| `03-lrbrl-atribuciones-del-pleno` | `indice` → 0; `articulo … a22` → 0 | ninguno | ninguna | **sí** |
| `04-lgt-prescripcion` | `indice` → 0; `articulo … a66` → 0 | ninguno | ninguna | **sí** |
| `05-trlrhl-impuestos-municipales` | `buscar texto refundido ley reguladora haciendas locales` → 5 (`127.0.0.1:9`, `local`); `indice` → 0; `articulo … a59` → 0 | ninguno | ninguna | **sí** |
| `06-irpf-rendimientos-del-trabajo` | `indice` → 0; `articulo … a17` → 0 | ninguno | ninguna | **sí** |
| `07-lrjsp-principio-de-legalidad` | `indice` → 0; `articulo … a25` → 0 | ninguno | ninguna | **sí** |
| `08-ltaibg-plazo-de-resolucion` | `indice` → 0; `articulo … a20` → 0 | ninguno | ninguna | **sí** |
| `09-constitucion-articulo-140` | `indice` → 0; `articulo … a140` → 0 | ninguno | ninguna | **sí** |
| `10-et-vacaciones` | `indice` → 0; `articulo … a38` → 0 | ninguno | ninguna | **sí** |
| `11-no-activa-programacion` | ninguna | ninguno | ninguna | **sí** |
| `12-no-activa-acuerdo-entre-amigos` | ninguna | ninguno | ninguna | **sí** |

Las respuestas (anexo A) y sus citas, tal como las extrae `formaDeCita` desde T046 (`citas_encontradas` con la pareja
esperada en las once sesiones con la skill activada): la 01 y la de prueba de red exponen el artículo 21 por apartados,
cada uno con `[art. 21.N de la Ley 39/2015, BOE-A-2015-10565, bloque a21]` (la de prueba de red no dice qué
devolvieron las dos órdenes de `a9998` que su pregunta pide, pero la traza muestra que las ejecutó, con 5 y 4, como en
los intentos 4 a 6); la 02, el artículo 118 con seis citas `[art. 118.N de la LCSP, BOE-A-2017-12902, bloque a1-30]`
(esta vez sin corchete exterior); la 03, el artículo 22 con la forma fija y la forma legible delante,
`artículo 22 de la Ley 7/1985 … [BOE-A-1985-5392, bloque a22]`, la única sesión con esa forma; la 04,
`[art. 66, Ley General Tributaria, BOE-A-2003-23186, bloque a66]`; la 05, sola en su línea tras la lista de impuestos,
`[Art. 59 del TRLRHL, BOE-A-2004-4214, bloque a59]`; la 06, `[art. 17.1 de la Ley 35/2006, BOE-A-2006-20764, bloque
a17]` y, sola en su línea al final, `[art. 17 de la Ley 35/2006, BOE-A-2006-20764, bloque a17]`; la 07, cinco veces
`[art. 25.N, Ley 40/2015, BOE-A-2015-10566, bloque a25]`; la 08, tres veces
`[art. 20.N de la LTAIBG, BOE-A-2013-12887, bloque a20]` (la forma exacta que el intento 6 rechazó); la 09, tras el
texto del artículo y sola en su línea, `[art. 140 de la Constitución Española, BOE-A-1978-31229, bloque a140]`; la 10,
`[Real Decreto Legislativo 2/2015 de 23 de octubre (Estatuto de los Trabajadores), art. 38.1, BOE-A-2015-11430, bloque
a38]`, con paréntesis dentro de los corchetes; la 11 responde con código Go y la 12 con texto. Ninguna sesión citó de
una forma que la extracción no admita (sin `bloque`, sin corchetes o con el identificador fuera).

## 4. Lo que la ejecución confirma y lo que queda para T031

Ningún defecto: el paso de retirada, la preparación de cada sesión, las trece sesiones, `strace`, la lectura de las
trece trazas, la atribución de conexiones, la extracción de citas y el informe hicieron lo que el contrato dice, y
ninguna eval falló. En concreto:

- **T044 y T045 (la llamada que el fin del proceso deja sin terminar).** Ninguna sesión es `sesión ilegible`: las trece
  trazas se leyeron enteras, con sus 25 invocaciones atribuidas. En el intento 6 la carrera del runtime de Go dejó una
  línea `clone(… <unfinished ...>) = ?` en un hilo de una invocación y la sesión quedó sin leer; el informe no dice si
  en esta ejecución alguna de las 25 invocaciones dejó una línea así (el lector las admite sin señalarlas, data-model
  §9, regla 5), así que lo que se comprueba es lo que la tarea pedía: ninguna traza de una sesión no cortada quedó
  fuera. Ningún `connect` en curso apareció en ninguna invocación (ninguna conexión sin resultado en el informe).
- **T046 (la forma legible dentro de los corchetes).** Diez de las once sesiones con la skill activada pusieron la forma
  legible dentro de los corchetes, delante del identificador (01, prueba de red, 02, 04, 05, 06, 07, 08, 09 y 10), con
  etiquetas distintas (artículo y apartado con la norma, la sigla, el nombre completo con paréntesis), y solo la 03
  citó con la forma fija; la extracción cuenta las once (`citas_encontradas` con la pareja esperada en todas). Con la
  expresión anterior a T046, diez sesiones habrían fallado por `cita ausente`: la decisión (B) de la persona es lo que
  hace alcanzable el cierre de 10 de 10 (FR-082, SC-003) con el modelo fijado (D13).
- **T040, T041, T042 y T043** siguen funcionando: las diez positivas leyeron el índice (la 05 tras su búsqueda) y el
  bloque que nombra la pregunta a la primera, ninguna pidió un bloque fuera de lo grabado y ninguna compuso un id del
  número del artículo (la 02 copió `a1-30` del índice).
- **Lo nuevo, sin defecto**: la búsqueda con términos propios de la 05, rechazada por el proxy con 5 y nombrada en
  «fuera de lo grabado» sin hacer fallar nada (§3.3), ejerce por primera vez en la plataforma la distinción de
  FR-076 (1) sobre un paso legítimo del protocolo. La ejecución de cierre (T031) puede mostrar filas así, y su veredicto
  no depende de ellas: solo de FR-072 y de que ninguna petición llegue a la red de una fuente.

Lo que queda para T031 es la ejecución de cierre con la etiqueta `evals` (quickstart §12.3), sobre esta misma cabeza o
una posterior que solo difiera en ficheros del directorio del hito (FR-082, SC-003), con `gates/evals-cierre.md` y
`gates/aceptacion.md`.

## 5. Supuestos de research D22

| Supuesto | Qué muestra esta ejecución | Estado |
|---|---|---|
| **S12** (identificar la ejecución) | (1) La API de eventos devuelve, tal cual y en este orden, trece eventos de la etiqueta antes de la séptima orden: los doce de los intentos 1 a 6 (`2026-09-15T02:48:51Z	labeled` … `2026-09-15T11:20:54Z	unlabeled`) y `2026-09-15T17:12:52Z	labeled` (todos de `evals-prueba-de-red`, actor `jmorenobl`; `gh api --paginate 'repos/{owner}/{repo}/issues/27/events?per_page=100' --jq '.[] \| select(.event == "labeled" or .event == "unlabeled") \| "\(.created_at)\t\(.event)\t\(.label.name)\t\(.actor.login)"'`), y un decimocuarto, `2026-09-15T17:22:25Z	unlabeled`, tras ella; las órdenes ordenan los instantes y eligen `2026-09-15T17:12:52Z`. (2) `created_at` (`…T17:12:52Z`) y `createdAt` (`…T17:12:54Z`) con el mismo formato ISO 8601 UTC con `Z`. (3) La ejecución se creó 2 s después del evento. (5) **Ejercido**: la rama ya tenía seis ejecuciones de `evals` anteriores (34922606273, 34930222593, 34936425178, 34941499481, 34956596912 y 34961757559), y la orden no eligió ninguna: `posteriores_a_la_etiqueta` lista solo 34999845098. (6) `workflowName` `evals` en `posteriores_a_la_etiqueta`, aunque `evals.yml` solo está en la rama de la propuesta. (4) Sin ejercer: la etiqueta no estaba puesta al empezar (la quitó el intento 6) y al final `gh pr view` la listaba, así que se quitó estando puesta | **se cumple** en (1), (2), (3), (5) y (6); (4), sin ejercer |
| **S2** (lo que trae `ubuntu-24.04`) | `sudo` sin contraseña en el paso de retirada y `sudo -n` en la comprobación 3 (`usuario: root`). `strace` 6.8 ya en la imagen (`strace is already the newest version (6.8-0ubuntu2).`). `npm install -g @anthropic-ai/claude-code@2.1.270`: `added 2 packages in 4s`, `claude --version` → `2.1.270 (Claude Code)`. `timeout`: ejercido en las trece sesiones (devolvió el 0 de `claude`). GNU findutils y coreutils: la línea `búsqueda:` seguida de 921 `retirado:` y de `búsqueda tras retirar: ninguno` (`find` con `-perm /111`, `-iname`, `-prune`, y `-H … -quit` en la regla del prefijo; `readlink -e` en lo usado; `rm -rf` sin ningún error). Sin `bubblewrap` ni `socat`, que el job no necesita (T036) | **se cumple** en todo lo ejercido |
| **S7** (retirada de Python) | (1) **Lo que trae**: las mismas 921 rutas de los intentos 3 a 6 (§3.2). (2) **Buscar como root**: `find` recorrió toda la imagen salvo `/proc` y `/sys` en 1 m 25 s y terminó con 0. (3) **Retirar**: 921 borrados sin ningún error; nada en solo lectura. (4) **No romper el job**: la comprobación de lo usado pasó, y después del paso corrieron `bash`, `sudo`, `find` (comprobación 3), `go` (los tests del guion), `make`, `git`, `timeout`, `strace`, `claude` (trece sesiones enteras, con el modelo) y el binario (25 invocaciones leídas de las trece trazas, con sus códigos 0, 4 y 5 y el texto que las respuestas citan) | **se cumple** en (1)-(4) |
| **S4** (formato de `strace -ff` en el runner) | **Se cumple en las trece sesiones.** Las trece se leyeron enteras, con la línea `vfork()` con relleno de alineación en el fichero principal de `claude` de cada una (V63), las `execve` de `bash` y del binario con su argv entero (`-s 131072`, V62), las líneas de creación de hilos de Go (`clone` con `CLONE_THREAD`) y de Claude Code (`clone3`), y sus líneas finales; ninguna sesión no cortada con traza ilegible, que es lo que la columna del supuesto pide. Si alguna invocación dejó una llamada sin terminar (la forma A del intento 6, V65), el lector la admitió desde T045 sin señalarla. La invocación `boe articulo BOE-A-2015-10565 a9998 --json` de `01-lpac-articulo-21-prueba-de-red` tiene una sola conexión, `127.0.0.1:9` de clase `local` (la del proxy que rechaza; la pareja destino y clase presentada una vez), atribuida a ella y no a la que lleva `--offline`, que no conectó; la búsqueda de la 05, una sola conexión igual; las 22 invocaciones con código 0 y la de código 4 no tienen ninguna. Ninguna conexión de clase `red` | **se cumple** en las trece: las líneas `connect`, las de creación de hilos y procesos, las de señal y la atribución por hilos funcionan con las trazas reales del runner; las formas B, C y D y el fichero huérfano de V65 siguen sin evidencia en el runner, y sin evidencia en contra |
| **S9** (una sesión cabe en 240 s) | Las trece terminaron por sí mismas, entre 7 y 28 s cada una (§3.3), con hasta 30 turnos disponibles; ninguna con código 124 ni 137 | **se cumple** en las trece |
| **S10** (códigos de la sesión) | Las trece con `codigo_de_la_sesion` 0: `timeout --kill-after=10s 240s strace -ff …` devolvió el 0 de `claude` cuando este terminó bien, y el guion lo escribió en `codigo-de-la-sesion` | **se cumple** en el 0 y en la propagación; 124 y 137 no se provocan |
| S1 (fuera de la lista de esta tarea) | `pull_request` con `types: [labeled]` ejecutó el `evals.yml` de la rama (`COMMIT_EVALUADO: 091facd…`, `PRUEBA_DE_RED: true`) y el secreto llegó al paso y a la sesión: las trece se autenticaron | se cumple en lo ejercido |
| S5, S6 (Claude Code en `-p`: skills, proxy, credencial, modelo) | S5: las diez positivas y la de prueba de red activaron la skill y las dos de no activación no; Bash ejecutó el binario por el enlace de la skill (las 25 invocaciones, con `KITLEGAL_CACHE_DIR` y el proxy heredados: las de código 0 leyeron la caché sin conectar, y las dos de código 5 conectaron solo a `127.0.0.1:9`), sin sandbox. S6: `CLAUDE_CODE_OAUTH_TOKEN` autenticó las trece sesiones y `--model claude-haiku-4-5-20251001` se aceptó (`Modelos de las sesiones: claude-haiku-4-5-20251001`) | se cumplen en lo ejercido |
| S11 (solo `api.anthropic.com`) | Con `NO_PROXY=api.anthropic.com` y el proxy que rechaza para todo lo demás, las trece sesiones llegaron al modelo y terminaron con `result success`; ninguna invocación del binario conectó fuera del bucle local | se cumple en lo ejercido |

## Anexos

Los dos volcados siguientes son la salida entera de la quinta y de la sexta orden de quickstart §12.2, tal cual las
imprimieron (cada línea con el prefijo de tarea, paso e instante que pone `gh run view --log`), más la línea final con
el código de cada orden. Van entre vallas de cinco acentos graves porque el informe contiene vallas de tres y de cuatro.

## Anexo A · Quinta orden de §12.2: informe entre marcas, tal cual

`````text
evals	Ejecutar las evals	2026-09-15T17:20:39.6369760Z --- inicio de informe.md ---
evals	Ejecutar las evals	2026-09-15T17:20:39.6382160Z # Informe de evals de boe-legislacion
evals	Ejecutar las evals	2026-09-15T17:20:39.6391357Z 
evals	Ejecutar las evals	2026-09-15T17:20:39.6395070Z ## Veredicto
evals	Ejecutar las evals	2026-09-15T17:20:39.6404771Z 
evals	Ejecutar las evals	2026-09-15T17:20:39.6406288Z Veredicto: aprobado
evals	Ejecutar las evals	2026-09-15T17:20:39.6406726Z 
evals	Ejecutar las evals	2026-09-15T17:20:39.6407124Z Motivos: ninguno
evals	Ejecutar las evals	2026-09-15T17:20:39.6407467Z 
evals	Ejecutar las evals	2026-09-15T17:20:39.6407629Z ## Cabecera
evals	Ejecutar las evals	2026-09-15T17:20:39.6407800Z 
evals	Ejecutar las evals	2026-09-15T17:20:39.6407995Z Modelo del job: claude-haiku-4-5-20251001
evals	Ejecutar las evals	2026-09-15T17:20:39.6408348Z 
evals	Ejecutar las evals	2026-09-15T17:20:39.6408598Z Modelos de las sesiones: claude-haiku-4-5-20251001
evals	Ejecutar las evals	2026-09-15T17:20:39.6408998Z 
evals	Ejecutar las evals	2026-09-15T17:20:39.6409384Z Versiones de Claude Code: 2.1.270
evals	Ejecutar las evals	2026-09-15T17:20:39.6409693Z 
evals	Ejecutar las evals	2026-09-15T17:20:39.6412623Z Commit: 091facd82d3744dffcfb3b17115c5ea25452d7d5
evals	Ejecutar las evals	2026-09-15T17:20:39.6413084Z 
evals	Ejecutar las evals	2026-09-15T17:20:39.6413524Z ## Comprobación sin Python
evals	Ejecutar las evals	2026-09-15T17:20:39.6413802Z 
evals	Ejecutar las evals	2026-09-15T17:20:39.6414026Z ```text
evals	Ejecutar las evals	2026-09-15T17:20:39.6415533Z búsqueda: find / ( -path /proc -o -path /sys ) -prune -o ( ( -type f -perm /111 ( -iname python* -o -iname pypy* ) ) -o ( -type l ( -iname python* -o -iname pypy* ) ) -o ( ( -type f -o -type l ) ( -iname libpython* -o -iname libpypy* ) ) ) -print
evals	Ejecutar las evals	2026-09-15T17:20:39.6416737Z usuario: root
evals	Ejecutar las evals	2026-09-15T17:20:39.6417052Z resultado: ninguno
evals	Ejecutar las evals	2026-09-15T17:20:39.6417356Z ```
evals	Ejecutar las evals	2026-09-15T17:20:39.6417489Z 
evals	Ejecutar las evals	2026-09-15T17:20:39.6417644Z ## Ficheros mal formados
evals	Ejecutar las evals	2026-09-15T17:20:39.6417853Z 
evals	Ejecutar las evals	2026-09-15T17:20:39.6417965Z ninguno
evals	Ejecutar las evals	2026-09-15T17:20:39.6418112Z 
evals	Ejecutar las evals	2026-09-15T17:20:39.6418267Z ## Invocaciones fuera de lo grabado
evals	Ejecutar las evals	2026-09-15T17:20:39.6418512Z 
evals	Ejecutar las evals	2026-09-15T17:20:39.6418791Z | Sesión | Eval | Orden | Código |
evals	Ejecutar las evals	2026-09-15T17:20:39.6419512Z | --- | --- | --- | --- |
evals	Ejecutar las evals	2026-09-15T17:20:39.6420334Z | 01-lpac-articulo-21-prueba-de-red | 01-lpac-articulo-21.yaml | boe articulo BOE-A-2015-10565 a9998 --json | 5 |
evals	Ejecutar las evals	2026-09-15T17:20:39.6421706Z | 01-lpac-articulo-21-prueba-de-red | 01-lpac-articulo-21.yaml | boe articulo BOE-A-2015-10565 a9998 --offline --json | 4 |
evals	Ejecutar las evals	2026-09-15T17:20:39.6423424Z | 05-trlrhl-impuestos-municipales | 05-trlrhl-impuestos-municipales.yaml | boe buscar texto refundido ley reguladora haciendas locales --json | 5 |
evals	Ejecutar las evals	2026-09-15T17:20:39.6424307Z 
evals	Ejecutar las evals	2026-09-15T17:20:39.6424495Z ## Peticiones llegadas a la red
evals	Ejecutar las evals	2026-09-15T17:20:39.6424772Z 
evals	Ejecutar las evals	2026-09-15T17:20:39.6425140Z ninguna petición llegó a la red de una fuente
evals	Ejecutar las evals	2026-09-15T17:20:39.6425482Z 
evals	Ejecutar las evals	2026-09-15T17:20:39.6425619Z ## Sesiones
evals	Ejecutar las evals	2026-09-15T17:20:39.6425795Z 
evals	Ejecutar las evals	2026-09-15T17:20:39.6426460Z | Sesión | Eval | Activa | Activada | Sesión terminada | Comandos ausentes | Citas ausentes | Resultado |
evals	Ejecutar las evals	2026-09-15T17:20:39.6427321Z | --- | --- | --- | --- | --- | --- | --- | --- |
evals	Ejecutar las evals	2026-09-15T17:20:39.6428315Z | 01-lpac-articulo-21 | 01-lpac-articulo-21.yaml | sí | sí | sí (código 0) | ninguno | ninguna | pasa |
evals	Ejecutar las evals	2026-09-15T17:20:39.6430699Z | 01-lpac-articulo-21-prueba-de-red | 01-lpac-articulo-21.yaml | sí | sí | sí (código 0) | ninguno | ninguna | pasa |
evals	Ejecutar las evals	2026-09-15T17:20:39.6432229Z | 02-lcsp-contrato-menor | 02-lcsp-contrato-menor.yaml | sí | sí | sí (código 0) | ninguno | ninguna | pasa |
evals	Ejecutar las evals	2026-09-15T17:20:39.6433902Z | 03-lrbrl-atribuciones-del-pleno | 03-lrbrl-atribuciones-del-pleno.yaml | sí | sí | sí (código 0) | ninguno | ninguna | pasa |
evals	Ejecutar las evals	2026-09-15T17:20:39.6435304Z | 04-lgt-prescripcion | 04-lgt-prescripcion.yaml | sí | sí | sí (código 0) | ninguno | ninguna | pasa |
evals	Ejecutar las evals	2026-09-15T17:20:39.6436809Z | 05-trlrhl-impuestos-municipales | 05-trlrhl-impuestos-municipales.yaml | sí | sí | sí (código 0) | ninguno | ninguna | pasa |
evals	Ejecutar las evals	2026-09-15T17:20:39.6438400Z | 06-irpf-rendimientos-del-trabajo | 06-irpf-rendimientos-del-trabajo.yaml | sí | sí | sí (código 0) | ninguno | ninguna | pasa |
evals	Ejecutar las evals	2026-09-15T17:20:39.6440769Z | 07-lrjsp-principio-de-legalidad | 07-lrjsp-principio-de-legalidad.yaml | sí | sí | sí (código 0) | ninguno | ninguna | pasa |
evals	Ejecutar las evals	2026-09-15T17:20:39.6442486Z | 08-ltaibg-plazo-de-resolucion | 08-ltaibg-plazo-de-resolucion.yaml | sí | sí | sí (código 0) | ninguno | ninguna | pasa |
evals	Ejecutar las evals	2026-09-15T17:20:39.6444116Z | 09-constitucion-articulo-140 | 09-constitucion-articulo-140.yaml | sí | sí | sí (código 0) | ninguno | ninguna | pasa |
evals	Ejecutar las evals	2026-09-15T17:20:39.6444977Z | 10-et-vacaciones | 10-et-vacaciones.yaml | sí | sí | sí (código 0) | ninguno | ninguna | pasa |
evals	Ejecutar las evals	2026-09-15T17:20:39.6445726Z | 11-no-activa-programacion | 11-no-activa-programacion.yaml | no | no | sí (código 0) | ninguno | ninguna | pasa |
evals	Ejecutar las evals	2026-09-15T17:20:39.6446627Z | 12-no-activa-acuerdo-entre-amigos | 12-no-activa-acuerdo-entre-amigos.yaml | no | no | sí (código 0) | ninguno | ninguna | pasa |
evals	Ejecutar las evals	2026-09-15T17:20:39.6447107Z 
evals	Ejecutar las evals	2026-09-15T17:20:39.6447265Z ## Sesión 01-lpac-articulo-21
evals	Ejecutar las evals	2026-09-15T17:20:39.6447428Z 
evals	Ejecutar las evals	2026-09-15T17:20:39.6447546Z Eval: 01-lpac-articulo-21.yaml
evals	Ejecutar las evals	2026-09-15T17:20:39.6447709Z 
evals	Ejecutar las evals	2026-09-15T17:20:39.6447790Z Pregunta:
evals	Ejecutar las evals	2026-09-15T17:20:39.6447912Z 
evals	Ejecutar las evals	2026-09-15T17:20:39.6447995Z ```text
evals	Ejecutar las evals	2026-09-15T17:20:39.6448245Z ¿qué dice el art. 21 de la Ley 39/2015?
evals	Ejecutar las evals	2026-09-15T17:20:39.6448494Z ```
evals	Ejecutar las evals	2026-09-15T17:20:39.6448608Z 
evals	Ejecutar las evals	2026-09-15T17:20:39.6448699Z Invocaciones:
evals	Ejecutar las evals	2026-09-15T17:20:39.6448810Z 
evals	Ejecutar las evals	2026-09-15T17:20:39.6448950Z | Orden | Código | Conexiones |
evals	Ejecutar las evals	2026-09-15T17:20:39.6449594Z | --- | --- | --- |
evals	Ejecutar las evals	2026-09-15T17:20:39.6450076Z | boe indice BOE-A-2015-10565 --json | 0 | sin conexiones |
evals	Ejecutar las evals	2026-09-15T17:20:39.6450544Z | boe articulo BOE-A-2015-10565 a21 --json | 0 | sin conexiones |
evals	Ejecutar las evals	2026-09-15T17:20:39.6450794Z 
evals	Ejecutar las evals	2026-09-15T17:20:39.6450878Z Respuesta:
evals	Ejecutar las evals	2026-09-15T17:20:39.6450994Z 
evals	Ejecutar las evals	2026-09-15T17:20:39.6451081Z ```text
evals	Ejecutar las evals	2026-09-15T17:20:39.6451672Z El **art. 21 de la Ley 39/2015** regula la **obligación de la Administración de resolver y notificar**. Sus apartados principales son:
evals	Ejecutar las evals	2026-09-15T17:20:39.6452113Z 
evals	Ejecutar las evals	2026-09-15T17:20:39.6454330Z **1. Obligación de resolución expresa y notificación:** La Administración debe dictar resolución expresa y notificarla en todos los procedimientos, cualquiera que sea su forma de iniciación. Cuando concurran circunstancias como prescripción, renuncia, caducidad o desistimiento, la resolución debe declarar la circunstancia con indicación de hechos y normas. Se exceptúan los procedimientos que terminan por pacto o convenio, y los relativos a derechos sometidos solo a declaración responsable o comunicación [art. 21.1 de la Ley 39/2015, BOE-A-2015-10565, bloque a21].
evals	Ejecutar las evals	2026-09-15T17:20:39.6456389Z 
evals	Ejecutar las evals	2026-09-15T17:20:39.6457578Z **2. Plazo máximo general:** El plazo máximo de notificación será el que fije la norma reguladora de cada procedimiento. Este plazo no puede exceder de **seis meses**, salvo que una norma de rango de Ley establezca uno mayor o lo prevea el Derecho de la Unión Europea [art. 21.2 de la Ley 39/2015, BOE-A-2015-10565, bloque a21].
evals	Ejecutar las evals	2026-09-15T17:20:39.6458568Z 
evals	Ejecutar las evals	2026-09-15T17:20:39.6460022Z **3. Plazo por defecto:** Cuando no haya plazo establecido, será de **tres meses**. Se cuenta desde la fecha del acuerdo de iniciación (procedimientos de oficio) o desde la entrada de la solicitud en el registro electrónico (procedimientos iniciados a solicitud) [art. 21.3 de la Ley 39/2015, BOE-A-2015-10565, bloque a21].
evals	Ejecutar las evals	2026-09-15T17:20:39.6461130Z 
evals	Ejecutar las evals	2026-09-15T17:20:39.6462647Z **4. Publicidad y comunicación:** Las Administraciones deben publicar en su portal web las relaciones de procedimientos con sus plazos máximos y los efectos del silencio administrativo. También informarán al interesado del plazo máximo, incluyéndolo en el acuerdo de iniciación o en comunicación dentro de diez días desde la recepción de la solicitud [art. 21.4 de la Ley 39/2015, BOE-A-2015-10565, bloque a21].
evals	Ejecutar las evals	2026-09-15T17:20:39.6463915Z 
evals	Ejecutar las evals	2026-09-15T17:20:39.6464870Z **5. Habilitación de medios:** Si el volumen de solicitudes podría incumplir el plazo máximo, el órgano competente o su superior jerárquico pueden habilitar medios personales y materiales adicionales [art. 21.5 de la Ley 39/2015, BOE-A-2015-10565, bloque a21].
evals	Ejecutar las evals	2026-09-15T17:20:39.6465697Z 
evals	Ejecutar las evals	2026-09-15T17:20:39.6466668Z **6. Responsabilidad:** El personal y los titulares de órganos competentes son responsables del cumplimiento de la obligación de resolver en plazo. El incumplimiento genera responsabilidad disciplinaria [art. 21.6 de la Ley 39/2015, BOE-A-2015-10565, bloque a21].
evals	Ejecutar las evals	2026-09-15T17:20:39.6467579Z ```
evals	Ejecutar las evals	2026-09-15T17:20:39.6467683Z 
evals	Ejecutar las evals	2026-09-15T17:20:39.6467852Z ## Sesión 01-lpac-articulo-21-prueba-de-red
evals	Ejecutar las evals	2026-09-15T17:20:39.6468066Z 
evals	Ejecutar las evals	2026-09-15T17:20:39.6468182Z Eval: 01-lpac-articulo-21.yaml
evals	Ejecutar las evals	2026-09-15T17:20:39.6468340Z 
evals	Ejecutar las evals	2026-09-15T17:20:39.6468426Z Pregunta:
evals	Ejecutar las evals	2026-09-15T17:20:39.6468536Z 
evals	Ejecutar las evals	2026-09-15T17:20:39.6468610Z ```text
evals	Ejecutar las evals	2026-09-15T17:20:39.6468866Z ¿qué dice el art. 21 de la Ley 39/2015?
evals	Ejecutar las evals	2026-09-15T17:20:39.6469170Z 
evals	Ejecutar las evals	2026-09-15T17:20:39.6470244Z Antes de responder, ejecuta también exactamente estas dos órdenes y di qué devolvieron: `~/.claude/skills/boe-legislacion/scripts/boe articulo BOE-A-2015-10565 a9998 --json` y `~/.claude/skills/boe-legislacion/scripts/boe articulo BOE-A-2015-10565 a9998 --offline --json`.
evals	Ejecutar las evals	2026-09-15T17:20:39.6471222Z ```
evals	Ejecutar las evals	2026-09-15T17:20:39.6471329Z 
evals	Ejecutar las evals	2026-09-15T17:20:39.6471426Z Invocaciones:
evals	Ejecutar las evals	2026-09-15T17:20:39.6471545Z 
evals	Ejecutar las evals	2026-09-15T17:20:39.6471684Z | Orden | Código | Conexiones |
evals	Ejecutar las evals	2026-09-15T17:20:39.6471923Z | --- | --- | --- |
evals	Ejecutar las evals	2026-09-15T17:20:39.6472248Z | boe articulo BOE-A-2015-10565 a9998 --json | 5 | 127.0.0.1:9 (local) |
evals	Ejecutar las evals	2026-09-15T17:20:39.6472728Z | boe articulo BOE-A-2015-10565 a9998 --offline --json | 4 | sin conexiones |
evals	Ejecutar las evals	2026-09-15T17:20:39.6473170Z | boe indice BOE-A-2015-10565 --json | 0 | sin conexiones |
evals	Ejecutar las evals	2026-09-15T17:20:39.6473573Z | boe articulo BOE-A-2015-10565 a21 --json | 0 | sin conexiones |
evals	Ejecutar las evals	2026-09-15T17:20:39.6473810Z 
evals	Ejecutar las evals	2026-09-15T17:20:39.6473896Z Respuesta:
evals	Ejecutar las evals	2026-09-15T17:20:39.6474007Z 
evals	Ejecutar las evals	2026-09-15T17:20:39.6474081Z ```text
evals	Ejecutar las evals	2026-09-15T17:20:39.6474598Z El artículo 21 de la Ley 39/2015 regulala **obligación de resolver** de la Administración. Sus puntos esenciales son:
evals	Ejecutar las evals	2026-09-15T17:20:39.6474988Z 
evals	Ejecutar las evals	2026-09-15T17:20:39.6476411Z **1. Obligación de resolución expresa:** La Administración está obligada a dictar resolución expresa y notificarla en todos los procedimientos, cualquiera que sea su forma de iniciación. Se exceptúan los procedimientos que terminan por pacto, convenio o los sometidos solo a declaración responsable o comunicación [art. 21.1 de la Ley 39/2015, BOE-A-2015-10565, bloque a21].
evals	Ejecutar las evals	2026-09-15T17:20:39.6477693Z 
evals	Ejecutar las evals	2026-09-15T17:20:39.6478939Z **2. Plazo máximo:** La resolución debe notificarse en el plazo que fije la norma reguladora del procedimiento, con un límite máximo de **seis meses**, salvo que una ley establezca un plazo mayor o lo prevea el Derecho de la Unión Europea. Si la norma no fija plazo, será de **tres meses** [art. 21.2 y 21.3 de la Ley 39/2015, BOE-A-2015-10565, bloque a21].
evals	Ejecutar las evals	2026-09-15T17:20:39.6480202Z 
evals	Ejecutar las evals	2026-09-15T17:20:39.6481126Z **3. Conteo de plazos:** Se cuentan desde la fecha del acuerdo de iniciación (procedimientos de oficio) o desde la entrada de la solicitud en el registro electrónico (procedimientos a solicitud) [art. 21.3 de la Ley 39/2015, BOE-A-2015-10565, bloque a21].
evals	Ejecutar las evals	2026-09-15T17:20:39.6482027Z 
evals	Ejecutar las evals	2026-09-15T17:20:39.6483318Z **4. Información a los interesados:** Las Administraciones deben informar de los plazos máximos y los efectos del silencio administrativo, incluyendo esta información en la iniciación del procedimiento o en comunicación dirigida en los diez días siguientes a la recepción de la solicitud [art. 21.4 de la Ley 39/2015, BOE-A-2015-10565, bloque a21].
evals	Ejecutar las evals	2026-09-15T17:20:39.6484421Z 
evals	Ejecutar las evals	2026-09-15T17:20:39.6485092Z **5. Responsabilidad:** El incumplimiento de la obligación de dictar resolución en plazo genera responsabilidad disciplinaria [art. 21.6 de la Ley 39/2015, BOE-A-2015-10565, bloque a21].
evals	Ejecutar las evals	2026-09-15T17:20:39.6485774Z ```
evals	Ejecutar las evals	2026-09-15T17:20:39.6485875Z 
evals	Ejecutar las evals	2026-09-15T17:20:39.6486020Z ## Sesión 02-lcsp-contrato-menor
evals	Ejecutar las evals	2026-09-15T17:20:39.6486195Z 
evals	Ejecutar las evals	2026-09-15T17:20:39.6486305Z Eval: 02-lcsp-contrato-menor.yaml
evals	Ejecutar las evals	2026-09-15T17:20:39.6486478Z 
evals	Ejecutar las evals	2026-09-15T17:20:39.6486564Z Pregunta:
evals	Ejecutar las evals	2026-09-15T17:20:39.6486676Z 
evals	Ejecutar las evals	2026-09-15T17:20:39.6486764Z ```text
evals	Ejecutar las evals	2026-09-15T17:20:39.6487291Z ¿Qué debe incluir el expediente de un contrato menor según el artículo 118 de la Ley de Contratos del Sector Público?
evals	Ejecutar las evals	2026-09-15T17:20:39.6487788Z ```
evals	Ejecutar las evals	2026-09-15T17:20:39.6487884Z 
evals	Ejecutar las evals	2026-09-15T17:20:39.6487978Z Invocaciones:
evals	Ejecutar las evals	2026-09-15T17:20:39.6488096Z 
evals	Ejecutar las evals	2026-09-15T17:20:39.6488228Z | Orden | Código | Conexiones |
evals	Ejecutar las evals	2026-09-15T17:20:39.6488475Z | --- | --- | --- |
evals	Ejecutar las evals	2026-09-15T17:20:39.6488753Z | boe indice BOE-A-2017-12902 --json | 0 | sin conexiones |
evals	Ejecutar las evals	2026-09-15T17:20:39.6489290Z | boe articulo BOE-A-2017-12902 a1-30 --json | 0 | sin conexiones |
evals	Ejecutar las evals	2026-09-15T17:20:39.6489540Z 
evals	Ejecutar las evals	2026-09-15T17:20:39.6489632Z Respuesta:
evals	Ejecutar las evals	2026-09-15T17:20:39.6489741Z 
evals	Ejecutar las evals	2026-09-15T17:20:39.6489828Z ```text
evals	Ejecutar las evals	2026-09-15T17:20:39.6490304Z Según el artículo 118 de la LCSP, el expediente de un contrato menor debe incluir los siguientes elementos:
evals	Ejecutar las evals	2026-09-15T17:20:39.6490676Z 
evals	Ejecutar las evals	2026-09-15T17:20:39.6490791Z ## Elementos obligatorios generales
evals	Ejecutar las evals	2026-09-15T17:20:39.6490968Z 
evals	Ejecutar las evals	2026-09-15T17:20:39.6492275Z 1. **Informe de justificación**: Un informe del órgano de contratación que justifique de manera motivada la necesidad del contrato y que no se está alterando su objeto con el fin de evitar la aplicación de los umbrales de contrato menor (40.000 euros para obras; 15.000 euros para suministros o servicios) [art. 118.2 de la LCSP, BOE-A-2017-12902, bloque a1-30]
evals	Ejecutar las evals	2026-09-15T17:20:39.6493383Z 
evals	Ejecutar las evals	2026-09-15T17:20:39.6493840Z 2. **Aprobación del gasto**: La aprobación presupuestaria del gasto [art. 118.3 de la LCSP, BOE-A-2017-12902, bloque a1-30]
evals	Ejecutar las evals	2026-09-15T17:20:39.6494252Z 
evals	Ejecutar las evals	2026-09-15T17:20:39.6494959Z 3. **Factura**: La incorporación de la factura correspondiente que reúna los requisitos que las normas de desarrollo de la LCSP establezcan [art. 118.3 de la LCSP, BOE-A-2017-12902, bloque a1-30]
evals	Ejecutar las evals	2026-09-15T17:20:39.6495580Z 
evals	Ejecutar las evals	2026-09-15T17:20:39.6495730Z ## Elemento adicional para contratos menores de obras
evals	Ejecutar las evals	2026-09-15T17:20:39.6495945Z 
evals	Ejecutar las evals	2026-09-15T17:20:39.6496223Z Para los contratos menores de obras, además de lo anterior, debe incluirse:
evals	Ejecutar las evals	2026-09-15T17:20:39.6496524Z 
evals	Ejecutar las evals	2026-09-15T17:20:39.6497121Z 4. **Presupuesto de las obras**: Sin perjuicio de que deba existir el correspondiente proyecto cuando sea requerido por disposiciones vigentes [art. 118.4 de la LCSP, BOE-A-2017-12902, bloque a1-30]
evals	Ejecutar las evals	2026-09-15T17:20:39.6497853Z 
evals	Ejecutar las evals	2026-09-15T17:20:39.6498894Z 5. **Informe de supervisión** (cuando proceda): Si el trabajo afecta a la estabilidad, seguridad o estanqueidad de la obra, se debe solicitar el informe de las oficinas o unidades de supervisión a que se refiere el artículo 235 LCSP [art. 118.4 de la LCSP, BOE-A-2017-12902, bloque a1-30]
evals	Ejecutar las evals	2026-09-15T17:20:39.6500018Z 
evals	Ejecutar las evals	2026-09-15T17:20:39.6500105Z ## Excepciones
evals	Ejecutar las evals	2026-09-15T17:20:39.6500225Z 
evals	Ejecutar las evals	2026-09-15T17:20:39.6501337Z El requisito del informe de justificación (apartado 2) no se aplica en contratos cuyo pago se verifique a través de sistemas de anticipos de caja fija u otros similares para realizar pagos menores, cuando el valor estimado no exceda de 5.000 euros [art. 118.5 de la LCSP, BOE-A-2017-12902, bloque a1-30]
evals	Ejecutar las evals	2026-09-15T17:20:39.6502440Z ```
evals	Ejecutar las evals	2026-09-15T17:20:39.6502547Z 
evals	Ejecutar las evals	2026-09-15T17:20:39.6502713Z ## Sesión 03-lrbrl-atribuciones-del-pleno
evals	Ejecutar las evals	2026-09-15T17:20:39.6502908Z 
evals	Ejecutar las evals	2026-09-15T17:20:39.6503038Z Eval: 03-lrbrl-atribuciones-del-pleno.yaml
evals	Ejecutar las evals	2026-09-15T17:20:39.6503229Z 
evals	Ejecutar las evals	2026-09-15T17:20:39.6503314Z Pregunta:
evals	Ejecutar las evals	2026-09-15T17:20:39.6503421Z 
evals	Ejecutar las evals	2026-09-15T17:20:39.6503498Z ```text
evals	Ejecutar las evals	2026-09-15T17:20:39.6504039Z ¿Qué atribuciones tiene el Pleno del ayuntamiento según el artículo 22 de la Ley reguladora de las Bases del Régimen Local?
evals	Ejecutar las evals	2026-09-15T17:20:39.6504526Z ```
evals	Ejecutar las evals	2026-09-15T17:20:39.6504623Z 
evals	Ejecutar las evals	2026-09-15T17:20:39.6504708Z Invocaciones:
evals	Ejecutar las evals	2026-09-15T17:20:39.6504826Z 
evals	Ejecutar las evals	2026-09-15T17:20:39.6504970Z | Orden | Código | Conexiones |
evals	Ejecutar las evals	2026-09-15T17:20:39.6505212Z | --- | --- | --- |
evals	Ejecutar las evals	2026-09-15T17:20:39.6505482Z | boe indice BOE-A-1985-5392 --json | 0 | sin conexiones |
evals	Ejecutar las evals	2026-09-15T17:20:39.6505878Z | boe articulo BOE-A-1985-5392 a22 --json | 0 | sin conexiones |
evals	Ejecutar las evals	2026-09-15T17:20:39.6506133Z 
evals	Ejecutar las evals	2026-09-15T17:20:39.6506222Z Respuesta:
evals	Ejecutar las evals	2026-09-15T17:20:39.6506337Z 
evals	Ejecutar las evals	2026-09-15T17:20:39.6506416Z ```text
evals	Ejecutar las evals	2026-09-15T17:20:39.6507493Z Según el **artículo 22 de la Ley 7/1985, de 2 de abril, Reguladora de las Bases del Régimen Local** [BOE-A-1985-5392, bloque a22], el Pleno municipal **—integrado por todos los Concejales y presidido por el Alcalde—** tiene las siguientes atribuciones principales:
evals	Ejecutar las evals	2026-09-15T17:20:39.6508325Z 
evals	Ejecutar las evals	2026-09-15T17:20:39.6508433Z **Atribuciones del Pleno:**
evals	Ejecutar las evals	2026-09-15T17:20:39.6508588Z 
evals	Ejecutar las evals	2026-09-15T17:20:39.6508785Z 1. **Control y fiscalización** de los órganos de gobierno
evals	Ejecutar las evals	2026-09-15T17:20:39.6509954Z 2. **Participación en organizaciones supramunicipales**, alteración del término municipal, creación o supresión de municipios y cambios en la capitalidad, nombre, bandera, enseña o escudo del municipio
evals	Ejecutar las evals	2026-09-15T17:20:39.6511153Z 3. **Aprobación del planeamiento urbanístico** (planeamiento general y otros instrumentos de ordenación)
evals	Ejecutar las evals	2026-09-15T17:20:39.6511797Z 4. **Aprobación del reglamento orgánico y ordenanzas** municipales
evals	Ejecutar las evals	2026-09-15T17:20:39.6512543Z 5. **Determinación de recursos tributarios**, aprobación y modificación de presupuestos, disposición de gastos y aprobación de cuentas
evals	Ejecutar las evals	2026-09-15T17:20:39.6513347Z 6. **Gestión de servicios** (aprobación de formas de gestión y expedientes de municipalización)
evals	Ejecutar las evals	2026-09-15T17:20:39.6513943Z 7. **Aceptación de delegaciones** de competencias de otras Administraciones
evals	Ejecutar las evals	2026-09-15T17:20:39.6514421Z 8. **Planteamiento de conflictos** de competencias con otras entidades
evals	Ejecutar las evals	2026-09-15T17:20:39.6515024Z 9. **Aprobación de plantilla de personal**, relación de puestos y retribuciones de funcionarios
evals	Ejecutar las evals	2026-09-15T17:20:39.6515596Z 10. **Ejercicio de acciones judiciales** y defensa de la corporación
evals	Ejecutar las evals	2026-09-15T17:20:39.6516057Z 11. **Declaración de lesividad** de actos del Ayuntamiento
evals	Ejecutar las evals	2026-09-15T17:20:39.6516550Z 12. **Alteración de la calificación jurídica** de bienes de dominio público
evals	Ejecutar las evals	2026-09-15T17:20:39.6517128Z 13. **Operaciones de crédito** (cuando superen umbrales de porcentaje de presupuesto)
evals	Ejecutar las evals	2026-09-15T17:20:39.6517711Z 14. **Aprobación de proyectos** de obras y servicios no previstos en presupuestos
evals	Ejecutar las evals	2026-09-15T17:20:39.6518426Z 15. **Otras atribuciones** que exijan mayoría especial o que las leyes le confieran
evals	Ejecutar las evals	2026-09-15T17:20:39.6518738Z 
evals	Ejecutar las evals	2026-09-15T17:20:39.6519467Z El Pleno **puede delegar** sus atribuciones en el Alcalde y en la Junta de Gobierno Local, **excepto** en las enunciadas en los apartados a), b), c), d), e), f), g), h), i), l) y p), que son **indelegables**.
evals	Ejecutar las evals	2026-09-15T17:20:39.6520161Z ```
evals	Ejecutar las evals	2026-09-15T17:20:39.6520257Z 
evals	Ejecutar las evals	2026-09-15T17:20:39.6520417Z ## Sesión 04-lgt-prescripcion
evals	Ejecutar las evals	2026-09-15T17:20:39.6520584Z 
evals	Ejecutar las evals	2026-09-15T17:20:39.6520693Z Eval: 04-lgt-prescripcion.yaml
evals	Ejecutar las evals	2026-09-15T17:20:39.6520858Z 
evals	Ejecutar las evals	2026-09-15T17:20:39.6520955Z Pregunta:
evals	Ejecutar las evals	2026-09-15T17:20:39.6521059Z 
evals	Ejecutar las evals	2026-09-15T17:20:39.6521147Z ```text
evals	Ejecutar las evals	2026-09-15T17:20:39.6521766Z ¿En cuántos años prescribe el derecho de la Administración a liquidar una deuda tributaria según el artículo 66 de la Ley General Tributaria?
evals	Ejecutar las evals	2026-09-15T17:20:39.6522407Z ```
evals	Ejecutar las evals	2026-09-15T17:20:39.6522504Z 
evals	Ejecutar las evals	2026-09-15T17:20:39.6522604Z Invocaciones:
evals	Ejecutar las evals	2026-09-15T17:20:39.6522719Z 
evals	Ejecutar las evals	2026-09-15T17:20:39.6522853Z | Orden | Código | Conexiones |
evals	Ejecutar las evals	2026-09-15T17:20:39.6523099Z | --- | --- | --- |
evals	Ejecutar las evals	2026-09-15T17:20:39.6523368Z | boe indice BOE-A-2003-23186 --json | 0 | sin conexiones |
evals	Ejecutar las evals	2026-09-15T17:20:39.6523761Z | boe articulo BOE-A-2003-23186 a66 --json | 0 | sin conexiones |
evals	Ejecutar las evals	2026-09-15T17:20:39.6524007Z 
evals	Ejecutar las evals	2026-09-15T17:20:39.6524093Z Respuesta:
evals	Ejecutar las evals	2026-09-15T17:20:39.6524196Z 
evals	Ejecutar las evals	2026-09-15T17:20:39.6524278Z ```text
evals	Ejecutar las evals	2026-09-15T17:20:39.6525000Z Según el **artículo 66 de la Ley 58/2003, de 17 de diciembre, General Tributaria**, el derecho de la Administración a liquidar una deuda tributaria prescribe a los **cuatro años**.
evals	Ejecutar las evals	2026-09-15T17:20:39.6525580Z 
evals	Ejecutar las evals	2026-09-15T17:20:39.6526425Z Específicamente, el artículo establece que «prescriben a los cuatro años» el derecho de la Administración para determinar la deuda tributaria mediante liquidación [art. 66, Ley General Tributaria, BOE-A-2003-23186, bloque a66].
evals	Ejecutar las evals	2026-09-15T17:20:39.6527246Z ```
evals	Ejecutar las evals	2026-09-15T17:20:39.6527360Z 
evals	Ejecutar las evals	2026-09-15T17:20:39.6527514Z ## Sesión 05-trlrhl-impuestos-municipales
evals	Ejecutar las evals	2026-09-15T17:20:39.6527710Z 
evals	Ejecutar las evals	2026-09-15T17:20:39.6527839Z Eval: 05-trlrhl-impuestos-municipales.yaml
evals	Ejecutar las evals	2026-09-15T17:20:39.6528028Z 
evals	Ejecutar las evals	2026-09-15T17:20:39.6528303Z Pregunta:
evals	Ejecutar las evals	2026-09-15T17:20:39.6528414Z 
evals	Ejecutar las evals	2026-09-15T17:20:39.6528490Z ```text
evals	Ejecutar las evals	2026-09-15T17:20:39.6529297Z ¿Qué impuestos pueden exigir los ayuntamientos según el artículo 59 del texto refundido de la Ley reguladora de las Haciendas Locales?
evals	Ejecutar las evals	2026-09-15T17:20:39.6530090Z ```
evals	Ejecutar las evals	2026-09-15T17:20:39.6530197Z 
evals	Ejecutar las evals	2026-09-15T17:20:39.6530283Z Invocaciones:
evals	Ejecutar las evals	2026-09-15T17:20:39.6530404Z 
evals	Ejecutar las evals	2026-09-15T17:20:39.6530561Z | Orden | Código | Conexiones |
evals	Ejecutar las evals	2026-09-15T17:20:39.6530810Z | --- | --- | --- |
evals	Ejecutar las evals	2026-09-15T17:20:39.6531202Z | boe buscar texto refundido ley reguladora haciendas locales --json | 5 | 127.0.0.1:9 (local) |
evals	Ejecutar las evals	2026-09-15T17:20:39.6531716Z | boe indice BOE-A-2004-4214 --json | 0 | sin conexiones |
evals	Ejecutar las evals	2026-09-15T17:20:39.6532112Z | boe articulo BOE-A-2004-4214 a59 --json | 0 | sin conexiones |
evals	Ejecutar las evals	2026-09-15T17:20:39.6532364Z 
evals	Ejecutar las evals	2026-09-15T17:20:39.6532444Z Respuesta:
evals	Ejecutar las evals	2026-09-15T17:20:39.6532561Z 
evals	Ejecutar las evals	2026-09-15T17:20:39.6532648Z ```text
evals	Ejecutar las evals	2026-09-15T17:20:39.6533061Z Según el artículo 59 del TRLRHL, los ayuntamientos pueden exigir los siguientes impuestos:
evals	Ejecutar las evals	2026-09-15T17:20:39.6533393Z 
evals	Ejecutar las evals	2026-09-15T17:20:39.6533496Z **Impuestos obligatorios:**
evals	Ejecutar las evals	2026-09-15T17:20:39.6533827Z Los ayuntamientos exigirán, de acuerdo con la ley:
evals	Ejecutar las evals	2026-09-15T17:20:39.6534034Z 
evals	Ejecutar las evals	2026-09-15T17:20:39.6534149Z 1. **Impuesto sobre Bienes Inmuebles**
evals	Ejecutar las evals	2026-09-15T17:20:39.6534484Z 2. **Impuesto sobre Actividades Económicas**
evals	Ejecutar las evals	2026-09-15T17:20:39.6534872Z 3. **Impuesto sobre Vehículos de Tracción Mecánica**
evals	Ejecutar las evals	2026-09-15T17:20:39.6535078Z 
evals	Ejecutar las evals	2026-09-15T17:20:39.6535197Z **Impuestos potestivos (voluntarios):**
evals	Ejecutar las evals	2026-09-15T17:20:39.6535590Z Asimismo, los ayuntamientos **podrán** establecer y exigir:
evals	Ejecutar las evals	2026-09-15T17:20:39.6535844Z 
evals	Ejecutar las evals	2026-09-15T17:20:39.6536009Z 4. **Impuesto sobre Construcciones, Instalaciones y Obras**
evals	Ejecutar las evals	2026-09-15T17:20:39.6536597Z 5. **Impuesto sobre el Incremento de Valor de los Terrenos de Naturaleza Urbana**
evals	Ejecutar las evals	2026-09-15T17:20:39.6536902Z 
evals	Ejecutar las evals	2026-09-15T17:20:39.6537031Z [Art. 59 del TRLRHL, BOE-A-2004-4214, bloque a59]
evals	Ejecutar las evals	2026-09-15T17:20:39.6537245Z 
evals	Ejecutar las evals	2026-09-15T17:20:39.6538314Z La diferencia clave es que los tres primeros son impuestos cuya exacción es obligatoria, mientras que los dos últimos quedan a la decisión de cada ayuntamiento, siendo necesario en todos los casos que se establezcan de acuerdo con la ley y las respectivas ordenanzas fiscales locales.
evals	Ejecutar las evals	2026-09-15T17:20:39.6539397Z ```
evals	Ejecutar las evals	2026-09-15T17:20:39.6539498Z 
evals	Ejecutar las evals	2026-09-15T17:20:39.6539665Z ## Sesión 06-irpf-rendimientos-del-trabajo
evals	Ejecutar las evals	2026-09-15T17:20:39.6539856Z 
evals	Ejecutar las evals	2026-09-15T17:20:39.6539984Z Eval: 06-irpf-rendimientos-del-trabajo.yaml
evals	Ejecutar las evals	2026-09-15T17:20:39.6540186Z 
evals	Ejecutar las evals	2026-09-15T17:20:39.6540264Z Pregunta:
evals	Ejecutar las evals	2026-09-15T17:20:39.6540453Z 
evals	Ejecutar las evals	2026-09-15T17:20:39.6540539Z ```text
evals	Ejecutar las evals	2026-09-15T17:20:39.6541027Z ¿Qué rendimientos se consideran rendimientos íntegros del trabajo según el artículo 17 de la ley del IRPF?
evals	Ejecutar las evals	2026-09-15T17:20:39.6541492Z ```
evals	Ejecutar las evals	2026-09-15T17:20:39.6541595Z 
evals	Ejecutar las evals	2026-09-15T17:20:39.6541681Z Invocaciones:
evals	Ejecutar las evals	2026-09-15T17:20:39.6541791Z 
evals	Ejecutar las evals	2026-09-15T17:20:39.6541928Z | Orden | Código | Conexiones |
evals	Ejecutar las evals	2026-09-15T17:20:39.6542166Z | --- | --- | --- |
evals	Ejecutar las evals	2026-09-15T17:20:39.6542445Z | boe indice BOE-A-2006-20764 --json | 0 | sin conexiones |
evals	Ejecutar las evals	2026-09-15T17:20:39.6542843Z | boe articulo BOE-A-2006-20764 a17 --json | 0 | sin conexiones |
evals	Ejecutar las evals	2026-09-15T17:20:39.6543093Z 
evals	Ejecutar las evals	2026-09-15T17:20:39.6543171Z Respuesta:
evals	Ejecutar las evals	2026-09-15T17:20:39.6543281Z 
evals	Ejecutar las evals	2026-09-15T17:20:39.6543360Z ```text
evals	Ejecutar las evals	2026-09-15T17:20:39.6545108Z Según el artículo 17 de la Ley 35/2006 (LIRPF), **se consideran rendimientos íntegros del trabajo todas las contraprestaciones o utilidades, cualquiera que sea su denominación o naturaleza, dinerarias o en especie, que deriven, directa o indirectamente, del trabajo personal o de la relación laboral o estatutaria y no tengan el carácter de rendimientos de actividades económicas** [art. 17.1 de la Ley 35/2006, BOE-A-2006-20764, bloque a17].
evals	Ejecutar las evals	2026-09-15T17:20:39.6546482Z 
evals	Ejecutar las evals	2026-09-15T17:20:39.6546656Z ## Rendimientos que se incluyen específicamente:
evals	Ejecutar las evals	2026-09-15T17:20:39.6546868Z 
evals	Ejecutar las evals	2026-09-15T17:20:39.6546974Z **1. En el apartado 1 se detallan:**
evals	Ejecutar las evals	2026-09-15T17:20:39.6547234Z - a) Los sueldos y salarios
evals	Ejecutar las evals	2026-09-15T17:20:39.6547489Z - b) Las prestaciones por desempleo
evals	Ejecutar las evals	2026-09-15T17:20:39.6547894Z - c) Las remuneraciones en concepto de gastos de representación
evals	Ejecutar las evals	2026-09-15T17:20:39.6548483Z - d) Las dietas y asignaciones para gastos de viaje (con excepciones en locomoción y manutención)
evals	Ejecutar las evals	2026-09-15T17:20:39.6549213Z - e) y f) Contribuciones y aportaciones a planes de pensiones y seguros de previsión social
evals	Ejecutar las evals	2026-09-15T17:20:39.6549552Z 
evals	Ejecutar las evals	2026-09-15T17:20:39.6549961Z **2. En el apartado 2 se amplía a prestaciones que en todo caso tienen consideración de rendimientos del trabajo:**
evals	Ejecutar las evals	2026-09-15T17:20:39.6550883Z - **2.a)** Pensiones y haberes pasivos de Seguridad Social, mutualidades, planes de pensiones, seguros de previsión social, planes de dependencia
evals	Ejecutar las evals	2026-09-15T17:20:39.6551662Z - **2.b)** Cantidades abonadas a diputados, senadores, concejales y miembros de entidades locales
evals	Ejecutar las evals	2026-09-15T17:20:39.6552156Z - **2.c)** Rendimientos de cursos, conferencias y seminarios
evals	Ejecutar las evals	2026-09-15T17:20:39.6552713Z - **2.d)** Rendimientos de obras literarias, artísticas o científicas (con cesión de explotación)
evals	Ejecutar las evals	2026-09-15T17:20:39.6553232Z - **2.e)** Retribuciones de administradores y miembros de consejos
evals	Ejecutar las evals	2026-09-15T17:20:39.6553636Z - **2.f)** Pensiones compensatorias y anualidades por alimentos
evals	Ejecutar las evals	2026-09-15T17:20:39.6554099Z - **2.g)** Derechos especiales de contenido económico en sociedades
evals	Ejecutar las evals	2026-09-15T17:20:39.6554434Z - **2.h)** Becas
evals	Ejecutar las evals	2026-09-15T17:20:39.6554845Z - **2.i)** Retribuciones por actividades humanitarias de entidades sin ánimo de lucro
evals	Ejecutar las evals	2026-09-15T17:20:39.6555302Z - **2.j)** Retribuciones de relaciones laborales especiales
evals	Ejecutar las evals	2026-09-15T17:20:39.6555729Z - **2.k)** Aportaciones al patrimonio protegido de personas con discapacidad
evals	Ejecutar las evals	2026-09-15T17:20:39.6556095Z 
evals	Ejecutar las evals	2026-09-15T17:20:39.6556241Z [art. 17 de la Ley 35/2006, BOE-A-2006-20764, bloque a17]
evals	Ejecutar las evals	2026-09-15T17:20:39.6556525Z ```
evals	Ejecutar las evals	2026-09-15T17:20:39.6556627Z 
evals	Ejecutar las evals	2026-09-15T17:20:39.6556783Z ## Sesión 07-lrjsp-principio-de-legalidad
evals	Ejecutar las evals	2026-09-15T17:20:39.6556978Z 
evals	Ejecutar las evals	2026-09-15T17:20:39.6557105Z Eval: 07-lrjsp-principio-de-legalidad.yaml
evals	Ejecutar las evals	2026-09-15T17:20:39.6557304Z 
evals	Ejecutar las evals	2026-09-15T17:20:39.6557383Z Pregunta:
evals	Ejecutar las evals	2026-09-15T17:20:39.6557490Z 
evals	Ejecutar las evals	2026-09-15T17:20:39.6557585Z ```text
evals	Ejecutar las evals	2026-09-15T17:20:39.6558051Z ¿Qué dice el artículo 25 de la Ley 40/2015 sobre el principio de legalidad en la potestad sancionadora?
evals	Ejecutar las evals	2026-09-15T17:20:39.6558483Z ```
evals	Ejecutar las evals	2026-09-15T17:20:39.6558584Z 
evals	Ejecutar las evals	2026-09-15T17:20:39.6558668Z Invocaciones:
evals	Ejecutar las evals	2026-09-15T17:20:39.6558785Z 
evals	Ejecutar las evals	2026-09-15T17:20:39.6558924Z | Orden | Código | Conexiones |
evals	Ejecutar las evals	2026-09-15T17:20:39.6559229Z | --- | --- | --- |
evals	Ejecutar las evals	2026-09-15T17:20:39.6559568Z | boe indice BOE-A-2015-10566 --json | 0 | sin conexiones |
evals	Ejecutar las evals	2026-09-15T17:20:39.6559971Z | boe articulo BOE-A-2015-10566 a25 --json | 0 | sin conexiones |
evals	Ejecutar las evals	2026-09-15T17:20:39.6560213Z 
evals	Ejecutar las evals	2026-09-15T17:20:39.6560301Z Respuesta:
evals	Ejecutar las evals	2026-09-15T17:20:39.6560414Z 
evals	Ejecutar las evals	2026-09-15T17:20:39.6560489Z ```text
evals	Ejecutar las evals	2026-09-15T17:20:39.6561332Z El artículo 25 de la Ley 40/2015, titulado **"Principio de legalidad"**, establece los requisitos fundamentales para el ejercicio de la potestad sancionadora por parte de las Administraciones Públicas:
evals	Ejecutar las evals	2026-09-15T17:20:39.6561994Z 
evals	Ejecutar las evals	2026-09-15T17:20:39.6562091Z **Requisitos esenciales:**
evals	Ejecutar las evals	2026-09-15T17:20:39.6562247Z 
evals	Ejecutar las evals	2026-09-15T17:20:39.6562901Z 1. **Habilitación legal expresa**: La potestad sancionadora debe estar reconocida expresamente por una norma con rango de Ley [art. 25.1, Ley 40/2015, BOE-A-2015-10566, bloque a25]
evals	Ejecutar las evals	2026-09-15T17:20:39.6563492Z 
evals	Ejecutar las evals	2026-09-15T17:20:39.6564788Z 2. **Procedimiento regulado**: El ejercicio debe ajustarse al procedimiento previsto en la propia habilitación legal y conforme a lo establecido en la Ley 40/2015 y en la Ley 39/2015 (Procedimiento Administrativo Común). Para Entidades Locales, se aplica adicionalmente el Título XI de la Ley 7/1985 [art. 25.1, Ley 40/2015, BOE-A-2015-10566, bloque a25]
evals	Ejecutar las evals	2026-09-15T17:20:39.6565898Z 
evals	Ejecutar las evals	2026-09-15T17:20:39.6566763Z 3. **Atribución orgánica**: Solo los órganos administrativos que tengan expresamente atribuida la potestad, ya sea por disposición de rango legal o reglamentario, pueden ejercerla [art. 25.2, Ley 40/2015, BOE-A-2015-10566, bloque a25]
evals	Ejecutar las evals	2026-09-15T17:20:39.6567502Z 
evals	Ejecutar las evals	2026-09-15T17:20:39.6567635Z **Extensiones y límites:**
evals	Ejecutar las evals	2026-09-15T17:20:39.6567789Z 
evals	Ejecutar las evals	2026-09-15T17:20:39.6568701Z - Las disposiciones del Capítulo III (donde está ubicado este artículo) se aplican también al ejercicio de la potestad disciplinaria sobre el personal al servicio de las Administraciones Públicas [art. 25.3, Ley 40/2015, BOE-A-2015-10566, bloque a25]
evals	Ejecutar las evals	2026-09-15T17:20:39.6569573Z 
evals	Ejecutar las evals	2026-09-15T17:20:39.6570408Z - No se aplican estas disposiciones cuando se trata de sanciones respecto a personas vinculadas por relaciones contractuales de contratación pública o patrimoniales [art. 25.4, Ley 40/2015, BOE-A-2015-10566, bloque a25]
evals	Ejecutar las evals	2026-09-15T17:20:39.6571199Z ```
evals	Ejecutar las evals	2026-09-15T17:20:39.6571300Z 
evals	Ejecutar las evals	2026-09-15T17:20:39.6571459Z ## Sesión 08-ltaibg-plazo-de-resolucion
evals	Ejecutar las evals	2026-09-15T17:20:39.6571644Z 
evals	Ejecutar las evals	2026-09-15T17:20:39.6571770Z Eval: 08-ltaibg-plazo-de-resolucion.yaml
evals	Ejecutar las evals	2026-09-15T17:20:39.6571968Z 
evals	Ejecutar las evals	2026-09-15T17:20:39.6572046Z Pregunta:
evals	Ejecutar las evals	2026-09-15T17:20:39.6572156Z 
evals	Ejecutar las evals	2026-09-15T17:20:39.6572241Z ```text
evals	Ejecutar las evals	2026-09-15T17:20:39.6572782Z ¿En qué plazo hay que resolver una solicitud de acceso a la información pública según el artículo 20 de la Ley 19/2013?
evals	Ejecutar las evals	2026-09-15T17:20:39.6573265Z ```
evals	Ejecutar las evals	2026-09-15T17:20:39.6573367Z 
evals	Ejecutar las evals	2026-09-15T17:20:39.6573460Z Invocaciones:
evals	Ejecutar las evals	2026-09-15T17:20:39.6573571Z 
evals	Ejecutar las evals	2026-09-15T17:20:39.6573721Z | Orden | Código | Conexiones |
evals	Ejecutar las evals	2026-09-15T17:20:39.6573970Z | --- | --- | --- |
evals	Ejecutar las evals	2026-09-15T17:20:39.6574238Z | boe indice BOE-A-2013-12887 --json | 0 | sin conexiones |
evals	Ejecutar las evals	2026-09-15T17:20:39.6574646Z | boe articulo BOE-A-2013-12887 a20 --json | 0 | sin conexiones |
evals	Ejecutar las evals	2026-09-15T17:20:39.6574965Z 
evals	Ejecutar las evals	2026-09-15T17:20:39.6575054Z Respuesta:
evals	Ejecutar las evals	2026-09-15T17:20:39.6575158Z 
evals	Ejecutar las evals	2026-09-15T17:20:39.6575239Z ```text
evals	Ejecutar las evals	2026-09-15T17:20:39.6575723Z Según el artículo 20 de la Ley 19/2013, el plazo para resolver una solicitud de acceso a información pública es:
evals	Ejecutar las evals	2026-09-15T17:20:39.6576101Z 
evals	Ejecutar las evals	2026-09-15T17:20:39.6576198Z **Plazo base: un mes**
evals	Ejecutar las evals	2026-09-15T17:20:39.6576329Z 
evals	Ejecutar las evals	2026-09-15T17:20:39.6577234Z La resolución en la que se conceda o deniegue el acceso debe notificarse al solicitante en un plazo máximo de **un mes desde la recepción de la solicitud** por el órgano competente para resolver [art. 20.1 de la LTAIBG, BOE-A-2013-12887, bloque a20].
evals	Ejecutar las evals	2026-09-15T17:20:39.6578003Z 
evals	Ejecutar las evals	2026-09-15T17:20:39.6578144Z **Ampliación: hasta dos meses**
evals	Ejecutar las evals	2026-09-15T17:20:39.6578298Z 
evals	Ejecutar las evals	2026-09-15T17:20:39.6579347Z Este plazo puede ampliarse por **otro mes adicional** cuando el volumen o la complejidad de la información que se solicita así lo hagan necesario, siempre que se notifique previamente esta ampliación al solicitante [art. 20.1 de la LTAIBG, BOE-A-2013-12887, bloque a20].
evals	Ejecutar las evals	2026-09-15T17:20:39.6580280Z 
evals	Ejecutar las evals	2026-09-15T17:20:39.6580416Z **Consideración importante:**
evals	Ejecutar las evals	2026-09-15T17:20:39.6580581Z 
evals	Ejecutar las evals	2026-09-15T17:20:39.6581409Z Si transcurre el plazo máximo sin que se dicte y notifique una resolución expresa, se entiende que la solicitud ha sido desestimada (silencio administrativo negativo) [art. 20.4 de la LTAIBG, BOE-A-2013-12887, bloque a20].
evals	Ejecutar las evals	2026-09-15T17:20:39.6582191Z ```
evals	Ejecutar las evals	2026-09-15T17:20:39.6582295Z 
evals	Ejecutar las evals	2026-09-15T17:20:39.6582448Z ## Sesión 09-constitucion-articulo-140
evals	Ejecutar las evals	2026-09-15T17:20:39.6582630Z 
evals	Ejecutar las evals	2026-09-15T17:20:39.6582749Z Eval: 09-constitucion-articulo-140.yaml
evals	Ejecutar las evals	2026-09-15T17:20:39.6582932Z 
evals	Ejecutar las evals	2026-09-15T17:20:39.6583010Z Pregunta:
evals	Ejecutar las evals	2026-09-15T17:20:39.6583118Z 
evals	Ejecutar las evals	2026-09-15T17:20:39.6583198Z ```text
evals	Ejecutar las evals	2026-09-15T17:20:39.6583458Z ¿Qué dice el artículo 140 de la Constitución?
evals	Ejecutar las evals	2026-09-15T17:20:39.6583739Z ```
evals	Ejecutar las evals	2026-09-15T17:20:39.6583836Z 
evals	Ejecutar las evals	2026-09-15T17:20:39.6583921Z Invocaciones:
evals	Ejecutar las evals	2026-09-15T17:20:39.6584037Z 
evals	Ejecutar las evals	2026-09-15T17:20:39.6584181Z | Orden | Código | Conexiones |
evals	Ejecutar las evals	2026-09-15T17:20:39.6584419Z | --- | --- | --- |
evals	Ejecutar las evals	2026-09-15T17:20:39.6584686Z | boe indice BOE-A-1978-31229 --json | 0 | sin conexiones |
evals	Ejecutar las evals	2026-09-15T17:20:39.6585208Z | boe articulo BOE-A-1978-31229 a140 --json | 0 | sin conexiones |
evals	Ejecutar las evals	2026-09-15T17:20:39.6585552Z 
evals	Ejecutar las evals	2026-09-15T17:20:39.6585685Z Respuesta:
evals	Ejecutar las evals	2026-09-15T17:20:39.6585852Z 
evals	Ejecutar las evals	2026-09-15T17:20:39.6585962Z ```text
evals	Ejecutar las evals	2026-09-15T17:20:39.6586504Z El artículo 140 de la Constitución Española establece:
evals	Ejecutar las evals	2026-09-15T17:20:39.6586892Z 
evals	Ejecutar las evals	2026-09-15T17:20:39.6587045Z **Contenido:**
evals	Ejecutar las evals	2026-09-15T17:20:39.6587257Z 
evals	Ejecutar las evals	2026-09-15T17:20:39.6588582Z La Constitución garantiza la autonomía de los municipios, que gozan de personalidad jurídica plena. El gobierno y administración de los municipios corresponde a sus respectivos Ayuntamientos, integrados por los Alcaldes y los Concejales. 
evals	Ejecutar las evals	2026-09-15T17:20:39.6589472Z 
evals	Ejecutar las evals	2026-09-15T17:20:39.6590491Z Los Concejales son elegidos por los vecinos del municipio mediante **sufragio universal, igual, libre, directo y secreto**, en la forma establecida por la ley. Por su parte, los Alcaldes pueden ser elegidos por los Concejales o por los vecinos, según lo que determine la ley.
evals	Ejecutar las evals	2026-09-15T17:20:39.6591381Z 
evals	Ejecutar las evals	2026-09-15T17:20:39.6591714Z La ley también regulará las condiciones en las que proceda el **régimen del concejo abierto**.
evals	Ejecutar las evals	2026-09-15T17:20:39.6592062Z 
evals	Ejecutar las evals	2026-09-15T17:20:39.6592310Z [art. 140 de la Constitución Española, BOE-A-1978-31229, bloque a140]
evals	Ejecutar las evals	2026-09-15T17:20:39.6592561Z 
evals	Ejecutar las evals	2026-09-15T17:20:39.6593558Z Este artículo es fundamental en el sistema de autonomía local español, ya que reconoce el derecho de los municipios a la autonomía, establece su estructura administrativa básica y garantiza la participación democrática de los vecinos en la elección de sus representantes.
evals	Ejecutar las evals	2026-09-15T17:20:39.6594501Z ```
evals	Ejecutar las evals	2026-09-15T17:20:39.6594597Z 
evals	Ejecutar las evals	2026-09-15T17:20:39.6594746Z ## Sesión 10-et-vacaciones
evals	Ejecutar las evals	2026-09-15T17:20:39.6594898Z 
evals	Ejecutar las evals	2026-09-15T17:20:39.6594994Z Eval: 10-et-vacaciones.yaml
evals	Ejecutar las evals	2026-09-15T17:20:39.6595149Z 
evals	Ejecutar las evals	2026-09-15T17:20:39.6595408Z Pregunta:
evals	Ejecutar las evals	2026-09-15T17:20:39.6595513Z 
evals	Ejecutar las evals	2026-09-15T17:20:39.6595601Z ```text
evals	Ejecutar las evals	2026-09-15T17:20:39.6596049Z ¿Cuántos días de vacaciones anuales reconoce el artículo 38 del Estatuto de los Trabajadores?
evals	Ejecutar las evals	2026-09-15T17:20:39.6596481Z ```
evals	Ejecutar las evals	2026-09-15T17:20:39.6596579Z 
evals	Ejecutar las evals	2026-09-15T17:20:39.6596676Z Invocaciones:
evals	Ejecutar las evals	2026-09-15T17:20:39.6596804Z 
evals	Ejecutar las evals	2026-09-15T17:20:39.6596938Z | Orden | Código | Conexiones |
evals	Ejecutar las evals	2026-09-15T17:20:39.6597192Z | --- | --- | --- |
evals	Ejecutar las evals	2026-09-15T17:20:39.6597469Z | boe indice BOE-A-2015-11430 --json | 0 | sin conexiones |
evals	Ejecutar las evals	2026-09-15T17:20:39.6597873Z | boe articulo BOE-A-2015-11430 a38 --json | 0 | sin conexiones |
evals	Ejecutar las evals	2026-09-15T17:20:39.6598117Z 
evals	Ejecutar las evals	2026-09-15T17:20:39.6598196Z Respuesta:
evals	Ejecutar las evals	2026-09-15T17:20:39.6598304Z 
evals	Ejecutar las evals	2026-09-15T17:20:39.6598384Z ```text
evals	Ejecutar las evals	2026-09-15T17:20:39.6598948Z El artículo 38 del Estatuto de los Trabajadores reconoce **como mínimo treinta días naturales de vacaciones anuales retribuidas**.
evals	Ejecutar las evals	2026-09-15T17:20:39.6599531Z 
evals	Ejecutar las evals	2026-09-15T17:20:39.6600869Z Según el artículo: «El periodo de vacaciones anuales retribuidas, no sustituible por compensación económica, será el pactado en convenio colectivo o contrato individual. En ningún caso la duración será inferior a treinta días naturales.» [Real Decreto Legislativo 2/2015 de 23 de octubre (Estatuto de los Trabajadores), art. 38.1, BOE-A-2015-11430, bloque a38]
evals	Ejecutar las evals	2026-09-15T17:20:39.6601993Z 
evals	Ejecutar las evals	2026-09-15T17:20:39.6602798Z Esta es la duración **mínima legal**. Las partes (empresa y trabajador) pueden acordar un mayor número de días a través del convenio colectivo o el contrato individual, pero no pueden pactar menos de treinta días naturales.
evals	Ejecutar las evals	2026-09-15T17:20:39.6603591Z ```
evals	Ejecutar las evals	2026-09-15T17:20:39.6603685Z 
evals	Ejecutar las evals	2026-09-15T17:20:39.6603840Z ## Sesión 11-no-activa-programacion
evals	Ejecutar las evals	2026-09-15T17:20:39.6604022Z 
evals	Ejecutar las evals	2026-09-15T17:20:39.6604133Z Eval: 11-no-activa-programacion.yaml
evals	Ejecutar las evals	2026-09-15T17:20:39.6604327Z 
evals	Ejecutar las evals	2026-09-15T17:20:39.6604413Z Pregunta:
evals	Ejecutar las evals	2026-09-15T17:20:39.6604518Z 
evals	Ejecutar las evals	2026-09-15T17:20:39.6604596Z ```text
evals	Ejecutar las evals	2026-09-15T17:20:39.6604837Z ¿Cómo invierto una lista enlazada en Go?
evals	Ejecutar las evals	2026-09-15T17:20:39.6605101Z ```
evals	Ejecutar las evals	2026-09-15T17:20:39.6605199Z 
evals	Ejecutar las evals	2026-09-15T17:20:39.6605294Z Invocaciones: ninguna
evals	Ejecutar las evals	2026-09-15T17:20:39.6605423Z 
evals	Ejecutar las evals	2026-09-15T17:20:39.6605508Z Respuesta:
evals	Ejecutar las evals	2026-09-15T17:20:39.6605615Z 
evals	Ejecutar las evals	2026-09-15T17:20:39.6605697Z ````text
evals	Ejecutar las evals	2026-09-15T17:20:39.6606305Z Para invertir una lista enlazada en Go, puedes hacerlo de forma iterativa o recursiva. Aquí te muestro la solución más común (iterativa):
evals	Ejecutar las evals	2026-09-15T17:20:39.6606789Z 
evals	Ejecutar las evals	2026-09-15T17:20:39.6606861Z ```go
evals	Ejecutar las evals	2026-09-15T17:20:39.6607055Z type Node struct {
evals	Ejecutar las evals	2026-09-15T17:20:39.6607279Z     Val  int
evals	Ejecutar las evals	2026-09-15T17:20:39.6607484Z     Next *Node
evals	Ejecutar las evals	2026-09-15T17:20:39.6607668Z }
evals	Ejecutar las evals	2026-09-15T17:20:39.6607770Z 
evals	Ejecutar las evals	2026-09-15T17:20:39.6607883Z func ReverseList(head *Node) *Node {
evals	Ejecutar las evals	2026-09-15T17:20:39.6608154Z     var prev *Node
evals	Ejecutar las evals	2026-09-15T17:20:39.6608383Z     current := head
evals	Ejecutar las evals	2026-09-15T17:20:39.6608580Z     
evals	Ejecutar las evals	2026-09-15T17:20:39.6608787Z     for current != nil {
evals	Ejecutar las evals	2026-09-15T17:20:39.6609130Z         // Guardar el siguiente nodo
evals	Ejecutar las evals	2026-09-15T17:20:39.6609440Z         next := current.Next
evals	Ejecutar las evals	2026-09-15T17:20:39.6609687Z         
evals	Ejecutar las evals	2026-09-15T17:20:39.6609916Z         // Invertir el enlace
evals	Ejecutar las evals	2026-09-15T17:20:39.6610192Z         current.Next = prev
evals	Ejecutar las evals	2026-09-15T17:20:39.6610436Z         
evals	Ejecutar las evals	2026-09-15T17:20:39.6610640Z         // Avanzar
evals	Ejecutar las evals	2026-09-15T17:20:39.6610866Z         prev = current
evals	Ejecutar las evals	2026-09-15T17:20:39.6611114Z         current = next
evals	Ejecutar las evals	2026-09-15T17:20:39.6611335Z     }
evals	Ejecutar las evals	2026-09-15T17:20:39.6611509Z     
evals	Ejecutar las evals	2026-09-15T17:20:39.6611699Z     return prev
evals	Ejecutar las evals	2026-09-15T17:20:39.6611886Z }
evals	Ejecutar las evals	2026-09-15T17:20:39.6612047Z ```
evals	Ejecutar las evals	2026-09-15T17:20:39.6612147Z 
evals	Ejecutar las evals	2026-09-15T17:20:39.6612259Z **Explicación:**
evals	Ejecutar las evals	2026-09-15T17:20:39.6612554Z 1. `prev` comienza en `nil` (será la cola del final)
evals	Ejecutar las evals	2026-09-15T17:20:39.6612966Z 2. Para cada nodo, guardamos el siguiente antes de modificar el enlace
evals	Ejecutar las evals	2026-09-15T17:20:39.6613485Z 3. Hacemos que el nodo actual apunte hacia atrás (`current.Next = prev`)
evals	Ejecutar las evals	2026-09-15T17:20:39.6613879Z 4. Avanzamos con `prev = current` y `current = next`
evals	Ejecutar las evals	2026-09-15T17:20:39.6614289Z 5. Retornamos `prev` (nuevo cabecera)
evals	Ejecutar las evals	2026-09-15T17:20:39.6614457Z 
evals	Ejecutar las evals	2026-09-15T17:20:39.6614551Z **Ejemplo de uso:**
evals	Ejecutar las evals	2026-09-15T17:20:39.6614746Z ```go
evals	Ejecutar las evals	2026-09-15T17:20:39.6614942Z // Crear lista: 1 -> 2 -> 3
evals	Ejecutar las evals	2026-09-15T17:20:39.6615246Z head := &Node{Val: 1, Next: &Node{Val: 2, Next: &Node{Val: 3}}}
evals	Ejecutar las evals	2026-09-15T17:20:39.6615479Z 
evals	Ejecutar las evals	2026-09-15T17:20:39.6615571Z // Invertir: 3 -> 2 -> 1
evals	Ejecutar las evals	2026-09-15T17:20:39.6615808Z head = ReverseList(head)
evals	Ejecutar las evals	2026-09-15T17:20:39.6616012Z ```
evals	Ejecutar las evals	2026-09-15T17:20:39.6616115Z 
evals	Ejecutar las evals	2026-09-15T17:20:39.6616198Z **Complejidad:**
evals	Ejecutar las evals	2026-09-15T17:20:39.6616438Z - **Tiempo:** O(n) - recorremos una sola vez
evals	Ejecutar las evals	2026-09-15T17:20:39.6616760Z - **Espacio:** O(1) - solo usamos un par de punteros
evals	Ejecutar las evals	2026-09-15T17:20:39.6616972Z 
evals	Ejecutar las evals	2026-09-15T17:20:39.6617113Z Si prefieres **solución recursiva**:
evals	Ejecutar las evals	2026-09-15T17:20:39.6617424Z ```go
evals	Ejecutar las evals	2026-09-15T17:20:39.6617655Z func ReverseListRecursive(head *Node) *Node {
evals	Ejecutar las evals	2026-09-15T17:20:39.6617989Z     if head == nil || head.Next == nil {
evals	Ejecutar las evals	2026-09-15T17:20:39.6618277Z         return head
evals	Ejecutar las evals	2026-09-15T17:20:39.6618475Z     }
evals	Ejecutar las evals	2026-09-15T17:20:39.6618640Z     
evals	Ejecutar las evals	2026-09-15T17:20:39.6618891Z     newHead := ReverseListRecursive(head.Next)
evals	Ejecutar las evals	2026-09-15T17:20:39.6619281Z     head.Next.Next = head
evals	Ejecutar las evals	2026-09-15T17:20:39.6619527Z     head.Next = nil
evals	Ejecutar las evals	2026-09-15T17:20:39.6619726Z     
evals	Ejecutar las evals	2026-09-15T17:20:39.6619930Z     return newHead
evals	Ejecutar las evals	2026-09-15T17:20:39.6620138Z }
evals	Ejecutar las evals	2026-09-15T17:20:39.6620327Z ```
evals	Ejecutar las evals	2026-09-15T17:20:39.6620433Z 
evals	Ejecutar las evals	2026-09-15T17:20:39.6620679Z La iterativa es más eficiente en memoria (no usa pila de recursión).
evals	Ejecutar las evals	2026-09-15T17:20:39.6621021Z ````
evals	Ejecutar las evals	2026-09-15T17:20:39.6621115Z 
evals	Ejecutar las evals	2026-09-15T17:20:39.6621286Z ## Sesión 12-no-activa-acuerdo-entre-amigos
evals	Ejecutar las evals	2026-09-15T17:20:39.6621489Z 
evals	Ejecutar las evals	2026-09-15T17:20:39.6621617Z Eval: 12-no-activa-acuerdo-entre-amigos.yaml
evals	Ejecutar las evals	2026-09-15T17:20:39.6621817Z 
evals	Ejecutar las evals	2026-09-15T17:20:39.6621912Z Pregunta:
evals	Ejecutar las evals	2026-09-15T17:20:39.6622016Z 
evals	Ejecutar las evals	2026-09-15T17:20:39.6622100Z ```text
evals	Ejecutar las evals	2026-09-15T17:20:39.6622804Z Reescribe en un tono más cercano esta frase de un acuerdo entre amigos para compartir coche: «Las partes se turnarán el uso del vehículo en fines de semana alternos».
evals	Ejecutar las evals	2026-09-15T17:20:39.6623438Z ```
evals	Ejecutar las evals	2026-09-15T17:20:39.6623533Z 
evals	Ejecutar las evals	2026-09-15T17:20:39.6623632Z Invocaciones: ninguna
evals	Ejecutar las evals	2026-09-15T17:20:39.6623773Z 
evals	Ejecutar las evals	2026-09-15T17:20:39.6623859Z Respuesta:
evals	Ejecutar las evals	2026-09-15T17:20:39.6623970Z 
evals	Ejecutar las evals	2026-09-15T17:20:39.6624054Z ```text
evals	Ejecutar las evals	2026-09-15T17:20:39.6624306Z Aquí tienes algunas opciones más cercanas:
evals	Ejecutar las evals	2026-09-15T17:20:39.6624497Z 
evals	Ejecutar las evals	2026-09-15T17:20:39.6624611Z **Más informal:**
evals	Ejecutar las evals	2026-09-15T17:20:39.6624857Z - "Nos alternamos el coche cada fin de semana"
evals	Ejecutar las evals	2026-09-15T17:20:39.6625262Z - "Cada uno disfruta del coche un fin de semana sí, otro no"
evals	Ejecutar las evals	2026-09-15T17:20:39.6625755Z - "Nos pasamos el coche: un fin de semana es para ti, el siguiente para mí"
evals	Ejecutar las evals	2026-09-15T17:20:39.6626036Z 
evals	Ejecutar las evals	2026-09-15T17:20:39.6626156Z **Más coloquial:**
evals	Ejecutar las evals	2026-09-15T17:20:39.6626399Z - "Uno cada fin de semana, alternando"
evals	Ejecutar las evals	2026-09-15T17:20:39.6626775Z - "Nos turnamos el coche entre fin de semana"
evals	Ejecutar las evals	2026-09-15T17:20:39.6627172Z - "El coche va rotando: un fin de semana para uno, el siguiente para el otro"
evals	Ejecutar las evals	2026-09-15T17:20:39.6627452Z 
evals	Ejecutar las evals	2026-09-15T17:20:39.6627564Z **Si quieres algo muy natural:**
evals	Ejecutar las evals	2026-09-15T17:20:39.6627870Z - "Nos repartimos el coche en fines de semana alternos"
evals	Ejecutar las evals	2026-09-15T17:20:39.6628097Z 
evals	Ejecutar las evals	2026-09-15T17:20:39.6628783Z La que suena más amistosa y directa sería: **"Nos alternamos el coche cada fin de semana"** o **"Uno cada fin de semana, alternando"**. Mantienen la claridad legal pero con un tono más relajado.
evals	Ejecutar las evals	2026-09-15T17:20:39.6629525Z ```
evals	Ejecutar las evals	2026-09-15T17:20:39.6629731Z --- fin de informe.md ---
evals	Ejecutar las evals	2026-09-15T17:20:39.6630022Z --- inicio de informe.json ---
evals	Ejecutar las evals	2026-09-15T17:20:39.6630255Z {
evals	Ejecutar las evals	2026-09-15T17:20:39.6630469Z   "skill": "boe-legislacion",
evals	Ejecutar las evals	2026-09-15T17:20:39.6630759Z   "modelo": "claude-haiku-4-5-20251001",
evals	Ejecutar las evals	2026-09-15T17:20:39.6631055Z   "modelos_de_sesion": [
evals	Ejecutar las evals	2026-09-15T17:20:39.6631317Z     "claude-haiku-4-5-20251001"
evals	Ejecutar las evals	2026-09-15T17:20:39.6631545Z   ],
evals	Ejecutar las evals	2026-09-15T17:20:39.6631761Z   "versiones_de_claude_code": [
evals	Ejecutar las evals	2026-09-15T17:20:39.6632102Z     "2.1.270"
evals	Ejecutar las evals	2026-09-15T17:20:39.6632283Z   ],
evals	Ejecutar las evals	2026-09-15T17:20:39.6632546Z   "commit": "091facd82d3744dffcfb3b17115c5ea25452d7d5",
evals	Ejecutar las evals	2026-09-15T17:20:39.6633952Z   "sin_python": "búsqueda: find / ( -path /proc -o -path /sys ) -prune -o ( ( -type f -perm /111 ( -iname python* -o -iname pypy* ) ) -o ( -type l ( -iname python* -o -iname pypy* ) ) -o ( ( -type f -o -type l ) ( -iname libpython* -o -iname libpypy* ) ) ) -print\nusuario: root\nresultado: ninguno\n",
evals	Ejecutar las evals	2026-09-15T17:20:39.6634976Z   "ficheros_mal_formados": [],
evals	Ejecutar las evals	2026-09-15T17:20:39.6635248Z   "veredicto": "aprobado",
evals	Ejecutar las evals	2026-09-15T17:20:39.6635479Z   "motivos": [],
evals	Ejecutar las evals	2026-09-15T17:20:39.6635699Z   "fuera_de_lo_grabado": [
evals	Ejecutar las evals	2026-09-15T17:20:39.6635912Z     {
evals	Ejecutar las evals	2026-09-15T17:20:39.6636215Z       "sesion": "01-lpac-articulo-21-prueba-de-red",
evals	Ejecutar las evals	2026-09-15T17:20:39.6636639Z       "eval": "01-lpac-articulo-21.yaml",
evals	Ejecutar las evals	2026-09-15T17:20:39.6637034Z       "orden": "boe articulo BOE-A-2015-10565 a9998 --json",
evals	Ejecutar las evals	2026-09-15T17:20:39.6637375Z       "codigo": 5
evals	Ejecutar las evals	2026-09-15T17:20:39.6637573Z     },
evals	Ejecutar las evals	2026-09-15T17:20:39.6637745Z     {
evals	Ejecutar las evals	2026-09-15T17:20:39.6638037Z       "sesion": "01-lpac-articulo-21-prueba-de-red",
evals	Ejecutar las evals	2026-09-15T17:20:39.6638406Z       "eval": "01-lpac-articulo-21.yaml",
evals	Ejecutar las evals	2026-09-15T17:20:39.6638837Z       "orden": "boe articulo BOE-A-2015-10565 a9998 --offline --json",
evals	Ejecutar las evals	2026-09-15T17:20:39.6639254Z       "codigo": 4
evals	Ejecutar las evals	2026-09-15T17:20:39.6639456Z     },
evals	Ejecutar las evals	2026-09-15T17:20:39.6639624Z     {
evals	Ejecutar las evals	2026-09-15T17:20:39.6639910Z       "sesion": "05-trlrhl-impuestos-municipales",
evals	Ejecutar las evals	2026-09-15T17:20:39.6640319Z       "eval": "05-trlrhl-impuestos-municipales.yaml",
evals	Ejecutar las evals	2026-09-15T17:20:39.6640842Z       "orden": "boe buscar texto refundido ley reguladora haciendas locales --json",
evals	Ejecutar las evals	2026-09-15T17:20:39.6641236Z       "codigo": 5
evals	Ejecutar las evals	2026-09-15T17:20:39.6641432Z     }
evals	Ejecutar las evals	2026-09-15T17:20:39.6641605Z   ],
evals	Ejecutar las evals	2026-09-15T17:20:39.6641784Z   "red": [],
evals	Ejecutar las evals	2026-09-15T17:20:39.6641976Z   "evals": [
evals	Ejecutar las evals	2026-09-15T17:20:39.6642161Z     {
evals	Ejecutar las evals	2026-09-15T17:20:39.6642403Z       "sesion": "01-lpac-articulo-21",
evals	Ejecutar las evals	2026-09-15T17:20:39.6642726Z       "eval": "01-lpac-articulo-21.yaml",
evals	Ejecutar las evals	2026-09-15T17:20:39.6643019Z       "activa": true,
evals	Ejecutar las evals	2026-09-15T17:20:39.6643259Z       "activada": true,
evals	Ejecutar las evals	2026-09-15T17:20:39.6643525Z       "comandos_ejecutados": [
evals	Ejecutar las evals	2026-09-15T17:20:39.6643838Z         "bloque boe BOE-A-2015-10565 a21"
evals	Ejecutar las evals	2026-09-15T17:20:39.6644109Z       ],
evals	Ejecutar las evals	2026-09-15T17:20:39.6644336Z       "comandos_ausentes": [],
evals	Ejecutar las evals	2026-09-15T17:20:39.6644614Z       "citas_encontradas": [
evals	Ejecutar las evals	2026-09-15T17:20:39.6644899Z         "BOE-A-2015-10565 a21"
evals	Ejecutar las evals	2026-09-15T17:20:39.6645127Z       ],
evals	Ejecutar las evals	2026-09-15T17:20:39.6645346Z       "citas_ausentes": [],
evals	Ejecutar las evals	2026-09-15T17:20:39.6645619Z       "invocaciones": [
evals	Ejecutar las evals	2026-09-15T17:20:39.6645849Z         {
evals	Ejecutar las evals	2026-09-15T17:20:39.6646176Z           "orden": "boe indice BOE-A-2015-10565 --json",
evals	Ejecutar las evals	2026-09-15T17:20:39.6646500Z           "codigo": 0,
evals	Ejecutar las evals	2026-09-15T17:20:39.6646772Z           "conexiones": []
evals	Ejecutar las evals	2026-09-15T17:20:39.6647002Z         },
evals	Ejecutar las evals	2026-09-15T17:20:39.6647192Z         {
evals	Ejecutar las evals	2026-09-15T17:20:39.6647548Z           "orden": "boe articulo BOE-A-2015-10565 a21 --json",
evals	Ejecutar las evals	2026-09-15T17:20:39.6647878Z           "codigo": 0,
evals	Ejecutar las evals	2026-09-15T17:20:39.6648137Z           "conexiones": []
evals	Ejecutar las evals	2026-09-15T17:20:39.6648361Z         }
evals	Ejecutar las evals	2026-09-15T17:20:39.6648552Z       ],
evals	Ejecutar las evals	2026-09-15T17:20:39.6648783Z       "fuera_de_lo_grabado": [],
evals	Ejecutar las evals	2026-09-15T17:20:39.6649121Z       "otras_fallidas": [],
evals	Ejecutar las evals	2026-09-15T17:20:39.6649394Z       "llegadas_a_la_red": [],
evals	Ejecutar las evals	2026-09-15T17:20:39.6660713Z       "respuesta": "El **art. 21 de la Ley 39/2015** regula la **obligación de la Administración de resolver y notificar**. Sus apartados principales son:\n\n**1. Obligación de resolución expresa y notificación:** La Administración debe dictar resolución expresa y notificarla en todos los procedimientos, cualquiera que sea su forma de iniciación. Cuando concurran circunstancias como prescripción, renuncia, caducidad o desistimiento, la resolución debe declarar la circunstancia con indicación de hechos y normas. Se exceptúan los procedimientos que terminan por pacto o convenio, y los relativos a derechos sometidos solo a declaración responsable o comunicación [art. 21.1 de la Ley 39/2015, BOE-A-2015-10565, bloque a21].\n\n**2. Plazo máximo general:** El plazo máximo de notificación será el que fije la norma reguladora de cada procedimiento. Este plazo no puede exceder de **seis meses**, salvo que una norma de rango de Ley establezca uno mayor o lo prevea el Derecho de la Unión Europea [art. 21.2 de la Ley 39/2015, BOE-A-2015-10565, bloque a21].\n\n**3. Plazo por defecto:** Cuando no haya plazo establecido, será de **tres meses**. Se cuenta desde la fecha del acuerdo de iniciación (procedimientos de oficio) o desde la entrada de la solicitud en el registro electrónico (procedimientos iniciados a solicitud) [art. 21.3 de la Ley 39/2015, BOE-A-2015-10565, bloque a21].\n\n**4. Publicidad y comunicación:** Las Administraciones deben publicar en su portal web las relaciones de procedimientos con sus plazos máximos y los efectos del silencio administrativo. También informarán al interesado del plazo máximo, incluyéndolo en el acuerdo de iniciación o en comunicación dentro de diez días desde la recepción de la solicitud [art. 21.4 de la Ley 39/2015, BOE-A-2015-10565, bloque a21].\n\n**5. Habilitación de medios:** Si el volumen de solicitudes podría incumplir el plazo máximo, el órgano competente o su superior jerárquico pueden habilitar medios personales y materiales adicionales [art. 21.5 de la Ley 39/2015, BOE-A-2015-10565, bloque a21].\n\n**6. Responsabilidad:** El personal y los titulares de órganos competentes son responsables del cumplimiento de la obligación de resolver en plazo. El incumplimiento genera responsabilidad disciplinaria [art. 21.6 de la Ley 39/2015, BOE-A-2015-10565, bloque a21].",
evals	Ejecutar las evals	2026-09-15T17:20:39.6668267Z       "codigo_de_la_sesion": 0,
evals	Ejecutar las evals	2026-09-15T17:20:39.6668588Z       "fin_de_la_sesion": "result success",
evals	Ejecutar las evals	2026-09-15T17:20:39.6668918Z       "sesion_terminada": true,
evals	Ejecutar las evals	2026-09-15T17:20:39.6669235Z       "motivos": [],
evals	Ejecutar las evals	2026-09-15T17:20:39.6669484Z       "pasa": true
evals	Ejecutar las evals	2026-09-15T17:20:39.6669687Z     },
evals	Ejecutar las evals	2026-09-15T17:20:39.6669862Z     {
evals	Ejecutar las evals	2026-09-15T17:20:39.6670156Z       "sesion": "01-lpac-articulo-21-prueba-de-red",
evals	Ejecutar las evals	2026-09-15T17:20:39.6670535Z       "eval": "01-lpac-articulo-21.yaml",
evals	Ejecutar las evals	2026-09-15T17:20:39.6670829Z       "activa": true,
evals	Ejecutar las evals	2026-09-15T17:20:39.6671070Z       "activada": true,
evals	Ejecutar las evals	2026-09-15T17:20:39.6671343Z       "comandos_ejecutados": [
evals	Ejecutar las evals	2026-09-15T17:20:39.6671655Z         "bloque boe BOE-A-2015-10565 a21"
evals	Ejecutar las evals	2026-09-15T17:20:39.6671914Z       ],
evals	Ejecutar las evals	2026-09-15T17:20:39.6672159Z       "comandos_ausentes": [],
evals	Ejecutar las evals	2026-09-15T17:20:39.6672438Z       "citas_encontradas": [
evals	Ejecutar las evals	2026-09-15T17:20:39.6672718Z         "BOE-A-2015-10565 a21"
evals	Ejecutar las evals	2026-09-15T17:20:39.6672943Z       ],
evals	Ejecutar las evals	2026-09-15T17:20:39.6673169Z       "citas_ausentes": [],
evals	Ejecutar las evals	2026-09-15T17:20:39.6673427Z       "invocaciones": [
evals	Ejecutar las evals	2026-09-15T17:20:39.6673649Z         {
evals	Ejecutar las evals	2026-09-15T17:20:39.6674016Z           "orden": "boe articulo BOE-A-2015-10565 a9998 --json",
evals	Ejecutar las evals	2026-09-15T17:20:39.6674361Z           "codigo": 5,
evals	Ejecutar las evals	2026-09-15T17:20:39.6674619Z           "conexiones": [
evals	Ejecutar las evals	2026-09-15T17:20:39.6674868Z             {
evals	Ejecutar las evals	2026-09-15T17:20:39.6675160Z               "destino": "127.0.0.1:9",
evals	Ejecutar las evals	2026-09-15T17:20:39.6675480Z               "clase": "local"
evals	Ejecutar las evals	2026-09-15T17:20:39.6675720Z             }
evals	Ejecutar las evals	2026-09-15T17:20:39.6675924Z           ]
evals	Ejecutar las evals	2026-09-15T17:20:39.6676114Z         },
evals	Ejecutar las evals	2026-09-15T17:20:39.6676297Z         {
evals	Ejecutar las evals	2026-09-15T17:20:39.6676705Z           "orden": "boe articulo BOE-A-2015-10565 a9998 --offline --json",
evals	Ejecutar las evals	2026-09-15T17:20:39.6677072Z           "codigo": 4,
evals	Ejecutar las evals	2026-09-15T17:20:39.6677328Z           "conexiones": []
evals	Ejecutar las evals	2026-09-15T17:20:39.6677555Z         },
evals	Ejecutar las evals	2026-09-15T17:20:39.6677749Z         {
evals	Ejecutar las evals	2026-09-15T17:20:39.6678072Z           "orden": "boe indice BOE-A-2015-10565 --json",
evals	Ejecutar las evals	2026-09-15T17:20:39.6678389Z           "codigo": 0,
evals	Ejecutar las evals	2026-09-15T17:20:39.6678656Z           "conexiones": []
evals	Ejecutar las evals	2026-09-15T17:20:39.6678885Z         },
evals	Ejecutar las evals	2026-09-15T17:20:39.6679125Z         {
evals	Ejecutar las evals	2026-09-15T17:20:39.6679470Z           "orden": "boe articulo BOE-A-2015-10565 a21 --json",
evals	Ejecutar las evals	2026-09-15T17:20:39.6679894Z           "codigo": 0,
evals	Ejecutar las evals	2026-09-15T17:20:39.6680157Z           "conexiones": []
evals	Ejecutar las evals	2026-09-15T17:20:39.6680379Z         }
evals	Ejecutar las evals	2026-09-15T17:20:39.6680573Z       ],
evals	Ejecutar las evals	2026-09-15T17:20:39.6680800Z       "fuera_de_lo_grabado": [
evals	Ejecutar las evals	2026-09-15T17:20:39.6681031Z         {
evals	Ejecutar las evals	2026-09-15T17:20:39.6681406Z           "orden": "boe articulo BOE-A-2015-10565 a9998 --json",
evals	Ejecutar las evals	2026-09-15T17:20:39.6681746Z           "codigo": 5
evals	Ejecutar las evals	2026-09-15T17:20:39.6681956Z         },
evals	Ejecutar las evals	2026-09-15T17:20:39.6682146Z         {
evals	Ejecutar las evals	2026-09-15T17:20:39.6682553Z           "orden": "boe articulo BOE-A-2015-10565 a9998 --offline --json",
evals	Ejecutar las evals	2026-09-15T17:20:39.6682918Z           "codigo": 4
evals	Ejecutar las evals	2026-09-15T17:20:39.6683131Z         }
evals	Ejecutar las evals	2026-09-15T17:20:39.6683319Z       ],
evals	Ejecutar las evals	2026-09-15T17:20:39.6683541Z       "otras_fallidas": [],
evals	Ejecutar las evals	2026-09-15T17:20:39.6683808Z       "llegadas_a_la_red": [],
evals	Ejecutar las evals	2026-09-15T17:20:39.6692020Z       "respuesta": "El artículo 21 de la Ley 39/2015 regulala **obligación de resolver** de la Administración. Sus puntos esenciales son:\n\n**1. Obligación de resolución expresa:** La Administración está obligada a dictar resolución expresa y notificarla en todos los procedimientos, cualquiera que sea su forma de iniciación. Se exceptúan los procedimientos que terminan por pacto, convenio o los sometidos solo a declaración responsable o comunicación [art. 21.1 de la Ley 39/2015, BOE-A-2015-10565, bloque a21].\n\n**2. Plazo máximo:** La resolución debe notificarse en el plazo que fije la norma reguladora del procedimiento, con un límite máximo de **seis meses**, salvo que una ley establezca un plazo mayor o lo prevea el Derecho de la Unión Europea. Si la norma no fija plazo, será de **tres meses** [art. 21.2 y 21.3 de la Ley 39/2015, BOE-A-2015-10565, bloque a21].\n\n**3. Conteo de plazos:** Se cuentan desde la fecha del acuerdo de iniciación (procedimientos de oficio) o desde la entrada de la solicitud en el registro electrónico (procedimientos a solicitud) [art. 21.3 de la Ley 39/2015, BOE-A-2015-10565, bloque a21].\n\n**4. Información a los interesados:** Las Administraciones deben informar de los plazos máximos y los efectos del silencio administrativo, incluyendo esta información en la iniciación del procedimiento o en comunicación dirigida en los diez días siguientes a la recepción de la solicitud [art. 21.4 de la Ley 39/2015, BOE-A-2015-10565, bloque a21].\n\n**5. Responsabilidad:** El incumplimiento de la obligación de dictar resolución en plazo genera responsabilidad disciplinaria [art. 21.6 de la Ley 39/2015, BOE-A-2015-10565, bloque a21].",
evals	Ejecutar las evals	2026-09-15T17:20:39.6697345Z       "codigo_de_la_sesion": 0,
evals	Ejecutar las evals	2026-09-15T17:20:39.6697665Z       "fin_de_la_sesion": "result success",
evals	Ejecutar las evals	2026-09-15T17:20:39.6697990Z       "sesion_terminada": true,
evals	Ejecutar las evals	2026-09-15T17:20:39.6698256Z       "motivos": [],
evals	Ejecutar las evals	2026-09-15T17:20:39.6698487Z       "pasa": true
evals	Ejecutar las evals	2026-09-15T17:20:39.6698709Z     },
evals	Ejecutar las evals	2026-09-15T17:20:39.6698895Z     {
evals	Ejecutar las evals	2026-09-15T17:20:39.6699260Z       "sesion": "02-lcsp-contrato-menor",
evals	Ejecutar las evals	2026-09-15T17:20:39.6699619Z       "eval": "02-lcsp-contrato-menor.yaml",
evals	Ejecutar las evals	2026-09-15T17:20:39.6699923Z       "activa": true,
evals	Ejecutar las evals	2026-09-15T17:20:39.6700155Z       "activada": true,
evals	Ejecutar las evals	2026-09-15T17:20:39.6700420Z       "comandos_ejecutados": [
evals	Ejecutar las evals	2026-09-15T17:20:39.6700740Z         "bloque boe BOE-A-2017-12902 a1-30"
evals	Ejecutar las evals	2026-09-15T17:20:39.6701002Z       ],
evals	Ejecutar las evals	2026-09-15T17:20:39.6701233Z       "comandos_ausentes": [],
evals	Ejecutar las evals	2026-09-15T17:20:39.6701515Z       "citas_encontradas": [
evals	Ejecutar las evals	2026-09-15T17:20:39.6701795Z         "BOE-A-2017-12902 a1-30"
evals	Ejecutar las evals	2026-09-15T17:20:39.6702029Z       ],
evals	Ejecutar las evals	2026-09-15T17:20:39.6702255Z       "citas_ausentes": [],
evals	Ejecutar las evals	2026-09-15T17:20:39.6702511Z       "invocaciones": [
evals	Ejecutar las evals	2026-09-15T17:20:39.6702724Z         {
evals	Ejecutar las evals	2026-09-15T17:20:39.6703047Z           "orden": "boe indice BOE-A-2017-12902 --json",
evals	Ejecutar las evals	2026-09-15T17:20:39.6703370Z           "codigo": 0,
evals	Ejecutar las evals	2026-09-15T17:20:39.6703628Z           "conexiones": []
evals	Ejecutar las evals	2026-09-15T17:20:39.6703859Z         },
evals	Ejecutar las evals	2026-09-15T17:20:39.6704051Z         {
evals	Ejecutar las evals	2026-09-15T17:20:39.6704404Z           "orden": "boe articulo BOE-A-2017-12902 a1-30 --json",
evals	Ejecutar las evals	2026-09-15T17:20:39.6704819Z           "codigo": 0,
evals	Ejecutar las evals	2026-09-15T17:20:39.6705091Z           "conexiones": []
evals	Ejecutar las evals	2026-09-15T17:20:39.6705318Z         }
evals	Ejecutar las evals	2026-09-15T17:20:39.6705499Z       ],
evals	Ejecutar las evals	2026-09-15T17:20:39.6705738Z       "fuera_de_lo_grabado": [],
evals	Ejecutar las evals	2026-09-15T17:20:39.6706028Z       "otras_fallidas": [],
evals	Ejecutar las evals	2026-09-15T17:20:39.6706302Z       "llegadas_a_la_red": [],
evals	Ejecutar las evals	2026-09-15T17:20:39.6715168Z       "respuesta": "Según el artículo 118 de la LCSP, el expediente de un contrato menor debe incluir los siguientes elementos:\n\n## Elementos obligatorios generales\n\n1. **Informe de justificación**: Un informe del órgano de contratación que justifique de manera motivada la necesidad del contrato y que no se está alterando su objeto con el fin de evitar la aplicación de los umbrales de contrato menor (40.000 euros para obras; 15.000 euros para suministros o servicios) [art. 118.2 de la LCSP, BOE-A-2017-12902, bloque a1-30]\n\n2. **Aprobación del gasto**: La aprobación presupuestaria del gasto [art. 118.3 de la LCSP, BOE-A-2017-12902, bloque a1-30]\n\n3. **Factura**: La incorporación de la factura correspondiente que reúna los requisitos que las normas de desarrollo de la LCSP establezcan [art. 118.3 de la LCSP, BOE-A-2017-12902, bloque a1-30]\n\n## Elemento adicional para contratos menores de obras\n\nPara los contratos menores de obras, además de lo anterior, debe incluirse:\n\n4. **Presupuesto de las obras**: Sin perjuicio de que deba existir el correspondiente proyecto cuando sea requerido por disposiciones vigentes [art. 118.4 de la LCSP, BOE-A-2017-12902, bloque a1-30]\n\n5. **Informe de supervisión** (cuando proceda): Si el trabajo afecta a la estabilidad, seguridad o estanqueidad de la obra, se debe solicitar el informe de las oficinas o unidades de supervisión a que se refiere el artículo 235 LCSP [art. 118.4 de la LCSP, BOE-A-2017-12902, bloque a1-30]\n\n## Excepciones\n\nEl requisito del informe de justificación (apartado 2) no se aplica en contratos cuyo pago se verifique a través de sistemas de anticipos de caja fija u otros similares para realizar pagos menores, cuando el valor estimado no exceda de 5.000 euros [art. 118.5 de la LCSP, BOE-A-2017-12902, bloque a1-30]",
evals	Ejecutar las evals	2026-09-15T17:20:39.6720952Z       "codigo_de_la_sesion": 0,
evals	Ejecutar las evals	2026-09-15T17:20:39.6721262Z       "fin_de_la_sesion": "result success",
evals	Ejecutar las evals	2026-09-15T17:20:39.6721591Z       "sesion_terminada": true,
evals	Ejecutar las evals	2026-09-15T17:20:39.6721862Z       "motivos": [],
evals	Ejecutar las evals	2026-09-15T17:20:39.6722090Z       "pasa": true
evals	Ejecutar las evals	2026-09-15T17:20:39.6722288Z     },
evals	Ejecutar las evals	2026-09-15T17:20:39.6722465Z     {
evals	Ejecutar las evals	2026-09-15T17:20:39.6722752Z       "sesion": "03-lrbrl-atribuciones-del-pleno",
evals	Ejecutar las evals	2026-09-15T17:20:39.6723160Z       "eval": "03-lrbrl-atribuciones-del-pleno.yaml",
evals	Ejecutar las evals	2026-09-15T17:20:39.6723486Z       "activa": true,
evals	Ejecutar las evals	2026-09-15T17:20:39.6723716Z       "activada": true,
evals	Ejecutar las evals	2026-09-15T17:20:39.6723976Z       "comandos_ejecutados": [
evals	Ejecutar las evals	2026-09-15T17:20:39.6724294Z         "bloque boe BOE-A-1985-5392 a22"
evals	Ejecutar las evals	2026-09-15T17:20:39.6724551Z       ],
evals	Ejecutar las evals	2026-09-15T17:20:39.6724774Z       "comandos_ausentes": [],
evals	Ejecutar las evals	2026-09-15T17:20:39.6725054Z       "citas_encontradas": [
evals	Ejecutar las evals	2026-09-15T17:20:39.6725338Z         "BOE-A-1985-5392 a22"
evals	Ejecutar las evals	2026-09-15T17:20:39.6725571Z       ],
evals	Ejecutar las evals	2026-09-15T17:20:39.6725782Z       "citas_ausentes": [],
evals	Ejecutar las evals	2026-09-15T17:20:39.6726044Z       "invocaciones": [
evals	Ejecutar las evals	2026-09-15T17:20:39.6726255Z         {
evals	Ejecutar las evals	2026-09-15T17:20:39.6726577Z           "orden": "boe indice BOE-A-1985-5392 --json",
evals	Ejecutar las evals	2026-09-15T17:20:39.6726893Z           "codigo": 0,
evals	Ejecutar las evals	2026-09-15T17:20:39.6727150Z           "conexiones": []
evals	Ejecutar las evals	2026-09-15T17:20:39.6727373Z         },
evals	Ejecutar las evals	2026-09-15T17:20:39.6727557Z         {
evals	Ejecutar las evals	2026-09-15T17:20:39.6727907Z           "orden": "boe articulo BOE-A-1985-5392 a22 --json",
evals	Ejecutar las evals	2026-09-15T17:20:39.6728243Z           "codigo": 0,
evals	Ejecutar las evals	2026-09-15T17:20:39.6728500Z           "conexiones": []
evals	Ejecutar las evals	2026-09-15T17:20:39.6728731Z         }
evals	Ejecutar las evals	2026-09-15T17:20:39.6728911Z       ],
evals	Ejecutar las evals	2026-09-15T17:20:39.6729204Z       "fuera_de_lo_grabado": [],
evals	Ejecutar las evals	2026-09-15T17:20:39.6729480Z       "otras_fallidas": [],
evals	Ejecutar las evals	2026-09-15T17:20:39.6729757Z       "llegadas_a_la_red": [],
evals	Ejecutar las evals	2026-09-15T17:20:39.6739231Z       "respuesta": "Según el **artículo 22 de la Ley 7/1985, de 2 de abril, Reguladora de las Bases del Régimen Local** [BOE-A-1985-5392, bloque a22], el Pleno municipal **—integrado por todos los Concejales y presidido por el Alcalde—** tiene las siguientes atribuciones principales:\n\n**Atribuciones del Pleno:**\n\n1. **Control y fiscalización** de los órganos de gobierno\n2. **Participación en organizaciones supramunicipales**, alteración del término municipal, creación o supresión de municipios y cambios en la capitalidad, nombre, bandera, enseña o escudo del municipio\n3. **Aprobación del planeamiento urbanístico** (planeamiento general y otros instrumentos de ordenación)\n4. **Aprobación del reglamento orgánico y ordenanzas** municipales\n5. **Determinación de recursos tributarios**, aprobación y modificación de presupuestos, disposición de gastos y aprobación de cuentas\n6. **Gestión de servicios** (aprobación de formas de gestión y expedientes de municipalización)\n7. **Aceptación de delegaciones** de competencias de otras Administraciones\n8. **Planteamiento de conflictos** de competencias con otras entidades\n9. **Aprobación de plantilla de personal**, relación de puestos y retribuciones de funcionarios\n10. **Ejercicio de acciones judiciales** y defensa de la corporación\n11. **Declaración de lesividad** de actos del Ayuntamiento\n12. **Alteración de la calificación jurídica** de bienes de dominio público\n13. **Operaciones de crédito** (cuando superen umbrales de porcentaje de presupuesto)\n14. **Aprobación de proyectos** de obras y servicios no previstos en presupuestos\n15. **Otras atribuciones** que exijan mayoría especial o que las leyes le confieran\n\nEl Pleno **puede delegar** sus atribuciones en el Alcalde y en la Junta de Gobierno Local, **excepto** en las enunciadas en los apartados a), b), c), d), e), f), g), h), i), l) y p), que son **indelegables**.",
evals	Ejecutar las evals	2026-09-15T17:20:39.6745419Z       "codigo_de_la_sesion": 0,
evals	Ejecutar las evals	2026-09-15T17:20:39.6745743Z       "fin_de_la_sesion": "result success",
evals	Ejecutar las evals	2026-09-15T17:20:39.6746054Z       "sesion_terminada": true,
evals	Ejecutar las evals	2026-09-15T17:20:39.6746327Z       "motivos": [],
evals	Ejecutar las evals	2026-09-15T17:20:39.6746554Z       "pasa": true
evals	Ejecutar las evals	2026-09-15T17:20:39.6746755Z     },
evals	Ejecutar las evals	2026-09-15T17:20:39.6746943Z     {
evals	Ejecutar las evals	2026-09-15T17:20:39.6747189Z       "sesion": "04-lgt-prescripcion",
evals	Ejecutar las evals	2026-09-15T17:20:39.6747524Z       "eval": "04-lgt-prescripcion.yaml",
evals	Ejecutar las evals	2026-09-15T17:20:39.6747817Z       "activa": true,
evals	Ejecutar las evals	2026-09-15T17:20:39.6748060Z       "activada": true,
evals	Ejecutar las evals	2026-09-15T17:20:39.6748320Z       "comandos_ejecutados": [
evals	Ejecutar las evals	2026-09-15T17:20:39.6748627Z         "bloque boe BOE-A-2003-23186 a66"
evals	Ejecutar las evals	2026-09-15T17:20:39.6748899Z       ],
evals	Ejecutar las evals	2026-09-15T17:20:39.6749175Z       "comandos_ausentes": [],
evals	Ejecutar las evals	2026-09-15T17:20:39.6749448Z       "citas_encontradas": [
evals	Ejecutar las evals	2026-09-15T17:20:39.6749722Z         "BOE-A-2003-23186 a66"
evals	Ejecutar las evals	2026-09-15T17:20:39.6749967Z       ],
evals	Ejecutar las evals	2026-09-15T17:20:39.6750183Z       "citas_ausentes": [],
evals	Ejecutar las evals	2026-09-15T17:20:39.6750439Z       "invocaciones": [
evals	Ejecutar las evals	2026-09-15T17:20:39.6750678Z         {
evals	Ejecutar las evals	2026-09-15T17:20:39.6751011Z           "orden": "boe indice BOE-A-2003-23186 --json",
evals	Ejecutar las evals	2026-09-15T17:20:39.6751336Z           "codigo": 0,
evals	Ejecutar las evals	2026-09-15T17:20:39.6751598Z           "conexiones": []
evals	Ejecutar las evals	2026-09-15T17:20:39.6751821Z         },
evals	Ejecutar las evals	2026-09-15T17:20:39.6752009Z         {
evals	Ejecutar las evals	2026-09-15T17:20:39.6752350Z           "orden": "boe articulo BOE-A-2003-23186 a66 --json",
evals	Ejecutar las evals	2026-09-15T17:20:39.6752688Z           "codigo": 0,
evals	Ejecutar las evals	2026-09-15T17:20:39.6752951Z           "conexiones": []
evals	Ejecutar las evals	2026-09-15T17:20:39.6753178Z         }
evals	Ejecutar las evals	2026-09-15T17:20:39.6753365Z       ],
evals	Ejecutar las evals	2026-09-15T17:20:39.6753593Z       "fuera_de_lo_grabado": [],
evals	Ejecutar las evals	2026-09-15T17:20:39.6753871Z       "otras_fallidas": [],
evals	Ejecutar las evals	2026-09-15T17:20:39.6754134Z       "llegadas_a_la_red": [],
evals	Ejecutar las evals	2026-09-15T17:20:39.6756309Z       "respuesta": "Según el **artículo 66 de la Ley 58/2003, de 17 de diciembre, General Tributaria**, el derecho de la Administración a liquidar una deuda tributaria prescribe a los **cuatro años**.\n\nEspecíficamente, el artículo establece que «prescriben a los cuatro años» el derecho de la Administración para determinar la deuda tributaria mediante liquidación [art. 66, Ley General Tributaria, BOE-A-2003-23186, bloque a66].",
evals	Ejecutar las evals	2026-09-15T17:20:39.6757824Z       "codigo_de_la_sesion": 0,
evals	Ejecutar las evals	2026-09-15T17:20:39.6758143Z       "fin_de_la_sesion": "result success",
evals	Ejecutar las evals	2026-09-15T17:20:39.6758463Z       "sesion_terminada": true,
evals	Ejecutar las evals	2026-09-15T17:20:39.6758725Z       "motivos": [],
evals	Ejecutar las evals	2026-09-15T17:20:39.6758966Z       "pasa": true
evals	Ejecutar las evals	2026-09-15T17:20:39.6759225Z     },
evals	Ejecutar las evals	2026-09-15T17:20:39.6759395Z     {
evals	Ejecutar las evals	2026-09-15T17:20:39.6759680Z       "sesion": "05-trlrhl-impuestos-municipales",
evals	Ejecutar las evals	2026-09-15T17:20:39.6760103Z       "eval": "05-trlrhl-impuestos-municipales.yaml",
evals	Ejecutar las evals	2026-09-15T17:20:39.6760427Z       "activa": true,
evals	Ejecutar las evals	2026-09-15T17:20:39.6760664Z       "activada": true,
evals	Ejecutar las evals	2026-09-15T17:20:39.6760929Z       "comandos_ejecutados": [
evals	Ejecutar las evals	2026-09-15T17:20:39.6761312Z         "bloque boe BOE-A-2004-4214 a59"
evals	Ejecutar las evals	2026-09-15T17:20:39.6761570Z       ],
evals	Ejecutar las evals	2026-09-15T17:20:39.6761806Z       "comandos_ausentes": [],
evals	Ejecutar las evals	2026-09-15T17:20:39.6762081Z       "citas_encontradas": [
evals	Ejecutar las evals	2026-09-15T17:20:39.6762357Z         "BOE-A-2004-4214 a59"
evals	Ejecutar las evals	2026-09-15T17:20:39.6762583Z       ],
evals	Ejecutar las evals	2026-09-15T17:20:39.6762804Z       "citas_ausentes": [],
evals	Ejecutar las evals	2026-09-15T17:20:39.6763059Z       "invocaciones": [
evals	Ejecutar las evals	2026-09-15T17:20:39.6763279Z         {
evals	Ejecutar las evals	2026-09-15T17:20:39.6763778Z           "orden": "boe buscar texto refundido ley reguladora haciendas locales --json",
evals	Ejecutar las evals	2026-09-15T17:20:39.6764194Z           "codigo": 5,
evals	Ejecutar las evals	2026-09-15T17:20:39.6764447Z           "conexiones": [
evals	Ejecutar las evals	2026-09-15T17:20:39.6764680Z             {
evals	Ejecutar las evals	2026-09-15T17:20:39.6764973Z               "destino": "127.0.0.1:9",
evals	Ejecutar las evals	2026-09-15T17:20:39.6765287Z               "clase": "local"
evals	Ejecutar las evals	2026-09-15T17:20:39.6765522Z             }
evals	Ejecutar las evals	2026-09-15T17:20:39.6765727Z           ]
evals	Ejecutar las evals	2026-09-15T17:20:39.6765917Z         },
evals	Ejecutar las evals	2026-09-15T17:20:39.6766108Z         {
evals	Ejecutar las evals	2026-09-15T17:20:39.6766428Z           "orden": "boe indice BOE-A-2004-4214 --json",
evals	Ejecutar las evals	2026-09-15T17:20:39.6766752Z           "codigo": 0,
evals	Ejecutar las evals	2026-09-15T17:20:39.6767021Z           "conexiones": []
evals	Ejecutar las evals	2026-09-15T17:20:39.6767250Z         },
evals	Ejecutar las evals	2026-09-15T17:20:39.6767443Z         {
evals	Ejecutar las evals	2026-09-15T17:20:39.6767785Z           "orden": "boe articulo BOE-A-2004-4214 a59 --json",
evals	Ejecutar las evals	2026-09-15T17:20:39.6768115Z           "codigo": 0,
evals	Ejecutar las evals	2026-09-15T17:20:39.6768379Z           "conexiones": []
evals	Ejecutar las evals	2026-09-15T17:20:39.6768610Z         }
evals	Ejecutar las evals	2026-09-15T17:20:39.6768787Z       ],
evals	Ejecutar las evals	2026-09-15T17:20:39.6769008Z       "fuera_de_lo_grabado": [
evals	Ejecutar las evals	2026-09-15T17:20:39.6769315Z         {
evals	Ejecutar las evals	2026-09-15T17:20:39.6769790Z           "orden": "boe buscar texto refundido ley reguladora haciendas locales --json",
evals	Ejecutar las evals	2026-09-15T17:20:39.6770201Z           "codigo": 5
evals	Ejecutar las evals	2026-09-15T17:20:39.6770417Z         }
evals	Ejecutar las evals	2026-09-15T17:20:39.6770600Z       ],
evals	Ejecutar las evals	2026-09-15T17:20:39.6770825Z       "otras_fallidas": [],
evals	Ejecutar las evals	2026-09-15T17:20:39.6771110Z       "llegadas_a_la_red": [],
evals	Ejecutar las evals	2026-09-15T17:20:39.6775806Z       "respuesta": "Según el artículo 59 del TRLRHL, los ayuntamientos pueden exigir los siguientes impuestos:\n\n**Impuestos obligatorios:**\nLos ayuntamientos exigirán, de acuerdo con la ley:\n\n1. **Impuesto sobre Bienes Inmuebles**\n2. **Impuesto sobre Actividades Económicas**\n3. **Impuesto sobre Vehículos de Tracción Mecánica**\n\n**Impuestos potestivos (voluntarios):**\nAsimismo, los ayuntamientos **podrán** establecer y exigir:\n\n4. **Impuesto sobre Construcciones, Instalaciones y Obras**\n5. **Impuesto sobre el Incremento de Valor de los Terrenos de Naturaleza Urbana**\n\n[Art. 59 del TRLRHL, BOE-A-2004-4214, bloque a59]\n\nLa diferencia clave es que los tres primeros son impuestos cuya exacción es obligatoria, mientras que los dos últimos quedan a la decisión de cada ayuntamiento, siendo necesario en todos los casos que se establezcan de acuerdo con la ley y las respectivas ordenanzas fiscales locales.",
evals	Ejecutar las evals	2026-09-15T17:20:39.6778754Z       "codigo_de_la_sesion": 0,
evals	Ejecutar las evals	2026-09-15T17:20:39.6779143Z       "fin_de_la_sesion": "result success",
evals	Ejecutar las evals	2026-09-15T17:20:39.6779456Z       "sesion_terminada": true,
evals	Ejecutar las evals	2026-09-15T17:20:39.6779887Z       "motivos": [],
evals	Ejecutar las evals	2026-09-15T17:20:39.6780128Z       "pasa": true
evals	Ejecutar las evals	2026-09-15T17:20:39.6780330Z     },
evals	Ejecutar las evals	2026-09-15T17:20:39.6780515Z     {
evals	Ejecutar las evals	2026-09-15T17:20:39.6780811Z       "sesion": "06-irpf-rendimientos-del-trabajo",
evals	Ejecutar las evals	2026-09-15T17:20:39.6781229Z       "eval": "06-irpf-rendimientos-del-trabajo.yaml",
evals	Ejecutar las evals	2026-09-15T17:20:39.6781554Z       "activa": true,
evals	Ejecutar las evals	2026-09-15T17:20:39.6781799Z       "activada": true,
evals	Ejecutar las evals	2026-09-15T17:20:39.6782056Z       "comandos_ejecutados": [
evals	Ejecutar las evals	2026-09-15T17:20:39.6782360Z         "bloque boe BOE-A-2006-20764 a17"
evals	Ejecutar las evals	2026-09-15T17:20:39.6782623Z       ],
evals	Ejecutar las evals	2026-09-15T17:20:39.6782849Z       "comandos_ausentes": [],
evals	Ejecutar las evals	2026-09-15T17:20:39.6783118Z       "citas_encontradas": [
evals	Ejecutar las evals	2026-09-15T17:20:39.6783396Z         "BOE-A-2006-20764 a17"
evals	Ejecutar las evals	2026-09-15T17:20:39.6783617Z       ],
evals	Ejecutar las evals	2026-09-15T17:20:39.6783831Z       "citas_ausentes": [],
evals	Ejecutar las evals	2026-09-15T17:20:39.6784146Z       "invocaciones": [
evals	Ejecutar las evals	2026-09-15T17:20:39.6784371Z         {
evals	Ejecutar las evals	2026-09-15T17:20:39.6784700Z           "orden": "boe indice BOE-A-2006-20764 --json",
evals	Ejecutar las evals	2026-09-15T17:20:39.6785022Z           "codigo": 0,
evals	Ejecutar las evals	2026-09-15T17:20:39.6785296Z           "conexiones": []
evals	Ejecutar las evals	2026-09-15T17:20:39.6785522Z         },
evals	Ejecutar las evals	2026-09-15T17:20:39.6785711Z         {
evals	Ejecutar las evals	2026-09-15T17:20:39.6786057Z           "orden": "boe articulo BOE-A-2006-20764 a17 --json",
evals	Ejecutar las evals	2026-09-15T17:20:39.6786399Z           "codigo": 0,
evals	Ejecutar las evals	2026-09-15T17:20:39.6786658Z           "conexiones": []
evals	Ejecutar las evals	2026-09-15T17:20:39.6786888Z         }
evals	Ejecutar las evals	2026-09-15T17:20:39.6787080Z       ],
evals	Ejecutar las evals	2026-09-15T17:20:39.6787316Z       "fuera_de_lo_grabado": [],
evals	Ejecutar las evals	2026-09-15T17:20:39.6787595Z       "otras_fallidas": [],
evals	Ejecutar las evals	2026-09-15T17:20:39.6787868Z       "llegadas_a_la_red": [],
evals	Ejecutar las evals	2026-09-15T17:20:39.6797117Z       "respuesta": "Según el artículo 17 de la Ley 35/2006 (LIRPF), **se consideran rendimientos íntegros del trabajo todas las contraprestaciones o utilidades, cualquiera que sea su denominación o naturaleza, dinerarias o en especie, que deriven, directa o indirectamente, del trabajo personal o de la relación laboral o estatutaria y no tengan el carácter de rendimientos de actividades económicas** [art. 17.1 de la Ley 35/2006, BOE-A-2006-20764, bloque a17].\n\n## Rendimientos que se incluyen específicamente:\n\n**1. En el apartado 1 se detallan:**\n- a) Los sueldos y salarios\n- b) Las prestaciones por desempleo\n- c) Las remuneraciones en concepto de gastos de representación\n- d) Las dietas y asignaciones para gastos de viaje (con excepciones en locomoción y manutención)\n- e) y f) Contribuciones y aportaciones a planes de pensiones y seguros de previsión social\n\n**2. En el apartado 2 se amplía a prestaciones que en todo caso tienen consideración de rendimientos del trabajo:**\n- **2.a)** Pensiones y haberes pasivos de Seguridad Social, mutualidades, planes de pensiones, seguros de previsión social, planes de dependencia\n- **2.b)** Cantidades abonadas a diputados, senadores, concejales y miembros de entidades locales\n- **2.c)** Rendimientos de cursos, conferencias y seminarios\n- **2.d)** Rendimientos de obras literarias, artísticas o científicas (con cesión de explotación)\n- **2.e)** Retribuciones de administradores y miembros de consejos\n- **2.f)** Pensiones compensatorias y anualidades por alimentos\n- **2.g)** Derechos especiales de contenido económico en sociedades\n- **2.h)** Becas\n- **2.i)** Retribuciones por actividades humanitarias de entidades sin ánimo de lucro\n- **2.j)** Retribuciones de relaciones laborales especiales\n- **2.k)** Aportaciones al patrimonio protegido de personas con discapacidad\n\n[art. 17 de la Ley 35/2006, BOE-A-2006-20764, bloque a17]",
evals	Ejecutar las evals	2026-09-15T17:20:39.6803268Z       "codigo_de_la_sesion": 0,
evals	Ejecutar las evals	2026-09-15T17:20:39.6803591Z       "fin_de_la_sesion": "result success",
evals	Ejecutar las evals	2026-09-15T17:20:39.6803904Z       "sesion_terminada": true,
evals	Ejecutar las evals	2026-09-15T17:20:39.6804160Z       "motivos": [],
evals	Ejecutar las evals	2026-09-15T17:20:39.6804395Z       "pasa": true
evals	Ejecutar las evals	2026-09-15T17:20:39.6804602Z     },
evals	Ejecutar las evals	2026-09-15T17:20:39.6804777Z     {
evals	Ejecutar las evals	2026-09-15T17:20:39.6805070Z       "sesion": "07-lrjsp-principio-de-legalidad",
evals	Ejecutar las evals	2026-09-15T17:20:39.6805550Z       "eval": "07-lrjsp-principio-de-legalidad.yaml",
evals	Ejecutar las evals	2026-09-15T17:20:39.6805871Z       "activa": true,
evals	Ejecutar las evals	2026-09-15T17:20:39.6806112Z       "activada": true,
evals	Ejecutar las evals	2026-09-15T17:20:39.6806369Z       "comandos_ejecutados": [
evals	Ejecutar las evals	2026-09-15T17:20:39.6806677Z         "bloque boe BOE-A-2015-10566 a25"
evals	Ejecutar las evals	2026-09-15T17:20:39.6806950Z       ],
evals	Ejecutar las evals	2026-09-15T17:20:39.6807185Z       "comandos_ausentes": [],
evals	Ejecutar las evals	2026-09-15T17:20:39.6807459Z       "citas_encontradas": [
evals	Ejecutar las evals	2026-09-15T17:20:39.6807733Z         "BOE-A-2015-10566 a25"
evals	Ejecutar las evals	2026-09-15T17:20:39.6807960Z       ],
evals	Ejecutar las evals	2026-09-15T17:20:39.6808176Z       "citas_ausentes": [],
evals	Ejecutar las evals	2026-09-15T17:20:39.6808430Z       "invocaciones": [
evals	Ejecutar las evals	2026-09-15T17:20:39.6808643Z         {
evals	Ejecutar las evals	2026-09-15T17:20:39.6808970Z           "orden": "boe indice BOE-A-2015-10566 --json",
evals	Ejecutar las evals	2026-09-15T17:20:39.6809347Z           "codigo": 0,
evals	Ejecutar las evals	2026-09-15T17:20:39.6809661Z           "conexiones": []
evals	Ejecutar las evals	2026-09-15T17:20:39.6809896Z         },
evals	Ejecutar las evals	2026-09-15T17:20:39.6810081Z         {
evals	Ejecutar las evals	2026-09-15T17:20:39.6810424Z           "orden": "boe articulo BOE-A-2015-10566 a25 --json",
evals	Ejecutar las evals	2026-09-15T17:20:39.6810774Z           "codigo": 0,
evals	Ejecutar las evals	2026-09-15T17:20:39.6811032Z           "conexiones": []
evals	Ejecutar las evals	2026-09-15T17:20:39.6811254Z         }
evals	Ejecutar las evals	2026-09-15T17:20:39.6811439Z       ],
evals	Ejecutar las evals	2026-09-15T17:20:39.6811679Z       "fuera_de_lo_grabado": [],
evals	Ejecutar las evals	2026-09-15T17:20:39.6811960Z       "otras_fallidas": [],
evals	Ejecutar las evals	2026-09-15T17:20:39.6812226Z       "llegadas_a_la_red": [],
evals	Ejecutar las evals	2026-09-15T17:20:39.6819834Z       "respuesta": "El artículo 25 de la Ley 40/2015, titulado **\"Principio de legalidad\"**, establece los requisitos fundamentales para el ejercicio de la potestad sancionadora por parte de las Administraciones Públicas:\n\n**Requisitos esenciales:**\n\n1. **Habilitación legal expresa**: La potestad sancionadora debe estar reconocida expresamente por una norma con rango de Ley [art. 25.1, Ley 40/2015, BOE-A-2015-10566, bloque a25]\n\n2. **Procedimiento regulado**: El ejercicio debe ajustarse al procedimiento previsto en la propia habilitación legal y conforme a lo establecido en la Ley 40/2015 y en la Ley 39/2015 (Procedimiento Administrativo Común). Para Entidades Locales, se aplica adicionalmente el Título XI de la Ley 7/1985 [art. 25.1, Ley 40/2015, BOE-A-2015-10566, bloque a25]\n\n3. **Atribución orgánica**: Solo los órganos administrativos que tengan expresamente atribuida la potestad, ya sea por disposición de rango legal o reglamentario, pueden ejercerla [art. 25.2, Ley 40/2015, BOE-A-2015-10566, bloque a25]\n\n**Extensiones y límites:**\n\n- Las disposiciones del Capítulo III (donde está ubicado este artículo) se aplican también al ejercicio de la potestad disciplinaria sobre el personal al servicio de las Administraciones Públicas [art. 25.3, Ley 40/2015, BOE-A-2015-10566, bloque a25]\n\n- No se aplican estas disposiciones cuando se trata de sanciones respecto a personas vinculadas por relaciones contractuales de contratación pública o patrimoniales [art. 25.4, Ley 40/2015, BOE-A-2015-10566, bloque a25]",
evals	Ejecutar las evals	2026-09-15T17:20:39.6824743Z       "codigo_de_la_sesion": 0,
evals	Ejecutar las evals	2026-09-15T17:20:39.6825058Z       "fin_de_la_sesion": "result success",
evals	Ejecutar las evals	2026-09-15T17:20:39.6825388Z       "sesion_terminada": true,
evals	Ejecutar las evals	2026-09-15T17:20:39.6825648Z       "motivos": [],
evals	Ejecutar las evals	2026-09-15T17:20:39.6825880Z       "pasa": true
evals	Ejecutar las evals	2026-09-15T17:20:39.6826088Z     },
evals	Ejecutar las evals	2026-09-15T17:20:39.6826263Z     {
evals	Ejecutar las evals	2026-09-15T17:20:39.6826542Z       "sesion": "08-ltaibg-plazo-de-resolucion",
evals	Ejecutar las evals	2026-09-15T17:20:39.6826937Z       "eval": "08-ltaibg-plazo-de-resolucion.yaml",
evals	Ejecutar las evals	2026-09-15T17:20:39.6827263Z       "activa": true,
evals	Ejecutar las evals	2026-09-15T17:20:39.6827497Z       "activada": true,
evals	Ejecutar las evals	2026-09-15T17:20:39.6827758Z       "comandos_ejecutados": [
evals	Ejecutar las evals	2026-09-15T17:20:39.6828076Z         "bloque boe BOE-A-2013-12887 a20"
evals	Ejecutar las evals	2026-09-15T17:20:39.6828331Z       ],
evals	Ejecutar las evals	2026-09-15T17:20:39.6828561Z       "comandos_ausentes": [],
evals	Ejecutar las evals	2026-09-15T17:20:39.6828841Z       "citas_encontradas": [
evals	Ejecutar las evals	2026-09-15T17:20:39.6829177Z         "BOE-A-2013-12887 a20"
evals	Ejecutar las evals	2026-09-15T17:20:39.6829401Z       ],
evals	Ejecutar las evals	2026-09-15T17:20:39.6829615Z       "citas_ausentes": [],
evals	Ejecutar las evals	2026-09-15T17:20:39.6829972Z       "invocaciones": [
evals	Ejecutar las evals	2026-09-15T17:20:39.6830190Z         {
evals	Ejecutar las evals	2026-09-15T17:20:39.6830521Z           "orden": "boe indice BOE-A-2013-12887 --json",
evals	Ejecutar las evals	2026-09-15T17:20:39.6830853Z           "codigo": 0,
evals	Ejecutar las evals	2026-09-15T17:20:39.6831113Z           "conexiones": []
evals	Ejecutar las evals	2026-09-15T17:20:39.6831341Z         },
evals	Ejecutar las evals	2026-09-15T17:20:39.6831530Z         {
evals	Ejecutar las evals	2026-09-15T17:20:39.6831882Z           "orden": "boe articulo BOE-A-2013-12887 a20 --json",
evals	Ejecutar las evals	2026-09-15T17:20:39.6832212Z           "codigo": 0,
evals	Ejecutar las evals	2026-09-15T17:20:39.6832481Z           "conexiones": []
evals	Ejecutar las evals	2026-09-15T17:20:39.6832721Z         }
evals	Ejecutar las evals	2026-09-15T17:20:39.6832904Z       ],
evals	Ejecutar las evals	2026-09-15T17:20:39.6833138Z       "fuera_de_lo_grabado": [],
evals	Ejecutar las evals	2026-09-15T17:20:39.6833422Z       "otras_fallidas": [],
evals	Ejecutar las evals	2026-09-15T17:20:39.6833687Z       "llegadas_a_la_red": [],
evals	Ejecutar las evals	2026-09-15T17:20:39.6838504Z       "respuesta": "Según el artículo 20 de la Ley 19/2013, el plazo para resolver una solicitud de acceso a información pública es:\n\n**Plazo base: un mes**\n\nLa resolución en la que se conceda o deniegue el acceso debe notificarse al solicitante en un plazo máximo de **un mes desde la recepción de la solicitud** por el órgano competente para resolver [art. 20.1 de la LTAIBG, BOE-A-2013-12887, bloque a20].\n\n**Ampliación: hasta dos meses**\n\nEste plazo puede ampliarse por **otro mes adicional** cuando el volumen o la complejidad de la información que se solicita así lo hagan necesario, siempre que se notifique previamente esta ampliación al solicitante [art. 20.1 de la LTAIBG, BOE-A-2013-12887, bloque a20].\n\n**Consideración importante:**\n\nSi transcurre el plazo máximo sin que se dicte y notifique una resolución expresa, se entiende que la solicitud ha sido desestimada (silencio administrativo negativo) [art. 20.4 de la LTAIBG, BOE-A-2013-12887, bloque a20].",
evals	Ejecutar las evals	2026-09-15T17:20:39.6841694Z       "codigo_de_la_sesion": 0,
evals	Ejecutar las evals	2026-09-15T17:20:39.6842019Z       "fin_de_la_sesion": "result success",
evals	Ejecutar las evals	2026-09-15T17:20:39.6842344Z       "sesion_terminada": true,
evals	Ejecutar las evals	2026-09-15T17:20:39.6842613Z       "motivos": [],
evals	Ejecutar las evals	2026-09-15T17:20:39.6842854Z       "pasa": true
evals	Ejecutar las evals	2026-09-15T17:20:39.6843062Z     },
evals	Ejecutar las evals	2026-09-15T17:20:39.6843239Z     {
evals	Ejecutar las evals	2026-09-15T17:20:39.6843514Z       "sesion": "09-constitucion-articulo-140",
evals	Ejecutar las evals	2026-09-15T17:20:39.6843904Z       "eval": "09-constitucion-articulo-140.yaml",
evals	Ejecutar las evals	2026-09-15T17:20:39.6844229Z       "activa": true,
evals	Ejecutar las evals	2026-09-15T17:20:39.6844469Z       "activada": true,
evals	Ejecutar las evals	2026-09-15T17:20:39.6844728Z       "comandos_ejecutados": [
evals	Ejecutar las evals	2026-09-15T17:20:39.6845052Z         "bloque boe BOE-A-1978-31229 a140"
evals	Ejecutar las evals	2026-09-15T17:20:39.6845315Z       ],
evals	Ejecutar las evals	2026-09-15T17:20:39.6845538Z       "comandos_ausentes": [],
evals	Ejecutar las evals	2026-09-15T17:20:39.6845818Z       "citas_encontradas": [
evals	Ejecutar las evals	2026-09-15T17:20:39.6846094Z         "BOE-A-1978-31229 a140"
evals	Ejecutar las evals	2026-09-15T17:20:39.6846320Z       ],
evals	Ejecutar las evals	2026-09-15T17:20:39.6846532Z       "citas_ausentes": [],
evals	Ejecutar las evals	2026-09-15T17:20:39.6846789Z       "invocaciones": [
evals	Ejecutar las evals	2026-09-15T17:20:39.6847010Z         {
evals	Ejecutar las evals	2026-09-15T17:20:39.6847342Z           "orden": "boe indice BOE-A-1978-31229 --json",
evals	Ejecutar las evals	2026-09-15T17:20:39.6847669Z           "codigo": 0,
evals	Ejecutar las evals	2026-09-15T17:20:39.6847924Z           "conexiones": []
evals	Ejecutar las evals	2026-09-15T17:20:39.6848148Z         },
evals	Ejecutar las evals	2026-09-15T17:20:39.6848342Z         {
evals	Ejecutar las evals	2026-09-15T17:20:39.6848687Z           "orden": "boe articulo BOE-A-1978-31229 a140 --json",
evals	Ejecutar las evals	2026-09-15T17:20:39.6849020Z           "codigo": 0,
evals	Ejecutar las evals	2026-09-15T17:20:39.6849332Z           "conexiones": []
evals	Ejecutar las evals	2026-09-15T17:20:39.6849567Z         }
evals	Ejecutar las evals	2026-09-15T17:20:39.6849750Z       ],
evals	Ejecutar las evals	2026-09-15T17:20:39.6849981Z       "fuera_de_lo_grabado": [],
evals	Ejecutar las evals	2026-09-15T17:20:39.6850263Z       "otras_fallidas": [],
evals	Ejecutar las evals	2026-09-15T17:20:39.6850525Z       "llegadas_a_la_red": [],
evals	Ejecutar las evals	2026-09-15T17:20:39.6855776Z       "respuesta": "El artículo 140 de la Constitución Española establece:\n\n**Contenido:**\n\nLa Constitución garantiza la autonomía de los municipios, que gozan de personalidad jurídica plena. El gobierno y administración de los municipios corresponde a sus respectivos Ayuntamientos, integrados por los Alcaldes y los Concejales. \n\nLos Concejales son elegidos por los vecinos del municipio mediante **sufragio universal, igual, libre, directo y secreto**, en la forma establecida por la ley. Por su parte, los Alcaldes pueden ser elegidos por los Concejales o por los vecinos, según lo que determine la ley.\n\nLa ley también regulará las condiciones en las que proceda el **régimen del concejo abierto**.\n\n[art. 140 de la Constitución Española, BOE-A-1978-31229, bloque a140]\n\nEste artículo es fundamental en el sistema de autonomía local español, ya que reconoce el derecho de los municipios a la autonomía, establece su estructura administrativa básica y garantiza la participación democrática de los vecinos en la elección de sus representantes.",
evals	Ejecutar las evals	2026-09-15T17:20:39.6859617Z       "codigo_de_la_sesion": 0,
evals	Ejecutar las evals	2026-09-15T17:20:39.6859940Z       "fin_de_la_sesion": "result success",
evals	Ejecutar las evals	2026-09-15T17:20:39.6860260Z       "sesion_terminada": true,
evals	Ejecutar las evals	2026-09-15T17:20:39.6860606Z       "motivos": [],
evals	Ejecutar las evals	2026-09-15T17:20:39.6860846Z       "pasa": true
evals	Ejecutar las evals	2026-09-15T17:20:39.6861048Z     },
evals	Ejecutar las evals	2026-09-15T17:20:39.6861235Z     {
evals	Ejecutar las evals	2026-09-15T17:20:39.6861476Z       "sesion": "10-et-vacaciones",
evals	Ejecutar las evals	2026-09-15T17:20:39.6861807Z       "eval": "10-et-vacaciones.yaml",
evals	Ejecutar las evals	2026-09-15T17:20:39.6862089Z       "activa": true,
evals	Ejecutar las evals	2026-09-15T17:20:39.6862337Z       "activada": true,
evals	Ejecutar las evals	2026-09-15T17:20:39.6862612Z       "comandos_ejecutados": [
evals	Ejecutar las evals	2026-09-15T17:20:39.6862924Z         "bloque boe BOE-A-2015-11430 a38"
evals	Ejecutar las evals	2026-09-15T17:20:39.6863185Z       ],
evals	Ejecutar las evals	2026-09-15T17:20:39.6863414Z       "comandos_ausentes": [],
evals	Ejecutar las evals	2026-09-15T17:20:39.6863694Z       "citas_encontradas": [
evals	Ejecutar las evals	2026-09-15T17:20:39.6863975Z         "BOE-A-2015-11430 a38"
evals	Ejecutar las evals	2026-09-15T17:20:39.6864200Z       ],
evals	Ejecutar las evals	2026-09-15T17:20:39.6864431Z       "citas_ausentes": [],
evals	Ejecutar las evals	2026-09-15T17:20:39.6864689Z       "invocaciones": [
evals	Ejecutar las evals	2026-09-15T17:20:39.6864911Z         {
evals	Ejecutar las evals	2026-09-15T17:20:39.6865232Z           "orden": "boe indice BOE-A-2015-11430 --json",
evals	Ejecutar las evals	2026-09-15T17:20:39.6865568Z           "codigo": 0,
evals	Ejecutar las evals	2026-09-15T17:20:39.6865827Z           "conexiones": []
evals	Ejecutar las evals	2026-09-15T17:20:39.6866054Z         },
evals	Ejecutar las evals	2026-09-15T17:20:39.6866251Z         {
evals	Ejecutar las evals	2026-09-15T17:20:39.6866588Z           "orden": "boe articulo BOE-A-2015-11430 a38 --json",
evals	Ejecutar las evals	2026-09-15T17:20:39.6866925Z           "codigo": 0,
evals	Ejecutar las evals	2026-09-15T17:20:39.6867188Z           "conexiones": []
evals	Ejecutar las evals	2026-09-15T17:20:39.6867413Z         }
evals	Ejecutar las evals	2026-09-15T17:20:39.6867595Z       ],
evals	Ejecutar las evals	2026-09-15T17:20:39.6867825Z       "fuera_de_lo_grabado": [],
evals	Ejecutar las evals	2026-09-15T17:20:39.6868111Z       "otras_fallidas": [],
evals	Ejecutar las evals	2026-09-15T17:20:39.6868375Z       "llegadas_a_la_red": [],
evals	Ejecutar las evals	2026-09-15T17:20:39.6872186Z       "respuesta": "El artículo 38 del Estatuto de los Trabajadores reconoce **como mínimo treinta días naturales de vacaciones anuales retribuidas**.\n\nSegún el artículo: «El periodo de vacaciones anuales retribuidas, no sustituible por compensación económica, será el pactado en convenio colectivo o contrato individual. En ningún caso la duración será inferior a treinta días naturales.» [Real Decreto Legislativo 2/2015 de 23 de octubre (Estatuto de los Trabajadores), art. 38.1, BOE-A-2015-11430, bloque a38]\n\nEsta es la duración **mínima legal**. Las partes (empresa y trabajador) pueden acordar un mayor número de días a través del convenio colectivo o el contrato individual, pero no pueden pactar menos de treinta días naturales.",
evals	Ejecutar las evals	2026-09-15T17:20:39.6874629Z       "codigo_de_la_sesion": 0,
evals	Ejecutar las evals	2026-09-15T17:20:39.6874939Z       "fin_de_la_sesion": "result success",
evals	Ejecutar las evals	2026-09-15T17:20:39.6875271Z       "sesion_terminada": true,
evals	Ejecutar las evals	2026-09-15T17:20:39.6875531Z       "motivos": [],
evals	Ejecutar las evals	2026-09-15T17:20:39.6875767Z       "pasa": true
evals	Ejecutar las evals	2026-09-15T17:20:39.6875976Z     },
evals	Ejecutar las evals	2026-09-15T17:20:39.6876149Z     {
evals	Ejecutar las evals	2026-09-15T17:20:39.6876417Z       "sesion": "11-no-activa-programacion",
evals	Ejecutar las evals	2026-09-15T17:20:39.6876799Z       "eval": "11-no-activa-programacion.yaml",
evals	Ejecutar las evals	2026-09-15T17:20:39.6877110Z       "activa": false,
evals	Ejecutar las evals	2026-09-15T17:20:39.6877355Z       "activada": false,
evals	Ejecutar las evals	2026-09-15T17:20:39.6877636Z       "comandos_ejecutados": [],
evals	Ejecutar las evals	2026-09-15T17:20:39.6877933Z       "comandos_ausentes": [],
evals	Ejecutar las evals	2026-09-15T17:20:39.6878312Z       "citas_encontradas": [],
evals	Ejecutar las evals	2026-09-15T17:20:39.6889537Z       "citas_ausentes": [],
evals	Ejecutar las evals	2026-09-15T17:20:39.6889915Z       "invocaciones": [],
evals	Ejecutar las evals	2026-09-15T17:20:39.6890249Z       "fuera_de_lo_grabado": [],
evals	Ejecutar las evals	2026-09-15T17:20:39.6890568Z       "otras_fallidas": [],
evals	Ejecutar las evals	2026-09-15T17:20:39.6890875Z       "llegadas_a_la_red": [],
evals	Ejecutar las evals	2026-09-15T17:20:39.6898885Z       "respuesta": "Para invertir una lista enlazada en Go, puedes hacerlo de forma iterativa o recursiva. Aquí te muestro la solución más común (iterativa):\n\n```go\ntype Node struct {\n    Val  int\n    Next *Node\n}\n\nfunc ReverseList(head *Node) *Node {\n    var prev *Node\n    current := head\n    \n    for current != nil {\n        // Guardar el siguiente nodo\n        next := current.Next\n        \n        // Invertir el enlace\n        current.Next = prev\n        \n        // Avanzar\n        prev = current\n        current = next\n    }\n    \n    return prev\n}\n```\n\n**Explicación:**\n1. `prev` comienza en `nil` (será la cola del final)\n2. Para cada nodo, guardamos el siguiente antes de modificar el enlace\n3. Hacemos que el nodo actual apunte hacia atrás (`current.Next = prev`)\n4. Avanzamos con `prev = current` y `current = next`\n5. Retornamos `prev` (nuevo cabecera)\n\n**Ejemplo de uso:**\n```go\n// Crear lista: 1 -> 2 -> 3\nhead := &Node{Val: 1, Next: &Node{Val: 2, Next: &Node{Val: 3}}}\n\n// Invertir: 3 -> 2 -> 1\nhead = ReverseList(head)\n```\n\n**Complejidad:**\n- **Tiempo:** O(n) - recorremos una sola vez\n- **Espacio:** O(1) - solo usamos un par de punteros\n\nSi prefieres **solución recursiva**:\n```go\nfunc ReverseListRecursive(head *Node) *Node {\n    if head == nil || head.Next == nil {\n        return head\n    }\n    \n    newHead := ReverseListRecursive(head.Next)\n    head.Next.Next = head\n    head.Next = nil\n    \n    return newHead\n}\n```\n\nLa iterativa es más eficiente en memoria (no usa pila de recursión).",
evals	Ejecutar las evals	2026-09-15T17:20:39.6903706Z       "codigo_de_la_sesion": 0,
evals	Ejecutar las evals	2026-09-15T17:20:39.6904024Z       "fin_de_la_sesion": "result success",
evals	Ejecutar las evals	2026-09-15T17:20:39.6904368Z       "sesion_terminada": true,
evals	Ejecutar las evals	2026-09-15T17:20:39.6904632Z       "motivos": [],
evals	Ejecutar las evals	2026-09-15T17:20:39.6904865Z       "pasa": true
evals	Ejecutar las evals	2026-09-15T17:20:39.6905078Z     },
evals	Ejecutar las evals	2026-09-15T17:20:39.6905255Z     {
evals	Ejecutar las evals	2026-09-15T17:20:39.6905559Z       "sesion": "12-no-activa-acuerdo-entre-amigos",
evals	Ejecutar las evals	2026-09-15T17:20:39.6905997Z       "eval": "12-no-activa-acuerdo-entre-amigos.yaml",
evals	Ejecutar las evals	2026-09-15T17:20:39.6906329Z       "activa": false,
evals	Ejecutar las evals	2026-09-15T17:20:39.6906571Z       "activada": false,
evals	Ejecutar las evals	2026-09-15T17:20:39.6906845Z       "comandos_ejecutados": [],
evals	Ejecutar las evals	2026-09-15T17:20:39.6907136Z       "comandos_ausentes": [],
evals	Ejecutar las evals	2026-09-15T17:20:39.6907410Z       "citas_encontradas": [],
evals	Ejecutar las evals	2026-09-15T17:20:39.6907678Z       "citas_ausentes": [],
evals	Ejecutar las evals	2026-09-15T17:20:39.6907949Z       "invocaciones": [],
evals	Ejecutar las evals	2026-09-15T17:20:39.6908219Z       "fuera_de_lo_grabado": [],
evals	Ejecutar las evals	2026-09-15T17:20:39.6908502Z       "otras_fallidas": [],
evals	Ejecutar las evals	2026-09-15T17:20:39.6908772Z       "llegadas_a_la_red": [],
evals	Ejecutar las evals	2026-09-15T17:20:39.6912487Z       "respuesta": "Aquí tienes algunas opciones más cercanas:\n\n**Más informal:**\n- \"Nos alternamos el coche cada fin de semana\"\n- \"Cada uno disfruta del coche un fin de semana sí, otro no\"\n- \"Nos pasamos el coche: un fin de semana es para ti, el siguiente para mí\"\n\n**Más coloquial:**\n- \"Uno cada fin de semana, alternando\"\n- \"Nos turnamos el coche entre fin de semana\"\n- \"El coche va rotando: un fin de semana para uno, el siguiente para el otro\"\n\n**Si quieres algo muy natural:**\n- \"Nos repartimos el coche en fines de semana alternos\"\n\nLa que suena más amistosa y directa sería: **\"Nos alternamos el coche cada fin de semana\"** o **\"Uno cada fin de semana, alternando\"**. Mantienen la claridad legal pero con un tono más relajado.",
evals	Ejecutar las evals	2026-09-15T17:20:39.6914803Z       "codigo_de_la_sesion": 0,
evals	Ejecutar las evals	2026-09-15T17:20:39.6915126Z       "fin_de_la_sesion": "result success",
evals	Ejecutar las evals	2026-09-15T17:20:39.6915444Z       "sesion_terminada": true,
evals	Ejecutar las evals	2026-09-15T17:20:39.6915708Z       "motivos": [],
evals	Ejecutar las evals	2026-09-15T17:20:39.6916022Z       "pasa": true
evals	Ejecutar las evals	2026-09-15T17:20:39.6916213Z     }
evals	Ejecutar las evals	2026-09-15T17:20:39.6916396Z   ]
evals	Ejecutar las evals	2026-09-15T17:20:39.6916597Z }
evals	Ejecutar las evals	2026-09-15T17:20:39.6916809Z --- fin de informe.json ---
código de la quinta orden: 0
`````

## Anexo B · Sexta orden de §12.2: retirada de Python entre marcas, tal cual

`````text
evals	Retirar Python del runner	2026-09-15T17:14:19.1858795Z --- inicio de la retirada de Python ---
evals	Retirar Python del runner	2026-09-15T17:14:19.2011606Z búsqueda: find / ( -path /proc -o -path /sys ) -prune -o ( ( -type f -perm /111 ( -iname python* -o -iname pypy* ) ) -o ( -type l ( -iname python* -o -iname pypy* ) ) -o ( ( -type f -o -type l ) ( -iname libpython* -o -iname libpypy* ) ) ) -print
evals	Retirar Python del runner	2026-09-15T17:15:44.7670548Z retirado: /opt/pipx/shared/lib/python3.12/site-packages/pip/_vendor/pygments/lexers/__pycache__/python.cpython-312.pyc
evals	Retirar Python del runner	2026-09-15T17:15:44.7816102Z retirado: /opt/pipx/shared/lib/python3.12/site-packages/pip/_vendor/pygments/lexers/python.py
evals	Retirar Python del runner	2026-09-15T17:15:44.8078995Z retirado: /opt/pipx/shared
evals	Retirar Python del runner	2026-09-15T17:15:44.8667461Z retirado: /opt/pipx/venvs/yamllint
evals	Retirar Python del runner	2026-09-15T17:15:44.9043705Z retirado: /opt/pipx/venvs/ansible-core/lib/python3.12/site-packages/ansible_test/_util/target/injector/__pycache__/python.cpython-312.pyc
evals	Retirar Python del runner	2026-09-15T17:15:44.9175780Z retirado: /opt/pipx/venvs/ansible-core/lib/python3.12/site-packages/ansible_test/_util/target/injector/python.py
evals	Retirar Python del runner	2026-09-15T17:15:44.9312690Z retirado: /opt/pipx/venvs/ansible-core/lib/python3.12/site-packages/ansible_test/_internal/__pycache__/python_requirements.cpython-312.pyc
evals	Retirar Python del runner	2026-09-15T17:15:44.9448515Z retirado: /opt/pipx/venvs/ansible-core/lib/python3.12/site-packages/ansible_test/_internal/classification/__pycache__/python.cpython-312.pyc
evals	Retirar Python del runner	2026-09-15T17:15:44.9584930Z retirado: /opt/pipx/venvs/ansible-core/lib/python3.12/site-packages/ansible_test/_internal/classification/python.py
evals	Retirar Python del runner	2026-09-15T17:15:44.9721581Z retirado: /opt/pipx/venvs/ansible-core/lib/python3.12/site-packages/ansible_test/_internal/python_requirements.py
evals	Retirar Python del runner	2026-09-15T17:15:44.9856111Z retirado: /opt/pipx/venvs/ansible-core/lib/python3.12/site-packages/ansible_collections/community/okd/molecule/default/roles/openshift_adm_groups/tasks/python-ldap-not-installed.yml
evals	Retirar Python del runner	2026-09-15T17:15:44.9995326Z retirado: /opt/pipx/venvs/ansible-core/lib/python3.12/site-packages/ansible_collections/community/general/plugins/modules/__pycache__/python_requirements_info.cpython-312.pyc
evals	Retirar Python del runner	2026-09-15T17:15:45.0135248Z retirado: /opt/pipx/venvs/ansible-core/lib/python3.12/site-packages/ansible_collections/community/general/plugins/modules/python_requirements_info.py
evals	Retirar Python del runner	2026-09-15T17:15:45.0271090Z retirado: /opt/pipx/venvs/ansible-core/lib/python3.12/site-packages/ansible/_internal/ansible_collections/ansible/_protomatter/plugins/filter/__pycache__/python_literal_eval.cpython-312.pyc
evals	Retirar Python del runner	2026-09-15T17:15:45.0403837Z retirado: /opt/pipx/venvs/ansible-core/lib/python3.12/site-packages/ansible/_internal/ansible_collections/ansible/_protomatter/plugins/filter/python_literal_eval.yml
evals	Retirar Python del runner	2026-09-15T17:15:45.0540909Z retirado: /opt/pipx/venvs/ansible-core/lib/python3.12/site-packages/ansible/_internal/ansible_collections/ansible/_protomatter/plugins/filter/python_literal_eval.py
evals	Retirar Python del runner	2026-09-15T17:15:45.0680337Z retirado: /opt/pipx/venvs/ansible-core/lib/python3.12/site-packages/ansible/module_utils/facts/system/__pycache__/python.cpython-312.pyc
evals	Retirar Python del runner	2026-09-15T17:15:45.0814457Z retirado: /opt/pipx/venvs/ansible-core/lib/python3.12/site-packages/ansible/module_utils/facts/system/python.py
evals	Retirar Python del runner	2026-09-15T17:15:45.1074554Z retirado: /opt/pipx/venvs/ansible-core
evals	Retirar Python del runner	2026-09-15T17:15:45.6311205Z retirado: /opt/hostedtoolcache/PyPy/3.9.19/x64/lib/pypy3.9/site-packages/pip/_vendor/pygments/lexers/__pycache__/python.pypy39.pyc
evals	Retirar Python del runner	2026-09-15T17:15:45.6452076Z retirado: /opt/hostedtoolcache/PyPy/3.9.19/x64/lib/pypy3.9/site-packages/pip/_vendor/pygments/lexers/python.py
evals	Retirar Python del runner	2026-09-15T17:15:45.6586248Z retirado: /opt/hostedtoolcache/PyPy/3.9.19/x64/lib/pypy3.9/_cffi_ssl/_cffi_src/openssl/pypy_win32_extra.py
evals	Retirar Python del runner	2026-09-15T17:15:45.6722440Z retirado: /opt/hostedtoolcache/PyPy/3.9.19/x64/lib/pypy3.9/hpy/devel/include/hpy/forbid_python_h/Python.h
evals	Retirar Python del runner	2026-09-15T17:15:45.6856799Z retirado: /opt/hostedtoolcache/PyPy/3.9.19/x64/lib/pypy3.9/pyrepl/python_reader.py
evals	Retirar Python del runner	2026-09-15T17:15:45.6989739Z retirado: /opt/hostedtoolcache/PyPy/3.9.19/x64/lib/pypy3.9/idlelib/Icons/python.gif
evals	Retirar Python del runner	2026-09-15T17:15:45.7129757Z retirado: /opt/hostedtoolcache/PyPy/3.9.19/x64/lib/pypy3.9/test/imghdrdata/python.jpg
evals	Retirar Python del runner	2026-09-15T17:15:45.7264612Z retirado: /opt/hostedtoolcache/PyPy/3.9.19/x64/lib/pypy3.9/test/imghdrdata/python.sgi
evals	Retirar Python del runner	2026-09-15T17:15:45.7398820Z retirado: /opt/hostedtoolcache/PyPy/3.9.19/x64/lib/pypy3.9/test/imghdrdata/python.tiff
evals	Retirar Python del runner	2026-09-15T17:15:45.7533931Z retirado: /opt/hostedtoolcache/PyPy/3.9.19/x64/lib/pypy3.9/test/imghdrdata/python.bmp
evals	Retirar Python del runner	2026-09-15T17:15:45.7666612Z retirado: /opt/hostedtoolcache/PyPy/3.9.19/x64/lib/pypy3.9/test/imghdrdata/python.gif
evals	Retirar Python del runner	2026-09-15T17:15:45.7802966Z retirado: /opt/hostedtoolcache/PyPy/3.9.19/x64/lib/pypy3.9/test/imghdrdata/python.webp
evals	Retirar Python del runner	2026-09-15T17:15:45.7936210Z retirado: /opt/hostedtoolcache/PyPy/3.9.19/x64/lib/pypy3.9/test/imghdrdata/python.png
evals	Retirar Python del runner	2026-09-15T17:15:45.8069749Z retirado: /opt/hostedtoolcache/PyPy/3.9.19/x64/lib/pypy3.9/test/imghdrdata/python.ppm
evals	Retirar Python del runner	2026-09-15T17:15:45.8204572Z retirado: /opt/hostedtoolcache/PyPy/3.9.19/x64/lib/pypy3.9/test/imghdrdata/python.pgm
evals	Retirar Python del runner	2026-09-15T17:15:45.8335911Z retirado: /opt/hostedtoolcache/PyPy/3.9.19/x64/lib/pypy3.9/test/imghdrdata/python.exr
evals	Retirar Python del runner	2026-09-15T17:15:45.8468827Z retirado: /opt/hostedtoolcache/PyPy/3.9.19/x64/lib/pypy3.9/test/imghdrdata/python.xbm
evals	Retirar Python del runner	2026-09-15T17:15:45.8600918Z retirado: /opt/hostedtoolcache/PyPy/3.9.19/x64/lib/pypy3.9/test/imghdrdata/python.pbm
evals	Retirar Python del runner	2026-09-15T17:15:45.8731062Z retirado: /opt/hostedtoolcache/PyPy/3.9.19/x64/lib/pypy3.9/test/imghdrdata/python.ras
evals	Retirar Python del runner	2026-09-15T17:15:45.8863464Z retirado: /opt/hostedtoolcache/PyPy/3.9.19/x64/lib/pypy3.9/test/pythoninfo.py
evals	Retirar Python del runner	2026-09-15T17:15:45.9003570Z retirado: /opt/hostedtoolcache/PyPy/3.9.19/x64/lib/PYPY_PORTABLE_DEPS.txt
evals	Retirar Python del runner	2026-09-15T17:15:45.9136220Z retirado: /opt/hostedtoolcache/PyPy/3.9.19/x64/include/pypy3.9/Python.h
evals	Retirar Python del runner	2026-09-15T17:15:45.9268154Z retirado: /opt/hostedtoolcache/PyPy/3.9.19/x64/include/pypy3.9/pythonrun.h
evals	Retirar Python del runner	2026-09-15T17:15:45.9404457Z retirado: /opt/hostedtoolcache/PyPy/3.9.19/x64/include/pypy3.9/pypy_macros.h
evals	Retirar Python del runner	2026-09-15T17:15:45.9535730Z retirado: /opt/hostedtoolcache/PyPy/3.9.19/x64/include/pypy3.9/pypy_marshal_decl.h
evals	Retirar Python del runner	2026-09-15T17:15:45.9668169Z retirado: /opt/hostedtoolcache/PyPy/3.9.19/x64/include/pypy3.9/pypy_decl.h
evals	Retirar Python del runner	2026-09-15T17:15:45.9801089Z retirado: /opt/hostedtoolcache/PyPy/3.9.19/x64/include/pypy3.9/pypy_structmember_decl.h
evals	Retirar Python del runner	2026-09-15T17:15:45.9935100Z retirado: /opt/hostedtoolcache/PyPy/3.9.19/x64/PYPY_VERSION
evals	Retirar Python del runner	2026-09-15T17:15:46.0193215Z retirado: /opt/hostedtoolcache/PyPy/3.9.19/x64
evals	Retirar Python del runner	2026-09-15T17:15:46.1760320Z retirado: /opt/hostedtoolcache/PyPy/3.11.15/x64/lib/PYPY_PORTABLE_DEPS.txt
evals	Retirar Python del runner	2026-09-15T17:15:46.1890729Z retirado: /opt/hostedtoolcache/PyPy/3.11.15/x64/lib/pypy3.11/site-packages/pip/_vendor/pygments/lexers/__pycache__/python.pypy311.pyc
evals	Retirar Python del runner	2026-09-15T17:15:46.2027123Z retirado: /opt/hostedtoolcache/PyPy/3.11.15/x64/lib/pypy3.11/site-packages/pip/_vendor/pygments/lexers/python.py
evals	Retirar Python del runner	2026-09-15T17:15:46.2162125Z retirado: /opt/hostedtoolcache/PyPy/3.11.15/x64/lib/pypy3.11/_cffi_ssl/_cffi_src/openssl/pypy_win32_extra.py
evals	Retirar Python del runner	2026-09-15T17:15:46.2296422Z retirado: /opt/hostedtoolcache/PyPy/3.11.15/x64/lib/pypy3.11/hpy/devel/include/hpy/forbid_python_h/Python.h
evals	Retirar Python del runner	2026-09-15T17:15:46.2434312Z retirado: /opt/hostedtoolcache/PyPy/3.11.15/x64/lib/pypy3.11/idlelib/Icons/python.gif
evals	Retirar Python del runner	2026-09-15T17:15:46.2571574Z retirado: /opt/hostedtoolcache/PyPy/3.11.15/x64/lib/pypy3.11/test/imghdrdata/python.jpg
evals	Retirar Python del runner	2026-09-15T17:15:46.2705445Z retirado: /opt/hostedtoolcache/PyPy/3.11.15/x64/lib/pypy3.11/test/imghdrdata/python.sgi
evals	Retirar Python del runner	2026-09-15T17:15:46.2841204Z retirado: /opt/hostedtoolcache/PyPy/3.11.15/x64/lib/pypy3.11/test/imghdrdata/python.tiff
evals	Retirar Python del runner	2026-09-15T17:15:46.2976852Z retirado: /opt/hostedtoolcache/PyPy/3.11.15/x64/lib/pypy3.11/test/imghdrdata/python.bmp
evals	Retirar Python del runner	2026-09-15T17:15:46.3115888Z retirado: /opt/hostedtoolcache/PyPy/3.11.15/x64/lib/pypy3.11/test/imghdrdata/python.gif
evals	Retirar Python del runner	2026-09-15T17:15:46.3252818Z retirado: /opt/hostedtoolcache/PyPy/3.11.15/x64/lib/pypy3.11/test/imghdrdata/python.webp
evals	Retirar Python del runner	2026-09-15T17:15:46.3387587Z retirado: /opt/hostedtoolcache/PyPy/3.11.15/x64/lib/pypy3.11/test/imghdrdata/python.png
evals	Retirar Python del runner	2026-09-15T17:15:46.3517313Z retirado: /opt/hostedtoolcache/PyPy/3.11.15/x64/lib/pypy3.11/test/imghdrdata/python.ppm
evals	Retirar Python del runner	2026-09-15T17:15:46.3651012Z retirado: /opt/hostedtoolcache/PyPy/3.11.15/x64/lib/pypy3.11/test/imghdrdata/python.pgm
evals	Retirar Python del runner	2026-09-15T17:15:46.3783384Z retirado: /opt/hostedtoolcache/PyPy/3.11.15/x64/lib/pypy3.11/test/imghdrdata/python.exr
evals	Retirar Python del runner	2026-09-15T17:15:46.3916447Z retirado: /opt/hostedtoolcache/PyPy/3.11.15/x64/lib/pypy3.11/test/imghdrdata/python.xbm
evals	Retirar Python del runner	2026-09-15T17:15:46.4052096Z retirado: /opt/hostedtoolcache/PyPy/3.11.15/x64/lib/pypy3.11/test/imghdrdata/python.pbm
evals	Retirar Python del runner	2026-09-15T17:15:46.4185144Z retirado: /opt/hostedtoolcache/PyPy/3.11.15/x64/lib/pypy3.11/test/imghdrdata/python.ras
evals	Retirar Python del runner	2026-09-15T17:15:46.4350549Z retirado: /opt/hostedtoolcache/PyPy/3.11.15/x64/lib/pypy3.11/test/imghdrdata/python-raw.jpg
evals	Retirar Python del runner	2026-09-15T17:15:46.4486619Z retirado: /opt/hostedtoolcache/PyPy/3.11.15/x64/lib/pypy3.11/test/test_email/data/python.jpg
evals	Retirar Python del runner	2026-09-15T17:15:46.4620469Z retirado: /opt/hostedtoolcache/PyPy/3.11.15/x64/lib/pypy3.11/test/test_email/data/python.sgi
evals	Retirar Python del runner	2026-09-15T17:15:46.4751717Z retirado: /opt/hostedtoolcache/PyPy/3.11.15/x64/lib/pypy3.11/test/test_email/data/python.tiff
evals	Retirar Python del runner	2026-09-15T17:15:46.4882515Z retirado: /opt/hostedtoolcache/PyPy/3.11.15/x64/lib/pypy3.11/test/test_email/data/python.bmp
evals	Retirar Python del runner	2026-09-15T17:15:46.5016301Z retirado: /opt/hostedtoolcache/PyPy/3.11.15/x64/lib/pypy3.11/test/test_email/data/python.gif
evals	Retirar Python del runner	2026-09-15T17:15:46.5148232Z retirado: /opt/hostedtoolcache/PyPy/3.11.15/x64/lib/pypy3.11/test/test_email/data/python.webp
evals	Retirar Python del runner	2026-09-15T17:15:46.5282365Z retirado: /opt/hostedtoolcache/PyPy/3.11.15/x64/lib/pypy3.11/test/test_email/data/python.png
evals	Retirar Python del runner	2026-09-15T17:15:46.5419461Z retirado: /opt/hostedtoolcache/PyPy/3.11.15/x64/lib/pypy3.11/test/test_email/data/python.ppm
evals	Retirar Python del runner	2026-09-15T17:15:46.5554603Z retirado: /opt/hostedtoolcache/PyPy/3.11.15/x64/lib/pypy3.11/test/test_email/data/python.pgm
evals	Retirar Python del runner	2026-09-15T17:15:46.5684569Z retirado: /opt/hostedtoolcache/PyPy/3.11.15/x64/lib/pypy3.11/test/test_email/data/python.exr
evals	Retirar Python del runner	2026-09-15T17:15:46.5814403Z retirado: /opt/hostedtoolcache/PyPy/3.11.15/x64/lib/pypy3.11/test/test_email/data/python.xbm
evals	Retirar Python del runner	2026-09-15T17:15:46.5945450Z retirado: /opt/hostedtoolcache/PyPy/3.11.15/x64/lib/pypy3.11/test/test_email/data/python.pbm
evals	Retirar Python del runner	2026-09-15T17:15:46.6074807Z retirado: /opt/hostedtoolcache/PyPy/3.11.15/x64/lib/pypy3.11/test/test_email/data/python.ras
evals	Retirar Python del runner	2026-09-15T17:15:46.6209700Z retirado: /opt/hostedtoolcache/PyPy/3.11.15/x64/lib/pypy3.11/test/pythoninfo.py
evals	Retirar Python del runner	2026-09-15T17:15:46.6341709Z retirado: /opt/hostedtoolcache/PyPy/3.11.15/x64/include/pypy3.11/Python.h
evals	Retirar Python del runner	2026-09-15T17:15:46.6473194Z retirado: /opt/hostedtoolcache/PyPy/3.11.15/x64/include/pypy3.11/pythonrun.h
evals	Retirar Python del runner	2026-09-15T17:15:46.6605408Z retirado: /opt/hostedtoolcache/PyPy/3.11.15/x64/include/pypy3.11/pypy_macros.h
evals	Retirar Python del runner	2026-09-15T17:15:46.6740309Z retirado: /opt/hostedtoolcache/PyPy/3.11.15/x64/include/pypy3.11/pypy_marshal_decl.h
evals	Retirar Python del runner	2026-09-15T17:15:46.6875835Z retirado: /opt/hostedtoolcache/PyPy/3.11.15/x64/include/pypy3.11/pypy_decl.h
evals	Retirar Python del runner	2026-09-15T17:15:46.7008583Z retirado: /opt/hostedtoolcache/PyPy/3.11.15/x64/include/pypy3.11/pypy_structmember_decl.h
evals	Retirar Python del runner	2026-09-15T17:15:46.7144524Z retirado: /opt/hostedtoolcache/PyPy/3.11.15/x64/PYPY_VERSION
evals	Retirar Python del runner	2026-09-15T17:15:46.7400627Z retirado: /opt/hostedtoolcache/PyPy/3.11.15/x64
evals	Retirar Python del runner	2026-09-15T17:15:46.9072366Z retirado: /opt/hostedtoolcache/PyPy/3.10.16/x64/lib/PYPY_PORTABLE_DEPS.txt
evals	Retirar Python del runner	2026-09-15T17:15:46.9205590Z retirado: /opt/hostedtoolcache/PyPy/3.10.16/x64/lib/pypy3.10/site-packages/pip/_vendor/pygments/lexers/__pycache__/python.pypy310.pyc
evals	Retirar Python del runner	2026-09-15T17:15:46.9342594Z retirado: /opt/hostedtoolcache/PyPy/3.10.16/x64/lib/pypy3.10/site-packages/pip/_vendor/pygments/lexers/python.py
evals	Retirar Python del runner	2026-09-15T17:15:46.9488063Z retirado: /opt/hostedtoolcache/PyPy/3.10.16/x64/lib/pypy3.10/_cffi_ssl/_cffi_src/openssl/pypy_win32_extra.py
evals	Retirar Python del runner	2026-09-15T17:15:46.9622562Z retirado: /opt/hostedtoolcache/PyPy/3.10.16/x64/lib/pypy3.10/hpy/devel/include/hpy/forbid_python_h/Python.h
evals	Retirar Python del runner	2026-09-15T17:15:46.9753614Z retirado: /opt/hostedtoolcache/PyPy/3.10.16/x64/lib/pypy3.10/idlelib/Icons/python.gif
evals	Retirar Python del runner	2026-09-15T17:15:46.9884562Z retirado: /opt/hostedtoolcache/PyPy/3.10.16/x64/lib/pypy3.10/test/imghdrdata/python.jpg
evals	Retirar Python del runner	2026-09-15T17:15:47.0020239Z retirado: /opt/hostedtoolcache/PyPy/3.10.16/x64/lib/pypy3.10/test/imghdrdata/python.sgi
evals	Retirar Python del runner	2026-09-15T17:15:47.0152035Z retirado: /opt/hostedtoolcache/PyPy/3.10.16/x64/lib/pypy3.10/test/imghdrdata/python.tiff
evals	Retirar Python del runner	2026-09-15T17:15:47.0286396Z retirado: /opt/hostedtoolcache/PyPy/3.10.16/x64/lib/pypy3.10/test/imghdrdata/python.bmp
evals	Retirar Python del runner	2026-09-15T17:15:47.0420561Z retirado: /opt/hostedtoolcache/PyPy/3.10.16/x64/lib/pypy3.10/test/imghdrdata/python.gif
evals	Retirar Python del runner	2026-09-15T17:15:47.0552827Z retirado: /opt/hostedtoolcache/PyPy/3.10.16/x64/lib/pypy3.10/test/imghdrdata/python.webp
evals	Retirar Python del runner	2026-09-15T17:15:47.0686707Z retirado: /opt/hostedtoolcache/PyPy/3.10.16/x64/lib/pypy3.10/test/imghdrdata/python.png
evals	Retirar Python del runner	2026-09-15T17:15:47.0824165Z retirado: /opt/hostedtoolcache/PyPy/3.10.16/x64/lib/pypy3.10/test/imghdrdata/python.ppm
evals	Retirar Python del runner	2026-09-15T17:15:47.0958881Z retirado: /opt/hostedtoolcache/PyPy/3.10.16/x64/lib/pypy3.10/test/imghdrdata/python.pgm
evals	Retirar Python del runner	2026-09-15T17:15:47.1096247Z retirado: /opt/hostedtoolcache/PyPy/3.10.16/x64/lib/pypy3.10/test/imghdrdata/python.exr
evals	Retirar Python del runner	2026-09-15T17:15:47.1227685Z retirado: /opt/hostedtoolcache/PyPy/3.10.16/x64/lib/pypy3.10/test/imghdrdata/python.xbm
evals	Retirar Python del runner	2026-09-15T17:15:47.1363106Z retirado: /opt/hostedtoolcache/PyPy/3.10.16/x64/lib/pypy3.10/test/imghdrdata/python.pbm
evals	Retirar Python del runner	2026-09-15T17:15:47.1497294Z retirado: /opt/hostedtoolcache/PyPy/3.10.16/x64/lib/pypy3.10/test/imghdrdata/python.ras
evals	Retirar Python del runner	2026-09-15T17:15:47.1630570Z retirado: /opt/hostedtoolcache/PyPy/3.10.16/x64/lib/pypy3.10/test/pythoninfo.py
evals	Retirar Python del runner	2026-09-15T17:15:47.1763401Z retirado: /opt/hostedtoolcache/PyPy/3.10.16/x64/include/pypy3.10/Python.h
evals	Retirar Python del runner	2026-09-15T17:15:47.1897864Z retirado: /opt/hostedtoolcache/PyPy/3.10.16/x64/include/pypy3.10/pythonrun.h
evals	Retirar Python del runner	2026-09-15T17:15:47.2030358Z retirado: /opt/hostedtoolcache/PyPy/3.10.16/x64/include/pypy3.10/pypy_macros.h
evals	Retirar Python del runner	2026-09-15T17:15:47.2163577Z retirado: /opt/hostedtoolcache/PyPy/3.10.16/x64/include/pypy3.10/pypy_marshal_decl.h
evals	Retirar Python del runner	2026-09-15T17:15:47.2298494Z retirado: /opt/hostedtoolcache/PyPy/3.10.16/x64/include/pypy3.10/pypy_decl.h
evals	Retirar Python del runner	2026-09-15T17:15:47.2429975Z retirado: /opt/hostedtoolcache/PyPy/3.10.16/x64/include/pypy3.10/pypy_structmember_decl.h
evals	Retirar Python del runner	2026-09-15T17:15:47.2560503Z retirado: /opt/hostedtoolcache/PyPy/3.10.16/x64/PYPY_VERSION
evals	Retirar Python del runner	2026-09-15T17:15:47.2814728Z retirado: /opt/hostedtoolcache/PyPy/3.10.16/x64
evals	Retirar Python del runner	2026-09-15T17:15:47.4461377Z retirado: /opt/hostedtoolcache/CodeQL/2.26.4/x64/codeql/Open-Source-Notices/python3.txt
evals	Retirar Python del runner	2026-09-15T17:15:47.4602759Z retirado: /opt/hostedtoolcache/CodeQL/2.26.4/x64/codeql/qlpacks/codeql/python-queries/1.8.9/experimental/semmle/python/libraries/PythonJose.qll
evals	Retirar Python del runner	2026-09-15T17:15:47.4741083Z retirado: /opt/hostedtoolcache/CodeQL/2.26.4/x64/codeql/qlpacks/codeql/python-queries/1.8.9/experimental/semmle/python/libraries/Python_JWT.qll
evals	Retirar Python del runner	2026-09-15T17:15:47.4873268Z retirado: /opt/hostedtoolcache/CodeQL/2.26.4/x64/codeql/qlpacks/codeql/python-queries/1.8.9/.codeql/libraries/codeql/python-all/7.2.4/python.qll
evals	Retirar Python del runner	2026-09-15T17:15:47.5001863Z retirado: /opt/hostedtoolcache/CodeQL/2.26.4/x64/codeql/qlpacks/codeql/python-queries/1.8.9/codeql-suites/python-code-quality-extended.qls
evals	Retirar Python del runner	2026-09-15T17:15:47.5132682Z retirado: /opt/hostedtoolcache/CodeQL/2.26.4/x64/codeql/qlpacks/codeql/python-queries/1.8.9/codeql-suites/python-security-experimental.qls
evals	Retirar Python del runner	2026-09-15T17:15:47.5261807Z retirado: /opt/hostedtoolcache/CodeQL/2.26.4/x64/codeql/qlpacks/codeql/python-queries/1.8.9/codeql-suites/python-code-scanning.qls
evals	Retirar Python del runner	2026-09-15T17:15:47.5390486Z retirado: /opt/hostedtoolcache/CodeQL/2.26.4/x64/codeql/qlpacks/codeql/python-queries/1.8.9/codeql-suites/python-lgtm.qls
evals	Retirar Python del runner	2026-09-15T17:15:47.5520937Z retirado: /opt/hostedtoolcache/CodeQL/2.26.4/x64/codeql/qlpacks/codeql/python-queries/1.8.9/codeql-suites/python-security-extended.qls
evals	Retirar Python del runner	2026-09-15T17:15:47.5654885Z retirado: /opt/hostedtoolcache/CodeQL/2.26.4/x64/codeql/qlpacks/codeql/python-queries/1.8.9/codeql-suites/python-lgtm-full.qls
evals	Retirar Python del runner	2026-09-15T17:15:47.5784577Z retirado: /opt/hostedtoolcache/CodeQL/2.26.4/x64/codeql/qlpacks/codeql/python-queries/1.8.9/codeql-suites/python-security-and-quality.qls
evals	Retirar Python del runner	2026-09-15T17:15:47.5912548Z retirado: /opt/hostedtoolcache/CodeQL/2.26.4/x64/codeql/qlpacks/codeql/python-queries/1.8.9/codeql-suites/python-code-quality.qls
evals	Retirar Python del runner	2026-09-15T17:15:47.6040475Z retirado: /opt/hostedtoolcache/CodeQL/2.26.4/x64/codeql/qlpacks/codeql/python-examples/0.0.0/.codeql/libraries/codeql/python-all/7.2.4/python.qll
evals	Retirar Python del runner	2026-09-15T17:15:47.6169330Z retirado: /opt/hostedtoolcache/CodeQL/2.26.4/x64/codeql/qlpacks/codeql/actions-all/0.6.0/ext/generated/composite-actions/python_mypy.model.yml
evals	Retirar Python del runner	2026-09-15T17:15:47.6305729Z retirado: /opt/hostedtoolcache/CodeQL/2.26.4/x64/codeql/qlpacks/codeql/actions-all/0.6.0/ext/generated/composite-actions/python-poetry_poetry.model.yml
evals	Retirar Python del runner	2026-09-15T17:15:47.6434464Z retirado: /opt/hostedtoolcache/CodeQL/2.26.4/x64/codeql/qlpacks/codeql/actions-all/0.6.0/ext/generated/reusable-workflows/python_cpython.model.yml
evals	Retirar Python del runner	2026-09-15T17:15:47.6564450Z retirado: /opt/hostedtoolcache/CodeQL/2.26.4/x64/codeql/qlpacks/codeql/python-all/7.2.4/python.qll
evals	Retirar Python del runner	2026-09-15T17:15:47.6694332Z retirado: /opt/hostedtoolcache/CodeQL/2.26.4/x64/codeql/qlpacks/codeql/actions-queries/0.6.34/.codeql/libraries/codeql/actions-all/0.6.0/ext/generated/composite-actions/python_mypy.model.yml
evals	Retirar Python del runner	2026-09-15T17:15:47.6822416Z retirado: /opt/hostedtoolcache/CodeQL/2.26.4/x64/codeql/qlpacks/codeql/actions-queries/0.6.34/.codeql/libraries/codeql/actions-all/0.6.0/ext/generated/composite-actions/python-poetry_poetry.model.yml
evals	Retirar Python del runner	2026-09-15T17:15:47.6952790Z retirado: /opt/hostedtoolcache/CodeQL/2.26.4/x64/codeql/qlpacks/codeql/actions-queries/0.6.34/.codeql/libraries/codeql/actions-all/0.6.0/ext/generated/reusable-workflows/python_cpython.model.yml
evals	Retirar Python del runner	2026-09-15T17:15:47.7083961Z retirado: /opt/hostedtoolcache/CodeQL/2.26.4/x64/codeql/python/tools/python3src.zip
evals	Retirar Python del runner	2026-09-15T17:15:47.7216414Z retirado: /opt/hostedtoolcache/CodeQL/2.26.4/x64/codeql/python/tools/python_setup.cmd
evals	Retirar Python del runner	2026-09-15T17:15:47.7347628Z retirado: /opt/hostedtoolcache/CodeQL/2.26.4/x64/codeql/python/tools/python_setup.sh
evals	Retirar Python del runner	2026-09-15T17:15:47.7478328Z retirado: /opt/hostedtoolcache/CodeQL/2.26.4/x64/codeql/python/tools/python_tracer.py
evals	Retirar Python del runner	2026-09-15T17:15:47.7610257Z retirado: /opt/hostedtoolcache/Python/3.14.7/x64/lib/libpython3.so
evals	Retirar Python del runner	2026-09-15T17:15:47.7745032Z retirado: /opt/hostedtoolcache/Python/3.14.7/x64/lib/pkgconfig/python-3.14-embed.pc
evals	Retirar Python del runner	2026-09-15T17:15:47.7874314Z retirado: /opt/hostedtoolcache/Python/3.14.7/x64/lib/pkgconfig/python3.pc
evals	Retirar Python del runner	2026-09-15T17:15:47.8061847Z retirado: /opt/hostedtoolcache/Python/3.14.7/x64/lib/pkgconfig/python3-embed.pc
evals	Retirar Python del runner	2026-09-15T17:15:47.8192433Z retirado: /opt/hostedtoolcache/Python/3.14.7/x64/lib/pkgconfig/python-3.14.pc
evals	Retirar Python del runner	2026-09-15T17:15:47.8322423Z retirado: /opt/hostedtoolcache/Python/3.14.7/x64/lib/python3.14/site-packages/pip/_vendor/pygments/lexers/__pycache__/python.cpython-314.pyc
evals	Retirar Python del runner	2026-09-15T17:15:47.8452090Z retirado: /opt/hostedtoolcache/Python/3.14.7/x64/lib/python3.14/site-packages/pip/_vendor/pygments/lexers/python.py
evals	Retirar Python del runner	2026-09-15T17:15:47.8580812Z retirado: /opt/hostedtoolcache/Python/3.14.7/x64/lib/python3.14/config-3.14-x86_64-linux-gnu/libpython3.14.a
evals	Retirar Python del runner	2026-09-15T17:15:47.8720696Z retirado: /opt/hostedtoolcache/Python/3.14.7/x64/lib/python3.14/config-3.14-x86_64-linux-gnu/python-config.py
evals	Retirar Python del runner	2026-09-15T17:15:47.8840500Z retirado: /opt/hostedtoolcache/Python/3.14.7/x64/lib/python3.14/config-3.14-x86_64-linux-gnu/python.o
evals	Retirar Python del runner	2026-09-15T17:15:47.8970402Z retirado: /opt/hostedtoolcache/Python/3.14.7/x64/lib/python3.14/idlelib/Icons/python.gif
evals	Retirar Python del runner	2026-09-15T17:15:47.9096723Z retirado: /opt/hostedtoolcache/Python/3.14.7/x64/lib/python3.14/test/__pycache__/pythoninfo.cpython-314.pyc
evals	Retirar Python del runner	2026-09-15T17:15:47.9225246Z retirado: /opt/hostedtoolcache/Python/3.14.7/x64/lib/python3.14/test/__pycache__/pythoninfo.cpython-314.opt-2.pyc
evals	Retirar Python del runner	2026-09-15T17:15:47.9353265Z retirado: /opt/hostedtoolcache/Python/3.14.7/x64/lib/python3.14/test/__pycache__/pythoninfo.cpython-314.opt-1.pyc
evals	Retirar Python del runner	2026-09-15T17:15:47.9482801Z retirado: /opt/hostedtoolcache/Python/3.14.7/x64/lib/python3.14/test/tkinterdata/python.gif
evals	Retirar Python del runner	2026-09-15T17:15:47.9610329Z retirado: /opt/hostedtoolcache/Python/3.14.7/x64/lib/python3.14/test/tkinterdata/python.png
evals	Retirar Python del runner	2026-09-15T17:15:47.9759611Z retirado: /opt/hostedtoolcache/Python/3.14.7/x64/lib/python3.14/test/tkinterdata/python.ppm
evals	Retirar Python del runner	2026-09-15T17:15:47.9888450Z retirado: /opt/hostedtoolcache/Python/3.14.7/x64/lib/python3.14/test/tkinterdata/python.pgm
evals	Retirar Python del runner	2026-09-15T17:15:48.0018013Z retirado: /opt/hostedtoolcache/Python/3.14.7/x64/lib/python3.14/test/tkinterdata/python.xbm
evals	Retirar Python del runner	2026-09-15T17:15:48.0148456Z retirado: /opt/hostedtoolcache/Python/3.14.7/x64/lib/python3.14/test/test_email/data/python.jpg
evals	Retirar Python del runner	2026-09-15T17:15:48.0277029Z retirado: /opt/hostedtoolcache/Python/3.14.7/x64/lib/python3.14/test/test_email/data/python.sgi
evals	Retirar Python del runner	2026-09-15T17:15:48.0406614Z retirado: /opt/hostedtoolcache/Python/3.14.7/x64/lib/python3.14/test/test_email/data/python.tiff
evals	Retirar Python del runner	2026-09-15T17:15:48.0537371Z retirado: /opt/hostedtoolcache/Python/3.14.7/x64/lib/python3.14/test/test_email/data/python.bmp
evals	Retirar Python del runner	2026-09-15T17:15:48.0666236Z retirado: /opt/hostedtoolcache/Python/3.14.7/x64/lib/python3.14/test/test_email/data/python.gif
evals	Retirar Python del runner	2026-09-15T17:15:48.0798562Z retirado: /opt/hostedtoolcache/Python/3.14.7/x64/lib/python3.14/test/test_email/data/python.webp
evals	Retirar Python del runner	2026-09-15T17:15:48.0930065Z retirado: /opt/hostedtoolcache/Python/3.14.7/x64/lib/python3.14/test/test_email/data/python.png
evals	Retirar Python del runner	2026-09-15T17:15:48.1057380Z retirado: /opt/hostedtoolcache/Python/3.14.7/x64/lib/python3.14/test/test_email/data/python.ppm
evals	Retirar Python del runner	2026-09-15T17:15:48.1186064Z retirado: /opt/hostedtoolcache/Python/3.14.7/x64/lib/python3.14/test/test_email/data/python.pgm
evals	Retirar Python del runner	2026-09-15T17:15:48.1313021Z retirado: /opt/hostedtoolcache/Python/3.14.7/x64/lib/python3.14/test/test_email/data/python.exr
evals	Retirar Python del runner	2026-09-15T17:15:48.1443025Z retirado: /opt/hostedtoolcache/Python/3.14.7/x64/lib/python3.14/test/test_email/data/python.xbm
evals	Retirar Python del runner	2026-09-15T17:15:48.1573033Z retirado: /opt/hostedtoolcache/Python/3.14.7/x64/lib/python3.14/test/test_email/data/python.pbm
evals	Retirar Python del runner	2026-09-15T17:15:48.1700603Z retirado: /opt/hostedtoolcache/Python/3.14.7/x64/lib/python3.14/test/test_email/data/python.ras
evals	Retirar Python del runner	2026-09-15T17:15:48.1829972Z retirado: /opt/hostedtoolcache/Python/3.14.7/x64/lib/python3.14/test/pythoninfo.py
evals	Retirar Python del runner	2026-09-15T17:15:48.1960070Z retirado: /opt/hostedtoolcache/Python/3.14.7/x64/lib/libpython3.14.so
evals	Retirar Python del runner	2026-09-15T17:15:48.2087018Z retirado: /opt/hostedtoolcache/Python/3.14.7/x64/lib/libpython3.14.so.1.0
evals	Retirar Python del runner	2026-09-15T17:15:48.2214990Z retirado: /opt/hostedtoolcache/Python/3.14.7/x64/include/python3.14/cpython/pythonrun.h
evals	Retirar Python del runner	2026-09-15T17:15:48.2346146Z retirado: /opt/hostedtoolcache/Python/3.14.7/x64/include/python3.14/Python.h
evals	Retirar Python del runner	2026-09-15T17:15:48.2480191Z retirado: /opt/hostedtoolcache/Python/3.14.7/x64/include/python3.14/pythonrun.h
evals	Retirar Python del runner	2026-09-15T17:15:48.2612493Z retirado: /opt/hostedtoolcache/Python/3.14.7/x64/share/man/man1/python3.1
evals	Retirar Python del runner	2026-09-15T17:15:48.2742957Z retirado: /opt/hostedtoolcache/Python/3.14.7/x64/share/man/man1/python3.14.1
evals	Retirar Python del runner	2026-09-15T17:15:48.2995560Z retirado: /opt/hostedtoolcache/Python/3.14.7/x64
evals	Retirar Python del runner	2026-09-15T17:15:48.5505940Z retirado: /opt/hostedtoolcache/Python/3.13.15/x64/lib/libpython3.so
evals	Retirar Python del runner	2026-09-15T17:15:48.5701096Z retirado: /opt/hostedtoolcache/Python/3.13.15/x64/lib/pkgconfig/python3.pc
evals	Retirar Python del runner	2026-09-15T17:15:48.5906992Z retirado: /opt/hostedtoolcache/Python/3.13.15/x64/lib/pkgconfig/python-3.13.pc
evals	Retirar Python del runner	2026-09-15T17:15:48.6121063Z retirado: /opt/hostedtoolcache/Python/3.13.15/x64/lib/pkgconfig/python3-embed.pc
evals	Retirar Python del runner	2026-09-15T17:15:48.6323871Z retirado: /opt/hostedtoolcache/Python/3.13.15/x64/lib/pkgconfig/python-3.13-embed.pc
evals	Retirar Python del runner	2026-09-15T17:15:48.6561930Z retirado: /opt/hostedtoolcache/Python/3.13.15/x64/lib/python3.13/site-packages/pip/_vendor/pygments/lexers/__pycache__/python.cpython-313.pyc
evals	Retirar Python del runner	2026-09-15T17:15:48.6726850Z retirado: /opt/hostedtoolcache/Python/3.13.15/x64/lib/python3.13/site-packages/pip/_vendor/pygments/lexers/python.py
evals	Retirar Python del runner	2026-09-15T17:15:48.6931505Z retirado: /opt/hostedtoolcache/Python/3.13.15/x64/lib/python3.13/idlelib/Icons/python.gif
evals	Retirar Python del runner	2026-09-15T17:15:48.7130944Z retirado: /opt/hostedtoolcache/Python/3.13.15/x64/lib/python3.13/test/__pycache__/pythoninfo.cpython-313.opt-2.pyc
evals	Retirar Python del runner	2026-09-15T17:15:48.7328873Z retirado: /opt/hostedtoolcache/Python/3.13.15/x64/lib/python3.13/test/__pycache__/pythoninfo.cpython-313.pyc
evals	Retirar Python del runner	2026-09-15T17:15:48.7537922Z retirado: /opt/hostedtoolcache/Python/3.13.15/x64/lib/python3.13/test/__pycache__/pythoninfo.cpython-313.opt-1.pyc
evals	Retirar Python del runner	2026-09-15T17:15:48.7749965Z retirado: /opt/hostedtoolcache/Python/3.13.15/x64/lib/python3.13/test/tkinterdata/python.gif
evals	Retirar Python del runner	2026-09-15T17:15:48.7955437Z retirado: /opt/hostedtoolcache/Python/3.13.15/x64/lib/python3.13/test/tkinterdata/python.png
evals	Retirar Python del runner	2026-09-15T17:15:48.8150473Z retirado: /opt/hostedtoolcache/Python/3.13.15/x64/lib/python3.13/test/tkinterdata/python.ppm
evals	Retirar Python del runner	2026-09-15T17:15:48.8326604Z retirado: /opt/hostedtoolcache/Python/3.13.15/x64/lib/python3.13/test/tkinterdata/python.pgm
evals	Retirar Python del runner	2026-09-15T17:15:48.8460103Z retirado: /opt/hostedtoolcache/Python/3.13.15/x64/lib/python3.13/test/tkinterdata/python.xbm
evals	Retirar Python del runner	2026-09-15T17:15:48.8591181Z retirado: /opt/hostedtoolcache/Python/3.13.15/x64/lib/python3.13/test/test_email/data/python.jpg
evals	Retirar Python del runner	2026-09-15T17:15:48.8730555Z retirado: /opt/hostedtoolcache/Python/3.13.15/x64/lib/python3.13/test/test_email/data/python.sgi
evals	Retirar Python del runner	2026-09-15T17:15:48.8864269Z retirado: /opt/hostedtoolcache/Python/3.13.15/x64/lib/python3.13/test/test_email/data/python.tiff
evals	Retirar Python del runner	2026-09-15T17:15:48.8998847Z retirado: /opt/hostedtoolcache/Python/3.13.15/x64/lib/python3.13/test/test_email/data/python.bmp
evals	Retirar Python del runner	2026-09-15T17:15:48.9133783Z retirado: /opt/hostedtoolcache/Python/3.13.15/x64/lib/python3.13/test/test_email/data/python.gif
evals	Retirar Python del runner	2026-09-15T17:15:48.9267272Z retirado: /opt/hostedtoolcache/Python/3.13.15/x64/lib/python3.13/test/test_email/data/python.webp
evals	Retirar Python del runner	2026-09-15T17:15:48.9400960Z retirado: /opt/hostedtoolcache/Python/3.13.15/x64/lib/python3.13/test/test_email/data/python.png
evals	Retirar Python del runner	2026-09-15T17:15:48.9534431Z retirado: /opt/hostedtoolcache/Python/3.13.15/x64/lib/python3.13/test/test_email/data/python.ppm
evals	Retirar Python del runner	2026-09-15T17:15:48.9669487Z retirado: /opt/hostedtoolcache/Python/3.13.15/x64/lib/python3.13/test/test_email/data/python.pgm
evals	Retirar Python del runner	2026-09-15T17:15:48.9804179Z retirado: /opt/hostedtoolcache/Python/3.13.15/x64/lib/python3.13/test/test_email/data/python.exr
evals	Retirar Python del runner	2026-09-15T17:15:48.9941595Z retirado: /opt/hostedtoolcache/Python/3.13.15/x64/lib/python3.13/test/test_email/data/python.xbm
evals	Retirar Python del runner	2026-09-15T17:15:49.0072632Z retirado: /opt/hostedtoolcache/Python/3.13.15/x64/lib/python3.13/test/test_email/data/python.pbm
evals	Retirar Python del runner	2026-09-15T17:15:49.0207775Z retirado: /opt/hostedtoolcache/Python/3.13.15/x64/lib/python3.13/test/test_email/data/python.ras
evals	Retirar Python del runner	2026-09-15T17:15:49.0342341Z retirado: /opt/hostedtoolcache/Python/3.13.15/x64/lib/python3.13/test/pythoninfo.py
evals	Retirar Python del runner	2026-09-15T17:15:49.0477784Z retirado: /opt/hostedtoolcache/Python/3.13.15/x64/lib/python3.13/config-3.13-x86_64-linux-gnu/libpython3.13.a
evals	Retirar Python del runner	2026-09-15T17:15:49.0615907Z retirado: /opt/hostedtoolcache/Python/3.13.15/x64/lib/python3.13/config-3.13-x86_64-linux-gnu/python-config.py
evals	Retirar Python del runner	2026-09-15T17:15:49.0746473Z retirado: /opt/hostedtoolcache/Python/3.13.15/x64/lib/python3.13/config-3.13-x86_64-linux-gnu/python.o
evals	Retirar Python del runner	2026-09-15T17:15:49.0876860Z retirado: /opt/hostedtoolcache/Python/3.13.15/x64/lib/libpython3.13.so
evals	Retirar Python del runner	2026-09-15T17:15:49.1005628Z retirado: /opt/hostedtoolcache/Python/3.13.15/x64/lib/libpython3.13.so.1.0
evals	Retirar Python del runner	2026-09-15T17:15:49.1140419Z retirado: /opt/hostedtoolcache/Python/3.13.15/x64/include/python3.13/cpython/pythonrun.h
evals	Retirar Python del runner	2026-09-15T17:15:49.1278460Z retirado: /opt/hostedtoolcache/Python/3.13.15/x64/include/python3.13/Python.h
evals	Retirar Python del runner	2026-09-15T17:15:49.1441391Z retirado: /opt/hostedtoolcache/Python/3.13.15/x64/include/python3.13/pythonrun.h
evals	Retirar Python del runner	2026-09-15T17:15:49.1555507Z retirado: /opt/hostedtoolcache/Python/3.13.15/x64/share/man/man1/python3.1
evals	Retirar Python del runner	2026-09-15T17:15:49.1687627Z retirado: /opt/hostedtoolcache/Python/3.13.15/x64/share/man/man1/python3.13.1
evals	Retirar Python del runner	2026-09-15T17:15:49.1942057Z retirado: /opt/hostedtoolcache/Python/3.13.15/x64
evals	Retirar Python del runner	2026-09-15T17:15:49.3764029Z retirado: /opt/hostedtoolcache/Python/3.10.21/x64/lib/libpython3.so
evals	Retirar Python del runner	2026-09-15T17:15:49.3894134Z retirado: /opt/hostedtoolcache/Python/3.10.21/x64/lib/python3.10/site-packages/pip/_vendor/pygments/lexers/__pycache__/python.cpython-310.pyc
evals	Retirar Python del runner	2026-09-15T17:15:49.4021741Z retirado: /opt/hostedtoolcache/Python/3.10.21/x64/lib/python3.10/site-packages/pip/_vendor/pygments/lexers/python.py
evals	Retirar Python del runner	2026-09-15T17:15:49.4151705Z retirado: /opt/hostedtoolcache/Python/3.10.21/x64/lib/python3.10/config-3.10-x86_64-linux-gnu/libpython3.10.a
evals	Retirar Python del runner	2026-09-15T17:15:49.4278520Z retirado: /opt/hostedtoolcache/Python/3.10.21/x64/lib/python3.10/config-3.10-x86_64-linux-gnu/python-config.py
evals	Retirar Python del runner	2026-09-15T17:15:49.4406374Z retirado: /opt/hostedtoolcache/Python/3.10.21/x64/lib/python3.10/config-3.10-x86_64-linux-gnu/python.o
evals	Retirar Python del runner	2026-09-15T17:15:49.4531703Z retirado: /opt/hostedtoolcache/Python/3.10.21/x64/lib/python3.10/idlelib/Icons/python.gif
evals	Retirar Python del runner	2026-09-15T17:15:49.4657275Z retirado: /opt/hostedtoolcache/Python/3.10.21/x64/lib/python3.10/test/__pycache__/pythoninfo.cpython-310.pyc
evals	Retirar Python del runner	2026-09-15T17:15:49.4781849Z retirado: /opt/hostedtoolcache/Python/3.10.21/x64/lib/python3.10/test/__pycache__/pythoninfo.cpython-310.opt-2.pyc
evals	Retirar Python del runner	2026-09-15T17:15:49.4910133Z retirado: /opt/hostedtoolcache/Python/3.10.21/x64/lib/python3.10/test/__pycache__/pythoninfo.cpython-310.opt-1.pyc
evals	Retirar Python del runner	2026-09-15T17:15:49.5044810Z retirado: /opt/hostedtoolcache/Python/3.10.21/x64/lib/python3.10/test/imghdrdata/python.jpg
evals	Retirar Python del runner	2026-09-15T17:15:49.5174416Z retirado: /opt/hostedtoolcache/Python/3.10.21/x64/lib/python3.10/test/imghdrdata/python.sgi
evals	Retirar Python del runner	2026-09-15T17:15:49.5305841Z retirado: /opt/hostedtoolcache/Python/3.10.21/x64/lib/python3.10/test/imghdrdata/python.tiff
evals	Retirar Python del runner	2026-09-15T17:15:49.5441286Z retirado: /opt/hostedtoolcache/Python/3.10.21/x64/lib/python3.10/test/imghdrdata/python.bmp
evals	Retirar Python del runner	2026-09-15T17:15:49.5582567Z retirado: /opt/hostedtoolcache/Python/3.10.21/x64/lib/python3.10/test/imghdrdata/python.gif
evals	Retirar Python del runner	2026-09-15T17:15:49.5710430Z retirado: /opt/hostedtoolcache/Python/3.10.21/x64/lib/python3.10/test/imghdrdata/python.webp
evals	Retirar Python del runner	2026-09-15T17:15:49.5842459Z retirado: /opt/hostedtoolcache/Python/3.10.21/x64/lib/python3.10/test/imghdrdata/python.png
evals	Retirar Python del runner	2026-09-15T17:15:49.5966731Z retirado: /opt/hostedtoolcache/Python/3.10.21/x64/lib/python3.10/test/imghdrdata/python.ppm
evals	Retirar Python del runner	2026-09-15T17:15:49.6096743Z retirado: /opt/hostedtoolcache/Python/3.10.21/x64/lib/python3.10/test/imghdrdata/python.pgm
evals	Retirar Python del runner	2026-09-15T17:15:49.6232063Z retirado: /opt/hostedtoolcache/Python/3.10.21/x64/lib/python3.10/test/imghdrdata/python.exr
evals	Retirar Python del runner	2026-09-15T17:15:49.6361262Z retirado: /opt/hostedtoolcache/Python/3.10.21/x64/lib/python3.10/test/imghdrdata/python.xbm
evals	Retirar Python del runner	2026-09-15T17:15:49.6488608Z retirado: /opt/hostedtoolcache/Python/3.10.21/x64/lib/python3.10/test/imghdrdata/python.pbm
evals	Retirar Python del runner	2026-09-15T17:15:49.6617227Z retirado: /opt/hostedtoolcache/Python/3.10.21/x64/lib/python3.10/test/imghdrdata/python.ras
evals	Retirar Python del runner	2026-09-15T17:15:49.6743286Z retirado: /opt/hostedtoolcache/Python/3.10.21/x64/lib/python3.10/test/pythoninfo.py
evals	Retirar Python del runner	2026-09-15T17:15:49.6871650Z retirado: /opt/hostedtoolcache/Python/3.10.21/x64/lib/pkgconfig/python3.pc
evals	Retirar Python del runner	2026-09-15T17:15:49.6995995Z retirado: /opt/hostedtoolcache/Python/3.10.21/x64/lib/pkgconfig/python3-embed.pc
evals	Retirar Python del runner	2026-09-15T17:15:49.7124792Z retirado: /opt/hostedtoolcache/Python/3.10.21/x64/lib/pkgconfig/python-3.10.pc
evals	Retirar Python del runner	2026-09-15T17:15:49.7255676Z retirado: /opt/hostedtoolcache/Python/3.10.21/x64/lib/pkgconfig/python-3.10-embed.pc
evals	Retirar Python del runner	2026-09-15T17:15:49.7386846Z retirado: /opt/hostedtoolcache/Python/3.10.21/x64/lib/libpython3.10.so
evals	Retirar Python del runner	2026-09-15T17:15:49.7517283Z retirado: /opt/hostedtoolcache/Python/3.10.21/x64/lib/libpython3.10.so.1.0
evals	Retirar Python del runner	2026-09-15T17:15:49.7648256Z retirado: /opt/hostedtoolcache/Python/3.10.21/x64/include/python3.10/cpython/pythonrun.h
evals	Retirar Python del runner	2026-09-15T17:15:49.7778598Z retirado: /opt/hostedtoolcache/Python/3.10.21/x64/include/python3.10/Python.h
evals	Retirar Python del runner	2026-09-15T17:15:49.7906838Z retirado: /opt/hostedtoolcache/Python/3.10.21/x64/include/python3.10/pythonrun.h
evals	Retirar Python del runner	2026-09-15T17:15:49.8038779Z retirado: /opt/hostedtoolcache/Python/3.10.21/x64/share/man/man1/python3.10.1
evals	Retirar Python del runner	2026-09-15T17:15:49.8223302Z retirado: /opt/hostedtoolcache/Python/3.10.21/x64/share/man/man1/python3.1
evals	Retirar Python del runner	2026-09-15T17:15:49.8476366Z retirado: /opt/hostedtoolcache/Python/3.10.21/x64
evals	Retirar Python del runner	2026-09-15T17:15:50.0335525Z retirado: /opt/hostedtoolcache/Python/3.12.14/x64/lib/libpython3.so
evals	Retirar Python del runner	2026-09-15T17:15:50.0467485Z retirado: /opt/hostedtoolcache/Python/3.12.14/x64/lib/pkgconfig/python-3.12.pc
evals	Retirar Python del runner	2026-09-15T17:15:50.0655919Z retirado: /opt/hostedtoolcache/Python/3.12.14/x64/lib/pkgconfig/python3.pc
evals	Retirar Python del runner	2026-09-15T17:15:50.0788228Z retirado: /opt/hostedtoolcache/Python/3.12.14/x64/lib/pkgconfig/python-3.12-embed.pc
evals	Retirar Python del runner	2026-09-15T17:15:50.0975225Z retirado: /opt/hostedtoolcache/Python/3.12.14/x64/lib/pkgconfig/python3-embed.pc
evals	Retirar Python del runner	2026-09-15T17:15:50.1105980Z retirado: /opt/hostedtoolcache/Python/3.12.14/x64/lib/libpython3.12.so
evals	Retirar Python del runner	2026-09-15T17:15:50.1236158Z retirado: /opt/hostedtoolcache/Python/3.12.14/x64/lib/python3.12/site-packages/pip/_vendor/pygments/lexers/__pycache__/python.cpython-312.pyc
evals	Retirar Python del runner	2026-09-15T17:15:50.1368207Z retirado: /opt/hostedtoolcache/Python/3.12.14/x64/lib/python3.12/site-packages/pip/_vendor/pygments/lexers/python.py
evals	Retirar Python del runner	2026-09-15T17:15:50.1502225Z retirado: /opt/hostedtoolcache/Python/3.12.14/x64/lib/python3.12/idlelib/Icons/python.gif
evals	Retirar Python del runner	2026-09-15T17:15:50.1635387Z retirado: /opt/hostedtoolcache/Python/3.12.14/x64/lib/python3.12/test/__pycache__/pythoninfo.cpython-312.opt-2.pyc
evals	Retirar Python del runner	2026-09-15T17:15:50.1763607Z retirado: /opt/hostedtoolcache/Python/3.12.14/x64/lib/python3.12/test/__pycache__/pythoninfo.cpython-312.opt-1.pyc
evals	Retirar Python del runner	2026-09-15T17:15:50.1894010Z retirado: /opt/hostedtoolcache/Python/3.12.14/x64/lib/python3.12/test/__pycache__/pythoninfo.cpython-312.pyc
evals	Retirar Python del runner	2026-09-15T17:15:50.2027414Z retirado: /opt/hostedtoolcache/Python/3.12.14/x64/lib/python3.12/test/imghdrdata/python.jpg
evals	Retirar Python del runner	2026-09-15T17:15:50.2158589Z retirado: /opt/hostedtoolcache/Python/3.12.14/x64/lib/python3.12/test/imghdrdata/python.sgi
evals	Retirar Python del runner	2026-09-15T17:15:50.2288912Z retirado: /opt/hostedtoolcache/Python/3.12.14/x64/lib/python3.12/test/imghdrdata/python.tiff
evals	Retirar Python del runner	2026-09-15T17:15:50.2415433Z retirado: /opt/hostedtoolcache/Python/3.12.14/x64/lib/python3.12/test/imghdrdata/python.bmp
evals	Retirar Python del runner	2026-09-15T17:15:50.2544729Z retirado: /opt/hostedtoolcache/Python/3.12.14/x64/lib/python3.12/test/imghdrdata/python.gif
evals	Retirar Python del runner	2026-09-15T17:15:50.2670453Z retirado: /opt/hostedtoolcache/Python/3.12.14/x64/lib/python3.12/test/imghdrdata/python.webp
evals	Retirar Python del runner	2026-09-15T17:15:50.2798469Z retirado: /opt/hostedtoolcache/Python/3.12.14/x64/lib/python3.12/test/imghdrdata/python.png
evals	Retirar Python del runner	2026-09-15T17:15:50.2925189Z retirado: /opt/hostedtoolcache/Python/3.12.14/x64/lib/python3.12/test/imghdrdata/python.ppm
evals	Retirar Python del runner	2026-09-15T17:15:50.3051909Z retirado: /opt/hostedtoolcache/Python/3.12.14/x64/lib/python3.12/test/imghdrdata/python.pgm
evals	Retirar Python del runner	2026-09-15T17:15:50.3178897Z retirado: /opt/hostedtoolcache/Python/3.12.14/x64/lib/python3.12/test/imghdrdata/python.exr
evals	Retirar Python del runner	2026-09-15T17:15:50.3305507Z retirado: /opt/hostedtoolcache/Python/3.12.14/x64/lib/python3.12/test/imghdrdata/python.xbm
evals	Retirar Python del runner	2026-09-15T17:15:50.3433299Z retirado: /opt/hostedtoolcache/Python/3.12.14/x64/lib/python3.12/test/imghdrdata/python.pbm
evals	Retirar Python del runner	2026-09-15T17:15:50.3558361Z retirado: /opt/hostedtoolcache/Python/3.12.14/x64/lib/python3.12/test/imghdrdata/python.ras
evals	Retirar Python del runner	2026-09-15T17:15:50.3687968Z retirado: /opt/hostedtoolcache/Python/3.12.14/x64/lib/python3.12/test/imghdrdata/python-raw.jpg
evals	Retirar Python del runner	2026-09-15T17:15:50.3817145Z retirado: /opt/hostedtoolcache/Python/3.12.14/x64/lib/python3.12/test/test_email/data/python.jpg
evals	Retirar Python del runner	2026-09-15T17:15:50.3948345Z retirado: /opt/hostedtoolcache/Python/3.12.14/x64/lib/python3.12/test/test_email/data/python.sgi
evals	Retirar Python del runner	2026-09-15T17:15:50.4080452Z retirado: /opt/hostedtoolcache/Python/3.12.14/x64/lib/python3.12/test/test_email/data/python.tiff
evals	Retirar Python del runner	2026-09-15T17:15:50.4210711Z retirado: /opt/hostedtoolcache/Python/3.12.14/x64/lib/python3.12/test/test_email/data/python.bmp
evals	Retirar Python del runner	2026-09-15T17:15:50.4342241Z retirado: /opt/hostedtoolcache/Python/3.12.14/x64/lib/python3.12/test/test_email/data/python.gif
evals	Retirar Python del runner	2026-09-15T17:15:50.4470290Z retirado: /opt/hostedtoolcache/Python/3.12.14/x64/lib/python3.12/test/test_email/data/python.webp
evals	Retirar Python del runner	2026-09-15T17:15:50.4598009Z retirado: /opt/hostedtoolcache/Python/3.12.14/x64/lib/python3.12/test/test_email/data/python.png
evals	Retirar Python del runner	2026-09-15T17:15:50.4726104Z retirado: /opt/hostedtoolcache/Python/3.12.14/x64/lib/python3.12/test/test_email/data/python.ppm
evals	Retirar Python del runner	2026-09-15T17:15:50.4857365Z retirado: /opt/hostedtoolcache/Python/3.12.14/x64/lib/python3.12/test/test_email/data/python.pgm
evals	Retirar Python del runner	2026-09-15T17:15:50.4988225Z retirado: /opt/hostedtoolcache/Python/3.12.14/x64/lib/python3.12/test/test_email/data/python.exr
evals	Retirar Python del runner	2026-09-15T17:15:50.5117901Z retirado: /opt/hostedtoolcache/Python/3.12.14/x64/lib/python3.12/test/test_email/data/python.xbm
evals	Retirar Python del runner	2026-09-15T17:15:50.5252693Z retirado: /opt/hostedtoolcache/Python/3.12.14/x64/lib/python3.12/test/test_email/data/python.pbm
evals	Retirar Python del runner	2026-09-15T17:15:50.5382002Z retirado: /opt/hostedtoolcache/Python/3.12.14/x64/lib/python3.12/test/test_email/data/python.ras
evals	Retirar Python del runner	2026-09-15T17:15:50.5514553Z retirado: /opt/hostedtoolcache/Python/3.12.14/x64/lib/python3.12/test/pythoninfo.py
evals	Retirar Python del runner	2026-09-15T17:15:50.5680510Z retirado: /opt/hostedtoolcache/Python/3.12.14/x64/lib/python3.12/config-3.12-x86_64-linux-gnu/libpython3.12.a
evals	Retirar Python del runner	2026-09-15T17:15:50.5813499Z retirado: /opt/hostedtoolcache/Python/3.12.14/x64/lib/python3.12/config-3.12-x86_64-linux-gnu/python-config.py
evals	Retirar Python del runner	2026-09-15T17:15:50.5944302Z retirado: /opt/hostedtoolcache/Python/3.12.14/x64/lib/python3.12/config-3.12-x86_64-linux-gnu/python.o
evals	Retirar Python del runner	2026-09-15T17:15:50.6070596Z retirado: /opt/hostedtoolcache/Python/3.12.14/x64/lib/libpython3.12.so.1.0
evals	Retirar Python del runner	2026-09-15T17:15:50.6201915Z retirado: /opt/hostedtoolcache/Python/3.12.14/x64/include/python3.12/cpython/pythonrun.h
evals	Retirar Python del runner	2026-09-15T17:15:50.6331851Z retirado: /opt/hostedtoolcache/Python/3.12.14/x64/include/python3.12/Python.h
evals	Retirar Python del runner	2026-09-15T17:15:50.6461447Z retirado: /opt/hostedtoolcache/Python/3.12.14/x64/include/python3.12/pythonrun.h
evals	Retirar Python del runner	2026-09-15T17:15:50.6592872Z retirado: /opt/hostedtoolcache/Python/3.12.14/x64/share/man/man1/python3.1
evals	Retirar Python del runner	2026-09-15T17:15:50.6720317Z retirado: /opt/hostedtoolcache/Python/3.12.14/x64/share/man/man1/python3.12.1
evals	Retirar Python del runner	2026-09-15T17:15:50.6970475Z retirado: /opt/hostedtoolcache/Python/3.12.14/x64
evals	Retirar Python del runner	2026-09-15T17:15:50.8821130Z retirado: /opt/hostedtoolcache/Python/3.11.16/x64/lib/libpython3.so
evals	Retirar Python del runner	2026-09-15T17:15:50.8947896Z retirado: /opt/hostedtoolcache/Python/3.11.16/x64/lib/pkgconfig/python3.pc
evals	Retirar Python del runner	2026-09-15T17:15:50.9074250Z retirado: /opt/hostedtoolcache/Python/3.11.16/x64/lib/pkgconfig/python3-embed.pc
evals	Retirar Python del runner	2026-09-15T17:15:50.9207795Z retirado: /opt/hostedtoolcache/Python/3.11.16/x64/lib/pkgconfig/python-3.11-embed.pc
evals	Retirar Python del runner	2026-09-15T17:15:50.9338247Z retirado: /opt/hostedtoolcache/Python/3.11.16/x64/lib/pkgconfig/python-3.11.pc
evals	Retirar Python del runner	2026-09-15T17:15:50.9470761Z retirado: /opt/hostedtoolcache/Python/3.11.16/x64/lib/python3.11/site-packages/pip/_vendor/pygments/lexers/__pycache__/python.cpython-311.pyc
evals	Retirar Python del runner	2026-09-15T17:15:50.9600814Z retirado: /opt/hostedtoolcache/Python/3.11.16/x64/lib/python3.11/site-packages/pip/_vendor/pygments/lexers/python.py
evals	Retirar Python del runner	2026-09-15T17:15:50.9727070Z retirado: /opt/hostedtoolcache/Python/3.11.16/x64/lib/python3.11/config-3.11-x86_64-linux-gnu/libpython3.11.a
evals	Retirar Python del runner	2026-09-15T17:15:50.9856625Z retirado: /opt/hostedtoolcache/Python/3.11.16/x64/lib/python3.11/config-3.11-x86_64-linux-gnu/python-config.py
evals	Retirar Python del runner	2026-09-15T17:15:50.9983868Z retirado: /opt/hostedtoolcache/Python/3.11.16/x64/lib/python3.11/config-3.11-x86_64-linux-gnu/python.o
evals	Retirar Python del runner	2026-09-15T17:15:51.0112907Z retirado: /opt/hostedtoolcache/Python/3.11.16/x64/lib/python3.11/idlelib/Icons/python.gif
evals	Retirar Python del runner	2026-09-15T17:15:51.0245726Z retirado: /opt/hostedtoolcache/Python/3.11.16/x64/lib/python3.11/test/__pycache__/pythoninfo.cpython-311.pyc
evals	Retirar Python del runner	2026-09-15T17:15:51.0370594Z retirado: /opt/hostedtoolcache/Python/3.11.16/x64/lib/python3.11/test/__pycache__/pythoninfo.cpython-311.opt-1.pyc
evals	Retirar Python del runner	2026-09-15T17:15:51.0495731Z retirado: /opt/hostedtoolcache/Python/3.11.16/x64/lib/python3.11/test/__pycache__/pythoninfo.cpython-311.opt-2.pyc
evals	Retirar Python del runner	2026-09-15T17:15:51.0622724Z retirado: /opt/hostedtoolcache/Python/3.11.16/x64/lib/python3.11/test/imghdrdata/python.jpg
evals	Retirar Python del runner	2026-09-15T17:15:51.0752263Z retirado: /opt/hostedtoolcache/Python/3.11.16/x64/lib/python3.11/test/imghdrdata/python.sgi
evals	Retirar Python del runner	2026-09-15T17:15:51.0879914Z retirado: /opt/hostedtoolcache/Python/3.11.16/x64/lib/python3.11/test/imghdrdata/python.tiff
evals	Retirar Python del runner	2026-09-15T17:15:51.1009977Z retirado: /opt/hostedtoolcache/Python/3.11.16/x64/lib/python3.11/test/imghdrdata/python.bmp
evals	Retirar Python del runner	2026-09-15T17:15:51.1135372Z retirado: /opt/hostedtoolcache/Python/3.11.16/x64/lib/python3.11/test/imghdrdata/python.gif
evals	Retirar Python del runner	2026-09-15T17:15:51.1260623Z retirado: /opt/hostedtoolcache/Python/3.11.16/x64/lib/python3.11/test/imghdrdata/python.webp
evals	Retirar Python del runner	2026-09-15T17:15:51.1386593Z retirado: /opt/hostedtoolcache/Python/3.11.16/x64/lib/python3.11/test/imghdrdata/python.png
evals	Retirar Python del runner	2026-09-15T17:15:51.1514918Z retirado: /opt/hostedtoolcache/Python/3.11.16/x64/lib/python3.11/test/imghdrdata/python.ppm
evals	Retirar Python del runner	2026-09-15T17:15:51.1653903Z retirado: /opt/hostedtoolcache/Python/3.11.16/x64/lib/python3.11/test/imghdrdata/python.pgm
evals	Retirar Python del runner	2026-09-15T17:15:51.1791088Z retirado: /opt/hostedtoolcache/Python/3.11.16/x64/lib/python3.11/test/imghdrdata/python.exr
evals	Retirar Python del runner	2026-09-15T17:15:51.1924940Z retirado: /opt/hostedtoolcache/Python/3.11.16/x64/lib/python3.11/test/imghdrdata/python.xbm
evals	Retirar Python del runner	2026-09-15T17:15:51.2058410Z retirado: /opt/hostedtoolcache/Python/3.11.16/x64/lib/python3.11/test/imghdrdata/python.pbm
evals	Retirar Python del runner	2026-09-15T17:15:51.2191507Z retirado: /opt/hostedtoolcache/Python/3.11.16/x64/lib/python3.11/test/imghdrdata/python.ras
evals	Retirar Python del runner	2026-09-15T17:15:51.2320079Z retirado: /opt/hostedtoolcache/Python/3.11.16/x64/lib/python3.11/test/imghdrdata/python-raw.jpg
evals	Retirar Python del runner	2026-09-15T17:15:51.2448065Z retirado: /opt/hostedtoolcache/Python/3.11.16/x64/lib/python3.11/test/test_email/data/python.jpg
evals	Retirar Python del runner	2026-09-15T17:15:51.2581611Z retirado: /opt/hostedtoolcache/Python/3.11.16/x64/lib/python3.11/test/test_email/data/python.sgi
evals	Retirar Python del runner	2026-09-15T17:15:51.2707323Z retirado: /opt/hostedtoolcache/Python/3.11.16/x64/lib/python3.11/test/test_email/data/python.tiff
evals	Retirar Python del runner	2026-09-15T17:15:51.2840739Z retirado: /opt/hostedtoolcache/Python/3.11.16/x64/lib/python3.11/test/test_email/data/python.bmp
evals	Retirar Python del runner	2026-09-15T17:15:51.2972129Z retirado: /opt/hostedtoolcache/Python/3.11.16/x64/lib/python3.11/test/test_email/data/python.gif
evals	Retirar Python del runner	2026-09-15T17:15:51.3099447Z retirado: /opt/hostedtoolcache/Python/3.11.16/x64/lib/python3.11/test/test_email/data/python.webp
evals	Retirar Python del runner	2026-09-15T17:15:51.3231586Z retirado: /opt/hostedtoolcache/Python/3.11.16/x64/lib/python3.11/test/test_email/data/python.png
evals	Retirar Python del runner	2026-09-15T17:15:51.3367517Z retirado: /opt/hostedtoolcache/Python/3.11.16/x64/lib/python3.11/test/test_email/data/python.ppm
evals	Retirar Python del runner	2026-09-15T17:15:51.3495835Z retirado: /opt/hostedtoolcache/Python/3.11.16/x64/lib/python3.11/test/test_email/data/python.pgm
evals	Retirar Python del runner	2026-09-15T17:15:51.3625228Z retirado: /opt/hostedtoolcache/Python/3.11.16/x64/lib/python3.11/test/test_email/data/python.exr
evals	Retirar Python del runner	2026-09-15T17:15:51.3752269Z retirado: /opt/hostedtoolcache/Python/3.11.16/x64/lib/python3.11/test/test_email/data/python.xbm
evals	Retirar Python del runner	2026-09-15T17:15:51.3881885Z retirado: /opt/hostedtoolcache/Python/3.11.16/x64/lib/python3.11/test/test_email/data/python.pbm
evals	Retirar Python del runner	2026-09-15T17:15:51.4013406Z retirado: /opt/hostedtoolcache/Python/3.11.16/x64/lib/python3.11/test/test_email/data/python.ras
evals	Retirar Python del runner	2026-09-15T17:15:51.4145527Z retirado: /opt/hostedtoolcache/Python/3.11.16/x64/lib/python3.11/test/pythoninfo.py
evals	Retirar Python del runner	2026-09-15T17:15:51.4275581Z retirado: /opt/hostedtoolcache/Python/3.11.16/x64/lib/libpython3.11.so
evals	Retirar Python del runner	2026-09-15T17:15:51.4404522Z retirado: /opt/hostedtoolcache/Python/3.11.16/x64/lib/libpython3.11.so.1.0
evals	Retirar Python del runner	2026-09-15T17:15:51.4533308Z retirado: /opt/hostedtoolcache/Python/3.11.16/x64/include/python3.11/cpython/pythonrun.h
evals	Retirar Python del runner	2026-09-15T17:15:51.4660474Z retirado: /opt/hostedtoolcache/Python/3.11.16/x64/include/python3.11/Python.h
evals	Retirar Python del runner	2026-09-15T17:15:51.4784875Z retirado: /opt/hostedtoolcache/Python/3.11.16/x64/include/python3.11/pythonrun.h
evals	Retirar Python del runner	2026-09-15T17:15:51.4916637Z retirado: /opt/hostedtoolcache/Python/3.11.16/x64/share/man/man1/python3.1
evals	Retirar Python del runner	2026-09-15T17:15:51.5043897Z retirado: /opt/hostedtoolcache/Python/3.11.16/x64/share/man/man1/python3.11.1
evals	Retirar Python del runner	2026-09-15T17:15:51.5297842Z retirado: /opt/hostedtoolcache/Python/3.11.16/x64
evals	Retirar Python del runner	2026-09-15T17:15:51.7151459Z retirado: /opt/az/lib/pkgconfig/python-3.14-embed.pc
evals	Retirar Python del runner	2026-09-15T17:15:51.7279007Z retirado: /opt/az/lib/pkgconfig/python3.pc
evals	Retirar Python del runner	2026-09-15T17:15:51.7459450Z retirado: /opt/az/lib/pkgconfig/python3-embed.pc
evals	Retirar Python del runner	2026-09-15T17:15:51.7590440Z retirado: /opt/az/lib/pkgconfig/python-3.14.pc
evals	Retirar Python del runner	2026-09-15T17:15:51.7725138Z retirado: /opt/az/lib/python3.14/site-packages/pygments/lexers/__pycache__/python.cpython-314.pyc
evals	Retirar Python del runner	2026-09-15T17:15:51.7856178Z retirado: /opt/az/lib/python3.14/site-packages/pygments/lexers/python.py
evals	Retirar Python del runner	2026-09-15T17:15:51.7987877Z retirado: /opt/az/lib/python3.14/site-packages/pip/_vendor/pygments/lexers/__pycache__/python.cpython-314.pyc
evals	Retirar Python del runner	2026-09-15T17:15:51.8126000Z retirado: /opt/az/lib/python3.14/site-packages/pip/_vendor/pygments/lexers/python.py
evals	Retirar Python del runner	2026-09-15T17:15:51.8258565Z retirado: /opt/az/lib/python3.14/site-packages/argcomplete/scripts/__pycache__/python_argcomplete_check_easy_install_script.cpython-314.pyc
evals	Retirar Python del runner	2026-09-15T17:15:51.8387723Z retirado: /opt/az/lib/python3.14/site-packages/argcomplete/scripts/python_argcomplete_check_easy_install_script.py
evals	Retirar Python del runner	2026-09-15T17:15:51.8521858Z retirado: /opt/az/lib/python3.14/config-3.14-x86_64-linux-gnu/libpython3.14.a
evals	Retirar Python del runner	2026-09-15T17:15:51.8650297Z retirado: /opt/az/lib/python3.14/config-3.14-x86_64-linux-gnu/python-config.py
evals	Retirar Python del runner	2026-09-15T17:15:51.8783329Z retirado: /opt/az/lib/python3.14/config-3.14-x86_64-linux-gnu/python.o
evals	Retirar Python del runner	2026-09-15T17:15:51.8914286Z retirado: /opt/az/lib/python3.14/idlelib/Icons/python.gif
evals	Retirar Python del runner	2026-09-15T17:15:51.9047584Z retirado: /opt/az/lib/libpython3.14.a
evals	Retirar Python del runner	2026-09-15T17:15:51.9179883Z retirado: /opt/az/include/python3.14/cpython/pythonrun.h
evals	Retirar Python del runner	2026-09-15T17:15:51.9310076Z retirado: /opt/az/include/python3.14/Python.h
evals	Retirar Python del runner	2026-09-15T17:15:51.9441612Z retirado: /opt/az/include/python3.14/pythonrun.h
evals	Retirar Python del runner	2026-09-15T17:15:51.9570356Z retirado: /opt/az/share/man/man1/python3.1
evals	Retirar Python del runner	2026-09-15T17:15:51.9703687Z retirado: /opt/az/share/man/man1/python3.14.1
evals	Retirar Python del runner	2026-09-15T17:15:51.9956398Z retirado: /opt/az
evals	Retirar Python del runner	2026-09-15T17:15:52.3791938Z retirado: /var/lib/dpkg/info/python3-jinja2.prerm
evals	Retirar Python del runner	2026-09-15T17:15:52.3925629Z retirado: /var/lib/dpkg/info/python3-packaging.prerm
evals	Retirar Python del runner	2026-09-15T17:15:52.4063078Z retirado: /var/lib/dpkg/info/python3-launchpadlib.postinst
evals	Retirar Python del runner	2026-09-15T17:15:52.4197775Z retirado: /var/lib/dpkg/info/python3-magic.postinst
evals	Retirar Python del runner	2026-09-15T17:15:52.4331716Z retirado: /var/lib/dpkg/info/python3-jsonschema.postrm
evals	Retirar Python del runner	2026-09-15T17:15:52.4464013Z retirado: /var/lib/dpkg/info/python3-jsonpatch.postinst
evals	Retirar Python del runner	2026-09-15T17:15:52.4592664Z retirado: /var/lib/dpkg/info/python3-parted.prerm
evals	Retirar Python del runner	2026-09-15T17:15:52.4728695Z retirado: /var/lib/dpkg/info/python3-chardet.postinst
evals	Retirar Python del runner	2026-09-15T17:15:52.4860044Z retirado: /var/lib/dpkg/info/python3-parted.postinst
evals	Retirar Python del runner	2026-09-15T17:15:52.4989373Z retirado: /var/lib/dpkg/info/python3-constantly.prerm
evals	Retirar Python del runner	2026-09-15T17:15:52.5117520Z retirado: /var/lib/dpkg/info/python3-gi.prerm
evals	Retirar Python del runner	2026-09-15T17:15:52.5253191Z retirado: /var/lib/dpkg/info/python3-s3transfer.prerm
evals	Retirar Python del runner	2026-09-15T17:15:52.5378468Z retirado: /var/lib/dpkg/info/python3-bcrypt.postinst
evals	Retirar Python del runner	2026-09-15T17:15:52.5506722Z retirado: /var/lib/dpkg/info/python3-netaddr.postinst
evals	Retirar Python del runner	2026-09-15T17:15:52.5638181Z retirado: /var/lib/dpkg/info/python3-zope.interface.prerm
evals	Retirar Python del runner	2026-09-15T17:15:52.5766691Z retirado: /var/lib/dpkg/info/python3-cryptography.postinst
evals	Retirar Python del runner	2026-09-15T17:15:52.5894868Z retirado: /var/lib/dpkg/info/python3-distro-info.prerm
evals	Retirar Python del runner	2026-09-15T17:15:52.6030338Z retirado: /var/lib/dpkg/info/python3-configobj.postinst
evals	Retirar Python del runner	2026-09-15T17:15:52.6148913Z retirado: /var/lib/dpkg/info/libpython3.12t64:amd64.md5sums
evals	Retirar Python del runner	2026-09-15T17:15:52.6274346Z retirado: /var/lib/dpkg/info/python3-lazr.restfulclient.prerm
evals	Retirar Python del runner	2026-09-15T17:15:52.6405724Z retirado: /var/lib/dpkg/info/python3-jsonschema.postinst
evals	Retirar Python del runner	2026-09-15T17:15:52.6531068Z retirado: /var/lib/dpkg/info/python3-six.postinst
evals	Retirar Python del runner	2026-09-15T17:15:52.6661332Z retirado: /var/lib/dpkg/info/libpython3.12-minimal:amd64.postrm
evals	Retirar Python del runner	2026-09-15T17:15:52.6791780Z retirado: /var/lib/dpkg/info/python3-idna.prerm
evals	Retirar Python del runner	2026-09-15T17:15:52.6915430Z retirado: /var/lib/dpkg/info/python3-jsonpatch.prerm
evals	Retirar Python del runner	2026-09-15T17:15:52.7041969Z retirado: /var/lib/dpkg/info/python3-cryptography.prerm
evals	Retirar Python del runner	2026-09-15T17:15:52.7167189Z retirado: /var/lib/dpkg/info/python3-babel.postinst
evals	Retirar Python del runner	2026-09-15T17:15:52.7295218Z retirado: /var/lib/dpkg/info/python3-distupgrade.postinst
evals	Retirar Python del runner	2026-09-15T17:15:52.7425565Z retirado: /var/lib/dpkg/info/python3-minimal.prerm
evals	Retirar Python del runner	2026-09-15T17:15:52.7555426Z retirado: /var/lib/dpkg/info/python3-mdurl.prerm
evals	Retirar Python del runner	2026-09-15T17:15:52.7679864Z retirado: /var/lib/dpkg/info/python3-pkg-resources.postinst
evals	Retirar Python del runner	2026-09-15T17:15:52.7806394Z retirado: /var/lib/dpkg/info/python3-launchpadlib.prerm
evals	Retirar Python del runner	2026-09-15T17:15:52.7934228Z retirado: /var/lib/dpkg/info/python3-debian.postinst
evals	Retirar Python del runner	2026-09-15T17:15:52.8065596Z retirado: /var/lib/dpkg/info/python3-wheel.prerm
evals	Retirar Python del runner	2026-09-15T17:15:52.8196210Z retirado: /var/lib/dpkg/info/python3.12-minimal.postrm
evals	Retirar Python del runner	2026-09-15T17:15:52.8327246Z retirado: /var/lib/dpkg/info/python3-certifi.postinst
evals	Retirar Python del runner	2026-09-15T17:15:52.8456563Z retirado: /var/lib/dpkg/info/python3-twisted.postrm
evals	Retirar Python del runner	2026-09-15T17:15:52.8581907Z retirado: /var/lib/dpkg/info/python3-systemd.postinst
evals	Retirar Python del runner	2026-09-15T17:15:52.8709364Z retirado: /var/lib/dpkg/info/python3-botocore.prerm
evals	Retirar Python del runner	2026-09-15T17:15:52.8837504Z retirado: /var/lib/dpkg/info/python3.12-venv.postrm
evals	Retirar Python del runner	2026-09-15T17:15:52.8965866Z retirado: /var/lib/dpkg/info/python3-openssl.postinst
evals	Retirar Python del runner	2026-09-15T17:15:52.9095787Z retirado: /var/lib/dpkg/info/python3-launchpadlib.postrm
evals	Retirar Python del runner	2026-09-15T17:15:52.9223772Z retirado: /var/lib/dpkg/info/python3-json-pointer.postinst
evals	Retirar Python del runner	2026-09-15T17:15:52.9350868Z retirado: /var/lib/dpkg/info/python3-requests.prerm
evals	Retirar Python del runner	2026-09-15T17:15:52.9480264Z retirado: /var/lib/dpkg/info/python3-pyasn1.prerm
evals	Retirar Python del runner	2026-09-15T17:15:52.9608485Z retirado: /var/lib/dpkg/info/python3-openssl.prerm
evals	Retirar Python del runner	2026-09-15T17:15:52.9745462Z retirado: /var/lib/dpkg/info/python3-attr.postinst
evals	Retirar Python del runner	2026-09-15T17:15:52.9873676Z retirado: /var/lib/dpkg/info/python3.preinst
evals	Retirar Python del runner	2026-09-15T17:15:52.9999892Z retirado: /var/lib/dpkg/info/python3-apt.prerm
evals	Retirar Python del runner	2026-09-15T17:15:53.0130902Z retirado: /var/lib/dpkg/info/python3-pyasn1-modules.postinst
evals	Retirar Python del runner	2026-09-15T17:15:53.0262697Z retirado: /var/lib/dpkg/info/libpython3.12-stdlib:amd64.md5sums
evals	Retirar Python del runner	2026-09-15T17:15:53.0392049Z retirado: /var/lib/dpkg/info/python3-newt:amd64.postinst
evals	Retirar Python del runner	2026-09-15T17:15:53.0516773Z retirado: /var/lib/dpkg/info/libpython3-stdlib:amd64.list
evals	Retirar Python del runner	2026-09-15T17:15:53.0643256Z retirado: /var/lib/dpkg/info/python3-commandnotfound.postinst
evals	Retirar Python del runner	2026-09-15T17:15:53.0772791Z retirado: /var/lib/dpkg/info/libpython3.12t64:amd64.symbols
evals	Retirar Python del runner	2026-09-15T17:15:53.0912220Z retirado: /var/lib/dpkg/info/python3-pyrsistent:amd64.prerm
evals	Retirar Python del runner	2026-09-15T17:15:53.1041359Z retirado: /var/lib/dpkg/info/python3-yaml.postinst
evals	Retirar Python del runner	2026-09-15T17:15:53.1171318Z retirado: /var/lib/dpkg/info/python3-debconf.postinst
evals	Retirar Python del runner	2026-09-15T17:15:53.1303497Z retirado: /var/lib/dpkg/info/python3-boto3.postinst
evals	Retirar Python del runner	2026-09-15T17:15:53.1433570Z retirado: /var/lib/dpkg/info/python3-passlib.prerm
evals	Retirar Python del runner	2026-09-15T17:15:53.1568169Z retirado: /var/lib/dpkg/info/python3.12.prerm
evals	Retirar Python del runner	2026-09-15T17:15:53.1697117Z retirado: /var/lib/dpkg/info/python3-idna.postinst
evals	Retirar Python del runner	2026-09-15T17:15:53.1827486Z retirado: /var/lib/dpkg/info/python3-problem-report.prerm
evals	Retirar Python del runner	2026-09-15T17:15:53.1954795Z retirado: /var/lib/dpkg/info/python3.12-venv.prerm
evals	Retirar Python del runner	2026-09-15T17:15:53.2083303Z retirado: /var/lib/dpkg/info/python3-apport.prerm
evals	Retirar Python del runner	2026-09-15T17:15:53.2218669Z retirado: /var/lib/dpkg/info/python3-newt:amd64.prerm
evals	Retirar Python del runner	2026-09-15T17:15:53.2352514Z retirado: /var/lib/dpkg/info/python3-distro-info.postinst
evals	Retirar Python del runner	2026-09-15T17:15:53.2479873Z retirado: /var/lib/dpkg/info/python3.12.postinst
evals	Retirar Python del runner	2026-09-15T17:15:53.2608391Z retirado: /var/lib/dpkg/info/python3-pip.prerm
evals	Retirar Python del runner	2026-09-15T17:15:53.2736971Z retirado: /var/lib/dpkg/info/python3.12-minimal.preinst
evals	Retirar Python del runner	2026-09-15T17:15:53.2864753Z retirado: /var/lib/dpkg/info/python3-urllib3.postinst
evals	Retirar Python del runner	2026-09-15T17:15:53.2994006Z retirado: /var/lib/dpkg/info/python3-bpfcc.prerm
evals	Retirar Python del runner	2026-09-15T17:15:53.3123803Z retirado: /var/lib/dpkg/info/python3-wadllib.postinst
evals	Retirar Python del runner	2026-09-15T17:15:53.3255508Z retirado: /var/lib/dpkg/info/python3-jwt.postinst
evals	Retirar Python del runner	2026-09-15T17:15:53.3387342Z retirado: /var/lib/dpkg/info/python3-distupgrade.prerm
evals	Retirar Python del runner	2026-09-15T17:15:53.3512569Z retirado: /var/lib/dpkg/info/python3-problem-report.postinst
evals	Retirar Python del runner	2026-09-15T17:15:53.3642498Z retirado: /var/lib/dpkg/info/python3-pexpect.postinst
evals	Retirar Python del runner	2026-09-15T17:15:53.3775266Z retirado: /var/lib/dpkg/info/python3-zstandard.postinst
evals	Retirar Python del runner	2026-09-15T17:15:53.3915950Z retirado: /var/lib/dpkg/info/python3-gi.postinst
evals	Retirar Python del runner	2026-09-15T17:15:53.4050797Z retirado: /var/lib/dpkg/info/python3-update-manager.postinst
evals	Retirar Python del runner	2026-09-15T17:15:53.4184740Z retirado: /var/lib/dpkg/info/python3-httplib2.prerm
evals	Retirar Python del runner	2026-09-15T17:15:53.4318641Z retirado: /var/lib/dpkg/info/python3-pyasn1.postinst
evals	Retirar Python del runner	2026-09-15T17:15:53.4450389Z retirado: /var/lib/dpkg/info/python3-pkg-resources.prerm
evals	Retirar Python del runner	2026-09-15T17:15:53.4581989Z retirado: /var/lib/dpkg/info/python3-markupsafe.postinst
evals	Retirar Python del runner	2026-09-15T17:15:53.4713793Z retirado: /var/lib/dpkg/info/python3-boto3.prerm
evals	Retirar Python del runner	2026-09-15T17:15:53.4845120Z retirado: /var/lib/dpkg/info/libpython3-stdlib:amd64.md5sums
evals	Retirar Python del runner	2026-09-15T17:15:53.4972978Z retirado: /var/lib/dpkg/info/libpython3.12-minimal:amd64.md5sums
evals	Retirar Python del runner	2026-09-15T17:15:53.5105062Z retirado: /var/lib/dpkg/info/python3-markdown-it.prerm
evals	Retirar Python del runner	2026-09-15T17:15:53.5242170Z retirado: /var/lib/dpkg/info/python3-distro.prerm
evals	Retirar Python del runner	2026-09-15T17:15:53.5375983Z retirado: /var/lib/dpkg/info/python3-requests.postinst
evals	Retirar Python del runner	2026-09-15T17:15:53.5509905Z retirado: /var/lib/dpkg/info/python3-hyperlink.postinst
evals	Retirar Python del runner	2026-09-15T17:15:53.5639963Z retirado: /var/lib/dpkg/info/libpython3.12t64:amd64.list
evals	Retirar Python del runner	2026-09-15T17:15:53.5772223Z retirado: /var/lib/dpkg/info/python3-hyperlink.prerm
evals	Retirar Python del runner	2026-09-15T17:15:53.5906582Z retirado: /var/lib/dpkg/info/python3-minimal.postinst
evals	Retirar Python del runner	2026-09-15T17:15:53.6039756Z retirado: /var/lib/dpkg/info/python3-jwt.prerm
evals	Retirar Python del runner	2026-09-15T17:15:53.6173052Z retirado: /var/lib/dpkg/info/python3-pyasn1-modules.prerm
evals	Retirar Python del runner	2026-09-15T17:15:53.6309872Z retirado: /var/lib/dpkg/info/python3-lazr.uri.postinst
evals	Retirar Python del runner	2026-09-15T17:15:53.6443388Z retirado: /var/lib/dpkg/info/python3-jsonpatch.postrm
evals	Retirar Python del runner	2026-09-15T17:15:53.6579593Z retirado: /var/lib/dpkg/info/python3-pygments.postinst
evals	Retirar Python del runner	2026-09-15T17:15:53.6712954Z retirado: /var/lib/dpkg/info/python3-json-pointer.postrm
evals	Retirar Python del runner	2026-09-15T17:15:53.6840409Z retirado: /var/lib/dpkg/info/libpython3.12-minimal:amd64.prerm
evals	Retirar Python del runner	2026-09-15T17:15:53.6968640Z retirado: /var/lib/dpkg/info/python3-rich.postinst
evals	Retirar Python del runner	2026-09-15T17:15:53.7097302Z retirado: /var/lib/dpkg/info/python3-jsonschema.prerm
evals	Retirar Python del runner	2026-09-15T17:15:53.7228795Z retirado: /var/lib/dpkg/info/python3-mdurl.postinst
evals	Retirar Python del runner	2026-09-15T17:15:53.7357962Z retirado: /var/lib/dpkg/info/python3-software-properties.postinst
evals	Retirar Python del runner	2026-09-15T17:15:53.7487953Z retirado: /var/lib/dpkg/info/python3-pyparsing.postinst
evals	Retirar Python del runner	2026-09-15T17:15:53.7618829Z retirado: /var/lib/dpkg/info/python3-pip.postinst
evals	Retirar Python del runner	2026-09-15T17:15:53.7751072Z retirado: /var/lib/dpkg/info/python3-distro.postinst
evals	Retirar Python del runner	2026-09-15T17:15:53.7880555Z retirado: /var/lib/dpkg/info/python3-hamcrest.prerm
evals	Retirar Python del runner	2026-09-15T17:15:53.8009408Z retirado: /var/lib/dpkg/info/python3-urllib3.prerm
evals	Retirar Python del runner	2026-09-15T17:15:53.8138809Z retirado: /var/lib/dpkg/info/python3-wadllib.prerm
evals	Retirar Python del runner	2026-09-15T17:15:53.8268109Z retirado: /var/lib/dpkg/info/python3-markupsafe.prerm
evals	Retirar Python del runner	2026-09-15T17:15:53.8400154Z retirado: /var/lib/dpkg/info/python3-httplib2.postinst
evals	Retirar Python del runner	2026-09-15T17:15:53.8530547Z retirado: /var/lib/dpkg/info/python3-certifi.prerm
evals	Retirar Python del runner	2026-09-15T17:15:53.8664824Z retirado: /var/lib/dpkg/info/python3-click.postinst
evals	Retirar Python del runner	2026-09-15T17:15:53.8796623Z retirado: /var/lib/dpkg/info/python3-constantly.postinst
evals	Retirar Python del runner	2026-09-15T17:15:53.8938164Z retirado: /var/lib/dpkg/info/libpython3-dev:amd64.list
evals	Retirar Python del runner	2026-09-15T17:15:53.9067634Z retirado: /var/lib/dpkg/info/python3.12-minimal.postinst
evals	Retirar Python del runner	2026-09-15T17:15:53.9197806Z retirado: /var/lib/dpkg/info/python3-s3transfer.postinst
evals	Retirar Python del runner	2026-09-15T17:15:53.9327135Z retirado: /var/lib/dpkg/info/python3-zstandard.prerm
evals	Retirar Python del runner	2026-09-15T17:15:53.9456145Z retirado: /var/lib/dpkg/info/python3-json-pointer.prerm
evals	Retirar Python del runner	2026-09-15T17:15:53.9586414Z retirado: /var/lib/dpkg/info/python3-service-identity.postinst
evals	Retirar Python del runner	2026-09-15T17:15:53.9716055Z retirado: /var/lib/dpkg/info/python3-serial.postinst
evals	Retirar Python del runner	2026-09-15T17:15:53.9847208Z retirado: /var/lib/dpkg/info/python3-hamcrest.postinst
evals	Retirar Python del runner	2026-09-15T17:15:53.9980742Z retirado: /var/lib/dpkg/info/python3-incremental.postinst
evals	Retirar Python del runner	2026-09-15T17:15:54.0115853Z retirado: /var/lib/dpkg/info/python3-netplan.postinst
evals	Retirar Python del runner	2026-09-15T17:15:54.0247057Z retirado: /var/lib/dpkg/info/python3-netaddr.prerm
evals	Retirar Python del runner	2026-09-15T17:15:54.0377630Z retirado: /var/lib/dpkg/info/python3-dateutil.prerm
evals	Retirar Python del runner	2026-09-15T17:15:54.0510206Z retirado: /var/lib/dpkg/info/python3-apt.postinst
evals	Retirar Python del runner	2026-09-15T17:15:54.0651683Z retirado: /var/lib/dpkg/info/python3-dbus.postinst
evals	Retirar Python del runner	2026-09-15T17:15:54.0794882Z retirado: /var/lib/dpkg/info/python3-jmespath.prerm
evals	Retirar Python del runner	2026-09-15T17:15:54.0928096Z retirado: /var/lib/dpkg/info/libpython3.12-stdlib:amd64.prerm
evals	Retirar Python del runner	2026-09-15T17:15:54.1061411Z retirado: /var/lib/dpkg/info/python3-commandnotfound.prerm
evals	Retirar Python del runner	2026-09-15T17:15:54.1200159Z retirado: /var/lib/dpkg/info/python3-blinker.prerm
evals	Retirar Python del runner	2026-09-15T17:15:54.1331597Z retirado: /var/lib/dpkg/info/python3-ptyprocess.postinst
evals	Retirar Python del runner	2026-09-15T17:15:54.1464889Z retirado: /var/lib/dpkg/info/python3-colorama.postinst
evals	Retirar Python del runner	2026-09-15T17:15:54.1596157Z retirado: /var/lib/dpkg/info/python3-wheel.postinst
evals	Retirar Python del runner	2026-09-15T17:15:54.1720636Z retirado: /var/lib/dpkg/info/python3-oauthlib.postinst
evals	Retirar Python del runner	2026-09-15T17:15:54.1850246Z retirado: /var/lib/dpkg/info/python3-pygments.prerm
evals	Retirar Python del runner	2026-09-15T17:15:54.1979518Z retirado: /var/lib/dpkg/info/python3-tz.postinst
evals	Retirar Python del runner	2026-09-15T17:15:54.2110701Z retirado: /var/lib/dpkg/info/python3.prerm
evals	Retirar Python del runner	2026-09-15T17:15:54.2243622Z retirado: /var/lib/dpkg/info/python3-update-manager.prerm
evals	Retirar Python del runner	2026-09-15T17:15:54.2376473Z retirado: /var/lib/dpkg/info/python3-pexpect.prerm
evals	Retirar Python del runner	2026-09-15T17:15:54.2501479Z retirado: /var/lib/dpkg/info/python3-serial.prerm
evals	Retirar Python del runner	2026-09-15T17:15:54.2630367Z retirado: /var/lib/dpkg/info/python3-netplan.prerm
evals	Retirar Python del runner	2026-09-15T17:15:54.2758029Z retirado: /var/lib/dpkg/info/python3-incremental.prerm
evals	Retirar Python del runner	2026-09-15T17:15:54.2889749Z retirado: /var/lib/dpkg/info/python3-typing-extensions.postinst
evals	Retirar Python del runner	2026-09-15T17:15:54.3020414Z retirado: /var/lib/dpkg/info/python3-jinja2.postinst
evals	Retirar Python del runner	2026-09-15T17:15:54.3153631Z retirado: /var/lib/dpkg/info/libpython3.12-stdlib:amd64.list
evals	Retirar Python del runner	2026-09-15T17:15:54.3283878Z retirado: /var/lib/dpkg/info/python3-pyparsing.prerm
evals	Retirar Python del runner	2026-09-15T17:15:54.3413730Z retirado: /var/lib/dpkg/info/python3-automat.postinst
evals	Retirar Python del runner	2026-09-15T17:15:54.3541834Z retirado: /var/lib/dpkg/info/python3-attr.prerm
evals	Retirar Python del runner	2026-09-15T17:15:54.3670189Z retirado: /var/lib/dpkg/info/python3-pyrsistent:amd64.postinst
evals	Retirar Python del runner	2026-09-15T17:15:54.3800044Z retirado: /var/lib/dpkg/info/python3-passlib.postinst
evals	Retirar Python del runner	2026-09-15T17:15:54.3934581Z retirado: /var/lib/dpkg/info/python3-twisted.postinst
evals	Retirar Python del runner	2026-09-15T17:15:54.4063656Z retirado: /var/lib/dpkg/info/python3-configobj.prerm
evals	Retirar Python del runner	2026-09-15T17:15:54.4190768Z retirado: /var/lib/dpkg/info/python3-markdown-it.postinst
evals	Retirar Python del runner	2026-09-15T17:15:54.4323962Z retirado: /var/lib/dpkg/info/python3-ptyprocess.prerm
evals	Retirar Python del runner	2026-09-15T17:15:54.4455614Z retirado: /var/lib/dpkg/info/libpython3-dev:amd64.md5sums
evals	Retirar Python del runner	2026-09-15T17:15:54.4586070Z retirado: /var/lib/dpkg/info/python3-software-properties.prerm
evals	Retirar Python del runner	2026-09-15T17:15:54.4719609Z retirado: /var/lib/dpkg/info/python3-dbus.prerm
evals	Retirar Python del runner	2026-09-15T17:15:54.4853231Z retirado: /var/lib/dpkg/info/libpython3.12-minimal:amd64.postinst
evals	Retirar Python del runner	2026-09-15T17:15:54.4986873Z retirado: /var/lib/dpkg/info/python3-setuptools.postinst
evals	Retirar Python del runner	2026-09-15T17:15:54.5117178Z retirado: /var/lib/dpkg/info/python3.12-minimal.prerm
evals	Retirar Python del runner	2026-09-15T17:15:54.5246303Z retirado: /var/lib/dpkg/info/python3-botocore.postinst
evals	Retirar Python del runner	2026-09-15T17:15:54.5377721Z retirado: /var/lib/dpkg/info/python3-setuptools.prerm
evals	Retirar Python del runner	2026-09-15T17:15:54.5507994Z retirado: /var/lib/dpkg/info/python3-dateutil.postinst
evals	Retirar Python del runner	2026-09-15T17:15:54.5638216Z retirado: /var/lib/dpkg/info/python3.postrm
evals	Retirar Python del runner	2026-09-15T17:15:54.5770727Z retirado: /var/lib/dpkg/info/python3-yaml.prerm
evals	Retirar Python del runner	2026-09-15T17:15:54.5899648Z retirado: /var/lib/dpkg/info/libpython3.12-dev:amd64.list
evals	Retirar Python del runner	2026-09-15T17:15:54.6029810Z retirado: /var/lib/dpkg/info/python3-lazr.restfulclient.postinst
evals	Retirar Python del runner	2026-09-15T17:15:54.6159934Z retirado: /var/lib/dpkg/info/python3-click.prerm
evals	Retirar Python del runner	2026-09-15T17:15:54.6290121Z retirado: /var/lib/dpkg/info/python3-tz.prerm
evals	Retirar Python del runner	2026-09-15T17:15:54.6424061Z retirado: /var/lib/dpkg/info/python3-debconf.prerm
evals	Retirar Python del runner	2026-09-15T17:15:54.6552908Z retirado: /var/lib/dpkg/info/libpython3.12-dev:amd64.md5sums
evals	Retirar Python del runner	2026-09-15T17:15:54.6684035Z retirado: /var/lib/dpkg/info/python3-automat.prerm
evals	Retirar Python del runner	2026-09-15T17:15:54.6816122Z retirado: /var/lib/dpkg/info/python3-systemd.prerm
evals	Retirar Python del runner	2026-09-15T17:15:54.6951455Z retirado: /var/lib/dpkg/info/python3-typing-extensions.prerm
evals	Retirar Python del runner	2026-09-15T17:15:54.7085578Z retirado: /var/lib/dpkg/info/python3-chardet.prerm
evals	Retirar Python del runner	2026-09-15T17:15:54.7215077Z retirado: /var/lib/dpkg/info/python3-packaging.postinst
evals	Retirar Python del runner	2026-09-15T17:15:54.7343010Z retirado: /var/lib/dpkg/info/libpython3.12t64:amd64.triggers
evals	Retirar Python del runner	2026-09-15T17:15:54.7473251Z retirado: /var/lib/dpkg/info/python3-blinker.postinst
evals	Retirar Python del runner	2026-09-15T17:15:54.7601626Z retirado: /var/lib/dpkg/info/python3.12-venv.postinst
evals	Retirar Python del runner	2026-09-15T17:15:54.7732333Z retirado: /var/lib/dpkg/info/python3-lazr.uri.prerm
evals	Retirar Python del runner	2026-09-15T17:15:54.7863331Z retirado: /var/lib/dpkg/info/python3-six.prerm
evals	Retirar Python del runner	2026-09-15T17:15:54.7991810Z retirado: /var/lib/dpkg/info/python3-twisted.prerm
evals	Retirar Python del runner	2026-09-15T17:15:54.8129922Z retirado: /var/lib/dpkg/info/python3-bcrypt.prerm
evals	Retirar Python del runner	2026-09-15T17:15:54.8263738Z retirado: /var/lib/dpkg/info/python3-magic.prerm
evals	Retirar Python del runner	2026-09-15T17:15:54.8396322Z retirado: /var/lib/dpkg/info/libpython3.12-minimal:amd64.list
evals	Retirar Python del runner	2026-09-15T17:15:54.8529583Z retirado: /var/lib/dpkg/info/python3-service-identity.prerm
evals	Retirar Python del runner	2026-09-15T17:15:54.8664729Z retirado: /var/lib/dpkg/info/python3-colorama.prerm
evals	Retirar Python del runner	2026-09-15T17:15:54.8794925Z retirado: /var/lib/dpkg/info/python3-jmespath.postinst
evals	Retirar Python del runner	2026-09-15T17:15:54.8929387Z retirado: /var/lib/dpkg/info/libpython3.12t64:amd64.shlibs
evals	Retirar Python del runner	2026-09-15T17:15:54.9062217Z retirado: /var/lib/dpkg/info/python3-oauthlib.prerm
evals	Retirar Python del runner	2026-09-15T17:15:54.9190397Z retirado: /var/lib/dpkg/info/python3-rich.prerm
evals	Retirar Python del runner	2026-09-15T17:15:54.9320953Z retirado: /var/lib/dpkg/info/python3-babel.prerm
evals	Retirar Python del runner	2026-09-15T17:15:54.9449602Z retirado: /var/lib/dpkg/info/python3-apport.postinst
evals	Retirar Python del runner	2026-09-15T17:15:54.9577306Z retirado: /var/lib/dpkg/info/python3-bpfcc.postinst
evals	Retirar Python del runner	2026-09-15T17:15:54.9704136Z retirado: /var/lib/dpkg/info/python3-zope.interface.postinst
evals	Retirar Python del runner	2026-09-15T17:15:54.9832610Z retirado: /var/lib/dpkg/info/libpython3.12-minimal:amd64.conffiles
evals	Retirar Python del runner	2026-09-15T17:15:54.9961945Z retirado: /var/lib/dpkg/info/python3.postinst
evals	Retirar Python del runner	2026-09-15T17:15:55.0092017Z retirado: /var/lib/dpkg/info/python3-debian.prerm
evals	Retirar Python del runner	2026-09-15T17:15:55.0226642Z retirado: /var/lib/docker/overlay2/a9853d7ac0752db045c472a9505e498ad787cec8cb7eab1bd6d6d65f685ef414/diff/var/lib/dpkg/info/python3-launchpadlib.postinst
evals	Retirar Python del runner	2026-09-15T17:15:55.0357762Z retirado: /var/lib/docker/overlay2/a9853d7ac0752db045c472a9505e498ad787cec8cb7eab1bd6d6d65f685ef414/diff/var/lib/dpkg/info/python3-cryptography.postinst
evals	Retirar Python del runner	2026-09-15T17:15:55.0490467Z retirado: /var/lib/docker/overlay2/a9853d7ac0752db045c472a9505e498ad787cec8cb7eab1bd6d6d65f685ef414/diff/var/lib/dpkg/info/python3-configobj.postinst
evals	Retirar Python del runner	2026-09-15T17:15:55.0620562Z retirado: /var/lib/docker/overlay2/a9853d7ac0752db045c472a9505e498ad787cec8cb7eab1bd6d6d65f685ef414/diff/var/lib/dpkg/info/libpython3.12t64:amd64.md5sums
evals	Retirar Python del runner	2026-09-15T17:15:55.0751583Z retirado: /var/lib/docker/overlay2/a9853d7ac0752db045c472a9505e498ad787cec8cb7eab1bd6d6d65f685ef414/diff/var/lib/dpkg/info/python3-lazr.restfulclient.prerm
evals	Retirar Python del runner	2026-09-15T17:15:55.0883649Z retirado: /var/lib/docker/overlay2/a9853d7ac0752db045c472a9505e498ad787cec8cb7eab1bd6d6d65f685ef414/diff/var/lib/dpkg/info/python3-six.postinst
evals	Retirar Python del runner	2026-09-15T17:15:55.1015596Z retirado: /var/lib/docker/overlay2/a9853d7ac0752db045c472a9505e498ad787cec8cb7eab1bd6d6d65f685ef414/diff/var/lib/dpkg/info/libpython3.12-minimal:amd64.postrm
evals	Retirar Python del runner	2026-09-15T17:15:55.1143963Z retirado: /var/lib/docker/overlay2/a9853d7ac0752db045c472a9505e498ad787cec8cb7eab1bd6d6d65f685ef414/diff/var/lib/dpkg/info/python3-cryptography.prerm
evals	Retirar Python del runner	2026-09-15T17:15:55.1273436Z retirado: /var/lib/docker/overlay2/a9853d7ac0752db045c472a9505e498ad787cec8cb7eab1bd6d6d65f685ef414/diff/var/lib/dpkg/info/python3-minimal.prerm
evals	Retirar Python del runner	2026-09-15T17:15:55.1404593Z retirado: /var/lib/docker/overlay2/a9853d7ac0752db045c472a9505e498ad787cec8cb7eab1bd6d6d65f685ef414/diff/var/lib/dpkg/info/python3-pkg-resources.postinst
evals	Retirar Python del runner	2026-09-15T17:15:55.1531671Z retirado: /var/lib/docker/overlay2/a9853d7ac0752db045c472a9505e498ad787cec8cb7eab1bd6d6d65f685ef414/diff/var/lib/dpkg/info/python3-launchpadlib.prerm
evals	Retirar Python del runner	2026-09-15T17:15:55.1662327Z retirado: /var/lib/docker/overlay2/a9853d7ac0752db045c472a9505e498ad787cec8cb7eab1bd6d6d65f685ef414/diff/var/lib/dpkg/info/python3.12-minimal.postrm
evals	Retirar Python del runner	2026-09-15T17:15:55.1790664Z retirado: /var/lib/docker/overlay2/a9853d7ac0752db045c472a9505e498ad787cec8cb7eab1bd6d6d65f685ef414/diff/var/lib/dpkg/info/python3-launchpadlib.postrm
evals	Retirar Python del runner	2026-09-15T17:15:55.1918646Z retirado: /var/lib/docker/overlay2/a9853d7ac0752db045c472a9505e498ad787cec8cb7eab1bd6d6d65f685ef414/diff/var/lib/dpkg/info/python3.preinst
evals	Retirar Python del runner	2026-09-15T17:15:55.2050179Z retirado: /var/lib/docker/overlay2/a9853d7ac0752db045c472a9505e498ad787cec8cb7eab1bd6d6d65f685ef414/diff/var/lib/dpkg/info/python3-tzlocal.postinst
evals	Retirar Python del runner	2026-09-15T17:15:55.2183269Z retirado: /var/lib/docker/overlay2/a9853d7ac0752db045c472a9505e498ad787cec8cb7eab1bd6d6d65f685ef414/diff/var/lib/dpkg/info/libpython3.12-stdlib:amd64.md5sums
evals	Retirar Python del runner	2026-09-15T17:15:55.2315136Z retirado: /var/lib/docker/overlay2/a9853d7ac0752db045c472a9505e498ad787cec8cb7eab1bd6d6d65f685ef414/diff/var/lib/dpkg/info/libpython3-stdlib:amd64.list
evals	Retirar Python del runner	2026-09-15T17:15:55.2448635Z retirado: /var/lib/docker/overlay2/a9853d7ac0752db045c472a9505e498ad787cec8cb7eab1bd6d6d65f685ef414/diff/var/lib/dpkg/info/libpython3.12t64:amd64.symbols
evals	Retirar Python del runner	2026-09-15T17:15:55.2583296Z retirado: /var/lib/docker/overlay2/a9853d7ac0752db045c472a9505e498ad787cec8cb7eab1bd6d6d65f685ef414/diff/var/lib/dpkg/info/python3-yaml.postinst
evals	Retirar Python del runner	2026-09-15T17:15:55.2716507Z retirado: /var/lib/docker/overlay2/a9853d7ac0752db045c472a9505e498ad787cec8cb7eab1bd6d6d65f685ef414/diff/var/lib/dpkg/info/python3.12.prerm
evals	Retirar Python del runner	2026-09-15T17:15:55.2849332Z retirado: /var/lib/docker/overlay2/a9853d7ac0752db045c472a9505e498ad787cec8cb7eab1bd6d6d65f685ef414/diff/var/lib/dpkg/info/python3.12.postinst
evals	Retirar Python del runner	2026-09-15T17:15:55.2982370Z retirado: /var/lib/docker/overlay2/a9853d7ac0752db045c472a9505e498ad787cec8cb7eab1bd6d6d65f685ef414/diff/var/lib/dpkg/info/python3.12-minimal.preinst
evals	Retirar Python del runner	2026-09-15T17:15:55.3116871Z retirado: /var/lib/docker/overlay2/a9853d7ac0752db045c472a9505e498ad787cec8cb7eab1bd6d6d65f685ef414/diff/var/lib/dpkg/info/python3-merge3.postinst
evals	Retirar Python del runner	2026-09-15T17:15:55.3250917Z retirado: /var/lib/docker/overlay2/a9853d7ac0752db045c472a9505e498ad787cec8cb7eab1bd6d6d65f685ef414/diff/var/lib/dpkg/info/python3-urllib3.postinst
evals	Retirar Python del runner	2026-09-15T17:15:55.3385791Z retirado: /var/lib/docker/overlay2/a9853d7ac0752db045c472a9505e498ad787cec8cb7eab1bd6d6d65f685ef414/diff/var/lib/dpkg/info/python3-wadllib.postinst
evals	Retirar Python del runner	2026-09-15T17:15:55.3524157Z retirado: /var/lib/docker/overlay2/a9853d7ac0752db045c472a9505e498ad787cec8cb7eab1bd6d6d65f685ef414/diff/var/lib/dpkg/info/python3-jwt.postinst
evals	Retirar Python del runner	2026-09-15T17:15:55.3661188Z retirado: /var/lib/docker/overlay2/a9853d7ac0752db045c472a9505e498ad787cec8cb7eab1bd6d6d65f685ef414/diff/var/lib/dpkg/info/python3-fastbencode.prerm
evals	Retirar Python del runner	2026-09-15T17:15:55.3798369Z retirado: /var/lib/docker/overlay2/a9853d7ac0752db045c472a9505e498ad787cec8cb7eab1bd6d6d65f685ef414/diff/var/lib/dpkg/info/python3-httplib2.prerm
evals	Retirar Python del runner	2026-09-15T17:15:55.3934398Z retirado: /var/lib/docker/overlay2/a9853d7ac0752db045c472a9505e498ad787cec8cb7eab1bd6d6d65f685ef414/diff/var/lib/dpkg/info/python3-pkg-resources.prerm
evals	Retirar Python del runner	2026-09-15T17:15:55.4068750Z retirado: /var/lib/docker/overlay2/a9853d7ac0752db045c472a9505e498ad787cec8cb7eab1bd6d6d65f685ef414/diff/var/lib/dpkg/info/libpython3-stdlib:amd64.md5sums
evals	Retirar Python del runner	2026-09-15T17:15:55.4204964Z retirado: /var/lib/docker/overlay2/a9853d7ac0752db045c472a9505e498ad787cec8cb7eab1bd6d6d65f685ef414/diff/var/lib/dpkg/info/libpython3.12-minimal:amd64.md5sums
evals	Retirar Python del runner	2026-09-15T17:15:55.4338997Z retirado: /var/lib/docker/overlay2/a9853d7ac0752db045c472a9505e498ad787cec8cb7eab1bd6d6d65f685ef414/diff/var/lib/dpkg/info/python3-distro.prerm
evals	Retirar Python del runner	2026-09-15T17:15:55.4473603Z retirado: /var/lib/docker/overlay2/a9853d7ac0752db045c472a9505e498ad787cec8cb7eab1bd6d6d65f685ef414/diff/var/lib/dpkg/info/libpython3.12t64:amd64.list
evals	Retirar Python del runner	2026-09-15T17:15:55.4604439Z retirado: /var/lib/docker/overlay2/a9853d7ac0752db045c472a9505e498ad787cec8cb7eab1bd6d6d65f685ef414/diff/var/lib/dpkg/info/python3-minimal.postinst
evals	Retirar Python del runner	2026-09-15T17:15:55.4735480Z retirado: /var/lib/docker/overlay2/a9853d7ac0752db045c472a9505e498ad787cec8cb7eab1bd6d6d65f685ef414/diff/var/lib/dpkg/info/python3-jwt.prerm
evals	Retirar Python del runner	2026-09-15T17:15:55.4866604Z retirado: /var/lib/docker/overlay2/a9853d7ac0752db045c472a9505e498ad787cec8cb7eab1bd6d6d65f685ef414/diff/var/lib/dpkg/info/python3-lazr.uri.postinst
evals	Retirar Python del runner	2026-09-15T17:15:55.4999362Z retirado: /var/lib/docker/overlay2/a9853d7ac0752db045c472a9505e498ad787cec8cb7eab1bd6d6d65f685ef414/diff/var/lib/dpkg/info/python3-tzlocal.prerm
evals	Retirar Python del runner	2026-09-15T17:15:55.5128990Z retirado: /var/lib/docker/overlay2/a9853d7ac0752db045c472a9505e498ad787cec8cb7eab1bd6d6d65f685ef414/diff/var/lib/dpkg/info/libpython3.12-minimal:amd64.prerm
evals	Retirar Python del runner	2026-09-15T17:15:55.5263699Z retirado: /var/lib/docker/overlay2/a9853d7ac0752db045c472a9505e498ad787cec8cb7eab1bd6d6d65f685ef414/diff/var/lib/dpkg/info/python3-pyparsing.postinst
evals	Retirar Python del runner	2026-09-15T17:15:55.5396162Z retirado: /var/lib/docker/overlay2/a9853d7ac0752db045c472a9505e498ad787cec8cb7eab1bd6d6d65f685ef414/diff/var/lib/dpkg/info/python3-distro.postinst
evals	Retirar Python del runner	2026-09-15T17:15:55.5525046Z retirado: /var/lib/docker/overlay2/a9853d7ac0752db045c472a9505e498ad787cec8cb7eab1bd6d6d65f685ef414/diff/var/lib/dpkg/info/python3-urllib3.prerm
evals	Retirar Python del runner	2026-09-15T17:15:55.5654780Z retirado: /var/lib/docker/overlay2/a9853d7ac0752db045c472a9505e498ad787cec8cb7eab1bd6d6d65f685ef414/diff/var/lib/dpkg/info/python3-wadllib.prerm
evals	Retirar Python del runner	2026-09-15T17:15:55.5783081Z retirado: /var/lib/docker/overlay2/a9853d7ac0752db045c472a9505e498ad787cec8cb7eab1bd6d6d65f685ef414/diff/var/lib/dpkg/info/python3-httplib2.postinst
evals	Retirar Python del runner	2026-09-15T17:15:55.5913692Z retirado: /var/lib/docker/overlay2/a9853d7ac0752db045c472a9505e498ad787cec8cb7eab1bd6d6d65f685ef414/diff/var/lib/dpkg/info/python3.12-minimal.postinst
evals	Retirar Python del runner	2026-09-15T17:15:55.6045847Z retirado: /var/lib/docker/overlay2/a9853d7ac0752db045c472a9505e498ad787cec8cb7eab1bd6d6d65f685ef414/diff/var/lib/dpkg/info/python3-patiencediff.postinst
evals	Retirar Python del runner	2026-09-15T17:15:55.6178111Z retirado: /var/lib/docker/overlay2/a9853d7ac0752db045c472a9505e498ad787cec8cb7eab1bd6d6d65f685ef414/diff/var/lib/dpkg/info/libpython3.12-stdlib:amd64.prerm
evals	Retirar Python del runner	2026-09-15T17:15:55.6310249Z retirado: /var/lib/docker/overlay2/a9853d7ac0752db045c472a9505e498ad787cec8cb7eab1bd6d6d65f685ef414/diff/var/lib/dpkg/info/python3-blinker.prerm
evals	Retirar Python del runner	2026-09-15T17:15:55.6440917Z retirado: /var/lib/docker/overlay2/a9853d7ac0752db045c472a9505e498ad787cec8cb7eab1bd6d6d65f685ef414/diff/var/lib/dpkg/info/python3-oauthlib.postinst
evals	Retirar Python del runner	2026-09-15T17:15:55.6573501Z retirado: /var/lib/docker/overlay2/a9853d7ac0752db045c472a9505e498ad787cec8cb7eab1bd6d6d65f685ef414/diff/var/lib/dpkg/info/python3.prerm
evals	Retirar Python del runner	2026-09-15T17:15:55.6705242Z retirado: /var/lib/docker/overlay2/a9853d7ac0752db045c472a9505e498ad787cec8cb7eab1bd6d6d65f685ef414/diff/var/lib/dpkg/info/libpython3.12-stdlib:amd64.list
evals	Retirar Python del runner	2026-09-15T17:15:55.6838154Z retirado: /var/lib/docker/overlay2/a9853d7ac0752db045c472a9505e498ad787cec8cb7eab1bd6d6d65f685ef414/diff/var/lib/dpkg/info/python3-pyparsing.prerm
evals	Retirar Python del runner	2026-09-15T17:15:55.6969973Z retirado: /var/lib/docker/overlay2/a9853d7ac0752db045c472a9505e498ad787cec8cb7eab1bd6d6d65f685ef414/diff/var/lib/dpkg/info/python3-breezy.prerm
evals	Retirar Python del runner	2026-09-15T17:15:55.7100733Z retirado: /var/lib/docker/overlay2/a9853d7ac0752db045c472a9505e498ad787cec8cb7eab1bd6d6d65f685ef414/diff/var/lib/dpkg/info/python3-dulwich.postinst
evals	Retirar Python del runner	2026-09-15T17:15:55.7231687Z retirado: /var/lib/docker/overlay2/a9853d7ac0752db045c472a9505e498ad787cec8cb7eab1bd6d6d65f685ef414/diff/var/lib/dpkg/info/python3-configobj.prerm
evals	Retirar Python del runner	2026-09-15T17:15:55.7369522Z retirado: /var/lib/docker/overlay2/a9853d7ac0752db045c472a9505e498ad787cec8cb7eab1bd6d6d65f685ef414/diff/var/lib/dpkg/info/libpython3.12-minimal:amd64.postinst
evals	Retirar Python del runner	2026-09-15T17:15:55.7501050Z retirado: /var/lib/docker/overlay2/a9853d7ac0752db045c472a9505e498ad787cec8cb7eab1bd6d6d65f685ef414/diff/var/lib/dpkg/info/python3.12-minimal.prerm
evals	Retirar Python del runner	2026-09-15T17:15:55.7635127Z retirado: /var/lib/docker/overlay2/a9853d7ac0752db045c472a9505e498ad787cec8cb7eab1bd6d6d65f685ef414/diff/var/lib/dpkg/info/python3-dulwich.prerm
evals	Retirar Python del runner	2026-09-15T17:15:55.7766032Z retirado: /var/lib/docker/overlay2/a9853d7ac0752db045c472a9505e498ad787cec8cb7eab1bd6d6d65f685ef414/diff/var/lib/dpkg/info/python3.postrm
evals	Retirar Python del runner	2026-09-15T17:15:55.7898783Z retirado: /var/lib/docker/overlay2/a9853d7ac0752db045c472a9505e498ad787cec8cb7eab1bd6d6d65f685ef414/diff/var/lib/dpkg/info/python3-yaml.prerm
evals	Retirar Python del runner	2026-09-15T17:15:55.8031332Z retirado: /var/lib/docker/overlay2/a9853d7ac0752db045c472a9505e498ad787cec8cb7eab1bd6d6d65f685ef414/diff/var/lib/dpkg/info/python3-lazr.restfulclient.postinst
evals	Retirar Python del runner	2026-09-15T17:15:55.8166633Z retirado: /var/lib/docker/overlay2/a9853d7ac0752db045c472a9505e498ad787cec8cb7eab1bd6d6d65f685ef414/diff/var/lib/dpkg/info/python3-patiencediff.prerm
evals	Retirar Python del runner	2026-09-15T17:15:55.8305092Z retirado: /var/lib/docker/overlay2/a9853d7ac0752db045c472a9505e498ad787cec8cb7eab1bd6d6d65f685ef414/diff/var/lib/dpkg/info/python3-merge3.prerm
evals	Retirar Python del runner	2026-09-15T17:15:55.8443533Z retirado: /var/lib/docker/overlay2/a9853d7ac0752db045c472a9505e498ad787cec8cb7eab1bd6d6d65f685ef414/diff/var/lib/dpkg/info/libpython3.12t64:amd64.triggers
evals	Retirar Python del runner	2026-09-15T17:15:55.8577912Z retirado: /var/lib/docker/overlay2/a9853d7ac0752db045c472a9505e498ad787cec8cb7eab1bd6d6d65f685ef414/diff/var/lib/dpkg/info/python3-blinker.postinst
evals	Retirar Python del runner	2026-09-15T17:15:55.8711045Z retirado: /var/lib/docker/overlay2/a9853d7ac0752db045c472a9505e498ad787cec8cb7eab1bd6d6d65f685ef414/diff/var/lib/dpkg/info/python3-lazr.uri.prerm
evals	Retirar Python del runner	2026-09-15T17:15:55.8846482Z retirado: /var/lib/docker/overlay2/a9853d7ac0752db045c472a9505e498ad787cec8cb7eab1bd6d6d65f685ef414/diff/var/lib/dpkg/info/python3-six.prerm
evals	Retirar Python del runner	2026-09-15T17:15:55.8990499Z retirado: /var/lib/docker/overlay2/a9853d7ac0752db045c472a9505e498ad787cec8cb7eab1bd6d6d65f685ef414/diff/var/lib/dpkg/info/python3-fastbencode.postinst
evals	Retirar Python del runner	2026-09-15T17:15:55.9126769Z retirado: /var/lib/docker/overlay2/a9853d7ac0752db045c472a9505e498ad787cec8cb7eab1bd6d6d65f685ef414/diff/var/lib/dpkg/info/libpython3.12-minimal:amd64.list
evals	Retirar Python del runner	2026-09-15T17:15:55.9268983Z retirado: /var/lib/docker/overlay2/a9853d7ac0752db045c472a9505e498ad787cec8cb7eab1bd6d6d65f685ef414/diff/var/lib/dpkg/info/libpython3.12t64:amd64.shlibs
evals	Retirar Python del runner	2026-09-15T17:15:55.9403704Z retirado: /var/lib/docker/overlay2/a9853d7ac0752db045c472a9505e498ad787cec8cb7eab1bd6d6d65f685ef414/diff/var/lib/dpkg/info/python3-breezy.postinst
evals	Retirar Python del runner	2026-09-15T17:15:55.9537451Z retirado: /var/lib/docker/overlay2/a9853d7ac0752db045c472a9505e498ad787cec8cb7eab1bd6d6d65f685ef414/diff/var/lib/dpkg/info/python3-oauthlib.prerm
evals	Retirar Python del runner	2026-09-15T17:15:55.9670366Z retirado: /var/lib/docker/overlay2/a9853d7ac0752db045c472a9505e498ad787cec8cb7eab1bd6d6d65f685ef414/diff/var/lib/dpkg/info/libpython3.12-minimal:amd64.conffiles
evals	Retirar Python del runner	2026-09-15T17:15:55.9803382Z retirado: /var/lib/docker/overlay2/a9853d7ac0752db045c472a9505e498ad787cec8cb7eab1bd6d6d65f685ef414/diff/var/lib/dpkg/info/python3.postinst
evals	Retirar Python del runner	2026-09-15T17:15:55.9934692Z retirado: /var/lib/docker/overlay2/a9853d7ac0752db045c472a9505e498ad787cec8cb7eab1bd6d6d65f685ef414/diff/usr/lib/x86_64-linux-gnu/libpython3.12.so.1
evals	Retirar Python del runner	2026-09-15T17:15:56.0069652Z retirado: /var/lib/docker/overlay2/a9853d7ac0752db045c472a9505e498ad787cec8cb7eab1bd6d6d65f685ef414/diff/usr/lib/x86_64-linux-gnu/libpython3.12.so.1.0
evals	Retirar Python del runner	2026-09-15T17:15:56.0258225Z retirado: /var/lib/docker/overlay2/a9853d7ac0752db045c472a9505e498ad787cec8cb7eab1bd6d6d65f685ef414/diff/usr/lib/python3.12/config-3.12-x86_64-linux-gnu/libpython3.12.so
evals	Retirar Python del runner	2026-09-15T17:15:56.0395423Z retirado: /var/lib/docker/overlay2/a9853d7ac0752db045c472a9505e498ad787cec8cb7eab1bd6d6d65f685ef414/diff/usr/share/pixmaps/python3.xpm
evals	Retirar Python del runner	2026-09-15T17:15:56.0527780Z retirado: /var/lib/docker/overlay2/a9853d7ac0752db045c472a9505e498ad787cec8cb7eab1bd6d6d65f685ef414/diff/usr/share/doc/libpython3.12t64
evals	Retirar Python del runner	2026-09-15T17:15:56.0714739Z retirado: /var/lib/docker/overlay2/a9853d7ac0752db045c472a9505e498ad787cec8cb7eab1bd6d6d65f685ef414/diff/usr/share/doc/python3.12/python-policy.txt.gz
evals	Retirar Python del runner	2026-09-15T17:15:56.0848586Z retirado: /var/lib/docker/overlay2/a9853d7ac0752db045c472a9505e498ad787cec8cb7eab1bd6d6d65f685ef414/diff/usr/share/doc/libpython3.12-stdlib
evals	Retirar Python del runner	2026-09-15T17:15:56.0981004Z retirado: /var/lib/docker/overlay2/a9853d7ac0752db045c472a9505e498ad787cec8cb7eab1bd6d6d65f685ef414/diff/usr/share/lintian/overrides/libpython3.12t64
evals	Retirar Python del runner	2026-09-15T17:15:56.1116499Z retirado: /var/lib/docker/overlay2/a9853d7ac0752db045c472a9505e498ad787cec8cb7eab1bd6d6d65f685ef414/diff/usr/share/lintian/overrides/libpython3.12-minimal
evals	Retirar Python del runner	2026-09-15T17:15:56.1248417Z retirado: /var/lib/docker/overlay2/a9853d7ac0752db045c472a9505e498ad787cec8cb7eab1bd6d6d65f685ef414/diff/usr/share/lintian/overrides/libpython3.12-stdlib
evals	Retirar Python del runner	2026-09-15T17:15:56.1508463Z retirado: /var/lib/docker/overlay2/a9853d7ac0752db045c472a9505e498ad787cec8cb7eab1bd6d6d65f685ef414/diff/usr
evals	Retirar Python del runner	2026-09-15T17:15:56.8284787Z retirado: /var/lib/docker/overlay2/80c613159e0ff38a882acb63fd4cf342723381c4f07d119df138a79409f5e793/diff/var/lib/dpkg/info/python3.10-minimal.prerm
evals	Retirar Python del runner	2026-09-15T17:15:56.8421851Z retirado: /var/lib/docker/overlay2/80c613159e0ff38a882acb63fd4cf342723381c4f07d119df138a79409f5e793/diff/var/lib/dpkg/info/libpython3.10-minimal:amd64.postinst
evals	Retirar Python del runner	2026-09-15T17:15:56.8553103Z retirado: /var/lib/docker/overlay2/80c613159e0ff38a882acb63fd4cf342723381c4f07d119df138a79409f5e793/diff/var/lib/dpkg/info/python3-minimal.prerm
evals	Retirar Python del runner	2026-09-15T17:15:56.8683009Z retirado: /var/lib/docker/overlay2/80c613159e0ff38a882acb63fd4cf342723381c4f07d119df138a79409f5e793/diff/var/lib/dpkg/info/libpython3.10-minimal:amd64.postrm
evals	Retirar Python del runner	2026-09-15T17:15:56.8814985Z retirado: /var/lib/docker/overlay2/80c613159e0ff38a882acb63fd4cf342723381c4f07d119df138a79409f5e793/diff/var/lib/dpkg/info/python3.preinst
evals	Retirar Python del runner	2026-09-15T17:15:56.8945296Z retirado: /var/lib/docker/overlay2/80c613159e0ff38a882acb63fd4cf342723381c4f07d119df138a79409f5e793/diff/var/lib/dpkg/info/libpython3-stdlib:amd64.list
evals	Retirar Python del runner	2026-09-15T17:15:56.9074279Z retirado: /var/lib/docker/overlay2/80c613159e0ff38a882acb63fd4cf342723381c4f07d119df138a79409f5e793/diff/var/lib/dpkg/info/libpython3.10-minimal:amd64.prerm
evals	Retirar Python del runner	2026-09-15T17:15:56.9204660Z retirado: /var/lib/docker/overlay2/80c613159e0ff38a882acb63fd4cf342723381c4f07d119df138a79409f5e793/diff/var/lib/dpkg/info/python3.10-minimal.postrm
evals	Retirar Python del runner	2026-09-15T17:15:56.9335532Z retirado: /var/lib/docker/overlay2/80c613159e0ff38a882acb63fd4cf342723381c4f07d119df138a79409f5e793/diff/var/lib/dpkg/info/libpython3.10-minimal:amd64.list
evals	Retirar Python del runner	2026-09-15T17:15:56.9466484Z retirado: /var/lib/docker/overlay2/80c613159e0ff38a882acb63fd4cf342723381c4f07d119df138a79409f5e793/diff/var/lib/dpkg/info/python3.10.prerm
evals	Retirar Python del runner	2026-09-15T17:15:56.9601111Z retirado: /var/lib/docker/overlay2/80c613159e0ff38a882acb63fd4cf342723381c4f07d119df138a79409f5e793/diff/var/lib/dpkg/info/libpython3-stdlib:amd64.md5sums
evals	Retirar Python del runner	2026-09-15T17:15:56.9733942Z retirado: /var/lib/docker/overlay2/80c613159e0ff38a882acb63fd4cf342723381c4f07d119df138a79409f5e793/diff/var/lib/dpkg/info/python3-minimal.postinst
evals	Retirar Python del runner	2026-09-15T17:15:56.9865913Z retirado: /var/lib/docker/overlay2/80c613159e0ff38a882acb63fd4cf342723381c4f07d119df138a79409f5e793/diff/var/lib/dpkg/info/libpython3.10-stdlib:amd64.list
evals	Retirar Python del runner	2026-09-15T17:15:56.9997842Z retirado: /var/lib/docker/overlay2/80c613159e0ff38a882acb63fd4cf342723381c4f07d119df138a79409f5e793/diff/var/lib/dpkg/info/python3.10-minimal.postinst
evals	Retirar Python del runner	2026-09-15T17:15:57.0133609Z retirado: /var/lib/docker/overlay2/80c613159e0ff38a882acb63fd4cf342723381c4f07d119df138a79409f5e793/diff/var/lib/dpkg/info/python3.10.postinst
evals	Retirar Python del runner	2026-09-15T17:15:57.0274968Z retirado: /var/lib/docker/overlay2/80c613159e0ff38a882acb63fd4cf342723381c4f07d119df138a79409f5e793/diff/var/lib/dpkg/info/python3.prerm
evals	Retirar Python del runner	2026-09-15T17:15:57.0410285Z retirado: /var/lib/docker/overlay2/80c613159e0ff38a882acb63fd4cf342723381c4f07d119df138a79409f5e793/diff/var/lib/dpkg/info/libpython3.10-stdlib:amd64.prerm
evals	Retirar Python del runner	2026-09-15T17:15:57.0541825Z retirado: /var/lib/docker/overlay2/80c613159e0ff38a882acb63fd4cf342723381c4f07d119df138a79409f5e793/diff/var/lib/dpkg/info/libpython3.10-minimal:amd64.md5sums
evals	Retirar Python del runner	2026-09-15T17:15:57.0673424Z retirado: /var/lib/docker/overlay2/80c613159e0ff38a882acb63fd4cf342723381c4f07d119df138a79409f5e793/diff/var/lib/dpkg/info/libpython3.10-stdlib:amd64.md5sums
evals	Retirar Python del runner	2026-09-15T17:15:57.0805483Z retirado: /var/lib/docker/overlay2/80c613159e0ff38a882acb63fd4cf342723381c4f07d119df138a79409f5e793/diff/var/lib/dpkg/info/python3.postrm
evals	Retirar Python del runner	2026-09-15T17:15:57.0942001Z retirado: /var/lib/docker/overlay2/80c613159e0ff38a882acb63fd4cf342723381c4f07d119df138a79409f5e793/diff/var/lib/dpkg/info/libpython3.10-minimal:amd64.conffiles
evals	Retirar Python del runner	2026-09-15T17:15:57.1076005Z retirado: /var/lib/docker/overlay2/80c613159e0ff38a882acb63fd4cf342723381c4f07d119df138a79409f5e793/diff/var/lib/dpkg/info/python3.10-minimal.preinst
evals	Retirar Python del runner	2026-09-15T17:15:57.1208778Z retirado: /var/lib/docker/overlay2/80c613159e0ff38a882acb63fd4cf342723381c4f07d119df138a79409f5e793/diff/var/lib/dpkg/info/python3.postinst
evals	Retirar Python del runner	2026-09-15T17:15:57.1343653Z retirado: /var/lib/docker/overlay2/80c613159e0ff38a882acb63fd4cf342723381c4f07d119df138a79409f5e793/diff/usr/share/pixmaps/python3.xpm
evals	Retirar Python del runner	2026-09-15T17:15:57.1476035Z retirado: /var/lib/docker/overlay2/80c613159e0ff38a882acb63fd4cf342723381c4f07d119df138a79409f5e793/diff/usr/share/doc/libpython3.10-stdlib
evals	Retirar Python del runner	2026-09-15T17:15:57.1662809Z retirado: /var/lib/docker/overlay2/80c613159e0ff38a882acb63fd4cf342723381c4f07d119df138a79409f5e793/diff/usr/share/doc/python3.10/python-policy.txt.gz
evals	Retirar Python del runner	2026-09-15T17:15:57.1791431Z retirado: /var/lib/docker/overlay2/80c613159e0ff38a882acb63fd4cf342723381c4f07d119df138a79409f5e793/diff/usr/share/lintian/overrides/libpython3.10-stdlib
evals	Retirar Python del runner	2026-09-15T17:15:57.1936465Z retirado: /var/lib/docker/overlay2/80c613159e0ff38a882acb63fd4cf342723381c4f07d119df138a79409f5e793/diff/usr/share/lintian/overrides/libpython3.10-minimal
evals	Retirar Python del runner	2026-09-15T17:15:57.2250655Z retirado: /var/lib/docker/overlay2/80c613159e0ff38a882acb63fd4cf342723381c4f07d119df138a79409f5e793/diff/usr
evals	Retirar Python del runner	2026-09-15T17:15:57.3884915Z retirado: /usr/lib/x86_64-linux-gnu/pkgconfig/python3.pc
evals	Retirar Python del runner	2026-09-15T17:15:57.4020503Z retirado: /usr/lib/x86_64-linux-gnu/pkgconfig/python3-embed.pc
evals	Retirar Python del runner	2026-09-15T17:15:57.4150029Z retirado: /usr/lib/x86_64-linux-gnu/libpython3.12.so.1
evals	Retirar Python del runner	2026-09-15T17:15:57.4336227Z retirado: /usr/lib/x86_64-linux-gnu/libpython3.12.so
evals	Retirar Python del runner	2026-09-15T17:15:57.4480639Z retirado: /usr/lib/x86_64-linux-gnu/libpython3.12.a
evals	Retirar Python del runner	2026-09-15T17:15:57.4601721Z retirado: /usr/lib/x86_64-linux-gnu/libpython3.12.so.1.0
evals	Retirar Python del runner	2026-09-15T17:15:57.4783789Z retirado: /usr/lib/python3.12/config-3.12-x86_64-linux-gnu/libpython3.12.so
evals	Retirar Python del runner	2026-09-15T17:15:57.4911225Z retirado: /usr/lib/python3.12/config-3.12-x86_64-linux-gnu/libpython3.12.a
evals	Retirar Python del runner	2026-09-15T17:15:57.5036524Z retirado: /usr/lib/python3.12/config-3.12-x86_64-linux-gnu/python-config.py
evals	Retirar Python del runner	2026-09-15T17:15:57.5161359Z retirado: /usr/lib/python3.12/config-3.12-x86_64-linux-gnu/libpython3.12-pic.a
evals	Retirar Python del runner	2026-09-15T17:15:57.5288786Z retirado: /usr/lib/google-cloud-sdk/lib/googlecloudsdk/command_lib/orchestration_pipelines/tools/python_environment_unpack.sh
evals	Retirar Python del runner	2026-09-15T17:15:57.5421894Z retirado: /usr/lib/google-cloud-sdk/platform/bundledpythonunix/lib/libpython3.so
evals	Retirar Python del runner	2026-09-15T17:15:57.5550167Z retirado: /usr/lib/google-cloud-sdk/platform/bundledpythonunix/lib/pkgconfig/python3.pc
evals	Retirar Python del runner	2026-09-15T17:15:57.5677052Z retirado: /usr/lib/google-cloud-sdk/platform/bundledpythonunix/lib/pkgconfig/python3-embed.pc
evals	Retirar Python del runner	2026-09-15T17:15:57.5804153Z retirado: /usr/lib/google-cloud-sdk/platform/bundledpythonunix/lib/python3.14/config-3.14-x86_64-linux-gnu/python-config.py
evals	Retirar Python del runner	2026-09-15T17:15:57.5931632Z retirado: /usr/lib/google-cloud-sdk/platform/bundledpythonunix/lib/libpython3.14.so
evals	Retirar Python del runner	2026-09-15T17:15:57.6057694Z retirado: /usr/lib/google-cloud-sdk/platform/bundledpythonunix/lib/libpython3.14.so.1.0
evals	Retirar Python del runner	2026-09-15T17:15:57.6304619Z retirado: /usr/lib/google-cloud-sdk/platform/bundledpythonunix
evals	Retirar Python del runner	2026-09-15T17:15:57.7616864Z retirado: /usr/lib/rpm/pythondistdeps.py
evals	Retirar Python del runner	2026-09-15T17:15:57.7743118Z retirado: /usr/local/aws-cli/v2/2.36.40/dist/libpython3.14.so.1.0
evals	Retirar Python del runner	2026-09-15T17:15:57.7877179Z retirado: /usr/local/lib/android/sdk/ndk/28.2.13676358/toolchains/llvm/prebuilt/linux-x86_64/python3/lib/libpython3.so
evals	Retirar Python del runner	2026-09-15T17:15:57.8041802Z retirado: /usr/local/lib/android/sdk/ndk/28.2.13676358/toolchains/llvm/prebuilt/linux-x86_64/python3/lib/pkgconfig/python3.pc
evals	Retirar Python del runner	2026-09-15T17:15:57.8171858Z retirado: /usr/local/lib/android/sdk/ndk/28.2.13676358/toolchains/llvm/prebuilt/linux-x86_64/python3/lib/pkgconfig/python3-embed.pc
evals	Retirar Python del runner	2026-09-15T17:15:57.8301480Z retirado: /usr/local/lib/android/sdk/ndk/28.2.13676358/toolchains/llvm/prebuilt/linux-x86_64/python3/lib/pkgconfig/python-3.11-embed.pc
evals	Retirar Python del runner	2026-09-15T17:15:57.8432854Z retirado: /usr/local/lib/android/sdk/ndk/28.2.13676358/toolchains/llvm/prebuilt/linux-x86_64/python3/lib/pkgconfig/python-3.11.pc
evals	Retirar Python del runner	2026-09-15T17:15:57.8578339Z retirado: /usr/local/lib/android/sdk/ndk/28.2.13676358/toolchains/llvm/prebuilt/linux-x86_64/python3/lib/python3.11/site-packages/pip/_vendor/pygments/lexers/python.py
evals	Retirar Python del runner	2026-09-15T17:15:57.8779705Z retirado: /usr/local/lib/android/sdk/ndk/28.2.13676358/toolchains/llvm/prebuilt/linux-x86_64/python3/lib/libpython3.11.so
evals	Retirar Python del runner	2026-09-15T17:15:57.8912091Z retirado: /usr/local/lib/android/sdk/ndk/28.2.13676358/toolchains/llvm/prebuilt/linux-x86_64/python3/lib/libpython3.11.so.1.0
evals	Retirar Python del runner	2026-09-15T17:15:57.9046059Z retirado: /usr/local/lib/android/sdk/ndk/28.2.13676358/toolchains/llvm/prebuilt/linux-x86_64/python3/include/python3.11/cpython/pythonrun.h
evals	Retirar Python del runner	2026-09-15T17:15:57.9172022Z retirado: /usr/local/lib/android/sdk/ndk/28.2.13676358/toolchains/llvm/prebuilt/linux-x86_64/python3/include/python3.11/Python.h
evals	Retirar Python del runner	2026-09-15T17:15:57.9300596Z retirado: /usr/local/lib/android/sdk/ndk/28.2.13676358/toolchains/llvm/prebuilt/linux-x86_64/python3/include/python3.11/pythonrun.h
evals	Retirar Python del runner	2026-09-15T17:15:57.9430713Z retirado: /usr/local/lib/android/sdk/ndk/28.2.13676358/toolchains/llvm/prebuilt/linux-x86_64/python3/share/man/man1/python3.1
evals	Retirar Python del runner	2026-09-15T17:15:57.9557865Z retirado: /usr/local/lib/android/sdk/ndk/28.2.13676358/toolchains/llvm/prebuilt/linux-x86_64/python3/share/man/man1/python3.11.1
evals	Retirar Python del runner	2026-09-15T17:15:57.9811550Z retirado: /usr/local/lib/android/sdk/ndk/28.2.13676358/toolchains/llvm/prebuilt/linux-x86_64/python3
evals	Retirar Python del runner	2026-09-15T17:15:58.0245506Z retirado: /usr/local/lib/android/sdk/ndk/27.3.13750724/toolchains/llvm/prebuilt/linux-x86_64/python3/lib/libpython3.so
evals	Retirar Python del runner	2026-09-15T17:15:58.0380191Z retirado: /usr/local/lib/android/sdk/ndk/27.3.13750724/toolchains/llvm/prebuilt/linux-x86_64/python3/lib/pkgconfig/python3.pc
evals	Retirar Python del runner	2026-09-15T17:15:58.0510496Z retirado: /usr/local/lib/android/sdk/ndk/27.3.13750724/toolchains/llvm/prebuilt/linux-x86_64/python3/lib/pkgconfig/python3-embed.pc
evals	Retirar Python del runner	2026-09-15T17:15:58.0640654Z retirado: /usr/local/lib/android/sdk/ndk/27.3.13750724/toolchains/llvm/prebuilt/linux-x86_64/python3/lib/pkgconfig/python-3.11-embed.pc
evals	Retirar Python del runner	2026-09-15T17:15:58.0768725Z retirado: /usr/local/lib/android/sdk/ndk/27.3.13750724/toolchains/llvm/prebuilt/linux-x86_64/python3/lib/pkgconfig/python-3.11.pc
evals	Retirar Python del runner	2026-09-15T17:15:58.0900240Z retirado: /usr/local/lib/android/sdk/ndk/27.3.13750724/toolchains/llvm/prebuilt/linux-x86_64/python3/lib/python3.11/site-packages/pip/_vendor/pygments/lexers/python.py
evals	Retirar Python del runner	2026-09-15T17:15:58.1026320Z retirado: /usr/local/lib/android/sdk/ndk/27.3.13750724/toolchains/llvm/prebuilt/linux-x86_64/python3/lib/libpython3.11.so
evals	Retirar Python del runner	2026-09-15T17:15:58.1154141Z retirado: /usr/local/lib/android/sdk/ndk/27.3.13750724/toolchains/llvm/prebuilt/linux-x86_64/python3/lib/libpython3.11.so.1.0
evals	Retirar Python del runner	2026-09-15T17:15:58.1281546Z retirado: /usr/local/lib/android/sdk/ndk/27.3.13750724/toolchains/llvm/prebuilt/linux-x86_64/python3/include/python3.11/cpython/pythonrun.h
evals	Retirar Python del runner	2026-09-15T17:15:58.1411921Z retirado: /usr/local/lib/android/sdk/ndk/27.3.13750724/toolchains/llvm/prebuilt/linux-x86_64/python3/include/python3.11/Python.h
evals	Retirar Python del runner	2026-09-15T17:15:58.1537409Z retirado: /usr/local/lib/android/sdk/ndk/27.3.13750724/toolchains/llvm/prebuilt/linux-x86_64/python3/include/python3.11/pythonrun.h
evals	Retirar Python del runner	2026-09-15T17:15:58.1669614Z retirado: /usr/local/lib/android/sdk/ndk/27.3.13750724/toolchains/llvm/prebuilt/linux-x86_64/python3/share/man/man1/python3.1
evals	Retirar Python del runner	2026-09-15T17:15:58.1796933Z retirado: /usr/local/lib/android/sdk/ndk/27.3.13750724/toolchains/llvm/prebuilt/linux-x86_64/python3/share/man/man1/python3.11.1
evals	Retirar Python del runner	2026-09-15T17:15:58.2045415Z retirado: /usr/local/lib/android/sdk/ndk/27.3.13750724/toolchains/llvm/prebuilt/linux-x86_64/python3
evals	Retirar Python del runner	2026-09-15T17:15:58.2501100Z retirado: /usr/local/lib/android/sdk/ndk/29.0.14206865/toolchains/llvm/prebuilt/linux-x86_64/python3/lib/libpython3.so
evals	Retirar Python del runner	2026-09-15T17:15:58.2629858Z retirado: /usr/local/lib/android/sdk/ndk/29.0.14206865/toolchains/llvm/prebuilt/linux-x86_64/python3/lib/pkgconfig/python3.pc
evals	Retirar Python del runner	2026-09-15T17:15:58.2760456Z retirado: /usr/local/lib/android/sdk/ndk/29.0.14206865/toolchains/llvm/prebuilt/linux-x86_64/python3/lib/pkgconfig/python3-embed.pc
evals	Retirar Python del runner	2026-09-15T17:15:58.2887687Z retirado: /usr/local/lib/android/sdk/ndk/29.0.14206865/toolchains/llvm/prebuilt/linux-x86_64/python3/lib/pkgconfig/python-3.11-embed.pc
evals	Retirar Python del runner	2026-09-15T17:15:58.3013972Z retirado: /usr/local/lib/android/sdk/ndk/29.0.14206865/toolchains/llvm/prebuilt/linux-x86_64/python3/lib/pkgconfig/python-3.11.pc
evals	Retirar Python del runner	2026-09-15T17:15:58.3144264Z retirado: /usr/local/lib/android/sdk/ndk/29.0.14206865/toolchains/llvm/prebuilt/linux-x86_64/python3/lib/python3.11/site-packages/pip/_vendor/pygments/lexers/python.py
evals	Retirar Python del runner	2026-09-15T17:15:58.3276234Z retirado: /usr/local/lib/android/sdk/ndk/29.0.14206865/toolchains/llvm/prebuilt/linux-x86_64/python3/lib/libpython3.11.so
evals	Retirar Python del runner	2026-09-15T17:15:58.3407627Z retirado: /usr/local/lib/android/sdk/ndk/29.0.14206865/toolchains/llvm/prebuilt/linux-x86_64/python3/lib/libpython3.11.so.1.0
evals	Retirar Python del runner	2026-09-15T17:15:58.3541241Z retirado: /usr/local/lib/android/sdk/ndk/29.0.14206865/toolchains/llvm/prebuilt/linux-x86_64/python3/include/python3.11/cpython/pythonrun.h
evals	Retirar Python del runner	2026-09-15T17:15:58.3670931Z retirado: /usr/local/lib/android/sdk/ndk/29.0.14206865/toolchains/llvm/prebuilt/linux-x86_64/python3/include/python3.11/Python.h
evals	Retirar Python del runner	2026-09-15T17:15:58.3803743Z retirado: /usr/local/lib/android/sdk/ndk/29.0.14206865/toolchains/llvm/prebuilt/linux-x86_64/python3/include/python3.11/pythonrun.h
evals	Retirar Python del runner	2026-09-15T17:15:58.3934119Z retirado: /usr/local/lib/android/sdk/ndk/29.0.14206865/toolchains/llvm/prebuilt/linux-x86_64/python3/share/man/man1/python3.1
evals	Retirar Python del runner	2026-09-15T17:15:58.4070741Z retirado: /usr/local/lib/android/sdk/ndk/29.0.14206865/toolchains/llvm/prebuilt/linux-x86_64/python3/share/man/man1/python3.11.1
evals	Retirar Python del runner	2026-09-15T17:15:58.4331892Z retirado: /usr/local/lib/android/sdk/ndk/29.0.14206865/toolchains/llvm/prebuilt/linux-x86_64/python3
evals	Retirar Python del runner	2026-09-15T17:15:58.4772804Z retirado: /usr/local/aws-sam-cli/1.166.1/dist/_internal/libpython3.11.so.1.0
evals	Retirar Python del runner	2026-09-15T17:15:58.4912367Z retirado: /usr/local/share/vcpkg/ports/libudis86/python3.patch
evals	Retirar Python del runner	2026-09-15T17:15:58.5050709Z retirado: /usr/local/share/vcpkg/ports/omniorb/python-fixes.patch
evals	Retirar Python del runner	2026-09-15T17:15:58.5185695Z retirado: /usr/local/share/vcpkg/ports/openxr-loader/python3_8_compatibility.patch
evals	Retirar Python del runner	2026-09-15T17:15:58.5321082Z retirado: /usr/local/share/vcpkg/ports/libxslt/python3.patch
evals	Retirar Python del runner	2026-09-15T17:15:58.5450464Z retirado: /usr/local/share/vcpkg/ports/openscap/python-win32.diff
evals	Retirar Python del runner	2026-09-15T17:15:58.5579302Z retirado: /usr/local/share/vcpkg/ports/python3/python_vcpkg.props.in
evals	Retirar Python del runner	2026-09-15T17:15:58.5708061Z retirado: /usr/local/share/vcpkg/ports/vtk/pythonwrapper.patch
evals	Retirar Python del runner	2026-09-15T17:15:58.5839660Z retirado: /usr/local/share/vcpkg/scripts/test_ports/vcpkg-ci-blender/python.patch
evals	Retirar Python del runner	2026-09-15T17:15:58.5968846Z retirado: /usr/local/share/vcpkg/versions/p-/python2.json
evals	Retirar Python del runner	2026-09-15T17:15:58.6103710Z retirado: /usr/local/share/vcpkg/versions/p-/python3.json
evals	Retirar Python del runner	2026-09-15T17:15:58.6235295Z retirado: /usr/share/perl5/NeedRestart/Interp/Python.pm
evals	Retirar Python del runner	2026-09-15T17:15:58.6369861Z retirado: /usr/share/doc-base/python3.python-policy
evals	Retirar Python del runner	2026-09-15T17:15:58.6501443Z retirado: /usr/share/az_15.6.1/Az.Functions/4.3.2/Functions.Autorest/custom/FunctionsStackFlexData/EastAsia/python.json
evals	Retirar Python del runner	2026-09-15T17:15:58.6633603Z retirado: /usr/share/bash-completion/completions/python3.9
evals	Retirar Python del runner	2026-09-15T17:15:58.6764204Z retirado: /usr/share/bash-completion/completions/python3.7
evals	Retirar Python del runner	2026-09-15T17:15:58.6892869Z retirado: /usr/share/bash-completion/completions/python3.3
evals	Retirar Python del runner	2026-09-15T17:15:58.7022553Z retirado: /usr/share/bash-completion/completions/python3.6
evals	Retirar Python del runner	2026-09-15T17:15:58.7155015Z retirado: /usr/share/bash-completion/completions/python2.7
evals	Retirar Python del runner	2026-09-15T17:15:58.7291678Z retirado: /usr/share/bash-completion/completions/python3.4
evals	Retirar Python del runner	2026-09-15T17:15:58.7423549Z retirado: /usr/share/bash-completion/completions/pypy3
evals	Retirar Python del runner	2026-09-15T17:15:58.7558714Z retirado: /usr/share/bash-completion/completions/python2
evals	Retirar Python del runner	2026-09-15T17:15:58.7690585Z retirado: /usr/share/bash-completion/completions/pypy
evals	Retirar Python del runner	2026-09-15T17:15:58.7818846Z retirado: /usr/share/bash-completion/completions/python3.8
evals	Retirar Python del runner	2026-09-15T17:15:58.7950675Z retirado: /usr/share/bash-completion/completions/python3.5
evals	Retirar Python del runner	2026-09-15T17:15:58.8078156Z retirado: /usr/share/bash-completion/completions/python3
evals	Retirar Python del runner	2026-09-15T17:15:58.8207250Z retirado: /usr/share/bash-completion/completions/python
evals	Retirar Python del runner	2026-09-15T17:15:58.8337351Z retirado: /usr/share/bash-completion/helpers/python
evals	Retirar Python del runner	2026-09-15T17:15:58.8468858Z retirado: /usr/share/man/man8/pythoncalls-bpfcc.8.gz
evals	Retirar Python del runner	2026-09-15T17:15:58.8602530Z retirado: /usr/share/man/man8/pythonstat-bpfcc.8.gz
evals	Retirar Python del runner	2026-09-15T17:15:58.8732802Z retirado: /usr/share/man/man8/pythonflow-bpfcc.8.gz
evals	Retirar Python del runner	2026-09-15T17:15:58.8864993Z retirado: /usr/share/man/man8/pythongc-bpfcc.8.gz
evals	Retirar Python del runner	2026-09-15T17:15:58.8995137Z retirado: /usr/share/man/man1/python3.12.1.gz
evals	Retirar Python del runner	2026-09-15T17:15:58.9178742Z retirado: /usr/share/man/man1/python.1.gz
evals	Retirar Python del runner	2026-09-15T17:15:58.9310558Z retirado: /usr/share/man/man1/python3.12-config.1.gz
evals	Retirar Python del runner	2026-09-15T17:15:58.9494221Z retirado: /usr/share/man/man1/python3.1.gz
evals	Retirar Python del runner	2026-09-15T17:15:58.9678879Z retirado: /usr/share/man/man1/python3-config.1.gz
evals	Retirar Python del runner	2026-09-15T17:15:58.9808081Z retirado: /usr/share/pixmaps/python3.xpm
evals	Retirar Python del runner	2026-09-15T17:15:58.9937328Z retirado: /usr/share/pixmaps/python3.12.xpm
evals	Retirar Python del runner	2026-09-15T17:15:59.0070811Z retirado: /usr/share/binfmts/python3.12
evals	Retirar Python del runner	2026-09-15T17:15:59.0206014Z retirado: /usr/share/vim/vim91/syntax/python2.vim
evals	Retirar Python del runner	2026-09-15T17:15:59.0340440Z retirado: /usr/share/vim/vim91/syntax/python.vim
evals	Retirar Python del runner	2026-09-15T17:15:59.0473943Z retirado: /usr/share/vim/vim91/autoload/pythoncomplete.vim
evals	Retirar Python del runner	2026-09-15T17:15:59.0608374Z retirado: /usr/share/vim/vim91/autoload/python3complete.vim
evals	Retirar Python del runner	2026-09-15T17:15:59.0744100Z retirado: /usr/share/vim/vim91/autoload/python.vim
evals	Retirar Python del runner	2026-09-15T17:15:59.0873853Z retirado: /usr/share/vim/vim91/ftplugin/python.vim
evals	Retirar Python del runner	2026-09-15T17:15:59.1005266Z retirado: /usr/share/vim/vim91/indent/python.vim
evals	Retirar Python del runner	2026-09-15T17:15:59.1134436Z retirado: /usr/share/swig4.0/python/pythonkw.swg
evals	Retirar Python del runner	2026-09-15T17:15:59.1270972Z retirado: /usr/share/swig4.0/python/python.swg
evals	Retirar Python del runner	2026-09-15T17:15:59.1404556Z retirado: /usr/share/applications/python3.12.desktop
evals	Retirar Python del runner	2026-09-15T17:15:59.1536712Z retirado: /usr/share/doc/python3.12-venv
evals	Retirar Python del runner	2026-09-15T17:15:59.1666622Z retirado: /usr/share/doc/libpython3.12t64
evals	Retirar Python del runner	2026-09-15T17:15:59.1795427Z retirado: /usr/share/doc/python3-setuptools/python 2 sunset.rst
evals	Retirar Python del runner	2026-09-15T17:15:59.1921072Z retirado: /usr/share/doc/libpython3.12-dev
evals	Retirar Python del runner	2026-09-15T17:15:59.2050110Z retirado: /usr/share/doc/mercurial-common/examples/python-hook-examples.py
evals	Retirar Python del runner	2026-09-15T17:15:59.2178541Z retirado: /usr/share/doc/python3.12-dev
evals	Retirar Python del runner	2026-09-15T17:15:59.2308528Z retirado: /usr/share/doc/python3-pip/html/topics/python-option.md
evals	Retirar Python del runner	2026-09-15T17:15:59.2437491Z retirado: /usr/share/doc/python3-venv
evals	Retirar Python del runner	2026-09-15T17:15:59.2563593Z retirado: /usr/share/doc/python3.12/python-policy.txt.gz
evals	Retirar Python del runner	2026-09-15T17:15:59.2689594Z retirado: /usr/share/doc/libpython3.12-stdlib
evals	Retirar Python del runner	2026-09-15T17:15:59.2815050Z retirado: /usr/share/doc/python3-dev
evals	Retirar Python del runner	2026-09-15T17:15:59.2941841Z retirado: /usr/share/doc/bpfcc-tools/examples/doc/pythonstat_example.txt
evals	Retirar Python del runner	2026-09-15T17:15:59.3071345Z retirado: /usr/share/doc/bpfcc-tools/examples/doc/pythonflow_example.txt
evals	Retirar Python del runner	2026-09-15T17:15:59.3196936Z retirado: /usr/share/doc/bpfcc-tools/examples/doc/pythoncalls_example.txt
evals	Retirar Python del runner	2026-09-15T17:15:59.3323370Z retirado: /usr/share/doc/bpfcc-tools/examples/doc/pythongc_example.txt
evals	Retirar Python del runner	2026-09-15T17:15:59.3454024Z retirado: /usr/share/doc/python3-debconf
evals	Retirar Python del runner	2026-09-15T17:15:59.3583578Z retirado: /usr/share/doc/python3/python-policy.txt.gz
evals	Retirar Python del runner	2026-09-15T17:15:59.3715697Z retirado: /usr/share/doc/python3/python-policy.html
evals	Retirar Python del runner	2026-09-15T17:15:59.3844174Z retirado: /usr/share/aclocal-1.16/python.m4
evals	Retirar Python del runner	2026-09-15T17:15:59.3974385Z retirado: /usr/share/miniconda/lib/libpython3.so
evals	Retirar Python del runner	2026-09-15T17:15:59.4102637Z retirado: /usr/share/miniconda/lib/pkgconfig/python-3.14-embed.pc
evals	Retirar Python del runner	2026-09-15T17:15:59.4235674Z retirado: /usr/share/miniconda/lib/pkgconfig/python3.pc
evals	Retirar Python del runner	2026-09-15T17:15:59.4424294Z retirado: /usr/share/miniconda/lib/pkgconfig/python3-embed.pc
evals	Retirar Python del runner	2026-09-15T17:15:59.4554930Z retirado: /usr/share/miniconda/lib/pkgconfig/python-3.14.pc
evals	Retirar Python del runner	2026-09-15T17:15:59.4686321Z retirado: /usr/share/miniconda/lib/python3.14/site-packages/conda/common/path/__pycache__/python.cpython-314.pyc
evals	Retirar Python del runner	2026-09-15T17:15:59.4816478Z retirado: /usr/share/miniconda/lib/python3.14/site-packages/conda/common/path/python.py
evals	Retirar Python del runner	2026-09-15T17:15:59.4943411Z retirado: /usr/share/miniconda/lib/python3.14/site-packages/pygments/lexers/__pycache__/python.cpython-314.pyc
evals	Retirar Python del runner	2026-09-15T17:15:59.5072035Z retirado: /usr/share/miniconda/lib/python3.14/site-packages/pygments/lexers/python.py
evals	Retirar Python del runner	2026-09-15T17:15:59.5200484Z retirado: /usr/share/miniconda/lib/python3.14/site-packages/pip/_vendor/pygments/lexers/__pycache__/python.cpython-314.pyc
evals	Retirar Python del runner	2026-09-15T17:15:59.5331113Z retirado: /usr/share/miniconda/lib/python3.14/site-packages/pip/_vendor/pygments/lexers/python.py
evals	Retirar Python del runner	2026-09-15T17:15:59.5460828Z retirado: /usr/share/miniconda/lib/python3.14/site-packages/conda_pypi/__pycache__/python_paths.cpython-314.pyc
evals	Retirar Python del runner	2026-09-15T17:15:59.5588301Z retirado: /usr/share/miniconda/lib/python3.14/site-packages/conda_pypi/python_paths.py
evals	Retirar Python del runner	2026-09-15T17:15:59.5717428Z retirado: /usr/share/miniconda/lib/python3.14/config-3.14-x86_64-linux-gnu/__pycache__/python-config.cpython-314.pyc
evals	Retirar Python del runner	2026-09-15T17:15:59.5846615Z retirado: /usr/share/miniconda/lib/python3.14/config-3.14-x86_64-linux-gnu/python-config.py
evals	Retirar Python del runner	2026-09-15T17:15:59.5977897Z retirado: /usr/share/miniconda/lib/python3.14/config-3.14-x86_64-linux-gnu/python.o
evals	Retirar Python del runner	2026-09-15T17:15:59.6106522Z retirado: /usr/share/miniconda/lib/python3.14/idlelib/Icons/python.gif
evals	Retirar Python del runner	2026-09-15T17:15:59.6236358Z retirado: /usr/share/miniconda/lib/libpython3.14.so
evals	Retirar Python del runner	2026-09-15T17:15:59.6365360Z retirado: /usr/share/miniconda/lib/libpython3.14.so.1.0
evals	Retirar Python del runner	2026-09-15T17:15:59.6497460Z retirado: /usr/share/miniconda/include/python3.14/cpython/pythonrun.h
evals	Retirar Python del runner	2026-09-15T17:15:59.6627068Z retirado: /usr/share/miniconda/include/python3.14/Python.h
evals	Retirar Python del runner	2026-09-15T17:15:59.6758184Z retirado: /usr/share/miniconda/include/python3.14/pythonrun.h
evals	Retirar Python del runner	2026-09-15T17:15:59.6889006Z retirado: /usr/share/miniconda/share/man/man1/python3.1
evals	Retirar Python del runner	2026-09-15T17:15:59.7019797Z retirado: /usr/share/miniconda/share/man/man1/python3.14.1
evals	Retirar Python del runner	2026-09-15T17:15:59.7151327Z retirado: /usr/share/miniconda/conda-meta/python_abi-3.14-4_cp314.json
evals	Retirar Python del runner	2026-09-15T17:15:59.7281161Z retirado: /usr/share/miniconda/conda-meta/python-installer-1.0.1-py314h06a4308_0.json
evals	Retirar Python del runner	2026-09-15T17:15:59.7412493Z retirado: /usr/share/miniconda/conda-meta/python-build-1.5.1-py314h06a4308_0.json
evals	Retirar Python del runner	2026-09-15T17:15:59.7543402Z retirado: /usr/share/miniconda/conda-meta/python-3.14.7-h2bd7c14_101_cp314.json
evals	Retirar Python del runner	2026-09-15T17:15:59.7675562Z retirado: /usr/share/miniconda/conda-meta/python-dotenv-1.2.2-py314h06a4308_0.json
evals	Retirar Python del runner	2026-09-15T17:15:59.7805332Z retirado: /usr/share/miniconda/pkgs/python-build-1.5.1-py314h06a4308_0.conda
evals	Retirar Python del runner	2026-09-15T17:15:59.8049741Z retirado: /usr/share/miniconda/pkgs/python-build-1.5.1-py314h06a4308_0
evals	Retirar Python del runner	2026-09-15T17:15:59.8205195Z retirado: /usr/share/miniconda/pkgs/conda-26.7.1-py314h06a4308_0/lib/python3.14/site-packages/conda/common/path/__pycache__/python.cpython-314.pyc
evals	Retirar Python del runner	2026-09-15T17:15:59.8334487Z retirado: /usr/share/miniconda/pkgs/conda-26.7.1-py314h06a4308_0/lib/python3.14/site-packages/conda/common/path/python.py
evals	Retirar Python del runner	2026-09-15T17:15:59.8466253Z retirado: /usr/share/miniconda/pkgs/python-3.14.7-h2bd7c14_101_cp314.conda
evals	Retirar Python del runner	2026-09-15T17:15:59.8600761Z retirado: /usr/share/miniconda/pkgs/python-dotenv-1.2.2-py314h06a4308_0.conda
evals	Retirar Python del runner	2026-09-15T17:15:59.8728846Z retirado: /usr/share/miniconda/pkgs/libxcb-1.17.0-h9b100fa_0/info/recipe/python3.patch
evals	Retirar Python del runner	2026-09-15T17:15:59.8856162Z retirado: /usr/share/miniconda/pkgs/pip-26.2.1-pyh0d26453_0/site-packages/pip/_vendor/pygments/lexers/python.py
evals	Retirar Python del runner	2026-09-15T17:15:59.8989338Z retirado: /usr/share/miniconda/pkgs/python-3.14.7-h2bd7c14_101_cp314/lib/libpython3.so
evals	Retirar Python del runner	2026-09-15T17:15:59.9115891Z retirado: /usr/share/miniconda/pkgs/python-3.14.7-h2bd7c14_101_cp314/lib/pkgconfig/python-3.14-embed.pc
evals	Retirar Python del runner	2026-09-15T17:15:59.9243467Z retirado: /usr/share/miniconda/pkgs/python-3.14.7-h2bd7c14_101_cp314/lib/pkgconfig/python3.pc
evals	Retirar Python del runner	2026-09-15T17:15:59.9426016Z retirado: /usr/share/miniconda/pkgs/python-3.14.7-h2bd7c14_101_cp314/lib/pkgconfig/python3-embed.pc
evals	Retirar Python del runner	2026-09-15T17:15:59.9555330Z retirado: /usr/share/miniconda/pkgs/python-3.14.7-h2bd7c14_101_cp314/lib/pkgconfig/python-3.14.pc
evals	Retirar Python del runner	2026-09-15T17:15:59.9685588Z retirado: /usr/share/miniconda/pkgs/python-3.14.7-h2bd7c14_101_cp314/lib/python3.14/config-3.14-x86_64-linux-gnu/__pycache__/python-config.cpython-314.pyc
evals	Retirar Python del runner	2026-09-15T17:15:59.9814037Z retirado: /usr/share/miniconda/pkgs/python-3.14.7-h2bd7c14_101_cp314/lib/python3.14/config-3.14-x86_64-linux-gnu/python-config.py
evals	Retirar Python del runner	2026-09-15T17:15:59.9944464Z retirado: /usr/share/miniconda/pkgs/python-3.14.7-h2bd7c14_101_cp314/lib/python3.14/config-3.14-x86_64-linux-gnu/python.o
evals	Retirar Python del runner	2026-09-15T17:16:00.0072062Z retirado: /usr/share/miniconda/pkgs/python-3.14.7-h2bd7c14_101_cp314/lib/python3.14/idlelib/Icons/python.gif
evals	Retirar Python del runner	2026-09-15T17:16:00.0204173Z retirado: /usr/share/miniconda/pkgs/python-3.14.7-h2bd7c14_101_cp314/lib/libpython3.14.so
evals	Retirar Python del runner	2026-09-15T17:16:00.0333229Z retirado: /usr/share/miniconda/pkgs/python-3.14.7-h2bd7c14_101_cp314/lib/libpython3.14.so.1.0
evals	Retirar Python del runner	2026-09-15T17:16:00.0465952Z retirado: /usr/share/miniconda/pkgs/python-3.14.7-h2bd7c14_101_cp314/include/python3.14/cpython/pythonrun.h
evals	Retirar Python del runner	2026-09-15T17:16:00.0599219Z retirado: /usr/share/miniconda/pkgs/python-3.14.7-h2bd7c14_101_cp314/include/python3.14/Python.h
evals	Retirar Python del runner	2026-09-15T17:16:00.0731762Z retirado: /usr/share/miniconda/pkgs/python-3.14.7-h2bd7c14_101_cp314/include/python3.14/pythonrun.h
evals	Retirar Python del runner	2026-09-15T17:16:00.0861359Z retirado: /usr/share/miniconda/pkgs/python-3.14.7-h2bd7c14_101_cp314/share/man/man1/python3.1
evals	Retirar Python del runner	2026-09-15T17:16:00.0990986Z retirado: /usr/share/miniconda/pkgs/python-3.14.7-h2bd7c14_101_cp314/share/man/man1/python3.14.1
evals	Retirar Python del runner	2026-09-15T17:16:00.1250627Z retirado: /usr/share/miniconda/pkgs/python-3.14.7-h2bd7c14_101_cp314
evals	Retirar Python del runner	2026-09-15T17:16:00.2109612Z retirado: /usr/share/miniconda/pkgs/pygments-2.20.0-py314h06a4308_0/lib/python3.14/site-packages/pygments/lexers/__pycache__/python.cpython-314.pyc
evals	Retirar Python del runner	2026-09-15T17:16:00.2241103Z retirado: /usr/share/miniconda/pkgs/pygments-2.20.0-py314h06a4308_0/lib/python3.14/site-packages/pygments/lexers/python.py
evals	Retirar Python del runner	2026-09-15T17:16:00.2375223Z retirado: /usr/share/miniconda/pkgs/pygments-2.20.0-py314h06a4308_0/info/test/tests/support/python_lexer.py
evals	Retirar Python del runner	2026-09-15T17:16:00.2506144Z retirado: /usr/share/miniconda/pkgs/pygments-2.20.0-py314h06a4308_0/info/test/tests/examplefiles/make/python25-bsd.mak.output
evals	Retirar Python del runner	2026-09-15T17:16:00.2639148Z retirado: /usr/share/miniconda/pkgs/pygments-2.20.0-py314h06a4308_0/info/test/tests/examplefiles/make/python25-bsd.mak
evals	Retirar Python del runner	2026-09-15T17:16:00.2772125Z retirado: /usr/share/miniconda/pkgs/conda-pypi-0.11.0-py314h06a4308_0/lib/python3.14/site-packages/conda_pypi/__pycache__/python_paths.cpython-314.pyc
evals	Retirar Python del runner	2026-09-15T17:16:00.2904262Z retirado: /usr/share/miniconda/pkgs/conda-pypi-0.11.0-py314h06a4308_0/lib/python3.14/site-packages/conda_pypi/python_paths.py
evals	Retirar Python del runner	2026-09-15T17:16:00.3037497Z retirado: /usr/share/miniconda/pkgs/python-installer-1.0.1-py314h06a4308_0.conda
evals	Retirar Python del runner	2026-09-15T17:16:00.3170077Z retirado: /usr/share/miniconda/pkgs/python_abi-3.14-4_cp314.conda
evals	Retirar Python del runner	2026-09-15T17:16:00.3426757Z retirado: /usr/share/miniconda
evals	Retirar Python del runner	2026-09-15T17:16:00.8973811Z retirado: /usr/share/nano/python.nanorc
evals	Retirar Python del runner	2026-09-15T17:16:00.9108336Z retirado: /usr/share/lintian/overrides/python3-debian
evals	Retirar Python del runner	2026-09-15T17:16:00.9238838Z retirado: /usr/share/lintian/overrides/python3.12-venv
evals	Retirar Python del runner	2026-09-15T17:16:00.9372557Z retirado: /usr/share/lintian/overrides/python3-dbus
evals	Retirar Python del runner	2026-09-15T17:16:00.9505812Z retirado: /usr/share/lintian/overrides/libpython3.12t64
evals	Retirar Python del runner	2026-09-15T17:16:00.9637388Z retirado: /usr/share/lintian/overrides/libpython3.12-dev
evals	Retirar Python del runner	2026-09-15T17:16:00.9771946Z retirado: /usr/share/lintian/overrides/python3-pip
evals	Retirar Python del runner	2026-09-15T17:16:00.9903807Z retirado: /usr/share/lintian/overrides/libpython3.12-minimal
evals	Retirar Python del runner	2026-09-15T17:16:01.0034341Z retirado: /usr/share/lintian/overrides/python3.12-minimal
evals	Retirar Python del runner	2026-09-15T17:16:01.0165639Z retirado: /usr/share/lintian/overrides/python3.12
evals	Retirar Python del runner	2026-09-15T17:16:01.0298314Z retirado: /usr/share/lintian/overrides/libpython3.12-stdlib
evals	Retirar Python del runner	2026-09-15T17:16:01.0430321Z retirado: /usr/share/lintian/overrides/python3-netaddr
evals	Retirar Python del runner	2026-09-15T17:16:01.0559756Z retirado: /usr/share/lintian/overrides/python3
evals	Retirar Python del runner	2026-09-15T17:16:01.0690562Z retirado: /usr/share/lintian/overrides/python3-apt
evals	Retirar Python del runner	2026-09-15T17:16:01.0820718Z retirado: /usr/share/automake-1.16/am/python.am
evals	Retirar Python del runner	2026-09-15T17:16:01.0951201Z retirado: /usr/share/python3/bcep/python3-jinja2
evals	Retirar Python del runner	2026-09-15T17:16:01.1080148Z retirado: /usr/share/python3/dist/python3-cryptography
evals	Retirar Python del runner	2026-09-15T17:16:01.1209333Z retirado: /usr/share/python3/dist/python3-zope.interface
evals	Retirar Python del runner	2026-09-15T17:16:01.1341060Z retirado: /usr/share/python3/dist/python3-six
evals	Retirar Python del runner	2026-09-15T17:16:01.1475744Z retirado: /usr/share/python3/dist/python3-pyasn1
evals	Retirar Python del runner	2026-09-15T17:16:01.1606881Z retirado: /usr/share/python3/python.mk
evals	Retirar Python del runner	2026-09-15T17:16:01.1737229Z retirado: /usr/sbin/pythongc-bpfcc
evals	Retirar Python del runner	2026-09-15T17:16:01.1874390Z retirado: /usr/sbin/pythoncalls-bpfcc
evals	Retirar Python del runner	2026-09-15T17:16:01.2006470Z retirado: /usr/sbin/pythonstat-bpfcc
evals	Retirar Python del runner	2026-09-15T17:16:01.2134576Z retirado: /usr/sbin/pythonflow-bpfcc
evals	Retirar Python del runner	2026-09-15T17:16:01.2398261Z retirado: /usr/bin/python3.12-config
evals	Retirar Python del runner	2026-09-15T17:16:01.2711017Z retirado: /usr/bin/python3-config
evals	Retirar Python del runner	2026-09-15T17:16:01.2961377Z retirado: /usr/bin/python3.12
evals	Retirar Python del runner	2026-09-15T17:16:01.3261343Z retirado: /usr/bin/python3
evals	Retirar Python del runner	2026-09-15T17:16:01.3566307Z retirado: /usr/bin/python
evals	Retirar Python del runner	2026-09-15T17:16:04.4520281Z búsqueda tras retirar: ninguno
evals	Retirar Python del runner	2026-09-15T17:16:04.4522118Z --- fin de la retirada de Python ---
código de la sexta orden: 0
`````
