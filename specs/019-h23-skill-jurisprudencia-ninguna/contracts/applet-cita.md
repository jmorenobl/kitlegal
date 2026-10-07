# Contrato: el applet `cita`

Lo que fija el plan donde el spec lo deja abierto (FR-004, FR-010, FR-012, FR-020, FR-022, FR-025). Los ejemplos son
salidas del prototipo (research V1, M1), con sus bytes. **Fijo** es lo que este contrato escribe como código: claves,
valores de vocabulario, nombres de casilla, direcciones y la forma de `url`. **Libre** es el texto para la persona
—`motivo`, `explicacion` y el mensaje de un error—: cada uno dice qué tiene que nombrar, y sus bytes los fijan los
golden cuando existan.

## 1. Verbos y argumentos

`kitlegal cita preparar` y `kitlegal cita cotejar`. Ninguno es el verbo por omisión (FR-001): `kitlegal cita` termina
con 2 y nombra los dos.

| Argumento | En la orden | En la herramienta | Verbos | Qué es |
|---|---|---|---|---|
| `ecli` | de posición, opcional | `ecli` | los dos | Un ECLI español (FR-005) |
| `roj` | `--roj <ROJ>` | `roj` | los dos | Un ROJ |
| `resolucion` | `--resolucion <n>/<año>` | `resolucion` | los dos | Un número de resolución; va con `fecha` |
| `fecha` | `--fecha AAAA-MM-DD` | `fecha` | los dos | La fecha de esa resolución |
| `texto` | `--texto <texto>` | `texto` | `preparar` | El texto de una búsqueda por materia |
| `documento` | `--documento <texto>` | `documento` | `cotejar` | El texto del documento, con su ficha |

Todos son cadenas; `documento` conserva los bytes con los que llega (research D12). Un argumento escrito con valor
vacío **no** es un argumento no dado: cuenta como dado y es el error de su fila de §6, en la orden (`--roj ""`,
`--roj=`, un ECLI `""`) y en la llamada (`"roj": ""`). El verbo los distingue porque sus campos son punteros, que el
analizador deja en `nil` cuando el argumento no se escribe (research D3, V38); para `--describe`, para el esquema de
la herramienta y para la tabla de comandos siguen siendo cadenas (research V39). Las banderas globales valen lo que
en cualquier verbo; `--offline` y `--no-graph` no cambian nada.

## 2. El sobre

| Clave | `preparar` | `cotejar` |
|---|---|---|
| `fuente` | `kitlegal.cita` | `kitlegal.cita` |
| `url` | la `direccion` de `data`; `kitlegal:applet/cita` si `data` no lleva ninguna | `kitlegal:documento/sha256:<64 hex>`: el SHA-256 del texto recibido, byte a byte |
| `fecha_consulta` | el instante de la invocación, del reloj del kernel | ídem |
| `hash` | la huella de `data`, como en todo applet | ídem |

En un fallo, `fuente` es `kitlegal.cita` y `url`, `kitlegal:applet/cita`. `fuente` es del espacio reservado a lo que
el binario calcula: dice que no se ha consultado ninguna fuente (FR-004). El mismo texto da siempre la misma `url`, y
uno distinto, otra: para el fragmento de la evidencia, la de su manifiesto,
`kitlegal:documento/sha256:4886e0c8ca9836527ec08d8732b50315640a76803033374af5287b2a8321ee27` (research V12).

## 3. `data` de `cita preparar`

| Clave | Cuándo está | Forma |
|---|---|---|
| `referencia` | con una referencia | `{forma, valor, fecha?}`: `forma` es `ecli`, `roj` o `resolucion`; `valor`, tal como se dio; `fecha`, `AAAA-MM-DD`, solo con `resolucion` |
| `texto` | con `--texto` | el texto, tal como se dio |
| `equivalente` | solo si se deduce (FR-012) | `{forma, valor}`: `forma` es `roj` o `ecli` |
| `cobertura` | solo con un ECLI de órgano `TC` (FR-014) | `{cendoj, motivo}`: `cendoj` es `no-cubierto`, su único valor; `motivo`, libre, dice que el Tribunal Constitucional no está en el CENDOJ |
| `direccion` | si hay algo que abrir | `https://www.poderjudicial.es/search/indexAN.jsp`, o la de la búsqueda por texto (FR-013) |
| `casillas` | siempre | lista de `{nombre, campo?, valor}`, en el orden de FR-011; vacía con `--texto` y fuera de cobertura |

`cobertura` declara lo que queda fuera, y sin esa declaración la clave no está: de una referencia que no es el ECLI
del Tribunal Constitucional, y de un texto, `data` no dice que el CENDOJ lo tenga, porque el binario no lo sabe
(research D4). `--roj "STC 79/2024"` se prepara como cualquier otro ROJ, sin `cobertura` (spec, Edge Cases).

Las casillas, fijas: «ECLI»; «Nº ROJ»; y «Nº Resolución» seguida de «Fecha resolución» con `campo` «Desde» y de
«Fecha resolución» con `campo` «Hasta», las dos con la fecha `dd/mm/aaaa`.

Ejemplos, la salida entera de la orden con `--json`; los bytes, con `fecha_consulta` de seis decimales (research M1):

`kitlegal cita preparar ECLI:ES:TS:2023:3144 --json` · 474 bytes:

```json
{"ok":true,"fuente":"kitlegal.cita","url":"https://www.poderjudicial.es/search/indexAN.jsp","fecha_consulta":"2026-10-07T11:51:51.830392+02:00","hash":"sha256:7822e8eb4dd8b0ff30e558f31b6a0fb8175bcb816a57d5f86e74df04469206b3","data":{"referencia":{"forma":"ecli","valor":"ECLI:ES:TS:2023:3144"},"equivalente":{"forma":"roj","valor":"STS 3144/2023"},"direccion":"https://www.poderjudicial.es/search/indexAN.jsp","casillas":[{"nombre":"ECLI","valor":"ECLI:ES:TS:2023:3144"}]}}
```

`kitlegal cita preparar --resolucion 1088/2023 --fecha 2023-07-04 --json` · 572 bytes; su `data`:

```json
{"referencia":{"forma":"resolucion","valor":"1088/2023","fecha":"2023-07-04"},"direccion":"https://www.poderjudicial.es/search/indexAN.jsp","casillas":[{"nombre":"Nº Resolución","valor":"1088/2023"},{"nombre":"Fecha resolución","campo":"Desde","valor":"04/07/2023"},{"nombre":"Fecha resolución","campo":"Hasta","valor":"04/07/2023"}]}
```

`kitlegal cita preparar --texto "cláusula suelo" --json` · 389 bytes; su `data`:

```json
{"texto":"cláusula suelo","direccion":"https://www.poderjudicial.es/search/sentencias/cl%C3%A1usula%20suelo/1/AN","casillas":[]}
```

`kitlegal cita preparar ECLI:ES:TC:2024:79 --json` · 424 bytes, con `url` `kitlegal:applet/cita`; su `data`:

```json
{"referencia":{"forma":"ecli","valor":"ECLI:ES:TC:2024:79"},"cobertura":{"cendoj":"no-cubierto","motivo":"Las resoluciones del Tribunal Constitucional no están en el CENDOJ: su buscador no las tiene."},"casillas":[]}
```

Con `--roj "STS 3144/2023"`, 470 bytes: una casilla «Nº ROJ» y `equivalente` `{"forma":"ecli","valor":"ECLI:ES:TS:2023:3144"}`.
Sin equivalente —`ECLI:ES:AN:2019:1.2.A`, `--roj "SAP M 1234/2020"`, `--roj "STC 79/2024"`— la clave no está: 422,
412 y 404 bytes. El último, su `data`:

```json
{"referencia":{"forma":"roj","valor":"STC 79/2024"},"direccion":"https://www.poderjudicial.es/search/indexAN.jsp","casillas":[{"nombre":"Nº ROJ","valor":"STC 79/2024"}]}
```

## 4. `data` de `cita cotejar`

| Clave | Cuándo está | Forma |
|---|---|---|
| `ficha` | siempre | `{roj, ecli, organo, fecha, recurso, resolucion, ponente, tipo}`: los ocho datos de FR-021; `fecha`, `AAAA-MM-DD`; los demás, como en el documento, sin los blancos de alrededor |
| `correspondencia` | siempre | `se-corresponden`, `no-se-corresponden` o `no-se-deduce` (FR-023) |
| `pedida` | con una referencia | la referencia, con la forma de §3 |
| `es_la_pedida` | con una referencia | booleano (FR-024) |
| `hallazgos` | siempre | lista; vacía, o con un elemento si `es_la_pedida` es falso |

El hallazgo: `{clase, difiere, cruce?, explicacion}`. `clase` es `documento-distinto`. `difiere` es una lista de
`{dato, pedido, documento}`, con `dato` en `ecli`, `roj`, `resolucion` o `fecha`: uno por dato que difiere, y con
número y fecha, uno o los dos. `cruce`, solo si lo hay: `numero-de-resolucion` —se pidió un ROJ cuyo
`<número>/<año>` es el «Nº de Resolución» del documento— o `numero-del-roj` —se pidió un número de resolución que es
el `<número>/<año>` del ROJ del documento—. `explicacion`, libre: dice que el documento no es el pedido, nombra cada
dato pedido y el del documento y, con cruce, qué es en el documento ese número.

Ejemplos, con el fragmento de la evidencia en la entrada estándar:

`kitlegal cita cotejar --json` · 559 bytes; su `data`:

```json
{"ficha":{"roj":"STS 3144/2023","ecli":"ECLI:ES:TS:2023:3144","organo":"Tribunal Supremo. Sala de lo Civil","fecha":"2023-07-04","recurso":"4703/2019","resolucion":"1088/2023","ponente":"PEDRO JOSE VELA TORRES","tipo":"Sentencia"},"correspondencia":"se-corresponden","hallazgos":[]}
```

`kitlegal cita cotejar --roj "STS 1088/2023" --json` · 949 bytes; termina con 0 y `ok` verdadero; lo que su `data`
añade a la anterior:

```json
"pedida":{"forma":"roj","valor":"STS 1088/2023"},"es_la_pedida":false,"hallazgos":[{"clase":"documento-distinto","difiere":[{"dato":"roj","pedido":"STS 1088/2023","documento":"STS 3144/2023"}],"cruce":"numero-de-resolucion","explicacion":"El documento no es el pedido: se pidió el ROJ STS 1088/2023 y el del documento es STS 3144/2023. 1088/2023 es el número de resolución del documento, no su ROJ."}]
```

Con el documento pedido —por su ECLI, su ROJ o su número con su fecha—, de 628 a 652 bytes: `es_la_pedida` verdadero y
`hallazgos` vacío. Con `--resolucion 1088/2023 --fecha 2023-01-01`, 869 bytes: `difiere` solo con `fecha`
(`2023-01-01` y `2023-07-04`), sin `cruce`. Con `--resolucion 3144/2023 --fecha 2023-07-04`, 1 005 bytes: `difiere`
con `resolucion` y `cruce` `numero-del-roj`.

`data` no lleva nada más del documento: ni «Id Cendoj», «Sede», «Sección» o «Procedimiento», ni su texto (FR-022).
Por eso la ficha sola —las once primeras líneas del fragmento, 316 bytes— da los mismos `data` y `hash` que el
fragmento entero, y otra `url`, la de su texto (research V41): es lo que la skill pasa (§10).

## 5. La lectura de la ficha

Las reglas son las de FR-021, sin añadir ninguna. En el prototipo: el texto se parte por `\n`; de cada línea se
quitan los blancos de los dos extremos, y con ellos el `\r`; la ficha empieza en la primera línea cuyo comienzo es
`Roj` seguido de dos puntos; su valor se parte por ` - ` en el ROJ y el ECLI; cada uno de los otros seis datos es lo
que sigue a su etiqueta y a los dos puntos en la primera línea, desde la siguiente a la de `Roj:`, que empieza por
esa etiqueta. «Nº de Recurso» y «Nº de Resolución» no se confunden: la etiqueta entera tiene que ir delante de los
dos puntos. Casos comprobados: texto delante de la ficha y finales `\r\n`, misma `data`; sin la línea de «Ponente»,
error que la nombra; «Fecha: 31/02/2023», error que la nombra.

## 6. Errores de argumentos

Todos terminan con 2, `ok` falso y `data` `{clase: "argumentos", mensaje}` (FR-006, FR-015, FR-026). El mensaje
nombra lo que esta tabla dice. «Dado» es escrito, también con valor vacío (§1): un argumento vacío cae en la fila de
su argumento, nunca en la de «no dado».

| Caso | El mensaje nombra |
|---|---|
| ECLI mal formado, en minúsculas, con un blanco o vacío | el ECLI entre comillas y qué tiene de malo frente a su forma |
| ECLI de otro país | el ECLI y que no es español |
| ROJ sin su forma, o vacío | el ROJ entre comillas y su forma |
| Número de resolución sin su forma, o vacío | el número y su forma, `<número>/<año>` |
| `--resolucion` sin `--fecha` | que el número necesita su fecha |
| `--fecha`, también vacía, sin `--resolucion` | que la fecha solo vale con el número |
| Fecha sin su forma, vacía o que no es un día | la fecha entre comillas y su forma |
| Más de una forma de referencia, también si alguna va vacía | las formas dadas |
| `preparar`: ni referencia ni `--texto` | lo que se puede dar |
| `preparar`: referencia y `--texto`, también vacío | que son excluyentes |
| `preparar`: `--texto` vacío o solo de blancos | el texto entre comillas |
| `cotejar`: ningún texto: `--documento` vacío o, sin él, una entrada vacía o una llamada de herramienta | por dónde se da |
| `cotejar`: texto sin línea `Roj:` | que falta la ficha |
| `cotejar`: ficha sin un dato, o con uno de los cuatro sin su forma | el dato, con su etiqueta |

La referencia se comprueba antes que el texto. Un error del propio análisis de la línea de órdenes —una bandera que
no existe— es del kernel, como en cualquier verbo. Las invocaciones con un argumento vacío, ejecutadas en el
prototipo como orden y como llamada, en research V40.

## 7. La entrada estándar y las herramientas

- En la orden, sin `--documento`, `cotejar` lee el texto de la entrada estándar hasta su final. Con `--documento`,
  el texto es el suyo y la entrada no se lee, tampoco si va vacío: `kitlegal cita cotejar --documento ""` con un
  documento en la entrada termina con 2, «ningún texto» (FR-020, FR-026).
- El kernel le da la entrada al verbo: la del proceso en una orden, y ninguna en una llamada de herramienta
  (research D12). Una llamada a `cita_cotejar` sin `documento`, o con él vacío, es el error de argumentos «ningún
  texto»; el servidor no lee de su entrada nada más que el protocolo y sigue atendiendo (research V7).
- La llamada devuelve el sobre de su orden: con el mismo texto, los mismos bytes salvo `fecha_consulta` (research
  V10). Un hallazgo no es un error de herramienta; un error de argumentos sí, con su clase (FR-030).
- `mcp.go`, `herramientas.go` e `internal/mcp` no cambian. El servidor pasa a anunciar doce herramientas.

## 8. `--describe`

El documento de cada verbo lleva en `entrada` la anotación `x-banderas`, con los nombres de sus banderas propias en
su orden: `["roj","resolucion","fecha","texto"]` en `preparar` y `["roj","resolucion","fecha","documento"]` en
`cotejar`. La lleva todo verbo con banderas propias —también los tres de `skills`— y ninguno más; no va en el esquema
de entrada de una herramienta (research D14). De ella sale la orden de la tabla de comandos:

```text
kitlegal cita preparar [<ecli>] [--roj <roj>] [--resolucion <resolucion>] [--fecha <fecha>] [--texto <texto>]
kitlegal cita cotejar [<ecli>] [--roj <roj>] [--resolucion <resolucion>] [--fecha <fecha>] [--documento <documento>]
```

## 9. Golden

En `internal/app/testdata/cita/`, el sobre entero con `fecha_consulta` puesta a `0001-01-01T00:00:00Z`, y cada uno
válido contra `schemas/cita.json` (FR-080, FR-081; SC-004):

| Fichero | Invocación |
|---|---|
| `preparar-ecli.json` | `ECLI:ES:TS:2023:3144` |
| `preparar-roj.json` | `--roj "STS 3144/2023"` |
| `preparar-resolucion.json` | `--resolucion 1088/2023 --fecha 2023-07-04` |
| `preparar-texto.json` | `--texto "cláusula suelo"` |
| `cotejar-sin-referencia.json` | el fragmento, sin referencia |
| `cotejar-ecli.json`, `cotejar-roj.json`, `cotejar-resolucion.json` | el fragmento, pedido por sus tres formas |
| `cotejar-roj-cruzado.json` | `--roj "STS 1088/2023"` |
| `cotejar-otra-fecha.json` | `--resolucion 1088/2023 --fecha 2023-01-01` |
| `cotejar-resolucion-cruzada.json` | `--resolucion 3144/2023 --fecha 2023-07-04` |

El test lee el fragmento de `evidencias/adr-0036/`, sin copiarlo.

## 10. Uso

Quién pide cada salida y cuántas veces, en contracts/skill-jurisprudencia.md §3. Ninguna crece con lo acumulado: el
applet no guarda nada, y cada invocación depende solo de sus argumentos y de su texto. Tras meses de uso diario, con
cientos de sentencias preparadas y cotejadas, la salida de la siguiente tiene los bytes de esta tabla:

| Salida | Bytes | Qué la hace crecer |
|---|---|---|
| Una consulta preparada | 389 a 572 | Nada acumulado: como mucho tres casillas; el texto de búsqueda, que va una vez tal cual y dos codificado |
| Un cotejo | 559 a 1 005 | Nada acumulado, ni el tamaño del documento: del texto solo sale la ficha |
| Un error | 294 a 379 | El argumento que nombra |
| El esquema de cada herramienta, una vez por conexión | 5 385 y 6 472 con las globales | Nada |

**Lo que escribe quien invoca.** En `preparar`, la referencia o el texto de la búsqueda: unas decenas de bytes. En
`cotejar`, el texto que pasa, que con la herramienta escribe el modelo: la skill pasa la ficha —316 bytes en la del
fragmento, once líneas— y no el documento, de modo que tampoco la entrada crece con el tamaño de la sentencia
(contracts/skill-jurisprudencia.md §3). El verbo admite la ficha con lo que la siga (FR-020), para quien da el
fichero entero desde la terminal.
