# 0027 · Actuar en cualquier asunto: la fase 2 es genérica, con la cobertura por materia declarada y la redacción a una fecha

- **Estado**: aceptada
- **Fecha**: 2026-09-27
- **Hito**: transversal (antes de H7). Enmienda el encuadre de las fases 2 y 3 del ADR 0013 sin cambiar su orden;
  añade H20 entre H7 y H8; amplía H8, H9, H10 y H11; constitución 2.5.0 (principios II y IX).

## Contexto y problema

El roadmap llamaba a la fase 2 «Actuar en mi municipio» y a la 3 «Consultar mi municipio». Es la idea de partida del
proyecto, pero no el producto que se ofrece: la web tiene una portada para despachos y otra para la ciudadanía
(ADR 0024), con ejemplos de contratación, tributos, trabajo, alquiler e IRPF, y el README ya anuncia que los escritos
«valen ante cualquier administración, porque el procedimiento común es el mismo para un ministerio, una consejería,
una diputación o un ayuntamiento», y que después viene saber qué hacen las administraciones. El roadmap era el único
que seguía diciendo municipio, y es lo que lee el workflow: la sección del hito es la entrada del run (ADR 0018), y con
«lo no especificado no se implementa» y la lectura conservadora, H10 y H11 habrían salido municipales.

Lo que ataba la fase 2 al municipio era concreto:

- **H10** guardaba en `config.yaml` «el municipio de la persona usuaria», como el principio IX. Un despacho lleva
  asuntos de clientes de muchos municipios, y el art. 30.6 LPAC no mira a quien usa el kit: tiene por inhábil el día
  que lo sea en el municipio o la comunidad donde reside el interesado o en la sede del órgano. Son datos del asunto.
- **H11** preveía el recurso de reposición «contra un acto municipal» y sacaba el órgano destinatario de
  `territorio`, que solo conoce el DIR3 de los ayuntamientos; para el resto de órganos no hay fuente descargable
  (ADR 0017). No hace falta: toda notificación indica si el acto pone fin a la vía administrativa, los recursos que
  proceden, el órgano ante el que presentarlos y el plazo (art. 40.2 LPAC).

Generalizar abre una pregunta que el roadmap no contestaba: dónde acaba la base y empieza la vertical (ADR 0012).
Aplicar la LPAC a cualquier materia sería el error que la web promete evitar. La disposición adicional primera,
apartado 2, de la LPAC deja regidos por su normativa específica, con la LPAC como supletoria, la aplicación de los
tributos y su revisión, la Seguridad Social y el desempleo, los procedimientos sancionadores tributario, del orden
social, de tráfico y de extranjería, y la extranjería y el asilo; el art. 112.2 permite a las leyes sustituir la
alzada por otras vías, y el 112.4 remite las reclamaciones económico-administrativas a su legislación. Un recurso de
reposición de la LPAC contra una multa de tráfico, con su plazo, es un escrito mal fundado.

Y hay dos huecos para quien trabaja un asunto, sea un despacho o un ciudadano:

- **La redacción a una fecha.** Un asunto tiene la fecha del acto o de los hechos, y lo que se cita es la redacción
  vigente entonces. La respuesta del BOE a un bloque trae todas sus versiones y el binario se queda con la última
  (`internal/source/boe/doc.go`, nota [5]). El modelo no sabe de memoria, con fiabilidad, una redacción anterior, y
  `cita validar` (H8) no puede comprobar un escrito de hace años contra el texto de hoy sin dar por errónea una cita
  que era correcta en su fecha.
- **Verificar lo que otro ha escrito.** H8 validaba una cita; un despacho necesita pasar un escrito entero, o la
  respuesta de otro asistente, y saber qué citas no se resuelven, cuáles apuntan a una norma derogada y cuáles a una
  redacción que ha cambiado.

La entrada externa: el 2026-09-27 Jorge contrastó kitlegal con lo que existe —otros servidores sobre la API del BOE y
las herramientas de las editoriales jurídicas— en una conversación con otro agente, que propuso situarlo como la capa
de verificación que va debajo de cualquier agente: validar citas de cualquier origen, el historial de redacciones, un
benchmark público, presencia en directorios con un servidor MCP y telemetría para alimentar el grafo. Lo que encaja
con las decisiones cerradas entra aquí; lo que no, se deja escrito.

Y una medida que corrige una intuición: la búsqueda de un artículo por su materia, anotada en la bitácora el
2026-09-15, no está demostrado que haga falta. En la primera ejecución con Sonnet (2026-09-16), las cinco preguntas
por materia de `boe-legislacion` pasaron 3 de 3 (ADR 0016). Falta medirlo con normas y artículos menos conocidos.

## Opciones consideradas

1. **Dejar el roadmap y resolverlo en el spec de H10 y H11.** Rechazada: el run lee la sección del hito y la
   constitución tal como están, y la lectura conservadora produce la versión municipal. La entrada es de la persona
   (ADR 0018) y es aquí donde se decide.
2. **Fase 2 genérica en todas las materias**, procedimientos especiales incluidos. Rechazada: los procedimientos de
   la DA 1ª.2 tienen recursos, plazos y órganos propios; meterlos en la base es construir las verticales dentro de
   ella, contra el ADR 0012, y multiplica H10 y H11.
3. **Fase 2 genérica por procedimiento**: el procedimiento común y el acceso a la información, ante cualquier
   administración y en cualquier materia; lo que la ley regula aparte, declarado no cubierto, como hoy se declara un
   territorio no cubierto; la redacción a una fecha como hito propio antes de `cita`, y la verificación de textos
   ajenos dentro de H8. Elegida.

## Decisión

1. **La fase 2 pasa a ser «Actuar: llevar un asunto ante cualquier administración»**, para quien lleva un asunto
   propio o ajeno: un ciudadano o un despacho. El ayuntamiento es un destinatario más. El orden no cambia: H10 y H11
   van detrás de H9.
2. **El asunto lleva sus municipios.** El `.kitlegal/config.yaml` de cada asunto guarda los dos que mira el art. 30.6
   LPAC —el del interesado y el de la sede del órgano— y no el de quien usa el kit. La administración y el órgano son
   datos del asunto: el DIR3 por `territorio` cuando es un ayuntamiento y, si no, lo que dice la notificación
   (art. 40.2). El escrito se firma como interesado o como representante (art. 5 LPAC), y los datos personales van en
   el fichero, nunca en el grafo (principio VII).
3. **Cobertura por materia.** Las herramientas y las skills base cubren el procedimiento común de la LPAC y el acceso
   a la información de la LTAIBG. Cuando el asunto cae en un procedimiento que la ley regula aparte —la DA 1ª.2 LPAC,
   las vías que sustituyen a la alzada (art. 112.2) y las reclamaciones económico-administrativas (art. 112.4)—, la
   skill lo declara no cubierto en su respuesta, y la herramienta que reciba el procedimiento como dato, en `data`;
   las dos nombran la norma que lo rige y nunca aplican los recursos ni los plazos del común. La declaración lleva una
   forma fija, como los avisos de vigencia (H5.1), para que las evals la comprueben por su forma y no por la opinión
   de un modelo. Llega con la vertical que lo cubra (backlog). Es la regla del principio IX para el territorio, aplicada a la materia, y entra en la
   constitución (principio II) porque obliga a todas las skills. Empieza en H9, la primera herramienta que calcula
   recursos y plazos; la lista de procedimientos vive en `data/` y cita los preceptos que la fundan.
4. **H11 genera tres escritos genéricos**: la solicitud de acceso a la información (LTAIBG, art. 17); el recurso de
   alzada o el potestativo de reposición contra un acto de cualquier administración, según la notificación diga si
   pone fin a la vía; y las alegaciones (arts. 76 y 82 LPAC), que el README ya anuncia, con el plazo del trámite de
   audiencia que fije el órgano, entre diez y quince días (art. 82.2).
5. **H20, la redacción a una fecha**, entre H7 y H8: `boe articulo` devuelve la versión vigente en una fecha, que el
   BOE ya da, y `boe-legislacion` la cita cuando la pregunta tiene fecha. Va después de H7 para nacer emitiendo cada
   `BloqueVersion` que observa (ADR 0014) —y para que `graph check` no tome por caducada una redacción pedida a
   propósito—, y antes de H8 porque validar un texto exige la redacción de su fecha.
   Recibe el siguiente número libre (§6.7 del roadmap), como lo conservó H19.
6. **H8 valida textos ajenos**: `cita-verificada` revisa las citas de un texto entero —un escrito, un dictamen, la
   respuesta de otro asistente— a la fecha del texto, y devuelve cada hallazgo en `data` con salida 0 (ADR 0023). Sin
   PDF: extraer el texto de un PDF es una dependencia nueva que pide su propio hito.
7. **La fase 3 pasa a ser «Consultar lo que hacen las administraciones, empezando por el municipio»**, como la
   anuncia el README. Sus hitos no cambian: se validan con el ayuntamiento porque es el órgano cuyo DIR3 da
   `territorio`.
8. **El artículo por materia se queda en el backlog**, con una condición de entrada medible: evals informativas por
   materia con normas y artículos menos conocidos; si el modelo del uso real (ADR 0016) no las pasa, la herramienta
   recibe número.
9. **Sin telemetría.** kitlegal no envía preguntas, asuntos ni datos de uso a ningún servidor propio, tampoco con
   consentimiento: lo promete la web y lo exige el ADR 0014 para el asunto. El grafo es local. Si algún día se plantea
   otra cosa, es un ADR que sustituye esa promesa, no un hito.
10. **Lo que no es un hito**: la presencia en directorios de skills y de servidores MCP, un benchmark público y los
    acuerdos con colegios o universidades no entregan ni protegen una skill (principio VIII) y los decide la persona,
    como la web (ADR 0024). Si necesitan código del producto —el servidor MCP del backlog de distribución—, ese código
    sí es un candidato. Un benchmark se juzgaría como las evals: citas por identificador contra el BOE, sin la opinión
    de un modelo (constitución, capa 1).

No se reabre ninguna decisión cerrada: base antes que vertical (ADR 0012) sale reforzado con una frontera explícita;
el orden por tiempo hasta el uso (ADR 0013), el grafo en tres piezas (ADR 0014), la frontera humana (ADR 0004) y el
contrato de resultados (ADR 0023) no cambian; CENDOJ sigue siendo solo ECLI y metadatos (ADR 0003).

## Consecuencias

**A favor**

- El roadmap dice lo mismo que la web y el README, y los runs de H10 y H11 construirán lo que se ofrece.
- La frontera entre base y vertical deja de ser una intuición: la traza la LPAC y se comprueba con evals («¿qué plazo
  tengo para recurrir una liquidación de la Agencia Tributaria?» tiene que declarar la cobertura, no dar un mes).
- Quien lleva un asunto puede citar la redacción aplicable a sus hechos y revisar las citas de un texto ajeno, con la
  misma huella y la misma fuente que cualquier otra consulta.

**En contra, y asumido**

- La fase 1 crece un hito (H20) y H8, H9 y H11 crecen: el bucle completo de la fase 2 llega un hito más tarde.
- La numeración vuelve a no seguir el orden (H7, H20, H8), como con H19; el siguiente número libre evita una cuarta
  renumeración.
- La lista de procedimientos que la ley regula aparte hay que mantenerla; su fundamento son la DA 1ª.2 y los
  arts. 112.2 y 112.4, y la amplía quien añade cada vertical.
- La condición de entrada del artículo por materia depende de que alguien escriba esas evals; si nadie lo hace, sigue
  en el backlog.
