# Contrato: las tres skills reescritas (FR-020 a FR-024)

Lo que consume cada skill es su propio texto: el agente carga la descripción en toda sesión, el cuerpo de `SKILL.md`
al activarla y una referencia cuando el cuerpo se lo dice. Este contrato fija, por skill, el reparto que implementan
las tareas: los apartados del cuerpo, las referencias a mano con sus apartados y la línea del cuerpo que dice cuándo
leer cada una, lo que tiene que quedarse en el cuerpo, los ejemplos que cambian y un borrador de la descripción en
tercera persona. La tabla de cada regla de `main` con su sitio está en research.md, «Tabla de reglas (FR-021)».

## Reglas comunes a las tres

1. **Cuerpo** (FR-004, FR-020): menos de 500 líneas y como mucho 11 000 caracteres contados como runas, con la región
   de la tabla de comandos como la deja `make skills-sync`, que no cambia. Lleva el protocolo, las reglas invariantes
   (`## Reglas`, con su número de hoy), las formas fijas que las evals comprueban sin modelo (las líneas `⚠`, en el
   bloque `text` en que las piden los controles), la tabla de comandos y una línea por referencia —a mano o generada—
   con un enlace de Markdown `[…](references/<fichero>.md)` fuera de los bloques de código, que dice cuándo leerla.
2. **Descripción** (FR-005, FR-022): tercera persona, qué hace la skill y cuándo se usa, como mucho 900 caracteres, sin
   ninguna etiqueta XML. «Se usa…», «Se activa…», «el agente…»; una pregunta citada entre «» puede llevar primera o
   segunda persona, y un verbo con la skill de sujeto («Consulta», «Calcula», «Lee») es tercera persona.
3. **Referencias a mano** (FR-012, FR-013): `references/<fichero>.md`, nombres en `snake_case` como las generadas, sin
   la cabecera «generado desde», sin declararlas en `kitlegal-referencias`; cada una abre con un párrafo que dice lo que
   lleva; ninguna enlaza otra; ninguna repite lo que sale de `data/`; ninguna lleva una línea `⚠` (nombran la forma del
   cuerpo por su etiqueta y su apartado), y si alguna la llevara, sería igual a una línea del cuerpo (D16). Las de más
   de 100 líneas abren con su índice (FR-009); ninguna de las previstas pasa de 100.
4. **Lo que no cambia** (FR-024): el `name`, `kitlegal-applets`, `kitlegal-referencias` y el contenido de las
   referencias generadas.
5. **Órdenes de ejemplo** (FR-019): toda orden concreta —`kitlegal boe`, un verbo y al menos un argumento— del cuerpo
   o de una referencia a mano es de lo que sirve la caché de una sesión (V3, V4: `indice`, `metadatos` y `articulo`,
   con y sin `--fecha`, de las normas de `data/normas.yaml`; ningún `buscar` ni `analisis`); las demás llevan
   marcadores (`<términos>`, `<norma>`). La mención del verbo sin argumentos, como «`kitlegal boe articulo` o
   `articulos`» de la regla 6, no es una orden y se queda como está (contracts/controles-de-contenido.md §10).
6. **Prosa** (FR-016, FR-017): en `boe-legislacion`, ni el cuerpo ni sus referencias a mano llevan una expresión de
   `evals/boe-legislacion/expresiones-prohibidas.yaml` fuera del código ni una fecha AAAAMMDD con cifras; en las tres,
   toda norma nombrada como «Ley N/AAAA», «Real Decreto…» empieza un título de `data/normas.yaml`, y nada lleva los
   términos de las instrucciones de las evals.

Las cifras de cada reparto son de un borrador completo escrito en la sesión del plan y medido con `wc -lm` (research
V2), no versionado: no son el texto final. Cada tarea escribe el suyo, y los gates de `make ci` lo miden.

## boe-legislacion

### Estructura

#### Cuerpo de `SKILL.md`

Las cifras son de un borrador del cuerpo escrito entero con las reglas de la tabla (borrador medido en la sesión del
plan, no versionado): runas con `LC_ALL=en_US.UTF-8 wc -m`, líneas de 120 columnas y la sangría de las listas.

| Apartado | Contenido | Runas |
|---|---|---|
| línea en blanco, `# Consultar y citar legislación consolidada del BOE` y `## Protocolo` con su párrafo | pasos en orden; orden o, si la tienes, siempre la herramienta de su fila, tras leer la referencia de las herramientas; sin `fuente`, `url`, `fecha_consulta` y `hash`, no hay cita | 379 |
| paso 1, **Identificar la norma** | tabla de normas antes de nada; `BOE-A-…` de ahí; no exhaustiva; si no está, la referencia de búsqueda | 229 |
| paso 2, **Leer** | índice y copia del id (ejemplo de la Ley 9/2017); ningún id fuera del índice; una lectura sin fecha por bloque, con `kitlegal graph check` detrás; la orden de `articulo` en Bash y en PowerShell; con fecha, «Redacción en una fecha»; varios bloques y orden que falla, sus referencias | 1222 |
| paso 3, **Completar** | remisiones; metadatos y análisis si la pregunta depende de la vigencia o de una modificación | 210 |
| paso 4, **Responder** | B48–B57, B63, B67–B69 | 992 |
| paso 5, **De un precepto no leído, nada** | B58–B62, con un ejemplo de afirmar y otro de negar | 631 |
| `## Cómo se cita` | forma, ejemplo en línea, «sin nada tras el id», una por bloque, la referencia de otras formas, la línea `⚠` entera | 457 |
| `## Redacción en una fecha` | cuándo leer la referencia; las tres formas fijas en un bloque `text`; que solo salen de una lectura con fecha | 793 |
| `## Redacción modificada` | la forma fija; nada más de `kitlegal graph` (ni `fuente-caducada`, ni nada de un bloque sin `version-obsoleta`); solo `check`; cuándo leer la referencia | 641 |
| `## Comandos` | `kitlegal` desde el `PATH` | 53 |
| región generada | sin cambios (FR-024) | 3076 |
| `## Reglas` | las ocho, con su número de hoy; la 3 con las tres etiquetas de aviso; la 7 sin la frase; la 8 con la línea sin consulta | 2198 |
| **Total** | 130 líneas | **10 881** |

Son 11 030 runas con las reglas en su sitio menos unas 150 de los tres ejemplos sin regla que se retiran (las notas
los dan): quedan 119 de margen hasta el límite de 11 000. Es la skill más ajustada: la región generada (3076) se queda
como está (FR-024), las formas fijas y la orden de `articulo` con su forma para PowerShell suman unas 1750, y la
prosa, unas 6050, va escrita al límite; la tarea mide cada apartado mientras lo escribe. Llevar las reglas 4 y 5 a una referencia queda descartado: el ADR
0046, decisión 1, deja las reglas invariantes en el cuerpo.

#### Referencias escritas a mano

`references/` es plano; ninguna referencia enlaza otra, y ninguna lleva el carácter `⚠`: nombran las formas de
`SKILL.md` por su etiqueta y su apartado. `normas.md` sigue generada, sin cambios, y se enlaza desde el paso 1. Todas
tienen menos de 100 líneas, así que ninguna necesita índice.

| Fichero | Apartados (`## `) | Líneas | Línea del cuerpo que dice cuándo leerla |
|---|---|---|---|
| `con_herramientas.md` | Nombres · Argumentos · Las órdenes del protocolo · Encadenar y fallar | ~25 | «Protocolo»: «Sigue los pasos en orden, cada operación con su orden, `kitlegal <applet> <verbo> … --json`, o, si la tienes, siempre con la herramienta de su fila en «Comandos», tras leer [con_herramientas](references/con_herramientas.md).» |
| `buscar_la_norma.md` | Buscar por el título · Varias posibles · Normas autonómicas · Si no aparece | ~20 | paso 1: «Si no está, se busca por su título con [buscar_la_norma](references/buscar_la_norma.md).» |
| `varios_bloques.md` | Cuándo · La orden | ~12 | paso 2: «Para varios bloques a la vez, [varios_bloques](references/varios_bloques.md);» |
| `lecturas_que_fallan.md` | Los códigos · Un bloque que no está · Una lectura de varios bloques · Qué dice la respuesta · Si `kitlegal graph check` falla | ~40 | paso 2: «si una orden, también `kitlegal graph check`, no termina con `0`, antes de seguir, [lecturas_que_fallan](references/lecturas_que_fallan.md).» (la regla 7 la enlaza también) |
| `como_se_cita.md` | Lo que hace cita · La forma legible dentro · Lo que no es cita · Sola o bajo una cita textual · Ids que terminan en punto · Bloques remitidos · Los avisos | ~28 | «Cómo se cita»: «Una cita por bloque; repasa cada una, y para otra forma, [como_se_cita](references/como_se_cita.md).» |
| `redaccion_en_una_fecha.md` | Qué es una fecha · Cómo se lee · Un hecho sin día · Un año o un mes · Lo que se cita · La redacción de hoy · Sin redacción vigente · Las tres líneas | ~40 | «Redacción en una fecha»: «Con una fecha en la pregunta —un día, un mes, un año o un hecho fechado—, lee antes [redaccion_en_una_fecha](references/redaccion_en_una_fecha.md): cómo se lee, sin `kitlegal graph check`, y cuándo va cada una de estas líneas, que solo salen de una lectura con fecha:» |
| `redaccion_modificada.md` | Qué compara · La línea, por bloque · No es un aviso de vigencia · La redacción superada · Si se pregunta qué cambió | ~25 | «Redacción modificada»: «Con `version-obsoleta`, o si se pregunta qué cambió, [redaccion_modificada](references/redaccion_modificada.md).» |

Qué lleva cada una:

- `con_herramientas.md`: B13 (el prefijo), B14, B16, B17, B32, B41 (`boe_metadatos`, `boe_analisis`), B82 (`fecha`)
  y B107, y cómo se llama con herramientas cada orden del protocolo, por el nombre de su fila.
- `buscar_la_norma.md`: B21–B26, con `kitlegal boe buscar <términos> --json`.
- `varios_bloques.md`: B37 y la parte de `articulos` de B31: la orden de Bash
  `kitlegal boe articulos <norma> <bloques>... --json && kitlegal graph check <norma> <bloques>... --json` y, detrás, su
  forma para PowerShell; con herramientas, `boe_articulos` y `graph_check` con esos mismos bloques. El control de las
  órdenes para PowerShell lee la skill entera (FR-016, FR-018), así que la orden de Bash de `articulos` cumple aquí la
  de «una orden de Bash por verbo».
- `lecturas_que_fallan.md`: B39, B43, B108, B109, B112–B115 y B125–B128, con la frase de la regla 7 sola en un
  bloque `text`.
- `como_se_cita.md`: B72–B78 en detalle, con los ejemplos de cita de hoy.
- `redaccion_en_una_fecha.md`: B81–B94, con `kitlegal boe articulo BOE-A-1985-5392 a22 --fecha 2005-06-15 --json`.
- `redaccion_modificada.md`: B98, B99, B101, B104, B105 y la explicación de l. 200–203.

### Notas

#### Ejemplos que cambian

Ninguno es una regla.

- `kitlegal boe buscar régimen jurídico del sector público --json` pasa a `kitlegal boe buscar <términos> --json`, en
  references/buscar_la_norma.md: la caché de las sesiones no sirve `buscar`.
- `kitlegal boe analisis BOE-A-2017-12902 --json` se retira: el paso 3 nombra los metadatos y el análisis, la tabla de
  comandos da su orden y references/con_herramientas.md sus herramientas. La caché no sirve `analisis`; si se quiere un
  ejemplo, `kitlegal boe metadatos BOE-A-2017-12902 --json` sí se sirve.
- `kitlegal boe articulo BOE-A-1985-5392 a22 --fecha 2005-06-15 --json` pasa a references/redaccion_en_una_fecha.md
  (se sirve). `kitlegal boe indice BOE-A-2015-10565 --json` y la orden de `a21` de la LPAC se quedan en el cuerpo; la
  de `articulos`, con su forma para PowerShell, pasa a references/varios_bloques.md.
- La cita de ejemplo `art. 21 de la Ley 39/2015 [BOE-A-2015-10565, bloque a21]` pasa de un bloque `text` a código en
  línea; `[art. 20.1 de la LTAIBG, …]`, `[BOE-A-2015-10565, bloque a21, art. 21]` y la cita de la Constitución pasan a
  references/como_se_cita.md.
- Se retiran el bloque `⚠ NORMA DEROGADA: esta norma ha sido derogada.` (las tres etiquetas siguen en la regla 3), el
  par «el artículo no lleva avisos de vigencia» / «la lectura no trajo avisos», «Ley 40/2015» y «LRJSP» del paso 1,
  «los artículos que regulan <materia> (N y M)» del paso 5 y la explicación «quien acaba de instalar…» de la regla 8.
- Los tres ejemplos sin regla que se retiran para el margen: en el paso 1, «por su nombre, su número y año o su
  abreviatura» (las formas de nombrar una norma siguen en la descripción, B06); en el paso 5, «el art. N sobre
  <materia>» (quedan «el art. N, que es…» y «no puedo decir qué <regla> fija»); y en la regla 5, «(competencias
  compartidas o cedidas, desarrollo autonómico, régimen foral)».
- `mcp__kitlegal__boe_articulo` y el `world.db` ilegible pasan a sus referencias. «En PowerShell, que en su versión 5.1
  no tiene `&&`» queda como «En PowerShell 5.1, sin `&&`». Las tres formas fechadas van juntas en un bloque `text` (el
  control solo pide que cada una esté en un bloque `text`).
- En la descripción, «tributos y haciendas locales» queda como «tributos», y «o pedirse el texto, hoy o en una fecha,
  de una norma estatal o autonómica» se funde con «lo que decía en una fecha» y «La norma, estatal o autonómica».

#### Riesgos

- Reglas que se quedan en el cuerpo porque las necesita toda respuesta: B49–B62 (la respuesta, otra conversación y el
  precepto no leído, que juzga `afirma_lo_no_leido`), B12, B27–B29, B34–B36, B38, B42, B55, B66, B79, B80, B97, B100,
  B102, B103, B110–B124 y B130–B132. También las formas fijas: las tres etiquetas de aviso, `⚠ REDACCIÓN MODIFICADA:`,
  las tres fechadas en un bloque `text`, la línea sin consulta sola en su bloque `text` y la orden de `articulo` con
  su forma para PowerShell.
- Paso 5 con un ejemplo de afirmar y uno de negar, en lugar de los tres de afirmar de hoy: lo decide la tasa de
  `afirma_lo_no_leido`.
- Modo con herramientas (la mitad de las sesiones del job): depende de que el agente lea references/con_herramientas.md
  para pedir `graph_check` tras cada `boe_articulo`. La tabla de comandos ya empareja cada orden con su herramienta; si
  el margen lo permite, la primera frase que conviene devolver al cuerpo es «con herramientas, `graph_check` con esa
  `norma` y el bloque en `bloques`», unas 90 runas.
- Varios bloques a la vez: la orden de `articulos` solo está en references/varios_bloques.md, y el cuerpo dice cuándo
  leerla; lo deciden las evals de dos bloques, como la 20.
- Preguntas con fecha: el cuerpo lleva el disparador, las tres formas y que solo salen de una lectura con fecha; cuándo
  van la segunda y la tercera, y el año o el mes sin día, están en la referencia. Lo deciden la 22 y la 23.
- La regla 7 queda en el cuerpo sin su frase, que está en references/lecturas_que_fallan.md. Ningún control la exige en
  `SKILL.md`: solo está en `formas_fijas` de la lista de expresiones.
- Prosa de las referencias: el control de expresiones también las lee. Hay que vigilar sobre todo
  lecturas_que_fallan.md y redaccion_modificada.md:
  - «graph check», «kitlegal graph», «hallazgos», «json» y «version-obsoleta», solo como código en línea;
  - nunca «la comprobación no», «la comprobación de la redacción», «consulta anterior», «lectura anterior», «se consultó
    antes», «cambio(s) de/en la redacción», «redacción posterior» (`redaccion_posterior` como código, sí), «vigente
    hasta», «haya cambiado desde» ni «código 3»;
  - ninguna fecha AAAAMMDD con cifras, «siempre» nunca en la misma frase que `--offline`, y ni «eval» ni «job»;
  - la frase de la regla 7, solo en un bloque `text`.
- Cada referencia se enlaza con un enlace de Markdown fuera de los bloques de código. El texto del enlace puede ser el
  nombre corto (`[normas](references/normas.md)`), que ahorra unas 15 runas por enlace frente a repetir la ruta.
- El margen es de 119 runas: una frase que se alargue al escribirla deja el cuerpo por encima de 11 000. El control lo
  dirá, pero conviene medir cada apartado con `wc -m` mientras se escribe.

#### Descripción propuesta

En tercera persona, 878 runas, medidas con `wc -m` sobre el texto en una línea:

```text
Consulta y cita normativa consolidada del Boletín Oficial del Estado (BOE) de cualquier materia —procedimiento administrativo, contratación pública, régimen local, tributos, transparencia, relaciones laborales…—: lee con kitlegal índice y artículos y cita identificador y bloque. Se usa siempre que la respuesta dependa de lo que dice una norma —qué dice un artículo, ley o real decreto, qué plazo, requisito o procedimiento fija, dónde se regula una materia— o de lo que decía en una fecha, la del acto o de los hechos, aunque el agente crea saberlo: sin leer la norma con kitlegal no hay cita. Se activa antes de boe_articulo u otra herramienta boe_…: qué pedir y cómo citar lo dice la skill. La norma, estatal o autonómica, se nombra por número y año (Ley 39/2015, Real Decreto Legislativo 2/2004), abreviatura (LPAC, LCSP, LRBRL, LGT, TRLRHL, LRJSP) o identificador BOE-A-….
```

Cubre B01–B07. No lleva «Úsala», «Actívala», «tú» ni «creas», ni ninguna etiqueta XML. Las normas que nombra están en
`data/normas.yaml`, y no lleva ninguna expresión de la lista ni ninguna instrucción de las evals.
## legal-core

### Estructura

#### El cuerpo

Medida sobre un borrador completo del cuerpo, contado con `LC_ALL=en_US.UTF-8 wc -m` (borrador medido en la sesión del
plan, no versionado). Cada cifra cuenta el apartado con su encabezado y sus líneas en blanco; el cuerpo es el fichero sin
el frontmatter, desde la línea en blanco que sigue al `---`.

| Apartado del cuerpo | Qué lleva | Líneas | Caracteres |
|---|---|---|---|
| `# Territorio, plazos, jerarquía normativa y leyes vertebrales` | título y una frase: identifica, calcula y orienta; no tramita ni sustituye a un profesional | 6 | 211 |
| `## Referencias` | «Cuándo leer cada una:» y una línea con enlace por cada una de las seis referencias | 14 | 897 |
| `## Protocolo` | el reparto de los pasos y los pasos 1 a 11, uno o dos enunciados por paso, con todo lo que las evals comprueban sin modelo | 33 | 2837 |
| `## Cómo se presenta el territorio` | cada dato como lo da `data`, el boletín por `codigo` y `nombre`, y el bloque de las tres líneas de cobertura | 11 | 323 |
| `## Cómo se presenta un plazo` | las nueve formas `⚠` sin cambio (plazo y su ejemplo del contencioso, redacción a fecha y de hoy, sin calcular, no cubierto ×2, cuatro avisos) y la regla de la línea entera | 42 | 1986 |
| `## Comandos` (párrafo) | herramienta o orden, `PATH`, `data.clase` y los códigos 0, 2 y 3 | 8 | 539 |
| `## Comandos` (región generada) | sin cambio | 19 | 1498 |
| `## Reglas` | las nueve reglas, la 7 con la línea sin consulta en su bloque `text`, sin cambio | 28 | 2199 |
| **Cuerpo** | | **161** | **10 490** |

El borrador queda 510 caracteres por debajo del límite de 11 000; tiene menos de 500 líneas
con margen. La descripción del borrador (en «Notas») tiene 888 caracteres.

#### Las referencias a mano

Sin cabecera y no declaradas en `kitlegal-referencias`; medidas sobre su borrador (borrador medido en la sesión del plan,
no versionado).

1. `references/municipios.md` — 64 líneas, 3576 caracteres; sin índice, porque no pasa de 100 líneas. Abre con un
   párrafo que dice lo que lleva. Apartados:
   - `## El municipio que falta` (L19, L20, L22)
   - ``## Lo que trae `data` `` (L23, L24, L25, L92)
   - `## La plantilla del territorio` (la plantilla de ocho líneas de `main`, sin cambio; L33, L34)
   - `## Qué dice cada aspecto de la cobertura` (la tabla de `main`; L29 a L32, L41)
   - `## Cuando no sale con 0` (L35 a L38, L65)

   Línea del cuerpo, en «Referencias»:

   ```text
   - [municipios.md](references/municipios.md): al presentar un territorio, si falta el municipio o no sale con 0, y
     para explicar su cobertura.
   ```

2. `references/calculo_de_plazos.md` — 59 líneas, 3750 caracteres; sin índice. Abre con un párrafo que dice lo que
   lleva. Apartados:
   - `## Lo que la ley regula aparte` (L47 el porqué, L48, L49, L50 con el ejemplo de `daprimera`, L52)
   - `## El acto presunto` (L54)
   - `## El silencio` (L61, L62, L63; la norma del silencio leída en la fecha de inicio, L68)
   - `## La reclamación del art. 24.2 LTAIBG` (L64)
   - `## Los municipios que cuentan` (L57 «pueden ser el mismo», L58, L60 el porqué)
   - `## El plazo sin calcular` (L66 el porqué, L67 «sin nombrar de memoria la norma que regía»)
   - `## La lectura en la fecha de inicio` (el porqué de L68, L72, L73, L74)

   Línea del cuerpo, en «Referencias»:

   ```text
   - [calculo_de_plazos.md](references/calculo_de_plazos.md): en lo que la ley regula aparte, un acto presunto, un
     silencio, una reclamación de acceso, un plazo sin calcular y una lectura fallida o con `redaccion_posterior`.
   ```

Ninguna de las dos lleva una línea `⚠` (nombran la forma del cuerpo con palabras: «la línea de no cubierto del
cuerpo», «su línea de redacción de hoy»), ni enlaza a otra referencia (nombran la de la cobertura por materia con
palabras), ni una orden con un identificador real, ni una norma escrita como «Ley N/AAAA», ni las palabras que veta
`instruccionDeEvals` (`eval`, `job`, `siempre … --offline`…).

#### Las referencias generadas

Sin cambio de contenido (FR-024). Sus líneas en «Referencias»:

```text
- [leyes_vertebrales.md](references/leyes_vertebrales.md): al nombrar una norma, su identificador y su rango.
- [jerarquia_normativa.md](references/jerarquia_normativa.md): para el nivel que regula algo, con el autonómico de la
  `comunidad` de `data`, y la norma que prevalece.
- [cobertura_por_materia.md](references/cobertura_por_materia.md): en toda pregunta de plazos o de recursos.
- [recursos_y_plazos.md](references/recursos_y_plazos.md): para elegir el tipo de plazo y nombrarlo.
```

En el resto del cuerpo, los pasos nombran cada referencia con palabras («la cobertura por materia», «los recursos y
plazos», «la referencia de los municipios»), sin otro enlace: el único de cada una es su línea de «Referencias».

### Notas

#### Ejemplos y textos que cambian

- La descripción: «Úsala cuando se pregunte…» pasa a «Se usa cuando se pregunta…» y va en segunda posición; la
  capacidad de calcular se resume en «calcula el último día del plazo», porque la lista de activación ya nombra el
  recurso, el plazo administrativo, el contencioso y el silencio; «Constitución» pasa a «CE»; «del recurso
  contencioso-administrativo» a «del contencioso-administrativo».
- El ejemplo de las cuotas de la Seguridad Social con `daprimera` fallida sale del paso 7 y va, con su texto, a
  `calculo_de_plazos.md`. La eval 12 de legal-core es ese caso: la regla que la decide («la línea va falle o no una
  lectura», regla 9 y paso 7) queda en el cuerpo.
- La lista de materias del paso 7 («tributos, Seguridad Social, sancionadores…») sale del cuerpo: es la de la
  referencia generada `cobertura_por_materia.md`, que la línea de «Referencias» manda leer en toda pregunta de plazos.
- La plantilla del territorio (cinco líneas sin forma comprobada y las tres de cobertura) pasa entera a
  `municipios.md`; el cuerpo guarda las tres de cobertura, que son las que la eval 02 comprueba sin modelo. Devolverla
  al cuerpo costaría unos 330 caracteres (unos 10 820): cabría en 11 000, con menos margen.
- La orden sin fecha del paso 6 (`kitlegal boe articulo <norma> <bloque>`) desaparece del cuerpo: el paso 6 remite a
  boe-legislacion (regla 2), y el paso 10 conserva `boe_articulo` y la orden con `--fecha <inicio>`. Volver a ponerla
  en la regla 2 cuesta unos 55 caracteres.
- Ningún ejemplo nuevo. Las órdenes del cuerpo llevan marcadores (`<nombre o código INE>`, `<tipo> <inicio>
  <municipios>...`, `<norma> <bloque> --fecha <inicio>`); no hay ninguna `kitlegal boe buscar` ni `analisis`, ni
  ninguna orden de lectura y comprobación (el paso 10 dice «sin `kitlegal graph check`»), así que legal-core sigue sin
  forma para PowerShell, como en `main`. El ejemplo del contencioso (`BOE-A-1998-16718`, bloques `a46` y `a128`) es una
  línea `⚠` y se queda en el cuerpo, sin cambio.

#### Riesgos: lo que tiene que quedarse en el cuerpo

Lo comprueban sin modelo las evals de `evals/legal-core/` (`internal/evals/plazos.go`, `territorio.go`, `citas.go`,
`fechadas.go`) o lo necesita toda respuesta:

- Las nueve formas `⚠` de «Cómo se presenta un plazo», byte a byte: `⚠ PLAZO:` con la cita en la misma línea (05, 06,
  09, 11), `⚠ NO CUBIERTO:` con el identificador (07, 08, 12), `⚠ PLAZO SIN CALCULAR:` con el identificador (10),
  `⚠ REDACCIÓN A FECHA:` con la fecha y la vigencia (05, 06, 09), `⚠ FALTAN FESTIVOS LOCALES:` (06), y las demás.
- Las tres líneas de cobertura `<aspecto>: <valor>` (02: `boletin_autonomico: no-configurado`,
  `boletin_provincial: no-configurado`).
- La línea `⚠ SIN CONSULTA AL BOE: …` exacta y sola en su bloque `text`, en la regla 7 (04 y el control
  `linea-sin-consulta` de `make ci`).
- Paso 7: «antes de preguntar nada», `⚠ NO CUBIERTO:` y «ni `plazos calcular`» (07, 08 y 12 lo prohíben); y «si no
  dice la materia, no la preguntes ni detengas la respuesta» (05, 06, 09 y 11 no dicen la materia: preguntarla deja la
  respuesta sin `⚠ PLAZO:`).
- Paso 8: la sede del juzgado en el contencioso (09 exige `territorio resolver Madrid`), y el interesado y la sede del
  órgano (05 y 06 exigen Leganés y Tordesillas).
- Paso 9: el 3 con «el plazo de tipo» → `⚠ PLAZO SIN CALCULAR:` y ninguna línea del plazo (10).
- Paso 10: leer cada precepto con `--fecha <inicio>` (05, 06 y 09 exigen `boe articulo … --fecha`) y, con 4 o 5,
  mantener `⚠ PLAZO:` (11).
- Paso 11: cada aviso tal cual (06).
- Comandos: herramienta o orden, con el ejemplo `mcp__kitlegal__territorio_resolver` (las evals corren en los dos
  modos).
- Lo que pasa a una referencia sin que ninguna eval lo compruebe, y que un modelo podría no leer: los códigos 2 y 3 de
  un municipio, la clase de boletín que no suple un nivel, el DIR3 que no se compone, el trato propio, el silencio, la
  reclamación de acceso, las citas del acto presunto, la lectura fallida sin la línea de la regla 7 en lo no cubierto,
  y la plantilla del territorio. La regla 1 y la 5 los cubren en lo esencial. Si sobra presupuesto tras la reescritura
  real, lo primero que conviene devolver al cuerpo es «sin la línea de la regla 7» en el paso 7 (unos 40 caracteres).
- El margen frente al límite de 11 000 es de 510 caracteres.
- Los nombres `municipios.md` y `calculo_de_plazos.md` no chocan con los generados (`leyes_vertebrales`,
  `jerarquia_normativa`, `recursos_y_plazos`, `cobertura_por_materia`), y no están en `kitlegal-referencias`, que no
  cambia: según FR-012 son referencias a mano mientras no abran con la cabecera «generado desde».

#### Borrador de la descripción

En tercera persona; 888 caracteres, contados con `wc -m` sobre el texto sin salto final, tal como queda tras plegar el
`>-`:

```text
Punto de partida de las preguntas de derecho público español que dependen de un municipio o de un plazo. Se usa cuando se pregunta hasta cuándo se puede recurrir o qué plazo hay para un recurso, cuándo vence un plazo administrativo o el del contencioso-administrativo o se produce el silencio, qué comunidad, provincia o boletines corresponden a un ayuntamiento, si un municipio es foral o qué norma prevalece. Con kitlegal identifica el territorio de un municipio —código INE, provincia, comunidad, régimen, DIR3 del ayuntamiento y boletines— y calcula el último día del plazo con los festivos de los municipios que cuentan; dice qué asuntos regula aparte la ley (tributos, multas de tráfico, Seguridad Social, extranjería) y razona con la jerarquía normativa y las leyes vertebrales (CE, LPAC, LRJSP, LRBRL, TRLRHL, LCSP, LGT…). Para el texto de un artículo se apoya en boe-legislacion.
```

Sin «Úsala», «tu» ni «creas»; los verbos («identifica», «calcula», «dice», «razona», «se apoya») tienen la skill de
sujeto, y «Se usa cuando se pregunta» es impersonal. No lleva ningún `<`. La de `main` tiene 993 caracteres.
## jurisprudencia

### Estructura

Cuerpo medido, no estimado, sobre un borrador completo, sección por sección, contado con `LC_ALL=en_US.UTF-8 wc -lm`
(borrador medido en la sesión del plan, no versionado). Los títulos de los apartados son los de `main`.

| Apartado del cuerpo | Runas | Líneas | Qué cambia frente a `main` |
|---|---:|---:|---|
| (línea en blanco), `# Sentencias: la consulta exacta, el documento cotejado y ninguna cita de memoria` y la entrada | 783 | 10 | entrada condensada (977 → 783), con J06–J10 |
| `## Protocolo` | 681 | 12 | J15 pasa a «Comandos» (836 → 681) |
| `### 1. Qué sentencias hacen falta, y cuáles están delante` | 810 | 11 | igual |
| `### 2. La sentencia que no está delante: la línea y la consulta preparada` | 1 645 | 29 | la tabla pierde la columna «Orden» y su última fila absorbe J24; J26, J28 y J29 a la referencia; J27 se queda; línea de cuándo leerla (2 364 → 1 645) |
| `### 3. Con el documento delante: cotejar su ficha` | 1 103 | 16 | quedan J31–J33 (regla), J35 y J40; lo demás a la referencia; línea de cuándo leerla (2 956 → 1 103) |
| `### 4. Citar` | 836 | 17 | igual |
| `### 5. Qué se dice de una sentencia` | 578 | 9 | igual |
| `### 6. Una pregunta por materia` | 398 | 7 | sin el bloque de la orden ni J57, que van a la referencia; línea de cuándo leerla (543 → 398) |
| `### 7. Una sentencia del Tribunal Constitucional` | 395 | 6 | igual |
| `## Comandos`, el párrafo a mano | 586 | 8 | absorbe J15 (552 → 586) |
| región generada de la tabla de comandos | 1 303 | 15 | igual byte a byte (1 302 y la línea en blanco que la sigue) |
| `## Reglas` | 1 487 | 17 | igual |
| **Total del cuerpo** | **10 605** | **157** | de 13 637 y 193; por debajo de 11 000 y de 500 líneas, con J27 en el paso 2 |

Referencias a mano (directorio nuevo `skills/jurisprudencia/references/`, plano, sin cabecera «generado desde», sin
enlaces entre ellas, sin ninguna línea `⚠`, sin índice porque ninguna pasa de 100 líneas):

#### `references/consulta_preparada.md`

La consulta que prepara `cita_preparar` (pasos 2 y 6). 38 líneas y 1 739 runas en el borrador. Abre con un párrafo que
dice que la línea, lo que se pide en cada caso y lo que se le dice a la persona son los del paso 2. Apartados:

- `## La orden de cada forma`: la tabla «Cómo se dio / Orden» con las tres órdenes de `main` (l. 65–67) y el porqué de
  J24 (los ROJ son correlativos).
- `## Las casillas`: J26 (`direccion`, `casillas`, el nombre que ve la persona, «Desde» y «Hasta»).
- `## El equivalente`: J28 y J29.
- `## La búsqueda por materia`: la orden de ejemplo `kitlegal cita preparar --texto "cláusula suelo" --json`, que la
  dirección abre la búsqueda ya hecha (J56) y J57.

Cuándo leerla, dos líneas del cuerpo:

- al final del paso 2: «Antes de dar la consulta, lee [references/consulta_preparada.md](references/consulta_preparada.md):
  la orden de cada forma, cómo se dan las casillas y qué hacer con el equivalente que devuelva la operación.»
- al principio del paso 6: «Ante «qué dice la jurisprudencia sobre…», lee «La búsqueda por materia» en
  [references/consulta_preparada.md](references/consulta_preparada.md), pide `cita_preparar` con los términos de la
  pregunta en `texto` y da la dirección que devuelve. **Mientras tanto no cites ninguna sentencia de memoria**, ni
  como ejemplo: si la respuesta necesita una, va por el paso 2.»

#### `references/documento_aportado.md`

El documento que trae la persona (paso 3). 50 líneas y 2 623 runas en el borrador. Abre con un párrafo que recuerda
que el cotejo sigue siendo lo primero. Apartados:

- `## La ficha y nada más`: J34 con su porqué, J36 con el bloque `<<'DOCUMENTO'` de `main` (l. 101–106) y J37.
- `## Lo que devuelve`: J38.
- `## El documento que no es el pedido`: el ejemplo de J33, J39 y J40 nombrada («va por el paso 2, con su línea y su
  consulta»), sin repetir la línea `⚠`.
- `## El ROJ y el ECLI que no se corresponden`: J41 y J42.
- `## Un texto sin ficha reconocible`: J43–J45, con «el código 2 (`argumentos`)» como en `main`.

Cuándo leerla, al final del paso 3: «Con cada documento, antes del cotejo, lee
[references/documento_aportado.md](references/documento_aportado.md): qué se pasa del documento y cómo, y qué se dice
si no es la pedida, si su ROJ y su ECLI no se corresponden o si no lleva ficha.»

### Notas

#### Lo que tiene que quedarse en el cuerpo

Lo comprueban sin modelo `internal/evals/sentencias.go`, `umbrales.go` y las evals de `evals/jurisprudencia/`, o lo
mira el juez:

- J20, la línea `⚠ SENTENCIA NO COMPROBADA:` con su bloque de ejemplo, al principio de línea
  (`ExtraerNoComprobada`; `no_comprobada` en 01, 02, 06 y 07). Es la única `⚠` de la skill; las referencias la
  nombran («su línea», «la línea del paso 2») y no la repiten.
- J46 y J47, la cita `[<ECLI>, ROJ: <ROJ>]` con los datos de la ficha (`formaDeCitaDeSentencia`; `citas` exactas en
  04, 08 y 09; `ninguna_cita` en 01, 02, 03, 05, 07 y 10; `sin_cita_del_roj` en 06).
- J49, ningún ECLI sin origen, y J31–J32, el cotejo primero: son los umbrales sin modelo `cita_sin_documento` y
  `documento_sin_cotejar:<modelo>:<modo>`, con 0 (04, 06, 08 y 09 piden el cotejo).
- J22 (lo que se pide), J23 (número de resolución, `AAAA-MM-DD` y el año del número) y J25: de ellos salen las
  casillas «Nº Resolución» y «Fecha resolución» con 1088/2023 y 04/07/2023 (01, 02, 07), que se buscan literales en la
  respuesta. J27, decisión del plan, también se queda: la pide toda respuesta con una sentencia que no está delante.
- J35 y J40: la eval 06 exige `cita cotejar --roj "STS 1088/2023"` y después `cita preparar --roj "STS 1088/2023"`, con
  la casilla «Nº ROJ».
- J55, J56 y J58: 03 y 10 exigen `cita preparar` con `--texto` no vacío y la dirección de búsqueda devuelta en la
  respuesta.
- J59: la dirección https://hj.tribunalconstitucional.es/ la escribe la skill, no una operación, y 05 la busca literal.
- J51–J54, J30 y J67: los mira el juez (`afirma_lo_no_leido` decide, como mucho 1 de 60; `afirma_que_existe` se
  publica).
- J70: `boe-legislacion` sigue nombrada en el cuerpo, así que el cierre por nombre del conjunto de dependencias (H30,
  FR-014) no cambia.
- J72–J74 se quedan enteras, con https://kitlegal.es/instalar/, aunque solo las pide una sesión sin binario: es una
  regla invariante y el equivalente de la «línea sin consulta» de `boe-legislacion`, que FR-018 deja en el cuerpo.
- Las reglas de CENDOJ (ADR 0036) quedan todas en el cuerpo o en la descripción: J04, J06–J08, J16, J48 y J67. Las
  referencias no añaden ninguna consulta: solo preparan y cotejan sin red.

#### Riesgos y decisiones a revisar

- J34 (solo la ficha) sale del cuerpo: si el modelo no lee la referencia y pasa el documento entero, el cotejo da lo
  mismo («el resto del texto no cambia lo que devuelve»), y solo se alarga.
- J15 pasa de «Protocolo» a «Comandos», fundida en la frase de los códigos («Códigos de salida de la orden, y entre
  paréntesis la `data.clase` del resultado que la herramienta marca como error»): en el cuerpo nuevo ningún paso dice
  ya que una orden termina con un código; lo dice la referencia del documento.
- La entrada se condensa (977 → 783) sin quitar J06–J10; es cambio de forma que leen las evals.
- «Con cada documento, antes del cotejo, lee…» no contradice J31 (leer la referencia no dice nada del documento); si
  la revisión lo ve ambiguo, «Al cotejar un documento, lee…» cuesta lo mismo.
- Leer una referencia añade una lectura en las respuestas de los pasos 2, 3 y 6: alarga las sesiones y hay que mirar
  el tope del trabajo de jurisprudencia en el job.
- El cuerpo queda 395 runas por debajo del límite de 11 000, con J27 en el cuerpo.
- Comprobado en el borrador: ninguna norma nombrada como «Ley N/AAAA», «Ley Orgánica…» o «Real Decreto…» (hoy la skill
  no nombra ninguna); ningún término de `instruccionDeEvals`; ningún enlace en las referencias; nombres en
  snake_case; ninguna referencia pasa de 100 líneas.

#### Borrador de la descripción

842 runas, contadas con `LC_ALL=en_US.UTF-8 wc -m` sobre el texto sin el plegado del YAML. Es el caso 084 de
`evidencias/adr-0046/casos.yaml` («base en tercera persona», etiqueta `correcto`, sin ningún voto en la medida): solo
cambia la persona frente a `main` («Úsala» → «Se usa», «creas conocerla» → «el agente crea conocerla», «Actívala» →
«Se activa»), y la activación conserva cada disparador. El caso 087 («El agente la activa antes de llamar…») también
está medido como correcto.

> Sentencias y jurisprudencia de los tribunales españoles, sin citar ninguna de memoria. Se usa siempre que la
> pregunta nombre, pida, cite o resuma una sentencia o un auto —«STS 1088/2023, de 4 de julio», un ECLI, un ROJ, «la
> sentencia del Supremo sobre…»—, pregunte si una sentencia existe o qué dice, pida jurisprudencia sobre una materia
> («qué dice la jurisprudencia sobre la cláusula suelo») o traiga el texto o el PDF de una resolución, también cuando
> el agente crea conocerla: una sentencia solo se cita con su documento delante. kitlegal no consulta el buscador del
> CENDOJ: prepara la consulta exacta para que la persona la busque con su navegador y coteja el documento que trae.
> Se activa antes de llamar a cita_preparar o a cita_cotejar: qué pedir, qué decir de la sentencia que no está delante
> y cómo citar la que sí lo dice la skill.
