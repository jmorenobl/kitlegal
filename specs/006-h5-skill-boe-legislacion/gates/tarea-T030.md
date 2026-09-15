# T030 · intento 4 (el tercero que cuenta el workflow, 2026-09-15): sin marcar

Cabeza `537e5d6f7ed5ff48f17f343323f0cef026cb9e33` (`feat(H5): T039`), propuesta de cambio
[#27](https://github.com/jmorenobl/kitlegal/pull/27), la misma de los intentos 1 a 3. Este intento es el cuarto de la
tarea y el tercero que cuenta `gates/tareas-intentos.json` (`T030: 3`), porque la supervisión del run restauró el
contador tras el intento 3 (abajo, «Historial de intentos»). El motivo del intento 3 está resuelto: **con T038 y T039
las trece trazas se leen enteras**, y con ellas la prueba de red cumple todo lo que SC-012 espera (las dos filas de
`a9998` con 5 y 4, la conexión `127.0.0.1:9` de clase `local` atribuida a la invocación sin `--offline`, ninguna sesión
ilegible ni cortada, ninguna petición a la red de una fuente, y la sesión de prueba de red igual que la 01). **Un motivo
nuevo, que solo podía descubrir esta prueba de red con las trazas leídas: el veredicto es `fallo` porque seis de las
diez positivas no pasan** (02, 03, 05, 07 y 08 por `comando ausente` y `cita ausente`; 10 solo por `cita ausente`), por
lo que el modelo hizo frente a lo grabado y no por el job, el lector ni el informe. La línea de T030 detiene la tarea
sin marcarla ante cualquier defecto que la prueba de red descubra, y el arreglo va en una tarea nueva antes de T030:
**T040** (el protocolo de la skill), más una decisión que solo puede tomar una persona (las evals 05 y 07).

## Lo que sí quedó hecho

1. `git push -u origin h5-skill-boe-legislacion`: avance rápido `857ec46..537e5d6` (el gancho `pre-push` la deja pasar).
2. `gh pr view h5-skill-boe-legislacion`: #27 existe (`OPEN`, base `main`); **no se creó ninguna propuesta**. Tras
   actualizar *Pendientes* de `gates/pr-h5.md`, el cuerpo se sincronizó con `gh pr edit 27 --body-file` (el mismo
   fichero).
3. Checks, `gh run list` y check-runs y estados por la API: `gates/evidencia-plataforma.md`. `ci` en verde
   (34941036417, 4 m 44 s); los cuatro estados de Codecov en verde, con las cifras de los intentos 2 y 3 (T039 no añade
   ninguna línea sin cubrir).
4. Quickstart §12.1: el secreto `CLAUDE_CODE_OAUTH_TOKEN` y las etiquetas `evals` y `evals-prueba-de-red` existen.
5. Quickstart §12.2, tal cual, las siete órdenes: ejecución 34941499481 identificada por el último evento `labeled`
   (`2026-09-15T07:23:53Z`) y `workflowName` `evals`; el paso «Retirar Python del runner» termina con 0 (921 rutas, las
   mismas del intento 3, 2 m 55 s, `búsqueda tras retirar: ninguno`); el job llega al informe (quinta y sexta órdenes
   con código 0, informe e informe JSON enteros entre marcas) y el veredicto es `fallo` por las seis positivas. Todo en
   `gates/prueba-de-red.md`, con la evidencia de S2, S4, S7, S9 y S12 y la etiqueta retirada con la séptima orden.
6. S8: lo comprobable con la propuesta abierta sobre la cabeza nueva, en `gates/pr-h5.md` (*Pendientes*): 368 ficheros
   en la propuesta, bajo `.agents/` solo `.agents/.gitattributes`, la misma lectura de lenguajes.
7. `make ci` local sobre la cabeza, en primer plano de la sesión (log en `gates/ci.log`, ignorado): abajo, «Verificación
   local».

## Motivo · el modelo frente a lo grabado: seis positivas sin pasar

`gates/prueba-de-red.md` §3.3 y §4 tienen la evidencia completa (invocaciones con código y conexiones de las trece
sesiones, respuestas, medidas locales). Resumen por causa:

- **(a) `articulos` es todo o nada y la regla 2 hace que el modelo se rinda** (03, 08; la 06 se recuperó sola). El
  modelo leyó el índice, acertó el bloque esperado y lo pidió con `scripts/boe articulos` junto a un vecino que no está
  grabado; el verbo falla en el primer bloque que no tiene (comprobado en local con la caché preparada: `a22 a21` termina
  con 4 y `no se ha podido resolver el bloque a21, en la posición 2 de 2`, y `articulo a22` con 0), la invocación entera
  terminó con 5 y el modelo dijo que no pudo consultar (regla 2) sin pedir el bloque esperado por separado; la 03 repitió
  el vecino, no el esperado.
- **(b) Ids compuestos del número del artículo** (02): pidió `a117` y `a118`, que no existen en el índice de la Ley
  9/2017 (sus ids son `a1-2`, `a1-3`… y el artículo 118 es `a1-30`), contra el paso 3 del protocolo; y en el job un id
  inexistente no devuelve 3 sino 5, así que la vuelta al índice «ante un código 3» no se activa.
- **(c) Otro artículo** (05: `a2` en vez de `a59`; 07: `a140` a `a145` en vez de `a25`). Los ids existen; el error es de
  conocimiento del modelo, y el índice de la fuente no ayuda: sus títulos son solo «Artículo N», «TÍTULO I», «CAPÍTULO
  III», sin rúbrica (comprobado en las diez normas de las evals), así que sirve para pasar del número al id, no para
  encontrar un artículo por su materia. Con red se podría explorar; en el job todo bloque no grabado responde 5.
- **(d) El nombre de la norma dentro de los corchetes** (10): leyó `a38` con 0 y citó `[Real Decreto Legislativo 2/2015,
  BOE-A-2015-11430, bloque a38]`; la forma fija es `[<identificador>, bloque <id>]` y el extractor, como debe, no la
  reconoce (SC-009). El SKILL.md prohíbe «art. 21» y «artículo 21» dentro de los corchetes, pero no dice nada del nombre
  ni del rango de la norma.

**Por qué no se vio antes**: en los intentos 1 y 2 ninguna sesión corrió o arrancó, y en el 3 ninguna traza se pudo
leer, así que este es el primer informe con invocaciones. Las respuestas del intento 3 ya apuntaban a (a) y (c) (05, 06
y 10 decían «código 5»), y la versión anterior de este fichero lo dejó como lo primero que mirar.

**Lo que dice del diseño**: es el comportamiento que FR-074 y FR-076 quieren —la caché de la sesión solo tiene lo que
los comandos esperados necesitan, el proxy no deja salir nada, y el informe nombra cada invocación fuera de lo grabado
con su eval— y lo que los evals están para medir: el protocolo, con el modelo del job, no llega al bloque esperado en
cuatro de las siete evals que no nombran el artículo. (a), (b) y (d) son del protocolo y se arreglan en él; (c) no.

## Arreglo elegido: T040

`skills/boe-legislacion/SKILL.md`, sin cambiar sus cinco pasos ni sus cinco reglas (contrato de la skill §2.3 y §2.4):

1. Paso 3: el id de un bloque se copia de la entrada del índice cuyo `titulo` es el artículo («Artículo 118» de la Ley
   9/2017 → `a1-30`) y nunca se compone del número, porque los ids de muchas normas no son `a<número>` — (b).
2. Paso 3: los bloques se leen de uno en uno con `scripts/boe articulo`; `scripts/boe articulos` solo cuando hacen falta
   varios a la vez y todos salen del índice; y si una orden con varios bloques termina con 4 o 5, se pide cada bloque por
   separado con `articulo` antes de dar ninguno por no consultado, porque el fallo de un bloque no impide leer los demás
   — (a).
3. «Cómo se cita»: dentro de los corchetes no va nada más que el identificador y el id: ni «art. 21», ni «artículo 21»,
   ni el nombre, el número o el rango de la norma, que van delante; `[Ley 39/2015, BOE-A-2015-10565, bloque a21]` no
   vale — (d).
4. Regla 2: se dice qué no se pudo consultar solo después de haber pedido cada bloque por separado.

Con el contrato de la skill (§2.3, §2.4, §3), research (una fila de verificación con las medidas locales y una decisión
D23 con las alternativas rechazadas) y `gates/pr-h5.md`. Sin nombrar evals, job, plataforma ni modelos (contrato §2.5),
con la región generada intacta y menos de 300 líneas. No hay comprobación local posible con un modelo (FR-044): la
evidencia es el intento siguiente de T030.

**Alternativas rechazadas**: grabar los bloques vecinos que pidió el modelo (persigue cada sesión y deja lo grabado sin
regla; FR-074 lo acota a lo que los comandos esperados necesitan); admitir texto delante del identificador dentro de los
corchetes en `formaDeCita` (cambia la forma fija de FR-008 y el contrato §3 para tolerar un desvío del protocolo, que es
lo que SC-009 quiere detectar); que `Juzgar` cuente un bloque leído con 4 o 5 (FR-072 y FR-076 lo prohíben); cambiar el
modelo de las sesiones (la clarificación del spec lo fija en la gama económica: decisión cerrada); y una herramienta
que encuentre un artículo por su materia dentro de una norma (es lo que la skill necesitaría para (c), pero es una
herramienta nueva, fuera del alcance de H5: backlog).

## Lo que decide la persona antes del intento siguiente

1. **Las evals 05 y 07** dependen de que el modelo sepa el número del artículo (59 del TRLRHL, 25 de la LRJSP), porque
   el índice de la fuente no tiene rúbricas y en el job no se puede explorar. Con el modelo que fija la clarificación del
   spec (gama económica, Haiku 4.5), la 05 falló en los intentos 3 y 4 y la 07 en el 4. T040 no las arregla. Opciones:
   (i) dejarlas como están y repetir la prueba de red tras T040: si el modelo acierta, pasan; si no, la ejecución de
   cierre (10 de 10, FR-082) no será alcanzable de forma fiable con ese modelo; (ii) que sus preguntas nombren el artículo
   («¿Qué dice el artículo 59 del texto refundido…?»), como hacen la 01 y la 09, en una tarea nueva que cambie las dos
   evals y la tabla del contrato de evals §2 (con lo que pasan a esperar solo el bloque, como la 01, y quedan cinco evals
   que exigen el índice: 02, 03, 06, 08 y 10); (iii) cambiar el modelo de las sesiones, que reabre la clarificación del
   spec y D13; (iv) una herramienta que encuentre artículos por materia dentro de una norma (backlog, para un hito
   posterior; es la raíz). **Recomendación**: (ii) ahora, porque mide lo que el job puede medir (el protocolo: índice →
   id → bloque → cita) sin fingir que mide conocimiento, y (iv) en el backlog, anotado en `docs/USO.md`; (iii) solo si la
   persona quiere que las evals midan también el conocimiento del modelo. No se ha hecho aquí porque cambia el diseño de
   dos evals fijado en el contrato y no es un defecto del protocolo.
2. **El contador de intentos**: cuando el workflow llegue a T030 tras T040, el contador (`T030: 3`) lo detendrá. Si se
   concede otro intento, es el quinto de la tarea.

## Lo que la ejecución sí confirma del job

Todo lo que los intentos anteriores dejaron por comprobar: S4 se cumple entera (las trece trazas legibles, con la
`vfork()` con relleno, las `execve` enteras, la atribución de la conexión `local` de `a9998` a su invocación y ninguna
de clase `red`); S2, S7 y S10 como en el intento 3; S9 en las trece (entre 6 y 38 s); S12 en (1), (2), (3), (5) y (6),
con (4) sin ejercer; S1, S5, S6 y S11 en lo ejercido. Las 26 invocaciones del binario que las trazas atribuyen tienen
código (0, 2, 4 o 5) y conexiones coherentes con él (las de 5, solo `127.0.0.1:9` `local`; las de 0 y 4, ninguna).

## Lo que hay que mirar primero en el intento siguiente

1. **Las sesiones 02, 03 y 08**: si con T040 leen su bloque por separado tras el 5 (o de entrada) y citan en la forma
   fija. Si la 02 vuelve a pedir `a118`, el modelo no está usando la entrada «Artículo 118» del índice: entonces habría
   que saber qué le entregó la herramienta Bash (la salida del índice de la LCSP tiene 34 720 octetos y `a1-30` está en
   el octeto 9 376), y el job no conserva los transcripts; se anotaría como supuesto nuevo.
2. **La sesión 10**: la cita con la forma fija.
3. **Las sesiones 05 y 07**: qué bloque piden; si vuelven a fallar por el artículo, es la decisión 1 de arriba.
4. Lo de siempre: ninguna `sesión ilegible`, `Modelos de las sesiones` y `Versiones de Claude Code` con valor, ninguna
   sesión con código 124 ni 137, ninguna petición de clase `red`, y las dos filas de `a9998` con 5 y 4.

## Tareas nuevas

T040 va tras T039 y antes de la fase 12. T033 a T039 siguen marcadas. En la trazabilidad de `tasks.md`, la fila de
FR-004 a FR-012 y FR-015, la de SC-004 y la de SC-012 nombran T040, y el punto 10 de la Definition of Done también. La
línea de T040 no lleva etiqueta de datos ni de plataforma y declara `skills/boe-legislacion/SKILL.md` y ficheros del
directorio del feature (más `CHANGELOG.md`, siempre permitido).

## Qué hace el intento siguiente de T030

Con T040 en verde (y, si la persona elige la opción (ii), la tarea de las evals 05 y 07 también antes de T030):
`git push -u origin h5-skill-boe-legislacion` (avance rápido); `gh pr view` encuentra #27 y no se crea nada; checks y
check-runs de la cabeza nueva; quickstart §12.1 y §12.2 enteras (la etiqueta está quitada); y reescribe
`gates/evidencia-plataforma.md` y `gates/prueba-de-red.md`. Como `gates/pr-h5.md` cambiará con T040, conviene volver a
sincronizar el cuerpo de #27 con `gh pr edit 27 --body-file specs/006-h5-skill-boe-legislacion/gates/pr-h5.md`.

## Historial de intentos

| Intento | Cabeza | Lo que descubrió la plataforma | Arreglo |
|---|---|---|---|
| 1 | `6d68c21` | `codecov/patch` en rojo; la purga de paquetes de Python rompía `apt` (100) | T034, T035; T033 |
| 2 | `417635e` | `CLAUDE_CODE_SUBPROCESS_ENV_SCRUB=1` exigía `bubblewrap` y forzaba el modo de permisos; `strace -s 4096` cortaba el argv | T036; T037 |
| 3 | `857ec46` | `vfork()` con relleno de alineación en x86_64 dejaba ilegibles las trece trazas | T038, T039 |
| 4 | `537e5d6` | las trazas se leen; seis positivas no pasan por el protocolo frente a lo grabado (a, b, d) y por el artículo elegido (c) | T040; decisión de la persona |

Tras el intento 3, `siguiente_tarea` detuvo el run al llegar T030 a 4 intentos. El supervisor del hito (sesión que
vigila el run por encargo de Jorge) restauró el contador a 2 en `gates/tareas-intentos.json` para permitir **un** intento
más, este, porque cada intento había descubierto en la plataforma un defecto distinto y real del job, con su arreglo en
una tarea propia, y ninguno repitió un fallo anterior. No se ha tocado ningún veredicto, umbral ni control.

## Verificación local

`make ci` sobre el árbol de la cabeza `537e5d6` (este intento solo escribe en `gates/` y en `tasks.md`, y borra antes
los ficheros temporales que usó para leer el registro y medir los índices), en primer plano de la sesión, con la salida
en `gates/ci.log` (ignorado): **código 0**, `ci: todos los controles en verde`, ningún `FAIL` y `go mod tidy -diff`
sin cambios. La tarea queda `[ ]`, como manda su línea cuando la prueba de red descubre un defecto; el árbol deja los
cuatro ficheros de `gates/` de este intento y `tasks.md` con T040, y ningún fichero temporal.
