# Quickstart: validar H39

Escenarios para comprobar la entrega cuando las tareas estén hechas. Todos se ejecutan desde la raíz del repositorio.
Los de §1 a §4 no abren sesiones con modelo, no usan la red más que la de las herramientas de Go y no escriben en el
árbol de trabajo, el índice ni el historial: solo en las cachés de Go y, los tests, en sus temporales. Los de §5 y §6
los hace una persona fuera de cualquier run (ADR 0032): abren sesiones con modelo o ponen una credencial en el entorno,
que la política de las sesiones del run veda.

## 1. La ayuda llega a `--describe` (US1; FR-001, FR-003)

```sh
go run ./cmd/kitlegal boe articulo BOE-A-2015-10565 a21 --describe | grep -F '"description": "Id del bloque, del campo id de kitlegal boe indice (a21, da-3, a1-30…): no se compone con el número del artículo, que no siempre es su id."'
go run ./cmd/kitlegal skills install --describe | grep -F '"description": "Actúa en el ámbito global: $HOME/.agents/skills y los hosts de HOME."'
go run ./cmd/kitlegal boe articulo BOE-A-2015-10565 a21 --describe | grep -cF '"description": "Emite el sobre en JSON'
```

Esperado: las dos primeras imprimen su línea y salen con 0; la tercera imprime `0` (las globales no ganan descripción)
y sale con 1. Sin los dos argumentos, `boe articulo --describe` sale con 2: el análisis va antes que la descripción
(contracts/esquemas.md §1).

## 2. Los controles de `make ci` del hito

```sh
go test -count=1 -run '^TestAyudaComoDescripcion$' ./internal/cli/
go test -count=1 -run '^(TestAyudasDeLosArgumentos|TestEsquemasPublicados|TestHerramientasDelServidor)$' ./internal/app/
go test -count=1 -run '^(TestReglaTestigo|TestVariantesDelLaboratorio|TestEvalsDelLaboratorio|TestLaboratorioFueraDelProducto|TestTablaDelLaboratorio|TestArnesDelLaboratorio|TestServidorDelLaboratorio)$' ./internal/evals/
make schema-check
make skills-check
```

Esperado: todos en verde. `make schema-check` y `make skills-check` no escriben nada; `skills/` sale sin cambios
(`git status --short skills/` vacío). El guion e2e del hito, ya activado en `internal/app/testdata/script/`:

```sh
go test -count=1 -run '^TestEntregaDelHito$/^h39-describe$' ./internal/app/
```

## 3. La tabla de unas sesiones fijas (US3; FR-040 a FR-043)

```sh
go run ./cmd/laboratorio tabla internal/evals/testdata/laboratorio/sondeo | diff - internal/evals/testdata/laboratorio/tabla.md
make laboratorio-tabla DIRECTORIO=internal/evals/testdata/laboratorio/sondeo
```

Esperado: `diff` no imprime nada y sale con 0; `make` imprime la misma tabla. Con un directorio sin ningún sondeo
(`go run ./cmd/laboratorio tabla internal/evals/testdata/laboratorio/respuestas`), sale con 1 y lo dice en la salida
de error.

## 4. El laboratorio fuera del producto (US2; FR-010)

```sh
go test -count=1 -run '^TestLaboratorioFueraDelProducto$' -v ./internal/evals/
git ls-files laboratorio | grep -c '^laboratorio/'
```

Esperado: el test en verde (lo empotrado, lo que instala `skills install --dir` en un directorio vacío y la definición
del job, sin ningún fichero del laboratorio); la segunda orden cuenta los ficheros versionados de `laboratorio/`.

## 5. Los rechazos del arnés, sin abrir ninguna sesión (persona; FR-030, FR-031)

Con una credencial cualquiera en el entorno, para que el guion del llavero no pida la suya: los argumentos se
comprueban antes y no se usa.

```sh
CLAUDE_CODE_OAUTH_TOKEN=no-se-usa make evals-laboratorio VARIANTES=instrucciones MODOS=orden SESIONES=40 \
  MODELO=claude-sonnet-5-5 CLAUDE_CODE=2.1.317 CONCURRENCIA=5 DIRECTORIO="$TMPDIR/h39-rechazo"
CLAUDE_CODE_OAUTH_TOKEN=no-se-usa make evals-laboratorio VARIANTES=control MODOS=orden SESIONES=6 \
  MODELO=sonnet CLAUDE_CODE=2.1.317 CONCURRENCIA=5 DIRECTORIO="$TMPDIR/h39-rechazo"
test ! -e "$TMPDIR/h39-rechazo" && echo "sin directorio"
```

Esperado: las dos salen con un código distinto de 0; la primera dice `VARIANTES y MODOS: instrucciones no se mide en
modo orden`; la segunda, el error de `SESIONES` y el de `MODELO`; la tercera imprime `sin directorio`.

## 6. El sondeo de la aceptación (persona, antes de fusionar; FR-050, FR-051)

Con la credencial del llavero y la versión de Claude Code que dé `claude --version`:

```sh
make evals-laboratorio VARIANTES=control,entrada,referencia-secundaria,referencia-paso,referencia-en-el-punto,descripcion-sin-herramientas \
  MODOS=orden,herramienta SESIONES=40 MODELO=claude-sonnet-5-5 CLAUDE_CODE=<versión> CONCURRENCIA=5 \
  DIRECTORIO="$HOME/kitlegal-sondeos/h39/a"
make evals-laboratorio VARIANTES=instrucciones,parametro MODOS=herramienta SESIONES=40 MODELO=claude-sonnet-5-5 \
  CLAUDE_CODE=<versión> CONCURRENCIA=5 DIRECTORIO="$HOME/kitlegal-sondeos/h39/b"
make laboratorio-tabla DIRECTORIO="$HOME/kitlegal-sondeos/h39"
```

Esperado: 480 y 80 sesiones; la tabla con 14 filas, con el modelo y la versión en cada una. La persona la copia a
`evidencias/` y escribe el ADR que precisa el 0046 con sus cifras; los dos van a la propuesta de cambio antes de
fusionarla (FR-052).
