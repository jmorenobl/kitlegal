---
name: jurisprudencia
description: >-
  Sentencias y jurisprudencia de los tribunales españoles, sin citar ninguna de memoria. Úsala siempre que la pregunta
  nombre, pida, cite o resuma una sentencia o un auto —«STS 1088/2023, de 4 de julio», un ECLI, un ROJ, «la sentencia
  del Supremo sobre…»—, pregunte si una sentencia existe o qué dice, pida jurisprudencia sobre una materia («qué dice
  la jurisprudencia sobre la cláusula suelo») o traiga el texto o el PDF de una resolución, también cuando creas
  conocerla: una sentencia solo se cita con su documento delante. kitlegal no consulta el buscador del CENDOJ: prepara
  la consulta exacta para que la persona la busque con su navegador y coteja el documento que trae. Actívala antes de
  llamar a cita_preparar o a cita_cotejar: qué pedir, qué decir de la sentencia que no está delante y cómo citar la
  que sí lo dice la skill.
metadata:
  kitlegal-applets: cita
---

# Sentencias: la consulta exacta, el documento cotejado y ninguna cita de memoria

Esta skill responde cuando hace falta una sentencia o un auto de un tribunal español. kitlegal **no consulta el
buscador de jurisprudencia del CENDOJ**: el buscador pide un CAPTCHA a los programas, y no se sortea. La búsqueda la
hace la persona con su navegador, y la skill le quita el trabajo de alrededor: le da la consulta exacta —la dirección
y cada casilla con su valor— y, cuando trae el documento, comprueba que es el que se pidió. Una sentencia citada de
memoria puede no existir, y un escrito que la lleve se paga caro: por eso **solo se cita la sentencia cuyo documento
está en la conversación**. La skill prepara, coteja y cita: no tramita nada y no sustituye el asesoramiento de un
profesional.

## Protocolo

Sigue los pasos que pida la pregunta. Las dos operaciones se piden de una de dos formas, que devuelven lo mismo:

- **Con su herramienta**, si entre las tuyas hay una con su nombre —`cita_preparar`, `cita_cotejar`—, solo o detrás
  del prefijo que le ponga tu agente, como `mcp__kitlegal__cita_preparar`. Sus argumentos van por su nombre (`ecli`,
  `roj`, `resolucion`, `fecha`, `texto`, `documento`), y solo los que uses: uno escrito vacío es un error de
  argumentos. Si la tienes, úsala siempre.
- **Con su orden**, `kitlegal cita <verbo> … --json`, si no la tienes. Siempre con `--json`.

Donde un paso dice que una orden termina con un código, con la herramienta el resultado viene marcado como error y su
`data.clase` dice cuál («Comandos»). Ninguna de las dos pide nada a la red: no dicen si una sentencia existe.

### 1. Qué sentencias hacen falta, y cuáles están delante

- Una sentencia **está delante** si su documento está en la conversación —su texto pegado, o un PDF adjunto que has
  leído— y lleva su **ficha**, el encabezamiento con el que el CENDOJ abre cada documento: una línea
  `Roj: <ROJ> - <ECLI>` y, debajo, `Órgano:`, `Fecha:`, `Nº de Recurso:`, `Nº de Resolución:`, `Ponente:` y
  `Tipo de Resolución:`.
- **No está delante** la que la persona nombra sin traerla, ni la que tú recuerdas o crees conocer, ni la que otro
  documento menciona de pasada: las que cita el texto de una sentencia son texto de ese documento.
- La que no está delante va por el paso 2; la que sí, por el paso 3. Las del Tribunal Constitucional, por el paso 7.

### 2. La sentencia que no está delante: la línea y la consulta preparada

Por cada sentencia que la respuesta necesita y no está delante, la respuesta lleva una línea con esta forma fija, al
principio de una línea —`⚠`, la etiqueta `SENTENCIA NO COMPROBADA`, dos puntos y la referencia tal como se dio—:

```text
⚠ SENTENCIA NO COMPROBADA: STS 1088/2023, de 4 de julio
```

Y debajo, la consulta para encontrarla, que sale de `cita_preparar` o de su orden. Cómo se pide depende de cómo se
dio la referencia:

| Cómo se dio | Qué se pide | Orden |
|---|---|---|
| Un ECLI (`ECLI:ES:TS:2023:3144`) | `ecli` | `kitlegal cita preparar ECLI:ES:TS:2023:3144 --json` |
| Un ROJ dicho como tal («ROJ: STS 3144/2023») | `roj` | `kitlegal cita preparar --roj "STS 3144/2023" --json` |
| Número y fecha («STS 1088/2023, de 4 de julio») | `resolucion` y `fecha` | `kitlegal cita preparar --resolucion 1088/2023 --fecha 2023-07-04 --json` |
| Número sin fecha («la STS 1088/2023») | nada: se pide la fecha | — |

- Una cita escrita como «STS 1088/2023, de 4 de julio» lleva el **número de resolución**, no el ROJ: se prepara con
  `resolucion` y con `fecha` en la forma `AAAA-MM-DD`. El año de la fecha es el del número, salvo que la cita diga otro.
- **Sin fecha no se prepara de ninguna forma**, tampoco como ROJ: los ROJ son correlativos, y ese mismo número es el
  ROJ de otra sentencia. La respuesta lleva la línea y pide la fecha.
- La respuesta da la consulta con lo que devuelve la operación, sin escribir nada de memoria: la **dirección** del
  buscador (`direccion`) y **cada casilla con su valor** (`casillas`: el nombre que la persona ve en la pantalla, la
  parte de la casilla si tiene dos —«Desde» y «Hasta»— y lo que hay que escribir). Dile a la persona qué hacer: abrir
  la dirección, escribir cada valor en su casilla, buscar, descargar el documento y traerlo —pegado o adjunto—.
- Si la operación da además un **equivalente** —el ROJ de un ECLI o el ECLI de un ROJ—, puedes nombrarlo junto a la
  referencia. No lo compongas tú: solo se deduce para el Tribunal Supremo, y cuando no viene no lo hay.
- **La respuesta no dice que la sentencia existe ni que no existe**: nadie lo ha comprobado. Tampoco dice de qué
  trata, qué resolvió ni si es firme (paso 5).

### 3. Con el documento delante: cotejar su ficha

Pasa a `cita_cotejar`, en `documento`, **la ficha y nada más**: sus líneas tal como están en el documento, desde la
de `Roj:` hasta la de `Tipo de Resolución:`, con las que haya entre ellas. La operación solo lee la ficha: el resto
del texto no cambia lo que devuelve, y copiarlo alarga cada cotejo con el tamaño de la sentencia. Si se había pedido
una sentencia concreta, da además la referencia en la forma en que se pidió (`ecli`, `roj`, o `resolucion` con
`fecha`). Con la orden, la ficha va por la entrada estándar, o en `--documento` si tu shell no tiene `<<`:

```bash
kitlegal cita cotejar --roj "STS 1088/2023" --json <<'DOCUMENTO'
Roj: STS 3144/2023 - ECLI:ES:TS:2023:3144
…las demás líneas de la ficha, hasta la de «Tipo de Resolución:»…
DOCUMENTO
```

- Un PDF lo lees tú y pasas su ficha: kitlegal no abre ficheros.
- Devuelve los ocho datos de la ficha (`ficha`), si el documento es el pedido (`es_la_pedida`, solo si diste una
  referencia) y, si no lo es, un hallazgo que dice qué difiere (`hallazgos`).
- **Si el documento no es el pedido**, dilo con lo que difiere, con las palabras del hallazgo —por ejemplo, que
  1088/2023 es el número de resolución de ese documento y no su ROJ—, y **no lo cites como si fuera el pedido**. La
  sentencia pedida sigue sin estar delante: va por el paso 2, con su línea y su consulta, preparada en la forma en que
  se pidió.
- Mira también `correspondencia`. Con `no-se-corresponden`, el ROJ y el ECLI de la ficha no son de la misma
  resolución: **no cites el documento**, di que no se corresponden, da los dos tal como están en la ficha y pide a la
  persona que vuelva a descargar el documento del buscador y lo traiga. Con cualquier otro valor, cita y no digas nada
  de esto.
- Si termina con el código 2 (`argumentos`) porque el texto no lleva una ficha reconocible, mira el dato que nombra
  el mensaje. Si está en el documento, no pasaste la ficha entera: repite el cotejo con todas sus líneas. Si no
  está, no hay sentencia que citar: di qué falta —el encabezamiento del documento del CENDOJ, o ese dato— y pide el
  documento completo. Sobre un texto sin ficha puedes responder, sin atribuirlo a una sentencia identificada.

### 4. Citar

La cita tiene una forma fija: las siglas, el número de resolución y la fecha, y entre corchetes, en la misma línea,
el ECLI, una coma, `ROJ:` y el ROJ.

```text
STS 1088/2023, de 4 de julio [ECLI:ES:TS:2023:3144, ROJ: STS 3144/2023]
```

- **Todos sus datos son los de la ficha** que leyó `cita_cotejar`, también cuando la persona dio otros: las siglas
  son las del ROJ; el número, el de resolución; la fecha, la de la ficha.
- La respuesta dice que la cita **sale del documento aportado**: kitlegal no ha consultado el buscador.
- **No escribas ningún ECLI que no venga de una de estas operaciones o de la persona**, ni dentro de una cita ni
  suelto. Un ECLI que recuerdas no es un dato.
- Las sentencias que el documento menciona no se citan con esta forma ni se les pone ECLI: se nombran como las
  nombra el texto.

### 5. Qué se dice de una sentencia

- **Con la ficha sola**, la cita y los datos de la ficha: órgano, fecha, número de recurso, ponente y tipo de
  resolución. Nada de su contenido.
- **Con su texto**, léelo y responde sobre él, con sus palabras. Lo que el texto pegado no trae —lo que va tras un
  `[…]`, los fundamentos que faltan— no lo has leído: dilo así.
- **No resumas ni caracterices una sentencia cuyo texto no está en la conversación**, aunque creas conocerla: ni qué
  resolvió, ni su doctrina, ni si la confirma o la corrige otra. Sería jurisprudencia sin fuente.

### 6. Una pregunta por materia

Ante «qué dice la jurisprudencia sobre…», pide `cita_preparar` con los términos de la pregunta en `texto`, o su orden:

```bash
kitlegal cita preparar --texto "cláusula suelo" --json
```

Da la dirección que devuelve, que abre el buscador con esa búsqueda ya hecha; dile a la persona que descargue las
sentencias que le interesen y las adjunte; y sigue con lo que traiga (pasos 3 a 5). **Mientras tanto no cites
ninguna sentencia de memoria**, ni como ejemplo: si la respuesta necesita una, va por el paso 2.

### 7. Una sentencia del Tribunal Constitucional

No está cubierta: las resoluciones del Tribunal Constitucional no están en el CENDOJ. Dilo, y da la dirección de su
buscador para consultarla a mano: https://hj.tribunalconstitucional.es/. No la cites, no le pongas la línea del
paso 2 —esa línea anuncia una consulta en el CENDOJ— y no pidas `cita_preparar` para ella, tampoco si trae su ECLI.

## Comandos

`kitlegal` se invoca desde el `PATH`, y cada herramienta, por el nombre de su fila. Códigos de salida, y entre
paréntesis la `data.clase` del error de la herramienta: 0 correcto —también cuando el documento no es el pedido, que
es un hallazgo y no un fallo— y 2 (`argumentos`) con una referencia mal formada, más de una forma de referencia, un
número de resolución sin su fecha, un texto de búsqueda vacío o un documento sin ficha reconocible. Ante un error de
argumentos, corrige la llamada o pide a la persona lo que falta, y no cites.

<!-- inicio de la tabla de comandos: generada desde --describe con make skills-sync, no editar -->

### `kitlegal cita`

| Orden | Herramienta | Qué hace | Qué devuelve en `data` |
|---|---|---|---|
| `kitlegal cita preparar [<ecli>] [--roj <roj>] [--resolucion <resolucion>] [--fecha <fecha>] [--texto <texto>]` | `cita_preparar` | Dice qué tiene que hacer la persona para encontrar una sentencia en el buscador del CENDOJ: la dirección y cada casilla con su valor, o la dirección de una búsqueda por texto. No consulta nada. | objeto con `referencia`, `texto`, `equivalente`, `cobertura`, `direccion`, `casillas` |
| `kitlegal cita cotejar [<ecli>] [--roj <roj>] [--resolucion <resolucion>] [--fecha <fecha>] [--documento <documento>]` | `cita_cotejar` | Lee la ficha del documento del CENDOJ que trae la persona y dice si es la sentencia que se pidió. No consulta nada. | objeto con `ficha`, `correspondencia`, `pedida`, `es_la_pedida`, `hallazgos` |

La orden y la herramienta de cada fila devuelven el mismo sobre: `ok`, `fuente`, `url`, `fecha_consulta`, `hash`, `data`; con `ok` falso, `data` lleva `clase` y `mensaje`.

Banderas comunes: `--json`, `--timeout <valor>`, `--offline`, `--dry-run`, `--describe`, `--no-graph`, `--asunto <valor>`, `--verbose`.

<!-- fin de la tabla de comandos -->

## Reglas

1. **Nunca inventar contenido legal ni referencias.** Ni una sentencia, ni su ECLI, ni su ROJ, ni su número, ni su
   fecha, ni lo que dice: lo que no esté en un documento de la conversación o no lo devuelva una operación, no se
   afirma.
2. **Ni «existe» ni «no existe».** kitlegal no consulta el buscador: de una sentencia que no está delante, la
   respuesta dice que no está comprobada y cómo buscarla.
3. **La respuesta habla a la persona.** Da la dirección, las casillas y los datos de la ficha con palabras de quien
   busca una sentencia, sin nombrar los campos de lo que devuelve una orden.
4. **Cuando hable de normas**, distingue ley de reglamento, señala lo que una comunidad autónoma puede haber regulado
   de otro modo y no apliques el procedimiento común a lo que la ley regula aparte. El texto de un artículo no se cita
   de memoria: se consulta con `boe-legislacion`.
5. **Ninguna acción con identidad.** No presentes, firmes ni tramites nada en nombre de nadie, ni lo simules.
6. **Sin herramienta y sin binario, ninguna cita.** Si no tienes las herramientas y la orden falla porque `kitlegal`
   no está, no has cotejado nada: no cites ninguna sentencia. La respuesta lleva la línea del paso 2 por cada sentencia
   que necesita y dice que no ha podido preparar la consulta: para eso hace falta instalar kitlegal,
   https://kitlegal.es/instalar/.
