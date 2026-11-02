# El cálculo de los plazos

El detalle de los pasos 8 a 10 de `SKILL.md`: los municipios que cuentan, el plazo que `plazos` no calcula y la
lectura de un precepto en la fecha de inicio. Las líneas que da la respuesta son las de «Cómo se presenta un plazo»
del cuerpo.

## Los municipios que cuentan

- En un plazo administrativo, el del interesado y el de la sede del órgano, que pueden ser el mismo; en el recurso
  contencioso-administrativo, el de la sede del juzgado. Si la persona no los ha dado, pregúntalos; nunca supongas el
  de quien usa el kit.
- A `plazos` pásale solo municipios que `territorio` haya resuelto en esta conversación, cada uno como lo dio la
  persona y en su orden. `plazos` acepta un nombre, pero sin el paso 8 la respuesta no tiene el territorio ni la
  cobertura de ese municipio.

## El plazo sin calcular

Con el código 3 y un `data.mensaje` que empieza por «el plazo de tipo», no hay fecha. Un plazo se rige por la
redacción vigente el día en que empieza, y algún precepto con que `plazos` lo cuenta no decía entonces lo que dice hoy
(cada línea del mensaje nombra uno). Qué regía lo dicen esa redacción y el régimen transitorio de la norma que la
cambió, y `plazos` no elige: tú tampoco. Da la línea de plazo sin calcular del cuerpo, con el tipo como lo nombra
la referencia del paquete, la fecha de inicio y cada línea del mensaje, con su norma, su bloque y su día; sin la línea
del plazo, sin ninguna fecha de vencimiento y sin nombrar de memoria la norma que regía.

## La lectura en la fecha de inicio

- Un plazo se rige por la redacción vigente el día en que empieza, no por la de hoy: por eso cada precepto de
  `data.preceptos`, y la norma que leas para el silencio, se lee en la fecha de inicio que pasaste a `plazos`. Su
  línea de redacción a fecha lleva la fecha de inicio y la fecha de vigencia que da la lectura.
- Si esa lectura sale con 4 o 5 en un precepto de `data.preceptos`, la línea del plazo se mantiene, con la fecha de
  `plazos` y la cita de cada precepto. En lugar de su línea de redacción a fecha, di aparte que no has podido leer esa
  redacción, como se dice cualquier lectura fallida: qué bloque falta y por qué, y no afirmes qué dice el precepto.
- Si el `data` de la lectura trae `redaccion_posterior`, el bloque cambió después: va también su línea de redacción de
  hoy, con la norma y la fecha de vigencia de ese campo. En un precepto de `data.preceptos`, `plazos` ya comprobó que su apartado decía lo mismo ese día: lo que cambió
  es otra parte del bloque. En la norma del silencio, la duración que pasaste es la de ese día, y si rige la de hoy lo
  diría el régimen transitorio de la norma que la cambió, que no has leído. Dilo así.
