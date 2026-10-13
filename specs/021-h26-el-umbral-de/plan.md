# Implementation Plan: H26 · El umbral de las clases del juez pasa a ser una tasa sobre los dos modos

**Branch**: `021-h26-el-umbral-de` | **Date**: 2026-10-13 | **Spec**: [spec.md](./spec.md)

**Input**: Feature specification from `/specs/021-h26-el-umbral-de/spec.md`

**Modo**: desatendido. Las decisiones técnicas se tomaron con el «Criterio de decisión autónoma» de la constitución y
están en [research.md](./research.md) (D1-D20), cada una con su alternativa rechazada. Lo que se afirma del
repositorio remite a su tabla V; lo medido en esta sesión, sobre un prototipo fuera del repositorio que ya no existe, a
su tabla M; y lo que no se pudo comprobar son los supuestos S1-S5.

## Summary

El job de evals decide hoy `afirma_lo_no_leido` con 0 respuestas marcadas en cada modo. El ADR 0038 decidió que una
clase del juez que decide lleve una tasa sobre los dos modos juntos. H26 hace que el job la aplique, en cuatro piezas:

1. **El umbral conjunto** (`internal/evals/umbrales.go`): por cada clase con `decide: true`, un elemento
   `<clase>:<modelo>` que suma las marcadas y las juzgadas de todos los modos, compara con el `umbral` de la clase y
   decide. Va detrás de los umbrales de las respuestas de los modos y delante de los de la medida. Su motivo nombra
   cada respuesta marcada con su modo.
2. **El recuento por modo**: `<clase>:<modelo>:<modo>` sigue igual y pasa a `decide: false`.
3. **Los valores**: `umbral: 0.028` en `evals/boe-legislacion/juez/clases.yaml` y `umbral: 0.017` en
   `evals/jurisprudencia/juez/clases.yaml`, con el diff de [contracts/clases-yaml.diff](./contracts/clases-yaml.diff).
4. **Los documentos**: lo que `CHANGELOG.md` (*Unreleased*), `CONTRIBUTING.md`, `docs/JURISPRUDENCIA.md`,
   `docs/WORKFLOW.md` y `CLAUDE.md` dicen del umbral ([contracts/documentos.md](./contracts/documentos.md)).

El cambio de código son unas cincuenta líneas en un fichero, sin tipo, campo, función exportada, paquete ni dependencia
nuevos. Un prototipo lo confirma: con él, `umbrales` lleva trece elementos y nueve que deciden en las dos skills; 3
marcadas de 108 dan `aprobado` en los tres repartos y 4 dan `fallo`; y 1 de 60 da `aprobado` y 2, `fallo` (research M4
a M6). El sondeo, la medida versionada, las copias del juez y la definición del job no cambian (research M2).

No hay ADR nuevo. No se tocan ninguna skill ni ninguna eval, `schemas/`, `evidencias/`, `testdata/`, `.github/`,
`scripts/`, la constitución, `docs/ROADMAP.md`, los ADR ni el binario fuera de `internal/evals` (FR-036, FR-055,
FR-062).

## Technical Context

**Language/Version**: Go 1.27 (`go 1.27.0`, `toolchain go1.27.2`), `CGO_ENABLED=0`, `-trimpath` (research V1); YAML de
las dos declaraciones de clases; Markdown de los documentos.

**Primary Dependencies**: las fijadas, sin ninguna nueva. El código que cambia usa `fmt` y `strings`, que
`umbrales.go` ya importa (research V2); los tests, `stretchr/testify`.

**Storage**: ninguno. Dos ficheros de datos del repositorio cambian un número cada uno.

**Testing**: `go test -race` en `make ci`: tablas con subtests con nombre y `t.Parallel`, ejecuciones sintéticas en
`t.TempDir()` escritas con `EscribirInforme` y un votante de votos grabados. Ningún test nuevo con la etiqueta
`evals`. El job de evals, sobre la cabeza, lo lanza el workflow.

**Target Platform**: el job, `ubuntu-24.04`; `make ci`, macOS y Linux.

**Project Type**: instrumento de medida de las skills (`internal/evals`, el flujo `evals`). El binario `kitlegal` no
lo enlaza (research V14).

**Performance Goals**: ninguno nuevo. El hito no añade ni quita ninguna sesión ni ningún voto (FR-035): las cotas de
900 s y los peores casos del job son los de hoy.

**Constraints**: ningún test de `make ci` usa la red ni abre una sesión con modelo; ningún `//nolint` ni `t.Skip`;
`evidencias/`, `schemas/` y `testdata/` no cambian; la rúbrica, los casos, la medida, el modelo del juez y su versión no
cambian; el run no elige ningún valor de umbral (FR-020, FR-061).

**Scale/Scope**: 108 respuestas juzgadas por job en `boe-legislacion` y 60 en `jurisprudencia`; 13 elementos de
`umbrales` en cada una; 1 función de test nueva y 8 que cambian lo que esperan.

## Constitution Check

*GATE: pasado antes de la fase 0 y re-evaluado tras el diseño (fase 1): sin violaciones; la única desviación
justificada está en Complexity Tracking.*

### Principios

| Principio | Cómo lo cumple H26 |
|---|---|
| I · Fuentes públicas y frontera humana | No consulta ninguna fuente ni toca `internal/httpx` o `docs/SOURCES.md`. El valor del umbral lo fijó una persona en el ADR 0038 y solo una persona lo estrecha: el run no lo elige (FR-020) |
| II · Nada sin cita ni fuente | Es lo que mide el umbral que el hito pone a decidir: que la respuesta no afirme lo que no leyó. Ningún sobre ni ninguna cita cambian. La regla «nunca inventar contenido legal» sigue en las skills, que no se tocan |
| III · Tests primero y offline | «Aceptación e2e: no aplica» (abajo, con el motivo). Cada pieza entra con sus tests en `make ci`, offline y sin modelo. La aceptación es el job de cierre (SC-001) |
| IV · Hexagonal y errores tipados | Ningún paquete del binario, ningún error tipado y ningún código de salida de `kitlegal` cambian. El job sigue saliendo en rojo con el veredicto `fallo`, y el sondeo, con 0 |
| V · Simplicidad y dependencias | Ninguna dependencia nueva. Un método, una función y dos constantes nuevas (`umbralesConjuntos`, `nombreDelUmbralConjunto`, `descripcionDelUmbralConjunto` y `sesionConSuModo`) en un fichero que ya existe, y una rama que deja de usarse, retirada. Cada mecanismo se traza a un FR («Trazabilidad») |
| VI · Un binario, convenciones de agente | El binario, sus banderas, `--describe` y las tablas de comandos de las skills no cambian |
| VII · Grafo y privacidad | Nada del grafo cambia. Los tests usan directorios temporales |
| VIII · Skills primero | No entrega ni cambia ninguna skill: protege las tres, haciendo que el verde y el rojo de sus evals vuelvan a decir algo. `SKILL.md`, `references/` y evals, sin tocar (FR-036) |
| IX · Genericidad territorial | Nada se particulariza para un municipio. El umbral conjunto tampoco lleva ningún caso por skill (FR-001) |

### Reglas de dependencia (`docs/ROADMAP.md` §2, constitución IV)

| Regla | Cómo la cumple |
|---|---|
| R1 · `internal/core/**` no importa adaptadores | Sin cambios en `internal/core` |
| R2 · Solo `internal/httpx` importa `net/http` | Ningún import nuevo |
| R3 · Solo `cache`/`store`/`graph` importan SQLite y `database/sql` | Ningún import nuevo |
| R4 · Solo `cli` y `cmd/` llaman a `os.Exit` | Ningún `os.Exit` nuevo |
| R5 · Solo `render` escribe en stdout | Ningún `fmt.Print*`, `os.Stdout` ni `os.Stderr` nuevo: el motivo y los elementos van al informe, como hoy |
| R6 · `internal/graph` no importa `source/*` ni `render` | Sin cambios |

## Project Structure

### Documentation (this feature)

```text
specs/021-h26-el-umbral-de/
├── plan.md                  # este fichero
├── research.md              # V1-V23, M1-M16, S1-S5, D1-D20
├── data-model.md            # la clase, los dos elementos, sus invariantes, lo que cambia
├── quickstart.md            # §0-§8, en make ci; §9, el job de cierre; §10, una persona
├── contracts/
│   ├── umbrales-del-informe.md      # los trece elementos, el conjunto, el recuento, el motivo, el uso, los mutantes
│   ├── clases-yaml.diff             # los dos valores, aplicable con git apply
│   └── documentos.md                # qué pasaje cambia en cada documento
├── checklists/              # del spec
├── mutantes.md              # lo escribe la tarea de FR-078
└── tasks.md                 # la escribe /speckit-tasks
```

### Source Code (repository root)

```text
evals/boe-legislacion/juez/clases.yaml       # umbral 0.028
evals/jurisprudencia/juez/clases.yaml        # umbral 0.017
internal/evals/
├── umbrales.go              # el umbral conjunto, el recuento por modo sin decidir, el motivo con el modo
├── doc.go                   # lo que dice de los umbrales de cada skill (comentario)
├── informe.go               # cómo se componen los umbrales (comentario)
├── umbrales_test.go         # TestUmbralConjunto, nuevo; el arnés y los umbrales esperados
├── informe_test.go          # los tests del informe con el juez y los de jurisprudencia
└── conjunto_test.go         # TestEvalsDelRepositorio: los dos valores
CHANGELOG.md, CONTRIBUTING.md, docs/JURISPRUDENCIA.md, docs/WORKFLOW.md, CLAUDE.md
```

**Structure Decision**: la de `docs/ROADMAP.md` §2 y `CLAUDE.md`, sin paquetes ni ficheros Go nuevos. Todo el código
cabe en `umbrales.go`, que ya tiene los umbrales y sus motivos. `sondeo.go`, `juez.go`, `conjunto.go`, `medida.go`,
`definicion.go`, los guiones de `scripts/` y el `Makefile` no cambian (research V8, M2).

## Aceptación e2e

**Aceptación e2e: no aplica.** El hito no cambia nada que vea quien usa `kitlegal`: el binario no enlaza
`internal/evals` (research V14) y cambiarlo fuera de ese paquete está fuera de alcance (FR-062), así que no hay
comportamiento del binario que un guion `testscript` describa (research D14). Su aceptación es la del instrumento: el
job de evals de cierre (SC-001). En `make ci` la fijan, sin modelo:

| Historia | Tests | Requisitos |
|---|---|---|
| US1 | `TestUmbralConjunto`, `TestUmbralesDeJurisprudencia` | FR-001 a FR-005, FR-072, FR-073; SC-002, SC-003 |
| US2 | `TestUmbralConjunto`, `TestUmbralesDeJurisprudencia`, `TestUmbralesDelInforme`, `TestInformeMarkdownDeLosUmbrales` | FR-010, FR-012, FR-074, FR-075; SC-004, SC-009 |
| US3 | `TestEvalsDelRepositorio`, `schema-check` | FR-020 a FR-022, FR-077; SC-007 |
| US4 | `TestInformeConElJuez`, `TestUmbralesDeJurisprudencia`, `TestUmbralesDelInforme`, `TestInformeEnDosModos`, `TestCopiasDelJuez`, `TestMedidaVersionada`, `TestEjecucionSinMedir`, `TestDefinicionDelJob` | FR-011, FR-030 a FR-036, FR-075, FR-076; SC-005, SC-006 |
| US5 | `TestJuicioDelSondeo`, `TestSalidaDelSondeo`, `TestSondear`, sin cambiar | FR-040; SC-014 |
| US6 | Ninguno: lo comprueba la revisión final por lectura (FR-055) | FR-050 a FR-055; SC-013 |

Ninguna tarea `[aceptacion]`.

## Controles mecánicos que este hito añade o toca

### Objetivos del `Makefile`

Ninguno nuevo ni cambiado. `test` lleva un test más, y `skills-check`, `TestEvalsDelRepositorio` con los dos valores.
`schema-check` no cambia.

### Tests nuevos

| Test | Fichero | Cubre |
|---|---|---|
| `TestUmbralConjunto` | `internal/evals/umbrales_test.go` | FR-001 a FR-005, FR-010, FR-072, FR-074; SC-002, SC-004 |

Sus siete casos, sobre la ejecución sintética «como el cierre» en dos modos —54 respuestas del modelo que decide por
modo— con la `clases.yaml` de `boe-legislacion` leída del repositorio (research D6, D7): `3-y-0`, `0-y-3`, `2-y-1`,
`4-y-0`, `2-y-2`, `104-con-2` y `104-con-3`. Cada uno exige los trece elementos de `umbrales` con sus valores escritos
a mano, los motivos de la raíz y el veredicto; y `0-y-3`, además, que el conjunto y el recuento del modo herramienta se
escriban como los dos ejemplos de `boe-legislacion` de contracts/umbrales-del-informe.md §2 y §3, tal cual. `3-y-0`
compara lo mismo, con el recuento del modo orden: el ejemplo de §3 con `orden` donde lleva `herramienta`.

### Tests existentes que cambian

Los ocho que el prototipo deja en rojo (research M2, M3), y ninguno más:

- `TestUmbralesDelInforme` y `TestUmbralesConJuezYConSentencias` (`umbrales_test.go`): los umbrales esperados, con el
  recuento por modo sin decidir y el conjunto en su sitio; los doce del contrato, escritos a mano, pasan a ser los
  trece, con nueve que deciden.
- `TestInformeConElJuez` (`informe_test.go`): los umbrales y los motivos de sus casos, con el nombre del conjunto y el
  modo de cada respuesta. `un-voto-que-no-llega` fija FR-075: el conjunto se cumple y el veredicto es `fallo` por la
  respuesta sin juzgar. `cuenta-su-proceso-con-tres` fija FR-076.
- `TestUmbralesDeJurisprudencia` (`informe_test.go`): trece elementos y nueve que deciden; `una-marcada-en-el-modo-orden`
  y `una-marcada-en-las-dos-clases` pasan a `aprobado`; dos casos nuevos, `dos-marcadas-en-el-modo-orden` y
  `una-marcada-en-cada-modo`, dan `fallo` (FR-073); y `el-elemento-del-contrato` compara, donde hoy compara el elemento
  de H25, los dos ejemplos de `jurisprudencia` de contracts/umbrales-del-informe.md §2 y §3, tal cual.
- `TestInformeMarkdownDeLosUmbrales` (`informe_test.go`): la fila del conjunto y las dos del recuento en la tabla
  «Umbrales».
- `TestInformeEnDosModos` y `TestInformeConSesionesSinMedir` (`informe_test.go`): solo la lista esperada de umbrales,
  que sale de los mismos ayudantes.
- `TestEvalsDelRepositorio` (`conjunto_test.go`): las clases leídas del repositorio, con `Umbral: 0.028` y
  `Umbral: 0.017` (FR-022, FR-077).

Ninguno se desactiva ni se salta. Ninguno se retira sin el que lo sustituye. Los que no cambian lo que esperan y lo
prueban: los del sondeo (FR-040), `TestCopiasDelJuez`, `TestMedidaVersionada`, `TestEjecucionSinMedir` y
`TestDefinicionDelJob` (FR-021, FR-033, FR-035).

### Puntos de entrada fuera de `make ci` (etiqueta `evals`)

`TestEjecucionDelJob`, `TestMedidaDelJuez` y `TestSondeo` no cambian. Ninguna tarea los ejecuta.

### Fixtures, `testdata/` y `schemas/`

Nada cambia en `schemas/` ni en `testdata/` (research V9, V13). Las tareas que cambian las dos `clases.yaml` llevan
`[datos]` porque lo pide FR-020; no dejan ningún manifiesto de grabación (research D15). Ninguna tarea escribe en
`evidencias/`.

### CI

Ningún fichero de `.github/` cambia. El job de evals se ejecuta en la propuesta de cambio, tras la revisión final, y
lo lanza el workflow (FR-079).

## Controles de umbral

Las dieciocho primeras filas son los umbrales que deciden en los informes del cierre, nueve por skill (FR-071): el
conjunto de `afirma_lo_no_leido` y los que ya decidían. El recuento por modo de una clase no tiene fila: no decide.
Las demás son los umbrales que se miden en `make ci`.

| Requisito | Umbral | Control | Dónde |
|---|---|---|---|
| FR-001, FR-003, FR-004, SC-001, SC-011 (`boe-legislacion`) | como mucho 0.028 de las respuestas juzgadas de los dos modos, marcadas en `afirma_lo_no_leido`: 3 de 108 | umbral del informe con `decide: true`: por encima, motivo con cada respuesta y su modo, veredicto `fallo` y el trabajo en rojo, que el cierre cuenta | `evals:boe-legislacion:afirma_lo_no_leido:claude-sonnet-5-5` |
| FR-031, SC-001 (`boe-legislacion`, modo orden) | 0 respuestas sin la skill activada | el de hoy, que sigue | `evals:boe-legislacion:sin_activar:claude-sonnet-5-5:orden` |
| FR-031, SC-001 (`boe-legislacion`, modo herramienta) | ídem | ídem | `evals:boe-legislacion:sin_activar:claude-sonnet-5-5:herramienta` |
| FR-033, SC-001 (`boe-legislacion`, defectos) | 0 defectos sin marcar, de 212 | el de hoy, con el recuento de la medida versionada | `evals:boe-legislacion:medida_del_juez:afirma_lo_no_leido:defectos_sin_marcar` |
| FR-033, SC-001 (`boe-legislacion`, correctos) | 0 correctos marcados, de 47 | ídem | `evals:boe-legislacion:medida_del_juez:afirma_lo_no_leido:correctos_marcados` |
| FR-034, SC-001 (`boe-legislacion`, sesiones, modo orden) | como mucho 900 s de la tanda del modo | el de hoy: por encima, motivo de la ejecución y `fallo` | `evals:boe-legislacion:duracion_de_las_sesiones:orden` |
| FR-034, SC-001 (`boe-legislacion`, sesiones, modo herramienta) | ídem | ídem | `evals:boe-legislacion:duracion_de_las_sesiones:herramienta` |
| FR-034, SC-001 (`boe-legislacion`, juez, modo orden) | como mucho 900 s de votos del modo | el de hoy | `evals:boe-legislacion:duracion_del_juez:orden` |
| FR-034, SC-001 (`boe-legislacion`, juez, modo herramienta) | ídem | ídem | `evals:boe-legislacion:duracion_del_juez:herramienta` |
| FR-001, FR-003, FR-004, SC-001, SC-011 (`jurisprudencia`) | como mucho 0.017 de las respuestas juzgadas de los dos modos, marcadas en `afirma_lo_no_leido`: 1 de 60 | umbral del informe con `decide: true`, como el de `boe-legislacion` | `evals:jurisprudencia:afirma_lo_no_leido:claude-sonnet-5-5` |
| FR-031, SC-001 (`jurisprudencia`, modo orden) | 0 respuestas sin la skill activada | el de hoy, que sigue | `evals:jurisprudencia:sin_activar:claude-sonnet-5-5:orden` |
| FR-031, SC-001 (`jurisprudencia`, modo herramienta) | ídem | ídem | `evals:jurisprudencia:sin_activar:claude-sonnet-5-5:herramienta` |
| FR-032, SC-001 (`jurisprudencia`, modo orden) | 0 respuestas con una cita sin documento cotejado o un ECLI sin origen | el de hoy, que sigue | `evals:jurisprudencia:cita_sin_documento:claude-sonnet-5-5:orden` |
| FR-032, SC-001 (`jurisprudencia`, modo herramienta) | ídem | ídem | `evals:jurisprudencia:cita_sin_documento:claude-sonnet-5-5:herramienta` |
| FR-033, SC-001 (`jurisprudencia`, defectos) | 0 defectos sin marcar, de 125 | el de hoy | `evals:jurisprudencia:medida_del_juez:afirma_lo_no_leido:defectos_sin_marcar` |
| FR-033, SC-001 (`jurisprudencia`, correctos) | 0 correctos marcados, de 124 | ídem | `evals:jurisprudencia:medida_del_juez:afirma_lo_no_leido:correctos_marcados` |
| FR-034, SC-001 (`jurisprudencia`, juez, modo orden) | como mucho 900 s de votos del modo | el de hoy | `evals:jurisprudencia:duracion_del_juez:orden` |
| FR-034, SC-001 (`jurisprudencia`, juez, modo herramienta) | ídem | ídem | `evals:jurisprudencia:duracion_del_juez:herramienta` |
| FR-072, SC-002 | 7 de 7 casos de `boe-legislacion`: 3 de 108 se cumple en 3 repartos, 4 da `fallo` en 2, y con 104, 2 se cumple y 3 da `fallo` | la ejecución sintética con la declaración del repositorio y votos grabados; umbrales, motivo y veredicto escritos a mano | `ci:internal/evals/umbrales_test.go:TestUmbralConjunto` |
| FR-073, SC-003 | 3 de 3 casos de `jurisprudencia`: 1 de 60 se cumple y 2 da `fallo` en 2 repartos | las sesiones sintéticas de sus diez evals, con su carpeta del juez | `ci:internal/evals/informe_test.go:TestUmbralesDeJurisprudencia` |
| FR-074, SC-004 | 3 marcadas en un solo modo: su recuento con `cumple: false` y `decide: false`, veredicto `aprobado` y 0 motivos | los casos `3-y-0` y `0-y-3` | `ci:internal/evals/umbrales_test.go:TestUmbralConjunto` |
| FR-030, FR-075, SC-005 | 1 respuesta sin juzgar con el conjunto cumplido: `fallo`, con 1 motivo de la ejecución que la nombra | el caso `un-voto-que-no-llega`, con un voto que no llega | `ci:internal/evals/informe_test.go:TestInformeConElJuez` |
| FR-011, FR-076, SC-006 | la clase que solo se publica: 1 elemento por modo con `decide: false` y 0 sin modo; con 3 respuestas con sí, el mismo veredicto | la lista exacta de umbrales de los casos `cuenta-su-proceso-con-tres` y `tres-con-si-en-afirma-que-existe` | `ci:internal/evals/informe_test.go:TestInformeConElJuez`, `ci:internal/evals/informe_test.go:TestUmbralesDeJurisprudencia` |
| FR-012, FR-075, SC-009 | 13 elementos y 9 que deciden, en su orden, en cada una de las dos skills; 0 en `legal-core` | los trece de `boe-legislacion` escritos a mano; los de `jurisprudencia`, contados; `legal-core`, con sus evals del repositorio | `ci:internal/evals/umbrales_test.go:TestUmbralesDelInforme`, `ci:internal/evals/informe_test.go:TestUmbralesDeJurisprudencia`, `ci:internal/evals/informe_test.go:TestInformeEnDosModos` |
| FR-022, FR-077, SC-007 | 2 de 2 valores iguales a los del ADR 0038: 0.028 y 0.017 | las clases leídas del repositorio, comparadas con los dos literales; con otro valor, rojo | `ci:internal/evals/conjunto_test.go:TestEvalsDelRepositorio` |
| FR-078, SC-008 | 4 de 4: cada uno de los 2 mutantes hace fallar un test de cada skill | los casos de las dos filas de FR-072 y FR-073, que la tarea de los mutantes ve fallar en una copia fuera del árbol | `ci:internal/evals/umbrales_test.go:TestUmbralConjunto`, `ci:internal/evals/informe_test.go:TestUmbralesDeJurisprudencia` |
| FR-033 (en `make ci`) | la medida versionada corresponde, con 0 y 0, en las dos skills | la comprobación sin modelo, que no cambia | `ci:internal/evals/medida_test.go:TestMedidaVersionada`, `ci:internal/evals/ejecucion_test.go:TestEjecucionSinMedir` |
| FR-021, FR-035 | 4 de 4 copias idénticas en cada carpeta del juez; los esquemas publicados y los peores casos del job, sin cambiar | los dos tests y `schema-check`, sin cambiar lo que esperan | `ci:internal/evals/medida_test.go:TestCopiasDelJuez`, `ci:internal/evals/definicion_test.go:TestDefinicionDelJob`, `ci:Makefile:schema-check` |
| FR-040, SC-014 | 1 línea por clase y 0 con un umbral o un veredicto en la salida del sondeo | los tests del sondeo, sin cambiar sus líneas | `ci:internal/evals/sondeo_test.go:TestSalidaDelSondeo` |
| FR-036 | las reglas del conjunto de las tres skills, sin defectos | `skills-check`, que no cambia | `ci:Makefile:skills-check` |
| FR-070, SC-010 | `make ci` en verde | el objetivo entero | `ci:Makefile:ci` |

Sin fila: FR-010, que no es un umbral que decida; SC-012, que lo mide una persona; y SC-013, que comprueba la revisión
final por lectura (FR-055).

**El recorrido del umbral nuevo.** Con 4 marcadas de 108 en `boe-legislacion`: `umbralesConjuntos` da `cumple: false`
con `decide: true`; `motivosDeLosUmbrales` da su motivo; el veredicto es `fallo`; el trabajo `evals (boe-legislacion)`
sale en rojo; `medir_cierre` lo cuenta y guarda el informe; y el informe final lee `umbrales` por el nombre exacto de la
fila y lo da como no cumplido (research V11). Con 3, todo lo anterior en verde, y «comprobado por su control». En
`make ci`, el mismo recorrido lo hace `TestUmbralConjunto` con votos grabados, y los dos mutantes prueban que el test
falla si el conjunto deja de sumar los modos o de decidir (research M9).

## Uso, de fuera adentro (criterio de uso, ADR 0028)

El detalle, con los ejemplos y sus bytes, está en contracts/umbrales-del-informe.md §2 a §5 y §7. Ninguna skill pide
ninguna de estas salidas ni las lee un `SKILL.md`: son 0 invocaciones por pregunta. Nada crece con lo consultado en
meses de uso —cientos de normas y miles de bloques—: cada job parte de las evals del repositorio y mide de nuevo.

- **`umbrales`** (el job, una vez; el informe final, una vez por medición del cierre; la persona): 13 elementos, uno
  más que hoy. 3 822 bytes en `boe-legislacion` y 4 099 en `jurisprudencia`, medidos; eran 3 482 y 3 757. El conjunto
  mide 335 bytes. Fijo: depende de las clases y de los modos. El `cumple: false` del conjunto deja de darse en el
  primer job en que las marcadas caben en la tasa; el de un recuento por modo, que no decide, cuando caben las de ese
  modo.
- **El motivo del conjunto** (la persona y la reparación del cierre, que no cambia por él ninguna skill; como mucho uno
  por skill y por job): 90 bytes de cabecera y unos 440 por respuesta marcada, con las frases del último cierre. Con
  las 108 y las 60 respuestas juzgadas, los más cortos que pueden darse llevan 4 respuestas en `boe-legislacion`,
  1,9 KB, y 2 en `jurisprudencia`, 1 KB; con menos juzgadas pueden llevar menos: 3, de 107 o de 104, y 1, de 58.
  Como mucho, las juzgadas: 108 y 60. Deja de darse en el primer job en que el conjunto se cumple; no se acumula.
- **La fila de `informe.md`** (la persona): una más, 85 bytes.
- **Las líneas del sondeo** (quien lo lanza, una vez por sondeo): una por clase, como hoy. No cambian.
- **El `umbral` de `clases.yaml`** (el job y `make ci`, una vez por ejecución; la persona que lo estrecha): un número
  por clase. No da señales.
- **`mutantes.md`** (la revisión final, una vez): los dos mutantes, con la línea que cambia y los subtests que fallan.
  No crece.
- **Los cinco documentos** (quien contribuye o lee un informe): las frases y las filas que hoy dicen el umbral de
  antes. No dan señales.

## Decisiones

- **El conjunto es la suma de los recuentos por modo**, en un método de `juicioDelJuez`, y `umbralesDelInforme` lo
  coloca (D1). No puede contar otra cosa que ellos.
- **El recuento por modo lleva `decide: false` siempre** (D2), y el nombre del conjunto es `<clase>:<modelo>` (D3).
- **El motivo escribe ` (modo <modo>)` detrás de cada sesión** (D4); la rama que nombraba las marcadas de un recuento
  por modo, que ya ningún umbral que decide alcanza, se retira.
- **Los tests de `boe-legislacion` usan la ejecución sintética que ya hay, con la declaración de clases del
  repositorio** (D6): no se arma un arnés nuevo con sus 21 evals, y el umbral que comparan es el dato que leerá el job.
- **Las 104 respuestas salen de cuatro sesiones cortadas tras su respuesta** (D7).
- **`jurisprudencia` se prueba en su tabla de hoy** (D8), y **el test de los datos es `TestEvalsDelRepositorio`** (D9).
- **La respuesta sin juzgar y la clase que solo se publica las fijan casos que ya existen** (D10).
- **Los mutantes se aplican en una copia fuera del árbol**, con el escenario 6 del quickstart, y la tarea deja su
  registro (D11).
- **El orden: el mecanismo, los valores, los mutantes, los documentos** (D12).
- **Las dos `clases.yaml` cambian el valor y una frase de su cabecera**, con el diff del contrato (D13).
- **El sondeo no cambia ni gana un test** (D16).
- **Los comentarios de `internal/evals` que el cambio deja falsos se corrigen con él; el de `evals.yml`, no** (D17).
- **La comparación sigue sin tolerancia**: no hay empate alcanzable (D19).
- Las demás, en research D1-D20.

## Trazabilidad: cada mecanismo y su requisito

| Mecanismo | Requisito |
|---|---|
| `umbralesConjuntos` y su llamada en `umbralesDelInforme` | FR-001, FR-002, FR-003, FR-005 |
| `nombreDelUmbralConjunto` y la descripción del conjunto | FR-001 |
| `Decide: false` en `umbralesDeLasClases` | FR-010, FR-011 |
| `marcadasDelUmbral` con el nombre del conjunto y el modo de cada respuesta | FR-004 |
| Los dos valores y la frase de su cabecera | FR-020 |
| `TestUmbralConjunto` y sus siete casos | FR-072, FR-074; SC-002, SC-004 |
| El campo `clasesDelJuez` del arnés de los tests | FR-072: la declaración del repositorio en la ejecución sintética |
| Los tres casos de `TestUmbralesDeJurisprudencia` | FR-073; SC-003 |
| Los umbrales y los motivos esperados de los tests que ya hay | FR-005, FR-012, FR-075, FR-076; SC-005, SC-006, SC-009 |
| Los dos literales de `TestEvalsDelRepositorio` | FR-022, FR-077; SC-007 |
| Los mutantes en una copia y `mutantes.md` | FR-078; SC-008 |
| Los comentarios de `umbrales.go`, `doc.go` e `informe.go` | FR-012: dicen cuántos umbrales hay y cuáles deciden |
| Los cinco documentos | FR-050 a FR-055; SC-013 |

Nada del diseño atiende a un estado que no pasa el umbral de materialidad: una `clases.yaml` que no se puede leer o
con un `umbral` fuera de 0 a 1 sigue la regla que ya hay —un fichero mal formado, con su motivo—, sin caso propio; y
un plan sin ningún modo no es un plan que el job pueda pedir.

## Datos externos

Ninguno (research D15): ni fuente ni grabación. Ningún manifiesto `grabaciones.json` ni test `TestGrabar*` cambia, y
el paso `grabar_datos` no tiene nada que grabar. Los dos valores no vienen de ninguna fuente externa: los fija el ADR
0038, que está en el repositorio, y llegan a los ficheros con el diff del contrato.

## Orden de implementación (de dentro afuera)

1. **El mecanismo** (`umbrales.go`, con los comentarios de `doc.go` e `informe.go`): el conjunto, el recuento por modo
   sin decidir y el motivo con el modo. Con los valores todavía en 0, ningún veredicto cambia —0 de 108 es 0 de 54 en
   cada modo—, y los siete tests que esperan la forma de antes (research M2) pasan a esperar la nueva: trece elementos,
   nueve que deciden, el motivo del conjunto. Aquí quedan fijados FR-075 y FR-076.
2. **Los valores**, en una tarea `[datos]`: las dos `clases.yaml` con `git apply` del diff del contrato;
   `TestEvalsDelRepositorio` con los dos literales; el campo `clasesDelJuez` del arnés y `TestUmbralConjunto`; y en
   `TestUmbralesDeJurisprudencia`, los casos de una y de dos marcadas y los dos elementos del contrato.
3. **Los mutantes**: el escenario 6 del quickstart sobre la cabeza, y su registro en
   `specs/021-h26-el-umbral-de/mutantes.md`.
4. **Los documentos**: los cinco, con contracts/documentos.md.

**Obligaciones para `tasks.md`**:

- ninguna tarea `[aceptacion]`; la tarea del paso 2 lleva `[datos]`, y ninguna otra;
- cada tarea deja `make ci` en verde: por eso el paso 1 lleva el código con todos los tests que cambian de forma, y el
  paso 2, los valores con todo lo que depende de ellos;
- cada fila `ci:` de «Controles de umbral» tiene la tarea que construye o actualiza su control, con un caso que lo ve
  fallar; los de FR-072 y FR-073, además, con los mutantes del paso 3;
- la tarea de los mutantes declara `specs/021-h26-el-umbral-de/mutantes.md`, no cambia nada bajo `internal/` y termina
  con el worktree retirado;
- la tarea de los documentos declara `CHANGELOG.md`, `CONTRIBUTING.md`, `docs/JURISPRUDENCIA.md`, `docs/WORKFLOW.md` y
  `CLAUDE.md`, y ningún otro;
- ninguna tarea ni corrección cumple un umbral cambiando su valor, dejándolo en `decide: false`, sacando respuestas de
  su total, ni cambiando la rúbrica, los casos o la medida (FR-061);
- ninguna tarea toca una skill, una eval, `schemas/`, `testdata/`, `.github/`, `scripts/`, `.specify/`,
  `docs/ROADMAP.md`, un ADR, `docs/USO.md`, `README.md`, la web, `specs/001-…` a `specs/020-…` ni el binario fuera de
  `internal/evals` (FR-036, FR-055, FR-062); ninguna escribe en la carpeta de evidencias ni nombra en su línea una
  ruta de ella o de `schemas/`;
- ninguna usa la red, abre una sesión con modelo ni ejecuta `make evals`, `make evals-sondeo`, `make evals-medir-juez`,
  `scripts/evals*.sh`, `TestEjecucionDelJob`, `TestMedidaDelJuez` o `TestSondeo` (FR-060);
- ninguna publica ni mide en la plataforma, y los escenarios 9 y 10 del quickstart no son tareas.

## Complexity Tracking

| Desviación | Por qué hace falta | Alternativa más simple rechazada |
|---|---|---|
| «Aceptación e2e: no aplica» (principio III: «cada hito empieza por el test e2e») | El hito no cambia el binario. La aceptación es la del job (SC-001), y cada pieza entra con sus tests en `make ci`. Precedentes: H24 y H25 | Un guion que ejerciera el binario sin cambios: no describiría la entrega |

## Comprobación contra la rúbrica del juez (`juez_plan`) y `precheck.sh plan`

- `precheck.sh plan`: existen `plan.md` y `research.md`; ninguna marca de aclaración pendiente; están
  `## Constitution Check`, la línea «Aceptación e2e» y `## Controles de umbral`, con cada fila nombrando requisitos del
  spec y un control con forma. Ejecutado en esta sesión: sin defectos.
- a · Constitution Check: un ítem por principio (I-IX) y por regla de dependencia (R1-R6).
- b · Dependencias: ninguna nueva (research V2).
- c · Reglas de dependencia: tabla R1-R6; ningún import nuevo.
- d · Errores y códigos: ningún error ni código de `kitlegal` cambia; el job y el sondeo salen como hoy.
- e · Tests primero: «Aceptación e2e: no aplica» con el motivo; un test nuevo, ocho que cambian, los que no cambian, los
  puntos de entrada y los fixtures.
- f · Alcance: nada fuera del spec. Los comentarios de `internal/evals` cambian porque el hito los deja falsos.
- g · Sin atajos: ningún `//nolint`, `t.Skip`, TODO ni error silenciado previstos.
- h · Mejor alternativa: cada decisión con la rechazada (research D1-D20).
- i · Afirmaciones verificadas: V1-V23 con fichero y línea u orden; M1-M16 medidas en esta sesión; S1-S5 como
  supuestos.
- j · Quickstart ejecutable: órdenes, rutas y nombres reales; dice qué vale ya hoy y qué deja cada escenario; las
  órdenes de lectura del 1, el 5, el 7 y el 9 se han ejecutado sobre el árbol de hoy, y las del 6, sobre el prototipo y
  sobre una copia de la cabeza (research M10, M13 a M16).
- k · Datos externos: ninguno.
- l · Autonomía: ninguna tarea para una persona; el sondeo y la medida no los lanza el run; el cierre en la plataforma
  lo hace el workflow.
- m · Uso: «Uso, de fuera adentro» y contracts §7, con bytes medidos, veces por job y lo que apaga cada señal.
- n · Proporcionalidad: «Trazabilidad»; sin tipos, paquetes ni ficheros Go nuevos; una rama retirada.
- o · Controles de umbral: los dieciocho que deciden en el cierre y una fila por cada umbral medido en `make ci`.
