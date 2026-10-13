# H26 · Cierre de la Definition of Done

Cierre de T005 (FR-020, FR-021, FR-035, FR-036, FR-055, FR-060, FR-061, FR-062, FR-070, FR-071; SC-010, SC-011) sobre
`44f2315`, que es T004, el último commit antes de T005. La rama está al día con `main`: `git merge-base main HEAD` da
la cabeza de `main`, `869eaf8`, así que `git diff --name-status main` da exactamente lo que cambia el hito. Las cinco
partes son las de la tarea: lo creado, lo modificado y lo que no cambia (§1), el quickstart (§2), la cobertura (§3),
los controles de umbral (§4) y los umbrales, el instrumento y los atajos (§5).

Todo lo que se cita aquí se ejecutó en la sesión de T005, en primer plano, y se vio terminar, o está en el
repositorio. Lo que no se pudo medir se dice tal cual (§2, «Cómo se ejecutó» y los escenarios 6, 9 y 10; §3, «El
diff» y lo que `main` cubría; §4, las dieciocho filas `evals:`; §5, «Sin modelo»; §6).

**T005 no toca código ni añade ningún test.** Las cifras de cobertura quedan sobre sus umbrales (§3), y la tarea solo
pide tests si alguna queda debajo: `internal/evals/umbrales.go`, que la tarea declara para ese caso, no cambia. Sus
únicos cambios son este fichero, dos líneas en `gates/supuestos.md` y su marca en tasks.md. Las cinco partes cuadran
con lo que la tarea pide comprobar, y las diecisiete órdenes de quickstart.md §0 a §5, §7 y §8 dan lo esperado (§2).

## 1. Lo creado y lo modificado en el hito, y lo que no cambia

`git diff --name-status main` da **44 ficheros: 31 `A` y 13 `M`**, ninguna `D` ni `R`. Los 31 `A` están todos en
`specs/021-h26-el-umbral-de/`; este `cierre.md` no está entre ellos, porque no tiene seguimiento hasta el commit de la
tarea, y con él serán 45. Los 13 `M` son todo lo que el hito cambia fuera de `specs/`
(`git diff --shortstat main -- . ':!specs'`: 13 ficheros, 1 071 líneas añadidas y 313 quitadas). El hito no crea
ningún fichero fuera de su directorio. `git status --porcelain` da, antes de §2 y después de las medidas de este
fichero, los dos que el workflow lleva modificados (`gates/tarea-actual.json` y `gates/tareas-intentos.json`), y
`git ls-files --others --exclude-standard` no daba ninguno antes de escribirlo.

### Los 13 ficheros

Cada uno con la tarea que lo tocó (`git log --format=%s main..HEAD -- <ruta>`) y sus líneas
(`git diff --numstat main`). Son los trece de plan.md, «Source Code», y ninguno más.

**Las dos declaraciones de clases** (T002):

| Fichero | Líneas | Qué cambia |
|---|---|---|
| `evals/boe-legislacion/juez/clases.yaml` | +3 −2 | `umbral: 0` pasa a `umbral: 0.028` en `afirma_lo_no_leido`, con la frase de su cabecera (§5). |
| `evals/jurisprudencia/juez/clases.yaml` | +3 −1 | `umbral: 0` pasa a `umbral: 0.017` en `afirma_lo_no_leido`, con la frase de su cabecera (§5). |

**Los ficheros de Go del paquete de evals** (6, todos en `internal/evals/`):

| Fichero | Tarea | Líneas | Qué cambia |
|---|---|---|---|
| `umbrales.go` | T001 | +179 −82 | `umbralesConjuntos` y `nombreDelUmbralConjunto`, nuevos; `Decide: false` en el recuento por modo; el motivo del conjunto, con el modo de cada respuesta; y sus comentarios. |
| `doc.go` | T001 | +28 −13 | Solo comentarios. |
| `informe.go` | T001 | +38 −16 | Solo comentarios. |
| `umbrales_test.go` | T001, T002 | +435 −38 | `TestUmbralConjunto`, nuevo; el arnés y los umbrales esperados. |
| `informe_test.go` | T001, T002 | +294 −111 | Los tests del informe con el juez y los de `jurisprudencia`. |
| `conjunto_test.go` | T002 | +22 −13 | Los dos valores de `TestEvalsDelRepositorio`. |

Son 245 líneas añadidas y 111 quitadas en los tres que no son tests, y 751 y 162 en los tres de test
(`git diff --shortstat main`, sobre cada trío). Ningún fichero ni paquete de Go es nuevo. En `doc.go` e `informe.go`,
toda línea añadida o quitada empieza por `//`: la búsqueda de las que no lo hacen da 0. `umbrales.go` importa lo que
en `main`, `fmt`, `strconv` y `strings`; `go.mod` y `go.sum` no cambian; y `gofmt -l internal/evals/` no nombra ningún
fichero.

**Los cinco documentos** (T004): `CHANGELOG.md` (+31 −11), `CONTRIBUTING.md` (+30 −19), `docs/JURISPRUDENCIA.md`
(+5 −4), `docs/WORKFLOW.md` (+1 −1) y `CLAUDE.md` (+2 −2). En `CHANGELOG.md`, la cabecera `## [0.5.0]` está en la
línea 285 de `main` y en la 305 del árbol, veinte más abajo, que son las que el hito añade encima; las ocho cabeceras
de su diff quedan por debajo de la 285 de `main` (§2, escenario 7).

### Lo que no cambia (FR-021, FR-036, FR-055, FR-062)

Primero, las órdenes de quickstart.md §5, tal cual (§2): la segunda, sobre `skills schemas evidencias .github scripts
.specify docs/ROADMAP.md docs/ADR docs/USO.md README.md web cmd`, **no imprime nada**, y la tercera, sobre
`evals internal`, da las dos `clases.yaml` y los seis `.go` de `internal/evals/`, y nada más.

Después, la misma comprobación —`git diff --name-status main -- <rutas>`, que **no imprime nada**— sobre cada cosa
que la tarea lista, también las que esas dos órdenes no nombran. Entre paréntesis, los ficheros versionados que hay
en esas rutas (`git ls-files`), para que el cero no sea el de una ruta vacía; y, donde se dice, el árbol, que
`git rev-parse` da igual en `main` y en la cabeza:

- **Ninguna skill**: `skills` (6), con el mismo árbol, `5d4e8147…`. Sus `SKILL.md` tienen 299, 197 y 190 líneas
  (`boe-legislacion`, `jurisprudencia` y `legal-core`).
- **Ninguna eval**: `evals` sin las dos `clases.yaml` (44 de sus 46: 27 en `evals/boe-legislacion`, 15 en
  `evals/jurisprudencia` y 4 en `evals/legal-core`, que sigue sin carpeta del juez y con el mismo árbol,
  `20d74b03…`). En las dos carpetas del juez, la rúbrica, el esquema de la respuesta, los casos y la medida son los de
  `main` (§5).
- **El directorio de esquemas**: `schemas` (16), con el mismo árbol, `3d689100…`.
- **La carpeta de la evidencia**: `evidencias` (151), con el mismo árbol, `e033857d…`.
- **Ningún fixture ni grabación**: el `testdata/` de la raíz (62); todo lo que hay bajo un `testdata/` en cualquier
  directorio (`'*testdata/*'`, 622); y los manifiestos `grabaciones.json` con sus `grabacion_test.go` (4). Tampoco
  `docs/SOURCES.md`. El paso `grabar_datos` no tenía nada que grabar.
- **La definición del job**: `.github` (7), con el mismo árbol, `b15e7083…`. `evals.yml` sigue con
  `MODELO_DE_EVALS: claude-sonnet-5-5`, `MODELO_DEL_JUEZ: claude-opus-5-5`, `REPETICIONES_DE_EVALS: 3`,
  `VERSION_DE_CLAUDE_CODE: 2.1.284` y `VERSION_DE_CLAUDE_CODE_DEL_JUEZ: 2.1.289`, y con su comentario de las líneas
  152 a 154, que dice «doce umbrales, diez que deciden»: el hito lo deja falso y no lo puede tocar (§6).
- **Los guiones de evals y los del workflow**: `scripts` (31), con el mismo árbol, `96622818…`; dentro,
  `scripts/workflow` (16) y `scripts/evals*` (6).
- **La constitución**: `.specify` (45), con el mismo árbol, `6578722f…`.
- **El roadmap, los ADR, la bitácora de uso, el léeme y la web**: `docs/ROADMAP.md`, `docs/ADR` (38, con el mismo
  árbol, `4dcdf7ae…`), `docs/USO.md`, `README.md` y `web` (66, con el mismo árbol, `c785b74a…`). De `docs/` (48), el
  diff solo nombra `docs/JURISPRUDENCIA.md` y `docs/WORKFLOW.md`.
- **Los artefactos de otros hitos**: `-- specs ':!specs/021-h26-el-umbral-de'` (1 060).
- **El binario fuera del paquete de evals**: `cmd` (3, con el mismo árbol, `74d1e211…`); `internal` sin
  `internal/evals` (551), con el mismo árbol en `internal/core` (`0ab0f68b…`), `internal/app` (`4807f918…`) e
  `internal/cli` (`64d54cc8…`); `data` (26, `b340039d…`); `skills.go`; `go.mod`, `go.sum`, `tools` y `Makefile` (13);
  y `.golangci.yml`, `codecov.yml` y `.goreleaser.yaml`.
- **El resto del paquete de evals**: `internal/evals` sin los seis ficheros de arriba (414 de sus 420). Entre ellos,
  `sondeo.go`, `juez.go`, `juzgar.go`, `conjunto.go`, `medida.go` y `definicion.go`, y los tests que la tabla de
  controles nombra y el hito no cambia: `sondeo_test.go`, `medida_test.go`, `ejecucion_test.go` y
  `definicion_test.go` (§4).

La misma orden sí imprime donde hay cambios: con `-- internal/evals` salen 6 líneas; con `-- evals`, 2; y con los
cinco documentos, 5.

`git worktree list` da el árbol y dos worktrees de un scratchpad de otra sesión, bajo `$TMPDIR` (`wt-antes` y
`wt-proto`), que no son de este run y que esta sesión no ha tocado. El de los mutantes de T003 no está.

### La Definition of Done, punto a punto (`ROADMAP.md` §1)

| Punto | Aplica | Qué se ve en la cabeza |
|---|---|---|
| 1 · `make ci` en verde | sí | §2, escenario 8: `ci: todos los controles en verde`, con la caché de tests vacía. |
| 2 · tests offline; fixtures si toca red | tests sí; fixtures no | Los tests de §2 y §4. El hito no toca la red: ninguna grabación ni fuente (arriba). |
| 3 · sin `net/http`, `os.Exit`, `fmt.Print*` fuera de lo autorizado | sí | En las 245 líneas añadidas a los tres `.go` que no son tests, 0 coincidencias de `"net/http"`, `"net"`, `os.Exit`, `fmt.Print`, `os.Stdout`, `os.Stderr`, `panic(`, `"database/sql"`, `modernc.org` e `internal/httpx`. `golangci-lint`, en el `make ci` de §2: `0 issues.` |
| 4 · esquemas y `schema-check` | ningún esquema cambia | `schemas` sin cambios; `make schema-check`, `ok` (§2, escenario 1, y dentro de `make ci`). |
| 5 · errores con código estable | ningún código de `kitlegal` cambia | `cmd`, `internal/cli`, `internal/core` e `internal/app` sin cambios. |
| 6 · e2e y `CHANGELOG.md` | «Aceptación e2e: no aplica»; `CHANGELOG.md`, sí | `CHANGELOG.md` +31 −11, en *Unreleased*, con la entrada del umbral conjunto en su línea 289 (T004; SC-010). |
| 7 · ADR | no | `docs/ADR` sin cambios: la decisión es el ADR 0038. |
| 8 · `SOURCES.md` | no | `docs/SOURCES.md` sin cambios: ninguna fuente. |
| 9 · cobertura | sí | §3. |
| 10 · skill: evals antes, `SKILL.md` < 300, sin drift | el hito no entrega ni cambia ninguna skill | `skills` y las evals sin cambios; 299, 197 y 190 líneas; `skills-check` en verde dentro de `make ci` (§2). |
| 11 · dimensión territorial | no | — |
| 12 · operaciones de grafo | no | Ningún applet cambia. |
| 13 · recursos, plazos o escritos | no | — |

## 2. Quickstart (§0 a §5, §7 y §8)

Los ocho escenarios se ejecutaron en su orden, en primer plano, sobre `44f2315` y con el árbol limpio salvo los dos
ficheros de `gates/`. Son diecisiete órdenes, y **las diecisiete dan lo esperado**. Cada una se ejecutó sola y con
`rtk proxy` delante, que da la salida sin resumir.

| § | Orden | Esperado | Obtenido | ¿Cuadra? |
|---|---|---|---|---|
| 0 | `go version` | Go 1.27 | `go version go1.27.2 darwin/arm64` | sí |
| 1 | `grep -n '^    umbral:' evals/boe-legislacion/juez/clases.yaml evals/jurisprudencia/juez/clases.yaml` | cuatro líneas: `umbral: 0.028` y `umbral: 0` en `boe-legislacion`, `umbral: 0.017` y `umbral: 0` en `jurisprudencia` | las cuatro: `…/boe-legislacion/juez/clases.yaml:9:    umbral: 0.028` y `:12:    umbral: 0`; `…/jurisprudencia/juez/clases.yaml:9:    umbral: 0.017` y `:12:    umbral: 0` | sí |
| 1 | `git apply --reverse --check specs/021-h26-el-umbral-de/contracts/clases-yaml.diff` | sin salida y código 0 | sin salida, código 0 | sí |
| 1 | `go test -count=1 -run '^TestEvalsDelRepositorio$' ./internal/evals/` | `ok` | `ok  	…/internal/evals	0.656s` | sí |
| 1 | `make schema-check` | `ok` | `go test -count=1 -run '^TestEsquemasPublicados$' ./internal/app/` y `ok  	…/internal/app	2.009s` | sí |
| 2 | `go test -count=1 -run '^TestUmbralConjunto$' ./internal/evals/` | `ok` | `ok  	…/internal/evals	1.484s` | sí |
| 3 | `go test -count=1 -run '^(TestUmbralesDelInforme\|TestUmbralesConJuezYConSentencias\|TestInformeConElJuez\|TestInformeMarkdownDeLosUmbrales\|TestInformeEnDosModos\|TestInformeConSesionesSinMedir)$' ./internal/evals/` | `ok` | `ok  	…/internal/evals	2.244s` | sí |
| 4 | `go test -count=1 -run '^TestUmbralesDeJurisprudencia$' ./internal/evals/` | `ok` | `ok  	…/internal/evals	1.172s` | sí |
| 5 | `go test -count=1 -run '^(TestCopiasDelJuez\|TestMedidaVersionada\|TestEjecucionSinMedir\|TestDefinicionDelJob\|TestJuicioDelSondeo\|TestSalidaDelSondeo\|TestSondear)$' ./internal/evals/` | `ok` | `ok  	…/internal/evals	7.048s` | sí |
| 5 | `git diff --name-only main -- skills schemas evidencias .github scripts .specify docs/ROADMAP.md docs/ADR docs/USO.md README.md web cmd` | sin salida | sin salida, código 0 | sí |
| 5 | `git diff --name-only main -- evals internal` | las dos `clases.yaml` y ficheros `.go` de `internal/evals/`: ninguno bajo `testdata/`, ninguna eval y nada de otro paquete | ocho líneas: las dos `clases.yaml` y `conjunto_test.go`, `doc.go`, `informe.go`, `informe_test.go`, `umbrales.go` y `umbrales_test.go`, de `internal/evals/` | sí |
| 7 | `grep -n -e 'umbral 0: la respuesta' -e 'diez de ellos decidiendo' -e 'deciden diez' -e 'otros doce' -e 'en el otro no haya ninguna' -e 'herramienta: 1 de 54' -e 'umbral 0 en cada modo' -e 'sigue con el umbral 0' CONTRIBUTING.md docs/JURISPRUDENCIA.md docs/WORKFLOW.md CLAUDE.md` | sin salida | sin salida; código 1, el de `grep` sin coincidencias | sí |
| 7 | `sed -n '/^## \[Unreleased\]/,/^## \[0\.5\.0\]/p' CHANGELOG.md \| grep -n -e '(0, decide)' -e 'decide\*\* con 0' -e '^    con 0: la respuesta' -e 'diez de ellos decidiendo' -e 'en un modo da `fallo`' -e 'Los doce de'` | sin salida | sin salida; código 1, el de `grep` sin coincidencias | sí |
| 7 | ``grep -c 'afirma_lo_no_leido:<modelo>`' CONTRIBUTING.md docs/WORKFLOW.md CHANGELOG.md`` | un número mayor que 0 en cada uno de los tres | `CONTRIBUTING.md:2`, `docs/WORKFLOW.md:1` y `CHANGELOG.md:3` | sí |
| 7 | `git diff --stat main -- CHANGELOG.md CONTRIBUTING.md docs/JURISPRUDENCIA.md docs/WORKFLOW.md CLAUDE.md` | los cinco ficheros | los cinco, y `5 files changed, 69 insertions(+), 37 deletions(-)` | sí |
| 7 | `git diff -U0 main -- CHANGELOG.md \| grep '^@@'` | todas las cabeceras, con su primer número por debajo de 285 | ocho cabeceras: `-129`, `-140,2`, `-146`, `-214,2`, `-224,2`, `-227`, `-231,2` y `-283,0` | sí |
| 8 | `make ci` | en verde, con `schema-check`, `skills-check` y las reglas del conjunto | `ci: todos los controles en verde`, con la caché de tests vacía (`go clean -testcache` justo antes). Registro de 102 líneas: 53 `ok`, ningún `FAIL` ni `(cached)`; `0 issues.` del lint; `No vulnerabilities found.` y `Your code is affected by 0 vulnerabilities.`; `ok` en `TestEsquemasPublicados` y en los tres paquetes de `skills-check`, que lleva `TestEvalsDelRepositorio` con las reglas del conjunto; `1 configuration file(s) validated`; `no leaks found`; `all modules verified`, seis veces; `go mod tidy -diff` sin salida | sí |

`govulncheck` dice además que hay dos vulnerabilidades en módulos requeridos que el código no llama. Es la misma línea
que trae el registro de la verificación de T004 (`gates/ci.log`), y no hace fallar el control. En ese `make ci`,
`go test` imprime 97,2 % y 97,7 % de cobertura para `internal/evals`, en `make test` y en `make test-integration`.

**Los tests, por su nombre.** Los dieciséis tests de las órdenes de `go test` de §1 a §5 se ejecutaron además juntos,
en una sola orden con `-v`, para contar sus subpruebas y que ninguna pasara en vacío: 312 líneas `--- PASS`, ningún
`--- FAIL` y ningún `--- SKIP`. Las cifras son líneas `--- PASS` en todos los niveles, contando el propio test:

- §1, 16: `TestEvalsDelRepositorio`, con `formato` y `conjunto-jurisprudencia`, que llevan los dos literales
  (`conjunto_test.go:1933`, `Umbral: 0.028`, y `:1974`, `Umbral: 0.017`).
- §2, 8: `TestUmbralConjunto`, con sus siete casos: `3-y-0`, `0-y-3`, `2-y-1`, `4-y-0`, `2-y-2`, `104-con-2` y
  `104-con-3`.
- §3, 58: `TestUmbralesDelInforme` (21, con `sin-activar-solo-en-orden`, `sin-activar-solo-en-herramienta`,
  `901-s-en-orden-y-900-en-herramienta`, `901-s-en-herramienta-y-900-en-orden` y `la-medida-del-juez-sin-cumplirse`),
  `TestUmbralesConJuezYConSentencias` (1, sin subpruebas), `TestInformeConElJuez` (16, con `un-voto-que-no-llega`,
  `cuenta-su-proceso-con-tres`, `la-sin-binario-ni-servidor-marcada`, `dos-marcadas-en-herramienta` y
  `901-s-del-juez-en-orden-y-900-en-herramienta`), `TestInformeMarkdownDeLosUmbrales` (8, con
  `las-trece-filas-de-los-dos-modos` y `las-filas-del-contrato`), `TestInformeEnDosModos` (6, con `legal-core`) y
  `TestInformeConSesionesSinMedir` (6).
- §4, 12: `TestUmbralesDeJurisprudencia`, con sus once casos: `el-elemento-del-contrato`, `ninguna-cuenta`,
  `una-marcada-en-el-modo-orden`, `una-marcada-en-las-dos-clases`, `dos-marcadas-en-el-modo-orden`,
  `una-marcada-en-cada-modo`, `tres-con-si-en-afirma-que-existe`, `901-s-del-juez-en-el-modo-orden`,
  `una-sin-documento-en-el-modo-orden`, `dos-sin-documento-en-el-modo-herramienta` y `una-sin-activar`.
- §5, 218: `TestCopiasDelJuez` (21), `TestMedidaVersionada` (25), `TestEjecucionSinMedir` (25),
  `TestDefinicionDelJob` (97), `TestJuicioDelSondeo` (31), `TestSalidaDelSondeo` (9) y `TestSondear` (10). Los seis
  primeros dan las mismas cifras que dejó escritas el cierre de H25 tras sus reparaciones
  (`specs/020-h25-el-juez-de/cierre.md`, §4), y ninguno de sus ficheros de test está en el diff (§1).

**Cómo se ejecutó.** Tres cosas difieren de teclear el quickstart en una terminal:

- **Los códigos de salida.** En la sesión de un paso, una orden con `$?` o con `$(…)` pide una aprobación que nadie
  puede dar, y no se ha buscado otra forma de pasarla. El código de cada orden es el que da la herramienta de la
  sesión —no dice nada con 0, y dio 1 en las dos búsquedas sin coincidencias de §7— o una sonda encadenada con `&&`
  (`git apply --reverse --check` y las dos órdenes de `git diff` de §5). **El de `make ci` no se recogió**: su salida
  fue a un registro, y lo que se tiene es su última línea, `ci: todos los controles en verde`, que `make` solo
  escribe si todos los objetivos anteriores terminan bien.
- El registro de `make ci`, el de la ejecución con `-v` y los perfiles filtrados de §3 se guardaron en un directorio
  bajo `$TMPDIR`. Ninguno está en el repositorio.
- `coverage.out` y `coverage-integration.out`, que deja §8 en la raíz, los ignora git (`git check-ignore`:
  `.gitignore:6` y `:5`), y también `gates/ci.log` (`.gitignore:22`).

**§6, §9 y §10, no ejecutados aquí.**

- **§6, los dos mutantes: lo ejecutó T003**, y su registro es `specs/021-h26-el-umbral-de/mutantes.md`: 4 de 4, cada
  mutante con `TestUmbralConjunto` y `TestUmbralesDeJurisprudencia` en rojo y `go test` con código 1, sobre `8ae2405`
  (T002). El `umbrales.go` que T003 mutó es el de hoy: su registro da el blob `d52a8aa`, y es el que dan
  `git rev-parse HEAD:internal/evals/umbrales.go` y `git hash-object` del árbol. Las dos líneas de las que cuelgan
  las sustituciones siguen donde T003 las encontró: `range j.deLosModos() {` en la 349 y `Decide:      true,` en la
  360. Esta sesión no ha repetido el escenario.
- **§9, el job de cierre: lo hace el workflow** tras la revisión final. Empuja la rama, abre la propuesta de cambio,
  pone la etiqueta `evals` y guarda el informe de cada skill en `gates/evals/`, que hoy no existe (SC-001, SC-011).
- **§10, después del run: una persona** lee las respuestas marcadas del cierre y decide la release (SC-012), corrige
  fuera del hito el comentario de `evals.yml` y escribe la viñeta de H26 de `CLAUDE.md`.

## 3. Cobertura (Definition of Done §1.9)

Se mide sobre los dos perfiles que deja el `make ci` de §2, escenario 8, en verde y con la caché de tests vacía (el
registro no trae ningún `(cached)`): `coverage.out` (`make test`) y `coverage-integration.out`
(`make test-integration`). El porcentaje sale de `go tool cover -func`: el global, sobre cada perfil; el de cada árbol
o fichero, sobre el perfil filtrado a sus líneas con `grep`, con su cabecera `mode:`, escrito en un directorio bajo
`$TMPDIR`. Los recuentos de sentencias salen de los bloques del perfil —su segundo campo—, contando cada bloque una
vez, con un `awk` en línea sobre los dos perfiles del árbol.

| Árbol | `coverage.out` | `coverage-integration.out` | Umbral |
|---|---|---|---|
| Global | **96,5 %** (10 728 de 11 123 sentencias) | **96,9 %** (10 777 de 11 123) | ≥ 70 % (Definition of Done §1.9; `codecov.yml`) |
| `internal/core/**` (dominio) | **98,3 %** (2 100 de 2 136) | **98,3 %** (2 100 de 2 136) | ≥ 85 % (Definition of Done §1.9; `codecov.yml`) |
| `internal/cli/**` | 98,0 % (398 de 406) | 98,0 % (398 de 406) | ≥ 90 % (`codecov.yml`) |
| `internal/evals` (el paquete de evals) | **97,3 %** (3 800 de 3 908) | **97,7 %** (3 817 de 3 908) | — |
| `internal/evals/umbrales.go` | 99,2 % (124 de 125) | 99,2 % (124 de 125) | — |
| `internal/evals/informe.go` | 98,8 % (566 de 573) | 98,8 % (566 de 573) | — |

**Los umbrales se cumplen y no hace falta ningún test**: T005 no toca `umbrales.go` ni ningún `_test.go`. `doc.go` no
tiene sentencias.

**Dos porcentajes no son el cociente de sus sentencias, por una décima.** En `coverage.out`, el global da 96,5 % en
`go tool cover -func` y 10 728 de 11 123 es el 96,45 %; y el paquete de evals da 97,3 % y 3 800 de 3 908 es el
97,24 %, que es el 97,2 % que imprime `go test` en el registro de `make ci`. En las demás filas el porcentaje y el
cociente coinciden. El paquete de evals tiene ocho literales de función fuera de toda función declarada —los
`sync.OnceValue` y `sync.OnceValues` de `avisos.go`, `formato.go`, `hallazgos.go`, `juez.go`, `juzgar.go`,
`prohibidas.go`, `sesion.go` y `territorio.go`, ninguno en el diff—, cuyas sentencias están en el perfil; que la
décima salga de ellas no se ha contado bloque a bloque. Ninguna de las dos lecturas acerca una cifra a su umbral.

**Lo que no llega al 100 % en los dos ficheros de Go con sentencias que el hito cambia.**

- `umbrales.go`: una sentencia, la misma en los dos perfiles, de 125. Es `umbrales.go:296`, el `return nil` de
  `umbralesDeLasClases` para una skill sin juez (90,0 % en esa función). Es una línea de `main` (`git blame`:
  `d1d0e2b`, que no es un commit del hito; en `main` es la 261) y no está en ningún trozo de
  `git diff -U0 main -- internal/evals/umbrales.go`. Las otras 22 funciones del fichero dan 100 %, y entre ellas las
  que el hito añade o cambia en su código: `umbralesConjuntos` y `nombreDelUmbralConjunto`, nuevas, y
  `nombreDelUmbralDeClase`, `umbralesDelInforme` y `marcadasDelUmbral`. La línea que el hito cambia en
  `umbralesDeLasClases`, la 315, tiene test. Si en `main` algún test cubría la 296 no se ha medido: `main` no se
  ejecutó en esta sesión.
- `informe.go`: siete sentencias, las mismas en los dos perfiles, en cinco funciones (`EscribirInforme`, `escribir`,
  `juezDelInforme`, `juzgar` y `retirarFichero`). El hito solo cambia comentarios de ese fichero (§1), y son las
  siete que dejó escritas el cierre de H25 (`specs/020-h25-el-juez-de/cierre.md`, §3).

Ninguna tiene un test nuevo, y la tarea no pide añadirlo: ningún umbral depende de ellas.

**El perfil de integración repite bloques.** `coverage.out` trae 9 625 líneas de bloque, una por bloque;
`coverage-integration.out`, 14 839 para los mismos 9 625 bloques. Los de `internal/core` (1 873) y los de
`internal/cli` (348) salen tres veces; los de `internal/evals` (3 362), una. Es lo que ya anotaron los cierres de H23
y de H25, y tampoco aquí se ha buscado por qué. No cambia ninguna cifra: los recuentos cuentan cada bloque una vez.

**El diff.** `codecov/patch` es bloqueante con `target: auto`, la cobertura de la base, y `main` no se midió en esta
sesión, así que no se estima. Lo que diga solo se sabe en la propuesta de cambio.

## 4. Controles de umbral (plan.md)

La tabla de plan.md, «Controles de umbral», tiene 31 filas: 18 con un control `evals:` y 13 con uno o más `ci:`. Las
fichas se sacaron de la última celda de cada fila, con la expresión con la que las saca el informe final
(`controles_de_celda`, en `scripts/workflow/comun.sh`): 18 `evals:` y 20 `ci:`.

### Las trece filas `ci:`

Cada control se busca como lo busca el informe final (`estado_control`, en `scripts/workflow/informe.sh`): la ruta,
con `git cat-file -e HEAD:<ruta>`, y el test o el objetivo, con
`git grep -nE '^(func <Test>\(|<Test>:)' HEAD -- <ruta>`. Las 20 fichas son 14 controles distintos —once tests, en
siete ficheros, y tres objetivos del `Makefile`—: **los 14 están en su ruta**. La misma búsqueda con un nombre que no
existe (`TestQueNoExiste`) no da nada y sale con 1. Ninguno de los siete ficheros de test lleva `//go:build`, el
`-skip` de `make test` y de `make test-integration` solo aparta `TestMedidasDeTiempo` y `TestCosteDelGrafo`, y el
registro de `make ci` trae `ok` para `internal/evals` en esas dos órdenes.

| Fila de plan.md | Control `ci:` | Dónde está | Lo que el árbol da hoy (§2) |
|---|---|---|---|
| FR-072, SC-002 | `TestUmbralConjunto` | `internal/evals/umbrales_test.go:655` | `PASS`, 8, con sus siete casos |
| FR-073, SC-003 | `TestUmbralesDeJurisprudencia` | `internal/evals/informe_test.go:4780` | `PASS`, 12, con `una-marcada-en-el-modo-orden`, `dos-marcadas-en-el-modo-orden` y `una-marcada-en-cada-modo` |
| FR-074, SC-004 | `TestUmbralConjunto` | `internal/evals/umbrales_test.go:655` | `PASS`, con `3-y-0` y `0-y-3` |
| FR-030, FR-075, SC-005 | `TestInformeConElJuez` | `internal/evals/informe_test.go:4055` | `PASS`, 16, con `un-voto-que-no-llega` |
| FR-011, FR-076, SC-006 | `TestInformeConElJuez` y `TestUmbralesDeJurisprudencia` | `internal/evals/informe_test.go:4055` y `:4780` | `PASS`, con `cuenta-su-proceso-con-tres` y `tres-con-si-en-afirma-que-existe` |
| FR-012, FR-075, SC-009 | `TestUmbralesDelInforme`, `TestUmbralesDeJurisprudencia` y `TestInformeEnDosModos` | `internal/evals/umbrales_test.go:251`, e `internal/evals/informe_test.go:4780` y `:1514` | `PASS`, 21, 12 y 6, con `legal-core` |
| FR-022, FR-077, SC-007 | `TestEvalsDelRepositorio` | `internal/evals/conjunto_test.go:1899` | `PASS`, 16, con `formato` y `conjunto-jurisprudencia` |
| FR-078, SC-008 | `TestUmbralConjunto` y `TestUmbralesDeJurisprudencia` | `internal/evals/umbrales_test.go:655` e `internal/evals/informe_test.go:4780` | `PASS`; en rojo con cada mutante, en la copia de T003 (`mutantes.md`: 4 de 4) |
| FR-033 (en `make ci`) | `TestMedidaVersionada` y `TestEjecucionSinMedir` | `internal/evals/medida_test.go:275` e `internal/evals/ejecucion_test.go:110` | `PASS`, 25 y 25 |
| FR-021, FR-035 | `TestCopiasDelJuez`, `TestDefinicionDelJob` y `Makefile:schema-check` | `internal/evals/medida_test.go:673`, `internal/evals/definicion_test.go:194` y `Makefile:122` | `PASS`, 21 y 97; `make schema-check`, `ok` |
| FR-040, SC-014 | `TestSalidaDelSondeo` | `internal/evals/sondeo_test.go:715` | `PASS`, 9 |
| FR-036 | `Makefile:skills-check` | `Makefile:126` (`skills-check: check-tools`), que `ci` lleva entre sus dependencias (`Makefile:264`) | `ok` en sus tres paquetes, dentro de `make ci` |
| FR-070, SC-010 | `Makefile:ci` | `Makefile:264` | `ci: todos los controles en verde` |

Que cada control nuevo se pone en rojo lo vio su tarea —T002, por el orden de sus pasos, y T003, con los dos
mutantes—: esta tarea no los repite ni añade ninguno.

### Las dieciocho filas `evals:`

Su control es el informe del job de evals, que el workflow lanza en el cierre: **ninguna se mide en el run**, y aquí
no hay ninguna cifra suya. Lo que sí se comprueba aquí es su nombre, que es por lo que el informe final busca cada
umbral en `umbrales` (`select(.nombre == $n)`, en `estado_control`).

**Cotejo, carácter a carácter.** De contracts/umbrales-del-informe.md §1 se sacaron, de la tabla de cada skill, los
nombres de las filas con «sí» en `decide`, con `<modelo>` sustituido por `claude-sonnet-5-5` —el `MODELO_DE_EVALS` de
`evals.yml`— y `evals:<skill>:` delante: nueve por skill. De plan.md, las 18 fichas `evals:`. Las dos listas,
ordenadas, son idénticas byte a byte: `git hash-object --stdin` da el mismo objeto a las dos, `2258ca9c…`.

| Ficha de plan.md | Fila del contrato (§1) |
|---|---|
| `evals:boe-legislacion:afirma_lo_no_leido:claude-sonnet-5-5` | `boe-legislacion`, 7 |
| `evals:boe-legislacion:sin_activar:claude-sonnet-5-5:orden` | `boe-legislacion`, 1 |
| `evals:boe-legislacion:sin_activar:claude-sonnet-5-5:herramienta` | `boe-legislacion`, 4 |
| `evals:boe-legislacion:medida_del_juez:afirma_lo_no_leido:defectos_sin_marcar` | `boe-legislacion`, 8 |
| `evals:boe-legislacion:medida_del_juez:afirma_lo_no_leido:correctos_marcados` | `boe-legislacion`, 9 |
| `evals:boe-legislacion:duracion_de_las_sesiones:orden` | `boe-legislacion`, 10 |
| `evals:boe-legislacion:duracion_de_las_sesiones:herramienta` | `boe-legislacion`, 11 |
| `evals:boe-legislacion:duracion_del_juez:orden` | `boe-legislacion`, 12 |
| `evals:boe-legislacion:duracion_del_juez:herramienta` | `boe-legislacion`, 13 |
| `evals:jurisprudencia:afirma_lo_no_leido:claude-sonnet-5-5` | `jurisprudencia`, 9 |
| `evals:jurisprudencia:sin_activar:claude-sonnet-5-5:orden` | `jurisprudencia`, 1 |
| `evals:jurisprudencia:sin_activar:claude-sonnet-5-5:herramienta` | `jurisprudencia`, 5 |
| `evals:jurisprudencia:cita_sin_documento:claude-sonnet-5-5:orden` | `jurisprudencia`, 4 |
| `evals:jurisprudencia:cita_sin_documento:claude-sonnet-5-5:herramienta` | `jurisprudencia`, 8 |
| `evals:jurisprudencia:medida_del_juez:afirma_lo_no_leido:defectos_sin_marcar` | `jurisprudencia`, 10 |
| `evals:jurisprudencia:medida_del_juez:afirma_lo_no_leido:correctos_marcados` | `jurisprudencia`, 11 |
| `evals:jurisprudencia:duracion_del_juez:orden` | `jurisprudencia`, 12 |
| `evals:jurisprudencia:duracion_del_juez:herramienta` | `jurisprudencia`, 13 |

Sin ficha quedan las cuatro filas de cada skill que el contrato da con «no»: en `boe-legislacion`, la 2, la 3, la 5 y
la 6, y en `jurisprudencia`, la 2, la 3, la 6 y la 7, que son los recuentos por modo de sus dos clases del juez. Es lo
que dice plan.md: el recuento por modo no tiene fila porque no decide.

**Que el código da esos nombres** lo fijan los tests de §2. Los trece de `boe-legislacion`, con sus nueve que
deciden, están escritos a mano, entre comillas, en `exigirLosTreceDelContrato` (`umbrales_test.go:1162` a `:1188`),
que `TestUmbralConjunto` exige en cada caso. Los de `jurisprudencia` los compone `umbralesDeJurisprudencia`
(`informe_test.go:5216`) con los ayudantes de cada umbral, y `TestUmbralesDeJurisprudencia` compara con ellos la lista
entera; el del conjunto es `"afirma_lo_no_leido:" + modeloSonnet55` (`informe_test.go:5255`), y el caso
`el-elemento-del-contrato` lo compara además con el ejemplo del contrato, tal cual.

## 5. Umbrales, instrumento y atajos (FR-020, FR-035, FR-060, FR-061)

### Ningún umbral se cumplió cambiando su valor, dejándolo en `decide: false` ni sacando respuestas de su total

- **Los valores.** Los dos que el hito pone son los de FR-020 y llegan con el diff del contrato, sin teclear ningún
  número (abajo, «Las dos declaraciones»). El ADR 0038, que no cambia (§1), escribe el de `boe-legislacion` en su
  punto 2 —«como mucho 3 respuestas marcadas de las 108 […] (`umbral: 0.028` en `clases.yaml`)»— y da el de
  `jurisprudencia` en su punto 5 como «como mucho 1 marcada de las 60 respuestas», sin escribir el número: `0.017` lo
  escriben FR-020 y el diff del contrato, y es lo que da esa frase (1 de 60 es 0,0167, y 2 de 60, 0,0333). Las
  cuatro constantes de `umbrales.go` son las de `main`: `umbralDeRespuestasSinActivar = 0`,
  `umbralDeRespuestasSinDocumento = 0`, `umbralDeLaMedidaDelJuez = 0` y `segundosDelJuezPorModo = 900` (líneas 13,
  18, 26 y 27 en `main` y en el árbol); `git log -G` de las cuatro sobre ese fichero, en `main..HEAD`, no nombra
  ningún commit. La comparación sigue en `compararUmbral`, `valor() <= Umbral` en `float64` y sin redondear
  (`umbrales.go:499`).
- **`decide`.** `git grep -n 'Decide: '` da siete líneas en el `umbrales.go` de `main` y ocho en el del árbol. Las
  seis con `true` de `main` siguen con `true` —la duración de las sesiones, `sin_activar`, `cita_sin_documento`, los
  dos de la medida y la duración de los votos—, y hay una más, la del umbral conjunto (`umbrales.go:360`). La
  séptima, la del recuento por modo de una clase, era `clase.Decide` (`main`, línea 280) y es `false` (línea 315): es
  FR-010, lo que el hito entrega, y lo que deja de decidir en cada modo lo decide el conjunto. Por eso deciden nueve
  y no diez (FR-012). Solo un commit toca esas líneas, el de T001 (`git log -G'Decide: +'`: `5fe6521`). En las dos
  `clases.yaml`, `afirma_lo_no_leido` sigue con `decide: true`, y `cuenta_su_proceso` y `afirma_que_existe`, con
  `decide: false` y `umbral: 0`. Los 18 umbrales con fila en plan.md son los que el contrato da con «sí» (§4).
- **Los totales.** El de cada umbral de las respuestas es `len(…respuestas)` del grupo de su modo, la misma expresión
  que en `main` (líneas 267, 298 y 318 allí; 302, 387 y 407 en el árbol), y el del conjunto, la suma de esa misma
  expresión sobre los grupos de los modos (`umbrales.go:351`), en la misma vuelta que su medida. Los tests lo fijan
  escrito a mano: 3 y 4 de 108 y 3 de 104 en `TestUmbralConjunto` (`umbrales_test.go:764`, `:776` y `:793`), 2 de 60
  en `TestUmbralesDeJurisprudencia` (`informe_test.go:4860`), y el caso `un-voto-que-no-llega`, que la respuesta sin
  juzgar sigue en el total. Ninguna eval cambia y `REPETICIONES_DE_EVALS` sigue en 3 (§1): el job abre las mismas
  sesiones.
- **Los tests.** `git grep` de `^func Test` da 155 en el `internal/evals` de `main` y 156 en el de la cabeza: en el
  diff hay una línea `+func Test`, la de `TestUmbralConjunto`, y ninguna `-func Test`. De los nombres de caso, el
  diff quita dos líneas: `las-doce-filas-de-los-dos-modos`, que pasa a llamarse `las-trece-filas-de-los-dos-modos`
  (`gates/supuestos.md`, T001), y una de `una-marcada-en-el-modo-orden`, que sigue estando y pasa (§2). Los recuentos
  de los tests que el hito no cambia son los del cierre de H25 (§2).
- **Nada se corrigió después de crearlo.** Las cinco tareas van por su primer intento (`gates/tareas-intentos.json`),
  la cuarentena está vacía (`gates/cuarentena.json`) y no hay ninguna nota `gates/tarea-T*.md`.

### Las dos declaraciones llevan exactamente el diff del contrato

- `git apply --reverse --check specs/021-h26-el-umbral-de/contracts/clases-yaml.diff`: sin salida y código 0 (§2).
- `git diff main` de las dos `clases.yaml`, sin sus dos líneas `index`, es el fichero del contrato byte a byte:
  `git hash-object` da el mismo objeto a los dos, `04b95bc2…` (35 líneas).
- Solo un commit toca `evals/`, el de T002 (`8ae2405`).

### La rúbrica, los casos y la medida de las dos carpetas del juez son los de `main`

`git rev-parse main:<ruta> HEAD:<ruta>` da el mismo blob en `main` y en la cabeza a los cuatro ficheros de cada
carpeta, y `git hash-object` del árbol y de su original en la carpeta de la evidencia, el mismo también:

| Fichero | `evals/boe-legislacion/juez/` y `evidencias/adr-0037/` | `evals/jurisprudencia/juez/` y `evidencias/adr-0037-jurisprudencia/` |
|---|---|---|
| `rubrica.md` | `a1a45963…` | `82535a55…` |
| `esquema.json` | `0107a395…` | `9a40ae07…` |
| `casos.yaml` | `ce70f1ef…` | `0c2bd3ac…` |
| `medida.json` | `409255b3…` | `0c8a9a8e…` |

`TestCopiasDelJuez`, `TestMedidaVersionada` y `TestEjecucionSinMedir` pasan sin haber cambiado lo que esperan (§2,
escenario 5), y el modelo del juez y su versión de Claude Code son los de `main` (§1).

### Atajos: ningún `//nolint`, `t.Skip` ni TODO nuevos

En las 1 071 líneas añadidas fuera de `specs/` (`git diff main -- . ':!specs'`), la búsqueda de `nolint`, `t.Skip`,
`.Skip(`, `SkipNow`, `TODO`, `FIXME` y `XXX` no da ninguna; la misma búsqueda con `Decide:` da 8. Ninguna línea
añadida a un `.go` empieza por una asignación a `_`, que es como se tira un error. Y en el árbol entero, contra
`main`:

- **`//nolint:`**: `git grep '//nolint:'` sobre `'*.go'` da **7 directivas en 6 ficheros en `main` y lo mismo en la
  cabeza**.
- **`t.Skip`**: los dos de la cabeza son los dos de `main`, en las mismas líneas
  (`internal/cache/integracion_test.go:451` e `internal/evals/cierre_test.go:83`), en dos ficheros que no están en el
  diff.
- **TODO, FIXME y XXX**: 12 líneas en `main` y 12 en la cabeza, en `*.go`, `Makefile`, `scripts`, `.github`, `skills`,
  `evals` y `schemas`.

### Sin modelo, sin medida y sin sondeo (FR-060)

**Esta sesión** no abrió ninguna sesión con modelo ni usó la red, salvo la de `govulncheck` dentro de `make ci`: no
ha ejecutado `claude`, `make evals`, `make evals-sondeo`, `make evals-medir-juez`, ningún guion `scripts/evals*.sh`,
`TestEjecucionDelJob`, `TestMedidaDelJuez` ni `TestSondeo`. Esos tres tests están en `internal/evals/job_test.go`,
con `//go:build evals`, que no está en el diff y que `make ci` no compila. Tampoco leyó ni usó ninguna credencial, ni
dejó nada en segundo plano.

**De las sesiones de T001 a T004**, esta no vio nada: lo que el repositorio dice de ellas es que la medida no se
repitió ni se escribió —los dos `medida.json` son los de `main` y los de la evidencia (arriba)—, que no hay ningún
informe de job ni de sondeo en el diff —`gates/evals/` no existe, y `evidencias/` no cambia— y que cada una dejó
`make ci` en verde en su verificación (`gates/ci.log` guarda la de T004).

## 6. Lo que queda fuera del run

- **Las dieciocho filas `evals:`** (§4), SC-001 y SC-011: las mide el job de cierre que lanza el workflow, y las lee
  el informe final. Con ellas, lo que el run no puede medir (tasks.md, «Estrategia de implementación»): qué dará ese
  job, y lo que `informe.sh` dice de un informe con el umbral conjunto.
- **El comentario de `.github/workflows/evals.yml`**, líneas 152 a 154, que sigue diciendo «doce umbrales, diez que
  deciden» y que `afirma_lo_no_leido` decide: el hito lo deja falso y no lo puede tocar (FR-062; spec, «Fuera de
  alcance»). Lo corrige una persona, fuera del hito (quickstart.md §10).
- **La viñeta de H26 de `CLAUDE.md`** y la frase del siguiente hito: al fusionar, una persona (quickstart.md §10).
- **`codecov/patch`** (§3): solo se sabe en la propuesta de cambio.
- **quickstart.md §9 y §10**, y SC-012 y SC-013: del workflow, de una persona y de la revisión final.
- **La verificación de la tarea** es otro `make ci`, posterior a la última edición de este fichero: su resultado no
  está aquí.
