# Implementation Plan: H7.4 · `boe-legislacion` sin el estado de la comprobación ni lo que no ha leído, y un job que juzga la respuesta a la pregunta

**Branch**: `014-h7-4-boe-legislacion-sin` | **Date**: 2026-09-30 | **Spec**: [spec.md](./spec.md)

**Input**: Feature specification from `/specs/014-h7-4-boe-legislacion-sin/spec.md`

**Modo**: desatendido. Las decisiones técnicas se tomaron con el «Criterio de decisión autónoma» de la constitución y
están en [research.md](./research.md) (D1-D21), cada una con su alternativa rechazada. Toda afirmación sobre Claude Code,
GitHub Actions, `gh`, `gosec`, PowerShell, Go o el repositorio remite a la tabla V de research.md, comprobada en local; lo
que no se puede comprobar así son los supuestos S1-S9. Las medidas de la lista y de la prosa se tomaron con dos
herramientas de un solo uso fuera de lo versionado (research, cabecera).

## Summary

H7.4 no toca el binario. Arregla `boe-legislacion` y el instrumento que la mide, en siete piezas:

1. **`boe-legislacion` v0.1.4** (research «Causa de raíz», D1-D5): la `description` dice que la skill se usa siempre
   que la respuesta dependa de lo que dice una norma, también si el modelo cree saberla, porque sin leerla no hay cita
   (la 04: dos sesiones de 3 sin activar); la prosa pierde «lectura anterior», «consulta anterior», «la comprobación» y
   «si lo leído basta», que la respuesta repite (27 de 51 en la línea de base); «Redacción modificada» dice por qué no
   se describe la redacción superada, qué sí se dice y qué responder si se pregunta qué cambió; cada línea
   `⚠ REDACCIÓN MODIFICADA:` lleva la cita de su bloque; y cada orden de lectura y comprobación lleva su forma para
   PowerShell (`…; if ($LASTEXITCODE -eq 0) { … }`).
2. **La lista por clases** (D6-D8): `anuncio` gana 25 formas de la clase A y la familia nueva `redaccion_no_leida`
   tiene 10 de la clase B; antes de buscar se quitan las `formas_fijas` que declara la lista (la línea
   `⚠ REDACCIÓN MODIFICADA:`, con su cita o sin ella, y la frase de la regla 7) y las etiquetas de los avisos. Marca
   exactamente 36, 11 y 9 en los tres informes versionados, eval por eval (medido).
3. **Dos umbrales nuevos que deciden** (D13, D14): `sin_activar:claude-sonnet-5-5` y
   `redaccion_no_leida:claude-sonnet-5-5`, los dos `<=` 0; con `expresiones_prohibidas:` (≤ 0,05) y la duración (≤ 900 s),
   cuatro que deciden. Los recuentos y los umbrales cuentan solo las sesiones terminadas.
4. **La eval 20** (D9-D11): lee `a1-30` y `da-3` de la LCSP con las dos redacciones cambiadas y exige dos líneas
   `⚠ REDACCIÓN MODIFICADA:`, cada una con su cita y sus fechas; el formato de eval gana `redacciones_modificadas` y
   `no_se_activan` (las tres de `legal-core` declaran que `boe-legislacion` no se activa).
5. **La respuesta a la pregunta** (D12): el primer `result` del transcript, no el último.
6. **Una tanda por commit** (D15-D17): un trabajo `tanda` decide, antes de `evals`, si otra ejecución anterior sobre el
   mismo commit ya mide; si mide, `evals` se salta entero (sin comprobación roja ni una `evals (<skill>)` de más).
7. **El sondeo sin traza en sus errores de uso** (D18).

No cambia ninguna decisión de arquitectura: no hay ADR nuevo, y `specs/011-…`, `specs/012-…` y `specs/013-…` no se
editan (FR-090).

## Technical Context

**Language/Version**: Go 1.27 (`go 1.27.0`, `toolchain go1.27.1`), `CGO_ENABLED=0`, `-trimpath`; bash (el guion del
sondeo, sin nada posterior a 3.2); YAML de GitHub Actions.

**Primary Dependencies**: las fijadas, sin ninguna nueva: `go.yaml.in/yaml/v3` (la del repositorio para yaml.v3),
`santhosh-tekuri/jsonschema` v6 (la lista y la eval con sus esquemas), `stretchr/testify`; biblioteca estándar
(`regexp`, `os/exec`, `encoding/json`, `time`). En el job, `gh` (ya en el runner y en el paso de `cambios`).

**Storage**: ninguno nuevo. Ficheros del repositorio: la lista y la eval 20 (`evals/`), sus esquemas (`schemas/`), la
skill (`skills/`), la definición del job, las derivadas del grafo previo (`testdata/evals/grafo-previo/`) y un caso de
lectura de sesión (`internal/evals/testdata/`).

**Testing**: `go test -race` en `make ci` (tablas, `t.Parallel`, sesiones y transcripts sintéticos, sustitutos POSIX de
`go`); puntos de entrada con la etiqueta `evals` fuera de `make ci` (`TestTandaDelCommit` nuevo, `TestSondeo`); el job
de evals en la propuesta de cambio.

**Target Platform**: el job, `ubuntu-24.04`; `make ci` y el sondeo, macOS y Linux. El binario distribuido no cambia
salvo la skill empotrada (`skills.go`).

**Project Type**: skill (`boe-legislacion`) + formato de eval, job y sondeo (`internal/evals`, `scripts/`, el flujo
`evals`).

**Performance Goals**: sesiones de `boe-legislacion` ≤ 900 s en el job (421 s con 93 sesiones en la línea de base; 96
ahora); peor caso del trabajo ≤ 122 min (7 285 s, D17); la tanda decide en segundos.

**Constraints**: ningún test de `make ci` usa la red ni abre una sesión con modelo; `SKILL.md` < 300 líneas (295 en v0.1.4:
las 294 del prototipo y el párrafo que cierra «Redacción modificada», que se queda, supuesto T010) y `description`
≤ 1024 caracteres (924); ninguna fecha `AAAAMMDD` con cifras en `SKILL.md`; ningún
`//nolint`.

**Scale/Scope**: 20 evals de `boe-legislacion` (97 sesiones con la prueba de red: 54 del modelo que decide y 30 del
informativo en evals que activan la skill) y 3 de `legal-core`; una lista de 87 expresiones y 2 formas fijas.

## Constitution Check

*GATE: pasado antes de la fase 0 y re-evaluado tras el diseño (fase 1): sin violaciones; las desviaciones justificadas
están en Complexity Tracking.*

### Principios

| Principio | Cómo lo cumple H7.4 |
|---|---|
| I · Fuentes públicas y frontera humana | No toca ninguna fuente legal, `internal/httpx` ni `docs/SOURCES.md`, y no graba nada (D21). Las sesiones siguen tras el proxy que solo deja pasar al modelo. La única red nueva es la del trabajo `tanda` en CI: `gh` leyendo la API de GitHub, como ya hace el paso de `cambios`; no es el binario, no es una fuente y no importa `net/http` (R2). |
| II · Nada sin cita ni fuente | El sobre no cambia. La skill deja de describir una redacción que ninguna orden devolvió (FR-011, FR-022), cada línea `⚠ REDACCIÓN MODIFICADA:` lleva su cita (FR-023), y la activación (FR-001) evita la respuesta sin sobre; `sin_activar` y `redaccion_no_leida` deciden con 0 (FR-041, FR-042). El comentario de la eval 20 nombra las normas modificadoras por el identificador de la grabación. |
| III · Tests primero y offline | «Aceptación e2e: no aplica» (abajo, con el motivo). Cada pieza entra con sus tests en `make ci`, offline, con sesiones y transcripts sintéticos y JSON sintético de `gh`; los datos de prueba nuevos (un transcript, dos derivadas) van en tareas `[datos]`; la aceptación es el job (SC-001). Cobertura dentro de `codecov.yml`. |
| IV · Hexagonal y errores tipados | Ningún paquete del binario ni código de salida de `kitlegal` cambian. `internal/evals` devuelve `error` con el fichero, la orden o el argumento delante; `errorDeUso` es el error tipado de los errores de uso del sondeo; los guiones salen con 1; ningún `panic`. Ni `os.Exit` ni salida estándar desde Go (R4, R5). |
| V · Simplicidad y dependencias | Ninguna dependencia nueva. Cada mecanismo se traza a un FR («Trazabilidad»); se rechazaron una familia `comprobacion` aparte, el anidado por clases, las formas fijas en Go, varias carpetas por grafo previo y la decisión de la tanda en bash con `jq` (research D2, D6, D7, D11, D15). |
| VI · Un binario, convenciones de agente | El binario, sus banderas y `--describe` no cambian; ningún ejecutable nuevo en `cmd/`: la decisión de la tanda es un punto de entrada de test con la etiqueta `evals`, como `TestEjecucionDelJob` y `TestSondeo`. |
| VII · Grafo y privacidad | Nada nuevo en el grafo; las derivadas del grafo previo las produce código (`TestGrabacionesDerivadas -actualizar-derivadas`) desde grabaciones de H4. El trabajo `tanda` solo lee ejecuciones y trabajos del flujo con `actions: read`. |
| VIII · Skills primero | Mejora `boe-legislacion` (v0.1.4) y la mide con sus evals, con dos umbrales más que deciden y una eval nueva (FR-050); protege el instrumento que la mide (la respuesta juzgada, la tanda única, el sondeo). `SKILL.md` < 300 líneas, sin nombrar evals ni modelos. |
| IX · Genericidad territorial | Nada se particulariza para un municipio: ni la lista, ni la skill, ni el job. Los municipios de las evals de `legal-core` (Leganés, Tordesillas) ya estaban; solo ganan `no_se_activan`. |

### Reglas de dependencia (`docs/ROADMAP.md` §2, constitución IV)

| Regla | Cómo la cumple |
|---|---|
| R1 · `internal/core/**` no importa adaptadores | Sin cambios en `internal/core`. |
| R2 · Solo `internal/httpx` importa `net/http` | Ningún fichero nuevo lo importa: `tanda.go` ejecuta `gh` con `os/exec` (research D16). |
| R3 · Solo `cache`/`store`/`graph` importan SQLite y `database/sql` | `internal/evals` no los importa; el grafo previo se prepara con el binario en proceso, como hoy. |
| R4 · Solo `cli` y `cmd/` llaman a `os.Exit` | Ningún `os.Exit` nuevo: los puntos de entrada son tests; el guion del sondeo sale con su código. |
| R5 · Solo `render` escribe en stdout | Ningún `fmt.Print*`, `os.Stdout` ni `os.Stderr` nuevo en Go: `TestTandaDelCommit` escribe en el fichero de `-salida` y registra con `t.Log`; `TestSondeo`, en `uso.txt` o `salida.txt`, que imprime el guion. |
| R6 · `internal/graph` no importa `source/*` ni `render` | Sin cambios. |

## Project Structure

### Documentation (this feature)

```text
specs/014-h7-4-boe-legislacion-sin/
├── plan.md                  # este fichero
├── research.md              # V1-V27, O1-O2, S1-S9, causa de raíz, calibrado, D1-D21
├── data-model.md            # lista, eval, sesión, juicio, recuento, umbrales, tanda, error de uso, ficheros
├── quickstart.md            # §1-§9; §8, el sondeo; §9, el job de cierre
├── contracts/
│   ├── skill-boe-legislacion.md   # v0.1.4: C1-C10, lo que se queda, PowerShell, uso, CHANGELOG
│   ├── lista-de-expresiones.md    # familias, formas fijas, esquema, calibrado, formas de la skill, frases de Juzgar
│   ├── evals-y-juicio.md          # no_se_activan, redacciones_modificadas, eval 20, grafo previo, legal-core, LeerSesion
│   ├── informe-del-job.md         # qué se cuenta, umbrales, motivos, informe.md, sondeo, tests
│   ├── tanda-del-job.md           # evals.yml, decidirLaTanda, TestTandaDelCommit, TestDefinicionDelJob, uso
│   └── sondeo.md                  # errorDeUso, uso.txt, el guion, tests
├── checklists/              # del spec
└── tasks.md                 # la escribe /speckit-tasks
```

### Source Code (repository root)

```text
schemas/expresiones-prohibidas.yaml.json            # redaccion_no_leida, formas_fijas [datos]
schemas/eval.yaml.json                              # no_se_activan, redacciones_modificadas [datos]
evals/boe-legislacion/expresiones-prohibidas.yaml   # anuncio +25, redaccion_no_leida, formas_fijas
evals/boe-legislacion/20-lcsp-dos-bloques-redaccion-cambiada.yaml   # nueva [datos]
evals/legal-core/01-…yaml, 02-…yaml, 03-…yaml       # no_se_activan: [boe-legislacion]
testdata/evals/grafo-previo/lcsp-a1-30-y-da-3-redaccion-original/   # dos derivadas [datos]
skills/boe-legislacion/SKILL.md                     # v0.1.4
.github/workflows/evals.yml                         # trabajo tanda; evals: needs, if, timeout-minutes 122
scripts/evals-sondeo.sh                             # uso.txt
internal/app/grafo_test.go                          # derivadasDelGrafoPrevio: dos entradas
internal/evals/
├── prohibidas.go            # RedaccionNoLeida, FormasFijas; quitar las formas; esDeLaClaseB
├── redacciones.go           # nuevo: ExtraerRedaccionesModificadas (sus casos, en los de la eval 20 de TestJuzgar…)
├── formato.go               # Eval.NoSeActivan, Eval.RedaccionesModificadas, RedaccionEsperada
├── juzgar.go                # no_se_activan, redacciones esperadas, campos del resultado
├── sesion.go                # la respuesta del primer result
├── informe.go               # recuento solo de terminadas (+ sin activar, clase B), formas exigidas, informe.md
├── umbrales.go              # sin_activar y redaccion_no_leida
├── tanda.go                 # nuevo: decidirLaTanda, esperarLaDecision, lectura del JSON de gh, órdenes de gh
├── definicion.go            # lee jobs.tanda y needs/if de jobs.evals
├── sondeo.go                # errorDeUso
├── job_test.go              # TestTandaDelCommit (evals); TestSondeo con uso.txt; banderas -ejecucion, -salida
├── testdata/sesiones/leer-sesion/respuesta-antes-de-una-tarea-en-segundo-plano/   # [datos]
├── doc.go
└── conjunto_test.go, formato_test.go, prohibidas_test.go, juzgar_test.go, informe_test.go, umbrales_test.go,
    sesion_test.go, definicion_test.go, sondeo_test.go
CHANGELOG.md, CONTRIBUTING.md
```

**Structure Decision**: la de `docs/ROADMAP.md` §2 y `CLAUDE.md`, sin paquetes nuevos: el formato de eval, el juicio,
el informe, la decisión de la tanda y el sondeo en `internal/evals`; sus guiones en `scripts/`; la definición del job en
`.github/workflows/`; la skill en `skills/`; los datos de eval en `evals/` y `testdata/evals/`. Ningún fichero de
producto del binario cambia salvo `SKILL.md`, que el binario empotra.

## Aceptación e2e

**Aceptación e2e: no aplica.** El hito no cambia el binario —cambiarlo está fuera de alcance, y `cmd/kitlegal` no
enlaza `internal/evals`—, así que no hay comportamiento de `kitlegal` que un guion `testscript` describa (research D20).
Su aceptación es la de la skill y la del instrumento: el job de evals de cierre (SC-001: los cuatro umbrales con
`decide: true` y `cumple: true`, la 04 pasando, `legal-core` aprobado, la eval 20 con sus dos líneas y su tasa, veredicto
aprobado y `red` vacío). En `make ci` la fijan, sin modelo: por US1, `TestEvalsDelRepositorio` (`expresiones-calibradas`,
`expresiones-en-los-bloques`, `expresiones-de-la-skill`, `prosa-de-la-skill`), `TestJuzgarLasClasesDeLaRespuesta` y
`TestExtraerExpresionesProhibidas` (FR-010 a FR-012, FR-021, FR-030 a FR-034, FR-093, FR-094; SC-002 a SC-004, SC-007);
por US2, `TestUmbralesDelInforme` y los casos de `no_se_activan` de `TestJuzgar…` (FR-001 a FR-004, FR-041; SC-006,
SC-007); por US3, los casos de la eval 20 de `TestJuzgar…`, `TestOrdenesParaPowerShell`,
`TestEvalsDelRepositorio/ordenes-para-powershell` y `grafo-previo`, y `TestGrabacionesDerivadas` (FR-023, FR-024,
FR-050 a FR-055, FR-095, FR-096; SC-005, SC-012); por US4, `TestUmbralesDelInforme` (FR-040 a FR-048, FR-097; SC-006);
por US5, `TestLeerSesion` y el caso `tres-de-54-y-seis-sin-terminar` (FR-060 a FR-062, FR-098; SC-008); por US6,
`TestDefinicionDelJob` (FR-070 a FR-072, FR-100; SC-010); por US7, `TestGuionDelSondeo`, `TestComprobarElSondeo` y
`TestSondear` (FR-080 a FR-082, FR-099; SC-009). Ninguna tarea `[aceptacion]`.

## Controles mecánicos que este hito añade o toca

### Objetivos del `Makefile`

Ninguno nuevo ni cambiado. Con más dentro: `test` (los tests nuevos de `internal/evals` e `internal/app`),
`skills-check` (las subpruebas nuevas y cambiadas de `TestEvalsDelRepositorio`, el tamaño y el frontmatter de
`SKILL.md`). `schema-check` no cambia: los dos esquemas que se tocan no salen de `--describe`.

### Tests nuevos

| Test | Fichero | Cubre |
|---|---|---|
| `TestJuzgarLasClasesDeLaRespuesta` | `juzgar_test.go` | FR-010 a FR-012, FR-022, FR-031, FR-034, FR-094; SC-007 |
| Casos de `no_se_activan` y de la eval 20 en `TestJuzgar…` | `juzgar_test.go` | FR-003, FR-004, FR-052, FR-053, FR-094 |
| `TestOrdenesParaPowerShell` y la subprueba `ordenes-para-powershell` | `conjunto_test.go` | FR-024, FR-095; SC-005 |
| Subpruebas `segundo-disparo` y `estado-de-la-tanda` de `TestDefinicionDelJob` (la decisión y la lectura del JSON de `gh`, sin tests aparte) | `definicion_test.go` | FR-070, FR-071, FR-100; SC-010 |
| Casos de §6 de [contracts/informe-del-job.md](./contracts/informe-del-job.md) en `TestUmbralesDelInforme` | `umbrales_test.go` | FR-040 a FR-046, FR-061, FR-097; SC-006 |
| Caso `respuesta-antes-de-una-tarea-en-segundo-plano` de `TestLeerSesion` | `sesion_test.go` | FR-060, FR-098; SC-008 |
| Casos de `uso.txt` en `TestGuionDelSondeo`; `errorDeUso` en `TestComprobarElSondeo` y `TestSondear` | `sondeo_test.go` | FR-080, FR-081, FR-099; SC-009 |

### Tests existentes que cambian

- `TestEvalsDelRepositorio`: `expresiones-calibradas` con tres informes y cinco columnas (contracts/lista-de-expresiones.md
  §4; SC-002); `listaDelRepositorio` exige también `redaccion_no_leida` y `formas_fijas`; `expresiones-de-la-skill` con
  la respuesta compuesta y las formas fijas (§5; SC-003); `grafo-previo`, con la comprobación de las redacciones
  esperadas (contracts/evals-y-juicio.md §4).
- `TestDefinicionDelJob`: `del-repositorio` y `sinteticas` con las claves de `tanda` y de `evals`, y el tope de 122
  (contracts/tanda-del-job.md §4).
- `TestGrabacionesDerivadas` (`internal/app/grafo_test.go`): dos entradas más (FR-096; SC-012).
- `formato_test.go`: los casos de los dos esquemas (contracts/lista-de-expresiones.md §2, contracts/evals-y-juicio.md
  §1). `TestLeerConjunto`, si arma la lista del repositorio.
- Todo test cuyo texto sintético lleve una forma nueva y que aplique la lista del repositorio o una con esas formas
  (research V27): hoy, el caso de `juzgar_test.go:642` («La redacción ha cambiado desde la consulta anterior.», que gana
  `consulta anterior` en sus expresiones y motivos esperados), el de `prohibidas_test.go:174` (la frase de la regla 7,
  que con una lista con `formas_fijas` no se marca) y dos casos de `TestCondicionesDeLaConsultaRepetida`
  (`consulta_repetida_test.go:73-74` y `:90-91`, la línea escrita con otras palabras o con fechas de nueve cifras, que
  ganan la línea de `se consultó antes` en las esperadas). La línea sin cita entera (`juzgar_test.go:111-112` y
  `:136-137`, `prohibidas_test.go:156-157`) no cambia: la quita su forma fija. La tarea que cambia la lista busca antes,
  en `internal/evals/*_test.go`, las 35 formas nuevas, escritas en UTF-8 y con escapes `\x` (la orden de research V27), y
  declara cada fichero que las lleva.
- `informe_test.go`, `juzgar_test.go`, `sesion_test.go`, `umbrales_test.go`: los resultados e informes esperados ganan
  `redacciones_modificadas_encontradas`/`…_ausentes`, los umbrales nuevos donde hay lista, las columnas de
  `informe.md` y las respuestas medidas sin las sesiones sin terminar. Ninguno se desactiva ni se salta.

### Puntos de entrada fuera de `make ci` (etiqueta `evals`)

`TestTandaDelCommit` (nuevo; lo ejecuta el trabajo `tanda` del job), `TestEjecucionDelJob` (sin cambios), `TestSondeo`
(escribe `uso.txt` con un `errorDeUso`) y `TestComprobarConsultaRepetida` (sin cambios). `golangci-lint` los lintea.

### Fixtures, `testdata/` y `schemas/` (tareas `[datos]`)

- `schemas/expresiones-prohibidas.yaml.json` y `schemas/eval.yaml.json` (contratos de la lista §2 y de las evals §1),
  con los casos de `formato_test.go` que los fijan y los ficheros que los cumplen.
- `testdata/evals/grafo-previo/lcsp-a1-30-y-da-3-redaccion-original/` (dos derivadas, escritas por
  `TestGrabacionesDerivadas -actualizar-derivadas`) con la eval 20 que las usa.
- `internal/evals/testdata/sesiones/leer-sesion/respuesta-antes-de-una-tarea-en-segundo-plano/` (transcript sintético,
  `codigo-de-la-sesion` y `sesion.err`).

### CI

`.github/workflows/evals.yml`: el trabajo `tanda` (su paso `decidir` ejecuta `go test` con `-timeout 12m`, para que la
espera de 10 min llegue a medir; supuesto T008), `needs`, `if` y `timeout-minutes` de `evals` (contracts/tanda-del-job.md
§1). Lo comprueba `TestDefinicionDelJob` en `make ci`. `ci.yml` no cambia. El job de evals se ejecuta en la propuesta de
cambio, tras la revisión final, y lo lanza el workflow (FR-102).

## Controles de umbral

Las cuatro filas del cierre que fija FR-092, y, como pide el ADR 0029 (criterio o), una por cada umbral que se mide en
`make ci` o en el cierre, con el test o la comprobación que lo pone en rojo. El de las expresiones del modelo informativo
(FR-044) no tiene fila: solo se publica.

| Requisito | Umbral | Control | Dónde |
|---|---|---|---|
| FR-040, SC-001 (expresiones, las dos clases) | ≤ 5 % de las respuestas medidas del modelo que decide en evals que activan la skill (≤ 2 de 54) | umbral del informe con `decide: true`: por encima, motivo, veredicto `fallo` y el trabajo en rojo, que el cierre cuenta | `evals:boe-legislacion:expresiones_prohibidas:claude-sonnet-5-5` |
| FR-041, SC-001 (sin activar) | 0 respuestas del modelo que decide sin la skill activada en evals que la activan | ídem | `evals:boe-legislacion:sin_activar:claude-sonnet-5-5` |
| FR-042, SC-001 (clase B) | 0 respuestas del modelo que decide con una expresión de `redaccion_no_leida` | ídem | `evals:boe-legislacion:redaccion_no_leida:claude-sonnet-5-5` |
| FR-043, SC-001 (duración) | ≤ 900 s de sesiones | umbral del informe con `decide: true`: por encima, motivo de la ejecución y `fallo` | `evals:boe-legislacion:duracion_de_las_sesiones` |
| SC-001 (ninguna sin medir, H7.3 FR 043) | 0 sesiones sin medir por un límite de uso | con alguna, motivo propio y `fallo` en el job; el test lo ve fallar con sesiones sintéticas | `ci:internal/evals/informe_test.go:TestInformeConSesionesSinMedir` |
| FR-032, FR-093, SC-002 | exactamente 36, 11 y 9 de 93, eval por eval y familia por familia | subprueba `expresiones-calibradas` | `ci:internal/evals/conjunto_test.go:TestEvalsDelRepositorio` |
| FR-033, FR-093, SC-003 | 0 expresiones en los bloques grabados y en la respuesta hecha de las formas de la skill | subpruebas `expresiones-en-los-bloques` y `expresiones-de-la-skill` | `ci:internal/evals/conjunto_test.go:TestEvalsDelRepositorio` |
| FR-021, FR-093, SC-004 | 0 expresiones en la prosa de `SKILL.md` y 0 fechas `AAAAMMDD` con cifras | subprueba `prosa-de-la-skill` | `ci:internal/evals/conjunto_test.go:TestEvalsDelRepositorio` |
| FR-024, FR-095, SC-005 | 2 de 2 órdenes de lectura y comprobación con su forma para PowerShell, misma norma y mismos bloques | subprueba `ordenes-para-powershell` y `TestOrdenesParaPowerShell` | `ci:internal/evals/conjunto_test.go:TestEvalsDelRepositorio`, `ci:internal/evals/conjunto_test.go:TestOrdenesParaPowerShell` |
| FR-041, FR-042, FR-045, FR-097, SC-006 | 1 sin activar o 1 con la clase B → `fallo`, 0 → se cumple; 3 de 54 → `fallo`, 2 de 54 → se cumple; con seis sin terminar, total 54 | casos sintéticos del informe | `ci:internal/evals/umbrales_test.go:TestUmbralesDelInforme` |
| FR-034, FR-094, SC-007 | cada frase marcada en su clase; 0 marcas en las formas fijas, en «No hay avisos de vigencia sobre este bloque» y en la respuesta de FR-022 | tabla de frases | `ci:internal/evals/juzgar_test.go:TestJuzgarLasClasesDeLaRespuesta` |
| FR-080, FR-099, SC-009 | 0 líneas de `go test` o de testify en la salida de error de un error de uso; 0 sesiones abiertas | el guion con el `go` sustituto; `sondear` sin preparar el árbol | `ci:internal/evals/sondeo_test.go:TestGuionDelSondeo`, `ci:internal/evals/sondeo_test.go:TestSondear` |
| FR-070, FR-071, FR-100, SC-010 | ≤ 1 tanda a la vez por commit y skill; 0 comprobaciones rojas o `evals (<skill>)` de más del segundo disparo; la etiqueta tras una tanda terminada vuelve a medir | la definición y `segundo-disparo` | `ci:internal/evals/definicion_test.go:TestDefinicionDelJob` |
| FR-054 (tope del trabajo) | `timeout-minutes` ≥ el peor caso con las evals del repositorio (7 285 s) | `fallosDelTope` | `ci:internal/evals/definicion_test.go:TestDefinicionDelJob` |
| FR-026, SC-011 | `SKILL.md` < 300 líneas; `description` ≤ 1024 caracteres | `TestSkillsDelRepositorio` en `skills-check` | `ci:Makefile:skills-check` |

SC-008 y SC-012 no tienen un umbral numérico (su control es el test de la tabla de «Tests nuevos» y
`TestGrabacionesDerivadas`).

## Uso, de fuera adentro (criterio de uso, ADR 0028)

Detalle, con ejemplos y bytes, en contracts/skill-boe-legislacion.md §5, contracts/informe-del-job.md §2,
contracts/tanda-del-job.md §5 y contracts/sondeo.md §3. Resumen:

- **La respuesta de la skill** (la persona; una por pregunta): lo que ocupa la norma y 0 líneas sobre la comprobación;
  221-242 B por bloque cuya redacción cambió, con su cita (0 en la mayoría de las preguntas, 2 en la de la eval 20, como
  mucho k × ≈ 240 B con k bloques leídos), con cientos de normas y miles de bloques consultados igual que con uno. Cada
  orden de lectura lleva detrás una `graph check` con sus bloques (≈ 325 B sin nada, ≈ 1 000 B por bloque cambiado,
  ≤ 3 800 B con cinco; H7.1, H7.3). La línea se apaga con la lectura siguiente del bloque (H7.1).
- **`umbrales`** (el job, el informe final, la persona; una vez por job y skill): cinco elementos de ≈ 250 B en
  `boe-legislacion` (≈ 1,2 KB), `[]` en `legal-core`; se miden de nuevo en cada job.
- **Las líneas esperadas de la eval 20** en el informe (la persona; una vez por job): 2 por sesión en 3 sesiones, < 1 KB.
- **La comprobación `tanda`** (el cierre; una por disparo que quiere medir): termina en segundos; sus consultas a `gh`,
  acotadas a las ejecuciones del commit (`--commit`, `--limit 100`).
- **La salida de error del sondeo** en un error de uso (quien lo lanza; una por lanzamiento): una línea por error, como
  mucho seis, < 1 KB.
- **`SKILL.md`** (el modelo; una vez por conversación): 295 líneas; **la lista**: 87 expresiones y 2 formas fijas,
  ≈ 2,9 KB con su comentario de cabecera (2 966 B), fija.

## Decisiones

- **La activación se corrige en la `description`**, con la razón y no con una prohibición, y sin tocar la 04 (research
  D1).
- **La clase A va a `anuncio`; la B, a una familia nueva `redaccion_no_leida`**, y la lista no gana las formas que
  nombran la redacción anterior, para no marcar la respuesta honrada de FR-022 (D2, D6).
- **Las formas fijas las declara la lista** (`formas_fijas`, con `<cita>` y `<fecha>`), y `expresiones-de-la-skill`
  comprueba que la lista y la skill dicen lo mismo (D7, D8). `<cita>` es opcional: la línea sin cita de los informes del
  calibrado también se quita, y `se consultó antes`, en `anuncio`, marca esas palabras fuera de la línea (D6, D7).
- **La línea `⚠ REDACCIÓN MODIFICADA:` lleva la cita detrás de la etiqueta** (D4); **PowerShell**,
  `…; if ($LASTEXITCODE -eq 0) { … }` (D5).
- **La respuesta a la pregunta es el primer `result`** (D12).
- **Una tanda por commit con un trabajo `tanda` que decide antes de `evals`** y una marca por paso que leen los
  disparos posteriores; solo se espera a que decidan las anteriores, nunca a su tanda (D15, D16).
- Las demás, con su alternativa rechazada, en research D1-D21.

## Trazabilidad: cada mecanismo y su requisito

| Mecanismo | Requisito |
|---|---|
| `description` de v0.1.4 (C1) | FR-001 a FR-003, FR-026 |
| Prosa sin el vocabulario de la clase A (C2-C8, C10) | FR-010, FR-020, FR-021, FR-025 |
| «Redacción modificada»: por qué, qué sí, qué responder (C9) | FR-011, FR-020, FR-022 |
| Línea `⚠ REDACCIÓN MODIFICADA:` con la cita (C9) | FR-023 |
| Formas para PowerShell (C2, C4) y `defectosDeLasOrdenesParaPowerShell` (en `conjunto_test.go`) | FR-024, FR-095 |
| `anuncio` +25, `RedaccionNoLeida`, `esDeLaClaseB` | FR-010, FR-011, FR-030, FR-031, FR-036, FR-042 |
| `FormasFijas` y su supresión en `ExtraerExpresionesProhibidas` (con las etiquetas de aviso; `<cita>` opcional) | FR-012, FR-031, FR-032, FR-033 |
| Esquema de la lista | FR-035 |
| Calibrado de tres informes y cinco columnas | FR-032, FR-093 |
| `expresiones-de-la-skill` con la respuesta compuesta | FR-033, FR-093 |
| `Eval.NoSeActivan` y su juicio; las tres de `legal-core` | FR-003, FR-004 |
| `Eval.RedaccionesModificadas`, `ExtraerRedaccionesModificadas`, sus campos, motivos y formas exigidas | FR-023, FR-052, FR-053 |
| Esquema de eval | FR-003, FR-053 |
| Eval 20 y su grafo previo; dos entradas de `derivadasDelGrafoPrevio`; la comprobación en `compruebaLaSesion` | FR-050 a FR-055, FR-096 |
| `Sesion.Respuesta` del primer `result`; el caso de `TestLeerSesion` | FR-060, FR-098 |
| Recuento solo de terminadas, con sin activar y clase B | FR-045, FR-061, FR-082 |
| Umbrales `sin_activar` y `redaccion_no_leida`; sus motivos | FR-040 a FR-048 |
| Trabajo `tanda`, `decidirLaTanda`, `esperarLaDecision`, `TestTandaDelCommit`, marca por paso | FR-070 a FR-072 |
| `TestDefinicionDelJob` con la tanda y `segundo-disparo` | FR-100 |
| `timeout-minutes: 122` | FR-054 (H7.3 FR 035) |
| `errorDeUso`, `uso.txt`, el guion | FR-080 a FR-082, FR-099 |
| `CHANGELOG.md`; `CONTRIBUTING.md`, `doc.go` y comentarios de `evals.yml` | FR-027, FR-101; lo que el hito deja falso (research D19) |

Nada del diseño atiende a un estado que no pasa el umbral de materialidad: un transcript, una respuesta de `gh`, una
definición o una lista que no se pueden leer siguen las reglas que ya hay —sesión ilegible, fichero mal formado, test que
falla, error con código 1—, sin caso propio.

## Cambios de `SKILL.md` trazados a la causa de raíz (FR-002, FR-020)

C1-C10 de contracts/skill-boe-legislacion.md §1, cada uno con el texto de v0.1.3, el de v0.1.4, la causa medida (research,
«Causa de raíz») y su requisito. En resumen: la activación (C1: la skill como comprobación opcional; la 04 pide un dato);
«la comprobación» y «desde una lectura anterior» (C2, C3, C5, C8, C10); «si lo leído basta» (C6); «si la redacción
cambió» (C7); y la clase B sin razón ni respuesta (C9).

## Datos externos

Ninguno (research D21): ni fuente ni grabación nuevas; las respuestas de `a1-30` y `da-3` están desde H4 y las derivadas
las produce `TestGrabacionesDerivadas -actualizar-derivadas`. Ningún manifiesto `grabaciones.json` ni test `TestGrabar*`
cambia, y el paso `grabar_datos` no tiene nada que grabar.

## Orden de implementación (de dentro afuera)

1. **`[datos]`** `schemas/expresiones-prohibidas.yaml.json` y `schemas/eval.yaml.json`, con sus casos de
   `formato_test.go`, `ExpresionesProhibidas` (`RedaccionNoLeida`, `FormasFijas`, la supresión y `esDeLaClaseB`),
   `Eval` (`NoSeActivan`, `RedaccionesModificadas`) y las tres evals de `legal-core` con `no_se_activan`. La lista del
   repositorio gana en esta tarea sus `formas_fijas` (las dos, definitivas) y la primera forma de `redaccion_no_leida`
   (`ya no exige`), que no marca ninguna respuesta de los informes de H7.1 y H7.2 (research V20): el esquema se cumple y
   el calibrado de hoy no cambia. Las otras nueve formas de `redaccion_no_leida` y las 25 de `anuncio` entran en el paso
   6, con `SKILL.md` v0.1.4: con ellas, la prosa de v0.1.3 da 9 defectos (research V19).
2. `redacciones.go` y el juicio de `Juzgar` (`no_se_activan`, redacciones esperadas), con sus tests.
3. `sesion.go` (el primer `result`) con su caso **`[datos]`** de `TestLeerSesion`.
4. `informe.go` y `umbrales.go` (recuento solo de terminadas, los dos umbrales, las columnas y las formas exigidas de
   `informe.md`), con sus tests.
5. **`[datos]`** la eval 20, las dos entradas de `derivadasDelGrafoPrevio`, las derivadas escritas por
   `TestGrabacionesDerivadas -actualizar-derivadas` y la comprobación de `compruebaLaSesion`; y `timeout-minutes: 122`
   en `evals.yml` (sin él, `TestDefinicionDelJob` falla con 20 evals).
6. `skills/boe-legislacion/SKILL.md` v0.1.4 con la lista entera (`anuncio` +25 y `redaccion_no_leida` con sus 10), el
   calibrado de tres informes y cinco columnas, `expresiones-de-la-skill`, `ordenes-para-powershell` y
   `TestOrdenesParaPowerShell`, `TestJuzgarLasClasesDeLaRespuesta` y los tests con textos que llevan formas nuevas, en
   la misma tarea: con la lista nueva, `prosa-de-la-skill` da 9 defectos en v0.1.3 (research V19).
7. `tanda.go`, `TestTandaDelCommit`, el trabajo `tanda` y `needs`/`if` de `evals` en `evals.yml`, y
   `TestDefinicionDelJob` con la tanda.
8. `errorDeUso`, `TestSondeo`, `scripts/evals-sondeo.sh` y sus tests.
9. `CHANGELOG.md`, `CONTRIBUTING.md`, `internal/evals/doc.go` y los comentarios de `evals.yml`.

**Obligaciones para `tasks.md`**: ninguna tarea `[aceptacion]` («Aceptación e2e: no aplica»); cada tarea deja `make ci`
en verde; las tareas `[datos]` son las de los pasos 1, 3 y 5, con las rutas de este plan; cada fila de «Controles de
umbral» tiene la tarea que construye su control, con un test que lo ve fallar; ninguna tarea ni corrección cumple un
umbral de FR-040 a FR-043 rebajándolo, dejándolo en `decide: false`, sacando evals del total o recortando la lista
(FR-047), ni cambia la regla por serie; ninguna tarea edita `specs/011-…` a `specs/013-…` ni `scripts/workflow/`
(FR-072, FR-090); ninguna tarea usa la red, graba, abre una sesión con modelo, ejecuta `make evals`, `make evals-sondeo`
con modelo o `TestTandaDelCommit`, publica ni mide en la plataforma; el quickstart §8 con modelo lo ejecuta la persona al
leer el informe final.

## Complexity Tracking

| Desviación | Por qué hace falta | Alternativa más simple rechazada |
|---|---|---|
| «Aceptación e2e: no aplica» (principio III: «cada hito empieza por el test e2e») | El hito no cambia el binario: no hay comportamiento suyo que un guion describa. La aceptación es la del job (SC-001), y cada pieza entra con sus tests en `make ci` | Un guion que ejerciera el binario sin cambios: no describiría la entrega |
| Tareas `[datos]` con código y tests (esquemas, tipos, casos que los fijan y ficheros que los cumplen) | Con las claves obligatorias en el esquema, la lista sin ellas es un fichero mal formado: separarlos deja `make ci` en rojo entre dos tareas. Precedente: los esquemas de la lista en H7.2 y H7.3 | Separar esquema, tipo y datos |
| Un trabajo más en el flujo `evals`, con Go, antes de las sesiones | La decisión de la tanda se prueba en `make ci` solo si es Go (sin `jq`, que no es del stack) y tiene que tomarse antes de `evals` para saltarlo entero (research D15) | La decisión dentro de `evals` o en bash con `jq` |

## Comprobación contra la rúbrica del juez (`juez_plan`) y `precheck.sh plan`

- `precheck.sh plan`: existen `plan.md` y `research.md`; ninguno conserva marcas de aclaración pendiente; están
  `## Constitution Check`, la línea «Aceptación e2e» y `## Controles de umbral`, con cada fila nombrando requisitos del
  spec y un control con forma (`evals:<skill>:<nombre>` o `ci:<ruta>[:<Test>]`).
- a · Constitution Check: un ítem por principio (I-IX) y por regla de dependencia (R1-R6).
- b · Dependencias: ninguna nueva.
- c · Reglas de dependencia: tabla R1-R6; ni `net/http`, ni SQLite, ni `os.Exit`, ni salida estándar nuevos.
- d · Errores y códigos: ningún código de `kitlegal` nuevo; errores de `internal/evals` con su fichero, orden o
  argumento; `errorDeUso` tipado; los guiones salen con 1 (research D16, D18).
- e · Tests primero: «Aceptación e2e: no aplica» con el motivo; «Tests nuevos», «Tests existentes que cambian», puntos de
  entrada, fixtures y las tareas `[datos]`.
- f · Alcance: nada fuera del spec; la documentación que se toca es la que el hito deja falsa (research D19).
- g · Sin atajos: ningún `//nolint`, `t.Skip`, TODO ni error silenciado previstos (G204, con órdenes constantes, D16).
- h · Mejor alternativa: cada decisión con la rechazada (D1-D21).
- i · Afirmaciones verificadas: V1-V27 con fichero, línea u orden; O1-O2 observadas; S1-S9 como supuestos.
- j · Quickstart ejecutable: órdenes y nombres de test reales (los que fija el plan); lo que escribe cada escenario,
  declarado; nada versionado cambia.
- k · Datos externos: ninguno.
- l · Autonomía: ninguna tarea para una persona; el §8 con modelo y el §9 los ejecuta la persona o el workflow al
  cerrar; el cierre en la plataforma lo hace el workflow.
- m · Uso: «Uso, de fuera adentro» y los contratos (bytes, órdenes por pregunta, apagado de cada señal).
- n · Proporcionalidad: «Trazabilidad»; sin mecanismos para estados sin vía real.
- o · Controles de umbral: las cuatro filas del cierre de FR-092, sin la del informativo, y una por cada umbral medido en
  `make ci`.
