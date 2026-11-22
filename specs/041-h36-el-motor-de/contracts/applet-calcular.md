# Contrato: el esquema de `plazos calcular` y el Go que no nombra los datos

Spec §C.7 y §C.8, FR-010 y FR-012. La orden, la herramienta `plazos_calcular`, sus argumentos, sus banderas, la
descripción del verbo y el sobre no cambian.

## 1. Lo que cambia en `internal/app/plazos.go` e `internal/core/plazos/salida.go`

- Las etiquetas `help` de `Tipo`, `Inicio`, `Municipios`, `Duracion` y `Unidad` de `argumentosDeCalcular` son, byte a
  byte, las cinco de §C.7.
- `Plazo.Tipo` y `Aviso.Codigo` llevan `jsonschema:"pattern=^[a-z0-9]+(-[a-z0-9]+)*$"` y ningún `enum`. El comentario de
  cabecera de `salida.go` deja de nombrar `CodigosDeTipo` y `CodigosDeAviso` entre los vocabularios que repiten los
  enumerados; los de `festivos[].ambito`, `cobertura[].ambito` y `cobertura[].falta` siguen, y su test también.
- Nada más del applet: `--help` y `--describe` ya no cargan los datos (research D5).

## 2. Los goldens que cambian (tarea `[datos]`, FR-017)

| Fichero | Cómo se regenera | Qué cambia |
|---|---|---|
| `internal/app/testdata/plazos/describe.json` | `go test -count=1 -run '^TestAppletPlazos$' ./internal/app/ -actualizar-golden-de-plazos` | las cinco ayudas; `tipo` y `codigo` con `pattern` y sin `enum` |
| `schemas/plazo.json` (la parte de `plazos calcular`) | `go test -count=1 -run '^TestEsquemasPublicados$' ./internal/app/ -actualizar-esquemas` | lo mismo |
| `internal/app/testdata/script/h39-describe.txtar` y `specs/031-h39-c-omo-se/aceptacion/describe.txtar` | a mano, idénticos | las líneas 39 a 43, con las ayudas de §C.7 |

`caso-a.json`, `caso-b.json` y `caso-e.json` no cambian (FR-021): la bandera los reescribe con la salida de la rama, que
es la misma (FR-011), así que quedan con los mismos bytes; si alguno cambiara, lo que está mal es el código, y se
arregla el código.

## 3. Los tests de `internal/app` que cambian o entran

- `compruebaDeclaracionDePlazos` (`plazos_test.go`): deja de exigir que la ayuda del tipo nombre cada código de
  `CodigosDeTipo`; sigue exigiendo que la de la unidad nombre las tres unidades.
- `compruebaEnumeradosDePlazos` (`plazos_test.go`): `tipo` de `plazos.Plazo` y `codigo` de `plazos.Aviso` llevan el
  `pattern` de §C.8 y ningún `enum` en `schemas/plazo.json`; los otros tres enumerados, como hoy.
- **`TestDescribeDePlazosSinDatos`** (nuevo, `internal/app/plazos_describe_test.go`; US4.1, SC-005): ejecuta
  `kitlegal plazos calcular x x x --describe` sobre tres juegos de `FicherosDePlazos`, por `registroDePlazos`:
  `fuentesDePlazos(t)`; los mismos con `Tipos` = `data/plazos.yaml` más un tipo de `lpac` y un periodo `aviso` de `lpac`
  con su código, su etiqueta y su texto (por sustitución de texto; la copia la lee `LeerTipos` sin defecto, que es lo
  que hace de ella «un tipo y un aviso más»); y los mismos con `Tipos` vacío, que no carga. Exige en los tres código 0 y
  la salida igual, byte a byte, a `describe.json`.
- `compruebaDefectoDeComposicionDePlazos` y el caso «sin-los-tipos» no cambian: ya exigen 0 en `--help` y
  `--describe` y 1 (`inesperado`) al ejecutar (US4.2).

## 4. `TestNingunCodigoDePlazosEnGo` (FR-010, SC-003; research D6)

`internal/app/plazos_codigos_test.go`.

- **Los códigos**: `tipos[].codigo`, `computos[].codigo` y `computos[].periodos[].aviso.codigo` de `data/plazos.yaml`
  (`../../data/plazos.yaml`) y de `../core/plazos/testdata/computo-de-ensayo.yaml`. Hoy: los siete tipos, `lpac`,
  `contencioso`, `civil`, `dias-del-art-183-lopj`, `ensayo-fijo`, `ensayo-libre`, `ensayo` y `periodo-de-ensayo`. El
  test exige que haya al menos uno de cada clase en cada fichero, para que un fichero que no se lee no lo deje en verde.
- **El Go**: cada `.go` que no acaba en `_test.go` de `../core/plazos`, `.`, `../skills` y `../evals`, sin
  subdirectorios, leído con `go/parser.ParseFile`.
- **Un defecto** por cada `*ast.BasicLit` de cadena cuyo valor sin comillas (`strconv.Unquote`) es entero un código, y
  por cada etiqueta de campo de un `struct` (`ast.Field.Tag`) que lo lleva como palabra entera (`\b<código>\b`, como
  `grep -w`). Cada defecto dice el fichero, la línea y el código.
- **Fuera**, sin código propio: los tres avisos del motor, que no son códigos de esos ficheros, y la descripción del
  verbo, que no es una etiqueta ni ningún literal suyo es entero un código.
- **En rojo** (mutantes de US3.6 y US4.4): `if tipo.Codigo == "silencio"` en `internal/core/plazos`, y la ayuda o los
  enumerados de hoy en las etiquetas.
