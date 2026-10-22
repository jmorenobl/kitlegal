# Cierre de la Definition of Done · H30 (T018)

Sesión del 2026-10-22 sobre la cabeza `ee0a18d` (T017) de `026-h30-las-evals-miden`, con `main` en `43d46f3`, que es
también la base de la rama (`git merge-base HEAD main`): `git diff --name-status main` es lo que traen los 20 commits de
la rama (`3b610cf`, `9a9be67`, `9337061` y T001 a T017), más los tres ficheros de `gates/` que el workflow tenía
modificados al empezar la tarea (`modelos.jsonl`, `tarea-actual.json`, `tareas-intentos.json`). Todo lo que se afirma
aquí se ejecutó y se vio terminar en esta sesión, o está en el repositorio; lo que no se pudo medir se dice tal cual.
Ninguna tarea está en cuarentena (`gates/cuarentena.json`, `"tareas": {}`), T001 a T017 están en `[X]` y la checklist del
spec no tiene ninguna casilla sin marcar.

Cómo se ejecutó (desviaciones de entorno respecto al quickstart, ninguna de producto):

- Las salidas de `go test` y `make` se leyeron con `rtk proxy`, y las de `go test -v`, filtradas con `grep` a las líneas
  `--- PASS`/`--- FAIL`, `PASS`, `FAIL` y `ok` (y, en §1 y §7, al registro que el escenario pide mirar).
- El sandbox de la sesión no deja comprobar antes de ejecutarla una orden con una variable (`$TMPDIR`, `$?`) ni `awk -f`:
  los registros de `make ci` fueron a `/tmp/claude-501/h30-t018/` con la ruta escrita entera, el código de salida se leyó
  con `&& echo "codigo: 0" || echo "codigo: distinto de 0"`, y las sumas por área de §4 las hizo un programa de un solo
  uso en `/tmp/claude-501/h30-t018/areas/main.go`, fuera del repositorio, con `go run` desde la raíz.
- `make ci` se ejecutó primero por objetivos, en tres órdenes consecutivas en primer plano, el mismo control
  (`Makefile:276`), porque una orden de la sesión no puede pasar de 10 minutos; la ejecución final, entera
  («Verificación de la tarea»).

## 1. Lo creado o modificado, y lo que no cambia

`git diff --name-status main`: 91 entradas, 61 `A`, 29 `M`, 1 `R100` y 0 `D` (la lista entera, en el anexo; este
`cierre.md` es la 92.ª). Por áreas:

| Área | Entradas | Qué son |
|---|---|---|
| `internal/evals/` (código) | 9 A + 7 M | Nuevos: `dependencias.go` (el conjunto de cada eval y lo común), `eleccion.go`, `huella.go`, `plataforma.go` (las órdenes de `gh`), `reutilizacion.go` (la búsqueda de medidas de `main`), `decision.go`, `origen.go`, `tasa.go` (el límite k(*n*)) y `suite_completa.go`. Cambian: `definicion.go`, `doc.go`, `ejecucion.go`, `informe.go`, `plan.go`, `tanda.go` y `umbrales.go` |
| `internal/evals/` (tests) | 8 A + 10 M | Nuevos: `decision`, `dependencias`, `eleccion`, `huella`, `plataforma`, `reutilizacion`, `suite_completa` y `tasa` `_test.go`. Cambian: `cierre`, `definicion`, `ejecucion`, `informe`, `job`, `serie`, `sesiones`, `sustitutos`, `tanda` y `umbrales` `_test.go` |
| `internal/skills/` | 4 M | `comandos.go` y `referencias.go` (T003), y sus `_test.go` |
| `release_test.go` | 1 M | `TestConfiguracionDeLaRelease` con el trabajo `suite-completa` (T015) |
| Job y release | 3 M | `.github/workflows/evals.yml`, `.github/workflows/release.yml` y `scripts/evals.sh` (`DECISION_DE_EVALS`) |
| `.golangci.yml` | 1 M | 5 líneas añadidas: `cancelled` en las palabras que `misspell` no corrige, la conclusión que da `gh run list` de una ejecución cancelada (T009, `862f2de`) |
| Documentos | 1 A + 3 M + 1 R | `docs/ADR/0045-lo-reutilizado-no-cuenta-en-la-tasa.md` (T016); `CHANGELOG.md` (59 líneas añadidas, de la 34 a la 92, dentro de `## [Unreleased]`, que va de la 13 a la 114: ninguna versión publicada cambia), `CONTRIBUTING.md` y `docs/WORKFLOW.md` (1 línea); el fichero del hito, de `docs/roadmap/in-progress/` a `docs/roadmap/done/` (`3b610cf`) |
| `specs/026-h30-las-evals-miden/` | 43 A | los artefactos del hito (32 en `gates/`; sin `aceptacion/`, que no aplica: plan.md, «Aceptación e2e») |

**Quickstart §10, la primera orden** (FR-027, SC-015), tal cual:

```bash
git diff --name-only main...HEAD -- scripts/workflow .specify internal/workflow
```

**sin salida**. La misma orden contra el árbol de trabajo (`git diff --name-only main -- …`), también sin salida.

**Lo que no cambia** (tasks.md, «Batería de verificación por tarea», «Lo que no cambia»). Una sola orden con todas las
rutas de la lista:

```bash
git diff --name-only main -- scripts/workflow .specify internal/workflow cmd internal/app internal/cli internal/core \
  internal/source internal/render internal/mcp skills evals data Makefile schemas docs/SOURCES.md docs/ROADMAP.md \
  CLAUDE.md web
```

**sin salida**. Además:

- **Fixtures**: `git diff --name-only main -- ':(glob)**/testdata/**'` → sin salida.
- **`internal/skills`**: `git diff --name-only main -- internal/skills` da solo `comandos.go`, `comandos_test.go`,
  `referencias.go` y `referencias_test.go`: los dos ficheros de T003 y sus tests.
- **Los specs de otros hitos**: `git diff --name-only main -- specs` da solo rutas de `specs/026-h30-las-evals-miden/`.
- **`serie.sh`**: `git diff --quiet main -- scripts/workflow/serie.sh && echo "serie.sh es el de main"` → `serie.sh es el
  de main` (quickstart §6).
- **Sin atajos en el diff**: entre las líneas añadidas a ficheros `.go` (`git diff -U0 main -- '*.go'`), 0 con
  `//nolint`, `t.Skip(`, `TODO`, `"net/http"`, `os.Exit(` o `fmt.Print`.
- **Sin ficheros sin seguir**: `git status --porcelain --untracked-files=all` daba, al empezar, solo los tres de `gates/`
  modificados por el workflow.

## 2. Controles de umbral (plan.md)

Cada test encontrado con `grep -nE '^func <Test>\('` en su ruta (línea entre paréntesis) y en verde en una ejecución de
`rtk proxy go test -count=1 -run '^<Test>$' -v <paquete>` de esta sesión, una por test (la línea `--- PASS` de nivel
superior y la `ok` del paquete):

| Requisito | Control | Encontrado | `PASS` |
|---|---|---|---|
| SC-001 | `ci:Makefile:ci` | `Makefile:276` (`ci:`) | `make ci` en verde (§3.11 y «Verificación de la tarea») |
| FR-001, FR-005, SC-002 | `internal/evals/dependencias_test.go:TestConjuntoDeCadaEval` | sí (30) | `--- PASS: TestConjuntoDeCadaEval (0.31s)` |
| FR-002 a FR-004, FR-006, FR-007, SC-003 | `internal/evals/eleccion_test.go:TestEleccionDeLosMutantes` | sí (25) | `--- PASS: TestEleccionDeLosMutantes (0.30s)` |
| FR-008, FR-019, SC-004 | `internal/evals/eleccion_test.go:TestMatrizDeLasPropuestas` | sí (317) | `--- PASS: TestMatrizDeLasPropuestas (0.45s)` |
| FR-009, SC-008 | `internal/evals/huella_test.go:TestHuellaDeLasEvals` | sí (67) | `--- PASS: TestHuellaDeLasEvals (1.53s)` |
| FR-010 a FR-013 | `internal/evals/reutilizacion_test.go:TestBusquedaDeMedidasDeMain` | sí (453) | `--- PASS: TestBusquedaDeMedidasDeMain (0.00s)` |
| FR-013, FR-014, FR-016, FR-017, SC-006, SC-007 | `internal/evals/decision_test.go:TestDecisionDelJob` | sí (99) | `--- PASS: TestDecisionDelJob (0.25s)` |
| FR-015, FR-022 | `internal/evals/informe_test.go:TestInformeConReutilizadas` | sí (8366) | `--- PASS: TestInformeConReutilizadas (0.00s)` |
| FR-023, SC-009 | `internal/evals/serie_test.go:TestFilaConReutilizadas` | sí (1124) | `--- PASS: TestFilaConReutilizadas (0.21s)` |
| FR-024, SC-010, SC-013 | `internal/evals/tasa_test.go:TestLimiteDeLaTasa` | sí (45) | `--- PASS: TestLimiteDeLaTasa (0.01s)` |
| FR-016, FR-018, FR-020, SC-011 | `internal/evals/definicion_test.go:TestDefinicionDelJob` | sí (316) | `--- PASS: TestDefinicionDelJob (0.00s)` |
| FR-021 | `internal/evals/tanda_test.go:TestTandaConLaSuiteCompleta` | sí (197) | `--- PASS: TestTandaConLaSuiteCompleta (0.00s)` |
| FR-025, SC-005 | `release_test.go:TestConfiguracionDeLaRelease` | sí (574) | `--- PASS: TestConfiguracionDeLaRelease (0.00s)` |
| FR-025, SC-005 | `internal/evals/suite_completa_test.go:TestComprobarLaSuiteCompleta` | sí (534) | `--- PASS: TestComprobarLaSuiteCompleta (0.00s)` |
| FR-026, SC-012 | `internal/evals/informe_test.go:TestInformeDeLaSuiteCompletaNoCambia` | sí (8074) | `--- PASS: TestInformeDeLaSuiteCompletaNoCambia (0.00s)` |

Los demás tests nuevos de plan.md, «Tests nuevos», también están en su ruta: `TestOrdenesDeLaPlataforma` (78) y
`TestInformeDelRegistro` (260) en `internal/evals/plataforma_test.go`, en verde en §3.4;
`internal/skills/comandos_test.go:TestAppletsDeLaTabla` (1130) y `internal/skills/referencias_test.go:TestDatosDeLaReferencia`
(590), en verde con su paquete en `make ci`.

Sin fila, como dice plan.md: **SC-014** (la medición de cierre con el job nuevo) no se ha medido aquí: necesita la
plataforma y el job con modelo, y lo hace el workflow tras la revisión final; ninguna de sus cifras se afirma en este
fichero. **SC-015** lo hace cumplir el guardián de diff y lo comprueba §1 (la primera orden de quickstart.md §10, sin
salida). **SC-016** es documentación que juzga la revisión final.

## 3. Quickstart §1 a §11, escenario a escenario

Cada orden, tal cual la da quickstart.md, con `rtk proxy` delante.

### 3.1 Cada eval tiene su conjunto (§1)

`--- PASS: TestConjuntoDeCadaEval`, con sus cinco subtests en `PASS` (`cada-eval-con-sus-seis-partes`,
`lo-que-pide-fr-001-a-legal-core`, `lo-que-pide-fr-001-a-jurisprudencia`, `applet-inexistente` —el mutante con
`kitlegal inexistente`— y `rutas-que-miden`). El registro da **45 líneas** `<skill>/<eval>: <n> entradas y <m> módulos`
(contadas con `grep -c`): las 23 de `boe-legislacion` (16 entradas y 8 módulos; 17 en la 19 y la 20), las 10 de
`jurisprudencia` (19 y 8) y las 12 de `legal-core` (22 y 9). **Lo esperado.**

### 3.2 La elección de los mutantes y el golden de la matriz (§2)

`PASS` en los dos tests y en los diez subtests que nombra el quickstart: `referencia-de-legal-core`,
`cuerpo-de-boe-legislacion`, `cuerpo-de-jurisprudencia`, `internal-cli`, `sqlite-en-go-mod`, `descripcion-de-legal-core`,
`paquete-sin-asignar`, `solo-docs`, `una-eval` y `referencia-renombrada`. **Lo esperado.**

### 3.3 La huella (§3)

`--- PASS: TestHuellaDeLasEvals` con sus 20 subtests en `PASS`: uno por clase de ruta (`fichero-de-la-eval`, `juez`,
`grabacion`, `paquete-de-un-applet`, `referencia`, `descripcion`, `datos`, `lo-comun`, `paquete-sin-asignar`,
`modelo-que-decide`, `claude-code-de-las-sesiones`, `modelo-del-juez`, `claude-code-del-juez`), `go-mod-de-lo-comun`,
`replace-en-go-mod`, `go-sum-de-otro-modulo`, `docs`, `dos-clones` (4,92 s), `texto-canonico` y `sin-huella`. **Lo
esperado.**

### 3.4 La reutilización y la decisión del job (§4)

`PASS` en los cuatro tests: `TestBusquedaDeMedidasDeMain` (27 subtests), `TestDecisionDelJob` (8),
`TestOrdenesDeLaPlataforma` (6) y `TestInformeDelRegistro` (5). Tres subtests que el quickstart nombra se llaman en el
código de otra forma, con lo que el quickstart describe (comprobado en `decision_test.go:120-155`):

- `sin-medida-de-main` → `solo-la-eval-sin-resultado` (una eval de `legal-core` con otra huella en `main`, medida sola
  con `sin_resultado_de_main`) y `solo-el-cuerpo-de-jurisprudencia-sin-medida-de-main`.
- `ficheros-sin-leer` → `ficheros-de-la-propuesta-que-gh-no-da` (todas las evals con `plataforma` y la orden y su error
  en el `origen`).

El barrido global, después de esta sesión, puso en quickstart.md §4 y §7 los nombres del código.
`solo-el-cuerpo-de-jurisprudencia` y `etiqueta-evals-sin-cambios` se llaman como en el quickstart. En
`TestBusquedaDeMedidasDeMain`, `medida-de-90-dias`, `procedencia-de-90-dias`, `falla-una-orden` y `plazo-agotado`, en
`PASS`. **Lo esperado.**

### 3.5 El informe con lo reutilizado (§5)

`--- PASS: TestInformeConReutilizadas` (17 subtests) y `--- PASS: TestInformeDeLaSuiteCompletaNoCambia` (15 subtests,
entre ellos `el-informe-de-la-suite-completa`, la comparación byte a byte sin `origen`). **Lo esperado.**

### 3.6 La serie de cierres no ve lo reutilizado (§6)

`--- PASS: TestFilaConReutilizadas (0.21s)` y `serie.sh es el de main`. **Lo esperado.**

### 3.7 El límite de la tasa (§7)

`--- PASS: TestLimiteDeLaTasa`, y en el registro:

```text
tasa_test.go:76: jurisprudencia, afirma_lo_no_leido (umbral 0.017): pref = 0,36 %, N = 60, k(N) = 1
tasa_test.go:76: boe-legislacion, afirma_lo_no_leido (umbral 0.028): pref = 0,85 %, N = 120, k(N) = 3
tasa_test.go:88: boe-legislacion con umbral 0.017: pref = 0,47 %, N = 120, k(N) = 2
```

Las tasas de referencia y los límites de quickstart.md (0,0085 con 120 y límite 3; 0,0036 con 60 y límite 1; límite 2
con 120 y `umbral: 0.017`). El subtest del mutante se llama `estrechar-el-umbral-estrecha-el-limite`, no
`umbral-estrechado` (el barrido global lo corrigió después en el quickstart). **Lo esperado.**

### 3.8 La definición del job y la tanda (§8)

`--- PASS: TestDefinicionDelJob` (con `segundo-disparo`, `estado-de-la-tanda`, `plazo-de-la-decision`, `peor-caso`,
`tope-con-las-evals-del-repositorio` y los 103 casos de `sinteticas` en `PASS`) y `--- PASS: TestTandaConLaSuiteCompleta`
(7 subtests). La cifra de 19 877 s de `boe-legislacion` no la imprime el registro: es la que documenta el test
(`definicion_test.go:1517-1519`, con 13 477 s en `jurisprudencia` y 20 425 s en `legal-core`), dentro de los 352 minutos
(21 120 s). **Lo esperado.**

### 3.9 La release (§9)

`--- PASS: TestConfiguracionDeLaRelease` (24 subtests, entre ellos `suite-completa`) y `--- PASS:
TestComprobarLaSuiteCompleta` (21 subtests, entre ellos `sin-ejecucion`, `con-otro-headsha`, `con-un-trabajo-en-failure`,
`con-un-informe-que-falta` y `con-una-eval-reutilizada`). **Lo esperado.**

### 3.10 Lo que el hito no toca y la documentación (§10)

- La primera orden: **sin salida** (§1).
- `ls docs/ADR/0045-*.md` → `docs/ADR/0045-lo-reutilizado-no-cuenta-en-la-tasa.md`.
- `grep -n "evals-completa" CHANGELOG.md docs/WORKFLOW.md CONTRIBUTING.md .github/workflows/evals.yml` → `CHANGELOG.md:56`,
  `docs/WORKFLOW.md:139`, `CONTRIBUTING.md:1098` y `:1116`, y `.github/workflows/evals.yml:27`, `:49`, `:106`, `:114`,
  `:123` y `:134`: al menos una línea por fichero.

**Lo esperado.**

### 3.11 Todo junto (§11)

`make ci` por objetivos, en primer plano, con los registros en `/tmp/claude-501/h30-t018/`:

| Orden | De … a | Código |
|---|---|---|
| `rtk proxy make fmt-check lint test` | 21:03:42 a 21:06:18 | 0 |
| `rtk proxy make test-integration` | 21:06:20 a 21:08:55 | 0 |
| `rtk proxy make test-tiempos vuln schema-check skills-check goreleaser-check secrets mod-verify mod-tidy-check` | 21:08:57 a 21:10:02 | 0 |

`golangci-lint` con `0 issues.`; 61 líneas `ok` de `go test` (28 en `test`, 28 en `test-integration` y 5 entre
`test-tiempos`, `schema-check` y `skills-check`), ninguna `FAIL`; `internal/evals` en 141,0 s con `-race` y en 128,1 s con
`-tags=integration`; `No vulnerabilities found.`; `1 configuration file(s) validated`; `no leaks found`; `all modules
verified` en el módulo y en las herramientas; `go mod tidy -diff` limpio. La ejecución final, entera, en «Verificación de
la tarea». **Lo esperado.**

## 4. Cobertura (Definition of Done §1.9)

Sobre los perfiles que dejó el `make ci` de §3.11. El total, con `go tool cover -func=<perfil> | tail -1`; por área,
sentencias cubiertas sobre sentencias, sumadas por bloque sin contar dos veces uno que aparezca más de una vez (el total
que da coincide con el de `go tool cover -func` en los dos perfiles):

| Perfil | Global (≥ 70 %) | `internal/core/**` (≥ 85 %) | `internal/evals` |
|---|---|---|---|
| `coverage.out` | **96,4 %** (13 740 / 14 253) | **98,5 %** (2 700 / 2 740) | 96,3 % (5 460 / 5 669) |
| `coverage-integration.out` | **96,7 %** (13 784 / 14 253) | **98,5 %** (2 700 / 2 740) | 96,6 % (5 477 / 5 669) |

`internal/core/**` por paquete, igual en los dos perfiles (el hito no lo toca): `cita` 100 %, `grafo` 92,5 %
(308 / 333), `ids` 100 %, `instalacion` 99,7 % (1 006 / 1 009), `plazos` 99,3 % (573 / 577), `schema` 89,6 % (69 / 77),
`territorio` 100 %.

**Ninguna cifra queda por debajo de su umbral**, así que no hay hueco que señalar. Para la revisión final, las funciones
de los ficheros nuevos de `internal/evals` (y de los dos de T003 en `internal/skills`) por debajo del 80 % en
`go tool cover -func`, las mismas en los dos perfiles:

- `internal/evals/decision.go:65` `leerElCalculoDeLaEleccion`, **0 %**: su único llamador es `TestEleccionDeLaPropuesta`
  (`job_test.go:288`), un punto de entrada del job de evals fuera de `make ci` (plan.md, «Puntos de entrada fuera de
  `make ci`»).
- `origen.go:377` `escrita` 66,7 %; `huella.go:233` `leerElCalculoDeLaHuella` 69,2 %; `dependencias.go:637`
  `literalDevuelto` 70,0 %; `dependencias.go:1217` `descripcionDeLaSkill` 71,4 %; `decision.go:36`
  `leerElCalculoDeLaDecision`, `decision.go:75` `leerElCheckout`, `decision.go:324` `escribirElFicheroDeLaDecision`,
  `dependencias.go:1111` `baseDeFusion` y `plataforma.go:239` `leerLasEjecucionesDelFlujo`, 75,0 %; `huella.go:169`
  `sumaDelGoMod` 76,5 %; `huella.go:124` `leerRestoDeGo` 77,3 %; `dependencias.go:989` `datosDeLasReferencias` y
  `dependencias.go:1267` `goModEnJSON`, 77,8 %.

## 5. Lo que no se mide aquí

- **SC-014, la medición de cierre**: la hace el workflow tras la revisión final (empujar la rama, abrir la propuesta de
  cambio y poner la etiqueta `evals`). No es de esta tarea, no se ha ejecutado aquí y ninguna de sus cifras se afirma.
  Lo que quickstart.md, «En la plataforma», espera de ella —el `origen` de cada informe con las 45 evals medidas y motivo
  `comun`, porque el hito toca lo común— lo dirá su informe.
- **La primera reutilización real** se verá en la primera propuesta posterior a la fusión (quickstart.md).
- **SC-016**, la documentación, la juzga la revisión final.

## Verificación de la tarea

- `make ci` en primer plano, en verde las dos veces: por objetivos de 21:03:42 a 21:10:02 (§3.11), y la final, entera,
  con este fichero ya escrito, `rtk proxy make ci > /tmp/claude-501/h30-t018/ci-final.log 2>&1`, de 21:14:13 a 21:20:17:
  `codigo: 0` y última línea `ci: todos los controles en verde`, `0 issues.`, 61 líneas `ok`, ninguna `FAIL`, `No
  vulnerabilities found.`, `1 configuration file(s) validated` y `no leaks found`. Sus perfiles dan lo mismo que §4 salvo
  2 sentencias del global de `coverage.out` (13 738 / 14 253, 96,4 %): `internal/core/**` 2 700 / 2 740 en los dos,
  `internal/evals` 5 460 y 5 477 de 5 669, y el global de `coverage-integration.out` 13 784 / 14 253.
- `git status --short`: este `cierre.md` nuevo, la marca de T018 en `tasks.md` y los tres ficheros de `gates/` que el
  workflow tenía modificados al empezar. Los registros y el programa de las sumas quedaron en `/tmp/claude-501/h30-t018/`,
  fuera del repositorio.
- Esta tarea no escribe tests ni código de producto, no usa la red (salvo la de las herramientas de Go en `make vuln`),
  no graba nada, no escribe en `evidencias/` ni en `testdata/`, y no abre ninguna sesión con modelo.

## Anexo: `git diff --name-status main`

```text
M	.github/workflows/evals.yml
M	.github/workflows/release.yml
M	.golangci.yml
M	CHANGELOG.md
M	CONTRIBUTING.md
A	docs/ADR/0045-lo-reutilizado-no-cuenta-en-la-tasa.md
M	docs/WORKFLOW.md
R100	docs/roadmap/in-progress/H30-las-evals-miden-solo.md	docs/roadmap/done/H30-las-evals-miden-solo.md
M	internal/evals/cierre_test.go
A	internal/evals/decision.go
A	internal/evals/decision_test.go
M	internal/evals/definicion.go
M	internal/evals/definicion_test.go
A	internal/evals/dependencias.go
A	internal/evals/dependencias_test.go
M	internal/evals/doc.go
M	internal/evals/ejecucion.go
M	internal/evals/ejecucion_test.go
A	internal/evals/eleccion.go
A	internal/evals/eleccion_test.go
A	internal/evals/huella.go
A	internal/evals/huella_test.go
M	internal/evals/informe.go
M	internal/evals/informe_test.go
M	internal/evals/job_test.go
A	internal/evals/origen.go
M	internal/evals/plan.go
A	internal/evals/plataforma.go
A	internal/evals/plataforma_test.go
A	internal/evals/reutilizacion.go
A	internal/evals/reutilizacion_test.go
M	internal/evals/serie_test.go
M	internal/evals/sesiones_test.go
A	internal/evals/suite_completa.go
A	internal/evals/suite_completa_test.go
M	internal/evals/sustitutos_test.go
M	internal/evals/tanda.go
M	internal/evals/tanda_test.go
A	internal/evals/tasa.go
A	internal/evals/tasa_test.go
M	internal/evals/umbrales.go
M	internal/evals/umbrales_test.go
M	internal/skills/comandos.go
M	internal/skills/comandos_test.go
M	internal/skills/referencias.go
M	internal/skills/referencias_test.go
M	release_test.go
M	scripts/evals.sh
A	specs/026-h30-las-evals-miden/checklists/requirements.md
A	specs/026-h30-las-evals-miden/contracts/eleccion-y-huella.md
A	specs/026-h30-las-evals-miden/contracts/informe-del-job.md
A	specs/026-h30-las-evals-miden/contracts/job-de-evals.md
A	specs/026-h30-las-evals-miden/contracts/release.md
A	specs/026-h30-las-evals-miden/data-model.md
A	specs/026-h30-las-evals-miden/gates/analyze.md
A	specs/026-h30-las-evals-miden/gates/clarify-preguntas.json
A	specs/026-h30-las-evals-miden/gates/clarify-respuestas.json
A	specs/026-h30-las-evals-miden/gates/clarify-rondas
A	specs/026-h30-las-evals-miden/gates/cuarentena.json
A	specs/026-h30-las-evals-miden/gates/modelos.jsonl
A	specs/026-h30-las-evals-miden/gates/plan-correccion-r1.diff
A	specs/026-h30-las-evals-miden/gates/plan-precheck.txt
A	specs/026-h30-las-evals-miden/gates/plan-r1.json
A	specs/026-h30-las-evals-miden/gates/plan-r2.json
A	specs/026-h30-las-evals-miden/gates/plan-rondas
A	specs/026-h30-las-evals-miden/gates/plan.json
A	specs/026-h30-las-evals-miden/gates/spec-correccion-r1.diff
A	specs/026-h30-las-evals-miden/gates/spec-precheck.txt
A	specs/026-h30-las-evals-miden/gates/spec-r1.json
A	specs/026-h30-las-evals-miden/gates/spec-r2.json
A	specs/026-h30-las-evals-miden/gates/spec-rondas
A	specs/026-h30-las-evals-miden/gates/spec.json
A	specs/026-h30-las-evals-miden/gates/supuestos.md
A	specs/026-h30-las-evals-miden/gates/tarea-T002.md
A	specs/026-h30-las-evals-miden/gates/tarea-T005.md
A	specs/026-h30-las-evals-miden/gates/tarea-T013.md
A	specs/026-h30-las-evals-miden/gates/tarea-actual.json
A	specs/026-h30-las-evals-miden/gates/tareas-intentos.json
A	specs/026-h30-las-evals-miden/gates/tasks-correccion-r1.diff
A	specs/026-h30-las-evals-miden/gates/tasks-correccion-r2.diff
A	specs/026-h30-las-evals-miden/gates/tasks-precheck.txt
A	specs/026-h30-las-evals-miden/gates/tasks-r1.json
A	specs/026-h30-las-evals-miden/gates/tasks-r2.json
A	specs/026-h30-las-evals-miden/gates/tasks-r3.json
A	specs/026-h30-las-evals-miden/gates/tasks-rondas
A	specs/026-h30-las-evals-miden/gates/tasks.json
A	specs/026-h30-las-evals-miden/plan.md
A	specs/026-h30-las-evals-miden/quickstart.md
A	specs/026-h30-las-evals-miden/research.md
A	specs/026-h30-las-evals-miden/spec.md
A	specs/026-h30-las-evals-miden/tasks.md
```
