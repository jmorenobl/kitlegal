# Research: H0 · Esqueleto del repo y gates de CI

**Fecha**: 2026-09-10 · **Rama**: `h0-esqueleto-del-repo` · **Spec**: [spec.md](./spec.md)

Modo desatendido. Cada decisión se toma con la sección «Criterio de decisión autónoma» de
`.specify/memory/constitution.md` y se registra aquí con alternativas y motivo, conforme al criterio §3
(«elegir con criterio y dejar rastro»). Ninguna de las decisiones de este documento afecta a alcance,
frontera humana, privacidad, TOS ni a una decisión cerrada de `CLAUDE.md`; por tanto ninguna dispara el
criterio §4 (escalar).

**Las decisiones marcadas «verificado» se comprobaron ejecutando el comando en esta máquina**
(Go 1.26.6 darwin/arm64, 2026-09-10) en módulos de prueba desechables fuera del repositorio. Las marcadas
«pendiente de medir» generan una obligación explícita para `tasks.md`. Los experimentos de D1 se hicieron
con los valores de aquel momento (`toolchain go1.26.6`); la corrección del 2026-09-11 (ver D1, «Por qué
go1.27.1») subió el parche fijado y volvió a ejecutar `make ci` con él, y las conclusiones sobre el
mecanismo —qué es suelo y qué es pin— no dependen del número concreto.

Cuando lo verificado es el **comportamiento documentado de una herramienta** y no su ejecución, se cita la
fuente: la documentación oficial (go.dev/doc/toolchain) o el propio código del go command instalado
(`$GOROOT/src/cmd/go/internal/toolchain/select.go`). Ninguna afirmación de este documento se apoya en
memoria sobre el funcionamiento de una herramienta.

---

## D1 · Versión de Go: `go 1.27.0` + `toolchain go1.27.1`, y el `Makefile` exporta `GOTOOLCHAIN`

- **Decisión**: dos piezas, no una.
  1. `go.mod` declara **dos** directivas: `go 1.27.0`, que es el suelo del lenguaje, y
     `toolchain go1.27.1`, que nombra el parche con el que se compila y se analiza. La directiva es la
     **única fuente de verdad** de la versión de Go del proyecto.
  2. El `Makefile` **deriva de esa directiva** el valor de `GOTOOLCHAIN` y lo **exporta** a todas sus
     recetas:

     ```make
     GO_TOOLCHAIN := $(shell awk '$$1 == "toolchain" { print $$2; exit }' go.mod)

     ifeq ($(GO_TOOLCHAIN),)
     $(error go.mod no declara una directiva `toolchain`; ver research.md D1)
     endif

     export GOTOOLCHAIN := $(GO_TOOLCHAIN)
     ```

  La CI sigue instalando Go con `actions/setup-go` y `go-version-file: go.mod`, **sin** `go-version` ni
  `check-latest`. Subir de parche es un commit explícito de una línea sobre la directiva `toolchain`; el
  `Makefile` no se toca.
- **Por qué el parche tiene que ser exacto, y no «al menos»**: `govulncheck` analiza también la biblioteca
  estándar del toolchain en uso, así que el veredicto de `make vuln` depende del parche de Go. Si local y
  CI ejecutan parches distintos, la misma revisión puede pasar `make vuln` en una máquina y fallarlo en la
  propuesta de cambio (o al revés): eso rompe SC-004 en su letra —«un árbol que pasa `ci` en local pasa los
  controles de la PR»— y vacía FR-042 (a), «versión fijada y explícita, idéntica en la máquina de quien
  contribuye y en la integración continua». El mismo argumento, en menor grado, vale para los analizadores
  que `golangci-lint` compila contra el toolchain instalado.
- **Por qué la directiva `toolchain` sola NO basta** (corrección de la ronda anterior de este documento,
  que afirmaba lo contrario): con `GOTOOLCHAIN=auto` —el valor por defecto— la directiva es un **suelo**,
  no un pin exacto. La documentación oficial es literal: «*In the standard configuration, the `go` command
  uses its own bundled toolchain when that toolchain is at least as new as the `go` or `toolchain` lines in
  the main module or workspace*» (go.dev/doc/toolchain). El código lo confirma: en
  `$GOROOT/src/cmd/go/internal/toolchain/select.go`, el toolchain por defecto es el local
  (`minToolchain := gover.LocalToolchain()`) y la directiva solo lo sustituye si es **estrictamente mayor**
  (`if gover.Compare(toolVers, minVers) > 0`).
  **Verificado en esta máquina** (go1.26.6 darwin/arm64, módulo desechable con `toolchain go1.26.1` y
  `GOTOOLCHAIN=auto`): `go version` responde `go version go1.26.6 darwin/arm64` — ejecuta el local, **no**
  el de la directiva. Es decir: cualquier parche posterior al fijado que alguien instale en su máquina se
  ejecutaría allí mientras la CI ejecuta el fijado, que es exactamente la divergencia que esta decisión
  existe para cerrar. La premisa anterior —«quien contribuye ejecuta el parche de la directiva aunque tenga
  instalado otro, sin hacer nada»— era falsa en esa dirección, y con ella caían el diseño de `check-tools`
  ([D18](#d18--comprobación-de-prerrequisitos-fr-010)) y el escenario 7 de `quickstart.md`.
- **Cómo se cumple ahora, con los datos del mecanismo** (no con supuestos):
  - **`GOTOOLCHAIN=<nombre>` sí es un pin exacto, en las dos direcciones.** Documentación oficial:
    «*When `GOTOOLCHAIN` is set to `<name>` (for example, `GOTOOLCHAIN=go1.21.0`), the `go` command always
    runs that specific Go toolchain. If a binary with that name is found in the system PATH, the `go`
    command uses it. Otherwise the `go` command uses a Go toolchain it downloads and verifies*». En
    `select.go` se ve por qué: con un nombre sin sufijo `+auto`/`+path`, `mode` queda vacío y **el bloque
    que lee `go.mod` para decidir el toolchain ni siquiera se ejecuta** (`if mode == "auto" || mode ==
    "path"`), de modo que el valor pedido es el que se ejecuta, sea más nuevo o más viejo que el local.
    **Verificado** con el `Makefile` de arriba sobre un módulo desechable con `toolchain go1.26.1` y Go
    local 1.26.6: la receta imprime `go version go1.26.1 darwin/arm64` tras descargarlo sola
    (`go: downloading go1.26.1 (darwin/arm64)`). Con `auto`, el mismo módulo daba go1.26.6.
  - **En local**: toda orden del `Makefile` corre el parche de la directiva, tenga quien contribuye el
    parche que tenga. No hay nada que instalar ni que configurar: si el toolchain no está, el go command lo
    descarga y lo verifica, así que FR-042 (d) y FR-010 se mantienen (el único prerrequisito sigue siendo
    `go` ≥ 1.21 y `git`).
  - **En la CI**: `parseGoVersionFile` de `actions/setup-go` (`src/installer.ts`) devuelve el valor de la
    directiva `toolchain` cuando existe, y **solo** en su ausencia cae en la directiva `go`; `findMatch`
    resuelve con `semver.satisfies(version, versionSpec)`, de modo que una versión exacta se resuelve a esa
    versión exacta y **nunca** al último parche de la serie. La acción exporta además `GOTOOLCHAIN=local`
    en el entorno del job (`setGoToolchain()`, `src/main.ts`), pero eso **no debilita el pin**: una
    asignación del `Makefile` prevalece sobre el entorno heredado (GNU Make, «Variables from the
    Environment»; solo `make -e` invertiría la precedencia, y no se usa). **Verificado**:
    `GOTOOLCHAIN=local make …` sobre el módulo desechable ejecuta igualmente el toolchain de la directiva.
    Resultado: la CI ejecuta **el parche de la directiva** (`go1.27.1`) por dos vías concordantes, y si
    alguna vez discreparan, manda el `Makefile` —que es la superficie única de invocación de los
    controles—.
  - **Coste en la CI: ninguno.** Cuando el toolchain pedido coincide con el que ya se está ejecutando,
    `select.go` corta en seco sin descargar nada
    (`if gotoolchain == "local" || gotoolchain == gover.LocalToolchain() { … return }`). Como setup-go
    instala precisamente el parche de la directiva, el pin es un no-op. **Verificado** (con los valores
    del momento): con Go local 1.26.6 y el pin en `go1.26.6`, la receta no imprime ninguna línea
    `downloading`.
- **Alternativas rechazadas**:
  - **La directiva `toolchain` sola, sin exportar `GOTOOLCHAIN`** (la decisión de la ronda anterior, aquí
    corregida): cierra la divergencia solo hacia abajo —protege de un Go **más viejo**— y la deja abierta
    hacia arriba, que es el caso frecuente: cualquier parche nuevo de Go instalado en local produce un
    toolchain distinto al de la CI, con `make vuln` capaz de dar veredictos distintos a ambos lados
    (SC-004, FR-042 a). Además obligaba a `check-tools` a comparar cadenas contra un valor que un Go más
    nuevo nunca cumple, con dos remedios que no funcionaban en ese caso (dejar `GOTOOLCHAIN=auto` sigue
    eligiendo el local más nuevo; instalar el parche fijado exige degradar la instalación de Go).
  - **`GOTOOLCHAIN=<nombre>+auto`** (p. ej. `go1.27.1+auto`): parece más tolerante, pero reabre justo la
    puerta que se quiere cerrar. Con el sufijo `+auto`, `select.go` vuelve a leer los `go.mod` y **sube**
    el toolchain si encuentra una directiva `go` o `toolchain` mayor. Desde el módulo raíz no cambiaría
    nada —el valor sale de su propia directiva—, pero las órdenes que operan sobre los módulos de
    herramientas (`go tool -modfile=tools/<n>/go.mod …`, [D2](#d2--herramientas-de-los-controles-un-módulo-go-por-herramienta-bajo-tools))
    sí podrían subir en silencio si un bump de Dependabot eleva el `go` de una herramienta: los
    analizadores de `golangci-lint` pasarían a compilarse con otro parche sin que nadie se entere. Con el
    nombre a secas ese caso **falla ruidosamente y con la salida escrita**. **Verificado** con los valores
    del momento (pin `go1.26.6` y un submódulo desechable con una directiva `go` mayor, `go 1.27.0`):
    `go: go.mod requires go >= 1.27.0 (running go 1.26.6; GOTOOLCHAIN=go1.26.6)`, exit ≠ 0. El mensaje
    nombra las dos versiones y la variable, y el arreglo es el commit de una línea sobre la directiva
    `toolchain`. Un fallo accionable es preferible a un cambio silencioso de parche: mismo criterio que en
    [D19](#d19--govulncheck-sin-red-caso-límite-del-spec).
  - **Fijar `GOTOOLCHAIN` en el entorno del job de CI, o con `go env -w` en local**: dos fuentes de verdad
    más (los flujos, y la configuración personal de cada máquina) que pueden discrepar de `go.mod`, y en el
    caso de `go env -w` un estado invisible en el control de versiones que además contamina otros
    proyectos. Derivarlo del `Makefile` mantiene una sola fuente de verdad, versionada, y aplica igual en
    los dos lados.
  - **`go-version: '1.27'` o `'~1.27.0'` con `check-latest: true` en `ci.yml`/`nightly.yml`**, dejando
    `go.mod` sin `toolchain`: sí resuelve al último parche publicado —es la forma correcta de expresar «el
    último parche» en la acción—, pero mueve la versión de Go fuera del control de versiones del módulo y
    la desacopla de la local, que sigue siendo la que cada quien tenga instalada. Con eso, el parche que
    evalúa `govulncheck` en la PR y el que lo evalúa en local pueden diferir, que es exactamente el defecto
    que hay que cerrar; además el salto de parche llega sin PR ni revisión (el mismo defecto de
    `go-version: stable`, en menor grado) y duplica la fuente de verdad de la versión de Go en tres
    ficheros (`go.mod` y los dos flujos).
  - **`go-version: stable`**: desacopla la versión de CI de la del `go.mod` y hace que un salto de **minor**
    llegue sin PR ni revisión. Contradice FR-042 (a).
- **Coste asumido —que el parche se quede atrás— y cómo se cubre sin depender de nadie**:
  - El detector es `govulncheck`, que ya es un gate: cuando se publique un parche de Go que corrija un
    fallo de la biblioteca estándar, `make vuln` **empieza a fallar** sobre el parche fijado, tanto en
    la PR como en el flujo `nightly` sobre `main` (FR-027), que existe precisamente para que un hallazgo
    nuevo aparezca sin esperar a la siguiente propuesta de cambio. El arreglo es subir la directiva
    `toolchain`. El control no se vuelve un adorno: falla ruidosamente y con un hallazgo accionable.
    El pin exportado por el `Makefile` es además lo que hace **fiable** a este detector en local: sin él,
    quien tuviera ya instalado el parche corregido vería `make vuln` en verde sobre una stdlib que la CI
    no está usando, y el hallazgo solo aparecería en la PR.
  - `CONTRIBUTING.md` documenta el procedimiento «cómo se sube el parche de Go»: cambiar `toolchain` en
    `go.mod`, ejecutar `make ci`, un commit `chore(deps): go1.27.x`.
  - **Pendiente de verificar en la implementación** (obligación para `tasks.md`, no se asume aquí): si el
    ecosistema `gomod` de Dependabot propone la subida de la directiva `toolchain`. Si la propone, la
    actualización queda además cubierta por FR-028 sin trabajo manual; si no, el mecanismo de detección
    sigue siendo el párrafo anterior y basta con dejarlo escrito en `CONTRIBUTING.md`. No se ha podido
    comprobar aquí porque exige una ejecución real de Dependabot sobre el repositorio, que todavía no
    existe; **no se da por hecho** en ninguna de las dos direcciones.
- **Efecto en `check-tools`** (FR-010, ver [D18](#d18--comprobación-de-prerrequisitos-fr-010)): al exportar
  el `Makefile` el pin, `check-tools` **deja de comparar la instalación de quien contribuye contra la
  directiva**. El parche ya no es algo que el contribuidor deba tener, sino algo que la orden garantiza; lo
  único que puede fallar es que el toolchain fijado **no se pueda obtener**. La comprobación pasa a ser
  «`go` presente y capaz de cambiar de toolchain, `git` presente, y el toolchain fijado obtenible», y el
  mensaje de fallo es el de esa causa real, no el de una versión local «equivocada» que ya no lo es.
- **Por qué go1.27.1** (corrección del 2026-09-11, ronda de revisión final, motivo [i] del juez B): la
  constitución («Restricciones técnicas») y FR-001 piden «Go estable actual», y *Assumptions* del spec lo
  interpreta como «la última versión estable publicada en el momento de implementar el hito». La ronda
  anterior de este documento fijaba `go1.26.6` porque era el `go` instalado en la máquina —eso es lo que
  verificaba `go version`—, **no** porque fuera la última publicada: la comprobación correcta es contra el
  proxy de módulos, no contra la instalación local. Hecha esa comprobación
  (`go list -m -json golang.org/toolchain@v0.0.1-<versión>.darwin-arm64`, que responde con la fecha de
  publicación o con `404 Not Found`): go1.27.0 se publicó el 2026-08-18; go1.27.1 y go1.26.8 el
  2026-08-28; go1.27.2 y go1.26.9 no existen a 2026-09-11. La estable actual es, por tanto, `go1.27.1`, y
  el hito se implementó (2026-09-10) y se corrigió (2026-09-11) después de su publicación.
  - **Verificado** que las herramientas fijadas funcionan con Go 1.27: con `toolchain go1.27.1` en `go.mod`
    y `go 1.27.1` en los cuatro `tools/*/go.mod`, `make ci` termina en verde (`golangci-lint` v2.13.2
    `0 issues`, `go test -race` ok, `govulncheck` sin hallazgos, `gitleaks` sin fugas, `go mod verify` en
    los cinco módulos, `go mod tidy -diff` vacío en la raíz y en cada módulo de herramienta) y
    `go version -m bin/kitlegal` responde `go1.27.1`.
  - La directiva `go` sube a `1.27.0` por coherencia con la regla de esta decisión —«`go X.Y.0` es el
    suelo del lenguaje y `toolchain goX.Y.Z` el parche»—: `kitlegal` es un binario, no una biblioteca, y
    no hay ningún consumidor que necesite compilarlo con Go 1.26. Los `tools/*/go.mod` pasan a `go 1.27.1`,
    que es lo que `go get -tool` escribiría hoy al crearlos.
  - **Alternativa rechazada**: quedarse en la rama 1.26 subiendo solo a `go1.26.8`. Habría sido la salida
    si alguna herramienta fijada no soportara aún Go 1.27; verificado que no es el caso, no hay motivo
    para fijar un minor que ya no es el estable actual, y hacerlo dejaría FR-001 incumplido con una
    justificación falsa.
  - **Alternativa rechazada**: `toolchain go1.27.1` manteniendo `go 1.26.0`. Mantiene el suelo del lenguaje
    un minor por debajo del parche sin ninguna razón que lo exija, y contradice la propia regla «`go X.Y.0`
    + `toolchain goX.Y.Z`» de esta decisión.
- **Nota**: `go test` de Go 1.27+ ejecuta `stdversion` por defecto; con el `go` directive en `1.27.0` el
  código no puede usar API posterior al suelo declarado sin que el test lo señale. El suelo del lenguaje y
  el parche de construcción son dos cosas distintas y por eso son dos directivas distintas.
- **Rastro**: FR-001, FR-010, FR-016, FR-027, FR-042 (a) y (d), SC-004, *Assumptions* del spec,
  constitución «Restricciones técnicas».

---

## D2 · Herramientas de los controles: un módulo Go por herramienta bajo `tools/`

- **Decisión**: cada herramienta vive en su propio módulo dedicado, sin código fuente, solo con su
  directiva `tool`:

  ```
  tools/golangci-lint/{go.mod,go.sum}   tool github.com/golangci/golangci-lint/v2/cmd/golangci-lint
  tools/govulncheck/{go.mod,go.sum}     tool golang.org/x/vuln/cmd/govulncheck
  tools/gitleaks/{go.mod,go.sum}        tool github.com/zricethezav/gitleaks/v8
  tools/lefthook/{go.mod,go.sum}        tool github.com/evilmartians/lefthook
  ```

  El `Makefile` las invoca con `go tool -modfile=tools/<nombre>/go.mod <nombre> …`.

- **Alternativas rechazadas**:
  - **Un único módulo `tools/` compartido por las cuatro herramientas** (era la opción obvia y la que
    sugería la respuesta a Q5 del `clarify`). **Se rechaza porque no compila**: verificado que al poner
    `golangci-lint` + `gofumpt` + `gitleaks` en el mismo módulo, la Selección de Versión Mínima eleva
    `github.com/charmbracelet/x/ansi` a una versión incompatible con el
    `github.com/charmbracelet/x/cellbuf` que exige gitleaks, y `go tool … gitleaks` falla con
    `b.SlowBlink undefined (type ansi.Style has no field or method SlowBlink)` y una decena de errores
    más. Aislado en su propio módulo, `go tool … gitleaks version` funciona (verificado). Un módulo por
    herramienta hace que ese choque sea estructuralmente imposible, no algo que haya que parchear con
    `replace`/`exclude` en cada bump de Dependabot.
  - `go install <mod>@<versión>` desde el `Makefile`: no queda cubierto por ningún `go.sum`
    (incumple FR-042 b), contamina `$GOBIN` de quien contribuye y Dependabot no ve la versión
    (incumple FR-042 c).
  - Directivas `tool` en el `go.mod` del producto: arrastra ~200 dependencias indirectas al fichero de
    dependencias del producto (incumple FR-005 y FR-042 e) y hace ruidoso `govulncheck`.
  - `golangci-lint-action`, `gitleaks-action` y demás acciones de terceros que descargan binarios:
    incumple FR-025 («la CI no aplica ningún control por una vía distinta de las órdenes del `Makefile`»)
    y rompe la equivalencia local ↔ CI de SC-004.
  - `mise`/`asdf`: declarados fuera de alcance por el propio spec.
- **Verificado**: `go help tool` documenta `-modfile` («The -modfile=file.mod build flag causes tool to
  use an alternate file instead of the go.mod in the module root directory»);
  `go get -tool -modfile=…`, `go tool -modfile=…`, `go mod verify -modfile=…` y
  `go mod tidy -diff -modfile=…` funcionan los cuatro. El fichero de sumas se deriva del `-modfile`
  (`tools/<n>/go.sum`), que es exactamente lo que Dependabot necesita ver.
  `go list ./...` desde la raíz **no** incluye `tools/**`, porque cada subdirectorio con `go.mod` propio
  queda fuera del patrón: el producto no se contamina.
- **Cómo cumple FR-042**: (a) versión fijada en `tools/<n>/go.mod`, misma en local y CI porque es la
  misma orden del `Makefile`; (b) integridad verificable con `tools/<n>/go.sum` + `go mod verify`;
  (c) Dependabot con una entrada `gomod` por directorio; (d) el único prerrequisito sigue siendo `go`
  (y `git`); (e) el `go.mod` del producto no cambia.
- **Rastro**: FR-042, FR-025, FR-019, spec §Clarifications Q5, constitución §V y «Criterio de decisión
  autónoma» §1.

### D2.1 · Ruta de módulo de gitleaks

`github.com/gitleaks/gitleaks/v8` **no existe como módulo**: `go get` responde
`module declares its path as: github.com/zricethezav/gitleaks/v8`. La ruta correcta es
`github.com/zricethezav/gitleaks/v8` (v8.30.1 en el momento de escribir esto). Se anota aquí porque la
ruta «obvia» derivada del nombre de la organización en GitHub es la equivocada.

### D2.2 · `go mod tidy` obligatorio tras `go get -tool`

Verificado: `go get -tool` deja el `go.sum` incompleto y `go mod tidy -diff -modfile=…` falla con
diferencias. Hay que ejecutar una vez `go mod tidy -modfile=tools/<n>/go.mod` al crear cada módulo de
herramienta; después el `-diff` sale vacío. Es un paso de la tarea de creación, no un fallo del diseño.

---

## D3 · `gofumpt` y `goimports` solo a través de `golangci-lint`, sin pin propio

- **Decisión**: no hay un módulo `tools/gofumpt/`. El formato lo aplica `golangci-lint` mediante su
  sección `formatters` (`gofumpt` con `extra-rules: true`, y `goimports`), y por tanto la versión de
  `mvdan.cc/gofumpt` queda fijada como dependencia indirecta explícita en
  `tools/golangci-lint/go.{mod,sum}` (verificado: `mvdan.cc/gofumpt v0.12.0`), donde Dependabot la ve.
- **Alternativas rechazadas**: pinar `mvdan.cc/gofumpt` aparte y llamarlo directamente desde el gancho de
  pre-commit. Se rechaza porque introduce **dos** versiones de gofumpt en el proyecto —la del pin y la que
  `golangci-lint` lleva embebida— que pueden discrepar tras cualquier bump. Un fichero que `make fmt`
  deja «formateado» y que `make ci` sigue rechazando (o al revés) destruye justo la propiedad que FR-012
  y SC-004 exigen: mismo control, misma configuración, mismo veredicto.
- **Cómo cumple FR-042 para `gofumpt`**: (a) versión explícita en `tools/golangci-lint/go.mod`;
  (b) cubierta por `tools/golangci-lint/go.sum`; (c) Dependabot la actualiza al actualizar
  `golangci-lint`; (d) sin prerrequisitos extra; (e) fuera del `go.mod` del producto.
- **Rastro**: FR-012, FR-042, `docs/ROADMAP.md` §3 («gofumpt + goimports (vía golangci-lint formatters)»).

---

## D4 · Dos modos del control de formato (FR-012)

- **Decisión**:
  - modo **mutante** → `make fmt` → `golangci-lint fmt ./...`
  - modo **verificación** → `make fmt-check` → `golangci-lint fmt --diff ./...`

  `make ci` invoca **solo** `fmt-check`. `make fmt` no forma parte de `ci`.
- **Verificado**: `golangci-lint fmt --help` expone `-d, --diff` («Display diffs instead of rewriting
  files»), y sobre un fichero mal formateado imprime el diff unificado **y termina con código ≠ 0**,
  sin tocar el fichero. Es exactamente el gate demostrable que pide FR-012 y el escenario 6 de US1.
- **Alternativa rechazada**: `gofumpt -l` + `test -z`. Duplica la configuración del formateador fuera del
  `.golangci.yml` (ver D3) y pierde `goimports`.
- **Nota**: `fmt-check` es una orden adicional a las doce que enumera FR-006 («al menos»); el modo de
  verificación necesita un nombre propio para poder invocarse desde `ci` y desde la CI.
- **Rastro**: FR-006, FR-007, FR-012, US1 escenarios 6 y 7, SC-002.

---

## D5 · Sin dependencias de producto en H0, y qué pasa con `go.sum`

- **Decisión**: el `go.mod` del producto no declara ninguna dependencia externa. El punto de entrada usa
  solo la biblioteca estándar (`os`, `io`, `fmt`) y `cmd/kitlegal/main_test.go` usa solo `testing`.
  **En consecuencia, el repositorio no tiene `go.sum` en la raíz**, porque un módulo sin dependencias
  externas no genera ninguno.
- **Por qué**: Kong es alcance de H1 y `testify`/`testscript` son alcance de H1/H3 según «Fuera de
  alcance» del spec; introducirlos en H0 sería anticipar una fase (criterio §2). Con un solo verbo
  (`version`) y un `switch` sobre `os.Args`, Kong no aporta nada que H0 necesite.
- **Tensión con el spec, resuelta**: FR-005 y SC-008 dan por hecho que existe «el fichero de sumas de
  verificación». Se cumple el **fin** de esos requisitos —integridad verificable de todo lo que el
  proyecto descarga— con `tools/<n>/go.sum` (uno por herramienta, todos no vacíos) y con
  `go mod verify` sobre el módulo raíz y sobre cada módulo de herramienta dentro de `make ci`. Sobre el
  módulo raíz `go mod verify` responde `all modules verified` de forma vacía, que es el resultado
  correcto para un módulo sin dependencias.
- **Alternativa rechazada**: añadir `testify` en H0 solo para que exista un `go.sum` en la raíz. Es una
  dependencia sin necesidad funcional introducida para satisfacer la letra de un criterio: exactamente la
  «ñapa» que prohíbe el criterio §1. Además contradice «Fuera de alcance».
- **Consecuencia para el gate**: se deja anotado para `/speckit-analyze` como divergencia consciente
  spec ↔ plan, no como omisión.
- **Rastro**: FR-005, FR-019, SC-008, «Fuera de alcance» del spec, constitución §V y «Criterio de decisión
  autónoma» §1 y §2.

---

## D6 · Estructura del punto de entrada: `run(args, stdout, stderr) int`

- **Decisión**: `cmd/kitlegal/main.go` contiene

  ```go
  var version, commit, fecha = "dev", "none", "unknown"   // inyectadas por -ldflags

  func main() { os.Exit(run(os.Args[1:], os.Stdout, os.Stderr)) }

  func run(args []string, stdout, stderr io.Writer) int { … }
  ```

  `main()` no hace nada más que llamar a `run` y propagar su código: es la única línea no cubierta por
  tests. `cmd/kitlegal/main_test.go` (paquete `main`, table-driven, `t.Parallel()`) ejerce `run` con
  escritores en memoria.
- **Alternativas rechazadas**:
  - Extraer la lógica a `internal/version` o `internal/cli`: prohibido por FR-002 y por la respuesta a Q4
    del `clarify` («H0 no crea `internal/`»).
  - Tests e2e sobre el binario compilado (`testscript`): alcance de H1.
- **Por qué**: es el patrón estándar para hacer testable un `main` sin crear paquetes, cubre FR-040 y el
  umbral global de FR-029 con el propio código del hito, y respeta la regla de dependencia «solo
  `internal/cli` y `cmd/` llaman a `os.Exit`».
- **Rastro**: FR-002, FR-003, FR-004, FR-040, FR-029, spec §Clarifications Q3 y Q4.

---

## D7 · Comportamiento de `run` ante entradas que no son `version`

- **Decisión**: `version` es el único verbo reconocido. Sin argumentos o con un verbo desconocido, `run`
  escribe en **stderr** una línea de uso que nombra el verbo disponible y devuelve **2**.
- **Por qué**: el binario tiene que hacer *algo* cuando no le pasan `version`, y de los códigos estables
  del proyecto el 2 es literalmente «args». Devolver 0 ante un verbo inexistente contradiría FR-039 de
  facto, aunque no lo diga ningún requisito. Devolver 1 introduciría un código fuera de la tabla.
- **Fuera de alcance deliberado**: no hay `--help`, ni `-h`, ni `--version`, ni banderas globales, ni
  sobre de salida, ni `--json`. Todo eso es H1 (criterio §2).
- **Rastro**: FR-039, *Assumptions* del spec, `CLAUDE.md` («Exit codes estables»).

---

## D8 · Datos de versión inyectados y formato de la salida

- **Decisión**: `make build` e `install` inyectan
  `-X main.version=$(VERSION) -X main.commit=$(COMMIT) -X main.fecha=$(FECHA)` con

  ```make
  VERSION ?= $(shell git describe --tags --always --dirty 2>/dev/null || echo 0.0.0-dev)
  COMMIT  ?= $(shell git rev-parse HEAD 2>/dev/null || echo desconocido)
  FECHA   ?= $(shell date -u +%Y-%m-%dT%H:%M:%SZ)
  ```

  Los valores por defecto en el código (`dev`/`none`/`unknown`) existen solo para `go run` y `go test`;
  el binario que produce `make build` nunca los muestra (FR-004). El formato de salida es texto plano,
  tres líneas, definido en [`contracts/cli-version.md`](./contracts/cli-version.md).
- **`COMMIT` es el SHA completo** de `git rev-parse HEAD`, para que SC-003 («el commit impreso coincide
  con el de la revisión construida») se compruebe con una igualdad exacta y no con un prefijo.
- **Alternativa rechazada**: leer los datos de `runtime/debug.ReadBuildInfo()` (que Go rellena solo con
  `go build` desde un repo git). Se rechaza porque `go test` no la puebla igual y porque FR-004 exige
  explícitamente inyección desde la orden de construcción.
- **Nota sobre reproducibilidad**: `FECHA` es la hora de construcción, así que dos `make build` del mismo
  commit producen binarios distintos byte a byte. Es lo que pide FR-003 («fecha de construcción»); la
  reproducibilidad del artefacto publicado es asunto de H6 (goreleaser). No se añade soporte de
  `SOURCE_DATE_EPOCH` en H0 por criterio §2.
- **Rastro**: FR-003, FR-004, SC-003, `--trimpath`/`CGO_ENABLED=0` de la constitución.

---

## D9 · Configuración de `golangci-lint`

- **Decisión**: `.golangci.yml` en formato v2 (`version: "2"`), con `linters.default: standard` más el
  `enable` explícito de todos los de FR-013 (`revive`, `gocritic`, `errorlint`, `exhaustive`, `nilerr`,
  `bodyclose`, `noctx`, `contextcheck`, `sqlclosecheck`, `rowserrcheck`, `nolintlint`, `misspell`,
  `gocyclo`, `dupl`, `testifylint`, `thelper`, `paralleltest`, `tparallel`, `gosec`) más `forbidigo`
  (FR-014). `errcheck`, `govet` y `staticcheck` llegan por `default: standard`.
  `depguard` **no** se activa (es H1, FR-014 lo dice explícitamente).
- **Alternativa rechazada**: `default: none` con lista cerrada. Apagaría `unused` e `ineffassign`, que
  `standard` trae gratis y que ningún requisito pide apagar; menos estricto sin ganancia.
- **`forbidigo` acotado por ruta**: la regla prohíbe `^fmt\.Print(|f|ln)$` y se limita a `internal/**`
  con una exclusión invertida:

  ```yaml
  linters:
    exclusions:
      rules:
        - linters: [forbidigo]
          path-except: '^internal/'
  ```

  Con eso, `cmd/kitlegal/main.go` puede escribir en stdout en H0 y un `fmt.Println` bajo `internal/`
  falla (SC-005). La ampliación a «solo `internal/render` escribe en stdout» es H1.
- **`misspell` y el español**: el diccionario de `misspell` es **solo inglés** (`locale: US|UK`); no
  existe diccionario español. La lectura honesta de «misspell (es/en)» de `docs/ROADMAP.md` §3 es:
  `misspell` corre con `locale: US` sobre todo el repositorio, incluidos los comentarios y literales en
  español, y los falsos positivos que produzca el español se neutralizan uno a uno en
  `linters.settings.misspell.ignore-rules`, cada uno con su comentario. En H0 no hay ninguno. Se anota
  como límite conocido de la herramienta, no como incumplimiento.
- **Ajustes**: `dupl.threshold: 100`, `gocyclo.min-complexity: 13`, `errcheck.check-type-assertions: true`,
  `nolintlint.require-explanation: true` + `require-specific: true` (hace ejecutable la prohibición de
  `//nolint` sin justificación del criterio §1), `run.tests: true`,
  `issues.max-issues-per-linter: 0`, `issues.max-same-issues: 0`.
- **`go vet`**: la constitución lo lista aparte en la capa 1 de gates; lo cubre `govet` dentro de
  `golangci-lint`, que es el mismo analizador. No se añade un paso `go vet` separado (sería el mismo
  control ejecutado dos veces).
- **Verificado**: `golangci-lint` v2.13.2 construido con `go tool`; `run --fast-only` existe (el `--fast`
  de v1 fue renombrado en v2, dato relevante para el gancho de pre-commit).
- **Rastro**: FR-013, FR-014, SC-005, `docs/ROADMAP.md` §3, constitución §IV y «Gates» capa 1.

---

## D10 · Detección de secretos y vía de exclusión trazable

- **Decisión**: `make secrets` → `go tool -modfile=tools/gitleaks/go.mod gitleaks dir . --redact --no-banner`.
  La **misma** orden se usa en `make ci` y en el gancho de pre-commit: un único camino, un único veredicto.
- **Verificado**: en gitleaks v8.30.1 los subcomandos son `dir`, `git` y `stdin`; `detect` y `protect`
  (la sintaxis que aparece en casi toda la documentación antigua) ya no existen.
- **Vía de exclusión** (caso límite del spec): fichero `.gitleaksignore` en la raíz, con la huella
  (*fingerprint*) del hallazgo concreto y un comentario que justifique cada línea; `CONTRIBUTING.md`
  documenta el procedimiento. No se añade `.gitleaks.toml`: se usa el conjunto de reglas por defecto.
  **H0 sí crea el fichero, con dos huellas reales.** La premisa original de esta decisión —que en H0 no
  había ningún falso positivo que excluir— era falsa: el árbol heredado trae dos líneas de ejemplo
  marcadas `// DON'T` en la documentación de la skill `golang-security` del paquete vendorizado
  `samber/cc-skills-golang`, que la regla `generic-api-key` detecta y que no son credenciales. Sin
  exclusión, `make secrets` no puede quedar en verde y el control no queda instalado, que es justo lo que
  el hito entrega. El fichero nace con contenido justificado, no como marcador de posición, así que SC-008
  se cumple; figura en el inventario de [data-model.md](./data-model.md) y entre las rutas de T001.
  Decisión humana del 2026-09-10 en
  [gates/decision-humana-secretos.md](./gates/decision-humana-secretos.md).
- **Verificado sobre el árbol real** (gitleaks v8.30.1, subcomando `dir`): `.gitignore` **no** se
  respeta, así que dejar de versionar un directorio no lo saca del escaneo —solo borrarlo del disco—; una
  ruta desnuda en `.gitleaksignore` tampoco excluye nada; la huella `fichero:regla:línea` sí. Es la única
  vía por hallazgo que funciona en este modo.
- **Alternativa rechazada**: `.gitleaks.toml` con `[allowlist]` por expresión regular. Excluye por patrón,
  no por hallazgo, así que puede silenciar secretos futuros que nadie ha visto; `.gitleaksignore` excluye
  exactamente una coincidencia identificada, que es lo que pide «trazable, sin desactivar el control».
- **Rastro**: FR-018, SC-009, «Edge Cases» del spec.

---

## D11 · Qué comprueba la cadena de suministro y sobre qué módulos

- **Decisión**:
  - `make mod-verify`: `go mod verify` sobre el módulo raíz **y** sobre cada `tools/<n>/go.mod`.
  - `make mod-tidy-check`: `go mod tidy -diff` **solo sobre el módulo raíz**.
- **Por qué el `tidy -diff` no se extiende a los módulos de herramientas**: FR-020 habla de «el fichero de
  dependencias» del producto. En un módulo de herramientas, lo que determina qué binario se ejecuta es la
  directiva `tool` más el `go.sum`; que su `go.mod` esté o no «saneado» es cosmético y su comprobación
  obliga a descargar además el grafo de dependencias *de test* de cada herramienta (verificado: `go mod
  tidy` sobre el módulo de `gofumpt` descarga `go-quicktest`, `kr/pretty`, `kr/text`), encareciendo la CI
  sin cubrir ningún riesgo nuevo. El riesgo real —una versión de herramienta alterada a mano— lo cubre
  `go mod verify` contra `tools/<n>/go.sum`.
- **Rastro**: FR-019, FR-020, FR-042 (b), SC-009.

---

## D12 · Un solo flujo de CI que ejecuta `make ci`

- **Decisión**: `.github/workflows/ci.yml` con un job, disparado por `pull_request` y `push` a `main`:
  `checkout` → `setup-go` (con `go-version-file: go.mod` y caché; **sin** `go-version` ni `check-latest`,
  porque la versión sale de la directiva `toolchain` del `go.mod` y el parche efectivo lo fija el
  `GOTOOLCHAIN` que exporta el `Makefile`,
  [D1](#d1--versión-de-go-go-1270--toolchain-go1271-y-el-makefile-exporta-gotoolchain))
  → `actions/cache` de `GOCACHE` +
  `GOMODCACHE` con clave derivada de `tools/*/go.sum` → **`make ci`** → subida del perfil de cobertura.
  Ningún paso aplica un control por su cuenta.
- **Por qué la caché es parte del diseño y no una optimización**: las cuatro herramientas se compilan
  desde fuente la primera vez. En frío eso domina el tiempo del flujo y SC-006 («PR limpia en < 3 min»)
  se mide en caliente (*Assumptions* del spec). El flujo `nightly` sobre `main` (FR-027) tiene aquí un
  efecto de segundo orden que conviene explicitar: mantiene viva la caché de la rama por defecto, de la
  que heredan por *fallback* todas las PR, así que la primera ejecución de una rama nueva ya arranca en
  caliente.
- **Obligación pendiente de medir** (*Assumptions* del spec lo exige explícitamente al plan): hay que
  registrar el tiempo del flujo en **frío** (caché invalidada) y en **caliente**, y dejar ambos números
  en la PR del hito. Es una tarea de `tasks.md`, no algo que un documento pueda afirmar de antemano: el
  número depende del ejecutor de GitHub, no de esta máquina.
- **Contingencia si el escenario en caliente superara los 3 min**: partir `make ci` en jobs paralelos
  (`lint`, `test`, `vuln`, `secrets+cadena`), cada uno invocando **una orden del `Makefile`**, y que la
  orden agregada `ci` siga existiendo para el ciclo local. Corrige el diseño del flujo sin tocar el
  criterio de aceptación ni romper FR-025 ni SC-004.
- **Rastro**: FR-025, FR-027, SC-004, SC-006, *Assumptions* del spec.

---

## D13 · CodeQL semanal

- **Decisión**: `.github/workflows/codeql.yml` con `github/codeql-action` (`init` + `analyze`),
  `languages: go`, `queries: security-extended`, `schedule` semanal, permisos mínimos
  (`security-events: write`, `contents: read`). Sin fichero de configuración aparte.
- **Por qué `security-extended` y no `security-and-quality`**: la suite `-and-quality` añade consultas de
  mantenibilidad que solapan con `golangci-lint`; CodeQL está aquí como SAST (FR-017), y el solape
  produciría hallazgos duplicados en dos herramientas distintas.
- **Por qué CodeQL no pasa por el `Makefile`**: no opera sobre un árbol de trabajo local sino sobre una
  base de datos que construye la propia acción. SC-004 lo excluye de la equivalencia local ↔ CI de forma
  explícita, así que no contradice FR-025 (que restringe el flujo `ci`, no el flujo `codeql`).
- **Rastro**: FR-017, FR-026, SC-004.

---

## D14 · Dependabot

- **Decisión**: `.github/dependabot.yml` con periodicidad semanal y seis entradas: `gomod` en `/`,
  `gomod` en `/tools/golangci-lint`, `/tools/govulncheck`, `/tools/gitleaks`, `/tools/lefthook`, y
  `github-actions` en `/`. Cada entrada `gomod` de herramienta agrupa sus actualizaciones en una sola PR
  (`groups`), porque el interés está en la versión de la herramienta, no en sus indirectas.
- **Por qué una entrada por directorio**: es la consecuencia directa de D2. Dependabot descubre módulos Go
  por ficheros llamados `go.mod` en el `directory` indicado, y no soporta nombres alternativos; ese es
  también el motivo por el que los módulos de herramienta van en subdirectorios con `go.mod` de nombre
  canónico y no en ficheros `go.tools.mod` en la raíz.
- **Rastro**: FR-028, FR-042 (c).

---

## D15 · Cobertura

- **Decisión**: `codecov.yml` declara `coverage.status.project.default` con `target: 70%` e
  `informational: false`, y un componente `internal/core` (`paths: ["internal/core/**"]`) con
  `target: 85%`, también no informativo. `make test` genera `coverage.out`; el flujo `ci` lo publica con
  `codecov/codecov-action`.
- **Por qué el componente no rompe en H0**: Codecov no evalúa un componente cuyo `paths` no casa con
  ningún fichero del informe. `internal/core/**` no existe en H0 y el estado empieza a aplicarse solo
  cuando aparezcan ficheros, sin configuración condicional que haya que editar en H7. **Esto se comprueba,
  no se afirma**: en la PR de prueba limpia hay que verificar que el estado del componente `internal/core`
  no aparece como fallo ni como pendiente que bloquee la fusión (obligación para `tasks.md`). Si apareciera
  bloqueando, la corrección es declarar el componente con `informational: true` **solo hasta H7** dejando
  el motivo en `codecov.yml`, nunca retirar el umbral.
- **`fail_ci_if_error: true`** en el paso de subida: si la subida falla en silencio, el estado de
  cobertura nunca llega y el gate se vuelve vacuo. Un fallo de subida debe verse.
- **Prerrequisito de plataforma: `CODECOV_TOKEN`.** Las versiones actuales de `codecov/codecov-action`
  exigen el token incluso para subir desde ramas del propio repositorio, así que con
  `fail_ci_if_error: true` **su ausencia hace fallar el flujo `ci` por la subida, no por un control**. La
  configuración de secretos de la plataforma está fuera del alcance del hito (nada de lo que el
  repositorio contiene puede crearla), pero no puede quedar tácita: se declara aquí como prerrequisito,
  `CONTRIBUTING.md` lo recoge junto al resto de la configuración del repositorio en GitHub, y `tasks.md`
  lleva la obligación de comprobar en la PR de prueba limpia que el secreto está dado de alta y que el
  paso de subida termina bien. **No** se rebaja `fail_ci_if_error` para esquivarlo: eso convertiría el
  gate de cobertura en un adorno, que es justo lo que la bandera existe para impedir.
- **`coverage.out` y SC-002**: ya está cubierto por el `.gitignore` existente (`*.out`, `coverage.*`), así
  que `make ci` no deja diferencias en `git status`.
- **Rastro**: FR-008, FR-029, SC-002, SC-011, `docs/ROADMAP.md` §1.9.

---

## D16 · Ganchos de pre-commit

- **Decisión**: `lefthook.yml` con un `pre-commit` que ejecuta **órdenes del `Makefile`**, no comandos
  propios: `make fmt` (con `stage_fixed: true` y `glob: "*.go"`), `make lint-fast`
  (`golangci-lint run --fast-only ./...`), `make secrets` y `make mod-tidy-check`.
- **Por qué a través del `Makefile`**: un segundo camino de invocación es un segundo sitio donde la
  configuración puede divergir; toda la arquitectura de gates de este hito se apoya en que hay un único
  camino (SC-004). Se renuncia deliberadamente a acotar el lint y el formato a los ficheros preparados:
  con un repositorio de este tamaño, la optimización no compensa el riesgo de divergencia.
- **Instalación**: orden `make hooks` → `lefthook install`. `README.md` y `CONTRIBUTING.md` la documentan;
  el gancho **no** es la autoridad final (caso límite del spec): la CI ejecuta los mismos controles.
- **Verificado en T003 (2026-09-10, lefthook v1.13.6)**: `stage_fixed: true` **sí** vuelve a preparar los
  ficheros que la orden corrige aunque `run` no reciba `{staged_files}`. Comprobado ejecutando el escenario
  8 completo: con `func  main( )  {` y una línea de comentario en el índice, el gancho ejecutó `make fmt`,
  devolvió la declaración a su forma canónica y el commit resultante contiene **solo** la línea de
  comentario ya formateada; `git diff HEAD -- cmd/kitlegal/main.go` quedó vacío. No hace falta el
  `git add {staged_files}` explícito que preveía la alternativa.
- **Rastro**: FR-022, SC-004, `docs/ROADMAP.md` §3, «Edge Cases» del spec.

---

## D17 · Órdenes con el objeto todavía ausente

Se implementan exactamente como fijó la respuesta a Q1 del `clarify`, con mensajes en español que nombran
el objeto y el hito:

| Orden | Comportamiento | Salida |
|---|---|---|
| `test-integration` | comando real `go test -race -tags=integration ./...` | exit 0 sobre conjunto vacío |
| `test-e2e` | anuncia y termina bien | `test-e2e: sin tests e2e todavía; los aporta H1 (testscript)` · exit 0 |
| `schema-check` | anuncia y termina bien | `schema-check: no hay schemas/ todavía; los aportan H4 (borrador) y H11 (contrato)` · exit 0 |
| `skills-sync` | anuncia y termina bien | `skills-sync: no hay skills/ ni data/*.yaml todavía; los aporta H5` · exit 0 |
| `release` | falla | `release: sin configurar hasta H6 (.goreleaser.yaml)` · exit 1 |

`release` no entra en `ci` ni en `nightly`. Las cinco se documentan en `CONTRIBUTING.md`.

**Rastro**: FR-011, SC-012, spec §Clarifications Q1.

---

## D18 · Comprobación de prerrequisitos (FR-010)

- **Decisión**: una orden `check-tools` del `Makefile`, de la que dependen las órdenes reales, comprueba
  **tres** cosas y nada más, porque tras [D2](#d2--herramientas-de-los-controles-un-módulo-go-por-herramienta-bajo-tools)
  no hay más prerrequisitos externos —las cuatro herramientas las construye `go tool`— y tras
  [D1](#d1--versión-de-go-go-1270--toolchain-go1271-y-el-makefile-exporta-gotoolchain) el parche de Go lo
  fija la propia orden:

  | # | Comprobación | Cómo | Si falla |
  |---|---|---|---|
  | 1 | `go` presente | `command -v go` | Falla nombrando `go`, la versión mínima (**1.21**, la primera capaz de cambiar de toolchain) y `https://go.dev/dl/` |
  | 2 | `git` presente | `command -v git` | Falla nombrando `git` y cómo instalarlo |
  | 3 | El toolchain fijado es **obtenible** | `go version` bajo el `GOTOOLCHAIN` que exporta el `Makefile`, y su salida nombra ese toolchain | Falla nombrando el toolchain fijado, la directiva `toolchain` de `go.mod` de la que sale y el error literal del go command |

- **Qué ha cambiado respecto a la ronda anterior, y por qué**: antes la comprobación 3 era una **comparación
  de cadenas entre la instalación local y la directiva**. Eso era incorrecto: con `GOTOOLCHAIN=auto` la
  directiva es un suelo, de modo que cualquiera con un parche **más nuevo** —la situación normal a las pocas
  semanas de cualquier parche de Go— habría visto fallar `check-tools` sin tener nada mal, y los dos
  remedios que anunciaba el mensaje no le servían: dejar `GOTOOLCHAIN=auto` seguía eligiendo su parche más
  nuevo, e «instalar go1.26.6» significaba degradar su instalación de Go para poder trabajar en el
  proyecto. Con el pin exportado por el `Makefile`, **el contribuidor ya no tiene que tener nada concreto
  instalado**: cualquier Go ≥ 1.21 sirve, y la orden se encarga de que lo que compila y analiza sea el
  parche de la directiva. La comprobación deja de imponer una precondición y pasa a verificar una
  **postcondición** de la propia orden.
- **Por qué la comprobación 3 sigue existiendo si el pin ya lo garantiza**: porque puede no poder
  cumplirse, y entonces conviene que el diagnóstico salga de `check-tools` y no del primer control que se
  tropiece. Las causas reales son dos: un `go` anterior a 1.21, que no entiende `GOTOOLCHAIN` ni sabe
  cambiar de toolchain; y un toolchain que no se puede descargar (sin red, `GOPROXY=off`, o una directiva
  `toolchain` que nombra una versión inexistente). **Verificado** que la segunda produce
  `go: download go1.26.99 for darwin/arm64: toolchain not available` con exit ≠ 0; `check-tools` envuelve
  ese error con el contexto que le falta —de qué fichero y qué línea sale el valor pedido—. Con esto FR-010
  se cumple en su letra («presencia **y versión**»): la presencia se comprueba con 1 y 2, y la versión con
  3, que es donde ahora vive de verdad.
- **El entorno no puede debilitar el pin**: una asignación del `Makefile` prevalece sobre la variable
  heredada del entorno (verificado con `GOTOOLCHAIN=local make …`, [D1](#d1--versión-de-go-go-1270--toolchain-go1271-y-el-makefile-exporta-gotoolchain)),
  así que `GOTOOLCHAIN=local` en la sesión de alguien —o el que exporta `actions/setup-go` en el job— ya no
  es un modo de fallo. Sí queda una palanca deliberada para probar la orden: `make … GO_TOOLCHAIN=<otro>`,
  porque una asignación en la línea de órdenes de `make` tiene precedencia sobre el fichero. Es la que usa
  el escenario 7 de [`quickstart.md`](./quickstart.md) para demostrar el camino de fallo sin tener que
  instalar un Go antiguo.
- **Rastro**: FR-001, FR-010, FR-042 (a) y (d), SC-010, «Edge Cases» del spec.

---

## D19 · `govulncheck` sin red (caso límite del spec)

- **Decisión**: `make vuln` ejecuta `govulncheck ./...` sin ninguna bandera que degrade el fallo. Sin red,
  `govulncheck` no puede consultar la base de datos y **termina con error**, no en verde. `CONTRIBUTING.md`
  lo documenta de forma explícita: *el análisis de vulnerabilidades requiere red; sin ella la orden falla,
  y ese fallo no debe confundirse con «no hay vulnerabilidades»*.
- **Alternativa rechazada**: capturar el error de red y devolver 0 con un aviso. Es exactamente «capturar
  errores para silenciarlos» (criterio §1) y convierte un control en un adorno.
- **Alcance del análisis**: `govulncheck` evalúa el código del módulo, sus dependencias **y la biblioteca
  estándar del toolchain en uso**. De ahí que el parche de Go tenga que ser el mismo en local y en la CI
  —y que no baste la directiva `toolchain`, que es solo un suelo: lo fija el `GOTOOLCHAIN` que exporta el
  `Makefile` ([D1](#d1--versión-de-go-go-1270--toolchain-go1271-y-el-makefile-exporta-gotoolchain))— y que este control sea, además, el detector de un
  `toolchain` que se ha quedado atrás: un parche de seguridad de la stdlib se manifiesta como un hallazgo
  en `make vuln`, en la PR y en el flujo `nightly`.
- **Rastro**: FR-016, FR-027, SC-004, «Edge Cases» del spec.

---

## D20 · Formato de los ADR

- **Decisión**: MADR corto (`Context and Problem Statement` / `Considered Options` / `Decision Outcome` /
  `Consequences`), con los encabezados en español, numerados `docs/ADR/NNNN-slug.md`. Los cuatro ADR de
  H0 **registran decisiones ya cerradas** en `CLAUDE.md` y en la constitución; no las reabren ni las
  matizan: se limitan a dejar por escrito contexto, opciones consideradas y consecuencias.
- **Rastro**: FR-024, FR-033 a FR-036, constitución «Restricciones técnicas» y «Gobernanza».

---

## Cuestiones NEEDS CLARIFICATION resueltas

Ninguna queda abierta. Las cinco preguntas del `clarify` ya estaban cerradas en el spec; este documento
resuelve las que aparecieron al bajar a diseño (D2 módulo por herramienta, D2.1 ruta de gitleaks,
D3 gofumpt sin pin propio, D5 ausencia de `go.sum` raíz, D7 verbo desconocido, D9 misspell en español,
D11 alcance del `tidy -diff`) y deja **cinco obligaciones de verificación** para `tasks.md`:

1. Medir el flujo `ci` en frío y en caliente y registrar ambos números (D12).
2. ~~Comprobar el comportamiento de `stage_fixed` de lefthook v1.13.6 (D16).~~ **Resuelta en T003**: sí
   re-prepara; ver D16.
3. Comprobar si el ecosistema `gomod` de Dependabot propone la subida de la directiva `toolchain` de
   `go.mod`; si no lo hace, dejar el procedimiento manual escrito en `CONTRIBUTING.md` (D1).
4. Comprobar en la PR de prueba limpia que el estado del componente `internal/core` de Codecov no aparece
   como fallo ni como pendiente que bloquee, dado que en H0 su `paths` no casa con ningún fichero (D15).
5. Comprobar que el secreto `CODECOV_TOKEN` está dado de alta en el repositorio y que el paso de subida de
   cobertura termina bien con `fail_ci_if_error: true`; si no lo estuviera, el flujo `ci` fallaría por la
   subida y no por un control (D15).
