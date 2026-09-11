# Implementation Plan: H1 · Kernel CLI: multicall, flags globales, exit codes, sobre de salida

**Branch**: `h1-kernel-cli-multicall` | **Date**: 2026-09-11 | **Spec**: [spec.md](./spec.md)

**Input**: Feature specification from `/specs/002-h1-kernel-cli-multicall/spec.md`

**Modo**: desatendido. Las decisiones técnicas se tomaron con «Criterio de decisión autónoma» de
`.specify/memory/constitution.md` y están registradas —decisión, alternativas y motivo— en
[research.md](./research.md). Lo que no se pudo verificar sin red está declarado como **supuesto no
verificado** en [research.md D24](./research.md#d24--supuestos-no-verificados-por-falta-de-red), no como
hecho.

## Summary

H0 dejó un repositorio blindado cuyo binario solo sabe decir su versión. H1 escribe **una vez** el patrón
que repetirán los veinticinco hitos siguientes: **añadir un applet debe ser declarar un comando y devolver
un `data`**, y todo lo demás —banderas, sobre, huella, renderizado, códigos de salida, registro de
eventos— venir dado.

El enfoque técnico se apoya en cuatro decisiones que sostienen el resto:

1. **El despacho ocurre antes que el análisis** ([D1](./research.md#d1--kong-gramática-por-invocación-construida-tras-resolver-el-applet)).
   `internal/app` elige el applet por `os.Args[0]` o por el primer argumento; solo entonces `internal/cli`
   construye **una** gramática de Kong —las ocho banderas globales embebidas más los verbos de ese
   applet— y la analiza. Así las globales se declaran en un único sitio y ningún applet puede
   redefinirlas, y la superficie de Kong de la que depende el kernel queda en lo mínimo.
2. **El contrato del applet no menciona ni banderas, ni sobre, ni exit codes, ni presentación**
   ([D3](./research.md#d3--contrato-del-applet-applet-verbo-argumentos-resultado)): `Applet` declara
   nombre, descripción y verbos; cada verbo declara su `struct` de argumentos, el tipo de su `data` y si
   es el verbo que se asume cuando no se nombra ninguno
   ([D26](./research.md#d26--verbo-por-omisión-cómo-funciona-kitlegal-echo-hola-sin-nombrar-verbo), que es
   lo que hace válida la entrega literal `kitlegal echo hola`); su método `Ejecutar` recibe el contexto de
   ejecución y el registrador ya montado y devuelve `{Procedencia, Datos}`. Eso es SC-010 por
   construcción.
3. **El sobre lo monta un único punto, también cuando falla** ([D7](./research.md#d7--data-del-sobre-de-fallo-dos-claves-clase-y-mensaje),
   [D8](./research.md#d8--clases-de-error-sentinelas-y-traducción-exhaustiva-el-inesperado-sale-con-1)): el
   mismo sitio que traduce el error a código de salida emite el sobre de fallo, con `ok: false` y la causa
   (`clase`, `mensaje`) dentro de `data`. La representación es idéntica para todos los applets y para los
   fallos anteriores a llegar al applet.
4. **Las reglas de arquitectura se hacen cumplir en dos capas independientes**
   ([D18](./research.md#d18--reparto-de-las-reglas-de-arquitectura-depguard-forbidigo-e-internalarch_testgo)):
   `depguard` + `forbidigo` en el lint y `internal/arch_test.go` sobre el grafo real de `go list -deps`.
   Una configuración se puede borrar; un test que falla, no.

Todo el código nuevo vive en `internal/{core/schema,cli,app,render}` más `internal/arch_test.go` y el
material de test bajo `internal/app/testdata/`. `cmd/kitlegal/main.go` se reduce a inyectar los escritores
y el registro de producción, que en H1 está **vacío**: el primer applet de producto llega en H4.

## Technical Context

**Language/Version**: Go 1.27, sin cambios respecto de H0. `go.mod` mantiene `go 1.27.0` + `toolchain
go1.27.1` y el `Makefile` sigue derivando de ahí el `GOTOOLCHAIN` que exporta a todas las recetas
([H0 research D1](../001-h0-esqueleto-del-repo/research.md)). H1 **no toca** ninguna de las dos
directivas. Verificado en local: `go version` → `go1.27.1 darwin/arm64`.

**Primary Dependencies**: cinco módulos nuevos, todos de la lista cerrada de la constitución §V y de
`docs/ROADMAP.md` §3 ([D22](./research.md#d22--dependencias-nuevas-cuáles-y-cómo-se-fijan)):
`github.com/alecthomas/kong` (análisis de la línea de órdenes) e `github.com/invopop/jsonschema`
(generación del esquema de `--describe`) entran en el binario; `github.com/stretchr/testify`,
`github.com/rogpeppe/go-internal` (testscript) y `github.com/santhosh-tekuri/jsonschema/v6` (validación
formal) solo en tests. **Ninguna otra.** Con ellas aparece por primera vez un `go.sum` en la raíz, lo que
cierra la divergencia que H0 registró en su *Complexity Tracking*.

De las cinco, cuatro están en el módulo local y se pudieron verificar leyendo su código
(`invopop/jsonschema@v0.14.0`, `rogpeppe/go-internal@v1.16.0`, `stretchr/testify@v1.12.1`,
`santhosh-tekuri/jsonschema/v6@v6.0.3`). **Kong no está y su descarga quedó denegada**: todo lo que este
plan supone sobre su API está declarado como supuesto no verificado, con comprobación y contingencia, en
[research.md D24](./research.md#d24--supuestos-no-verificados-por-falta-de-red). Ninguna de las
contingencias cambia el diseño; solo la mecánica interna de `internal/cli`.

**Herramientas de control**: las cuatro de H0, sin cambios de versión (`golangci-lint` v2.13.2 con
`depguard` v2.2.1 y `forbidigo` v2.3.1 dentro, `govulncheck`, `gitleaks` v8.30.1, `lefthook` v1.13.6).
H1 **no añade ninguna herramienta**: activa dos linters que ya venían incluidos y **retira** una exclusión
de `errcheck` heredada de H0 —la de `fmt.Fprint*`— porque desde este hito tapa un camino que el contrato
obliga a vigilar ([D27](./research.md#d27--la-exclusión-de-errcheck-para-fmtfprint-se-retira)). Retirar
una exclusión endurece el control: no es una excepción nueva de las que FR-057 prohíbe.

**Storage**: N/A. H1 no abre ninguna base de datos ni escribe en disco fuera del directorio de
construcción y de los temporales de los tests (FR-059). SQLite entra en H3.

**Testing**: `go test -race -shuffle=on -coverprofile=coverage.out ./...` (unitarios, table-driven,
`t.Parallel()`, `testify` como ayudante) más el e2e con `testscript` contra un binario construido por el
propio test en un directorio temporal
([D20](./research.md#d20--test-e2e-con-testscript-binario-real-construido-por-el-test)). `make test-e2e`
deja de ser un anuncio y pasa a ejecutar ese paquete. Toda salida de applet se valida contra su
descripción formal con `santhosh-tekuri/jsonschema/v6` **con `AssertFormat()`**
([D13](./research.md#d13--validación-formal-en-tests-santhosh-tekurijsonschemav6-con-assertformat)); sin
esa llamada las aserciones de `format` son anotaciones y un `url` vacío pasaría (verificado en
`compiler.go:47-53`). **Cero red en todos los tests.**

**Target Platform**: binario multiplataforma sin cgo (`CGO_ENABLED=0`, `-trimpath` en `build` e
`install`, como en H0). El e2e construye su binario con las mismas restricciones no impuestas: usa el
`go build` por omisión porque lo que prueba es comportamiento, no distribución. CI en `ubuntu-latest`;
desarrollo verificado en darwin/arm64. El despacho por `os.Args[0]` normaliza el sufijo `.exe` para que el
mismo código valga en windows.

**Project Type**: herramienta de línea de órdenes multicall. H1 es el hito en que el multicall deja de ser
una decisión escrita y pasa a ser código.

**Performance Goals**: ninguna meta de rendimiento propia. Se conserva la de H0: el flujo `ci` de una PR
limpia por debajo de **3 minutos** con la caché poblada. El e2e añade una compilación por ejecución del
paquete; si esa compilación empujara el flujo por encima del umbral, la contingencia está en
*Obligaciones que este plan traslada a `tasks.md`*.

**Constraints**: `make ci` sigue sin modificar ningún fichero del árbol de trabajo (SC-013). El código de
H1 no hace ninguna petición de red (FR-058), no escribe fuera del directorio de construcción ni de los
temporales de test (FR-059), no contiene ningún `panic` en rutas de usuario (FR-033) y no introduce ningún
código de salida fuera de los reservados más el 1 de FR-031
([D8](./research.md#d8--clases-de-error-sentinelas-y-traducción-exhaustiva-el-inesperado-sale-con-1)).

**Scale/Scope**: cuatro paquetes nuevos bajo `internal/`, un test de arquitectura, dos applets de ejemplo
y sus guiones de e2e. Del orden de 1000-1400 líneas de Go entre producto y test. El valor del hito está en
que ese código no se vuelva a escribir nunca más.

## Constitution Check

*GATE: debe pasar antes de la fase 0 y volver a evaluarse tras la fase 1.*

### Principios

| # | Principio | Cómo lo cumple H1 | Veredicto |
|---|---|---|---|
| **I** | Fuentes públicas y frontera humana | H1 no toca ninguna fuente: FR-058 prohíbe toda petición de red y el test de arquitectura lo hace mecánico, porque **ningún paquete de H1 puede importar `net/http`** (solo `internal/httpx`, que no existe todavía). La frontera humana se hace *observable* aquí: el código de salida **6** y el error tipado `ErrIdentidadHumana` nacen en este hito, con al menos un caso de prueba que lo fuerza (SC-006), aunque lo que lo produzca llegue en H25. Ningún applet de ejemplo realiza ninguna acción con efectos externos. | ✅ Cumple |
| **II** | Nada sin cita ni fuente | Es **el** principio que H1 convierte en contrato ejecutable. El sobre `{ok, fuente, url, fecha_consulta, hash, data}` se define en `internal/core/schema` como único lugar del que dependen todos los applets (FR-015); `fuente` y `url` nunca van vacías, ni en éxito ni en fallo (FR-016, FR-045), y la validación formal **rechaza** un `url` vacío o no-URI (FR-017, verificado en `format.go:535`). Un applet sin fuente externa usa el espacio de nombres reservado `kitlegal.` / `kitlegal:`, que declara explícitamente que el resultado es calculado y no una cita de fuente pública ([D6](./research.md#d6--espacio-de-nombres-reservado-kitlegal--kitlegal)). H1 no emite ningún contenido legal, así que no hay nada que inventar. | ✅ Cumple |
| **III** | Tests primero y offline | El hito empieza por el guion de `testscript` que describe la entrega (orden de implementación, paso 1), como manda la constitución. Todos los tests son offline: no hay red, no hay fixtures que grabar y `KITLEGAL_RECORD=1` no tiene objeto. Toda salida de applet se valida contra su descripción formal en test (FR-017, SC-015). Cobertura: `internal/cli` ≥ 90 % declarado como componente bloqueante de Codecov ([D21](./research.md#d21--cobertura-de-internalcli--90--dónde-se-declara-el-umbral)), y los umbrales de H0 (`internal/core/**` ≥ 85 %, global ≥ 70 %) siguen vigentes sin exclusiones nuevas (FR-056, FR-057). | ✅ Cumple |
| **IV** | Arquitectura hexagonal con reglas ejecutables | Es el otro principio que H1 materializa. Dominio puro en `internal/core/schema`: sin I/O, sin Kong y **sin `log/slog`** —el `*slog.Logger` no viaja dentro de `schema.Contexto` sino como parámetro explícito de `Ejecutar` ([D2](./research.md#d2--reparto-entre-internalcli-internalapp-e-internalcoreschema-sin-ciclos))—, y la pureza deja de ser una afirmación: `depguard` deniega bajo `internal/core/**` los ocho paquetes internos **y** `log`, `log/slog`, `os`, `io`, `net/http` y `database/sql`. Adaptador de presentación en `internal/render`; composición manual en `internal/app`, sin framework de inyección. Las cinco reglas de dependencia pasan de «aún no tienen objeto» a **activas y vigiladas en dos capas** (`depguard` + `forbidigo` + `internal/arch_test.go`, [D18](./research.md#d18--reparto-de-las-reglas-de-arquitectura-depguard-forbidigo-e-internalarch_testgo)). Errores tipados con `errors.Is`, traducción **exhaustiva** vigilada por el linter `exhaustive` y **un único punto de salida por binario**. Ningún `panic` en rutas de usuario (FR-033); el único pánico posible es el de un registro mal construido al arrancar, que es un defecto de compilación y no una ruta de usuario ([D17](./research.md#d17--precedencia-del-despacho-nombres-no-registrados-y-registro-que-se-rechaza)). | ✅ Cumple |
| **V** | Simplicidad y dependencias fijadas (YAGNI) | Las cinco dependencias nuevas están **todas** en la lista cerrada de §V; ninguna otra entra, y *Complexity Tracking* no contiene ninguna dependencia. Sin framework de DI (composición manual en `internal/app`), sin ORM, sin generador de CLI. `internal/cli` **no** se extrae a librería y ni siquiera se plantea: la regla de los tres applets no se cumple hasta H11 y `echo` no cuenta (no se registra en el binario distribuido). Todo lo que el spec lista en «Fuera de alcance» queda fuera del plan: no hay `httpx`, ni caché, ni grafo, ni `schemas/`, ni markdown, ni `pkg/legalkit`. `--offline`, `--no-graph` y `--asunto` se aceptan y se propagan, y no se les inventa semántica. | ✅ Cumple |
| **VI** | Un binario, convenciones de agente | Nombre único `kitlegal`; multicall por `os.Args[0]` **o** primer argumento, con precedencia fijada y comprobable (FR-004); las ocho banderas globales declaradas una sola vez en `cli.Globales` y heredadas por todo applet sin que ninguno las declare (FR-018, SC-010); `--describe` emite el esquema JSON del que en H11 y H22 se generarán las tools MCP y las tablas de comandos. Skills: ninguna en H1 (H5), y por eso `make skills-sync` sigue anunciando el hito que las aporta. | ✅ Cumple |
| **VII** | Grafo y privacidad | H1 no crea ninguna base de datos, ningún nodo y ninguna arista: `--no-graph` y `--asunto` se aceptan y se propagan sin abrir ni crear nada (FR-023, FR-024), y ningún applet implementa `Emit`. No se persiste ningún dato personal. El registro de eventos **no incluye por omisión contenido que identifique a una persona física** (FR-039): registra applet, verbo, duración y clase de error, nunca `data` ni los argumentos en bruto salvo en `debug` ([D14](./research.md#d14--registro-de-eventos-logslog-a-stderr---verbose-y-kitlegal_log)). La descripción de `--dry-run` sí nombra los argumentos, porque son su objeto, y por eso es un mensaje explícito para la persona y no un registro ([D10](./research.md#d10---dry-run-no-corta-viaja-en-el-contexto-y-la-descripción-va-a-stderr)). | ✅ Cumple |

### Reglas de dependencia (`docs/ROADMAP.md` §2, constitución §IV)

Las cinco pasan en H1 de «sin objeto» a **activas**. FR-052 exige además sustituir la forma acotada por
ruta que H0 dejó como provisional.

| Regla | Situación en H1 | Cómo se hace cumplir | Veredicto |
|---|---|---|---|
| `internal/core/**` no importa `internal/{source,httpx,cache,store,graph,render,cli,app}` | **Activa**: `internal/core/schema` existe y solo importa biblioteca estándar (`crypto/sha256`, `encoding/hex`, `encoding/json`, `time`). El registrador **no** es un campo de `schema.Contexto` (D2), de modo que el dominio tampoco importa `log/slog`. | `depguard`, lista `core`: `files: ["**/internal/core/**"]`, `deny` sobre los ocho paquetes **y sobre el I/O de la biblioteca estándar** (`log`, `log/slog`, `os`, `io`, `net/http`, `database/sql`), para que «dominio puro» sea un control y no una afirmación. **Y** `internal/arch_test.go`, que recorre el grafo real con `go list -deps` y nombra la regla violada. | ✅ Cumple |
| Solo `internal/httpx` importa `net/http` | **Activa y vacía**: `httpx` no existe todavía, así que **ningún** paquete puede importar `net/http`. Eso hace mecánico FR-058. | `depguard`, lista `red`: `files: ["**/internal/**", "**/cmd/**"]` con excepción `!**/internal/httpx/**`, `deny: net/http`. **Y** el test de arquitectura. | ✅ Cumple |
| Solo `internal/{cache,store,graph}` importan SQLite y `database/sql` | **Activa y vacía**: ninguno de los tres existe; ningún paquete importa `database/sql` ni `modernc.org/sqlite`, que ni siquiera está en `go.mod`. | `depguard`, lista `sql`, con las tres rutas exceptuadas. **Y** el test de arquitectura. | ✅ Cumple |
| Solo `internal/cli` y `cmd/` llaman a `os.Exit` | **Activa, y cumplida en forma más estricta que la regla**: `os.Exit` aparece exactamente **dos** veces, una en cada **raíz de composición** —`cmd/kitlegal/main.go` y `internal/app/testdata/kitlegal-e2e/main.go`, el `main` del binario que construye el e2e (FR-009, [D19](./research.md#d19--dónde-viven-los-applets-de-ejemplo-y-cómo-se-compilan-lintan-y-ejecutan))—. No hay una tercera: `internal/cli` **podría** llamarlo según la regla y no lo hace ([D2](./research.md#d2--reparto-entre-internalcli-internalapp-e-internalcoreschema-sin-ciclos)), `app.Main` devuelve `int`, y el `TestMain` del e2e retorna en vez de salir (verificado en `$GOROOT/src/testing/testing.go:379-380`). Kong no debe terminar el proceso por su cuenta; es el supuesto **S3** de [D24](./research.md#d24--supuestos-no-verificados-por-falta-de-red) y es **bloqueante**. | `forbidigo` con `analyze-types: true` y patrón `^os\.Exit$` / `pkg: ^os$`, con `msg` marcado `R4:`, exceptuando **solo** `cmd/**` e `internal/app/testdata/kitlegal-e2e/**` —las dos raíces de composición— y **no** `internal/cli/**`, para que el control vigile la forma estricta que el plan elige (verificado: el campo `pkg` se ignora sin `analyze-types`, `forbidigo/patterns.go:23-26`). | ✅ Cumple |
| Solo `internal/render` escribe en stdout; logs (`log/slog`) a stderr | **Activa en su forma completa**, que es lo que FR-052 pide: desaparece la excepción de H0 para `cmd/`. `internal/render` es el único paquete que escribe, en stdout y en stderr; a Kong se le entregan **sus** escritores, de modo que ni la ayuda se salta la regla ([D11](./research.md#d11--precedencia-entre---describe---help-y-la-ejecución)), e `internal/cli` escribe a través de una interfaz que declara él mismo, sin importar `render` (D2). `os.Stdout` y `os.Stderr` solo se nombran en las **dos raíces de composición**, que los inyectan. `slog` escribe siempre en el escritor de error (FR-036), y lo que debe verse siempre no viaja por `slog` ([D10](./research.md#d10---dry-run-no-corta-viaja-en-el-contexto-y-la-descripción-va-a-stderr)). | `forbidigo` con tres patrones: `^fmt\.Print(\|f\|ln)$` con `msg` marcado `R5:` (**todo el árbol, sin ninguna excepción**, ni siquiera en las raíces de composición), y `^os\.Stdout$` / `^os\.Stderr$` con `msg` marcado `R5-descriptores:`, exceptuando solo esas dos raíces. Las excepciones se escriben como `path` + `text` sobre la marca, así que no pueden desactivar por accidente la prohibición de `fmt.Print*`. Demostrable retirando la excepción y viendo fallar `make lint`. | ✅ Cumple |

### Gates mecánicos (constitución «Gates», capa 1)

Los que este hito **activa por primera vez**: test de arquitectura sobre las reglas de dependencia;
`depguard`; `forbidigo` con `analyze-types` y en su forma completa (sin la acotación de H0); validación de
toda salida de applet contra su descripción formal; **fixtures que fuerzan los códigos de salida 2, 3, 4,
5 y 6** (SC-006, SC-014) más el 0 y el 1 de FR-031; e2e con `testscript` (`make test-e2e` deja de anunciar
un objeto ausente); umbral de cobertura de `internal/cli`; y **`errcheck` sobre las escrituras**, que es lo
que convierte en control la promesa de traducir un fallo de escritura a un código de salida (D27).

Los que **siguen sin objeto**, con su hito asignado: grabación y reproducción de fixtures HTTP (H2/H4);
corrección de citas contra la respuesta grabada del BOE (H4); `schemas/*.json` persistido y su
comprobación de deriva (H4 borrador, H11); frontmatter de `SKILL.md` y diff vacío de `references/` (H5).
Ninguno se simula: `make schema-check` y `make skills-sync` siguen anunciando el objeto ausente y el hito
que lo aporta.

Los de H0 siguen todos activos y **sin ninguna exclusión nueva** (FR-057): formato, lint, `-race`,
`govulncheck`, `gosec`, CodeQL, `gitleaks`, `go mod verify`, `go mod tidy -diff`, pre-commit, cobertura.
Uno de ellos se **endurece**: se retira la exclusión de `errcheck` para `fmt.Fprint*` que H0 justificó
para su punto de entrada, porque en H1 `cmd/` ya no escribe y `internal/render` sí, y el contrato exige
traducir el fallo de escritura a un código de salida
([D27](./research.md#d27--la-exclusión-de-errcheck-para-fmtfprint-se-retira)). Las dos excepciones de
`forbidigo` para las raíces de composición no relajan ninguna regla de H0: sustituyen a la acotación por
ruta que H0 dejó como provisional, y la sustituyen por una más estrecha (FR-052).

**Veredicto del gate: PASA.** No hay ninguna violación sin justificar. Las seis divergencias conscientes
están en *Complexity Tracking*, ninguna afecta a alcance, frontera humana, privacidad, términos de uso ni
a una decisión cerrada.

### Re-evaluación tras la fase 1 (diseño)

El diseño de la fase 1 ([data-model.md](./data-model.md), [contracts/](./contracts/),
[quickstart.md](./quickstart.md)) no introduce ninguna capacidad nueva sobre lo evaluado arriba. Los
elementos que aparecieron al bajar a diseño y que podrían tocar un principio:

- **El valor `1` para el fallo inesperado** (§IV, códigos de salida estables): no contradice ninguno de
  los reservados por `CLAUDE.md`, que fija 0, 2, 3, 4, 5 y 6 y deja 1 libre. FR-031 solo exige «ni 0 ni
  reservado». Queda documentado en [`contracts/banderas-y-exit-codes.md`](./contracts/banderas-y-exit-codes.md)
  para que ningún hito posterior lo reutilice para otra cosa.
- **El pánico de arranque ante un registro mal construido** (§IV, «ningún `panic` en rutas de usuario»):
  no es una ruta de usuario sino un defecto de compilación, y FR-008 exige precisamente que no se
  manifieste como error de usuario. Documentado en [D17](./research.md#d17--precedencia-del-despacho-nombres-no-registrados-y-registro-que-se-rechaza)
  y en el contrato del registro.
- **Los applets de ejemplo pasan el lint completo** (§I criterio de decisión autónoma, «ni atajos ni
  ñapas»): como los comodines de Go no descienden a `testdata`, hay que enumerarlos explícitamente en el
  `Makefile` ([D19](./research.md#d19--dónde-viven-los-applets-de-ejemplo-y-cómo-se-compilan-lintan-y-ejecutan)).
  Es una línea de configuración, no una excepción: los dos applets de ejemplo son la implementación de
  referencia que copiará cada applet posterior y no pueden estar por debajo del listón.
- **Someterlos al lint obliga a declarar qué es cada uno** (§IV, R4 y R5): `internal/app/testdata/ejemplo`
  es código de applet y no puede nada; `internal/app/testdata/kitlegal-e2e` es un `package main` y es la
  **raíz de composición del binario de e2e**, así que necesita `os.Exit` y los dos descriptores, igual que
  `cmd/`. La excepción se acota a esa ruta, lleva la marca de la regla y no alcanza al paquete hermano
  (D18, D19). Lo contrario —dejar el lint tal cual— haría fallar `make lint` sobre código legítimo del
  propio hito, y las dos salidas de ahí serían atajos: renunciar al control o dejar el código fuera.
- **El verbo por omisión** (§V, YAGNI): la entrega literal del hito (`kitlegal echo hola`) no nombra
  verbo, y el contrato del applet exige al menos uno. Se resuelve con un booleano en un `struct` que el
  applet ya declara y una normalización de diez líneas en `internal/app`, sin apoyarse en Kong ni añadir
  supuestos ([D26](./research.md#d26--verbo-por-omisión-cómo-funciona-kitlegal-echo-hola-sin-nombrar-verbo)).
  No se anticipa nada: es lo mínimo para que la entrega del hito exista.
- **El pre-escaneo acotado de `argv`** (§V y §IV): SC-014 exige sobre de fallo en dos casos anteriores al
  análisis de la gramática, así que hay que saber si se pidió `--json` antes de poder analizarlo. Se
  declara como mecanismo único, estrecho y probado
  ([D25](./research.md#d25--pre-escaneo-acotado-de-argv-antes-de-la-gramática)),
  y se comprueba en test que coincide con el análisis de Kong. No es una segunda gramática ni una segunda
  fuente de verdad: es un valor provisional que se descarta en cuanto existe el definitivo.
- **Retirar una exclusión de `errcheck`** (§I, «ni atajos ni ñapas»): FR-057 exige que no haya exclusiones
  nuevas sin justificar, no que las viejas sean intocables. La de `fmt.Fprint*` protegía a un punto de
  entrada que a partir de H1 ya no escribe, y taparía el único camino que el contrato manda vigilar en
  `internal/render` (D27). Se retira entera.

**Veredicto tras el diseño: PASA**, con dos entradas nuevas en *Complexity Tracking* —la enumeración
explícita de los paquetes bajo `testdata` con su excepción de `forbidigo`, y el verbo por omisión— y sin
ninguna violación.

## Project Structure

### Documentation (this feature)

```text
specs/002-h1-kernel-cli-multicall/
├── plan.md              # Este fichero
├── research.md          # Fase 0: 27 decisiones con alternativas y motivo, y los supuestos no verificados
├── data-model.md        # Fase 1: entidades del kernel y sus invariantes
├── quickstart.md        # Fase 1: guía de validación ejecutable
├── contracts/
│   ├── sobre-de-salida.md          # El sobre, la huella, la procedencia y el sobre de fallo
│   ├── banderas-y-exit-codes.md    # Las ocho banderas globales, la ayuda y la tabla de códigos
│   ├── registro-y-describe.md      # Applet, registro, despacho multicall y --describe
│   └── reglas-de-arquitectura.md   # Las cinco reglas y cómo se comprueba cada una
├── spec.md
├── checklists/
├── gates/
└── tasks.md             # Fase 2 (`/speckit-tasks`, no lo crea este comando)
```

### Source Code (repository root)

Estructura de `docs/ROADMAP.md` §2 y `CLAUDE.md`. En negrita lo que H1 crea; el resto ya existe de H0.

```text
cmd/kitlegal/
├── main.go                    ← se reduce: main() → os.Exit(app.Main(os.Args, app.RegistroDeProduccion(),
│                                os.Stdout, os.Stderr, version, commit, fecha)). Deja de escribir en stdout
└── main_test.go               ← se adapta: el contrato de `version` no cambia (D16)

internal/
├── arch_test.go               **NUEVO** paquete de solo test que recorre `go list -deps` y comprueba las
│                              tres reglas de importación —incluida la mitad de R1 que prohíbe el I/O de
│                              la biblioteca estándar en el dominio—, nombrando la regla violada
│                              (FR-051, FR-053)
├── core/
│   └── schema/                **NUEVO** dominio puro, sin I/O. Solo biblioteca estándar
│       ├── sobre.go             Sobre, Procedencia, Resultado, constructor y validación
│       ├── huella.go            forma canónica de `data` + hash sha256 con prefijo (D4)
│       ├── contexto.go          Contexto de ejecución: las SEIS decisiones globales que viajan al applet.
│       │                        El *slog.Logger NO está aquí: viaja como parámetro de Ejecutar (D2)
│       ├── error.go             Clase (constantes) y DatosError {clase, mensaje}
│       └── *_test.go
├── cli/                       **NUEVO** el kernel
│   ├── globales.go              Globales: las ocho banderas, con etiquetas Kong
│   ├── preescaneo.go          **NUEVO** pre-escaneo acotado de argv: --json, --verbose, --help (D25)
│   ├── parse.go                 construcción y análisis de la gramática de una invocación (D1)
│   ├── errors.go                sentinelas, clasificación con errors.Is y la ÚNICA tabla clase→código
│   ├── sobre.go                 montaje del sobre de éxito y del sobre de fallo (FR-045)
│   ├── presentador.go           la interfaz que cli consume; la implementa render y la inyecta app (D2)
│   ├── describe.go              esquema de entrada y salida con invopop/jsonschema, `data` condicionado
│   ├── log.go                   slog a stderr; nivel por --verbose y KITLEGAL_LOG
│   └── *_test.go
├── app/                       **NUEVO** raíz de composición y registro
│   ├── applet.go                Applet, Verbo (con PorOmision), Argumentos
│   ├── registro.go              Registro: Registrar (con validación FR-008), Buscar, Nombres
│   ├── despacho.go              os.Args[0] → primer argumento → verbos reservados (D17) y
│   │                            normalización del verbo por omisión antes de la gramática (D26)
│   ├── main.go                  Main(argv, registro, stdout, stderr, datos de versión) int
│   ├── ayuda.go                 ayuda del binario derivada del registro (FR-026, sin lista paralela)
│   ├── e2e_test.go              construye el binario y ejecuta los guiones con testscript (D20);
│   │                            TestMain retorna, no llama a os.Exit (testing.go:379-380)
│   ├── *_test.go
│   └── testdata/              **NUEVO** material de test; los comodines de Go NO descienden aquí (D19)
│       ├── ejemplo/             los DOS applets de ejemplo: `echo` (verbo `repetir` por omisión) y el
│       │                        segundo de SC-010 (dos verbos, ninguno por omisión). Código de applet:
│       │                        sin excepción de lint de ninguna clase
│       ├── kitlegal-e2e/        package main: kernel real + registro de ejemplo (FR-009). RAÍZ DE
│       │                        COMPOSICIÓN del binario de e2e: os.Exit y os.Stdout/os.Stderr, igual
│       │                        que cmd/ y exceptuado por ruta exacta en el lint (D18)
│       └── script/*.txtar       guiones de testscript: ayuda, enlace simbólico, exit 2, JSON analizable
└── render/                    **NUEVO** único escritor, de stdout y de los mensajes para la persona
    ├── render.go                Presentador sobre (stdout, stderr); elige forma según --json; toda
    │                            escritura comprueba y propaga su error (D15, D27)
    ├── json.go                  el sobre serializado, y nada más (FR-042)
    ├── tabla.go                 tabla mínima: 4 líneas de procedencia + `data` aplanado por rutas (D15)
    ├── texto.go                 texto para personas: `version`, ayuda derivada del registro (stdout) y
    │                            avisos —fallo, lista de applets, descripción de --dry-run— (stderr)
    └── *_test.go

.golangci.yml                  ← se amplía: depguard (5 listas; la de `core` deniega también el I/O de la
                                 biblioteca estándar), forbidigo con analyze-types, marcas de regla en
                                 `msg` y excepciones solo para las dos raíces de composición (desaparece
                                 la acotación a internal/** de H0, FR-052); y se RETIRA la exclusión de
                                 errcheck para fmt.Fprint* (D27)
codecov.yml                    ← se amplía: componente internal_cli con objetivo 90 %, bloqueante (D21)
Makefile                       ← test-e2e pasa a ejecutar el paquete de e2e; lint/fmt/fmt-check enumeran
                                 además los dos paquetes bajo internal/app/testdata (D19)
go.mod, go.sum                 ← 5 dependencias nuevas; aparece go.sum en la raíz por primera vez
CHANGELOG.md                   ← sección Unreleased
docs/ADR/
├── 0005-contrato-de-applet.md          **NUEVO** (D23)
└── 0006-sobre-de-salida-y-huella.md    **NUEVO** (D23)
```

**Structure Decision**: la de `docs/ROADMAP.md` §2 sin desviaciones. La única elección no evidente es en
qué paquete vive cada tipo, y está resuelta por la dirección de dependencia que exige el roadmap:
`cmd` → `app` → `cli` → `core/schema`, con `render` → `core/schema`. Como el roadmap sitúa el registro de
applets en `internal/app` y las banderas y el mapeo de errores en `internal/cli`, la función de `cli` que
analiza la gramática recibe el `struct` como `any` y así **no hay ciclo**
([D2](./research.md#d2--reparto-entre-internalcli-internalapp-e-internalcoreschema-sin-ciclos)). Por la
misma razón `cli` **no importa** `render`: declara la interfaz `Presentador` que consume y `app` le
inyecta la implementación, que es lo que permite que el kernel escriba sin invertir la dirección de
dependencia del roadmap.
`internal/source/`, `internal/httpx`, `internal/cache`, `internal/store`, `internal/graph` y
`pkg/legalkit` **no se crean**: están fuera de alcance y crearlos vacíos sería un marcador de posición,
que SC-012 prohíbe.

## Controles mecánicos que este hito añade o toca

Según la sección «Gates» de la constitución. Cada fila indica la orden que lo ejecuta, si forma parte del
veredicto agregado y **con qué comprobación concreta se demuestra que el control está vivo** —es decir,
qué falla si el control se retira.

| # | Control | Herramienta / forma | Orden | ¿En `make ci`? | Demostración de que está activo |
|---|---|---|---|---|---|
| 1 | **Regla de arquitectura: importaciones** (3 de las 5) | `depguard` v2.2.1, cinco listas con `files` sobre ruta absoluta (`**/internal/core/**`, …); la lista `core` deniega además el I/O de la biblioteca estándar | `lint` | Sí | Un `import "net/http"` en `internal/cli` hace fallar `make lint` nombrando la lista `red`; un `import "log/slog"` en `internal/core/schema` falla nombrando la lista `core` (quickstart esc. 8) |
| 2 | **Regla de arquitectura: símbolos** (2 de las 5) | `forbidigo` v2.3.1 con `analyze-types: true`: `^os\.Exit$` (marca `R4:`), `^fmt\.Print(\|f\|ln)$` (marca `R5:`), `^os\.Stdout$` / `^os\.Stderr$` (marca `R5-descriptores:`); excepciones por `path`+`text` solo en las dos raíces de composición | `lint` | Sí | Un `os.Exit` en `internal/app` falla; el mismo `os.Exit` en `cmd/` o en `internal/app/testdata/kitlegal-e2e` pasa; el mismo `os.Exit` en `internal/app/testdata/ejemplo` **falla** (quickstart esc. 8 y 9) |
| 3 | **Test de arquitectura** | `internal/arch_test.go` sobre `go list -deps`, segunda capa independiente del lint (FR-051) | `test` | Sí | Borrar las cinco listas de `.golangci.yml` **no** basta: el test sigue fallando ante la misma violación (quickstart esc. 8) |
| 4 | **Tests de contrato del sobre** | `santhosh-tekuri/jsonschema/v6` con `AssertFormat()`, sobre la salida real de los applets de ejemplo | `test` | Sí | Emitir un sobre con `url: ""` hace fallar el test; sin `AssertFormat()` pasaría (D13) |
| 5 | **Tests de contrato de `--describe`** | el sobre de fallo emitido se valida contra el esquema que el propio applet emite (SC-015, FR-047) | `test` | Sí | Quitar el `else` del condicional hace fallar el test del sobre de fallo |
| 6 | **Fixtures de códigos de salida** | tabla de casos que fuerza 0, 1, 2, 3, 4, 5 y 6 desde un applet de prueba, y comprueba código, descriptor del mensaje y sobre de fallo (SC-006, SC-014) | `test` | Sí | Añadir una `Clase` nueva sin rama en la traducción hace fallar el linter `exhaustive` **antes** que el test (D8) |
| 7 | **Reproducibilidad de la huella** | mismo `data` → misma huella; un byte de diferencia → huella distinta (SC-005) | `test` | Sí | Serializar sin el viaje de ida y vuelta hace fallar el caso de orden de claves (D4) |
| 8 | **E2E con `testscript`** | `internal/app/e2e_test.go` + `internal/app/testdata/script/*.txtar`; construye el binario en un temporal y lo pone en `PATH` | **`test-e2e`** (deja de ser un anuncio) y `test` | Sí (vía `test`) | Los cuatro casos del hito: ayuda, enlace simbólico, exit 2 con argumentos malos y JSON analizable (SC-009) |
| 9 | **Cobertura de `internal/cli` ≥ 90 %** | componente `internal_cli` en `codecov.yml`, bloqueante, sin exclusiones (D21) | `test` (el perfil) + publicación en `ci.yml` | Sí (el perfil); el estado lo emite Codecov | Borrar la mitad de los tests de `internal/cli` baja del umbral y bloquea (SC-004) |
| 10 | **Lint de los applets de ejemplo** | `TESTDATA_PKGS` enumerado explícitamente en `lint`, `fmt` y `fmt-check` (D19) | `lint`, `fmt-check` | Sí | Un defecto deliberado en `internal/app/testdata/ejemplo` hace fallar `make lint`, **incluidos** `os.Exit` y `fmt.Println`: la excepción de la raíz de composición no alcanza al paquete hermano (quickstart esc. 9) |
| 11 | **Error de escritura propagado** | `errcheck` sin la exclusión de `fmt.Fprint*` (D27) más `TestEscrituraFallida`, con un `io.Writer` que siempre falla | `lint` + `test` | Sí | Descartar el error de una escritura en `internal/render` hace fallar `make lint`; el test comprueba que el fallo sale con código 1 y que **no** se emite un segundo sobre |
| 12 | **El pre-escaneo coincide con el análisis** | `TestPreescaneo`: tabla de invocaciones bien formadas donde `PreEscanear` y el resultado de Kong deben dar el mismo `--json`, `--verbose` y `--help` (D25) | `test` | Sí | Añadir una forma sintáctica que Kong acepte y el pre-escaneo no reconozca hace fallar el test antes de que el sobre de fallo salga con la forma equivocada |
| 13 | **Verbo por omisión** | `TestVerboPorOmision` sobre el despacho: con verbo por omisión, sin él, y con `--help` (D26); y la validación del registro, que rechaza dos verbos marcados | `test` | Sí | Quitar la normalización hace fallar `kitlegal echo hola` con código 2 —la entrega literal del hito— y lo ve el e2e; marcar dos verbos por omisión hace fallar la construcción del registro |
| 14 | Formato, lint general, `-race`, `govulncheck`, `gosec`, CodeQL, `gitleaks`, `go mod verify`, `go mod tidy -diff`, pre-commit | los de H0, **sin ninguna exclusión nueva**; uno se endurece, al retirarse la exclusión de `errcheck` (FR-057, D27) | varias | Sí (salvo CodeQL) | Los mismos de [H0](../001-h0-esqueleto-del-repo/plan.md); `go mod verify` y `go mod tidy -diff` cubren ahora también el `go.sum` de la raíz, que aparece por primera vez |

**Fixtures.** H1 **no crea `testdata/<fuente>/`** ni graba nada: no hay fuentes externas, así que no hay
respuesta que grabar y `KITLEGAL_RECORD=1` no tiene objeto. El único `testdata/` que nace es
`internal/app/testdata/`, que es **material de test, no fixtures de fuente**: contiene código Go (los dos
applets de ejemplo y el `main` del e2e) y los guiones `.txtar`. La regla de capa 3 de la constitución
—«toda modificación de ficheros existentes en `testdata/` … solo en tareas etiquetadas `[datos]`, con
pausa obligatoria»— está pensada para fixtures grabados de fuentes; se anota aquí explícitamente que
`internal/app/testdata/` no lo es, para que el guardián de diff del workflow no lo trate como tal ni,
al revés, lo deje sin vigilar. Las tareas que lo toquen deben declarar esas rutas como cualquier otra.

**Tests de contrato.** Los cuatro contratos de H1 están en [`contracts/`](./contracts/) y cada uno tiene su
comprobación mecánica: el sobre (filas 4, 7 y 11), las banderas y los códigos de salida (filas 6, 11 y
12), el registro y `--describe` (filas 5, 8 y 13) y las reglas de arquitectura (filas 1, 2 y 3). Ninguno
se comprueba «leyendo el código»: el criterio de la constitución —«lo que se puede comprobar con un script
nunca se delega a un modelo»— se aplica a los cuatro, y por eso las tres garantías que el diseño añadió al
bajar a contrato —error de escritura, pre-escaneo y verbo por omisión— entran en la tabla con control
propio en lugar de quedarse en una frase.

**Objetivos del `Makefile` que H1 toca**, y ninguno más: `test-e2e` (pasa a ejecutar el paquete de e2e),
`lint`, `fmt` y `fmt-check` (añaden `TESTDATA_PKGS`). `schema-check`, `skills-sync` y `release` siguen
anunciando el hito que las aporta; `build`, `install`, `test`, `test-integration`, `vuln`, `secrets`,
`mod-verify`, `mod-tidy-check`, `check-tools`, `hooks`, `ci` y `help` **no cambian**.

## Orden de implementación

La constitución fija `core → adaptador → applet → skill`. En H1 no hay adaptador de fuente ni skill, así
que el orden es `core → kernel → presentación → composición → ejemplos`, con el test e2e escrito **antes**
que todo, como manda el ritual del roadmap §6.

1. **El guion de e2e que describe la entrega**, en `internal/app/testdata/script/`: `kitlegal echo hola
   --json`, el enlace simbólico, `--help`, exit 2 y `--describe`. Falla porque no hay nada. Es la
   definición ejecutable del hito.
2. **Comprobar los supuestos sobre Kong** (S1-S5 de [D24](./research.md#d24--supuestos-no-verificados-por-falta-de-red)):
   `go get github.com/alecthomas/kong`, `go doc`, y un test mínimo que fuerza `--help` y una bandera
   desconocida y comprueba que **el proceso no termina**. S3 es bloqueante: si Kong no permite interceptar
   la salida del proceso, se aplica la contingencia antes de escribir nada más.
3. **`internal/core/schema`** con sus tests: sobre, procedencia, contexto, clases, forma canónica y huella.
   Es lo único de lo que todo lo demás depende, y no depende de nada.
4. **`internal/cli`**: `errors.go` (sentinelas + clasificación + tabla exhaustiva) primero, porque es lo
   que el resto del kernel usa para fallar; luego `preescaneo.go` y `log.go`, porque el nivel del registro
   y la forma de un fallo temprano se deciden antes que nada ([D25](./research.md#d25--pre-escaneo-acotado-de-argv-antes-de-la-gramática));
   luego `presentador.go` (la interfaz que `cli` consume), `globales.go` y `parse.go`; luego `sobre.go`
   (éxito y fallo); `describe.go` al final, porque necesita el sobre y la gramática ya fijados.
5. **`internal/render`**: la forma JSON antes que la tabla; la tabla necesita el sobre ya estable; y
   `texto.go` con los mensajes para la persona, que es lo que hace visible la descripción de `--dry-run`
   sin depender del nivel de registro (D10). Toda escritura comprueba y propaga su error, con la exclusión
   de `errcheck` ya retirada (D27), para que el compilador y el lint acompañen desde la primera línea.
6. **`internal/app`**: `applet.go` y `registro.go` (con la validación de FR-008, incluido el verbo por
   omisión único), `despacho.go` —precedencia del multicall y normalización del verbo
   ([D26](./research.md#d26--verbo-por-omisión-cómo-funciona-kitlegal-echo-hola-sin-nombrar-verbo))—,
   `ayuda.go`, y `main.go` como raíz de composición, que es quien inyecta el `render.Presentador`.
7. **Los dos applets de ejemplo** en `internal/app/testdata/ejemplo/` —`echo`, con verbo por omisión, y el
   segundo de SC-010, con dos verbos y ninguno por omisión— y el `main` del e2e en
   `internal/app/testdata/kitlegal-e2e/`. Aquí el guion del paso 1 empieza a pasar.
8. **`cmd/kitlegal/main.go`** se reduce a la inyección, y `main_test.go` se adapta sin cambiar el contrato
   de `version` ([D16](./research.md#d16--kitlegal-version-de-h0-sigue-existiendo-y-sale-por-render)).
9. **Las reglas mecánicas**: `depguard` y `forbidigo` en `.golangci.yml` (retirando la acotación de H0,
   FR-052, y dejando como únicas excepciones las dos raíces de composición) e `internal/arch_test.go`. Van
   al final porque necesitan que los paquetes a los que se refieren existan; y se validan **rompiendo cada
   regla a propósito** y viendo fallar las dos capas, incluida la comprobación de que la excepción de la
   raíz de composición **no** alcanza a `internal/app/testdata/ejemplo`.
10. **Cobertura y documentación**: componente `internal_cli` en `codecov.yml`, `CHANGELOG.md` y los dos
    ADR. `make ci` en verde de principio a fin.

## Complexity Tracking

> Divergencias conscientes respecto de la lectura literal del spec o del roadmap, y estructura añadida que
> no era la opción por defecto. Ninguna afecta a alcance, frontera humana, privacidad, términos de uso ni
> a una decisión cerrada, por lo que ninguna dispara el criterio §4 (escalar). **No hay ninguna
> dependencia fuera de la lista de la constitución §V**, de modo que esta tabla no contiene ninguna.

| Violación / divergencia | Por qué es necesaria | Alternativa más simple, y por qué se rechaza |
|---|---|---|
| **Las reglas de dependencia se comprueban dos veces** (lint y test de arquitectura), con la duplicación que eso supone | Lo exige FR-051 palabra por palabra: «de modo que la garantía no dependa solo de la configuración del lint». Una configuración de lint se desactiva borrando tres líneas de YAML o con un `//nolint`; un test que falla en `make ci` no se desactiva sin que se vea en el diff. Las dos capas cubren además cosas distintas: `depguard` ve importaciones declaradas, el test ve el grafo **transitivo** real de `go list -deps`. | Solo `depguard`: más simple y más rápido, pero deja la garantía a merced de una línea de configuración, que es exactamente el riesgo que el hito existe para cerrar. Solo el test: perdería el mensaje en el momento de editar y la integración con el pre-commit. |
| **`os.Exit` vive solo en las raíces de composición —`cmd/kitlegal` y el `main` del binario de e2e—, y no en `internal/cli`**, aunque la regla autoriza `internal/cli` y FR-035 los nombra juntos | Un único punto de salida **por binario** es lo que permite que `app.Main` devuelva `int` y que todo el kernel sea testable sin lanzar un subproceso —que es lo que sostiene la cobertura ≥ 90 % de SC-004—. Es además el patrón que H0 ya dejó establecido con `run(args, stdout, stderr) int`. Cumplir un subconjunto estricto de la regla no la viola, y la excepción del lint se acota a ese subconjunto: `internal/cli/**` **no** está exceptuado. | Permitir `os.Exit` también en `internal/cli`: obligaría a probar los caminos de salida con subprocesos, que no cuentan para la cobertura, y multiplicaría los puntos por los que el proceso puede terminar sin pasar por la traducción única de FR-030. |
| **Los dos paquetes bajo `internal/app/testdata/` se enumeran explícitamente** en `lint`, `fmt`, `fmt-check` y el test de arquitectura, y **uno de los dos —y solo uno— queda exceptuado de `^os\.Exit$` y `^os\.Stdout$`/`^os\.Stderr$`** | Verificado empíricamente que los comodines de Go **no** descienden a `testdata` en ningún caso —ni `./...`, ni `./internal/app/testdata/...`, ni el patrón con ruta de módulo—, mientras que los paquetes enumerados explícitamente sí se resuelven y compilan ([D19](./research.md#d19--dónde-viven-los-applets-de-ejemplo-y-cómo-se-compilan-lintan-y-ejecutan)). Los applets de ejemplo son la implementación de referencia que copiará cada applet posterior: dejarlos sin lintar sería el atajo que el criterio §1 prohíbe. Pero someterlos al lint obliga a reconocer que `kitlegal-e2e` es un `package main`: **ningún `main` puede propagar un código de salida sin `os.Exit` ni inyectar descriptores sin nombrarlos**, así que es una raíz de composición y se le trata como a `cmd/`. La excepción va por ruta exacta y con la marca de la regla, de modo que no alcanza a `ejemplo/` ni desactiva `fmt.Print*`. | Dejarlos fuera de los controles: es la opción por omisión y no cuesta nada configurar, pero produce la peor consecuencia posible —que el código que todo el mundo va a copiar sea el único que no pasa el listón—. Moverlos fuera de `testdata/`: contradice FR-009, que fija su ubicación. Dejar el lint sin la excepción: `make lint` fallaría sobre código legítimo del propio hito, y de ahí solo se sale con un atajo. |
| **`Verbo` gana un campo `PorOmision`** y el kernel normaliza la lista de argumentos antes de construir la gramática | La entrega literal del hito es `kitlegal echo hola --json`, que no nombra verbo, y el contrato del applet exige `Verbos()` no vacío con nombres únicos: sin verbo por omisión declarado, esa invocación terminaría en código 2 y el hito no entregaría lo que dice. Es un booleano en un `struct` que el applet ya declara, con su invariante («como máximo uno») comprobada al construir el registro ([D26](./research.md#d26--verbo-por-omisión-cómo-funciona-kitlegal-echo-hola-sin-nombrar-verbo)). | Asumir implícitamente el verbo cuando el applet tiene uno solo: la invocación cambiaría de comportamiento sola el día que el applet añada el segundo verbo, rompiendo llamadas que antes funcionaban —exactamente lo que `boe` no puede permitirse—. Un método `VerboPorOmision() string` en `Applet`: más superficie en la interfaz que SC-010 quiere mínima, y permite nombrar un verbo inexistente. Apoyarse en el comando predeterminado de Kong: añadiría un supuesto más a [D24](./research.md#d24--supuestos-no-verificados-por-falta-de-red) para ahorrar diez líneas deterministas. |
| **Se retira una exclusión de `errcheck` heredada de H0** (`fmt.Fprint*`), en un hito que promete no tocar los controles de H0 | La justificación de H0 estaba acotada a su punto de entrada, que a partir de H1 **ya no escribe**; y el contrato de H1 exige traducir el fallo de escritura en stdout a un código de salida. Mientras la exclusión siguiera en pie cubriría a `internal/render` —el único escritor del binario— y el requisito quedaría sin control mecánico ([D27](./research.md#d27--la-exclusión-de-errcheck-para-fmtfprint-se-retira)). Retirar una exclusión endurece; FR-057 prohíbe exclusiones nuevas, no prohíbe quitar las viejas. | Acotarla a `cmd/**`: quedaría una exclusión muerta —no hay escrituras en `cmd/`— que alguien reactivaría por costumbre. Dejarla y confiar en un test: el test cubre lo que se recuerde probar; `errcheck` cubre todas las escrituras, también las de H4 en adelante. |
| **El registro rechaza un applet inválido con un pánico de arranque**, y no devolviendo un error al usuario | FR-008 exige que el fallo «NO DEBE manifestarse como un error de usuario en tiempo de invocación». Un registro duplicado solo puede existir si alguien compiló mal el binario: es un defecto de programación, detectable en el arranque, y convertirlo en un error de usuario lo escondería detrás de un código de salida que significa otra cosa. FR-033 prohíbe el pánico «en rutas de usuario», y esta no lo es. | Devolver error y salir con un código: el usuario recibiría un fallo incomprensible que no puede corregir, y el defecto llegaría a producción en lugar de reventar en el primer test. Validar solo en un test: el binario podría publicarse roto si alguien registra un applet sin añadir el caso. |

## Obligaciones que este plan traslada a `tasks.md`

1. **Comprobar los cinco supuestos sobre Kong (S1-S5)** antes de escribir el kernel, y dejar el resultado
   por escrito. **S3 es bloqueante**: si Kong no permite impedir que termine el proceso por su cuenta, se
   aplica la contingencia de [D24](./research.md#d24--supuestos-no-verificados-por-falta-de-red) —generar
   la ayuda desde el registro y desactivar la de Kong— antes de continuar, y se anota en la PR.
2. **Comprobar que `golangci-lint` v2 no excluye `testdata/` por omisión**, introduciendo un defecto
   deliberado en `internal/app/testdata/ejemplo` y viendo fallar `make lint`. Si resultara que sí lo
   excluye, declararlo en `.golangci.yml` en lugar de renunciar al control.
3. **Romper cada una de las cinco reglas de dependencia a propósito**, una por una, y comprobar que
   **las dos capas** fallan y nombran la regla (SC-008). Es el único modo de demostrar que el control está
   vivo y no solo escrito. En una copia desechable o revirtiendo el cambio; `git status` limpio al final.
4. **Medir el flujo `ci` con el e2e dentro** y compararlo con el número que H0 registró. El e2e añade una
   compilación por ejecución del paquete. Si el flujo superara los 3 minutos en caliente, la contingencia
   es reutilizar el binario entre guiones o mover el e2e a un job propio del flujo —se corrige el diseño
   del flujo, nunca el criterio.
5. **Comprobar en la PR que el estado de cobertura del componente `internal_cli`** aparece y bloquea, igual
   que se hizo con `internal/core` en H0. Si no apareciera, la causa es la configuración de Codecov, no el
   umbral: nunca se rebaja el objetivo para que el estado pase.
6. **Anotar la prohibición pendiente para H4**: ningún adaptador de `internal/source/<fuente>` puede usar
   el prefijo `kitlegal.` ni el esquema `kitlegal:` (FR-016). En H1 no hay nada que comprobar porque no hay
   adaptadores; la comprobación mecánica nace con el primero, en H4, y debe quedar escrita en la PR de H1
   para que no se pierda.
7. **Declarar las rutas de cada tarea que toque `internal/app/testdata/`**, dejando claro en la propia
   tarea que ese directorio es material de test y **no** fixtures grabados de fuente, de modo que el
   guardián de diff del workflow no lo confunda con los `testdata/<fuente>/` que la constitución protege
   en la capa 3.
8. **Demostrar que las dos excepciones de `forbidigo` son estrechas**: la misma violación que pasa en
   `internal/app/testdata/kitlegal-e2e` debe **fallar** en `internal/app/testdata/ejemplo`, y `fmt.Println`
   debe fallar en las dos. Si alguna pasara donde no debe, se corrige la excepción; nunca se amplía.
9. **Cerrar el pre-escaneo contra la realidad de Kong** ([D25](./research.md#d25--pre-escaneo-acotado-de-argv-antes-de-la-gramática)):
   una vez descargado Kong, enumerar las formas sintácticas que acepta para una bandera booleana larga y
   comprobar en `TestPreescaneo` que el pre-escaneo coincide con el análisis en todas ellas. Lo que no se
   soporte se declara por escrito en el contrato; no se deja a la suerte.
