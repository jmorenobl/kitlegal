# Implementation Plan: H41.3 · Por qué `normativa` no se activa con todas las skills instaladas, y si unas descripciones cortas lo arreglan

**Branch**: `037-h41-3-por-que-normativa` | **Date**: 2026-11-16 | **Spec**: [spec.md](spec.md)

**Input**: Feature specification from `specs/037-h41-3-por-que-normativa/spec.md`

## Summary

El hito entrega lo que hace falta para medir si la caída de la eval 04 de `normativa` en modo herramienta es de la
instalación completa, si la causa es el presupuesto del listado de skills de Claude Code y si unas descripciones cortas
la quitan, sin tocar ninguna skill, eval, umbral ni versión del job. En `laboratorio/`, fuera del producto, la variante
`producto-descripciones-cortas`: cuatro textos de `description`, uno por skill de `skills/`, de 417, 420, 426 y 427 runas
(1690 frente a 3313 hoy), que `make ci` comprueba (FR-001, FR-002). El arnés del laboratorio gana
`INSTALACION=producto-descripciones-cortas`, que instala las cuatro skills de `skills/` con la `description` de la
variante y todo lo demás byte a byte, solo con `VARIANTES=normativa-producto` (FR-010, FR-011), y `PRESUPUESTO=<n>`, que
pone `SLASH_COMMAND_TOOL_CHAR_BUDGET=<n>` en el entorno de cada sesión (FR-013, FR-014); con cualquiera de las dos,
`sondeo.json` lo anota, y sin ellas todo queda como hoy (FR-012, FR-015). La tabla da, con esa constancia, la instalación
y el presupuesto de cada fila (FR-020 a FR-024). El run no abre sesiones con modelo: el sondeo de cuatro filas y 640
sesiones, su tabla en `evidencias/` y el ADR los hace la persona antes de fusionar (FR-030, FR-031).

Enfoque técnico (research D1 a D16): la variante, un fichero de texto por skill, leído como hoy se leen los textos de las
variantes de H39; una sola función que la comprueba para `make ci` y para el arnés; la composición, con `yaml.Marshal` de
la clave `description` en el lugar de sus líneas, con la misma clasificación de líneas que ya usa el laboratorio; la
variable, detrás del entorno cerrado de hoy; la constancia, dos claves con `omitempty`; y las dos columnas, solo con la
constancia, con la instalación delante de las skills instaladas en el orden de las filas.

## Technical Context

**Language/Version**: Go 1.27 (`go 1.27.0` y `toolchain go1.27.4` en `go.mod`; `go version` da `go1.27.4`, research
V13).

**Primary Dependencies**: las fijadas, sin ninguna nueva. Biblioteca estándar (`encoding/json`, `os`, `path/filepath`,
`slices`, `strconv`, `strings`, `unicode/utf8`, `iter`); `go.yaml.in/yaml/v3`, la biblioteca de YAML que el repositorio
fijó en H5 y que ya está en `go.mod` (research V11), para escribir la `description` compuesta; `stretchr/testify` y
`rogpeppe/go-internal` (testscript) en los tests.

**Storage**: ficheros: la variante en `laboratorio/producto-descripciones-cortas/`; la copia compuesta en el temporal del
arnés; `sondeo.json` y las sesiones de cada sondeo, en un `DIRECTORIO` de la persona fuera del repositorio; fixtures en
`internal/evals/testdata/laboratorio/familias/instalacion/`.

**Testing**: `go test -race` en `make ci`; un guion testscript congelado; tests de tabla con `t.Parallel()`; árboles de
prueba en `t.TempDir()`; el sustituto de `claude` de `TestArnesDelLaboratorio`, que además vuelca su entorno; ningún test
con red ni con modelo. Skills aplicadas: research, cabecera.

**Target Platform**: macOS y Linux, donde corren `make ci`, el job de evals y el sondeo de la persona.

**Project Type**: programas y tests del repositorio (`internal/evals`, `cmd/laboratorio`, guiones de `scripts/`), una
función pura exportada en `internal/skills` y cuatro textos en `laboratorio/`; el binario `kitlegal` no cambia de órdenes
ni de salidas, y las skills no cambian.

**Performance Goals**: ninguno. El arnés lee cuatro ficheros más y reescribe cuatro `SKILL.md` una vez por sondeo; la
tabla lee dos claves más de cada `sondeo.json` que ya lee.

**Constraints**: las suites congeladas de H39 a H41.2 siguen en verde sin cambiar un fichero (FR-032); `skills/` no
cambia (FR-033); el run no graba, no escribe en `evidencias/` y no abre sesiones con modelo (FR-017, FR-031).

**Scale/Scope**: 4 ficheros de texto; 1 fichero nuevo de código y su test; 1 función exportada; 1 valor nuevo de un
argumento y 1 argumento nuevo; 2 claves nuevas en `sondeo.json`; 2 columnas en la tabla de las familias del producto; 1
juego de fixtures; 1 guion congelado; 4 filas y 640 sesiones en la aceptación de la persona.

## Constitution Check

*GATE: Must pass before Phase 0 research. Re-check after Phase 1 design.*

| Principio | Cómo lo cumple este hito |
|---|---|
| I · Fuentes públicas y frontera humana (INNEGOCIABLE) | No toca ninguna fuente ni `internal/httpx` y no añade ninguna petición. Las sesiones del sondeo se sirven de las grabaciones del job, con los proxies que rechazan toda petición, como en H39 a H41.2. Nada se presenta ni se firma; el CENDOJ no se toca |
| II · Nada sin cita ni fuente | El sobre de ningún applet cambia y ninguna skill cambia (FR-033): sus reglas de cita, sus formas fijas y su declaración de lo no cubierto siguen como están. La variante solo cambia la `description` de las copias del laboratorio |
| III · Tests primero y offline | Primero el guion de contracts/aceptacion.md, que hoy falla por una aserción. Todos los tests nuevos, offline y sin modelo, contra árboles de prueba en `t.TempDir()` o contra fixtures que escribe una tarea `[datos]`; ninguna grabación. `internal/core` no cambia y la cobertura global no baja: el código nuevo lleva sus tests |
| IV · Arquitectura hexagonal con reglas ejecutables | El código va en `internal/evals`, que no va en el binario; `cmd/laboratorio` sigue siendo la raíz que nombra `os.Std*` y llama a `os.Exit`. Errores: los de los argumentos, `errorDeUso` juntos como hoy (el guion sale con 1); una variante que no se puede leer, error del arnés sin `errorDeUso` (registro de `go test`, salida 1); en la tabla, defecto `inesperado` con 1. `internal/skills` gana una función pura, sin E/S. `kitlegal` y sus códigos no cambian. Ningún `panic` |
| V · Simplicidad y dependencias fijadas (YAGNI) | Ninguna dependencia nueva (`go.yaml.in/yaml/v3` es la biblioteca de YAML fijada en H5). Cada mecanismo se traza a un FR («Trazabilidad»). Lo que llega manipulado —una variante o un `sondeo.json` editados a mano, un fichero ajeno en la variante— es la regla genérica del spec, sin comprobación, mensaje ni test propios más allá de lo que FR-002 pide. Las desviaciones, en Complexity Tracking |
| VI · Un binario, convenciones de agente | Un solo binario `kitlegal`, sin banderas, variables ni órdenes nuevas. Los dos cambios de argumentos son de `make evals-laboratorio`, una orden del repositorio; `cmd/laboratorio` no se distribuye y no cambia de órdenes. `SLASH_COMMAND_TOOL_CHAR_BUDGET` solo se pone en las sesiones del laboratorio, nunca en el equipo de quien instala kitlegal |
| VII · Grafo y privacidad | No toca el grafo. `sondeo.json` gana el nombre de una instalación y un número; ningún dato de una persona |
| VIII · Skills primero; el binario es la herramienta | Protege las skills que ya existen: mide si la caída de `sin_activar` de `normativa` es de la instalación, del presupuesto del listado o de la longitud de las descripciones antes de que el hito siguiente cambie las del producto (spec, «Uso, de fuera adentro»). Ninguna skill cambia ni gana `scripts/`; sus evals se miden en el cierre (FR-034) |
| IX · Genericidad territorial, validación local | Ningún dato ni código nombra un municipio; las evals del sondeo son las de hoy |
| X · Genericidad por materia, validación administrativa | Ningún caso especial para una materia: la comprobación y la composición recorren las skills que haya en `skills/`, sean las que sean; los textos son uno por skill de capacidad. La única regla con nombre, `normativa-producto`, es de una skill de capacidad, vive en el laboratorio, fuera del producto, y la fija FR-010. La eval 04 solo está en las órdenes de la persona |
| Regla 1 · `internal/core/**` no importa adaptadores | `internal/core` no cambia |
| Regla 2 · solo `internal/httpx` importa `net/http` | Nada nuevo importa `net/http` |
| Regla 3 · solo `internal/{cache,store,graph}` importan SQLite | Nada nuevo importa `database/sql` ni `modernc.org/sqlite` |
| Regla 4 · solo `internal/cli` y `cmd/` llaman a `os.Exit` | Solo `cmd/laboratorio/main.go`, como hoy |
| Regla 5 · solo `internal/render` escribe en stdout; logs a stderr | `internal/evals` escribe en ficheros del temporal y de `DIRECTORIO` y en el `io.Writer` que le inyecta `cmd/laboratorio`, como hoy; nada nuevo nombra `os.Stdout` |
| Regla 6 · `internal/graph` no importa `internal/source/*` ni `internal/render` | `internal/graph` no cambia |
| Regla 7 · solo `internal/mcp` importa el SDK de MCP | Nada nuevo lo importa: el servidor de las sesiones es el `kitlegal mcp serve` de hoy |
| Applets de ejemplo fuera del binario (ADR 0010; `depguard`, regla `ejemplo`) | Nada nuevo importa `internal/app/ejemplo` |
| Garantías de `internal/arch_test.go` (`TestArquitectura`, `TestDependenciasDelBinario` y las demás) | Ninguna importación del cierre de `cmd/kitlegal` cambia: `internal/skills` gana una función sin importaciones nuevas, y `go.yaml.in/yaml/v3` entra en `internal/evals`, que no está en ese cierre |

Sin violaciones sin justificar. Re-evaluado tras el diseño de la fase 1: igual.

## Project Structure

### Documentation (this feature)

```text
specs/037-h41-3-por-que-normativa/
├── plan.md
├── research.md
├── data-model.md
├── quickstart.md
├── contracts/
│   ├── variante.md        # FR-001, FR-002: los cuatro textos y la comprobación
│   ├── arnes.md           # FR-010 a FR-018: INSTALACION=producto-descripciones-cortas, PRESUPUESTO, sondeo.json
│   ├── tabla.md           # FR-020 a FR-024: las dos columnas y su ejemplo
│   └── aceptacion.md      # el guion congelado
├── aceptacion/laboratorio-tabla.txtar   # la primera tarea; congelado
└── tasks.md
```

### Source Code (repository root)

```text
laboratorio/producto-descripciones-cortas/consulta-juridica.txt     # NUEVO: los textos de contracts/variante.md §2
laboratorio/producto-descripciones-cortas/jurisprudencia.txt        # NUEVO
laboratorio/producto-descripciones-cortas/normativa.txt             # NUEVO
laboratorio/producto-descripciones-cortas/plazos-y-territorio.txt   # NUEVO
internal/skills/frontmatter.go                 # PrimeraEtiquetaXML, que defectosDeLaEtiqueta usa
internal/skills/frontmatter_test.go            # TestPrimeraEtiquetaXML
internal/evals/laboratorio_descripciones.go    # NUEVO: descripcionesCortas, skillConLaDescripcion, componerLasDescripciones, límites y carpeta
internal/evals/laboratorio_descripciones_test.go # NUEVO: TestDescripcionesCortas, TestConLaDescripcion
internal/evals/laboratorio.go                  # la clasificación de las líneas de descripcionYCuerpo, en una función iteradora (lineasDelSkill) que comparte con skillConLaDescripcion
internal/evals/laboratorio_familias.go         # varianteDelProductoDeNormativa e instalacionConLasDescripcionesCortas, junto a la familia de normativa
internal/evals/laboratorio_sondeo.go           # PRESUPUESTO; el valor nuevo de INSTALACION; comprobar con la raíz; la composición en preparar; el entorno; la constancia
internal/evals/laboratorio_sondeo_test.go      # el sustituto vuelca su entorno; TestArnesConLasDescripcionesCortas, TestArnesConElPresupuesto; TestRecetasDelSondeo y TestGuionDelLaboratorio con diez argumentos
internal/evals/laboratorio_tabla.go            # sondeoLeido; claveDeFila; conLaInstalacion; las dos columnas; instalacionDeLaFila; compararClaves
internal/evals/laboratorio_tabla_test.go       # TestTablaConLaInstalacion
internal/evals/job_test.go                     # TestSondeoDelLaboratorio (-tags evals): -presupuesto
internal/evals/testdata/laboratorio/familias/instalacion/            # NUEVO [datos]: sondeo/{h41-2,h41-3-control,h41-3-diagnostico,h41-3-remedio,h41-3-sola}/ y tabla.md
scripts/evals-laboratorio.sh                   # diez argumentos; -presupuesto
scripts/evals-sondeo-llavero.sh                # comentario: diez argumentos
Makefile                                       # receta de evals-laboratorio con "$(PRESUPUESTO)"; su comentario
CONTRIBUTING.md                                # «El sondeo del laboratorio»: la variante, los dos argumentos, las columnas
CHANGELOG.md                                   # Unreleased
```

**Structure Decision**: la de `docs/ROADMAP.md` §2 y `CLAUDE.md`. El código, con el del arnés y la tabla, en
`internal/evals`, que no va en el binario; lo de la variante, en un fichero propio del paquete; la variante, en
`laboratorio/`, fuera de `skills/`, de `evals/` y de lo que se empotra. Ningún paquete nuevo. `skills/`, `evals/`,
`data/` y `.github/workflows/evals.yml` no se tocan.

## Diseño

- **La variante** (contracts/variante.md; data-model §1; research D1 a D4): cuatro ficheros de una línea; la comprobación
  `descripcionesCortas`, que da los defectos de FR-002 en el orden y con las líneas de contracts/variante.md §3, con la
  definición de etiqueta XML de `internal/skills`.
- **El arnés** (contracts/arnes.md; data-model §2 a §5; research D5 a D9 y D13): el valor nuevo de `INSTALACION` y sus
  reglas; `PRESUPUESTO`, décimo argumento; la composición tras la copia de `INSTALACION=producto`; la variable detrás del
  entorno cerrado de hoy; las dos claves de `sondeo.json`. Orden de trabajo: el sustituto que vuelca su entorno y el
  caso sin `PRESUPUESTO`, con la variable en el entorno de quien lanza —ninguna sesión la tiene, y el entorno de cada
  una es, valor a valor, el que componen `entornoDelSondeo`, `entornoDeLaSesion` y `sinElDirectorioEnElPATH`, que el hito
  no cambia (contracts/arnes.md §6)—, se escriben y pasan con el arnés de hoy, antes de añadir la opción; así lo que fija
  «hoy» no sale del código nuevo.
- **La tabla** (contracts/tabla.md; data-model §6; research D10 y D11): con la constancia, la instalación y el presupuesto
  en la clave, en la cabecera y en cada fila, y en el orden de las filas delante de las skills instaladas; sin ella, la de
  hoy.

**Aceptación e2e:** un guion testscript que escribe la primera tarea en `specs/037-h41-3-por-que-normativa/aceptacion/` y
queda congelado (contracts/aceptacion.md), activado como `internal/app/testdata/script/h41-3-laboratorio-tabla.txtar` y
ejecutado por `TestEntregaDelHito`:
- `aceptacion/laboratorio-tabla.txtar` (historia 3): `laboratorio tabla` sobre los cuatro sondeos fijos de
  `normativa-producto` en modo herramienta del sondeo de H41.3 —sola, control, diagnóstico con `presupuesto` y remedio con
  `instalacion`, misma versión— y uno de H41.2 sin la constancia da su tabla y su detalle byte a byte: una fila por
  instalación y presupuesto, en el orden de FR-021, con «Instalación» y «Presupuesto», sesiones que activan `normativa`
  antes, solo otra skill y ninguna, una sin terminar en la de remedio, y la fila de H41.2 con la instalación deducida y
  su única sesión sin terminar (`n` = 0). Cubre FR-020, FR-021, FR-022, FR-024; SC-006.

La historia 1 no tiene una orden del binario ni del programa del laboratorio: su aceptación en `make ci` es
`TestDescripcionesCortas` (FR-001, FR-002; SC-001). La 2 abre sesiones con modelo y credencial: su aceptación en `make ci`
son `TestArnesConLasDescripcionesCortas`, `TestArnesConElPresupuesto` y `TestConLaDescripcion`, con el sustituto de
`claude` (FR-010 a FR-018; SC-002 a SC-005). La 4 la hace la persona (FR-030, FR-031; quickstart.md §7).

## Uso, de fuera adentro

Ninguna `SKILL.md` del producto ni el binario que se distribuye consume nada de lo que entrega el hito: 0 invocaciones por
pregunta. Cada salida, su lector, su tamaño y cuándo deja de darse; ninguna crece con el uso del producto —cientos de
normas y miles de bloques consultados no cambian ninguna—, solo con las filas y las sesiones que la persona pide:

| Salida | Quién la lee y cuántas veces | Forma y ejemplo | Tamaño | Deja de darse |
|---|---|---|---|---|
| Los cuatro textos de la variante | el modelo de cada sesión de la fila de remedio, una vez, en el listado del arranque: 160 en el sondeo | una línea por fichero (contracts/variante.md §2) | 1690 runas, 1732 bytes, fijo | cuando el ADR que lee el sondeo decide; si el hito siguiente los aplica, la comprobación sigue exigiendo uno por skill |
| La comprobación de la variante | quien hace `make ci`, solo en rojo | `laboratorio/producto-descripciones-cortas: falta la descripción de quinta` (contracts/variante.md §3) | una línea por defecto: como mucho dos por skill, una por entrada de más y la de la suma | al corregir la variante; vuelve si un hito añade una skill sin su texto |
| Los errores de los argumentos | la persona, cuando la orden no vale | `PRESUPUESTO: «0100» no es un entero mayor que 0` (contracts/arnes.md §2) | una línea por argumento, y una por defecto de la variante | cuando la orden vale |
| La constancia de `sondeo.json` | el guion de la tabla, una vez por sondeo | `,"presupuesto":1000000`, `,"instalacion":"producto-descripciones-cortas"` (contracts/arnes.md §5) | 22 y 46 bytes, una vez por sondeo, solo con la opción | no se escribe sin las opciones |
| Las dos columnas de la tabla | la persona, una vez, para la lectura y el ADR | contracts/tabla.md §4 | 85 bytes en las cuatro filas del sondeo, más 37 en la cabecera y el separador; el detalle, unos 93 KB, no cambia | solo salen con la constancia; dejan de leerse cuando el ADR está escrito |

## Trazabilidad

| Mecanismo | FR/SC |
|---|---|
| Los cuatro ficheros de `laboratorio/producto-descripciones-cortas/` | FR-001, FR-033 |
| `descripcionesCortas` y sus defectos | FR-002, FR-016, SC-001 |
| `skills.PrimeraEtiquetaXML`, que usa `defectosDeLaEtiqueta` | FR-002 («la definición de ADR 0046, punto 3») |
| `TestDescripcionesCortas` | FR-002, SC-001 |
| La clasificación de líneas compartida por `descripcionYCuerpo` y `skillConLaDescripcion` | FR-011 («cada otro byte») |
| `skillConLaDescripcion` con `yaml.Marshal`; `componerLasDescripciones` en `preparar` | FR-010, FR-011 |
| `TestConLaDescripcion` | FR-011 (caso límite de los caracteres que YAML cita), SC-002 |
| `varianteDelProductoDeNormativa`, `instalacionConLasDescripcionesCortas` y sus reglas en `errorDeLaInstalacion` | FR-010, FR-016 |
| `comprobar` con la raíz; la variante, solo leída con el valor nuevo | FR-015, FR-016 |
| `PRESUPUESTO`: `ArgumentosDelLaboratorio.Presupuesto`, su forma, su error con H39 | FR-013, FR-016 |
| La receta, `scripts/evals-laboratorio.sh`, `-presupuesto` en `TestSondeoDelLaboratorio`, comentario del llavero | FR-013, FR-017 |
| `SLASH_COMMAND_TOOL_CHAR_BUDGET` detrás del entorno de hoy | FR-013, FR-014 |
| `instalacion` y `presupuesto` en `sondeoLeido` | FR-012, FR-020 |
| El sustituto de `claude` vuelca su entorno | FR-018 |
| `TestArnesConLasDescripcionesCortas` | FR-010, FR-011, FR-016, FR-018, SC-002, SC-005 |
| `TestArnesConElPresupuesto` | FR-012, FR-013, FR-014, FR-015, FR-016, FR-018, SC-003, SC-004, SC-005 |
| `TestRecetasDelSondeo` y `TestGuionDelLaboratorio` con diez argumentos | FR-013 |
| `conLaInstalacion`; la instalación y el presupuesto en `claveDeFila`, solo con la constancia; `compararClaves` | FR-020, FR-021, FR-022 |
| `instalacionDeLaFila` | FR-020, FR-022 |
| Las columnas «Instalación» y «Presupuesto» | FR-020, FR-021 |
| `TestTablaConLaInstalacion` y `testdata/laboratorio/familias/instalacion/` | FR-020, FR-022, FR-024, SC-006 |
| Guion congelado | FR-020, FR-021, FR-022, FR-024, SC-006 |
| `CONTRIBUTING.md`, comentario del `Makefile`, cabeceras de los guiones, `CHANGELOG.md` | FR-010, FR-013, FR-020: lo que hoy documentan deja de ser cierto (research D14) |
| Líneas de `gates/supuestos.md` del plan, entre ellas la del sondeo pendiente | FR-031 (research D15) |

## Gates: controles mecánicos que añade o toca

- **`make ci`**: ningún objetivo nuevo; corre todo lo que sigue.
- **Objetivos de `make`**: `evals-laboratorio` gana `"$(PRESUPUESTO)"` en su receta (`TestRecetasDelSondeo`) y cambia su
  comentario. `laboratorio-tabla`, sin cambios.
- **Tests nuevos en `make ci`**: `TestDescripcionesCortas` y `TestConLaDescripcion` (`laboratorio_descripciones_test.go`);
  `TestArnesConLasDescripcionesCortas` y `TestArnesConElPresupuesto` (`laboratorio_sondeo_test.go`);
  `TestTablaConLaInstalacion` (`laboratorio_tabla_test.go`); y `TestPrimeraEtiquetaXML`
  (`internal/skills/frontmatter_test.go`, T002).
- **Tests que cambian**: `TestRecetasDelSondeo`, la orden del laboratorio con `PRESUPUESTO=1000000` y `"1000000"`
  detrás de `"04"`; `TestGuionDelLaboratorio`, diez argumentos en sus casos, `ordenDelLaboratorio` con `-presupuesto`
  delante de `-temporal`, y el caso de uso con nueve, que nombra `<presupuesto>` y no ejecuta go; el sustituto de `claude`
  deja además `entorno.XXXXXX`, que ningún test de hoy lee. Sin cambiar ningún valor esperado: `TestArnesDelLaboratorio`,
  `TestArnesDeNormativa`, `TestArnesDeLasFamilias`, `TestArnesConLaInstalacionDelProducto`, `TestArnesConUnaEval`,
  `TestTablaConLasActivaciones`, `TestTablaDeLasFamilias`, `TestTablaDeNormativa`, `TestTablaDelLaboratorio`,
  `TestLaboratorioFueraDelProducto`, los de `internal/skills` (frontmatter) y los de H39 que usan `descripcionYCuerpo`.
  Fuera del plan, `TestTopeDeLaSesion` (`internal/evals/sesiones_test.go`) escucha SIGHUP en cada subtest antes de abrir
  la sesión, sin cambiar sus valores esperados, para que `muere-por-otra-senal` no dependa de que el proceso de test la
  herede ignorada (`gates/supuestos.md`, reparador T001).
- **Suites congeladas**: la de H41.3 (contracts/aceptacion.md); las de H39 a H41.2, sin cambiar un fichero (FR-032).
- **Fixtures** (tarea `[datos]`, sin red): `internal/evals/testdata/laboratorio/familias/instalacion/` (cinco sondeos y su
  `tabla.md`, contracts/tabla.md §6). Ninguna grabación; ningún fichero de `testdata/` existente cambia.
- **Fuera de `make ci`**: `TestSondeoDelLaboratorio` (`-tags evals`), la entrada de la persona.

## Datos externos

Ninguno (research, «Fuentes externas»): no hay manifiesto `grabaciones.json`, test `TestGrabar*` ni tarea que grabe; los
cuatro textos de la variante los fija este plan, no salen de ninguna fuente; las sesiones del sondeo se sirven de las
grabaciones que ya usa el job; el run no escribe en `evidencias/` ni en `data/`.

## Controles de umbral

| Requisito | Umbral | Control | Dónde |
|---|---|---|---|
| SC-001, FR-001, FR-002 | 1 descripción por skill de `skills/` y 0 de más; cada una ≤ 450 runas; todas ≤ 1700; 0 etiquetas XML | La variante del repositorio, sin ningún defecto; cada caso roto (falta, sobra, 451, 1701, `<ejemplo>`, quinta skill), su línea exacta | ci:internal/evals/laboratorio_descripciones_test.go:TestDescripcionesCortas |
| SC-002, FR-010, FR-011, FR-018 | 4 skills por sesión, 0 repetidas; el 100 % de las `description` iguales a las de la variante; 0 bytes de diferencia con `skills/` en todo lo demás; 0 defectos de frontmatter en cada `SKILL.md` compuesto | El sustituto de `claude`: las skills de cada sesión, fichero a fichero; `skills.LeerFrontmatter` y `skills.ValidarFrontmatter` sobre cada `SKILL.md` compuesto, también con textos con `: `, comillas y `#` | ci:internal/evals/laboratorio_sondeo_test.go:TestArnesConLasDescripcionesCortas, ci:internal/evals/laboratorio_descripciones_test.go:TestConLaDescripcion |
| SC-003, FR-013, FR-014 | Con `PRESUPUESTO=N`, el 100 % de las sesiones con `SLASH_COMMAND_TOOL_CHAR_BUDGET=N` y los mismos nombres de variables que sin él más esa; sin él, 0 sesiones con la variable, aunque esté en el entorno de quien lanza | El entorno que vuelca el sustituto de cada sesión, con `=500` en el entorno de quien lanza | ci:internal/evals/laboratorio_sondeo_test.go:TestArnesConElPresupuesto |
| SC-004, FR-012, FR-015 | Sin las dos opciones, 0 bytes de diferencia con hoy en las skills, las sesiones, las invocaciones, el entorno, `sondeo.json`, la salida y los mensajes de una petición que no vale | Los tests del arnés de hoy con sus valores esperados sin cambiar, para todo salvo el entorno, que no leen; el entorno, en el caso sin `PRESUPUESTO` de `TestArnesConElPresupuesto`, escrito y en verde con el arnés de hoy: el que vuelca cada sesión es, nombre a nombre y valor a valor, el que componen `entornoDelSondeo`, `entornoDeLaSesion` y `sinElDirectorioEnElPATH` con el entorno del test, salvo `PWD`, `SHLVL` y `_` (contracts/arnes.md §6) | ci:internal/evals/laboratorio_sondeo_test.go:TestArnesDelLaboratorio, ci:internal/evals/laboratorio_sondeo_test.go:TestArnesDeNormativa, ci:internal/evals/laboratorio_sondeo_test.go:TestArnesDeLasFamilias, ci:internal/evals/laboratorio_sondeo_test.go:TestArnesConLaInstalacionDelProducto, ci:internal/evals/laboratorio_sondeo_test.go:TestArnesConUnaEval, ci:internal/evals/laboratorio_sondeo_test.go:TestArnesConElPresupuesto |
| SC-005, FR-016 | 0 sesiones abiertas, nada construido y `DIRECTORIO` sin crear con cada petición inválida de FR-016, con su mensaje | Las peticiones de contracts/arnes.md §6, y las de `INSTALACION` de hoy | ci:internal/evals/laboratorio_sondeo_test.go:TestArnesConLasDescripcionesCortas, ci:internal/evals/laboratorio_sondeo_test.go:TestArnesConElPresupuesto, ci:internal/evals/laboratorio_sondeo_test.go:TestArnesConLaInstalacionDelProducto |
| SC-006, FR-020, FR-021, FR-024 | 0 diferencias en filas, instalaciones, presupuestos, recuentos y extremos de Wilson con un decimal de los sondeos fijos; la instalación de una fila sin la constancia, en sus cuatro casos | La salida, byte a byte, contra `tabla.md`; los cuatro casos de `instalacionDeLaFila`; el guion congelado | ci:internal/evals/laboratorio_tabla_test.go:TestTablaConLaInstalacion, ci:internal/app/e2e_test.go:TestEntregaDelHito |
| SC-007, FR-022, FR-023, FR-032 | 0 bytes de diferencia en la tabla y el detalle de los sondeos fijos de H39 a H41.2; 0 fallos en sus suites congeladas, sin cambiar un fichero suyo | Sus tablas fijas; sus guiones activados | ci:internal/app/e2e_test.go:TestEntregaDelHito, ci:internal/evals/laboratorio_tabla_test.go:TestTablaConLasActivaciones, ci:internal/evals/laboratorio_tabla_test.go:TestTablaDeLasFamilias, ci:internal/evals/laboratorio_tabla_test.go:TestTablaDeNormativa, ci:internal/evals/laboratorio_tabla_test.go:TestTablaDelLaboratorio |
| SC-008, FR-034 | Las evals de `normativa` en verde en el cierre, en los dos modos, con `sin_activar` ≤ 0: sus umbrales con `decide: true` | El job de evals del cierre | evals:normativa:sin_activar:claude-sonnet-5-5:orden, evals:normativa:sin_activar:claude-sonnet-5-5:herramienta, evals:normativa:redaccion_sin_fechar:claude-sonnet-5-5:orden, evals:normativa:redaccion_sin_fechar:claude-sonnet-5-5:herramienta, evals:normativa:afirma_lo_no_leido:claude-sonnet-5-5, evals:normativa:medida_del_juez:afirma_lo_no_leido:defectos_sin_marcar, evals:normativa:medida_del_juez:afirma_lo_no_leido:correctos_marcados, evals:normativa:duracion_de_las_sesiones:orden, evals:normativa:duracion_de_las_sesiones:herramienta, evals:normativa:duracion_del_juez:orden, evals:normativa:duracion_del_juez:herramienta |
| SC-008, FR-034 | Las evals de `jurisprudencia` en verde en el cierre, en los dos modos | El job de evals del cierre | evals:jurisprudencia:sin_activar:claude-sonnet-5-5:orden, evals:jurisprudencia:sin_activar:claude-sonnet-5-5:herramienta, evals:jurisprudencia:cita_sin_documento:claude-sonnet-5-5:orden, evals:jurisprudencia:cita_sin_documento:claude-sonnet-5-5:herramienta, evals:jurisprudencia:documento_sin_cotejar:claude-sonnet-5-5:orden, evals:jurisprudencia:documento_sin_cotejar:claude-sonnet-5-5:herramienta, evals:jurisprudencia:afirma_lo_no_leido:claude-sonnet-5-5, evals:jurisprudencia:medida_del_juez:afirma_lo_no_leido:defectos_sin_marcar, evals:jurisprudencia:medida_del_juez:afirma_lo_no_leido:correctos_marcados, evals:jurisprudencia:duracion_del_juez:orden, evals:jurisprudencia:duracion_del_juez:herramienta |
| SC-008, FR-034 | Las evals de `plazos-y-territorio` en verde en el cierre, en los dos modos | El job de evals del cierre | evals:plazos-y-territorio:sin_activar:claude-sonnet-5-5:orden, evals:plazos-y-territorio:sin_activar:claude-sonnet-5-5:herramienta, evals:plazos-y-territorio:plazo_sin_fecha:claude-sonnet-5-5:orden, evals:plazos-y-territorio:plazo_sin_fecha:claude-sonnet-5-5:herramienta, evals:plazos-y-territorio:cobertura_sin_declarar:claude-sonnet-5-5:orden, evals:plazos-y-territorio:cobertura_sin_declarar:claude-sonnet-5-5:herramienta, evals:plazos-y-territorio:vigencia_sin_declarar:claude-sonnet-5-5:orden, evals:plazos-y-territorio:vigencia_sin_declarar:claude-sonnet-5-5:herramienta, evals:plazos-y-territorio:redaccion_sin_fechar:claude-sonnet-5-5:orden, evals:plazos-y-territorio:redaccion_sin_fechar:claude-sonnet-5-5:herramienta, evals:plazos-y-territorio:lectura_sin_intentar:claude-sonnet-5-5:orden, evals:plazos-y-territorio:lectura_sin_intentar:claude-sonnet-5-5:herramienta |

SC-008 usa los umbrales del job tal como están (`.github/workflows/evals.yml` y las carpetas de evals), que el hito no
toca: las filas nombran los que el cierre de H41.2 dio como publicados con `decide: true` (research V12), y el veredicto
del job sale en rojo con cualquiera de ellos. `consulta-juridica` no publica ninguno con `decide: true` y su skill no
cambia; su veredicto lo deciden sus series (ADR 0016), configuración que el hito no toca. SC-009 lo mide la persona
(capa 3): el sondeo abre sesiones con modelo que ni `make ci` ni el run abren.

## Cierre y pendientes

- **FR-031**: el run no lanza el sondeo ni escribe el ADR; `gates/supuestos.md` lo deja dicho, y el informe final lo da
  en «Supuestos y pendientes»: la propuesta queda pendiente de la tabla de las cuatro filas en `evidencias/` y de su
  lectura, que no deciden la fusión.
- **FR-034**: si la medición del cierre da en rojo `sin_activar` de la eval 04 de `normativa` en modo herramienta, la
  reparación no toca ninguna skill, eval, umbral ni versión del job: el informe final da el umbral sin cumplir con sus
  cifras y la deja dicha para la persona.

## Complexity Tracking

| Desviación | Por qué hace falta | Alternativa más simple rechazada y por qué |
|---|---|---|
| `internal/skills` exporta `PrimeraEtiquetaXML` para un uso del laboratorio | FR-002 pide la definición de etiqueta XML de ADR 0046, punto 3, que es una sola expresión en `internal/skills` (research D4) | Copiarla en `internal/evals`: dos definiciones que pueden divergir |
| `descripcionYCuerpo` pasa a usar una clasificación de líneas compartida | FR-011: lo que no es la `description` tiene que ser lo mismo para la composición y para el test que la compara; los tests de H39 que la usan siguen en verde sin cambiar (research D5) | Una segunda clasificación en `skillConLaDescripcion`: puede divergir de la que mide el test |
| `TestRecetasDelSondeo` y `TestGuionDelLaboratorio` cambian sus valores esperados | FR-013 pide un argumento más, que cambia la orden de la receta y el número de argumentos del guion (research D7) | Tomar el presupuesto del entorno de quien lanza: es lo que FR-014 prohíbe |
| El sustituto de `claude` anota también su entorno | FR-018 pide el sustituto «que anota sus invocaciones y su entorno»; las anotaciones de hoy se leen por su prefijo y no cambian (research D13) | Un test unitario de la función del entorno: no ve lo que llega a la sesión |
| La instalación y el presupuesto solo entran en la clave de la fila con la constancia | FR-022 (la tabla de hoy byte a byte) con FR-021 y US3 (el orden `sola`, `control`, `diagnóstico`, `remedio`) (research D10) | Llenarlos siempre: la instalación deducida pondría `normativa` antes que las cuatro y cambiaría la tabla congelada de H41.2 |
| `internal/evals` importa `go.yaml.in/yaml/v3` fuera de sus tests | FR-011: la `description` leída tiene que ser el texto exacto, también con lo que YAML cita o no admite sin escapar (research D5, V3, V15). No es una dependencia nueva: es la biblioteca de YAML que el repositorio usa desde H5 en lugar de `gopkg.in/yaml.v3`, ya en `go.mod` y en el cierre del binario (`internal/arch_test.go:221-227`; research V16) | `json.Marshal` como cadena entrecomillada: no escapa lo que YAML no admite sin escapar |
