# Quickstart: validación de H3

Guía **ejecutable** para comprobar que H3 entrega lo que dice. Cada escenario se ejecuta tal cual, desde
la raíz del repositorio, sobre la rama `h3-internal-cache-sqlite` y **con el hito ya implementado**: no
es la implementación, es cómo se verifica. Los nombres de test son los del inventario de
[plan.md](./plan.md) («Inventario de tests»).

**Sin efectos colaterales.** Ningún escenario crea ni modifica un fichero versionado ni toca la caché
real de la cuenta (`~/.cache/kitlegal/`): todos los tests trabajan en directorios temporales que Go
borra al terminar, y el escenario 1 lo mide. Los que necesitan introducir violaciones deliberadas
(escenario 9) lo hacen sobre una **copia desechable del árbol fuera del repositorio** (`mktemp -d`, sin
`.git`) que borran al terminar. Lo único que queda en el árbol tras los escenarios es `bin/` y
`coverage.out`, ambos en `.gitignore` desde H0. Ningún escenario toca el índice ni el historial de git.

## Prerrequisitos

```bash
go version                          # go1.27.1 (la directiva `toolchain` de go.mod)
git --version
make --version
git rev-parse --abbrev-ref HEAD     # h3-internal-cache-sqlite
git status --porcelain | grep -v '^?? specs/' ; true   # vacío: nada versionado modificado
```

Los bloques se ejecutan con **`bash` o `zsh`**. Fuera de `go`, `git` y `make`, solo herramientas POSIX
(`grep`, `sed`, `diff`, `mktemp`, `tar`, `wc`, `find`, `shasum`, `test`). **Ninguna comprobación necesita
red**, salvo la primera ejecución de `make ci` con la caché fría (descarga de módulos y compilación de
herramientas) y `make vuln`, que consulta la base de vulnerabilidades.

Varios escenarios filtran con `grep` la salida **en crudo** de `go test` o de `git`. Si el terminal
lleva un envoltorio que la resume (por ejemplo `rtk`, que sustituye las líneas `--- PASS` por un
recuento y reescribe `git status --porcelain`), ejecutar la orden sin envoltorio (`rtk proxy go test …`,
`rtk proxy git status …`) o leer el registro íntegro que ese envoltorio guarda; el resultado esperado no
cambia.

Los tests del hito **no conocen ninguna dirección de una máquina real**: la única dirección que
aparece es el host ficticio `fuente.prueba` del fixture de reproducción y de las claves de caché, más la
dirección de la identificación que H2 graba en la cabecera `User-Agent` del fixture, que se envía y no
se consulta. Se comprueba mecánicamente:

```bash
grep -rhoE 'https?://[^"'"'"'`) ]+' internal/cache/*_test.go internal/cache/testdata \
  | sed -E 's/:+$//' \
  | grep -viE '^https?://fuente\.prueba(/|$)' \
  | grep -vx 'https://ventanillalegal.es/bot' ; test $? -eq 1 && echo "solo direcciones ficticias"
```

**Esperado**: «solo direcciones ficticias» y ninguna línea antes.

---

## Escenario 1 — La suite entera, sin red, con detector de carreras y sin tocar la caché real (SC-009, SC-014, FR-040)

```bash
ANTES=$(ls -laR "$HOME/.cache/kitlegal" 2>&1 | shasum)
make test
DESPUES=$(ls -laR "$HOME/.cache/kitlegal" 2>&1 | shasum)
test "$ANTES" = "$DESPUES" && echo "la caché real de la cuenta no se ha tocado"
rtk proxy git status --porcelain -- testdata internal/cache/testdata internal/source 2>/dev/null; echo "fin del estado"
```

**Esperado**: `ok  	github.com/jmorenobl/kitlegal/internal/cache` (y el resto de paquetes) sin
`FAIL` ni `WARNING: DATA RACE`; «la caché real de la cuenta no se ha tocado»; entre la orden de `git
status` y «fin del estado» **no aparece ninguna línea** (ningún fichero nuevo ni modificado bajo
`testdata/`, `internal/cache/testdata/` ni `internal/source/`: el único fixture del hito ya está
versionado y ningún test escribe ahí).

---

## Escenario 2 — El criterio de aceptación: la segunda consulta idéntica no toca la red (SC-001, US1)

```bash
rtk proxy go test -race -count=1 -v \
  -run 'TestAdaptadorDePruebaConElKernel/(primera-consulta|segunda-consulta-sin-red|sin-cache-la-reproduccion-falla)$' \
  ./internal/cache/ | grep -E '^\s*--- (PASS|FAIL): TestAdaptadorDePruebaConElKernel/'
```

El filtro exige la barra del subtest: con `-v`, `go test` imprime también la línea del test padre
(`--- PASS: TestAdaptadorDePruebaConElKernel (…)`), que aquí no se cuenta.

**Esperado**: tres líneas `--- PASS`, una por subtest:

```
    --- PASS: TestAdaptadorDePruebaConElKernel/primera-consulta (…)
    --- PASS: TestAdaptadorDePruebaConElKernel/segunda-consulta-sin-red (…)
    --- PASS: TestAdaptadorDePruebaConElKernel/sin-cache-la-reproduccion-falla (…)
```

La primera consulta se sirve desde la grabación de `GET http://fuente.prueba/norma` y deja la entrada
en la caché (`origen: fuente`); la segunda se ejecuta con un cliente de reproducción **estricto** sobre
un directorio vacío y termina con código 0 y `origen: cache` con el mismo cuerpo byte a byte; la tercera
demuestra que ese cliente estricto falla de verdad cuando la caché no sirve: código 1 y el mensaje
nombra `GET http://fuente.prueba/norma`.

---

## Escenario 3 — Lo caducado no se sirve, con reloj inyectado y sin esperar (SC-002, US2)

```bash
rtk proxy go test -race -count=1 -v -run '^TestExpiracionConRelojInyectado$' ./internal/cache/ \
  | grep -E '^--- PASS: TestExpiracionConRelojInyectado \(0\.[0-9]+s\)$' && echo "sin espera real"
```

**Esperado**: la línea `--- PASS: TestExpiracionConRelojInyectado (0.0Ns)` y «sin espera real». El test
guarda con una vigencia de una hora y lee en las tres posiciones del borde (antes, en el instante
exacto y después) moviendo el reloj inyectado; su duración es de milisegundos.

---

## Escenario 4 — `--offline` por el kernel: sirve lo guardado, código 4 si falta y el fichero queda intacto (SC-003, SC-011, US3)

```bash
rtk proxy go test -race -count=1 -v -run 'TestAdaptadorDePruebaConElKernel/offline-' ./internal/cache/ \
  | grep -E '^\s*--- (PASS|FAIL): TestAdaptadorDePruebaConElKernel/offline-'
```

El filtro repite el prefijo `offline-` tras la barra del subtest: con `-v`, `go test` imprime también la
línea del test padre (`--- PASS: TestAdaptadorDePruebaConElKernel (…)`), que aquí no se cuenta. Los
subtests cuyo nombre empieza por `offline-` son **exactamente seis** (inventario de tests de `plan.md` y
research D12); ninguno más puede llamarse así sin cambiar este esperado.

**Esperado**: seis líneas `--- PASS`, una por subtest:

```
    --- PASS: TestAdaptadorDePruebaConElKernel/offline-presente (…)
    --- PASS: TestAdaptadorDePruebaConElKernel/offline-ausente (…)
    --- PASS: TestAdaptadorDePruebaConElKernel/offline-expirada (…)
    --- PASS: TestAdaptadorDePruebaConElKernel/offline-sin-base (…)
    --- PASS: TestAdaptadorDePruebaConElKernel/offline-directorio-inexistente (…)
    --- PASS: TestAdaptadorDePruebaConElKernel/offline-guardar (…)
```

Cada subtest invoca `app.Main` en proceso con `--offline` sobre un registro que solo existe en el test:
presente → 0 y `origen: cache`; ausente y expirada → 4; directorio vacío o inexistente → 4 y sigue vacío
o inexistente (ni `cache.db`, ni `-wal`, ni `-shm`, ni migración); `guardar` → 1. La variante del
directorio inexistente porque su padre es un fichero se mide sin el kernel, en el escenario 5. En todos, el SHA-256
de `cache.db` antes y después es el mismo (o el fichero sigue sin existir).

---

## Escenario 5 — La caché vive donde la persona usuaria decide (SC-004, US4)

```bash
rtk proxy go test -race -count=1 -v \
  -run '^(TestRutaEfectivaPrecedencia|TestRutaInservible|TestDirectorioNoCreable|TestEsInexistente|TestNewPorOmisionUsaElDirectorioDeLaCuenta|TestAdaptadorConVariableDeEntorno)$' \
  ./internal/cache/ | grep -E '^--- (PASS|FAIL)'
```

**Esperado**: seis líneas `--- PASS`. Precedencia opción > variable > omisión; la ruta por omisión es
`<HOME>/.cache/kitlegal/cache.db` con `HOME` redirigido por el test a un directorio temporal; variable
vacía o que apunta a un fichero → código 2 nombrando `KITLEGAL_CACHE_DIR` y la ruta, sin caída a la ruta
por omisión; un directorio no creable porque su padre es un fichero → 2 en modo normal y, en solo
lectura, cliente sin base y `Get` → 4 con el fichero padre intacto (la regla «inexistente» reconoce el
`ENOTDIR` de `Stat`, que no es `fs.ErrNotExist`). Ninguno escribe bajo el `HOME` real.

---

## Escenario 6 — Esquema, migraciones y `PRAGMA` (SC-005, SC-006, US5)

```bash
rtk proxy go test -race -count=1 -v \
  -run '^(TestMigracionesEmbebidasBienFormadas|TestMigracionesIdempotentes|TestVersionMayorQueLaConocida|TestMigracionAtomica|TestSoloLecturaNoMigra|TestDosClientesMigranUnaVez|TestFicheroInutilizable|TestAbrirAplicaLosPragma|TestAbrirCreaDirectorioYFicheroConPermisosReservados)$' \
  ./internal/cache/ | grep -E '^--- (PASS|FAIL)'
```

**Esperado**: nueve líneas `--- PASS`. Abrir dos veces deja `schema_version` en 1 sin error; una base
con `version = 99` termina con código 1, el mensaje dice «esquema en la versión 99 y este binario
conoce la 1» y el fichero es idéntico byte a byte; un fichero de texto termina con 1 y no se borra; la
migración interrumpida (tabla `entradas` ajena preexistente) deja la base sin `schema_version`; la base
creada responde `journal_mode = wal`, `synchronous = 2`, `busy_timeout = 5000`, y en solo lectura
`query_only = 1`; el directorio nace con `0700` y `cache.db` con `0600`.

---

## Escenario 7 — Dos clientes a la vez: mismo proceso y dos procesos; los tests de integración solo escriben en su directorio temporal (SC-007, SC-009)

```bash
rtk proxy go test -race -count=1 -run '^(TestDosClientesEnElMismoProceso|TestMismaClaveDosEscritores|TestSoloLecturaVeLoConfirmadoEnElWAL)$' ./internal/cache/
SONDA=$(mktemp -d)
TMPDIR="$SONDA" rtk proxy go test -race -count=1 -tags=integration -v -run '^TestIntegracion' ./internal/cache/ \
  | grep -E '^\s*--- (PASS|FAIL|SKIP)'
find "$SONDA" -mindepth 1 | wc -l
rm -rf "$SONDA"
make test-integration
```

**Esperado**: la primera orden termina en `ok`; la segunda muestra solo líneas `--- PASS` (ninguna
`SKIP`: en una máquina sin privilegios los permisos se hacen valer; este escenario se ejecuta en local, sin
la variable `CI`, que es donde la precondición incumplida salta en vez de fallar) —aquí el filtro **sí** recoge las
líneas de los tests padre además de las subpruebas sangradas, porque este escenario no cuenta líneas,
sino que comprueba qué nombres aparecen y que ninguno es `SKIP`—, entre ellas
`TestIntegracionDosProcesos/escritor-padre-lector-solo-lectura-hijo`,
`TestIntegracionDosProcesos/escritor-hijo-lector-padre`,
`TestIntegracionDirectorioNoEscribible/solo-lectura-lee`,
`TestIntegracionDirectorioNoEscribible/normal-argumentos`, `TestIntegracionDirectorioDenegado` y
`TestIntegracionWALSinMemoriaCompartida`; el `find` imprime **`0`** (todo lo que los tests crearon bajo el
directorio temporal se borró con ellos; en macOS `TMPDIR` es lo que `os.TempDir` honra); `make
test-integration` termina en verde. El hijo de los dos procesos es el propio binario de test relanzado
con `-test.run=^TestProcesoAuxiliar$` y el papel en el entorno.

---

## Escenario 8 — Las quince situaciones de fallo, con su código de salida (SC-011, FR-033)

```bash
rtk proxy go test -race -count=1 -v -run '^TestClasesDeError$' ./internal/cache/ \
  | grep -cE '^\s*--- PASS: TestClasesDeError/[^/ ]+ '
```

El filtro cuenta solo las subpruebas de **primer nivel** (tras la barra, un nombre sin ninguna otra
barra): la línea del test padre no la cuenta, y una subprueba anidada tampoco la contaría —`tasks.md`
(T009) fija además que las trece filas son trece `t.Run` sin anidar, de modo que las dos comprobaciones
de cada fila (el error tal como sale del paquete y envuelto con `%w`) viven dentro de su subprueba.

**Esperado**: **`13`**, una subprueba por cada fila provocable sin permisos de la tabla de
[`contracts/errores-y-codigos.md`](./contracts/errores-y-codigos.md) §3 (filas 1-11, 14 y 15), cada una
comprobando `cli.Clasificar` y `cli.CodigoSalida` sobre el error real. Las filas 12 y 13 (permisos) las
cubren `TestIntegracionDirectorioDenegado` y `TestIntegracionWALSinMemoriaCompartida` del escenario 7.

---

## Escenario 9 — Los controles mecánicos fallan ante cada intento de saltarse una regla (SC-013, SC-008, FR-041)

Todo sobre una copia desechable **fuera del repositorio**; el árbol de trabajo no se toca.

```bash
COPIA=$(mktemp -d)
tar --exclude=.git -cf - . | tar -xf - -C "$COPIA"
```

**a. `database/sql` fuera de `internal/{cache,store,graph}` (R3)** — lint **y** test de arquitectura:

```bash
printf 'package app\n\nimport _ "database/sql"\n' > "$COPIA/internal/app/violacion_r3.go"
( cd "$COPIA" && make lint 2>&1 | grep -c 'R3: el acceso a base de datos vive en internal/{cache,store,graph}' )
( cd "$COPIA" && go test ./internal/ -run 'TestArquitectura/R3' 2>&1 | grep -c 'R3 · importación reservada' )
rm "$COPIA/internal/app/violacion_r3.go"
```

**Esperado**: `1` y `1` (o más: una línea por hallazgo).

**b. El dominio importa el adaptador de caché (R1)** — en un subpaquete nuevo del dominio, porque
`internal/cache` importa `internal/core` y un import en el propio `internal/core` sería un ciclo (el
compilador lo rechazaría antes de que ninguna regla se nombrara):

```bash
mkdir -p "$COPIA/internal/core/violacion"
printf 'package violacion\n\nimport _ "github.com/jmorenobl/kitlegal/internal/cache"\n' > "$COPIA/internal/core/violacion/violacion.go"
( cd "$COPIA" && make lint 2>&1 | grep -c 'R1: el dominio no importa adaptadores; la caché se usa desde internal/source' )
( cd "$COPIA" && go test ./internal/ -run 'TestArquitectura/R1' 2>&1 | grep -c 'R1 · el dominio es puro' )
rm -r "$COPIA/internal/core/violacion"
```

**Esperado**: `1` y `1`.

**c. Un `*sql.Rows` sin cerrar en un fichero etiquetado `integration`** — `sqlclosecheck` lo ve porque
`run.build-tags` alcanza el fichero (FR-041):

```bash
cat > "$COPIA/internal/cache/violacion_sqlclose_test.go" <<'EOF'
//go:build integration

package cache_test

import (
	"context"
	"database/sql"
)

func filasSinCerrar(ctx context.Context, db *sql.DB) error {
	rows, err := db.QueryContext(ctx, "SELECT 1")
	if err != nil {
		return err
	}
	_ = rows
	return nil
}
EOF
( cd "$COPIA" && make lint 2>&1 | grep -c 'sqlclosecheck' )
rm "$COPIA/internal/cache/violacion_sqlclose_test.go"
```

**Esperado**: `1` (o más; `unused` puede añadir el suyo sobre `filasSinCerrar`). Sin `run.build-tags`
en `.golangci.yml`, el fichero quedaría fuera del análisis y el recuento sería `0`.

**d. `net/http` en un test de `internal/cache` (R2)**:

```bash
printf 'package cache_test\n\nimport _ "net/http"\n' > "$COPIA/internal/cache/violacion_r2_test.go"
( cd "$COPIA" && make lint 2>&1 | grep -c 'R2: la red se usa a través de internal/httpx' )
rm -rf "$COPIA"
```

**Esperado**: `1`. Al terminar, `git status --porcelain` en el repositorio sigue sin mostrar nada nuevo.

---

## Escenario 10 — La superficie del binario no cambia (SC-012, FR-044)

```bash
make build
./bin/kitlegal --help
./bin/kitlegal version
go list -deps ./cmd/kitlegal | grep -cE 'internal/cache|modernc\.org' ; true
go version -m bin/kitlegal | grep -c 'modernc.org' ; true
rtk proxy go test -count=1 ./internal/ -run '^(TestElBinarioNoEnlazaCache|TestElBinarioNoEnlazaHTTPX|TestDependenciasDelBinario|TestElBinarioNoEnlazaLosEjemplos)$'
grep -rln --include='*.go' --exclude='*_test.go' -e 'adaptadorDePrueba' -e '"consultar"' -e '"guardar"' internal/app cmd internal/cache ; test $? -eq 1 && echo "el adaptador de prueba solo existe en tests"
```

**Esperado**: la ayuda del binario es la misma que al cerrar H2 (sin ningún applet registrado; solo
`version` y la ayuda); los dos `grep -c` imprimen **`0`**; los cuatro tests pasan; «el adaptador de prueba
solo existe en tests».

---

## Escenario 11 — `make ci` en verde, con `test-integration` dentro, sin supresiones nuevas y con los umbrales (SC-008, SC-009, SC-014)

```bash
make ci
grep -nE '^ci:' Makefile
rtk proxy git diff main -- '*.go' | grep -c '^+.*//nolint' ; true
rtk proxy git diff main -- .golangci.yml
go tool cover -func=coverage.out | tail -1
```

**Esperado**: «ci: todos los controles en verde» (con `test-integration` ejecutado entre `test` y
`vuln`); la línea `ci: fmt-check lint test test-integration vuln schema-check secrets mod-verify mod-tidy-check`;
el recuento de `//nolint` añadidos es **`0`**; el diff de `.golangci.yml` frente a `main` contiene
únicamente la clave `build-tags` bajo `run`, comentarios de la lista `sql` y, si S2 de research.md se
activó, palabras bajo `misspell.ignore-rules` (ninguna regla ni exclusión nueva); la cobertura total es
≥ 70,0 %.

---

## Escenario 12 — Un lector de solo lectura ve lo que otra invocación ya confirmó y respeta sus bloqueos (FR-015, FR-032, US3 escenario 6)

```bash
rtk proxy go test -race -count=1 -v \
  -run '^(TestSoloLecturaVeLoConfirmadoEnElWAL|TestDosClientesEnElMismoProceso)$' ./internal/cache/ \
  | grep -E '^\s*--- (PASS|FAIL)'
```

**Esperado**: cuatro líneas `--- PASS` —las dos de los tests padre,
`TestSoloLecturaVeLoConfirmadoEnElWAL` (sin subtests) y `TestDosClientesEnElMismoProceso`, que `-v`
imprime sin sangrar, y las dos subpruebas sangradas
`TestDosClientesEnElMismoProceso/normal-normal` y `TestDosClientesEnElMismoProceso/normal-solo-lectura`—
y ninguna `--- FAIL`. El escritor mantiene su cliente abierto (sin checkpoint) mientras el lector de solo lectura lee: toda
entrada confirmada se encuentra, ninguna a medias, y una escritura sin confirmar no se ve.
