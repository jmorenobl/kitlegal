# Quickstart: validación de H5.1

Guía **ejecutable** para comprobar que H5.1 entrega lo que dice. Cada escenario se ejecuta tal cual, desde la raíz del
repositorio, sobre la rama `007-h5-1-avisos-de-vigencia` **con el hito ya implementado y confirmado**. Los nombres de
test son los del inventario de [plan.md](./plan.md); los formatos, los de [contracts/](./contracts/).

**Sin efectos colaterales.** Los escenarios 1 a 10 no crean ni modifican ningún fichero del árbol de trabajo, no tocan
el índice ni el historial de git y no escriben en la caché de `kitlegal` de la cuenta. Lo que se rompe a propósito se
rompe en `/tmp/kitlegal-quickstart-h51/`: un **clon desechable** de la rama (`git clone` solo lee el repositorio), que los
prerrequisitos crean vacío y la limpieza borra. Las órdenes `go test` y `make` de la guía no escriben perfiles de
cobertura. El escenario 11 es de plataforma: lo ejecuta la tarea `[plataforma]` y, además de esa carpeta temporal,
escribe solo lo que declara —la rama remota, la propuesta de cambio y, si hay que repetir la aceptación, su etiqueta
`evals`—; su evidencia la escribe la tarea en `specs/007-h5-1-avisos-de-vigencia/gates/evals-aceptacion.md`.

**El directorio del hito queda fuera de las comprobaciones del árbol**: el workflow `hito` reescribe ahí sus ficheros de
estado, así que `git status` descarta `specs/007-h5-1-avisos-de-vigencia/` en cualquier estado. Una ejecución retomada
empieza siempre por los prerrequisitos.

**Cada orden se ejecuta tal cual en el modo desatendido**, con las formas que la sesión acepta sin pedir aprobación
(`.claude/settings.json`; research V32):

| Necesidad | Forma de la guía |
|---|---|
| Crear, borrar o editar en `/tmp/kitlegal-quickstart-h51/` | `rtk proxy mkdir`, `rtk proxy rm`, `rtk proxy chmod`, `rtk proxy perl -0pi` |
| Leer el código de salida, encadenar órdenes o redirigir | `rtk proxy sh -c '…; echo "código $?"'` |
| Leer o filtrar la salida de `make`, `go test`, `git status` o `git diff` | `rtk proxy` delante de la orden y de cada etapa de la tubería (el filtro de salida de la sesión resume o reescribe esas órdenes) |
| Probar sobre el clon | `rtk proxy go -C /tmp/kitlegal-quickstart-h51/repo test …` y `git -C /tmp/kitlegal-quickstart-h51/repo …` |
| Una orden larga | en una sola línea |

`git`, `wc` sobre ficheros del repositorio y `echo` con texto literal van tal cual; `make` y `go test` van con
`rtk proxy` delante para que su salida llegue entera.

## Prerrequisitos

```bash
go version
git rev-parse --abbrev-ref HEAD
rtk proxy git status --porcelain | rtk proxy grep -vE '^.. specs/007-h5-1-avisos-de-vigencia/' ; echo "fin del estado"
rtk proxy mkdir -p /tmp/kitlegal-quickstart-h51
rtk proxy chmod -R u+w /tmp/kitlegal-quickstart-h51
rtk proxy rm -r /tmp/kitlegal-quickstart-h51
rtk proxy mkdir /tmp/kitlegal-quickstart-h51
git clone --quiet --branch 007-h5-1-avisos-de-vigencia . /tmp/kitlegal-quickstart-h51/repo
rtk proxy git -C /tmp/kitlegal-quickstart-h51/repo status --porcelain ; echo "fin del estado del clon"
```

Esperado: `go version go1.27.1 …`; la rama `007-h5-1-avisos-de-vigencia`; el estado del árbol muestra solo
`fin del estado`; el clon se crea sin mensajes y su estado muestra solo `fin del estado del clon`.

## 1. Controles en verde (FR-060, SC-006)

```bash
rtk proxy make skills-check
rtk proxy make schema-check
rtk proxy go test -count=1 ./internal/source/boe/ ./internal/evals/
```

Esperado: las tres terminan en 0; `make skills-check` da `ok` para `internal/app`, `internal/skills` e
`internal/evals` (con los subtests `avisos-del-esquema` y `avisos-de-la-skill` de `TestEvalsDelRepositorio` dentro);
la tercera, `ok` para los dos paquetes.

## 2. `SKILL.md` fija la forma de los avisos (US1, FR-001 a FR-005, SC-001)

```bash
wc -l skills/boe-legislacion/SKILL.md
rtk proxy grep -nF '⚠ NORMA DEROGADA:' skills/boe-legislacion/SKILL.md
rtk proxy grep -nF '⚠ VIGENCIA AGOTADA:' skills/boe-legislacion/SKILL.md
rtk proxy grep -nF '⚠ TEXTO POSIBLEMENTE DESACTUALIZADO:' skills/boe-legislacion/SKILL.md
rtk proxy grep -nE '^(## |### )' skills/boe-legislacion/SKILL.md
rtk proxy git diff --unified=0 main...HEAD -- skills/boe-legislacion/SKILL.md
```

Esperado (research V33): 199 líneas, menos de 300; la primera búsqueda da dos líneas, la 147 (la lista) y la 152 (el
ejemplo, `⚠ NORMA DEROGADA: esta norma ha sido derogada.`), y las otras dos, una cada una, la 148 y la 149, todas entre
la línea de `## Cómo se cita` (121) y la de `## Comandos` (158); la lista de títulos da el número de línea de cada
sección (`### 5. Responder citando` en la 102 y `## Reglas` en la 182); en el diff, tres bloques, `@@ -114 +114,3 @@`,
`@@ -140,0 +143,15 @@` y `@@ -174,3 +191,4 @@`, cada uno dentro de `### 5. Responder citando`, de `## Cómo se cita` o de
la regla 3 de `## Reglas`, y lo quitado y lo añadido son exactamente los textos del
[contrato de la forma fija](./contracts/forma-fija-de-los-avisos.md) §3. Que el texto no nombra evals, el job ni modelos
y que la región generada no cambia lo comprueban `make skills-check` (escenario 1) y el diff.

## 3. La salida del binario no cambia (US4-4, FR-012)

```bash
rtk proxy git diff --stat main...HEAD -- schemas/norma.json schemas/bloque.json internal/source/boe/testdata internal/app/testdata ; echo "fin del diff"
rtk proxy git diff --stat main...HEAD -- internal/source/boe
rtk proxy make test-e2e
```

Esperado: el primer diff no lista nada (solo `fin del diff`); el segundo lista solo `internal/source/boe/avisos.go` e
`internal/source/boe/avisos_test.go`; `make test-e2e` termina en 0 con los guiones `testscript` y los golden de H4 sin
cambios, y `make schema-check` ya pasó en el escenario 1.

## 4. Las comprobaciones mecánicas fallan nombrando el código (US4-2, US4-3, SC-005)

Todo sobre el clon. Cada mutación va seguida de su diff, que demuestra que se aplicó, de la comprobación, que tiene que
fallar, y de la restauración.

```bash
rtk proxy perl -0pi -e 's/"consolidacion-no-finalizada", "derogada", "vigencia-agotada"/"consolidacion-no-finalizada", "vigencia-agotada"/' /tmp/kitlegal-quickstart-h51/repo/schemas/eval.yaml.json
rtk proxy git -C /tmp/kitlegal-quickstart-h51/repo diff --stat
rtk proxy sh -c 'go -C /tmp/kitlegal-quickstart-h51/repo test -count=1 -run "^TestEvalsDelRepositorio\$/^avisos-del-esquema\$" ./internal/evals/; echo "código $?"'
git -C /tmp/kitlegal-quickstart-h51/repo checkout -- schemas/eval.yaml.json
rtk proxy perl -0pi -e 's/"vigencia-agotada"\]/"vigencia-agotada", "otro-aviso"]/' /tmp/kitlegal-quickstart-h51/repo/schemas/eval.yaml.json
rtk proxy git -C /tmp/kitlegal-quickstart-h51/repo diff --stat
rtk proxy sh -c 'go -C /tmp/kitlegal-quickstart-h51/repo test -count=1 -run "^TestEvalsDelRepositorio\$/^avisos-del-esquema\$" ./internal/evals/; echo "código $?"'
git -C /tmp/kitlegal-quickstart-h51/repo checkout -- schemas/eval.yaml.json
rtk proxy perl -0pi -e 's/⚠ NORMA DEROGADA:/⚠ :/g' /tmp/kitlegal-quickstart-h51/repo/skills/boe-legislacion/SKILL.md
rtk proxy git -C /tmp/kitlegal-quickstart-h51/repo diff --stat
rtk proxy sh -c 'go -C /tmp/kitlegal-quickstart-h51/repo test -count=1 -run "^TestEvalsDelRepositorio\$/^avisos-de-la-skill\$" ./internal/evals/; echo "código $?"'
git -C /tmp/kitlegal-quickstart-h51/repo checkout -- skills/boe-legislacion/SKILL.md
rtk proxy perl -0pi -e 's/⚠ NORMA DEROGADA:/NORMA DEROGADA:/g' /tmp/kitlegal-quickstart-h51/repo/skills/boe-legislacion/SKILL.md
rtk proxy git -C /tmp/kitlegal-quickstart-h51/repo diff --stat
rtk proxy sh -c 'go -C /tmp/kitlegal-quickstart-h51/repo test -count=1 -run "^TestEvalsDelRepositorio\$/^avisos-de-la-skill\$" ./internal/evals/; echo "código $?"'
git -C /tmp/kitlegal-quickstart-h51/repo checkout -- skills/boe-legislacion/SKILL.md
rtk proxy perl -0pi -e 's/⚠ NORMA DEROGADA:/⚠ NORMA DEROGADA/g' /tmp/kitlegal-quickstart-h51/repo/skills/boe-legislacion/SKILL.md
rtk proxy git -C /tmp/kitlegal-quickstart-h51/repo diff --stat
rtk proxy sh -c 'go -C /tmp/kitlegal-quickstart-h51/repo test -count=1 -run "^TestEvalsDelRepositorio\$/^avisos-de-la-skill\$" ./internal/evals/; echo "código $?"'
git -C /tmp/kitlegal-quickstart-h51/repo checkout -- skills/boe-legislacion/SKILL.md
rtk proxy git -C /tmp/kitlegal-quickstart-h51/repo status --porcelain ; echo "fin del estado del clon"
```

Esperado: cada `diff --stat` del esquema da `1 file changed, 1 insertion(+), 1 deletion(-)` y cada uno de `SKILL.md`,
`1 file changed, 2 insertions(+), 2 deletions(-)` (la lista y el ejemplo); cada prueba termina en `FAIL` con `código 1`
y su mensaje: `el esquema de eval no enumera el código de aviso derogada en avisos` (un código de menos),
`el esquema de eval enumera en avisos el código otro-aviso, que no es un código de aviso del binario` (un código de más)
y, en las tres de `SKILL.md` (sin etiqueta, sin marca, sin dos puntos),
`falta la forma fija del aviso derogada: ⚠ NORMA DEROGADA:`; al final el clon queda limpio. Las mismas mutaciones con
los otros dos códigos las cubren `TestComprobarCodigosDeAviso` y `TestComprobarFormasDeAviso`.

## 5. La forma fija y el juicio (US2-3 a US2-8, FR-030 a FR-035, SC-003)

```bash
rtk proxy go test -count=1 -v -run '^(TestExtraerAvisos|TestJuzgar)$' ./internal/evals/ | rtk proxy grep -E -- '--- (PASS|FAIL)'
```

Esperado: todo `PASS` y ningún `FAIL`, con los subtests de `TestExtraerAvisos` del
[contrato de la forma fija](./contracts/forma-fija-de-los-avisos.md) §4 y, en `TestJuzgar`, además de los de H5,
`aviso-con-su-forma-fija`, `aviso-con-variantes-toleradas`, `aviso-ausente`, `aviso-con-otra-redaccion`,
`aviso-negado`, `forma-fija-y-lo-contrario`, `aviso-no-esperado`, `avisos-en-el-orden-de-la-eval`, `aviso-repetido` y
`cita-y-aviso-ausentes`.

## 6. `avisos` en el formato (US2-1, US2-2, US2-9, US2-10, FR-020 a FR-023, SC-002)

```bash
rtk proxy go test -count=1 -v -run '^(TestLeerEval|TestComprobarCodigosDeAviso)$' ./internal/evals/ | rtk proxy grep -E -- '--- (PASS|FAIL)'
```

Esperado: todo `PASS`, con `avisos`, `aviso-desconocido`, `no-activa-con-avisos`, `avisos-vacio`, `citas-vacio`,
`aviso-repetido`, `cita-repetida` e `informativa-con-avisos` entre los casos de `TestLeerEval`, y `exacto`,
`falta-un-codigo`, `sobra-un-codigo`, `sin-enumerado` y `sin-avisos` en `TestComprobarCodigosDeAviso`. Que las 17 evals
de antes siguen válidas lo da `make skills-check` (escenario 1).

## 7. El informe publica los avisos (US3, FR-040 a FR-043, SC-004)

```bash
rtk proxy go test -count=1 -v -run '^(TestInforme|TestInformeConAvisos)$' ./internal/evals/ | rtk proxy grep -E -- '--- (PASS|FAIL)'
```

Esperado: todo `PASS`, con `TestInforme/aprobado` y los dos subtests de `TestInformeConAvisos`,
`uno-encontrado-y-otro-ausente` y `aviso-detras-de-la-cita`.

## 8. La eval de la norma derogada, su norma y lo grabado (US5, FR-050 a FR-057, SC-006, SC-009)

```bash
rtk proxy ls evals/boe-legislacion
rtk proxy cat evals/boe-legislacion/18-lrjpac-norma-derogada.yaml
rtk proxy grep -n 'BOE-A-1992-26318' data/normas.yaml skills/boe-legislacion/references/normas.md
rtk proxy grep -ci 'derog' data/normas.yaml skills/boe-legislacion/references/normas.md ; echo "fin del recuento"
rtk proxy git diff --name-status main...HEAD -- testdata internal/source/boe/testdata internal/evals/testdata
rtk proxy sh -c 'set -e; e=$(git log --reverse --format=%H main..HEAD -- evals/boe-legislacion/18-lrjpac-norma-derogada.yaml | head -n 1); s=$(git log --reverse --format=%H main..HEAD -- skills/boe-legislacion/SKILL.md | head -n 1); test -n "$e"; test -n "$s"; test "$e" != "$s"; git merge-base --is-ancestor "$e" "$s"; echo "la eval ($e) precede al primer cambio de SKILL.md ($s)"'
rtk proxy go test -count=1 -v -run '^(TestEvalsDelRepositorio|TestIdentificadoresDeLasNormas|TestManifiestoDeGrabaciones|TestGrabacionesSinSolape)$' ./internal/evals/ | rtk proxy grep -E -- '--- (PASS|FAIL)'
rtk proxy rm /tmp/kitlegal-quickstart-h51/repo/testdata/evals/boe.legislacion-consolidada/GET_https_www.boe.es_datosabiertos_api_legislacion-consolidada_id_BOE-A-1992-26318_texto_indice.json
rtk proxy sh -c 'go -C /tmp/kitlegal-quickstart-h51/repo test -count=1 -run "^TestEvalsDelRepositorio\$/^grabado\$" ./internal/evals/; echo "código $?"'
git -C /tmp/kitlegal-quickstart-h51/repo checkout -- testdata/evals/boe.legislacion-consolidada
rtk proxy git -C /tmp/kitlegal-quickstart-h51/repo status --porcelain ; echo "fin del estado del clon"
```

Esperado: 18 ficheros de eval, el último `18-lrjpac-norma-derogada.yaml`, con el contenido del
[contrato de la eval](./contracts/eval-norma-derogada-y-grabacion.md) §6; `BOE-A-1992-26318` en una línea de cada fichero;
`data/normas.yaml:0` y `skills/boe-legislacion/references/normas.md:0` seguidos de `fin del recuento`; el diff de
material de prueba con exactamente dos líneas, `M testdata/evals/grabaciones.json` y
`A testdata/evals/boe.legislacion-consolidada/GET_https_www.boe.es_datosabiertos_api_legislacion-consolidada_id_BOE-A-1992-26318_texto_indice.json`;
la línea `la eval (…) precede al primer cambio de SKILL.md (…)`; todo `PASS` (con los subtests `formato`, `conjunto`,
`normas-conocidas`, `grabado`, `avisos-del-esquema` y `avisos-de-la-skill`, y los cuatro negativos de
`TestIdentificadoresDeLasNormas`); sin el índice grabado, `FAIL` con `18-lrjpac-norma-derogada.yaml: la norma
BOE-A-1992-26318 sin indice` y `código 1`; y el clon limpio al final.

## 9. Las etiquetas salen solo del binario (US4-1, US4-5, FR-010, FR-011, SC-005)

```bash
rtk proxy sh -c 'for f in internal/evals/*.go; do case "$f" in *_test.go) ;; *) grep -nHE "NORMA DEROGADA|VIGENCIA AGOTADA|TEXTO POSIBLEMENTE DESACTUALIZADO" "$f" ;; esac; done; echo "fin de la búsqueda"'
rtk proxy go test -count=1 -v -run '^(TestEtiquetasDeAviso|TestAvisosDe|TestCodigosDeAviso|TestEtiquetasSoloDesdeBoe)$' ./internal/source/boe/ ./internal/evals/ | rtk proxy grep -E -- '--- (PASS|FAIL)'
```

Esperado: la búsqueda solo imprime `fin de la búsqueda`; todo `PASS`, con `exactamente-tres`, `frases-con-su-forma` y
`cada-llamada-su-mapa` en `TestEtiquetasDeAviso`.

## 10. Documentación (FR-061, SC-008)

```bash
rtk proxy grep -n 'avisos' README.md CONTRIBUTING.md
rtk proxy grep -n 'H5.1' CHANGELOG.md
```

Esperado: en `README.md` y en `CONTRIBUTING.md`, la fila `avisos` de la tabla del formato común de eval y, en
`CONTRIBUTING.md`, la frase de cuándo pasa una eval con los avisos; en `CHANGELOG.md`, el bloque *De H5.1* de
*Unreleased* con la forma fija en la skill, el campo `avisos`, el informe y la eval nueva.

## 11. Plataforma: la ejecución de aceptación (FR-070 a FR-073, SC-007)

Solo la tarea `[plataforma]`. Qué hace cada paso y cuándo vale: [contrato de la ejecución de
aceptación](./contracts/ejecucion-de-aceptacion.md). La ejecución elegida se guarda en
`/tmp/kitlegal-quickstart-h51/ejecucion.txt`, que leen los pasos siguientes.

### 11.1 Prerrequisitos

```bash
rtk proxy gh auth status
rtk proxy gh secret list
rtk proxy gh label list --search evals
rtk proxy test -s specs/007-h5-1-avisos-de-vigencia/gates/pr-h5.1.md && echo "cuerpo de la propuesta de cambio presente"
```

Esperado: sesión iniciada; `CLAUDE_CODE_OAUTH_TOKEN` entre los secretos; la etiqueta `evals`; el cuerpo presente. Si
falta algo, la tarea se detiene y lo anota.

### 11.2 Publicar

```bash
git push -u origin 007-h5-1-avisos-de-vigencia
rtk proxy sh -c 'if url=$(gh pr view 007-h5-1-avisos-de-vigencia --json url --jq .url 2>/dev/null); then echo "propuesta existente: $url"; else gh pr create --base main --head 007-h5-1-avisos-de-vigencia --title "feat(H5.1): Avisos de vigencia en las evals" --body-file specs/007-h5-1-avisos-de-vigencia/gates/pr-h5.1.md; fi'
gh pr view 007-h5-1-avisos-de-vigencia --json number,state,baseRefName,headRefName,headRefOid,url
```

Esperado: la rama empujada; la dirección de la propuesta de cambio; `state` `OPEN`, base `main` y `headRefOid` igual a
`git rev-parse HEAD`.

### 11.3 Identificar la ejecución de apertura y esperarla

```bash
rtk proxy sh -c 'set -e; gh run list --branch 007-h5-1-avisos-de-vigencia --event pull_request --limit 100 --json databaseId,workflowName,headSha,status,conclusion,createdAt,url --jq "map(select(.workflowName == \"evals\")) | sort_by(.createdAt) | .[0] // \"todavía no hay ninguna ejecución de evals de la rama\""'
rtk proxy sh -c 'set -e; id=$(gh run list --branch 007-h5-1-avisos-de-vigencia --event pull_request --limit 100 --json databaseId,workflowName,createdAt --jq "map(select(.workflowName == \"evals\")) | sort_by(.createdAt) | .[0].databaseId // empty"); test -n "$id"; echo "$id" > /tmp/kitlegal-quickstart-h51/ejecucion.txt; echo "ejecución de apertura: $id"'
rtk proxy sh -c 'set -e; id=$(cat /tmp/kitlegal-quickstart-h51/ejecucion.txt); test -n "$id"; set +e; gh run watch "$id" --exit-status --interval 60 > /dev/null; echo "código $?"'
```

Esperado: la primera orden se repite hasta que muestre una ejecución (la plataforma tarda unos segundos en crearla), con
su `url`, que se registra; la segunda guarda su `databaseId`; la tercera se repite tal cual hasta que la ejecución
termine (si la herramienta la corta por su tope de tiempo, vuelve enseguida con el resultado una vez terminada) y da
`código 0` con el veredicto `aprobado` o `código 1` si el job falló.

### 11.4 Leer el informe

```bash
rtk proxy sh -c 'set -e; id=$(cat /tmp/kitlegal-quickstart-h51/ejecucion.txt); test -n "$id"; test "$(gh run view "$id" --json status --jq .status)" = completed; gh run view "$id" --log > /tmp/kitlegal-quickstart-h51/registro.txt; for f in informe.md informe.json; do grep -qF -- "--- inicio de $f ---" /tmp/kitlegal-quickstart-h51/registro.txt; grep -qF -- "--- fin de $f ---" /tmp/kitlegal-quickstart-h51/registro.txt; sed -n "/--- inicio de $f ---/,/--- fin de $f ---/p" /tmp/kitlegal-quickstart-h51/registro.txt | sed "1d;\$d" | cut -f3- | sed -E "s/^[^ ]+ //" > "/tmp/kitlegal-quickstart-h51/$f"; done; jq -c "{commit, veredicto, modelo_que_decide, repeticiones, umbral}" /tmp/kitlegal-quickstart-h51/informe.json; echo "paso Ejecutar las evals, primera y última línea:"; awk -F "\t" "\$2 == \"Ejecutar las evals\" { print \$3 }" /tmp/kitlegal-quickstart-h51/registro.txt | sed -n "1p;\$p" | cut -d " " -f 1'
```

Solo si la anterior falla con la ejecución terminada porque falta una marca (el job se detuvo antes del informe):

```bash
rtk proxy sh -c 'set -e; id=$(cat /tmp/kitlegal-quickstart-h51/ejecucion.txt); test -n "$id"; gh run view "$id" --log-failed'
```

Esperado: una línea con el `commit`, `aprobado`, `claude-sonnet-5`, `3` y `2`, y los instantes de la primera y la
última línea del paso «Ejecutar las evals», cuya diferencia es la duración que se registra (supuesto S4). Con la orden de
`--log-failed`, la tarea registra su salida, se detiene y lo anota.

### 11.5 Comprobar la aceptación

```bash
rtk proxy jq -e --arg eval 18-lrjpac-norma-derogada.yaml '.modelo_que_decide as $m | (.tasas | map(select(.eval == $eval and .modelo == $m and (.pregunta_ampliada | not)))) as $tasas | (.evals | map(select(.eval == $eval and .modelo == $m))) as $sesiones | {commit, veredicto, red, modelo_que_decide: $m, tasa: ($tasas | map({planificada, decide, sesiones, pasan, pasa})), sesiones: ($sesiones | map({sesion, pasa, avisos_encontrados, avisos_ausentes, motivos}))}, (.veredicto == "aprobado" and .red == [] and ($tasas | length) == 1 and $tasas[0].planificada and ($tasas[0].decide | not) and $tasas[0].sesiones == 3 and ($sesiones | length) == 3 and all($sesiones[]; ((.avisos_encontrados + .avisos_ausentes) | sort) == ["derogada", "vigencia-agotada"]))' /tmp/kitlegal-quickstart-h51/informe.json
rtk proxy sh -c 'set -e; id=$(cat /tmp/kitlegal-quickstart-h51/ejecucion.txt); test -n "$id"; c=$(gh run view "$id" --json headSha --jq .headSha); test -n "$c"; i=$(jq -r .commit /tmp/kitlegal-quickstart-h51/informe.json); test "$i" = "$c"; echo "commit evaluado: $c (el del informe)"; cambios=$(git diff --name-only "$c" HEAD); echo "ficheros cambiados entre el commit evaluado y la cabeza:"; printf "%s\n" "$cambios"; printf "%s\n" "$cambios" | while IFS= read -r f; do case "$f" in "") ;; specs/007-h5-1-avisos-de-vigencia/*) ;; *) echo "fuera del directorio del hito: $f"; exit 1 ;; esac; done; echo "todos bajo specs/007-h5-1-avisos-de-vigencia/"'
```

Esperado: la primera imprime el resumen —`commit`, `veredicto` `aprobado`, `red` `[]`, la tasa de la eval 18 con
`planificada` `true`, `decide` `false`, `sesiones` 3 y sus `pasan`, y sus tres sesiones con su reparto de avisos y sus
motivos— y termina con la línea `true` (código 0); la segunda imprime `commit evaluado: <sha> (el del informe)`, la
lista de ficheros (vacía o solo bajo el directorio del hito) y `todos bajo specs/007-h5-1-avisos-de-vigencia/`. Cuántas
sesiones pasan no decide nada: se registra. Si la primera da `false` o la segunda falla, la aceptación no vale: la tarea
registra las dos salidas, se detiene y lo anota. Las dos salidas enteras van a `gates/evals-aceptacion.md` con el enlace,
el `databaseId`, el `headSha`, la conclusión y la duración (data-model §10).

### 11.6 Repetir por etiqueta (solo si un commit posterior cambió algo fuera del directorio del hito)

```bash
rtk proxy sh -c 'set -e; puestas=$(gh pr view 007-h5-1-avisos-de-vigencia --json labels --jq ".labels[].name"); if printf "%s\n" "$puestas" | grep -qxF evals; then gh pr edit 007-h5-1-avisos-de-vigencia --remove-label evals; fi; echo "la etiqueta evals no está puesta"'
gh pr edit 007-h5-1-avisos-de-vigencia --add-label evals
rtk proxy sh -c 'set -e; n=$(gh pr view 007-h5-1-avisos-de-vigencia --json number --jq .number); marcas=$(gh api --paginate "repos/{owner}/{repo}/issues/$n/events?per_page=100" --jq ".[] | select(.event == \"labeled\" and .label.name == \"evals\") | .created_at"); t=$(printf "%s\n" "$marcas" | sort | tail -n 1); test -n "$t"; echo "etiqueta puesta: $t"; id=$(gh run list --branch 007-h5-1-avisos-de-vigencia --event pull_request --limit 100 --json databaseId,workflowName,createdAt --jq "map(select(.workflowName == \"evals\" and .createdAt >= \"$t\")) | sort_by(.createdAt) | .[0].databaseId // empty"); test -n "$id"; echo "$id" > /tmp/kitlegal-quickstart-h51/ejecucion.txt; echo "ejecución de la repetición: $id"'
```

Esperado: `la etiqueta evals no está puesta`; la etiqueta puesta; la tercera se repite hasta que imprima
`etiqueta puesta: <instante>` y `ejecución de la repetición: <id>`. Después se ejecutan tal cual la tercera orden de
§11.3 y las de §11.4 y §11.5, que leen esa ejecución; su evidencia se añade como sección nueva y vigente de
`gates/evals-aceptacion.md`.

## Limpieza

```bash
rtk proxy chmod -R u+w /tmp/kitlegal-quickstart-h51
rtk proxy rm -r /tmp/kitlegal-quickstart-h51
rtk proxy git status --porcelain | rtk proxy grep -vE '^.. specs/007-h5-1-avisos-de-vigencia/' ; echo "fin del estado"
```

Esperado: la carpeta desaparece y el estado del árbol muestra solo `fin del estado`.
