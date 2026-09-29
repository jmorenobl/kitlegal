# Implementation Plan: H7.2 · La consulta repetida con evidencia coherente, y respuestas sin la maquinaria interna

**Branch**: `012-h7-2-la-consulta-repetida` | **Date**: 2026-09-29 | **Spec**: [spec.md](./spec.md)

**Input**: Feature specification from `/specs/012-h7-2-la-consulta-repetida/spec.md`

**Modo**: desatendido. Las decisiones técnicas se tomaron con el «Criterio de decisión autónoma» de la constitución y
están en [research.md](./research.md) (D1-D20), cada una con su alternativa rechazada. Toda afirmación sobre Go,
`jsonschema`, Claude Code o el propio repositorio remite a la tabla V de research.md, comprobada en local y sin red; lo
que no se puede comprobar así son los supuestos S1-S4.

## Summary

H7.2 no toca el binario: arregla lo que H7 y H7.1 dieron a `boe-legislacion` en cuatro piezas.

1. **La eval de la consulta repetida sobre un artículo modificado de verdad** (research D12-D15): la 19 pasa a ser
   `19-lcsp-contrato-menor-redaccion-cambiada.yaml`, sobre el art. 118 LCSP (`BOE-A-2017-12902`, `a1-30`). Su grafo
   previo, `testdata/evals/grafo-previo/lcsp-a1-30-redaccion-original/`, es la grabación H4 del bloque sin la redacción
   de 2020, producida por código del repositorio (`-actualizar-derivadas`) y comprobada por `TestGrabacionesDerivadas`
   byte a byte contra esa derivación y, servida en lugar de la grabación, por lo que da `boe`: exactamente una
   redacción de la grabada, la de 20180309. La subprueba `grafo-previo` comprueba en `make ci` que la sesión preparada da
   el `version-obsoleta` 20180309 → 20200206. Se retiran la eval del art. 21 LPAC, su grafo previo escrito a mano y su
   párrafo sintético.
2. **`boe-legislacion` v0.1.2** (research, «Causa de raíz», D16): el ruido es un párrafo de transición que Sonnet 5
   escribe tras la última orden —`graph check` en 51 de 51 sesiones— y que `claude -p` entrega como principio de la
   respuesta; `SKILL.md` lo favorece con el paso 5 («traslada lo que encuentre»), una prohibición en otra sección y
   estrecha (10 de 34 la cumplen y narran igual) y sin porqué. Los cambios C1-C9 van a ese sitio, con ese alcance y con
   ese porqué, más FR-041 a FR-044.
3. **La lista de expresiones prohibidas** (research D1-D11): `evals/<skill>/expresiones-prohibidas.yaml` con su esquema,
   dos familias y 38 expresiones literales, comparadas con la tolerancia de H5.1 y delimitadas como palabra. `Juzgar` la
   aplica a las evals que activan la skill (una expresión, un motivo y la sesión no pasa); el informe publica las
   encontradas por sesión y el recuento por modelo. Calibrada: 35 de las 93 respuestas de H7.1 con el reparto de FR-084,
   0 en los bloques grabados y 0 en lo que `SKILL.md` enseña a escribir.
4. **El escenario del quickstart** (research D17-D18): binario, caché y skill del hito en un directorio temporal, dos
   `claude -p` con `--setting-sources project` para que cargue la skill de ese directorio y no la de la cuenta y con
   `CLAUDE_ENV_FILE` apuntando a un guion que pone el binario de ese directorio primero en el `PATH` de su Bash —antes,
   una comprobación sin modelo resuelve `kitlegal` como lo hará la conversación y no la abre si no es ese—, y
   `TestComprobarConsultaRepetida` (etiqueta `evals`) con el mismo código que el juicio.

No cambia ninguna decisión de arquitectura: no hay ADR nuevo, y `specs/010-…` y `specs/011-…` no se editan (FR-070).

## Technical Context

**Language/Version**: Go 1.27 (`go 1.27.0`, `toolchain go1.27.1`), `CGO_ENABLED=0`, `-trimpath`.

**Primary Dependencies**: las fijadas, sin ninguna nueva: `gopkg.in/yaml.v3` y `santhosh-tekuri/jsonschema` v6.0.3 (la
lista y su esquema, con el lector común de `internal/skills`), `stretchr/testify`; biblioteca estándar (`regexp`,
`encoding/xml`, `encoding/json`, `encoding/json/v2` del informe).

**Storage**: ninguno nuevo. Ficheros del repositorio: la lista y la eval (`evals/boe-legislacion/`), el esquema de la
lista (`schemas/`) y la derivada (`testdata/evals/grafo-previo/`). El grafo y la caché de la sesión, los de siempre en temporales.

**Testing**: `go test -race` en `make ci` (unitarios con tabla y `t.Parallel`; subpruebas de `TestEvalsDelRepositorio`
en `make skills-check`); puntos de entrada con la etiqueta `evals` fuera de `make ci`; el job de evals en la propuesta
de cambio.

**Target Platform**: los del repositorio; el binario distribuido no cambia (`internal/evals` no llega a él, V17).

**Project Type**: skill (`boe-legislacion`) + formato común de eval y job (`internal/evals`) + datos de eval.

**Performance Goals**: ninguno nuevo. Las 38 expresiones se compilan una vez; el calibrado lee un JSON de ≈ 300 KB.

**Constraints**: ningún test ni tarea usa la red; `SKILL.md` < 300 líneas; `internal/evals` y el formato de eval no
cambian más allá de FR-056; el escenario del quickstart no escribe fuera de su directorio temporal.

**Scale/Scope**: 19 evals de `boe-legislacion`, 93 sesiones por job (51 de Sonnet 5 y 30 de Haiku 4.5 en evals que
activan la skill); una lista de 38 expresiones.

## Constitution Check

*GATE: pasado antes de la fase 0 y re-evaluado tras el diseño (fase 1): sin violaciones; las desviaciones justificadas
están en Complexity Tracking.*

### Principios

| Principio | Cómo lo cumple H7.2 |
|---|---|
| I · Fuentes públicas y frontera humana | No toca ninguna fuente, ni `internal/httpx`, ni graba nada (FR-014; «Datos externos»). Ningún test ni tarea usa la red. En el quickstart §6, si una conversación pide `metadatos` pasados cinco minutos, el binario los pide al BOE por `internal/httpx`, como en cualquier uso (fuente con fila revisada). |
| II · Nada sin cita ni fuente | El sobre no cambia. La skill sigue citando `[<identificador>, bloque <id>]` y trasladando los avisos con su forma (FR-046); la derivada no inventa contenido legal: es la redacción original que el BOE publicó y la grabación trae (FR-010, FR-011). |
| III · Tests primero y offline | «Aceptación e2e: no aplica» (abajo, con el motivo): el hito no cambia el binario. Cada pieza entra con sus tests en `make ci`, offline contra `testdata/` y el informe versionado de H7.1; la aceptación es el job (SC-001) y el quickstart §6 (SC-002). Cobertura dentro de `codecov.yml`. |
| IV · Hexagonal y errores tipados | No cambia ningún paquete del binario ni ningún código de salida. `internal/evals` es la herramienta de desarrollo que ya compone `internal/app`; sus errores siguen siendo `error` con el fichero delante; una lista mal formada es un fichero mal formado, no un `panic`. |
| V · Simplicidad y dependencias | Ninguna dependencia nueva. Cada mecanismo se traza a un FR («Trazabilidad»); la firma de `Juzgar` no cambia (el valor cero de la lista es «sin lista», research D5). |
| VI · Un binario, convenciones de agente | El binario, sus banderas y `--describe` no cambian; la tabla de comandos de `SKILL.md` no cambia. |
| VII · Grafo y privacidad | Nada nuevo en el grafo. El del quickstart vive en el temporal y se borra; la lista y el informe no guardan datos personales. |
| VIII · Skills primero | Mejora `boe-legislacion` (v0.1.2) y la mide con sus evals: la eval nueva y la lista en todas las que la activan (SC-001). `SKILL.md` < 300 líneas, sin nombrar evals ni modelos. |
| IX · Genericidad territorial | Nada se particulariza para un municipio: ni la lista, ni la eval (una norma estatal), ni la skill. |

### Reglas de dependencia (`docs/ROADMAP.md` §2, constitución IV)

| Regla | Cómo la cumple |
|---|---|
| R1 · `internal/core/**` no importa adaptadores | Sin cambios en `internal/core`. |
| R2 · Solo `internal/httpx` importa `net/http` | Sin cambios; ningún fichero nuevo lo importa. |
| R3 · Solo `cache`/`store`/`graph` importan SQLite y `database/sql` | `internal/evals` sigue leyendo el grafo con `graph.Leer` (la subprueba `grafo-previo` ya lo hace) y ejecuta `graph check` en proceso con `app.AppletGrafo`; ningún fichero nuevo importa `database/sql` ni el controlador. |
| R4 · Solo `cli` y `cmd/` llaman a `os.Exit` | Sin cambios; los puntos de entrada son tests. |
| R5 · Solo `render` escribe en stdout | Sin cambios en el binario; los tests escriben por `testing`. |
| R6 · `internal/graph` no importa `source/*` ni `render` | Sin cambios. |

## Project Structure

### Documentation (this feature)

```text
specs/012-h7-2-la-consulta-repetida/
├── plan.md                  # este fichero
├── research.md              # causa de raíz, D1-D20, V1-V23, S1-S4
├── data-model.md            # lista, expresión encontrada, recuento, eval, redacción, derivada, comprobación, retirada
├── quickstart.md            # escenarios de validación; §6, la consulta repetida en Claude Code
├── contracts/
│   ├── lista-y-juicio.md            # fichero, esquema, lista, comparación, juicio, informe, controles, uso
│   ├── eval-y-derivada.md           # eval 19, derivada, control de derivaciones, FR-002, retirada
│   ├── skill-boe-legislacion.md     # v0.1.2: cambios trazados, lo que se queda, controles, uso, CHANGELOG
│   └── comprobacion-del-quickstart.md  # punto de entrada, condiciones, tests
├── checklists/              # del spec
└── tasks.md                 # la escribe /speckit-tasks
```

### Source Code (repository root)

```text
schemas/expresiones-prohibidas.yaml.json                 # nuevo [datos]
evals/boe-legislacion/
├── expresiones-prohibidas.yaml                          # nuevo: la lista
├── 19-lcsp-contrato-menor-redaccion-cambiada.yaml       # nuevo: la eval de la consulta repetida
└── 19-lpac-articulo-21-redaccion-cambiada.yaml          # se retira
testdata/evals/grafo-previo/
├── lcsp-a1-30-redaccion-original/GET_…_BOE-A-2017-12902_texto_bloque_a1-30.json   # nuevo [datos]
└── lpac-a21-version-anterior/                           # se retira [datos]
internal/evals/
├── prohibidas.go (+ _test)          # nuevo: ExpresionesProhibidas, lectura, ExtraerExpresionesProhibidas
├── consulta_repetida.go (+ _test)   # nuevo: comprobarConsultaRepetida (quickstart)
├── formato.go                       # Eval.Prohibidas
├── conjunto.go                      # LeerConjunto reconoce la lista; Conjunto.Prohibidas
├── juzgar.go                        # ResultadoDeEval.ExpresionesProhibidas, motivo, Pasa
├── informe.go                       # RecuentoDeExpresiones, Informe.ExpresionesProhibidasPorModelo, columna y sección de informe.md
├── doc.go                           # el paquete lee también la lista
├── job_test.go                      # TestComprobarConsultaRepetida, con -skill y sus cuatro banderas nuevas (etiqueta evals)
└── formato_test.go, conjunto_test.go, juzgar_test.go, informe_test.go, consultas_test.go
internal/app/grafo_test.go           # derivación, -actualizar-derivadas, lista de derivadas del grafo previo (su tipo), comparación carpeta a carpeta, TestGrabacionesDerivadasInventadas; fuera la entrada retirada
skills/boe-legislacion/SKILL.md      # v0.1.2
CHANGELOG.md, CONTRIBUTING.md        # lo que el hito deja falso
```

**Structure Decision**: la de `docs/ROADMAP.md` §2 y `CLAUDE.md`, sin paquetes nuevos: el formato de eval y el job en
`internal/evals`, el control de derivaciones en los tests de `internal/app`, los datos de eval en `evals/` y
`testdata/evals/`, la skill en `skills/`. Ningún fichero de producto del binario cambia.

## Aceptación e2e

**Aceptación e2e: no aplica.** El hito no cambia el binario —cambiarlo está fuera de alcance, y `cmd/kitlegal` no
enlaza `internal/evals` (V17; `TestDependenciasDelBinario`)—, así que no hay comportamiento del binario que un guion
`testscript` pueda describir. Su aceptación es la de la skill: el job de evals de cierre (SC-001: la eval nueva con
`⚠ REDACCIÓN MODIFICADA:` y su tasa, las expresiones por sesión, el recuento por modelo dentro del umbral, el veredicto
aprobado y `red` vacío) y el escenario del quickstart §6 (SC-002). En `make ci` la fijan, sin modelo:
`TestGrabacionesDerivadas` y `TestGrabacionesDerivadasInventadas` (US2.3, US2.4; FR-010-FR-013, SC-005), la subprueba
`grafo-previo` (US2.2; FR-002), `formato`/`conjunto` (FR-030, SC-007), `expresiones-calibradas` (US4.2; FR-084,
SC-003), `expresiones-en-los-bloques` (FR-085, SC-004), `expresiones-de-la-skill` (FR-043, FR-051), los tests de
formato, de `Juzgar` y del informe (US3.1-US3.9; FR-050-FR-055, FR-082, FR-083, SC-006) y los de las condiciones del
quickstart (US5.3; FR-061). Las evals nuevas y cambiadas no las escribe una tarea `[aceptacion]`: las escriben las
tareas del orden de implementación, con sus tests.

## Controles mecánicos que este hito añade o toca

### Objetivos del `Makefile`

Ninguno nuevo ni cambiado. Lo nuevo entra en los que ya hay: `test` (unitarios de `internal/evals` e `internal/app`),
`skills-check` (las subpruebas nuevas de `TestEvalsDelRepositorio`, la de `grafo-previo` ampliada, y `SKILL.md`).
`schema-check` no cambia: compara los esquemas que salen de `--describe`, y ni `eval.yaml.json` ni el de la lista salen
de ahí.

### Tests nuevos

| Test | Paquete | Cubre |
|---|---|---|
| Casos del esquema de la lista (compila; una lista válida valida; una sin una familia, no) | `internal/evals` (`formato_test.go`) | FR-050, FR-082 |
| `TestLeerConjunto`: con lista bien formada (en `Conjunto.Prohibidas` y en cada `Eval.Prohibidas`), con lista mal formada (un `FicheroMalFormado` que la nombra) y sin lista (como antes) | `internal/evals` | FR-050, FR-055, FR-082 |
| `TestExtraerExpresionesProhibidas`: orden de la lista, sin repetir, extremos de palabra | `internal/evals` (`prohibidas_test.go`) | FR-051 |
| `TestJuzgarLasExpresionesProhibidas`, con la lista del repositorio: sin expresiones; «Sin hallazgos en la memoria de consultas»; «te confirmé»; mayúsculas, blancos y énfasis; la línea `⚠ REDACCIÓN MODIFICADA:` con sus dos fechas; una eval de no activación; una eval sin lista | `internal/evals` | FR-051, FR-052, FR-083, SC-006 |
| Informe: expresiones por sesión (JSON y columna); recuento por modelo, igual con la sesión de la prueba de red; skill sin lista; lista mal formada → `fallo`; serie que decide con dos sesiones con expresiones → `fallo`, y una informativa → tasa publicada sin decidir | `internal/evals` (`informe_test.go`) | FR-053, FR-054, FR-055 |
| Subtests `expresiones-calibradas`, `expresiones-en-los-bloques`, `expresiones-de-la-skill` de `TestEvalsDelRepositorio` | `internal/evals` | FR-084, FR-085, FR-043, FR-051, SC-003, SC-004 |
| `TestCondicionesDeLaConsultaRepetida` (las condiciones del quickstart) | `internal/evals` (`consulta_repetida_test.go`) | FR-061 |
| `TestGrabacionesDerivadasInventadas`: las tres derivadas inventadas no pasan, nombradas | `internal/app` (`grafo_test.go`) | FR-012, SC-005 |

### Tests existentes que cambian

- `TestGrabacionesDerivadas` / `grabacionesDerivadas()`: la derivación y `-actualizar-derivadas`; la lista de derivadas
  del grafo previo, de su propio tipo, con la entrada nueva; fuera la retirada con `parrafoDeLaVersionAnterior`; y la
  comparación carpeta a carpeta, para que la carpeta, y no la clase de la entrada, decida la comprobación: una entrada de
  otra clase con carpeta de grafo previo falla nombrada (contracts/eval-y-derivada.md §3; FR-011, FR-012). Las cuatro
  del e2e, sin cambios en `grabacionesDerivadas()` (FR-013).
- `compruebaElGrafoPrevio` (subprueba `grafo-previo`): lee como la sesión y ejecuta `graph check` (§4 del mismo
  contrato; FR-002).
- `formato_test.go`, `consultas_test.go` y `juzgar_test.go`: los nombres retirados pasan a los nuevos y la eval sintética
  de la consulta repetida, a la forma de la nueva (FR-021; research D15). Ninguno se desactiva ni se salta.
- `informe_test.go`: los informes esperados ganan `expresiones_prohibidas` por sesión, `expresiones_prohibidas_por_modelo`
  y la columna y la sección de `informe.md`.

### Puntos de entrada fuera de `make ci`

`TestComprobarConsultaRepetida` (etiqueta `evals`) con la `-skill` de siempre y cuatro banderas nuevas, `-primera`,
`-segunda`, `-fecha-superada` y `-fecha-leida`, todas obligatorias (contracts/comprobacion-del-quickstart.md). Lo ejecuta la persona en el quickstart §6; `golangci-lint` lo lintea (V19).

### Fixtures, `testdata/` y `schemas/` (tareas `[datos]`)

- `schemas/expresiones-prohibidas.yaml.json` (nuevo), con los casos de `formato_test.go` que lo fijan.
- `testdata/evals/grafo-previo/lcsp-a1-30-redaccion-original/…a1-30.json` (nuevo), escrito con
  `-actualizar-derivadas`, con su comprobación.
- `testdata/evals/grafo-previo/lpac-a21-version-anterior/` (se retira), con su entrada.

### CI

Sin cambios en `.github/workflows/` (research D1). El job de evals se ejecuta en la propuesta de cambio (lo lanza el
workflow tras la revisión final, FR-087) y su informe publica lo nuevo.

## Uso, de fuera adentro (criterio de uso, ADR 0028)

Detalle, con ejemplos y bytes, en contracts/skill-boe-legislacion.md §4 y contracts/lista-y-juicio.md §7. Resumen:

- **La respuesta** (la persona; una por pregunta): 0 líneas sobre la comprobación —hoy ≈ 85 B de transición en 2 de
  cada 3 respuestas de Sonnet 5—; una línea `⚠ REDACCIÓN MODIFICADA:` de ≈ 161 B por bloque cuya redacción cambió (0
  en la mayoría de las preguntas, como mucho k × 161 B con k bloques leídos); ≈ 83 B si `graph check` falla. No crece
  con lo acumulado. La línea sale en la respuesta cuya lectura ve la redacción nueva y se apaga con la lectura siguiente
  (H7.1 FR 024); la frase de un fallo, solo en esa respuesta; nada de consultas anteriores sin `version-obsoleta`.
- **Invocaciones por pregunta**: las de v0.1.1, sin ninguna nueva —cada bloque una vez con `kitlegal boe`, una
  `graph check` por norma citada (≈ 300 B sin cambios, ≤ 3 800 B con cinco bloques cambiados, H7.1 SC 005)—.
- **El informe del job** (el job y la persona; una vez por job): por sesión, `[]` lo habitual y ≤ 1,2 KB con las 38;
  el recuento, dos elementos de ≈ 100 B. Se recalcula en cada job; no crece con el uso del kit.
- **La lista**: 38 expresiones, ≈ 1 KB, fija.

## Trazabilidad: cada mecanismo y su requisito

| Mecanismo | Requisito |
|---|---|
| `schemas/expresiones-prohibidas.yaml.json` | FR-050, FR-055 |
| `ExpresionesProhibidas`, lectura y reconocimiento en `LeerConjunto`; `Conjunto.Prohibidas`, `Eval.Prohibidas` | FR-050, FR-052, FR-053, FR-055, FR-061 |
| `ExtraerExpresionesProhibidas` y su expresión regular | FR-051 |
| `ResultadoDeEval.ExpresionesProhibidas`, motivo `expresión prohibida: `, `Pasa` | FR-052, FR-054 |
| `RecuentoDeExpresiones`, `Informe.ExpresionesProhibidasPorModelo`, columna y sección de `informe.md`, párrafo de skill sin lista | FR-053 |
| `evals/boe-legislacion/expresiones-prohibidas.yaml` | FR-050, FR-051 |
| Subtests `expresiones-calibradas`, `expresiones-en-los-bloques`, `expresiones-de-la-skill` | FR-084, FR-085, FR-043 y FR-051 |
| Eval `19-lcsp-contrato-menor-redaccion-cambiada.yaml` | FR-001-FR-004, FR-030 |
| Derivada `lcsp-a1-30-redaccion-original`, la derivación y `-actualizar-derivadas` | FR-010, FR-014 |
| Lista de derivadas del grafo previo y comparación carpeta a carpeta en `TestGrabacionesDerivadas`; `TestGrabacionesDerivadasInventadas` | FR-011, FR-012, FR-013, FR-081 |
| `compruebaElGrafoPrevio` ampliada | FR-002 |
| Retirada de la eval 19 anterior, su derivada, su entrada y su párrafo; tests adaptados | FR-020, FR-021 |
| `SKILL.md` v0.1.2 (C1-C9) | FR-040-FR-047 |
| `comprobarConsultaRepetida`, `TestComprobarConsultaRepetida` y sus banderas | FR-060, FR-061, FR-062 |
| `CHANGELOG.md`; `CONTRIBUTING.md` («Formato común de eval», «Job de evals»); `doc.go` | FR-048, FR-086; FR-050 (la documentación del formato) |

Nada del diseño atiende a un estado que no pasa el umbral de materialidad: una lista, una derivada o una respuesta que
no se pueden leer siguen las reglas que ya hay —fichero mal formado, sesión ilegible, test que falla— sin caso propio.

## Cambios de `SKILL.md` trazados a la causa de raíz (FR-045)

| Cambio (contracts/skill-boe-legislacion.md §1) | Causa (research, «Causa de raíz») | Requisito |
|---|---|---|
| C1 · La respuesta empieza por lo que se pregunta, en el paso 5 | La regla estaba lejos del paso que precede a la respuesta; 34 de 34 en el primer párrafo, 22 anuncian la respuesta | FR-040 |
| C2 · Todo lo que la respuesta no nombra | La regla solo nombraba la memoria de consultas: 10 de 34 la cumplían y narraban igual; 7 con «código 0» | FR-040 |
| C3 · El porqué: quien pregunta no ve las órdenes; sin hallazgos no se distingue leído de no leído | La regla no daba motivo | FR-040, FR-041 |
| C4 · El paso 5 manda decir el cambio solo si hay `version-obsoleta` | «traslada lo que encuentre» invitaba a informar del resultado vacío (5 de 34 repiten «trasladar», 9 de 34 hallazgos y avisos) | FR-040, FR-042 |
| C5 · Nada de otra conversación | La respuesta que remató «te habría confirmado antes» | FR-041 |
| C6 · Las fechas tal como las da el hallazgo, en la misma línea | Precisión de H7.1 FR 042 | FR-042 |
| C7 · Regla 7 sin la memoria, `graph` ni el código | La regla vigente pedía decir «la memoria de consultas» | FR-043 |
| C8 · Regla 2 por lo que significa, sin el código | Los códigos en la respuesta | FR-044 |
| C9 · La última viñeta de «Memoria de consultas» remite al paso 5 | Prohibición repetida, más estrecha y en otro sitio | FR-040, FR-047 |

## Datos externos

Ninguno (research, «Datos externos»): ni fuente ni grabación nuevas, ningún manifiesto `grabaciones.json` ni test
`TestGrabar*` cambia, y el paso `grabar_datos` no tiene nada que grabar. La eval nueva usa las grabaciones de H4 de
`BOE-A-2017-12902` (`boe.legislacion-consolidada`, fila revisada en `docs/SOURCES.md`), y su derivada la produce código
del repositorio a partir de ellas.

## Orden de implementación (de dentro afuera)

1. **`[datos]`** `schemas/expresiones-prohibidas.yaml.json` con los casos de `internal/evals/formato_test.go` que lo
   fijan.
2. `internal/evals`: `prohibidas.go` (tipo, lectura, `ExtraerExpresionesProhibidas`), `LeerConjunto` y
   `Conjunto.Prohibidas`, `Eval.Prohibidas`, `doc.go`, con sus tests.
3. `internal/evals`: `Juzgar` y el informe, con sus tests.
4. `evals/boe-legislacion/expresiones-prohibidas.yaml` con las subpruebas `expresiones-calibradas`,
   `expresiones-en-los-bloques` y `expresiones-de-la-skill`.
5. **`[datos]`** La derivada `lcsp-a1-30-redaccion-original`, escrita con `-actualizar-derivadas`, con la derivación,
   la lista de derivadas del grafo previo con su entrada (comprobaciones 1 y 2) y `TestGrabacionesDerivadasInventadas`
   en `internal/app/grafo_test.go`. La comparación de ficheros y entradas sigue siendo la de hoy, con las dos carpetas
   juntas: `lpac-a21-version-anterior` sigue siendo una entrada del e2e con carpeta de grafo previo.
6. La eval nueva en lugar de la 19 anterior; los tests que nombran lo retirado (`formato_test.go`, `consultas_test.go`,
   `juzgar_test.go`); `compruebaElGrafoPrevio` ampliada.
7. **`[datos]`** Retirada de `lpac-a21-version-anterior/` con su entrada y `parrafoDeLaVersionAnterior`, y, con ella,
   la comparación carpeta a carpeta de `TestGrabacionesDerivadas` (contracts/eval-y-derivada.md §3), que, introducida
   antes, dejaría `make ci` en rojo mientras existiera esa entrada.
8. `skills/boe-legislacion/SKILL.md` v0.1.2.
9. `consulta_repetida.go` con sus tests y `TestComprobarConsultaRepetida` en `job_test.go`.
10. `CHANGELOG.md` y `CONTRIBUTING.md`.

**Obligaciones para `tasks.md`**: ninguna tarea `[aceptacion]` («Aceptación e2e: no aplica»); cada tarea deja
`make ci` en verde; las tareas `[datos]` son las de los pasos 1, 5 y 7, con las rutas que les da este plan; ninguna
tarea usa la red, graba, publica ni mide en la plataforma; el escenario del quickstart §6 no lo ejecuta ninguna tarea
(FR-062).

## Complexity Tracking

| Desviación | Por qué hace falta | Alternativa más simple rechazada |
|---|---|---|
| «Aceptación e2e: no aplica» (principio III: «cada hito empieza por el test e2e») | El hito no cambia el binario: no hay comportamiento suyo que un guion describa. La aceptación es la de la skill (job y quickstart), y cada pieza entra con sus tests en `make ci` | Un guion que ejerciera el binario sin cambios sobre los datos nuevos: no describiría la entrega, y lo que diría (FR-002) ya lo dice la subprueba `grafo-previo` |
| Tres tareas `[datos]` con código de test: el esquema con los casos que lo fijan; la derivada con su derivación y su comprobación; la retirada de la derivada con la de su entrada y la comparación carpeta a carpeta | `TestGrabacionesDerivadas` exige que cada fichero de las carpetas de derivadas tenga su comprobación y cada comprobación su fichero, y la derivada solo existe si la escribe la derivación (FR-010); la comparación carpeta a carpeta rechaza la entrada retirada mientras exista: separarlos deja una tarea en rojo. Precedente: la derivada `version-ulterior` y el esquema de eval de H7.1 | Separarlos, que dejaría `make ci` en rojo entre dos tareas |

## Comprobación contra la rúbrica del juez (`juez_plan`) y `precheck.sh plan`

- `precheck.sh plan`: existen `plan.md` y `research.md`; ninguno conserva marcas de aclaración pendiente;
  `## Constitution Check` y la línea «Aceptación e2e» están.
- a · Constitution Check: un ítem por principio (I-IX) y por regla de dependencia (R1-R6).
- b · Dependencias: ninguna nueva.
- c · Reglas de dependencia: tabla R1-R6; nada del binario cambia.
- d · Errores y códigos: ningún código nuevo; lista mal formada = fichero mal formado (research D8).
- e · Tests primero: «Aceptación e2e: no aplica» con el motivo; «Tests nuevos», «Tests existentes que cambian»,
  fixtures y puntos de entrada.
- f · Alcance: nada fuera del spec; la documentación que se toca es la que el hito deja falsa (research D20).
- g · Sin atajos: ningún `//nolint`, `t.Skip`, TODO ni error silenciado previstos.
- h · Mejor alternativa: cada decisión con la rechazada (D1-D20).
- i · Afirmaciones verificadas: V1-V23 con fichero:línea, orden o prototipo; S1-S4 como supuestos.
- j · Quickstart ejecutable: todo en un temporal que se borra; lo que escribe en el árbol, declarado (§2 con el mismo
  contenido, §7 en ficheros ignorados).
- k · Datos externos: ninguno; la derivada la produce código a partir de la grabación de H4.
- l · Autonomía: ninguna tarea para una persona; el §6 lo ejecuta la persona al leer el informe final; el cierre en la
  plataforma lo hace el workflow.
- m · Uso: «Uso, de fuera adentro» y los contratos (bytes, invocaciones por pregunta, apagado de cada señal).
- n · Proporcionalidad: «Trazabilidad»; `Juzgar` sin cambiar de firma; sin mecanismos para estados sin vía real.
