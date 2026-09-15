# T030 · intento 2 de 3 (2026-09-15): sin marcar

Cabeza `417635e39abf5807f8600b79debd6026fc320ce8` (`feat(H5): T035`), propuesta de cambio
[#27](https://github.com/jmorenobl/kitlegal/pull/27), la misma del intento 1. Los dos motivos del intento 1 están
resueltos: `codecov/patch` en verde (98,42 % del diff, objetivo 94,70 %) con T034 y T035, y el paso «Retirar Python
del runner» termina con 0 con T033. **Un motivo nuevo**, que solo podía descubrir esta prueba de red: **ninguna sesión
de Claude Code arranca en el runner**, porque `CLAUDE_CODE_SUBPROCESS_ENV_SCRUB=1` exige `bubblewrap` en Linux, y la
imagen no lo trae. Su arreglo va en una tarea nueva, **T036**, colocada antes de T030 en `tasks.md` (tras T035). T030
no puede llevarlo: sus rutas son de `gates/`, y el push del intento siguiente tiene que publicar la cabeza con el
arreglo.

El intento 1 está documentado en la versión anterior de este fichero (commit `c4c7613`, `feat(H5): T033`, y
anteriores); lo que sigue lo sustituye.

## Lo que sí quedó hecho

1. `git push -u origin h5-skill-boe-legislacion`: avance rápido `6d68c21..417635e` (el gancho `pre-push` la deja pasar).
2. `gh pr view h5-skill-boe-legislacion`: #27 existe (`OPEN`, base `main`); **no se creó ninguna propuesta**. Tras
   actualizar *Pendientes* y la cobertura de `gates/pr-h5.md`, el cuerpo se sincronizó con `gh pr edit 27 --body-file`
   (el mismo fichero).
3. Checks, `gh run list` y check-runs y estados por la API: `gates/evidencia-plataforma.md`. `ci` en verde
   (34929854868, 4 m 49 s); los cuatro estados de Codecov en verde, con cifra y objetivo, `codecov/patch` incluido.
4. Quickstart §12.1: el secreto `CLAUDE_CODE_OAUTH_TOKEN` y las etiquetas `evals` y `evals-prueba-de-red` existen.
5. Quickstart §12.2, tal cual, las siete órdenes: ejecución 34930222593 identificada por el último evento `labeled`
   (`2026-09-15T04:48:02Z`) y `workflowName` `evals`; el paso «Retirar Python del runner» **termina con 0** (921
   rutas retiradas, `búsqueda tras retirar: ninguno`, 1 m 47 s); el job llega al informe (quinta y sexta órdenes con
   código 0, informe e informe JSON enteros entre marcas), pero el veredicto es `fallo`: **las trece sesiones
   terminan con código 1 al arrancar**, sin ningún mensaje en el transcript. Todo en `gates/prueba-de-red.md`, con la
   evidencia de S2, S4, S7, S9 y S12 y la etiqueta retirada con la séptima orden.
6. S8: lo comprobable con la propuesta abierta sobre la cabeza nueva, en `gates/pr-h5.md` (*Pendientes*).
7. `make ci` local sobre la cabeza, en primer plano de la sesión (log en `gates/ci.log`, ignorado): código 0,
   `ci: todos los controles en verde`, ningún `FAIL`.

## Motivo · las sesiones no arrancan: `bubblewrap` y el modo de permisos con `CLAUDE_CODE_SUBPROCESS_ENV_SCRUB`

**Hecho** (`gates/prueba-de-red.md` §3 y anexo A): las trece sesiones (doce evals y la de prueba de red) tienen
`codigo_de_la_sesion` 1, `fin_de_la_sesion` `sin mensajes`, respuesta vacía, ninguna invocación y esta salida de
error, idéntica en todas (tras la cabecera de licencia y una línea de `import` del propio binario):

```text
error: bubblewrap is required for subprocess env scrubbing and isolation. Install with: sudo apt-get install -y bubblewrap, set sandbox.bwrapPath in managed settings, or set CLAUDE_CODE_SUBPROCESS_ENV_SCRUB=0 to disable (loses subprocess isolation).
      at nBn (/$bunfs/root/chunk-y58z5vzc.js:11:25510)
      at A (/$bunfs/root/chunk-xgzajagg.js:11:3535)
      at async <anonymous> (/$bunfs/root/chunk-mpjnf6jh.js:88:4824)
Bun v1.4.3 (Linux x64)
```

Claude Code 2.1.270 en Linux, con `CLAUDE_CODE_SUBPROCESS_ENV_SCRUB` activo, exige `bwrap` (bubblewrap) al arrancar,
antes de escribir el mensaje `system`/`init` y antes de ninguna petición al modelo; `ubuntu-24.04` no lo trae y el
job no lo instala. Cada sesión duró menos de un segundo; `timeout` y `strace` devolvieron el código de `claude` (1).
El informe lo recoge bien: cada eval no pasa con `la sesión no terminó: código 1` y, en las positivas, activación,
comandos y citas ausentes; `fuera_de_lo_grabado` y `red` vacíos; `sin_python` con sus tres líneas; ninguna sesión
`ilegible`.

**Por qué el diseño no lo previó**: research V12 comprobó `CLAUDE_CODE_SUBPROCESS_ENV_SCRUB` leyendo el binario
2.1.270 (que retira del entorno de los subprocesos las variables con credenciales), no ejecutando una sesión en
Linux, y la exigencia de `bwrap` es solo del binario de Linux (el de macOS de esta máquina, la misma versión, no
contiene el mensaje; V11 sí anotaba que el *sandbox* de Bash exige `bwrap` y `socat`, y D16 lo rechazó por eso, sin
que nadie supiera que el `ENV_SCRUB` lo exige también).

**Segundo efecto de la misma variable, leído del binario 2.1.270 de esta máquina** (funciones `Ekn` y `cNn`, las que
resuelven el modo de permisos; búsqueda por desplazamiento de byte, sin ejecutar nada):

```js
if(gd()){let V=r||v&&v!=="default"||C&&C!=="default"||m&&m!=="default",re="Permission mode forced to default — CLAUDE_CODE_SUBPROCESS_ENV_SCRUB is set "+"(allowed_non_write_users hardening). Declare allowedTools explicitly, or set CLAUDE_CODE_SUBPROCESS_ENV_SCRUB=0 to opt out.";if(V)process.stderr.write(`⚠ ${re}\n`);return{mode:"default",notification:V?re:void 0,modeSuppliedOnInvocation:!0}}
```

y, en la versión pura de la misma función, `if(Ie(n.CLAUDE_CODE_SUBPROCESS_ENV_SCRUB)){…return{mode:"default",…}}`.
Es decir: con la variable activa, **el modo de permisos se fuerza a `default`** e ignora `--permission-mode
bypassPermissions` (con un aviso `⚠ Permission mode forced to default …` en la salida de error), y las herramientas
que la sesión necesita hay que declararlas con `--allowedTools`. En una sesión `-p` sin terminal, en modo `default`,
una llamada a Bash que ninguna regla permite se deniega: **instalar `bubblewrap` solo haría que las sesiones
arrancaran y que ninguna ejecutara el binario**, y el intento 3 (el último) volvería a fallar, esta vez con
`comando ausente` y `cita ausente` en las diez positivas. Las dos cosas van juntas en T036.

**Alternativas para el arreglo**:

- **Instalar `bubblewrap` en el runner y declarar las herramientas permitidas** (elegida, **T036**): es lo que el
  propio binario pide («Install with: sudo apt-get install -y bubblewrap» y «Declare allowedTools explicitly») y
  conserva la decisión del plan (§VII: la credencial del modelo no llega a las órdenes de la sesión) y la garantía de
  red de D16 (proxy que rechaza, observado por invocación). Cambia el paso de instalación del job (`bubblewrap` junto
  a `strace`) y la orden de la sesión en `scripts/evals.sh` (`--allowedTools Skill Bash` en lugar de
  `--permission-mode bypassPermissions`, que el binario ignora con la variable activa). Lo que el aislamiento de
  `bwrap` hace con los subprocesos (qué espacios de nombres desune, qué monta de solo lectura, si la caché de la
  sesión y el enlace de la skill siguen accesibles, si `strace` sigue a los descendientes a través de `bwrap` y con
  qué líneas `clone`) no se puede comprobar con una sesión real fuera de la plataforma; T036 lo observa hasta donde
  se puede sin modelo, en contenedores desechables de Linux con un token inválido, y lo que quede lo dice la prueba
  de red del intento 3.
- *`CLAUDE_CODE_SUBPROCESS_ENV_SCRUB=0`* (o sin la variable): las sesiones arrancarían sin `bwrap` y
  `bypassPermissions` se respetaría, pero la credencial del modelo entraría en el entorno de cada orden de la sesión
  (del binario y de lo que el modelo ejecute), en contra del plan §VII; rechazada mientras la elegida sea viable. Si
  la comprobación de T036 mostrara que el aislamiento de `bwrap` es incompatible con el contrato del job (proxy,
  caché, enlace de la skill o traza), T036 lo documenta y toma esta con su justificación, actualizando plan §VII.
- *Ajuste `processWrapper` de Claude Code* para anteponer `env -u CLAUDE_CODE_OAUTH_TOKEN` a cada subproceso: sin
  documentar, solo honrado desde ajustes gestionados (lista `wo` del binario) y sin comprobar; rechazada.
- *Sandbox de Bash de Claude Code*: sigue rechazado por D16 (cambia el contexto de ejecución de la skill y la
  petición permitida la haría el proxy interno, invisible por invocación).

**Lo que la ejecución sí confirma del job** (antes del arranque de `claude`): el paso de retirada de Python sin
purga (T033) funciona en la imagen real, con `find` como root en 1 m 23 s y sin ninguna ruta en solo lectura; la
comprobación 3 del guion repite la búsqueda y escribe `resultado: ninguno`; las comprobaciones 4 a 6 pasan; las trece
preparaciones de sesión (`TestPrepararSesion`) pasan con las grabaciones; `timeout`, `strace` y el guion propagan el
código de `claude` y el informe se escribe e imprime entre marcas aunque el veredicto sea `fallo`.

## Tarea nueva

T036 va tras T035 y antes de la fase 12. T033, T034 y T035 siguen marcadas. En la trazabilidad de `tasks.md`, FR-070
y SC-012 nombran T036, y el punto 10 de la Definition of Done, también. La línea de T036 no lleva ninguna etiqueta:
la extracción del workflow da `.github/workflows/evals.yml`, `scripts/evals.sh` y ficheros del feature.

## Qué hace el intento siguiente de T030

Con T036 en verde: `git push -u origin h5-skill-boe-legislacion` (avance rápido); `gh pr view` encuentra #27 y no se
crea nada; checks y check-runs de la cabeza nueva; quickstart §12.1 y §12.2 enteras (la etiqueta está quitada, así
que la primera orden no quita nada y la segunda crea un evento `labeled` nuevo); y reescribe
`gates/evidencia-plataforma.md` y `gates/prueba-de-red.md`. Como `gates/pr-h5.md` cambiará con T036, conviene volver a
sincronizar el cuerpo de #27 con `gh pr edit 27 --body-file specs/006-h5-skill-boe-legislacion/gates/pr-h5.md`. Lo
que hay que mirar primero en el informe: `Modelos de las sesiones` y `Versiones de Claude Code` con valor (la sesión
arrancó), ninguna línea `⚠ Permission mode forced` ni `permission` en las salidas de error, y las invocaciones de
`a9998` en `fuera_de_lo_grabado`.

## Actualización tras T036 (2026-09-15)

T036 comprobó en contenedores de `ubuntu:24.04` la alternativa elegida arriba y no basta: con `bubblewrap` y
`--allowedTools Skill Bash` las sesiones arrancan, pero la primera orden de Bash exige también `socat` y, con él, se
ejecuta dentro de `bwrap` con `--unshare-pid`, cuyos `clone` devuelven números del espacio de nombres nuevo que
`LeerTrazas` no puede atribuir; además deja el disco de solo lectura (research V61 (3)). Como prevé su línea, T036 tomó
`CLAUDE_CODE_SUBPROCESS_ENV_SCRUB=0` y conserva `--permission-mode bypassPermissions` (research V12, V61 y D13; plan
§VII): el job no instala `bubblewrap`, y Claude Code sigue sin pasar el token al entorno de las órdenes. La misma
comprobación encontró un defecto que no depende de la variable: `strace -s 4096` corta el argv de la instantánea de
shell que Claude Code crea antes de la primera orden de Bash, y toda sesión que ejecute Bash saldría con `sesión
ilegible`; lo arregla **T037**, también antes de T030.

Lo que hay que mirar primero en el informe del intento 3, en lugar de lo anterior: `Modelos de las sesiones` y
`Versiones de Claude Code` con valor; salidas de error sin `bubblewrap`, `socat`, `Sandbox is required` ni `Permission
mode forced`; ninguna sesión con `sesión ilegible` (si la hay, el fichero, la línea y su texto, para S4); y las dos
invocaciones de `a9998` en `fuera_de_lo_grabado`, la de `--json` con 5 y la conexión `127.0.0.1:9` `local`, y la de
`--offline` con 4 y sin conexiones, como en research V61 (4).
