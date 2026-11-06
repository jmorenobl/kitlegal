# Contrato: las remisiones de las tres skills y lo que se mueve

FR-001 a FR-010; SC-001 a SC-004. Research D1 a D7, V2 a V6; data-model §5.

## 1. Términos

- **Cuerpo**: lo que sigue a la línea `---` que cierra el frontmatter de un `SKILL.md`, contado como runas (research V2).
- **Control**: los ficheros de `laboratorio/<skill>/<skill>-control/skills/<skill>/`, que son los de `skills/<skill>` en
  `v0.9.0` (research V1).
- **Remisión**: la frase del cuerpo que enlaza `references/<f>` (§5).
- Los cambios van como `diff -u` del `SKILL.md` de control al de `skills/<skill>` tras el hito. Cada `-` y cada `+` es
  una línea entera del fichero, con su sangría; nada fuera de estos tramos cambia. Fuera de ellos, el frontmatter, la
  tabla de comandos y el bloque de los paquetes son los de control, byte a byte (FR-010).

| Skill | Cuerpo de control | Cuerpo tras el hito | Referencias | Movimientos |
|---|---|---|---|---|
| `normativa` | 10 929 | **10 926** | las ocho de hoy | dos (§3) |
| `plazos-y-territorio` | 10 930 | **10 982** | las cuatro de hoy | ninguno |
| `jurisprudencia` | 10 842 | **10 842** | las dos de hoy | ninguno |

Medidas en esta sesión sobre copias con estos cambios (research V3). Ninguna referencia nueva.

**Corrección de T003** (gates/supuestos.md): la vertical de ensayo de H32 (`TestVerticalDeEnsayo`) añade 70 caracteres
al bloque de los paquetes de las tres skills, así que cada cuerpo, sin paquetes, cabe en 10 930 como mucho. Con los
10 981 de arriba, `normativa` daba 11 051: su segundo movimiento (§3) la deja en 10 926. El 10 982 de
`plazos-y-territorio` tampoco cabe con la vertical de ensayo.

## 2. `normativa` (FR-004)

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
+  fecha van siempre. Por cada entrada `version-obsoleta`, y si se pregunta qué cambió, lee
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
```

Cada tramo es una **remisión** salvo el quinto (`@@ -70,8 +73,6 @@`), que es el **movimiento** de §3; R4 y R5 salen
de los tramos sexto y séptimo y entran en el cuarto:

| # | Referencia | Sitio | Condición (FR-002) | Lo que trae | Qué hace con control |
|---|---|---|---|---|---|
| R1 | `con_herramientas.md` | «Protocolo», antes de la primera orden o llamada | ninguna | cómo se pide cada una con herramienta | quita «tras leer», que sigue a «si la tienes», y pone la orden delante |
| R2 | `normas.md` | paso 1, antes de toda orden | ninguna | su tabla, con el `BOE-A-…` | la mención dentro de «toma … de la tabla de» pasa a orden (la de `remisiones-en-orden`) |
| R2 | `buscar_la_norma.md` | paso 1, antes de `boe buscar` | la norma no está en la tabla | cómo buscarla («búscala como dice») | ídem |
| R3 | `redaccion_en_una_fecha.md` | paso 2, antes de leer el bloque | una fecha en la pregunta | cómo se lee | la remisión entre rayas pasa a orden con su condición |
| R4 | `redaccion_modificada.md` | paso 2, tras la comprobación que trae la entrada | cada entrada `version-obsoleta`, o se pregunta qué cambió | qué se dice | sale de la sección «Redacción modificada» (no es un paso) al paso 2; la sección conserva encabezado y forma |
| R5 | `como_se_cita.md` | paso 2, tras la lectura que trae el caso | un bloque con avisos o un id que termina en punto | cómo se citan | «repásalas, y las dudas, en» de «Cómo se cita» sale, y la orden va al paso 2 (research D1) |
| R6 | `varios_bloques.md` | paso 2, antes de leer varios bloques | varios bloques a la vez | su orden | mención tras dos puntos pasa a orden |
| R6 | `lecturas_que_fallan.md` | paso 2, antes de seguir | una orden que no termina con `0` | qué hacer | ídem |

Ninguna lleva «si la tienes», «antes de escribir la respuesta» ni «antes de dar la consulta». Lo que no es un enlace no
cambia: «si la tienes» sigue en la frase que dice cómo pedir cada operación, que no remite a nada. Las cinco formas fijas
siguen en sus bloques `text`, y los pasos 3 y 6 y la primera viñeta del paso 4, línea a línea, como exige
`h32-normativa.txtar` (research V4).

## 3. El movimiento de `normativa` (FR-009)

| De | A | Texto | Por qué |
|---|---|---|---|
| el paso 4, su segunda viñeta | el final de `references/redaccion_modificada.md`, detrás de una línea en blanco | «Nada de otra conversación» (abajo) | lo único que llega de otra conversación a una respuesta es una entrada `version-obsoleta`, y es el caso en que R4 manda leer esa referencia; sin entrada lo cubre la frase del paso 2 que no se mueve (research D2) |

Sale del `SKILL.md` (las dos líneas `-` del quinto tramo de §2) y llega así como sus tres últimas líneas, partida tras
«la» y no tras «línea»: el control 7 de H31 no deja que una línea de una referencia a mano empiece por `⚠` sin ser una
de `SKILL.md`, y FR-009 deja cambiar los saltos de línea (corrección de T003):

```diff
--- control/references/redaccion_modificada.md
+++ skills/normativa/references/redaccion_modificada.md
@@ -42,3 +42,6 @@
 ## Si se pregunta qué cambió
 
 La respuesta dice que cita la redacción vigente y que la de antes no la ha leído.
+
+- **Nada de otra conversación**, ni para afirmar, confirmar o desmentir: de ellas solo va la
+  línea `⚠ REDACCIÓN MODIFICADA:` de cada bloque que la trae.
```

La referencia pasa de 2359 a 2520 bytes y de 44 a 47 líneas, lejos de las 100 a partir de las que `skills-check` le
pide un índice.

El segundo movimiento (corrección de T003; §1):

| De | A | Texto | Por qué |
|---|---|---|---|
| la regla 2, tras su enunciado | el final de `references/lecturas_que_fallan.md`, detrás de una línea en blanco | «Si una lectura falla, di qué bloques faltan y por qué.» | es el caso en que R6 manda leer esa referencia («Si una orden no termina con `0`»), que ya lo dice en «Qué dice la respuesta»; la regla 2 conserva su enunciado y lo no cubierto |

```diff
--- control/SKILL.md
+++ skills/normativa/SKILL.md
@@ -161 +159 @@
-2. **Nunca inventar contenido legal.** Si una lectura falla, di qué bloques faltan y por qué. Lo que lo leído no da
+2. **Nunca inventar contenido legal.** Lo que lo leído no da
--- control/references/lecturas_que_fallan.md
+++ skills/normativa/references/lecturas_que_fallan.md
@@ -51,2 +51,4 @@
 La frase es solo de ese caso: con `0`, la respuesta no la lleva, ni afirmada, ni negada, ni con otras palabras. Si lo
 que falla es la lectura, la orden termina ahí, sin `kitlegal graph check`, y vale la regla 2.
+
+Si una lectura falla, di qué bloques faltan y por qué.
```

La referencia pasa de 2776 a 2834 bytes y de 52 a 54 líneas. Ningún otro byte de ninguna referencia de las tres skills
cambia.

## 4. `plazos-y-territorio` (FR-005)

```diff
--- control/SKILL.md
+++ skills/plazos-y-territorio/SKILL.md
@@ -27,12 +27,12 @@
 <!-- paquetes: lo genera make skills-sync -->
 - Si la pregunta es de lo administrativo, lee [administrativo](references/administrativo.md).
 <!-- fin de los paquetes -->
-- [jerarquia_normativa.md](references/jerarquia_normativa.md): para el nivel que regula algo, con el autonómico de la
-  `comunidad` de `data`, y la norma que prevalece.
-- [municipios.md](references/municipios.md): al presentar un territorio, si falta el municipio o no sale con 0, y
-  para explicar su cobertura.
-- [calculo_de_plazos.md](references/calculo_de_plazos.md): en los municipios que cuentan, un plazo sin calcular y una
-  lectura fallida o con `redaccion_posterior`.
+- Para el nivel que regula algo, con el autonómico de la `comunidad` de `data`, y la norma que prevalece, lee
+  [jerarquia_normativa.md](references/jerarquia_normativa.md): rangos y reglas.
+- Al presentar un territorio, si falta el municipio o no sale con 0, y para explicar su cobertura, lee
+  [municipios.md](references/municipios.md): qué hacer.
+- En los municipios que cuentan, un plazo sin calcular y una lectura fallida o con `redaccion_posterior`, lee
+  [calculo_de_plazos.md](references/calculo_de_plazos.md): su detalle.
 
 ## Protocolo
 
```

| # | Referencia | Sitio | Condición | Lo que trae |
|---|---|---|---|---|
| P1 | `jerarquia_normativa.md` | «Referencias», antes del «Protocolo» y de toda llamada | el nivel que regula algo y la norma que prevalece | rangos y reglas |
| P2 | `municipios.md` | ídem | presentar un territorio; un municipio que falta o no sale con 0; su cobertura | qué hacer |
| P3 | `calculo_de_plazos.md` | ídem | los municipios que cuentan; un plazo sin calcular; una lectura fallida o con `redaccion_posterior` | su detalle |

La orden generada de `administrativo.md`, entre sus marcas, y la línea que la introduce no cambian (FR-005), y siguen
donde se abren en 58 de 60 sesiones en modo herramienta (research D3).

## 5. `jurisprudencia` (FR-006)

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
| J1 | `consulta_preparada.md` | paso 2, en la frase que nombra `cita_preparar`, antes de la tabla de lo que se pide | una sentencia que no está delante; antes de pedir la operación | la orden de cada forma, las casillas y el equivalente |
| J2 | `documento_aportado.md` | paso 3, su primer párrafo, antes de la viñeta que pide el cotejo | cada documento; antes del cotejo | qué se pasa del documento y qué se dice en cada caso |
| J3 | `consulta_preparada.md`, «La búsqueda por materia» | paso 6, sin cambios | una pregunta por materia; antes de `cita_preparar` | ese apartado |

## 6. La regla de las remisiones (FR-001 a FR-003, SC-001)

`TestRemisionesDelRepositorio` (`internal/skills/remisiones_test.go`, en el paquete, con `raizDelRepositorio = "../.."`;
research D5), sobre el `SKILL.md` de cada skill de `skills/` que tiene `references/`, tal como está en el árbol (con sus
regiones generadas, que `TestSkillsDelRepositorio` mantiene iguales a lo que da `make skills-sync`):

1. Las líneas de los bloques de código cuentan como líneas en blanco (`lineasFueraDeBloques`; research D6).
2. Por cada enlace `[…](references/<f>)` o `[…](references/<f>#…)` del cuerpo, su **frase** empieza tras el último
   límite anterior al enlace —un punto seguido de un espacio o de un salto de línea, una línea en blanco, o el principio
   de una línea que, tras su sangría, empieza por `- `, `* `, un número y `. `, o `#`— y acaba en el primer punto
   seguido de un espacio o de un salto de línea, la primera línea en blanco o el final del cuerpo que haya tras el
   enlace. El enlace entero, con el punto de su `.md`, queda dentro.
3. Falla, una vez por referencia y regla, en el orden de los nombres de las referencias, con
   `<skill>: references/<f>: <regla>`:
   - `un enlace suyo no va en una frase que lo mande leer («lee» antes del enlace)` — si entre el principio de la frase y
     el enlace no está la palabra `lee` o `Lee` (entera: ni `leer` ni `léelo`);
   - `la frase de un enlace suyo lleva «<expresión>»` — por cada una de `si la tienes`, `antes de escribir la respuesta`
     y `antes de dar la consulta` que esté en la frase, con sus blancos y saltos de línea juntados en un espacio y sin
     distinguir mayúsculas de minúsculas, como `lee` o `Lee` en el verbo: «Antes de dar la consulta, lee […]», al
     principio de una frase, falla igual que en medio; el mensaje nombra la expresión tal como está en esta lista;
   - `ningún enlace del cuerpo la manda leer` — si el cuerpo no tiene ningún enlace a ella.
4. Un subtest, **«control»**, aplica la misma regla al `SKILL.md` de cada `<skill>-control`
   (`laboratorio/<skill>/<skill>-control/skills/<skill>/`) y exige exactamente estos fallos: en `normativa`, dos en
   `con_herramientas.md` (el verbo y «si la tienes») y uno en `normas.md`, `buscar_la_norma.md`,
   `redaccion_en_una_fecha.md`, `varios_bloques.md`, `lecturas_que_fallan.md` y `como_se_cita.md`; en
   `plazos-y-territorio`, el del verbo en cada una de las tres de §4; en `jurisprudencia`, el de «antes de dar la
   consulta» en `consulta_preparada.md`. Otro, **«frases»**, dos cuerpos escritos en el test con lo que el cuerpo real
   tiene: una orden de leer escrita antes de un bloque de código para un enlace escrito después, que falla, y «si la» y
   «tienes» en dos líneas de la frase de un enlace, que falla. Así el control se ve fallar por encima del umbral.

`make skills-check` gana el nombre del test en su expresión, junto a `TestNormasDelRepositorio` y los demás; ninguna otra
receta cambia. Ejemplo de lo que da con el párrafo del «Protocolo» de control (dos líneas, 122 y 93 bytes con su salto):

```text
normativa: references/con_herramientas.md: un enlace suyo no va en una frase que lo mande leer («lee» antes del enlace)
normativa: references/con_herramientas.md: la frase de un enlace suyo lleva «si la tienes»
```

Lo lee quien cambia una skill, una vez por `make skills-check` o `make ci`: como mucho cinco líneas por referencia; con el
`SKILL.md` de control, ocho en `normativa`, unos 900 bytes. No depende del uso del producto y deja de darse cuando cada
remisión es una orden. Con lo de §2 a §5, ninguna.

## 7. El producto se calcula desde control (FR-007, FR-009, FR-010; SC-003, SC-004)

`TestSkillsDelProductoDesdeControl` (`internal/evals/laboratorio_familias_test.go`), por skill, con la declaración de
data-model §5 escrita con los bytes de §2 a §5:

1. Lee control y `skills/<skill>`, sin lo que genera `make skills-sync`: sin las referencias cuya primera línea empieza
   por `<!-- generado desde `, y con las regiones del `SKILL.md` entre `<!-- paquetes: …` y `<!-- fin de los paquetes -->`
   y entre `<!-- inicio de la tabla de comandos…` y `<!-- fin de la tabla de comandos -->` reducidas a sus marcas
   (research D7).
2. Aplica al `SKILL.md` de control cada cambio, cuyo `antes` tiene que estar una sola vez; y añade a cada referencia el
   `enLaReferencia` de sus movimientos.
3. Exige, ruta a ruta y byte a byte, el resultado en `skills/<skill>`; falla con la ruta del primer fichero que difiere,
   falta o sobra, o con el cambio cuyo `antes` ya no está una vez.
4. Por cada movimiento: su texto y su `enLaReferencia`, con los blancos juntados y `references/<x>` escrito `<x>`, son
   iguales, y el texto no está en el `SKILL.md` de `skills/<skill>`.
5. El frontmatter del `SKILL.md` de `skills/<skill>`, desde la primera línea a la que lo cierra, es el de control byte a
   byte (FR-010).

Que lo generado sea lo que da `make skills-sync` lo sigue comprobando `TestSkillsDelRepositorio` (FR-010).

## 8. Uso, de fuera adentro

- **El cuerpo.** Lo lee el agente una vez por pregunta, al activar la skill: 10 926, 10 982 y 10 842 caracteres (§1),
  nunca más de 11 000 (FR-008), tampoco con la línea de un paquete (corrección de T003). No crece con el uso: cientos de normas y miles de bloques consultados no lo tocan; el
  bloque de los paquetes crece con los paquetes declarados en `data/`, no con las consultas.
- **Las referencias**, cada una una vez por pregunta, cuando llega su condición. `normativa`, en toda pregunta:
  `con_herramientas.md` (2442 bytes) y `normas.md` (3510): 5952 bytes, lo que hoy casi nunca se lee en modo herramienta
  y es lo que el hito quiere que se lea (research S1). Con su condición, de 909 (`varios_bloques.md`) a 2834
  (`lecturas_que_fallan.md`, con lo movido); `redaccion_modificada.md`, 2520 con lo movido, que deja de leerse en las preguntas sin
  entrada `version-obsoleta`. Una pregunta con fecha: cuerpo + 5952 + 2809 bytes. `plazos-y-territorio`: `administrativo.md`
  (9930) en toda pregunta de plazos o recursos, como hoy, y con su condición `municipios.md` (3944), `calculo_de_plazos.md`
  (2760) y `jerarquia_normativa.md` (1543). `jurisprudencia`: `consulta_preparada.md` (2038) con una sentencia que no
  está delante o una pregunta por materia, `documento_aportado.md` (2805) con un documento. `normas.md` crece con las
  normas declaradas en `data/normas.yaml`, no con las consultadas; las demás, solo cuando un hito las cambia.
- **Las invocaciones por pregunta** de la tabla de comandos no cambian: el hito no añade ni quita ninguna orden ni
  llamada a `kitlegal`, solo lecturas de referencias, una por referencia y pregunta.
- **Señales.** Ninguna: el contenido de cuerpo y referencias no depende de la consulta y no caduca. Los fallos de §6 los
  lee quien cambia una skill, en `make skills-check` o `make ci`, y dejan de darse cuando cada remisión es una orden.
