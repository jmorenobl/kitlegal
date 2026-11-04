# Implementation Plan: H40 · Por qué `normativa` no lee sus referencias en modo herramienta

**Branch**: `032-h40-por-qu-e` | **Date**: 2026-11-04 | **Spec**: [spec.md](spec.md)

**Input**: Feature specification from `specs/032-h40-por-qu-e/spec.md`

## Summary

El hito entrega lo que hace posible medir cuál de las tres diferencias entre `normativa` y la skill de laboratorio de
H39 explica que, en modo herramienta, `normativa` no abra sus referencias: cuatro variantes de `normativa` fuera del
producto —`control`, idéntica a `skills/normativa`, y una por diferencia con el texto exacto del hito— (FR-001 a
FR-008); el arnés de H39 ampliado para lanzarlas con las evals 04, 05, 15 y 22 de `normativa` y sus lecturas grabadas
(FR-010 a FR-013); y el guion de la tabla, que cuenta por variante y modo la activación, la apertura de cada una de las
ocho referencias con su turno y la regla testigo, con intervalos de Wilson, sin cambiar lo que H39 mide (FR-020 a
FR-030). El run no abre ninguna sesión con modelo: el sondeo de 200 sesiones y el ADR los hace la persona antes de
fusionar (FR-040 a FR-042).

Enfoque técnico (research D1 a D16): las variantes, copias completas en `laboratorio/normativa/<variante>/skills/normativa/`,
con nombres `normativa-…` que no chocan con los de H39; la familia de `normativa` declarada en Go en
`internal/evals/laboratorio_normativa.go`; el arnés y la tabla de H39, en `internal/evals`, eligen la familia por los
nombres de las variantes; `make ci` calcula cada variante desde la de control con su declaración y la compara byte a
byte; y la suite congelada es un guion que ejecuta `laboratorio tabla` sobre las sesiones de H39 de `main` y sobre unas
de `normativa`, para lo que el arnés e2e construye también `cmd/laboratorio`.

## Technical Context

**Language/Version**: Go 1.27 (`go 1.27.0`, `toolchain go1.27.3` en `go.mod`).

**Primary Dependencies**: las fijadas, sin ninguna nueva: `stretchr/testify` y `rogpeppe/go-internal` v1.16.0
(testscript) en los tests; el código nuevo usa la biblioteca estándar y lo que ya usa `internal/evals`. El SDK de MCP
sigue solo en `internal/mcp` (R7).

**Storage**: ficheros: `laboratorio/normativa/` versionado; cada sondeo, en un directorio de la persona; la caché SQLite
de cada sesión, como hoy.

**Testing**: `go test -race` en `make ci`; un guion testscript congelado (`aceptacion/laboratorio-tabla.txtar`); tests de
tabla con `t.Parallel()` y fixtures en `internal/evals/testdata/laboratorio/normativa/`; el sustituto de `claude` de
`TestArnesDelLaboratorio`; ningún test con red ni con modelo. Skills aplicadas: `golang-how-to` con `golang-testing`
(tests de tabla, fixtures en `testdata/`, testscript), `golang-project-layout` (ningún paquete nuevo: el código va donde
está el de H39), `golang-lint` (ningún `//nolint`) y `golang-cli` (ningún argumento ni bandera nuevos).

**Target Platform**: macOS y Linux, donde corren `make ci` y el sondeo de la persona.

**Project Type**: CLI multicall (`kitlegal`, sin cambios) y programas del repositorio (`cmd/laboratorio`, guiones de
`scripts/`).

**Performance Goals**: ninguno. Cada sesión del sondeo, con el tope de hoy.

**Constraints**: `kitlegal`, `skills/`, los esquemas, las instrucciones del servidor y las recetas de `make` no cambian;
el run no graba, no escribe en `evidencias/` y no abre sesiones con modelo (FR-012, FR-013, FR-042); las sesiones del
sondeo, con los proxies que rechazan toda petición (ADR 0043).

**Scale/Scope**: 4 variantes (36 ficheros), 4 evals, 2 lanzamientos, 5 filas y 200 sesiones en la aceptación; una tabla
de 16 columnas.

## Constitution Check

*GATE: Must pass before Phase 0 research. Re-check after Phase 1 design.*

| Principio | Cómo lo cumple este hito |
|---|---|
| I · Fuentes públicas y frontera humana (INNEGOCIABLE) | No toca ninguna fuente ni `internal/httpx` y no añade ninguna petición. Las sesiones del sondeo se sirven de `testdata/evals/` con los proxies que rechazan toda petición, como las del job (ADR 0043). Nada se presenta ni se firma |
| II · Nada sin cita ni fuente | El sobre de ningún applet cambia. Las variantes son `normativa` con una diferencia de redacción cada una: citan igual, y no las ve ninguna persona ni skill del producto |
| III · Tests primero y offline | Primero `aceptacion/laboratorio-tabla.txtar`, que hoy falla por una aserción (la tabla no conoce la familia de `normativa` y sale con 1; contracts/arnes-e2e.md §2). Todos los tests nuevos, offline y sin modelo, contra fixtures escritos por las tareas; ninguna grabación nueva. `internal/core` no cambia; el código nuevo de `internal/evals` lleva sus tests en `make ci` |
| IV · Arquitectura hexagonal con reglas ejecutables | El código va en `internal/evals`, que ningún paquete de producción importa, y `cmd/laboratorio` sigue siendo la raíz que nombra `os.Std*` y llama a `os.Exit`. Errores: la tabla, defecto `inesperado` con 1; el arnés, como en H39 (un error de uso, 1 con su mensaje). `kitlegal` y sus códigos no cambian. Ningún `panic` |
| V · Simplicidad y dependencias fijadas (YAGNI) | Ninguna dependencia nueva. Cada mecanismo se traza a un FR (§Diseño, Complexity Tracking): ni argumentos, ni opciones, ni estados que el spec no pida; lo que llega de fuera manipulado es defecto `inesperado` |
| VI · Un binario, convenciones de agente | Un solo binario `kitlegal`, sin banderas, variables ni órdenes nuevas. `cmd/laboratorio` no se distribuye (`.goreleaser.yaml` construye solo `./cmd/kitlegal`) |
| VII · Grafo y privacidad | No toca el grafo. Las sesiones usan su caché y su grafo en su directorio, como las del job; ningún dato de una persona |
| VIII · Skills primero; el binario es la herramienta | Mide, antes de reorganizar las skills del producto (ADR 0050, punto 5), qué hace que `normativa` lea sus referencias en modo herramienta, y protege las de hoy: `skills/` no cambia y las evals del cierre miden las cuatro (FR-043). Las variantes son skills sin código; el tope de 11 000 sigue en `skills/` (FR-007) |
| IX · Genericidad territorial, validación local | Sin dimensión territorial: ninguna herramienta, dato ni skill con un municipio |
| X · Genericidad por materia, validación administrativa | Ningún caso especial para una materia en código, `data/` ni protocolo de skill. Las cuatro evals del sondeo son las del producto —tributos, haciendas locales y contratación pública— y no cambian; las variantes viven fuera del producto |
| Regla 1 · `internal/core/**` no importa adaptadores | `internal/core` no cambia |
| Regla 2 · solo `internal/httpx` importa `net/http` | Nada nuevo importa `net/http` |
| Regla 3 · solo `internal/{cache,store,graph}` importan SQLite | Nada nuevo importa `database/sql` ni `modernc.org/sqlite` |
| Regla 4 · solo `internal/cli` y `cmd/` llaman a `os.Exit` | Solo `cmd/laboratorio/main.go`, como hoy |
| Regla 5 · solo `internal/render` escribe en stdout; logs a stderr | `internal/evals` escribe en el `io.Writer` que le inyecta `cmd/laboratorio`, como hoy |
| Regla 6 · `internal/graph` no importa `internal/source/*` ni `internal/render` | `internal/graph` no cambia |
| Regla 7 · solo `internal/mcp` importa el SDK de MCP | Nada nuevo lo importa: el servidor de las sesiones es el `kitlegal mcp serve` de hoy |

Sin violaciones. Re-evaluado tras el diseño de la fase 1: igual.

## Project Structure

### Documentation (this feature)

```text
specs/032-h40-por-qu-e/
├── plan.md
├── research.md
├── data-model.md
├── quickstart.md
├── contracts/
│   ├── variantes.md      # FR-001 a FR-008
│   ├── arnes.md          # FR-008, FR-010 a FR-013
│   ├── tabla.md          # FR-020 a FR-023, FR-030
│   └── arnes-e2e.md      # la suite congelada y lo que gana el arnés e2e
├── aceptacion/laboratorio-tabla.txtar   # la primera tarea; congelado
└── tasks.md
```

### Source Code (repository root)

```text
laboratorio/normativa/                         # NUEVO: las cuatro variantes (contracts/variantes.md §1)
└── normativa-{control,remisiones-en-orden,cuerpo-corto,testigo-en-referencia}/skills/normativa/
    ├── SKILL.md
    └── references/*.md                        # las ocho
internal/evals/
├── laboratorio_normativa.go                   # NUEVO: la familia de normativa (data-model §1)
├── laboratorio_normativa_test.go              # NUEVO: TestVariantesDeNormativa, TestReglaTestigoDeNormativa
├── laboratorio_sondeo.go                      # VARIANTES de las dos familias; skills, evals y plan de la familia
├── laboratorio_sondeo_test.go                 # + TestArnesDeNormativa
├── laboratorio_tabla.go                       # familia de la tabla; aperturas por referencia; tabla de normativa
├── laboratorio_tabla_test.go                  # + TestTablaDeNormativa
├── laboratorio_test.go                        # TestLaboratorioFueraDelProducto, generalizado (research D12)
└── testdata/laboratorio/normativa/            # NUEVO [datos]: respuestas/, sondeo/, tabla.md
internal/app/e2e_test.go                       # cmd/laboratorio y KITLEGAL_LABORATORIO_BIN (contracts/arnes-e2e.md §1)
cmd/laboratorio/main.go                        # comentario de paquete: su tabla cuenta las dos familias
Makefile                                       # ayuda de evals-laboratorio y laboratorio-tabla; ninguna receta
CONTRIBUTING.md                                # «El sondeo del laboratorio»: las variantes de normativa
CHANGELOG.md                                   # Unreleased
```

**Structure Decision**: la de `docs/ROADMAP.md` §2 y `CLAUDE.md`. Las variantes, en el laboratorio de H39, fuera del
producto; el código, junto al del arnés y la tabla de H39 en `internal/evals`; ningún paquete nuevo.

## Diseño

- **Variantes** (contracts/variantes.md): copias completas, escritas por las tareas (research D4). `normativa-control` es
  `skills/normativa`; las otras tres, control con su declaración: cinco sustituciones de bytes fijos (D5, D6), una regla
  por secciones con un encabezado renombrado, y una inserción y una sección con el párrafo de `cierre.md` de H39. Cuerpos
  de hoy: 10 929, 11 084, 3298 y 10 973 caracteres (research V1 a V3).
- **Familia** (data-model §1; research D3): la de H39, de `variantes.yaml`, sin cambios; la de `normativa`, en Go.
- **Arnés** (contracts/arnes.md; research D7, D8): `VARIANTES` de una sola familia, con la línea de la mezcla; skills de
  la familia, `evals/normativa` como directorio de evals y el plan de las cuatro; servidor de hoy. Las cinco filas, dos
  lanzamientos.
- **Tabla** (contracts/tabla.md; research D9, D10): la familia por la primera variante; las ocho aperturas por sesión; 16
  columnas con una celda por referencia; la regla testigo solo en `normativa-testigo-en-referencia`; la de H39, byte a
  byte la de hoy.
- **Fuera del producto** (research D12): `TestLaboratorioFueraDelProducto` recorre todo `laboratorio/`, con las copias de
  `skills/` fuera de la comprobación de bytes y los nombres de las variantes de `normativa` entre los prohibidos.

**Aceptación e2e:** `aceptacion/laboratorio-tabla.txtar` (activado como
`internal/app/testdata/script/h40-laboratorio-tabla.txtar`) describe la historia 3: `laboratorio tabla` sobre cinco
sesiones fijas de las variantes de `normativa` da, byte a byte, la tabla y el detalle de contracts/tabla.md §3 y §4
—activación, las ocho referencias con su turno, una abierta dos veces y otras que ninguna abre, `Skill` y la primera
herramienta en el mismo turno, una sin terminar, una fila con `n` = 0, la regla que se cumple y la que no, y la celda
vacía fuera de `normativa-testigo-en-referencia`— (FR-020, FR-021, FR-030, SC-005); y sobre las sesiones fijas de H39
de `main`, su `tabla.md` de `main` (FR-023, SC-007). La primera tarea lo escribe y amplía el arnés e2e con
`KITLEGAL_LABORATORIO_BIN` (contracts/arnes-e2e.md; research D13). Las historias 1 y 2 no tienen una salida de un
programa que un guion lea sin modelo: su aceptación son `TestVariantesDeNormativa`, `TestLaboratorioFueraDelProducto` y
`TestArnesDeNormativa` en `make ci` (FR-001 a FR-011; SC-001 a SC-003, SC-006); los extremos de Wilson de 0, 20 y 40 de
40, `TestTablaDeNormativa` (FR-022). La historia 4 la hace la persona antes de fusionar (FR-040 a FR-042).

## Gates: controles mecánicos que añade o toca

- **`make ci`**: ningún objetivo nuevo; corre todo lo que sigue. `make skills-check` y `TestSkillsDelRepositorio` siguen
  exigiendo `skills/` con su forma y su tope.
- **Objetivos de `make`**: `evals-laboratorio` y `laboratorio-tabla`, con las mismas recetas; cambia solo su ayuda.
  `TestRecetasDelSondeo` sigue fijando las recetas.
- **Tests nuevos en `make ci`**: `TestVariantesDeNormativa` y `TestReglaTestigoDeNormativa`
  (`laboratorio_normativa_test.go`), `TestArnesDeNormativa` (`laboratorio_sondeo_test.go`) y `TestTablaDeNormativa`
  (`laboratorio_tabla_test.go`), en `internal/evals`.
- **Tests que cambian**: `TestLaboratorioFueraDelProducto`, generalizado a las dos familias, sin cambiar lo que comprueba
  de H39 (research D12). `TestArnesDelLaboratorio`, `TestTablaDelLaboratorio`, `TestReglaTestigo`,
  `TestVariantesDelLaboratorio`, `TestEvalsDelLaboratorio` y `TestServidorDelLaboratorio`, sin cambiar ningún valor
  esperado (FR-023).
- **Arnés e2e**: `KITLEGAL_LABORATORIO_BIN` (contracts/arnes-e2e.md §1), que `TestVariablesDeLosGuiones` cuenta entre
  las rutas absolutas; el guion congelado, ejecutado por `TestEntregaDelHito`.
- **Fixtures** (tareas `[datos]`, sin red): `internal/evals/testdata/laboratorio/normativa/respuestas/` (seis respuestas
  fijas), `…/sondeo/` (sesiones fijas de los dos modos) y `…/tabla.md`. Ninguna grabación; ningún fichero de
  `testdata/` existente cambia.
- **Fuera de `make ci`**: `TestSondeoDelLaboratorio` (`-tags evals`), la entrada de la persona.

## Datos externos

Ninguno (research «Fuentes externas»): no hay manifiesto `grabaciones.json`, test `TestGrabar*` ni tarea que grabe; las
sesiones del sondeo se sirven de las grabaciones que ya usa el job, y nada se escribe en `evidencias/` ni en `data/`.

## Controles de umbral

| Requisito | Umbral | Control | Dónde |
|---|---|---|---|
| SC-001, FR-002 | 0 bytes de diferencia y 0 ficheros de más o de menos entre `normativa-control` y `skills/normativa` | Compara los dos árboles y falla con la ruta del primer fichero que difiere, falta o sobra | ci:internal/evals/laboratorio_normativa_test.go:TestVariantesDeNormativa |
| SC-002, FR-003 a FR-006 | 0 bytes de diferencia fuera de lo declarado en cada una de las tres; 100 % de los textos declarados iguales al del hito; 0 declaraciones que ya no se aplican; 8 remisiones, una por referencia | Calcula cada variante desde control con su declaración y la compara byte a byte | ci:internal/evals/laboratorio_normativa_test.go:TestVariantesDeNormativa |
| SC-003, FR-001 | 0 ficheros de una variante en `kitlegal.Skills()`, en lo que instala `skills install --dir` y en lo que el job ejecuta o elige; 0 rutas de `laboratorio/` que midan | Recorre lo empotrado, una instalación en un temporal, la definición del job, `evals/` y `rutaQueMide` | ci:internal/evals/laboratorio_test.go:TestLaboratorioFueraDelProducto |
| FR-007 | Cuerpo de como mucho 11 000 caracteres en cada skill de `skills/` (sin cambios) | La forma de las skills del repositorio | ci:internal/app/skills_test.go:TestSkillsDelRepositorio |
| SC-006, FR-008, FR-010, FR-011 | 0 sesiones abiertas con una variante inexistente, una en un modo que no es suyo, `SESIONES` no múltiplo de 4, una mezcla de familias, otra versión de Claude Code o un alias de modelo; n/4 por eval en cada variante y modo pedidos con una válida; 1 skill, la de su variante, en cada sesión | Sustituto de `claude` que anota sus invocaciones, con los dos lanzamientos de la aceptación (`normativa-control` en orden y herramienta; las otras tres en herramienta) y las tres variantes de solo herramienta en orden, de modo que se piden los ocho pares variante × modo de la familia; en cada sesión, solo `normativa` con el `SKILL.md` de su variante, sin `servidor.json` en orden y con `kitlegal mcp serve` en herramienta (contracts/arnes.md §5) | ci:internal/evals/laboratorio_sondeo_test.go:TestArnesDeNormativa |
| SC-004, FR-030 | 100 % de las respuestas fijas, y la de H32, con su veredicto | El guion de la regla sobre `respuestas/` y la respuesta del informe | ci:internal/evals/laboratorio_normativa_test.go:TestReglaTestigoDeNormativa |
| SC-005, FR-020 a FR-022 | 0 diferencias en recuentos, turnos y extremos de Wilson con un decimal; 0, 20 y 40 de 40 → `[0,0–8,8 %]`, `[35,2–64,8 %]`, `[91,2–100,0 %]` | La salida, byte a byte, contra `tabla.md`, y las celdas de las filas de 40 | ci:internal/evals/laboratorio_tabla_test.go:TestTablaDeNormativa |
| SC-005, SC-007, FR-020, FR-021, FR-023, FR-030 | 0 bytes de diferencia entre la salida de `laboratorio tabla` y las dos tablas del guion congelado (la de `normativa` y la de H39 de `main`) | El guion `h40-laboratorio-tabla.txtar` | ci:internal/app/e2e_test.go:TestEntregaDelHito |
| SC-007, FR-023 | 0 columnas y 0 cifras distintas de `testdata/laboratorio/tabla.md` para las sesiones fijas de H39 | La salida, byte a byte, contra `tabla.md` | ci:internal/evals/laboratorio_tabla_test.go:TestTablaDelLaboratorio |
| SC-008, FR-043 | Las evals de `normativa` en verde en el cierre: sus umbrales con `decide: true` | El job de evals del cierre | evals:normativa:sin_activar:claude-sonnet-5-5:orden, evals:normativa:sin_activar:claude-sonnet-5-5:herramienta, evals:normativa:redaccion_sin_fechar:claude-sonnet-5-5:orden, evals:normativa:redaccion_sin_fechar:claude-sonnet-5-5:herramienta, evals:normativa:afirma_lo_no_leido:claude-sonnet-5-5 |
| SC-008, FR-043 | Las evals de `jurisprudencia` en verde en el cierre | El job de evals del cierre | evals:jurisprudencia:sin_activar:claude-sonnet-5-5:orden, evals:jurisprudencia:sin_activar:claude-sonnet-5-5:herramienta, evals:jurisprudencia:cita_sin_documento:claude-sonnet-5-5:orden, evals:jurisprudencia:cita_sin_documento:claude-sonnet-5-5:herramienta, evals:jurisprudencia:documento_sin_cotejar:claude-sonnet-5-5:orden, evals:jurisprudencia:documento_sin_cotejar:claude-sonnet-5-5:herramienta, evals:jurisprudencia:afirma_lo_no_leido:claude-sonnet-5-5 |
| SC-008, FR-043 | Las evals de `plazos-y-territorio` en verde en el cierre | El job de evals del cierre | evals:plazos-y-territorio:sin_activar:claude-sonnet-5-5:orden, evals:plazos-y-territorio:sin_activar:claude-sonnet-5-5:herramienta, evals:plazos-y-territorio:plazo_sin_fecha:claude-sonnet-5-5:orden, evals:plazos-y-territorio:plazo_sin_fecha:claude-sonnet-5-5:herramienta, evals:plazos-y-territorio:cobertura_sin_declarar:claude-sonnet-5-5:orden, evals:plazos-y-territorio:cobertura_sin_declarar:claude-sonnet-5-5:herramienta, evals:plazos-y-territorio:vigencia_sin_declarar:claude-sonnet-5-5:orden, evals:plazos-y-territorio:vigencia_sin_declarar:claude-sonnet-5-5:herramienta, evals:plazos-y-territorio:redaccion_sin_fechar:claude-sonnet-5-5:orden, evals:plazos-y-territorio:redaccion_sin_fechar:claude-sonnet-5-5:herramienta, evals:plazos-y-territorio:lectura_sin_intentar:claude-sonnet-5-5:orden, evals:plazos-y-territorio:lectura_sin_intentar:claude-sonnet-5-5:herramienta |

SC-008 usa los umbrales del job tal como están (`.github/workflows/evals.yml` y las carpetas de evals), que el hito no
toca: las filas nombran los que los informes del cierre de H39 publican con `decide: true` (research V20), y el veredicto
del job sale en rojo con cualquiera de ellos. `consulta-juridica` no publica ningún umbral con `decide: true`: su
veredicto lo deciden sus series (ADR 0016), configuración del repositorio que el hito no toca, y sale en el cierre como
el de las demás. SC-009 lo comprueba la persona (capa 3).

## Cierre y pendientes

- **FR-042**: el run no lanza el sondeo ni escribe el ADR; la propuesta de cambio queda pendiente de los dos, y el
  informe final lo dice por `gates/supuestos.md`.
- **FR-043**: si la medición del cierre da un rojo de partida en modo herramienta (como la eval 04 de `normativa`,
  `sin_activar`, en la ejecución 41474010633), la reparación no toca `skills/` ni reorganiza una skill del producto, y el
  informe da sus cifras para que decida una persona (anotado en `gates/supuestos.md`).

## Complexity Tracking

| Desviación | Por qué hace falta | Alternativa más simple rechazada y por qué |
|---|---|---|
| `laboratorio/normativa/`, cuatro copias completas de `normativa` (36 ficheros) | FR-001, FR-002 y FR-008: una skill completa por variante, que el arnés copia como directorio de skills, y una copia de control que `make ci` compara con `skills/normativa` | Generarlas al lanzar el sondeo, o enlazarlas: FR-002 no mediría nada, y `os.CopyFS` copia un enlace como enlace (research D4, V10) |
| `cmd/laboratorio` en el arnés e2e (`KITLEGAL_LABORATORIO_BIN`), en la tarea `[aceptacion]` | FR-023 y SC-007: la suite congelada es un guion de `TestEntregaDelHito` que ejecuta `laboratorio tabla`, y tiene que fallar hoy por una aserción | Una orden del arnés que llame a la tabla en proceso (no es el programa que lanza la persona); una precondición y el cambio del arnés en otra tarea (el guion fallaría por no poder ejecutarse; research D13, V17) |
| Copias de `skills/` fuera de la comprobación de bytes de `TestLaboratorioFueraDelProducto` | FR-001 junto a FR-002: `normativa-control` es byte a byte `skills/normativa`, que el binario empotra | Un test aparte para `normativa`: el de H39, que recorre todo `laboratorio/`, seguiría en rojo (research D12, V5) |
