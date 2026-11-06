# Implementation Plan: H41 · Las skills del producto mandan leer sus referencias

**Branch**: `034-h41-las-skills-del` | **Date**: 2026-11-06 | **Spec**: [spec.md](spec.md)

**Input**: Feature specification from `specs/034-h41-las-skills-del/spec.md`

## Summary

El hito vuelve a escribir cada remisión de `normativa` y `jurisprudencia` como una orden de leer —«lee», el enlace,
cuándo y lo que trae—, en su sitio, como en la #245 (FR-001 a FR-005), y hace en `normativa` el sitio que necesitan con
los dos movimientos de la #245 (FR-006): los cuerpos quedan en 10 926 y 10 842 caracteres (FR-008). Deja
`skills/plazos-y-territorio` como la de `v0.9.0` (FR-010) y pone en el laboratorio, junto a su control, tres variantes
que solo cambian la forma de las órdenes de `municipios.md` y `calculo_de_plazos.md` (FR-023), para que el sondeo diga
cuál aplica H41.1. Amplía el arnés de H39 y H40 a las tres familias del producto, cada una con su `<skill>-control`
congelada y su `<skill>-producto` instalada desde `skills/` (FR-020 a FR-033), y la tabla, que cuenta las tres familias
y la lectura con `cd` al directorio de las referencias (FR-040, FR-041). El run no abre sesiones con modelo: el sondeo de
diecinueve filas y 760 sesiones y su lectura los hace la persona antes de fusionar (FR-052, FR-053).

Enfoque técnico (research D1 a D20): lo que la #245 midió vuelve tal cual desde la rama `033-h41-las-skills-del`, con
lo que cambia y su porqué en «Lo que cambia respecto a la #245»; las órdenes de las variantes, exactas y medidas sobre
copias en esta sesión (contracts/variantes.md); una sola definición de «frase» (`skills.Frases`) para la regla de las
remisiones y para el control de las variantes; el producto calculado desde su control; las familias declaradas en Go con
un campo `producto`; y dos guiones congelados, uno de lo que instala el binario y otro de la tabla.

## Technical Context

**Language/Version**: Go 1.27 (`go 1.27.0`, `toolchain go1.27.3` en `go.mod`; `go version`: go1.27.3).

**Primary Dependencies**: las fijadas, sin ninguna nueva. Biblioteca estándar en el código (`crypto/sha256`,
`encoding/json`, `os`, `regexp`, `strings`, `unicode`) y lo que ya usan `internal/skills` e `internal/evals`;
`stretchr/testify` y `rogpeppe/go-internal` v1.16.0 (testscript) en los tests.

**Storage**: ficheros: `skills/` y `laboratorio/` versionados; cada sondeo, en un `DIRECTORIO` de la persona, con su
`sondeo.json` (data-model §4).

**Testing**: `go test -race` en `make ci`; dos guiones testscript congelados; tests de tabla con `t.Parallel()` y
fixtures en `internal/evals/testdata/laboratorio/familias/`; el sustituto de `claude` de `TestArnesDelLaboratorio`;
ningún test con red ni con modelo. Skills aplicadas: `golang-how-to`, con `golang-testing` (tests de tabla, fixtures en
`testdata/`, testscript), `golang-project-layout` (ningún paquete nuevo: el código va con el de H39 y H40 en
`internal/evals`, y `Frases` y el test de las remisiones, con los lectores de Markdown y los tests del repositorio de
`internal/skills`), `golang-lint` (ningún `//nolint`; `producto` entre las palabras de misspell, research V13),
`golang-naming` (el fichero de las familias se llama por lo que declara) y `golang-cli` (ningún argumento ni bandera
nuevos).

**Target Platform**: macOS y Linux, donde corren `make ci` y el sondeo de la persona.

**Project Type**: skills sin código del producto, y programas del repositorio (`cmd/laboratorio`, guiones de
`scripts/`); el binario `kitlegal` no cambia, salvo las skills que empotra.

**Performance Goals**: ninguno. Cada sesión del sondeo, con el tope de hoy.

**Constraints**: cuerpo de cada `SKILL.md` de `skills/` ≤ 11 000 caracteres, también con la línea de 70 de la vertical
de ensayo de H32; las suites congeladas de H39, H40 y H32 siguen en verde sin cambiar un fichero (research V4); el run no
graba, no escribe en `evidencias/` y no abre sesiones con modelo (FR-031, FR-052).

**Scale/Scope**: 2 skills con 10 frases de remisión (7 y 3) y 2 movimientos; 3 familias; 8 ficheros de control nuevos y
15 de las tres variantes; 19 filas y 760 sesiones en la aceptación; 3 tablas.

## Constitution Check

*GATE: Must pass before Phase 0 research. Re-check after Phase 1 design.*

| Principio | Cómo lo cumple este hito |
|---|---|
| I · Fuentes públicas y frontera humana (INNEGOCIABLE) | No toca ninguna fuente ni `internal/httpx` y no añade ninguna petición. Las sesiones del sondeo se sirven de las grabaciones del job, con los proxies que rechazan toda petición (ADR 0043). `jurisprudencia` sigue sin consultar el CENDOJ: sus remisiones cambian de sitio, no lo que piden (ADR 0036). Nada se presenta ni se firma |
| II · Nada sin cita ni fuente | El sobre de ningún applet cambia. Las reglas de cita, las formas fijas y la declaración de lo no cubierto siguen en los cuerpos (contracts/remisiones.md §2 a §4); lo movido es una regla entera que llega, palabra por palabra, a la referencia que su caso manda leer. `plazos-y-territorio` no cambia |
| III · Tests primero y offline | Primero los dos guiones de contracts/aceptacion.md, que hoy fallan por una aserción. Todos los tests nuevos, offline y sin modelo, contra fixtures que escriben las tareas `[datos]`; ninguna grabación. `internal/core` no cambia, y la cobertura global no baja: el código nuevo de `internal/evals` e `internal/skills` lleva sus tests |
| IV · Arquitectura hexagonal con reglas ejecutables | El código va en `internal/evals` y en `internal/skills`, que no van en el binario (research V10); `cmd/laboratorio` sigue siendo la raíz que nombra `os.Std*` y llama a `os.Exit`. Errores: la tabla, defecto `inesperado` con 1 y el de otra familia con 1, como H40; el arnés, error de uso como en H39; `kitlegal` y sus códigos no cambian. Ningún `panic` |
| V · Simplicidad y dependencias fijadas (YAGNI) | Ninguna dependencia nueva. Cada mecanismo se traza a un FR («Trazabilidad»); lo que el guion no sabe leer es la regla genérica del spec (casos límite: defecto `inesperado`, código 1, sin más casos), sin comprobación, mensaje ni test propios para un estado manipulado (research D17, D20). Las desviaciones, en Complexity Tracking |
| VI · Un binario, convenciones de agente | Un solo binario `kitlegal`, sin banderas, variables ni órdenes nuevas. `cmd/laboratorio` no se distribuye (`.goreleaser.yaml:30` construye solo `./cmd/kitlegal`) |
| VII · Grafo y privacidad | No toca el grafo. Las sesiones usan su caché y su grafo en su directorio, como las del job; ningún dato de una persona |
| VIII · Skills primero; el binario es la herramienta | Mejora dos skills del producto, medibles con sus evals (FR-051) y con el sondeo de la persona (FR-052), según la guía de autoría (ADR 0046, 0050, 0051 y 0052): lo que toda respuesta necesita sigue en el cuerpo, que cabe en 11 000; lo demás, en referencias que el cuerpo manda leer; sin `scripts/`. Prepara con la medida de las variantes el cambio de la tercera (H41.1) |
| IX · Genericidad territorial, validación local | Ninguna remisión ni dato nombra un municipio; las órdenes de las variantes hablan de «el municipio» y de «los municipios que cuentan». Las evals del sondeo de `plazos-y-territorio` son las de hoy, con un municipio cubierto (Leganés) y otro no (Tordesillas), y no cambian |
| X · Genericidad por materia, validación administrativa | Ningún caso especial para una materia en código, `data/` ni protocolo: las órdenes se escriben por capacidad; el generador y la orden generada de cada paquete no cambian. La lista de remisiones y el cálculo desde control dejan fuera lo generado, para que un paquete entre como datos (research D7, D8). La comparación total de `plazos-y-territorio` que pide FR-010 incluye lo generado hasta H41.1 (Complexity Tracking) |
| Regla 1 · `internal/core/**` no importa adaptadores | `internal/core` no cambia |
| Regla 2 · solo `internal/httpx` importa `net/http` | Nada nuevo importa `net/http` |
| Regla 3 · solo `internal/{cache,store,graph}` importan SQLite | Nada nuevo importa `database/sql` ni `modernc.org/sqlite` |
| Regla 4 · solo `internal/cli` y `cmd/` llaman a `os.Exit` | Solo `cmd/laboratorio/main.go`, como hoy |
| Regla 5 · solo `internal/render` escribe en stdout; logs a stderr | `internal/evals` escribe en el `io.Writer` que le inyecta `cmd/laboratorio`, como hoy; `skills.Frases` no escribe nada |
| Regla 6 · `internal/graph` no importa `internal/source/*` ni `internal/render` | `internal/graph` no cambia |
| Regla 7 · solo `internal/mcp` importa el SDK de MCP | Nada nuevo lo importa: el servidor de las sesiones es el `kitlegal mcp serve` de hoy |
| Garantías sin red de H19 y H23 (`TestArquitectura`) | Ninguna importación cambia en lo que instala las skills ni en lo que prepara y coteja una cita: el hito cambia el contenido empotrado, no el código que lo instala |

Sin violaciones sin justificar. Re-evaluado tras el diseño de la fase 1: igual.

## Project Structure

### Documentation (this feature)

```text
specs/034-h41-las-skills-del/
├── plan.md
├── research.md
├── data-model.md
├── quickstart.md
├── contracts/
│   ├── remisiones.md      # FR-001 a FR-010: los cambios exactos, la lista, la frase, la regla y el producto desde control
│   ├── variantes.md       # FR-020 a FR-025: las tres familias, las tres formas de plazos-y-territorio y sus cuerpos
│   ├── arnes.md           # FR-030 a FR-033
│   ├── tabla.md           # FR-040 a FR-042
│   └── aceptacion.md      # los dos guiones congelados
├── aceptacion/{skills-instaladas,laboratorio-tabla}.txtar   # la primera tarea; congelados
└── tasks.md
```

### Source Code (repository root)

```text
skills/normativa/SKILL.md                                   # R1–R6 y la salida de lo movido (contracts/remisiones.md §2)
skills/normativa/references/{redaccion_modificada,lecturas_que_fallan}.md   # lo movido, al final (§3)
skills/jurisprudencia/SKILL.md                              # J1, J2 (§4)
laboratorio/plazos-y-territorio/plazos-y-territorio-control/skills/plazos-y-territorio/**             # NUEVO: 5 ficheros
laboratorio/plazos-y-territorio/plazos-y-territorio-{sin-condicion,condicion-por-orden,lista-con-verbo}/skills/plazos-y-territorio/**   # NUEVO: 3 × 5
laboratorio/jurisprudencia/jurisprudencia-control/skills/jurisprudencia/**                            # NUEVO: 3 ficheros
internal/skills/frases.go                                   # NUEVO: Frases (contracts/remisiones.md §6)
internal/skills/frases_test.go                              # NUEVO: TestFrases
internal/skills/remisiones_test.go                          # NUEVO: TestRemisionesDelRepositorio (§7)
internal/evals/laboratorio_familias.go                      # antes laboratorio_normativa.go: el tipo, `producto`, las tres familias
internal/evals/laboratorio_familias_test.go                 # NUEVO: TestVariantesDeControl, TestSkillsDelProductoDesdeControl, TestVariantesDePlazosYTerritorio
internal/evals/laboratorio_normativa_test.go                # TestVariantesDeNormativa sin sus dos subtests de control; ayudantes con la familia
internal/evals/laboratorio_sondeo.go                        # cuatro familias, la mezcla, `<skill>-producto` desde skills/, la lista
internal/evals/laboratorio_sondeo_test.go                   # + TestArnesDeLasFamilias
internal/evals/laboratorio_tabla.go                         # `referencias_del_producto` en sondeoLeido; tabla de una familia del producto; lectura con cd
internal/evals/laboratorio_tabla_test.go                    # + TestTablaDeLasFamilias, TestLecturaConCd
internal/evals/laboratorio_test.go                          # TestLaboratorioFueraDelProducto, a las tres familias
internal/evals/testdata/laboratorio/familias/{normativa,plazos-y-territorio,jurisprudencia}/   # NUEVO [datos]: sondeo/, tabla.md
internal/evals/testdata/laboratorio/familias/mezcla/sondeo/h41/                                 # NUEVO [datos]: sin tabla.md
cmd/laboratorio/main.go                                     # comentario de paquete: las cuatro familias
scripts/evals-laboratorio.sh                                # cabecera: las cuatro familias (comentario)
Makefile                                                    # skills-check: + TestRemisionesDelRepositorio; ayuda de evals-laboratorio y laboratorio-tabla
.golangci.yml                                               # misspell: `producto` (research V13)
CONTRIBUTING.md                                             # tabla de objetivos y «El sondeo del laboratorio»: las familias del producto
docs/SKILLS.md                                              # §1, tras «Cómo se remite» (research D19)
CHANGELOG.md                                                # Unreleased
```

**Structure Decision**: la de `docs/ROADMAP.md` §2 y `CLAUDE.md`. Las skills, en `skills/`; las variantes, en el
laboratorio de H39 y H40, fuera del producto; el código, junto al del arnés y la tabla en `internal/evals`, y la frase y
el test de las remisiones, con los lectores de Markdown y los tests del repositorio en `internal/skills`; ningún paquete
nuevo. `skills/plazos-y-territorio` no se toca.

## Diseño

- **Skills** (contracts/remisiones.md §2 a §4; research D1 a D3): los cambios, como `diff -u` del control; cada remisión
  con su sitio, condición y lo que trae; dos movimientos en `normativa`, ninguno en `jurisprudencia`.
- **La frase** (§6; research D5, D6): `skills.Frases`, segmentos del cuerpo con las reglas de FR-001.
- **Regla y lista de las remisiones** (§5, §7; research D7): «lee» antes de cada enlace, ninguna de las tres
  expresiones, cada referencia con su remisión y cada frase la de la lista; `plazos-y-territorio` fuera hasta H41.1.
- **Producto desde control** (§8; research D4, D8): sustituciones y movimientos declarados, comparación byte a byte sin
  lo generado, el frontmatter idéntico; `plazos-y-territorio`, igual entera.
- **Familias y variantes** (contracts/variantes.md; data-model §1, §2, §6; research D9 a D13): `producto` en
  `familiaDelLaboratorio`; huellas SHA-256 del control; las tres formas, con su declaración y el control de que solo
  tocan sus órdenes.
- **Arnés** (contracts/arnes.md; research D14, D15): `<skill>-producto` desde `skills/`; `referencias_del_producto` en
  `sondeo.json`; el mensaje de la mezcla; las evals de cada familia.
- **Tabla** (contracts/tabla.md; research D16, D17): la familia por la primera variante; las referencias, unión de las
  declaradas y las listas; la lectura con `cd` solo en las familias del producto; celdas vacías donde la variante no
  tiene la referencia; «Cumplen la regla» solo en `normativa`.

**Aceptación e2e:** dos guiones testscript que escribe la primera tarea en `specs/034-h41-las-skills-del/aceptacion/` y
quedan congelados (contracts/aceptacion.md), activados como `internal/app/testdata/script/h41-*.txtar` y ejecutados por
`TestEntregaDelHito`:
- `aceptacion/skills-instaladas.txtar` (historias 1, 2 y 3): lo que pone `kitlegal skills install --dir` lleva cada
  orden de leer de `normativa` y `jurisprudencia` en su sitio, ninguna de las remisiones de antes y lo movido en su
  referencia; `plazos-y-territorio`, con la lista de control y «pregúntalo y espera» en el paso 1; y ninguna variante
  del laboratorio. Cubre FR-001 a FR-006, FR-010, FR-025; SC-001, SC-004, SC-005, SC-009.
- `aceptacion/laboratorio-tabla.txtar` (historia 5): `laboratorio tabla` sobre sesiones fijas de las tres familias
  —con una sesión de cada una de las cinco variantes de `plazos-y-territorio`, lecturas con `cd` que cuentan y que no, la
  columna de una referencia de la lista del producto vacía en las variantes que no la tienen y «Cumplen la regla» solo
  en `normativa`— da sus tablas byte a byte, y un directorio de otra familia es un defecto con código 1. Cubre FR-032,
  FR-040, FR-041; SC-006.

La historia 4 abre sesiones con modelo y credencial: su aceptación en `make ci` es `TestArnesDeLasFamilias` con el
sustituto de `claude` (FR-022, FR-030, FR-032, FR-033; SC-007, SC-008). La 6 la hace la persona antes de fusionar
(FR-052, FR-053). La 3 va en el primer guion porque uno solo suyo pasaría hoy (research D18).

## Lo que cambia respecto a la #245

| Qué | En la #245 | En este hito | Por qué |
|---|---|---|---|
| Orden de `redaccion_modificada.md` | «…, y si se pregunta qué cambió, lee» | «…, o si se pregunta qué cambió, lee» | FR-002 y ADR 0052, punto 2: dos casos no se juntan con «y» (research D1) |
| `skills/plazos-y-territorio` | P1 a P4 y dos movimientos | sin cambios | FR-010; ADR 0052, punto 4 |
| Variantes de `plazos-y-territorio` | control | control y tres formas | FR-023; ADR 0052, punto 3 (research D11) |
| `TestRemisionesDelRepositorio` | regla sobre las tres skills | regla sobre `normativa` y `jurisprudencia`, con la lista del plan; `plazos-y-territorio` fuera hasta H41.1 | FR-001, FR-010 (research D7) |
| Frase | lector propio del test | `skills.Frases`, compartida con el control de las variantes | FR-001, FR-023 (research D5) |
| Producto desde control | tres skills sin lo generado | `normativa` y `jurisprudencia` sin lo generado; `plazos-y-territorio` entera | FR-007, FR-010 (research D4, D8) |
| Tabla: `sondeo.json` de un sondeo con `<skill>-producto` sin `referencias_del_producto` | comprobación y mensaje propios | no vuelven; la regla genérica del spec | solo una edición a mano llega a ese estado: el arnés siempre escribe la lista (FR-032; research D17, D20) |
| Arnés | 13 filas, 520 sesiones | 19 filas, 760 sesiones; el test prueba las cuatro invocaciones | FR-030, FR-033 (research D14) |
| Guiones | `remisiones.txtar` y `laboratorio-tabla.txtar` | `skills-instaladas.txtar` (con la historia 3) y `laboratorio-tabla.txtar` con las cinco variantes | research D18 |
| `docs/SKILLS.md` | lo de la reorganización | lo mismo, con el límite del ADR 0052, punto 4; sin su punto 2 | FR-050 (research D19) |

## Trazabilidad

| Mecanismo | FR/SC |
|---|---|
| Remisiones reescritas de `normativa` y `jurisprudencia`; sus sitios | FR-001 a FR-005 |
| Movimientos de `normativa` (M-a, M-b) | FR-006, FR-008 |
| `skills.Frases` y `TestFrases` | FR-001 (la frase), FR-023 (el control de las variantes) |
| `TestRemisionesDelRepositorio`, su lista, su exclusión y su nombre en `make skills-check` | FR-001 a FR-003, FR-010, SC-001 |
| `TestSkillsDelProductoDesdeControl` y su declaración | FR-006, FR-007, FR-009, FR-010, SC-003 a SC-005 |
| `laboratorio/<skill>/<skill>-control` (dos nuevas) y sus huellas | FR-020, SC-009 |
| Tres variantes de `plazos-y-territorio`, su declaración y `TestVariantesDePlazosYTerritorio` | FR-023, FR-024, SC-009 |
| Ayudantes de H40 con la familia; `TestVariantesDeNormativa` sin sus subtests de control | FR-020, FR-021 |
| Campo `producto`; familias de `plazos-y-territorio` y `jurisprudencia` | FR-022, FR-030 |
| Copia de `skills/<skill>` en el sondeo; `referencias_del_producto` | FR-022, FR-032 |
| Mensaje de la mezcla de familias | FR-030 |
| `TestArnesDeLasFamilias` | FR-022, FR-030, FR-032, FR-033, SC-007, SC-008 |
| Tabla de una familia del producto; unión de referencias; celdas vacías | FR-032, FR-040 |
| Lectura con `cd` (solo familias del producto) | FR-041 |
| `TestLaboratorioFueraDelProducto` generalizado | FR-025, SC-009 |
| Fixtures `testdata/laboratorio/familias/` | FR-040, FR-041, SC-006 |
| Guiones congelados | FR-001 a FR-006, FR-010, FR-025, FR-032, FR-040, FR-041 |
| `.golangci.yml` (`producto` en misspell) | FR-022 (el nombre que fija el spec) |
| `CHANGELOG.md`, `docs/SKILLS.md` | FR-050 |
| `CONTRIBUTING.md`, ayuda de `make`, comentarios de `cmd/laboratorio` y del guion | FR-020, FR-030, FR-040: lo que hoy documentan deja de ser cierto (research D19) |

## Gates: controles mecánicos que añade o toca

- **`make ci`**: ningún objetivo nuevo; corre todo lo que sigue. **`make skills-check`**: su expresión de tests gana
  `TestRemisionesDelRepositorio`; nada más de su receta cambia.
- **Objetivos de `make`**: `evals-laboratorio` y `laboratorio-tabla`, con las mismas recetas (`TestRecetasDelSondeo`);
  cambia su ayuda.
- **Tests nuevos en `make ci`**: `TestFrases` y `TestRemisionesDelRepositorio` (`internal/skills`);
  `TestVariantesDeControl`, `TestSkillsDelProductoDesdeControl` y `TestVariantesDePlazosYTerritorio`
  (`internal/evals/laboratorio_familias_test.go`); `TestArnesDeLasFamilias` (`laboratorio_sondeo_test.go`);
  `TestTablaDeLasFamilias` y `TestLecturaConCd` (`laboratorio_tabla_test.go`).
- **Tests que cambian**: `TestVariantesDeNormativa` (salen «control es skills/normativa» y «directorios de la familia»,
  que cubre `TestVariantesDeControl`); `TestLaboratorioFueraDelProducto`, a las tres familias. Sin cambiar ningún valor
  esperado: `TestArnesDeNormativa`, `TestArnesDelLaboratorio`, `TestTablaDeNormativa`, `TestTablaDelLaboratorio`,
  `TestReglaTestigo`, `TestReglaTestigoDeNormativa`, `TestOrdenesDeBash`, `TestServidorDelLaboratorio`,
  `TestRecetasDelSondeo`, `TestSkillsDelRepositorio`, `TestVerticalDeEnsayo`.
- **Suites congeladas**: las dos de H41 (contracts/aceptacion.md); las de H39, H40 y H32, sin cambiar un fichero
  (FR-042; research V4).
- **Fixtures** (tareas `[datos]`, sin red): `internal/evals/testdata/laboratorio/familias/<skill>/sondeo/` (sesiones
  fijas de los dos modos y sus `sondeo.json`; en `plazos-y-territorio`, de las cinco variantes) y `…/tabla.md`, por
  familia; y `internal/evals/testdata/laboratorio/familias/mezcla/sondeo/h41/`, un sondeo de `jurisprudencia` con un
  directorio de `normativa`, sin `tabla.md`. Ninguna grabación; ningún fichero de `testdata/` existente cambia.
- **Fuera de `make ci`**: `TestSondeoDelLaboratorio` (`-tags evals`), la entrada de la persona.

## Datos externos

Ninguno (research, «Fuentes externas»): no hay manifiesto `grabaciones.json`, test `TestGrabar*` ni tarea que grabe;
las sesiones del sondeo se sirven de las grabaciones que ya usa el job, y el run no escribe en `evidencias/` ni en
`data/`.

## Controles de umbral

| Requisito | Umbral | Control | Dónde |
|---|---|---|---|
| SC-001, FR-001 a FR-003 | 100 % de las referencias de `normativa` y `jurisprudencia` con alguna remisión; 0 frases con enlace sin «lee» antes o con «si la tienes», «antes de escribir la respuesta» o «antes de dar la consulta»; 0 frases distintas de la lista del plan | La regla y la lista sobre `skills/`; el subtest «control», que tiene que dar los fallos de los `<skill>-control` (contracts/remisiones.md §7) | ci:internal/skills/remisiones_test.go:TestRemisionesDelRepositorio |
| SC-002, FR-008 | 0 skills de `skills/` con un cuerpo de más de 11 000 caracteres | El defecto `cuerpo de N caracteres (máximo 11000)`, con la skill y su recuento | ci:internal/app/skills_test.go:TestSkillsDelRepositorio |
| SC-003, FR-009 | 0 bytes de diferencia entre el frontmatter de `normativa` y `jurisprudencia` y el de su control; 0 derivas tras `make skills-sync` | El frontmatter, byte a byte; la regeneración en memoria contra el árbol | ci:internal/evals/laboratorio_familias_test.go:TestSkillsDelProductoDesdeControl, ci:internal/app/skills_test.go:TestSkillsDelRepositorio |
| SC-004, FR-006, FR-007 | los dos textos movidos en su referencia y fuera del cuerpo; 0 bytes de diferencia con el control fuera de lo declarado | Calcula cada skill desde su control con su declaración y la compara byte a byte | ci:internal/evals/laboratorio_familias_test.go:TestSkillsDelProductoDesdeControl |
| SC-005, FR-010 | 0 bytes de diferencia entre `skills/plazos-y-territorio` y su control; 0 ficheros de más o de menos | La comparación ruta a ruta, con lo generado | ci:internal/evals/laboratorio_familias_test.go:TestSkillsDelProductoDesdeControl |
| SC-001, SC-004, SC-005, SC-009, FR-001 a FR-006, FR-010, FR-025 | 100 % de las órdenes de contracts/aceptacion.md §2 en lo instalado; 0 remisiones de antes; lo movido en su referencia; la lista de control en `plazos-y-territorio`; 0 variantes instaladas | El guion congelado `h41-skills-instaladas.txtar` | ci:internal/app/e2e_test.go:TestEntregaDelHito |
| SC-006, FR-040, FR-041 | 0 diferencias en recuentos, turnos y extremos de Wilson de las sesiones fijas de las tres familias, también de las cinco variantes de `plazos-y-territorio`, con las lecturas con `cd` que cuentan y las que no | La salida, byte a byte, contra cada `tabla.md`; la regla, orden a orden; el guion congelado `h41-laboratorio-tabla.txtar` | ci:internal/evals/laboratorio_tabla_test.go:TestTablaDeLasFamilias, ci:internal/evals/laboratorio_tabla_test.go:TestLecturaConCd, ci:internal/app/e2e_test.go:TestEntregaDelHito |
| SC-007, FR-022, FR-032 | 0 bytes de diferencia entre la skill de una sesión de `<skill>-producto` y `skills/<skill>`; la lista de sus referencias en `sondeo.json` y en las columnas de la tabla; 0 directorios `<skill>-producto` en el repositorio | Las sesiones del sustituto de `claude`; la tabla con una referencia de la lista que la familia no declara; las carpetas de las familias | ci:internal/evals/laboratorio_sondeo_test.go:TestArnesDeLasFamilias, ci:internal/evals/laboratorio_tabla_test.go:TestTablaDeLasFamilias, ci:internal/evals/laboratorio_familias_test.go:TestVariantesDeControl |
| SC-008, FR-030, FR-033 | 0 sesiones con una petición inválida; n/4 por eval en cada variante y modo pedidos con una válida, también en las cuatro invocaciones de FR-033 | Sustituto de `claude` que anota sus invocaciones (contracts/arnes.md §5) | ci:internal/evals/laboratorio_sondeo_test.go:TestArnesDeLasFamilias |
| SC-009, FR-020 | 0 bytes de diferencia y 0 ficheros de más o de menos en cada `<skill>-control` respecto de `v0.9.0` | Las huellas SHA-256 por ruta | ci:internal/evals/laboratorio_familias_test.go:TestVariantesDeControl |
| SC-009, FR-021 | 0 bytes fuera de lo declarado entre cada variante de H40 y `normativa-control` | El cálculo de H40, sin cambiar sus valores | ci:internal/evals/laboratorio_normativa_test.go:TestVariantesDeNormativa |
| SC-009, FR-023 | 0 bytes fuera de lo declarado entre cada variante de `plazos-y-territorio` y su control; 0 segmentos distintos fuera de las órdenes de `municipios.md` y `calculo_de_plazos.md` | La declaración, byte a byte; las frases de los dos cuerpos sin esas órdenes (contracts/variantes.md §4) | ci:internal/evals/laboratorio_familias_test.go:TestVariantesDePlazosYTerritorio |
| SC-009, FR-025 | 0 ficheros de variantes versionadas en lo empotrado, en lo instalado y en lo que el job ejecuta o elige | Recorre lo empotrado, una instalación en un temporal, la definición del job, `evals/` y `rutaQueMide` | ci:internal/evals/laboratorio_test.go:TestLaboratorioFueraDelProducto |
| SC-009, FR-042 | 0 fallos en las suites congeladas de H39 y H40, sin cambiar un fichero suyo | Sus guiones activados | ci:internal/app/e2e_test.go:TestEntregaDelHito |
| SC-010, FR-051 | Las evals de `normativa` en verde en el cierre, en los dos modos: sus umbrales con `decide: true` | El job de evals del cierre | evals:normativa:sin_activar:claude-sonnet-5-5:orden, evals:normativa:sin_activar:claude-sonnet-5-5:herramienta, evals:normativa:redaccion_sin_fechar:claude-sonnet-5-5:orden, evals:normativa:redaccion_sin_fechar:claude-sonnet-5-5:herramienta, evals:normativa:afirma_lo_no_leido:claude-sonnet-5-5, evals:normativa:medida_del_juez:afirma_lo_no_leido:defectos_sin_marcar, evals:normativa:medida_del_juez:afirma_lo_no_leido:correctos_marcados, evals:normativa:duracion_de_las_sesiones:orden, evals:normativa:duracion_de_las_sesiones:herramienta, evals:normativa:duracion_del_juez:orden, evals:normativa:duracion_del_juez:herramienta |
| SC-010, FR-051 | Las evals de `jurisprudencia` en verde en el cierre, en los dos modos | El job de evals del cierre | evals:jurisprudencia:sin_activar:claude-sonnet-5-5:orden, evals:jurisprudencia:sin_activar:claude-sonnet-5-5:herramienta, evals:jurisprudencia:cita_sin_documento:claude-sonnet-5-5:orden, evals:jurisprudencia:cita_sin_documento:claude-sonnet-5-5:herramienta, evals:jurisprudencia:documento_sin_cotejar:claude-sonnet-5-5:orden, evals:jurisprudencia:documento_sin_cotejar:claude-sonnet-5-5:herramienta, evals:jurisprudencia:afirma_lo_no_leido:claude-sonnet-5-5, evals:jurisprudencia:medida_del_juez:afirma_lo_no_leido:defectos_sin_marcar, evals:jurisprudencia:medida_del_juez:afirma_lo_no_leido:correctos_marcados, evals:jurisprudencia:duracion_del_juez:orden, evals:jurisprudencia:duracion_del_juez:herramienta |
| SC-010, FR-051 | Las evals de `plazos-y-territorio` en verde en el cierre, en los dos modos | El job de evals del cierre | evals:plazos-y-territorio:sin_activar:claude-sonnet-5-5:orden, evals:plazos-y-territorio:sin_activar:claude-sonnet-5-5:herramienta, evals:plazos-y-territorio:plazo_sin_fecha:claude-sonnet-5-5:orden, evals:plazos-y-territorio:plazo_sin_fecha:claude-sonnet-5-5:herramienta, evals:plazos-y-territorio:cobertura_sin_declarar:claude-sonnet-5-5:orden, evals:plazos-y-territorio:cobertura_sin_declarar:claude-sonnet-5-5:herramienta, evals:plazos-y-territorio:vigencia_sin_declarar:claude-sonnet-5-5:orden, evals:plazos-y-territorio:vigencia_sin_declarar:claude-sonnet-5-5:herramienta, evals:plazos-y-territorio:redaccion_sin_fechar:claude-sonnet-5-5:orden, evals:plazos-y-territorio:redaccion_sin_fechar:claude-sonnet-5-5:herramienta, evals:plazos-y-territorio:lectura_sin_intentar:claude-sonnet-5-5:orden, evals:plazos-y-territorio:lectura_sin_intentar:claude-sonnet-5-5:herramienta |

SC-010 usa los umbrales del job tal como están (`.github/workflows/evals.yml` y las carpetas de evals), que el hito no
toca: las filas nombran los que publican con `decide: true` los informes del cierre de H40 (research V11), y el veredicto
del job sale en rojo con cualquiera de ellos. `consulta-juridica` no publica ninguno con `decide: true` y su skill no
cambia; su veredicto lo deciden sus series (ADR 0016), configuración que el hito no toca. Los 10 930 caracteres de FR-024
no son un umbral: el spec dice que ninguna comprobación sale en rojo con una variante por encima, y el plan da los
cuerpos y lo que movería H41.1 (contracts/variantes.md §3). SC-011 y SC-012 los mide la persona (capa 3): el sondeo abre
sesiones con modelo que ni `make ci` ni el run abren.

## Cierre y pendientes

- **FR-052**: el run no lanza el sondeo; la propuesta de cambio queda pendiente de sus tablas y su lectura, y el informe
  final lo dice; la propuesta no se fusiona sin ellas, y si `normativa` o `jurisprudencia` no cumplen, lo medido va a un
  ADR que dice qué medir después.
- **FR-053**: la tabla de `plazos-y-territorio` va a un ADR que escribe la persona, con la regla de FR-052 frente a
  `plazos-y-territorio-control` y el detalle por sesión; research D15 le da a qué condiciones llegan las evals en cada
  forma.
- **FR-051**: si la medición del cierre da un rojo, la reparación corrige la skill dentro de lo que este plan declara
  (contracts/remisiones.md) y el informe da sus cifras; ningún umbral se rebaja.

## Complexity Tracking

| Desviación | Por qué hace falta | Alternativa más simple rechazada y por qué |
|---|---|---|
| Dos copias más de una skill en `laboratorio/` (8 ficheros) y tres variantes (15) | FR-020: `<skill>-control` congelada por familia; FR-023: las tres formas que mide el sondeo | Instalarlas desde la etiqueta `v0.9.0` al lanzar el sondeo: `make ci` no podría comprobar que no cambian y el clon del cierre es superficial (research D9) |
| El cálculo de `normativa` y `jurisprudencia` desde control y su lista dejan fuera lo generado | FR-007 y FR-001 sin romper el principio X: un paquete nuevo cambia lo generado y no puede pedir un cambio de código | Compararlo también: pondría los tests en rojo con cada paquete; lo cubre `TestSkillsDelRepositorio` (research D7, D8) |
| La comparación de `plazos-y-territorio` incluye lo generado | FR-010 pide cada fichero idéntico byte a byte hasta H41.1 | Dejar fuera lo generado: no cumple FR-010. Mientras dure, un paquete que añadiera una línea a la región de `plazos-y-territorio` pondría el test en rojo; este hito no añade ninguno (research D4) |
| `skills.Frases`, exportada y usada solo por tests | Una sola definición de la frase de FR-001 para la regla de las remisiones (`internal/skills`) y para el control de FR-023 (`internal/evals`) | Dos implementaciones, una por paquete de test: dos definiciones que pueden divergir (research D5). `internal/skills` no va en el binario (research V10) |
