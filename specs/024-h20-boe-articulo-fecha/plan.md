# Implementation Plan: H20 · `boe articulo --fecha`: la redacción vigente en una fecha + `boe-legislacion` v0.2

**Branch**: `024-h20-boe-articulo-fecha` | **Date**: 2026-10-19 | **Spec**: [spec.md](./spec.md)

**Input**: Feature specification from `/specs/024-h20-boe-articulo-fecha/spec.md`

**Modo**: desatendido. Las decisiones técnicas se tomaron con el «Criterio de decisión autónoma» de la constitución y
están en [research.md](./research.md) (D1-D22), cada una con su alternativa rechazada. Lo que se afirma del repositorio
y de las dependencias remite a su tabla V; lo medido en esta sesión, a su tabla M; y lo que no se pudo comprobar son
los supuestos S1-S3. El prototipo de esta sesión —un árbol desacoplado de `0b2a21d` bajo `$TMPDIR`, ya retirado— solo
midió: ninguna tarea parte de él.

## Summary

Hoy `kitlegal boe articulo` da la última redacción de un bloque aunque la respuesta del BOE las trae todas. H20 lo
resuelve en cinco piezas, de dentro afuera:

1. **El grafo acepta lo observado sin pisar lo que guarda.** `schema.Observado` y `core.Lote` ganan `SoloLoNuevo`; con
   él, la entrega a `world.db` escribe solo los nodos, las aristas y los textos que no tiene (research D9).
2. **La lectura fechada en el adaptador del BOE** (`internal/source/boe`). El lector del bloque guarda todas las
   versiones; `--fecha AAAA-MM-DD` elige la de mayor fecha de vigencia que no es posterior a la pedida, o sale con 3
   (FR-003, FR-004); `data` es el artículo de hoy de esa versión más `fecha_pedida` y, si la hay,
   `redaccion_posterior`; la lectura se guarda en una entrada nueva con las versiones del bloque, y emite solo su
   `BloqueVersion` y su texto, con `SoloLoNuevo` (contracts/articulo-fecha.md). La lectura sin fecha no cambia.
3. **La orden y la herramienta** (`internal/app/boe.go`): `--fecha` y el parámetro `fecha` de `boe_articulo`, que el
   registro genera solo; `schemas/bloque.json` regenerado.
4. **Las evals con fecha y su umbral** (`internal/evals`): `fecha` en un comando de bloque, `redaccion_fechada`, las
   evals 22 y 23, la caché preparada que sirve cualquier fecha y `redaccion_sin_fechar:<modelo>:<modo>`, con 0 y que
   decide: `boe-legislacion` pasa a quince umbrales, once que deciden (contracts/evals-redaccion-fechada.md).
5. **`boe-legislacion` v0.2** y los documentos: la sección «Redacción en una fecha» con tres formas fijas, en 298
   líneas (contracts/skill-boe-legislacion.md y su prototipo); `CHANGELOG.md`, `README.md`, `CONTRIBUTING.md` y la
   frase de `docs/WORKFLOW.md` que cuenta los umbrales que deciden en `boe-legislacion` (contracts/documentos.md).

No hay ADR nuevo, ni dependencia, ni fuente, ni grabación nueva (FR-006). No cambian `boe articulos`, los demás verbos,
`legal-core`, `jurisprudencia`, `graph check` fuera de lo dicho, `.github/workflows/evals.yml`, el juez ni su medida
(FR-059, FR-070).

## Technical Context

**Language/Version**: Go 1.27 (`go 1.27.0`, `toolchain go1.27.2`), `CGO_ENABLED=0`, `-trimpath` (research V1); YAML de
las evals; JSON Schema de `schemas/`; Markdown de `SKILL.md` y de los documentos.

**Primary Dependencies**: las fijadas, sin ninguna nueva (research V2): `alecthomas/kong` (la bandera),
`invopop/jsonschema` (el `--describe`), `santhosh-tekuri/jsonschema` (la validación de salidas y evals),
`modelcontextprotocol/go-sdk` (la herramienta, sin código nuevo), `modernc.org/sqlite` (caché y grafo, sin SQL nuevo),
`rogpeppe/go-internal` (testscript), `stretchr/testify` y `gopkg.in/yaml.v3`. Del estándar, `time` (`time.DateOnly`).

**Storage**: la caché SQLite de hoy, con una clase de entrada más (`versiones`, data-model §5); `world.db`, sin
migración ni tabla nueva: `SoloLoNuevo` cambia qué se escribe, no el esquema.

**Testing**: `go test -race -shuffle=on` en `make ci`: tests unitarios del adaptador contra lo grabado de H4 y con
respuestas construidas en el test; tests del almacén y de integración del grafo con lotes sintéticos en
`t.TempDir()`; el applet con el registro y el almacén reales; los guiones testscript del arnés de hoy; las evals con
sesiones sintéticas, sin ninguna sesión con modelo (FR-072). El job de evals, sobre la cabeza, lo lanza el workflow.

**Target Platform**: el binario, macOS, Linux y Windows como hoy; `make ci`, macOS y Linux; el job, `ubuntu-24.04`.

**Project Type**: CLI multicall con servidor MCP local, una skill sin código y el instrumento de medida de las evals.

**Performance Goals**: ninguno nuevo. Una lectura fechada pide lo mismo que una sin fecha (el bloque y sus metadatos) y
elige recorriendo sus versiones (nueve en el `a22`). El job: con veintitrés evals a cinco a la vez, 18 977 s de peor
caso, por debajo de 352 minutos (S1; lo comprueba `TestDefinicionDelJob`), y el objetivo de 900 s por tanda, que no
cambia.

**Constraints**: la salida sin `--fecha` byte a byte la de `main` (FR-012); la entrada de hoy sin cambiar (FR-021);
`SKILL.md` por debajo de 300 líneas y su `description` en 1 024 caracteres (FR-048; research V35); ningún `//nolint`,
`t.Skip` ni test desactivado; ninguna red en los tests; ningún cambio en `.github/`, `evidencias/`, `docs/SOURCES.md`
ni en lo grabado (`testdata/` solo gana lo que digan las tareas `[datos]`, y este plan no pide ninguno).

**Scale/Scope**: unas 30 líneas nuevas de `SKILL.md` en las mismas 299; dos evals; quince umbrales en
`boe-legislacion`; seis guiones de aceptación; los ficheros de «Project Structure».

## Constitution Check

*GATE: pasado antes de la fase 0 y re-evaluado tras el diseño (fase 1): sin violaciones y sin desviaciones que
justificar (Complexity Tracking, vacío).*

### Principios

| Principio | Cómo lo cumple H20 |
|---|---|
| I · Fuentes públicas y frontera humana | La lectura fechada pide, por `internal/httpx`, los mismos dos recursos GET que la de hoy, de la misma fuente con fila revisada (research V37); ninguna petición, fuente ni grabación nueva (FR-006). Nada de identidad ni de sedes |
| II · Nada sin cita ni fuente | El sobre de la lectura fechada es el obligatorio, con la dirección del bloque y su fecha de consulta; lo que entrega al grafo lleva la procedencia del sobre como `source` (FR-030). La skill cita la redacción de la fecha y dice en línea fija qué fecha pidió y cuál es la de hoy, sin describir lo que no leyó (FR-041 a FR-043) |
| III · Tests primero y offline | Seis guiones de aceptación congelados (abajo); tests unitarios offline contra lo grabado de H4; toda salida nueva validada contra `schemas/bloque.json` (FR-014). Cobertura: el código nuevo vive en `internal/source/boe`, `internal/graph`, `internal/cli` e `internal/evals`, con sus tests; `internal/core` solo gana dos campos y su copia, que cubren los tests del grafo |
| IV · Hexagonal y errores tipados | La elección vive en el adaptador (research D1); los dos 3 nuevos son `errorDeNoEncontrado`, y la fecha mal escrita `errorDeArgumentos`, que el kernel ya traduce a 3 y 2 (ADR 0023; research D7). Ningún `os.Exit`, `panic` ni escritura a stdout nuevos |
| V · Simplicidad y dependencias | Ninguna dependencia nueva. Cada mecanismo se traza a un requisito («Trazabilidad»). Un tipo exportado de salida (`ArticuloFechado`, con `RedaccionPosterior`), una función exportada (`ValidarFecha`, como `ValidarNorma`) y dos campos en contratos que ya existen (`Observado.SoloLoNuevo`, `Lote.SoloLoNuevo`). Nada para estados por debajo del umbral de materialidad: una versión con una fecha de vigencia ilegible cae en la misma regla que una sin ella (FR-004) |
| VI · Un binario, convenciones de agente | Una bandera nueva de un verbo que ya existe; `--describe` la anuncia y de él salen la herramienta y la tabla de la skill (FR-001, FR-085). Las banderas globales valen igual (FR-007) |
| VII · Grafo y privacidad | Ids naturales de hoy (ELI); ningún dato de persona; `world.db` solo gana lo que no tenía, sin cambiar lo que tenía (FR-031) |
| VIII · Skills primero | Mejora una skill medible, `boe-legislacion` v0.2, con dos evals que deciden y un umbral que la hace cumplir (FR-047, FR-054); va a Go solo lo determinista —la elección por fecha—, y la skill decide cuándo pedirla. `SKILL.md` sin código, por debajo de 300 líneas, con su tabla generada (FR-048) |
| IX · Genericidad territorial | Nada se particulariza para un municipio ni una norma: la regla de elección es la de cualquier bloque, y el umbral sale de las claves de cada eval (FR-055). El hito no tiene dimensión territorial: la matriz de municipios no aplica |

### Reglas de dependencia (`docs/ROADMAP.md` §2, constitución IV)

| Regla | Cómo la cumple |
|---|---|
| R1 · `internal/core/**` no importa adaptadores | `internal/core/schema`, `internal/core` (`graphstore.go`) e `internal/core/grafo` solo ganan un campo `bool` y su copia; ningún import nuevo |
| R2 · Solo `internal/httpx` importa `net/http` | La lectura fechada pide con el `Pedidor` de la invocación, como hoy; ningún import nuevo |
| R3 · Solo `cache`/`store`/`graph` importan SQLite y `database/sql` | La entrada nueva usa el puerto `core.Cache`; la entrega con `SoloLoNuevo`, el SQL de hoy en `internal/graph`; los tests del grafo, sus ayudantes de hoy |
| R4 · Solo `cli` y `cmd/` llaman a `os.Exit` | Ningún `os.Exit` nuevo |
| R5 · Solo `render` escribe en stdout | La salida legible es la tabla de `render` de hoy; ningún `fmt.Print*` ni `os.Stdout` nuevo |
| R6 · `internal/graph` no importa `source/*` ni `render` | `aplicar.go` lee `SoloLoNuevo` del `Consolidado` de `internal/core/grafo`; ningún import nuevo |

## Project Structure

### Documentation (this feature)

```text
specs/024-h20-boe-articulo-fecha/
├── plan.md                  # este fichero
├── research.md              # V1-V39, M1-M11, S1-S3, D1-D22
├── data-model.md            # consulta, versiones, elección, salida, entrada, grafo, evals, errores
├── quickstart.md            # §0-§8 en make ci o en un árbol aparte; §9 el cierre; §10 una persona
├── contracts/
│   ├── articulo-fecha.md                  # la orden y la herramienta: elección, golden, errores, salida, caché, grafo, uso
│   ├── evals-redaccion-fechada.md         # formato, las dos evals, líneas, umbral, los quince, tests, mutantes, uso
│   ├── skill-boe-legislacion.md           # las formas fijas, cada requisito, cómo cabe, lo que consume
│   ├── SKILL-boe-legislacion.prototipo.md # el SKILL.md prototipo, 298 líneas
│   └── documentos.md                      # qué pasaje cambia en cada documento
├── aceptacion/              # los seis guiones; los escribe la primera tarea
├── mutantes.md              # lo escribe la tarea de FR-090
├── checklists/              # del spec
└── tasks.md                 # la escribe /speckit-tasks
```

### Source Code (repository root)

```text
internal/core/schema/grafo.go          # Observado.SoloLoNuevo
internal/core/graphstore.go            # Lote.SoloLoNuevo
internal/core/grafo/lote.go            # Consolidado.SoloLoNuevo, copiado en Consolidar (+ lote_test.go)
internal/cli/entrega.go                # LoteDe copia SoloLoNuevo (+ entrega_test.go)
internal/graph/aplicar.go              # con SoloLoNuevo, lo guardado no se toca (+ almacen_test.go, integracion_test.go)
internal/source/boe/
├── bloque.go                          # bloqueLeido.versiones, versionLeida (+ bloque_test.go)
├── ids.go                             # ValidarFecha (+ ids_test.go)
├── consultas.go                       # ConsultaArticulo.Fecha
├── datos.go                           # ArticuloFechado, RedaccionPosterior
├── entradas.go                        # versionesDelBloque, claveDeLasVersiones, la unión datosDeEntrada
├── fecha.go                           # nuevo: la lectura fechada, la elección, sus dos mensajes (+ fecha_test.go)
├── articulo.go                        # la rama de articulo con Fecha
├── grafo.go                           # nodoDeLaVersion, observadoDeLaLecturaFechada (+ grafo_test.go)
└── doc.go                             # un párrafo del comentario de paquete
internal/app/boe.go                    # argumentosDeArticulo.Fecha; Salida de articulo, ArticuloFechado (+ grafo_test.go, boe_test.go)
schemas/bloque.json                    # regenerado desde --describe ([datos])
schemas/eval.yaml.json                 # comando-bloque-esperado, redaccion_fechada ([datos])
internal/evals/
├── formato.go                         # ComandoEsperado.Fecha, RedaccionFechadaEsperada, Eval.RedaccionFechada (+ formato_test.go)
├── juzgar.go                          # qué casa con un comando de bloque, con fecha y sin ella (+ juzgar_test.go)
├── fechadas.go                        # nuevo: las tres formas fijas y ExtraerRedaccionesFechadas (+ fechadas_test.go)
├── informe.go                         # medirLasRedaccionesFechadas; campos de respuestaDelJuez y respuestasMedidas
├── umbrales.go                        # umbralDeRedaccionSinFechar, su nombre, su descripción, su sitio y su motivo
├── consultas.go                       # la lectura de cada bloque con --fecha 9999-12-31 (+ consultas_test.go)
├── conjunto.go                        # las evals con fecha, aparte; maximoDeEvals 23 (+ conjunto_test.go)
├── doc.go                             # lo que dice de los umbrales de cada skill (comentario)
├── informe_test.go                    # TestUmbralesDeBoeLegislacion; TestUmbralesDeLasTresSkills
├── definicion_test.go                 # el literal peorCasoDeBoeLegislacion
└── plan_test.go                       # las series y sesiones de boe-legislacion en probarElPlanDelRepositorio
evals/boe-legislacion/22-lcsp-contrato-menor-en-2019.yaml
evals/boe-legislacion/23-lpac-articulo-21-en-2016.yaml
skills/boe-legislacion/SKILL.md
CHANGELOG.md, README.md, CONTRIBUTING.md
docs/WORKFLOW.md                           # la frase de medir_cierre con los umbrales que deciden en boe-legislacion
```

**Structure Decision**: la de `docs/ROADMAP.md` §2 y `CLAUDE.md`, sin paquetes nuevos. La lectura fechada va al
adaptador, junto a la lectura de hoy (research D1); la marca de la entrega, a los dos contratos que ya la llevan de un
sitio a otro (`Observado` y `Lote`) y al almacén (research D9); el umbral, junto a los demás umbrales de las respuestas
(research D14). `internal/mcp`, `internal/app/herramientas.go`, `internal/cache`, `internal/httpx`, `graph check`,
`internal/evals/juez.go`, `medida.go`, `sondeo.go`, `definicion.go`, el `Makefile` y `.github/` no cambian.

## Aceptación e2e

**Aceptación e2e:** seis guiones `testscript`, que escribe la primera tarea en `specs/024-h20-boe-articulo-fecha/aceptacion/`
y quedan congelados; el workflow los activa como `internal/app/testdata/script/h20-<nombre>.txtar`. Usan solo lo que
el arnés ya tiene (`exec`, `stdout`, `stderr`, `cmp`, `cp`, `rm`, `exists`, `env`, `mcp`, `exec sed`,
`exec sh -c '…; test $? -eq <n>'`, `KITLEGAL_BIN`, `KITLEGAL_T0_BIN`, lo grabado de H4 en `reproduccion/`). Cada uno
empieza por la precondición `exec kitlegal boe articulo --help` y `stdout '--fecha'` (`stdout` solo toma por bandera
un primer argumento `-count=`: `$(go list -m -f '{{.Dir}}' github.com/rogpeppe/go-internal)/testscript/cmd.go:662-676`,
v1.16.0), que hoy se ejecuta y falla por su aserción.

| Guion | Historia | Qué describe | Requisitos |
|---|---|---|---|
| `fecha-version.txtar` | US1, US2 | Con `KITLEGAL_T0_BIN` y `--json`: el `a1-30` el `2019-06-01` sale con 0, con `fecha_vigencia` `20180309`, `norma_modificadora` `BOE-A-2017-12902`, `fecha_pedida` y `redaccion_posterior` `20200206` y `BOE-A-2020-1651`, y la salida legible con sus filas; el `a22` el `2005-06-15` (`20040101`, `BOE-A-2003-23103`, posterior `20151031`) y el `2019-06-01`, con el mismo `hash_texto` que sin `--fecha` y sin `redaccion_posterior`; el `a21` el `2016-06-01` sale con 3, `no-encontrado`, sin texto, con `2016-06-01` y `20161002` en el mensaje | FR-001, FR-002, FR-003, FR-010, FR-013; SC-004 |
| `fecha-codigos.txtar` | US1, US2 | `--fecha` con `2019-02-30`, `2019-6-1` y vacía: 2, sin crear la caché; con una norma o un bloque mal formados, el mismo 2 que sin `--fecha`; con `a9999`, el mismo 3 y el mismo mensaje que sin ella, distinto del de una fecha anterior; con `--dry-run` y la caché vacía, 0 y ninguna escritura | FR-005, FR-007, FR-003 |
| `fecha-offline.txtar` | US1 | Tras la lectura sin fecha, la fechada con `--offline` sale con 4 (la entrada de hoy no la sirve); tras una fechada, la sin fecha con `--offline` da byte a byte lo de antes; sin `reproduccion/`, la misma lectura fechada con `--offline` da el mismo `data` y el mismo `hash`, y otra fecha del mismo bloque también sale con 0 | FR-020, FR-021, FR-022, FR-082 |
| `fecha-main.txtar` | US3 | Sin `--fecha`, `boe articulo BOE-A-1985-5392 a22`, con `--json` y sin él, y el 3 de `a9999` y el 2 de una norma mal formada, cada uno igual byte a byte a la salida de `main` que el guion lleva escrita (las fechas de consulta que no fija el reloj del arnés, igualadas con `exec sed`); y `boe articulos … --describe`, igual al de `main` | FR-012, FR-070, FR-083; SC-005 |
| `fecha-herramienta.txtar` | US1, US2 | `mcp serve` anuncia `boe_articulo`; con `fecha` `2019-06-01` sobre el `a1-30`, el sobre de la orden con `--json`; con `fecha` `2016-06-01` sobre el `a21`, el error `no-encontrado` con `20161002`; y el servidor sigue | FR-001, FR-084 |
| `fecha-grafo.txtar` | US4 | Con el `a22` leído sin fecha, la lectura del `2005-06-15` añade una `BloqueVersion` (`graph stats`) y deja `graph check BOE-A-1985-5392 a22` byte a byte igual; con `--no-graph` y con `--dry-run`, `graph stats` no cambia; la siguiente lectura sin fecha da el `graph check` del mismo mundo sin la lectura fechada | FR-030, FR-031, FR-032 |

La salida de `main` de `fecha-main.txtar` la escribe la primera tarea dentro del guion, como secciones del txtar,
desde lo que da hoy el binario del arnés, antes de que exista el código del hito (research D20). US5 no tiene guion: no
es comportamiento del binario; su aceptación son `TestUmbralesDeBoeLegislacion` y `TestJuzgarLecturasFechadas` en
`make ci` y el job de cierre (SC-001). La primera tarea no escribe ninguna eval: las dos entran con la tarea del
formato y del umbral, porque `make ci` valida toda eval de `evals/` (precedente: H9).

## Controles mecánicos que este hito añade o toca

### Objetivos del `Makefile`

Ninguno nuevo ni cambiado. `test`, `schema-check` y `skills-check` llevan más dentro: los tests de abajo,
`schemas/bloque.json` regenerado y la tabla regenerada de `boe-legislacion`.

### Tests nuevos

| Test | Fichero | Cubre |
|---|---|---|
| `TestValidarFecha` | `internal/source/boe/ids_test.go` | FR-005: las formas de research M5 |
| `TestArticuloEnUnaFecha` | `internal/source/boe/fecha_test.go` | FR-002, FR-003, FR-004, FR-010, FR-011, FR-013, FR-081, SC-004: los siete casos de contracts/articulo-fecha.md §2; un bloque sin versiones y otro con una versión sin `fecha_vigencia`, construidos en el test, con fechas anteriores, intermedias y posteriores; los avisos de lo grabado sintético de `testdata/sintetico/avisos/` con `--fecha` |
| `TestArticuloEnUnaFechaYLaCache` | ídem | FR-020, FR-021, FR-022, FR-087, SC-007: la entrada de hoy no sirve la lectura fechada con `--offline` (4); tras una fechada, la entrada de hoy igual byte a byte y la sin fecha con `--offline` igual; la fechada con `--offline` y su entrada, el mismo `data` y la misma fecha de consulta |
| `TestApplySoloLoNuevo` | `internal/graph/almacen_test.go` | FR-030, FR-031: con `SoloLoNuevo`, un nodo, una arista y un texto guardados quedan iguales byte a byte; uno nuevo entra con la procedencia del lote; dos veces, nada la segunda |
| `TestIntegracionLecturaFechada` | `internal/graph/integracion_test.go` | FR-031, FR-086, SC-006: con fila de lecturas y sin ella (el mundo de H7), con `version-obsoleta` y con `fuente-caducada` presentes, la fila y `grafo.Comprobar` iguales tras el lote de una lectura fechada, y la siguiente lectura sin fecha con el `Comprobar` del mundo sin ella |
| `TestLecturaFechadaEnElGrafo` | `internal/app/grafo_test.go` | FR-030, FR-032, FR-086: con el applet y el almacén reales, la `BloqueVersion` del `2005-06-15` del `a22` con la procedencia del sobre; la fila de lecturas y `graph check` iguales; con `--no-graph` y `--dry-run`, `world.db` igual |
| `TestExtraerRedaccionesFechadas` | `internal/evals/fechadas_test.go` | FR-052: contracts/evals-redaccion-fechada.md §4 |
| `TestFormatoDeLasRedaccionesFechadas` | `internal/evals/formato_test.go` | FR-051, FR-052: ídem |
| `TestJuzgarLecturasFechadas` | `internal/evals/juzgar_test.go` | FR-051, FR-089 (caso 9): ídem |
| `TestUmbralesDeBoeLegislacion` | `internal/evals/informe_test.go` | FR-054, FR-055, FR-089 (casos 1 a 8), SC-010: ídem |

### Tests y guiones existentes que cambian

- **Por el grafo**: `TestEntregaDelMontador` (`internal/cli/entrega_test.go`), que el lote lleva `SoloLoNuevo`;
  `TestConsolidar` (`internal/core/grafo/lote_test.go`), que el consolidado lo lleva; `TestObservadoDeBoe`
  (`internal/source/boe/grafo_test.go`), un caso de la lectura fechada —el nodo y el texto, `SoloLoNuevo`, nada sin ELI
  ni bajo ensayo— que no cambia los de hoy; `TestLaSalidaDeBoeNoCambiaConElGrafo` (`internal/app/grafo_test.go`), con
  la lectura fechada.
- **Por los fallos de la fuente con fecha** (FR-007, FR-082): `TestFallosNoSeGuardan`
  (`internal/source/boe/fuente_test.go`) gana tres casos con `ConsultaArticulo{Norma: normaVigente, Bloque:
  bloqueDelArticulo21, Fecha}` el `2019-06-01`, sobre lo sintético de H4 y sin fichero nuevo: `bloque-ilegible` (4),
  `metadatos-caidos` —el bloque obtenido y los metadatos con el 503 de la fuente caída— (4) y el 429 de los metadatos
  tras obtener el bloque, `salvoEn(metadatosVigente, reproduce(sinteticoDelLimite))` (5), como la fila 18 de
  `TestClasesDeErrorDeBoe`. Cada uno resuelve además su consulta sin `Fecha` en un banco nuevo con el mismo Pedidor y
  exige la misma clase (`cli.Clasificar`), el mismo código (`cli.CodigoSalida`) y el mismo mensaje; y, como los casos
  de hoy, las dos veces el sobre sin datos con la dirección y el instante de la petición que falla, las mismas
  peticiones repetidas —nada guardado: ni los metadatos ni la entrada de las versiones— y, con `--offline`, el 4 de la
  entrada de las versiones ausente (`falloSinEntradaDeLasVersiones`, con el verbo `versiones` y su clave). El 503 y el
  429 van, como en H4, por la reproducción de `httpx` y sobre los metadatos: es el único momento en que hay algo
  obtenido que una escritura prematura de la entrada guardaría; un fallo de la primera petición no deja nada que
  escribir. Los cuatro casos de hoy, iguales.
- **Por el lector**: `TestLeerBloque` (`internal/source/boe/bloque_test.go`), con las versiones de los tres bloques de
  research M1 y la lista vacía de un bloque sin versiones; lo que esperaba hoy, igual.
- **Por la salida**: `TestEsquemasPublicados` (`make schema-check`) con `schemas/bloque.json` regenerado;
  `TestSalidaDeBoeContraSchemas` (`internal/app/boe_test.go`), con salidas fechadas; `TestSkillsDelRepositorio`, con la
  tabla regenerada. `TestHerramientasDelServidor` no cambia lo que espera y comprueba la herramienta con `fecha`.
  `TestGolden`, sus trece casos, igual (FR-012).
- **Por las evals y el job**: `TestConjuntoDeEvals` (`internal/evals/conjunto_test.go`), con el máximo de 23 y una eval
  con fecha fuera de las positivas; `TestEvalsDelRepositorio`, con las veintitrés evals, su grabado y el subtest nuevo
  `formas-de-la-redaccion-fechada`; `TestConsultasNecesarias` (`consultas_test.go`), con la lectura de cada bloque con
  `--fecha 9999-12-31`; `TestUmbralesDeLasTresSkills` (`informe_test.go`), con los quince de `boe-legislacion`;
  `TestDefinicionDelJob` (`definicion_test.go`), solo el literal `peorCasoDeBoeLegislacion` y el caso de
  `probarElPeorCasoDelTrabajo` que lo lleva (research D16); `TestPlanEnDosModos` (`plan_test.go`, subtest
  `del-repositorio`), solo las series y sesiones de `boe-legislacion` en `probarElPlanDelRepositorio` —de 32, 32 y 2 y
  96, 96 y 6 a 36, 36 y 2 y 108, 108 y 6— y su comentario, con los modelos y las repeticiones del job, iguales.

Ninguno se desactiva ni se salta, y ninguno pierde lo que comprobaba: cada cambio es el dato nuevo de una lista, una
cifra o un caso más. Ningún guion de `internal/app/testdata/script/` de hoy cambia.

### Puntos de entrada fuera de `make ci`

`TestEjecucionDelJob`, `TestMedidaDelJuez`, `TestSondeo` y `TestGrabar*` no cambian, y ninguna tarea los ejecuta
(FR-072).

### Fixtures, `testdata/` y `schemas/` (tareas `[datos]`)

- `schemas/bloque.json`, regenerado desde `--describe` con `-actualizar-esquemas`.
- `schemas/eval.yaml.json`: `comando-bloque-esperado` y `redaccion_fechada`.
- Ningún fichero nuevo ni cambiado en `testdata/`: las respuestas sin versiones o con una sin `fecha_vigencia` de
  FR-004 se construyen en el test, y lo demás sale de lo grabado de H4 y de sus sintéticos `avisos/`,
  `bloque-ilegible/`, `metadatos-caidos/` y `limite/` de `internal/source/boe/testdata/sintetico/`.

### CI

Ningún fichero de `.github/` cambia (FR-059). El job de evals lo lanza el workflow sobre la propuesta de cambio, tras
la revisión final (FR-091).

## Controles de umbral

Una fila por cada requisito o criterio con un umbral que se mide sin persona. En el cierre, los dos umbrales nuevos y
los que ya decidían en `boe-legislacion`, `jurisprudencia` y `legal-core` (FR-088); en `make ci`, cada control del
hito. `red` vacía, ninguna sesión sin medir y ninguna respuesta sin juzgar no son elementos de `umbrales`: los hace
cumplir el veredicto del job, como desde H7.3 y H24. La serie de 2 de 3 no tiene fila: no es un umbral agregado.

| Requisito | Umbral | Control | Dónde |
|---|---|---|---|
| FR-054, FR-055, SC-001 (modo orden) | 0 respuestas sin la lectura fechada o sin las líneas esperadas, de las 6 del modo en las evals con fecha | umbral del informe con `decide: true`: con una, su motivo con la sesión, veredicto `fallo` y el trabajo de `boe-legislacion` en rojo, que el cierre cuenta | `evals:boe-legislacion:redaccion_sin_fechar:claude-sonnet-5-5:orden` |
| FR-054, FR-055, SC-001 (modo herramienta) | ídem | ídem | `evals:boe-legislacion:redaccion_sin_fechar:claude-sonnet-5-5:herramienta` |
| FR-047, SC-001 (modo orden) | 0 respuestas sin la skill activada, de 60 | el de hoy, que sigue | `evals:boe-legislacion:sin_activar:claude-sonnet-5-5:orden` |
| FR-047, SC-001 (modo herramienta) | ídem | ídem | `evals:boe-legislacion:sin_activar:claude-sonnet-5-5:herramienta` |
| FR-058, SC-001 | como mucho 0,028 de las 120 respuestas juzgadas de los dos modos: 3 | el de hoy, que sigue | `evals:boe-legislacion:afirma_lo_no_leido:claude-sonnet-5-5` |
| FR-058, SC-001 (defectos) | 0 defectos sin marcar | el de hoy, con la medida versionada | `evals:boe-legislacion:medida_del_juez:afirma_lo_no_leido:defectos_sin_marcar` |
| FR-058, SC-001 (correctos) | 0 correctos marcados | ídem | `evals:boe-legislacion:medida_del_juez:afirma_lo_no_leido:correctos_marcados` |
| FR-059, SC-001 (sesiones, modo orden) | como mucho 900 s por tanda | el de hoy | `evals:boe-legislacion:duracion_de_las_sesiones:orden` |
| FR-059, SC-001 (sesiones, modo herramienta) | ídem | ídem | `evals:boe-legislacion:duracion_de_las_sesiones:herramienta` |
| FR-059, SC-001 (juez, modo orden) | como mucho 900 s de votos | ídem | `evals:boe-legislacion:duracion_del_juez:orden` |
| FR-059, SC-001 (juez, modo herramienta) | ídem | ídem | `evals:boe-legislacion:duracion_del_juez:herramienta` |
| FR-071, FR-088 (`jurisprudencia` sigue como decidía) | sus once umbrales que deciden, cumplidos; si no, queda en el informe para la persona | los de hoy, que no cambian | `evals:jurisprudencia:sin_activar:claude-sonnet-5-5:orden`, `evals:jurisprudencia:sin_activar:claude-sonnet-5-5:herramienta`, `evals:jurisprudencia:cita_sin_documento:claude-sonnet-5-5:orden`, `evals:jurisprudencia:cita_sin_documento:claude-sonnet-5-5:herramienta`, `evals:jurisprudencia:documento_sin_cotejar:claude-sonnet-5-5:orden`, `evals:jurisprudencia:documento_sin_cotejar:claude-sonnet-5-5:herramienta`, `evals:jurisprudencia:afirma_lo_no_leido:claude-sonnet-5-5`, `evals:jurisprudencia:medida_del_juez:afirma_lo_no_leido:defectos_sin_marcar`, `evals:jurisprudencia:medida_del_juez:afirma_lo_no_leido:correctos_marcados`, `evals:jurisprudencia:duracion_del_juez:orden`, `evals:jurisprudencia:duracion_del_juez:herramienta` |
| FR-071, FR-088 (`legal-core` sigue como decidía) | sus seis umbrales, cumplidos; si no, queda en el informe para la persona | los de hoy, que no cambian | `evals:legal-core:sin_activar:claude-sonnet-5-5:orden`, `evals:legal-core:sin_activar:claude-sonnet-5-5:herramienta`, `evals:legal-core:plazo_sin_fecha:claude-sonnet-5-5:orden`, `evals:legal-core:plazo_sin_fecha:claude-sonnet-5-5:herramienta`, `evals:legal-core:cobertura_sin_declarar:claude-sonnet-5-5:orden`, `evals:legal-core:cobertura_sin_declarar:claude-sonnet-5-5:herramienta` |
| FR-002, FR-003, FR-013, FR-081, SC-004 | 7 casos con su código, su versión, sus fechas, su norma, su huella y su redacción posterior: 0 diferencias | la tabla escrita a mano contra lo grabado | `ci:internal/source/boe/fecha_test.go:TestArticuloEnUnaFecha`, `ci:internal/app/testdata/script/h20-fecha-version.txtar` |
| FR-004, FR-011 | 0 textos devueltos de un bloque no situable, con fechas anteriores, intermedias y posteriores; los avisos de hoy con `--fecha` | los casos construidos en el test y lo grabado sintético de avisos | `ci:internal/source/boe/fecha_test.go:TestArticuloEnUnaFecha` |
| FR-005, FR-007, FR-082 | cada caso con su código y su mensaje; con la fuente caída, con el límite de ritmo y con el bloque o los metadatos ilegibles o caídos, 0 diferencias de clase, código y mensaje con la lectura sin `--fecha` y 0 entradas de las versiones escritas (la fechada con `--offline` después, 4); 0 cachés creadas con una fecha mal escrita | la tabla, los fallos de la fuente con y sin fecha y el guion | `ci:internal/source/boe/ids_test.go:TestValidarFecha`, `ci:internal/source/boe/fuente_test.go:TestFallosNoSeGuardan`, `ci:internal/app/testdata/script/h20-fecha-codigos.txtar` |
| FR-012, FR-083, SC-005 | 0 bytes de diferencia con `main` sin `--fecha`, con `--json` y sin él, en éxito y en fallo; los trece golden de hoy, iguales | el guion con la salida de `main` dentro; los golden | `ci:internal/app/testdata/script/h20-fecha-main.txtar`, `ci:internal/source/boe/fuente_test.go:TestGolden` |
| FR-020, FR-021, FR-022, FR-082, FR-087, SC-007 | la entrada de hoy sirve la lectura fechada 0 veces y cambia en 0 bytes; con `--offline` y la entrada de las versiones, 0 diferencias de `data` y `hash` | los tests de la caché y el guion | `ci:internal/source/boe/fecha_test.go:TestArticuloEnUnaFechaYLaCache`, `ci:internal/app/testdata/script/h20-fecha-offline.txtar` |
| FR-030, FR-031, FR-032, FR-086, SC-006 | 0 elementos cambiados en la fila de lecturas y en `graph check`; 0 entregas con `--no-graph` o `--dry-run`; 0 cambios en la segunda entrega | los tests del grafo y el guion | `ci:internal/graph/almacen_test.go:TestApplySoloLoNuevo`, `ci:internal/graph/integracion_test.go:TestIntegracionLecturaFechada`, `ci:internal/app/grafo_test.go:TestLecturaFechadaEnElGrafo`, `ci:internal/app/testdata/script/h20-fecha-grafo.txtar` |
| FR-001, FR-084 | `boe_articulo` con `fecha`: el sobre de la orden; con una fecha anterior, el error `no-encontrado` | el guion y el control de conformidad de H21 | `ci:internal/app/testdata/script/h20-fecha-herramienta.txtar`, `ci:internal/app/herramientas_test.go:TestHerramientasDelServidor` |
| FR-014, FR-085 | 0 salidas fechadas que no validan; 0 diferencias entre `schemas/` y `--describe` | la validación y el golden | `ci:Makefile:schema-check`, `ci:internal/app/boe_test.go:TestSalidaDeBoeContraSchemas` |
| FR-048, SC-011 | `SKILL.md` de `boe-legislacion` < 300 líneas; su tabla y sus `references/`, las regeneradas; las tres etiquetas en sus bloques `text` | la comprobación de skills | `ci:Makefile:skills-check`, `ci:internal/evals/conjunto_test.go:TestEvalsDelRepositorio` |
| FR-050 | el tope de 352 minutos ≥ el peor caso de `boe-legislacion` con las evals del repositorio, 18 977 s | la definición del job contra las evals del repositorio | `ci:internal/evals/definicion_test.go:TestDefinicionDelJob` |
| FR-051, FR-052 | las 23 evals válidas contra el esquema y con sus reglas; cada fila de la tabla de qué casa con un comando de bloque | el esquema, las reglas y el juicio | `ci:internal/evals/formato_test.go:TestFormatoDeLasRedaccionesFechadas`, `ci:internal/evals/juzgar_test.go:TestJuzgarLecturasFechadas`, `ci:internal/evals/conjunto_test.go:TestEvalsDelRepositorio` |
| FR-053 | 0 consultas necesarias, con fecha o sin ella, sin servir desde lo grabado | `Preparar` y `ComprobarSinRed` sobre la unión de grabaciones | `ci:internal/evals/conjunto_test.go:TestEvalsDelRepositorio` |
| FR-055, FR-089, SC-010 | 1 sesión que cuenta → `fallo` y su motivo; los nueve casos de sesiones sintéticas y el de los elementos: 10 de 10 | las sesiones sintéticas, con los quince umbrales, los motivos y el veredicto escritos a mano | `ci:internal/evals/informe_test.go:TestUmbralesDeBoeLegislacion`, `ci:internal/evals/juzgar_test.go:TestJuzgarLecturasFechadas`, `ci:internal/evals/informe_test.go:TestUmbralesDeLasTresSkills` |
| FR-056, SC-009 | 15 elementos con 11 que deciden en `boe-legislacion`, 6 con 6 en `legal-core` y 15 con 11 en `jurisprudencia`, en su orden | el informe con las evals del repositorio y el plan del job | `ci:internal/evals/informe_test.go:TestUmbralesDeLasTresSkills` |
| FR-090, SC-008 | con cada uno de los 5 mutantes, al menos 1 test falla: 5 de 5 | los tests de contracts/evals-redaccion-fechada.md §5; la tarea de FR-090 los ve fallar y lo deja en `mutantes.md` | `ci:internal/source/boe/fecha_test.go:TestArticuloEnUnaFecha`, `ci:internal/source/boe/grafo_test.go:TestObservadoDeBoe`, `ci:internal/app/grafo_test.go:TestLecturaFechadaEnElGrafo`, `ci:internal/evals/informe_test.go:TestUmbralesDeBoeLegislacion`, `ci:internal/evals/informe_test.go:TestUmbralesDeLasTresSkills` |
| FR-080 | `make ci` en verde | el objetivo entero | `ci:Makefile:ci` |

Sin fila: SC-002, que decide el informe final con las filas de arriba; SC-003, que mide una persona; SC-012, la
revisión final por lectura (FR-062); y FR-040 a FR-046 y FR-049, que no tienen umbral propio: su efecto lo miden las
dos primeras filas y la serie de cada eval, y el texto lo leen los jueces de la revisión final.

**El recorrido del umbral nuevo.** Con una sesión de la eval 22 del modo orden que lee el `a1-30` sin `--fecha`:
`medirLasRedaccionesFechadas` la marca; `umbralDeRedaccionSinFechar` da 1 de 6, `cumple: false` y `decide: true`; su
motivo nombra la sesión; el veredicto es `fallo`; el trabajo `evals (boe-legislacion)` sale en rojo; `medir_cierre` lo
cuenta y guarda el informe; y el informe final busca en `umbrales` el nombre exacto de la fila y lo da como no
cumplido. Con todas leyendo en su fecha y con sus líneas, todo en verde, y «comprobado por su control». En `make ci`,
el mismo recorrido lo hace `TestUmbralesDeBoeLegislacion` con sesiones sintéticas, y el mutante del `decide` prueba
que falla si deja de decidir.

## Uso, de fuera adentro (criterio de uso, ADR 0028)

El detalle, con los ejemplos y sus bytes, está en contracts/articulo-fecha.md §4 y §7, contracts/evals-redaccion-fechada.md
§6 y contracts/skill-boe-legislacion.md §4. Nada de lo que lee una skill crece con lo consultado en meses —cientos de
normas y miles de bloques—: cada lectura da un bloque, el que pide la pregunta.

- **El sobre de la lectura fechada** (`boe-legislacion`, una vez por bloque de la respuesta, dos con un año o un mes
  sin día, y una lectura sin fecha más solo si la respuesta dice qué dice hoy): el del `a1-30` el `2019-06-01`, 2 226
  bytes de `data` medidos, 119 de ellos los dos campos nuevos, y unos 2,5 KB con el sobre (research M2). La redacción
  posterior deja de darse con una fecha desde la última vigencia del bloque y cambia cuando el BOE añade una versión.
- **El 3 de FR-003** (la misma lectura, para la línea C): un mensaje de unos 200 bytes; solo con una fecha anterior a
  la primera vigencia.
- **Las líneas A, B y C** (la persona, una vez por respuesta y bloque, dos con un año o un mes; el job, sin modelo): una
  línea cada una; solo en respuestas con fecha.
- **La entrada de las versiones** (solo el binario, por su clave): una por bloque leído con fecha, del tamaño de la
  respuesta del BOE (3,8 KB el `a1-30`, 37,6 KB el `a22`, research M3); sin purga, como las de hoy (research V8);
  ninguna skill la lee y no da señales.
- **La `BloqueVersion` y su texto** (ninguna skill: no es una lectura): uno por versión distinta leída con fecha; no
  es la redacción vista de ningún bloque, así que nunca da `version-obsoleta` ni `fuente-caducada`.
- **Los dos umbrales nuevos y su motivo** (el job, una vez; el informe final; la persona): unos 450 bytes cada uno;
  el motivo, menos de 0,5 KB con las 6 sesiones del modo; dejan de dar `cumple: false` en el primer job en que ninguna
  sesión cuenta.
- **`SKILL.md`** (el agente que activa la skill, una vez por conversación): 298 líneas en el prototipo; no da señales.

## Decisiones

- **La lectura fechada vive en el adaptador** (D1), con `--fecha` como `*string` (D2) validado con
  `time.Parse(time.DateOnly, …)` antes de abrir nada (D3).
- **La regla de elección**: la última versión cuya vigencia no es posterior a la fecha; 3 antes de la primera; 3 con
  otro mensaje si alguna versión no tiene `fecha_vigencia` o no hay versiones (D4, D7).
- **La salida**, `ArticuloFechado`, que embebe `Articulo` y añade `fecha_pedida` y `redaccion_posterior`, solo en la
  lectura fechada (D5, D6).
- **La caché**, una entrada nueva con las versiones del bloque, que sirve cualquier fecha (D8).
- **El grafo**, solo la `BloqueVersion` y su texto, con `SoloLoNuevo` (D9, D10).
- **Las evals**: `fecha` en el comando de bloque y `redaccion_fechada` (D11); qué casa con cada comando (D12); tres
  formas fijas con su cita (D13); el umbral `redaccion_sin_fechar` (D14); las evals con fecha aparte en las reglas del
  conjunto (D15); el literal del peor caso del job (D16); la caché preparada con `--fecha 9999-12-31` (D18).
- **La skill** sin `graph check` tras una lectura fechada (D17); **los golden** (D19); **la aceptación** (D20); **los
  documentos** (D21); **ningún dato externo** (D22).

## Trazabilidad: cada mecanismo y su requisito

| Mecanismo | Requisito |
|---|---|
| `argumentosDeArticulo.Fecha` y `ConsultaArticulo.Fecha` (`*string`) | FR-001, FR-005 (la fecha vacía) |
| `ValidarFecha`, antes de abrir nada | FR-005 |
| `bloqueLeido.versiones` y `versionLeida` | FR-002, FR-010 |
| La elección y sus dos mensajes de 3 (`fecha.go`) | FR-002, FR-003, FR-004, FR-013 |
| `ArticuloFechado`, `RedaccionPosterior` y la `Salida` de `articulo` | FR-010, FR-012, FR-014, FR-085 |
| `versionesDelBloque`, `claveDeLasVersiones` y su lectura y escritura, solo con el bloque y sus metadatos obtenidos e interpretados | FR-007 (un fallo no deja nada escrito, como hoy), FR-020, FR-021, FR-022 |
| `nodoDeLaVersion` y `observadoDeLaLecturaFechada` | FR-030, FR-031 |
| `Observado.SoloLoNuevo`, `Lote.SoloLoNuevo`, `Consolidado.SoloLoNuevo`, su copia en `LoteDe` y `Consolidar`, y la entrega que no toca lo guardado | FR-030, FR-031 (el nodo ya guardado: research D9) |
| El párrafo de `doc.go` | FR-001: el comentario de paquete, que dice qué hace cada verbo frente a `refs/boe.py`, queda incompleto sin la lectura fechada |
| `schemas/bloque.json` regenerado | FR-014, FR-085 |
| `ComandoEsperado.Fecha`, `comando-bloque-esperado` y la lectura de `--fecha` en el juicio | FR-051 |
| `RedaccionFechadaEsperada`, `Eval.RedaccionFechada` y su esquema | FR-052 |
| `fechadas.go`: las tres etiquetas y `ExtraerRedaccionesFechadas` | FR-041, FR-043, FR-048, FR-052, FR-054 |
| `medirLasRedaccionesFechadas`, `respuestaDelJuez.conFecha` y `.sinFechar`, `respuestasMedidas.conRedaccionFechada` | FR-054, FR-055 |
| `umbralDeRedaccionSinFechar`, su nombre, su descripción, su sitio y su caso en `nombradasPorSuSesion` | FR-054, FR-055, FR-056 |
| La lectura de cada bloque con `--fecha 9999-12-31` en `ConsultasNecesarias` | FR-053 |
| Las evals con fecha aparte y `maximoDeEvals` 23 | FR-050, FR-052, FR-080 |
| Las evals 22 y 23 | FR-050, FR-057 |
| El literal del peor caso en `definicion_test.go` | FR-050 |
| Las series y sesiones de `boe-legislacion` en `plan_test.go` | FR-050, FR-059 |
| El subtest `formas-de-la-redaccion-fechada` | FR-048 (las formas tal cual, para el guion) |
| `SKILL.md` | FR-040 a FR-049 |
| Los documentos | FR-060 a FR-062 |
| Los seis guiones | FR-001 a FR-007, FR-010, FR-012, FR-013, FR-020 a FR-022, FR-030 a FR-032, FR-070, FR-082 a FR-084 |
| Los tests nuevos y los que cambian | sus filas de «Controles de umbral» |
| Los mutantes y `mutantes.md` | FR-090, SC-008 |

Nada del diseño atiende a un estado por debajo del umbral de materialidad: una `fecha_vigencia` ilegible del BOE cae
en la misma condición que una ausente (FR-004), sin caso propio ni test; dos versiones con la misma fecha de vigencia
no tienen regla aparte, porque la elección recorre en orden; y la entrada de las versiones manipulada a mano es la regla
genérica, defecto `inesperado` y código 1, como toda entrada ilegible hoy.

## Datos externos

Ninguno (research D22). Lo grabado de H4 (`internal/source/boe/testdata/boe.legislacion-consolidada/`) tiene los tres
bloques con todas sus versiones y sus metadatos, y la caché de las evals lo une con lo de H5 y H9
(`UnionDeGrabaciones`). La única fuente es `boe.legislacion-consolidada`, con fila revisada en `docs/SOURCES.md` de
`main` (research V37), y la lectura fechada pide sus mismos recursos (FR-006). Ningún manifiesto `grabaciones.json`,
test `TestGrabar*`, evidencia ni fichero de `data/` cambia, y el paso `grabar_datos` no tiene nada que grabar.

## Orden de implementación (de dentro afuera)

1. **La aceptación**: los seis guiones en `specs/024-h20-boe-articulo-fecha/aceptacion/`, con la salida de `main` de
   `fecha-main.txtar` escrita desde el binario del arnés de hoy.
2. **El grafo**: `SoloLoNuevo` en `schema.Observado`, `core.Lote`, `grafo.Consolidado` y `cli.LoteDe`, y la entrega
   que no toca lo guardado, con `TestApplySoloLoNuevo`, `TestIntegracionLecturaFechada` y los tests de la copia.
3. **El lector**: `bloqueLeido.versiones`, con `TestLeerBloque`, y los trece golden de hoy sin cambiar.
4. **La lectura fechada en el adaptador**: `ValidarFecha`, `ConsultaArticulo.Fecha`, `fecha.go`, `datos.go`,
   `entradas.go`, `grafo.go`, la rama de `articulo.go` y el párrafo de `doc.go`, con `TestValidarFecha`,
   `TestArticuloEnUnaFecha`, `TestArticuloEnUnaFechaYLaCache`, los tres casos con fecha de `TestFallosNoSeGuardan` y
   el caso de `TestObservadoDeBoe`.
5. **La orden y la herramienta**: `argumentosDeArticulo.Fecha` y la `Salida` de `articulo`; `schemas/bloque.json`
   regenerado (`[datos]`); la tabla de `boe-legislacion` regenerada con `make skills-sync` (solo su región generada);
   `TestLecturaFechadaEnElGrafo` y los cambios de `TestLaSalidaDeBoeNoCambiaConElGrafo` y
   `TestSalidaDeBoeContraSchemas`.
6. **El formato de las evals**: `ComandoEsperado.Fecha`, `RedaccionFechadaEsperada`, `schemas/eval.yaml.json`
   (`[datos]`), el juicio de un comando de bloque y `fechadas.go`, con `TestFormatoDeLasRedaccionesFechadas`,
   `TestJuzgarLecturasFechadas` y `TestExtraerRedaccionesFechadas`.
7. **Las evals y el umbral**: las evals 22 y 23, las reglas del conjunto, `ConsultasNecesarias`, la medida y el umbral,
   el literal del peor caso y los comentarios de `internal/evals` que dejan de ser ciertos, con
   `TestUmbralesDeBoeLegislacion` y los cambios de `TestUmbralesDeLasTresSkills`, `TestConjuntoDeEvals`,
   `TestEvalsDelRepositorio`, `TestConsultasNecesarias`, `TestDefinicionDelJob` y `TestPlanEnDosModos`. Una sola tarea: añadir las evals
   cambia a la vez el conjunto, la caché preparada, el peor caso y los umbrales, y los tests que los fijan tienen que
   cambiar con ellas para dejar `make ci` en verde.
8. **La skill**: `skills/boe-legislacion/SKILL.md` desde contracts/skill-boe-legislacion.md y su prototipo, con el
   subtest `formas-de-la-redaccion-fechada`.
9. **Los mutantes**: el escenario 7 del quickstart y `specs/024-h20-boe-articulo-fecha/mutantes.md`.
10. **Los documentos**: con contracts/documentos.md.

**Obligaciones para `tasks.md`**:

- una sola tarea `[aceptacion]`, la primera, que escribe solo los seis guiones: ninguna eval ni ningún dato;
- cada tarea deja `make ci` en verde; las que tocan `schemas/` llevan `[datos]`; ninguna toca `testdata/`;
- cada fila `ci:` de «Controles de umbral» tiene la tarea que construye su control con un caso que lo ve fallar;
- la tarea de la skill declara `skills/boe-legislacion/SKILL.md` e `internal/evals/conjunto_test.go`, y ningún otro
  fichero de `skills/`; la de la orden declara además `skills/boe-legislacion/SKILL.md` solo por su tabla generada;
- la tarea de los mutantes declara `specs/024-h20-boe-articulo-fecha/mutantes.md`, no cambia nada bajo `internal/` y
  termina con el árbol aparte retirado;
- la de los documentos declara `CHANGELOG.md`, `README.md`, `CONTRIBUTING.md` y `docs/WORKFLOW.md`, y en este solo
  cambia la frase de `medir_cierre` que cuenta los umbrales que deciden en `boe-legislacion` (contracts/documentos.md);
- ninguna tarea ni corrección cumple un umbral cambiando su valor, dejándolo en `decide: false`, sacando sesiones de su
  `total`, ni cambiando las evals, la rúbrica, los casos o la medida del juez (FR-071);
- ninguna toca `skills/legal-core/`, `skills/jurisprudencia/`, `evals/legal-core/`, `evals/jurisprudencia/`,
  `evals/boe-legislacion/juez/`, `evals/boe-legislacion/expresiones-prohibidas.yaml`, las veintiuna evals de hoy,
  `.github/`, `scripts/`, `.specify/`, `evidencias/`, `docs/SOURCES.md`, `docs/ROADMAP.md`, `docs/USO.md`,
  `docs/HISTORIAL.md`, los ADR, la web, `CLAUDE.md` ni `specs/001-…` a `specs/023-…` (FR-063, FR-070);
- ninguna usa la red, abre una sesión con modelo ni ejecuta `make evals`, `make evals-sondeo`, `make evals-medir-juez`,
  `scripts/evals*.sh`, `TestEjecucionDelJob`, `TestMedidaDelJuez`, `TestSondeo` ni `TestGrabar*` (FR-072);
- ninguna publica ni mide en la plataforma: los escenarios 9 y 10 del quickstart no son tareas.

## Complexity Tracking

Ninguna desviación que justificar: ni dependencias fuera de la lista, ni paquetes nuevos, ni reglas de dependencia
rotas, ni excepciones a los principios.

## Comprobación contra la rúbrica del juez (`juez_plan`) y `precheck.sh plan`

- `precheck.sh plan`: existen `plan.md` y `research.md`; ninguna marca de aclaración pendiente; están
  `## Constitution Check`, la línea «Aceptación e2e» y `## Controles de umbral`, con cada fila nombrando requisitos del
  spec y un control con la forma `ci:` o `evals:`. El guion pide aprobación en el sandbox de esta sesión, así que sus
  comprobaciones se hicieron con su misma lógica (`controles_de_umbral`, `ids_de_linea` y `controles_de_celda` de
  `scripts/workflow/comun.sh`) en `awk` y `grep`: 29 filas, todas con requisitos definidos en el spec y con control; 0
  defectos.
- a · Constitution Check: un ítem por principio (I-IX) y por regla de dependencia (R1-R6).
- b · Dependencias: ninguna nueva (research V2).
- c · Reglas de dependencia: tabla R1-R6; ningún import nuevo.
- d · Errores y códigos: los dos 3 nuevos con `errorDeNoEncontrado`, la fecha con `errorDeArgumentos`; la fila de
  FR-004, decidida y razonada (research D7).
- e · Tests primero: los seis guiones con su historia y sus requisitos; los tests nuevos y los que cambian, entre ellos
  los fallos de la fuente con fecha de FR-007 en `TestFallosNoSeGuardan`; ningún fixture nuevo; las tareas `[datos]`.
- f · Alcance: nada fuera del spec; `articulos`, `analisis`, el diff entre versiones, la derogación en la fecha, otras
  skills y el workflow del job, fuera.
- g · Sin atajos: ningún `//nolint`, `t.Skip`, TODO ni error silenciado previstos.
- h · Mejor alternativa: cada decisión con la rechazada (research D1-D22).
- i · Afirmaciones verificadas: V1-V39 con fichero y línea u orden; M1-M11 medidos en esta sesión; S1-S3 como supuestos.
- j · Quickstart ejecutable: órdenes, rutas y nombres reales; las de §1 a §6 ejecutadas hoy en esta sesión, todas con
  `ok` y el árbol de trabajo sin cambios; §7 dice el efecto de su árbol aparte y cómo se retira (la mecánica de
  `git worktree add --detach` y `remove --force` se usó en esta sesión para el prototipo); §8, `make ci`, vale tras el
  hito.
- k · Datos externos: ninguno; la única fuente, con fila revisada.
- l · Autonomía: ninguna tarea para una persona; el cierre en la plataforma lo hace el workflow.
- m · Uso: «Uso, de fuera adentro» y los contratos, con bytes medidos, veces por pregunta y lo que apaga cada señal.
- n · Proporcionalidad: «Trazabilidad»; nada para estados por debajo del umbral de materialidad.
- o · Controles de umbral: los dos nuevos y los veintiséis que ya decidían en el cierre, y una fila por cada umbral
  medido en `make ci`.
