# Implementation Plan: H41 · Las skills del producto mandan leer sus referencias

**Branch**: `033-h41-las-skills-del` | **Date**: 2026-11-06 | **Spec**: [spec.md](spec.md)

**Input**: Feature specification from `specs/033-h41-las-skills-del/spec.md`

## Summary

El hito reescribe cada remisión de `normativa`, `plazos-y-territorio` y `jurisprudencia` como una orden de leer —«lee»,
el enlace, cuándo y lo que trae—, en su sitio (FR-001 a FR-006), y hace el sitio que las órdenes necesitan llevando a
la referencia que se manda leer en su caso lo que solo usa ese caso (FR-009): en `normativa`, a `redaccion_modificada.md`
y a `lecturas_que_fallan.md`, y en `plazos-y-territorio`, a `calculo_de_plazos.md` y a `municipios.md` (correcciones de
T003, T004 y T009 en gates/supuestos.md); los tres cuerpos quedan en 10 926, 10 893 y 10 842 caracteres (FR-008). Amplía el arnés de H39 y H40 a tres familias del producto, cada una con su
`<skill>-control` congelada (la de `v0.9.0`) y su `<skill>-producto` instalada desde `skills/` en cada sondeo, con cuatro
evals de la skill (FR-020 a FR-033); y la tabla, que cuenta las tres familias y la lectura con `cd` al directorio de las
referencias (FR-040, FR-041). El run no abre sesiones con modelo: el sondeo de 520 sesiones y su lectura los hace la
persona antes de fusionar (FR-052).

Enfoque técnico (research D1 a D16): los cambios de las skills, exactos, en contracts/remisiones.md, medidos sobre copias
en esta sesión; la regla de las remisiones como un test del repositorio en `internal/skills`; el producto
calculado desde el control con su declaración, en un test de `internal/evals`, sin lo que genera `make skills-sync`; las
familias del producto declaradas en Go con un campo `producto`; el arnés copia `skills/<skill>` y deja en `sondeo.json`
los nombres de sus referencias; y dos guiones congelados, uno de la tabla y otro de lo que instala el binario.

## Technical Context

**Language/Version**: Go 1.27 (`go 1.27.0`, `toolchain go1.27.3` en `go.mod`; research V8).

**Primary Dependencies**: las fijadas, sin ninguna nueva. Biblioteca estándar en el código (`crypto/sha256`,
`encoding/json`, `os`, `regexp`, `unicode`) y lo que ya usan `internal/skills` e `internal/evals`; `stretchr/testify` y
`rogpeppe/go-internal` v1.16.0 (testscript) en los tests.

**Storage**: ficheros: `skills/` y `laboratorio/` versionados; cada sondeo, en un `DIRECTORIO` de la persona, con su
`sondeo.json` (data-model §4).

**Testing**: `go test -race` en `make ci`; dos guiones testscript congelados; tests de tabla con `t.Parallel()` y fixtures
en `internal/evals/testdata/laboratorio/familias/`; el sustituto de `claude` de `TestArnesDelLaboratorio`; ningún test con
red ni con modelo. Skills aplicadas: `golang-how-to`, con `golang-testing` (tests de tabla, fixtures en `testdata/`,
testscript), `golang-project-layout` (ningún paquete nuevo: el código va con el de H39 y H40 en `internal/evals`, y el
test de las remisiones con los del repositorio de `internal/skills`), `golang-lint` (ningún `//nolint`), `golang-naming` (el fichero de las familias se
llama por lo que declara) y `golang-cli` (ningún argumento ni bandera nuevos).

**Target Platform**: macOS y Linux, donde corren `make ci` y el sondeo de la persona.

**Project Type**: skills sin código del producto, y programas del repositorio (`cmd/laboratorio`, guiones de
`scripts/`); el binario `kitlegal` no cambia, salvo las skills que empotra.

**Performance Goals**: ninguno. Cada sesión del sondeo, con el tope de hoy.

**Constraints**: cuerpo de cada `SKILL.md` ≤ 11 000 caracteres; las suites congeladas (también `h32-normativa.txtar`, que
fija las cinco formas fijas de `normativa` en bloques `text`, research V4) siguen en verde; el run no graba, no escribe
en `evidencias/` y no abre sesiones con modelo (FR-031, FR-052).

**Scale/Scope**: 3 skills, 14 remisiones reescritas, movidas o nuevas (8, 4 y 2), 4 movimientos (2 y 2); 3 familias, 8 ficheros de control nuevos; 13 filas y
520 sesiones en la aceptación; 3 tablas.

## Constitution Check

*GATE: Must pass before Phase 0 research. Re-check after Phase 1 design.*

| Principio | Cómo lo cumple este hito |
|---|---|
| I · Fuentes públicas y frontera humana (INNEGOCIABLE) | No toca ninguna fuente ni `internal/httpx` y no añade ninguna petición. Las sesiones del sondeo se sirven de `testdata/evals/` con los proxies que rechazan toda petición (ADR 0043). `jurisprudencia` sigue sin consultar el CENDOJ: su remisión cambia de sitio, no lo que pide (ADR 0036). Nada se presenta ni se firma |
| II · Nada sin cita ni fuente | El sobre de ningún applet cambia. Las reglas de cita, las formas fijas y la declaración de lo no cubierto siguen en los cuerpos (contracts/remisiones.md §2 a §5); lo movido es una regla entera que llega a la referencia que el caso manda leer |
| III · Tests primero y offline | Primero los dos guiones de contracts/aceptacion.md, que hoy fallan por una aserción. Todos los tests nuevos, offline y sin modelo, contra fixtures que escriben las tareas; ninguna grabación. `internal/core` no cambia |
| IV · Arquitectura hexagonal con reglas ejecutables | El código va en `internal/evals`, que no importa ningún paquete del binario, y el test de las remisiones, en `internal/skills`, sin código nuevo del paquete (research V11); `cmd/laboratorio` sigue siendo la raíz que nombra `os.Std*` y llama a `os.Exit`. Errores: la tabla, defecto `inesperado` con 1; el arnés, error de uso como en H39. `kitlegal` y sus códigos no cambian. Ningún `panic` |
| V · Simplicidad y dependencias fijadas (YAGNI) | Ninguna dependencia nueva. Cada mecanismo se traza a un FR («Trazabilidad»); lo que llega de fuera manipulado (un `sondeo.json` sin lista, una transcripción rota) es defecto `inesperado` con la regla genérica de H39 (código 1 y sin tabla), sin test propio; el de un `sondeo.json` sin lista nombra el fichero y la variante |
| VI · Un binario, convenciones de agente | Un solo binario `kitlegal`, sin banderas, variables ni órdenes nuevas. `cmd/laboratorio` no se distribuye (`.goreleaser.yaml` construye solo `./cmd/kitlegal`) |
| VII · Grafo y privacidad | No toca el grafo. Las sesiones usan su caché y su grafo en su directorio, como las del job; ningún dato de una persona |
| VIII · Skills primero; el binario es la herramienta | Mejora tres skills del producto, medibles con sus evals (FR-051) y con el sondeo de la persona (FR-052), según la guía de autoría (ADR 0046, 0050 y 0051): lo que toda respuesta necesita sigue en el cuerpo, que cabe en 11 000; lo demás, en referencias que el cuerpo manda leer; sin `scripts/` |
| IX · Genericidad territorial, validación local | Ninguna remisión ni dato nombra un municipio; las evals del sondeo de `plazos-y-territorio` son las de hoy, con un municipio cubierto (Leganés) y otro no (Tordesillas), y no cambian |
| X · Genericidad por materia, validación administrativa | Ningún caso especial para una materia en código, `data/` ni protocolo: las órdenes se escriben por capacidad. La orden generada de cada paquete y su generador no cambian (FR-005), y el control del producto desde el control no compara lo que genera `make skills-sync`, para que un paquete nuevo entre como datos sin tocar código (research D7) |
| Regla 1 · `internal/core/**` no importa adaptadores | `internal/core` no cambia |
| Regla 2 · solo `internal/httpx` importa `net/http` | Nada nuevo importa `net/http` |
| Regla 3 · solo `internal/{cache,store,graph}` importan SQLite | Nada nuevo importa `database/sql` ni `modernc.org/sqlite` |
| Regla 4 · solo `internal/cli` y `cmd/` llaman a `os.Exit` | Solo `cmd/laboratorio/main.go`, como hoy |
| Regla 5 · solo `internal/render` escribe en stdout; logs a stderr | `internal/evals` escribe en el `io.Writer` que le inyecta `cmd/laboratorio`, como hoy; `internal/skills` no cambia fuera de sus tests |
| Regla 6 · `internal/graph` no importa `internal/source/*` ni `internal/render` | `internal/graph` no cambia |
| Regla 7 · solo `internal/mcp` importa el SDK de MCP | Nada nuevo lo importa: el servidor de las sesiones es el `kitlegal mcp serve` de hoy |

Sin violaciones. Re-evaluado tras el diseño de la fase 1: igual.

## Project Structure

### Documentation (this feature)

```text
specs/033-h41-las-skills-del/
├── plan.md
├── research.md
├── data-model.md
├── quickstart.md
├── contracts/
│   ├── remisiones.md      # FR-001 a FR-010: los cambios exactos de las tres skills, la regla y el cálculo desde control
│   ├── variantes.md       # FR-020 a FR-023
│   ├── arnes.md           # FR-030 a FR-033
│   ├── tabla.md           # FR-040 a FR-042
│   └── aceptacion.md      # los dos guiones congelados
├── aceptacion/{laboratorio-tabla,remisiones}.txtar   # la primera tarea; congelados
└── tasks.md
```

### Source Code (repository root)

```text
skills/normativa/SKILL.md                                   # R1–R6 y la salida de lo movido (contracts/remisiones.md §2)
skills/normativa/references/redaccion_modificada.md         # lo movido, al final (§3)
skills/normativa/references/lecturas_que_fallan.md          # lo movido, al final (§3; corrección de T003)
skills/plazos-y-territorio/SKILL.md                         # P1–P4 y la salida de lo movido (§4)
skills/plazos-y-territorio/references/{calculo_de_plazos,municipios}.md   # lo movido, al final (§4; T004 y T009)
skills/jurisprudencia/SKILL.md                              # J1, J2 (§5)
laboratorio/plazos-y-territorio/plazos-y-territorio-control/skills/plazos-y-territorio/**   # NUEVO: 5 ficheros
laboratorio/jurisprudencia/jurisprudencia-control/skills/jurisprudencia/**                  # NUEVO: 3 ficheros
internal/skills/remisiones_test.go                          # NUEVO: TestRemisionesDelRepositorio (§6)
internal/evals/laboratorio_familias.go                      # antes laboratorio_normativa.go: el tipo, `producto` y las tres familias
internal/evals/laboratorio_familias_test.go                 # NUEVO: TestVariantesDeControl, TestSkillsDelProductoDesdeControl
internal/evals/laboratorio_normativa_test.go                # TestVariantesDeNormativa sin los dos subtests que pasan a TestVariantesDeControl
internal/evals/laboratorio_sondeo.go                        # cuatro familias, la mezcla, `<skill>-producto` desde skills/, la lista
internal/evals/laboratorio_sondeo_test.go                   # + TestArnesDeLasFamilias
internal/evals/laboratorio_tabla.go                         # `sondeoLeido` + `referencias_del_producto` (T005); tabla de una familia del producto, sus referencias, la lectura con cd
internal/evals/laboratorio_tabla_test.go                    # + TestTablaDeLasFamilias, TestLecturaConCd
internal/evals/laboratorio_test.go                          # TestLaboratorioFueraDelProducto, generalizado
internal/evals/testdata/laboratorio/familias/{normativa,plazos-y-territorio,jurisprudencia}/   # NUEVO [datos]: sondeo/, tabla.md
internal/evals/testdata/laboratorio/familias/mezcla/sondeo/h41/   # NUEVO [datos]: normativa-control y plazos-y-territorio-control, sin tabla.md
cmd/laboratorio/main.go                                     # comentario de paquete: las cuatro familias
scripts/evals-laboratorio.sh                                # cabecera: las cuatro familias (comentario)
Makefile                                                    # skills-check: + TestRemisionesDelRepositorio; ayuda de evals-laboratorio y laboratorio-tabla
CONTRIBUTING.md                                             # «El sondeo del laboratorio»: las familias del producto y la aceptación
docs/SKILLS.md                                              # §1 «Cómo se remite» (research D16)
CHANGELOG.md                                                # Unreleased
```

**Structure Decision**: la de `docs/ROADMAP.md` §2 y `CLAUDE.md`. Las skills, en `skills/`; las variantes, en el
laboratorio de H39 y H40, fuera del producto; el código, junto al del arnés y la tabla en `internal/evals`, y el test de
las remisiones, con los del repositorio en `internal/skills`; ningún paquete nuevo.

## Diseño

- **Skills** (contracts/remisiones.md §2 a §5; research D1 a D4): los cambios, como `diff -u` del control; cada remisión
  con su sitio, condición y lo que trae; dos movimientos en `normativa`, dos en `plazos-y-territorio` y ninguno en
  `jurisprudencia`.
- **Regla de las remisiones** (§6; research D5, D6): por cada enlace a `references/<f>` fuera de un bloque de código,
  «lee» antes del enlace en su frase y ninguna de las tres expresiones; un test del repositorio, que también exige los
  fallos de cada `<skill>-control`.
- **Producto desde control** (§7; research D7): sustituciones y añadidos declarados, comparación byte a byte, sin lo
  generado; el frontmatter, idéntico.
- **Familias y variantes** (contracts/variantes.md; data-model §1 a §3; research D8, D9): `producto` en
  `familiaDelLaboratorio`; huellas SHA-256 del control.
- **Arnés** (contracts/arnes.md; research D10 a D12): `<skill>-producto` desde `skills/`; `referencias_del_producto` en
  `sondeo.json`; el mensaje de la mezcla; las evals de cada familia.
- **Tabla** (contracts/tabla.md; research D13, D14): la familia por la primera variante; las referencias, unión de las
  declaradas y las listas; la lectura con `cd` solo en las familias del producto; celdas vacías donde la variante no
  tiene la referencia; «Cumplen la regla» solo en `normativa`.

**Aceptación e2e:** dos guiones testscript que escribe la primera tarea en `specs/033-h41-las-skills-del/aceptacion/` y
quedan congelados (contracts/aceptacion.md), activados como `internal/app/testdata/script/h41-*.txtar` y ejecutados por
`TestEntregaDelHito`:
- `aceptacion/laboratorio-tabla.txtar` (historia 4): `laboratorio tabla` sobre sesiones fijas de las tres familias da,
  byte a byte, sus tablas —lecturas con `cd` que cuentan y que no, la columna de una referencia de la lista del producto
  vacía en las variantes que no la tienen, «Cumplen la regla» solo en `normativa`— y un directorio de otra familia es un
  defecto con código 1. Cubre FR-040, FR-041 y SC-005.
- `aceptacion/remisiones.txtar` (historias 1 y 2): lo que instala `kitlegal skills install --dir` lleva cada orden de leer
  de las tres skills en su sitio, ninguna de las remisiones de antes, y lo movido en su referencia y fuera del cuerpo.
  Cubre FR-001 a FR-006, FR-009, SC-001 y SC-004.

La historia 3 abre sesiones con modelo y credencial: su aceptación es `TestArnesDeLasFamilias` con el sustituto de
`claude` en `make ci` (FR-030, FR-033; SC-006, SC-007). La historia 5 la hace la persona antes de fusionar (FR-052).

## Trazabilidad

| Mecanismo | FR/SC |
|---|---|
| Remisiones reescritas de las tres skills; sus sitios | FR-001 a FR-006 |
| Movimientos: «Nada de otra conversación» y la lectura que falla, en `normativa`; `⚠ PLAZO:` con 4 o 5 y el municipio que falta, en `plazos-y-territorio` | FR-008, FR-009 |
| `TestRemisionesDelRepositorio` y su nombre en `make skills-check` | FR-001 a FR-003, SC-001 |
| `TestSkillsDelProductoDesdeControl` y su declaración | FR-007, FR-009, FR-010, SC-003, SC-004 |
| `laboratorio/<skill>/<skill>-control` (dos nuevas) y sus huellas | FR-020, SC-008 |
| Campo `producto`; familias de `plazos-y-territorio` y `jurisprudencia` | FR-022, FR-030 |
| Copia de `skills/<skill>` en el sondeo; `referencias_del_producto` | FR-022, FR-033 |
| Mensaje de la mezcla de familias | FR-030 |
| Tabla de una familia del producto; unión de referencias; celdas vacías | FR-040 |
| Lectura con `cd` (solo familias del producto) | FR-041 |
| `TestLaboratorioFueraDelProducto` generalizado | FR-022, FR-023 |
| Fixtures `testdata/laboratorio/familias/` | FR-040, FR-041, SC-005 |
| Guiones congelados | FR-001 a FR-006, FR-009, FR-040, FR-041 |
| `CHANGELOG.md`, `docs/SKILLS.md` | FR-050 |
| `CONTRIBUTING.md`, ayuda de `make`, comentarios de `cmd/laboratorio` y del guion | FR-020, FR-030, FR-040: lo que hoy documentan deja de ser cierto (research D16) |

## Gates: controles mecánicos que añade o toca

- **`make ci`**: ningún objetivo nuevo; corre todo lo que sigue. **`make skills-check`**: su expresión de tests gana
  `TestRemisionesDelRepositorio`; nada más de su receta cambia.
- **Objetivos de `make`**: `evals-laboratorio` y `laboratorio-tabla`, con las mismas recetas (`TestRecetasDelSondeo`);
  cambia su ayuda.
- **Tests nuevos en `make ci`**: `TestRemisionesDelRepositorio` (`internal/skills/remisiones_test.go`: las skills de
  `skills/`, sin fallos; en «control», el `SKILL.md` de cada `<skill>-control`, con exactamente sus fallos; en «frases»,
  la frase cortada por un bloque de código y la expresión partida por un salto de línea); `TestVariantesDeControl` y
  `TestSkillsDelProductoDesdeControl`
  (`laboratorio_familias_test.go`); `TestArnesDeLasFamilias` (`laboratorio_sondeo_test.go`); `TestTablaDeLasFamilias`
  y `TestLecturaConCd` (`laboratorio_tabla_test.go`).
- **Tests que cambian**: `TestVariantesDeNormativa` (salen «control es skills/normativa» y «directorios de la familia»,
  que pasan a `TestVariantesDeControl`); `TestLaboratorioFueraDelProducto`, a las tres familias. Sin cambiar ningún valor
  esperado: `TestArnesDeNormativa`, `TestArnesDelLaboratorio`, `TestTablaDeNormativa`, `TestTablaDelLaboratorio`,
  `TestReglaTestigo`, `TestReglaTestigoDeNormativa`, `TestOrdenesDeBash`, `TestServidorDelLaboratorio`,
  `TestRecetasDelSondeo`, `TestSkillsDelRepositorio`.
- **Suites congeladas**: las dos de H41 (contracts/aceptacion.md); las de H39 y H40 y `h32-normativa.txtar`, sin cambiar
  un fichero (FR-042; research V4).
- **Fixtures** (tareas `[datos]`, sin red): `internal/evals/testdata/laboratorio/familias/<skill>/sondeo/` (sesiones fijas
  de los dos modos y sus `sondeo.json`) y `…/tabla.md`, por familia; y `internal/evals/testdata/laboratorio/familias/mezcla/sondeo/h41/`,
  un sondeo de `normativa` con un directorio de `plazos-y-territorio`, sin `tabla.md`. Ninguna grabación; ningún fichero de `testdata/`
  existente cambia.
- **Fuera de `make ci`**: `TestSondeoDelLaboratorio` (`-tags evals`), la entrada de la persona.

## Datos externos

Ninguno (research, «Fuentes externas»): no hay manifiesto `grabaciones.json`, test `TestGrabar*` ni tarea que grabe;
las sesiones del sondeo se sirven de las grabaciones que ya usa el job, y el run no escribe en `evidencias/` ni en
`data/`.

## Controles de umbral

| Requisito | Umbral | Control | Dónde |
|---|---|---|---|
| SC-001, FR-001 a FR-003 | 100 % de las referencias de las tres skills con al menos una remisión que es una orden; 0 frases con un enlace a una referencia sin «lee» o «Lee» antes del enlace, o con «si la tienes», «antes de escribir la respuesta» o «antes de dar la consulta», en mayúscula o minúscula | La frase de cada enlace a una referencia de cada skill de `skills/`, y cada referencia con algún enlace; y las frases de control, que tienen que fallar, también «Antes de dar la consulta» al principio de la frase de `jurisprudencia-control` (contracts/remisiones.md §6.3 y §6.4) | ci:internal/skills/remisiones_test.go:TestRemisionesDelRepositorio |
| SC-002, FR-008 | 0 skills de `skills/` con un cuerpo de más de 11 000 caracteres | El defecto `cuerpo de N caracteres (máximo 11000)`, con la skill y su recuento | ci:internal/app/skills_test.go:TestSkillsDelRepositorio |
| SC-003, FR-010 | 0 bytes de diferencia entre el frontmatter de cada skill y el de su `<skill>-control`; 0 derivas tras `make skills-sync` | El frontmatter, byte a byte, contra el de control; y la regeneración en memoria contra el árbol | ci:internal/evals/laboratorio_familias_test.go:TestSkillsDelProductoDesdeControl, ci:internal/app/skills_test.go:TestSkillsDelRepositorio |
| SC-004, FR-007, FR-009 | 100 % de los movimientos en su referencia y fuera del cuerpo; 0 bytes de diferencia con `<skill>-control` fuera de lo declarado, en el cuerpo y en las referencias | Calcula cada skill desde su control con su declaración y la compara byte a byte | ci:internal/evals/laboratorio_familias_test.go:TestSkillsDelProductoDesdeControl |
| SC-001, SC-004, FR-001 a FR-006, FR-009 | 100 % de las órdenes de contracts/aceptacion.md §2 en lo que instala el binario; 0 remisiones de antes; lo movido en su referencia | El guion congelado `h41-remisiones.txtar` | ci:internal/app/e2e_test.go:TestEntregaDelHito |
| SC-005, FR-040, FR-041 | 0 diferencias en recuentos, turnos y extremos de Wilson de las sesiones fijas de las tres familias, con las lecturas con `cd` que cuentan y las que no | La salida, byte a byte, contra cada `tabla.md`; la regla, orden a orden; y el guion congelado `h41-laboratorio-tabla.txtar` | ci:internal/evals/laboratorio_tabla_test.go:TestTablaDeLasFamilias, ci:internal/evals/laboratorio_tabla_test.go:TestLecturaConCd, ci:internal/app/e2e_test.go:TestEntregaDelHito |
| SC-006, FR-022, FR-033 | 0 bytes de diferencia entre la skill de una sesión de `<skill>-producto` y `skills/<skill>`; la lista de sus referencias en `sondeo.json`; 0 directorios de variante `<skill>-producto` en el repositorio | Las sesiones del sustituto de `claude`; las carpetas de las familias; las rutas de `laboratorio/` | ci:internal/evals/laboratorio_sondeo_test.go:TestArnesDeLasFamilias, ci:internal/evals/laboratorio_familias_test.go:TestVariantesDeControl, ci:internal/evals/laboratorio_test.go:TestLaboratorioFueraDelProducto |
| SC-007, FR-030 | 0 sesiones con una petición inválida; n/4 por eval en cada variante y modo pedidos con una válida, en cada familia | Sustituto de `claude` que anota sus invocaciones (contracts/arnes.md §5) | ci:internal/evals/laboratorio_sondeo_test.go:TestArnesDeLasFamilias |
| SC-008, FR-020 | 0 bytes de diferencia, 0 ficheros de más o de menos, en cada `<skill>-control` respecto de `v0.9.0` | Las huellas SHA-256 por ruta | ci:internal/evals/laboratorio_familias_test.go:TestVariantesDeControl |
| SC-008, FR-021 | 0 bytes fuera de lo declarado entre cada variante de H40 y `normativa-control` | El cálculo de H40, sin cambiar sus valores | ci:internal/evals/laboratorio_normativa_test.go:TestVariantesDeNormativa |
| SC-008, FR-023 | 0 ficheros de variantes versionadas en lo empotrado, en lo instalado y en lo que el job ejecuta o elige | Recorre lo empotrado, una instalación en un temporal, la definición del job, `evals/` y `rutaQueMide` | ci:internal/evals/laboratorio_test.go:TestLaboratorioFueraDelProducto |
| SC-008, FR-042 | 0 fallos en las suites congeladas de H39 y H40, sin cambiar un fichero suyo | Sus guiones activados | ci:internal/app/e2e_test.go:TestEntregaDelHito |
| SC-009, FR-051 | Las evals de `normativa` en verde en el cierre, en los dos modos: sus umbrales con `decide: true` | El job de evals del cierre | evals:normativa:sin_activar:claude-sonnet-5-5:orden, evals:normativa:sin_activar:claude-sonnet-5-5:herramienta, evals:normativa:redaccion_sin_fechar:claude-sonnet-5-5:orden, evals:normativa:redaccion_sin_fechar:claude-sonnet-5-5:herramienta, evals:normativa:afirma_lo_no_leido:claude-sonnet-5-5, evals:normativa:medida_del_juez:afirma_lo_no_leido:defectos_sin_marcar, evals:normativa:medida_del_juez:afirma_lo_no_leido:correctos_marcados, evals:normativa:duracion_de_las_sesiones:orden, evals:normativa:duracion_de_las_sesiones:herramienta, evals:normativa:duracion_del_juez:orden, evals:normativa:duracion_del_juez:herramienta |
| SC-009, FR-051 | Las evals de `jurisprudencia` en verde en el cierre, en los dos modos | El job de evals del cierre | evals:jurisprudencia:sin_activar:claude-sonnet-5-5:orden, evals:jurisprudencia:sin_activar:claude-sonnet-5-5:herramienta, evals:jurisprudencia:cita_sin_documento:claude-sonnet-5-5:orden, evals:jurisprudencia:cita_sin_documento:claude-sonnet-5-5:herramienta, evals:jurisprudencia:documento_sin_cotejar:claude-sonnet-5-5:orden, evals:jurisprudencia:documento_sin_cotejar:claude-sonnet-5-5:herramienta, evals:jurisprudencia:afirma_lo_no_leido:claude-sonnet-5-5, evals:jurisprudencia:medida_del_juez:afirma_lo_no_leido:defectos_sin_marcar, evals:jurisprudencia:medida_del_juez:afirma_lo_no_leido:correctos_marcados, evals:jurisprudencia:duracion_del_juez:orden, evals:jurisprudencia:duracion_del_juez:herramienta |
| SC-009, FR-051 | Las evals de `plazos-y-territorio` en verde en el cierre, en los dos modos | El job de evals del cierre | evals:plazos-y-territorio:sin_activar:claude-sonnet-5-5:orden, evals:plazos-y-territorio:sin_activar:claude-sonnet-5-5:herramienta, evals:plazos-y-territorio:plazo_sin_fecha:claude-sonnet-5-5:orden, evals:plazos-y-territorio:plazo_sin_fecha:claude-sonnet-5-5:herramienta, evals:plazos-y-territorio:cobertura_sin_declarar:claude-sonnet-5-5:orden, evals:plazos-y-territorio:cobertura_sin_declarar:claude-sonnet-5-5:herramienta, evals:plazos-y-territorio:vigencia_sin_declarar:claude-sonnet-5-5:orden, evals:plazos-y-territorio:vigencia_sin_declarar:claude-sonnet-5-5:herramienta, evals:plazos-y-territorio:redaccion_sin_fechar:claude-sonnet-5-5:orden, evals:plazos-y-territorio:redaccion_sin_fechar:claude-sonnet-5-5:herramienta, evals:plazos-y-territorio:lectura_sin_intentar:claude-sonnet-5-5:orden, evals:plazos-y-territorio:lectura_sin_intentar:claude-sonnet-5-5:herramienta |

SC-009 usa los umbrales del job tal como están (`.github/workflows/evals.yml` y las carpetas de evals), que el hito no
toca: las filas nombran los que publican con `decide: true` los informes del cierre de H40 (research V12), y el veredicto
del job sale en rojo con cualquiera de ellos. `consulta-juridica` no publica ninguno con `decide: true` y su skill no
cambia; su veredicto lo deciden sus series (ADR 0016), configuración que el hito no toca. FR-032 (trece filas, 520
sesiones) y SC-010 los mide la persona (capa 3): el sondeo abre sesiones con modelo que ni `make ci` ni el run abren.

## Cierre y pendientes

- **FR-052**: el run no lanza el sondeo; la propuesta de cambio queda pendiente de sus tablas y su lectura, y el informe
  final lo dice por su línea `[alcance]` en `gates/supuestos.md`; la propuesta no se fusiona sin ellas, y si no cumplen,
  lo medido va a un ADR que dice qué medir después.
- **FR-051**: si la medición del cierre da un rojo, la reparación corrige la skill dentro de lo que este plan declara
  (contracts/remisiones.md) y el informe da sus cifras; ningún umbral se rebaja.

## Complexity Tracking

| Desviación | Por qué hace falta | Alternativa más simple rechazada y por qué |
|---|---|---|
| Dos copias más de una skill en `laboratorio/` (8 ficheros) | FR-020: `<skill>-control` congelada por familia, que el arnés instala tal cual | Instalarlas desde la etiqueta `v0.9.0` al lanzar el sondeo: `make ci` no podría comprobar que no cambian y el clon de la entrega es superficial (research D8) |
| El cálculo del producto desde control deja fuera lo generado | FR-007 sin romper el principio X: un paquete nuevo en `data/` cambia lo generado y no puede pedir un cambio de código | Comparar también lo generado: pondría el test en rojo con cada paquete; FR-010 ya lo cubre (research D7) |
| Un guion que lee lo que instala el binario | ADR 0018: la entrega de las historias 1 y 2 es lo que `kitlegal skills install` pone en el proyecto de quien lo usa | Solo tests sobre `skills/`: no describen lo que recibe quien instala (research D15) |
