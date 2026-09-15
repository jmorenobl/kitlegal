# T030 · intento 3 de 3 (2026-09-15): sin marcar

Cabeza `857ec465074a273bd0d7c0c175185f830d3112ea` (`feat(H5): T037`), propuesta de cambio
[#27](https://github.com/jmorenobl/kitlegal/pull/27), la misma de los intentos 1 y 2. Los dos motivos del intento 2
están resueltos: con T036, las trece sesiones arrancan, se autentican, terminan con código 0 y responden (`Modelos de
las sesiones: claude-haiku-4-5-20251001`, `Versiones de Claude Code: 2.1.270`, salida de error vacía en las trece, sin
`bubblewrap`, `socat`, `Sandbox is required` ni `Permission mode forced`), y con T037 ninguna traza se declara ilegible
por la instantánea de shell. **Un motivo nuevo, que solo podía descubrir esta prueba de red: las trece sesiones salen
con `sesión ilegible` por la línea `vfork()` con relleno de alineación** que `strace` escribe en el runner de x86_64 y
que `LeerTrazas` no admite. Su arreglo va en dos tareas nuevas, **T038** (de datos) y **T039**, colocadas tras T037 y
antes de T030 en `tasks.md`. T030 no puede llevarlo: sus rutas son de `gates/`, la línea real tiene que entrar en las
trazas sintéticas en una tarea de datos (research S4; la propia línea de T030) y el push del intento siguiente tiene
que publicar la cabeza con el arreglo.

**Este era el último de los tres intentos de T030.** La tarea queda sin marcar, como manda su línea; el intento
siguiente lo concede una persona (o quien supervise el hito) después de T038 y T039, porque el contador de
`gates/tareas-intentos.json` no es de esta tarea y no se toca.

Los intentos 1 y 2 están documentados en las versiones anteriores de este fichero (commits `c4c7613` y `857ec46`); lo
que sigue lo sustituye.

## Lo que sí quedó hecho

1. `git push -u origin h5-skill-boe-legislacion`: avance rápido `417635e..857ec46` (el gancho `pre-push` la deja pasar).
2. `gh pr view h5-skill-boe-legislacion`: #27 existe (`OPEN`, base `main`); **no se creó ninguna propuesta**. Tras
   actualizar *Pendientes* y la cobertura de `gates/pr-h5.md`, el cuerpo se sincronizó con `gh pr edit 27 --body-file`
   (el mismo fichero).
3. Checks, `gh run list` y check-runs y estados por la API: `gates/evidencia-plataforma.md`. `ci` en verde
   (34935998623, 4 m 30 s); los cuatro estados de Codecov en verde, con las cifras del intento 2 (T036 y T037 no tocan
   ningún fichero Go).
4. Quickstart §12.1: el secreto `CLAUDE_CODE_OAUTH_TOKEN` y las etiquetas `evals` y `evals-prueba-de-red` existen.
5. Quickstart §12.2, tal cual, las siete órdenes: ejecución 34936425178 identificada por el último evento `labeled`
   (`2026-09-15T06:19:28Z`) y `workflowName` `evals`; el paso «Retirar Python del runner» termina con 0 (921 rutas,
   3 m 7 s, `búsqueda tras retirar: ninguno`); el job llega al informe (quinta y sexta órdenes con código 0, informe e
   informe JSON enteros entre marcas), pero el veredicto es `fallo`: **las trece sesiones son ilegibles por la misma
   línea**. Todo en `gates/prueba-de-red.md`, con la evidencia de S2, S4, S7, S9 y S12 y la etiqueta retirada con la
   séptima orden.
6. S8: lo comprobable con la propuesta abierta sobre la cabeza nueva, en `gates/pr-h5.md` (*Pendientes*).
7. `make ci` local sobre la cabeza, en primer plano de la sesión (log en `gates/ci.log`, ignorado): código 0,
   `ci: todos los controles en verde`, ningún `FAIL` (abajo, «Verificación local»).

## Motivo · `vfork()` con el relleno de alineación de `strace`

**Hecho** (`gates/prueba-de-red.md` §3.3 y anexo A): las trece sesiones tienen `codigo_de_la_sesion` 0,
`fin_de_la_sesion` `result success`, respuesta del modelo, y un único motivo, idéntico salvo el fichero, la línea y el
número:

```text
sesión ilegible: traza: traza ilegible: /home/runner/work/_temp/kitlegal-evals-boe-legislacion/sesiones/01-lpac-articulo-21/traza/t.11488, línea 7: no es ninguna de las formas de línea de la traza: execve, clone, clone3, fork, vfork o connect con su resultado, una señal o la línea final: «vfork()                                 = 11494»
```

(`t.11488` es el fichero del hilo principal de `claude`, el primero por número; en las otras doce sesiones, línea 5.)

La línea tiene 47 octetos: `vfork()` (7), 33 espacios y `= 11494` (7); el `=` está en la columna 41. Es la alineación
por defecto de `strace` (`-a 40`, «Align return values in a specific column»): tras el paréntesis de cierre escribe un
espacio, rellena hasta la columna 40 si la llamada es más corta y después `= ` y el resultado. `formaDeLlamada`
(`internal/evals/trazas.go`) exige `\) = ` con un solo espacio: casa con toda llamada de más de 40 columnas, que son
todas las que las sondas de research vieron (`execve` con su argv, `connect` con su dirección, `clone` y `clone3` con
sus banderas), y no con `vfork()`, la única del filtro `-e trace=execve,connect,clone,clone3,fork,vfork` sin
argumentos. `vfork` sí está en la lista de llamadas de la expresión y en `creaHilo`: el lector la trataría como la
creación de un proceso (data-model §9, regla 2) si la línea casara.

**Por qué no se vio antes**: V53, V54, V61 y V62 se comprobaron en contenedores arm64, donde la llamada `vfork` no existe
(la biblioteca de C la hace con `clone` y `CLONE_VFORK`, y así crean sus procesos `bash` y Claude Code en esa
arquitectura); el binario de Claude Code para Linux x86_64 —un ejecutable de Bun— crea los procesos de sus órdenes con
`vfork`. El intento 2 no llegó a esta línea porque ninguna sesión pasó del arranque. S4 preveía `fork` o `vfork` como
posible línea de creación, pero no el relleno, que es de `strace` y no de la llamada.

**Lo que dice el defecto del diseño**: es el comportamiento que data-model §9, regla 5, quiere ante un formato no
comprobado: la traza es ilegible, la sesión no pasa y el motivo nombra el fichero, la línea y su texto, en lugar de dejar
`red` vacío en silencio. La misma regla dice que esa línea real entra en las sesiones sintéticas en una tarea de datos y
que `LeerTrazas` se ajusta antes de la ejecución de cierre (research S4; plan, obligación 1).

**Arreglo elegido** (T038 y T039): `formaDeLlamada` admite entre el paréntesis de cierre y `= ` uno o más espacios (el
espacio de siempre y el relleno de alineación), sin cambiar nada más de la forma: el resultado sigue siendo obligatorio
y una llamada sin él sigue haciendo ilegible la traza. Antes del código, la línea real entra en las trazas sintéticas
como el caso `proceso-por-vfork` de `TestLeerTrazas` (T038, de datos): `t.1000`, el proceso de `claude` en x86_64, crea
el proceso 2000 con `vfork()` + 33 espacios + `= 2000`, y `t.2000` es la invocación del enlace `boe` de la skill a través de `bash`,
como en `connect-fuera-de-la-invocacion`; con el código de hoy, `LeerTrazas` falla sobre ese caso con el mismo motivo
que el runner (rojo), y con T039 devuelve la invocación (verde). T039 añade además la línea real (47 octetos, con su
número 11494) como literal de un test del paquete, y actualiza data-model §9 (regla 5: el relleno), el contrato del job
§4, research V53 (por qué sus sondas no vieron el relleno), una fila nueva de verificación, S4 y `gates/pr-h5.md`.

**Alternativas rechazadas**:

- *`strace -a 1` en la orden de la sesión*, que quitaría el relleno (con la columna en 1, ninguna llamada es más corta):
  cambia el contrato del job §3.2 y obliga a repetir en contenedores las comprobaciones de V61 y V62; deja al lector
  dependiendo de una opción para leer el formato por defecto de `strace`, que es el que reproducen todas las trazas
  sintéticas y las sondas de research; y no añade ninguna garantía, porque el relleno no pierde información.
- *Admitir en `LeerTrazas` cualquier blanco* (`\s+`): `strace` rellena con espacios; admitir tabuladores sería admitir
  una forma que nadie ha visto.
- *Ignorar las líneas que no casan*: rechazado desde el diseño (data-model §9, regla 5; FR-076): dejaría hilos sin
  atribuir y `red` vacío en falso.

## Lo que la ejecución sí confirma del job

Antes de la línea del `vfork()`, todo lo que el intento 2 dejó por comprobar: las sesiones arrancan con
`CLAUDE_CODE_SUBPROCESS_ENV_SCRUB=0` y `--permission-mode bypassPermissions` sin ningún aviso (salida de error vacía en
las trece); se autentican con el secreto y aceptan el modelo; las diez positivas activan la skill y las dos de no
activación no; Bash ejecuta el binario (las respuestas citan lo que devolvió); cada sesión termina por sí misma entre 7
y 40 s (S9), muy por debajo de los 240 s del tope, y `strace` y `timeout` devuelven su 0 (S10); el paso de retirada de
Python y la comprobación 3 dan lo mismo que en el intento 2 (S2, S7); la ejecución se identifica por la etiqueta y
`workflowName` (S12); y las cuatro líneas anteriores al `vfork()` del fichero principal de `claude` (seis en la 01) —la
`execve` y las líneas con que Claude Code de amd64 crea sus primeros hilos— tienen formas admitidas.

## Lo que hay que mirar primero en el intento siguiente

1. **Ninguna sesión con `sesión ilegible`** (si la hay, el fichero, la línea y su texto, para S4: sería otra forma de
   x86_64 que ninguna sonda vio, y su arreglo iría como T038 y T039, en una tarea de datos y otra de código). Con las
   trazas leídas, la conexión `127.0.0.1:9` de clase `local` de `a9998 --json` y las dos filas de `fuera_de_lo_grabado`
   de la sesión de prueba de red, con 5 y 4, como en research V61 (4).
2. **Las sesiones 05, 06 y 10.** En esta ejecución el modelo respondió en ellas que no pudo consultar la norma porque la
   fuente devolvió «error de límite de ritmo (código 5)», es decir, que alguna de sus invocaciones fue al proxy que
   rechaza: una consulta fuera de lo grabado. Sin traza no se sabe cuál (la 05 y la 10 esperan `indice` y después
   `articulo` de los bloques `a59` y `a38`; la 06, `indice` y `a17`; y las siete positivas restantes, con las mismas
   grabaciones, sí obtuvieron su artículo). Con las trazas leídas, esas invocaciones saldrán en `fuera_de_lo_grabado`
   con su orden: si el modelo pidió otro bloque u otro verbo, esas tres evals no pasan por `comando ausente` y `cita
   ausente`, y el arreglo —en el protocolo de la skill, en la eval o en lo grabado (tarea de datos)— iría en una tarea
   nueva antes de T031, porque la ejecución de cierre exige `aprobado` con 10 de 10; la prueba de red (SC-012) no exige
   el veredicto, pero la línea de T030 detiene la tarea ante cualquier defecto que descubra.
3. Lo mismo que en los intentos anteriores: `Modelos de las sesiones` y `Versiones de Claude Code` con valor, salidas
   de error vacías, ninguna sesión con código 124 ni 137.

## Tareas nuevas

T038 y T039 van tras T037 y antes de la fase 12. T033 a T037 siguen marcadas. En la trazabilidad de `tasks.md`, FR-076 y
SC-012 nombran T038 y T039, FR-071 nombra T039, y los puntos 2 y 10 de la Definition of Done nombran las dos. La línea de T038 lleva la
etiqueta de datos y solo declara los dos ficheros nuevos bajo el directorio de datos de prueba del paquete de evals y
ficheros del directorio del feature (sin pausa: research V36); la de T039 no lleva ninguna etiqueta y declara
`internal/evals/trazas.go`, `internal/evals/trazas_test.go` y ficheros del feature.

## Qué hace el intento siguiente de T030

Con T038 y T039 en verde: `git push -u origin h5-skill-boe-legislacion` (avance rápido); `gh pr view` encuentra #27 y no
se crea nada; checks y check-runs de la cabeza nueva (`codecov/patch` medirá las líneas nuevas de T039); quickstart
§12.1 y §12.2 enteras (la etiqueta está quitada, así que la primera orden no quita nada y la segunda crea un evento
`labeled` nuevo); y reescribe `gates/evidencia-plataforma.md` y `gates/prueba-de-red.md`. Como `gates/pr-h5.md`
cambiará con T039, conviene volver a sincronizar el cuerpo de #27 con `gh pr edit 27 --body-file
specs/006-h5-skill-boe-legislacion/gates/pr-h5.md`.

## Verificación local

`make ci` sobre el árbol de la cabeza `857ec46` (este intento solo escribe en `gates/` y en `tasks.md`), en primer
plano de la sesión, con la salida en `gates/ci.log` (ignorado): **código 0**, `ci: todos los controles en verde`,
ningún `FAIL` y `go mod tidy -diff` sin cambios. La tarea queda `[ ]`, como manda su línea cuando la prueba de red
descubre un defecto; el árbol deja los cuatro ficheros de `gates/` de este intento y `tasks.md` con T038 y T039.
