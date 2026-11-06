# Data model: H41 · Las skills del producto mandan leer sus referencias

Lo que el hito declara, escribe o lee. Todo vive en el repositorio o en el directorio de un sondeo; nada en el binario
`kitlegal`, cuya salida no cambia (solo las skills que empotra).

## 1. Familia del laboratorio (`internal/evals/laboratorio_familias.go`)

El tipo `familiaDelLaboratorio` de H40 (data-model §1 de H40), con un campo más (research D10):

| Campo | Qué es | H39 | Familias del producto |
|---|---|---|---|
| `skill` | el `name` de la skill de cada variante | `articulos-del-boe` | `normativa`, `plazos-y-territorio`, `jurisprudencia` |
| `carpeta` | directorio de sus variantes versionadas | `laboratorio/variantes` | `laboratorio/<skill>` |
| `variantes` | nombre y modos, en orden; la primera, la de control | de `variantes.yaml` | las de la tabla de abajo |
| `producto` (**nuevo**) | la variante que se instala desde `skills/<skill>` en cada sondeo | vacío | `<skill>-producto` |
| `referencias` | las de las variantes versionadas, en orden de nombre | la de la declaración | las de `<skill>-control` |
| `evals` | directorio de evals de los repartos | `laboratorio/evals` | `evals/<skill>` |
| `evalsDelSondeo` | las cuatro evals del plan, en orden | todas | research D15 |
| `conLaRegla` | variante con regla testigo | vacío | `normativa-testigo-en-referencia` en `normativa`; vacío en las otras |

Una familia es **del producto** si tiene `producto`. Las tres, en este orden, son `familiasDelProducto`; el arnés y la
tabla las prueban detrás de la de H39.

| Familia | Variantes (modos) | `referencias` | `evalsDelSondeo` |
|---|---|---|---|
| `normativa` | `normativa-control`, `normativa-producto` (orden, herramienta) · `normativa-remisiones-en-orden`, `normativa-cuerpo-corto`, `normativa-testigo-en-referencia` (herramienta) | las ocho de H40 | `04-lgt-prescripcion.yaml`, `05-trlrhl-impuestos-municipales.yaml`, `15-irpf-rendimientos-por-materia.yaml`, `22-lcsp-contrato-menor-en-2019.yaml` |
| `plazos-y-territorio` | `plazos-y-territorio-control`, `plazos-y-territorio-producto`, `plazos-y-territorio-sin-condicion`, `plazos-y-territorio-condicion-por-orden`, `plazos-y-territorio-lista-con-verbo` (orden, herramienta) | `administrativo.md`, `calculo_de_plazos.md`, `jerarquia_normativa.md`, `municipios.md` | `02-territorio-municipio-no-cubierto.yaml`, `05-plazo-reposicion-leganes.yaml`, `10-plazo-reposicion-anterior-a-la-lpac.yaml`, `11-plazo-reposicion-sin-leer-el-precepto.yaml` |
| `jurisprudencia` | `jurisprudencia-control`, `jurisprudencia-producto` (orden, herramienta) | `consulta_preparada.md`, `documento_aportado.md` | `01-existe-con-numero-y-fecha.yaml`, `03-por-materia.yaml`, `04-documento-pegado.yaml`, `06-documento-que-no-es-el-pedido.yaml` |

Reglas: los nombres de las variantes son únicos entre las cuatro familias; la carpeta de una familia del producto tiene un
directorio por variante versionada y nada más, y ninguno se llama `<skill>-producto` (FR-022).

## 2. `<skill>-control` (FR-020)

`laboratorio/<skill>/<skill>-control/skills/<skill>/`: los ficheros de `skills/<skill>` en `v0.9.0` (research V1), con
las mismas rutas relativas. `normativa` ya existe (9 ficheros); nuevos, `plazos-y-territorio` (5) y `jurisprudencia`
(3). Su huella: una tabla `ruta → sha256` por fichero en el test (research D9). Fuera de `skills/`, de `evals/` y de lo
que mide el job (FR-025). El tope de 11 000 caracteres no se les aplica.

## 3. `<skill>-producto` (FR-022, FR-032)

No tiene directorio en el repositorio. En cada sondeo que la pide, el arnés copia `skills/<skill>` del árbol en
`<temporal>/skills/<skill>-producto/<skill>/`, que es el directorio de skills de sus sesiones, y anota sus referencias en
`sondeo.json` (§4). `plazos-y-territorio-producto` es, mientras FR-010 se cumpla, igual que su control.

## 4. `sondeo.json` (FR-032)

El de H39 (`{"modelo","claude_code"}`), con una clave más, solo si el sondeo pide `<skill>-producto`. Un solo tipo lo
escribe y lo lee: `sondeoLeido` (`internal/evals/laboratorio_tabla.go:66`), con el que el arnés lo codifica hoy, gana el
campo `ReferenciasDelProducto []string` con `json:"referencias_del_producto,omitempty"`, tras `Modelo` y `ClaudeCode`:

| Clave | Tipo | Cuándo |
|---|---|---|
| `modelo` | texto | siempre |
| `claude_code` | texto | siempre |
| `referencias_del_producto` | lista de nombres de fichero, en orden de nombre | si el sondeo pide `<skill>-producto` |

Ejemplo, el de la primera invocación de `normativa` de la aceptación (una línea y su salto, 256 bytes; el de
`plazos-y-territorio`, 168; el de `jurisprudencia`, 132):

```json
{"modelo":"claude-sonnet-5-5","claude_code":"2.1.317","referencias_del_producto":["buscar_la_norma.md","como_se_cita.md","con_herramientas.md","lecturas_que_fallan.md","normas.md","redaccion_en_una_fecha.md","redaccion_modificada.md","varios_bloques.md"]}
```

Sin `<skill>-producto`, el de hoy, byte a byte: `{"modelo":"claude-sonnet-5-5","claude_code":"2.1.317"}`. Los guiones
de H39 y H40 leen de él lo mismo que hoy (research V9: claves sin campo, ignoradas). Lo escribe el arnés una vez por
sondeo y lo lee la tabla una vez; no crece con el uso del producto.

## 5. Declaración del producto (`TestSkillsDelProductoDesdeControl`; FR-006, FR-007)

Por skill, en `internal/evals/laboratorio_familias_test.go`, con los bytes de contracts/remisiones.md §2 a §4:

- **cambio del cuerpo** `{antes, despues}`: `antes` está una sola vez en el `SKILL.md` de control, sin lo generado; cada
  remisión reescrita es uno o dos (una que cambia de sitio sale de uno y entra en otro). `normativa`, los tramos de §2;
  `jurisprudencia`, los de §4.
- **movimiento** `{texto, referencia, enLaReferencia}`: `texto` sale del `SKILL.md` (es el `antes` de un cambio con
  `despues` vacío o reducido); `enLaReferencia` son los bytes que se añaden al final de `references/<referencia>`; los
  dos son el mismo texto con los blancos juntados. Dos, en `normativa` (§3); ninguno en `jurisprudencia`.
- **`plazos-y-territorio`**: sin declaración; la igualdad es total, con lo generado (FR-010).

## 6. Declaración de las variantes de `plazos-y-territorio` (`TestVariantesDePlazosYTerritorio`; FR-023)

El tipo `declaracionDeLaVariante` de H40 (`internal/evals/laboratorio_normativa_test.go:221`), solo con
`sustituciones` `{fichero, antes, despues, hito}`: `fichero` es siempre `SKILL.md`; `antes`, bytes que están una vez en
el de `plazos-y-territorio-control`; `despues`, los de contracts/variantes.md §2; `hito`, el `despues` normalizado
(`textoDelHito`). Una lista `declaracionesDePlazosYTerritorio`, en el orden de las variantes de la familia tras la de
control y la del producto. Lo que no declara es el de control.

## 7. La lista de remisiones (`TestRemisionesDelRepositorio`; FR-001)

`remisionesDelPlan map[string][]string`: por skill, las frases de contracts/remisiones.md §5, en orden, como las da
`skills.Frases`. `fueraDeLasRemisiones map[string]string`: `{"plazos-y-territorio": "H41.1"}` (FR-010).

## 8. Sesión contada y referencias de la tabla (FR-040, FR-041)

`sesionContada` de H40 sin cambios: `aperturas` va por cada referencia **de la tabla**, con el turno de su primera
apertura o 0. La regla de apertura de una orden de `Bash` gana, en las familias del producto, la lectura con `cd`
(contracts/tabla.md §2).

Las referencias de la tabla son la unión, en orden de nombre, de las `referencias` de la familia por cada directorio de
variante versionada que lee y de la lista `referencias_del_producto` de cada sondeo que tiene un directorio
`<skill>-producto`. Una variante versionada tiene las de la familia; `<skill>-producto`, las de la unión de sus listas.
La celda de una referencia que la variante de la fila no tiene va vacía, en la fila y en el detalle.

## 9. Estados y transiciones

No hay estado que cambie con el uso del producto. `<skill>-control` y las variantes versionadas no cambian nunca (§2,
§6); `<skill>-producto` es `skills/<skill>` en el momento de cada sondeo; un sondeo se crea entero en un `DIRECTORIO` que
no existía (H39) y no se toca después.
