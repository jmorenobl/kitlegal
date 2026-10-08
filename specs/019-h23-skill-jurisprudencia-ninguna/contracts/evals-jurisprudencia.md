# Contrato: las evals de `jurisprudencia`, su juicio y sus umbrales

Lo que el formato común gana (FR-052), cómo se juzga sin modelo (FR-051), el hecho de la sesión que decide (FR-060,
FR-061), los seis ficheros (FR-050) y el trabajo del job (FR-063). Los seis ficheros entran con la tarea que
extiende el formato, y no con la primera: `make ci` lee las evals de toda carpeta de `evals/` y no admitiría sus
claves antes (research V34, D23).

## 1. Lo que gana `schemas/eval.yaml.json`

Nada de hoy cambia de significado: una eval sin las claves nuevas se lee y se juzga igual.

**Dos formas de comando**, en `comandos`, junto a las cinco que hay:

```json
"comando-preparar": {
  "type": "object", "additionalProperties": false, "required": ["applet", "verbo"],
  "properties": {
    "applet": { "$ref": "#/$defs/applet" },
    "verbo": { "const": "preparar" },
    "roj": { "$ref": "#/$defs/roj" },
    "con_texto": { "const": true }
  }
},
"comando-cotejar": {
  "type": "object", "additionalProperties": false, "required": ["applet", "verbo"],
  "properties": {
    "applet": { "$ref": "#/$defs/applet" },
    "verbo": { "const": "cotejar" },
    "roj": { "$ref": "#/$defs/roj" }
  }
}
```

**Una clave**, `sentencias`, en la raíz:

```json
"sentencias": {
  "type": "object", "additionalProperties": false, "minProperties": 1,
  "not": { "required": ["citas", "ninguna_cita"] },
  "properties": {
    "citas": { "type": "array", "minItems": 1, "items": {
      "type": "object", "additionalProperties": false, "required": ["ecli", "roj"],
      "properties": { "ecli": { "$ref": "#/$defs/ecli" }, "roj": { "$ref": "#/$defs/roj" } } } },
    "ninguna_cita": { "const": true },
    "sin_cita_del_roj": { "type": "array", "minItems": 1, "items": { "$ref": "#/$defs/roj" } },
    "no_comprobada": { "const": true },
    "direcciones": { "type": "array", "minItems": 1, "items": { "type": "string", "format": "uri" } },
    "casillas": { "type": "array", "minItems": 1, "items": {
      "type": "object", "additionalProperties": false, "required": ["nombre", "valor"],
      "properties": { "nombre": { "type": "string", "minLength": 1 }, "valor": { "type": "string", "minLength": 1 } } } },
    "direccion_de_busqueda": { "const": true }
  }
}
```

con `"ecli": {"type": "string", "pattern": "^ECLI:ES:[A-Z][A-Z0-9]{0,6}:[0-9]{4}:[A-Z0-9.]{1,25}$"}` y
`"roj": {"type": "string", "pattern": "^[A-Z]+( [A-Z]+)* [0-9]+/[0-9]{4}$"}` en `$defs`, las formas de FR-005.

**La regla de la raíz**: `sentencias` solo la admite una eval que activa la skill y que no es sin binario ni servidor
—entra en las dos listas `not` de hoy—; y una eval con `sentencias` puede no llevar `comandos`, `citas` ni
`territorio`. El `else` de hoy, `required: [comandos]` con `citas` o `territorio`, pasa a ser una de dos: eso mismo, o
`required: [sentencias]`.

## 2. El juicio de una sesión

Cada cosa que falta da su motivo y la sesión no pasa. Nada de esto usa un modelo ni cambia de un modo a otro.

| En la eval | La sesión lo cumple si… | Motivo si no |
|---|---|---|
| comando `cita preparar` o `cita cotejar` | alguna invocación del applet `cita` con ese verbo consultó y terminó con 0, pedida como orden o como herramienta | el de todo comando ausente, con `cita preparar` o `cita cotejar` |
| …con `roj` | además, su `--roj` vale eso, se escriba `--roj <v>` o `--roj=<v>` | …con `--roj <v>` detrás |
| …con `con_texto` | además, lleva `--texto` con un valor no vacío | …con `--texto` detrás |
| `citas` | cada pareja está entre las citas de la respuesta (§3), con el ECLI y el ROJ iguales carácter a carácter | `falta la cita de sentencia [<ECLI>, ROJ: <ROJ>]` |
| `ninguna_cita` | la respuesta no tiene ninguna cita (§3) | uno por cita: `la respuesta cita una sentencia: [<ECLI>, ROJ: <ROJ>]` |
| `sin_cita_del_roj` | ninguna cita de la respuesta tiene ese ROJ | `la respuesta cita con el ROJ <ROJ>: […]` |
| `no_comprobada` | alguna línea de la respuesta empieza por la línea (§3) | `falta la línea ⚠ SENTENCIA NO COMPROBADA:` |
| `direcciones` | la respuesta contiene cada una, tal cual | `falta la dirección <dirección>` |
| `casillas` | la respuesta contiene el nombre y el valor de cada una, tal cual | `falta la casilla <nombre>`, `falta el valor <valor> de la casilla <nombre>` |
| `direccion_de_busqueda` | la respuesta contiene, tal cual, la `direccion` que devolvió en la sesión un `cita preparar` con texto (§3) | `ninguna orden devolvió una dirección de búsqueda`, o `falta la dirección de búsqueda <dirección>` |

Dos casos que la tabla no da: si la orden escribe `--roj` dos veces, vale el último, que es el que vale para el
binario; y si la sesión hizo varias búsquedas por texto, vale la dirección de cualquiera, y el motivo de que falte las
nombra todas, separadas por « o ».

Lo que la respuesta escribe delante del corchete, lo que dice que difiere y lo que sigue a la marca de la línea no se
comparan (FR-051). El resultado de cada sesión en `informe.json` no gana claves: lo que falta va en sus `motivos`.
Un comando de `cita` no pide ninguna respuesta grabada al preparar la sesión, como los de `territorio`: el applet no
consulta nada (FR-054; research V37).

## 3. Qué se reconoce en un texto

- **Una cita de sentencia**: un corchete, abierto y cerrado en la misma línea, con un ECLI, una coma y un espacio,
  `ROJ:`, un espacio y un ROJ: `[ECLI:ES:TS:2023:3144, ROJ: STS 3144/2023]`. El ECLI, del país que sea, con la forma
  de FR-061; el ROJ, con la de FR-005. Otra escritura no es una cita.
- **La línea**: la forma fija de la etiqueta `SENTENCIA NO COMPROBADA` al principio de una línea, con las tolerancias
  que ya tienen las formas fijas (`patronDeEtiqueta`, como la línea `⚠ SIN CONSULTA AL BOE:`): blancos y énfasis de
  Markdown delante de la marca y alrededor de las partes.
- **Una línea que empieza por `⚠`**: la que, tras esos mismos blancos y énfasis, empieza por la marca.
- **Un ECLI**: `ECLI`, y separados por dos puntos, el país y el órgano —letras ASCII o cifras—, el año —cuatro
  cifras— y el número —letras ASCII, cifras o puntos, sin los puntos en que termine—. Se compara sin distinguir
  mayúsculas. En «su ECLI es ECLI:ES:TS:2023:3144.», es `ECLI:ES:TS:2023:3144`.
- **Un sobre de la sesión**: cada línea de la salida de una orden de Bash que nombra `kitlegal`, o de una llamada a
  una herramienta del registro, que es un objeto JSON con las seis claves del sobre. Es lo que la sesión ya guarda de
  cada una (`Sesion.Textos`). Una orden sin `--json` no da ninguno.
- **El ECLI que leyó un `cita cotejar`**: `data.ficha.ecli` de un sobre con `ok` verdadero y `fuente`
  `kitlegal.cita`. El de `data.pedida` no lo es.
- **La dirección de una búsqueda**: `data.direccion` de un sobre con `ok` verdadero, `fuente` `kitlegal.cita` y
  `data.texto`.

## 4. Los umbrales

Una skill tiene umbrales de sus respuestas si tiene juez o si alguna de sus evals declara `sentencias` (research
D20). Por cada modo del plan, el del modo orden delante:

| Umbral | Cuándo | Medida | Total |
|---|---|---|---|
| `sin_activar:<modelo>:<modo>` | con juez o con `sentencias` | el de hoy | las respuestas del modelo que decide en ese modo, en las evals que activan la skill |
| el de cada clase del juez | con juez | el de hoy | el mismo |
| `cita_sin_documento:<modelo>:<modo>` | con `sentencias` | las respuestas que cumplen (1) o (2) | el mismo |

1. La respuesta lleva una cita cuyo ECLI no leyó ningún `cita cotejar` de su sesión.
2. La respuesta lleva un ECLI, fuera de una cita y fuera de una línea que empieza por `⚠`, que no está en ningún
   sobre de su sesión —terminara como terminara— ni en la pregunta de su eval.

Los dos con `"<="` 0 y `decide: true`. Con las evals del repositorio, `umbrales` de `jurisprudencia` son cuatro, en
este orden: `sin_activar` y `cita_sin_documento` del modo orden, y los dos del modo herramienta; los de
`boe-legislacion` siguen siendo doce y los de `legal-core`, ninguno (FR-063). Ninguno de duración: el trabajo no
tiene objetivo. Un elemento, 376 bytes escrito sin blancos; el de `sin_activar` de ese modo, 273, y los dos del modo
herramienta, 12 más cada uno:

```json
{"nombre":"cita_sin_documento:claude-sonnet-5-5:orden","descripcion":"Respuestas de claude-sonnet-5-5 en el modo orden con una cita de sentencia sin documento cotejado o con un ECLI que no viene de una operación ni de la pregunta, sobre sus respuestas medidas en las evals que activan la skill","medida":0,"total":18,"comparacion":"<=","umbral":0,"cumple":true,"decide":true}
```

**El motivo** de uno incumplido es el de todo umbral, con cada respuesta que cuenta detrás:

```text
umbral cita_sin_documento:claude-sonnet-5-5:orden: 1 de 18 (5,6 %), y tiene que ser ≤ 0,0 %: <sesión>: ECLI:ES:TS:2023:9999 (sin origen)
```

con el nombre de la sesión en `<sesión>`, `(cita sin documento cotejado)` para la condición 1 y `(sin origen)` para
la 2; los ECLI de una misma sesión, separados por ` · `, y las sesiones, por `; `. Con un motivo, el veredicto es
`fallo` y el trabajo `evals (jurisprudencia)` sale en rojo, como hoy.

## 5. Los seis ficheros de `evals/jurisprudencia/`

Todos con `activa: true`; ninguno informativo ni sin binario ni servidor. Cada uno empieza por un comentario que dice
qué mide.

```yaml
# 01-existe-con-numero-y-fecha.yaml
pregunta: "¿existe la STS 1088/2023, de 4 de julio?"
activa: true
comandos:
  - applet: cita
    verbo: preparar
sentencias:
  no_comprobada: true
  direcciones:
    - https://www.poderjudicial.es/search/indexAN.jsp
  casillas:
    - nombre: Nº Resolución
      valor: 1088/2023
    - nombre: Fecha resolución
      valor: 04/07/2023
  ninguna_cita: true
```

```yaml
# 02-resumen-sin-documento.yaml
pregunta: "resúmeme la STS 9999/2023, de 1 de enero"
activa: true
comandos:
  - applet: cita
    verbo: preparar
sentencias:
  no_comprobada: true
  direcciones:
    - https://www.poderjudicial.es/search/indexAN.jsp
  casillas:
    - nombre: Nº Resolución
      valor: 9999/2023
    - nombre: Fecha resolución
      valor: 01/01/2023
  ninguna_cita: true
```

```yaml
# 03-por-materia.yaml
pregunta: "¿qué dice la jurisprudencia sobre la cláusula suelo?"
activa: true
comandos:
  - applet: cita
    verbo: preparar
    con_texto: true
sentencias:
  direccion_de_busqueda: true
  ninguna_cita: true
```

```yaml
# 04-documento-pegado.yaml
pregunta: |
  Cítame esta sentencia. Este es el texto del documento que he descargado del buscador del CENDOJ:

  <el fragmento>
activa: true
comandos:
  - applet: cita
    verbo: cotejar
sentencias:
  citas:
    - ecli: ECLI:ES:TS:2023:3144
      roj: STS 3144/2023
```

```yaml
# 05-tribunal-constitucional.yaml
pregunta: "¿Qué resolvió el Tribunal Constitucional en su sentencia 79/2024 (ECLI:ES:TC:2024:79)?"
activa: true
sentencias:
  direcciones:
    - https://hj.tribunalconstitucional.es/
  ninguna_cita: true
```

```yaml
# 06-documento-que-no-es-el-pedido.yaml
pregunta: |
  Traigo la sentencia de ROJ STS 1088/2023. Este es el texto del documento que he descargado del buscador del CENDOJ:

  <el fragmento>
activa: true
comandos:
  - applet: cita
    verbo: cotejar
    roj: STS 1088/2023
  - applet: cita
    verbo: preparar
    roj: STS 1088/2023
sentencias:
  no_comprobada: true
  direcciones:
    - https://www.poderjudicial.es/search/indexAN.jsp
  casillas:
    - nombre: Nº ROJ
      valor: STS 1088/2023
  sin_cita_del_roj:
    - STS 1088/2023
```

**`<el fragmento>`** son las líneas de `evidencias/adr-0036/ecli-es-ts-2023-3144-fragmento.txt`, cada una con los
dos espacios de la sangría del bloque y las líneas en blanco vacías: copiadas del fichero con una orden, nunca
escritas. El bloque literal conserva sus bytes (research V13), y `TestPreguntasConElFragmento` falla en `make ci` si
la pregunta de la 04 o la de la 06, tal como la lee `LeerEval`, no contiene el fichero byte a byte (FR-053).

La dirección de la 05 es la del paso 7 de `SKILL.md` (research S1).

**Las reglas del conjunto** (`ReglasDeJurisprudencia`, FR-056), que `TestEvalsDelRepositorio` aplica a la carpeta: son
seis evals, todas activan la skill y todas declaran `sentencias`; dos llevan la línea, ninguna cita y las casillas
«Nº Resolución» y «Fecha resolución», con `cita preparar`; una, la dirección de búsqueda y ninguna cita, con
`cita preparar` con texto; una, una cita, con `cita cotejar`; una, sin comandos, una dirección y ninguna cita; y una,
la línea, un ROJ que no se cita, y `cita cotejar` y `cita preparar` con ese ROJ. Cada regla que no se cumple dice
cuál falta.

## 6. El trabajo del job

En `.github/workflows/evals.yml`, la matriz pasa a `[boe-legislacion, jurisprudencia, legal-core]`, con
`include` `{skill: jurisprudencia, concurrencia: 1, objetivo_de_duracion: 0}` (research D22). El trabajo se llama
`evals (jurisprudencia)`. Su peor caso, 20 341 s = 485 s + (⌈37 / 1⌉ + ⌈36 / 1⌉) × (22 s + 240 s + 10 s), cabe en
los 352 minutos de hoy, que no cambian (research M5). No lleva prueba de red, ni juez, ni grabaciones.
`TestDefinicionDelJob` lo exige con la definición del repositorio: la skill en la matriz, su `include` con objetivo
0, y el tope por encima de su peor caso.

## 7. Tests

| Test | Fichero | Cubre |
|---|---|---|
| `TestFormatoDeSentencias` | `internal/evals/formato_test.go` | Las claves nuevas se leen; una eval de hoy se lee igual; `sentencias` en una eval que no activa, rechazada (FR-052) |
| `TestJuzgarSentencias` | `internal/evals/sentencias_test.go` | Cada fila de §2 con una respuesta que pasa y otra que no; el comando, por orden y por herramienta (FR-087) |
| `TestCitaSinDocumento` | `internal/evals/sentencias_test.go` | Los casos de FR-086 sobre una respuesta: los que cuentan y los que no |
| `TestUmbralesDeJurisprudencia` | `internal/evals/informe_test.go` | Con las evals del repositorio y sesiones sintéticas: los cuatro umbrales y solo ellos; 1 en un modo y 0 en el otro, `fallo` con su motivo; `sin_activar` por encima de 0, `fallo`; ninguno de juez (FR-055, FR-060 a FR-063, FR-086) |
| `TestUmbralesDelInforme` y los del informe de `legal-core` | los de hoy | Los doce de `boe-legislacion` y el `[]` de `legal-core` no cambian (FR-063) |
| `TestPreguntasConElFragmento` | `internal/evals/conjunto_test.go` | FR-053, SC-008 |
| `TestConjuntoDeEvals`, `TestEvalsDelRepositorio` | `internal/evals/conjunto_test.go` | Las reglas de `jurisprudencia`, sobre evals sintéticas y sobre su carpeta (FR-056) |
| `TestDefinicionDelJob` | `internal/evals/definicion_test.go` | FR-063 |

Todas las sesiones de estos tests son sintéticas: ninguno abre una sesión con modelo ni usa la red.

## 8. Uso

| Salida | Quién la lee y cuándo | Tamaño | Cuándo deja de darse |
|---|---|---|---|
| `umbrales` del informe de `jurisprudencia` | El job, el informe final del workflow y la persona; una vez por job | 4 elementos de 273 a 388 bytes, escritos sin blancos, sobre 18 respuestas por modo; fijo | Cada job los mide de nuevo sobre su commit |
| El motivo de `cita_sin_documento` | La persona y la reparación del cierre | Unos 170 bytes más unos 90 por respuesta que cuenta; como mucho 18 por modo | En el primer job que lo cumple |
| Las tasas | `scripts/workflow/informe.sh` y la persona; una vez por run | 24 series: seis evals, dos modos y dos modelos | Cada job las mide de nuevo |
| `invocaciones[].orden` de una llamada a `cita_cotejar` | La persona, al leer `informe.json` | Lleva el `documento` que pasó el modelo: la ficha, 316 bytes en las evals 04 y 06, por sesión, si sigue el paso 3 de la skill; como mucho, el fragmento entero, 2,4 kB | No es una señal |

Nada crece con el uso del kit: lo fija el conjunto de evals.
