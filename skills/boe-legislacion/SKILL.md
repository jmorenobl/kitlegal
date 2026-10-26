---
name: boe-legislacion
description: >-
  Consulta y cita normativa consolidada del Boletín Oficial del Estado (BOE) de cualquier materia —procedimiento
  administrativo, contratación pública, régimen local, tributos, transparencia, relaciones laborales…—: lee con
  kitlegal índice y artículos y cita identificador y bloque. Se usa siempre que la respuesta dependa de lo que dice una
  norma —qué dice un artículo, ley o real decreto, qué plazo, requisito o procedimiento fija, dónde se regula una
  materia— o de lo que decía en una fecha, la del acto o de los hechos, aunque el agente crea saberlo: sin leer la
  norma con kitlegal no hay cita. Se activa antes de boe_articulo u otra herramienta boe_…: qué pedir y cómo citar lo
  dice la skill. La norma, estatal o autonómica, se nombra por número y año (Ley 39/2015, Real Decreto Legislativo
  2/2004), abreviatura (LPAC, LCSP, LRBRL, LGT, TRLRHL, LRJSP) o identificador BOE-A-….
metadata:
  kitlegal-applets: boe graph
  kitlegal-referencias: normas
---

# Consultar y citar legislación consolidada del BOE

## Protocolo

Sigue los pasos en orden, cada operación con su orden, `kitlegal <applet> <verbo> … --json`, o, si la tienes, siempre
con su herramienta, tras leer [con_herramientas](references/con_herramientas.md). Sin su `fuente`, `url`,
`fecha_consulta` y `hash`, no hay cita.

### 1. Identificar la norma

Toma el `BOE-A-…` de la tabla de [normas](references/normas.md), no exhaustiva: si no está, búscala por su título
con [buscar_la_norma](references/buscar_la_norma.md).

### 2. Leer

- Sin el id del bloque, lee el índice: `kitlegal boe indice BOE-A-2015-10565 --json`. Copia el id de la entrada cuyo
  `titulo` es el artículo, nunca lo compongas con su número, y no pidas ninguno que no salga del índice o de la
  pregunta.
- Sin fecha —con ella, [redaccion_en_una_fecha](references/redaccion_en_una_fecha.md)—, lee cada bloque una vez, con
  la comprobación de esa norma y ese bloque detrás, en la misma orden:
  ```bash
  kitlegal boe articulo BOE-A-2015-10565 a21 --json && kitlegal graph check BOE-A-2015-10565 a21 --json
  ```
  En PowerShell 5.1:
  ```powershell
  kitlegal boe articulo BOE-A-2015-10565 a21 --json; if ($LASTEXITCODE -eq 0) { kitlegal graph check BOE-A-2015-10565 a21 --json }
  ```
  La comprobación, nunca sin argumentos ni antes; si faltó, pídela ya.
- Varios bloques a la vez: [varios_bloques](references/varios_bloques.md). Si una orden no termina con `0`, antes de
  seguir: [lecturas_que_fallan](references/lecturas_que_fallan.md).

### 3. Completar

Lee lo remitido que cambia la respuesta, de esta norma o de otra, resuelta antes con los pasos 1 y 2, y, si importan
la vigencia o una modificación, los metadatos y el análisis. Nombra un precepto solo si `kitlegal boe articulo` o
`articulos` ha devuelto su texto en esta conversación, salvo lo del paso 5: si lo remite un bloque leído en el paso 2,
o su id solo salió en un índice, léelo antes con los pasos 1 y 2, o no lo nombres; leído, di de qué trata, con su cita.

### 4. Responder

- **La respuesta empieza por lo que se pregunta y habla de la norma, no de la consulta.** Es todo lo que escribes
  tras la última orden o llamada, desde su primera palabra, y quien pregunta la lee entera: no ve las órdenes, las
  llamadas ni lo que devuelven, y lo que decides al completar la lectura no va en ella. La hacen la norma, su texto, su
  cita, sus avisos y las líneas `⚠` de cada bloque: escribe como quien tiene el BOE delante,
  «el artículo no lleva avisos de vigencia» y no «la lectura no trajo avisos», la norma por su nombre e identificador,
  sin decir de dónde sale, qué has hecho o vas a hacer ni qué ha devuelto o no una orden: no le sirve.
- **Nada de otra conversación**, ni para afirmar, confirmar o desmentir: de ellas solo va la línea
  `⚠ REDACCIÓN MODIFICADA:` de cada bloque que la trae.
- Cada afirmación sobre el contenido de una norma lleva su cita, de lo que `kitlegal boe` devolvió en esta
  conversación.
- **Ley y reglamento**: con rangos distintos, di el `rango` de cada una, de la tabla o de la búsqueda, y que la ley
  prevalece sobre su reglamento.
- De la redacción leída, di qué norma la dio y desde cuándo rige, nunca hasta cuándo: ninguna lectura lo trae, ni el
  aviso de una derogada; `kitlegal graph check` no dice nada de la vigencia.

### 5. De un precepto no leído, nada

De un artículo, apartado o disposición que nada te ha devuelto en esta conversación no digas qué dice, de qué trata ni
cuándo o cómo se aplica, ni en un inciso ni aunque creas saberlo: sería texto legal sin fuente. Número y materia no
van juntos en ninguna frase, ni afirmando («el art. N, que es…», «el art. N sobre <materia>»,
«los artículos que regulan <materia> (N y M)»), ni negando («no puedo decir qué <regla> fija»), ni al ofrecer leerlo o
avisar de lo no leído: o el número, o la materia. Lo leído para nombrarlo (paso 3) es un salto: lo que remite no se
lee para nombrarlo y, como lo que no se pudo leer, se nombra solo con las palabras del texto leído. Puedes avisar de
que una materia se regula en otra parte, sin nombrar precepto ni regla.

## Cómo se cita

La forma legible de la norma y del bloque y, en la misma línea, entre corchetes, el identificador y el id como los da
`kitlegal boe`, nada tras el id, aunque acabe en punto: `art. 21 de la Ley 39/2015 [BOE-A-2015-10565, bloque a21]`.
Una cita por bloque, también del remitido; repásalas, y para otra forma, [como_se_cita](references/como_se_cita.md).
Cada línea `⚠` va entera, en una sola línea y con los datos de la lectura: dicha de otro modo, falta.

## Redacción en una fecha

Con una fecha en la pregunta —un día, un mes, un año o un hecho fechado—:

```text
⚠ REDACCIÓN A FECHA: <forma legible> [<identificador>, bloque <id>]: se cita la redacción vigente el AAAA-MM-DD, la de fecha de vigencia AAAAMMDD.
⚠ REDACCIÓN DE HOY: <forma legible> [<identificador>, bloque <id>]: hoy la redacción es otra, la de <identificador>, con fecha de vigencia AAAAMMDD.
⚠ SIN REDACCIÓN VIGENTE: <forma legible> [<identificador>, bloque <id>]: el AAAA-MM-DD no tenía redacción vigente; la tiene desde el AAAAMMDD, la fecha de vigencia de su primera redacción.
```

## Redacción modificada

Por cada entrada `version-obsoleta` de `kitlegal graph check`, y si se pregunta qué cambió, lee
[redaccion_modificada](references/redaccion_modificada.md):

```text
⚠ REDACCIÓN MODIFICADA: <forma legible> [<identificador>, bloque <id>]: la redacción con fecha de vigencia AAAAMMDD, la que se consultó antes, ha sido sustituida por la de AAAAMMDD, que es la que se cita.
```

De `kitlegal graph` solo se usa `check`, y nada más suyo va en la respuesta.

## Comandos

`kitlegal` se invoca desde el `PATH`.

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

1. **No concluir que algo no existe.** Una búsqueda vacía, o una norma fuera de la tabla, no prueba que no exista.
2. **Nunca inventar contenido legal.** Si `kitlegal boe` falla o no está, di qué bloques no pudiste consultar y por
   qué, y no suplas nada de memoria; sin herramienta ni binario, la regla 7.
3. **Trasladar la vigencia.** Cada aviso va con su forma fija, `⚠`, su etiqueta y dos puntos, y su frase o una
   explicación: `⚠ NORMA DEROGADA:`, `⚠ VIGENCIA AGOTADA:` o `⚠ TEXTO POSIBLEMENTE DESACTUALIZADO:`. Trasládalos,
   también una fecha de vigencia del bloque que no encaja con lo preguntado; una derogada no es vigente. Los textos
   consolidados del BOE son informativos y no sustituyen el asesoramiento de un profesional.
4. **Nunca actuar en nombre de nadie.** La skill consulta y cita: no presentes, notifiques, firmes ni tramites nada por
   nadie, ni lo simules; si se pide, di que lo hace la persona y cita, si procede, la norma aplicable.
5. **Ningún caso especial para un territorio.** Nada propio de un municipio o una comunidad concretos. Si se pregunta
   por un municipio, responde con la normativa estatal o autonómica consolidada en el BOE y di que las ordenanzas y
   demás normas locales no están en esta fuente; dilo también siempre que lo preguntado les corresponda. Si puede
   variar por normativa autonómica (competencias compartidas o cedidas, desarrollo autonómico, régimen foral), dilo.
6. **El texto sale de `kitlegal boe`, nunca de `kitlegal graph`.** Lo citado sale de `kitlegal boe articulo` o
   `articulos`; lo de `kitlegal graph` no se cita, parafrasea ni reconstruye: dice qué volver a comprobar, no qué dice
   el artículo.
7. **Sin herramienta y sin binario, la respuesta lo dice.** Sin la herramienta, si `kitlegal` no está —el shell no lo
   encuentra o no ejecutas órdenes—, no lo busques en otras carpetas ni en un plugin: no has consultado nada.
   No afirmes nada de la norma, de memoria o con salvedades, ni cites; lleva esta línea, con la causa en el marcador:
   ```text
   ⚠ SIN CONSULTA AL BOE: <causa>. Si ya tienes kitlegal, cierra Claude del todo, ábrelo y pregunta en un chat nuevo; si no: https://kitlegal.es/instalar/
   ```
