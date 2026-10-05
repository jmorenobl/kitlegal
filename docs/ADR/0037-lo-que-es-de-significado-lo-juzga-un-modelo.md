# 0037 · Lo que es de significado en una respuesta lo juzga un modelo, con rúbrica cerrada y cita comprobada; la lista de expresiones deja de decidir

- **Estado**: aceptada (2026-10-05), con la validación de abajo.
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

Dentro de eso, **dejar que el juez ponga la frontera** se rechaza. La validación lo midió: con una rúbrica que no
decía si una glosa breve cuenta, un cambio de una línea movió 3 de las 21 marcas; con la frontera escrita, los tres
votos coinciden en todos los casos. Y **contar todo lo que la respuesta diga de un precepto no leído** también se
rechaza: marca el aviso con el que la skill dice lo que no ha mirado («el cómputo se regula en otros artículos, que
no he leído»), que le sirve a quien lee. Se elige la frontera del **precepto identificado**: lo que se nombra pide
cita.

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
   - La rúbrica y los casos validados viven en `evidencias/`, que ningún run escribe. Lo que el job usa es su copia,
     y `make ci` comprueba que es idéntica.
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
   - `afirma_lo_no_leido`. **Decide**, con umbral 0, y sustituye a `redaccion_no_leida`, que es un caso suyo.
     Cuenta que la respuesta:
     - exponga una regla que no está en ningún texto que devolvieran las herramientas, nombre o no el precepto, o
       describa una redacción que no devolvieron;
     - cambie una remisión del texto leído por la descripción de lo remitido;
     - diga de qué trata un artículo, un apartado o una disposición que nombra y que no leyó, aunque sea con una
       etiqueta breve.

     No cuenta que avise de que una materia se regula en otra parte sin nombrar ningún precepto ni decir la regla,
     que glose un título o un capítulo, ni que diga qué norma derogó o desarrolla a otra. Tampoco la fidelidad, que
     la respuesta parafrasee mal un precepto que sí leyó: es otra clase, queda sin medir y es un candidato del
     backlog.
   - `cuenta_su_proceso`: la respuesta dice el estado de una comprobación o lo que el agente tiene, necesita o va a
     hacer. **Solo se publica**, con un voto. Pasa a decidir, o se retira, cuando una persona lo decida con las
     medidas de varios cierres.
8. **La lista de expresiones deja de juzgar respuestas.** Sale del veredicto y de `umbrales`. Lo decide una persona
   en la entrada de H24. No es un corrector rebajando un umbral: esa prohibición del ADR 0029 sigue vigente. El
   fichero se queda como el vocabulario que la prosa de `SKILL.md` no usa, que `make ci` ya comprueba
   (`TestProsaDeLaSkill`): una lista sirve para un texto que escribe el proyecto, no para el de un modelo.

## Validación

Hecha el 2026-10-04 y el 2026-10-05, en una sesión interactiva a petición de Jorge y fuera de cualquier run
(ADR 0032), con `claude-opus-5-5` y Claude Code 2.1.289. Cada voto fue una sesión sin herramientas, sin servidores
MCP, sin skills y sin ninguna fuente de ajustes. La rúbrica final, los casos, los votos de la medida final y los
guiones están en `evidencias/adr-0037/`.

**Material.** Las respuestas del modelo que decide, en las evals que activan la skill, de los seis informes
versionados (`specs/01{1..6}-*/gates/evals/boe-legislacion.json`): 423 (51 en cada uno de los de H7.1, H7.2 y H7.3, 54
en el de H7.4 y 108 en cada uno de los de H21 y H22). Los informes guardan la respuesta y las órdenes de cada sesión,
no lo que devolvieron: esos textos se reconstruyeron repitiendo las órdenes contra las grabaciones, sin modelo. Son
un superconjunto de lo que la sesión vio, que pudo filtrar la salida de una orden: eso puede hacer que el juez no
marque algo, no que marque de más. En el job, el juez recibe los textos del transcript.

**Casos conocidos de antemano**, 145:

- **Defectos que una persona leyó**: las cinco respuestas que describen una redacción que ninguna orden devolvió
  (`docs/USO.md`, 2026-09-30). Son todas de la eval 19.
- **Defectos derivados**, 140: por cada informe, eval y modo, la respuesta de la primera sesión que pasó, con el
  texto del primer bloque que cita quitado. Lo que la respuesta dice de ese bloque es, por construcción, un
  precepto que ninguna herramienta devolvió.

**Dos grupos.** El ajuste son los informes de H7.1, H7.2 y H7.3 y sus derivados (203 casos, con tres votos cada
uno). La medida son los de H7.4, H21 y H22, que nadie había leído, y los suyos (360 casos, votados por orden).

### Cuatro vueltas

| Rúbrica | Qué dice de la frontera | Ajuste: sin etiqueta marcadas (de 148) | Ajuste: casos con votos divididos (de 203) | Medida: sin etiqueta marcadas (de 270) |
|---|---|---|---|---|
| 1 | nada | 21 | 5 | 8 |
| 2 | excluye qué norma derogó a otra | 21 | 6 | sin medir |
| 3 | además, toda glosa breve cuenta | 49 | 2 | 57 |
| 4 | la del precepto identificado | 47 | 0 | 20 |

En las cuatro, **los 145 casos conocidos quedan marcados por unanimidad**, y en 4.238 votos no hubo ninguno nulo ni
con error.

- **La primera** encontró lo que nadie había buscado: la skill glosa de memoria preceptos que no ha leído. Las 29
  respuestas que marcó se leyeron, y nueve se comprobaron contra los textos: ninguna era un error del juez. Tres de
  las glosas estaban mal o se contradecían: dos respuestas daban contenidos distintos a las letras derogadas n) y o)
  del art. 22.2 de la LRBRL, y otra invertía los arts. 26 y 27 de la Ley 40/2015.
- **Entre la primera y la segunda**, que solo difieren en una línea, 3 de las 21 marcas cambiaron por otras 3: todas,
  glosas breves. El juez era estable dentro de una tanda y no entre rúbricas: la frontera la ponía él.
- **La tercera** escribió que toda glosa cuenta. Quedó estable, pero marcó una de cada cinco respuestas recientes, y
  unas 23 de las 57 eran avisos de lo que la respuesta no cubre, sin ningún precepto nombrado.
- **La cuarta** es la decisión.

### La medida final

Jorge eligió la frontera del precepto identificado el 2026-10-05. Las 120 respuestas que alguna rúbrica había
marcado se etiquetaron leyéndolas, antes de escribir la rúbrica final: 67 defectos, 47 avisos correctos, 5 que ya
eran defectos leídos y 1 dudosa, que no entra en el criterio.

| Grupo | Clase | Casos | Marcados por unanimidad |
|---|---|---|---|
| Ajuste | Defectos leídos en septiembre | 5 | 5 |
| Ajuste | Defectos derivados | 50 | 50 |
| Ajuste | Defecto, según la lectura | 47 | 47 |
| Ajuste | Aviso correcto, según la lectura | 4 | 0 |
| Ajuste | Sin etiqueta | 97 | 0 |
| Medida | Defectos derivados | 90 | 90 |
| Medida | Defecto, según la lectura | 20 | 20 |
| Medida | Aviso correcto, según la lectura | 43 | 0 |
| Medida | Dudosa | 1 | 0 |
| Medida | Sin etiqueta | 206 | 0 |

- **Defectos**: 212 de 212 marcados con los tres votos.
- **Avisos correctos**: 0 de 47, ni con un solo voto.
- **Estabilidad**: en el ajuste, los tres votos coinciden en los 203 casos.
- **Consumo**: unos 21.000 tokens de entrada y 8 s por voto. Una ejecución del job con 108 respuestas son unos 15
  minutos de sesión.

**Lo que marcaría hoy el job**: 3 respuestas de 54 en el informe de H7.4; 5 y 6 en el de H21, por modo; y 2 y 4 en
el de H22. Con el umbral en 0 saldría en rojo por la skill, no por el juez: H24 la corrige.

**`cuenta_su_proceso`**, con un voto: en H7.3 marca las siete respuestas de la lectura a mano y una más. En H7.4,
H21 y H22 marca entre 12 y 30 de cada 54, donde la lista daba 0: «El sobre no trae avisos de vigencia»,
«(`fecha_vigencia` 20161002)». Sale de la prosa de `SKILL.md`, que nombra el sobre y los dos campos al decir qué
tiene que llevar la respuesta.

### Lo que la validación no prueba

- **Las etiquetas de la lectura y la rúbrica final son de la misma mano**, y las etiquetas salen de las frases que
  el juez citó en las vueltas anteriores. El acuerdo total dice que la rúbrica transmite la frontera sin
  ambigüedad; no es una estimación independiente de los errores del juez.
- **Las 303 respuestas que ninguna rúbrica marcó no se han leído enteras.** Los derivados dicen que al juez no se le
  escapa un bloque entero sin leer; no dicen nada de una frase suelta.
- **Votar por orden pierde algo**: con la primera rúbrica, 2 de las 26 respuestas con algún voto afirmativo
  tuvieron el primero negativo.
- **La primera medida sobre respuestas nuevas** es el cierre de H24, que lee una persona.

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

- **El juez no es determinista, y es sensible a la rúbrica.** Lo acotan la unanimidad, la frase comprobada y la
  medida del punto 3, pero una línea de la rúbrica mueve los casos de la frontera. Por eso la rúbrica dice dónde
  está, vive en `evidencias/` y no cambia sin repetir la medida.
- **La frontera deja fuera cosas que son contenido legal sin fuente**: la glosa de un título o de un capítulo, y qué
  norma derogó a otra. Se eligió así porque ahí el juez se dividía o el dato servía a quien lee.
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

## Enmiendas que provoca

Van en la propuesta que acepta este ADR:

- **Constitución** (2.11.0). En «Gates», la quinta regla dice «un control» donde decía «un control mecánico», y
  dice cuál es cada uno. La capa 2 gana el significado de la respuesta de una eval, cuando no tiene forma fija.
- **ADR 0029**: su estado dice que este ADR lo sustituye en parte.
- **`docs/ROADMAP.md`**: enmienda fechada en la «Decisión del mecanismo» de H5.1 y en la fila «Evals de skills» de
  §3; la sección de H24; y en H23, cómo se mide «ni resume».
- **`CLAUDE.md`**: una decisión cerrada más, el juez de las evals, y H24 como siguiente hito.
- **`evidencias/adr-0037/`**: la rúbrica, los casos y los votos.
- **`docs/WORKFLOW.md`**: la fila de la evidencia de un ADR en su tabla de rutas.
- **`docs/USO.md`**: la entrada del 2026-10-05, con lo que la validación encontró en la skill.

Lo que cambia con H24 y no aquí: el job (`internal/evals`, `.github/workflows/evals.yml`), lo que `docs/WORKFLOW.md`
y `CONTRIBUTING.md` dicen de él, y `boe-legislacion`.
