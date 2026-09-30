# 0030 · Lo que cambia el cierre vuelve a la revisión final: revisión y cierre en un solo bucle, el rango de cambios para los jueces y el informe por commit

- **Estado**: aceptada
- **Fecha**: 2026-09-30
- **Hito**: transversal (workflow `hito` 2.3.0 y constitución 2.8.0, tras H7.3 y antes de H7.4). Enmienda el ADR 0018 en
  la secuencia del final del run, el ADR 0028 en la regla de convergencia de la revisión final y el ADR 0029 en la
  sección 7 del informe final.

## Contexto y problema

En el workflow 2.2.0, después de la revisión final (`ronda_revision` → `cerrar_revision`) venían `publicar` y
`bucle_cierre` (`medir_cierre` → `reparar_cierre` → `cerrar_cierre`, hasta tres mediciones) y el informe final. Nada de
lo que hacía el cierre volvía a pasar por un juez. Mientras el cierre solo medía, no importaba: casi nunca salía en
rojo y, si salía, era por la plataforma. Desde el ADR 0029 los umbrales del job de evals deciden, así que un cierre en
rojo es normal, y `reparar_cierre` cambia código, la skill y los artefactos para ponerlo en verde.

**Lo que pasó en H7.3** (run `f41e85d1`, PR #85, fusionada en `eafa386`):

1. **El cambio.** La primera medición, sobre `6ab3add`, salió en rojo por el umbral nuevo: 5 de 51 respuestas de
   Sonnet 5 con una expresión prohibida (9,8 %; el umbral era ≤ 5 %) y la eval 06 con 1 de 3. `reparar_cierre` hizo
   `eb6b4c8` («fix(H7.3): cierre en la plataforma»), que cambió el protocolo de `skills/boe-legislacion/SKILL.md`: la
   comprobación de la redacción pasó de ir una vez por norma citada, tras la última lectura, como pedía la sección del
   hito («manteniendo lo que H7.1 decidió: una vez por norma citada, después de leer y antes de responder»), a ir en la
   misma orden que cada lectura, con `&&`. También tocó `spec.md`, `plan.md`, `research.md`, el contrato de la skill y el
   `CHANGELOG.md`, y dejó tres líneas `reparar_cierre` en `gates/supuestos.md`.
2. **Nadie lo revisó.** La segunda medición, sobre `eb6b4c8`, dio 0 de 51 y 540 s, en verde, y el run terminó. El
   informe listó el commit en «Cambios posteriores a la revisión final» con la frase «que ningún juez juzgó», entre
   registros y el propio informe: una línea más, sin nada que la destacara.
3. **El arreglo fue a mano.** Una sesión aparte relanzó los jueces con `scripts/paso.sh revision_juez_a|b H7.3`. Su
   regla de convergencia solo trataba como corrección «el último commit `fix(H7.3): motivos de la revisión final`»
   (`cerrar_correccion` y el prompt de `revision_juez`), así que la sesión tuvo que escribir una nota en
   `gates/revision-pendiente.md` que les señalara `eb6b4c8` y les diera las referencias de la sección del hito y de
   H7.1. Con ella, los dos aprobaron sin motivos en una tercera ronda (`gates/revision-{a,b}-r3.json`): juzgaron en el
   criterio i que «una por cada orden que lee bloques» era la mejor solución sin tocar el binario, y que estaba
   registrada como decisión. Como esa sesión también arregló un test inestable (`0222c39`), hubo que medir otra vez
   (1 de 51, 492 s).

El resultado fue bueno, pero por suerte y a mano. Las causas son cuatro, y ninguna es de H7.3:

- **La secuencia.** La revisión iba antes del cierre y nada volvía a ella: un cambio del producto hecho por el cierre
  podía terminar el run sin juez.
- **La convergencia de los jueces dependía de un mensaje de commit.** Lo que cambiaba otro paso no era, para ellos,
  «la corrección».
- **El informe no distinguía.** Mezclaba registros del run con cambios del producto y no decía si alguna ronda los
  había visto.
- **La medición no se ataba al producto.** Nada decía que una medición deja de valer cuando cambia lo que midió. Una
  revisión después del cierre habría podido corregir sin que nadie volviera a medir.

## Opciones consideradas

Para que lo que cambia el cierre lo vean los jueces:

1. **Juzgar al final lo acumulado**: el bucle de cierre como estaba y, después, una revisión de todo lo que cambió.
   Rechazada. Cualquier corrección de esa revisión invalida la última medición y obliga a medir otra vez, fuera de las
   tres, o a terminar con una cabeza sin medir. Además, una reparación que los jueces rechazarían se mide antes, y
   quizá se construye otra encima.
2. **Medir primero y juzgar solo lo que sale en verde.** Rechazada por lo mismo: la revisión queda al final y su
   corrección invalida la medición. Ahorra juzgar una reparación que no arregla nada, pero a costa de una medición más
   por cada rechazo (15-40 min, y una de las tres).
3. **Juzgar cada reparación antes de volver a medirla.** Elegida (abajo). La última acción del bucle es siempre una
   medición, y todo lo que mide lo ha juzgado ya alguien: la invariante se cumple en cada salida del bucle, sin
   mediciones extra.
4. **Un bucle de revisión duplicado dentro del de cierre**, con anclas YAML. Rechazada. Exige renombrar cada paso
   anidado (`preparar_jueces_cierre`, `leer_revision_cierre`…) y reescribir cada condición y la referencia de los
   items del fan-out: dos definiciones de la misma revisión que se separarán con el tiempo. Además, `scripts/paso.sh`
   y `scripts/coste-run.sh` resuelven la plantilla `revision_juez` por su id y darían con la primera de las dos.
5. **Un juez del cierre**, con una rúbrica reducida. Rechazada por la misma razón que la opción 6 del ADR 0028: la
   rúbrica es la de la revisión final. La ronda manual de H7.3 se decidió en el criterio i (alcance y supuestos), que
   una rúbrica «del cierre» no habría tenido.
6. **Prohibir a `reparar_cierre` que toque la skill o el código.** Rechazada. Contradice el ADR 0029, según el cual un
   umbral que no se cumple se arregla en la skill o en la herramienta, y convertiría cada cierre en rojo en un
   pendiente.
7. **Solo destacarlo en el informe.** Rechazada como solución, aunque se hace además: publicar no es un control, que
   es la lección del ADR 0029.

Para que los jueces sepan qué mirar:

8. **Ampliar la búsqueda por mensaje** («motivos de la revisión final» o «cierre en la plataforma»). Rechazada: sigue
   dependiendo de un mensaje. No ve un artefacto que otro paso arrastra, como el que en H7.2 llegó con los registros
   del run (ADR 0029), ni un commit hecho a mano.
9. **Un rango desde la cabeza que juzgó el último veredicto**, sin mirar quién hizo cada commit. Elegida.

Para los topes:

10. **Un tope más corto en los ciclos del cierre** (2 veredictos y 1 corrección). Rechazada. Solo cambia el peor
    caso: los jueces convergen en una o dos rondas (H7: 2; H7.1: 1; H7.2: 2; H7.3: 2 y la manual, 1). Y justo en ese
    peor caso, la segunda y la tercera corrección son lo que la calidad pide. El tope de la constitución (tres
    correcciones por gate) vale igual para un cambio del cierre que para cualquier otro.
11. **Un presupuesto de rondas común a todo el run.** Rechazada: una revisión final que agota sus cuatro rondas
    dejaría la primera reparación con un solo veredicto y sin corrector.

Para cuándo vuelve a medir:

12. **Reutilizar la medición siempre que el producto no cambie.** Rechazada: un rojo de la plataforma (una ejecución
    cancelada, un runner caído) no se arregla en el código, y solo se aclara midiendo otra vez. Se reutiliza solo si,
    además, ninguna reparación ha terminado desde la medición: es el caso de una reanudación.

## Decisión

1. **Revisión final y cierre en un solo bucle, `bucle_final`** (`workflow.yml`, hasta tres vueltas). Cada vuelta:
   - `pendiente_revision` (`scripts/workflow/revision.sh pendiente`) decide, con el estado del disco, si hay algo que
     juzgar;
   - si lo hay, un **ciclo** de la revisión final (`ronda_revision`: los dos jueces en paralelo, el corrector sobre la
     unión, `cerrar_correccion`; hasta cuatro veredictos y tres correcciones) y `cerrar_revision`;
   - `publicar`, `medir_cierre` y, si la medición sale en rojo y queda medición, `reparar_cierre` y `cerrar_cierre`.

   El primer ciclo juzga el hito. Cada ciclo siguiente juzga lo que cambió fuera de `gates/` después del último
   veredicto, sea en commits o sin commitear. Una reparación que solo toca `gates/` no abre ninguno.

   Cada ciclo tiene su propio tope; las rondas se numeran seguidas en todo el run, para que `revision-{a,b}-r<n>.json`
   no se pisen, y el ciclo abierto vive en `gates/revision-ciclo.json`. Si un ciclo agota sus rondas, los motivos van a
   `gates/revision-pendiente.md` y al informe. Si uno posterior aprueba, el pendiente queda como
   `revision-pendiente-resuelto-r<n>.md`.
2. **Los jueces reciben el rango exacto.** `gate.sh leer revision` apunta en `gates/revision-juzgado.json` la cabeza
   que juzgó cada ronda con veredicto válido. `preparar_jueces` llama a `revision.sh cambios`, que escribe
   `gates/revision-cambios.md` y el diff fuera de `gates/` en `gates/revision-cambios-r<n>.diff`: el hito entero en la
   primera ronda y, en las demás, el rango desde el último veredicto, con cada commit, el paso que lo hizo y sus
   ficheros. En un run anterior, las cabezas juzgadas salen de los commits «veredictos de la revisión final».
   - **La convergencia deja de depender del mensaje.** Todo el rango es «la corrección», lo haya hecho el corrector o
     una reparación del cierre.
   - **Una reparación del cierre se juzga con la rúbrica entera.** El prompt de los jueces tiene un párrafo propio
     («CAMBIOS DEL CIERRE»): la reparación se juzga con la rúbrica entera, y si se aparta de lo que piden la sección
     del hito, el spec o lo que decidió un hito anterior, solo cumple si es la mejor solución a la vista de la medición
     y está registrada como decisión con su evidencia.
   - **El corrector y la reparación lo saben.** El corrector no deshace lo que arreglaba la reparación, y
     `reparar_cierre` sabe que lo que cambie irá a los jueces y que tiene que registrar cada desviación en `research.md`
     o `plan.md` y en `gates/supuestos.md`.
3. **Una medición vale solo para el producto que midió.** El orden (juzgar, después medir) hace que lo que corrige la
   revisión entre siempre en la medición siguiente. `medir_cierre` (`cierre.sh medir <hito> 3`) mide en cada vuelta,
   salvo en un caso: si desde la última medición nada ha cambiado fuera de `gates/` y ninguna reparación ha terminado
   (`gates/cierre-reparaciones`, que cuenta `cierre.sh reparado`), la reutiliza sin gastar otra. Su salida `reparar`
   (rojo y ronda < 3) es la condición del bucle y la de reparar.
4. **Reanudación.** El motor vuelve a ejecutar el bucle entero desde `pendiente_revision`. Por eso todo lo que decide
   está en ficheros, y `global.sh base <mensaje>` guarda el mensaje del paso con modelo hasta que `cerrar` termina. Al
   reanudar:
   - `global.sh retomar` hace el `cerrar` que faltaba, y el trabajo a medias del corrector o de la reparación queda
     commiteado o apartado, y a la vista del ciclo siguiente;
   - un ciclo abierto continúa con las rondas que le quedan;
   - una revisión aprobada sin cambios no se repite;
   - la medición se reutiliza si sigue valiendo.
5. **Informe final.** La sección 7 lista cada commit posterior a lo que juzgó la primera ronda, con la ronda que lo vio
   y el veredicto de sus dos jueces, o «solo registros de `gates/`». Aparte va **«Cambios que ningún juez vio»**, vacía
   en un run sano; cualquier línea ahí lleva la marca **Anomalía**, y los cambios sin commitear fuera de `gates/`
   también cuentan. En la sección 1:
   - la revisión cuenta sus rondas y sus ciclos;
   - una línea dice cuántos cambios no vio ningún juez;
   - la medición del cierre dice si es el producto de la cabeza, o qué cambió después de medir.
6. **Constitución 2.8.0**, capa 2: todo lo que cambia el producto después de un veredicto de la revisión final lo
   juzgan los dos jueces antes de que el run termine, con la misma rúbrica y un tope propio de tres correcciones;
   cada juez recibe el rango de lo cambiado desde su último veredicto; y una medición vale solo para el producto que
   midió.

**El límite del motor.** spec-kit 1.0.4 anida bucles: `ronda_revision`, un `do-while`, va dentro de un `if` dentro de
`bucle_final`, otro `do-while`, con el fan-out de los jueces dentro. Se ha comprobado con el motor real. Lo que no hace
es reanudar dentro de un anidado: vuelve a ejecutar el paso de nivel superior entero («resume will re-run the parent
step and its nested body», `engine.py`). Por eso nada de lo que decide el bucle está en las salidas de los pasos:
todas las condiciones leen pasos que corren sin condición en la misma vuelta (`pendiente_revision`, `leer_revision`,
`medir_cierre`), y el resto está en el disco. No ha hecho falta resolverlo con pasos shell.

## Coste en tiempo

Medido en los logs de los runs de H7 (`c378ced4`), H7.1 (`411e4b8c`), H7.2 (`6335565c`) y H7.3 (`f41e85d1`):

| Pieza | Duración |
|---|---|
| Una ronda de la revisión (los dos jueces en paralelo: el más lento) | 9-25 min, mediana ≈ 18 |
| Una corrección (`corrector_revision` + `cerrar_correccion` con `make ci`) | 10-23 min |
| Una medición (`medir_cierre`) | 36-41 min hasta H7.2; 14,5 min con el job en paralelo de H7.3 |
| Una reparación (`reparar_cierre` + `cerrar_cierre`) | ≈ 28 min (H7.3) |

Lo que añade cada camino frente al 2.2.0:

| Camino | Extra |
|---|---|
| Cierre en verde a la primera (H7, H7.1, H7.2) | nada |
| Reparación que solo toca `gates/` | nada |
| Una reparación que cambia el producto y los jueces aprueban (H7.3) | una ronda: 10-25 min, en lugar de una sesión a mano y una medición más |
| Lo mismo, con un rechazo | ronda + corrección + ronda: 30-75 min |
| Peor caso de un ciclo del cierre (4 veredictos, 3 correcciones) | ≈ 2 h de mediana; ≈ 2 h 50 min con las duraciones más altas |
| Peor caso del run (dos reparaciones, cada una con su ciclo agotado) | ≈ 4-6 h, dentro del presupuesto de 48 h de `hito.sh` |
| Reanudación entre una medición y el final de su reparación | ahorra una medición (15-40 min) y no gasta una de las tres |

Frente a las alternativas: juzgar al final lo acumulado (opción 1) ahorra un ciclo solo cuando hay dos reparaciones,
unos 18 min de mediana. Pero cada rechazo suyo añade una medición (15-40 min) fuera de las tres, o deja la cabeza sin
medir. Un tope de dos veredictos por ciclo (opción 10) recorta el peor caso a unos 50 min, a cambio de dejar pendiente
lo que una segunda o tercera corrección habría resuelto.

## Validación

Las salidas están en el scratchpad de la sesión (`arnes/resultados/`, `h73-validacion/`). Los clones no llevan este ADR
(ADR 0029, «Validación»). La cabecera de `revision.sh` remite aquí sin contar el caso.

**Mecánica, sin motor ni modelo** (`arnes/mecanicas.sh`, repo sintético). 25 de 25 comprobaciones:

- sin veredictos se abre el ciclo 1, que juzga el hito entero;
- el tope de cuatro veredictos cuenta dentro de cada ciclo;
- `cerrar_revision` cierra el ciclo y commitea veredictos, cabezas juzgadas y rango;
- un commit que solo toca `gates/` no abre ciclo; un cambio sin commitear en `SKILL.md` sí, con aviso en
  `revision-cambios.md`;
- el ciclo 2 empieza en la ronda 5 y tiene su propio tope;
- una aprobación posterior deja el pendiente anterior como resuelto;
- `medir` reutiliza solo si el producto no cambió y no terminó ninguna reparación;
- el informe cuenta y nombra lo que ningún juez vio, incluido lo sin commitear, y dice si la medición es la de la
  cabeza;
- `global.sh cerrar` usa el mensaje que guardó `base`, y sin ninguno sale con 2.

**Motor real y pasos falsos** (`arnes/crear-repo.sh`, `arnes/escenario.sh`). El tramo de `iniciar_revision` a
`informe_final` se extrae tal cual de `workflow.yml` y lo ejecuta `specify workflow run`, con un `claude` y un `gh`
falsos guiados por un escenario y un remoto desnudo local:

| Escenario | Qué pasa | Resultado |
|---|---|---|
| e0 | revisión aprobada, cierre verde | 1 ronda, 1 medición, nada sin juez |
| e1 | medición roja → la reparación toca `SKILL.md` → B rechaza → corrección → aprueban → medición verde | ciclo 2 con rondas 2-3; la ronda 2 ve la reparación, la 3 la corrección; 2 mediciones |
| e2 | la reparación solo toca `gates/` | sin ciclo nuevo; 2 mediciones |
| e3 | siempre rojo | 3 mediciones, 2 reparaciones, 3 ciclos; termina en rojo, sin nada sin juez |
| e4 | B rechaza siempre la reparación | ciclo 2 con 4 veredictos y 3 correcciones; motivos en `revision-pendiente.md`; se mide la cabeza corregida |
| e5a | la reparación se interrumpe con cambios | al reanudar, `retomar` los commitea, el ciclo 2 los juzga y se mide |
| e5b | la reparación se interrumpe sin cambios | al reanudar se reutiliza la medición (sin llamar a `gh`) y se repite la reparación |
| e5c | falla un juez en la ronda del ciclo 2 | al reanudar continúa el ciclo 2 con su presupuesto |
| e5d | el corrector del ciclo 2 se interrumpe | al reanudar se commitea lo que dejó y la ronda 3 lo juzga |
| e6 | falla un juez en la ronda 2 de la revisión del hito | al reanudar continúa el ciclo 1, sin abrir otro |

En todos, el árbol queda limpio, la medición final es la del producto de la cabeza y «Cambios que ningún juez vio»
está vacía.

**Sobre H7.3** (clon con la rama de H7.3 en `eb6b4c8`, antes de la sesión manual, y la de este cambio aplicada encima,
sin commitear):

- `revision.sh pendiente H7.3` abre el ciclo 2 con `eb6b4c8`, sin nota a mano: `revision-pendiente.md` no existe.
- `revision.sh cambios` da el rango `6ab3add..HEAD`, con `eb6b4c8` «paso `reparar_cierre`, tras una medición del
  cierre en rojo», sus seis ficheros, las 243 líneas de su diff y dónde está la medición que lo motivó.
- `scripts/paso.sh revision_juez_a H7.3` y `revision_juez_b H7.3`, una ejecución de cada uno, en paralelo como el
  fan-out (Opus 5 y Fable 5.1 con `xhigh`, unos 20 min). Los dos tratan el cambio de protocolo sin ninguna nota a mano.
  - **Criterio i.** Los dos lo contrastan con `docs/ROADMAP.md:339` y con el FR-040 de H7.1, y juzgan la desviación:
    «una vez por norma citada» pasa a «una vez por cada orden que lee bloques». A la vista de la medición es la mejor
    solución, y está registrada (`plan.md` «Decisiones», `research.md` «Cierre», `spec.md`, contrato §8 y el supuesto
    `[skill]`). A la recuenta en `gates/evals/boe-legislacion.json`, y B descarta la alternativa que conservaba la letra.
  - **Criterio j.** A mide lo que cuesta la comprobación en cada orden: 326 B sin nada que decir y 1 048 B con un
    `version-obsoleta`, sin crecer con lo acumulado.
  - **Veredictos.** B aprueba sin motivos. A rechaza con un motivo [f] que el rango introdujo: `plan.md:22-23` sigue
    diciendo «La comprobación no cambia de sitio» y remite a una «Decisiones» que desde `eb6b4c8` dice lo contrario, y
    el plan solo traza C1-C11. B lo deja en observaciones. La ronda manual de H7.3, con la nota, lo había dejado también
    en observaciones; aquí la convergencia lo atribuye al rango, que es donde nació. En el workflow, ese rechazo lleva al
    corrector, a otra ronda y a la medición de la cabeza corregida.
  - Veredictos en `h73-validacion/revision-{a,b}-r3.json`; rango en `h73/…/gates/revision-cambios.md`.
- `informe.sh H7.3 --solo-ver` sobre `eb6b4c8` pone en la sección 1 «Cambios que ningún juez vio: 1» y que la medición
  «NO es el producto de la cabeza». En la sección 7 marca `eb6b4c8` como **Anomalía**, con sus seis ficheros.
- Sobre la cabeza final de H7.3 (`bf67ec5`), el mismo informe da «lo vio la ronda 3 (juez A: aprobado; juez B:
  aprobado)» para `eb6b4c8` y `0222c39`, y «Ninguno» en lo que ningún juez vio.

## Consecuencias

**A favor**

- El run no puede terminar con un cambio del producto que ningún juez haya visto. Si se agotan las rondas, el
  pendiente va al informe. Si algo se escapa (una interrupción sin reanudar, un commit a mano), la sección 7 lo marca
  como anomalía.
- Los jueces juzgan lo que cambió, venga de donde venga, sin notas a mano ni convenciones de mensaje.
- La medición final es siempre la del producto que se fusionaría, y el informe lo comprueba.
- Una reanudación a mitad del cierre retoma en su sitio: no pierde trabajo, no lo deja sin juez y no gasta mediciones
  en medir otra vez lo mismo.

**En contra, y asumido**

- Un cierre que repara añade al menos una ronda de jueces (10-25 min) por reparación que cambie el producto.
- La revisión y el cierre comparten bucle: el primer ciclo de la revisión, el del hito, va dentro de `bucle_final`.
  Para quien lee el log, `ronda_revision` ya no es un paso de primer nivel.
- `publicar` corre en cada vuelta. Tras la primera solo empuja, porque la propuesta ya existe.
- La cabeza juzgada se toma al leer los veredictos, no al preparar a los jueces. Es la misma porque los jueces no
  commitean, y `revision.sh cambios` avisa si hay cambios sin commitear fuera de `gates/`, que no entran en el rango.
