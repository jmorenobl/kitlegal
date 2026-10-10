# Implementation Plan: H25 · El juez de `jurisprudencia`: resumir o caracterizar una sentencia que no se ha leído decide

**Branch**: `020-h25-el-juez-de` | **Date**: 2026-10-10 | **Spec**: [spec.md](./spec.md)

**Input**: Feature specification from `/specs/020-h25-el-juez-de/spec.md`

**Modo**: desatendido. Las decisiones técnicas se tomaron con el «Criterio de decisión autónoma» de la constitución y
están en [research.md](./research.md) (D1-D20), cada una con su alternativa rechazada. Lo que se afirma del
repositorio remite a su tabla V; lo medido en esta sesión, a su tabla M; y lo que no se pudo comprobar son los
supuestos S1-S9.

## Summary

H25 no cambia el binario ni el applet `cita`. Lleva a `jurisprudencia` el juez con modelo de H24 y corrige dos frases
de la skill, en seis piezas:

1. **La carpeta del juez** (`evals/jurisprudencia/juez/`): cuatro copias de lo que una persona validó y su
   `clases.yaml`, con `afirma_lo_no_leido`, que decide con 0, y `afirma_que_existe`, que solo se publica. Con ella, el
   código de H24 juzga las respuestas de la skill, comprueba su medida y publica sus doce umbrales sin cambiar: no
   nombra ninguna skill (research V4 a V6).
2. **El campo `sentencia`** del voto, junto a `precepto`, en el informe (research D2).
3. **La reconstrucción de los 249 casos**, que es lo que el código de hoy no sabe hacer: el caso cuyo informe es el de
   un sondeo, el derivado que quita texto de la pregunta, y las órdenes del applet `cita`, con valores con espacios y
   con la entrada estándar (research D3 a D10). Un prototipo con estas reglas resuelve los 249 (research M2).
4. **Cuatro evals** con preguntas del sondeo, y las reglas del conjunto con las diez (research D13, D14).
5. **El trabajo del job**: `jurisprudencia` a cuatro sesiones a la vez en `evals` y en la matriz de `medida`, con los
   topes de hoy (research M6, D12).
6. **`jurisprudencia` v0.1**: el CAPTCHA es de los programas y la respuesta no lo anuncia; el equivalente se da como
   deducido. Es un diff de 9 líneas añadidas y 7 quitadas, que deja 197 (research D16, M11).

No hay ADR nuevo. No se editan `specs/001-…` a `specs/019-…`, la constitución, `scripts/workflow/`, `schemas/`,
`evidencias/`, las seis evals de H23 ni nada de `boe-legislacion` y `legal-core` (FR-005, FR-096). Esto es el plan;
el cierre se aparta de él en un punto, registrado en «Decisiones»: la reparación R1 cambia una viñeta de
`skills/boe-legislacion/SKILL.md`, que pasa a v0.1.8.

## Technical Context

**Language/Version**: Go 1.27 (`go 1.27.0`, `toolchain go1.27.2`), `CGO_ENABLED=0`, `-trimpath`; YAML de GitHub
Actions y de las evals; Markdown de la skill.

**Primary Dependencies**: las fijadas, sin ninguna nueva (research V2): el lector de YAML del repositorio,
`santhosh-tekuri/jsonschema` v6 (cada voto contra `esquema.json`), `stretchr/testify`; biblioteca estándar
(`encoding/json/v2`, `strings`, `io/fs`, `regexp`).

**Storage**: ninguno nuevo. Ficheros del repositorio: la carpeta del juez, cuatro evals, la skill y la definición del
job. La reconstrucción lee, sin escribir, el informe del cierre de H23 y los dos informes de sondeo de la evidencia.

**Testing**: `go test -race` en `make ci`: tablas con subtests con nombre y `t.Parallel`, sesiones y transcripts
sintéticos, informes sintéticos en `t.TempDir()`, un votante que responde según la etiqueta de cada caso y la
reconstrucción en proceso de los casos del repositorio. Cada test vive en el `_test.go` de su fichero. Ningún test
nuevo con la etiqueta `evals`. El job de evals, sobre la cabeza, lo lanza el workflow.

**Target Platform**: el job, `ubuntu-24.04`; `make ci`, macOS y Linux.

**Project Type**: skill (`jurisprudencia`) e instrumento de medida (`internal/evals`, el flujo `evals`).

**Performance Goals**: votos de un modo ≤ 900 s (FR-023); con 7,6 s por voto son unos 228 s (research S3). Peor caso
del trabajo `evals`, 12 577 s, y del de `medida`, 15 505 s (research M6). La reconstrucción de los 249 casos tarda 39
ms en el prototipo (research M2).

**Constraints**: ningún test de `make ci` usa la red ni abre una sesión con modelo; `SKILL.md` < 300 líneas; ningún
`//nolint`; `evidencias/` no se escribe; la rúbrica, los casos, la medida, la orden y el mensaje del voto no cambian
(FR-003, FR-011, FR-030); el producto y `make ci` no necesitan Python.

**Scale/Scope**: 10 evals y 120 sesiones por job de `jurisprudencia`; 60 respuestas juzgadas, con 60 votos si ninguna
se marca; 249 casos etiquetados y 499 votos en una ejecución de la medida.

## Constitution Check

*GATE: pasado antes de la fase 0 y re-evaluado tras el diseño (fase 1): sin violaciones; la única desviación
justificada está en Complexity Tracking.*

### Principios

| Principio | Cómo lo cumple H25 |
|---|---|
| I · Fuentes públicas y frontera humana | No consulta ninguna fuente ni toca `internal/httpx` o `docs/SOURCES.md` (research V24, D19). El CENDOJ no se consulta: la reconstrucción repite órdenes de `cita`, que no pide nada a la red (ADR 0036). La skill deja de anunciar a la persona un CAPTCHA, y sigue sin sortearlo. La medida la lanza y la versiona una persona (FR-050) |
| II · Nada sin cita ni fuente | Es lo que el hito hace medir: que la respuesta no resuma ni caracterice una sentencia que no tenía delante, con umbral 0 (FR-020). Ningún caso, respuesta ni texto se escribe a mano (FR-040), y el fragmento llega a las evals con una orden (FR-066). El sobre no cambia |
| III · Tests primero y offline | «Aceptación e2e: no aplica» (abajo, con el motivo). Cada pieza entra con sus tests en `make ci`, offline y sin modelo. La aceptación es el job de cierre (SC-001) |
| IV · Hexagonal y errores tipados | Ningún paquete del binario ni código de salida de `kitlegal` cambian. `internal/evals` devuelve `error` con el caso o el fichero delante; ningún `panic`, `os.Exit` ni salida estándar desde Go (R4, R5) |
| V · Simplicidad y dependencias | Ninguna dependencia nueva. El juez, los umbrales, la medida y sus guiones se usan como están; solo se escribe lo que el código de hoy no sabe hacer. Cada mecanismo se traza a un FR («Trazabilidad»). Rechazados un paquete aparte, un campo genérico del voto, un segundo camino de reconstrucción y una regla con los valores de las evals (research D1, D2, D4, D13) |
| VI · Un binario, convenciones de agente | El binario, sus banderas, `--describe` y la tabla de comandos de la skill no cambian. Ningún ejecutable nuevo |
| VII · Grafo y privacidad | Nada nuevo en el grafo. La reconstrucción usa una caché y un grafo temporales, que se retiran, y no toca los de quien la ejecuta (FR-044) |
| VIII · Skills primero | Mejora `jurisprudencia` (v0.1) y hace medible lo que H23 dejó sin control. Es la condición para que una release la lleve. `SKILL.md`, 197 líneas, sin `scripts/` y con la tabla de comandos sin drift |
| IX · Genericidad territorial | Nada se particulariza para un municipio: ni el juez, ni las evals, ni la skill |

### Reglas de dependencia (`docs/ROADMAP.md` §2, constitución IV)

| Regla | Cómo la cumple |
|---|---|
| R1 · `internal/core/**` no importa adaptadores | Sin cambios en `internal/core` |
| R2 · Solo `internal/httpx` importa `net/http` | Ningún fichero nuevo lo importa; la reconstrucción no abre ninguna conexión |
| R3 · Solo `cache`/`store`/`graph` importan SQLite y `database/sql` | `internal/evals` no los importa: abre la caché y el grafo de la sesión con las opciones de sus paquetes, como hoy |
| R4 · Solo `cli` y `cmd/` llaman a `os.Exit` | Ningún `os.Exit` nuevo: el código de una orden repetida lo devuelve `app.Main` |
| R5 · Solo `render` escribe en stdout | Ningún `fmt.Print*`, `os.Stdout` ni `os.Stderr` nuevo: la salida de cada orden repetida va a un búfer |
| R6 · `internal/graph` no importa `source/*` ni `render` | Sin cambios |

## Project Structure

### Documentation (this feature)

```text
specs/020-h25-el-juez-de/
├── plan.md                  # este fichero
├── research.md              # V1-V24, M1-M12, S1-S9, la traza de las dos frases, D1-D20
├── data-model.md            # la carpeta, el voto, el caso, las preguntas, el texto pegado, los ficheros
├── quickstart.md            # §0-§9, en make ci; §10, el job de cierre; §11, una persona
├── contracts/
│   ├── juez-de-jurisprudencia.md          # carpeta, mensaje, voto con sentencia, informe, doce umbrales, motivos
│   ├── medida-y-casos.md                  # casos, reconstrucción, derivaciones, ejecución de la medida
│   ├── evals-jurisprudencia.md            # las cuatro evals, las reglas del conjunto, las preguntas
│   ├── job-de-evals.md                    # evals.yml, peores casos, TestDefinicionDelJob
│   ├── skill-jurisprudencia.md            # causa, C1 y C2, lo que se queda, uso
│   └── skill-jurisprudencia-v0.1.diff     # C1 y C2, aplicable con git apply
├── checklists/              # del spec
└── tasks.md                 # la escribe /speckit-tasks
```

### Source Code (repository root)

```text
evals/jurisprudencia/juez/                          # nueva: clases.yaml y las cuatro copias
evals/jurisprudencia/07-resumen-de-una-conocida.yaml
evals/jurisprudencia/08-doctrina-con-el-fallo-delante.yaml
evals/jurisprudencia/09-de-que-trata-con-la-ficha-sola.yaml
evals/jurisprudencia/10-doctrina-dada-por-hecha.yaml
skills/jurisprudencia/SKILL.md                      # v0.1: dos pasajes
.github/workflows/evals.yml                         # concurrencia, matriz de medida, descripción, comentarios
internal/evals/
├── juez.go                  # el campo sentencia del voto
├── informe.go               # la columna del campo propio de la clase en informe.md
├── medida.go                # Quitado con texto; el informe de un sondeo; el texto pegado; las órdenes cita
├── conjunto.go              # las reglas del conjunto con las diez
├── doc.go                   # lo que dice de jurisprudencia
├── umbrales.go              # solo el comentario de umbralesDelInforme, que el hito dejaba falso (T009)
└── juez_test.go, sesion_test.go, informe_test.go, medida_test.go, ejecucion_test.go,
    conjunto_test.go, definicion_test.go, sentencias_test.go (tasks.md, T003), y un comentario
    en umbrales_test.go y en consultas_test.go (T009)
CHANGELOG.md, CONTRIBUTING.md, docs/JURISPRUDENCIA.md, docs/WORKFLOW.md

# Después, con las reparaciones del cierre («Decisiones»; research R1 y R2) y la revisión final que las juzga:
skills/boe-legislacion/SKILL.md                     # v0.1.8: una viñeta (R1)
internal/evals/juez.go, informe.go, doc.go          # el voto que el tope corta se pide otra vez; votos_cortados (R2)
internal/evals/juez_test.go, informe_test.go, medida_test.go, ejecucion_test.go, sondeo_test.go   # (R2)
internal/app/skills_test.go                         # el caso dos-inicios, sin una línea de más (revisión final, R1)
CHANGELOG.md, CONTRIBUTING.md
```

**Structure Decision**: la de `docs/ROADMAP.md` §2 y `CLAUDE.md`, sin paquetes ni ficheros Go nuevos: todo cabe en
los ficheros de `internal/evals` que ya tienen el juez, la medida y las reglas. `ejecucion.go`, `sesion.go`,
`definicion.go`, los guiones de `scripts/` y el `Makefile` no cambian, y de `umbrales.go` no cambia el código: solo
un comentario, que decía que la skill no tiene juez (research V4 a V6, V23; tasks.md, T009). Fuera de
`internal/evals`, el único fichero de Go que cambia es un test de `internal/app`, en un caso, con la revisión final
de la reparación R1.

## Aceptación e2e

**Aceptación e2e: no aplica.** El hito no cambia el binario ni el applet `cita` —cambiarlos está fuera de alcance
(FR-096), y `cmd/kitlegal` no enlaza `internal/evals`—, así que no hay comportamiento de `kitlegal` que un guion
`testscript` describa (research D18). Su aceptación es la de la skill y la del instrumento: el job de evals de cierre
(SC-001). En `make ci` la fijan, sin modelo:

| Historia | Tests | Requisitos |
|---|---|---|
| US1, US2 | `TestVotoDelJuez`, `TestTextosDeLaSesion`, `TestMensajeDelVoto`, `TestUmbralesDeJurisprudencia` | FR-010 a FR-013, FR-020 a FR-025, FR-108, FR-109, FR-111; SC-008, SC-009, SC-013 |
| US3 | `TestCopiasDelJuez`, `TestMedidaVersionada`, `TestEjecucionSinMedir`, `TestEvalsDelRepositorio` | FR-001, FR-002, FR-031, FR-032, FR-102, FR-104; SC-002, SC-004 |
| US4 | `TestArgumentosDeLaInvocacion`, `TestTextoQuitado`, `TestResolverCasos`, `TestReconstruccionDeJurisprudencia`, `TestGrabacionesDerivadas`, `TestEjecucionDeLaMedida`, `TestDefinicionDelJob` | FR-040 a FR-045, FR-050, FR-051, FR-105 a FR-107; SC-005 a SC-007 |
| US5 | `TestPreguntasDelSondeo`, `TestPreguntasConElFragmento`, `TestConjuntoDeEvals`, `TestEvalsDelRepositorio` | FR-060 a FR-066, FR-110; SC-010 |
| US6 | `TestDefinicionDelJob` | FR-070 a FR-072, FR-103; SC-003 |
| US7 | `skills-check` (`TestSkillsDelRepositorio`) | FR-083; SC-011 |

Ninguna tarea `[aceptacion]`.

## Controles mecánicos que este hito añade o toca

### Objetivos del `Makefile`

Ninguno nuevo ni cambiado. `test` y `skills-check` llevan más dentro: los tests nuevos, y en
`TestEvalsDelRepositorio`, la carpeta del juez de `jurisprudencia` y sus diez evals. `schema-check` no cambia.
`evals-medir-juez` no cambia: recibe la skill.

### Tests nuevos

| Test | Fichero | Cubre |
|---|---|---|
| `TestTextoQuitado` | `medida_test.go` | FR-043 |
| `TestReconstruccionDeJurisprudencia` | `medida_test.go` | FR-040 a FR-044, FR-105; SC-005 |
| `TestPreguntasDelSondeo` | `conjunto_test.go` | FR-060, FR-110; SC-010 |

### Tests existentes que cambian

- `TestCopiasDelJuez`: la fila de `jurisprudencia` en su tabla.
- `TestMedidaVersionada` y `TestEjecucionSinMedir`: sus seis mutaciones, por cada fila de la tabla y no solo por la
  primera (research V13).
- `TestEvalsDelRepositorio/conjunto-jurisprudencia`: la skill tiene juez, con sus dos clases, y sus diez evals cumplen
  las reglas. `TestConjuntoDeEvals/jurisprudencia`: las reglas con las diez y sus mutaciones.
- `TestUmbralesDeJurisprudencia`: de cuatro umbrales a doce, con un votante de pega, las 30 respuestas por modo y los
  casos de FR-111; deja de exigir que el informe no lleve `juez`.
- `TestVotoDelJuez`: un caso con `sentencia` y el esquema de esta skill.
- `TestTextosDeLaSesion` y `TestMensajeDelVoto`: una sesión de cada modo con una orden o una llamada `cita` sobre la
  pregunta con el fragmento. Lo que ya esperan de la orden y del mensaje no cambia (FR-011).
- `TestArgumentosDeLaInvocacion` y `TestResolverCasos`: los casos de las órdenes `cita`, del informe de un sondeo y
  del derivado por texto. Los de H24, sin cambiar.
- `TestGrabacionesDerivadas` y `TestEjecucionDeLaMedida`: por skill. Lo que esperan de `boe-legislacion` no cambia.
- `TestPreguntasConElFragmento`: además, la eval (h) con el fragmento y la (i) con su ficha.
- `TestDefinicionDelJob`: `del-repositorio`, `sinteticas`, `peor-caso` y `tope-con-las-evals-del-repositorio`, con lo
  de contracts/job-de-evals.md §4.
- Y cuatro que esta lista no nombraba y fija tasks.md: `TestInformeConElJuez` y `TestInformeMarkdownDeLosUmbrales`,
  con el voto que lleva `sentencia` y su columna (T001); `TestLeerCasosEtiquetados`, con un derivado con `texto`
  (T002); y `TestJuzgarSentencias`, con las sesiones modelo de las diez evals (T003).
- Con la reparación R2 del cierre: `TestVotoDelJuez` y `TestOrdenDelVoto`, con el voto que el tope corta y se pide
  otra vez; `TestInformeConElJuez` y `TestInformeMarkdownDeLosUmbrales`, con `votos_cortados` y su tabla;
  `TestJuicioDelSondeo`, `TestSalidaDelSondeo` y `TestEjecucionDeLaMedida`, con el tope agotado dos veces o un voto
  cortado que llega al repetirse; y lo que `ejecucion_test.go` exige de la clave `juez` de un informe sin votos.
- Con la revisión final de la reparación R1: el caso `dos-inicios` de `TestSkillsDelRepositorio`
  (`internal/app/skills_test.go`), que pone la marca repetida en el lugar de una línea en blanco y no en una línea
  más, para que valga con una skill de 299 líneas.

Ninguno se desactiva ni se salta. Ninguno se retira sin el que lo sustituye.

### Puntos de entrada fuera de `make ci` (etiqueta `evals`)

`TestEjecucionDelJob`, `TestMedidaDelJuez` y `TestSondeo` no cambian. Ninguna tarea los ejecuta.

### Fixtures, `testdata/` y `schemas/`

Nada cambia en `schemas/` ni en `testdata/`: ninguna tarea `[datos]` (research D17). Los tests leen, sin escribir, la
evidencia de la validación del juez, el fragmento y el informe del cierre de H23, y escriben lo sintético en su
directorio temporal. Ninguna tarea escribe en `evidencias/`.

### CI

`.github/workflows/evals.yml` (contracts/job-de-evals.md): lo comprueba `TestDefinicionDelJob` en `make ci`. `ci.yml`
no cambia. El job de evals se ejecuta en la propuesta de cambio, tras la revisión final, y lo lanza el workflow
(FR-112).

## Controles de umbral

Los diez umbrales que deciden en el informe del cierre de `jurisprudencia` —los seis nuevos que fija FR-101 y los
cuatro que siguen (FR-024), que SC-001 cuenta— y, como pide el ADR 0029, una fila por cada umbral que se mide en
`make ci`. `afirma_que_existe` no tiene fila: solo se publica (FR-021).

| Requisito | Umbral | Control | Dónde |
|---|---|---|---|
| FR-020, SC-001 (modo orden) | 0 respuestas marcadas en `afirma_lo_no_leido`, de las juzgadas del modo | umbral del informe con `decide: true`: con una, motivo con la sesión y sus tres frases, veredicto `fallo` y el trabajo en rojo, que el cierre cuenta | `evals:jurisprudencia:afirma_lo_no_leido:claude-sonnet-5-5:orden` |
| FR-020, SC-001 (modo herramienta) | ídem | ídem | `evals:jurisprudencia:afirma_lo_no_leido:claude-sonnet-5-5:herramienta` |
| FR-022, SC-001 (defectos) | 0 defectos sin marcar, de 125 | umbral del informe con `decide: true`, con el recuento de la medida versionada; con otro valor, `make ci` falla y el job termina en `fallo` sin abrir sesiones | `evals:jurisprudencia:medida_del_juez:afirma_lo_no_leido:defectos_sin_marcar` |
| FR-022, SC-001 (correctos) | 0 correctos marcados, de 124 | ídem | `evals:jurisprudencia:medida_del_juez:afirma_lo_no_leido:correctos_marcados` |
| FR-023, SC-001 (modo orden) | ≤ 900 s de votos del modo | umbral del informe con `decide: true`: por encima, motivo de la ejecución y `fallo` | `evals:jurisprudencia:duracion_del_juez:orden` |
| FR-023, SC-001 (modo herramienta) | ídem | ídem | `evals:jurisprudencia:duracion_del_juez:herramienta` |
| FR-024, SC-001 (cita sin documento, modo orden) | 0 respuestas con una cita sin documento cotejado o un ECLI sin origen | el de hoy, que sigue, ahora sobre las diez evals | `evals:jurisprudencia:cita_sin_documento:claude-sonnet-5-5:orden` |
| FR-024, SC-001 (cita sin documento, modo herramienta) | ídem | ídem | `evals:jurisprudencia:cita_sin_documento:claude-sonnet-5-5:herramienta` |
| FR-024, SC-001 (sin activar, modo orden) | 0 respuestas sin la skill activada | el de hoy, que sigue | `evals:jurisprudencia:sin_activar:claude-sonnet-5-5:orden` |
| FR-024, SC-001 (sin activar, modo herramienta) | ídem | ídem | `evals:jurisprudencia:sin_activar:claude-sonnet-5-5:herramienta` |
| SC-001 (sin juzgar) | 0 respuestas sin juzgar | con alguna, motivo de la ejecución y `fallo` en el job, como en H24; el test lo ve fallar con un voto que no llega. Desde la reparación R2 del cierre, el voto que el tope corta se pide otra vez, una sola, y no llega si lo corta también | `ci:internal/evals/informe_test.go:TestInformeConElJuez` |
| FR-001, FR-102, SC-002 | 4 de 4 copias idénticas | comparación byte a byte de la fila de la skill; con un byte cambiado o sin una, el defecto que la nombra | `ci:internal/evals/medida_test.go:TestCopiasDelJuez` |
| FR-072, FR-103, SC-003 | cada `timeout-minutes` cubre su peor caso (12 577 s bajo 352 min y 15 505 s bajo 269); 2 de 2 mutaciones «de una en una» en rojo | la definición del repositorio y las sintéticas, con las evals y los casos del repositorio | `ci:internal/evals/definicion_test.go:TestDefinicionDelJob` |
| FR-031, FR-104, SC-004 | la medida del repositorio corresponde, con 0 y 0; 6 de 6 mutaciones en rojo, en `make ci` y en el job | la comprobación sobre el repositorio y sobre copias cambiadas; el recorrido del job con quien abre sesiones contado | `ci:internal/evals/medida_test.go:TestMedidaVersionada`, `ci:internal/evals/ejecucion_test.go:TestEjecucionSinMedir` |
| FR-044, FR-105, SC-005 | 249 de 249 casos resueltos (138, 39 y 72); 0 órdenes `cotejar` en los 45 sin el documento | reconstrucción en proceso de los casos del repositorio | `ci:internal/evals/medida_test.go:TestReconstruccionDeJurisprudencia` |
| FR-045, FR-106, SC-006 | 249 respuestas iguales a las de su informe; 127 derivados que solo pierden lo quitado | el control de derivaciones, con los casos de la skill | `ci:internal/evals/medida_test.go:TestGrabacionesDerivadas` |
| FR-051, FR-107, SC-007 | 0 de 125 y 0 de 124 con 499 votos; 1 defecto sin marcar o 1 correcto marcado da error con el caso | votante según la etiqueta, que cuenta sus llamadas | `ci:internal/evals/medida_test.go:TestEjecucionDeLaMedida` |
| FR-013, FR-108, SC-008 | 1 de 1 voto afirmativo de la clase que decide, publicado con `sentencia` | el voto contra el esquema de la skill; el informe con un votante de pega | `ci:internal/evals/juez_test.go:TestVotoDelJuez`, `ci:internal/evals/informe_test.go:TestUmbralesDeJurisprudencia` |
| FR-010, FR-109, SC-009 | 0 bytes de `SKILL.md`, de la eval o del juicio sin modelo en el mensaje, en los 2 modos | los textos de la sesión y el mensaje, byte a byte | `ci:internal/evals/sesion_test.go:TestTextosDeLaSesion`, `ci:internal/evals/juez_test.go:TestMensajeDelVoto` |
| FR-065, FR-066, FR-110, SC-010 | 4 de 4 preguntas iguales a las de `preguntas.json`; el fragmento y la ficha byte a byte; 10 evals que cumplen las reglas, y con 9 o con 11, no | las preguntas leídas con `LeerEval`; las reglas del conjunto | `ci:internal/evals/conjunto_test.go:TestPreguntasDelSondeo`, `ci:internal/evals/conjunto_test.go:TestPreguntasConElFragmento`, `ci:internal/evals/conjunto_test.go:TestConjuntoDeEvals`, `ci:internal/evals/conjunto_test.go:TestEvalsDelRepositorio` |
| FR-083, SC-011 | `SKILL.md` < 300 líneas y sin drift | `TestSkillsDelRepositorio`, en `skills-check` | `ci:Makefile:skills-check` |
| FR-111, SC-013 | 12 elementos, 10 que deciden; 1 marcada, `fallo`; 0, se cumple; 3 en `afirma_que_existe`, mismo veredicto; 901 s, `fallo` | sesiones sintéticas de las diez evals y un votante de pega | `ci:internal/evals/informe_test.go:TestUmbralesDeJurisprudencia` |

SC-012 es `make ci` entero; SC-014 lo mide una persona; SC-015 lo comprueba la revisión final por lectura (FR-093): no
tienen fila.

## Uso, de fuera adentro (criterio de uso, ADR 0028)

El detalle, con ejemplos y bytes medidos (research M10), está en contracts/juez-de-jurisprudencia.md §8,
contracts/medida-y-casos.md §10, contracts/evals-jurisprudencia.md §5, contracts/job-de-evals.md §6 y
contracts/skill-jurisprudencia.md §4. Nada de lo que el hito entrega crece con lo consultado en meses de uso: el applet
`cita` no pide nada a la red y no guarda nada, y el job y la medida parten de las evals y de los casos del repositorio.

- **La respuesta de la skill** (la persona; una por pregunta): las órdenes por pregunta son las de v0 —un
  `cita preparar` por sentencia que no está delante, un `cita cotejar` por documento, un `cita preparar` con texto por
  pregunta por materia—. De 768 a 2 695 bytes en el sondeo. Por sentencia sin documento, la línea y su consulta, unos
  360 bytes; salen mientras la sentencia no esté delante, y con su documento cotejado las sustituye la cita. Ninguna
  frase sobre un CAPTCHA.
- **`umbrales`** (el job, el informe final, la persona; una vez por job): doce elementos, 3 767 bytes con todas sus
  medidas en 0, fijos. Se miden de nuevo en cada job.
- **`juez`** del informe, con `sentencia` (la persona; una vez por job): 1 619 bytes por respuesta con sus tres votos;
  nada si ningún voto dice sí; unos 195 KB como mucho. Acotado por 60 respuestas.
- **Los motivos** (la persona y la reparación del cierre): 374 bytes por respuesta marcada; 128 el del instrumento sin
  medir. Cada uno deja de darse cuando deja de darse su causa.
- **La medida impresa** (quien pone la etiqueta; una por lanzamiento y por skill): unos 650 bytes; si no se cumple,
  unos 280 más por caso mal juzgado.
- **Las líneas de `TestDefinicionDelJob`, de las reglas del conjunto y de un caso que no se resuelve** (quien cambia
  el job, las evals o el reconstructor; en `make ci`): una línea por defecto, hasta que se corrige.
- **`SKILL.md`** (el modelo; una vez por conversación): 197 líneas.

## Decisiones

- **El juez, los umbrales y la medida de H24 se usan como están**: con la carpeta del juez, el código ya juzga, mide
  y publica para esta skill (research V4 a V6, V23).
- **`sentencia` es un segundo campo opcional del voto**, y la columna de `informe.md` toma su nombre de los votos que
  lleva (D2).
- **`quitado` gana `texto`**; `regla` no se lee (D3).
- **Las órdenes `cita` se repiten en proceso, con su propia regla de argumentos y con la entrada estándar**; las de
  `boe` y `graph`, como en H24 (D4, D5).
- **El código de la orden repetida se compara con el del informe en las órdenes `cita`** (D7): sin ello, el
  reconstructor de hoy da por buenos textos de error (M3).
- **Un informe es de un sondeo por su clave `sondeo`**, y la pregunta de una eval se lee de las evals de hoy (D8, D9).
- **Los tests no leen los votos de la evidencia** (D11).
- **La matriz de `medida` esperada se deriva de las skills con carpeta de juez** (D12).
- **Las reglas del conjunto cuentan las diez por clase, sin valores** (D13).
- **Las cuatro evals se numeran 07 a 10** (D14) y las dos con texto pegado se componen con `cp`, `sed` y `printf`
  (D15).
- **Ninguna tarea `[datos]`** (D17).
- Las demás, con su alternativa rechazada, en research D1-D20.
- **Reparaciones del cierre** (FR-027), las dos con la medición 2 del cierre como evidencia (research «Reparaciones
  del cierre», R1 y R2), **y las dos apartándose de lo que piden la sección del hito y el spec**:
  - **`boe-legislacion` v0.1.8**, porque el juez marca en `afirma_lo_no_leido` dos respuestas de la eval 15 que ponen
    la materia del art. 7 con «sobre» en el aviso final de las remisiones no seguidas, 1 de 54 en cada modo (R1). Se
    aparta de «Fuera de alcance» de la sección del hito, que nombra a `boe-legislacion`, y de FR-005, que manda dejar
    su `SKILL.md` como está. Se hace porque la aceptación pide que su veredicto siga aprobado y el umbral que no se
    cumple es de la skill, no de la ejecución; se descarta no tocarla y volver a medir, con la misma construcción ya
    vista en dos sesiones independientes. `SKILL.md` queda en 299 líneas, su máximo, con un solo trozo de diff con
    `main`; para que `make ci` las admita, el caso `dos-inicios` de `TestSkillsDelRepositorio`
    (`internal/app/skills_test.go`) deja de añadir una línea a la copia que altera (revisión final, ronda 3).
  - **El voto que el tope corta se pide otra vez, una sola**, dentro de las dos peticiones por voto que ya tenía el
    nulo, con el corte publicado en `juez.votos_cortados` y en informe.md, porque el job de `jurisprudencia` salió en
    rojo con sus doce umbrales cumplidos por un voto cortado (R2). Se aparta de lo que H24 dejó sin reintento y a la
    decisión de una persona (H24 research D6); a ella le queda decidir si lo quiere así.
  - Ninguna cambia el peor caso, los topes, una eval, la rúbrica, los casos, la medida ni un umbral.

## Trazabilidad: cada mecanismo y su requisito

| Mecanismo | Requisito |
|---|---|
| La carpeta del juez, con sus cuatro copias y `clases.yaml`; la fila de `TestCopiasDelJuez` | FR-001 a FR-004, FR-030, FR-102 |
| `Sentencia` en el voto y en lo leído de él; la columna de `informe.md` | FR-013, FR-108 |
| `Quitado` con `texto`, y el nombre del derivado en un error | FR-040, FR-043 |
| La lectura del informe de un sondeo y de `preguntas.json`, con `{fragmento}` y `{ficha}` | FR-042 |
| El texto pegado y sus tres recortes | FR-043 |
| `AppletCita` en el registro de la sesión; los argumentos de una orden `cita`; la entrada estándar de `cotejar` | FR-041, FR-044 |
| El `codigo` de la invocación del informe y su comparación | FR-041 |
| El recuerdo por informe, sesión y texto quitado | FR-043 |
| Las reglas del conjunto con las diez, y la dirección en «número y fecha» | FR-061, FR-065 |
| Las cuatro evals | FR-060 a FR-066 |
| La concurrencia de `evals`, `jurisprudencia` en la matriz de `medida` y la descripción de la entrada | FR-050, FR-070, FR-071 |
| Lo que `TestDefinicionDelJob` exige de más: la matriz de `medida` y su concurrencia | FR-072, FR-103 |
| Las mutaciones de la medida por cada fila de la tabla | FR-031, FR-032, FR-104 |
| Los doce umbrales en `TestUmbralesDeJurisprudencia` | FR-020 a FR-025, FR-111 |
| Los tests del mensaje y de los textos con una sesión de `cita` | FR-010 a FR-012, FR-109 |
| Los tests de la reconstrucción, de las derivaciones y de la medida con los casos de la skill | FR-045, FR-051, FR-105 a FR-107 |
| Los tests de las preguntas | FR-066, FR-110 |
| `SKILL.md` v0.1 (C1 y C2) | FR-080 a FR-083 |
| `CHANGELOG.md`, `docs/JURISPRUDENCIA.md`, `docs/WORKFLOW.md`, `CONTRIBUTING.md` | FR-084, FR-090 a FR-093 |
| `doc.go` y los comentarios de `evals.yml` que el hito deja falsos | FR-024, FR-070 |

Nada del diseño atiende a un estado que no pasa el umbral de materialidad: un informe, un `preguntas.json` o unos
casos que no se pueden leer, y un `quitado` que no dice qué quitar, siguen la regla que ya hay —el caso no se
resuelve, con un error que lo nombra—, sin caso propio.

## Cambios de `SKILL.md` trazados a la causa (FR-080)

C1 y C2 de contracts/skill-jurisprudencia.md §2, cada uno con su pasaje de v0, su texto de v0.1 y su causa medida.

- **C1, el CAPTCHA.** v0 dice «el buscador pide un CAPTCHA a los programas, y no se sortea», y 8 de las 35 respuestas
  del cierre de H23 y 6 de las 18 del sondeo lo nombran; cinco se lo anuncian a la persona o le piden que lo resuelva.
  v0.1 dice que el obstáculo es de los programas, que a la persona no le sale y que la respuesta no lo anuncia.
- **C2, el equivalente.** v0 dice «puedes nombrarlo junto a la referencia», y cinco de las seis respuestas del cierre
  de H23 que lo dan no dicen que es deducido. v0.1 pide darlo como deducido de la referencia y sin comprobar.

Las sesiones y sus frases, una a una, están en research.md, «Traza de las dos frases».

## Datos externos

Ninguno (research D19): ni fuente ni grabación nuevas. Ningún manifiesto `grabaciones.json` ni test `TestGrabar*`
cambia, y el paso `grabar_datos` no tiene nada que grabar. El CENDOJ no se consulta. El fragmento, los informes de los
sondeos y las preguntas ya están versionados en `evidencias/`, que el run solo lee.

## Orden de implementación (de dentro afuera)

1. **El campo `sentencia`** (`juez.go`, `informe.go`), con un juez sintético cuyo esquema lo lleva: `TestVotoDelJuez`
   y la columna de `informe.md`.
2. **La reconstrucción** (`medida.go`): `Quitado` con `texto`, el informe de un sondeo y sus preguntas, el texto
   pegado, las órdenes `cita` con su entrada estándar y su código. Con `TestArgumentosDeLaInvocacion`,
   `TestTextoQuitado` y `TestResolverCasos`, sobre informes sintéticos: no necesita la carpeta del juez.
3. **Las cuatro evals, las reglas del conjunto con las diez y `jurisprudencia` a cuatro en el trabajo `evals`**, en
   una tarea: con diez evals de una en una, el tope no cubre el peor caso (33 397 s, con la fórmula de
   `definicion.go`). Con `TestPreguntasDelSondeo`, `TestPreguntasConElFragmento`, `TestConjuntoDeEvals`,
   `TestEvalsDelRepositorio`, `TestDefinicionDelJob` y las sesiones sintéticas de `TestUmbralesDeJurisprudencia`.
4. **La carpeta del juez, su fila en la tabla de las copias y `jurisprudencia` en la matriz de `medida`**, en una
   tarea: la carpeta sola deja cuatro tests en rojo (research M4). Con `TestCopiasDelJuez`, las mutaciones por fila de
   `TestMedidaVersionada` y `TestEjecucionSinMedir`, los doce umbrales de `TestUmbralesDeJurisprudencia`,
   `TestTextosDeLaSesion`, `TestMensajeDelVoto` y `TestDefinicionDelJob`.
5. **Los tests con los casos del repositorio**: `TestReconstruccionDeJurisprudencia`, `TestGrabacionesDerivadas` y
   `TestEjecucionDeLaMedida`.
6. **`SKILL.md` v0.1**, con el diff del contrato.
7. **La documentación**: `CHANGELOG.md`, `CONTRIBUTING.md` (los pasajes de FR-093, hoy en torno a sus líneas 625,
   702, 922, 942, 955, 1042, 1084, 1107 y 1129), `docs/JURISPRUDENCIA.md` (su línea 228), la fila «Evidencia de un
   ADR» de `docs/WORKFLOW.md`, `internal/evals/doc.go` y los comentarios de `evals.yml`.

**Obligaciones para `tasks.md`**: ninguna tarea `[aceptacion]` ni `[datos]`; cada tarea deja `make ci` en verde, y por
eso los pasos 3 y 4 son una tarea cada uno; cada fila de «Controles de umbral» tiene la tarea que construye su
control, con un test que lo ve fallar; ninguna tarea ni corrección cumple un umbral de FR-020, FR-022 o FR-023
rebajándolo, dejándolo en `decide: false`, sacando respuestas o casos de su total, ni cambiando la rúbrica, los casos
o la medida (FR-026); ninguna tarea escribe en `evidencias/` ni nombra en su línea una ruta de esa carpeta —la nombra
como «la evidencia de la validación del juez» o «el fragmento» y remite al contrato—; ninguna edita las seis evals de
H23, `evals/boe-legislacion/`, `evals/legal-core/`, el applet `cita`, `schemas/`, `specs/001-…` a `specs/019-…`, la
constitución ni `scripts/workflow/` (FR-005, FR-096); ninguna usa la red, abre una sesión con modelo ni ejecuta
`make evals`, `make evals-sondeo`, `make evals-medir-juez`, `scripts/evals*.sh`, `TestEjecucionDelJob`,
`TestMedidaDelJuez`, `TestSondeo` o los guiones de Python de la evidencia (FR-095); la tarea de documentación lleva
`CONTRIBUTING.md` entre sus rutas (FR-093); ninguna publica ni mide en la plataforma; y los escenarios 10 y 11 del
quickstart no son tareas.

## Complexity Tracking

| Desviación | Por qué hace falta | Alternativa más simple rechazada |
|---|---|---|
| «Aceptación e2e: no aplica» (principio III: «cada hito empieza por el test e2e») | El hito no cambia el binario ni el applet `cita`. La aceptación es la del job (SC-001), y cada pieza entra con sus tests en `make ci`. Precedente: H24 | Un guion que ejerciera el binario sin cambios: no describiría la entrega |

## Comprobación contra la rúbrica del juez (`juez_plan`) y `precheck.sh plan`

- `precheck.sh plan`: existen `plan.md` y `research.md`; ninguno conserva marcas de aclaración pendiente; están
  `## Constitution Check`, la línea «Aceptación e2e» y `## Controles de umbral`, con cada fila nombrando requisitos
  del spec y un control con forma.
- a · Constitution Check: un ítem por principio (I-IX) y por regla de dependencia (R1-R6).
- b · Dependencias: ninguna nueva (research V2).
- c · Reglas de dependencia: tabla R1-R6; ni `net/http`, ni SQLite, ni `os.Exit`, ni salida estándar nuevos.
- d · Errores y códigos: ningún código de `kitlegal` nuevo; errores de `internal/evals` con su caso o su fichero.
- e · Tests primero: «Aceptación e2e: no aplica» con el motivo; «Tests nuevos», «Tests existentes que cambian»,
  puntos de entrada y fixtures.
- f · Alcance: nada fuera del spec. `doc.go` y los comentarios de `evals.yml` cambian porque el hito los deja falsos.
- g · Sin atajos: ningún `//nolint`, `t.Skip`, TODO ni error silenciado previstos.
- h · Mejor alternativa: cada decisión con la rechazada (research D1-D20).
- i · Afirmaciones verificadas: V1-V24 con fichero y línea u orden; M1-M12 medidas en esta sesión; S1-S9 como
  supuestos, con lo que pasa si no se cumplen.
- j · Quickstart ejecutable: órdenes, rutas y nombres de test reales (los que fija el plan); dice qué vale ya hoy; los
  escenarios 0 a 9 solo dejan ficheros que git ignora; el 10 es del workflow y el 11, de una persona.
- k · Datos externos: ninguno.
- l · Autonomía: ninguna tarea para una persona; la medida y el sondeo no los lanza el run; el cierre en la
  plataforma lo hace el workflow.
- m · Uso: «Uso, de fuera adentro» y los contratos, con bytes medidos, veces por job o por pregunta y lo que apaga
  cada señal.
- n · Proporcionalidad: «Trazabilidad»; el código de H24 se usa como está; sin mecanismos para estados sin vía real.
- o · Controles de umbral: los diez que deciden en el cierre y una fila por cada umbral medido en `make ci`.
