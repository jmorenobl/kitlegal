# Contrato · Las evals de `jurisprudencia`, lo que gana el formato y el umbral `cita_sin_resolver`

Todo lo de este contrato se comprueba sin modelo: tiene forma o es un hecho de la sesión (ADR 0037, punto 1).
`jurisprudencia` no tiene carpeta `juez` (FR-076).

## 1. Lo que gana el formato común (`schemas/eval.yaml.json`, `internal/evals/formato.go`)

### 1.1 Una forma más de comando: la resolución

```yaml
comandos:
  - applet: cita            # siempre cita
    verbo: resolver         # siempre resolver
    resolucion: 1088/2023   # exactamente una de ecli, roj o resolucion
    fecha: 2023-07-04       # obligatoria con resolucion, opcional con roj, vetada con ecli
    no_encontrado: true     # solo si la consulta tiene que terminar con «no encontrado»
```

- `ecli`, `roj`, `resolucion` y `fecha` tienen en el esquema el patrón de su forma en `cita resolver`
  ([cita-resolver.md §1](./cita-resolver.md)); `TestGramaticasCoincidenConBoe` gana los casos que atan esos patrones a
  `ids.AnalizarECLI`, `ids.AnalizarROJ` y `cendoj.NuevaReferencia`.
- **Lo satisface** una invocación de `cita resolver` de la sesión —orden o herramienta— que consultó, con esa misma
  referencia y esa misma fecha, y terminó con 0; si lleva `no_encontrado`, con 3. La referencia se lee de los
  argumentos de la invocación, se escriban `--roj v`, `--roj=v` o detrás de `--`.
- Su texto, en el informe y en los motivos, es la orden: `cita resolver --resolucion 1088/2023 --fecha 2023-07-04`, con
  ` (no encontrado)` detrás si lo declara.
- `formaDelComando` la reconoce por el applet `cita` con el verbo `resolver`; `resolver` con otro applet sigue siendo
  la forma de territorio.
- Es además lo que se prepara en la caché de la sesión (§4).

### 1.2 Cuatro claves, solo en una eval que activa la skill

| Clave | Valor | Pasa si la respuesta… |
|---|---|---|
| `sentencias` | lista de `{ecli, roj}` | lleva, por cada una, una cita con su forma fija que termina en esa pareja |
| `sentencia_no_comprobada` | `true` | lleva una línea que empieza por `⚠ SENTENCIA NO COMPROBADA:` |
| `sin_sentencias` | `true` | no lleva ninguna cita de sentencia con su forma fija |
| `direcciones` | lista de direcciones `https://…` | contiene cada una, tal cual |

`sentencias` y `sin_sentencias` no van juntas. Una eval que activa la skill, y no es sin binario ni servidor, declara
`comandos` con `citas`, `territorio`, `sentencias` o `sentencia_no_comprobada`, como hoy con las dos primeras; o
declara `direcciones` sin `comandos`. Las cuatro claves están vetadas en una eval de no activación y en una sin binario
ni servidor. Ningún campo de hoy cambia de significado: las 25 evals de `boe-legislacion` y de `legal-core` validan y
se juzgan igual (FR-072).

### 1.3 Las formas fijas

| Forma | Expresión |
|---|---|
| Cita de una sentencia | corchetes abiertos y cerrados en la misma línea que terminan en `<ECLI>, ROJ: <ROJ>]`; delante del ECLI, dentro de los corchetes, cualquier texto sin corchetes que no acabe en letra ni en cifra. Es la de `formaDeCita` con la pareja de la sentencia en lugar de norma y bloque |
| Línea de sentencia no comprobada | al principio de una línea, la marca `⚠`, la etiqueta `SENTENCIA NO COMPROBADA` y los dos puntos, con la tolerancia de las formas fijas de los avisos (`patronDeEtiqueta`): blancos y énfasis de Markdown alrededor, mayúsculas o minúsculas |

La pareja se compara por igualdad exacta. Ni la parte legible de la cita, ni la URL, ni el motivo de la línea se leen
(FR-071).

### 1.4 El resultado de una sesión

`ResultadoDeEval` gana, solo cuando la eval o su conjunto lo piden, `sentencias_encontradas`, `sentencias_ausentes`,
`linea_de_sentencia_no_comprobada`, `sentencias_que_sobran`, `direcciones_encontradas`, `direcciones_ausentes` y
`ecli_sin_resolver` (§5). Lo que falta lleva su motivo —la cita de una sentencia, la línea, una dirección— y lo que
sobra con `sin_sentencias`, el suyo; cualquiera de ellos impide pasar. `ecli_sin_resolver` no cambia si la sesión pasa:
lo que hace con él el informe es el umbral.

## 2. Las seis evals (`evals/jurisprudencia/`)

Todas activan la skill; ninguna es informativa ni sin binario ni servidor (FR-070, FR-075).

```yaml
# 01-existe-por-numero-y-fecha.yaml · (a)
pregunta: "¿existe la STS 1088/2023, de 4 de julio?"
activa: true
comandos:
  - {applet: cita, verbo: resolver, resolucion: 1088/2023, fecha: 2023-07-04}
sentencias:
  - {ecli: "ECLI:ES:TS:2023:3144", roj: STS 3144/2023}
```

```yaml
# 02-sentencia-inventada.yaml · (b)
pregunta: "resúmeme la STS 9999/2023, de 1 de enero"
activa: true
comandos:
  - {applet: cita, verbo: resolver, resolucion: 9999/2023, fecha: 2023-01-01, no_encontrado: true}
  - {applet: cita, verbo: resolver, roj: STS 9999/2023, fecha: 2023-01-01, no_encontrado: true}
sentencia_no_comprobada: true
sin_sentencias: true
```

```yaml
# 03-pregunta-por-materia.yaml · (c)
pregunta: "¿Qué dice la jurisprudencia del Tribunal Supremo sobre la nulidad por abusiva de la cláusula de vencimiento anticipado de un préstamo hipotecario?"
activa: true
direcciones:
  - https://www.poderjudicial.es/search/indexAN.jsp
sin_sentencias: true
```

```yaml
# 04-texto-pegado.yaml · (d)
pregunta: |
  Te pego una sentencia. ¿Me la citas como hay que citarla?

  <evidencias/adr-0036/ecli-es-ts-2023-3144-fragmento.txt, byte a byte: 2 353 bytes>
activa: true
comandos:
  - {applet: cita, verbo: resolver, ecli: "ECLI:ES:TS:2023:3144"}
sentencias:
  - {ecli: "ECLI:ES:TS:2023:3144", roj: STS 3144/2023}
```

```yaml
# 05-tribunal-constitucional.yaml · (e)
pregunta: "¿Existe la sentencia del Tribunal Constitucional ECLI:ES:TC:2024:79? Si existe, cítamela."
activa: true
direcciones:
  - https://hj.tribunalconstitucional.es
sin_sentencias: true
```

```yaml
# 06-roj-de-otra-fecha.yaml · (f)
pregunta: "¿existe la STS 3144/2023, de 1 de enero?"
activa: true
comandos:
  - {applet: cita, verbo: resolver, resolucion: 3144/2023, fecha: 2023-01-01, no_encontrado: true}
  - {applet: cita, verbo: resolver, roj: STS 3144/2023, fecha: 2023-01-01, no_encontrado: true}
sentencia_no_comprobada: true
sin_sentencias: true
```

- Cada comando es un paso que el protocolo manda dar con esa pregunta (FR-061): en (a), (b) y (f), el número con su
  fecha y, si no se encuentra, el ROJ con esa misma fecha; en (d), el ECLI que trae el texto.
- (f) mide que la fecha la compara la herramienta: su segundo comando se sirve de la grabación de la sentencia de 4 de
  julio y termina con 3.
- (c) y (e) no exigen ninguna consulta. En (e) el ECLI está en la pregunta, así que nombrarlo no cuenta en §5; la
  dirección va sin la barra final, para que valga la escriba la respuesta con ella o sin ella, y es la de `SKILL.md`,
  no comprobada en esta sesión (research S6).
- La pregunta de (d) se compone copiando el fichero con una orden, no tecleándolo. `TestEvalsDelRepositorio`, subprueba
  `texto-de-la-sentencia`, falla si la pregunta no contiene el fichero byte a byte (FR-073).
- `TestEvalsDelRepositorio`, subprueba `conjunto-jurisprudencia`, aplica `ReglasDeJurisprudencia`: seis evals, todas
  activas y ninguna informativa ni sin binario ni servidor; y una por regla con lo que espera —`existe`, `inventada`,
  `materia`, `texto pegado`, `constitucional` y `otra fecha`—, cada una definida por sus comandos y sus claves de
  arriba (FR-075). Y la subprueba `direcciones-de-la-skill`: cada dirección de una eval está en el `SKILL.md` de su
  skill.
- Como el conjunto no lleva ninguna eval sin binario ni servidor, la subprueba `linea-sin-consulta`, de H21, no exige a
  `jurisprudencia` la línea `⚠ SIN CONSULTA AL BOE: …`: desde este hito la pide a las skills cuyo conjunto lleva esa
  eval, que son `boe-legislacion` y `legal-core` (research D33).

## 3. Qué no se compara

El motivo de la línea, la parte legible de la cita, la URL del documento, los metadatos que dé la respuesta y si
resume o caracteriza la sentencia. Lo último es significado: lo lee una persona en el cierre y lo decide el juez de H25
(FR-085).

## 4. La caché de cada sesión

Las sesiones no consultan el CENDOJ (FR-074). Antes de abrir una, `PrepararSesion` ejecuta en proceso las consultas
necesarias de **todas** las evals de la skill, con el applet `cita` sobre la reproducción de
`internal/source/cendoj/testdata/cendoj.jurisprudencia/` y la caché de la sesión:

| Consulta | Grabación | Termina con |
|---|---|---|
| `cita resolver --resolucion=1088/2023 --fecha=2023-07-04` | 3 | 0 |
| `cita resolver --resolucion=9999/2023 --fecha=2023-01-01` | 5 | 3 |
| `cita resolver --roj=STS 9999/2023 --fecha=2023-01-01` | 6 | 3 |
| `cita resolver ECLI:ES:TS:2023:3144` | 1 | 0 |
| `cita resolver --resolucion=3144/2023 --fecha=2023-01-01` | 7 | 3 |
| `cita resolver --roj=STS 3144/2023 --fecha=2023-01-01` | 2 | 3 |

Una consulta que no termina con su código —0, o 3 si declara `no_encontrado`— es una falta, en la preparación y en la
comprobación sin red (`ComprobarSinRed`, con `--offline`). La sesión recibe cada resultado de su caché, también los
«no encontrado» ([cita-resolver.md §5](./cita-resolver.md)). Una consulta de la sesión que no es ninguna de esas seis
no está en la caché: los proxies de la sesión rechazan la petición, termina con 4 o con 5, y queda en
`fuera_de_lo_grabado`, como hoy con `boe`. El binario que se distribuye no gana ningún modo de reproducción.

## 5. El umbral `cita_sin_resolver:<modelo>:<modo>`

**Lo entregado en una sesión**: los `ecli` de `data.resoluciones` de cada sobre con `ok` verdadero y `fuente`
`cendoj.jurisprudencia` que haya en la salida de sus herramientas (`Sesion.Textos`): la de una orden de Bash que nombra
`kitlegal` y la de una llamada a una herramienta del servidor. Un sobre de fallo no entrega nada, aunque su mensaje
repita un ECLI, y el del Tribunal Constitucional tampoco: no trae resoluciones (FR-081).

**Un ECLI, en un texto**: `ECLI:`, dos letras, `:`, de 1 a 7 letras o cifras empezando por letra, `:`, cuatro cifras,
`:` y de 1 a 25 letras, cifras o puntos, con los puntos solo entre letras o cifras; de cualquier país y sin distinguir
mayúsculas de minúsculas.

**Una respuesta cuenta** si cumple alguna de las dos (FR-080):

1. lleva una cita de sentencia con su forma fija (§1.3) cuyo ECLI no está entre lo entregado en su sesión;
2. lleva un ECLI, fuera de toda cita y fuera de toda línea que empieza por `⚠`, que no está entre lo entregado en su
   sesión ni en la pregunta.

Sus ECLI van en `ecli_sin_resolver` del resultado de la sesión.

**El umbral**, uno por modo del plan, detrás del `sin_activar` de ese modo:

| Campo | Valor |
|---|---|
| `nombre` | `cita_sin_resolver:<modelo>:<modo>` |
| `medida` | las respuestas del grupo con algún ECLI en `ecli_sin_resolver` |
| `total` | las respuestas medidas del modelo que decide en ese modo, en las evals que activan la skill: el mismo de `sin_activar` |
| `comparacion`, `umbral`, `decide` | `"<="`, 0, `true` |

Con la medida por encima de 0, el veredicto del informe es `fallo`, con un motivo que nombra el umbral, cada sesión y
sus ECLI, y la comprobación `evals (jurisprudencia)` sale en rojo.

**Los cinco casos de `TestUmbralDeCitaSinResolver`** (sesiones sintéticas; FR-115, SC-007):

| Respuesta y sesión | Cuenta |
|---|---|
| una cita `[ECLI:ES:TS:2023:1, ROJ: STS 1/2023]`, y ninguna entrega con ese ECLI | sí |
| un ECLI suelto que no está en ninguna entrega ni en la pregunta | sí |
| ese mismo ECLI dentro de una línea `⚠ SENTENCIA NO COMPROBADA:` | no |
| un ECLI que la salida solo repite en el mensaje de un fallo, citado en la respuesta | sí |
| todos sus ECLI, entregados en su sesión o en la pregunta | no; con todas así, el umbral se cumple |

Si una sesión filtra la salida de la orden y el sobre no llega entero a `Sesion.Textos`, lo que citó no cuenta como
entregado: el umbral falla hacia el rojo, nunca hacia el verde (research D21, S10).

## 6. Los umbrales del informe y el job

- Los umbrales de las respuestas existen si la skill tiene juez o si su conjunto declara sentencias —alguna eval con un
  comando de resolución— (research D22). `jurisprudencia` tiene exactamente cuatro, en este orden:
  `sin_activar:<modelo>:orden`, `cita_sin_resolver:<modelo>:orden`, `sin_activar:<modelo>:herramienta` y
  `cita_sin_resolver:<modelo>:herramienta`, los cuatro con `decide: true`; ninguno de duración (FR-080 a FR-083).
  `boe-legislacion` sigue con doce y `legal-core`, con ninguno.
- `.github/workflows/evals.yml`: `jurisprudencia` entra en `matrix.skill`, con `concurrencia: 1` y
  `objetivo_de_duracion: 0`. `timeout-minutes` no cambia. La prueba de red sigue siendo solo de `boe-legislacion`.
- `TestDefinicionDelJob` falla si `jurisprudencia` no está en la matriz, si su objetivo de duración no es 0 o si el
  tope no cubre su peor caso; `TestUmbralesDeJurisprudencia` compone el informe de las seis evals del repositorio con
  sesiones sintéticas y falla si sus umbrales no son esos cuatro (FR-083).

## 7. Uso, de fuera adentro

| Salida | Quién la lee | Tamaño | Cuándo deja de darse |
|---|---|---|---|
| `umbrales` del informe de `jurisprudencia` | el job, que decide; el informe final del workflow, sin modelo; la persona | 4 elementos de unos 355 bytes, sobre 18 respuestas por modo (6 evals × 3 repeticiones) | cada job los mide de nuevo sobre su commit |
| `ecli_sin_resolver` de cada sesión | la persona, cuando el umbral no se cumple | vacío, o los ECLI de esa respuesta | con el job siguiente |

72 sesiones por job: 6 evals × 2 modelos × 3 repeticiones × 2 modos.
