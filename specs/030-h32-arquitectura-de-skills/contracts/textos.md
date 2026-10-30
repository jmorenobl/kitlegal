# Contrato · Los textos: los de `c991808` y los de la entrega (FR-009, FR-015 a FR-018, FR-021, FR-040)

Los cuerpos enteros de la entrega son los de `borradores/<skill>.SKILL.md`, medidos en esta sesión con
`awk 'f>=2{print} /^---/{f++}' <fichero> | wc -lm` (`docs/SKILLS.md` §6). El hueco que dejan es el de la línea que
`make skills-sync` añade por un paquete, la de la vertical de ensayo, 70 caracteres con su salto (FR-024):

| Skill | `c991808` | Entrega | Con la línea de la vertical de ensayo |
|---|---|---|---|
| `normativa` (de `boe-legislacion`) | 10 995 | 10 858 | 10 928 |
| `jurisprudencia` | 10 859 | 10 842 | 10 912 |
| `plazos-y-territorio` (de `legal-core`) | 10 944 | 10 896 (con la línea de lo administrativo, 94) | 10 966 |
| `consulta-juridica` | — | 3 660 | unos 3 725 (la línea de la lista de paquetes) |

Las descripciones: `normativa`, unos 887 caracteres; `plazos-y-territorio`, unos 880; `consulta-juridica`, unos 705;
`jurisprudencia`, la de hoy.

## 1. `normativa`: lo que cambia del cuerpo de `boe-legislacion`

| # | Dónde | `c991808` | Entrega | Por qué |
|---|---|---|---|---|
| N1 | paso 1 | — | las dos marcas de los paquetes, vacías | FR-023 |
| N2 | paso 2, viñeta 1 | `kitlegal boe indice BOE-A-2015-10565 --json` | `kitlegal boe indice BOE-A-1978-31229 --json` | FR-021 |
| N3 | paso 2, viñeta 2 | las órdenes de `BOE-A-2015-10565 a21`, en Bash y PowerShell | las mismas, de `BOE-A-1978-31229 a140` | FR-021; la caché sirve el bloque (eval 09) |
| N4 | paso 2, viñeta 2, tras «La comprobación, nunca sin argumentos ni antes; si faltó, pídela ya.» | — | «Habla de otras conversaciones, que quien pregunta no conoce: un bloque sin entrada `version-obsoleta` no da nada a la respuesta, ni una línea, ni una frase junto a los avisos o a la fecha de vigencia.» | FR-015: el *gotcha* de `e1bfb77` (líneas 215 a 220) en el paso que pide la comprobación |
| N5 | paso 4, viñeta 2 | «**Nada de otra conversación**, ni para afirmar, confirmar o desmentir: de ellas solo va la línea `⚠ REDACCIÓN MODIFICADA:` de cada bloque que la trae.» | — (lo dicen N4 y «Redacción modificada») | sitio para N4 (FR-009) |
| N6 | paso 4, viñeta 3 | «Cada afirmación sobre el contenido de una norma lleva su cita, de lo que `kitlegal boe` devolvió en esta conversación.» | «Cada afirmación sobre una norma, con su cita de lo que `kitlegal boe` devolvió aquí.» | FR-009, más corto |
| N7 | paso 4, viñeta 5 | «De la redacción leída, di qué norma la dio…» | «De la redacción leída, qué norma la dio…» (el resto igual) | FR-009 |
| N8 | paso 5 | «…ni negando («no puedo decir qué <regla> fija»), ni al ofrecer leerlo o avisar de lo no leído: o el número, o la materia. Lo leído para nombrarlo (paso 3) es un salto: lo que remite no se lee para nombrarlo y, como lo que no se pudo leer, se nombra solo con las palabras del texto leído. Puedes avisar de que una materia se regula en otra parte, sin nombrar precepto ni regla.» y el ejemplo «los artículos que regulan <materia> (N y M)» | «…ni negando («no puedo decir qué <regla> fija»). Una remisión que no has leído se repite con las palabras del texto leído, sin materia: quitarle el número y dejarle la materia sigue diciendo lo que no has leído. Y explicar una palabra del texto leído con lo que regula otro precepto es afirmar ese precepto: es decir qué dice.» | FR-016: los dos cambios, con su porqué; ningún ejemplo nombra una norma |
| N9 | «Cómo se cita» | «…nada tras el id aunque acabe en punto: `art. 21 de la Ley 39/2015 [BOE-A-2015-10565, bloque a21]`. Una cita por bloque, también del remitido; repásalas, y para otra forma, [como_se_cita](…). Cada línea `⚠` va entera, en una línea y con los datos de la lectura: dicha de otro modo, falta.» | «…y nada más: `art. 140 de la Constitución Española [BOE-A-1978-31229, bloque a140]`. Una por bloque, también del remitido; las dudas, en [como_se_cita](…). Cada línea `⚠` va entera y con los datos de la lectura.» | FR-010 (no hay «otra forma»), FR-021 |
| N10 | regla 2 | «Si una lectura falla, di qué bloques faltan y por qué, y no suplas nada de memoria.» | «Si una lectura falla, di qué bloques faltan y por qué. Lo que lo leído no da —un plazo, un órgano, una vía— no está cubierto: dilo, sin suplirlo de memoria ni con el procedimiento común.» | FR-006: invariantes de lo no cubierto y del procedimiento común |
| N11 | regla 4 | «…ni lo simules: lo hace la persona.» | «…ni lo simules.» | FR-009 |
| N12 | regla 5 | «**Ningún caso especial para un territorio.** Las ordenanzas y demás normas locales no están en esta fuente: dilo si se pregunta por un municipio o lo preguntado les corresponde. Si puede variar por normativa autonómica o foral, dilo.» | «**Territorio.** Las normas locales no están en esta fuente; si puede variar por normativa autonómica o foral, dilo.» | FR-009 |
| N13 | regla 6 | «**El texto sale de `kitlegal boe`, nunca de `kitlegal graph`.** Del grafo, solo `check`.» | «**Del grafo, solo `kitlegal graph check`.** El texto sale de `kitlegal boe`.» | FR-009 |
| N14 | descripción | «…relaciones laborales…—» | «…relaciones laborales, seguros…—» | FR-004: cualquier materia |

No cambian (FR-017, FR-018): el paso 3 entero, con su encabezado; el paso 6 entero; la primera viñeta del paso 4; las
formas de «Redacción en una fecha» y «Redacción modificada»; la regla 3 y la 7 con `⚠ SIN CONSULTA AL BOE:`; la tabla
de comandos. La comprobación (FR-015) no pide ningún cambio en los pasos 3 y 6 ni en esa viñeta: la tabla de FR-017
no tiene ninguna fila.

## 2. `jurisprudencia`

| # | Dónde | `c991808` | Entrega | Por qué |
|---|---|---|---|---|
| J1 | introducción | «Esta skill responde cuando hace falta una sentencia o un auto de un tribunal español. kitlegal…» | «kitlegal…» | sitio para J2 (FR-009); lo dice la descripción |
| J2 | protocolo, tras «Ninguna de las dos pide nada a la red…» | — | las dos marcas de los paquetes, vacías | FR-023 |
| J3 | regla 4 | «El texto de un artículo no se cita de memoria: se consulta con `boe-legislacion`.» | «Ningún artículo, plazo, órgano ni vía de memoria: aquí no están cubiertos.» | FR-006: sin remisión a otra skill, con la invariante de lo no cubierto |

## 3. `plazos-y-territorio`: lo que cambia del cuerpo de `legal-core`

| # | Dónde | `c991808` | Entrega | Por qué |
|---|---|---|---|---|
| P1 | título e introducción | «# Territorio, plazos, jerarquía normativa y leyes vertebrales» · «Esta skill identifica el territorio, calcula los plazos y orienta…» | «# Plazos y territorio» · «Esta skill identifica el territorio de un municipio y calcula los plazos que cuenta `kitlegal plazos`…» | FR-007, FR-008 |
| P2 | «Referencias» | las líneas de `leyes_vertebrales.md`, `cobertura_por_materia.md` y `recursos_y_plazos.md` | «La de cada paquete, en toda pregunta de plazos o de recursos y al nombrar una norma:» y las marcas, con la línea de `administrativo` | FR-019, FR-023 |
| P3 | «Referencias», `calculo_de_plazos.md` | «en lo que la ley regula aparte, un acto presunto, un silencio, una reclamación de acceso, un plazo sin calcular y una lectura fallida o con `redaccion_posterior`» | «en los municipios que cuentan, un plazo sin calcular y una lectura fallida o con `redaccion_posterior`» | FR-019: lo del procedimiento común pasa al paquete |
| P4 | paso 5 | «con las leyes vertebrales» | «con las leyes del paquete» | FR-019 |
| P5 | paso 6 | «**El texto de un artículo**, y el identificador que no está en las referencias, con `boe-legislacion` en esta conversación, citado como ella cita (regla 2).» | «**El texto de un artículo**: el id, del índice (`kitlegal boe indice <identificador> --json`), y el texto, de `kitlegal boe articulo <identificador> <id> --json` o sus herramientas, citado como un precepto del plazo.» | FR-006: responde sola, sin remitir |
| P6 | paso 7 | «Si la materia está en la cobertura por materia, `⚠ NO CUBIERTO:`…» | «Si el asunto está en la cobertura de su paquete, la línea `⚠ NO CUBIERTO:` de su fila, entera…» | FR-012 |
| P7 | paso 9 | «el tipo, de los recursos y plazos» | «el tipo, de la referencia del paquete» | FR-019 |
| P8 | paso 10 | «`kitlegal boe articulo <norma> <bloque> --fecha <inicio> --json`» | «`kitlegal boe articulo <identificador> <id> --fecha <inicio> --json`» | FR-010: el marcador del id |
| P9 | «Cómo se presenta un plazo» | la forma `⚠ PLAZO: <AAAA-MM-DD> — <tipo> [<norma>, <precepto>, <identificador>, bloque <bloque>]` y el ejemplo de la LJCA `[art. 46.1 de la LJCA, BOE-A-1998-16718, bloque a46] […]` | una frase de la cita en la forma única y `⚠ PLAZO: <AAAA-MM-DD> — <tipo>: <forma legible> [<identificador>, bloque <id>]`, «una tras otra» | FR-010, FR-021 |
| P10 | redacción a fecha y de hoy | `[<identificador>, bloque <bloque>]` | las líneas de `normativa`, con `<id>` | FR-013 |
| P11 | `⚠ PLAZO SIN CALCULAR:` | «…dicen hoy: <identificador>, bloque <bloque>, apartado <apartado>…» | «…dicen hoy: [<identificador>, bloque <id>], apartado <apartado>…» | FR-010 |
| P12 | `⚠ NO CUBIERTO:` | `(<identificador>)… (<precepto de la LPAC>)` y `la <precepto de la LPAC> (BOE-A-2015-10565)` | las dos formas de [forma-de-cita.md](forma-de-cita.md) §3, y «Lo no cubierto (paso 7) es la línea de su fila en la referencia del paquete» | FR-011, FR-012, FR-021 |
| P13 | regla 1 | «…de `kitlegal`, de las referencias o de `boe-legislacion`…» | «…de `kitlegal` o de las referencias…» | FR-006 |
| P14 | regla 2 | «El identificador sale de las referencias o de `boe-legislacion`; el texto de un artículo, solo de `boe-legislacion`.» | «El identificador sale de las referencias o de `kitlegal boe`; el texto de un artículo, solo de `kitlegal boe articulo`, con su cita.» | FR-006 |
| P15 | regla 7 | «…con la causa en el marcador; quien acaba de instalar suele tenerlo y sin arrancar aún:» | «…con la causa en el marcador:» | FR-009; la línea lo dice |
| P16 | regla 9 | «Si el asunto está en la cobertura por materia, se declara no cubierto…» | «Si el asunto está en la cobertura de un paquete, se declara no cubierto…» y «Un plazo, un órgano o una vía de otra materia no están cubiertos: dilo, sin darlos de memoria.» | FR-006 |
| P17 | descripción | «Punto de partida de las preguntas de derecho público español…» | la de `borradores/plazos-y-territorio.SKILL.md`: qué calcula, qué plazos cuenta hoy (los de la LPAC y la interposición del contencioso) y cuándo se usa | FR-004, FR-008 |

`kitlegal-referencias` pasa de `leyes_vertebrales jerarquia_normativa recursos_y_plazos cobertura_por_materia` a
`jerarquia_normativa`.

## 4. Las referencias que cambian

| Fichero de la entrega | De | Cambio |
|---|---|---|
| `normativa/references/como_se_cita.md` | `boe-legislacion/references/como_se_cita.md` | «Lo que hace cita» dice la forma única sin el patrón literal fuera de corchetes; sale «La forma legible dentro» (era una segunda forma); «Lo que no es cita» lo dice con palabras, sin escribir ninguna cita fuera de la forma (el control la daría por defecto); el ejemplo, el de la Constitución |
| `plazos-y-territorio/references/administrativo.md` | nueva | `borradores/administrativo.md`: la prosa de «Lo que la ley regula aparte», «El acto presunto», «El silencio» y «La reclamación del art. 24.2 LTAIBG» de `calculo_de_plazos.md`, con `boe-legislacion` cambiado por `kitlegal boe articulo`, y las tres regiones generadas |
| `plazos-y-territorio/references/calculo_de_plazos.md` | `legal-core/references/calculo_de_plazos.md` | sin las cuatro secciones que pasan al paquete; su introducción dice los pasos que detalla; «como `boe-legislacion` dice cualquier lectura fallida» pasa a «como se dice cualquier lectura fallida: qué bloque falta y por qué» |
| `plazos-y-territorio/references/{cobertura_por_materia,recursos_y_plazos,leyes_vertebrales}.md` | las de `legal-core` | salen: las retira `make skills-sync` (generadas y ya no declaradas) |

Las demás referencias pasan sin cambios con su carpeta.

## 5. La suite de H31.1 (FR-040)

Se retiran `internal/app/testdata/script/h31-1-h31-1-lee-lo-que-nombra.txtar`, `h31-1-h31-1-responder.txtar` y
`h31-1-h31-1-sitio-en-el-cuerpo.txtar`. Lo que sigue valiendo lo lleva el guion de aceptación `normativa.txtar`
([aceptacion.md](aceptacion.md)).

`h19-skills-invocacion.txtar` pasa a los nombres nuevos: la orden de ejemplo que busca en `normativa` es la de la
Constitución (`kitlegal boe articulo BOE-A-1978-31229 a140`), la tabla de `kitlegal territorio`, la de
`plazos-y-territorio`, y exige `PATH` en el `SKILL.md` de cada una de las cuatro skills: la de entrada lo dice al
remitir las órdenes a la skill de cada capacidad.
