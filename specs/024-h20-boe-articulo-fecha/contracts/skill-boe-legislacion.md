# Contrato: `boe-legislacion` v0.2

Qué cambia en `skills/boe-legislacion/SKILL.md`, qué no, cómo cabe y qué salida del binario consume. El prototipo
entero está en [SKILL-boe-legislacion.prototipo.md](./SKILL-boe-legislacion.prototipo.md): 298 líneas, con la fila de
la tabla como la regenerará `make skills-sync`; con la fila de `main`, los siete tests de `make skills-check` pasan sobre
él (research M6). La tarea puede escribirlo de otra manera; lo que fija este contrato son las formas fijas, las reglas
y los límites.

## 1. Las formas fijas (FR-041, FR-043, FR-048)

Van tal cual, cada una en su bloque `text`, y un guion las compara sin modelo (contracts/evals-redaccion-fechada.md §2):

```text
⚠ REDACCIÓN A FECHA: <forma legible> [<identificador>, bloque <id>]: se cita la redacción vigente el AAAA-MM-DD, la de fecha de vigencia AAAAMMDD.
```

```text
⚠ REDACCIÓN DE HOY: <forma legible> [<identificador>, bloque <id>]: hoy la redacción es otra, la de <identificador>, con fecha de vigencia AAAAMMDD.
```

```text
⚠ SIN REDACCIÓN VIGENTE: <forma legible> [<identificador>, bloque <id>]: el AAAA-MM-DD no tenía redacción vigente; la tiene desde el AAAAMMDD, la fecha de vigencia de su primera redacción.
```

Con los datos de US1, las dos primeras quedan:

```text
⚠ REDACCIÓN A FECHA: art. 118 de la LCSP [BOE-A-2017-12902, bloque a1-30]: se cita la redacción vigente el 2019-06-01, la de fecha de vigencia 20180309.
⚠ REDACCIÓN DE HOY: art. 118 de la LCSP [BOE-A-2017-12902, bloque a1-30]: hoy la redacción es otra, la de BOE-A-2020-1651, con fecha de vigencia 20200206.
```

y con los de US2, la tercera:

```text
⚠ SIN REDACCIÓN VIGENTE: art. 21 de la Ley 39/2015 [BOE-A-2015-10565, bloque a21]: el 2016-06-01 no tenía redacción vigente; la tiene desde el 20161002, la fecha de vigencia de su primera redacción.
```

La de hoy no dice que rija ni que esté vigente el día de la consulta; con una vigencia futura, su fecha lo dice
(FR-041, clarificación 3). Ninguna dice hasta cuándo rigió la citada.

## 2. Lo que dice la skill, requisito a requisito

| Requisito | Dónde, en el prototipo | Qué dice |
|---|---|---|
| FR-040 | «Redacción en una fecha», párrafo de entrada; paso 3 | con una fecha —un día, un año o un mes, o un hecho cuyo día dan la pregunta o la conversación—, cada bloque se lee con `--fecha` o con `boe_articulo` y `fecha`, y se cita esa redacción como en «Cómo se cita»; con un hecho sin día, se pide el día y no se lee ni se cita ninguna redacción como la de ese momento |
| FR-041 | primera y segunda viñeta | la línea A siempre tras un 0; la B si la lectura da la redacción de hoy; nada de hasta cuándo |
| FR-042 | segunda viñeta | de la redacción de hoy, nada más salvo leerla sin fecha (paso 3) y citarla |
| FR-043 | tercera viñeta | con el 3 que da la fecha de vigencia de la primera redacción, la línea C, sin citar el texto como el de entonces ni decir qué regía sin leerlo; con otro 3, la regla 2, que gana la razón «no fecha sus redacciones» (el 3 de FR-004) |
| FR-044 | primera viñeta | sus avisos, los de la norma hoy (regla 3) |
| FR-045 | párrafo de entrada | con un año o un mes, dos lecturas, el primer y el último día; la misma fecha de vigencia y la misma norma, o los dos 3, la línea de cada una; si no, ninguna cita y se pide el día |
| FR-046 | paso 3, «Redacción modificada», regla 7 | sin fecha, como hoy; la lectura fechada va sin `kitlegal graph check` (research D17), y «Redacción modificada» y la regla 7 hablan solo de la lectura sin fecha; las tres líneas, solo tras una lectura con fecha |
| FR-047 | `description` | gana «o de lo que decía en una fecha —la del acto, la de los hechos—» y «el texto, hoy o en una fecha»; 1 024 caracteres como mucho |
| FR-048 | todo el fichero | menos de 300 líneas; la tabla, la regenerada; `references/`, sin cambios |
| FR-049 | «Redacción modificada» | «`kitlegal boe articulo` sin fecha da solo la redacción de hoy» |

## 3. Cómo cabe en 299 líneas

`main` tiene 299; la sección nueva, 28. El prototipo hace sitio sin quitar ninguna regla. Líneas por sección,
contadas en esta sesión en `main` y en el prototipo:

| Sección | `main` | Prototipo | Qué se comprime y por qué no se pierde nada |
|---|---|---|---|
| Cabecera y `description` | 18 | 17 | la `description` gana las preguntas con fecha y, para caber en 1 024 caracteres, pierde «régimen jurídico del sector público» de la lista de materias (queda «LRJSP» entre las abreviaturas; ninguna pregunta de las evals de hoy depende de esas palabras) y abrevia el final |
| Título y entrada | 9 | 8 | la entrada sin la lista de materias, que da la `description`: presentación, no regla |
| 2. Resolver | 13 | 12 | las dos últimas viñetas en una, con el mismo texto |
| 3. Leer | 45 | 38 | «Con herramientas…» en una línea; el ejemplo de `boe articulos` y su forma PowerShell en código en línea, que el test de las formas PowerShell lee igual (`internal/evals/conjunto_test.go:3040-3108`); el ejemplo de órdenes de metadatos y análisis, con una de las dos; gana la frase de la lectura con fecha |
| 4. Evaluar | 11 | 8 | sin la frase de entrada, que dice el título; la última frase como viñeta |
| 5. Responder | 37 | 32 | «Repasa cada cita…» pasa a la entrada de «Cómo se cita», cuyas viñetas tienen sus tres reglas; la viñeta de la cita sin el detalle que «Cómo se cita» repite |
| Cómo se cita | 36 | 32 | las etiquetas de los avisos en la frase de entrada; dos viñetas en una, con el mismo texto |
| Redacción en una fecha | — | 28 | nueva |
| Redacción modificada | 36 | 30 | la forma en prosa, que repetía el bloque `text`, en una frase; sin los blancos alrededor del bloque; la última viñeta unida a la anterior; `fuente-caducada` en una línea |
| Comandos y tablas | 36 | 36 | la fila de `boe articulo`, regenerada, en la misma línea |
| Reglas | 39 | 38 | la 2 gana una razón; la 7 dice «la lectura sin fecha» y deja la remisión que repetía |
| Total | 299 | 298 | |

Ninguna regla pierde lo que manda; lo leen los jueces de la revisión final (FR-048).

## 4. Lo que consume de la salida (de fuera adentro)

| Lo que lee | De dónde | Para qué | Veces por pregunta |
|---|---|---|---|
| El texto, la cita, los avisos, la fecha de vigencia y la norma modificadora de la versión elegida | el sobre de la lectura fechada | citar esa redacción y su línea A | una por bloque; dos con un año o un mes sin día |
| La fecha de vigencia y la norma de la redacción de hoy | `redaccion_posterior` de ese sobre | la línea B | la misma lectura |
| La fecha de vigencia de la primera redacción | el mensaje del 3 de FR-003 | la línea C | la misma lectura |
| El texto de hoy | una lectura sin fecha, con su `kitlegal graph check` | decir qué dice hoy, solo si la respuesta lo dice | como mucho una por bloque |

La prosa no nombra los campos (`expresiones-prohibidas.yaml`, `salida_de_las_herramientas`): dice «la redacción de hoy»,
«su norma» y «su fecha». `SKILL.md` lo lee el agente una vez por conversación y no da señales.
