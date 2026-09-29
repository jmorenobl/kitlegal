# Implementation Plan: H7.3 · El umbral de expresiones prohibidas decide, `boe-legislacion` sin el vocabulario que provoca el ruido, y evals en paralelo con un sondeo local

**Branch**: `013-h7-3-el-umbral-de` | **Date**: 2026-09-30 | **Spec**: [spec.md](./spec.md)

**Input**: Feature specification from `/specs/013-h7-3-el-umbral-de/spec.md`

**Modo**: desatendido. Las decisiones técnicas se tomaron con el «Criterio de decisión autónoma» de la constitución y
están en [research.md](./research.md) (D1-D22), cada una con su alternativa rechazada. Toda afirmación sobre Claude Code,
GitHub Actions, `gh`, `gosec`, Go o el repositorio remite a la tabla V de research.md, comprobada en local y sin red (la
2.1.270 del job, en la imagen Docker local de H5); lo que no se puede comprobar así son los supuestos S1-S7.

## Summary

H7.3 no toca el binario. Arregla `boe-legislacion`, hace cumplir su umbral y hace barato medirla, en seis piezas:

1. **El umbral decide** (research D5): `informe.json` lleva siempre `umbrales` (contrato del ADR 0029). En
   `boe-legislacion`, `expresiones_prohibidas:claude-sonnet-5` (≤ 0,05, `decide: true`), el de Haiku 4.5 (`decide:
   false`) y `duracion_de_las_sesiones` (≤ 900 s, `decide: true`); en `legal-core`, `[]`. Un umbral que decide y no se
   cumple es un motivo, así que el veredicto es `fallo` y el job sale en rojo.
2. **`boe-legislacion` v0.1.3** (research «Causa de raíz», D1): el ruido de H7.2 repite la prosa de `SKILL.md` —la
   regla 7 empareja «con hallazgos o sin ellos» con «trasládalos»; el paso 5 dice «antes de redactar la respuesta» y
   enumera lo que la respuesta no nombra; el ejemplo trae 20250101—. Once cambios (C1-C11) quitan ese vocabulario de la
   prosa, reescriben la regla 7 y ponen `AAAAMMDD` en el ejemplo. La comprobación no cambia de sitio (§ Decisiones).
3. **La tercera familia** (D2, D3): `anuncio`, 14 expresiones medidas en los cierres de H7.1 y H7.2; con ella la lista
   marca exactamente 35 y 10 respuestas, eval por eval.
4. **El job en paralelo** (D7-D15, D19): un repartidor en Go, compartido con el sondeo, abre como mucho
   `CONCURRENCIA_DE_EVALS` sesiones a la vez (4 y 1), cada una con su estado de Claude Code, y las cierra con el tope de
   siempre (240 s, 124/137); un límite de la cuenta deja la sesión **sin medir**, con un motivo propio, y
   tras el mensaje del límite de uso no abre más; el segundo disparo sobre el mismo commit espera por su grupo de
   concurrencia; y `TestDefinicionDelJob` comprueba en `make ci` la definición, con el tope que cubre el peor caso.
5. **El sondeo** (D16-D18): `make evals-sondeo SKILL=… EVALS=… MODELO=… REPETICIONES=… [CONCURRENCIA=…]`, en macOS o Linux,
   sin strace ni sudo, con el binario y la skill del árbol, sin la configuración de Claude Code de quien lo lanza y con
   solo `CLAUDE_CODE_OAUTH_TOKEN`; el mismo repartidor, la misma preparación y el mismo juez sin lo que sale de la traza;
   una salida que empieza diciendo que no es un veredicto.
6. **El escenario del quickstart** (§6): el sondeo con la `SKILL.md` de `main` y con la del hito.

No cambia ninguna decisión de arquitectura: no hay ADR nuevo, y `specs/012-…` no se edita (FR-080).

## Technical Context

**Language/Version**: Go 1.27 (`go 1.27.0`, `toolchain go1.27.1`), `CGO_ENABLED=0`, `-trimpath`; bash (los guiones nuevos,
sin nada posterior a 3.2).

**Primary Dependencies**: las fijadas, sin ninguna nueva: `go.yaml.in/yaml/v3` (la del repositorio para yaml.v3; ya la
usa `release_test.go`), `santhosh-tekuri/jsonschema` v6 (la lista y su esquema), `stretchr/testify`; biblioteca
estándar (`os/exec`, `syscall`, `os/signal`, `encoding/json/v2`, `regexp`).

**Storage**: ninguno nuevo. Ficheros del repositorio: la lista (`evals/boe-legislacion/`), su esquema (`schemas/`), la
definición del job (`.github/workflows/evals.yml`), la skill (`skills/`), dos guiones (`scripts/`). Las sesiones, en
temporales.

**Testing**: `go test -race` en `make ci` (unitarios con tabla y `t.Parallel`; sustitutos POSIX de `claude`, `strace` y
`go` escritos por los tests en `t.TempDir()`; `TestPrepararElArbolDelSondeo` con la etiqueta `integration`); puntos de
entrada con la etiqueta `evals` fuera de `make ci`; el job de evals en la propuesta de cambio.

**Target Platform**: el job, `ubuntu-24.04` con strace; el sondeo y los tests del repartidor, macOS y Linux. El binario
distribuido no cambia (`internal/evals` no llega a él).

**Project Type**: skill (`boe-legislacion`) + formato común de eval, job y sondeo (`internal/evals`, `scripts/`, el flujo
`evals`).

**Performance Goals**: sesiones de `boe-legislacion` ≤ 900 s en el job (se esperan ≈ 520 s con 4 a la vez, D14); peor
caso de cada trabajo ≤ 120 min (D13).

**Constraints**: ningún test ni tarea usa la red ni abre una sesión con modelo; `SKILL.md` < 300 líneas; el sondeo no
escribe fuera de su temporal salvo la caché de compilación de Go; ninguna variable de credencial de quien lanza el
sondeo llega a sus sesiones salvo `CLAUDE_CODE_OAUTH_TOKEN`.

**Scale/Scope**: 19 evals de `boe-legislacion` (94 sesiones con la prueba de red: 51 de Sonnet 5 y 30 de Haiku 4.5 en
evals que activan la skill) y 3 de `legal-core` (18 sesiones); una lista de 52 expresiones.

## Constitution Check

*GATE: pasado antes de la fase 0 y re-evaluado tras el diseño (fase 1): sin violaciones; las desviaciones justificadas
están en Complexity Tracking.*

### Principios

| Principio | Cómo lo cumple H7.3 |
|---|---|
| I · Fuentes públicas y frontera humana | No toca ninguna fuente, ni `internal/httpx`, ni graba nada (D22). Las sesiones del job y del sondeo siguen con el proxy que rechaza toda petición salvo la del modelo; ningún test ni tarea usa la red. |
| II · Nada sin cita ni fuente | El sobre no cambia. La skill sigue citando `[<identificador>, bloque <id>]`, trasladando los avisos con su forma y diciendo la redacción cambiada con las fechas que da el binario; el ejemplo pierde una fecha inventada (FR-013). |
| III · Tests primero y offline | «Aceptación e2e: no aplica» (abajo, con el motivo). Cada pieza entra con sus tests en `make ci`, offline, con sesiones y transcripts sintéticos y sustitutos de `claude`, `strace` y `go`; la aceptación es el job (SC-001) y el quickstart §6 (SC-002). Cobertura dentro de `codecov.yml`. |
| IV · Hexagonal y errores tipados | Ningún paquete del binario ni ningún código de salida de `kitlegal` cambian. `internal/evals` devuelve `error` con el fichero o el argumento delante; los guiones salen con 1 ante un error, como el job hoy; ningún `panic`. Ni `os.Exit` ni escritura en la salida estándar desde Go (V9): el guion del sondeo imprime su salida. |
| V · Simplicidad y dependencias | Ninguna dependencia nueva (`errgroup` rechazado, D7). Cada mecanismo se traza a un FR («Trazabilidad»); se retiran las tres entradas del job que ya nadie ejecuta (D19). |
| VI · Un binario, convenciones de agente | El binario, sus banderas y `--describe` no cambian; ningún ejecutable nuevo en `cmd/` (D16). |
| VII · Grafo y privacidad | Nada nuevo en el grafo; el de cada sesión vive en su temporal. El sondeo no lee ni pasa las credenciales de la persona salvo `CLAUDE_CODE_OAUTH_TOKEN`, ni su configuración de Claude Code (D10, D16). |
| VIII · Skills primero | Mejora `boe-legislacion` (v0.1.3) y la mide con sus evals, ahora con un umbral que decide (SC-001); protege el instrumento que la mide (job en paralelo, límites, sondeo). `SKILL.md` < 300 líneas, sin nombrar evals ni modelos. |
| IX · Genericidad territorial | Nada se particulariza para un municipio: ni la lista, ni la skill, ni el job; la concurrencia y el objetivo van por skill en la definición del job. |

### Reglas de dependencia (`docs/ROADMAP.md` §2, constitución IV)

| Regla | Cómo la cumple |
|---|---|
| R1 · `internal/core/**` no importa adaptadores | Sin cambios en `internal/core`. |
| R2 · Solo `internal/httpx` importa `net/http` | Ningún fichero nuevo lo importa: el repartidor ejecuta procesos (`os/exec`), no pide nada por red. |
| R3 · Solo `cache`/`store`/`graph` importan SQLite y `database/sql` | `internal/evals` no los importa; las sesiones usan el binario. |
| R4 · Solo `cli` y `cmd/` llaman a `os.Exit` | Ningún `os.Exit` nuevo: los puntos de entrada son tests que fallan; los sustitutos son guiones POSIX. |
| R5 · Solo `render` escribe en stdout | Ningún `fmt.Print*`, `os.Stdout` ni `os.Stderr` nuevo en Go: la salida del sondeo la imprime `scripts/evals-sondeo.sh` desde un fichero (D16). |
| R6 · `internal/graph` no importa `source/*` ni `render` | Sin cambios. |

## Project Structure

### Documentation (this feature)

```text
specs/013-h7-3-el-umbral-de/
├── plan.md                  # este fichero
├── research.md              # V1-V18, S1-S7, causa de raíz, D1-D22
├── data-model.md            # umbral, sesión leída, sin medir, informe, lista, definición, repartidor, sondeo
├── quickstart.md            # §1-§8; §6, el sondeo con modelo; §7, el job de cierre
├── contracts/
│   ├── lista-de-expresiones.md    # tercera familia, esquema, calibrado, comprobaciones
│   ├── skill-boe-legislacion.md   # v0.1.3: C1-C11, lo que se queda, la prosa, uso, CHANGELOG
│   ├── informe-del-job.md         # umbrales, motivos, sin medir, reintentos, duración, informe.md
│   ├── ejecucion-del-job.md       # evals.sh, evals-sesion.sh, repartidor, entorno, evals.yml, definición
│   └── sondeo.md                  # orden, guion, TestSondeo, salida, entorno, códigos
├── checklists/              # del spec
└── tasks.md                 # la escribe /speckit-tasks
```

### Source Code (repository root)

```text
schemas/expresiones-prohibidas.yaml.json            # anuncio [datos]
evals/boe-legislacion/expresiones-prohibidas.yaml   # anuncio
skills/boe-legislacion/SKILL.md                     # v0.1.3
.github/workflows/evals.yml                         # name, concurrency, include, env; cambios gana evals-sesion.sh
scripts/
├── evals.sh                 # CONCURRENCIA_DE_EVALS; TestEjecucionDelJob en lugar del bucle
├── evals-sesion.sh          # nuevo: la orden de la sesión, con strace si KITLEGAL_EVALS_TRAZA=si
└── evals-sondeo.sh          # nuevo: temporal, TestSondeo, salida
Makefile                     # objetivo evals-sondeo
internal/evals/
├── prohibidas.go            # Anuncio; ExtraerExpresionesProhibidas con tres familias
├── sesion.go                # api_retry, TerminaEnReintento, ErrorDelResultado, motivo con el texto
├── limites.go (+ _test)     # nuevo: clasificación (a) (b) (c)
├── umbrales.go (+ _test)    # nuevo: Umbral, construcción, motivos
├── juzgar.go                # ResultadoDeEval: ReintentosPorLimiteDeRitmo, SinMedir
├── informe.go               # Umbrales, SesionesSinMedir, reintentos, duración, SinAbrir, series sin medir, informe.md
├── sesiones.go (+ _test)    # nuevo: repartidor, entorno, tope, cancelación
├── definicion.go (+ _test)  # nuevo: leerDefinicionDelJob; TestDefinicionDelJob
├── sondeo.go (+ _test)      # nuevo: comprobar, árbol, sondear, juicio sin traza, salida; TestGuionDelSondeo
├── sondeo_integracion_test.go  # nuevo (integration): TestPrepararElArbolDelSondeo
├── sustitutos_test.go       # nuevo: los guiones sustitutos de claude, strace y go
├── job_test.go              # TestEjecucionDelJob, TestSondeo; fuera TestPlanDeSesiones, TestPrepararSesion, TestInformeDelJob
├── doc.go
└── conjunto_test.go, formato_test.go, prohibidas_test.go, juzgar_test.go, informe_test.go, sesion_test.go,
    consulta_repetida_test.go
CHANGELOG.md, CONTRIBUTING.md
```

**Structure Decision**: la de `docs/ROADMAP.md` §2 y `CLAUDE.md`, sin paquetes nuevos: el formato de eval, el job y el
sondeo en `internal/evals`; sus guiones en `scripts/`; la definición del job en `.github/workflows/`; la skill en
`skills/`; los datos de eval en `evals/`. Ningún fichero de producto del binario cambia.

## Aceptación e2e

**Aceptación e2e: no aplica.** El hito no cambia el binario —cambiarlo está fuera de alcance, y `cmd/kitlegal` no
enlaza `internal/evals` (`TestDependenciasDelBinario`)—, así que no hay comportamiento de `kitlegal` que un guion
`testscript` describa (research D21). Su aceptación es la de la skill y la del instrumento: el job de evals de cierre
(SC-001: `umbrales` con los dos que deciden cumplidos, ninguna sesión sin medir, veredicto aprobado, `red` vacío) y el
escenario del quickstart §6 (SC-002). En `make ci` la fijan, sin modelo: por US1, `TestEvalsDelRepositorio`
(`expresiones-calibradas`, `expresiones-en-los-bloques`, `expresiones-de-la-skill`, `prosa-de-la-skill`),
`TestProsaDeLaSkill`, `TestExtraerExpresionesProhibidas` y `TestJuzgarLasExpresionesProhibidas` (FR-010 a FR-024,
FR-091, FR-095; SC-003 a SC-005); por US2, `TestUmbralesDelInforme` y `TestInformeMarkdownDeLosUmbrales` (FR-001 a FR-008,
FR-092; SC-006); por US3, `TestInformeConSesionesSinMedir`, `TestLeerSesionConReintentos`, `TestClasificarElLimite`,
`TestEjecutarSesiones*`, `TestTopeDeLaSesion` y `TestDefinicionDelJob` (FR-030 a FR-052, FR-093, FR-094; SC-007, SC-008);
por US4, `TestJuicioDelSondeo`, `TestSalidaDelSondeo`, `TestComprobarElSondeo`, `TestSondear`, `TestGuionDelSondeo` y
`TestPrepararElArbolDelSondeo` (FR-060 a FR-068, FR-096; SC-009). Ninguna tarea `[aceptacion]`.

## Controles mecánicos que este hito añade o toca

### Objetivos del `Makefile`

- **Nuevo**: `evals-sondeo` (en `make help`; fuera de `make ci`: abre sesiones con modelo).
- **Sin cambios**, con más dentro: `test` y `test-integration` (los tests nuevos de `internal/evals`), `skills-check`
  (la subprueba `prosa-de-la-skill` y la de calibrado ampliada, dentro de `TestEvalsDelRepositorio`). `evals` sigue
  siendo `scripts/evals.sh "$(SKILL)"`. `schema-check` no cambia: `expresiones-prohibidas.yaml.json` no sale de
  `--describe`.

### Tests nuevos

| Test | Fichero | Cubre |
|---|---|---|
| `TestUmbralesDelInforme` | `umbrales_test.go` | FR-001 a FR-006, FR-051, FR-092; SC-006 |
| `TestInformeMarkdownDeLosUmbrales` | `informe_test.go` | FR-001 |
| `TestInformeConSesionesSinMedir` | `informe_test.go` | FR-033, FR-040 a FR-044, FR-093; SC-007 |
| `TestLeerSesionConReintentos` | `sesion_test.go` | FR-033, FR-040, FR-063, FR-065 |
| `TestClasificarElLimite` | `limites_test.go` | FR-040, FR-041, FR-093 |
| `TestEjecutarSesionesEnParalelo`, `TestEjecutarSesionesTrasElLimiteDeUso`, `TestEjecutarSesionesConElContextoCancelado`, `TestTopeDeLaSesion` | `sesiones_test.go` | FR-030 a FR-032, FR-036, FR-037, FR-044, FR-064, FR-094; SC-007, SC-008 |
| `TestDefinicionDelJob` | `definicion_test.go` | FR-030, FR-034, FR-035, FR-051, FR-094; SC-008 |
| `TestJuicioDelSondeo`, `TestSalidaDelSondeo`, `TestComprobarElSondeo`, `TestSondear`, `TestGuionDelSondeo` | `sondeo_test.go` | FR-060 a FR-068, FR-096; SC-009 |
| `TestPrepararElArbolDelSondeo` (etiqueta `integration`) | `sondeo_integracion_test.go` | FR-062, FR-064 |
| Subprueba `prosa-de-la-skill` y `TestProsaDeLaSkill` | `conjunto_test.go` | FR-013, FR-091; SC-005 |

### Tests existentes que cambian

- `TestEvalsDelRepositorio/expresiones-calibradas`: dos informes, tres familias y «alguna», con los repartos de
  contracts/lista-de-expresiones.md §4 (FR-021; SC-003). `listaDelRepositorio` exige también `anuncio` no vacía.
- `formato_test.go` (casos del esquema), `TestLeerConjunto`, `TestExtraerExpresionesProhibidas`,
  `TestJuzgarLasExpresionesProhibidas`: la tercera familia (FR-023, FR-024).
- Todo test que aplica la lista del repositorio, o la del contrato de H7.2 si gana `anuncio`, a un texto con alguna de
  las 14 expresiones de `anuncio` (hoy, la transición de la memoria de H7.1, `transicionDeLaMemoria` de
  `juzgar_test.go` y sus variantes, que lleva «tengo todo lo necesario») gana esa expresión, detrás de las de la
  maquinaria, en sus expresiones y motivos esperados: `TestExtraerExpresionesProhibidas` (`prohibidas_test.go`),
  `TestJuzgarLasExpresionesProhibidas` (`juzgar_test.go`), los informes armados con `copiarLaListaDelRepositorio`
  (`informe_test.go`) y el caso `tres-fallos-a-la-vez` de `TestCondicionesDeLaConsultaRepetida`
  (`consulta_repetida_test.go`, que espera hoy «la primera respuesta lleva expresiones prohibidas: memoria de
  consultas, hallazgos»). Los cambia T002, que declara esos ficheros (FR-024).
- `informe_test.go`, `juzgar_test.go` y `sesion_test.go`: los informes y resultados esperados ganan `umbrales`,
  `duracion_de_las_sesiones`, `reintentos_por_limite_de_ritmo`, `sesiones_sin_medir`, `sin_medir`, las secciones y
  columnas nuevas de `informe.md` y el texto del error en el motivo sin terminar (`TestLeerSesion/result-con-is-error`,
  el único transcript versionado con un `result` con `is_error`, pasa a `result con is_error: API Error: 529 …`;
  research V18). Ninguno se desactiva ni se salta.

### Puntos de entrada fuera de `make ci` (etiqueta `evals`)

`TestEjecucionDelJob` (lo ejecuta `scripts/evals.sh`, en el job) y `TestSondeo` (lo ejecuta
`scripts/evals-sondeo.sh`, la persona); `TestComprobarConsultaRepetida`, sin cambios. Se retiran `TestPlanDeSesiones`,
`TestPrepararSesion` y `TestInformeDelJob` (research D19). `golangci-lint` los lintea.

### Fixtures, `testdata/` y `schemas/` (tareas `[datos]`)

- `schemas/expresiones-prohibidas.yaml.json`: `anuncio`, con los casos de `formato_test.go` que lo fijan y la lista
  que lo cumple. Ningún fichero de `testdata/` cambia: los sustitutos y las sesiones sintéticas son constantes de los
  tests en `t.TempDir()`.

### CI

`.github/workflows/evals.yml`, trabajo `evals`: `name`, `concurrency`, `include` con la concurrencia y el objetivo por
skill, y dos variables de `env`; el trabajo `cambios` gana `scripts/evals-sesion.sh` en su filtro
(contracts/ejecucion-del-job.md §6). Lo comprueba `TestDefinicionDelJob` en `make ci`. `ci.yml` no cambia. El job de
evals se ejecuta en la propuesta de cambio, tras la revisión final, y lo lanza el workflow (FR-098).

## Controles de umbral

Las dos filas del cierre que fija FR-099 —el umbral de Haiku 4.5 (FR-004) no tiene fila: solo se publica— y, como pide
el ADR 0029 (criterio o), una por cada umbral que se mide en `make ci`, con el test que lo pone en rojo.

| Requisito | Umbral | Control | Dónde |
|---|---|---|---|
| FR-002, FR-003, SC-001 (Sonnet 5) | ≤ 5 % de las respuestas de Sonnet 5 en evals que activan la skill (≤ 2 de 51) | umbral del informe con `decide: true`: por encima, motivo, veredicto `fallo` y el trabajo en rojo, que el cierre cuenta | `evals:boe-legislacion:expresiones_prohibidas:claude-sonnet-5` |
| FR-051, SC-001 (duración) | ≤ 900 s de sesiones | umbral del informe con `decide: true`: por encima, motivo de la ejecución y `fallo` | `evals:boe-legislacion:duracion_de_las_sesiones` |
| FR-043, SC-001 (ninguna sin medir) | 0 sesiones sin medir por límite de uso | con alguna, motivo propio y `fallo` en el job; el test lo ve fallar con sesiones sintéticas | `ci:internal/evals/informe_test.go:TestInformeConSesionesSinMedir` |
| FR-021, FR-095, SC-003 | exactamente 35 de 93 (H7.1) y 10 de 93 (H7.2), eval por eval y familia por familia | subprueba `expresiones-calibradas` | `ci:internal/evals/conjunto_test.go:TestEvalsDelRepositorio` |
| FR-022, FR-095, SC-004 | 0 expresiones en los bloques grabados y en las formas que enseña la skill | subpruebas `expresiones-en-los-bloques` y `expresiones-de-la-skill` | `ci:internal/evals/conjunto_test.go:TestEvalsDelRepositorio` |
| FR-013, FR-091, SC-005 | 0 expresiones en la prosa de `SKILL.md` y 0 fechas `AAAAMMDD` con cifras | subprueba `prosa-de-la-skill` | `ci:internal/evals/conjunto_test.go:TestEvalsDelRepositorio` |
| FR-030, FR-094, SC-008 (repartidor) | ≤ `CONCURRENCIA_DE_EVALS` sesiones a la vez; 0 directorios compartidos | test con los sustitutos | `ci:internal/evals/sesiones_test.go:TestEjecutarSesionesEnParalelo` |
| FR-030, FR-034, FR-035, SC-008 (definición) | concurrencia 4 y 1; tope ≥ peor caso de cada skill; una tanda por commit | comprobación de `evals.yml` | `ci:internal/evals/definicion_test.go:TestDefinicionDelJob` |
| FR-064, SC-009 | 0 ficheros fuera del temporal y 0 ficheros suyos al terminar | tests del sondeo con los sustitutos (sesiones y guion) y con la construcción real del árbol (`test-integration`) | `ci:internal/evals/sondeo_test.go:TestSondear`, `ci:internal/evals/sondeo_test.go:TestGuionDelSondeo`, `ci:internal/evals/sondeo_integracion_test.go:TestPrepararElArbolDelSondeo` |
| FR-017, SC-010 | `SKILL.md` < 300 líneas | `TestSkillsDelRepositorio` en `skills-check` | `ci:Makefile:skills-check` |

SC-002 no tiene fila: lo mide una persona (ADR 0029).

## Uso, de fuera adentro (criterio de uso, ADR 0028)

Detalle, con ejemplos y bytes, en contracts/skill-boe-legislacion.md §6, contracts/informe-del-job.md §7,
contracts/ejecucion-del-job.md §9 y contracts/sondeo.md §4 y §8. Resumen:

- **La respuesta de la skill** (la persona; una por pregunta): lo que ocupa la norma y 0 líneas sobre la comprobación;
  161 B por bloque cuya redacción cambió (0 en la mayoría, como mucho k × 161 B), 83 B si la comprobación falla. Las
  órdenes por pregunta no cambian: cada bloque una vez y una `graph check` por norma citada, acotada a la pregunta
  (H7.1: ≈ 300 B sin nada, ≤ 3 800 B con cinco bloques cambiados), con cientos de normas y miles de bloques consultados
  igual que con uno. La línea se apaga con la lectura siguiente del bloque (H7.1 FR 024).
- **`umbrales`** (el job, el informe final, la persona; una vez por job y skill): ≈ 1 KB en `boe-legislacion`, `[]` en
  `legal-core`; se miden de nuevo en cada job.
- **Sesiones sin medir, reintentos, duración** (el job y quien ajusta la concurrencia; una vez por job): `[]`, 0 y un
  número lo habitual; ≤ 22 KB en el peor caso (las 94 sin medir); solo en ese job.
- **La salida del sondeo** (quien ajusta `SKILL.md`; una por sondeo): ≈ 0,9 KB con 5 evals, < 3 KB con todas sus sesiones
  listadas; no deja nada tras de sí.
- **`SKILL.md`** (el modelo; una vez por conversación): < 300 líneas (270 en el prototipo, research V17). **La lista**:
  52 expresiones, ≈ 1,3 KB, fija.

## Decisiones

- **La comprobación de la redacción no cambia de sitio (FR-015).** `graph check` es la última orden en 51 de 51 sesiones
  de Sonnet 5 del cierre de H7.2 —en las 41 sin ruido igual que en las 10 con él— y en 30 de 30 de Haiku 4.5, que no lo
  escribe nunca: el sitio es común a las respuestas limpias y a las que tienen ruido, y lo que distingue a estas es el
  vocabulario de la prosa (research, «Causa de raíz»). Se queda lo que decidió H7.1 —una vez por norma citada, después
  de leer sus bloques y antes de la respuesta—; lo que cambia es que la orden ya no va unida a «antes de redactar la
  respuesta» (C4). Rechazado moverla al final del paso 3: con varias normas seguiría siendo la última orden, y una
  lectura posterior apagaría `version-obsoleta`.
- **El modelo repite la prosa, no la salida del binario (FR-014)**: el cambio va a `SKILL.md`, no a cómo se lee `graph
  check`, cuya clave `data.hallazgos` se queda en código (research, «Causa de raíz»; D1).
- **Una tanda por commit: el segundo disparo espera** (FR-034; research D11).
- **La preparación de `TestPrepararSesion`, en proceso (FR-061).** Lo que prepara `TestPrepararSesion` es
  `PrepararSesion` con las evals de la skill y `UnionDeGrabaciones()`; el repartidor lo llama igual, justo antes de
  cada sesión, en el job y en el sondeo, así que la preparación es la misma en los dos. La entrada `TestPrepararSesion`
  se retira porque ya nadie la ejecuta (research D7, D19); un `go test` por sesión solo para llamar a esa función se
  rechazó.
- Las demás, con su alternativa rechazada, en research D1-D22.

## Trazabilidad: cada mecanismo y su requisito

| Mecanismo | Requisito |
|---|---|
| `Umbral`, su construcción por modelo y por objetivo, su comparación, `Informe.Umbrales` siempre presente, sección «Umbrales» de `informe.md`; la regla por serie, sin cambios | FR-001 a FR-007, FR-051 |
| Motivo de umbral y motivos «de la ejecución, no de la skill» | FR-003, FR-043, FR-051 |
| `Sesion.Reintentos`, `TerminaEnReintento`, `ErrorDelResultado` y el texto en el motivo sin terminar, sea cual sea el código (research D6, V18) | FR-033, FR-040, FR-063, FR-065 |
| Clasificación (a) (b) (c) | FR-040 a FR-042 |
| `ResultadoDeEval.SinMedir` y `ReintentosPorLimiteDeRitmo`; `TasaDelInforme.SinMedir`; `Informe.SesionesSinMedir`, `ReintentosPorLimiteDeRitmo`, `DuracionDeLasSesiones`; recuento sin las sin medir | FR-002, FR-033, FR-042, FR-043, FR-050 |
| `InformeAEscribir.SinAbrir`, `ObjetivoDeDuracion`, `DuracionDeLasSesiones` | FR-044, FR-050, FR-051 |
| Repartidor: orden, concurrencia, preparación con `PrepararSesion`, parar tras (a), duración | FR-030, FR-032, FR-036, FR-037, FR-044, FR-050, FR-061 |
| Tope en Go con grupo de procesos (124/137) | FR-031, FR-062 |
| Cancelación con `SIGINT`/`SIGTERM` | FR-037, FR-064 |
| `claude/` con enlaces a las skills, `tmp/`, `TMPDIR` y `CLAUDE_CODE_TMPDIR` por sesión | FR-032, FR-063 |
| `scripts/evals-sesion.sh` | FR-031 |
| `scripts/evals.sh`: `CONCURRENCIA_DE_EVALS`, `TestEjecucionDelJob`; retirada de las tres entradas y de `plan.tsv` | FR-030, FR-031, FR-050 |
| `evals.yml`: `concurrency`, `name`, `include`, `env`, filtro de `cambios` | FR-030, FR-031, FR-034, FR-035, FR-051 |
| `leerDefinicionDelJob`, `TestDefinicionDelJob` | FR-060 (concurrencia del sondeo), FR-094 |
| `make evals-sondeo`, `scripts/evals-sondeo.sh`, `TestSondeo`, comprobación de argumentos y credencial, árbol de trabajo, entorno del sondeo, juicio sin traza, salida | FR-060 a FR-067 |
| Sustitutos de `claude`, `strace` y `go` en los tests | FR-068, FR-094, FR-096 |
| `anuncio` en la lista, el esquema y `ExpresionesProhibidas` | FR-020, FR-023, FR-024 |
| `expresiones-calibradas` con dos informes y tres familias | FR-021, FR-095 |
| `prosa-de-la-skill`, `TestProsaDeLaSkill` | FR-013, FR-091 |
| `SKILL.md` v0.1.3 (C1-C11) | FR-010 a FR-017 |
| `CHANGELOG.md`; `CONTRIBUTING.md` («Job de evals», lista de expresiones); `doc.go` | FR-018, FR-097; lo que el hito deja falso (research D20) |

Nada del diseño atiende a un estado que no pasa el umbral de materialidad: un transcript, una definición o una lista
que no se pueden leer siguen las reglas que ya hay —sesión ilegible, fichero mal formado, test que falla, error con
código 1—, sin caso propio.

## Cambios de `SKILL.md` trazados a la causa de raíz (FR-014)

C1-C11 de contracts/skill-boe-legislacion.md §1, cada uno con la línea de v0.1.2, el texto nuevo, la causa medida
(research, «Causa de raíz») y su requisito. En resumen: «memoria de consultas» (C1, C4, C7, C10), los códigos junto a la
palabra «código» (C2, C3, C8, C9), la enumeración de lo prohibido y el caso vacío (C4, C5, C6), «antes de redactar la
respuesta» (C4), el ejemplo con fechas (C7) y la regla 7 (C11).

## Datos externos

Ninguno (research D22): ni fuente ni grabación nuevas, ningún manifiesto `grabaciones.json` ni test `TestGrabar*`
cambia, y el paso `grabar_datos` no tiene nada que grabar. El calibrado usa los informes versionados de H7.1 y H7.2.

## Orden de implementación (de dentro afuera)

1. **`[datos]`** `schemas/expresiones-prohibidas.yaml.json` y `evals/boe-legislacion/expresiones-prohibidas.yaml` con
   `anuncio`, `ExpresionesProhibidas.Anuncio`, `ExtraerExpresionesProhibidas` y `recontarExpresiones` con tres familias,
   `listaDelRepositorio`, y sus tests (`formato_test.go`, `prohibidas_test.go`, `juzgar_test.go`, `conjunto_test.go`
   para `TestLeerConjunto`, e `informe_test.go` y `consulta_repetida_test.go`, que aplican la lista del repositorio a
   la transición de la memoria de H7.1; «Tests existentes que cambian»).
2. `expresiones-calibradas` sobre los dos informes y las tres familias.
3. `sesion.go` (reintentos, texto del error) y `limites.go`, con sus tests.
4. `umbrales.go` e `informe.go` (umbrales, sin medir, `SinAbrir`, reintentos, duración, `informe.md`), con sus tests.
5. `sesiones.go`, `scripts/evals-sesion.sh` y `sustitutos_test.go`, con los tests del repartidor.
6. `.github/workflows/evals.yml`, `definicion.go` y `TestDefinicionDelJob`.
7. `TestEjecucionDelJob` en `job_test.go` y `scripts/evals.sh`; fuera las tres entradas retiradas.
8. `sondeo.go`, `TestSondeo`, `scripts/evals-sondeo.sh`, el objetivo `evals-sondeo` del `Makefile` y sus tests.
9. `skills/boe-legislacion/SKILL.md` v0.1.3 con `prosa-de-la-skill` y `TestProsaDeLaSkill` en la misma tarea (la
   subprueba con v0.1.2 dejaría `make ci` en rojo).
10. `CHANGELOG.md`, `CONTRIBUTING.md` e `internal/evals/doc.go`.

**Obligaciones para `tasks.md`**: ninguna tarea `[aceptacion]` («Aceptación e2e: no aplica»); cada tarea deja `make ci`
en verde; la única tarea `[datos]` es la del paso 1, con las rutas de este plan; cada fila de «Controles de umbral»
tiene la tarea que construye su control, con un test que lo ve fallar; ninguna tarea ni corrección cumple un umbral de
FR-002 o FR-051 rebajándolo, retirándolo, dejándolo en `decide: false`, sacando las evals informativas del total o
recortando la lista (FR-008), ni cambia la regla por serie (FR-007); ninguna tarea usa la red, graba, abre una sesión
con modelo, ejecuta `make evals` o `make evals-sondeo`, publica ni mide en la plataforma (FR-068); el quickstart §6 lo
ejecuta la persona al leer el informe final.

## Complexity Tracking

| Desviación | Por qué hace falta | Alternativa más simple rechazada |
|---|---|---|
| «Aceptación e2e: no aplica» (principio III: «cada hito empieza por el test e2e») | El hito no cambia el binario: no hay comportamiento suyo que un guion describa. La aceptación es la del job (SC-001) y la del quickstart (SC-002), y cada pieza entra con sus tests en `make ci` | Un guion que ejerciera el binario sin cambios: no describiría la entrega |
| Una tarea `[datos]` con código y tests (el esquema, la lista, el tipo y los casos que los fijan) | Con `anuncio` obligatoria en el esquema, la lista sin ella es un fichero mal formado: separarlos deja `make ci` en rojo entre dos tareas. Precedente: el esquema de la lista en H7.2 | Separar esquema y lista |

## Comprobación contra la rúbrica del juez (`juez_plan`) y `precheck.sh plan`

- `precheck.sh plan`: existen `plan.md` y `research.md`; ninguno conserva marcas de aclaración pendiente; están
  `## Constitution Check`, la línea «Aceptación e2e» y `## Controles de umbral`, con cada fila nombrando requisitos del
  spec y un control con forma (`evals:<skill>:<nombre>` o `ci:<ruta>[:<Test>]`).
- a · Constitution Check: un ítem por principio (I-IX) y por regla de dependencia (R1-R6).
- b · Dependencias: ninguna nueva.
- c · Reglas de dependencia: tabla R1-R6; ni `net/http`, ni SQLite, ni `os.Exit`, ni salida estándar nuevos.
- d · Errores y códigos: ningún código de `kitlegal` nuevo; errores de `internal/evals` con su fichero o argumento; los
  guiones salen con 1 (research D16, D19).
- e · Tests primero: «Aceptación e2e: no aplica» con el motivo; «Tests nuevos», «Tests existentes que cambian», puntos
  de entrada, fixtures y la tarea `[datos]`.
- f · Alcance: nada fuera del spec; la documentación que se toca es la que el hito deja falsa, y el sondeo no se
  documenta fuera de `make help`, del quickstart y de `CHANGELOG.md` (research D20).
- g · Sin atajos: ningún `//nolint`, `t.Skip`, TODO ni error silenciado previstos (G204 se resuelve con el guion de la
  sesión, D8).
- h · Mejor alternativa: cada decisión con la rechazada (D1-D22).
- i · Afirmaciones verificadas: V1-V18 con fichero, offset u orden; S1-S7 como supuestos. El motivo de una sesión sin
  terminar lleva el texto del `result` con `is_error` también con código 1, el que da una credencial que no sirve (V18).
- j · Quickstart ejecutable: cada escenario con órdenes y nombres de test reales; las copias salen de `git archive` y se
  borran; lo que escribe en el árbol, declarado (§8, ficheros ignorados).
- k · Datos externos: ninguno.
- l · Autonomía: ninguna tarea para una persona; el §6 lo ejecuta la persona al leer el informe final; el cierre en la
  plataforma lo hace el workflow.
- m · Uso: «Uso, de fuera adentro» y los contratos (bytes, órdenes por pregunta, apagado de cada señal).
- n · Proporcionalidad: «Trazabilidad»; sin mecanismos para estados sin vía real.
- o · Controles de umbral: las dos filas del cierre de FR-099, sin la de Haiku, y una por cada umbral medido en
  `make ci`.
