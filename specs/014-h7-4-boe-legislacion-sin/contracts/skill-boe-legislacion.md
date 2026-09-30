# Contrato: `boe-legislacion` v0.1.4

`skills/boe-legislacion/SKILL.md`. Cada cambio lleva el texto de v0.1.3 que sustituye, el de v0.1.4, su causa (research,
«Causa de raíz») y su requisito. El prototipo con todos los cambios (research V19) tiene 294 líneas, una `description`
de ≈ 925 caracteres y 0 párrafos de prosa con alguna expresión de la lista nueva o una fecha `AAAAMMDD` con cifras; el
`SKILL.md` de v0.1.4 tenía 295 líneas con C1-C10, porque se quedaba la frase que cierra «Redacción modificada» (C9), y
tiene 298 con C11, que la absorbe en una viñeta (reparación del cierre), sigue en 298 con C12 (segunda reparación del
cierre), y tiene una `description` de 924 caracteres. Todo lo que no nombra este contrato queda como en v0.1.3 (§2).

## 1. Cambios

### C1 · La `description` (activación; FR-001, FR-002, FR-003, FR-026)

v0.1.3: «… Úsala cuando se pregunte qué dice un artículo, una ley o un real decreto; cuando se nombre una norma por su
número y año (…), por su abreviatura (…) o por su identificador BOE-A-…; o cuando haya que citar el texto vigente de
una norma estatal o autonómica consolidada en el BOE. Lee el índice…».

v0.1.4 (la frase de las materias, delante, no cambia):

```yaml
description: >-
  Consulta y cita normativa consolidada del Boletín Oficial del Estado (BOE) de cualquier materia: procedimiento
  administrativo, contratación pública, régimen local, tributos y haciendas locales, régimen jurídico del sector
  público, transparencia, relaciones laborales… Úsala siempre que la respuesta dependa de lo que dice una norma —qué
  dice un artículo, una ley o un real decreto, qué plazo, requisito o procedimiento fija, dónde se regula una
  materia—, también cuando creas conocer la respuesta: sin leer la norma con el binario, la respuesta no tiene cita.
  La norma puede nombrarse por su número y año (Ley 39/2015, Real Decreto Legislativo 2/2004), por su abreviatura
  (LPAC, LCSP, LRBRL, LGT, TRLRHL, LRJSP) o por su identificador BOE-A-…, o pedirse el texto vigente de una norma
  estatal o autonómica consolidada en el BOE. Lee el índice y los artículos con el binario kitlegal y responde citando
  identificador y bloque.
```

Causa: el modelo trata la skill como una comprobación opcional de lo que cree saber, y la pregunta de la 04 pide un
dato («¿En cuántos años…?») y no «qué dice» (research, activación, 1-3).

### C2 · Paso 3, la orden de `articulo` y su forma para PowerShell (FR-021, FR-024, FR-025)

v0.1.3: «Lee los bloques de uno en uno con `kitlegal boe articulo`. La orden que lee un bloque comprueba también, detrás
de la lectura, si su redacción ha cambiado desde una lectura anterior:» y el bloque `bash`.

v0.1.4:

````markdown
- Lee los bloques de uno en uno con `kitlegal boe articulo`, con `kitlegal graph check` de la misma norma y el mismo
  bloque detrás, en la misma orden:

  ```bash
  kitlegal boe articulo BOE-A-2015-10565 a21 --json && kitlegal graph check BOE-A-2015-10565 a21 --json
  ```

  En PowerShell, que en su versión 5.1 no tiene `&&`, la misma orden es:

  ```powershell
  kitlegal boe articulo BOE-A-2015-10565 a21 --json; if ($LASTEXITCODE -eq 0) { kitlegal graph check BOE-A-2015-10565 a21 --json }
  ```
````

Causa: «comprueba también… si su redacción ha cambiado desde una lectura anterior» (líneas 65-66) es la frase que repite
la respuesta; y la orden de Bash no funciona en Windows PowerShell 5.1 (research D5, S1).

### C3 · Paso 3, los dos sobres (FR-021, FR-025)

v0.1.3 (líneas 72-77): «Devuelve dos sobres. El primero trae el texto del bloque y sus avisos de vigencia. El segundo es
el de la comprobación: por cada entrada que traiga con la clase `version-obsoleta`, la respuesta lleva la línea
`⚠ REDACCIÓN MODIFICADA:` de «Redacción modificada desde una lectura anterior»; si la comprobación termina con otro
código que `0`, la regla 7. La comprobación va siempre así, …: no la pidas nunca sin argumentos ni antes de leer, y si un
bloque se ha leído sin ella, pídela a continuación con esa norma y ese bloque.»

v0.1.4: «Devuelve dos sobres: el del bloque, con su texto y sus avisos de vigencia, y el de `kitlegal graph check`. Por
cada entrada de clase `version-obsoleta` del segundo, la respuesta lleva una línea `⚠ REDACCIÓN MODIFICADA:`
(«Redacción modificada»); si `kitlegal graph check` termina con otro código que `0`, la regla 7. `kitlegal graph check`
va siempre así, detrás de la lectura, en su misma orden y con su misma norma y sus mismos bloques: no lo pidas nunca
sin argumentos ni antes de leer, y si un bloque se ha leído sin él, pídelo a continuación con esa norma y ese bloque.»

Causa: «la comprobación» y «desde una lectura anterior» (research, clase A, 72-77).

### C4 · Paso 3, la orden de `articulos` y su forma para PowerShell (FR-021, FR-024, FR-025)

v0.1.3 (líneas 79-81): «…con la comprobación de esos mismos bloques detrás (`kitlegal boe articulos <norma> <bloques>...
--json && kitlegal graph check <norma> <bloques>... --json`).»

v0.1.4: «…solo cuando necesites varios bloques a la vez y todos salgan del índice, con `kitlegal graph check` de esos
mismos bloques detrás:» y, detrás, los dos bloques:

````markdown
  ```bash
  kitlegal boe articulos <norma> <bloques>... --json && kitlegal graph check <norma> <bloques>... --json
  ```

  ```powershell
  kitlegal boe articulos <norma> <bloques>... --json; if ($LASTEXITCODE -eq 0) { kitlegal graph check <norma> <bloques>... --json }
  ```
````

### C5 · Paso 3, una lectura por bloque (FR-021)

v0.1.3 (82-83): «…apagaría el aviso de que su redacción ha cambiado desde una lectura anterior (más en «Redacción
modificada desde una lectura anterior»).» → v0.1.4: «…apagaría su línea `⚠ REDACCIÓN MODIFICADA:` (más en «Redacción
modificada»).»

### C6 · Paso 4 (FR-010, FR-020)

v0.1.3: «Comprueba si lo leído basta:» y, al final, «Si falta algo que no puedes leer con `kitlegal boe`, dilo en la
respuesta en lugar de suplirlo.»

v0.1.4: «Mira si hace falta leer algo más para responder:» y, al final, «Si falta algo que no puedes leer con
`kitlegal boe`, dilo en la respuesta en lugar de suplirlo. Lo que decidas en este paso no va en la respuesta: quien
pregunta no ve los pasos.»

Causa: «tengo suficiente para responder…; no necesito el artículo 60» y «I have enough to respond now» (14-02 y 14-01 de
`196ee05`) dicen en la respuesta la decisión de este paso.

### C7 · Paso 5, lo que lleva la respuesta (FR-010)

v0.1.3: «…sus avisos de vigencia y, si la redacción cambió, la línea `⚠ REDACCIÓN MODIFICADA:`. La respuesta está hecha
de eso: no cuentes lo que has hecho ni lo que ha devuelto ninguna orden.»

v0.1.4: «…sus avisos de vigencia y la línea `⚠ REDACCIÓN MODIFICADA:` de cada bloque que la trae. La respuesta está hecha
de eso: no cuentes lo que has hecho, lo que vas a hacer ni lo que ha devuelto ninguna orden.»

### C8 · Paso 5, «Nada de otra conversación» (FR-011, FR-021)

v0.1.3: «Lo único que la respuesta dice de una lectura anterior es la línea `⚠ REDACCIÓN MODIFICADA:` de un bloque cuya
redacción ha cambiado.» → v0.1.4: «De lo leído en otras conversaciones, la respuesta solo lleva la línea
`⚠ REDACCIÓN MODIFICADA:` de cada bloque que la trae (más en «Redacción modificada»).»

### C9 · La sección «Redacción modificada» (FR-011, FR-021, FR-022, FR-023)

Título: «## Redacción modificada desde una lectura anterior» → «## Redacción modificada». Cuerpo de v0.1.4:

````markdown
`kitlegal` recuerda en local los bloques que ha leído con `kitlegal boe articulo` o `articulos` y qué redacción vio
cada vez. `kitlegal graph check <norma> <bloques>... --json` compara, para esa norma y esos bloques, la redacción que
acabas de leer con la que se leyó la vez anterior, y devuelve en `data.hallazgos` una entrada por cada bloque en que
encuentra algo, con su `clase`. De `kitlegal graph`, el protocolo solo usa `check`, y siempre detrás de una lectura, en
su misma orden (paso 3): una vez por cada orden que lee bloques.

- Una entrada de `clase` `version-obsoleta` dice que la redacción que acabas de leer de ese bloque no es la que se leyó
  la vez anterior. La respuesta lo dice con una línea por bloque, con su forma fija: `⚠ REDACCIÓN MODIFICADA:` —`⚠`, la
  etiqueta `REDACCIÓN MODIFICADA` y dos puntos—, la cita del bloque como en «Cómo se cita» y dos puntos, y detrás, en
  la misma línea, las dos fechas de vigencia tal como las da esa entrada (`AAAAMMDD`): la de la redacción superada
  (`fecha_vigencia`) y la de la que acabas de leer (`fecha_vigencia_reciente`). La forma, con marcadores en lugar de
  datos:

  ```text
  ⚠ REDACCIÓN MODIFICADA: <forma legible> [<identificador>, bloque <id>]: la redacción con fecha de vigencia AAAAMMDD, la que se consultó antes, ha sido sustituida por la de AAAAMMDD, que es la que se cita.
  ```

  Con dos bloques en `version-obsoleta`, dos líneas, cada una con su cita y sus fechas. Decirlo con otras palabras no
  vale: la línea va con su forma fija.
- Una entrada de `clase` `fuente-caducada` no va en la respuesta: la respuesta cita el texto que acabas de leer, que la
  caché no sirve pasada su vigencia.
- **La redacción superada no la has leído.** `kitlegal boe articulo` da solo la redacción vigente, y
  `kitlegal graph check`, dos fechas: nada de lo que devuelven dice qué decía la redacción superada ni en qué se
  diferencia de la vigente. La respuesta no lo dice, ni lo resume, ni lo compara, aunque creas saberlo: sería texto
  legal sin fuente. Sí dice lo que da la lectura: el texto vigente con su cita, qué norma le dio esa redacción
  (`norma_modificadora`) y desde cuándo está vigente (`fecha_vigencia`).
- Si quien pregunta quiere saber qué cambió, la respuesta dice que cita la redacción vigente y que la que había antes
  no la ha leído.
- La línea `⚠ REDACCIÓN MODIFICADA:` es todo lo que la respuesta dice de lo leído en otras conversaciones. Si
  `kitlegal graph check` termina con otro código que `0`, la regla 7.
````

Detrás de la lista, la frase que cerraba la sección en v0.1.3, «La etiqueta `REDACCIÓN MODIFICADA` no es la de ningún
aviso de vigencia.», se queda como estaba (supuesto de T010).

Causa: la clase A (título, 185, 198) y la B (research, clase B, 1-3). La línea con la cita: research D4.

### C10 · Regla 7 (FR-021, FR-025)

Título: «**Si la comprobación de la redacción no termina con `0`.**» → «**Si `kitlegal graph check` no termina con `0`.**».
«…y lleva esta frase, sin afirmar que la redacción ha cambiado ni que no:» → «…y lleva esta frase, tal cual y sin nada
más sobre la redacción:». «…la orden termina ahí, sin comprobación, y vale la regla 2. La línea `⚠ REDACCIÓN
MODIFICADA:` solo va cuando la comprobación termina con `0` y trae `version-obsoleta` («Redacción modificada desde una
lectura anterior»).» → «…la orden termina ahí, sin `kitlegal graph check`, y vale la regla 2. La línea `⚠ REDACCIÓN
MODIFICADA:` solo va cuando `kitlegal graph check` termina con `0` y trae `version-obsoleta` («Redacción modificada»).».
**La frase fija no cambia**: «No se ha podido comprobar si la redacción ha cambiado desde una consulta anterior.».

### C11 · El sobre de `kitlegal graph check` sin entradas no da nada a la respuesta (reparación del cierre; FR-010, FR-021, FR-025)

Lo añade `reparar_cierre` tras la medición sobre `4cb52cb` (research V28, D22): con C1-C10, 15 de 54 respuestas del
modelo que decide llevan una expresión de la lista, 14 de ellas porque cuentan, casi siempre en un párrafo final de
vigencia, que `kitlegal graph check` no encontró nada. Cinco cambios, todos en la prosa, sin tocar ninguna forma fija
ni la región generada; 298 líneas.

- **Paso 3, los dos sobres** (C3). v0.1.4 con C1-C10: «…y el de `kitlegal graph check`. Por cada entrada de clase
  `version-obsoleta` del segundo, la respuesta lleva una línea `⚠ REDACCIÓN MODIFICADA:` («Redacción modificada»); si
  `kitlegal graph check` termina con otro código que `0`, la regla 7.» → «…y el de `kitlegal graph check`, del que la
  respuesta lleva una línea `⚠ REDACCIÓN MODIFICADA:` por cada entrada de clase `version-obsoleta` de su
  `data.hallazgos`, y nada más: vacío, que es lo habitual, no da nada a la respuesta («Redacción modificada»); si
  `kitlegal graph check` termina con otro código que `0`, la regla 7.» Lo demás del párrafo no cambia.
- **Paso 5, lo que lleva la respuesta** (C7). «…ni lo que ha devuelto ninguna orden.» → «…ni lo que ha devuelto ninguna
  orden; tampoco lo que una orden no ha devuelto: un `data.hallazgos` vacío no deja rastro en la respuesta.»
- **Paso 5, los avisos de vigencia.** Entre «(más en «Cómo se cita»).» y «Recuerda que los textos consolidados…» entra:
  «De la vigencia del bloque, la respuesta dice lo que trae el sobre de `kitlegal boe`; el de `kitlegal graph check`
  no dice nada de ella.»
- **«Redacción modificada»** (C9). En el párrafo de apertura, «…con la que se leyó la vez anterior, y devuelve…» →
  «…con la que se leyó la vez anterior, en otra conversación que quien pregunta no conoce, y devuelve…». En la viñeta
  de `version-obsoleta`, «Decirlo con otras palabras no vale: la línea va con su forma fija.» → «Decirlo con otras
  palabras no vale, ni en lugar de la línea ni además de ella: la línea va con su forma fija y lo dice entera.» Detrás
  de la viñeta de `fuente-caducada` entra una viñeta nueva, que absorbe la última de C9 («La línea
  `⚠ REDACCIÓN MODIFICADA:` es todo lo que la respuesta dice de lo leído en otras conversaciones. Si `kitlegal graph
  check` termina con otro código que `0`, la regla 7.», cuya remisión ya hace el paso 3) y la frase que cerraba la
  sección («La etiqueta `REDACCIÓN MODIFICADA` no es la de ningún aviso de vigencia.», supuesto de T010):

  ````markdown
  - **No es un aviso de vigencia.** Un aviso es un dato de la norma, que el BOE da con el bloque, y la etiqueta
    `REDACCIÓN MODIFICADA` no es la de ninguno: una entrada de `kitlegal graph check` es un dato de kitlegal sobre otras
    conversaciones, que quien pregunta no conoce. Por eso un bloque sin entrada de `clase` `version-obsoleta` no da nada
    a la respuesta: ni una línea, ni una frase junto a los avisos o a la fecha de vigencia, ni una palabra al final.
    Mencionarlo sería contar lo que devolvió una orden y hablar de otras conversaciones: la línea es todo lo que la
    respuesta dice de ellas y de `kitlegal graph check`.
  ````

- **Regla 7** (C10). Delante de «Si lo que falla es la lectura…» entra: «La frase es solo de ese caso: con `0`, la
  respuesta no la lleva, ni afirmada ni negada, ni con otras palabras.»
- **Paso 3, una línea en blanco** entre las viñetas «Una orden de `kitlegal boe articulos` falla entera…» y «Sigue las
  remisiones…» se quita, sin cambiar el texto: con 300 líneas o más `skills-check` falla, y el mutante `dos-inicios` de
  `TestSkillsDelRepositorio` añade una línea a la skill, así que el tope efectivo es 298.

Causa: la clase A no es solo vocabulario. Con la prosa de C1-C10, el modelo que decide sigue contando el resultado
negativo de `kitlegal graph check`, con las palabras de las dos formas fijas (que no cambian, FR-025) y de la etiqueta:
«no se ha detectado cambio de redacción», «no consta que su redacción haya cambiado desde una consulta anterior», «la
comprobación de cambios de redacción no encontró ninguno». La skill decía qué lleva la respuesta cuando hay una entrada
y nunca qué pasa cuando no la hay, y el modelo trataba ese resultado como un dato de la vigencia del bloque, al lado de
«no trae avisos de vigencia» (que sí es derecho, FR-012). C11 lo dice con la razón, como hizo C9 con la clase B (que da
0 de 54): por qué (es un dato sobre otras conversaciones, no de la norma), qué sí (la línea por entrada) y dónde no (ni
junto a los avisos ni al final). Ninguna frase nueva enseña una forma de la lista: `prosa-de-la-skill` y
`expresiones-de-la-skill` siguen en verde.

### C12 · El fin de una redacción no lo trae ningún sobre (segunda reparación del cierre; FR-011, FR-020, FR-022)

Lo añade `reparar_cierre` tras la medición sobre `ad68a70` (research V30, D23): con C1-C11, `redaccion_no_leida` da 1 de
54 (umbral 0), la 18-01, que presenta la última redacción del art. 42 de la Ley 30/1992, derogada, como «Su redacción
vigente hasta la derogación es la de la Ley 4/1999 (vigencia desde 14/04/1999)». Dos cambios en la prosa, sin tocar
ninguna forma fija, la regla 3, la forma de los avisos ni la región generada; 298 líneas.

- **Paso 5, los avisos de vigencia** (C11). «De la vigencia del bloque, la respuesta dice lo que trae el sobre de
  `kitlegal boe`; el de `kitlegal graph check` no dice nada de ella.» → «De la vigencia del bloque, la respuesta dice
  lo que trae el sobre de `kitlegal boe`: sus avisos y, de la redacción leída, qué norma la dio (`norma_modificadora`)
  y desde cuándo rige (`fecha_vigencia`). Hasta cuándo, nunca: ningún sobre trae el fin de una redacción, tampoco en
  una norma derogada, cuyo aviso no lleva fecha, y darlo sería texto legal sin fuente. El de `kitlegal graph check` no
  dice nada de ella.» Lo demás de la viñeta no cambia.
- **«Redacción modificada», la viñeta «La redacción superada no la has leído»** (C9). «…qué norma le dio esa redacción
  (`norma_modificadora`) y desde cuándo está vigente (`fecha_vigencia`).» → «…qué norma le dio esa redacción
  (`norma_modificadora`) y desde cuándo rige (`fecha_vigencia`); hasta cuándo, nunca (paso 5).»
- **Paso 3, dos líneas en blanco** entre el cierre de un bloque de código y la viñeta siguiente (la de `indice` y la de
  `metadatos`) se quitan, sin cambiar el texto, como ya iba la de `articulos` con «Lee cada bloque una sola vez»: con
  las dos líneas que gana el paso 5, la skill se queda en 298, el tope efectivo (C11).

Causa: la prosa decía qué se dice de una redacción —la norma que la dio y «desde cuándo está vigente»— y la regla 3,
que no se presente como vigente el texto de una norma derogada; con una norma derogada, el modelo concilia las dos
convirtiendo la fecha de inicio en un intervalo cuyo fin ningún sobre da: `kitlegal boe articulo` trae `fecha_vigencia`
y `norma_modificadora`, `kitlegal boe metadatos` trae `estatus_derogacion` sin fecha, y el aviso `derogada` dice que lo
está, sin cuándo. Es la clase B por la definición de la lista (§1 de su cabecera: «hasta cuándo rigió»), y la forma
`vigente hasta` no puede salir de la lista ni acotarse: es la única de la clase B que marca la 19-02 de `196ee05` en el
calibrado (FR-032, FR-047). Ya en H7.1 la 18-02 razonaba igual sin caer en una forma («la última redacción vigente
(fecha de vigencia 1999-04-14) antes de la derogación»). C12 lo dice con la razón, como C9 y C11: qué sí (la norma y
desde cuándo) y por qué no el fin (ningún sobre lo trae; sería texto legal sin fuente).

## 2. Lo que se queda de v0.1.3 (FR-025, FR-026)

La lectura de uno en uno, con `kitlegal graph check` detrás en la misma orden y con su misma norma y sus mismos bloques,
una vez por cada orden que lee bloques; «Cómo se cita» y los avisos de vigencia, con sus formas y sus ejemplos; la frase
fija de la regla 7; las reglas 1 a 6; «Nada de otra conversación» (su primera frase); la región generada de la tabla de
comandos, intacta (`make skills-check`); ninguna fecha `AAAAMMDD` escrita con cifras; ningún nombre de eval, del job ni
de un modelo (H5 FR 077); frontmatter con `name`, `description` y `metadata` como hoy.

## 3. La prosa y las formas, comprobadas en `make ci`

- **Prosa** (`prosa-de-la-skill`, FR-021, FR-093; SC-004): ningún párrafo fuera de los bloques delimitados, del código en
  línea y de la región generada lleva una expresión de la lista, y ninguna línea lleva una fecha `AAAAMMDD` con cifras.
- **Formas** (`expresiones-de-la-skill`, FR-033; SC-003): la respuesta compuesta con los bloques `text` de `SKILL.md` y
  la forma escrita de cada aviso y de cada clase de hallazgo, con datos de ejemplo, no se marca una vez quitadas las
  formas fijas
  ([lista-de-expresiones.md](./lista-de-expresiones.md) §5).
- **PowerShell** (`ordenes-para-powershell`, FR-095; SC-005): §4.
- **Tamaño y frontmatter** (`skills-check`, FR-026; SC-011): < 300 líneas, `description` ≤ 1024 caracteres.

## 4. La comprobación de las órdenes para PowerShell (FR-095)

`defectosDeLasOrdenesParaPowerShell(markdown string) []string`, en `internal/evals/conjunto_test.go` (solo la usan los
tests, como `defectosDeLaProsa`), sobre el texto entero de `SKILL.md`; una línea por defecto, nil sin ninguno:

1. Busca, en su orden, cada **orden de Bash** —`kitlegal boe <verbo> <norma> <bloques…> --json && kitlegal graph check
   <norma> <bloques…> --json`, con `<verbo>` `articulo` o `articulos`— y cada **orden de PowerShell** —`kitlegal boe
   <verbo> <norma> <bloques…> --json; if ($LASTEXITCODE -eq 0) { kitlegal graph check <norma> <bloques…> --json }`—,
   en bloques delimitados o en código en línea. Las palabras se separan por blancos; `<norma>` y cada bloque son una
   palabra sin blancos (un marcador como `<bloques>...` es una palabra).
2. Error si no hay ninguna orden de Bash con `articulo` o ninguna con `articulos` (SC-005: 2 de 2).
3. Por cada orden de Bash, la **siguiente** orden encontrada tiene que ser de PowerShell, con el mismo verbo, la misma
   norma y los mismos bloques, en el mismo orden; error que nombra la orden de Bash si no lo es, si falta o si es otra
   forma (la de `&&`, sin el `if`, o con otra condición). Una orden sin el `if` o con otra condición no es ninguna de
   las dos formas y no se cuenta; la de PowerShell escrita con `&&` es otra orden de Bash, y da también su propio
   error si no la sigue la suya (supuesto de T010).
4. En cada orden, de las dos formas, la comprobación lleva la misma norma y los mismos bloques que la lectura; error si no.

La usa la subprueba `ordenes-para-powershell` de `TestEvalsDelRepositorio` sobre `SKILL.md`; `TestOrdenesParaPowerShell`
la fija con Markdown sintético: las dos órdenes con sus dos formas, en bloques o en código en línea (sin error); sin la
forma de PowerShell de la primera o de la última; con `&&` en PowerShell; sin el `if`; con otra condición; con otra
norma o con otros bloques en la de PowerShell; con otros bloques en la comprobación, en la de Bash o en la de
PowerShell; sin argumentos en la comprobación; sin la de `articulos`; y sin ninguna orden.

## 5. Uso, de fuera adentro

| Salida | Quién, cuántas veces | Tamaño con meses de uso | Cuándo se apaga |
|---|---|---|---|
| La respuesta | la persona; una por pregunta | lo que ocupe la norma citada; 0 líneas sobre la comprobación; una línea `⚠ REDACCIÓN MODIFICADA:` por bloque cambiado, de 221 bytes con `art. 118 de la Ley 9/2017 [BOE-A-2017-12902, bloque a1-30]` y 242 con `disposición adicional tercera de la Ley 9/2017 [BOE-A-2017-12902, bloque da-3]` (medidos; la de H7.3, 161): 0 en la mayoría de las preguntas, 2 en la de la eval 20, como mucho k × ≈ 240 B con k bloques leídos; con cientos de normas y miles de bloques consultados, la misma, porque solo cuenta lo leído en la pregunta | la línea sale en la respuesta de la lectura que ve la redacción nueva y no en la siguiente lectura del bloque (H7.1); al cabo de un mes, solo si el BOE publica otra |
| `kitlegal graph check <norma> <bloques…>` | la skill; una por orden que lee bloques, detrás de la lectura | ≈ 325 B sin cambios, ≈ 1 000 B por bloque cambiado (H7.3), ≤ 3 800 B con cinco; acotada por los bloques pedidos | cada señal se da una vez y la apaga la lectura siguiente (H7.1) |
| La `description` | el agente, en cada sesión con la skill instalada | ≈ 925 caracteres (≤ 1024) | no da señales |
| El cuerpo de `SKILL.md` | el modelo, una vez por conversación en que se activa; la orden de PowerShell, en cada lectura en Windows con PowerShell | 298 líneas (< 300; 294 en el prototipo, 295 con C1-C10, 298 con C11 y con C12) | no da señales |

## 6. `CHANGELOG.md` (*Unreleased*) (FR-027, FR-101)

Una entrada `boe-legislacion` v0.1.4 en *Cambiado*: se activa ante toda pregunta cuya respuesta dependa de lo que dice
una norma, también si el modelo cree saberla; la respuesta no cuenta la comprobación ni lo que el agente va a hacer, y no
describe una redacción que no ha leído (dice que cita la vigente y que la anterior no la ha leído); cada línea
`⚠ REDACCIÓN MODIFICADA:` dice de qué precepto es, con su cita; y cada orden de lectura y comprobación tiene su forma
para PowerShell.
