# Implementation Plan: H9 · `internal/core/plazos` + applet `plazos` + skill `legal-core` v1

**Branch**: `023-h9-internal-core-plazos` | **Date**: 2026-10-18 | **Spec**: [spec.md](./spec.md)

**Input**: Feature specification from `/specs/023-h9-internal-core-plazos/spec.md`

**Modo**: desatendido. Las decisiones técnicas se tomaron con el «Criterio de decisión autónoma» de la constitución y
están en [research.md](./research.md) (D1-D30), cada una con su alternativa rechazada. Lo comprobado en esta sesión
—el repositorio, las evidencias del ADR 0039, dos sondas sobre el BOCM y el fichero CKAN, y un prototipo del applet en
una copia de trabajo bajo `$TMPDIR`— está en la tabla V; lo medido, en M; y lo que no se pudo comprobar sin red ni
modelo —el formato de la resolución del BOE, los festivos nacionales y autonómicos, el texto vigente de los preceptos y
lo que harán las sesiones—, en los supuestos S1-S7.

## Summary

De fuera adentro: quien pregunta hasta cuándo puede recurrir recibe la fecha calculada por el binario, con el precepto
leído y la razón de cada día que la movió; y si el asunto lo regula la ley aparte, una declaración en lugar del plazo
del común. Lo entregan seis piezas:

1. **`legal-core` v1** (contracts/skill-legal-core.md; borrador de 248 líneas en
   contracts/SKILL-legal-core.prototipo.md): cinco pasos nuevos —cobertura antes de preguntar nada, situar, plazo,
   precepto, respuesta— y tres
   formas fijas, `⚠ PLAZO:`, `⚠ NO CUBIERTO:` y la línea de cada aviso; dos referencias generadas.
2. **El applet `plazos`** (contracts/applet-plazos.md), con `calcular` y su herramienta `plazos_calcular`: el art. 30
   LPAC con los dos municipios del art. 30.6, y el contencioso con agosto, el art. 182.1 LOPJ y el aviso del 183. Su
   dominio, puro, en `internal/core/plazos`. Sin red, sin caché, sin grafo.
3. **`data/festivos/`** (contracts/datos-de-plazos.md §1), derivado por código de lo que graba `grabar_datos`: los
   nacionales y autonómicos de las tres resoluciones del BOE y los locales de la Comunidad de Madrid de las siete
   disposiciones del BOCM, con las correcciones en su orden. Una sonda de esta sesión, con la regla de cruce y la
   gramática de research D13, da 179 de 179 municipios en 2025 y 2026 y exactamente las diez diferencias del
   `LEEME.md` con el fichero CKAN (research V9).
4. **`data/plazos.yaml` y `data/cobertura.yaml`** (§2 y §3), comprobados contra lo grabado del BOE por bloque, por
   duración y por palabras clave (§4).
5. **El formato de eval y cinco evals** (contracts/evals-legal-core.md): `plazo`, `no_cubierto` y el comando
   `calcular`; (a) a (e), con sus fechas a mano (contracts/casos-de-plazos.md).
6. **Dos umbrales que deciden**, `plazo_sin_fecha` y `cobertura_sin_declarar`, por modo, y `sin_activar` de
   `legal-core`, que pasa a decidir.

No hay ADR nuevo ni dependencia nueva. Hay dos adaptadores de fuente nuevos, sin consulta en ejecución, y tres
manifiestos de grabación.

## Technical Context

**Language/Version**: Go 1.27 (`go.mod`; toolchain `go1.27.2`), `CGO_ENABLED=0`, `-trimpath`; YAML de `data/` y de las
evals; Markdown de la skill.

**Primary Dependencies**: las fijadas, ninguna nueva: `alecthomas/kong` (la gramática del verbo), `invopop/jsonschema`
(`--describe`), `santhosh-tekuri/jsonschema` (salidas contra su esquema, evals y `data/cobertura.yaml` contra el suyo),
`go.yaml.in/yaml/v3` v3.0.5, la ruta con la que `go.mod` fija `yaml.v3` (`data/`), `rogpeppe/go-internal`
(`testscript`), `stretchr/testify`; biblioteca estándar (`time`,
`encoding/xml` y `html` o `strings` para los boletines, `slices`, `strconv`). Las property tests van sobre una rejilla
exhaustiva, sin `testing/quick` ni `pgregory.net/rapid` (research D22).

**Storage**: ninguno en ejecución. Ficheros nuevos empotrados: `data/festivos/nacionales-y-autonomicos.yaml`,
`data/festivos/locales/13.yaml` y `data/plazos.yaml`. Fuera del binario: `data/cobertura.yaml`, las grabaciones en
`evidencias/h9/` e `internal/evals/testdata/`, y los golden.

**Testing**: `go test -race` en `make ci`: tablas, golden contra su esquema, property tests, derivación byte a byte,
cotejo con el fichero CKAN, verificación de los preceptos contra lo grabado, cinco guiones `testscript`, el test de
arquitectura y sesiones sintéticas para el juicio y el informe. El job de evals, en la propuesta de cambio, tras la
revisión final.

**Target Platform**: el binario, macOS, Linux y Windows, como hoy; el job, `ubuntu-24.04`.

**Project Type**: skill y CLI multicall (un applet nuevo en el binario único).

**Performance Goals**: ninguno con umbral. Un cómputo es lineal en la duración pedida y no lee nada fuera de lo
empotrado.

**Constraints**: el código de `plazos` no alcanza `net`, `net/http` ni `internal/httpx` (FR-013); `SKILL.md` < 300
líneas; `.github/workflows/evals.yml`, `skills/boe-legislacion/`, `skills/jurisprudencia/`, sus evals, `internal/mcp`,
`docs/ROADMAP.md`, los ADR, `docs/USO.md`, `docs/HISTORIAL.md`, `web/` y los specs de otros hitos no cambian; ningún
`//nolint`; ningún test saltado.

**Scale/Scope**: 1 verbo, 1 herramienta (13 en producción, 16 con las de ejemplo), 7 tipos, unos 250 festivos
nacionales y autonómicos y unos 720 locales, 9 entradas de cobertura, 5 evals (de 42 a 102 sesiones del trabajo de
`legal-core`: 48, 48 y 6, más la prueba de red), 6 umbrales en `legal-core`, 5 guiones de aceptación, 20 casos en la
tabla; y lo que la entrega rompe sin tocarlo y hay que alinear: 6 tests de Go y 4 guiones (research V11, V15), más
los que el hito amplía a propósito («Tests y guiones existentes que cambian»).

## Constitution Check

*GATE: pasado antes de la fase 0 y re-evaluado tras el diseño (fase 1): sin violaciones; las desviaciones
justificadas están en Complexity Tracking.*

### Principios

| Principio | Cómo lo cumple H9 |
|---|---|
| I · Fuentes públicas y frontera humana | Las tres fuentes que se graban tienen fila revisada en `docs/SOURCES.md` de `main` (research V1): `boe.fiestas-laborales` a 1 s, `bocm.fiestas-locales` a 10 s y `boe.legislacion-consolidada` a 1 s, siempre por `internal/httpx`, GET, con `robots.txt` y User-Agent; solo en el paso `grabar_datos`. Ningún applet pide festivos en ejecución (FR-028). Ninguna acción con identidad: la skill dice hasta cuándo, no presenta nada (regla 6). |
| II · Nada sin cita ni fuente | El sobre de `plazos` lleva su procedencia del espacio reservado; cada festivo, en `data/festivos/` y en `data`, su `source`, la disposición que lo fija; cada tipo, su precepto, que la skill lee antes de citarlo. Lo que la ley regula aparte se declara no cubierto, con la norma que lo rige solo si el plan la fija y lo grabado la confirma (FR-031, FR-042; ADR 0027): es el principio que el hito aplica a los plazos, con un umbral que lo hace cumplir (FR-054). |
| III · Tests primero y offline | Cinco guiones `testscript` escritos antes que el código y congelados; unitarios y property tests offline; las grabaciones, solo del paso `grabar_datos`, y los tests las leen sin red; cada salida, contra `schemas/plazo.json`; `internal/core/plazos` con su cobertura ≥ 85 %. |
| IV · Hexagonal y errores tipados | Dominio puro en `internal/core/plazos`, que recibe bytes; adaptadores de los dos boletines en `internal/source/{boe/fiestas,bocm}`; composición en `internal/app`. Errores de clase `argumentos` (2) y `no-encontrado` (3), en el orden de FR-012; unas fuentes que no cargan, `inesperado` (1); nunca 4, 5 ni 6 (contracts/applet-plazos.md §6). Ningún `panic`. |
| V · Simplicidad y dependencias | Ninguna dependencia nueva. Cada mecanismo, con su requisito («Trazabilidad»). Rechazados: tres tipos genéricos, banderas por papel de municipio, una tabla de alias de nombres, el fichero CKAN como tabla de códigos, fragmentos de texto legal en `data/`, un esquema para `data/festivos/` y otro para `data/plazos.yaml`, `testing/quick` (research D2, D3, D13, D16, D17, D22, D25). |
| VI · Un binario, convenciones de agente | Un applet más en el multicall, con las ocho banderas globales heredadas; la herramienta y la fila de la tabla salen de `--describe` (research V13). |
| VII · Grafo y privacidad | `plazos` no emite operaciones de grafo ni abre `world.db` (FR-013); los municipios del asunto los da quien llama y no se guardan en ningún sitio. |
| VIII · Skills primero | El hito mejora `legal-core`, medida por cinco evals nuevas y dos umbrales; solo va al binario lo que exige determinismo, las fechas. `SKILL.md`, 248 líneas, sin `scripts/`, con cada operación como orden y como herramienta. |
| IX · Genericidad territorial | Ningún municipio con trato propio en código, `data/` ni la skill (FR-073): los festivos locales se configuran por comunidad (`data/festivos/locales/<comunidad>.yaml`) y el cruce de nombres es una regla para todos (research D13). Matriz territorial en e2e: Leganés, con festivos locales, y Tordesillas, sin ellos, con la cobertura en `data` (FR-085). Fuera de la Comunidad de Madrid, o en 2027, `cobertura` y un aviso dicen qué falta, nunca que no hay festivos (FR-026). |

### Reglas de dependencia (`docs/ROADMAP.md` §2, constitución IV)

| Regla | Cómo la cumple |
|---|---|
| R1 · `internal/core/**` no importa adaptadores | `internal/core/plazos` importa `internal/core/schema`, `go.yaml.in/yaml/v3` y la biblioteca estándar; ni `os` ni `io` (research V16). Recibe los municipios ya resueltos —código INE, nombre y comunidad—: los resuelve el applet con `territorio`. `internal/core/territorio` gana `Registro.MunicipiosDe`, que da los municipios de una comunidad con sus formas plegadas, para el cruce del BOCM (research D13); ningún paquete del dominio importa otro nuevo. Los tests del dominio usan calendarios sintéticos; los que necesitan los ficheros reales viven en `internal/app` (research D23). |
| R2 · Solo `internal/httpx` importa `net/http` | Ningún fichero nuevo lo importa. Los dos adaptadores piden con `httpx` en su `TestGrabar*` y leen con `httpx.Replay`. Más allá de R2: `internal/core/plazos` y `internal/app/plazos.go` no alcanzan `net` ni `internal/httpx` (`TestArquitectura`). |
| R3 · Solo `cache`/`store`/`graph` importan SQLite y `database/sql` | Nada nuevo los importa. |
| R4 · Solo `cli` y `cmd/` llaman a `os.Exit` | Ninguna llamada nueva. |
| R5 · Solo `render` escribe en stdout | El applet devuelve un `Resultado`; los tests de derivación escriben `data/festivos/` con `-actualizar-festivos`, como los golden. |
| R6 · `internal/graph` no importa `source/*` ni `render` | Sin cambios. |
| R7 (`internal/arch_test.go`) · Solo `internal/mcp` importa el SDK | Sin cambios: la herramienta sale del registro. |

## Project Structure

### Documentation (this feature)

```text
specs/023-h9-internal-core-plazos/
├── plan.md                         # este fichero
├── research.md                     # V1-V29, M1-M5, S1-S8, D1-D30
├── data-model.md                   # calendario, tipos, petición, plazo, derivación, evals, umbrales
├── quickstart.md                   # §0-§13; §14, el cierre; §15, la persona; §16, limpieza
├── contracts/
│   ├── applet-plazos.md            # verbo, cómputo, sobre, data, avisos, errores, --describe, uso
│   ├── datos-de-plazos.md          # data/festivos/, data/plazos.yaml, data/cobertura.yaml, grabaciones, verificación
│   ├── casos-de-plazos.md          # la tabla de casos, con sus razones
│   ├── evals-legal-core.md         # formato, formas fijas, juicio, umbrales, reglas del conjunto, las cinco evals
│   ├── skill-legal-core.md         # qué lleva SKILL.md y por qué, referencias, uso
│   └── SKILL-legal-core.prototipo.md   # el borrador de SKILL.md, con su tabla generada
├── checklists/                     # del spec
├── gates/supuestos.md              # con los supuestos del plan
├── mutantes.md                     # lo escribe la tarea de FR-089
└── tasks.md                        # la escribe /speckit-tasks
```

### Source Code (repository root)

```text
skills/legal-core/
├── SKILL.md                                  # v1, desde el borrador
└── references/recursos_y_plazos.md, cobertura_por_materia.md   # generadas
evals/legal-core/05-…09-*.yaml                # las cinco evals nuevas
data/
├── datos.go                                  # empotra festivos/ y plazos.yaml
├── plazos.yaml, cobertura.yaml               # nuevos
└── festivos/nacionales-y-autonomicos.yaml, festivos/locales/13.yaml   # generados
internal/core/plazos/                         # nuevo paquete, dominio puro
├── doc.go, tipos.go, calendario.go, festivos.go       # LeerTipos, CargarCalendario, EscribirFestivos…
├── calcular.go, computo.go, avisos.go, errores.go, salida.go
internal/core/territorio/registro.go          # Registro.MunicipiosDe: los municipios de una comunidad con sus formas
internal/source/boe/fiestas/                  # nuevo: boe.fiestas-laborales
├── doc.go, resolucion.go, derivar.go
├── grabacion_test.go                         # //go:build grabacion — TestGrabarFiestasLaborales
└── testdata/grabaciones.json                 # [datos]
internal/source/bocm/                         # nuevo: bocm.fiestas-locales
├── doc.go, disposicion.go, cruce.go, derivar.go
├── grabacion_test.go                         # //go:build grabacion — TestGrabarFiestasLocales
└── testdata/grabaciones.json                 # [datos]
internal/app/
├── plazos.go                                 # el applet
├── registro.go, ejemplo/kitlegal-e2e/main.go # en los dos registros; FuentesDePlazos
└── testdata/plazos/                          # golden [datos]
internal/evals/
├── formato.go, juzgar.go, consultas.go, conjunto.go   # plazo, no_cubierto, calcular, reglas
├── plazos.go                                 # nuevo: formas fijas y su juicio
├── informe.go, umbrales.go                   # medirLosPlazos y los dos umbrales
├── grabaciones.go, grabacion_test.go         # la tercera raíz; el segundo manifiesto
└── testdata/grabaciones.json, testdata/boe.legislacion-consolidada/   # [datos]
internal/skills/referencias.go, cobertura.go  # dos generadores; LeerCobertura
internal/arch_test.go                         # plazos sin red
schemas/plazo.json, cobertura.yaml.json       # nuevos [datos]
schemas/eval.yaml.json                        # plazo, no_cubierto, calcular [datos]
README.md, CONTRIBUTING.md, CHANGELOG.md, internal/evals/doc.go
```

**Structure Decision**: la de `docs/ROADMAP.md` §2 y `CLAUDE.md`: el dominio en `internal/core/plazos`, que §2 ya
nombra; los adaptadores de los dos boletines en `internal/source/<fuente>`; la composición en `internal/app`; la skill
en `skills/`, sus evals en `evals/`, y lo que configura cada territorio en `data/` (research D1, D12).

## Aceptación e2e

**Aceptación e2e:** cinco guiones `testscript`, que escribe la primera tarea en
`specs/023-h9-internal-core-plazos/aceptacion/` y quedan congelados; el workflow los activa como
`internal/app/testdata/script/h9-<nombre>.txtar`. Usan solo los casos cuyo día decisivo es cierto antes de grabar
(research D27; contracts/casos-de-plazos.md, casos 1, 2, 3, 15, 17 y 18).

| Guion | Historia | Qué describe | Requisitos |
|---|---|---|---|
| `plazos-calcular.txtar` | US1, US5 | Caso (a) con `--json` y con `--offline`: 0, `fecha` `2026-10-13`, el precepto `BOE-A-2015-10565` `a124` `124.1`, Leganés con su código, los festivos `2026-10-09` (local, `28074`, `BOCM-20251212-34`) y `2026-10-12` (nacional, `BOE-A-2025-21667`), sin avisos; con y sin `--offline`, el mismo `data` y el mismo `hash`. Caso (e): `2026-09-15` con los preceptos `a46` y `a128` de `BOE-A-1998-16718`. Caso 15: `2026-12-28` con `dias-del-art-183-lopj` | FR-001 a FR-003, FR-005, FR-006, FR-010, FR-013; SC-009 |
| `plazos-cobertura.txtar` | US2 | Tordesillas (caso 2): 0, `2026-10-09`, `faltan-festivos-locales` cuyo texto nombra Tordesillas y 2026 y no dice que no tenga festivos; `cobertura` con `sin-festivos-locales`. Leganés hacia 2027 (caso 18): `2027-01-14` y el aviso de 2027. Antes de 2025 (caso 17): `2025-01-03` con los dos avisos | FR-026, FR-085; US2.1 a US2.3 |
| `plazos-codigos.txtar` | US1, US2 | Cada fila de 2 de contracts/applet-plazos.md §6 —tipo, fecha, duración (falta, no entera, a un tipo que la fija), unidad (falta, a un tipo que la fija), uno y tres municipios en la LPAC, dos en el contencioso—; Arroyomolinos, con sus dos candidatos tras el municipio pedido (2); un municipio que no existe (3); y el orden: un tipo mal con un municipio que no existe da 2 | FR-007, FR-010, FR-012 |
| `plazos-herramienta.txtar` | US1 | El servidor anuncia `plazos_calcular`; con los argumentos del caso (a) da el sobre de la orden salvo `fecha_consulta`; con un tipo que no existe, el error `argumentos`, y el servidor sigue | FR-001, FR-085 |
| `plazos-sin-efectos.txtar` | US1 | Con la caché y el grafo con contenido, `arbol` da el mismo listado antes y después de la orden y de la herramienta, también con `--offline`, `--no-graph` y `--dry-run` | FR-013, FR-028; SC-009 |

Lo que usan del arnés es lo que ya hay: `exec`, `cmp`, `stdout`, `arbol` y `mcp`. Afirman códigos, clases, claves,
valores de vocabulario, fechas, preceptos y `source`, y del texto de un aviso solo el municipio y el año que debe
nombrar. Cada uno falla hoy por su primera aserción: el binario no tiene el applet `plazos`.

La primera tarea no escribe ninguna eval ni ningún dato: las cinco evals entran con la tarea del formato que las
admite, porque `make ci` lee las evals de toda carpeta de `evals/` (precedente: H23). US3 no tiene guion: su
aceptación son los tests de derivación y de cotejo. US4 y US6, las evals en el job de cierre (SC-001) y, en `make ci`,
los tests con sesiones sintéticas.

## Controles mecánicos que este hito añade o toca

### Objetivos del `Makefile`

Ninguno nuevo ni cambiado. `test`, `test-integration`, `schema-check` y `skills-check` llevan más dentro: los tests de
abajo, dos esquemas publicados más (`plazo.json`, `cobertura.yaml.json`) y dos referencias más.

### Tests nuevos

| Test | Fichero | Cubre |
|---|---|---|
| `TestLeerTipos` | `internal/core/plazos/tipos_test.go` | FR-008, FR-030 (los siete y sus reglas) |
| `TestCargarCalendario`, `TestEscribirFestivos` | `calendario_test.go`, `festivos_test.go` | FR-020 a FR-026: la forma canónica que escribe `EscribirFestivos` y su lectura, ida y vuelta, y las consultas del calendario; ninguna regla de carga propia (contracts/datos-de-plazos.md §1) |
| `TestCalcular` | `calcular_test.go` | FR-002 a FR-012, con calendarios sintéticos: cada regla del cómputo, cada aviso, cada fila de errores y su orden |
| `TestPropiedadesDelComputo` | `propiedades_test.go` | FR-081, SC-004: las tres propiedades sobre la rejilla de research D22 |
| `TestAvisosDePlazos` | `avisos_test.go` | FR-048: códigos, etiquetas y textos; las etiquetas de `internal/evals` son las del dominio |
| `TestLeerResolucion`, `TestFestivosNacionalesDerivados` | `internal/source/boe/fiestas/*_test.go` | FR-020, FR-025, FR-083: la derivación igual a `data/festivos/nacionales-y-autonomicos.yaml`, cada fila con `source`, las 19 comunidades con festivos cada año |
| `TestLeerDisposicion`, `TestCruceConElTerritorio`, `TestFestivosLocalesDerivados`, `TestCotejoConElCatalogo` | `internal/source/bocm/*_test.go` | FR-021 a FR-024, FR-083, FR-084, SC-006, SC-007: las erratas de research V8, los nombres de V7, la derivación igual a `data/festivos/locales/13.yaml`, 179 de 179, cada línea, y las diez diferencias del `LEEME.md` |
| `TestAppletPlazos` | `internal/app/plazos_test.go` | FR-001, FR-013: los golden de `testdata/plazos/` contra `schemas/plazo.json`, ninguna operación de grafo, el mismo `data` y `hash` con `--offline`, el golden de `--describe` (FR-085), los enumerados de la salida iguales a los del dominio |
| `TestFestivoDeUnaParte` | `internal/app/plazos_test.go` | FR-011, FR-025 |
| `TestCasosDePlazos` | `internal/app/plazos_casos_test.go` | FR-082, SC-005: los veinte casos de contracts/casos-de-plazos.md |
| `TestPreceptosDeLosPlazos` | `internal/evals/grabaciones_test.go` | FR-010, FR-030, FR-031: contracts/datos-de-plazos.md §4.2 |
| `TestFormatoDePlazos`, `TestJuzgarPlazos` | `internal/evals/formato_test.go`, `plazos_test.go` | FR-048, FR-051, FR-057: claves, esquema, enumerado atado a `plazos.CodigosDeAviso`, formas fijas, motivos, el comando `calcular`, `prohibidos` con orden y con herramienta |
| `TestUmbralesDeLegalCore` | `internal/evals/informe_test.go` | FR-053 a FR-055, FR-087, SC-012: los siete casos de FR-087 con sesiones sintéticas |
| `TestUmbralesDeLasTresSkills` | `internal/evals/informe_test.go` | FR-056, SC-011: con las evals del repositorio y el plan del job, los seis de `legal-core`, los trece de `boe-legislacion` (nueve deciden) y los quince de `jurisprudencia` (once deciden), en su orden |
| `TestLeerCobertura`, `TestRenderizarRecursosYPlazos`, `TestRenderizarCobertura` | `internal/skills/*_test.go` | FR-031, FR-032 |

### Tests y guiones existentes que cambian

La lista la da el prototipo (research V11, V15):

- **Por el applet**: `TestPuntoDeEntrada` (`cmd/kitlegal/main_test.go`), `TestRegistroDeProduccion`
  (`internal/app/registro_test.go`) y `TestRegistroDeE2E` (`internal/app/ejemplo/kitlegal-e2e/main_test.go`), la
  lista de applets; `TestHerramientasDelServidor` (`internal/app/herramientas_test.go`), de 12 y 15 herramientas a 13 y
  16; `TestEsquemasPublicados` y `TestEsquemasCubrenTodosLosVerbos` (`internal/app/esquemas_test.go`), la fila de
  `plazo.json` y su descripción; `TestArquitectura` (`internal/arch_test.go`), los ficheros sin red.
- **Guiones e2e**, en `internal/app/testdata/script/`, todos `[datos]`: `argumentos.txtar` (la lista de applets) y
  `h21-mcp-herramientas.txtar`, `h21-mcp-proceso.txtar` y `h21-mcp-protocolo.txtar` (de 15 a 16 herramientas).
- **Por las evals y el job**: `TestEvalsDelRepositorio` y `TestConjuntoDeEvals` (`internal/evals/conjunto_test.go`), con
  las nueve evals, la regla nueva y el subtest de grabado de `legal-core`; `TestConsultasNecesarias`, con la forma
  `calcular`; `TestManifiestoDeGrabaciones` y `TestGrabacionesSinSolape`, con el segundo manifiesto y la tercera raíz;
  `TestDefinicionDelJob` (`internal/evals/definicion_test.go`), solo el literal `peorCasoDeLegalCore` y las líneas de
  los dos subtests que lo llevan (research D24).
- **Por la skill**: `TestSkillsDelRepositorio` regenera la tabla y las dos referencias nuevas.

`internal/empaquetado` y los guiones de la instalación de desarrollo no cambian (research V11, V14). Ninguno se
desactiva ni se salta, y ninguno pierde lo que comprobaba: cada cambio es el dato nuevo de una lista o una cifra.

### Fixtures, `testdata/` y `schemas/` (tareas `[datos]`)

- Los tres manifiestos de research D15, y lo que graba `grabar_datos`: `evidencias/h9/` y
  `internal/evals/testdata/boe.legislacion-consolidada/`.
- `internal/app/testdata/plazos/`: `describe.json` y los golden de los casos (a), (b) y (e).
- `schemas/plazo.json` (desde `--describe`), `schemas/cobertura.yaml.json` y `schemas/eval.yaml.json`.
- Los cuatro guiones de arriba.

### CI

`.github/workflows/` no cambia (FR-058). El job de evals lo lanza el workflow sobre la propuesta de cambio, tras la
revisión final; `legal-core` ya va a dos sesiones (#157).

## Controles de umbral

Una fila por cada requisito o criterio con un umbral que se mide sin persona. En el cierre, los umbrales del informe
de `legal-core` que deciden desde este hito y los de `boe-legislacion` y `jurisprudencia` que ya decidían (FR-086); en
`make ci`, cada umbral de los controles del hito. `red` vacía y ninguna sesión sin medir no son elementos de
`umbrales`: los hace cumplir el veredicto del job desde H7.3, como antes.

| Requisito | Umbral | Control | Dónde |
|---|---|---|---|
| FR-053, FR-088, SC-001 (modo orden) | 0 respuestas sin la línea `⚠ PLAZO:` con la fecha esperada, de las 9 de las evals de plazo | umbral del informe con `decide: true`: con una, motivo con la sesión, veredicto `fallo` y el trabajo de `legal-core` en rojo, que el cierre cuenta | `evals:legal-core:plazo_sin_fecha:claude-sonnet-5-5:orden` |
| FR-053, FR-088, SC-001 (modo herramienta) | ídem | ídem | `evals:legal-core:plazo_sin_fecha:claude-sonnet-5-5:herramienta` |
| FR-054, FR-088, SC-001 (modo orden) | 0 respuestas sin la declaración con la norma esperada o con una línea `⚠ PLAZO:`, de las 6 de las evals de cobertura | ídem | `evals:legal-core:cobertura_sin_declarar:claude-sonnet-5-5:orden` |
| FR-054, FR-088, SC-001 (modo herramienta) | ídem | ídem | `evals:legal-core:cobertura_sin_declarar:claude-sonnet-5-5:herramienta` |
| FR-045, SC-001 (modo orden) | 0 respuestas sin la skill activada, de las 21 del modo | umbral del informe con `decide: true` | `evals:legal-core:sin_activar:claude-sonnet-5-5:orden` |
| FR-045, SC-001 (modo herramienta) | ídem | ídem | `evals:legal-core:sin_activar:claude-sonnet-5-5:herramienta` |
| FR-071, FR-086 (`boe-legislacion` sigue como decidía) | sus nueve umbrales que deciden, cumplidos; si no, queda en el informe para la persona | los de hoy, que no cambian | `evals:boe-legislacion:sin_activar:claude-sonnet-5-5:orden`, `evals:boe-legislacion:sin_activar:claude-sonnet-5-5:herramienta`, `evals:boe-legislacion:afirma_lo_no_leido:claude-sonnet-5-5`, `evals:boe-legislacion:medida_del_juez:afirma_lo_no_leido:defectos_sin_marcar`, `evals:boe-legislacion:medida_del_juez:afirma_lo_no_leido:correctos_marcados`, `evals:boe-legislacion:duracion_de_las_sesiones:orden`, `evals:boe-legislacion:duracion_de_las_sesiones:herramienta`, `evals:boe-legislacion:duracion_del_juez:orden`, `evals:boe-legislacion:duracion_del_juez:herramienta` |
| FR-071, FR-086 (`jurisprudencia` sigue como decidía) | sus once umbrales que deciden, cumplidos; si no, queda en el informe para la persona | los de hoy, que no cambian | `evals:jurisprudencia:sin_activar:claude-sonnet-5-5:orden`, `evals:jurisprudencia:sin_activar:claude-sonnet-5-5:herramienta`, `evals:jurisprudencia:cita_sin_documento:claude-sonnet-5-5:orden`, `evals:jurisprudencia:cita_sin_documento:claude-sonnet-5-5:herramienta`, `evals:jurisprudencia:documento_sin_cotejar:claude-sonnet-5-5:orden`, `evals:jurisprudencia:documento_sin_cotejar:claude-sonnet-5-5:herramienta`, `evals:jurisprudencia:afirma_lo_no_leido:claude-sonnet-5-5`, `evals:jurisprudencia:medida_del_juez:afirma_lo_no_leido:defectos_sin_marcar`, `evals:jurisprudencia:medida_del_juez:afirma_lo_no_leido:correctos_marcados`, `evals:jurisprudencia:duracion_del_juez:orden`, `evals:jurisprudencia:duracion_del_juez:herramienta` |
| FR-050, SC-001 (duración) | el tope de 352 minutos ≥ el peor caso de `legal-core` con las evals del repositorio, 14 629 s | la definición del job contra las evals del repositorio | `ci:internal/evals/definicion_test.go:TestDefinicionDelJob` |
| FR-055, FR-087, SC-012 | 1 sesión que cuenta → `fallo`; 0 → se cumple; los siete casos de FR-087, 7 de 7 | sesiones sintéticas, por respuesta y por informe | `ci:internal/evals/informe_test.go:TestUmbralesDeLegalCore` |
| FR-056, SC-011 | 6 elementos en `legal-core`, los 6 con `decide: true`; 13 con 9 en `boe-legislacion`; 15 con 11 en `jurisprudencia`; en su orden | el informe con las evals del repositorio y el plan del job | `ci:internal/evals/informe_test.go:TestUmbralesDeLasTresSkills` |
| FR-081, SC-004 | 0 contraejemplos de las tres propiedades sobre la rejilla | el recorrido entero de la rejilla | `ci:internal/core/plazos/propiedades_test.go:TestPropiedadesDelComputo` |
| FR-082, SC-005 | 20 casos, 0 diferencias de fecha o de avisos | la tabla | `ci:internal/app/plazos_casos_test.go:TestCasosDePlazos` |
| FR-011 | el festivo de una parte cuenta como hábil y avisa: 0 diferencias | el primer festivo con parte de lo grabado | `ci:internal/app/plazos_test.go:TestFestivoDeUnaParte` |
| FR-020, FR-083, SC-006 (nacionales y autonómicos) | 0 bytes de diferencia con lo derivado; 0 filas sin `source`; 19 de 19 comunidades con festivos cada año | la derivación desde `evidencias/h9/` | `ci:internal/source/boe/fiestas/derivar_test.go:TestFestivosNacionalesDerivados` |
| FR-021, FR-022, FR-083, SC-006 (locales) | 0 bytes de diferencia; 0 filas sin `source`; 179 de 179 municipios con festivos o falta en 2025 y en 2026; 0 líneas sin festivos ni falta | la derivación desde `evidencias/h9/` y `data/territorio/` | `ci:internal/source/bocm/derivar_test.go:TestFestivosLocalesDerivados` |
| FR-084, SC-007 | exactamente las 10 diferencias del `LEEME.md`, ninguna más ni ninguna menos | el cotejo por código INE con las filas de `entidad_codigo` 0 | `ci:internal/source/bocm/cotejo_test.go:TestCotejoConElCatalogo` |
| FR-030, FR-031, FR-010 | 0 normas o bloques sin grabar; 0 duraciones fuera de su apartado; 0 identificadores distintos de la búsqueda grabada; 0 palabras clave ausentes | la lectura de cada bloque con `boe articulo` sobre la caché preparada | `ci:internal/evals/grabaciones_test.go:TestPreceptosDeLosPlazos` |
| FR-052 | 0 consultas necesarias de las evals de `legal-core` sin su grabación | `Preparar` y `ComprobarSinRed` sobre la unión de grabaciones | `ci:internal/evals/conjunto_test.go:TestEvalsDelRepositorio` |
| FR-051 | 9 de 9 evals de `legal-core` válidas contra el esquema y con sus reglas del conjunto | la validación y las reglas | `ci:internal/evals/conjunto_test.go:TestEvalsDelRepositorio`, `ci:internal/evals/formato_test.go:TestFormatoDePlazos` |
| FR-013, SC-009 (red) | 0 caminos de dependencia a `net`, `net/http` o `internal/httpx` | el recorrido del grafo de importaciones | `ci:internal/arch_test.go:TestArquitectura` |
| FR-013, SC-009 (grafo, caché, `--offline`) | 0 operaciones de grafo; 0 cambios en la caché y el grafo; `data` y `hash` iguales con `--offline` | el resultado del applet; `arbol` antes y después; las dos salidas | `ci:internal/app/plazos_test.go:TestAppletPlazos`, `ci:internal/app/testdata/script/h9-plazos-sin-efectos.txtar`, `ci:internal/app/testdata/script/h9-plazos-calcular.txtar` |
| FR-012, FR-085 | cada fila de §6 con su código; 2 municipios cubiertos y no cubiertos con su cobertura | los guiones | `ci:internal/app/testdata/script/h9-plazos-codigos.txtar`, `ci:internal/app/testdata/script/h9-plazos-cobertura.txtar` |
| FR-001, FR-085 (herramienta) | 13 herramientas en producción y 16 con las de ejemplo; `plazos_calcular` con el sobre de su orden | el control de conformidad de H21 y el guion | `ci:internal/app/herramientas_test.go:TestHerramientasDelServidor`, `ci:internal/app/testdata/script/h9-plazos-herramienta.txtar` |
| FR-046, SC-010 | `SKILL.md` de `legal-core` < 300 líneas; tabla y referencias iguales a las regeneradas | la comprobación de skills | `ci:Makefile:skills-check`, `ci:internal/app/skills_test.go:TestSkillsDelRepositorio` |
| FR-032, SC-010 | 0 diferencias en `references/recursos_y_plazos.md` y `cobertura_por_materia.md` | ídem | `ci:Makefile:skills-check` |
| FR-089, SC-008 | con cada mutante, al menos 1 test falla: todos de todos | los tests de esta tabla; la tarea de FR-089 los ejecuta con cada mutante y lo deja en `mutantes.md` | `ci:internal/app/plazos_casos_test.go:TestCasosDePlazos`, `ci:internal/source/bocm/derivar_test.go:TestFestivosLocalesDerivados`, `ci:internal/evals/informe_test.go:TestUmbralesDeLegalCore`, `ci:internal/evals/informe_test.go:TestUmbralesDeLasTresSkills` |

SC-002 lo decide el informe final con las filas de arriba; SC-003 lo mide una persona; SC-013, la revisión final por
lectura: ninguno tiene fila.

## Uso, de fuera adentro (criterio de uso, ADR 0028)

El detalle, con ejemplos y bytes, está en contracts/applet-plazos.md §8, contracts/evals-legal-core.md §6 y
contracts/skill-legal-core.md §5.

- **El resultado de `plazos calcular`** (la skill, paso 9; una invocación por plazo, casi siempre una por pregunta): la
  skill escribe menos de 100 bytes y recibe 772 bytes en el caso (a) y 858 en el (b) (research M1, M2), con las claves
  de FR-001 y ninguna más; cada festivo
  saltado, unos 90 bytes; cada aviso, unos 250; cada tramo de `cobertura`, unos 100. Acotado sea cual sea la
  duración: menos de 15 KB en el peor caso (contracts/applet-plazos.md §8).
- **Las señales** (avisos y `cobertura`): cada una solo si el tramo de esa llamada pasa por lo que la causa: es
  información de esa fecha. La de los locales de 2027 de la Comunidad de Madrid deja de darse cuando una release lleve
  el BOCM de 2027; la de una comunidad sin festivos locales, cuando entren los de su boletín (backlog «territorios»);
  la de una parte, cuando el tramo no pasa por ese día; la del art. 183, cuando no pasa por el 24 de diciembre al 6 de
  enero. Hasta qué disposición llegan los datos lo lleva el aviso de cobertura, y la skill no lo dice sin él (FR-041).
- **Las referencias** (la skill, una vez por conversación de plazos): unos 2 kB cada una, fijas.
- **`SKILL.md`** (el agente, una vez por conversación): 248 líneas, 15,9 kB.
- **Los umbrales** (el job, el informe final, la persona; una vez por job): seis elementos de unos 300 bytes en
  `legal-core`; un motivo, menos de 0,6 KB.

**Nada crece con lo acumulado.** `plazos` lee solo lo empotrado: tras meses de uso diario, con cientos de normas y
miles de bloques consultados, la llamada siguiente devuelve lo mismo que la primera. Las referencias y `data/festivos/`
son fijas hasta la tarea `[datos]` siguiente. Ninguna señal se repite sin información nueva: la disposición hasta la
que llegan los datos, que es igual en cada llamada, solo sale dentro de un aviso que la causa.

## Decisiones

- **El dominio en `internal/core/plazos`; los boletines, en dos adaptadores sin consulta en ejecución** (D1).
- **Todo argumento es una cadena que decide el dominio** (D2): un `enum` de Kong o una bandera entera rompen la tabla
  generada (research V12).
- **Un solo `generico` con unidad, como el silencio** (D3).
- **El silencio se produce el día siguiente al vencimiento del plazo máximo** (D8): arts. 21.3, 24.1, 24.4, 30.3, 30.4.
- **Agosto no corre: el vencimiento pasa un mes si el tramo lo toca; el último día, con el 182.1 LOPJ; los días del
  183 se cuentan y avisan** (D9): la fecha más temprana.
- **`cobertura` en `data`, por tramos de años, y un aviso que lleva la disposición de su ámbito** (D5, D6).
- **El cruce de nombres del BOCM: forma compacta y, si no, prefijo seguido de «de»; la línea con erratas de forma se
  lee** (D13): sin ello no se cumple FR-084 (research V7 a V9).
- **Tres manifiestos, uno por fuente y paquete; los preceptos, en una raíz nueva de `internal/evals`** (D15).
- **Ningún texto legal en `data/`: la verificación lee lo grabado por apartado y por palabras clave** (D16).
- **La tabla de casos en `internal/app`; los guiones de aceptación, solo con días decisivos ciertos** (D23, D27).
- **`TestDefinicionDelJob` cambia solo el literal del peor caso** (D24).
- Las demás, con su alternativa rechazada, en research D1-D30.

## Trazabilidad: cada mecanismo y su requisito

| Mecanismo | Requisito |
|---|---|
| `internal/core/plazos`: `Calendario`, `CargarCalendario` | FR-003, FR-006, FR-007, FR-011, FR-020 a FR-026 |
| `LeerTipos` y `data/plazos.yaml` | FR-001, FR-008, FR-030 |
| `Calcular` y el cómputo de la LPAC | FR-002 a FR-007, FR-009 |
| El cómputo del contencioso | FR-010 |
| Los cuatro avisos y sus textos | FR-010, FR-011, FR-026, FR-041, FR-048 |
| `cobertura` en `data` | FR-001, FR-026; principio IX |
| Los errores y su orden | FR-012 |
| `EscribirFestivos` (forma canónica) | FR-020, FR-083 |
| El applet `plazos`, sus dos registros, `FuentesDePlazos` y `data/datos.go` | FR-001, FR-013, FR-028 |
| `schemas/plazo.json`, los golden y el de `--describe` | FR-085; constitución III |
| `internal/source/boe/fiestas`: manifiesto, `TestGrabarFiestasLaborales`, `LeerResolucion`, la derivación | FR-020, FR-025, FR-027, FR-083 |
| `internal/source/bocm`: manifiesto, `TestGrabarFiestasLocales`, `LeerDisposicion`, el cruce, la derivación | FR-020 a FR-024, FR-027, FR-083 |
| `territorio.Registro.MunicipiosDe`, con su test en `registro_test.go` | FR-022: el cruce por código INE con los municipios de la comunidad del boletín, sin un segundo lector de `data/territorio/` |
| `TestCotejoConElCatalogo` | FR-084 |
| `data/cobertura.yaml`, `schemas/cobertura.yaml.json`, `LeerCobertura` | FR-031, FR-032 |
| El manifiesto de `internal/evals/testdata`, la raíz en `UnionDeGrabaciones` y `TestGrabarEvals` con dos manifiestos | FR-027, FR-030, FR-031, FR-052 |
| `TestPreceptosDeLosPlazos` | FR-010, FR-030, FR-031 |
| Los dos generadores de referencias | FR-032 |
| `skills/legal-core/SKILL.md` v1 | FR-040 a FR-048 |
| `plazo`, `no_cubierto` y `calcular` en el formato y su esquema | FR-051 |
| Las formas fijas y su juicio (`internal/evals/plazos.go`) | FR-048, FR-057 |
| Las dos reglas del conjunto de `legal-core` | FR-051, FR-088 |
| Las cinco evals | FR-050, FR-052, FR-088 |
| `medirLosPlazos` y los dos umbrales, con su motivo | FR-053 a FR-056 |
| `TestPropiedadesDelComputo`, `TestCasosDePlazos`, `TestFestivoDeUnaParte` | FR-011, FR-081, FR-082 |
| `TestUmbralesDeLegalCore`, `TestUmbralesDeLasTresSkills` | FR-087, FR-056 |
| `internal/core/plazos` e `internal/app/plazos.go` en la garantía sin red | FR-013 |
| El guion `plazos-sin-efectos` | FR-013, FR-028 |
| La tarea de los mutantes y `mutantes.md` | FR-089 |
| `CHANGELOG.md`, `README.md`, `CONTRIBUTING.md`, `internal/evals/doc.go` | FR-060 a FR-062 |
| La nota del cierre | FR-091: la lleva el informe final del workflow; ninguna tarea |

Nada del diseño atiende a un estado que no pasa el umbral de materialidad. Un fichero de `data/` empotrado que no
carga, una grabación tocada a mano o una evidencia que falta siguen la regla genérica —defecto `inesperado`, código 1,
o el test que falla—, sin caso propio ni mensaje prometido.

## Datos externos

Tres fuentes, las tres con fila revisada en `docs/SOURCES.md` de `main` (research V1), grabadas solo por el paso
`grabar_datos` tras el commit de la tarea `[datos]` que deja su manifiesto y su `TestGrabar*` (research D15;
contracts/datos-de-plazos.md §4.1):

| Fuente | Manifiesto | Test | Material | Lo que sale de él |
|---|---|---|---|---|
| `boe.fiestas-laborales` | `internal/source/boe/fiestas/testdata/grabaciones.json` | `TestGrabarFiestasLaborales` | `evidencias/h9/boe.fiestas-laborales/` (tres resoluciones) | `data/festivos/nacionales-y-autonomicos.yaml`, por `TestFestivosNacionalesDerivados -actualizar-festivos` |
| `bocm.fiestas-locales` | `internal/source/bocm/testdata/grabaciones.json` | `TestGrabarFiestasLocales` | `evidencias/h9/bocm.fiestas-locales/` (siete disposiciones) | `data/festivos/locales/13.yaml`, por `TestFestivosLocalesDerivados -actualizar-festivos` |
| `boe.legislacion-consolidada` | `internal/evals/testdata/grabaciones.json` | `TestGrabarEvals` | `internal/evals/testdata/boe.legislacion-consolidada/` | ningún fichero de `data/`: las sesiones de las evals (FR-052) y la verificación de `data/plazos.yaml` y `data/cobertura.yaml` (FR-030, FR-031) |

Ningún fichero de `data/festivos/` se escribe a mano. `data/plazos.yaml` y `data/cobertura.yaml` los fija este plan
desde el spec, como `data/normas.yaml`, y los comprueba lo grabado; el identificador del texto refundido de Tráfico
sale de la búsqueda grabada (research S5). Cada id de bloque del manifiesto de preceptos es el del índice grabado de su
norma —en la LOPJ, `acientoochentaydos`, `acientoochentaytres` y `acientoochentaycinco`; en el Código Civil, `art5`
(research V20, V29)—, salvo el `a1` del texto refundido de Tráfico, cuyo índice no está grabado: si no existe,
`TestGrabarEvals` falla y el run se detiene con el dossier de `grabar_datos` (research S8). El fichero CKAN y el `LEEME.md` de `evidencias/adr-0039/` solo los lee el
cotejo; ninguna tarea los cambia. Ninguna fuente sin fila revisada entra: no la hay en este hito.

## Orden de implementación (de dentro afuera)

1. **`[aceptacion]`** Los cinco guiones de `aceptacion/`, sin código de producto, sin datos y sin evals.
2. **`[datos]`** Los tres manifiestos y sus `TestGrabar*` (`internal/source/boe/fiestas`, `internal/source/bocm` con
   su `doc.go`, y el segundo manifiesto de `internal/evals` con `TestGrabarEvals` recorriendo los dos y
   `TestManifiestoDeGrabaciones`). Tras su commit graba `grabar_datos`.
3. `internal/core/plazos`: tipos, calendario, festivos, cómputo, avisos, errores y salida, con sus tests y la rejilla
   de las propiedades, sobre calendarios sintéticos.
4. **`[datos]`** `data/plazos.yaml`, `data/cobertura.yaml`, `schemas/cobertura.yaml.json`, `LeerCobertura`, la
   tercera raíz en `UnionDeGrabaciones`, `TestGrabacionesSinSolape` y `TestPreceptosDeLosPlazos`.
5. **`[datos]`** El BOE: `LeerResolucion`, la derivación, `data/festivos/nacionales-y-autonomicos.yaml` y sus tests.
6. **`[datos]`** El BOCM: `territorio.Registro.MunicipiosDe`, `LeerDisposicion`, el cruce, la derivación,
   `data/festivos/locales/13.yaml`, sus tests y el cotejo con el fichero CKAN.
7. **`[datos]`** El applet en los dos registros: `internal/app/plazos.go`, `data/datos.go`, `registro.go`, el binario
   de e2e, `schemas/plazo.json`, los golden, `internal/arch_test.go`, los siete tests de la lista de applets y de
   herramientas, los cuatro guiones e2e del applet y la línea de `CONTRIBUTING.md` que enumera los applets.
8. La tabla de casos y el festivo de una parte (`internal/app/plazos_casos_test.go`, `plazos_test.go`).
9. **`[datos]`** El formato de eval y su juicio: `schemas/eval.yaml.json`, `formato.go`, `plazos.go`, `juzgar.go`,
   `consultas.go`, las reglas del conjunto, las cinco evals de `evals/legal-core/` y `TestDefinicionDelJob`.
10. Los umbrales: `informe.go` y `umbrales.go`, con `TestUmbralesDeLegalCore` y `TestUmbralesDeLasTresSkills`.
11. La skill: los dos generadores, `skills/legal-core/SKILL.md` desde el borrador y `make skills-sync`.
12. Los mutantes de FR-089: cada uno aplicado, el test que falla anotado en `mutantes.md` y el código restaurado.
13. La documentación: `CHANGELOG.md` (*Unreleased*), `README.md`, `CONTRIBUTING.md` e `internal/evals/doc.go`.

**Obligaciones para `tasks.md`**: una sola tarea `[aceptacion]`, la primera; cada tarea deja `make ci` en verde, y
por eso las de los pasos 7 y 9 llevan en la misma tarea el código, los tests y los guiones que su cambio rompe; las
tareas `[datos]` son las de los pasos 2, 4, 5, 6, 7 y 9, con las rutas de este plan —cada guion e2e por su nombre—; la
del paso 2 deja los manifiestos y no deriva nada, y las de los pasos 4 a 6 parten de lo grabado; cada fila de
«Controles de umbral» tiene la tarea que construye su control, con un test que lo ve fallar; ninguna tarea ni
corrección cumple un umbral de FR-053, FR-054 o `sin_activar` rebajándolo, dejándolo en `decide: false`, sacando
sesiones de su total o cambiando las evals (FR-071); ninguna escribe en `evidencias/`, ni un fichero de
`data/festivos/` a mano, ni un festivo, un identificador o un texto de una norma de memoria; ninguna toca
`.github/workflows/`, `skills/boe-legislacion/`, `skills/jurisprudencia/`, `evals/boe-legislacion/`,
`evals/jurisprudencia/`, `internal/mcp`, `docs/SOURCES.md`, `docs/ROADMAP.md`, `docs/ADR/`, `docs/USO.md`,
`docs/HISTORIAL.md`, `CLAUDE.md` ni `web/`; ninguna usa la red ni `KITLEGAL_RECORD`, abre una sesión con modelo ni
ejecuta `make evals`, `make evals-sondeo`, `scripts/evals*.sh` o `TestSondeo`; ninguna publica ni mide en la
plataforma; y los escenarios 14 y 15 del quickstart no son tareas.

## Complexity Tracking

| Desviación | Por qué hace falta | Alternativa más simple rechazada |
|---|---|---|
| Dos adaptadores de fuente nuevos bajo `internal/source/` que no consultan nada en ejecución | Cada fuente que se graba lleva su manifiesto con una sola `fuente` en su paquete (research V2), y el formato de un boletín es de un adaptador (constitución IV). Son una revisión de la persona al leer el informe (constitución, capa 3) | Un solo paquete para las dos fuentes: `grabar_datos` no lo admite. El analizador en el dominio: no es dominio |
| `internal/evals` gana una tercera raíz de grabaciones y `TestGrabarEvals` un segundo manifiesto | `grabar_datos` no puede grabar en `testdata/evals/` (research V3) y el manifiesto de `internal/source/boe` se graba entero (V4) | Ampliar uno de los dos manifiestos de hoy |
| El cruce de nombres del BOCM va más allá del pliegue de `territorio`: forma compacta y prefijo seguido de «de» | Seis nombres del BOCM no son los del INE (research V7); sin la regla, FR-084 da cuatro diferencias de más | Una tabla de alias, que es un caso por municipio (principio IX) y no sale de ninguna fuente; el fichero CKAN como tabla, que el ADR 0039 excluye |
| La gramática de la línea del BOCM tolera erratas de forma | Seis líneas de 2025 y 2026 las tienen (research V8); tratarlas como líneas sin leer da otras tantas diferencias con el fichero CKAN y deja sin festivos a municipios que el boletín sí da | La gramática estricta |
| `TestDefinicionDelJob` cambia su literal del peor caso de `legal-core`, aunque FR-058 dice que no se tocan los tests que leen la concurrencia | Con nueve evals el peor caso es otro (research V15), y FR-050 hace de ese test el control del tope; no cambia ningún valor de concurrencia ni la definición | Dejarlo: `make ci` en rojo |
| Tareas `[datos]` con código y tests | El applet, su esquema, sus golden y los guiones que enumeran applets tienen que entrar juntos para que `make ci` quede en verde; igual el formato con las evals. Precedente: H7.2 a H23.1 | Separar datos y código: `make ci` en rojo entre dos tareas |

## Comprobación contra la rúbrica del juez (`juez_plan`) y `precheck.sh plan`

- `precheck.sh plan`: existen `plan.md` y `research.md`; ninguno conserva marcas de aclaración pendiente; están
  `## Constitution Check`, la línea «Aceptación e2e» y `## Controles de umbral`, con cada fila nombrando requisitos
  definidos en el spec y un control con forma (`ci:` o `evals:`).
- a · Constitution Check: un ítem por principio (I-IX) y por regla de dependencia (R1-R6, y R7 del test).
- b · Dependencias: ninguna nueva.
- c · Reglas de dependencia: el dominio sin adaptadores ni `os`/`io`; ni `net/http`, ni SQLite, ni `os.Exit`, ni salida
  estándar nuevos.
- d · Errores y códigos: `argumentos` → 2, `no-encontrado` → 3, `inesperado` → 1, en el orden de FR-012.
- e · Tests primero: cinco guiones congelados con sus requisitos; tests nuevos, los que cambian, fixtures y tareas
  `[datos]`.
- f · Alcance: nada fuera del spec. Lo que el spec no nombraba y el plan añade para cumplirlo —la regla de cruce, la
  gramática tolerante, la raíz nueva, el literal de `TestDefinicionDelJob`, la unidad rechazada en un tipo que la fija—
  está en Complexity Tracking o en `gates/supuestos.md`, con su requisito.
- g · Sin atajos: ningún `//nolint`, `t.Skip`, TODO ni error silenciado previstos.
- h · Mejor alternativa: cada decisión con la rechazada (D1-D30).
- i · Afirmaciones verificadas: V1-V29 con fichero, orden, sonda o prototipo —cada id de bloque del manifiesto de
  preceptos, contra el índice grabado de su norma (V20, V29)—; M1-M5 medidas; S1-S8 como supuestos, con lo que pasa si
  no se cumplen.
- j · Quickstart ejecutable: órdenes, rutas, banderas y datos reales; todo bajo `$T`, borrado al final; el 14 es del
  workflow y el 15 de una persona.
- k · Datos externos: tres fuentes revisadas, tres manifiestos, tres `TestGrabar*`, lo derivado por código.
- l · Autonomía: ninguna tarea para una persona; el cierre en la plataforma lo hace el workflow.
- m · Uso: «Uso, de fuera adentro» y los tres contratos, con bytes medidos, invocaciones por pregunta y lo que apaga
  cada señal.
- n · Proporcionalidad: «Trazabilidad»; sin mecanismos para estados sin vía real.
- o · Controles de umbral: los seis de `legal-core` que deciden en el cierre, los que ya decidían en las otras dos
  skills, y una fila por cada umbral medido en `make ci`.
