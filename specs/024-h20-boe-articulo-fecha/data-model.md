# Data model: H20 · `boe articulo --fecha` + `boe-legislacion` v0.2

Cada entidad con sus campos, sus reglas y el requisito que la exige. Lo que no está aquí no cambia. Los nombres Go son
los que el plan fija; los JSON y YAML, los que ven el binario, las skills y las evals.

## 1. La consulta: `boe.ConsultaArticulo`

| Campo | Tipo | Regla | Requisito |
|---|---|---|---|
| `Norma` | `string` | como hoy, `ValidarNorma` | — |
| `Bloque` | `string` | como hoy, `ValidarBloque` | — |
| `Fecha` | `*string` (nuevo) | `nil`: la lectura de hoy, por el camino de hoy. No `nil`: `ValidarFecha` —`time.Parse(time.DateOnly, …)`— detrás de la norma y el bloque y antes de abrir nada; si no es un día AAAA-MM-DD, código 2 | FR-001, FR-005, FR-012 |

La rellena `argumentosDeArticulo` (`internal/app/boe.go`) desde su campo nuevo `Fecha *string`
(`placeholder:"AAAA-MM-DD"`), que Kong expone como `--fecha` y `--describe` como `fecha`, en `x-banderas`. La
herramienta `boe_articulo` lo recibe como el parámetro opcional `fecha` sin código nuevo (research V14).

## 2. Las versiones de un bloque: `versionLeida` y `bloqueLeido`

`leerBloque` (`internal/source/boe/bloque.go`) guarda, además de lo que da hoy, cada versión en su orden:

| Campo de `versionLeida` | De dónde sale | Como hoy la última |
|---|---|---|
| `fechaVersion` | `fecha_publicacion`, aunque venga vacía; `original` si falta | nota [5] |
| `fechaVigencia` | `fecha_vigencia`; vacía si falta | nota [5] |
| `normaModificadora` | `id_norma`; vacía si falta | nota [5] |
| `texto` | el de la versión, normalizado como hoy (nota [6]) | — |

`bloqueLeido` gana `versiones []versionLeida`. Sus campos de hoy siguen siendo los de la última versión, o los del
bloque sin versiones, byte a byte (FR-012: los golden de hoy no cambian). Un bloque sin versiones deja `versiones`
vacía.

**Situable**: un bloque cuyas versiones, al menos una, tienen todas una `fechaVigencia` que se lee como día AAAAMMDD.
Solo un bloque situable da una versión (FR-004).

## 3. La elección (`internal/source/boe/fecha.go`)

Con un bloque situable y la fecha pedida *D* (FR-002, FR-003, FR-010):

- **versión elegida**: la última, en el orden del bloque, cuya fecha de vigencia no es posterior a *D*;
- **sin versión elegida** (todas posteriores a *D*): 3, con el mensaje de FR-003, que nombra *D* y la fecha de vigencia
  de la primera versión;
- **redacción posterior**: la última versión del bloque, si no es la elegida; si lo es, ninguna (FR-013).

Un bloque no situable da 3 con el mensaje de FR-004, para cualquier *D* (FR-004). Los mensajes, en
contracts/articulo-fecha.md §3.

## 4. La salida: `boe.ArticuloFechado` y `boe.RedaccionPosterior`

```go
type ArticuloFechado struct {
	Articulo
	FechaPedida        string              `json:"fecha_pedida,omitempty"`
	RedaccionPosterior *RedaccionPosterior `json:"redaccion_posterior,omitempty"`
}

type RedaccionPosterior struct {
	FechaVigencia     string `json:"fecha_vigencia"`
	NormaModificadora string `json:"norma_modificadora"`
}
```

- `Articulo` lleva los datos de la versión elegida tal como la nota [5] los da de la última: `fecha_version`,
  `fecha_vigencia`, `norma_modificadora`, `texto` y `hash_texto` de esa versión; `titulo`, `tipo`, `url` y los de hoy;
  `avisos` y `url_eli`, los de los metadatos de la norma hoy (FR-010, FR-011).
- `FechaPedida`: la fecha tal como se pidió, AAAA-MM-DD (FR-010).
- `RedaccionPosterior`: `nil` si la elegida es la última (FR-010, FR-013).
- Es la `Salida` que el registro declara para `articulo` (`internal/app/boe.go`); la lectura sin fecha sigue
  devolviendo `Articulo`, y su JSON no lleva ninguna de las dos claves (FR-012).

## 5. La entrada de las versiones (`internal/source/boe/entradas.go`)

| | |
|---|---|
| Clave | `claveDeLasVersiones(norma, bloque)`: `boe.legislacion-consolidada\|1\|versiones\|<dirección de la API del bloque>` |
| Datos | `versionesDelBloque`: `titulo`, `tipo`, `versiones` (cada una con `fecha_version`, `fecha_vigencia`, `norma_modificadora` y `texto`), `avisos` y `url_eli` |
| Fecha de consulta | la más antigua de la del bloque y la de sus metadatos, como el artículo de hoy |
| Vigencia | 7 días (`vigenciaLarga`), como el artículo |
| Quién la lee y la escribe | solo la lectura fechada; la escribe antes de elegir, tras obtener e interpretar el bloque y sus metadatos; con un fallo de cualquiera de los dos, nada (FR-007) |

`versionesDelBloque` entra en la unión `datosDeEntrada`. La entrada de hoy (`claveDelArticulo`, `Articulo`) no cambia
de clave, de forma ni de contenido, y la lectura fechada no la lee ni la escribe (FR-020, FR-021). `versionDeLasEntradas`
sigue en `1`. Con `--offline`: con la entrada vigente, la lectura fechada de cualquier fecha se sirve de ella (FR-022);
sin ella, 4, como toda entrada ausente (FR-020).

## 6. El grafo

| Pieza | Cambio | Requisito |
|---|---|---|
| `schema.Observado` (`internal/core/schema/grafo.go`) | gana `SoloLoNuevo bool`: con él, lo observado se añade a `world.db` solo donde no está | FR-031 |
| `core.Lote` (`internal/core/graphstore.go`) | gana `SoloLoNuevo bool` | FR-031 |
| `cli.LoteDe` (`internal/cli/entrega.go`) | lo copia del `Observado` al lote | FR-030, FR-031 |
| `grafo.Consolidado` (`internal/core/grafo/lote.go`) | gana `SoloLoNuevo`, que `Consolidar` copia del lote | FR-031 |
| `aplicarConsolidado` y `tablaDelGrafo.aplicar` (`internal/graph/aplicar.go`) | con `SoloLoNuevo`, un nodo, una arista o un texto que `world.db` ya guarda se deja como está; uno que no guarda se escribe con la procedencia del lote | FR-030, FR-031 |
| `observadoDeLaLecturaFechada` (`internal/source/boe/grafo.go`, nueva) | el nodo `BloqueVersion` de la versión elegida, con el id y los datos de la lectura sin fecha (`<eli>#<bloque>@<fecha_vigencia>:<hash_texto>`; `fecha_vigencia`, `fecha_version`, `norma_modificadora`, `hash_texto`) y su `Texto`; `Vigencia` de 7 días y `SoloLoNuevo`; nada sin ELI ni bajo ensayo | FR-030 |

**Sin cambios**: la fila de lecturas, `Lecturas()`, `graph check` y su regla de la redacción vista, la entrega de la
lectura sin fecha (`SoloLoNuevo` falso) y `graph.Nulo`. Una lectura fechada no lleva `eli:has_version`, así que no
es una lectura (FR-031).

## 7. Las evals (`internal/evals`, `schemas/eval.yaml.json`)

### 7.1 El comando de bloque con fecha

`ComandoEsperado.Fecha string` (`fecha`, AAAA-MM-DD), solo en la forma bloque. En el esquema, `comandos` admite
`comando-bloque-esperado` (`applet`, `norma`, `bloque` y `fecha` opcional); `grafo_previo` sigue con `comando-bloque`
(FR-051).

| Comando | Casa con |
|---|---|
| sin `fecha` | `articulo` o `articulos` de la norma con el bloque entre sus bloques, sin `--fecha`, con código 0 (como hoy, más «sin `--fecha`») |
| con `fecha` *D* | `articulo` de la norma y el bloque con `--fecha` *D* —orden o herramienta—, con código 0 o 3 |

### 7.2 Lo esperado de una redacción fechada

`Eval.RedaccionFechada *RedaccionFechadaEsperada` (`redaccion_fechada`):

| Campo | YAML | Forma | Regla |
|---|---|---|---|
| `Norma` | `norma` | `BOE-A-…` | obligatorio |
| `Bloque` | `bloque` | id de bloque | obligatorio |
| `Fecha` | `fecha` | AAAA-MM-DD | obligatorio |
| `FechaVigencia` | `fecha_vigencia` | AAAAMMDD | con `posterior`, excluye `vigente_desde` |
| `Posterior` | `posterior` | `{fecha_vigencia: AAAAMMDD, norma_modificadora: BOE-A-…}` | solo con `fecha_vigencia` |
| `VigenteDesde` | `vigente_desde` | AAAAMMDD | excluye `fecha_vigencia` y `posterior` |

Una eval es **con fecha** si lleva `redaccion_fechada` (`esperaUnaRedaccionFechada`, `internal/evals/informe.go`). El esquema la prohíbe en una eval
con `activa: false` y en la eval sin binario ni servidor (FR-052). `normasEsperadas` no la mira: la norma de
`posterior` es un dato de la respuesta, no una norma que la eval lea.

### 7.3 Las líneas de la respuesta

`ExtraerRedaccionesFechadas(respuesta) []LineaFechada` (`internal/evals/fechadas.go`, nuevo): por cada línea con una de
las tres formas fijas y una cita, la forma, la norma y el bloque de su primera cita y, detrás de la cita, la primera
fecha AAAA-MM-DD, la primera fecha AAAAMMDD y, en la de la redacción de hoy, el primer identificador `BOE-A-…`
(contracts/evals-redaccion-fechada.md §2).

### 7.4 Lo medido de cada respuesta y el umbral

`respuestaDelJuez` gana `conFecha` (su eval lleva `redaccion_fechada`) y `sinFechar` (cuenta en el umbral);
`respuestasMedidas` gana `conRedaccionFechada`. El umbral `redaccion_sin_fechar:<modelo>:<modo>`: `total`, las
respuestas `conFecha` del grupo del modo; `medida`, las `sinFechar`; `comparacion` `<=`; `umbral` 0; `decide: true`
(FR-054, FR-055). Contrato completo en contracts/evals-redaccion-fechada.md §3 y §4.

### 7.5 Las reglas del conjunto

`conjuntoAComprobar` aparta las evals con `redaccion_fechada`, como la eval sin binario ni servidor: no son
`positivas` ni cuentan en `citadaPor`. `maximoDeEvals` pasa de 21 a 23; lo demás, igual (research D15).

## 8. Los errores (ADR 0023)

| Caso | Clase | Código | Dirección | Instante |
|---|---|---|---|---|
| `--fecha` que no es un día AAAA-MM-DD | `argumentos` | 2 | ninguna (firma el kernel) | ninguno |
| Fecha anterior a la primera vigencia (FR-003) | `no-encontrado` | 3 | la del bloque | la fecha de consulta de las versiones |
| Bloque no situable (FR-004) | `no-encontrado` | 3 | la del bloque | ídem |
| Norma o bloque mal formados, bloque que no está en la norma, fuente caída o con límite, bloque ilegible, metadatos caídos | los de hoy | 2, 3, 4, 5 | los de hoy | los de hoy (FR-007); sin escribir la entrada de las versiones |
| `--offline` sin la entrada de las versiones | `fuente-no-disponible` | 4 | la del bloque | ninguno (FR-020) |
