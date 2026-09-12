# Implementation Plan: H3 · `internal/cache`: SQLite con TTL y `--offline`

**Branch**: `h3-internal-cache-sqlite` | **Date**: 2026-09-12 | **Spec**: [spec.md](./spec.md)

**Input**: Feature specification from `/specs/004-h3-internal-cache-sqlite/spec.md`

**Modo**: desatendido. Las decisiones técnicas se tomaron con el «Criterio de decisión autónoma» de
`.specify/memory/constitution.md` y están registradas —decisión, alternativas y motivo— en
[research.md](./research.md). Toda afirmación sobre una herramienta o dependencia externa está
verificada en local (código, documentación o **sonda ejecutada**) y research.md cita dónde; lo que no se
pudo verificar está declarado como supuesto en [research.md D18](./research.md#d18--supuestos-no-verificados),
no como hecho.

## Summary

H2 dejó la única puerta a la red; H3 construye **la memoria de lo que ya se pidió**: una caché local
en SQLite con vigencia por entrada, que la segunda consulta idéntica responda sin tocar la red, que lo
caducado no se sirva y que `--offline` signifique «solo lee, y código 4 si falta». No entrega ningún
verbo ni toca ninguna skill (principio VIII: fundación que protege a las skills que vendrán; el primer
applet que la use es `boe`, H4).

Cinco decisiones sostienen el diseño:

1. **El puerto vive en el dominio** ([D1](./research.md#d1--el-puerto-vive-en-internalcore-paquete-core-como-corecache-con-dos-métodos)):
   nace el paquete `core` en `internal/core/` con `core.Cache` —`Get(ctx, clave) ([]byte, bool, error)`
   y `Put(ctx, clave, contenido, vigencia) error`, la firma de la clarificación Q3— sin `Close` y sin
   importar nada del módulo. R3 gana su **primer dueño real**: `internal/cache`.
2. **Un tipo, un constructor con contexto, cuatro opciones** ([D2](./research.md#d2--superficie-de-internalcache-un-tipo-un-constructor-con-contexto-cuatro-opciones-tres-métodos)):
   `cache.New(ctx, ConDirectorio, SoloLectura, ConReloj, ConRegistrador)`. El modo lo pone quien
   construye a partir de `schema.Contexto.Offline` (FR-015); la caché no lee ninguna bandera.
3. **Solo lectura que no crea, no migra y ve todo lo confirmado** ([D5](./research.md#d5--apertura-en-solo-lectura-stat-primero-modero-y-immutable1-solo-cuando-el-directorio-no-admite-escritura)):
   `Stat` antes de abrir con una única regla «inexistente» —`fs.ErrNotExist` **o** `syscall.ENOTDIR`, este
   último cuando el padre del directorio es un fichero— (inexistente → sin base → ausencia → código 4;
   cualquier otro fallo de `Stat` → 1), `mode=ro` + `query_only` sin
   `immutable` para ver el WAL de otra invocación, y `immutable=1` **solo** cuando el directorio no
   admite crear `-shm` y no hay `-wal` (el contenedor de solo lectura de US3). Todo verificado con
   sondas contra el driver.
4. **Esquema mínimo, migraciones atómicas, `PRAGMA` conservadores** ([D4](./research.md#d4--apertura-normal-fichero-pre-creado-a-0600-dsn-con-pragma-seguros-una-conexión),
   [D6](./research.md#d6--esquema-v1-dos-tablas-strict-expiración-como-integer-de-nanosegundos-unix-sin-índice),
   [D7](./research.md#d7--migraciones-embebidas-schema_version-atomicidad-y-carrera-entre-dos-procesos)):
   dos tablas `STRICT`, expiración en nanosegundos comparada en Go, una migración embebida aplicada en
   una transacción inmediata que relee la versión, `WAL` + `synchronous=FULL` + espera ante bloqueo de
   5 s por tramos de `busy_timeout=100` que miran el contexto (`espera.go`; revisión final),
   fichero pre-creado a `0600` y directorio a `0700`, y la ruta codificada para el URI del DSN.
5. **Errores con clase, tabla cerrada de quince situaciones** ([D9](./research.md#d9--errores-tipados-cacheerror-declara-su-clase-tabla-cerrada-de-quince-situaciones)):
   `cache.Error` implementa `schema.ConClase`; la ausencia en solo lectura es «fuente no disponible»
   (4); nunca 3, 5 ni 6.

Todo el código nuevo vive en `internal/core` (el puerto) e `internal/cache` (producto, tests y el único
fixture), más el test de arquitectura, una línea del `Makefile`, dos líneas de `.golangci.yml`,
`go.mod`/`go.sum` y una nota en `docs/PENDIENTES.md`. El binario distribuido no cambia y no enlaza el
paquete ni el driver (SC-012).

## Technical Context

**Language/Version**: Go 1.27, sin cambios: `go.mod` mantiene `go 1.27.0` + `toolchain go1.27.1` y el
`Makefile` sigue exportando ese `GOTOOLCHAIN`. Verificado: `go version` → `go1.27.1 darwin/arm64`. La
dependencia nueva exige menos (`modernc.org/sqlite` declara `go 1.25.0`).

**Primary Dependencies**: **una** dependencia nueva, la que el hito nombra y la constitución §V y
`docs/ROADMAP.md` §3 fijan (FR-043): `modernc.org/sqlite` **v1.58.0** (SQLite 3.53.4 traducido a Go,
sin cgo, ADR 0002), descargada en la caché de módulos y ejercitada con cuatro sondas (research.md,
tabla de verificación). Arrastra módulos que quedan en `go.mod` como indirectos: la sonda 6 de
research.md (`go mod tidy` sin red sobre un módulo que importa el driver) resolvió `modernc.org/libc`
v1.75.6, `mathutil`, `memory`, `golang.org/x/sys` v0.47.0, `go-humanize`, `uuid`, `go-isatty`,
`go-strftime` y `bigfft`, y **no** `modernc.org/fileutil` ni `github.com/google/pprof`, que el `go.mod`
del driver declara pero ningún paquete alcanzable importa; la lista definitiva es la que `go mod tidy`
escriba en el repositorio (supuesto S1) y ningún documento la copia a mano. **Ninguna llega al binario
distribuido** en H3: ningún paquete de producción importa `internal/cache`, y `TestDependenciasDelBinario`
más `TestElBinarioNoEnlazaCache` lo fijan. Todo lo demás es biblioteca estándar: `database/sql`,
`embed`, `context`, `time`, `os`, `path/filepath`, `log/slog`, `errors`, `fmt`, `sync`, `io/fs` y
`syscall` (solo en `ruta.go` y solo por la constante `ENOTDIR` de la regla «inexistente», D3); en
tests, `os/exec` y `crypto/sha256`. Ni `net/http` ni `net/http/httptest` en ningún fichero de
`internal/cache` (R2, `depguard` por prefijo).

**Herramientas de control**: las de H0/H1/H2 sin cambio de versión (`golangci-lint` v2.13.2,
`govulncheck`, `gitleaks`, `lefthook`). H3 no añade ninguna, no activa ningún linter (`sqlclosecheck` y
`rowserrcheck` están activos desde H0; por primera vez tienen código que vigilar) y no añade ninguna
exclusión ni `//nolint` (SC-008). Añade `run.build-tags: [integration]` para que el análisis alcance los
ficheros etiquetados (FR-041; clave verificada en `pkg/config/run.go`).

**Storage**: `cache.db` (SQLite, WAL) en `~/.cache/kitlegal/` o en `KITLEGAL_CACHE_DIR`, con esquema
versionado en `schema_version` (v1: tabla `entradas(clave, contenido, expira_en)`). Solo en directorios
temporales durante los tests. Ningún otro fichero; ningún `world.db` (H17).

**Testing**: `go test -race -shuffle=on -coverprofile=coverage.out ./...` (`make test`, sin cambios) más
`go test -race -tags=integration ./...` (`make test-integration`, receta de H0 sin cambios) que **entra
en `make ci`**. Todos los tests trabajan sobre una base real en `t.TempDir()`; los etiquetados
`integration` son los que dependen del entorno (permisos, dos procesos). Un solo fixture, escrito a
mano (`internal/cache/testdata/reproduccion/prueba/GET_http_fuente.prueba_norma.json`), para que la
primera consulta del adaptador de prueba pase por `httpx.Replay`; la segunda usa `Replay` sobre un
directorio vacío (estricto, FR-046). **Cero red en todos los tests**, **ninguna grabación real**, y
ningún test toca la caché de la cuenta (`HOME` se redirige con `t.Setenv` donde hace falta). Sin guion
`testscript` nuevo: el spec lo excluye porque H3 no cambia ningún comportamiento visible del binario.

**Target Platform**: sin cambios (binario sin cgo, `-trimpath`; CI en `ubuntu-latest`; desarrollo en
darwin/arm64). `modernc.org/sqlite` publica SQLite 3.53.4 para darwin/linux/windows × amd64/arm64
(`doc.go`). Los tests de permisos se miden donde el sistema de ficheros los hace valer (SC-004) y en CI el
ejecutor no es privilegiado (supuesto S3): allí la precondición incumplida es `t.Fatalf` y pone el trabajo
en rojo; fuera de la integración continua es `t.Skip` nombrando la causa.

**Project Type**: paquete interno de una herramienta de línea de órdenes multicall; ningún applet nuevo.

**Performance Goals**: ninguna propia. Se conserva la de H0 (`ci` de una PR limpia < 3 min en caliente;
supuesto S5 con `test-integration` dentro). El test de caducidad no espera tiempo real (SC-002): el
reloj se inyecta.

**Constraints**: por omisión `synchronous=FULL`, espera ante bloqueo de 5 s en total —`busy_timeout=100`
por intento y reintento del cliente mirando el contexto entre tramos—, una conexión por cliente; no
existe ninguna opción exportada para desactivar WAL, bajar la sincronización, cambiar la espera ni abrir
en solo lectura con `immutable` a voluntad (FR-031; D4, D5). En solo lectura no se crea ni el directorio ni el
fichero ni se aplica ninguna migración, y `cache.db` queda idéntico byte a byte (SC-003). Ningún
`panic` en ninguna ruta del paquete (FR-034). Sin goroutines propias. `make ci` no modifica ningún
fichero versionado.

**Scale/Scope**: dos paquetes nuevos (`internal/core`, 2 ficheros sin sentencias; `internal/cache`, 8
ficheros de producto más un `.sql` embebido, del orden de 600-900 líneas de Go de producto y bastantes
más de test), un fixture, un fichero de H1 tocado (`internal/arch_test.go`), `Makefile` (una línea),
`.golangci.yml` (dos líneas y comentarios), `go.mod`/`go.sum`, `docs/PENDIENTES.md`.

## Constitution Check

*GATE: debe pasar antes de la fase 0 y volver a evaluarse tras la fase 1.*

### Principios

| # | Principio | Cómo lo cumple H3 | Veredicto |
|---|---|---|---|
| **I** | Fuentes públicas y frontera humana | H3 no emite ninguna petición HTTP: no toca `internal/httpx` (spec, *Fuera de alcance*) y `internal/cache` no importa `net/http` en ningún fichero, tests incluidos (`depguard`, lista `red`, por prefijo; R2). La caché sirve precisamente para **no volver a pedir** a un servidor público lo que ya dio. Ninguna acción con identidad: el código 6 no se produce (FR-033). El adaptador de prueba consulta solo por `httpx.Replay`, que no abre conexión alguna. | ✅ Cumple |
| **II** | Nada sin cita ni fuente | La caché guarda contenido **opaco** (FR-006): lo que el sobre necesita para citar —dirección final, estado, fecha de consulta— lo serializa dentro del contenido quien llama (H4). Y **lo caducado no se sirve, tampoco sin red** (FR-008, FR-018): una cita se sostiene sobre texto vigente; el borde de la vigencia es exacto y se prueba en las tres posiciones (SC-002). No se genera ni interpreta contenido legal. | ✅ Cumple |
| **III** | Tests primero y offline | Todos los tests del hito son offline por construcción: base real en `t.TempDir()`, reproducción de H2 sobre un fixture escrito a mano y sobre un directorio vacío; ningún test conoce una dirección real (quickstart, prerrequisitos) ni toca `~/.cache/kitlegal`. El control literal del hito («unit de expiración de TTL con reloj inyectado») y el criterio de aceptación («`Replay` en modo estricto») tienen test nombrado (§Inventario). Sin e2e nuevo, por la razón que el spec fija (no hay comportamiento visible que describir); el orden «test primero» se concreta por fichero (§Orden de implementación). Cobertura: rigen los umbrales generales (SC-014); `internal/core` (paquete `core`) no tiene sentencias y no altera el componente. | ✅ Cumple |
| **IV** | Arquitectura hexagonal con reglas ejecutables | El puerto `Cache` está en el dominio (`internal/core`) y la implementación en el adaptador `internal/cache` (patrón *Repository* de `docs/ROADMAP.md` §2): el dominio no sabe que hay SQLite. R3 pasa de «activa y vacía» a **activa con dueño**, vigilada en las dos capas (`depguard` lista `sql` + subprueba R3); R1 sigue vigilada sobre el nuevo paquete `core`. Errores tipados que mapean a códigos estables: `cache.Error` + `schema.ConClase` + `cli.Clasificar` (sin cambios en el kernel), tabla cerrada de 15 situaciones → {1, 2, 4} (contrato de errores §3); nunca 3, 5 ni 6. Ningún `panic`. Composición manual con opciones funcionales. `log/slog` por el registrador recibido; `internal/cache` no escribe en stdout (R5) ni llama a `os.Exit` (R4). | ✅ Cumple |
| **V** | Simplicidad y dependencias fijadas (YAGNI) | Una dependencia nueva, de la lista de §V y nombrada por el hito (FR-043); sus indirectos se declaran (S1). Sin DI, sin ORM, sin herramienta de migraciones (D7 rechaza la sugerencia de la skill `golang-database` por esto). Todo lo del spec en *Fuera de alcance* queda fuera del plan: sin `Source`, sin esquema de claves, sin revalidación, sin `store`/`graph`, sin verbos de mantenimiento, sin desalojo, sin índice de expiración, sin cifrado, sin métricas, sin `os.UserCacheDir`, sin cambios en `httpx`, sin e2e, sin `CHANGELOG.md`, sin ADR. Las divergencias de forma están en *Complexity Tracking*; ninguna añade una capacidad. | ✅ Cumple |
| **VI** | Un binario, convenciones de agente | El binario distribuido **no cambia** (FR-044, SC-012): mismo registro vacío, mismos verbos, ningún applet ni bandera nueva, y no enlaza `internal/cache` ni `modernc.org/*` (`TestElBinarioNoEnlazaCache`, `TestDependenciasDelBinario` con la lista intacta). `--offline` adquiere el significado que el hito le da —solo lectura y código 4 si falta— en el único sitio donde H3 lo convierte en modo: el applet de prueba, que lee `schema.Contexto.Offline` (FR-047). `--timeout` llega como plazo del contexto a `New`, `Get` y `Put`. `--dry-run`, `--no-graph` y `--asunto` siguen solo propagados. | ✅ Cumple |
| **VII** | Grafo y privacidad | Sin grafo (H17) y sin `Emit`. Privacidad: la caché guarda respuestas de fuentes públicas en un directorio y un fichero con acceso reservado a la cuenta (`0700`/`0600`, FR-021); los eventos van a nivel `debug` y los mensajes de error nombran rutas, variables y claves, nunca datos de personas; ningún dato sale del disco de la persona usuaria. | ✅ Cumple |
| **VIII** | Skills primero; el binario es la herramienta | H3 es un hito de **fundación** de los que §VIII admite: no entrega skill, pero protege a todas las que vendrán, porque cada adaptador de fuente de cada skill (H4 `boe`, H12 `placsp`, H13 `bdns`…) usará esta caché para no volver a pedir lo que ya tiene y para funcionar con `--offline`. La caché es una de las capacidades que §VIII reserva al binario («caché»). No se construye ninguna herramienta que ninguna skill vaya a usar: el único adaptador del hito es de prueba y vive en material de test (FR-045). | ✅ Cumple |
| **IX** | Genericidad territorial, validación local | H3 no tiene dimensión territorial (spec, *Fuera de alcance*): ningún municipio, boletín ni territorio aparece en código, `data/` ni fixtures. El host del fixture es `fuente.prueba`. El punto 11 de la Definition of Done no aplica. | ✅ Cumple |

### Reglas de dependencia (`docs/ROADMAP.md` §2, constitución §IV)

| Regla | Situación en H3 | Cómo se hace cumplir | Veredicto |
|---|---|---|---|
| `internal/core/**` no importa `internal/{source,httpx,cache,store,graph,render,cli,app}` ni I/O de la biblioteca estándar | **Nuevo paquete `core`** (`internal/core/{doc,cache}.go`) que importa solo `context` y `time`. `internal/cache` importa `internal/core` y `internal/core/schema`; nunca al revés (FR-038). | `depguard` lista `core` (`**/internal/core/**` cubre el paquete raíz) + `compruebaDominioPuro` (`paquetesBajo` incluye el propio prefijo); demostrado rompiéndola en copia desechable (quickstart 9.b) | ✅ Cumple |
| Solo `internal/httpx` importa `net/http` | Sin cambio de dueño. `internal/cache` no importa `net/http` ni `net/http/httptest` en **ningún** fichero: la lista `red` casa por prefijo y cubre los tests (`run.tests: true`); por eso el adaptador de prueba usa `httpx.Replay` y el literal `"GET"` (D12). | `depguard` lista `red` + subprueba R2 (sin cambios); quickstart 9.d | ✅ Cumple |
| Solo `internal/{cache,store,graph}` importan SQLite y `database/sql` | **Activa con su primer dueño**: `internal/cache` importa `database/sql`, `modernc.org/sqlite` y `modernc.org/sqlite/lib` (cubierto por prefijo). `duenoObligatorio` sigue en `false` con la razón escrita (FR-037: faltan `store` y `graph`). | `depguard` lista `sql` + subprueba R3; quickstart 9.a | ✅ Cumple |
| Solo `internal/cli` y `cmd/` llaman a `os.Exit` | Sin cambio: `internal/cache` no llama a `os.Exit`; el kernel se ejercita con `app.Main` en proceso; el proceso auxiliar de los tests de integración es el propio binario de test y termina devolviendo de su función de test. | `forbidigo` `^os\.Exit$` (sin cambios) | ✅ Cumple |
| Solo `internal/render` escribe en stdout; logs a stderr | Sin cambio: `internal/cache` no nombra `os.Stdout`/`os.Stderr` ni usa `fmt.Print*`; escribe únicamente por el `*slog.Logger` recibido (D14). | `forbidigo` `^fmt\.Print(\|f\|ln)$`, `^os\.Stdout$`, `^os\.Stderr$` (sin cambios) | ✅ Cumple |

### Gates mecánicos (constitución «Gates», capa 1)

Los que este hito **activa por primera vez o endurece**: `sqlclosecheck` y `rowserrcheck` con código
que vigilar, alcanzando también los ficheros etiquetados (`run.build-tags`); tests de integración
etiquetados dentro de `make ci`; R3 con dueño; fixtures de códigos de salida para las 15 situaciones de
fallo (13 deterministas + 2 en integración; SC-011); `-race` sobre dos clientes en el mismo proceso y
sobre dos procesos (SC-007); el binario sin `internal/cache` ni `modernc.org/*` (SC-012); test de
superficie exportada (FR-005); medida byte a byte de `cache.db` bajo `--offline` (SC-003).

Los que **siguen sin objeto**, con su hito: corrección de citas contra la respuesta grabada del BOE (H4);
`schemas/*.json` y su deriva (H4/H10); frontmatter y `references/` de skills (H5); matriz territorial
(H7). Ninguno se simula.

Los de H0, H1 y H2 siguen todos activos y **sin ninguna exclusión nueva** (SC-008).

**Reglas del modo desatendido** (constitución): `KITLEGAL_RECORD=1` **no se ejecuta** en ninguna tarea
del hito; la única tarea que crea material bajo `internal/cache/testdata/` lleva `[datos]` y es un
fixture escrito a mano contra un host ficticio (`docs/WORKFLOW.md`: exige etiqueta y guardián, sin
pausa); ninguna tarea toca `testdata/` de raíz, `schemas/`, `docs/SOURCES.md` ni `internal/source/`.

**Veredicto del gate: PASA.** No hay ninguna violación. Las divergencias conscientes están en
*Complexity Tracking*; ninguna afecta a alcance, frontera humana, privacidad, términos de uso ni a una
decisión cerrada.

### Re-evaluación tras la fase 1 (diseño)

El diseño ([data-model.md](./data-model.md), [contracts/](./contracts/), [quickstart.md](./quickstart.md))
no añade ninguna capacidad sobre lo evaluado. Lo que apareció al bajar a diseño y roza un principio:

- **El paquete `core` nace ahora** (§IV): es el sitio que `docs/ROADMAP.md` §2 reserva a los puertos y
  no reabre ninguna decisión; contiene una interfaz y un comentario, sin sentencias. No exige ADR: el
  patrón *Repository* y «interfaz `Cache` definida en `core`» ya estaban decididos.
- **`immutable=1` como reapertura de contingencia** (§IV, §V): no es una opción configurable ni un modo
  general (Q2 lo descartó); se activa solo ante `SQLITE_READONLY_DIRECTORY` (1544) o `SQLITE_CANTOPEN`
  (14) **sin `-wal`** —los dos códigos que produce un directorio que no admite crear `-shm`, según qué
  auxiliares existan—, condición en la que ningún escritor puede existir, y está verificado con sondas
  (D5). Sin él, el caso más común de
  `--offline` (directorio de solo lectura) terminaría en código 1 contra el DEBE de FR-015.
- **El adaptador de prueba usa `Replay` en las dos consultas y un fixture a mano** (§III, R2): se
  detectó al escribir el contrato de arquitectura que un servidor local exigiría `net/http` en
  `internal/cache`, que `depguard` deniega por prefijo también en tests. La alternativa —una exclusión
  para `internal/cache/*_test.go`— sería una supresión (SC-008). El fixture es material de test contra
  `fuente.prueba`, en una tarea `[datos]` sin pausa, como los diez de H2 (D12).
- **Dos situaciones de fallo no enumeradas en FR-033** (§IV, ningún fallo sin clase): ruta por omisión
  indeterminable → «argumentos» (2), contexto cancelado o vencido → «fuente no disponible» (4). Declaradas
  en *Complexity Tracking* y en el contrato de errores (filas 7 y 15).
- **`t.Skip` condicionado a que el sistema de ficheros haga valer los permisos, y solo fuera de la
  integración continua** (§V, «nada de tests desactivados»): es la precondición de entorno que SC-004
  declara y nunca se activa en un entorno donde la aserción tenga sentido; dentro de la integración
  continua (`CI` no vacía) la misma precondición incumplida es `t.Fatalf`, de modo que S3 lo hace valer el
  propio gate —el trabajo se pone en rojo— y no la lectura de un registro. Declarado en
  *Complexity Tracking*.
- **`make ci` gana `test-integration`** (§III, «todo lo que entra se queda como gate de CI»): la receta es
  contrato de H0 y no cambia; solo cambia la lista de prerrequisitos de `ci`, donde `test-integration`
  entra entre `test` y `vuln`. Coste: repetir la suite unitaria (≈ 22 s hoy).

**Veredicto tras el diseño: PASA**, sin ninguna violación.

## Project Structure

### Documentation (this feature)

```text
specs/004-h3-internal-cache-sqlite/
├── plan.md              # Este fichero
├── research.md          # Fase 0: 18 secciones (decisiones con alternativas y motivo, verificación con sondas, supuestos)
├── data-model.md        # Fase 1: entidades, invariantes, estados del cliente, flujo de una consulta
├── quickstart.md        # Fase 1: guía de validación ejecutable, 12 escenarios
├── contracts/
│   ├── puerto-y-cliente.md          # core.Cache, la superficie de internal/cache, New por modo, Get/Put, Close, valores por omisión
│   ├── errores-y-codigos.md         # cache.Error, la tabla cerrada de 15 situaciones → clase → código, mensajes
│   ├── esquema-y-apertura.md        # esquema v1, migraciones, DSN por modo, PRAGMA, apertura normal y de solo lectura
│   └── reglas-de-arquitectura.md    # R1-R5 en H3, sqlclosecheck y rowserrcheck, importaciones, qué falla ante cada violación
├── spec.md
├── checklists/
├── gates/
└── tasks.md             # Fase 2 (`/speckit-tasks`, no lo crea este comando)
```

### Source Code (repository root)

Estructura de `docs/ROADMAP.md` §2 y `CLAUDE.md`. En negrita lo que H3 crea; lo demás existe de
H0/H1/H2 y se indica qué cambia.

```text
internal/
├── core/                        el dominio; hoy solo tenía el subpaquete schema
│   ├── doc.go                   **NUEVO** package core: comentario de paquete («los puertos del dominio»)     (D1)
│   ├── cache.go                 **NUEVO** el puerto Cache (Get/Put); importa solo context y time              (D1)
│   └── schema/                  sin cambios (Contexto.Offline, Clase, ConClase ya existen)
├── cache/                       **NUEVO** el adaptador de caché: primer dueño de R3
│   ├── doc.go                   qué garantiza el paquete y qué no hace (no lee banderas, no borra, no cifra)
│   ├── cliente.go               Cliente, New, Opcion, ConDirectorio, SoloLectura, ConReloj, ConRegistrador,
│   │                            Close; VariableDirectorio                                                      (D2)
│   ├── ruta.go                  resolución de la ruta efectiva (opción > variable > omisión), Stat del
│   │                            directorio y su validación por modo; la regla «inexistente»
│   │                            (esInexistente: fs.ErrNotExist o syscall.ENOTDIR); origen para los mensajes   (D3)
│   ├── abrir.go                 apertura normal (MkdirAll 0700, OpenFile 0600, DSN con la ruta codificada
│   │                            para el URI, pool de 1) y de solo lectura (Stat de cache.db con la misma
│   │                            regla, mode=ro; ante 1544 o 14, primero si cache.db se deja leer, después
│   │                            reapertura immutable sin -wal; con -wal, código 1); clasificadores que
│   │                            miran el contexto y distinguen bloqueo de fichero inutilizable            (D4, D5)
│   ├── espera.go                espera ante bloqueo: presupuesto de 5 s por tramos de busy_timeout de
│   │                            100 ms que miran el contexto entre intentos (revisión final)               (D4, D10)
│   ├── migraciones.go           embed.FS, versión conocida, versión registrada, aplicación en transacción
│   │                            inmediata con relectura, comprobación en solo lectura                          (D7)
│   ├── migraciones/
│   │   └── 0001_entradas.sql    esquema v1: schema_version y entradas, STRICT                                  (D6)
│   ├── entradas.go              Get (ausencia, expiración en Go, normalización del BLOB vacío, clase 4 en
│   │                            solo lectura) y Put (validación, upsert)                                       (D8)
│   ├── errores.go               Error{Operacion, Ruta, Origen, Clave, Causa}, Clase(), Error(), Unwrap()      (D9)
│   ├── *_test.go                un fichero de test por fichero de código, mismo orden; más
│   │                            concurrencia_test.go, superficie_test.go (package cache);
│   │                            adaptador_test.go e integracion_test.go (package cache_test)                 (D11, D12)
│   └── testdata/                **NUEVO** material de test escrito a mano, no fixtures reales; tarea [datos]
│       └── reproduccion/prueba/
│           └── GET_http_fuente.prueba_norma.json   la única grabación del hito                                 (D12)
├── arch_test.go                 ← comentario de R3 («primer dueño real; sin duenoObligatorio hasta H17»);
│                                  TestElBinarioNoEnlazaCache                                                   (D13)
Makefile                         ← ci: fmt-check lint test test-integration vuln schema-check secrets mod-verify mod-tidy-check (D11)
.golangci.yml                    ← run.build-tags: [integration]; comentarios de la lista sql                  (D11, D13)
go.mod, go.sum                   ← + modernc.org/sqlite v1.58.0 y sus indirectos                               (S1)
docs/PENDIENTES.md               ← «En H4»: retirar TestElBinarioNoEnlazaCache y ampliar modulosDelBinario     (D13)
```

**Structure Decision**: la de `docs/ROADMAP.md` §2 sin desviaciones: `internal/cache` es el adaptador
que el diagrama sitúa junto a `httpx`, `store`, `graph` y `render` («Repository sobre SQLite»), y el
puerto va al paquete raíz del dominio, que el diagrama describe como «define los PUERTOS». No se crean
`internal/source/`, `internal/store/`, `internal/graph/` ni `pkg/`: están fuera de alcance y crearlos
vacíos sería un marcador de posición. `internal/cache/testdata/` sigue la convención de Go (fixtures
junto al paquete, H2 D17); la decisión sobre los fixtures reales de fuentes queda para H4
(`docs/PENDIENTES.md`).

## Controles mecánicos que este hito añade o toca

Según la sección «Gates» de la constitución. Cada fila dice qué orden lo ejecuta, si está en `make ci` y
**qué falla si el control se retira**.

| # | Control | Herramienta / forma | Orden | ¿En `make ci`? | Demostración de que está activo |
|---|---|---|---|---|---|
| 1 | **R3 con dueño** | subprueba R3 de `TestArquitectura` (sin `duenoObligatorio`, FR-037) + `depguard` lista `sql` | `test`, `lint` | Sí | Un `import "database/sql"` en `internal/app` falla en `depguard` **y** en R3 (quickstart 9.a) |
| 2 | **R1 sobre el paquete `core` nuevo** | `compruebaDominioPuro` + `depguard` lista `core` | `test`, `lint` | Sí | Un `import ".../internal/cache"` en `internal/core/cache.go` falla en los dos (quickstart 9.b) |
| 3 | **`sqlclosecheck` / `rowserrcheck`** | activos desde H0; alcanzan también `integracion_test.go` por `run.build-tags` | `lint` | Sí | Un `*sql.Rows` sin `Close` en el fichero etiquetado → «Rows/Stmt/NamedStmt was not closed» (quickstart 9.c) |
| 4 | **R2 en `internal/cache`** | `depguard` lista `red` (por prefijo, tests incluidos) | `lint` | Sí | Un `import "net/http"` en un `_test.go` de `internal/cache` → «R2: …» (quickstart 9.d) |
| 5 | **Tests de integración en CI** | `make ci` incluye `test-integration` (`go test -race -tags=integration ./...`) | `test-integration` | Sí | Quitar `test-integration` de `ci` deja `TestIntegracion*` sin ejecutar; quickstart 11 comprueba la línea `ci:` |
| 6 | **Superficie exportada** | `TestSuperficieExportada`: `go/parser` + `go/ast` sobre las declaraciones exportadas; falla ante `sql.*`, `sqlite.*` o una declaración fuera del contrato | `test` | Sí | Exportar un método que devuelva `*sql.DB` hace fallar el test (FR-005) |
| 7 | **Puerto implementado** | `var _ core.Cache = (*Cliente)(nil)` | compilación | Sí | Cambiar la firma de `Get` no compila (FR-001, FR-002) |
| 8 | **Criterio de aceptación** | `TestAdaptadorDePruebaConElKernel/{primera-consulta,segunda-consulta-sin-red,sin-cache-la-reproduccion-falla}` con `httpx.Replay` estricto sobre un directorio vacío | `test` | Sí | Sin `Put` tras la primera consulta, la segunda termina con clase 1 nombrando `GET http://fuente.prueba/norma` (SC-001) |
| 9 | **Expiración con reloj inyectado** | `TestExpiracionConRelojInyectado` (tres posiciones del borde; duración < 1 s) | `test` | Sí | Comparar con `<=` en vez de `Before` hace fallar la posición «instante exacto» (SC-002, FR-008) |
| 10 | **`--offline` por el kernel** | seis subtests `offline-*` con SHA-256 de `cache.db` antes y después, y listado del directorio | `test` | Sí | Migrar en solo lectura crea `cache.db` en `offline-sin-base` y el listado deja de estar vacío; servir lo caducado hace pasar `offline-expirada` con 0 en vez de 4 (SC-003) |
| 11 | **Ubicación y precedencia** | `TestRutaEfectivaPrecedencia`, `TestRutaInservible`, `TestDirectorioNoCreable`, `TestNewPorOmisionUsaElDirectorioDeLaCuenta`, `TestAdaptadorConVariableDeEntorno` (`t.Setenv`, `HOME` redirigido) | `test` | Sí | Tratar la variable vacía como ausente hace pasar la caída a la ruta por omisión que `TestRutaInservible` prohíbe (SC-004, FR-022); reconocer solo `fs.ErrNotExist` como «inexistente» hace salir `TestDirectorioNoCreable/solo-lectura` (padre fichero, `Stat` → `ENOTDIR`) con 1 en vez de 4 (FR-015, SC-011) |
| 12 | **Esquema y migraciones** | `TestMigracionesEmbebidasBienFormadas`, `TestMigracionesIdempotentes`, `TestVersionMayorQueLaConocida`, `TestMigracionAtomica`, `TestSoloLecturaNoMigra`, `TestDosClientesMigranUnaVez`, `TestFicheroInutilizable`, `TestContextoCanceladoDuranteLaMigracion` | `test` | Sí | Aplicar la migración sin transacción deja `schema_version` creada tras el fallo de `TestMigracionAtomica`; borrar el fichero inutilizable hace fallar la comparación de SHA-256; quitar la rama del contexto de `falloAlAplicar` hace salir `TestContextoCanceladoDuranteLaMigracion` con 1 («no se pudo aplicar la migración») en vez de 4 (SC-005, FR-003, FR-026 a FR-028) |
| 13 | **`PRAGMA`, permisos y ruta en el URI** | `TestAbrirAplicaLosPragma` (`journal_mode`, `synchronous`, `busy_timeout` = tramo de 100 ms, `query_only`), `TestAbrirCreaDirectorioYFicheroConPermisosReservados` (`0700`/`0600`), `TestRutaParaURI` y `TestDirectorioConCaracteresDeURI` (directorios con `%41`, `?`, `#`, espacio y `&`: la base dentro del directorio declarado, a `0600`, nada fuera, y un lector de solo lectura la encuentra) | `test` | Sí | Quitar `_pragma=journal_mode(WAL)` → `delete`; dejar que SQLite cree el fichero → `0644`; componer el DSN sin escapar la ruta deja `cache.db` a 0 bytes y la base en un hermano del directorio con `?` y `#`, y `New` en 1 con `%41` (SC-006, FR-020, FR-021) |
| 14 | **Concurrencia y espera ante bloqueo** | `TestDosClientesEnElMismoProceso` (`normal-normal`, `normal-solo-lectura`), `TestMismaClaveDosEscritores`, `TestSoloLecturaVeLoConfirmadoEnElWAL`, `TestClienteDesdeVariasGoroutines` bajo `-race`; `TestIntegracionDosProcesos` (dos subtests) relanzando el binario de test; `TestEsperaAQueSueltenElBloqueo`, `TestNewConLaBaseBloqueadaRespetaElContexto`, `TestPutConLaBaseBloqueadaRespetaElContexto`, `TestBloqueoQueNoSeSueltaAgotaLaEspera` con otra conexión reteniendo `BEGIN IMMEDIATE` | `test`, `test-integration` | Sí | Un presupuesto de 0, o quitar el reintento, produce `SQLITE_BUSY` en el segundo escritor y hace fallar `TestEsperaAQueSueltenElBloqueo`; subir `tramoDeEspera` a 5 s —la espera entera dentro del motor, sin tramos que miren el contexto— deja `TestNewConLaBaseBloqueadaRespetaElContexto` y `TestPutConLaBaseBloqueadaRespetaElContexto` en 5,08 s (medido), por encima del margen de 2,5 s que exigen; quitar la rama `esBloqueo` de `falloAlAbrir` o de `falloAlOperar`, o presentar `SQLITE_BUSY` como fichero inutilizable, hace fallar `TestBloqueoQueNoSeSueltaAgotaLaEspera`; abrir en solo lectura con `immutable=1` siempre hace fallar `normal-solo-lectura` (no ve el WAL) (SC-007, FR-003, FR-015, FR-031, FR-032) |
| 15 | **Permisos: directorio y fichero** | `TestIntegracionDirectorioNoEscribible` (`solo-lectura-lee`, `normal-argumentos`), `TestIntegracionDirectorioDenegado`, `TestIntegracionFicheroDenegado` (`cache.db` a `0000` en un directorio escribible: 1 en solo lectura y 2 en normal, nombrando el fichero y el acceso denegado, nunca el directorio), `TestIntegracionReaperturaInmutableFalla` (`cache.db` legible y con las páginas estropeadas, en un directorio `0500` sin `-wal`: la reapertura inmutable falla → 1 nombrando el fichero, ningún cliente, nada creado ni cambiado), `TestIntegracionWALSinMemoriaCompartida`, con precondición de que el sistema de ficheros haga valer los permisos (`t.Fatalf` en integración continua, `t.Skip` fuera de ella) y restauración de los permisos en `t.Cleanup` | `test-integration` | Sí | Sin la reapertura `immutable` ante 1544 o 14 sin `-wal`, `solo-lectura-lee` termina con 1 en vez de 0; devolver `nil` en `falloAlLeerLoInmutable` entrega un cliente sin base y `TestIntegracionReaperturaInmutableFalla` recibe un cliente y un 0 en vez del 1 (la ausencia falsa que FR-015 prohíbe); disparar la comprobación de `-wal` solo ante 1544 deja `TestIntegracionWALSinMemoriaCompartida` con un mensaje que no nombra los dos auxiliares (el caso llega con 14); degradar `ErrPermission` a ausencia hace terminar `TestIntegracionDirectorioDenegado` con 4 en vez de 1; culpar al directorio de un fichero ilegible hace fallar `TestIntegracionFicheroDenegado` (FR-015, FR-022, FR-035, SC-004, SC-011) |
| 16 | **Fixtures de códigos de salida** | `TestClasesDeError`: 13 subpruebas (filas 1-11, 14, 15 del contrato de errores §3) comprobando `cli.Clasificar` y `cli.CodigoSalida` sobre el error real; filas 12 y 13 en los tests de la fila 15 | `test`, `test-integration` | Sí | Devolver la ausencia en solo lectura como «inesperado» hace salir la fila 8 con 1 en vez de 4; pasar la rama por omisión de `falloAlOperar` a «fuente no disponible» hace salir la fila 14 y `TestFalloDelControladorAlOperar` con 4 en vez de 1 (SC-011) |
| 17 | **Binario sin `internal/cache` ni `modernc.org/*`** | `TestElBinarioNoEnlazaCache` (temporal, se retira en H4) + `TestDependenciasDelBinario` (lista sin cambios) | `test` | Sí | Importar `internal/cache` desde `internal/app` (no test) hace fallar los dos (SC-012, FR-044) |
| 18 | **Direcciones solo ficticias en tests y fixture** | `grep` del quickstart (prerrequisitos): toda dirección `http(s)://` de `internal/cache/*_test.go` y `internal/cache/testdata` es `fuente.prueba`, salvo la identificación exacta | manual | No | Cualquier `https://www.boe.es` o `127.0.0.1` en un test o fixture aparece y la confirmación no se imprime |
| 19 | **Ni caché real ni ficheros fuera de `t.TempDir()`** | medida diferencial de `ls -laR ~/.cache/kitlegal` (quickstart 1) y `TMPDIR` desechable vacío tras los tests de integración (quickstart 7); guardián de diff del workflow con `[datos]` | manual + workflow | No (el guardián sí, por tarea) | Un test que olvidara `t.Setenv("HOME", …)` cambiaría la medida; un fichero temporal no borrado deja el `find` distinto de 0 (SC-009, FR-040) |
| 20 | Formato, lint general, `-race`, `govulncheck`, `gosec`, CodeQL, `gitleaks`, `go mod verify`, `go mod tidy -diff`, pre-commit, cobertura | los de H0/H1/H2, **sin ninguna exclusión nueva** (SC-008) | varias | Sí (salvo CodeQL) | Los mismos de H2; `go mod tidy -diff` cubre la dependencia nueva; `gosec` G204/G301/G302/G304 pasan sin `//nolint` por D11, D4 y D3 |

### Fixtures

H3 crea **un** fichero bajo `internal/cache/testdata/reproduccion/prueba/`, **escrito a mano** contra el
host ficticio `fuente.prueba` (D12): es material de test —la grabación que sirve la primera consulta
del adaptador de prueba por `httpx.Replay`— y no una grabación de una fuente real. La tarea que lo crea
lleva `[datos]`; el guardián del workflow exige la etiqueta y no pausa (material nuevo bajo
`internal/<pkg>/testdata/`). **No se crea `testdata/<fuente>/` en la raíz** ni se graba nada real.

**Lista cerrada**: este fichero y ninguno más. Es la ruta que la tarea `[datos]` declara y la que el
guardián de diff comprueba. El nombre es el que produce la regla del contrato de grabación de H2 §2
(`GET` + `http` + `fuente.prueba` sin puerto + `/norma` → `_norma`), y el contenido sigue su §1:

| Fichero | Petición grabada | Respuesta grabada | Lo usa |
|---|---|---|---|
| `GET_http_fuente.prueba_norma.json` | `GET http://fuente.prueba/norma` | `"formato": 1`; `grabado_en` `2026-09-12T00:00:00Z`; `peticion.cabeceras` solo `User-Agent: kitlegal/dev (+https://ventanillalegal.es/bot)`; `200`; `Content-Length: 24`, `Content-Type: text/plain; charset=utf-8`, `Date` fija; `cuerpo` `<norma>contenido</norma>` | `TestAdaptadorDePruebaConElKernel/primera-consulta` y, como siembra, `offline-presente`; `sin-cache-la-reproduccion-falla` no lo usa (directorio vacío) |

**Tests de contrato.** Los cuatro contratos de [`contracts/`](./contracts/) tienen comprobación
mecánica: puerto y cliente (filas 6-11, 13), errores y códigos (16), esquema y apertura (12, 13, 14, 15),
reglas de arquitectura (1, 2, 3, 4, 17).

**Objetivos del `Makefile` que H3 toca**, y ninguno más: la lista de prerrequisitos de `ci` (gana
`test-integration`). Ninguna receta cambia.

## Inventario de tests

Nombres fijados aquí para que `tasks.md` y `quickstart.md` los usen tal cual. Sin guiones bajos; un
fichero de test por fichero de código y en el mismo orden que el árbol de «Project Structure»
(`golang-testing`), con las excepciones declaradas: `doc.go` e `internal/core/{doc,cache}.go` no tienen
test porque no contienen sentencias. `t.Parallel()` en todos salvo los que usan `t.Setenv`, que
`paralleltest` reconoce.

| Fichero | Tests | Cubre |
|---|---|---|
| `cliente_test.go` (`cache`) | `TestNewRechazaOpcionesInvalidas` (tabla: `ConDirectorio("")`, `ConReloj(nil)`, `ConRegistrador(nil)` → 2), `TestNewPorOmisionUsaElDirectorioDeLaCuenta` (`t.Setenv` de `HOME` y `USERPROFILE` a un temporal; aparece `<HOME>/.cache/kitlegal/cache.db`), `TestCloseEsIdempotente`, `TestOperacionTrasCierre` (`Get`/`Put` → 1, sin tocar el fichero), `TestClienteDesdeVariasGoroutines` (un cliente, N goroutines de `Get`/`Put`, `-race`) | FR-004, FR-019, FR-023, D2 |
| `ruta_test.go` | `TestRutaEfectivaPrecedencia` (`t.Setenv`: opción > variable > omisión; la variable se lee una vez), `TestRutaInservible` (tabla × modo: vacía por opción, vacía por variable, fichero como directorio → 2 nombrando origen y ruta; en solo lectura igual), `TestDirectorioNoCreable` (directorio `<fichero>/sub` cuyo padre es un fichero, `Stat` → `ENOTDIR`: subtest `normal` → 2 nombrando origen y ruta; subtest `solo-lectura` → `New` sin error, cliente sin base, `Get` → 4 y el fichero padre intacto), `TestEsInexistente` (tabla del predicado: `fs.ErrNotExist` → true, `syscall.ENOTDIR` → true, `fs.ErrPermission` → false, `nil` → false, error ajeno → false; los errores reales salen de `os.Stat` sobre rutas construidas en `t.TempDir()`) | FR-020, FR-022, FR-023, US4, D3 |
| `abrir_test.go` | `TestAbrirCreaDirectorioYFicheroConPermisosReservados` (`0700`, `0600`), `TestAbrirAplicaLosPragma` (`journal_mode = wal`, `synchronous = 2`, `busy_timeout = 100`, el tramo; en solo lectura `query_only = 1`), `TestRutaParaURI` (tabla del escape `%`, `?`, `#`), `TestDirectorioConCaracteresDeURI` (tabla de directorios con `%41`, `?`, `#`, espacio y `&`), `TestSoloLecturaSinBaseNoCreaNada` (tabla: directorio vacío, directorio inexistente y directorio cuyo padre es un fichero: listado del padre idéntico antes y después, `Get` → 4), `TestSoloLecturaBaseVacia` (0 bytes: ausencia, sin migrar, fichero intacto), `TestSoloLecturaNoModificaElFichero` (SHA-256 antes y después de `Get` presente y ausente), `TestFicheroInutilizable` (normal y solo lectura → 1; SHA-256 igual; no se borra), `TestSoloLecturaVeLoConfirmadoEnElWAL` (escritor abierto sin checkpoint; el lector `ro` ve lo confirmado y no lo pendiente) | FR-015, FR-021, FR-028, FR-030, FR-031, SC-003, SC-006, US3-6, D4, D5 |
| `migraciones_test.go` | `TestMigracionesEmbebidasBienFormadas` (numeración contigua desde `0001`; versión conocida = número de ficheros), `TestMigracionesIdempotentes` (dos aperturas: una fila en `schema_version`, versión 1), `TestVersionMayorQueLaConocida` (`INSERT INTO schema_version VALUES (99, …)` con `database/sql`; normal y solo lectura → 1; mensaje con esperada y encontrada; SHA-256 igual), `TestMigracionAtomica` (tabla `entradas` ajena preexistente: `New` → 1; sin `schema_version`; la tabla ajena sigue), `TestSoloLecturaNoMigra` (base en versión 0 sigue en 0), `TestDosClientesMigranUnaVez` (dos `New` concurrentes sobre un directorio vacío: cero errores, una versión), `TestContextoCanceladoDuranteLaMigracion` (el reloj inyectado cancela el contexto dentro de la transacción de la migración: 4, `context.Canceled` alcanzable, sin cliente, sin `schema_version`, y la siguiente apertura migra) | FR-003, FR-024 a FR-029, SC-005, US5-1 a US5-3, fila 15 del contrato de errores, D7 |
| `entradas_test.go` | `TestPutYGetIntegros` (tabla: texto, binario con bytes altos, cero bytes → `[]byte{}` presente, 1 MiB, clave larga), `TestGetAusente`, `TestPutSustituyeLaEntrada`, `TestClavesIndependientes`, `TestClaveVacia` (`Get` y `Put` → 2), `TestVigenciaInvalida` (0 y negativa → 2; nada escrito), `TestExpiracionConRelojInyectado` (antes → presente; instante exacto → ausencia; después → ausencia; duración < 1 s), `TestPutSobreEntradaCaducada`, `TestContextoCancelado` (`Get`, `Put` y `New` con contexto cancelado → 4), `TestSoloLecturaAusenciaEsFuenteNoDisponible`, `TestSoloLecturaExpiradaEsFuenteNoDisponible`, `TestSoloLecturaPutFalla` (→ 1; SHA-256 igual), `TestFalloDelControladorAlOperar` (páginas de entradas a 0xFF desde la tercera: `Get` y `Put` → 1 nombrando operación, clave y ruta, causa `SQLITE_CORRUPT`, nunca «bloqueada»; SHA-256 igual) | FR-006 a FR-014, FR-016 a FR-018, FR-028, FR-033, SC-002, SC-010, US1-4, US1-5, US2, fila 14 del contrato de errores, D8 |
| `errores_test.go` | `TestClasesDeError` (13 filas del contrato §3; la 14 incluye el bloqueo que dura más que la espera, acortada con la opción de las pruebas, y el fallo del controlador sobre páginas de entradas estropeadas en `Get` y `Put`; la 15, el contexto que vence durante la espera ante bloqueo en `New` y `Put` —cada llamada con su propio plazo— y el que se cancela durante la migración), `TestErrorMensajes` (una fila por forma de mensaje del contrato de errores §6: qué nombra cada situación —ruta, origen, variable, clave, versiones, el fichero que no se deja abrir o leer con su acceso denegado, la base bloqueada con la espera agotada y, en el fallo de WAL sin memoria compartida de la fila 13, la ruta y los dos auxiliares), `TestErrorEnvueltoConservaLaClase`, `TestErrorNuloOCero` (`Error()` no vacío, `Clase()` inesperado) | FR-033 a FR-035, SC-011, D9 |
| `espera_test.go` | `TestNewConLaBaseBloqueadaRespetaElContexto` (otra conexión retiene `BEGIN IMMEDIATE` sobre una base sin migrar; `New` con plazo de 300 ms → 4 en mucho menos de 5 s, sin cliente y sin conexión abierta: al soltar el bloqueo no quedan `-wal` ni `-shm`), `TestPutConLaBaseBloqueadaRespetaElContexto` (igual con `Put`; `Get` no espera), `TestEsperaAQueSueltenElBloqueo` (el bloqueo se suelta a los dos tramos y `New` y `Put` terminan bien), `TestBloqueoQueNoSeSueltaAgotaLaEspera` (espera acortada a 300 ms con la opción de las pruebas: 1, mensaje «bloqueada» con la ruta y la espera, causa `SQLITE_BUSY`, nunca «inutilizable») | FR-003, FR-031, FR-035, filas 14 y 15 del contrato de errores, D4, D10 |
| `concurrencia_test.go` | `TestDosClientesEnElMismoProceso` (`normal-normal`, `normal-solo-lectura`: escritor confirma N entradas y avisa por canal; el lector encuentra todas, ninguna a medias), `TestMismaClaveDosEscritores` (última escritura completa gana; el contenido leído es uno de los dos, nunca una mezcla) | FR-007, FR-032, SC-007, US5-4, D10 |
| `superficie_test.go` | `TestSuperficieExportada` | FR-005, D2 |
| `adaptador_test.go` (`cache_test`) | tipos `adaptadorDePrueba`, `argumentosConsultar`, `argumentosGuardar`, `cuerpoDeLaFuente`; `TestAdaptadorDePruebaConElKernel` con **nueve** subtests, de los que **seis y solo seis** llevan el prefijo `offline-` que filtra el escenario 4 del quickstart: `primera-consulta` (que es también lo que acredita US3 escenario 5 y FR-014: sin `--offline`, la ausencia sigue su curso hacia la fuente y el código no cambia), `segunda-consulta-sin-red`, `sin-cache-la-reproduccion-falla`, `offline-presente`, `offline-ausente`, `offline-expirada`, `offline-sin-base`, `offline-directorio-inexistente`, `offline-guardar`; esta lista es la única, y `tasks.md`, `quickstart.md` y research D12 la usan tal cual; `TestAdaptadorConVariableDeEntorno` (`t.Setenv`: base bajo la variable y no bajo `HOME`; vacía → 2; fichero → 2) | SC-001, SC-003, SC-004, SC-011, US1-1 a US1-3, US3-1 a US3-5, FR-045 a FR-047, D12 |
| `integracion_test.go` (`cache_test`, `//go:build integration`) | `TestIntegracionDosProcesos` (`escritor-padre-lector-solo-lectura-hijo`: el padre confirma N entradas y mantiene el cliente abierto; el hijo, con `SoloLectura()` y `KITLEGAL_CACHE_DIR`, las encuentra todas; `escritor-hijo-lector-padre`), `TestIntegracionDirectorioNoEscribible` (`solo-lectura-lee` con `0500` y sin auxiliares → 0; `normal-argumentos` → 2), `TestIntegracionDirectorioDenegado` (`0000` → 1 en solo lectura; 2 en normal), `TestIntegracionFicheroDenegado` (`cache.db` a `0000` en un directorio escribible → 1 en solo lectura y 2 en normal, nombrando el fichero y el acceso denegado y sin nombrar el directorio), `TestIntegracionWALSinMemoriaCompartida` (`cache.db` + `-wal` copiados a un directorio `0500` → 1), `TestIntegracionReaperturaInmutableFalla` (`cache.db` legible, en WAL, con todas las páginas menos la primera a 0xFF, en un directorio `0500` sin `-wal` → la reapertura inmutable falla: 1 nombrando el fichero, sin culpar al directorio ni afirmar un acceso denegado, sin cliente, causa `SQLITE_CORRUPT`, huellas intactas), `TestProcesoAuxiliar` (retorna sin la variable de papel); precondición común de que el sistema de ficheros hace valer los permisos: `t.Fatalf` nombrando la causa cuando `CI` no está vacía, `t.Skip` nombrando la misma causa fuera de la integración continua | FR-015, FR-022, FR-032, SC-004, SC-007, SC-009, SC-011 filas 12 y 13, D5, D11 |
| `internal/arch_test.go` | comentario de R3; `TestElBinarioNoEnlazaCache` | FR-036, FR-037, FR-044, SC-012, SC-013 |

## Orden de implementación

La constitución fija `core → adaptador → applet → skill`. En H3 no hay applet de producto ni skill: el
orden es `dominio → adaptador (de dentro afuera: errores, ruta, apertura, migraciones, entradas) →
adaptador de prueba con el kernel → integración → controles`, y en cada paso **el test de tabla se
escribe antes** que el código que lo hace pasar, con `make ci` en verde al cerrar cada tarea (regla del
modo desatendido).

1. **Dependencia**: `go get modernc.org/sqlite@v1.58.0`, `go mod tidy`, comprobar S1 (`go mod tidy -diff`
   limpio, `TestDependenciasDelBinario` en verde sin tocar su lista).
2. **Dominio** (`internal/core/doc.go`, `cache.go`): el puerto; `make lint` y el test de arquitectura en
   verde (R1 sobre el paquete nuevo).
3. **Errores y ruta** (`errores.go`, `ruta.go`, `cliente.go` con las opciones y `New` sin abrir aún):
   `errores_test.go` (salvo `TestClasesDeError`, que va en el paso 7 cuando existen sus trece
   situaciones provocables sin permisos),
   `ruta_test.go` (`TestEsInexistente` primero: fija la regla «inexistente» antes de que `abrir.go` la
   use), `TestNewRechazaOpcionesInvalidas`.
4. **Apertura y migraciones** (`abrir.go`, `migraciones.go`, `migraciones/0001_entradas.sql`, y de nuevo
   `cliente.go`: `New` pasa a abrir y migrar, `Close` a cerrar la conexión y `Cliente` gana `db` y
   `versionEsquema`): `abrir_test.go`, `migraciones_test.go`, `TestCloseEsIdempotente`,
   `TestNewPorOmisionUsaElDirectorioDeLaCuenta`.
5. **Entradas** (`entradas.go`): `entradas_test.go`, `TestOperacionTrasCierre`,
   `TestClienteDesdeVariasGoroutines`, `concurrencia_test.go`, `superficie_test.go`.
6. **Fixture** (`[datos]`): `internal/cache/testdata/reproduccion/prueba/GET_http_fuente.prueba_norma.json`.
7. **Adaptador de prueba** (`adaptador_test.go`) y `TestAdaptadorDePruebaConElKernel`,
   `TestAdaptadorConVariableDeEntorno`: el criterio de aceptación del hito con el kernel en proceso. Con
   todas las rutas de fallo ya existentes, aquí se escribe **entera** `TestClasesDeError` (es el caso en
   que test e implementación van en tareas distintas).
8. **Integración** (`integracion_test.go`), `.golangci.yml` (`run.build-tags`) y `Makefile` (`ci` con
   `test-integration`): en la misma tarea, porque el fichero etiquetado no debe existir sin que el lint
   lo alcance ni sin que CI lo ejecute.
9. **Controles y documentación**: `internal/arch_test.go` (comentario R3, `TestElBinarioNoEnlazaCache`),
   comentarios de `.golangci.yml`, `docs/PENDIENTES.md`. Validación de SC-013 **en copia desechable**
   (quickstart 9). `make ci` en verde de principio a fin.

Sin `CHANGELOG.md` ni ADR: H3 no cambia ningún comportamiento visible ni se aparta de ADR 0002 (spec,
*Fuera de alcance*).

## Complexity Tracking

> Divergencias conscientes respecto de la lectura literal del spec o del roadmap. Ninguna afecta a
> alcance, frontera humana, privacidad, términos de uso ni a una decisión cerrada. **Ninguna dependencia
> está fuera de la lista de la constitución §V**: la única nueva es la que el hito nombra.

| Divergencia | Por qué es necesaria | Alternativa más simple, y por qué se rechaza |
|---|---|---|
| **`New` recibe `context.Context`** (el hito escribe `cache.Get/Put(key, ttl)` y no dice nada del constructor) | Construir la caché crea el directorio, abre la base y migra: es una operación con entrada y salida y FR-003 exige contexto en todas y prohíbe crear uno de fondo ([D2](./research.md#d2--superficie-de-internalcache-un-tipo-un-constructor-con-contexto-cuatro-opciones-tres-métodos)). | *Abrir perezosamente en la primera operación*: mezclaría en `Get` fallos de ruta (2) y de esquema (1) con los de lectura, y obligaría a sincronizar una apertura diferida. |
| **`ConReloj` exportado** (FR-009 solo exige que el reloj se pueda inyectar al construir) | El adaptador de prueba vive en el paquete externo `cache_test` (Q4) y necesita sembrar una entrada caducada para `offline-expirada` sin esperar (SC-002); H4 tendrá la misma necesidad ([D2](./research.md#d2--superficie-de-internalcache-un-tipo-un-constructor-con-contexto-cuatro-opciones-tres-métodos)). | *Reloj interno como en H2*: el test externo solo podría caducar con una vigencia de un nanosegundo y una espera real. |
| **Reapertura con `immutable=1` cuando el directorio no admite crear `-shm` y no hay `-wal`** (la comprobación de `-wal` se dispara ante 1544 **y** ante 14, que es el código que llega cuando `-wal` existe) | Sin ella, `mode=ro` falla con `SQLITE_READONLY_DIRECTORY` en un directorio de solo lectura (sonda 3) y el caso más común de `--offline` terminaría en código 1 contra el DEBE de FR-015. Solo se activa cuando ningún escritor puede existir, y Q2 descartó `immutable` como modo general, no como contingencia ([D5](./research.md#d5--apertura-en-solo-lectura-stat-primero-modero-y-immutable1-solo-cuando-el-directorio-no-admite-escritura)). | *`immutable=1` siempre*: no vería el WAL de otra invocación (Q2). *Código 1*: incumple FR-015. |
| **`Get` en solo lectura devuelve la ausencia como error de clase «fuente no disponible»** (FR-013 dice que la ausencia no es un error de la caché) | FR-016 y Key Entities «Modo de solo lectura» dicen que en ese modo la ausencia «DEBE producir un fallo de la clase» y que «el modo … convierte la ausencia en un fallo»; hacerlo en la caché deja al adaptador sin decisión que repetir ([D8](./research.md#d8--get-y-put-validación-upsert-ausencia-expiración-y-solo-lectura)). Fuera de ese modo, FR-013 rige intacto. | *Devolver `false, nil` y que cada adaptador traduzca*: dos lecturas del spec y una decisión duplicada en cada fuente. |
| **Dos situaciones de fallo no enumeradas en FR-033**: ruta por omisión indeterminable → «argumentos» (2); contexto cancelado o vencido → «fuente no disponible» (4) | Ninguna ruta de fallo puede quedar sin clase (FR-033). La primera es corregible por quien invoca (declarar `KITLEGAL_CACHE_DIR`), como las demás de «argumentos»; la segunda es la clase que H1 define para «plazo agotado» y la que H2 dio a la misma situación, y `conPlazoAgotado` del kernel ya la impone bajo `--timeout` ([D3](./research.md#d3--ruta-efectiva-opción-variable-omisión-qué-es-inservible-y-en-qué-modo), [D9](./research.md#d9--errores-tipados-cacheerror-declara-su-clase-tabla-cerrada-de-quince-situaciones)). | *«Inesperado» para las dos*: la primera dejaría de decirle a la persona qué hacer; la segunda daría dos códigos a un mismo `--timeout` según dónde venciera. |
| **El adaptador de prueba sirve la primera consulta por `httpx.Replay` con un fixture escrito a mano**, en vez de contra un servidor local como H2 | Un servidor local exige `net/http` en `internal/cache`, que `depguard` deniega por prefijo también en los tests (R2); una exclusión sería una supresión (SC-008). El spec prevé material de reproducción escrito a mano ([D12](./research.md#d12--el-adaptador-de-prueba-applet-prueba-con-consultar-y-guardar-el-kernel-en-proceso-y-la-reproducción-en-las-dos-consultas)). | *Copiar el fixture de H2 desde `internal/httpx/testdata`*: acopla dos paquetes por sus fixtures. *Generar el JSON en el test*: duplica el contrato de formato de H2. |
| **`t.Skip` condicionado en los tests de permisos, y solo fuera de la integración continua**, cuando el sistema de ficheros no los hace valer (usuario `root`) | SC-004 declara que esas variantes «se miden donde el sistema de ficheros los hace valer»; una aserción de permisos como `root` no mide nada. El salto nombra la causa y no se produce en un puesto de desarrollo normal. En la integración continua la misma precondición incumplida es `t.Fatalf`: allí el entorno **debe** hacer valer los permisos (S3), y hacerlo fallar es la única forma de comprobarlo, porque la receta de `test-integration` es contrato de H0, corre sin `-v` y un salto no deja rastro en el registro ([D11](./research.md#d11--tests-todos-sobre-una-base-real-en-ttempdir-la-etiqueta-integration-marca-lo-que-depende-del-entorno-ci-los-ejecuta)). | *Saltar también en CI*: dejaría S3 sin comprobación observable —la tarea `[plataforma]` tendría que leer un `SKIP` que la receta no imprime—. *Fallar también en local*: convertiría un puesto de desarrollo inadecuado en un rojo falso del hito. *No comprobar permisos*: dejaría FR-015 y FR-022 sin medida. |
| **`make ci` gana `test-integration`** (la receta y la lista de `ci` son contrato de H0) | FR-041 y `docs/ROADMAP.md` §3 («todo lo que entra se queda como gate de CI»); `ci.yml` solo llama a `make ci`, así que es el único sitio donde entra sin romper la identidad local/CI ([D11](./research.md#d11--tests-todos-sobre-una-base-real-en-ttempdir-la-etiqueta-integration-marca-lo-que-depende-del-entorno-ci-los-ejecuta)). Coste: la suite unitaria se repite (≈ 22 s). | *Un job aparte en `ci.yml`*: rompe la identidad. *Acotar la receta a `./internal/cache/`*: cambia un contrato de H0 y dejará de valer en H12. |
| **`TestElBinarioNoEnlazaCache`, un test temporal** que H4 retirará | SC-012 pide que la superficie del binario no cambie; que no enlace un paquete con una decena de módulos nuevos es la forma mecánica de verlo. Cuando H4 lo enlace, el test desaparece junto con la ampliación justificada de `modulosDelBinario` ([D13](./research.md#d13--controles-mecánicos-qué-se-añade-y-qué-se-toca)). | *Solo `TestDependenciasDelBinario`*: cubre los módulos, no el paquete. |
| **Espera ante bloqueo por tramos** (revisión final): `busy_timeout` pasa de 5000 a 100 ms por intento y el cliente reintenta la sentencia mirando el contexto hasta 5 s en total (`espera.go`) | La espera de `busy_timeout` vive dentro del motor y no mira el contexto: `sqlite3_interrupt`, que es lo que el driver hace al terminar el contexto, no la corta, así que con 5000 un `Put` o un `New` con plazo de 300 ms tardaban 5 s en volver, contra FR-003 y contra lo que `doc.go` y el contrato del cliente §3 prometían (motivo [e][f] de la revisión). Con tramos de 100 ms la espera total sigue siendo la de FR-031 y termina en cuanto acaba el tramo en curso tras vencer el contexto. Solo se reintentan sentencias en autocommit y el comienzo de una transacción, donde `SQLITE_BUSY` garantiza que no se hizo nada; el presupuesto se mide con el reloj real, no con el inyectado ([D4](./research.md#d4--apertura-normal-fichero-pre-creado-a-0600-dsn-con-pragma-seguros-una-conexión), [D10](./research.md#d10--concurrencia-dos-clientes-son-dos-pools-sin-goroutines-propias--race)). | *Mantener `busy_timeout(5000)`*: la espera no respeta el contexto y no hay forma de interrumpirla desde Go. *Registrar un manejador de espera propio con `sqlite3_busy_handler`*: exige el puntero de la conexión y trampolines de `libc` que el driver no expone; frágil y con `unsafe`. *Un contexto derivado con plazo*: no cambia nada, porque el motor no lo mira. |
| **Clasificación del fallo al abrir con el contexto y el bloqueo delante** (revisión final): `falloAlAbrir`, `falloAlAplicar`, `falloAlOperar` y `falloAlLeerLoInmutable` miran `ctx.Err()` y `SQLITE_INTERRUPT` además del error del contexto, y `SQLITE_BUSY` tiene su propio mensaje | Con el contexto vencido durante la espera, la causa que llega es el `SQLITE_BUSY` de la última espera y no el error del contexto: mirar solo `errors.Is(causa, context.DeadlineExceeded)` clasificaba la fila 15 como fichero inutilizable (motivo [e][b]). Un bloqueo que agota la espera tampoco es un fichero inutilizable (FR-028, FR-035): es la fila 14 y su mensaje dice que la base está bloqueada y nombra la espera. La causa del fallo de la fila 15 lleva el error del contexto además del del driver, para que `errors.Is` alcance los dos. | *Envolver todo fallo con el contexto terminado en el kernel (`conPlazoAgotado`)*: ya lo hace bajo `--timeout`, pero el contrato de errores promete la clase desde el paquete y `TestClasesDeError` la mide sobre el error real. |
| **Diagnóstico del fichero antes que el del directorio** (revisión final): ante `SQLITE_CANTOPEN` o 1544 en solo lectura se comprueba con `os.Open` si `cache.db` se deja leer antes de mirar `-wal`; en modo normal, si `OpenFile` falla y `cache.db` ya existe, el mensaje nombra el fichero | El código del driver no distingue «el fichero no se deja leer» de «el directorio no admite crear `-shm`», y el mensaje culpaba al directorio en los dos modos con un `cache.db` a `0000` en un directorio escribible (motivo [e][b], FR-035). Comprobar primero el fichero es una lectura que no escribe nada, cabe en el modo de solo lectura y va antes de mirar `-wal` porque con `-wal` presente el mensaje de la fila 13 también culparía al directorio. Las clases del contrato no cambian: 1 en solo lectura (fila 12), 2 en modo normal (fila 6) ([D5](./research.md#d5--apertura-en-solo-lectura-stat-primero-modero-y-immutable1-solo-cuando-el-directorio-no-admite-escritura)). | *Comprobar la escribibilidad del directorio*: no es portable y escribiría en un modo que promete no escribir. *Diagnosticar después de la reapertura inmutable*: llega tarde cuando hay `-wal`, y hace un intento condenado antes de explicar. |
| **La ruta del DSN codificada para el URI** (revisión final): `rutaParaURI` escapa `%`, `?` y `#` y convierte las barras con `filepath.ToSlash` | El driver entrega el DSN `file:` entero al motor con `SQLITE_OPEN_URI`, y el motor decodifica `%HH` y corta el camino en `?` y `#`: con `ConDirectorio("…/con?x")` la base acababa en el hermano `…/con` a `0644` y `cache.db` a 0 bytes, y con `%41` en el nombre `New` fallaba con 1 (motivo [b][e]; FR-020, FR-021, FR-022, FR-028). Se escapan solo los tres caracteres que el motor interpreta, para que la ruta siga reconocible en el DSN ([D4](./research.md#d4--apertura-normal-fichero-pre-creado-a-0600-dsn-con-pragma-seguros-una-conexión)). | *`url.PathEscape` por segmento*: escapa también espacios y otros caracteres que el motor no interpreta, sin ganar nada. *Rechazar esos directorios*: FR-020 exige que la base viva donde la persona declara. |

## Obligaciones que este plan traslada a `tasks.md`

1. **Comprobar S1** al añadir la dependencia: `go mod tidy -diff` limpio, `go mod graph` sin módulos más
   allá de `modernc.org/sqlite` y los que su `go.mod` declara (con las versiones que la selección mínima
   escoja; que falten `modernc.org/fileutil` y `github.com/google/pprof` es lo esperado, sonda 6), y
   `TestDependenciasDelBinario` en verde sin tocar su lista. Los indirectos que `tidy` escriba se
   comparan con la predicción de S1 y la diferencia, si la hay, se anota en la PR.
2. **Sin `//nolint`** (SC-008): el proceso auxiliar se lanza por una función cuyo **parámetro** es el
   ejecutable (`os.Args[0]` lo pasa el test) y cuyo único argumento es el literal
   `-test.run=^TestProcesoAuxiliar$` (G204, D11); toda ruta que llegue a `os.OpenFile`, y en los tests
   a `os.ReadFile`/`os.WriteFile`, pasa por `filepath.Clean` (G304; `gosec` analiza también los tests,
   y H2 lo resolvió igual en `grabar_test.go` con el auxiliar `contenidoDe`); `MkdirAll` con `0o700`,
   `OpenFile` con `0o600` y `WriteFile` en tests con `0o600` (G301, G302, G306, D4). La tarea de
   `integracion_test.go` lo declara y `make lint` lo confirma.
3. **Comprobar S2** (`misspell`) en la primera tarea que cree ficheros de `internal/cache`; un falso
   positivo va a `ignore-rules` como en H1 y H2, y la tarea que lo asuma **declara `.golangci.yml`** entre
   sus rutas, acotado a `misspell.ignore-rules`. No se usan como palabra suelta `directorios`,
   `transaccion` ni `configuracion` (D15).
4. **Etiquetar `[datos]`** la tarea que cree `internal/cache/testdata/reproduccion/prueba/GET_http_fuente.prueba_norma.json`,
   dejando claro que es material escrito a mano contra `fuente.prueba` y no una grabación real, con el
   nombre y el formato del contrato de grabación de H2.
5. **Declarar en cada tarea sus rutas** (guardián de diff): las de H0/H1 que se tocan son exactamente
   `internal/arch_test.go`, `Makefile`, `.golangci.yml`, `docs/PENDIENTES.md`, más `go.mod` y `go.sum`.
   Ninguna tarea toca `internal/core/schema`, `internal/cli`, `internal/app`, `internal/httpx`,
   `internal/render` ni `cmd/`.
6. **Restaurar permisos en `t.Cleanup`** en toda prueba que ponga un directorio a `0500` o `0000`,
   registrando la restauración **después** de `t.TempDir()` (la limpieza de `TempDir` es `RemoveAll` y
   falla el test si no puede borrar); y comprobar la precondición de permisos antes de la aserción (D11).
7. **La tarea de `integracion_test.go` incluye `.golangci.yml` (`run.build-tags`) y `Makefile` (`ci`)**:
   un fichero etiquetado no entra sin que el lint lo alcance ni sin que CI lo ejecute (FR-041).
8. **Validar SC-013 sobre una copia desechable** fuera del repositorio (quickstart 9), nunca sobre el
   árbol de trabajo, y dejar la salida como evidencia en la PR.
9. **Medir SC-003 y SC-009 de forma diferencial** (quickstart 1, 4 y 7) tras `make ci`, y anotar el
   resultado.
10. **Anotar para H4** en `docs/PENDIENTES.md`: retirar `TestElBinarioNoEnlazaCache` y ampliar
    `modulosDelBinario` con **los módulos que `go list -deps ./cmd/kitlegal` muestre al enlazar
    `internal/cache`, justificados uno a uno** por escrito (constitución §V). La nota **no lleva una lista
    escrita a mano**: el `go.mod` del driver declara módulos (`modernc.org/fileutil`, `github.com/google/pprof`)
    que `go mod tidy` no incorpora al grafo (sonda 6 de research.md), y copiarlos fijaría en
    `docs/PENDIENTES.md` dependencias que el binario nunca enlaza. Si se quiere orientar a H4, se anota
    la salida de la sonda 6 marcada como «medida en un módulo de sonda, sustituir por la medida real».
    La clave de caché por fuente y su TTL salen de la fuente (`Source.TTL()`).
11. **Mantener los umbrales**: `internal/core` ≥ 85 % (sin sentencias nuevas), global ≥ 70 % con los
    tests unitarios de `internal/cache` (los de integración no alimentan `coverage.out`). Nunca se rebaja
    un umbral para que pase un estado.
12. **Solo el host ficticio `fuente.prueba` en tests y fixture** (control 18): la última tarea de tests
    ejecuta la orden entera de los prerrequisitos del quickstart contra el árbol y deja «solo direcciones
    ficticias» como evidencia en la PR.
13. **Comprobar S3 y S5** en la tarea `[plataforma]`: S3 se comprueba viendo el trabajo `ci` de la PR **en
    verde con `test-integration` dentro** (`gh pr checks`, `gh run list`), no leyendo el registro —la
    precondición de permisos falla con `t.Fatalf` en integración continua, así que un ejecutor privilegiado
    pondría el trabajo en rojo, y la receta sin `-v` no imprime ningún `PASS` ni `SKIP` que leer—; la
    duración de `ci` de la PR queda anotada. El cuerpo de la
    propuesta de cambio sale de un fichero del directorio del hito que la tarea nombra
    (`gates/pr-h3.md`), con la justificación de la dependencia nueva que exige la constitución §V.
14. **Mensajes y registro**: cada `cache.Error` nombra lo que la tabla del contrato de errores §3 exige
    (columna «Mensaje nombra»); `TestErrorMensajes` lo fija.

## Comprobación contra la rúbrica del juez (`juez_plan`, criterios a-j)

| Criterio | Dónde se cumple |
|---|---|
| a. constitution_check | Un ítem por principio (I-IX) y por regla de dependencia (5), con veredicto y evidencia; gates mecánicos, reglas del modo desatendido y re-evaluación tras el diseño |
| b. dependencias | Una, de la lista §V y nombrada por el hito (FR-043); sus indirectos declarados (S1); *Complexity Tracking* lo afirma |
| c. reglas_dependencia | `core` sin importaciones del módulo; `cache → core, schema`; SQLite y `database/sql` solo en `cache`; `net/http` en ningún fichero de `cache` (por eso el adaptador usa `Replay`); sin `os.Exit`; sin stdout; contrato de arquitectura |
| d. errores_exit_codes | `cache.Error` + `schema.ConClase` + `cli.Clasificar` sin cambios en el kernel; tabla cerrada de 15 situaciones → {1, 2, 4}; `TestClasesDeError` (13) + dos tests de integración; una sola regla «inexistente» (`fs.ErrNotExist` o `syscall.ENOTDIR`) para los dos `Stat` del modo de solo lectura, escrita igual en D3, D5, los tres contratos, el data-model y el inventario, de modo que ningún fallo de `Stat` tiene dos códigos ni se queda sin clase |
| e. tests_primero | Inventario de tests con nombres fijos; fixture único con lista cerrada; controles mecánicos (tabla de 20 filas); e2e: ninguno, con la razón del spec; orden de implementación con el test antes que el código |
| f. alcance | Todo lo de *Fuera de alcance* del spec queda fuera; lo único no enumerado (dos clases de fallo, `immutable` como contingencia, `ConReloj` exportado, `New` con contexto) está declarado y justificado en *Complexity Tracking* |
| g. sin_atajos | Ningún `nolint`, ninguna exclusión, ningún TODO, ningún error silenciado; G204 resuelto por parámetro y no por supresión; el único `t.Skip` es una precondición de entorno declarada —y solo fuera de la integración continua, donde la misma precondición es `t.Fatalf`—, no un test desactivado; `synchronous=FULL` en vez de `NORMAL` |
| h. mejor_alternativa | Cada decisión de research.md lleva la alternativa rechazada y el motivo; las que dependían del comportamiento del driver se decidieron por sondas ejecutadas (D4, D5, D6, D7), no por documentación de memoria |
| i. afirmaciones_verificadas | Tabla de verificación en research.md con la ruta de cada comprobación (driver: `sqlite.go`, `doc.go`, `go.mod`, cuatro sondas; `gosec`: `subproc.go`, `resolve.go`, `fileperms.go`, `readfile.go`; `sqlclosecheck`, `rowserrcheck`, `golangci-lint`, `depguard`, `paralleltest`, `misspell`: su código; biblioteca estándar: fuente y `go doc`); lo no verificado en D18 como supuesto |
| j. quickstart_ejecutable | 12 escenarios con órdenes, rutas y nombres de test reales; violaciones solo en copia desechable; medidas diferenciales sobre la caché real (`ls -laR | shasum`) y sobre un `TMPDIR` desechable; nota sobre `rtk`; ningún escenario toca el árbol, el índice ni el historial |
