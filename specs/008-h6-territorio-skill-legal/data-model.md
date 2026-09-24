# Modelo de datos: H6 · `territorio` + skill `legal-core` v0

Los tipos que H6 añade, con su forma exacta, sus reglas y de dónde sale cada dato. Las claves JSON y los nombres de
verbo van en español (`CLAUDE.md`, «Convenciones de idioma»); los identificadores Go siguen la convención del lenguaje.

Convenciones heredadas que aquí no se repiten: el sobre de salida y su huella (ADR 0006), el contrato del applet
(ADR 0005), la procedencia con fecha (ADR 0015) y el ensayo (ADR 0011). Convención de los tipos de `data`: **ninguna
etiqueta lleva `omitempty`**, todo campo se emite siempre, lo que no hay va como cadena vacía y una lista vacía va como
`[]` (research.md V13).

## 1. Identificadores (`internal/core/ids`)

### 1.1 Código INE de municipio

| Concepto | Forma | Regla |
|---|---|---|
| Código | `PPMMM`, cinco cifras | `PP` es el código de provincia, `01`-`52`; `MMM` es el del municipio dentro de ella, `001`-`999` |
| Código con dígito | `PPMMMD`, seis cifras | Las cinco primeras son el código; la sexta es el dígito de control **declarado por quien escribe** |
| Dígito de control | una cifra | Es **dato oficial** de la relación del INE, no se calcula (research.md D9) |

- `CodigoINE` es un valor inmutable que solo se construye analizando. `String()` devuelve siempre las cinco cifras con
  sus ceros por delante: **analizar y volver a escribir da la misma cadena** y normalizar lo normalizado no lo cambia
  (FR-032).
- `Provincia()` devuelve las dos primeras cifras.
- `ComprobarDigito(declarado, oficial byte) error` es **método de `CodigoINE`**, no función de paquete (research.md
  D8): compara el dígito declarado con el oficial y su error nombra los dos. La firma es la misma en este documento,
  en D8 y en el contrato [identificadores-ine-y-dir3](./contracts/identificadores-ine-y-dir3.md) §1 y §3, porque
  `TestSuperficieDeIds` exige que lo exportado sea **exactamente** esa lista.
- Entradas inválidas: cadena vacía, cualquier carácter que no sea cifra ASCII, menos de cinco o más de seis cifras,
  provincia fuera de `01`-`52`, municipio `000`. Todas dan un error de clase `argumentos` (código 2), nunca `panic`
  ni valor por omisión (FR-033).

### 1.2 Código DIR3 de un ayuntamiento

| Concepto | Forma | Regla |
|---|---|---|
| Código | `L01PPMMMD` | `L` + `01` + código INE de cinco cifras + dígito de control |
| Relación con el INE | `DIR3.CodigoINE()` | Devuelve `PPMMM`; `DIR3.Digito()`, la última cifra |
| Construcción | `DIR3DeAyuntamiento(ine, digito)` | Es la única forma de componer uno desde el dominio |

- La letra se analiza sin distinguir mayúscula y minúscula y se **normaliza a mayúscula**; el resto son cifras.
- `AnalizarDIR3` rechaza cualquier otra forma con clase `argumentos`. El paquete no implementa ELI, ECLI, CELEX ni NIF
  (FR-034).
- Ida y vuelta estable: `AnalizarDIR3(d.String())` devuelve el mismo valor (FR-032).

### 1.3 Errores

Un tipo propio del paquete con `Clase() schema.Clase` = `schema.ClaseArgumentos`, que el kernel traduce a código 2 sin
que el dominio importe `internal/cli` (research.md V7, V18). El mensaje nombra la entrada con `%q` y dice qué tiene de
malo, como los de `internal/source/boe/ids.go`.

## 2. Territorio (`internal/core/territorio`)

### 2.1 Fuentes y registro

```go
type Fuentes struct {
    Municipios  []byte            // data/territorio/municipios.yaml
    DIR3        []byte            // data/territorio/dir3.yaml
    Estado      []byte            // data/territorio/estado.yaml
    Comunidades map[string][]byte // data/territorio/comunidades/<código>.yaml, por código
}

// Los cuatro ficheros ya analizados, en la forma de §3. Los exporta el dominio
// —con FicheroDeMunicipios, FicheroDeDIR3, FicheroDeEstado y FicheroDeComunidad—
// para que internal/skills los valide contra su esquema con
// ValidarDocumentoYAML[T] sin repetir los tipos (FR-044, research.md D27).
type Ficheros struct {
    Municipios  FicheroDeMunicipios
    DIR3        FicheroDeDIR3
    Estado      FicheroDeEstado
    Comunidades map[string]FicheroDeComunidad
}

func Cargar(f Fuentes) (*Registro, error)
func (r *Registro) Resolver(entrada string) (Territorio, error)

// El pliegue de nombres (§2.7), exportado porque lo comparten el registro, el
// juicio de evals (contrato de evals §2) y el control del corpus congelado
// (research.md D10, D27). Una sola definición del pliegue en el árbol.
func Plegar(nombre string) string
```

`Cargar` decodifica `Fuentes` en `Ficheros`, comprueba la integridad entre ficheros y construye los índices.
`Resolver` no vuelve a leer nada. El dominio recibe bytes y no un sistema de ficheros (research.md D3), y **ni él ni
sus tests importan el paquete `data`**: los tests del dominio son sintéticos y los que miran los ficheros congelados
reales viven en `internal/skills` (research.md D27).

**Integridad que `Cargar` exige** (un fallo es error de carga, no de resolución):

1. Toda provincia citada por un municipio está declarada por alguna comunidad, y ninguna provincia está en dos.
2. Todo municipio de `dir3.yaml` existe en `municipios.yaml`, y el DIR3 es coherente con su código INE y su dígito
   (`DIR3.CodigoINE() == código` y `DIR3.Digito() == dc`).
3. Ningún código de municipio, de provincia o de comunidad repetido (la clave repetida ya la rechaza el lector, V37).
4. El régimen de toda comunidad es `comun` o `foral`.
5. El nombre del fichero de cada comunidad coincide con el código que declara dentro.
6. **La `comunidad` de cada municipio es la de la comunidad que declara su provincia.** Es la única redundancia del
   modelo: FR-040 exige la columna en la fila, y la provincia lleva a la misma comunidad por el otro camino. El
   camino **autoritativo es el de la provincia** —es el que sostiene el régimen y los boletines, que viven en el
   fichero de la comunidad—, y esta comprobación impide que la columna de la fila discrepe de él en silencio: una
   discrepancia es error de carga y hace fallar `make ci` nombrando el municipio, las dos comunidades y sus ficheros.

### 2.2 Municipio, provincia, comunidad, régimen

| Entidad | Campos | Origen | Id natural (H7) |
|---|---|---|---|
| **Municipio** | nombre oficial, código INE, dígito de control, provincia, comunidad | `ine.municipios` | código INE |
| **Provincia** | código de dos cifras, nombre, comunidad | `ine.codigos-territoriales` (nombres, en el fichero de su comunidad) | código |
| **Comunidad** | código, nombre, régimen, provincias, boletines (opcional) | `ine.codigos-territoriales` (nombres) y la configuración | código |
| **Ayuntamiento** | código DIR3 | `mpt.rel`, solo si está verificado | código DIR3 |
| **Régimen** | `comun` \| `foral` | La configuración de su comunidad | — |

La **comunidad del municipio** viaja en su fila porque FR-040 lo exige (§3.1), y el registro la resuelve por su
provincia, que es el camino autoritativo; los dos nunca pueden discrepar, porque la integridad de §2.1 (punto 6) los
ata. La `comunidad` que emite `data` (§2.5) es, por tanto, una sola.

El **régimen es dato nacional**: lo tienen las 19 comunidades y ciudades autónomas, esté o no configurada su parte de
boletines (FR-055). Son forales Navarra y el País Vasco; el resto, común, incluidas Ceuta y Melilla (spec, US3 y casos
límite).

### 2.3 Boletín aplicable

| Campo | Clave JSON | Regla |
|---|---|---|
| Nivel | `nivel` | `estatal` \| `autonomico` \| `provincial` |
| Código | `codigo` | El del boletín (`BOE`, `BOCM`…) |
| Nombre | `nombre` | El oficial |
| Dirección | `url` | La pública del boletín |
| Motivo | `motivo` | Vacío salvo cuando el mismo boletín cubre dos niveles: entonces, la razón que fija su configuración |
| Procedencia | `source` | El fichero de `data/territorio/` que lo fija |

La lista lleva **siempre** el estatal y **solo** los niveles configurados (FR-008, FR-021). En una comunidad
uniprovincial la equivalencia **no se deduce**: la declara su configuración (spec, caso límite; FR-052).

### 2.4 Cobertura

Exactamente tres claves, siempre presentes, con vocabulario cerrado (FR-020):

| Clave | Valores | Significa |
|---|---|---|
| `boletin_autonomico` | `configurado` \| `no-configurado` | Si la comunidad tiene declarado su boletín autonómico |
| `boletin_provincial` | `configurado` \| `no-configurado` | Ídem para el provincial |
| `dir3` | `verificado` \| `no-verificado` | Si el DIR3 de ese ayuntamiento está en la correspondencia verificada |

**Ningún valor significa «no existe»**, y esa ausencia es lo que hace mecánicamente cierto FR-022: un test exige que
los enumerados sean exactamente esos.

### 2.5 Territorio resuelto: el `data` del verbo

Ocho claves de primer nivel. Cada dato que sale de un fichero lleva su `source` (FR-005, FR-006); `cobertura`, que dice
lo que el registro tiene configurado o verificado, no lo lleva:

```json
{
  "municipio":  {"nombre": "…", "source": "ine.municipios"},
  "codigo_ine": {"codigo": "PPMMM", "digito_de_control": "D", "source": "ine.municipios"},
  "provincia":  {"codigo": "PP", "nombre": "…", "source": "ine.codigos-territoriales"},
  "comunidad":  {"codigo": "CC", "nombre": "…", "source": "ine.codigos-territoriales"},
  "dir3":       {"codigo": "L01PPMMMD", "source": "mpt.rel"},
  "regimen":    {"valor": "comun", "source": "data/territorio/comunidades/CC.yaml"},
  "boletines":  [ {"nivel": "…", "codigo": "…", "nombre": "…", "url": "…", "motivo": "", "source": "…"} ],
  "cobertura":  {"boletin_autonomico": "…", "boletin_provincial": "…", "dir3": "…"}
}
```

**Vocabulario de `source`** (FR-005): el identificador de la fila de `docs/SOURCES.md` cuando el dato viene de una
descarga (`ine.municipios`, `ine.codigos-territoriales`, `mpt.rel`), o la ruta del fichero de `data/territorio/` cuando lo fija la configuración.
Un test exige que todo `source` emitido sea una de las dos cosas y que los identificadores existan en
`docs/SOURCES.md`.

**Invariante del DIR3 no verificado** (FR-023), atada en test:
`dir3.codigo == ""` ⟺ `dir3.source == ""` ⟺ `cobertura.dir3 == "no-verificado"`.
Nunca se emite un código derivado como si fuera registral.

### 2.6 Entrada y resolución

```
entrada plegada → ¿vacía? ── sí ──→ no nombra nada (2)
                            └─ no ──→ ¿alguna letra? ── no ──→ código INE (5 o 6 cifras)
                                                      └─ sí ──→ nombre
```

| Caso | Código | Qué devuelve |
|---|---|---|
| Código de cinco cifras de un municipio de la relación | 0 | El territorio |
| Código de seis cifras con el dígito oficial | 0 | El mismo `data` y la misma huella que por cinco cifras y que por nombre (SC-001) |
| Código de seis cifras con dígito distinto del oficial | 2 | Mensaje que dice el dígito recibido y el oficial |
| Entrada vacía o hecha solo de espacios y separadores | 2 | «la consulta … no nombra ningún municipio: no tiene ninguna letra ni ninguna cifra» |
| Sin ninguna letra y sin formar un código (4 o 7 cifras, provincia 00 o 53…, municipio 000, un separador o una cifra no ASCII entre las cifras, como `28-074`) | 2 | Mensaje que dice qué tiene de malo |
| Código **bien formado** —provincia `01`-`52` y municipio `001`-`999`— que no está en la relación | 3 | — |
| Nombre que corresponde a un municipio | 0 | El territorio |
| Nombre que corresponde a varios | 2 | Mensaje con **todos** los candidatos, ordenados por código INE |
| Nombre que no corresponde a ninguno | 3 | — |

El applet **nunca** decide 4, 5 ni 6 (FR-016); solo el plazo agotado de `--timeout`, que pone el kernel para todo
applet, termina en 4 (FR-020 de H1).

### 2.7 Normalización de nombres

1. Se pliega la entrada y cada forma conocida del municipio: minúsculas, letras con diacrítico a su letra base, signos
   separadores y espacios colapsados.
2. Formas conocidas de un municipio, derivadas **del nombre oficial** y sin caso especial por municipio (FR-024):
   - el nombre oficial tal como lo escribe el INE (`Coruña, A`);
   - con el artículo pospuesto antepuesto (`A Coruña`);
   - cada lado de un nombre bilingüe partido por `/` (`Donostia`, `San Sebastián`), y la forma con artículo de cada uno.
3. Índice: forma plegada → conjunto de municipios. Una forma que lleva a más de uno es **ambigüedad declarada**
   (código 2 con candidatos), nunca una elección.
4. Tres controles mecánicos sobre el corpus congelado (research.md D10), que por leer los ficheros reales viven en
   `internal/skills` y no en el dominio (research.md D27): toda runa de todo nombre está cubierta por el pliegue,
   ningún municipio queda inalcanzable por efecto de la normalización —o se resuelve él, o la ambigüedad lo nombra
   entre sus candidatos— y ningún nombre plegado se queda sin letras, forma que §2.6 lee como código.

### 2.8 Fecha de la respuesta

`Procedencia.FechaConsulta` es la **más antigua** de las fechas (`fecha` de la raíz) de los ficheros que sostienen ese
`data`: municipios, la comunidad del municipio, estado y, si la trae, la correspondencia DIR3. A medianoche UTC
(research.md D6). De ahí que dos ejecuciones den la misma salida byte a byte, con `--offline` y sin él.

## 3. Ficheros congelados (`data/territorio/`)

Todos llevan en la raíz `fecha` (la del fichero de origen, `AAAA-MM-DD`) y `source`, y se validan contra su esquema en
`make ci`. Se generan en tareas `[datos]` con pausa (FR-045).

### 3.1 `municipios.yaml` — esquema `schemas/territorio-municipios.yaml.json`

```yaml
fecha: 2026-02-04
source: ine.municipios
municipios:
  "28074": {dc: "5", nombre: "Leganés", provincia: "28", comunidad: "13"}
```

`propertyNames.pattern` `^(0[1-9]|[1-4][0-9]|5[0-2])(00[1-9]|0[1-9][0-9]|[1-9][0-9]{2})$`, el código INE de §1.1:
provincia de `01` a `52` y municipio de `001` a `999`; `dc` una cifra; `nombre` no vacío; `provincia`
`^(0[1-9]|[1-4][0-9]|5[0-2])$`, de `01` a `52` (el mismo patrón que la clave de `provincias` de §3.4); `comunidad` dos
cifras. Los patrones del código INE y de la provincia aceptan exactamente lo que acepta `AnalizarCodigoINE`, y los ata
el subtest `TestTerritorioDelRepositorio/gramaticas` (contrato de identificadores §2). Las cinco columnas de FR-040,
**siempre las cinco**: es la misma forma que enseñan §2.2 y research.md D4, y la coherencia de `comunidad` con la
comunidad que declara la provincia la exige §2.1, punto 6. Una línea por municipio (research.md D4). Los valores de
este ejemplo son ilustrativos: los fija la tarea `[datos]` desde la relación del INE (research.md S1, S5).

### 3.2 `dir3.yaml` — esquema `schemas/territorio-dir3.yaml.json`

```yaml
fecha: 2026-09-21
source: mpt.rel
correspondencia:
  "28074": "L01280745"
```

Solo filas verificadas (FR-048). `propertyNames.pattern` el del código INE de §3.1,
`^(0[1-9]|[1-4][0-9]|5[0-2])(00[1-9]|0[1-9][0-9]|[1-9][0-9]{2})$`; valor
`^[Ll]01(0[1-9]|[1-4][0-9]|5[0-2])(00[1-9]|0[1-9][0-9]|[1-9][0-9]{2})[0-9]$`, el DIR3 de §1.2: la letra en mayúscula o
en minúscula, como `AnalizarDIR3`, que la normaliza a mayúscula (el fichero la escribe en mayúscula), el tipo `01`, un
código INE en rango y el dígito. Los dos aceptan exactamente lo que aceptan `AnalizarCodigoINE` y `AnalizarDIR3`, y los
ata el mismo subtest `gramaticas`.

### 3.3 `estado.yaml` — esquema `schemas/territorio-estado.yaml.json`

```yaml
fecha: 2026-09-20
source: data/territorio/estado.yaml
boletin:
  codigo: BOE
  nombre: "Boletín Oficial del Estado"
  url: "https://www.boe.es/"
```

### 3.4 `comunidades/<código>.yaml` — esquema `schemas/territorio-comunidad.yaml.json`

```yaml
fecha: 2026-09-20
source: ine.codigos-territoriales
codigo: "13"
nombre: "Comunidad de Madrid"
regimen: comun
provincias:
  "28": "Madrid"
boletines:                       # ausente = territorio no configurado
  autonomico: {codigo: BOCM, nombre: "Boletín Oficial de la Comunidad de Madrid", url: "…"}
  provincial: {codigo: BOCM, nombre: "…", url: "…", motivo: "Comunidad uniprovincial: el BOCM hace también de boletín provincial."}
```

- `regimen` es obligatorio en las 19 (FR-055); `boletines` solo lo trae la Comunidad de Madrid en este hito (FR-051).
- `provincias` lleva al menos una.
- Añadir un territorio es rellenar `boletines` en su fichero: ni código, ni skills (FR-053).

## 4. Normas vertebrales y jerarquía

### 4.1 `data/normas.yaml` gana `vertebral`

Campo booleano **opcional** en cada norma (`vertebral: true`), añadido a `schemas/normas.yaml.json`, que declara
`additionalProperties: false` (spec, *Clarifications* Q2; FR-067). Lo llevan las quince leyes de la tabla de
`refs/mapa-sistema-legal-skills.md` §1.4 y ninguna más (FR-070). Ocho ya están en el fichero; siete entran con este
hito y sus identificadores se copian **de la respuesta grabada** de `boe buscar`, nunca de `refs/` (FR-071;
research.md S7).

### 4.2 `data/jerarquia.yaml` — esquema `schemas/jerarquia.yaml.json`

```yaml
niveles:
  - nivel: ue
    nombre: "Unión Europea"
    boletin: "DOUE"
    normas: ["Reglamento", "Directiva", "Decisión"]
  # estado, comunidad-autonoma, provincia, municipio
reglas:
  - regla: competencia-antes-que-jerarquia
    enunciado: "…"
```

`nivel` es un enumerado de cinco valores en orden; `regla`, un enumerado de las cuatro reglas de interpretación
(competencia antes que jerarquía, ley posterior, ley especial, reglamento nunca contra ley). El esquema fija además,
con `prefixItems` e `items: false`, que `niveles` trae los cinco niveles y `reglas` las cuatro reglas, cada uno una
vez y en el orden de su enumerado: un nivel que falta, sobra, se repite o cambia de sitio es un defecto en la línea del
elemento, y el orden del documento es ya el del enumerado. `boletin` es la clase de boletín del nivel, no un boletín
concreto: el de cada comunidad y provincia lo da `territorio resolver` desde `data/territorio/` (FR-051). `normas`
son los tipos de norma del nivel de mayor a menor rango, sin repetir. El contenido sale de
`refs/mapa-sistema-legal-skills.md` §1.1 y §1.2. De aquí sale `references/jerarquia_normativa.md` (FR-066).

## 5. Skill `legal-core`

| Pieza | Contenido |
|---|---|
| `SKILL.md` | Frontmatter válido, < 300 líneas, protocolo que empieza por el territorio, región generada de comandos, reglas invariantes (FR-061 a FR-064) |
| `metadata.kitlegal-applets` | `territorio` |
| `metadata.kitlegal-referencias` | `leyes_vertebrales jerarquia_normativa` |
| `references/leyes_vertebrales.md` | Generada desde `data/normas.yaml`, solo `vertebral: true` |
| `references/jerarquia_normativa.md` | Generada desde `data/jerarquia.yaml` |
| `scripts/territorio` | Enlace a `../../../bin/instalado/kitlegal` |

**Tabla de generadores** (research.md D20), que sustituye al `switch` por nombre:

| Referencia | Fichero de datos | Qué emite |
|---|---|---|
| `normas` | `data/normas.yaml` | Todas las normas (sin cambio) |
| `leyes_vertebrales` | `data/normas.yaml` | Las marcadas `vertebral: true`, mismas columnas |
| `jerarquia_normativa` | `data/jerarquia.yaml` | Una tabla de niveles y una lista de reglas |

## 6. Formato de eval: lo que H6 añade

### 6.1 Cuarta variante de comando esperado

| Variante | Claves | Discriminante |
|---|---|---|
| Bloque (H5) | `applet`, `norma`, `bloque` | sin `verbo` |
| Consulta de norma (H5) | `applet`, `verbo` ∈ {`indice`,`metadatos`,`analisis`}, `norma` | verbo del enumerado |
| Búsqueda (H5) | `applet`, `verbo: buscar`, `terminos` | `buscar` |
| **Territorio (H6)** | `applet`, `verbo: resolver`, `municipio` | `resolver` |

La forma se decide **en un solo sitio** y la consumen el juicio, el texto del comando y las consultas necesarias
(research.md D21, punto 2). Una invocación satisface el comando de territorio cuando su applet es `territorio`, su
verbo `resolver` y su argumento, plegado, es el municipio esperado.

### 6.2 Esperado de territorio

```yaml
territorio:
  comunidad: "Comunidad de Madrid"
  provincia: "Madrid"
  boletines: ["BOCM"]
  cobertura: ["boletin_autonomico: configurado"]
```

Cada elemento se busca en la respuesta por su **forma fija**, como la cita por su identificador y el aviso por su
etiqueta; se reparte en encontrados y ausentes, y un ausente impide que la eval pase.

### 6.3 Consultas necesarias

Un comando de territorio **no genera ninguna consulta que grabar**: el applet no pide nada por red ni usa caché
(FR-043). La regla «lo grabado basta para cada eval» sigue aplicándose solo a los comandos del BOE.

### 6.4 Reglas del conjunto de `legal-core`

| Regla | Exige |
|---|---|
| `tamaño` | Al menos tres evals bien formadas |
| `cubierto` | Al menos una eval activa sobre un municipio del territorio configurado |
| `no cubierto` | Al menos una eval activa sobre un municipio de una comunidad sin configuración |
| `no activación` | Al menos una eval con `activa: false` |
| `esperado verificable` | Toda eval activa declara `citas` o `territorio` |

Las diez reglas de `boe-legislacion` no cambian (research.md D22).
