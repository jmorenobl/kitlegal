# Redacción en una fecha

Cómo se lee y se cita la redacción de un bloque vigente en una fecha, y cuándo lleva la respuesta cada una de las tres
líneas de «Redacción en una fecha» de `SKILL.md`.

## Qué es una fecha

Una pregunta lleva una fecha si da un día, un mes o un año, o un hecho cuyo día dan la pregunta o la conversación: la
del acto, la de los hechos. Entonces se cita la redacción vigente ese día, no la de hoy.

## Cómo se lee

Cada bloque, con `--fecha AAAA-MM-DD` detrás, como en
`kitlegal boe articulo BOE-A-1985-5392 a22 --fecha 2005-06-15 --json`, o con `boe_articulo` y el día en `fecha`. La
lectura con fecha va sin `kitlegal graph check`.

## Un hecho sin día

Si un hecho sitúa el momento y nadie da su día, pide el día: no leas ni cites ninguna redacción como la de entonces.

## Un año o un mes

Con un año o un mes sin día, lee con su primer día y con el último. Si las dos lecturas dan la misma fecha de vigencia y
la misma norma que la dio, o las dos terminan con el `3` de «Sin redacción vigente», responde con la línea de cada una;
si no, no cites ninguna y pide el día.

## Lo que se cita

Lo leído con fecha se cita como en «Cómo se cita» de `SKILL.md`, con sus avisos, que son los de la norma hoy (regla 3),
y con la línea REDACCIÓN A FECHA, con la fecha pedida y la de vigencia de lo citado tal como las da la lectura. De la
redacción citada, la prosa dice qué norma la dio y desde cuándo rige; hasta cuándo, nunca: ninguna lectura lo trae,
tampoco en una norma derogada, cuyo aviso no lleva fecha, y entre ella y la de hoy puede haber otras.

## La redacción de hoy

Si el `data` de la lectura trae `redaccion_posterior`, hoy la redacción es otra, y va también la línea REDACCIÓN DE HOY,
con la norma y la fecha de vigencia de `redaccion_posterior`. De la redacción de hoy, la lectura da su norma y su
fecha, y esa línea es el único sitio de la respuesta en que van: «hay otra redacción, dada por <identificador>, de
<fecha>» no la sustituye, y repetido junto a ella no añade nada. Qué dice, qué cambió o si rige ya no lo has leído,
salvo que la leas sin fecha (paso 2 de `SKILL.md`).

## Sin redacción vigente

Si la lectura termina con `3` y su mensaje da la fecha de vigencia de la primera redacción, ese día el precepto no
tenía redacción vigente: no cites su texto como el de entonces ni digas qué regía sin haberlo leído, y responde con la
línea SIN REDACCIÓN VIGENTE, con la fecha del mensaje. Con otro `3`, vale la regla 2.

## Las tres líneas

Las tres salen solo de una lectura con fecha, cada una con su forma fija, entera y en una sola línea, tal como la
escribe `SKILL.md`.
