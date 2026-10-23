# Implementation Plan: H30 · Las evals miden solo lo que cambia: dependencias de cada eval y resultados por huella

**Branch**: `026-h30-las-evals-miden` | **Date**: 2026-10-22 | **Spec**: [spec.md](spec.md)

**Input**: Feature specification from `specs/026-h30-las-evals-miden/spec.md`

## Summary

El job de evals deja de medir las tres skills enteras con cualquier cambio. El arnés (`internal/evals`, que el binario no
enlaza) calcula el conjunto de dependencias de cada eval sin listas a mano —sus skills y las que nombran, sus
referencias y los `data/*.yaml` de los que salen, los paquetes Go de los applets de sus tablas de comandos (`go list`),
sus grabaciones y su juez—, elige las evals cuyo conjunto toca la propuesta (todas con lo común, una descripción o una
ruta sin asignar) y da a cada eval una huella, el sha256 de un texto canónico con todo lo que la elige. Cada trabajo
`evals (<skill>)` mide lo elegido y, para lo demás, busca en las 10 medidas de `main` más recientes un resultado con la
misma huella cuyas sesiones se abrieron hace menos de 90 días y lo reutiliza sin abrir sesión; lo que no encuentra lo
mide, y un error de la plataforma queda escrito junto a lo que se mide por él. El informe separa lo medido (los campos de
hoy, que lee `serie.sh` sin cambios) de lo reutilizado (bloque `origen` y umbrales `:reutilizadas`). La etiqueta
`evals-completa` y el despacho miden la suite completa, la tanda por commit distingue los dos modos, la release exige la
suite completa del commit etiquetado, y el ADR 0045 fija el límite binomial de la tasa de una clase para cualquier
número de respuestas medidas, que con todas es el de hoy (research D1 a D20).

## Technical Context

**Language/Version**: Go 1.27 (`go 1.27.0`, `toolchain go1.27.3` en `go.mod`); YAML de GitHub Actions; bash en
`scripts/evals.sh`.

**Primary Dependencies**: ninguna nueva. Biblioteca estándar (`crypto/sha256`, `go/parser`, `go/token`, `math`,
`math/big` en el test de la tasa, `os/exec`, `encoding/json/v2`) y las ya fijadas (`go.yaml.in/yaml/v3`, `stretchr/testify`); `gh` y `git` en el runner,
como hoy.

**Storage**: ninguno nuevo. Los resultados viven en el `informe.json` de cada trabajo, en el registro de la plataforma
(90 días); la decisión, en un fichero de `$RUNNER_TEMP`.

**Testing**: `go test` con `testify`; la plataforma, en memoria y con el `gh` sustituto de la tanda (research D15); `git
clone` local en `t.TempDir()`; `scripts/workflow/serie.sh` ejecutado tal cual. Sin red ni sesiones con modelo.

**Target Platform**: GitHub Actions `ubuntu-24.04` (el job y la release) y `make ci` en macOS y Linux; el cálculo fija
`GOOS=linux GOARCH=amd64 CGO_ENABLED=0` para no depender de dónde corre (research D3, V2).

**Project Type**: arnés de CI del binario multicall `kitlegal` (las evals de sus skills).

**Performance Goals**: la decisión de cada trabajo, como mucho 15 minutos (`-timeout`), con 10 para la plataforma
(research D17); el cálculo de `make ci` (un `go list` y el sha256 de los ficheros que miden) no se ha medido en esta
sesión.

**Constraints**: no más sesiones a la vez que hoy (5, 4 y 2; FR-020); nada cambia en `scripts/workflow/`, `.specify/` ni
`internal/workflow/` (FR-027); ninguna eval, umbral, clase, repetición, modelo ni versión cambia (FR-026).

**Scale/Scope**: 45 evals en 3 skills hoy; con materias como paquetes, cientos. Lo que crece está calculado en «Uso».

## Constitution Check

*GATE: pasado antes de la fase 0 y re-evaluado tras el diseño (fase 1): sin violaciones; la única desviación
justificada, «Aceptación e2e: no aplica», está en Complexity Tracking.*

### Principios

| Principio | Cómo lo cumple H30 |
|---|---|
| I · Fuentes públicas y frontera humana | Ninguna fuente nueva ni petición HTTP en Go: la plataforma se lee con `gh`, solo lecturas (listas de ejecuciones, trabajos, registros, propuestas y ficheros), como la tanda desde H7.4 (research V7). El binario no cambia. Crear la etiqueta `evals-completa`, etiquetar y publicar siguen siendo de una persona |
| II · Nada sin cita ni fuente | Ningún sobre ni ninguna skill cambia. Cada resultado reutilizado lleva su procedencia (ejecución, trabajo, commit, propuesta, fecha) y su huella: nada se da por medido sin decir dónde se midió |
| III · Tests primero y offline | «Aceptación e2e: no aplica» (abajo; Complexity Tracking). Cada pieza entra con su test en `make ci`, offline: la plataforma en memoria o con el `gh` sustituto, `go list` y `git clone` locales; ningún test toca la red. No hay fixtures nuevos en `testdata/` (research D15, D16) |
| IV · Hexagonal y errores tipados | Ningún paquete del binario cambia salvo dos funciones de lectura en `internal/skills`, puras; ningún error tipado ni código de salida de `kitlegal` cambia. En el arnés, un fallo es un error que nombra su causa; la plataforma que no responde no falla el trabajo, deja medir (FR-013); un estado manipulado, la regla genérica |
| V · Simplicidad y dependencias | Ninguna dependencia nueva. Dos funciones exportadas, en `internal/skills`, para no duplicar sus formatos (research D11); en `internal/evals`, nada exportado. Cada fichero y mecanismo, trazado a su FR («Trazabilidad») |
| VI · Un binario, convenciones de agente | El binario, sus banderas, `--describe` y las tablas de comandos no cambian; el arnés lee esas tablas |
| VII · Grafo y privacidad | Nada del grafo cambia. Lo que el arnés lee de la plataforma son informes de evals, sin datos de ningún asunto |
| VIII · Skills primero | Protege las tres skills: cada cambio de una skill se mide sin esperar a las demás, y lo que no cambia conserva el resultado que tiene en `main`; la release exige su suite completa. Ningún `SKILL.md` ni eval cambia |
| IX · Genericidad territorial | Nada territorial: el cálculo no nombra ningún municipio, comunidad ni boletín; sale de las rutas del repositorio |
| X · Genericidad por materia | El cálculo es por eval y por skill, sin ninguna marca de materia (Q2 de clarify, lectura conservadora): las referencias de una skill son dependencias de todas sus evals, y un paquete de materia futuro entrará como datos sin tocar este código hasta que H32 fije su molde |

### Reglas de dependencia (`docs/ROADMAP.md` §2, constitución IV, `internal/arch_test.go`)

| Regla | Cómo la cumple |
|---|---|
| R1 · `internal/core/**` no importa adaptadores | Sin cambios en `internal/core` |
| R2 · Solo `internal/httpx` importa `net/http` | Ningún import de `net/http`: la plataforma va por `gh` (proceso) desde el arnés |
| R3 · Solo `cache`/`store`/`graph` importan SQLite y `database/sql` | El arnés no abre ninguna base; `go list` solo lee el grafo de paquetes |
| R4 · Solo `cli` y `cmd/` llaman a `os.Exit` | Ningún `os.Exit`: los puntos de entrada son tests que fallan con `t.Fatal` |
| R5 · Solo `render` escribe en stdout; logs a stderr | Ningún `fmt.Print*` ni `os.Stdout` nuevo: los puntos de entrada escriben en los ficheros de sus banderas (`$GITHUB_OUTPUT`, la decisión) y registran con `t.Log` |
| R6 · `internal/graph` no importa `source/*` ni `render` | Sin cambios en `internal/graph` |
| R7 · Solo `internal/mcp` importa el SDK de MCP | Ningún import del SDK |
| «Sin red» de `arch_test.go` (instalación, disco, cita, ids, plazos, raíz y applets sin red) | Las funciones nuevas de `internal/skills` no importan nada de red; ningún paquete de esa lista cambia de importaciones |

## Project Structure

### Documentation (this feature)

```text
specs/026-h30-las-evals-miden/
├── spec.md · plan.md · research.md · data-model.md · quickstart.md
├── contracts/
│   ├── eleccion-y-huella.md      # conjunto, elección, huella y su texto canónico
│   ├── informe-del-job.md        # bloque origen, umbrales :reutilizadas, tasa, veredicto, tamaño
│   ├── job-de-evals.md           # disparadores, modo, trabajos, pasos, órdenes de la plataforma
│   └── release.md                # trabajo suite-completa y su mensaje
├── checklists/  gates/
└── tasks.md                      # /speckit-tasks
```

### Source Code (repository root)

```text
internal/evals/                    # el arnés (lo común); nada exportado nuevo
├── dependencias.go (+_test)       # rutas que miden, lo común, conjunto de cada eval, árbol de cabeza y base   FR-001..005
├── eleccion.go (+_test)           # elección y motivos; golden de la matriz en el test                            FR-002..008, FR-019
├── huella.go (+_test)             # texto canónico y sha256                                                        FR-009
├── plataforma.go (+_test)         # interfaz, implementación con gh, informe de un registro                       FR-007, FR-010, FR-025
├── reutilizacion.go (+_test)      # medidas de main y búsqueda del resultado por huella                           FR-010..013
├── decision.go (+_test)           # decisión de cada trabajo y de la apertura                                     FR-012..014, FR-016..018
├── origen.go                      # bloque origen, aportes, umbrales :reutilizadas                                FR-011, FR-015, FR-022, FR-023
├── tasa.go (+_test)               # N, K, pref y k(n)                                                              FR-024
├── suite_completa.go (+_test)     # la comprobación de la release                                                 FR-025
├── informe.go, umbrales.go        # componen con origen; la regla de la tasa (tests en informe_test.go y serie_test.go)
├── plan.go, ejecucion.go          # el plan, solo con las evals medidas
├── tanda.go (tanda_test.go)       # el modo de cada tanda                                                         FR-021
├── definicion.go (definicion_test.go)  # lo que se comprueba de evals.yml                                       FR-016..021
├── job_test.go                    # entradas: TestEleccionDeLaPropuesta, TestDecisionDelTrabajo, TestSuiteCompletaDelCommit; -decision y -modo
└── doc.go                         # el paquete, con lo nuevo
internal/skills/comandos.go, referencias.go (+_test)   # AppletsDeLaTabla, DatosDeLaReferencia                   FR-001(c)(d)
.github/workflows/evals.yml        # cambios con Go, tanda con modo, decisión, marca del informe, permisos      FR-010, FR-016..021
.github/workflows/release.yml      # trabajo suite-completa; publicar depende de él                             FR-025
release_test.go                    # subtest suite-completa de TestConfiguracionDeLaRelease                       FR-025
scripts/evals.sh                   # DECISION_DE_EVALS obligatoria → -decision                                    FR-012, FR-019
docs/ADR/0045-lo-reutilizado-no-cuenta-en-la-tasa.md                                                              FR-024, SC-013
CHANGELOG.md · docs/WORKFLOW.md · CONTRIBUTING.md                                                                  SC-016
```

**Structure Decision**: todo el cálculo en el arnés, que es lo común y el binario no enlaza (research D1); los puntos de
entrada son tests con la etiqueta `evals`, el patrón de `TestTandaDelCommit` y `TestEjecucionDelJob`. Ningún paquete
nuevo; nueve ficheros de código nuevos en `internal/evals`, uno por responsabilidad con su FR.

## Aceptación e2e

**Aceptación e2e: no aplica.** El hito no cambia nada que vea quien usa `kitlegal`: el binario no enlaza
`internal/evals`, ningún `SKILL.md` ni eval cambia (FR-026), y las dos funciones nuevas de `internal/skills` solo leen
(research D20). No hay comportamiento del binario que un guion `testscript` describa, ni evals nuevas. La aceptación del
spec (SC-006 y SC-007) son tests de `make ci` sobre la decisión del job, con la plataforma en memoria
(`TestDecisionDelJob`, subtests `solo-el-cuerpo-de-jurisprudencia` y `etiqueta-evals-sin-cambios`), y la medición de
cierre del workflow corre ya con el job nuevo (SC-014).

## Controles mecánicos que este hito añade o toca

### Objetivos del `Makefile`

Ninguno nuevo ni cambiado: `make ci` ejecuta los tests nuevos por `test` (`go test -race -shuffle=on ./...`), y `make
evals` sigue llamando a `scripts/evals.sh`, que ahora exige `DECISION_DE_EVALS`.

### Tests nuevos (`make ci`)

| Test | Fichero | Qué fija |
|---|---|---|
| `TestConjuntoDeCadaEval` | `internal/evals/dependencias_test.go` | Las evals del repositorio, cada una con conjunto calculable, no vacío y sus seis partes; lo que FR-001 pide para una de `legal-core` y una de `jurisprudencia`; mutante con un applet inexistente (FR-001, FR-005, SC-002) |
| `TestEleccionDeLosMutantes` | `internal/evals/eleccion_test.go` | Los mutantes de SC-003 y de FR-002 a FR-007 (los diez de quickstart.md §2), eval por eval; el esperado sale del listado de `evals/<skill>/`, no de los conjuntos |
| `TestMatrizDeLasPropuestas` | `internal/evals/eleccion_test.go` | Golden (constante del test) de lo que mide y reutiliza cada skill para los mutantes, una propuesta sin nada elegido y la suite completa, con una medida de `main` sintética con la huella de cada eval (FR-008, FR-019, SC-004) |
| `TestHuellaDeLasEvals` | `internal/evals/huella_test.go` | Cada clase de ruta cambia la huella de las que elige y no la de las demás; la versión de un módulo de lo común en `go.mod`, todas; un `replace` y una línea de `go.sum` de un módulo de ningún conjunto, todas; `docs/`, ninguna; dos clones, las mismas (FR-009, SC-008) |
| `TestOrdenesDeLaPlataforma`, `TestInformeDelRegistro` | `internal/evals/plataforma_test.go` | Las órdenes de `gh` con el sustituto; el informe sacado de un registro con prefijos y marca de orden de bytes |
| `TestBusquedaDeMedidasDeMain` | `internal/evals/reutilizacion_test.go` | Canceladas, sin fusionar, sin marca, de 90 días o más y otro commit de la propuesta no cuentan; la más reciente gana; la undécima no se recorre; sesión sin medir o respuesta sin juzgar no se reutiliza; la procedencia de una reutilizada en `main` es la original, y con 90 días o más la eval se mide con `sin_resultado_de_main`; una orden de `gh` que falla o el plazo agotado paran la búsqueda, lo que queda se mide y la orden y su error quedan en el resultado (FR-010 a FR-013) |
| `TestDecisionDelJob` | `internal/evals/decision_test.go` | SC-006, SC-007, FR-014 (solo la eval sin resultado, con su motivo), FR-016 (sin elección, nada), suite completa sin búsqueda; ficheros de la propuesta que `gh` no da: todas las evals medidas con `plataforma` y el error en `origen` (D17) |
| `TestInformeConReutilizadas` | `internal/evals/informe_test.go` | Veredicto con serie reutilizada en fallo y con respuesta reutilizada sin activar; `:reutilizadas`; suma de aportes = umbral; tasa solo con lo medido (FR-015, FR-022, FR-024) |
| `TestInformeDeLaSuiteCompletaNoCambia` | `internal/evals/informe_test.go` | El informe de la suite completa de datos de prueba, sin `origen`, byte a byte igual al fijado antes de tocar `informe.go` (FR-026, SC-012) |
| `TestFilaConReutilizadas` | `internal/evals/serie_test.go` | `scripts/workflow/serie.sh fila` sale con 0 y da la misma fila que con solo lo medido (FR-023, SC-009) |
| `TestLimiteDeLaTasa` | `internal/evals/tasa_test.go` | Para cada skill con juez y cada *n* de 1 a *N*, P(Bin(*n*, *p*ref) > k(*n*)) ≤ 0,02 con racionales exactos (`math/big`); k(*N*) = límite de hoy; mutante `umbral: 0.017` → k(120) = 2 (FR-024, SC-010) |
| `TestTandaConLaSuiteCompleta` | `internal/evals/tanda_test.go` | Las dos combinaciones de FR-021 |
| `TestComprobarLaSuiteCompleta` | `internal/evals/suite_completa_test.go` | Los casos de [contracts/release.md](contracts/release.md) §3 (FR-025, SC-005) |
| `TestAppletsDeLaTabla`, `TestDatosDeLaReferencia` | `internal/skills/comandos_test.go`, `referencias_test.go` | Las dos lecturas con las tablas y cabeceras de las tres skills, y sin marcas o sin cabecera |

### Tests existentes que cambian lo que esperan

- `TestDefinicionDelJob` (`definicion_test.go`): el trabajo `cambios` con `TestEleccionDeLaPropuesta`; la condición y
  las salidas de `tanda` con el modo; en `evals`, los permisos, el `fetch-depth`, el paso de la decisión antes de
  «Ejecutar las evals» con `GH_TOKEN` solo en él, `DECISION_DE_EVALS` y la marca final con `!cancelled()`; el peor caso
  con los 900 s de la decisión; y, como hoy, una sola combinación por skill con 5, 4 y 2 (FR-016 a FR-020, SC-011).
  Sus subtests `segundo-disparo` y `estado-de-la-tanda` añaden el modo.
- `TestConfiguracionDeLaRelease` (`release_test.go`): el conjunto de trabajos y su orden incluyen `suite-completa`.
- Hoy ningún test de `make ci` ejecuta `scripts/evals.sh` (solo `scripts/evals-sesion.sh`, en `sesiones_test.go`;
  comprobado con `grep` al escribir tasks.md). Uno nuevo, en `sustitutos_test.go` y con el `claude` sustituto delante
  en el `PATH`, comprueba que sin `DECISION_DE_EVALS`, o con una ruta que no es un fichero legible, sale con 1 y el
  mensaje «evals: falta DECISION_DE_EVALS» sin que llegue ninguna sesión al sustituto.
- Los tests del umbral conjunto de una clase en `umbrales_test.go` e `informe_test.go` con menos respuestas que las de
  su plan entero esperan el límite k(*n*) (FR-024); con todas, lo de hoy.

### Puntos de entrada fuera de `make ci` (etiqueta `evals`)

`TestEleccionDeLaPropuesta`, `TestDecisionDelTrabajo`, `TestSuiteCompletaDelCommit` (nuevos) y `TestTandaDelCommit`
(`-modo`) y `TestEjecucionDelJob` (`-decision`), en `job_test.go`; banderas en [contracts/job-de-evals.md](contracts/job-de-evals.md) §4.

### Fixtures, `testdata/` y `schemas/`

Ninguno: las respuestas de la plataforma, la medida de `main` y el golden de la matriz los construye cada test en
memoria o en `t.TempDir()` (research D15, D16). Ningún esquema nuevo.

### CI

`.github/workflows/evals.yml` y `.github/workflows/release.yml` (contratos). `make ci` no cambia de objetivos.

## Controles de umbral

| Requisito | Umbral | Control | Dónde |
|---|---|---|---|
| SC-001 | `make ci` termina con 0 | Todos los tests de esta tabla y los de siempre | `ci:Makefile:ci` |
| FR-001, FR-005, SC-002 | 0 evals sin conjunto calculable o con conjunto vacío; las seis partes y lo que FR-001 pide y excluye; 1 applet inexistente → rojo | Recorre todas las evals del repositorio y el mutante | `ci:internal/evals/dependencias_test.go:TestConjuntoDeCadaEval` |
| FR-002 a FR-004, FR-006, FR-007, SC-003 | 0 evals de diferencia con lo esperado en cada mutante | Un subtest por mutante, eval por eval | `ci:internal/evals/eleccion_test.go:TestEleccionDeLosMutantes` |
| FR-008, FR-019, SC-004 | 0 bytes de diferencia con el golden | Comparación byte a byte | `ci:internal/evals/eleccion_test.go:TestMatrizDeLasPropuestas` |
| FR-009, SC-008 | 0 huellas cambiadas fuera de las elegidas, 0 sin cambiar entre las elegidas, 0 diferencias entre dos directorios | Un subtest por clase de ruta y los dos clones | `ci:internal/evals/huella_test.go:TestHuellaDeLasEvals` |
| FR-010 a FR-013 | 0 medidas inválidas usadas (cancelada, sin fusionar, sin marca, de 90 días o más, otro commit); como mucho 10 recorridas; 0 resultados incompletos reutilizados; 0 procedencias que no sean la de la medida que abrió las sesiones; 0 resultados reutilizados con procedencia de 90 días o más; 0 errores de la plataforma sin escribir | Respuestas sintéticas de la plataforma y un instante fijo | `ci:internal/evals/reutilizacion_test.go:TestBusquedaDeMedidasDeMain` |
| FR-013, FR-014, FR-016, FR-017, SC-006, SC-007 | 0 sesiones de `boe-legislacion` y `legal-core` con solo el cuerpo de `jurisprudencia`; 0 sesiones y 45 reutilizadas con la etiqueta sin cambios; solo las sesiones de la eval sin resultado; con los ficheros de la propuesta sin leer, 0 evals de la skill sin medir y el error escrito | Sesiones del plan de cada trabajo | `ci:internal/evals/decision_test.go:TestDecisionDelJob` |
| FR-015, FR-022 | 1 serie reutilizada en fallo o 1 respuesta reutilizada sin activar → fallo | Informes compuestos con datos de prueba | `ci:internal/evals/informe_test.go:TestInformeConReutilizadas` |
| FR-023, SC-009 | `serie.sh fila` sale con 0 y 0 diferencias con la fila de solo lo medido | Ejecuta el guion de `main` | `ci:internal/evals/serie_test.go:TestFilaConReutilizadas` |
| FR-024, SC-010, SC-013 | P(rojo sin motivo) ≤ 0,02 para cada *n* de 1 a *N*; k(*N*) = límite de hoy; mutante 2 de 120 → k(120) = 2 | Binomial exacta con `math/big` | `ci:internal/evals/tasa_test.go:TestLimiteDeLaTasa` |
| FR-016, FR-018, FR-020, SC-011 | Modo `completa` solo con el despacho, `evals-completa` y `evals-prueba-de-red`; la tanda, solo con uno de ellos, `evals` o `elige=si`; 1 trabajo por skill; 5, 4 y 2 a la vez; 11 en total | Lee la definición del job | `ci:internal/evals/definicion_test.go:TestDefinicionDelJob` |
| FR-021 | Una con reutilización no mide con otra en curso; una completa no deja de medir por una con reutilización | Las dos combinaciones | `ci:internal/evals/tanda_test.go:TestTandaConLaSuiteCompleta` |
| FR-025, SC-005 | 0 publicaciones sin la suite completa: rojo sin el trabajo, sin la dependencia o con otro commit | Lee `release.yml` | `ci:release_test.go:TestConfiguracionDeLaRelease` |
| FR-025, SC-005 | Rojo con otra ejecución de commit, en rojo, con una eval reutilizada o sin un informe | Respuestas sintéticas de la plataforma | `ci:internal/evals/suite_completa_test.go:TestComprobarLaSuiteCompleta` |
| FR-026, SC-012 | 0 bytes de diferencia, sin `origen`, con el informe de antes del hito | Golden fijado antes de tocar `informe.go` | `ci:internal/evals/informe_test.go:TestInformeDeLaSuiteCompletaNoCambia` |

Sin fila, porque no son un umbral que mida `make ci` ni el job: SC-014 lo comprueba el cierre con la comprobación de
`cierre.sh` que ya existe (la fila de la serie que no casa con los umbrales de su informe lo detiene); SC-015 (el diff
sin `scripts/workflow/`, `.specify/` ni `internal/workflow/`) lo hace cumplir el guardián de diff del workflow, porque
ninguna tarea declara esas rutas; SC-016 es documentación que juzga la revisión final.

## Uso, de fuera adentro (criterio de uso, ADR 0028)

Ninguna skill pide estas salidas: las leen personas y el cierre, y crecen con las evals, no con las preguntas ni con
los meses de uso de `kitlegal` (sus normas y bloques consultados no entran en ninguna sesión de evals).

| Salida | Quién, cuántas veces | Forma y ejemplo | Tamaño hoy → con cientos de evals | Cuándo deja de darse |
|---|---|---|---|---|
| Elección con su motivo (registro de `cambios` y de cada trabajo) | Quien revisa, una vez por ejecución | `elegida legal-core/05-plazo-reposicion-leganes: conjunto: skills/legal-core/references/recursos_y_plazos.md` (107 bytes) | ≤ 45 líneas de unos 60 + 50·*r* bytes (*r*, rutas que la eligen) → crece con las elegidas y las rutas de la propuesta | Al fusionar: sobre `main` no hay propuesta, y la suite completa no elige |
| Bloque `origen` del informe | Quien revisa el informe y el cierre (`gates/evals/<skill>.json`), una vez por ejecución | [contracts/informe-del-job.md](contracts/informe-del-job.md) §2: huella de 64, procedencia de ~150 bytes, aportes de ~70 | Medida: 0,7–1,7 KB; reutilizada: +1,4 KB de series. Hoy 30–140 KB entre los tres informes (que ya son 1,8 MiB) → ~1 MB con 300 evals. No crece con la historia: un resultado por eval | La marca «reutilizada» de una eval cae cuando cambia su huella, cuando la ejecución en que se abrieron sus sesiones (su procedencia) cumple 90 días —la vida de su registro—, aunque una medida posterior la haya reutilizado, o cuando su medida sale de las 10 más recientes: entonces se mide (data-model §8) |
| Error de la plataforma (`origen.error_de_la_plataforma`, registro del paso de la decisión y de `cambios`) | Quien revisa, en la ejecución en que falla una orden de `gh` o se agota el plazo | La orden y su salida de error, o «plazo de 10 minutos agotado» con la orden en curso ([contracts/informe-del-job.md](contracts/informe-del-job.md) §2) | Unas líneas, uno como mucho por trabajo y ejecución: la búsqueda para en el primer error. No crece con las evals | En la primera ejecución en que la plataforma responde |
| Búsqueda de medidas de `main` | Cada trabajo con reutilización y evals sin elegir, una vez por ejecución | Lista de ejecuciones (≤ 200), propuestas cerradas (≤ 100), trabajos de cada ejecución mirada, y el registro del trabajo de su skill | Lo normal, 1 registro con un informe de ≤ 922 KiB; como mucho 10 por trabajo, 30 entre los tres. El tamaño del registro entero de un trabajo no se ha medido en esta sesión | Para en cuanto cada eval no elegida tiene resultado; con la cota, se mide |
| Fallo de la release | Quien empuja la etiqueta, una vez por etiqueta | [contracts/release.md](contracts/release.md) §2 | 3–12 líneas | Cuando el commit tiene su suite completa en verde |
| Fila de la serie de cierres | El cierre, una por skill con juez y medición | La de hoy (`schemas/serie-de-cierres.json`) | La de hoy: cuenta solo lo medido | Como hoy |

Ninguna señal se repite sin información nueva: la elección y la decisión salen una vez por ejecución, y lo reutilizado
lleva una sola procedencia por eval.

## Decisiones

Todas en [research.md](research.md), con su alternativa: D1 dónde vive; D2 conjunto por entradas; D3 `go list` y
`go/parser`; D4 lo común y las rutas que miden, sin los tests ni `testdata/` de los paquetes de un conjunto; D5 base local y rutas de la plataforma; D6
huella; D7 `go.mod` y `go.sum`, por versiones y resto; D8 medidas de `main` validadas por un paso; D9 series y aportes; D10 `origen` y
`:reutilizadas`; D11 lecturas en `internal/skills`; D12 límite y publicación de la tasa; D13 tanda con modo; D14 release;
D15 plataforma en los tests; D16 golden en el test; D17 decisión acotada; D18 `DECISION_DE_EVALS` obligatoria; D19
motivo; D20 aceptación e2e. Supuestos no verificados sin red: S1 a S8.

## Trazabilidad: cada mecanismo y su requisito

| Mecanismo | Requisito |
|---|---|
| `skills.AppletsDeLaTabla`, `skills.DatosDeLaReferencia` | FR-001(c)(d), FR-005 |
| Rutas que miden, lo común, conjunto por entradas, árbol de cabeza y base (`dependencias.go`) | FR-001 a FR-005, FR-009 |
| Elección, motivos y su orden (`eleccion.go`) | FR-002 a FR-008 |
| Huella y su texto canónico (`huella.go`) | FR-009 |
| Interfaz de la plataforma y su implementación con `gh`; extracción del informe de un registro (`plataforma.go`) | FR-007, FR-010, FR-025 |
| Medidas de `main`, la marca «El informe del trabajo está escrito», búsqueda con cota 10, vida de 90 días de la procedencia (`reutilizacion.go`, `evals.yml`) | FR-010 a FR-013 |
| `origen.error_de_la_plataforma` y el motivo `plataforma`: la orden de `gh` que falla, o el plazo, escrita junto a lo que se mide por ella (`decision.go`, `eleccion.go`) | FR-007, FR-013, FR-022 (y FR-005 como lectura conservadora de una elección incalculable) |
| Decisión de cada trabajo y de la apertura (`decision.go`), `TestDecisionDelTrabajo`, `TestEleccionDeLaPropuesta`, trabajo `cambios` con Go | FR-012 a FR-014, FR-016, FR-017 |
| Bloque `origen`, aportes, umbrales `:reutilizadas`, veredicto (`origen.go`, `informe.go`, `umbrales.go`) | FR-011, FR-015, FR-022, FR-023 |
| Plan solo con las evals medidas; `-decision`; `DECISION_DE_EVALS` en `scripts/evals.sh` | FR-012, FR-014, FR-019 |
| Límite binomial y publicación del umbral (`tasa.go`, `umbrales.go`); ADR 0045 | FR-024, SC-010, SC-013 |
| Modo de la tanda y su segunda marca (`tanda.go`, `evals.yml`) | FR-018, FR-021 |
| Permisos, `fetch-depth`, paso de la decisión y peor caso del tope (`evals.yml`, `definicion.go`) | FR-016 a FR-020 |
| Trabajo `suite-completa`, `TestSuiteCompletaDelCommit`, `suite_completa.go` | FR-025 |
| Golden del informe de la suite completa, fijado primero | FR-026 |
| CHANGELOG, `docs/WORKFLOW.md`, `CONTRIBUTING.md`, cabecera de `evals.yml`, `doc.go` | SC-016 |

Nada se diseña para estados que no pasan el umbral de materialidad: un informe, registro o respuesta manipulados o
malformados no se reutilizan y la eval se mide, o el paso falla como defecto `inesperado` (spec, «Edge Cases»).

## Datos externos

Ninguno (research, «Datos externos»): ninguna fuente de `docs/SOURCES.md`, ningún manifiesto `grabaciones.json`,
ningún paso `grabar_datos` y ningún fichero de `data/`. Las respuestas de la plataforma de los tests son sintéticas.

## Orden de implementación (de dentro afuera)

1. Fijar en `TestInformeDeLaSuiteCompletaNoCambia` el informe de hoy de la suite completa de unos datos de prueba, antes
   de tocar `informe.go` (FR-026).
2. `internal/skills`: `AppletsDeLaTabla` y `DatosDeLaReferencia`, con sus tests.
3. `dependencias.go` y `TestConjuntoDeCadaEval`.
4. `eleccion.go` y `TestEleccionDeLosMutantes`.
5. `huella.go` y `TestHuellaDeLasEvals`.
6. `tasa.go`, la regla en `umbrales.go` y `TestLimiteDeLaTasa`.
7. `origen.go` y la composición del informe con `origen`, aportes y `:reutilizadas`; `TestInformeConReutilizadas` y
   `TestFilaConReutilizadas`.
8. `plataforma.go` y sus tests.
9. `reutilizacion.go` y sus tests.
10. `decision.go`, el plan con solo lo medido, `-decision`, `TestDecisionDelTrabajo`, `TestEleccionDeLaPropuesta` y
    `TestDecisionDelJob`; después `TestMatrizDeLasPropuestas`.
11. `tanda.go` con el modo y `TestTandaConLaSuiteCompleta`.
12. `.github/workflows/evals.yml`, `scripts/evals.sh` y `TestDefinicionDelJob`.
13. `suite_completa.go`, `TestSuiteCompletaDelCommit`, `TestComprobarLaSuiteCompleta`, `release.yml` y `release_test.go`.
14. ADR 0045, `CHANGELOG.md`, `docs/WORKFLOW.md`, `CONTRIBUTING.md` y `internal/evals/doc.go`. El ADR toma la tasa
    acumulada de la serie de `boe-legislacion` de `specs/*/gates/serie.jsonl` y `evidencias/adr-0038/serie-previa.jsonl`
    con `scripts/workflow/serie.sh`, sin copiarla del spec.

## Complexity Tracking

| Violation | Why Needed | Simpler Alternative Rejected Because |
|---|---|---|
| Principio III, «cada hito empieza por el test e2e (`testscript`)»: Aceptación e2e no aplica | El hito no cambia ningún comportamiento del binario; lo que entrega es el job, su arnés y la release (research D20) | Un guion `testscript` sobre `kitlegal` no describiría nada de la entrega; la aceptación del spec (SC-006, SC-007) son tests de `make ci` sobre la decisión del job, como en H24, H26 y H26.1 |

## Comprobación contra la rúbrica del juez (`juez_plan`) y `precheck.sh plan`

- `precheck.sh plan`: `research.md` existe; ninguna marca de aclaración pendiente; `## Constitution Check` con una fila por cada
  principio I a X; la línea «Aceptación e2e»; `## Controles de umbral` con filas que nombran FR/SC definidos y un `ci:`
  en cada «Dónde».
- a, c: cada principio y cada regla de dependencia, con lo que hace este plan. b: sin dependencias nuevas. d: el binario
  no cambia de errores ni códigos; el arnés falla con errores que nombran su causa. e: aceptación «no aplica» con
  motivo, tests nuevos y cambiados, sin fixtures. f: nada fuera del spec («Fuera de alcance» respetado, research). g:
  ningún `nolint`, `t.Skip` ni TODO previsto. h: cada decisión con su alternativa (research). i: lo comprobado, con
  dónde (V1 a V8); lo no comprobable sin red, S1 a S8. j: el quickstart solo ejecuta tests y lecturas de git. k: sin
  datos externos. l: ninguna tarea pide a una persona ni toca la plataforma. m: «Uso». n: «Trazabilidad». o: «Controles
  de umbral».
