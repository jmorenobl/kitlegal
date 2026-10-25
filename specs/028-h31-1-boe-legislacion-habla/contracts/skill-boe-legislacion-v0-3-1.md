# Contrato: `boe-legislacion` v0.3.1, el cuerpo y sus referencias a mano

FR-001 a FR-008. El texto de §1 y de §4 es el que se escribe, byte a byte salvo el salto de línea final de cada
fichero; se validó en una copia de `HEAD` (`3dec6e3`) bajo `$TMPDIR` con los tests de `make skills-check` y
`TestSkillsEmpotradas` en verde (research V6). Lo que §1 no muestra del cuerpo no cambia.

## 1. El cuerpo de `skills/boe-legislacion/SKILL.md`

El frontmatter (líneas 1 a 15) no cambia, y con él la `description` (FR-008). La región de la tabla de comandos, entre
sus marcas, tampoco (FR-006), ni la línea «`kitlegal` se invoca desde el `PATH`.» que la precede, ni las reglas 3 a 6.

### 1.1 Del título a «Redacción modificada»

Sustituye las líneas 17 a 107 de hoy (del título a la frase que acaba en «[redaccion_modificada](…).»):

````markdown
# Consultar y citar legislación consolidada del BOE

## Protocolo

Sigue los pasos en orden, cada operación con su orden, `kitlegal <applet> <verbo> … --json`, o, si la tienes, siempre
con su herramienta, tras leer [con_herramientas](references/con_herramientas.md). Sin su `fuente`, `url`,
`fecha_consulta` y `hash`, no hay cita.

### 1. Identificar la norma

Toma el `BOE-A-…` de la tabla de [normas](references/normas.md), no exhaustiva: si no está, búscala por su título
con [buscar_la_norma](references/buscar_la_norma.md).

### 2. Leer

- Sin el id del bloque, lee el índice: `kitlegal boe indice BOE-A-2015-10565 --json`. Copia el id de la entrada cuyo
  `titulo` es el artículo, nunca lo compongas con su número, y no pidas ninguno que no salga del índice o de la
  pregunta.
- Sin fecha —con ella, [redaccion_en_una_fecha](references/redaccion_en_una_fecha.md)—, lee cada bloque una vez, con
  la comprobación de esa norma y ese bloque detrás, en la misma orden:
  ```bash
  kitlegal boe articulo BOE-A-2015-10565 a21 --json && kitlegal graph check BOE-A-2015-10565 a21 --json
  ```
  En PowerShell 5.1:
  ```powershell
  kitlegal boe articulo BOE-A-2015-10565 a21 --json; if ($LASTEXITCODE -eq 0) { kitlegal graph check BOE-A-2015-10565 a21 --json }
  ```
  La comprobación, nunca sin argumentos ni antes; si faltó, pídela ya.
- Varios bloques a la vez: [varios_bloques](references/varios_bloques.md). Si una orden no termina con `0`, antes de
  seguir: [lecturas_que_fallan](references/lecturas_que_fallan.md).

### 3. Completar

Lee lo remitido que cambia la respuesta, de esta norma o de otra, resuelta antes con los pasos 1 y 2, y, si importan
la vigencia o una modificación, los metadatos y el análisis. Nombra un precepto solo si `kitlegal boe articulo` o
`articulos` ha devuelto su texto en esta conversación, salvo lo del paso 5: si lo remite un bloque leído en el paso 2,
o su id solo salió en un índice, léelo antes con los pasos 1 y 2, o no lo nombres; leído, di de qué trata, con su cita.

### 4. Responder

- **La respuesta empieza por lo que se pregunta y habla de la norma, no de la consulta.** Es todo lo que escribes
  tras la última orden o llamada, desde su primera palabra, y quien pregunta la lee entera: no ve las órdenes, las
  llamadas ni lo que devuelven, y lo que decides al completar la lectura no va en ella. La hacen la norma, su texto, su
  cita, sus avisos y las líneas `⚠` de cada bloque: escribe como quien tiene el BOE delante,
  «el artículo no lleva avisos de vigencia» y no «la lectura no trajo avisos», la norma por su nombre e identificador,
  sin decir de dónde sale, qué has hecho o vas a hacer ni qué ha devuelto o no una orden: no le sirve.
- **Nada de otra conversación**, ni para afirmar, confirmar o desmentir: de ellas solo va la línea
  `⚠ REDACCIÓN MODIFICADA:` de cada bloque que la trae.
- Cada afirmación sobre el contenido de una norma lleva su cita, de lo que `kitlegal boe` devolvió en esta
  conversación.
- **Ley y reglamento**: con rangos distintos, di el `rango` de cada una, de la tabla o de la búsqueda, y que la ley
  prevalece sobre su reglamento.
- De la redacción leída, di qué norma la dio y desde cuándo rige, nunca hasta cuándo: ninguna lectura lo trae, ni el
  aviso de una derogada; `kitlegal graph check` no dice nada de la vigencia.

### 5. De un precepto no leído, nada

De un artículo, apartado o disposición que nada te ha devuelto en esta conversación no digas qué dice, de qué trata ni
cuándo o cómo se aplica, ni en un inciso ni aunque creas saberlo: sería texto legal sin fuente. Número y materia no
van juntos en ninguna frase, ni afirmando («el art. N, que es…», «el art. N sobre <materia>»,
«los artículos que regulan <materia> (N y M)»), ni negando («no puedo decir qué <regla> fija»), ni al ofrecer leerlo o
avisar de lo no leído: o el número, o la materia. Lo leído para nombrarlo (paso 3) es un salto: lo que remite no se
lee para nombrarlo y, como lo que no se pudo leer, se nombra solo con las palabras del texto leído. Puedes avisar de
que una materia se regula en otra parte, sin nombrar precepto ni regla.

## Cómo se cita

La forma legible de la norma y del bloque y, en la misma línea, entre corchetes, el identificador y el id como los da
`kitlegal boe`, nada tras el id, aunque acabe en punto: `art. 21 de la Ley 39/2015 [BOE-A-2015-10565, bloque a21]`.
Una cita por bloque, también del remitido; repásalas, y para otra forma, [como_se_cita](references/como_se_cita.md).
Cada línea `⚠` va entera, en una sola línea y con los datos de la lectura: dicha de otro modo, falta.

## Redacción en una fecha

Con una fecha en la pregunta —un día, un mes, un año o un hecho fechado—:

```text
⚠ REDACCIÓN A FECHA: <forma legible> [<identificador>, bloque <id>]: se cita la redacción vigente el AAAA-MM-DD, la de fecha de vigencia AAAAMMDD.
⚠ REDACCIÓN DE HOY: <forma legible> [<identificador>, bloque <id>]: hoy la redacción es otra, la de <identificador>, con fecha de vigencia AAAAMMDD.
⚠ SIN REDACCIÓN VIGENTE: <forma legible> [<identificador>, bloque <id>]: el AAAA-MM-DD no tenía redacción vigente; la tiene desde el AAAAMMDD, la fecha de vigencia de su primera redacción.
```

## Redacción modificada

Por cada entrada `version-obsoleta` de `kitlegal graph check`, y si se pregunta qué cambió, lee
[redaccion_modificada](references/redaccion_modificada.md):

```text
⚠ REDACCIÓN MODIFICADA: <forma legible> [<identificador>, bloque <id>]: la redacción con fecha de vigencia AAAAMMDD, la que se consultó antes, ha sido sustituida por la de AAAAMMDD, que es la que se cita.
```

De `kitlegal graph` solo se usa `check`, y nada más suyo va en la respuesta.
````

### 1.2 Las reglas

Las reglas 1 y 2 cambian, la 7 de hoy sale (§3, fila 1) y la 8 pasa a ser la 7 con el mismo texto; la 3 a la 6, igual:

````markdown
1. **No concluir que algo no existe.** Una búsqueda vacía, o una norma fuera de la tabla, no prueba que no exista.
2. **Nunca inventar contenido legal.** Si `kitlegal boe` falla o no está, di qué bloques no pudiste consultar y por
   qué, y no suplas nada de memoria; sin herramienta ni binario, la regla 7.
````

````markdown
7. **Sin herramienta y sin binario, la respuesta lo dice.** Sin la herramienta, si `kitlegal` no está —el shell no lo
   encuentra o no ejecutas órdenes—, no lo busques en otras carpetas ni en un plugin: no has consultado nada.
   No afirmes nada de la norma, de memoria o con salvedades, ni cites; lleva esta línea, con la causa en el marcador:
   ```text
   ⚠ SIN CONSULTA AL BOE: <causa>. Si ya tienes kitlegal, cierra Claude del todo, ábrelo y pregunta en un chat nuevo; si no: https://kitlegal.es/instalar/
   ```
````

### 1.3 Presupuesto

Con §1.1 y §1.2, el cuerpo (desde el byte que sigue a la línea `---` que cierra el frontmatter) tiene **10 993
caracteres**, uno más que hoy (10 992) y 7 por debajo del máximo (FR-005; research V6). Lo mide el gate de H31 con
`utf8.RuneCount` (`internal/skills/sincronia.go`, `maximoDeCaracteresDelCuerpo`): en la copia del plan, el mismo
cuerpo con 8 caracteres más dio `cuerpo de 11001 caracteres (máximo 11000)` y `TestSkillsDelRepositorio` en rojo
(research V6). Si la revisión cambia una frase, el corrector
deja el cuerpo en 11 000 o menos con la regla de FR-006, nunca sacando nada de §2.

Ningún ejemplo de §2 queda partido entre dos líneas: cada uno se encuentra con `grep -F` (quickstart, escenario 2).
Las líneas que la rama añade al cuerpo no nombran `BOE-A-2004-4214`, `BOE-A-2006-20764`, `BOE-A-2013-12887`,
«Ley 35/2006», «Real Decreto Legislativo 2/2004», «Ley 19/2013», «IRPF» ni «TRLRHL» (FR-004; quickstart, escenario 3);
el único ejemplo de norma del cuerpo sigue siendo la Ley 39/2015, que ninguna eval por materia pide.

## 2. Las reglas de `27aeccc` y su frase en v0.3.1 (FR-001, SC-003)

Una fila por regla de los dos pasajes del cuerpo de antes de H31 (`skills/boe-legislacion/SKILL.md` del árbol
`27aeccc`, `d14bb86^`): la última viñeta de su paso 4 (línea 98) y su paso 5 entero (líneas 100 a 130). El id es el de
la «Tabla de reglas» de `specs/027-h31-las-skills-siguen/research.md`, que da también su sitio en v0.3. «Cómo vuelve»:
*igual*, el cuerpo de hoy ya la dice con el mismo contenido y se queda como está; *vuelve*, H31 la comprimió o le quitó
sus ejemplos y vuelve con ellos; *FR-002/FR-003*, es de lo remitido o del precepto no leído y vuelve como la dejan
esos requisitos.

| Id | Regla en `27aeccc` | Cómo vuelve | Frase del cuerpo de v0.3.1 |
|---|---|---|---|
| B47 | Lo que no puedas leer, no lo suplas (paso 4) | igual | Regla 2: «Si `kitlegal boe` falla o no está, di qué bloques no pudiste consultar y por qué, y no suplas nada de memoria»; paso 5: «como lo que no se pudo leer, se nombra solo con las palabras del texto leído» |
| B48 | Lo que decidas al evaluar el contexto no va en la respuesta: quien pregunta no lo ve (paso 4) | vuelve (H31: «ni lo que devuelven o decides») | Paso 4: «quien pregunta la lee entera: no ve las órdenes, las llamadas ni lo que devuelven, y lo que decides al completar la lectura no va en ella» |
| B49 | La respuesta empieza por lo que se pregunta y habla de la norma, no de la consulta | igual | Paso 4, la negrita de la primera viñeta |
| B50 | La respuesta es todo lo que escribes después de la última orden o llamada, desde su primera palabra; quien pregunta lo lee entero y no ve las órdenes, las llamadas ni lo que devuelven | vuelve (H31: «Es lo escrito tras la última orden o llamada; quien pregunta no las ve») | Paso 4: «Es todo lo que escribes tras la última orden o llamada, desde su primera palabra, y quien pregunta la lee entera: no ve las órdenes, las llamadas ni lo que devuelven» |
| B51 | Le sirven la norma, su texto, su cita, sus avisos de vigencia y las líneas `⚠` de cada bloque, y de eso está hecha | igual | Paso 4: «La hacen la norma, su texto, su cita, sus avisos y las líneas `⚠` de cada bloque» |
| B52 | Escribe como quien tiene el BOE delante: «el artículo no lleva avisos de vigencia», y no «la lectura no trajo avisos»; la norma por su nombre e identificador, sin decir de qué tabla u orden sale | vuelve con sus dos ejemplos | Paso 4: «escribe como quien tiene el BOE delante, «el artículo no lleva avisos de vigencia» y no «la lectura no trajo avisos», la norma por su nombre e identificador, sin decir de dónde sale» |
| B53 | Lo que has hecho, lo que vas a hacer y lo que ha devuelto o no una orden no le sirven | vuelve (H31: «qué haces ni qué devolvió una orden») | Paso 4: «sin decir […] qué has hecho o vas a hacer ni qué ha devuelto o no una orden: no le sirve» |
| B54 | Nada de otra conversación: ni para afirmar, confirmar o desmentir | igual | Paso 4: «**Nada de otra conversación**, ni para afirmar, confirmar o desmentir» |
| B55 | De otras conversaciones, solo la línea `⚠ REDACCIÓN MODIFICADA:` de cada bloque que la trae | igual | Paso 4: «de ellas solo va la línea `⚠ REDACCIÓN MODIFICADA:` de cada bloque que la trae» |
| B56 | Cada afirmación sobre el contenido de una norma lleva su cita, de lo que devolvió `kitlegal boe` en esta conversación | igual | Paso 4, tercera viñeta |
| B57 | La cita: forma legible y, en la misma línea, `[<identificador>, bloque <id>]` como los da la fuente | igual | «Cómo se cita», primera frase |
| B58 | De un precepto no leído, nada: ni qué dice, ni de qué trata, ni cuándo o cómo se aplica, ni en un inciso, ni creyendo saberlo | FR-002/FR-003 (sin cambio en ella) | Paso 5, primera frase |
| B59 | El índice no lo dice: da el número de cada bloque y nada más | FR-002 | Paso 3: «Nombra un precepto solo si `kitlegal boe articulo` o `articulos` ha devuelto su texto en esta conversación […]: si […] su id solo salió en un índice, léelo antes con los pasos 1 y 2, o no lo nombres» |
| B60 | Número y materia nunca juntos, ni afirmando («el art. N, que es…», «el art. N sobre <materia>», «los artículos que regulan <materia> (N y M)»), ni negando («no puedo decir qué <regla> fija»), ni al ofrecer leerlo, ni al avisar de lo no leído | FR-002/FR-003; vuelven los dos ejemplos que H31 quitó | Paso 5, segunda frase, con los cuatro ejemplos |
| B61 | Para decir de qué trata lo remitido, léelo y cítalo; si no, la remisión con las palabras del texto leído y nada detrás | FR-002/FR-003 | Paso 3: «salvo lo del paso 5: si lo remite un bloque leído en el paso 2, […] léelo antes con los pasos 1 y 2, o no lo nombres; leído, di de qué trata, con su cita»; paso 5: «Lo leído para nombrarlo (paso 3) es un salto: lo que remite no se lee para nombrarlo y, como lo que no se pudo leer, se nombra solo con las palabras del texto leído» |
| B62 | Puedes avisar de que una materia se regula en otra parte, sin nombrar precepto ni regla | igual | Paso 5, última frase |
| B63 | Ley y reglamento: el rango de cada una y que la ley prevalece | igual | Paso 4, «**Ley y reglamento**» |
| B64 | Señalar la variación autonómica | igual | Regla 5, última frase |
| B65 | Ordenanzas y normas locales, fuera de esta fuente | igual | Regla 5 |
| B66 | Cada aviso de vigencia con su forma fija | igual | Regla 3 |
| B67 | De la redacción leída, qué norma la dio y desde cuándo rige | igual | Paso 4, última viñeta |
| B68 | Hasta cuándo, nunca: ninguna lectura lo trae, tampoco en una derogada | igual | Paso 4, última viñeta |
| B69 | `kitlegal graph check` no dice nada de la vigencia | igual | Paso 4, última viñeta |

FR-002 y FR-003, además de B59 a B61, en las frases del paso 3 («Nombra un precepto solo si …») y del paso 5 («Lo
leído para nombrarlo (paso 3) es un salto: …»). «Precepto» es el término del título del paso 5, cuya primera frase lo
dice: un artículo, un apartado o una disposición. El paso 3 conserva su regla de hoy, «Lee lo remitido
que cambia la respuesta, de esta norma o de otra, resuelta antes con los pasos 1 y 2», sin límite de saltos; el salto
es solo de lo leído **para nombrarlo**, como fija la Clarificación Q2 (research D4).

## 3. Lo que sale del cuerpo (FR-006)

Siete textos que solo pide una clase de pregunta. Cada uno va a una referencia a mano que ya existe, y el cuerpo la
enlaza en la frase que dice cuándo leerla. Ninguno es del protocolo que toda respuesta necesita, de una regla invariante,
de FR-001 a FR-004 ni una forma fija (research D6).

| # | Texto que sale | De | A | Clase de pregunta | Frase del cuerpo que la enlaza y cuándo |
|---|---|---|---|---|---|
| 1 | La regla 7 de hoy entera: «**Si `kitlegal graph check` no termina con `0`.** Si, tras una lectura sin fecha, termina con otro código o `graph_check` da un error, cita lo leído y lleva, solo entonces, la frase de [lecturas_que_fallan]…» | Reglas, 7 | `lecturas_que_fallan.md`, «Si `kitlegal graph check` falla», que ya la dice entera (sin cambios) | La comprobación tras una lectura sin fecha falla | Paso 2: «Si una orden no termina con `0`, antes de seguir: [lecturas_que_fallan]» (la comprobación va en la misma orden que la lectura; con herramientas, su error, `con_herramientas.md`) |
| 2 | «se cita la redacción vigente ese día, leída sin `kitlegal graph check`» y «Cada línea, solo de una lectura con fecha» | «Redacción en una fecha» | `redaccion_en_una_fecha.md`, «Qué es una fecha», «Cómo se lee» y «Las tres líneas» (sin cambios) | Pregunta con fecha | Paso 2: «Sin fecha —con ella, [redaccion_en_una_fecha]—, lee cada bloque…»; el enlace pasa del apartado al paso que lee |
| 3 | «una línea, con su cita y sus dos fechas» | «Redacción modificada» | `redaccion_modificada.md`, «La línea, por bloque» (sin cambios) | Un bloque con `version-obsoleta` | «Redacción modificada»: «Por cada entrada `version-obsoleta` de `kitlegal graph check`, y si se pregunta qué cambió, lee [redaccion_modificada]» (la comprobación es la del paso 2) |
| 4 | «ni `fuente-caducada` ni nada de un bloque sin `version-obsoleta`» | «Redacción modificada» | `redaccion_modificada.md`, apartado nuevo «Lo demás de la comprobación» (§4) | La comprobación trae otras entradas | La misma de la fila 3; el cuerpo conserva «De `kitlegal graph` solo se usa `check`, y nada más suyo va en la respuesta» |
| 5 | «Releer el bloque apagaría su línea `⚠ REDACCIÓN MODIFICADA:`» | Paso 2 | `redaccion_modificada.md`, apartado nuevo «Una lectura por bloque» (§4) | Un bloque con `version-obsoleta` | La misma de la fila 3; el paso 2 conserva «lee cada bloque una vez» |
| 6 | «di «no encontrada con esta búsqueda» y propón reformularla» | Reglas, 1 | `buscar_la_norma.md`, «Si no aparece» (§4) | La norma no está en la tabla y la búsqueda no la da | Paso 1: «si no está, búscala por su título con [buscar_la_norma]»; la regla 1 conserva que no aparecer no prueba que no exista |
| 7 | «(«Comandos»)», de dónde salen los nombres de las herramientas | Protocolo | `con_herramientas.md`, «Nombres», que ya lo dice (sin cambios) | Con herramientas | Protocolo: «tras leer [con_herramientas]» |

Consecuencia de la fila 1: la regla 8 de hoy pasa a ser la 7, la regla 2 remite a «la regla 7» y `CONTRIBUTING.md`
(línea 743) deja de llamar «la frase de la regla 7» a la de las lecturas que fallan (contracts/guia-y-changelog.md §3).

## 4. Las referencias a mano que cambian

Las demás (`como_se_cita.md`, `con_herramientas.md`, `lecturas_que_fallan.md`, `redaccion_en_una_fecha.md`,
`varios_bloques.md`) y la generada `normas.md` no cambian. «Si faltó, pídela ya», de la comprobación, se queda en el
paso 2 (§1.1): es del protocolo y vale en los dos modos (B36 de la tabla de H31).

### 4.1 `references/redaccion_modificada.md`

El párrafo de entrada pasa a decir también lo que no lleva:

```markdown
Qué compara `kitlegal graph check` tras una lectura sin fecha, cuándo y cómo lleva la respuesta la línea de
«Redacción modificada» de `SKILL.md`, y qué no lleva.
```

Y, delante de «## La redacción superada», dos apartados nuevos (filas 4 y 5 de §3):

```markdown
## Lo demás de la comprobación

Una entrada `fuente-caducada`, o cualquiera de un bloque sin `version-obsoleta`, no va en la respuesta, que cita el
texto que acabas de leer.

## Una lectura por bloque

No releas un bloque sin fecha: una segunda lectura apagaría su línea REDACCIÓN MODIFICADA.
```

Ninguna línea empieza por `⚠`: la etiqueta va sin la marca, como en el resto de la referencia
(`lineas-de-aviso-de-las-referencias`).

### 4.2 `references/buscar_la_norma.md`

«Si no aparece» (fila 6 de §3) pasa a ser:

```markdown
Reformula la búsqueda con otras palabras del título. Que no aparezca no prueba que no exista (regla 1 de `SKILL.md`):
di «no encontrada con esta búsqueda» y propón reformularla.
```

## 5. Lo que comprueba `make ci` de este contrato

Los gates y controles de H31, sin cambios, sobre la skill entera (plan.md, «Controles de umbral»): el cuerpo de 11 000
caracteres como mucho y de menos de 500 líneas, el frontmatter, cada referencia enlazada desde `SKILL.md` y ninguna
desde otra, `prosa-de-la-skill`, `ordenes-para-powershell`, `linea-sin-consulta`, `formas-de-la-redaccion-fechada`,
`avisos-de-la-skill`, `hallazgos-de-la-skill`, `lineas-de-aviso-de-las-referencias`, `ordenes-de-ejemplo`,
`normas-nombradas` y `sin-instrucciones-de-evals`, y `TestSkillsEmpotradas`. Ninguna orden concreta nueva de
`kitlegal boe` entra en el cuerpo ni en una referencia (FR-007): las menciones nuevas de `kitlegal boe articulo` y
`articulos` van sin argumentos. La tabla de §2, el paso 3 y el paso 5 los coteja la revisión final (SC-003).
