---

description: "Tareas de H41.2 · Por qué la eval 04 no activa normativa en modo herramienta en el job"
---

# Tasks: H41.2 · Por qué la eval 04 no activa `normativa` en modo herramienta en el job

**Input**: `specs/036-h41-2-por-qu-e/` — [spec.md](spec.md), [plan.md](plan.md), [research.md](research.md),
[data-model.md](data-model.md), [quickstart.md](quickstart.md) y [contracts/](contracts/) (`informe.md`, `arnes.md`,
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
- **[aceptacion]**: la suite congelada (solo T001). **[datos]**: la tarea toca ficheros bajo `testdata/` (T003 y T008).

## Convenciones de todas las tareas

- `make ci` en verde al terminar cada una, en primer plano y por objetivos si hace falta (`make fmt-check`,
  `make lint`, `make test`, `make test-integration`, `make vuln`, `make schema-check`, `make skills-check`…). Ninguna
  usa la red más que la de las herramientas de Go, ninguna graba (no hay datos externos: plan.md «Datos externos»; ni
  manifiesto `grabaciones.json` ni `TestGrabar*`), ninguna escribe en `data/` ni fuera del repositorio y de `$TMPDIR`,
  ninguna abre una sesión con modelo ni ejecuta `TestSondeoDelLaboratorio`, `make evals-laboratorio`, `make evals`,
  `make evals-sondeo` ni los guiones `scripts/evals*.sh` desde la terminal (FR-014).
- No se tocan `skills/`, `evals/`, `laboratorio/`, `.github/workflows/evals.yml`, ningún umbral, el binario `kitlegal`
  ni sus órdenes, `cmd/laboratorio` ni ningún fichero de `testdata/` que ya exista (plan.md «Structure Decision»;
  data-model §5). Las suites congeladas de H39, H40 y H41 (`specs/031-h39-c-omo-se/aceptacion/`,
  `specs/032-h40-por-qu-e/aceptacion/`, `specs/034-h41-las-skills-del/aceptacion/` y sus copias activadas en
  `internal/app/testdata/script/`) siguen en verde sin cambiar un fichero suyo (FR-032).
- Los textos son los de los contratos: la lista, su JSON y su párrafo, los de [contracts/informe.md](contracts/informe.md)
  §1 a §3; los argumentos, sus mensajes, la copia y `sondeo.json`, los de [contracts/arnes.md](contracts/arnes.md) §2 y
  §3; las columnas, su forma y el detalle, los de [contracts/tabla.md](contracts/tabla.md) §2 a §4. Nada se inventa
  fuera de ellos.
- `internal/evals` no nombra `os.Stdin`, `os.Stdout`, `os.Stderr` ni llama a `os.Exit` (R4, R5); ningún `panic`; los
  errores de los argumentos del arnés, juntos con `errors.Join` como hoy; lo que llega manipulado (un transcript que
  Claude Code no escribe, un `sondeo.json` editado a mano) es la regla genérica de hoy, sin comprobación, mensaje ni test
  propios (research D1, D5). Ningún `//nolint`. Tests de tabla con `t.Parallel()`.
- Cada control se ve fallar con un mutante (la medida pasada del umbral) antes de darlo por bueno, y el mutante se
  revierte (ADR 0029). Los prototipos y guiones de un solo uso van bajo `$TMPDIR` y no se versionan. Los fixtures
  nuevos no llevan nada con forma de secreto (gitleaks corre en `make ci`).

---

## Phase 1: Setup — la suite de aceptación

**Propósito**: fijar, antes de cualquier código, la tabla que el guion del laboratorio tiene que dar con las sesiones
fijas de `normativa-producto` (ADR 0018; plan.md «Aceptación e2e»; contracts/aceptacion.md).

- [X] T001 [aceptacion] Escribir, a partir del spec y de [contracts/aceptacion.md](contracts/aceptacion.md) y [contracts/tabla.md](contracts/tabla.md) §4 y §6, sin código de producto y sin evals nuevas (ninguna skill cambia: plan.md «Structure Decision»), el guion testscript specs/036-h41-2-por-qu-e/aceptacion/laboratorio-tabla.txtar (historia 3; el workflow lo activa como `h41-2-laboratorio-tabla.txtar` en la carpeta de guiones de internal/app al terminar el bucle y lo ejecuta `TestEntregaDelHito`; el arnés e2e no cambia: `KITLEGAL_LABORATORIO_BIN` es el `laboratorio` del árbol). **Cabecera**: un comentario que nombra lo que cubre, FR-020, FR-021, FR-022, FR-024 y SC-005, sin los textos `usage: `, `unknown command`, `unexpected command` ni `cannot parse`. **Órdenes**, en este orden: `exec sh -c '"$KITLEGAL_LABORATORIO_BIN" tabla activaciones > activaciones.out || true'`, `cmp activaciones.out activaciones-tabla.md`, `exec $KITLEGAL_LABORATORIO_BIN tabla activaciones`, `cmp stdout activaciones-tabla.md` y `! stderr .`. **Ficheros del archivo**: bajo `activaciones/`, los tres sondeos de contracts/tabla.md §6 —`h41` (`sondeo.json` `{"modelo":"claude-sonnet-5-5","claude_code":"2.1.317","referencias_del_producto":[…]}` sin `skills_instaladas`), `h41-2-a` (`"claude_code":"2.1.284"`, `"skills_instaladas":["normativa"]`) y `h41-2-b` (`"claude_code":"2.1.284"`, `"skills_instaladas":["consulta-juridica","jurisprudencia","normativa","plazos-y-territorio"]`), los tres con `referencias_del_producto` = `buscar_la_norma.md`, `como_se_cita.md`, `con_herramientas.md`, `lecturas_que_fallan.md`, `normas.md`, `redaccion_en_una_fecha.md`, `redaccion_modificada.md` y `varios_bloques.md`, una línea con su salto final (data-model §3)— y sus ocho sesiones `normativa-producto/herramienta/04-lgt-prescripcion-<nn>/` con lo que hace cada una según la tabla de contracts/tabla.md §6 (`h41-2-a` `-01` y `-02`; `h41-2-b` `-01` a `-05`; `h41` `-01`), cada una con `codigo-de-la-sesion`, `sesion.err` y `sesion.jsonl` con la forma de las sesiones fijas de la familia `normativa` de H41 que lee `TestTablaDeLasFamilias`: `system/init` con el modelo y la versión de su sondeo, un `assistant` con su `message.id` propio por respuesta y sus bloques `tool_use` (`Skill` con `input.skill`; `mcp__kitlegal__boe_articulo` en `h41-2-a` y `boe_articulo` en `h41-2-b`, los dos con una consulta de la eval 04), los `tool_result` en mensajes `user`, y `result` con la respuesta; las dos sin terminar (`h41-2-b/…-05` y `h41/…-01`), código `124` y sin `result`; ninguna abre una referencia; y `activaciones-tabla.md`, la tabla de contracts/tabla.md §4 byte a byte (título, tabla, «## Detalle» y su tabla, con su salto final). El guion falla hoy en el primer `cmp` por una aserción (el guion de hoy ignora `skills_instaladas` y da la tabla sin las columnas nuevas: research V11), lo que el workflow comprueba antes de congelarlo; ninguna tarea posterior lo toca.

**Checkpoint**: la suite está escrita y en rojo por una comparación; desde aquí queda congelada.

---

## Phase 2: Foundational — una sola lectura de las activaciones

**Propósito**: la definición de FR-001, una vez, para el informe del job (US1) y para la tabla (US3).

- [X] T002 La lectura compartida de los usos y la regla de la primera herramienta en un fichero propio, con la lista de activaciones y su test (FR-001, FR-020; SC-001; contracts/informe.md §1; data-model §1; research D1, D2). **Código**: internal/evals/activaciones.go, nuevo, con, movidos sin cambiar de internal/evals/laboratorio_tabla.go, `usoDelTranscript`, `leerLosUsos` y sus ayudantes `usosDelMensaje`, `bloqueConTipo`, `conResultado`, `activaAntes` y `skillDelUso`, con un único cambio: `leerLosUsos` deja de exigir `message.id` en cada `assistant` (el turno se decide con la misma regla de H39 y el id tal como se lee, también vacío; se retira el error «el mensaje assistant no tiene message.id»); la regla de `esDeKitlegal` y `consultaKitlegal` de la tabla pasa, sin cambiar, a `reconocedorDeKitlegal{herramientas, interprete}` con su método `esDeKitlegal(uso, modo)` —en `herramienta`, un uso cuyo nombre, o lo que sigue a su último `__`, es una herramienta del registro; en `orden`, una orden de `Bash` con una orden simple que invoca `kitlegal` con un verbo que da herramienta y consulta; en cualquier otro modo, incluido el `""` de una eval sin binario ni servidor, ninguno— y una función que lo construye con las herramientas del registro y `nuevoInterprete()`, como hoy la tabla; y, nuevos, `Activacion{Skill string; Antes bool}` (claves JSON `skill` y `antes`) y `activacionesDeLosUsos(usos, esDeKitlegal)`: el turno de la primera herramienta y, por cada uso de `Skill` en orden, su `Activacion` (antes si su turno es anterior al de la primera, o si no hay ninguna; en el mismo turno, no; una entrada por bloque, también si repite skill), nunca nil (`[]Activacion{}` sin ninguna). En internal/evals/laboratorio_tabla.go, `contadorDelLaboratorio` guarda un `reconocedorDeKitlegal` en lugar de sus `herramientas` e `interprete` y lo usa donde hoy usa la regla; nada más de la tabla cambia. En .golangci.yml, en `linters.settings.misspell.ignore-rules`, la entrada `activacion`, primera por orden alfabético, con su comentario como las demás (`misspell`, con su diccionario solo inglés, lee el tipo `Activacion` como «Activation»; la convención del fichero, `.golangci.yml:294-295`, es neutralizar cada falso positivo del español con una entrada, nunca con `//nolint:misspell`, del que el repositorio no tiene ninguno; gates/tarea-T002.md). **Test**: internal/evals/activaciones_test.go, nuevo, `TestActivacionesDeLosUsos`, de tabla con `t.Parallel()`, con usos escritos en el test, en los modos `herramienta`, `orden` y `""`: una activación antes de la primera herramienta, una en el mismo turno (después), una sin ninguna herramienta (antes), una skill activada dos veces (dos entradas), una orden `kitlegal` de `Bash` en una sesión de modo `""` (antes), `which kitlegal` y `ToolSearch` que no cuentan como primera herramienta, y ninguna activación (`[]`, no nil). Mutantes que se ven en rojo y se revierten: el mismo turno contado como antes; la orden `kitlegal` de `Bash` contada en el modo `""`; nil en lugar de `[]`. Sin cambiar ningún valor esperado, siguen en verde `TestTablaDeLasFamilias`, `TestTablaDeNormativa`, `TestTablaDelLaboratorio`, `TestLecturaConCd`, `TestReglaTestigo`, `TestOrdenesDeBash` y las suites congeladas de H39, H40 y H41 (FR-022, FR-023, FR-032; SC-006). `make ci` en verde.

**Checkpoint**: la lista de activaciones de una sesión se calcula con una sola regla; la tabla de hoy sale igual.

---

## Phase 3: User Story 1 — El informe del job dice qué skills activó cada sesión (P1)

**Goal**: cada sesión de `informe.json` y de `informe.md` lleva su lista de activaciones, y nada más del informe cambia.

**Independent Test**: `TestInformeConLasActivaciones` sobre las diez sesiones fijas de los dos modos, y
`TestInformeDeLaSuiteCompletaNoCambia` con sus dos constantes de hoy.

- [X] T003 [datos] [US1] Las sesiones fijas de los dos modos y las aserciones de lo que no cambia del informe, en una sola tarea porque el test es el que lee esas sesiones, cubre código que ya existe (`informe.go` de hoy, sin tocarlo) y sin él los fixtures no prueban nada (FR-004, FR-006; SC-002; contracts/informe.md §5; data-model §5; research D6, D14; plan.md «Diseño», el orden de trabajo del informe). **Fixtures**: internal/evals/testdata/sesiones/activaciones/, nuevo, con `evals/22-lcsp-contrato-menor-en-2019.yaml`, copia byte a byte de evals/normativa/22-lcsp-contrato-menor-en-2019.yaml (`activa: true` y `redaccion_fechada`, con la que el informe publica sin juez `sin_activar:claude-sonnet-5-5:<modo>`: research V10), y `sesiones/` con diez sesiones, `orden-antes`, `orden-ninguna`, `orden-otra`, `orden-las-dos`, `orden-mismo-turno` y las cinco `herramienta-…` de los mismos nombres, cada una con lo que hace según la tabla de contracts/informe.md §5 (la consulta, `Bash` `kitlegal boe articulo … --json` en modo orden y `mcp__kitlegal__boe_articulo` en modo herramienta; `Skill` `normativa` o `consulta-juridica` en el turno que dice la tabla; `mismo-turno`, `Skill` y la consulta en la misma respuesta), con la forma de las de internal/evals/testdata/sesiones/informe/ (modo orden) y de internal/evals/testdata/sesiones/leer-llamadas/ (modo herramienta, con `servidor.json`) y un `message.id` propio por respuesta. Ningún fichero de `testdata/` existente cambia; nada se graba. **Test**: en internal/evals/informe_test.go, `TestInformeConLasActivaciones`: escribe con `EscribirInforme`, en un directorio de `t.TempDir()`, el informe de esas sesiones con la skill `normativa`, el modelo que decide `claude-sonnet-5-5` y los dos modos, y exige con valores literales escritos en el test, sesión a sesión, `activada` (sí en `antes`, `las-dos` y `mismo-turno`; no en `ninguna` y `otra`), `sin_activar:claude-sonnet-5-5:orden` y `:herramienta` con 2 de 5, las tasas, los motivos y el veredicto que da hoy el informe; estas aserciones pasan con el `informe.go` de hoy, que la tarea no toca. Mutante que se ve en rojo y se revierte: en internal/evals/informe.go, una sesión que activa otra skill contada como activada (2 de 5 pasa a 1 de 5). `make ci` en verde.
- [X] T004 [US1] La lista de activaciones en `informe.json` y en `informe.md` (FR-001, FR-002, FR-003, FR-005, FR-006; SC-001, SC-002; contracts/informe.md §1 a §3 y §5; data-model §1 y §2; research D3, D4, D5, D6). **Código**: internal/evals/juzgar.go, `ResultadoDeEval` gana `Activaciones []Activacion` con `json:"activaciones,omitzero"`, detrás de `Activada`; internal/evals/informe.go: `EscribirInforme` construye el `reconocedorDeKitlegal` una vez (sus errores, como los de hoy al preparar el informe); de cada sesión que `LeerSesion` leyó, también de la ilegible por su eval, su traza o su modo, la lista de `activacionesDeLosUsos` con los usos de `leerLosUsos` y el reconocedor en el modo de la sesión, junto a donde hoy se fija `Activada` (`informe.go`, `juzgada.resultado.Activada`); de una sesión sin leer, nil y sin clave; un error al leer la lista es un motivo más, `sesión ilegible: <error>`, y la lista queda nil; en la sección de cada sesión de `informe.md`, detrás de «Modelo de la sesión:», el párrafo `Skills activadas: <texto>` con `textoDeLasActivaciones`, nueva en internal/evals/activaciones.go (`normativa (antes), consulta-juridica (después)`; `ninguna` con la lista vacía), y `Skills activadas: sin leer` sin lista, con el `sinLeer` de hoy. `activada`, la tabla «Sesiones» con su columna «Activada», las tasas, los umbrales, los motivos y el veredicto se calculan como hoy. **Tests**: en internal/evals/informe_test.go, `TestInformeConLasActivaciones` gana, sesión a sesión, la lista exacta de `informe.json` y el párrafo de `informe.md` de las diez sesiones de T003 según contracts/informe.md §5 (`antes`: `normativa (antes)`; `ninguna`: `[]` y `ninguna`; `otra`: `consulta-juridica (antes)`, sin `normativa`; `las-dos`: `consulta-juridica (antes), normativa (después)`; `mismo-turno`: `normativa (después)`), sin cambiar ninguna de sus aserciones de T003; `TestInforme`, su caso `sesion-ilegible` exige además que el resultado no lleve `activaciones` y que su sección diga `Skills activadas: sin leer`, sin cambiar ningún otro valor esperado; `TestInformeDeLaSuiteCompletaNoCambia` quita de lo escrito, además del bloque `origen` y de «Medidas y reutilizadas» de H30, las claves `activaciones` de `informe.json` y los párrafos `Skills activadas:` de `informe.md`, y lo compara byte a byte con `informeJSONDeLaSuiteCompleta` e `informeMDDeLaSuiteCompleta`, que no cambian; sus casos de lo que se quita ganan los que fijan que se quita eso y nada más; un test de `textoDeLasActivaciones` en internal/evals/activaciones_test.go con la lista vacía, una entrada y dos; en internal/evals/sondeo_test.go, `juicioSinLaTraza` deja también `Activaciones` en nil, con su comentario, y nada más de `TestJuicioDelSondeo` cambia: el sondeo de `make evals-sondeo` juzga con su propio `juzgarSesionDelSondeo` (internal/evals/sondeo.go, que no cambia), que no lee la lista porque el hito no la pide en el sondeo, y el test compara su juicio con el `ResultadoDeEval` del job, que desde esta tarea la lleva (gates/tarea-T004.md). Mutantes que se ven en rojo y se revierten: «antes» y «después» cambiados en el párrafo; `omitzero` quitado (la sesión ilegible escribe `"activaciones":[]`); el párrafo de una sesión sin lista dado como `ninguna`; un byte más en la sección de una sesión fuera del párrafo (rojo en `TestInformeDeLaSuiteCompletaNoCambia`). Sin cambiar ningún valor esperado siguen en verde los demás tests del informe, de `LeerSesion`, de la serie de cierres y de los umbrales, y `TestJuzgarSinLaLista`, que lee informes versionados de H7.1 a H7.3 (FR-005). `make ci` en verde.

**Checkpoint**: el informe del job dice, por sesión, qué skills activó y cuándo; lo demás del informe sale igual.

---

## Phase 4: User Story 2 — El arnés instala, si se le pide, las skills del producto como el job (P1)

**Goal**: `make evals-laboratorio` gana `INSTALACION=producto` y `EVAL=<nn>`, y cada sondeo de una familia del producto
deja en `sondeo.json` las skills que tenía instaladas.

**Independent Test**: `TestArnesConLaInstalacionDelProducto` y `TestArnesConUnaEval` con el sustituto de `claude`, y
`TestArnesDeLasFamilias`, `TestArnesDeNormativa` y `TestArnesDelLaboratorio` sin cambiar lo que exigen de cada sesión.

- [X] T005 [US2] La constancia de las skills instaladas en todo sondeo de una familia del producto, sin la opción todavía (FR-012, FR-015; SC-003; contracts/arnes.md §3 punto 3 y §5; data-model §3; research D11). **Código**: internal/evals/laboratorio_tabla.go, solo el tipo `sondeoLeido`, que gana `SkillsInstaladas []string` con `json:"skills_instaladas,omitempty"` detrás de `ReferenciasDelProducto` (la tabla aún no lo usa); internal/evals/laboratorio_sondeo.go: `crearElSondeoDelLaboratorio` escribe, en un sondeo de una familia del producto, `skills_instaladas` con los nombres de las entradas de `skills/<variante>/` del temporal en orden de nombre (sin la opción, `["<skill>"]`), y en uno de H39, el `sondeo.json` de hoy; ni la copia ni su huella se guardan en `DIRECTORIO`. **Tests**: en internal/evals/laboratorio_sondeo_test.go, `exigirElSondeoDelArnes` espera en las familias del producto el `sondeo.json` de hoy con `"skills_instaladas":["<skill>"]` detrás (también con `referencias_del_producto`), y en H39 el de hoy; lo que exigen de cada sesión `TestArnesDeLasFamilias`, `TestArnesDeNormativa` y `TestArnesDelLaboratorio`, y las sesiones que abren, no cambian. Mutantes que se ven en rojo y se revierten: la clave escrita también en H39 (rojo en `TestArnesDelLaboratorio`); la clave quitada en `<skill>-control` (rojo en `TestArnesDeLasFamilias`). `make ci` en verde.
- [X] T006 [US2] `INSTALACION=producto`: la instalación de las cuatro skills del producto en cada sesión, sus errores y su paso por la cadena del sondeo (FR-010, FR-011, FR-012, FR-014, FR-015; SC-003, SC-004; contracts/arnes.md §1 a §3 y §5; data-model §3; research D7, D9, D10). **Código**: internal/evals/laboratorio_sondeo.go: `ArgumentosDelLaboratorio` gana `Instalacion`, detrás de `Directorio`, tal como llega; `comprobar` da, detrás del error de `DIRECTORIO` y juntos con los de hoy, `INSTALACION: «<v>» no es producto` siempre que el valor no es vacío ni `producto`, con cualquier `VARIANTES`, e `INSTALACION: las variantes del laboratorio de H39 no son de ninguna familia del producto` solo si las variantes que valen son de la familia de H39 (sin familia, porque ninguna vale o las que valen mezclan familias, no se da: contracts/arnes.md §2); con cualquiera no se abre ninguna sesión, no se construye nada y no se crea `DIRECTORIO`; `laboratorioComprobado` gana `conElProducto bool`; con él, `copiarLasSkillsDeLaVariante` deja en `skills/<variante>/` del temporal, además de la copia de la variante de hoy, la de cada directorio de `skills/` del árbol que no es la skill de la familia, con `os.CopyFS` como `copiarLasSkills` (en `<skill>-producto`, `skills/` entera), una vez en `preparar`, antes de abrir ninguna sesión; `skills_instaladas` da entonces los cuatro nombres; sin la opción, lo de hoy byte a byte. internal/evals/job_test.go: `TestSondeoDelLaboratorio` (`-tags evals`, que `make lint` compila y `make ci` no ejecuta) gana la bandera `-instalacion` y la pasa a `ArgumentosDelLaboratorio.Instalacion`. scripts/evals-laboratorio.sh exige ocho argumentos y da el octavo a `go test` como `-instalacion`; Makefile, la receta de `evals-laboratorio` pasa `"$(INSTALACION)"` detrás de `"$(DIRECTORIO)"`. La comprobación de la versión de Claude Code y del modelo, y su anotación, no cambian (FR-014). **Tests**: en internal/evals/laboratorio_sondeo_test.go, `TestArnesConLaInstalacionDelProducto`, con la construcción sustituida y el sustituto de `claude` de `TestArnesDelLaboratorio`: con `INSTALACION=producto`, una invocación por familia con `VARIANTES=<skill>-control,<skill>-producto MODOS=herramienta SESIONES=4`: en cada sesión, `claude/skills/` tiene exactamente `consulta-juridica`, `jurisprudencia`, `normativa` y `plazos-y-territorio`; la de la familia es, ruta a ruta y byte a byte, la de `laboratorio/<skill>/<skill>-control/skills/<skill>` o la de `skills/<skill>`, y cada otra, la de `skills/<otra>`; y `sondeo.json` lleva las cuatro en `skills_instaladas`; falla con la primera sesión cuyas skills difieren. Con `INSTALACION=todas`, con `VARIANTES=otra INSTALACION=todas` y con `VARIANTES=control INSTALACION=producto`: 0 sesiones, nada construido, `DIRECTORIO` sin crear y los mensajes de contracts/arnes.md §2 (con `VARIANTES=otra`, `VARIANTES: «otra» no es ninguna variante del laboratorio` y después `INSTALACION: «todas» no es producto`). `TestRecetasDelSondeo`, su caso de `evals-laboratorio` pasa también `INSTALACION=producto` y espera la orden del llavero con los ocho argumentos. `TestGuionDelLaboratorio` (internal/evals/laboratorio_sondeo_test.go:1434-1600, en `make ci`, con el sustituto de go), que fija scripts/evals-laboratorio.sh y hoy le pasa exactamente siete argumentos, pasa a ocho: cada caso de `casosDelGuionDelLaboratorio` pasa un octavo, `producto` en `go-test-sale-con-0-y-directorio-absoluto` y vacío en los demás; `ordenDelLaboratorio` espera `-instalacion` con el octavo detrás de `-directorio <directorio>`, con `-temporal <temporal>` como último argumento, que es donde lo lee `exigirElGoDelLaboratorio` (:1498); el caso `sin-siete-argumentos` pasa a `sin-ocho-argumentos`, con los siete de hoy, que sale con 1, da en la salida de error la línea de uso y no ejecuta go (el mensaje de su aserción `sinGo` dice ocho); la línea de uso del guion (scripts/evals-laboratorio.sh:27) gana `<instalacion>` detrás de `<directorio>`; y los comentarios de `TestGuionDelLaboratorio`, `ordenDelLaboratorio` y `casosDelGuionDelLaboratorio` dicen ocho. Sin cambiar ningún valor esperado de lo que instalan ni de lo que abren, siguen en verde `TestArnesDeLasFamilias`, `TestArnesDeNormativa` y `TestArnesDelLaboratorio`. Mutantes que se ven en rojo y se revierten: la skill de la familia copiada también de `skills/` sobre la variante (rojo con `<skill>-control`); una de las otras tres sin copiar; la comprobación hecha después de crear `DIRECTORIO`; la copia hecha también sin la opción (rojo en `TestArnesDeLasFamilias`); el guion sin pasar el octavo argumento a `go test`, o pasándolo con otra bandera o detrás de `-temporal` (rojo en `TestGuionDelLaboratorio`); el guion que acepta siete (rojo en `sin-ocho-argumentos`). `make ci` en verde.
- [X] T007 [US2] `EVAL=<nn>`: el plan de una sola eval, sus errores, la regla de `SESIONES` con él y su paso por la cadena del sondeo (FR-011, FR-013, FR-014, FR-015; SC-004, SC-009; contracts/arnes.md §1 a §3 y §5; data-model §3; research D8, D9). **Código**: internal/evals/laboratorio_sondeo.go (e internal/evals/laboratorio_familias.go si lo pide la elección de la eval en `evalsDelPlan`): `ArgumentosDelLaboratorio` gana `Eval`, detrás de `Instalacion`, tal como llega; `comprobar`: con `EVAL` no vacío, `SESIONES` es un entero mayor que 0 (`SESIONES: «<n>» no es un entero mayor que 0`, en el sitio del error de `SESIONES` de hoy) y sin él la regla de hoy, decidida por que `EVAL` venga o no, valga o no; detrás de los de `INSTALACION`, `EVAL: «<v>» no es el número de ninguna eval del plan de <skill>` si no es el número de dos cifras de una de las cuatro evals del plan de la familia (la que empieza por `<nn>-`), y `EVAL: las variantes del laboratorio de H39 no tienen el plan de una familia del producto` con variantes de H39, los dos solo con una familia (contracts/arnes.md §2); `laboratorioComprobado` gana `eval`, el fichero de la eval pedida (vacío sin `EVAL`), y `porEval` es `SESIONES` con `EVAL` y `SESIONES/4` sin él; con `EVAL`, el plan de cada par variante × modo son `SESIONES` sesiones de esa eval, de nombre `<eval sin .yaml>-01` a `-<SESIONES>`; sin él, el de hoy byte a byte. internal/evals/job_test.go: `TestSondeoDelLaboratorio` gana la bandera `-eval` y la pasa a `ArgumentosDelLaboratorio.Eval`. scripts/evals-laboratorio.sh exige nueve argumentos y da el noveno a `go test` como `-eval`; Makefile, la receta de `evals-laboratorio` pasa `"$(EVAL)"` detrás de `"$(INSTALACION)"`. **Tests**: en internal/evals/laboratorio_sondeo_test.go, `TestArnesConUnaEval`, con el sustituto de `claude`: `normativa-producto` con `EVAL=04 SESIONES=3`, `plazos-y-territorio-control` con `EVAL=05 SESIONES=2` y `jurisprudencia-producto` con `EVAL=03 SESIONES=1`, en modo herramienta, abren solo `<eval>-01` a `<eval>-<SESIONES>` de esa eval y ninguna más; las dos invocaciones de FR-030 (`VARIANTES=normativa-producto MODOS=herramienta EVAL=04`) con `SESIONES=2`, con `INSTALACION=producto` y sin ella, abren `04-lgt-prescripcion-01` y `-02` con las skills de T006 y de hoy; con `EVAL=07` en `normativa`, `EVAL=4`, `EVAL=04` y `VARIANTES=control`, y `EVAL=04` con `SESIONES=0`: 0 sesiones, nada construido, `DIRECTORIO` sin crear y su mensaje; y el ejemplo de contracts/arnes.md §2 con `VARIANTES=control MODOS=orden SESIONES=0 … INSTALACION=todas EVAL=04` da sus tres líneas en ese orden. `TestRecetasDelSondeo`, su caso de `evals-laboratorio` pasa `INSTALACION=producto EVAL=04` y espera la orden del llavero con los nueve argumentos. `TestGuionDelLaboratorio` pasa de ocho (T006) a nueve argumentos: cada caso de `casosDelGuionDelLaboratorio` pasa un noveno, `04` en `go-test-sale-con-0-y-directorio-relativo` (con el octavo vacío, para que el octavo y el noveno se vean por separado) y vacío en los demás; `ordenDelLaboratorio` espera además `-eval` con el noveno detrás de `-instalacion <octavo>`, con `-temporal <temporal>` como último argumento; el caso `sin-ocho-argumentos` pasa a `sin-nueve-argumentos`, con ocho, que sale con 1, da la línea de uso y no ejecuta go (el mensaje de su aserción `sinGo` dice nueve); la línea de uso del guion gana `<eval>` detrás de `<instalacion>`; y los comentarios de `TestGuionDelLaboratorio`, `ordenDelLaboratorio` y `casosDelGuionDelLaboratorio` dicen nueve. Sin cambiar ningún valor esperado, siguen en verde `TestArnesDeLasFamilias`, `TestArnesDeNormativa`, `TestArnesDelLaboratorio` y `TestArnesConLaInstalacionDelProducto`. Mutantes que se ven en rojo y se revierten: `SESIONES/4` sesiones con `EVAL`; `EVAL=4` aceptado; `SESIONES=3` rechazado con `EVAL` por no ser múltiplo de 4; el plan de `EVAL` aplicado también sin él (rojo en `TestArnesDeLasFamilias`); el guion sin pasar el noveno argumento a `go test` como `-eval`, o pasándolo con otra bandera o detrás de `-temporal` (rojo en `TestGuionDelLaboratorio`); el guion que acepta ocho (rojo en `sin-nueve-argumentos`). `make ci` en verde.

**Checkpoint**: las cuatro invocaciones de quickstart.md §8 son órdenes válidas del arnés y abren lo que piden, comprobado
sin modelo.

---

## Phase 5: User Story 3 — La tabla dice qué skills tenía instaladas cada fila y qué skill activó (P1)

**Goal**: la tabla de una familia del producto con constancia da la instalación en la fila, las columnas de las otras
skills y de ninguna, y la lista de cada sesión en el detalle; sin constancia, la de hoy byte a byte.

**Independent Test**: `TestTablaConLasActivaciones` y el guion congelado de T001; `TestTablaDeLasFamilias`,
`TestTablaDeNormativa` y `TestTablaDelLaboratorio` sin cambiar.

- [ ] T008 [datos] [US3] La tabla con las skills instaladas y las activaciones, con sus sesiones fijas y su tabla esperada, en una sola tarea porque el control de SC-005 compara la salida del guion byte a byte con esa tabla y no hay test sin los fixtures ni fixtures que prueben algo sin el código (FR-020, FR-021, FR-022, FR-023, FR-024; SC-005, SC-006; contracts/tabla.md §2 a §4 y §6; data-model §1, §4 y §5; research D12, D13, D14). **Código**: internal/evals/laboratorio_tabla.go: `claveDeFila` gana `instaladas`, los nombres de `skills_instaladas` del sondeo unidos por «, », o la skill de la familia si el sondeo no la tiene, y `compararClaves` ordena por variante, modo, modelo, versión e instalación, en ese orden; `sesionContada` gana `activaciones []Activacion`, la de `activacionesDeLosUsos` con el reconocedor en el modo de la sesión, solo de una sesión terminada; `familiaDeLaTabla` y `celdasDelProducto` ganan `conLasActivaciones`, verdadero si algún sondeo de la tabla tiene `skills_instaladas`: con él, «Skills instaladas» tras «Claude Code», «Activan otra skill», «Otras skills» y «No activan ninguna» tras «Activan antes», y en el detalle «Activaciones» tras «Activa antes» con `textoDeLasActivaciones` (vacía en una sesión sin terminar), con la forma de contracts/tabla.md §3 (Wilson al 95 %, z = 1,96, un decimal y coma decimal; cada sesión una vez en cada recuento; «Otras skills» en orden de nombre, unidas por «; », o `ninguna`; con `n` = 0, los recuentos vacíos y la instalación llena); sin él, la tabla y el detalle de hoy, byte a byte; la de H39 no lo usa. **Fixtures**: internal/evals/testdata/laboratorio/familias/activaciones/, nuevo: `sondeo/h41/`, `sondeo/h41-2-a/` y `sondeo/h41-2-b/`, copia byte a byte de los sondeos bajo `activaciones/` del archivo de specs/036-h41-2-por-qu-e/aceptacion/laboratorio-tabla.txtar (T001), y `tabla.md`, copia byte a byte de su `activaciones-tabla.md` (la extracción, con un guion de un solo uso bajo `$TMPDIR`); ningún fichero de `testdata/` existente cambia; nada se graba. **Test**: en internal/evals/laboratorio_tabla_test.go, `TestTablaConLasActivaciones`, con `t.Parallel()`: la salida de la tabla sobre `familias/activaciones/sondeo/` es su `tabla.md` byte a byte, con la salida de error vacía y código 0; y sobre un sondeo que el test monta en `t.TempDir()` con 39 copias de `h41-2-a/…-01` y una de `h41-2-b/…-03`, con las cuatro skills en `skills_instaladas`, la fila lleva `40 de 40 · 100,0 % [91,2–100,0 %]`, `1 de 40 · 2,5 % [0,4–12,9 %]`, `consulta-juridica: 1 de 40 · 2,5 % [0,4–12,9 %]` y `0 de 40 · 0,0 % [0,0–8,8 %]` (contracts/tabla.md §4; research V3), como `probarCuarentaSesiones`; falla con la primera fila, recuento, intervalo, nombre o lista que difiere. Comprobar además, sin versionar nada, que `go run ./cmd/laboratorio tabla internal/evals/testdata/laboratorio/familias/activaciones/sondeo` da byte a byte ese `tabla.md`, que es lo que pide el guion congelado. Mutantes que se ven en rojo y se revierten: una skill activada dos veces contada dos veces en «Otras skills»; la instalación ordenada después del modelo; un extremo de Wilson con otro redondeo; las columnas nuevas dadas también sin constancia (rojo en `TestTablaDeLasFamilias` y en `TestTablaDeNormativa`). Sin cambiar ningún valor esperado ni fichero, siguen en verde `TestTablaDeLasFamilias`, `TestTablaDeNormativa`, `TestTablaDelLaboratorio`, `TestLecturaConCd`, `TestReglaTestigo`, `TestOrdenesDeBash` y las suites congeladas de H39, H40 y H41 (FR-022, FR-023, FR-032; SC-006). `make ci` en verde.

**Checkpoint**: la tabla de las cuatro filas de FR-030 se puede contar sin modelo; las tablas de H39, H40 y H41 no cambian;
el guion de T001 pasa contra el código de T008.

---

## Phase 6: User Story 4 — La medida y su ADR, fuera del run (P2)

Sin tareas: FR-030, FR-031 y SC-008 se hacen fuera del run (ADR 0032; FR-014) con lo que entregan T005 a T008 y las
órdenes de quickstart.md §8. El informe final lo deja pendiente por la línea `[alcance] plan: FR-030 y FR-031 …` que ya
está en `gates/supuestos.md` (research D17).

---

## Phase 7: Polish & cross-cutting

- [ ] T009 La documentación de lo que deja de ser cierto (FR-001, FR-002, FR-010, FR-012, FR-013, FR-014, FR-020; DoD §1; research D16): CHANGELOG.md, en *Unreleased*, la lista de activaciones de cada sesión en `informe.json` y en `informe.md` del job, `INSTALACION=producto` y `EVAL=<nn>` de `make evals-laboratorio`, `skills_instaladas` en `sondeo.json` y las columnas nuevas de la tabla de una familia del producto; CONTRIBUTING.md, en «Job de evals», `activaciones` en la lista de lo que lleva `informe.json` y el párrafo `Skills activadas:` de `informe.md`, y en «El sondeo del laboratorio», la tabla de argumentos con `INSTALACION` y `EVAL` (y `SESIONES` entero mayor que 0 con `EVAL`), `skills_instaladas` y las columnas «Skills instaladas», «Activan otra skill», «Otras skills», «No activan ninguna» y, en el detalle, «Activaciones», con las órdenes de contracts/arnes.md §1; en el Makefile, solo el comentario que precede a `evals-laboratorio` («siete argumentos» → nueve, con los dos nuevos y que vacíos son la instalación y el plan de hoy) y su línea `##` de ayuda si nombra los argumentos, sin cambiar la receta (`TestRecetasDelSondeo` sigue en verde); la cabecera de comentario de scripts/evals-laboratorio.sh («Los siete argumentos» → los nueve) y la de scripts/evals-sondeo-llavero.sh («con sus siete» → nueve), sin cambiar código ni órdenes. Ningún ADR (el que lee el sondeo se escribe fuera del run: FR-031, research D16), ninguna fila nueva en docs/SOURCES.md (no hay fuente nueva) y ningún esquema. `make ci` en verde.

---

## Dependencies & Execution Order

### Orden de las fases

- **Fase 1 (T001)**: primero; congela la suite.
- **Fase 2 (T002)**: antes de US1 y US3; mueve la lectura y la regla que usan el informe (T004) y la tabla (T008).
- **Fase 3 (T003 → T004)**: T003 fija con el `informe.go` de hoy lo que no cambia; T004 añade la lista y
  `textoDeLasActivaciones`, sobre la lectura de T002 y las sesiones de T003.
- **Fase 4 (T005 → T006 → T007)**: T005 escribe `skills_instaladas` con una skill; T006 la da con cuatro y pasa
  `INSTALACION` por la cadena (ocho argumentos, en la receta, el guion y `TestGuionDelLaboratorio`); T007 pasa `EVAL`
  (nueve) y prueba las invocaciones de FR-030 con la
  instalación de T006.
- **Fase 5 (T008)**: necesita la regla de T002, `textoDeLasActivaciones` de T004, el campo `SkillsInstaladas` de T005
  y el archivo de T001. Tras T008, `h41-2-laboratorio-tabla.txtar` pasa.
- **Fase 7 (T009)**: al final; documenta lo de T004 a T008.

### Dentro de cada tarea

Test y la implementación mínima que lo pone en verde, en la misma tarea (T003, el test sobre el `informe.go` de hoy con
sus fixtures; T008, el código, sus fixtures y su test); el mutante que ve fallar el control, revertido; `make ci` en
verde antes de marcarla.

### Oportunidades de paralelismo

Ninguna tarea lleva [P]: T002, T005 y T008 comparten internal/evals/laboratorio_tabla.go; T002, T004 y T008,
internal/evals/activaciones.go; T003 y T004, internal/evals/informe_test.go; T005 a T007,
internal/evals/laboratorio_sondeo.go, y T006, T007 y T009, el Makefile y scripts/evals-laboratorio.sh. Se ejecutan en
serie, en el orden de los ids.

## Parallel Example

```text
Ninguno: el bucle del workflow ejecuta y verifica las tareas una a una, en el orden de los ids.
```

## Implementation Strategy

### MVP

T001 → T002 → T003 → T004: el informe del job dice por sesión qué skills activó, que es lo que lee la medición del
cierre si la eval 04 sale en rojo (FR-033).

### Entrega incremental

T005 a T007 dan el arnés de las cuatro filas de FR-030; T008, su tabla y el guion congelado en verde; T009, la
documentación. Cada tarea deja `make ci` en verde.

## Notes

- 9 tareas: T001 [aceptacion]; base: T002; US1: T003, T004; US2: T005, T006, T007; US3: T008; US4: ninguna; T009.
- Tareas [datos]: T003 (internal/evals/testdata/sesiones/activaciones/) y T008
  (internal/evals/testdata/laboratorio/familias/activaciones/), las dos sin red; cada una dice en su línea por qué va
  con su test. El hito no tiene datos externos (plan.md «Datos externos»): ni manifiesto `grabaciones.json`, ni
  `TestGrabar*`, ni grabación.
- Definition of Done (docs/ROADMAP.md §1): `make ci` y tests offline en cada tarea; ningún sobre ni esquema cambia (sin
  tarea de `schemas/`); e2e: T001; CHANGELOG: T009; ADR: no aplica en el run (el de FR-031 se escribe fuera, research
  D16 y D17); `docs/SOURCES.md`: no aplica (ninguna fuente); fixtures: T003 y T008; skills con evals: las de hoy, sin
  cambiar, que mide el job del cierre (FR-033); dimensión territorial: ningún dato ni código nombra un municipio.
- Cada fila de «Controles de umbral» de plan.md y la tarea que construye o toca su control y lo ve fallar con un
  mutante:
  - SC-001, FR-001, FR-002, FR-003, FR-006 — `TestActivacionesDeLosUsos`: T002 (mismo turno como antes; `kitlegal` en
    modo `""`; nil); `TestInformeConLasActivaciones`: T004 («antes» y «después» cambiados; `omitzero` quitado).
  - SC-002, FR-004 — `TestInformeConLasActivaciones`: T003 (otra skill contada como activada);
    `TestInformeDeLaSuiteCompletaNoCambia`: T004 (un byte más fuera del párrafo).
  - SC-003, FR-010, FR-012, FR-015 — `TestArnesConLaInstalacionDelProducto`: T006 (la de la familia copiada de
    `skills/`; una sin copiar; la copia sin la opción); `TestArnesDeLasFamilias`, `TestArnesDeNormativa`,
    `TestArnesDelLaboratorio` (de hoy, `exigirElSondeoDelArnes` con la clave): T005 (clave en H39; clave quitada).
  - SC-004, FR-011 — `TestArnesConLaInstalacionDelProducto`: T006 (comprobación después de crear `DIRECTORIO`);
    `TestArnesConUnaEval`: T007 (`EVAL=4` aceptado); `TestArnesDeLasFamilias` (de hoy, sin cambiar sus errores): T006 y
    T007.
  - SC-009, FR-013 — `TestArnesConUnaEval`: T007 (`SESIONES/4` con `EVAL`; `SESIONES=3` rechazado);
    `TestArnesDeLasFamilias`: T007 (el plan de `EVAL` sin él).
  - SC-005, FR-020, FR-021, FR-024 — `TestTablaConLasActivaciones`: T008 (doble recuento; orden de la instalación;
    redondeo); `TestEntregaDelHito` con `h41-2-laboratorio-tabla.txtar`: T001, que lo ve fallar hoy en su primer `cmp`,
    y T008, que lo pone en verde.
  - SC-006, FR-022, FR-023, FR-032 — `TestTablaDeLasFamilias`, `TestTablaDeNormativa`, `TestTablaDelLaboratorio` y las
    suites de H39, H40 y H41 en `TestEntregaDelHito` (de hoy, sin cambiar): T002, que mueve la lectura y retira la
    exigencia de `message.id` con ellas en verde, y T008 (las columnas nuevas sin constancia, en rojo).
  - SC-007, FR-033 — las filas `evals:` de `normativa`, `jurisprudencia` y `plazos-y-territorio`: las mide el job de
    evals del cierre, que hace el workflow tras la revisión final con los umbrales de hoy, sin tarea que los toque.
- Requisitos sin tarea propia: FR-030, FR-031, SC-008 (fuera del run; US4); FR-033 y SC-007 (el cierre del workflow);
  FR-032 (todas: ninguna declara una ruta de las suites congeladas, y lo guarda el guardián de diff).
