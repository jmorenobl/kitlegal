# 0041 · Un plazo se lee en la redacción vigente el día en que empieza

- **Estado**: aceptada (2026-10-20)
- **Fecha**: 2026-10-20
- **Hito**: sin hito; un cambio puntual de `legal-core` y de sus evals de plazo, tras H20. No toca ningún applet, ni el
  contrato `Applet` (ADR 0005), ni el sobre (ADR 0006), ni el umbral de ninguna clase (ADR 0038)

## Contexto y problema

H9 dio a `legal-core` el cálculo de plazos, y sus evals de plazo esperan leer el precepto que fija cada plazo sin fecha,
en la redacción de hoy. H20 dio a `boe-legislacion` la lectura en una fecha, con una regla general: si la pregunta lleva
«un hecho cuyo día dan la pregunta o la conversación», se cita la redacción vigente ese día. Una pregunta de plazo la
cumple, porque da el día de la notificación. Las dos skills pedían cosas distintas para la misma pregunta, y
`legal-core` no decía cuál: en los jobs de evals de la #170 y de la #171, 1 y 3 de las 6 sesiones de Sonnet 5.5 en modo
herramienta de las evals 05 y 06 leyeron el art. 124 LPAC con `--fecha` el día de la notificación. La eval las contó
como `comando ausente`, y la 05 cayó en la #171. Esas sesiones tampoco pusieron la línea `⚠ REDACCIÓN A FECHA:` que
`boe-legislacion` pide a toda lectura fechada. Ninguna de las dos lecturas era un error de forma: falta decidir cuál es
la buena.

## Lo que dice la ley

Leído con kitlegal el 2026-10-20:

- **La disposición transitoria tercera de la Ley 39/2015** [BOE-A-2015-10565, bloque dttercera] dice, en su letra c),
  que «los actos y resoluciones dictados con posterioridad a la entrada en vigor de esta Ley se regirán, en cuanto al
  régimen de recursos, por las disposiciones de la misma»; en la a), que a los procedimientos ya iniciados «no les será
  de aplicación la misma, rigiéndose por la normativa anterior»; y en la e), que, «a falta de previsiones expresas», las
  cuestiones de derecho transitorio «en materia de procedimiento administrativo se resolverán de acuerdo con los
  principios establecidos en los apartados anteriores».
- **El art. 2.3 del Código Civil** [BOE-A-1889-4763, bloque art2]: «Las leyes no tendrán efecto retroactivo, si no
  dispusieren lo contrario».
- **La disposición transitoria segunda de la Ley 29/1998** [BOE-A-1998-16718, bloque dtsegunda] aplica el mismo
  principio al proceso contencioso: los recursos interpuestos antes de su entrada en vigor siguen «conforme a las normas
  que regían a la fecha de su iniciación». Su disposición transitoria tercera [bloque dttercera] enseña lo contrario
  cuando la ley lo dice: el régimen nuevo de casación se aplica también a resoluciones anteriores cuyo plazo no había
  terminado.

De ahí sale la regla: el régimen de un recurso, y con él su plazo, es el vigente cuando se dictó el acto; el de un
procedimiento, el vigente cuando se inició; y una norma posterior solo lo cambia si su propio régimen transitorio lo
dice. La redacción de hoy rige un plazo solo si es la misma que la de entonces o si la norma que la dio lo dispone.

## Decisión

1. **Un plazo se lee en la redacción vigente el día en que empieza**: la fecha de inicio que `legal-core` pasa a
   `plazos calcular` —la de la notificación o la publicación del acto o, en el silencio, la del acuerdo de iniciación o
   la entrada de la solicitud—. Cada precepto de `data.preceptos`, y la norma del procedimiento que se lee para el
   silencio, se lee con `boe articulo --fecha <inicio>` (o `boe_articulo` con `fecha`) y se cita como una redacción a
   fecha, con su línea `⚠ REDACCIÓN A FECHA:`. Es lo que la regla de `boe-legislacion` ya pedía, y no cambia.
2. **La fecha de la notificación hace de la del acto.** La ley ata el régimen de recursos al día en que se dictó el
   acto, que la pregunta casi nunca da; la notificación es el día que da y desde el que cuenta el plazo (art. 30.4
   LPAC), y la notificación tiene que cursarse en los diez días siguientes al acto (art. 40.2 LPAC [BOE-A-2015-10565,
   bloque a40]). Que una reforma entre en vigor entre los dos días es un supuesto que se acepta.
3. **Si el precepto cambió después, se dice y no se resuelve.** Si la lectura trae `redaccion_posterior`, la respuesta
   lleva también la línea `⚠ REDACCIÓN DE HOY:`, y dice que la fecha de `plazos` cuenta con la duración de la redacción
   de hoy y que qué redacción rige el plazo lo dice el régimen transitorio de la norma que lo cambió, que no ha leído.
   `plazos` no calcula con una redacción anterior: `data/plazos.yaml` lleva la duración de hoy.
4. **Las evals piden lo mismo que la skill.** Las evals de plazo de `legal-core` (05, 06 y 09) esperan el comando de
   bloque con `fecha:` el día de la notificación y llevan `redaccion_fechada`, de modo que `legal-core` gana el umbral
   sin modelo `redaccion_sin_fechar`, con 0, como `boe-legislacion` desde H20. Ningún umbral se relaja.

## Consecuencias

- La respuesta a una pregunta de plazo lleva, detrás de `⚠ PLAZO:`, una línea `⚠ REDACCIÓN A FECHA:` por precepto, que
  dice en qué redacción se apoya la fecha.
- La lectura fechada no lleva `kitlegal graph check` detrás ni entra en `graph check` como lectura del bloque (H20): un
  plazo deja de alimentar la memoria de redacciones de lo leído sin fecha.
- El acto presunto, que no tiene fecha de inicio, se sigue leyendo sin `--fecha`.
- Calcular con la duración de una redacción anterior, o leer el régimen transitorio de una norma modificadora, sería
  trabajo de `plazos` y de `data/plazos.yaml`, no de la skill; queda en el backlog hasta que el uso lo pida.
