# H23.1 · Cierre de la Definition of Done

Cierre de T005 (FR-032, FR-033, FR-034, FR-045, FR-050, FR-051, FR-052, FR-060, FR-061; SC-009) sobre `ff27b7d`,
que es T004, el último commit antes de T005. La rama está al día con `main`: `git merge-base main HEAD` da la cabeza
de `main`, `ef35b40`, así que `git diff --name-status main` da exactamente lo que cambia el hito. Las cinco partes son
las de la tarea: lo creado, lo modificado y lo que no cambia (§1), el quickstart (§2), la cobertura (§3), los controles
de umbral (§4) y los umbrales, el instrumento y los atajos (§5).

Todo lo que se cita aquí se ejecutó en la sesión de T005, en primer plano, y se vio terminar, o está en el
repositorio. Lo que no se pudo medir se dice tal cual (§3, «El diff»; §4, las once filas `evals:`; §5, «Sin modelo»;
§6).

**T005 no toca código ni añade ningún test.** Las cifras de cobertura quedan sobre sus umbrales (§3), y la tarea solo
pide tests si alguna queda debajo: `internal/evals/informe_test.go` e `internal/evals/umbrales_test.go`, que la tarea
declara para ese caso, no cambian en esta tarea. Sus únicos cambios son este fichero, una línea en `gates/supuestos.md`
y su marca en tasks.md.

**Un hallazgo para la persona y la revisión final** (§1, «Lo que no cambia», último punto): el diff del hito añade
`specs/021-h26-el-umbral-de/gates/modelos.jsonl`, un fichero del directorio de H26. No lo escribió ninguna tarea: lo
trae el commit `92b434a` («docs(H23.1): spec, clarificaciones y plan del hito») y es una línea del paso `specify` del
workflow. T005 no lo retira (`gates/supuestos.md`).

## 1. Lo creado y lo modificado en el hito, y lo que no cambia

`git diff --name-status main` da **43 ficheros: 33 `A` y 10 `M`**, ninguna `D` ni `R`. De los 33 `A`, 32 están en
`specs/022-h23-1-jurisprudencia-coteja-el/` y uno en `specs/021-h26-el-umbral-de/` (abajo). Este `cierre.md` no está
entre ellos, porque no tiene seguimiento hasta el commit de la tarea. Los 10 `M` son todo lo que el hito cambia fuera de
`specs/` (`git diff --shortstat main -- . ':!specs'`: 10 ficheros, 663 líneas añadidas y 138 quitadas). El hito no
crea ningún fichero fuera de `specs/`. `git status --short` da, antes de §2, los tres ficheros de `gates/` que el
workflow lleva modificados (`modelos.jsonl`, `tarea-actual.json` y `tareas-intentos.json`).

### Los 10 ficheros fuera de `specs/`

Cada uno con la tarea que lo tocó (`git log --name-only main..HEAD -- . ':!specs'`) y sus líneas
(`git diff --numstat main`, añadidas y quitadas):

| Fichero | Tarea | Líneas |
|---|---|---|
| `internal/evals/informe.go` | T001 | +130 −11 |
| `internal/evals/umbrales.go` | T001 | +138 −38 |
| `internal/evals/doc.go` | T001 | +20 −14 |
| `internal/evals/informe_test.go` | T001 | +273 −67 |
| `internal/evals/conjunto_test.go` | T001 | +40 −0 |
| `skills/jurisprudencia/SKILL.md` | T003 | +11 −1 |
| `CHANGELOG.md` | T004 | +28 −0 |
| `CONTRIBUTING.md` | T004 | +19 −4 |
| `docs/JURISPRUDENCIA.md` | T004 | +3 −2 |
| `CLAUDE.md` | T004 | +1 −1 |

Son los cinco ficheros Go del paquete de evals, `SKILL.md` de `jurisprudencia` y los cuatro documentos. T002 solo
añadió `mutantes.md`, dentro del directorio del feature.

**Lo que cambia en el código, sin los comentarios.** `git diff -U0 main` de `informe.go`, `umbrales.go` y `doc.go`,
quitando las líneas de comentario y las vacías, solo da lo del umbral nuevo: los campos `conCotejo`, `pideElCotejo` y
`cotejo`; `queLePidenElCotejo`, `sinCotejar`, `pideElCotejo`, `comandoDelCotejo`, `cotejaUnDocumento` y
`medirLosCotejos`; la constante `umbralDeRespuestasSinCotejar = 0`, el prefijo `documento_sin_cotejar:` y la
descripción; el `append` de `umbrales()` detrás de `cita_sin_documento`; `umbralDeDocumentoSinCotejar`,
`nombreDelUmbralDeDocumentoSinCotejar` y `sinCotejarDelUmbral`; y en `nombradasEn`, el `return` que ahora prueba
`sinDocumentoDelUmbral` y después `sinCotejarDelUmbral`. Las únicas líneas de código quitadas son las seis del bloque
de prefijos, que vuelven con los mismos valores y otra alineación (`gofmt`), y ese `return`. `doc.go` solo cambia
comentarios.

### Lo que no cambia (FR-032, FR-033, FR-034, FR-045, FR-052)

- **Las órdenes de quickstart.md §3** (§2, escenario 3): la segunda,
  `git diff --name-only main -- internal/evals/sondeo.go scripts skills/boe-legislacion skills/legal-core evals schemas evidencias testdata internal/app/testdata .github .specify cmd docs/ROADMAP.md docs/ADR docs/USO.md docs/HISTORIAL.md docs/WORKFLOW.md README.md web`,
  no da ninguna línea (contada con `wc -l`: 0). Cubre `sondeo.go`, los guiones de `scripts/` (también
  `scripts/workflow/` y `scripts/evals-sondeo.sh`), `boe-legislacion` y `legal-core`, las evals, el directorio de
  esquemas, la carpeta de la evidencia, los fixtures de `testdata/`, `.github/`, `.specify/` (con la constitución),
  `cmd/`, el roadmap, los ADR, la bitácora de uso, `HISTORIAL.md`, `WORKFLOW.md`, el léeme y la web. La tercera,
  `git diff --name-only main -- internal skills`, da solo los cinco ficheros de `internal/evals/` y
  `skills/jurisprudencia/SKILL.md`: el binario no cambia fuera del paquete de evals.
- **Los mismos árboles, por su objeto**: `git rev-parse main:<ruta> HEAD:<ruta>` da el mismo a `skills/boe-legislacion`
  (`eed129f8…`), `skills/legal-core` (`01a63a0a…`), `.github` (`951ca3fb…`), `.specify` (`28032037…`), `schemas`
  (`d70359a6…`), `internal/evals/sondeo.go` (`a5f96eaa…`), `evals` (`10bc1ca0…`) y `evidencias` (`ee52f274…`).
- **`docs/`**: `git diff --name-only main -- docs` da solo `docs/JURISPRUDENCIA.md`, que es de T004.
- **Los artefactos de otros hitos.**
  `git diff --name-only main -- specs ':(exclude)specs/022-h23-1-jurisprudencia-coteja-el'` da **un fichero**:
  `specs/021-h26-el-umbral-de/gates/modelos.jsonl`, nuevo (`A`), de una línea. Lo trae `92b434a`, el commit del spec,
  las clarificaciones y el plan; ninguna tarea lo ha tocado (`git log main..HEAD -- specs/021-h26-el-umbral-de/` solo
  da ese commit). Su línea es la del paso `specify` de este run —`"paso":"specify"`, `"origen":"workflow"`,
  `"inicio":"2026-10-16T06:54:25Z"`—, escrita en el directorio del último hito porque el del nuevo aún no existía. No
  es el spec, el plan ni un informe de H26, que son lo que FR-045 protege, y FR-052 no nombra `specs/`. Pero es un
  fichero de otro hito en el diff, y esta tarea pide constatar que no hay ninguno: no se puede. T005 no lo retira,
  porque su ruta no está entre las declaradas y el guardián de diff solo admite el directorio del feature; la causa
  está en el paso `specify` del workflow, que no se toca dentro de un hito. Lo decide la persona: quitarlo de la
  propuesta de cambio, o dejarlo y corregir el workflow fuera del hito.

### La Definition of Done, punto a punto (`ROADMAP.md` §1)

| Punto | Estado |
|---|---|
| 1 · `make ci` en verde | Sí (§2, escenario 7); y la verificación de la tarea, otro `make ci` posterior a este fichero |
| 2 · Tests offline; fixtures si toca red | Sí: sesiones sintéticas y votos grabados; ningún fixture nuevo (`testdata` fuera del diff) |
| 3 · Sin `net/http`, `os.Exit`, `fmt.Print*` fuera de lo autorizado | Sí: el diff de código (arriba) no añade ningún import ni ninguna de esas llamadas; `lint` en verde (`depguard`, `forbidigo`) |
| 4 · Esquemas y `schema-check` | Ningún esquema cambia; `make schema-check`, `ok` (§2) |
| 5 · Errores tipados y códigos estables | Ningún código de `kitlegal` cambia: el binario no cambia fuera del paquete de evals |
| 6 · e2e y `CHANGELOG.md` | «Aceptación e2e: no aplica» (plan.md); `CHANGELOG.md` en *Unreleased* (§2, escenario 6) |
| 7 · ADR | No aplica (plan.md, «Summary»): `docs/ADR` fuera del diff |
| 8 · `SOURCES.md` | No aplica: ninguna fuente nueva ni cambiada |
| 9 · Cobertura | Sí (§3) |
| 10 · Skill: evals antes, `SKILL.md` < 300, sin drift | Sí: 207 líneas, `make skills-check` `ok`; las evals que piden el cotejo son las de `main` (§5) |
| 11 a 13 | No aplican |

## 2. Quickstart (§0 a §3, §5 a §7)

Cada orden se ejecutó tal como la da quickstart.md, salvo donde se dice, y su salida se leyó con `rtk proxy`. Las del
escenario 3 y 6 que el sandbox de la sesión no dejó encadenar con `echo` se ejecutaron sueltas o con `wc -l` detrás.

- **§0.** `go version`: `go1.27.2 darwin/arm64` (Go 1.27). `git status --short`: los tres ficheros de `gates/` que
  lleva modificados el workflow (`modelos.jsonl`, `tarea-actual.json`, `tareas-intentos.json`).
- **§1.** `go test -count=1 -v -run '^TestUmbralesDeJurisprudencia$' ./internal/evals/`: `ok` y **quince subtests en
  `PASS`**: `el-elemento-del-contrato`, los diez de antes (`ninguna-cuenta`, `una-sin-activar`,
  `una-sin-documento-en-el-modo-orden`, `dos-sin-documento-en-el-modo-herramienta`, `tres-con-si-en-afirma-que-existe`,
  `una-marcada-en-el-modo-orden`, `dos-marcadas-en-el-modo-orden`, `una-marcada-en-cada-modo`,
  `una-marcada-en-las-dos-clases`, `901-s-del-juez-en-el-modo-orden`) y los cuatro nuevos
  (`una-sin-cotejar-en-el-modo-orden`, `una-sin-cotejar-en-el-modo-herramienta`, `una-sin-activar-y-sin-cotejar`,
  `una-sin-medir-y-sin-cotejar`). Lo esperado.
- **§2.** (1) `go test -count=1 -v -run '^TestEvalsDelRepositorio$' ./internal/evals/`: `ok`, 16 subtests en `PASS`,
  entre ellos `piden-el-cotejo`, que exige la 04, la 06, la 08 y la 09 de `jurisprudencia` y ninguna de las otras dos
  (`conjunto_test.go:1988` a `:2007`). (2) `make skills-check`: `ok` en `internal/app`, `internal/skills` e
  `internal/evals`. (3) `make schema-check`: `ok` en `internal/app`. Lo esperado.
- **§3.** (1) La orden con los diecinueve tests: `ok`, los diecinueve en `PASS` —`TestUmbralesDelInforme`,
  `TestUmbralesConJuezYConSentencias`, `TestUmbralConjunto`, `TestInforme`, `TestInformeConElJuez`,
  `TestInformeMarkdownDeLosUmbrales`, `TestInformeEnDosModos`, `TestInformeConSesionesSinMedir`,
  `TestInformeFinalDeLosUmbrales`, `TestJuzgar`, `TestJuzgarLasLlamadas`, `TestJuzgarSentencias`,
  `TestCitaSinDocumento`, `TestCopiasDelJuez`, `TestMedidaVersionada`, `TestDefinicionDelJob`, `TestJuicioDelSondeo`,
  `TestSalidaDelSondeo` y `TestSondear`—. `TestInformeFinalDeLosUmbrales` sigue esperando «jurisprudencia: 13
  umbrales, los 9 que deciden se cumplen…» (`informe_test.go:1854`), con `umbralesDeJurisprudenciaComoEnH26`. (2) Sin
  salida. (3) Los seis ficheros esperados. Lo esperado (§1, «Lo que no cambia»).
- **§4** lo ejecutó T002, en una copia fuera del árbol: `mutantes.md`, «Resultado: 2 de 2». Esta tarea no lo repite.
- **§5.** `wc -l`: **207** (< 300). Las tres órdenes `cmp`, sin salida, y el directorio temporal borrado: no cambian
  las líneas 1 a 38, el paso 2 (33 líneas) ni del paso 4 al final (84 líneas). `git diff main` de `SKILL.md`: dos
  trozos, `@@ -48 +48,2 @@`, en el paso 1 (líneas 40 a 50), y `@@ -84,0 +86,9 @@`, en el paso 3 (desde la 84): el
  cotejo es lo primero, nada del documento antes del cotejo, y si es la pedida lo dice `es_la_pedida`, con el caso de
  1088/2023 y el ROJ STS 3144/2023. Lo esperado; la lectura contra contracts/skill-jurisprudencia.md es de la revisión
  final (SC-007).
- **§6.** (1) Sin salida (`wc -l`: 0). (2) `CONTRIBUTING.md:3`, `docs/JURISPRUDENCIA.md:1`, `CLAUDE.md:1`: mayor que 0
  en los tres. (3) `4`: mayor que 1. (4) Sin salida (`wc -l`: 0). (5) `@@ -44,0 +45,15 @@` y `@@ -63,0 +79,13 @@`:
  44 y 63, por debajo de 65, la línea de `## [0.6.0]` en `main` (en el árbol, la 93). Lo esperado.
- **§7.** `make ci`, con la caché de tests vaciada antes (`go clean -testcache`) y el registro en `$TMPDIR`: **sale
  con 0**, de 11:24:47 a 11:29:14, y termina con «ci: todos los controles en verde». El registro no trae ningún
  `(cached)` y lleva, en su orden: `golangci-lint fmt --diff`, `golangci-lint run` («0 issues.»), `go test -race` con
  `coverage.out`, el de `-tags=integration` con `coverage-integration.out`, `test-tiempos`, `govulncheck` («No
  vulnerabilities found.»), `schema-check`, `skills-check`, `goreleaser check`, `gitleaks` («no leaks found»),
  `go mod verify` y `go mod tidy -diff`.
- **§8** lo hace el workflow tras la revisión final: empuja, abre la propuesta de cambio, pone la etiqueta `evals` y
  guarda los informes en `gates/evals/`, que hoy no existe.
- **§9** queda para después del run, para una persona: el sondeo de la eval 06 (SC-010), el comentario de
  `.github/workflows/evals.yml:157` —que sigue diciendo «trece umbrales, nueve que deciden»— y la viñeta de H23.1 de
  `CLAUDE.md`.

## 3. Cobertura (Definition of Done §1.9)

Sobre los dos perfiles que deja el `make ci` de §2, escenario 7: `coverage.out` (`make test`) y
`coverage-integration.out` (`make test-integration`). El porcentaje sale de `go tool cover -func`: el global, sobre
cada perfil; el de cada árbol o fichero, sobre el perfil filtrado a sus líneas con `grep` y su cabecera `mode:`,
escrito bajo `$TMPDIR`. Los recuentos de sentencias salen de los bloques del perfil —su segundo campo—, contando cada
bloque una vez, con un `awk` en línea sobre los dos perfiles.

| Árbol | `coverage.out` | `coverage-integration.out` | Umbral |
|---|---|---|---|
| Global | **96,5 %** (10 776 de 11 168 sentencias) | **96,9 %** (10 822 de 11 168) | ≥ 70 % (Definition of Done §1.9; `codecov.yml`) |
| `internal/core/**` (dominio) | **98,3 %** (2 100 de 2 136) | **98,3 %** (2 100 de 2 136) | ≥ 85 % (Definition of Done §1.9; `codecov.yml`) |
| `internal/cli/**` | 98,0 % (398 de 406) | 98,0 % (398 de 406) | ≥ 90 % (`codecov.yml`) |
| `internal/evals` (el paquete de evals) | **97,3 %** (3 845 de 3 953) | **97,7 %** (3 862 de 3 953) | — |
| `internal/evals/umbrales.go` | 99,3 % (140 de 141) | 99,3 % (140 de 141) | — |
| `internal/evals/informe.go` | 98,8 % (595 de 602) | 98,8 % (595 de 602) | — |

**Los umbrales se cumplen y no hace falta ningún test.** `doc.go` no tiene sentencias.

**Las funciones nuevas del paquete de evals**, en `go tool cover -func` sobre los dos perfiles: las nueve dan
**100 %** —`queLePidenElCotejo` (`informe.go:686`), `sinCotejar` (`:703`), `pideElCotejo` (`:1407`),
`cotejaUnDocumento` (`:1429`), `medirLosCotejos` (`:1446`), `umbralDeDocumentoSinCotejar` (`umbrales.go:472`),
`nombreDelUmbralDeDocumentoSinCotejar` (`:488`) y `sinCotejarDelUmbral` (`:650`); `comandoDelCotejo` es una variable—.
Las que el hito cambia en su código dan también 100 %: `medirLasRespuestas` (`informe.go:1375`), `umbrales`
(`umbrales.go:299`) y `nombradasEn` (`:632`).

**Lo que no llega al 100 % en los dos ficheros de Go con sentencias que el hito cambia**: ocho bloques sin cubrir, de
una sentencia cada uno, los mismos en los dos perfiles. En `informe.go`, siete, en `EscribirInforme` (líneas 913 y
917), `escribir` (969), `juezDelInforme` (1344 y 1349), `juzgar` (1506) y `retirarFichero` (1996): son las cinco
funciones que el cierre de H26 nombra con sus siete sentencias (`specs/021-h26-el-umbral-de/cierre.md`, §3). En
`umbrales.go`, uno, en `umbralesDeLasClases` (línea 333), el `return nil` de una skill sin juez (`if j == nil`), que
el mismo cierre nombra. Ninguna de esas líneas está en un trozo de `git diff -U0 main` de su fichero (trozos nuevos de `informe.go`:
470–475, 553–559, 683–714, 809, 811–817, 819–820, 1369–1374, 1387, 1394–1397, 1403–1464 y 1722–1723; de `umbrales.go`, el más
cercano a la 333 es el de 309–312). La tarea no pide un test para ellas: ningún umbral depende de ellas.

**El diff.** `codecov/patch` es bloqueante con `target: auto`, la cobertura de la base, y `main` no se midió en esta
sesión, así que no se estima. Lo que diga solo se sabe en la propuesta de cambio.

## 4. Controles de umbral (plan.md)

La tabla de plan.md, «Controles de umbral», tiene 20 filas: 11 con un control `evals:` y 9 con uno o más `ci:`. Las
fichas se sacaron de la última celda de cada fila con la expresión con la que las saca el informe final
(`controles_de_celda`, en `scripts/workflow/comun.sh:165`): **11 `evals:` y 16 `ci:`**.

### Las nueve filas `ci:`

Cada control se busca como lo busca el informe final (`estado_control`, en `scripts/workflow/informe.sh:346`): la
ruta, con `git cat-file -e HEAD:<ruta>`, y el test o el objetivo, con
`git grep -nE '^(func <Test>\(|<Test>:)' HEAD -- <ruta>`. Las 16 fichas son 14 controles distintos —once tests, en
siete ficheros, y tres objetivos del `Makefile`—: **las ocho rutas existen y los 14 están en su ruta**. La misma
búsqueda con un nombre que no existe (`TestQueNoExiste`) no da nada y sale con un código distinto de 0.

| Fila de plan.md | Control `ci:` | Dónde está | Lo que el árbol da hoy (§2) |
|---|---|---|---|
| FR-005, FR-062, SC-002 a SC-004 | `TestUmbralesDeJurisprudencia` | `internal/evals/informe_test.go:5124` | `PASS`, 15 subtests |
| FR-006, FR-063, SC-005 (`jurisprudencia`) | `TestUmbralesDeJurisprudencia` | `internal/evals/informe_test.go:5124` | `PASS`, la lista de quince en cada caso |
| FR-007, FR-063, SC-005 (las otras dos) | `TestUmbralesDelInforme`, `TestInformeEnDosModos` y `TestEvalsDelRepositorio` | `internal/evals/umbrales_test.go:251`, `internal/evals/informe_test.go:1516` e `internal/evals/conjunto_test.go:1899` | `PASS` los tres; `piden-el-cotejo` en `PASS` |
| FR-064, SC-006 | `TestUmbralesDeJurisprudencia` | `internal/evals/informe_test.go:5124` | `PASS`; en rojo con cada mutante, en la copia de T002 (`mutantes.md`: 2 de 2) |
| FR-023, SC-007 | `Makefile:skills-check` | `Makefile:126` (`skills-check: check-tools`) | `ok`, suelto y dentro de `make ci` |
| FR-030, FR-031 | `TestUmbralesConJuezYConSentencias`, `TestInformeConElJuez`, `TestJuzgar` y `TestInforme` | `internal/evals/umbrales_test.go:573`, `internal/evals/informe_test.go:4367`, `internal/evals/juzgar_test.go:293` e `internal/evals/informe_test.go:267` | `PASS` los cuatro |
| FR-032 | `TestCopiasDelJuez`, `TestDefinicionDelJob` y `Makefile:schema-check` | `internal/evals/medida_test.go:673`, `internal/evals/definicion_test.go:194` y `Makefile:122` | `PASS`, `PASS` y `ok` |
| FR-033 | `TestSalidaDelSondeo` | `internal/evals/sondeo_test.go:715` | `PASS` |
| FR-060 | `Makefile:ci` | `Makefile:264` | «ci: todos los controles en verde» |

Que cada control nuevo se pone en rojo lo vieron T001, por el orden de sus pasos (`gates/tarea-T001.md`), y T002, con
los dos mutantes: esta tarea no los repite ni añade ninguno.

### Las once filas `evals:`

Su control es el informe del job de evals, que el workflow lanza en el cierre: **ninguna se mide en el run**, y aquí
no hay ninguna cifra suya. Lo que sí se comprueba aquí es su nombre, que es por lo que el informe final busca cada
umbral en `umbrales` (`select(.nombre == $n)`, en `estado_control`).

**Cotejo, carácter a carácter.** De contracts/umbral-documento-sin-cotejar.md §1 se sacaron, de las quince filas, los
nombres de las once con «sí» en `decide` —el contrato ya los escribe con `claude-sonnet-5-5`, que es el
`MODELO_DE_EVALS` de `.github/workflows/evals.yml:220`—, con `evals:jurisprudencia:` delante. De plan.md, las 11 fichas
`evals:`. Las dos listas, ordenadas con `LC_ALL=C sort`, son idénticas byte a byte: `git hash-object --stdin` da el
mismo objeto a las dos, `574ac13b…`.

| Ficha de plan.md | Fila del contrato (§1) |
|---|---|
| `evals:jurisprudencia:documento_sin_cotejar:claude-sonnet-5-5:orden` | 5 |
| `evals:jurisprudencia:documento_sin_cotejar:claude-sonnet-5-5:herramienta` | 10 |
| `evals:jurisprudencia:sin_activar:claude-sonnet-5-5:orden` | 1 |
| `evals:jurisprudencia:sin_activar:claude-sonnet-5-5:herramienta` | 6 |
| `evals:jurisprudencia:cita_sin_documento:claude-sonnet-5-5:orden` | 4 |
| `evals:jurisprudencia:cita_sin_documento:claude-sonnet-5-5:herramienta` | 9 |
| `evals:jurisprudencia:afirma_lo_no_leido:claude-sonnet-5-5` | 11 |
| `evals:jurisprudencia:medida_del_juez:afirma_lo_no_leido:defectos_sin_marcar` | 12 |
| `evals:jurisprudencia:medida_del_juez:afirma_lo_no_leido:correctos_marcados` | 13 |
| `evals:jurisprudencia:duracion_del_juez:orden` | 14 |
| `evals:jurisprudencia:duracion_del_juez:herramienta` | 15 |

Sin ficha quedan las cuatro filas que el contrato da con «no» —la 2, la 3, la 7 y la 8, los recuentos por modo de las
dos clases del juez—, que es lo que dicen FR-061 y plan.md: un umbral que no decide no tiene fila.

**Que el código da esos nombres.** El de `documento_sin_cotejar` lo compone `nombreDelUmbralDeDocumentoSinCotejar`
(`umbrales.go:488`): `documento_sin_cotejar:` (`umbrales.go:67`), el modelo, `:` y el modo. El test no lo saca del
código: `umbralDocumentoSinCotejar` (`informe_test.go:5787`) lo escribe a mano, `"documento_sin_cotejar:" + modelo +
":" + string(modo)`, y el caso `el-elemento-del-contrato` compara el elemento del modo orden con el ejemplo de
contracts §2 tal cual, con su nombre `documento_sin_cotejar:claude-sonnet-5-5:orden` (`informe_test.go:5184` a
`:5192`).

## 5. Umbrales, instrumento y atajos (FR-050, FR-051, FR-060)

### Ningún umbral se cumplió cambiando su valor, dejándolo en `decide: false` ni sacando sesiones de su total

- **Los valores.** Las cuatro constantes de `main` siguen igual en el árbol: `umbralDeRespuestasSinActivar = 0` y
  `umbralDeRespuestasSinDocumento = 0` (líneas 13 y 18 en los dos), y `umbralDeLaMedidaDelJuez = 0` y
  `segundosDelJuezPorModo = 900` (26 y 27 en `main`; 32 y 33 en el árbol). `git log -G` de cada una sobre
  `umbrales.go`, en `main..HEAD`, no da ningún commit. La nueva, `umbralDeRespuestasSinCotejar = 0` (línea 24), solo
  la toca T001 (`3387608`), y es el 0 de FR-004. Los umbrales en forma de tasa viven en las `clases.yaml`, que no
  cambian (abajo).
- **`decide`.** `grep -n 'Decide: '` da ocho líneas en el `umbrales.go` de `main` —siete `true` y la `false` del
  recuento por modo de una clase (línea 315)— y nueve en el del árbol: las mismas ocho, en el mismo orden, y una más,
  `Decide: true,` en `umbralDeDocumentoSinCotejar` (línea 481). `git log -G'Decide: '` en `main..HEAD` solo da T001.
- **Los totales.** Ninguna expresión de un total de antes cambia: las únicas líneas de código quitadas de
  `umbrales.go` son las seis de los prefijos, realineadas, y el `return` de `nombradasEn` (§1). El total nuevo es
  `len(delModo.queLePidenElCotejo())` (`umbrales.go:473`): todas las respuestas medidas del modo de las evals que piden
  el cotejo, sin sacar ninguna. Los tests lo fijan escrito a mano: 12 por modo, y 11 con la sesión sin medir, que lleva
  su motivo (§2, escenario 1).
- **Los tests.** `git grep -c '^func Test'` da 165 en el `internal/evals` de `main` y 165 en el de la cabeza; el diff
  no tiene ninguna línea `+func Test` ni `-func Test`. De los nombres de caso, el diff solo quita la línea de
  `ninguna-cuenta`, que sigue estando, en otra forma, y pasa (§2); añade los cuatro casos nuevos.
- **Nada se corrigió después de crearlo.** Las cinco tareas van por su primer intento (`gates/tareas-intentos.json`),
  la cuarentena está vacía (`gates/cuarentena.json`) y la única nota es `gates/tarea-T001.md`, que deja el rojo de
  antes del código y la verificación en verde.

### Las evals, la rúbrica, los casos y la medida de la carpeta del juez son los de `main`

`git rev-parse main:<ruta> HEAD:<ruta>` da el mismo objeto en `main` y en la cabeza a `evals/` entero (`10bc1ca0…`) y a
los cinco ficheros de `evals/jurisprudencia/juez/`; y `git hash-object` del árbol de trabajo y de su copia en
`evidencias/adr-0037-jurisprudencia/`, el mismo también, a los cuatro que tienen copia:

| Fichero | `main`, cabeza, árbol y evidencia |
|---|---|
| `rubrica.md` | `82535a55…` |
| `esquema.json` | `9a40ae07…` |
| `casos.yaml` | `0c2bd3ac…` |
| `medida.json` | `0c8a9a8e…` |
| `clases.yaml` (sin copia en la evidencia) | `655c58dc…` |

`REPETICIONES_DE_EVALS` sigue en 3 y `MODELO_DE_EVALS` en `claude-sonnet-5-5` (`.github/workflows/evals.yml:228` y
`:220`; `.github/` no cambia): el job abre las mismas sesiones. `TestCopiasDelJuez`, `TestMedidaVersionada` y
`TestDefinicionDelJob` pasan sin haber cambiado lo que esperan (§2, escenario 3).

### Atajos: ningún `//nolint`, `t.Skip` ni TODO nuevos

En las 663 líneas añadidas fuera de `specs/` (`git diff main -- . ':!specs'`), la búsqueda de `nolint`, `t.Skip`,
`.Skip(`, `SkipNow`, `TODO`, `FIXME` y `XXX` no da ninguna; la misma búsqueda con `Decide:` da 2, la de
`umbrales.go:481` y la del ayudante del test. Ninguna línea añadida empieza por una asignación a `_`. El `_` de
`invocaciones, _ := deUsarYTirar.invocacionesDe(sesion)` (`cotejaUnDocumento`) no es un error: `invocacionesDe`
devuelve un booleano que una sesión leída nunca pone a falso (research D3). En el árbol entero, contra `main`:

- **`//nolint:`**: `git grep -c '//nolint:'` sobre `'*.go'` da **13 directivas en 9 ficheros en `main` y lo mismo en
  la cabeza**.
- **`t.Skip`**: los nueve de la cabeza (`t.Skip` y `t.Skipf`) son los nueve de `main`, en los mismos ficheros y
  líneas (`internal/cache/integracion_test.go`, `internal/evals/cierre_test.go`, `internal/evals/serie_test.go` e
  `internal/workflow/arbol_test.go`), ninguno en el diff.
- **TODO, FIXME y XXX**: 12 líneas en `main` y 12 en la cabeza, en `*.go`, `Makefile`, `scripts`, `.github`,
  `skills`, `evals` y `schemas`.

### Sin modelo, sin medida y sin sondeo (FR-051, FR-060)

**Esta sesión** no abrió ninguna sesión de evals ni el juez, ni usó la red salvo la de `govulncheck` dentro de
`make ci`: no ha ejecutado `claude`, `make evals`, `make evals-sondeo`, `make evals-medir-juez`, ningún guion
`scripts/evals*.sh`, `TestEjecucionDelJob`, `TestMedidaDelJuez` ni `TestSondeo`. Tampoco leyó ni usó ninguna
credencial, ni dejó nada en segundo plano.

**De las sesiones de T001 a T004**, esta no vio nada: lo que el repositorio dice de ellas es que la medida no se
repitió ni se escribió —`medida.json` es el de `main` y el de la evidencia (arriba)—, que no hay ningún informe de job
ni de sondeo en el diff —`gates/evals/` no existe, y `evidencias/` no cambia— y que cada una dejó `make ci` en verde
en su verificación (`gates/ci.log` guarda la de T004).

## 6. Lo que queda fuera del run

- **Las once filas `evals:`** (§4), SC-001 y SC-009: las mide el job de cierre que lanza el workflow, y las lee el
  informe final. Qué dará ese job con la skill v0.1.1 no se puede medir en el run (tasks.md, «Estrategia de
  implementación»).
- **`specs/021-h26-el-umbral-de/gates/modelos.jsonl`** (§1): lo decide la persona.
- **El comentario de `.github/workflows/evals.yml:157`**, que sigue diciendo «trece umbrales, nueve que deciden»: lo
  corrige una persona, fuera del hito (quickstart.md §9).
- **La viñeta de H23.1 de `CLAUDE.md`** y la frase del siguiente hito: al fusionar, una persona (quickstart.md §9).
- **`codecov/patch`** (§3): solo se sabe en la propuesta de cambio.
- **quickstart.md §8 y §9**, y SC-010: del workflow y de una persona.
- **La verificación de la tarea** es otro `make ci`, posterior a la última edición de este fichero: su resultado no
  está aquí.
