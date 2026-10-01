# Quickstart: validar H21 · `kitlegal mcp serve`

Guía de validación de la entrega, escenario a escenario. Cada uno remite a su contrato y a sus FR/SC; no repite el
detalle. Se ejecuta **después** de implementar el hito: antes, el applet `mcp`, los tests que se nombran y las dos
`SKILL.md` nuevas no existen.

## Antes de empezar

- Desde la raíz del repositorio, en **una sola sesión** de shell (los escenarios comparten `REPO`, `T`, `K`, `E` y las
  funciones `k` y `hablar`), con el toolchain de `go.mod`, `git` y `make`; `jq` para leer los mensajes de §5, §6 y §12
  (el mismo que ya usa `scripts/workflow/`).
- **Sin red de ninguna fuente y sin sesiones con modelo**, salvo §11 (el sondeo, que lanza una persona), §12 (el job, que
  lanza el workflow) y §13 (la prueba humana). `boe` responde con el binario de e2e, que sirve las grabaciones de H4
  desde `reproduccion/` y nunca abre una conexión; `territorio` y `graph` no usan la red. La única red es la de `make ci`
  (`make vuln` consulta su base de datos), como en todos los hitos.
- **Efectos**: `go test` y `go build` escriben solo en las cachés de Go; `make ci` deja además `coverage.out` y
  `coverage-integration.out` en la raíz, que `.gitignore` ignora. Todo lo demás va bajo `$T`, que §14 borra. Nada toca
  `~/.cache/kitlegal` ni `~/.agents/`: cada invocación del binario pasa por `k`, que le da un `HOME` y un
  `KITLEGAL_CACHE_DIR` bajo `$T`. Las órdenes de `go` y de `make` se ejecutan sin `k`, con el entorno de siempre. El
  último escenario compara `git status --porcelain` con el del principio.

```sh
REPO=$(git rev-parse --show-toplevel)
cd "$REPO"
INICIAL=$(git status --porcelain)
T=$(mktemp -d)
mkdir -p "$T/bin" "$T/home" "$T/w/reproduccion" "$T/con espacios"
go build -o "$T/bin/kitlegal" ./cmd/kitlegal
go build -ldflags "-X main.reloj=2026-09-28T12:00:00Z" -o "$T/bin/kitlegal-e2e" ./internal/app/ejemplo/kitlegal-e2e
cp "$T/bin/kitlegal" "$T/con espacios/kitlegal"
cp -R internal/source/boe/testdata/boe.legislacion-consolidada "$T/w/reproduccion/"
K="$T/bin/kitlegal"; E="$T/bin/kitlegal-e2e"
k() { env HOME="$T/home" KITLEGAL_CACHE_DIR="$T/cache" "$@"; }
hablar() {
  for m in '{"jsonrpc":"2.0","id":1,"method":"initialize","params":{"protocolVersion":"2025-11-25","capabilities":{},"clientInfo":{"name":"quickstart","version":"0"}}}' \
    '{"jsonrpc":"2.0","method":"notifications/initialized"}' "$@"; do
    printf '%s\n' "$m"
    sleep 2
  done
}
```

`hablar` es un cliente de la especificación 2025-11-25 hecho con `printf`: escribe el saludo y después cada mensaje que
recibe como argumento, uno cada dos segundos, y al terminar se cierra la entrada del servidor. La espera es lo que deja
al servidor responder antes del cierre (research V12). El binario de e2e busca `reproduccion/` en su directorio de
trabajo: por eso §6 se ejecuta desde `$T/w`.

## 1. `make ci` en verde (FR-079; SC-013)

```sh
make ci
```

Termina con `ci: todos los controles en verde`, con `schema-check` y `skills-check` sin diferencias, los guiones
`h21-mcp-*` dentro de `TestEntregaDelHito` y `depguard` sin hallazgos.

## 2. Los guiones de aceptación (FR-071 a FR-073, FR-075, FR-076; SC-004 a SC-006, SC-008)

```sh
go test -race -count=1 -run '^TestEntregaDelHito$/^h21-mcp-' -v ./internal/app/ 2>&1 | grep -E '^ *--- (PASS|FAIL)'
```

Seis líneas, todas `--- PASS`: la de `TestEntregaDelHito` y las de `h21-mcp-llamadas`, `h21-mcp-errores`,
`h21-mcp-herramientas`, `h21-mcp-protocolo` y `h21-mcp-proceso` ([contracts/arnes-e2e.md §4](./contracts/arnes-e2e.md)).
Los guiones llevan ese prefijo desde que el workflow activa la suite, tras el bucle de tareas.

## 3. La conformidad de las herramientas (FR-070; SC-003)

```sh
go test -race -count=1 -run '^TestHerramientasDelServidor$' -v ./internal/app/
go test -count=1 -run '^(TestLineaDeLlamada|TestEsquemasDeHerramienta)$' ./internal/cli/
go test -count=1 -run '^TestEsquemasPublicados$' ./internal/app/
```

Pasan: con el registro de producción, diez herramientas; con el que lleva los applets de ejemplo, trece; cada una con
el nombre, la descripción y los dos esquemas de `--describe` de su verbo y de solo lectura; `capabilities` es
`{"tools":{}}`; y `schemas/servidor.json` es el de `mcp serve --describe`.

## 4. Llamadas simultáneas, el cierre y el ensayo (FR-014, FR-022, FR-024, FR-074; SC-007)

```sh
go test -race -count=1 -run '^(TestLlamadasSimultaneas|TestServirSinEntrada|TestServirEnEnsayo|TestDependenciasDeRed)$' ./internal/app/
go test -race -count=1 ./internal/mcp/...
go test -race -count=1 -run '^TestRitmoCompartido$' ./internal/httpx/
```

Pasan, sin carreras: cada llamada recibe el sobre de su entrada; la entrada que se cierra con una llamada en curso da 0;
`--dry-run` vuelve sin leer la entrada; dos invocaciones reciben cada una su cliente de red y los dos esperan turno en
el mismo ritmo por sitio; y un `robots.txt` que falla para un cliente se vuelve a pedir con el siguiente.

## 5. A mano: listar y llamar con el binario distribuido (US1, US4, US6; FR-002, FR-005, FR-006, FR-010, FR-011)

```sh
k "$K" territorio resolver Leganés --json > "$T/orden.json"
hablar '{"jsonrpc":"2.0","id":2,"method":"tools/list"}' \
  '{"jsonrpc":"2.0","id":3,"method":"tools/call","params":{"name":"territorio_resolver","arguments":{"consulta":"Leganés"}}}' \
  '{"jsonrpc":"2.0","id":4,"method":"tools/call","params":{"name":"boe_articulo","arguments":{"norma":"BOE-A-2015-10565"}}}' \
  '{"jsonrpc":"2.0","id":5,"method":"tools/call","params":{"name":"skills_install","arguments":{}}}' \
  | k "$K" mcp serve > "$T/a.jsonl" 2> "$T/a.err"; echo "código $?"
wc -l < "$T/a.jsonl"
jq -c 'select(.id==1) | .result | .protocolVersion, .capabilities, .serverInfo.name, (.instructions | utf8bytelength)' "$T/a.jsonl"
jq -r 'select(.id==2) | .result.tools[] | "\(.name) \(.annotations.readOnlyHint)"' "$T/a.jsonl"
jq -r 'select(.id==3) | .result.content[0].text' "$T/a.jsonl" | cmp - "$T/orden.json" && echo "mismo sobre"
jq -c 'select(.id==3) | .result | (.structuredContent == (.content[0].text | fromjson)), .isError' "$T/a.jsonl"
jq -c 'select(.id==4) | .result | .isError, .structuredContent.ok, .structuredContent.data.clase' "$T/a.jsonl"
jq -c 'select(.id==5) | .error.code, (.result == null)' "$T/a.jsonl"
```

Esperado: `código 0`; `5` líneas (una respuesta por petición con `id`); `"2025-11-25"`, `{"tools":{}}`, `"kitlegal"` y
`485`; diez líneas, de `boe_analisis true` a `territorio_resolver true`; `mismo sobre`; `true` y `null` (el resultado
que termina bien no lleva `isError`); `true`, `false` y `"argumentos"` (falta `bloque`: el sobre de fallo, como error de
herramienta, sin pedir nada a ninguna fuente); y `-32602` y `true` (una herramienta que el servidor no anuncia la
responde el protocolo, sin sobre). El servidor siguió atendiendo después de la llamada que falló (FR-015).

## 6. A mano: el sobre de `boe_articulo` es el de su orden, y el grafo lo ve (US1.1, US1.6; FR-010, FR-013)

```sh
cd "$T/w"
k "$E" boe articulo BOE-A-2015-10565 a21 --json > "$T/a21.json"
hablar '{"jsonrpc":"2.0","id":2,"method":"tools/call","params":{"name":"boe_articulo","arguments":{"norma":"BOE-A-2015-10565","bloque":"a21"}}}' \
  '{"jsonrpc":"2.0","id":3,"method":"tools/call","params":{"name":"graph_check","arguments":{"norma":"BOE-A-2015-10565","bloques":["a21"]}}}' \
  | k "$E" mcp serve > "$T/b.jsonl" 2> "$T/b.err"; echo "código $?"
cd "$REPO"
jq -r 'select(.id==2) | .result.content[0].text' "$T/b.jsonl" | cmp - "$T/a21.json" && echo "mismo sobre"
jq -c 'select(.id==3) | .result | .isError, .structuredContent.ok' "$T/b.jsonl"
ls "$T/cache"
```

Esperado: `código 0`; `mismo sobre` (el binario de e2e tiene el reloj fijo, así que también coincide `fecha_consulta`);
`null` y `true` (`graph_check` responde con un resultado, no con un error); y en `$T/cache`, `cache.db` y `world.db`
(con sus auxiliares `-wal` y `-shm` si SQLite los deja). La llamada lee de la caché que dejó la orden: ninguna de las
dos abre una conexión.

## 7. A mano: el proceso (US4.3 a US4.5, US4.7; FR-021, FR-022, FR-024, FR-025; SC-008)

```sh
k "$K" mcp serve < /dev/null > "$T/fin.out"; echo "código $?"; wc -c < "$T/fin.out"
k "$K" mcp serve --asunto x < /dev/null > "$T/asunto.out" 2> "$T/asunto.err"; echo "código $?"
wc -c < "$T/asunto.out"; cat "$T/asunto.err"
k "$K" mcp serve --dry-run > "$T/ensayo.out" 2> "$T/ensayo.err"; echo "código $?"
wc -c < "$T/ensayo.out"; cat "$T/ensayo.err"
(cd / && hablar '{"jsonrpc":"2.0","id":2,"method":"tools/call","params":{"name":"territorio_resolver","arguments":{"consulta":"Leganés"}}}' \
  | k "$T/con espacios/kitlegal" mcp serve > "$T/raiz.jsonl" 2> "$T/raiz.err"; echo "código $?")
jq -r 'select(.id==2) | .result.content[0].text' "$T/raiz.jsonl" | cmp - "$T/orden.json" && echo "mismo sobre"
```

Esperado: con la entrada cerrada, `código 0` y `0` bytes. Con `--asunto x`, `código 2`, `0` bytes en la salida estándar
y, en la de error, el evento de la invocación (`level=WARN msg=invocación applet=mcp … clase=argumentos`, como en
cualquier orden que falla) y `argumentos inválidos: mcp serve no admite --asunto: el servidor no expone nada del asunto`. Con
`--dry-run` y la entrada de la terminal abierta, vuelve al momento con `código 0`, `0` bytes en la salida estándar y la
línea `--dry-run: no se ha ejecutado nada; se habría ejecutado el applet "mcp", el verbo "serve", …` en la de error.
Desde `/` y desde una ruta con espacios, `código 0` y `mismo sobre`.

## 8. Las dos skills (US2, US3; FR-030, FR-031, FR-035, FR-036; SC-009)

```sh
make skills-check
wc -l skills/boe-legislacion/SKILL.md skills/legal-core/SKILL.md
grep -c '^ *⚠ SIN CONSULTA AL BOE: <causa>\. Para consultarlo hace falta instalar kitlegal: https://kitlegal\.es/instalar/$' skills/boe-legislacion/SKILL.md skills/legal-core/SKILL.md
grep -c '^| Orden | Herramienta | Qué hace | Qué devuelve en `data` |$' skills/boe-legislacion/SKILL.md skills/legal-core/SKILL.md
go test -count=1 -run '^(TestOrdenesDeLasSkillsEmpotradas|TestTablaDeComandosCoincideConLaGramatica)$' ./internal/app/
go test -count=1 -run '^TestRenderizarTabla$' ./internal/skills/
```

Esperado: `skills-check` en verde, sin drift y con el calibrado de H7.4 en 36, 11 y 9; cada `SKILL.md` con menos de 300
líneas (297 y 190 en el prototipo); `1` en cada fichero para la línea; `2` y `1` cabeceras de tabla con la columna
«Herramienta»; y los tests pasan ([contracts/skills.md](./contracts/skills.md)).

## 9. Las `instructions` y la arquitectura (FR-006, FR-026, FR-077, FR-079; SC-010, SC-013)

```sh
go test -count=1 -run '^TestInstrucciones$' -v ./internal/mcp/
go test -count=1 -run '^(TestArquitectura|TestDependenciasDelBinario)$' ./internal/
go list -deps ./... | grep -c '^github.com/modelcontextprotocol/go-sdk/mcp$'
grep -rlE --include='*.go' '"github.com/modelcontextprotocol/go-sdk/[a-z0-9/]+"$' cmd internal | sort
```

Esperado: los tests pasan; `1`; y, con una línea de importación del SDK, solo ficheros de `internal/mcp/` y de
`internal/mcp/mcptest/` (`internal/arch_test.go` nombra el módulo en sus listas, no lo importa).

## 10. Las evals en dos modos, sin modelo (FR-042 a FR-047, FR-080, FR-081, FR-083; SC-011, SC-012)

```sh
go test -count=1 -run '^(TestUmbralesDelInforme|TestInformeEnDosModos|TestJuzgarLasLlamadas|TestJuzgarSinBinarioNiServidor|TestLeerLasLlamadas)$' ./internal/evals/
go test -count=1 -run '^(TestPlanEnDosModos|TestSesionesPorModo|TestGuionDeLaSesion|TestDefinicionDelJob|TestEvalsDelRepositorio)$' ./internal/evals/
go test -count=1 -run '^(TestComprobarElSondeo|TestSondear|TestGuionDelSondeo)$' ./internal/evals/
grep -n 'timeout-minutes: 240' .github/workflows/evals.yml
ls evals/boe-legislacion/21-sin-binario-ni-servidor.yaml evals/legal-core/04-sin-binario-ni-servidor.yaml
```

Pasan, con sesiones y transcripts sintéticos y el `claude` sustituto: los casos de
[contracts/evals-en-dos-modos.md §8](./contracts/evals-en-dos-modos.md). El tope del trabajo es de 240 minutos y las dos
evals nuevas están.

## 11. El sondeo rechaza una eval sin binario ni servidor (FR-050, FR-084)

Lo lanza una persona, fuera del run: un paso del workflow no puede (`scripts/evals-sondeo-llavero.sh` lo rechaza; ADR
0032), y lo que cubre este caso en `make ci` son los tests de §10. En macOS con el llavero del sondeo, pide antes su
contraseña. No abre ninguna sesión.

```sh
make evals-sondeo SKILL=boe-legislacion EVALS=01,21 MODELO=claude-sonnet-5-5 REPETICIONES=1; echo "código: $?"
```

Esperado: en la salida de error, solo `EVALS: 21 es una eval sin binario ni servidor: solo la mide el job de evals` (y
la línea de error de `make`); la salida estándar, vacía; `código: 2`.

## 12. El job de cierre (FR-078; SC-001)

Lo lanza el workflow sobre la propuesta de cambio, tras la revisión final; aquí solo se lee su informe, que el cierre
deja en `specs/015-h21-kitlegal-mcp-serve/gates/evals/`:

```sh
jq -c '.veredicto, .red, [.umbrales[] | {nombre, medida, total, cumple, decide}]' specs/015-h21-kitlegal-mcp-serve/gates/evals/boe-legislacion.json
jq -c '[.tasas[] | select(.planificada and .decide and (.pasa | not)) | {eval, modelo}]' specs/015-h21-kitlegal-mcp-serve/gates/evals/boe-legislacion.json specs/015-h21-kitlegal-mcp-serve/gates/evals/legal-core.json
jq -c '[.tasas[] | select(.eval | test("sin-binario")) | {eval, modelo, pasan, sesiones}]' specs/015-h21-kitlegal-mcp-serve/gates/evals/boe-legislacion.json specs/015-h21-kitlegal-mcp-serve/gates/evals/legal-core.json
jq -c '.veredicto, .umbrales' specs/015-h21-kitlegal-mcp-serve/gates/evals/legal-core.json
```

Esperado: `"aprobado"`, `[]` y diez umbrales, los ocho de «Controles de umbral» de plan.md con `decide: true` y
`cumple: true` (`expresiones_prohibidas:claude-sonnet-5-5:orden` y `…:herramienta`, como mucho 2 de 54 cada uno;
`sin_activar:…` y `redaccion_no_leida:…`, 0; `duracion_de_las_sesiones:orden` y `…:herramienta`, ≤ 900); `[]` y `[]`
(ninguna serie que decide sin pasar, en ningún modo); la serie de la eval sin binario ni servidor de cada skill con el
modelo que decide, con al menos 2 de 3; y `legal-core`, `"aprobado"` y `[]`.

## 13. La prueba humana, después de fusionar (SC-002)

Fuera del run y sin control en él. Con `kitlegal` de la rama fusionada instalado y las skills en `~/.agents/skills/`
(`kitlegal skills install -g`):

1. En la app de escritorio de ChatGPT, añadir el servidor en *Settings > MCP servers* (orden `kitlegal`, argumentos
   `mcp serve`) y preguntar «¿qué dice el art. 21 de la Ley 39/2015?». La respuesta lleva
   `art. 21 de la Ley 39/2015 [BOE-A-2015-10565, bloque a21]` y el agente no ejecuta ninguna orden.
2. En la app de escritorio de Claude, sin carpeta, instalar con doble clic un `.mcpb` hecho a mano con ese binario, como
   el de la prueba del ADR 0035, y hacer la misma pregunta: la misma cita.

## 14. Limpieza

```sh
cd "$REPO"
rm -rf "$T"
test "$(git status --porcelain)" = "$INICIAL" && echo "árbol intacto"
```

Esperado: `árbol intacto`.
