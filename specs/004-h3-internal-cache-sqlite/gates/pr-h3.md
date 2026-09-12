## Objetivo

«Las normas cambian poco; no volver a pedir lo que ya tenemos» (`docs/ROADMAP.md` §4, H3). H2 dejó la
única puerta del módulo hacia la red. Faltaba **la memoria de lo que ya se pidió**: sin ella, cada consulta
de una skill volvería a pedir a un servidor público exactamente el mismo texto. H3 construye esa memoria,
una caché local en SQLite con vigencia por entrada. Es la condición previa de H4 (`boe`) y de todos los
hitos de fuente posteriores.

Tres caras, y las tres se comprueban sin red:

1. **No volver a pedir lo que ya tenemos.** La segunda consulta idéntica se sirve de lo guardado. El
   criterio de aceptación se comprueba de la forma más dura: la segunda consulta usa un `httpx.Replay`
   **estricto sobre un directorio vacío**, así que cualquier petición emitida haría fallar el test
   nombrándola. Un control negativo demuestra que ese test pasa *porque* no pide: con la caché vacía, la
   misma reproducción falla con código 1 y nombra `GET http://fuente.prueba/norma`.
2. **Lo caducado no se sirve**, tampoco sin red. El borde de la vigencia se prueba en sus tres posiciones
   (antes, en el instante exacto y después) con un reloj inyectado. El test no espera tiempo real.
3. **Trabajar sin red con lo que ya hay.** `--offline`, que H1 declaró y H2 dejó sin significado, lo
   adquiere: quien construye la caché la abre en **solo lectura**. En ese modo no crea la base, no migra y
   no escribe, y lo que falta termina en **código 4**. La caché no lee ninguna bandera.

Hito de **fundación** (principio VIII): no entrega ni cambia ninguna skill. Al cerrarlo, la superficie
visible del binario es exactamente la misma que al abrirlo. El primer applet que enlazará la caché es
`boe`, en H4.

## Alcance

- **`internal/core`**, paquete nuevo con el sitio que `docs/ROADMAP.md` §2 reserva a los puertos:
  `doc.go` y `cache.go` con la interfaz `core.Cache`. `Get` usa el idioma «coma ok» y `Put` exige
  vigencia. Importa solo `context` y `time` y no tiene sentencias.
- **`internal/cache`**, paquete nuevo (8 ficheros de producto, una migración embebida y 11 de test):
  - `cliente.go`: `New(ctx, opciones...)`, `Close` idempotente y cuatro opciones (`ConDirectorio`,
    `SoloLectura`, `ConReloj`, `ConRegistrador`). `ruta.go`: la precedencia opción >
    `KITLEGAL_CACHE_DIR` > `~/.cache/kitlegal`.
  - `abrir.go`: directorio `0700` y fichero `0600`; la ruta entra en el DSN codificada para el camino
    del URI (`%`, `?` y `#`). En modo normal, `journal_mode=WAL`, `synchronous=FULL`,
    `busy_timeout=100` —el tramo de cada intento— y `_txlock=immediate`. En solo lectura, `mode=ro` y
    `query_only`, más la reapertura `immutable=1` como contingencia. Los mensajes culpan a quien toca:
    al fichero cuando es él el que no se deja abrir o leer, al directorio solo cuando lo es.
  - `espera.go`: la espera ante bloqueo, hasta 5 s en total, hecha de reintentos por tramos de 100 ms
    que miran el contexto entre tramo y tramo, porque la espera de `busy_timeout` vive dentro del motor
    y no lo mira. Un plazo que vence durante la espera es «fuente no disponible» (4) en cuanto acaba el
    tramo en curso; un bloqueo que agota la espera es «inesperado» (1) diciendo que la base está
    bloqueada, nunca «inutilizable».
  - `migraciones.go` y `migraciones/0001_entradas.sql`: `embed` y `schema_version`. Cada migración va en
    una transacción inmediata. Una versión desconocida o un fichero inutilizable se rechazan sin
    modificar ni borrar nada.
  - `entradas.go`: `Get` y `Put`, con `var _ core.Cache = (*Cliente)(nil)`.
  - `errores.go`: `cache.Error` implementa `schema.ConClase`. Tabla cerrada de **quince** situaciones →
    códigos `{1, 2, 4}`.
- **Un solo fichero de reproducción**, escrito a mano contra el host ficticio `fuente.prueba`:
  `internal/cache/testdata/reproduccion/prueba/GET_http_fuente.prueba_norma.json`. **Ninguna grabación
  contra una fuente real**; `KITLEGAL_RECORD` no se ejecutó en ninguna tarea.
- **Ficheros de H0/H1/H2 tocados, y ninguno más**:
  - `Makefile`: solo la línea de prerrequisitos, `ci: fmt-check lint test test-integration vuln
    schema-check secrets mod-verify mod-tidy-check`. Ninguna receta cambia.
  - `.golangci.yml`: `run.build-tags: [integration]`, comentarios de la lista `sql` y dos palabras
    españolas bajo `misspell.ignore-rules`.
  - `internal/arch_test.go`: comentario de R3 y `TestElBinarioNoEnlazaCache`.
  - `docs/PENDIENTES.md`, `go.mod` y `go.sum`.
  - No se tocan `internal/core/schema`, `internal/cli`, `internal/app`, `internal/httpx`,
    `internal/render` ni `cmd/`.
- **Artefactos del hito**: `specs/004-h3-internal-cache-sqlite/` completo (spec, plan, research con sus
  apartados D1 a D18, data-model, cuatro contratos, quickstart, tasks y `gates/`).

**Fuera de alcance, y por qué no aplican seis puntos de la Definition of Done**:

- Punto 4: no hay applet ni salida de applet, y `schemas/` no se toca.
- Punto 6: ningún comportamiento visible cambia, así que no hay guion e2e ni `CHANGELOG.md`. Lo que gana
  `--offline` se comprueba con el kernel **en proceso**.
- Punto 7: no hay ADR nuevo. SQLite sin cgo ya es ADR 0002, y el patrón *Repository* está en el roadmap
  §2.
- Punto 8: no se consulta ninguna fuente real.
- Punto 10: no hay skill.
- Punto 11: no hay dimensión territorial.

Tampoco se abren `internal/source/`, `internal/store/` ni `internal/graph/`.

## Dependencias (constitución §V)

**Una** entrada directa nueva en `go.mod`, **de la lista fijada** de la constitución §V y de
`docs/ROADMAP.md` §3. La nombra el propio hito (FR-043), así que no requiere *Complexity Tracking*:

| Dependencia | Versión | Para qué | Alternativa descartada |
|---|---|---|---|
| `modernc.org/sqlite` | v1.58.0 | SQLite 3.53.4 traducido a Go, **sin cgo** (ADR 0002), para `CGO_ENABLED=0` y `-trimpath`. Aporta los códigos de error extendidos (`SQLITE_READONLY_DIRECTORY`, `SQLITE_CANTOPEN`, `SQLITE_NOTADB`) que la apertura en solo lectura necesita distinguir | `mattn/go-sqlite3`: exige cgo y rompe la compilación cruzada del binario único. Un almacén propio en ficheros: sin transacciones, sin WAL y sin lectores concurrentes que vean lo confirmado |

**Indirectos que escribió `go mod tidy`, frente a la predicción del supuesto S1** (detalle en
`specs/004-h3-internal-cache-sqlite/gates/s1-dependencias.md`):

- **Acertó**: los nueve indirectos que la sonda 6 había medido entraron exactamente, con las mismas
  versiones. Son `modernc.org/libc v1.75.6`, `modernc.org/mathutil v1.7.1`, `modernc.org/memory
  v1.12.1`, `golang.org/x/sys v0.47.0` (sube desde v0.26.0, como S1 anticipaba),
  `github.com/dustin/go-humanize v1.0.1`, `github.com/google/uuid v1.6.0`, `github.com/mattn/go-isatty
  v0.0.24`, `github.com/ncruces/go-strftime v1.0.0` y `github.com/remyoudompheng/bigfft
  v0.0.0-20230129092748-24d4a6f8daec`. Tampoco entraron, como predijo la sonda, `modernc.org/fileutil`
  ni `github.com/google/pprof`: el `go.mod` del driver los declara, pero ningún paquete alcanzable los
  importa.
- **La diferencia**: `golang.org/x/tools` **sube de v0.26.0 a v0.48.0**. No es un módulo nuevo: ya era
  indirecto del repositorio por `github.com/rogpeppe/go-internal`. La versión mayor la pide
  `modernc.org/libc v1.75.6` en su `go.mod` (`go mod graph`: `modernc.org/libc@v1.75.6
  golang.org/x/tools@v0.48.0`). La sonda no pudo verlo porque su módulo no tenía ningún otro
  `x/tools` con el que hacer selección mínima. Queda dentro de lo que S1 admitía: nada más allá de lo
  que declara el `go.mod` del driver.

**Nada llega al binario distribuido en H3**: ningún paquete de producción importa `internal/cache`. Lo
fijan `TestDependenciasDelBinario`, con su lista `modulosDelBinario` intacta, y
`TestElBinarioNoEnlazaCache`. `go mod tidy -diff`, `go mod verify` y `govulncheck` están en verde.

## Controles añadidos

Lo que hasta aquí era disciplina pasa a ser mecánico:

- **La regla R3 estrena dueño.** «Solo `internal/{cache,store,graph}` importan SQLite y `database/sql`»
  estaba activa y vacía desde H0. Ahora `internal/cache` es su primer dueño real y la vigilan las dos
  capas: `depguard`, lista `sql`, y la subprueba R3 de `internal/arch_test.go`. `duenoObligatorio` sigue
  en `false`, con la razón escrita: la exigencia pide los tres dueños, y `store` y `graph` llegan en H12
  y H17.
- **Los tests de integración son gate.** `make ci` incluye `test-integration` entre `test` y `vuln`. Con
  `run.build-tags: [integration]`, además, `sqlclosecheck`, `rowserrcheck` y el resto de linters alcanzan
  el fichero etiquetado. Sin esa clave, un fichero etiquetado que ni compila pasaría el lint en verde
  (demostrado en la contraprueba de la variante c).
- **`TestIntegracionDosProcesos`**: el segundo proceso es el propio binario de test relanzado. El lector
  encuentra todas las entradas confirmadas, ninguna a medias y sin fallar por bloqueo. Los tests de
  permisos cubren las dos filas de la tabla de errores que dependen del sistema de ficheros:
  - `TestIntegracionDirectorioNoEscribible`: directorio `0500`, lee gracias a `immutable=1`.
  - `TestIntegracionDirectorioDenegado`: directorio `0000`; solo lectura → 1 y nunca una ausencia falsa.
  - `TestIntegracionFicheroDenegado`: `cache.db` a `0000` en un directorio escribible → 1 en solo
    lectura y 2 en normal, nombrando el fichero y el acceso denegado, nunca el directorio.
  - `TestIntegracionReaperturaInmutableFalla`: `cache.db` legible y con las páginas estropeadas en un
    directorio `0500` → la reapertura `immutable=1` vuelve a fallar → 1 nombrando el fichero, ningún
    cliente y nada creado ni cambiado.
  - `TestIntegracionWALSinMemoriaCompartida`: `-wal` sin `-shm` → 1, nombrando los dos auxiliares.
- **`TestSuperficieExportada`** recorre con `go/parser` todas las declaraciones exportadas. Falla si
  alguna nombra `sql` o `sqlite`, o si aparece una fuera de la lista del contrato. Que no se pueda
  ejecutar SQL arbitrario desde fuera es **por construcción**, y este test lo vigila.
- **`TestClasesDeError`**: trece subpruebas deterministas, una por fila de la tabla cerrada. Cada una
  comprueba clase y código sobre el error tal como sale del paquete **y** envuelto con `%w`, y que nunca
  se producen 3, 5 ni 6 ni un `panic`. Las dos filas restantes las cubren los tests de integración.
- **`TestAdaptadorDePruebaConElKernel`**: `app.Main` en proceso sobre un adaptador de prueba, que solo
  existe en `package cache_test`. Nueve subtests, entre ellos `segunda-consulta-sin-red`, su control
  negativo y seis `offline-*`. Estos comparan el SHA-256 de `cache.db` y el listado del directorio antes
  y después.
- **`TestElBinarioNoEnlazaCache`** (temporal): `go list -deps` confirma que el binario no enlaza
  `internal/cache` ni `modernc.org/*`. Su retirada ya está anotada para H4.
- **SC-013, demostrado, no afirmado.** El escenario 9 del quickstart se ejecutó en cuatro variantes
  sobre copias desechables fuera del repositorio:
  - a: `database/sql` en `internal/app` → `depguard` **y** R3.
  - b: el dominio importa la caché → `depguard` **y** R1.
  - c: un `*sql.Rows` sin cerrar en un fichero `integration` → `sqlclosecheck`, que solo lo alcanza
    gracias a `run.build-tags`.
  - d: `net/http` en un test de la caché → `depguard` R2, también por prefijo.

  Cada variante hace fallar `make ci` nombrando la regla.

## Evidencia

- `specs/004-h3-internal-cache-sqlite/gates/evidencia-sc013.md`: los controles fallan ante cada intento
  de saltárselos (T012). Incluye un hallazgo de método. Con las sondas originales, `make ci` fallaba pero
  nombraba a `revive` o `rowserrcheck` en vez de la regla, porque golangci-lint deduplica los hallazgos
  por línea (`uniq-by-line`). Se corrigieron las sondas del quickstart, no los controles.
- `specs/004-h3-internal-cache-sqlite/gates/evidencia-cierre.md`: validación agregada (T013). Cubre los
  escenarios 1, 4, 7, 10, 11 y 12 con `make ci` sin caché de tests y `test-integration` ejecutado de
  verdad. También la medida diferencial de `~/.cache/kitlegal` (idéntica en toda la sesión), un `TMPDIR`
  desechable con 0 ficheros tras la integración, una ayuda del binario idéntica byte a byte a la de H2 y
  «solo direcciones ficticias». Cada medida lleva su sonda positiva.
- `specs/004-h3-internal-cache-sqlite/gates/s1-dependencias.md`: los indirectos medidos frente a la
  predicción de S1.

| Umbral | Exigido | Medido (local, `go tool cover`) |
|---|---|---|
| Global | ≥ 70 % | **92,9 %** |
| `internal/core/**` | ≥ 85 % | **90,1 %** |
| `internal/cli` | ≥ 90 % | **98,6 %** |
| `internal/cache` | — | 87,8 % |

Ninguno se rebajó: `codecov.yml` no aparece en el diff frente a `main`. Sin componente nuevo de Codecov
(S4): rigen los umbrales generales. `internal/core` no gana sentencias.

**Sin ninguna supresión nueva** (SC-008): `0` `//nolint` añadidos frente a `main`. `gosec` se resuelve
sin supresiones: G204 con el ejecutable por parámetro, G304 con `filepath.Clean`, y G301, G302 y G306 con
`0o700` y `0o600`.

**Aviso de método, por si se repite la medida.** El envoltorio de terminal de esta máquina reescribe la
salida de `go test`, `git status --porcelain`, `git diff` y `gh pr checks`. Una medida diferencial hecha
sobre esas salidas **sale verde siempre**. Todo lo de arriba está medido con el paso directo.

## Decisiones

Todas están razonadas en `specs/004-h3-internal-cache-sqlite/research.md` (D1 a D18). Las que cambian algo fuera
del paquete, o se apartan de la letra del roadmap (*Complexity Tracking* del plan):

- **El puerto vive en el dominio (D1).** `core.Cache` no tiene `Close`, y el dominio no sabe que hay
  SQLite detrás. Nace el paquete `core`, sin sentencias. No exige ADR: el patrón *Repository* y la
  interfaz `Cache` en `core` ya estaban decididos.
- **`New` recibe `context.Context` (D2).** Construir la caché crea el directorio, abre y migra, y FR-003
  exige contexto en toda operación con entrada y salida. Alternativa rechazada: abrir en la primera
  operación, que mezclaría en `Get` los fallos de ruta (2) y de esquema (1) con los de lectura.
- **`ConReloj` exportado (D2).** El adaptador de prueba vive en `cache_test` y necesita sembrar una
  entrada caducada sin esperar. H4 tendrá la misma necesidad.
- **`synchronous=FULL`, no `NORMAL` (D4).** Integridad antes que velocidad en un fichero que sobrevive a
  muchas versiones del binario. No existe ninguna opción para bajar la sincronización ni para desactivar
  WAL.
- **Solo lectura con `mode=ro` e `immutable=1` solo como contingencia (D5).** `mode=ro` ve lo que otra
  invocación ya confirmó en el WAL. Cuando el directorio no admite crear `-shm` (SQLite devuelve 1544 o
  14) y **no hay `-wal`**, se reabre con `immutable=1`: en esa condición ningún escritor puede existir.
  Sin la reapertura, el caso más común de `--offline`, un contenedor de solo lectura, terminaría en
  código 1. Con `-wal` presente, el resultado es 1 nombrando la ruta y los dos auxiliares. `immutable`
  como modo general se descartó porque no vería el WAL. Decidido con sondas ejecutadas, no por
  documentación de memoria.
- **Migraciones propias, sin herramienta (D7).** `embed`, `schema_version` y una transacción inmediata
  que relee la versión: dos clientes concurrentes migran una sola vez. Se rechazó la sugerencia de la
  skill `golang-database` de usar una herramienta de migraciones, por la constitución §V.
- **En solo lectura, la ausencia es «fuente no disponible» (4) dentro de la caché (D8).** Así lo exigen
  FR-016 y la entidad «modo de solo lectura», y cada adaptador no tiene que repetir la traducción. Fuera
  de ese modo, la ausencia sigue siendo `nil, false, nil` (FR-013).
- **Dos situaciones de fallo que FR-033 no enumeraba (D3, D9)**, para que ninguna ruta quede sin clase:
  - Ruta por omisión indeterminable → «argumentos» (2): quien invoca lo corrige declarando
    `KITLEGAL_CACHE_DIR`.
  - Contexto cancelado o vencido → «fuente no disponible» (4), la clase que H1 y H2 dan al plazo agotado.
- **El adaptador de prueba sirve también la primera consulta por `httpx.Replay` (D12).** Un servidor
  local exigiría `net/http` en `internal/cache`, que `depguard` deniega por prefijo también en los tests,
  y una exclusión sería una supresión.
- **`t.Skip` condicionado solo fuera de la integración continua (D11).** Los tests de permisos comprueban
  primero que el sistema de ficheros los hace valer. En un puesto de desarrollo privilegiado saltan
  nombrando la causa. En integración continua (`CI` no vacía), la misma precondición incumplida es
  `t.Fatalf`: un ejecutor `root` pone el trabajo `ci` en rojo en vez de pasar de largo. Es el único
  `t.Skip` del hito.
- **La contingencia de `misspell` se resolvió sin atajo (S2).** `versiones` es el plural que el mensaje de
  versión desconocida debe nombrar. `reproduccion` la fijan dos contratos: el tramo del directorio de
  grabaciones de H2 y el nombre del subtest `sin-cache-la-reproduccion-falla`. Como no podía reescribirse,
  se detuvo T008 y se redelimitó para declarar `.golangci.yml`, en vez de un `//nolint`.

## Pendientes

- **Para H4, ya anotado en `docs/PENDIENTES.md`:**
  - Retirar `TestElBinarioNoEnlazaCache` cuando el primer applet enlace la caché.
  - Ampliar `modulosDelBinario` con los módulos que muestre entonces `go list -deps` sobre el binario,
    **justificados uno a uno** y sin copiar ninguna lista escrita a mano.
  - La clave de caché y su vigencia salen de cada fuente (`Source.TTL()`), no de un valor global.
- **Para la revisión humana** (propuesta de `evidencia-sc013.md`, fuera del alcance de H3): fijar
  `issues.uniq-by-line: false` en `.golangci.yml`, para que el primer `make ci` nombre todas las reglas
  incumplidas de una línea aunque otro linter la marque también. Hoy `TestArquitectura` nombra R3 y R1
  en cualquier caso.
- **Estados de esta propuesta**, leídos tras abrirla (detalle en
  `specs/004-h3-internal-cache-sqlite/gates/evidencia-plataforma.md`):
  - **S3 se cumple**: el trabajo `ci` está en verde con `test-integration` dentro (run `34718912959`). La
    precondición de permisos es `t.Fatalf` bajo `CI`, así que el ejecutor no es privilegiado.
  - **S5, anotado**: 6 m 19 s **en frío**. `go.sum` cambia, así que la caché de `setup-go` tiene otra
    clave. `test-integration` cuesta 46 s en la integración continua. La medida en caliente llega con el
    siguiente push de la rama.
  - **Cobertura de la Definition of Done, medida y en verde**: `codecov/project` 90,99 % (≥ 70 %),
    `internal/core` 90,36 % (≥ 85 %) e `internal/cli` 98,07 % (≥ 90 %). Ningún verde vacío.
- **Decisión humana antes de fusionar: `codecov/patch` está en rojo**, con `85.64% of diff hit (target
  92.47%)`. `codecov.yml` no declara ningún estado `patch`: Codecov aplica su objetivo por omisión
  (`auto`, la cobertura de la base), que sube con `main` (83,33 % en H1, 89,52 % en H2). No lo pide la
  Definition of Done, y el spec dice que `internal/cache` no tiene umbral propio. Parte del hueco es
  código que solo ejercitan los tests de integración, que por el plan no alimentan `coverage.out`: con
  ellos, `internal/cache` pasa de 87,8 % a 91,7 %. El resto son ramas defensivas sin test. Este hito no
  lo resuelve por su cuenta, porque es una decisión sobre qué estados son gate (H0, FR-029). Tampoco se
  rebaja nada. Opciones, con su coste en la evidencia:
  1. Declarar `patch` en `codecov.yml` con objetivo fijo o `informational`.
  2. Tests unitarios para las ramas defensivas de `abrir.go` y `migraciones.go`, en una tarea nueva de
     H3.
  3. Subir también la cobertura de integración, lo que cambia un contrato de H0.
- **Revisión pendiente del ritual** (§6 del roadmap, punto 4): `/code-review` y `/security-review` sobre
  esta propuesta. No toca `docs/SOURCES.md` porque no hay fuente.
- **La fusión es humana.** Squash-merge. El gancho `pre-push` rechaza `main`, los push forzados, los
  borrados y las etiquetas (ADR 0007).

🤖 Generated with [Claude Code](https://claude.com/claude-code)
