# Implementation Plan: H41.2 · Por qué la eval 04 no activa `normativa` en modo herramienta en el job

**Branch**: `036-h41-2-por-qu-e` | **Date**: 2026-11-11 | **Spec**: [spec.md](spec.md)

**Input**: Feature specification from `specs/036-h41-2-por-qu-e/spec.md`

## Summary

El hito entrega lo que hace falta para medir de dónde viene la caída de la eval 04 de `normativa` en modo herramienta, sin
tocar ninguna skill, eval, umbral ni versión del job. El informe del job da, por sesión, la lista de sus activaciones de
skills —nombre tal cual y si van antes o después de la primera herramienta de kitlegal— en `informe.json` y en la sección
de la sesión de `informe.md`, con todo lo demás igual (FR-001 a FR-006). El arnés del laboratorio gana dos argumentos de
`make evals-laboratorio`: `INSTALACION=producto`, que deja en cada sesión las cuatro skills de `skills/` con la variante en
el lugar de la de su familia (FR-010 a FR-012), y `EVAL=<nn>`, que pide una sola eval del plan de la familia (FR-013);
cada sondeo de una familia del producto deja en `sondeo.json` los nombres de las skills que tenía instaladas. La tabla da
por fila esas skills y cuenta las sesiones que activan otra skill y las que no activan ninguna, con sus intervalos de
Wilson, y el detalle da la lista de cada sesión (FR-020 a FR-024). El run no abre sesiones con modelo: el sondeo de
cuatro filas y 160 sesiones, su tabla en `evidencias/` y el ADR los hace la persona antes de fusionar (FR-030, FR-031).

Enfoque técnico (research D1 a D17): una sola lectura de los usos con su turno y una sola regla de la primera herramienta
para el informe y para la tabla, en `internal/evals/activaciones.go`; la lista en `informe.json` con `omitzero`, para que
«sin leer» no se confunda con «ninguna»; la instalación, con la misma copia con `os.CopyFS` que H41 usa para
`<skill>-producto`; la constancia, en `sondeo.json`; y la tabla nueva solo cuando algún sondeo tiene esa constancia, para
que la de los sondeos de antes salga igual byte a byte.

## Technical Context

**Language/Version**: Go 1.27 (`go 1.27.0`, `toolchain go1.27.3` en `go.mod`).

**Primary Dependencies**: las fijadas, sin ninguna nueva. Biblioteca estándar en el código (`encoding/json`,
`encoding/json/v2` y `encoding/json/jsontext` donde ya se usan, `os`, `path/filepath`, `slices`, `strings`) y lo que ya usan
`internal/evals` e `internal/app`; `stretchr/testify` y `rogpeppe/go-internal` (testscript) en los tests.

**Storage**: ficheros: `informe.json` e `informe.md` del job; `sondeo.json` y las sesiones de cada sondeo, en un
`DIRECTORIO` de la persona fuera del repositorio; fixtures en `internal/evals/testdata/`.

**Testing**: `go test -race` en `make ci`; un guion testscript congelado; tests de tabla con `t.Parallel()` y fixtures en
`testdata/`; el sustituto de `claude` de `TestArnesDelLaboratorio`; ningún test con red ni con modelo. Skills aplicadas,
con el reparto de `golang-how-to`: `golang-testing` (tests de tabla, fixtures en `testdata/`, testscript; las aserciones
de lo que no cambia, antes de cambiar el código), `golang-refactoring` (mover dentro del paquete lo que pasa a ser
compartido, sin cambiar su comportamiento, con los tests de hoy de red), `golang-project-layout` y `golang-naming`
(ningún paquete nuevo; el fichero nuevo se llama por lo que declara, `activaciones.go`), `golang-error-handling` (los
errores de los argumentos, juntos con `errors.Join` como hoy; ningún error silenciado), `golang-lint` (ningún `//nolint`)
y `golang-cli` (dos argumentos de una orden de `make` del repositorio, validados antes de hacer nada; ninguna bandera del
binario).

**Target Platform**: macOS y Linux, donde corren `make ci`, el job de evals y el sondeo de la persona.

**Project Type**: programas y tests del repositorio (`internal/evals`, `cmd/laboratorio`, guiones de `scripts/`); el
binario `kitlegal` y las skills no cambian.

**Performance Goals**: ninguno. El informe lee una vez más el transcript de cada sesión (research D1).

**Constraints**: las suites congeladas de H39, H40 y H41 siguen en verde sin cambiar un fichero (FR-032); el run
no graba, no escribe en `evidencias/` y no abre sesiones con modelo (FR-014, FR-031).

**Scale/Scope**: 1 fichero nuevo de código y su test; 2 argumentos nuevos; 1 clave nueva en `informe.json` y 1 en
`sondeo.json`; 4 columnas nuevas en la tabla de las familias del producto y 1 en su detalle; 2 juegos de fixtures; 1 guion
congelado; 4 filas y 160 sesiones en la aceptación de la persona.

## Constitution Check

*GATE: Must pass before Phase 0 research. Re-check after Phase 1 design.*

| Principio | Cómo lo cumple este hito |
|---|---|
| I · Fuentes públicas y frontera humana (INNEGOCIABLE) | No toca ninguna fuente ni `internal/httpx` y no añade ninguna petición. Las sesiones del sondeo se sirven de las grabaciones del job, con los proxies que rechazan toda petición, como en H39 a H41. Nada se presenta ni se firma; el CENDOJ no se toca |
| II · Nada sin cita ni fuente | El sobre de ningún applet cambia y ninguna skill cambia: sus reglas de cita, sus formas fijas y su declaración de lo no cubierto siguen como están |
| III · Tests primero y offline | Primero el guion de contracts/aceptacion.md, que hoy falla por una aserción. Todos los tests nuevos, offline y sin modelo, contra fixtures que escriben tareas `[datos]`; ninguna grabación. Las aserciones de lo que no cambia del informe se escriben y pasan antes de tocar `informe.go`. `internal/core` no cambia y la cobertura global no baja: el código nuevo lleva sus tests |
| IV · Arquitectura hexagonal con reglas ejecutables | El código va en `internal/evals`, que no va en el binario; `cmd/laboratorio` sigue siendo la raíz que nombra `os.Std*` y llama a `os.Exit`. Errores: los de los argumentos del arnés, `errorDeUso` juntos como hoy (el guion sale con 1); los de la tabla, defecto `inesperado` con 1; en el informe, motivo de sesión ilegible. `kitlegal` y sus códigos no cambian. Ningún `panic` |
| V · Simplicidad y dependencias fijadas (YAGNI) | Ninguna dependencia nueva. Cada mecanismo se traza a un FR («Trazabilidad»). Lo que llega manipulado —un transcript que Claude Code no escribe, un `sondeo.json` editado a mano— es la regla genérica del spec, sin comprobación, mensaje ni test propios (research D1, D5). Las desviaciones, en Complexity Tracking |
| VI · Un binario, convenciones de agente | Un solo binario `kitlegal`, sin banderas, variables ni órdenes nuevas. Los dos argumentos son de `make evals-laboratorio`, una orden del repositorio; `cmd/laboratorio` no se distribuye y no cambia de órdenes |
| VII · Grafo y privacidad | No toca el grafo. `sondeo.json` gana nombres de skills, ningún dato de una persona; la lista del informe, nombres de skills |
| VIII · Skills primero; el binario es la herramienta | Protege las skills que ya existen: mide de dónde viene una caída de `sin_activar` de `normativa` antes de cambiar la skill, su umbral o el job (spec, «Uso, de fuera adentro»). Ninguna skill cambia ni gana `scripts/`; sus evals se miden en el cierre (FR-033) |
| IX · Genericidad territorial, validación local | Ningún dato ni código nombra un municipio; las evals del sondeo son las de hoy |
| X · Genericidad por materia, validación administrativa | Ningún caso especial para una materia ni una skill: la lista da el nombre de cualquier skill tal cual; `INSTALACION=producto` copia todas las de `skills/`, sean las que sean; las columnas de la tabla cuentan cualquier otra skill por su nombre; `EVAL` vale para cualquier eval del plan de cualquier familia. La eval 04 solo está en las órdenes de la persona |
| Regla 1 · `internal/core/**` no importa adaptadores | `internal/core` no cambia |
| Regla 2 · solo `internal/httpx` importa `net/http` | Nada nuevo importa `net/http` |
| Regla 3 · solo `internal/{cache,store,graph}` importan SQLite | Nada nuevo importa `database/sql` ni `modernc.org/sqlite` |
| Regla 4 · solo `internal/cli` y `cmd/` llaman a `os.Exit` | Solo `cmd/laboratorio/main.go`, como hoy |
| Regla 5 · solo `internal/render` escribe en stdout; logs a stderr | `internal/evals` escribe en los ficheros del informe y en el `io.Writer` que le inyecta `cmd/laboratorio`, como hoy; `activaciones.go` no escribe nada |
| Regla 6 · `internal/graph` no importa `internal/source/*` ni `internal/render` | `internal/graph` no cambia |
| Regla 7 · solo `internal/mcp` importa el SDK de MCP | Nada nuevo lo importa: el servidor de las sesiones es el `kitlegal mcp serve` de hoy |
| Garantías sin red de H19 y H23 (`TestArquitectura`) | Ninguna importación cambia en lo que instala las skills ni en lo que prepara y coteja una cita |

Sin violaciones sin justificar. Re-evaluado tras el diseño de la fase 1: igual.

## Project Structure

### Documentation (this feature)

```text
specs/036-h41-2-por-qu-e/
├── plan.md
├── research.md
├── data-model.md
├── quickstart.md
├── contracts/
│   ├── informe.md         # FR-001 a FR-006: la lista en informe.json e informe.md
│   ├── arnes.md           # FR-010 a FR-015: INSTALACION, EVAL y skills_instaladas
│   ├── tabla.md           # FR-020 a FR-024: las columnas nuevas y su ejemplo
│   └── aceptacion.md      # el guion congelado
├── aceptacion/laboratorio-tabla.txtar   # la primera tarea; congelado
└── tasks.md
```

### Source Code (repository root)

```text
internal/evals/activaciones.go                 # NUEVO: Activacion, activacionesDeLosUsos, textoDeLasActivaciones, reconocedorDeKitlegal; leerLosUsos y sus ayudantes, movidos de laboratorio_tabla.go
internal/evals/activaciones_test.go            # NUEVO: TestActivacionesDeLosUsos
internal/evals/informe.go                      # ResultadoDeEval.Activaciones; la lista de cada sesión; el párrafo de informe.md
internal/evals/informe_test.go                 # + TestInformeConLasActivaciones; TestInforme (caso sesion-ilegible); TestInformeDeLaSuiteCompletaNoCambia (quita la lista)
internal/evals/laboratorio_sondeo.go           # INSTALACION y EVAL; la copia de las otras skills; el plan de una eval; skills_instaladas
internal/evals/laboratorio_sondeo_test.go      # + TestArnesConLaInstalacionDelProducto, TestArnesConUnaEval; exigirElSondeoDelArnes con skills_instaladas; TestRecetasDelSondeo con nueve argumentos
internal/evals/laboratorio_tabla.go            # sondeoLeido.SkillsInstaladas; la instalación en la fila; las columnas nuevas; el reconocedor compartido
internal/evals/laboratorio_tabla_test.go       # + TestTablaConLasActivaciones
internal/evals/job_test.go                     # TestSondeoDelLaboratorio (-tags evals): -instalacion y -eval
internal/evals/testdata/sesiones/activaciones/                       # NUEVO [datos]: evals/ y diez sesiones
internal/evals/testdata/laboratorio/familias/activaciones/           # NUEVO [datos]: sondeo/{h41,h41-2-a,h41-2-b}/ y tabla.md
scripts/evals-laboratorio.sh                   # nueve argumentos; -instalacion y -eval
scripts/evals-sondeo-llavero.sh                # comentario: nueve argumentos
Makefile                                       # receta de evals-laboratorio con "$(INSTALACION)" "$(EVAL)"; su comentario («siete argumentos»)
CONTRIBUTING.md                                # «Job de evals»: la lista; «El sondeo del laboratorio»: los dos argumentos y las columnas
CHANGELOG.md                                   # Unreleased
```

**Structure Decision**: la de `docs/ROADMAP.md` §2 y `CLAUDE.md`. El código, con el del informe, el arnés y la tabla en
`internal/evals`, que no va en el binario; lo compartido por el informe y la tabla, en un fichero propio del paquete;
ningún paquete nuevo. `skills/`, `evals/`, `laboratorio/` y `.github/workflows/evals.yml` no se tocan.

## Diseño

- **La lectura compartida** (contracts/informe.md §1; data-model §1; research D1, D2): `leerLosUsos` y la regla de la
  primera herramienta pasan a `activaciones.go` y las usan el informe y la tabla; `leerLosUsos` deja de exigir
  `message.id`; en el modo de una eval sin binario ni servidor no hay primera herramienta.
- **El informe** (contracts/informe.md §2, §3; data-model §2; research D3 a D6): `activaciones` en `informe.json`,
  `Skills activadas:` en `informe.md`; nada más cambia. Orden de trabajo: las aserciones de FR-004 de
  `TestInformeConLasActivaciones` sobre las fixtures se escriben y pasan con el `informe.go` de hoy, y después se añade
  la lista; así los valores de «hoy» no salen del código nuevo.
- **El arnés** (contracts/arnes.md; data-model §3; research D7 a D11): dos argumentos detrás de los siete de hoy; la
  copia de las otras skills con la opción; el plan de una eval; `skills_instaladas` en todo sondeo de una familia del
  producto.
- **La tabla** (contracts/tabla.md; data-model §4; research D12, D13): con la constancia, la instalación en la fila y
  las columnas «Skills instaladas», «Activan otra skill», «Otras skills», «No activan ninguna» y, en el detalle,
  «Activaciones»; sin ella, la de hoy.

**Aceptación e2e:** un guion testscript que escribe la primera tarea en `specs/036-h41-2-por-qu-e/aceptacion/` y queda
congelado (contracts/aceptacion.md), activado como `internal/app/testdata/script/h41-2-laboratorio-tabla.txtar` y
ejecutado por `TestEntregaDelHito`:
- `aceptacion/laboratorio-tabla.txtar` (historia 3): `laboratorio tabla` sobre tres sondeos fijos de
  `normativa-producto` en modo herramienta —dos instalaciones de la misma versión y un sondeo de H41 sin constancia—
  da su tabla y su detalle byte a byte: una fila por instalación, las skills instaladas, «Activan otra skill», «Otras
  skills» y «No activan ninguna» con sus intervalos, una fila sin sesiones terminadas, sesiones sin terminar, y la lista
  de cada sesión. Cubre FR-020, FR-021, FR-022, FR-024; SC-005.

La historia 1 no tiene una orden del binario que un guion pueda lanzar —el informe lo escribe `EscribirInforme` dentro
de `go test`—: su aceptación en `make ci` es `TestInformeConLasActivaciones` (FR-001 a FR-006; SC-001, SC-002). La 2 abre
sesiones con modelo y credencial: su aceptación en `make ci` son `TestArnesConLaInstalacionDelProducto` y
`TestArnesConUnaEval`, con el sustituto de `claude` (FR-010 a FR-015; SC-003, SC-004, SC-009). La 4 la hace la persona
(FR-030, FR-031; quickstart.md §8).

## Trazabilidad

| Mecanismo | FR/SC |
|---|---|
| `activaciones.go`: `Activacion`, `activacionesDeLosUsos`, `reconocedorDeKitlegal`, `leerLosUsos` movida | FR-001, FR-020 (la misma definición en el informe y en la tabla) |
| `leerLosUsos` sin exigir `message.id` | FR-004, FR-005 (ninguna sesión legible pasa a ilegible) |
| Ninguna primera herramienta en el modo sin binario ni servidor | FR-001 («en una sesión sin ninguna, toda activación es antes») |
| `textoDeLasActivaciones` | FR-002, FR-020 (el mismo texto en `informe.md` y en el detalle) |
| `ResultadoDeEval.Activaciones` con `omitzero` | FR-001, FR-003, FR-005 |
| Párrafo `Skills activadas:` | FR-002, FR-005 |
| Error al leer la lista → motivo de sesión ilegible | FR-005 |
| `TestActivacionesDeLosUsos` | FR-001, SC-001 |
| `TestInformeConLasActivaciones` y `testdata/sesiones/activaciones/` | FR-006, SC-001, SC-002 |
| `TestInforme` (caso `sesion-ilegible`) | FR-005 |
| `TestInformeDeLaSuiteCompletaNoCambia` quita la lista | FR-004, SC-002 |
| `INSTALACION=producto`, sus errores y la copia de las otras skills | FR-010, FR-011 |
| `EVAL=<nn>`, sus errores, `SESIONES` sin múltiplo de 4 y el plan de una eval | FR-011, FR-013 |
| Receta, `scripts/evals-laboratorio.sh`, `-instalacion`/`-eval` en `TestSondeoDelLaboratorio`, comentario del llavero | FR-010, FR-013, FR-014 |
| `skills_instaladas` en `sondeo.json` y en `sondeoLeido` | FR-012, FR-020 |
| `TestArnesConLaInstalacionDelProducto`, `TestArnesConUnaEval`; `exigirElSondeoDelArnes` con la clave | FR-015, SC-003, SC-004, SC-009 |
| `TestRecetasDelSondeo` con nueve argumentos | FR-010, FR-013, FR-014 |
| La instalación en `claveDeFila`; la de la familia sola sin constancia | FR-020, FR-022 |
| `conLasActivaciones`: columnas nuevas solo con constancia | FR-020, FR-022, FR-023 |
| Columnas «Activan otra skill», «Otras skills», «No activan ninguna», «Activaciones» | FR-020, FR-021 |
| `TestTablaConLasActivaciones` y `testdata/laboratorio/familias/activaciones/` | FR-024, SC-005 |
| Guion congelado | FR-020, FR-021, FR-022, FR-024, SC-005 |
| `CONTRIBUTING.md`, comentario de `evals-laboratorio` en el `Makefile`, cabeceras de los guiones, `CHANGELOG.md` | FR-001, FR-010, FR-013, FR-020: lo que hoy documentan deja de ser cierto (research D16) |
| Línea de `gates/supuestos.md` del sondeo pendiente | FR-031 (research D17) |

## Gates: controles mecánicos que añade o toca

- **`make ci`**: ningún objetivo nuevo; corre todo lo que sigue.
- **Objetivos de `make`**: `evals-laboratorio` gana dos argumentos en su receta (`TestRecetasDelSondeo`); cambia su
  comentario. `laboratorio-tabla`, sin cambios.
- **Tests nuevos en `make ci`**: `TestActivacionesDeLosUsos` (`activaciones_test.go`); `TestInformeConLasActivaciones`
  (`informe_test.go`); `TestArnesConLaInstalacionDelProducto` y `TestArnesConUnaEval` (`laboratorio_sondeo_test.go`);
  `TestTablaConLasActivaciones` (`laboratorio_tabla_test.go`).
- **Tests que cambian**: `TestInformeDeLaSuiteCompletaNoCambia` quita la lista antes de comparar con sus dos constantes,
  que no cambian, y gana los casos de lo que se quita; `TestInforme`, su caso `sesion-ilegible` exige además que no haya
  lista; `exigirElSondeoDelArnes` (de `TestArnesDeLasFamilias` y `TestArnesDeNormativa`) espera `skills_instaladas` en
  `sondeo.json` de una familia del producto; `TestRecetasDelSondeo`, nueve argumentos. Sin cambiar ningún valor
  esperado: los demás tests del informe, de `LeerSesion`, de la serie y de los umbrales, `TestJuzgarSinLaLista`,
  `TestArnesDelLaboratorio`, lo que exigen de cada sesión `TestArnesDeLasFamilias` y `TestArnesDeNormativa`,
  `TestTablaDeLasFamilias`, `TestTablaDeNormativa`, `TestTablaDelLaboratorio`, `TestLecturaConCd`, `TestReglaTestigo`,
  `TestOrdenesDeBash`, `TestServidorDelLaboratorio`, `TestLaboratorioFueraDelProducto`.
- **Suites congeladas**: la de H41.2 (contracts/aceptacion.md); las de H39, H40 y H41, sin cambiar un fichero
  (FR-032).
- **Fixtures** (tareas `[datos]`, sin red): `internal/evals/testdata/sesiones/activaciones/` (una eval y diez sesiones,
  contracts/informe.md §5) e `internal/evals/testdata/laboratorio/familias/activaciones/` (tres sondeos y su `tabla.md`,
  contracts/tabla.md §6). Ninguna grabación; ningún fichero de `testdata/` existente cambia.
- **Fuera de `make ci`**: `TestSondeoDelLaboratorio` (`-tags evals`), la entrada de la persona.

## Datos externos

Ninguno (research, «Fuentes externas»): no hay manifiesto `grabaciones.json`, test `TestGrabar*` ni tarea que grabe; las
sesiones del sondeo se sirven de las grabaciones que ya usa el job, y el run no escribe en `evidencias/` ni en `data/`.

## Controles de umbral

| Requisito | Umbral | Control | Dónde |
|---|---|---|---|
| SC-001, FR-001, FR-002, FR-003, FR-006 | 100 % de las sesiones fijas de los dos modos con su lista exacta —nombres, orden y «antes»/«después»— en `informe.json` y en `informe.md`, también la de `Skill` en el mismo turno; y cada caso de la regla con su lista | Las diez sesiones de contracts/informe.md §5, sesión a sesión; los usos escritos en el test, caso a caso | ci:internal/evals/informe_test.go:TestInformeConLasActivaciones, ci:internal/evals/activaciones_test.go:TestActivacionesDeLosUsos |
| SC-002, FR-004 | 0 diferencias en `activada`, `sin_activar` de cada modo, tasas, motivos y veredicto de esas sesiones respecto de hoy; 0 bytes de diferencia en el informe de la suite completa fuera de la lista | Los valores de hoy, fijados antes de cambiar el código; el informe de la suite completa sin la lista contra sus dos constantes | ci:internal/evals/informe_test.go:TestInformeConLasActivaciones, ci:internal/evals/informe_test.go:TestInformeDeLaSuiteCompletaNoCambia |
| SC-003, FR-010, FR-012, FR-015 | Sin la opción, 0 bytes de diferencia con lo que se instala hoy; con ella, 4 skills por sesión, 0 repetidas y 0 bytes de diferencia con la variante y con `skills/`, en una variante versionada y en la `-producto` de las tres familias; la constancia en `sondeo.json` | El sustituto de `claude`: las skills de cada sesión, ruta a ruta y byte a byte, y `sondeo.json` | ci:internal/evals/laboratorio_sondeo_test.go:TestArnesConLaInstalacionDelProducto, ci:internal/evals/laboratorio_sondeo_test.go:TestArnesDeLasFamilias, ci:internal/evals/laboratorio_sondeo_test.go:TestArnesDeNormativa, ci:internal/evals/laboratorio_sondeo_test.go:TestArnesDelLaboratorio |
| SC-004, FR-011 | 0 sesiones abiertas, nada construido y `DIRECTORIO` sin crear con cada petición inválida de `INSTALACION` y de `EVAL`, y con las de hoy | Las peticiones de contracts/arnes.md §5, con su mensaje | ci:internal/evals/laboratorio_sondeo_test.go:TestArnesConLaInstalacionDelProducto, ci:internal/evals/laboratorio_sondeo_test.go:TestArnesConUnaEval, ci:internal/evals/laboratorio_sondeo_test.go:TestArnesDeLasFamilias |
| SC-009, FR-013 | Con `EVAL`, el 100 % de las sesiones de la eval pedida y tantas como `SESIONES`, en una variante de cada familia y en las dos invocaciones de FR-030; sin él, 0 diferencias con las de hoy | Las sesiones que anota el sustituto de `claude` | ci:internal/evals/laboratorio_sondeo_test.go:TestArnesConUnaEval, ci:internal/evals/laboratorio_sondeo_test.go:TestArnesDeLasFamilias |
| SC-005, FR-020, FR-021, FR-024 | 0 diferencias en filas, recuentos, extremos de Wilson con un decimal, nombres y listas del detalle de las sesiones fijas, también con 0, 1 y 40 de 40 | La salida, byte a byte, contra `tabla.md`; la fila de 40 copias; el guion congelado | ci:internal/evals/laboratorio_tabla_test.go:TestTablaConLasActivaciones, ci:internal/app/e2e_test.go:TestEntregaDelHito |
| SC-006, FR-022, FR-023, FR-032 | 0 bytes de diferencia en la tabla y el detalle de los sondeos fijos de H39, H40 y H41; 0 fallos en las suites congeladas de H39, H40 y H41 sin cambiar un fichero suyo | Sus tablas fijas; sus guiones activados | ci:internal/app/e2e_test.go:TestEntregaDelHito, ci:internal/evals/laboratorio_tabla_test.go:TestTablaDeLasFamilias, ci:internal/evals/laboratorio_tabla_test.go:TestTablaDeNormativa, ci:internal/evals/laboratorio_tabla_test.go:TestTablaDelLaboratorio |
| SC-007, FR-033 | Las evals de `normativa` en verde en el cierre, en los dos modos, con `sin_activar` ≤ 0: sus umbrales con `decide: true` | El job de evals del cierre | evals:normativa:sin_activar:claude-sonnet-5-5:orden, evals:normativa:sin_activar:claude-sonnet-5-5:herramienta, evals:normativa:redaccion_sin_fechar:claude-sonnet-5-5:orden, evals:normativa:redaccion_sin_fechar:claude-sonnet-5-5:herramienta, evals:normativa:afirma_lo_no_leido:claude-sonnet-5-5, evals:normativa:medida_del_juez:afirma_lo_no_leido:defectos_sin_marcar, evals:normativa:medida_del_juez:afirma_lo_no_leido:correctos_marcados, evals:normativa:duracion_de_las_sesiones:orden, evals:normativa:duracion_de_las_sesiones:herramienta, evals:normativa:duracion_del_juez:orden, evals:normativa:duracion_del_juez:herramienta |
| SC-007, FR-033 | Las evals de `jurisprudencia` en verde en el cierre, en los dos modos | El job de evals del cierre | evals:jurisprudencia:sin_activar:claude-sonnet-5-5:orden, evals:jurisprudencia:sin_activar:claude-sonnet-5-5:herramienta, evals:jurisprudencia:cita_sin_documento:claude-sonnet-5-5:orden, evals:jurisprudencia:cita_sin_documento:claude-sonnet-5-5:herramienta, evals:jurisprudencia:documento_sin_cotejar:claude-sonnet-5-5:orden, evals:jurisprudencia:documento_sin_cotejar:claude-sonnet-5-5:herramienta, evals:jurisprudencia:afirma_lo_no_leido:claude-sonnet-5-5, evals:jurisprudencia:medida_del_juez:afirma_lo_no_leido:defectos_sin_marcar, evals:jurisprudencia:medida_del_juez:afirma_lo_no_leido:correctos_marcados, evals:jurisprudencia:duracion_del_juez:orden, evals:jurisprudencia:duracion_del_juez:herramienta |
| SC-007, FR-033 | Las evals de `plazos-y-territorio` en verde en el cierre, en los dos modos | El job de evals del cierre | evals:plazos-y-territorio:sin_activar:claude-sonnet-5-5:orden, evals:plazos-y-territorio:sin_activar:claude-sonnet-5-5:herramienta, evals:plazos-y-territorio:plazo_sin_fecha:claude-sonnet-5-5:orden, evals:plazos-y-territorio:plazo_sin_fecha:claude-sonnet-5-5:herramienta, evals:plazos-y-territorio:cobertura_sin_declarar:claude-sonnet-5-5:orden, evals:plazos-y-territorio:cobertura_sin_declarar:claude-sonnet-5-5:herramienta, evals:plazos-y-territorio:vigencia_sin_declarar:claude-sonnet-5-5:orden, evals:plazos-y-territorio:vigencia_sin_declarar:claude-sonnet-5-5:herramienta, evals:plazos-y-territorio:redaccion_sin_fechar:claude-sonnet-5-5:orden, evals:plazos-y-territorio:redaccion_sin_fechar:claude-sonnet-5-5:herramienta, evals:plazos-y-territorio:lectura_sin_intentar:claude-sonnet-5-5:orden, evals:plazos-y-territorio:lectura_sin_intentar:claude-sonnet-5-5:herramienta |

SC-007 usa los umbrales del job tal como están (`.github/workflows/evals.yml` y las carpetas de evals), que el hito no
toca: las filas nombran los que el informe final de H41.1 da como publicados con `decide: true` en su cierre (research
V14), y el veredicto del job sale en rojo con cualquiera de ellos. `consulta-juridica` no publica ninguno con
`decide: true` y su skill no cambia; su veredicto lo deciden sus series (ADR 0016), configuración que el hito no toca.
SC-008 lo mide la persona (capa 3): el sondeo abre sesiones con modelo que ni `make ci` ni el run abren.

## Cierre y pendientes

- **FR-031**: el run no lanza el sondeo ni escribe el ADR; `gates/supuestos.md` lo deja dicho, y el informe final lo da
  en «Supuestos y pendientes»: la propuesta queda pendiente de la tabla de las cuatro filas en `evidencias/` y de su
  lectura, que no deciden la fusión.
- **FR-033**: si la medición del cierre da en rojo `sin_activar` de la eval 04 de `normativa` en modo herramienta, la
  reparación no toca ninguna skill, eval, umbral ni versión del job: el informe final da el umbral sin cumplir con sus
  cifras, y las sesiones dicen sus activaciones en el `informe.json` y el `informe.md` del job.

## Complexity Tracking

| Desviación | Por qué hace falta | Alternativa más simple rechazada y por qué |
|---|---|---|
| `leerLosUsos` deja de exigir `message.id` | FR-001 (una sola definición para el informe y la tabla) con FR-004: la exigencia volvería ilegibles sesiones que hoy son legibles, empezando por los transcripts de los tests del informe (research D1, V4) | Mantenerla: cambia motivos y veredictos de hoy, o silencia el error. En todo transcript con `message.id` los turnos son los de hoy, y la tabla de los sondeos del arnés sale igual (FR-022) |
| `TestInformeDeLaSuiteCompletaNoCambia` quita la lista antes de comparar | SC-002: todo lo demás del informe, byte a byte, contra las constantes de hoy | Reescribir las constantes con la lista: el test dejaría de fijar que lo de antes no cambia (research D6) |
| `exigirElSondeoDelArnes` espera `skills_instaladas` en las familias del producto | FR-012 pide la constancia también sin la opción | Escribirla solo con la opción: la tabla no daría la instalación de dos de las cuatro filas de FR-030 (research D11) |
| La tabla de una familia del producto tiene dos formas, con constancia y sin ella | FR-022: sin constancia, la de hoy byte a byte; FR-020: con ella, las columnas nuevas | Dar siempre las columnas nuevas: cambiaría las tablas fijas y las suites congeladas de H40 y H41 (FR-032) |
