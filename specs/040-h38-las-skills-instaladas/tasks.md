---

description: "Tareas de H38 · Las skills instaladas se actualizan con el binario (ADR 0060)"
---

# Tasks: H38 · Las skills instaladas se actualizan con el binario (ADR 0060)

**Input**: `specs/040-h38-las-skills-instaladas/` — [spec.md](spec.md), [plan.md](plan.md), [research.md](research.md),
[data-model.md](data-model.md), [quickstart.md](quickstart.md) y [contracts/](contracts/) (`aviso-y-reinstalacion.md`,
`aceptacion.md`).

**Tests**: obligatorios (constitución III). La primera tarea escribe la suite de aceptación, que el workflow congela;
cada tarea posterior trae su test y el cambio que lo pone en verde, y deja `make ci` en verde por sí sola.

**Organización**: las fases van en el orden en que se ejecutan, una tras otra. La fundacional deja las tres piezas que
el aviso compone (el dominio que decide, el candado y el permiso del kernel) sin cambiar todavía lo que hace ninguna
invocación; la de US1 las compone y, en la misma tarea, cambia los e2e de H19 que pasan a esperar la reinstalación. La
línea de cada tarea declara todas las rutas que crea o modifica (el guardián rechaza cualquier otra; `go.mod`,
`go.sum`, `CHANGELOG.md` y el directorio del feature siempre se admiten, y declarar `x.go` admite `x_test.go`).

## Format: `[ID] [P?] [Story] Descripción`

- **[P]**: puede ir en paralelo (ficheros distintos, sin dependencia de una tarea sin terminar): T003 (solo
  `internal/cache`), T006 (solo depende de T002) y T007 (solo documentación).
- **[Story]**: la historia de [spec.md](spec.md) a la que sirve. Las fundacionales (T002 a T004) no llevan etiqueta y
  citan sus FR; US2 y US3 no tienen tarea propia: su decisión la construyen T002 y T004, y su comportamiento visible lo
  enciende T005 y lo fijan `aceptacion/tocadas.txtar` y `aceptacion/excluidas.txtar`.
- **[aceptacion]**: la suite congelada (solo T001). **[datos]**: solo T005, la única que toca ficheros de
  `internal/app/testdata/` (los tres guiones de H19 de FR-018; plan.md «Fixtures»). Ninguna toca `schemas/`.

## Convenciones de todas las tareas

- `make ci` en verde al terminar cada una, en primer plano y por objetivos si hace falta (`make fmt-check`, `make lint`,
  `make test`, `make test-integration`, `make vuln`, `make schema-check`). Ninguna usa la red más que la de las
  herramientas de Go; ninguna graba (no hay datos externos: plan.md «Datos externos»; ni manifiesto `grabaciones.json`
  ni `TestGrabar*`); ninguna escribe en `data/`, en `docs/SOURCES.md` ni fuera del repositorio y de `$TMPDIR`; ninguna
  abre una sesión con modelo ni ejecuta `make evals` ni `make evals-sondeo`.
- Ningún test escribe en la cuenta de quien lo ejecuta: todo test que pueda llegar al candado fija `HOME` y
  `KITLEGAL_CACHE_DIR` dentro de `t.TempDir()` con `t.Setenv`, sin `t.Parallel()` (research.md D7).
- No cambian `kitlegal skills install`, `list` ni `doctor`, la forma de `kitlegal.json`, los hosts, ninguna `SKILL.md`,
  referencia, eval, `data/`, la web, `.golangci.yml`, `Makefile`, `.github/`, `scripts/`, `cmd/`, el arnés
  (`internal/app/e2e_test.go`), `internal/arch_test.go`, `internal/app/ejemplo/kitlegal-e2e/main.go`,
  `h21-mcp-protocolo.txtar` ni ningún guion o suite congelada distintos de los seis de FR-018 (FR-016, FR-018).
  Ninguna tarea lo cambia: las dos referencias de `normativa` y `plazos-y-territorio` que cambian las cambia, después
  de las tareas, la reparación del cierre (research.md, «Reparaciones del cierre», R1).
- Nada nuevo: ni dependencia (la precedencia SemVer es propia, research.md D6), ni paquete, orden, bandera, variable,
  código de salida o sobre (plan.md «Summary»). Ningún `panic`, ningún `os.Exit`, nada en la salida estándar; ningún
  error de la reinstalación se propaga (research.md D11). Ningún `//nolint`.
- Los textos son los de los contratos: las tres líneas, las de contracts/aviso-y-reinstalacion.md §3 y data-model §5;
  la decisión, la de §2; el candado, el de §4; los guiones, los de contracts/aceptacion.md. Nada se inventa fuera de
  ellos.
- Cada control se ve fallar con un mutante (la medida pasada del umbral) antes de darlo por bueno (ADR 0029). El mutante
  se aplica sobre una copia del árbol bajo `$TMPDIR`, nunca sobre el árbol versionado, y no se versiona; la tarea
  anota en su informe qué mutante aplicó y qué test vio en rojo.

---

## Phase 1: Setup — la suite de aceptación

**Propósito**: fijar, antes de cualquier cambio, lo que la primera consulta con el binario nuevo tiene que hacer y lo
que no (ADR 0018; plan.md «Aceptación e2e»; contracts/aceptacion.md).

- [X] T001 [aceptacion] Escribir, a partir del spec y de contracts/aceptacion.md §1 a §5, sin código de producto y sin evals (ninguna skill cambia: spec «Fuera de alcance»), los cuatro guiones testscript specs/040-h38-las-skills-instaladas/aceptacion/reinstala.txtar (US1 · FR-001, FR-003, FR-004, FR-005, FR-006, FR-008, FR-012, FR-013 · SC-001, SC-003), specs/040-h38-las-skills-instaladas/aceptacion/tocadas.txtar (US2 · FR-003, FR-004, FR-005, FR-007, FR-013 · SC-002, SC-003, SC-006), specs/040-h38-las-skills-instaladas/aceptacion/excluidas.txtar (US3 y el local inservible de «Edge Cases» · FR-001, FR-002, FR-003, FR-009, FR-013, FR-014, FR-015 · SC-003, SC-006) y specs/040-h38-las-skills-instaladas/aceptacion/candado.txtar (US4.1 a US4.4 y US4.6 · FR-010, FR-011, FR-013 · SC-003, SC-004), cada uno con un comentario de cabecera que nombra los FR y SC que cubre y que falla hoy por una aserción
  - **Lo común** (contracts/aceptacion.md §1): los binarios `$KITLEGAL_V1_BIN` (`v0.1.0`) y `$KITLEGAL_V2_BIN`
    (`v0.2.0`) del arnés de H19; `env HOME=$WORK/<caso>` con su `mkdir` por caso, y con hosts `mkdir $HOME/.claude
    $HOME/.gemini/config` antes de instalar; `env KITLEGAL_CACHE_DIR=$WORK/cache-<caso>`, que no existe hasta que algo
    la crea; la consulta `territorio resolver Leganés --json --no-graph`, y la del exit 4, `boe articulo
    BOE-A-2015-10565 a99 --offline` dentro de `exec sh -c '…; test $? -eq 4'` (research.md D10).
  - **Referencias** al principio de cada guion: cada orden del guion con un `HOME` vacío y sin skills, su salida
    estándar (`cp stdout`), su salida de error (`cp stderr`) y su código (`test $? -eq N`). En cada caso, `cmp stdout`
    con la referencia y el mismo código (SC-003, FR-012).
  - **Las tres líneas**, como regex: FR-006 `aviso: kitlegal v0\.2\.0 ha actualizado las skills globales, que eran de
    v0\.1\.0; se aplican en la próxima conversación`; FR-009 `aviso: kitlegal v0\.2\.0 se actualizarían las skills
    globales, que eran de v0\.1\.0; se aplican en la próxima conversación`; la de hoy `aviso: las skills instaladas son
    de kitlegal v0\.1\.0 y este binario es kitlegal v0\.2\.0; ejecuta: kitlegal skills install -g`. Donde la referencia
    deja vacía la salida de error, `stderr '\A<línea>\n\z'`; donde ya escribe (`--dry-run`, un error sin `--json`),
    `una-linea-mas.sh <ref> <obtenida> <patrón>`, un guion de shell del propio archivo txtar (la obtenida tiene
    exactamente una línea que empieza por `aviso: `, casa con el patrón, y sin ella es byte a byte la referencia).
  - **«Nada escrito»** (FR-013): `arbol $HOME` antes y después comparados con `cmp`, y `! exists $KITLEGAL_CACHE_DIR`.
    **«Reinstala»** (FR-005): `arbol $HOME` igual al de un `HOME` gemelo con la misma preparación sobre el que se
    ejecuta `$KITLEGAL_V2_BIN skills install -g` directo, `! grep 'v0\.1\.0' $HOME/.agents/skills/kitlegal.json` y
    `! exists $KITLEGAL_CACHE_DIR/skills-global.lock`.
  - **reinstala.txtar**, los seis casos de §2: primera y segunda consulta (US1.1-1.3, con `V2 skills doctor -g --json`
    sin hallazgos y la segunda con `! stderr .` y el árbol igual), un fichero de `references/` que falta (US1.4), una
    skill en otra versión con `V2 skills install -g plazos-y-territorio` (US1.5, la línea nombra `v0.1.0`), un nombre
    retirado intacto sin hosts (`mv` de `normativa` a `boe-legislacion` y `sed 's/"normativa/"boe-legislacion/g'` en
    `kitlegal.json`; US1.6), el exit 4 (US1.7, `! stdout .`, código 4, `una-linea-mas.sh` con FR-006) y el multicall
    con `symlink $WORK/bin/territorio -> $KITLEGAL_V2_BIN` (US1.8).
  - **tocadas.txtar**, §3: primero el control (hosts, `V1 install -g` sin tocar, la consulta reinstala con
    `\A<FR-006>\n\z`); después, cada uno en su `HOME`, el fichero editado (US2.1, con `grep '"version": "v0\.1\.0"'`
    en el manifiesto), el fichero ajeno con `olvida.sh` del propio archivo, la entrada de host ajena (`rm` y `symlink`
    a `$WORK/fuera`), la retirada «sin retirar» (US2.2), el binario más viejo (US2.3) y los metadatos de construcción
    con `sed 's/"v0\.2\.0"/"0.2.0+abc"/g'` (US2.4), con las líneas de hoy exactas de §3 y «nada escrito» en todos.
  - **excluidas.txtar**, §4: el mismo control en su `HOME`; después, sobre otro `HOME` con hosts y `V1 install -g`,
    `mcp llamadas.txt $KITLEGAL_V2_BIN mcp serve --no-graph` (US3.3, `grep -count=1 '^aviso: '` y que es la de hoy),
    `--describe` (US3.4, `\A<hoy>\n\z`), el análisis fallido `boe articulo` (código 2, `! stdout .`, `una-linea-mas.sh`
    con la de hoy), la ayuda (`territorio resolver --help`, `territorio --help`, `version`: `! stderr .`) y
    `--dry-run` (US3.5, `una-linea-mas.sh` con FR-009), con «nada escrito» tras cada una; y el ámbito local de §4: el
    local viejo (US3.1, la línea de hoy sin ` -g`), el local al día con el global viejo (US3.2, `! stderr .`) y el
    local inservible con `./.agents` fichero del propio archivo (FR-003 a, `! stderr .`).
  - **candado.txtar**, §5: el candado reciente copiado de `candado.txt` del propio archivo (US4.1, la de hoy, el árbol
    sin cambios y `cmp` del candado), el huérfano con `exec touch -t 202001010000` (US4.2, reinstala y no queda), sin
    directorio de caché (US4.3, lo crea, reinstala y no queda), sin permiso con `exec chmod -R a-w
    $HOME/.agents/skills` (US4.4, la de hoy, el árbol igual tras `exec chmod -R u+w`, sin candado y la consulta
    siguiente reinstala) y dos a la vez con `exec … &a&`, `exec … &b&`, `wait a` y `wait b` (US4.6, una sola línea de
    FR-006 entre las dos, la otra la de hoy o vacía, la salida estándar de cada una la de la referencia y el árbol el
    de una reinstalación; research.md D8). Los permisos se devuelven antes de terminar el caso.
  - Los guiones afirman solo lo que el spec fija. Verificación: `scripts/workflow/aceptacion.sh rojo-primero` los ve
    fallar a los cuatro por una aserción contra el binario de hoy (que solo da la línea de H19); `make ci` en verde.

**Checkpoint**: la suite existe, falla hoy por una aserción y queda congelada.

---

## Phase 2: Fundacional — el dominio, el candado y el permiso

**Propósito**: las tres piezas que T005 compone, cada una con sus unitarios, sin que ninguna invocación cambie todavía
(los e2e de H19 siguen en verde tal como están).

- [X] T002 El dominio que decide y reinstala, con sus unitarios (FR-003, FR-004, FR-005, FR-006, FR-007, FR-009, FR-011, FR-013, FR-014, FR-019; SC-001, SC-002, SC-006; contracts/aviso-y-reinstalacion.md §2 y §3; data-model §1 a §6; research.md D1, D3, D6, D7, D11): internal/core/instalacion/reinstalacion.go (nuevo), internal/core/instalacion/puertos.go, internal/core/instalacion/version.go, internal/core/instalacion/discoEnMemoria_test.go
  - **puertos.go**: `type TomarCandado func() (liberar func(), tomado bool)`, con el comentario de data-model §3 (con
    `tomado == false`, `liberar` es nil y no queda nada escrito).
  - **version.go**: la precedencia SemVer 2.0.0 §11, no exportada, junto a `FormaSemVer` y `MismaVersion` (research.md
    D6): sin la `v` inicial ni los metadatos de construcción; el núcleo número a número; sin prerelease va después de la
    misma con prerelease; prereleases identificador a identificador (numéricos por valor, alfanuméricos por bytes
    ASCII, numérico antes que alfanumérico, más identificadores después); cada número por longitud y, con la misma, por
    bytes. Quien la llama ya comprobó la forma con `FormaSemVer`.
  - **reinstalacion.go**: `PermisoDelAviso` con `SoloAvisar`, `Ensayar` y `Reinstalar` (data-model §1); `Reinstalacion`
    con los campos de data-model §2 (`Disco`, `Enlazador`, `Escritor`, `Candado`, `Home`, `Version`, `Empotradas`,
    `Retiradas`); y `Avisar(permiso) (linea string, hay bool)`, que aplica contracts §2 paso a paso: hoy (`Aviso`), solo
    avisar, ámbito (el manifiesto que compara el aviso es el local → la de hoy), versiones (FR-004 con la precedencia de
    `version.go`), plan (`Planificar` sin `--host` con el pedido de data-model §4 —declaradas y empotradas más
    `Retiradas.Herederas` de las retiradas, en orden de nombre y una vez—; error o «sin retirar» → la de hoy), ensayo
    (la línea de FR-009, sin candado ni escritura), candado (sin tomarlo → la de hoy; tomado, `liberar` diferido pase lo
    que pase), otra vez bajo el candado (pasos 1, 3, 4 y 5 sobre el disco de ahora; sin línea de hoy → ninguna) y
    aplicar (`Aplicar` con el plan del paso 8; error → la de hoy del paso 8; bien → la de FR-006 con la versión que
    nombra la de hoy del paso 8). Las dos líneas nuevas se componen aquí con los bytes de contracts §3. Ningún error
    se devuelve; el `switch` sobre el permiso nombra sus tres valores (`exhaustive`).
  - **Tests** en internal/core/instalacion/reinstalacion_test.go (paquete `instalacion_test`, sobre `discoEnMemoria`
    y con un candado falso que cuenta las llamadas a `liberar`; discoEnMemoria_test.go solo si le falta algo que el
    test necesita): `TestReglaDeVersiones` (FR-004, FR-019), la tabla de research.md D7 con `Avisar(Ensayar)` —menor,
    prerelease anterior a su final y al revés, `1.0.0-alpha` < `1.0.0-alpha.1`, `1.0.0-2` < `1.0.0-11`, `1.0.0-1` <
    `1.0.0-alpha`, `0.9.0` < `0.10.0`, solo metadatos (`0.1.0+abc` y `0.1.0`: la de hoy), binario más viejo, una skill
    declarada más nueva que el binario con el manifiesto más viejo, una versión instalada sin forma— que espera la línea
    de FR-009 o la de hoy y el disco sin cambios; y `TestReinstalacion` (FR-003, FR-005 a FR-007, FR-009, FR-013,
    FR-014) con `Avisar(Reinstalar)` sobre los estados de D7 —sin tocar, un fichero que falta, uno editado, uno ajeno,
    una entrada de host ajena, una retirada intacta y una «sin retirar», un manifiesto local viejo, uno al día con el
    global viejo y un local ilegible con el global viejo—: la línea, el estado del disco (`instantanea`) igual al de
    `Planificar`+`Aplicar` del mismo pedido o sin cambios, `liberar` llamado una vez por candado tomado y el candado sin
    pedir donde no se cumplen FR-003 (a)-(c); más `SoloAvisar` (la de hoy, sin candado) y `Ensayar` sobre el estado sin
    tocar (la de FR-009, sin candado y sin cambios), y el candado que no se toma (la de hoy, sin cambios).
  - Control de umbral (ADR 0029): un mutante que trata `0.1.0+abc` como menor que `0.1.0`, o uno que reinstala con
    un binario más viejo, pone `TestReglaDeVersiones` en rojo. Cobertura de `internal/core/**` ≥ 85 % (DoD §1.9).
    Verificación: `go test -race -count=1 -run 'TestReglaDeVersiones|TestReinstalacion' ./internal/core/instalacion/`
    y `make ci` en verde.

- [X] T003 [P] El candado en la caché, con su test y la superficie del paquete (FR-010, FR-013, FR-016; SC-004; contracts/aviso-y-reinstalacion.md §4; data-model §3; research.md D4, D5, D7): internal/cache/candado.go (nuevo), internal/cache/superficie_test.go
  - **candado.go**: `func TomarCandado(nombre string, caducidad time.Duration) (liberar func(), tomado bool)`: resuelve
    `Directorio()`; `os.MkdirAll` con `permisosDelDirectorio` (0o700); `os.OpenFile` con `O_WRONLY|O_CREATE|O_EXCL` y
    0o600 sobre `filepath.Clean(filepath.Join(directorio, nombre))`, vacío; si ya existe, `os.Lstat` y `time.Since` de
    su fecha de modificación: menos que la caducidad → no tomado y sin tocarlo; la caducidad o más → `os.Remove` y una
    sola creación exclusiva más. Cualquier otro fallo (`Directorio()`, `MkdirAll`, la creación, la retirada) → no
    tomado, `liberar` nil y nada creado que quede. `liberar` borra el fichero; un borrado que falla se deja (research.md
    D11). Sin esperas ni reintentos; solo `os`, `path/filepath`, `time`, `errors` e `io/fs` (R3).
  - **superficie_test.go**: la entrada `"func TomarCandado"` en `superficieDelContrato`, con su comentario y el de
    cabecera nombrando el candado de H38, como H7 hizo con `"func Directorio"` (research.md D4).
  - **Test** `TestTomarCandado` en internal/cache/candado_test.go con `t.Setenv` de la variable de `ruta.go` y sin
    `t.Parallel()`: libre → tomado y el fichero existe con 0o600 y vacío; `liberar` → no existe; reciente (`os.Chtimes`
    a hace 9 min 30 s) → no tomado e intacto (mismo contenido y misma fecha); huérfano a hace 10 min y a hace un año →
    tomado; sin directorio → lo crea con 0o700 y tomado; la variable vacía y la que apunta a un fichero → no tomado y
    nada creado.
  - Control de umbral (ADR 0029): un mutante con `<=` en lugar de `<` en la caducidad, o que trata el reciente como
    huérfano, pone `TestTomarCandado` en rojo; sin la entrada de la lista, `TestSuperficieExportada` sale en rojo.
    Verificación: `go test -race -count=1 -run 'TestTomarCandado|TestSuperficieExportada' ./internal/cache/` y `make ci`
    en verde.

- [X] T004 El permiso de cada invocación, del kernel al avisador, sin cambiar todavía lo que el avisador del binario hace (FR-001, FR-002, FR-008, FR-009, FR-015; contracts/aviso-y-reinstalacion.md §1; data-model §1; research.md D2, D12): internal/app/aviso.go, internal/app/main.go, internal/app/registro.go, internal/app/mcp_test.go
  - **aviso.go**: `type Avisador func(permiso instalacion.PermisoDelAviso) (linea string, hay bool)`, con su comentario;
    `AvisoDeVersion` devuelve un avisador de esa firma que sigue dando, con cualquier permiso, la línea de hoy tal como
    hoy (la composición de `Reinstalacion` llega con T005, junto a los e2e de H19 que cambian).
  - **main.go**: en `resolverApplet`, en el punto en que hoy llama a `avisar` y solo si `buscaElAviso` (sin cambios),
    el permiso de contracts §1: análisis fallido → `SoloAvisar`; `DecisionDescribir` → `SoloAvisar`; argumentos del
    verbo analizado que son un `servidor` (`mcp serve`, con o sin `--dry-run`), reconocido por la interfaz como en
    `ejecutarVerbo` y sin escribir su nombre → `SoloAvisar`; ejecutar con `--dry-run` → `Ensayar`; cualquier otra
    ejecución → `Reinstalar`. `avisar(p, registro, permiso)`; la ayuda, `skills`, `version` y las llamadas de
    herramienta siguen sin avisador (`sinAvisador`).
  - **registro.go**: `Registro.aviso(permiso)`, que lo pasa al avisador; sin avisador, ninguna línea, como hoy.
  - **Tests**: en internal/app/aviso_test.go, `casosDelAvisoDelKernel` gana el permiso esperado de cada caso que avisa
    y el avisador sintético lo registra (FR-001, FR-002, FR-009): `verbo-con-json`, `verbo-en-tabla`,
    `verbo-por-omisión`, `por-nombre-de-invocacion` y `fallo-del-verbo` (exit 4) → `Reinstalar`; `dry-run` y
    `dry-run-que-falla` → `Ensayar`; `describe` y los errores de argumentos (`error-de-argumentos-del-verbo`, su
    variante por omisión, `argumento-de-mas`, `bandera-desconocida`, `plazo-invalido`) → `SoloAvisar`; los que no
    avisan siguen sin llamarlo. `TestAvisoDeVersion` con la firma nueva y las mismas expectativas. En
    internal/app/registro_test.go, el avisador de prueba con la firma nueva y que recibe el permiso que se le pasa. En
    internal/app/mcp_test.go, el avisador de prueba con la firma nueva y que `mcp serve`, con y sin `--dry-run`, lo
    llama una sola vez por arranque y con `SoloAvisar`, y sus llamadas ninguna (FR-015).
  - Control de umbral (ADR 0029): un mutante que da `Reinstalar` a `--describe`, a un análisis fallido o a `mcp serve`,
    o `Reinstalar` en lugar de `Ensayar` con `--dry-run`, pone `TestAvisoDelKernel` o el test de `mcp_test.go` en rojo.
    Verificación: `go test -race -count=1 -run 'TestAvisoDelKernel|TestAvisoDeVersion' ./internal/app/` y los tests de
    registro_test.go y mcp_test.go que cambian; `make ci` en verde con los e2e de H19 y `h21-mcp-protocolo.txtar` sin
    tocar.

**Checkpoint**: el dominio decide, el candado se toma y el kernel sabe qué permite cada invocación; ninguna invocación
cambia todavía.

---

## Phase 3: User Story 1 · El binario nuevo pone al día las skills globales en la primera consulta (P1) 🎯 MVP

**Goal**: la primera consulta con el binario nuevo, sobre un global sin tocar, deja `~/.agents/skills` como `install
-g` y lo dice en una línea; lo tocado (US2), lo excluido y el local (US3) siguen con la línea de hoy.

**Independent Test**: `aceptacion/reinstala.txtar`, `tocadas.txtar` y `excluidas.txtar`, que el workflow activa al
terminar el bucle; y en la tarea, `TestAvisoDeVersion` y los tres e2e de H19 que cambian.

- [X] T005 [datos] [US1] La composición del avisador que reinstala y los seis e2e de H19 que pasan a esperarla, indivisibles y **nada más**: con la composición nueva y los guiones de hoy, `TestEntregaDelHito/h19-skills-aviso*` y `h19-skills-conflictos-dentro` salen en rojo, y con los guiones nuevos y sin la composición, también (FR-001, FR-003, FR-005, FR-006, FR-010, FR-012, FR-013, FR-018; SC-001, SC-005, SC-007; contracts/aviso-y-reinstalacion.md §2 y §4; contracts/aceptacion.md §6; research.md D4, D12; precedente: H19 T014, H7.1 T014, H33 T004): internal/app/aviso.go, internal/app/testdata/script/h19-skills-aviso.txtar, internal/app/testdata/script/h19-skills-aviso-sin-aviso.txtar, internal/app/testdata/script/h19-skills-conflictos-dentro.txtar, specs/009-h19-instalar-sin-clonar/aceptacion/skills-aviso.txtar, specs/009-h19-instalar-sin-clonar/aceptacion/skills-aviso-sin-aviso.txtar, specs/009-h19-instalar-sin-clonar/aceptacion/skills-conflictos-dentro.txtar
  - **aviso.go** (research.md D12): `AvisoDeVersion` compone `instalacion.Reinstalacion` con `disco.Lector{}`,
    `dependencias.Enlazador`, `disco.NuevoEscritor(dependencias.Enlazador)`, `Candado: func() (func(), bool) { return
    cache.TomarCandado(ficheroDelCandado, caducidadDelCandado) }` con las constantes `ficheroDelCandado =
    "skills-global.lock"` y `caducidadDelCandado = 10 * time.Minute`, `HOME` del entorno en cada llamada, la versión, lo
    empotrado (`skillsEmpotradasDe`) y `skillsRetiradas` (`internal/app/retiradas.go`), y llama `Avisar(permiso)`. Se
    quedan la versión sin forma SemVer y lo empotrado ausente o ilegible (ninguna línea, como hoy); sin `Enlazador`, solo
    la línea de hoy (`instalacion.Aviso`), sin `panic` (constitución IV).
  - **Test** `TestAvisoDeVersion` en internal/app/aviso_test.go: cada subtest fija `HOME` y `KITLEGAL_CACHE_DIR` en
    `t.TempDir()`; los casos de hoy, con `SoloAvisar`, dan las mismas líneas; casos nuevos sobre un global de otra
    versión: `Reinstalar` → la línea de FR-006, el manifiesto en la versión del binario y sin candado; `Ensayar` → la de
    FR-009 y nada escrito (ni el directorio de la caché); `SoloAvisar` → la de hoy y nada escrito; un candado de hace
    9 min 30 s (`os.Chtimes`) → la de hoy y el global sin cambios; uno de hace 10 min → la de FR-006; sin `Enlazador` →
    la de hoy y nada escrito. Así la caducidad de la composición queda fijada por un test (FR-010).
  - **Los seis guiones** (contracts/aceptacion.md §6), cada activo y su copia congelada con el mismo cambio: en
    h19-skills-aviso.txtar y skills-aviso.txtar, la línea 119 pasa a `stderr '\A<FR-006>\n\z'` y tras ella `exec
    $KITLEGAL_V1_BIN skills install -g` antes de los casos de 121-132; en h19-skills-aviso-sin-aviso.txtar y
    skills-aviso-sin-aviso.txtar, las líneas 152, 156 y 160 pasan a `stderr '\A<FR-006>\n\z'`, con `exec
    $KITLEGAL_V1_BIN skills install -g` antes de los controles de 155 y 159 y tras el de 160, antes de los casos de
    162-189; en h19-skills-conflictos-dentro.txtar y skills-conflictos-dentro.txtar, la 203 pasa a `stderr
    '\A<FR-006>\n\z'` y tras la 204 `exec $KITLEGAL_V1_BIN skills install -g`, antes del caso (h). Los comentarios de
    esos controles dicen que ahora reinstalan y por qué se devuelve la v1. Ninguna expectativa local cambia; ningún
    otro guion cambia (FR-018). `<FR-006>` es la regex de contracts/aceptacion.md §1.
  - Control de umbral (ADR 0029), SC-005: `TestArquitectura` sin cambios en su código; su subprueba «sin red» cubre
    ahora `internal/cache` por la importación de `internal/app/aviso.go`. Mutante sobre una copia bajo `$TMPDIR`: un
    `import _ "net"` en `internal/app/aviso.go` pone la subprueba en rojo. Y un mutante con `caducidadDelCandado = 0`
    pone `TestAvisoDeVersion` en rojo.
  - Verificación: `go test -race -count=1 -run 'TestAvisoDeVersion|TestAvisoDelKernel' ./internal/app/`; `go test
    -count=1 -run 'TestArquitectura' ./internal/`; `go test -race -count=1 -run
    'TestEntregaDelHito/^(h19-skills-aviso|h19-skills-aviso-sin-aviso|h19-skills-conflictos-dentro|h21-mcp-protocolo)$'
    ./internal/app/` (quickstart §2); `diff` entre cada guion activo y su copia congelada solo da las líneas que ya
    difieren hoy; y `make ci` en verde.

**Checkpoint**: la CLI reinstala el global sin tocar en la primera consulta; los cuatro guiones de `aceptacion/`, una
vez activados, pueden pasar.

---

## Phase 4: User Story 4 · El candado y los fallos no dejan nada roto (P2)

**Goal**: de dos invocaciones a la vez reinstala como mucho una, y un fallo a mitad deja lo que `install -g` completa.

**Independent Test**: `TestReinstalacionConcurrente` y `TestReinstalacionFalloAMitad`; en e2e, `aceptacion/candado.txtar`.

- [X] T006 [P] [US4] Los unitarios de la concurrencia y del fallo a mitad, sobre el dominio de T002 (FR-010, FR-011, FR-019; SC-004; US4.5, US4.6; contracts/aviso-y-reinstalacion.md §2 pasos 7 a 9; research.md D3, D7): internal/core/instalacion/reinstalacion.go, internal/core/instalacion/discoEnMemoria_test.go
  - **Tests** en internal/core/instalacion/reinstalacion_test.go, sin gorrutinas ni esperas: `TestReinstalacionConcurrente`
    con dos `Reinstalacion` sobre el mismo `discoEnMemoria` y un candado falso compartido, y los dos intercalados de
    research.md D7 forzados desde los propios puertos: (1) la primera escritura de A ejecuta B entera, que encuentra el
    candado tomado y da la línea de hoy, y A da la de FR-006; (2) el `TomarCandado` de B, ya evaluado el estado viejo,
    ejecuta antes A entera, que reinstala y suelta; B toma el candado, vuelve a evaluar y no da ninguna línea. En los
    dos, una sola línea de FR-006 y el estado final el de una reinstalación. Y los dos del paso 8 que no se cumple
    (revisión final): (3) el `TomarCandado` de B ejecuta antes, entera, una `Reinstalacion` v0.3.0, y (4) antes de que B
    lo tome se edita un fichero del global; en los dos, B da la línea de hoy de ahora y no escribe nada. `TestReinstalacionFalloAMitad` (patrón de
    `TestFalloAMitadSeCompleta`, `aplicar_test.go:341-393`): para cada operación n de una reinstalación con las cuatro
    fases (una retirada intacta, un fichero que se reescribe, un enlace de host nuevo y el manifiesto), un `Escritor`
    que falla en la n: la línea de hoy del estado anterior, `liberar` llamado, y después `Planificar`+`Aplicar` con el
    pedido de `install -g` sin conflictos y con el estado final de la reinstalación sin fallos; si la n es el enlace,
    `Aplicar` sigue con su copia (FR 024 de H19) y la reinstalación da la línea de FR-006, como `install`.
  - reinstalacion.go cambia solo si uno de estos tests descubre que el código de T002 no cumple contracts §2 (y entonces
    lo mínimo que lo cumple); discoEnMemoria_test.go, solo si le falta el gancho de la escritura que falla.
  - Control de umbral (ADR 0029), la fila de SC-004 con `TestReinstalacionConcurrente`: un mutante sin el candado (el
    paso 7 que siempre toma) o sin la segunda evaluación (el paso 8 que reutiliza el plan del paso 5) pone el test en
    rojo: sin el candado salen dos líneas de FR-006, y sin la segunda evaluación B reescribe el disco de ahora en los
    intercalados (3) y (4); un mutante que no llama a `liberar` tras un fallo pone en rojo
    `TestReinstalacionFalloAMitad`. Verificación: `go test -race -count=1 -run 'TestReinstalacion'
    ./internal/core/instalacion/` y `make ci` en verde.

**Checkpoint**: la concurrencia y el fallo a mitad tienen su control determinista; el e2e de `candado.txtar` lo comprueba
con procesos reales al activarse.

---

## Phase 5: User Story 5 · La documentación dice cómo se actualiza ahora (P3)

**Goal**: quien lee `README.md` y `CHANGELOG.md` sabe qué se pone al día solo y qué sigue pidiendo la orden.

**Independent Test**: lectura del párrafo «Actualizar» de `README.md` y de «Unreleased» > «Cambiado» de
`CHANGELOG.md` (US5.1, US5.2).

- [X] T007 [P] [US5] La documentación y la verificación final de la rama (FR-017; SC-007; US5.1, US5.2; DoD §1.6; plan.md «Diseño» 5): README.md, CHANGELOG.md
  - **README.md**: el párrafo «**Actualizar**» (el de `brew upgrade kitlegal` y `scoop update kitlegal`) dice que las
    skills globales se ponen al día solas en la primera consulta con el programa nuevo, salvo que se hayan tocado; que
    se aplican en la próxima conversación; y que las skills nuevas del producto y las locales siguen pidiendo `kitlegal
    skills install`, con `-g` o sin él. Ningún otro párrafo cambia.
  - **CHANGELOG.md**: en «## [Unreleased]» → «### Cambiado», una viñeta con lo mismo, que nombra H38 y el ADR 0060,
    dice que si algo se ha tocado sigue la línea de siempre, que el servidor MCP y el ámbito local no reinstalan, y
    declara Windows sin probar porque la CI es solo Ubuntu.
  - Verificación final: quickstart §1 (los cuatro `go test` de los unitarios) y la segunda orden de quickstart §2 (los
    e2e de H19 y de H21), todos `ok`; `make ci` en verde. La primera orden de quickstart §2 (`h38-*`) la ejecuta el
    workflow tras activar la suite.

**Checkpoint**: el hito completo; el workflow activa la suite congelada y cierra.

---

## Definition of Done (docs/ROADMAP.md §1) aplicada al hito

| Punto | Cómo se cumple | Tarea |
|---|---|---|
| 1 · `make ci` en verde | al terminar cada tarea | todas |
| 2 · Unitarios offline del código nuevo; fixtures si toca red | `TestReglaDeVersiones`, `TestReinstalacion`, `TestReinstalacionConcurrente`, `TestReinstalacionFalloAMitad`, `TestTomarCandado`, `TestAvisoDelKernel`, `TestAvisoDeVersion`; no toca red: ninguna grabación ni fixture nuevo | T002 a T006 |
| 3 · Sin `net/http`, `os.Exit`, `fmt.Print*` fuera de lo autorizado | `depguard`, `forbidigo` y la subprueba «sin red» de `TestArquitectura` | T002 a T005 |
| 4 · Salidas contra `schemas/` y `make schema-check` | no aplica: ninguna salida de applet ni `--describe` cambia (plan.md «Controles mecánicos») | — |
| 5 · Errores tipados; ningún `panic` en rutas de usuario | ningún código de salida nuevo; ningún error de la reinstalación se propaga (D11); el `Enlazador` nulo no entra en pánico (D12) | T002, T005 |
| 6 · Comportamiento visible: e2e y `CHANGELOG.md` | los cuatro guiones de `aceptacion/` y los seis de FR-018; la viñeta de «Unreleased» | T001, T005, T007 |
| 7 · ADR si cambia una decisión de arquitectura | no hace falta uno nuevo: el ADR 0060 ya existe y su estado («propuesta») no lo cambia el run (spec, «Assumptions») | — |
| 8 · Fuente externa: `docs/SOURCES.md` | no aplica: ninguna fuente (research.md «Datos externos») | — |
| 9 · Cobertura de `internal/core/**` ≥ 85 % | los unitarios del dominio | T002, T006 |
| 10 · Skill: evals | ninguna tarea cambia una skill; las dos referencias que cambia la reparación del cierre las miden las evals 13, 14 y 17 de `normativa`, que ya existen, en la medición siguiente (research.md, «Reparaciones del cierre», R1) | — |
| 11 a 13 · territorial, grafo, plazos | no aplican | — |

## Controles de umbral (plan.md) → tarea que los construye

| Fila | Control | Tarea | Lo ve fallar |
|---|---|---|---|
| SC-001 | `h38-reinstala` | T001 | `rojo-primero`: hoy la primera consulta da la línea de H19, no la de FR-006 |
| SC-002 | `h38-tocadas` | T001 | `rojo-primero`: el control del guion no reinstala hoy |
| SC-003, FR-008, FR-012 | los cuatro guiones | T001 | `rojo-primero` de cada guion |
| SC-004, FR-010 (e2e) | `h38-candado` | T001 | `rojo-primero`: el huérfano y sin directorio no reinstalan hoy |
| SC-004, FR-010 (concurrencia) | `TestReinstalacionConcurrente` | T006 | mutante sin candado: dos líneas de FR-006; sin segunda evaluación: la que llega reescribe el disco de ahora |
| FR-010 (caducidad) | `TestTomarCandado` | T003 | mutante con `<=` o que trata el reciente como huérfano; y en T005 `TestAvisoDeVersion` con `caducidadDelCandado = 0` |
| SC-005, FR-012 | subprueba «sin red» de `TestArquitectura` | T005 | mutante `import _ "net"` en `internal/app/aviso.go` |
| SC-006, FR-013, FR-009 | `h38-excluidas` y `h38-tocadas` | T001 | `rojo-primero`: el control de cada guion y el `--dry-run` no dan hoy sus líneas |
| SC-007, FR-018, FR-019 | `make ci` | todas (T005 cambia los e2e de FR-018; T002 y T006, los unitarios de FR-019) | cualquier rojo de `make ci` |

## Trazabilidad de requisitos

| Requisito | Tareas |
|---|---|
| FR-001, FR-002 | T001, T004, T005 |
| FR-003, FR-004, FR-005, FR-006, FR-007 | T001, T002, T005 |
| FR-008 | T001, T004 |
| FR-009 | T001, T002, T004, T005 |
| FR-010 | T001, T002, T003, T005, T006 |
| FR-011 | T001, T002, T006 |
| FR-012, FR-013 | T001, T002, T003, T005 |
| FR-014 | T001, T002 |
| FR-015 | T001, T004 |
| FR-016 | T003 (el candado fuera de `~/.agents/skills`); ninguna tarea toca `install`, `list`, `doctor` ni el manifiesto |
| FR-017 | T007 |
| FR-018 | T005 |
| FR-019 | T002, T006 |
| SC-001 a SC-007 | ver «Controles de umbral» |

## Dependencies & Execution Order

- **T001** primero (aceptación congelada; sin código de producto).
- **T002** tras T001. **T003** no depende de nada más que T001 ([P] con T002). **T004** tras T002 (usa
  `instalacion.PermisoDelAviso`).
- **T005** tras T002, T003 y T004 (compone las tres piezas).
- **T006** tras T002 ([P] con T003 a T005: solo toca el dominio y su test).
- **T007** cuando se quiera tras T001 ([P]: solo documentación); su verificación final pide las demás hechas, así que
  va la última.

Orden de ejecución en secuencia: T001 → T002 → T003 → T004 → T005 → T006 → T007.

### Parallel Opportunities

```text
Tras T001:  T002 (internal/core/instalacion)  ∥  T003 (internal/cache)
Tras T002:  T004 (internal/app)  ∥  T006 (internal/core/instalacion, solo tests)
En cualquier momento: T007 (README.md, CHANGELOG.md)
```

## Implementation Strategy

- **MVP**: T001 → T005 (US1): con eso la primera consulta reinstala el global sin tocar, y lo tocado y lo excluido
  siguen con la línea de hoy (US2 y US3 salen de las mismas piezas).
- **Incremento**: T006 (US4) fija con unitarios deterministas la concurrencia y el fallo a mitad, que el e2e de
  `candado.txtar` comprueba con procesos reales.
- **Cierre**: T007 (US5) y la verificación final; el workflow activa la suite congelada, ejecuta `make ci` y hace el
  cierre.
