# Research · H21 · `kitlegal mcp serve`

Fase 0 del plan. Modo desatendido: cada decisión se tomó con el «Criterio de decisión autónoma» de la constitución y
lleva su alternativa rechazada. Toda afirmación sobre una herramienta o dependencia externa remite a la tabla V
(comprobada en local, en esta sesión) o a la tabla S (supuesto que no se pudo comprobar sin red o sin abrir una sesión con
modelo). V43 a V45 son de la sesión del corrector del plan, que rehízo D8: leídas, con las órdenes que nombran, en el
repositorio y en la biblioteca estándar.

**Material de un solo uso, fuera de lo versionado** (en el directorio temporal de la sesión, borrado con ella):

- `h21-sdk/`: un módulo con `modelcontextprotocol/go-sdk` v1.7.0 y dos prototipos, un servidor por la entrada y la
  salida estándar con una herramienta de esquemas propios (`servidor/main.go`) y dos clientes contra él
  (`cliente_test.go`, el cliente del SDK; `anterior_test.go`, uno de la especificación 2025-11-25 escrito línea a línea).
- `h21-informe/`: el programa `jq` de `seccion_evals` de `scripts/workflow/informe.sh` copiado literal (`tasas.jq`), un
  `informe.json` sintético de dos modos, el guion `medir.sh` (tamaño de los esquemas de cada herramienta con el binario
  de `main`) y el texto de las `instructions`.
- Un worktree de `22b5bda` con el prototipo de los dos `SKILL.md`, retirado al terminar (`git worktree list` solo da el
  árbol del repositorio). Su diff está en [contracts/skills-prototipo.diff](./contracts/skills-prototipo.diff).

Ninguna de las tres cosas abrió una sesión con modelo ni leyó credenciales. La red usada fue la de las herramientas de Go
(el proxy de módulos), y solo en las órdenes que V1 y V2 nombran.

## V · Verificado en local, en esta sesión

| # | Afirmación | Dónde se comprobó |
|---|---|---|
| V1 | La caché de módulos del equipo tiene `go-sdk` v1.6.0 y v1.7.0, con sus sumas ya verificadas. El proxy publica además la v1.8.0 (2026-09-04), pero **no se pudo descargar**: verificarla exige escribir `~/go/pkg/sumdb/sum.golang.org/latest`, y el sandbox de las sesiones de un paso solo deja escribir en la caché de módulos, no en `sumdb` | `ls ~/go/pkg/mod/cache/download/github.com/modelcontextprotocol/go-sdk/@v/`; `go -C <temporal> list -m -versions github.com/modelcontextprotocol/go-sdk` → `verifying go.mod: …@v1.8.0/go.mod: open …/pkg/sumdb/sum.golang.org/latest: operation not permitted`; `scripts/claude-modelo.sh:120-128` (`allowWrite`: temporal, `GOCACHE`, `GOMODCACHE`, telemetría, golangci) |
| V2 | Con los `require` de `go.mod` de `main`, `go get github.com/modelcontextprotocol/go-sdk@v1.7.0` termina bien dentro del sandbox; `go mod tidy` falla después porque la selección mínima elige `golang.org/x/oauth2` v0.35.0 y `github.com/segmentio/asm` v1.1.3, que la caché no tiene verificadas (mismo error de `sumdb`). Con `go get …/go-sdk@v1.7.0 golang.org/x/oauth2@v0.37.0 github.com/segmentio/asm@v1.2.1`, `go mod tidy` termina bien sin escribir en `sumdb`, y los prototipos compilan y pasan con `-race` | las tres órdenes, en `h21-sdk/` |
| V3 | `go-sdk` v1.7.0 declara `go 1.25.0` y requiere `golang-jwt/jwt/v5`, `google/go-cmp`, `google/jsonschema-go` v0.4.3, `segmentio/encoding` v0.5.4, `yosida95/uritemplate/v3` v3.0.2, `x/oauth2`, `x/time`, `x/tools`; sin cgo | `go.mod` del módulo en la caché |
| V4 | Un binario que importa `…/go-sdk/mcp` enlaza, además de lo que ya enlaza kitlegal (`x/sys`, `x/time`), **siete módulos nuevos**: `modelcontextprotocol/go-sdk`, `google/jsonschema-go`, `segmentio/encoding`, `segmentio/asm`, `yosida95/uritemplate/v3`, `golang.org/x/oauth2` y `golang.org/x/sync`. No enlaza `golang-jwt` ni `go-cmp`. Medido en darwin/arm64 | `go list -deps -f '{{if .Module}}{{.Module.Path}} {{.Module.Version}}{{end}}' ./servidor` en `h21-sdk/` |
| V5 | El paquete `mcp` del SDK importa `net/http` (transporte HTTP en el mismo paquete), `golang.org/x/oauth2` y `…/jsonschema-go/jsonschema` | `go list -deps … github.com/modelcontextprotocol/go-sdk/mcp` |
| V6 | El SDK v1.7.0 atiende en el mismo proceso las versiones `2026-07-28`, `2025-11-25`, `2025-06-18`, `2025-03-26` y `2024-11-05`; el servidor no tiene opción para estrechar la lista salvo que el transporte implemente `ProtocolVersionSupporter` | `mcp/shared.go:45-63`; `mcp/server.go:912-927` |
| V7 | `(*Server).AddTool(*Tool, ToolHandler)` es la interfaz de bajo nivel: no valida la entrada ni la salida, exige un `InputSchema` no nulo de `type: object` (pánico si no) y acepta en `InputSchema` y `OutputSchema` cualquier valor que se serialice a JSON, también `json.RawMessage`; un nombre de herramienta admite letras, cifras, `_`, `-` y `.` hasta 128 caracteres. La función genérica `AddTool[In, Out]` valida la entrada con `jsonschema-go` y devuelve el fallo como texto propio, no como un resultado que el llamante componga | `mcp/server.go:250-321`, `:323-449`; `mcp/tool.go:17-57`, `:161-194` |
| V8 | Sin `Capabilities`, el servidor anuncia `{"logging":{}}` más lo que tenga registrado; con `Capabilities: &ServerCapabilities{Tools: &ToolCapabilities{}}` anuncia exactamente `{"tools":{}}`, sin `listChanged`, sin `logging`, sin `prompts` ni `resources` | `mcp/server.go:109-137`, `:615-663`; prototipo: `"capabilities":{"tools":{}}` en los dos clientes |
| V9 | `StdioTransport` usa `os.Stdin` y `os.Stdout` directamente; `IOTransport{Reader io.ReadCloser, Writer io.WriteCloser}` recibe los dos | `mcp/transport.go:112-138` |
| V10 | `(*Server).Run` vuelve cuando el cliente cierra la conexión o se cancela su contexto, con el error de la sesión. Cerrada la entrada **sin llamadas en curso**, devuelve `nil` | `mcp/server.go:1272-1312`; prototipo, `TestClienteModerno`: `Run devolvió: <nil>`, `exit status 0` |
| V11 | Cerrada la entrada **con una llamada en curso**, el SDK espera a que el manejador termine, no escribe su respuesta, y `Run` devuelve un error con el texto `server is closing: EOF`, que no envuelve `io.EOF` (`fmt.Errorf("%w: %v", ErrServerClosing, …)`) y cuyo centinela vive en un paquete `internal` | prototipo, `TestClienteAnterior`: salida vacía tras cerrar la entrada, `Run devolvió: server is closing: EOF`; `internal/jsonrpc2/conn.go:478-495`, `:541-559`, `:749-763` |
| V12 | Si todas las peticiones llegan de golpe y la entrada se cierra a continuación (un fichero por la entrada estándar), ninguna respuesta se escribe: mismo caso que V11. Un cliente tiene que esperar cada respuesta antes de cerrar | prototipo, segunda ejecución de `TestClienteAnterior` con las peticiones en un `strings.Reader`: salida estándar vacía |
| V13 | Una línea de la entrada que no es JSON termina la sesión: `Run` devuelve el error del analizador (`invalid character 'e' looking for beginning of value`) | prototipo, primera ejecución de `TestClienteAnterior` |
| V14 | El cliente del SDK negocia por omisión `2026-07-28` (con `server/discover`, sin `initialize`); la versión que pide es un campo sin exportar de `ClientSessionOptions` («for testing»): desde fuera del paquete no se le puede pedir otra | `mcp/client.go:249-253`, `:307-376`; prototipo: `"protocolVersion":"2026-07-28"` |
| V15 | Un cliente de la `2025-11-25` escrito a mano (`initialize`, `notifications/initialized`, `tools/list`, `tools/call`) recibe: `protocolVersion` `2025-11-25`, `capabilities` `{"tools":{}}`, las `instructions` y `serverInfo`; en `tools/list`, cada herramienta con `annotations` `{"idempotentHint":false,"readOnlyHint":true}` y sus dos esquemas **con los bytes que se le dieron**; en `tools/call`, `content[0].text` con el texto dado, `structuredContent` con el JSON dado e `isError: true` en el resultado de fallo. Cada línea de la salida estándar es un mensaje JSON-RPC 2.0 | prototipo, tercera ejecución de `TestClienteAnterior` (siete respuestas) |
| V16 | El codificador del SDK escapa `<`, `>` y `&` como `<`… en lo que escribe por el cable, también dentro de `structuredContent`; decodificado, el texto es idéntico (`a < b & «c» é`) | prototipo: línea del cable frente a `resultado.Content[0].(*mcp.TextContent).Text` |
| V17 | Una llamada a una herramienta que el servidor no tiene se responde como error del protocolo (`-32602`, `unknown tool "…"`), sin llegar a ningún manejador. `prompts/list` y `resources/list` se responden con listas vacías aunque la capacidad no se anuncie | prototipo; `mcp/server.go:955-962` |
| V18 | El servidor atiende las llamadas a la vez: cada petición corre en su gorrutina, y seis llamadas simultáneas (dos de 300 ms y cuatro inmediatas) reciben cada una su eco, las inmediatas antes; sin carreras con `-race` | `internal/jsonrpc2/conn.go:652-690`; prototipo, `go test -race` |
| V19 | Con un registrador, el SDK escribe en él líneas en inglés: a nivel `INFO`, el arranque, la conexión y una por petición de un cliente `2026-07-28` (`client log level set`); a nivel `ERROR`, `server session ended with error` cuando la sesión acaba como en V11 o V13. Sin registrador, las descarta | prototipo con `slog.LevelDebug` en la salida de error |
| V20 | El cliente MCP que lleva dentro Claude Code 2.1.284 valida el `structuredContent` de un resultado contra el `outputSchema` de la herramienta y rechaza la llamada si no casa (`Structured content does not match the tool's output schema`); una de sus dos variantes lo valida también cuando `isError` es verdadero | `grep -a -o` sobre `~/.local/share/claude/versions/2.1.284` (el método con el que H7.3 leyó V1-V5 de su research): `has an output schema but did not return structured content…` |
| V21 | Claude Code 2.1.284 declara `--mcp-config <configs...>` («Load MCP servers from JSON files or strings (space-separated)») y `--strict-mcp-config`, y compone el nombre de una herramienta MCP como `` `mcp__${serverName}__${toolName}` `` | `grep -a -o` sobre el mismo binario |
| V22 | El kernel pone a todo verbo el plazo de `--timeout` (`context.WithTimeout`), convierte el plazo vencido en `fuente-no-disponible` aunque el applet vuelva sin error, y presenta un sobre con lo que el applet devuelve: un verbo que sirve durante horas saldría con 4 y un sobre en la salida estándar | `internal/app/main.go:429-473`, `:532-543`; `internal/cli/sobre.go:91-115` |
| V23 | El kernel entrega al grafo **después** de presentar y dentro de `Emitir`, de forma síncrona; busca el aviso de versión en toda invocación que resuelve un applet distinto de `skills`; y `--dry-run` no corta antes del applet | `internal/cli/sobre.go:104-114`; `internal/app/main.go:352-358`, `:402-406`, `:455-464` |
| V24 | El terminador `--` funciona en el binario de hoy: lo que va detrás es posicional (`kitlegal territorio resolver --json -- Leganés` devuelve el sobre) y no nombra verbo ni bandera | la orden, con el binario de `main`; `internal/cli/parse_test.go:192`; `internal/app/despacho_test.go:245-248` |
| V25 | Los diez verbos de consulta solo tienen argumentos posicionales de cadena o de lista de cadenas; su `entrada` de `--describe` sin las ocho banderas ocupa entre 62 y 171 B en JSON compacto (1 190 B los diez) y ninguna referencia `$defs`; su `salida` con los `$defs` del documento, entre 1 247 y 3 661 B (19 289 B los diez); sus descripciones, 971 B | `h21-informe/medir.sh` sobre el binario de `22b5bda` |
| V26 | El ritmo por sitio y las reglas de `robots.txt` viven en cada `httpx.Cliente` (`sitios`), que es seguro para varias gorrutinas; `DependenciasDeRed` construye un cliente nuevo en cada invocación. H2 fijó que la copia de `robots.txt` «dura lo que vive el cliente» porque «un cliente vive una invocación» | `internal/httpx/sitio.go`; `internal/app/boe.go:42-52`; `specs/003-h2-internal-httpx-cliente/spec.md:31`, `:238` |
| V27 | `forbidigo` prohíbe `os.Exit`, `fmt.Print*`, `os.Stdout` y `os.Stderr` fuera de las raíces de composición; no dice nada de `os.Stdin`. `depguard` compara cada lista con la ruta absoluta del fichero y también linta los `_test.go` | `.golangci.yml:8-44`, `:78-215`, `:229-261`, `:378-410` |
| V28 | `TestDependenciasDelBinario` falla con cualquier módulo enlazado que no esté en `modulosDelBinario`, en seis plataformas; `TestArquitectura` comprueba R2 sobre las importaciones **directas** de los paquetes del módulo, así que un tercero que importe `net/http` no la incumple | `internal/arch_test.go:155-203`, `:243-279`, `:751-779` |
| V29 | `TestEsquemasCubrenTodosLosVerbos` exige que cada verbo del registro de producción tenga su parte en un fichero de `schemas/`, y `-actualizar-esquemas` los regenera desde `--describe` | `internal/app/esquemas_test.go:28`, `:55-65`, `:340-`, `:479` |
| V30 | El juicio de un `comando` usa las invocaciones de la traza de `strace` (`satisface`: consulta, código 0, mismo applet y, según la forma, verbo, norma, bloque, términos o municipio); un prohibido cuenta termine como termine. `LeerSesion` solo lee del transcript el `tool_use` de la herramienta `Skill` | `internal/evals/juzgar.go:420-457`, `:596-639`; `internal/evals/sesion.go:383-419` |
| V31 | El entorno de cada sesión lo compone `entornoDeLaSesion` (caché, proxies cerrados, `CLAUDE_CONFIG_DIR`, temporal); el guion lee `../pregunta.txt` y `../modelo.txt` | `internal/evals/sesiones.go:481-503`; `scripts/evals-sesion.sh` |
| V32 | El peor caso de un trabajo es `485 + ⌈N/C⌉ × (22 + 240 + 10)` s, y `TestDefinicionDelJob` falla si `timeout-minutes` no lo cubre; el cierre del workflow espera 10 800 s salvo `KITLEGAL_CIERRE_ESPERA_MAX`, y recoge el informe de la comprobación `evals (<skill>)` de entre las marcas de su registro | `internal/evals/definicion.go:27-44`, `:265-331`; `scripts/workflow/cierre.sh:52`, `:64-84` |
| V33 | `scripts/workflow/informe.sh` agrupa `tasas` por fila (`eval` más un sufijo si la pregunta es ampliada o la serie está fuera del plan) y enseña, de cada fila, **la primera** serie de cada valor de `modelo`; saca las columnas de `modelo_que_decide`, `modelos_informativos` y los `modelo` de `tasas`; marca `informativa` la fila con una serie cuyo `modelo` es exactamente `modelo_que_decide`; y da una fila de expresiones por elemento, con su `modelo`. Con un informe sintético en el que las series del modo herramienta llevan `modelo` `<id> (herramienta)`, la tabla da cuatro columnas, una celda por serie de cada modo, el `✗` en la que falla y la marca `informativa` | `scripts/workflow/informe.sh:156-179`, `:204-228`; `jq -r --argjson cambios '{}' --arg cabeza '' -f tasas.jq informe.json` |
| V34 | El informe final resuelve `evals:<skill>:<nombre>` partiendo por los dos primeros `:` (el nombre puede llevar más) y `ci:<ruta>:<Test>` con `git grep '^(func <Test>\(|<Test>:)'` en la ruta | `scripts/workflow/informe.sh:250-273`; `scripts/workflow/comun.sh:160-166` |
| V35 | El prototipo de `boe-legislacion` con las dos formas y la regla nueva tiene **297 líneas** y pasa `prosa-de-la-skill`, `expresiones-de-la-skill`, `ordenes-para-powershell`, `expresiones-calibradas` (36, 11 y 9) y `expresiones-en-los-bloques` de `TestEvalsDelRepositorio`, y `TestSkillsDelRepositorio`, `TestOrdenesDeLasSkillsEmpotradas` y `TestTablaDeComandosCoincideConLaGramatica`; el de `legal-core`, 190 líneas, pasa los tres últimos | `go -C <worktree> test -run …` en el worktree temporal; `wc -l` |
| V36 | Las reglas del conjunto de `boe-legislacion` exigen entre 10 y 20 evals y exactamente 10 positivas que deciden, y cada positiva tiene que citar una norma propia; las de `legal-core`, al menos 3 | `internal/evals/conjunto.go:202-228`, `:324-436` |
| V37 | El texto de las `instructions` del contrato ocupa 485 bytes | `wc -c h21-informe/instrucciones.txt` |
| V38 | `/bin/`, `/dist/` y `/coverage.*` están en `.gitignore` | `.gitignore:2-6` |
| V39 | Con el binario de hoy: un error de argumentos escribe en la salida de error el evento `level=WARN msg=invocación applet=boe verbo="" duracion=… clase=argumentos` y, detrás, `argumentos inválidos: expected "<bloque>"`, y sale con 2; con `--json`, el sobre de fallo lleva `"fuente":"kitlegal.cli","url":"kitlegal:cli"` y `data` `{"clase":"argumentos","mensaje":"argumentos inválidos: expected \"<bloque>\""}`. `--dry-run`, con `--json` o sin él, deja la salida estándar con 0 bytes y escribe en la de error `--dry-run: no se ha ejecutado nada; se habría ejecutado el applet "territorio", el verbo "resolver", con los argumentos [...]`: el kernel llama a `Ejecutar` y después describe, sin sobre | `go run ./cmd/kitlegal boe articulo BOE-A-2015-10565 --no-graph`; `go run ./cmd/kitlegal territorio resolver --dry-run --no-graph [--json] Leganés` con la salida estándar a un fichero (`wc -c`: 0); `internal/app/main.go:445-464` |
| V40 | `scripts/workflow/aceptacion.sh rojo-primero` da por inválido un guion cuya salida lleva `unknown command`, `cannot parse`, `unexpected command` o `usage: `, y por bueno el que falla sin ellos: una orden del arnés que aún no existe o un `exec` que falla no valen como rojo. La ayuda general de hoy empieza por `uso: kitlegal <applet> [verbo] [banderas]`, lista `boe`, `graph`, `skills` y `territorio`, uno por línea con dos espacios delante, y no lleva `mcp`; `kitlegal mcp …` sale hoy con 2 (`"mcp" no es ningún applet de kitlegal`) | `scripts/workflow/aceptacion.sh:45-63`; `go run ./cmd/kitlegal --help`; `go run ./cmd/kitlegal mcp serve --describe` |
| V41 | Registrar un applet cambia cuatro listas literales: `internal/app/registro_test.go:248`, `cmd/kitlegal/main_test.go:20`, `internal/app/ejemplo/kitlegal-e2e/main_test.go:166` e `internal/app/testdata/script/argumentos.txtar:24` y `:30` (`applets disponibles: boe, contar, echo, graph, skills, territorio`); `CONTRIBUTING.md:85` la repite en un ejemplo. H6, H19 y H7 lo hicieron en una tarea `[datos]` indivisible con el esquema publicado del applet. `Verbo.Salida` nulo es un verbo que no declara su `data`, y `--describe` lo deja sin restringir | `grep` de `applets disponibles` y de las listas; `specs/010-h7-internal-graph-grafo/tasks.md` T016 y `specs/009-h19-instalar-sin-clonar/tasks.md` T014; `internal/app/applet.go:47-51` |
| V42 | `make evals-sondeo` entra por `scripts/evals-sondeo-llavero.sh`, que rechaza un paso del workflow y, en macOS con el llavero y sin la credencial en el entorno, pide su contraseña antes de comprobar los argumentos; `scripts/evals-sesion.sh` termina su orden en `--disallowedTools WebFetch WebSearch`; `scripts/evals.sh` exige `kitlegal` en el `PATH`; y el peor caso del trabajo cuenta la prueba de red en las dos skills, la lleve o no el trabajo | `scripts/evals-sondeo-llavero.sh:25-51`; `scripts/evals-sesion.sh:20-22`; `scripts/evals.sh:121-124`; `internal/evals/definicion.go:301-331` |
| V43 | Lo que un cliente recuerda de `robots.txt` por sitio no son solo las reglas: `reglasDelSitio.denegado` guarda también, sin caducidad, el no haberlo podido obtener —un 429, que no se reintenta; un 5xx o un fallo de transporte agotados los tres intentos; un contenido ilegible; una cadena de redirecciones en bucle o excedida—, y desde ahí toda petición de ese cliente a ese sitio devuelve `limite-o-tos` sin pedir nada. Solo el contexto vencido o cancelado no deja entrada. Esa copia y el limitador viven en la misma entrada `sitio`, una por cliente y sitio. H2 fijó la copia «sin caducidad» porque «un cliente vive una invocación de la línea de órdenes, que dura segundos», y dejó «cualquier vigencia entre invocaciones» fuera de la memoria del cliente | `internal/httpx/robots.go:26-38`, `:116-137`, `:259-312`; `internal/httpx/sitio.go:20-49`; `internal/httpx/cliente.go:51`, `:263-294`; `specs/003-h2-internal-httpx-cliente/spec.md:31`, `:238` (leído en la sesión del corrector del plan) |
| V44 | El turno de un sitio es un cubo de un solo token, y `Wait` no espera un turno que no cabe en el plazo del contexto: devuelve el error en el acto y la petición sale como `fuente-no-disponible` («el turno en el sitio no llega dentro del plazo de la operación»). El test de H2 que mide el ritmo compara cada llegada al servidor de prueba con su turno, contado desde el comienzo, sin holgura. Cada cliente clona el transporte por omisión de la biblioteca, cuyas conexiones ociosas se cierran solas a los 90 s (`IdleConnTimeout`): un cliente que deja de usarse no deja nada abierto pasado ese tiempo | `internal/httpx/sitio.go:51-55`, `:91-94`; `internal/httpx/ritmo.go:41-73`; `internal/httpx/ritmo_test.go:34-88`; `internal/httpx/transporte.go:15-34`; `src/net/http/transport.go:47-58` de go1.27.1 (sesión del corrector del plan) |
| V45 | `net/http/httptest` solo se usa en los tests de `internal/httpx`: R2 reserva `net/http` a ese árbol, por prefijo y también en los `_test.go` (V27), y no veda `net`, que hoy ningún fichero del módulo importa solo. Las tablas que recorren las opciones del cliente son `TestOpcionesInvalidas`, `TestReplayRechazaOpciones` y `TestErrorMensajesEnEspanol`; `TestNewSinOpciones` y `TestMapaDeSitios` leen por dentro el intervalo y el limitador de un sitio; y `TestDependenciasDeRed` construye hoy un cliente y pide en ensayo, sin emitir nada | `grep -rl httptest internal cmd`; `.golangci.yml:145-160`; `internal/arch_test.go:575-580`; `grep -rn -E '^\s*"net"$' --include='*.go' .` (sin resultados); `internal/httpx/cliente_test.go:26-56`, `internal/httpx/reproducir_test.go:503`, `internal/httpx/errores_test.go:283`, `internal/httpx/sitio_test.go:121`; `internal/app/boe_test.go:616-640` (sesión del corrector del plan) |

## S · Supuestos no verificados

| # | Supuesto | Por qué no se pudo comprobar | Qué lo mide |
|---|---|---|---|
| S1 | La v1.8.0 del SDK se comporta como la v1.7.0 en lo que usa el hito | No se pudo descargar (V1) | No se usa: el hito fija la v1.7.0 (D1) |
| S2 | `govulncheck` no encuentra vulnerabilidades alcanzables en `go-sdk` v1.7.0 ni en los seis módulos que arrastra | `make vuln` pide a `vuln.go.dev`; no se ejecutó en esta sesión | `make ci` (`vuln`) en la tarea que añade la dependencia y en CI |
| S3 | En las otras cinco plataformas de distribución el SDK enlaza los mismos siete módulos | Se midió solo en darwin/arm64 (V4): cambiar `GOOS` exige una variable de entorno que el modo de permisos de la sesión no deja poner | `TestDependenciasDelBinario` en `make ci` |
| S4 | Con `claude -p … --mcp-config <fichero>`, las herramientas del servidor están disponibles desde el primer turno como `mcp__kitlegal__<applet>_<verbo>`, y `--permission-mode bypassPermissions` las deja llamar sin preguntar | Exige abrir una sesión con modelo (ADR 0032); V21 solo da la bandera y la forma del nombre | El job de cierre, en el modo herramienta |
| S5 | En el transcript `stream-json`, una llamada a una herramienta MCP es un bloque `tool_use` con `name` e `input`, y su resultado, un bloque `tool_result` de un mensaje `user` con `tool_use_id`, `content` e `is_error` | Ningún transcript del repositorio trae una; los sintéticos de `internal/evals/testdata/` traen la misma forma para `Skill` | El job de cierre; el juicio no depende solo de `is_error` (D17) |
| S6 | El validador del cliente de Claude Code acepta el esquema de salida de `--describe` (`if`/`then`/`else`, `$defs`, `format`) y el sobre real | V20 dice que valida, no con qué opciones; no hay un validador JavaScript en el stack para probarlo | El job de cierre; los sobres ya casan con ese esquema con el validador de los tests (`TestSalidaContraSuEsquema`) |
| S7 | Claude Code no pasa su entorno entero al servidor que arranca | No se pudo leer con certeza del binario | No se depende de ello: el fichero de configuración del servidor lleva su `env` (D16) |
| S8 | El shell de la herramienta Bash de Claude Code, en el runner, no vuelve a poner en el `PATH` el directorio de `kitlegal` | Depende del perfil del runner | El juicio: una orden de `kitlegal` ejecutada en una sesión sin binario en el `PATH` la deja sin pasar (D18) |
| S9 | Las sesiones del modo herramienta tardan como las del modo orden (452 s con 96 sesiones en el cierre de H7.4) | No se ha medido | `duracion_de_las_sesiones:herramienta`, que decide |
| S10 | Un trabajo de GitHub Actions en un runner alojado admite `timeout-minutes: 240` | Documentación de la plataforma, sin copia local | El propio job |
| S11 | Lo que el ADR 0035 dice de Codex, de Antigravity y de la app de Claude | El spec lo toma como está | La prueba humana de SC-002 |

## D · Decisiones

### D1 · La dependencia se fija en `go-sdk` v1.7.0, con dos indirectas en la versión que la caché tiene verificada

`require github.com/modelcontextprotocol/go-sdk v1.7.0`, y como indirectas `golang.org/x/oauth2 v0.37.0` y
`github.com/segmentio/asm v1.2.1` (más las que `go mod tidy` añade: `google/jsonschema-go` v0.4.3, `segmentio/encoding`
v0.5.4, `yosida95/uritemplate/v3` v3.0.2, `golang.org/x/sync` v0.22.0). La orden es la de V2, en una sola línea, para que
la selección mínima no pase por versiones sin verificar.

**Por qué.** Es la única versión cuyo código se pudo leer y ejecutar (V3-V19), ya atiende la especificación 2026-07-28 y
las anteriores (V6, V14, V15), y es la única que el ejecutor puede añadir dentro del sandbox del run (V1, V2): con la
v1.8.0, o con las indirectas que el SDK declara, `go get` y `go mod tidy` fallan al escribir `sumdb`.

**Alternativas rechazadas.** *v1.8.0*, la que nombra el ADR 0035: no se puede leer ni añadir en el run (V1); la subida
queda para Dependabot, con la conformidad y los e2e de este hito como red. *Desactivar la comprobación de sumas*
(`GONOSUMDB`, `GOFLAGS=-insecure`): quita una garantía de la cadena de suministro para esquivar un permiso.

**Fuera del hito.** Que el sandbox de `scripts/claude-modelo.sh` no deje escribir en `$GOPATH/pkg/sumdb` es un defecto del
proceso (ADR 0032 quería permitir «módulos, sumas y vulnerabilidades»): cualquier hito que añada un módulo sin verificar
se parará ahí. Arreglarlo es un cambio de `scripts/`, que va en su propia rama.

### D2 · Un paquete nuevo, `internal/mcp`, y nada del SDK fuera de él

`internal/mcp` es el adaptador del protocolo: el único paquete que importa el SDK (FR-026). Exporta cinco cosas
—`Instrucciones`, `Herramienta`, `Resultado`, `Servicio` y `Servir`— y no importa nada del módulo salvo la biblioteca
estándar: recibe las herramientas ya hechas. El applet `mcp` vive en `internal/app`, que es quien tiene el registro y el
kernel. `internal/mcp/mcptest` lleva los dos clientes de prueba (D12).

**Alternativa rechazada.** El SDK en `internal/app`: el «paquete del servidor» sería la raíz de composición entera, y
`depguard` no aislaría nada.

### D3 · Una llamada es una invocación del kernel, por el mismo código

`internal/app` construye por herramienta un `Llamar(argumentos)` que: convierte los argumentos en la línea de órdenes
del verbo (D5); ejecuta `resolverApplet` —la gramática de Kong, el contexto con el plazo, `Ejecutar`, el plazo vencido—
con un presentador cuyo escritor de salida es un búfer de la llamada; y termina con `Montador.Emitir`, que escribe el
sobre en ese búfer y entrega al grafo antes de volver (V23). El sobre de la llamada son los bytes del búfer sin el salto
final: los que la orden escribe con `--json`. `Main` se parte en dos funciones para compartir ese final (`emitir`), sin
cambiar lo que hace.

- El despacho no se repite: el applet y el verbo son los de la herramienta.
- Las banderas de la llamada son las del servidor: `--json`, `--timeout=<plazo>` y, si se dieron a `mcp serve`,
  `--offline` y `--no-graph`. Los argumentos de la herramienta van detrás de `--` (V24): ninguno puede ser una bandera.
- El registro con el que se resuelve es una copia del registro sin avisador: el aviso de versión lo da el kernel una
  vez, al analizar `mcp serve` (V23), y no en cada llamada (FR-023).
- El registrador es el del servidor: una línea por llamada, con el nivel ya resuelto.
- La entrega al grafo ocurre dentro de la llamada, antes de devolver su resultado: la llamada que el cliente envía
  después la ve (FR-013).

**Alternativas rechazadas.** *Rellenar el struct de argumentos desde el JSON*, sin Kong: se saltaría los valores por
omisión, los enumerados y los tipos propios (`cli.Literal`), y «los mismos argumentos» dejaría de ser cierto.
*Lanzar un subproceso por llamada*: el ritmo por sitio no se compartiría (D8) y dependería de la ruta del ejecutable
(FR-025).

### D4 · Un verbo que sirve: interfaz sin exportar en `internal/app`

`ejecutarVerbo` mira si los argumentos del verbo implementan `servidor` (sin exportar; solo `mcp serve`). Si la
implementan y no hay `--dry-run`, los llama con un contexto sin plazo, el contexto de ejecución, el registrador, el
presentador del kernel y el registro, y la invocación termina **sin sobre**: 0 si vuelve sin error, y el fallo de siempre
si no. Con `--dry-run` sigue el camino de todo verbo: `Ejecutar`, que en `mcp serve` valida y no sirve, y la descripción
del kernel (FR-022). La entrada estándar llega por las dependencias del applet (`DependenciasDeMCPDelSistema`, con
`os.Stdin`, que `forbidigo` no veda, V27), y la salida, por `Salida()` del presentador, que es de `internal/render` (R5).

**Por qué.** Sin esto, V22: plazo, código 4 y un sobre en la salida estándar.

**Alternativas rechazadas.** *Atender `mcp serve` fuera del registro*, como `version`: no tendría `--describe`, ni
ayuda, ni banderas globales (FR-021, FR-022). *Un campo exportado en `Verbo`* o una interfaz exportada: API de applet
para un solo verbo (YAGNI). *`os.Stdout` en las dependencias*: lo veda R5.

### D5 · Argumentos de una llamada: comprobados y convertidos en un solo sitio, sin validador

`cli.LineaDeLlamada(verbo, argumentos)` recorre los campos del struct de argumentos —los mismos que recorre
`--describe`— y devuelve la línea de órdenes del verbo, o un error de la clase `argumentos`, si:

- los argumentos no son un objeto JSON, o falta `arguments` (cuenta como `{}`);
- sobra una propiedad (también una bandera global: FR-020);
- un valor no es del tipo de su campo: cadena para un campo de cadena, booleano, número entero, o lista de ellos;
- falta un posicional y se da uno posterior (`bloques` sin `norma` en `graph_check`), que no tiene orden equivalente.

Que falte uno obligatorio lo dice Kong al analizar la línea, con el mensaje de la orden. Los posicionales van en el orden
de sus campos detrás de `--`; un campo que no es posicional, como `--<nombre>=<valor>` delante.

**Por qué sin validador.** Lo que el esquema de entrada puede decir sale de esos mismos campos (`describe.go`), y lo demás
(obligatorios, enumerados, patrones) lo aplica Kong, que es quien decide en la orden. Un validador de JSON Schema sería
una dependencia directa más (`google/jsonschema-go`, fuera de la lista del principio V, o `santhosh-tekuri` enlazado en
el binario con `x/text`) para comprobar tres cosas. La conformidad comprueba, verbo a verbo, que el tipo que declara el
esquema es el que `LineaDeLlamada` exige (FR-070).

**Alternativa rechazada.** La `AddTool` genérica del SDK, que valida: devuelve su propio texto de error y no el sobre de
la clase `argumentos` (V7; FR-011).

### D6 · Los dos esquemas de una herramienta salen del generador de `--describe`

`cli.EsquemasDeHerramienta(verbo)` usa el mismo `generador`: la entrada, con los campos del verbo y **sin** los de
`Globales`; la salida, el sobre con `data` condicionado a `ok`. Cada uno lleva en `$defs` las definiciones que referencia
(hoy, ninguna en la entrada, V25), sin `$schema`, en JSON compacto. La conformidad los compara con `--describe`.

**Alternativa rechazada.** Recortar el documento de `--describe` ya serializado: dos caminos para el mismo esquema, y las
referencias `#/$defs/…` quedarían rotas al separar las dos partes.

### D7 · El servidor: solo herramientas, de solo lectura, y la entrada cerrada es un final normal

`Servir` registra cada herramienta con `(*Server).AddTool` (V7), `Annotations.ReadOnlyHint: true`, y
`Capabilities: &ServerCapabilities{Tools: &ToolCapabilities{}}` (V8): ni `logging`, ni `listChanged`, ni `prompts`, ni
`resources` (FR-008). El transporte es `IOTransport` (V9) sobre la entrada de las dependencias, envuelta en un lector que
anota si llegó a su fin, y el escritor del presentador. Cuando `Run` vuelve: si la entrada llegó a su fin, `Servir`
devuelve `nil`, haya o no error (V10, V11); si no, el error de `Run` (V13), que el kernel presenta como cualquier fallo
(`inesperado`, código 1).

El registrador del servidor se pasa al SDK (V19): con el nivel por omisión solo deja sus errores; con `--verbose`,
también sus líneas informativas. No pasarlo descartaría sus errores internos.

**Alternativas rechazadas.** *Reconocer el cierre por el texto del error* (`server is closing: EOF`): depende de una
cadena de un paquete interno del SDK. *Anunciar `prompts` o `resources` vacíos*: fuera de alcance.

### D8 · Cada llamada, su cliente HTTP, como la orden; el ritmo por sitio, de todo el proceso

`internal/httpx` gana un valor, `Ritmo` —los turnos por sitio, con su intervalo—, y una opción, `ConRitmo`: el cliente
que la recibe espera turno en ese `Ritmo`, que comparte con todo cliente que lo reciba; lo demás sigue siendo de cada
cliente, también lo que recuerda de `robots.txt`. `DependenciasDeRed` crea un `Ritmo` con el intervalo del BOE y sigue
construyendo **un cliente por invocación**, como hoy (V26), cada uno con ese `Ritmo`.

- **En la orden no cambia nada**: una invocación por proceso, un cliente y un `Ritmo` que nadie más usa.
- **En el servidor**, cada llamada que llega a la red construye su cliente (D3). Lo que obtiene de `robots.txt` —las
  reglas, o la denegación por no haberlo podido obtener (V43)— vale para esa llamada y se va con ella: es lo que H2 fijó
  para una invocación, y sigue siendo cierto palabra por palabra, porque el cliente sigue viviendo una (H2 FR 015, que no
  se sustituye). Los turnos sí son de todas: el BOE no recibe del servidor más de una petición por intervalo, sea de una
  llamada o de cinco, y el `robots.txt` de cada una ocupa el suyo (FR-014).

Lo que eso da, visto desde quien llama:

- Una llamada que llega a la red hace las peticiones de su orden —su `robots.txt` y lo que pide— y tarda lo que ella. Una
  que se sirve de la caché no construye ningún cliente, como hoy.
- Un fallo al obtener `robots.txt` no pasa de la llamada que lo sufre: la siguiente lo vuelve a pedir y, si la fuente
  responde, lee. Es el sobre de la orden para la misma entrada y el mismo estado (FR-010), y las reglas no se usan nunca
  más allá de la llamada que las leyó.
- Entre dos llamadas no queda en memoria más que el turno de cada sitio. Un cliente que ya no se usa no deja nada
  abierto: sus conexiones ociosas se cierran solas (V44).
- Varias llamadas a la vez que piden al BOE esperan turno una detrás de otra, y la que no lo alcanza dentro de su plazo
  devuelve `fuente-no-disponible`: es lo que el ritmo ya hace con un plazo que no da para el turno (V44).

`ConRitmo` declara el ritmo del cliente en lugar de `ConIntervalo`; si se dan las dos vale la última, como con una opción
repetida. Un `Ritmo` nulo o de intervalo no positivo es un error de la clase `argumentos` al construir el cliente, como
`ConIntervalo(0)`, y `Replay` la rechaza como rechaza `ConIntervalo`: su cadena no tiene ritmo. Sin `ConRitmo`, un
cliente tiene su propio ritmo, como hoy.

**Tests.** `TestRitmoCompartido`, en `internal/httpx`, con el servidor local de sus tests: dos clientes con el mismo
`Ritmo` no adelantan ninguna de sus cuatro llegadas —el `robots.txt` y el recurso de cada uno— a su turno, con la medida
de H2 (V44); y con un `robots.txt` que falla la primera vez que se pide (429) y responde después, el primer cliente
recibe `limite-o-tos` y el segundo lee. `TestDependenciasDeRed`, en `internal/app`, comprueba la composición: dos
invocaciones reciben dos clientes distintos que comparten los turnos. Lo hace sin esperar nada: con un `Ritmo` de una
hora, un plazo en el contexto de cada petición y un sitio local que responde 404 a todo y cuenta sus conexiones, el
primer cliente gasta el turno en su `robots.txt`; el segundo falla en el acto, sin abrir ninguna conexión, porque su
turno no cabe en el plazo (V44); y uno de otras dependencias sí la abre. Ese sitio es un `net.Listener` del propio test:
`net/http` sigue vedado fuera de `internal/httpx` (V45).

**Alternativas rechazadas.**

- *Un cliente por proceso*, que es lo que este plan decía. Lo que el cliente recuerda de `robots.txt` viviría lo que el
  servidor, que una app de escritorio mantiene horas o días. Si la primera llamada que necesita la red llega sin
  conexión, o el BOE responde 5xx o 429 a `/robots.txt`, el sitio quedaría denegado hasta reiniciar la app: toda llamada
  que no esté en la caché devolvería `limite-o-tos` mientras la orden, con su cliente nuevo, lee en cuanto vuelve la
  red. Es otro sobre para la misma entrada y el mismo estado (FR-010) y una señal que no se apaga; y las reglas pasarían
  días sin releerse, cuando RFC 9309 §2.4 pide no usar la copia más de 24 h (principio I).
- *Un cliente por proceso y vigencia para la copia* dentro de `internal/httpx` (la denegación, hasta la llamada
  siguiente; las reglas, 24 h). Son dos vigencias y un reloj más en el cliente, y aun así el servidor respondería
  distinto que la orden mientras durase una denegación. Lo que ahorra —no repetir `robots.txt` en cada llamada que llega
  a la red— se lo ahorraría igual a las órdenes: es una copia de `robots.txt` entre invocaciones, que H2 dejó fuera de
  la memoria del cliente (V43). El hito pide que la llamada recorra el camino de la orden, no uno más corto.
- *Un cliente por llamada sin nada compartido*, como hoy: cinco llamadas a la vez serían cinco peticiones simultáneas al
  BOE (FR-014).
- *Los turnos en una variable del paquete `httpx`*, sin opción: estado global, que la raíz de composición no ve ni los
  tests pueden aislar.

### D9 · `--timeout`, `--offline`, `--no-graph`, `--verbose`, `--asunto`, `--dry-run`

- `--timeout`, `--offline` y `--no-graph` de `mcp serve` pasan a cada llamada (D3).
- `--verbose` fija el nivel del registrador del servidor, que es el de las llamadas.
- `--asunto <valor>`: `mcp serve` devuelve un error de la clase `argumentos` antes de servir, en `Ejecutar` y en
  `servir` (FR-021).
- `--dry-run`: D4.
- El contexto de cada llamada es el de la orden: el de fondo con el plazo. La cancelación que envíe el cliente no
  interrumpe la llamada; su resultado se descarta (V11). No se añade un contexto al kernel para un caso que el spec no
  pide.

### D10 · El texto de las `instructions`

Cinco frases, 485 bytes (V37), en una constante de `internal/mcp`; el texto, en
[contracts/servidor-mcp.md §5](./contracts/servidor-mcp.md). El test cuenta **bytes**, que acota por arriba a los
caracteres: la documentación de Codex que da la cifra no se pudo leer en local.

### D11 · `schemas/servidor.json` para `mcp serve`

`TestEsquemasCubrenTodosLosVerbos` exige la parte de cada verbo de producción (V29): la tabla de `esquemas_test.go` gana
`{applet: "mcp", nombre: "servidor.json", entidad: "servidor", verbos: ["serve"]}` y el fichero lo escribe
`-actualizar-esquemas`, en una tarea `[datos]`.

### D12 · El arnés e2e: una orden `mcp` de testscript con dos clientes

Los guiones no pueden dar las peticiones por un fichero (V12). El arnés gana la orden `mcp`, que arranca el programa del
guion, conecta un cliente, hace las llamadas de un fichero esperando cada respuesta y cierra la entrada. Los clientes
viven en `internal/mcp/mcptest`: el del SDK, que negocia `2026-07-28` (V14), y, con `-anterior`, uno de la `2025-11-25`
escrito con la biblioteca estándar (V15), que además comprueba que cada línea de la salida estándar del servidor es un
mensaje del protocolo. Contrato en [contracts/arnes-e2e.md](./contracts/arnes-e2e.md).

**Alternativas rechazadas.** *Importar el SDK desde el arnés de `internal/app`*: lo veda `depguard` (V27; FR-079).
*Un binario cliente aparte*: otro `package main` con excepciones de `forbidigo`. *Forzar al cliente del SDK a la versión
anterior* interceptando `server/discover`: depende de un detalle interno de su negociación.

### D13 · Conformidad y concurrencia, en proceso

`TestHerramientasDelServidor` (FR-070) y `TestLlamadasSimultaneas` (FR-074) arrancan `Main` con `mcp serve` en una
gorrutina, con tuberías como entrada y salida, y hablan con él con `mcptest`: así corren con `-race` sobre el código del
servidor, que en un subproceso el detector no ve. La conformidad se ejerce con el registro de producción (diez
herramientas) y con uno que lleva además los applets de ejemplo (FR-004). Del mismo modo, `TestPlazoDeCadaLlamada`
(FR-020) arranca `mcp serve --timeout=<plazo corto>` con un registro que lleva un verbo del test que espera a su
contexto: es lo que ve fallar un servidor que heredara el plazo que el kernel pone a todo verbo (V22, D4) o una llamada
a la que no le llegara (D9); en un guion no cabe, porque ningún verbo del binario de e2e se queda esperando.

### D14 · La tabla de comandos: una columna más

`RenderizarTabla` pasa a `| Orden | Herramienta | Qué hace | Qué devuelve en data |`, con `<applet>_<verbo>` en la
columna nueva, y la línea del sobre dice que la orden y la herramienta de cada fila devuelven el mismo. Ninguna fila
nueva: el recuento de líneas de cada `SKILL.md` no cambia por la tabla. `TestOrdenesDeLasSkillsEmpotradas` comprueba
además que cada herramienta de la tabla está entre las que da el registro (FR-031).

### D15 · Las dos skills: dónde va cada cosa

Un bloque al principio de «Protocolo» dice las dos formas, cuándo se usa cada una, cómo se reconoce la herramienta con
el prefijo del agente, que el `&&` son dos llamadas y que el código es `data.clase`; cada paso nombra su herramienta
junto a su orden; «Comandos» da la correspondencia entera de códigos y clases; y una regla nueva, la última, da la línea
`⚠ SIN CONSULTA AL BOE:`. Para caber, tres órdenes sueltas pasan de bloque a código en línea y se quitan líneas en
blanco junto a bloques: 297 líneas en `boe-legislacion` (V35). Texto, en
[contracts/skills.md](./contracts/skills.md).

**La línea.** `⚠ SIN CONSULTA AL BOE: <causa>. Para consultarlo hace falta instalar kitlegal: https://kitlegal.es/instalar/`.
No lleva ninguna expresión de la lista (V35), así que no hace falta declararla forma fija.

**Alternativa rechazada.** Una skill por modo (ADR 0035). Y repetir cada orden con su llamada en un bloque: no cabe.

### D16 · Los modos de una sesión de eval

- **Orden**: la de hoy, sin cambios.
- **Herramienta**: el repartidor escribe `servidor.json` en el directorio de la sesión
  (`{"mcpServers":{"kitlegal":{"command":"<ruta absoluta de kitlegal>","args":["mcp","serve"],"env":{…}}}}`, con la
  caché y los proxies de la sesión en `env`: S7) y quita del `PATH` de la sesión el directorio del binario; el guion
  añade `--mcp-config ../servidor.json` si el fichero está (V21). El modo de una sesión se lee de ahí: lo es si tiene
  `servidor.json`.
- **Sin binario ni servidor**: la de una eval que lo declara; sin `servidor.json` y con el `PATH` sin el binario.

El job abre tres tandas seguidas con la concurrencia de hoy —orden, herramienta y las de las evals sin binario ni
servidor—, y mide cada una por separado: la duración de un modo es la de su tanda (FR-043).

**Alternativas rechazadas.** *Los dos modos a la vez*: ocho sesiones simultáneas sobre la suscripción, que nadie ha
medido. *Un trabajo por modo*: dos informes por skill, contra FR-048. *Un `modo.txt`*: obligaría a tocar cada sesión
sintética de `testdata/`.

### D17 · Una llamada a una herramienta, en el juicio

`LeerSesion` lee del transcript cada bloque `tool_use` cuyo nombre, o lo que sigue a su último `__`, es el de una
herramienta del registro, con su `input`, y su `tool_result` (S5). El juicio la convierte en una invocación: el applet y
el verbo de la herramienta, los argumentos en el orden de `cli.LineaDeLlamada`, y como código, 0 si el resultado no es un
error y el de su `data.clase` si lo es (1 si no se puede leer). Es un error si `is_error` es verdadero **o** si el
contenido es un sobre con `ok` falso. Con eso, `satisface`, los prohibidos y «fuera de lo grabado» valen sin cambios
(FR-042, FR-041).

El proceso del servidor aparece en la traza como una invocación de `mcp serve`: no es una consulta, y sus conexiones se
atribuyen a ella, de modo que una petición que llegue a la red sigue poniendo el informe en `fallo`.

### D18 · Una orden de `kitlegal` en una sesión que no debía tenerlo

En una sesión del modo herramienta o sin binario ni servidor, una invocación de applet en la traza que no sea `mcp serve`
deja la sesión sin pasar, con su motivo. Es lo que hace cierta la definición del modo (FR-040, FR-046) sin depender del
perfil del runner (S8). La vía es real: `servidor.json`, que está en el directorio de la sesión, lleva la ruta absoluta
del binario, y una sesión que la lea puede ejecutarlo sin el `PATH`; sin esta regla, sus órdenes satisfarían los
`comandos` y el modo herramienta mediría órdenes.

### D19 · El informe por modo

- `tasas`: una serie por eval, modelo y modo. Su `modelo` es el id en el modo orden y en las evals sin binario ni
  servidor, y `<id> (herramienta)` en el modo herramienta; gana `modo`. Así el `informe.sh` de hoy da una celda por serie
  y conserva la marca de las informativas (V33).
- `expresiones_prohibidas_por_modelo`: un elemento por modelo y modo, con `modelo` `<id> (<modo>)` y `modo`.
- `umbrales`: `<recuento>:<modelo>:<modo>` y `duracion_de_las_sesiones:<modo>` (V34): diez en `boe-legislacion`, ocho
  que deciden.
- Los motivos de una serie y de un umbral ya nombran su `modelo` o su nombre, y con ellos el modo.

**Alternativas rechazadas.** *El modo en `eval`*: el informe final contaría 41 evals y perdería la marca de las nuevas.
*Sumar los modos*: 3 de 108 cumpliría (FR-043).

### D20 · El tope del trabajo: 240 minutos

Con tres tandas, el peor caso es `485 + (⌈97/4⌉ + ⌈96/4⌉ + ⌈6/4⌉) × 272` = 14 357 s (239,3 min) en `boe-legislacion` y
`485 + (19 + 18 + 6) × 272` = 12 181 s (203 min) en `legal-core`: `timeout-minutes: 240`. Lo esperado son unos 17 min de
sesiones en `boe-legislacion` (S9).

**Consecuencia que se declara.** El cierre del workflow espera 3 h (V32). Solo en el peor caso —todas las sesiones
agotando su tope— el trabajo duraría más que esa espera, y el cierre se detendría con su dossier antes de leer un informe
que sería `fallo` de todos modos. Queda anotado como supuesto.

**Alternativa rechazada.** Subir la concurrencia a 6 para que el peor caso quepa en 163 min: cambia el punto en el que
H7.3 midió los límites de la suscripción, sin poder medirlo aquí.

### D21 · La eval sin binario ni servidor

Clave nueva del formato, `sin_binario_ni_servidor: true`, solo con `activa: true` y sin comandos, citas ni nada que exija
consultar. El juicio añade dos comprobaciones: la línea con su dirección, y ninguna cita. Las reglas del conjunto dejan
de contarla como positiva y exigen exactamente una por skill; el tamaño de `boe-legislacion` pasa a 21. Preguntas, en
[contracts/evals-en-dos-modos.md §1](./contracts/evals-en-dos-modos.md).

### D22 · El sondeo

`evalsPedidas` da, por cada eval pedida que declara `sin_binario_ni_servidor`, el error
`EVALS: <nn> es una eval sin binario ni servidor: solo la mide el job de evals`, en el orden de `EVALS` y junto a los
demás; es un `errorDeUso` como los de hoy. Nada más cambia (FR-050).

### D23 · Datos externos

Ninguno. Los guiones usan las grabaciones de H4 y las derivadas de H7 y H7.1 que ya están en `testdata/`; ninguna
respuesta nueva del BOE hace falta, y `grabar_datos` no tiene nada que grabar.

### D24 · Lo que la prueba de FR-080 hace con `informe.sh`

El test lee `scripts/workflow/informe.sh`, toma el programa `jq` de `seccion_evals` y el de `recuentos_y_umbrales` de
entre sus marcas de texto, y comprueba sobre el `Informe` de dos modos la regla que esos programas aplican (V33): ninguna
pareja de fila y `modelo` repetida, y cada recuento con un `modelo` distinto que nombra su modo. **No ejecuta `jq`**: no
es del stack, y `make ci` no lo exige. Si el texto de esos programas cambia, el test falla y nombra la línea: la regla
que comprueba dejaría de ser la de «el `informe.sh` de hoy». La ejecución real del programa sobre un informe de dos modos
se hizo una vez, en esta sesión (V33).

**Alternativa rechazada.** Ejecutar el guion entero en el test: necesita `jq`, un repositorio con `main` y el directorio
del feature.

### D25 · La precondición de los guiones de aceptación

Cada guion empieza por `exec kitlegal --help` y `stdout '^\s+mcp\s'`: la orden termina bien hoy y después, y la aserción
falla hoy porque la ayuda no lista `mcp` (V40). Así el «rojo primero» es por una aserción, y ningún guion llega hoy a la
orden `mcp` del arnés, que todavía no existe.

**Alternativas rechazadas.** *Empezar por la orden `mcp` del arnés*: hoy da `unknown command`, que el workflow toma por
un guion que no se puede leer (V40). *Empezar por `exec kitlegal mcp serve --describe`*: hoy falla el propio `exec`, que
`testscript` informa como `unexpected command failure`, también inválido.

### D26 · Registrar `mcp`: una tarea `[datos]` indivisible, con `argumentos.txtar`

El registro del applet en producción y en el binario de e2e va en una sola tarea con `schemas/servidor.json`, su fila de
`esquemas_test.go` y las cuatro listas literales de applets (V41), entre ellas las líneas 24 y 30 de
`internal/app/testdata/script/argumentos.txtar`, que ganan `mcp`. El cuerpo del applet va antes, probado sobre un
registro local del test. `mcp serve` no declara `Salida` (V41): al servir no emite ningún sobre.

**Por qué, si el spec deja fuera «editar los guiones de los hitos anteriores».** FR-001 pide el applet en el registro y
FR-071, el e2e contra el binario con replay, que es el de e2e; ese guion de H1 fija la lista de applets de ese binario,
así que registrar `mcp` lo cambia, como lo cambiaron `territorio` (H6), `skills` (H19) y `graph` (H7). La lectura
conservadora es tocar exactamente esa lista y nada más de ningún guion anterior. Queda anotado como supuesto.

**Alternativa rechazada.** Un segundo binario de e2e, solo para `mcp`: el e2e dejaría de ejercer el registro que se
distribuye, y FR-004 se prueba precisamente con «un registro con un applet más».
