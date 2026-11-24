# Implementation Plan: H45.1 · Los tests del molde, sin el estado de hoy del repositorio (ADR 0062)

**Branch**: `043-h45-1-los-tests-del` | **Date**: 2026-11-24 | **Spec**: [spec.md](spec.md)

**Input**: Feature specification from `specs/043-h45-1-los-tests-del/spec.md`

## Summary

Cuatro tests dejan de fijar lo que hay hoy en `data/`, `evals/` y los `SKILL.md`, y con los datos de `main` comprueban lo
mismo. `TestCoberturaDelRepositorio` escribe las nueve tuplas de lo regulado aparte sin el texto del procedimiento y exige
que las filas de lo administrativo leídas sean, en su orden, de ellas: una fila que pasa a otro paquete sale de lo leído
sin poner el test en rojo (research D1). `mutantesDeLasOrdenes` cuenta la línea de la orden del paso 2 en el `SKILL.md`
de `normativa` que copia el caso (D2). `cuatro-evals-mas` mide una carpeta de 16 evals de modo, con el tope de la
definición (D3). `llevaNoCubierto` deja fuera las evals de lo no declarado (D4). La copia con un paquete más de H45 pasa
a estrechar una fila de la base, dar norma a otra, llevar una referencia en `normativa` y añadir cinco evals, corre los
dos tests que faltaban en su lista y comprueba con `cmp` que es igual a su copia congelada (D5). Un prototipo en un clon
lo midió entero: verde con `main` y con la copia, y en rojo con cada mutante y con cada test devuelto a `main` (research
P1). Ningún cambio de producción, de datos, de skills ni de `.github/`.

## Technical Context

**Language/Version**: Go 1.27 (`go1.27.4`, research V1).

**Primary Dependencies**: las fijadas, ninguna nueva: `stretchr/testify` y `rogpeppe/go-internal` v1.16.0 (testscript:
`cmp`, `grep`, `stdout`; V4) en los tests; biblioteca estándar (`os.Root`, `strings.Lines`; V2) en los tests y en el
programa del guion.

**Storage**: ningún fichero de `data/`, `evals/`, `skills/` ni `schemas/` cambia. Cambian tres ficheros de test de Go y
el guion `h45-paquete-de-mas` con su copia congelada (`testdata/`, tarea `[datos]`). Nada en la caché ni en el grafo.

**Testing**: `go test -race` en `make ci`; el guion testscript corre `go test` en una copia del módulo en su `$WORK`, sin
red (`GOPROXY=off`), como hoy. Sin modelo. Skills de Go en research D9.

**Target Platform**: Linux y macOS, donde corre `make ci`.

**Project Type**: CLI multicall con skills empotradas; el cambio es solo de tests (`internal/skills`, `internal/evals`) y
de un guion de `internal/app`.

**Performance Goals**: ninguno del spec. El guion, 7,87 s y 8,03 s de reloj en el prototipo (research P1);
`TestControlesSobreLasReferenciasAMano` entero en la copia, 0,40 s a 0,66 s.

**Constraints**: con `main`, ninguna salida del binario ni referencia distinta (FR-014); cada tarea deja `make ci` en
verde (research V8, D6); ningún `.go` de producción en el diff (US1.3).

**Scale/Scope**: 3 ficheros de test (`internal/skills/cobertura_test.go`, `internal/evals/conjunto_test.go`,
`internal/evals/definicion_test.go`); 1 función de test nueva sin exportar (`lineaDeLaOrdenDelPasoDos`) y 1 firma que gana
`t` (`mutantesDeLasOrdenes`); en el programa del guion, 1 variable, 1 constante y 2 funciones (`cambiarUnaVez`,
`lineaDeLaOrden`) y 2 ficheros nuevos en su archivo; 2 guiones `.txtar` iguales; el registro `mutantes.md`.

## Constitution Check

*GATE: Must pass before Phase 0 research. Re-check after Phase 1 design.*

| Principio o regla | Cómo lo cumple este hito |
|---|---|
| I · Fuentes públicas y frontera humana (INNEGOCIABLE) | Ninguna petición HTTP: el hito no graba nada («Datos externos»); la copia del guion corre sin red y solo usa lo ya grabado de `BOE-A-2015-11722`, bloque `a1`. Ninguna identidad, POST ni CAPTCHA |
| II · Nada sin cita ni fuente | Ningún sobre, cita ni declaración de no cubierto cambia (FR-014). Los tests siguen exigiendo la norma que rige cada fila de lo regulado aparte de la base (L06, FR-003) y que cada eval de lo regulado aparte nombre el precepto de una fila (FR-009); a las de lo no declarado, que no llevan precepto, no se les pide |
| III · Tests primero y offline | «Aceptación e2e: no aplica» (abajo, con el motivo; Complexity Tracking): el hito no cambia el binario, y su aceptación es la copia `h45-paquete-de-mas` cambiada en una tarea `[datos]`, que con cada test devuelto a `main` sale en rojo (SC-005, P1). Todo offline, en `make ci`; ningún esquema cambia; la cobertura no cambia (solo código de test) |
| IV · Arquitectura hexagonal con reglas ejecutables | Ningún código de producción: ningún paquete, error, código de salida ni `panic` nuevo. Los tests fallan con `require`/`assert`; el programa del guion sale con 1 si no puede instalar, como hoy |
| V · Simplicidad y dependencias fijadas (YAGNI) | Ninguna dependencia ni paquete nuevo; cada mecanismo con su requisito («Trazabilidad»). Rechazados (research D1-D6): el esperado leído en crudo, quitar tuplas repetidas, exigir que cada tupla que falta tenga su fila de otro paquete, un campo nuevo en los mutantes, vaciar la carpeta de evals, más subtests de `TestDefinicionDelJob` en la copia, un guion de aceptación nuevo, `CHANGELOG.md` |
| VI · Un binario, convenciones de agente | Ninguna orden, bandera ni argumento cambia; `--describe` no cambia |
| VII · Grafo y privacidad | Nada toca el grafo ni la caché; ningún dato de nadie |
| VIII · Skills primero; el binario es la herramienta | Protege el molde de las skills de capacidad: un paquete (H42 y los que sigan) entra con datos, referencias y evals sin que un test ajeno caiga. Ninguna skill, referencia ni eval cambia (FR-014) |
| IX · Genericidad territorial, validación local | Ningún caso por municipio; la copia sigue corriendo `h36-plazos-de-hoy`, con Leganés, como hoy |
| X · Genericidad por materia, validación administrativa | Es el objeto del hito: los tests dejan de fijar con lo de hoy qué filas, líneas y evals hay; el paquete de una fila sale del `vertical` de `data/normas.yaml` y ningún código nombra una materia nueva. Las nueve tuplas a mano son las de lo administrativo, la materia de validación, que el spec permite escribir sin su texto (FR-001). La copia lo valida con un paquete sobre una norma real que estrecha una fila, da norma a otra —de la DA 1ª.2.c, con tuplas repetidas— y lleva referencia en dos skills |
| R1 · `internal/core/**` no importa el kernel, adaptadores ni entrada y salida | `internal/core` no cambia |
| R2 · solo `internal/httpx` importa `net/http` | Ningún import nuevo |
| R3 · solo `internal/{cache,store,graph}` importan SQLite y `database/sql` | Ningún import nuevo |
| R4 · solo `internal/cli` y `cmd/` llaman a `os.Exit` (`forbidigo`) | Ningún `os.Exit` nuevo en el módulo; el del programa del guion vive en el `.txtar`, fuera del módulo, como hoy |
| R5 · solo `internal/render` escribe en stdout; logs con `log/slog` a stderr (`forbidigo`) | Nada del módulo escribe en stdout; el programa del guion, que imprime sus pasos, vive fuera del módulo |
| R6 · `internal/graph` no importa `internal/source` ni `internal/render` | `internal/graph` no cambia |
| R7 · solo el paquete del servidor importa el SDK de MCP | Ningún import nuevo |
| Garantía sin red de `TestArquitectura` (`paquetesSinRed`, `ficherosSinRed`) | Solo cambian ficheros `_test.go`, que no entran en el binario, y ninguno gana una importación: usan `fmt`, `os`, `path/filepath`, `strings`, `time` y `maps`, que ya importaban (P1, `go vet` y `golangci-lint` en verde) |

Re-evaluado tras la fase 1: sin violaciones sin justificar; la única desviación, en Complexity Tracking.

## Project Structure

### Documentation (this feature)

```text
specs/043-h45-1-los-tests-del/
├── spec.md
├── plan.md                              # este fichero
├── research.md                          # P1, D1-D9, V1-V13, datos externos, supuestos
├── data-model.md                        # la tupla, la orden del paso 2, la carpeta de 16, el paquete de la copia
├── quickstart.md                        # Q1-Q8
├── contracts/
│   ├── tests-genericos.md               # FR-001 a FR-009: cada test, su forma, su mensaje y sus rojos
│   └── copia-con-un-paquete-mas.md      # FR-010 a FR-013: el programa, el archivo y las comprobaciones del guion
├── mutantes.md                          # registro de SC-004, SC-005 y SC-007; lo escribe la última tarea
└── tasks.md                             # /speckit-tasks
```

### Source Code (repository root)

```text
internal/skills/
└── cobertura_test.go                    # TestCoberturaDelRepositorio (contracts/tests-genericos.md §1)
internal/evals/
├── conjunto_test.go                     # mutantesDeLasOrdenes(t), lineaDeLaOrdenDelPasoDos (§2); llevaNoCubierto (§4)
└── definicion_test.go                   # probarCuatroEvalsMas, el subtest cuatro-evals-mas (§3)
internal/app/testdata/script/
└── h45-paquete-de-mas.txtar             # [datos] (contracts/copia-con-un-paquete-mas.md)
specs/042-h45-el-molde-admite/aceptacion/
└── paquete-de-mas.txtar                 # [datos], el mismo fichero
```

**Structure Decision**: la de `docs/ROADMAP.md` §2 y `CLAUDE.md`, sin paquetes ni ficheros de Go nuevos: los tests siguen
junto a lo que prueban y el guion en `internal/app/testdata/script/`.

## Diseño

1. **`TestCoberturaDelRepositorio`** (FR-001 a FR-004; contracts/tests-genericos.md §1; research D1): las nueve tuplas a
   mano, el filtro de la base sobre lo leído y la subsecuencia.
2. **La línea del paso 2** (FR-005; §2; D2): `mutantesDeLasOrdenes(t)` y `lineaDeLaOrdenDelPasoDos`.
3. **`cuatro-evals-mas`** (FR-006 a FR-008; §3; D3): la carpeta de 16 copias y el tope de la definición.
4. **`llevaNoCubierto`** (FR-009; §4; D4).
5. **La copia con un paquete más** (FR-010 a FR-013; contracts/copia-con-un-paquete-mas.md; D5), en una tarea `[datos]`
   que cambia los dos `.txtar` a la vez.
6. **El registro de mutantes** (SC-004, SC-005, SC-007; D7), en un clon del temporal.

Orden de las tareas (cada una deja `make ci` en verde; D6): 1, 2, 3 y 4, que con los datos de `main` comprueban lo
mismo que hoy; 5, que solo pasa con 1 a 4 hechos; 6 al final.

## Uso, de fuera adentro

Ninguna skill pide nada nuevo y ninguna `SKILL.md` invoca nada por este hito: cero invocaciones por pregunta (FR-014).
Lo que cambia lo leen quien escribe un paquete y `TestEntregaDelHito`, nunca en tiempo de consulta; todo sale de `data/`,
`evals/` y `skills/`, no de la caché ni del grafo, así que con meses de uso diario —cientos de normas y miles de bloques
consultados— es igual.

| Salida | Quién la consume y cuántas veces | Tamaño | Cuándo deja de darse |
|---|---|---|---|
| Mensaje de `TestCoberturaDelRepositorio` | quien escribe un paquete, persona o run, una vez por `make ci` en rojo | 751 bytes con L06 (dos listas de tuplas, unos 40 bytes cada una; P1); lo acotan las filas de lo regulado aparte de `data/cobertura.yaml`, hoy 9 | cuando las filas de lo administrativo vuelven a ser, en su orden, de las nueve |
| Defecto esperado de los dos casos de `mutantesDeLasOrdenes` | ídem, al caer `TestControlesSobreLasReferenciasAMano` | el de hoy byte a byte con `main` (una línea, con la orden y su forma); solo cambia el número | cuando el control da la orden en su línea |
| Mensajes de `cuatro-evals-mas` | ídem | una línea con la cuenta de `peorCaso`, de 16 evals fijas | cuando la definición vuelve a dar cabida a 16 a su concurrencia y no a dos |
| Mensaje de `precepto-de-la-cobertura` | ídem | una línea por eval de lo regulado aparte con otro precepto (149 bytes la de la 07, P1) | cuando la eval nombra el precepto de una fila |
| Salida del guion `h45-paquete-de-mas` | `TestEntregaDelHito`, en cada `make ci` | cinco líneas, 109 bytes en verde (P1), más hasta 1 200 bytes por orden de `go` que no sale con 0, como hoy | cuando el test que cae vuelve a ser genérico |

## Trazabilidad

| Mecanismo | Requisito |
|---|---|
| `tupla` sin procedimiento y las nueve a mano | FR-001, FR-003 |
| el filtro de la base sobre lo leído, escrito en el test | FR-002 |
| la subsecuencia en `TestCoberturaDelRepositorio` | FR-002, FR-003 |
| `mutantesDeLasOrdenes(t)` y `lineaDeLaOrdenDelPasoDos`, con su cuenta propia | FR-005 |
| `evalsDeLaCarpeta = 16`, los `Remove` y las 16 copias; sin `topeEnMinutos` | FR-006, FR-007, SC-006 |
| `!FueraDeLoDeclarado` en `llevaNoCubierto` y el mensaje del `NotEmpty` | FR-009 |
| `cambiosDeLaBase` y `cambiarUnaVez` | FR-010 (a, b) |
| `paquete/prueba-de-normativa.md` en `skills/normativa/references/prueba.md` | FR-010 (c) |
| `96-` a `99-` y `paquete/fuera-de-lo-declarado.yaml` | FR-010 (d), «Assumptions» del spec |
| `ordenDelPasoDos`, `lineaDeLaOrden`, su línea en la salida y su `stdout` | US6.4, FR-010 |
| los `grep` de la fila estrechada, la del orden social y la línea en `normativa` | US6.3, US6.4, FR-010 |
| los dos filtros del programa | FR-011 |
| el `cmp` del guion con su copia congelada; la tarea `[datos]` que cambia los dos | FR-013 |
| `mutantes.md` | SC-004, SC-005, SC-007, FR-012 |

## Controles mecánicos que el hito añade o toca

Según «Gates» de la constitución, capa 1:

- **`make ci`**, sin objetivos nuevos: `make test` y `make test-integration` corren los cuatro tests cambiados y
  `TestEntregaDelHito` con el guion cambiado; `make skills-check`, igual y en verde; `make lint`, con la configuración de
  hoy, sin `//nolint` ni `t.Skip` (P1: `0 issues`).
- **Tests de contrato**: ningún esquema ni sobre cambia.
- **Fixtures** (tarea `[datos]`): `internal/app/testdata/script/h45-paquete-de-mas.txtar` y
  `specs/042-h45-el-molde-admite/aceptacion/paquete-de-mas.txtar`, iguales. Ninguna grabación ni `evidencias/`.
- **Test de arquitectura**: `TestArquitectura`, sin cambios.
- **Exit codes**: ninguno nuevo.

**Mutantes** (`mutantes.md`; quickstart Q8): cada uno en un clon del temporal; el registro dice la orden, lo esperado y lo
visto.

| Control | Mutante | En rojo |
|---|---|---|
| FR-003 | L06: fuera `data/cobertura.yaml:39` | `TestCoberturaDelRepositorio` («solo las 5 primeras») |
| FR-003 | L07: `LeerCobertura` con `Norma = nil` | `TestCoberturaDelRepositorio` (`expected: 9`, `actual: 0`) |
| FR-004, SC-007 | L04 y L05, con la sincronía | `h32-paquetes` (`no match for ^filas: 33 de 33$`) |
| FR-005 | P6: `defectosDeLasOrdenes(delSkillMd, nil)` | `forma-en-una-referencia-con-otros-bloques` y `forma-en-ninguna-referencia` |
| FR-005 | otra línea: `linea: numero + 1` en `tramosDeCodigo` | los mismos dos casos |
| FR-006, FR-007 | T1; L29 | `TestDefinicionDelJob/cuatro-evals-mas` |
| FR-008 | ocho evals de más en `plazos-y-territorio` | `/tope-con-las-evals-del-repositorio` (y `/cuatro-evals-mas` en verde) |
| FR-009 | L33; las tres de lo regulado aparte pasadas a lo no declarado | `TestEvalsDelRepositorio/precepto-de-la-cobertura` |
| FR-012, SC-005 | cada uno de los cuatro tests devuelto a `main` | `h45-paquete-de-mas` (`tests enteros: 1` o `subtests: 1`, con el test nombrado) |
| FR-013 | la copia congelada distinta | `h45-paquete-de-mas` (`cmp`: «… differ») |

**Aceptación e2e: no aplica.** El hito no cambia ningún comportamiento del binario: ningún applet, verbo, bandera,
salida, código de salida, skill, referencia ni eval (FR-014), así que no hay guion nuevo que describa su entrega; una
tarea `[aceptacion]` dejaría en `specs/043-h45-1-los-tests-del/aceptacion/` un guion que el workflow activaría como
`h45-1-…`, una segunda copia con un paquete más que el spec no pide (research D6). La aceptación del hito, la del spec,
es la copia `h45-paquete-de-mas` con los cuatro casos (US6 · FR-010 a FR-013 · SC-003), cambiada en una tarea `[datos]`
junto con su copia congelada de H45, en verde dentro de `make ci`, y `make ci` en verde con `main` (SC-001). Que la copia
ejerce los cuatro casos —que con cada test de `main` sale en rojo— lo registra `mutantes.md` (SC-005).

**Datos externos:** ninguno. No hay manifiesto `grabaciones.json` ni test `TestGrabar*`; el paso `grabar_datos` no tiene
qué hacer (research, «Datos externos»).

**Autonomía:** ninguna tarea pide nada a una persona ni publica la rama, abre la propuesta de cambio o mide CI y evals
remotas; el cierre lo hace el workflow tras la revisión final. Comprobar que el prototipo de H42 pasa sin ningún `.go` es
de la persona antes de lanzar H42, fuera de este run.

## Controles de umbral

| Requisito | Umbral | Control | Dónde |
|---|---|---|---|
| SC-001 | `make ci` con 0 fallos | `make ci` | `ci:Makefile:ci` |
| SC-002, FR-014 | 0 diferencias entre las referencias de `skills/` y lo generado desde `data/`, y 0 en las salidas que fijan los tests del binario | `make skills-check` (`TestSkillsDelRepositorio` y los demás de su lista), en rojo con una diferencia | `ci:Makefile:skills-check`, `ci:internal/app/skills_test.go:TestSkillsDelRepositorio` |
| SC-003, FR-011 | `sincronía`, `tests enteros` y `subtests` con 0 en la copia | `h45-paquete-de-mas`, en rojo si un código no es 0 | `ci:internal/app/e2e_test.go:TestEntregaDelHito` |
| FR-010 | la fila estrechada en `administrativo.md`; la del orden social en `prueba.md` y 0 veces en `administrativo.md`; la línea del paquete en el `SKILL.md` de `normativa`; la orden del paso 2, 1 línea más abajo | los `grep` y el `stdout` del guion | `ci:internal/app/e2e_test.go:TestEntregaDelHito` |
| FR-013 | 0 bytes de diferencia entre los dos guiones | el `cmp` del guion | `ci:internal/app/e2e_test.go:TestEntregaDelHito` |
| FR-003 | las 9 tuplas; 0 filas de lo administrativo leídas fuera de ellas o de su orden | `TestCoberturaDelRepositorio` | `ci:internal/skills/cobertura_test.go:TestCoberturaDelRepositorio` |
| FR-004, SC-007 | 33 de 33 filas de `c991808` con sus datos | `h32-paquetes` | `ci:internal/app/e2e_test.go:TestEntregaDelHito` |
| FR-005 | 1 defecto exacto por caso, en la línea de la orden en el `SKILL.md` comprobado (36 con `main`) | `TestControlesSobreLasReferenciasAMano` | `ci:internal/evals/conjunto_test.go:TestControlesSobreLasReferenciasAMano` |
| SC-006, FR-006, FR-007 | 16 evals exactas; peor caso ≤ tope de la definición a su concurrencia y > tope a dos | `TestDefinicionDelJob/cuatro-evals-mas` | `ci:internal/evals/definicion_test.go:TestDefinicionDelJob` |
| FR-008 | peor caso de las evals del repositorio ≤ tope | `TestDefinicionDelJob/tope-con-las-evals-del-repositorio`, sin cambios | `ci:internal/evals/definicion_test.go:TestDefinicionDelJob` |
| FR-009 | ≥ 1 eval de lo regulado aparte; 0 con un precepto que no es de una fila | `TestEvalsDelRepositorio/precepto-de-la-cobertura` | `ci:internal/evals/conjunto_test.go:TestEvalsDelRepositorio` |

SC-004, SC-005 y FR-012 no son umbrales de `make ci` ni del cierre: su control, por el spec, es el registro de mutantes
que escribe una tarea del run (`mutantes.md`), con la orden y lo visto de cada uno; sus rojos son los de los controles de
esta tabla. Que el diff contra `main` no lleve ningún fichero de `data/`, `evals/`, `skills/`, `.github/` ni de Go de
producción (US1.3, FR-014) lo hace cumplir el guardián de diff del workflow con las rutas que declara cada tarea
(`scripts/workflow/guardian-diff.sh`, que el hito no toca), fuera de `make ci`; su comprobación sin modelo es Q3.

## Complexity Tracking

| Violation | Why Needed | Simpler Alternative Rejected Because |
|---|---|---|
| «Aceptación e2e: no aplica» (principio III: «cada hito empieza por el test e2e») | El hito no cambia el binario (FR-014); su aceptación es la copia de H45 que el spec manda cambiar en una tarea `[datos]`, y que solo puede ir detrás de los cuatro tests porque cada tarea deja `make ci` en verde (research D6) | Un guion nuevo en `specs/043-…/aceptacion/`: el workflow lo activaría como `h45-1-…`, una segunda copia con un paquete más que el spec no pide y que doblaría su coste en `make ci`. Precedentes: H7.3, H24 y H31 |

## Decisiones

Todas en research.md, con la alternativa descartada: D1 (la subsecuencia de las nueve tuplas), D2 (la línea del paso 2,
contada aparte), D3 (la carpeta de 16 evals), D4 (`llevaNoCubierto`), D5 (la copia: qué filas, la referencia en
`normativa`, la eval de lo no declarado, la línea del paso 2 en la salida y el `cmp`), D6 (aceptación y orden de las
tareas), D7 (el registro en un clon), D8 (sin `CHANGELOG.md`), D9 (skills de Go). Lo medido, en P1.
