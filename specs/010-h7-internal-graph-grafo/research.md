# Research: H7 · `internal/graph`: grafo del mundo, operaciones en el `Resultado` y `graph check` mínimo

Fase 0 del plan, en modo desatendido. Tres partes: **V** (lo que se comprobó en local, con dónde), **D** (decisiones,
con la alternativa rechazada y por qué) y **S** (supuestos que no se pueden comprobar sin red, sin la plataforma o
sin el código que todavía no existe). Nada de lo que depende de un supuesto se afirma como hecho en el plan.

Criterio aplicado: «Criterio de decisión autónoma» de `.specify/memory/constitution.md` (la mejor solución; lo no
especificado no se implementa; ante alcance, privacidad o decisión cerrada, la lectura conservadora). Skills de Go
aplicadas: `golang-how-to` (orquesta), `golang-project-layout` (paquetes), `golang-database` (SQLite, transacciones,
contexto), `golang-testing` (tablas, `t.TempDir`, etiquetas), `golang-cli` (códigos, stdout/stderr),
`golang-error-handling` (errores con clase), `golang-design-patterns` (puertos, Null Object), `golang-lint`
(`depguard`, `dupl`, `misspell`), `golang-naming`.

## V · Verificaciones en local

| # | Afirmación | Dónde se comprobó |
|---|---|---|
| V1 | Toolchain `go1.27.1`; `go.mod` declara `go 1.27.0` + `toolchain go1.27.1`; `modernc.org/sqlite v1.59.0` | `go version`; `go.mod` líneas 3-16 |
| V2 | `time.Time.MarshalJSON` —lo que escribe `fecha_consulta` en el sobre— es `appendStrictRFC3339` → `appendFormatRFC3339(b, true)`, el mismo camino que `Format(time.RFC3339Nano)`; con desplazamiento 0 escribe `Z` y la fracción de segundo va sin ceros a la derecha | `$(go env GOROOT)/src/time/time.go:1591-1600`; `format_rfc3339.go:18-60`; `format.go:44` («RFC3339Nano removes trailing zeros»), `:120`, `:659-661` |
| V3 | `time.Parse` con un desplazamiento numérico que coincide con una zona de `Local` devuelve el instante **en `Local`**: sumarle una duración que cruce un cambio de horario cambiaría el desplazamiento al escribirlo. Para conservar el de `fecha_consulta` hace falta `time.FixedZone` con ese desplazamiento | `go doc time.Parse` (líneas 34-37 de su salida) |
| V4 | `os.UserHomeDir` devuelve error si `HOME` falta **o está vacía** («$HOME is not defined») | `$(go env GOROOT)/src/os/file.go:605-624` |
| V5 | `encoding/json/jsontext.Value.Canonicalize` implementa JCS (RFC 8785); `encoding/json/v2` ya se usa en el dominio | `go doc encoding/json/jsontext.Value.Canonicalize`; `internal/core/instalacion/manifiesto.go` (importa `encoding/json/v2`) |
| V6 | `regexp` (RE2): `\d` es `[0-9]` ASCII, pero la bandera `i` **no** es ASCII: pliega cada clase con `unicode.SimpleFold`, así que con `(?i)` la clase `[a-z]` casa también `ſ` (U+017F) y el signo Kelvin (U+212A), y `[^a-z0-9]` deja de admitirlos como límite (`ſ12345678Z` no casaría y `12345678` + U+212A sí). No hay aserciones de contorno, así que los límites de FR-025 se escriben como clases. Por eso la expresión va **sin** `(?i)` y con clases ASCII explícitas (D34) | `go doc regexp/syntax` («Perl character classes (all ASCII-only)», `i  case-insensitive`); `$(go env GOROOT)/src/regexp/syntax/parse.go:2020` (`appendFoldedRange`, que añade los plegados de cada rango); sonda V37 |
| V7 | `invopop/jsonschema v0.14.0` marca como obligatoria toda clave sin `omitempty`/`omitzero` | `$(go env GOMODCACHE)/github.com/invopop/jsonschema@v0.14.0/reflect.go:4`, `:104-107`, `:952` |
| V8 | `testscript` (go-internal v1.16.0): `exec … &` lanza en segundo plano y `wait` hace fallar el guion si una orden de fondo falló; `env K=` deja la variable **presente y vacía** (para quitarla hace falta `exec env -u K …`) | `…/go-internal@v1.16.0/testscript/doc.go:151-170`, `:240-249`; `cmd.go:585-595`, `:644-651` |
| V9 | **Sonda de lectura** (módulo desechable en `/tmp`, sin red, `modernc.org/sqlite v1.59.0` de la caché de módulos, binario `go1.27.1` con `GOTOOLCHAIN=local`): sobre un `world.db` en WAL cerrado, **`mode=ro` deja `world.db-wal` y `world.db-shm` creados al cerrar**; **`mode=rw&_pragma=query_only(1)` lee y no deja ningún auxiliar**, con `world.db` byte a byte igual; un fichero de 0 bytes se lee como base sin tablas, sigue en 0 bytes y en `journal_mode=delete`; `mode=rw` sobre un fichero ausente falla con 14 (`SQLITE_CANTOPEN`) **sin crear nada**; un fichero que no es base falla con 26 (`SQLITE_NOTADB`) y queda intacto; un directorio en la ruta da 14; con el directorio a `0500`, `mode=rw` da 1544 (`SQLITE_READONLY_DIRECTORY`) y `mode=ro&immutable=1` lee sin crear nada; con un escritor abierto, el lector ve lo confirmado en el WAL y no puede escribir (8), y no cambia `world.db` ni `world.db-wal`, **pero sí `world.db-shm`**, donde deja sus marcas de lectura (la huella de `-shm` cambia al cerrar el lector, como ya mostraba la salida de esta sonda; V36 F lo repite con los dos modos). La afirmación «el lector no toca los auxiliares del escritor» de la primera versión de este plan era falsa. Esta sonda no ejerce un `-wal` huérfano: lo hace V36 D | salida de `sh /tmp/kitlegal-sonda-h7/sonda.sh` (2026-09-28) |
| V10 | **Sonda de escritura**: abrir un `world.db` de 0 bytes con `_pragma=journal_mode(WAL)` en la cadena de conexión lo lleva a **4096 bytes aunque la transacción se deshaga**; con `foreign_keys(1)`, una arista hacia nodos inexistentes falla con 787; ocho escritores concurrentes con `_txlock=immediate` que migran dentro de la transacción si la versión es 0 aplican **una** migración y las ocho filas | salida de `sh /tmp/kitlegal-sonda-h7/escritura.sh` (2026-09-28) |
| V11 | **Sonda de WAL explícito** (tres ejecuciones): sin `journal_mode` en la cadena de conexión, leer la versión de un fichero de 0 bytes no escribe nada; ocho escritores que fijan `PRAGMA journal_mode=WAL` (con reintento ante `SQLITE_BUSY`) solo si la versión es 0 y después abren la transacción inmediata terminan los ocho, con una sola migración y la base en WAL; en una base ya en WAL, una transacción deshecha deja `world.db` con los mismos bytes | salida de `sh /tmp/kitlegal-sonda-h7/wal.sh` (2026-09-28) |
| V12 | Lo que H3 ya verificó del controlador y sigue valiendo (misma familia de versiones): los `_pragma=` se aplican **por conexión** con `busy_timeout` el primero; `_txlock=immediate` fija el modo de `BeginTx`; tablas `STRICT` e `INSERT … ON CONFLICT … DO UPDATE`; DDL dentro de una transacción se deshace; `*sqlite.Error` con `Code()`; al cerrar el último cliente desaparecen `-wal` y `-shm`; un `BeginTx` con el contexto de quien llama lo deshace `database/sql` desde otra goroutine (por eso `context.WithoutCancel`) | `specs/004-h3-internal-cache-sqlite/research.md:25-26`, `:28`; `internal/cache/migraciones.go:220-247` |
| V13 | Andamiaje de la caché que el grafo imita: cadena de conexión normal y de solo lectura, tramo de espera de 100 ms y espera total de 5 s, reintento por tramos que mira el contexto, migración releída **dentro** de la transacción inmediata, `schema_version(version, aplicada_en)` | `internal/cache/abrir.go:294-323`; `espera.go:11-27`, `:45-68`; `migraciones.go:83-143`, `:149-218`; `migraciones/0001_entradas.sql` |
| V14 | La regla de ubicación de la caché vive en `rutaEfectiva` + `compruebaRuta` (variable presente y vacía → «argumentos»; ruta que existe y no es directorio → «argumentos»; sin variable y sin `HOME` → «argumentos» con «declara HOME o KITLEGAL_CACHE_DIR») y **no se exporta**; la superficie exportada de `internal/cache` es una lista cerrada vigilada por un test | `internal/cache/ruta.go:48-86`, `:123-132`; `internal/cache/superficie_test.go:68-89` |
| V15 | `internal/core/schema` solo puede importar `crypto/sha256`, `encoding/hex`, `encoding/json` y `time`; lo exige un test | `internal/core/schema/contexto_test.go:51-62` |
| V16 | Los campos de `schema.Resultado` son hoy `Procedencia, Datos, Legible, Ensayo`, y un test lo fija | `internal/core/schema/sobre.go:89-112`; `sobre_test.go:178-199` |
| V17 | Flujo del kernel: `Main` → `resolver` → `ejecutarVerbo` (contexto con el plazo de `--timeout`, cancelado al volver del applet) → `Montador.Emitir`, que presenta y devuelve el código; `--dry-run` y la ayuda vuelven **antes** de `Emitir`; un fallo sale por `emitirFallo` | `internal/app/main.go:138-165`, `:395-437`; `internal/cli/sobre.go:72-94`, `:112-140` |
| V18 | Si la procedencia no declara fecha, el sobre la toma del reloj del montador en el momento de montarlo | `internal/cli/sobre.go:212-227` |
| V19 | La tabla mínima escribe `fecha_consulta` con `RFC3339Nano` y aplana `data` en pares ruta/valor | `internal/render/tabla.go:67-81`, `:115-130` |
| V20 | `boe articulo` firma con la dirección del bloque y `boe articulos` con la de la norma, con la fecha de consulta más antigua de lo que sostiene la respuesta; la vigencia de los dos es `vigenciaLarga` = 604 800 s; `hash_texto` es `sha256:` + SHA-256 de los bytes de `Texto`; `Articulo` lleva `URLELI` de los metadatos | `internal/source/boe/articulo.go:36-91`, `:287-302`; `fuente.go:24-32`, `:219-229`; `datos.go:28-58`, `:226-230`; `metadatos.go:20`, `:74` |
| V21 | La reproducción de `httpx` fecha cada respuesta con la hora del cliente al servirla, y `httpx.ConHora` la sustituye | `internal/httpx/cliente.go:210-229`, `:315-318` |
| V22 | Arnés e2e: cada guion tiene `KITLEGAL_CACHE_DIR=$WORK/cache` y una copia de las grabaciones en `$WORK/reproduccion/boe.legislacion-consolidada`; los binarios de variante se construyen con `-X` sobre variables del `package main` de e2e, que **no lee del entorno nada que decida su comportamiento**; existen `cronometra` y `arbol` | `internal/app/e2e_test.go:67-113`, `:178-209`, `:499-519`, `:598-607`; `internal/app/ejemplo/kitlegal-e2e/main.go` (comentarios de `version`, `enlazador` y `directorioDeReproduccion`) |
| V23 | La tarea `[aceptacion]` se verifica con el guardián, el rojo-primero **y `make ci`**; el rojo-primero rechaza un guion que falla por orden desconocida o de uso | `scripts/workflow/verificar.sh:33-52`; `scripts/workflow/aceptacion.sh` (`rojo-primero`) |
| V24 | Las evals se leen y validan contra `schemas/eval.yaml.json` dentro de `make ci` (`TestEvalsDelRepositorio`, en `skills-check`); `formaDelComando` decide la forma por el verbo; `Juzgar` exige comandos, citas, avisos y territorio | `internal/evals/formato.go:19-58`, `:119-130`, `:189-203`; `juzgar.go:206-241`, `:381-400`; `Makefile:123-124` |
| V25 | La preparación de una sesión ejecuta cada consulta **en proceso** con `app.Main` y un registro que solo tiene `boe`; la sesión corre con `KITLEGAL_CACHE_DIR="$d/cache"`; el intérprete de trazas monta `app.RegistroDeProduccion("")` para reconocer applets | `internal/evals/preparar.go:134-155`, `:226-273`, `:275-290`; `scripts/evals.sh:145`; `internal/evals/trazas.go:366-369` |
| V26 | La tabla de comandos de una skill lleva **todos** los verbos de cada applet de `kitlegal-applets` | `internal/skills/comandos.go:490-525`; `internal/skills/frontmatter.go:42` |
| V27 | Guiones e2e existentes afectados: `territorio-matriz.txtar:206` afirma `! exists cache` tras invocaciones de `territorio resolver` que terminan en 0; `boe-codigos.txtar:41` y `boe-verbos.txtar:58` solo tras invocaciones que fallan o que no ejecutan, así que siguen valiendo; listas literales de applets en `argumentos.txtar:24` y `:30`, `internal/app/registro_test.go:245-246`, `cmd/kitlegal/main_test.go:19` e `internal/app/ejemplo/kitlegal-e2e/main_test.go:90`; `ayuda.txtar:11-19` compara por expresión y `territorio` sigue siendo el nombre más largo | `git grep` sobre `internal/app/testdata/script/` y `*_test.go` |
| V28 | En `data/territorio/` los 8132 municipios tienen DIR3 verificado, así que ningún municipio real ejerce FR-044; el DIR3 de Leganés es `L01280745` (no el `L01280740` del ejemplo de `refs/`); la fecha de `municipios.yaml` es 2026-02-04 | `grep -c '^  "'` en `municipios.yaml` y `dir3.yaml`; `dir3.yaml:4367`; `municipios.yaml:1` |
| V29 | Reglas de arquitectura vigentes: `depguard` `core` ya deniega `internal/graph` al dominio y `sql` exceptúa `internal/graph/**`; `TestArquitectura` R3 no exige dueño (`duenoObligatorio` falso) porque falta `internal/store`, y `paquetesInternos` ya lista `graph`; `grafo.alcanza` recorre transitivamente dentro del módulo | `.golangci.yml` (listas `core` y `sql`); `internal/arch_test.go:91-114`, `:683-708`, `:782-784` |
| V30 | `misspell` v0.8.0 marca `observacion`→observation y `emision`→emission como palabras sueltas (también entre `_`); no marca `relacion`, `operacion`, `procedencia`, `hallazgo(s)`, `explicacion`, `vigencia`, `nodos`, `aristas`, `textos`, `huella`, `entrantes`, `salientes`, `recuento`, `persona`, `documento`, `separador`, `lote`, `caducada`, `obsoleta`, `reciente`, `instante`, `almacen`, `lectura`, `nulo`, `desempate`, `consolidado`; `versiones` ya está en `ignore-rules` | `…/golangci/misspell@v0.8.0/words.go:8516`, `:25592`, `:20323`; guion `/tmp/kitlegal-sonda-h7/misspell.sh`; `.golangci.yml` (`misspell.ignore-rules`) |
| V31 | Objetivos de `make` que el hito usa: `test` y `test-integration` saltan `MEDIDAS_DE_TIEMPO`, que `test-tiempos` ejecuta sola en `internal/app` | `Makefile:78-92` |
| V32 | `cksum` y `env -u` existen en la máquina de desarrollo (POSIX) | `command -v cksum env`; `env -u HOME true` |
| V33 | `dupl` tiene umbral 100 y `gocyclo` 13 en la configuración del repositorio; `testscript` tiene la orden `symlink file -> target` | `.golangci.yml:199`, `:244`; `…/go-internal@v1.16.0/testscript/doc.go:231-232` |
| V34 | Un applet sin verbo por omisión invocado sin verbo sale con 2 y un mensaje que nombra el applet y enumera sus verbos en el orden en que los declara | `cmd/kitlegal/main_test.go:105-129` (casos de `boe`, `skills` y `territorio`) |
| V35 | `TestEsquemasCubrenTodosLosVerbos` exige que cada verbo del registro de producción tenga su parte en exactamente un fichero publicado, y `ficherosDeEsquemas` es la tabla de esos ficheros; `go help build` no dice que `-o` cree los directorios intermedios (el quickstart los crea antes) | `internal/app/esquemas_test.go:59-64`, `:326-338`; `go help build` (líneas 26-28 de su salida) |
| V36 | **Sonda del corrector del plan** (módulo desechable, `modernc.org/sqlite v1.59.0` de la caché de módulos, `go1.27.1`, sin red; seis casos). **A**: un `world.db` de 0 bytes abierto sin `journal_mode` en la cadena de conexión, con `PRAGMA journal_mode=WAL` y cerrado **sin abrir ninguna transacción**, pasa a **4096 bytes** (una página: la cabecera), en `wal` y sin tablas, sin ningún otro fichero. **B**: una base en modo rollback sin el esquema del grafo (8192 bytes) con `PRAGMA journal_mode=WAL` y cerrada conserva el tamaño y cambia los bytes **18, 19, 27 y 95** de la cabecera (versiones de escritura y lectura 1 → 2, y el byte bajo del contador de cambios y de `version-valid-for`); esa base la escribió este mismo controlador, con la cabecera que él deja: con cabeceras de otro origen cambian más bytes (V41). **C**: un temporal de `os.CreateTemp(dir, "world.db-nuevo-*")` (`0600`) en WAL, con esquema y filas en una transacción confirmada, queda **solo** al cerrarlo (sin `-wal` ni `-shm`); `os.Link(temporal, world.db)` lo publica; un segundo `os.Link` sobre el `world.db` que ya existe falla con un error que cumple `errors.Is(err, fs.ErrExist)`; tras borrar el temporal, `world.db` se abre en `wal` con sus filas y, al cerrarlo, sigue con los mismos bytes y sin auxiliares. **D**: con un `-wal` huérfano con marcos (copia de los tres ficheros de un escritor abierto) y ninguna otra conexión, `mode=rw&_pragma=query_only(1)` lee las dos filas y **al cerrar hace un checkpoint**: `world.db` pasa de 4096 a 8192 bytes y desaparecen `-wal` y `-shm`; `mode=ro&_pragma=query_only(1)` lee las dos filas y deja `world.db` y `world.db-wal` **byte a byte iguales**, y solo cambia `world.db-shm` (o lo crea, de 32 768 bytes, si faltaba); `mode=ro&immutable=1` no ve lo que está en el WAL («no such table»). **E**: sobre una base en WAL cerrada limpia, `mode=rw` + `query_only` no deja nada y `mode=ro` deja `world.db-shm` (32 768 bytes) y un `world.db-wal` de 0 bytes (lo que ya decía V9). **F**: con un escritor abierto con marcos en el WAL, un lector en `mode=ro` o en `mode=rw` + `query_only` deja `world.db` y `world.db-wal` iguales y cambia `world.db-shm` | salida de `go -C /tmp/corrector-h7/wal run . /tmp/corrector-h7/datos-2` (2026-09-28) |
| V37 | **Sonda de las clases de caracteres**: la expresión de contracts/almacen-world-db.md §5 (sin `(?i)`, clases ASCII explícitas) rechaza los 20 valores de rechazo de ese contrato y acepta los 11 de aceptación, con 0 fallos; la anterior, con `(?i)`, aceptaba `ſ12345678Z` y rechazaba `12345678` + U+212A. `unicode.IsSpace` es la propiedad White_Space de Unicode (en Latin-1: `\t`, `\n`, `\v`, `\f`, `\r`, espacio, U+0085, U+00A0; fuera, la tabla `White_Space`, que incluye U+2003 y no U+200B); `unicode.IsControl` solo es cierto en Latin-1 y ahí coincide con la categoría Cc (U+0000-U+001F, U+007F-U+009F); U+0085 es de las dos clases; un byte que no es UTF-8 se lee como U+FFFD, que no es de ninguna. `time.Parse("20060102", v)` rechaza un signo, cifras de anchura completa, una longitud distinta de 8 y una fecha que no existe | `go -C /tmp/corrector-h7/regex run .` y `go -C /tmp/corrector-h7/regex test -run TestFechaVigencia -v .` (2026-09-28); `go doc unicode.IsSpace`, `go doc unicode.IsControl`; `$(go env GOROOT)/src/unicode/graphic.go:81-87`, `:128-138` |
| V38 | Cotas de tiempo existentes que la entrega alcanza: `internal/app/testdata/script/boe-cache-rapida.txtar:19-57` (diez `cronometra 200ms $KITLEGAL_BIN boe articulo … --json` servidos de la caché, cada uno con `cmp stdout` y `! stderr .`) y `territorio-matriz.txtar:192-202` (tres `cronometra 200ms $KITLEGAL_BIN territorio resolver … --json`, con `cmp stdout` y `! stderr .`); `TestMedidasDeTiempo` ejecuta los guiones que cronometran (`internal/app/e2e_test.go:487-495`) en `make test-tiempos` (`Makefile:80`, `:91-92`). En esas invocaciones la entrega no cambia nada: la primera invocación de cada guion ya observó lo mismo, con la misma `fecha_consulta` (la de la caché, o `2026-02-04T00:00:00Z` en `territorio`) | lectura de los guiones y del `Makefile`; D13 (una observación idéntica no cambia nada) |
| V39 | El binario no atiende `SIGINT` ni `SIGTERM`: `Main` solo ignora `SIGPIPE`, así que una señal termina el proceso sin ejecutar ningún `defer`; el plazo por omisión de `--timeout` es 30 s | `internal/app/main.go:185`; `internal/cli/globales.go:37` |
| V40 | `testscript` parte los argumentos solo por espacio ASCII, tabulador, `\r` y `#` fuera de comillas, así que U+00A0, U+2003, U+007F y U+0085 viajan dentro de un argumento de `exec`; U+0000 no puede viajar en ningún argumento de un proceso (las cadenas de `argv` terminan en NUL) | `…/go-internal@v1.16.0/testscript/testscript.go:1310` |
| V41 | **Sonda de la cabecera** (módulo desechable, `modernc.org/sqlite v1.59.0`, que lleva SQLite 3.53.4 —`sqlite_version()`; `SQLITE_VERSION_NUMBER` 3053004—, `go1.27.1`, sin red). Camino del paso 4 de D11 sobre un `world.db` sin esquema fuera de WAL: abrir con la cadena de escritura, leer la versión, `PRAGMA journal_mode=WAL` (devuelve `wal`); otra conexión toma la transacción inmediata y la mantiene; la propia falla con `SQLITE_BUSY` (5); se cierran las dos. Queda solo `world.db`, del mismo tamaño salvo el de 0 bytes, en `wal` y sin el esquema del grafo, y cambian exactamente estos bytes de la cabecera de 100: **0 bytes** → 4096 bytes (cabecera nueva entera); **base en rollback escrita por este controlador** (contador 2, tamaño en cabecera 2 válido, versión 3053004) → 18, 19, 27, 95; **escrita por otra versión** (3043002 en 96-99) → 18, 19, 27, 95, 98, 99 (96-99 pasa a 3053004); **tamaño en cabecera 0** (28-31), con `version-valid-for` igual al contador o distinto (7) → 18, 19, 27, 31, 95 (28-31 pasa al número de páginas, 2); **`version-valid-for` distinto (7) con el tamaño correcto** → 18, 19, 27, 95 (28-31 ya tenía el valor que se escribe); **contador 511 = `version-valid-for`** → 18, 19, 26, 27, 94, 95 (el acarreo alcanza más bytes de 24-27 y 92-95); **una página de más al final del fichero** (12 288 bytes, tamaño en cabecera 2 válido) → 18, 19, 27, 95 y el fichero no se trunca; **las tres cosas a la vez** → 18, 19, 27, 31, 95, 98, 99. Un tamaño en cabecera válido **mayor** que el fichero da `SQLITE_CORRUPT` (11) al leer la versión y no cambia nada (inutilizable). Estos conjuntos son exactos **para esas bases**, sin páginas libres y sin diario; no agotan lo que una confirmación hace sobre cualquier base de fuera (páginas libres con `auto_vacuum=full` y diarios fríos o vacíos, V45) | salida de `go -C /tmp/corrector-h7-r3/sonda run . /tmp/corrector-h7-r3/datos-1`, líneas `residuo` (2026-09-28) |
| V42 | **WAL dentro de una transacción** (misma sonda): con `BEGIN IMMEDIATE` abierto, `PRAGMA journal_mode=WAL` sobre un `world.db` de **0 bytes** devuelve `delete` **sin error**, y tras crear una tabla y confirmar la base queda en `delete` (8192 bytes, fuera de WAL); sobre una **base en rollback** falla con «SQL logic error: cannot change into wal mode from within a transaction (1)». El modo WAL no se puede fijar dentro de la transacción que crea el esquema, y con 0 bytes el fallo es silencioso: solo se ve en el valor que devuelve el pragma | ídem, líneas `tx` (2026-09-28) |
| V43 | **Diario de rollback** (misma sonda). **Caliente** —copia de `world.db` y `world.db-journal` de un escritor en rollback con la transacción a medias y páginas ya volcadas (`cache_size(2)`), sin ningún bloqueo—: `mode=rw&_pragma=query_only(1)` lee y **deshace** la transacción interrumpida: `world.db` cambia (vuelve a lo confirmado) y `world.db-journal` desaparece; la cadena de escritura hace lo mismo en su primera lectura; `mode=ro&_pragma=query_only(1)` falla con 776 (`SQLITE_READONLY_ROLLBACK`, «attempt to write a readonly database») y deja `world.db` y `world.db-journal` con los mismos bytes; `mode=ro&immutable=1` lee sin deshacer nada, así que ve las páginas que la transacción sin confirmar ya había volcado. **Frío** (diario con la cabecera a cero, como el que deja `journal_mode=PERSIST`): `mode=ro` y `mode=rw` + `query_only` leen y no cambian ni `world.db` ni el diario. **De un escritor vivo** (en rollback, con su transacción abierta): `mode=ro` lee lo confirmado y no cambia nada | ídem, líneas `diario`; `go -C /tmp/corrector-h7-r3/sonda run . /tmp/corrector-h7-r3/datos-2 diario2`, líneas `diario2` (2026-09-28) |
| V44 | **`-wal` huérfano en una entrega que falla** (misma sonda): copia de `world.db` (4096 bytes), `-wal` y `-shm` de un escritor abierto con dos nodos confirmados en el WAL; una conexión con la cadena de escritura lee la versión (1), abre la transacción inmediata, inserta y la **deshace** (un rechazo), y cierra: `world.db` pasa a 16 384 bytes (el cierre, última conexión, hace el checkpoint) y desaparecen `-wal` y `-shm`; el grafo sigue con sus dos nodos. Con una versión posterior (2) confirmada en el WAL, solo leer la versión y cerrar hace lo mismo | ídem, líneas `huerf.` (2026-09-28) |
| V45 | **Lo que escribe el paso a WAL de una base de fuera** (sonda del corrector tras la ronda 3 del juez: módulo desechable, `modernc.org/sqlite v1.59.0`, `go1.27.1`, sin red; mismo camino que V41: cadena de escritura, leer la versión, `PRAGMA journal_mode=WAL`, otra conexión retiene la transacción inmediata, la propia falla con 5, se cierran las dos; datos con `zeroblob`, deterministas). V41 midió solo cabeceras sobre bases sin páginas libres y sin diario; el paso a WAL es **una confirmación en modo rollback** y hace todo lo que SQLite hace al confirmar: **(a)** con `auto_vacuum=full` y páginas libres —el estado que deja una aplicación que usa `sqlite3_autovacuum_pages`; en la sonda, `auto_vacuum=INCREMENTAL`, una tabla `a` con 40 filas de `zeroblob(3000)` borradas y los bytes 64-67 puestos a 0— la confirmación **vacía todas las páginas libres y trunca el fichero**: 176 128 → 12 288 bytes, y en los 12 288 que quedan cambian 18, 19, 27, 31, 35, 39 y 95 (28-31: 43 → 3 páginas; 32-35, primera página de la lista libre: 4 → 0; 36-39, páginas libres: 40 → 0); con dos tablas y las libres al final, 180 224 → 16 384 y los mismos siete bytes; con una página en uso **detrás** de las libres (tablas `a` y `otra` creadas antes, 40 filas en `a`, después 2 en `otra`, `a` vaciada), 188 416 → 24 576 bytes y 59 bytes distintos: los siete de la cabecera y otros en las páginas 2 (mapa de punteros), 4 (la página que apunta a las reubicadas) y 5 (una reubicada). En todos, las tablas y sus filas son las mismas y `auto_vacuum` sigue en `full`; con `auto_vacuum=INCREMENTAL` no se vacía nada (18, 19, 27, 95, mismo tamaño). **(b)** Un `world.db-journal` **frío** de `journal_mode=PERSIST` (8720 bytes) o **vacío** de `TRUNCATE` (0 bytes) **desaparece ya con el pragma**, con la conexión aún abierta y sin ninguna otra: la confirmación en modo `DELETE` borra el diario al terminar; `world.db` cambia en 18, 19, 27 y 95, como en V41. Con `world.db` de 0 bytes y un diario vacío: 0 → 4096 bytes y el diario desaparece. **(c)** Al leer (`mode=ro` y `mode=rw` + `query_only`), el diario frío de `PERSIST`, el vacío de `TRUNCATE` y el vacío junto a un `world.db` de 0 bytes quedan con los mismos bytes, igual que `world.db` | `go -C /tmp/corrector-h7-r4/sonda run . /tmp/corrector-h7-r4/sonda/datos residuo` y `… lectura` (2026-09-28); sonda del juez, reejecutada: `go -C /tmp/juez-h7-r3/sonda run . /tmp/corrector-h7-r4/sonda/datos-juez persist`; fuente: `_autoVacuumCommit` (`modernc.org/sqlite@v1.59.0/lib/sqlite_g_000000000001feab.go:1540-1616`: con `incrVacuum` a 0 vacía hasta `nFin`, escribe 32-35 y 36-39 a 0 en `:1605-1606` y 28-31 en `:1608`, y marca el truncado en `:1609`), `_pager_end_transaction` (`sqlite_g_000000000001ffff.go:4693-4749`, borra el diario en `:4749`) y `_hasHotJournal` (`:3578-3622`, que borra el de una base de 0 páginas en `:3622`); `Xsqlite3_autovacuum_pages` en `lib/sqlite.go:10188` |
| V46 | **`world.db` que el proceso no puede escribir** (misma sonda, `… solo-lectura`; sonda del juez `… solo-lectura`): `world.db` en WAL con permisos `0400`, el directorio en `0700` y sin auxiliares. `mode=rw` + `query_only` y `mode=ro` leen sin error —SQLite cae a solo lectura sin avisar— y al cerrar dejan `world.db-shm` (32 768 bytes) y `world.db-wal` (0 bytes), los dos con los permisos de `world.db` (`0400`), que una conexión de solo lectura no puede borrar; la cadena de escritura lee la versión y abre `BEGIN IMMEDIATE` sin error y deja los mismos dos ficheros (sonda del juez). En modo rollback (`0400`) ninguno deja nada. `os.OpenFile(ruta, os.O_RDWR, 0)` falla con `permission denied`, `errors.Is(err, fs.ErrPermission)` es cierto, y no cambia ni los bytes ni la fecha de modificación de `world.db`; sobre un `world.db` que sí se puede escribir, abrirlo así y cerrarlo tampoco (misma huella y misma fecha). `mode=ro&immutable=1&_pragma=query_only(1)` lee la fila de `world.db` `0400`, en WAL y en rollback, y **no crea ningún fichero** ni cambia `world.db`. Con los auxiliares que dejó una lectura anterior (`-wal` de 0 bytes y `-shm`, los dos `0400`), `mode=ro` e `immutable=1` leen y ningún fichero cambia | `go -C /tmp/corrector-h7-r4/sonda run . /tmp/corrector-h7-r4/sonda/datos solo-lectura` (2026-09-28); sonda del juez, reejecutada: `go -C /tmp/juez-h7-r3/sonda run . /tmp/corrector-h7-r4/sonda/datos-juez solo-lectura`; `go doc os.OpenFile`, `go doc io/fs.ErrPermission` |
| V47 | **`world.db` que es un enlace simbólico** (misma sonda, `… enlace`): con `cache/world.db` → `real/grafo.db` (en WAL), una conexión abierta por el nombre crea `real/grafo.db-wal` y `real/grafo.db-shm`, **junto al destino y con su nombre**, y ninguno junto al enlace: fuera de Windows, SQLite resuelve cada componente de la ruta (`_appendOnePathElement`, `sqlite_g_0000000000000003.go:13625-13670` en darwin y `sqlite_linux_amd64.go:36` en linux/amd64, que sigue `S_IFLNK`). En Windows, `_winFullPathnameNoMutex` (`sqlite_windows.go:125069-125126`) usa `GetFullPathNameW` (`_aSyscall[25]`, que apunta a `libc.XGetFullPathNameW` en `:130590`), que no sigue enlaces: allí los nombra por la ruta sin resolver (leído en el fuente, sin ejecutar: S5). Con un `-wal` huérfano junto al destino, `os.Stat("cache/world.db-wal")` no lo ve y `os.Stat(<ruta resuelta>-wal)` sí; `mode=rw` + `query_only` por el nombre hace el checkpoint al cerrar (el destino pasa de 4096 a 8192 bytes y sus auxiliares desaparecen) y `mode=ro` lo deja con los mismos bytes | `go -C /tmp/corrector-h7-r4/sonda run . /tmp/corrector-h7-r4/sonda/datos enlace` (2026-09-28) |
| V48 | **`world.db-shm` suelto, sin `-wal`**; el resultado depende del **origen** del `-shm`. **De un lector sobre una base limpia, sin marcos en el índice** (sonda del corrector, `… shm-suelto`, `/tmp/corrector-h7-r4/sonda/main.go:402-425`; `world.db` en WAL cerrado limpio, que se puede escribir, y una copia del `-shm` de esa conexión abierta): `mode=ro` lee, deja `world.db` y `world.db-shm` con los mismos bytes y **crea un `world.db-wal` de 0 bytes**; `mode=rw` + `query_only` lee y, al cerrar, **borra** `world.db-shm`. **De un escritor con marcos en el WAL** (sonda del juez, caso E: copia del `-shm` de un escritor abierto con marcos, sin su `-wal`): `mode=ro` lee, deja `world.db` igual, **reescribe** `world.db-shm` (huella 31996ab8 → fd4c9fda: SQLite reconstruye el índice porque el `-wal` que describe no está) y crea un `world.db-wal` de 0 bytes. Las dos caen en la causa declarada de D10 (lo que SQLite escribe en los auxiliares de WAL para leer); que el `-shm` quede igual **no** es general. Ningún cierre del binario deja ese estado: SQLite borra `-wal` y `-shm` al cerrar la última conexión (V12) | `go -C /tmp/corrector-h7-r4/sonda run . /tmp/corrector-h7-r4/sonda/datos shm-suelto` (2026-09-28); sonda del juez de la ronda 4 del plan, caso E (`gates/plan-pendiente.md`, motivo 2) |
| V49 | **`world.db` de 0 bytes con un `world.db-wal` no vacío** (sonda del juez de la ronda 4 del plan, casos A, A2, B y G: `world.db` de 0 bytes y `world.db-wal` de 12 392 bytes, copia del WAL de un escritor con marcos): `mode=ro&_pragma=busy_timeout(100)&_pragma=query_only(1)` lee 0 tablas y **borra** `world.db-wal`; igual con el `-shm` presente (el `-shm` queda igual) y con `world.db` en `0400`; `mode=rw` + `query_only` y la cadena de escritura hacen lo mismo en su primera lectura; solo `mode=ro&immutable=1` lo deja intacto. Con un `-wal` de 0 bytes no se borra nada (SQLite no lo cuenta como existente). Causa: `_pagerOpenWalIfPresent` borra el `-wal` de una base de 0 páginas (`_sqlite3OsDelete`, l. 4478) en cualquier modo que no sea `immutable`, **sin** llevar sus marcos a `world.db`: no es una recuperación sino un descarte. Un fichero de 0 bytes es siempre una base sin esquema (versión 0), así que su versión se sabe sin abrir SQLite | `go -C /tmp/juez-h7-r4/sonda run . /tmp/juez-h7-r4/datos` (`gates/plan-pendiente.md`, motivo 1); `modernc.org/sqlite@v1.59.0/lib/sqlite_g_000000000001ffff.go:4461-4490` |

## D · Decisiones

### D1 · Dónde vive cada pieza

**Decisión.** Las **operaciones** viajan en `schema.Resultado` y sus tipos se declaran en `internal/core/schema`
(`schema.Observado`, `schema.Operacion`, `schema.Nodo`, `schema.Arista`, `schema.Texto`), sin importar nada nuevo
(V15). El **puerto** `core.GraphStore` y su argumento `core.Lote` viven en la raíz del dominio,
`internal/core/graphstore.go`, junto a `Cache` y `Source` (el comentario de `internal/core/doc.go` ya lo anunciaba).
La **lógica del grafo** —validación del lote, rechazo de `Persona`, fusión de observaciones, `check`, explicaciones y
las formas de salida de los verbos— es dominio puro en un paquete nuevo, `internal/core/grafo`. El **adaptador**
SQLite es `internal/graph` (nombre fijado por el hito). La **composición** —applet `graph`, registro, entrega— vive en
`internal/app`, y el **kernel** (`internal/cli`) monta el lote y lo entrega.

**Alternativas.** Tipos de operación en `internal/core/grafo` y `schema.Resultado` importándolo: rompe
`TestContexto` (V15) y crearía un ciclo en cuanto `grafo` necesitara `schema`. La lógica de `check` en SQL dentro de
`internal/graph`: saca reglas del dominio a un adaptador, fuera del umbral de cobertura de `internal/core/**` y más
difícil de probar con tablas.

### D2 · Nombres del campo y de los tipos (ADR 0014 los deja al plan)

**Decisión.** `schema.Resultado` gana el campo `Grafo schema.Observado`; `Observado{Vigencia time.Duration;
Operaciones []Operacion}`; `Operacion` es una interfaz **sellada** (método no exportado) que implementan `Nodo{ID,
Tipo string; Datos map[string]any}`, `Arista{Origen, Relacion, Destino string}` y `Texto{Huella, Cuerpo string}`. El
puerto es `GraphStore{ Apply(ctx context.Context, lote Lote) error }` —el nombre que usa el hito— y
`Lote{Fuente, URL, FechaConsulta string; Vigencia time.Duration; Operaciones []schema.Operacion}`.

**Alternativas.** Un `struct` con un campo de clase y tres punteros: admite valores sin sentido (dos a la vez, o
ninguno) que habría que rechazar en ejecución. Nombres con «Emision» u «Observacion» sueltos: `misspell` los marca
(V30).

### D3 · La vigencia viaja en `Observado`, no en `Procedencia`

**Decisión.** La vigencia que la fuente declara para la consulta (FR-065) es `Observado.Vigencia`; cero significa «no
la declara». `boe` pone la vigencia con la que guarda la consulta en la caché (604 800 s, V20); `territorio`, ninguna.

**Alternativa.** `Procedencia.Vigencia`: la procedencia es lo que hace citable un sobre y la usan todos los applets y
todos los fallos; meterle un dato que solo le importa al grafo cambiaría un tipo del contrato (V16) sin necesidad.

### D4 · La entrega la hace el `Montador`, justo después de presentar

**Decisión.** `cli.Montador` gana el campo `Grafo core.GraphStore` (nulo: no entrega) y `Emitir` recibe un
`context.Context` como primer argumento. Tras presentar un resultado correcto —y solo entonces—, si
`res.Grafo.Operaciones` no está vacío, monta el lote con la procedencia **del sobre ya montado** y llama a
`Apply`. `internal/app/main.go` guarda en el desenlace el plazo de la invocación y `--no-graph`, crea el contexto con
ese plazo y elige el almacén: `graph.Nulo{}` con `--no-graph`, el del registro sin ella. Así FR-026 (después de
presentar), FR-030 (sin operaciones no se abre nada), FR-032 (un fallo no entrega) y FR-034 (`--dry-run` vuelve antes
de `Emitir`, V17) son propiedades de la estructura y no de la disciplina de cada applet.

**Alternativas.** *Volver a montar el sobre en `Main` para leer su fecha*: con una procedencia sin fecha, el reloj
daría otro instante que el del sobre presentado (V18). *Un gancho de entrega en `internal/app`*: partiría la
garantía «la procedencia la pone el kernel» entre dos paquetes. *Un `EmitirYEntregar` aparte*: dos caminos para
presentar un resultado.

### D5 · La fecha del lote es la del sobre, escrita igual

**Decisión.** `Lote.FechaConsulta = sobre.FechaConsulta.Format(time.RFC3339Nano)`: el mismo texto que el sobre
escribe en `fecha_consulta` (V2). El almacén guarda ese texto tal cual y compara instantes interpretándolo.

### D6 · `--no-graph` es `graph.Nulo`; un registro sin grafo no entrega

**Decisión.** `graph.Nulo` implementa `core.GraphStore` y descarta. El registro gana `EntregarAlGrafo(almacen
core.GraphStore)`; el de producción y el de e2e lo llaman con `graph.Nuevo()`; el valor cero del registro —el de los
tests y el de las consultas con que la preparación de evals llena la caché— no entrega nada, de modo que ningún test
escriba en el `world.db` de la cuenta de quien los ejecuta. Quien necesita un grafo lo pide con un directorio
explícito: los tests, con `graph.Nuevo(graph.ConDirectorio(<temporal>))`, y la preparación del grafo previo de una
eval (D26), con `graph.Nuevo(graph.ConDirectorio(<caché de la sesión>))`.

**Alternativa.** Que el valor cero entregue al `world.db` de la cuenta: cualquier test que ejecute un verbo con un
registro propio escribiría en `$HOME`.

### D7 · El aviso de una entrega fallida es una línea y su escritura no se propaga

**Decisión.** Si `Apply` falla, el kernel escribe por `Presentador.Aviso` exactamente una línea:
`kitlegal: lo observado no ha llegado al grafo del mundo: <causa>`, con los saltos de línea de la causa sustituidos por
espacios. El error de escribir esa línea no cambia el código ni se propaga (FR-033: el sobre ya dijo `ok: true`), con
el mismo razonamiento que el aviso de versión de H19 (`internal/app/main.go`, `avisar`). Todo mensaje de
`internal/graph` nombra `world.db` (su ruta cuando se conoce), así que la línea nombra `world.db` y la causa.

### D8 · Ubicación: la regla de la caché, exportada una vez

**Decisión.** `internal/cache` exporta `func Directorio() (string, error)`, que es `rutaEfectiva("", os.LookupEnv)` +
`compruebaRuta` (V14), y `internal/graph` la usa: `world.db` queda en el mismo directorio que `cache.db`, con la misma
regla y los mismos errores de clase «argumentos» (FR-001, FR-011). `superficie_test.go` gana `func Directorio` en su
lista cerrada; la función no nombra la base de datos, así que la garantía de FR-005 de H3 se mantiene.

**Alternativas.** *Copiar la regla en `internal/graph`*: dos reglas que pueden divergir. *Un paquete compartido fuera
de `cache`, `store` y `graph`*: R3 solo concede SQLite a esos tres (constitución IV), y sería otra regla. *Refactorizar
`internal/cache` para exponer su andamiaje*: rompe FR-005 de H3 (la superficie no puede dejar asomar la base) y no lo
pide el hito.

### D9 · Andamiaje propio en `internal/graph`, con el diseño de la caché

**Decisión.** `internal/graph` implementa su apertura, su migración y su espera con las mismas decisiones que la caché
(V13): un solo `*sql.DB` por operación con una conexión, `busy_timeout(100)` por tramo, espera total de 5 s que mira
el contexto entre tramos, `synchronous(FULL)`, `_txlock=immediate`, migraciones embebidas en
`internal/graph/migraciones/*.sql` y `schema_version(version, aplicada_en)`. Es código propio, no copiado: sus
errores nombran `world.db`, su lectura elige el modo según los auxiliares y según que el proceso pueda escribir
`world.db` (D10), y su escritura comprueba ese permiso antes de abrir SQLite, crea `world.db` en un temporal que
publica con `os.Link` y, en un `world.db` que ya existe sin esquema, fija WAL fuera de la cadena de conexión (D11). `dupl`
(umbral 100) vigila que no nazca un clon; si lo marcara, se reestructura, nunca con `//nolint`.

**Alternativa.** Exportar el andamiaje de la caché: ver D8.

### D10 · Lectura: el modo según los auxiliares y el permiso de escritura, e inmutable como recurso

**Decisión.** Antes de abrir, los verbos de `graph`:

1. hacen `os.Stat` de `world.db`: ausente (también un enlace simbólico sin destino) → grafo vacío sin abrir nada;
   directorio → inutilizable; un fichero de **0 bytes** → grafo vacío **sin abrir SQLite**, haya los auxiliares que
   haya: un fichero de 0 bytes es siempre una base sin esquema (versión 0), y abrirlo en cualquier modo que no sea
   `immutable` borraría un `world.db-wal` no vacío junto a él (`_pagerOpenWalIfPresent`, V49), contra FR-004;
2. buscan sus tres auxiliares posibles, `-wal`, `-shm` y `-journal`, **con el nombre que les da SQLite**: fuera de
   Windows, a partir de la ruta resuelta con `filepath.EvalSymlinks`, porque SQLite sigue los enlaces y crea los
   auxiliares junto al destino y con su nombre; en Windows, a partir de la ruta tal cual, porque allí no los sigue
   (V47). Sin enlaces, las dos son la misma ruta;
3. comprueban si el proceso puede escribir `world.db`: `os.OpenFile(<ruta>, os.O_RDWR, 0)` y cerrarlo enseguida, sin
   leer, escribir ni truncar, que no cambia sus bytes ni su fecha de modificación (V46). `nil` → puede;
   `fs.ErrNotExist` (lo retiraron después del `Stat`) → grafo vacío, como en 1; cualquier otro error (sin permiso,
   sistema de ficheros de solo lectura…) → no puede;
4. eligen el modo por lo que encuentran:

- **Puede escribir y no hay ningún auxiliar** (el estado normal: ninguna otra conexión y el último cierre limpio):
  `file:<ruta>?mode=rw&_pragma=busy_timeout(100)&_pragma=query_only(1)`. Lee sin crear el fichero si no existe, sin
  poder escribir y sin dejar ningún auxiliar, con `world.db` byte a byte igual (V9, V36 E).
- **No puede escribir y no hay `-wal` ni `-journal`** (p. ej., `chmod a-w` sobre el `world.db` que creó el binario, en
  WAL): `file:<ruta>?mode=ro&immutable=1&_pragma=query_only(1)`, que lee `world.db` tal cual y no crea nada (V46).
  `mode=rw` + `query_only` y `mode=ro` también lo leerían —SQLite cae a solo lectura sin avisar—, pero sobre una base
  en WAL dejan `world.db-shm` y `world.db-wal`, que ninguna conexión de solo lectura puede borrar (V46), contra FR-004
  y SC-004. Sin `-wal` no hay nada confirmado fuera de `world.db`, así que `immutable` lee todo el grafo. Que no tome
  bloqueos no importa aquí: ninguna invocación de la cuenta puede escribir ese fichero —sin ese permiso su entrega
  falla antes de abrirlo (D11)— y el directorio de la caché es de la cuenta (D8); es la misma premisa con la que ya se
  lee un directorio de solo lectura (última viñeta).
- **Con algún auxiliar si puede escribir, o con `-wal` o `-journal` si no puede** (otra conexión abierta, un `-wal`
  huérfano de un escritor interrumpido, un `-shm` suelto, un `world.db-journal`, o los que dejó otro programa que leyó
  sin permiso de escritura), con un `world.db` de más de 0 bytes (el de 0 bytes no llega aquí: paso 1):
  `file:<ruta>?mode=ro&_pragma=busy_timeout(100)&_pragma=query_only(1)`. Lee todo lo
  confirmado, también lo que sigue en el WAL, y deja `world.db`, `world.db-wal` y el diario byte a byte iguales
  (V36 D, F, V43); junto a una base de 0 páginas borraría el `-wal` (V49), y por eso ese caso se lee sin abrir
  SQLite; `mode=rw` + `query_only`, si al cerrar es la última conexión, haría un checkpoint que reescribe
  `world.db` y borra los auxiliares (V36 D). Un `world.db-journal` es el diario de un escritor en modo rollback, un
  estado que el binario no busca (el grafo se escribe siempre en WAL, D11): llega con un `world.db` de fuera o tras
  una interrupción. Si es **frío** (cabecera a cero, el de `journal_mode=PERSIST`), **vacío** (0 bytes, el de
  `TRUNCATE`) o de un escritor **vivo**, `mode=ro` lee lo confirmado sin cambiar nada (V43, V45 c). Si está
  **caliente** (una transacción interrumpida sin deshacer), `mode=ro` falla con 776 (`SQLITE_READONLY_ROLLBACK`) sin
  tocar ni `world.db` ni el diario (V43), y el verbo sale con 1, `inesperado`: «tiene una transacción interrumpida sin
  deshacer; no se modifica» (FR-010: el sistema no deja abrirlo para leer sin modificarlo). La deshará la entrega
  siguiente que pueda escribir `world.db`, que es un escritor (D11).
- En los modos que no son `immutable`, si la primera consulta da 1544 o 14 (directorio que no admite los auxiliares):
  si el fichero no se deja leer, inutilizable; si no hay `world.db-wal` ni `world.db-journal`, se reabre con
  `mode=ro&immutable=1&_pragma=query_only(1)` (V9); si hay alguno, inutilizable (`immutable` no leería lo que está en
  el WAL, V36 D, ni desharía un diario caliente, así que vería páginas sin confirmar, V43).

**Lo que no se puede evitar**, desviación declarada de FR-004, FR-031 y SC-004 en sus bytes (no en el contenido del
grafo, FR-005), en *Complexity Tracking* y en `gates/supuestos.md`, declarada por su causa —lo que SQLite escribe en
los auxiliares de WAL para leer lo confirmado en ellos— y con sus cotas: nunca cambia `world.db`, ni `world.db-wal`
si existe, ni el diario, ni el contenido del grafo. La cota del `-wal` se cumple también junto a un `world.db` de 0
bytes porque ese caso no abre SQLite (paso 1; V49):

- **Con algún auxiliar de WAL** (`-wal` o `-shm`) y un `world.db` de más de 0 bytes, SQLite escribe en `world.db-shm`
  —el índice del WAL: las marcas de lectura del lector y, tras un escritor interrumpido, la reconstrucción del
  índice— y lo crea, de 32 768 bytes, si faltaba (V36 D, F). Con un `world.db-shm` suelto, sin `-wal` —que ningún
  cierre del binario deja—, crea además un `world.db-wal` de 0 bytes, y lo que hace con el `-shm` depende de su origen
  (V48): el de un lector sobre una base limpia, sin marcos en el índice, queda igual; el de un escritor con marcos en
  el WAL se reescribe, porque SQLite reconstruye el índice de un `-wal` que ya no está. Cambian o aparecen esos dos
  auxiliares; `world.db` y lo que dicen `graph stats` y `graph show`, no.
- **Con una entrega concurrente**, si un escritor abre `world.db` entre el `Stat` y la apertura del lector y cierra
  antes que él, el lector sin auxiliares es la última conexión y su cierre lleva a `world.db` lo que ese escritor
  confirmó: los bytes cambian por la entrega, no por la lectura, y el contenido es el que esa entrega dejó.

Sin auxiliares y sin ninguna entrega a la vez —el estado de los guiones, del quickstart y de la matriz de SC-004—, la
lectura no cambia ni un byte de nada, pueda el proceso escribir `world.db` (V9, V36 E) o no (V46), y tampoco si
`world.db` es un enlace simbólico, porque los auxiliares se buscan donde SQLite los crea (V47). Con un diario frío,
vacío o de un escritor vivo, tampoco (V43, V45 c). Con un diario caliente no se lee, pero tampoco se cambia nada: no es
una desviación, es la salida de FR-010.

**Alternativas.** *`mode=ro` siempre* (la de la caché): deja `-wal` y `-shm` en el estado normal (V9, V36 E), contra
FR-004 y US3.7 en el caso común. *Elegir el modo sin mirar el permiso de escritura* (la versión anterior de este
plan): sobre un `world.db` en WAL que el proceso no puede escribir, cualquier modo que no sea `immutable` deja
`world.db-shm` y `world.db-wal` (V46), y `chmod a-w` basta para llegar a ese estado. *Comprobar el permiso con
`access(2)`* (`golang.org/x/sys/unix.Access`): mira el usuario real y no el efectivo, no existe en Windows y añadiría
el primer import de un módulo que `go.mod` declara `// indirect` (l. 29); `os.OpenFile` es de la biblioteca estándar, mira lo mismo
que mirará SQLite al abrir y en Windows ve el atributo de solo lectura. *Leer la cabecera de `world.db` para usar
`immutable` solo si está en WAL*: sin permiso, en rollback, `mode=ro` tampoco deja nada (V46), pero añade un analizador
de la cabecera para distinguir dos lecturas que dan lo mismo. *Borrar los auxiliares que deje una lectura sin permiso*:
podrían ser de otra conexión. *Buscar los auxiliares siempre junto al nombre*: con un `world.db` que es un enlace, no ve
un `-wal` huérfano junto al destino, lee con `mode=rw` + `query_only` y el cierre hace el checkpoint sobre el destino
(V47). *Buscarlos siempre junto a la ruta resuelta*: en Windows, donde SQLite no sigue enlaces, no vería los suyos
(V47). *Leer con `immutable=1` también con un `-shm` suelto y permiso de escritura*: sin bloqueos, una entrega que
empiece a la vez y haga su checkpoint durante la lectura daría una instantánea incoherente. *Mirar solo `-wal` y `-shm`* (la versión anterior de este plan): con un diario
caliente, `mode=rw` + `query_only` deshace la transacción interrumpida al leer y reescribe `world.db` (V43), contra
FR-004. *Leer un diario caliente con `immutable=1`*: daría páginas de una transacción sin confirmar (V43).
*`mode=rw` + `query_only` siempre* (la primera versión de este plan): con un `-wal` huérfano, el cierre reescribe
`world.db` y borra los auxiliares (V36 D). *`immutable=1` cuando hay auxiliares*: no
toma bloqueos ni lee el WAL, así que daría un grafo sin lo confirmado en él (V36 D), contra FR-005 y FR-053. *Copiar
`world.db` y `-wal` a un temporal y leer la copia*: dos lecturas no atómicas de dos ficheros que un escritor puede estar
cambiando dan una instantánea incoherente, y la copia no distingue un `-wal` huérfano de uno en uso. *Desactivar el
checkpoint al cerrar* (`SQLITE_DBCONFIG_NO_CKPT_ON_CLOSE`): no evita la escritura en `-shm` y exige llegar a la API C del
controlador, que ninguna verificación de este plan cubre. *Abrir con el modo que dictan los auxiliares también un
`world.db` de 0 bytes* (la versión anterior de este plan): con un `-wal` no vacío junto a él, cualquier modo que no sea
`immutable` lo borra sin llevar nada a `world.db` (V49), contra FR-004, FR-031 y SC-004. *Leer el de 0 bytes con
`mode=ro&immutable=1`*: tampoco toca el `-wal` (V49), pero abre SQLite para saber lo que el tamaño ya dice (versión 0,
grafo vacío) y deja la corrección de la lectura a un detalle del modo; no abrir nada es lo más simple y no depende de
SQLite.

### D11 · Escritura: validar antes de tocar el disco; crear en un temporal y publicarlo con `os.Link`

**Decisión.** `Almacen.Apply`:

1. valida y consolida el lote en el dominio (`grafo.Consolidar`, que llama a `ValidarLote`: FR-024 lo que no necesita
   la base, FR-025, y unos datos sin forma JSON canónica) sin tocar el disco, rechaza una vigencia que no es un número
   entero de segundos —la columna `ttl` guarda segundos, y recortarla guardaría una vigencia que la fuente no
   declaró—, y mira el contexto: con el plazo ya agotado no toca nada;
2. resuelve el directorio (D8) y hace `os.Stat` de `world.db`: un directorio es inutilizable, sin tocar nada; ausente,
   paso 3; un fichero, paso 4;
3. **`world.db` ausente: se construye en un temporal y se publica de una vez.**
   1. comprueba contra un grafo vacío que cada extremo de arista es un nodo del lote (`ValidarContraGrafoVacio`): un
      rechazo no crea nada;
   2. crea los directorios que falten, de uno en uno desde el primer antecesor que no existe (`0700`), y anota los que
      crea ella (un `EEXIST` es de otra invocación y no se anota);
   3. `os.CreateTemp(<dir>, "world.db-nuevo-*")` (`0600`); si falla porque el directorio ya no existe (otra invocación
      que falló lo retiró vacío), vuelve a 3.2 mientras quede plazo, solo si el antecesor más cercano que existe
      (`os.Lstat`), él mismo incluido, es un directorio; si es un enlace sin destino o algo que no es un directorio,
      volver daría siempre lo mismo y la entrega falla enseguida («no se puede escribir world.db en …») sin crear nada;
   4. abre el temporal con la cadena de escritura del paso 4, fija `journal_mode=WAL` fuera de toda transacción y
      comprueba que el pragma devuelve `wal` —dentro de una transacción SQLite no lo fija, y sobre un fichero de 0
      bytes lo calla: devuelve `delete` sin error (V42)—; nadie más conoce el temporal, así que no espera a nadie; y en
      **una** transacción crea el esquema (migración 0001 y `schema_version`) y aplica
      el lote; confirma y cierra, y comprueba que el cierre no da error y que no queda `<temporal>-wal`: el temporal
      queda solo y completo (V36 C);
   5. mira el contexto: con el plazo agotado no publica;
   6. `os.Link(<temporal>, <dir>/world.db)`: si enlaza, `world.db` aparece de una vez, en WAL, con el esquema y el
      lote (FR-003, FR-013: nunca un `world.db` a medias); si falla con `fs.ErrExist`, otra invocación lo publicó antes
      y este lote se aplica por el paso 4 sobre ese `world.db`, con el mismo contexto; con cualquier otro error, la
      entrega falla (S7);
   7. borra siempre el temporal (y su `-wal`, `-shm` o `-journal`, si quedaran) y, si la entrega falla, retira los
      directorios que anotó en 3.2, del más profundo al menos, con `os.Remove`, que no borra uno que otra invocación
      ya use (`ENOTEMPTY`). Un fallo de esta limpieza se une a la causa (`errors.Join`) y sale en la misma línea de
      aviso (D7), sin silenciarse;
4. **`world.db` existe: se aplica en su sitio.** Antes de abrir SQLite comprueba, como D10, que el proceso puede
   escribir `world.db` (`os.OpenFile(<ruta>, os.O_RDWR, 0)` y cerrarlo enseguida, V46); con cualquier error —sin
   permiso, sistema de ficheros de solo lectura, o `fs.ErrNotExist` de un enlace simbólico sin destino o de un
   fichero retirado entre medias— la entrega falla con «no se puede escribir» (`inesperado`) sin abrir SQLite: no
   cambia ni crea nada, ni recupera nada. Un enlace sin destino llega aquí desde el paso 3, cuyo `os.Link` da
   `fs.ErrExist` porque el nombre está ocupado: así SQLite no crea el destino a través del enlace. Si `world.db` tiene
   **0 bytes**, la comprobación de 3.1 contra el grafo vacío (`ValidarContraGrafoVacio`) va aquí, **antes** de abrir
   SQLite —la versión 0 se sabe sin abrir—: un lote que el grafo vacío rechaza no toca ningún fichero, tampoco un
   `world.db-wal` junto a él, que la primera lectura de la conexión de escritura borraría (V49). Si puede, abre
   **sin** `journal_mode` en la cadena de conexión
   (`busy_timeout`, `synchronous(FULL)`, `foreign_keys(1)`, `_txlock=immediate`) y lee la versión: de un esquema
   posterior, o de un fichero que no es base, sale sin modificar nada (FR-010, FR-012) salvo la recuperación que se
   describe al final de este paso; si es 0, repite la
   comprobación de 3.1 contra el grafo vacío y, si la base no está ya en WAL, fija `PRAGMA journal_mode=WAL` con
   reintento por tramos, fuera de toda transacción, y comprueba que devuelve `wal` (V10, V11, V42). Si `world.db`
   tiene un `-wal` huérfano o un diario caliente, esta conexión los recupera como cualquier escritor de SQLite: el
   diario se deshace en su primera lectura (V43) y el `-wal` se lleva a `world.db` en el checkpoint del cierre si es
   la última conexión (V44); junto a un `world.db` de 0 bytes, en cambio, un `-wal` no vacío no se recupera: la
   primera lectura lo **descarta** —lo borra sin llevar nada a `world.db`— (V49);
5. abre la transacción inmediata con reintento por tramos, relee la versión **dentro**, crea el esquema si es 0
   (FR-013, atómico) y aplica el lote con las comprobaciones que necesitan la base (extremos presentes, tipo guardado,
   cuerpo de una huella ya guardada); cualquier rechazo deshace todo (V11: en WAL y sin `-wal` huérfano, los mismos
   bytes; con él, lo que dice la viñeta del `-wal` huérfano de abajo).

**Lo que queda tras una entrega que falla** (FR-033):

- **`world.db` ausente**: nada —ni `world.db`, ni el temporal, ni los directorios que creó la entrega—, falle donde
  falle (plazo, espera, rechazo, E/S u `os.Link`).
- **`world.db` que el proceso no puede escribir** (sin permiso, sistema de ficheros de solo lectura, enlace sin
  destino): nada; la entrega falla en la comprobación del paso 4, antes de abrir SQLite, así que no cambia ningún
  byte, no aparece ningún auxiliar y no se recupera ningún `-wal` huérfano ni ningún diario caliente (V46).
- **`world.db` de versión 1, o base ya en WAL sin esquema, que el proceso puede escribir, sin `-wal` huérfano**: los
  mismos bytes (V11); los auxiliares que abrió la conexión desaparecen al cerrarla si es la última (V12).
- **`world.db` que no es base, de esquema posterior o inutilizable**: los mismos bytes (FR-010, FR-012), salvo la
  recuperación de un `-wal` huérfano o de un diario caliente de una base de esquema posterior (viñeta siguiente, V44).
- **`world.db` con un `-wal` huérfano o un diario de rollback caliente** (estado «Con auxiliares» de data-model §3.1,
  que deja un escritor interrumpido): la conexión de la entrega los recupera (paso 4) aunque la entrega
  falle después —rechazo, esquema posterior, plazo o espera agotados, E/S—: el diario caliente se deshace en su
  primera lectura (`world.db` vuelve a lo que confirmó la última transacción y el diario desaparece, V43), y el `-wal`
  huérfano se lleva a `world.db` en el checkpoint del cierre, si es la última conexión (`world.db` cambia de bytes y
  de tamaño, y `-wal` y `-shm` desaparecen, V44). El contenido del grafo es el confirmado: el que ya leían los
  verbos de `graph` con el `-wal` huérfano (D10) y, con el diario caliente, el de la última transacción confirmada, que
  los verbos no podían leer (salían con 1, D10). Los bytes de `world.db` cambian: **desviación declarada** de «el
  grafo MUST quedar como estaba» en sus bytes, no en su contenido, en *Complexity Tracking* y en `gates/supuestos.md`.
  Con otra conexión abierta, el cierre de la entrega no es el último y no hay checkpoint. Junto a un `world.db` de
  **0 bytes**, un `-wal` no vacío no se recupera sino que se **descarta**: la primera lectura de la conexión de
  escritura lo borra sin llevar nada a `world.db` (`_pagerOpenWalIfPresent`, V49), aunque la entrega falle después
  (plazo, espera o E/S); es la misma desviación declarada, por su causa. Un lote que el grafo vacío rechaza no llega a
  abrir SQLite (paso 4) y no lo toca.
- **`world.db` existente, sin esquema y que no está en WAL** —un fichero de 0 bytes o una base en modo rollback sin el
  esquema del grafo; el binario nunca crea ninguno de los dos: solo llegan de fuera, escritos por cualquier programa
  con cualquier versión y configuración de SQLite—, si la entrega falla **después** de fijar WAL y antes de confirmar
  (plazo o espera propia agotados esperando la transacción inmediata, o un fallo de E/S). Se declara **por su causa y
  con sus cotas**, no por una lista de bytes, porque la lista depende de la base que llega:
  - **causa**: en una base en rollback, `PRAGMA journal_mode=WAL` es una confirmación en modo rollback, y SQLite hace
    en ella todo lo que hace al confirmar; no depende del lote ni de este binario, y es lo que haría cualquier
    programa que la pusiera en WAL;
  - **cotas**, que valen para cualquier base de fuera: el contenido —las tablas, las filas y el esquema de esa base—
    no cambia; `world.db` sigue en versión 0, sin el esquema del grafo, ahora en WAL; y, cerrada la última conexión,
    no queda ningún fichero que no estuviera antes (SQLite borra al cerrar el `-wal` y el `-shm` que abre, V12);
  - **qué cambia**, por esa causa: la cabecera de la página 1 —siempre las versiones de escritura y lectura (18 y 19,
    1 → 2), el contador de cambios (24-27) y `version-valid-for` (92-95), y los campos que la confirmación reescribe
    cuando no coinciden con lo que ella calcula: el tamaño en páginas (28-31), `SQLITE_VERSION_NUMBER` (96-99) y la
    lista de páginas libres (32-35 y 36-39)—; con `auto_vacuum=full` y páginas libres (lo que deja una aplicación que
    usa `sqlite3_autovacuum_pages`), la confirmación las vacía todas: reubica las páginas en uso que quedan detrás de
    ellas —cambian bytes de esas páginas, del mapa de punteros y de las páginas que las apuntan— y **trunca** el
    fichero (V45 a); un `world.db-journal` frío (`PERSIST`) o vacío (`TRUNCATE`) **desaparece** (V45 b; uno caliente
    es la viñeta anterior); y el de **0 bytes** pasa a 4096 bytes, la cabecera entera de una base en WAL sin tablas
    (V36 A, V41);
  - **ejemplos medidos**, que `aplicar_test.go` fija con su resultado exacto (V36 B, V41, V45), sin darlos por
    exhaustivos para otras bases: la cabecera de este controlador → 18, 19, 27, 95 y el mismo tamaño; escrita por otra
    versión → además 98, 99; tamaño en cabecera 0 → además 31; acarreo del contador → 18, 19, 26, 27, 94, 95;
    `auto_vacuum=full` con 40 páginas libres al final → de 176 128 a 12 288 bytes y 18, 19, 27, 31, 35, 39, 95; con
    una página en uso detrás de las libres → de 188 416 a 24 576 bytes, esos siete de la cabecera y otros en las
    páginas 2, 4 y 5; con el diario frío de `PERSIST` o el vacío de `TRUNCATE` → el diario ya no existe y 18, 19, 27,
    95; 0 bytes con un diario vacío → 4096 bytes y sin diario.

  En todos los casos `world.db` sigue sin esquema (versión 0): los verbos de `graph` lo leen como grafo vacío (FR-004)
  y la entrega siguiente lo completa (FR-013). Es la **desviación declarada** de «el grafo MUST quedar como estaba» en
  sus bytes (no en su contenido), en *Complexity Tracking* y en `gates/supuestos.md`. Un rechazo del lote no llega a
  ese punto: contra un grafo vacío, todo motivo de rechazo se comprueba antes de fijar WAL (paso 4), salvo que otra
  invocación entregue a la vez, y es esa la que cambia los bytes.
- **Proceso terminado por una señal** durante la entrega: no es un fallo de entrega que el programa trate (FR-033
  enumera los que sí, y el binario no atiende señales, V39), así que no se ejecuta ninguna limpieza: pueden quedar el
  temporal `world.db-nuevo-*` (con sus auxiliares) y los directorios creados, además de los ficheros de recuperación
  que SQLite deje ante la interrupción (un `-wal` huérfano o un diario caliente), que la conexión de escritura
  siguiente recupera (viñeta del `-wal` huérfano, V43, V44). Ningún lector mira el temporal y ninguna entrega lo
  toca. Borrarlo en una entrega posterior podría borrar el de otra invocación en curso, y atender señales no lo pide
  el hito: lectura conservadora, en `gates/supuestos.md`.

**Alternativas.** *WAL en la cadena de conexión, como la caché*: un `world.db` de 0 bytes pasaría a 4096 bytes aunque
el lote se rechace (V10), contra FR-033. *Fijar WAL dentro de la transacción que crea el esquema, para que un fallo no
deje nada*: SQLite no lo permite (V42): sobre una base en rollback el pragma falla, y sobre un fichero de 0 bytes
devuelve `delete` sin error y la base confirmada queda fuera de WAL, contra FR-003; por eso los pasos 3.4 y 4 exigen
que el pragma devuelva `wal`. *Evitar la recuperación de un `-wal` huérfano o de un diario caliente cuando la entrega
va a fallar*: SQLite la hace al abrir o al cerrar la conexión de escritura, antes de saber si el lote entra; evitarla
exigiría leer antes con `mode=ro` y abrir después para escribir, lo que solo la ahorra en los fallos que esa lectura
ve (esquema posterior, rechazo contra lo guardado), no en un plazo o una espera agotados ni con un diario caliente,
que `mode=ro` no deja leer (V43); a cambio de una segunda apertura con su propia carrera entre las dos. Lo que deja
la recuperación es lo que dejaría la primera entrega que sí entra, o cualquier otro escritor de SQLite, y no cambia el
contenido. *Desactivar el checkpoint al cerrar*: ver D10. *Crear `world.db` en su sitio y fijar WAL antes de la
transacción* (la primera versión de este plan): un plazo agotado esperando la transacción inmediata dejaría un `world.db` de 4096 bytes sin
esquema, y el directorio, donde no había nada (V36 A), contra FR-033. *Borrar el `world.db` creado si la entrega
falla*: con varias invocaciones a la vez borraría un fichero que otra ya usa. *Publicar el temporal con `os.Rename`*:
sustituye un `world.db` que otra invocación haya creado entre medias (en Unix y en Windows), y lo que esa invocación
entregue iría a un fichero ya desenlazado. *Publicar también por un temporal cuando `world.db` ya existe sin esquema*:
exigiría sustituirlo con `os.Rename`, con la misma pérdida para quien lo tenga abierto. *Con un `world.db` existente
sin esquema, crear el esquema en modo rollback y fijar WAL después de confirmar*: una base de versión 1 que no está en
WAL hasta la entrega siguiente (contra FR-003), y el binario escribiría el lote del grafo con un diario de rollback,
que una interrupción dejaría caliente sobre datos del grafo, y los verbos de `graph` no pueden leer un diario caliente
sin deshacerlo (D10, V43); todo a cambio de un residuo que no cambia el contenido.
*Migrar en una transacción y aplicar en otra*: un lote rechazado dejaría el esquema creado en un fichero que estaba sin
él. *Si `os.Link` no es posible, crear en su sitio*: devolvería el residuo que este diseño elimina, por un camino que
ningún test de CI ejerce (S7). *Abrir sin comprobar el permiso de escritura* (la versión anterior de este plan): sobre
un `world.db` en WAL que el proceso no puede escribir, SQLite cae a solo lectura sin avisar, lee la versión y abre
`BEGIN IMMEDIATE` sin error, y al cerrar deja `world.db-shm` y `world.db-wal` (V46), contra FR-033. *Evitar el
vaciado de las páginas libres al fijar WAL*: exige registrar una función con `sqlite3_autovacuum_pages` por la API C
del controlador (`Xsqlite3_autovacuum_pages`, V45), que ninguna verificación de este plan cubre, a cambio de un
residuo que no cambia el contenido. *No fijar WAL en una base de fuera con páginas libres o con un diario*: dejaría
ese `world.db` fuera de WAL o sin el esquema del grafo, contra FR-003 y FR-013. *Enumerar los bytes como el conjunto
exacto de cualquier base de fuera* (la versión anterior de este plan): la lista depende de la configuración con que
llega la base (V45), y afirmarla para cualquiera no es comprobable; lo comprobable son las cotas y los ejemplos.

### D12 · Esquema de `world.db` (migración 0001)

**Decisión.** `nodes`, `edges`, `texts` y `schema_version`, todas `STRICT`, con los nombres de columna de
`refs/kitlegal-grafo.md` §3 adaptados a FR-002 y FR-023 (data-model §3): la arista tiene **una fila por terna**
(`PRIMARY KEY (src, rel, dst)`) con su primera y su última observación; cada nodo y cada arista guardan la fuente,
la url y la vigencia de la última (`ttl`, NULL si no se declaró) y la url y la fuente de la primera, que el desempate
exacto de FR-023 necesita; `props` es el JSON canónico (RFC 8785) de los datos identificativos; `REFERENCES` con
`foreign_keys(1)` es la segunda red de FR-024 (V10). Índices: `nodes(type, source)` para `stats`, `edges(dst, rel)`
para las aristas entrantes.

**Alternativa.** La de `refs/` con `observed_at` en la clave de la arista: una fila por observación, contra «una
arista por terna» (FR-022).

### D13 · FR-023 como un orden total, en el dominio

**Decisión.** Cada observación es una tupla (instante, url, fuente, texto de la fecha, vigencia, datos canónicos). La
**última** guardada es la de mayor instante y, entre instantes iguales, la **menor** por (url, fuente, texto de la
fecha, sin vigencia antes que con ella y la menor, datos canónicos), comparando bytes; la **primera** es la de menor
instante y, a igualdad, la menor por (url, fuente, texto de la fecha). Como solo se guardan esas dos y la de un texto
(la más antigua con el mismo orden), el resultado no depende del orden de llegada ni de repetir un lote (FR-022). Las
funciones son puras en `internal/core/grafo` y se prueban en los dos órdenes de llegada. Quedaron así: un registro por
tabla (`RegistroDeNodo`, `RegistroDeArista`, `RegistroDeTexto`, con los campos de las columnas de data-model §3) y
`FusionarNodo`, `FusionarArista` y `FusionarTexto(guardado, llegado)`, que devuelven el registro fusionado —el guardado
si nada cambia, para que el adaptador escriba solo si difiere— y rechazan con un `*Rechazo` un tipo o un cuerpo
distintos del guardado; `Consolidar` reduce el lote a un registro por clave, ordenado por la clave.

**Alternativa.** Guardar solo la fecha de la primera: el desempate por url no se podría reproducir y dependería del
orden de llegada.

### D14 · Validación del lote y rechazo de `Persona`

**Decisión.** `grafo.ValidarLote` rechaza el lote entero —sin tocar el disco— si falta fuente, url o fecha, si la url
no es un URI absoluto (el mismo criterio que `schema.Procedencia.Validar`), si la fecha no es RFC 3339, si la vigencia
es negativa, si un nodo no tiene id o tipo, si una arista no tiene origen, relación o destino, si un texto no tiene la
huella `sha256:` + SHA-256 de su cuerpo, si el lote da dos tipos a un id, o si un `Persona` lleva en su id o en
cualquier cadena de sus datos (claves y valores, a cualquier profundidad) algo con la forma de DNI, NIE o NIF. La
expresión (RE2, V6) es una sola, **sin** la bandera `(?i)`: las mayúsculas y minúsculas van en clases ASCII explícitas,
con las clases de caracteres de D34 y los separadores y los límites de FR-025 (contracts/almacen-world-db.md §5). Lo que necesita la base (extremos, tipo guardado, huella guardada con otro cuerpo) se
comprueba dentro de la transacción (D11). También rechaza una operación nula o un puntero a `Nodo`, `Arista` o
`Texto` (la interfaz sellada los admite), y `Consolidar` rechaza unos datos sin forma JSON canónica; las cadenas de una
`Persona` se examinan en esa forma canónica, y una `Persona` cuyos datos no la tienen se rechaza.

### D15 · `check` sobre una instantánea, en el dominio

**Decisión.** `internal/graph` lee todos los nodos y aristas en una transacción de lectura (`Lectura.Instantanea`) y
`grafo.Comprobar(instantanea, ahora) ([]Hallazgo, error)` aplica las dos reglas y ordena los hallazgos; sin hallazgos
devuelve una lista vacía, nunca nula, y lo que ninguna entrega guarda (una fecha que no es RFC 3339, una vigencia
negativa o con fracción de segundo) es un error sin hallazgos, que el applet da como `inesperado`. Diez mil nodos y
diez mil aristas caben de sobra en memoria (SC-008).

**Alternativa.** Consultas SQL con `json_extract`: reglas de dominio escritas en SQL, peor probadas y fuera del umbral
de `internal/core/**`.

### D16 · `stats` por `GROUP BY` y orden en Go

**Decisión.** Recuentos con `SELECT type, source, count(*) … GROUP BY type, source` (y lo mismo en aristas); el orden
por bytes de FR-054 lo pone Go con `strings.Compare`, sin depender de la colación del motor.

### D17 · Datos canónicos con `jsontext.Canonicalize`

**Decisión.** Los datos identificativos se guardan y se comparan en su forma JCS (V5), que es la que FR-023 nombra.

**Alternativa.** La forma canónica de `schema.Huella`: ordena claves por bytes y escapa como `encoding/json`, que no
es RFC 8785.

### D18 · El instante de caducidad conserva el desplazamiento

**Decisión.** `caducó = fecha_consulta + vigencia`, calculado en `time.FixedZone("", desplazamiento)` de la propia
`fecha_consulta` y escrito con `RFC3339Nano` (V2, V3): `Z` si la fecha lleva `Z`, sin fracción si es cero.

### D19 · El applet `graph`

**Decisión.** Tres verbos sin verbo por omisión: `show <id>`, `stats` y `check`, con los datos de
contracts/applet-graph.md. Firma en el espacio reservado —`kitlegal.graph`, `kitlegal:applet/graph`— con la fecha de
su reloj, que lee **una vez** por invocación y que es también el instante de la comprobación (FR-051, FR-067): así el
sobre de `check` dice cuándo se comprobó, y un test o un binario de e2e con el reloj fijo da una salida reproducible.
Sin `--json`, la **tabla mínima** del kernel (FR-055), que por construcción no dice nada que no diga el sobre. Esquema
publicado `schemas/grafo.json` (entidad `grafo`, verbos `check`, `show`, `stats`), como los demás. Con `--dry-run`
los verbos leen igual (no tienen efectos) y el kernel escribe su descripción, como `territorio`.

**Alternativa.** Texto legible propio (ADR 0026): más código para lo mismo que ya dice la tabla.

### D20 · Emisión de `boe` en el adaptador

**Decisión.** `internal/source/boe/grafo.go` compone el `Observado` de `articulo` y de `articulos` a partir de los
`Articulo` ya compuestos (FR-040), con la vigencia de la consulta (`vigenciaLarga`); `articulo.go` lo pone en el
resultado de éxito. El id de la norma sale de `url_eli` con `net/url`: desde el primer segmento de la ruta que es
exactamente `eli` hasta el final, con los segmentos de la ruta tal como va escrita (`EscapedPath`: una `%2F` no parte
un segmento); sin él no se emite nada (FR-040). `indice`, `buscar`, `metadatos` y `analisis` no
emiten.

**Alternativa.** Componerlo en `internal/app/boe.go`: sacaría de la fuente el conocimiento de sus propios
identificadores (ADR 0014, «un applet diseñado sin el grafo es un applet peor»).

### D21 · Emisión de `territorio` en el dominio

**Decisión.** `territorio.Territorio.Observado()` (fichero nuevo `internal/core/territorio/grafo.go`) da el
`Municipio` `ine:<código>` con datos `{codigo_ine, nombre}` y, si hay DIR3, el `Organo` `<DIR3>` con datos `{dir3}` y
la arista `lb:pertenece_a`; sin vigencia. `internal/app/territorio.go` lo pone en el resultado. FR-044 (sin DIR3) se
prueba con fuentes sintéticas del dominio, porque ningún municipio real carece de DIR3 (V28).

### D22 · Fixtures derivadas, con marca de sintéticas y un control de derivación

**Decisión.** Cuatro grabaciones derivadas de las de H4, cada una con el nombre de la que sustituye, por tareas
`[datos]` (FR-095):

- `internal/app/testdata/derivadas/version-posterior/…_texto_bloque_a21.json`: el bloque `a21` con `fecha_vigencia`
  `20250101` y un párrafo más, `[Redacción sintética de prueba: versión posterior derivada de la grabación de H4.]`;
- `internal/app/testdata/derivadas/sin-eli/…_metadatos.json`: los metadatos de la Ley 39/2015 con `url_eli` vacío;
- `internal/app/testdata/derivadas/eli-sin-segmento/…_metadatos.json`: con `url_eli`
  `https://www.boe.es/buscar/act.php?id=BOE-A-2015-10565`;
- `testdata/evals/grafo-previo/lpac-a21-version-anterior/…_texto_bloque_a21.json`: el bloque con `fecha_vigencia`
  `20151002` y el párrafo `[Redacción sintética de prueba: versión anterior derivada de la grabación de H4.]`.

El párrafo marca el texto como sintético: ninguna fixture inventa contenido legal. `TestGrabacionesDerivadas`
(`internal/app`) comprueba que cada una se lee con `boe` y da el mismo `Articulo` o los mismos metadatos que la
grabación original salvo exactamente lo que dice su nombre.

**Alternativas.** Escribirlas dentro de los guiones: FR-095 exige tareas `[datos]`. Generarlas en tiempo de test: no
serían ficheros revisables por la persona.

### D23 · Arnés: tres relojes fijos, derivadas en `$WORK` y comparación con `cksum`

**Decisión.** El `package main` de e2e gana la variable de cadena `reloj` (`-X main.reloj=<RFC 3339>`): si lleva
valor, la reproducción de `boe` fecha con `httpx.ConHora` en ese instante y el applet `graph` usa ese reloj. `TestMain`
construye tres binarios más —`$KITLEGAL_T0_BIN` (2026-09-28T12:00:00Z), `$KITLEGAL_T1_BIN` (2026-09-29T12:00:00Z) y
`$KITLEGAL_T8_BIN` (2026-10-06T12:00:00Z)— y el `Setup` copia `internal/app/testdata/derivadas/` en `$WORK/derivadas/`.
Los guiones comparan los bytes de `world.db` con `exec cksum cache/world.db` y quitan variables con `exec env -u`
(V8, V32). Un valor de `reloj` que no es RFC 3339 es un defecto de composición (`registroDeE2E` falla).

**Alternativas.** *Una variable de entorno para el reloj*: el binario de e2e no lee del entorno nada que decida su
comportamiento (V22). *El reloj real*: `fecha_consulta` no se podría afirmar literal y `fuente-caducada` dependería del
día en que se ejecuten los tests.

### D24 · Once guiones de aceptación, con la precondición del applet

**Decisión.** Los once de la tabla «Aceptación e2e» del plan, cada uno encabezado por la precondición `exec kitlegal
--help` + `stdout '^  graph +\S'`, que falla por aserción mientras el applet no esté registrado (V23).

### D25 · La eval de FR-085 no va en T001

**Decisión.** T001 escribe solo los guiones. La eval nueva exige el formato ampliado (FR-086) y `make ci` valida cada
eval contra el esquema (V24), así que en T001 dejaría `make ci` en rojo (V23). Entra en su propia tarea, **después**
del formato y **antes** de tocar `SKILL.md` (Definition of Done §1.10).

### D26 · Formato de eval ampliado

**Decisión.** (contracts/evals-y-skill.md) Tres añadidos, todos opcionales y solo con `activa: true`: la forma de
comando `comando-comprobacion` (`applet` + `verbo: check`); `prohibidos`, comandos (`applet` + `verbo`) que ninguna
invocación de la sesión puede ejecutar —cuenta toda invocación que consulta, termine como termine—; y `grafo_previo`,
un conjunto de grabaciones derivadas (`testdata/evals/grafo-previo/<nombre>/`) y los comandos de bloque con los que se
preparó el grafo antes de la sesión. `Juzgar` gana `comandos_prohibidos_ejecutados` y un motivo por cada uno; la
preparación ejecuta el grafo previo **antes** de llenar la caché, con una caché temporal que se descarta y el grafo en
el directorio de la caché de la sesión, y trata como falta cualquier salida de error (una entrega fallida). `avisos` no
cambia (FR-086).

### D27 · `boe-legislacion` declara el applet `graph` entero

**Decisión.** `kitlegal-applets: boe graph`: la tabla de comandos gana `graph show`, `graph stats` y `graph check`
(V26); el protocolo solo pide `graph check` y prohíbe citar de cualquier verbo de `graph` (FR-081).

**Alternativa.** Seleccionar verbos en el frontmatter: sintaxis nueva que el hito no pide.

### D28 · Costes medidos solos

**Decisión.** `TestCosteDelGrafo` (`internal/app`, no paralelo) mide SC-007 (mediana de 20 `boe articulo` desde la
caché con y sin `--no-graph`, con el binario de e2e) y SC-008 (`graph check` y `graph stats` sobre un `world.db` de
10 000 nodos y 10 000 aristas, mediana de 5). Entra en `MEDIDAS_DE_TIEMPO`, que `test-tiempos` ejecuta sola (V31).

**Alternativa.** Medirlo dentro de `make test`: en paralelo con todo el módulo mide la carga (la razón de
`test-tiempos`).

### D29 · El guion de la matriz territorial deja de afirmar que no hay directorio de caché

**Decisión.** `territorio-matriz.txtar:206` pasa de `! exists cache` a `! exists cache/cache.db`: `territorio` sigue
sin tocar la caché, pero desde H7 el kernel deja `world.db` en su directorio (V27). Va en la tarea `[datos]` que
registra la entrega en el binario de e2e, que es la que lo pone en rojo.

### D30 · Regla de arquitectura del grafo

**Decisión.** `.golangci.yml` gana la lista `grafo` (`files: **/internal/graph/**`, deniega `internal/source` e
`internal/render`) y `TestArquitectura` la subprueba «R6 · internal/graph no importa internal/source ni
internal/render», transitiva con `grafo.alcanza` (V29) y con la exigencia de que `internal/graph` esté en el grafo. R1
y R3 no cambian; el comentario de R3 se corrige (el grafo llega en H7; `duenoObligatorio` sigue falso porque falta
`internal/store`, H10).

### D31 · `misspell`

**Decisión.** `observacion` entra en `misspell.ignore-rules`, con su comentario, en la tarea que crea las claves JSON
`primera_observacion` y `ultima_observacion` (el `_` no protege, V30). Ningún identificador suelto se llama
`Observacion` ni `Emision`; los comentarios llevan tilde.

### D32 · Datos externos: ninguno

**Decisión.** H7 no graba ni deriva nada de ninguna fuente: no hay manifiesto `grabaciones.json` nuevo, ni test
`//go:build grabacion`, ni paso `grabar_datos`, ni material en `evidencias/h7/`, ni fila nueva en `docs/SOURCES.md`.
Las cuatro fixtures derivadas (D22) salen de grabaciones de H4 ya versionadas; `world.db` es un almacén local.

### D33 · Pruebas del almacén: unitarias y de integración

**Decisión.** `internal/graph` tiene tests unitarios sobre `t.TempDir()` (sin etiqueta) y `integracion_test.go` con
`//go:build integration` para la matriz de FR-088 por la API pública (lo ejecuta `make test-integration`, dentro de
`make ci`), como la caché; los casos de un `world.db` que es un enlace simbólico van en `integracion_enlace_test.go`,
con `//go:build integration && unix`, porque el nombre de sus auxiliares es el de Unix (V47; precedente de la
restricción: `internal/disco/noregular_unix_test.go`), sin `t.Skip`.

### D34 · Clases de caracteres de FR-025 y de FR-064: ASCII

**Decisión.** En FR-025, «letra» es `A`-`Z` o `a`-`z`; «cifra», `0`-`9`; y el espacio que hace de separador, U+0020;
«sin distinguir mayúsculas de minúsculas» vale solo para esas letras. Cualquier otro carácter —`á`, `í`, `ñ`, `Ñ`,
`ſ` (U+017F), el signo Kelvin (U+212A), U+00A0, las cifras de anchura completa— no forma parte de ninguna de las tres
formas y sí sirve de límite. La expresión se escribe sin `(?i)` y con clases explícitas: `[A-Za-z]`, `[0-9]`,
`[XYZxyz]`, `[0-9A-Za-z]` y, para los límites, `[^A-Za-z0-9]` (V6, V37). Consecuencias, que fija
`TestPersonaSinDocumento`: rechazan `ſ12345678Z`, `Martí12345678Z` y `12345678Zá` (el carácter junto a la secuencia no
es letra ni cifra ASCII, así que es límite); entran `12345678Ñ` y `12345678` + U+212A (la última no es una letra
ASCII), `12` + U+00A0 + `345 678 Z` (U+00A0 no es separador) y `１２３４５６７８Z` (no son cifras ASCII). En FR-064,
las «ocho cifras AAAAMMDD» de una `fecha_vigencia` son también cifras ASCII: es válida si y solo si
`time.Parse("20060102", v)` no da error, que ya rechaza un signo, cifras de otra escritura, otra longitud y una fecha
que no existe (V37).

Es la lectura literal de «cifra ASCII» de FR-025, la del alfabeto de los documentos de identidad españoles (letras y
cifras ASCII) y, en los límites, la que más rechaza (constitución VII: ante la duda sobre un dato de una persona,
prevalece el rechazo). Es la que proponía el motivo que el gate del spec dejó pendiente (`gates/spec-pendiente.md`) y
queda como supuesto en `gates/supuestos.md`.

**Alternativas.** *Clases Unicode* (`\p{L}`, `\p{Nd}` y cualquier espacio como separador): una letra acentuada junto
al documento dejaría de ser límite y `Martí12345678Z` o `12345678Zá` entrarían, contra la lectura conservadora; y
casarían formas que ningún documento español tiene, como `12345678Ñ`. *`(?i)` con clases en minúscula* (la primera
versión de este plan): el plegado de RE2 es Unicode (V6), así que mezclaba las dos lecturas —`ſ` dejaba de ser límite
y el signo Kelvin pasaba por letra— y el contrato decía otra cosa que la expresión.

### D35 · `graph show`: id vacío, de espacios o con caracteres de control (FR-052)

**Decisión.** `grafo.ValidarID(id)` da un error de clase `argumentos` (2) si el id es la cadena vacía, si **todos**
sus caracteres son de espacio en blanco —la propiedad White_Space de Unicode, `unicode.IsSpace`: U+0020, `\t`, `\n`,
U+0085, U+00A0, U+2003…— o si contiene **algún** carácter de control —la categoría Cc de Unicode,
`unicode.IsControl`: U+0000-U+001F y U+007F-U+009F— (V37). Un id con espacios y algo más (`a b`, ` a`) es válido y se
busca tal cual, sin recortar; si no está, 3 (FR-053). Un carácter de formato como U+200B no es de espacio en blanco ni
de control, y un byte que no es UTF-8 se lee como U+FFFD, que tampoco: los dos son ids válidos que, si no están, dan 3.
Ejemplos que fija `TestValidarID`, uno por clase: U+00A0 solo y U+2003 solo (espacio en blanco → 2), `a` + U+0000 +
`b` y U+007F solo (control → 2), U+0085 solo (de las dos clases → 2), y `a b` y U+200B (válidos). El guion
`h7-grafo-codigos` ejerce los mismos salvo U+0000, que no puede viajar en un argumento (V40).

Es la definición que da el mismo resultado se escriba como se escriba el hueco, la que coincide con «caracteres de
control» en Unicode (Cc) y la que rechaza como argumento todo id que no puede ser el id natural de ningún nodo (un
ELI, `ine:<código>` o un DIR3 no llevan espacios en blanco ni controles). Queda como supuesto en `gates/supuestos.md`
(el otro motivo que el gate del spec dejó pendiente).

**Alternativas.** *Solo U+0020 como espacio y solo C0 (U+0000-U+001F) como control*: un id de U+00A0 o de U+2003
saldría con 3 en vez de 2, y U+007F y U+0085 no serían control, aunque son tan invisibles como el espacio y el
tabulador. *Toda la categoría C de Unicode* (`unicode.Is(unicode.C, r)`, con los formatos como U+200B y los
sustitutos): va más allá de «control» (Cc) sin que el spec lo pida (lo no especificado no se implementa). *Recortar los
espacios del id antes de buscarlo*: cambiaría el id que se busca, y FR-053 lo nombra tal cual.

## S · Supuestos (no se afirman como hechos)

| # | Supuesto | Por qué no se puede comprobar ahora | Qué pasa si no se cumple |
|---|---|---|---|
| S1 | La entrega que no cambia nada (una consulta servida de la caché) cuesta en el ejecutor de CI bastante menos de 150 ms (SC-007) | Solo se mide en el runner | `TestCosteDelGrafo` falla en `test-tiempos` y la tarea optimiza (p. ej., decidir sin transacción de escritura cuando nada cambia); nunca se relaja la cota |
| S2 | `graph check` y `graph stats` sobre 10 000 nodos y 10 000 aristas cumplen 3 s y 1 s en el runner (SC-008) | Ídem | Ídem |
| S3 | `cksum` y `env -u` están en el runner de Linux (POSIX; `coreutils`) | Comprobado solo en macOS (V32) | Los guiones fallarían por orden no encontrada; se ve en el primer `make ci` de CI |
| S4 | La eval informativa se ejecuta en el job con su tasa (SC-012) | Lo mide el cierre del workflow, no una tarea | Queda en el informe final |
| S5 | En Windows el grafo se comporta como en Unix (rutas, permisos, `os.Link` sobre NTFS, `cksum` no se usa allí porque los guiones de `cksum` y `env -u` llevan `[unix]`), con dos diferencias que D10 ya trata así: `os.OpenFile` con `O_RDWR` falla sobre un fichero con el atributo de solo lectura, y SQLite nombra los auxiliares de un `world.db` que es un enlace por la ruta sin resolver (V47, leído en el fuente de `modernc.org/sqlite`, sin ejecutar) | Sin CI en Windows | Sin efecto en `make ci` |
| S6 | Desde el paso 12, cada invocación de los guiones cronometrados existentes entrega al grafo (abrir `world.db`, transacción inmediata, `synchronous(FULL)`), y aun así cabe en su cota: los diez `cronometra 200ms` de `boe articulo` desde la caché de `internal/app/testdata/script/boe-cache-rapida.txtar` y los tres de `territorio resolver` de `territorio-matriz.txtar:192-202`, que ejecuta `TestMedidasDeTiempo` en `make test-tiempos` (V38). En esas invocaciones la entrega no cambia nada (misma observación que la primera del guion). S1 solo cubre los 150 ms de SC-007, que, sumados al coste actual de la invocación, podrían pasar de 200 ms: cumplir SC-007 no basta para cumplir estas cotas | Solo se mide en el runner, más lento que la máquina de desarrollo | `TestMedidasDeTiempo` falla en `test-tiempos`, y la tarea optimiza en `internal/graph` la entrega que no cambia nada —p. ej., decidir con una transacción de lectura, antes de `BEGIN IMMEDIATE`, que el lote consolidado coincide con lo guardado y no abrir la de escritura— **sin tocar esos guiones ni sus cotas**; nunca se relaja una cota ni se saca un guion de `test-tiempos` |
| S7 | El directorio de la caché está en un sistema de ficheros con enlaces duros (APFS, ext4, NTFS…), que `os.Link` necesita para publicar `world.db` (D11) | Comprobado en APFS (V36 C); el del runner se ve en su primer `make ci` | `os.Link` falla con un error distinto de `fs.ErrExist`: la entrega falla con la línea de aviso, no crea nada y la salida y el código del applet no cambian (FR-033); los verbos de `graph` leen un grafo ausente. En el runner, los guiones que esperan `world.db` fallarían en el primer `make ci` |
