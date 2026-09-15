# Research: H5 · Skill `boe-legislacion` + andamiaje de skills

**Modo**: desatendido. Cada decisión se tomó con el «Criterio de decisión autónoma» de
`.specify/memory/constitution.md` y queda aquí con la alternativa rechazada y el motivo. Toda afirmación sobre una
herramienta o dependencia externa remite a la tabla de verificación (V1-V60), que dice dónde se comprobó en local; lo
que no se puede comprobar sin red o sin la plataforma está en D22 como **supuesto**, no como hecho. Ninguna afirmación se
apoya en la memoria de sesiones anteriores.

## Tabla de verificación

| # | Afirmación | Dónde se comprobó |
|---|---|---|
| V1 | Claude Code instalado: `2.1.270`, binario nativo empaquetado con Bun | `claude --version` → `2.1.270 (Claude Code)`; `which claude` → enlace a `~/.local/share/claude/versions/2.1.270`; `file` → Mach-O arm64 |
| V2 | Banderas de `-p`: `--output-format text\|json\|stream-json`, `--model`, `--allowedTools`/`--disallowedTools`, `--permission-mode` (`acceptEdits`, `auto`, `bypassPermissions`, `manual`, `dontAsk`, `plan`), `--settings <file-or-json>`, `--setting-sources user,project,local`, `--no-session-persistence`, `--disable-slash-commands` («Disable all skills»), `--bare` (omite ganchos) | `claude --help` (2.1.270) |
| V3 | `stream-json` con `-p` exige `--verbose` | binario 2.1.270: «When using --print, --output-format=stream-json requires --verbose.» |
| V4 | `--max-turns` existe, oculto, solo con `--print`; al agotarse, `result` con `subtype` `error_max_turns` | binario 2.1.270 |
| V5 | Skills personales en `$CLAUDE_CONFIG_DIR/skills` o `~/.claude/skills`; el cargador acepta directorios enlazados | binario 2.1.270 (cargador: `if(!D.isDirectory()&&!D.isSymbolicLink())return null`); en local, `~/.claude/skills/skill-creator` es un enlace y la sesión la lista |
| V6 | La herramienta que activa una skill se llama `Skill`, con entrada `{skill, args?}` | binario 2.1.270 (`var mo="Skill"`, esquema de entrada); transcript JSONL local: `"name":"Skill","input":{"skill":"speckit-plan",…}` |
| V7 | Mensaje `system`/`init` con `model`, `claude_code_version`, `skills[]` | binario 2.1.270, esquemas del formato `stream-json` |
| V8 | Mensaje `result` con `subtype: "success"` y `result: string` | binario 2.1.270 |
| V9 | El resultado de la herramienta Bash no lleva código de salida; en fallo, el texto empieza por `Exit code N` con `is_error: true` | binario 2.1.270 (esquema de salida de Bash; clase `ShellError`); transcript local con `"content":"Exit code 1\n…","is_error":true` |
| V10 | El cliente de la API del modelo usa `https_proxy`/`HTTPS_PROXY`/`http_proxy`/`HTTP_PROXY` salvo que el host case con `NO_PROXY` | binario 2.1.270 (`fetchOptions:$s({forAnthropicAPI:!0})`, `ry()`, `Uw`) |
| V11 | Sandbox de Bash en Linux: `bwrap` con `--unshare-net`, `--unshare-user`, `--unshare-pid`, `--ro-bind / /`; escribible solo el directorio de trabajo, temporales de Claude y `sandbox.filesystem.allowWrite`; proxy inyectado en `http://localhost:3128` que responde `403` `X-Proxy-Error: blocked-by-allowlist`; exige `bwrap` y `socat`; `failIfUnavailable` termina con 1 | binario 2.1.270 |
| V12 | `CLAUDE_CODE_DISABLE_NONESSENTIAL_TRAFFIC` desactiva tráfico no esencial. `CLAUDE_CODE_SUBPROCESS_ENV_SCRUB` activo (`1`), que hasta T036 se leía solo como «retira del entorno de los subprocesos las variables con credenciales», hace además en Linux tres cosas, comprobadas en V61: (1) **exige `bwrap`** al arrancar, antes del mensaje `system`/`init` («bubblewrap is required for subprocess env scrubbing and isolation…», función `nBn` del binario de x86_64 en la ejecución 34930222593 de T030, `nHn` en el de arm64), y crea ficheros y directorios vacíos para montarlos de solo lectura (`.gitconfig`, `.npmrc`, `.bashrc`… en el directorio personal; `package.json`, `.env*`, `.gitmodules`, `node_modules/.bin`, `.claude/commands`… en el de trabajo; `.gitmodules`, `.github` y ficheros de `.git` en `GITHUB_WORKSPACE`); (2) **fuerza el modo de permisos `default`** (funciones `Ekn` y `cNn` del binario de macOS, `NLn` en el de Linux arm64): ignora `--permission-mode bypassPermissions`, con el aviso «⚠ Permission mode forced to default — CLAUDE_CODE_SUBPROCESS_ENV_SCRUB is set (allowed_non_write_users hardening). Declare allowedTools explicitly, or set CLAUDE_CODE_SUBPROCESS_ENV_SCRUB=0 to opt out.»; (3) **ejecuta cada orden de Bash dentro de `bwrap`** aunque `--settings` desactive el sandbox (con la variable activa y el sandbox desactivado, `dIe` devuelve si hay `bwrap`), lo que exige también `socat` («Sandbox is required but failed to initialize: Sandbox dependencies not available: socat not installed», `MVt` y `Ktt`), con `--new-session --die-with-parent --ro-bind / /`, escritura solo en `/home`, `/root`, `/tmp`, `/var`, `/opt`, `/run` y `/mnt` salvo una lista de ficheros de configuración de solo lectura (`oHn`), y `--dev /dev --unshare-pid --unshare-user --cap-drop ALL --proc /proc` (`DVt`). El gancho `SessionStart` no se ejecuta dentro de `bwrap`. Con la variable a `0` no ocurre nada de eso, y Claude Code sigue sin pasar `CLAUDE_CODE_OAUTH_TOKEN` al entorno de las órdenes ni al de los ganchos (V61 (4)) | binario 2.1.270: el de macOS de la máquina de desarrollo y el de Linux arm64 del paquete de npm, leídos por desplazamiento de byte; V61 |
| V13 | Los ganchos `PostToolUse` reciben `tool_input` y `tool_response`, sin código de salida | binario 2.1.270 (esquemas de ganchos) |
| V14 | `gh workflow run --ref`: «Branch or tag name which contains the version of the workflow file you'd like to run»; la ayuda no dice que el fichero tenga que estar en la rama principal | `gh workflow run --help` (gh 2.100.0) |
| V15 | Validador del estándar Agent Skills: claves admitidas `name`, `description`, `license`, `allowed-tools`, `metadata`, `compatibility`; `name` `^[a-z0-9-]+$` sin guion al principio, al final ni doble, ≤ 64; `description` ≤ 1024 y sin `<` ni `>`; `compatibility` ≤ 500 | `~/.agents/skills/skill-creator/scripts/quick_validate.py` 42-92 (copia idéntica en el plugin `skill-creator`) |
| V16 | `go install` deja el ejecutable en `GOBIN`, o `$GOPATH/bin`, o `$HOME/go/bin` | `go help install` |
| V17 | `go list` expone `Target` («install path»): `go list -f '{{.Target}}' ./cmd/kitlegal` → `/Users/jorge/go/bin/kitlegal` | `go help list` línea 25; ejecución local |
| V18 | `GOENV=off` desactiva el fichero de configuración de Go | `go help environment` líneas 40-44 |
| V19 | `http.DefaultTransport` usa `ProxyFromEnvironment` (`HTTP_PROXY`, `HTTPS_PROXY`, `NO_PROXY`, también en minúsculas) | `go doc net/http ProxyFromEnvironment`; `go doc net/http DefaultTransport` |
| V20 | `internal/httpx` clona `http.DefaultTransport`, «proxy del entorno» incluido | `internal/httpx/transporte.go` 5-18 |
| V21 | `httpx.Replay` no abre conexiones, no pide `robots.txt` (su grabación queda sin usar) y no reintenta; `httpx.New` hace 3 intentos por omisión | `internal/httpx/cliente.go` 51, 298-319 |
| V22 | Vigencias de la fuente: 300 s (`buscar`, `metadatos`) y 604 800 s (`indice`, `articulo`, `articulos`) | `internal/source/boe/fuente.go` 25-31, 215-225 |
| V23 | En la caché, lo caducado es ausencia; en solo lectura (`--offline`), la ausencia es clase «fuente no disponible» (4) | `internal/cache/entradas.go` 53-76 |
| V24 | `--describe` exige los argumentos del verbo; el documento lleva `title`, `description`, `properties.entrada` (argumentos y las ocho banderas) y `properties.salida` con `if`/`then`/`else` y `$defs` | `go run ./cmd/kitlegal boe articulo --describe` → código 2 («expected "<norma> <bloque>"»); con `BOE-A-2015-10565 a21 --describe` → el documento |
| V25 | `describir` no está exportado; `cli.Describir(p, cli.Verbo)` y `cli.Verbo` sí; `render.Nuevo(salida, errores io.Writer)` | `internal/app/main.go` 463-479; `internal/cli/describe.go` 61-99; `internal/render/render.go` 44 |
| V26 | `app.Main(argv, registro, stdout, stderr, …) int`; valor cero de `app.Registro` listo para usar; `app.AppletBoe(app.DependenciasDeBoe{Cliente, Cache})`; `cache.ConDirectorio`, `cache.ConReloj`; `httpx.ConFuente`, `httpx.ConRaizDeGrabacion`, `httpx.ConIntervalo`; `boe.NombreDeLaFuente`, `boe.IntervaloEntrePeticiones` | `internal/app/main.go` 78-127; `internal/app/registro.go` 42-60, 164-172; `internal/app/boe.go` 25-60; `internal/cache/cliente.go` 50, 78; `internal/httpx/cliente.go` 105, 135, 157 |
| V27 | Gramáticas: norma `^BOE-A-[0-9]{4}-[0-9]{1,9}$`; bloque `^[A-Za-z0-9][A-Za-z0-9.-]{0,63}$` (incluye `a85bis.`) | `internal/source/boe/ids.go` 14-95 |
| V28 | Todos los argumentos de los seis verbos de `boe` van por posición (`arg:""`) | `internal/app/boe.go` 138-190 |
| V29 | Grabaciones de H4: metadatos de `BOE-A-2017-12902` con título «Ley 9/2017…» y de `BOE-A-1985-5392` con «Ley 7/1985…»; en el índice de la LCSP, `a1-30` es «Artículo 118»; la búsqueda grabada devuelve los rangos `Ley`, `Real Decreto-ley` y `Real Decreto` | `internal/source/boe/testdata/boe.legislacion-consolidada/` (metadatos e índice); `internal/source/boe/testdata/golden/buscar-procedimiento-administrativo-comun.json` |
| V30 | `refs/boe.py` documenta `indice BOE-A-2006-20764` y `articulo BOE-A-2006-20764 a17` | `refs/boe.py` 8-12, 760-768 |
| V31 | `go.yaml.in/yaml/v3` lo mantiene la organización YAML desde que `go-yaml` se declaró sin mantenimiento (abril de 2025); es el mismo código que `gopkg.in/yaml.v3` (misma comprobación de claves repetidas en `decode.go:776` de los dos, que solo corre en las condiciones de V43); ya está en el grafo del módulo por `testify` (`go.mod` 27, `v3.0.5 // indirect`) y el binario no lo enlaza. La organización lo declara legado congelado, que solo recibe correcciones de seguridad: todo el trabajo en curso, correcciones incluidas, se hace en v4, y a los proyectos nuevos les pide `go.yaml.in/yaml/v4`. Eso lo dice el README de v4 (V58 (2)), no el de v3.0.5 | `go.yaml.in/yaml/v3@v3.0.5/README.md` («Project Status», sin sección «Version Intentions»); `go.yaml.in/yaml/v4@v4.0.0-rc.2/README.md` 40-49; `decode.go` 774-777 de `go.yaml.in/yaml/v3@v3.0.5` y de `gopkg.in/yaml.v3@v3.0.1`; `go mod why -m go.yaml.in/yaml/v3`; `internal/arch_test.go`, `modulosDelBinario`, que no lo lista |
| V32 | `santhosh-tekuri/jsonschema/v6`: `NewCompiler`, `AssertFormat`, `AddResource`, `Compile`, `(*Schema).Validate(any)`; vocabularios propios, con un ejemplo `uniqueKeys` | `compiler.go` 53, 121, 178; `validator.go` 15; `example_vocab_uniquekeys_test.go`; uso existente en `internal/app/esquema_test.go` 103 |
| V33 | `.gitattributes`: los patrones siguen las reglas de `.gitignore` salvo que no admiten negación; un `.agents/.gitattributes` con `skills/** linguist-vendored linguist-generated` asigna los dos atributos a `.agents/skills/**` y a nada más | página man local `gitattributes.5` 81-91; experimento en un repositorio desechable en `/tmp` con `git check-attr` (git 2.50.1): `set` bajo `.agents/skills/`, `unspecified` en `.agents/otro.md`, `skills/boe/SKILL.md` y el propio `.agents/.gitattributes` |
| V34 | Un enlace `scripts/boe -> ../../../bin/instalado/kitlegal` dentro de `repo/skills/s`, invocado a través de `home/.claude/skills/s -> repo/skills/s`, resuelve al binario enlazado desde `repo/bin/instalado/kitlegal`, y el programa ve como nombre de invocación la ruta con `…/scripts/boe` | experimento desechable en `/tmp` (darwin): `argv0=/tmp/…/home/.claude/skills/s/scripts/boe` |
| V35 | El extractor de rutas del workflow reconoce `(\.?[A-Za-z0-9_-]+/)+…`, ficheros con extensión conocida, `.*ignore`, `.editorconfig`, `Makefile` y `LICENSE`: `.gitattributes` en la raíz no es declarable y `.agents/.gitattributes` sí | `.specify/workflows/hito/workflow.yml` 615 |
| V36 | El guardián marca como protegido todo `testdata/*`, `*/testdata/*` y `schemas/*` (una tarea sin `[datos]` no puede tocarlo). `clasificar_datos` pausa una tarea `[datos]` solo si **modifica** un fichero ya existente bajo `testdata/` (a cualquier profundidad) o `schemas/`, o **añade** un fichero bajo `testdata/` de la raíz, bajo `internal/source/` o bajo `schemas/`. Un fichero nuevo bajo `internal/<paquete>/testdata/` fuera de `internal/source/` no pausa: `internal/evals/testdata/grabaciones.json` daría `pausa:false`, y `testdata/evals/grabaciones.json`, `pausa:true` | `.specify/workflows/hito/workflow.yml` 702 y 713 (guardián; 816 en el de reparación), 931-935 (`clasificar_datos`) |
| V37 | El despacho multicall toma el último componente de `argv[0]` | `internal/app/despacho.go` 218-225 |
| V38 | `run.build-tags` actuales: `integration`, `fuentes`, `grabacion`; excepciones de `forbidigo` solo en `cmd/` y en el binario de e2e | `.golangci.yml` 28-31, 306-321 |
| V39 | Reglas de las skills `golang-*` citadas en D21 | `.agents/skills/golang-testing/SKILL.md` 49-64 y 360-387; `.agents/skills/golang-project-layout/SKILL.md` 78-82; `.agents/skills/golang-continuous-integration/SKILL.md` 47-50 y 264-277; `.agents/skills/golang-lint/SKILL.md` 91-93 |
| V40 | GNU make termina con código 2 cuando falla una receta | `rtk proxy sh -c 'printf "x:\n\tfalse\n" \| make -f - x; echo "código $?"'` → `código 2` |
| V41 | Con `HTTP(S)_PROXY=http://127.0.0.1:9` y caché vacía, `kitlegal boe metadatos BOE-A-2015-10565 --json` termina con **5** (`limite-o-tos`: «no se ha podido obtener el robots.txt del sitio») en unos 3 s, sin respuesta de la fuente; con `--offline`, **4** sin pedir nada | binario construido del árbol en un temporal, `KITLEGAL_CACHE_DIR` temporal; ejecución local |
| V42 | `golangci-lint` v2.13.2 y `misspell` v0.8.0 (los de `tools/golangci-lint/go.mod`) montan el replacer con `DictMain`, le añaden `DictAmerican` por `locale: US`, le retiran las `ignore-rules` con `RemoveRule` y, sin `mode: restricted`, pasan `Replace` por el **fichero entero**, no solo por los comentarios. `misspell` solo marca una palabra **suelta** —delimitada por todo lo que no es `[a-zA-Z0-9']`, así que `-`, `_`, `.`, `/` y las letras con tilde la parten— escrita toda en minúsculas, toda en mayúsculas o solo con la inicial en mayúscula, y cuya forma en minúsculas es una errata del diccionario; dentro de un identificador camelCase no marca nada. El barrido cubre también spec.md, porque los comentarios que citan un FR copian su redacción. Con las catorce `ignore-rules` actuales, sobre el texto final de spec.md, plan.md, research.md, data-model.md, quickstart.md y contracts/, marca **exactamente** estas veintiséis: `activacion` (words.go:10394), `candidatas` (11092, «candidates»), `comando` (25278), `comandos` (21247), `componentes` (6363), `constitucion` (3124), `contradice` (11710), `decisiones` (11920), `declaracion` (6735), `defectos` (21506), `definitivo` (11977), `directorios` (6948), `distribuye` (12294), `evaluacion` (12580), `informativo` (7835), `legislativo` (8118), `patrones` (23220), `preparacion` (8929), `prescripcion` (4787), `procede` (26582), `producto` (23463), `programas` (19409), `recorre` (26678, «recorder»), `secretos` (23874), `terminaron` (16190, «terminator») y `variantes` (20300). Solo sobre spec.md marca doce, todas entre ellas: `candidatas` (FR-005 y el caso límite «Varias normas candidatas»), `comando`, `comandos`, `decisiones`, `defectos`, `directorios`, `distribuye`, `informativo`, `legislativo`, `procede`, `producto` y `terminaron` (FR-071 y *Key Entities*: «las invocaciones que terminaron sin texto»). Ninguna otra: tampoco las palabras de los nombres de subtest del inventario de plan.md. Entre las que no marca: `sesion`, `terminada`, `terminar`, `acabaron`, `candidata`, `posibles`, `activada`, `activa`, `repetida`, `clave`, `codigo`, `motivos`, `ilegible`, `instalacion`, `sincronia`, `grabaciones`, `invocaciones`, `trazas`, `citas`, `informe`, `normas`, `esquemas`, `enlaces`, `referencias`, `bloque`, `busqueda`, `senal`, `linea`, `leido`, `conexion`, `peticion`, `recorrer` y `recorrido`. La decisión sobre cuáles entran en `ignore-rules` está en D20 | `github.com/golangci/golangci-lint/v2@v2.13.2`: `pkg/golinters/misspell/misspell.go` 44-73 (`createMisspellReplacer`: `DictMain`, `AddRuleList(DictAmerican)`, `RemoveRule`, `Compile`) y 89-97 (`Replace` salvo `restricted`); `github.com/golangci/misspell@v0.8.0`: `replace.go` 19 (`wordRegexp`) y 218-258 (`recheckLine`: descarta `CaseUnknown` y exige la corrección del diccionario), `case.go` 19-44, `words.go`. Lista calculada con una réplica de ese replacer (programa desechable fuera del repositorio, sobre esos dos módulos de la caché local: `DictMain`, `DictAmerican`, `RemoveRule` de las catorce y `Replace`) pasada por spec.md, por esos ficheros y por los nombres de subtest propuestos, y vuelta a pasar sobre el texto final después de la última corrección del plan; las alternativas de D20 (`recorrer`, `recorrido`, «sin terminar», «que no acabaron», `candidata`, «posibles»), pasadas por la misma réplica, no dan ninguna. `terminaron` la encontró la comprobación del juez del plan en los artefactos (`gates/plan-r6.json`, criterio i); `candidatas`, el barrido de spec.md |
| V43 | Con `go.yaml.in/yaml/v3`, `yaml.Unmarshal(doc, &nodo)` con `nodo` de tipo `yaml.Node` **no** comprueba claves repetidas: `unmarshal` copia el nodo y vuelve antes de mirar su contenido, y la comprobación de repetidos (`uniqueKeys`, activa por omisión) solo corre en `mapping`, es decir, al decodificar a un mapa, a `any` o a una estructura. `(*yaml.Node).Decode(&v)` crea un decodificador nuevo con `uniqueKeys: true` y devuelve `*yaml.TypeError` con «line N: mapping key "…" already defined at line M». Al decodificar a `any`, un mapa con todas las claves de texto da `map[string]any`, y si no, `map[any]any` | `go.yaml.in/yaml/v3@v3.0.5`: `decode.go` 344 (`uniqueKeys: true`), 484-494 (copia del nodo), 767-776 (comprobación), 791 y 864 (`isStringMap`); `yaml.go` 140-151 (`Decode`); confirmado con un experimento del juez del plan (`gates/plan-r1.json`, criterio i) |
| V44 | `jsonschema.UnmarshalJSON(r io.Reader) (any, error)` decodifica JSON a `any` con `json.Number` («without losing number precision»); es la función con la que la propia biblioteca lee sus documentos JSON y la que ya usa el repositorio para validar salidas | `github.com/santhosh-tekuri/jsonschema/v6@v6.0.3`: `loader.go` 38, 124 y 253-256, `content.go` 48; `internal/app/esquema_test.go` 96 |
| V45 | `gh` 2.100.0: en `gh api`, `{owner}` y `{repo}` se sustituyen por los del repositorio del directorio actual, `--paginate` pide todas las páginas y `--jq` filtra la respuesta; `gh run list` y `gh run view` exponen en `--json` `databaseId`, `createdAt`, `event`, `headSha`, `status`, `conclusion`, `url` y `workflowName` (entre otros: `attempt`, `displayTitle`, `headBranch`, `name`, `number`, `startedAt`, `updatedAt`, `workflowDatabaseId`); `gh run list` filtra por `--branch` y `--event`, admite `--limit`, y su `--workflow` filtra por un flujo que `gh` tiene que resolver por nombre, sin que la ayuda diga qué pasa con uno que no está en la rama principal (por eso quickstart §12 no lo usa y filtra por `workflowName`); `gh run view <run-id> --log` imprime el registro completo de la ejecución; `gh pr edit` tiene `--add-label` y `--remove-label`; `gh run watch <run-id> --exit-status` sigue la ejecución hasta que termina y sale con código distinto de 0 si falla. `--jq` recibe una consulta con la sintaxis de jq y no necesita `jq` instalado; las dos consultas de quickstart §12 que eligen la ejecución dan, sobre una lista con ejecuciones de `ci` y de `evals` anteriores y posteriores a la etiqueta en desorden, la primera de `evals` creada desde la etiqueta (y `posteriores_a_la_etiqueta` con todas las posteriores) y su `databaseId`; sin ninguna de `evals` posterior, el mensaje «todavía no hay ninguna ejecución de evals posterior a la etiqueta» y un `databaseId` vacío | `gh api --help`, `gh run list --help` («JSON FIELDS»), `gh run view --help` («JSON FIELDS», `--log`), `gh pr edit --help`, `gh run watch --help`, `gh help formatting` (`--jq`); las dos consultas, con `jq` local sobre una lista sintética |
| V46 | Los tests leen ficheros de la raíz con rutas relativas al directorio del paquete (`carpetaDeLosEsquemas = "../../schemas"`); la grabación de H4 declara como raíz de grabación la carpeta `testdata` de su paquete, y `httpx` exige que la raíz exista y sea un directorio y crea bajo ella `<fuente>/` | `internal/app/esquemas_test.go` 32-33; `internal/source/boe/grabacion_test.go` 24-26 y 51-55; `internal/httpx/grabar.go` 371-384 |
| V47 | El mensaje `result` de `stream-json` lleva `subtype` (`success`, `error_max_turns`, `error_during_execution`, `error_max_budget_usd`, `error_max_structured_output_retries`) y `is_error` booleano, además de `result` | binario 2.1.270: `subtype:"success",is_error:!1,num_turns:0,result:""`, esquema del mensaje con `is_error:I()`, y los cuatro `subtype:"error_…"` |
| V48 | La extracción del informe de quickstart §12.2 y §12.3 —con `set -e`, comprobar primero con `grep -qF --` que el registro contiene las cuatro marcas del contrato del job §3.3 y después imprimir cada fichero con `sed -n "/--- inicio de <f> ---/,/--- fin de <f> ---/p"`— imprime todas las líneas de cada marca de inicio a su marca de fin, las dos incluidas y sin tope, aunque cada línea lleve delante un prefijo; si falta cualquiera de las cuatro marcas, termina con 1 sin imprimir nada | `rtk proxy sh -c` local sobre registros sintéticos con un prefijo de tarea, paso e instante separados por tabuladores: completo, con líneas antes y después → los dos ficheros con sus marcas y código 0; sin `--- fin de informe.json ---` → nada y código 1; sin marcas → nada y código 1; con un `informe.md` de mil líneas → las mil, sus dos marcas y las tres líneas de `informe.json`, seguidas del `código 0` de la sonda (1006 líneas) |
| V49 | La comprobación de ficheros cambiados de quickstart §12.3 escribe la lista entera de `git diff --name-only <commit> HEAD` y la recorre con `while IFS= read -r`: termina con 1 en el primer fichero que no empieza por `specs/006-h5-skill-boe-legislacion/`; con la lista vacía (commit igual a la cabeza) escribe `todos bajo …` y termina con 0; con un commit que no existe, `git diff` falla y `set -e` detiene la orden con su código antes de escribir la lista | `rtk proxy sh -c` local sobre el historial de este repositorio: con `HEAD` → lista vacía, `todos bajo specs/006-h5-skill-boe-legislacion/` y código 0; con `HEAD~2`, anterior a H4 → la lista y `fuera del directorio del hito: .github/workflows/nightly.yml`, código 1; con `0000000000000000000000000000000000000000` → `fatal: bad object …` y código 128 |
| V50 | En bash, `exec 3<>/dev/tcp/127.0.0.1/9` contra un puerto cerrado falla con «Connection refused», y con `set -euo pipefail` esa orden **sola** termina el guion con código 1. En una subshell en la condición de un `if` (`if (exec 3<>/dev/tcp/127.0.0.1/9) 2>/dev/null; then …; fi`), el mismo fallo solo deja la condición en falso y el guion sigue; contra un puerto abierto, la condición es cierta y se ejecuta el cuerpo del `if` | sonda local con `set -euo pipefail` en `/bin/bash` 3.2.57 (macOS) y en bash 5.2.21 de la imagen `ubuntu:24.04`, la del runner, en un contenedor local sin red (`docker run --network none`): la orden sola contra `127.0.0.1:9` → `connect: Connection refused` y código 1, sin llegar a la orden siguiente; la forma del `if` → la orden siguiente al `if` se ejecuta y código 0. Con bash 3.2 y un puerto local abierto (una escucha de `perl` en `127.0.0.1:45679`), la forma del `if` → el mensaje de la comprobación 4 del contrato del job §3.1 y código 1 |
| V51 | El runtime de Go 1.27.1 crea cada hilo del sistema con la llamada `clone`, no con `clone3`: `newosproc` llama a `clone(cloneFlags, …)`, con `cloneFlags` = `CLONE_VM`, `CLONE_FS`, `CLONE_FILES`, `CLONE_SIGHAND`, `CLONE_SYSVSEM` y `CLONE_THREAD`, y en amd64 `runtime·clone` añade `CLONE_SETTLS` antes de la llamada `SYS_clone`. Los hilos los crea `newm`, al que llaman `sysmon`, `templateThread` y `startm`, y `startm` lo ejecuta cualquier hilo del proceso: un hilo puede crearse desde otro que no es el principal | `runtime/os_linux.go` 156-161 y 184, `runtime/sys_linux_amd64.s` 580, 602 (`ORQ $0x00080000, DI`, «add flag CLONE_SETTLS») y 604 (`MOVL $SYS_clone, AX`), `runtime/proc.go` 179, 2875, 2962 y 3111, del toolchain `go1.27.1` (`go env GOROOT`); en arm64, observado en la traza de V53 |
| V52 | GNU `timeout` (coreutils 9.4, la versión de Ubuntu 24.04) termina con 124 si la orden agota el tope, con 137 si tiene que enviarle `KILL` (con `--kill-after`, cuando la orden sigue viva tras la señal inicial) y con el código de la orden si esta termina antes | documentación local: `timeout --help`, apartado «Exit status» (`124 if COMMAND times out, and --preserve-status is not specified`; `137 if COMMAND (or timeout itself) is sent the KILL (9) signal (128+9)`; `the exit status of COMMAND otherwise`), en un contenedor local de `ubuntu:24.04` sin red (`docker run --network none`); sondas en ese contenedor: `timeout 1s tail -f /dev/null` → 124; `timeout --kill-after=1s 1s sh -c 'trap "" TERM; tail -f /dev/null'` → 137; `timeout 5s sh -c 'exit 3'` → 3; `timeout 5s true` → 0 |
| V53 | `strace` 6.8, la versión de Ubuntu 24.04. (1) **Código**: termina con el de la orden que sigue (3 y 0) y, si la orden muere por una señal, se mata con la misma (137 con `KILL`); bajo `timeout --kill-after=10s 2s`, `strace -ff … tail -f /dev/null` da 124 y no queda vivo ningún `tail`; bajo `timeout --kill-after=10s 240s`, `strace -ff … sh -c 'exit 7'` da 7. (2) **Traza**: con `-ff -e trace=execve,connect,clone,clone3,fork,vfork -e signal=none -s 4096 -o <dir>/t`, las opciones del contrato del job §3.2 hasta que V54 mostró que `-e signal=none` suprime la línea final de todo proceso que muere por una señal (el contrato ya no la usa, y V54 (2) repite esta orden sin ella), sobre `bash -c` que invoca dos veces un `kitlegal` de Go 1.27.1 a través de un enlace `boe` con `HTTP(S)_PROXY=http://127.0.0.1:9` y caché vacía, hay un fichero por hilo; ninguna línea `<unfinished …>` ni `<… resumed>`; cada fichero, también el de cada hilo, termina en `+++ exited with N +++`; exactamente un fichero, el de la orden que arrancó `strace`, no lo crea ninguna línea de otro; `bash` crea cada proceso con `clone(child_stack=NULL, flags=CLONE_CHILD_CLEARTID\|CLONE_CHILD_SETTID\|SIGCHLD, child_tidptr=0x…) = N`, y el binario sus hilos con `clone(child_stack=0x…, flags=CLONE_VM\|CLONE_FS\|CLONE_FILES\|CLONE_SIGHAND\|CLONE_THREAD\|CLONE_SYSVSEM) = N`, también desde hilos que no son el principal (`t.43` crea `t.44` y `t.45` crea `t.46`, hilos de la invocación `t.41`); `execve("/tmp/s/bin/boe", ["/tmp/s/bin/boe", "articulo", "BOE-A-2015-10565", "a9998", "--json"], 0x… /* 12 vars */) = 0`, con los bytes no ASCII del argv en octal (`c\303\263digo` en el de `bash`); `connect(9, {sa_family=AF_INET, sin_port=htons(9), sin_addr=inet_addr("127.0.0.1")}, 16) = -1 EINPROGRESS (Operation now in progress)` y, de `bash`, `connect(3, {sa_family=AF_UNIX, sun_path="/var/run/nscd/socket"}, 110) = -1 ENOENT (No such file or directory)`; con `HTTP(S)_PROXY=http://[::1]:9`, `connect(9, {sa_family=AF_INET6, sin6_port=htons(9), sin6_flowinfo=htonl(0), inet_pton(AF_INET6, "::1", &sin6_addr), sin6_scope_id=0}, 28) = -1 EINPROGRESS (Operation now in progress)`. La invocación sin `--offline` termina con 5 tras tres `connect` a `127.0.0.1:9`, los tres en su hilo principal; la de `--offline`, con 4 y sin ningún `connect`; un proceso de `bash` que no ejecuta nada (una sustitución de orden) deja un fichero con solo su línea final | contenedor local desechable de `ubuntu:24.04` (aarch64) con `strace` instalado del archivo de Ubuntu; las sondas, con `docker run --network none --cap-add SYS_PTRACE`; el binario, compilado del árbol en un temporal con `GOOS=linux GOARCH=arm64 CGO_ENABLED=0 go build -trimpath`. Lo que puede diferir en el runner (x86_64, Claude Code) es el supuesto S4 |
| V54 | `strace` 6.8 y `timeout` de coreutils 9.4, con `-ff -e trace=execve,connect,clone,clone3,fork,vfork -s 4096`, con y sin `-e signal=none`. (1) **Con `-e signal=none`** no se escribe `+++ killed by SIG… +++`: el fichero de un proceso que muere por TERM o por KILL acaba en su última llamada, sin línea final, aunque no haya tope (`bash -c 'sh -c "kill -TERM \$\$"; true'`, y lo mismo con KILL: el de `sh` acaba en su `execve`, y `strace` sale con 0). Bajo `timeout --kill-after=10s <tope>`, el corte deja así sin línea final todos los procesos vivos: con `bash -c 'sleep 30 & tail -f /dev/null; true'` y 2 s, 124 y los tres ficheros sin ella; con un programa de Go dentro de un `connect` bloqueante y 3 s, 124, cinco ficheros sin ella, tres de ellos vacíos (hilos sin ninguna llamada trazada) y la llamada interrumpida como última línea, con `= ? ERESTARTSYS (To be restarted if SA_RESTART is set)`; con `bash -c '<boe> articulo BOE-A-2015-10565 a9998 --json …; sleep 30'`, el proxy que rechaza y 8 s, 124, sin línea final solo el fichero de `bash`, vivo en el corte, y con `+++ exited with 5 +++` los siete de la invocación, que terminó antes, con sus tres `connect`, uno de ellos en un hilo creado por otro hilo que no es el principal. Es lo que observó el juez del plan (`gates/plan-r6.json`, criterio i). (2) **Sin `-e signal=none`**, esos mismos procesos dejan `+++ killed by SIGTERM +++` o `+++ killed by SIGKILL +++`, y cada señal entregada, una línea `--- SIGNOMBRE {…} ---`: `--- SIGTERM {si_signo=SIGTERM, si_code=SI_USER, si_pid=N, si_uid=N} ---` (también con `si_code=SI_TKILL`), `--- SIGURG {si_signo=SIGURG, si_code=SI_TKILL, si_pid=N, si_uid=N} ---` del runtime de Go, `--- SIGCHLD {si_signo=SIGCHLD, si_code=CLD_EXITED, si_pid=N, si_uid=N, si_status=N, si_utime=N, si_stime=N} ---` de `bash`, a veces con un comentario `/* 0.01 s */` antes de cerrar la llave, y `--- SIGCONT {…} ---` del mismo remitente que TERM; KILL no deja línea de señal. Con la orden de V53: código 0, 13 ficheros, todos con línea final, uno solo sin línea de creación, ninguna línea `<unfinished …>`, `<… resumed>` ni `<detached …>`, y toda línea es de una forma de V53, de señal (7 `SIGURG` y 2 `SIGCHLD`) o final, con los códigos 5 y 4. Con los tres cortes de (1): 124 y **todos** los ficheros con su línea final; la llamada bloqueante interrumpida conserva `= ? ERESTARTSYS (…)`, seguida de sus líneas `--- SIGTERM {…} ---` y de `+++ killed by SIGTERM +++`, y el fichero de cada hilo sin llamadas trazadas tiene solo esa línea final. (3) **Una señal solo a `strace`**: con TERM, `strace` sigue hasta que termina la orden y sale con 0, con todas las líneas finales; con KILL sale con 137, y los ficheros de los procesos vivos quedan sin línea final, con sus líneas completas. Es el caso del 137: pasados los 10 s de `--kill-after`, `timeout` envía `KILL` a `strace` (V52). (4) **`timeout` dentro de `strace`**: `strace` espera a todo lo que traza (`bash -c 'sleep 3 & true'`: 3 s y código 0), y `strace … -- timeout --kill-after=1s 2s bash -c 'setsid sleep 8 & tail -f /dev/null'` dura 8 s y sale con 124: un descendiente en otra sesión no recibe la señal del tope y mantiene vivo `strace` | contenedor local desechable de `ubuntu:24.04` (aarch64) con `strace` 6.8 del archivo de Ubuntu y coreutils 9.4, con `docker run --network none --cap-add SYS_PTRACE`; el binario, compilado del árbol como en V53; el programa del `connect` bloqueante, una sonda desechable fuera del repositorio que escucha en `127.0.0.1` con cola 0, sin aceptar, y conecta con sockets bloqueantes hasta quedarse esperando. En cada caso se leyeron la última línea y el tamaño de cada fichero, las líneas de creación y el recuento de formas de línea |
| V55 | Al leer JSON a una estructura, `encoding/json` (v1) se queda con el **último** valor de una clave repetida y no devuelve error, también con `Decoder.DisallowUnknownFields`, que es como lee H4 sus datos de prueba (`leerDatoDePrueba`). `encoding/json/v2`, que en Go 1.27.1 forma parte de la biblioteca estándar sin `GOEXPERIMENT`, rechaza por omisión la clave repetida, también dentro de un objeto anidado (con su ruta JSON); con `RejectUnknownMembers(true)` rechaza el miembro desconocido; rechaza lo que va detrás del valor; y compara los nombres distinguiendo mayúsculas | `go doc encoding/json/v2` (diferencias con v1: «By default, v1 allows for the presence of duplicate names, while v2 rejects duplicate names»; `RejectUnknownMembers`) y `go doc encoding/json/jsontext AllowDuplicateNames`; `go list -f '{{.GoFiles}}' encoding/json/v2` con go1.27.1 y `go env GOEXPERIMENT` vacío lista los ficheros del paquete; `internal/source/boe/casos_test.go` 646-664; sonda desechable fuera del repositorio con go1.27.1, sobre `{"busqueda": "a", "bloques": ["a21"], "busqueda": "b"}` leído a una estructura: v1 con `DisallowUnknownFields`, y `json.Unmarshal` de v1 → sin error y `Busqueda` igual a `b`; v2 → `jsontext: duplicate object member name "busqueda"`; v2 con un miembro `otra` → `unknown object member name "otra"`; con `{…} {}` → `invalid character '{' after top-level value`; con `Busqueda` → miembro desconocido; con la clave repetida en un objeto dentro de una lista → `duplicate object member name "busqueda" within "/0"` |
| V56 | El paso «Retirar Python del runner» **tal como era hasta T033** (contrato del job §1 en `6d68c21`) y la comprobación 3 de §3.1, en su forma literal —el paso con su primera orden `set -euo pipefail`, su consulta de paquetes y su línea `paquetes a purgar:`—, y la sexta orden de quickstart §12.2. **Desde T033 el paso no consulta ni purga paquetes** (D17): lo que aquí depende de la purga —la línea `paquetes a purgar:` y la salida de `apt-get` en (1) y (5), el paquete retenido de (7) y el control de (7) con la purga fallida— describe el paso anterior; la comprobación 3, (3) y (6) no cambian, y V60 repite con el paso nuevo (1) en su (2), (2) en su (3), (4) en su (4), (7) en su (5) y (5) en su (6). (1) **Runner con Python fuera del `PATH` y paquetes de Python**: CPython con la estructura de la caché de herramientas (`/opt/hostedtoolcache/Python/3.12.3/x64`, con `bin/python3.12` y sus enlaces `python3` y `python`, `lib/python3.12/` y `lib/libpython3.12.so.1.0`), PyPy igual (`/opt/hostedtoolcache/PyPy/3.10.14/x64`, con `bin/pypy3.10`, el enlace `pypy3`, `bin/libpypy3.10-c.so` y `lib/pypy3.10/`), Miniconda bajo `/usr/share/miniconda` (`bin/python3.12`, `bin/python3`, `lib/python3.12/` y un paquete `pkgs/python-3.12.4-h5148396_1/bin/python3.12`), un entorno de pipx (`/opt/pipx/venvs/yamllint`) con su enlace `bin/python` a `/usr/bin/python3.12`, que la purga deja roto, un `python3` suelto junto a otro ejecutable en `/opt/herramienta/bin`, un `python3` y un `lib/python3.12` en la instalación que contiene `claude` (`/opt/claude-code`, con `/usr/local/bin/claude` enlazado a ella), un enlace `/usr/local/bin/python3` con `/usr/local/lib/python3.12/dist-packages`, la `libpython3.13.so.1.0` de un programa congelado y un ejecutable `/dev/shm/python3`; los paquetes de V57 (1): `python3.12-minimal` con `/usr/bin/python3.12`, `libpython3.12t64:arm64` con su biblioteca, que depende de `libexpat-falsa:arm64`, instalada como automática, `pypy3` con `/usr/bin/pypy3`, `ipython3` y `python3-ajustes` con solo su configuración (`rc`); y, como señuelos, `python.vim`, la página de manual `python3.1.gz`, un `python.go` de modo 0444 en una caché de módulos y `/usr/share/doc/python3/`. La comprobación 3 termina con 1 y escribe `evals: hay Python accesible: <ruta>` para veintiuna rutas, una por línea —entre ellas `/var/lib/dpkg/info/libpython3.12t64:arm64.list` y `.md5sums`, del registro de dpkg—, y ninguna de un señuelo; `sin-python.txt` lleva `búsqueda: find / ( -path /proc -o -path /sys ) -prune -o ( ( -type f -perm /111 ( -iname python* -o -iname pypy* ) ) -o ( -type l ( -iname python* -o -iname pypy* ) ) -o ( ( -type f -o -type l ) ( -iname libpython* -o -iname libpypy* ) ) ) -print`, `usuario: root` y `resultado:` con esas veintiuna. El paso, ejecutado con `bash <paso>` sin ninguna opción, termina con 0: imprime sus dos marcas, `paquetes a purgar: libpython3.12t64:arm64 pypy3 python3.12-minimal` y la salida de `apt-get`, que retira esos tres y `libexpat-falsa` (V57 (3)), la búsqueda, nueve líneas `retirado:` y `búsqueda tras retirar: ninguno`; retira enteros `/usr/share/miniconda`, `/opt/hostedtoolcache/PyPy/3.10.14/x64`, `/opt/hostedtoolcache/Python/3.12.3/x64` y `/opt/pipx/venvs/yamllint`, y solo el fichero en `/usr/local/bin/python3` (el prefijo contiene `claude`), `/opt/claude-code/bin/python3` (la instalación contiene `claude`), `/opt/herramienta/bin/python3` (sin biblioteca estándar: `herramienta` sigue), la biblioteca del programa congelado (sigue su ejecutable) y `/dev/shm/python3`; siguen los señuelos, los dos `x64.complete`, `/usr/local/lib/python3.12/`, `/opt/claude-code/lib/python3.12`, `claude`, `/usr/bin/ipython3`, la configuración de `python3-ajustes`, el espacio de trabajo, el binario instalado y la skill instalada. Después, la comprobación 3 termina con 0 y `resultado: ninguno`. (2) **Imagen limpia**: la comprobación 3 da 0 con `búsqueda: …`, `usuario: root` y `resultado: ninguno`; el paso, 0 con `paquetes a purgar: ninguno` y sin ninguna línea `retirado:`. Con solo un enlace `/opt/otro/python3` y un ejecutable `/usr/share/otro/python3`, la comprobación 3 termina con 1 nombrando los dos. (3) **Sin root**: como `nobody` (`setpriv`), `find /` con la poda de `/proc` y `/sys` termina con 1 (`Permission denied` en `/root` y en `/var/cache/apt/archives/partial`, entre otros), y la comprobación 3, con un `sudo -n` que ejecuta como `nobody`, termina con 1 con `evals: no se pudo buscar Python como root en todo el sistema de ficheros` y sin escribir `sin-python.txt`. (4) **Solo lectura**: con un `python3` en un tmpfs remontado de solo lectura, el paso escribe su línea `retirado:` y termina con 1 tras `rm: cannot remove '/solo-lectura/usr/bin/python3': Read-only file system`. En todos los casos como root, `find` recorre `/dev` sin errores. (5) **Registro**: la extracción de la sexta orden, sobre registros sintéticos con un prefijo de tarea, paso e instante separados por tabuladores que reproducen también el texto del paso (`marca="retirada de Python"` y `echo "--- inicio de la $marca ---"`) y llevan la salida de (1), imprime solo de la marca de inicio de la salida a la de fin, las dos incluidas, con `paquetes a purgar:` y la salida de `apt-get`, y termina con 0; sin la marca de fin, porque el paso se detuvo, termina con 1 sin imprimir nada. (6) `gh run view --help` de gh 2.100.0 lista `--log-failed` («View the log for any failed steps in a run or specific job»). (7) **Las opciones las fija el paso**, ejecutado siempre con `bash <paso>`, sin opciones: con `python3.12-minimal` retenido (`apt-mark hold`), `apt-get purge` escribe `E: Held packages were changed and -y was used without --allow-change-held-packages` y el paso termina con 100 sin buscar ni escribir su marca de fin; con un `sudo` cuya búsqueda en `/` da su salida completa y termina con 1, el paso termina con 1 tras la línea `búsqueda:`, sin ninguna línea `retirado:`; sin `GITHUB_WORKSPACE`, con 1 y `GITHUB_WORKSPACE: unbound variable`; y con `sh <paso>` (`dash` 0.5.12), con 2 y `set: Illegal option -o pipefail` en la línea `set`, sin retirar nada. Control: el mismo paso sin la línea `set -euo pipefail` (79 líneas en lugar de 80) termina con 0 en los dos primeros casos —con la purga fallida retira los ficheros de los paquetes, `/var/lib/dpkg/info/libpython3.12t64:arm64.list` y `.md5sums` incluidos, y con la búsqueda fallida, las nueve rutas de (1), y en los dos escribe `búsqueda tras retirar: ninguno` y su marca de fin—: sin esa línea, un fallo de la purga o de la búsqueda pasaría en silencio | contenedores desechables de `ubuntu:24.04` (aarch64; bash 5.2.21, GNU findutils 4.9.0, GNU coreutils 9.4, dpkg 1.22.6, apt 2.8.3, dash 0.5.12, perl 5.38.2), con `docker run --network none` y, en (4), `--cap-add SYS_ADMIN --tmpfs /solo-lectura`; como root, con `sudo` sustituido por un guion que ejecuta sus argumentos (en (7), por uno cuya búsqueda en `/` termina con 1), `strace`, `node`, `go`, `make`, `git`, `claude` y el `kitlegal` instalado por ejecutables vacíos, y `GITHUB_WORKSPACE` y `HOME` con la estructura del runner; los paquetes, construidos e instalados como en V57; paso y comprobación extraídos tal cual, con `perl`, del contrato montado en solo lectura, dentro de cada contenedor, y ejecutados sin cambios (el control de (7) solo quita la línea `set`); (5), `rtk proxy sh` local sobre esos registros. Lo que trae y admite el runner real son los supuestos S2 y S7 |
| V57 | `dpkg-query` de dpkg 1.22.6 y `apt-get` de apt 2.8.3, los de Ubuntu 24.04, en lo que usaba el paso «Retirar Python del runner» hasta T033 (contrato del job §1 en `6d68c21`); **desde T033 el paso no usa ninguno de los dos** (D17, V60). En (3) ningún paquete del sistema dependía de los de Python; en `ubuntu-24.04` sí dependen, y la purga terminó con 100 (ejecución 34922606273 de T030; reproducido en V60 (1)). (1) **Sin patrón**: `dpkg-query -W -f='${db:Status-Abbrev} ${binary:Package}\n'` termina con 0 y da una línea por paquete registrado salvo los que no están instalados (99 líneas y ninguna `un` en la imagen con los paquetes de la prueba): el estado abreviado de tres caracteres —acción deseada, estado y marca de error, un espacio si no hay error—, un espacio y el nombre, con `:<arquitectura>` detrás en los paquetes `Multi-Arch: same` (`ii  libpython3.12t64:arm64`, `ii  libexpat-falsa:arm64`) y sin ella en los demás (`ii  python3.12-minimal`, `ii  pypy3`, `ii  ipython3`); un paquete retirado sin purgar sale `rc` (`rc  python3-ajustes`) y uno desempaquetado sin configurar, `iU` (`iU  python3-desempaquetado`); `read -r estado paquete` separa los dos campos. (2) **Con patrones**: `dpkg-query -W 'python*' 'libpython*'` da además los paquetes que dpkg solo conoce por las relaciones de otros, en estado `un` (`libpython3.4-minimal`, `libpython3.5-minimal`, `python-4suite`, `python-debian` y `python3-iptables`, también en la imagen limpia, que no tiene ningún Python), y termina con 0; con un patrón que no casa con nada (`libpypy*`) escribe `dpkg-query: no packages found matching libpypy*` y termina con 1 aunque el otro patrón case, con lo que casó en la salida; solo con uno que no casa (`zzzno*`), también 1; con una opción desconocida, 2. `dpkg-query --help` no lista códigos de salida y la imagen no trae la página de manual (`/etc/dpkg/dpkg.cfg.d/excludes` retira `/usr/share/man`), así que los códigos son los observados. (3) **Filtro y purga del paso**: con los estados `ii`, `rc` e `iU` a la vez, el paso elige `libpython3.12t64:arm64`, `pypy3`, `python3-desempaquetado` y `python3.12-minimal`, y no `python3-ajustes` (`rc`) ni `ipython3` (otro nombre); `apt-get purge -y --auto-remove` con esos nombres, con arquitectura, sin red y sin `apt-get update`, retira los cuatro y `libexpat-falsa` —una dependencia instalada como automática que ya nada necesita—, el paso termina con 0 sin ninguna línea `retirado:`, y en dpkg quedan solo `ii  ipython3` y `rc  python3-ajustes`. Con el filtro anterior, que solo elegía los instalados (`?i*`), `python3-desempaquetado` quedaba fuera y `apt-get` lo configuraba al purgar los demás (`Setting up python3-desempaquetado (1.0) ...`), con sus guiones de instalación. (4) **Retenido**: con `python3.12-minimal` retenido (`apt-mark hold`), `apt-get purge -y` termina con 100 y `E: Held packages were changed and -y was used without --allow-change-held-packages`, sin retirar ningún paquete | contenedores desechables de `ubuntu:24.04` (aarch64) con `docker run --network none`; paquetes falsos construidos con `dpkg-deb --build --root-owner-group` (uno `Multi-Arch: same` con `Depends`, otro con `conffiles`), instalados con `dpkg -i`, uno marcado con `apt-mark auto`, otro retirado con `dpkg -r` y otro desempaquetado con `dpkg --unpack`, y el árbol de construcción borrado antes de la prueba; el paso, extraído tal cual del contrato (V56). En el runner, la ejecución 34922606273 dio (1) con `:amd64` y terminó con 100 en (3) (D22, S2) |
| V58 | `go.yaml.in/yaml/v4`, la otra ruta de la organización YAML. (1) **Ya está en el módulo y en el binario**: `go.mod` 28 la lleva como `v4.0.0-rc.2 // indirect`; `go mod why -m go.yaml.in/yaml/v4` da `internal/cli` → `github.com/invopop/jsonschema` → `github.com/pb33f/ordered-map/v2` → `go.yaml.in/yaml/v4`; en `go mod graph` la piden `invopop/jsonschema@v0.14.0` y `pb33f/ordered-map/v2@v2.3.1`, las dos en `v4.0.0-rc.2`; y `modulosDelBinario` la lista («H1: lo importa github.com/pb33f/ordered-map/v2»). En una construcción hay una sola versión de cada módulo, así que la de las herramientas y la del binario serían la misma. (2) **Su README** (líneas 40-49, iguales en rc.2 y en rc.6), sección «Version Intentions»: «Versions `v1`, `v2`, and `v3` will remain as **frozen legacy**. They will receive **security-fixes only**…»; «All ongoing work, including new features and routine bug-fixes, will happen in **`v4`**»; «If you’re starting a new project or upgrading an existing one, please use the `go.yaml.in/yaml/v4` import path». (3) **Solo versiones candidatas**: `go list -m -versions go.yaml.in/yaml/v4` (2026-09-14) da `v4.0.0-rc.1` a `v4.0.0-rc.6` y ninguna estable; la misma orden con v3 da `v3.0.2` a `v3.0.5`. (4) **Su API cambia entre candidatas**: en rc.2, `(*Node).Decode` y `TypeError` están en `yaml.go` (142 y 358), con `UnmarshalError` (340), y la comprobación de claves repetidas en `decode.go` 783, cuyos tests esperan `yaml: unmarshal errors:` (`decode_test.go` 1771); en rc.6, `(*Node).Decode` está en `internal/libyaml/node.go` 282, el nuevo `LoadErrors` y `TypeError` son alias de `internal/libyaml` (`yaml.go` 500 y 513), `UnmarshalError` no se define en ningún fichero (solo lo nombran comentarios y nombres de sus tests), y la comprobación está en `internal/libyaml/constructor.go` 601, cuyos tests esperan `yaml: construct errors:` (`yaml_test.go` 1833) | `go.mod` 27-28; `go mod why -m go.yaml.in/yaml/v4 go.yaml.in/yaml/v3`; `go mod graph`; `internal/arch_test.go` (`modulosDelBinario`); `README.md`, `yaml.go`, `decode.go`, `decode_test.go` de `go.yaml.in/yaml/v4@v4.0.0-rc.2` y `README.md`, `yaml.go`, `yaml_test.go`, `internal/libyaml/node.go`, `errors.go` y `constructor.go` de `@v4.0.0-rc.6`, en la caché de módulos; `go list -m -versions go.yaml.in/yaml/v4 go.yaml.in/yaml/v3` |
| V59 | `contextcheck` v1.1.6, el de `tools/golangci-lint/go.mod`, marca toda llamada hecha desde una función con un `context.Context` entre sus parámetros o sus variables libres a otra que, sin recibir contexto, pasa a una tercera uno que no hereda, directamente o a través de sus propias llamadas estáticas: «Function … should pass the context parameter». El análisis cruza paquetes con hechos exportados y solo lo cortan un `nolint` en la declaración de la función llamada, una llamada dinámica (por una variable, un campo, un parámetro o una interfaz) o una función que devuelve un contexto; una función sin contexto que llama a otra así queda marcada igual, de modo que mover la llamada a un auxiliar sin contexto no cambia nada. `app.Main` es de esas: `ejecutarVerbo` abre el contexto de la invocación con `context.WithTimeout(context.Background(), ejecucion.Timeout)`, el plazo de `--timeout`, sin recibir el de quien llama. Por eso `Preparar(ctx, …)` con `app.Main` dentro no pasa el lint de `make ci` | `github.com/kkHAIKE/contextcheck@v1.1.6`: `contextcheck.go` 224-259 (`checkIsEntry`: contexto entre los parámetros o las variables libres), 519-563 (`checkFuncWithCtx`, el aviso), 566-637 (`checkFuncWithoutCtx`, el hecho que se propaga), 704-727 (`getFunction`: solo llamadas estáticas y cierres) y `README.md` («skip the check for the specified function»); `internal/app/main.go` 127-154 y 334-357; primera versión de `internal/evals/preparar.go` en T006, con `Preparar(ctx, …)`: `go tool -modfile=tools/golangci-lint/go.mod golangci-lint run ./internal/evals/...` → «Function Main->resolver->atender->resolverApplet->ejecutarVerbo should pass the context parameter (contextcheck)» en la llamada a `app.Main` |
| V60 | El paso «Retirar Python del runner» sin consulta ni purga de paquetes (contrato del job §1, desde T033) y la comprobación 3 de §3.1, en su forma literal, frente al paso anterior, y la sexta orden de quickstart §12.2. (1) **El caso del runner**: un paquete falso `sistema-esencial`, con `Essential: yes`, que depende de `python3.12-minimal` (con `/usr/bin/python3.12` y un `postinst`) y de `libpython3.12t64:arm64` (`Multi-Arch: same`, con `/usr/lib/aarch64-linux-gnu/libpython3.12.so.1.0`), como en `ubuntu-24.04` dependen paquetes del sistema de los de Python. La comprobación 3 termina con 1 y cinco rutas: el intérprete, la biblioteca, `/var/lib/dpkg/info/python3.12-minimal.postinst` y la `.list` y la `.md5sums` de `libpython3.12t64:arm64`. El paso anterior imprime su marca de inicio y `paquetes a purgar: libpython3.12t64:arm64 python3.12-minimal`, y `apt-get` escribe `sistema-esencial : Depends: python3.12-minimal but it is not going to be installed`, lo mismo de `libpython3.12t64`, y `E: Error, pkgProblemResolver::Resolve generated breaks, this may be caused by held packages.`, el error de la ejecución 34922606273 de T030; termina con 100 sin la línea `búsqueda:` ni la marca de fin, y en dpkg siguen los tres paquetes `ii` con sus ficheros. El paso nuevo, sobre ese mismo estado, termina con 0: marca de inicio, línea `búsqueda:`, `retirado:` de las cinco rutas, `búsqueda tras retirar: ninguno` y marca de fin; en dpkg siguen los tres `ii`, sigue el fichero de `sistema-esencial` y siguen `claude`, el espacio de trabajo, el binario instalado y la skill. Después, la comprobación 3 termina con 0 y `resultado: ninguno`. (2) **V56 (1) con el paso nuevo** (las mismas instalaciones y señuelos, y los paquetes de V57 (1)): la comprobación 3 termina con 1 y veintiuna rutas; el paso, con 0 y catorce líneas `retirado:` —enteros `/usr/share/miniconda`, `/opt/hostedtoolcache/PyPy/3.10.14/x64`, `/opt/hostedtoolcache/Python/3.12.3/x64` y `/opt/pipx/venvs/yamllint`; solo el fichero en `/usr/bin/python3.12`, `/usr/bin/pypy3`, `/usr/lib/aarch64-linux-gnu/libpython3.12.so.1.0`, `/usr/local/bin/python3` (el prefijo contiene `claude`), la `.list` y la `.md5sums` de `libpython3.12t64:arm64`, la biblioteca del programa congelado, `/opt/claude-code/bin/python3`, `/opt/herramienta/bin/python3` y `/dev/shm/python3`— y `búsqueda tras retirar: ninguno`. En dpkg siguen `ii` `ipython3`, `libexpat-falsa:arm64` (que la purga de V56 (1) retiraba como automática), `libpython3.12t64:arm64`, `pypy3` y `python3.12-minimal`, y `rc` `python3-ajustes`; siguen los señuelos, los dos `x64.complete`, `/usr/local/lib/python3.12`, `/opt/claude-code/lib/python3.12`, `claude`, `/usr/bin/ipython3`, la configuración de `python3-ajustes`, `herramienta`, el programa congelado, el espacio de trabajo, el binario instalado y la skill. Después, la comprobación 3 termina con 0 y `resultado: ninguno`. (3) **Imagen limpia**: la comprobación 3 da 0 y `resultado: ninguno`, y el paso, 0 sin ninguna línea `retirado:`; con un enlace `/opt/otro/python3` y un ejecutable `/usr/share/otro/python3`, la comprobación 3 termina con 1 nombrando los dos, el paso con 0 retirando los dos, y la comprobación 3, de nuevo, con 0. (4) **Solo lectura**: con un `python3` en un tmpfs remontado de solo lectura, el paso escribe `retirado: /solo-lectura/usr/bin/python3` y termina con 1 tras `rm: cannot remove '/solo-lectura/usr/bin/python3': Read-only file system`. (5) **Opciones y paquete retenido**, con `bash <paso>` salvo donde se dice: sin `GITHUB_WORKSPACE`, 1 y `GITHUB_WORKSPACE: unbound variable`; con `sh <paso>` (`dash`), 2 y `set: Illegal option -o pipefail`; con un `sudo` cuya búsqueda en `/` da su salida completa y termina con 1, 1 tras la línea `búsqueda:`, sin ninguna línea `retirado:` y con `/opt/otro/python3` en disco. Control: el mismo paso sin la línea `set -euo pipefail` (64 líneas en lugar de 65), con ese `sudo`, retira `/opt/otro/python3`, escribe `búsqueda tras retirar: ninguno` y su marca de fin y termina con 0. Con `python3.12-minimal` retenido (`apt-mark hold`), el paso anterior termina con 100 (`E: Held packages were changed and -y was used without --allow-change-held-packages.`) y el nuevo con 0, retirando `/usr/bin/python3.12`; el paquete sigue `hi` y la comprobación 3 da `resultado: ninguno`. (6) **Registro**: sobre registros sintéticos con el prefijo de tarea, paso e instante separados por tabuladores, que reproducen también el texto del paso, la parte de la sexta orden de quickstart §12.2 que lee el registro termina con 1 sin imprimir nada con la salida del paso anterior del caso (1), que no tiene marca de fin, y con 0 con la del nuevo, imprimiendo de la marca de inicio a la de fin de la salida, las dos incluidas, con la línea `búsqueda:`, las cinco `retirado:` y `búsqueda tras retirar: ninguno`. (7) **`dpkg --purge --force-depends` con los paquetes de Python de Ubuntu**: el `prerm` de `python3` (3.12.3-0ubuntu2.1) ejecuta `py3clean -p python3` si `which` lo encuentra, y `/usr/bin/py3clean` empieza por `#! /usr/bin/python3`. Retirado primero `python3.12-minimal` (3.12.3-1ubuntu0.16), con 0, `dpkg --purge --force-depends python3` escribe `py3clean: not found` e `installed python3 package pre-removal script subprocess returned error exit status 127`, falla también al deshacer (`py3compile: not found`, 127), termina con 1 y deja `python3` en `pF`; en el orden contrario, los dos terminan con 0. En todos los casos, el paso nuevo extraído del contrato es igual byte a byte al de `.github/workflows/evals.yml` y no tiene ninguna orden `dpkg` ni `apt`, y la comprobación 3 es la de antes de T033, con su línea `busqueda=` igual a la del paso | (1)-(6): contenedores desechables de `ubuntu:24.04` (aarch64; bash 5.2.21, GNU findutils 4.9.0, dpkg 1.22.6, apt 2.8.3, dash, perl), con `docker run --network none` y, en (4), `--cap-add SYS_ADMIN --tmpfs /solo-lectura`; como root, con `sudo` sustituido por un guion que ejecuta sus argumentos (en (5), por uno cuya búsqueda en `/` termina con 1), `strace`, `node`, `go`, `make`, `git`, `claude` y el `kitlegal` instalado por ejecutables vacíos, y `GITHUB_WORKSPACE` y `HOME` con la estructura del runner; los paquetes, construidos con `dpkg-deb --build --root-owner-group` e instalados con `dpkg -i`, como en V57; el paso anterior, extraído con `perl` del contrato de `6d68c21`, y el nuevo, del contrato del árbol y de `.github/workflows/evals.yml`, montados en solo lectura, ejecutados sin cambios (el control de (5) solo quita la línea `set`); en (6), la sexta orden desde `m="la retirada de Python"`, extraída con `perl` de quickstart.md y ejecutada con `sh`. (7): contenedores desechables sin red y como root de una imagen local de Ubuntu 24.04.4 con dpkg 1.22.6 y los paquetes de Python del archivo de Ubuntu; `/var/lib/dpkg/info/*.prerm` y `/usr/bin/py3clean` leídos en ella |
| V61 | La orden de la sesión del contrato del job §3.2 con Claude Code 2.1.270 en Linux, con `CLAUDE_CODE_SUBPROCESS_ENV_SCRUB` a `1` y a `0` (T036). (1) **Sin `bubblewrap`, con la variable a `1` y `--permission-mode bypassPermissions`** (la orden de la ejecución 34930222593 de T030): código 1, transcript vacío y, tras la cabecera del binario, la salida de error del runner («error: bubblewrap is required for subprocess env scrubbing and isolation. Install with: sudo apt-get install -y bubblewrap, …», en `nHn`); `LeerTrazas` lee la traza, sin invocaciones. (2) **Con `bubblewrap` y la misma orden**: la salida de error es solo el aviso «⚠ Permission mode forced to default — CLAUDE_CODE_SUBPROCESS_ENV_SCRUB is set …»; el transcript tiene `hook_started` y `hook_response` del gancho, `system`/`init` con `claude_code_version` `2.1.270` y `permissionMode` `default`, dos `api_retry` y `result` con `is_error` y «Failed to authenticate. API Error: 401 OAuth access token is invalid.»; código 1. Con el servidor local que sustituye a la API (debajo), `Skill` se ejecuta («Launching skill: boe-legislacion») y Bash se deniega («This Bash command contains multiple operations. The following parts require approval: …», en `permission_denials` y en un mensaje `system`/`permission_denied`): ninguna invocación del binario. (3) **Con `bubblewrap` y `--allowedTools Skill Bash`** en lugar de `--permission-mode`: sin aviso, con la salida de error vacía, `system`/`init` con `permissionMode` `default` y el mismo `result` de autenticación. El gancho ve `HTTP_PROXY`, `HTTPS_PROXY`, `NO_PROXY` y `KITLEGAL_CACHE_DIR` y no `CLAUDE_CODE_OAUTH_TOKEN`, con los espacios de nombres y los montajes del guion —no corre dentro de `bwrap`—; escribe en la caché, `127.0.0.1:9` rechaza la conexión y `boe articulo BOE-A-2015-10565 a21 --describe` por el enlace de la skill termina con 0. Con el servidor local, la orden de Bash no se ejecuta: su resultado es «Sandbox is required but failed to initialize: Sandbox dependencies not available: socat not installed. Restart to retry.». Con `socat` instalado se ejecuta bajo `bwrap`, con una línea `execve("/usr/bin/bwrap", ["bwrap", "--new-session", "--die-with-parent", "--ro-bind", "/", "/", "--bind", "/home", "/home", "--bind", "/root", "/root", "--bind", "/tmp", "/tmp", "--bind", "/var", "/var", "--bind", "/opt", "/opt", "--bind", "/run", "/run", "--bind", "/mnt", "/mnt", …, "--dev", "/dev", "--unshare-pid", "--unshare-user", "--cap-drop", "ALL", "--proc", "/proc", "--", "/bin/bash", "-c", "source <instantánea> … && eval '<orden>' …"], 0x… /* 36 vars */) = 0` de 22 001 octetos, cuyo argv enumera además montajes de solo lectura sobre `.claude`, `.git`, `.mcp.json`, `.gitmodules`, `.github` y ficheros de configuración del directorio personal, del de trabajo, de sus antecesores y de `GITHUB_WORKSPACE`. La orden ve espacios de nombres de usuario, PID y montaje nuevos, `CapEff` 0 y `NoNewPrivs` 1, seis procesos y ninguno con `CLAUDE_CODE_OAUTH_TOKEN` en su entorno, las variables de proxy y `KITLEGAL_CACHE_DIR` sin el token y `/` de solo lectura; la caché bajo `RUNNER_TEMP` y el directorio de trabajo son escribibles, `127.0.0.1:9` rechaza, el enlace de la skill resuelve al binario, `--describe` termina con 0 y las dos órdenes de la prueba de red de §6, con 5 y 4. Pero la traza no se puede atribuir: dentro del espacio de nombres de PID, el `bash` de la orden crea sus procesos con `clone(…) = 3`, `= 4`, `= 44` y `= 50`, sus números en ese espacio, mientras `strace -ff` nombra los ficheros con los del anfitrión (hasta `t.216`), y `LeerTrazas` falla con «dos líneas crean t.50: t.49, línea 1, y t.163, línea 8» (un hilo de `claude` y el `bash` del sandbox); sin la coincidencia, los ficheros de esos procesos se quedarían sin línea de creación. `strace` 6.8 anota el número del anfitrión con `--decode-pids=pidns`, que cambia las formas de línea de data-model §9. En Docker, `bwrap` necesita además `--security-opt systempaths=unconfined`: sin ella, la orden termina con «bwrap: Can't mount proc on /newroot/proc: Operation not permitted». (4) **Con la variable a `0` y `--permission-mode bypassPermissions`** (la orden del contrato desde T036), sin `bubblewrap` ni `socat`: sin aviso, con la salida de error vacía, `system`/`init` con `permissionMode` `bypassPermissions` y el mismo `result` de autenticación, código 1, y ningún fichero nuevo en el directorio de trabajo; el gancho ve el proxy y `KITLEGAL_CACHE_DIR` y no `CLAUDE_CODE_OAUTH_TOKEN`. Con el servidor local, `Skill` y Bash se ejecutan sin denegaciones, en los espacios de nombres del guion, con `CapEff` 0 y `NoNewPrivs` 0: la orden ve el proxy y `KITLEGAL_CACHE_DIR` y tampoco `CLAUDE_CODE_OAUTH_TOKEN`, que Claude Code no le pasa, pero lo encuentra en `/proc/<pid>/environ` de seis de los once procesos visibles, entre ellos el guion que lanza la sesión; escribe en la caché y en el directorio de trabajo, `127.0.0.1:9` rechaza, `--describe` termina con 0, `a9998 --json` con 5 y `a9998 --offline --json` con 4. `LeerTrazas` falla en la primera línea del proceso que crea la instantánea de shell antes de la primera orden de Bash, `execve("/bin/bash", ["/bin/bash", "-c", "-l", "SNAPSHOT_FILE=…"...], 0x… /* 27 vars */) = 0`, con «el argv de execve no es una lista de cadenas entre comillas»: `-s 4096` corta el tercer argumento con `"..."`, y lo mismo ocurre en (3) con `socat`. Con `-s 131072` en su lugar la línea queda entera (unos 7 000 octetos) y `LeerTrazas` lee la traza: cuatro invocaciones, las dos `--describe` del gancho y de la sonda, sin consulta y con 0; `a9998 --json`, con 5 y tres `connect` a `127.0.0.1:9` de clase `local`; y `a9998 --offline --json`, con 4 y sin conexiones. Ese corte no depende de la variable: lo arregla T037 | contenedores desechables de `ubuntu:24.04` (aarch64) con `strace` 6.8, `nodejs` y `npm` del archivo de Ubuntu, Claude Code 2.1.270 instalado con `npm install -g` y, según el caso, `bubblewrap` 0.9.0 y `socat` 1.8.0.0 del archivo; red solo para instalar paquetes. Como `runner` (uid 1001), con `docker run --security-opt seccomp=unconfined --security-opt apparmor=unconfined --cap-add SYS_PTRACE` y, en (3) con `socat`, `--security-opt systempaths=unconfined`; la skill del árbol con su enlace `scripts/boe` y `bin/instalado/kitlegal` hacia un `kitlegal` compilado del árbol para linux/arm64, como lo deja `make install`, y un gancho `SessionStart` en `~/.claude/settings.json` que escribe los nombres de las variables de su entorno, sus espacios de nombres, sus montajes, si escribe en la caché, la conexión a `127.0.0.1:9` y el `--describe` por el enlace de la skill. El bloque de la sesión, extraído con `perl` de `scripts/evals.sh`: del commit `417635e` en (1) y (2), del árbol con `--allowedTools Skill Bash` en (3) y del árbol final en (4), ejecutado sin cambios salvo el `-s` de la última traza. Con el token inválido, `CLAUDE_CODE_OAUTH_TOKEN` con un valor que no es un token, la red de Docker y la sesión fuera de `/home`: la API responde 401 y ninguna sesión llega al modelo. Para ejecutar herramientas sin modelo, con el directorio personal, `GITHUB_WORKSPACE` y `RUNNER_TEMP` en las rutas del runner, `--network none`, `api.anthropic.com` a `127.0.0.1` en `/etc/hosts`, `ANTHROPIC_BASE_URL=http://api.anthropic.com:8080` y un servidor local de usar y tirar que a cada petición a `/v1/messages` responde con `Skill`, después con Bash (una sonda que repite lo del gancho y busca el token en `/proc/*/environ` sin imprimir ningún valor, y las dos órdenes de la prueba de red) y por último con un texto. Las trazas, leídas por `LeerTrazas` en un arnés temporal borrado antes de `make ci` |

---

## D1 · Orden: de fuera adentro al planificar, de dentro afuera al implementar

**Decisión.** Se planifica desde lo que la skill debe resolver (sus evals) hacia las herramientas; se implementa: formato
de eval → preparación y arnés de grabación → grabación (pausa) → esquema de normas (pausa) → evals y `data/normas.yaml` →
generación de skills → pegamento y `make skills-sync`/`skills-check` → `SKILL.md` → instalación (guiones e2e antes del
código) → comparación e informe → job → documentación → plataforma. Detalle en plan.md, «Orden de implementación».

**Por qué.** Constitución, «Flujo de trabajo», punto 2. FR-065 exige las evals commiteadas antes que `SKILL.md` y que el
código de `scripts/skills-sync.sh`; y las evals necesitan identificadores verificados, que solo existen después de grabar
(FR-023). Por eso el formato y el arnés van antes que las evals, y las evals antes que la generación y la skill.

**Alternativas rechazadas.** *Evals primero con identificadores de memoria*: contradice FR-023 y dejaría a la tarea de
evals escribir datos que luego corrige una persona. *Skill primero y evals después*: contra FR-065 y la Definition of
Done §1.10.

## D2 · Dónde vive el código

**Decisión.** Dos paquetes nuevos de herramienta de desarrollo, que el binario distribuido no enlaza:
`internal/skills` (skill, frontmatter, normas, referencias, tabla, enlaces, deriva, validación YAML contra esquema) e
`internal/evals` (formato, conjunto, consultas necesarias, grabaciones, preparación de caché, trazas, sesión, citas,
comparación, informe). El único punto que necesita `--describe` en proceso, `TestSkillsDelRepositorio`, vive en
`internal/app/skills_test.go` (paquete `app`), igual que `TestEsquemasPublicados`, y usa la función `describir` del
kernel sin exportarla. Las órdenes son tests (D7) y guiones de shell.

**Por qué.** La tabla debe salir de lo que declara `--describe` (FR-032); la función que lo emite no está exportada (V25)
y `--describe` exige argumentos del verbo (V24), así que invocar el binario pediría inventar argumentos por verbo. El
precedente de H4 resuelve exactamente esto dentro de `internal/app`. El spec deja fuera los cambios del kernel (spec,
*Fuera de alcance*). La skill `golang-project-layout` pide que todo `package main` viva en `cmd/` (V39), y un segundo
binario en `cmd/` contradiría el ejecutable único (`CLAUDE.md`, «Multicall»).

**Alternativas rechazadas.** *Exportar `app.Describir`*: cambio del kernel que el hito no necesita. *Herramientas como
`package main` bajo `internal/`* (como el binario de e2e): nuevas excepciones de `forbidigo` para `os.Exit` y
`os.Stdout` y contra la regla de `golang-project-layout`. *`cmd/skills-sync`*: segundo binario. *`internal/docgen`*: en
`refs/kitlegal-estructura-y-ecosistema.md` §1 es el generador de escritos, no de skills.

**Dependencias del binario.** `TestDependenciasDelBinario` (`internal/arch_test.go`) sigue igual: si `internal/skills` o
`internal/evals` llegaran a enlazarse, aparecerían `go.yaml.in/yaml/v3` y `santhosh-tekuri/jsonschema` y el test fallaría.

## D3 · Frontmatter válido según el estándar Agent Skills

**Decisión.** Reglas de data-model §1.1: claves admitidas `name`, `description`, `license`, `allowed-tools`, `metadata`,
`compatibility`; `name` `^[a-z0-9]+(-[a-z0-9]+)*$` y ≤ 64, igual al directorio; `description` no vacía, ≤ 1024 runas y
sin `<` ni `>`; `compatibility` ≤ 500; `metadata` mapa de cadena a cadena.

**Por qué.** Es lo que aplica el validador del estándar que distribuye Anthropic con `skill-creator` (V15); en local no
hay otro texto del estándar. `^[a-z0-9]+(-[a-z0-9]+)*$` es equivalente a sus dos comprobaciones (clase y guiones). Dos
diferencias con el validador, las dos pedidas por el spec: se exige `description` no vacía (FR-040: el validador la deja
pasar vacía) y `name` igual al directorio (FR-001). Se cuenta en runas porque el validador cuenta caracteres, no bytes.

**Alternativas rechazadas.** *Solo `name` y `description`*: aceptaría claves que el validador rechaza. *Admitir claves
extra que Claude Code tolera* (`user-invocable`…): no son del estándar portable (constitución, «Restricciones técnicas»).

## D4 · Qué genera la sincronización para cada skill

**Decisión.** La skill declara en `metadata`: `kitlegal-applets` (applets, en orden, de los que genera tabla y enlaces) y
`kitlegal-referencias` (nombres `n` de `data/n.yaml` de los que genera `references/n.md`). `boe-legislacion` declara
`boe` y `normas`.

**Por qué.** FR-035 exige que la generación salga de la propia skill y de `data/`, sin listas escritas para una skill; y
FR-042 exige detectar un enlace o fichero **ausente**, lo que requiere una declaración independiente de lo que hay en
disco. `metadata` es la clave del estándar para datos de una herramienta (V15) y viaja con la skill.

**Alternativas rechazadas.** *Deducir los applets de los enlaces de `scripts/`*: un enlace borrado borraría la declaración
y no se detectaría. *Deducirlos de las marcas de la región*: igual. *Un manifiesto aparte en la skill*: fichero fuera del
estándar dentro del producto. *Leer `scripts/boe` en el texto de `SKILL.md`*: frágil.

## D5 · Destino de `scripts/<applet>` y enlace del binario instalado

**Decisión.** `skills/<skill>/scripts/<applet>` es un enlace versionado a `../../../bin/instalado/kitlegal`. `make
install` crea `bin/instalado/kitlegal` (ignorado, `/bin/`) como enlace a la ruta que da `go list -f '{{.Target}}'`
(V17), donde `go install` dejó el binario.

**Por qué.** El enlace versionado tiene que ser relativo y resolver al **binario instalado** (FR-051), cuya ruta depende
de cada cuenta; ese salto solo lo puede dar la instalación. Resolver a través de `~/.claude/skills/<skill>` funciona
porque el sistema resuelve el destino relativo desde el directorio físico del enlace, y el multicall sigue viendo `boe`
(V34, V37). Un nombre propio (`bin/instalado/`) separa lo que ejecutan las skills instaladas de lo que construye
`make build` (`bin/kitlegal`).

**Alternativas rechazadas.** *`../../../bin/kitlegal`* (`refs/kitlegal-estructura-y-ecosistema.md` §1): `make build`
escribe ahí, así que la skill instalada ejecutaría la última construcción y no el binario instalado, o `make install`
tendría que convertir la salida de `make build` en un enlace. *Wrapper de shell*: decisión cerrada de `CLAUDE.md` y
FR-084. *Enlace por skill a `$GOBIN` dentro de `scripts/`*: un fichero no versionado dentro del producto, uno por skill.
*Destino absoluto*: no portable.

## D6 · Tabla de comandos desde `--describe`

**Decisión.** Formato de contrato de sincronización §3. La sintaxis se deriva de la entrada: obligatorio → `<nombre>`,
de varios valores → `<nombre>...`, opcional → `[--nombre]`; las banderas globales se obtienen describiendo un verbo sin
argumentos con `cli.Describir`. `TestTablaDeComandosCoincideConLaGramatica` invoca cada sintaxis generada contra la
gramática real con `--describe`.

**Por qué.** `--describe` declara obligatoriedad, orden y tipo, pero no si un argumento va por posición (V24). Hoy todos
van por posición (V28); la comprobación contra la gramática convierte esa inferencia en un control: si un verbo tuviera
una bandera obligatoria, la invocación generada fallaría. El spec prohíbe ampliar el applet para la tabla (*Fuera de
alcance*).

**Alternativas rechazadas.** *Tabla sin sintaxis* (solo nombres): menos útil para el agente sin ganar corrección.
*Leer las etiquetas de Kong en la generación*: sería otra fuente que `--describe`, contra FR-032.

## D7 · Sincronización y comprobación como tests

**Decisión.** `make skills-check` = una sola invocación de `go test` sobre `internal/app`, `internal/skills` e
`internal/evals` con los tests «del repositorio»; `scripts/skills-sync.sh` = el mismo test de `internal/app` con
`-args -regenerar-skills`. `skills-check` entra en `make ci`.

**Por qué.** Es el patrón de `make schema-check` (`TestEsquemasPublicados -actualizar-esquemas`): la comprobación
regenera en memoria y compara sin escribir (FR-042), la regeneración es el mismo código con escritura, y no hace falta
ningún `main`. Una receta de dos órdenes se detendría en la primera que falla (V40) y ocultaría la segunda.

**Alternativas rechazadas.** *Un binario de herramienta*: D2. *Solo `make test`*: los tests ya corren ahí, pero la
comprobación mecánica de skills es un control con nombre en el hito y `make skills-check` la hace ejecutable sola.

## D8 · `data/normas.yaml`: mapa por identificador, esquema y biblioteca YAML

**Decisión.** Documento `normas:` indexado por identificador (contrato de normas §1-§2). Todo documento YAML del hito
(`data/normas.yaml`, las evals y el frontmatter de `SKILL.md`) pasa por el mismo lector de `internal/skills/esquemas.go`
(data-model, «Lectura de documentos YAML»): `go.yaml.in/yaml/v3` lo analiza a un `yaml.Node`, que conserva las líneas;
el lector recorre el nodo y rechaza toda clave repetida en cualquier mapa, nombrando la ruta, la clave y sus dos líneas;
convierte el nodo a `any` con `(*yaml.Node).Decode`; lo normaliza a tipos JSON (lo codifica en JSON y lo lee con
`jsonschema.UnmarshalJSON`), y lo valida con `santhosh-tekuri/jsonschema/v6` contra su esquema, aquí
`schemas/normas.yaml.json`. `rango` es un `enum` con el vocabulario de las búsquedas grabadas.

**Por qué.** JSON Schema 2020-12 no puede exigir unicidad de un campo entre elementos de una lista; en un mapa el
identificador es el nombre de la propiedad y el esquema lo valida con `propertyNames`, pero la unicidad de la clave es
cosa del lector, porque el documento que llega al validador ya no puede tenerla repetida. Ese lector no puede ser
`yaml.Unmarshal` a un `yaml.Node`: copia el nodo sin comprobar repetidos, y al convertirlo la última clave ganaría en
silencio (V43). El recorrido propio presenta el defecto nombrando la entrada y sus dos líneas (FR-025, US6-3) con un
mensaje del proyecto, y no con el texto de la biblioteca; `Decode` aplica además la comprobación de repetidos de la
biblioteca (V43), así que una clave repetida no pasa por ninguno de los dos caminos. La normalización a tipos JSON es la
que el propio validador usa para leer JSON (`json.Number`, V44), y un mapa con claves que no son texto falla al
codificarse en lugar de validarse a medias. `go.yaml.in/yaml/v3` es el mismo código que `gopkg.in/yaml.v3`, mantenido, y
ya está en el grafo (V31) — Complexity Tracking. El `enum` de `rango` sale de la fuente y no de una lista escrita a mano
(V29), y un test lo compara con lo grabado.

Por qué v3 y no `go.yaml.in/yaml/v4`, aunque su organización deje v3 como legado congelado que solo recibe correcciones
de seguridad y pida v4 a los proyectos nuevos (V31, V58 (2)):
- **v4 no tiene ninguna versión estable**: solo `v4.0.0-rc.1` a `v4.0.0-rc.6` (V58 (3)). Entre la candidata que ya está
  en el grafo (rc.2) y la última (rc.6), `(*Node).Decode` se define en un paquete interno, `TypeError` es un alias de un tipo
  de ese paquete, `UnmarshalError` desapareció y el error de clave repetida cambió de texto (V58 (4)). Lo que el lector usa y V43 comprueba —que analizar
  a `yaml.Node` no detecta la clave repetida y que `Decode` sí— no está fijado en v4 de una candidata a otra; en v3 sí.
- **v4 la enlaza el binario**, por `internal/cli` → `invopop/jsonschema` → `pb33f/ordered-map/v2` (V58 (1)). En una
  construcción hay una sola versión de cada módulo: hacerla directa en las herramientas ata su versión a la del binario,
  y subirla para unas —a otra candidata, con otra API— movería la del otro. v3 no llega al binario (V31), y
  `TestDependenciasDelBinario` sigue sin cambios.
- **v3 es la biblioteca de la lista de la constitución** (§V: `gopkg.in/yaml.v3`, el mismo código que
  `go.yaml.in/yaml/v3`, V31); v4 es otra versión mayor, con otra API.

El coste que se acepta con v3 es que un fallo que no sea de seguridad no se corregirá en ella. El lector solo usa el análisis a
`yaml.Node` y `(*yaml.Node).Decode`, comprobados en V43, y cualquier desviación la harían visible `TestLeerEval`,
`TestLeerNormas` y `TestValidarDocumentoYAML`.

**Alternativas rechazadas.** *`yaml.Unmarshal` a un `yaml.Node` y conversión propia*: no detecta la clave repetida
(V43). *Solo `yaml.Unmarshal` o `Decode` a `any`*: detecta la repetida, pero con el mensaje de la biblioteca, y los
errores del esquema se quedan sin la línea del documento. *Lista con comprobación de repetidos en Go*: el esquema no la
declararía. *Lista con un vocabulario propio `uniqueKeys`* (V32): palabra clave no estándar que otros validadores
ignoran. *`gopkg.in/yaml.v3`*: sin mantenimiento según su sucesor y añadiría un módulo nuevo. *`go.yaml.in/yaml/v4`*,
la ruta que su README pide a los proyectos nuevos y que ya está en el grafo (`go.mod` 28): sin ninguna versión estable,
con la API y el error de clave repetida cambiados entre candidatas, y con una sola versión para las herramientas y el
binario (V58; arriba). *`rango` como cadena libre*: aceptaría rangos que la fuente no usa.

## D9 · Verificación offline de identificadores

**Decisión.** `TestIdentificadoresDeLasNormas` reproduce en proceso, con el applet `boe` sobre `httpx.Replay`, cada
búsqueda del manifiesto, y exige para cada norma un resultado con el mismo identificador y título.

**Por qué.** FR-024 pide repetir sin red la verificación con `boe buscar`; hacerlo con el mismo applet sobre las mismas
respuestas es la misma verificación. Detecta una edición posterior del identificador o del título (SC-008).

**Alternativas rechazadas.** *Leer los JSON grabados directamente*: duplicaría la lectura de la fuente. *Comprobación
única anotada en la tarea*: el spec la descarta (*Assumptions*).

## D10 · Formato de eval y `schemas/eval.yaml.json`

**Decisión.** YAML por eval con `pregunta`, `activa`, `reproduce` opcional, `comandos` (tres formas excluyentes) y
`citas`, validado contra un esquema publicado (contrato de evals §1).

**Por qué.** `docs/ROADMAP.md` §3 fija `evals/<skill>/*.yaml` y los cuatro campos; FR-060 y la clarificación fijan que el
comando declara lo consultado. Un esquema publicado documenta el formato «válido para cualquier skill» y hace mecánica la
detección de ficheros mal formados (US4-4) con la biblioteca de validación ya fijada. `reproduce` es la marca que exige
FR-064 dentro del fichero. `comandos` es la palabra del spec; `misspell` la marca (V42), así que entra en
`misspell.ignore-rules` con su motivo, como las palabras de contrato de H4.

**Alternativas rechazadas.** *Línea de órdenes literal*: la clarificación la descarta. *Validación solo en Go*: el
formato quedaría documentado en prosa. *JSON*: el roadmap dice YAML. *Otra palabra que `comandos`*: el formato dejaría
de hablar como el spec.

## D11 · Grabación: manifiesto en `testdata/` de la raíz, arnés con el applet, reutilización de H4

**Decisión.** Manifiesto `testdata/evals/grabaciones.json` (búsqueda, prefijo del título y bloques por norma); arnés
`TestGrabarEvals` (etiqueta `grabacion`, en `internal/evals`) que siembra una caché con las grabaciones de H4, ejecuta en
proceso `boe buscar`, `metadatos`, `indice` y `articulo` con un cliente que graba con raíz `testdata/evals/`
(`../../testdata/evals` desde el paquete, V46), y resuelve el identificador por el prefijo; grabaciones nuevas en
`testdata/evals/boe.legislacion-consolidada/`. Las sesiones y trazas sintéticas, que no son respuestas de ninguna fuente,
siguen en `internal/evals/testdata/sesiones/`.

**Por qué.** El ejecutor no toca la red, así que la grabación la hace una persona en la pausa de la tarea `[datos]` del
manifiesto, como en H4. Pero esa pausa no la provoca cualquier fichero nuevo de datos de prueba: `clasificar_datos` solo
pausa si la tarea modifica material existente o **añade** un fichero bajo `testdata/` de la raíz, bajo
`internal/source/` o bajo `schemas/` (V36). En H4 el manifiesto estaba en `internal/source/boe/testdata/` y pausaba; en
H5 el manifiesto tiene que ser un fichero nuevo en una de esas tres ubicaciones o la persona nunca grabaría.
`internal/source/` queda descartada porque H5 no cambia la fuente (*Fuera de alcance*), y `schemas/` no es sitio para
datos de prueba; queda `testdata/` de la raíz, que es también donde `docs/ROADMAP.md` §3 sitúa lo que graba `httpx`
(«graba a `testdata/<fuente>/`»). Dentro, `evals/` agrupa el manifiesto y lo grabado para las evals de las skills; la
raíz de grabación es ese directorio y `httpx` crea bajo él el de la fuente (V46). Grabar ejecutando el applet es
literalmente «verificar con `kitlegal boe buscar`» (FR-023), sin depender de funciones no exportadas de
`internal/source/boe`. Sembrar desde H4 toda consulta, también la búsqueda, hace que lo ya grabado no se vuelva a pedir (FR-074): la búsqueda
`procedimiento administrativo común` de la primera entrada es la misma que grabó H4, y pedirla a la red escribiría en
H5 un fichero con el mismo nombre que el de H4 (`httpx` nombra cada grabación por método y URL), que
`TestGrabacionesSinSolape` rechazaría sobre grabaciones que el ejecutor no puede tocar. Resolver por prefijo evita
escribir identificadores antes de conocerlos.

**Alternativas rechazadas.** *`internal/evals/testdata/`, junto al código que lo usa*: un fichero nuevo ahí no hace
pausar al workflow (`clasificar_datos` da `pausa:false`, V36); la persona nunca grabaría, la tarea del esquema de normas
se quedaría sin rangos grabados y las respuestas de la fuente no pasarían por la revisión humana de la capa 3. *Añadir
las grabaciones al directorio de H4*: pausaría, pero cambia `internal/source/boe`. *`testdata/` de la raíz como raíz de
grabación* (`testdata/boe.legislacion-consolidada/`, con el manifiesto suelto en `testdata/`): pausa igual, pero mezcla
en la raíz de `testdata/` un manifiesto de las evals con el directorio de una fuente, con el mismo nombre que el de H4.
*Hacer que `clasificar_datos` pause también bajo `internal/*/testdata/`*: es proceso, fuera de la rama del hito (spec,
*Fuera de alcance*). *Manifiesto con identificadores*: identificadores de memoria. *Reutilizar
`scripts/grabar-fixtures.sh`*: volvería a grabar lo de H4.

**Lectura del manifiesto** (contrato de evals §3.1; data-model §7.2). Un único lector, `LeerManifiesto(contenido []byte)
(Manifiesto, error)` en `internal/evals/grabaciones.go`, del que dependen `TestGrabarEvals`,
`TestIdentificadoresDeLasNormas` y `TestManifiestoDeGrabaciones`: así el manifiesto que se graba, el que verifica los
identificadores y el que se comprueba son el mismo documento leído con las mismas reglas. Decodifica con
`encoding/json/v2` y `RejectUnknownMembers(true)`, que rechazan la clave repetida, el miembro desconocido y lo que va
detrás del valor (V55), y después comprueba la fuente, las entradas y la norma repetida. Como el manifiesto no lleva
identificadores, una norma repetida es una pareja de entradas cuyos `titulo_empieza_por` son iguales o uno empieza por el
otro: una entrada resuelve el único resultado cuyo título empieza por su prefijo, y un mismo título solo empieza por dos
prefijos si uno empieza por el otro, así que la regla recoge toda pareja que podría grabar dos veces la misma norma sin
necesitar ninguna grabación. El test valida contenidos sintéticos escritos como constantes, que llegan con el lector en el
paso 3 del orden de implementación, y el manifiesto real en su subtest `repositorio`, que llega en el paso 6 porque el
manifiesto lo crea el paso 4.

*Alternativas rechazadas.* *`encoding/json` con `DisallowUnknownFields`, como `leerDatoDePrueba` de H4*: con una clave
repetida se queda con el último valor sin error (V55), el mismo defecto que el lector de YAML evita recorriendo el nodo
(V43). *Un esquema del manifiesto en `schemas/`*: un fichero nuevo bajo `schemas/` más, con su pausa, y JSON Schema no
expresa que un prefijo empiece por el de otra entrada, así que la regla seguiría en Go. *Norma repetida como
`titulo_empieza_por` exactamente igual*: deja pasar `Ley 39/2015,` junto a `Ley 39/2015, de 1 de octubre`, que resuelven
la misma norma. *Detectar la repetición solo por los identificadores resueltos*: necesita las grabaciones, así que
llegaría con `TestIdentificadoresDeLasNormas` en el paso 7 y dejaría el lector sin esa regla en los pasos 3 a 6. *Leer
el manifiesto real en el paso 3*: el fichero no existe hasta el paso 4 y `make ci` fallaría.

## D12 · Qué se registra de una sesión y cómo se compara

**Decisión.** Activación, respuesta, modelo y versión, del transcript `stream-json`; si la sesión terminó, de su código
de salida (que el guion escribe siempre en `codigo-de-la-sesion`) y del último mensaje del transcript, que tiene que ser
`result` con `subtype: success` e `is_error: false` (V47); invocaciones (argv y código) y conexiones, de una traza
`strace -ff` de la sesión entera. Una eval cuya sesión no terminó no pasa, y el informe da el motivo (tope de 240 s,
error de la API, turnos agotados, sin `result`). Comparación mecánica de data-model §6.1, §6.2 y §9: applet por
el nombre de invocación, banderas globales retiradas (con valor en `--timeout` y `--asunto`, según el tipo que declara
`cli.Globales`), `--describe` y `--dry-run` sin consulta, bloques por `articulo`/`articulos`, términos de `buscar` como
palabras.

**Por qué.** La herramienta Bash no informa del código de salida de cada invocación y una orden puede encadenar varias
(V9); los ganchos tampoco lo dan (V13). La traza del sistema registra lo que se ejecutó de verdad, sin tocar la skill ni
el binario (FR-077). `--describe` y `--dry-run` no consultan (V24; y `descripcionDeLaOperacion`, en `internal/app/main.go`, que con
`--dry-run` escribe «no se ha ejecutado nada»): con FR-072 «consultó lo mismo» y la
clarificación («consultar es obtener el texto»), no pueden satisfacer un comando esperado.

**Alternativas rechazadas.** *Resultado de Bash en el transcript*: sin código por invocación. *Ganchos `PostToolUse`*:
igual, y `--bare` los desactiva. *Un envoltorio del binario que registre*: cambia lo que ejecuta la skill instalada.
*Leer las citas con un modelo*: FR-072 lo prohíbe.

**Atribución de hilos y conexiones** (data-model §9; contrato del job §4). `strace -ff` deja un fichero por hilo (V53), y
una conexión solo cuenta para una invocación si su hilo es del proceso de esa invocación. El runtime de Go crea sus hilos
con `clone`, no con `clone3`, y a menudo desde un hilo que no es el principal (V51; en la traza de V53, `t.43` crea
`t.44`), así que la pertenencia se sigue por `CLONE_THREAD` de forma transitiva, con `clone` y con `clone3`. Lo no
atribuido se ignora, porque las conexiones de `claude` a la API del modelo son públicas; y justo por eso un fallo de la
atribución no puede ser silencioso: un hilo sin atribuir dejaría su conexión ignorada y `red` vacío en falso (FR-076). La
traza tiene que tener exactamente un fichero sin la línea que lo crea y solo líneas de las formas comprobadas en V53 y,
las de señal, en V54, salvo en una sesión cortada (abajo); cualquier otra cosa la hace ilegible y la sesión no pasa, con
el fichero, la línea y su texto en el informe. En la traza real de V53 los `connect` de la invocación salían de su hilo
principal, pero en V54 (1) y (2) la misma invocación los hace también desde hilos que no lo son, uno de ellos creado por
otro hilo: la conexión `local` de la prueba de red demuestra en el runner la lectura de `connect` y su atribución a la
invocación, que puede pasar por hilos, y la atribución por hilos queda demostrada con trazas reales porque ninguna sesión
no cortada sea ilegible, es decir, porque cada fichero de hilo tiene su línea de creación reconocida (S4). Cada regla la fija un caso de `TestLeerTrazas`: `hilo-por-clone` y `hilo-por-clone3` (las dos
llamadas), `hilo-de-un-hilo` (la transitividad), `connect-fuera-de-la-invocacion` (se ignoran el `connect` de otro
proceso y el anterior a la `execve` del applet), y `fichero-sin-origen` y `fichero-ilegible` (lo que no se entiende no se
ignora).

*Alternativas rechazadas.* *Reconocer solo `clone3`*: no ve ningún hilo de Go (V51). *Atribuir solo los hilos que crea el
hilo principal*: no ve los que crea otro hilo (V53). *Ignorar las líneas y los ficheros que no se entienden*: un formato
distinto en el runner dejaría conexiones sin atribuir y `red` vacío sin que nada fallara. *Contar las conexiones de todos
los procesos de la sesión*: las de `claude` a la API del modelo harían fallar cada ejecución. *Atribuir a la invocación
también los procesos que crea*: el binario no crea procesos (en `internal/` y `cmd/`, `os/exec` solo lo importan
ficheros de test), y un hijo que ejecuta otro programa ya no es el applet.

**Señales y sesión cortada por el tope** (data-model §9, reglas 5 y 6; contrato del job §3.2 y §4). La traza se toma
**sin** `-e signal=none`. Con esa opción, `strace` no escribe `+++ killed by SIG… +++`, y el fichero de todo proceso que
muere por una señal se queda sin línea final, haya tope o no (V54 (1)): con la regla de que cada fichero termina en ella,
una orden de la sesión muerta por una señal haría ilegible una sesión sin corte, y un corte haría ilegible toda sesión
cortada, que es lo que observó el juez del plan. Sin ella, cada proceso que muere por una señal deja
`+++ killed by SIG… +++`, y cada señal entregada, una línea `--- SIGNOMBRE {…} ---`, que `LeerTrazas` admite y no cuenta
(V54 (2): siete `SIGURG` del runtime de Go en la orden de V53). Con 0 y con 124 todos los ficheros terminan en su línea
final (V54 (2); con TERM, `strace` sigue hasta el final, V54 (3)); con 137, `timeout` mata `strace` con `KILL` y los
ficheros de los procesos vivos se quedan sin ella (V54 (3)). Por eso `LeerTrazas(dir, cortada)` recibe si el tope cortó
la sesión (`codigo-de-la-sesion` 124 o 137, que lee `LeerSesion`) y, solo entonces, admite lo que puede dejar un corte:
ficheros sin línea final o vacíos, y una llamada con el resultado `? ERRNO (descripción)` a la que en su fichero solo
siguen líneas de señal y, como mucho, la línea final, que es la que interrumpió la señal (V54 (1) y (2)). Sigue exigiendo
el origen de cada fichero, las formas de las demás líneas y la atribución. La invocación cuyo hilo principal no llegó a
su línea final queda sin código y no satisface ningún comando esperado; un `connect` interrumpido se clasifica por su
dirección, y fuera del bucle local cuenta como `red`. La sesión lleva el motivo del tope y no `sesión ilegible`, y sus
conexiones y sus invocaciones fuera de lo grabado se informan (FR-071, FR-076). La regla vale con 124 y con 137, aunque
con 124 V54 no muestre ficheros sin línea final: una sesión cortada no pasa en ningún caso, y su traza no es evidencia del
formato de una traza completa (S4), así que se anota como S9. Cada regla la fija un caso: `lineas-de-senal`,
`cortada-por-el-tope`, `sin-linea-final-sin-corte`, `cortada-con-llamada-interrumpida`, `llamada-interrumpida-sin-corte`
y `llamada-interrumpida-antes-del-final` de `TestLeerTrazas`; `sesion-sin-terminar` (124) y
`sesion-cortada-con-invocaciones` (137) de `TestInforme`; y `bloque-leido-sin-codigo` de `TestJuzgar`.

*Alternativas rechazadas.* *Mantener `-e signal=none` y tratar como ilegible todo fichero sin línea final* (la regla
anterior): toda sesión cortada daría `sesión ilegible` en lugar del motivo del tope, sus conexiones no llegarían a `red`
ni sus invocaciones a `fuera_de_lo_grabado`, un tope en la prueba de red se tomaría por un formato de traza distinto
(S4), y lo mismo pasaría con una sesión sin corte en la que un proceso muriera por una señal (V54 (1)). *Mantener
`-e signal=none` y admitir siempre ficheros sin línea final*: la invocación muerta por una señal quedaría sin código en
lugar de con uno distinto de 0, y un fichero incompleto por cualquier otra causa pasaría en silencio. *Una lista de
señales en `-e signal=`*: habría que comprobar qué deja cada señal que faltara de la lista; sin la opción, V54 (2) muestra
el comportamiento completo, y las líneas de señal no cambian nada de lo que se lee. *Tratar las líneas de señal como
defecto*: el runtime de Go las produce en cada invocación (V54 (2)). *Que `LeerTrazas` lea `codigo-de-la-sesion` por su
cuenta*: dos lectores del mismo fichero que podrían discrepar cuando falta o no es un entero, algo que `LeerSesion` ya
resuelve. *Deducir el corte de la propia traza* (algún fichero sin línea final): convertiría el defecto que detecta la
regla 5 en su propia excusa. *Admitir el resultado `? ERRNO (…)` seguido de otra llamada, o en una sesión sin corte*: V54
solo lo muestra en la llamada que interrumpe el corte, seguida de líneas de señal y de la línea final; en otro sitio no
está comprobado y sería un formato distinto (S4). *Contar como `bloqueada`, o ignorar, un `connect` interrumpido a una
dirección que no es local*: nada muestra que no llegara a la red, y FR-076 no deja sin detectar una llegada. *Poner
`timeout` dentro de `strace`* (`strace … -- timeout … claude`), para que `strace` escriba siempre las líneas finales:
`strace` espera a todo lo que traza, y un descendiente en otra sesión, al que no llega la señal del tope, lo mantiene vivo
más allá de los 240 s (V54 (4)), de modo que la sesión dejaría de estar acotada.

## D13 · El job: plataforma, disparadores, modelo y sesión

**Decisión.** `.github/workflows/evals.yml` en `ubuntu-24.04` con `workflow_dispatch`, `schedule` semanal y
`pull_request` `labeled` (`evals`, `evals-prueba-de-red`); modelo `claude-haiku-4-5-20251001` fijado en el fichero; Claude
Code 2.1.270; una sesión por eval en un directorio de trabajo vacío, `--setting-sources user`, `--max-turns 30`, tope de
240 s, sin `WebFetch` ni `WebSearch` (contrato del job).

**Por qué.** FR-070 pide ejecución semanal, manual y sobre la rama del hito antes de fusionar. `schedule` corre en la
rama principal; para una rama cuyo fichero aún no está en la principal, el evento de etiqueta de la propuesta de cambio
usa el fichero de esa rama (supuesto S1, que prueba la tarea `[plataforma]`); `gh workflow run --ref` no documenta ese
caso (V14). `ubuntu-24.04` fija la imagen, y Linux es necesario para `strace`. La versión 2.1.270 es aquella cuyo
comportamiento se comprobó (V1-V13). El tope de 240 s queda por debajo de la vigencia de 300 s (V22). Un directorio vacío
fuera del repositorio y `--setting-sources user` evalúan solo lo que dejó `make install` (FR-077).

**Alternativas rechazadas.** *Contenedor sin Python*: añade supuestos sobre `ptrace` en contenedores y sobre acciones en
contenedor, sin ganar nada que la retirada de D17, comprobada antes de cada ejecución, no dé. *`anthropics/claude-code-action`*: no hay nada que
comprobar en local y no expone la traza. *`push` a la rama*: evaluaría en cada empuje. *Modelo como entrada del job*:
FR-070 lo prohíbe.

**Lectura del job desde la plataforma** (quickstart §12; contrato del job §3.3 y §7). La ejecución de una etiqueta se
elige entre las de la rama con evento `pull_request` por el campo `workflowName` (`evals`, el `name:` del fichero), que
`gh run list --json` expone (V45), y no con `--workflow evals.yml`, que obliga a `gh` a resolver por su fichero un flujo
que hasta la fusión solo está en la rama de la propuesta de cambio (V14 no lo cubre); el valor del campo en ese caso es
el supuesto S12 (6), y la orden que muestra la ejecución imprime también las demás posteriores a la etiqueta con su
`workflowName` para registrarlo. El guion imprime `informe.md` e `informe.json` entre cuatro marcas literales, y las
órdenes los sacan del registro de marca a marca, sin tope, fallando si falta una (V48). La evidencia de SC-003 es una
sola salida, la misma que se registra: la lista completa de ficheros cambiados y la comprobación de que todos están
bajo el directorio del hito (V49).

*Alternativas rechazadas.* *`gh run list --workflow evals.yml`*: supuesto sobre `gh` y la plataforma que ninguna
comprobación local cubre y que, si falla, deja las órdenes sin ejecución que leer. *Buscar el encabezado del informe con
`grep -A <n>`*: corta el informe cuando pasa de `n` líneas, no detecta que falte el final y depende de un encabezado que
nada fija. *Subir el informe como artefacto*: una acción más en el flujo y una descarga que no está entre las formas
permitidas de la sesión. *Registrar la lista completa y comprobar aparte, con una exclusión de ruta que se espera vacía*:
son dos salidas distintas, y lo registrado dejaría de ser lo que la orden comprueba.

**La sesión y la credencial del modelo** (T036; V12 y V61). *Decisión.* La orden de la sesión fija
`CLAUDE_CODE_SUBPROCESS_ENV_SCRUB=0` y conserva `--permission-mode bypassPermissions`; el job no instala `bubblewrap` ni
`socat`. Claude Code no pasa `CLAUDE_CODE_OAUTH_TOKEN` al entorno de las órdenes ni al de los ganchos de la sesión, pero
la variable queda en el entorno inicial de `claude` y de los procesos que lo lanzan, que una orden del mismo usuario
puede leer en `/proc/<pid>/environ` (V61 (4)).

*Por qué.* Con la variable a `1`, el binario de Linux ejecuta cada orden de Bash dentro de `bwrap` (V12), y ese
aislamiento es incompatible con dos piezas del contrato del job, comprobado en V61 (3). **La traza**: `bwrap` siempre
desune el espacio de nombres de PID, y los `clone` de los procesos de la orden devuelven los números de ese espacio, no
los de los ficheros `t.<n>` de `strace -ff`, así que la atribución por la línea que crea cada fichero (D12; data-model §9,
reglas 1 y 2) no es posible: falla con una coincidencia de números o deja ficheros sin origen. **El contexto de la
skill** (FR-077): monta `/` de solo lectura, con escritura solo bajo `/home`, `/root`, `/tmp`, `/var`, `/opt`, `/run` y
`/mnt`, lo mismo que D16 (2) rechaza del sandbox de Bash. El proxy, la caché bajo `RUNNER_TEMP` y el enlace de la skill
sí funcionan bajo `bwrap`. El riesgo que se asume (plan.md §VII) es que una orden de la sesión lea el token del entorno
de un proceso antecesor: el runner es desechable y el job tiene `contents: read`; solo lo disparan quien pone una
etiqueta en una propuesta de cambio del propio repositorio (GitHub no pasa los secretos a las de un *fork*), la ejecución
manual o la semanal sobre la rama principal; las sesiones leen las evals del repositorio y las respuestas grabadas de la
fuente; y el token de `claude setup-token` se revoca y se sustituye sin tocar el repositorio. El valor va explícito, no
ausente, para que un valor por defecto del binario no active el aislamiento.

*Alternativas rechazadas.*
- *`CLAUDE_CODE_SUBPROCESS_ENV_SCRUB=1` con `bubblewrap` en el runner y `--allowedTools Skill Bash`* (el arreglo que
  proponía la línea de T036 tras la ejecución 34930222593): además de romper la traza y el contexto de la skill, exige
  `socat` (V61 (3)) y espacios de nombres de usuario sin privilegios en el runner, que Ubuntu 24.04 puede restringir con
  AppArmor (`kernel.apparmor_restrict_unprivileged_userns`: el paquete `apparmor` de 24.04 trae el perfil
  `unprivileged_userns`, que niega las capacidades, y ninguno para `bwrap`), algo que no se puede comprobar fuera del
  runner. Adaptar la traza con `strace --decode-pids=pidns` cambiaría las formas de línea y la atribución de data-model
  §9 justo antes del último intento de la prueba de red, y la skill seguiría ejecutándose con el disco de solo lectura.
- *Ajuste `processWrapper` de Claude Code* para anteponer `env -u CLAUDE_CODE_OAUTH_TOKEN` a cada orden: sin documentar
  y, según el binario, solo honrado desde ajustes gestionados; tampoco retiraría el token del entorno de los antecesores.
- *Pasar el token por descriptor* (`CLAUDE_CODE_OAUTH_TOKEN_FILE_DESCRIPTOR`, que existe en el binario 2.1.270): sin
  documentar, y el secreto seguiría en el entorno del paso del job, de `make` y del guion, legible igual en `/proc`.
- *Sandbox de Bash de Claude Code*: rechazado por D16.

## D14 · Caché preparada por sesión y comprobación sin red

**Decisión.** `Preparar` llena una caché nueva con **todas** las consultas necesarias de todas las evals, en proceso,
con `boe` sobre `httpx.Replay` de la unión H4 + H5; `ComprobarSinRed` las ejecuta con `--offline` sobre un registro cuyo
cliente es `Replay` de un directorio vacío. En `make ci`, una caché por test; en el job, una por sesión, justo antes.

**Por qué.** El binario no expone la reproducción (clarificación); componerla en proceso no toca ni la fuente ni el
kernel (V26). `metadatos` y `buscar` caducan a los 300 s (V22) y lo caducado es ausencia con `--offline` (V23): preparar
una vez por ejecución dejaría sin metadatos a las sesiones posteriores. Llenarla con todo evita que una eval dependa del
orden. El `Replay` de directorio vacío hace que una petición escapada falle en vez de salir a la red.

**Alternativas rechazadas.** *Una caché versionada*: caduca (clarificación). *Una caché por ejecución del job*: V22.
*Solo las consultas de la propia eval*: igual de válida, pero dos preparaciones distintas para `make ci` y el job.

**Sin contexto** (T006, al implementar). `Preparar`, `ComprobarSinRed` y `PrepararSesion` no reciben un
`context.Context`: cada consulta es una invocación entera de `app.Main`, la raíz de composición del binario, que abre su
propio contexto con el plazo de `--timeout` y no admite el de quien llama (V59). Un contexto en la firma solo se podría
mirar entre dos invocaciones y nunca llegaría dentro de ninguna: una cancelación a medias, que es lo que `contextcheck`
rechaza en `make ci`. *Alternativas rechazadas*: un `//nolint:contextcheck`, que el plan no admite (obligación 5) y que
taparía una cancelación que no llega; llamar a `app.Main` a través de una variable o de un campo, que solo esconde la
llamada al análisis; cambiar `app.Main` para que reciba el contexto, que es cambiar el kernel (obligación 6; spec,
*Fuera de alcance*); y ejecutar los verbos del applet sin `app.Main`, que dejaría fuera la lectura de banderas,
`--offline` y la traducción a códigos de salida que la preparación y la comprobación tienen que ejercer tal cual (V26).

## D15 · Pruebas de `make install` sobre un árbol mínimo

**Decisión.** `TestInstalacion` (etiqueta `integration`) copia en `$WORK/repo` `Makefile`, `go.mod`, `go.sum`, el
guion, `skills/` con sus enlaces y los ficheros de los paquetes de `./cmd/kitlegal` según `go list -deps`, y ejecuta
`make install` con `HOME`, `GOBIN` y `GOPATH` temporales, `GOPROXY=off` y `GOENV=off` (contrato de instalación §4).

**Por qué.** FR-055 prohíbe escribir en el directorio personal y en el de binarios reales; ejecutar `make install` sobre
el repositorio real reescribiría además su `bin/instalado/kitlegal` hacia un temporal que desaparece, rompiendo la
instalación de quien ejecuta los tests. El árbol mínimo ejerce la receta real. `GOMODCACHE` y `GOCACHE` se toman del
proceso de test para no descargar nada (misma práctica que el e2e de `internal/app`).

**Alternativas rechazadas.** *Probar solo el guion*: no ejerce `make install` (FR-055). *Clon con `git clone`*: ve lo
commiteado y no el árbol de la tarea. *Variable de raíz en el `Makefile` para los tests*: costura solo de test en la
superficie de invocación.

## D16 · Garantía de red del job

**Decisión.** La sesión corre con `HTTP_PROXY`/`HTTPS_PROXY` (y en minúsculas) apuntando a `127.0.0.1:9`, cerrado (el
guion lo comprueba), y `NO_PROXY=api.anthropic.com`; el sandbox de Claude Code se desactiva explícitamente. Cada
invocación del binario pasa por ese proxy y termina sin texto con 5, o con 4 bajo `--offline` (V41; en Linux y bajo
`strace`, V53). Independientemente, la traza registra cada `connect` de cada hilo del proceso de cada invocación,
atribuido como dice D12: una conexión a una dirección que no es local con resultado `0` o `EINPROGRESS` es una petición
llegada a la red y hace fallar el veredicto (FR-076).

**Por qué.** Bloquea con un mecanismo que el binario ya respeta sin cambiarlo (V19, V20) y que no depende de los flags del
agente (FR-074); la red del modelo sigue abierta porque el cliente de Claude Code respeta `NO_PROXY` (V10). La detección
no confía en el mecanismo: observa en el sistema al único proceso que habla con las fuentes, así que una evasión
(p. ej. una orden que vacía las variables de proxy) no pasa inadvertida. La evaluación ocurre en el contexto normal de la
skill instalada (FR-077).

**Alternativas rechazadas.**
- *Sandbox de Bash de Claude Code* (V11): bloqueo más fuerte, pero (1) exige espacios de nombres de usuario sin
  privilegios, `bwrap` y `socat` en el runner, que no se pueden comprobar en local; (2) cambia el contexto de ejecución de
  la skill (disco de solo lectura salvo el directorio de trabajo, `TMPDIR`, caché obligada a vivir en el directorio de
  trabajo); y (3) la petición permitida la haría el proxy dentro del proceso de Claude Code, invisible para la observación
  por invocación, de modo que la detección de FR-076 dependería de fiarse de la decisión del propio proxy.
- *Cortafuegos por usuario y proxy con lista de permitidos*: dos componentes más y listas de direcciones del proveedor
  del modelo, que cambian.
- *Leer `--offline` del entorno en el binario*: cambio del kernel prohibido por el spec.

## D17 · Sin Python en el job y en la aceptación

**Decisión.** Tras `make install`, el paso «Retirar Python del runner» (contrato del job §1), que fija sus opciones con
`set -euo pipefail` como primera orden, no descarta ningún error y no consulta ni purga paquetes: (1) busca como root en
**todo** el sistema de ficheros, salvo `/proc` y `/sys`, los ficheros ejecutables y los enlaces cuyo nombre empieza por
`python` o `pypy`, y los ficheros y enlaces `libpython*` y `libpypy*`, sin distinguir mayúsculas; (2) retira cada ruta
encontrada y, si está en `<prefijo>/bin/` y `<prefijo>/lib/` tiene una entrada `python*` o `pypy*`, la instalación
`<prefijo>` entera, salvo que contenga algo de lo que el job usa después (el espacio de trabajo, `~/.claude`, el binario
instalado y las órdenes `bash`, `sudo`, `find`, `rm`, `timeout`, `strace`, `claude`, `node`, `go`, `make` y `git`, por
su ruta y por su destino); y (3) repite la búsqueda y falla si queda alguna ruta o si falta algo de lo usado.
`scripts/evals.sh` repite la misma búsqueda, como root, antes de la primera sesión (comprobación 3 del contrato del job
§3.1): escribe en `sin-python.txt` la orden, el usuario y el resultado, que van al informe, y termina con 1 nombrando
cada ruta encontrada, o si no puede buscar como root. La aceptación (FR-080, FR-081) se registra de la ejecución de
cierre del job. Paso y comprobación están verificados en su forma literal en contenedores de `ubuntu:24.04` sin red
(V56 y, sin la purga de paquetes que el paso tuvo hasta T033, V60); lo que suponen del runner real son los supuestos S2
y S7.

Por qué las opciones las fija el propio paso: sin `-e`, una búsqueda que termina con error no detiene el paso, que sigue
retirando lo que ve y escribe `búsqueda tras retirar: ninguno` con código 0 (V60 (5), el mismo paso sin su línea `set`).
Con la línea en el paso, la garantía no depende de las opciones con que GitHub invoque `shell: bash`, que no se pueden
comprobar sin la plataforma, y ejecutado por otro intérprete el paso falla en esa misma línea sin hacer nada (V60 (5),
con `dash`). `-u` lo detiene también si falta una variable, como `GITHUB_WORKSPACE`, en lugar de seguir con una ruta
vacía entre lo usado; `pipefail`, como en `scripts/`, hace que una tubería que se añada después no esconda el fallo de
una etapa. «Instalar strace y Claude Code», el otro paso de más de una orden, empieza por la misma línea.

Por qué no se consultan ni se purgan paquetes: hasta T033 el paso purgaba antes de buscar, con `apt-get purge -y
--auto-remove`, los paquetes `python*`, `libpython*`, `pypy*` y `libpypy*` con ficheros en disco, listados con
`dpkg-query -W` sin patrón (V57). En la prueba de red del intento 1 de T030 (ejecución 34922606273,
`gates/prueba-de-red.md`) esa purga terminó con 100 en `ubuntu-24.04` antes de buscar nada: el filtro eligió 110
paquetes, entre ellos `python3`, `python3-minimal`, `python3-apt`, `python3-debconf` y `python3-netplan`, de los que
dependen paquetes del sistema, y retirarlos con sus dependientes llegaba a la cadena de arranque (`shim-signed`,
`grub-efi-amd64-signed`, `grub2-common`), que el resolvedor rechazó con «pkgProblemResolver::Resolve generated breaks».
V57 no lo vio porque en su contenedor ningún paquete del sistema dependía de Python; V60 (1) lo reproduce con un paquete
esencial que depende de dos de Python. La purga no retiraba ningún intérprete que la búsqueda no encuentre: el
intérprete, sus enlaces y sus bibliotecas se encuentran por nombre (V56 (1), V60 (1) y (2)), y lo que un paquete deja en
disco sin ellos —la biblioteca estándar, `dist-packages`, guiones con `#!/usr/bin/python3`— no se ejecuta. Sin purga, el
paso tampoco depende del grafo de paquetes de cada versión de la imagen, el mismo motivo por el que se rechazan las
listas de sitios. Lo que cambia: los paquetes siguen registrados en dpkg y pierden los ficheros que la búsqueda
encuentra, entre ellos los de su registro con esos nombres (las listas y sumas de los `libpython*` y los guiones de
mantenimiento ejecutables de los `python*`, V60 (1)); ningún paso posterior del job usa dpkg ni apt, y el runner es
desechable.

Por qué la búsqueda es así:
- **Todo el sistema de ficheros**: un intérprete fuera de los sitios de una lista (la caché de herramientas, Miniconda,
  un entorno de pipx, un SDK que trae el suyo) se ejecuta igual por su ruta absoluta, y la sesión corre con
  `bypassPermissions`. Solo se podan `/proc` y `/sys`, en los que el núcleo no deja crear ficheros y cuyas entradas
  cambian mientras se recorren; `/dev` se recorre, porque `/dev/shm` es un tmpfs que admite ejecutables, y como root lo
  hace sin errores (V56).
- **Como root**: la sesión corre con el usuario del runner, que tiene `sudo` sin contraseña (S2), así que todo lo que hay
  en disco le es accesible. Un usuario sin privilegios no puede leer, p. ej., `/root`: `find` termina con 1 y la búsqueda
  no sería completa (V56). Por eso la comprobación 3 falla si no puede buscar como root, en lugar de escribir «ninguno».
- **Ejecutables, enlaces y bibliotecas**: un intérprete es un ejecutable o un enlace a uno; las bibliotecas se buscan con
  cualquier modo porque se cargan sin permiso de ejecución, y con una de ellas un programa que la enlaza, como uno
  congelado que la lleva al lado, ejecuta Python. Un fichero de datos con ese nombre (`python.vim`, una página de
  manual, un `python.go` de solo lectura de la caché de módulos de Go) no ejecuta nada, y ni se busca ni se retira (V56).
- **Por nombre, y qué afirma**: la comprobación afirma exactamente lo que buscó, y lo escribe en `sin-python.txt` con la
  orden; un intérprete cuyo nombre no empieza por `python` ni por `pypy`, o sin permiso de ejecución, fuera de una
  instalación retirada, no lo ve, y que la imagen no traiga ninguno así es parte del supuesto S7.

Por qué la instalación entera: lo que queda de una instalación sin su intérprete no se ejecuta, pero retirarla entera no
deja en disco lo que la búsqueda por nombre no ve dentro de ella (su biblioteca estándar, un ejecutable con otro nombre
que lleva el intérprete). `<prefijo>/lib/python*` o `<prefijo>/lib/pypy*` distingue una instalación de Python (CPython,
PyPy, conda, un entorno virtual) de un prefijo que solo tiene un `python3` suelto en `bin/`, del que se retira solo el
fichero. Un prefijo que contiene algo de lo que el job usa —`/usr`, `/usr/local`, el espacio de trabajo, la instalación
de `claude`— tampoco se retira entero: lo usado se deriva de las propias órdenes del job, por su ruta y por su destino,
no de una lista de directorios, y la comprobación final del paso falla si la retirada se hubiera llevado algo de ello.

**Por qué.** FR-073 y FR-081 piden que la sesión no tenga Python accesible y que conste cómo se comprobó; el runner es
desechable. En macOS no se puede retirar `/usr/bin/python3` del sistema, así que una sesión local no podría cumplirlo; las
sesiones del job son sesiones de Claude Code con la skill y el binario instalados con `make install` (FR-080).

**Alternativas rechazadas.**
- *`PATH` recortado*: una ruta absoluta lo esquiva.
- *Buscar y retirar solo en el `PATH` y en `/usr/bin`, `/usr/local/bin` y `/bin`*: la misma debilidad con otra lista
  cerrada de sitios. Un intérprete de la caché de herramientas (`/opt/hostedtoolcache/Python`,
  `/opt/hostedtoolcache/PyPy`), de Miniconda (`/usr/share/miniconda`) o de un entorno de pipx seguiría en disco, la
  sesión lo ejecutaría por su ruta absoluta y `sin-python.txt` diría «sin Python» en falso; V56 encuentra y retira todos
  esos casos con la búsqueda completa.
- *Retirar una lista de instalaciones conocidas* (`/opt/hostedtoolcache/Python`, `/opt/hostedtoolcache/PyPy`,
  `/usr/share/miniconda`): depende de lo que traiga cada versión de la imagen, y deja fuera lo que no esté en la lista.
- *Podar también `/dev`*: `/dev/shm` admite ejecutables, y recorrer `/dev` como root no da errores (V56).
- *Buscar por contenido* (los símbolos de la biblioteca de Python en cada ejecutable y cada biblioteca): obliga a leer
  enteros todos los ficheros de la imagen, y lo que un programa que enlaza la biblioteca puede ejecutar ya se corta al
  retirar la biblioteca, que la búsqueda por nombre encuentra.
- *Cualquier fichero cuyo nombre empiece por `python`*: toma por intérpretes ficheros de datos que no ejecutan nada
  (V56) y, borrados de la caché de módulos de Go, la dejaría inconsistente.
- *Retirar el prefijo de todo `bin/` que tenga un `python*`, sin más condición*: con el `python3` de `/usr/bin` o de
  `/usr/local/bin` se llevaría el sistema o `claude`, y con uno suelto en el `bin/` de otra herramienta, la herramienta.
- *Fallar, en lugar de retirar solo el fichero, cuando el prefijo contiene algo usado*: `/usr` y `/usr/local` contienen
  siempre `bash` o `claude`, y un `python3` en ellos es lo normal; retirado el fichero, decide la búsqueda final.
- *Confiar en las opciones con que GitHub ejecuta un paso `shell: bash`*: no se pueden comprobar sin la plataforma, y
  el mismo paso sin `set -e` sigue tras una búsqueda fallida y termina con 0 (V60 (5); con la purga, V56 (7)).
- *Purgar antes los paquetes de Python con `apt-get purge -y --auto-remove`* (el paso hasta T033, con la consulta sin
  patrón y el filtro por estado de V57): en `ubuntu-24.04` termina con 100 sin retirar nada, porque de esos paquetes
  dependen paquetes del sistema (ejecución 34922606273 de T030; V60 (1)). Forzarla (`--allow-remove-essential`, o
  aceptar lo que el resolvedor arrastre) se llevaría lo que depende de Python, hasta la cadena de arranque, distinto en
  cada versión de la imagen; y un paquete retenido la detendría igual (V57 (4)).
- *`dpkg --purge --force-depends`*: no pasa por el resolvedor, pero ejecuta los guiones de mantenimiento de cada
  paquete, y los de Python necesitan el intérprete que se retira: el `prerm` de `python3` ejecuta `py3clean`, un guion
  que empieza por `#! /usr/bin/python3`. Retirado antes el intérprete, falla con 127 y deja `python3` a medias (`pF`);
  en el orden contrario termina con 0 (V60 (7)): el resultado depende del orden y de los guiones de cada versión.
- *Retirar los ficheros de las listas de dpkg* (`dpkg-query -L` de cada paquete de Python, sin `apt` ni guiones): no
  retira ningún intérprete que la búsqueda no encuentre, que es lo que FR-073 y FR-081 piden, y añade como supuestos el
  formato de esas listas, sus desvíos (`dpkg-divert`) y los directorios que comparten con otros paquetes.
- *Contenedor*: D13. *Sesión local*: no cumple FR-081 en macOS.

## D18 · `.agents/.gitattributes` y pendientes de estructura

**Decisión.** Un `.agents/.gitattributes` con `skills/** linguist-vendored linguist-generated`. Se borran de
`docs/PENDIENTES.md` las tres entradas «En H5».

**Por qué.** FR-085 pide los dos efectos solo sobre `.agents/skills/`. Git aplica los atributos exactamente ahí (V33). El
fichero va anidado porque el guardián solo admite rutas declarables y `.gitattributes` en la raíz no lo es (V35); cambiar
el workflow es proceso, fuera de la rama del hito (spec, *Fuera de alcance*). Se ponen los dos atributos porque FR-085 y la
clarificación los asocian a los dos efectos; que GitHub los aplique así, también desde un fichero anidado, es el supuesto
S8.

**Alternativas rechazadas.** *Solo `linguist-vendored`*: la clarificación señala que el plegado de diffs es de
`linguist-generated`. *`-diff`*: afecta a `git diff` local. *Desversionar `.agents/skills/`*: la clarificación lo
descarta.

## D19 · Documentación

**Decisión.** `README.md`: sección de skills con los tres directorios, `make install`, `make skills-sync`,
`make skills-check`, `make evals`, el formato común de eval (dónde viven las evals de cada skill, sus campos `pregunta`,
`activa`, `comandos`, `citas` y `reproduce`, y su validación con `eval.yaml.json` en `make ci`) y el job de evals
(manual, semanal y por etiqueta sobre la rama de un hito, lanzado con `make evals`); «Qué entrega este hito» pasa a H5; y la línea 188, que dice que `make help` enumera
también las órdenes que «reciben su contenido en un hito posterior (`skills-sync`, `release`)», pasa a nombrar solo
`release`, porque desde H5 `skills-sync` tiene contenido. En «Ejecutar los controles», «encadena los nueve controles»
(línea 156) pasa a «encadena los diez controles», porque T017 añade `skills-check` a los nueve prerrequisitos de `ci`; la
tabla de órdenes con la columna «¿En `ci`?» (líneas 166-180) recibe las filas de `make skills-check` (sí: frontmatter y
límite de líneas, derivas de las referencias, de la tabla de comandos y de los enlaces, tabla de normas contra su
esquema y sus identificadores, formato y conjunto de evals y lo grabado), `make skills-sync` (no: escribe en el árbol) y
`make evals` (no: sesiones con modelo que lanza el job de evals), y la de `make test-integration` nombra también la
instalación de las skills en un directorio personal temporal (T019). En «Construir e instalar» (línea 114), el
comentario de `make install` dice que, además de instalar el binario, enlaza las skills en el directorio personal de
skills de Claude Code. `CONTRIBUTING.md`: formato común de eval y job de evals; en la tabla «Los controles» (líneas
94-112), las filas de `make skills-check` (sí), `make skills-sync` (no — escribe en el árbol) y `make evals` (no — sesiones
con modelo y credencial, fuera de `make ci`; las lanza el job de evals), y la fila de los tests con la etiqueta
`integration` nombra la instalación de las skills; se retira la fila `make skills-sync` de la tabla de «Órdenes que existen
pero reciben su contenido en un hito posterior» (líneas 156-163), que queda solo con `make release`; y el párrafo que
enumera las órdenes que salieron de esa tabla (líneas 168-172) añade «desde H5, `make skills-sync`» regenera las
referencias y la región de comandos de las skills. `CHANGELOG.md` (*Unreleased*): el párrafo de introducción, que
enumera los hitos cerrados hasta H4, añade **H5 — skill `boe-legislacion`**; *Añadido* skill, `data/normas.yaml`,
esquemas, formato común de eval, evals, job de evals, órdenes; *Cambiado* `make install` y `make ci` (diez controles, con
`skills-check`). Los tres documentos escriben literalmente «formato común de eval» y «job de evals», y quickstart §11
busca esas expresiones, la cifra, las filas y sus valores de «¿En `ci`?», el comentario de `make install`, el párrafo de
`CONTRIBUTING.md` y la entrada de H5 en la introducción del `CHANGELOG.md`.

**Por qué.** FR-083; y la documentación debe coincidir con el `Makefile` en la misma rama: no basta con nombrar las
órdenes; un README que diga «nueve controles» u omita `skills-check` de su tabla contradice el `ci:` que deja T017.

## D20 · Lint

**Decisión.** `.golangci.yml` añade `evals` a `run.build-tags` y, a `misspell.ignore-rules`, exactamente estas cinco
palabras (en el paso 2 del orden de implementación), cada una con un comentario que da su motivo, como las catorce de
H0-H4:

| Palabra | Por qué la escribe suelta el código de H5 | `misspell` la lee como (V42) |
|---|---|---|
| `comando` | singular del formato: «comando esperado» y «comando ausente» en los comentarios y en los motivos de `Juzgar` y del informe (data-model §6.1 y §10.2) | «commando» |
| `comandos` | clave `comandos` del formato de eval (FR-060, contrato de evals §1) en la etiqueta YAML del campo `Comandos` de `Eval`, y claves `comandos_ejecutados` y `comandos_ausentes` del informe (contrato del job §5) | «commandos» |
| `defectos` | «defectos de la skill»: lo que `Regenerar` devuelve junto a las derivas y lo que `make skills-check` presenta (data-model §5, contrato de sincronización §2) | «defects» |
| `legislativo` | «Real Decreto Legislativo», rango que la expresión de `normas-nombradas` tiene que casar tal cual en `SKILL.md` (contrato de sincronización §2) | «legislation» |
| `patrones` | los `pattern` de `NORMA` y `BLOQUE` de `schemas/normas.yaml.json` y `schemas/eval.yaml.json`, que `TestGramaticasCoincidenConBoe` compara con `boe.ValidarNorma` y `boe.ValidarBloque` y que sus mensajes nombran (data-model, cabecera; control 10) | «patrons» |

Las otras veintiuna palabras que V42 encuentra en los artefactos y en spec.md no entran, y el código no las escribe
sueltas (plan.md, obligación 4): las que en español llevan tilde (`activacion`, `constitucion`, `declaracion`,
`evaluacion`, `preparacion`, `prescripcion`) van con ella o dentro de un identificador camelCase; las demás
(`candidatas`, `componentes`, `contradice`, `decisiones`, `definitivo`, `directorios`, `distribuye`, `informativo`,
`procede`, `producto`, `programas`, `recorre`, `secretos`, `terminaron`, `variantes`), con otra palabra que `misspell` no
marca o dentro de un identificador: `recorre`, con `recorrer` o `recorrido`; `terminaron`, con «sin terminar» o «que no
acabaron»; `candidatas`, con el singular `candidata` o con «posibles» (alternativas comprobadas en V42). Nada de
`//nolint`. Los tests que ejecutan programas (`make`, `go`) escriben el programa y los argumentos como constantes y pasan
lo variable por `cmd.Env` y `cmd.Dir`. Lecturas y escrituras de un mismo fichero en funciones distintas.

**Por qué.** Sin la etiqueta, `job_test.go` quedaría fuera del lint (V38). `.golangci.yml` neutraliza el español
«falso positivo a falso positivo»: cada entrada responde a una palabra que el código tiene que escribir suelta, y lo dice
en su comentario. Las cinco lo son por contrato: esquivarlas cambiaría la clave del formato que fija el spec, el término
del modelo de datos o el texto que una expresión tiene que casar. Las otras veintiuna son prosa de los artefactos o del
spec, o palabras que en español llevan tilde; ninguna regla las exige literales en Go, así que una entrada para ellas no
respondería a ningún falso positivo del código. `recorre`, con la que los artefactos describen el lector común, se
redacta en Go con `recorrer` o `recorrido`. `terminaron`, con la que FR-071 y *Key Entities* describen las invocaciones
fuera de lo grabado, no es ninguna clave del informe —la clave es `fuera_de_lo_grabado` (data-model §10.2)—, así que los
comentarios de `juzgar.go` e `informe.go` dicen «sin terminar». `candidatas`, la de FR-005, no nombra nada del código: un
comentario que la necesite usa el singular. `misspell` no marca ninguna de esas alternativas (V42). `gosec` analiza
también los tests (`run.tests: true`).

**Alternativas rechazadas.** *Solo `comando` y `comandos`*: el campo `Defectos`, la expresión con «Real Decreto
Legislativo» y los mensajes sobre los patrones harían fallar el lint en las tareas que los escriben. *Ignorar también
`recorre`, `terminaron` y las demás de V42*: entradas sin ninguna palabra del código que las pida. *Esquivar las cinco*: renombraría la
clave `comandos` del spec, o escribiría la expresión de `normas-nombradas` con clases de caracteres (`[Ll]egislativo`)
solo para no tocar la lista.

## D21 · Aplicación de las skills `golang-*`

`golang-how-to` orquesta: `golang-project-layout` (D2: sin `package main` fuera de `cmd/`; código en `internal/`),
`golang-testing` (tests de tabla con subtests con nombre, `t.Parallel()` salvo con `t.Setenv`, un fichero de test por
fichero de código con las excepciones declaradas en plan.md, y la etiqueta `integration` para `TestInstalacion`, que
ejecuta `make` y compila), `golang-lint` (sin `//nolint`; si alguno fuera inevitable, con linter y motivo, que
`nolintlint` exige), `golang-continuous-integration` (acciones con versión mayor fijada, permisos mínimos por job,
`-count=1` en las comprobaciones) y `golang-dependency-management` (una sola dependencia directa nueva, ya presente en el
grafo). Reglas citadas en V39.

## D22 · Supuestos no verificados

| # | Supuesto | Quién lo comprueba |
|---|---|---|
| S1 | En GitHub, `pull_request` con `types: [labeled]` sobre una propuesta de cambio del mismo repositorio ejecuta el fichero del job de la rama de la propuesta, con los secretos del repositorio y `github.event.label.name`; `schedule` corre en la rama principal; el contexto `inputs` está vacío fuera de `workflow_dispatch` | tarea `[plataforma]`, prueba de red |
| S2 | `ubuntu-24.04` tiene `sudo` sin contraseña (también con `sudo -n`, que usa la comprobación 3 del contrato del job §3.1), `apt` con `strace`, `node` y `npm`, `timeout` y las mismas GNU findutils y coreutils de la imagen `ubuntu:24.04` que comprueban V56 y V60 (`find` con `-perm /111`, `-iname`, `-H` y `-quit`; `readlink -e`); y `npm install -g @anthropic-ai/claude-code@2.1.270` instala la versión comprobada. El paso de retirada ya no usa dpkg ni apt (D17): lo que este supuesto decía de ellos lo desmintió la prueba de red del intento 1 de T030 (ejecución 34922606273), en la que `dpkg-query -W` sin patrón terminó con 0, con `:amd64` en los paquetes `Multi-Arch: same`, pero `apt-get purge -y --auto-remove` no retiró lo que elegía el filtro y terminó con 100. En esa misma ejecución, `sudo apt-get` no pidió contraseña, `strace` 6.8 ya venía en la imagen y `claude --version` dio `2.1.270 (Claude Code)`. El job no instala `bubblewrap` ni `socat`: con `CLAUDE_CODE_SUBPROCESS_ENV_SCRUB=0` (D13, desde T036) Claude Code no los usa, y este supuesto no dice nada de los espacios de nombres de usuario del runner (V61) | tarea `[plataforma]`, en la prueba de red: registra en `gates/prueba-de-red.md`, de la salida del paso de retirada que imprime la sexta orden de quickstart §12.2, la línea `búsqueda:` seguida de las líneas `retirado:` y de `búsqueda tras retirar: ninguno`, que muestran que `sudo` y `find` terminaron, junto a la evidencia de S7. Si el paso se detiene en la búsqueda o en un borrado, la ejecución no llega al informe: la tarea registra la salida de la orden de `--log-failed` de quickstart §12.2, se detiene y lo anota como en S7 |
| S3 | Cada búsqueda del manifiesto devuelve su norma entre sus diez resultados; existen los bloques `a66`, `a59`, `a17`, `a25`, `a20`, `a140` y `a38`; los rangos de un real decreto legislativo y de la Constitución son los que devuelva la búsqueda | la persona, en la pausa de grabación |
| S4 | Formato de `strace -ff -o` **en el runner** (x86_64, con Claude Code), en lo que V53 y V54 no cubren porque se comprobaron en arm64 con `bash` y el binario: la línea `clone` de los hilos de Go en amd64, que además llevan `CLONE_SETTLS` (V51) y, con él, previsiblemente un argumento más; las líneas con las que Claude Code crea sus hilos y los procesos de sus órdenes (`clone3` con `{flags=…}`, `clone`, `fork` o `vfork`), la de `+++ killed by SIG… +++` y las de señal `--- SIGNOMBRE {…} ---` que produzca Claude Code; que las formas de V53 y V54 son las mismas (`execve` con su argv, `connect` en `AF_INET`, `AF_INET6` y `AF_UNIX`, `clone` con `CLONE_THREAD`, un fichero por hilo con su línea final y sin `<unfinished …>`); y que `strace` puede seguir a Claude Code y a sus descendientes | tarea `[plataforma]`, en la prueba de red: registra en `gates/prueba-de-red.md`, del informe copiado del registro, la conexión `127.0.0.1:9` de clase `local` de la invocación `a9998 --json` de la sesión `01-lpac-articulo-21-prueba-de-red` (contrato del job §5 y §6) y que ninguna sesión no cortada por el tope (`codigo_de_la_sesion` distinto de 124 y 137, como el 0 de las terminadas) tiene traza ilegible, como evidencia de que las líneas `connect`, las de creación de hilos y procesos, las de señal y la atribución por hilos funcionan con trazas reales. La conexión prueba la lectura de `connect` y su atribución a la invocación; la legibilidad de las sesiones no cortadas, que cada fichero de hilo tiene su línea de creación reconocida y su línea final (D12). Una sesión cortada por el tope no es evidencia de S4 ni en contra: su traza puede quedar sin líneas finales o con una llamada interrumpida (V54), y se anota como S9. Si falta esa conexión o una sesión no cortada es ilegible por su traza, la tarea se detiene y lo anota con el fichero, la línea y el texto que da el motivo; esa línea real entra en las sesiones sintéticas en una tarea `[datos]` y `LeerTrazas` se ajusta antes de la ejecución de cierre (plan.md, obligación 1) |
| S5 | En `-p`, Claude Code carga `~/.claude/skills` y el modelo activa la skill con `Skill`; Bash hereda el entorno del proceso (proxy y `KITLEGAL_CACHE_DIR`) y corre sin sandbox, con `--settings` en línea y `CLAUDE_CODE_SUBPROCESS_ENV_SCRUB=0` (D13). V61 (4) lo comprueba en un contenedor arm64 sin modelo, con un servidor local en lugar de la API: `Skill` se ejecuta, y la orden de Bash ve el proxy y la caché, escribe en ella y corre en los espacios de nombres del guion, sin `bwrap`. Queda para la plataforma que el modelo active la skill y que el binario de x86_64 haga lo mismo | tarea `[plataforma]` |
| S6 | Con `CLAUDE_CODE_OAUTH_TOKEN` (token de suscripción de `claude setup-token`; el proyecto no usa clave de API de pago por uso), `claude -p` se autentica en el runner y acepta `--model claude-haiku-4-5-20251001`. Lo que este supuesto decía de `CLAUDE_CODE_SUBPROCESS_ENV_SCRUB` (que retiraba el token del entorno de las órdenes) lo sustituye D13 desde T036: con la variable a `0`, V61 (4) comprueba en un contenedor que Claude Code no pasa el token al entorno de las órdenes ni al de los ganchos, y que sigue legible en `/proc/<pid>/environ` de sus antecesores, el riesgo que asume plan.md §VII | tarea `[plataforma]` |
| S7 | Lo que la retirada de Python (contrato del job §1; D17) supone del runner `ubuntu-24.04`, fuera de lo que V56 y V60 comprueban en un contenedor. (1) **Lo que trae**: sus intérpretes y bibliotecas de Python —los de los paquetes del sistema, que en la ejecución 34922606273 de T030 eran 110 con nombre `python*` o `libpython*`, varios de ellos dependencias de paquetes del sistema, y, según la documentación pública de `actions/runner-images`, no comprobada, `/opt/hostedtoolcache/Python`, `/opt/hostedtoolcache/PyPy` y Miniconda en `/usr/share/miniconda`, además de lo que haya en otros sitios— tienen nombres que empiezan por `python`, `pypy`, `libpython` o `libpypy`, y cada intérprete tiene permiso de ejecución o está dentro de una instalación que se retira entera: la búsqueda afirma exactamente eso, y es lo que `sin-python.txt` registra. (2) **Que se puede buscar**: como root, `find` recorre todo el sistema de ficheros salvo `/proc` y `/sys` y termina con 0, sin entradas ilegibles para root ni que desaparezcan mientras se recorre. (3) **Que se puede retirar**: nada de lo encontrado está en un sistema de ficheros de solo lectura, como el de un snap. (4) **Que no rompe el job**: retirar lo encontrado —sin purgar paquetes, así que los que dependen de los de Python siguen instalados sin sus intérpretes ni sus bibliotecas, y el registro de dpkg de esos paquetes queda incompleto (V60 (1))— no afecta a `bash`, `sudo`, `find`, `rm`, `timeout`, `strace`, `claude`, `node`, `go`, `make`, `git` ni al binario instalado, y ningún paso posterior usa dpkg ni apt; el paso comprueba que lo usado sigue en disco, no que funcione. La ejecución 34922606273 no llegó a ejercer (2)-(4): el paso anterior se detuvo en la purga, que suponía además en (3) que ningún paquete de Python estaba retenido y en (4) que `apt` podía retirarlos con sus dependientes, y esto último no se cumplió | tarea `[plataforma]`, en la prueba de red: registra en `gates/prueba-de-red.md` la salida del paso de retirada que imprime la sexta orden de quickstart §12.2 (la búsqueda, cada `retirado:` y `búsqueda tras retirar: ninguno`), que muestra qué intérpretes, bibliotecas e instalaciones traía la imagen y que la búsqueda y los borrados terminaron; el `sin_python` del informe (la búsqueda, `usuario: root` y `resultado: ninguno`), que muestra que la comprobación 3 repitió la búsqueda completa como root antes de la primera sesión; y, para (4), que las sesiones ejecutaron `claude` y el binario con los códigos del informe. Si el paso o la comprobación 3 fallan —una ruta que queda, un borrado en solo lectura, una búsqueda con errores, algo usado que ya no está, Python accesible—, la ejecución no llega al informe: la tarea registra la salida de la orden de `--log-failed` de quickstart §12.2, se detiene y lo anota con la ruta y el mensaje, y el arreglo va en una tarea nueva antes de la ejecución de cierre (plan.md, obligación 1). Un intérprete con otro nombre, o sin permiso de ejecución, fuera de una instalación retirada, no lo detecta ninguna evidencia del job: (1) queda como supuesto |
| S8 | GitHub excluye de las estadísticas los ficheros con `linguist-vendored` y pliega en los diffs los que llevan `linguist-generated`, también desde un `.gitattributes` anidado | tras la fusión (estadísticas del repositorio) y en la primera propuesta de cambio que toque `.agents/skills/`; anotado en `gates/pr-h5.md` |
| S9 | Una sesión de Haiku con ≤ 30 turnos cabe en 240 s; `TestInstalacion` reutiliza la caché de construcción y no alarga `make ci` de forma apreciable | tarea `[plataforma]`: en la prueba de red y en la ejecución de cierre, cada sesión con `codigo_de_la_sesion` 124 o 137 se anota en `gates/prueba-de-red.md` o en `gates/evals-cierre.md` como evidencia en contra, con su motivo, también si su traza es ilegible, y no como evidencia de S4 (D12, V54); revisión final |
| S10 | En el runner, con Claude Code, `timeout --kill-after=10s 240s` termina `strace`, `claude` y sus descendientes, y los códigos de la sesión son los de V52 y V53 (las mismas versiones de coreutils y de `strace` de Ubuntu 24.04, en x86_64): 124 al agotar el tope, 137 si hay que enviar `KILL` y, si la sesión termina antes, el código de `claude`; y que lo que queda en la traza de una sesión cortada es lo de V54 (2) y (3) | tarea `[plataforma]`, en la prueba de red: registra en `gates/prueba-de-red.md` el `codigo_de_la_sesion` 0 de cada sesión terminada del informe, que comprueba en el runner que `strace` devuelve el código de `claude` cuando este termina bien. 124 y 137 no se provocan en la plataforma: los fijan V52, V53 y V54, y eligen el texto del motivo y si la traza se lee como la de una sesión cortada (data-model §9, regla 6), nunca si la sesión pasa, porque con cualquier código distinto de 0 la sesión no terminó (contrato del job §4) |
| S11 | Con el tráfico no esencial desactivado, Claude Code solo necesita `api.anthropic.com` | tarea `[plataforma]` |
| S12 | Lo que suponen las órdenes de quickstart §12 para identificar la ejecución de una etiqueta. (1) **Orden de los eventos**: la API de eventos de la incidencia devuelve los `labeled` en orden cronológico y, con `--paginate`, el último evento es la última línea. Las órdenes no dependen de ello: toman el último instante **ordenando** los `created_at` (`sort \| tail -n 1`), lo que solo exige (2). (2) **Formato de los instantes**: `created_at` de la API de eventos y `createdAt` de `gh run list` son instantes UTC con el mismo formato ISO 8601 (`AAAA-MM-DDTHH:MM:SSZ`), así que su orden de texto es el cronológico, tanto para `sort` como para el `>=` de `--jq`. (3) **Instante de la ejecución**: la ejecución de `evals.yml` que crea un evento `labeled` tiene `createdAt` igual o posterior al `created_at` de ese evento. (4) **Quitar una etiqueta que no está puesta**: `gh pr edit --remove-label` de una etiqueta existente que no está puesta termina en 0. Las órdenes tampoco dependen de ello: solo la quitan si `gh pr view --json labels` la lista, y un fallo de `gh pr view` detiene la orden (`set -e`). (5) **Quitar y volver a poner**: crea un evento `labeled` nuevo y, con S1, una ejecución nueva de `evals.yml`. (6) **Nombre del flujo de la ejecución**: `gh run list --json` da `workflowName` `evals` (el `name:` del contrato del job §1) a la ejecución de `evals.yml` que crea el evento, aunque el fichero solo esté en la rama de la propuesta de cambio. Las órdenes no dependen de que `gh` resuelva el flujo por el nombre de su fichero (`--workflow evals.yml`, que V14 no cubre): filtran por ese campo, que V45 comprueba que existe | tarea `[plataforma]`, en la prueba de red, antes de leer ningún informe: registra en `gates/prueba-de-red.md` la lista de eventos `labeled` de la etiqueta tal como la devuelve la API (1), el instante elegido y el `createdAt` y `databaseId` de la ejecución leída (2, 3), que esa ejecución no es ninguna de las anteriores a la etiqueta (5), y su `workflowName` con la lista `posteriores_a_la_etiqueta` que imprime la tercera orden de quickstart §12.2 (6). Si algo difiere, la tarea se detiene, lo anota en su `gates/tarea-Tnnn.md` y las órdenes de §12 se corrigen antes de la ejecución de cierre |
