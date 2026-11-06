# La consulta preparada

La consulta que prepara `cita_preparar`, o su orden, para la sentencia que no está delante (paso 2 de `SKILL.md`) y
para una pregunta por materia (paso 6). La línea que va encima de la consulta, lo que se pide en cada caso y lo que se
le dice a la persona que haga son los del paso 2; aquí está el detalle.

## La orden de cada forma

Sin la herramienta, cada forma en que se dio la referencia tiene su orden:

| Cómo se dio | Orden |
|---|---|
| Un ECLI (`ECLI:ES:TS:2023:3144`) | `kitlegal cita preparar ECLI:ES:TS:2023:3144 --json` |
| Un ROJ dicho como tal («ROJ: STS 3144/2023») | `kitlegal cita preparar --roj "STS 3144/2023" --json` |
| Número y fecha («STS 1088/2023, de 4 de julio») | `kitlegal cita preparar --resolucion 1088/2023 --fecha 2023-07-04 --json` |

Un número sin fecha no tiene orden, tampoco como ROJ: los ROJ son correlativos, y ese mismo número es el ROJ de otra
sentencia. Por eso la respuesta lleva la línea del paso 2 y pide la fecha.

## Las casillas

La operación devuelve la dirección del buscador (`direccion`) y cada casilla con su valor (`casillas`): el nombre que
la persona ve en la pantalla, la parte de la casilla si tiene dos —«Desde» y «Hasta»— y lo que hay que escribir.

## El equivalente

Si la operación da además un **equivalente** —el ROJ de un ECLI o el ECLI de un ROJ—, puedes nombrarlo junto a la
referencia, y entonces dilo como lo que es: deducido de esa referencia, sin que nadie lo haya comprobado. No lo
compongas tú: solo se deduce para el Tribunal Supremo, y cuando no viene no lo hay.

## La búsqueda por materia

Ante «qué dice la jurisprudencia sobre…», `cita_preparar` lleva los términos de la pregunta en `texto`; sin la
herramienta, su orden:

```bash
kitlegal cita preparar --texto "cláusula suelo" --json
```

La dirección que devuelve abre el buscador con esa búsqueda ya hecha. Dásela a la persona, dile que descargue las
sentencias que le interesen y las adjunte, y sigue con lo que traiga (pasos 3 a 5).
