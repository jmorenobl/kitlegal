---

description: "Tareas de H41.3 · Por qué normativa no se activa con todas las skills instaladas, y si unas descripciones cortas lo arreglan"
---

# Tasks: H41.3 · Por qué `normativa` no se activa con todas las skills instaladas, y si unas descripciones cortas lo arreglan

**Input**: `specs/037-h41-3-por-que-normativa/` — [spec.md](spec.md), [plan.md](plan.md), [research.md](research.md),
[data-model.md](data-model.md), [quickstart.md](quickstart.md) y [contracts/](contracts/) (`variante.md`, `arnes.md`,
`tabla.md`, `aceptacion.md`).

**Tests**: obligatorios (constitución III). La primera tarea escribe la suite de aceptación, que el workflow congela;
cada tarea posterior trae su test y la implementación que lo pone en verde, y deja `make ci` en verde por sí sola.

**Organización**: una fase por historia, en el orden en que se ejecutan una tras otra. La línea de cada tarea declara
todas las rutas que crea o modifica (el guardián rechaza cualquier otra; `go.mod`, `go.sum`, `CHANGELOG.md` y el
directorio del feature siempre se admiten, y declarar `x.go` admite `x_test.go`).

## Format: `[ID] [P?] [Story] Descripción`

- **[P]**: puede ir en paralelo (ficheros distintos, sin dependencia de una tarea sin terminar). Ninguna lo lleva.
- **[Story]**: la historia de [spec.md](spec.md) a la que sirve (US1 a US3; US4 no tiene tareas: se hace fuera del run,
  FR-030 y FR-031).
- **[aceptacion]**: la suite congelada (solo T001). **[datos]**: la tarea toca ficheros bajo `testdata/` (solo T010).

## Convenciones de todas las tareas

- `make ci` en verde al terminar cada una, en primer plano y por objetivos si hace falta (`make fmt-check`,
  `make lint`, `make test`, `make test-integration`, `make vuln`, `make schema-check`, `make skills-check`…). Ninguna
  usa la red más que la de las herramientas de Go, ninguna graba (no hay datos externos: plan.md «Datos externos»; ni
  manifiesto `grabaciones.json` ni `TestGrabar*`), ninguna escribe en `data/` ni fuera del repositorio y de `$TMPDIR`,
  ninguna abre una sesión con modelo ni ejecuta `TestSondeoDelLaboratorio`, `make evals-laboratorio`, `make evals`,
  `make evals-sondeo` ni los guiones `scripts/evals*.sh` desde la terminal (FR-017).
- No se tocan `skills/`, `evals/`, `data/`, `.github/workflows/evals.yml`, ningún umbral, el binario `kitlegal` ni sus
  órdenes, `cmd/laboratorio` ni ningún fichero de `testdata/` que ya exista (plan.md «Structure Decision»; FR-033). Las
  suites congeladas de H39 a H41.2 (`specs/031-h39-c-omo-se/aceptacion/`, `specs/032-h40-por-qu-e/aceptacion/`,
  `specs/034-h41-las-skills-del/aceptacion/`, `specs/035-h41-1-plazos-y-territorio/aceptacion/`,
  `specs/036-h41-2-por-qu-e/aceptacion/` y sus copias activadas en `internal/app/testdata/script/`) siguen en verde sin
  cambiar un fichero suyo (FR-032).
- Los textos son los de los contratos: los cuatro de la variante y los defectos de su comprobación, los de
  [contracts/variante.md](contracts/variante.md) §2 y §3; los argumentos, sus mensajes, la composición, el entorno y
  `sondeo.json`, los de [contracts/arnes.md](contracts/arnes.md) §1 a §5; las columnas, su orden y la tabla de ejemplo,
  los de [contracts/tabla.md](contracts/tabla.md) §2 a §4. Nada se inventa fuera de ellos.
- `internal/evals` no nombra `os.Stdin`, `os.Stdout`, `os.Stderr` ni llama a `os.Exit` (R4, R5); ningún `panic`; los
  errores de los argumentos del arnés, juntos con `errors.Join` y envueltos en `errorDeUso` como hoy; lo que llega
  manipulado (una variante, un transcript o un `sondeo.json` editados a mano) es la regla genérica de hoy, defecto
  `inesperado`, sin comprobación, mensaje ni test propios más allá de los de FR-002 (plan.md, principio V). Ningún
  `//nolint`: si `misspell` (`locale: US`) lee como inglés mal escrito un identificador o comentario nuevo en español, la
  tarea que lo introduce añade su entrada en `linters.settings.misspell.ignore-rules` de `.golangci.yml`, en su orden
  alfabético y con su comentario, como las demás (por eso las tareas de código declaran `.golangci.yml`). Tests de tabla
  con `t.Parallel()`.
- Cada control se ve fallar con un mutante (la medida pasada del umbral) antes de darlo por bueno, y el mutante se
  revierte (ADR 0029). Los prototipos y guiones de un solo uso van bajo `$TMPDIR` y no se versionan. Los fixtures
  nuevos no llevan nada con forma de secreto (gitleaks corre en `make ci`).

---

## Phase 1: Setup — la suite de aceptación

**Propósito**: fijar, antes de cualquier código, la tabla que el guion del laboratorio tiene que dar con los sondeos
fijos de las cuatro filas de H41.3 y uno de H41.2 (ADR 0018; plan.md «Aceptación e2e»; contracts/aceptacion.md).

- [X] T001 [aceptacion] [US3] Escribir, a partir del spec y de [contracts/aceptacion.md](contracts/aceptacion.md), [contracts/tabla.md](contracts/tabla.md) §4 y §6 y [contracts/arnes.md](contracts/arnes.md) §5, sin código de producto y sin evals nuevas (ninguna skill cambia: FR-033; plan.md «Structure Decision»), el guion testscript specs/037-h41-3-por-que-normativa/aceptacion/laboratorio-tabla.txtar (historia 3; el workflow lo activa como `h41-3-laboratorio-tabla.txtar` en la carpeta de guiones de internal/app al terminar el bucle y lo ejecuta `TestEntregaDelHito`; el arnés e2e no cambia: `KITLEGAL_LABORATORIO_BIN` es el `laboratorio` del árbol). **Cabecera**: un comentario que dice qué fija el guion y lleva la línea `# Cubre: US3 · FR-020, FR-021, FR-022, FR-024 · SC-006`, sin los textos `usage: `, `unknown command`, `unexpected command` ni `cannot parse`. **Órdenes**, en este orden: `exec sh -c '"$KITLEGAL_LABORATORIO_BIN" tabla instalacion > instalacion.out || true'`, `cmp instalacion.out instalacion-tabla.md`, `exec $KITLEGAL_LABORATORIO_BIN tabla instalacion`, `cmp stdout instalacion-tabla.md` y `! stderr .`. **Ficheros del archivo**: bajo `instalacion/`, los cinco sondeos de contracts/tabla.md §6 —`h41-3-sola`, `h41-3-control`, `h41-3-diagnostico` y `h41-3-remedio`, cada uno con el `sondeo.json` de su fila de contracts/arnes.md §5 (`claude_code` `2.1.284`, con `R` y `P` desarrollados, una línea con su salto final), y `h41-2`, con el de control y `"claude_code":"2.1.322"`— y sus once sesiones `normativa-producto/herramienta/04-lgt-prescripcion-<nn>/` con lo que hace cada una según la tabla de contracts/tabla.md §6 (`h41-3-sola` `-01` y `-02`; `h41-3-control` `-01` a `-03`; `h41-3-diagnostico` `-01` y `-02`; `h41-3-remedio` `-01` a `-03`; `h41-2` `-01`), cada una con `codigo-de-la-sesion`, `sesion.err` vacío y `sesion.jsonl` con la forma de las sesiones del archivo congelado specs/036-h41-2-por-qu-e/aceptacion/laboratorio-tabla.txtar: `system/init` con el modelo y la versión de su sondeo, un `assistant` con su `message.id` propio por respuesta y sus bloques `tool_use` («activa `normativa` antes»: `Skill` con `{"skill":"normativa"}` y, en el turno siguiente, `mcp__kitlegal__boe_articulo` con una consulta de la eval 04; «solo `consulta-juridica`, antes»: lo mismo con `{"skill":"consulta-juridica"}`; «ninguna»: `mcp__kitlegal__boe_articulo` sin ningún `Skill`), los `tool_result` en mensajes `user` y `result` con la respuesta; las dos sin terminar (`h41-3-remedio/…-03` y `h41-2/…-01`), código `124`, `Skill` de `consulta-juridica` y `mcp__kitlegal__boe_articulo`, sin `result`; ninguna abre una referencia; y `instalacion-tabla.md`, la tabla de contracts/tabla.md §4 byte a byte (título, tabla, «## Detalle» y su tabla, con su salto final). El guion falla hoy en el primer `cmp` por una aserción (el guion de hoy ignora `instalacion` y `presupuesto`, junta en una fila control, diagnóstico y remedio y da la tabla sin las dos columnas: contracts/aceptacion.md), lo que el workflow comprueba antes de congelarlo; ninguna tarea posterior lo toca. **Redelimitada en el intento 1** (gates/tarea-T001.md): la verificación de esta tarea corre `make ci` en un proceso que llega con SIGHUP ignorada y encuentra en rojo un test que ningún fichero de la tarea toca, así que la tarea declara además internal/evals/sesiones_test.go, solo para que ese test fije su propia condición de partida: en `TestTopeDeLaSesion`, cada subtest, antes de `abrirSesion`, escucha `syscall.SIGHUP` con `signal.Notify` en un canal propio con hueco para una señal y deja de escucharla con `signal.Stop` en `t.Cleanup`, con un comentario que dice por qué (la sesión hereda las señales que el proceso de test ignora y la shell del sustituto no puede restaurar una que le llega ignorada: con SIGHUP ignorada, el `kill -s HUP` que el sustituto se envía a sí mismo en `muere-por-otra-senal` no lo mata y el código es `0` y no `129`; mientras el test la escucha, Go no la ignora y la sesión la recibe con su disposición por omisión). Ningún caso, valor esperado ni aserción del test cambia, ni ningún otro fichero de Go, de test o de producto.

**Checkpoint**: la suite está escrita y en rojo por una comparación; desde aquí queda congelada.

---

## Phase 2: Foundational — una sola definición de etiqueta XML

**Propósito**: la definición de ADR 0046, punto 3, una vez, para la validación del frontmatter de hoy y para la
comprobación de la variante (US1) (research D4).

- [X] T002 Exportar la primera etiqueta XML de un texto desde `internal/skills` y usarla en su propia validación (FR-002; research D4; data-model §1; plan.md «Complexity Tracking», primera fila). **Código**: internal/skills/frontmatter.go gana `PrimeraEtiquetaXML(texto string) string`, la primera coincidencia de la expresión `etiquetaXML` de hoy (`:67-71`), o `""` sin ninguna, con su comentario de documentación; `defectosDeLaEtiqueta` (`:481-488`) la usa en lugar de llamar a la expresión; ningún mensaje ni regla de la validación cambia, ninguna importación nueva. **Test**: en internal/skills/frontmatter_test.go, `TestPrimeraEtiquetaXML`, de tabla con `t.Parallel()`: `<ejemplo>` en medio de un texto da `<ejemplo>`; dos etiquetas dan la primera; una de cierre (`</b>`) da la suya; un texto sin `<`, `a < b` y `<1>` dan `""`. Mutantes que se ven en rojo y se revierten: la última coincidencia en lugar de la primera; `""` siempre. Sin cambiar ningún valor esperado siguen en verde los tests de frontmatter de internal/skills y `make skills-check`. `make ci` en verde.

**Checkpoint**: la etiqueta XML tiene una sola definición, que la validación de las skills ya usa.

---

## Phase 3: User Story 1 — Una variante del producto con las cuatro descripciones cortas (P1)

**Goal**: `laboratorio/producto-descripciones-cortas/` guarda los cuatro textos, y `make ci` sale en rojo si dejan de
ser uno por skill de `skills/` o pasan de sus límites.

**Independent Test**: `TestDescripcionesCortas`, sin red ni modelo: la variante del repositorio sin ningún defecto, y
cada caso roto de contracts/variante.md §3 con su línea exacta.

- [X] T003 [US1] La variante y su comprobación, con su test (FR-001, FR-002, FR-033; SC-001; contracts/variante.md §1 a §4; data-model §1; research D1, D2, D3). **Variante**: laboratorio/producto-descripciones-cortas/consulta-juridica.txt, laboratorio/producto-descripciones-cortas/jurisprudencia.txt, laboratorio/producto-descripciones-cortas/normativa.txt y laboratorio/producto-descripciones-cortas/plazos-y-territorio.txt, nuevos, cada uno con exactamente la línea de su apartado de contracts/variante.md §2 y su salto de línea final (417, 420, 426 y 427 runas; 1690 en total), y nada más en el directorio. **Código**: internal/evals/laboratorio_descripciones.go, nuevo, con la carpeta de la variante (`laboratorio/producto-descripciones-cortas`, la ruta con que empieza cada defecto), los límites (450 runas por descripción y 1700 en total) y `descripcionesCortas(raiz string) (map[string]string, []string, error)`: las skills son los subdirectorios de `<raiz>/skills/` (la regla de `copiarLasDemasSkillsDelArbol`); el texto de cada una, su fichero `<skill>.txt` sin el salto de línea final (la regla de `textoDeLaVariante`); los defectos, uno por línea, con el texto y en el orden de contracts/variante.md §3 —por cada skill en orden de nombre, el de que falte o, si está, el de su longitud y el de su etiqueta (con `skills.PrimeraEtiquetaXML` de T002); después, por cada entrada de la variante que no es `<skill>.txt` de una skill, en orden de nombre; al final, el de la suma de las que hay—, contando runas con `unicode/utf8`; y como error solo un directorio que no se puede leer, también uno que no existe (defecto `inesperado`, ADR 0023). Nada la llama todavía fuera de su test. **Test**: internal/evals/laboratorio_descripciones_test.go, nuevo, `TestDescripcionesCortas` (contracts/variante.md §4), de tabla con `t.Parallel()`: con la raíz del repositorio, ningún defecto y un texto por cada skill de `skills/`, igual a su línea de contracts/variante.md §2; y, sobre árboles que el test monta en `t.TempDir()` (un `skills/` con subdirectorios vacíos y la variante con sus ficheros), cada fila de la tabla de §3 con su línea exacta —falta una, sobra `otra.txt`, una de 451 runas, cuatro de 450, 450, 400 y 401 (solo `… las descripciones suman 1701 caracteres (máximo 1700)`), una que empieza por `<ejemplo>` y una quinta skill sin descripción— y el ejemplo de cinco líneas de §3 en su orden; falla con la primera línea que difiere. Mutantes que se ven en rojo y se revierten: el límite de 450 con `>=`; contar bytes en lugar de runas (la variante del repositorio, de 1732 bytes, da el defecto de la suma); la quinta skill sin defecto; el defecto de la entrada que sobra antes que los de las skills. Sin cambiar ningún valor esperado sigue en verde `TestLaboratorioFueraDelProducto` (la variante no se empotra ni se instala, y ninguna de sus rutas tiene por tramo el nombre de una variante del producto). Ningún fichero de `skills/` cambia. `make ci` en verde; .golangci.yml solo si lo pide `misspell` (convenciones).

**Checkpoint**: la variante existe y `make ci` la protege; nada del arnés la usa todavía.

---

## Phase 4: User Story 2 — El arnés instala la variante compuesta y fija el presupuesto del listado (P1)

**Goal**: `make evals-laboratorio` gana `INSTALACION=producto-descripciones-cortas` y `PRESUPUESTO=<caracteres>`;
`sondeo.json` anota cada opción; sin ellas, el arnés de hoy byte a byte, también el entorno de cada sesión.

**Independent Test**: `TestConLaDescripcion`, `TestArnesConLasDescripcionesCortas` y `TestArnesConElPresupuesto` con el
sustituto de `claude`; `TestArnesDelLaboratorio`, `TestArnesDeNormativa`, `TestArnesDeLasFamilias`,
`TestArnesConLaInstalacionDelProducto` y `TestArnesConUnaEval` sin cambiar ningún valor esperado; `TestRecetasDelSondeo`
y `TestGuionDelLaboratorio` con diez argumentos.

- [X] T004 [US2] La composición de un `SKILL.md` con otra `description`, sobre una clasificación de líneas compartida con `descripcionYCuerpo` (FR-011; SC-002; contracts/arnes.md §3; data-model §3; research D5; plan.md «Complexity Tracking», filas 2 y 6). **Código**: internal/evals/laboratorio.go: la clasificación de las líneas de `descripcionYCuerpo` (`:303-346`: el frontmatter entre el primer delimitador y el siguiente; la línea que empieza por `description:` y las sangradas con espacio o tabulador que la siguen) pasa, sin cambiar su regla, a una función iteradora del mismo fichero (`iter.Seq2` de cada línea con si es de la `description`), que `descripcionYCuerpo` usa para dar exactamente lo mismo que hoy; internal/evals/laboratorio_descripciones.go gana `conLaDescripcion(contenido []byte, texto string) ([]byte, error)`, que con esa clasificación sustituye las líneas de la `description`, en su sitio, por las que da `yaml.Marshal(map[string]string{"description": texto})` de `go.yaml.in/yaml/v3` (ya en go.mod; ninguna dependencia nueva) y deja todas las demás tal cual y en su orden; un `SKILL.md` sin frontmatter o sin la clave es un error (defecto `inesperado`). Nada la llama todavía fuera de su test. **Test**: en internal/evals/laboratorio_descripciones_test.go, `TestConLaDescripcion`, de tabla con `t.Parallel()`: con cada `SKILL.md` de `skills/` del repositorio y, por cada uno, el texto de su skill en la variante de T003 y textos con `: `, comillas simples y dobles, `#` y `«»`: la `description` que lee `skills.LeerFrontmatter` del compuesto es el texto exacto; `skills.ValidarFrontmatter`, con la raíz del repositorio y los applets que declara el propio frontmatter como registrados, no da ningún defecto (research D12); y el cuerpo de `descripcionYCuerpo` del compuesto es, byte a byte, el del original. Mutantes que se ven en rojo y se revierten: el texto escrito sin codificar detrás de `description: ` (rojo con `: ` y `#`); solo la línea de la clave sustituida, con las sangradas del bloque de hoy detrás (rojo en el cuerpo y en la lectura); la clasificación de la composición sin las líneas con tabulador. Sin cambiar ningún valor esperado siguen en verde los tests de H39 que usan `descripcionYCuerpo` y `TestLaboratorioFueraDelProducto`. `make ci` en verde (`internal/arch_test.go` no cambia: `internal/evals` no está en el cierre del binario); .golangci.yml solo si lo pide `misspell`.
- [X] T005 [US2] El sustituto de `claude` anota su entorno, y el caso sin `PRESUPUESTO` de `TestArnesConElPresupuesto`, escritos y en verde con el arnés de hoy, sin tocar código del arnés: es un test que cubre código que ya existe, el entorno de las sesiones de hoy, y se escribe antes de la opción para que lo que fija «hoy» no salga del código nuevo (FR-014, FR-015, FR-018; SC-003, SC-004; contracts/arnes.md §4 y §6; research D8, D13; plan.md «Diseño»). **Test**: en internal/evals/laboratorio_sondeo_test.go, `guionDelClaudeDelLaboratorio` (`:49-62`) deja además, en cada sesión, una anotación `entorno.XXXXXX` con la ruta de la sesión en la primera línea y la salida de `env` detrás; las anotaciones de hoy (`sesion.`, `version.`) y su lectura por prefijo no cambian; un ayudante lee, de las anotaciones `entorno.`, el entorno de cada sesión por su ruta. `TestArnesConElPresupuesto`, nuevo, con `t.Parallel()` y `SLASH_COMMAND_TOOL_CHAR_BUDGET=500` en el entorno de quien lanza (el `Entorno` del sondeo que monta el test), con la construcción sustituida y el sustituto de `claude` de `TestArnesDelLaboratorio`: la petición del control (`VARIANTES=normativa-producto MODOS=herramienta SESIONES=2 EVAL=04 INSTALACION=producto`): ninguna sesión tiene la variable; el entorno que vuelca cada sesión es, como conjunto, nombre a nombre y valor a valor, el de `sinElDirectorioEnElPATH(entornoDeLaSesion(entornoDelSondeo(<entorno del test>, <temporal>), <sesión>, false), <temporal>/bin)`, con `<sesión>` `DIRECTORIO/normativa-producto/herramienta/<nombre>` con ruta absoluta, sin ninguna de más ni de menos salvo `PWD`, `SHLVL` y `_`; y `sondeo.json` es el de control de contracts/arnes.md §5 (con el `claude_code` del sustituto). Mutantes que se ven en rojo y se revierten, en internal/evals/sondeo.go o internal/evals/laboratorio_sondeo.go: `SLASH_COMMAND_TOOL_CHAR_BUDGET` en `variablesQueVenLasSesionesDelSondeo`; una variable más en `entornoDelSondeo`; `HOME` quitado del entorno de las sesiones. Sin cambiar ningún valor esperado siguen en verde `TestArnesDelLaboratorio`, `TestArnesDeNormativa`, `TestArnesDeLasFamilias`, `TestArnesConLaInstalacionDelProducto` y `TestArnesConUnaEval`. `make ci` en verde.
- [X] T006 [US2] `INSTALACION=producto-descripciones-cortas`: sus reglas, la composición en `preparar` y la constancia en `sondeo.json` (FR-010, FR-011, FR-012, FR-015, FR-016, FR-018; SC-002, SC-004, SC-005; contracts/arnes.md §2, §3, §5 y §6; data-model §1 a §3 y §5; research D3, D5, D6, D9). **Código**: internal/evals/laboratorio_familias.go: las constantes `varianteDelProductoDeNormativa = "normativa-producto"`, que la declaración de la familia de `normativa` usa en sus dos sitios en lugar del literal (`:82` y `:87`), e `instalacionConLasDescripcionesCortas = "producto-descripciones-cortas"`, a su lado; internal/evals/laboratorio_sondeo.go: `comprobar` recibe además la raíz (y `comprobarElLaboratorio` se la da); `errorDeLaInstalacion` gana el valor nuevo y las variantes que valen, con las reglas 1 a 5 de research D6 y los mensajes de contracts/arnes.md §2 (de las cuatro primeras, la primera que se da y una línea como mucho; «no es producto» y el de H39, los de hoy byte a byte; `INSTALACION: producto-descripciones-cortas solo vale con VARIANTES=normativa-producto` con cualquier variante que vale y no es `normativa-producto`, también en una mezcla y sin mirar la familia, salvo con la de H39; y, solo con el valor nuevo, un `INSTALACION: <defecto>` por cada defecto de `descripcionesCortas` de T003, que solo se lee con ese valor); un error al leer la variante no es un error de uso (sin `errorDeUso`, como el de una declaración que no se puede leer); con cualquiera, ninguna sesión, nada construido y `DIRECTORIO` sin crear; `laboratorioComprobado` gana `instalacion` (vacía, `producto` o la nueva; `conElProducto` es que no esté vacía) y `descripciones` (los textos con el valor nuevo; nil sin él); `preparar`, con el valor nuevo y tras la copia de `INSTALACION=producto`, llama a `componerLasDescripciones`, nueva en internal/evals/laboratorio_descripciones.go, que reescribe en el temporal cada `SKILL.md` de `skills/normativa-producto/` con `conLaDescripcion` de T004 y el texto de su skill; la copia compuesta no se guarda en `DIRECTORIO`; internal/evals/laboratorio_tabla.go, solo el tipo `sondeoLeido`, gana `Instalacion string` con `json:"instalacion,omitempty"` detrás de `SkillsInstaladas` (la tabla aún no lo usa); `crearElSondeoDelLaboratorio` la escribe solo con el valor nuevo; sin él, `sondeo.json` el de hoy byte a byte. **Test**: en internal/evals/laboratorio_sondeo_test.go, `TestArnesConLasDescripcionesCortas` (contracts/arnes.md §6), con `t.Parallel()`, la construcción sustituida y el sustituto de `claude`: errores de uso con su mensaje, ninguna sesión, nada construido, el temporal vacío y `DIRECTORIO` sin crear, con `normativa-control`, con `jurisprudencia-producto`, con `control` (de H39, el mensaje de hoy), con `normativa-producto,normativa-control`, con `normativa-producto,jurisprudencia-producto` (las líneas de hoy de la mezcla y, detrás, la de «solo vale con»), y, sobre una raíz de prueba en `t.TempDir()` con `laboratorio/variantes.yaml` y `skills/` del repositorio y la variante sin `plazos-y-territorio.txt`, `INSTALACION: laboratorio/producto-descripciones-cortas: falta la descripción de plazos-y-territorio`; y la petición del remedio, `VARIANTES=normativa-producto MODOS=herramienta SESIONES=2 EVAL=04 INSTALACION=producto-descripciones-cortas`: dos sesiones; cada una enlaza exactamente las cuatro skills de `skills/`, en orden de nombre y sin repetir, a la copia de su variante en el temporal; la `description` de cada `SKILL.md` enlazado, leída con `skills.LeerFrontmatter`, es el texto de su fichero de la variante; el cuerpo de `descripcionYCuerpo` de ese `SKILL.md` es el de `skills/`; cada otro fichero, ruta a ruta y byte a byte, el de `skills/`; `skills.ValidarFrontmatter` no da ningún defecto; su servidor, el de la familia; `sondeo.json`, el de remedio de contracts/arnes.md §5; y ningún `SKILL.md` en `DIRECTORIO`; falla con la primera sesión cuyas skills, descripción o bytes difieren. Mutantes que se ven en rojo y se revierten: la `description` de una skill sin sustituir; la composición hecha también con `INSTALACION=producto` (rojo en `TestArnesConLaInstalacionDelProducto`); «solo vale con» dado solo con familia (rojo con la mezcla con `jurisprudencia-producto`); la comprobación después de crear `DIRECTORIO`; `instalacion` escrita también sin la opción (rojo en `TestArnesConLaInstalacionDelProducto`). Sin cambiar ningún valor esperado siguen en verde `TestArnesDelLaboratorio`, `TestArnesDeNormativa`, `TestArnesDeLasFamilias`, `TestArnesConLaInstalacionDelProducto`, `TestArnesConUnaEval` y el caso de T005. `make ci` en verde; .golangci.yml solo si lo pide `misspell`.
- [X] T007 [US2] `PRESUPUESTO` en el arnés: su forma y sus errores, la variable en el entorno de cada sesión y la constancia en `sondeo.json` (FR-012, FR-013, FR-014, FR-016, FR-018; SC-003, SC-005; contracts/arnes.md §2, §4 a §6; data-model §2, §4 y §5; research D7, D8, D9). **Código**: internal/evals/laboratorio_sondeo.go: `ArgumentosDelLaboratorio` gana `Presupuesto`, detrás de `Eval`, tal como llega; `comprobar` da, el último, detrás del de `EVAL`, una sola línea: `PRESUPUESTO: «<v>» no es un entero mayor que 0` si no es vacío y no tiene la forma `formaDeUnEnteroPositivo` (internal/evals/sondeo.go:34) o no cabe en un `int` (`strconv.Atoi`), y si la tiene, `PRESUPUESTO: las variantes del laboratorio de H39 no son de ninguna familia del producto` solo si las variantes que valen son de la familia de H39; `laboratorioComprobado` gana `presupuesto int` (0 sin él); con `presupuesto > 0`, el entorno base de las sesiones es el de hoy, `entornoDelSondeo(s.Entorno, temporal)` (`:282`), con `SLASH_COMMAND_TOOL_CHAR_BUDGET=<presupuesto>` detrás, y `entornoDeLaSesion` pone encima lo de cada sesión sin cambios; sin él, el código de hoy; internal/evals/laboratorio_tabla.go, solo el tipo `sondeoLeido`, gana `Presupuesto int` con `json:"presupuesto,omitempty"` detrás de `Instalacion`; `crearElSondeoDelLaboratorio` lo escribe solo con la opción. **Test**: en internal/evals/laboratorio_sondeo_test.go, `TestArnesConElPresupuesto` gana, con la variable a 500 en el entorno de quien lanza como en T005: errores de uso con ninguna sesión, nada construido, el temporal vacío y `DIRECTORIO` sin crear, con `PRESUPUESTO=0`, `0100`, `+1000000` y `99999999999999999999`, cada uno con su línea; `PRESUPUESTO=1000000` con `control`, de H39; y los dos ejemplos de contracts/arnes.md §2, con sus dos líneas en su orden; con `PRESUPUESTO=1000000`, la petición del diagnóstico (la del control de T005 con `INSTALACION=producto`), la del remedio de T006 con presupuesto y `jurisprudencia-control` sin `INSTALACION`: cada sesión tiene una sola línea `SLASH_COMMAND_TOOL_CHAR_BUDGET=1000000`, y `sondeo.json` lleva `"presupuesto":1000000` detrás de lo de contracts/arnes.md §5 (el de diagnóstico; el de remedio con las dos claves); y los nombres de las variables de cada sesión del control de T005 son los de la misma sesión del diagnóstico menos `SLASH_COMMAND_TOOL_CHAR_BUDGET`; el caso sin `PRESUPUESTO` de T005, sin cambiar. Mutantes que se ven en rojo y se revierten: `0100` aceptado (solo `strconv.Atoi`); la variable puesta también sin `PRESUPUESTO`; el valor tomado del entorno de quien lanza; la línea de `PRESUPUESTO` delante de la de `EVAL`; las dos líneas de `PRESUPUESTO` a la vez; `presupuesto` escrito como cadena. Sin cambiar ningún valor esperado siguen en verde los tests del arnés de T005 y T006 y los de hoy. `make ci` en verde; .golangci.yml solo si lo pide `misspell`.
- [ ] T008 [US2] `PRESUPUESTO` por la cadena del sondeo, de la receta a `go test` (FR-013, FR-017; SC-003; contracts/arnes.md §1 y §6; research D7, D14). **Código**: internal/evals/job_test.go: `TestSondeoDelLaboratorio` (`-tags evals`, que `make lint` compila y `make ci` no ejecuta) gana la bandera `-presupuesto` y la pasa a `ArgumentosDelLaboratorio.Presupuesto`; scripts/evals-laboratorio.sh exige diez argumentos y da el décimo a `go test` como `-presupuesto`, delante de `-temporal`; su línea de uso (`:29`) gana `<presupuesto>` detrás de `<eval>`, y su cabecera de comentario dice diez argumentos y que `PRESUPUESTO` vacío es sin presupuesto; Makefile: la receta de `evals-laboratorio` pasa `"$(PRESUPUESTO)"` detrás de `"$(EVAL)"`, y el comentario que la precede (`:157-167`) dice diez argumentos, el valor nuevo de `INSTALACION` y `PRESUPUESTO`, con su línea `##` de ayuda si nombra los argumentos; scripts/evals-sondeo-llavero.sh, solo su comentario («con sus nueve» → diez), sin cambiar código ni órdenes. **Tests**: en internal/evals/laboratorio_sondeo_test.go, `TestRecetasDelSondeo`, su caso de `evals-laboratorio` pasa también `PRESUPUESTO=1000000` y espera la orden del llavero con `"1000000"` detrás de `"04"`; `TestGuionDelLaboratorio` pasa de nueve a diez argumentos: cada caso de `casosDelGuionDelLaboratorio` pasa un décimo, `1000000` en `go-test-sale-con-0-y-directorio-absoluto` (con el noveno vacío, para que el noveno y el décimo se vean por separado) y vacío en los demás; `ordenDelLaboratorio` espera además `-presupuesto` con el décimo detrás de `-eval <noveno>`, con `-temporal <temporal>` como último argumento; el caso `sin-nueve-argumentos` pasa a `sin-diez-argumentos`, con nueve, que sale con 1, da en la salida de error la línea de uso con `<presupuesto>` y no ejecuta go (el mensaje de su aserción `sinGo` dice diez); y los comentarios de `TestGuionDelLaboratorio`, `ordenDelLaboratorio` y `casosDelGuionDelLaboratorio` dicen diez. Mutantes que se ven en rojo y se revierten: el guion sin pasar el décimo a `go test`, o con otra bandera, o detrás de `-temporal` (rojo en `TestGuionDelLaboratorio`); el guion que acepta nueve (rojo en `sin-diez-argumentos`); la receta sin `"$(PRESUPUESTO)"` (rojo en `TestRecetasDelSondeo`). `make lint` compila `job_test.go` con `-tags evals`. `make ci` en verde.

**Checkpoint**: las cuatro invocaciones de contracts/arnes.md §1 son órdenes válidas del arnés y preparan lo que piden,
comprobado sin modelo; sin las dos opciones, lo de hoy byte a byte.

---

## Phase 5: User Story 3 — La tabla dice la instalación y el presupuesto de cada fila (P1)

**Goal**: con algún sondeo con la constancia, la tabla de una familia del producto da «Instalación» y «Presupuesto»
detrás de «Skills instaladas» y una fila por cada instalación y presupuesto, en el orden de FR-021; sin ella, la de hoy
byte a byte.

**Independent Test**: `TestTablaConLaInstalacion` y el guion congelado de T001; `TestTablaConLasActivaciones`,
`TestTablaDeLasFamilias`, `TestTablaDeNormativa` y `TestTablaDelLaboratorio` sin cambiar.

- [ ] T009 [US3] La instalación y el presupuesto en la clave, la cabecera y las filas de la tabla, solo con la constancia (FR-020, FR-021, FR-022, FR-023; SC-006, SC-007; contracts/tabla.md §2 a §4; data-model §6; research D10, D11). **Código**: internal/evals/laboratorio_tabla.go: `instalacionDeLaFila(leido sondeoLeido, skill string) string` —su `instalacion`; sin ella, `producto` si `skills_instaladas` tiene alguna skill además de `skill`, y si no (solo ella, o sin la clave), `skill`—; `familiaDeLaTabla` gana `conLaInstalacion`, verdadero si algún sondeo de la tabla tiene `instalacion` o `presupuesto` (junto a `algunoConLaConstancia`); `claveDeFila` gana `instalacion string` y `presupuesto int`, que solo se llenan con `conLaInstalacion` (sin él, vacíos en todas las claves); `compararClaves` (`:910`) ordena por variante, modo, modelo, versión, instalación, presupuesto (0, el vacío, primero; después de menor a mayor) y skills instaladas; con `conLaInstalacion`, la cabecera, el separador y cada fila dan «Instalación» y «Presupuesto» detrás de «Skills instaladas», con el presupuesto vacío si es 0, también en una fila con `n` = 0; el detalle, los recuentos y los intervalos, como en H41.2; sin `conLaInstalacion`, la tabla y el detalle de hoy byte a byte; la de H39 no lo usa. **Test**: en internal/evals/laboratorio_tabla_test.go, `TestTablaConLaInstalacion`, con `t.Parallel()`: el subtest de `instalacionDeLaFila` en sus cuatro casos (`instalacion` en el `sondeo.json`; las cuatro skills; solo la de la familia; sin `skills_instaladas`), y uno sobre sondeos que el test monta en `t.TempDir()` como copias de los sondeos fijos `h41-2-a` y `h41-2-b` de la familia de las activaciones que lee `TestTablaConLasActivaciones` (sin tocar los originales), uno de ellos con `"presupuesto":1000000` y otro con `"instalacion":"producto-descripciones-cortas"` añadidos a su `sondeo.json`: la cabecera con las dos columnas en su sitio, una fila por instalación y presupuesto en el orden de contracts/tabla.md §3, y en cada fila su instalación, su presupuesto y los recuentos que da hoy su sondeo sin la constancia. Comprobar además, sin versionar nada, que `go run ./cmd/laboratorio tabla` sobre el directorio `instalacion/` del archivo de specs/037-h41-3-por-que-normativa/aceptacion/laboratorio-tabla.txtar (T001), extraído en `$TMPDIR` con un guion de un solo uso, da byte a byte su `instalacion-tabla.md`, que es lo que pide el guion congelado. Mutantes que se ven en rojo y se revierten: la instalación ordenada detrás de las skills instaladas (la fila `normativa` sale la última); el presupuesto vacío ordenado después de un valor; `producto` solo con cuatro nombres fijos; las columnas dadas también sin la constancia (rojo en `TestTablaConLasActivaciones`, `TestTablaDeLasFamilias` y `TestTablaDeNormativa`). Sin cambiar ningún valor esperado ni fichero, siguen en verde `TestTablaConLasActivaciones`, `TestTablaDeLasFamilias`, `TestTablaDeNormativa`, `TestTablaDelLaboratorio` y las suites congeladas de H39 a H41.2 (FR-022, FR-023, FR-032; SC-007). `make ci` en verde; .golangci.yml solo si lo pide `misspell`.
- [ ] T010 [datos] [US3] Los sondeos fijos de las cuatro filas y uno de H41.2 con su tabla esperada, y el subtest que compara con ella la salida del guion: un test que cubre código que ya existe desde T009, sin tocar código (FR-020, FR-021, FR-022, FR-024; SC-006; contracts/tabla.md §4 y §6; research D12). **Fixtures**: internal/evals/testdata/laboratorio/familias/instalacion/, nuevo: `sondeo/h41-2/`, `sondeo/h41-3-control/`, `sondeo/h41-3-diagnostico/`, `sondeo/h41-3-remedio/` y `sondeo/h41-3-sola/`, copia byte a byte de los sondeos bajo `instalacion/` del archivo de specs/037-h41-3-por-que-normativa/aceptacion/laboratorio-tabla.txtar (T001), y `tabla.md`, copia byte a byte de su `instalacion-tabla.md` (la extracción, con un guion de un solo uso bajo `$TMPDIR`); ningún fichero de `testdata/` existente cambia; nada se graba. **Test**: en internal/evals/laboratorio_tabla_test.go, `TestTablaConLaInstalacion` gana el subtest `sondeos-fijos`: la salida de la tabla sobre `familias/instalacion/sondeo/` es su `tabla.md` byte a byte, con la salida de error vacía y código 0, como `probarLosSondeosFijosConLasActivaciones`; falla con la primera fila, instalación, presupuesto, recuento o intervalo que difiere. Comprobar, sin versionar nada, que `go run ./cmd/laboratorio tabla internal/evals/testdata/laboratorio/familias/instalacion/sondeo` da ese `tabla.md` (quickstart.md §3). Mutantes que se ven en rojo y se revierten, en internal/evals/laboratorio_tabla.go: «Presupuesto» delante de «Instalación»; el presupuesto de `h41-3-diagnostico` dado vacío; la instalación de `h41-2` dada como `normativa`. `make ci` en verde.

**Checkpoint**: la tabla de las cuatro filas de FR-030 se puede contar sin modelo; las tablas de H39 a H41.2 no
cambian; el guion de T001 pasa contra el código de T009.

---

## Phase 6: User Story 4 — La medida y su ADR, fuera del run (P2)

Sin tareas: FR-030, FR-031 y SC-009 se hacen fuera del run (ADR 0032; FR-017) con lo que entregan T003 a T010 y las
órdenes de contracts/arnes.md §1 y quickstart.md §7. El informe final lo deja pendiente por la línea
`[alcance] plan: FR-030 y FR-031 …` que ya está en `gates/supuestos.md` (research D15).

---

## Phase 7: Polish & cross-cutting

- [ ] T011 La documentación de lo que deja de ser cierto (FR-001, FR-010, FR-012, FR-013, FR-020; DoD §1; research D14): CHANGELOG.md, en *Unreleased*, la variante `producto-descripciones-cortas` en `laboratorio/` y su comprobación en `make ci`, `INSTALACION=producto-descripciones-cortas` y `PRESUPUESTO=<caracteres>` de `make evals-laboratorio`, `instalacion` y `presupuesto` en `sondeo.json` y las columnas «Instalación» y «Presupuesto» de la tabla de una familia del producto; CONTRIBUTING.md, en «El sondeo del laboratorio», la variante y su forma (un fichero `<skill>.txt` por skill de `skills/`, sus límites y que solo la lee el arnés), la sinopsis y la tabla de argumentos con el valor nuevo de `INSTALACION` (solo con `VARIANTES=normativa-producto`) y `PRESUPUESTO` (sus reglas y que pone `SLASH_COMMAND_TOOL_CHAR_BUDGET` solo en las sesiones del laboratorio), las claves `instalacion` y `presupuesto` y las dos columnas, con las órdenes de contracts/arnes.md §1, y la mención de `scripts/evals-laboratorio.sh` «con sus nueve» (`:1660`) → diez; sin cambiar código, órdenes ni recetas. Ningún ADR (el que lee el sondeo se escribe fuera del run: FR-031, research D15), ninguna fila nueva en docs/SOURCES.md (no hay fuente nueva) y ningún esquema. `make ci` en verde.

---

## Dependencies & Execution Order

### Orden de las fases

- **Fase 1 (T001)**: primero; congela la suite.
- **Fase 2 (T002)**: antes de US1; exporta la etiqueta XML que usa la comprobación de T003.
- **Fase 3 (T003)**: la variante y `descripcionesCortas`, que usa T006.
- **Fase 4 (T004 → T005 → T006 → T007 → T008)**: T004 da `conLaDescripcion`, que usa la composición de T006; T005 fija
  el entorno de hoy con el arnés de hoy, antes de cualquier opción; T006 añade `INSTALACION=producto-descripciones-cortas`
  sobre T003 y T004; T007 añade `PRESUPUESTO` y sus casos, que usan las peticiones de T005 y T006 (los ejemplos de
  contracts/arnes.md §2 llevan el valor nuevo de `INSTALACION`); T008 pasa `PRESUPUESTO` por la cadena (diez argumentos)
  con el campo de T007.
- **Fase 5 (T009 → T010)**: T009 necesita los campos `Instalacion` (T006) y `Presupuesto` (T007) de `sondeoLeido` y el
  archivo de T001; T010 copia el archivo de T001 y fija la salida del código de T009. Tras T009,
  `h41-3-laboratorio-tabla.txtar` pasa.
- **Fase 7 (T011)**: al final; documenta lo de T003 a T010.

### Dentro de cada tarea

Test y la implementación mínima que lo pone en verde, en la misma tarea (T005 y T010, tests que cubren código que ya
existe, sin tocarlo); el mutante que ve fallar el control, revertido; `make ci` en verde antes de marcarla.

### Oportunidades de paralelismo

Ninguna tarea lleva [P]: T003, T004 y T006 comparten internal/evals/laboratorio_descripciones.go; T005 a T008,
internal/evals/laboratorio_sondeo_test.go; T006, T007 y T009, internal/evals/laboratorio_tabla.go; T009 y T010,
internal/evals/laboratorio_tabla_test.go. Se ejecutan en serie, en el orden de los ids.

## Parallel Example

```text
Ninguno: el bucle del workflow ejecuta y verifica las tareas una a una, en el orden de los ids.
```

## Implementation Strategy

### MVP

T001 → T002 → T003: la variante de las cuatro descripciones cortas existe y `make ci` la protege (US1), que es el
remedio candidato que mide la cuarta fila del sondeo.

### Entrega incremental

T004 a T008 dan el arnés de las cuatro filas de FR-030 (US2); T009 y T010, su tabla y el guion congelado en verde (US3);
T011, la documentación. Cada tarea deja `make ci` en verde.

## Notes

- 11 tareas: T001 [aceptacion]; base: T002; US1: T003; US2: T004, T005, T006, T007, T008; US3: T009, T010; US4:
  ninguna; T011.
- Tareas [datos]: T010 (internal/evals/testdata/laboratorio/familias/instalacion/), sin red, con su subtest que cubre el
  código de T009 y sin tocar código. El hito no tiene datos externos (plan.md «Datos externos»): ni manifiesto
  `grabaciones.json`, ni `TestGrabar*`, ni grabación.
- Definition of Done (docs/ROADMAP.md §1): `make ci` y tests offline en cada tarea; ningún sobre ni esquema cambia (sin
  tarea de `schemas/`); e2e: T001; CHANGELOG: T011; ADR: no aplica en el run (el de FR-031 se escribe fuera, research
  D15); `docs/SOURCES.md`: no aplica (ninguna fuente); fixtures: T010; skills con evals: las de hoy, sin cambiar, que
  mide el job del cierre (FR-034); dimensión territorial: ningún dato ni código nombra un municipio.
- Cada fila de «Controles de umbral» de plan.md y la tarea que construye o toca su control y lo ve fallar con un
  mutante:
  - SC-001, FR-001, FR-002 — `TestDescripcionesCortas`: T003 (450 con `>=`; bytes en lugar de runas; la quinta skill sin
    defecto; el orden de los defectos).
  - SC-002, FR-010, FR-011, FR-018 — `TestConLaDescripcion`: T004 (texto sin codificar; sangradas del bloque de hoy
    detrás); `TestArnesConLasDescripcionesCortas`: T006 (una `description` sin sustituir).
  - SC-003, FR-013, FR-014 — `TestArnesConElPresupuesto`: T005 (la variable dejada pasar desde quien lanza) y T007 (la
    variable sin `PRESUPUESTO`; el valor tomado de quien lanza); `TestRecetasDelSondeo` y `TestGuionDelLaboratorio`:
    T008 (el décimo argumento sin pasar o mal puesto).
  - SC-004, FR-012, FR-015 — el caso sin `PRESUPUESTO` de `TestArnesConElPresupuesto`: T005 (una variable más o una
    menos en el entorno de hoy); `TestArnesDelLaboratorio`, `TestArnesDeNormativa`, `TestArnesDeLasFamilias`,
    `TestArnesConLaInstalacionDelProducto` y `TestArnesConUnaEval` (de hoy, sin cambiar sus valores): T006 (la
    composición o `instalacion` también con `INSTALACION=producto`) y T007.
  - SC-005, FR-016 — `TestArnesConLasDescripcionesCortas`: T006 («solo vale con» solo con familia; la comprobación
    después de crear `DIRECTORIO`); `TestArnesConElPresupuesto`: T007 (`0100` aceptado; el orden de las líneas);
    `TestArnesConLaInstalacionDelProducto` (de hoy, sin cambiar sus mensajes): T006.
  - SC-006, FR-020, FR-021, FR-024 — `TestTablaConLaInstalacion`: T009 (el orden de la instalación y del presupuesto;
    `producto` con cuatro nombres fijos) y T010 (las columnas cambiadas; un presupuesto o una instalación de otra fila);
    `TestEntregaDelHito` con `h41-3-laboratorio-tabla.txtar`: T001, que lo ve fallar hoy en su primer `cmp`, y T009, que
    lo pone en verde.
  - SC-007, FR-022, FR-023, FR-032 — `TestTablaConLasActivaciones`, `TestTablaDeLasFamilias`, `TestTablaDeNormativa`,
    `TestTablaDelLaboratorio` y las suites de H39 a H41.2 en `TestEntregaDelHito` (de hoy, sin cambiar): T009 (las
    columnas nuevas sin la constancia, en rojo).
  - SC-008, FR-034 — las filas `evals:` de `normativa`, `jurisprudencia` y `plazos-y-territorio`: las mide el job de
    evals del cierre, que hace el workflow tras la revisión final con los umbrales de hoy, sin tarea que los toque.
- Requisitos sin tarea propia: FR-030, FR-031 y SC-009 (fuera del run; US4); FR-034 y SC-008 (el cierre del workflow);
  FR-017 (ninguna tarea lanza el arnés; lo guardan las convenciones y quickstart.md §5); FR-032 y FR-033 (todas:
  ninguna declara una ruta de las suites congeladas ni de `skills/`, y lo guarda el guardián de diff).
