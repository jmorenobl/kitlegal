# 0037 · Lo que es de significado en una respuesta lo juzga un modelo, con rúbrica cerrada y cita comprobada; la lista de expresiones deja de decidir

- **Estado**: propuesta. No se acepta sin la medida de «Validación», que exige sesiones con modelo lanzadas por una
  persona, y la decide Jorge.
- **Fecha**: 2026-10-04
- **Hito**: transversal (tras H22; antes de H24, que lo implementa en el job de evals, y de H23, que lo usa).
  Sustituye del ADR 0029 una palabra del punto 1 de su decisión —el control de un umbral era «mecánico»— y, del
  roadmap, la «Decisión del mecanismo» de H5.1 en lo que descartaba todo juicio con modelo. El resto del ADR 0029
  sigue vigente: el contrato de `umbrales`, «Controles de umbral» en el plan y el informe final sin modelo.

## Contexto y problema

Que el job de evals juzga sin ningún modelo no lo decidió un ADR. Lo decidió Jorge el 2026-09-16 al revisar el plan
de H5.1 (`docs/ROADMAP.md`, «Decisión del mecanismo»): el aviso de vigencia se compara por su forma fija, como la cita
por su identificador, y «un juicio semántico (modelo o similitud) queda descartado». Para lo que tiene forma, la
decisión acertó y sigue en pie: la cita con su identificador y su bloque, las líneas `⚠ NORMA DEROGADA:`,
`⚠ REDACCIÓN MODIFICADA:` y `⚠ SIN CONSULTA AL BOE:`, las órdenes que la sesión ejecutó, si la skill se activó y lo
que salió a la red son hechos, y un guion los comprueba.

H7.2 llevó el mismo mecanismo a lo que no tiene forma: una lista de expresiones que la respuesta no debe llevar
(`evals/boe-legislacion/expresiones-prohibidas.yaml`). El ADR 0029 hizo que su umbral decidiera, y escribió que todo
umbral medido sin persona tiene «un control mecánico». Desde entonces la lista ha fallado por los dos lados:

- **No ve lo que tiene que ver.** Leídas a mano las respuestas del cierre de H7.3 (`docs/USO.md`, 2026-09-30), 9 de
  51 cuentan la comprobación o describen una redacción que ninguna orden devolvió; la lista marcó 1. En las otras dos
  mediciones de ese cierre, 12 frente a 5 y 6 frente a 0. Con el modelo que decide ahora, 28 de 51 frente a 0
  (`docs/ROADMAP.md`, H7.4).
- **Ve lo que no es.** El 2026-10-04 el job salió en rojo por dos respuestas correctas de tres en la eval sin binario
  ni servidor: ofrecían repetir la consulta («lo consulto y te respondo con el texto y su cita») y casaban con
  «respondo con el texto» (`docs/USO.md`, 2026-10-04; #116).
- **Cada arreglo traslada el problema.** H7.3 quitó de `SKILL.md` el vocabulario que provocaba el ruido, H7.4 rehízo
  la lista por clases y la #116 sacó de ella una eval. Los cierres de H7.4, H21 y H22 dan 0 de 54, en cada modo,
  pero nadie ha leído esas respuestas a mano: el cero dice que ninguna forma de la lista aparece, no que ninguna
  respuesta cuente su proceso o afirme lo que no ha leído.

Lo que la lista quiere medir es significado. Dos de las cosas que mide importan de forma distinta: que una respuesta
**afirme el contenido de una norma que ninguna herramienta de la sesión devolvió** es contenido legal sin fuente
(constitución, principio II); que **cuente su proceso** («la comprobación terminó», «tengo suficiente para
responder») es ruido.

H23 llega con el mismo problema. Su umbral dice que ninguna respuesta «cita ni resume» una sentencia inventada. «Cita»
tiene forma: un ECLI, un ROJ o un número con su fecha. «Resume» no la tiene, y con el ADR 0029 como está H23
escribiría otra lista.

La constitución ya dice dónde va cada cosa («Gates»): cada comprobación vive en la capa más baja que pueda
verificarla; lo que se puede comprobar con un guion nunca se delega a un modelo, y lo que exige criterio se juzga con
rúbrica cerrada. El significado de un texto libre no lo verifica un guion. El defecto fue dejarlo en la capa 1.

## Opciones consideradas

1. **Seguir afinando la lista.** Rechazada: es lo que se ha hecho en H7.2, H7.3, H7.4 y la #116, y la lista sigue
   sin detectar lo dicho con otras palabras, que su propia cabecera declara, y marca lo correcto que comparte
   palabras con lo incorrecto.
2. **Similitud con frases de referencia.** Rechazada ya en H5.1: no separa «ha sido derogada» de «no ha sido
   derogada». Además usa un modelo sin dar una evidencia que una persona pueda leer.
3. **Dar forma fija a todo lo que la respuesta diga.** Rechazada como solución (se mantiene para lo que ya la tiene):
   una forma fija comprueba lo que la respuesta tiene que llevar, no lo que no tiene que decir, y obligar a una
   plantilla en todo el texto empeora la respuesta para quien la lee.
4. **Publicar y que lea una persona.** Rechazada: es lo que el ADR 0029 corrigió. Nadie ha leído a mano las
   respuestas de los tres últimos cierres.
5. **Un juez con modelo que valore la respuesta entera** («¿es una buena respuesta?»). Rechazada: no es reproducible,
   no deja evidencia y juzgaría también lo que un guion ya comprueba.
6. **Un juez con modelo acotado**: una pregunta cerrada por clase, con los textos que la sesión leyó delante, una cita
   literal que el job comprueba sin modelo, mayoría de tres y medido contra respuestas etiquetadas antes de decidir.
   **Elegida.**

Para el modelo del juez, **el mismo que se juzga** se rechaza: comparte sus puntos ciegos. Se propone el de los jueces
del workflow (`claude-opus-5-5`, ADR 0033), fijado por su id completo. La validación lo confirma o lo cambia.

## Decisión

1. **Dos clases de comprobación en el job, y cada una en su capa.** Lo que tiene forma o es un hecho de la sesión se
   sigue comprobando sin modelo: citas, avisos, formas fijas, órdenes y herramientas ejecutadas, activación, red y
   duración. La «Decisión del mecanismo» de H5.1 sigue vigente para todo eso, y nada que un guion pueda comprobar pasa
   al juez. Lo que es de significado lo juzga un modelo, con las reglas de abajo.
2. **El juez.** Una pregunta cerrada por clase, que se responde con sí o no.
   - Recibe la pregunta de la eval, la respuesta de la sesión y los textos que devolvieron sus herramientas. No recibe
     la skill, ni lo que la eval espera, ni el veredicto de las comprobaciones sin modelo.
   - Si responde que sí, copia de la respuesta la frase que lo prueba. El job comprueba sin modelo que esa frase está
     en la respuesta. Un voto afirmativo cuya frase no está es nulo.
   - Cada voto es un proceso nuevo, sin herramientas y sin acceso al repositorio. Tres votos por respuesta, y decide
     la mayoría de los válidos. Una respuesta sin mayoría queda «sin juzgar», y el job falla con ese motivo, como
     falla hoy con una sesión sin medir.
   - El informe publica, de cada respuesta que el juez marca, la frase y los tres votos.
3. **El juez se mide antes de decidir.** Una clase solo decide si el juez se ha medido contra respuestas etiquetadas
   por una persona, versionadas con la procedencia de cada etiqueta. Ninguna respuesta etiquetada como defecto queda
   sin marcar, y las marcadas de más no pasan del umbral que fije la validación. El job repite esa medida y, si el
   juez no la cumple, sale en rojo diciendo que el instrumento no vale, sin dar veredicto de la skill.
4. **El modelo del juez** se fija por su id completo junto al modelo que decide (`.github/workflows/evals.yml`) y
   cambia solo con un diff, con la medida del punto 3 repetida.
5. **Dónde corre.** En el job de evals y en el sondeo que lanza una persona. Nunca en un paso de un run del workflow
   (ADR 0032): dentro de un run, el juez se prueba con votos grabados.
6. **Los umbrales del juez usan el contrato del ADR 0029**: entran en `umbrales` con su `nombre`, su `medida` y su
   `decide`, y el informe final los lee como hoy. Del punto 1 de la decisión del ADR 0029 cambia una palabra: todo
   umbral medido sin persona tiene **un control** que pone en rojo su comprobación; ese control es un guion si lo
   medido tiene forma, y el juez de este ADR si es de significado.
7. **Las dos primeras clases**, en `boe-legislacion`:
   - `afirma_lo_no_leido`: la respuesta dice qué dice, decía o exige un precepto, y ese contenido no está en ningún
     texto que devolvieran las herramientas de la sesión. **Decide**, con umbral 0. Sustituye a `redaccion_no_leida`,
     que es un caso suyo.
   - `cuenta_su_proceso`: la respuesta dice el estado de una comprobación o lo que el agente tiene, necesita o va a
     hacer. **Solo se publica.** Pasa a decidir cuando una persona lo decida, con las medidas de varios cierres.
8. **La lista de expresiones deja de juzgar respuestas.** Sale del veredicto y de `umbrales`. Lo decide una persona
   en la entrada de H24. No es un corrector rebajando un umbral: esa prohibición del ADR 0029 sigue vigente. El
   fichero se queda como el vocabulario que la prosa de `SKILL.md` no usa, que `make ci` ya comprueba
   (`TestProsaDeLaSkill`): una lista sirve para un texto que escribe el proyecto, no para el de un modelo.

## Validación

**Pendiente.** Hay que hacerla antes de aceptar este ADR y de lanzar H24, con sesiones con modelo que lanza una
persona, o una sesión interactiva a petición suya, fuera de cualquier run (ADR 0032).

Material: las respuestas del modelo que decide, en las evals que activan la skill, de los seis informes versionados
(`specs/01{1..6}-*/gates/evals/boe-legislacion.json`): 423 (51 en cada uno de los de H7.1, H7.2 y H7.3, 54 en el de
H7.4 y 108 en cada uno de los de H21 y H22). Los informes guardan la respuesta y las órdenes de cada sesión, no lo
que devolvieron: esos textos se reconstruyen repitiendo las órdenes contra las grabaciones, sin modelo.

1. **Ajuste**, sobre los tres informes que una persona leyó respuesta a respuesta (`docs/USO.md`, 2026-09-30). Tienen
   cinco respuestas que describen una redacción que ninguna orden devolvió: la 19-01 de H7.1, la 19-01 y la 19-02 de
   H7.2, y la 19-01 y la 19-02 de H7.3. La rúbrica se escribe y se corrige aquí, dos veces como mucho.
2. **Medida**, una sola vez y con la rúbrica ya cerrada, sobre los tres informes que nadie ha leído (H7.4, H21 y
   H22). Una persona lee cada respuesta que el juez marque, y además las de las evals 19 y 20 de esos tres informes,
   las marque o no, porque son las que traen una redacción cambiada.
3. **Lo que se anota**: por informe, las marcadas, las etiquetadas, las que el juez no vio y las que marcó de más;
   los votos nulos y las respuestas sin mayoría; y los votos, el tiempo y el consumo de la suscripción por ejecución.

**Se acepta si** en el ajuste el juez marca las cinco, en la medida no deja sin marcar ninguna que la persona
etiquete, y las marcadas de más, leídas una a una, no pasan del umbral que Jorge fije con la cifra delante. **Se
rechaza** si para lograrlo la rúbrica tiene que nombrar los casos. Entonces queda la opción 4 con una lectura a mano
en cada cierre, y H23 se escribe sin «ni resume» en su umbral.

La rúbrica y las etiquetas de esta validación entran en `main` con el ADR aceptado: son la entrada de H24, como la
fila de `docs/SOURCES.md` lo es de una grabación.

## Consecuencias

**A favor**

- Lo que el principio II prohíbe, afirmar contenido legal sin fuente, tiene un control que mira lo que la respuesta
  dice y no sus palabras.
- El job deja de salir en rojo por una respuesta correcta que comparte palabras con una incorrecta.
- H23 mide «ni resume» con la misma clase, sin otra lista.
- Cada marca del juez lleva una frase que está en la respuesta: quien lee el informe ve por qué.

**En contra, y asumido**

- **El juez no es determinista.** Lo acotan la mayoría de tres, la frase comprobada y la medida del punto 3, pero una
  misma respuesta puede recibir votos distintos en dos ejecuciones. El informe publica los votos para que se vea.
- **Cuesta suscripción y tiempo.** Con 54 respuestas del modelo que decide por modo, son 324 votos por ejecución de
  `boe-legislacion`, más los de la medida del juez. La cifra real la da la validación; H24 le pone un umbral.
- **`cuenta_su_proceso` se queda sin un control que decida.** Hoy lo decide la lista, mal. Es una pérdida declarada
  respecto del criterio de éxito de H7.2, hasta que una persona lo promueva.
- **Juez y juzgado son modelos del mismo proveedor**, y pueden compartir errores. Lo mitigan que sean modelos
  distintos, que el juez tenga delante los textos leídos y que tenga que citar.
- **La respuesta es texto ajeno para el juez.** Por eso no tiene herramientas ni acceso al repositorio.
- **Si el id del juez se retira**, hay que repetir la medida con el que lo sustituya antes de que el job vuelva a
  decidir.

## Enmiendas que provoca, al aceptarse

- **Constitución** (2.11.0). En «Gates», la quinta regla dice «un control» donde dice «un control mecánico», con la
  remisión a este ADR. La capa 2 gana «el significado de la respuesta de una eval, cuando no tiene forma fija»; la
  capa 1 conserva «citas esperadas de las evals comparadas por identificador, no por opinión».
- **ADR 0029**: su estado pasa a «aceptada; sustituida en parte por el ADR 0037».
- **`docs/ROADMAP.md`**: enmienda fechada en la «Decisión del mecanismo» de H5.1 y en la fila «Evals de skills» de
  §3; la sección de H24; y en H23, cómo se mide «ni resume».
- **`CLAUDE.md`**: «Modelo que decide las evals» nombra también el del juez, y el siguiente hito pasa a ser H24.
- **`docs/WORKFLOW.md`**, donde describe el job de evals.
