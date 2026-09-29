# Data model: H7.2

Entidades del spec («Key Entities») y lo que el plan les da en el repositorio. Ninguna vive en el binario: todas son del
formato común de eval y del job (`internal/evals`), de sus tests o de los datos de las evals. Decisiones en
[research.md](./research.md) (D1-D18); formatos exactos en [contracts/](./contracts/).

## 1. Lista de expresiones prohibidas

Una por skill, opcional. Hoy solo `boe-legislacion` la tiene (FR-050).

| Aspecto | Valor |
|---|---|
| Fichero | `evals/<skill>/expresiones-prohibidas.yaml` |
| Esquema | `schemas/expresiones-prohibidas.yaml.json` (research D1, D2) |
| Claves | `maquinaria` y `otra_conversacion`, las dos obligatorias; ninguna otra |
| Cada familia | lista no vacía de expresiones |
| Expresión | una o más palabras separadas por un espacio; sin blancos en los extremos, sin `*` ni `_` (`^[^\s*_]+( [^\s*_]+)*$`) |
| Tipo Go | `evals.ExpresionesProhibidas{Maquinaria, OtraConversacion []string}` (`yaml:"maquinaria"`, `yaml:"otra_conversacion"`) |
| Quién la lee | `LeerConjunto`, que la reconoce por su nombre exacto y la valida; mal formada, es un `FicheroMalFormado` (FR-055) |
| Dónde queda | `Conjunto.Prohibidas` y, copiada, `Eval.Prohibidas` de cada eval de la carpeta (research D5); vacía si la carpeta no la tiene o está mal formada |

Contenido de la de `boe-legislacion`: las 38 expresiones de research D4 (contracts/lista-y-juicio.md §2).

## 2. Expresión encontrada

Una expresión de la lista presente en una respuesta, con la tolerancia de H5.1.

- **Comparación** (research D3): cada expresión se compila una vez en
  `(?:^|[^\p{L}\p{N}])` + palabras `(?i:QuoteMeta(p))` unidas por `entrePalabrasDeAviso` + `(?:$|[^\p{L}\p{N}])`.
- **Función**: `evals.ExtraerExpresionesProhibidas(texto string, lista ExpresionesProhibidas) []string` —las
  encontradas, en el orden de la lista (maquinaria y después otra conversación), sin repetir; `nil` si ninguna—. La
  usan `Juzgar`, las subpruebas de calibrado, de bloques y de la skill, y la comprobación del quickstart (FR-061: «el
  mismo código que el job»).
- **En el resultado**: `ResultadoDeEval.ExpresionesProhibidas []string` (`json:"expresiones_prohibidas"`), vacía (`[]`)
  si no hay ninguna o si la eval es de no activación; cada una, un motivo `expresión prohibida: <expresión>` y la
  sesión no pasa (FR-052).

## 3. Recuento por modelo

`evals.RecuentoDeExpresiones{Modelo string; ConAlguna int; Respuestas int}` (`json:"modelo"`, `"con_alguna"`,
`"respuestas"`), en `Informe.ExpresionesProhibidasPorModelo` (`json:"expresiones_prohibidas_por_modelo"`), detrás de
`tasas` (FR-053; research D7).

- Un elemento por modelo del job: el que decide y después los informativos, en su orden.
- `Respuestas`: sesiones juzgadas (no ilegibles) de las series planificadas cuya eval activa la skill.
- `ConAlguna`: de esas, las que tienen alguna expresión encontrada.
- No cuentan las series no planificadas: la de la prueba de red y la de una sesión cuya eval, modelo o pregunta no se
  pudo leer.
- Skill sin lista: `[]`.

Con el conjunto actual: `claude-sonnet-5` sobre 51 y `claude-haiku-4-5-20251001` sobre 30, con o sin prueba de red.

## 4. Eval de la consulta repetida

`evals/boe-legislacion/19-lcsp-contrato-menor-redaccion-cambiada.yaml` (FR-001 a FR-004; research D12):

| Campo | Valor |
|---|---|
| `pregunta` | «Hace tiempo te pregunté qué exige el artículo 118 de la LCSP para el expediente de un contrato menor. ¿Qué dice ahora?» |
| `activa`, `informativa` | `true`, `true` |
| `grafo_previo` | `grabaciones: lcsp-a1-30-redaccion-original`; comando `boe BOE-A-2017-12902 a1-30` |
| `comandos` | bloque `boe BOE-A-2017-12902 a1-30`; comprobación `graph check` con `norma: BOE-A-2017-12902` |
| `prohibidos` | `graph show` |
| `citas` | `BOE-A-2017-12902 a1-30` |
| `hallazgos` | `version-obsoleta` |

Juicio: el de toda eval que activa la skill, lista incluida.

## 5. Redacción de la grabada

Cada `<version>` hija del `<bloque>` en el cuerpo de la respuesta grabada de un bloque, con `fecha_publicacion`,
`fecha_vigencia`, `id_norma` y su texto. Lo que de ella da `boe articulo` (`fecha_version`, `fecha_vigencia`,
`norma_modificadora`, `texto`, `hash_texto` y lo demás del `Articulo`) es lo que da sobre la grabación reducida a esa
sola versión. La grabación H4 del bloque `a1-30` de `BOE-A-2017-12902` trae dos (research V15):

| Versión | `id_norma` | `fecha_publicacion` | `fecha_vigencia` |
|---|---|---|---|
| Original | `BOE-A-2017-12902` | 20171109 | 20180309 |
| Vigente | `BOE-A-2020-1651` | 20200205 | 20200206 |

## 6. Derivada del grafo previo

Una respuesta grabada del BOE sin alguna de sus redacciones; servida en lugar de la grabada, da una redacción que la
grabada trae (FR-010 a FR-012; research D13).

| Aspecto | Valor |
|---|---|
| Carpeta | `testdata/evals/grafo-previo/<nombre>/`, un fichero con el nombre de la grabación de H4 que sustituye |
| La de H7.2 | `lcsp-a1-30-redaccion-original/GET_https_www.boe.es_datosabiertos_api_legislacion-consolidada_id_BOE-A-2017-12902_texto_bloque_a1-30.json` |
| Cómo se produce | la derivación de `internal/app/grafo_test.go` sobre la grabación y la fecha 20180309: quita las `<version>` posteriores y nada más; la escribe `-actualizar-derivadas` |
| Cómo se comprueba | `TestGrabacionesDerivadas`, por estar en su carpeta: byte a byte igual a la derivación, y servida en lugar de la grabación, `boe articulo` da exactamente una redacción de la grabada, la de 20180309 |
| Declaración | entrada de la lista de derivadas del grafo previo, de su propio tipo y separada de `grabacionesDerivadas()`: subcarpeta, fichero, argumentos de `boe` y fecha de vigencia de la redacción que da |

La carpeta decide la comprobación (contracts/eval-y-derivada.md §3): `TestGrabacionesDerivadas` compara los ficheros de
cada carpeta con las entradas de su lista, en los dos sentidos, y una entrada de otra clase con carpeta de grafo previo
lo hace fallar nombrándola. Las derivadas del e2e (`internal/app/testdata/derivadas/…`) siguen en
`grabacionesDerivadas()`, de su clase actual, con su comprobación (FR-013).

## 7. Comprobación de la consulta repetida (quickstart)

Entrada: dos directorios de conversación, cada uno con `sesion.jsonl`, `sesion.err` y `codigo-de-la-sesion` (los de
`LeerSesion`), la lista de la skill y las dos fechas. Salida: una línea por condición que falla (contracts/
comprobacion-del-quickstart.md §2). Condiciones: la primera respuesta lleva `⚠ REDACCIÓN MODIFICADA:` en una línea con
las dos fechas; la segunda no la lleva; ninguna lleva una expresión de la lista (FR-061). Antes, las dos
conversaciones legibles y terminadas.

## 8. Lo que se retira

| Qué | Dónde |
|---|---|
| Eval 19 anterior | `evals/boe-legislacion/19-lpac-articulo-21-redaccion-cambiada.yaml` |
| Su grafo previo | `testdata/evals/grafo-previo/lpac-a21-version-anterior/` |
| Su comprobación | la entrada `lpac-a21-version-anterior` de `grabacionesDerivadas()`; con ella, la comparación de las dos carpetas juntas, que pasa a ser carpeta a carpeta |
| Su párrafo sintético | `parrafoDeLaVersionAnterior`, en `internal/app/grafo_test.go` |
