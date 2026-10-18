---
name: legal-core
description: >-
  Punto de partida de las preguntas de derecho público español que dependen de un municipio o de un plazo: identifica
  el territorio de un municipio —código INE, provincia, comunidad, régimen, DIR3 del ayuntamiento y boletines— y
  calcula con kitlegal hasta qué día se puede recurrir una resolución, cuándo vence un plazo administrativo
  o el del recurso contencioso-administrativo y cuándo se produce el silencio, con los festivos de los municipios que
  cuentan; y dice cuándo un asunto lo regula la ley aparte (tributos, multas de tráfico, Seguridad Social,
  extranjería). Razona con la jerarquía normativa y las leyes vertebrales (Constitución, LPAC, LRJSP, LRBRL, TRLRHL,
  LCSP, LGT…). Úsala cuando se pregunte hasta cuándo se puede recurrir o qué plazo hay para un recurso, cuándo vence un
  plazo o se produce el silencio, qué comunidad, provincia o boletines corresponden a un ayuntamiento, si un municipio
  es foral o qué norma prevalece. Para el texto de un artículo se apoya en boe-legislacion.
metadata:
  kitlegal-applets: territorio plazos
  kitlegal-referencias: leyes_vertebrales jerarquia_normativa recursos_y_plazos cobertura_por_materia
---

# Territorio, plazos, jerarquía normativa y leyes vertebrales

Esta skill es el punto de partida de las preguntas de derecho público español que dependen de dónde se plantean o de
un plazo. Identifica el territorio con `kitlegal territorio`, calcula los plazos con `kitlegal plazos` y razona con
cuatro referencias: `references/leyes_vertebrales.md`, con el identificador `BOE-A-…` de cada ley vertebral;
`references/jerarquia_normativa.md`, con qué nivel regula qué; `references/recursos_y_plazos.md`, con cada tipo de
plazo, su duración y el precepto que la fija; y `references/cobertura_por_materia.md`, con lo que la ley regula aparte.
No da el texto de ningún artículo: eso lo hace `boe-legislacion`. La skill identifica, calcula y orienta: no tramita
nada y no sustituye el asesoramiento de un profesional.

## Protocolo

Una pregunta de territorio sigue del 1 al 4, y el 5 y el 6 si va más allá del territorio. Una pregunta de plazos o de
recursos empieza por el 7, que decide si hay plazo que calcular antes de preguntar nada; solo si lo hay, el 8 resuelve
los municipios que cuentan con los pasos 1 a 4, y siguen del 9 al 11. Cada operación se pide de una de dos formas,
que devuelven el mismo sobre, con `fuente`, `url`, `fecha_consulta` y `hash` junto a `data`:

- **Con la herramienta** de su fila en «Comandos», si entre las tuyas hay una con ese nombre, solo o detrás del prefijo
  que le ponga tu agente, como `mcp__kitlegal__territorio_resolver`. Si la tienes, úsala siempre.
- **Con la orden** `kitlegal … --json`, si no la tienes.

Donde un paso dice que la orden termina con un código, con la herramienta el resultado viene marcado como error y su
`data.clase` dice cuál («Comandos»).

### 1. Identificar el territorio

- Toma el municipio de la conversación: su nombre o su código INE, tal como los haya dado la persona.
- Si la conversación no lo dice, **pregúntalo** y espera la respuesta. No lo supongas, no lo deduzcas de otros datos
  ni sigas con un municipio de ejemplo, tampoco con el de quien usa el kit. Si solo se nombra una provincia o una
  comunidad, pregunta también por el municipio: `kitlegal territorio` resuelve municipios.

### 2. Resolverlo con `kitlegal territorio`

Llama a `territorio_resolver` con el municipio en `consulta` o ejecuta `kitlegal territorio resolver <nombre o código
INE> --json`.

- Pasa el nombre o el código tal como los dio la persona; no conviertas de memoria un nombre en un código ni al revés.
- Ningún dato de territorio —comunidad, provincia, régimen, DIR3, boletines— se da por sabido ni se escribe de
  memoria: todos salen de `data` en esta conversación. Ningún municipio ni ninguna comunidad tiene un trato propio.
- Con el código 0, `data` trae siempre `municipio`, `codigo_ine`, `provincia`, `comunidad`, `dir3`, `regimen`,
  `boletines` y `cobertura`, y cada dato lleva su `source`.
- Si no tienes ni la herramienta ni el binario, vale la regla 7.

### 3. Leer `cobertura` y trasladarla a la respuesta

| Aspecto | Valores | Qué dice |
|---|---|---|
| `boletin_autonomico` | `configurado`, `no-configurado` | si kitlegal tiene declarado el boletín de la comunidad |
| `boletin_provincial` | `configurado`, `no-configurado` | si kitlegal tiene declarado el de la provincia |
| `dir3` | `verificado`, `no-verificado` | si el DIR3 del ayuntamiento está en la correspondencia verificada |

- Responde con lo que dice `data` y escribe la cobertura con sus tres aspectos en su forma fija (más en «Cómo se
  presenta el territorio»). Lo `no-configurado` o `no-verificado` dilo con sus palabras: **no** significa que no
  exista (regla 5).
- **No nombres ningún boletín que el applet no haya devuelto**, aunque creas conocerlo; la clase de boletín de
  `references/jerarquia_normativa.md` no suple un nivel `no-configurado`.
- Con `dir3` en `no-verificado`, no compongas el DIR3 a partir del código INE. Si `regimen.valor` es `foral`, dilo
  (regla 4). Di de cuándo son los datos: `fecha_consulta`.

### 4. Ambigüedad y ausencia

- **Código 2 (`argumentos`) con candidatos**: `data.mensaje` los da todos, cada uno como `<código INE> <nombre>
  (<provincia>)`. Ofrécelos, pregunta a cuál se refiere la persona y vuelve al paso 2 con el código que elija.
- **Código 2 sin candidatos**: la consulta no forma un código INE válido o su dígito de control no es el oficial. Díselo
  y pide el nombre o el código; no corrijas tú el código.
- **Código 3 (`no-encontrado`)**: no está en la relación de municipios que lleva kitlegal. Dilo así, sin concluir que
  no exista, y pide otra forma del nombre o su código INE.
- Con cualquier otro código o `clase`, di que no se pudo resolver y no suplas sus datos.

### 5. Razonar con las referencias

- Nombra cada norma con el identificador de `references/leyes_vertebrales.md`, con su rango.
- Usa `references/jerarquia_normativa.md` para decir en qué nivel se regula lo preguntado y qué norma prevalece. El
  nivel autonómico es el de la `comunidad` de `data`; los boletines concretos, solo los de `boletines`.

### 6. Delegar el texto en `boe-legislacion`

- **Afirmar lo que dice un artículo exige leerlo con la skill `boe-legislacion` en esta misma conversación**
  (`kitlegal boe articulo <norma> <bloque>` o `boe_articulo`), y citarlo como ella cita. No lo cites de memoria ni desde
  las referencias, que no llevan el texto de ninguna norma.
- El identificador de una norma que no está en las referencias también se resuelve con `boe-legislacion`.

### 7. Decidir si hay plazo que calcular

Con lo que la conversación ya dice del acto o de la solicitud, de la administración que lo dictó o lo recibe y de su
materia, mira estos dos casos **antes de preguntar nada**, tampoco el municipio ni la fecha: ninguno los necesita,
porque no se les calcula nada.

- **Lo que la ley regula aparte.** Busca la materia del asunto en `references/cobertura_por_materia.md`. Si cae en una
  de sus entradas —tributos, Seguridad Social, sancionadores tributarios, sociales, de tráfico o de extranjería,
  extranjería y asilo, una impugnación que sustituye a la alzada, una reclamación económico-administrativa—, declara que
  no está cubierto con la línea de «Cómo se presenta un plazo», con la norma que lo rige si la entrada la trae y, si no,
  con el precepto de la LPAC que lo excluye; no nombres ninguna norma que no venga de la entrada. **No pidas `plazos
  calcular`** y no des ni la línea del plazo ni ninguna fecha: el plazo del común no es el suyo.
  Si la conversación no dice la materia, el asunto no cae en ninguna entrada: **no la preguntes ni detengas la
  respuesta por ella**. Sigue con el 8 y, al responder, di que la fecha no vale si el asunto es de los que la ley regula
  aparte, nombrándolos con la referencia y sin ninguna norma que no venga de ella.
- **Acto presunto**: no hay fecha. Cita los arts. 122.1 y 124.1 LPAC y el 46.1 LJCA, leídos con `boe-legislacion`, y
  no pidas ningún plazo.

### 8. Situar el plazo

- La fecha de inicio: la de la notificación o la publicación; en el silencio, la del acuerdo de iniciación o la de
  entrada de la solicitud en el registro. Si falta, **pregúntala**: `plazos` no la deduce y tú tampoco.
- Los municipios que cuentan: en un plazo administrativo, **el del interesado y el de la sede del órgano** (pueden ser
  el mismo); en el recurso contencioso-administrativo, **el de la sede del juzgado**. Tómalos y resuélvelos con los
  pasos 1 a 4: si la persona no los ha dado, pregúntalos; nunca supongas el de quien usa el kit.

### 9. Pedir el plazo a `plazos`

Elige el tipo en `references/recursos_y_plazos.md` y llama a `plazos_calcular` o ejecuta `kitlegal plazos calcular
<tipo> <inicio> <municipios>... --json`, con los municipios en su orden (interesado y sede del órgano, o sede del
juzgado).

- **Solo municipios que `territorio` haya resuelto en esta conversación** (paso 8), cada uno como lo dio la persona.
  `plazos` acepta un nombre, pero sin el paso 8 la respuesta no tiene el territorio ni la cobertura de ese municipio:
  no se lo pases sin resolver.
- **Silencio**: el plazo máximo para resolver lo fija la norma del procedimiento; léela con `boe-legislacion` y pásalo
  con `--duracion` y `--unidad`. Solo si no fija ninguno, los tres meses del art. 21.3 LPAC, y dilo citándolo.
  `plazos` no dice si el silencio es estimatorio o desestimatorio: eso lo dice la norma, no la fecha.
- **Reclamación del art. 24.2 LTAIBG**: señala como variación autonómica ante quién se reclama, el Consejo de
  Transparencia y Buen Gobierno o el órgano que la comunidad haya designado (art. 24.6).
- Con el código 2 o 3, como en el paso 4: los candidatos de un municipio van en `data.mensaje` tras el municipio
  pedido.

### 10. Leer el precepto

Lee con `boe-legislacion` cada precepto de `data.preceptos` antes de citarlo, y cítalo como ella cita.

### 11. Responder con el plazo

- Da la fecha de `data.fecha` en la línea del plazo, con su tipo y una cita por cada precepto de `data.preceptos`, y
  ninguna otra fecha de vencimiento. **No calcules ni corrijas ninguna fecha**, tampoco para «comprobarla».
- Di qué festivos movieron la fecha, con `data.festivos`: su fecha, su ámbito y, si es local, su municipio.
- Traslada cada aviso de `data.avisos` con su línea `texto`, tal cual.
- Hasta qué disposición llegan los festivos lo dice ya el aviso de cobertura que lo trae; sin aviso, no lo copies.

## Cómo se presenta el territorio

Cada dato sale de `data`, con el nombre y el código tal como los da:

```text
Municipio: <nombre>, código INE <codigo> (dígito de control <digito_de_control>)
Provincia: <nombre de la provincia>
Comunidad o ciudad autónoma: <nombre de la comunidad>, régimen <común o foral, según regimen.valor>
DIR3 del ayuntamiento: <código, o «no verificado»>
Boletines: <código> — <nombre> (<nivel>), uno por cada entrada de `boletines`
boletin_autonomico: <valor>
boletin_provincial: <valor>
dir3: <valor>
```

La cobertura va con la clave y el valor exactos de `data`; detrás puedes explicar qué significa. Un boletín se nombra por
su `codigo` y su `nombre` exactos.

## Cómo se presenta un plazo

La línea del plazo, con la fecha de `data.fecha` tal cual, el tipo como lo nombra `references/recursos_y_plazos.md` y la
cita de cada precepto de `data.preceptos`, leído en el paso 10:

```text
⚠ PLAZO: <AAAA-MM-DD> — <tipo> [<norma>, <precepto>, <identificador>, bloque <bloque>]
```

Cada precepto va en sus propios corchetes, uno detrás de otro en esa misma línea: lo que hace cita es que los corchetes
terminen en `<identificador>, bloque <bloque>]`, con un solo bloque. Con los dos preceptos del recurso
contencioso-administrativo:

```text
⚠ PLAZO: <AAAA-MM-DD> — recurso contencioso-administrativo [art. 46.1 de la LJCA, BOE-A-1998-16718, bloque a46] [art. 128.2 de la LJCA, BOE-A-1998-16718, bloque a128]
```

Dos bloques en unos mismos corchetes —«bloques a46 y a128», o «bloque a46; …, bloque a128»— no son una cita de ninguno
de los dos.

La declaración de que el asunto no está cubierto, con el procedimiento como lo nombra
`references/cobertura_por_materia.md`; la primera si la entrada trae la norma que lo rige, la segunda si no:

```text
⚠ NO CUBIERTO: <procedimiento> se rige por <norma> (<identificador>), no por el procedimiento común: kitlegal no calcula sus plazos ni sus recursos (<precepto de la LPAC>).
⚠ NO CUBIERTO: <procedimiento> lo regula aparte la <precepto de la LPAC> (BOE-A-2015-10565): kitlegal no calcula sus plazos ni sus recursos.
```

Los avisos, cada uno con el `texto` de `data.avisos`, que empieza por una de estas formas:

```text
⚠ FALTAN FESTIVOS NACIONALES Y AUTONÓMICOS: …
⚠ FALTAN FESTIVOS LOCALES: …
⚠ DÍAS DEL ART. 183 LOPJ: …
⚠ FESTIVO DE UNA PARTE: …
```

Detrás de cada línea puedes explicarla con tus palabras; las líneas no se cambian.

## Comandos

`kitlegal` se invoca desde el `PATH`, y la herramienta, por el nombre de su fila. Los dos applets responden con lo que
lleva dentro el binario, sin consultar ninguna fuente. Códigos de salida, y entre paréntesis la `data.clase` del error de
la herramienta: 0 correcto, 2 (`argumentos`) argumentos inválidos —también un nombre que es el de más de un municipio—,
3 (`no-encontrado`) un municipio que no está en la relación.

<!-- inicio de la tabla de comandos: generada desde --describe con make skills-sync, no editar -->

### `kitlegal territorio`

| Orden | Herramienta | Qué hace | Qué devuelve en `data` |
|---|---|---|---|
| `kitlegal territorio resolver <consulta>` | `territorio_resolver` | Devuelve el territorio de un municipio, por su nombre o por su código INE, con la cobertura de lo que está configurado y verificado. | objeto con `municipio`, `codigo_ine`, `provincia`, `comunidad`, `dir3`, `regimen`, `boletines`, `cobertura` |

### `kitlegal plazos`

| Orden | Herramienta | Qué hace | Qué devuelve en `data` |
|---|---|---|---|
| `kitlegal plazos calcular <tipo> <inicio> <municipios>... [--duracion <duracion>] [--unidad <unidad>]` | `plazos_calcular` | Devuelve la fecha de un plazo —un recurso, el acceso a la información pública, el silencio o uno genérico— desde la fecha de inicio que se da, con los preceptos que lo fijan, los festivos saltados, los avisos y la cobertura de los festivos. | objeto con `fecha`, `tipo`, `preceptos`, `municipios`, `festivos`, `avisos`, `cobertura` |

La orden y la herramienta de cada fila devuelven el mismo sobre: `ok`, `fuente`, `url`, `fecha_consulta`, `hash`, `data`; con `ok` falso, `data` lleva `clase` y `mensaje`.

Banderas comunes: `--json`, `--timeout <valor>`, `--offline`, `--dry-run`, `--describe`, `--no-graph`, `--asunto <valor>`, `--verbose`.

<!-- fin de la tabla de comandos -->

## Reglas

1. **Nunca inventar contenido legal ni citar de memoria.** Ni el texto de una norma, ni su identificador, ni ningún
   dato de territorio, ni ninguna fecha: lo que no salga en esta conversación de `kitlegal`, de las referencias o de
   `boe-legislacion`, no se afirma.
2. **Cada afirmación sobre una norma, con su identificador.** El identificador sale de las referencias o de
   `boe-legislacion`; el texto de un artículo, solo de `boe-legislacion`.
3. **Distinguir ley de reglamento.** Di el rango de cada norma que nombres y recuerda que un reglamento nunca puede
   contradecir la ley.
4. **Señalar la variación autonómica.** Di lo que una comunidad puede haber regulado de otro modo, y más aún si su
   régimen es foral; las ordenanzas y demás normas locales no están en las referencias.
5. **No concluir «no existe»** a partir de un resultado sin cobertura completa: ni de un municipio que no está en la
   relación, ni de un boletín `no-configurado`, ni de un DIR3 `no-verificado`, ni de unos festivos que faltan, ni de
   una norma que falta en una referencia.
6. **Ninguna acción con identidad.** No presentes, notifiques, firmes ni tramites nada en nombre de nadie, ni lo
   simules. Si la pregunta lo pide, di que es una acción que hace la persona.
7. **Sin herramienta y sin binario, la respuesta lo dice.** Si no tienes la herramienta y la orden falla porque
   `kitlegal` no está, no has consultado nada: no afirmes ningún dato de territorio, ninguna fecha ni ningún plazo, ni
   nada del contenido de una norma, tampoco de memoria ni con salvedades, y no escribas ninguna cita. La respuesta lleva
   esta línea, con la causa en lugar del marcador y la dirección en la misma línea:

   ```text
   ⚠ SIN CONSULTA AL BOE: <causa>. Para consultarlo hace falta instalar kitlegal: https://kitlegal.es/instalar/
   ```

8. **Ninguna fecha que no dé `plazos`.** La fecha de un plazo sale de `data.fecha`; ni la cuentes ni la retoques.
9. **Lo que la ley regula aparte no recibe el plazo del común.** Si el asunto está en
   `references/cobertura_por_materia.md`, se declara no cubierto y no se pide ningún plazo.
