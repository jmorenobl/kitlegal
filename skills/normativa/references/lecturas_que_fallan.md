# Lecturas que fallan

Qué se hace cuando una orden del protocolo no termina con `0`, o su herramienta da un error: qué dice cada código, cómo
se sigue según lo que falló y qué lleva entonces la respuesta.

## Los códigos

Cada orden termina con `0` si todo fue bien y, si no, con el código de lo que falló; una herramienta da un error, y su
`data.clase` lo nombra:

- `2` o `argumentos`: los argumentos no son válidos.
- `3` o `no-encontrado`: no encuentra lo pedido.
- `4` o `fuente-no-disponible`: la fuente no está disponible.
- `5` o `limite-o-tos`: la fuente limita el ritmo de las consultas.
- `6` o `identidad-humana`: hace falta la identidad de una persona.
- `1` o `inesperado`: un fallo inesperado, como un `world.db` que no se puede leer en `kitlegal graph`.

Una lectura de `kitlegal boe` falla si termina con `3`, `4` o `5`, o si no hay caché con `--offline`: vale la regla 2
de `SKILL.md`.

## Un bloque que no está

Si `kitlegal boe` termina con `3`, vuelve al índice en lugar de probar otros ids: el id es el de la entrada cuyo
`titulo` es el artículo, no su número (en la Ley 9/2017, «Artículo 118» es `a1-30`; `a118` no está). Si el artículo no
existe en la norma, dilo.

## Una lectura de varios bloques

Falla entera en cuanto falla uno de sus bloques. Si termina con `4` o `5`, pide cada bloque por separado con
`kitlegal boe articulo` (`boe_articulo`) antes de dar ninguno por no consultado: el fallo de uno no impide leer los
demás. Que no se pudieron consultar, dilo solo después de pedirlos así.

## Qué dice la respuesta

De lo que no se pudo leer, la respuesta dice qué bloques, por su número, y por qué: no están en la norma, la fuente no
estaba disponible, limitó las consultas o no fecha sus redacciones. No da el código, ni de qué tratan esos bloques, ni
para qué los pedías, y no suple su texto.

## Si `kitlegal graph check` falla

Si la lectura sin fecha devuelve el texto del bloque y, detrás, `kitlegal graph check` termina con otro código que `0`
—o `graph_check` da un error—, la respuesta cita igual el texto leído con `kitlegal boe` y lleva esta frase, tal cual y
sin nada más sobre la redacción:

```text
No se ha podido comprobar si la redacción ha cambiado desde una consulta anterior.
```

La frase es solo de ese caso: con `0`, la respuesta no la lleva, ni afirmada, ni negada, ni con otras palabras. Si lo
que falla es la lectura, la orden termina ahí, sin `kitlegal graph check`, y vale la regla 2.
