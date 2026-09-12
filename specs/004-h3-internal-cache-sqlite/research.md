# Research: H3 · `internal/cache`: SQLite con TTL y `--offline`

**Modo**: desatendido. Cada decisión de abajo se tomó con el «Criterio de decisión autónoma» de
`.specify/memory/constitution.md` (siempre la mejor solución; lo no especificado no se implementa; elegir
con criterio y dejar rastro) y registra la alternativa rechazada y por qué. Ninguna afecta a alcance,
frontera humana, privacidad, términos de uso, reglas de anomalías ni a una decisión cerrada de
`CLAUDE.md`, de modo que ninguna dispara el criterio 4 (escalar).

**Verificación.** Toda afirmación sobre el comportamiento de una herramienta o dependencia externa se ha
comprobado en local contra su código, su documentación o **una sonda ejecutada** contra ella, y cada
decisión cita dónde. La dependencia nueva del hito no estaba en la caché de módulos y se descargó con
`go mod download modernc.org/sqlite@v1.58.0` sin tocar `go.mod` ni `go.sum`; su comportamiento se
comprobó con cuatro sondas (`go run`) en un módulo desechable del scratchpad de la sesión, fuera del
repositorio; al corregir el plan se añadieron dos más, también en el scratchpad: la sonda 5 (qué error
devuelve `os.Stat` cuando un componente de la ruta es un fichero) y la sonda 6 (qué módulos incorpora
`go mod tidy` al importar el driver). Los módulos de las herramientas de control ya estaban (los trae
`tools/golangci-lint/go.mod`). Lo que no se ha podido verificar en local está en
[D18 · Supuestos no verificados](#d18--supuestos-no-verificados), declarado como supuesto y no como
hecho. Rutas de verificación usadas en este documento:

| Qué | Dónde se comprobó |
|---|---|
| Toolchain de Go | `go version` → `go1.27.1 darwin/arm64`; `go.mod` declara `go 1.27.0` + `toolchain go1.27.1` |
| Biblioteca estándar | `go doc <símbolo>` sobre ese toolchain y su fuente en `$(go env GOROOT)/src`: `database/sql/sql.go` 931-936 (`DB.Close` es idempotente: «Make DB.Close idempotent», devuelve `nil` la segunda vez) y 1026-1034 (`SetMaxOpenConns`, «0 = sin límite»); `go doc os.UserHomeDir` («en Unix, incluido macOS, devuelve `$HOME`»; error si la variable no está); `go doc embed` (`//go:embed` sobre `embed.FS` con patrones `path.Match`); `go doc testing.T.TempDir` y `testing/testing.go` 1616-1617 (la limpieza es `RemoveAll` y **falla el test** si no puede borrar: un directorio dejado en `0500` debe restaurarse antes); `go help build` 162-166 (`-tags`); `go help testflag` 130-134 (`-run`) y 228-230 (todo flag de test se reconoce con prefijo `test.`, obligatorio al invocar el binario de test directamente); `syscall/syscall_unix.go` 126-127 (`Errno.Is`: **solo** `ENOENT` es `fs.ErrNotExist`; `ENOTDIR`, que es lo que devuelve `Stat` cuando un componente de la ruta es un fichero, **no** lo es) y `GOOS=windows go doc syscall.ENOTDIR` (`ENOTDIR Errno = ERROR_PATH_NOT_FOUND`: la constante existe en las tres plataformas del binario y en Windows es el mismo error que `Errno.Is` ya trata como `fs.ErrNotExist`) |
| `modernc.org/sqlite` v1.58.0 | `go list -m -versions modernc.org/sqlite` (última: v1.58.0); `go mod download -json` (dir `/Users/jorge/go/pkg/mod/modernc.org/sqlite@v1.58.0`); su `go.mod` (`go 1.25.0`; **declara** como requisitos `modernc.org/libc v1.75.6`, `modernc.org/mathutil`, `modernc.org/fileutil`, `golang.org/x/sys v0.47.0`, `github.com/google/pprof` e indirectos `go-humanize`, `uuid`, `go-isatty`, `go-strftime`, `bigfft`, `modernc.org/memory`; lo que de ahí **llega al grafo de un módulo que lo importa** lo mide la sonda 6, no esta lista); `doc.go` (SQLite **3.53.4** en darwin/arm64 y linux/amd64; «use en go.mod la misma versión exacta de `modernc.org/libc`»; DSN por `sql.Open("sqlite", dsnURI)`); `sqlite.go` 285-468 (`applyQueryParams`: los `_pragma=` se ejecutan **por conexión** al abrirla, `busy_timeout` primero y el resto en orden alfabético; `_txlock=immediate` fija el modo de `BeginTx`; `_journal_mode`, `_synchronous`, `_query_only` y `_time_*` como atajos); `go doc modernc.org/sqlite Error` (`Code() int`, `Error() string`); `modernc.org/sqlite/lib` exporta `SQLITE_CANTOPEN=14`, `SQLITE_READONLY=8`, `SQLITE_READONLY_DIRECTORY=1544`, `SQLITE_NOTADB=26`, `SQLITE_BUSY=5` (sonda 2) |
| Sonda 1 (comportamiento de SQLite por el driver) | Módulo desechable `sonda` en el scratchpad, `go run` con go1.27.1. Resultados: `PRAGMA journal_mode`→`wal`, `synchronous`→`2` (FULL), `busy_timeout`→`5000`; DDL dentro de una transacción se deshace con `Rollback` (0 tablas después); tablas `STRICT` y `INSERT … ON CONFLICT(clave) DO UPDATE` aceptados; un BLOB de cero bytes se lee como `[]byte(nil)` (longitud 0) y la ausencia es `sql.ErrNoRows`; un lector `mode=ro` **ve** una fila confirmada por otra conexión abierta y aún sin checkpoint, `query_only`→`1`, y rechaza `INSERT` y DDL con «attempt to write a readonly database (8)»; `cache.db` queda **byte a byte igual** tras el lector; un segundo escritor con `busy_timeout` espera a que el primero confirme y termina sin error; un lector no ve una fila sin confirmar; al cerrar el último cliente desaparecen `-wal` y `-shm`; `mode=ro` sobre un fichero inexistente falla con «unable to open database file (14)» **sin crear nada**; un fichero que no es una base falla con «file is not a database (26)» en `ro` y en normal, y queda intacto; un fichero de 0 bytes se abre en `ro` como base sin tablas; un contexto cancelado corta la consulta con `context.Canceled`; el error del driver es `*sqlite.Error` con `Code()`; SQLite crea `cache.db` con `0644` si no existe, y con el fichero pre-creado a `0600` los auxiliares `-wal`/`-shm` heredan `0600` |
| Sonda 2 y 3 (directorio no escribible) | Mismo módulo. Con el directorio a `0500` y **sin** `-wal`/`-shm`, `mode=ro` falla con `SQLITE_READONLY_DIRECTORY` (1544) en la primera consulta —también la que lee `sqlite_master`— y `mode=ro&immutable=1` **lee** sin crear nada; con `-shm` presente y `-wal` ausente, lo mismo (1544); con `-wal` presente y `-shm` ausente, `mode=ro` falla con `SQLITE_CANTOPEN` (14); con `-wal` **y** `-shm` presentes (escritor abierto), `mode=ro` lee y ve lo confirmado en el WAL; un cliente normal sobre `0500` falla con 1544; un lector `ro` en un directorio **escribible** crea `-wal` y `-shm` y no los borra al cerrar; `os.Stat` bajo un directorio `0000` devuelve `os.ErrPermission` y bajo un directorio inexistente `os.ErrNotExist`; `os.MkdirAll` sobre una ruta que es un fichero falla con «not a directory»; con `busy_timeout(0)` un segundo escritor recibe `SQLITE_BUSY` (5) de inmediato |
| Sonda 4 (migraciones) | `ExecContext` ejecuta **varias sentencias** de una sola cadena (un fichero `.sql` entero) dentro de una transacción; `SELECT COALESCE(MAX(version), 0)` sobre una tabla inexistente falla con «no such table»; la existencia se consulta en `sqlite_master`; dos `BeginTx` con `_txlock=immediate` sobre la misma base: la segunda espera (o falla con `SQLITE_BUSY` si no hay espera) |
| Sonda 5 (`os.Stat` con un componente de la ruta que es un fichero) | Programa `go run` en el scratchpad, go1.27.1 darwin/arm64. Con `<tmp>/fichero` un fichero regular: `os.Stat("<tmp>/fichero/sub/cache.db")` y `os.Stat("<tmp>/fichero/cache.db")` fallan con «not a directory», `errors.Is(err, fs.ErrNotExist)` → **false**, `errors.Is(err, syscall.ENOTDIR)` → **true**, `errors.Is(err, fs.ErrPermission)` → false; `os.Stat("<tmp>/noexiste/cache.db")` y `os.Stat("<tmp>/noexiste")` fallan con «no such file or directory», `fs.ErrNotExist` → true, `ENOTDIR` → false; `os.MkdirAll("<tmp>/fichero/sub", 0o700)` falla con «mkdir <tmp>/fichero: not a directory», `ENOTDIR` → true. Conclusión: la regla «inexistente» de D3 necesita **las dos** comprobaciones para que el padre-fichero en solo lectura sea «sin base» (4) y no «inesperado» (1) |
| Sonda 6 (`go mod tidy` al importar el driver) | Módulo desechable `sonda` en el scratchpad con `go 1.27.0`, `require modernc.org/sqlite v1.58.0` y un `main.go` que importa `modernc.org/sqlite` y `modernc.org/sqlite/lib`; `GOFLAGS=-mod=mod GOPROXY=off go mod tidy` (sin red, todo desde la caché de módulos). El `go.mod` resultante lleva como indirectos exactamente `github.com/dustin/go-humanize v1.0.1`, `github.com/google/uuid v1.6.0`, `github.com/mattn/go-isatty v0.0.24`, `github.com/ncruces/go-strftime v1.0.0`, `github.com/remyoudompheng/bigfft v0.0.0-20230129092748-24d4a6f8daec`, `golang.org/x/sys v0.47.0`, `modernc.org/libc v1.75.6`, `modernc.org/mathutil v1.7.1`, `modernc.org/memory v1.12.1`. **Ni `modernc.org/fileutil` ni `github.com/google/pprof` entran en el grafo**, aunque el `go.mod` del driver los declare: ningún paquete alcanzable desde `modernc.org/sqlite` los importa y la poda del grafo de módulos los deja fuera. Por eso ninguna lista de módulos «que H4 deberá añadir» se escribe a mano (D13, S1) |
| `gosec` v2.28.0 (`tools/golangci-lint/go.mod` 168) | `rules/subproc.go` 62-94 (`func (r *subprocess) Match` entero, releído línea a línea al corregir el plan) y `resolve.go` 84-108 (G204: se marca todo argumento no resoluble a constante; `os.Args[0]` es un `IndexExpr` cuyo `X` es un selector → **no resoluble** → marcado; **excepción**: el bloque `if i == 0 { … }` de las líneas 72-81, encabezado por el comentario literal «Special case: struct fields OR function parameters/receivers used as executable name (i==0) -> skip», **omite** el nombre del ejecutable cuando es un campo de struct (`v.IsField()`) o un **parámetro o receptor** de la función (`obj.Pos() < bodyStart`); un `BasicLit` siempre resuelve. Si al implementar la versión de `gosec` cambiara, la cita que vale es el comentario, no el número de línea); `rules/fileperms.go` 79-112 (G301: `Mkdir`/`MkdirAll` con modo ⊄ `0750`; G302: `OpenFile`/`Chmod` con modo ⊄ `0600`; G306: `WriteFile` ⊄ `0600`); `rules/readfile.go` 143-246 (G304: `os.Open`/`OpenFile`/`ReadFile`/`Create` con ruta variable, **salvo** si la ruta es el resultado directo de `filepath.Clean`, `filepath.Rel` o `filepath.EvalSymlinks` o una variable asignada desde ellos; `os.Stat` y `os.MkdirAll` no están en la lista) |
| `sqlclosecheck` v0.6.0 | `pkg/analyzer/analyzer.go` 29: «Checks that sql.Rows, sql.Stmt, sqlx.NamedStmt, pgx.Query are closed» |
| `rowserrcheck` (golangci fork, `go.mod` 96) | `passes/rowserr/rowserr.go` 25 y 107: exige que se compruebe `Rows.Err` tras iterar un `*sql.Rows` («rows.Err must be checked») |
| `golangci-lint` v2.13.2 | `pkg/config/run.go` 22 (`run.build-tags`, lista) y `pkg/lint/package.go` 216-225 (se aplican como `-tags` al cargar los paquetes: el análisis alcanza los ficheros etiquetados) |
| `depguard` v2.2.1 | `settings.go` 224-248 (`strInPrefixList`: un paquete denegado casa por **prefijo**, salvo si termina en `$`; `modernc.org/sqlite` cubre también `modernc.org/sqlite/lib`, y `net/http` de la lista `red` cubre `net/http/httptest` en todo fichero fuera de `internal/httpx/**`, tests incluidos porque `run.tests: true`) |
| `paralleltest` v1.0.15 | `paralleltest.go` 87 y 150 (un test que llama a `t.Setenv` no puede ser paralelo y el linter lo reconoce) |
| `misspell` v0.8.0 | `words.go`: `directorios` (línea 6948, glosado como «directors»), `transaccion` («transaction») y `configuracion` («configuration») **están** en el diccionario como faltas de ortografía en inglés, y por eso los marca; `directorio`, `migracion`, `expiracion`, `vigencia`, `esquema`, `solo`, `lectura`, `contenido`, `entradas`, `clave`, `reloj`, `registrador`, `opcion` no están. `version` aparece dos veces como parte de otras entradas; H1 y H2 ya lo usan sin incidencia |
| Kernel de H1/H2 | `internal/core/schema/{contexto,error,sobre}.go` (`Contexto.Offline`, `Clase`, `ConClase`, `Resultado`), `internal/cli/errors.go` (`Clasificar`: sentinelas, después `errors.As` sobre `ConClase`; `CodigoSalida`), `internal/app/main.go` 305 (`context.WithTimeout` con `--timeout`) y 386-397 (`conPlazoAgotado`: un contexto vencido se envuelve en `cli.ErrFuenteNoDisponible`), `go doc ./internal/app` (`Main`, `Registro`, `Verbo`, `Argumentos`), `go doc ./internal/httpx` (`New`, `Replay(dir, opciones...)`, `Pedir(ctx, ejecucion, Peticion)`, `Respuesta.Cuerpo`), `internal/httpx/adaptador_test.go` (el patrón del adaptador de prueba y de `invocarAlKernel`), `internal/arch_test.go` (R3 sin `duenoObligatorio` con su razón; `TestElBinarioNoEnlazaHTTPX`; `modulosDelBinario`), `internal/httpx/grabar.go` 224-229 y 265-268 (H2 pasó G304 con `filepath.Clean`) |
| `Makefile`, CI, `.golangci.yml` | `Makefile` (`test`, `test-integration` = `go test -race -tags=integration ./...`, `ci` **no** incluye `test-integration`); `specs/001-h0-esqueleto-del-repo/contracts/make-targets.md` 22 (la receta de `test-integration` es contrato de H0); `.github/workflows/ci.yml` (solo `make ci` + Codecov); `.golangci.yml` (lista `sql` de `depguard` con `!**/internal/cache/**`; `run.tests: true`; sin `build-tags`); duración de la suite unitaria hoy: `go test -count=1 ./...` ≈ 22 s, dominada por `internal/httpx` (14 s) |
| Skills `golang-*` | `.agents/skills/golang-{how-to,database,testing,project-layout,design-patterns}/SKILL.md` (D17) |

---

## D1 · El puerto vive en `internal/core`, paquete `core`, como `core.Cache` con dos métodos

**Decisión.** Se crea el paquete `core` en `internal/core/` (hoy ese directorio solo tiene el
subpaquete `schema`) con dos ficheros: `doc.go` (comentario de paquete: «los puertos del dominio») y
`cache.go`:

```go
package core

type Cache interface {
    Get(ctx context.Context, clave string) (contenido []byte, presente bool, err error)
    Put(ctx context.Context, clave string, contenido []byte, vigencia time.Duration) error
}
```

Es la firma que fijó la clarificación Q3 (FR-001, FR-013): idioma «coma ok», ausencia = `presente=false,
err=nil`, fallo = `err != nil` con `schema.ConClase`. `Cache`, `Get` y `Put` son los nombres literales
del hito («Interfaz `Cache` definida en `core`», «`cache.Get/Put(key, ttl)`»); `clave`, `contenido` y
`vigencia` siguen la convención de identificadores en español del proyecto. Solo importa `context` y
`time`: R1 se cumple sin tocar la lista de `depguard`, cuyo patrón `**/internal/core/**` ya cubre el
paquete, igual que `compruebaDominioPuro` (`paquetesBajo` incluye el propio prefijo). **`Close` no forma
parte del puerto**: FR-004 dice que cierra «quien lo construyó», y un adaptador de fuente que reciba un
`core.Cache` no debe poder cerrar una caché que no es suya (interfaz mínima, `golang-structs-interfaces`).
El adaptador declara `var _ core.Cache = (*cache.Cliente)(nil)`.

**Alternativas.** *Ponerlo en `internal/core/schema`*: `schema` es el contrato del sobre y de `--describe`
(tipos JSON); un puerto de almacenamiento no pertenece a ese contrato y lo engordaría con algo que nunca
se serializa. *Un subpaquete `internal/core/puerto`*: nombre genérico que no dice qué contiene y que
obligaría a escribir `puerto.Cache` donde el roadmap y el hito dicen `core`; además, cuando lleguen
`Source`, `Fetcher` y `GraphStore` (H4, H17) el sitio natural es el mismo paquete raíz del dominio, que el
diagrama de `docs/ROADMAP.md` §2 describe como «define los PUERTOS». *Un método `Close` en el puerto*:
rechazado por lo dicho.

---

## D2 · Superficie de `internal/cache`: un tipo, un constructor con contexto, cuatro opciones, tres métodos

**Decisión.** El paquete exporta exactamente:

```go
const VariableDirectorio = "KITLEGAL_CACHE_DIR"

type Cliente struct{ /* privado */ }
func New(ctx context.Context, opciones ...Opcion) (*Cliente, error)

type Opcion func(*configuracion) error
func ConDirectorio(dir string) Opcion            // FR-023: precedencia máxima
func SoloLectura() Opcion                        // FR-015: el modo lo pone quien construye
func ConReloj(ahora func() time.Time) Opcion     // FR-009
func ConRegistrador(registrador *slog.Logger) Opcion

func (c *Cliente) Get(ctx context.Context, clave string) ([]byte, bool, error)
func (c *Cliente) Put(ctx context.Context, clave string, contenido []byte, vigencia time.Duration) error
func (c *Cliente) Close() error

type Error struct{ … }  // D9
```

- **`New` recibe `context.Context`** porque construir la caché **es** una operación con entrada y salida
  (crear el directorio, abrir la base, migrar): FR-003 exige contexto en toda operación y prohíbe crear
  uno de fondo. H2 no lo necesitaba en `New` porque no hacía nada hasta `Pedir`.
- **Un solo tipo `Cliente`** con y sin `SoloLectura()`: FR-017 prohíbe partir el puerto o el cliente en
  dos tipos; el rechazo de `Put` en solo lectura es un fallo en ejecución con clase «inesperado», medido
  por SC-011. No se añade ningún «refuerzo por construcción»: sería un segundo tipo o una interfaz
  parcial, y FR-017 lo permite solo como añadido que aquí no aporta nada que SC-011 no mida.
- **Opciones funcionales que devuelven error** (`golang-design-patterns`, patrón de H2): la validación
  ocurre al construir. `ConDirectorio("")`, `ConReloj(nil)` y `ConRegistrador(nil)` son «argumentos»
  (código 2). Se aplican en orden y la última repetida gana.
- **`ConReloj` es exportado.** FR-009 pide inyectar el reloj «al construir el cliente»; hacerlo por opción
  exportada permite que el adaptador de prueba de este hito (paquete externo `cache_test`, FR-045) siembre
  una entrada caducada para el escenario `--offline` con entrada expirada (US3 escenario 2) sin esperar
  tiempo real, y que H4 pruebe su adaptador sin abrir el paquete. Por omisión, `time.Now`.
- **`ConRegistrador`**: el mismo registrador que el kernel entrega al applet, como en H2 (D15 de H2);
  sin la opción los eventos se descartan (`slog.DiscardHandler`). Eventos a nivel `debug`: apertura
  (ruta, modo, versión de esquema), migración aplicada, acierto, ausencia, entrada expirada, escritura.
  No hay contadores ni métricas (spec, *Fuera de alcance*).
- **`Close`**: cierra el `*sql.DB`; la segunda llamada devuelve `nil` (la de `database/sql` ya es
  idempotente y el cliente además recuerda que está cerrado); `Get`/`Put` después de cerrar fallan con
  clase «inesperado» sin tocar el disco (FR-004, caso límite «cierre del cliente»).
- **Sin nombre de fichero exportado**: `cache.db` es una constante privada; el contrato lo fija y los
  tests externos usan el literal. Nada más se exporta: ni el `*sql.DB`, ni SQL, ni la ruta efectiva
  (FR-005). Un test de superficie (`TestSuperficieExportada`, patrón de H2) recorre las declaraciones
  exportadas y falla si aparece un tipo de `database/sql` o de `modernc.org/sqlite`.

**Alternativas.** *`New(opciones...)` sin contexto y abrir perezosamente en la primera operación*: haría
que el error de ruta inservible (código 2) o de esquema (código 1) apareciera en la primera lectura y no
al construir, mezclando dos clases de fallo en un mismo punto, y obligaría a `Get` a sincronizar una
apertura diferida. *Dos tipos (`Cliente` y `Lector`)*: prohibido por FR-017. *Reloj solo inyectable
desde el propio paquete (como H2)*: dejaría al adaptador de prueba sin forma de sembrar una entrada
caducada salvo con una vigencia de un nanosegundo y una espera real, que es justo lo que SC-002
prohíbe. *Exponer `Ruta()`*: nada del spec lo pide; los tests conocen el directorio porque lo declaran.

---

## D3 · Ruta efectiva: opción, variable, omisión; qué es inservible y en qué modo

**Decisión.** `New` resuelve la ruta **una sola vez** (FR-023) en este orden:

1. `ConDirectorio(dir)` si se declaró;
2. si no, `os.LookupEnv("KITLEGAL_CACHE_DIR")`: **presente** → ese valor (aunque esté vacío, que es
   inservible); ausente → omisión;
3. por omisión, `filepath.Join(home, ".cache", "kitlegal")` con `home` de `os.UserHomeDir()`
   (FR-019, al pie de la letra: `$HOME` en Unix y macOS, `%USERPROFILE%` en Windows).

La base es siempre `<directorio>/cache.db` (FR-020). La ruta pasa por `filepath.Clean` antes de
cualquier `os.OpenFile`, que es lo que G304 de `gosec` reconoce como saneado (verificado en
`readfile.go`; H2 hizo lo mismo en `grabar.go`).

Validación, cualificada por el modo (FR-022, FR-015). Tras resolver la ruta, `New` hace **un**
`os.Stat(dir)` sobre el directorio efectivo y clasifica el resultado con la **regla «inexistente»**, que
es la misma que D5 aplica después a `cache.db` y la única que decide qué fallo de `Stat` es una ausencia
y cuál un error:

- **existe** (`err == nil`): si `IsDir()` → se sigue; si no → «argumentos» (2) en cualquier modo (la
  ruta «existe y no es un directorio» de FR-022).
- **inexistente**: `errors.Is(err, fs.ErrNotExist) || errors.Is(err, syscall.ENOTDIR)`. Hacen falta
  **las dos**: `Stat("<fichero>/sub")` —un componente de la ruta, el padre, es un fichero— devuelve
  `ENOTDIR` («not a directory»), y en Go 1.27 `syscall.Errno.Is` solo reconoce `ENOENT` como
  `fs.ErrNotExist` (`syscall_unix.go` 126-127; sonda 5). Un directorio con un padre que es un fichero
  **no existe como directorio**, que es exactamente lo que FR-015 manda tratar «como un `cache.db`
  inexistente» en solo lectura y lo que FR-022 llama «no existe y no se puede crear» en modo normal. En
  Windows `syscall.ENOTDIR` es `ERROR_PATH_NOT_FOUND`, que `Errno.Is` ya trata como `fs.ErrNotExist`:
  la regla vale igual en las tres plataformas del binario y compila en todas.
- **otro** (`fs.ErrPermission` porque el padre deniega el acceso, error de E/S…): no se sabe si el
  directorio está. La clase depende del modo (tabla).

Ningún resultado de `Stat` queda fuera de estas tres ramas, así que ninguna forma de fallo queda sin
clase (FR-033). El predicado vive en `ruta.go` (`esInexistente(err) bool`), único sitio donde se importa
`syscall`, y solo por la constante `ENOTDIR`.

| Situación (resultado de `Stat(dir)`) | Modo normal | Modo de solo lectura |
|---|---|---|
| Valor vacío (opción o variable presente y vacía); no se llega a `Stat` | «argumentos» (2), nombrando el origen | «argumentos» (2) |
| **existe** y no es un directorio | «argumentos» (2) | «argumentos» (2) |
| **inexistente** (`ErrNotExist`: no hay nada en la ruta; o `ENOTDIR`: el padre es un fichero) | se crea con `MkdirAll(dir, 0o700)`; si falla —con `ENOTDIR` cuando el padre es un fichero (sonda 5), o por permisos— → «argumentos» (2) | se trata como `cache.db` inexistente: cliente **sin base** (D5), sin `MkdirAll` ni ningún otro acceso al disco; toda lectura es ausencia → 4. El caso «padre que es un fichero» lo miden `TestDirectorioNoCreable/solo-lectura` y `TestSoloLecturaSinBaseNoCreaNada`; el de la ruta sin nada, además, `offline-directorio-inexistente` por el kernel |
| **existe**, es un directorio y no se puede escribir | «argumentos» (2): falla `OpenFile(cache.db, O_RDWR|O_CREATE, 0o600)` o, si `cache.db` ya existía, la migración/primera operación con `SQLITE_READONLY_DIRECTORY` | se lee (D5) |
| **existe**, es un directorio y deniega el acceso (`0000`): `Stat(dir)` no falla, falla el siguiente paso | «argumentos» (2): `MkdirAll` no falla pero `OpenFile` devuelve `ErrPermission` | `os.Stat(dir/cache.db)` devuelve `ErrPermission` (sonda 2 E) → «inesperado» (1), nunca ausencia |
| **otro** error de `Stat(dir)` (el padre deniega el acceso, E/S) | «argumentos» (2), sin intentar `MkdirAll` (fallaría igual) | «inesperado» (1), nunca ausencia |
| `os.UserHomeDir` falla (sin `HOME`) | «argumentos» (2), nombrando `HOME` y `KITLEGAL_CACHE_DIR` | igual |

La última fila es una situación que FR-033 no enumera; se declara en *Complexity Tracking* del plan con
el mismo criterio que H2 D19 (ningún fallo sin clase; es corregible por quien invoca, declarando la
variable). El mensaje de cualquier fallo de ruta nombra **de dónde vino** la ruta («opción
`ConDirectorio`», «variable `KITLEGAL_CACHE_DIR`», «ruta por omisión») y **cuál era** (FR-022, FR-035).

**Alternativas.** *Tratar la variable vacía como ausente*: es la «caída silenciosa a la ruta por omisión»
que FR-022 prohíbe. *`os.UserCacheDir`*: decisión cerrada de `CLAUDE.md`, excluida por el spec.
*Comprobar la escribibilidad del directorio con `unix.Access`*: no es portable y añadiría un import
directo de `golang.org/x/sys` fuera de la lista de la constitución §V; la detección por el intento real
de crear el fichero (y por el código 1544 del driver cuando el fichero ya existía) cubre los dos casos
sin dependencia nueva.

---

## D4 · Apertura normal: fichero pre-creado a `0600`, DSN con `PRAGMA` seguros, una conexión

**Decisión.** En modo normal, `New`:

1. `os.MkdirAll(dir, 0o700)` (G301: ⊆ `0750`);
2. `os.OpenFile(filepath.Clean(ruta), os.O_RDWR|os.O_CREATE, 0o600)` y cierre inmediato (G302: ⊆ `0600`;
   G304: ruta saneada). Un fichero de 0 bytes es una base SQLite válida (sonda 1 G) y con él pre-creado
   los auxiliares `-wal`/`-shm` heredan `0600` (sonda 1 I). Sin este paso SQLite crea `cache.db` a
   `0644` (sonda 1 I), que incumple FR-021. Sin `O_EXCL`: dos procesos que arrancan a la vez comparten
   el fichero y se serializan en la migración (D7);
3. `sql.Open("sqlite", "file:<ruta>?_pragma=busy_timeout(100)&_pragma=journal_mode(WAL)&_pragma=synchronous(FULL)&_txlock=immediate")`,
   con `<ruta>` codificada para el camino del URI (`%` → `%25`, `?` → `%3F`, `#` → `%23`;
   `rutaParaURI`): el driver entrega el DSN `file:` entero al motor con `SQLITE_OPEN_URI`, que
   decodifica `%HH` y corta el camino en `?` y `#` (revisión final, motivo [b][e]);
4. `db.SetMaxOpenConns(1)`;
5. `migrar(ctx)` (D7).

`PRAGMA` y por qué (FR-030, FR-031):

| `PRAGMA` | Valor | Motivo |
|---|---|---|
| `journal_mode` | `WAL` | Literal del hito; es lo que permite leer mientras otra invocación escribe (sonda 1 D y E). Es persistente en el fichero; se declara en cada apertura normal por idempotencia |
| `synchronous` | `FULL` (2) | Es el valor por omisión de SQLite (sonda 1 A lo confirma) y el más conservador: cada confirmación se sincroniza con el disco. FR-031 prohíbe sacrificar integridad por velocidad; la caché escribe unos KB por consulta y el coste de la red domina |
| `busy_timeout` | `100` ms por intento; 5 s en total | Espera ante bloqueo (FR-031): una invocación simultánea espera hasta cinco segundos en vez de fallar de inmediato (sonda 1 E frente a sonda 2 G). Cinco segundos cubre cualquier transacción de la caché —una migración o un `Put`— con margen; un valor mayor solo alargaría un fallo real. **Revisión final**: la espera de `busy_timeout` vive dentro del motor y no mira el contexto —`sqlite3_interrupt`, que es lo que el driver hace al terminar el contexto, no la corta—, así que la planificación inicial (`5000` en el DSN) dejaba un `Put` o un `New` con plazo de 300 ms cinco segundos esperando, contra FR-003. Ahora el motor espera un **tramo** de 100 ms y el cliente repite la sentencia mientras reciba `SQLITE_BUSY`, mirando el contexto entre tramos, hasta un presupuesto de 5 s medido con el reloj real (`espera.go`, contrato de apertura §4). Se reintentan solo sentencias en autocommit y el comienzo de una transacción, donde `SQLITE_BUSY` garantiza que no se hizo nada |
| `_txlock` | `immediate` | Toda transacción explícita (solo las de migración, D7) toma el bloqueo de escritura al empezar, no al primer `INSERT`, de modo que dos procesos que migran a la vez se serializan sin `SQLITE_BUSY` por escalada de bloqueo (sonda 4) |
| `query_only` | solo en solo lectura (D5) | Refuerzo del modo `ro` dentro de la propia conexión |

Los `_pragma` del DSN se aplican en **cada** conexión nueva (`applyQueryParams`), así que no dependen
del pool. **Una sola conexión por cliente** (`SetMaxOpenConns(1)`): un cliente de la caché atiende una
invocación secuencial; con una conexión no hay contención dentro del proceso, `Get` y `Put` no
necesitan transacción explícita (el upsert es una sentencia, autocommit) y dos clientes en el mismo
proceso son dos pools independientes que se comportan como dos procesos (FR-032). La `golang-database`
exige configurar el pool: es esta línea.

**Alternativas.** *`synchronous=NORMAL`*: es la recomendación habitual con WAL y no corrompe la base,
pero sí puede perder las últimas confirmaciones ante un corte de energía; un juez podría leerlo, con
razón, como «sacrificar por velocidad» algo que FR-031 protege, y la ganancia aquí es inapreciable.
*`foreign_keys`, `temp_store`, `mmap_size`, `cache_size`*: nada del esquema los necesita (YAGNI).
*Pool por omisión (ilimitado)*: dos goroutines del mismo cliente competirían por el bloqueo de escritura
dentro del proceso sin necesidad. *Dejar que SQLite cree el fichero*: `0644`, contra FR-021.

---

## D5 · Apertura en solo lectura: `Stat` primero, `mode=ro`, y `immutable=1` solo cuando el directorio no admite escritura

**Decisión.** Con `SoloLectura()`, `New` no crea el directorio ni el fichero y no migra (FR-015):

1. Si la validación de D3 ya clasificó el directorio como **inexistente** (`Stat(dir)` con
   `fs.ErrNotExist` o `syscall.ENOTDIR`), el cliente queda **sin base** sin tocar el disco. Si el
   directorio existe, `os.Stat(<dir>/cache.db)` y la **misma regla «inexistente» de D3**:
   `errors.Is(err, fs.ErrNotExist) || errors.Is(err, syscall.ENOTDIR)` → el cliente queda **sin base**:
   no se abre nada, el directorio queda intacto también en `-wal`/`-shm`, y toda lectura es ausencia,
   que FR-016 convierte en «fuente no disponible» (código 4). Cualquier otro error de `Stat`
   (`ErrPermission` en un directorio `0000`, sonda 2 E; E/S) → «inesperado» (1), nunca una ausencia
   falsa. Las dos ramas cubren todo fallo posible de `Stat`: lo que no es inexistente es inesperado.
2. Si existe: `sql.Open("sqlite", "file:<ruta>?mode=ro&_pragma=busy_timeout(100)&_pragma=query_only(1)")`,
   con `<ruta>` codificada para el camino del URI como en D4 (`rutaParaURI`) y el mismo tramo de 100 ms
   por intento de la espera por tramos (D4, revisión final),
   `SetMaxOpenConns(1)`, y la **comprobación de esquema** (D7) en la propia construcción, que es la
   primera consulta real. Sin `immutable`: así la lectura ve lo que otra invocación ya confirmó en el WAL
   y respeta sus bloqueos (sonda 1 D, sonda 2 D; clarificación Q2, FR-015, FR-032).
3. Si esa primera consulta falla con `*sqlite.Error` de código `SQLITE_READONLY_DIRECTORY` (1544) **o**
   `SQLITE_CANTOPEN` (14), SQLite no pudo abrir lo que necesita para leer, y el código no dice si el
   culpable es el fichero o el directorio. **Revisión final** (motivo [e][b]): antes de nada se comprueba
   con `os.Open` + cierre —una lectura que no escribe nada— si `cache.db` se deja leer; si no, el fallo
   es «inesperado» (1) nombrando el fichero y el acceso denegado (fila 12), sin culpar al directorio, y
   va antes de mirar `-wal` porque con `-wal` presente el mensaje de la fila 13 también lo culparía. Si
   el fichero se deja leer, el directorio no permite crear `-shm`: **cuál de los dos códigos llega depende
   de qué auxiliares existan** —sin `-wal`, o con `-shm` y sin `-wal`, es 1544 (sonda 3 B y D); con `-wal`
   presente y `-shm` ausente es 14 (sonda 3 E)—, así que la condición que se comprueba es la misma para
   los dos y es la presencia de `cache.db-wal`, no el código. Ningún escritor puede estar trabajando ahí,
   porque escribir exige crear `-wal` y `-shm` en ese mismo directorio (sonda 3 G). Entonces:
   - si **no existe** `cache.db-wal`, se cierra y se reabre con `mode=ro&immutable=1`, que lee sin crear
     nada (sonda 3 C). Todo lo confirmado está en `cache.db`, porque no hay WAL, así que no se pierde
     ninguna entrada. Si la reapertura o su primera consulta vuelven a fallar (el fallo era otro, p. ej.
     permisos del propio fichero), «inesperado» (1) nombrando la ruta, nunca una ausencia falsa;
   - si **existe** `cache.db-wal` (y falta `-shm`, que es el único caso que llega aquí con `-wal`, y el que
     llega con el código 14), el fallo es «inesperado» (1): SQLite no puede leer un WAL sin su memoria
     compartida en un directorio donde no puede crearla (sonda 3 E), y abrirlo como inmutable ignoraría lo
     que hay en el WAL. El mensaje nombra la ruta y los dos auxiliares (fila 13 del contrato de errores).
4. Cualquier otro error de la primera consulta → «inesperado» (1) sin degradar a ausencia
   (`SQLITE_NOTADB` (26) para un fichero que no es una base, sonda 1 G).

Un directorio **no escribible** se lee, por tanto, «con normalidad» (FR-015) por una de dos vías, las
dos verificadas: si otra invocación está escribiendo ahí (el caso «el disco era escribible cuando
empezó»), `-wal` y `-shm` existen y `mode=ro` lee lo confirmado (sonda 2 D); si nadie escribe y no hay
WAL, `immutable=1` es correcto porque el fichero no puede cambiar. `immutable=1` **no se usa nunca** en un
directorio escribible: un lector inmutable no toma bloqueos y un escritor que apareciera después podría
hacer checkpoint bajo sus pies (por eso la clarificación Q2 lo descartó como modo general).

En modo de solo lectura, un `cache.db` de 0 bytes (dejado por una ejecución normal interrumpida antes de
migrar) se abre como base sin tablas (sonda 1 G): versión 0, sin migración, toda lectura ausencia.

**Alternativas.** *`immutable=1` siempre*: descartado en Q2 y peligroso con un escritor concurrente.
*Detectar la escribibilidad del directorio antes de abrir (crear y borrar un fichero de prueba)*:
escribiría en un modo que promete no escribir nada. *Tratar 1544 como «inesperado» sin más*: dejaría el
caso más común del modo —el contenedor de solo lectura de US3— en código 1, contra el DEBE de FR-015.
*Disparar la comprobación de `-wal` solo ante 1544*: el caso con `-wal` presente llega con el código 14
(sonda 3 E), de modo que caería en «cualquier otro error» y su mensaje no nombraría los dos auxiliares
que la fila 13 del contrato de errores exige y `TestIntegracionWALSinMemoriaCompartida` mide.

---

## D6 · Esquema v1: dos tablas `STRICT`, expiración como `INTEGER` de nanosegundos Unix, sin índice

**Decisión.** Una única migración, `internal/cache/migraciones/0001_entradas.sql`:

```sql
CREATE TABLE schema_version (
    version     INTEGER PRIMARY KEY,
    aplicada_en TEXT    NOT NULL
) STRICT;

CREATE TABLE entradas (
    clave     TEXT    PRIMARY KEY NOT NULL,
    contenido BLOB    NOT NULL,
    expira_en INTEGER NOT NULL
) STRICT;
```

- `schema_version` guarda **una fila por migración aplicada** (versión e instante en RFC 3339 UTC); la
  versión actual es `MAX(version)`. Deja historia sin coste y no exige una fila «singleton» que actualizar.
- `entradas`: la clave opaca es la clave primaria (FR-011, FR-007: un upsert sustituye contenido y
  vigencia sin duplicar); el contenido es un BLOB opaco (FR-006, FR-012); `expira_en` es el instante de
  expiración como `time.Time.UnixNano()` (UTC implícito), y la comparación de vigencia se hace **en Go**
  —`ahora.Before(time.Unix(0, expiraEn))`— y no en SQL, para que el borde de FR-008 (vigente mientras el
  reloj es *anterior*; en el instante exacto, ausencia) no dependa de la resolución ni del formato de
  fecha del motor, y para que el reloj inyectado sea la única fuente de «ahora».
- **`expira_en` tiene un intervalo representable, y `Put` satura a él** (revisión final): `UnixNano`
  solo está definido entre el 21 de septiembre de 1677 y el 11 de abril de 2262 (`go doc
  time.Time.UnixNano`), y FR-010 admite cualquier vigencia mayor que cero. Con 240 años desde 2026, o con
  `time.Duration(math.MaxInt64)`, la conversión daba un número cualquiera y `Get` leía la entrada recién
  guardada como ausencia —bajo `--offline`, 4 en vez de 0—. `instanteDeExpiracion` (`entradas.go`)
  calcula `reloj().Add(vigencia)` como `time.Time`, que no desborda, y guarda `math.MaxInt64` cuando la
  suma no es anterior al último instante representable y `math.MinInt64` cuando no es posterior al primero
  (solo alcanzable con un reloj inyectado). Para todo reloj dentro del intervalo la comparación decide
  igual que con el instante exacto: vigente hasta el último instante lo que caducaría después, caducado lo
  que caducó antes del primero. No es una situación de fallo y no toca la tabla cerrada de D9
  (`TestExpiracionFueraDelIntervaloRepresentable`).
- `STRICT` (SQLite ≥ 3.37; aquí 3.53.4, sonda 1 C) hace que un valor de tipo equivocado sea un error y
  no una conversión silenciosa.
- **Sin índice sobre `expira_en`**: H3 no borra ni desaloja lo caducado (*Fuera de alcance*); `Get` va
  por clave primaria. Se añadirá con la primera política de desalojo, si llega.
- Lo caducado **no se borra al leer** (decisión de diseño que el spec deja abierta): `Get` en solo
  lectura no puede escribir, y borrar en modo normal convertiría una lectura en una escritura con bloqueo
  sin beneficio observable. El upsert de la clave sustituye la entrada caducada (FR-007, US2 escenario 4).

**Alternativas.** *`expira_en` como texto RFC 3339 o como `time.Time` del driver*: legible desde un
cliente `sqlite3`, pero introduce formato, zona horaria y redondeo de subsegundos en la comparación del
borde. *Guardar `creada_en` o metadatos de la respuesta*: contenido opaco por FR-006; nada los usa.
*Una tabla por fuente*: la clave ya distingue lo que deba distinguir (FR-011) y el esquema de claves
llega con H4. *Rechazar con «argumentos» (2) la vigencia cuya expiración no cabe en `expira_en`*: FR-010
la declara válida, sería una fila nueva en la tabla cerrada de D9 y un adaptador que declarase «para
siempre» fallaría, cuando la saturación da el mismo resultado observable para todo reloj del intervalo.
*Guardar `expira_en` como texto o con más bits*: cambia el esquema por un caso que la saturación resuelve
sin tocarlo.

---

## D7 · Migraciones embebidas, `schema_version`, atomicidad y carrera entre dos procesos

**Decisión.** `//go:embed migraciones/*.sql` en un `embed.FS`. Cada fichero se llama `NNNN_<nombre>.sql`;
la versión conocida por el binario es el número de ficheros, y un test (`TestMigracionesEmbebidasBienFormadas`)
exige numeración contigua desde `0001`, sin huecos ni repetidos. `migrar(ctx)`:

1. `versionRegistrada`: `SELECT count(*) FROM sqlite_master WHERE type='table' AND name='schema_version'`
   (sonda 4); si no existe → 0; si existe → `SELECT COALESCE(MAX(version), 0) FROM schema_version`. Un
   fichero que no es una base falla aquí con `SQLITE_NOTADB` → «inesperado» (1) sin tocarlo (FR-028).
2. Si la registrada es **mayor** que la conocida → «inesperado» (1) con «esperaba v1, encontró vN»; el
   fichero no se modifica (FR-027).
3. Por cada versión pendiente, en orden: `BeginTx` (inmediata por `_txlock`), **releer la versión dentro
   de la transacción** (si otro proceso la aplicó mientras se esperaba el bloqueo, `Rollback` y seguir),
   `ExecContext` con el fichero entero (varias sentencias en una cadena: sonda 4), `INSERT INTO
   schema_version`, `Commit`. DDL y `INSERT` van en la misma transacción y SQLite los deshace juntos
   (sonda 1 B): una interrupción deja la versión anterior, nunca a medias (FR-026).
4. Volver a abrir una base ya migrada no aplica nada y no falla (FR-025): la versión registrada es igual
   a la conocida y el bucle está vacío.

En **solo lectura** no se aplica ninguna migración; la versión registrada solo se **comprueba** (D5):
0 → base sin esquema, toda lectura ausencia; igual a la conocida → se lee; **cualquier otra** (mayor,
o menor cuando existan más versiones) → «inesperado» (1) con esperada y encontrada, porque leer un
esquema distinto del que el binario conoce sería adivinar. Con una sola versión conocida, la rama
«menor» es inalcanzable en H3 y queda escrita para que la regla sea completa.

La atomicidad se **prueba** sin matar procesos: el test pre-crea en la base una tabla `entradas` con otra
forma, de modo que la migración 1 falla en su segunda sentencia después de haber creado
`schema_version`; tras el fallo (código 1) la base no tiene `schema_version` y conserva la tabla ajena:
o entra todo o no entra nada.

**Alternativas.** *`golang-migrate` u otra herramienta* (lo que sugiere la skill `golang-database`):
dependencia fuera de la lista de la constitución §V para dos sentencias; el hito dice literalmente
`embed` + tabla `schema_version`. *`PRAGMA user_version`* en lugar de la tabla: no deja historia y el
hito nombra la tabla. *Dividir el fichero por `;`*: innecesario, el driver ejecuta la cadena entera.
*Aplicar sin transacción explícita*: rompe FR-026.

---

## D8 · `Get` y `Put`: validación, upsert, ausencia, expiración y solo lectura

**Decisión.**

`Put(ctx, clave, contenido, vigencia)`:
1. `clave == ""` → «argumentos» (2); `vigencia <= 0` → «argumentos» (2); sin tocar el disco (FR-010,
   FR-011). El orden de comprobación es clave, vigencia, modo.
2. En solo lectura (o sin base) → «inesperado» (1) nombrando la clave, sin escribir (FR-017); la
   comprobación es anterior a cualquier acceso, así que también vale para el cliente sin base.
3. Cerrado → «inesperado» (1).
4. `expiraEn := instanteDeExpiracion(reloj(), vigencia)`, que es `reloj().Add(vigencia).UnixNano()`
   saturado a `math.MaxInt64` cuando la suma no es anterior al último instante representable (11 de abril
   de 2262) y a `math.MinInt64` cuando no es posterior al primero (D6; revisión final); una sentencia:
   `INSERT INTO entradas(clave, contenido, expira_en) VALUES (?, ?, ?) ON CONFLICT(clave) DO UPDATE SET contenido = excluded.contenido, expira_en = excluded.expira_en`
   (sonda 1 C), en autocommit. Un `contenido` nulo se guarda como BLOB vacío (`NOT NULL`).
5. Un error del driver → «inesperado» (1) con la causa envuelta; `ctx.Err()` → «fuente no disponible»
   (4), ver D9.

`Get(ctx, clave)`:
1. `clave == ""` → «argumentos» (2).
2. Cerrado → «inesperado» (1).
3. Sin base (solo lectura con `cache.db` inexistente) o esquema en versión 0 → ausencia.
4. `QueryRowContext("SELECT contenido, expira_en FROM entradas WHERE clave = ?").Scan(&contenido, &expiraEn)`:
   `sql.ErrNoRows` → ausencia; otro error → «inesperado» (1) / contexto → 4.
5. Si `!reloj().Before(time.Unix(0, expiraEn))` → ausencia (FR-008, FR-018): la fila se deja (D6).
6. Presente: se devuelve `contenido` con `presente = true`; un BLOB de cero bytes llega del driver como
   `nil` (sonda 1 C) y se normaliza a `[]byte{}` para que «presente y vacío» no se confunda con nada.
7. **En solo lectura, la ausencia (pasos 3, 4 y 5) no se devuelve como ausencia sino como fallo de la
   clase «fuente no disponible»** (código 4), nombrando la clave y que la invocación es de solo lectura
   (FR-016): es «el modo … convierte la ausencia en un fallo» de Key Entities, y deja al adaptador sin
   nada que decidir bajo `--offline`. Fuera de ese modo la ausencia es `false, nil` (FR-013, FR-014).

`QueryRowContext(...).Scan` no deja `*sql.Rows` abiertos ni exige `Rows.Err`: `sqlclosecheck` y
`rowserrcheck` no tienen nada que objetar y tampoco se les esquiva; ninguna consulta del paquete devuelve
más de una fila.

**Alternativas.** *Devolver la ausencia en solo lectura como `false, nil` y que el adaptador decida el
código 4*: dos lecturas posibles de FR-016 y una decisión repetida en cada adaptador; el spec la sitúa en
el modo. *Comparar la vigencia en SQL (`WHERE expira_en > ?`)*: el borde quedaría escrito dos veces (SQL
y Go) y la resolución del instante dependería del motor. *Borrar al leer lo caducado*: D6.

---

## D9 · Errores tipados: `cache.Error` declara su clase; tabla cerrada de quince situaciones

**Decisión.** Un único tipo de error, con el patrón de `httpx.Error` (H2 D4):

```go
type Error struct {
    Operacion string // «construir», «migrar», «leer», «escribir», «cerrar»
    Ruta      string // fichero o directorio implicado, si lo hay
    Origen    string // de dónde salió la ruta: opción, variable o por omisión; vacío si no aplica
    Clave     string // clave implicada, si la hay
    Causa     error  // error de origen (driver, sistema de ficheros, contexto); Unwrap lo expone
    // clase, privada
}
func (e *Error) Error() string        // en español; nombra ruta, origen, clave y versiones según toque (FR-035); nunca vacío ni panic
func (e *Error) Unwrap() error
func (e *Error) Clase() schema.Clase  // implementa schema.ConClase
```

`cli.Clasificar` ya reconoce `schema.ConClase` con `errors.As` tras los cinco sentinelas (H2), así que
`internal/cache` no importa `internal/cli`: la dependencia sigue yendo al dominio. Envolver con `%w`
conserva la clase. La tabla cerrada situación → clase → código está en
[`contracts/errores-y-codigos.md`](./contracts/errores-y-codigos.md) §3: las **ocho** situaciones de
SC-011 —que corresponden a las **nueve** filas marcadas ★ (1, 2, 4, 5, 6, 8, 9, 10, 11), porque un bullet
de SC-011 puede agrupar varias filas, como explica el §3 del contrato— y las **seis** restantes (3, 7, 12,
13, 14, 15) que el diseño hace inevitables y que FR-033 exige clasificar: opción inválida, ruta por
omisión indeterminable, acceso denegado, WAL sin `-shm` en directorio no escribible, la fila que reúne el
fallo de E/S o del driver con la operación tras el cierre, y el contexto cancelado o vencido. Ninguna
produce 3, 5 ni 6.

**Contexto cancelado o vencido → «fuente no disponible» (4).** No está en la lista de FR-033. Se elige
la clase que H1 define como «la fuente que no responde, y también el plazo agotado» y que H2 asignó a la
misma situación (contrato de errores de H2, fila 2), de modo que el kernel y los tres paquetes dicen lo
mismo; además `conPlazoAgotado` del kernel ya envuelve en `cli.ErrFuenteNoDisponible` todo error cuando
el contexto de `--timeout` vence (`main.go` 386-397), así que cualquier otra clase aquí sería ignorada en
el camino real. Alternativa rechazada: «inesperado» (1), que haría que un `--timeout` corto en H4
terminara con dos códigos distintos según dónde venciera el contexto.

**Alternativas.** *Sentinelas propios (`ErrAusente`)*: prohibido por FR-013 (la ausencia no es error) y
un error sin datos no puede nombrar ruta ni clave. *Importar `internal/cli` para envolver sus
sentinelas*: adaptador dependiendo del kernel (H2 D4).

---

## D10 · Concurrencia: dos clientes son dos pools; sin goroutines propias; `-race`

**Decisión.** El cliente guarda un `*sql.DB` con una conexión, un `sync.Mutex` que protege el estado
«cerrado» y nada más. `Get` y `Put` son sentencias sueltas en autocommit; solo la migración abre
transacciones (inmediatas). Dos clientes sobre la misma base —en el mismo proceso o en dos— son dos
conexiones que SQLite arbitra con WAL y `busy_timeout` (sonda 1 D y E): el lector ve entradas completas o
ninguna, y toda entrada confirmada antes de leer (FR-032, SC-007). El escenario de **dos procesos** se
acredita relanzando el binario de test (clarificación Q5), ver D11. La suite pasa con `-race`, que
`make test` y `make test-integration` ya llevan. El paquete no arranca ninguna goroutine.

**Alternativas.** *Un mutex global del paquete para serializar clientes del mismo proceso*: no cubre dos
procesos y esconde lo que SQLite ya resuelve; además impediría medir la pareja «lector de solo lectura
frente a escritor» en el mismo proceso. *Reintentar ante `SQLITE_BUSY`*: la planificación inicial lo
descartó porque `busy_timeout` ya esperaba dentro del motor; la **revisión final** lo adoptó, porque esa
espera no mira el contexto y dejaba fuera de FR-003 a toda operación bloqueada (motivo [e][f]): el
motor espera un tramo de 100 ms y el cliente reintenta mirando el contexto hasta 5 s en total
(`espera.go`, D4). Sigue sin haber goroutines propias —la pausa entre tramos es un temporizador del
runtime— y un bloqueo que dura más que la espera sigue siendo un fallo real, declarado como «inesperado»
(1) con un mensaje que dice que la base está bloqueada y nombra la espera, nunca como fichero
inutilizable.

---

## D11 · Tests: todos sobre una base real en `t.TempDir()`; la etiqueta `integration` marca lo que depende del entorno; CI los ejecuta

**Decisión.**

- **Todos** los tests de `internal/cache` trabajan sobre un `cache.db` real en `t.TempDir()`: el motor
  va dentro del proceso, abrir una base cuesta milisegundos y no existe un doble fiel de SQLite que valga
  la pena mantener (`golang-database`: SQL explícito, sin *mocks* del motor). Ningún test toca
  `~/.cache/kitlegal`: los que ejercitan la ruta por omisión redefinen `HOME` (y `USERPROFILE`) con
  `t.Setenv` hacia un directorio temporal; los que ejercitan la variable usan `t.Setenv` y por tanto no
  son paralelos (`paralleltest` lo reconoce). Todo lo demás lleva `t.Parallel()`.
- La etiqueta **`//go:build integration`** (`integracion_test.go`, paquete `cache_test`) marca los tests
  que **dependen del entorno o lanzan procesos**: dos procesos (relanzando el binario de test), permisos
  del sistema de ficheros (directorio `0500` y `0000`, WAL sin `-shm`). Es la línea que la skill
  `golang-testing` traza («unit tests fast; build tags for integration tests») y la que el roadmap §3
  describe («integración: `//go:build integration`, SQLite en `t.TempDir()`»). Los tests unitarios
  cubren el resto y son los que alimentan `coverage.out`.
- **CI ejecuta los etiquetados** (FR-041): el objetivo `ci` del `Makefile` gana `test-integration`
  inmediatamente después de `test` y antes de `vuln`, de modo que la línea queda
  `ci: fmt-check lint test test-integration vuln schema-check secrets mod-verify mod-tidy-check`, que es
  la que el escenario 11 del quickstart compara literalmente. La receta de `test-integration`
  (`go test -race -tags=integration ./...`) es
  contrato de H0 y **no cambia**; el coste es repetir la suite unitaria una vez más (≈ 22 s hoy, medido),
  aceptable frente a la alternativa de un segundo objetivo o de acotar la receta a un paquete.
- **El análisis estático alcanza los ficheros etiquetados** (FR-041, SC-009): `.golangci.yml` gana
  `run.build-tags: [integration]` (clave verificada en `pkg/config/run.go` 22; se aplica como `-tags` al
  cargar los paquetes, `package.go` 216-225). Sin ello `sqlclosecheck`, `gofumpt` y `govet` no verían
  `integracion_test.go`. El quickstart lo demuestra introduciendo un `*sql.Rows` sin cerrar en ese fichero
  sobre una copia desechable.
- **Dos procesos sin `//nolint`**: G204 marca `exec.CommandContext(ctx, os.Args[0], …)` porque
  `os.Args[0]` no resuelve a constante, pero **omite el nombre del ejecutable cuando es un parámetro de
  la función** (`subproc.go` 72-81, verificado). El lanzador es por tanto
  `lanzarProcesoAuxiliar(ctx, ejecutable string, dir string, rol string) *exec.Cmd`, que recibe
  `os.Args[0]` desde el test y construye `exec.CommandContext(ctx, ejecutable, "-test.run=^TestProcesoAuxiliar$")`
  —el único argumento es un literal, que G204 resuelve—; el papel (`escritor` o `lector-solo-lectura`) y
  el directorio viajan en `cmd.Env` (`KITLEGAL_CACHE_DIR` incluido, que ejercita FR-020 desde otro
  proceso). `TestProcesoAuxiliar` retorna sin hacer nada cuando la variable del papel no está (patrón de
  `TestHelperProcess` de la biblioteca estándar; no es un `t.Skip`), y con ella actúa y termina con 0 o
  con un mensaje en la salida estándar que el padre incorpora al fallo. Bajo `-race` el hijo hereda la
  instrumentación porque es el mismo binario.
  **Resultado (implementación, T010 intento 1)**: la versión fijada de `gosec` (v2.28.0) trae además
  G702, un análisis de propagación que trata `os.Args` y `os.Getenv` como datos no confiables y
  `exec.CommandContext` como destino, y marcó la llamada aunque el ejecutable llegara como parámetro. La
  ruta del binario sale por eso de `os.Executable()` —la que da el sistema operativo, que quien lanza el
  proceso no fija, y la que usa `internal/testenv.Executable` de la biblioteca estándar para relanzar sus
  binarios de test—; el lanzador la sigue recibiendo como parámetro, con el literal como único argumento
  y sin ninguna supresión. El hijo deja además en un fichero de informe cuántas entradas confirmó o
  encontró, porque un hijo que no llegara a ejecutar `TestProcesoAuxiliar` también terminaría con 0.
- **Permisos**: los tests que ponen un directorio a `0500`/`0000` registran con `t.Cleanup` la
  restauración a `0700` **después** de `t.TempDir()`, porque la limpieza de `TempDir` es `RemoveAll` y
  falla el test si no puede borrar (`testing.go` 1616). Comprueban antes que el sistema de ficheros hace
  valer los permisos (crean un directorio `0000` y exigen que `os.Stat` dentro devuelva `ErrPermission`);
  si no los hace valer (usuario `root`), el desenlace **depende de dónde corra el proceso**: con la
  variable de entorno `CI` no vacía —que GitHub Actions exporta como `CI=true`— es `t.Fatalf`
  **nombrando la causa**, y fuera de ella `t.Skip` nombrando la misma causa. Es la precondición de entorno
  que SC-004 declara («se miden donde el sistema de ficheros los hace valer»), no un test desactivado: en
  cualquier entorno donde la aserción tenga sentido se ejecuta, y donde el hito **exige** que la tenga —la
  integración continua, supuesto S3— un ejecutor privilegiado pone el trabajo en rojo en vez de pasar de
  largo. Es también lo que hace comprobable S3: la receta de `test-integration` es contrato de H0, corre
  sin `-v` e imprime solo `ok <paquete>`, de modo que un `t.Skip` no dejaría ningún rastro que leer en el
  registro de CI.
- **Atomicidad y versión mayor** se prueban con la base manipulada desde el propio test (paquete
  `cache`, caja blanca, que puede importar `database/sql` porque R3 se lo permite a `internal/cache/**`,
  tests incluidos): pre-crear una tabla ajena; escribir `INSERT INTO schema_version VALUES (99, …)`.
- **Un fichero de test por fichero de código** y en el mismo orden (`golang-testing`), más
  `concurrencia_test.go`, `superficie_test.go`, `adaptador_test.go` e `integracion_test.go`; `doc.go` y
  `internal/core/{doc,cache}.go` no tienen test porque no contienen sentencias (una interfaz y comentarios).

**Alternativas.** *Etiquetar como `integration` todo lo que abre SQLite*: dejaría `make test` casi sin
cobertura de `internal/cache` y el umbral global del 70 % en riesgo, y convertiría en «integración» tests
de milisegundos. *Ejecutar los etiquetados en un job aparte de CI en lugar de en `ci`*: rompería la
identidad «lo que pasa en local pasa en CI» de H0 (`ci.yml` solo llama a `make ci`). *Un `package main`
auxiliar para el segundo proceso*: descartado en Q5 (añade un binario y corre sin `-race`). *`go test
-run` desde el test para el hijo*: recompila y exige `go` en la máquina; el binario ya está en
`os.Args[0]`. *`//nolint:gosec` para G204 como en `arch_test.go` (H1)*: SC-008 prohíbe toda supresión
nueva y el patrón del parámetro lo hace innecesario.

---

## D12 · El adaptador de prueba: applet `prueba` con `consultar` y `guardar`, el kernel en proceso y la reproducción en las dos consultas

**Decisión.** `internal/cache/adaptador_test.go`, paquete `cache_test` (FR-045, Q4), copia el patrón de
`internal/httpx/adaptador_test.go`:

- `adaptadorDePrueba{directorio string, reloj func() time.Time, cliente *httpx.Cliente}` implementa
  `app.Applet` con nombre `prueba` y dos verbos:
  - `consultar <url>`: construye la caché con `ConDirectorio(directorio)` (si no está vacío; si lo está,
    la ruta sale de `KITLEGAL_CACHE_DIR`, que es lo que un subtest con `t.Setenv` ejercita),
    `ConRegistrador(registrador)`, `ConReloj(reloj)` si lo hay y **`SoloLectura()` si y solo si
    `ejecucion.Offline`** (FR-015, FR-047: el único sitio de H3 donde la bandera se convierte en modo);
    `Get(ctx, "prueba:"+url)`; si presente, devuelve `{cuerpo, origen: "cache"}`; si ausente,
    `Pedir(ctx, ejecucion, httpx.Peticion{Metodo: "GET", URL: url})` con el cliente que el test le dio,
    `Put` con vigencia de una hora, y devuelve `{cuerpo, origen: "fuente"}`. `defer` que cierra la caché
    y une su error al de retorno (`errcheck` exige mirar `Close`).
  - `guardar <clave> <contenido>`: construye la caché igual y llama a `Put`. Existe para acreditar por el
    kernel la escritura rechazada en solo lectura (FR-017, FR-047, SC-011) y para sembrar entradas en los
    escenarios de `--offline`.
- **Las dos consultas usan `httpx.Replay`**, ninguna `httpx.New` ni un servidor local. Razón mecánica,
  verificada: la lista `red` de `depguard` deniega `net/http` en todo fichero fuera de
  `internal/httpx/**` —tests incluidos (`run.tests: true`)— y casa por **prefijo** (`settings.go`
  224-248), así que también deniega `net/http/httptest`; un servidor local exige nombrar `net/http`
  (`http.Handler`, `http.ResponseWriter`), y el adaptador de H2 pudo hacerlo solo porque vive bajo
  `internal/httpx`. Relajar la lista para `internal/cache/*_test.go` sería una exclusión nueva (SC-008 y
  R2). Por eso:
  - la **primera** consulta se sirve desde un directorio de reproducción que contiene **una grabación
    escrita a mano** de `GET http://fuente.prueba/norma` (200, `text/plain`, cuerpo
    `<norma>contenido</norma>`), el único fixture del hito:
    `internal/cache/testdata/reproduccion/prueba/GET_http_fuente.prueba_norma.json`, con el nombre que
    dicta el contrato de grabación de H2 §2 y el formato de su §1 (`"formato": 1`). Que «la consulta
    llega a la fuente» (US1 escenario 1) se acredita porque el sobre trae `origen: fuente` y el cuerpo de
    la grabación, y porque la reproducción de H2 falla ante cualquier petición distinta de la grabada;
  - la **segunda** consulta usa `httpx.Replay(t.TempDir())` sobre un directorio **vacío** (estricto,
    FR-046): cualquier petición emitida termina con clase 1 nombrándola. Código 0, `origen: cache`, cuerpo
    idéntico byte a byte al de la primera (SC-001).
- `TestAdaptadorDePruebaConElKernel` invoca `app.Main` sobre un registro construido en el test
  (`invocarAlKernel`, como H2) con **nueve** subtests —seis de ellos, y ninguno más, con el prefijo
  `offline-` que filtra el escenario 4 del quickstart—: `primera-consulta` (reproducción con la grabación;
  `origen: fuente`; `cache.db` aparece; es también lo que acredita US3 escenario 5 y FR-014, porque parte
  de una caché vacía y la ausencia no altera ni el curso ni el código), `segunda-consulta-sin-red` (misma caché, reproducción estricta
  vacía; 0, `origen: cache`, mismo cuerpo), `sin-cache-la-reproduccion-falla` (caché vacía y reproducción
  vacía: código 1 y el mensaje nombra `GET http://fuente.prueba/norma`, que es lo que demuestra que el
  subtest anterior pasa *porque* no se emitió ninguna petición, US1 escenario 3), `offline-presente` (0,
  `origen: cache`, `cache.db` idéntico por SHA-256), `offline-ausente` (4, el mensaje nombra la clave),
  `offline-expirada` (sembrada con `ConReloj` en el pasado; 4), `offline-sin-base` (directorio vacío: 4,
  y sigue vacío), `offline-directorio-inexistente` (4, y sigue sin existir), `offline-guardar` (1, sin
  fichero nuevo). `TestAdaptadorConVariableDeEntorno` (`t.Setenv`, no paralelo): la base aparece bajo
  `KITLEGAL_CACHE_DIR` y no bajo `HOME`; variable vacía → 2 nombrándola; variable que apunta a un fichero
  → 2.
- El fixture es **material de test escrito a mano contra el host ficticio `fuente.prueba`**, no una
  grabación real: la tarea que lo crea lleva `[datos]` (`docs/WORKFLOW.md`: material nuevo bajo
  `internal/<pkg>/testdata/` exige etiqueta y guardián, sin pausa) y el spec lo prevé («el material de
  reproducción que este hito necesite se escribe a mano, como el de H2»). Es el único fichero de
  `internal/cache/testdata/` y no se crea `testdata/` en la raíz.
- Nada de esto se registra en ningún binario (FR-044): `TestElBinarioNoEnlazaLosEjemplos` (H1) y la
  ausencia de `adaptadorDePrueba` y de los literales `"consultar"`/`"guardar"` fuera de `_test.go` en
  `internal/app`, `cmd` e `internal/cache` (quickstart).

**Alternativas.** *Primera consulta con `httpx.New` contra un servidor `httptest` local, como H2*:
imposible sin importar `net/http` en `internal/cache`, que `depguard` deniega por R2; se detectó al
escribir el contrato de arquitectura. *Copiar en el test el fixture de H2
(`internal/httpx/testdata/reproduccion/prueba/…`)*: acopla los tests de un paquete al `testdata` de otro
y va contra la co-localización de fixtures (H2 D17). *Construir el JSON en el test a partir del contrato
de formato*: duplica en `internal/cache` la regla del nombre y del formato de H2. *Sin el verbo
`guardar`*: no habría camino por el kernel para FR-017, que FR-047 exige acreditar «por ese mismo
camino». *Un applet de ejemplo en `internal/app/ejemplo`*: se enlazaría en el binario de e2e (H2 D16).

---

## D13 · Controles mecánicos: qué se añade y qué se toca

**Decisión.**

- **`internal/arch_test.go`**: la subprueba R3 conserva `duenoObligatorio: false` **a propósito** (FR-037:
  exigir dueño requeriría que `store` y `graph` existieran) y su comentario pasa a decir que R3 tiene ya
  su primer dueño real (`internal/cache`) y por qué sigue sin exigirse. Se añade
  `TestElBinarioNoEnlazaCache` (temporal hasta H4, como `TestElBinarioNoEnlazaHTTPX`): el cierre de
  `go list -deps ./cmd/kitlegal` no contiene `internal/cache` ni ningún paquete bajo `modernc.org/`.
  `TestDependenciasDelBinario` **no cambia su lista** (FR-044, SC-012): el binario no enlaza la caché.
- **`.golangci.yml`**: `run.build-tags: [integration]` (D11) y comentarios de la lista `sql` («R3 tiene
  dueño desde H3»). Ninguna regla nueva, ninguna exclusión, ningún `//nolint` (SC-008). Contingencia de
  `misspell` (S2): igual que H2, con la tarea declarando el fichero.
- **`Makefile`**: `ci` gana `test-integration` entre `test` y `vuln` (D11). Ningún objetivo nuevo ni
  receta cambiada.
- **`go.mod` / `go.sum`**: `modernc.org/sqlite v1.58.0` y sus módulos indirectos (FR-043).
- **`docs/PENDIENTES.md`**: entrada «En H4» ampliada: retirar `TestElBinarioNoEnlazaCache` y ampliar
  `modulosDelBinario` con **los módulos que `go list -deps ./cmd/kitlegal` muestre al enlazar
  `internal/cache`**, justificados uno a uno por escrito (constitución §V). La nota **no lleva ninguna
  lista escrita a mano**: la sonda 6 demuestra que el `go.mod` del driver declara módulos
  (`modernc.org/fileutil`, `github.com/google/pprof`) que no llegan al grafo, y una lista copiada de ahí
  acabaría fijando en `docs/PENDIENTES.md` dependencias que el binario nunca enlaza. Lo que se puede
  anotar como orientación es la salida de la sonda 6 (`libc`, `mathutil`, `memory`, `golang.org/x/sys`,
  `go-humanize`, `uuid`, `go-isatty`, `go-strftime`, `bigfft`), marcada como **medida en un módulo de
  sonda, no en el binario**, y con la instrucción de sustituirla por la medida real en H4.
- **Fixtures**: uno, escrito a mano (D12): `internal/cache/testdata/reproduccion/prueba/GET_http_fuente.prueba_norma.json`,
  en una tarea `[datos]`. Ningún `testdata/` en la raíz; ninguna grabación real. **Tests de contrato**:
  los cuatro contratos de `contracts/` tienen comprobación mecánica (tabla del plan). **ADR**,
  **`docs/SOURCES.md`**, `schemas/`: sin cambio (spec, *Fuera de alcance*). **`CHANGELOG.md`**: una
  entrada bajo *Cambiado*, porque `make ci` gana `test-integration` (D11) y ese cambio es visible para
  quien ejecuta los controles, como H1 registró el de `make test-e2e`; `README.md` y `CONTRIBUTING.md`
  lo reflejan en sus tablas (revisión final).

---

## D14 · Registro de eventos: `slog` a nivel `debug` por el registrador recibido

**Decisión.** Igual que H2 D15: sin `ConRegistrador` se descarta todo (`slog.New(slog.DiscardHandler)`);
con él, eventos `debug` con atributos estables (`ruta`, `modo`, `version_esquema`, `clave`, `resultado`
∈ {`acierto`, `ausencia`, `expirada`}). Ningún mensaje de error va al registrador y a la vez al sobre:
el sobre lleva `Error.Error()`, el registro lleva la `Causa` técnica (FR-035).

---

## D15 · Nombres en español que `misspell` acepta

**Decisión.** Verificado en `words.go` (tabla de verificación): los identificadores y palabras sueltas
del hito son `directorio` (singular), `migracion`/`migraciones`, `expira`, `vigencia`, `esquema`,
`entradas`, `clave`, `contenido`, `reloj`, `registrador`, `opcion`, `soloLectura`. **No se usan** como
palabra suelta sin acento `directorios`, `transaccion` ni `configuracion`, que el diccionario marca:
en comentarios se escriben con acento (`transacción`, `configuración`, que no casan) y en identificadores
se usa `directorio`, `tx` y `ajustes`/`opciones`. Si aun así apareciera un falso positivo, rige la
contingencia S2.

---

## D16 · Dónde no hay nada que decidir (y por qué no se construye)

Lo que el spec deja en *Fuera de alcance* queda fuera del plan sin excepción: ningún adaptador real, ni
puerto `Source`/`Fetcher`, ni esquema de claves, ni revalidación condicional, ni `store`/`graph`, ni
verbos de mantenimiento, ni desalojo, ni cifrado, ni métricas, ni cambio en `internal/httpx`, ni
`os.UserCacheDir`, ni semántica de `--dry-run` en la caché, ni e2e nuevo, ni ADR, ni umbral propio de
cobertura; en `CHANGELOG.md`, solo la entrada que el cambio de `make ci` exige (D13, revisión final). Un test de superficie (`TestSuperficieExportada`) fija que el paquete no
exporta más de lo que D2 enumera.

---

## D17 · Aplicación de las skills `golang-*` instaladas

`golang-how-to` orquesta (se leyó su `SKILL.md`; el intento de invocarla como skill falló en este
entorno y se aplicó leyendo los ficheros de `.agents/skills/`):

| Skill | Qué fija en este plan |
|---|---|
| `golang-database` | SQL explícito y parametrizado (`?`), sin ORM; `*Context` en toda operación; `sql.ErrNoRows` con `errors.Is` para la ausencia (patrón «`exists bool`» de la skill, que es la firma de D1); `QueryRowContext` en vez de `Query` para una fila; pool configurado (D4); transacciones para la migración multi-sentencia (D7). Su recomendación de una herramienta externa de migraciones se rechaza por la constitución §V y el literal del hito (D7) |
| `golang-design-patterns` | Opciones funcionales que devuelven error (D2); sin estado global ni `init()`; el recurso (`*sql.DB`) tiene dueño y `Close` explícito |
| `golang-testing` | Tabla con subtests, `t.Parallel()` salvo con `t.Setenv`, `t.TempDir()`, etiqueta `integration` para lo dependiente del entorno (D11), `-race`, un fichero de test por fichero de código, caja negra para el adaptador (`cache_test`) |
| `golang-project-layout` | `internal/cache` con un fichero por responsabilidad; `migraciones/` junto al paquete; el puerto en `internal/core` (D1); ningún `pkg/` |
| `golang-error-handling` | Tipo de error con `Unwrap`, clasificación con `errors.As`, mensajes que nombran ruta, origen y clave (D9) |
| `golang-context` / `golang-concurrency` | Contexto obligatorio como primer parámetro, también en `New`; sin goroutines propias; mutex solo para el estado de cierre (D10) |
| `golang-lint` / `golang-security` | Sin `//nolint` (SC-008): G204 por parámetro, G301/G302 con `0700`/`0600`, G304 con `filepath.Clean` (D3, D4, D11) |
| `golang-naming` | Identificadores en español con la convención del lenguaje (`Cliente`, `ConDirectorio`, `SoloLectura`), tests sin guiones bajos |

---

## D18 · Supuestos no verificados

Todo lo anterior está comprobado en local. Quedan como **supuestos**, con su comprobación asignada a
`tasks.md`:

- **S1 · `go mod tidy` añade a `go.mod` exactamente `modernc.org/sqlite v1.58.0` y sus indirectos.** La
  sonda 6 (`go mod tidy` sin red sobre un módulo que importa `modernc.org/sqlite` y `modernc.org/sqlite/lib`)
  resolvió como indirectos `modernc.org/libc v1.75.6`, `modernc.org/mathutil v1.7.1`,
  `modernc.org/memory v1.12.1`, `golang.org/x/sys v0.47.0`, `github.com/dustin/go-humanize v1.0.1`,
  `github.com/google/uuid v1.6.0`, `github.com/mattn/go-isatty v0.0.24`, `github.com/ncruces/go-strftime v1.0.0`
  y `github.com/remyoudompheng/bigfft`; **no** incorporó `modernc.org/fileutil` ni `github.com/google/pprof`,
  que el `go.mod` del driver declara pero ningún paquete alcanzable importa. Sigue siendo un supuesto
  porque la sonda es un módulo que no tiene las demás dependencias de `kitlegal`: la selección mínima de
  versiones puede subir `golang.org/x/sys` (hoy v0.26.0 indirecto) y no se ha ejecutado `tidy` sobre el
  repositorio para no tocar `go.mod` durante la planificación. La lista de arriba es una **predicción
  medida**, no la lista final: la final es la que `go mod tidy` escriba, y ningún documento del repositorio
  la copia a mano (D13). Comprobación: `go mod tidy -diff` limpio y `TestDependenciasDelBinario` en verde sin
  tocar su lista en la primera tarea que añada la dependencia. `doc.go` del driver pide usar la misma
  versión de `modernc.org/libc` que su `go.mod`: la selección mínima la respeta mientras nada exija una
  mayor.
- **S2 · `misspell` no marca ningún identificador ni comentario nuevo.** Verificado para la lista de D15;
  cualquier otra palabra que aparezca al escribir se comprueba con `make lint` en la primera tarea que
  cree ficheros de `internal/cache`, y la contingencia es la de H2: entrada en `misspell.ignore-rules`
  declarada por la tarea, o reescritura del término, nunca `//nolint`.
  **Resultado (implementación, T008 intento 1)**: el supuesto se quedó corto en un término que ningún
  identificador nuevo trae y que dos contratos fijan: `reproduccion`, en el nombre del subtest
  `sin-cache-la-reproduccion-falla` (inventario de tests del plan, escenario 2 del quickstart) y en el
  tramo del directorio de grabaciones de H2 que el adaptador de prueba reproduce. `misspell` separa las
  palabras por cualquier carácter fuera de `[a-zA-Z0-9']`, así que el guion no la protege. Salida: la línea
  de T008 declara `.golangci.yml` acotado a esa palabra en `misspell.ignore-rules` y la tarea queda sin
  marcar para el intento siguiente. La entrada es durable: cada adaptador de fuente de H4 en adelante
  volverá a escribir ese directorio en sus tests.
- **S3 · El ejecutor de la integración continua no es `root`.** Los tests de permisos (D11) lo exigen
  para medir algo; si no se cumple, `TestIntegracionDirectorioNoEscribible`,
  `TestIntegracionDirectorioDenegado` y `TestIntegracionWALSinMemoriaCompartida` terminan en `t.Fatalf`
  nombrando la causa —allí no saltan— y hay que corregir el entorno, no el test.
  Comprobación: la tarea `[plataforma]` ve el trabajo `ci` de la propuesta de cambio **en verde con
  `test-integration` dentro** (`gh pr checks`, `gh run list`). No se busca ningún `PASS` ni `SKIP` en el
  registro: la receta corre sin `-v` y no los imprime; el rojo sería la única señal de que el supuesto no
  se cumple, y su ausencia es la comprobación.
- **S4 · Codecov**: sin componente nuevo (SC-014: rigen los umbrales generales); el estado global y el de
  `internal/core` siguen en verde. `internal/core` (paquete `core`) no tiene sentencias, así que no
  altera la cobertura del componente.
- **S5 · Tiempo de `make ci` con `test-integration`**: la suite unitaria dura ≈ 22 s en local y se
  repetirá una vez; se asume que `ci` de una PR limpia sigue por debajo de los 3 minutos de H0. Se mide en
  la tarea `[plataforma]`.
