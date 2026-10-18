# Quickstart: validar H9 · `plazos` + `data/festivos/` + `legal-core` v1

Guía de validación de la entrega, escenario a escenario. Cada uno remite a su contrato y a sus FR/SC; no repite el
detalle. Se ejecuta **después** de implementar el hito: antes, el applet `plazos`, `data/festivos/`, las evals nuevas y
los tests que se nombran no existen.

## 0. Antes de empezar

- Desde la raíz del repositorio, en **una sola sesión** de `bash` o `zsh` (los escenarios comparten `REPO`, `T`, `K` y
  las funciones `k` y `hablar`), con el toolchain de `go.mod`, `git`, `make` y `jq`.
- **Sin red de ninguna fuente y sin sesiones con modelo**, salvo §14 (el job, que lanza el workflow) y §15 (la prueba
  humana). `plazos` y `territorio` no usan la red; los tests leen lo grabado. La única red es la de `make ci` (`make
  vuln` consulta su base de datos), como en todos los hitos.
- **Efectos**: `go test`, `go run` y `go build` escriben solo en las cachés de Go; `make ci` deja además
  `coverage.out` y `coverage-integration.out`, que `.gitignore` ignora. Todo lo demás va bajo `$T`, que §16 borra. Cada
  invocación del binario pasa por `k`, que le da un `HOME` y un `KITLEGAL_CACHE_DIR` bajo `$T`: nada toca
  `~/.cache/kitlegal`. Ningún escenario escribe en el árbol: los tests de derivación se ejecutan **sin**
  `-actualizar-festivos`. §16 compara `git status --porcelain` con el del principio.

```sh
REPO=$(git rev-parse --show-toplevel)
cd "$REPO"
INICIAL=$(git status --porcelain)
T=$(mktemp -d)
mkdir -p "$T/bin" "$T/home"
go build -o "$T/bin/kitlegal" ./cmd/kitlegal
K="$T/bin/kitlegal"
k() { env HOME="$T/home" KITLEGAL_CACHE_DIR="$T/cache" "$@"; }
hablar() {
  for m in '{"jsonrpc":"2.0","id":1,"method":"initialize","params":{"protocolVersion":"2025-11-25","capabilities":{},"clientInfo":{"name":"quickstart","version":"0"}}}' \
    '{"jsonrpc":"2.0","method":"notifications/initialized"}' "$@"; do
    printf '%s\n' "$m"
    sleep 2
  done
}
```

`hablar` es el cliente de `printf` del quickstart de H21: escribe el saludo y cada mensaje que recibe, uno cada dos
segundos, y al terminar cierra la entrada del servidor.

## 1. `make ci` en verde (FR-080)

```sh
make ci
```

Termina con `ci: todos los controles en verde`, con `schema-check` (`schemas/plazo.json` y `schemas/cobertura.yaml.json`
sin diferencias) y `skills-check` (la tabla y las cuatro referencias de `legal-core` sin diferencias).

## 2. Los guiones de aceptación (FR-001 a FR-013, FR-026, FR-085; SC-009)

```sh
go test -race -count=1 -run '^TestEntregaDelHito$' -v ./internal/app/ 2>&1 | grep -E '^ *--- (PASS|FAIL): TestEntregaDelHito/h9-'
```

Cinco líneas `--- PASS`: `h9-plazos-calcular`, `h9-plazos-cobertura`, `h9-plazos-codigos`, `h9-plazos-herramienta` y
`h9-plazos-sin-efectos` (plan.md, «Aceptación e2e»).

## 3. El caso (a): la fecha, el precepto y los festivos (US1; contracts/applet-plazos.md §4)

```sh
k "$K" plazos calcular reposicion 2026-09-09 Leganés Leganés --json > "$T/a.json"; echo "código $?"
jq -c '.data | {fecha, preceptos, municipios, festivos, avisos}' "$T/a.json"
k "$K" plazos calcular reposicion 2026-09-09 Leganés Leganés --json --offline > "$T/a-offline.json"; echo "código $?"
diff <(jq -c '{data, hash}' "$T/a.json") <(jq -c '{data, hash}' "$T/a-offline.json") && echo "iguales"
```

Código 0 las dos veces; `fecha` `2026-10-13`; `preceptos` `[{"norma":"BOE-A-2015-10565","bloque":"a124","apartado":"124.1"}]`;
`municipios` `[{"codigo_ine":"28074","nombre":"Leganés"}]`; `festivos` el `2026-10-09` (`local`, `28074`,
`BOCM-20251212-34`) y el `2026-10-12` (`nacional`, `BOE-A-2025-21667`); `avisos` `[]`; e `iguales` (FR-013, SC-009).

## 4. Tordesillas y 2027: la fecha avisa de lo que le falta (US2; FR-026)

```sh
k "$K" plazos calcular reposicion 2026-09-09 Tordesillas Tordesillas --json | jq -c '.data | {fecha, avisos, cobertura}'
k "$K" plazos calcular reposicion 2026-12-14 Leganés Leganés --json | jq -c '.data | {fecha, avisos: [.avisos[].codigo]}'
k "$K" plazos calcular generico 2024-12-30 Leganés Leganés --duracion 3 --unidad dias-habiles --json | jq -c '.data | {fecha, avisos: [.avisos[].codigo]}'
```

`2026-10-09` con un aviso `faltan-festivos-locales` cuyo texto empieza por `⚠ FALTAN FESTIVOS LOCALES:` y nombra
Tordesillas y 2026, y `cobertura` con `{"ambito":"local","municipio":"47165",…,"falta":"sin-festivos-locales"}`;
`2027-01-14` con `faltan-festivos-locales`; y `2025-01-03` con `faltan-festivos-nacionales-y-autonomicos` y
`faltan-festivos-locales` (contracts/casos-de-plazos.md, casos 2, 18 y 17).

## 5. El contencioso: agosto y los días del art. 183 (US5; FR-010)

```sh
k "$K" plazos calcular contencioso 2026-06-15 Madrid --json | jq -c '.data | {fecha, preceptos, avisos}'
k "$K" plazos calcular contencioso 2026-10-27 Madrid --json | jq -c '.data | {fecha, avisos: [.avisos[].codigo]}'
```

`2026-09-15` con los preceptos `a46` y `a128` de `BOE-A-1998-16718` y sin avisos; y `2026-12-28` con
`dias-del-art-183-lopj` (casos 3 y 15).

## 6. Los códigos de salida y su orden (FR-012; contracts/applet-plazos.md §6)

```sh
for args in "inventado 2026-09-09 Leganés Leganés" "reposicion 2026-02-30 Leganés Leganés" \
  "generico 2026-09-09 Leganés Leganés" "reposicion 2026-09-09 Leganés Leganés --duracion 2" \
  "silencio 2026-09-09 Leganés Leganés --duracion 3" "reposicion 2026-09-09 Leganés" \
  "contencioso 2026-09-09 Madrid Leganés" "reposicion 2026-09-09 Arroyomolinos Leganés" \
  "reposicion 2026-09-09 Inexistente Leganés" "inventado 2026-09-09 Inexistente Leganés"; do
  eval "k \"\$K\" plazos calcular $args --json" > "$T/e.json" 2>/dev/null; c=$?
  echo "$c $(jq -r '.data.clase' "$T/e.json") ← $args"
done
```

`2 argumentos` en todas salvo `reposicion 2026-09-09 Inexistente Leganés`, que da `3 no-encontrado`; la de
Arroyomolinos lleva en `data.mensaje` sus dos candidatos tras «municipio «Arroyomolinos»: », y la última da 2: el tipo
va antes que el municipio.

## 7. La herramienta (FR-001, FR-085)

```sh
hablar '{"jsonrpc":"2.0","id":2,"method":"tools/list"}' \
  '{"jsonrpc":"2.0","id":3,"method":"tools/call","params":{"name":"plazos_calcular","arguments":{"tipo":"reposicion","inicio":"2026-09-09","municipios":["Leganés","Leganés"]}}}' \
  '{"jsonrpc":"2.0","id":4,"method":"tools/call","params":{"name":"plazos_calcular","arguments":{"tipo":"inventado","inicio":"2026-09-09","municipios":["Leganés","Leganés"]}}}' \
  | k "$K" mcp serve > "$T/mcp.jsonl" 2> "$T/mcp.err"; echo "código $?"
jq -c 'select(.id==2) | [.result.tools[].name] | (length, index("plazos_calcular") != null)' "$T/mcp.jsonl"
jq -c 'select(.id==3) | .result.structuredContent.data.fecha' "$T/mcp.jsonl"
jq -c 'select(.id==4) | .result.isError, .result.structuredContent.data.clase' "$T/mcp.jsonl"
```

Código 0; 13 herramientas, con `plazos_calcular`; `"2026-10-13"`; y `true` con `"argumentos"`: el servidor responde
a la llamada siguiente al error.

## 8. Sin red, sin caché y sin grafo (FR-013, FR-028; SC-009)

```sh
find "$T/home" "$T/cache" -type f 2>/dev/null | sort > "$T/antes"
k "$K" plazos calcular reposicion 2026-09-09 Leganés Leganés --json > /dev/null
k "$K" plazos calcular contencioso 2026-06-15 Madrid --json --offline > /dev/null
find "$T/home" "$T/cache" -type f 2>/dev/null | sort > "$T/despues"
diff "$T/antes" "$T/despues" && echo "sin cambios"
go test -count=1 -run '^TestArquitectura$' ./internal/
```

`sin cambios` y `ok`: ni la caché ni el grafo ganan nada, y el código de `plazos` no alcanza la red.

## 9. `data/festivos/` sale de lo grabado (US3; FR-020 a FR-025, FR-083, FR-084; SC-006, SC-007)

```sh
go test -count=1 -run '^(TestFestivosNacionalesDerivados|TestLeerResolucion)$' -v ./internal/source/boe/fiestas/ 2>&1 | grep -E '^(--- |ok|FAIL)'
go test -count=1 -run '^(TestFestivosLocalesDerivados|TestCotejoConElCatalogo|TestLeerDisposicion|TestCruceConElTerritorio)$' -v ./internal/source/bocm/ 2>&1 | grep -E '^(--- |ok|FAIL)'
grep -E 'municipio: "28080"' data/festivos/locales/13.yaml | grep 2025
grep -E 'municipio: "28026"' data/festivos/locales/13.yaml | grep 2026
```

Todos `--- PASS`. Majadahonda (`28080`) en 2025: el `2025-09-15` con `BOCM-20250203-16` y el `2025-11-25` (US3.2).
Brunete (`28026`) en 2026: el `2026-05-15` y el `2026-09-14`, con `BOCM-20251229-19`, la corrección que los da (US3.3).

## 10. Los preceptos, contra lo grabado (FR-010, FR-030, FR-031)

```sh
go test -count=1 -run '^(TestPreceptosDeLosPlazos|TestIdentificadoresDeLasNormas|TestGrabacionesSinSolape|TestManifiestoDeGrabaciones)$' -v ./internal/evals/ 2>&1 | grep -E '^(--- |ok|FAIL)'
```

Todos `--- PASS`.

## 11. La tabla de casos y las propiedades (FR-011, FR-081, FR-082; SC-004, SC-005)

```sh
go test -count=1 -run '^(TestCasosDePlazos|TestFestivoDeUnaParte|TestAppletPlazos)$' -v ./internal/app/ 2>&1 | grep -E '^(--- |ok|FAIL)'
go test -count=1 -run '^(TestPropiedadesDelComputo|TestCalcular)$' -v ./internal/core/plazos/ 2>&1 | grep -E '^(--- |ok|FAIL)'
go test -count=1 -cover ./internal/core/plazos/
```

Todos `--- PASS`, con un subtest por cada uno de los 21 casos de contracts/casos-de-plazos.md (19 con fecha y
2 de cobertura); la cobertura de
`internal/core/plazos`, al menos el 85 %.

## 12. Las evals y los umbrales con sesiones sintéticas (FR-050 a FR-057, FR-087; SC-011, SC-012)

```sh
go test -count=1 -run '^(TestUmbralesDeLegalCore|TestUmbralesDeLasTresSkills|TestFormatoDePlazos|TestJuzgarPlazos|TestEvalsDelRepositorio|TestDefinicionDelJob)$' -v ./internal/evals/ 2>&1 | grep -E '^(--- |ok|FAIL)'
ls evals/legal-core/
```

Todos `--- PASS`; nueve evals, las cuatro de hoy y `05` a `09` (contracts/evals-legal-core.md §5).

## 13. La skill (FR-032, FR-040 a FR-048; SC-010)

```sh
make skills-check
wc -l < skills/legal-core/SKILL.md
head -1 skills/legal-core/references/recursos_y_plazos.md skills/legal-core/references/cobertura_por_materia.md
grep -c '^⚠ ' skills/legal-core/SKILL.md
```

`skills-check` sin diferencias; menos de 300 líneas; cada referencia con `<!-- generado desde data/…, no editar -->`; y
`8`: las dos líneas `⚠ PLAZO:` (la plantilla y la forma con los dos preceptos del contencioso), las dos `⚠ NO CUBIERTO:`
y las cabezas de los cuatro avisos, en sus bloques `text`. La de
`⚠ SIN CONSULTA AL BOE:` va sangrada dentro de la regla 7, como en v0.

## 14. El cierre (lo hace el workflow, tras la revisión final; SC-001, SC-002)

El job de evals sobre la propuesta de cambio. En el `informe.json` de `legal-core`:

```sh
jq -c '.veredicto, [.umbrales[] | {nombre, medida, total, cumple, decide}], (.red | length)' informe.json
```

`aprobado`; seis umbrales, `sin_activar`, `plazo_sin_fecha` y `cobertura_sin_declarar` de los modos orden y
herramienta, con `medida` 0 sobre 21, 9 y 6, `cumple` y `decide` verdaderos; y `0`. El informe final da cada uno
«comprobado por su control». `boe-legislacion` y `jurisprudencia`, si salen en rojo, quedan en el informe (FR-071).

## 15. La persona (SC-003)

Con la release publicada, en la app de escritorio de Claude con el plugin: «me notificaron el 9 de septiembre de 2026
una resolución del Ayuntamiento de Leganés y vivo en Leganés, ¿hasta cuándo puedo presentar el recurso de reposición?»
se responde con el 13 de octubre de 2026, el art. 124.1 LPAC citado y la fiesta local del 9 de octubre como razón; y
«¿qué plazo tengo para recurrir una liquidación de la Agencia Tributaria?», con la declaración de no cubierto, sin dar
un mes.

## 16. Limpieza

```sh
rm -rf "$T"
[ "$(git status --porcelain)" = "$INICIAL" ] && echo "el árbol está como al principio"
```
