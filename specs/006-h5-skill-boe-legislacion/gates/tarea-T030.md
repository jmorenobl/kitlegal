# T030 · intento 6 (el tercero que cuenta el workflow, 2026-09-15): sin marcar

Cabeza `a99295f50716010b560c1998ac4e71e62cd24da4` (`feat(H5): T043`), propuesta de cambio
[#27](https://github.com/jmorenobl/kitlegal/pull/27), la misma de los intentos 1 a 5. Este intento es el sexto de la
tarea y el tercero que cuenta `gates/tareas-intentos.json` (`T030: 3`), porque la supervisión del run restauró el
contador tras los intentos 3, 4 y 5 (abajo, «Historial de intentos»). Los dos motivos del intento 5 están resueltos y
comprobados en la plataforma: **con T042 y T043 las diez positivas leen el índice y el bloque que nombra la pregunta a
la primera** (la 03, la 06 y la 08 piden `a22`, `a17` y `a20`; ninguna positiva pide un bloque fuera de lo grabado) **y
la 09 cita sola en su línea, tras la transcripción, en la forma exacta de T042**. **Dos motivos nuevos, de dos tipos
distintos**: **(e)** la sesión 09, terminada con código 0 y no cortada, es `sesión ilegible` porque la línea 1 del
fichero de un hilo de su invocación es `clone(… <unfinished ...>) = ?`, la forma con la que `strace` escribe una llamada
en curso cuando el proceso termina antes de que tenga resultado, que ninguna sonda de research vio y que `LeerTrazas` no
admite: un defecto del job, el primero que el supuesto S4 no cubría; y **(f)** la 08 leyó `a20` con 0 y citó
`[art. 20.1 de la LTAIBG, BOE-A-2013-12887, bloque a20]`, con la forma legible dentro de los corchetes, la tercera
sesión distinta en tres intentos que la pone dentro pese a T040 y T042. La línea de T030 detiene la tarea sin marcarla
ante cualquier defecto que la prueba de red descubra, y el arreglo va en tareas nuevas antes de T030: **T044** (de
datos: la línea real en las trazas sintéticas), **T045** (`LeerTrazas` admite la llamada que el fin del proceso deja
sin resultado) y **T046** (la parte mecánica de la cita admite la forma legible dentro de los corchetes), más la
revisión de esta última por la persona antes del intento siguiente («Lo que decide la persona»).

## Lo que sí quedó hecho

1. `git push -u origin h5-skill-boe-legislacion`: avance rápido `a40d16a..a99295f` (el gancho `pre-push` la deja pasar).
2. `gh pr view h5-skill-boe-legislacion`: #27 existe (`OPEN`, base `main`); **no se creó ninguna propuesta**. Tras
   actualizar *Pendientes* de `gates/pr-h5.md`, el cuerpo se sincronizó con `gh pr edit 27 --body-file` (el mismo
   fichero).
3. Checks, `gh run list` y check-runs y estados por la API: `gates/evidencia-plataforma.md`. `ci` en verde
   (34961223411, 4 m 27 s); los cuatro estados de Codecov en verde, con las cifras de los intentos 2 a 5 (T042 y T043 no
   tocan ningún fichero Go).
4. Quickstart §12.1: el secreto `CLAUDE_CODE_OAUTH_TOKEN` y las etiquetas `evals` y `evals-prueba-de-red` existen.
5. Quickstart §12.2, tal cual, las siete órdenes: ejecución 34961757559 identificada por el último evento `labeled`
   (`2026-09-15T11:09:25Z`) y `workflowName` `evals`; el paso «Retirar Python del runner» termina con 0 (las mismas 921
   rutas de los intentos 3 a 5, 2 m 21 s, `búsqueda tras retirar: ninguno`); el job llega al informe (quinta y sexta
   órdenes con código 0, informe e informe JSON enteros entre marcas) y el veredicto es `fallo` por (e) y (f). La cuarta
   orden terminó a la primera (la ejecución duró 8 m 26 s, dentro del tope de la herramienta de la sesión) y la séptima
   también, sin ningún error de la plataforma. Todo en `gates/prueba-de-red.md`, con la evidencia de S2, S4, S7, S9 y
   S12.
6. S8: lo comprobable con la propuesta abierta sobre la cabeza nueva, en `gates/pr-h5.md` (*Pendientes*): 369 ficheros
   en la propuesta, bajo `.agents/` solo `.agents/.gitattributes`, la misma lectura de lenguajes.
7. `make ci` local sobre la cabeza, en primer plano de la sesión (log en `gates/ci.log`, ignorado): abajo, «Verificación
   local».

## Motivo (e) · La llamada que el fin del proceso deja sin resultado (09; del job)

`gates/prueba-de-red.md` §3.3, §4 y §5 tienen la evidencia completa. La sesión terminó con código 0 (`result success`,
respuesta con el artículo 140 y la cita `art. 140 de la Constitución Española [BOE-A-1978-31229, bloque a140]`, la
forma fija), así que no está cortada y `LeerTrazas` exige que cada línea tenga una de las formas de data-model §9, regla
5. En `traza/t.14465`, línea 1, encontró
`clone(child_stack=0x2a559d472000, flags=CLONE_VM|CLONE_FS|CLONE_FILES|CLONE_SIGHAND|CLONE_THREAD|CLONE_SYSVSEM|CLONE_SETTLS <unfinished ...>) = ?`
y dio la traza por ilegible con el motivo exacto del informe (`no es ninguna de las formas de línea de la traza: execve,
clone, clone3, fork, vfork o connect con su resultado, una señal o la línea final`).

- **Qué es la línea.** Las banderas son las del runtime de Go en amd64 (`CLONE_THREAD` y `CLONE_SETTLS`, research V51 y
  V53): `t.14465` es un hilo del binario `kitlegal` de una de las dos invocaciones de la sesión (el índice y `a140`, las
  dos con respuesta), creado por otro hilo del proceso, cuya primera llamada trazada fue crear a su vez otro hilo. Y
  ` <unfinished ...>) = ?` es la forma con la que `strace` cierra una llamada que estaba en curso cuando el proceso
  terminó sin llegar a su parada de salida: el binario, que termina en cuanto escribe su respuesta, salió
  (`exit_group` desde otro hilo) mientras el runtime creaba un hilo, y el núcleo mató ese hilo con el proceso.
- **Por qué no se vio antes.** Es una carrera de todo proceso corto de Go, no de esta sesión ni de este runner: puede
  tocar a cualquier invocación en cualquier ejecución, y no tocó a ninguna de las 59 invocaciones leídas en los intentos
  4 y 5 ni a las otras 22 de este. Research V53 y V54 la daban por ausente («ninguna línea `<unfinished …>`»), porque
  sus sondas trazaban procesos que terminan por su hilo principal; la única forma sin resultado que el lector admite es
  `? ERRNO (descripción)`, y solo en una sesión cortada (regla 6), que es otra cosa: la señal del tope interrumpe una
  llamada bloqueante.
- **Lo que no se sabe de la traza.** Solo se conoce la línea 1 de `t.14465`, porque el lector se detiene en el primer
  defecto y el job no publica las trazas. Con toda probabilidad la sigue `+++ exited with 0 +++` (el fin de cada hilo de
  un proceso que sale con `exit_group` es esa línea, V54 (2)), y puede haber además un fichero `t.<n>` del hilo que esa
  `clone` llegó a crear, si el núcleo lo creó antes de matarlo, con solo su línea final y sin ninguna línea que lo cree
  (la `clone` que lo creó no tiene resultado): con la regla 1 de hoy sería un segundo fichero sin origen y la traza
  seguiría siendo ilegible. Las dos cosas las fijan T044 y T045, y la sonda nueva V65 las comprueba en un contenedor.
- **Por qué es un defecto del job y no del modelo.** La sesión hizo todo lo que la eval espera (leyó el índice y `a140`
  con 0 y citó en la forma fija); es el lector el que no admite lo que `strace` escribe. Sin este arreglo, la ejecución
  de cierre (10 de 10, FR-082) depende de que la carrera no toque a ninguna de las 22 invocaciones de la siguiente
  ejecución, y el job semanal fallaría de cuando en cuando sin que nada hubiera cambiado.

## Motivo (f) · La forma legible dentro de los corchetes, tercera vez (08; el modelo frente a la forma fija)

La sesión leyó `a20` con 0 (`comandos_ausentes` vacío) y respondió con el plazo, la ampliación y el silencio, cada
punto con su cita, pero las tres citas llevan la forma legible dentro de los corchetes, delante del identificador:
`[art. 20.1 de la LTAIBG, BOE-A-2013-12887, bloque a20]` (dos veces) y `[art. 20.4 de la LTAIBG, BOE-A-2013-12887,
bloque a20]`. La expresión de extracción (`formaDeCita`, `internal/evals/citas.go`) exige el identificador justo tras el
corchete, así que no cuenta ninguna y la cita esperada queda ausente.

- **La serie.** Es la tercera sesión distinta en tres intentos con la forma legible dentro de los corchetes: la 10 del
  intento 4 (`[Real Decreto Legislativo 2/2015, BOE-A-2015-11430, bloque a38]`), la 09 del intento 5
  (`[Constitución Española, BOE-A-1978-31229, bloque a140]`, sola tras la transcripción) y ahora la 08, con el artículo,
  el apartado y la sigla de la norma. Cada una con una etiqueta distinta, y cada una después de un refuerzo: T040
  prohibió cualquier texto dentro de los corchetes con dos ejemplos; T042 escribió la forma mecánica en el propio paso 5
  y pidió comprobar antes de responder que cada corchete de apertura va seguido de `BOE-`. En esta ejecución la 09
  cumplió exactamente lo que T042 escribió y la 02 puso la etiqueta fuera del corchete mecánico pero dentro de otro
  (`[art. 118.2, LCSP [BOE-A-2017-12902, bloque a1-30]]`, seis veces; la extracción la lee por el corchete interior):
  el modelo de las sesiones quiere una etiqueta legible pegada a la cita y, con el protocolo reforzado dos veces, sigue
  metiéndola dentro en una sesión de cada diez.
- **Lo que dice de la forma fija.** Esas citas llevan todo lo que FR-008 exige de una cita —la norma en forma legible,
  su identificador `BOE-A-…` y el id del bloque tal como los da la fuente, extraíbles mecánicamente, y el contenido del
  texto que devolvió `scripts/boe`— y SC-009 quiere comparar por identificador y no por redacción. Lo único que las deja
  fuera es dónde va la etiqueta legible respecto al corchete, una restricción que solo sirve a la expresión de
  extracción: para quien lee la respuesta, `[art. 20.1 de la LTAIBG, BOE-A-2013-12887, bloque a20]` es una cita
  completa, y `cita` (H8) la resolverá igual. Con el modelo fijado por la clarificación del spec (D13), la ejecución de
  cierre con 10 de 10 (FR-082, SC-003) no es alcanzable de forma fiable mientras la extracción rechace esa forma: con una
  sesión de cada diez, diez positivas seguidas pasan poco más de una de cada tres veces.

## Arreglo elegido: T044, T045 y T046

**T044**, de datos, cinco casos nuevos de `TestLeerTrazas` en `internal/evals/testdata/sesiones/leer-trazas/`, con la
línea real de `t.14465` (con los números de hilo del caso): `clone-sin-terminar` (la `clone` sin terminar como línea 1
del fichero de un hilo de la invocación, seguida de `+++ exited with 0 +++`), `hilo-de-clone-sin-terminar` (lo mismo,
más el fichero del hilo que esa `clone` pudo crear, con solo `+++ exited with 0 +++` y sin ninguna línea que lo cree),
`connect-sin-terminar` (la misma forma en la otra llamada del filtro que puede quedar en curso), y dos negativos,
`huerfano-sin-clone-sin-terminar` (un fichero sin origen y sin llamadas cuando ninguna creación quedó sin terminar:
sigue ilegible) y `clone-sin-terminar-seguido-de-otra-llamada` (a una llamada sin resultado solo pueden seguirla líneas
de señal y la línea final: ilegible). Validados con un arnés temporal, como T038: con el lector de hoy, los cuatro
primeros fallan con el motivo exacto del runner. Sin pausa (ficheros nuevos bajo el directorio de datos de prueba del
paquete, research V36). Once ficheros; el árbol de §9.1 pasa a 55 casos y 208 ficheros.

**T045**, `internal/evals/trazas.go` y su test, y nada más: (1) `formaDeLlamada` admite `<llamada>(<argumentos>
<unfinished ...>) = ?` en cualquier sesión, cortada o no —no la produce el corte sino el fin del proceso—, como llamada
sin resultado a la que solo pueden seguir líneas de señal y la línea final (el mismo mecanismo que `? ERRNO (…)`, que
sigue admitiéndose solo con la sesión cortada); no crea ningún hilo, una `execve` así no es una `execve` con 0 y un
`connect` así se clasifica por su dirección (`red` fuera del bucle local, FR-076); (2) la regla 1 de data-model §9
admite, además del fichero raíz (el único sin línea de creación que tiene alguna llamada), un fichero sin línea de
creación **solo si** no tiene ninguna llamada y alguna creación de la traza quedó sin terminar: es el hilo que esa
llamada pudo crear y que el núcleo mató con el proceso antes de ninguna llamada trazada; no pertenece a ningún proceso
ni tiene conexiones que atribuir. Todo lo demás sigue siendo ilegible con su error. Con data-model §9, el contrato del
job §4 y §9, research V53, V54, S4 y una sonda nueva, V65, en un contenedor desechable con `strace` 6.8: un programa de
Go que sale con `os.Exit` mientras su runtime crea hilos, repetido hasta provocar la línea, que registra qué la sigue y
si aparece el fichero huérfano, y sobre cuya traza real se lee `LeerTrazas` con el cambio y se rechaza sin él.

**Alternativas rechazadas para (e)**: ignorar las líneas que no casan (dejaría hilos y conexiones sin atribuir en
silencio: `red` vacío en falso, FR-076); admitir `<unfinished ...>` solo con la sesión cortada (no es el corte lo que la
produce: esta sesión terminó con 0); publicar las trazas como artefacto del job para leerlas después (no arregla la
lectura, y las trazas llevan el argv de cada orden de la sesión); cambiar la orden de la sesión para que el binario no
salga mientras crea hilos (no existe tal opción: es el runtime de Go).

**T046**, `internal/evals/citas.go`, su test y `skills/boe-legislacion/SKILL.md`: `formaDeCita` extrae, de cada pareja
de corchetes, el `<identificador>, bloque <id>` con el que termina, con cualquier texto sin corchetes delante del
identificador; la forma fija de FR-008 pasa a ser «los corchetes terminan en `<identificador>, bloque <id>]`», la
comparación sigue siendo la igualdad exacta de la pareja y ninguna cita esperada cambia. `SKILL.md` conserva sus cinco
pasos, sus cinco reglas y la forma recomendada (la forma legible delante del corchete), y sustituye la prohibición de
texto dentro de los corchetes, los dos ejemplos que no valen y la comprobación de `BOE-` tras el corchete por la regla
nueva. Con el contrato de la skill §2.3 y §3, el de evals §6, data-model §6.2, research D23 (párrafo nuevo) y
`gates/pr-h5.md`. Sin comprobación local con un modelo (FR-044): la evidencia es el intento siguiente de T030.

**Por qué revertir lo que D23 rechazó dos veces.** D23 rechazó «admitir texto delante del identificador dentro de los
corchetes en la extracción» porque «cambia la forma fija de FR-008 para tolerar un desvío del protocolo, que es lo que
la comparación mecánica tiene que detectar». Lo que la comparación tiene que detectar, según el spec, es otra cosa:
SC-009 (otro bloque u otra norma fallan; la redacción no cuenta), FR-072 (el comando esperado con código 0 y la cita
esperada) y FR-076 (ninguna petición a la red). Dónde va la etiqueta legible respecto al corchete no es nada de eso: es
una convención de escritura que el modelo de las sesiones no sigue de forma fiable, y la evidencia de tres intentos (tres
sesiones, tres etiquetas distintas, dos refuerzos del protocolo entre medias) dice que un cuarto refuerzo no lo
eliminará. Es la misma razón por la que la persona decidió que las positivas nombren el artículo (T041 y T043): que las
evals midan lo que el job puede medir —la norma, el bloque y la cita por su identificador— sin depender de lo que el
modelo hace de forma no fiable. Y la extracción sigue siendo mecánica e inequívoca: `BOE-A-` seguido de cuatro cifras y
un número no aparece en ningún otro sitio de una cita.

**Alternativas rechazadas para (f)**: reforzar `SKILL.md` por cuarta vez, con el apartado y la sigla como ejemplos de lo
que va delante del corchete (no elimina lo que dos refuerzos no eliminaron, cuesta un intento de plataforma por cada
variante nueva y deja el cierre de 10 de 10 al albur de la sesión; es la opción (A) de «Lo que decide la persona»);
dejar la extracción y repetir la prueba de red (D23); cambiar el modelo de las sesiones (reabre la clarificación del
spec, D13); exigir una línea final de fuente en toda respuesta (rechazada en T042: añade una regla que el contrato §2.4
no tiene y no evita la forma inválida dentro de esa línea).

## Lo que decide la persona antes del intento siguiente

1. **T046, como está escrita (B) o la opción (A)**. (B): la extracción admite la forma legible dentro de los corchetes,
   delante del identificador, y `SKILL.md` deja de prohibirla; revierte lo que research D23 rechazó dos veces, con la
   evidencia de los intentos 4, 5 y 6. (A): un cuarto refuerzo de `SKILL.md` (el artículo con apartado y la sigla de la
   norma como ejemplos de lo que va delante del corchete) sin tocar la extracción, y repetir la prueba de red. Se escribe
   ya como tarea (B) y no como pregunta por la misma razón que T043: es la decisión que la persona ya tomó para las
   preguntas, aplicada a la cita. Si la persona prefiere (A), basta reescribir la línea de T046 en `tasks.md` antes de
   lanzar el run (y con ello el cierre de 10 de 10 seguirá dependiendo de dónde ponga el modelo la etiqueta en cada
   sesión).
2. **El contador de intentos**: cuando el workflow llegue a T030 tras T044, T045 y T046, el contador (`T030: 3`) lo
   detendrá. Si se concede otro intento, es el séptimo de la tarea.

## Lo que la ejecución sí confirma del job

Lo mismo que los intentos 4 y 5 en doce de las trece sesiones: S4 se cumple en ellas (las doce trazas legibles, con la
`vfork()` con relleno, las `execve` enteras, la atribución de la conexión `local` de `a9998` a su invocación y ninguna
de clase `red`); S2, S7 y S10 como en los intentos 3 a 5 (la búsqueda de la retirada tardó 1 m 52 s en este runner);
S9 en las trece (entre 6 y 32 s); S12 en (1), (2), (3), (5) y (6), con (4) sin ejercer; S1, S5, S6 y S11 en lo
ejercido. Las 22 invocaciones del binario que las doce trazas atribuyen tienen código (20 con 0, 1 con 5 y 1 con 4) y
conexiones coherentes con él (la de 5, solo `127.0.0.1:9` `local`; las demás, ninguna). Y lo nuevo: T042 funciona en
la plataforma (la 09 cita en la forma fija tras la transcripción) y T043 también (las seis positivas que cambió leen el
bloque nombrado a la primera; ninguna positiva pide un bloque fuera de lo grabado, por primera vez).

## Lo que hay que mirar primero en el intento siguiente

1. **Ninguna `sesión ilegible`**: las trece trazas leídas, con sus invocaciones en el informe; si alguna sesión vuelve
   a ser ilegible, el fichero, la línea y el texto del motivo (y si es un fichero sin origen, T045 (2)).
2. **Las citas**: con T046 (B), toda cita que termine en `<identificador>, bloque <id>]` cuenta; mirar si alguna sesión
   cita de una forma que tampoco case (sin `bloque`, sin corchetes, con el identificador fuera).
3. Lo de siempre: `Modelos de las sesiones` y `Versiones de Claude Code` con valor, ninguna sesión con código 124 ni
   137, ninguna petición de clase `red`, y las dos filas de `a9998` con 5 y 4.

## Tareas nuevas

T044, T045 y T046 van tras T043 y antes de la fase 12, en ese orden (T044 solo toca once ficheros nuevos bajo el
directorio de datos de prueba de `internal/evals`; T045, `trazas.go` y su test; T046, `citas.go`, su test y la skill).
T033 a T043 siguen marcadas. En `tasks.md`: el formato (nueve tareas de datos, con T044), las rebanadas (T044 entre las
de datos y T045 entre las de código que las leen), la fila 2 y la fila 10 de la Definition of Done, y en la
trazabilidad las filas de FR-004 a FR-012 y FR-015 (T046), FR-071 (T045), FR-072 (T046), FR-076 (T044 y T045), SC-004
(T046), SC-009 (T046) y SC-012 (las tres); y en «Dependencias y orden», T033 a T046 antes de T030 y T044 antes de T045.
Ninguna de las tres lleva etiqueta de plataforma; T044 lleva la de datos y declara sus once ficheros por su ruta
completa; T045 declara `internal/evals/trazas.go` e `internal/evals/trazas_test.go`; T046 declara
`internal/evals/citas.go`, `internal/evals/citas_test.go` y `skills/boe-legislacion/SKILL.md`; las tres, ficheros del
directorio del feature (y `CHANGELOG.md`, siempre permitido).

## Qué hace el intento siguiente de T030

Con T044, T045 y T046 en verde: `git push -u origin h5-skill-boe-legislacion` (avance rápido); `gh pr view` encuentra
#27 y no se crea nada; checks y check-runs de la cabeza nueva (T045 y T046 tocan ficheros Go: `codecov/patch` mide de
nuevo); quickstart §12.1 y §12.2 enteras (la etiqueta está quitada); y reescribe `gates/evidencia-plataforma.md` y
`gates/prueba-de-red.md`. Como `gates/pr-h5.md` cambiará con T045 y T046, conviene volver a sincronizar el cuerpo de
#27 con `gh pr edit 27 --body-file specs/006-h5-skill-boe-legislacion/gates/pr-h5.md`. Si la ejecución dura más de
600 s, la cuarta orden se repite tal cual en primer plano (este intento: 8 m 26 s; el 5, 13 m 33 s).

## Historial de intentos

| Intento | Cabeza | Lo que descubrió la plataforma | Arreglo |
|---|---|---|---|
| 1 | `6d68c21` | `codecov/patch` en rojo; la purga de paquetes de Python rompía `apt` (100) | T034, T035; T033 |
| 2 | `417635e` | `CLAUDE_CODE_SUBPROCESS_ENV_SCRUB=1` exigía `bubblewrap` y forzaba el modo de permisos; `strace -s 4096` cortaba el argv | T036; T037 |
| 3 | `857ec46` | `vfork()` con relleno de alineación en x86_64 dejaba ilegibles las trece trazas | T038, T039 |
| 4 | `537e5d6` | las trazas se leen; seis positivas no pasan por el protocolo frente a lo grabado (a, b, d) y por el artículo elegido (c) | T040; T041 (decisión de la persona) |
| 5 | `a40d16a` | T040 y T041 funcionan (02, 05, 07 y 10 pasan); cuatro positivas no pasan: 03, 06 y 08 por el artículo elegido (c'), 09 por la cita sola tras una transcripción (d') | T042; T043 (misma decisión, extendida) |
| 6 | `a99295f` | T042 y T043 funcionan (las diez leen el bloque nombrado; la 09 cita en la forma fija); la 09 es ilegible por `clone(… <unfinished ...>) = ?` (e); la 08 cita con la forma legible dentro de los corchetes (f) | T044, T045; T046 (revisión de la persona) |

Tras el intento 3, `siguiente_tarea` detuvo el run al llegar T030 a 4 intentos. El supervisor del hito (sesión que
vigila el run por encargo de Jorge) restauró el contador a 2 en `gates/tareas-intentos.json` para permitir **un** intento
más, el 4, porque cada intento había descubierto en la plataforma un defecto distinto y real del job, con su arreglo en
una tarea propia, y ninguno repitió un fallo anterior. Tras el intento 4 volvió a restaurarlo a 2 para el 5, con el
mismo criterio y la decisión de la persona sobre las evals 05 y 07 aplicada en T041; y tras el 5 (T042 y T043, con la
decisión de la persona sobre el alcance de T043: las seis positivas por materia), otra vez a 2 para este, el 6. No se ha
tocado ningún veredicto, umbral ni control. Este intento no repite ningún fallo anterior (lo que T042 y T043 arreglaban
está comprobado en verde); lo que descubre es una forma de `strace` que ninguna sonda vio, en una sesión que hizo todo
lo que la eval espera, y una tercera variante de la forma de cita inválida.

**Decisión de la persona sobre el alcance de T043 (supervisión del run, 2026-09-15).** Jorge eligió **las seis**
positivas que preguntan por materia (02, 03, 04, 06, 08 y 10), como está escrita T043: las evals miden el protocolo
(índice → id → bloque → cita) sin depender de la memoria del modelo, con la misma razón que T041; la 02 sigue exigiendo
el índice de hecho (artículo 118 → `a1-30`, que solo sale de él), aunque, como aclaró la supervisión tras revisar el
commit de T043, ya no lleva `indice` entre sus comandos esperados, y la búsqueda de un artículo por su materia queda en
el backlog (`docs/USO.md`). Este intento lo confirma: las seis pasan a la primera.

**Tercera ampliación del tope (supervisión del run, 2026-09-15).** Tras T042 y T043 el workflow volvió a detenerse por
el tope de intentos de T030. La supervisión restauró el contador a 2 para un intento más, el sexto de la tarea, con el
mismo criterio: el intento 5 cumplió la prueba de red (SC-012), no repitió ningún fallo anterior y los motivos que dejó
quedaban arreglados en T042 y T043. Este es ese intento.

## Verificación local

`make ci` sobre el árbol de la cabeza `a99295f` (este intento solo escribe en `gates/` y en `tasks.md`, y borró antes
los ficheros temporales que usó para copiar el informe, la retirada de Python, el registro y la lista de ficheros de la
propuesta), en primer plano de la sesión, con la salida en `gates/ci.log` (ignorado): **código 0**, `ci: todos los controles en verde`, `0 issues.`, `No
vulnerabilities found.`, ningún `FAIL` y `go mod tidy -diff` sin cambios. La tarea
queda `[ ]`, como manda su línea cuando la prueba de red descubre un defecto; el árbol deja los cuatro ficheros de
`gates/` de este intento y `tasks.md` con T044, T045 y T046 (y su trazabilidad), y ningún fichero temporal. La etiqueta
`evals-prueba-de-red` está quitada y el cuerpo de #27 sincronizado con `gates/pr-h5.md`.
