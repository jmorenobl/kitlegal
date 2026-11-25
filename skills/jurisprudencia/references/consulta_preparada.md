# La consulta preparada

La consulta que prepara `cita_preparar`, o su orden, para la sentencia que no está delante (paso 2 de `SKILL.md`) y
para una pregunta por materia (paso 6). La línea que va encima de la consulta y lo que se le dice a la persona que
haga son los del paso 2; aquí está el detalle, y lo que se pide según cómo se dio la referencia.

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

## Lo que se pide

Lo que se pide depende de cómo se dio la referencia:

| Cómo se dio | Qué se pide |
|---|---|
| Un ECLI (`ECLI:ES:TS:2023:3144`) | `ecli` |
| Un ROJ dicho como tal («ROJ: STS 3144/2023») | `roj` |
| Número y fecha («STS 1088/2023, de 4 de julio») | `resolucion` y `fecha` |
| Número sin fecha («la STS 1088/2023») | nada: **sin fecha no se prepara de ninguna forma**, tampoco como ROJ; lleva la línea y se pide la fecha |

- Una cita escrita como «STS 1088/2023, de 4 de julio» lleva el **número de resolución**, no el ROJ: se prepara con
  `resolucion` y con `fecha` en la forma `AAAA-MM-DD`. El año de la fecha es el del número, salvo que la cita diga otro.
