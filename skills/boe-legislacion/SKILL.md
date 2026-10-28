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

## Protocolo

Sigue los pasos en orden, cada operación con `kitlegal … --json` desde el `PATH` o, si la tienes, con su
herramienta, tras leer [con_herramientas](references/con_herramientas.md).

### 1. Identificar la norma

Toma el `BOE-A-…` de la tabla de [normas](references/normas.md); si no está, búscala con
[buscar_la_norma](references/buscar_la_norma.md).

### 2. Leer

- Sin el id del bloque, lee el índice: `kitlegal boe indice BOE-A-2015-10565 --json`. Copia el id de la entrada cuyo
  `titulo` es el artículo; nunca lo compongas con su número ni pidas uno que no salga del índice o de la pregunta.
- Sin fecha —con ella, [redaccion_en_una_fecha](references/redaccion_en_una_fecha.md)—, lee cada bloque una vez, con
  su comprobación detrás, en la misma orden:
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

Si un bloque leído remite a otro precepto («sin perjuicio de lo dispuesto en el artículo N», «en los términos
del artículo N»), de esta norma o de otra, o viste su id solo en un índice, y la respuesta va a hablar de él,
aunque sea para avisar, léelo con los pasos 1 y 2 antes de escribir. Es un salto: lo que remita esa segunda
lectura no se lee. Si importan la vigencia o una modificación, lee también los metadatos y el análisis.

Por qué: quien pregunta va a actuar con lo que escribes y no puede comprobar lo que no lleva cita; con el texto
delante, lo que sabes de ese precepto se puede decir y citar.

Ojo: una remisión no dice de qué trata lo remitido. Lo que recuerdes de él, aunque sea cierto, no está en lo
leído: si te parece útil decirlo, esa es la señal de que hay que leerlo.

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
- **Ley y reglamento**: di el `rango` de cada una y que la ley prevalece sobre su reglamento.
- De la redacción leída, di qué norma la dio y desde cuándo rige, nunca hasta cuándo: ninguna lectura lo trae, ni el
  aviso de una derogada.

### 5. De un precepto no leído, nada

De un artículo, apartado o disposición que nada te ha devuelto en esta conversación no digas qué dice, de qué trata ni
cuándo o cómo se aplica, ni en un inciso ni aunque creas saberlo: sería texto legal sin fuente. Número y materia no
van juntos en ninguna frase, ni afirmando («el art. N, que es…», «el art. N sobre <materia>»,
«los artículos que regulan <materia> (N y M)»), ni negando («no puedo decir qué <regla> fija»), ni al ofrecer leerlo o
avisar de lo no leído: o el número, o la materia. Lo leído para nombrarlo (paso 3) es un salto: lo que remite no se
lee para nombrarlo y, como lo que no se pudo leer, se nombra solo con las palabras del texto leído. Puedes avisar de
que una materia se regula en otra parte, sin nombrar precepto ni regla.

### 6. Repasar antes de enviar

Con la respuesta escrita, y antes de enviarla, repásala frase a frase. Es el último paso: no lo saltes.

1. Por cada cosa que la respuesta dice de un precepto, con su número o sin él, también de aquello a lo que
   remite el texto leído, busca en esta conversación la orden `kitlegal boe articulo` o `articulos`, o la
   llamada `boe_articulo` o `boe_articulos`, que devolvió su texto. Si no está, léelo ahora con los pasos 1
   y 2, un salto, y deja lo dicho con su cita; si no se puede leer, quita la frase o déjala con las palabras
   del texto leído, nada más.
2. Quita toda frase que hable de la consulta y no de la norma: lo que hiciste, una orden, lo que devolvió o
   no devolvió.

Envía la respuesta corregida, no la primera.

## Cómo se cita

La forma legible y, en la misma línea, entre corchetes, el identificador y el id como los da `kitlegal boe`, nada
tras el id aunque acabe en punto: `art. 21 de la Ley 39/2015 [BOE-A-2015-10565, bloque a21]`. Una cita por bloque,
también del remitido; repásalas, y para otra forma, [como_se_cita](references/como_se_cita.md). Cada línea `⚠` va
entera, en una línea y con los datos de la lectura: dicha de otro modo, falta.

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

## Comandos

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
2. **Nunca inventar contenido legal.** Si una lectura falla, di qué bloques faltan y por qué, y no suplas
   nada de memoria.
3. **Trasladar la vigencia.** Cada aviso, con su frase o una explicación: `⚠ NORMA DEROGADA:`, `⚠ VIGENCIA AGOTADA:`
   o `⚠ TEXTO POSIBLEMENTE DESACTUALIZADO:`, y la fecha de vigencia del bloque que no encaja con lo preguntado; una
   derogada no es vigente. El texto consolidado es informativo y no sustituye el asesoramiento de un profesional.
4. **Nunca actuar en nombre de nadie.** No presentes, notifiques, firmes ni tramites nada, ni lo simules: lo hace la
   persona.
5. **Ningún caso especial para un territorio.** Las ordenanzas y demás normas locales no están en esta fuente: dilo si
   se pregunta por un municipio o lo preguntado les corresponde. Si puede variar por normativa autonómica o foral, dilo.
6. **El texto sale de `kitlegal boe`, nunca de `kitlegal graph`.** Del grafo, solo `check`.
7. **Sin herramienta y sin binario, la respuesta lo dice.** Sin la herramienta, si `kitlegal` no está —el shell no lo
   encuentra o no ejecutas órdenes—, no lo busques en otras carpetas ni en un plugin: no has consultado nada.
   No afirmes nada de la norma, de memoria o con salvedades, ni cites; lleva esta línea, con la causa en el marcador:
   ```text
   ⚠ SIN CONSULTA AL BOE: <causa>. Si ya tienes kitlegal, cierra Claude del todo, ábrelo y pregunta en un chat nuevo; si no: https://kitlegal.es/instalar/
   ```
