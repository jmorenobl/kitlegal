# Tasks: H9 · `internal/core/plazos` + applet `plazos` + skill `legal-core` v1

**Input**: `specs/023-h9-internal-core-plazos/` (spec.md, plan.md, research.md, data-model.md, contracts/,
quickstart.md)

**Prerrequisitos**: spec.md y plan.md leídos; `.specify/memory/constitution.md` 2.13.0 aplicada (umbral de
materialidad, criterio de uso, proporcionalidad y convergencia en «Gates»; controles de umbral del ADR 0029; ADR 0018,
0023, 0027, 0032, 0038 y 0039); H23.1 y #157 en `main` (`legal-core` a dos sesiones, `umbrales` de `legal-core`
vacío, trece umbrales en `boe-legislacion` y quince en `jurisprudencia`). **Ninguna tarea crea el esqueleto ni los
gates: `make ci` existe y pasa**, y cada tarea lo deja en verde al terminar.

**Aceptación**: cinco guiones `testscript` (plan.md, «Aceptación e2e»), que escribe T001 en
`specs/023-h9-internal-core-plazos/aceptacion/` y que el workflow congela y, tras el bucle, activa como
`h9-plazos-<nombre>.txtar` en el directorio de guiones del arnés e2e de `internal/app`. Usan solo los casos de
contracts/casos-de-plazos.md cuyo día decisivo es cierto antes de grabar (1, 2, 3, 15, 17 y 18; research D27). Lo que
no tiene guion, y por qué: **US3** (los festivos salen del boletín) se acepta con los tests de derivación y de cotejo de
T007 y T008; **US4** y **US6**, y US1.2, con las evals que entran en T011 y el job de cierre, que lanza el workflow tras
la revisión final (SC-001), y en `make ci` con los tests de sesiones sintéticas de T011 y T012; **US1.3** (la skill
pregunta el municipio que falta y nunca usa el de quien usa el kit) y FR-043, FR-044, los leen los jueces de la
revisión final, como fija el spec (FR-040, «Comprobable»), porque FR-050 no admite más evals que las cinco.

**Tests**: obligatorios (constitución §III; spec FR-080 a FR-089; plan.md, «Controles mecánicos»). Cada tarea trae
sus tests y la implementación que los hace pasar en el mismo diff; los nombres de test y sus ficheros son los de
plan.md, «Tests nuevos», porque el informe final busca cada control `ci:<ruta>:<Test>` como `^func <Test>\(` en esa
ruta.

**Rebanadas verticales y excepciones declaradas.** Son rebanada completa —test e implementación en el mismo diff,
`make ci` en verde por sí solas— T002 a T013. Las demás, cada una por su razón:

- **T001** es la suite congelada: sin código de producto; el «rojo primero» del workflow la verifica.
- **T014** es la prueba de FR-089 sobre el código y los tests que dejan T004 a T012: aplica los mutantes en una copia
  fuera del árbol y deja su registro; no cambia ninguna línea de código de producto, y solo toca un test si un mutante
  no hace fallar ninguno: entonces refuerza ese test, en su fichero declarado entre sus rutas, porque las rutas se
  congelan al empezar cada intento y una tarea ya marcada no se reabre.
- **T015** es documentación; su verificación es que cada cifra, fichero, test y umbral que nombra existe en el árbol, y
  `make ci`.
- **T016** es el cierre de la Definition of Done, sin código de producto.

**Tareas `[datos]` con código** (plan.md, *Complexity Tracking*, última fila; precedente H7.2 a H23.1): T005, T007,
T008, T009 y T011 tocan el directorio de esquemas o uno de fixtures y además código, porque `make ci` no queda en verde
si entran por separado: el lector de un fichero de `data/` y su esquema (T005), el analizador de un boletín y el
fichero que deriva (T007, T008), el applet con su esquema, sus golden y los guiones que enumeran applets y herramientas
(T009), y el formato de eval con su esquema y las evals que lo usan (T011). T002 solo deja manifiestos y sus tests de
grabación, y no deriva nada. Ningún fichero de `data/festivos/` se escribe a mano: lo escribe el test de derivación con
`-actualizar-festivos` desde lo grabado. `data/plazos.yaml` y `data/cobertura.yaml` no son datos de una fuente sino
lo que fija el plan desde el spec (los siete tipos, sus códigos, el precepto de cada uno y las nueve entradas de la
DA 1ª.2 y del art. 112 LPAC), como `data/normas.yaml` en H5: T005 los escribe tal como los dan contracts/datos-de-plazos.md
§2 y §3, con el único identificador que el plan no conoce —el del texto refundido de Tráfico— tomado de la búsqueda
grabada, y `TestPreceptosDeLosPlazos` los comprueba contra lo grabado (FR-030, FR-031; research D16, D17).

**Lo que este fichero añade al plan, comprobado en el árbol en esta sesión** (lectura y `grep`, sin ejecutar ningún
test):

- **El paso 3 del plan se parte en dos tareas** (T003, la lectura y la escritura de los ficheros y el calendario; T004,
  el cómputo, los avisos, los errores y la salida), y **el paso 6, en dos** (T006, lo nuevo de `territorio`; T008, el
  BOCM), para que cada sesión de un paso tenga un diff acotado; cada parte deja `make ci` en verde.
- **`internal/core/territorio` no expone hoy las comunidades ni sus nombres** (`registro.go`: `comunidadRegistrada`
  sin exportar; `nombres.go`: `Plegar` y `formasDelNombre`), y el BOE cruza sus columnas con las 19 comunidades «con la
  regla compacta de D13» (research D14, S2), la misma que el BOCM. T006 añade, junto a `Registro.MunicipiosDe` que pide
  el plan, `Registro.Comunidades` y la forma compacta de un nombre, para que la regla tenga una sola definición y
  `data/territorio/` un solo lector; ningún paquete del dominio importa nada nuevo (R1).
- **T016, el cierre**, no está en los trece pasos del plan: `make ci` deja los perfiles de cobertura y no los compara
  con ningún umbral, así que el punto 9 de la Definition of Done (`internal/core/plazos` ≥ 85 %) y lo que el hito no
  puede tocar (FR-058, FR-063, FR-070) se comprueban una vez, sobre el diff entero, antes de la revisión final. Es el
  cierre de H7.3, H7.4, H23, H24, H25, H26 y H23.1.
- **FR-061 nombra `docs/WORKFLOW.md`**, y research V28 dice que hoy no nombra `legal-core`: T015 lo declara y solo lo
  toca si un `grep` encuentra una frase que el hito deje falsa.
- Existen hoy con el nombre que les da el plan: `TestManifiestoDeGrabaciones`, `TestGrabacionesSinSolape` y
  `TestIdentificadoresDeLasNormas` (`internal/evals/grabaciones_test.go`), `TestGrabarEvals` (`grabacion_test.go`),
  `UnionDeGrabaciones` (`grabaciones.go`), `TestConsultasNecesarias` (`consultas_test.go`), `TestConjuntoDeEvals` y
  `TestEvalsDelRepositorio` (`conjunto_test.go`), `TestDefinicionDelJob` (`definicion_test.go`), `ReglasDeLegalCore`
  (`conjunto.go`), `FuentesEmbebidas` (`internal/app/registro.go`), `TestEntregaDelHito` (`internal/app/e2e_test.go`)
  y el empotrado de `data/territorio/` en `data/datos.go`. La línea de `CONTRIBUTING.md` que enumera los applets es la
  salida de consola de su línea 85. El empotrado de las skills (`skills.go`) usa comodines: dos referencias nuevas no
  piden cambiarlo. Los nombres que el plan fija como nuevos no existen todavía.

**Lo que no se ha medido en esta sesión.** No se ha ejecutado ningún test ni `make ci`. Las cifras de las tareas son
las de research.md (tablas V y M) y las de los contratos, que salen de sondas y de un prototipo que ya no existe; las
tareas las vuelven a medir con sus tests. Lo grabado del BOE (formato de la resolución, festivos nacionales y
autonómicos, texto vigente de los preceptos) no existe todavía: research S1 a S5 y S8 dicen qué pasa si contradice al
plan.

**Modo**: desatendido (ADR 0018). El workflow `hito` ejecuta y verifica las tareas **una a una**, con guardián de diff:
cada tarea declara en su propia línea, tras «Rutas:», **todas** las rutas que crea o modifica, y solo toca esas (más
`go.mod`, `go.sum`, `CHANGELOG.md`, el directorio del feature y el `x_test.go` de cada `x.go` declarado). Los demás
nombres que aparecen en una línea son contexto: referencias a los documentos del feature (contracts/, plan.md…) y
nombres de ficheros que la tarea solo lee. El cierre en la plataforma —publicar la rama, abrir la propuesta de cambio y
medir CI y el job de evals— lo hace el workflow tras la revisión final; quickstart.md §14 queda para el workflow y §15
para después del run.

## Formato: `- [ ] Tnnn [etiqueta?] [Story?] Descripción — FR/SC. Rutas: …`

- **[P]**: ninguna tarea lo lleva; en el bucle del workflow todo va en secuencia.
- **[aceptacion]**: solo T001, la primera.
- **[datos]**: T002, T005, T007, T008, T009 y T011. Tras el commit de T002, el paso `grabar_datos` graba sus tres
  manifiestos sin modelo; las tareas siguientes parten de lo grabado y ninguna graba ni cambia un manifiesto.
- **[Story]**: US1 hasta qué día puedo recurrir · US2 la fecha avisa de lo que le falta · US3 los festivos salen del
  boletín que los declara · US4 lo que la ley regula aparte no recibe el plazo del común · US5 el contencioso y agosto ·
  US6 el job sale en rojo si la respuesta no da la fecha o da la del común. La aceptación, la fundación y el cierre no
  llevan historia; una tarea que sirve a varias lleva la principal y nombra las demás.

## Batería de verificación por tarea

- **`make ci` en primer plano** y la tarea marcada `[X]` en el mismo turno en que termina en verde: formato · lint
  (`depguard`, `forbidigo`, `gosec`, `misspell`, `revive`, `dupl`, `gocyclo`, `paralleltest`…, también sobre los
  ficheros `//go:build grabacion`) · tests con `-race` · `test-integration` · `test-tiempos` · `vuln` · `schema-check` ·
  `skills-check` · `goreleaser-check` · secretos · módulos. Ningún registro de la verificación en la raíz del
  repositorio. Las salidas de `go test`, de `git diff` y de `make` se leen con `rtk proxy` o con sondas positivas,
  porque el proxy de la sesión las resume.
- **Sin red ni modelo** (FR-072; ADR 0032): ninguna tarea usa la red salvo la de las herramientas de Go (`make vuln`),
  ni graba nada, ni escribe en la carpeta de la evidencia; ninguna ejecuta `claude`, `make evals`, `make
  evals-sondeo`, los guiones `evals*.sh`, `TestEjecucionDelJob`, `TestMedidaDelJuez` ni `TestSondeo`. Los tests de
  `make ci` usan sesiones sintéticas, que escriben en `t.TempDir()`, y no abren ninguna sesión con modelo.
- **Sin personas a mitad**: ninguna tarea la hace, la revisa ni la desbloquea una persona.
- **Nada de memoria** (constitución, «Criterio de decisión autónoma»; plan.md, «Obligaciones para tasks.md»): ningún
  festivo, identificador `BOE-A-…` o `BOCM-…`, id de bloque ni texto de una norma se escribe de memoria. Los festivos
  salen de lo grabado por código; los identificadores, de `data/normas.yaml`, de los índices grabados o de la búsqueda
  grabada; si lo grabado contradice al plan (research S1 a S5, S8), manda lo grabado, se corrige el artefacto del
  feature que lo recordaba y se anota en `gates/supuestos.md`.
- **Ningún umbral se cumple cambiándolo** (FR-071; ADR 0029): ninguna tarea ni corrección cumple un umbral de plan.md,
  «Controles de umbral», cambiando su valor, dejándolo en `decide: false`, sacando sesiones de su `total`, retirando su
  test, quitando un caso de una tabla o cambiando las evals. Cada control nuevo se ve en rojo antes de darlo por bueno
  (la verificación de cada tarea y los mutantes de T014, que no quedan en el diff).
- **Lo que no cambia** (FR-058, FR-063, FR-070; plan.md, «Obligaciones para tasks.md»): `.github/workflows/`,
  `skills/boe-legislacion/`, `skills/jurisprudencia/`, `evals/boe-legislacion/`, `evals/jurisprudencia/`,
  `internal/mcp`, `internal/empaquetado`, `docs/SOURCES.md`, `scripts/verify-sources.sh`, `docs/ROADMAP.md`,
  `docs/ADR/`, `docs/USO.md`, `docs/HISTORIAL.md`, `CLAUDE.md`, `web/`, las cuatro evals de hoy de `legal-core`, el
  manifiesto de H5, el de `internal/source/boe` y los specs, planes e informes de otros hitos. Ningún valor de
  concurrencia ni ningún tope del job.
- **Genericidad territorial** (FR-073; principio IX): ningún municipio con trato propio en código, `data/` ni la
  skill; Leganés, Tordesillas, Madrid, Brunete, El Vellón, Majadahonda y Arroyomolinos solo aparecen en tests, golden,
  guiones, evals y la tabla de casos.
- **Ningún test escribe fuera de `t.TempDir()`**, salvo los de derivación y los golden con su bandera de actualizar; lo
  temporal de una tarea va bajo `$TMPDIR`, nunca bajo `/tmp`.
- **Sin atajos** (plan.md, «Comprobación contra la rúbrica», g): ningún `//nolint`, `t.Skip`, TODO ni error
  silenciado; ningún test se desactiva ni se retira sin el que lo sustituye; ninguno pierde lo que comprobaba. Si
  `dupl`, `gocyclo`, `paralleltest` o `misspell` marcan algo, se reestructura o se renombra; lo que `misspell` no
  admite de una palabra española va a `ignore-rules`, nunca a `//nolint`.
- **Antes de tocar una función, una constante o un ayudante que usan los tests**, `grep` de sus usos en todos los
  `_test.go` del paquete: lo que haya que alinear entra en la misma tarea, y sus ficheros están en sus rutas.
- **Proporcionalidad** (constitución, «Gates»; plan.md, «Trazabilidad», último párrafo): ninguna tarea añade un caso,
  un mensaje ni un test para un estado sin vía real —un fichero empotrado que no carga, una grabación tocada a mano, una
  evidencia que falta—: los cubre la regla genérica (defecto `inesperado`, código 1, o el test que falla).

---

## Fase 1: Aceptación — la suite congelada de la entrega

**Objetivo**: los cinco guiones que describen `plazos calcular` desde la terminal y desde un cliente MCP, en rojo por
una aserción.

**Prueba independiente**: el «rojo primero» del workflow sobre cada guion; después del bucle, `TestEntregaDelHito` con
los cinco `h9-plazos-*` (quickstart.md §2).

- [X] T001 [aceptacion] Los cinco guiones `testscript` de la entrega, escritos desde spec.md (US1.1, US2.1 a US2.3, US5.1 y US5.2; FR-001 a FR-013, FR-026, FR-028, FR-085) y desde contracts/applet-plazos.md §1 a §6, sin código de producto, sin datos y sin ninguna eval (plan.md, «Aceptación e2e»: las cinco evals entran con el formato que las admite, en T011). Son `plazos-calcular.txtar`, `plazos-cobertura.txtar`, `plazos-codigos.txtar`, `plazos-herramienta.txtar` y `plazos-sin-efectos.txtar`, con esos nombres exactos, porque plan.md, «Controles de umbral», nombra sus copias activadas. **Común a los cinco**: un comentario de cabecera que nombra los FR, SC y escenarios que cubre y que no contiene los textos `usage: `, `unknown command`, `unexpected command` ni `cannot parse` (el «rojo primero» los toma por un guion ilegible); antes de cualquier otra orden, la precondición `exec kitlegal --help` seguida de `stdout '^\s+plazos\s'`, que hoy se ejecuta bien y falla en su aserción (el precedente, los guiones `h23-cita-`); después, solo las órdenes y variables que el arnés ya tiene (`exec`, `stdout`, `stderr`, `cmp`, `arbol`, `mcp`, `KITLEGAL_T0_BIN`; `exec sed` para quitar `fecha_consulta` y `exec sh -c '…; test $? -eq <n>'` para un código exacto, como en `h23-cita-herramientas` y `h23-cita-sin-efectos`), sin ninguna nueva; los municipios, por su nombre. Afirman lo que el spec y el contrato fijan —códigos, `ok`, la clase, `"fuente":"kitlegal.plazos"`, `"url":"kitlegal:applet/plazos"`, claves y valores de vocabulario de `data`, fechas, preceptos y `source`— y, del texto de un aviso, solo el municipio y el año que debe nombrar; nada del mensaje de un error salvo lo de Arroyomolinos que fija FR-012, ni `hash` salvo su igualdad con `--offline`, ni `fecha_consulta`. **plazos-calcular** (US1.1, US5.1, US5.2; FR-001 a FR-003, FR-005, FR-006, FR-010, FR-013; SC-009): `plazos calcular reposicion 2026-09-09 Leganés Leganés --json` termina con 0 y da `"fecha":"2026-10-13"`, `"tipo":"reposicion"`, `"preceptos":[{"norma":"BOE-A-2015-10565","bloque":"a124","apartado":"124.1"}]`, `"municipios":[{"codigo_ine":"28074","nombre":"Leganés"}]` (una vez), los festivos `{"fecha":"2026-10-09","ambito":"local","municipio":"28074","source":"BOCM-20251212-34"}` y `{"fecha":"2026-10-12","ambito":"nacional","source":"BOE-A-2025-21667"}` y `"avisos":[]`, en el orden de claves del ejemplo de contracts §4; la misma orden con `--offline` da el mismo sobre salvo `fecha_consulta` (los dos pasados por `exec sed` y comparados con `cmp`: mismo `data` y mismo `hash`); `plazos calcular contencioso 2026-06-15 Madrid --json`, 0, `"fecha":"2026-09-15"` y los preceptos `a46` (`46.1`) y `a128` (`128.2`) de `BOE-A-1998-16718`; y `plazos calcular contencioso 2026-10-27 Madrid --json`, 0, `"fecha":"2026-12-28"` y un aviso de código `dias-del-art-183-lopj`. **plazos-cobertura** (US2.1 a US2.3; FR-026, FR-085): `reposicion 2026-09-09 Tordesillas Tordesillas`, 0, `"fecha":"2026-10-09"`, un aviso `faltan-festivos-locales` cuyo texto nombra Tordesillas y 2026, ninguna frase «Tordesillas no tiene festivos» (aserción negativa), y en `cobertura` la entrada `{"ambito":"local","municipio":"47165","desde":2026,"hasta":2026,"falta":"sin-festivos-locales"}`; `reposicion 2026-12-14 Leganés Leganés`, 0, `"fecha":"2027-01-14"` y un aviso `faltan-festivos-locales` cuyo texto nombra Leganés y 2027; y `generico 2024-12-30 Leganés Leganés --duracion 3 --unidad dias-habiles`, 0, `"fecha":"2025-01-03"` y los avisos `faltan-festivos-nacionales-y-autonomicos` y `faltan-festivos-locales`. **plazos-codigos** (FR-007, FR-010, FR-012): terminan con 2, `"ok":false` y `"clase":"argumentos"` cada fila de 2 de contracts §6 —el tipo `inventado`; la fecha `2026-02-30` y una que no es `AAAA-MM-DD`; `generico` sin `--duracion`, con `--duracion 1.5` y con `--duracion 0`; `reposicion --duracion 2`; `silencio --duracion 3` sin `--unidad`, `generico --duracion 3 --unidad semanas` y `reposicion --unidad meses`; `reposicion` con un municipio y con tres; `contencioso` con dos—; `reposicion 2026-09-09 Arroyomolinos Leganés` con 2 y un mensaje que lleva «municipio «Arroyomolinos»: » y los dos candidatos con los códigos INE que dan `data/territorio/municipios.yaml` y el guion `territorio-matriz.txtar`; `reposicion 2026-09-09 Inexistente Leganés` con 3 y `"clase":"no-encontrado"`; y `inventado 2026-09-09 Inexistente Leganés` con 2 (el tipo va antes que el municipio). **plazos-herramienta** (FR-001, FR-085): con la orden `mcp` y `KITLEGAL_T0_BIN`, el servidor anuncia `plazos_calcular`; una llamada con `tipo` `reposicion`, `inicio` `2026-09-09` y `municipios` `["Leganés","Leganés"]` da `resultado`, y su sobre, quitado `fecha_consulta`, es igual con `cmp` al de la orden de los mismos argumentos; una con `tipo` `inventado` da `error argumentos`, y la llamada siguiente del mismo proceso da `resultado`. **plazos-sin-efectos** (FR-013, FR-028; SC-009): con la caché y el grafo ya con contenido —una lectura de `boe articulo` sobre la reproducción, como en `h23-cita-sin-efectos`—, `arbol cache` antes; después, `plazos calcular` con una orden que termina con 0, una que falla con 2, una con `--offline --no-graph` y una con `--dry-run`, y una llamada a `plazos_calcular` con `resultado`, afirmado en la salida de `mcp`; `arbol cache` otra vez y `cmp` de los dos listados. **Verificación**: los cinco se validan contra un prototipo de un solo uso bajo `$TMPDIR`, fuera del repositorio —una copia de `HEAD` exportada con `git archive`, con el applet mínimo de contracts y un calendario mínimo escrito solo para el prototipo con los días decisivos de research D27 (el 9 de octubre de 2026 local de Leganés, como lo da el BOCM de la evidencia del ADR 0039; el 12 de octubre de 2026 y el 1 de enero de 2025, nacionales)—, ejecutado con `go -C`: allí los cinco pasan, y con un mutante por guion (la prórroga del art. 30.5 quitada, el aviso local quitado, el orden de los errores invertido, la herramienta sin registrar, un fichero escrito en la caché por el verbo) el suyo falla; el prototipo no se versiona; en el repositorio, el «rojo primero» da los cinco en rojo por la aserción de su precondición; `make ci` en verde, sin ningún fichero fuera del directorio de la suite — FR-001, FR-002, FR-003, FR-005, FR-006, FR-007, FR-010, FR-012, FR-013, FR-026, FR-028, FR-085; SC-009. Rutas: specs/023-h9-internal-core-plazos/aceptacion/

**Checkpoint**: la entrega está descrita y congelada; nada del producto ha cambiado.

---

## Fase 2: Fundación — lo que se graba, el dominio de los plazos y los datos que fija el plan

**Objetivo**: los tres manifiestos que graba `grabar_datos`; el paquete `internal/core/plazos`, puro, con el cómputo
entero sobre calendarios sintéticos; y `data/plazos.yaml` y `data/cobertura.yaml` comprobados contra lo grabado.

**Prueba independiente**: quickstart.md §10 y §11 (la parte de `internal/core/plazos`).

- [X] T002 [datos] Los tres manifiestos de grabación y sus tests `//go:build grabacion`, y **nada más**: ningún analizador, ninguna derivación, ningún fichero de `data/` (contracts/datos-de-plazos.md §4.1; research D15, V2 a V5, V20, V21, V29, S5, S8; plan.md paso 2, «Datos externos»). **`internal/source/boe/fiestas`**, paquete nuevo: `doc.go` con el comentario del paquete (el adaptador de `boe.fiestas-laborales`, sin consulta en ejecución); `testdata/grabaciones.json` con `"fuente":"boe.fiestas-laborales"` y las tres resoluciones de FR-020, cada una con su año y su dirección `https://www.boe.es/diario_boe/txt.php?id=<id>` (`BOE-A-2024-21316` 2025, `BOE-A-2025-21667` 2026, `BOE-A-2026-21254` 2027); y `grabacion_test.go`, `TestGrabarFiestasLaborales`, que pide cada una con `internal/httpx` (`ConFuente("boe.fiestas-laborales")`, `ConRaizDeGrabacion` con la raíz de `KITLEGAL_EVIDENCIAS`, que exige, e intervalo de 1 s, el de su fila de `docs/SOURCES.md`), solo GET. **`internal/source/bocm`**, paquete nuevo: `doc.go`; `testdata/grabaciones.json` con `"fuente":"bocm.fiestas-locales"` y las siete disposiciones de FR-020 con su año (`BOCM-20241213-23`, `BOCM-20241227-32`, `BOCM-20250203-16` y `BOCM-20250217-24` de 2025; `BOCM-20251212-34`, `BOCM-20251229-19` y `BOCM-20260108-18` de 2026), cada una en XML en `https://www.bocm.es/boletin/CM_Orden_BOCM/AAAA/MM/DD/<id>.xml` con la fecha de su identificador, comprobada contra las direcciones de la evidencia del ADR 0039; y `grabacion_test.go`, `TestGrabarFiestasLocales`, igual, a 10 s. **`internal/evals`**: `testdata/grabaciones.json`, segundo manifiesto, con `"fuente":"boe.legislacion-consolidada"` y la forma del de H5 (`busqueda`, `titulo_empieza_por`, `bloques`, `para`): LPAC (`a24`, `a30`, `a112`, `a122`, `a124`, `daprimera`), LJCA (`a46`, `a128`), LOPJ (`acientoochentaydos`, `acientoochentaytres`, `acientoochentaycinco`), LTAIBG (`a24`), LGT (`a97`), Código Civil (`art5`) y el texto refundido de la Ley sobre Tráfico (`a1`, con su búsqueda por título, sin identificador escrito: lo da la búsqueda grabada, S5); cada id de bloque, salvo el `a1` de Tráfico (S8), comprobado contra el índice grabado de su norma; `TestGrabarEvals` recorre los dos manifiestos, siembra la caché desde lo grabado y pide solo lo que falta, y escribe lo del segundo en `internal/evals/testdata/boe.legislacion-consolidada/`; `TestManifiestoDeGrabaciones` valida también el segundo (fuente, forma, ids de bloque con `boe.ValidarBloque`). Verificación: con una `fuente` vacía un momento en el segundo manifiesto, `TestManifiestoDeGrabaciones` falla (se restaura); `go vet -tags grabacion` de los tres paquetes compila; `make ci` en verde. Tras el commit graba `grabar_datos`; ninguna tarea vuelve a tocar un manifiesto — FR-020, FR-027, FR-030, FR-031, FR-052. Rutas: internal/source/boe/fiestas/doc.go, internal/source/boe/fiestas/grabacion_test.go, internal/source/boe/fiestas/testdata/grabaciones.json, internal/source/bocm/doc.go, internal/source/bocm/grabacion_test.go, internal/source/bocm/testdata/grabaciones.json, internal/evals/testdata/grabaciones.json, internal/evals/grabacion_test.go, internal/evals/grabaciones.go

- [X] T003 El dominio de los plazos, primera parte: los tipos, el calendario y la forma canónica de los ficheros de festivos, puros y sobre datos sintéticos, y **nada más** (data-model §1; contracts/datos-de-plazos.md §1.1, §1.2 y §2; research D1, D3, D12, D17, D25, V16, V27; plan.md paso 3). Paquete nuevo `internal/core/plazos`, que solo importa `internal/core/schema`, `go.yaml.in/yaml/v3` y la biblioteca estándar, ni `os` ni `io` (R1, también en sus tests). **`doc.go`**: el comentario del paquete. **`tipos.go`**: `Tipo` (`codigo`, `nombre`, `computo` `lpac` o `contencioso`, `duracion` opcional con `cantidad` y `unidad`, `preceptos` con `norma`, `bloque`, `apartado` y `fija`, y `unidades` en `generico`) y `LeerTipos(bytes)`, con `KnownFields`, que exige los siete códigos de contracts §2 y ninguno más, `duracion` en los cinco que la fijan y en ningún otro, `unidades` solo en `generico` con las tres unidades, y `fija: duracion` en un precepto de cada tipo con duración. **`calendario.go`**: `Calendario`, inmutable tras cargar y seguro entre goroutines, y `CargarCalendario(nacionales []byte, locales map[string][]byte)`, que lee las dos formas de contracts §1.1 y §1.2 —lo que no se puede leer es un error que el applet dará como `inesperado`, sin reglas de carga propias (D25)—; las consultas de data-model §1: si un día es inhábil para unos municipios dados (sábado, domingo, nacional o autonómico sin `parte` de la comunidad de alguno, local de alguno), los festivos de un día con su ámbito, municipio y `source`, las partes de un día en una comunidad y la cobertura de unos años para unos municipios, por ámbito y tramos (`disposicion` o `falta`: `sin-cargar`, `sin-festivos-locales`, `no-comunicado`, `linea-sin-leer`, `sin-linea`); y los vocabularios cerrados de ámbitos y faltas, con `sin-municipio` el de una línea. **`festivos.go`**: la escritura canónica de los dos ficheros (`EscribirFestivosNacionales`, `EscribirFestivosLocales`), con la línea de cabecera de fichero generado de contracts §1, sus claves en ese orden y sus listas ordenadas como dice §1.1 y §1.2. **Tests**: `TestLeerTipos` (`tipos_test.go`), con el `data/plazos.yaml` de contracts §2 como constante y una variante por cada regla que lo rechaza; `TestCargarCalendario` (`calendario_test.go`), con un calendario sintético de dos comunidades, una con fichero local, festivos con y sin `parte`, una falta de cada motivo y dos años, cada consulta; y `TestEscribirFestivos` (`festivos_test.go`), ida y vuelta —lo escrito se carga y se vuelve a escribir igual byte a byte— y el orden de filas y claves. Verificación: con la comprobación de los siete códigos quitada un momento, `TestLeerTipos` falla (se restaura); `make ci` en verde — FR-003, FR-008, FR-011, FR-020, FR-022, FR-025, FR-026, FR-030. Rutas: internal/core/plazos/doc.go, internal/core/plazos/tipos.go, internal/core/plazos/calendario.go, internal/core/plazos/festivos.go

- [X] T004 El dominio de los plazos, segunda parte: el cómputo de la LPAC, del silencio y del contencioso, los avisos, los errores de argumentos y la salida, con la rejilla de las propiedades, todo sobre calendarios sintéticos (contracts/applet-plazos.md §2, §4, §5 y §6; data-model §1; research D2 a D10, D22, D28, D29; plan.md paso 3). **`calcular.go`**: `Peticion` (tipo, inicio, duración y unidad como cadenas, y los municipios ya resueltos como `{codigo_ine, nombre, comunidad}`) y `Calcular(calendario, tipos, peticion)`, que da el `Plazo` o un error de argumentos; nunca resuelve un municipio (lo hace el applet con `territorio`). **`computo.go`**: desde el día siguiente a `inicio` (FR-002); días hábiles, el enésimo hábil; días naturales, el enésimo día y la prórroga; meses, de fecha a fecha o el último día del mes, y la prórroga (art. 30.2 a 30.6; D7); el silencio, el día siguiente al vencimiento calculado así (D8); el contencioso, con un municipio, agosto que no corre (partida desde el 31 de agosto si `inicio` o el día siguiente es de agosto; vencimiento al mes siguiente si el tramo toca agosto), el último día con los inhábiles del art. 182.1 (sábado, domingo, 24 y 31 de diciembre, nacional, autonómico y local de la sede) y la prórroga del 185.2, y los días del 24 de diciembre al 6 de enero contados (D9); los festivos saltados, como dice contracts §2, nunca un sábado ni un domingo. **`avisos.go`**: `CodigosDeAviso`, las cuatro etiquetas y los textos de contracts §5, con `<años>`, `<motivo>` y la disposición de su ámbito, cada uno solo si el tramo pasa por lo que lo causa, en el orden de §4, y ninguno por una entidad menor (FR-024). **`errores.go`**: los defectos de clase `argumentos` en el orden de D10 —tipo, fecha, duración, unidad, número de municipios (dos en `lpac`, uno en `contencioso`)—, con un mensaje que nombra la entrada, construidos como los de `territorio`. **`salida.go`**: el `Plazo` y sus partes con las claves de contracts §4 y ninguna más, `cobertura` por tramos de años (como mucho cinco por ámbito), los municipios una vez cada uno en el orden dado, y los enumerados en las etiquetas de los tipos para `--describe`. **Tests**: `TestCalcular` (`calcular_test.go`), una tabla con cada regla del cómputo, cada aviso, cada fila de errores y su orden, sobre calendarios sintéticos; `TestAvisosDePlazos` (`avisos_test.go`), códigos, etiquetas y textos; y `TestPropiedadesDelComputo` (`propiedades_test.go`), sobre la rejilla de research D22 —cada fecha de inicio de 2025 a 2027, cada N de 1 a 40 y cada pareja de un juego de municipios sintéticos con festivos en días distintos—: (1) N días hábiles nunca acaban en un inhábil de ninguno de los municipios; (2) entre el día siguiente a `inicio` y la fecha hay exactamente N hábiles y la fecha es hábil; (3) un plazo de meses acaba el mismo día del mes, el último, o el hábil siguiente si ese es inhábil; sin `testing/quick` ni dependencias nuevas. Verificación: con la prórroga del art. 30.5 quitada un momento, `TestPropiedadesDelComputo` y `TestCalcular` fallan; con un solo municipio en lugar de los dos, `TestCalcular` falla (se restauran); `go test -cover ./internal/core/plazos/` ≥ 85 %; `make ci` en verde — FR-002 a FR-012, FR-024, FR-026, FR-048, FR-081; SC-004. Rutas: internal/core/plazos/calcular.go, internal/core/plazos/computo.go, internal/core/plazos/avisos.go, internal/core/plazos/errores.go, internal/core/plazos/salida.go, internal/core/plazos/propiedades_test.go

- [ ] T005 [datos] Los tipos y la cobertura por materia en `data/`, con el esquema y el lector de la cobertura, la tercera raíz de grabaciones y la verificación de los preceptos contra lo grabado (contracts/datos-de-plazos.md §2, §3 y §4.2; research D15 a D17, S4, S5, S8; plan.md paso 4). **`data/plazos.yaml`**: el de contracts §2, tal cual. **`data/cobertura.yaml`**: las nueve entradas de contracts §3 en su orden, con la norma solo en la aplicación de los tributos (LGT `BOE-A-2003-23186`, de `data/normas.yaml`, `a97`) y en los sancionadores de tráfico, cuyo identificador se toma de la respuesta de la búsqueda grabada en `internal/evals/testdata/boe.legislacion-consolidada/`, nunca de memoria. **`schemas/cobertura.yaml.json`**: el esquema de ese fichero, escrito como `normas.yaml.json` y `jerarquia.yaml.json`. **`internal/skills/cobertura.go`**: `LeerCobertura`, que valida con `ValidarDocumentoYAML` contra ese esquema, como `jerarquia.go`; `TestLeerCobertura` con un documento válido y uno por cada forma que el esquema rechaza. **`internal/evals/grabaciones.go`**: `UnionDeGrabaciones` da el tercer conjunto, la raíz nueva de `internal/evals/testdata/`. **`internal/evals/grabaciones_test.go`**: `TestGrabacionesSinSolape` cubre la raíz nueva; y `TestPreceptosDeLosPlazos`, nuevo, con la caché preparada desde la unión, lee cada bloque con `boe articulo` en proceso y comprueba los cuatro puntos de contracts §4.2 —cada norma y cada bloque de los dos ficheros grabados; la duración en palabras en el apartado de cada precepto con `fija: duracion`; cada `norma.identificador` de la cobertura igual al que resuelve la búsqueda grabada; y las palabras clave de la tabla de §4.2—, nombrando el precepto y la palabra que falta. Si lo grabado contradice al plan (S4: el art. 182.1 sin el 24 y el 31 de diciembre, una duración en otro apartado, una palabra clave distinta), se corrige el fichero de `data/` o la tabla del contrato con lo grabado y se anota en `gates/supuestos.md`; y si cambia una regla del cómputo, se ajusta en esta tarea con todo lo que la fija: la regla en `computo.go` (y en `calcular.go` si toca la petición), su caso en la tabla de `TestCalcular` (`calcular_test.go`), la propiedad de `TestPropiedadesDelComputo` (`propiedades_test.go`) que la recorra, y el texto del aviso y su caso de `TestAvisosDePlazos` (`avisos.go`, `avisos_test.go`) si el aviso nombra la regla, como `dias-del-art-183-lopj` con los inhábiles del art. 182. Es la última tarea que escribe código en `internal/core/plazos` (T014 solo puede reforzar tests): si su cobertura baja del 85 %, los tests que la suben van en el `_test.go` del fichero que la pierde, o en `ramas_test.go` si ese fichero no está entre estas rutas. Verificación: con `reposicion` a dos meses un momento en `data/plazos.yaml`, con un bloque que no está grabado y con una palabra clave cambiada, `TestPreceptosDeLosPlazos` falla nombrándolos (se restauran); `go test -cover ./internal/core/plazos/` ≥ 85 %; `make ci` en verde — FR-008, FR-010, FR-030, FR-031, FR-052, FR-088. Rutas: data/plazos.yaml, data/cobertura.yaml, schemas/cobertura.yaml.json, internal/skills/cobertura.go, internal/evals/grabaciones.go, internal/core/plazos/computo.go, internal/core/plazos/calcular.go, internal/core/plazos/avisos.go, internal/core/plazos/propiedades_test.go, internal/core/plazos/ramas_test.go

**Checkpoint**: lo grabado está en el árbol; el cómputo entero existe y se prueba sin `data/`; los tipos y la cobertura
están en `data/` y comprobados contra lo grabado.

---

## Fase 3: US3 — Los festivos salen del boletín que los declara (P1)

**Objetivo**: `data/festivos/` derivado por código de lo grabado, byte a byte, con cada fila con su `source`, los 179
municipios de la Comunidad de Madrid cubiertos o con su motivo, y el cotejo con el fichero CKAN con exactamente las
diez diferencias de su `LEEME.md`.

**Prueba independiente**: quickstart.md §9.

- [ ] T006 [US3] Lo que `territorio` da a los dos adaptadores de los boletines: los municipios de una comunidad con sus formas, las comunidades con sus nombres y la forma compacta de un nombre, una sola definición para los dos (research D13, D14, S2, V7; plan.md, R1 y «Trazabilidad»; «Lo que este fichero añade al plan», arriba). **`registro.go`**: `Registro.MunicipiosDe(comunidad)`, los municipios de esa comunidad con su código INE, su nombre oficial y sus formas plegadas (las de `formasDelNombre`), en orden de código; y `Registro.Comunidades()`, las comunidades y ciudades autónomas cargadas con su código y su nombre, en orden de código. **`nombres.go`**: `FormaCompacta(nombre)`, el pliegue de `Plegar` sin las palabras «de», «del», «la», «las», «los» y «el» ni los blancos. Ninguna lista de alias ni ningún municipio con trato propio. **Tests** (`registro_test.go`, `nombres_test.go`), con las fuentes sintéticas de los tests de hoy, sin `os` (R1): una comunidad con un municipio de artículo pospuesto y otra comunidad con un municipio del mismo nombre, que `MunicipiosDe` no mezcla; `Comunidades` en su orden; y `FormaCompacta` de «Talamanca del Jarama» y «Talamanca de Jarama», iguales, y de «Moraleja de En medio» y «Moraleja de Enmedio», iguales (research V7). Verificación: con `MunicipiosDe` sin filtrar por comunidad un momento, su test falla (se restaura); `make ci` en verde — FR-022. Rutas: internal/core/territorio/registro.go, internal/core/territorio/nombres.go

- [ ] T007 [datos] [US3] Los festivos nacionales y autonómicos, derivados de las tres resoluciones del BOE grabadas tras T002 (contracts/datos-de-plazos.md §1.1 y §1.3; data-model §2; research D14, S1, S2; plan.md paso 5). **`internal/source/boe/fiestas/resolucion.go`**: `LeerResolucion(bytes)`, que lee el anexo de una resolución tal como está grabado —una fila por fecha, una columna por comunidad o ciudad autónoma, la marca «***» o «**» (nacional) o «*» (autonómico) y las notas que limitan una fiesta a una parte de la comunidad—; una fila o una nota que no se lee sin interpretarla queda fuera, como en el BOCM (ADR 0039, punto 3). **`internal/source/boe/fiestas/derivar.go`**: la derivación, que recibe las resoluciones (leídas con `httpx.Replay` sobre la raíz donde las dejó `grabar_datos`) y el registro de `territorio`, cruza cada columna con una de las 19 comunidades de `Registro.Comunidades` por `FormaCompacta` (S2: sin lista de alias; una columna que no casa hace fallar la derivación nombrándola), y escribe con `plazos.EscribirFestivosNacionales` una fila por fecha, ámbito y parte, con `source` la resolución de su año y `parte` solo en las limitadas (FR-025). **`data/festivos/nacionales-y-autonomicos.yaml`**: el que escribe `TestFestivosNacionalesDerivados` con `-actualizar-festivos`; no se edita a mano. **Tests**: `TestLeerResolucion` (`resolucion_test.go`), con lo grabado de cada año: el 12 de octubre de 2026 nacional en las 19 columnas, al menos una fila con `parte` (FR-025; si lo grabado no trae ninguna, se anota en `gates/supuestos.md` y T010 falla, como dice contracts/casos-de-plazos.md) y una fila autonómica; y `TestFestivosNacionalesDerivados` (`derivar_test.go`), que deriva desde lo grabado y el registro real (cargado de `data/territorio/` en el test) y falla si el fichero no es igual byte a byte, si una fila no lleva `source` o si alguna de las 19 comunidades no tiene festivos en 2025, 2026 o 2027; sin la bandera no escribe nada. Si el formato grabado contradice a S1, manda lo grabado y se anota. Verificación: con una línea del fichero cambiada un momento, el test falla (se regenera); con una fila sin `source`, falla; `make ci` en verde — FR-011, FR-020, FR-025, FR-083; SC-006. Rutas: internal/source/boe/fiestas/resolucion.go, internal/source/boe/fiestas/derivar.go, data/festivos/nacionales-y-autonomicos.yaml

- [ ] T008 [datos] [US3] Los festivos locales de la Comunidad de Madrid, derivados de las siete disposiciones del BOCM grabadas tras T002, con el cruce con `data/territorio/`, las correcciones en su orden y el cotejo con el fichero CKAN (contracts/datos-de-plazos.md §1.2 y §1.3; data-model §2; research D13, V6 a V10, D30; plan.md paso 6). **`internal/source/bocm/disposicion.go`**: `LeerDisposicion(bytes)`, que lee las líneas `.⇥<nombre>[:] <cuerpo>[.]` del XML con la gramática de D13 —el nombre hasta los dos puntos o, sin ellos, hasta la primera cifra; «no comunicado por el ayuntamiento» sin distinguir mayúsculas; fechas `<d>[ y <d>] [de ]<mes>` unidas por « y », con blancos y puntos sueltos tolerados (V8)—, y da por línea festivos, `no-comunicado` o `linea-sin-leer` (FR-023: la de Valdeolmos-Alalpardo con paréntesis). **`internal/source/bocm/cruce.go`**: el cruce solo con `Registro.MunicipiosDe` de la comunidad del boletín: primero la forma compacta igual a la de una forma de un único municipio; si ninguno, el nombre que empieza por todas las palabras de una forma de un único municipio seguidas de «de» o «del»; si no, `sin-municipio` (FR-022, FR-024). **`internal/source/bocm/derivar.go`**: la derivación de un año desde la lista de disposiciones que recibe, en orden de publicación (la fecha de su identificador): cada línea de una posterior sustituye entera la del municipio, con su `source` (FR-021; data-model §2, transiciones); al final, el municipio sin línea, `sin-linea`; escribe con `plazos.EscribirFestivosLocales`. Recibe la lista para que T010 componga el caso de Brunete con la principal de 2026 sola (D30). **`data/festivos/locales/13.yaml`**: el que escribe `TestFestivosLocalesDerivados` con `-actualizar-festivos`; no se edita a mano. **Tests**: `TestLeerDisposicion` (`disposicion_test.go`), las erratas de V8 y la línea de Valdeolmos-Alalpardo de 2025 y de 2026; `TestCruceConElTerritorio` (`cruce_test.go`), los seis nombres de V7 a su código INE, Cerceda, Mataelpino y El Espartal a `sin-municipio` y la Cerceda de A Coruña (`15024`) nunca; `TestFestivosLocalesDerivados` (`derivar_test.go`), que deriva desde lo grabado y el registro real y falla si el fichero no es igual byte a byte, si una fila no lleva `source`, si uno de los 179 municipios de la comunidad 13 no tiene en 2025 o en 2026 festivos o una falta (o tiene las dos cosas), o si una línea de lo grabado no da ni festivos de un municipio ni una falta; y afirma US3.2 (Majadahonda, `28080`, el `2025-09-15` con `BOCM-20250203-16` y no el 14) y US3.3 (Brunete, `28026`, el `2026-05-15` y el `2026-09-14` con `BOCM-20251229-19`); y `TestCotejoConElCatalogo` (`cotejo_test.go`), que lee el fichero CKAN y el `LEEME.md` de la evidencia del ADR 0039 sin cambiarlos, coteja los festivos locales de 2025 por código INE contra las filas de `entidad_codigo` 0 y exige exactamente las diez diferencias de FR-084, ninguna más ni ninguna menos, cada una nombrada en el test. Verificación: con las disposiciones de un año sin ordenar un momento (la corrección antes que la principal), `TestFestivosLocalesDerivados` falla; con la regla de prefijo quitada, `TestCotejoConElCatalogo` da diferencias de más y falla (se restauran); `make ci` en verde — FR-020 a FR-024, FR-083, FR-084; SC-006, SC-007. Rutas: internal/source/bocm/disposicion.go, internal/source/bocm/cruce.go, internal/source/bocm/derivar.go, internal/source/bocm/cotejo_test.go, data/festivos/locales/13.yaml

**Checkpoint**: `data/festivos/` sale de lo grabado y lo protege `make ci`.

---

## Fase 4: US1, US2 y US5 — `plazos calcular` en el binario y como herramienta (P1, P2) 🎯 MVP

**Objetivo**: el applet `plazos` en los dos registros, sin red, sin caché y sin grafo, con su esquema, sus golden, su
herramienta y la tabla de casos calculados a mano.

**Prueba independiente**: quickstart.md §3 a §8 y §11; los cinco guiones de T001, ejercidos con copias momentáneas.

- [ ] T009 [datos] [US1] El applet `plazos` con su verbo `calcular` en los dos registros, con su esquema, sus golden, la garantía sin red y las listas literales que su llegada cambia, en una sola tarea porque con el applet registrado `schema-check` exige su esquema y cuatro guiones enumeran los applets o cuentan las herramientas (contracts/applet-plazos.md §1 a §8; research D1, D2, D4, D5, D10, D11, V11 a V13, V17, V26; plan.md paso 7, «Tests y guiones existentes que cambian»). **`internal/app/plazos.go`**, nuevo: `AppletPlazos(fuentes)`, con el verbo `calcular` y ningún verbo por omisión; argumentos de cadena (D2): `tipo`, `inicio` y `municipios` de posición (lista, uno o más) y `--duracion` y `--unidad` como banderas, sin `enum` de Kong ni bandera entera (V12); comprueba primero con el dominio tipo, fecha, duración, unidad y número de municipios, después resuelve cada municipio con `territorio.Registro.Resolver` —2 con el mensaje de `territorio` de cada uno tras «municipio «<pedido>»: », separados por « \| »; 3 `no-encontrado` solo si todo lo demás vale (FR-012; V17)— y calcula con `plazos.Calcular`; sobre con `fuente` `kitlegal.plazos`, `url` `kitlegal:applet/plazos` y la fecha del reloj del kernel; ninguna operación de grafo; unas fuentes que no cargan, `inesperado` (1). **`internal/app/registro.go`**: `FuentesDePlazos`, los ficheros de `data/festivos/` y `data/plazos.yaml` del binario, y el applet en la raíz de producción. **`data/datos.go`**: el empotrado de `festivos/nacionales-y-autonomicos.yaml`, `festivos/locales/*.yaml` y `plazos.yaml`, con sus accesores, como el de `territorio/`. **`internal/app/ejemplo/kitlegal-e2e/main.go`**: el applet en el registro de e2e. **El esquema** `schemas/plazo.json`, regenerado con `-actualizar-esquemas` y no editado, y su fila en `ficherosDeEsquemas` con su descripción (`esquemas_test.go`). **Golden** en `internal/app/testdata/plazos/`: `describe.json` (`--describe` de `plazos calcular`, con `x-banderas` `["duracion","unidad"]`) y la salida de los casos (a), (b) y (e), regenerados con `-actualizar-golden`. **Tests**: `TestAppletPlazos` (`plazos_test.go`): cada golden valida contra `schemas/plazo.json`; `Resultado.Grafo` vacío; el mismo `data` y el mismo `hash` con `--offline`; el golden de `--describe`; y los enumerados de la salida (`tipo`, `festivos[].ambito`, `avisos[].codigo`, `cobertura[].ambito`, `cobertura[].falta`) iguales a los vocabularios del dominio, como `TestTerritorioResuelto`. **Lo que el applet nuevo cambia** (research V11): las listas de applets de `TestPuntoDeEntrada` (`cmd/kitlegal/main_test.go`), `TestRegistroDeProduccion` (`registro_test.go`) y `TestRegistroDeE2E` (`kitlegal-e2e/main_test.go`); `TestHerramientasDelServidor` de 12 y 15 herramientas a 13 y 16 con `plazos_calcular`; `TestEsquemasCubrenTodosLosVerbos`; `TestArquitectura` (`internal/arch_test.go`), con `internal/core/plazos` e `internal/app/plazos.go` en la garantía sin `net`, `net/http` ni `internal/httpx` (FR-013); los guiones `argumentos.txtar` (la lista de applets), `h21-mcp-herramientas.txtar`, `h21-mcp-proceso.txtar` y `h21-mcp-protocolo.txtar` (de 15 a 16 herramientas); y la salida de consola de `CONTRIBUTING.md` que enumera los applets. Cada cambio es el dato nuevo de una lista o una cifra. Verificación: con una operación de grafo puesta un momento en el resultado, `TestAppletPlazos` falla; con `internal/httpx` importado un momento desde `plazos.go`, `TestArquitectura` falla (se restauran); los cinco guiones de T001, copiados un momento con prefijo `zz-` al directorio de guiones del arnés e2e, pasan con `TestEntregaDelHito` y se borran; `make ci` en verde — FR-001, FR-007, FR-010, FR-012, FR-013, FR-026, FR-028, FR-085; SC-009. Rutas: internal/app/plazos.go, internal/app/registro.go, data/datos.go, internal/app/ejemplo/kitlegal-e2e/main.go, internal/app/ejemplo/kitlegal-e2e/main_test.go, schemas/plazo.json, internal/app/testdata/plazos/, internal/app/esquemas_test.go, internal/app/herramientas_test.go, internal/app/registro_test.go, cmd/kitlegal/main_test.go, internal/arch_test.go, internal/app/testdata/script/argumentos.txtar, internal/app/testdata/script/h21-mcp-herramientas.txtar, internal/app/testdata/script/h21-mcp-proceso.txtar, internal/app/testdata/script/h21-mcp-protocolo.txtar, CONTRIBUTING.md

- [ ] T010 [US1] La tabla de casos calculados a mano y el festivo de una parte, sobre el applet y los ficheros empotrados (contracts/casos-de-plazos.md; research D23, D30, S1, S3; plan.md paso 8). **`internal/app/plazos_casos_test.go`**, nuevo: `TestCasosDePlazos`, un subtest por cada uno de los veinte casos de la tabla, llamando al applet como lo llama la skill —municipios por su nombre, `--duracion` y `--unidad` cuando el caso los lleva— y comparando la fecha y la lista de códigos de aviso: (a), (b) y (e) (casos 1 a 3; US1, US2, US5), el art. 30.6 con el festivo del interesado y con el de la sede (4, 5), el último día en sábado (6), el cambio de año hacia 2027 (7, 16, 18), Brunete antes de su corrección, con el applet compuesto con los festivos locales que deriva `internal/source/bocm` de la principal de 2026 sola y escribe `plazos.EscribirFestivosLocales` (8; D30), y con ella (9), la entidad menor, el 9 de octubre de 2026 en El Vellón hábil y sin aviso (10; FR-024), el silencio (11; FR-009), días hábiles y naturales (12, 13), el mes sin día equivalente (14), el contencioso con el art. 183 (15, 16), el año anterior a 2025 (17), y las coberturas de (c) y (d) contra `data/cobertura.yaml` leído con `skills.LeerCobertura` (19, 20: la entrada con la DA 1ª.2.a y la LGT, y la de la DA 1ª.2.c con el identificador de Tráfico). **`internal/app/plazos_test.go`**: `TestFestivoDeUnaParte`, que toma de los ficheros empotrados el primer festivo con `parte` en día laborable y un municipio de su comunidad de `Registro.MunicipiosDe`, y con un `generico` de un día hábil desde la víspera da ese día como fecha y el aviso `festivo-de-una-parte` con el día y la parte; falla si lo empotrado no trae ninguno (FR-011, S1). Si lo grabado contradice un razonamiento de la tabla (S3), se corrigen el razonamiento y la fecha en contracts/casos-de-plazos.md con lo grabado y se anota en `gates/supuestos.md`; los casos de los guiones congelados (1, 2, 3, 15, 17, 18) no dependen de un día incierto (D27). Verificación: con el art. 30.6 quitado un momento (solo el primer municipio), fallan los casos 4 o 5; con agosto contado en el contencioso, el 3 (daría el 17 de agosto); con la prórroga quitada, el 1 y el 6 (se restauran); `make ci` en verde — FR-002 a FR-011, FR-024, FR-026, FR-082; SC-005. Rutas: internal/app/plazos_casos_test.go, internal/app/plazos_test.go, internal/app/plazos.go

**Checkpoint**: `kitlegal plazos calcular` y `plazos_calcular` dan la fecha con su precepto, sus festivos y sus avisos;
la entrega del binario está completa y los cinco guiones pasan.

---

## Fase 5: US6 y US4 — Las evals de plazo y de cobertura, y los umbrales que deciden (P1)

**Objetivo**: el formato de eval admite el esperado de un plazo y el de una declaración de no cubierto; las cinco evals;
y los dos umbrales por modo, con `sin_activar`, que hacen fallar el job.

**Prueba independiente**: quickstart.md §12.

- [ ] T011 [datos] [US6] El formato de eval gana `plazo`, `no_cubierto` y el comando `calcular`, con su juicio sin modelo, las reglas del conjunto de `legal-core` y las cinco evals, en una sola tarea porque `make ci` lee las evals de toda carpeta y el lector estricto rechaza una clave que el tipo no tiene (contracts/evals-legal-core.md §1, §2, §4 y §5; data-model §3; research D18, D19, D24, D26, V15, V18, V24; plan.md paso 9). **El esquema** `schemas/eval.yaml.json`: `plazo` (`fecha` `AAAA-MM-DD`, `avisos` sin repetir con el enumerado de `plazos.CodigosDeAviso`, `razonamiento` no vacío) y `no_cubierto` (`norma` `BOE-A-…`), solo en una eval activa que no es sin binario ni servidor y nunca juntas; y la forma `{applet: plazos, verbo: calcular, tipo}` en el `oneOf` de los comandos. **`internal/evals/formato.go`**: `Eval.Plazo`, `Eval.NoCubierto` y `ComandoEsperado.Tipo`; `TestFormatoDePlazos` (`formato_test.go`): claves, esquema y el enumerado de `plazo.avisos` atado a `plazos.CodigosDeAviso`, como `ComprobarCodigosDeAviso`. **`internal/evals/plazos.go`**, nuevo: las tres formas fijas de contracts §2 —`⚠ PLAZO:` con la fecha justo detrás de los dos puntos y al menos una cita `formaDeCita` en la misma línea, `⚠ NO CUBIERTO:` con cada `BOE-A-…` de su línea, y la etiqueta de cada aviso— con la tolerancia de `patronDeEtiqueta`, y su juicio; `TestJuzgarPlazos` (`plazos_test.go`). **`internal/evals/juzgar.go`**: `cumpleLoEsperado` gana los motivos de contracts §2 (`plazo ausente: <fecha>`, `aviso del plazo ausente: <código>`, `declaración de no cubierto ausente: <norma>`, `línea del plazo en un asunto no cubierto: <fecha>`), y `ResultadoDeEval`, `plazo_encontrado`, `avisos_del_plazo_ausentes`, `no_cubierto_declarado` y `lineas_del_plazo`; el comando `calcular` lo satisface una orden `plazos calcular <tipo> …` que termina con 0 sin `--dry-run`, `--describe` ni la ayuda, o una llamada a `plazos_calcular` con ese `tipo` que no da error; `prohibidos` con `{applet: plazos, verbo: calcular}` cubre la orden y la llamada. **`internal/evals/consultas.go`**: `calcular` no pide ninguna consulta grabada (`TestConsultasNecesarias`). **`internal/evals/conjunto.go`**: en `ReglasDeLegalCore`, «esperado verificable» admite `plazo` y `no_cubierto`, y la regla nueva «cobertura por materia», al menos una eval de plazo y una de cobertura (FR-088); `TestConjuntoDeEvals` con un conjunto sin eval de cobertura que la incumple, y `TestEvalsDelRepositorio` con las nueve evals de `legal-core` válidas y sus consultas servidas sin red por la unión de grabaciones (FR-052). **Las cinco evals** `evals/legal-core/05-plazo-reposicion-leganes.yaml`, `06-plazo-reposicion-tordesillas.yaml`, `07-cobertura-liquidacion-tributaria.yaml`, `08-cobertura-multa-de-trafico.yaml` y `09-plazo-contencioso-agosto.yaml`, con la pregunta, los comandos, las citas, el esperado y los prohibidos de contracts §5, la fecha y los avisos de la tabla de casos (casos 1, 2 y 3, con la fecha que dejó T010) y su razonamiento a mano, y en la 08 la norma que lleva `data/cobertura.yaml`; ninguna declara `no_se_activan`, ninguna es informativa ni por materia; las cuatro de hoy no cambian. **`internal/evals/definicion_test.go`**: solo el literal `peorCasoDeLegalCore`, a `14629 s = 485 s + (⌈49 / 2⌉ + ⌈48 / 2⌉ + ⌈6 / 2⌉) × (22 s + 240 s + 10 s)`, y las líneas de los dos subtests que lo llevan (D24); ningún valor de concurrencia, ningún tope y ningún cambio de `.github/workflows/evals.yml` (FR-058). Verificación: con la forma `⚠ PLAZO:` sin la fecha delante de la cita, `TestJuzgarPlazos` da `plazo ausente`; con la eval 07 sin `no_cubierto` un momento, la regla «cobertura por materia» falla en `TestEvalsDelRepositorio`; con el tope del job por debajo de 14 629 s en una definición sintética, `TestDefinicionDelJob` falla (se restauran); `make ci` en verde — FR-048, FR-050, FR-051, FR-052, FR-057, FR-058, FR-088; SC-001. Rutas: schemas/eval.yaml.json, internal/evals/formato.go, internal/evals/plazos.go, internal/evals/juzgar.go, internal/evals/consultas.go, internal/evals/conjunto.go, internal/evals/definicion_test.go, evals/legal-core/05-plazo-reposicion-leganes.yaml, evals/legal-core/06-plazo-reposicion-tordesillas.yaml, evals/legal-core/07-cobertura-liquidacion-tributaria.yaml, evals/legal-core/08-cobertura-multa-de-trafico.yaml, evals/legal-core/09-plazo-contencioso-agosto.yaml

- [ ] T012 [US6] Los umbrales `plazo_sin_fecha:<modelo>:<modo>` y `cobertura_sin_declarar:<modelo>:<modo>` en el informe del job, con `sin_activar` de `legal-core`, y los tests que fijan los umbrales de las tres skills, en una sola tarea porque el código cambia la forma de `umbrales` de `legal-core` y los tests que la esperan vacía tienen que cambiar con él (contracts/evals-legal-core.md §3; data-model §3; research D20, V19; plan.md paso 10). **`internal/evals/informe.go`**: `medirLosPlazos` marca en las respuestas medidas de cada grupo de modo las de las evals de plazo y de cobertura y las que cuentan —de plazo, sin la línea `⚠ PLAZO:` con `plazo.fecha`; de cobertura, sin la declaración con `no_cubierto.norma` o con una línea `⚠ PLAZO:` de cualquier fecha—, cada sesión una vez como mucho; los avisos no entran (FR-055); fuera del `total` y de la `medida`, las sesiones sin medir (con su motivo de la ejecución), las de un modelo informativo y las de la eval sin binario ni servidor; y el motivo de uno que no se cumple, `umbral <nombre>: <medida> de <total> (<p> %), y tiene que ser ≤ 0,0 %: <sesión>; <sesión>…`, en orden de sesión, como el de `documento_sin_cotejar`. **`internal/evals/umbrales.go`**: una skill tiene umbrales de sus respuestas si tiene juez o si alguna eval declara `sentencias`, `plazo` o `no_cubierto`, sin ningún nombre de skill en el código; por cada modo del plan, detrás de `documento_sin_cotejar`, `plazo_sin_fecha` si alguna eval declara `plazo` y `cobertura_sin_declarar` si alguna declara `no_cubierto`, con las descripciones de contracts §3, `comparacion` `"<="`, `umbral` 0 y `decide: true`. **Tests** en `informe_test.go`, con sesiones sintéticas escritas en `t.TempDir()`: `TestUmbralesDeLegalCore`, los siete casos de FR-087 —(1) una sesión de plazo sin la línea con la fecha cuenta y da `fallo` con su motivo; con ella no cuenta; (2) una de cobertura sin la declaración, o con una línea `⚠ PLAZO:`, cuenta y da `fallo`; (3) con la declaración y sin esa línea no cuenta aunque lleve una fecha suelta en el texto; (4) una de plazo con la línea y sin un aviso esperado no cuenta en el umbral y no cumple lo esperado de su serie; (5) una eval que no es de plazo ni de cobertura no entra en ningún `total`; (6) una sesión sin medir no cuenta y da su motivo de la ejecución; (7) una sesión de Haiku no entra— y US6.1 y US6.2 (1 de 9 → `fallo` nombrando la sesión; 0 de 6 → se cumple y ningún motivo); y `TestUmbralesDeLasTresSkills`, con las evals del repositorio y el plan del job: en `legal-core`, exactamente `sin_activar`, `plazo_sin_fecha` y `cobertura_sin_declarar` del modo orden y los mismos del modo herramienta, los seis con `decide: true`, totales 21, 9 y 6; trece con nueve que deciden en `boe-legislacion` y quince con once en `jurisprudencia`, en su orden y sin cambiar ninguno. Los tests de hoy que esperan `umbrales` de `legal-core` vacío pasan a los seis (sustituye H23.1 FR 007). Verificación: con la comparación de `plazo_sin_fecha` como «ninguna sesión pasa» un momento, y con su `decide` en falso, `TestUmbralesDeLegalCore` o `TestUmbralesDeLasTresSkills` fallan (se restauran; T014 los registra con los demás); `make ci` en verde — FR-053 a FR-056, FR-071, FR-072, FR-087, FR-088; SC-001, SC-011, SC-012. Rutas: internal/evals/informe.go, internal/evals/umbrales.go

**Checkpoint**: el job de evals de `legal-core` sabe salir en rojo por una fecha que falta o por el plazo del común.

---

## Fase 6: US4 y US1 — `legal-core` v1 (P1)

**Objetivo**: la skill comprueba la cobertura antes de preguntar nada, sitúa el plazo con `territorio`, lo pide a
`plazos`, lee el precepto y responde con la fecha en su forma fija; sus dos referencias, generadas.

**Prueba independiente**: quickstart.md §13 y `make skills-check`; su efecto, el job de cierre (SC-001).

- [ ] T013 [US4] La skill `legal-core` v1 y sus dos referencias generadas (contracts/skill-legal-core.md §1 a §4; contracts/SKILL-legal-core.prototipo.md; research D21, V13, V14, V22, M4; plan.md paso 11). **`internal/skills/referencias.go`**: dos filas nuevas en la tabla de generadores —`recursos_y_plazos.md` desde `data/plazos.yaml` con `plazos.LeerTipos`, título «Recursos y plazos», columnas Tipo · Código · Duración · Cómputo · Precepto, una fila por tipo y por unidad del genérico; y `cobertura_por_materia.md` desde `data/cobertura.yaml` con `skills.LeerCobertura`, título «Cobertura por materia», columnas Procedimiento · Lo regula aparte · Norma que lo rige, «—» sin norma—, cada una con la cabecera `<!-- generado desde data/<fichero>, no editar -->` y sin texto de ninguna norma; `TestRenderizarRecursosYPlazos` y `TestRenderizarCobertura` (`referencias_test.go`). **`skills/legal-core/SKILL.md`**: el borrador de contracts, con su frontmatter (`description` de menos de 1 024 caracteres que gana las preguntas de plazos y recursos y de lo regulado aparte y sigue ganando las del territorio, FR-045; `kitlegal-applets: territorio plazos`; `kitlegal-referencias` con `recursos_y_plazos` y `cobertura_por_materia`), los pasos 7 a 11 —cobertura antes de preguntar nada; situar con `territorio`, preguntando el municipio que falte y nunca el de quien usa el kit; pedir el plazo a `plazos`; leer cada precepto de `data.preceptos` con `boe articulo` por medio de `boe-legislacion`; responder con la línea del plazo, los festivos y la línea de cada aviso, y la disposición solo con el aviso de cobertura que la trae (FR-040, FR-041)—, el acto presunto sin fecha (FR-043), el art. 24.6 LTAIBG y el plazo máximo del silencio (FR-044), las reglas de v0 con la 7 ampliada y la 8 y la 9 nuevas (FR-047), y las formas fijas `⚠ PLAZO:`, las dos `⚠ NO CUBIERTO:` y las cabezas de los cuatro avisos en bloques `text` (FR-048); la tabla de comandos y las referencias regeneradas con `make skills-sync`, y si la tabla regenerada difiere de la del borrador, manda la regenerada. **`skills/legal-core/references/recursos_y_plazos.md`** y **`cobertura_por_materia.md`**: las que escribe `make skills-sync`, sin editar. Verificación: `make skills-check` sin diferencias; `wc -l` de `SKILL.md` por debajo de 300; siete líneas que empiezan por `⚠ ` en `SKILL.md` (quickstart.md §13); con una referencia editada un momento, `make skills-check` falla (se regenera); `make ci` en verde. Su efecto lo mide el job de cierre con los umbrales de T012 — FR-032, FR-040 a FR-048; SC-010. Rutas: internal/skills/referencias.go, skills/legal-core/SKILL.md, skills/legal-core/references/recursos_y_plazos.md, skills/legal-core/references/cobertura_por_materia.md

**Checkpoint**: la entrega completa —binario, datos, evals, umbrales y skill— está en el árbol.

---

## Fase 7: Pulido — mutantes, documentos y cierre

- [ ] T014 Los mutantes de FR-089, aplicados uno cada vez en una copia de la cabeza fuera del árbol, y su registro (quickstart.md §11 y §12; plan.md paso 12 y fila de FR-089 de «Controles de umbral»). Sobre una copia de `HEAD` exportada con `git archive` bajo `$TMPDIR` y ejecutada con `go -C`, sin tocar el árbol: (1) el cómputo de la LPAC con un solo municipio (sin el art. 30.6); (2) sin la prórroga del art. 30.5; (3) agosto contado en el contencioso; (4) una corrección del BOCM sin aplicar (las disposiciones de un año reducidas a la principal en la derivación); (5) la comparación de `plazo_sin_fecha` en «ninguna sesión pasa» en lugar de «cada sesión»; (6) la de `cobertura_sin_declarar`, igual; (7) `decide` falso en `plazo_sin_fecha`; (8) `decide` falso en `cobertura_sin_declarar`. Con cada uno, por separado, se ejecutan `TestCasosDePlazos`, `TestCalcular`, `TestPropiedadesDelComputo`, `TestFestivosLocalesDerivados`, `TestUmbralesDeLegalCore` y `TestUmbralesDeLasTresSkills`, y al menos uno falla. **`specs/023-h9-internal-core-plazos/mutantes.md`**: por mutante, el cambio en una línea, la orden ejecutada y cada test que falla, con su subtest y su mensaje, tal como se vio en esta sesión. Ningún mutante queda en el árbol y la copia se borra. Si un mutante no hace fallar ningún test, se refuerza en esta misma tarea el test del control que debía verlo, en su fichero declarado entre estas rutas (`TestCasosDePlazos` en `internal/app/plazos_casos_test.go`, `TestCalcular` en `internal/core/plazos/calcular_test.go`, `TestPropiedadesDelComputo` en `internal/core/plazos/propiedades_test.go`, `TestFestivosLocalesDerivados` en `internal/source/bocm/derivar_test.go`, `TestUmbralesDeLegalCore` y `TestUmbralesDeLasTresSkills` en `internal/evals/informe_test.go`), sin tocar código de producto ni quitar ningún caso; se anota en `mutantes.md` y se repiten el mutante y `make ci`. Verificación: los ocho mutantes con al menos un test en rojo en `mutantes.md`; `make ci` en verde; `git status` sin nada fuera de estas rutas — FR-089; SC-008. Rutas: specs/023-h9-internal-core-plazos/mutantes.md, internal/app/plazos_casos_test.go, internal/core/plazos/calcular_test.go, internal/core/plazos/propiedades_test.go, internal/source/bocm/derivar_test.go, internal/evals/informe_test.go

- [ ] T015 Lo que los documentos dicen de `plazos`, de los festivos, de `legal-core`, de sus evals y de sus umbrales, con cada cifra, fichero, test y umbral sacados del árbol y no de memoria (research V28; plan.md paso 13). **`CHANGELOG.md`**, dentro de `## [Unreleased]` y en ninguna versión publicada: el applet `plazos calcular` y su herramienta `plazos_calcular`, con sus tipos y sus códigos; `data/festivos/` con sus dos fuentes, sus años y su cobertura (locales solo de la Comunidad de Madrid, 2025 y 2026); los tipos y la lista de cobertura por materia; `legal-core` v1; el formato de eval con `plazo`, `no_cubierto` y `calcular`; las cinco evals; y los dos umbrales nuevos —qué cuentan, sobre qué sesiones, con 0 y que deciden— y que `umbrales` de `legal-core` deja de estar vacío (FR-060, FR-090). **`README.md`** (líneas 24 a 37 y la lista de herramientas si la hay): lo que `legal-core` y `plazos` hacen para quien pregunta, sin prometer ninguna release. **`CONTRIBUTING.md`**, en las líneas de research V28: la lista de esquemas, el verbo con su herramienta, la tabla del formato de eval con los esperados nuevos y las reglas del conjunto de `legal-core`, su concurrencia de 2 (#157), sus nueve evals y sus sesiones (48, 48 y 6 más la prueba de red, contadas con el plan del job), sus seis umbrales que deciden y lo de la línea 1210 que quede falso. **`internal/evals/doc.go`**: el comentario del paquete donde dice qué umbrales tiene una skill. **`docs/WORKFLOW.md`**: solo si un `grep` de `legal-core`, `plazos` y `umbrales` encuentra una frase que el hito deje falsa (V28 dice que hoy no nombra `legal-core`). Solo lo falso o incompleto; nada de lo que sigue siendo cierto se reescribe y ningún test ata este texto al código (FR-062). Verificación: cada test, fichero, umbral y cifra nombrados existen en el árbol (`grep`); `make ci` en verde — FR-060, FR-061, FR-062, FR-090; SC-013. Rutas: CHANGELOG.md, README.md, CONTRIBUTING.md, internal/evals/doc.go, docs/WORKFLOW.md

- [ ] T016 Cierre de la Definition of Done, sin código de producto (`ROADMAP.md` §1; quickstart.md §0 a §13 y §16; plan.md, «Obligaciones para tasks.md»): specs/023-h9-internal-core-plazos/cierre.md con (1) la lista, sacada de `git diff --name-status main`, de lo creado o modificado, y la comprobación de que no cambia nada de «Lo que no cambia» de la batería —`.github/workflows/`, las skills y evals de `boe-legislacion` y `jurisprudencia`, `internal/mcp`, `internal/empaquetado`, `docs/SOURCES.md`, `scripts/verify-sources.sh`, `docs/ROADMAP.md`, `docs/ADR/`, `docs/USO.md`, `docs/HISTORIAL.md`, `CLAUDE.md`, `web/`, las cuatro evals de hoy de `legal-core`, los manifiestos de H4 y H5 y los specs de otros hitos— (FR-028, FR-058, FR-063, FR-070), de que los dos ficheros de `data/festivos/` son los que escriben sus tests de derivación (ejecutados sin bandera) y de que ningún municipio aparece en código Go que no sea de test, en `data/plazos.yaml`, en `data/cobertura.yaml` ni en `SKILL.md` (FR-073); (2) la cobertura sobre `coverage.out` y `coverage-integration.out` de `make ci` con `go tool cover -func` —global ≥ 70 % e `internal/core/plazos` ≥ 85 % (Definition of Done §1.9)—, con la cifra de cada una y, si alguna queda por debajo, el hueco (paquete, fichero y funciones sin cubrir) para la revisión final; esta tarea no escribe tests: la de `internal/core/plazos` ya la hace cumplir la verificación de T005, la última tarea que escribe código en el paquete; (3) el resultado de quickstart.md §0 a §13 y §16, orden a orden, con lo esperado de cada una —los guiones `h9-plazos-` de §2, que el workflow activa tras el bucle, ejercidos con copias momentáneas de prefijo `zz-`, que se borran—, y la constancia de que §14 lo hace el workflow y §15 queda para después del run; (4) la tabla de plan.md, «Controles de umbral», con el test de cada fila `ci:` encontrado como `^func <Test>\(` en su ruta y verde, y cada fila `evals:` con el umbral que publica T012; y (5) la nota de que el informe final lleva la de FR-091 (cinco applets; revisar si `internal/cli` pide refactor, no extracción), que no es trabajo de ninguna tarea. Verificación: `make ci` en verde y `git status` sin nada fuera de `cierre.md` — FR-028, FR-058, FR-063, FR-070, FR-073, FR-080; SC-004 a SC-012. Rutas: specs/023-h9-internal-core-plazos/cierre.md

---

## Dependencias y orden de ejecución

El bucle ejecuta las tareas en su orden; ninguna depende de una posterior.

| Tarea | Depende de | Por qué |
|---|---|---|
| T001 | — | la suite se escribe antes que el código |
| T002 | — | los manifiestos; tras su commit graba `grabar_datos` |
| T003 | — | el dominio no lee ningún fichero |
| T004 | T003 | calcula sobre el `Calendario` y los `Tipo` |
| T005 | T002, T003, T004 | lee lo grabado del BOE y `data/plazos.yaml` con `LeerTipos`; ajusta el cómputo de T004 y sus tests si lo grabado cambia una regla, y mide la cobertura del paquete |
| T006 | — | solo `territorio` |
| T007 | T002, T003, T006 | lo grabado del BOE, `EscribirFestivosNacionales`, `Comunidades` y `FormaCompacta` |
| T008 | T002, T003, T006 | lo grabado del BOCM, `EscribirFestivosLocales`, `MunicipiosDe` y `FormaCompacta` |
| T009 | T004, T005, T007, T008 | el cómputo, `data/plazos.yaml` y `data/festivos/` empotrados |
| T010 | T008, T009 | el applet; la derivación con la principal sola; `LeerCobertura` (T005) |
| T011 | T004, T005, T009, T010 | `CodigosDeAviso`; la unión de grabaciones; el comando `calcular`; las fechas de la tabla |
| T012 | T011 | las claves `plazo` y `no_cubierto` y su juicio |
| T013 | T005, T009, T012 | `LeerTipos`, `LeerCobertura`, la tabla de `--describe`; las formas fijas que miden los umbrales |
| T014 | T004, T008, T010, T012 | los tests que deben fallar con cada mutante |
| T015 | T009 a T013 | lo que documenta ya está en el árbol |
| T016 | todas | el cierre sobre el diff entero |

### Historias

- **US1** (P1): T009, T010 y T013, sobre T002 a T008. **US2** (P1): con US1 (avisos y `cobertura` de T004, casos de
  T010). **US3** (P1): T006 a T008. **US4** (P1): T005, T011 y T013. **US5** (P2): T004 y T010. **US6** (P1): T011 y
  T012.

### Oportunidades de paralelismo

Ninguna tarea lleva `[P]`: el workflow las ejecuta en secuencia, con un guardián por tarea. Fuera del bucle, T006 es
independiente de T003 a T005, y T007 y T008 lo son entre sí.

---

## Trazabilidad

### Requisitos → tareas

| Requisito | Tareas |
|---|---|
| FR-001 | T001, T004, T009 |
| FR-002 a FR-006 | T001, T004, T010 |
| FR-007 | T001, T004, T009, T010 |
| FR-008 | T003, T005, T010 |
| FR-009 | T004, T010 |
| FR-010 | T001, T004, T005, T009, T010 |
| FR-011 | T003, T004, T007, T010 |
| FR-012 | T001, T004, T009 |
| FR-013 | T001, T009 |
| FR-020 | T002, T003, T007, T008 |
| FR-021 a FR-023 | T008 |
| FR-024 | T004, T008, T010 |
| FR-025 | T003, T007 |
| FR-026 | T001, T003, T004, T009, T010 |
| FR-027 | T002 |
| FR-028 | T001, T009, T016 |
| FR-030 | T002, T003, T005 |
| FR-031 | T002, T005 |
| FR-032 | T013 |
| FR-040 a FR-047 | T013 |
| FR-048 | T004, T011, T013 |
| FR-050 | T011 |
| FR-051 | T011 |
| FR-052 | T002, T005, T011 |
| FR-053 a FR-056 | T012 |
| FR-057 | T011 |
| FR-058 | T011, T016 |
| FR-060 a FR-062 | T015 |
| FR-063 | T016 |
| FR-070 | T016 |
| FR-071 | T012 (y la batería) |
| FR-072 | T012 (y la batería) |
| FR-073 | T016 (y la batería) |
| FR-080 | todas; T016 |
| FR-081 | T004 |
| FR-082 | T010 |
| FR-083 | T007, T008 |
| FR-084 | T008 |
| FR-085 | T001, T009 |
| FR-086 | plan.md, «Controles de umbral»; T012 publica sus umbrales |
| FR-087 | T012 |
| FR-088 | T005, T011, T012 |
| FR-089 | T014 |
| FR-090 | T015; el job, el workflow |
| FR-091 | el informe final del workflow; T016 lo constata |
| SC-001, SC-002 | T011, T012, T013; los mide el job de cierre y el informe final |
| SC-003 | una persona, después del run; ninguna tarea |
| SC-004 | T004 |
| SC-005 | T010 |
| SC-006 | T007, T008 |
| SC-007 | T008 |
| SC-008 | T014 |
| SC-009 | T001, T009 |
| SC-010 | T013 |
| SC-011, SC-012 | T012 |
| SC-013 | T015; la revisión final |

### Controles de umbral (plan.md) → tarea que construye el control

| Fila de plan.md | Tarea | Cómo se ve fallar |
|---|---|---|
| `plazo_sin_fecha` (orden, herramienta) | T012 (formas fijas y evals de T011) | `TestUmbralesDeLegalCore`, casos 1 y 4; mutantes 5 y 7 de T014 |
| `cobertura_sin_declarar` (orden, herramienta) | T012 (T011) | `TestUmbralesDeLegalCore`, casos 2 y 3; mutantes 6 y 8 de T014 |
| `sin_activar` de `legal-core` (orden, herramienta) | T012 (lo publica al tener umbrales de sus respuestas) | `TestUmbralesDeLasTresSkills` con otro `decide` u otro orden |
| `boe-legislacion` y `jurisprudencia` siguen como decidían | ninguna los cambia; T012 los fija | `TestUmbralesDeLasTresSkills` con otro elemento |
| `TestDefinicionDelJob` (FR-050) | T011 | definición sintética con el tope bajo el peor caso |
| `TestUmbralesDeLegalCore` (FR-055, FR-087) | T012 | sus siete casos; mutantes 5 a 8 |
| `TestUmbralesDeLasTresSkills` (FR-056) | T012 | `decide` en falso; mutantes 7 y 8 |
| `TestPropiedadesDelComputo` (FR-081) | T004 | sin la prórroga; mutante 2 |
| `TestCasosDePlazos` (FR-082) | T010 | sin el art. 30.6, con agosto, sin la prórroga; mutantes 1 a 3 |
| `TestFestivoDeUnaParte` (FR-011) | T010 | sin festivo de una parte en lo empotrado, o contado como inhábil |
| `TestFestivosNacionalesDerivados` (FR-020, FR-083) | T007 | una línea cambiada; una fila sin `source` |
| `TestFestivosLocalesDerivados` (FR-021, FR-022, FR-083) | T008 | disposiciones sin ordenar; mutante 4 |
| `TestCotejoConElCatalogo` (FR-084) | T008 | sin la regla de prefijo: diferencias de más |
| `TestPreceptosDeLosPlazos` (FR-010, FR-030, FR-031) | T005 | duración cambiada, bloque sin grabar, palabra clave cambiada |
| `TestEvalsDelRepositorio` (FR-052, FR-051) | T011 (unión de T005) | la eval 07 sin `no_cubierto` |
| `TestFormatoDePlazos` (FR-051) | T011 | un código de aviso que no es de `plazos.CodigosDeAviso` |
| `TestArquitectura` (FR-013) | T009 | `internal/httpx` importado desde `plazos.go` |
| `TestAppletPlazos` y los guiones `h9-plazos-sin-efectos` y `h9-plazos-calcular` (FR-013) | T009, T001 | una operación de grafo; mutante de T001 que escribe en la caché |
| guiones `h9-plazos-codigos` y `h9-plazos-cobertura` (FR-012, FR-085) | T001 (activos tras el bucle), T009 | mutantes de T001 |
| `TestHerramientasDelServidor` y `h9-plazos-herramienta` (FR-001, FR-085) | T009, T001 | la herramienta sin registrar |
| `skills-check` y `TestSkillsDelRepositorio` (FR-046, FR-032) | T013 | una referencia editada; 300 líneas o más |
| FR-089 | T014 | los ocho mutantes |

### Definition of Done (`ROADMAP.md` §1) → tareas

| Punto | Aplica | Tareas |
|---|---|---|
| 1 · `make ci` en verde | Sí | Todas; T016 |
| 2 · Tests offline; fixtures si toca red | Sí: tres fuentes grabadas por `grabar_datos` | T002 (manifiestos); T005, T007, T008 (leen lo grabado); T009 (golden) |
| 3 · Sin `net/http`, `os.Exit`, `fmt.Print*` fuera de lo autorizado | Sí | T009 (`TestArquitectura`); T003 y T004 sin `os` ni `io` |
| 4 · Esquemas y `schema-check` | Sí | T009 (`plazo.json`); T005 (`cobertura.yaml.json`); T011 (`eval.yaml.json`) |
| 5 · Errores tipados y códigos estables | Sí | T004, T009 (2, 3 y 1, nunca 4, 5 ni 6) |
| 6 · e2e y `CHANGELOG.md` | Sí | T001, T009; T015 |
| 7 · ADR | No: los festivos los fija el ADR 0039; la cobertura, el 0027; los umbrales, el 0038 (plan.md, «Summary») | — |
| 8 · `SOURCES.md` y `verify-sources.sh` | No: las tres fuentes tienen fila revisada en `main` (research V1) y FR-028 deja `verify-sources.sh` sin vigilarlas | T016 lo constata |
| 9 · Cobertura | Sí | T004 y T005 (≥ 85 % en `internal/core/plazos`, en su verificación); T016 (las dos cifras y cualquier hueco, en `cierre.md`) |
| 10 · Skill: evals antes, `SKILL.md` < 300, sin drift | Sí | T011 (evals, antes que la skill); T013 |
| 11 · Dimensión territorial | Sí | T001 (`plazos-cobertura`: Leganés y Tordesillas); T009; T016 (FR-073) |
| 12 · Operaciones de grafo | No emite ninguna (FR-013) | T009 (`Resultado.Grafo` vacío) |
| 13 · Recursos y plazos: común y regulado aparte | Sí | T005, T011 (evals 05 y 07, 08), T012, T013 |

---

## Estrategia de implementación

- **De dentro afuera** (plan.md): la suite; lo que se graba; el dominio; los datos que fija el plan; lo nuevo de
  `territorio`; los dos boletines; el applet; la tabla de casos; el formato de eval y las evals; los umbrales; la skill;
  los mutantes; los documentos; y el cierre.
- **Lo mínimo que ya da valor** es la fase 4 (T001 a T010): `kitlegal plazos calcular` da el 13 de octubre de 2026 en el
  caso (a), con el precepto y los festivos, desde la terminal y como herramienta. No es una entrega parcial: el run
  ejecuta las dieciséis, y SC-001 pide el cierre entero, con la skill de T013 cumpliendo los umbrales de T012.
- **Lo que el run no puede medir** (research S6): qué dará el job de cierre con `legal-core` v1. Si el cierre sale en
  rojo por un umbral de `legal-core`, se repara lo que entrega el hito, sin cambiar el umbral, su `decide`, su `total`
  ni las evals (FR-071); si sale en rojo por `boe-legislacion` o `jurisprudencia`, queda en el informe final.
- **Si lo grabado contradice al plan** (research S1 a S5, S8): manda lo grabado; T005, T007, T008 y T010 corrigen el
  artefacto del feature que lo recordaba y lo anotan en `gates/supuestos.md`; los guiones congelados de T001 solo
  dependen de días ciertos (D27). Si el bloque `a1` del texto refundido de Tráfico no existe, `grabar_datos` se
  detiene tras T002 con su dossier (S8).
- **Después del run** (SC-003; quickstart.md §15): con la release publicada, la prueba en la app de escritorio, y la
  viñeta de H9 de `CLAUDE.md` al fusionar (spec, «Fuera de alcance»).

---

## Comprobación contra la rúbrica de `juez_tasks` y `precheck.sh tasks`

- `precheck.sh tasks`: el guion no se pudo ejecutar en esta sesión (el sandbox pide aprobación para lanzarlo); sus
  comprobaciones de la fase tasks se reprodujeron con `grep` y sus mismos patrones sobre las líneas de tarea, sin
  ningún hallazgo: dieciséis líneas con el formato `- [ ] Tnnn`, ids únicos y correlativos (T001 a T016); ninguna sin
  rutas; las que nombran el directorio de esquemas o uno de fixtures llevan `[datos]` (T002, T005, T007, T008, T009 y
  T011, y ninguna otra línea lleva esa etiqueta); ninguna nombra la carpeta de la evidencia por su ruta, ni la variable
  de grabación, ni una acción de plataforma, ni pide nada a una persona a mitad; una sola `[aceptacion]`, la primera;
  `checklists/requirements.md` sin ítems sin marcar. El único defecto que puede dar antes del análisis es que falta
  `gates/analyze.md`, que escribe el paso siguiente.
- a · `analisis_critico`: lo escribe el paso de análisis después de este fichero.
- b · `trazabilidad`: cada tarea cita sus FR y SC; «Requisitos → tareas» cubre FR-001 a FR-091 y SC-001 a SC-013; y
  «Controles de umbral» da la tarea y cómo se ve fallar cada fila de plan.md.
- c · `rebanadas_verdes`: T002 a T013 traen sus tests y su implementación; las excepciones (T001, T014 a T016) están
  declaradas arriba; el orden es el del plan, con los pasos 3 y 6 partidos, y la tabla de dependencias no tiene ninguna
  hacia delante.
- d · `rutas_declaradas`: tras «Rutas:», ficheros concretos; los únicos directorios son el de la suite (T001) y el de
  los golden del applet (T009), que nacen con la tarea; ninguna ruta genérica.
- e · `datos_separados`: las tareas que tocan el directorio de esquemas o uno de fixtures llevan `[datos]`, y su código
  va con ellas por la razón declarada arriba (plan.md, *Complexity Tracking*); lo externo sale de los tres manifiestos
  de T002 y de `grabar_datos`, nunca de la red, de una persona ni de la memoria; `data/festivos/` lo escribe el código.
- f · `dod`: tabla «Definition of Done → tareas»; `CHANGELOG.md` en T015; sin ADR ni `SOURCES.md`, con su motivo;
  esquemas en T005, T009 y T011; fixtures en T002 y T009.
- g · `checklist_veraz`: `checklists/requirements.md` no cambia.
- h · `aceptacion_primero`: T001 es la tarea `[aceptacion]`; sus guiones cubren US1.1, US2.1 a US2.3, US5.1 y US5.2;
  lo que no tiene guion (US3, US4, US6, US1.2, US1.3) se acepta como dice la cabecera, como fija el plan.
- i · `autonomia`: ninguna tarea exige a una persona, publica, ni mide en la plataforma.
