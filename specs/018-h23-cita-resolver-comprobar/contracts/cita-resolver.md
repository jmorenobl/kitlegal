# Contrato · `kitlegal cita resolver` y la herramienta `cita_resolver`

Lo que ve quien invoca el verbo: su gramática, lo que pide a la fuente, su `data`, sus códigos, lo que guarda y su
tamaño. Los bytes son de M3 y M4 de [research.md](../research.md), con una URL de documento supuesta de 93 caracteres:
la real no se ha visto (S1).

## 1. Invocación

```text
kitlegal cita resolver <ecli>
kitlegal cita resolver --roj "<roj>" [--fecha <AAAA-MM-DD>]
kitlegal cita resolver --resolucion <número/año> --fecha <AAAA-MM-DD>
```

El applet `cita` tiene un solo verbo, `resolver`, sin verbo por omisión, y hereda las ocho banderas globales (FR-001).
Los argumentos, con el nombre que llevan también en la herramienta:

| Nombre | Cómo se da | Forma |
|---|---|---|
| `ecli` | de posición, opcional | `ECLI:ES:<órgano>:<año>:<número>`: órgano de 1 a 7 letras mayúsculas o cifras, la primera una letra; año de cuatro cifras; número de 1 a 25 letras mayúsculas, cifras o puntos (FR-003) |
| `roj` | `--roj` | una o más palabras de letras mayúsculas, de la A a la Z, separadas por un espacio; un espacio, cifras, `/` y cuatro cifras: `STS 3144/2023` (FR-004) |
| `resolucion` | `--resolucion` | cifras, `/` y cuatro cifras: `1088/2023` (FR-005) |
| `fecha` | `--fecha` | `AAAA-MM-DD`, de un día que existe (FR-005) |

Se valida todo antes de abrir la caché o pedir nada, en este orden, y lo primero que falla es un error de la clase
`argumentos` (código 2) cuyo mensaje dice qué falla (FR-002 a FR-005):

1. hay exactamente una forma de referencia: ni ninguna, ni dos;
2. un ECLI tiene sus cinco partes y es español; no lleva `--fecha`, tampoco si es del Tribunal Constitucional;
3. un ROJ tiene su forma; su `--fecha`, si la lleva, la suya;
4. `--resolucion` tiene su forma y lleva `--fecha`, con la suya.

No se recorta ni se pasa a mayúsculas nada: `ecli:es:ts:2023:3144`, ` ECLI:ES:TS:2023:3144` y `ROJ: STS 3144/2023`
terminan con 2. El reconocimiento de ECLI y de ROJ es de `internal/core/ids` (`AnalizarECLI`, `AnalizarROJ`), con su
fuzz (FR-113); la referencia la compone `cendoj.NuevaReferencia`.

## 2. Lo que pide

Tras validar, por este orden:

1. **Un ECLI cuyo órgano es `TC`**: no se pide nada ni se abre la caché; termina con 0 y el `data` de §3.3 (FR-015).
2. **La caché**, con la clave de §5. Una entrada vigente se sirve tal cual: una entrega, con 0; un «no encontrado»,
   con 3; las dos con la `fecha_consulta` y la `url` de la consulta que la guardó, sin construir el cliente.
3. Con `--offline` y sin entrada vigente, `fuente-no-disponible` (4), sin pedir nada (FR-017).
4. Con `--dry-run` y sin entrada vigente, nada sale: `Resultado.Ensayo` lleva `GET <página>` y `POST <formulario>`, y
   ni la caché ni el grafo cambian (FR-019).
5. **La fuente**, en una `Consulta` de `internal/httpx`
   ([httpx-formulario.md](./httpx-formulario.md)): la página del buscador y, si responde 200, el formulario con el
   campo de la referencia ([fuente-cendoj-y-grabacion.md §2](./fuente-cendoj-y-grabacion.md)). Con el `robots.txt` del
   sitio, que pide el cliente, son como mucho tres peticiones, separadas 5 s y sin reintentos (FR-020, FR-021).
6. Lo que la fuente devuelve se filtra por `--fecha` (FR-011): solo quedan las resoluciones cuya fecha es la pedida.

## 3. `data`

### 3.1 Forma

| Clave | Tipo | Qué es |
|---|---|---|
| `resoluciones` | lista de 0 a 10 objetos | las resoluciones encontradas que son la pedida, en el orden de la fuente |
| `resoluciones[].ecli` | texto no vacío | su ECLI |
| `resoluciones[].roj` | texto no vacío | su ROJ |
| `resoluciones[].organo` | texto | órgano y sala, como los da la fuente |
| `resoluciones[].fecha` | texto no vacío, `AAAA-MM-DD` | su fecha |
| `resoluciones[].numero_resolucion` | texto | su número de resolución |
| `resoluciones[].numero_recurso` | texto | su número de recurso |
| `resoluciones[].ponente` | texto | su ponente |
| `resoluciones[].url` | URI absoluto | la dirección de su documento que da el CENDOJ; si la da relativa, resuelta sobre la del formulario |
| `pagina_completa` | booleano | verdadero si la fuente devolvió 10 resultados, antes de filtrar por fecha: puede haber más resoluciones que las que lleva `resoluciones`; falso, están todas (FR-014) |
| `cobertura` | `cubierto` o `tribunal-constitucional-no-cubierto` | si la referencia es de un órgano que está en la fuente que consulta `cita resolver` (FR-015) |

`organo`, `numero_resolucion`, `numero_recurso` y `ponente` van vacíos si la fuente no los da; los otros cuatro nunca
(FR-010, FR-022). `data` no lleva el resumen del CENDOJ ni ningún otro texto de la resolución.

### 3.2 Una entrega

`kitlegal cita resolver ECLI:ES:TS:2023:3144 --json`, código 0. 625 bytes en una línea; aquí, con saltos para leerlo:

```json
{"ok":true,"fuente":"cendoj.jurisprudencia","url":"https://www.poderjudicial.es/search/search.action",
 "fecha_consulta":"2026-10-06T10:00:00.123456789Z",
 "hash":"sha256:0000000000000000000000000000000000000000000000000000000000000000",
 "data":{"resoluciones":[{"ecli":"ECLI:ES:TS:2023:3144","roj":"STS 3144/2023",
   "organo":"Tribunal Supremo. Sala de lo Civil","fecha":"2023-07-04","numero_resolucion":"1088/2023",
   "numero_recurso":"4703/2019","ponente":"PEDRO JOSE VELA TORRES",
   "url":"https://www.poderjudicial.es/search/AN/openDocument/0123456789abcdef0123456789abcdef/20230714"}],
  "pagina_completa":false,"cobertura":"cubierto"}}
```

`url` del sobre es la del formulario; `fecha_consulta`, el instante en que salió el envío. Las tres formas de la misma
sentencia dan el mismo `data`.

### 3.3 El Tribunal Constitucional

`kitlegal cita resolver ECLI:ES:TC:2024:79 --json`, código 0, 297 bytes:

```json
{"ok":true,"fuente":"kitlegal.cita","url":"kitlegal:applet/cita","fecha_consulta":"2026-10-06T10:00:00.123456789Z",
 "hash":"sha256:0000000000000000000000000000000000000000000000000000000000000000",
 "data":{"resoluciones":[],"pagina_completa":false,"cobertura":"tribunal-constitucional-no-cubierto"}}
```

No repite el ECLI pedido: el único sitio de un sobre correcto en el que hay un ECLI es una resolución entregada.

### 3.4 «No encontrado»

`kitlegal cita resolver --roj "STS 3144/2023" --fecha 2023-01-01 --json`, código 3, 422 bytes:

```json
{"ok":false,"fuente":"cendoj.jurisprudencia","url":"https://www.poderjudicial.es/search/search.action",
 "fecha_consulta":"2026-10-06T10:00:00.123456789Z",
 "hash":"sha256:0000000000000000000000000000000000000000000000000000000000000000",
 "data":{"clase":"no-encontrado","mensaje":"no se ha encontrado ninguna resolución con el ROJ STS 3144/2023 y la fecha 2023-01-01: ninguna de las encontradas con ese ROJ tiene esa fecha"}}
```

El mensaje nombra la forma, el valor y la fecha pedidos. De lo encontrado con otra fecha no lleva nada (FR-012). Con
cero resultados de la fuente: «no se ha encontrado ninguna resolución con el ECLI ECLI:ES:TS:2023:999999».

## 4. Códigos

Filas del ADR 0023; ningún código nuevo. `cita resolver` nunca termina con 6 ni con 7 (FR-020).

| Situación | `data.clase` | Código | Se guarda |
|---|---|---|---|
| Referencia mal dada (§1) | `argumentos` | 2 | no; no sale ninguna petición |
| Al menos una resolución que es la pedida | — | 0 | caché, 30 días; grafo |
| ECLI del Tribunal Constitucional | — | 0 | no |
| La fuente responde «No se ha encontrado ningún resultado», o ninguna encontrada tiene la fecha pedida | `no-encontrado` | 3 | caché, 30 días: solo el mensaje |
| La página o el formulario responden con un estado que no es 200 ni 5xx —403, 404, 410, 400, 429, una redirección—; o con un 200 que no es ninguna de las dos respuestas reconocidas; o con una lista con un resultado sin ECLI, ROJ, fecha o URL | `limite-o-tos` | 5 | no; no se pide nada más en esa invocación |
| El `robots.txt` no se obtiene sin que venza el plazo, o deniega la ruta | `limite-o-tos` | 5 | no |
| Un 5xx, un fallo de conexión, el plazo vencido o el turno que no cabe en él, en la página o en el formulario; `--offline` sin entrada vigente | `fuente-no-disponible` | 4 | no |
| Cualquier otra cosa: la caché, el grafo o una grabación que no se dejan leer | `inesperado` | 1 | regla genérica |

El «no encontrado» sale solo de la respuesta reconocida o del filtro por fecha, nunca de un estado (FR-022). Un 403 en
la página deja el formulario sin enviar (FR-016).

## 5. Caché

| | |
|---|---|
| Clave | `cendoj.jurisprudencia\|1\|resolver\|<forma>\|<valor>\|<fecha>`; `forma` es `ecli`, `roj` o `resolucion`; `fecha`, la pedida o vacía |
| Contenido de una entrega | `{"fecha_consulta":…,"url":…,"datos":<data>}`; 502 bytes con una resolución |
| Contenido de un «no encontrado» | `{"fecha_consulta":…,"url":…,"no_encontrado":"<mensaje>"}`; 270 bytes |
| Vigencia | 30 días (2 592 000 s) |

La misma forma con el mismo valor y la misma fecha es la misma consulta; cualquier otra es otra consulta y va a la
fuente (FR-040). Vencida la entrada, la consulta vuelve a la fuente, y con `--offline` termina con 4. Nunca se guardan
la página de resultados, el resumen, texto de la resolución, un fallo 4 o 5 ni la declaración de §3.3 (FR-041).

## 6. Grafo

Por cada resolución de una entrega, un nodo: id, su ECLI; tipo, `Resolucion`; datos, `ecli`, `roj`, `organo`, `fecha`,
`numero_resolucion`, `numero_recurso`, `ponente` y `url` (318 bytes); con la procedencia del sobre y la vigencia de la
caché. Sin aristas, sin texto y sin ningún nodo `Persona` (FR-042). Un fallo, la declaración de §3.3, `--no-graph` y
`--dry-run` no entregan nada. `graph show <ECLI>` lo enseña; `graph check` no da ningún hallazgo de él (FR-043).

## 7. Sin `--json`

`Resultado.Legible`, de los mismos datos (FR-018):

```text
ECLI:ES:TS:2023:3144
  ROJ: STS 3144/2023
  Órgano: Tribunal Supremo. Sala de lo Civil
  Fecha: 2023-07-04
  Resolución: 1088/2023
  Recurso: 4703/2019
  Ponente: PEDRO JOSE VELA TORRES
  Documento: https://www.poderjudicial.es/search/AN/openDocument/0123456789abcdef0123456789abcdef/20230714
```

Un bloque por resolución, separados por una línea en blanco. Con `pagina_completa` verdadero, una línea final: «La
fuente devolvió una página completa de 10 resoluciones: puede haber más.» En el caso de §3.3, una sola línea: «El
Tribunal Constitucional no está en la fuente que consulta cita resolver.»

## 8. La herramienta `cita_resolver`

Sale del registro, sin tocar el applet `mcp` ni el paso que empaqueta (FR-050, FR-051). Sus argumentos son los cuatro
de §1, todos opcionales para el esquema; lo que §1 exige lo dice el verbo, con su sobre de fallo:

```json
{"name":"cita_resolver","arguments":{"resolucion":"1088/2023","fecha":"2023-07-04"}}
```

equivale a `kitlegal cita resolver --resolucion=1088/2023 --fecha=2023-07-04 --json`. El resultado es el sobre de §3; en
un fallo, va marcado como error con su `data.clase`. Las llamadas comparten el ritmo del servidor, cada una con su
plazo (FR-024).

## 9. `--describe` y el esquema publicado

`kitlegal cita resolver --describe` da el documento del verbo. En su `entrada`, `roj`, `resolucion` y `fecha` llevan
`"title"` con su escritura (`"--roj"`, `"--resolucion"`, `"--fecha"`); `ecli` no lo lleva: va por su posición
(research D16). Su parte publicada es `schemas/resolucion.json`, y `make schema-check` falla si difieren.

La marca no es de este verbo: la lleva toda bandera propia de todo verbo. Los tres que ya tenían alguna la ganan con
este hito, y es lo único que cambia de lo que emite un verbo de hoy:

| Verbo | Propiedades de su `entrada` que ganan `"title"` |
|---|---|
| `skills install` | `global` (`"--global"`), `host` (`"--host"`), `dir` (`"--dir"`) |
| `skills list` y `skills doctor` | `global` (`"--global"`), `dir` (`"--dir"`) |

Con ellos cambia su parte publicada, `schemas/instalacion.json`: siete líneas más, 234 bytes, y nada de lo que valida
(research M5). Se regenera en la misma tarea que cambia `--describe`. Los verbos de `boe`, `territorio`, `graph` y `mcp`
no tienen banderas propias, y ni su documento ni su esquema publicado cambian.

La fila de la tabla de comandos que sale del documento de `cita resolver`:

```text
| `kitlegal cita resolver [<ecli>] [--roj=<roj>] [--resolucion=<resolucion>] [--fecha=<fecha>]` | `cita_resolver` | … | objeto con `resoluciones`, `pagina_completa`, `cobertura` |
```

Los esquemas de la herramienta no llevan `title`.

## 10. Uso, de fuera adentro

| Quién lo pide | Cuántas veces | Tamaño | Cuándo deja de darse |
|---|---|---|---|
| La skill `jurisprudencia`, de una en una (FR-061): una por sentencia dada por su ECLI o su ROJ; una o dos por sentencia escrita con número y fecha | una respuesta que comprueba cinco sentencias, de 5 a 10 | 625 B por entrega de una resolución; como mucho 3 495 B con diez. Diez entregas de una resolución: 6 250 B. Por herramienta el sobre va dos veces en el mensaje: 1 250 B por llamada | no crece con lo acumulado: lo acotan la referencia y la página de 10 |
| Una persona, desde la terminal | una | la salida de §7, unas 8 líneas por resolución | — |

La caché tiene una entrada por consulta distinta hecha en los últimos 30 días: con cien sentencias comprobadas en un
mes, cada una por una o dos formas, entre 100 y 200 entradas de 270 a 502 bytes, menos de 100 kB; las vencidas dejan de
servirse. El grafo tiene un nodo por resolución comprobada alguna vez, de 318 bytes de datos: cientos en meses de uso,
y nadie los recibe salvo quien pide uno con `graph show`.

Tiempo: una consulta que no está en la caché hace tres peticiones separadas 5 s, al menos 10 s, dentro del plazo por
omisión de 30 s; cinco sentencias, al menos 50 s, y hasta el doble con el segundo intento de cada una.
