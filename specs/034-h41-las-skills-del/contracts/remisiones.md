# Contrato: las remisiones de `normativa` y `jurisprudencia`, lo que se mueve y su control

FR-001 a FR-010; SC-001 a SC-005. Research D1 a D8, V2 a V5; data-model §5.

## 1. Términos y cuerpos

- **Cuerpo**: lo que sigue a la línea `---` que cierra el frontmatter de un `SKILL.md`, contado en runas (research V2).
- **Control**: `laboratorio/<skill>/<skill>-control/skills/<skill>/`, que es `skills/<skill>` de `v0.9.0` (research V1).
- **Remisión**: la frase del cuerpo que enlaza `references/<f>` (§6).
- Los cambios van como `diff -u` del `SKILL.md` de control al de `skills/<skill>` tras el hito; cada `-` y cada `+` es una
  línea entera, con su sangría, y nada fuera de esos tramos cambia: ni el frontmatter, ni la tabla de comandos, ni la
  región de los paquetes (FR-009).

| Skill | Cuerpo de control | Tras el hito | Con la vertical de ensayo (+70) | Movimientos |
|---|---|---|---|---|
| `normativa` | 10 929 | **10 926** | 10 996 | dos (§3) |
| `jurisprudencia` | 10 842 | **10 842** | 10 912 | ninguno |
| `plazos-y-territorio` | 10 930 | 10 930, sin cambios (FR-010) | 11 000 | ninguno |

Medidas en esta sesión sobre copias con estos cambios (research V3). Ninguna referencia nueva; `consulta-juridica` no
cambia.

## 2. `normativa` (FR-004)

Los cambios de la #245 (su contrato, §2), con uno solo distinto, marcado `(D1)`: la «y» que juntaba dos casos en la
orden de `redaccion_modificada.md` pasa a «o» (FR-002; research D1).

```diff
--- control/SKILL.md
+++ skills/normativa/SKILL.md
@@ -17,13 +17,14 @@
 
 ## Protocolo
 
-Sigue los pasos en orden, cada operación con `kitlegal … --json` desde el `PATH` o, si la tienes, con su
-herramienta, tras leer [con_herramientas](references/con_herramientas.md).
+Antes de la primera orden o llamada, lee [con_herramientas](references/con_herramientas.md): cómo se pide cada
+una con herramienta. Sigue los pasos en orden, cada operación con `kitlegal … --json` desde el `PATH` o, si la
+tienes, con su herramienta.
 
 ### 1. Identificar la norma
 
-Toma el `BOE-A-…` de la tabla de [normas](references/normas.md); si no está, búscala con
-[buscar_la_norma](references/buscar_la_norma.md).
+Lee [normas](references/normas.md) y toma el `BOE-A-…` de su tabla; si no está, lee
+[buscar_la_norma](references/buscar_la_norma.md) y búscala como dice.
 
 <!-- paquetes: lo genera make skills-sync -->
 <!-- fin de los paquetes -->
@@ -33,8 +34,8 @@
 - El id de cada bloque sale del índice, con fecha o sin ella: `kitlegal boe indice BOE-A-1978-31229 --json`. Copia el
   id de la entrada cuyo `titulo` es el artículo; nunca lo compongas con su número ni lo tomes de la pregunta: no
   siempre es su id.
-- Sin fecha —con ella, [redaccion_en_una_fecha](references/redaccion_en_una_fecha.md)—, lee cada bloque una vez, con
-  su comprobación detrás, en la misma orden:
+- Con una fecha en la pregunta, lee [redaccion_en_una_fecha](references/redaccion_en_una_fecha.md) antes de leer el
+  bloque: cómo se lee. Sin fecha, lee cada bloque una vez, con su comprobación detrás, en la misma orden:
   ```bash
   kitlegal boe articulo BOE-A-1978-31229 a140 --json && kitlegal graph check BOE-A-1978-31229 a140 --json
   ```
@@ -45,9 +46,11 @@
   La comprobación, nunca sin argumentos, ni antes, ni con fecha; sin fecha, si faltó, pídela ya. Habla de otras
   conversaciones, que quien pregunta no conoce: un bloque sin entrada `version-obsoleta` no da nada a la respuesta, ni
   una línea, ni una frase junto a los avisos o a la fecha de vigencia. Ni «no consta otra redacción». Las líneas de una
-  fecha van siempre.
-- Varios bloques a la vez: [varios_bloques](references/varios_bloques.md). Si una orden no termina con `0`, antes de
-  seguir: [lecturas_que_fallan](references/lecturas_que_fallan.md).
+  fecha van siempre. Por cada entrada `version-obsoleta`, o si se pregunta qué cambió, lee
+  [redaccion_modificada](references/redaccion_modificada.md): qué se dice.
+- Con avisos o un id que termina en punto, lee [como_se_cita](references/como_se_cita.md): cómo se citan.
+- Para varios bloques a la vez, lee antes [varios_bloques](references/varios_bloques.md): su orden. Si una orden no
+  termina con `0`, lee [lecturas_que_fallan](references/lecturas_que_fallan.md) antes de seguir: qué hacer.
 
 ### 3. Completar
 
@@ -70,8 +73,6 @@
   cita, sus avisos y las líneas `⚠` de cada bloque: escribe como quien tiene el BOE delante,
   «el artículo no lleva avisos de vigencia» y no «la lectura no trajo avisos», la norma por su nombre e identificador,
   sin decir de dónde sale, qué has hecho o vas a hacer ni qué ha devuelto o no una orden: no le sirve.
-- **Nada de otra conversación**, ni para afirmar, confirmar o desmentir: de ellas solo va la línea
-  `⚠ REDACCIÓN MODIFICADA:` de cada bloque que la trae.
 - Cada afirmación sobre el contenido de una norma lleva su cita, de lo que `kitlegal boe` devolvió en esta
   conversación.
 - **Ley y reglamento**: di el `rango` de cada una y que la ley prevalece sobre su reglamento.
@@ -105,8 +106,8 @@
 ## Cómo se cita
 
 La forma legible y, en la misma línea, entre corchetes, el identificador y el id como los da `kitlegal boe`, y nada
-más: `art. 140 de la Constitución Española [BOE-A-1978-31229, bloque a140]`. Una cita por bloque, también del remitido;
-repásalas, y las dudas, en [como_se_cita](references/como_se_cita.md). Cada línea `⚠` va entera y con los datos de la lectura.
+más: `art. 140 de la Constitución Española [BOE-A-1978-31229, bloque a140]`. Una cita por bloque, también del remitido.
+Cada línea `⚠` va entera y con los datos de la lectura.
 
 ## Redacción en una fecha
 
@@ -121,9 +122,6 @@
 
 ## Redacción modificada
 
-Por cada entrada `version-obsoleta` de `kitlegal graph check`, y si se pregunta qué cambió, lee
-[redaccion_modificada](references/redaccion_modificada.md):
-
 ```text
 ⚠ REDACCIÓN MODIFICADA: <forma legible> [<identificador>, bloque <id>]: la redacción con fecha de vigencia AAAAMMDD, la que se consultó antes, ha sido sustituida por la de AAAAMMDD, que es la que se cita.
 ```
@@ -158,7 +156,7 @@
 ## Reglas
 
 1. **No concluir que algo no existe.** Una búsqueda vacía, o una norma fuera de la tabla, no prueba que no exista.
-2. **Nunca inventar contenido legal.** Si una lectura falla, di qué bloques faltan y por qué. Lo que lo leído no da
+2. **Nunca inventar contenido legal.** Lo que lo leído no da
    —un plazo, un órgano, una vía— no está cubierto: dilo, sin suplirlo de memoria ni con el procedimiento común.
 3. **Trasladar la vigencia.** Cada aviso, con su frase o una explicación: `⚠ NORMA DEROGADA:`, `⚠ VIGENCIA AGOTADA:`
    o `⚠ TEXTO POSIBLEMENTE DESACTUALIZADO:`, y la fecha de vigencia del bloque que no encaja con lo preguntado; una
```

Cada tramo es una remisión salvo el quinto y el último, que son los movimientos de §3; la remisión de
`redaccion_modificada.md` sale de la sección «Redacción modificada» (tramo octavo) y entra en el cuarto, y la de
`como_se_cita.md` sale de «Cómo se cita» (séptimo) y entra en el cuarto.

| # | Referencia | Sitio | Condición (FR-002) | Lo que trae | Respecto a la #245 |
|---|---|---|---|---|---|
| R1 | `con_herramientas.md` | «Protocolo», antes de la primera orden o llamada | ninguna | cómo se pide cada una con herramienta | igual |
| R2 | `normas.md` | paso 1, antes de toda orden | ninguna | su tabla, con el `BOE-A-…` | igual |
| R2 | `buscar_la_norma.md` | paso 1, antes de `boe buscar` | la norma no está en la tabla | cómo buscarla | igual |
| R3 | `redaccion_en_una_fecha.md` | paso 2, antes de leer el bloque | una fecha en la pregunta | cómo se lee | igual |
| R4 | `redaccion_modificada.md` | paso 2, detrás de la comprobación que trae la entrada | cada entrada `version-obsoleta`, **o** se pregunta qué cambió | qué se dice | «y» → «o» (D1) |
| R5 | `como_se_cita.md` | paso 2, detrás de la lectura que trae el caso | avisos o un id que termina en punto | cómo se citan | igual |
| R6 | `varios_bloques.md` | paso 2, antes de leer varios bloques | varios bloques a la vez | su orden | igual |
| R6 | `lecturas_que_fallan.md` | paso 2, antes de seguir | una orden que no termina con `0` | qué hacer | igual |

Ninguna lleva «si la tienes», «antes de escribir la respuesta» ni «antes de dar la consulta»; «si la tienes» sigue en la
frase que dice cómo pedir cada operación, que no enlaza nada. Las cinco formas fijas siguen en sus bloques `text`, y los
pasos 3 y 6 y la primera viñeta del paso 4 no cambian (`h32-normativa.txtar`; research V4).

## 3. Los movimientos de `normativa` (FR-006)

Exactamente dos, los de la #245 (su contrato, §3), que llegan enteros y palabra por palabra, al final de la referencia
que el cuerpo manda leer en su caso, detrás de una línea en blanco; solo cambian los saltos de línea:

| # | De | A | Texto |
|---|---|---|---|
| M-a | paso 4, su segunda viñeta | `references/redaccion_modificada.md` (R4) | «- **Nada de otra conversación**, ni para afirmar, confirmar o desmentir: de ellas solo va la línea `⚠ REDACCIÓN MODIFICADA:` de cada bloque que la trae.» |
| M-b | regla 2, tras su enunciado | `references/lecturas_que_fallan.md` (R6) | «Si una lectura falla, di qué bloques faltan y por qué.» |

```diff
--- control/references/redaccion_modificada.md
+++ skills/normativa/references/redaccion_modificada.md
@@ -42,3 +42,6 @@
 ## Si se pregunta qué cambió
 
 La respuesta dice que cita la redacción vigente y que la de antes no la ha leído.
+
+- **Nada de otra conversación**, ni para afirmar, confirmar o desmentir: de ellas solo va la
+  línea `⚠ REDACCIÓN MODIFICADA:` de cada bloque que la trae.
--- control/references/lecturas_que_fallan.md
+++ skills/normativa/references/lecturas_que_fallan.md
@@ -50,3 +50,5 @@
 
 La frase es solo de ese caso: con `0`, la respuesta no la lleva, ni afirmada, ni negada, ni con otras palabras. Si lo
 que falla es la lectura, la orden termina ahí, sin `kitlegal graph check`, y vale la regla 2.
+
+Si una lectura falla, di qué bloques faltan y por qué.
```

M-a se parte tras «la» y no tras «línea»: el control 7 de H31 no deja que una línea de una referencia a mano empiece por
`⚠` sin ser una de `SKILL.md`. Las referencias pasan de 2359 a 2520 bytes y de 2776 a 2834 (research V5), lejos de las
100 líneas a partir de las que `skills-check` pide un índice. Ningún otro byte de ninguna referencia cambia.

## 4. `jurisprudencia` (FR-005)

Los de la #245 (su contrato, §5), sin cambios:

```diff
--- control/SKILL.md
+++ skills/jurisprudencia/SKILL.md
@@ -57,8 +57,9 @@
 ⚠ SENTENCIA NO COMPROBADA: STS 1088/2023, de 4 de julio
 ```
 
-Y debajo, la consulta para encontrarla, que sale de `cita_preparar` o de su orden. Lo que se pide depende de cómo se
-dio la referencia:
+Y debajo, la consulta para encontrarla, que sale de `cita_preparar` o de su orden. Antes de pedir una u otra, lee
+[references/consulta_preparada.md](references/consulta_preparada.md): la orden de cada forma, cómo se dan las casillas
+y qué hacer con el equivalente que devuelva la operación. Lo que se pide depende de cómo se dio la referencia:
 
 | Cómo se dio | Qué se pide |
 |---|---|
@@ -75,11 +76,12 @@
 - **La respuesta no dice que la sentencia existe ni que no existe**: nadie lo ha comprobado. Tampoco dice de qué
   trata, qué resolvió ni si es firme (paso 5).
 
-Antes de dar la consulta, lee [references/consulta_preparada.md](references/consulta_preparada.md): la orden de cada
-forma, cómo se dan las casillas y qué hacer con el equivalente que devuelva la operación.
-
 ### 3. Con el documento delante: cotejar su ficha
 
+Con cada documento, antes del cotejo, lee [references/documento_aportado.md](references/documento_aportado.md): qué
+se pasa del documento y cómo, y qué se dice si no es la pedida, si su ROJ y su ECLI no se corresponden o si no lleva
+ficha.
+
 - **El cotejo es lo primero** que se hace con un documento que está en la conversación, sea su texto pegado o un PDF
   que has leído: su ficha va a `cita_cotejar`, o a su orden, antes de escribir ninguna frase sobre él.
 - **Nada del documento antes del cotejo**: ni que es la sentencia pedida, ni que coincide con ella, ni de qué trata.
@@ -90,10 +92,6 @@
 - **Si el documento no es el pedido, no lo cites como si fuera el pedido.** La sentencia pedida sigue sin estar
   delante: va por el paso 2, con su línea y su consulta, preparada en la forma en que se pidió.
 
-Con cada documento, antes del cotejo, lee [references/documento_aportado.md](references/documento_aportado.md): qué
-se pasa del documento y cómo, y qué se dice si no es la pedida, si su ROJ y su ECLI no se corresponden o si no lleva
-ficha.
-
 ### 4. Citar
 
 La cita tiene una forma fija: las siglas, el número de resolución y la fecha, y entre corchetes, en la misma línea,
```

| # | Referencia | Sitio | Condición | Lo que trae |
|---|---|---|---|---|
| J1 | `consulta_preparada.md` | paso 2, en la frase que nombra `cita_preparar`, antes de la tabla de lo que se pide | una sentencia que no está delante (el paso); antes de pedir la operación | la orden de cada forma, las casillas y el equivalente |
| J2 | `documento_aportado.md` | paso 3, su primer párrafo, antes de la viñeta que pide el cotejo | cada documento; antes del cotejo | qué se pasa del documento y qué se dice en cada caso |
| J3 | `consulta_preparada.md` | paso 6, sin cambios | una pregunta por materia; antes de `cita_preparar` | «La búsqueda por materia» |

## 5. La lista de remisiones (FR-001)

Las frases con un enlace a una referencia, fuera de las regiones que genera `make skills-sync`, en el orden del cuerpo,
tal como las da `skills.Frases` (§6): con la marca de su viñeta y los blancos juntados en un espacio. Es lo que
`TestRemisionesDelRepositorio` compara, entero y en orden (§7):

**`normativa`**, siete frases para sus ocho referencias, una por línea:

```text
Antes de la primera orden o llamada, lee [con_herramientas](references/con_herramientas.md): cómo se pide cada una con herramienta.
Lee [normas](references/normas.md) y toma el `BOE-A-…` de su tabla; si no está, lee [buscar_la_norma](references/buscar_la_norma.md) y búscala como dice.
- Con una fecha en la pregunta, lee [redaccion_en_una_fecha](references/redaccion_en_una_fecha.md) antes de leer el bloque: cómo se lee.
Por cada entrada `version-obsoleta`, o si se pregunta qué cambió, lee [redaccion_modificada](references/redaccion_modificada.md): qué se dice.
- Con avisos o un id que termina en punto, lee [como_se_cita](references/como_se_cita.md): cómo se citan.
- Para varios bloques a la vez, lee antes [varios_bloques](references/varios_bloques.md): su orden.
Si una orden no termina con `0`, lee [lecturas_que_fallan](references/lecturas_que_fallan.md) antes de seguir: qué hacer.
```

**`jurisprudencia`**, tres frases para sus dos referencias:

```text
Antes de pedir una u otra, lee [references/consulta_preparada.md](references/consulta_preparada.md): la orden de cada forma, cómo se dan las casillas y qué hacer con el equivalente que devuelva la operación.
Con cada documento, antes del cotejo, lee [references/documento_aportado.md](references/documento_aportado.md): qué se pasa del documento y cómo, y qué se dice si no es la pedida, si su ROJ y su ECLI no se corresponden o si no lleva ficha.
Ante «qué dice la jurisprudencia sobre…», lee «La búsqueda por materia» en [references/consulta_preparada.md](references/consulta_preparada.md), pide `cita_preparar` con los términos de la pregunta en `texto`, o su orden, y da la dirección que devuelve.
```

## 6. Qué es una frase: `skills.Frases` (FR-001, FR-023)

`func Frases(cuerpo string) []string`, en `internal/skills/frases.go` (research D5, D6). Los segmentos del cuerpo, en
su orden:

1. Cada línea de un bloque de código —de la valla que lo abre a la que lo cierra, como `lineasFueraDeBloques`
   (`internal/skills/markdown.go:51`), vallas incluidas— es un segmento, tal cual, sin su salto.
2. Cada línea de encabezado (tras su sangría, empieza por `#`) y cada línea que, sin sus blancos de los extremos, es
   entera un comentario HTML (`<!--` … `-->`) es un segmento, sin sus blancos de los extremos.
3. En lo demás, cada **frase**: empieza en el primer carácter que no es un blanco tras el final de la anterior, una
   línea en blanco, un segmento de 1 o 2, o el principio del cuerpo; también al principio de una línea que, tras su
   sangría, empieza por `- `, `* ` o un número y `. ` (la marca abre la frase y va en ella). Acaba en el primer punto
   seguido de un espacio o de un salto de línea, o al final del cuerpo, o donde empieza otra frase o un segmento de 1
   o 2. El punto de `.md` dentro de un enlace no la cierra, porque no lo sigue un blanco.
4. Los blancos y saltos de línea de una frase se juntan en un espacio; los de sus extremos se quitan. Las líneas en
   blanco no dan segmento.

`TestFrases` (`internal/skills/frases_test.go`, de tabla) fija cada regla con un cuerpo escrito en el test: dos frases en
una línea; una frase partida en tres líneas con sangría; una viñeta y un elemento numerado que abren frase; un
encabezado y un comentario HTML, segmentos solos; un bloque de código que corta una frase y cuyas líneas son segmentos;
un enlace `[x.md](references/x.md)` seguido de `:`, que no cierra la frase.

## 7. La regla de las remisiones: `TestRemisionesDelRepositorio` (FR-001 a FR-003, FR-010; SC-001)

En `internal/skills/remisiones_test.go`, con `raizDelRepositorio = "../.."` como los demás tests del repositorio del
paquete. Por cada skill de `skills/` con `references/`, en orden de nombre, salvo las de la tabla
`fueraDeLasRemisiones = map[string]string{"plazos-y-territorio": "H41.1"}` (FR-010; el valor es el hito que la incluye,
y el comentario de la tabla lo dice):

1. **La regla.** Por cada segmento de `skills.Frases` del cuerpo con un enlace `[…](references/<f>)` o
   `[…](references/<f>#…)`, y por cada uno de esos enlaces, falla, una vez por referencia y motivo, en orden de nombre
   de las referencias, con `<skill>: references/<f>: <motivo>`:
   - `un enlace suyo no va en una frase que lo mande leer («lee» antes del enlace)` — si antes del enlace, en su frase,
     no está la palabra entera `lee` o `Lee`;
   - `la frase de un enlace suyo lleva «<expresión>»` — por cada una de `si la tienes`, `antes de escribir la respuesta`
     y `antes de dar la consulta` que esté en la frase, sin distinguir mayúsculas de minúsculas;
   - `ningún enlace del cuerpo la manda leer` — si el cuerpo no tiene ningún enlace a ella.
2. **La lista.** Los segmentos con un enlace a una referencia que no están entre los segmentos `<!-- paquetes: …` y
   `<!-- fin de los paquetes -->` son, en orden, los de `remisionesDelPlan[<skill>]` (§5). Si no, falla con
   `<skill>: la remisión <n> no es la del plan` y las dos frases (`assert.Equal` de las dos listas), o con `<skill>: no
   tiene lista de remisiones del plan` si la skill no está en el mapa ni en `fueraDeLasRemisiones`.
3. **Subtest «control».** La regla de 1 (sin la lista) sobre el `SKILL.md` de `normativa-control` y de
   `jurisprudencia-control` da exactamente estos fallos, así que el control se ve fallar por encima del umbral:
   - `normativa`: dos en `con_herramientas.md` (el verbo y «si la tienes») y el del verbo en `buscar_la_norma.md`,
     `como_se_cita.md`, `lecturas_que_fallan.md`, `normas.md`, `redaccion_en_una_fecha.md` y `varios_bloques.md`;
   - `jurisprudencia`: «antes de dar la consulta» en `consulta_preparada.md` (la frase empieza por «Antes»).

`make skills-check` gana el nombre del test en su expresión de `-run`, junto a `TestNormasDelRepositorio`; ninguna otra
receta cambia. Lo que da, con el `SKILL.md` de `normativa-control` (dos líneas de 122 y 93 bytes con su salto):

```text
normativa: references/con_herramientas.md: un enlace suyo no va en una frase que lo mande leer («lee» antes del enlace)
normativa: references/con_herramientas.md: la frase de un enlace suyo lleva «si la tienes»
```

## 8. El producto desde control: `TestSkillsDelProductoDesdeControl` (FR-006, FR-007, FR-009, FR-010; SC-003 a SC-005)

En `internal/evals/laboratorio_familias_test.go`, con la declaración de data-model §5 escrita con los bytes de §2 a §4:

1. **`normativa` y `jurisprudencia`.** Lee control y `skills/<skill>` sin lo que genera `make skills-sync` —sin las
   referencias cuya primera línea empieza por `<!-- generado desde `, y con las regiones del `SKILL.md` entre
   `<!-- paquetes: …` y `<!-- fin de los paquetes -->` y entre `<!-- inicio de la tabla de comandos…` y `<!-- fin de la
   tabla de comandos -->` reducidas a sus marcas—; aplica al `SKILL.md` de control cada cambio, cuyo `antes` tiene que
   estar una sola vez, y añade a cada referencia el `enLaReferencia` de sus movimientos; y exige el resultado en
   `skills/<skill>`, ruta a ruta y byte a byte. Falla con la ruta del primer fichero que difiere, falta o sobra, o con
   el cambio cuyo `antes` ya no está una vez.
2. **Cada movimiento**: su texto y su `enLaReferencia`, con los blancos juntados, son iguales, el primero no está en el
   `SKILL.md` de `skills/<skill>` y el segundo está en su referencia.
3. **El frontmatter** del `SKILL.md` de `skills/<skill>`, de la primera línea a la que lo cierra, es el de control byte
   a byte.
4. **`plazos-y-territorio`** (FR-010): `skills/plazos-y-territorio` y `plazos-y-territorio-control` tienen los mismos
   ficheros, ruta a ruta y byte a byte, **también lo generado**; falla con la ruta del que difiere, falta o sobra. Es
   así hasta H41.1, como dice el comentario del subtest (research D4).

Que lo generado sea lo que da `make skills-sync` lo sigue comprobando `TestSkillsDelRepositorio` (FR-009), y el tope de
11 000 caracteres, el mismo test (FR-008).

## 9. Uso, de fuera adentro

- **El cuerpo.** Lo lee el agente una vez por pregunta, al activar la skill: 10 926 y 10 842 caracteres (§1), nunca más
  de 11 000, tampoco con la línea de un paquete. No crece con el uso: cientos de normas y miles de bloques consultados no
  lo tocan; la región de los paquetes crece con los paquetes declarados en `data/`, no con las consultas.
- **Las referencias**, cada una una vez por pregunta, cuando llega su condición. `normativa`, en toda pregunta:
  `con_herramientas.md` (2442 bytes) y `normas.md` (3510), 5952 bytes, que hoy en modo herramienta casi nunca se leen y
  es lo que el hito quiere que se lea (research S1); con su condición, de 909 (`varios_bloques.md`) a 2834
  (`lecturas_que_fallan.md`, con lo movido), y `redaccion_modificada.md`, 2520 con lo movido. Una pregunta con fecha:
  cuerpo + 5952 + 2809 bytes. `jurisprudencia`: `consulta_preparada.md` (2038) con una sentencia que no está delante o
  una pregunta por materia, `documento_aportado.md` (2805) con un documento; una vez por pregunta aunque lleguen varios.
  `normas.md` crece con las normas declaradas en `data/normas.yaml`, no con las consultadas; las demás, solo cuando un
  hito las cambia. Lo movido se lee solo en su caso, en vez de en toda pregunta con el cuerpo.
- **Las invocaciones de `kitlegal` por pregunta** no cambian: el hito no añade ni quita ninguna orden ni llamada, solo
  lecturas de referencias, una por referencia y pregunta.
- **Señales.** Ninguna: el contenido de cuerpo y referencias no depende de la consulta ni caduca. Los fallos de §7 los
  lee quien cambia una skill, una vez por `make skills-check` o `make ci`: como mucho tres líneas por referencia y la
  diferencia de la lista; con el `SKILL.md` de control, ocho líneas en `normativa`, unos 900 bytes. No dependen del uso
  del producto y dejan de darse cuando cada remisión es la del plan.
