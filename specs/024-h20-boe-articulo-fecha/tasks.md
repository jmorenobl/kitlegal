# Tasks: H20 · `boe articulo --fecha`: la redacción vigente en una fecha + `boe-legislacion` v0.2

**Input**: `specs/024-h20-boe-articulo-fecha/` (spec.md, plan.md, research.md, data-model.md, contracts/,
quickstart.md)

**Prerrequisitos**: spec.md y plan.md leídos; `.specify/memory/constitution.md` 2.13.0 aplicada (umbral de
materialidad, criterio de uso, proporcionalidad y convergencia en «Gates»; controles de umbral del ADR 0029; ADR 0005,
0014, 0018, 0023, 0032, 0037 y 0038); #168 en `main` (`boe-legislacion` a cinco sesiones a la vez en el job, veintiuna
evals, trece umbrales en `boe-legislacion`, quince en `jurisprudencia` y seis en `legal-core`). **Ninguna tarea crea el
esqueleto ni los gates: `make ci` existe y pasa**, y cada tarea lo deja en verde al terminar.

**Aceptación**: seis guiones `testscript` (plan.md, «Aceptación e2e»), que escribe T001 en
`specs/024-h20-boe-articulo-fecha/aceptacion/` y que el workflow congela y, tras el bucle, activa como
`h20-<nombre>.txtar` en el directorio de guiones del arnés e2e de `internal/app`. Cubren US1.1, US2.1, US3.1 y US4.1.
Lo que no tiene guion, y por qué: **US5** no es comportamiento del binario; se acepta con `TestUmbralesDeBoeLegislacion`
y `TestJuzgarLecturasFechadas` en `make ci` (T008 y T010) y con el job de cierre, que lanza el workflow tras la
revisión final (SC-001). **US1.2, US2.2 y US3.2** son respuestas de la skill: las miden las evals 22 y 23 y las
veintiuna de hoy en el job de cierre, con el umbral de T010, y la serie de 2 de 3 de cada eval (FR-057). Ninguna eval
entra en T001: las dos entran con el formato que las admite y su umbral (T008 y T010), porque `make ci` valida toda
eval de la carpeta de evals (precedente: H9).

**Tests**: obligatorios (constitución §III; spec FR-080 a FR-090; plan.md, «Controles mecánicos»). Cada tarea trae sus
tests y la implementación que los hace pasar en el mismo diff; los nombres de test y sus ficheros son los de plan.md,
«Tests nuevos» y «Tests y guiones existentes que cambian», porque el informe final busca cada control
`ci:<ruta>:<Test>` como `^func <Test>\(` en esa ruta.

**Rebanadas verticales y excepciones declaradas.** Son rebanada completa —test e implementación en el mismo diff,
`make ci` en verde por sí solas— T002 a T011. Las demás, cada una por su razón:

- **T001** es la suite congelada: sin código de producto; el «rojo primero» del workflow la verifica.
- **T012** es la prueba de FR-090 sobre el código y los tests que dejan T005, T007 y T010: aplica los cinco mutantes
  en un árbol aparte fuera del repositorio y deja su registro; no cambia nada bajo `internal/` (plan.md, «Obligaciones
  para tasks.md»). Cada control ya se vio en rojo con su mutante en la verificación de la tarea que lo construye.
- **T013** es documentación; su verificación es que cada cifra, fichero, test y umbral que nombra existe en el árbol,
  y `make ci`.
- **T014** es el cierre de la Definition of Done, sin código de producto.

**Tareas `[datos]` con código** (precedente H7.2 a H9): T006 y T008 tocan el directorio de esquemas y además código,
porque `make ci` no queda en verde si entran por separado: con la bandera `--fecha` registrada, `schema-check` exige
`schemas/bloque.json` regenerado y `skills-check` la tabla regenerada de `boe-legislacion` (T006); y el formato de eval
no admite `fecha` ni `redaccion_fechada` sin el esquema que las valida, ni el esquema sin el tipo que el lector estricto
lee (T008). Ninguna tarea toca un directorio de fixtures: las respuestas sin versiones o con una versión sin
`fecha_vigencia` (FR-004) se construyen en el test, y lo demás sale de lo grabado de H4 (plan.md, «Fixtures»; research
D22). No hay manifiesto de grabación nuevo, y el paso `grabar_datos` no tiene nada que grabar.

**Lo que este fichero añade al plan, comprobado en el árbol en esta sesión** (lectura y `git grep`, sin ejecutar ningún
test):

- **El paso 4 del plan se parte en tres tareas** (T004, `ValidarFecha`; T005, la lectura fechada, la elección, la
  salida y la caché; T007, lo que entrega al grafo, con el test del applet que solo puede escribirse cuando existe la
  bandera de T006), **y el paso 7, en dos** (T009, la caché preparada de las evals; T010, las evals con fecha y el
  umbral), para que cada sesión de un paso tenga un diff acotado; cada parte deja `make ci` en verde.
- **`TestClavesDeLosDatos`** (`internal/source/boe/datos_test.go:129`) fija hoy que el verbo `articulo` describe
  exactamente `boe.Articulo` y `boe.Aviso`, con todas sus claves obligatorias: con la `Salida` de `articulo` en
  `boe.ArticuloFechado` (research D5) cambia en la misma tarea que la `Salida` (T006), que declara ese fichero.
- **`ConsultasNecesarias`** (`internal/evals/consultas.go:90`) la usan también `preparar.go` y `medida.go`, y sus
  tests nombran consultas `articulo BOE-A-…` literales (`preparar_test.go`, `medida_test.go`): T009 los declara por si
  la lectura con `--fecha 9999-12-31` cambia sus listas; ni `medida.go` ni la medida del juez cambian (FR-058).
- **Las cifras de las evals del repositorio** viven en dos tests, y los dos cambian en T010, que los declara:
  `internal/evals/definicion_test.go` (`peorCasoDeBoeLegislacion`, línea 1874, y su caso de la línea 1952, con
  `SesionesPorGrupo` 97, 96 y 6) y `internal/evals/plan_test.go` (`probarElPlanDelRepositorio`, líneas 285-306: series
  32, 32 y 2 y sesiones 96, 96 y 6 de `boe-legislacion`, que pasan a 36, 36 y 2 y a 108, 108 y 6). Buscadas también las
  cifras de series y sesiones (32, 36, 96, 97, 108, 109) y las de evals (21, 23) en los tests que leen `evals/`
  (`conjunto_test.go`, `job_test.go`, `medida_test.go`, `sondeo_test.go`, `informe_test.go`, `serie_test.go`): los
  demás recuentos son de ejecuciones sintéticas (`ejecucionConUmbrales`, `filaContada`) o de evals nombradas, que no
  dependen del número de evals; `TestConjuntoDeEvals` y `TestEvalsDelRepositorio`, que sí, ya están en T010.
- **T014, el cierre**, no está en los diez pasos del plan: `make ci` deja los perfiles de cobertura y no los compara con
  ningún umbral, así que el punto 9 de la Definition of Done y lo que el hito no puede tocar (FR-063, FR-070) se
  comprueban una vez, sobre el diff entero, antes de la revisión final. Es el cierre de H7.3 a H9.
- Existen hoy con el nombre que les da el plan: `TestLeerBloque`, `TestArticulo`, `TestFallosNoSeGuardan`,
  `TestGolden`, `TestObservadoDeBoe`, `TestValidarNorma`, `TestClavesDeLosDatos` (`internal/source/boe`);
  `TestEntregaDelMontador` y `LoteDe` (`internal/cli/entrega.go:28`); `TestConsolidar` y `Consolidar`
  (`internal/core/grafo/lote.go:82`); `TestLaSalidaDeBoeNoCambiaConElGrafo`, `TestSalidaDeBoeContraSchemas`,
  `TestHerramientasDelServidor` y `TestEsquemasPublicados` (`internal/app`); `TestConjuntoDeEvals`,
  `TestEvalsDelRepositorio` y su subtest `hallazgos-de-la-skill`, `maximoDeEvals` (21, `conjunto.go:227`),
  `TestConsultasNecesarias`, `TestDefinicionDelJob`, `probarElPeorCasoDelTrabajo`, `TestPlanEnDosModos`,
  `probarElPlanDelRepositorio`, `TestUmbralesDeLasTresSkills`,
  `umbralesDeBoeLegislacionComoEnH26`, `escribirSesionSinteticaDelPlan`, `informeDelPlanDelJob`, `leerElArticulo`,
  `respuestasQueSeJuzgan`, `nombradasPorSuSesion` y `fechasDeVigencia` (`internal/evals`). La validación de la norma y
  el bloque está en `internal/source/boe/articulo.go:38-45`, delante de abrir nada. Los nombres que el plan fija como
  nuevos no existen todavía.

**Lo que no se ha medido en esta sesión.** No se ha ejecutado ningún test ni `make ci`. Las cifras de las tareas son
las de research.md (tablas V y M) y las de los contratos, que salen de un prototipo que ya no existe; las tareas las
vuelven a medir con sus tests. El peor caso del job con veintitrés evals, 18 977 s, es el supuesto S1: lo comprueba
`TestDefinicionDelJob` en T010.

**Modo**: desatendido (ADR 0018). El workflow `hito` ejecuta y verifica las tareas **una a una**, con guardián de diff:
cada tarea declara en su propia línea, tras «Rutas:», **todas** las rutas que crea o modifica, y solo toca esas (más
`go.mod`, `go.sum`, `CHANGELOG.md`, el directorio del feature y el `x_test.go` de cada `x.go` declarado). Los demás
nombres que aparecen en una línea son contexto: referencias a los documentos del feature y ficheros que la tarea solo
lee. El cierre en la plataforma —publicar la rama, abrirla a revisión y medir CI y el job de evals— lo hace el workflow
tras la revisión final; quickstart.md §9 queda para el workflow y §10 para después del run.

## Formato: `- [ ] Tnnn [etiqueta?] [Story?] Descripción — FR/SC. Rutas: …`

- **[P]**: ninguna tarea lo lleva; en el bucle del workflow todo va en secuencia.
- **[aceptacion]**: solo T001, la primera.
- **[datos]**: T006 y T008, por la razón de arriba; ninguna toca fixtures ni manifiestos de grabación.
- **[Story]**: US1 la redacción que regía en la fecha del acto · US2 en esa fecha el precepto no tenía redacción
  vigente · US3 la pregunta sin fecha sigue igual · US4 una lectura fechada no engaña al grafo · US5 el job sale en rojo
  si la respuesta no da la redacción fechada. La aceptación y el pulido no llevan historia; una tarea que sirve a varias
  lleva la principal y nombra las demás.

## Batería de verificación por tarea

- **`make ci` en primer plano** y la tarea marcada `[X]` en el mismo turno en que termina en verde: formato · lint
  (`depguard`, `forbidigo`, `gosec`, `misspell`, `revive`, `dupl`, `gocyclo`, `paralleltest`, `unused`…) · tests con
  `-race` y `-shuffle=on` · `test-integration` · `test-tiempos` · `vuln` · `schema-check` · `skills-check` ·
  `goreleaser-check` · secretos · módulos. Ningún registro de la verificación en la raíz del repositorio. Las salidas de
  `go test`, de `git diff` y de `make` se leen con `rtk proxy` o con sondas positivas, porque el proxy de la sesión las
  resume.
- **Sin red ni modelo** (FR-006, FR-072; ADR 0032): ninguna tarea usa la red salvo la de las herramientas de Go
  (`make vuln`), ni graba nada, ni escribe en la carpeta de la evidencia; ninguna ejecuta `claude`, `make evals`,
  `make evals-sondeo`, `make evals-medir-juez`, los guiones `scripts/evals*.sh`, `TestEjecucionDelJob`,
  `TestMedidaDelJuez`, `TestSondeo` ni `TestGrabar*`. Los tests de `make ci` usan sesiones sintéticas, que escriben en
  `t.TempDir()`.
- **Sin personas a mitad**: ninguna tarea la hace, la revisa ni la desbloquea una persona.
- **Nada de memoria** (constitución, «Criterio de decisión autónoma»): ninguna fecha de vigencia, identificador
  `BOE-A-…`, id de bloque, huella ni texto de una norma se escribe de memoria: salen de lo grabado de H4 y de los
  contratos, que los midieron sobre él (research M1 a M4). Si lo grabado contradice a un contrato, manda lo grabado,
  se corrige el artefacto del feature que lo recordaba y se anota en `specs/024-h20-boe-articulo-fecha/gates/supuestos.md`.
- **Ningún umbral se cumple cambiándolo** (FR-071; ADR 0029): ninguna tarea ni corrección cumple un umbral de plan.md,
  «Controles de umbral», cambiando su valor, dejándolo en `decide: false`, sacando sesiones de su `total`, retirando su
  test, quitando un caso de una tabla, ni cambiando las evals, la rúbrica, los casos o la medida del juez. Cada control
  nuevo se ve en rojo antes de darlo por bueno: con el mutante de su fila, aplicado a mano y revertido antes de
  `make ci` (ningún mutante queda en el diff), y en T012.
- **Lo que no cambia** (FR-059, FR-063, FR-070; plan.md, «Obligaciones para tasks.md»): `skills/legal-core/`,
  `skills/jurisprudencia/`, `evals/legal-core/`, `evals/jurisprudencia/`, `evals/boe-legislacion/juez/`,
  `evals/boe-legislacion/expresiones-prohibidas.yaml`, las veintiuna evals de hoy de `boe-legislacion`, `.github/`,
  `scripts/`, `.specify/`, la carpeta de la evidencia, `docs/SOURCES.md`, `docs/ROADMAP.md`, `docs/USO.md`,
  `docs/HISTORIAL.md`, `docs/ADR/`, `web/`, `CLAUDE.md`, `Makefile`, `internal/mcp`, `internal/app/herramientas.go`,
  `internal/cache`, `internal/httpx`, `internal/evals/juez.go`, `medida.go`, `sondeo.go`, `definicion.go`, lo grabado y
  sus manifiestos, los guiones e2e de hoy, y los specs, planes e informes de otros hitos. Ningún valor de concurrencia
  ni ningún tope del job.
- **La lectura sin fecha no cambia** (FR-012): ninguna tarea cambia su camino de código, su entrada de caché, su salida
  ni lo que entrega al grafo; los trece casos de `TestGolden` siguen sin cambiar ni un byte.
- **Ningún test escribe fuera de `t.TempDir()`**, salvo los golden con su bandera de actualizar; lo temporal de una
  tarea va bajo `$TMPDIR`, nunca bajo `/tmp`, y todo árbol aparte (`git worktree add --detach`) va bajo `$TMPDIR`, nunca
  dentro del repositorio, y se retira al terminar.
- **Sin atajos** (plan.md, «Comprobación contra la rúbrica», g): ningún `//nolint`, `t.Skip`, TODO ni error
  silenciado; ningún test se desactiva ni se retira sin el que lo sustituye; ninguno pierde lo que comprobaba. Si
  `dupl`, `gocyclo`, `paralleltest` o `misspell` marcan algo, se reestructura o se renombra.
- **Antes de tocar una función, una constante, un tipo o un ayudante que usan los tests**, `git grep` de sus usos en
  todos los `_test.go` del módulo: lo que haya que alinear entra en la misma tarea, y sus ficheros están en sus rutas.
- **Proporcionalidad** (constitución, «Gates»; plan.md, «Trazabilidad», último párrafo): ninguna tarea añade un caso,
  un mensaje ni un test para un estado sin vía real —una `fecha_vigencia` ilegible (cae en la regla de FR-004), dos
  versiones con la misma vigencia, la entrada de las versiones manipulada a mano—: los cubre la regla genérica (defecto
  `inesperado`, código 1).

---

## Fase 1: Aceptación — la suite congelada de la entrega

**Objetivo**: los seis guiones que describen la lectura fechada desde la terminal y desde un cliente MCP, y la lectura
sin fecha igual a la de `main`, en rojo por una aserción.

**Prueba independiente**: el «rojo primero» del workflow sobre cada guion; después del bucle, `TestEntregaDelHito` con
los seis `h20-*` (quickstart.md §4).

- [X] T001 [aceptacion] Los seis guiones `testscript` de la entrega, escritos desde spec.md (US1.1, US2.1, US3.1 y US4.1) y desde contracts/articulo-fecha.md §1 a §6, sin código de producto, sin datos y sin ninguna eval (plan.md, «Aceptación e2e»: las dos evals entran con el formato que las admite, en T008 y T010). Son `fecha-version.txtar`, `fecha-codigos.txtar`, `fecha-offline.txtar`, `fecha-main.txtar`, `fecha-herramienta.txtar` y `fecha-grafo.txtar`, con esos nombres exactos, porque plan.md, «Controles de umbral», nombra sus copias activadas `h20-<nombre>.txtar`. **Común a los seis**: un comentario de cabecera que nombra los FR, SC y escenarios que cubre y que no contiene los textos `usage: `, `unknown command`, `unexpected command` ni `cannot parse` (el «rojo primero» los toma por un guion ilegible); antes de cualquier otra orden, la precondición `exec kitlegal boe articulo --help` seguida de `stdout '--fecha'`, que hoy se ejecuta bien y falla en su aserción; después, solo las órdenes y variables que el arnés ya tiene (`exec`, `stdout`, `stderr`, `cmp`, `cp`, `rm`, `exists`, `env`, `mcp`, `exec sed` para igualar `fecha_consulta`, `exec sh -c '…; test $? -eq <n>'` para un código exacto, `KITLEGAL_BIN`, `KITLEGAL_T0_BIN` y lo grabado de H4 en `reproduccion/`; precedentes `boe-offline`, `boe-codigos`, `h7-1-grafo-lecturas`, `h21-mcp-llamadas` y `h9-plazos-herramienta`), sin ninguna nueva. Afirman solo lo que el spec fija —códigos, `ok`, la clase, y en `data` `fecha_vigencia`, `fecha_version`, `norma_modificadora`, `fecha_pedida`, `redaccion_posterior` con su fecha de vigencia y su norma, la igualdad de `hash_texto` y `texto` donde el spec la pide, y del mensaje del 3 de FR-003 solo que nombra la fecha pedida y la fecha de vigencia de la primera redacción y que es distinto del 3 de un bloque que no está en la norma—; nada de `fecha_consulta`, ni del texto de un mensaje de 2. **fecha-version** (US1.1, US2.1; FR-001, FR-002, FR-003, FR-010, FR-013; SC-004): con `KITLEGAL_T0_BIN` y `--json`, `boe articulo BOE-A-2017-12902 a1-30 --fecha 2019-06-01` sale con 0, con `"fecha_vigencia":"20180309"`, `"norma_modificadora":"BOE-A-2017-12902"`, `"fecha_pedida":"2019-06-01"` y `"redaccion_posterior":{"fecha_vigencia":"20200206","norma_modificadora":"BOE-A-2020-1651"}`, y sin `--json` la salida legible lleva `2019-06-01`, `20200206` y `BOE-A-2020-1651` en sus filas; `BOE-A-1985-5392 a22 --fecha 2005-06-15`, 0, `20040101`, `BOE-A-2003-23103` y la redacción posterior `20151031`, `BOE-A-2015-11723`; `a22 --fecha 2019-06-01`, 0, el mismo `hash_texto` y el mismo `texto` que `a22` sin `--fecha` y ninguna clave `redaccion_posterior`; `BOE-A-2015-10565 a21 --fecha 2016-06-01`, 3, `"clase":"no-encontrado"`, sin `texto`, con `2016-06-01` y `20161002` en el mensaje. **fecha-codigos** (FR-003, FR-005, FR-007): `--fecha 2019-02-30`, `--fecha 2019-6-1` y `--fecha ''` salen con 2 y `"clase":"argumentos"`, sin crear la caché (`! exists` del fichero de la caché del `HOME` del arnés); con una norma mal formada y con un bloque mal formado, cada uno con `--fecha 2019-06-01`, el mismo código, la misma clase y el mismo mensaje que sin `--fecha` (comparados con `cmp` tras igualar `fecha_consulta`); con `a9999` y `--fecha 2019-06-01`, el mismo 3 y el mismo mensaje que sin ella, y distinto (`! cmp`) del de `a21 --fecha 2016-06-01`; con `--dry-run`, `--fecha 2019-06-01` y la caché vacía, 0 y la caché sin crear. **fecha-offline** (FR-020, FR-021, FR-022, FR-082): tras `a22` sin fecha sobre la reproducción, `a22 --fecha 2005-06-15 --offline` sale con 4; tras `a22 --fecha 2005-06-15` sobre la reproducción, `a22 --offline` sin fecha da byte a byte lo que daba antes de la lectura fechada (`cmp`); retirada `reproduccion/`, `a22 --fecha 2005-06-15 --offline --json` da el mismo `data` y el mismo `hash` que con la reproducción, y `a22 --fecha 2008-04-30 --offline` sale con 0. **fecha-main** (US3.1; FR-012, FR-070, FR-083; SC-005): sin `--fecha`, `boe articulo BOE-A-1985-5392 a22` con `--json` y sin él, el 3 de `a9999` y el 2 de una norma mal formada, cada uno con su código exacto y su salida igual byte a byte, con `cmp`, a la salida de `main` que el guion lleva escrita como sección del txtar (las fechas de consulta que no fija el reloj del arnés, igualadas con `exec sed` en los dos lados, research S3); y `boe articulos BOE-A-2015-10565 a21 a22 --describe` igual a la sección de `main`. Las secciones de `main` salen de la salida del binario del arnés en la cabeza de hoy, sin el código del hito (research D20), capturadas con una copia momentánea `zz-` del guion en el directorio de guiones del arnés de `internal/app` que vuelca cada salida, borrada antes de terminar. **fecha-herramienta** (US1.1, US2.1; FR-001, FR-084): con la orden `mcp` y `KITLEGAL_T0_BIN`, el servidor anuncia `boe_articulo` (su parámetro `fecha` lo compara con `--describe` el control de conformidad, `TestHerramientasDelServidor`, en T006); una llamada con `norma` `BOE-A-2017-12902`, `bloque` `a1-30` y `fecha` `2019-06-01` da `resultado`, y su sobre, quitado `fecha_consulta`, es igual con `cmp` al de la orden con `--fecha 2019-06-01 --json`; una con `BOE-A-2015-10565`, `a21` y `fecha` `2016-06-01` da el error `no-encontrado` con `20161002`; y la llamada siguiente del mismo proceso da `resultado`. **fecha-grafo** (US4.1; FR-030, FR-031, FR-032): con `a22` leído sin fecha, `graph stats` y `graph check BOE-A-1985-5392 a22` antes; tras `a22 --fecha 2005-06-15`, `graph stats` con un nodo `BloqueVersion` más y `graph check BOE-A-1985-5392 a22` byte a byte igual al de antes (`cmp`); con `--no-graph` y con `--dry-run`, otra lectura fechada deja `graph stats` igual; y en un `HOME` aparte con la misma secuencia sin la lectura fechada, la siguiente lectura sin fecha de `a22` da el mismo `graph check` en los dos (`cmp`). **Verificación**: los seis se validan contra un prototipo de un solo uso bajo `$TMPDIR`, fuera del repositorio —un árbol desacoplado de `HEAD` (`git worktree add --detach`) con la lectura fechada mínima de contracts/articulo-fecha.md (la bandera, la elección, los dos mensajes de 3, la entrada de las versiones y la emisión con `SoloLoNuevo`)—, ejecutado con `go -C`: allí los seis pasan, y con un mutante por guion (la versión por `fecha_publicacion`, el 2 de la fecha quitado, la lectura fechada servida de la entrada de hoy, `--describe` de `articulos` con la bandera, la herramienta sin `fecha`, la lectura fechada con `eli:has_version`) el suyo falla; el árbol aparte se retira con `git worktree remove --force` y no se versiona nada de él; en el repositorio, el «rojo primero» da los seis en rojo por la aserción de su precondición; `make ci` en verde, sin ningún fichero fuera del directorio de la suite — FR-001, FR-002, FR-003, FR-005, FR-006, FR-007, FR-010, FR-012, FR-013, FR-020, FR-021, FR-022, FR-030, FR-031, FR-032, FR-070, FR-082, FR-083, FR-084; SC-004, SC-005. Rutas: specs/024-h20-boe-articulo-fecha/aceptacion/

**Checkpoint**: la entrega está descrita y congelada; nada del producto ha cambiado.

---

## Fase 2: Fundación — el grafo acepta lo observado sin pisar lo que guarda

**Objetivo**: `SoloLoNuevo` viaja del `Resultado` a `world.db`, y con él la entrega escribe solo lo que no tiene
(research D9).

**Prueba independiente**: quickstart.md §3, primera y segunda órdenes.

- [X] T002 [US4] `SoloLoNuevo` en los contratos que llevan lo observado al grafo, y la entrega que no toca lo guardado (plan.md paso 2; contracts/articulo-fecha.md §6; data-model §6; research D9, D10). **Contratos** (R1: ningún import nuevo): `schema.Observado` gana `SoloLoNuevo bool` (`internal/core/schema/grafo.go`), `core.Lote` gana `SoloLoNuevo bool` (`internal/core/graphstore.go`), `grafo.Consolidado` lo gana y `grafo.Consolidar` lo copia (`internal/core/grafo/lote.go`), y `cli.LoteDe` lo copia del `Observado` al `Lote` (`internal/cli/entrega.go`); el contrato `Applet` no cambia (ADR 0005). **Almacén** (`internal/graph/aplicar.go`, R3 y R6: SQL de hoy, ningún import nuevo): con `SoloLoNuevo`, un nodo, una arista o un texto que `world.db` ya guarda queda como estaba, byte a byte, sin cambiar su última observación, su vigencia, sus datos ni su `source`; uno nuevo entra con la procedencia del lote; sin `SoloLoNuevo`, todo como hoy. **Tests**: `TestApplySoloLoNuevo` (nuevo, `internal/graph/almacen_test.go`): un nodo, una arista y un texto guardados quedan iguales byte a byte; uno nuevo entra con la procedencia del lote; el mismo lote aplicado dos veces no cambia nada la segunda; `TestIntegracionLecturaFechada` (nuevo, `internal/graph/integracion_test.go`), con lotes sintéticos con la forma del de una lectura fechada (una `BloqueVersion` con su `Texto`, sin `eli:has_version`, con `SoloLoNuevo`) en `t.TempDir()`: con fila de lecturas del bloque y sin ella (el mundo de H7), con `version-obsoleta` y con `fuente-caducada` presentes, la fila de lecturas y lo que da `grafo.Comprobar` iguales tras el lote, también cuando el nodo de la versión ya estaba guardado; y la siguiente lectura sin fecha da el `Comprobar` del mundo sin el lote; `TestEntregaDelMontador` (`internal/cli/entrega_test.go`) gana que el lote lleva `SoloLoNuevo`, y `TestConsolidar` (`internal/core/grafo/lote_test.go`) que el consolidado lo lleva; si el test de los campos de `Observado` (`internal/core/schema/grafo_test.go`) los enumera, gana el nuevo. **Verificación**: con el mutante «`SoloLoNuevo` ignorado en la entrega», `TestApplySoloLoNuevo` y `TestIntegracionLecturaFechada` fallan (revertido antes de `make ci`); quickstart.md §3, primera y segunda órdenes, `ok`; `make ci` en verde — FR-030, FR-031, FR-086; SC-006. Rutas: internal/core/schema/grafo.go, internal/core/graphstore.go, internal/core/grafo/lote.go, internal/cli/entrega.go, internal/graph/aplicar.go, internal/graph/almacen_test.go, internal/graph/integracion_test.go

**Checkpoint**: el grafo puede recibir una observación que no cuenta como lectura sin cambiar lo que ya sabía.

---

## Fase 3: US1, US2 y US3 — la lectura fechada en el adaptador, la orden y la herramienta (P1) 🎯 MVP

**Objetivo**: `kitlegal boe articulo <norma> <bloque> --fecha AAAA-MM-DD` y `boe_articulo` con `fecha` dan la versión
vigente en esa fecha, o el 3 con la primera fecha de vigencia; sin `--fecha`, nada cambia.

**Prueba independiente**: quickstart.md §2 y §5; tras el bucle, `h20-fecha-version`, `h20-fecha-codigos`,
`h20-fecha-offline`, `h20-fecha-main` y `h20-fecha-herramienta`.

- [ ] T003 [US1] El lector del bloque guarda todas sus versiones (plan.md paso 3; data-model §2; research D4, M1). `bloqueLeido` gana `versiones []versionLeida`, en el orden de la respuesta del BOE, cada una con los mismos campos que hoy da la última —`fecha_publicacion` (la `fecha_version` de la salida), `fecha_vigencia`, `id_norma` y su texto normalizado igual que hoy— (`internal/source/boe/bloque.go`); lo que el lector da hoy para la última no cambia, ni la entrada de hoy, ni la salida sin fecha (FR-012). **Tests**: `TestLeerBloque` (`internal/source/boe/bloque_test.go`) gana las versiones de los tres bloques de research M1 sobre lo grabado de H4 —nueve en el `a22` de `BOE-A-1985-5392`, con sus fechas de publicación y de vigencia y sus normas escritas a mano en el test, y las del `a1-30` de `BOE-A-2017-12902` y del `a21` de `BOE-A-2015-10565`— y la lista vacía de una respuesta sin versiones construida en el test; lo que esperaba hoy, igual. **Verificación**: `go test -count=1 -run '^(TestLeerBloque|TestGolden|TestArticulo)$' ./internal/source/boe/` `ok`, con los trece golden de hoy sin cambiar; `make ci` en verde — FR-002, FR-010, FR-012. Rutas: internal/source/boe/bloque.go

- [ ] T004 [US1] `ValidarFecha` (plan.md paso 4; contracts/articulo-fecha.md §3, primera fila; research D3, M5). En `internal/source/boe/ids.go`, junto a `ValidarNorma` y `ValidarBloque`: `ValidarFecha(fecha string) error` acepta solo un día `AAAA-MM-DD` que existe en el calendario, con `time.Parse(time.DateOnly, …)` y nada más, y su fallo es `errorDeArgumentos` con el mensaje `la fecha "<valor>" no es un día AAAA-MM-DD del calendario, como 2019-06-01`. **Tests**: `TestValidarFecha` (nuevo, `internal/source/boe/ids_test.go`), las formas de research M5: válidas `2019-06-01`, `2016-02-29`, `1985-04-22` y `9999-12-31`; inválidas `2019-02-30`, `2015-02-29`, `2019-6-1`, `2019-06-1`, `20190601`, `2019/06/01`, `2019-06-01T00:00:00`, ` 2019-06-01`, la cadena vacía y `2019-13-01`, cada una con su clase (`cli.Clasificar`: `argumentos`), su código (`cli.CodigoSalida`: 2) y su mensaje. **Verificación**: con el mutante «la cadena vacía aceptada», `TestValidarFecha` falla (revertido); `make ci` en verde — FR-005. Rutas: internal/source/boe/ids.go

- [ ] T005 [US1] La lectura fechada en el adaptador del BOE: la consulta, la elección, sus dos mensajes de 3, la salida y su entrada de caché (plan.md paso 4; contracts/articulo-fecha.md §2, §3, §4 y §5; data-model §1 a §5 y §8; research D1, D2, D4 a D8). **Consulta** (`internal/source/boe/consultas.go`): `ConsultaArticulo` gana `Fecha *string`; `nil` es la lectura de hoy, por el mismo camino de código que hoy (FR-012). **Rama fechada** (`internal/source/boe/articulo.go`): con `Fecha`, tras `ValidarNorma` y `ValidarBloque` y antes de abrir la caché o pedir nada, `ValidarFecha` (orden norma, bloque, fecha: con las tres mal, el 2 de la norma); después, `fecha.go`. **`internal/source/boe/fecha.go`** (nuevo): la lectura fechada pide los mismos dos recursos que la de hoy, el bloque y los metadatos de su norma, con el `Pedidor` de la invocación (FR-006); la elección de research D4 —bloque sin versiones o con alguna sin `fecha_vigencia` AAAAMMDD: 3 con `el BOE no da la fecha de vigencia de cada redacción del bloque <bloque> de <norma>: no se puede saber cuál tenía el <AAAA-MM-DD>`, sea cual sea la fecha (FR-004); si no, la última, en el orden del bloque, cuya `fecha_vigencia` no es posterior a la pedida, comparadas como días del calendario y nunca por `fecha_publicacion`; ninguna: 3 con `el bloque <bloque> de <norma> no tenía redacción vigente el <AAAA-MM-DD>: su primera redacción tiene fecha de vigencia <AAAAMMDD>`; la redacción posterior, la última del bloque cuando no es la elegida—; los dos 3 con `errorDeNoEncontrado`, la dirección del bloque y el instante de la consulta, y ninguno con texto (research D7). **Salida** (`internal/source/boe/datos.go`): `ArticuloFechado`, `Articulo` embebido con los datos de la versión elegida como la nota [5] los da para la última, más `FechaPedida` (`fecha_pedida`, `omitempty`) y `RedaccionPosterior *RedaccionPosterior` (`redaccion_posterior`, `omitempty`), y `RedaccionPosterior{FechaVigencia, NormaModificadora}` con los nombres y la forma AAAAMMDD de los campos del artículo (research D5, D6); los avisos, los de la norma hoy, los mismos de la lectura sin fecha (FR-011). **Caché** (`internal/source/boe/entradas.go`): `versionesDelBloque` y `claveDeLasVersiones` (`boe.legislacion-consolidada|1|versiones|<dirección del bloque>`), con la vigencia de 7 días del artículo, el título, el tipo, todas las versiones, los avisos y el ELI; se escribe solo con el bloque y sus metadatos obtenidos e interpretados y antes de elegir; la lectura fechada se sirve solo de ella y nunca lee ni escribe la entrada de hoy (`articuloDelBloque`), que no cambia de forma, clave ni contenido; con `--offline` y sin ella, 4 con el verbo `versiones` y su clave (FR-020, FR-021, FR-022). **`internal/source/boe/doc.go`**: el párrafo del comentario de paquete sobre `articulo` con `--fecha`, que `refs/boe.py` no tiene (contracts/documentos.md, último apartado); la tabla de notas no cambia. Esta tarea no emite nada al grafo con la lectura fechada (T007). **Tests**: `TestArticuloEnUnaFecha` (nuevo, `internal/source/boe/fecha_test.go`), una tabla con los siete casos de contracts/articulo-fecha.md §2 sobre lo grabado de H4, cada uno con su código, su clase, `fecha_version`, `fecha_vigencia`, `norma_modificadora`, `hash_texto`, `fecha_pedida`, su redacción posterior o su ausencia y su mensaje escritos a mano, la huella de cada texto igual a la de su texto, y `titulo`, `tipo`, `avisos`, `url` y `url_eli` los de la lectura sin fecha; el del 2019-06-01 del `a22`, con el `texto` y el `hash_texto` del golden de hoy de ese bloque (FR-013); un bloque sin versiones y otro con una versión sin `fecha_vigencia`, construidos en el test, con fechas anteriores, intermedias y posteriores a las versiones con fecha: 3 con el mensaje de FR-004 y sin texto; y lo grabado sintético de avisos de H4 leído con `--fecha`, con los avisos de la lectura sin fecha (FR-011); `TestArticuloEnUnaFechaYLaCache` (nuevo, ídem): con solo la entrada de hoy, la fechada con `--offline` sale con 4; tras una fechada de una versión que no es la de hoy, la entrada de hoy igual byte a byte y la sin fecha con `--offline` igual; la fechada con `--offline` y su entrada, el mismo `data`, el mismo `hash` y la misma fecha de consulta que la que la guardó, también con otra fecha del mismo bloque (FR-020 a FR-022, FR-087; SC-007); `TestFallosNoSeGuardan` (`internal/source/boe/fuente_test.go`) gana los tres casos de plan.md («Por los fallos de la fuente con fecha») con `ConsultaArticulo{Norma: normaVigente, Bloque: bloqueDelArticulo21, Fecha}` el 2019-06-01 sobre lo sintético de H4 y sin fichero nuevo —`bloque-ilegible` (4), `metadatos-caidos` (4) y el 429 de los metadatos tras obtener el bloque (5)—, cada uno con la misma clase, el mismo código y el mismo mensaje que su consulta sin `Fecha` en un banco nuevo con el mismo `Pedidor`, el sobre sin datos con la dirección y el instante de la petición que falla, las mismas peticiones repetidas, nada guardado y, con `--offline`, el 4 de la entrada de las versiones ausente (`falloSinEntradaDeLasVersiones`); los cuatro casos de hoy, iguales (FR-007, FR-082); `TestClaveDeEntrada` y `TestEntradaIdaYVuelta` (`entradas_test.go`) ganan la clave y la ida y vuelta de la entrada de las versiones, y `TestDocAnotaElPorte` (`doc_test.go`) sigue en verde con el párrafo. **Verificación**: con cada mutante de FR-090 de la elección —la versión elegida por `fecha_publicacion`, siempre la última, y con una fecha anterior al bloque la primera versión—, `TestArticuloEnUnaFecha` falla (revertidos); quickstart.md §2 `ok`, con `TestGolden` y sus trece casos sin cambiar; `make ci` en verde — FR-001, FR-002, FR-003, FR-004, FR-006, FR-007, FR-010, FR-011, FR-012, FR-013, FR-020, FR-021, FR-022, FR-081, FR-082, FR-087; SC-004, SC-007. Rutas: internal/source/boe/consultas.go, internal/source/boe/articulo.go, internal/source/boe/fecha.go, internal/source/boe/datos.go, internal/source/boe/entradas.go, internal/source/boe/doc.go, internal/source/boe/fuente_test.go

- [ ] T006 [datos] [US1] La orden `--fecha` y el parámetro `fecha` de `boe_articulo`, con `schemas/bloque.json` regenerado y la tabla de `boe-legislacion` regenerada, en una sola tarea porque con la bandera registrada `schema-check` exige el esquema y `skills-check` la tabla (plan.md paso 5; contracts/articulo-fecha.md §1 y §4; research D2, D5, D6, D19, V11, V12, V14, V39). **Verbo** (`internal/app/boe.go`): `argumentosDeArticulo` gana `Fecha *string` con `placeholder:"AAAA-MM-DD"` y su ayuda, que llega a `ConsultaArticulo.Fecha`; la `Salida` declarada de `articulo` pasa a `boe.ArticuloFechado` (la de `articulos`, igual: FR-070); la herramienta `boe_articulo` gana `fecha` desde el registro, sin código del servidor (`internal/mcp` e `internal/app/herramientas.go` no cambian). **`schemas/bloque.json`**: regenerado con `go test -count=1 -run '^TestEsquemasPublicados$' ./internal/app/ -args -actualizar-esquemas`, nunca a mano: la entrada de `articulo` con `fecha` y `x-banderas: ["fecha"]`, su salida en `boe.ArticuloFechado` con las doce propiedades de hoy obligatorias y `fecha_pedida` y `redaccion_posterior` opcionales, y la definición `boe.RedaccionPosterior`; la parte de `articulos`, igual byte a byte. **`skills/boe-legislacion/SKILL.md`**: solo su región generada, con `make skills-sync` (la fila de `boe articulo` con `[--fecha <fecha>]` y los dos campos), sin editar a mano ninguna otra línea. **Tests**: `TestClavesDeLosDatos` (`internal/source/boe/datos_test.go`) gana `boe.ArticuloFechado` con sus doce claves obligatorias y las dos opcionales y `boe.RedaccionPosterior`, y el verbo `articulo` describe `boe.ArticuloFechado`, `boe.Aviso` y `boe.RedaccionPosterior`; `articulos`, igual; `TestSalidaDeBoeContraSchemas` (`internal/app/boe_test.go`) gana salidas fechadas —con redacción posterior, sin ella y el 3 de FR-003— validadas contra `schemas/bloque.json`, y lo que `boe_test.go` comprueba de los campos del verbo gana `Fecha`; `TestHerramientasDelServidor` no cambia lo que espera y compara `boe_articulo` con `fecha` con `--describe`; `TestEsquemasPublicados` y `TestSkillsDelRepositorio` en verde sin la bandera de actualizar. **Verificación**: con el mutante «`fecha_pedida` sin `omitempty` en `ArticuloFechado`», `TestEsquemasPublicados` o `TestSalidaDeBoeContraSchemas` fallan (revertido); `go run ./cmd/kitlegal boe articulo BOE-A-2015-10565 a21 --describe` muestra `"fecha"` y `boe.ArticuloFechado` sin abrir la caché (quickstart.md §5); quickstart.md §5 `ok`; `make ci` en verde, con `schema-check` y `skills-check` — FR-001, FR-005, FR-010, FR-012, FR-014, FR-048, FR-070, FR-084, FR-085; SC-011. Rutas: internal/app/boe.go, schemas/bloque.json, skills/boe-legislacion/SKILL.md, internal/source/boe/datos_test.go

**Checkpoint**: la lectura fechada funciona de punta a punta desde la terminal y como herramienta; la lectura sin fecha,
igual que en `main`.

---

## Fase 4: US4 — Una lectura fechada no engaña al grafo (P2)

**Objetivo**: la lectura fechada entrega solo la `BloqueVersion` elegida y su texto, con `SoloLoNuevo`, y no cuenta como
lectura del bloque.

**Prueba independiente**: quickstart.md §3, tercera orden; tras el bucle, `h20-fecha-grafo`.

- [ ] T007 [US4] Lo que la lectura fechada entrega al grafo, del adaptador al applet (plan.md pasos 4 y 5; contracts/articulo-fecha.md §6; data-model §6; research D9, D10, V15 a V19). **Adaptador** (`internal/source/boe/grafo.go`): `nodoDeLaVersion`, el nodo `BloqueVersion` con el id `<eli>#<bloque>@<fecha_vigencia>:<hash_texto>` y los datos con que lo emite hoy la lectura sin fecha (fecha de vigencia, fecha de versión, norma modificadora y huella), y `observadoDeLaLecturaFechada`: `schema.Observado{Vigencia: vigenciaLarga, SoloLoNuevo: true}` con ese nodo y su `Texto`, y nada más —ni la `Norma`, ni el `Bloque`, ni `eli:has_version`, ni otra arista—; sin ELI o bajo ensayo, nada, como hoy; la rama fechada con código 0 (`internal/source/boe/fecha.go` o `articulo.go`) lo pone en su `Resultado`; la lectura sin fecha emite lo de hoy. **Tests**: `TestObservadoDeBoe` (`internal/source/boe/grafo_test.go`) gana el caso de la lectura fechada —el nodo y el texto, `SoloLoNuevo`, nada sin ELI ni bajo ensayo— sin cambiar los de hoy; `TestLecturaFechadaEnElGrafo` (nuevo, `internal/app/grafo_test.go`), con el applet, el registro y el almacén reales en `t.TempDir()`: la `BloqueVersion` del 2005-06-15 del `a22` (`fecha_vigencia` `20040101`) entra con la procedencia del sobre como `source`; la fila de lecturas del bloque y `graph check` iguales antes y después, también con el nodo de la versión ya guardado por una lectura sin fecha y en un mundo sin fila de lecturas; la siguiente lectura sin fecha da el `graph check` que daría sin la lectura fechada; aplicada dos veces, nada la segunda; con `--no-graph` y con `--dry-run`, `world.db` igual; `TestLaSalidaDeBoeNoCambiaConElGrafo` (ídem) gana la lectura fechada. **Verificación**: con el mutante 4 de FR-090 —en `observadoDeLaLecturaFechada`, las operaciones de la lectura sin fecha, con `eli:has_version`—, `TestObservadoDeBoe` y `TestLecturaFechadaEnElGrafo` fallan (revertido); quickstart.md §3 `ok`; `make ci` en verde — FR-030, FR-031, FR-032, FR-086; SC-006. Rutas: internal/source/boe/grafo.go, internal/source/boe/fecha.go, internal/source/boe/articulo.go, internal/app/grafo_test.go

**Checkpoint**: `graph check` no ve las lecturas fechadas; el mundo gana sus versiones sin cambiar lo que sabía.

---

## Fase 5: US5 — El job sale en rojo si la respuesta no da la redacción fechada (P1)

**Objetivo**: el formato de eval admite la fecha y lo esperado de una redacción fechada; la caché preparada sirve
cualquier fecha; las evals 22 y 23 entran, y `redaccion_sin_fechar:<modelo>:<modo>` decide con 0.

**Prueba independiente**: quickstart.md §6.

- [ ] T008 [datos] [US5] El formato de eval gana `fecha` en el comando de bloque y `redaccion_fechada`, con su esquema, el juicio de qué casa con un comando de bloque y las tres formas fijas, en una sola tarea porque el lector estricto rechaza una clave que el tipo no tiene y el esquema valida toda eval (plan.md paso 6; contracts/evals-redaccion-fechada.md §1.1, §1.2 y §2; research D11, D12, D13). **Formato** (`internal/evals/formato.go`): `ComandoEsperado` gana `Fecha` (`fecha`, AAAA-MM-DD) en la forma bloque; `RedaccionFechadaEsperada` con `norma`, `bloque`, `fecha` y, o `fecha_vigencia` y, si la hay, `posterior` (`fecha_vigencia`, `norma_modificadora`), o `vigente_desde`; `Eval` gana `RedaccionFechada` (`redaccion_fechada`). **`schemas/eval.yaml.json`**: la definición nueva `comando-bloque-esperado` (con `fecha`) solo en `comandos`, y `grafo_previo` sigue con `comando-bloque`; `redaccion_fechada` con sus dos formas, prohibida con `activa: false` y en la eval sin binario ni servidor. **Juicio** (`internal/evals/juzgar.go`): los argumentos de una invocación de `articulo` se leen separando `--fecha` (con `=` o con su valor detrás, en cualquier sitio, también tras `--`) de los de posición; un comando sin fecha casa con `articulo` o `articulos` de la norma con el bloque, sin `--fecha` y con 0, como hoy; uno con fecha casa solo con `articulo` de la norma y el bloque con esa fecha y con 0 o 3; la herramienta `boe_articulo` con `fecha`, como su orden. **`internal/evals/fechadas.go`** (nuevo): las tres etiquetas `REDACCIÓN A FECHA`, `REDACCIÓN DE HOY` y `SIN REDACCIÓN VIGENTE` con la gramática de las de los avisos (`patronDeEtiqueta`), y `ExtraerRedaccionesFechadas(respuesta) []LineaFechada`: por cada línea con una forma y al menos una cita (`formaDeCita`), su forma, la norma y el bloque de su primera cita y lo leído detrás de ella —A y C, la primera fecha AAAA-MM-DD y la primera AAAAMMDD; B, el primer `BOE-A-…` y la primera AAAAMMDD—, con `fechasDeVigencia` para las AAAAMMDD; una línea sin cita no da nada. **Tests**: `TestExtraerRedaccionesFechadas` (nuevo, `internal/evals/fechadas_test.go`): cada forma, con mayúsculas y con énfasis de Markdown; una línea sin cita; dos líneas; lo leído detrás de la cita y no delante; la etiqueta de un aviso de hoy, que no da nada; `TestFormatoDeLasRedaccionesFechadas` (nuevo, `internal/evals/formato_test.go`): el esquema admite las dos formas de `redaccion_fechada` y `fecha` en `comandos`, y rechaza `fecha_vigencia` con `vigente_desde`, `posterior` sin `fecha_vigencia`, una fecha que no es AAAA-MM-DD, `fecha` en `grafo_previo` y `redaccion_fechada` con `activa: false` o en la eval sin binario ni servidor; `TestJuzgarLecturasFechadas` (nuevo, `internal/evals/juzgar_test.go`): la tabla de contracts §1.1 —comando con fecha y orden con `--fecha` delante y detrás, con `=` y separada, y con la herramienta; con 0 y con 3; sin fecha, con otra fecha, con `articulos` y con 4, no— y el caso 9 de FR-089: el comando sin fecha de la eval 02 no lo cumple una lectura fechada del `a1-30`, que deja el motivo `comando ausente: bloque boe BOE-A-2017-12902 a1-30`; los tests de juicio de hoy, iguales. **Verificación**: con el mutante «un comando sin fecha casa también con una lectura fechada», `TestJuzgarLecturasFechadas` falla (revertido); quickstart.md §6, primera orden, `ok`; `make ci` en verde, con las veintiuna evals de hoy válidas contra el esquema nuevo — FR-051, FR-052, FR-089; SC-010. Rutas: internal/evals/formato.go, internal/evals/juzgar.go, internal/evals/fechadas.go, schemas/eval.yaml.json

- [ ] T009 [US5] La caché preparada de las evals sirve la lectura fechada de cada bloque (plan.md paso 7; contracts/evals-redaccion-fechada.md §4; research D18, M4, M5, V31). `ConsultasNecesarias` (`internal/evals/consultas.go`) añade, por cada consulta `articulo <norma> <bloque>` que anota, la misma con `--fecha 9999-12-31`, y el comando fechado de una eval anota además su bloque sin fecha y su lectura con `--fecha 9999-12-31` (no la fecha de la eval, que en la (b) sale con 3); `Preparar` y `ComprobarSinRed` siguen tomando por falta todo lo que no sale con 0. **Tests**: `TestConsultasNecesarias` (`internal/evals/consultas_test.go`) gana la lectura con `--fecha 9999-12-31` de cada bloque y la de un comando fechado de una eval sintética; si `TestPreparar` o los tests de la medida (`internal/evals/preparar_test.go`, `internal/evals/medida_test.go`) fijan listas literales de consultas, ganan las nuevas sin perder ninguna; `TestEvalsDelRepositorio` prepara y comprueba sin red la caché de las veintiuna evals de hoy con las lecturas fechadas, sobre la unión de grabaciones; `medida.go` y la medida del juez no cambian (FR-058). **Verificación**: con el mutante «sin la lectura con `--fecha 9999-12-31`», `TestConsultasNecesarias` falla (revertido); `make ci` en verde — FR-053. Rutas: internal/evals/consultas.go, internal/evals/preparar_test.go, internal/evals/medida_test.go, internal/evals/conjunto_test.go

- [ ] T010 [US5] Las evals 22 y 23 y el umbral `redaccion_sin_fechar:<modelo>:<modo>`, en una sola tarea porque añadir las evals cambia a la vez el conjunto, la caché preparada, el peor caso del job y los umbrales, y los tests que los fijan cambian con ellas (plan.md paso 7, «Controles de umbral» y su recorrido; contracts/evals-redaccion-fechada.md §1.3, §1.4, §3 y §4; research D14, D15, D16, S1). **Evals**: `evals/boe-legislacion/22-lcsp-contrato-menor-en-2019.yaml` y `evals/boe-legislacion/23-lpac-articulo-21-en-2016.yaml`, tal cual contracts/evals-redaccion-fechada.md §1.3, positivas que activan la skill y deciden (ni `informativa` ni `por_materia`), con su comando de bloque con `fecha` (`a1-30` de `BOE-A-2017-12902` el 2019-06-01; `a21` de `BOE-A-2015-10565` el 2016-06-01), sus citas, sus avisos y su `redaccion_fechada`; las veintiuna de hoy no cambian. **Conjunto** (`internal/evals/conjunto.go`): las evals con `redaccion_fechada` se apartan como la eval sin binario ni servidor —ni `positivas` ni `citadaPor`; sí «normas conocidas» y el esquema—; `maximoDeEvals` de 21 a 23; `normasEsperadas` no lee `redaccion_fechada`. **Medida y umbral** (`internal/evals/informe.go`, `internal/evals/umbrales.go`): `medirLasRedaccionesFechadas`, con `respuestaDelJuez.conFecha` y `.sinFechar` y `respuestasMedidas.conRedaccionFechada`; `umbralDeRedaccionSinFechar` con el nombre, la descripción, el `total`, la `medida`, `"<="`, 0 y `decide: true` de contracts §3, en su sitio tras `cuenta_su_proceso` de su modo, y su caso en `nombradasPorSuSesion`, con el motivo `umbral redaccion_sin_fechar:<modelo>:<modo>: <n> de <total> (<p> %), y tiene que ser ≤ 0,0 %: <sesiones>`; fuera del `total` y de la `medida`, las sesiones sin medir (que siguen con su motivo de la ejecución), las de un modelo informativo y las de la eval sin binario ni servidor (las de `respuestasQueSeJuzgan`); existe si alguna eval de la skill lleva `redaccion_fechada`, sin ningún caso por skill (FR-055). **`internal/evals/doc.go`**: lo que el comentario de paquete dice de los umbrales de cada skill y deja de ser cierto. **Tests**: `TestUmbralesDeBoeLegislacion` (nuevo, `internal/evals/informe_test.go`), con sesiones sintéticas del plan del job sobre las evals del repositorio (`escribirSesionSinteticaDelPlan`, `informeDelPlanDelJob`, y una lectura fechada como orden y como herramienta junto a `leerElArticulo`): los casos 1 a 8 de contracts §4 —`lee-sin-la-fecha` y `lee-otra-fecha`; `lee-con-la-fecha` en los dos modos; `sin-la-linea` y `con-otra-fecha`; `sin-la-de-hoy`, `la-de-hoy-con-otra-fecha` y `la-de-hoy-con-otra-norma`; `sin-redaccion-vigente` y `redaccion-a-fecha-en-la-b`; `una-eval-sin-fecha`; `sin-medir`; `de-haiku`—, cada uno con los quince umbrales, sus medidas, el veredicto y los motivos escritos a mano; `TestUmbralesDeLasTresSkills` (ídem) con los quince de `boe-legislacion` en el orden de contracts §3, once que deciden (un ayudante `umbralesDeBoeLegislacion` que añade los dos a `umbralesDeBoeLegislacionComoEnH26`), `afirma_lo_no_leido:<modelo>` sobre 120 con su umbral de 0,028 de `clases.yaml`, `sin_activar` sobre 60 por modo, y `legal-core` (seis, seis) y `jurisprudencia` (quince, once) como hoy; `TestConjuntoDeEvals` (`internal/evals/conjunto_test.go`), con el máximo de 23 y una eval con fecha que no es positiva; `TestEvalsDelRepositorio`, con las veintitrés evals válidas, sus reglas y su caché preparada y comprobada sin red; `TestConsultasNecesarias` (`internal/evals/consultas_test.go`) si enumera las consultas de las evals del repositorio; `TestDefinicionDelJob` (`internal/evals/definicion_test.go`), solo el literal `peorCasoDeBoeLegislacion` —18 977 s, con 109, 108 y 6 sesiones y 60, 60 y 3 respuestas— y el caso de `probarElPeorCasoDelTrabajo` que lo lleva, recalculados desde `definicion.go` con las evals del repositorio, que no cambia; la concurrencia (5), el objetivo de 900 s y el tope de 352 minutos, iguales (FR-059); `TestPlanEnDosModos` (`internal/evals/plan_test.go`, subtest `del-repositorio`), solo las cifras del caso `boe-legislacion` de `probarElPlanDelRepositorio` y su comentario —series 36, 36 y 2 por modo (orden, herramienta y sin modo) y sesiones 108, 108 y 6, hoy 32, 32 y 2 y 96, 96 y 6—, recalculadas con el `PlanDeEvals` de las evals del repositorio y los modelos y repeticiones de la definición del job, que no cambian; el caso de `legal-core`, la sesión de más con la prueba de red y la lógica del test, iguales (FR-059). **Verificación**: con el mutante 5 de FR-090 —el `decide` de `umbralDeRedaccionSinFechar` en falso—, `TestUmbralesDeBoeLegislacion` y `TestUmbralesDeLasTresSkills` fallan (revertido); quickstart.md §6 `ok`; `make ci` en verde — FR-047, FR-050, FR-051, FR-052, FR-053, FR-054, FR-055, FR-056, FR-057, FR-058, FR-059, FR-071, FR-072, FR-088, FR-089; SC-001, SC-009, SC-010. Rutas: evals/boe-legislacion/22-lcsp-contrato-menor-en-2019.yaml, evals/boe-legislacion/23-lpac-articulo-21-en-2016.yaml, internal/evals/conjunto.go, internal/evals/informe.go, internal/evals/umbrales.go, internal/evals/doc.go, internal/evals/consultas_test.go, internal/evals/definicion_test.go, internal/evals/plan_test.go

**Checkpoint**: el job de `boe-legislacion` decide sobre la redacción fechada; las evals con fecha existen antes que la
skill que las responde (Definition of Done §1.10).

---

## Fase 6: US1, US2 y US3 — `boe-legislacion` v0.2 (P1)

**Objetivo**: con una pregunta con fecha, la skill lee el bloque en esa fecha, cita esa redacción y da sus líneas de
forma fija; sin fecha, como hoy.

**Prueba independiente**: `make skills-check` y `TestEvalsDelRepositorio/formas-de-la-redaccion-fechada`; en el cierre,
las evals 22 y 23 y el umbral de T010.

- [ ] T011 [US1] `boe-legislacion` v0.2 (plan.md paso 8; contracts/skill-boe-legislacion.md §1 a §4; contracts/SKILL-boe-legislacion.prototipo.md; research D13, D17). `skills/boe-legislacion/SKILL.md`, desde el prototipo y su contrato: la `description` gana las preguntas con fecha y sigue ganando las de hoy, en 1 024 caracteres como mucho (FR-047); la sección «Redacción en una fecha» —leer con `--fecha` o con `boe_articulo` y `fecha` cuando la pregunta lleva un día, un año o un mes, o un hecho cuya fecha dan la pregunta o la conversación, y pedir el día ante un hecho sin fecha (FR-040); con un año o un mes, dos lecturas, el primer y el último día, y si dan resultados distintos ninguna cita y se pide el día (FR-045); la línea A tras un 0, la B si hay redacción de hoy, sin decir que rija el día de la consulta ni hasta cuándo rigió la citada (FR-041); de la redacción de hoy, nada salvo leerla sin fecha y citarla (FR-042); la línea C con el 3 de FR-003, sin citar el texto como el de entonces ni decir qué regía sin leerlo, y el 3 de FR-004 en la regla 2 (FR-043); los avisos como hoy (FR-044)—; las tres formas fijas tal cual, cada una en su bloque `text` (contracts/skill-boe-legislacion.md §1); el paso 3, «Redacción modificada» y la regla 7, solo para la lectura sin fecha, sin `kitlegal graph check` tras una lectura fechada (FR-046; research D17); y lo que deja de ser cierto con la lectura fechada, corregido para la lectura sin fecha y solo eso (FR-049); la prosa no nombra los campos de la salida (`expresiones-prohibidas.yaml`); por debajo de 300 líneas, sin perder ninguna regla, con la región generada que dejó T006 sin editar a mano y `references/` sin cambios (FR-048). **Tests**: `TestEvalsDelRepositorio` (`internal/evals/conjunto_test.go`) gana el subtest `formas-de-la-redaccion-fechada`: el `SKILL.md` de `boe-legislacion` lleva, en sus bloques `text`, las tres etiquetas de `fechadas.go`, como `hallazgos-de-la-skill` con la suya; el test de las formas PowerShell (`conjunto_test.go`) sigue en verde. **Verificación**: con el mutante «una etiqueta escrita de otra manera en el bloque `text`», `formas-de-la-redaccion-fechada` falla (revertido); `make skills-check` `ok` y `wc -l` de `SKILL.md` < 300; `go test -count=1 -run '^TestEvalsDelRepositorio$' ./internal/evals/` `ok`; `make ci` en verde — FR-040, FR-041, FR-042, FR-043, FR-044, FR-045, FR-046, FR-047, FR-048, FR-049; SC-001, SC-011. Rutas: skills/boe-legislacion/SKILL.md, internal/evals/conjunto_test.go

**Checkpoint**: la skill responde a la pregunta con fecha con la redacción de esa fecha; lo mide el job de cierre.

---

## Fase 7: Pulido — mutantes, documentos y cierre

- [ ] T012 Los cinco mutantes de FR-090, aplicados uno cada vez en un árbol aparte fuera del repositorio, y su registro (plan.md paso 9 y fila de FR-090 de «Controles de umbral»; quickstart.md §7; contracts/evals-redaccion-fechada.md §5). Con `git worktree add --detach "$TMPDIR/h20-mutantes" HEAD` y `go -C "$TMPDIR/h20-mutantes"`, sin tocar el árbol de trabajo: (1) la versión elegida por `fecha_publicacion`, (2) siempre la última y (3) con una fecha anterior al bloque la primera versión, en la elección de `fecha.go`, con `TestArticuloEnUnaFecha`; (4) la lectura fechada contada como lectura —las operaciones de la lectura sin fecha, con `eli:has_version`— en `observadoDeLaLecturaFechada`, con `TestObservadoDeBoe` y `TestLecturaFechadaEnElGrafo`; (5) el `decide` de `umbralDeRedaccionSinFechar` en falso, con `TestUmbralesDeBoeLegislacion` y `TestUmbralesDeLasTresSkills`; con cada uno, por separado, se ejecutan las órdenes de quickstart.md §7 y al menos un test falla; sin mutantes, todos pasan. **`specs/024-h20-boe-articulo-fecha/mutantes.md`**: por mutante, el cambio en una línea, la orden ejecutada y cada test que falla, con su subtest o su caso y su mensaje, tal como se vio en esta sesión. Esta tarea no cambia nada bajo `internal/`; si un mutante no hace fallar ningún test, lo anota en `mutantes.md` con su salida y la tarea no se da por terminada, porque el refuerzo es del test de la tarea que construyó el control. El árbol aparte se retira con `git worktree remove --force "$TMPDIR/h20-mutantes"`, y `git worktree list` ya no lo nombra. **Verificación**: los cinco mutantes con al menos un test en rojo en `mutantes.md`; `make ci` en verde; `git status` sin nada fuera de `mutantes.md` — FR-090; SC-008. Rutas: specs/024-h20-boe-articulo-fecha/mutantes.md

- [ ] T013 Lo que los documentos dicen de `boe articulo`, de `boe-legislacion`, de sus evals, de su formato y de sus umbrales, con cada cifra, fichero, test y umbral sacados del árbol y no de memoria (plan.md paso 10; contracts/documentos.md; research D21). **`CHANGELOG.md`**, dentro de `## [Unreleased]`, una sección `### Añadido` delante de `### Cambiado`, con los cuatro puntos de contracts/documentos.md: `--fecha` en `boe articulo` y `fecha` en `boe_articulo`, con la regla de elección, el 3 de una fecha anterior al bloque y la redacción posterior; que una lectura fechada no cuenta como lectura del bloque en `graph check`; `boe-legislacion` v0.2 con sus tres líneas; y las evals 22 y 23, el formato que las admite y el umbral `redaccion_sin_fechar:<modelo>:<modo>` (qué cuenta, sobre qué sesiones, con 0 y que decide); ninguna sección de una versión publicada cambia (FR-060, FR-091). **`README.md`**: solo la frase de «Qué sabe hacer tu agente con kitlegal» de contracts/documentos.md. **`CONTRIBUTING.md`**: los pasajes de la tabla de contracts/documentos.md —el formato de eval (`comandos` con `fecha`, `redaccion_fechada`, `sin_binario_ni_servidor`), las reglas del conjunto de `boe-legislacion`, el juicio de un comando de bloque, las respuestas que se juzgan (60 por modo), el trabajo del job (veintitrés evals, 222 sesiones, catorce que deciden, 123 respuestas), el peor caso de `boe-legislacion` (18 977 s, el literal de `TestDefinicionDelJob`), cuándo existen los umbrales de las respuestas, los quince umbrales de `boe-legislacion` con once que deciden y la entrega al grafo de `boe articulo --fecha`—, cada cifra comprobada contra el test o el fichero que la fija. **`docs/WORKFLOW.md`**: solo la frase de `medir_cierre` que enumera los umbrales que deciden en `boe-legislacion` (de nueve a once, con `redaccion_sin_fechar:<modelo>:<modo>` en cada modo), sin reescribir el resto. Solo lo falso o incompleto; nada de lo que sigue siendo cierto se reescribe, ninguna release se promete y ningún test ata este texto al código (FR-062). **Verificación**: cada test, fichero, umbral y cifra nombrados existen en el árbol (`git grep`); `git diff` de `docs/WORKFLOW.md` con una sola frase cambiada; `make ci` en verde — FR-060, FR-061, FR-062, FR-063, FR-091; SC-012. Rutas: CHANGELOG.md, README.md, CONTRIBUTING.md, docs/WORKFLOW.md

- [ ] T014 Cierre de la Definition of Done, sin código de producto (`docs/ROADMAP.md` §1; quickstart.md §0 a §8; plan.md, «Obligaciones para tasks.md»): `specs/024-h20-boe-articulo-fecha/cierre.md` con (1) la lista, sacada de `git diff --name-status main`, de lo creado o modificado, y la comprobación de que no cambia nada de «Lo que no cambia» de la batería —`skills/legal-core/`, `skills/jurisprudencia/`, `evals/legal-core/`, `evals/jurisprudencia/`, `evals/boe-legislacion/juez/`, `evals/boe-legislacion/expresiones-prohibidas.yaml`, las veintiuna evals de hoy, `.github/`, `scripts/`, `.specify/`, la carpeta de la evidencia, `docs/SOURCES.md`, `docs/ROADMAP.md`, `docs/USO.md`, `docs/HISTORIAL.md`, `docs/ADR/`, `web/`, `CLAUDE.md`, el `Makefile`, `internal/mcp`, lo grabado y sus manifiestos, los guiones e2e de hoy y los specs de otros hitos— (FR-006, FR-059, FR-063, FR-070), y de que ninguna petición nueva aparece en lo que pide la lectura fechada (FR-006); (2) la cobertura sobre `coverage.out` y `coverage-integration.out` de `make ci` con `go tool cover -func` —global ≥ 70 % e `internal/core/**` ≥ 85 % (Definition of Done §1.9)— con la cifra de cada una y la de `internal/source/boe`, `internal/graph` e `internal/evals`, y, si alguna queda por debajo, el hueco (paquete, fichero y funciones sin cubrir) para la revisión final; esta tarea no escribe tests; (3) el resultado de quickstart.md §0 a §8, orden a orden, con lo esperado de cada una —los guiones `h20-` de §4, que el workflow activa tras el bucle, ejercidos con copias momentáneas de prefijo `zz-` en el directorio de guiones del arnés de `internal/app`, que se borran—, y la constancia de que §9 lo hace el workflow y §10 queda para después del run; (4) la tabla de plan.md, «Controles de umbral», con el test de cada fila `ci:` encontrado como `^func <Test>\(` en su ruta y verde, y cada fila `evals:` con el umbral que publica T010 o el de hoy que sigue; y (5) la nota de FR-092 que lleva el informe final (la fase 1 sigue con H9.1, ADR 0040; H8 y H11 pueden dar por hecha la redacción de un bloque a una fecha), que no es trabajo de ninguna tarea. **Verificación**: `make ci` en verde y `git status` sin nada fuera de `cierre.md` — FR-006, FR-059, FR-063, FR-070, FR-080, FR-092; SC-002, SC-004 a SC-011. Rutas: specs/024-h20-boe-articulo-fecha/cierre.md

---

## Dependencias y orden de ejecución

El bucle ejecuta las tareas en su orden; ninguna depende de una posterior.

| Tarea | Depende de | Por qué |
|---|---|---|
| T001 | — | la suite se escribe antes que el código, con la salida de `main` de hoy |
| T002 | — | solo los contratos del grafo y el almacén |
| T003 | — | solo el lector del bloque |
| T004 | — | solo la validación de la fecha |
| T005 | T003, T004 | elige entre las versiones del lector y valida con `ValidarFecha` |
| T006 | T005 | la bandera llega a `ConsultaArticulo.Fecha`; la `Salida` es `ArticuloFechado` |
| T007 | T002, T005, T006 | `SoloLoNuevo`; la rama fechada; el applet con la bandera para `TestLecturaFechadaEnElGrafo` |
| T008 | T006 | la herramienta con `fecha` y la línea de su orden, que el juicio lee |
| T009 | T006, T008 | el binario prepara con `--fecha 9999-12-31`; `ComandoEsperado.Fecha` |
| T010 | T008, T009 | el formato, las formas fijas y la caché preparada de las evals nuevas |
| T011 | T006, T008, T010 | la tabla regenerada; las etiquetas de `fechadas.go`; las evals antes que la skill |
| T012 | T005, T007, T010 | los tests que deben fallar con cada mutante |
| T013 | T005 a T011 | lo que documenta ya está en el árbol, con sus cifras |
| T014 | todas | el cierre sobre el diff entero |

### Historias

- **US1** (P1): T003 a T006 y T011. **US2** (P1): con US1 (el 3 de FR-003 en T005; la línea C en T011). **US3** (P1):
  la lectura sin fecha intacta en T003, T005 y T006, comprobada por `fecha-main` (T001) y `TestGolden`. **US4** (P2):
  T002 y T007. **US5** (P1): T008 a T010.

### Oportunidades de paralelismo

Ninguna tarea lleva `[P]`: el workflow las ejecuta en secuencia, con un guardián por tarea. Fuera del bucle, T002, T003
y T004 son independientes entre sí.

---

## Trazabilidad

### Requisitos → tareas

| Requisito | Tareas |
|---|---|
| FR-001 | T001, T005, T006 |
| FR-002 | T001, T003, T005 |
| FR-003 | T001, T005 |
| FR-004 | T005 |
| FR-005 | T001, T004, T005, T006 |
| FR-006 | T001 (con la reproducción), T005, T014 |
| FR-007 | T001, T005 |
| FR-010 | T001, T003, T005, T006 |
| FR-011 | T005 |
| FR-012 | T001, T003, T005, T006 (y la batería) |
| FR-013 | T001, T005 |
| FR-014 | T006 |
| FR-020 a FR-022 | T001, T005 |
| FR-030 a FR-032 | T001, T002, T007 |
| FR-040 a FR-046 | T011 |
| FR-047 | T010 (las evals con fecha en `sin_activar`), T011 |
| FR-048 | T006 (la tabla), T011 |
| FR-049 | T011 |
| FR-050 | T010 |
| FR-051 | T008, T010 |
| FR-052 | T008, T010 |
| FR-053 | T009, T010 |
| FR-054 a FR-056 | T010 |
| FR-057, FR-058 | T010 |
| FR-059 | T010, T014 |
| FR-060 a FR-062 | T013 |
| FR-063 | T013, T014 |
| FR-070 | T001, T006, T014 (y la batería) |
| FR-071, FR-072 | T010 (y la batería) |
| FR-080 | todas; T014 |
| FR-081 | T005 |
| FR-082 | T001, T005 |
| FR-083 | T001 |
| FR-084 | T001, T006 |
| FR-085 | T006 |
| FR-086 | T002, T007 |
| FR-087 | T005 |
| FR-088 | plan.md, «Controles de umbral»; T010 publica el umbral |
| FR-089 | T008 (caso 9), T010 |
| FR-090 | T012 (y cada tarea ve su control en rojo: T005, T007, T010) |
| FR-091 | T013 (`CHANGELOG.md`); el job, el workflow tras la revisión final |
| FR-092 | el informe final del workflow; T014 la constata |
| SC-001 | T010, T011; lo mide el job de cierre |
| SC-002 | el informe final, con la fila de plan.md; T014 la constata |
| SC-003 | después del run, fuera de él; ninguna tarea |
| SC-004 | T001, T005 |
| SC-005 | T001 |
| SC-006 | T002, T007 |
| SC-007 | T005 |
| SC-008 | T012 |
| SC-009, SC-010 | T010 (T008, el caso 9) |
| SC-011 | T006, T011 |
| SC-012 | T013; la revisión final |

### Controles de umbral (plan.md) → tarea que construye el control

| Fila de plan.md | Tarea | Cómo se ve fallar |
|---|---|---|
| `redaccion_sin_fechar` (orden, herramienta) | T010 (formato y formas de T008) | `TestUmbralesDeBoeLegislacion`, casos 1, 3, 4 y 5; mutante 5 de T012 |
| `sin_activar` de `boe-legislacion` (orden, herramienta) | el de hoy; T010 añade las evals con fecha a su `total` | `TestUmbralesDeLasTresSkills` con otro `decide` u otro orden |
| `afirma_lo_no_leido:<modelo>`, la medida del juez, las duraciones | los de hoy; T010 las fija entre los quince | `TestUmbralesDeLasTresSkills` con otro elemento |
| `jurisprudencia` y `legal-core` siguen como decidían | ninguna los cambia; T010 los fija | `TestUmbralesDeLasTresSkills` con otro elemento |
| `TestArticuloEnUnaFecha` y `h20-fecha-version` (FR-002, FR-003, FR-013, FR-081) | T005, T001 | mutantes 1 a 3 de T012 |
| `TestArticuloEnUnaFecha` (FR-004, FR-011) | T005 | un bloque con una versión sin `fecha_vigencia` situado por su publicación |
| `TestValidarFecha`, `TestFallosNoSeGuardan`, `h20-fecha-codigos` (FR-005, FR-007, FR-082) | T004, T005, T001 | la cadena vacía aceptada; la entrada de las versiones escrita con un fallo |
| `h20-fecha-main` y `TestGolden` (FR-012, FR-083) | T001, T005 | cualquier byte de la lectura sin fecha cambiado |
| `TestArticuloEnUnaFechaYLaCache` y `h20-fecha-offline` (FR-020 a FR-022, FR-087) | T005, T001 | la lectura fechada servida de la entrada de hoy |
| `TestApplySoloLoNuevo`, `TestIntegracionLecturaFechada`, `TestLecturaFechadaEnElGrafo`, `h20-fecha-grafo` (FR-030 a FR-032, FR-086) | T002, T007, T001 | `SoloLoNuevo` ignorado; mutante 4 de T012 |
| `h20-fecha-herramienta` y `TestHerramientasDelServidor` (FR-001, FR-084) | T001, T006 | la herramienta sin `fecha` |
| `schema-check` y `TestSalidaDeBoeContraSchemas` (FR-014, FR-085) | T006 | `schemas/bloque.json` sin regenerar; un campo opcional como obligatorio |
| `skills-check` y `TestEvalsDelRepositorio` (FR-048) | T006, T011 | 300 líneas o más; una tabla editada a mano; una etiqueta cambiada |
| `TestDefinicionDelJob` (FR-050) | T010 | el literal sin recalcular con las veintitrés evals |
| `TestPlanEnDosModos` (FR-050, FR-059) | T010 | las series y sesiones de `boe-legislacion` sin recalcular con las veintitrés evals |
| `TestFormatoDeLasRedaccionesFechadas`, `TestJuzgarLecturasFechadas`, `TestEvalsDelRepositorio` (FR-051, FR-052) | T008, T010 | un comando sin fecha que casa con una lectura fechada |
| `TestEvalsDelRepositorio` (FR-053) | T009, T010 | sin la lectura con `--fecha 9999-12-31` |
| `TestUmbralesDeBoeLegislacion`, `TestJuzgarLecturasFechadas`, `TestUmbralesDeLasTresSkills` (FR-055, FR-056, FR-089) | T010, T008 | sus casos; mutante 5 |
| FR-090 | T012 | los cinco mutantes |
| `make ci` (FR-080) | todas | — |

### Definition of Done (`docs/ROADMAP.md` §1) → tareas

| Punto | Aplica | Tareas |
|---|---|---|
| 1 · `make ci` en verde | Sí | Todas; T014 |
| 2 · Tests offline; fixtures si toca red | Sí, sin fixtures nuevos: lo grabado de H4 sirve los tres bloques con todas sus versiones (research D22) | T003, T005 (leen lo grabado); T001 (con la reproducción) |
| 3 · Sin `net/http`, `os.Exit`, `fmt.Print*` fuera de lo autorizado | Sí | T005 (pide con el `Pedidor` de hoy); T006 (la salida legible del kernel); ningún import nuevo (R1 a R6) |
| 4 · Esquemas y `schema-check` | Sí | T006 (`schemas/bloque.json`); T008 (`schemas/eval.yaml.json`) |
| 5 · Errores tipados y códigos estables | Sí | T004 (2), T005 (los dos 3 nuevos, 4 y 5 como sin fecha) |
| 6 · e2e y `CHANGELOG.md` | Sí | T001; T013 |
| 7 · ADR | No: la lectura fechada y su entrega caben en los ADR 0005, 0014 y 0023; el umbral, en los 0029 y 0038 (plan.md, «Summary») | — |
| 8 · `SOURCES.md` y `verify-sources.sh` | No: la misma fuente y los mismos recursos, con fila revisada en `main` (research V37) | T014 lo constata |
| 9 · Cobertura | Sí | cada tarea con sus tests; T014 (las cifras y cualquier hueco, en `cierre.md`) |
| 10 · Skill: evals antes, `SKILL.md` < 300, sin drift | Sí | T010 (evals, antes que la skill); T006 y T011 |
| 11 · Dimensión territorial | No: la lectura fechada no depende de un municipio (plan.md, principio IX) | — |
| 12 · Operaciones de grafo | Sí | T002, T007 |
| 13 · Recursos y plazos: común y regulado aparte | No: el hito no calcula plazos ni genera escritos | — |

---

## Estrategia de implementación

- **De dentro afuera** (plan.md, «Orden de implementación»): la suite; el grafo que acepta lo observado sin pisarlo; el
  lector con todas las versiones; la fecha; la lectura fechada en el adaptador; la orden y la herramienta; la entrega
  al grafo; el formato de eval; la caché preparada; las evals y el umbral; la skill; los mutantes; los documentos; y el
  cierre.
- **Lo mínimo que ya da valor** es la fase 3 (T001 a T006): `kitlegal boe articulo BOE-A-2017-12902 a1-30 --fecha
  2019-06-01` da la redacción vigente desde el 20180309 y avisa de la de hoy, desde la terminal y como herramienta. No
  es una entrega parcial: el run ejecuta las catorce, y SC-001 pide el cierre entero, con la skill de T011 cumpliendo el
  umbral de T010.
- **Lo que el run no puede medir** (research S2): qué dará el job de cierre con `boe-legislacion` v0.2. Si el cierre
  sale en rojo por un umbral de `boe-legislacion`, se repara lo que entrega el hito —la skill—, sin cambiar el umbral,
  su `decide`, su `total`, las evals, la rúbrica, los casos ni la medida del juez (FR-071); si sale en rojo por
  `jurisprudencia` o `legal-core`, queda en el informe final.
- **Si lo grabado contradice a un contrato** (research M1 a M4): manda lo grabado; T003, T005 y T010 corrigen el
  artefacto del feature que lo recordaba y lo anotan en `gates/supuestos.md`; los guiones congelados de T001 se validan
  antes contra el prototipo.
- **Después del run** (SC-003; quickstart.md §10): con la release publicada, la prueba en la app de escritorio.

---

## Comprobación contra la rúbrica de `juez_tasks` y `precheck.sh tasks`

- `precheck.sh tasks`: sus comprobaciones de la fase tasks se reproducen con `grep` y sus mismos patrones sobre las
  líneas de tarea (resultado en la sesión que escribe este fichero): catorce líneas con el formato `- [ ] Tnnn`, ids
  únicos y correlativos (T001 a T014); ninguna sin rutas; las que nombran el directorio de esquemas llevan `[datos]`
  (T006 y T008, y ninguna otra línea lleva esa etiqueta) y ninguna nombra un directorio de fixtures; ninguna nombra la
  carpeta de la evidencia por su ruta, ni la variable de grabación, ni una acción de plataforma, ni pide nada a una
  persona a mitad; una sola `[aceptacion]`, la primera; `checklists/requirements.md` sin ítems sin marcar. El único
  defecto que puede dar antes del análisis es que falta `gates/analyze.md`, que escribe el paso siguiente.
- a · `analisis_critico`: lo escribe el paso de análisis después de este fichero.
- b · `trazabilidad`: cada tarea cita sus FR y SC; «Requisitos → tareas» cubre FR-001 a FR-092 y SC-001 a SC-012; y
  «Controles de umbral» da la tarea y cómo se ve fallar cada fila de plan.md.
- c · `rebanadas_verdes`: T002 a T011 traen sus tests y su implementación; las excepciones (T001, T012 a T014) están
  declaradas arriba; el orden es el del plan, con los pasos 4 y 7 partidos, y la tabla de dependencias no tiene ninguna
  hacia delante.
- d · `rutas_declaradas`: tras «Rutas:», ficheros concretos; el único directorio es el de la suite (T001), que nace con
  la tarea; ninguna ruta genérica.
- e · `datos_separados`: las tareas que tocan el directorio de esquemas llevan `[datos]`, y su código va con ellas por
  la razón declarada arriba; ningún dato externo: lo grabado de H4 ya sirve la lectura fechada (research D22), nada sale
  de la red, de una persona ni de la memoria.
- f · `dod`: tabla «Definition of Done → tareas»; `CHANGELOG.md` en T013; sin ADR ni `SOURCES.md`, con su motivo;
  esquemas en T006 y T008; sin fixtures nuevos, con su motivo.
- g · `checklist_veraz`: `checklists/requirements.md` no cambia.
- h · `aceptacion_primero`: T001 es la tarea `[aceptacion]`; sus guiones cubren US1.1, US2.1, US3.1 y US4.1; lo que no
  tiene guion (US5, US1.2, US2.2, US3.2) se acepta como dice la cabecera, como fija el plan.
- i · `autonomia`: ninguna tarea exige a una persona, publica, ni mide en la plataforma.
