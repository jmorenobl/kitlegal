# Research: H7.2 · La consulta repetida con evidencia coherente, y respuestas sin la maquinaria interna

Modo desatendido. Cada decisión (D1-D20) lleva su alternativa rechazada y el motivo, con el «Criterio de decisión
autónoma» de la constitución. Toda afirmación sobre una herramienta o una dependencia se comprobó en local y sin red; la
tabla V dice dónde. Lo que no se puede comprobar así son los supuestos S1-S4, que no se afirman como hechos.

## Causa de raíz del ruido (FR-045)

**Medida.** Sobre el informe de H7.1 (`specs/011-h7-1-graph-check-acotado/gates/evals/boe-legislacion.json`, 93
respuestas; V20):

| Medida | Sonnet 5 (evals que activan) | Haiku 4.5 (evals que activan) |
|---|---|---|
| Respuestas | 51 | 30 |
| Última invocación de la sesión: `graph check` | 51 de 51 | 28 de 30 |
| Respuestas con ruido interno | 34 (67 %) | 0 |
| De ellas, con el ruido en el primer párrafo | 34 de 34 | — |
| Primer párrafo que empieza por «Sin hallazgos» o «Código 0» | 34 de 34 | — |
| Primer párrafo que además anuncia la respuesta («Ya tengo todo lo necesario para responder», «Respuesta:») | 22 de 34 | — |
| Primer párrafo que da también el estado de los avisos («ni avisos de vigencia») | 9 de 34 | — |
| Primer párrafo que repite el verbo del protocolo («nada que trasladar», «Sin hallazgos que trasladar») | 5 de 34 | — |
| Con «memoria de consultas» | 24 de 34 | — |
| Sin «memoria de consultas», solo «hallazgos»: cumplen la letra de la regla vigente y narran igual | 10 de 34 | — |
| Con «código 0» | 7 de 34 | — |

En H7 (`d4a96db`, `docs/USO.md` del 2026-09-29, registro no versionado): 43 de 51 con Sonnet 5 (84 %), 41 en el primer
párrafo, 0 de 30 con Haiku 4.5; H7 pedía `graph check` dos veces, antes y después de leer.

**Causa.** El ruido no lo produce la salida de `graph check` ni una ambigüedad de la regla: es un párrafo de transición
que Sonnet 5 escribe tras el resultado de la **última** orden —el estado de lo que acaba de comprobar y el anuncio de que
va a responder— y que `claude -p` entrega como principio de la respuesta final (el `result` del transcript, V13). Haiku
4.5, con la misma última orden en 28 de 30 sesiones, no lo escribe nunca. Desde H7.1 la última orden es siempre
`graph check` (51 de 51), así que el párrafo habla de ella. `SKILL.md` v0.1.1 lo favorece en tres sitios:

1. **Dónde está la regla.** El paso 5, que es lo que el modelo sigue justo antes de redactar, dice «comprueba la memoria
   de consultas … y traslada lo que encuentre»: un paso con resultado que se «traslada», cuyo estado Sonnet 5 informa
   aunque esté vacío (5 de 34 repiten «trasladar»; 9 de 34 informan a la vez de hallazgos y avisos, las dos cosas que el
   paso manda trasladar). La prohibición está en otra sección («Memoria de consultas», última viñeta), lejos del paso.
2. **Qué prohíbe.** Solo «decir algo de la memoria de consultas». 10 de 34 respuestas no la nombran y narran igual
   («Sin hallazgos ni avisos de vigencia»); nada prohíbe los códigos de salida (7 de 34), los hallazgos, las clases ni
   el anuncio de la respuesta.
3. **Por qué.** No dice quién lee la respuesta ni por qué el estado vacío no le sirve: quien pregunta no ve las órdenes;
   «sin hallazgos» no le dice nada del derecho y, además, no distingue un bloque leído antes sin cambios de uno que nunca
   se leyó (FR-041), así que tampoco es cierto que «no haya cambiado».

**Cambios de `SKILL.md` v0.1.2, cada uno trazado a su causa** (contrato en
[contracts/skill-boe-legislacion.md](./contracts/skill-boe-legislacion.md)):

| Cambio | Causa que corrige | Requisito |
|---|---|---|
| C1. En el paso 5, junto a la orden de `graph check`, cómo empieza la respuesta: por lo que se pregunta; sin contar lo que se ha hecho ni lo que ha devuelto ninguna orden, tampoco para decir que no hay nada | 1 (posición) y el párrafo de transición | FR-040 |
| C2. La lista entera de lo que la respuesta no nombra: la memoria de consultas, `graph check` y los verbos de `kitlegal graph`, los códigos de salida, los hallazgos, las clases del binario, el JSON y el sobre; salvo la forma `⚠ REDACCIÓN MODIFICADA:` | 2 (alcance) | FR-040 |
| C3. El porqué: quien pregunta no ve las órdenes ni sus salidas, y la comprobación sin `version-obsoleta` no distingue un bloque leído antes de uno nunca leído | 3 (motivo) | FR-040, FR-041 |
| C4. El paso 5 manda decir el cambio solo si hay `version-obsoleta` («si da `version-obsoleta`…»), en lugar de «traslada lo que encuentre» | 1 (resultado vacío que se informa) | FR-040, FR-042 |
| C5. Nada de lo dicho o respondido en otra conversación: ni afirmarlo, ni confirmarlo, ni desmentirlo, ni decir que no se recuerda | Pregunta que dice «te pregunté» y la sesión que remató «te habría confirmado antes» | FR-041 |
| C6. Las dos fechas de la forma, tal como las da el hallazgo (`AAAAMMDD`) y en la misma línea | Precisión de H7.1 FR 042 | FR-042 |
| C7. Regla 7 reescrita: con otro código, responder con el texto y decir que no se ha podido comprobar si la redacción ha cambiado desde una consulta anterior, sin nombrar la memoria, `graph` ni el código | 2 (la regla vigente pedía decir «la memoria de consultas») | FR-043 |
| C8. Regla 2: lo que no se pudo consultar, por lo que significa (fuente no disponible, límite de consultas, artículo que no está), sin el código | 2 (códigos en la respuesta) | FR-044 |
| C9. La última viñeta de «Memoria de consultas» remite al paso 5 en lugar de repetir una prohibición más estrecha | 1 y 2 | FR-040, FR-047 |

No hay «una prohibición más» sin explicar por qué la actual no basta: la causa es de posición, alcance y motivo, y los
cambios C1-C4 y C9 la atacan ahí. Su efecto no se puede medir sin modelo (S2): lo mide SC-001 en el job de cierre, y la
lista (D1-D8) hace que, si no basta, la sesión no pase en lugar de pasar inadvertido.

## Decisiones

### D1 · Dónde vive la lista: un fichero por skill en su carpeta de evals, con su esquema

**Decisión.** `evals/<skill>/expresiones-prohibidas.yaml`, opcional, validado contra un esquema propio,
`schemas/expresiones-prohibidas.yaml.json` (título `evals/<skill>/expresiones-prohibidas.yaml`). `LeerConjunto` lo
reconoce por su nombre exacto; cualquier otra entrada sigue siendo un fichero de eval. Hoy solo `boe-legislacion` lo
tiene (FR-050).

**Por qué.** Es la convención del repositorio: un esquema por clase de documento YAML (`schemas/eval.yaml.json`,
`normas.yaml.json`, `jerarquia.yaml.json`, `territorio-*.yaml.json`; V3), cada uno compilado con
`skills.CompilarEsquema` desde su ruta, como ya hace `compilarEsquemaDeEval` (V3). La lista es «una por skill»: vive
junto a las evals de la skill que juzga, se lee con ellas y un fallo suyo es un fichero mal formado de esa carpeta
(FR-055) sin mecanismo nuevo.

**Alternativas.** (a) Un `$defs` dentro de `schemas/eval.yaml.json` compilado por fragmento: `jsonschema` v6 lo admite
(`Compiler.Compile` resuelve el fragmento, V2), pero mezcla dos clases de documento en un esquema titulado con la forma
de nombre de una eval y obliga a cambiar `skills.CompilarEsquema`, compartido, para compilar un fragmento. (b) Una clave
en cada eval: la lista es de la skill, no de la eval, y se repetiría en 17 ficheros. (c) `data/`: no es formato de
eval ni la lee `internal/evals` (FR-050).

`schemas/eval.yaml.json` no cambia. El nuevo esquema entra por una tarea `[datos]` (constitución, «Gates»).
`.github/workflows/evals.yml` no cambia: el filtro de apertura ya cubre `evals/*` e `internal/*`, y un cambio solo del
esquema de la lista lo detiene `make ci` antes que el job (la lista mal formada es un fichero mal formado,
`TestEvalsDelRepositorio`); no cambia ningún veredicto que el job pudiera dar distinto.

### D2 · Forma de la lista: dos familias con nombre, expresiones literales

**Decisión.**

```yaml
maquinaria:
  - memoria de consultas
  - …
otra_conversacion:
  - te dije
  - …
```

Dos claves obligatorias, una por familia de FR-050, cada una una lista no vacía de expresiones. Una expresión es una o
más palabras separadas por un espacio, sin blancos al principio ni al final y sin `*` ni `_` (patrón
`^[^\s*_]+( [^\s*_]+)*$`): la comparación tolera los blancos y el énfasis (D3), así que una expresión escrita con ellos
no tendría una forma única. Cualquier otra cosa —clave desconocida o repetida, familia que falta o vacía, expresión
vacía o con esos caracteres— es un fichero mal formado (FR-055), sin casos propios (spec, Edge Cases).

**Por qué.** Las familias son las del spec y las pide FR-083 («una con una expresión de cada familia») y la
comprobación de calibrado, que exige la familia de cada marca (FR-084, US4.2). Literales y no expresiones regulares:
FR-051 compara «por la forma», y una lista de frases la lee y la revisa cualquiera; una expresión regular en un fichero
de datos sería código sin tests en `data`.

**Alternativa.** Una lista plana con la familia en un campo de cada elemento (`{expresion, familia}`): más texto por
elemento para la misma información.

### D3 · Comparación: la tolerancia de H5.1, delimitada como palabra

**Decisión.** Cada expresión se compila una vez en la expresión regular:

```text
(?:^|[^\p{L}\p{N}])  palabra₁  SEP  palabra₂ … SEP  palabraₙ  (?:$|[^\p{L}\p{N}])
```

cada palabra `(?i:` + `regexp.QuoteMeta(palabra)` + `)` y `SEP` el `entrePalabrasDeAviso` de `avisos.go`
(`[\t\p{Zs}*_]*[\t\p{Zs}][\t\p{Zs}*_]*`): sin distinguir mayúsculas, con blancos de más y con el énfasis de Markdown
entre palabras y alrededor, como la forma fija de los avisos (H5.1) y de los hallazgos (H7.1). Una expresión se
encuentra si casa en algún punto de la respuesta. Los extremos piden que lo que rodea a la expresión no sea letra ni
cifra: `hallazgo` no se encuentra dentro de `hallazgos` (la lista lleva las dos) ni `json` dentro de otra palabra; sí
tras `(`, `` ` `` o `*`. No se pliegan tildes ni se toleran saltos de línea entre palabras, como en H5.1.

**Por qué.** FR-051 fija la tolerancia de H5.1; reutilizar su separador es «el mismo código que el job» (FR-061).
Los extremos de palabra evitan que una expresión corta choque con otra palabra (FR-050, «formas que no chocan»).
Verificado sobre las 93 respuestas y los bloques con un prototipo del mismo código fuera del repositorio (V1).

**Alternativas.** Subcadena sin extremos: `te dije` casaría en «te dijeron» y `json` en cualquier palabra que lo
contenga. Plegar tildes: no es tolerancia de H5.1 y el modelo escribe las tildes (V1: todas las marcas del calibrado
las llevan).

### D4 · Las expresiones

**Decisión.** 38 expresiones, calibradas (V1):

- **maquinaria** (22): `memoria de consultas`, `hallazgo`, `hallazgos`, `graph check`, `graph show`, `graph stats`,
  `kitlegal graph`, `version-obsoleta`, `fuente-caducada`, `código de salida`, `códigos de salida`, `código 0`,
  `código 1`, `código 2`, `código 3`, `código 4`, `código 5`, `código 6`, `código 7`, `exit code`, `json`,
  `sobre de salida`.
- **otra_conversacion** (16): `te dije`, `te respondí`, `te confirmé`, `te indiqué`, `te comenté`, `te contesté`,
  `te expliqué`, `te habría dicho`, `te habría respondido`, `te habría confirmado`, `te habría indicado`,
  `te habría comentado`, `te habría contestado`, `te habría explicado`, `conversación anterior`,
  `conversaciones anteriores`.

Cada elemento de FR-040 tiene forma en la lista, salvo el sobre, que solo la tiene como `sobre de salida`: «el sobre» y
«sobre» son castellano corriente (preposición; «en sobre cerrado» en contratación) y chocarían (FR-050); queda como
paráfrasis, visible en el informe (FR-051). Los verbos de `kitlegal graph` son los tres del applet (`show`, `stats`,
`check`; V21) y los códigos, los ocho de ADR 0023 (0-7). En la segunda familia, los siete verbos de decir de la medida de
`docs/USO.md` en pretérito indefinido y en condicional compuesto con «te», y «conversación anterior». **Quedan fuera**,
por chocar con una respuesta legítima o por no atribuir nada dicho en otro momento: `te diría` (condicional simple de
consejo), `te he dicho` (dentro de la misma respuesta), `como te` y `como ya te` (de la medida de USO; «tal como te
muestro»), `te habría` a secas («te habría correspondido la bonificación») y `consulta anterior`, que es justo lo que
FR-043 manda decir cuando `graph check` falla.

**Calibrado (V1).** Sobre las 93 respuestas de H7.1 la lista marca 35: 02 (1), 03 (3), 04 (3), 05 (3), 06 (3), 07 (3),
08 (2), 09 (2), 13 (3), 14 (3), 15 (3), 16 (2), 17 (3) por la maquinaria y 19 (1) por lo dicho en otra conversación
(«te habría confirmado»); 0 de las otras 58, entre ellas las tres de la eval 18 con `⚠ NORMA DEROGADA:` y
`⚠ VIGENCIA AGOTADA:`, las que trasladan `⚠ REDACCIÓN MODIFICADA:` y las dos con «No hay avisos de vigencia». Por
modelo, Sonnet 5 35 de 51 y Haiku 4.5 0 de 30. En el cuerpo de todos los bloques grabados (las dos carpetas de
grabaciones, las dos redacciones del bloque `a1-30` incluidas): 0. En las formas fijas de los avisos y de los hallazgos
y en los ejemplos de `SKILL.md`: 0.

**Alternativa.** Copiar los dos patrones de `docs/USO.md` tal cual: son expresiones regulares (D2) y el de lo dicho en
otra conversación incluye `como (ya )?te`, que choca.

### D5 · Cómo llega la lista a `Juzgar`: el lector la pone en cada eval

**Decisión.** `Eval` gana `Prohibidas ExpresionesProhibidas` (`yaml:"-"`, como `Fichero`): no es clave del YAML, la
pone `LeerConjunto` en cada eval del directorio tras leer la lista. `Conjunto` gana también `Prohibidas`, la lista leída
(vacía si la carpeta no la tiene o está mal formada). `Juzgar(eval, sesion, skill)` no cambia de firma: si `eval.Activa`,
busca en la respuesta las expresiones de `eval.Prohibidas`.

**Por qué.** Una eval de una skill sin lista lleva el valor cero y se juzga como antes, sin tocar ninguna llamada ni
test que ya exista (FR-050, FR-052). `Conjunto.Prohibidas` lo usan el informe (¿tiene la skill lista?, FR-053) y la
comprobación del quickstart (FR-061).

**Alternativa.** Un cuarto parámetro de `Juzgar`: cambia cada una de sus llamadas en `juzgar_test.go` (1 791 líneas) e
`informe.go` para decir lo mismo que el valor cero.

### D6 · El juicio: un motivo por expresión, y la respuesta no pasa

**Decisión.** `ResultadoDeEval` gana `ExpresionesProhibidas []string` (`json:"expresiones_prohibidas"`): las
encontradas, en el orden de la lista (maquinaria y después otra conversación), cada una una vez; vacía, `[]`, si ninguna
o si no se aplica. Cada una añade el motivo `expresión prohibida: <expresión>` detrás de los del territorio y antes del
del modelo, y `Pasa` exige que no haya ninguna. No se aplica a una eval de no activación (FR-052), ni cambia nada en
ella.

**Por qué.** Es «como una con una cita ausente» (FR-052): un motivo y no pasa. El umbral del ADR 0016 decide en las
series que deciden y la tasa se publica en las informativas, sin más cambios (FR-054).

### D7 · El informe: por sesión y por modelo

**Decisión.** Por sesión, el campo de D6 en `informe.json` y una columna «Expresiones prohibidas» en la tabla de
sesiones de `informe.md`, entre «Territorio ausente» y «Resultado» (`ninguna` si vacía). En la raíz,
`expresiones_prohibidas_por_modelo`, detrás de `tasas`: un elemento por modelo del job —el que decide y después los
informativos, en su orden— con `modelo`, `con_alguna` y `respuestas`. `respuestas` cuenta las sesiones juzgadas (no
ilegibles) de las series **planificadas** cuya eval activa la skill; `con_alguna`, las de esas con alguna expresión. La
sesión de la prueba de red forma una serie no planificada (V14): se juzga con la lista y publica sus expresiones, pero no
entra en el recuento (FR-053). En `informe.md`, una sección «Expresiones prohibidas por modelo», detrás de «Tasas por
eval», con una tabla `Modelo | Respuestas con alguna expresión | Respuestas en evals que activan la skill`. Una skill sin
lista da `[]` y el párrafo `la skill no tiene lista de expresiones prohibidas`: un recuento de 0 diría que se buscó.

**Por qué.** Es lo que pide FR-053 y lo que la persona compara con el umbral de SC-001 (≤ 2 de 51, ≤ 1 de 30). Contar
las sesiones juzgadas: una ilegible no tiene respuesta juzgada, y ya da `fallo`.

**Alternativa.** El umbral del 5 % en el veredicto: fuera de alcance (spec, «Fuera de alcance»).

### D8 · Lista mal formada

**Decisión.** Un `expresiones-prohibidas.yaml` que no es un fichero regular, que no se puede leer o que no valida es un
`FicheroMalFormado` con su error, como un fichero de eval (FR-055): `make ci` falla nombrándolo
(`TestEvalsDelRepositorio`, subtest `formato`), `PrepararSesion` y `TestPlanDeSesiones` no preparan nada y el informe lo
lista y da `fallo`, lo que ya hacen con cualquier mal formado (V14). Las evals de esa carpeta se leen sin lista.

### D9 · Calibrado contra el informe versionado de H7.1

**Decisión.** Subtest `expresiones-calibradas` de `TestEvalsDelRepositorio`: lee
`../../specs/011-h7-1-graph-check-acotado/gates/evals/boe-legislacion.json` en un tipo local (`evals[].eval`,
`evals[].respuesta`), aplica la lista del repositorio a cada respuesta con la función del juicio y exige el reparto de
FR-084 por las dos cifras de la eval y por familia, y 0 en las demás.

**Por qué.** FR-084 nombra ese fichero, versionado y que no se edita (FR-070): leerlo es la verdad terreno, sin copia.
Por las dos cifras y no por el nombre del fichero: el de la eval 19 retirada no puede aparecer en un test (FR-020).

**Alternativa.** Copiar las 93 respuestas a `testdata/`: una segunda copia del mismo dato, con una tarea `[datos]` más.

### D10 · Ninguna expresión en el texto de los bloques grabados

**Decisión.** Subtest `expresiones-en-los-bloques` de `TestEvalsDelRepositorio`: lee con `boe articulo`, en proceso y
sobre `UnionDeGrabaciones()`, cada bloque de las consultas necesarias de las evals de `boe-legislacion`
(`ConsultasNecesarias`) y, sobre las grabaciones de cada grafo previo, cada bloque de sus comandos; ninguna expresión de
la lista casa con el `texto` de ninguno (FR-085). Con eso entran las dos redacciones de `a1-30`: la vigente por la
grabada y la original por la derivada.

**Por qué.** Lo que una respuesta puede transcribir es el `texto` que da `boe`, no el XML; buscar en el `texto` es la
comparación de FR-051 sobre lo que se transcribe. Si una eval futura cita una norma cuyo texto lleva «hallazgo» (p. ej.
hallazgos arqueológicos), el subtest falla y obliga a revisar la lista antes de medir nada.

**Alternativa.** Buscar en el cuerpo XML de cada fichero `*_texto_bloque_*`: incluye bloques que ninguna eval consulta y
redacciones que nadie transcribe, y el marcado puede partir una expresión.

### D11 · Lo que `SKILL.md` pide decir no lleva ninguna expresión

**Decisión.** Subtest `expresiones-de-la-skill` de `TestEvalsDelRepositorio`: ninguna expresión de la lista casa con
(a) la forma fija escrita de cada aviso y de cada clase de hallazgo etiquetada (`formaEscrita` de
`boe.EtiquetasDeAviso` y `grafo.EtiquetasDeHallazgo`), ni (b) con el contenido de ningún bloque ```` ```text ```` de
`skills/boe-legislacion/SKILL.md`, que son los textos que la skill enseña a escribir en la respuesta (la cita, el aviso,
la línea `⚠ REDACCIÓN MODIFICADA:` y, desde v0.1.2, la frase de la regla 7).

**Por qué.** FR-051 (las formas no son expresiones prohibidas) y FR-043 («ni la regla 7 ni el ejemplo … MUST pedir a la
respuesta ninguna expresión de la lista») se comprueban por la forma, sin modelo (constitución, «Gates», capa 1). Por
eso la frase de la regla 7 va en v0.1.2 en un bloque `text`, como los demás ejemplos.

### D12 · La eval nueva

**Decisión.** `evals/boe-legislacion/19-lcsp-contrato-menor-redaccion-cambiada.yaml`, con la pregunta literal de
FR-001, `activa: true`, `informativa: true`, `grafo_previo` `lcsp-a1-30-redaccion-original` con el comando de bloque
`boe BOE-A-2017-12902 a1-30`, los comandos del bloque y de la comprobación con la norma, `graph show` prohibido, la cita
del bloque y `hallazgos: [version-obsoleta]` (contracts/eval-y-derivada.md §1). El nombre sigue la forma de la eval 02,
que pregunta por el mismo artículo. Cumple las reglas del conjunto sin cambiarlas (FR-030): 19 ficheros, 10 positivas
que deciden, informativa y positiva, norma en `data/normas.yaml`.

**Por qué.** Lo exigen FR-001 a FR-004. Las consultas que necesita (`boe articulo`, el índice y los metadatos de
`BOE-A-2017-12902`) ya están en las grabaciones de H4 (V16): no hay grabación nueva (FR-014).

### D13 · La derivada del grafo previo: la produce código, la comprueba `boe`

**Decisión.** `testdata/evals/grafo-previo/lcsp-a1-30-redaccion-original/` con un fichero, del nombre de la grabación de
H4 del bloque `a1-30` de `BOE-A-2017-12902`. En `internal/app/grafo_test.go`:

- **La derivación** es una función que toma la grabación y una fecha de vigencia y quita del `cuerpo` los elementos
  `<version>` del `<bloque>` posteriores al de esa fecha —desde el cierre de esa versión hasta el cierre de la última,
  blancos incluidos— sin cambiar nada más: ni la petición, ni las cabeceras, ni el estado, ni el resto del cuerpo. La
  grabación se lee y se escribe con un tipo local con los campos del formato de grabación de `httpx` en su orden y el
  mismo codificador (sangrado de dos espacios, sin escapar HTML, salto final); su premisa, que derivar sin quitar nada
  devuelve la grabación byte a byte.
- **La escribe** `go test -count=1 -run '^TestGrabacionesDerivadas$' ./internal/app/ -args -actualizar-derivadas`, que
  antes de comprobar escribe cada derivada del grafo previo desde su grabación; es el patrón de las banderas
  `-actualizar-esquemas`, `-regenerar-skills` y `-actualizar-golden` del repositorio y de `-update` en la biblioteca
  estándar (V18). La tarea `[datos]` la ejecuta; nada se escribe a mano (FR-010).
- **La comprueba** `TestGrabacionesDerivadas`, y la comprobación la decide la carpeta: las derivadas de
  `grafosPreviosDeLasEvals` se declaran en una lista propia, de un tipo nuevo —la derivada del grafo previo, que declara
  su subcarpeta, la grabación de la que sale, los argumentos de `boe` y la fecha de vigencia de la redacción que da, sin
  comprobación propia—, separada de `grabacionesDerivadas()`, que queda con las del e2e; y el test compara cada carpeta
  con su lista, en los dos sentidos. Una entrada de otra clase con carpeta de grafo previo (`versionDelArticulo21`, que
  se queda, admite cualquier carpeta) falla nombrada: es una entrada del e2e sin fichero en su carpeta. Toda derivada
  del grafo previo pasa, así, por 1 y 2; 3 es el test de que 2 rechaza lo inventado:
  1. la derivada es, byte a byte, la que da la derivación sobre la grabación (FR-010: reproducible y sin otro cambio);
  2. servida en lugar de la grabación, `boe articulo` da exactamente una de las redacciones que trae la grabada —la que
     da `boe` sobre la grabación reducida a cada una de sus versiones—, entera (fecha de versión y de vigencia, norma
     modificadora, texto, huella y todo lo demás), y es la de la fecha declarada (FR-011; SC-005: 20180309);
  3. un test aparte arma en `t.TempDir()`, desde la grabación, tres derivadas que no pasan —con la fecha 20151002, con un
     párrafo de más y con las dos cosas, como la retirada— y exige que la comprobación 2 las rechace nombrándolas
     (FR-012, SC-005). La huella sale del texto: no hay derivada con una huella inventada y el texto de la grabada.
- Las derivadas del e2e (`version-posterior`, `version-ulterior`, `sin-eli`, `eli-sin-segmento`) siguen con su
  comprobación, y todo fichero de las dos carpetas sigue con su comprobación, y toda comprobación con su fichero
  (FR-013), ahora carpeta a carpeta.
- La comparación carpeta a carpeta entra con la retirada de `lpac-a21-version-anterior` (D15), que es una entrada del
  e2e con carpeta de grafo previo y la haría fallar; hasta entonces sigue la de hoy, con las dos carpetas juntas.

**Por qué.** FR-010 pide que la produzca código del repositorio de forma reproducible; FR-011 y FR-012, una
comprobación semántica que no depende de cómo se hizo y que alcance a toda derivada del grafo previo. La comparación
byte a byte es la que dice «sin ningún otro cambio»; la de `boe`, la que dice «una redacción que la grabada trae»; la
carpeta, y no la clase que elige quien declara, la que dice que ninguna derivada del grafo previo se la salta.

**Alternativas.** (a) Un guion o un programa fuera del repositorio, como las derivadas de H7: no es código del
repositorio (FR-010). (b) Derivar al preparar la sesión, sin fichero: cambia `PrepararSesion` y el formato de eval, y
FR-010 dice que la derivada entra por una tarea `[datos]`. (c) Solo la comprobación de `boe`: no dice que la derivada
salga de la grabación sin nada más. (d) Una clase más en `grabacionesDerivadas()`, con la comparación de las dos
carpetas juntas: una entrada del e2e con carpeta de grafo previo pasaría con una fecha y un párrafo inventados, como la
retirada (FR-012).

### D14 · FR-002 en la preparación de la eval

**Decisión.** La subprueba `grafo-previo` de `TestEvalsDelRepositorio` (`compruebaElGrafoPrevio`) gana, para cada eval
con `grafo_previo`, lo que hace la sesión: tras preparar el grafo previo, lee en proceso cada bloque de sus comandos con
`boe articulo` sobre `UnionDeGrabaciones()` entregando al grafo de la sesión, y ejecuta `graph check` con cada norma y
sus bloques; exige código 0, que las clases de los hallazgos sean las de `hallazgos` de la eval y que cada
`version-obsoleta` lleve en `fecha_vigencia` la de la redacción que dejó el grafo previo y en `fecha_vigencia_reciente`
la que acaba de leer.

**Por qué.** Es la «subprueba `grafo-previo`» que el spec nombra como prueba independiente de US2 y lo que hace
comprobable FR-002 en `make ci`, sin modelo. Es genérica: sirve a toda eval con grafo previo, sin nombrar la 19.

### D15 · Lo que se retira y lo que se adapta (FR-020, FR-021)

**Decisión.** Salen `evals/boe-legislacion/19-lpac-articulo-21-redaccion-cambiada.yaml`,
`testdata/evals/grafo-previo/lpac-a21-version-anterior/`, su entrada en `grabacionesDerivadas()` y
`parrafoDeLaVersionAnterior`, y con ellos `TestGrabacionesDerivadas` pasa a comparar carpeta a carpeta (D13).
`versionDelArticulo21` se queda: la usan `version-posterior` y `version-ulterior`. Se
adaptan, sin cambiar lo que comprueban, los tests que los nombran (V21): `internal/evals/formato_test.go` (documentos
sintéticos con `grabaciones: lpac-a21-version-anterior`, que pasan a `lcsp-a1-30-redaccion-original`),
`internal/evals/consultas_test.go` (el grafo previo sintético) e `internal/evals/juzgar_test.go` (la eval sintética de la
consulta repetida, que pasa a la forma de la nueva: `BOE-A-2017-12902`, `a1-30`, la comprobación con la norma). El
nombre sintético `06-lpac-articulo-21-redaccion-cambiada.yaml` de `consultas_test.go` no es el fichero retirado y se
queda. `CHANGELOG.md` y `docs/` los nombran como historia y no son código ni test.

### D16 · `SKILL.md` v0.1.2

**Decisión.** Los cambios C1-C9 de la causa de raíz, y ninguno más (FR-047): se quedan H7 FR 081, la parte de H7 FR 082
del código 0, H7.1 FR 040, 041, 043 y 045, la forma de la cita y la de los avisos (FR-046). La tabla de comandos no cambia
(el binario no cambia), `make skills-check` sin drift, < 300 líneas (hoy 251; v0.1.2, ≈ 265), sin nombrar evals, el
job ni modelos. Contrato: [contracts/skill-boe-legislacion.md](./contracts/skill-boe-legislacion.md).

### D17 · El escenario del quickstart (FR-060, FR-061)

**Decisión.** Todo en `d=$(cd "$(mktemp -d)" && pwd -P)` (la ruta física, como `scripts/evals.sh`):

1. **Binario** del árbol de la rama: `CGO_ENABLED=0 go build -trimpath -o "$d/bin/kitlegal" ./cmd/kitlegal`, primero en
   el `PATH` de cada orden del escenario (`PATH="$d/bin:$PATH"`).
2. **Preparación**, la del job: `go test -tags evals -count=1 -run '^TestPrepararSesion$' ./internal/evals/ -args
   -skill boe-legislacion -eval 19-lcsp-contrato-menor-redaccion-cambiada.yaml -modelo claude-sonnet-5
   -sesion "$d/sesion"`, con `"$d/sesion/cache"` creado vacío antes (V14). Deja el grafo previo y la caché en
   `"$d/sesion/cache"` y la pregunta en `"$d/sesion/pregunta.txt"`.
3. **Skill**, con ese binario y en ámbito local, en el directorio de trabajo de las conversaciones:
   `kitlegal skills install boe-legislacion --host claude` desde `"$d/trabajo"` (`.agents/skills/` y el enlace relativo
   en `.claude/skills/`; V12).
4. **Cada conversación**, desde `"$d/trabajo"`, con `KITLEGAL_CACHE_DIR="$d/sesion/cache"`:

   ```text
   claude -p "$(cat "$d/sesion/pregunta.txt")" --model claude-sonnet-5 --output-format stream-json --verbose
     --no-session-persistence --setting-sources project --permission-mode dontAsk
     --allowedTools 'Bash(kitlegal *)' Read Skill
   ```

   con la salida en `<dir>/sesion.jsonl`, la de error en `<dir>/sesion.err` y el código en `<dir>/codigo-de-la-sesion`,
   los tres ficheros que lee `LeerSesion` (V13), con `<dir>` `"$d/primera"` y después `"$d/segunda"`. Sin `--continue`
   ni `--resume`: la segunda es una conversación nueva.
5. **Comprobación**: `go test -tags evals -count=1 -v -run '^TestComprobarConsultaRepetida$' ./internal/evals/ -args
   -skill boe-legislacion -primera "$d/primera" -segunda "$d/segunda" -fecha-superada 20180309 -fecha-leida 20200206`
   (contracts/comprobacion-del-quickstart.md).
6. `rm -rf "$d"`.

**Banderas, una por una.** `--model claude-sonnet-5`, el modelo que decide en el job y el del ruido medido (V22).
`--output-format stream-json --verbose`: el formato del job, que exige `--verbose` (V7); la respuesta final es el campo
`result` del último mensaje `result`, la misma que juzga el job (V13). `--no-session-persistence`: la conversación no se
guarda en disco (V6), así que no queda ninguna fuera de `"$d"`. `--setting-sources project`: **es lo que garantiza que
la skill es la instalada en `"$d/trabajo"`**: el cargador de skills de Claude Code solo lee `~/.claude/skills` si la
fuente `userSettings` está activa y lee `.claude/skills` del directorio de trabajo si lo está `projectSettings` (V8);
con solo `project`, la `boe-legislacion` de la cuenta de la persona, de otra versión, no se carga, y tampoco sus
ajustes, ganchos ni `CLAUDE.md` de usuario. `--permission-mode dontAsk` con `--allowedTools 'Bash(kitlegal *)' Read
Skill`: los permisos justos —activar la skill, leer sus `references/` y ejecutar `kitlegal`—; todo lo demás que pediría
permiso se deniega sin preguntar (V9), también `WebFetch` y `WebSearch`. `--permission-mode bypassPermissions`, el del
job, no es «justo» en la máquina de una persona.

**Caché.** `boe articulo` e `indice` se guardan siete días y un artículo en su entrada vigente se sirve sin pedir nada
(V10): las dos conversaciones leen el bloque de la caché preparada con la redacción de 20200206. `buscar` y `metadatos`
caducan a los cinco minutos; si una conversación los pide pasados, `kitlegal` los pide al BOE como en cualquier uso
—por `internal/httpx`, fuente con fila revisada—, y nada de lo que comprueba el escenario depende de ellos (ninguno es
una lectura de bloque).

**Alternativas.** `CLAUDE_CONFIG_DIR` vacío o un `HOME` temporal para aislar la cuenta: en macOS la credencial va en el
llavero, ligada a la configuración, y la conversación podría quedarse sin credencial (S3). `--bare`: solo admite
`ANTHROPIC_API_KEY`, y el proyecto usa la suscripción (V6). Instalar en global o con `make install`: fuera de alcance
(spec). Copiar las respuestas a mano o comprobarlas con órdenes de shell: fuera de alcance (spec).

### D18 · La comprobación del quickstart en Go

**Decisión.** `TestComprobarConsultaRepetida` en `internal/evals/job_test.go` (etiqueta `evals`, fuera de `make ci`
como los otros tres puntos de entrada del job) con cuatro banderas nuevas: `-primera`, `-segunda`, `-fecha-superada`,
`-fecha-leida`, y la `-skill` que ya hay. Lee la lista con `LeerConjunto` de `evals/<skill>/` y cada conversación con
`LeerSesion`, y llama a `comprobarConsultaRepetida`, una función sin exportar de `internal/evals/consulta_repetida.go`,
probada en `make ci` (`consulta_repetida_test.go`), que devuelve una línea por condición que falla:

1. la primera respuesta lleva la forma `⚠ REDACCIÓN MODIFICADA:` —la expresión de `formasDeHallazgo`, la del juicio—
   en una línea que tiene también las dos fechas como palabras (`contieneComoPalabra`);
2. la segunda no la lleva (`ExtraerHallazgos`);
3. ninguna de las dos lleva una expresión de la lista (`ExtraerExpresionesProhibidas`, la del juicio).

Antes, cada conversación tiene que haberse leído y haber terminado (`Terminada`); si no, la línea lo dice con el motivo
de `LeerSesion` o `MotivoSinTerminar`, porque una conversación sin respuesta no puede pasar las condiciones 2 y 3 en
vacío. El test falla con todas las líneas; si no hay ninguna, registra que se cumplen las tres.

**Por qué.** FR-061 pide un punto de entrada con la etiqueta `evals` y el mismo código que el job. La función, fuera del
fichero con etiqueta, se prueba en `make ci`: una comprobación que dijera «se cumple» sin cumplirse engañaría a la
persona que acepta el hito. Las fechas llegan por bandera: son las de FR-061 y no se escriben en código.

**Alternativa.** Todo dentro del test con etiqueta: sin tests en `make ci`.

### D19 · Aceptación e2e: no aplica

**Decisión.** El hito no cambia el binario (spec, «Fuera de alcance»; `cmd/kitlegal` no depende de `internal/evals`,
V17): no hay comportamiento del binario que un guion `testscript` pueda describir. Su aceptación es el job de evals de
cierre (SC-001) y el escenario del quickstart (SC-002); en `make ci` la fijan `TestGrabacionesDerivadas`, las subpruebas
de `TestEvalsDelRepositorio` y los tests de formato, de `Juzgar`, del informe y de la comprobación del quickstart.

### D20 · Documentación que el hito deja falsa

**Decisión.** `CHANGELOG.md` (*Unreleased*): la entrada de `boe-legislacion` pasa a v0.1.2 y dice que sustituye a la
v0.1.1, que no llegó a publicarse; la de la eval de la consulta repetida pasa a la eval nueva y dice qué retira; y una
entrada nueva, la lista de expresiones prohibidas y su decisión (FR-048, FR-086). `CONTRIBUTING.md`, «Formato común de
eval»: el fichero de la lista y la condición nueva de «Una sesión de una eval pasa si…»; «Job de evals»: lo que el
informe publica de la lista (FR-050, FR-053). `README.md` no dice nada de esto (V21) y no cambia.

## Datos externos

Ninguno. No hay fuente nueva ni grabación nueva (FR-014): la eval nueva usa la grabación de H4 del bloque `a1-30`, el
índice y los metadatos de `BOE-A-2017-12902` (`internal/source/boe/testdata/boe.legislacion-consolidada/`; fuente
`boe.legislacion-consolidada`, fila revisada en `docs/SOURCES.md`), y la derivada la produce código del repositorio a
partir de ella (D13). Ningún manifiesto `grabaciones.json` ni test `TestGrabar*` cambia; el paso `grabar_datos` no tiene
nada que grabar.

## Verificaciones (V)

| # | Afirmación | Dónde se comprobó |
|---|---|---|
| V1 | La lista de D4, con la comparación de D3 escrita en Go (`regexp`, `\p{Zs}`, `(?i:…)`, extremos `[^\p{L}\p{N}]`), marca 35 de las 93 respuestas con el reparto de FR-084 (34 maquinaria, 1 otra conversación, en la 19), Sonnet 5 35 de 51 y Haiku 4.5 0 de 30; 0 en el cuerpo de los 16 ficheros `*_texto_bloque_*` de las dos carpetas de grabaciones; 0 en las formas fijas y en los ejemplos de `SKILL.md` y en «No se ha podido comprobar si la redacción ha cambiado desde una consulta anterior.»; casa con «SIN HALLAZGOS», «Sin  hallazgos en la **memoria de consultas**», «la *memoria* de _consultas_», «te confirmé.» (con blanco sin separación U+00A0 también), «`graph check`», «(código 0)»; no casa con «hallazgoss» | Prototipo `go run /tmp/h72/main.go` desde la raíz, fuera del repositorio y sin red, contra `specs/011-…/gates/evals/boe-legislacion.json`, `internal/source/boe/testdata/boe.legislacion-consolidada/` y `testdata/evals/boe.legislacion-consolidada/` |
| V2 | `Compiler.Compile(loc)` resuelve un fragmento JSON pointer en `loc` | `github.com/santhosh-tekuri/jsonschema/v6@v6.0.3/compiler.go:178-188` (`absolute`, `resolveFragment`) |
| V3 | Un esquema por documento YAML en `schemas/*.yaml.json`, compilado desde su ruta; `compilarEsquemaDeEval(ruta)` llama a `skills.CompilarEsquema(contenido)`, que registra y compila la raíz | `ls schemas/`; `internal/evals/formato.go:224-239`; `internal/skills/esquemas.go:30-49`; `internal/skills/normas.go:19`, `jerarquia.go:13`, `territorio.go:19` |
| V4 | `encoding/json/v2` codifica un slice nil como `[]` salvo con `FormatNilSliceAsNull` | `go doc encoding/json/v2.FormatNilSliceAsNull` (go1.27.1) |
| V5 | `regexp/syntax`: `(?i)` sin distinguir mayúsculas; `\p{…}` clases Unicode de `unicode.Categories` | `go doc regexp/syntax` (líneas 23-26, 63, 97, 137) |
| V6 | Banderas de `claude` 2.1.284: `-p`, `--model`, `--output-format text\|json\|stream-json`, `--verbose`, `--no-session-persistence` («sessions will not be saved to disk … only works with --print»), `--setting-sources <user, project, local>`, `--permission-mode` con `dontAsk`, `--allowedTools <tools...>` con la forma `"Bash(git *) Edit"`, `-c/--continue`, `-r/--resume`, `--bare` («Anthropic auth is strictly ANTHROPIC_API_KEY or apiKeyHelper») | `claude --help` (2.1.284, `~/.local/share/claude/versions/2.1.284`) |
| V7 | Con `--print`, `stream-json` exige `--verbose` | binario 2.1.284: «Error: When using --print, --output-format=stream-json requires --verbose.» |
| V8 | El cargador de skills registra «Loading skills from: managed=…, user=…, project=[…]» y carga las de usuario con `lr("userSettings")&&!b?YM(r,"userSettings","skills",n):…` y las de proyecto solo si `lr("projectSettings")`, siendo `lr` `isSettingSourceEnabled` | binario 2.1.284: función `F1o` (búsqueda de `Loading skills from: managed=` y `lr("projectSettings")`) y la exportación `lr as isSettingSourceEnabled` |
| V9 | `dontAsk`: «don't ask (auto-deny anything that would prompt)» | binario 2.1.284 (tabla de descripciones de modos) |
| V10 | `buscar` y `metadatos` se guardan 300 s; `indice`, `articulo`, `articulos` y `analisis`, 604 800 s; un artículo en su entrada vigente se sirve sin pedir nada ni mirar los metadatos | `internal/source/boe/fuente.go:22-31` y `:215-231`; `internal/source/boe/articulo.go:209-238` |
| V11 | Cada bloque que devuelven `boe articulo`/`articulos` y llega al grafo es una lectura, «la sirva la fuente o la caché» | `CHANGELOG.md`, *Unreleased*, «Lo que se observa hoy» (H7.1) |
| V12 | `kitlegal skills install [<skill>...] [-g] [--host <host>]... [--dir <ruta>]`: sin `-g`, ámbito local | `internal/app/instalacion.go:164-167` |
| V13 | `LeerSesion(dir)` lee `sesion.jsonl`, `codigo-de-la-sesion` y `sesion.err`; `Respuesta` es el `result` del último mensaje `result` con `success` y sin `is_error`; `Terminada` y `MotivoSinTerminar` | `internal/evals/sesion.go:14-19`, `:55-127` |
| V14 | `TestPrepararSesion` recibe `-skill -eval -sesion -modelo`, y `PrepararSesion` escribe solo en el directorio de la sesión (con su `cache/` ya creado): grafo previo, caché, `eval.txt`, `modelo.txt`, `pregunta.txt`; rechaza un directorio con ficheros mal formados; el informe pone la sesión de la prueba de red en una serie no planificada | `internal/evals/job_test.go:88-115`; `internal/evals/preparar.go:198-298`; `internal/evals/informe.go:362-372` y `:554-597` |
| V15 | La grabación H4 del bloque `a1-30` trae dos `<version>`: `BOE-A-2017-12902`, publicación 20171109, vigencia 20180309; `BOE-A-2020-1651`, publicación 20200205, vigencia 20200206; sin cabecera `Content-Length` | `internal/source/boe/testdata/boe.legislacion-consolidada/GET_…_BOE-A-2017-12902_texto_bloque_a1-30.json` |
| V16 | Las grabaciones de H4 tienen el bloque, el índice y los metadatos de `BOE-A-2017-12902` | `ls internal/source/boe/testdata/boe.legislacion-consolidada/ \| grep 12902` |
| V17 | El binario no depende de `internal/evals` | `go list -deps ./cmd/kitlegal \| grep -c internal/evals` → 0 |
| V18 | Banderas de test que escriben ficheros versionados: `-actualizar-esquemas`, `-regenerar-skills`, `-actualizar-golden`; en la biblioteca estándar, `-update` | `internal/app/esquemas_test.go:28`, `internal/app/skills_test.go:33`, `internal/source/boe/fuente_test.go:35`; `$GOROOT/src/go/printer/printer_test.go:28` (go1.27.1) |
| V19 | `golangci-lint` carga los ficheros con la etiqueta `evals` | `.golangci.yml:39-43` |
| V20 | Las medidas de la causa de raíz (tabla de arriba) | `jq` sobre `specs/011-…/gates/evals/boe-legislacion.json`: última `invocaciones[].orden` por modelo; primer párrafo (`split("\n\n")[0]`) con `test(…; "i")` |
| V21 | Nombran lo retirado: `internal/evals/formato_test.go` (66, 412, 491, 497, 502, 507, 549), `consultas_test.go:120`, `juzgar_test.go:63, 1441`, `internal/app/grafo_test.go:2549-2601`; los verbos de `graph` son `show`, `stats` y `check`; `README.md` y `docs/` no describen la eval 19 ni la memoria de consultas | `git grep` de los nombres; `internal/app/grafo.go:77-100` |
| V22 | El job fija `claude-sonnet-5` como modelo que decide, `claude-haiku-4-5-20251001` como informativo, 3 repeticiones y umbral 2 | `.github/workflows/evals.yml:87-92` |
| V23 | Los dos patrones de `docs/USO.md` marcan 35 respuestas de H7.1 con el reparto de FR-084 | `jq … test("…"; "i")` sobre el informe de H7.1 |

## Supuestos no verificados

- **S1.** La versión de Claude Code de la persona (y la del runner del job) mantiene las banderas y el cargador de skills
  comprobados en 2.1.284 (V6-V9). No se puede comprobar otra versión sin instalarla.
- **S2.** Que v0.1.2 baje el ruido de Sonnet 5 por debajo del umbral: el cambio ataca la causa medida, pero solo una
  ejecución con modelo lo dice; lo mide SC-001 en el job de cierre, y el hito no lo afirma antes.
- **S3.** La credencial de la persona es la de su cuenta (llavero u OAuth), no un `apiKeyHelper` de sus ajustes de
  usuario, que con `--setting-sources project` no se leería.
- **S4.** Claude Code carga una skill cuyo directorio en `.claude/skills/` es un enlace relativo a `.agents/skills/`
  dentro del directorio de trabajo: lo comprobó H5 en 2.1.270 (el cargador admite directorios enlazados) y es el diseño
  de la instalación local con `--host claude` (ADR 0019, ADR 0025); en 2.1.284 no se ha vuelto a comprobar sin abrir
  una conversación.
