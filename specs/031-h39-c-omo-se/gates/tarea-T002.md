# T002: por qué no quedó en verde

## Intento 1: tarea mal delimitada, redelimitada (2026-11-03)

La tarea no declaraba `internal/app/grafo_test.go`, y el código que pide la deja en rojo: `compruebaEntradaDeCheck`
fija las propiedades `norma` y `bloques` de la entrada de `graph check --describe` sin `description`. Research V8 lista
los ficheros que fijan bytes de `--describe` (`schemas/`, el golden de plazos y `h20-fecha-main.txtar`), pero no este
test, que fija la forma decodificada. Salida de `go test -count=1 ./internal/app/ -run '^TestAppletGrafo$'` con el
código de la tarea y antes de tocar el test, ejecutada en este intento:

```
--- FAIL: TestAppletGrafo (0.00s)
    --- FAIL: TestAppletGrafo/ayuda-y-describe (0.00s)
        grafo_test.go:496:
            	Error Trace:	internal/app/grafo_test.go:612
            	Error:      	Not equal:
            	            	expected: map[string]interface {}{"type":"string"}
            	            	actual  : map[string]interface {}{"description":"Identificador BOE de la norma, BOE-A-<año>-<número>; sin él, todo lo consultado.", "type":"string"}
        grafo_test.go:496:
            	Error Trace:	internal/app/grafo_test.go:614
            	Error:      	Not equal:
            	            	expected: map[string]interface {}{"items":map[string]interface {}{"type":"string"}, "type":"array"}
            	            	actual  : map[string]interface {}{"description":"Ids de bloque de esa norma, como a21; sin ellos, todos los suyos.", "items":map[string]interface {}{"type":"string"}, "type":"array"}
```

El guardián de diff de este intento tiene congeladas las rutas de la línea anterior (`gates/tarea-actual.json`), sin
`internal/app/grafo_test.go`: cambiarlo y marcar la tarea habría hecho que rechazara el diff. La línea de T002 en
`tasks.md` lo declara ahora, y la tarea queda `[ ]`.

### Lo que deja este intento en el árbol (lo hereda el siguiente)

Todo lo de la tarea, más el arreglo del test que faltaba:

- `internal/cli/describe.go`: `describirArgumentos(entrada, campos)`, llamada en `esquema()` tras `anotarBanderas`, y
  `sinEscaparHTML` en `codificar`. `Schema.MarshalJSON` de invopop serializa con `json.Marshal`, que escapa `<`, `>` y
  `&` aunque el codificador de fuera lleve `SetEscapeHTML(false)` (el supuesto de T001 lo anotó): sin deshacerlo, la
  ayuda de `norma` salía con escapes de JSON y el guion congelado `describe.txtar` no pasaría. El reemplazo deja como
  está la barra escapada y la prueba primero, de modo que una barra que es texto de la ayuda no se toma por un escape.
- `internal/cli/herramienta.go`: `EsquemasDeHerramienta` llama a `describirArgumentos`.
- `internal/cli/describe_test.go`: `TestAyudaComoDescripcion` (los dos esquemas; una ayuda con `<`, `>`, `&` y `$`, una
  propiedad sin ayuda, una con ayuda vacía y las ocho globales; el resto del esquema igual al del gemelo sin ayudas; y
  la escritura con una barra literal). En rojo antes del código, por la `description` que faltaba; en verde después.
  Un mutante que quita el par de la barra escapada de `sinEscaparHTML` lo pone en rojo (revertido). Actualizada la
  expectativa del `Literal` de `TestDescribeGramatica`.
- `internal/cli/herramienta_test.go`: `entradaDeArticulo` con las dos `description` (229 bytes).
- `internal/app/grafo_test.go`: `compruebaEntradaDeCheck` con las dos `description`.
- `schemas/{norma,bloque,municipio,instalacion,grafo,cita,plazo}.json` regenerados con `-actualizar-esquemas` (antes,
  `make schema-check` en rojo con las 15 partes de esos siete ficheros); el diff solo añade 36 líneas `description`;
  `servidor.json` no cambia. `internal/app/testdata/plazos/describe.json` con `-actualizar-golden-de-plazos` (cinco
  `description`). La sección `main-articulos.describe` de `h20-fecha-main.txtar`, idéntica byte a byte a la salida de
  `go run ./cmd/kitlegal boe articulos BOE-A-2015-10565 a21 a22 --describe` (`git diff --no-index --exit-code`), con
  una nota en la cabecera del guion.
- `CHANGELOG.md`: la entrada de FR-005 en *Unreleased*, *Cambiado*.

### Verificación ejecutada en este intento, con ese árbol

En primer plano y por objetivos, todos en verde: `make fmt-check lint` (0 issues), `make test`, `make test-integration`,
y `make fmt-check lint test-tiempos vuln schema-check skills-check goreleaser-check secrets mod-verify mod-tidy-check`.
`git status --short` no lista ningún fichero de `skills/`. La salida de `go run ./cmd/kitlegal` con `boe articulo
BOE-A-2015-10565 a21 --describe`, `plazos calcular x x x --describe` y `skills install --describe` lleva cada
`description` que afirma `specs/031-h39-c-omo-se/aceptacion/describe.txtar` y las ocho globales sin ella, comprobado
con `grep` sobre la salida; el guion activado no se ha ejecutado en este intento.

### Lo que le queda al intento siguiente

Repetir `make ci` con el árbol heredado y, en verde, marcar T002 `[X]`.

## Intento 2: en verde, T002 marcada (2026-11-03)

Con el árbol heredado y sin cambiar ningún fichero de código ni de datos:

- El paso `[datos]` se repitió para verlo en este intento: `schemas/` y `internal/app/testdata/plazos/describe.json`
  restaurados de `HEAD`; `make schema-check` en rojo con las 15 partes de los siete ficheros (`servidor.json` no sale);
  regenerados con `-actualizar-esquemas` (`TestEsquemasPublicados`) y `-actualizar-golden-de-plazos`
  (`TestAppletPlazos`); `diff -r` contra las copias heredadas: idénticos. El diff de `schemas/` y del golden solo
  añade líneas `description` y la coma que las precede. La sección `main-articulos.describe` la compara
  `cmp stdout` del guion con la salida del binario del árbol en `make test-integration`.
- Verificación en primer plano y por objetivos, todos en verde: `make fmt-check lint` (0 issues), `make test`,
  `make test-integration` y `make test-tiempos vuln schema-check skills-check goreleaser-check secrets mod-verify
  mod-tidy-check`. Es la lista completa de `make ci`.
- `git status --short` no lista ningún fichero de `skills/`.
