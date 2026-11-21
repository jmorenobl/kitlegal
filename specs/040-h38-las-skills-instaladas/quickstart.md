# Quickstart: comprobar H38

Escenarios para validar lo entregado. Se ejecutan desde la raíz del repositorio, con Go 1.27 y el módulo descargado (sin
red). Ninguno escribe fuera de un directorio de `mktemp -d` (que vive en `$TMPDIR`), de los `t.TempDir()` de los tests
y de la caché de compilación de Go; ninguno toca el árbol de trabajo, el índice ni el historial de git, ni la cuenta de
quien los ejecuta (`HOME` y `KITLEGAL_CACHE_DIR` apuntan siempre dentro del temporal).

Describen el binario entregado: con el código de hoy fallan, porque el binario de hoy solo da la línea de H19, y en la
sesión del plan no se ejecutaron (research.md, «Lo medido en esta sesión»).

## 1 · Los unitarios

```sh
go test -race -count=1 -run 'TestReglaDeVersiones|TestReinstalacion' ./internal/core/instalacion/
go test -race -count=1 -run 'TestTomarCandado|TestSuperficieExportada' ./internal/cache/
go test -race -count=1 -run 'TestAvisoDelKernel|TestAvisoDeVersion' ./internal/app/
go test -count=1 -run 'TestArquitectura' ./internal/
```

Esperado: `ok` en los cuatro. `TestReinstalacionConcurrente` y `TestReinstalacionFalloAMitad` van en el primero
(research.md D7); `TestSuperficieExportada`, en el segundo, pasa con la entrada `"func TomarCandado"` en
`superficieDelContrato` (research.md D4); `TestArquitectura` vigila que `internal/app/aviso.go`, que ahora importa `internal/cache`, no alcanza
la red (SC-005).

## 2 · La aceptación y los e2e de H19 que cambian

Con la suite activada en `internal/app/testdata/script/` (lo hace el workflow tras el bucle de tareas):

```sh
go test -race -count=1 -run 'TestEntregaDelHito/^h38-' ./internal/app/
go test -race -count=1 -run 'TestEntregaDelHito/^(h19-skills-aviso|h19-skills-aviso-sin-aviso|h19-skills-conflictos-dentro|h21-mcp-protocolo)$' ./internal/app/
```

Esperado: `ok`; con `-v`, `PASS` en `h38-reinstala`, `h38-tocadas`, `h38-excluidas` y `h38-candado`, y en los cuatro
de H19 y H21. El arnés construye sus binarios en su propio temporal y lo borra.

## 3 · A mano, con dos binarios de verdad

```sh
t=$(mktemp -d)
go build -ldflags '-X main.version=v0.1.0' -o "$t/v1/kitlegal" ./cmd/kitlegal
go build -ldflags '-X main.version=v0.2.0' -o "$t/v2/kitlegal" ./cmd/kitlegal
mkdir -p "$t/home/.claude" "$t/trabajo"
cd "$t/trabajo"
HOME="$t/home" KITLEGAL_CACHE_DIR="$t/cache" "$t/v1/kitlegal" skills install -g >/dev/null
HOME="$t/home" KITLEGAL_CACHE_DIR="$t/cache" "$t/v2/kitlegal" territorio resolver Leganés --json --no-graph >/dev/null
```

Esperado, en la salida de error y como única línea:

```text
aviso: kitlegal v0.2.0 ha actualizado las skills globales, que eran de v0.1.0; se aplican en la próxima conversación
```

Y después:

```sh
grep -c 'v0\.1\.0' "$t/home/.agents/skills/kitlegal.json"     # 0
ls -A "$t/cache"                                              # nada: el candado ya no está
HOME="$t/home" KITLEGAL_CACHE_DIR="$t/cache" "$t/v2/kitlegal" territorio resolver Leganés --json --no-graph >/dev/null
                                                              # ninguna línea en la salida de error
HOME="$t/home" "$t/v2/kitlegal" skills doctor -g --json       # código 0 y ningún hallazgo en data
```

### Lo tocado no se pisa

```sh
mkdir -p "$t/home2"
HOME="$t/home2" KITLEGAL_CACHE_DIR="$t/cache2" "$t/v1/kitlegal" skills install -g >/dev/null
echo "una línea añadida a mano" >> "$t/home2/.agents/skills/normativa/SKILL.md"
HOME="$t/home2" KITLEGAL_CACHE_DIR="$t/cache2" "$t/v2/kitlegal" territorio resolver Leganés --json --no-graph >/dev/null
```

Esperado: la línea de hoy, `aviso: las skills instaladas son de kitlegal v0.1.0 y este binario es kitlegal v0.2.0;
ejecuta: kitlegal skills install -g`; `kitlegal.json` sigue en `v0.1.0` y `$t/cache2` no existe.

### `--dry-run` y el candado

```sh
mkdir -p "$t/home3"
HOME="$t/home3" KITLEGAL_CACHE_DIR="$t/cache3" "$t/v1/kitlegal" skills install -g >/dev/null
HOME="$t/home3" KITLEGAL_CACHE_DIR="$t/cache3" "$t/v2/kitlegal" territorio resolver Leganés --dry-run --no-graph
```

Esperado: en la salida de error, la descripción de `--dry-run` y la línea `aviso: kitlegal v0.2.0 se actualizarían las
skills globales, que eran de v0.1.0; se aplican en la próxima conversación`; nada cambia y `$t/cache3` no existe.

```sh
mkdir -p "$t/cache3" && : > "$t/cache3/skills-global.lock"
HOME="$t/home3" KITLEGAL_CACHE_DIR="$t/cache3" "$t/v2/kitlegal" territorio resolver Leganés --json --no-graph >/dev/null
touch -t 202001010000 "$t/cache3/skills-global.lock"
HOME="$t/home3" KITLEGAL_CACHE_DIR="$t/cache3" "$t/v2/kitlegal" territorio resolver Leganés --json --no-graph >/dev/null
```

Esperado: con el candado reciente, la línea de hoy y el candado sigue; con el huérfano, la línea de FR-006 y el
candado ya no está (`ls -A "$t/cache3"` vacío).

### Limpieza

```sh
cd - >/dev/null
rm -rf "$t"
```
