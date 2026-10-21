# Cierre de la Definition of Done · H26.1 (T016)

Sesión del 2026-10-21 sobre la cabeza `4eb8583` (T015) de `025-h26-1-las-evals-sirven`, con `main` en `6bf2da3`, que es
también la base de la rama (`git merge-base HEAD main`): `git diff --name-status main` es lo que traen los 19 commits de
la rama, más los tres ficheros de `gates/` que el workflow tenía modificados al empezar la tarea (`modelos.jsonl`,
`tarea-actual.json`, `tareas-intentos.json`). Todo lo que se afirma aquí se ejecutó y se vio terminar en esta sesión, o
está en el repositorio; lo que no se pudo medir se dice tal cual. Ninguna tarea está en cuarentena
(`gates/cuarentena.json`, `"tareas": {}`), T001 a T015 están en `[X]` y la checklist del spec tiene 16 de 16.

Cómo se ejecutó (desviaciones de entorno respecto al quickstart, ninguna de producto):

- El sandbox de la sesión no deja comprobar antes de ejecutarla una orden con una variable (`$TMPDIR`, `$?`, `$copia`,
  la de un `for`). Los registros fueron a `/tmp/claude-501/` (el `$TMPDIR` de la sesión) con la ruta escrita entera, y el
  árbol aparte de §7, a `/tmp/claude-501/h261-copia`; el código de `make ci` se lee en su última línea.
- Las salidas de `go test`, `git diff` y `make` se leyeron con `rtk proxy`.
- Las sumas por paquete de §2 se hicieron con un `awk` en línea sobre los perfiles, por bloque y sin contar dos veces uno
  que aparezca más de una vez; el total que da coincide con el de `go tool cover -func` en los dos perfiles.

## 1. Lo creado o modificado, y lo que no cambia

`git diff --name-status main`: 81 entradas, 48 `A`, 33 `M`, 0 `D` (la lista entera, en el anexo; este `cierre.md` es la
82.ª). Por áreas:

| Área | Entradas | Qué son |
|---|---|---|
| `internal/evals/` (código) | 9 M | `consultas.go` (lo que la skill lleva a leer, derivado de `data/`, y `lecturas`), `formato.go` (`lecturas` y `sin_servir`), `preparar.go` (la caché sin lo de `sin_servir`), `medida.go` (la base de la reconstrucción con lo de `data/`), `juzgar.go` (`lecturas_sin_servir`, el `motivo` y el `bash` de las fallidas), `sesion.go` (`SinSobre`, `OrdenesDeBash`), `umbrales.go` (`fuera_de_lo_grabado` y `lectura_sin_intentar`), `informe.go` y `doc.go` |
| `internal/evals/` (tests) | 14 M | `conjunto`, `consultas`, `definicion`, `ejecucion`, `formato`, `grabaciones`, `informe`, `juzgar`, `medida`, `plan`, `preparar`, `sesion`, `sondeo` y `umbrales` `_test.go` |
| `internal/evals/testdata/` | 4 A + 2 M | los bloques `a15` de `BOE-A-2004-4214`, `a18` de `BOE-A-2006-20764` y `a26` y `a27` de `BOE-A-2015-10566`, y el `robots.txt` regrabado, que dejó `grabar_datos` sin modelo (`f1e4bdf`; `gates/grabaciones.md`); y el manifiesto `grabaciones.json` (T002) |
| `schemas/` | 1 M | `eval.yaml.json`: `lecturas` y `sin_servir` (T001; 29 líneas más, 1 menos) |
| `evals/` | 2 A + 3 M | las evals 11 y 12 de `legal-core`; `lecturas:` en la 14, la 15 y la 16 de `boe-legislacion` (solo líneas añadidas: 4, 4 y 7) |
| `skills/legal-core/` | 1 M | `SKILL.md` v1.2.3, los pasos 7 y 10 (299 líneas; 18 más y 18 menos) |
| Documentos | 2 M | `CHANGELOG.md` (76 líneas añadidas, todas entre las líneas 15 y 118, dentro de `## [Unreleased]`, que va de la 13 a la 119: ninguna versión publicada cambia) y `CONTRIBUTING.md` (T014) |
| `go.mod` | 1 M | `toolchain go1.27.2` → `go1.27.3`, del commit `801c435 chore(deps): go1.27.3`, anterior a T001 y fuera de las tareas |
| `specs/025-h26-1-las-evals-sirven/` | 42 A | los artefactos del hito (30 en `gates/`; sin `aceptacion/`, que no aplica) |

**Quickstart §9** (FR-028; D8 y D15 de research), la orden tal cual:

```bash
git diff --name-only main -- cmd internal/app internal/cli internal/render internal/source internal/core internal/httpx internal/cache internal/graph internal/mcp .github testdata evidencias
```

**sin salida**. El binario, sus códigos de salida, su presentador y el job no cambian; ni las grabaciones de H4
(`internal/source/boe/testdata/`) ni las de H5 (`testdata/evals/`) ni ninguna evidencia.

**Lo que no cambia** (tasks.md, «Batería de verificación por tarea»). Una sola orden con todas las rutas de la lista:

```bash
git diff --name-only main -- cmd internal/app internal/cli internal/render internal/source internal/core \
  internal/httpx internal/cache internal/graph internal/mcp .github scripts .specify data testdata evidencias \
  docs/SOURCES.md docs/ROADMAP.md docs/ADR docs/WORKFLOW.md CLAUDE.md web Makefile skills/boe-legislacion \
  skills/jurisprudencia evals/jurisprudencia evals/boe-legislacion/juez internal/evals/juez.go \
  internal/evals/sondeo.go internal/evals/definicion.go | wc -l
```

dio **0**. Además:

- **Las evals de hoy**: `git diff --name-status main -- evals/` da solo `M` de la 14, la 15 y la 16 de
  `boe-legislacion` y `A` de la 11 y la 12 de `legal-core`; ninguna otra de `boe-legislacion` ni ninguna de las diez de
  `legal-core` de hoy.
- **Los specs de otros hitos**: `git diff --name-only main -- specs/ | grep -vc '^specs/025-h26-1-las-evals-sirven/'`
  → 0.
- **Sin atajos en el diff**: entre las líneas añadidas a ficheros `.go`, 0 con `"net/http"`, `os.Exit(` o `fmt.Print`
  (Definition of Done §1.3) y 0 con `//nolint`, `t.Skip(` o `TODO`.
- **FR-028 en el juicio**: lo comprueba, sesión a sesión, `TestJuzgarElMotivoDeLasFallidas` (`pasa` y `motivos` iguales
  a los de la misma sesión sin los campos nuevos), en verde en §3.3; y el código con que termina una orden que no
  puede escribir su salida no cambia, porque `internal/render` e `internal/app` no están en el diff (§9).

## 2. Cobertura (Definition of Done §1.9)

Sobre los perfiles que dejó el `make ci` de esta sesión (13:09:42 a 13:15:29). El total, con
`go tool cover -func=<perfil> | tail -1`; por área, sentencias cubiertas sobre sentencias:

| Perfil | Global (≥ 70 %) | `internal/core/**` (≥ 85 %) | `internal/evals` |
|---|---|---|---|
| `coverage.out` | **96,8 %** (12 559 / 12 978) | **98,5 %** (2 700 / 2 740) | 97,4 % (4 301 / 4 414) |
| `coverage-integration.out` | **97,2 %** (12 607 / 12 978) | **98,5 %** (2 700 / 2 740) | 97,8 % (4 318 / 4 414) |

`internal/core/**` por paquete, igual en los dos perfiles (el hito no lo toca): `cita` 100 %, `grafo` 92,5 %
(308 / 333), `ids` 100 %, `instalacion` 99,7 % (1 006 / 1 009), `plazos` 99,3 % (573 / 577), `schema` 89,6 % (69 / 77),
`territorio` 100 %.

**Ninguna cifra queda por debajo de su umbral**, así que no hay hueco que pase a la revisión final por esta vía. Para su
lectura, los bloques del código que el hito escribe o toca (cruzados con las líneas de `git diff -U0 main`) que ningún
test recorre (`count` 0 en los dos perfiles), todos de una sentencia y todos ramas de defensa:

- `internal/evals/juzgar.go:1156-1158` — en `traeLaLineaDeLaSalidaNoEscrita`, la rama del prefijo vacío.
- `internal/evals/juzgar.go:1182-1184` y `:1187-1189` — en `prefijoDeLaSalidaNoEscrita`, el presentador que no da error
  con una salida estándar que falla, o uno que no termina en la causa. Las dos dejan el prefijo vacío, y con él toda
  invocación como antes de H26.1; `TestLineaDeLaSalidaNoEscrita` exige que, con el binario de hoy, no lo sea.
- `internal/evals/medida.go:1074-1076` — en `leerLasEvalsDeHoy`, el error de `consultasDeLaCache` con un `data/` que no se
  lee.
- `internal/evals/umbrales.go:955-957` — en `lecturasMedidas.umbrales`, el receptor `nil`.

## 3. Quickstart §0 a §10, orden a orden

### 3.0 Antes de empezar (§0)

`go version` → `go version go1.27.3 darwin/arm64`; `git status --short` → los tres ficheros de `gates/` que el workflow
tenía modificados; `git worktree list` → un solo worktree, `~/Projects/kitlegal 4eb8583 [025-h26-1-las-evals-sirven]`.
Esperado: Go 1.27, lo que haya sin commitear y un solo worktree. **Lo esperado.**

### 3.1 La caché sirve lo que la skill lleva a leer (§1)

- `go test -count=1 -v -run '^TestEvalsDelRepositorio$' ./internal/evals/` → `--- PASS: TestEvalsDelRepositorio` y
  `ok … internal/evals 0.727s`, con sus 19 subtests en `PASS`, entre ellos `lo-que-lleva-a-leer` (con
  `lo-que-lleva-a-leer/boe-legislacion`, `/legal-core` y `/jurisprudencia` en `PASS`), `grabado`, `lecturas` y
  `conjunto-legal-core`. Esperado: `ok`, con esos subtests en `PASS`. **Lo esperado.** La cifra de «17 bloques y 19
  normas» del quickstart no se volvió a contar en esta sesión: el subtest comprueba cada lectura que enumera
  `loQueLlevaALeerElRepositorio` (`grabaciones_test.go:1112`) sin imprimir cuántas son.
- `go test -count=1 -v -run '^(TestCacheDeLasSesiones|TestLoQueLlevaALeerSinGrabar|TestBaseDeLaReconstruccion|TestConsultasNecesarias)$' ./internal/evals/`
  → los cuatro en `PASS` y `ok … 0.962s`; 20 subtests en `PASS` y 0 en `FAIL`, entre ellos
  `TestCacheDeLasSesiones/sin-servir`, `TestCacheDeLasSesiones/lecturas`,
  `TestBaseDeLaReconstruccion/sirve-lo-que-la-skill-lleva-a-leer`, `TestConsultasNecesarias/lecturas-de-la-eval` y
  `TestConsultasNecesarias/sin-servir`. Esperado: `ok`. **Lo esperado.**

### 3.2 `lecturas` y `sin_servir` en el formato (§2)

`go test -count=1 -v -run '^(TestFormatoDeLecturas|TestLeerEval|TestEsquemaDeEval)$' ./internal/evals/` → los tres en
`PASS` y `ok … 0.359s`; 90 subtests en `PASS`, 0 en `FAIL`. Esperado: `ok`. **Lo esperado.**

### 3.3 El juicio (§3)

`go test -count=1 -v -run '^(TestJuzgarLasLecturasSinServir|TestJuzgarElMotivoDeLasFallidas|TestLineaDeLaSalidaNoEscrita|TestLeerLasLlamadas|TestTextosDeLaSesion|TestJuzgar|TestJuzgarLasLlamadas|TestJuzgarPlazos)$' ./internal/evals/`
→ los ocho en `PASS` y `ok … 0.388s`; 220 subtests en `PASS`, 0 en `FAIL`. Esperado: `ok`. **Lo esperado.**

### 3.4 El informe (§4)

`go test -count=1 -v -run '^(TestUmbralesDeLasLecturas|TestUmbralesDeLasTresSkills|TestUmbralesDeLegalCore|TestUmbralesDeBoeLegislacion|TestUmbralesDeJurisprudencia|TestUmbralesDelInforme|TestInformeEnDosModos|TestInforme)$' ./internal/evals/`
→ los ocho en `PASS` y `ok … 6.147s`; 104 subtests en `PASS`, 0 en `FAIL`. Esperado: `ok`, con `legal-core` en 18
umbrales y 12 que deciden y las otras dos en 19 y 11, sin `lectura_sin_intentar`: es lo que exige
`TestUmbralesDeLasTresSkills` (`informe_test.go:6908-6909`, `:6928` y `:6948`), en verde. **Lo esperado.**

### 3.5 El tope del trabajo de `legal-core` (§5)

- `go test -count=1 -v -run '^TestDefinicionDelJob$' ./internal/evals/` → `--- PASS: TestDefinicionDelJob` y
  `ok … 0.454s` (96 subtests en `PASS`, 0 en `FAIL`).
- `grep -n 'peorCasoDeLegalCore = ' internal/evals/definicion_test.go` →
  `1898:	peorCasoDeLegalCore = "19525 s = 485 s + (\xe2\x8c\x8867 / 2\xe2\x8c\x89 + \xe2\x8c\x8866 / 2\xe2\x8c\x89 + " +`.
- `grep -n 'timeout-minutes: 352' .github/workflows/evals.yml` → `226:    timeout-minutes: 352`.

Esperado: `ok`, el literal de 19 525 s con los caracteres escritos `\xe2\x8c\x88` y `\xe2\x8c\x89`, y el tope, 352, sin
cambiar (19 525 s caben en 21 120 s). **Lo esperado.**

### 3.6 `legal-core` v1.2.3 (§6)

- `wc -l skills/legal-core/SKILL.md` → `299`.
- `make skills-check` → `ok` en `internal/app` (2.667s), `internal/skills` (2.879s) e `internal/evals` (1.551s).
- `grep -n 'legal-core. v1.2.3' CHANGELOG.md` → `108:- **\`legal-core\` v1.2.3: qué hace la respuesta con una lectura
  fallida en los pasos 7 y 10 (ADR 0043, puntos 6 y …`, dentro de `## [Unreleased]`.

Esperado: 299 líneas o menos, `ok` y la entrada de *Unreleased*. **Lo esperado.** Lo que dicen los pasos 7 y 10
(contracts/skill-legal-core.md §1) lo lee la revisión final.

### 3.7 Un precepto sin grabar y los mutantes, en una copia fuera del árbol (§7)

Ejecutado en esta sesión con `git worktree add --detach /tmp/claude-501/h261-copia HEAD` (cabeza `4eb8583`):

- **a.** `sed -i.orig 's/bloque: a128,/bloque: a129,/'` sobre `data/plazos.yaml` de la copia: sin «LA SUSTITUCIÓN NO SE
  HA APLICADO», y `git diff --no-index -U0` enseña la línea 29 con `bloque: a129`. `go -C … test -count=1 -run
  '^TestEvalsDelRepositorio$' ./internal/evals/` → `--- FAIL: TestEvalsDelRepositorio`,
  `--- FAIL: TestEvalsDelRepositorio/lo-que-lleva-a-leer` y `…/lo-que-lleva-a-leer/jurisprudencia` (la salida se cortó
  con `head` tras la primera skill), con `Should be empty, but was [data/plazos.yaml: lo que la skill lleva a leer boe
  articulo BOE-A-1998-16718 a129: «boe articulo BOE-A-1998-16718 a129» terminó con código 1: …` y la falta de la
  grabación del bloque `a129`. Restaurado con `mv`: `git diff --no-index --exit-code` contra `data/plazos.yaml` del
  árbol de trabajo, 0.
- **b.** `sed -i.orig 's/render: no se pudo escribir en %s/render: no se ha podido escribir en %s/'` sobre
  `internal/render/render.go` de la copia: sin «EL MUTANTE NO SE HA APLICADO», con la línea `return fmt.Errorf("render:
  no se ha podido escribir en %s: %w", nombre, err)`. `TestTuberiaCerrada` → `FAIL … internal/app 3.065s`, cada subtest
  con `"render: no se ha podido escribir en la salida estándar: write /dev/stdout: broken pipe\n" does not contain "no se
  pudo escribir en la salida estándar"`; `TestLineaDeLaSalidaNoEscrita` → `ok … internal/evals 0.507s`. Restaurado con
  `mv`, con `git diff --no-index --exit-code` en 0.
- `git worktree remove --force /tmp/claude-501/h261-copia` → `git worktree list`, el único de §0; `git status --short`,
  los tres de §0.

Esperado: en a, la falta que nombra `a129` y `FAIL`; en b, `TestTuberiaCerrada` en `FAIL` y
`TestLineaDeLaSalidaNoEscrita` en `ok`; y el árbol como en §0. **Lo esperado.** Los otros tres mutantes (FR-002, FR-013 y
FR-033) no se volvieron a ejecutar en esta sesión: los ejecutó T015 y su registro está en `mutantes.md` (secciones 2, 3 y
4), cada uno con al menos un test en rojo.

### 3.8 `make ci` (§8)

`rtk proxy make ci > /tmp/claude-501/h261-ci.log 2>&1`, en primer plano de 13:09:42 a 13:15:29 → última línea
`ci: todos los controles en verde`; `golangci-lint` con `0 issues.`; 61 líneas `ok` de `go test` (las dos pasadas con
`-race` —la segunda con `-tags=integration`—, `test-tiempos`, `schema-check` y `skills-check`), ninguna `FAIL`;
`No vulnerabilities found.`; `goreleaser check` con `1 configuration file(s) validated`; `gitleaks` con `no leaks found`;
`all modules verified` en el módulo y en las cinco herramientas, y `go mod tidy -diff` limpio. Esperado: `código: 0` y
`ci: todos los controles en verde`. **Lo esperado.** La ejecución final, con este fichero ya escrito, en «Verificación».

### 3.9 Lo que no cambia (§9)

Sin salida (§1 de este fichero). **Lo esperado.**

### 3.10 El cierre (§10)

**Lo hace el workflow** tras la revisión final: empujar la rama, abrir la propuesta de cambio y medir el job de evals
sobre la cabeza, que deja `gates/evals/<skill>.json`. No es de esta tarea, no se ha ejecutado aquí y ninguna de sus
cifras (SC-011 a SC-020 en el job) se afirma en este fichero.

## 4. Controles de umbral (plan.md)

Filas `ci:` — cada test encontrado como `^func <Test>\(` en su ruta (línea entre paréntesis) y en verde en esta sesión
(§3.1 a §3.6, `TestTuberiaCerrada` aparte —`ok … internal/app 2.231s`— y `make ci`):

| Requisito | Control | Encontrado | Verde |
|---|---|---|---|
| SC-016 | `internal/evals/informe_test.go:TestInforme` (278) | sí | sí: el caso `llegada-a-la-red` con `comprobarLlegadaALaRed` (`:3282`), cuyo cuerpo es igual al de `main` (comparado con `diff`), que exige la petición en `red`, su motivo y el veredicto `fallo`; T010 solo le añadió, por caso, el umbral `fuera_de_lo_grabado` que espera |
| SC-011, FR-001, FR-003, SC-003 | `internal/evals/conjunto_test.go:TestEvalsDelRepositorio` (1999), `internal/evals/preparar_test.go:TestLoQueLlevaALeerSinGrabar` (995) | sí | sí, y en rojo con el mutante de §3.7.a |
| FR-002, FR-005, SC-008 | `TestEvalsDelRepositorio` (1999), `internal/evals/preparar_test.go:TestCacheDeLasSesiones` (1070) | sí | sí; el mutante de FR-002, en `mutantes.md` §2 (T015) |
| FR-004 | `internal/evals/medida_test.go:TestBaseDeLaReconstruccion` (3080) | sí | sí |
| SC-018, FR-006, FR-010, FR-011, SC-004 | `TestCacheDeLasSesiones` (1070), `TestEvalsDelRepositorio` (1999) | sí | sí |
| FR-013, SC-005, SC-008 | `TestCacheDeLasSesiones` (1070) | sí | sí; el mutante de FR-013, en `mutantes.md` §3 (T015) |
| FR-008, FR-009, FR-012, FR-014, SC-004, SC-005 | `internal/evals/formato_test.go:TestFormatoDeLecturas` (1907), `TestEvalsDelRepositorio` (1999) | sí | sí |
| FR-007, FR-015 a FR-018, SC-005 | `internal/evals/juzgar_test.go:TestJuzgarLasLecturasSinServir` (3966) | sí | sí |
| FR-033, FR-015, SC-008, SC-013 | `internal/evals/umbrales_test.go:TestUmbralesDeLasLecturas` (2200) | sí | sí; el mutante de FR-033, en `mutantes.md` §4 (T015) |
| FR-021, FR-022, SC-006, SC-015 | `TestUmbralesDeLasLecturas` (2200), `internal/evals/informe_test.go:TestUmbralesDeLasTresSkills` (6862) | sí | sí |
| FR-023, FR-024, FR-026 a FR-028, SC-007, SC-017 | `internal/evals/juzgar_test.go:TestJuzgarElMotivoDeLasFallidas` (4480), `internal/evals/sesion_test.go:TestLeerLasLlamadas` (296) | sí | sí |
| FR-025, SC-007 | `internal/evals/juzgar_test.go:TestLineaDeLaSalidaNoEscrita` (4666), `internal/app/tuberia_unix_test.go:TestTuberiaCerrada` (45) | sí | sí, y con el mutante de §3.7.b, `TestTuberiaCerrada` en rojo y `TestLineaDeLaSalidaNoEscrita` en verde |
| FR-020, SC-009 | `internal/evals/definicion_test.go:TestDefinicionDelJob` (196) | sí | sí |
| FR-031, SC-021 | `Makefile:skills-check` (126), `internal/app/skills_test.go:TestSkillsDelRepositorio` (83) | sí | sí (§3.6 y `make ci`) |

Filas `evals:` — no se miden aquí (necesitan el job con modelo, §3.10); lo que hay en el árbol, visto en el informe de
un plan sintético con las evals del repositorio (`TestUmbralesDeLasTresSkills`, en verde):

| Umbral | Origen | En el árbol |
|---|---|---|
| `legal-core:lectura_sin_intentar:claude-sonnet-5-5:orden` y `:herramienta` | **nuevos, de T007** (con la 11 de T008 y los motivos de T013) | `lecturasMedidas.umbrales` (`internal/evals/umbrales.go:978-985`): `Umbral: umbralDeLecturasSinIntentar` (`= 0`, `:52`), `Decide: grupo.modelo == l.modeloQueDecide` (`:984`); en el plan, total 3 (las sesiones de la 11 del modo) y `decide: true` (`informe_test.go:6898`, `:6900`, `:6904` y `:6907`); los de `claude-haiku-4-5-20251001`, publicados sin decidir |
| `legal-core:plazo_sin_fecha:claude-sonnet-5-5:<modo>` | el de hoy, que sigue, con la 11 en su total (T008) | total 12 por modo y `decide: true` (`informe_test.go:6892`, `:6895` y `:6904`) |
| `legal-core:cobertura_sin_declarar:claude-sonnet-5-5:<modo>` | el de hoy, con la 12 en su total (T008) | total 9 por modo y `decide: true` |
| `legal-core:sin_activar:claude-sonnet-5-5:<modo>` | el de hoy (T008) | total 30 por modo y `decide: true` |
| `legal-core:vigencia_sin_declarar:…` y `redaccion_sin_fechar:…` | los de hoy, sin cambiar sus totales | 3 y 9 por modo y `decide: true` |
| `fuera_de_lo_grabado:<modelo>:<modo>` en las tres skills | **nuevos, de T010** | `umbrales.go:964-970`, sin `Decide` (falso): publicados, nunca deciden (FR-021); cuatro por skill, todos `decide: false` (`informe_test.go:6905-6907` y `:6975-6979`) |
| `boe-legislacion:afirma_lo_no_leido:claude-sonnet-5-5` y los otros 10 que deciden | los de hoy, que siguen | 19 umbrales y 11 que deciden (`informe_test.go:6928`), los que deciden iguales en forma a los de antes de H26.1 (`:6975-6979`); `afirma_lo_no_leido` con la tasa `0.028` de `evals/boe-legislacion/juez/clases.yaml:9`, sin cambios, sobre 120 respuestas (`:6939`); `evals/boe-legislacion/juez/` fuera del diff (§1) |
| Los 11 que deciden de `jurisprudencia` | los de hoy, que siguen | 19 umbrales y 11 que deciden (`informe_test.go:6948`), iguales en forma a los de antes; `afirma_lo_no_leido` con la tasa `0.017` de `evals/jurisprudencia/juez/clases.yaml:9`; `evals/jurisprudencia/` y `skills/jurisprudencia/` fuera del diff (§1) |
| `duracion_de_las_sesiones:<modo>`, `duracion_del_juez:<modo>`, `medida_del_juez:…` | los de hoy | sin cambios en `internal/evals/definicion.go`, `juez.go`, `sondeo.go` ni en `.github/` (§1) |

SC-002 lo comprueba `precheck.sh plan`; SC-010 y FR-032, la revisión final y el job en la propuesta de cambio.

## 5. Para el informe final: los dos límites conocidos y SC-017

El spec («Fuera de alcance») manda que el informe final enseñe dos límites conocidos, que este hito no mide:

1. **La línea `⚠ REDACCIÓN A FECHA:` de memoria en la 11.** El juicio de una eval con `sin_servir` no da un motivo por una
   línea `⚠ REDACCIÓN A FECHA:` de un bloque que no se sirvió: una respuesta de la 11 que ponga esa línea de memoria
   pasa el juicio del job.
2. **La exposición de la norma del procedimiento tras `⚠ NO CUBIERTO:`.** No hay una clase del juez con modelo para
   `legal-core` ni una prohibición en las evals 07, 08 y 12 de leer la norma que rige el procedimiento de su materia:
   la exposición de esa norma tras `⚠ NO CUBIERTO:` (los arts. 223 y 235 de la LGT en la 07) sigue sin medirse en el
   job.

**SC-017** se confirma con lo que publique el job de cierre: en el informe de `boe-legislacion`, ninguna llamada sin sobre
entre los `1`, y los de salida no escrita con su `motivo` (quickstart §10, la quinta orden). Lo que salga confirma o
descarta la causa del ADR 0043; en `make ci` solo está fijado, sin modelo, que el juicio los reparte así
(`TestJuzgarElMotivoDeLasFallidas` y `TestLineaDeLaSalidaNoEscrita`, en verde).

## Verificación de la tarea

- `make ci` en primer plano, en verde las dos veces: la de 13:09:42 a 13:15:29 (§3.8) y la final, con este fichero ya
  escrito y T016 marcada, de 13:24:37 a 13:30:16 (`ci: todos los controles en verde`, `0 issues.`, 61 líneas `ok`,
  ninguna `FAIL`, `No vulnerabilities found.`, `no leaks found`). Sus perfiles dan las cifras de §2: 96,8 % y 97,2 % de
  total, 2 700 / 2 740 en `internal/core/**` y 4 318 / 4 414 en `internal/evals` con el de integración.
- `git status --short`: solo este `cierre.md` nuevo, la marca de T016 en `tasks.md` y los tres ficheros de `gates/` que
  el workflow tenía modificados al empezar. Ningún árbol aparte queda en `git worktree list`; los registros de la sesión
  quedaron en `/tmp/claude-501/`, fuera del repositorio.
- Esta tarea no escribe tests ni código de producto, no usa la red (salvo la de las herramientas de Go en `make vuln`),
  no graba nada, no escribe en `evidencias/` y no abre ninguna sesión con modelo.

## Anexo: `git diff --name-status main`

```text
M	CHANGELOG.md
M	CONTRIBUTING.md
M	evals/boe-legislacion/14-trlrhl-impuestos-por-materia.yaml
M	evals/boe-legislacion/15-irpf-rendimientos-por-materia.yaml
M	evals/boe-legislacion/16-lrjsp-legalidad-por-materia.yaml
A	evals/legal-core/11-plazo-reposicion-sin-leer-el-precepto.yaml
A	evals/legal-core/12-cobertura-seguridad-social-sin-leer-la-da-primera.yaml
M	go.mod
M	internal/evals/conjunto_test.go
M	internal/evals/consultas.go
M	internal/evals/consultas_test.go
M	internal/evals/definicion_test.go
M	internal/evals/doc.go
M	internal/evals/ejecucion_test.go
M	internal/evals/formato.go
M	internal/evals/formato_test.go
M	internal/evals/grabaciones_test.go
M	internal/evals/informe.go
M	internal/evals/informe_test.go
M	internal/evals/juzgar.go
M	internal/evals/juzgar_test.go
M	internal/evals/medida.go
M	internal/evals/medida_test.go
M	internal/evals/plan_test.go
M	internal/evals/preparar.go
M	internal/evals/preparar_test.go
M	internal/evals/sesion.go
M	internal/evals/sesion_test.go
M	internal/evals/sondeo_test.go
A	internal/evals/testdata/boe.legislacion-consolidada/GET_https_www.boe.es_datosabiertos_api_legislacion-consolidada_id_BOE-A-2004-4214_texto_bloque_a15.json
A	internal/evals/testdata/boe.legislacion-consolidada/GET_https_www.boe.es_datosabiertos_api_legislacion-consolidada_id_BOE-A-2006-20764_texto_bloque_a18.json
A	internal/evals/testdata/boe.legislacion-consolidada/GET_https_www.boe.es_datosabiertos_api_legislacion-consolidada_id_BOE-A-2015-10566_texto_bloque_a26.json
A	internal/evals/testdata/boe.legislacion-consolidada/GET_https_www.boe.es_datosabiertos_api_legislacion-consolidada_id_BOE-A-2015-10566_texto_bloque_a27.json
M	internal/evals/testdata/boe.legislacion-consolidada/GET_https_www.boe.es_robots.txt.json
M	internal/evals/testdata/grabaciones.json
M	internal/evals/umbrales.go
M	internal/evals/umbrales_test.go
M	schemas/eval.yaml.json
M	skills/legal-core/SKILL.md
A	specs/025-h26-1-las-evals-sirven/checklists/requirements.md
A	specs/025-h26-1-las-evals-sirven/contracts/cache-de-la-sesion.md
A	specs/025-h26-1-las-evals-sirven/contracts/formato-de-eval.md
A	specs/025-h26-1-las-evals-sirven/contracts/juicio-e-informe.md
A	specs/025-h26-1-las-evals-sirven/contracts/skill-legal-core.md
A	specs/025-h26-1-las-evals-sirven/data-model.md
A	specs/025-h26-1-las-evals-sirven/gates/analyze.md
A	specs/025-h26-1-las-evals-sirven/gates/clarify-preguntas.json
A	specs/025-h26-1-las-evals-sirven/gates/clarify-respuestas.json
A	specs/025-h26-1-las-evals-sirven/gates/clarify-rondas
A	specs/025-h26-1-las-evals-sirven/gates/cuarentena.json
A	specs/025-h26-1-las-evals-sirven/gates/grabaciones.log
A	specs/025-h26-1-las-evals-sirven/gates/grabaciones.md
A	specs/025-h26-1-las-evals-sirven/gates/modelos.jsonl
A	specs/025-h26-1-las-evals-sirven/gates/plan-correccion-r1.diff
A	specs/025-h26-1-las-evals-sirven/gates/plan-precheck.txt
A	specs/025-h26-1-las-evals-sirven/gates/plan-r1.json
A	specs/025-h26-1-las-evals-sirven/gates/plan-r2.json
A	specs/025-h26-1-las-evals-sirven/gates/plan-rondas
A	specs/025-h26-1-las-evals-sirven/gates/plan.json
A	specs/025-h26-1-las-evals-sirven/gates/spec-correccion-r1.diff
A	specs/025-h26-1-las-evals-sirven/gates/spec-correccion-r2.diff
A	specs/025-h26-1-las-evals-sirven/gates/spec-precheck.txt
A	specs/025-h26-1-las-evals-sirven/gates/spec-r1.json
A	specs/025-h26-1-las-evals-sirven/gates/spec-r2.json
A	specs/025-h26-1-las-evals-sirven/gates/spec-r3.json
A	specs/025-h26-1-las-evals-sirven/gates/spec-rondas
A	specs/025-h26-1-las-evals-sirven/gates/spec.json
A	specs/025-h26-1-las-evals-sirven/gates/supuestos.md
A	specs/025-h26-1-las-evals-sirven/gates/tarea-T001.md
A	specs/025-h26-1-las-evals-sirven/gates/tarea-actual.json
A	specs/025-h26-1-las-evals-sirven/gates/tareas-intentos.json
A	specs/025-h26-1-las-evals-sirven/gates/tasks-precheck.txt
A	specs/025-h26-1-las-evals-sirven/gates/tasks-r1.json
A	specs/025-h26-1-las-evals-sirven/gates/tasks-rondas
A	specs/025-h26-1-las-evals-sirven/gates/tasks.json
A	specs/025-h26-1-las-evals-sirven/mutantes.md
A	specs/025-h26-1-las-evals-sirven/plan.md
A	specs/025-h26-1-las-evals-sirven/quickstart.md
A	specs/025-h26-1-las-evals-sirven/research.md
A	specs/025-h26-1-las-evals-sirven/spec.md
A	specs/025-h26-1-las-evals-sirven/tasks.md
```
