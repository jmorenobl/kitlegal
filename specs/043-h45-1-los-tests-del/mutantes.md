# Mutantes y comprobación contra `main` de H45.1

Registro de T006 (SC-001, SC-002, SC-004, SC-005, SC-007; FR-003 a FR-010, FR-012, FR-013, FR-014; US1.1-US1.4;
plan.md, «Mutantes» y «Controles de umbral»; quickstart Q1-Q8; research D7). Una entrada por fila de la tabla «Mutantes»
del plan y por cada orden de Q1 a Q7, con la orden, lo esperado y lo visto (el principio de la salida real). Todo lo que
sigue se ejecutó en la sesión de T006 y se vio terminar; nada se da por visto de una sesión anterior.

## El árbol de trabajo (Q1-Q7)

Todas en la raíz del repositorio, en la rama `043-h45-1-los-tests-del` con T001-T005 commiteados (`HEAD` en `c7633c45`,
T005). Solo leen el árbol.

### Q1 · `make ci` (SC-001; US1.1)

- Orden: `make ci`.
- Esperado: termina con `ci: todos los controles en verde`.
- Visto: el final de la salida, `go test … ./internal/app/ ./internal/skills/ ./internal/evals/` del `skills-check` con
  `ok` en los tres, `govulncheck` con `No vulnerabilities found.`, `goreleaser check` con `1 configuration file(s)
  validated`, `gitleaks` con `no leaks found`, `go mod verify` con `all modules verified` en el módulo y en los cinco de
  `tools/`, `go mod tidy -diff` sin diferencias y `ci: todos los controles en verde`. La pasada final, tras escribir este
  registro, en «El árbol al terminar».

### Q2 · `make skills-check` (SC-002; US1.2)

- Orden: `make skills-check`.
- Esperado: `ok` en `./internal/app/`, `./internal/skills/` y `./internal/evals/`.
- Visto: `ok github.com/jmorenobl/kitlegal/internal/app 0.844s`, `ok …/internal/skills 0.615s`, `ok …/internal/evals
  0.808s`. Su rojo, con una diferencia, en «SC-002 · una línea cambiada en una referencia generada».

### Q3 · El diff contra `main` (US1.3; FR-014)

- `git diff --name-only main...HEAD -- data evals skills .github` → sin salida.
- `git diff --name-only main...HEAD -- '*.go' ':!*_test.go'` → sin salida.
- `git diff --name-only main...HEAD` → `docs/roadmap/done/H45.1-los-tests-del-molde-sin-el-estado-de-hoy.md`,
  `internal/app/testdata/script/h45-paquete-de-mas.txtar`, `internal/evals/conjunto_test.go`,
  `internal/evals/definicion_test.go`, `internal/skills/cobertura_test.go`,
  `specs/042-h45-el-molde-admite/aceptacion/paquete-de-mas.txtar` y 36 rutas de `specs/043-h45-1-los-tests-del/`
  (`checklists/`, `contracts/`, `data-model.md`, `gates/`, `plan.md`, `quickstart.md`, `research.md`, `spec.md`,
  `tasks.md`). Ninguna otra: es la lista de Q3.

### Q4 · La copia con un paquete más, en verde (SC-003; FR-010, FR-011)

- Orden: `go test -count=1 -v -run '^TestEntregaDelHito$/^h45-paquete-de-mas$' ./internal/app/`.
- Esperado: `PASS`, con `copia: hecha`, `sincronía: 0`, `orden del paso 2 de normativa, líneas más abajo: 1`,
  `tests enteros: 0`, `subtests: 0`, y los `grep`, el `stdout` y el `cmp` sin fallo.
- Visto: esas cinco líneas, en ese orden; `> grep '^\| actuaciones y procedimientos en materia de extranjería y asilo,
  salvo los del paquete de prueba \|' repo/skills/plazos-y-territorio/references/administrativo.md`, `> grep '^\|
  procedimientos sancionadores en el orden social \|' repo/skills/plazos-y-territorio/references/prueba.md`, `> ! grep
  '^\| procedimientos sancionadores en el orden social \|' …/administrativo.md`, `> grep '^- Si la pregunta es de lo
  prueba, lee \[prueba\]\(references/prueba\.md\)\.$' repo/skills/normativa/SKILL.md`, `> stdout '^orden del paso 2 de
  normativa, líneas más abajo: 1$'` y `> cmp $KITLEGAL_SKILLS/../internal/app/testdata/script/h45-paquete-de-mas.txtar
  $KITLEGAL_SKILLS/../specs/042-h45-el-molde-admite/aceptacion/paquete-de-mas.txtar`, sin fallo; `--- PASS:
  TestEntregaDelHito/h45-paquete-de-mas (7.95s)`.

### Q5 · Los dos guiones, iguales byte a byte (FR-013)

- Orden: `cmp internal/app/testdata/script/h45-paquete-de-mas.txtar specs/042-h45-el-molde-admite/aceptacion/paquete-de-mas.txtar`.
- Esperado: sin salida y código 0.
- Visto: sin salida y código 0.

### Q6 · Con `main`, el defecto en la línea 36 (US1.4; FR-005)

- `grep -n 'kitlegal boe articulo BOE-A-1978-31229 a140 --json && kitlegal graph check' skills/normativa/SKILL.md` →
  `36:  kitlegal boe articulo BOE-A-1978-31229 a140 --json && kitlegal graph check BOE-A-1978-31229 a140 --json`.
- `go test -count=1 -v -run '^TestControlesSobreLasReferenciasAMano$/^(forma-en-una-referencia-con-otros-bloques|forma-en-ninguna-referencia)$' ./internal/evals/`
  → `--- PASS: TestControlesSobreLasReferenciasAMano/forma-en-una-referencia-con-otros-bloques`, `--- PASS:
  …/forma-en-ninguna-referencia`, `ok …/internal/evals 0.413s`. Que el defecto que esperan dice `línea 36` con `main` se
  ve en el mutante «otra línea» (`expected: … línea 36 …`).

### Q7 · Los otros tres, en verde (US2, US4, US5)

- `go test -count=1 -v -run '^TestCoberturaDelRepositorio$' ./internal/skills/` → `--- PASS: TestCoberturaDelRepositorio`.
- `go test -count=1 -v -run '^TestDefinicionDelJob$/^cuatro-evals-mas$' ./internal/evals/` → `--- PASS:
  TestDefinicionDelJob/cuatro-evals-mas`.
- `go test -count=1 -v -run '^TestEvalsDelRepositorio$/^precepto-de-la-cobertura$' ./internal/evals/` → `--- PASS:
  TestEvalsDelRepositorio/precepto-de-la-cobertura`.

## El clon (Q8)

- `<clon>` es `/tmp/claude/h451`: `git clone --no-hardlinks -q . /tmp/claude/h451`. Va con la ruta escrita y no con
  `"$TMPDIR/h451"` porque la capa de permisos de esta sesión no dejó correr ninguna orden que expandiera `$TMPDIR` (ni
  `printenv TMPDIR`), así que no se pudo comprobar que `$TMPDIR` sea `/tmp/claude`; `/tmp/claude` es el directorio
  temporal escribible del sandbox, el mismo de los clones de T002-T005 (`/tmp/claude/h451-t00N`). Ningún mutante se
  aplicó al árbol de trabajo, al índice ni al historial del repositorio.
- `git -C /tmp/claude/h451 rev-parse HEAD origin/main` → `c7633c45b8651d1d7ad1ea8b8af622af8a902172` (T005) y
  `e1555a7d6b56b685060e10f3d5661bbd0f76e909`, que es `git rev-parse main` del repositorio (V13).
- Línea base, sin mutante: `go -C /tmp/claude/h451 test -count=1 ./internal/skills/ ./internal/evals/` → `ok
  …/internal/skills 0.674s`, `ok …/internal/evals 20.776s`.
- Cada mutante: aplicado con un `sed -i`, un `cp` o un `git checkout origin/main -- <fichero>` en el clon; su control,
  con `go -C /tmp/claude/h451 test -count=1 …` (salida sin condensar); deshecho con `git -C /tmp/claude/h451 checkout --
  <ruta>`, `checkout HEAD --` o `clean -fdq --`. Tras deshacer cada uno, `git -C /tmp/claude/h451 status --short` no dio
  nada.
- Las órdenes son las de Q8 con `/tmp/claude/h451` en lugar de `"$TMPDIR/h451"`, y `$s` y `$c` escritos enteros
  (`/tmp/claude/h451/internal/app/testdata/script/h45-paquete-de-mas.txtar` y
  `/tmp/claude/h451/specs/042-h45-el-molde-admite/aceptacion/paquete-de-mas.txtar`).

### L06 · fuera `data/cobertura.yaml:39` (FR-003)

- Mutante: `sed -i 39d /tmp/claude/h451/data/cobertura.yaml`; la línea 39 era `    norma: {identificador:
  BOE-A-2015-11722, nombre: "texto refundido de la Ley sobre Tráfico, …", aplicable: {bloque: a1, apartado: "1"}}`.
- Orden: `go -C /tmp/claude/h451 test -count=1 -run '^TestCoberturaDelRepositorio$' ./internal/skills/`.
- Esperado: rojo, «solo las 5 primeras».
- Visto: `--- FAIL: TestCoberturaDelRepositorio`, `Error: Not equal:`, `expected: 9`, `actual  : 5`; `Messages:
  ../../data/cobertura.yaml: de sus filas de administrativo, [{BOE-A-2015-10565 DA 1ª.2.a BOE-A-2003-23186} …], solo las
  5 primeras son, en su orden, de las nueve tuplas de contracts/datos-de-plazos.md §3 de H9, [… {BOE-A-2015-10565 DA
  1ª.2.c BOE-A-2015-11722} …]`; `FAIL …/internal/skills`.
- Deshecho: `git -C /tmp/claude/h451 checkout -- data/`; `status --short` vacío.

### L07 · `LeerCobertura` con `Norma = nil` (FR-003)

- Mutante: el `sed -i 's|^\treturn ValidarDocumentoYAML\[Cobertura\](contenido, esquema)$|…|'
  /tmp/claude/h451/internal/skills/cobertura.go` de Q8 (`git diff --stat` → `internal/skills/cobertura.go | 7 ++++++-`).
- Orden: la de L06.
- Esperado: rojo, `expected: 9`, `actual: 0`.
- Visto: `--- FAIL: TestCoberturaDelRepositorio`, `Error Trace: /tmp/claude/h451/internal/skills/cobertura_test.go:395`,
  `expected: 9`, `actual  : 0`; `Messages: ../../data/cobertura.yaml: de sus filas de administrativo,
  [{BOE-A-2015-10565 DA 1ª.2.a } {BOE-A-2015-10565 DA 1ª.2.b } …`.
- Deshecho: `git -C /tmp/claude/h451 checkout -- internal/skills/cobertura.go`; `status --short` vacío.

### L04 y L05, con la sincronía (FR-004; SC-007)

- L04: `sed -i '35,36d' /tmp/claude/h451/data/cobertura.yaml` (las líneas eran `  - procedimiento: procedimientos
  sancionadores en el orden social` y su `precepto: {norma: BOE-A-2015-10565, bloque: daprimera, apartado: "DA
  1ª.2.c"}`); sincronía, `go -C /tmp/claude/h451 test -count=1 -run '^TestSkillsDelRepositorio$' ./internal/app/ -args
  -regenerar-skills` → `ok …/internal/app 1.367s`, y `status --short` → `M data/cobertura.yaml`, `M
  skills/plazos-y-territorio/references/administrativo.md`.
- L05: `sed -i 's/procedimientos sancionadores en el orden social$/procedimientos sancionadores en el orden laboral/'
  /tmp/claude/h451/data/cobertura.yaml` (`data/cobertura.yaml | 2 +-`); la misma sincronía → `ok …/internal/app
  0.813s`, con los mismos dos ficheros cambiados.
- Órdenes, con cada uno: `go -C /tmp/claude/h451 test -count=1 -v -run '^TestEntregaDelHito$/^h32-paquetes$'
  ./internal/app/` y `go -C /tmp/claude/h451 test -count=1 -v -run '^TestCoberturaDelRepositorio$' ./internal/skills/`.
- Esperado: `h32-paquetes` en rojo, `no match for ^filas: 33 de 33$`; `TestCoberturaDelRepositorio` en verde.
- Visto, igual con L04 y con L05: `filas: 32 de 33`, `> stdout '^filas: 33 de 33$'`, `FAIL:
  testdata/script/h32-paquetes.txtar:37: no match for `^filas: 33 de 33$` found in stdout`, `--- FAIL:
  TestEntregaDelHito/h32-paquetes`; y `--- PASS: TestCoberturaDelRepositorio`, `ok …/internal/skills`.
- Deshecho, cada uno: `git -C /tmp/claude/h451 checkout -- data/ skills/`; `status --short` vacío.

### P6 · `defectosDeLasOrdenes(delSkillMd, nil)` (FR-005)

- Mutante: `sed -i 's/defectosDeLasOrdenes(delSkillMd, deLasReferencias)/defectosDeLasOrdenes(delSkillMd, nil)/'
  /tmp/claude/h451/internal/evals/conjunto_test.go` (`conjunto_test.go | 2 +-`; la llamada, en la 4797).
- Orden: `go -C /tmp/claude/h451 test -count=1 -v -run '^TestControlesSobreLasReferenciasAMano$/^(forma-en-una-referencia-con-otros-bloques|forma-en-ninguna-referencia)$' ./internal/evals/`.
- Esperado: los dos casos en rojo.
- Visto: `--- FAIL: TestControlesSobreLasReferenciasAMano/forma-en-una-referencia-con-otros-bloques` y `--- FAIL:
  …/forma-en-ninguna-referencia`; en los dos, `Error Trace: …/conjunto_test.go:5094`, `Error: Should be empty, but was
  [normativa: SKILL.md: línea 36, «kitlegal boe articulo BOE-A-1978-31229 a140 --json && kitlegal graph check
  BOE-A-1978-31229 a140 --json»: no la sigue su forma para PowerShell, «…; if ($LASTEXITCODE -eq 0) { … }», con el
  mismo verbo, la misma norma y los mismos bloques]`, `Messages: sin mutar, la copia de normativa no da ningún defecto
  del control`. Caen en la premisa del caso, antes del mutante: sin leer las referencias, la orden del paso 2 ya sale
  sin su forma.
- Deshecho: `git -C /tmp/claude/h451 checkout -- internal/evals/conjunto_test.go`; `status --short` vacío.

### Otra línea · `linea: numero + 1` en `tramosDeCodigo` (FR-005)

- Mutante: `sed -i 's/tramos = append(tramos, tramoDeCodigo{linea: numero, texto: recortada})/tramos = append(tramos,
  tramoDeCodigo{linea: numero + 1, texto: recortada})/' /tmp/claude/h451/internal/evals/conjunto_test.go` (en la 4214;
  `conjunto_test.go | 2 +-`).
- Orden: la de P6.
- Esperado: los dos casos en rojo.
- Visto: `--- FAIL: …/forma-en-ninguna-referencia` y `--- FAIL: …/forma-en-una-referencia-con-otros-bloques`; en los
  dos, `Error Trace: …/conjunto_test.go:5104`, `expected: []string{"normativa: SKILL.md: línea 36, «kitlegal boe
  articulo BOE-A-1978-31229 a140 --json && kitlegal graph check …»: no la sigue su forma para PowerShell, …"}` y
  `actual  : []string{"normativa: SKILL.md: línea 37, «…»: …"}`.
- Deshecho: `git -C /tmp/claude/h451 checkout -- internal/evals/conjunto_test.go`; `status --short` vacío.

### T1 · `plazos-y-territorio` a dos en la definición (FR-006, FR-007; SC-006)

- Mutante: `sed -i '260s/concurrencia: 3/concurrencia: 2/' /tmp/claude/h451/.github/workflows/evals.yml` (la 259 es `-
  skill: plazos-y-territorio`) y `sed -i 's/"plazos-y-territorio": {Concurrencia: 3, ObjetivoDeDuracion:
  0},/"plazos-y-territorio": {Concurrencia: 2, ObjetivoDeDuracion: 0},/' /tmp/claude/h451/internal/evals/definicion_test.go`
  (`2 files changed, 2 insertions(+), 2 deletions(-)`).
- Orden: `go -C /tmp/claude/h451 test -count=1 -v -run '^TestDefinicionDelJob$/^cuatro-evals-mas$' ./internal/evals/`.
- Esperado: rojo, «"5h52m0s" is not greater than or equal to …».
- Visto: `--- FAIL: TestDefinicionDelJob/cuatro-evals-mas`, `Error Trace: …/definicion_test.go:3471`, `Error:
  "5h52m0s" is not greater than or equal to "7h47m49s"`, `Messages: con 16 evals de plazos-y-territorio y la definición
  del repositorio, el tope cubre su peor caso, 28069 s = 485 s + 900 s + (⌈97 / 2⌉ + ⌈96 / 2⌉) × (22 s + 240 s + 10 s) +
  60 s + 6 × (35 s + 5 s)`.
- Deshecho: `git -C /tmp/claude/h451 checkout -- .github/ internal/evals/definicion_test.go`; `status --short` vacío.

### L29 · un solo modo en `plazos-y-territorio` (FR-006, FR-007; SC-006)

- Mutante: `sed -i 's/Modos:               \[\]Modo{ModoOrden, ModoHerramienta},/Modos:               []Modo{ModoOrden},/'
  /tmp/claude/h451/internal/evals/definicion.go` (en la 773; `definicion.go | 2 +-`).
- Orden: la de T1.
- Esperado: `cuatro-evals-mas` en rojo.
- Visto: `--- FAIL: TestDefinicionDelJob/cuatro-evals-mas`; `Error Trace: …/definicion_test.go:3470`, `Not equal:`,
  `expected: evals.peorCasoDelTrabajo{Decision:900000000000, SesionesPorGrupo:[]int{97, 96}, Concurrencia:3, …}`,
  `actual  : evals.peorCasoDelTrabajo{Decision:900000000000, SesionesPorGrupo:[]int{97}, Concurrencia:3, …}`; y
  `Error Trace: …/definicion_test.go:3484`, `Error: "5h52m0s" is not less than "4h10m13s"`, `Messages: con 16 evals de
  plazos-y-territorio y la skill a dos, el tope no cubre su peor caso, 15013 s = 485 s + 900 s + (⌈97 / 2⌉) × …`.
- Deshecho: `git -C /tmp/claude/h451 checkout -- internal/evals/definicion.go`; `status --short` vacío.

### FR-008 · ocho evals de más en `plazos-y-territorio`

- Mutante: ocho `cp /tmp/claude/h451/evals/plazos-y-territorio/01-territorio-municipio-cubierto.yaml
  /tmp/claude/h451/evals/plazos-y-territorio/<n>-copia.yaml`, con `<n>` de 90 a 97, uno tras otro con `&&` (el bucle
  `for` de Q8 no lo dejó correr la capa de permisos; la carpeta queda igual). `status --short` → los ocho, de `??
  evals/plazos-y-territorio/90-copia.yaml` a `?? evals/plazos-y-territorio/97-copia.yaml`.
- Orden: `go -C /tmp/claude/h451 test -count=1 -v -run '^TestDefinicionDelJob$/^(tope-con-las-evals-del-repositorio|cuatro-evals-mas)$' ./internal/evals/`.
- Esperado: `/tope-con-las-evals-del-repositorio` en rojo y `/cuatro-evals-mas` en verde.
- Visto: `--- PASS: TestDefinicionDelJob/cuatro-evals-mas`; `--- FAIL: TestDefinicionDelJob/tope-con-las-evals-del-repositorio`,
  con `--- FAIL` en `el-del-repositorio`, `medida-sin-jurisprudencia`, `jurisprudencia-de-una-en-una-en-la-medida`,
  `descripciones-de-uno-en-uno-en-la-medida` y `jurisprudencia-de-una-en-una`, y `--- PASS` en los cinco
  `por-debajo-…`. En los cinco en rojo (`Error Trace: …/definicion_test.go:1849`), el defecto de más es el mismo,
  `jobs.evals.timeout-minutes: vale 352 (21120 s), y lo esperado es al menos el peor caso de plazos-y-territorio, 23173 s
  = 485 s + 900 s + (⌈115 / 3⌉ + ⌈114 / 3⌉ + ⌈6 / 3⌉) × (22 s + 240 s + 10 s) + 60 s + 6 × (35 s + 5 s)`; en uno,
  `expected: []string(nil)` y ese defecto solo en `actual`.
- Deshecho: `git -C /tmp/claude/h451 clean -fdq -- evals/`; `status --short` vacío.

### L33 · `a999` en la eval 12 (FR-009)

- Mutante: `sed -i '15s/bloque: daprimera}/bloque: a999}/'
  /tmp/claude/h451/evals/plazos-y-territorio/12-cobertura-seguridad-social-sin-leer-la-da-primera.yaml` (la 15 era `
  precepto: {norma: BOE-A-2015-10565, bloque: daprimera}`).
- Orden: `go -C /tmp/claude/h451 test -count=1 -v -run '^TestEvalsDelRepositorio$/^precepto-de-la-cobertura$' ./internal/evals/`.
- Esperado: rojo, con el mensaje de `a999` de la 12.
- Visto: `--- FAIL: TestEvalsDelRepositorio/precepto-de-la-cobertura`; `Error Trace: …/conjunto_test.go:2935`, `Error:
  Should be empty, but was [12-cobertura-seguridad-social-sin-leer-la-da-primera.yaml: no_cubierto.precepto
  BOE-A-2015-10565, bloque a999 no es el precepto de ninguna fila de data/cobertura.yaml]`, `Messages: cada
  no_cubierto.precepto de ../../evals/plazos-y-territorio es el de una fila de ../../data/cobertura.yaml`; y en la 2949
  el mutante `a999` de la 07, con ese mismo defecto de la 12 de más en `actual`.
- Deshecho: `git -C /tmp/claude/h451 checkout -- evals/`; `status --short` vacío.

### Las tres de lo regulado aparte, pasadas a lo no declarado (FR-009)

- Mutante: `sed -i -e 's/^  norma: .*$/  fuera_de_lo_declarado: true/' -e '/^  precepto: /d'` sobre la 07, la 08 y la
  12 de `/tmp/claude/h451/evals/plazos-y-territorio/` (`3 files changed, 3 insertions(+), 6 deletions(-)`).
- Orden: la de L33.
- Esperado: rojo, `Should NOT be empty`.
- Visto: `--- FAIL: TestEvalsDelRepositorio/precepto-de-la-cobertura`; `Error Trace: …/conjunto_test.go:2932`, `Error:
  Should NOT be empty, but was []`, `Messages: ../../evals/plazos-y-territorio tiene al menos una eval con no_cubierto
  de lo regulado aparte`.
- Deshecho: `git -C /tmp/claude/h451 checkout -- evals/`; `status --short` vacío.

### Cada uno de los cuatro tests, devuelto a `main` (FR-012; SC-005)

Orden, con cada uno: `go -C /tmp/claude/h451 test -count=1 -v -run '^TestEntregaDelHito$/^h45-paquete-de-mas$'
./internal/app/`. Esperado: el guion en rojo, `tests enteros: 1` con los dos primeros y `subtests: 1` con los dos
últimos, con el test nombrado en su registro.

1. `git -C /tmp/claude/h451 checkout origin/main -- internal/skills/cobertura_test.go` → `copia: hecha`, `sincronía:
   0`, `tests enteros: 1`, `subtests: 0`; `tests enteros: go test -count=1 -run ^(TestLeerCobertura|TestCoberturaDelRepositorio|…|TestControlesSobreLasReferenciasAMano)$
   ./internal/skills/ ./internal/evals/: exit status 1`, `--- FAIL: TestCoberturaDelRepositorio (0.01s)`; `FAIL:
   testdata/script/h45-paquete-de-mas.txtar:55: no match for `^tests enteros: 0$` found in stdout`. Deshecho con
   `git -C /tmp/claude/h451 checkout HEAD -- internal/skills/cobertura_test.go`.
2. `sed -i 's/lineaDeLaOrdenDelPasoDos(t), bashDelPasoDos, sinSuForma/36, bashDelPasoDos, sinSuForma/'
   /tmp/claude/h451/internal/evals/conjunto_test.go` (una línea, la 5324: los dos casos comparten
   `sinSuFormaEnElSkillMd`) → `copia: hecha`, `sincronía: 0`, `orden del paso 2 de normativa, líneas más abajo: 1`,
   `tests enteros: 1`, `subtests: 0`; `--- FAIL: TestControlesSobreLasReferenciasAMano (0.67s)`, `--- FAIL:
   TestControlesSobreLasReferenciasAMano/forma-en-una-referencia-con-otros-bloques`, con `expected: … línea 36 …` y
   `actual  : … línea 37 …`; `FAIL: …h45-paquete-de-mas.txtar:55: no match for `^tests enteros: 0$``. El programa
   imprime solo el principio de la salida de `go test`, y en él sale el primero de los dos casos. Deshecho con
   `git -C /tmp/claude/h451 checkout -- internal/evals/conjunto_test.go`.
3. `git -C /tmp/claude/h451 checkout origin/main -- internal/evals/definicion_test.go` → las cinco líneas con `tests
   enteros: 0` y `subtests: 1`; `subtests: go test -count=1 -run ^(TestEvalsDelRepositorio|TestCasosDePlazos|TestEntregaDelHito|TestDefinicionDelJob)$/^(precepto-de-la-cobertura|…|cuatro-evals-mas)$
   ./internal/evals/ ./internal/app/: exit status 1`, `--- FAIL: TestDefinicionDelJob/cuatro-evals-mas`, `Error:
   "5h52m0s" is not greater than or equal to "6h44m21s"`; `FAIL: …h45-paquete-de-mas.txtar:56: no match for
   `^subtests: 0$` found in stdout`. Deshecho con `git -C /tmp/claude/h451 checkout HEAD --
   internal/evals/definicion_test.go`.
4. `sed -i 's/ && !eval.NoCubierto.FueraDeLoDeclarado$//' /tmp/claude/h451/internal/evals/conjunto_test.go` (la 2963)
   → las cinco líneas con `tests enteros: 0` y `subtests: 1`; `--- FAIL:
   TestEvalsDelRepositorio/precepto-de-la-cobertura`, `Error: Should be empty, but was [95-fuera-de-lo-declarado.yaml:
   no_cubierto.precepto , bloque  no es el precepto de ninguna fila de data/cobertura.yaml]`; `FAIL:
   …h45-paquete-de-mas.txtar:56: no match for `^subtests: 0$``. Deshecho con `git -C /tmp/claude/h451 checkout --
   internal/evals/conjunto_test.go`.

Tras cada uno, `status --short` vacío.

### La copia congelada distinta (FR-013)

- Mutante: `printf '\n' >> /tmp/claude/h451/specs/042-h45-el-molde-admite/aceptacion/paquete-de-mas.txtar`.
- Orden: la de SC-005.
- Esperado: el guion en rojo, «… differ».
- Visto: las cinco líneas como en verde (`tests enteros: 0`, `subtests: 0`, `líneas más abajo: 1`); `> cmp
  $KITLEGAL_SKILLS/../internal/app/testdata/script/h45-paquete-de-mas.txtar
  $KITLEGAL_SKILLS/../specs/042-h45-el-molde-admite/aceptacion/paquete-de-mas.txtar`, `FAIL:
  testdata/script/h45-paquete-de-mas.txtar:75: /tmp/claude/h451/skills/../internal/app/testdata/script/h45-paquete-de-mas.txtar
  and /tmp/claude/h451/skills/../specs/042-h45-el-molde-admite/aceptacion/paquete-de-mas.txtar differ`.
- Deshecho: `git -C /tmp/claude/h451 checkout -- specs/`; `status --short` vacío.

### FR-010 (1) · `instalar` sin la llamada a `cambiarUnaVez`, igual en los dos guiones

Cada paso, sobre los dos guiones del clon a la vez; orden, la de SC-005.

- `sed -i 's/cambiarUnaVez(filepath.Join(datos, "cobertura.yaml"), cambiosDeLaBase)/error(nil)/' <s> <c>` (la 287; `2
  files changed, 2 insertions(+), 2 deletions(-)`) → las cinco líneas como en verde; `FAIL:
  testdata/script/h45-paquete-de-mas.txtar:66: no match for `^\| actuaciones y procedimientos en materia de extranjería
  y asilo, salvo los del paquete de prueba \|` found in repo/skills/plazos-y-territorio/references/administrativo.md`.
- `sed -i '/^grep .*salvo los del paquete de prueba/d' <s> <c>` → `FAIL: …:66: no match for `^\| procedimientos
  sancionadores en el orden social \|` found in repo/skills/plazos-y-territorio/references/prueba.md`.
- `sed -i '/^grep .*procedimientos sancionadores en el orden social/d' <s> <c>` → `FAIL: …:66: unexpected match for
  `^\| procedimientos sancionadores en el orden social \|` found in
  repo/skills/plazos-y-territorio/references/administrativo.md: | procedimientos sancionadores en el orden social |`.
- Esperado: esas tres, una tras otra. Visto: las tres, con `--- FAIL: TestEntregaDelHito/h45-paquete-de-mas` en cada
  pasada.
- Deshecho: `git -C /tmp/claude/h451 checkout -- internal/app/testdata/ specs/`; `status --short` vacío.

### FR-010 (2) · `instalar` sin copiar `paquete/prueba-de-normativa.md`, igual en los dos guiones

- `sed -i 's|copiarFichero(filepath.Join("paquete", "prueba-de-normativa.md"), filepath.Join(copia, "skills",
  "normativa", "references", "prueba.md"))|error(nil)|' <s> <c>` (la 296) → `copia: hecha`, `sincronía: 0`, `orden del
  paso 2 de normativa, líneas más abajo: 0`, `tests enteros: 0`, `subtests: 0`; `FAIL: …:71: no match for `^- Si la
  pregunta es de lo prueba, lee \[prueba\]\(references/prueba\.md\)\.$` found in repo/skills/normativa/SKILL.md`.
- `sed -i '/^grep .*repo\/skills\/normativa\/SKILL\.md$/d' <s> <c>` → `FAIL: …:71: no match for `^orden del paso 2 de
  normativa, líneas más abajo: 1$` found in stdout`.
- Esperado: la salida con `0`; el `grep` de la línea del paquete y, sin él, el `stdout`. Visto: eso.
- Deshecho: `git -C /tmp/claude/h451 checkout -- internal/app/testdata/ specs/`; `status --short` vacío.

### FR-010 (3) · un viejo de `cambiosDeLaBase` que no está una vez en `data/cobertura.yaml`

- Mutante: `sed -i 's/^  - procedimiento: actuaciones y procedimientos en materia de extranjería y asilo$/  -
  procedimiento: actuaciones y procedimientos en materia de extranjería y de asilo/' /tmp/claude/h451/data/cobertura.yaml`
  (la 42; `data/cobertura.yaml | 2 +-`).
- Orden: la de SC-005.
- Esperado: el `exec go run` sale con 1 y su salida de errores dice «no lleva una sola vez».
- Visto: `> exec go run paquete-de-mas.go $KITLEGAL_SKILLS/.. repo`, `paquete-de-mas: instalación:
  $WORK/repo/data/cobertura.yaml no lleva una sola vez "  - procedimiento: actuaciones y procedimientos en materia de
  extranjería y asilo\n"`, `[exit status 1]`, `FAIL: testdata/script/h45-paquete-de-mas.txtar:50: unexpected command
  failure`.
- Deshecho: `git -C /tmp/claude/h451 checkout -- data/`; `status --short` vacío.

### SC-002 · una línea cambiada en una referencia generada

- Mutante: `sed -i '34s/^| procedimientos sancionadores en el orden social |/| procedimientos sancionadores en el orden
  laboral |/' /tmp/claude/h451/skills/plazos-y-territorio/references/administrativo.md`; la 34 está entre `<!-- inicio de
  cobertura_por_materia: generada desde data/cobertura.yaml y data/normas.yaml con make skills-sync, no editar -->` (la
  28) y `<!-- fin de cobertura_por_materia -->` (la 40). `git diff --stat` → `administrativo.md | 2 +-`.
- Orden: `make -C /tmp/claude/h451 skills-check`.
- Esperado: rojo, con una diferencia.
- Visto: `--- FAIL: TestSkillsDelRepositorio/skills`, `Error Trace: /tmp/claude/h451/internal/app/skills_test.go:106`,
  `Error: Should be empty, but was [plazos-y-territorio: references/administrativo.md: contenido-distinto]`; los casos
  de `TestSkillsDelRepositorio/scripts/consulta-juridica/…`, en rojo con esa misma diferencia de más; `FAIL
  …/internal/app`, `ok …/internal/skills`, `ok …/internal/evals`, `make: *** [Makefile:137: skills-check] Error 1`.
- Deshecho: `git -C /tmp/claude/h451 checkout -- skills/`; `status --short` vacío.

### El clon, al final

`git -C /tmp/claude/h451 status --short` → sin salida. El `rm -rf /tmp/claude/h451` del final de Q8 lo denegó la capa de
permisos de esta sesión: el clon, limpio, sigue en el temporal y no está en el repositorio.

## El árbol al terminar

- `git diff --stat` → solo `specs/043-h45-1-los-tests-del/gates/modelos.jsonl`, `gates/tarea-actual.json` y
  `gates/tareas-intentos.json`, del workflow, que ya estaban cambiados al empezar la sesión (no los tocó ninguna orden de
  esta tarea). `git diff --stat HEAD -- data evals skills .github internal` → sin salida.
- `git status --short --untracked-files=all` → esos tres, `?? specs/043-h45-1-los-tests-del/mutantes.md` y los
  marcadores de 3 bytes y modo 666 que pone el sandbox en las rutas que no deja escribir (`.bash_profile`, `.bashrc`,
  `.gitconfig`, `.gitmodules`, `.idea`, `.profile`, `.ripgreprc`, `.vscode`, `.zprofile`, `.zshrc` y ocho bajo
  `.claude/`: `agents`, `commands`, `hooks`, `launch.json`, `loop.md`, `output-styles`, `routines`, `workflows`); no los
  escribió ninguna orden de esta tarea y no son del cambio. Lo que T006 cambia es solo `mutantes.md` (y la marca de T006
  en `tasks.md`).
- `make ci`, tras escribir este registro (filtrado a fallos y cierre): `0 issues.`, `No vulnerabilities found.`, `1
  configuration file(s) validated`, `no leaks found`, ningún `FAIL` ni `--- FAIL`, y `ci: todos los controles en
  verde`.
