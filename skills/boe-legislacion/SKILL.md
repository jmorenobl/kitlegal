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
- Lee cada bloque una sola vez por pregunta: una segunda lectura del mismo bloque apagaría lo que la memoria de
  consultas tiene que decirte (más en «Memoria de consultas»).
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

- Cuando ya no quede nada por leer, y antes de redactar la respuesta, comprueba la memoria de consultas una vez por
  cada norma cuyos bloques vas a citar, con esa norma y los bloques de ella que has leído:

  ```bash
  kitlegal graph check BOE-A-2015-10565 a21 --json
  ```

  No la pidas nunca sin argumentos ni antes de leer. Si da `version-obsoleta`, dilo con la forma fija de «Memoria de
  consultas»; si no, no digas nada de ella.
- **La respuesta empieza por lo que se pregunta.** Quien pregunta no ve las órdenes que ejecutas ni lo que devuelven:
  le sirven la norma, su texto y su cita. No cuentes lo que has hecho ni lo que ha devuelto ninguna orden, tampoco para
  decir que no hay nada que decir ni para anunciar que vas a responder. Salvo la forma `⚠ REDACCIÓN MODIFICADA:`, la
  respuesta no nombra la memoria de consultas, `kitlegal graph` ni ninguno de sus verbos, los códigos de salida, los
  hallazgos, las clases del binario (`version-obsoleta`, `fuente-caducada`), el JSON ni el sobre.
- **Nada de otra conversación.** No sabes qué se preguntó ni qué se respondió en otra conversación: no hables de ello,
  ni para afirmarlo, ni para confirmarlo, ni para desmentirlo. Sin `version-obsoleta`, no digas nada de lo consultado
  antes, ni que ha cambiado ni que no: la comprobación sin hallazgos no distingue un bloque leído antes y sin cambios de
  uno que nunca se leyó.
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

`kitlegal` recuerda en local los bloques que ha leído con `kitlegal boe articulo` o `articulos` y qué redacción vio
cada lectura. `kitlegal graph check <norma> <bloques>... --json` devuelve en `data.hallazgos` los de esa norma y esos
bloques, cada uno con su `clase`. De los verbos de `kitlegal graph`, el protocolo solo usa `check`: una vez por norma
citada, cuando ya no queda nada por leer y antes de responder (paso 5).

- `version-obsoleta`: la redacción del bloque ha cambiado desde la lectura anterior. Trasládalo con su forma fija,
  `⚠ REDACCIÓN MODIFICADA:` —`⚠`, la etiqueta `REDACCIÓN MODIFICADA` y dos puntos—, y detrás, en la misma línea, las dos
  fechas de vigencia tal como las da el hallazgo (`AAAAMMDD`): la de la redacción superada (`fecha_vigencia`) y la de
  la que acabas de leer (`fecha_vigencia_reciente`). Por ejemplo:

  ```text
  ⚠ REDACCIÓN MODIFICADA: la redacción con fecha de vigencia 20161002, la que se consultó antes, ha sido sustituida por la de 20250101, que es la que se cita.
  ```

  Decirlo con otras palabras no lo traslada.
- `fuente-caducada` no se traslada: la respuesta cita el texto que acabas de leer, que la caché no sirve pasada su
  vigencia.
- Sin `version-obsoleta`, la respuesta no dice nada de la memoria de consultas (paso 5); si `kitlegal graph check`
  termina con otro código, la regla 7.

La etiqueta de `version-obsoleta` no es la de ningún aviso de vigencia.

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
| `kitlegal graph check [<norma> [<bloques>...]]` | Comprueba lo consultado de una norma, de algunos de sus bloques o, sin argumentos, todo lo consultado, y lista como mucho 50 hallazgos: redacciones que han cambiado desde la lectura anterior y consultas caducadas. | objeto con `norma`, `bloques`, `version-obsoleta`, `fuente-caducada`, `omitidos`, `hallazgos` |

Todas devuelven el sobre `ok`, `fuente`, `url`, `fecha_consulta`, `hash`, `data`; con `ok` falso, `data` lleva `clase` y `mensaje`.

Banderas comunes: `--json`, `--timeout <valor>`, `--offline`, `--dry-run`, `--describe`, `--no-graph`, `--asunto <valor>`, `--verbose`.

<!-- fin de la tabla de comandos -->

## Reglas

1. **No concluir que algo no existe.** Una búsqueda vacía, o que una norma no esté en `references/normas.md`, no prueba
   que la norma o la regulación no existan: di «no encontrada con esta búsqueda» y propón reformular la búsqueda.
2. **Nunca inventar contenido legal.** Si `kitlegal boe` falla —código 3 (no encontrado), 4 (fuente no disponible) o 5
   (límite de ritmo), o sin caché con `--offline`— o no está disponible, di qué no se pudo consultar y por qué con lo
   que significa para quien pregunta —que el artículo no está en la norma, que la fuente no estaba disponible, que la
   fuente limitó las consultas—, sin el código, y no suplas el texto con conocimiento propio. Si la orden que falló
   pedía varios bloques, dilo solo después de haber pedido cada bloque por separado con `kitlegal boe articulo`
   (paso 3), y di cuáles no se pudieron consultar. Si `kitlegal` no está en el `PATH`, di que falta instalar kitlegal.
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
   con otro código, responde igual con el texto de `kitlegal boe` y di que no se ha podido comprobar si la redacción ha
   cambiado desde una consulta anterior, sin afirmar que ha cambiado ni que no, y sin nombrar la memoria de consultas,
   `kitlegal graph` ni el código:

   ```text
   No se ha podido comprobar si la redacción ha cambiado desde una consulta anterior.
   ```
