# Data model: H41 · Las skills del producto mandan leer sus referencias

Lo que el hito declara, escribe o lee. Todo vive en el repositorio o en el directorio de un sondeo; nada en el binario
`kitlegal`, cuya salida no cambia.

## 1. Familia del laboratorio (`internal/evals/laboratorio_familias.go`)

El tipo `familiaDelLaboratorio` de H40 (data-model §1 de H40), con un campo más:

| Campo | Qué es | H39 | Familias del producto |
|---|---|---|---|
| `skill` | el `name` de la skill de cada variante | `articulos-del-boe` | `normativa`, `plazos-y-territorio`, `jurisprudencia` |
| `carpeta` | directorio de sus variantes versionadas | `laboratorio/variantes` | `laboratorio/<skill>` |
| `variantes` | nombre y modos, en orden; la primera, la de control | de `variantes.yaml` | las de la tabla de abajo |
| `producto` (**nuevo**) | la variante que se instala desde `skills/<skill>` en cada sondeo | vacío | `<skill>-producto` |
| `referencias` | las de las variantes versionadas, en orden de nombre | la de la declaración | las de `<skill>-control` |
| `evals` | directorio de evals de los repartos | `laboratorio/evals` | `evals/<skill>` |
| `evalsDelSondeo` | las cuatro evals del plan, en orden | todas | research D12 |
| `conLaRegla` | variante con regla testigo | vacío | `normativa-testigo-en-referencia` en `normativa`; vacío en las otras |

Una familia es **del producto** si tiene `producto`. Las tres, en este orden, son `familiasDelProducto`; el arnés y la
tabla las prueban detrás de la de H39.

| Familia | Variantes (modos) | `referencias` | `evalsDelSondeo` |
|---|---|---|---|
| `normativa` | `normativa-control` (orden, herramienta) · `normativa-producto` (orden, herramienta) · `normativa-remisiones-en-orden`, `normativa-cuerpo-corto`, `normativa-testigo-en-referencia` (herramienta) | las ocho de H40 | 04, 05, 15 y 22 (las de H40) |
| `plazos-y-territorio` | `plazos-y-territorio-control`, `plazos-y-territorio-producto` (orden, herramienta) | `administrativo.md`, `calculo_de_plazos.md`, `jerarquia_normativa.md`, `municipios.md` | `02-territorio-municipio-no-cubierto.yaml`, `05-plazo-reposicion-leganes.yaml`, `10-plazo-reposicion-anterior-a-la-lpac.yaml`, `11-plazo-reposicion-sin-leer-el-precepto.yaml` |
| `jurisprudencia` | `jurisprudencia-control`, `jurisprudencia-producto` (orden, herramienta) | `consulta_preparada.md`, `documento_aportado.md` | `01-existe-con-numero-y-fecha.yaml`, `03-por-materia.yaml`, `04-documento-pegado.yaml`, `06-documento-que-no-es-el-pedido.yaml` |

Reglas: los nombres de las variantes son únicos entre las cuatro familias; la carpeta de una familia del producto tiene un
directorio por variante versionada y nada más, y ninguno se llama `<skill>-producto` (FR-022).

## 2. `<skill>-control` (FR-020)

`laboratorio/<skill>/<skill>-control/skills/<skill>/`: los ficheros de `skills/<skill>` en `v0.9.0` (research V1), con
las mismas rutas relativas. `normativa` ya existe (9 ficheros); nuevos, `plazos-y-territorio` (5: `SKILL.md` y sus cuatro
referencias) y `jurisprudencia` (3). Su huella: una tabla `ruta → sha256` por fichero en el test (research D8). Fuera de
`skills/`, de `evals/` y de lo que mide el job (FR-023). El tope de 11 000 caracteres no se les aplica.

## 3. `<skill>-producto` (FR-022, FR-033)

No tiene directorio en el repositorio. En cada sondeo que la pide, el arnés copia `skills/<skill>` del árbol en
`<temporal>/skills/<skill>-producto/<skill>/`, que es el directorio de skills de sus sesiones, y anota sus referencias en
`sondeo.json` (§4).

## 4. `sondeo.json` (FR-033)

El de H39 (`{"modelo","claude_code"}`), con una clave más, solo si el sondeo pide `<skill>-producto`:

| Clave | Tipo | Cuándo |
|---|---|---|
| `modelo` | texto | siempre |
| `claude_code` | texto | siempre |
| `referencias_del_producto` | lista de nombres de fichero, en orden de nombre | si el sondeo pide `<skill>-producto` |

Ejemplo, el del primer lanzamiento de `normativa` de la aceptación (una línea y su salto, 256 bytes; el de
`plazos-y-territorio`, 168; el de `jurisprudencia`, 132):

```json
{"modelo":"claude-sonnet-5-5","claude_code":"2.1.317","referencias_del_producto":["buscar_la_norma.md","como_se_cita.md","con_herramientas.md","lecturas_que_fallan.md","normas.md","redaccion_en_una_fecha.md","redaccion_modificada.md","varios_bloques.md"]}
```

Sin `<skill>-producto`, el de hoy, byte a byte: `{"modelo":"claude-sonnet-5-5","claude_code":"2.1.317"}`. Los guiones
de H39 y H40 leen de él lo mismo que hoy (research D10).

## 5. Declaración de la reorganización (test de FR-007 y FR-009)

Por skill, en `internal/evals/laboratorio_familias_test.go`, con los bytes de contracts/remisiones.md:

- **cambio del cuerpo** `{antes, despues}`: `antes` está una sola vez en el `SKILL.md` de control, sin lo generado; cada
  remisión reescrita es uno o dos (una que cambia de sitio sale de uno y entra en otro).
- **movimiento** `{texto, referencia, enLaReferencia}`: `texto` sale del `SKILL.md` (es el `antes` de un cambio con
  `despues` vacío); `enLaReferencia` son los bytes que se añaden al final de `references/<referencia>`; los dos son el
  mismo texto con los blancos juntados y los enlaces relativos a la referencia.

Lo generado por `make skills-sync` (referencias con «generado desde» y regiones entre marcas) queda fuera de la
comparación (research D7).

## 6. Sesión contada y tabla (FR-040, FR-041)

`sesionContada` de H40 sin cambios: `aperturas` va por cada referencia **de la tabla** (§7), con el turno de su primera
apertura o 0. La regla de apertura de una orden de `Bash` gana, en las familias del producto, la lectura con `cd`
(contracts/tabla.md §2).

## 7. Referencias de la tabla (FR-040)

La unión, en orden de nombre, de las `referencias` de la familia por cada directorio de variante versionada que lee y de
la lista `referencias_del_producto` de cada sondeo que tiene un directorio `<skill>-producto`. Una variante versionada
tiene las de la familia; `<skill>-producto`, las de la unión de sus listas. La celda de una referencia que la variante de
la fila no tiene va vacía, en la fila y en el detalle.

## 8. Estados y transiciones

No hay estado que cambie con el uso del producto. `<skill>-control` no cambia nunca (§2); `<skill>-producto` es
`skills/<skill>` en el momento de cada sondeo; un sondeo se crea entero en un `DIRECTORIO` que no existía (H39) y no se
toca después.
