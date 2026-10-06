# Contrato: `boe-legislacion` v0.1.7

Qué cambia en `skills/boe-legislacion/SKILL.md`, por qué y qué lo comprueba (FR-080 a FR-087, FR-110; SC-010). El
texto exacto es el de [skill-boe-legislacion-v0.1.7.diff](./skill-boe-legislacion-v0.1.7.diff), un prototipo que se
aplica sobre v0.1.6 con `git apply` (research V14: se aplica limpio, 19 líneas fuera y 19 dentro).

## 1. La causa, con las respuestas de la validación

Las cifras salen de `evidencias/adr-0037/lectura.md` y de los seis informes versionados (research M1, M5).

**Lo que no ha leído.** De las 67 respuestas con defecto de la lectura, 20 son de la skill de hoy (informes de H7.4,
H21 y H22: 3, 11 y 6). Leídas una a una (clasificación de este plan), tienen tres formas:

| Forma | Respuestas | Ejemplo (informe, sesión) |
|---|---|---|
| a. Dice que no ha leído o no ha podido leer un precepto, y de qué trata | 7 | «No pude leer el art. 60, que regula el régimen del impuesto sobre bienes inmuebles.» (H21, `14-…-01`) |
| b. Nombra un precepto remitido o no cubierto con su materia al lado | 10 | «Remite a otros preceptos de la propia ley, como el art. 7 sobre rentas exentas…» (H22, `15-…-02`) |
| c. Expone como contenido la regla de un precepto que no leyó | 3 | «También cabe la reclamación potestativa del art. 24, que es previa a ese recurso» (H22, `17-…-02`) |

Qué enseña v0.1.6, línea a línea, que lleva a cada una:

| Líneas de v0.1.6 | Qué dice | Qué le falta | Forma |
|---|---|---|---|
| 97-98 y 109-110 | «Sigue las remisiones que hagan falta…», «si… cambia la respuesta, léelo» | Dice cuándo leer una remisión; no dice qué hacer con la que no se lee. El modelo la cambia por su descripción | b |
| 116 | «Si falta algo que no puedes leer…, dilo en la respuesta en lugar de suplirlo» | Manda decirlo, no cómo: el modelo lo dice con su glosa | a |
| 263-269 (regla 2) | «di qué no se pudo consultar y por qué con lo que significa para quien pregunta… y di cuáles no se pudieron consultar» | «Lo que significa para quien pregunta» invita a explicar de qué trataba el bloque que falló (las evals 14, 15 y 16, cuando una lectura queda fuera de lo grabado) | a |
| 129-133 | «Cada afirmación sobre el contenido de una norma lleva su cita, y lo citado sale del texto que devolvió…» | La regla habla de lo que se cita. Una etiqueta entre paréntesis o un inciso no se toma por afirmación sobre el contenido | b, c |
| 215-219 | «La redacción superada no la has leído… sería texto legal sin fuente» | Es el único sitio que enseña «no lo has leído, no lo digas», y solo para la redacción anterior. Ahí funcionó (`redaccion_no_leida` en 0 desde H7.4): el caso general no tiene su frase | a, b, c |

**«El sobre» y los campos.** Medido en los informes (research M5): ninguna de las 51 respuestas de cada informe de
H7.1, H7.2 y H7.3 nombra «el sobre», `fecha_vigencia` ni `norma_modificadora`; lo hacen 21 de 54 en el de H7.4, 46 de
108 en el de H21 y 45 de 108 en el de H22. Entre el informe de H7.3 y el de H7.4 cambian dos cosas: el modelo que
decide (ADR 0031) y dos frases que H7.4 añadió para decir qué lleva la respuesta de la vigencia, las dos que el hito
señala (líneas 141-142 y 218-219 de v0.1.6). El resto del vocabulario ya estaba en v0.1.3 (research V15). La medición
no separa las dos causas: v0.1.7 quita el vocabulario de toda la prosa, y el cierre publica cuánto baja.

## 2. Los cambios

| Id | Dónde (v0.1.6) | Qué cambia | Causa | Requisito |
|---|---|---|---|---|
| C1 | Paso 5, viñeta nueva tras la de la cita | «**De un precepto que no has leído, nada.**»: de un artículo, un apartado o una disposición que nada devolvió en la conversación no se dice qué dice ni de qué trata, ni entre paréntesis ni aunque se crea saber, con su razón (texto legal sin fuente); si se nombra, por su número y nada más; una remisión va como el texto la da, y si importa se lee y se cita; se puede avisar de que una materia se regula en otra parte, sin nombrar precepto ni regla | formas a, b y c | FR-081, FR-082, FR-083 |
| C2 | Regla 2, líneas 267-269 | «di cuáles no se pudieron consultar» gana «por su número y sin decir de qué tratan» | forma a | FR-083 |
| C3 | Paso 5, líneas 139-144 | La vigencia se dice «con palabras de quien lee la norma… sin nombrar lo que devuelve una orden ni sus campos»; fuera «lo que trae el sobre de `kitlegal boe`» y las dos glosas de campo; «ningún sobre trae» pasa a «ninguna lectura trae» | la frase de H7.4 | FR-084, FR-085 |
| C4 | «Redacción modificada», líneas 218-219 | «qué norma le dio esa redacción y desde cuándo rige», sin las dos glosas de campo | la frase de H7.4 | FR-084, FR-085 |
| C5 | Paso 4, líneas 111-114 | «si la lectura trae avisos» en lugar de «si el sobre trae avisos»; «Modificaciones» sin la glosa `norma_modificadora` y en una línea | el vocabulario | FR-085 |
| C6 | «Cómo se cita», línea 171 | «Cada aviso de la lectura» en lugar de «Cada aviso del sobre» | el vocabulario | FR-084 |
| C7 | «Redacción modificada», líneas 197-199 | Las dos fechas de la línea se dicen por su orden («primero la de la redacción superada y después la más reciente, que es la de la que acabas de leer»), sin los dos nombres de campo | el vocabulario | FR-085 |
| C8 | Paso 4, líneas 109-110 | «Remisiones» en una línea: «a otro precepto o a otra norma» | las líneas de C1 | FR-086 (menos de 300) |
| C9 | Paso 5, la viñeta de C1 (reparación del cierre, §8) | La viñeta dice también «cuándo o cómo se aplica» y «ni en un paréntesis o un inciso», y de la remisión, que «va con las palabras del texto leído y sin un «que es…» ni un «que regula…» detrás: eso es contenido de lo remitido»; una línea más, que paga la línea en blanco entre el ejemplo `⚠ NORMA DEROGADA:` y su viñeta en «Cómo se cita» | forma c, medida en el cierre (§8) | FR-081, FR-082 |

**Las líneas.** v0.1.6 tiene 298 y ese es el tope efectivo: `skills-check` falla con 300 o más y el mutante
`dos-inicios` de `internal/app/skills_test.go` añade una (research V13). C1 cuesta cuatro. Salen de C3 (una: la forma
del aviso, que ese párrafo repetía, queda en «Cómo se cita» y en la regla 3, a las que remite; y «Recuerda que los
textos consolidados…», que sigue en la regla 3), de C5 (una), de C7 (una) y de C8 (una). El prototipo tiene 298
(research V14). C9 cuesta una, que paga el blanco de «Cómo se cita» (§8): sigue en 298.

## 3. Lo que se queda (FR-086)

Sin cambios: la `description` y el frontmatter; los pasos 1 a 3, con las dos órdenes de lectura y comprobación en la
misma orden y sus formas para PowerShell; la forma de la cita y su repaso; las tres formas de aviso y su ejemplo; la
línea `⚠ REDACCIÓN MODIFICADA:` con su bloque; «No es un aviso de vigencia»; «La redacción superada no la has leído»,
salvo las dos glosas de C4; las reglas 1 y 3 a 8, con la línea `⚠ SIN CONSULTA AL BOE:`; y la región generada.
`SKILL.md` no nombra evals, el job ni modelos (H5 FR 077).

## 4. El vocabulario que la prosa no usa (FR-085, FR-110)

- **En datos.** `evals/boe-legislacion/expresiones-prohibidas.yaml` gana la clave `salida_de_las_herramientas`, con
  `el sobre`, `fecha_vigencia` y `norma_modificadora`, y su comentario de cabecera dice lo que el fichero es desde H24:
  el vocabulario que la prosa de `SKILL.md` no usa. `schemas/expresiones-prohibidas.yaml.json` la exige: una lista de
  al menos un elemento, cada uno con el patrón `^[^\s*`]+( [^\s*`]+)*$`, que admite el guion bajo de un nombre de
  campo. Las dos cosas, en una tarea `[datos]`.
- **Cómo delimita la comprobación ese uso.** Las tres expresiones se buscan en **toda la prosa** de `SKILL.md`, la de
  `parrafosDeLaProsa`: fuera de los bloques delimitados, que son las órdenes, y de la región generada, que es la
  frontera que da el hito. Para esta familia el código en línea **cuenta**: se le quitan los acentos graves y se deja
  su texto, porque v0.1.6 escribe los dos campos como código. Las demás familias siguen sin mirar el código en línea.
  La comparación es la de siempre (`formaDeExpresion`): por palabras, sin distinguir mayúsculas. «del sobre», «el mismo
  sobre» o «sobre el texto» no son «el sobre»; `fecha_vigencia_reciente` sí lleva `fecha_vigencia`.
- **Qué da.** Con v0.1.6, cinco párrafos con defecto, los de las líneas 111, 113, 139, 194 y 215: entre ellos están
  las dos frases que el hito señala, en los de la 139 y la 215. Con el prototipo, ninguno, y ninguna expresión de las
  otras cuatro familias (research V14).
- **Tests.** `TestProsaDeLaSkill` gana casos con Markdown escrito en el test: los dos párrafos de v0.1.6, tal cual,
  dan su defecto con sus expresiones; un campo en código en línea cuenta y `graph check` en código en línea sigue sin
  contar; en un bloque delimitado y en la región generada no cuenta ninguna; «del sobre» no es «el sobre». La
  subprueba `prosa-de-la-skill` de `TestEvalsDelRepositorio` aplica lo mismo al `SKILL.md` del repositorio.

## 5. Uso, de fuera adentro

| Salida | Quién, cuántas veces | Tamaño | Cuándo deja de darse |
|---|---|---|---|
| La respuesta | La persona que pregunta; una por pregunta | Lo que ocupa la norma leída, con 0 frases sobre preceptos no leídos. El aviso de lo que no cubre es una frase: «El cómputo del plazo se regula en otros artículos de la ley, que no he leído.» son 80 B. Como mucho una por materia que el texto leído deja fuera; no crece con lo consultado en meses de uso | El aviso sale mientras la respuesta no lea esa materia; si la lee, lo sustituye la cita. «No he podido leer el art. 60.» sale solo en la pregunta en que esa lectura falla |
| `SKILL.md` | El modelo; una vez por conversación en que la skill se activa | 298 líneas; las órdenes por pregunta no cambian (una lectura con su comprobación por bloque, H7.3) | No da señales |

## 6. `CHANGELOG.md` (FR-087)

En *Unreleased*, `boe-legislacion` v0.1.7: no dice qué dice ni de qué trata un precepto que no ha leído; traslada las
remisiones como el texto las da; y dice la vigencia sin «el sobre» ni nombres de campos.

## 7. Lo que mide el cambio

En `make ci`: `skills-check` (menos de 300 líneas, frontmatter, región sin drift) y la prosa (§4). En el cierre:
`afirma_lo_no_leido:<modelo>:<modo>`, que decide con 0 en cada modo, y `cuenta_su_proceso:<modelo>:<modo>`, que se
publica ([informe-del-job.md](./informe-del-job.md) §2). Si el cierre marca alguna respuesta, la reparación corrige
`SKILL.md` con la frase marcada delante; no toca la rúbrica, los casos, la medida ni el umbral (FR-036).

## 8. Reparación del cierre: C9, el inciso detrás de una remisión (FR-081, FR-082)

Lo añade `reparar_cierre` tras la medición sobre `a5bda45` (research V23, D27): con C1-C8,
`afirma_lo_no_leido:claude-sonnet-5-5:orden` da 1 de 54 frente al umbral 0 —la sesión
`08-ltaibg-plazo-de-resolucion-claude-sonnet-5-5-01`, con tres votos sobre la frase «También cabe la reclamación
potestativa del artículo 24, que es previa a ese recurso»— y el del modo herramienta, 0 de 54; los demás umbrales se
cumplen y `legal-core` aprueba.

**Causa.** La sesión leyó solo el art. 20 de la Ley 19/2013, cuyo apartado 5 dice «sin perjuicio de la posibilidad de
interposición de la reclamación potestativa prevista en el artículo 24», y nada más del art. 24; que la reclamación
sea previa al recurso es la regla del art. 24, que no leyó. Es la forma c de §1, con la misma frase que marcó la
validación en los informes de H21 y H22 (`evidencias/adr-0037/casos.yaml`, dos casos `defecto`). C1 prohibía decir
«qué dice» o «de qué trata» el precepto no leído y mandaba trasladar la remisión «como el texto la da, sin describir
lo remitido»: el modelo parafrasea el art. 20.5, que sí leyó, nombra el art. 24 solo por su número, como C1 pide, y
cuelga de la remisión un inciso que sitúa lo remitido —cuándo va respecto del recurso— sin tomarlo por una
descripción de lo que dice el art. 24. Otra respuesta del mismo cierre hace la misma construcción con un inciso que no
afirma nada («del artículo 24, que es otra vía», `08-…-herramienta-…-01`, sin ningún voto afirmativo): es la que C1 no
nombraba.

**Qué cambia (C9).** La viñeta de C1 enumera lo que no se dice del precepto no leído («qué dice, de qué trata ni
cuándo o cómo se aplica»), nombra el inciso junto al paréntesis y dice de la remisión que va con las palabras del texto
leído y sin un «que es…» ni un «que regula…» detrás, porque eso es contenido de lo remitido; «si importa, léelo y
cítalo» y el aviso de lo que no cubre quedan como estaban. Una línea más (la viñeta pasa de cuatro a cinco), que paga
la línea en blanco entre el ejemplo `⚠ NORMA DEROGADA:` y la viñeta que lo sigue, en «Cómo se cita», como hizo H7.4
con C11: `SKILL.md` sigue en 298 líneas, con las mismas 14 de más de 120 caracteres, y la prosa sin ninguna expresión
de la lista (`make skills-check`, en verde en la sesión de la reparación). Ninguna forma fija, eval, rúbrica, caso,
medida ni umbral cambia (FR-036). Rechazado: la frase marcada como ejemplo en la skill, porque es contenido legal del
art. 24 y el modelo repite el vocabulario de la prosa (H7.3); y leer siempre el precepto remitido, que añade una
lectura a preguntas que no la necesitan cuando C1 ya manda leerlo si importa.

**Lo que no se ha medido.** El efecto de C9 sobre `afirma_lo_no_leido` y sobre las demás evals no se puede medir en
una sesión de un paso, que no abre sesiones con modelo (ADR 0032): lo mide el job de evals de la medición siguiente,
en los dos modos.
