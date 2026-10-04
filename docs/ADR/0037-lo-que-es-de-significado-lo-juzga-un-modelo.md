# 0037 · Lo que es de significado en una respuesta lo juzga un modelo, con rúbrica cerrada y cita comprobada; la lista de expresiones deja de decidir

- **Estado**: propuesta. La validación está hecha (2026-10-05); falta que una persona confirme sus lecturas, y la
  decide Jorge.
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
   literal que el job comprueba sin modelo, tres votos y medido contra casos etiquetados antes de decidir.
   **Elegida.**

Para el modelo del juez, **el mismo que se juzga** se rechaza. El juez tiene que ignorar lo que sabe: que el art. 66
de la LGT diga cuatro años es verdad, y aun así es un defecto si ninguna herramienta lo leyó. Un modelo distinto y más
capaz sigue mejor esa instrucción y no comparte los puntos ciegos del juzgado. Se elige el de los jueces del workflow
(`claude-opus-5-5`, ADR 0033), fijado por su id completo.

Para la clase que decide, **«la respuesta dice algo que no está en los textos leídos»** se rechaza: junta dos cosas.
Que la respuesta hable de un precepto que ninguna herramienta devolvió es casi un hecho. Que parafrasee mal uno que sí
leyó es un juicio de fidelidad, más difuso, y es donde un juez marca de más. Con umbral 0 y 108 respuestas por
ejecución, una marca de más por cada cien respuestas deja en rojo sin motivo dos ejecuciones de cada tres, y una por
cada mil, una de cada diez. Decide solo lo primero.

Para la regla de los votos, **la mayoría de tres** se rechaza en la clase que decide, porque los dos errores no
cuestan lo mismo. Una marca de más pone el job en rojo y, dentro de un run, hace que la reparación del cierre cambie
la skill por un defecto que no existe. Una de menos deja el control mejor que la lista, que vio 1 de 9. Se elige la
unanimidad, que además decide lo mismo votando por orden: un voto por respuesta, y dos más solo para las que el
primero marca.

Para cuándo se mide al juez, **en cada ejecución** se rechaza: con el modelo, la versión de Claude Code, la rúbrica y
los casos fijados, repetir la medida solo repite el azar del modelo y añade otra causa de rojos sin motivo. Se mide
cuando cambia alguno de los cuatro.

## Decisión

1. **Dos clases de comprobación en el job, y cada una en su capa.** Lo que tiene forma o es un hecho de la sesión se
   sigue comprobando sin modelo: citas, avisos, formas fijas, órdenes y herramientas ejecutadas, activación, red y
   duración. La «Decisión del mecanismo» de H5.1 sigue vigente para todo eso, y nada que un guion pueda comprobar pasa
   al juez. Lo que es de significado lo juzga un modelo, con las reglas de abajo.
2. **El juez.** Una pregunta cerrada por clase, que se responde con sí o no.
   - Recibe la pregunta de la eval, la respuesta de la sesión y los textos que devolvieron sus herramientas. No recibe
     la skill, ni lo que la eval espera, ni el veredicto de las comprobaciones sin modelo.
   - Si responde que sí, copia de la respuesta la frase que lo prueba. El job comprueba sin modelo que esa frase está
     en la respuesta. Un sí cuya frase no está es un voto nulo, y se repite una vez.
   - Cada voto es un proceso nuevo, sin herramientas y sin acceso al repositorio.
   - **En una clase que decide, una respuesta queda marcada solo si tres votos dicen que sí**, cada uno con su frase
     comprobada. Se vota por orden: el segundo y el tercero solo se piden si los anteriores marcaron.
   - Un voto que no llega a darse, por un límite de uso o por el tope de tiempo, deja la respuesta sin juzgar, y el
     job falla con ese motivo, como falla hoy con una sesión sin medir.
   - El informe publica, de cada respuesta con algún voto afirmativo, las frases y los votos.
3. **El juez se mide antes de decidir, y cada vez que cambia.**
   - Una clase solo decide si el juez se ha medido contra casos etiquetados, versionados con la procedencia de cada
     etiqueta. Hay dos clases de caso: respuestas que una persona leyó, y defectos derivados quitándole a una sesión
     correcta uno de los textos que leyó, sin escribir nada a mano.
   - Ningún caso etiquetado como defecto queda sin marcar, y ninguno etiquetado como correcto queda marcado.
   - La medida vale para la rúbrica, el id del modelo, la versión de Claude Code y los casos con los que se hizo. El
     job la repite cuando la medida versionada no corresponde a lo que hay, y si el juez no la cumple sale en rojo
     diciendo que el instrumento no vale, sin dar veredicto de la skill.
   - **Una marca que una persona lee y da por errónea entra en los casos como correcta.** El arreglo va a la rúbrica,
     no a la skill.
   - La rúbrica, los casos y el modelo del juez no los cambia un run del workflow.
4. **El modelo del juez** es `claude-opus-5-5`, fijado por su id completo junto al modelo que decide
   (`.github/workflows/evals.yml`). Cambia solo con un diff, con la medida del punto 3 repetida.
5. **Dónde corre.** En el job de evals y en el sondeo que lanza una persona. Nunca en un paso de un run del workflow
   (ADR 0032): dentro de un run, el juez se prueba con votos grabados.
6. **Los umbrales del juez usan el contrato del ADR 0029**: entran en `umbrales` con su `nombre`, su `medida` y su
   `decide`, y el informe final los lee como hoy. Del punto 1 de la decisión del ADR 0029 cambia una palabra: todo
   umbral medido sin persona tiene **un control** que pone en rojo su comprobación; ese control es un guion si lo
   medido tiene forma, y el juez de este ADR si es de significado.
7. **Las dos primeras clases**, en `boe-legislacion`:
   - `afirma_lo_no_leido`: la respuesta dice qué dice, decía o exige un precepto, o una redacción de un precepto, que
     ninguna herramienta de la sesión devolvió. **Decide**, con umbral 0. Sustituye a `redaccion_no_leida`, que es un
     caso suyo. No entra aquí la fidelidad, que la respuesta parafrasee mal un precepto que sí leyó: es otra clase,
     queda sin medir y es un candidato del backlog.
   - `cuenta_su_proceso`: la respuesta dice el estado de una comprobación o lo que el agente tiene, necesita o va a
     hacer. **Solo se publica**, con un voto. Pasa a decidir, o se retira, cuando una persona lo decida con las
     medidas de varios cierres.
8. **La lista de expresiones deja de juzgar respuestas.** Sale del veredicto y de `umbrales`. Lo decide una persona
   en la entrada de H24. No es un corrector rebajando un umbral: esa prohibición del ADR 0029 sigue vigente. El
   fichero se queda como el vocabulario que la prosa de `SKILL.md` no usa, que `make ci` ya comprueba
   (`TestProsaDeLaSkill`): una lista sirve para un texto que escribe el proyecto, no para el de un modelo.

## Validación

Hecha el 2026-10-04 y el 2026-10-05, en una sesión interactiva a petición de Jorge y fuera de cualquier run
(ADR 0032), con `claude-opus-5-5` y Claude Code 2.1.289. El resultado está al final de esta sección.

**Material.** Las respuestas del modelo que decide, en las evals que activan la skill, de los seis informes
versionados (`specs/01{1..6}-*/gates/evals/boe-legislacion.json`): 423 (51 en cada uno de los de H7.1, H7.2 y H7.3, 54
en el de H7.4 y 108 en cada uno de los de H21 y H22). Los informes guardan la respuesta y las órdenes de cada sesión,
no lo que devolvieron: esos textos se reconstruyen repitiendo las órdenes contra las grabaciones, sin modelo. Son un
superconjunto de lo que la sesión vio, que pudo filtrar la salida de una orden: eso puede hacer que el juez no marque
algo, no que marque de más. En el job, el juez recibe los textos del transcript.

**Casos.**

- **Defectos que una persona leyó**: las cinco respuestas que describen una redacción que ninguna orden devolvió
  (`docs/USO.md`, 2026-09-30): la 19-01 de H7.1, la 19-01 y la 19-02 de H7.2, y la 19-01 y la 19-02 de H7.3. Son todas
  de la misma eval.
- **Defectos derivados**: por cada informe, eval y modo, la respuesta de la primera sesión que pasó, con el texto
  del primer bloque que cita quitado. Lo que la respuesta dice de ese bloque es, por construcción, un precepto que
  ninguna herramienta devolvió. Son 140, de las dieciocho evals: 50 de los informes del ajuste y 90 de los de la
  medida.
- **Lo demás** no está etiquetado como correcto: la lectura del 2026-09-30 buscó dos clases, no esta. Una persona
  lee cada respuesta que el juez marque fuera de los dos grupos de arriba, y la etiqueta.

**Pasos.**

1. **Ajuste**, sobre los informes de H7.1, H7.2 y H7.3 y sus derivados, con los tres votos de cada respuesta para
   comparar la unanimidad con la mayoría. La rúbrica se escribe y se corrige aquí, dos veces como mucho.
2. **Medida**, una sola vez y con la rúbrica ya cerrada, sobre los informes de H7.4, H21 y H22, que nadie ha leído,
   y sus derivados. Una persona lee además las respuestas de las evals 19 y 20 de esos tres
   informes, las marque el juez o no, porque son las que traen una redacción cambiada.
3. **Lo que se anota**: por informe, las marcadas, las etiquetadas, las que el juez no vio y las que marcó de más,
   con cada regla de votos; los votos nulos; y los votos, el tiempo y el consumo de la suscripción por ejecución.

**Se acepta si**, con la unanimidad, el juez marca en el ajuste los cinco defectos leídos, en la medida marca todos
los derivados y todos los que la persona etiquete, y ninguna marca de la medida resulta errónea. **Se rechaza** si
para lograrlo la rúbrica tiene que nombrar los casos. Entonces queda la opción 4 con una lectura a mano en cada
cierre, y H23 se escribe sin «ni resume» en su umbral.

**Lo que la medida no puede dar.** Cero marcas erróneas en 270 respuestas solo acota la tasa en torno al 1 %, no
en el uno por mil que haría raros los rojos sin motivo. Por eso el punto 3 de la decisión hace entrar en los casos cada
marca errónea que aparezca después.

La rúbrica y los casos de esta validación entran en `main` con el ADR aceptado: son la entrada de H24, como la fila
de `docs/SOURCES.md` lo es de una grabación.

### Resultado

La rúbrica es la primera que se escribió: no hizo falta corregirla. Cada voto fue una sesión sin herramientas, sin
servidores MCP, sin skills y sin ninguna fuente de ajustes. Las lecturas las hizo el agente y están por confirmar.

| Grupo | Clase | Casos | Marcados por unanimidad | Por mayoría |
|---|---|---|---|---|
| Ajuste | Defectos leídos | 5 | 5 | 5 |
| Ajuste | Defectos derivados | 50 | 50 | 50 |
| Ajuste | Sin etiqueta | 148 | 21 | 25 |
| Medida | Defectos derivados | 90 | 90 | sin medir |
| Medida | Sin etiqueta | 270 | 8 | sin medir |

- **Los 145 defectos conocidos quedan marcados**, con los tres votos.
- **Ninguna marca resultó errónea.** Las 29 respuestas sin etiqueta marcadas por unanimidad, y las 7 con algún voto
  afirmativo, dicen algo de un precepto que ninguna orden devolvió: glosan una remisión («las entidades del art. 45
  (entidades de ámbito territorial inferior al municipio)»), dicen de qué trata el artículo que declaran no haber
  leído, o nombran la norma que sustituyó a la derogada. Tres de esas glosas están mal o se contradicen entre sí.
- **Las 30 respuestas de las evals 19 y 20 de la medida** no describen la redacción anterior, y el juez no marcó
  ninguna.
- **Estabilidad**: los tres votos coinciden en 198 de 203 casos del ajuste. Ningún voto nulo ni error en 1.167.
- **Votar por orden** perdió, en el ajuste, 2 de las 26 respuestas con algún voto afirmativo.
- **`cuenta_su_proceso`**, con un voto: en H7.3 marca las siete respuestas de la lectura a mano y una más. En H7.4,
  H21 y H22 marca entre 12 y 27 de cada 54, donde la lista daba 0: «El sobre no trae avisos de vigencia»,
  «(`fecha_vigencia` 20161002)».
- **Consumo**: unos 21.000 tokens de entrada y 8,4 s por voto. Una ejecución del job con 108 respuestas son unos 15
  minutos de sesión.

**Lo que la validación encontró y este ADR no esperaba.** Con el umbral en 0, el job saldría hoy en rojo por la
skill: por unanimidad, el juez marca 3 respuestas de cada 54 en el informe de H21 y 1 de cada 54 en el de H22, en
cada modo. H24 no puede cumplir su aceptación sin corregir `boe-legislacion`, que hoy deja fuera de su alcance.

## Consecuencias

**A favor**

- Lo que el principio II prohíbe, afirmar contenido legal sin fuente, tiene un control que mira lo que la respuesta
  dice y no sus palabras.
- El job deja de salir en rojo por una respuesta correcta que comparte palabras con una incorrecta.
- H23 mide «ni resume» con la misma clase, sin otra lista.
- Cada marca del juez lleva una frase que está en la respuesta: quien lee el informe ve por qué.
- Con la votación por orden, son 108 votos por ejecución de `boe-legislacion`, y dos más por cada respuesta que el
  primero marque.

**En contra, y asumido**

- **El juez no es determinista.** Lo acotan la unanimidad, la frase comprobada y la medida del punto 3, pero una
  misma respuesta puede recibir votos distintos en dos ejecuciones. El informe publica los votos para que se vea.
- **La unanimidad deja pasar los casos dudosos.** Una respuesta con dos votos afirmativos de tres no queda marcada.
  Se publica con sus frases, para que la lea una persona.
- **La fidelidad queda sin medir.** Una respuesta que lee un precepto y lo parafrasea mal no la marca nadie, ni antes
  ni ahora.
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
- **El guardián de diff del workflow** protege la rúbrica y los casos del juez (`evals/*/juez/`), como protege
  `testdata/` y `schemas/`: es un cambio del proceso, fuera de H24.
