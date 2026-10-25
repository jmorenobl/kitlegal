# Redacción modificada

Qué compara `kitlegal graph check` tras una lectura sin fecha, y cuándo y cómo lleva la respuesta la línea de
«Redacción modificada» de `SKILL.md`.

## Qué compara

`kitlegal` recuerda en local los bloques que ha leído sin fecha y qué redacción vio cada vez.
`kitlegal graph check <norma> <bloques>... --json` (`graph_check`) compara, para esa norma y esos bloques, la redacción
que acabas de leer con la que se leyó la vez anterior, en otra conversación que quien pregunta no conoce, y devuelve en
`data.hallazgos` una entrada por cada bloque en que encuentra algo, con su `clase`.

## La línea, por bloque

Una entrada de `clase` `version-obsoleta` dice que la redacción que acabas de leer de ese bloque no es la que se leyó la
vez anterior. La respuesta lo dice con una línea REDACCIÓN MODIFICADA por bloque, con la cita del bloque como en «Cómo
se cita» y las dos fechas de vigencia de la entrada: la superada y, después, la de lo que acabas de leer. Con dos
bloques en `version-obsoleta`, dos líneas, cada una con su cita y sus fechas. Decirlo con otras palabras no vale, ni en
lugar de la línea ni además de ella: la línea va con su forma fija y lo dice entera.

## No es un aviso de vigencia

Un aviso es un dato de la norma, que el BOE da con el bloque, y la etiqueta REDACCIÓN MODIFICADA no es la de ninguno:
una entrada de `kitlegal graph check` es un dato de kitlegal sobre otras conversaciones, que quien pregunta no conoce.
Mencionarla fuera de su línea sería contar lo que devolvió una orden y hablar de otras conversaciones.

## La redacción superada

No la has leído. Ni la lectura sin fecha, que da solo la redacción de hoy, ni `kitlegal graph check`, que da dos
fechas, dicen qué decía la superada ni en qué se diferencia de la vigente: la respuesta no lo dice, ni lo resume, ni lo
compara, aunque creas saberlo, que sería texto legal sin fuente.

## Si se pregunta qué cambió

La respuesta dice que cita la redacción vigente y que la de antes no la ha leído.
