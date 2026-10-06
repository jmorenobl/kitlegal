# Implementation Plan: H24 · Las evals juzgan el significado con un modelo: «afirma lo que no ha leído» decide, `boe-legislacion` deja de glosar lo que no ha leído, y la lista de expresiones deja de decidir

**Branch**: `017-h24-las-evals-juzgan` | **Date**: 2026-10-05 | **Spec**: [spec.md](./spec.md)

**Input**: Feature specification from `/specs/017-h24-las-evals-juzgan/spec.md`

**Modo**: desatendido. Las decisiones técnicas se tomaron con el «Criterio de decisión autónoma» de la constitución y
están en [research.md](./research.md) (D1-D26), cada una con su alternativa rechazada. Toda afirmación sobre Go, el
repositorio, los guiones del workflow o la definición del job remite a la tabla V de research.md, comprobada en local;
las medidas, a la tabla M; y lo que no se pudo comprobar en esta sesión —Claude Code, npm, Python y los tiempos— son
los supuestos S1-S8.

## Summary

H24 no toca el binario. Arregla `boe-legislacion` y cambia el instrumento que la mide, en siete piezas:

1. **El juez** (D2-D9): de cada respuesta del modelo que decide se leen del transcript los textos que devolvieron sus
   herramientas; con ellos, la pregunta y la respuesta se compone el mensaje de la validación del ADR 0037, y
   `scripts/evals-voto.sh` abre cada voto con su orden. El job comprueba sin modelo que la frase citada está en la
   respuesta, repite una vez el voto nulo y aplica la regla: marcada solo con tres síes, votando por orden.
2. **Las clases en datos** (D20): `evals/boe-legislacion/juez/clases.yaml`, con `afirma_lo_no_leido`, que decide con 0,
   y `cuenta_su_proceso`, que se publica; junto a ella, las cuatro copias de la evidencia del ADR 0037.
3. **La medida** (D11, D15-D19): `make ci` y el job comprueban sin modelo que la medida versionada corresponde a la
   rúbrica, a los casos, al modelo del juez y a su versión de Claude Code; si no, el job no abre ninguna sesión. La
   ejecución de la medida, que lanza una persona, reconstruye en proceso los textos de los 259 casos y los vota.
4. **El informe** (D10, D13): `umbrales` pasa a doce elementos, diez que deciden; la clave `juez` lleva los votos y las
   frases de cada respuesta con algún voto afirmativo, y las que quedan sin juzgar.
5. **La lista de expresiones** deja de aplicarse a toda respuesta (D14) y se queda como el vocabulario de la prosa,
   con tres expresiones más (D22).
6. **`boe-legislacion` v0.1.7** (D23): una regla nueva, con su razón, sobre el precepto que no se ha leído, y la prosa
   sin «el sobre» ni los nombres de campo. Es un diff que se aplica sobre v0.1.6 y deja 298 líneas.
7. **El job**: el modelo del juez y su versión de Claude Code fijados aparte, un segundo Claude Code instalado para
   sus votos, el trabajo `medida` con su etiqueta y su entrada, y los topes recalculados (D6, D21).

No hay ADR nuevo, y `specs/011-…` a `specs/016-…` y `scripts/workflow/` no se editan (FR-096).

## Technical Context

**Language/Version**: Go 1.27 (`go 1.27.0`, `toolchain go1.27.1`), `CGO_ENABLED=0`, `-trimpath`; bash (los dos guiones
nuevos, sin nada posterior a 3.2); YAML de GitHub Actions.

**Primary Dependencies**: las fijadas, sin ninguna nueva: el lector de YAML del repositorio, `santhosh-tekuri/jsonschema`
v6 (la declaración de clases, la lista y cada voto contra `esquema.json`), `stretchr/testify`; biblioteca estándar
(`os/exec`, `context`, `encoding/json`, `crypto/sha256`, `unicode`, `time`). En el job, `npm` y `gh`, que ya usa.

**Storage**: ninguno nuevo. Ficheros del repositorio: la carpeta del juez (`evals/boe-legislacion/juez/`), dos esquemas
(`schemas/`), la eval retirada y su derivada (`testdata/evals/retiradas/`), la skill, dos guiones y la definición del
job.

**Testing**: `go test -race` en `make ci`: tablas con `t.Parallel`, sesiones y transcripts sintéticos, votantes que
devuelven salidas grabadas, el `claude` sustituto de `sustitutos_test.go` y la reconstrucción en proceso de los casos.
Puntos de entrada con la etiqueta `evals`, fuera de `make ci`: `TestEjecucionDelJob`, `TestSondeo` y, nuevo,
`TestMedidaDelJuez`. El job de evals, en la propuesta de cambio.

**Target Platform**: el job, `ubuntu-24.04`; `make ci` y el sondeo, macOS y Linux. El binario distribuido no cambia
salvo la skill que empotra.

**Project Type**: skill (`boe-legislacion`) e instrumento de medida (`internal/evals`, `scripts/`, el flujo `evals`).

**Performance Goals**: votos de un modo ≤ 900 s (FR-033); con unos 8 s por voto (ADR 0037) y cuatro respuestas a la
vez, 54 votos son unos 108 s (estimado; S6). Peor caso del trabajo `evals`, 21 097 s; del trabajo `medida`, 16 105 s
(contracts/job-de-evals.md §4).

**Constraints**: ningún test de `make ci` usa la red ni abre una sesión con modelo; `SKILL.md` ≤ 298 líneas (V13);
ningún `//nolint`; `evidencias/` no se escribe; la rúbrica, los casos y la medida no cambian (FR-045).

**Scale/Scope**: 21 evals y 198 sesiones por job de `boe-legislacion`; 111 respuestas juzgadas (54, 54 y 3), con 111
votos si ninguna se marca; 259 casos etiquetados y 683 votos en una ejecución de la medida (M2).

## Constitution Check

*GATE: pasado antes de la fase 0 y re-evaluado tras el diseño (fase 1): sin violaciones; las desviaciones justificadas
están en Complexity Tracking.*

### Principios

| Principio | Cómo lo cumple H24 |
|---|---|
| I · Fuentes públicas y frontera humana | No toca ninguna fuente legal, `internal/httpx` ni `docs/SOURCES.md`, y no graba nada (D25). Las sesiones siguen tras el proxy que solo deja pasar al modelo; el voto del juez no tiene herramientas ni red propia. La medida la lanza y la versiona una persona (FR-050, FR-054). |
| II · Nada sin cita ni fuente | Es lo que el hito hace cumplir: la skill deja de decir qué dice o de qué trata un precepto que ninguna herramienta devolvió (FR-081 a FR-083), y `afirma_lo_no_leido` decide con 0 (FR-030). El sobre no cambia. Ningún caso ni ningún texto se escribe a mano (FR-024). |
| III · Tests primero y offline | «Aceptación e2e: no aplica» (abajo, con el motivo). Cada pieza entra con sus tests en `make ci`, offline y sin modelo; lo que entra en `schemas/` y `testdata/` va en tareas `[datos]`; la aceptación es el job (SC-001). |
| IV · Hexagonal y errores tipados | Ningún paquete del binario ni código de salida de `kitlegal` cambian. `internal/evals` devuelve `error` con el fichero o el voto delante; los guiones salen con 1; ningún `panic`. Ni `os.Exit` ni salida estándar desde Go (R4, R5). |
| V · Simplicidad y dependencias | Ninguna dependencia nueva. Cada mecanismo se traza a un FR («Trazabilidad»). Se rechazaron un paquete aparte, una guarda por variable de entorno, un plazo por modo, un esquema para la medida y los votos versionados como datos de test (research D1, D6, D18, D19, D26). |
| VI · Un binario, convenciones de agente | El binario, sus banderas y `--describe` no cambian; ningún ejecutable nuevo en `cmd/`: la medida es un punto de entrada de test con la etiqueta `evals`, como el job y el sondeo. |
| VII · Grafo y privacidad | Nada nuevo en el grafo. La reconstrucción de los casos usa grafos temporales que se descartan. |
| VIII · Skills primero | Mejora `boe-legislacion` (v0.1.7) y cambia cómo se mide, con un umbral que decide sobre lo que la respuesta dice. Deja el instrumento que H23 necesita. `SKILL.md` < 300 líneas, sin nombrar evals, el job ni modelos. |
| IX · Genericidad territorial | Nada se particulariza para un municipio: ni el juez, ni la rúbrica, ni la skill. |

### Reglas de dependencia (`docs/ROADMAP.md` §2, constitución IV)

| Regla | Cómo la cumple |
|---|---|
| R1 · `internal/core/**` no importa adaptadores | Sin cambios en `internal/core`. |
| R2 · Solo `internal/httpx` importa `net/http` | Ningún fichero nuevo lo importa: el voto es un proceso (`os/exec`). |
| R3 · Solo `cache`/`store`/`graph` importan SQLite y `database/sql` | `internal/evals` no los importa: la caché y el grafo de la reconstrucción se abren con las opciones de sus paquetes, como hoy en `preparar.go`. |
| R4 · Solo `cli` y `cmd/` llaman a `os.Exit` | Ningún `os.Exit` nuevo: los puntos de entrada son tests; los guiones salen con su código. |
| R5 · Solo `render` escribe en stdout | Ningún `fmt.Print*`, `os.Stdout` ni `os.Stderr` nuevo en Go: la medida se escribe en el fichero de `-salida` y la imprime su guion, como el informe. |
| R6 · `internal/graph` no importa `source/*` ni `render` | Sin cambios. |

## Project Structure

### Documentation (this feature)

```text
specs/017-h24-las-evals-juzgan/
├── plan.md                  # este fichero
├── research.md              # V1-V22, M1-M7, S1-S8, D1-D26
├── data-model.md            # juez, texto, voto, medida, caso, informe, job, ficheros
├── quickstart.md            # §1-§9; §7, el sondeo; §8, el job de cierre; §9, la medida
├── contracts/
│   ├── juez-y-voto.md                       # carpeta del juez, textos, mensaje, orden, lectura, frase, regla
│   ├── informe-del-job.md                   # qué se juzga, umbrales, juez, motivos, instrumento sin medir, sondeo
│   ├── medida-del-juez.md                   # medida, comprobación, copias, casos, reconstrucción, ejecución
│   ├── job-de-evals.md                      # evals.yml, guiones, topes, TestDefinicionDelJob
│   ├── skill-boe-legislacion.md             # causa, C1-C8, lo que se queda, la prosa, uso
│   └── skill-boe-legislacion-v0.1.7.diff    # el prototipo, aplicable con git apply
├── checklists/              # del spec
└── tasks.md                 # la escribe /speckit-tasks
```

### Source Code (repository root)

```text
schemas/juez-clases.yaml.json                       # nuevo [datos]
schemas/expresiones-prohibidas.yaml.json            # salida_de_las_herramientas [datos]
evals/boe-legislacion/juez/                         # nuevo: clases.yaml y las cuatro copias
evals/boe-legislacion/expresiones-prohibidas.yaml   # la clave nueva y su cabecera
testdata/evals/retiradas/                           # la eval 19 de H7.1 y su derivada, restauradas [datos]
skills/boe-legislacion/SKILL.md                     # v0.1.7
scripts/evals-voto.sh, scripts/evals-medir-juez.sh  # nuevos
scripts/evals.sh                                    # las tres variables del juez
Makefile                                            # objetivo evals-medir-juez
.github/workflows/evals.yml                         # variables del juez, segundo Claude Code, trabajo medida, topes
.golangci.yml                                       # «informativo» en ignore-rules de misspell (T018, la convergencia)
internal/app/grafo_test.go                          # la derivada restaurada en TestGrabacionesDerivadas
internal/evals/
├── juez.go                  # nuevo: clases, textos, mensaje, frase, lectura del voto, regla, votante
├── medida.go                # nuevo: medida, comprobación, casos, reconstrucción, ejecución de la medida
├── ejecucion.go             # nuevo: el recorrido del job, con la comprobación de la medida delante
├── sesion.go                # Sesion.Textos
├── conjunto.go              # la carpeta juez; Conjunto.Juez
├── formato.go               # Eval sin Prohibidas
├── prohibidas.go            # SalidaDeLasHerramientas
├── juzgar.go                # sin la lista
├── informe.go, umbrales.go  # el juez, la clave juez, los doce umbrales, los motivos, informe.md
├── sondeo.go                # el juez y sus líneas
├── consulta_repetida.go     # sin la lista
├── definicion.go            # variables del juez, trabajo medida, peores casos
├── grabaciones.go           # EvalsRetiradas: dónde están la eval restaurada y su grafo previo
├── preparar.go              # leerEvalsBienFormadas, que comparten PrepararSesion y la reconstrucción
├── job_test.go              # TestEjecucionDelJob con ejecutarElJob; TestMedidaDelJuez; TestSondeo
├── doc.go
└── juez_test.go, medida_test.go, ejecucion_test.go, sustitutos_test.go y los _test.go de lo que cambia
CHANGELOG.md, CONTRIBUTING.md, docs/WORKFLOW.md
```

**Structure Decision**: la de `docs/ROADMAP.md` §2 y `CLAUDE.md`, sin paquetes nuevos: el juez, la medida y el
recorrido del job en `internal/evals`; sus guiones en `scripts/`; la definición del job en `.github/workflows/`; la
skill en `skills/`; los datos del juez en `evals/<skill>/juez/`. Ningún fichero del binario cambia salvo `SKILL.md`,
que el binario empotra.

## Aceptación e2e

**Aceptación e2e: no aplica.** El hito no cambia el binario —cambiarlo está fuera de alcance, y `cmd/kitlegal` no
enlaza `internal/evals`—, así que no hay comportamiento de `kitlegal` que un guion `testscript` describa (research
D24). Su aceptación es la de la skill y la del instrumento: el job de evals de cierre (SC-001). En `make ci` la fijan,
sin modelo:

| Historia | Tests | Requisitos |
|---|---|---|
| US1 | `skills-check`, `TestProsaDeLaSkill` y `TestEvalsDelRepositorio/prosa-de-la-skill` | FR-085, FR-086, FR-110; SC-010 |
| US2 | `TestTextosDeLaSesion`, `TestMensajeDelVoto`, `TestOrdenDelVoto`, `TestFraseEnLaRespuesta`, `TestVotoDelJuez`, `TestReglaDeLosVotos`, `TestInformeConElJuez`, `TestUmbralesDelInforme` | FR-001 a FR-014, FR-030 a FR-035, FR-102 a FR-104, FR-107; SC-002 a SC-004, SC-007 |
| US3 | `TestMedidaVersionada`, `TestCopiasDelJuez`, `TestEjecucionSinMedir`, `TestEvalsDelRepositorio/formato` | FR-020 a FR-023, FR-040 a FR-045, FR-105, FR-108; SC-005, SC-008 |
| US4 | `TestGrabacionesDerivadas` (los dos), `TestEjecucionDeLaMedida`, `TestGuionDeLaMedida`, `TestDefinicionDelJob` | FR-024, FR-050 a FR-054, FR-106, FR-109, FR-112; SC-006, SC-009, SC-012 |
| US5 | `TestInformeConElJuez` | FR-060, FR-061; SC-004 |
| US6 | `TestJuzgarSinLaLista`, `TestUmbralesDelInforme` | FR-034, FR-062, FR-070, FR-071, FR-111; SC-011 |
| US7 | `TestJuicioDelSondeo`, `TestSalidaDelSondeo` | FR-075, FR-076 |

Ninguna tarea `[aceptacion]`.

## Controles mecánicos que este hito añade o toca

### Objetivos del `Makefile`

`evals-medir-juez`, nuevo y fuera de `make ci` (abre sesiones con modelo). `test` y `skills-check` llevan más dentro:
los tests nuevos y la lectura de `juez/` en `TestEvalsDelRepositorio`. `schema-check` no cambia: los dos esquemas que
se tocan no salen de `--describe`.

### Tests nuevos

| Test | Fichero | Cubre |
|---|---|---|
| `TestTextosDeLaSesion` | `sesion_test.go` | FR-001, FR-107 |
| `TestMensajeDelVoto`, `TestOrdenDelVoto` | `juez_test.go` | FR-001 a FR-004, FR-107; SC-007 |
| `TestFraseEnLaRespuesta`, `TestVotoDelJuez` | `juez_test.go` | FR-005 a FR-007, FR-102; SC-002 |
| `TestReglaDeLosVotos` | `juez_test.go` | FR-010, FR-011, FR-103; SC-003 |
| `TestInformeConElJuez` | `informe_test.go` | FR-012 a FR-014, FR-030, FR-031, FR-033, FR-035, FR-037, FR-060, FR-061, FR-104; SC-004 |
| `TestEjecucionSinMedir` | `ejecucion_test.go` | FR-043, FR-104; SC-004 |
| `TestMedidaVersionada`, `TestCopiasDelJuez` | `medida_test.go` | FR-023, FR-041, FR-042, FR-105, FR-108; SC-005, SC-008 |
| `TestGrabacionesDerivadas` (el de `internal/evals`) | `medida_test.go` | FR-024, FR-109; SC-009 |
| `TestEjecucionDeLaMedida`, `TestGuionDeLaMedida` | `medida_test.go` | FR-051 a FR-053, FR-106; SC-006 |
| `TestJuzgarSinLaLista` | `juzgar_test.go` | FR-070, FR-111; SC-011 |

### Tests existentes que cambian

- `TestEvalsDelRepositorio`: salen `expresiones-calibradas`, `expresiones-en-los-bloques` y `expresiones-de-la-skill`,
  con lo que solo ellas usan (FR-071), salvo las rutas de los tres informes del calibrado y la lectura de sus
  entradas, que pasan a `TestJuzgarSinLaLista`; `formato` lee además la carpeta del juez; `prosa-de-la-skill`
  aplica la clave nueva. `TestProsaDeLaSkill` gana los casos de contracts/skill-boe-legislacion.md §4. `TestLeerConjunto`, los de la
  carpeta `juez`.
- `TestUmbralesDelInforme`, `TestInforme…` y `TestJuzgar…`: sin `expresiones_prohibidas` en el resultado, en los
  motivos, en el recuento ni en `umbrales`; con los doce umbrales donde la skill sintética tiene juez.
- `TestExtraerExpresionesProhibidas` se queda: la comparación sigue sirviendo a la prosa.
- `TestCondicionesDeLaConsultaRepetida`: sin la condición de las expresiones (D14).
- `TestDefinicionDelJob`: `del-repositorio`, `sinteticas` y `peor-caso` con lo de contracts/job-de-evals.md §5.
- `TestJuicioDelSondeo`, `TestSalidaDelSondeo`, `TestSondear`: las líneas del juez en lugar del recuento.
- `TestGrabacionesDerivadas` de `internal/app`: la entrada de la derivada restaurada.
- Los tests de formato de `schemas/expresiones-prohibidas.yaml.json`: la clave nueva.

Ninguno se desactiva ni se salta.

### Puntos de entrada fuera de `make ci` (etiqueta `evals`)

`TestEjecucionDelJob` (llama a `ejecutarElJob`), `TestMedidaDelJuez` (nuevo) y `TestSondeo`. `golangci-lint` los
lintea. `TestComprobarConsultaRepetida` pierde la lista.

### Fixtures, `testdata/` y `schemas/` (tareas `[datos]`)

- `schemas/juez-clases.yaml.json`, con `evals/boe-legislacion/juez/clases.yaml`, el tipo y los casos que lo fijan.
- `schemas/expresiones-prohibidas.yaml.json`, con la clave nueva de la lista y los casos que la fijan.
- `testdata/evals/retiradas/`: los dos ficheros de contracts/medida-del-juez.md §4, escritos con `git show c4819d1^:…`,
  y la entrada de `TestGrabacionesDerivadas` de `internal/app` que comprueba la derivada.

Ninguna tarea escribe en `evidencias/`.

### CI

`.github/workflows/evals.yml` (contracts/job-de-evals.md): lo comprueba `TestDefinicionDelJob` en `make ci`. `ci.yml`
no cambia. El job de evals se ejecuta en la propuesta de cambio, tras la revisión final, y lo lanza el workflow
(FR-113).

## Controles de umbral

Los diez umbrales que deciden en el informe del cierre —los seis nuevos que fija FR-101 y los cuatro que siguen
(FR-034), que SC-001 cuenta— y, como pide el ADR 0029 (criterio o), una fila por cada umbral que se mide en `make ci`.
`cuenta_su_proceso` no tiene fila: solo se publica (FR-031).

| Requisito | Umbral | Control | Dónde |
|---|---|---|---|
| FR-030, SC-001 (modo orden) | 0 respuestas marcadas en `afirma_lo_no_leido`, de las juzgadas del modo | umbral del informe con `decide: true`: con una, motivo con la sesión y sus frases, veredicto `fallo` y el trabajo en rojo, que el cierre cuenta | `evals:boe-legislacion:afirma_lo_no_leido:claude-sonnet-5-5:orden` |
| FR-030, SC-001 (modo herramienta) | ídem | ídem | `evals:boe-legislacion:afirma_lo_no_leido:claude-sonnet-5-5:herramienta` |
| FR-032, SC-001 (defectos) | 0 defectos sin marcar, de 212 | umbral del informe con `decide: true`, con el recuento de la medida versionada; con otro valor, `make ci` falla y el job termina en `fallo` sin abrir sesiones | `evals:boe-legislacion:medida_del_juez:afirma_lo_no_leido:defectos_sin_marcar` |
| FR-032, SC-001 (correctos) | 0 correctos marcados, de 47 | ídem | `evals:boe-legislacion:medida_del_juez:afirma_lo_no_leido:correctos_marcados` |
| FR-033, SC-001 (modo orden) | ≤ 900 s de votos del modo | umbral del informe con `decide: true`: por encima, motivo de la ejecución y `fallo` | `evals:boe-legislacion:duracion_del_juez:orden` |
| FR-033, SC-001 (modo herramienta) | ídem | ídem | `evals:boe-legislacion:duracion_del_juez:herramienta` |
| FR-034, SC-001 (sin activar, modo orden) | 0 respuestas sin la skill activada | el de hoy, que sigue | `evals:boe-legislacion:sin_activar:claude-sonnet-5-5:orden` |
| FR-034, SC-001 (sin activar, modo herramienta) | ídem | ídem | `evals:boe-legislacion:sin_activar:claude-sonnet-5-5:herramienta` |
| FR-034, SC-001 (sesiones, modo orden) | ≤ 900 s de sesiones del modo | el de hoy, que sigue | `evals:boe-legislacion:duracion_de_las_sesiones:orden` |
| FR-034, SC-001 (sesiones, modo herramienta) | ídem | ídem | `evals:boe-legislacion:duracion_de_las_sesiones:herramienta` |
| FR-007, SC-001 (sin juzgar) | 0 respuestas sin juzgar | con alguna, motivo de la ejecución y `fallo` en el job; el test lo ve fallar con un voto que no llega | `ci:internal/evals/informe_test.go:TestInformeConElJuez` |
| FR-102, SC-002 | 4 de 4 casos del voto: vale, nulo y repetido 1 vez, vale con blancos y énfasis, sin juzgar | tabla con salidas grabadas | `ci:internal/evals/juez_test.go:TestVotoDelJuez` |
| FR-103, SC-003 | 3 votos y marcada; 1 voto y sin marcar; 3 votos, sin marcar y 2 frases; 1 voto en la clase que se publica | votante que cuenta sus llamadas | `ci:internal/evals/juez_test.go:TestReglaDeLosVotos` |
| FR-104, SC-004 | 1 marcada → `fallo`; 0 → se cumple; 3 en `cuenta_su_proceso` → mismo veredicto; la eval sin binario ni servidor marcada → mismos umbrales y 0 motivos; medida que no corresponde → `fallo` y 0 sesiones | sesiones sintéticas y un votante grabado; el recorrido del job con quien abre sesiones contado | `ci:internal/evals/informe_test.go:TestInformeConElJuez`, `ci:internal/evals/ejecucion_test.go:TestEjecucionSinMedir` |
| FR-042, FR-105, SC-005 | la medida del repositorio corresponde, con 0 y 0; 6 de 6 mutaciones en rojo | la comprobación sobre el repositorio y sobre copias cambiadas | `ci:internal/evals/medida_test.go:TestMedidaVersionada` |
| FR-106, SC-006 | 0 de 212 y 0 de 47 con 683 votos y una medida versionada que no corresponde; 1 defecto sin marcar o 1 correcto marcado → falla; 0 sesiones de evals | votante según la etiqueta | `ci:internal/evals/medida_test.go:TestEjecucionDeLaMedida` |
| FR-107, SC-007 | 0 bytes de `SKILL.md`, de la eval o del juicio sin modelo en el mensaje; la orden y el mensaje de la validación | el mensaje byte a byte; el guion con el `claude` sustituto | `ci:internal/evals/juez_test.go:TestMensajeDelVoto`, `ci:internal/evals/juez_test.go:TestOrdenDelVoto` |
| FR-023, FR-108, SC-008 | 4 de 4 copias idénticas | comparación byte a byte | `ci:internal/evals/medida_test.go:TestCopiasDelJuez` |
| FR-109, SC-009 | 259 respuestas iguales a las de su informe; 140 derivados que solo pierden su texto | reconstrucción en proceso | `ci:internal/evals/medida_test.go:TestGrabacionesDerivadas` |
| FR-086, FR-110, SC-010 | `SKILL.md` < 300 líneas; 0 usos de las tres expresiones en la prosa | `TestSkillsDelRepositorio` en `skills-check`; la prosa | `ci:Makefile:skills-check`, `ci:internal/evals/conjunto_test.go:TestProsaDeLaSkill`, `ci:internal/evals/conjunto_test.go:TestEvalsDelRepositorio` |
| FR-111, SC-011 | 0 sesiones que dejan de pasar por la lista | las dos respuestas del 2026-10-04 y las 56 del calibrado de H7.4: las 36, 11 y 9 que la lista del repositorio marca, aplicada a la `respuesta` de cada entrada, en los informes de H7.1, H7.2 y H7.3, exigidas como premisa | `ci:internal/evals/juzgar_test.go:TestJuzgarSinLaLista` |
| FR-092, FR-112, SC-012 | cada `timeout-minutes` ≥ su peor caso (21 097 s y 16 105 s); 2 versiones de Claude Code fijadas aparte; 0 formas de lanzar la medida fuera de su etiqueta y su entrada | la definición del repositorio y las sintéticas | `ci:internal/evals/definicion_test.go:TestDefinicionDelJob` |

SC-013 es `make ci` entero, y SC-014 lo mide una persona: no tienen fila.

## Uso, de fuera adentro (criterio de uso, ADR 0028)

El detalle, con ejemplos y bytes medidos (research M4), está en contracts/skill-boe-legislacion.md §5,
contracts/informe-del-job.md §2 a §5 y §8 y contracts/medida-del-juez.md §7 y §9. Nada de lo que el hito entrega
crece con lo consultado en meses de uso: el binario y sus órdenes no cambian, y el job parte de las evals y de los
casos del repositorio.

- **La respuesta de la skill** (la persona; una por pregunta): lo que ocupa la norma leída, con 0 frases sobre
  preceptos no leídos; el aviso de lo que no cubre, 80 B, como mucho uno por materia. Las órdenes por pregunta son las
  de v0.1.6. El aviso se apaga cuando la respuesta lee esa materia.
- **`umbrales`** (el job, el informe final, la persona; una vez por job): doce elementos de unos 325 B, 3,6 KB, fijos;
  `[]` en `legal-core`. Se miden de nuevo en cada job.
- **`juez`** del informe (la persona; una vez por job): de unos 20 a unos 60 KB con las cifras de la validación; 530 KB
  como mucho. Acotado por 111 respuestas. Una respuesta sin votos afirmativos no deja nada.
- **Los motivos** (la persona y la reparación del cierre): unos 400 B por respuesta marcada; 174 B el del instrumento;
  143 B el de una respuesta sin juzgar. Cada uno deja de darse cuando deja de darse su causa.
- **La medida impresa** (quien lanza su ejecución; una por lanzamiento): 656 B.
- **Las líneas del sondeo** (quien lo lanza): 432 B, más una línea por respuesta sin juzgar.
- **`SKILL.md`** (el modelo; una vez por conversación): 298 líneas.

## Decisiones

- **El mensaje y la orden del voto son los de la validación, literales**, y el voto lo abre un guion (D3, D4).
- **Tope de un voto, 35 s más 5 s; cuatro respuestas a la vez** (D6): es lo que cabe en el tope de un trabajo con el
  peor caso de las sesiones. Su cola no está medida (S6).
- **Los umbrales de las respuestas dependen de que la skill tenga juez, no lista** (D13).
- **La lista sale también de `comprobarConsultaRepetida`** (D14).
- **Dos de los 259 casos necesitan una eval y una derivada retiradas en H7.2: se restauran bajo
  `testdata/evals/retiradas/`** (D15). El spec suponía que todo estaba en `main`.
- **El control de derivaciones de los casos vive en `internal/evals`, con el nombre que da el hito** (D16).
- **Los tests no leen los votos de la evidencia** (D18).
- **El vocabulario nuevo se busca en toda la prosa, con el código en línea** (D22).
- **v0.1.7 quita también los nombres de campo de la línea de la redacción modificada y los del paso 4** (D23, C5 y
  C7): los pide la delimitación de D22.
- **Reparación del cierre: C9, el inciso detrás de una remisión** (D27, V23; contracts/skill-boe-legislacion.md §8):
  con C1-C8, `afirma_lo_no_leido:claude-sonnet-5-5:orden` dio 1 de 54 («del artículo 24, que es previa a ese
  recurso», la regla de un artículo no leído colgada de la remisión del art. 20.5); la viñeta de C1 nombra el inciso y
  la remisión va con las palabras del texto leído. Una línea, pagada con un blanco de «Cómo se cita»: 298 líneas. Su
  efecto lo mide el job de la medición siguiente.
- Las demás, con su alternativa rechazada, en research D1-D27.

## Trazabilidad: cada mecanismo y su requisito

| Mecanismo | Requisito |
|---|---|
| `Sesion.Textos` y su lectura del transcript | FR-001, FR-002 |
| `mensajeDelVoto` | FR-001, FR-002, FR-004 |
| `scripts/evals-voto.sh` y el votante que lo ejecuta, con su tope | FR-003, FR-004, FR-007 |
| La lectura del voto y su validación contra `esquema.json` | FR-007 |
| `fraseEsta` | FR-005 |
| El voto nulo y la regla | FR-006, FR-010, FR-011 |
| La selección de respuestas, con la eval sin binario ni servidor aparte | FR-012, FR-013 |
| `clases.yaml`, su esquema, `Conjunto.Juez` y `LeerConjunto` con `juez/` | FR-020, FR-021 |
| Las cuatro copias y `TestCopiasDelJuez` | FR-022, FR-023, FR-108 |
| Los casos, su reconstrucción en proceso y los dos ficheros restaurados | FR-024, FR-051, FR-109 |
| Los doce umbrales y sus motivos | FR-030 a FR-035 |
| `juez` en `informe.json` y la sección de `informe.md` | FR-060, FR-061 |
| `comprobarLaMedida`, `TestMedidaVersionada` | FR-041, FR-042, FR-105 |
| `ejecutarElJob` con la comprobación delante, e `InstrumentoSinMedir` | FR-043, FR-044 |
| `medirAlJuez`, `TestMedidaDelJuez`, `scripts/evals-medir-juez.sh`, el objetivo de `make` | FR-050 a FR-054 |
| La lista fuera de `Juzgar`, del resultado, del recuento, del sondeo y de `comprobarConsultaRepetida`; los tres subtests retirados | FR-062, FR-070, FR-071 |
| El juez en el sondeo y sus líneas; la versión del equipo desde los transcripts | FR-075, FR-076 |
| `SKILL.md` v0.1.7 (C1-C8) | FR-080 a FR-086 |
| `salida_de_las_herramientas`, su esquema y la prosa con código en línea | FR-085, FR-110 |
| Las dos variables del juez, el segundo Claude Code, el trabajo `medida`, la condición de `tanda` | FR-050, FR-090, FR-091 |
| Los peores casos y los dos topes | FR-092 |
| `TestDefinicionDelJob` con todo ello | FR-112 |
| `CHANGELOG.md`, `CONTRIBUTING.md`, `docs/WORKFLOW.md`, `doc.go` y los comentarios de `evals.yml` | FR-087, FR-095, FR-113 |

Nada del diseño atiende a un estado que no pasa el umbral de materialidad: un transcript, una medida, unos casos o un
informe versionado que no se pueden leer siguen las reglas que ya hay —sesión ilegible, fichero mal formado, test que
falla, error del punto de entrada—, sin caso propio.

## Cambios de `SKILL.md` trazados a la causa (FR-080)

C1-C8 de contracts/skill-boe-legislacion.md §2, cada uno con sus líneas de v0.1.6, lo que cambia, su causa medida (§1)
y su requisito. En resumen: v0.1.6 dice cuándo leer una remisión y no qué hacer con la que no se lee; manda decir lo
que falta y no cómo; y solo enseña «no lo has leído, no lo digas» para la redacción anterior (C1, C2). Y desde H7.4
dice lo que la respuesta lleva de la vigencia con «el sobre» y dos nombres de campo (C3 a C7). C8 es una línea del
presupuesto de 298. C9 (reparación del cierre; contracts §8, research V23 y D27) es la forma c medida en el cierre:
C1 no nombraba el inciso que cuelga de una remisión y sitúa lo remitido, y paga su línea con un blanco de «Cómo se
cita».

## Datos externos

Ninguno (research D25): ni fuente ni grabación nuevas. Ningún manifiesto `grabaciones.json` ni test `TestGrabar*`
cambia, y el paso `grabar_datos` no tiene nada que grabar. Lo que entra en `testdata/` sale de la historia del
repositorio y deriva de una grabación de H4, del BOE, con fila revisada en `docs/SOURCES.md`.

## Orden de implementación (de dentro afuera)

1. **`[datos]`** `schemas/juez-clases.yaml.json`, la carpeta `evals/boe-legislacion/juez/` con `clases.yaml` y sus
   cuatro copias, `Juez`, `LeerConjunto` con la carpeta `juez` y `TestCopiasDelJuez`.
2. `juez.go` sin procesos: los textos de la sesión (`sesion.go`), el mensaje, la frase, la lectura del voto y la regla,
   con sus tests.
3. `scripts/evals-voto.sh` y el votante que lo ejecuta, con `TestOrdenDelVoto`.
4. La medida: `MODELO_DEL_JUEZ` y `VERSION_DE_CLAUDE_CODE_DEL_JUEZ` en `evals.yml`, su lectura en `definicion.go`,
   `comprobarLaMedida` y `TestMedidaVersionada`.
5. La lista fuera de las respuestas (D14), con `TestJuzgarSinLaLista` y los subtests retirados; los umbrales de las
   respuestas pasan a depender del juez.
6. El juez en el informe: `informe.go`, `umbrales.go`, la clave `juez`, los motivos e `informe.md`, con
   `TestInformeConElJuez` y `TestUmbralesDelInforme`.
7. **`[datos]`** Los dos ficheros de `testdata/evals/retiradas/` y su entrada en `TestGrabacionesDerivadas` de
   `internal/app`.
8. Los casos, su reconstrucción, `TestGrabacionesDerivadas` de `internal/evals`, `medirAlJuez` y
   `TestEjecucionDeLaMedida`.
9. El recorrido del job (`ejecucion.go`, `TestEjecucionSinMedir`), los puntos de entrada, `scripts/evals.sh`,
   `scripts/evals-medir-juez.sh`, el `Makefile`, el resto de `evals.yml` y `TestDefinicionDelJob`.
10. El sondeo con el juez.
11. **`[datos]`** `schemas/expresiones-prohibidas.yaml.json` y la clave nueva de la lista, la prosa con el código en
    línea, los casos de `TestProsaDeLaSkill` y `SKILL.md` v0.1.7, en la misma tarea: con la clave nueva, la prosa de
    v0.1.6 da cinco defectos (V14).
12. `CHANGELOG.md`, `CONTRIBUTING.md`, `docs/WORKFLOW.md`, `internal/evals/doc.go` y los comentarios de `evals.yml`.

**Obligaciones para `tasks.md`**: ninguna tarea `[aceptacion]`; cada tarea deja `make ci` en verde; las tareas `[datos]`
son las de los pasos 1, 7 y 11, con las rutas de este plan; cada fila de «Controles de umbral» tiene la tarea que
construye su control, con un test que lo ve fallar; ninguna tarea ni corrección cumple un umbral de FR-030, FR-032 o
FR-033 rebajándolo, dejándolo en `decide: false`, sacando respuestas o casos de su total, ni cambiando la rúbrica, los
casos o la medida (FR-036, FR-045); ninguna tarea escribe en la carpeta de la evidencia ni la nombra por su ruta en su
línea —el paso 1 la nombra como «la evidencia del ADR 0037» y remite a contracts/medida-del-juez.md §3—, ni edita
`specs/011-…` a `specs/016-…` ni `scripts/workflow/` (FR-096); ninguna usa la red, abre una sesión con modelo ni
ejecuta `make evals`, `make evals-sondeo`, `make evals-medir-juez`, `scripts/evals*.sh`, `TestEjecucionDelJob`,
`TestMedidaDelJuez` o `TestSondeo` (FR-093); ninguna publica ni mide en la plataforma; y los escenarios 7 y 9 del
quickstart no son tareas.

## Complexity Tracking

| Desviación | Por qué hace falta | Alternativa más simple rechazada |
|---|---|---|
| «Aceptación e2e: no aplica» (principio III: «cada hito empieza por el test e2e») | El hito no cambia el binario. La aceptación es la del job (SC-001), y cada pieza entra con sus tests en `make ci` | Un guion que ejerciera el binario sin cambios: no describiría la entrega |
| Tareas `[datos]` con código y tests | Con una clave obligatoria en un esquema, el fichero que no la lleva es un fichero mal formado: separar esquema, datos y código deja `make ci` en rojo entre dos tareas. Precedente: H7.2 a H7.4 | Separar esquema, tipo y datos |
| Dos ficheros restaurados en `testdata/evals/retiradas/`, que el spec no prevé | Dos de los 259 casos son de una eval retirada: sin su pregunta y su grafo previo no se pueden reconstruir, y los casos no se pueden cambiar (research D15) | Leerlos de la historia de git al ejecutar (no la hay en el job ni en CI) o escribir la pregunta en el código (a mano) |
| Un trabajo más en el flujo `evals` y un segundo Claude Code en el de siempre | La medida no abre sesiones de evals, tiene su tope y no puede llamarse `evals (<skill>)`; y la versión de los votos se fija aparte de la de las sesiones (FR-050, FR-091, FR-092) | Un paso dentro de `evals`, o una sola versión de Claude Code |
| Dos guiones nuevos en `scripts/` | El voto se abre como una sesión, con el guion como único argumento variable (research D4); la medida imprime su fichero como el job imprime el informe | Los argumentos del voto en Go, que pediría un `//nolint` |

## Comprobación contra la rúbrica del juez (`juez_plan`) y `precheck.sh plan`

- `precheck.sh plan`: existen `plan.md` y `research.md`; ninguno conserva marcas de aclaración pendiente; están
  `## Constitution Check`, la línea «Aceptación e2e» y `## Controles de umbral`, con cada fila nombrando requisitos del
  spec y un control con forma.
- a · Constitution Check: un ítem por principio (I-IX) y por regla de dependencia (R1-R6).
- b · Dependencias: ninguna nueva.
- c · Reglas de dependencia: tabla R1-R6; ni `net/http`, ni SQLite, ni `os.Exit`, ni salida estándar nuevos.
- d · Errores y códigos: ningún código de `kitlegal` nuevo; errores de `internal/evals` con su fichero o su voto; los
  guiones salen con 1.
- e · Tests primero: «Aceptación e2e: no aplica» con el motivo; «Tests nuevos», «Tests existentes que cambian», puntos
  de entrada, fixtures y las tareas `[datos]`.
- f · Alcance: nada fuera del spec. Lo que el spec no preveía y el plan añade para cumplirlo —los dos ficheros
  restaurados— está en Complexity Tracking; lo que se retira de más —la lista en `comprobarConsultaRepetida`— sale de
  FR-070 (research D14).
- g · Sin atajos: ningún `//nolint`, `t.Skip`, TODO ni error silenciado previstos.
- h · Mejor alternativa: cada decisión con la rechazada (D1-D26).
- i · Afirmaciones verificadas: V1-V22 con fichero, línea u orden; M1-M7 medidas en esta sesión; S1-S8 como supuestos,
  con lo que pasa si no se cumplen.
- j · Quickstart ejecutable: órdenes, rutas y nombres de test reales (los que fija el plan); los escenarios 1 a 6 no
  escriben en el árbol; el 7 y el 9 son de una persona y el 8, del workflow.
- k · Datos externos: ninguno.
- l · Autonomía: ninguna tarea para una persona; la medida y el sondeo no los lanza el run; el cierre en la plataforma
  lo hace el workflow.
- m · Uso: «Uso, de fuera adentro» y los contratos, con bytes medidos, veces por job o por pregunta y lo que apaga
  cada señal.
- n · Proporcionalidad: «Trazabilidad»; sin mecanismos para estados sin vía real.
- o · Controles de umbral: los diez que deciden en el cierre y una fila por cada umbral medido en `make ci`.
