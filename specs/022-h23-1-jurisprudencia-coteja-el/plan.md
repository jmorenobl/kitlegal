# Implementation Plan: H23.1 · `jurisprudencia` coteja el documento antes de decir nada de él, y un umbral sin modelo lo hace cumplir

**Branch**: `022-h23-1-jurisprudencia-coteja-el` | **Date**: 2026-10-16 | **Spec**: [spec.md](./spec.md)

**Input**: Feature specification from `/specs/022-h23-1-jurisprudencia-coteja-el/spec.md`

**Modo**: desatendido. Las decisiones técnicas se tomaron con el «Criterio de decisión autónoma» de la constitución y
están en [research.md](./research.md) (D1-D16), cada una con su alternativa rechazada. Lo que se afirma del
repositorio remite a su tabla V; lo medido en esta sesión, sobre un prototipo fuera del repositorio, a su tabla M; y lo
que no se pudo comprobar en esta sesión son los supuestos S1-S4. Ninguna tarea parte del prototipo ni lo
necesita: el plan y sus contratos dicen todo lo que hay que construir.

## Summary

En la eval 06, 6 de 48 respuestas de `claude-sonnet-5-5` dan por buena una sentencia que no es la pedida sin ejecutar
`cita cotejar`, y ningún control lo hace fallar. H23.1 lo corrige en tres piezas:

1. **El umbral** (`internal/evals`): por cada modo, `documento_sin_cotejar:<modelo>:<modo>` cuenta las sesiones de una
   eval que pide `cita cotejar` sin ningún cotejo que consultara y terminara bien —como orden o como llamada a
   `cita_cotejar`, con el mismo criterio con el que el juicio da por ejecutado un comando esperado—, sobre las
   respuestas del `total` de `sin_activar` de ese modo en esas evals. Decide con 0; su motivo nombra cada sesión. Va
   detrás de `cita_sin_documento` de su modo: `umbrales` de `jurisprudencia` pasa de trece, nueve que deciden, a
   quince, once que deciden ([contracts/umbral-documento-sin-cotejar.md](./contracts/umbral-documento-sin-cotejar.md)).
2. **La skill** `jurisprudencia` v0.1.1: los pasos 1 y 3 de `SKILL.md` dicen que el cotejo de la ficha es lo primero
   que se hace con un documento de la conversación, que nada se dice de él antes, y que si es la pedida lo dice
   `es_la_pedida`, nunca una comparación a ojo ([contracts/skill-jurisprudencia.md](./contracts/skill-jurisprudencia.md)).
3. **Los documentos**: `CHANGELOG.md` (*Unreleased*), `CONTRIBUTING.md`, `docs/JURISPRUDENCIA.md` y `CLAUDE.md`
   ([contracts/documentos.md](./contracts/documentos.md)); `docs/WORKFLOW.md` no dice nada que el hito deje falso.

El cambio de código son unas 165 líneas de producción en dos ficheros que ya existen, `informe.go` y `umbrales.go`, sin
tipo exportado, paquete ni dependencia nuevos, y los comentarios que el cambio deja incompletos. Un prototipo lo
confirma: con él, los quince elementos y once que deciden; 1 sesión sin cotejo de 12 da `fallo` con su motivo en cada
modo; la sesión sin medir y la de Haiku quedan fuera; los dos mutantes hacen fallar los casos; y el resto de
`internal/evals`, `-race` y el linter fijado, en verde (research M2 a M8). Ni el sondeo, ni el juicio de las sesiones, ni
la definición del job, ni las evals cambian.

No hay ADR nuevo. No se tocan `boe-legislacion`, `legal-core`, ninguna eval, `schemas/`, `evidencias/`, `testdata/`,
`.github/`, `scripts/`, la constitución, `docs/ROADMAP.md`, los ADR ni el binario fuera de `internal/evals` (FR-032 a
FR-034, FR-045, FR-052).

## Technical Context

**Language/Version**: Go 1.27 (`go 1.27.0`, `toolchain go1.27.2`), `CGO_ENABLED=0`, `-trimpath` (research V1);
Markdown de `SKILL.md` y de los documentos.

**Primary Dependencies**: las fijadas, sin ninguna nueva. El código usa `fmt`, `slices` y `strings`, que los dos
ficheros ya importan (research V2); los tests, `stretchr/testify`.

**Storage**: ninguno. Ningún fichero de datos cambia.

**Testing**: `go test -race -shuffle=on` en `make ci`: los casos de `TestUmbralesDeJurisprudencia` escriben con
`EscribirInforme` las sesiones sintéticas del plan del job en `t.TempDir()`, con las evals y la carpeta del juez del
repositorio y un votante de votos grabados; `TestEvalsDelRepositorio` lee las evals del repositorio. Ningún test con la
etiqueta `evals`. El job de evals, sobre la cabeza, lo lanza el workflow.

**Target Platform**: el job, `ubuntu-24.04`; `make ci`, macOS y Linux.

**Project Type**: instrumento de medida de las skills (`internal/evals`, que el binario no enlaza, research V14), una
skill sin código (`skills/jurisprudencia/SKILL.md`) y documentación.

**Performance Goals**: ninguno nuevo. El hito no añade ninguna sesión ni ningún voto (FR-032): las cotas y los peores
casos del job son los de hoy. La medida recorre una vez las invocaciones de cada una de las 60 respuestas del modelo que
decide por job.

**Constraints**: ningún test de `make ci` usa la red ni abre una sesión con modelo; ningún `//nolint` ni `t.Skip`;
`evidencias/`, `schemas/` y `testdata/` no cambian; la rúbrica, los casos, la medida, el modelo del juez y su versión no
cambian; `internal/evals/sondeo.go` y `scripts/evals-sondeo.sh` no se tocan (FR-032, FR-033).

**Scale/Scope**: 12 respuestas por modo en el `total` del umbral; 15 elementos de `umbrales` en `jurisprudencia`;
4 casos nuevos de un test y 1 subtest nuevo de otro; `SKILL.md`, de 197 líneas a unas 207.

## Constitution Check

*GATE: pasado antes de la fase 0 y re-evaluado tras el diseño (fase 1): sin violaciones; la única desviación
justificada está en Complexity Tracking.*

### Principios

| Principio | Cómo lo cumple H23.1 |
|---|---|
| I · Fuentes públicas y frontera humana | No consulta ninguna fuente ni toca `internal/httpx` o `docs/SOURCES.md`. La skill sigue sin consultar el CENDOJ: coteja el documento que trae la persona, sin red (ADR 0036). El sondeo de SC-010 lo lanza una persona, fuera del run |
| II · Nada sin cita ni fuente | Es lo que el hito protege: una sentencia solo se cita con su documento cotejado, y si es la pedida lo dice lo que devuelve la operación (FR-020 a FR-022). Ningún sobre cambia. La regla «nunca inventar contenido legal» sigue en la skill, con sus reglas sin tocar (FR-024) |
| III · Tests primero y offline | «Aceptación e2e: no aplica» (abajo, con el motivo). El umbral entra con sus casos en `make ci`, offline y sin modelo; la forma de la skill, con `make skills-check`. La aceptación es el job de cierre (SC-001) |
| IV · Hexagonal y errores tipados | Ningún paquete del binario, ningún error tipado y ningún código de salida de `kitlegal` cambian. El job sigue saliendo en rojo con el veredicto `fallo` |
| V · Simplicidad y dependencias | Ninguna dependencia nueva. Sin tipo ni función exportados nuevos: tres campos y cuatro métodos sin exportar en tipos que ya existen (`respuestaDelJuez`, `respuestasMedidas`, `grupoJuzgado`), cuatro funciones sin exportar, tres constantes y una variable, en dos ficheros que ya existen. Ninguna función de test nueva: cuatro casos en una tabla y un subtest. Cada mecanismo se traza a un FR («Trazabilidad») |
| VI · Un binario, convenciones de agente | El binario, sus banderas, `--describe` y la tabla de comandos de la skill no cambian (FR-023) |
| VII · Grafo y privacidad | Nada del grafo cambia. Los tests usan directorios temporales |
| VIII · Skills primero | Corrige una skill, `jurisprudencia` v0.1.1, y protege su promesa con un control que decide (spec, «Resumen»). `SKILL.md` sigue sin código, por debajo de 300 líneas, con su tabla generada y sin herramientas nuevas |
| IX · Genericidad territorial | Nada se particulariza para un municipio. El umbral tampoco lleva ningún caso por skill: qué evals entran lo dicen sus `comandos` (FR-001) |

### Reglas de dependencia (`docs/ROADMAP.md` §2, constitución IV)

| Regla | Cómo la cumple |
|---|---|
| R1 · `internal/core/**` no importa adaptadores | Sin cambios en `internal/core` |
| R2 · Solo `internal/httpx` importa `net/http` | Ningún import nuevo |
| R3 · Solo `cache`/`store`/`graph` importan SQLite y `database/sql` | Ningún import nuevo |
| R4 · Solo `cli` y `cmd/` llaman a `os.Exit` | Ningún `os.Exit` nuevo |
| R5 · Solo `render` escribe en stdout | Ningún `fmt.Print*`, `os.Stdout` ni `os.Stderr` nuevo: el elemento y el motivo van al informe, como los demás |
| R6 · `internal/graph` no importa `source/*` ni `render` | Sin cambios |

## Project Structure

### Documentation (this feature)

```text
specs/022-h23-1-jurisprudencia-coteja-el/
├── plan.md                  # este fichero
├── research.md              # V1-V23, M1-M14, S1-S4, D1-D16
├── data-model.md            # la eval que pide el cotejo, el cotejo, lo medido de cada respuesta, el umbral
├── quickstart.md            # §0-§7, en make ci; §8, el job de cierre; §9, una persona
├── contracts/
│   ├── umbral-documento-sin-cotejar.md   # los quince, el elemento, el motivo, la fila, los casos, los mutantes, el uso
│   ├── skill-jurisprudencia.md           # qué cambia en SKILL.md, qué no, su forma y la salida que consume
│   └── documentos.md                     # qué pasaje cambia en cada documento
├── checklists/              # del spec
├── mutantes.md              # lo escribe la tarea de FR-064
└── tasks.md                 # la escribe /speckit-tasks
```

### Source Code (repository root)

```text
internal/evals/
├── informe.go               # pideElCotejo, cotejaUnDocumento, medirLosCotejos; campos de respuestaDelJuez y
│                            # respuestasMedidas; métodos de grupoJuzgado; comentarios de EscribirInforme
├── umbrales.go              # umbralDeDocumentoSinCotejar, su nombre y sus constantes; su sitio en umbrales();
│                            # sinCotejarDelUmbral en nombradasEn; comentarios
├── doc.go                   # lo que dice de los umbrales de cada skill (comentario)
├── informe_test.go          # TestUmbralesDeJurisprudencia: cuatro casos y su arnés; TestInformeFinalDeLosUmbrales
└── conjunto_test.go         # TestEvalsDelRepositorio/piden-el-cotejo
skills/jurisprudencia/SKILL.md     # pasos 1 y 3
CHANGELOG.md, CONTRIBUTING.md, docs/JURISPRUDENCIA.md, CLAUDE.md
```

**Structure Decision**: la de `docs/ROADMAP.md` §2 y `CLAUDE.md`, sin paquetes ni ficheros Go nuevos. La medida va
junto a la de `cita_sin_documento`, en `informe.go`, y el umbral junto a los demás, en `umbrales.go` (research D1).
`sondeo.go`, `juzgar.go`, `sentencias.go`, `juez.go`, `medida.go`, `definicion.go`, los guiones de `scripts/` y el
`Makefile` no cambian.

## Aceptación e2e

**Aceptación e2e: no aplica.** El hito no cambia el código del binario: `cmd/kitlegal` no enlaza `internal/evals`
(research V14) y cambiarlo fuera de ese paquete está fuera de alcance (FR-052). El `SKILL.md` de `jurisprudencia` viaja
empotrado, y `kitlegal skills install` lo instala igual que hoy: sus guiones ya comparan lo instalado con lo empotrado,
sea cual sea su texto (research V17). No hay comportamiento de `kitlegal` que un guion `testscript` describa (research
D14). La aceptación es la de la skill y la del instrumento: el job de evals de cierre (SC-001). En `make ci` la fijan,
sin modelo:

| Historia | Tests | Requisitos |
|---|---|---|
| US1 | `TestUmbralesDeJurisprudencia` (casos `ninguna-cuenta`, `una-sin-cotejar-en-el-modo-orden`, `una-sin-cotejar-en-el-modo-herramienta`, `una-sin-medir-y-sin-cotejar`), `TestEvalsDelRepositorio/piden-el-cotejo` | FR-001 a FR-005, FR-062, FR-064; SC-002 a SC-004, SC-006 |
| US2 | `make skills-check` (`TestSkillsDelRepositorio`, `TestOrdenesDeLasSkillsEmpotradas`); el resto lo leen la revisión final y el job | FR-020 a FR-024; SC-007 |
| US3 | `TestUmbralesDeJurisprudencia` (los quince, once que deciden; `una-sin-activar-y-sin-cotejar`), `TestUmbralesDelInforme`, `TestInformeEnDosModos`, `TestUmbralesConJuezYConSentencias`, `TestInformeConElJuez`, `TestInformeFinalDeLosUmbrales`, los del juicio y las series, `TestCopiasDelJuez`, `TestDefinicionDelJob`, `schema-check` y los del sondeo, sin cambiar lo que esperan | FR-006, FR-007, FR-030 a FR-034, FR-063; SC-005 |
| US4 | Ninguno: lo comprueba la revisión final por lectura (FR-045) | FR-040 a FR-045; SC-008 |

Ninguna tarea `[aceptacion]`.

## Controles mecánicos que este hito añade o toca

### Objetivos del `Makefile`

Ninguno nuevo ni cambiado. `test` y `skills-check` llevan los casos y el subtest nuevos; `schema-check` no cambia.

### Tests nuevos

Ninguna función de test nueva (research D8). Casos y subtests nuevos, en tests que ya existen:

| Test | Fichero | Qué añade | Cubre |
|---|---|---|---|
| `TestUmbralesDeJurisprudencia` | `internal/evals/informe_test.go` | Cuatro casos: `una-sin-cotejar-en-el-modo-orden`, `una-sin-cotejar-en-el-modo-herramienta`, `una-sin-activar-y-sin-cotejar` y `una-sin-medir-y-sin-cotejar` (contracts/umbral-documento-sin-cotejar.md §5) | FR-001 a FR-005, FR-030, FR-062, FR-064; SC-002, SC-004, SC-006; US3, escenario 3 |
| `TestEvalsDelRepositorio` | `internal/evals/conjunto_test.go` | El subtest `piden-el-cotejo`: con las evals del repositorio, piden el cotejo la 04, la 06, la 08 y la 09 de `jurisprudencia`, y ninguna de `boe-legislacion` ni de `legal-core` | FR-001, FR-007, FR-063; SC-005 |

### Tests existentes que cambian

Los que el prototipo deja en rojo (research M2, M4), y ninguno más:

- `TestUmbralesDeJurisprudencia`: exige quince umbrales y once que deciden donde exigía trece y nueve; sus diez casos de
  antes esperan además los dos `documento_sin_cotejar` a 0 de 12 —`ninguna-cuenta` lo fija con las 24 sesiones que
  cotejan por la traza y por la llamada—; su arnés gana lo de research D9 y D10 (`escribirSinElCotejo`,
  `umbralDocumentoSinCotejar`, los campos `sinCotejo` y `sinMedir` del caso y `sinCotejar` y `sinMedir` de
  `medidasDeJurisprudencia`, la constante de las 12 respuestas y la sesión sin medir en
  `mensajesDeLosVotosDeJurisprudencia`), y su comentario de cabecera dice los quince.
- `TestInformeFinalDeLosUmbrales`: su caso `como-en-el-cierre-de-h26` lee los trece de H26 con
  `umbralesDeJurisprudenciaComoEnH26` en lugar de `umbralesDeJurisprudencia(nil)`, que pasa a dar quince; lo que
  espera no cambia (research D11).

Ninguno se desactiva ni se salta. Los que no cambian lo que esperan y lo prueban: `TestUmbralesDelInforme`,
`TestUmbralesConJuezYConSentencias`, `TestUmbralConjunto`, `TestInforme`, `TestInformeConElJuez`,
`TestInformeMarkdownDeLosUmbrales`, `TestInformeEnDosModos`, `TestInformeConSesionesSinMedir`, los del juicio
(`TestJuzgar`, `TestJuzgarLasLlamadas`, `TestJuzgarSentencias`, `TestCitaSinDocumento`), `TestCopiasDelJuez`,
`TestMedidaVersionada`, `TestDefinicionDelJob` y los del sondeo (`TestJuicioDelSondeo`, `TestSalidaDelSondeo`,
`TestSondear`) (FR-030 a FR-033; quickstart §3).

### Puntos de entrada fuera de `make ci` (etiqueta `evals`)

`TestEjecucionDelJob`, `TestMedidaDelJuez` y `TestSondeo` no cambian. Ninguna tarea los ejecuta (FR-051).

### Fixtures, `testdata/` y `schemas/`

Nada cambia en `schemas/`, `testdata/` ni `evidencias/` (research D16). Las sesiones de los tests son sintéticas, en
`t.TempDir()`. Ninguna tarea lleva `[datos]`.

### CI

Ningún fichero de `.github/` cambia. El job de evals se ejecuta en la propuesta de cambio, tras la revisión final, y lo
lanza el workflow (FR-066).

## Controles de umbral

Las once primeras filas son los umbrales que deciden en el informe del job de cierre de `jurisprudencia` (FR-061): los
dos nuevos y los nueve que ya decidían. Los cuatro recuentos de las clases del juez no tienen fila: no deciden. Ni
`boe-legislacion` ni `legal-core` tienen fila: su rojo no es motivo de este hito (FR-050). Las demás son los umbrales
que se miden en `make ci`.

| Requisito | Umbral | Control | Dónde |
|---|---|---|---|
| FR-001, FR-002, FR-003, FR-004, SC-001, SC-009 (modo orden) | 0 sesiones sin cotejo de las 12 del modo orden en las evals que piden el cotejo | umbral del informe con `decide: true`: con 1, motivo con cada sesión, veredicto `fallo` y el trabajo en rojo, que el cierre cuenta | `evals:jurisprudencia:documento_sin_cotejar:claude-sonnet-5-5:orden` |
| FR-001, FR-002, FR-003, FR-004, SC-001, SC-009 (modo herramienta) | 0 de las 12 del modo herramienta | ídem | `evals:jurisprudencia:documento_sin_cotejar:claude-sonnet-5-5:herramienta` |
| FR-030, SC-001 (modo orden) | 0 respuestas sin la skill activada, de 30 | el de hoy, que sigue | `evals:jurisprudencia:sin_activar:claude-sonnet-5-5:orden` |
| FR-030, SC-001 (modo herramienta) | ídem | ídem | `evals:jurisprudencia:sin_activar:claude-sonnet-5-5:herramienta` |
| FR-030, SC-001 (modo orden) | 0 respuestas con una cita sin documento cotejado o un ECLI sin origen, de 30 | el de hoy, que sigue | `evals:jurisprudencia:cita_sin_documento:claude-sonnet-5-5:orden` |
| FR-030, SC-001 (modo herramienta) | ídem | ídem | `evals:jurisprudencia:cita_sin_documento:claude-sonnet-5-5:herramienta` |
| FR-030, SC-001 | como mucho 0.017 de las 60 respuestas juzgadas de los dos modos marcadas en `afirma_lo_no_leido`: 1 de 60 | el de hoy, que sigue | `evals:jurisprudencia:afirma_lo_no_leido:claude-sonnet-5-5` |
| FR-030, FR-032, SC-001 (defectos) | 0 defectos sin marcar, de 125 | el de hoy, con la medida versionada | `evals:jurisprudencia:medida_del_juez:afirma_lo_no_leido:defectos_sin_marcar` |
| FR-030, FR-032, SC-001 (correctos) | 0 correctos marcados, de 124 | ídem | `evals:jurisprudencia:medida_del_juez:afirma_lo_no_leido:correctos_marcados` |
| FR-030, SC-001 (juez, modo orden) | como mucho 900 s de votos del modo | el de hoy: por encima, motivo de la ejecución y `fallo` | `evals:jurisprudencia:duracion_del_juez:orden` |
| FR-030, SC-001 (juez, modo herramienta) | ídem | ídem | `evals:jurisprudencia:duracion_del_juez:herramienta` |
| FR-005, FR-062, SC-002, SC-003, SC-004 | 1 sin cotejo de 12 da `cumple: false`, `fallo` y su motivo en cada modo; 0 de 12 con las 24 sesiones cotejando por la traza y por la llamada; la sesión de Haiku, las de las evals que no lo piden y la sin medir, 0 en el `total` y en la `medida`, y la sin medir con su motivo | los casos de contracts §5, con los quince umbrales, los motivos y el veredicto escritos a mano | `ci:internal/evals/informe_test.go:TestUmbralesDeJurisprudencia` |
| FR-006, FR-063, SC-005 (`jurisprudencia`) | 15 elementos y 11 que deciden, en el orden de contracts §1 | la lista exacta en cada caso, y su longitud y los que deciden contados | `ci:internal/evals/informe_test.go:TestUmbralesDeJurisprudencia` |
| FR-007, FR-063, SC-005 (`boe-legislacion` y `legal-core`) | 13 y 9 en `boe-legislacion`; `[]` en `legal-core`; 0 evals que piden el cotejo en las dos | los trece escritos a mano, sin cambiar; `legal-core` con sus evals del repositorio; y las evals de las tres skills, leídas del repositorio | `ci:internal/evals/umbrales_test.go:TestUmbralesDelInforme`, `ci:internal/evals/informe_test.go:TestInformeEnDosModos`, `ci:internal/evals/conjunto_test.go:TestEvalsDelRepositorio` |
| FR-064, SC-006 | 2 de 2 mutantes hacen fallar al menos un caso | los casos de la fila de FR-062, que la tarea de los mutantes ve fallar en una copia fuera del árbol (quickstart §4) | `ci:internal/evals/informe_test.go:TestUmbralesDeJurisprudencia` |
| FR-023, SC-007 | `SKILL.md` por debajo de 300 líneas, su tabla la regenerada y 0 herramientas nuevas | `skills-check`, que no cambia | `ci:Makefile:skills-check` |
| FR-030, FR-031 | los trece de antes, el juicio de cada sesión y la tasa de cada serie, sin cambiar lo que esperan sus tests | los tests de hoy, sin cambiar | `ci:internal/evals/umbrales_test.go:TestUmbralesConJuezYConSentencias`, `ci:internal/evals/informe_test.go:TestInformeConElJuez`, `ci:internal/evals/juzgar_test.go:TestJuzgar`, `ci:internal/evals/informe_test.go:TestInforme` |
| FR-032 | 4 de 4 copias idénticas en la carpeta del juez; los peores casos del job y los esquemas, sin cambiar | los dos tests y `schema-check`, sin cambiar lo que esperan | `ci:internal/evals/medida_test.go:TestCopiasDelJuez`, `ci:internal/evals/definicion_test.go:TestDefinicionDelJob`, `ci:Makefile:schema-check` |
| FR-033 | las líneas y el código de salida del sondeo, sin cambiar | los tests del sondeo, sin cambiar | `ci:internal/evals/sondeo_test.go:TestSalidaDelSondeo` |
| FR-060 | `make ci` en verde | el objetivo entero | `ci:Makefile:ci` |

Sin fila: SC-008, que comprueba la revisión final por lectura (FR-045); SC-010, que mide una persona; y FR-020 a
FR-022, que no tienen umbral propio: los leen los jueces de la revisión final, y su efecto lo miden las dos primeras
filas (SC-001).

**El recorrido del umbral nuevo.** Con una sesión de la 06 del modo orden que no coteja: `medirLosCotejos` la deja con
`pideElCotejo` y sin `cotejo`; `umbralDeDocumentoSinCotejar` da 1 de 12, `cumple: false` y `decide: true`;
`motivosDeLosUmbrales` da su motivo con la sesión; el veredicto es `fallo`; el trabajo `evals (jurisprudencia)` sale en
rojo; `medir_cierre` lo cuenta y guarda el informe; y el informe final busca en `umbrales` el nombre exacto de la fila y
lo da como no cumplido (research V12). Con todas cotejando, todo lo anterior en verde, y «comprobado por su control». En
`make ci`, el mismo recorrido lo hace `TestUmbralesDeJurisprudencia` con sesiones sintéticas, y los dos mutantes
prueban que falla si el umbral deja de contar cada sesión o de decidir (research M8).

## Uso, de fuera adentro (criterio de uso, ADR 0028)

El detalle, con los ejemplos y sus bytes, está en contracts/umbral-documento-sin-cotejar.md §2 a §4 y §7 y en
contracts/skill-jurisprudencia.md §4. Nada crece con lo consultado en meses de uso —cientos de normas y miles de bloques—:
lo que consume la skill depende de los documentos de la pregunta, y cada job parte de las evals del repositorio y mide
de nuevo.

- **La salida de `cita cotejar`**, que no cambia (la skill, una vez por documento delante y antes de decir nada de
  él; con *n* documentos, *n* invocaciones): el sobre del caso de la 06 mide 937 bytes, y su `data`, 673 (research V20).
  Lo que se pasa es la ficha, no el texto. La señal —`es_la_pedida: false` y su hallazgo— vale para ese documento en
  esa pregunta y no se acumula.
- **Los dos elementos de `umbrales`** (el job, una vez; el informe final, una vez por medición del cierre; la persona):
  366 y 378 bytes con 0 de 12; los quince, unos 4,8 KB, frente a 4 100 bytes de los trece (research M9, M13). Ninguna
  skill los lee: 0 invocaciones por pregunta. El `cumple: false` de un modo deja de darse en el primer job en que todas
  sus sesiones cotejan.
- **El motivo** (la persona y la reparación del cierre; como mucho uno por modo): 151 bytes con una sesión, y como mucho
  723 en el modo orden y 873 en el herramienta, con las 12 (research M10, M12). Deja de darse en el primer job en que el
  umbral se cumple.
- **La fila de `informe.md`** (la persona): una más por modo, 92 a 99 bytes (research M11).
- **`SKILL.md`** (el agente que activa la skill, una vez por conversación): unas diez líneas más, por debajo de 300.
  No da señales.
- **`mutantes.md`** (la revisión final, una vez) y **los documentos** (quien contribuye o lee un informe): no crecen
  ni dan señales.

## Decisiones

- **La medida va con las respuestas de cada modo**, las del `total` de `sin_activar`, en `medirLasRespuestas`, como la
  de `cita_sin_documento`; el juicio de cada sesión no cambia (D1).
- **Cotejar es satisfacer el comando esperado `cita cotejar` sin ROJ**, sobre las invocaciones con las que se juzga la
  sesión: la traza y las llamadas (D2); el falso de `invocacionesDe`, al que una sesión leída no llega, no se trata
  aparte (D3).
- **El umbral vive dentro de los umbrales de las respuestas**: la condición con la que existen no cambia (D4), y qué
  evals entran lo dicen sus `comandos` (D5).
- **El nombre, la descripción y el sitio**, detrás de `cita_sin_documento` de su modo (D6); **el motivo nombra cada
  sesión**, sin el modo detrás (D7).
- **Los tests van en las tablas que ya hay**, con la 06 sin cotejo escrita de nuevo y una sesión sin medir (D8 a D10);
  **el caso del cierre de H26 sigue esperando trece** (D11).
- **Los mutantes, en una copia fuera del árbol, con una línea cada uno** (D12).
- **La skill cambia en los pasos 1 y 3, y nada más** (D13); **sin aceptación e2e del binario** (D14).
- **Los documentos dicen lo nuevo y nada más**, la skill en «Corregido» y el umbral en «Añadido»; `docs/WORKFLOW.md`
  no cambia (D15).
- **Ningún dato externo** (D16).

## Trazabilidad: cada mecanismo y su requisito

| Mecanismo | Requisito |
|---|---|
| `pideElCotejo(eval)` | FR-001: qué evals entran, por sus `comandos` |
| `comandoDelCotejo` y `cotejaUnDocumento(sesion)` | FR-002 |
| `respuestaDelJuez.pideElCotejo` y `.cotejo`, `respuestasMedidas.conCotejo` y `medirLosCotejos` | FR-001, FR-002, FR-005: las respuestas del `total` de `sin_activar`, con lo que se mide de ellas |
| `grupoJuzgado.queLePidenElCotejo()` y `.sinCotejar()` | FR-001: el `total` y la `medida`, cada sesión una vez |
| `umbralDeDocumentoSinCotejar`, `nombreDelUmbralDeDocumentoSinCotejar` y sus tres constantes (prefijo, descripción y umbral 0) | FR-001, FR-004 |
| Su llamada en `respuestasMedidas.umbrales`, detrás de `cita_sin_documento` | FR-001, FR-006 (el orden) |
| `sinCotejarDelUmbral` y su llamada en `nombradasEn` | FR-003 |
| Los comentarios de `doc.go`, `informe.go` y `umbrales.go` que dicen cuántos umbrales hay y cuáles | FR-006: el hito los deja falsos (research V21) |
| Los cuatro casos de `TestUmbralesDeJurisprudencia` y su arnés | FR-062, FR-064; SC-002 a SC-004, SC-006; US3, escenario 3 |
| Quince y once en `TestUmbralesDeJurisprudencia` | FR-006, FR-063; SC-005 |
| `umbralesDeJurisprudenciaComoEnH26` | FR-030: `TestInformeFinalDeLosUmbrales` no cambia lo que espera (research D11) |
| `TestEvalsDelRepositorio/piden-el-cotejo` | FR-001, FR-007, FR-063; SC-005 |
| Los mutantes en una copia y `mutantes.md` | FR-064; SC-006 |
| Los pasos 1 y 3 de `SKILL.md` | FR-020 a FR-024; SC-007 |
| Los cuatro documentos | FR-040 a FR-045, FR-065; SC-008 |

Nada del diseño atiende a un estado que no pasa el umbral de materialidad: una skill con una eval que pide el cotejo y
sin juez ni `sentencias` no la entrega ningún emisor (research D4); el falso de `invocacionesDe` no lo alcanza una
sesión leída (research D3); y los casos límite del criterio de cotejo —código 2, `--dry-run`, cotejo sin `--roj`, varios
cotejos— son los del juicio, que ya prueban sus tests (research D2, D8).

## Datos externos

Ninguno (research D16): ni fuente ni grabación. Ningún manifiesto `grabaciones.json` ni test `TestGrabar*` cambia, y el
paso `grabar_datos` no tiene nada que grabar. Las cifras de partida (6 de 48) son de `docs/USO.md`, que está en el
repositorio.

## Orden de implementación (de dentro afuera)

1. **El umbral** (`internal/evals/informe.go`, `umbrales.go` y `doc.go`, con `informe_test.go` y `conjunto_test.go`):
   la medida, el elemento, su motivo, los comentarios que deja incompletos, los quince en
   `TestUmbralesDeJurisprudencia` con sus cuatro casos nuevos, `umbralesDeJurisprudenciaComoEnH26` y `piden-el-cotejo`.
   Es una sola tarea porque el código cambia la forma de `umbrales` y los tests que la esperan tienen que cambiar con
   él para dejar `make ci` en verde (research M2).
2. **Los mutantes**: el escenario 4 del quickstart sobre la cabeza, y su registro en
   `specs/022-h23-1-jurisprudencia-coteja-el/mutantes.md`.
3. **La skill**: `skills/jurisprudencia/SKILL.md`, pasos 1 y 3, con contracts/skill-jurisprudencia.md y el
   escenario 5 del quickstart.
4. **Los documentos**: con contracts/documentos.md y el escenario 6 del quickstart.

**Obligaciones para `tasks.md`**:

- ninguna tarea `[aceptacion]` ni `[datos]`;
- cada tarea deja `make ci` en verde: por eso el paso 1 lleva el código con todos los tests que cambian de forma;
- cada fila `ci:` de «Controles de umbral» tiene la tarea que construye o mantiene su control, con un caso que lo ve
  fallar; la de FR-062, además, con los mutantes del paso 2;
- la tarea del paso 1 declara `internal/evals/informe.go`, `internal/evals/umbrales.go`, `internal/evals/doc.go`,
  `internal/evals/informe_test.go` e `internal/evals/conjunto_test.go`, y ningún otro fichero de `internal/`;
- la tarea de los mutantes declara `specs/022-h23-1-jurisprudencia-coteja-el/mutantes.md`, no cambia nada bajo
  `internal/` y termina con el worktree retirado;
- la tarea de la skill declara `skills/jurisprudencia/SKILL.md`, y ningún otro fichero de `skills/`;
- la tarea de los documentos declara `CHANGELOG.md`, `CONTRIBUTING.md`, `docs/JURISPRUDENCIA.md` y `CLAUDE.md`, y
  ningún otro; comprueba `docs/WORKFLOW.md` sin cambiarlo;
- ninguna tarea ni corrección cumple un umbral cambiando su valor, dejándolo en `decide: false`, sacando sesiones de
  su `total`, ni cambiando las evals, la rúbrica, los casos o la medida (FR-050);
- ninguna tarea toca `boe-legislacion`, `legal-core`, una eval, `schemas/`, `testdata/`, `evidencias/`, `.github/`,
  `scripts/`, `.specify/`, `internal/evals/sondeo.go`, `docs/ROADMAP.md`, `docs/WORKFLOW.md`, un ADR, `docs/USO.md`,
  `docs/HISTORIAL.md`, `README.md`, la web, `specs/001-…` a `specs/021-…` ni el binario fuera de `internal/evals`
  (FR-032 a FR-034, FR-045, FR-052); ninguna escribe en la carpeta de evidencias ni nombra en su línea una ruta de ella
  o de `schemas/`;
- ninguna usa la red, abre una sesión con modelo ni ejecuta `make evals`, `make evals-sondeo`, `make evals-medir-juez`,
  `scripts/evals*.sh`, `TestEjecucionDelJob`, `TestMedidaDelJuez` o `TestSondeo` (FR-051);
- ninguna publica ni mide en la plataforma, y los escenarios 8 y 9 del quickstart no son tareas.

## Complexity Tracking

| Desviación | Por qué hace falta | Alternativa más simple rechazada |
|---|---|---|
| «Aceptación e2e: no aplica» (principio III: «cada hito empieza por el test e2e») | El hito no cambia el código del binario. La aceptación es la del job (SC-001), y cada pieza entra con sus tests en `make ci`. Precedentes: H24, H25 y H26 | Un guion que ejerciera el binario sin cambios: no describiría la entrega |

## Comprobación contra la rúbrica del juez (`juez_plan`) y `precheck.sh plan`

- `precheck.sh plan`: existen `plan.md` y `research.md`; ninguna marca de aclaración pendiente; están
  `## Constitution Check`, la línea «Aceptación e2e» y `## Controles de umbral`, con cada fila nombrando requisitos del
  spec y un control con forma. Ejecutado en esta sesión: sin defectos.
- a · Constitution Check: un ítem por principio (I-IX) y por regla de dependencia (R1-R6).
- b · Dependencias: ninguna nueva (research V2).
- c · Reglas de dependencia: tabla R1-R6; ningún import nuevo.
- d · Errores y códigos: ningún error ni código de `kitlegal` cambia; el job sale en rojo con `fallo`, como hoy.
- e · Tests primero: «Aceptación e2e: no aplica» con el motivo; los casos y el subtest nuevos, los dos tests que
  cambian, los que no cambian, los puntos de entrada y los fixtures.
- f · Alcance: nada fuera del spec. Los comentarios de `internal/evals` cambian porque el hito los deja falsos.
- g · Sin atajos: ningún `//nolint`, `t.Skip`, TODO ni error silenciado previstos (el falso de `invocacionesDe`,
  research D3).
- h · Mejor alternativa: cada decisión con la rechazada (research D1-D16).
- i · Afirmaciones verificadas: V1-V23 con fichero y línea u orden; M1-M14 medidas en esta sesión; S1-S4 como
  supuestos.
- j · Quickstart ejecutable: órdenes, rutas y nombres reales; dice qué vale ya hoy y qué deja cada escenario; las
  órdenes del 0, las de `make` del 2, el 3, el 5 y el 6 se han ejecutado sobre el árbol de hoy, la mecánica del
  worktree del 4 también, y sus sustituciones y la orden del 1, sobre el prototipo (research M3, M8).
- k · Datos externos: ninguno.
- l · Autonomía: ninguna tarea para una persona; el sondeo no lo lanza el run; el cierre en la plataforma lo hace el
  workflow.
- m · Uso: «Uso, de fuera adentro», contracts/umbral-documento-sin-cotejar.md §7 y contracts/skill-jurisprudencia.md §4,
  con bytes medidos, veces por pregunta y por job, y lo que apaga cada señal.
- n · Proporcionalidad: «Trazabilidad»; sin tipos, funciones exportadas, paquetes, ficheros Go ni funciones de test
  nuevos; nada para estados por debajo del umbral de materialidad.
- o · Controles de umbral: los once que deciden en el cierre de `jurisprudencia` y una fila por cada umbral medido en
  `make ci`.
