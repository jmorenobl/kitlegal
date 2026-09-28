---
name: boe-legislacion
description: >-
  Consulta y cita normativa consolidada del Boletín Oficial del Estado (BOE) de cualquier materia: procedimiento
  administrativo, contratación pública, régimen local, tributos y haciendas locales, régimen jurídico del sector
  público, transparencia, relaciones laborales… Úsala cuando se pregunte qué dice un artículo, una ley o un real
  decreto; cuando se nombre una norma por su número y año (Ley 39/2015, Real Decreto Legislativo 2/2004), por su
  abreviatura (LPAC, LCSP, LRBRL, LGT, TRLRHL, LRJSP) o por su identificador BOE-A-…; o cuando haya que citar el texto
  vigente de una norma estatal o autonómica consolidada en el BOE. Lee el índice y los artículos con el binario
  kitlegal y responde citando identificador y bloque.
metadata:
  kitlegal-applets: boe graph
  kitlegal-referencias: normas
---

# Consultar y citar legislación consolidada del BOE

Esta skill responde preguntas sobre el contenido de normas consolidadas del Boletín Oficial del Estado —la
Constitución, leyes orgánicas y ordinarias, reales decretos legislativos, reales decretos y las normas autonómicas que
el BOE consolida— de cualquier ámbito: procedimiento administrativo, contratación pública, régimen local, tributos,
transparencia, relaciones laborales… Lo que dice de una norma sale del texto que devuelve `kitlegal boe` en la misma
conversación, y cada afirmación sobre ese texto va con su cita. La skill consulta y cita: no tramita nada y no sustituye
el asesoramiento de un profesional.

## Protocolo

Sigue los cinco pasos en orden. Escribe todas las órdenes con `--json`: el sobre trae entonces `fuente`, `url`,
`fecha_consulta` y `hash` junto a `data`, y sin ellos no hay cita.

### 1. Identificar la norma

Antes de consultar nada, lee `references/normas.md` y localiza en esa tabla la norma de la pregunta por su nombre, por
su número y año («Ley 40/2015») o por su abreviatura («LRJSP»). La tabla recoge las normas que esta skill consulta a
menudo; no es exhaustiva.

### 2. Resolver `BOE-A-…`

- Si la norma está en `references/normas.md`, toma de ahí su identificador `BOE-A-…`.
- Si no está, búscala con `kitlegal boe buscar` por las palabras de su título:

  ```bash
  kitlegal boe buscar régimen jurídico del sector público --json
  ```

  Elige entre los resultados por título y rango, y di en la respuesta qué norma elegiste. Si hay varias posibles —una
  ley y su texto refundido, una ley y el reglamento que la desarrolla, una norma estatal y otra autonómica de título
  parecido—, di cuáles y por qué eliges una; si la pregunta no permite elegir, pregunta o responde de ambas
  distinguiéndolas.
- Una norma autonómica consolidada en el BOE se resuelve igual: si `kitlegal boe buscar` la encuentra por su título, se
  lee y se cita como una estatal.
- Si la búsqueda no da la norma, reformúlala con otras palabras del título; que no aparezca no prueba que no exista
  (regla 1).
- Resuelto el `BOE-A-…` de la norma de la pregunta, y antes de leer ninguno de sus bloques, comprueba la memoria de
  consultas (más en «Memoria de consultas»):

  ```bash
  kitlegal graph check --json
  ```

### 3. Leer índice y bloques con `kitlegal boe`

- Si no conoces el id del bloque —saber el número del artículo no basta—, lee el índice de la norma:

  ```bash
  kitlegal boe indice BOE-A-2015-10565 --json
  ```

- Copia el id de la entrada del índice cuyo `titulo` es el artículo que buscas; nunca lo compongas a partir del número
  del artículo, porque en muchas normas los ids no son `a<número>`. En la Ley 9/2017, la entrada con `titulo`
  «Artículo 118» tiene el id `a1-30`, y `a118` no está en su índice.
- Lee los bloques de uno en uno con `kitlegal boe articulo`:

  ```bash
  kitlegal boe articulo BOE-A-2015-10565 a21 --json
  ```

  Usa `kitlegal boe articulos`, que los devuelve en el orden pedido, solo cuando necesites varios bloques a la vez y
  todos salgan del índice.
- Una orden de `kitlegal boe articulos` falla entera en cuanto falla uno de sus bloques. Si una orden con varios bloques
  termina con el código 4 o 5, pide cada bloque por separado con `kitlegal boe articulo` antes de dar ninguno por no
  consultado: el fallo de un bloque no impide leer los demás.

- Sigue las remisiones que hagan falta para responder: si el bloque remite a otro artículo, de la misma norma o de
  otra, lee también el bloque remitido, resolviendo antes la otra norma con los pasos 1 y 2.
- Si la pregunta depende de la vigencia de la norma o de sus modificaciones, lee sus metadatos y su análisis:

  ```bash
  kitlegal boe metadatos BOE-A-2017-12902 --json
  kitlegal boe analisis BOE-A-2017-12902 --json
  ```

- No pidas nunca un id de bloque que no salga del índice o de la propia pregunta. Si `kitlegal boe` termina con el
  código 3 (no encontrado), vuelve al índice en lugar de probar otros ids; si el artículo no existe en la norma, dilo.

### 4. Evaluar si falta contexto

Antes de responder, comprueba si lo leído basta:

- **Remisiones**: si el bloque remite a otro artículo, a otra ley o a un reglamento que cambia la respuesta, léelo
  (paso 3).
- **Vigencia**: si el sobre trae avisos (derogada, vigencia agotada, consolidación no finalizada) o la fecha de
  vigencia del bloque no encaja con la situación preguntada, tenlo en cuenta y trasládalo (regla 3).
- **Modificaciones**: si una norma posterior cambió el bloque (`norma_modificadora`) de un modo que importa para la
  pregunta, consulta `kitlegal boe metadatos` o `kitlegal boe analisis`.

Si falta algo que no puedes leer con `kitlegal boe`, dilo en la respuesta en lugar de suplirlo.

### 5. Responder citando

- Cuando ya no quede nada por leer, y antes de redactar la respuesta, vuelve a ejecutar `kitlegal graph check --json`
  y traslada lo que encuentren las dos comprobaciones como dice «Memoria de consultas».
- Cada afirmación sobre el contenido de una norma lleva su cita, y lo citado sale del texto que devolvió `kitlegal boe`
  en esta conversación. La cita es la forma legible de la norma y del bloque seguida, en la misma línea, de
  `[<identificador>, bloque <id>]`. Lo que la hace cita es que los corchetes terminen en
  `<identificador>, bloque <id>]`, con el identificador `BOE-A-…` y el id tal como los da la fuente (más en «Cómo se
  cita»).
- **Distingue ley y reglamento**: cuando cites normas de rango distinto, di el rango de cada una —el `rango` de
  `references/normas.md` o de la búsqueda— y recuerda que la ley prevalece sobre el reglamento que la desarrolla.
- **Señala la variación autonómica**: cuando lo preguntado pueda variar por normativa autonómica (competencias
  compartidas o cedidas, desarrollo autonómico, régimen foral), dilo; y cuando corresponda a ordenanzas u otras normas
  locales, di que no están en esta fuente.
- Traslada cada aviso de vigencia del sobre con su forma fija: `⚠`, la etiqueta del aviso tal como la da el binario y
  dos puntos, seguidos de la frase del binario o de una explicación (más en «Cómo se cita»). Recuerda que los textos
  consolidados del BOE tienen carácter informativo.
- Antes de responder, repasa cada cita: sus corchetes se abren y se cierran en la misma línea y terminan en
  `<identificador>, bloque <id>]`, con la palabra `bloque` y nada entre el id y el corchete de cierre. Si dentro de los
  corchetes va además la forma legible, va delante del identificador.

## Cómo se cita

Cada cita lleva la forma legible de la norma y del bloque y, entre corchetes, el identificador de la norma y el id del
bloque tal como los da la fuente. La forma recomendada pone la forma legible delante del corchete:

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
  el artículo: `art. 140 de la Constitución Española [BOE-A-1978-31229, bloque a140]`.
- El identificador y el id van tal como los devuelve `kitlegal boe`, también cuando el id termina en punto: el corchete
  de cierre lo delimita.
- Una cita por bloque. Un bloque remitido se cita por separado, con su norma y su id.

Los avisos de vigencia también tienen forma fija. Cada aviso del sobre va en la respuesta con `⚠`, la etiqueta del
aviso tal como la da el binario —lo que su `texto` lleva entre `⚠` y los dos puntos— y dos puntos, seguidos de la frase
del binario o de una explicación:

- `⚠ NORMA DEROGADA:` para el aviso `derogada`.
- `⚠ VIGENCIA AGOTADA:` para el aviso `vigencia-agotada`.
- `⚠ TEXTO POSIBLEMENTE DESACTUALIZADO:` para el aviso `consolidacion-no-finalizada`.

```text
⚠ NORMA DEROGADA: esta norma ha sido derogada.
```

- La etiqueta va entera y sin cambiar ninguna palabra, con `⚠` delante y los dos puntos detrás, todo en la misma línea.
  Decir con otras palabras que la norma está derogada no traslada el aviso.

## Memoria de consultas

`kitlegal` recuerda en local las normas y los bloques que ha leído con `kitlegal boe articulo` o `articulos`: cada
redacción, con su fecha de vigencia, y cuándo la consultó. `kitlegal graph check` repasa esa memoria y devuelve en
`data` una lista de hallazgos, cada uno con su `clase` y una `explicacion` que nombra la norma (`BOE-A-…`) o el bloque
(`[BOE-A-…, bloque <id>]`) del que habla. De los verbos de `kitlegal graph`, el protocolo solo usa `check`, dos veces:
resuelto el `BOE-A-…` y antes de leer (paso 2), y cuando ya no queda nada por leer y antes de responder (paso 5).

Al responder, reúne los hallazgos de las dos comprobaciones cuya `explicacion` nombra el `BOE-A-…` de la norma de la
pregunta, solo o en la cita de un bloque de esa norma que has leído para responder, y agrúpalos por `clase`. Di cada
clase presente una sola vez, aunque lleguen varios hallazgos de la misma clase o la misma clase en las dos
comprobaciones:

- `version-obsoleta`: di que la redacción ha cambiado respecto de la consultada antes, con las fechas de vigencia que
  traen los hallazgos: la de cada redacción superada (`fecha_vigencia`) y la de la más reciente
  (`fecha_vigencia_reciente`).
- `fuente-caducada`: di que la consulta anterior había caducado y que la respuesta se apoya en la lectura nueva.

Los demás hallazgos —los de otras normas, también las que hayas leído por una remisión, y los de bloques que no has
leído para responder— no se trasladan. Sin hallazgos de la norma de la pregunta, no hay nada que decir de la memoria de
consultas. Un hallazgo no es un aviso de vigencia: no lleva la forma fija de los avisos.

## Comandos

`kitlegal` se invoca desde el `PATH`. Códigos de salida: 0 correcto, 2 argumentos inválidos, 3
no encontrado, 4 fuente no disponible, 5 límite de ritmo de la fuente, 6 requiere identidad humana, 1 fallo inesperado
(por ejemplo, un `world.db` que no se puede leer en `kitlegal graph`).

<!-- inicio de la tabla de comandos: generada desde --describe con make skills-sync, no editar -->

### `kitlegal boe`

| Orden | Qué hace | Qué devuelve en `data` |
|---|---|---|
| `kitlegal boe buscar <texto>...` | Busca normas consolidadas por las palabras de su título o con una consulta de la fuente. | lista de objetos con `identificador`, `titulo`, `rango`, `vigencia_agotada`, `estado_consolidacion`, `url` |
| `kitlegal boe indice <norma>` | Devuelve los bloques de una norma consolidada, en el orden de la fuente. | objeto con `norma`, `url`, `bloques` |
| `kitlegal boe articulo <norma> <bloque>` | Devuelve el texto vigente de un bloque de una norma, con los avisos de su vigencia. | objeto con `norma`, `bloque`, `titulo`, `tipo`, `fecha_version`, `fecha_vigencia`, `norma_modificadora`, `texto`, `hash_texto`, `avisos`, `url`, `url_eli` |
| `kitlegal boe articulos <norma> <bloques>...` | Devuelve el texto vigente de varios bloques de una norma, en el orden pedido. | lista de objetos con `norma`, `bloque`, `titulo`, `tipo`, `fecha_version`, `fecha_vigencia`, `norma_modificadora`, `texto`, `hash_texto`, `avisos`, `url`, `url_eli` |
| `kitlegal boe metadatos <norma>` | Devuelve los datos de una norma y los avisos de su vigencia. | objeto con `norma`, `titulo`, `rango`, `numero_oficial`, `fecha_disposicion`, `fecha_publicacion`, `fecha_vigencia`, `estatus_derogacion`, `vigencia_agotada`, `estado_consolidacion`, `url_eli`, `avisos` |
| `kitlegal boe analisis <norma>` | Devuelve las materias, las notas y las referencias de una norma. | objeto con `norma`, `materias`, `notas`, `referencias` |

### `kitlegal graph`

| Orden | Qué hace | Qué devuelve en `data` |
|---|---|---|
| `kitlegal graph show <id>` | Devuelve un nodo del grafo del mundo con sus aristas y su procedencia, sin texto legal. | objeto con `nodo`, `salientes`, `entrantes` |
| `kitlegal graph stats` | Cuenta los nodos, las aristas y los textos del grafo del mundo por tipo, relación y fuente. | objeto con `nodos`, `aristas`, `textos`, `nodos_por_tipo`, `aristas_por_relacion` |
| `kitlegal graph check` | Comprueba el grafo del mundo y devuelve como hallazgos las versiones superadas y las consultas caducadas. | lista de objetos con `clase`, `id`, `explicacion`, `procedencia`, `fecha_vigencia`, `fecha_vigencia_reciente`, `vigencia_segundos` |

Todas devuelven el sobre `ok`, `fuente`, `url`, `fecha_consulta`, `hash`, `data`; con `ok` falso, `data` lleva `clase` y `mensaje`.

Banderas comunes: `--json`, `--timeout <valor>`, `--offline`, `--dry-run`, `--describe`, `--no-graph`, `--asunto <valor>`, `--verbose`.

<!-- fin de la tabla de comandos -->

## Reglas

1. **No concluir que algo no existe.** Una búsqueda vacía, o que una norma no esté en `references/normas.md`, no prueba
   que la norma o la regulación no existan: di «no encontrada con esta búsqueda» y propón reformular la búsqueda.
2. **Nunca inventar contenido legal.** Si `kitlegal boe` falla —código 3 (no encontrado), 4 (fuente no disponible) o 5
   (límite de ritmo), o sin caché con `--offline`— o no está disponible, di qué no se pudo consultar y no suplas el
   texto con conocimiento propio. Si la orden que falló pedía varios bloques, dilo solo después de haber pedido cada
   bloque por separado con `kitlegal boe articulo` (paso 3), y di cuáles no se pudieron consultar. Si `kitlegal` no
   está en el `PATH`, di que falta instalar kitlegal.
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
   salida de un verbo de `kitlegal graph`: la memoria de consultas dice qué hay que volver a comprobar, no qué dice el
   artículo.
7. **Una comprobación con hallazgos no es un fallo.** `kitlegal graph check` con código 0 es un resultado, con
   hallazgos o sin ellos, y nunca un fallo de la herramienta: trasládalos como dice «Memoria de consultas». Si termina
   con otro código, responde igual con el texto de `kitlegal boe` y di que no se ha podido comprobar la memoria de
   consultas.
