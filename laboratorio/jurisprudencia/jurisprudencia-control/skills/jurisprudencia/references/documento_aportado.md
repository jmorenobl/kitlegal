# El documento aportado

El documento que trae la persona, su texto pegado o un PDF (paso 3 de `SKILL.md`). El cotejo sigue siendo lo primero:
nada del documento antes de que `cita_cotejar`, o su orden, lea su ficha. Aquí está qué se le pasa y cómo, qué
devuelve y qué se dice cuando no es la pedida, cuando su ROJ y su ECLI no se corresponden y cuando no lleva ficha.

## La ficha y nada más

Pasa a `cita_cotejar`, en `documento`, **la ficha y nada más**: sus líneas tal como están en el documento, desde la
de `Roj:` hasta la de `Tipo de Resolución:`, con las que haya entre ellas. La operación solo lee la ficha: el resto
del texto no cambia lo que devuelve, y copiarlo alarga cada cotejo con el tamaño de la sentencia.

Con la orden, la ficha va por la entrada estándar, o en `--documento` si tu shell no tiene `<<`:

```bash
kitlegal cita cotejar --roj "STS 1088/2023" --json <<'DOCUMENTO'
Roj: STS 3144/2023 - ECLI:ES:TS:2023:3144
…las demás líneas de la ficha, hasta la de «Tipo de Resolución:»…
DOCUMENTO
```

Un PDF lo lees tú y pasas su ficha: kitlegal no abre ficheros.

## Lo que devuelve

Los ocho datos de la ficha (`ficha`), si el documento es el pedido (`es_la_pedida`, solo si diste una referencia) y,
si no lo es, un hallazgo que dice qué difiere (`hallazgos`).

## El documento que no es el pedido

Si es o no la pedida lo dice `es_la_pedida`, no una comparación hecha a ojo, que falla cuando el número de resolución
del documento es el ROJ pedido: 1088/2023 es el número de resolución del documento del ROJ STS 3144/2023, y el ROJ
STS 1088/2023 es el de otra sentencia.

Si el documento no es el pedido, dilo con lo que difiere, con las palabras del hallazgo —por ejemplo, que 1088/2023
es el número de resolución de ese documento y no su ROJ—, y no lo cites como si fuera el pedido. La sentencia pedida
va por el paso 2, con su línea y su consulta.

## El ROJ y el ECLI que no se corresponden

Mira también `correspondencia`. Con `no-se-corresponden`, el ROJ y el ECLI de la ficha no son de la misma
resolución: **no cites el documento**, di que no se corresponden, da los dos tal como están en la ficha y pide a la
persona que vuelva a descargar el documento del buscador y lo traiga. Con cualquier otro valor, cita y no digas nada
de esto.

## Un texto sin ficha reconocible

Si el cotejo termina con el código 2 (`argumentos`) porque el texto no lleva una ficha reconocible, mira el dato que
nombra el mensaje. Si está en el documento, no pasaste la ficha entera: repite el cotejo con todas sus líneas. Si no
está, no hay sentencia que citar: di qué falta —el encabezamiento del documento del CENDOJ, o ese dato— y pide el
documento completo. Sobre un texto sin ficha puedes responder, sin atribuirlo a una sentencia identificada.
