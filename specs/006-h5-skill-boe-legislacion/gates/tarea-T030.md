# T030 · intento 5 (el tercero que cuenta el workflow, 2026-09-15): sin marcar

Cabeza `a40d16a6d2bd4b4159bc6bf2788f77f20f3c5f4a` (`feat(H5): T041`), propuesta de cambio
[#27](https://github.com/jmorenobl/kitlegal/pull/27), la misma de los intentos 1 a 4. Este intento es el quinto de la
tarea y el tercero que cuenta `gates/tareas-intentos.json` (`T030: 3`), porque la supervisión del run restauró el
contador tras los intentos 3 y 4 (abajo, «Historial de intentos»). Los dos motivos del intento 4 están resueltos y
comprobados en la plataforma: **con T040 y T041 pasan las cuatro positivas que entonces fallaron por el protocolo o por
el artículo elegido** (02, 05, 07 y 10; la 02 por el camino exacto que T040 escribió), la prueba de red vuelve a cumplir
todo lo que SC-012 espera y el job, el lector de trazas y el informe no tienen ningún defecto. **Un motivo nuevo, del
mismo tipo que el anterior: el veredicto es `fallo` porque cuatro positivas no pasan** —03, 06 y 08 por `comando
ausente` y `cita ausente`, porque el modelo pidió otro artículo, tres sesiones que en el intento 4 acertaron; y 09 por
`cita ausente`, porque citó con el nombre de la norma dentro de los corchetes, la forma que T040 declaró inválida—, por
lo que el modelo hizo frente a lo grabado y no por el job. La línea de T030 detiene la tarea sin marcarla ante cualquier
defecto que la prueba de red descubra, y el arreglo va en tareas nuevas antes de T030: **T042** (la cita cuando va sola
y la forma mecánica en el paso 5 de la skill) y **T043** (las seis positivas que preguntan por materia nombran el
artículo, como la persona decidió para la 05 y la 07), más la revisión de esa segunda por la persona antes del intento
siguiente («Lo que decide la persona»).

## Lo que sí quedó hecho

1. `git push -u origin h5-skill-boe-legislacion`: avance rápido `537e5d6..a40d16a` (el gancho `pre-push` la deja pasar).
2. `gh pr view h5-skill-boe-legislacion`: #27 existe (`OPEN`, base `main`); **no se creó ninguna propuesta**. Tras
   actualizar *Pendientes* de `gates/pr-h5.md`, el cuerpo se sincronizó con `gh pr edit 27 --body-file` (el mismo
   fichero).
3. Checks, `gh run list` y check-runs y estados por la API: `gates/evidencia-plataforma.md`. `ci` en verde
   (34956091100, 4 m 27 s); los cuatro estados de Codecov en verde, con las cifras de los intentos 2 a 4 (T040 y T041 no
   tocan ningún fichero Go).
4. Quickstart §12.1: el secreto `CLAUDE_CODE_OAUTH_TOKEN` y las etiquetas `evals` y `evals-prueba-de-red` existen.
5. Quickstart §12.2, tal cual, las siete órdenes: ejecución 34956596912 identificada por el último evento `labeled`
   (`2026-09-15T10:11:15Z`) y `workflowName` `evals`; el paso «Retirar Python del runner» termina con 0 (las mismas 921
   rutas de los intentos 3 y 4, 6 m 32 s, `búsqueda tras retirar: ninguno`); el job llega al informe (quinta y sexta
   órdenes con código 0, informe e informe JSON enteros entre marcas) y el veredicto es `fallo` por las cuatro positivas.
   La cuarta orden se repitió tal cual (la ejecución duró 13 m 33 s, más que el tope de la herramienta de la sesión) y la
   séptima también (la primera ejecución quitó la etiqueta y después recibió un `502 Bad Gateway` de la plataforma).
   Todo en `gates/prueba-de-red.md`, con la evidencia de S2, S4, S7, S9 y S12.
6. S8: lo comprobable con la propuesta abierta sobre la cabeza nueva, en `gates/pr-h5.md` (*Pendientes*): 369 ficheros
   en la propuesta, bajo `.agents/` solo `.agents/.gitattributes`, la misma lectura de lenguajes.
7. `make ci` local sobre la cabeza, en primer plano de la sesión (log en `gates/ci.log`, ignorado): abajo, «Verificación
   local».

## Motivo · el modelo frente a lo grabado, segunda vuelta: cuatro positivas sin pasar

`gates/prueba-de-red.md` §3.3 y §4 tienen la evidencia completa (invocaciones con código y conexiones de las trece
sesiones, respuestas). Resumen por causa:

- **(c') Otro artículo, en tres sesiones que en el intento 4 acertaron** (03, 06, 08). Las tres leyeron el índice con 0
  y pidieron un solo bloque, como manda el paso 3 desde T040, pero no el esperado: la 03, `a21` (atribuciones del
  Alcalde) en vez de `a22` (del Pleno), con un reintento `--offline` (4); la 06, `a21` (rendimientos del capital) en vez
  de `a17` (del trabajo), con reintentos `--timeout 10000` (2) y `--timeout 10s` (5); la 08, `a12` (derecho de acceso)
  dos veces en vez de `a20` (plazo de resolución). Ante el 5 se rindieron sin suplir el texto (regla 2, cumplida). En
  el intento 4 las tres pidieron el bloque esperado (la 03, `a21 a22`; la 06, `a17 a18 a19 a20` y después `a17`; la 08,
  `a19 a20`): **el número del artículo que el modelo de las sesiones recuerda cambia de una sesión a otra**, y leer de
  uno en uno (T040) no lo compensa: la 03 del intento 4 pidió los dos candidatos a la vez; esta vez eligió uno y falló.
  Es la causa (c) del intento 4 en la 05 y la 07: el índice de la fuente solo da «Artículo N» sin rúbrica (V64 (1)),
  en el job todo bloque no grabado responde 5, y no hay redacción del protocolo que supla lo que el modelo no sabe (D23).
- **(d') El nombre de la norma dentro de los corchetes, otra vez** (09). Leyó `a140` con 0, transcribió el artículo
  como cita textual en bloque y, debajo, sola en su línea, escribió `[Constitución Española, BOE-A-1978-31229, bloque
  a140]`. «Cómo se cita» lo prohíbe desde T040 con `[Ley 39/2015, BOE-A-2015-10565, bloque a21]` como forma que no vale,
  y el paso 5 solo remite a esa sección. En la misma ejecución la 05 y la 10 pusieron el nombre de la norma delante de
  los corchetes y la 09 del intento 4 citó bien: la regla se sigue cuando la cita acompaña a una frase y se rompe cuando
  va sola tras una transcripción, que es cuando el modelo quiere una etiqueta legible y la mete dentro. La expresión de
  extracción exige el identificador justo tras el corchete, como debe (SC-009; D23 rechazó relajarla).

**Lo que dice del diseño**: T040 hizo lo suyo en las tres causas que arreglaba —ninguna sesión compuso un id, ninguna
se rindió tras un `articulos` fallido sin pedir por separado, y la 10 citó en la forma fija—, y T041 lo suyo en la 05 y
la 07. Lo que queda es lo que D23 ya había identificado como no arreglable en el protocolo, la memoria del modelo sobre
el número del artículo, en tres evals más; y una segunda ocasión de la forma de cita inválida en una posición (cita sola
tras una transcripción) que el texto de T040 no cubría con un ejemplo.

## Arreglo elegido: T042 y T043

**T042**, `skills/boe-legislacion/SKILL.md`, sin cambiar sus cinco pasos, sus cinco reglas ni la forma de la cita
(contrato de la skill §2.3, §2.4 y §3): el paso 5 escribe la forma mecánica en el propio paso —`[<identificador>,
bloque <id>]`, con el corchete de apertura seguido inmediatamente de `BOE-A-…`— en vez de solo remitir a «Cómo se cita»,
y pide comprobar antes de responder que cada corchete de apertura de una cita va seguido de `BOE-`; «Cómo se cita» dice
que la regla vale igual cuando la cita va sola en una línea o tras una cita textual (la forma legible va delante en la
misma línea: `art. 140 de la Constitución Española [BOE-A-1978-31229, bloque a140]`) y añade
`[Constitución Española, BOE-A-1978-31229, bloque a140]` como segunda forma que no vale. Con el contrato de la skill
§2.3 y §3, research D23 (párrafo nuevo con esta evidencia y las alternativas rechazadas) y `gates/pr-h5.md`
(*Decisiones*). Sin nombrar evals, job, plataforma ni modelos (contrato §2.5), con la región generada intacta y menos de
300 líneas. Sin comprobación local con un modelo (FR-044): la evidencia es el intento siguiente de T030.

**Alternativas rechazadas para (d')**: admitir texto delante del identificador dentro de los corchetes en la extracción
(rechazada ya en D23: cambia la forma fija de FR-008 para tolerar un desvío que SC-009 quiere detectar); exigir una
línea final de fuente en toda respuesta (añade una regla que el contrato §2.4 no tiene, y no evita que dentro de esa
línea la forma sea la inválida); mover «Cómo se cita» delante del protocolo (el orden de secciones lo fija el contrato
§2.2, FR-002).

**T043**, las seis positivas que preguntan por materia (02, 03, 04, 06, 08 y 10) nombran el artículo, como la 01, la 05,
la 07 y la 09, y sus comandos esperados quedan en el bloque solo; bloques, citas, identificadores, materias y lo grabado
no cambian; la 06 conserva `reproduce: boe-fiscal` (sigue siendo el artículo 17 de la LIRPF, `a17`, la consulta
documentada en `refs/boe.py`, FR-064); la 02, con «artículo 118» en la pregunta, sigue exigiendo el índice de hecho,
porque su id `a1-30` solo sale de él. Con el contrato de evals §2 (seis filas y la nota de comandos esperados), research
D23 («Las evals 05 y 07» pasa a cubrir las diez), `gates/pr-h5.md` (*Decisiones*) y la entrada de `docs/USO.md` del
2026-09-15 (evidencia nueva de 03, 06 y 08). Sin comprobación local con un modelo: la evidencia es el intento siguiente.

**Por qué las seis y no solo las tres que fallaron.** La decisión de la persona para la 05 y la 07 (T041, opción (ii))
se justificó así: que las evals midan «lo que el job puede medir (el protocolo: índice → id → bloque → cita) sin fingir
que mide conocimiento», y rechazó dejarlas y repetir porque, con el modelo fijado, «la ejecución de cierre (10 de 10,
FR-082) no será alcanzable de forma fiable». D23 dejó las otras seis por materia con una condición explícita: «en ese
intento el modelo eligió en todas el artículo esperado». Ese razonamiento se aplica ahora a las seis, no a tres: la
condición ya no se cumple en la 03, la 06 y la 08 (aciertan en el intento 4 y fallan en el 5), y en la 02, la 04 y la 10
dos aciertos en dos intentos son la misma dependencia con mejor suerte, no una garantía; cambiar solo las tres que
fallaron dejaría el cierre de 10 de 10 y el job semanal al albur de la sesión en las otras tres, que es lo que la
decisión quiso evitar. Las diez positivas siguen siendo de materias distintas (FR-062), incluyen LPAC, LCSP, LRBRL, LGT y
TRLRHL, la del art. 21 y las fiscales no cambian (data-model §6.3), y lo que no miden —encontrar el artículo por su
materia— sigue en el backlog (`docs/USO.md`, T041). **Alternativas rechazadas para (c')**: cambiar solo 03, 06 y 08
(arriba); dejarlas y repetir la prueba de red (D23); cambiar el modelo de las sesiones (reabre la clarificación del
spec, D13); grabar los bloques que el modelo pide por error (persigue cada sesión; FR-074); y una herramienta que
encuentre un artículo por su materia dentro de una norma (backlog, `docs/USO.md`).

## Lo que decide la persona antes del intento siguiente

1. **El alcance de T043**: seis evals (como está escrita) o solo las tres que fallaron (03, 06 y 08). Es la misma
   decisión que la persona tomó para la 05 y la 07, extendida con la evidencia de este intento, y por eso se escribe ya
   como tarea y no como pregunta; si la persona prefiere conservar por materia la 02, la 04 y la 10, basta editar la
   línea de T043 en `tasks.md` antes de lanzar el run (y con ello el cierre de 10 de 10 seguirá dependiendo de la memoria
   del modelo en esas tres).
2. **El contador de intentos**: cuando el workflow llegue a T030 tras T042 y T043, el contador (`T030: 3`) lo detendrá.
   Si se concede otro intento, es el sexto de la tarea.

## Lo que la ejecución sí confirma del job

Lo mismo que el intento 4, por segunda vez: S4 se cumple entera (las trece trazas legibles, con la `vfork()` con
relleno, las `execve` enteras, la atribución de la conexión `local` de `a9998` a su invocación y ninguna de clase
`red`); S2, S7 y S10 como en los intentos 3 y 4 (la búsqueda de la retirada tardó 6 m 7 s en este runner, frente a 2 m
32 s en el 4: el tiempo lo pone el disco); S9 en las trece (entre 6 y 38 s); S12 en (1), (2), (3), (5) y (6), con (4)
sin ejercer; S1, S5, S6 y S11 en lo ejercido. Las 33 invocaciones del binario que las trazas atribuyen tienen código
(22 con 0, 8 con 5, 2 con 4 y 1 con 2) y conexiones coherentes con él (las de 5, solo `127.0.0.1:9` `local`; las demás,
ninguna). Y lo nuevo: T040 funciona en la plataforma (la 02 pide `a1-30` por separado tras el 5 de `articulos` y pasa;
la 10 cita en la forma fija) y T041 también (la 05 y la 07 pasan a la primera).

## Lo que hay que mirar primero en el intento siguiente

1. **La sesión 09**: la cita con la forma fija tras la transcripción en bloque; y cualquier otra sesión que ponga una
   cita sola en su línea.
2. **Las seis de T043** (02, 03, 04, 06, 08 y 10): que lean el índice y el bloque nombrado a la primera, como la 05 y
   la 07 en este intento; la 02, que copie `a1-30` del índice con el artículo ya en la pregunta.
3. Lo de siempre: ninguna `sesión ilegible`, `Modelos de las sesiones` y `Versiones de Claude Code` con valor, ninguna
   sesión con código 124 ni 137, ninguna petición de clase `red`, y las dos filas de `a9998` con 5 y 4.

## Tareas nuevas

T042 y T043 van tras T041 y antes de la fase 12, en ese orden (T042 toca solo la skill; T043 solo las evals y
`docs/USO.md`). T033 a T041 siguen marcadas. En la trazabilidad de `tasks.md`, la fila de FR-004 a FR-012 y FR-015 y
la de SC-004 nombran T042; las de FR-062, FR-063, FR-064, FR-072 y SC-003 nombran T043; la de SC-012 nombra las dos; y el
punto 10 de la Definition of Done también. Ninguna de las dos lleva etiqueta de datos ni de plataforma; T042 declara
`skills/boe-legislacion/SKILL.md` y T043 los seis ficheros de `evals/boe-legislacion/` y `docs/USO.md`, más ficheros del
directorio del feature (y `CHANGELOG.md`, siempre permitido).

## Qué hace el intento siguiente de T030

Con T042 y T043 en verde: `git push -u origin h5-skill-boe-legislacion` (avance rápido); `gh pr view` encuentra #27 y no
se crea nada; checks y check-runs de la cabeza nueva; quickstart §12.1 y §12.2 enteras (la etiqueta está quitada); y
reescribe `gates/evidencia-plataforma.md` y `gates/prueba-de-red.md`. Como `gates/pr-h5.md` cambiará con T042 y T043,
conviene volver a sincronizar el cuerpo de #27 con `gh pr edit 27 --body-file
specs/006-h5-skill-boe-legislacion/gates/pr-h5.md`. Si la ejecución dura más de 600 s, la cuarta orden se repite tal
cual en primer plano (este intento: 13 m 33 s, porque la búsqueda de la retirada de Python tardó 6 m 7 s).

## Historial de intentos

| Intento | Cabeza | Lo que descubrió la plataforma | Arreglo |
|---|---|---|---|
| 1 | `6d68c21` | `codecov/patch` en rojo; la purga de paquetes de Python rompía `apt` (100) | T034, T035; T033 |
| 2 | `417635e` | `CLAUDE_CODE_SUBPROCESS_ENV_SCRUB=1` exigía `bubblewrap` y forzaba el modo de permisos; `strace -s 4096` cortaba el argv | T036; T037 |
| 3 | `857ec46` | `vfork()` con relleno de alineación en x86_64 dejaba ilegibles las trece trazas | T038, T039 |
| 4 | `537e5d6` | las trazas se leen; seis positivas no pasan por el protocolo frente a lo grabado (a, b, d) y por el artículo elegido (c) | T040; T041 (decisión de la persona) |
| 5 | `a40d16a` | T040 y T041 funcionan (02, 05, 07 y 10 pasan); cuatro positivas no pasan: 03, 06 y 08 por el artículo elegido (c'), 09 por la cita sola tras una transcripción (d') | T042; T043 (misma decisión, extendida) |

Tras el intento 3, `siguiente_tarea` detuvo el run al llegar T030 a 4 intentos. El supervisor del hito (sesión que
vigila el run por encargo de Jorge) restauró el contador a 2 en `gates/tareas-intentos.json` para permitir **un** intento
más, el 4, porque cada intento había descubierto en la plataforma un defecto distinto y real del job, con su arreglo en
una tarea propia, y ninguno repitió un fallo anterior. Tras el intento 4 volvió a restaurarlo a 2 para este, el 5, con el
mismo criterio y la decisión de la persona sobre las evals 05 y 07 aplicada en T041. No se ha tocado ningún veredicto,
umbral ni control. Este intento no repite ningún fallo anterior (lo que T040 y T041 arreglaban está comprobado en verde);
lo que descubre es la misma causa (c) en otras tres evals y una segunda posición de la forma de cita inválida.

## Verificación local

`make ci` sobre el árbol de la cabeza `a40d16a` (este intento solo escribe en `gates/` y en `tasks.md`, y borra antes
los ficheros temporales que usó para copiar el informe, la retirada de Python y el registro), en primer plano de la
sesión, con la salida en `gates/ci.log` (ignorado): **código 0**, `ci: todos los controles en verde`, ningún `FAIL` y
`go mod tidy -diff` sin cambios. La tarea queda `[ ]`, como manda su línea cuando la prueba de red descubre un defecto;
el árbol deja los cuatro ficheros de `gates/` de este intento y `tasks.md` con T042 y T043 (y su trazabilidad), y ningún
fichero temporal. La etiqueta `evals-prueba-de-red` está quitada y el cuerpo de #27 sincronizado con `gates/pr-h5.md`.
