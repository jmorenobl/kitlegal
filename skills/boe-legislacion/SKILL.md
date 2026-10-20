---
name: boe-legislacion
description: >-
  Consulta y cita normativa consolidada del Boletín Oficial del Estado (BOE) de cualquier materia: procedimiento
  administrativo, contratación pública, régimen local, tributos y haciendas locales, transparencia, relaciones
  laborales… Úsala siempre que la respuesta dependa de lo que dice una norma —qué dice un artículo, una ley o un real
  decreto, qué plazo, requisito o procedimiento fija, dónde se regula una materia— o de lo que decía en una fecha —la
  del acto, la de los hechos—, también cuando creas conocer la respuesta: sin leer la norma con kitlegal, no hay cita.
  Actívala antes de llamar a boe_articulo u otra herramienta boe_…: qué pedir y cómo citar lo dice la skill. La norma
  puede nombrarse por su número y año (Ley 39/2015, Real Decreto Legislativo 2/2004), por su abreviatura (LPAC, LCSP,
  LRBRL, LGT, TRLRHL, LRJSP) o por su identificador BOE-A-…, o pedirse el texto, hoy o en una fecha, de una norma
  estatal o autonómica consolidada en el BOE. Lee índice y artículos con kitlegal y cita identificador y bloque.
metadata:
  kitlegal-applets: boe graph
  kitlegal-referencias: normas
---

# Consultar y citar legislación consolidada del BOE

Esta skill responde preguntas sobre el contenido de normas consolidadas del Boletín Oficial del Estado —la Constitución,
leyes, reales decretos legislativos, reales decretos y las normas autonómicas que el BOE consolida— de cualquier ámbito.
Lo que dice de una norma sale del texto que devuelve `kitlegal boe` en la misma conversación, y cada afirmación sobre
ese texto va con su cita. La skill consulta y cita: no tramita nada y no sustituye el asesoramiento de un profesional.

## Protocolo

Sigue los cinco pasos en orden. Cada operación se pide de una de dos formas, que devuelven el mismo sobre —`fuente`,
`url`, `fecha_consulta` y `hash` junto a `data`, sin los que no hay cita—:

- **Con su herramienta**, si entre las tuyas hay una con su nombre —`boe_articulo`, `graph_check`…—, solo o detrás del
  prefijo que le ponga tu agente, como `mcp__kitlegal__boe_articulo`. Sus argumentos van por su nombre (`norma`,
  `bloque`, `bloques`, `texto`, `fecha`). Si la tienes, úsala siempre.
- **Con su orden**, `kitlegal <applet> <verbo> … --json`, si no la tienes.
- Los pasos dan las dos. Donde encadenan dos órdenes con `&&`, con herramientas son dos llamadas seguidas, y la segunda
  solo si la primera no falló. Y donde dicen que una orden termina con un código, con herramientas el resultado viene
  marcado como error y su `data.clase` dice cuál («Comandos»).

### 1. Identificar la norma

Antes de consultar nada, lee `references/normas.md` y localiza en esa tabla la norma de la pregunta por su nombre, por
su número y año («Ley 40/2015») o por su abreviatura («LRJSP»). La tabla recoge las normas que esta skill consulta a
menudo; no es exhaustiva.

### 2. Resolver `BOE-A-…`

- Si la norma está en `references/normas.md`, toma de ahí su identificador `BOE-A-…`.
- Si no está, búscala por las palabras de su título, con `boe_buscar` (en `texto`) o con su orden, como
  `kitlegal boe buscar régimen jurídico del sector público --json`. Elige entre los resultados por título y rango, y di
  en la respuesta qué norma elegiste. Si hay varias posibles —una ley y su texto refundido, una ley y el reglamento que
  la desarrolla, una norma estatal y otra autonómica de título parecido—, di cuáles y por qué eliges una; si la pregunta
  no permite elegir, pregunta o responde de ambas distinguiéndolas.
- Una norma autonómica consolidada en el BOE se resuelve igual: si la búsqueda la encuentra por su título, se lee y se
  cita como una estatal. Si no da la norma, reformúlala con otras palabras del título; que no aparezca no prueba que no
  exista (regla 1).

### 3. Leer índice y bloques con `kitlegal boe`

- Si no conoces el id del bloque —saber el número del artículo no basta—, lee el índice de la norma, con `boe_indice`
  o con su orden, como `kitlegal boe indice BOE-A-2015-10565 --json`.
- Copia el id de la entrada del índice cuyo `titulo` es el artículo que buscas; nunca lo compongas a partir del número
  del artículo, porque en muchas normas los ids no son `a<número>`. En la Ley 9/2017, la entrada con `titulo`
  «Artículo 118» tiene el id `a1-30`, y `a118` no está en su índice.
- Sin fecha en la pregunta —con ella, «Redacción en una fecha»—, lee los bloques de uno en uno con
  `kitlegal boe articulo`, con `kitlegal graph check` de la misma norma y el mismo bloque detrás, en la misma orden:
  ```bash
  kitlegal boe articulo BOE-A-2015-10565 a21 --json && kitlegal graph check BOE-A-2015-10565 a21 --json
  ```
  En PowerShell, que en su versión 5.1 no tiene `&&`, la misma orden es:
  ```powershell
  kitlegal boe articulo BOE-A-2015-10565 a21 --json; if ($LASTEXITCODE -eq 0) { kitlegal graph check BOE-A-2015-10565 a21 --json }
  ```
  Con herramientas, `graph_check` va con la misma `norma` y ese bloque en `bloques`. Devuelve dos sobres, el del
  bloque, con su texto y sus avisos de vigencia, y el de `kitlegal graph check`, que da a la respuesta lo que dice
  «Redacción modificada» o, si termina con otro código que `0`, la regla 7. Va siempre así, detrás de la lectura, en su
  misma orden y con su misma norma y sus mismos bloques: no lo pidas nunca sin argumentos ni antes de leer, y si un
  bloque se ha leído sin él, pídelo a continuación con esa norma y ese bloque.
- Usa `kitlegal boe articulos` (`boe_articulos`), que los devuelve en el orden pedido, solo cuando necesites varios
  bloques a la vez y todos salgan del índice, con `kitlegal graph check` (`graph_check`) de esos mismos bloques detrás:
  `kitlegal boe articulos <norma> <bloques>... --json && kitlegal graph check <norma> <bloques>... --json`, y en PowerShell
  `kitlegal boe articulos <norma> <bloques>... --json; if ($LASTEXITCODE -eq 0) { kitlegal graph check <norma> <bloques>... --json }`.
- Sin fecha, lee cada bloque una sola vez por pregunta: una segunda lectura apagaría su línea `⚠ REDACCIÓN MODIFICADA:`.
- Una lectura de varios bloques falla entera en cuanto falla uno de ellos. Si termina con `4` o `5`, pide cada bloque
  por separado con `kitlegal boe articulo` (`boe_articulo`) antes de dar ninguno por no consultado: el fallo de un
  bloque no impide leer los demás.
- Sigue las remisiones que hagan falta para responder: si el bloque remite a otro artículo, de la misma norma o de
  otra, lee también el bloque remitido, resolviendo antes la otra norma con los pasos 1 y 2.
- Si la pregunta depende de la vigencia de la norma o de sus modificaciones, lee sus metadatos y su análisis, con
  `boe_metadatos` y `boe_analisis` o con sus órdenes, como `kitlegal boe analisis BOE-A-2017-12902 --json`.
- No pidas nunca un id de bloque que no salga del índice o de la propia pregunta. Si `kitlegal boe` termina con `3`
  (no encontrado), vuelve al índice en lugar de probar otros ids; si el artículo no existe en la norma, dilo.

### 4. Evaluar si falta contexto

- **Remisiones**: si el bloque remite a otro precepto o a otra norma que cambia la respuesta, léelo (paso 3).
- **Vigencia**: si la lectura trae avisos (derogada, vigencia agotada, consolidación no finalizada) o la fecha de
  vigencia del bloque no encaja con la situación preguntada, tenlo en cuenta y trasládalo (regla 3).
- **Modificaciones**: si una norma posterior cambió el bloque y eso importa, lee sus metadatos o su análisis (paso 3).
- Lo que no puedas leer, no lo suplas (paso 5); lo que decidas aquí no va en la respuesta: quien pregunta no lo ve.

### 5. Responder citando

- **La respuesta empieza por lo que se pregunta y habla de la norma, no de la consulta.** La respuesta es todo lo que
  escribes después de la última orden o llamada, desde su primera palabra: quien pregunta lo lee entero, y no ve las
  órdenes ni las llamadas ni lo que devuelven. Le sirven la norma, su texto, su cita, sus avisos de vigencia y las
  líneas `⚠` de cada bloque, y de eso está hecha. Escribe como quien tiene el BOE delante: «el artículo no lleva avisos
  de vigencia», y no «la lectura no trajo avisos»; la norma, por su nombre y su identificador, sin decir de qué tabla
  o de qué orden sale. Lo que has hecho, lo que vas a hacer y lo que ha devuelto o no una orden no le sirven.
- **Nada de otra conversación.** No sabes qué se preguntó ni qué se respondió en otra conversación: no hables de ello,
  ni para afirmarlo, ni para confirmarlo, ni para desmentirlo. De lo leído en otras conversaciones, la respuesta solo
  lleva la línea `⚠ REDACCIÓN MODIFICADA:` de cada bloque que la trae (más en «Redacción modificada»).
- Cada afirmación sobre el contenido de una norma lleva su cita, y lo citado sale del texto que devolvió `kitlegal boe`,
  por su orden o por su herramienta, en esta conversación. La cita es la forma legible de la norma y del bloque seguida,
  en la misma línea, de `[<identificador>, bloque <id>]`, tal como los da la fuente («Cómo se cita»).
- **De un precepto que no has leído, nada.** De un artículo, un apartado o una disposición que nada te ha devuelto en
  esta conversación no digas qué dice, de qué trata ni cuándo o cómo se aplica, ni en un paréntesis o un inciso, ni
  aunque creas saberlo: sería texto legal sin fuente; el índice tampoco lo dice, da el número de cada bloque y nada
  más. Su número y su materia no van juntos en ninguna frase, ni afirmando («el art. N, que es…», «el art. N sobre
  <materia>», «los artículos que regulan <materia> (N y M)»), ni negando lo que puedes decir de él («no puedo decir
  qué <regla> fija»), ni al ofrecer leerlo, ni al avisar al final de lo que no has leído: o el número solo, o la
  materia sola. Para decir de qué trata lo remitido, léelo y cítalo; si no, la remisión va con las palabras del texto
  leído y nada detrás. Puedes avisar de que una materia se regula en otra parte, sin nombrar precepto ni regla.
- **Distingue ley y reglamento**: cuando cites normas de rango distinto, di el rango de cada una —el `rango` de
  `references/normas.md` o de la búsqueda— y recuerda que la ley prevalece sobre el reglamento que la desarrolla.
- **Señala la variación autonómica**: cuando lo preguntado pueda variar por normativa autonómica (competencias
  compartidas o cedidas, desarrollo autonómico, régimen foral), dilo; y cuando corresponda a ordenanzas u otras normas
  locales, di que no están en esta fuente.
- Traslada cada aviso de vigencia de la lectura con su forma fija («Cómo se cita», regla 3). De la vigencia del bloque,
  la respuesta dice sus avisos y, de la redacción leída, qué norma la dio y desde cuándo rige. Hasta cuándo, nunca:
  ninguna lectura trae el fin de una redacción, tampoco en una norma derogada, cuyo aviso no lleva fecha, y darlo sería
  texto legal sin fuente. `kitlegal graph check` no dice nada de la vigencia.

## Cómo se cita

Cada cita lleva la forma legible de la norma y del bloque y, entre corchetes, el identificador de la norma y el id del
bloque tal como los da la fuente; repasa cada cita de la respuesta con estas reglas. La forma recomendada pone la forma
legible delante del corchete:

```text
art. 21 de la Ley 39/2015 [BOE-A-2015-10565, bloque a21]
```

- Lo que hace cita es que los corchetes terminen en `<identificador>, bloque <id de bloque>]`: el identificador
  `BOE-A-…`, una coma, la palabra `bloque` y el id, sin nada entre el id y el corchete de cierre, y los dos corchetes en
  la misma línea.
- Si dentro de los corchetes va además la forma legible, va delante del identificador y separada de él por una coma:
  `[art. 20.1 de la LTAIBG, BOE-A-2013-12887, bloque a20]` también es una cita. Detrás del id no va nada:
  `[BOE-A-2015-10565, bloque a21, art. 21]` no es una cita, y tampoco lo son el identificador sin el id del bloque ni
  el identificador y el id sin corchetes.
- La regla vale igual cuando la cita va sola en una línea o debajo de una cita textual en bloque, como tras transcribir
  el artículo: `art. 140 de la Constitución Española [BOE-A-1978-31229, bloque a140]`. El identificador y el id van tal
  como los devuelve `kitlegal boe`, también cuando el id termina en punto: el corchete de cierre lo delimita.
- Una cita por bloque. Un bloque remitido se cita por separado, con su norma y su id.

Los avisos de vigencia también tienen forma fija. Cada aviso de la lectura va en la respuesta con `⚠`, la etiqueta del
aviso tal como la da el binario —lo que su `texto` lleva entre `⚠` y los dos puntos— y dos puntos, seguidos de la frase
del binario o de una explicación: `⚠ NORMA DEROGADA:` para el aviso `derogada`, `⚠ VIGENCIA AGOTADA:` para
`vigencia-agotada` y `⚠ TEXTO POSIBLEMENTE DESACTUALIZADO:` para `consolidacion-no-finalizada`:
```text
⚠ NORMA DEROGADA: esta norma ha sido derogada.
```

**Por qué las líneas `⚠` tienen forma fija.** Cada línea `⚠` de esta skill —un aviso, la redacción en una fecha, la
redacción modificada— la lee un programa sin leer el resto de la respuesta, y comprueba en ella su etiqueta, su cita y
los datos de la lectura que lleva. Lo que una línea dice solo se comprueba en ella: dicho con otras palabras, la línea
falta. Por eso va entera, en una sola línea, con su forma y sus datos tal como los da la lectura.

## Redacción en una fecha

Si la pregunta lleva una fecha —un día, un año o un mes, o un hecho cuyo día dan la pregunta o la conversación—, cita
la redacción vigente ese día, no la de hoy: lee cada bloque con `--fecha AAAA-MM-DD` detrás, como en
`kitlegal boe articulo BOE-A-1985-5392 a22 --fecha 2005-06-15 --json`, o con `boe_articulo` y `fecha`, y sin
`kitlegal graph check`. Si un hecho sitúa el momento y nadie da su día, pide el día: no leas ni cites ninguna redacción
como la de entonces. Con un año o un mes sin día, lee con su primer y con su último día: si las dos lecturas dan la
misma fecha de vigencia y la misma norma que la dio, o las dos el `3` de abajo, responde con la línea de cada una; si
no, no cites ninguna y pide el día.
- Lo leído así se cita como en «Cómo se cita», con sus avisos, que son los de la norma hoy (regla 3), y con esta línea,
  con la fecha pedida y la de vigencia de lo citado tal como las da la lectura:
  ```text
  ⚠ REDACCIÓN A FECHA: <forma legible> [<identificador>, bloque <id>]: se cita la redacción vigente el AAAA-MM-DD, la de fecha de vigencia AAAAMMDD.
  ```
- Si el `data` de la lectura trae `redaccion_posterior`, hoy la redacción es otra, y va también esta línea, con la
  norma y la fecha de vigencia de `redaccion_posterior`:
  ```text
  ⚠ REDACCIÓN DE HOY: <forma legible> [<identificador>, bloque <id>]: hoy la redacción es otra, la de <identificador>, con fecha de vigencia AAAAMMDD.
  ```
  De la redacción de hoy, la lectura da su norma y su fecha, y esta línea es el único sitio de la respuesta en que van:
  «hay otra redacción, dada por <identificador>, de <fecha>» no la sustituye, y repetido junto a ella no añade nada.
  Qué dice, qué cambió o si rige ya no lo has leído, salvo que la leas sin fecha (paso 3). De la citada, la prosa dice
  qué norma la dio y desde cuándo rige; hasta cuándo, nunca: entre las dos puede haber otras.
- Si la lectura termina con `3` y su mensaje da la fecha de vigencia de la primera redacción, ese día el precepto no
  tenía redacción vigente: no cites su texto como el de entonces ni digas qué regía sin haberlo leído, y responde con
  esta línea, con la fecha del mensaje; con otro `3`, la regla 2.
  ```text
  ⚠ SIN REDACCIÓN VIGENTE: <forma legible> [<identificador>, bloque <id>]: el AAAA-MM-DD no tenía redacción vigente; la tiene desde el AAAAMMDD, la fecha de vigencia de su primera redacción.
  ```
- Estas tres líneas salen solo de una lectura con fecha, y cada una con su forma fija («Cómo se cita»).

## Redacción modificada

`kitlegal` recuerda en local los bloques que ha leído sin fecha y qué redacción vio cada vez.
`kitlegal graph check <norma> <bloques>... --json` (`graph_check`) compara, para esa norma y esos bloques, la redacción
que acabas de leer con la que se leyó la vez anterior, en otra conversación que quien pregunta no conoce, y devuelve en
`data.hallazgos` una entrada por cada bloque en que encuentra algo, con su `clase`. De `kitlegal graph`, el protocolo
solo usa `check`, tras cada lectura sin fecha (paso 3).

- Una entrada de `clase` `version-obsoleta` dice que la redacción que acabas de leer de ese bloque no es la que se leyó
  la vez anterior. La respuesta lo dice con una línea por bloque, con esta forma fija, la cita del bloque como en «Cómo
  se cita» y las dos fechas de vigencia de la entrada (`AAAAMMDD`): la superada y después la de lo que acabas de leer.
  ```text
  ⚠ REDACCIÓN MODIFICADA: <forma legible> [<identificador>, bloque <id>]: la redacción con fecha de vigencia AAAAMMDD, la que se consultó antes, ha sido sustituida por la de AAAAMMDD, que es la que se cita.
  ```
  Con dos bloques en `version-obsoleta`, dos líneas, cada una con su cita y sus fechas. Decirlo con otras palabras no
  vale, ni en lugar de la línea ni además de ella: la línea va con su forma fija y lo dice entera.
- Una entrada de `clase` `fuente-caducada` no va en la respuesta, que cita el texto que acabas de leer.
- **No es un aviso de vigencia.** Un aviso es un dato de la norma, que el BOE da con el bloque, y la etiqueta
  `REDACCIÓN MODIFICADA` no es la de ninguno: una entrada de `kitlegal graph check` es un dato de kitlegal sobre otras
  conversaciones, que quien pregunta no conoce. Por eso un bloque sin entrada de `clase` `version-obsoleta` no da nada
  a la respuesta: ni una línea, ni una frase junto a los avisos o a la fecha de vigencia, ni una palabra al final.
  Mencionarlo sería contar lo que devolvió una orden y hablar de otras conversaciones: la línea es todo lo que la
  respuesta dice de ellas y de `kitlegal graph check`.
- **La redacción superada no la has leído.** Ni la lectura sin fecha, que da solo la redacción de hoy, ni
  `kitlegal graph check`, que da dos fechas, dicen qué decía la superada ni en qué se diferencia de la vigente: la
  respuesta no lo dice, ni lo resume, ni lo compara, aunque creas saberlo, que sería texto legal sin fuente. Si quien
  pregunta qué cambió, la respuesta dice que cita la redacción vigente y que la de antes no la ha leído.

## Comandos

`kitlegal` se invoca desde el `PATH`, y cada herramienta, por el nombre de su fila. Cada orden termina con `0` si todo
fue bien, y si no, con el código de lo que falló; el error de una herramienta lo dice en `data.clase`: `2` o
`argumentos`, con argumentos inválidos; `3` o `no-encontrado`, si no encuentra lo pedido; `4` o `fuente-no-disponible`,
si la fuente no está disponible; `5` o `limite-o-tos`, si la fuente limita el ritmo; `6` o `identidad-humana`, si hace
falta la identidad de una persona; y `1` o `inesperado`, ante un fallo inesperado (por ejemplo, un `world.db` que no
se puede leer en `kitlegal graph`).

<!-- inicio de la tabla de comandos: generada desde --describe con make skills-sync, no editar -->

### `kitlegal boe`

| Orden | Herramienta | Qué hace | Qué devuelve en `data` |
|---|---|---|---|
| `kitlegal boe buscar <texto>...` | `boe_buscar` | Busca normas consolidadas por las palabras de su título o con una consulta de la fuente. | lista de objetos con `identificador`, `titulo`, `rango`, `vigencia_agotada`, `estado_consolidacion`, `url` |
| `kitlegal boe indice <norma>` | `boe_indice` | Devuelve los bloques de una norma consolidada, en el orden de la fuente. | objeto con `norma`, `url`, `bloques` |
| `kitlegal boe articulo <norma> <bloque> [--fecha <fecha>]` | `boe_articulo` | Devuelve el texto vigente de un bloque de una norma, con los avisos de su vigencia; con --fecha AAAA-MM-DD, la redacción vigente ese día. | objeto con `norma`, `bloque`, `titulo`, `tipo`, `fecha_version`, `fecha_vigencia`, `norma_modificadora`, `texto`, `hash_texto`, `avisos`, `url`, `url_eli`, `fecha_pedida`, `redaccion_posterior` |
| `kitlegal boe articulos <norma> <bloques>...` | `boe_articulos` | Devuelve el texto vigente de varios bloques de una norma, en el orden pedido. | lista de objetos con `norma`, `bloque`, `titulo`, `tipo`, `fecha_version`, `fecha_vigencia`, `norma_modificadora`, `texto`, `hash_texto`, `avisos`, `url`, `url_eli` |
| `kitlegal boe metadatos <norma>` | `boe_metadatos` | Devuelve los datos de una norma y los avisos de su vigencia. | objeto con `norma`, `titulo`, `rango`, `numero_oficial`, `fecha_disposicion`, `fecha_publicacion`, `fecha_vigencia`, `estatus_derogacion`, `vigencia_agotada`, `estado_consolidacion`, `url_eli`, `avisos` |
| `kitlegal boe analisis <norma>` | `boe_analisis` | Devuelve las materias, las notas y las referencias de una norma. | objeto con `norma`, `materias`, `notas`, `referencias` |

### `kitlegal graph`

| Orden | Herramienta | Qué hace | Qué devuelve en `data` |
|---|---|---|---|
| `kitlegal graph show <id>` | `graph_show` | Devuelve un nodo del grafo del mundo con sus aristas y su procedencia, sin texto legal. | objeto con `nodo`, `salientes`, `entrantes` |
| `kitlegal graph stats` | `graph_stats` | Cuenta los nodos, las aristas y los textos del grafo del mundo por tipo, relación y fuente. | objeto con `nodos`, `aristas`, `textos`, `nodos_por_tipo`, `aristas_por_relacion` |
| `kitlegal graph check [<norma> [<bloques>...]]` | `graph_check` | Comprueba lo consultado de una norma, de algunos de sus bloques o, sin argumentos, todo lo consultado, y lista como mucho 50 hallazgos: redacciones que han cambiado desde la lectura anterior y consultas caducadas. | objeto con `norma`, `bloques`, `version-obsoleta`, `fuente-caducada`, `omitidos`, `hallazgos` |

La orden y la herramienta de cada fila devuelven el mismo sobre: `ok`, `fuente`, `url`, `fecha_consulta`, `hash`, `data`; con `ok` falso, `data` lleva `clase` y `mensaje`.

Banderas comunes: `--json`, `--timeout <valor>`, `--offline`, `--dry-run`, `--describe`, `--no-graph`, `--asunto <valor>`, `--verbose`.

<!-- fin de la tabla de comandos -->

## Reglas

1. **No concluir que algo no existe.** Una búsqueda vacía, o que una norma no esté en `references/normas.md`, no prueba
   que la norma o la regulación no existan: di «no encontrada con esta búsqueda» y propón reformular la búsqueda.
2. **Nunca inventar contenido legal.** Si `kitlegal boe` falla —termina con `3` (no encontrado), `4` (fuente no
   disponible) o `5` (límite de ritmo), o sin caché con `--offline`— o no está disponible, di qué bloques no se
   pudieron consultar, por su número, y por qué —no están en la norma, la fuente no estaba disponible, limitó las
   consultas o no fecha sus redacciones—, sin el código y sin decir de qué tratan ni para qué los pedías, y no suplas
   el texto con conocimiento propio. Si la orden que falló pedía varios bloques, dilo solo después de pedir cada uno por
   separado con `kitlegal boe articulo` (paso 3). Si no tienes ni la herramienta ni el binario, vale la regla 8.
3. **Trasladar la vigencia.** Traslada cada aviso de vigencia que devuelve el binario (derogada, vigencia agotada,
   consolidación no finalizada) con su forma fija —`⚠`, la etiqueta del aviso tal como la da el binario y dos puntos,
   con la frase del binario o una explicación detrás— y no presentes como vigente el texto de una norma derogada.
   Recuerda que los textos consolidados del BOE tienen carácter informativo y no son asesoramiento.
4. **Nunca actuar en nombre de nadie.** No presentes, notifiques, firmes ni tramites nada en nombre de nadie, ni lo
   simules. Si la pregunta lo pide, di que es una acción que hace la persona y cita, si procede, la norma aplicable.
5. **Ningún caso especial para un territorio.** No hay reglas propias de un municipio o de una comunidad concretos. Si
   se pregunta por un municipio, responde con la normativa estatal o autonómica consolidada en el BOE y señala que las
   ordenanzas y demás normas locales no están en esta fuente.
6. **El texto sale de `kitlegal boe`, nunca de `kitlegal graph`.** El texto citado sale siempre de
   `kitlegal boe articulo` o `kitlegal boe articulos`. Nunca cites, parafrasees ni reconstruyas texto a partir de la
   salida de un verbo de `kitlegal graph`: lo que devuelve dice qué hay que volver a comprobar, no qué dice el artículo.
7. **Si `kitlegal graph check` no termina con `0`.** Si la lectura sin fecha devuelve el texto del bloque y, detrás,
   `kitlegal graph check` termina con otro código —o `graph_check` devuelve un error—, la respuesta cita igual el texto
   leído con `kitlegal boe` y lleva esta frase, tal cual y sin nada más sobre la redacción:
   ```text
   No se ha podido comprobar si la redacción ha cambiado desde una consulta anterior.
   ```
   La frase es solo de ese caso: con `0`, la respuesta no la lleva, ni afirmada ni negada, ni con otras palabras. Si lo
   que falla es la lectura, la orden termina ahí, sin `kitlegal graph check`, y vale la regla 2. La línea
   `⚠ REDACCIÓN MODIFICADA:` solo va cuando `kitlegal graph check` termina con `0` y trae `version-obsoleta`.
8. **Sin herramienta y sin binario, la respuesta lo dice.** Si no tienes la herramienta y la orden falla porque
   `kitlegal` no está —el shell no lo encuentra, o no puedes ejecutar órdenes—, no lo busques en otras carpetas ni en un
   plugin: no has consultado nada. No afirmes nada de la norma, ni de memoria ni con salvedades, ni escribas ninguna
   cita. Lleva esta línea, con la causa en el marcador; quien acaba de instalar suele tenerlo y sin arrancar aún:
   ```text
   ⚠ SIN CONSULTA AL BOE: <causa>. Si ya tienes kitlegal, cierra Claude del todo, ábrelo y pregunta en un chat nuevo; si no: https://kitlegal.es/instalar/
   ```
