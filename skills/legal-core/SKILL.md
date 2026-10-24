---
name: legal-core
description: >-
  Punto de partida de las preguntas de derecho público español que dependen de un municipio o de un plazo. Se usa
  cuando se pregunta hasta cuándo se puede recurrir o qué plazo hay para un recurso, cuándo vence un plazo
  administrativo o el del contencioso-administrativo o se produce el silencio, qué comunidad, provincia o boletines
  corresponden a un ayuntamiento, si un municipio es foral o qué norma prevalece. Con kitlegal identifica el
  territorio de un municipio —código INE, provincia, comunidad, régimen, DIR3 del ayuntamiento y boletines— y calcula
  el último día del plazo con los festivos de los municipios que cuentan; dice qué asuntos regula aparte la ley
  (tributos, multas de tráfico, Seguridad Social, extranjería) y razona con la jerarquía normativa y las leyes
  vertebrales (CE, LPAC, LRJSP, LRBRL, TRLRHL, LCSP, LGT…). Para el texto de un artículo se apoya en boe-legislacion.
metadata:
  kitlegal-applets: territorio plazos
  kitlegal-referencias: leyes_vertebrales jerarquia_normativa recursos_y_plazos cobertura_por_materia
---

# Territorio, plazos, jerarquía normativa y leyes vertebrales

Esta skill identifica el territorio, calcula los plazos y orienta: no tramita nada y no sustituye el asesoramiento de
un profesional.

## Referencias

Cuándo leer cada una:

- [leyes_vertebrales.md](references/leyes_vertebrales.md): al nombrar una norma, su identificador y su rango.
- [jerarquia_normativa.md](references/jerarquia_normativa.md): para el nivel que regula algo, con el autonómico de la
  `comunidad` de `data`, y la norma que prevalece.
- [cobertura_por_materia.md](references/cobertura_por_materia.md): en toda pregunta de plazos o de recursos.
- [recursos_y_plazos.md](references/recursos_y_plazos.md): para elegir el tipo de plazo y nombrarlo.
- [municipios.md](references/municipios.md): al presentar un territorio, si falta el municipio o no sale con 0, y
  para explicar su cobertura.
- [calculo_de_plazos.md](references/calculo_de_plazos.md): en lo que la ley regula aparte, un acto presunto, un
  silencio, una reclamación de acceso, un plazo sin calcular y una lectura fallida o con `redaccion_posterior`.

## Protocolo

Territorio: del 1 al 4, y el 5 y el 6 si va más allá. Plazos o recursos: el 7 y, solo si hay plazo, del 8 al 11.

1. **El municipio**, de la conversación. Si no lo dice, **pregúntalo** y espera: no lo supongas, tampoco el de quien
   usa el kit.
2. **Resolverlo** con `territorio_resolver` o `kitlegal territorio resolver <nombre o código INE> --json`, tal como lo
   dio la persona. Sin herramienta ni binario, la regla 7.
3. **Responder con `data`**, la cobertura en su forma fija y de cuándo son los datos (`fecha_consulta`). **Ningún
   boletín que no esté en `boletines`.**
4. **Sin el 0**: con el 2 o el 3 de un municipio, lo que dice la referencia de los municipios; con otro código, di que
   no se pudo resolver, sin suplir lo que falte.
5. **Razonar** con las leyes vertebrales y la jerarquía normativa: identificador y rango de cada norma, nivel que
   regula lo preguntado y norma que prevalece.
6. **El texto de un artículo**, y el identificador que no está en las referencias, con `boe-legislacion` en esta
   conversación, citado como ella cita (regla 2).
7. **¿Hay plazo?** Con lo que ya dice la conversación del acto, la administración y la materia, **antes de preguntar
   nada**, ni el municipio ni la fecha. Si la materia está en la cobertura por materia, `⚠ NO CUBIERTO:`,
   **ni `plazos calcular`** ni ninguna fecha (regla 9). Si no dice la materia, **no la preguntes ni detengas la
   respuesta**: sigue con el 8, y al responder di que la fecha no vale si la ley regula el asunto aparte, nombrando
   esos asuntos solo con la referencia. Un acto presunto no tiene fecha: ningún plazo.
8. **Situar el plazo**: la fecha de inicio (notificación o publicación; en el silencio, acuerdo de iniciación o
   entrada en el registro), que si falta **se pregunta**; y los municipios que cuentan, con los pasos 1 a 4: **el del
   interesado y el de la sede del órgano**, o en el contencioso-administrativo **el de la sede del juzgado**.
9. **Pedir el plazo**: el tipo, de los recursos y plazos; `plazos_calcular` o `kitlegal plazos calcular <tipo> <inicio>
   <municipios>... --json`, con los municipios del 8 en su orden. **Con el 3 y un `data.mensaje` que empieza por «el
   plazo de tipo», no hay fecha**: `⚠ PLAZO SIN CALCULAR:`, sin la línea del plazo ni ninguna fecha, y sin el 10.
10. **Leer cada precepto de `data.preceptos` en la fecha de inicio**, con `boe_articulo` y `fecha` o
    `kitlegal boe articulo <norma> <bloque> --fecha <inicio> --json`, sin `kitlegal graph check`, y citarlo con
    `⚠ REDACCIÓN A FECHA:`. Si sale con 4 o 5, **`⚠ PLAZO:` se mantiene**: di aparte que no has podido leerlo, sin
    afirmar qué dice.
11. **Responder**: `data.fecha` en la línea del plazo y ninguna otra fecha (regla 8); los festivos que la movieron, de
    `data.festivos`, con fecha, ámbito y, si es local, municipio; cada aviso de `data.avisos`, su `texto` tal cual; el
    alcance de los festivos, solo si lo dice un aviso.

## Cómo se presenta el territorio

Cada dato como lo da `data`; un boletín, por su `codigo` y su `nombre` exactos; la cobertura, con la clave y el valor
exactos, una línea por aspecto, y detrás puedes explicarla:

```text
boletin_autonomico: <valor>
boletin_provincial: <valor>
dir3: <valor>
```

## Cómo se presenta un plazo

La línea del plazo, con `data.fecha` tal cual, el tipo como lo nombran los recursos y plazos y cada precepto de
`data.preceptos` en sus propios corchetes, con un solo bloque (dos en unos corchetes no son cita de ninguno); debajo,
la del contencioso:

```text
⚠ PLAZO: <AAAA-MM-DD> — <tipo> [<norma>, <precepto>, <identificador>, bloque <bloque>]
⚠ PLAZO: <AAAA-MM-DD> — recurso contencioso-administrativo [art. 46.1 de la LJCA, BOE-A-1998-16718, bloque a46] [art. 128.2 de la LJCA, BOE-A-1998-16718, bloque a128]
```

Detrás, la redacción a fecha de cada precepto leído en el paso 10 y, con `redaccion_posterior`, la de hoy:

```text
⚠ REDACCIÓN A FECHA: <forma legible> [<identificador>, bloque <bloque>]: se cita la redacción vigente el AAAA-MM-DD, la de fecha de vigencia AAAAMMDD.
⚠ REDACCIÓN DE HOY: <forma legible> [<identificador>, bloque <bloque>]: hoy la redacción es otra, la de <identificador>, con fecha de vigencia AAAAMMDD.
```

Con el 3 del paso 9, en lugar de la línea del plazo, con cada línea del mensaje:

```text
⚠ PLAZO SIN CALCULAR: <tipo> que empieza el <AAAA-MM-DD>: kitlegal no lo calcula, porque lo cuenta con preceptos que ese día no decían lo que dicen hoy: <identificador>, bloque <bloque>, apartado <apartado>, desde el <AAAA-MM-DD>; <…>.
```

Lo no cubierto (paso 7), con el procedimiento como lo nombra la cobertura por materia: la primera si la entrada trae
la norma que lo rige; si no, la segunda:

```text
⚠ NO CUBIERTO: <procedimiento> se rige por <norma> (<identificador>), no por el procedimiento común: kitlegal no calcula sus plazos ni sus recursos (<precepto de la LPAC>).
⚠ NO CUBIERTO: <procedimiento> lo regula aparte la <precepto de la LPAC> (BOE-A-2015-10565): kitlegal no calcula sus plazos ni sus recursos.
```

Los avisos de `data.avisos`, cada uno con su `texto`, que empieza así:

```text
⚠ FALTAN FESTIVOS NACIONALES Y AUTONÓMICOS: …
⚠ FALTAN FESTIVOS LOCALES: …
⚠ DÍAS DEL ART. 183 LOPJ: …
⚠ FESTIVO DE UNA PARTE: …
```

Un programa comprueba cada línea `⚠` sin leer nada más: dicha con otras palabras, falta. Va entera, en una sola línea
y con su forma; detrás puedes explicarla.

## Comandos

Con la herramienta de su fila si entre las tuyas hay una con ese nombre, sola o tras el prefijo de tu agente, como
`mcp__kitlegal__territorio_resolver`: si la tienes, úsala siempre. Si no, con la orden `kitlegal … --json`, desde el
`PATH`. Donde la orden sale con un código, la herramienta da un error con la `data.clase` entre paréntesis: 0
correcto, 2 (`argumentos`) argumentos inválidos —también un nombre de más de un municipio—, 3 (`no-encontrado`) un
municipio fuera de la relación o un plazo con preceptos que no regían el día de inicio.

<!-- inicio de la tabla de comandos: generada desde --describe con make skills-sync, no editar -->

### `kitlegal territorio`

| Orden | Herramienta | Qué hace | Qué devuelve en `data` |
|---|---|---|---|
| `kitlegal territorio resolver <consulta>` | `territorio_resolver` | Devuelve el territorio de un municipio, por su nombre o por su código INE, con la cobertura de lo que está configurado y verificado. | objeto con `municipio`, `codigo_ine`, `provincia`, `comunidad`, `dir3`, `regimen`, `boletines`, `cobertura` |

### `kitlegal plazos`

| Orden | Herramienta | Qué hace | Qué devuelve en `data` |
|---|---|---|---|
| `kitlegal plazos calcular <tipo> <inicio> <municipios>... [--duracion <duracion>] [--unidad <unidad>]` | `plazos_calcular` | Devuelve la fecha de un plazo —un recurso, el acceso a la información pública, el silencio o uno genérico— desde la fecha de inicio que se da, con los preceptos que lo fijan, los festivos saltados, los avisos y la cobertura de los festivos. No la da si alguno de sus preceptos no regía así el día de inicio. | objeto con `fecha`, `tipo`, `preceptos`, `municipios`, `festivos`, `avisos`, `cobertura` |

La orden y la herramienta de cada fila devuelven el mismo sobre: `ok`, `fuente`, `url`, `fecha_consulta`, `hash`, `data`; con `ok` falso, `data` lleva `clase` y `mensaje`.

Banderas comunes: `--json`, `--timeout <valor>`, `--offline`, `--dry-run`, `--describe`, `--no-graph`, `--asunto <valor>`, `--verbose`.

<!-- fin de la tabla de comandos -->

## Reglas

1. **Nunca inventar contenido legal ni citar de memoria.** Ni el texto de una norma, ni su identificador, ni ningún
   dato de territorio, ni ninguna fecha: lo que no salga en esta conversación de `kitlegal`, de las referencias o de
   `boe-legislacion`, no se afirma.
2. **Cada afirmación sobre una norma, con su identificador.** El identificador sale de las referencias o de
   `boe-legislacion`; el texto de un artículo, solo de `boe-legislacion`.
3. **Distinguir ley de reglamento.** Di el rango de cada norma que nombres y recuerda que un reglamento nunca puede
   contradecir la ley.
4. **Señalar la variación autonómica.** Di lo que una comunidad puede haber regulado de otro modo, y más aún si su
   régimen es foral; las ordenanzas y demás normas locales no están en las referencias.
5. **No concluir «no existe»** a partir de un resultado sin cobertura completa: ni de un municipio que no está en la
   relación, ni de un boletín `no-configurado`, ni de un DIR3 `no-verificado`, ni de unos festivos que faltan, ni de
   una norma que falta en una referencia.
6. **Ninguna acción con identidad.** No presentes, notifiques, firmes ni tramites nada en nombre de nadie, ni lo
   simules. Si la pregunta lo pide, di que es una acción que hace la persona.
7. **Sin herramienta y sin binario, la respuesta lo dice.** Si no tienes la herramienta y la orden falla porque
   `kitlegal` no está, no lo busques en otras carpetas ni en un plugin: no has consultado nada. No afirmes ningún dato
   de territorio, ninguna fecha ni ningún plazo, ni nada de una norma, ni de memoria ni con salvedades, ni escribas
   ninguna cita. Lleva esta línea, con la causa en el marcador; quien acaba de instalar suele tenerlo y sin arrancar aún:

   ```text
   ⚠ SIN CONSULTA AL BOE: <causa>. Si ya tienes kitlegal, cierra Claude del todo, ábrelo y pregunta en un chat nuevo; si no: https://kitlegal.es/instalar/
   ```

8. **Ninguna fecha que no dé `plazos`.** La fecha de un plazo sale de `data.fecha`; ni la cuentes ni la retoques,
   tampoco para «comprobarla».
9. **Lo que la ley regula aparte no recibe el plazo del común.** Si el asunto está en la cobertura por materia, se
   declara no cubierto, falle o no una lectura; ni se pide un plazo ni se exponen los de la norma que lo rige.
