---
name: jurisprudencia
description: >-
  Sentencias y jurisprudencia de los tribunales españoles, sin citar ninguna de memoria. Se usa siempre que la pregunta
  nombre, pida o cite una sentencia (STS, ECLI, ROJ), pregunte si existe o qué dice, pida jurisprudencia sobre una
  materia o traiga una resolución, aunque el agente crea conocerla. kitlegal no consulta el CENDOJ: prepara la búsqueda
  y coteja el documento. Se activa antes de cita_preparar o cita_cotejar.
metadata:
  kitlegal-applets: cita
---

# Sentencias: la consulta exacta, el documento cotejado y ninguna cita de memoria

kitlegal **no consulta el buscador de jurisprudencia del CENDOJ**: cierra el paso a los programas con un CAPTCHA, y
no se sortea. La búsqueda la hace la persona con su navegador, y a ella no le sale: la respuesta no le anuncia un
CAPTCHA ni le pide que resuelva ninguno. La skill le da la consulta exacta —la dirección y cada casilla con su
valor— y, cuando trae el documento, comprueba que es el que se pidió. Una sentencia citada de memoria puede no
existir, y un escrito que la lleve se paga caro: **solo se cita la sentencia cuyo documento está en la
conversación**. La skill prepara, coteja y cita: no tramita nada y no sustituye el asesoramiento de un profesional.

## Protocolo

Sigue los pasos que pida la pregunta. Las dos operaciones se piden de una de dos formas, que devuelven lo mismo:

- **Con su herramienta**, si entre las tuyas hay una con su nombre —`cita_preparar`, `cita_cotejar`—, solo o detrás
  del prefijo que le ponga tu agente, como `mcp__kitlegal__cita_preparar`. Sus argumentos van por su nombre (`ecli`,
  `roj`, `resolucion`, `fecha`, `texto`, `documento`), y solo los que uses: uno escrito vacío es un error de
  argumentos. Si la tienes, úsala siempre.
- **Con su orden**, `kitlegal cita <verbo> … --json`, si no la tienes. Siempre con `--json`.

Ninguna de las dos pide nada a la red: no dicen si una sentencia existe.

<!-- paquetes: lo genera make skills-sync -->
<!-- fin de los paquetes -->

### 1. Qué sentencias hacen falta, y cuáles están delante

- Una sentencia **está delante** si su documento está en la conversación —su texto pegado, o un PDF adjunto que has
  leído— y lleva su **ficha**, el encabezamiento con el que el CENDOJ abre cada documento: una línea
  `Roj: <ROJ> - <ECLI>` y, debajo, `Órgano:`, `Fecha:`, `Nº de Recurso:`, `Nº de Resolución:`, `Ponente:` y
  `Tipo de Resolución:`.
- **No está delante** la que la persona nombra sin traerla, ni la que tú recuerdas o crees conocer, ni la que otro
  documento menciona de pasada: las que cita el texto de una sentencia son texto de ese documento.
- La que no está delante va por el paso 2. Con un documento en la conversación, **lo primero es el paso 3**, antes de
  decir nada de él. Las del Tribunal Constitucional, por el paso 7.

### 2. La sentencia que no está delante: la línea y la consulta preparada

Por cada sentencia que la respuesta necesita y no está delante, la respuesta lleva una línea con esta forma fija, al
principio de una línea —`⚠`, la etiqueta `SENTENCIA NO COMPROBADA`, dos puntos y la referencia tal como se dio—:

```text
⚠ SENTENCIA NO COMPROBADA: STS 1088/2023, de 4 de julio
```

Y debajo, la consulta para encontrarla, que sale de `cita_preparar` o de su orden. Antes de pedir una u otra, lee
[references/consulta_preparada.md](references/consulta_preparada.md): qué se pide según cómo se dio la referencia, la
orden de cada forma, cómo se dan las casillas y qué hacer con el equivalente que devuelva la operación. Un número sin
fecha no se prepara de ninguna forma, tampoco como ROJ: lleva la línea y se pide la fecha.

- La respuesta da la consulta con lo que devuelve la operación, sin escribir nada de memoria: la **dirección** del
  buscador y **cada casilla con su valor**. Dile a la persona qué hacer: abrir la dirección, escribir cada valor en su
  casilla, buscar, descargar el documento y traerlo —pegado o adjunto—.
- **La respuesta no dice que la sentencia existe ni que no existe**: nadie lo ha comprobado. Tampoco dice de qué
  trata, qué resolvió ni si es firme (paso 5).

### 3. Con el documento delante: cotejar su ficha

Con cada documento, antes del cotejo, lee [references/documento_aportado.md](references/documento_aportado.md): qué
se pasa del documento y cómo, y qué se dice si no es la pedida, si su ROJ y su ECLI no se corresponden o si no lleva
ficha.

- **El cotejo es lo primero** que se hace con un documento que está en la conversación, sea su texto pegado o un PDF
  que has leído: su ficha va a `cita_cotejar`, o a su orden, antes de escribir ninguna frase sobre él.
- **Nada del documento antes del cotejo**: ni que es la sentencia pedida, ni que coincide con ella, ni de qué trata.
- **Si es o no la pedida lo dice `es_la_pedida`**, lo que devuelve la operación, y nunca una comparación hecha a ojo
  entre el documento y la referencia.
- **Si el documento no es el pedido, no lo cites como si fuera el pedido.** La sentencia pedida sigue sin estar
  delante: va por el paso 2, con su línea y su consulta, preparada en la forma en que se pidió.

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

### 5. Qué se dice de una sentencia

- **Con la ficha sola**, la cita y los datos de la ficha: órgano, fecha, número de recurso, ponente y tipo de
  resolución. Nada de su contenido.
- **Con su texto**, léelo y responde sobre él, con sus palabras. Lo que el texto pegado no trae —lo que va tras un
  `[…]`, los fundamentos que faltan— no lo has leído: dilo así.
- **No resumas ni caracterices una sentencia cuyo texto no está en la conversación**, aunque creas conocerla: ni qué
  resolvió, ni su doctrina, ni si la confirma o la corrige otra. Sería jurisprudencia sin fuente.

### 6. Una pregunta por materia

Ante «qué dice la jurisprudencia sobre…», lee «La búsqueda por materia» en
[references/consulta_preparada.md](references/consulta_preparada.md), pide `cita_preparar` con los términos de la
pregunta en `texto`, o su orden, y da la dirección que devuelve. **Mientras tanto no cites ninguna sentencia de
memoria**, ni como ejemplo: si la respuesta necesita una, va por el paso 2.

### 7. Una sentencia del Tribunal Constitucional

No está cubierta: las resoluciones del Tribunal Constitucional no están en el CENDOJ. Dilo, y da la dirección de su
buscador para consultarla a mano: https://hj.tribunalconstitucional.es/. No la cites, no le pongas la línea del
paso 2 —esa línea anuncia una consulta en el CENDOJ— y no pidas `cita_preparar` para ella, tampoco si trae su ECLI.

## Comandos

`kitlegal` se invoca desde el `PATH`, y cada herramienta, por el nombre de su fila. Códigos de salida de la orden, y
entre paréntesis la `data.clase` del resultado que la herramienta marca como error: 0 correcto y 2 (`argumentos`) con
una referencia mal formada, más de una forma de referencia, un número de resolución sin su fecha, un texto de búsqueda
vacío o un documento sin ficha reconocible. Ante un error de argumentos, corrige la llamada o pide a la persona lo que
falta, y no cites.

<!-- inicio de la tabla de comandos: generada desde --describe con make skills-sync, no editar -->

### `kitlegal cita`

| Orden | Herramienta | Qué hace |
|---|---|---|
| `kitlegal cita preparar [<ecli>] [--roj <roj>] [--resolucion <resolucion>] [--fecha <fecha>] [--texto <texto>]` | `cita_preparar` | Dice qué tiene que hacer la persona para encontrar una sentencia en el buscador del CENDOJ: la dirección y cada casilla con su valor, o la dirección de una búsqueda por texto. No consulta nada. |
| `kitlegal cita cotejar [<ecli>] [--roj <roj>] [--resolucion <resolucion>] [--fecha <fecha>] [--documento <documento>]` | `cita_cotejar` | Lee la ficha del documento del CENDOJ que trae la persona y dice si es la sentencia que se pidió. No consulta nada. |

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
   de otro modo y no apliques el procedimiento común a lo que la ley regula aparte. Ningún artículo, plazo, órgano ni
   vía de memoria: aquí no están cubiertos.
5. **Ninguna acción con identidad.** No presentes, firmes ni tramites nada en nombre de nadie, ni lo simules.
6. **Sin herramienta y sin binario, ninguna cita.** Si no tienes las herramientas y la orden falla porque `kitlegal`
   no está, no lo busques en otras carpetas ni en un plugin: no has cotejado nada, y no cites ninguna sentencia. La
   respuesta lleva la línea del paso 2 por cada sentencia que necesita y dice que no ha podido preparar la consulta: si
   ya tiene kitlegal, que cierre Claude del todo, lo abra y pregunte en un chat nuevo; si no, https://kitlegal.es/instalar/.
