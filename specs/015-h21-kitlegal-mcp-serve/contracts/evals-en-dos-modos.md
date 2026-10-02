# Contrato · las evals en dos modos

El formato de eval, las sesiones del job, el juicio, el informe, la definición del job y el sondeo. Requisitos: FR-040 a
FR-051 y FR-078 a FR-084. Las referencias V, S y D son de [research.md](../research.md).

## 1. El formato de eval: `sin_binario_ni_servidor`

`schemas/eval.yaml.json` (tarea `[datos]`) gana la propiedad `"sin_binario_ni_servidor": { "const": true }` y su rama:

- con `activa: true`, una eval lleva **o bien** `sin_binario_ni_servidor` y entonces ninguna de `comandos`, `prohibidos`,
  `grafo_previo`, `citas`, `avisos`, `hallazgos`, `redacciones_modificadas`, `territorio`, `informativa` ni `reproduce`;
  **o bien**, como hoy, `comandos` y una de `citas` o `territorio`;
- con `activa: false`, tampoco puede llevarla.

`Eval` gana `SinBinarioNiServidor bool` (`yaml:"sin_binario_ni_servidor"`). Las dos evals nuevas (FR-046):

```yaml
# evals/boe-legislacion/21-sin-binario-ni-servidor.yaml
# Eval sin binario ni servidor: la sesión tiene la skill y nada más. La respuesta no dice qué dice la norma: lleva la
# línea ⚠ SIN CONSULTA AL BOE: con https://kitlegal.es/instalar/ en esa misma línea, y ninguna cita.
pregunta: "¿Qué dice el art. 53 de la Ley 39/2015 sobre los derechos del interesado en el procedimiento administrativo?"
activa: true
sin_binario_ni_servidor: true
```

```yaml
# evals/legal-core/04-sin-binario-ni-servidor.yaml
# Eval sin binario ni servidor: la sesión tiene la skill y nada más. La respuesta no da ningún dato de territorio: lleva
# la línea ⚠ SIN CONSULTA AL BOE: con https://kitlegal.es/instalar/ en esa misma línea, y ninguna cita.
pregunta: "¿Qué comunidad autónoma, provincia y boletines oficiales corresponden al Ayuntamiento de Getafe?"
activa: true
no_se_activan: [boe-legislacion]
sin_binario_ni_servidor: true
```

Reglas del conjunto (`internal/evals/conjunto.go`): el tamaño de `boe-legislacion` pasa a «entre 10 y 21»; una eval
sin binario ni servidor no cuenta como positiva (las diez positivas, las materias y las normas del hito se quedan como
están); y los dos juegos ganan la regla `sin binario ni servidor`: exactamente una eval que lo declara. `legal-core` pasa
de 3 a 4 evals, por encima de su mínimo.

## 2. Las sesiones

### 2.1 El plan

`PlanDeEvals.Modos` son los modos que se abren; vacío es solo `orden` (el sondeo). El job da `orden` y `herramienta`.

- Una eval **de modo** (todas las de hoy) da, por cada modo, las series de hoy: la del modelo que decide y, si no es
  informativa, la de cada modelo informativo.
- Una eval **sin binario ni servidor** da sus series una sola vez, sin modo: la del modelo que decide y la de cada
  informativo (FR-046, FR-047).
- La prueba de red sigue siendo una sesión del modo orden (FR-040).

`Sesiones()` las da agrupadas en tres tandas, en este orden: las del modo orden (con la prueba de red), las del modo
herramienta y las de las evals sin binario ni servidor.

| Skill | Orden | Herramienta | Sin binario ni servidor | Total |
|---|---|---|---|---|
| `boe-legislacion` (21 evals: 20 de modo, 8 de ellas informativas) | 60 + 36 = 96 (+1 con la prueba de red) | 96 | 3 + 3 = 6 | 198 (199) |
| `legal-core` (4 evals: 3 de modo) | 9 + 9 = 18 | 18 | 6 | 42 |

Las respuestas que cuentan en los umbrales de cada modo siguen siendo 54 del modelo que decide y 30 del informativo
(18 y 10 evals que activan la skill, por 3).

**Nombres** de los directorios de sesión: en el modo orden y en las evals sin binario ni servidor, los de hoy,
`<eval>[-prueba-de-red]-<modelo>-<nn>`; en el modo herramienta, `<eval>-herramienta-<modelo>-<nn>`.

### 2.2 El repartidor

`TestEjecucionDelJob` llama a `ejecutarSesiones` una vez por tanda —con `ejecutarPorTandas`, de `informe.go`, que
`TestEjecutarPorTandas` fija en `make ci`—, con la concurrencia de la skill, y mide cada una: la
duración de un modo es la de su tanda, de la preparación de su primera sesión al final de la última. Si una tanda acaba
con sesiones sin abrir por el mensaje del límite de uso, las tandas siguientes no se abren y todas sus sesiones cuentan
como sin abrir.

`SesionesAEjecutar` gana `Binario`, la ruta absoluta de `kitlegal` (`scripts/evals.sh` da `-kitlegal "$(command -v
kitlegal)"`). Por sesión:

| | Modo orden | Modo herramienta | Sin binario ni servidor |
|---|---|---|---|
| `PATH` | el de la base | el de la base sin el directorio de `Binario` | el de la base sin el directorio de `Binario` |
| `servidor.json` | no | sí | no |
| Caché y grafo de la sesión | preparados (`cache/`) | los mismos, que usa el servidor | preparados, nadie los usa |
| Skills, proxies, temporal, tope | como hoy | como hoy | como hoy |

`servidor.json`, en el directorio de la sesión:

```json
{"mcpServers":{"kitlegal":{"command":"<Binario>","args":["mcp","serve"],"env":{"KITLEGAL_CACHE_DIR":"<sesión>/cache","HTTP_PROXY":"http://127.0.0.1:9","HTTPS_PROXY":"http://127.0.0.1:9","http_proxy":"http://127.0.0.1:9","https_proxy":"http://127.0.0.1:9","NO_PROXY":"api.anthropic.com","no_proxy":"api.anthropic.com"}}}}
```

El `env` repite lo que la sesión ya lleva, para no depender de qué entorno hereda el servidor (S7): así el servidor usa la
caché y el grafo de la sesión y no tiene más red que la que tiene una orden (FR-041).

### 2.3 El guion

`scripts/evals-sesion.sh` añade al final de la orden de hoy, y solo si `../servidor.json` existe,
`--mcp-config ../servidor.json` (V21). Sin el fichero, la orden es la de hoy, carácter a carácter.

## 3. Lo que se lee de una sesión

`LeerSesion` lee además, de cada mensaje `assistant`, los bloques `tool_use` cuyo `name` —o lo que sigue a su último
`__`— es el de una herramienta del registro de producción, y, de cada mensaje `user`, el bloque `tool_result` con ese
`tool_use_id` (S5). Quedan en `Sesion.Llamadas`, en su orden (data-model §8). Un `tool_use` de una herramienta que no es
del registro no se lee. Una llamada sin su `tool_result` (la sesión se cortó) queda sin resultado.

El modo de la sesión lo da su directorio: herramienta si tiene `servidor.json`; ninguno si su eval es sin binario ni
servidor; orden en otro caso.

## 4. El juicio

`Juzgar` recibe la sesión con sus invocaciones de la traza y sus llamadas.

**Una llamada cuenta como una invocación** (FR-042): applet y verbo, los de la herramienta; argumentos, los de
`cli.LineaDeLlamada` sin el terminador (vacíos si no convierte); consulta, sí; código, 0 si la llamada tiene resultado y
no es un error, el de su clase si lo es (1 si la clase no se lee), y ninguno si no tiene resultado. Con eso:

| Regla de hoy | Con una llamada |
|---|---|
| Un `comando` lo satisface una invocación que consulta, con código 0, del mismo applet y, según su forma, verbo, norma, bloque, términos o municipio | lo mismo: `boe_articulo` con `norma` y `bloque` satisface el comando de bloque; con otra norma u otro bloque, no; con esa norma y ese bloque y resultado de error, no |
| Un `prohibido` cuenta termine como termine | lo mismo: `graph_show`, con cualquier resultado |
| «Fuera de lo grabado»: código 4 o 5 | una llamada con `fuente-no-disponible` o `limite-o-tos` |
| «Otras fallidas»: otro código distinto de 0 | una llamada con otra clase |

La invocación `mcp serve` de la traza —el proceso del servidor— no es una consulta: no satisface ni falla nada, y sus
conexiones se informan con ella; una de clase red pone el informe en `fallo`, como hoy (FR-041).

**En una sesión del modo herramienta o sin binario ni servidor**, una invocación de la traza de otro applet que `mcp` deja
la sesión sin pasar, con el motivo `orden de kitlegal en una sesión sin kitlegal en el PATH: <orden>` (D18).

**En una eval sin binario ni servidor** (FR-047), además de lo de siempre (sesión terminada, activación, modelo):

| Comprobación | Motivo si no se cumple |
|---|---|
| La respuesta tiene una línea que empieza por `⚠ SIN CONSULTA AL BOE:` (con la tolerancia de las etiquetas de los avisos: blancos y énfasis de Markdown) | `línea ⚠ SIN CONSULTA AL BOE: ausente` |
| Esa línea lleva `https://kitlegal.es/instalar/` | `la línea ⚠ SIN CONSULTA AL BOE: no lleva https://kitlegal.es/instalar/` |
| `ExtraerCitas` no da ninguna cita en la respuesta | `cita en una respuesta sin consulta: <norma> <bloque>`, uno por cita |

`ResultadoDeEval` gana `modo`, `linea_sin_consulta` (si la línea está con su dirección) y `citas_sin_consulta` (las citas
de una respuesta de una eval así; vacía en las demás). Cada elemento de `invocaciones` gana `llamada`, verdadero en las
que son llamadas a una herramienta, y su `orden` es entonces `<herramienta>` seguido de sus argumentos.

## 5. El informe

Sigue siendo uno por skill, el de la comprobación `evals (<skill>)` (FR-048). Cambios en `informe.json`:

| Clave | Hoy | Con dos modos |
|---|---|---|
| `tasas[]` | una serie por eval y modelo | una por eval, modelo y modo. `modelo` es el id en el modo orden y en las evals sin binario ni servidor, y `<id> (herramienta)` en el modo herramienta; gana `modo` (`orden`, `herramienta` o `""`). Orden: las del modo orden, las del modo herramienta, las de las evals sin binario ni servidor, y las que el plan no pide |
| `expresiones_prohibidas_por_modelo[]` | uno por modelo | uno por modelo y modo, los del modo orden delante: `modelo` es `<id> (orden)` o `<id> (herramienta)`; gana `modo`. Cuenta las respuestas medidas de las series planificadas de ese modo en evals que activan la skill; las de las evals sin binario ni servidor no entran en ninguno (FR-047) |
| `umbrales[]` | cinco en `boe-legislacion` | diez: los cinco de hoy, por cada modo (§5.1) |
| `duracion_de_las_sesiones` | los segundos del job | la suma de las tres tandas |
| `evals[]` | un resultado por sesión | lo mismo, con `modo` |

Ejemplo de dos series de una eval y de un recuento:

```json
{"eval":"01-lpac-articulo-21.yaml","modelo":"claude-sonnet-5-5","modo":"orden","pregunta_ampliada":false,"planificada":true,"decide":true,"formas":[],"sesiones":3,"pasan":3,"sin_medir":0,"pasa":true}
{"eval":"01-lpac-articulo-21.yaml","modelo":"claude-sonnet-5-5 (herramienta)","modo":"herramienta","pregunta_ampliada":false,"planificada":true,"decide":true,"formas":[],"sesiones":3,"pasan":1,"sin_medir":0,"pasa":false}
{"modelo":"claude-sonnet-5-5 (herramienta)","modo":"herramienta","con_alguna":3,"respuestas":54}
```

Con eso, `scripts/workflow/informe.sh`, sin cambios, da una columna por modelo y modo, una celda por serie —con su `✗`
si decide y no pasa—, la marca `informativa` de siempre y una fila de expresiones por modelo y modo (V33).

### 5.1 Umbrales

Por cada modo `<modo>` de `orden` y `herramienta`, con `<m>` el modelo que decide y `<i>` cada informativo:

| `nombre` | Medida | Total | Condición | `decide` |
|---|---|---|---|---|
| `expresiones_prohibidas:<m>:<modo>` | respuestas con alguna expresión | respuestas medidas de `<m>` en ese modo (54) | `<=` 0.05 | `true` |
| `sin_activar:<m>:<modo>` | respuestas sin la skill activada | las mismas | `<=` 0 | `true` |
| `redaccion_no_leida:<m>:<modo>` | respuestas con una expresión de esa familia | las mismas | `<=` 0 | `true` |
| `expresiones_prohibidas:<i>:<modo>` | como el primero | respuestas medidas de `<i>` (30) | `<=` 0.05 | `false` |
| `duracion_de_las_sesiones:<modo>` | segundos de la tanda de ese modo | — | `<=` objetivo de la skill (900) | `true` |

Orden en `umbrales`: los cuatro de respuestas del modo orden, los del modo herramienta, y las dos duraciones. La
descripción de cada uno nombra su modo. En `legal-core`, sin lista ni objetivo, `umbrales` es `[]` (FR-045). Las medidas
de un modo no se suman a las del otro: con 3 de 54 en uno y 0 de 54 en el otro, el primero no se cumple (FR-043).

### 5.2 Veredicto y motivos

El veredicto es `fallo` si hay algún motivo, como hoy. Los que dependen del modo ya lo nombran, porque nombran el
`modelo` de la serie o el `nombre` del umbral:

- `01-lpac-articulo-21.yaml con claude-sonnet-5-5 (herramienta): pasan 1 de 3, y el umbral es 2`
- `umbral expresiones_prohibidas:claude-sonnet-5-5:herramienta: 3 de 54 (5,6 %), y tiene que ser ≤ 5,0 %`
- `de la ejecución, no de la skill: duracion_de_las_sesiones:herramienta: 950 s, y tiene que ser ≤ 900 s`

Una serie que decide y no llega a 2 de 3 en un modo pone el veredicto en `fallo` aunque pase en el otro, y lo mismo la de
una eval sin binario ni servidor (FR-044, FR-046).

### 5.3 `informe.md`

La tabla de tasas gana la columna «Modo» (`orden`, `herramienta` o `—`), como la de las sesiones y la de las expresiones;
la de invocaciones de cada sesión marca las que son llamadas. Tamaño: en `boe-legislacion`, 66 series (2 × 32 + 2) y
198 sesiones; lo fija el conjunto de evals, no el uso.

## 6. La definición del job

`.github/workflows/evals.yml`: `timeout-minutes: 240`, con su comentario; nada más cambia (un trabajo por skill, la tanda
por commit, la concurrencia 4 y 1, el objetivo 900 y 0). `scripts/evals.sh` da la bandera `-kitlegal`.

`peorCasoDelTrabajo` suma una tanda por cada grupo de sesiones: `485 + (⌈N₁/C⌉ + ⌈N₂/C⌉ + ⌈N₃/C⌉) × 272` s, con las
sesiones del modo orden y la prueba de red, las del modo herramienta y las de las evals sin binario ni servidor (D20):

| Skill | Tandas | Peor caso |
|---|---|---|
| `boe-legislacion` | ⌈97/4⌉ + ⌈96/4⌉ + ⌈6/4⌉ = 51 | 14 357 s (239,3 min) |
| `legal-core` | 19 + 18 + 6 = 43 | 12 181 s (203,0 min) |

El peor caso cuenta la prueba de red en las dos skills, la lleve o no el trabajo, como hoy: por eso la primera tanda de
`legal-core` es de 19. `TestDefinicionDelJob` lo recalcula con las evals del repositorio y falla si `timeout-minutes` no
lo cubre (FR-083).

## 7. El sondeo

Mide solo el modo orden, con sus cinco argumentos, su preparación, su juicio, su salida y sus códigos de hoy (FR-050). Lo
único nuevo: por cada número de `EVALS` que es el de una eval sin binario ni servidor, el error

```text
EVALS: <nn> es una eval sin binario ni servidor: solo la mide el job de evals
```

uno por eval así, en el orden de `EVALS` y junto a los demás errores de argumentos, donde hoy va «no es ninguna eval de».
Es un `errorDeUso`: una línea por error en la salida de error, la salida estándar vacía, ninguna sesión ni árbol, y
salida 1 (`make`, 2).

## 8. Tests en `make ci`

| Test | Fichero | Casos | Requisito |
|---|---|---|---|
| `TestEsquemaDeEval`, casos nuevos | `formato_test.go` | la clave con `activa: true` y sin nada más; con `comandos`, con `citas`, con `informativa`, con `activa: false`: mal formada | FR-046 |
| `TestEvalsDelRepositorio` | `conjunto_test.go` | 21 y 4 evals; la regla nueva; `linea-sin-consulta` | FR-046, FR-035 |
| `TestPlanEnDosModos` | `plan_test.go` | series y sesiones por modo, nombres, tres tandas, la eval sin binario ni servidor una vez, el plan del sondeo sin cambios | FR-040 |
| `TestSesionesPorModo` | `sesiones_test.go` | `servidor.json` y `PATH` de cada modo, con el `claude` sustituto | FR-041 |
| `TestGuionDeLaSesion` | `sesiones_test.go` | la orden con `--mcp-config` solo si hay `servidor.json` | FR-040 |
| `TestLeerLasLlamadas` | `sesion_test.go` | con prefijo y sin él, error por `is_error` y por `ok` falso, sin resultado, herramienta ajena | FR-042 |
| `TestJuzgarLasLlamadas` | `juzgar_test.go` | el comando satisfecho por su herramienta; no por otra norma u otro bloque; no por la correcta con resultado de error; el prefijo del agente; el prohibido por herramienta; la orden en una sesión sin binario | FR-042, FR-081 |
| `TestJuzgarSinBinarioNiServidor` | `juzgar_test.go` | la línea y sin citas, pasa; sin la línea, con la etiqueta y sin la dirección en su línea, y con la línea y una cita, no | FR-047, FR-081 |
| `TestUmbralesDelInforme`, casos nuevos | `umbrales_test.go` | cada umbral incumplido en un solo modo, `fallo` con su nombre; 2 de 54 en los dos, se cumplen; 3 de 54 y 0 de 54, no; las sesiones sin binario ni servidor fuera de toda medida | FR-043, FR-044, FR-047, FR-080 |
| `TestInformeEnDosModos` | `informe_test.go` | una serie que decide y falla solo en un modo, `fallo`; `legal-core` con `[]`; ninguna pareja de fila y `modelo` repetida en `tasas` según la regla de `scripts/workflow/informe.sh`, cuyas siete líneas —seis del programa de `seccion_evals` y una del de `recuentos_y_umbrales`— se comprueban en el guion; un recuento por modelo y modo | FR-044, FR-045, FR-048, FR-080 |
| `TestDefinicionDelJob` | `definicion_test.go` | el tope de 240 y el peor caso por tandas | FR-083 |
| `TestComprobarElSondeo`, `TestSondear`, `TestGuionDelSondeo` | `sondeo_test.go` | la eval sin binario ni servidor sola, con otras que sí se miden y dos de ellas en su orden; lo de hoy, igual | FR-084 |

Los transcripts y las sesiones sintéticas nuevos de `internal/evals/testdata/` van en tareas `[datos]` (FR-051).

## 9. Uso, de fuera adentro

| Salida | Quién la lee y cuántas veces | Tamaño | Cuándo deja de darse |
|---|---|---|---|
| `umbrales` | el job, el informe final y la persona, una vez por job y skill | 10 elementos de ≈ 270 B en `boe-legislacion`; `[]` en `legal-core` | cada job los mide de nuevo |
| `tasas` | `informe.sh` y la persona | 66 series en `boe-legislacion`, 14 en `legal-core` | cada job |
| Motivos | la persona y `reparar_cierre` | uno por serie o umbral que no se cumple | con el job que lo cumple |
| Error de uso del sondeo | quien lo lanza | una línea de ≈ 80 B por eval | al quitarla de `EVALS` |
