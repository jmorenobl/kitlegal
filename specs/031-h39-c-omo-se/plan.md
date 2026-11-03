# Implementation Plan: H39 · Cómo se organiza una skill para que se active y se lea en los dos modos

**Branch**: `031-h39-c-omo-se` | **Date**: 2026-11-03 | **Spec**: [spec.md](spec.md)

**Input**: Feature specification from `specs/031-h39-c-omo-se/spec.md`

## Summary

Dos entregas. **La base**: la ayuda de cada argumento y bandera propios de un verbo pasa a ser la `description` de su
propiedad en el esquema de entrada de `--describe` y, con él, en el `inputSchema` de su herramienta MCP; las banderas
globales, las tablas de comandos de las skills y lo demás del esquema no cambian (FR-001 a FR-005). **El laboratorio**:
una skill pequeña fuera del producto (`laboratorio/`), con una regla testigo de forma que decide un guion, ocho
variantes que difieren de la de control en las partes que declaran y cuatro evals servidas por la grabación de hoy; el
arnés con el que una persona lanza el sondeo por variante y modo con el modelo y la versión de Claude Code que pide, y
el guion que cuenta sus transcripciones con intervalos de Wilson (FR-010 a FR-043). El run no abre ninguna sesión con
modelo: el sondeo y el ADR que precisa el 0046 los hace la persona antes de fusionar (FR-050 a FR-052).

Enfoque técnico (research D1 a D13): `describirArgumentos` en `internal/cli`, como `anotarBanderas`; el laboratorio en
`laboratorio/`, que no se empotra, no se instala y no es de lo que mide el job; el código del arnés y del guion en
`internal/evals`, que reutiliza su reparto de sesiones, su preparación sin red y su lector de transcripts; y
`cmd/laboratorio`, un programa del repositorio como `cmd/empaquetar`, con dos órdenes: `servidor`, el filtro de líneas
que da a las sesiones de `instrucciones` y `parametro` otro texto sin que `kitlegal` gane ninguna bandera (FR-029), y
`tabla`, el guion.

## Technical Context

**Language/Version**: Go 1.27 (`go 1.27.0`, `toolchain go1.27.3` en `go.mod`).

**Primary Dependencies**: las fijadas, sin ninguna nueva: `alecthomas/kong` v1.16.1, `invopop/jsonschema` v0.14.0,
`gopkg.in/yaml.v3`, `stretchr/testify`, `rogpeppe/go-internal` (testscript) y `modelcontextprotocol/go-sdk` v1.8.0, que
sigue importando solo `internal/mcp` (R7): el filtro del laboratorio trabaja con líneas JSON (research V13, V14).

**Storage**: ficheros: `laboratorio/` versionado; el directorio de cada sondeo, en el equipo de la persona; la caché
SQLite de cada sesión, como hoy.

**Testing**: `go test -race` en `make ci`; un guion testscript de aceptación; fixtures nuevos en
`internal/evals/testdata/laboratorio/`; un sustituto de `claude` para el arnés (como los de
`internal/evals/sustitutos_test.go`); ningún test con red ni con modelo.

**Target Platform**: macOS y Linux, donde corren `make ci` y el sondeo de la persona.

**Project Type**: CLI multicall (`kitlegal`) y herramientas del repositorio (`cmd/laboratorio`, guiones de `scripts/`).

**Performance Goals**: ninguno. Cada sesión del sondeo tiene el tope de hoy (`topeDeUnaSesion`).

**Constraints**: `kitlegal` no gana ninguna bandera, variable ni orden (FR-029); `skills/` no cambia (FR-004); el run no
graba, no escribe en `evidencias/` y no abre sesiones con modelo (FR-012, FR-032); toda petición de las sesiones del
sondeo, a los proxies que rechazan (ADR 0043).

**Scale/Scope**: 28 ayudas de argumento (22 en verbos con herramienta), 7 de los 8 ficheros de `schemas/` que salen de
`--describe`, 8 variantes, 4 evals, 14 filas y 560 sesiones en la aceptación.

## Constitution Check

*GATE: Must pass before Phase 0 research. Re-check after Phase 1 design.*

| Principio | Cómo lo cumple este hito |
|---|---|
| I · Fuentes públicas y frontera humana (INNEGOCIABLE) | No toca ninguna fuente ni `internal/httpx`, y no añade ninguna petición. Las sesiones del sondeo se sirven de `testdata/evals/` con los proxies que rechazan toda petición del reparto de hoy (ADR 0043); el filtro de `laboratorio servidor` no hace red. Nada se presenta ni se firma |
| II · Nada sin cita ni fuente | El sobre de cada applet no cambia: solo el esquema de entrada gana `description`. La skill de laboratorio cita norma y bloque como las del producto, y no la ve ninguna persona ni skill del producto |
| III · Tests primero y offline | Primero el guion `aceptacion/describe.txtar`, que falla hoy por una aserción (no hay `description` en las propiedades). Todo test nuevo, offline y sin modelo, contra fixtures; `schemas/` regenerado y validado por `make schema-check`. `internal/core` no cambia; el código nuevo de `internal/evals` lleva sus tests en `make ci` |
| IV · Arquitectura hexagonal con reglas ejecutables | El cambio de los esquemas vive en `internal/cli`; el código del laboratorio, en `internal/evals`, que ningún paquete de producción importa; `cmd/laboratorio` es la raíz que nombra `os.Stdin`, `os.Stdout` y `os.Stderr` y llama a `os.Exit`, como `cmd/empaquetar` (exentos de R4 y R5 en `.golangci.yml`). Errores: defecto `inesperado` con 1 en `laboratorio`; el arnés, como el sondeo de hoy (uso → 1 con su mensaje). Ningún `panic` |
| V · Simplicidad y dependencias fijadas (YAGNI) | Ninguna dependencia nueva. Cada mecanismo se traza a un FR (§Diseño y Complexity Tracking): ni opciones, ni estados ni casos límite que el spec no pida; lo que llega de fuera manipulado es defecto `inesperado` |
| VI · Un binario, convenciones de agente | Un solo binario `kitlegal`, sin banderas nuevas; `--describe` sigue siendo la fuente de las herramientas MCP y de las tablas, y ahora lleva la ayuda de cada argumento. `cmd/laboratorio` no se distribuye ni se instala (`.goreleaser.yaml` construye solo `./cmd/kitlegal`) |
| VII · Grafo y privacidad | No toca el grafo. Las sesiones del sondeo usan su caché y su grafo en su directorio, como las del job; ningún dato de una persona |
| VIII · Skills primero; el binario es la herramienta | Mide dónde va en una skill lo que el agente tiene que cumplir, antes de reorganizar las del producto, y protege las de hoy: ninguna cambia y las evals del cierre las miden todas (FR-053). La ayuda de los argumentos en los esquemas es para las skills en modo herramienta. La skill de laboratorio cumple la forma de `docs/SKILLS.md` que comprueba `make ci` |
| IX · Genericidad territorial, validación local | Sin dimensión territorial: ninguna herramienta, dato ni skill con un municipio |
| X · Genericidad por materia, validación administrativa | Ningún caso especial para una materia en código, `data/` ni protocolo de skill. Las cuatro evals de laboratorio son de cuatro materias (constitucional, tributaria, administrativa, laboral) y viven fuera del producto |
| Regla 1 · `internal/core/**` no importa adaptadores | `internal/core` no cambia |
| Regla 2 · solo `internal/httpx` importa `net/http` | Nada nuevo importa `net/http`: el filtro usa `os/exec`, `bufio` y `encoding/json` |
| Regla 3 · solo `internal/{cache,store,graph}` importan SQLite | Nada nuevo importa `database/sql` ni `modernc.org/sqlite` |
| Regla 4 · solo `internal/cli` y `cmd/` llaman a `os.Exit` | Solo `cmd/laboratorio/main.go`, con el código que devuelve `internal/evals` |
| Regla 5 · solo `internal/render` escribe en stdout; logs a stderr | `internal/evals` escribe en el `io.Writer` que le inyecta `cmd/laboratorio`; `kitlegal` sigue escribiendo por su presentador |
| Regla 6 · `internal/graph` no importa `internal/source/*` ni `internal/render` | `internal/graph` no cambia |
| Regla 7 · solo `internal/mcp` importa el SDK de MCP | El filtro no lo importa; los tests del laboratorio usan `internal/mcp/mcptest`, como los de `internal/app` |

Sin violaciones. Re-evaluado tras el diseño de la fase 1: igual.

## Project Structure

### Documentation (this feature)

```text
specs/031-h39-c-omo-se/
├── plan.md
├── research.md
├── data-model.md
├── quickstart.md
├── contracts/
│   ├── esquemas.md          # FR-001 a FR-005
│   ├── laboratorio.md       # FR-010 a FR-014, FR-020 a FR-029
│   ├── arnes.md             # FR-029 a FR-033
│   └── tabla.md             # FR-040 a FR-043
├── aceptacion/describe.txtar   # la primera tarea; congelado
└── tasks.md
```

### Source Code (repository root)

```text
internal/cli/
├── describe.go                 # describirArgumentos, llamada en esquema() tras anotarBanderas (FR-001, FR-003)
├── herramienta.go              # la misma en EsquemasDeHerramienta (FR-002)
└── describe_test.go, herramienta_test.go   # TestAyudaComoDescripcion; las expectativas de structs con ayuda
internal/app/
├── ayudas_test.go              # NUEVO: TestAyudasDeLosArgumentos (SC-001)
└── testdata/
    ├── plazos/describe.json                # regenerado [datos] (FR-004)
    └── script/h20-fecha-main.txtar         # sección main-articulos.describe [datos] (FR-004)
schemas/{norma,bloque,municipio,instalacion,grafo,cita,plazo}.json   # regenerados [datos] (FR-004)
laboratorio/                    # NUEVO (contracts/laboratorio.md §1)
├── variantes.yaml
├── evals/0{1..4}-*.yaml
└── variantes/<variante>/…
internal/evals/
├── laboratorio.go              # declaración, variantes y partes, regla testigo (FR-011, FR-013, FR-020 a FR-028)
├── laboratorio_sondeo.go       # sondearElLaboratorio (FR-030 a FR-033)
├── laboratorio_tabla.go        # turnos, recuentos, Wilson, tabla y detalle (FR-040 a FR-043)
├── laboratorio_servidor.go     # el filtro de `laboratorio servidor` (FR-021, FR-024, FR-029)
├── sesiones.go                 # SesionesAEjecutar.Servidor (FR-021, FR-024)
├── job_test.go                 # TestSondeoDelLaboratorio (-tags evals) y sus banderas (FR-032)
├── laboratorio*_test.go        # los tests de §Gates
└── testdata/laboratorio/       # NUEVO [datos]: respuestas/, sondeo/, tabla.md
cmd/laboratorio/main.go         # NUEVO: `laboratorio servidor` y `laboratorio tabla`
scripts/evals-laboratorio.sh    # NUEVO (contracts/arnes.md §1)
scripts/evals-sondeo-llavero.sh # primer argumento `sondeo` o `laboratorio`
Makefile                        # evals-laboratorio y laboratorio-tabla; evals-sondeo con `sondeo` delante
CHANGELOG.md                    # Unreleased (FR-005)
CONTRIBUTING.md                 # «Sondeo local»: el del laboratorio y su tabla
```

**Structure Decision**: la de `docs/ROADMAP.md` §2 y `CLAUDE.md`, con dos piezas fuera del producto y justificadas en
Complexity Tracking: `laboratorio/` (material de un sondeo, como `evidencias/` o `refs/` no son producto) y
`cmd/laboratorio` (un programa del repositorio, como `cmd/empaquetar`).

## Diseño

- **Esquemas** (contracts/esquemas.md): `describirArgumentos(entrada, campos)` pone en la propiedad de cada campo del
  verbo con `help` no vacía su valor; las globales no pasan por ella (research D1). Se regeneran los siete esquemas y
  los dos golden que fijan esos bytes (research D3).
- **Laboratorio** (contracts/laboratorio.md): la declaración, las ocho variantes como copias completas, la regla
  testigo y sus respuestas fijas, y las cuatro evals con las preguntas y lecturas de las 09, 04, 07 y 10 de
  `normativa`.
- **Arnés** (contracts/arnes.md): argumentos y comprobaciones antes de nada; un reparto por variante y modo con
  `ejecutarSesiones`; `Servidor` solo en `instrucciones` y `parametro`; `DIRECTORIO/<variante>/<modo>/<eval>-<nn>/` y
  `sondeo.json`.
- **Guion** (contracts/tabla.md): turnos por respuesta del modelo, recuentos de FR-041, Wilson, tabla y detalle.

**Aceptación e2e**: `aceptacion/describe.txtar` (activado como `internal/app/testdata/script/h39-describe.txtar`)
describe la historia 1: `kitlegal boe articulo BOE-A-2015-10565 a21 --describe` da en `norma`, `bloque` y `fecha` su
ayuda como `description`, carácter a carácter; `kitlegal plazos calcular x x x --describe`, en sus argumentos;
`kitlegal skills install --describe` da en `global` la suya con `$HOME` tal cual; y ninguna de las banderas globales
(`json`, `timeout`…) lleva `description` (FR-001, FR-003; SC-001 en `--describe`). FR-002 lo comprueba
`TestAyudasDeLosArgumentos`, porque la orden `mcp` del arnés e2e no escribe el esquema de las herramientas (research
D12). Las historias 2 y 3 no tienen comportamiento visible del binario `kitlegal`: su aceptación son los tests de
`make ci` de SC-003 a SC-007 (§Gates). La historia 4 la hace la persona antes de fusionar (FR-050 a FR-052).

## Gates: controles mecánicos que añade o toca

- **`make ci`**: sin objetivos nuevos dentro; corre todo lo que sigue. `make schema-check` compara los siete esquemas
  regenerados; `make skills-check` y la deriva de `TestSkillsDelRepositorio` exigen `skills/` sin cambios.
- **Objetivos nuevos, fuera de `make ci`**: `make evals-laboratorio` (persona, con credencial) y
  `make laboratorio-tabla` (sin red ni modelo). `make evals-sondeo` sigue igual para quien lo lanza.
- **Tests nuevos en `make ci`**: `TestAyudaComoDescripcion` (`internal/cli`); `TestAyudasDeLosArgumentos`
  (`internal/app`); y en `internal/evals`: `TestReglaTestigo`, `TestVariantesDelLaboratorio`, `TestEvalsDelLaboratorio`,
  `TestLaboratorioFueraDelProducto`, `TestTablaDelLaboratorio`, `TestArnesDelLaboratorio` y
  `TestServidorDelLaboratorio`; y el caso de `servidor.json` con `Servidor` en el test de hoy de `escribirElServidor`.
- **Tests que cambian porque cambia lo que se fija**: los de `internal/cli` que comparan el esquema de un struct con
  `help`; `TestEsquemasPublicados` y los golden de `plazos` y de `h20-fecha-main.txtar`, regenerados.
- **Fixtures**: `internal/evals/testdata/laboratorio/` (respuestas fijas, sesiones fijas de un sondeo y `tabla.md`; la
  fila de 40 la monta el test en un temporal con copias de dos sesiones fijas de `sondeo/`), escritos por las tareas,
  sin red; ninguna grabación nueva.
- **Fuera de `make ci`**: `TestSondeoDelLaboratorio` (`-tags evals`), la entrada de la persona.

## Datos externos

Ninguno (research «Fuentes externas»): no hay manifiesto `grabaciones.json`, test `TestGrabar*` ni tarea que grabe; las
sesiones del sondeo se sirven de `testdata/evals/`, y nada se escribe en `evidencias/` ni en `data/`.

## Controles de umbral

| Requisito | Umbral | Control | Dónde |
|---|---|---|---|
| SC-001, FR-001, FR-002, FR-003 | 100 % de los argumentos con ayuda de todos los verbos (28 hoy, como mínimo) con esa ayuda como `description`, en `--describe` y en `tools/list`; 0 propiedades sin ayuda, y 0 globales, con `description` | Recorre `RegistroDeProduccion` contra el modelo de Kong y falla con el primer argumento que no cumple | ci:internal/app/ayudas_test.go:TestAyudasDeLosArgumentos |
| SC-001, FR-001, FR-003 | 0 propiedades sin ayuda con `description` y la ayuda sin reescribir, en los dos esquemas | Structs de prueba con y sin ayuda y con `<`, `>` y `$` | ci:internal/cli/describe_test.go:TestAyudaComoDescripcion |
| FR-004 | 0 diferencias entre `schemas/` y lo que emite `--describe` | `make schema-check` | ci:internal/app/esquemas_test.go:TestEsquemasPublicados |
| SC-002, FR-004 | 0 bytes de diferencia en las tablas de comandos de `skills/*/SKILL.md` regeneradas | Deriva de las skills | ci:internal/app/skills_test.go:TestSkillsDelRepositorio |
| FR-029 | Las instrucciones del binario, las de hoy carácter a carácter (485 bytes) | Texto fijo | ci:internal/mcp/instrucciones_test.go:TestInstrucciones |
| SC-003, FR-013, FR-014 | 100 % de las respuestas fijas con su veredicto; la línea del sello en 0 ficheros de `skills/` | Guion de la regla sobre `respuestas/` y la respuesta del producto | ci:internal/evals/laboratorio_test.go:TestReglaTestigo |
| SC-004, FR-011, FR-020 a FR-028 | 0 bytes de diferencia fuera de lo declarado y más de 0 dentro, en las siete variantes; cuerpo de como mucho 11 000 caracteres y descripción de como mucho 900 en cada variante, sin defectos ni deriva de `skills.Regenerar` | Partes frente a la de control y forma del producto | ci:internal/evals/laboratorio_test.go:TestVariantesDelLaboratorio |
| FR-012 | 4 evals, todas de la herramienta declarada; 0 consultas sin servir por la grabación | `LeerConjunto`, `Preparar` y `ComprobarSinRed` | ci:internal/evals/laboratorio_test.go:TestEvalsDelLaboratorio |
| SC-005, FR-010 | 0 ficheros del laboratorio en lo empotrado, en lo que instala `skills install --dir` y en lo que elige o ejecuta el job | Recorre `kitlegal.Skills()`, una instalación en un temporal, la definición del job y `rutaQueMide` | ci:internal/evals/laboratorio_test.go:TestLaboratorioFueraDelProducto |
| SC-006, FR-041, FR-042, FR-043 | 0 diferencias en los recuentos y en los extremos de Wilson con un decimal: 0, 20 y 40 de 40 → `[0,0–8,8 %]`, `[35,2–64,8 %]`, `[91,2–100,0 %]` | La salida, byte a byte, contra `tabla.md` y contra las celdas esperadas | ci:internal/evals/laboratorio_tabla_test.go:TestTablaDelLaboratorio |
| SC-007, FR-030, FR-031 | 0 sesiones abiertas con una petición inválida (versión, modelo, variante en otro modo, `SESIONES` no múltiplo de 4); n/4 por eval en cada variante y modo pedidos con una válida | Sustituto de `claude` que cuenta sus invocaciones | ci:internal/evals/laboratorio_sondeo_test.go:TestArnesDelLaboratorio |
| FR-030, FR-033 | 0 diferencias, carácter a carácter, entre la orden del guion del llavero que imprimen `make -n evals-sondeo …` y `make -n evals-laboratorio …` y la de contracts/arnes.md §1 (`sondeo` y sus cinco argumentos; `laboratorio` y sus siete, en su orden) | `make -n` desde la raíz del repositorio, por os/exec y sin el `MAKEFLAGS` del `make` que lo lanza: solo imprime, no abre nada | ci:internal/evals/laboratorio_sondeo_test.go:TestRecetasDelSondeo |
| FR-021, FR-024, FR-029 | 0 bytes cambiados fuera de `instructions` y de la `description` del parámetro declarado | El filtro sobre las líneas del servidor de producción | ci:internal/evals/laboratorio_servidor_test.go:TestServidorDelLaboratorio |
| SC-008, FR-053 | Las evals de `normativa` en verde en el cierre: sus umbrales con `decide: true` | El job de evals del cierre | evals:normativa:sin_activar:claude-sonnet-5-5:orden, evals:normativa:sin_activar:claude-sonnet-5-5:herramienta, evals:normativa:redaccion_sin_fechar:claude-sonnet-5-5:herramienta, evals:normativa:afirma_lo_no_leido:claude-sonnet-5-5 |
| SC-008, FR-053 | Las evals de `consulta-juridica` en verde en el cierre: su umbral con `decide: true` y sus series (ADR 0016) | El job de evals del cierre | evals:consulta-juridica:habla_en_primera_o_segunda_persona |
| SC-008, FR-053 | Las evals de `jurisprudencia` en verde en el cierre | El job de evals del cierre | evals:jurisprudencia:sin_activar:claude-sonnet-5-5:herramienta, evals:jurisprudencia:afirma_lo_no_leido:claude-sonnet-5-5 |
| SC-008, FR-053 | Las evals de `plazos-y-territorio` en verde en el cierre | El job de evals del cierre | evals:plazos-y-territorio:sin_activar:claude-sonnet-5-5:herramienta, evals:plazos-y-territorio:plazo_sin_fecha:claude-sonnet-5-5:herramienta |

SC-008 usa los umbrales del job tal como están (`.github/workflows/evals.yml` y las carpetas de evals), que el hito no
toca: las filas nombran los que publica el informe de cada skill con `decide: true` (comprobado en los informes de H32,
`specs/030-h32-arquitectura-de-skills/gates/evals/`), y el veredicto del job sale en rojo con cualquiera de ellos. SC-009
lo comprueba la persona (capa 3).

## Cierre y pendientes

- **FR-052**: el run no lanza el sondeo ni escribe el ADR; la propuesta de cambio queda pendiente de los dos, y el
  informe final lo dice por `gates/supuestos.md`.
- **FR-053**: si la medición del cierre da los rojos de partida en modo herramienta (eval 04 de `normativa`,
  `sin_activar`; eval 02 de `consulta-juridica`), la reparación no toca `skills/` ni reorganiza las skills del
  producto, y el informe da sus cifras para que decida una persona (anotado en `gates/supuestos.md`).

## Complexity Tracking

| Desviación | Por qué hace falta | Alternativa más simple rechazada y por qué |
|---|---|---|
| `laboratorio/` en la raíz del repositorio | FR-010: fuera de `skills/` y `evals/`, y fuera de lo que mide el job, para que un cambio del laboratorio no mida el producto (research D4, V9) | `internal/evals/laboratorio/` o su `testdata/`: es de lo común, y cada cambio mediría las cuatro skills del producto |
| `cmd/laboratorio`, un segundo programa del repositorio | FR-021, FR-024 y FR-029: el servidor de esas sesiones no puede ser `kitlegal`; y FR-040: la tabla en la salida estándar, sin el registro de `go test` | Una bandera o variable oculta en `kitlegal`, que prohíbe FR-029; reconstruir `kitlegal` con `-overlay` y el fuente reescrito (research D7) |
| `SesionesAEjecutar.Servidor` | FR-021 y FR-024: otra orden en `servidor.json` sin sacar `kitlegal` del lugar de `Binario`, cuyo directorio sale del `PATH` en el modo herramienta | Dar el filtro como `Binario`: dejaría `kitlegal` en el `PATH` de esas sesiones |
| Primer argumento de `scripts/evals-sondeo-llavero.sh` | FR-032: el arnés con la credencial del sondeo de hoy (ADR 0032) | Un segundo guion de llavero: duplicaría la parte que toca la credencial |
