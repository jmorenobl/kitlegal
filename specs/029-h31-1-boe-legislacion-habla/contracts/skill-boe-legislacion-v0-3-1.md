# Contrato: `boe-legislacion` v0.3.1 (FR-001 a FR-010)

`skills/boe-legislacion/` de v0.3.1 es la skill medida, `evidencias/sondeos-20261026/brazo-W/skill/`, con el cuerpo de
[`skill-md-v0.3.1.md`](skill-md-v0.3.1.md): ese fichero es el `SKILL.md` de la entrega, byte a byte, frontmatter
incluido. Su cuerpo —desde la línea que sigue al `---` que cierra el frontmatter, la 16— tiene **10 996 caracteres**
(`tail -n +16 … | wc -m`; el gate los cuenta como runas, `utf8.RuneCount` en `internal/skills/sincronia.go:442`) y 156
líneas, frente a 169: 1 085 caracteres menos que la skill medida (12 081) y 4 más que el de `main` (10 992). Con él y las referencias de la
skill medida, `make skills-check` y los controles de contenido salen en verde (research V3).

Las líneas que cita este contrato son las del `SKILL.md` de la skill medida.

## 1. Lo que no cambia

| Qué | Dónde, en la skill medida | Criterio | Lo comprueba |
|---|---|---|---|
| Frontmatter, con la descripción (FR-010) | líneas 1 a 15, iguales a las de `main` | idéntico | `h31-1-sitio-en-el-cuerpo` |
| Paso 3, «Completar» (FR-002) | 48 a 59 | idéntico línea a línea | `h31-1-lee-lo-que-nombra` |
| Primera viñeta del paso 4 (FR-002) | 63 a 68 | idéntica línea a línea | `h31-1-responder` |
| Paso 5, «De un precepto no leído, nada» (FR-002) | 78 a 86 | idéntico línea a línea | `h31-1-lee-lo-que-nombra` |
| Paso 6, «Repasar antes de enviar» (FR-002) | 88 a 100 | idéntico línea a línea | `h31-1-responder` |
| Formas fijas (FR-003): las tres de «Redacción en una fecha», la de «Redacción modificada», la de la regla 7 | bloques `text` de 113-117, 124-126 y 182-184 | cada línea, sin blancos de los extremos, es una línea del `SKILL.md` | `h31-1-sitio-en-el-cuerpo`; `formas-de-la-redaccion-fechada`, `hallazgos-de-la-skill`, `linea-sin-consulta` |
| Etiquetas de aviso de la regla 3 (FR-003) | 166-167 | las tres, en el `SKILL.md` | `avisos-de-la-skill` |
| Tabla de comandos (FR-004) | 134 a 159, entre sus marcas | idéntica | `h31-1-sitio-en-el-cuerpo`; `make skills-check` |
| Encabezados `##` y `###` (FR-006) | 14, en su orden | idénticos | `h31-1-sitio-en-el-cuerpo` |
| Número y encabezado en negrita de las siete reglas (FR-006) | 163 a 179 | idénticos | `h31-1-sitio-en-el-cuerpo` |
| Paso 1; los ejemplos de orden del paso 2 y su forma para PowerShell; «Si una orden no termina con `0`…»; «Redacción en una fecha»; «Redacción modificada»; reglas 1 y 7 | 25-28, 37-46, 109-128, 163, 179-184 | idénticos | revisión final |
| Las ocho referencias de `references/` (FR-001, FR-007) | `brazo-W/skill/references/` | cada fichero igual; ninguno más | `h31-1-sitio-en-el-cuerpo` |

Las referencias son las de la skill medida: frente a `main` solo difieren `buscar_la_norma.md` y
`redaccion_modificada.md`, lo que cambió la #200 (spec, «Contexto»). Ninguna recibe texto del cuerpo (§3), así que su
diff con la skill medida es vacío.

## 2. Los textos reescritos (FR-006)

Fuera de los pasajes medidos, las formas fijas y la tabla. Cada fila da el texto de la skill medida, el de la entrega
y por qué lo quitado no hace falta: o el agente no lo haría mal sin ello, o el cuerpo ya lo dice en otro sitio, que la
fila nombra. Lo demás del cuerpo no cambia (§1).

| # | Skill medida | Entrega | Lo quitado y por qué no hace falta |
|---|---|---|---|
| T1 | Línea 17, el título: `# Consultar y citar legislación consolidada del BOE` | — | No pide nada al agente: el nombre y la descripción de la skill ya la identifican, y no es un encabezado `##` ni `###` al que remita nada (FR-006). Ningún gate lo exige (`grep` en `internal/skills`, research V3) |
| T2 | «Protocolo»: «Sigue los pasos en orden, cada operación con su orden, `kitlegal <applet> <verbo> … --json`, o, si la tienes, siempre con su herramienta, tras leer [con_herramientas](references/con_herramientas.md). Sin su `fuente`, `url`, `fecha_consulta` y `hash`, no hay cita.» | «Sigue los pasos en orden, cada operación con `kitlegal … --json` o, si la tienes, siempre con su herramienta, tras leer [con_herramientas](references/con_herramientas.md). Sin `fuente`, `url`, `fecha_consulta` y `hash`, no hay cita.» | «su orden» y `<applet> <verbo>`: la forma de cada orden la dan la tabla de «Comandos» y los ejemplos del paso 2. El enlace a `con_herramientas` y su condición siguen como están |
| T3 | Paso 2: «Copia el id de la entrada cuyo `titulo` es el artículo, nunca lo compongas con su número, y no pidas ninguno que no salga del índice o de la pregunta.» | «Copia el id de la entrada cuyo `titulo` es el artículo; nunca lo compongas con su número ni pidas uno que no salga del índice o de la pregunta.» | Nada de lo que dice: las dos prohibiciones siguen, con menos palabras |
| T4 | Paso 2: «lee cada bloque una vez, con la comprobación de esa norma y ese bloque detrás, en la misma orden:» | «lee cada bloque una vez, con su comprobación detrás, en la misma orden:» | «de esa norma y ese bloque»: lo enseña la orden de ejemplo de debajo y lo dice «La comprobación, nunca sin argumentos» |
| T5 | Paso 4: «**Nada de otra conversación**, ni para afirmar, confirmar o desmentir: de ellas solo va la línea `⚠ REDACCIÓN MODIFICADA:` de cada bloque que la trae.» | «**Nada de otra conversación**, salvo la línea `⚠ REDACCIÓN MODIFICADA:`.» | «ni para afirmar, confirmar o desmentir»: lo dicen los pasos 5 y 6, medidos —de lo que nada ha devuelto «en esta conversación» no se dice nada, ni en un inciso, y cada cosa dicha se busca «en esta conversación»—. «de cada bloque que la trae»: «Redacción modificada», «Por cada entrada `version-obsoleta`» |
| T6 | Paso 4: «Cada afirmación sobre el contenido de una norma lleva su cita, de lo que `kitlegal boe` devolvió en esta conversación.» | Sale del paso 4 y abre «Cómo se cita» (T9): «Cada afirmación sobre una norma lleva su cita: …» | «de lo que `kitlegal boe` devolvió en esta conversación»: lo dicen el paso 5 («que nada te ha devuelto en esta conversación») y el 6.1 («busca en esta conversación la orden `kitlegal boe articulo` o `articulos`…»). La regla invariante conserva su frase (§4) |
| T7 | Paso 4: «**Ley y reglamento**: con rangos distintos, di el `rango` de cada una, de la tabla o de la búsqueda, y que la ley prevalece sobre su reglamento.» | «**Ley y reglamento**: di el `rango` de cada una y que la ley prevalece sobre su reglamento.» | «con rangos distintos»: una ley y su reglamento siempre lo tienen. «de la tabla o de la búsqueda»: de dónde sale el rango, que traen la tabla de normas, la búsqueda y los metadatos (tabla de «Comandos») |
| T8 | Paso 4: «De la redacción leída, di qué norma la dio y desde cuándo rige, nunca hasta cuándo: ninguna lectura lo trae, ni el aviso de una derogada; `kitlegal graph check` no dice nada de la vigencia.» | «De la redacción leída, di qué norma la dio y desde cuándo rige, nunca hasta cuándo: ninguna lectura lo trae, ni el aviso de una derogada.» | Lo de `kitlegal graph check`: «De `kitlegal graph` solo se usa `check`, y nada más suyo va en la respuesta» («Redacción modificada») |
| T9 | «Cómo se cita»: «La forma legible de la norma y del bloque y, en la misma línea, entre corchetes, el identificador y el id como los da `kitlegal boe`, nada tras el id, aunque acabe en punto: `art. 21 de la Ley 39/2015 [BOE-A-2015-10565, bloque a21]`. Una cita por bloque, también del remitido; repásalas, y para otra forma, [como_se_cita](references/como_se_cita.md). Cada línea `⚠` va entera, en una sola línea y con los datos de la lectura: dicha de otro modo, falta.» | «Cada afirmación sobre una norma lleva su cita: la forma legible y, en la misma línea, entre corchetes, el identificador y el id como los da `kitlegal boe`, nada tras el id aunque acabe en punto: `art. 21 de la Ley 39/2015 [BOE-A-2015-10565, bloque a21]`. Una cita por bloque; para otra forma, [como_se_cita](references/como_se_cita.md). Cada línea `⚠` va entera, en una línea y con los datos de la lectura: dicha de otro modo, falta.» | Gana la frase de T6. «de la norma y del bloque»: lo enseña el ejemplo. «también del remitido; repásalas»: el paso 6.1 repasa «también… aquello a lo que remite el texto leído» y deja «lo dicho con su cita». «sola». El enlace a `como_se_cita` y su condición siguen como están |
| T10 | «Comandos»: «`kitlegal` se invoca desde el `PATH`.» | — | La regla 7 dice qué hacer si el shell no lo encuentra («no lo busques en otras carpetas ni en un plugin») y la tabla nombra cada orden como `kitlegal <applet> <verbo>` |
| T11 | Regla 2: «Si `kitlegal boe` falla o no está, di qué bloques no pudiste consultar y por qué, y no suplas nada de memoria; sin herramienta ni binario, la regla 7.» | «Si una lectura falla, di qué bloques no pudiste consultar y por qué, y no suplas nada de memoria; sin herramienta ni binario, la regla 7.» | «o no está»: es lo de la regla 7, a la que la misma frase remite. `lecturas_que_fallan.md` (líneas 18 y 50) y `redaccion_en_una_fecha.md` (46) remiten a la regla 2 para una lectura que falla, y eso sigue |
| T12 | Regla 3: «Cada aviso va con su forma fija, `⚠`, su etiqueta y dos puntos, y su frase o una explicación: `⚠ NORMA DEROGADA:`, `⚠ VIGENCIA AGOTADA:` o `⚠ TEXTO POSIBLEMENTE DESACTUALIZADO:`. Trasládalos, también una fecha de vigencia del bloque que no encaja con lo preguntado; una derogada no es vigente. Los textos consolidados del BOE son informativos y no sustituyen el asesoramiento de un profesional.» | «Cada aviso, con su frase o una explicación: `⚠ NORMA DEROGADA:`, `⚠ VIGENCIA AGOTADA:` o `⚠ TEXTO POSIBLEMENTE DESACTUALIZADO:`, y la fecha de vigencia del bloque que no encaja con lo preguntado; una derogada no es vigente. El texto consolidado es informativo y no sustituye el asesoramiento de un profesional.» | «va con su forma fija, `⚠`, su etiqueta y dos puntos»: lo enseñan las tres formas que siguen y lo dice «Cómo se cita» (cada línea `⚠` entera; «dicha de otro modo, falta»). «Trasládalos»: es el encabezado. La última frase dice lo mismo en singular. Las tres etiquetas siguen; `como_se_cita.md` (42) y `redaccion_en_una_fecha.md` (29) remiten a ellas en la regla 3 |
| T13 | Regla 4: «La skill consulta y cita: no presentes, notifiques, firmes ni tramites nada por nadie, ni lo simules; si se pide, di que lo hace la persona y cita, si procede, la norma aplicable.» | «No presentes, notifiques, firmes ni tramites nada, ni lo simules: lo hace la persona.» | «La skill consulta y cita»: lo dice el protocolo entero. «por nadie»: el encabezado. «si se pide, di que» y «cita, si procede, la norma aplicable»: que lo hace la persona sigue, ahora como el porqué, y citar la norma de lo que se pregunta es lo que hace toda respuesta (pasos 2 a 6) |
| T14 | Regla 5: «Nada propio de un municipio o una comunidad concretos. Si se pregunta por un municipio, responde con la normativa estatal o autonómica consolidada en el BOE y di que las ordenanzas y demás normas locales no están en esta fuente; dilo también siempre que lo preguntado les corresponda. Si puede variar por normativa autonómica (competencias compartidas o cedidas, desarrollo autonómico, régimen foral), dilo.» | «Las ordenanzas y demás normas locales no están en esta fuente: dilo si se pregunta por un municipio o lo preguntado les corresponde. Si puede variar por normativa autonómica o foral, dilo.» | «Nada propio de un municipio o una comunidad concretos»: el encabezado. «responde con la normativa estatal o autonómica consolidada en el BOE»: es lo único que la skill lee (paso 1 y tabla de normas). El paréntesis queda en «o foral»: las competencias compartidas o cedidas y el desarrollo autonómico son normativa autonómica |
| T15 | Regla 6: «Lo citado sale de `kitlegal boe articulo` o `articulos`; lo de `kitlegal graph` no se cita, parafrasea ni reconstruye: dice qué volver a comprobar, no qué dice el artículo.» | — (queda el encabezado: «6. **El texto sale de `kitlegal boe`, nunca de `kitlegal graph`.**») | Lo citado sale de las órdenes que el paso 6.1 busca, `kitlegal boe articulo` o `articulos` (o sus herramientas); y de `kitlegal graph` «solo se usa `check`, y nada más suyo va en la respuesta» («Redacción modificada») |

La regla 7 no se reescribe: la eval 21, sin binario ni servidor, decide, y la regla lleva su porqué («no has consultado
nada») y sus prohibiciones concretas («de memoria o con salvedades»), lo que el hito pide conservar (research D2).

## 3. Los textos de una clase de pregunta (FR-007): ninguno sale

Ningún texto sale del cuerpo a una referencia, y ninguna referencia nueva existe. Cada texto de una clase de pregunta
de la skill medida, y por qué se queda:

| Texto de la skill medida | Clase | Dónde | Por qué se queda |
|---|---|---|---|
| «tras leer [con_herramientas](references/con_herramientas.md)» | herramientas en vez de órdenes | «Protocolo» | Fuera de los pasos numerados (FR-007); su enlace sigue como está |
| «si no está, búscala por su título con [buscar_la_norma](…)» | norma fuera de la tabla | paso 1 | Ya es el enlace con su condición; el detalle ya está en la referencia |
| «Sin el id del bloque, lee el índice: … no pidas ninguno que no salga del índice o de la pregunta.» | pregunta sin id de bloque | paso 2 | Decide la orden que se ejecuta y el id que se pide. Elegir del índice no le pide al agente ningún dato que no tenga, así que una referencia con esta guía no se abriría (`evidencias/sondeos-20261025/LEEME.md`, líneas 57 a 59) y el agente compondría el id; el paso 3, medido, cuenta con ella («viste su id solo en un índice») |
| «—con ella, [redaccion_en_una_fecha](…)—» | fecha | paso 2 | Ya es el enlace con su condición |
| «En PowerShell 5.1: …» y su orden | PowerShell | paso 2 | `ordenes-para-powershell` exige la forma para PowerShell en el mismo fichero que la orden de lectura (`docs/SKILLS.md` §1): fuera, el control sale en rojo (FR-008) |
| «Varios bloques a la vez: [varios_bloques](…). Si una orden no termina con `0`, antes de seguir: [lecturas_que_fallan](…).» | varios bloques; lectura que falla | paso 2 | Ya son los enlaces con su condición |
| «**Nada de otra conversación**…» | otra conversación; `version-obsoleta` | paso 4 | Reescrita (T5) en una línea; llevarla a `redaccion_modificada` pediría en el paso 4 un enlace con su condición casi igual de largo, y lo de otra conversación vale para toda respuesta |
| «**Ley y reglamento**…» | ley y reglamento de rango distinto | paso 4 | Regla invariante de `CLAUDE.md`: su frase va en el cuerpo (FR-006, §4); reescrita (T7) |
| «ni el aviso de una derogada» | norma derogada | paso 4 | Cuatro palabras dentro de una frase que vale para toda lectura |
| «Redacción en una fecha» y «Redacción modificada» | fecha; `version-obsoleta` o qué cambió | fuera de los pasos | Fuera de los pasos numerados (FR-007), con formas fijas (FR-003); sus enlaces siguen como están |
| Reglas 1, 3, 4, 5 y 7 | búsqueda vacía; avisos; se pide actuar; municipio; sin binario | «Reglas» | Fuera de los pasos numerados (FR-007) y reglas invariantes (FR-006, §4) |

## 4. Las reglas invariantes de `CLAUDE.md` en el cuerpo (FR-006)

| Regla invariante | Su frase en el cuerpo de v0.3.1 |
|---|---|
| No inventar contenido legal | Regla 2: «**Nunca inventar contenido legal.** Si una lectura falla, …, y no suplas nada de memoria» |
| Cada afirmación con su cita | «Cómo se cita»: «Cada afirmación sobre una norma lleva su cita»; y «Sin `fuente`, `url`, `fecha_consulta` y `hash`, no hay cita» («Protocolo») |
| Distinguir ley y reglamento | Paso 4: «**Ley y reglamento**: di el `rango` de cada una y que la ley prevalece sobre su reglamento.» |
| Señalar la variación autonómica | Regla 5: «Si puede variar por normativa autonómica o foral, dilo.» |
| No concluir que algo no existe | Regla 1: «**No concluir que algo no existe.** Una búsqueda vacía, o una norma fuera de la tabla, no prueba que no exista.» |
| No actuar en nombre de nadie | Regla 4: «**Nunca actuar en nombre de nadie.** No presentes, notifiques, firmes ni tramites nada, ni lo simules: lo hace la persona.» |
| El texto sale de `kitlegal boe` y no del grafo | Regla 6: «**El texto sale de `kitlegal boe`, nunca de `kitlegal graph`.**» |
| Sin herramienta ni binario, la respuesta lo dice | Regla 7, sin cambios, con su forma fija |

## 5. Las remisiones

Las referencias remiten al cuerpo por número y título, y cada destino conserva su texto y su sentido: «regla 1»
(`buscar_la_norma.md:24`), «regla 2» (`lecturas_que_fallan.md:18` y `:50`, `redaccion_en_una_fecha.md:46`), «regla 3»
(`como_se_cita.md:42`, `redaccion_en_una_fecha.md:29`), «paso 2» (`redaccion_en_una_fecha.md:40`), «Cómo se cita»
(`como_se_cita.md:3`, `redaccion_en_una_fecha.md:29`, `redaccion_modificada.md:16`), «Redacción en una fecha»
(`redaccion_en_una_fecha.md:4`), «Redacción modificada» (`redaccion_modificada.md:4`) y «Comandos», por la columna
«Herramienta» (`con_herramientas.md:8`). El cuerpo remite a su regla 7 desde la 2. Líneas de las referencias de la
skill medida, comprobadas con `grep -n` en esta sesión.

## 6. Una consecuencia fuera de la skill: `CONTRIBUTING.md`

`CONTRIBUTING.md:741-743` dice que las formas de `formas_fijas` de `boe-legislacion` son «la línea `⚠ REDACCIÓN
MODIFICADA:` y la frase de la regla 7 sin su punto final». En la skill medida, la regla 7 es «Sin herramienta y sin
binario» y la frase («No se ha podido comprobar si la redacción ha cambiado desde una consulta anterior») está en
`references/lecturas_que_fallan.md`, en `main` y en la skill medida. La entrega cambia «la frase de la regla 7» por
«la frase de `references/lecturas_que_fallan.md`» y nada más del fichero (research D10).
