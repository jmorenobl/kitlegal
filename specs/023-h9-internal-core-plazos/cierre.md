# Cierre de la Definition of Done · H9 (T016)

Sesión del 2026-10-18 sobre la cabeza `17610eb` (T014) de `023-h9-internal-core-plazos`, con `main` en `63e9ae9`,
que es también la base de la rama: `git diff --name-status main` es exactamente lo que traen los 20 commits de la
rama. Todo lo que se afirma aquí se ejecutó y se vio terminar en esta sesión, o está en el repositorio; lo que no se
pudo medir se dice tal cual. T015 (la documentación) está en cuarentena: este cierre no lo cubre ni lo sustituye.

Cómo se ejecutó el quickstart (desviaciones respecto a su §0, todas de entorno y ninguna de producto):

- El sandbox de la sesión rechaza en la línea de órdenes las rutas fuera del repositorio, las asignaciones de
  entorno (`env HOME=…`, `HOME=… orden`), los grupos `{ …; }` y `$?`, y no deja ejecutar un guion ni el binario
  construido. Por eso `$T` fue `bin/t016/` —dentro del repositorio pero ignorado por git (`.gitignore:2:/bin/`),
  comprobado con `git check-ignore`—, borrado al terminar; cada orden del binario se lanzó con `go run ./cmd/kitlegal`
  en vez de con el binario construido; y el código de salida de §6 se leyó de la línea `exit status N` que `go run`
  escribe en stderr.
- §2 se ejercitó con copias `zz-h9-*.txtar` de los cinco guiones de `aceptacion/`, que se borraron (quedan 0).
- §7 se ejercitó con el arnés e2e (una copia momentánea `zz-h9-quickstart-7.txtar`, borrada), porque desde la línea de
  órdenes no se pueden encadenar los mensajes con pausas; el detalle y una observación, en §3.7.
- §8 no pudo aislar `HOME` ni `KITLEGAL_CACHE_DIR` en línea; su evidencia es la del guion `plazos-sin-efectos`, que hace
  exactamente esa comprobación, y la de `TestAppletPlazos` y `TestArquitectura` (§3.8).

## 1. Lo creado o modificado, y lo que no cambia

`git diff --name-status main`: 198 entradas, 150 `A`, 48 `M`, 0 `D` (la lista entera, en el anexo del final). Por
áreas:

| Área | Entradas | Qué son |
|---|---|---|
| `internal/core/plazos/` | 16 A | el dominio de los plazos y sus tests |
| `internal/app/` | 3 A + 4 M de código y tests, 4 A de golden, 4 M de guiones e2e, 1 M del binario de ejemplo + su test | el applet `plazos`, el registro, los guiones de H21 con la herramienta nueva |
| `internal/source/bocm/`, `internal/source/boe/fiestas/` | 10 A + 7 A | los dos adaptadores de festivos con sus manifiestos |
| `internal/evals/` | 2 A + 16 M + 21 A de `testdata/` | el formato `plazo`/`no_cubierto`, el juicio sin modelo, los umbrales, la tercera raíz de grabaciones |
| `internal/skills/`, `internal/core/territorio/` | 2 A + 2 M, 4 M | `LeerCobertura`, el generador de referencias, `MunicipiosDe` |
| `data/`, `schemas/` | 4 A + 2 M, 2 A + 1 M | `plazos.yaml`, `cobertura.yaml`, los dos ficheros de `data/festivos/`, los esquemas |
| `skills/legal-core/` | 1 M + 2 A | `SKILL.md` v1 y sus dos referencias generadas |
| `evals/legal-core/` | 5 A | las evals `05` a `09` |
| `evidencias/h9/` | 13 A | las grabaciones y su manifiesto (las graba `grabar_datos`, sin modelo) |
| `specs/023-h9-internal-core-plazos/` | 57 A | los artefactos del hito (38 en `gates/`) |
| Raíz y workflow | `.golangci.yml`, `CHANGELOG.md`, `CLAUDE.md`, `CONTRIBUTING.md`, `cmd/kitlegal/main_test.go`, `internal/arch_test.go`, `internal/workflow/` (1 A + 1 M), `.specify/workflows/hito/workflow.yml`, `docs/HISTORIAL.md`, `docs/WORKFLOW.md`, `scripts/claude-modelo.sh`, `scripts/workflow/marcas.sh` (A), `scripts/workflow/verificar.sh` | ver abajo |

**Lo que no cambia** (FR-028, FR-058, FR-063, FR-070), comprobado contra esa lista:

| Elemento | En el diff |
|---|---|
| `.github/workflows/` | no |
| `skills/boe-legislacion/`, `skills/jurisprudencia/`, `evals/boe-legislacion/`, `evals/jurisprudencia/` | no |
| `internal/mcp`, `internal/empaquetado` | no |
| `docs/SOURCES.md`, `scripts/verify-sources.sh` | no |
| `docs/ROADMAP.md`, `docs/ADR/`, `docs/USO.md` | no |
| `web/` | no |
| Las cuatro evals de hoy de `legal-core` (`01` a `04`) | no: solo `A` de `05` a `09` |
| Los manifiestos de H4 (`internal/source/boe/testdata/grabaciones.json`) y de H5 (`testdata/evals/grabaciones.json`) | no: los tres manifiestos nuevos son los de `bocm`, `boe/fiestas` e `internal/evals/testdata/` |
| Los specs de otros hitos | no: todo lo de `specs/` está bajo `023-h9-internal-core-plazos/` |
| `CLAUDE.md`, `docs/HISTORIAL.md` | **sí, una línea cada uno**, y no por una tarea del hito: ver abajo |

**Para la revisión final.** `CLAUDE.md` y `docs/HISTORIAL.md` difieren de `main` solo en «El workflow `hito` va por la
2.9.0» → «2.9.1». Lo hace el commit `f266c12`, «fix(workflow): las sesiones de un paso retiran las marcas de
escritura vacías de Claude Code (workflow 2.9.1)», anterior a T001 y ajeno a las tareas (ninguna declara esas rutas;
`git log main..HEAD -- CLAUDE.md docs/HISTORIAL.md` da solo ese commit), que trae también `.specify/workflows/hito/
workflow.yml`, `docs/WORKFLOW.md`, `scripts/claude-modelo.sh`, `scripts/workflow/marcas.sh`, `scripts/workflow/
verificar.sh`, `internal/workflow/marcas_test.go`, `internal/workflow/arbol_test.go` y seis líneas de `CHANGELOG.md`.
Es el arreglo del workflow con el que se reanudó el run, no contenido de H9; esta tarea no lo revierte (dejaría al run
sin su arreglo) y lo deja anotado para que el informe final lo diga. `.golangci.yml` y `CONTRIBUTING.md` los tocan
T004, T008, T009 y T010 (rutas suyas; `CONTRIBUTING.md` siempre está permitido).

**`data/festivos/` es lo que escriben sus tests de derivación** (sin `-actualizar-festivos`):
`TestFestivosNacionalesDerivados` (`internal/source/boe/fiestas/`) y `TestFestivosLocalesDerivados`
(`internal/source/bocm/`) en verde en esta sesión (§3.9), y `git status --porcelain -- data/festivos/` vacío después
de ejecutarlos.

**Ningún municipio en código, `data/plazos.yaml`, `data/cobertura.yaml` ni `SKILL.md`** (FR-073):

- Los nombres que usan los fixtures, las evals y la tabla de casos (Leganés, Tordesillas, Majadahonda, Brunete,
  Arroyomolinos) y sus códigos INE entre comillas (`28074`, `47165`, `28080`, `28026`, `28079`): 0 apariciones en
  el Go que no es de test de todo el módulo (`cmd/`, `internal/`, `data/`), en `data/plazos.yaml`, en
  `data/cobertura.yaml` y en `skills/legal-core/SKILL.md`.
- Los 8 245 nombres distintos de `data/territorio/municipios.yaml`, como palabra entera: 0 apariciones en
  `data/plazos.yaml`, `data/cobertura.yaml` y `SKILL.md`; en los 34 ficheros Go de producto que toca el hito, las
  únicas apariciones son comentarios y palabras corrientes que coinciden con un nombre —`avisos.go:258` «Mira»;
  `internal/app/plazos.go:40` el tipo `territorio.Fuentes`; `bocm/derivar.go:303` `slices.Sort` («Sort», Lleida);
  `bocm/derivar.go:24` y `bocm/doc.go:2-4` «Comunidad de Madrid»; `bocm/cruce.go:16` el ejemplo «Braojos de la
  Sierra» en el comentario de la regla genérica de los nombres; `boe/fiestas/doc.go:4` «Real Decreto»;
  `boe/fiestas/derivar.go:245-265` los nombres de comunidades (Murcia, Madrid, Ceuta, Asturias) en el comentario de la
  normalización—; ningún caso especial en código.
- Códigos de cinco cifras: ninguno entre comillas en esos ficheros Go; ninguno en los tres ficheros de datos y skill
  fuera de los identificadores `BOE-A-…` y `BOCM-…`.

## 2. Cobertura (Definition of Done §1.9)

Sobre los perfiles que dejó el `make ci` de §3.1 (`coverage.out` y `coverage-integration.out`, 10 808 líneas cada
uno), con `go tool cover -func`; la cifra de `internal/core/plazos` sale del subperfil de sus 483 líneas con la misma
orden.

| Perfil | Global (`total:`) | ≥ 70 % | `internal/core/plazos` | ≥ 85 % |
|---|---|---|---|---|
| `coverage.out` | 96,7 % | sí | 99,3 % | sí |
| `coverage-integration.out` | 97,1 % | sí | 99,3 % | sí |

Las dos quedan por encima: no hay hueco que pasar a la revisión final. Para contexto, `internal/core/**` por paquete
en la salida de `make ci` (`test`): `cita` 100,0, `grafo` 92,5, `ids` 100,0, `instalacion` 99,7, `plazos` 99,3,
`schema` 89,6, `territorio` 100,0 (todos ≥ 85); `go test -count=1 -cover ./internal/core/plazos/` (§3.11) da 99,3 %.
Las dos funciones de `internal/core/plazos` por debajo del 100 % en los dos perfiles son `calendario.go:408`
`PartesDe` (85,7 %) y `festivos.go:187` `escribir` (81,2 %); el paquete cumple con holgura y esta tarea no escribe
tests (lo hizo cumplir T005). Los únicos paquetes por debajo del 70 % en el listado de `make ci` son `cmd/empaquetar`,
`cmd/kitlegal` e `internal/app/ejemplo` (0,0 %, anteriores a H9, fuera de `internal/core`); el total global ya los
incluye.

## 3. Quickstart §0 a §13 y §16, orden a orden

### 3.0 Antes de empezar

`git status --porcelain` inicial: las cinco entradas `M` de `gates/` que el workflow deja al lanzar la tarea
(`cuarentena.json`, `modelos.jsonl`, `supuestos.md`, `tarea-actual.json`, `tareas-intentos.json`) y `??
gates/cuarentena/T015.patch`. Binario: `go run ./cmd/kitlegal version` responde `kitlegal dev`. `$T`: `bin/t016/`,
ignorado por git.

### 3.1 `make ci` en verde (FR-080)

`make ci` termina con `ci: todos los controles en verde` y `RESULTADO: make ci en verde`. Pasos vistos: `fmt
--diff` sin salida; `lint` «0 issues»; `test` con `-race` y perfil, todos los paquetes `ok`; `test-integration`,
todos `ok`; `test-tiempos` `ok`; `govulncheck` «No vulnerabilities found» (0 en el código y en lo importado; 2 en
módulos requeridos que el código no llama); `schema-check` (`TestEsquemasPublicados` `ok`, con `schemas/plazo.json` y
`schemas/cobertura.yaml.json`); `skills-check` (los siete tests de repositorio `ok`); `goreleaser check` «1
configuration file(s) validated»; `gitleaks` «no leaks found»; `go mod verify` «all modules verified» en el módulo y
en las cinco herramientas; `go mod tidy -diff` sin salida.

### 3.2 Los guiones de aceptación (SC-009)

`go test -race -count=1 -run '^TestEntregaDelHito$' -v ./internal/app/` con las cinco copias `zz-h9-plazos-*.txtar`:

```
--- PASS: TestEntregaDelHito/zz-h9-plazos-sin-efectos (0.20s)
--- PASS: TestEntregaDelHito/zz-h9-plazos-cobertura (0.12s)
--- PASS: TestEntregaDelHito/zz-h9-plazos-calcular (0.16s)
--- PASS: TestEntregaDelHito/zz-h9-plazos-herramienta (0.37s)
--- PASS: TestEntregaDelHito/zz-h9-plazos-codigos (0.67s)
ok  	github.com/jmorenobl/kitlegal/internal/app	17.074s
```

Cinco `PASS`, como esperaba §2 (con el prefijo `zz-` de las copias). Copias borradas: quedan 0. Activarlos como
`h9-plazos-*.txtar` lo hace el workflow tras el bucle.

### 3.3 El caso (a) (US1)

`plazos calcular reposicion 2026-09-09 Leganés Leganés --json`: código 0 y

```
{"fecha":"2026-10-13","preceptos":[{"norma":"BOE-A-2015-10565","bloque":"a124","apartado":"124.1"}],"municipios":[{"codigo_ine":"28074","nombre":"Leganés"}],"festivos":[{"fecha":"2026-10-09","ambito":"local","municipio":"28074","source":"BOCM-20251212-34"},{"fecha":"2026-10-12","ambito":"nacional","source":"BOE-A-2025-21667"}],"avisos":[]}
```

Con `--offline`: código 0 y `diff` de `{data, hash}` vacío → `iguales` (FR-013, SC-009). Todo como lo esperado.

### 3.4 Tordesillas y 2027 (US2; FR-026)

- Tordesillas 2026-09-09: `2026-10-09`, un aviso `faltan-festivos-locales` cuyo texto empieza por `⚠ FALTAN FESTIVOS
  LOCALES:` y nombra Tordesillas (47165) y 2026; `cobertura` con
  `{"ambito":"local","municipio":"47165","desde":2026,"hasta":2026,"falta":"sin-festivos-locales"}`.
- Leganés 2026-12-14: `2027-01-14` con `["faltan-festivos-locales"]`.
- `generico 2024-12-30 … --duracion 3 --unidad dias-habiles`: `2025-01-03` con
  `["faltan-festivos-nacionales-y-autonomicos","faltan-festivos-locales"]`.

Los tres como lo esperado (casos 2, 18 y 17).

### 3.5 El contencioso (US5; FR-010)

- `contencioso 2026-06-15 Madrid`: `2026-09-15`, preceptos `a46` (46.1) y `a128` (128.2) de `BOE-A-1998-16718`,
  `avisos` `[]`.
- `contencioso 2026-10-27 Madrid`: `2026-12-28` con `["dias-del-art-183-lopj"]`.

Como lo esperado (casos 3 y 15).

### 3.6 Los códigos de salida y su orden (FR-012)

| Orden | Código | `data.clase` |
|---|---|---|
| `inventado 2026-09-09 Leganés Leganés` | 2 | `argumentos` |
| `reposicion 2026-02-30 Leganés Leganés` | 2 | `argumentos` |
| `generico 2026-09-09 Leganés Leganés` | 2 | `argumentos` |
| `reposicion 2026-09-09 Leganés Leganés --duracion 2` | 2 | `argumentos` |
| `silencio 2026-09-09 Leganés Leganés --duracion 3` | 2 | `argumentos` |
| `reposicion 2026-09-09 Leganés` | 2 | `argumentos` |
| `contencioso 2026-09-09 Madrid Leganés` | 2 | `argumentos` |
| `reposicion 2026-09-09 Arroyomolinos Leganés` | 2 | `argumentos` |
| `reposicion 2026-09-09 Inexistente Leganés` | 3 | `no-encontrado` |
| `inventado 2026-09-09 Inexistente Leganés` | 2 | `argumentos` |

El `data.mensaje` de Arroyomolinos: «municipio «Arroyomolinos»: el nombre "Arroyomolinos" es el de 2 municipios de la
relación; consulta uno por su código INE: 10023 Arroyomolinos (Cáceres); 28015 Arroyomolinos (Madrid)». Todo como lo
esperado: el tipo va antes que el municipio. El código lo da la línea `exit status N` de `go run`; el guion
`plazos-codigos` (§3.2) lo afirma además con `test $? -eq 2` y `-eq 3` dentro de `sh`.

### 3.7 La herramienta (FR-001, FR-085)

Con el arnés (copia momentánea `zz-h9-quickstart-7.txtar`, `PASS` en 0,67 s, borrada): `mcp anuncio.txt
$KITLEGAL_T0_BIN mcp serve` da `1 herramientas 16` y `salida 0`, con `plazos_calcular lectura=true` entre las 16 del
binario del arnés (las 13 de producción más `contar_letras`, `contar_palabras` y `echo_repetir`, los applets de
ejemplo del arnés; las 13 de producción las fija `TestHerramientasDelServidor`, en verde en esta sesión, §4); las dos
llamadas en un mismo proceso dan `1 plazos_calcular resultado`, `2 plazos_calcular error argumentos` y `salida 0`, con
`"data":{"fecha":"2026-10-13",` en `mcp-1.json` y `"data":{"clase":"argumentos",` en `mcp-2.json`. Lo esperado por §7
(13 herramientas, `2026-10-13`, el error `argumentos` y el servidor que sigue) queda cubierto entre el arnés y ese
test.

Observación, fuera de H9 (`internal/mcp` no cambia), para la revisión final: antes del arnés se probó `cat` de los
cinco mensajes de §7 de golpe en `go run ./cmd/kitlegal mcp serve`, es decir, con la entrada cerrada nada más
escribirlos; el servidor salió con 0 sin escribir ninguna respuesta ni nada en stderr. Con pausas entre mensajes (lo
que hacen `hablar` y el arnés) responde a todo. No se ha indagado más.

### 3.8 Sin red, sin caché y sin grafo (FR-013, FR-028; SC-009)

`go test -count=1 -run '^TestArquitectura$' ./internal/`: `ok` (8 pasados). La comprobación de que ni `$T/home` ni
`$T/cache` ganan un fichero no se pudo hacer en línea (el sandbox rechaza `HOME=…`/`KITLEGAL_CACHE_DIR=…` delante de la
orden); la hace el guion `plazos-sin-efectos` (§3.2, `PASS`): con la caché y el grafo con contenido, `arbol cache`
antes y después de `plazos calcular` como orden —una que termina con 0, una que falla, una con `--offline --no-graph`,
una con `--dry-run`— y como herramienta, `cmp` iguales; y `TestAppletPlazos` (§3.11, `PASS`): 0 operaciones de grafo.

### 3.9 `data/festivos/` sale de lo grabado (US3)

`TestLeerResolucion`, `TestFestivosNacionalesDerivados` (`boe/fiestas`, `ok` 0,380 s); `TestLeerDisposicion`,
`TestCruceConElTerritorio`, `TestCotejoConElCatalogo`, `TestFestivosLocalesDerivados` (`bocm`, `ok` 0,486 s): todos
`--- PASS`. Majadahonda (`28080`) en 2025: `2025-09-15` y `2025-11-25`, con `BOCM-20250203-16`; Brunete (`28026`) en
2026: `2026-05-15` y `2026-09-14`, con `BOCM-20251229-19`. Como lo esperado (US3.2, US3.3).

### 3.10 Los preceptos, contra lo grabado (FR-010, FR-030, FR-031)

`TestGrabacionesSinSolape`, `TestManifiestoDeGrabaciones`, `TestIdentificadoresDeLasNormas`,
`TestPreceptosDeLosPlazos`: todos `--- PASS` (`internal/evals`, `ok` 0,648 s).

### 3.11 La tabla de casos y las propiedades (SC-004, SC-005)

`TestFestivoDeUnaParte`, `TestAppletPlazos`, `TestCasosDePlazos` (`internal/app`, `ok` 1,889 s) y `TestCalcular`,
`TestPropiedadesDelComputo` (`internal/core/plazos`, `ok` 1,344 s): todos `--- PASS`. `go test -count=1 -cover
./internal/core/plazos/`: `coverage: 99.3% of statements` (≥ 85 %). (El recuento de subtests, uno por caso, lo fija
`TestCasosDePlazos` sobre los 20 casos de contracts/casos-de-plazos.md; aquí se vio el `PASS` del test.)

### 3.12 Las evals y los umbrales con sesiones sintéticas (SC-011, SC-012)

`TestJuzgarPlazos`, `TestDefinicionDelJob`, `TestFormatoDePlazos`, `TestUmbralesDeLasTresSkills`,
`TestUmbralesDeLegalCore`, `TestEvalsDelRepositorio`: todos `--- PASS` (`ok` 1,386 s). `ls evals/legal-core/`: nueve
evals, `01` a `04` y `05-plazo-reposicion-leganes`, `06-plazo-reposicion-tordesillas`,
`07-cobertura-liquidacion-tributaria`, `08-cobertura-multa-de-trafico`, `09-plazo-contencioso-agosto`.

### 3.13 La skill (SC-010)

`make skills-check` en verde (los siete tests `ok`); `wc -l < skills/legal-core/SKILL.md` → 248 (< 300); la primera
línea de `references/recursos_y_plazos.md` es `<!-- generado desde data/plazos.yaml, no editar -->` y la de
`references/cobertura_por_materia.md`, `<!-- generado desde data/cobertura.yaml, no editar -->`; `grep -c '^⚠ '` → 7.
Como lo esperado.

### 3.14 y 3.15

§14, el cierre (el job de evals sobre la propuesta de cambio y el `informe.json` de `legal-core` con sus seis
umbrales), lo hace el workflow tras la revisión final; no es trabajo de esta tarea ni de ninguna. §15, la persona en la
app de escritorio con la release, queda para después del run.

### 3.16 Limpieza

`bin/t016/` borrado al cerrar esta tarea; `git status --porcelain` al final del quickstart igual al del principio
(`diff` vacío → «el árbol está como al principio»): las copias `zz-` no quedan y `coverage.out` y
`coverage-integration.out` los ignora `.gitignore`.

## 4. Controles de umbral (plan.md)

Cada fila `ci:` con su test encontrado como `^func <Test>(` en su ruta (`grep -c` = 1) y en verde en esta sesión:

| Fila (requisito) | Control | Encontrado | Verde |
|---|---|---|---|
| FR-050, SC-001 | `internal/evals/definicion_test.go:TestDefinicionDelJob` | 1 | §3.12 |
| FR-055, FR-087, SC-012 | `internal/evals/informe_test.go:TestUmbralesDeLegalCore` | 1 | §3.12 |
| FR-056, SC-011 | `internal/evals/informe_test.go:TestUmbralesDeLasTresSkills` | 1 | §3.12 |
| FR-081, SC-004 | `internal/core/plazos/propiedades_test.go:TestPropiedadesDelComputo` | 1 | §3.11 |
| FR-082, SC-005 | `internal/app/plazos_casos_test.go:TestCasosDePlazos` | 1 | §3.11 |
| FR-011 | `internal/app/plazos_test.go:TestFestivoDeUnaParte` | 1 | §3.11 |
| FR-020, FR-083, SC-006 | `internal/source/boe/fiestas/derivar_test.go:TestFestivosNacionalesDerivados` | 1 | §3.9 |
| FR-021, FR-022, FR-083, SC-006 | `internal/source/bocm/derivar_test.go:TestFestivosLocalesDerivados` | 1 | §3.9 |
| FR-084, SC-007 | `internal/source/bocm/cotejo_test.go:TestCotejoConElCatalogo` | 1 | §3.9 |
| FR-030, FR-031, FR-010 | `internal/evals/grabaciones_test.go:TestPreceptosDeLosPlazos` | 1 | §3.10 |
| FR-052 | `internal/evals/conjunto_test.go:TestEvalsDelRepositorio` | 1 | §3.12 |
| FR-051 | `…conjunto_test.go:TestEvalsDelRepositorio`, `internal/evals/formato_test.go:TestFormatoDePlazos` | 1, 1 | §3.12 |
| FR-013, SC-009 (red) | `internal/arch_test.go:TestArquitectura` | 1 | §3.8 |
| FR-013, SC-009 (grafo, caché, `--offline`) | `internal/app/plazos_test.go:TestAppletPlazos`; guiones `h9-plazos-sin-efectos`, `h9-plazos-calcular` | 1; guiones en `aceptacion/` | §3.11; §3.2 (copias `zz-`) |
| FR-012, FR-085 | guiones `h9-plazos-codigos`, `h9-plazos-cobertura` | en `aceptacion/` | §3.2 |
| FR-001, FR-085 (herramienta) | `internal/app/herramientas_test.go:TestHerramientasDelServidor`; guion `h9-plazos-herramienta` | 1; en `aceptacion/` | `--- PASS` en esta sesión; §3.2 |
| FR-046, SC-010 | `Makefile:skills-check` (objetivo en la línea 126); `internal/app/skills_test.go:TestSkillsDelRepositorio` | sí; 1 | §3.13; `--- PASS` en esta sesión |
| FR-032, SC-010 | `Makefile:skills-check` | sí | §3.13 |
| FR-089, SC-008 | `TestCasosDePlazos`, `TestFestivosLocalesDerivados`, `TestUmbralesDeLegalCore`, `TestUmbralesDeLasTresSkills` | 1 cada uno | §3.9, §3.11, §3.12 (el registro de los mutantes, en `mutantes.md`, T014) |

Los guiones `h9-plazos-*.txtar` de las filas `ci:internal/app/testdata/script/…` están hoy en `aceptacion/` (suite
congelada); el workflow los activa en `testdata/script/` tras el bucle. En esta sesión pasan como copias `zz-` (§3.2).

Cada fila `evals:` con el umbral que publica T012 (`internal/evals/umbrales.go`, comprobado por
`TestUmbralesDeLasTresSkills`):

| Fila | Umbral publicado |
|---|---|
| `evals:legal-core:plazo_sin_fecha:claude-sonnet-5-5:orden` y `…:herramienta` | «Respuestas de %s en el modo %s sin la línea ⚠ PLAZO: con la fecha esperada, sobre sus respuestas medidas en las evals que esperan un plazo»; `comparacion` `<=`, `umbral` 0, `decide` `true`; total 9 por modo |
| `evals:legal-core:cobertura_sin_declarar:claude-sonnet-5-5:orden` y `…:herramienta` | «Respuestas de %s en el modo %s sin la línea ⚠ NO CUBIERTO: con la norma esperada, o con una línea ⚠ PLAZO:, sobre sus respuestas medidas en las evals que esperan una declaración de no cubierto»; `<=`, 0, `decide` `true`; total 6 por modo |
| `evals:legal-core:sin_activar:claude-sonnet-5-5:orden` y `…:herramienta` | «Respuestas de %s en el modo %s sin la skill activada, sobre sus respuestas medidas en las evals que la activan»; `<=`, 0, `decide` `true`; total 21 por modo |
| `evals:boe-legislacion:…` (nueve que deciden) | los de hoy, sin cambio: 13 umbrales con 9 que deciden, en su orden (`umbralesDeBoeLegislacionComoEnH26`) |
| `evals:jurisprudencia:…` (once que deciden) | los de hoy, sin cambio: 15 umbrales con 11 que deciden, en su orden (`umbralesDeJurisprudencia`) |

El informe de `legal-core` lleva exactamente los seis, en el orden `sin_activar`, `plazo_sin_fecha`,
`cobertura_sin_declarar` del modo orden y los mismos del modo herramienta, con totales 21, 9, 6, 21, 9, 6 y veredicto
`aprobado` con las sesiones sintéticas. Lo que midan en el job es §14.

## 5. Nota para el informe final (FR-091)

Con H9 hay cinco applets de producto: `boe`, `territorio`, `graph`, `cita` y `plazos` (la ayuda del binario los lista
junto a `mcp` y `skills`, los del kernel y la distribución). Es el momento de revisar si `internal/cli` pide refactor,
no extracción (CLAUDE.md: no se extrae a librería hasta que tres applets repitan el patrón). Es una nota, no trabajo
de ninguna tarea del hito.

## Verificación de la tarea

- `make ci` en verde al principio de la sesión (§3.1) y de nuevo al final, tras escribir este fichero: `ci: todos los
  controles en verde` las dos veces (el del workflow queda en `gates/ci.log`).
- `git status`: fuera de `cierre.md`, solo lo que el workflow dejó antes de la sesión (§3.0); ninguna copia `zz-`,
  ningún guion momentáneo, nada en `testdata/`, `schemas/` ni `evidencias/`.
- Sin supuestos nuevos: las desviaciones del quickstart son de entorno y están arriba; ninguna cambia lo que ve quien
  usa el binario, el alcance ni la skill.

## Anexo: `git diff --name-status main`

    M	.golangci.yml
    M	.specify/workflows/hito/workflow.yml
    M	CHANGELOG.md
    M	CLAUDE.md
    M	CONTRIBUTING.md
    M	cmd/kitlegal/main_test.go
    A	data/cobertura.yaml
    M	data/datos.go
    M	data/datos_test.go
    A	data/festivos/locales/13.yaml
    A	data/festivos/nacionales-y-autonomicos.yaml
    A	data/plazos.yaml
    M	docs/HISTORIAL.md
    M	docs/WORKFLOW.md
    A	evals/legal-core/05-plazo-reposicion-leganes.yaml
    A	evals/legal-core/06-plazo-reposicion-tordesillas.yaml
    A	evals/legal-core/07-cobertura-liquidacion-tributaria.yaml
    A	evals/legal-core/08-cobertura-multa-de-trafico.yaml
    A	evals/legal-core/09-plazo-contencioso-agosto.yaml
    A	evidencias/h9/bocm.fiestas-locales/GET_https_www.bocm.es_boletin_CM_Orden_BOCM_2024_12_13_BOCM-20241213-23.xml.json
    A	evidencias/h9/bocm.fiestas-locales/GET_https_www.bocm.es_boletin_CM_Orden_BOCM_2024_12_27_BOCM-20241227-32.xml.json
    A	evidencias/h9/bocm.fiestas-locales/GET_https_www.bocm.es_boletin_CM_Orden_BOCM_2025_02_03_BOCM-20250203-16.xml.json
    A	evidencias/h9/bocm.fiestas-locales/GET_https_www.bocm.es_boletin_CM_Orden_BOCM_2025_02_17_BOCM-20250217-24.xml.json
    A	evidencias/h9/bocm.fiestas-locales/GET_https_www.bocm.es_boletin_CM_Orden_BOCM_2025_12_12_BOCM-20251212-34.xml.json
    A	evidencias/h9/bocm.fiestas-locales/GET_https_www.bocm.es_boletin_CM_Orden_BOCM_2025_12_29_BOCM-20251229-19.xml.json
    A	evidencias/h9/bocm.fiestas-locales/GET_https_www.bocm.es_boletin_CM_Orden_BOCM_2026_01_08_BOCM-20260108-18.xml.json
    A	evidencias/h9/bocm.fiestas-locales/GET_https_www.bocm.es_robots.txt.json
    A	evidencias/h9/boe.fiestas-laborales/GET_https_www.boe.es_diario_boe_txt.php_q_id_BOE-A-2024-21316.json
    A	evidencias/h9/boe.fiestas-laborales/GET_https_www.boe.es_diario_boe_txt.php_q_id_BOE-A-2025-21667.json
    A	evidencias/h9/boe.fiestas-laborales/GET_https_www.boe.es_diario_boe_txt.php_q_id_BOE-A-2026-21254.json
    A	evidencias/h9/boe.fiestas-laborales/GET_https_www.boe.es_robots.txt.json
    A	evidencias/h9/manifiesto.json
    M	internal/app/ejemplo/kitlegal-e2e/main.go
    M	internal/app/ejemplo/kitlegal-e2e/main_test.go
    M	internal/app/esquemas_test.go
    M	internal/app/herramientas_test.go
    A	internal/app/plazos.go
    A	internal/app/plazos_casos_test.go
    A	internal/app/plazos_test.go
    M	internal/app/registro.go
    M	internal/app/registro_test.go
    A	internal/app/testdata/plazos/caso-a.json
    A	internal/app/testdata/plazos/caso-b.json
    A	internal/app/testdata/plazos/caso-e.json
    A	internal/app/testdata/plazos/describe.json
    M	internal/app/testdata/script/argumentos.txtar
    M	internal/app/testdata/script/h21-mcp-herramientas.txtar
    M	internal/app/testdata/script/h21-mcp-proceso.txtar
    M	internal/app/testdata/script/h21-mcp-protocolo.txtar
    M	internal/arch_test.go
    A	internal/core/plazos/avisos.go
    A	internal/core/plazos/avisos_test.go
    A	internal/core/plazos/calcular.go
    A	internal/core/plazos/calcular_test.go
    A	internal/core/plazos/calendario.go
    A	internal/core/plazos/calendario_test.go
    A	internal/core/plazos/computo.go
    A	internal/core/plazos/doc.go
    A	internal/core/plazos/errores.go
    A	internal/core/plazos/festivos.go
    A	internal/core/plazos/festivos_test.go
    A	internal/core/plazos/propiedades_test.go
    A	internal/core/plazos/salida.go
    A	internal/core/plazos/salida_test.go
    A	internal/core/plazos/tipos.go
    A	internal/core/plazos/tipos_test.go
    M	internal/core/territorio/nombres.go
    M	internal/core/territorio/nombres_test.go
    M	internal/core/territorio/registro.go
    M	internal/core/territorio/registro_test.go
    M	internal/evals/conjunto.go
    M	internal/evals/conjunto_test.go
    M	internal/evals/consultas.go
    M	internal/evals/consultas_test.go
    M	internal/evals/definicion_test.go
    M	internal/evals/formato.go
    M	internal/evals/formato_test.go
    M	internal/evals/grabacion_test.go
    M	internal/evals/grabaciones.go
    M	internal/evals/grabaciones_test.go
    M	internal/evals/informe.go
    M	internal/evals/informe_test.go
    M	internal/evals/juzgar.go
    M	internal/evals/plan_test.go
    A	internal/evals/plazos.go
    A	internal/evals/plazos_test.go
    A	internal/evals/testdata/boe.legislacion-consolidada/GET_https_www.boe.es_datosabiertos_api_legislacion-consolidada_id_BOE-A-1889-4763_texto_bloque_art5.json
    A	internal/evals/testdata/boe.legislacion-consolidada/GET_https_www.boe.es_datosabiertos_api_legislacion-consolidada_id_BOE-A-1985-12666_texto_bloque_acientoochentaycinco.json
    A	internal/evals/testdata/boe.legislacion-consolidada/GET_https_www.boe.es_datosabiertos_api_legislacion-consolidada_id_BOE-A-1985-12666_texto_bloque_acientoochentaydos.json
    A	internal/evals/testdata/boe.legislacion-consolidada/GET_https_www.boe.es_datosabiertos_api_legislacion-consolidada_id_BOE-A-1985-12666_texto_bloque_acientoochentaytres.json
    A	internal/evals/testdata/boe.legislacion-consolidada/GET_https_www.boe.es_datosabiertos_api_legislacion-consolidada_id_BOE-A-1998-16718_texto_bloque_a128.json
    A	internal/evals/testdata/boe.legislacion-consolidada/GET_https_www.boe.es_datosabiertos_api_legislacion-consolidada_id_BOE-A-1998-16718_texto_bloque_a46.json
    A	internal/evals/testdata/boe.legislacion-consolidada/GET_https_www.boe.es_datosabiertos_api_legislacion-consolidada_id_BOE-A-2003-23186_texto_bloque_a97.json
    A	internal/evals/testdata/boe.legislacion-consolidada/GET_https_www.boe.es_datosabiertos_api_legislacion-consolidada_id_BOE-A-2013-12887_texto_bloque_a24.json
    A	internal/evals/testdata/boe.legislacion-consolidada/GET_https_www.boe.es_datosabiertos_api_legislacion-consolidada_id_BOE-A-2015-10565_texto_bloque_a112.json
    A	internal/evals/testdata/boe.legislacion-consolidada/GET_https_www.boe.es_datosabiertos_api_legislacion-consolidada_id_BOE-A-2015-10565_texto_bloque_a122.json
    A	internal/evals/testdata/boe.legislacion-consolidada/GET_https_www.boe.es_datosabiertos_api_legislacion-consolidada_id_BOE-A-2015-10565_texto_bloque_a124.json
    A	internal/evals/testdata/boe.legislacion-consolidada/GET_https_www.boe.es_datosabiertos_api_legislacion-consolidada_id_BOE-A-2015-10565_texto_bloque_a24.json
    A	internal/evals/testdata/boe.legislacion-consolidada/GET_https_www.boe.es_datosabiertos_api_legislacion-consolidada_id_BOE-A-2015-10565_texto_bloque_a30.json
    A	internal/evals/testdata/boe.legislacion-consolidada/GET_https_www.boe.es_datosabiertos_api_legislacion-consolidada_id_BOE-A-2015-10565_texto_bloque_daprimera.json
    A	internal/evals/testdata/boe.legislacion-consolidada/GET_https_www.boe.es_datosabiertos_api_legislacion-consolidada_id_BOE-A-2015-11722_metadatos.json
    A	internal/evals/testdata/boe.legislacion-consolidada/GET_https_www.boe.es_datosabiertos_api_legislacion-consolidada_id_BOE-A-2015-11722_texto_bloque_a1.json
    A	internal/evals/testdata/boe.legislacion-consolidada/GET_https_www.boe.es_datosabiertos_api_legislacion-consolidada_id_BOE-A-2015-11722_texto_indice.json
    A	internal/evals/testdata/boe.legislacion-consolidada/GET_https_www.boe.es_datosabiertos_api_legislacion-consolidada_q_limit_10_query_7B_22query_22_3A_20_-524cb2ad.json
    A	internal/evals/testdata/boe.legislacion-consolidada/GET_https_www.boe.es_datosabiertos_api_legislacion-consolidada_q_limit_10_query_7B_22query_22_3A_20_-744167c4.json
    A	internal/evals/testdata/boe.legislacion-consolidada/GET_https_www.boe.es_robots.txt.json
    A	internal/evals/testdata/grabaciones.json
    M	internal/evals/umbrales.go
    M	internal/evals/umbrales_test.go
    A	internal/skills/cobertura.go
    A	internal/skills/cobertura_test.go
    M	internal/skills/referencias.go
    M	internal/skills/referencias_test.go
    A	internal/source/bocm/cotejo_test.go
    A	internal/source/bocm/cruce.go
    A	internal/source/bocm/cruce_test.go
    A	internal/source/bocm/derivar.go
    A	internal/source/bocm/derivar_test.go
    A	internal/source/bocm/disposicion.go
    A	internal/source/bocm/disposicion_test.go
    A	internal/source/bocm/doc.go
    A	internal/source/bocm/grabacion_test.go
    A	internal/source/bocm/testdata/grabaciones.json
    A	internal/source/boe/fiestas/derivar.go
    A	internal/source/boe/fiestas/derivar_test.go
    A	internal/source/boe/fiestas/doc.go
    A	internal/source/boe/fiestas/grabacion_test.go
    A	internal/source/boe/fiestas/resolucion.go
    A	internal/source/boe/fiestas/resolucion_test.go
    A	internal/source/boe/fiestas/testdata/grabaciones.json
    M	internal/workflow/arbol_test.go
    A	internal/workflow/marcas_test.go
    A	schemas/cobertura.yaml.json
    M	schemas/eval.yaml.json
    A	schemas/plazo.json
    M	scripts/claude-modelo.sh
    A	scripts/workflow/marcas.sh
    M	scripts/workflow/verificar.sh
    M	skills/legal-core/SKILL.md
    A	skills/legal-core/references/cobertura_por_materia.md
    A	skills/legal-core/references/recursos_y_plazos.md
    A	specs/023-h9-internal-core-plazos/aceptacion/plazos-calcular.txtar
    A	specs/023-h9-internal-core-plazos/aceptacion/plazos-cobertura.txtar
    A	specs/023-h9-internal-core-plazos/aceptacion/plazos-codigos.txtar
    A	specs/023-h9-internal-core-plazos/aceptacion/plazos-herramienta.txtar
    A	specs/023-h9-internal-core-plazos/aceptacion/plazos-sin-efectos.txtar
    A	specs/023-h9-internal-core-plazos/checklists/requirements.md
    A	specs/023-h9-internal-core-plazos/contracts/SKILL-legal-core.prototipo.md
    A	specs/023-h9-internal-core-plazos/contracts/applet-plazos.md
    A	specs/023-h9-internal-core-plazos/contracts/casos-de-plazos.md
    A	specs/023-h9-internal-core-plazos/contracts/datos-de-plazos.md
    A	specs/023-h9-internal-core-plazos/contracts/evals-legal-core.md
    A	specs/023-h9-internal-core-plazos/contracts/skill-legal-core.md
    A	specs/023-h9-internal-core-plazos/data-model.md
    A	specs/023-h9-internal-core-plazos/gates/aceptacion-congelada.json
    A	specs/023-h9-internal-core-plazos/gates/analyze.md
    A	specs/023-h9-internal-core-plazos/gates/clarify-preguntas.json
    A	specs/023-h9-internal-core-plazos/gates/clarify-respuestas.json
    A	specs/023-h9-internal-core-plazos/gates/clarify-rondas
    A	specs/023-h9-internal-core-plazos/gates/cuarentena.json
    A	specs/023-h9-internal-core-plazos/gates/cuarentena/T001.patch
    A	specs/023-h9-internal-core-plazos/gates/cuarentena/T002.patch
    A	specs/023-h9-internal-core-plazos/gates/cuarentena/T003.patch
    A	specs/023-h9-internal-core-plazos/gates/dossier.md
    A	specs/023-h9-internal-core-plazos/gates/grabaciones.log
    A	specs/023-h9-internal-core-plazos/gates/grabaciones.md
    A	specs/023-h9-internal-core-plazos/gates/modelos.jsonl
    A	specs/023-h9-internal-core-plazos/gates/plan-correccion-r1.diff
    A	specs/023-h9-internal-core-plazos/gates/plan-precheck.txt
    A	specs/023-h9-internal-core-plazos/gates/plan-r1.json
    A	specs/023-h9-internal-core-plazos/gates/plan-r2.json
    A	specs/023-h9-internal-core-plazos/gates/plan-rondas
    A	specs/023-h9-internal-core-plazos/gates/plan.json
    A	specs/023-h9-internal-core-plazos/gates/spec-correccion-r1.diff
    A	specs/023-h9-internal-core-plazos/gates/spec-precheck.txt
    A	specs/023-h9-internal-core-plazos/gates/spec-r1.json
    A	specs/023-h9-internal-core-plazos/gates/spec-r2.json
    A	specs/023-h9-internal-core-plazos/gates/spec-rondas
    A	specs/023-h9-internal-core-plazos/gates/spec.json
    A	specs/023-h9-internal-core-plazos/gates/supuestos.md
    A	specs/023-h9-internal-core-plazos/gates/tarea-T001.md
    A	specs/023-h9-internal-core-plazos/gates/tarea-T002.md
    A	specs/023-h9-internal-core-plazos/gates/tarea-T003.md
    A	specs/023-h9-internal-core-plazos/gates/tarea-T011.md
    A	specs/023-h9-internal-core-plazos/gates/tarea-actual.json
    A	specs/023-h9-internal-core-plazos/gates/tareas-intentos.json
    A	specs/023-h9-internal-core-plazos/gates/tasks-correccion-r1.diff
    A	specs/023-h9-internal-core-plazos/gates/tasks-precheck.txt
    A	specs/023-h9-internal-core-plazos/gates/tasks-r1.json
    A	specs/023-h9-internal-core-plazos/gates/tasks-r2.json
    A	specs/023-h9-internal-core-plazos/gates/tasks-rondas
    A	specs/023-h9-internal-core-plazos/gates/tasks.json
    A	specs/023-h9-internal-core-plazos/mutantes.md
    A	specs/023-h9-internal-core-plazos/plan.md
    A	specs/023-h9-internal-core-plazos/quickstart.md
    A	specs/023-h9-internal-core-plazos/research.md
    A	specs/023-h9-internal-core-plazos/spec.md
    A	specs/023-h9-internal-core-plazos/tasks.md
