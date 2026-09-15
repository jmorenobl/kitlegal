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
  kitlegal-applets: boe
  kitlegal-referencias: normas
---

# Consultar y citar legislación consolidada del BOE

Esta skill responde preguntas sobre el contenido de normas consolidadas del Boletín Oficial del Estado —la
Constitución, leyes orgánicas y ordinarias, reales decretos legislativos, reales decretos y las normas autonómicas que
el BOE consolida— de cualquier ámbito: procedimiento administrativo, contratación pública, régimen local, tributos,
transparencia, relaciones laborales… Lo que dice de una norma sale del texto que devuelve `scripts/boe` en la misma
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
- Si no está, búscala con `scripts/boe buscar` por las palabras de su título:

  ```bash
  scripts/boe buscar régimen jurídico del sector público --json
  ```

  Elige entre los resultados por título y rango, y di en la respuesta qué norma elegiste. Si hay varias posibles —una
  ley y su texto refundido, una ley y el reglamento que la desarrolla, una norma estatal y otra autonómica de título
  parecido—, di cuáles y por qué eliges una; si la pregunta no permite elegir, pregunta o responde de ambas
  distinguiéndolas.
- Una norma autonómica consolidada en el BOE se resuelve igual: si `scripts/boe buscar` la encuentra por su título, se
  lee y se cita como una estatal.
- Si la búsqueda no da la norma, reformúlala con otras palabras del título; que no aparezca no prueba que no exista
  (regla 1).

### 3. Leer índice y bloques con `scripts/boe`

- Si no conoces el id del bloque —saber el número del artículo no basta—, lee el índice de la norma:

  ```bash
  scripts/boe indice BOE-A-2015-10565 --json
  ```

- Copia el id de la entrada del índice cuyo `titulo` es el artículo que buscas; nunca lo compongas a partir del número
  del artículo, porque en muchas normas los ids no son `a<número>`. En la Ley 9/2017, la entrada con `titulo`
  «Artículo 118» tiene el id `a1-30`, y `a118` no está en su índice.
- Lee los bloques de uno en uno con `scripts/boe articulo`:

  ```bash
  scripts/boe articulo BOE-A-2015-10565 a21 --json
  ```

  Usa `scripts/boe articulos`, que los devuelve en el orden pedido, solo cuando necesites varios bloques a la vez y
  todos salgan del índice.
- Una orden de `scripts/boe articulos` falla entera en cuanto falla uno de sus bloques. Si una orden con varios bloques
  termina con el código 4 o 5, pide cada bloque por separado con `scripts/boe articulo` antes de dar ninguno por no
  consultado: el fallo de un bloque no impide leer los demás.

- Sigue las remisiones que hagan falta para responder: si el bloque remite a otro artículo, de la misma norma o de
  otra, lee también el bloque remitido, resolviendo antes la otra norma con los pasos 1 y 2.
- Si la pregunta depende de la vigencia de la norma o de sus modificaciones, lee sus metadatos y su análisis:

  ```bash
  scripts/boe metadatos BOE-A-2017-12902 --json
  scripts/boe analisis BOE-A-2017-12902 --json
  ```

- No pidas nunca un id de bloque que no salga del índice o de la propia pregunta. Si `scripts/boe` termina con el
  código 3 (no encontrado), vuelve al índice en lugar de probar otros ids; si el artículo no existe en la norma, dilo.

### 4. Evaluar si falta contexto

Antes de responder, comprueba si lo leído basta:

- **Remisiones**: si el bloque remite a otro artículo, a otra ley o a un reglamento que cambia la respuesta, léelo
  (paso 3).
- **Vigencia**: si el sobre trae avisos (derogada, vigencia agotada, consolidación no finalizada) o la fecha de
  vigencia del bloque no encaja con la situación preguntada, tenlo en cuenta y trasládalo (regla 3).
- **Modificaciones**: si una norma posterior cambió el bloque (`norma_modificadora`) de un modo que importa para la
  pregunta, consulta `scripts/boe metadatos` o `scripts/boe analisis`.

Si falta algo que no puedes leer con `scripts/boe`, dilo en la respuesta en lugar de suplirlo.

### 5. Responder citando

- Cada afirmación sobre el contenido de una norma lleva su cita, y lo citado sale del texto que devolvió `scripts/boe`
  en esta conversación. La cita es la forma legible de la norma y del bloque seguida, en la misma línea, de
  `[<identificador>, bloque <id>]`, con el corchete de apertura seguido inmediatamente del identificador `BOE-A-…`: el
  nombre de la norma va delante del corchete, nunca dentro (más en «Cómo se cita»).
- **Distingue ley y reglamento**: cuando cites normas de rango distinto, di el rango de cada una —el `rango` de
  `references/normas.md` o de la búsqueda— y recuerda que la ley prevalece sobre el reglamento que la desarrolla.
- **Señala la variación autonómica**: cuando lo preguntado pueda variar por normativa autonómica (competencias
  compartidas o cedidas, desarrollo autonómico, régimen foral), dilo; y cuando corresponda a ordenanzas u otras normas
  locales, di que no están en esta fuente.
- Traslada los avisos de vigencia del sobre y recuerda que los textos consolidados del BOE tienen carácter informativo.
- Antes de responder, repasa cada cita: su corchete de apertura va seguido de `BOE-`. Si dentro de los corchetes hay
  algo delante del identificador —el nombre de la norma, «art.», «artículo»—, sácalo delante del corchete, en la misma
  línea.

## Cómo se cita

Cada cita lleva primero la forma legible de la norma y del bloque y, detrás, entre corchetes, el identificador de la
norma y el id del bloque tal como los da la fuente:

```text
art. 21 de la Ley 39/2015 [BOE-A-2015-10565, bloque a21]
```

- La parte entre corchetes es exactamente `[<identificador>, bloque <id de bloque>]`, con el corchete de apertura
  seguido inmediatamente del identificador. No se admite otra forma para esa parte: dentro de los corchetes no va nada
  más que el identificador y el id. Ni «art. 21», ni «artículo 21», ni el nombre, el número o el rango de la norma, que
  van delante, fuera de los corchetes; ni el identificador sin el id del bloque. No valen
  `[Ley 39/2015, BOE-A-2015-10565, bloque a21]` ni `[Constitución Española, BOE-A-1978-31229, bloque a140]`: se
  escriben `art. 21 de la Ley 39/2015 [BOE-A-2015-10565, bloque a21]` y
  `art. 140 de la Constitución Española [BOE-A-1978-31229, bloque a140]`.
- La regla vale igual cuando la cita va sola en una línea o debajo de una cita textual en bloque, como tras transcribir
  el artículo: también entonces la forma legible va delante, en la misma línea,
  `art. 140 de la Constitución Española [BOE-A-1978-31229, bloque a140]`, y el nombre de la norma no pasa dentro de
  los corchetes.
- El identificador y el id van tal como los devuelve `scripts/boe`, también cuando el id termina en punto: el corchete
  de cierre lo delimita.
- Una cita por bloque. Un bloque remitido se cita por separado, con su norma y su id.

## Comandos

Invoca el binario por el enlace `scripts/boe` de esta skill. Códigos de salida: 0 correcto, 2 argumentos inválidos, 3
no encontrado, 4 fuente no disponible, 5 límite de ritmo de la fuente, 6 requiere identidad humana.

<!-- inicio de la tabla de comandos: generada desde --describe con make skills-sync, no editar -->

### `scripts/boe`

| Orden | Qué hace | Qué devuelve en `data` |
|---|---|---|
| `scripts/boe buscar <texto>...` | Busca normas consolidadas por las palabras de su título o con una consulta de la fuente. | lista de objetos con `identificador`, `titulo`, `rango`, `vigencia_agotada`, `estado_consolidacion`, `url` |
| `scripts/boe indice <norma>` | Devuelve los bloques de una norma consolidada, en el orden de la fuente. | objeto con `norma`, `url`, `bloques` |
| `scripts/boe articulo <norma> <bloque>` | Devuelve el texto vigente de un bloque de una norma, con los avisos de su vigencia. | objeto con `norma`, `bloque`, `titulo`, `tipo`, `fecha_version`, `fecha_vigencia`, `norma_modificadora`, `texto`, `hash_texto`, `avisos`, `url`, `url_eli` |
| `scripts/boe articulos <norma> <bloques>...` | Devuelve el texto vigente de varios bloques de una norma, en el orden pedido. | lista de objetos con `norma`, `bloque`, `titulo`, `tipo`, `fecha_version`, `fecha_vigencia`, `norma_modificadora`, `texto`, `hash_texto`, `avisos`, `url`, `url_eli` |
| `scripts/boe metadatos <norma>` | Devuelve los datos de una norma y los avisos de su vigencia. | objeto con `norma`, `titulo`, `rango`, `numero_oficial`, `fecha_disposicion`, `fecha_publicacion`, `fecha_vigencia`, `estatus_derogacion`, `vigencia_agotada`, `estado_consolidacion`, `url_eli`, `avisos` |
| `scripts/boe analisis <norma>` | Devuelve las materias, las notas y las referencias de una norma. | objeto con `norma`, `materias`, `notas`, `referencias` |

Todas devuelven el sobre `ok`, `fuente`, `url`, `fecha_consulta`, `hash`, `data`; con `ok` falso, `data` lleva `clase` y `mensaje`.

Banderas comunes: `--json`, `--timeout <valor>`, `--offline`, `--dry-run`, `--describe`, `--no-graph`, `--asunto <valor>`, `--verbose`.

<!-- fin de la tabla de comandos -->

## Reglas

1. **No concluir que algo no existe.** Una búsqueda vacía, o que una norma no esté en `references/normas.md`, no prueba
   que la norma o la regulación no existan: di «no encontrada con esta búsqueda» y propón reformular la búsqueda.
2. **Nunca inventar contenido legal.** Si `scripts/boe` falla —código 3 (no encontrado), 4 (fuente no disponible) o 5
   (límite de ritmo), o sin caché con `--offline`— o no está disponible, di qué no se pudo consultar y no suplas el
   texto con conocimiento propio. Si la orden que falló pedía varios bloques, dilo solo después de haber pedido cada
   bloque por separado con `scripts/boe articulo` (paso 3), y di cuáles no se pudieron consultar. Si `scripts/boe` no
   resuelve a un binario, di que falta instalar kitlegal.
3. **Trasladar la vigencia.** Traslada los avisos de vigencia que devuelve el binario (derogada, vigencia agotada,
   consolidación no finalizada) y no presentes como vigente el texto de una norma derogada. Recuerda que los textos
   consolidados del BOE tienen carácter informativo y no son asesoramiento.
4. **Nunca actuar en nombre de nadie.** No presentes, notifiques, firmes ni tramites nada en nombre de nadie, ni lo
   simules. Si la pregunta lo pide, di que es una acción que hace la persona y cita, si procede, la norma aplicable.
5. **Ningún caso especial para un territorio.** No hay reglas propias de un municipio o de una comunidad concretos. Si
   se pregunta por un municipio, responde con la normativa estatal o autonómica consolidada en el BOE y señala que las
   ordenanzas y demás normas locales no están en esta fuente.
