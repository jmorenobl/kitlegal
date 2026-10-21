# Tasks: H26.1 · Las evals sirven lo que la skill lleva a leer y hacen fallar a propósito la lectura que quieren probar

**Input**: `specs/025-h26-1-las-evals-sirven/` (spec.md, plan.md, research.md, data-model.md, contracts/, quickstart.md)

**Prerrequisitos**: spec.md y plan.md leídos; `.specify/memory/constitution.md` 2.13.0 aplicada (umbral de
materialidad, criterio de uso, proporcionalidad y convergencia en «Gates»; controles de umbral del ADR 0029; ADR 0016,
0018, 0023, 0031, 0032, 0037, 0038 y 0043); `main` en `6bf2da3` (#179: ADR 0043 aceptado, workflow 2.9.4). **Ninguna
tarea crea el esqueleto ni los gates: `make ci` existe y pasa**, y cada tarea lo deja en verde al terminar.

**Aceptación**: **no aplica** (plan.md, «Aceptación e2e», y Complexity Tracking; research D15). El hito no cambia el
binario: `cmd/kitlegal` no enlaza `internal/evals` (research V5), `internal/render` no se toca (D8) y FR-028 prohíbe
cambiar el código con que termina. Ninguna tarea `[aceptacion]`. La aceptación es la del job de evals de cierre
(SC-011 a SC-020), que lanza el workflow tras la revisión final; en `make ci` la fijan, sin modelo, los tests de la
tabla de plan.md, «Aceptación e2e», que construyen las tareas de abajo.

**Tests**: obligatorios (constitución §III; plan.md, «Controles mecánicos»). Cada tarea trae sus tests y la
implementación que los hace pasar en el mismo diff; los nombres de test y sus ficheros son los de plan.md, «Tests
nuevos» y «Tests existentes que cambian lo que esperan», porque el informe final busca cada control `ci:<ruta>:<Test>`
como `^func <Test>\(` en esa ruta.

**Rebanadas verticales y excepciones declaradas.** Son rebanada completa —test e implementación en el mismo diff,
`make ci` en verde por sí solas— T003 a T010, T012 y T013. Las demás, cada una por su razón:

- **T001** y **T002** son las dos tareas `[datos]` del plan (pasos 1 y 2), sin otro trabajo: el esquema y el
  manifiesto. Los tests que las ejercen ya existen y siguen en verde con ellas (`TestEsquemaDeEval` y
  `TestEvalsDelRepositorio` validan cada eval del repositorio contra el esquema; `TestManifiestoDeGrabaciones/repositorio`
  lee el manifiesto real). Los rechazos del esquema los fija T003, la tarea siguiente, a través de `LeerEval`, que es
  quien valida contra él (research V8).
- **T011** es el texto de una skill: lo fija `make skills-check` (`TestSkillsDelRepositorio`, que ya existe), y lo que
  dice lo juzga la revisión final (capa 2); las evals que lo miden entran antes, en T008 (Definition of Done §1.10).
- **T014** es documentación; su verificación es que cada cifra, fichero, test y umbral que nombra existe en el árbol,
  y `make ci`.
- **T015** son los mutantes sobre el código que dejan T004, T008, T007 y T013, en un árbol aparte, y su registro; no
  cambia nada bajo `internal/` (research D16).
- **T016** es el cierre de la Definition of Done, sin código de producto.

**Tareas `[datos]`**: T001 (`schemas/eval.yaml.json`) y T002 (`internal/evals/testdata/grabaciones.json`), cada una
sola, sin código. Tras el commit de T002, el paso `grabar_datos` del workflow ejecuta, sin modelo, `TestGrabarEvals`
(`internal/evals/grabacion_test.go`, etiqueta `grabacion`) con la fila de `boe.legislacion-consolidada` revisada en
`docs/SOURCES.md` de `main` (2026-09-13), y deja las cuatro respuestas nuevas en
`internal/evals/testdata/boe.legislacion-consolidada/`; T003 y las siguientes ya parten de ellas. Ninguna otra tarea
toca un directorio de fixtures ni de esquemas: las sesiones sintéticas se escriben en `t.TempDir()`, y ningún fichero
de `data/` cambia.

**Lo que este fichero añade al plan, comprobado en el árbol en esta sesión** (lectura y `grep`, sin ejecutar ningún
test):

- **Los ficheros de los tests que el plan nombra y ya existen**: `TestEvalsDelRepositorio` (`conjunto_test.go:1993`),
  `TestConsultasNecesarias` (`consultas_test.go:42`), `TestDefinicionDelJob` (`definicion_test.go:196`; el literal
  `peorCasoDeLegalCore`, línea 1896; el caso de 288 minutos, líneas 1047 y 1122), `TestLeerEval` y `TestEsquemaDeEval`
  (`formato_test.go:220` y `:1898`), `TestInforme`, `TestInformeEnDosModos`, `TestInformeConElJuez`,
  `TestUmbralesDeJurisprudencia`, `TestUmbralesDeLegalCore`, `TestUmbralesDeLasTresSkills` y
  `TestUmbralesDeBoeLegislacion` (`informe_test.go`), `TestJuzgar` (`juzgar_test.go:293`), `TestBaseDeLaReconstruccion`
  (`medida_test.go:3077`), `TestJuzgarPlazos` (`plazos_test.go:51`), `TestLeerLasLlamadas` y `TestTextosDeLaSesion`
  (`sesion_test.go:283` y `:800`), `TestUmbralesDelInforme` y `TestUmbralesConJuezYConSentencias` (`umbrales_test.go`),
  `TestManifiestoDeGrabaciones` y `TestGrabacionesSinSolape` (`grabaciones_test.go`), `TestSkillsDelRepositorio`
  (`internal/app/skills_test.go:83`) y `TestTuberiaCerrada` (`internal/app/tuberia_unix_test.go`, con
  `mensajeDeSalidaRota`, línea 23). No existen todavía, y los crean las tareas: `TestCacheDeLasSesiones`,
  `TestLoQueLlevaALeerSinGrabar`, `TestFormatoDeLecturas`, `TestJuzgarLasLecturasSinServir`,
  `TestJuzgarElMotivoDeLasFallidas`, `TestLineaDeLaSalidaNoEscrita` y `TestUmbralesDeLasLecturas`.
- **`InvocacionFallida.Codigo` es hoy `int`** (`juzgar.go:351`) y lo usan `juzgar.go`, `informe.go`, `informe_test.go`,
  `juzgar_test.go` y `sondeo_test.go` (que pone a `nil` las listas de la sesión, líneas 578-579): T006 los declara.
- **`TasaDelInforme` solo se construye en `informe.go`** (línea 1937): el sondeo no arma series propias con ella, así
  que `sondeo.go` no cambia; los tests que leen informes escritos con `EscribirInforme` (`sesiones_test.go`,
  `sondeo_test.go`, `serie_test.go`, `cierre_test.go`, `ejecucion_test.go`) se declaran en las tareas que cambian el
  JSON del informe (T007 y T010), por si alguno compara una serie o la lista de umbrales entera.
- **`ConsultasNecesarias` la usan `preparar.go:329` y `medida.go:1071`**, y la nombran `conjunto_test.go`,
  `consultas_test.go`, `job_test.go`, `medida_test.go` y `preparar_test.go`: T004 y T009 los declaran.
- **El conjunto de `legal-core` no tiene máximo** (`minimoDeEvalsDeLegalCore = 3`, `conjunto.go:245`); el de
  `boe-legislacion` sí (`maximoDeEvals = 23`, `conjunto.go:229`), y no cambia: T009 añade `lecturas` a tres evals que
  ya existen. Cuentan las evals o las series de `legal-core` `conjunto_test.go` (`evalsDeLegalCoreConLaDeVigencia`),
  `definicion_test.go` y `plan_test.go` (`probarElPlanDelRepositorio`, líneas 290-310): T008 los declara.
- **La versión de `legal-core`** solo se nombra en `CHANGELOG.md:28` y en el ADR 0043; `skills/legal-core/SKILL.md`
  tiene 299 líneas (`wc -l`), el máximo de `internal/skills/sincronia.go` (research V9).
- **Los lectores de `data/`** existen con el nombre del plan: `plazos.LeerTipos` y `plazos.Unidades`
  (`internal/core/plazos/tipos.go:151` y `:70`), `skills.LeerCobertura` (`internal/skills/cobertura.go:100`) y
  `skills.LeerNormas` (`internal/skills/normas.go:131`). `render.Nuevo` (`internal/render/render.go:44`) y el error de la
  escritura fallida (`render.go:99`, `render: no se pudo escribir en %s: %w`) también.
- **Los métodos sin exportar de data-model §1** (`LecturaSinServir.consulta`, `Eval.sinServir`,
  `Eval.comandoSinServir`) entran con su primer uso —T005 los dos primeros, T006 el tercero— y no en T003, para que
  ningún método quede sin llamar en el código de producto de una tarea.

**Lo que no se ha medido en esta sesión.** No se ha ejecutado ningún test ni `make ci`. Las cifras de las tareas (72
consultas de `data/`, 17 bloques y 19 normas, 19 525 s, 18 y 19 umbrales) son las de research.md (V1, V10) y los
contratos; las tareas las vuelven a medir con sus tests.

**Modo**: desatendido (ADR 0018). El workflow `hito` ejecuta y verifica las tareas **una a una**, con guardián de diff:
cada tarea declara en su propia línea, tras «Rutas:», **todas** las rutas que crea o modifica, y solo toca esas (más
`go.mod`, `go.sum`, `CHANGELOG.md`, el directorio del feature y el `x_test.go` de cada `x.go` declarado). Los demás
nombres que aparecen en una línea son contexto: referencias a los documentos del feature y ficheros que la tarea solo
lee. El cierre en la plataforma —publicar la rama, abrirla a revisión y medir CI y el job de evals— lo hace el workflow
tras la revisión final; quickstart.md §10 queda para el workflow.

## Formato: `- [ ] Tnnn [etiqueta?] [Story?] Descripción — FR/SC. Rutas: …`

- **[P]**: ninguna tarea lo lleva; en el bucle del workflow todo va en secuencia.
- **[aceptacion]**: ninguna (Aceptación e2e: no aplica).
- **[datos]**: T001 y T002, solas.
- **[Story]**: US1 la caché sirve lo que la skill lleva a leer · US2 una eval hace fallar a propósito la lectura que
  quiere probar · US3 una eval sirve la exploración razonable que la provoca · US4 el informe cuenta las lecturas fuera
  de lo grabado · US5 el motivo de un 1 que no es de kitlegal · US6 `legal-core` v1.2.3. Las fases de datos, fundación
  y pulido no llevan historia; una tarea que sirve a varias lleva la principal y nombra las demás.

## Batería de verificación por tarea

- **`make ci` en primer plano** y la tarea marcada `[X]` en el mismo turno en que termina en verde: formato · lint
  (`depguard`, `forbidigo`, `gosec`, `misspell`, `revive`, `dupl`, `gocyclo`, `paralleltest`, `unused`…) · tests con
  `-race` y `-shuffle=on` · `test-integration` · `test-tiempos` · `vuln` · `schema-check` · `skills-check` ·
  `goreleaser-check` · secretos · módulos. Ningún registro de la verificación en la raíz del repositorio. Las salidas de
  `go test`, de `git diff` y de `make` se leen con `rtk proxy` o con sondas positivas, porque el proxy de la sesión las
  resume. `go test -run` parte el patrón por `/`: un `-run` por test, y se cuentan los `PASS`.
- **Sin red ni modelo** (ADR 0032): ninguna tarea usa la red salvo la de las herramientas de Go (`make vuln`), ni graba
  nada, ni escribe en la carpeta de la evidencia; ninguna ejecuta `claude`, `make evals`, `make evals-sondeo`,
  `make evals-medir-juez`, los guiones `scripts/evals*.sh`, `TestEjecucionDelJob`, `TestMedidaDelJuez`, `TestSondeo`
  ni `TestGrabar*`. Los tests de `make ci` usan cachés preparadas desde `UnionDeGrabaciones()` y sesiones sintéticas,
  que escriben en `t.TempDir()`.
- **Sin personas a mitad**: ninguna tarea la hace, la revisa ni la desbloquea una persona.
- **Nada de memoria** (constitución, «Criterio de decisión autónoma»): ningún identificador `BOE-A-…`, id de bloque,
  fecha ni texto de una norma se escribe de memoria: salen de `data/`, de lo grabado y de los contratos, que los
  midieron sobre él (research V1 a V3). El arnés no copia ningún identificador de `data/` (FR-002) ni el mensaje del
  binario (FR-025). Si lo grabado contradice a un contrato, manda lo grabado, se corrige el artefacto del feature y se
  anota en `specs/025-h26-1-las-evals-sirven/gates/supuestos.md`.
- **Ningún umbral se cumple cambiándolo** (ADR 0029): ninguna tarea cumple un umbral de plan.md, «Controles de umbral»,
  cambiando su valor, dejándolo en `decide: false`, sacando sesiones de su `total`, retirando su test o un caso de una
  tabla, ni cambiando la regla de 2 de 3, las repeticiones, los modelos, la concurrencia, los umbrales del juez o el
  tope del job (spec, «Fuera de alcance»). Cada control nuevo se ve en rojo antes de darlo por bueno, con el mutante de
  su fila aplicado en un árbol aparte bajo `$TMPDIR` (`git worktree add --detach`, nunca dentro del repositorio) y
  retirado al terminar; ninguno queda en el diff.
- **Lo que no cambia** (FR-028; plan.md, «Constraints»; quickstart.md §9): `cmd/`, `internal/app`, `internal/cli`,
  `internal/render`, `internal/source`, `internal/core`, `internal/httpx`, `internal/cache`, `internal/graph`,
  `internal/mcp`, `.github/`, `scripts/`, `.specify/`, `data/`, `testdata/` de la raíz, la carpeta de la evidencia,
  `docs/SOURCES.md`, `docs/ROADMAP.md`, `docs/ADR/`, `docs/WORKFLOW.md`, `CLAUDE.md`, `web/`, el `Makefile`,
  `skills/boe-legislacion/`, `skills/jurisprudencia/`, `evals/jurisprudencia/`, `evals/boe-legislacion/juez/`, las
  evals de `boe-legislacion` salvo la 14, la 15 y la 16, las de `legal-core` de hoy, `internal/evals/juez.go`,
  `sondeo.go`, `definicion.go`, las grabaciones de H4 y H5 y los specs de otros hitos.
- **Ningún test escribe fuera de `t.TempDir()`**; lo temporal de una tarea va bajo `$TMPDIR`, nunca bajo `/tmp`.
- **Sin atajos** (plan.md, «Comprobación contra la rúbrica», g): ningún `//nolint`, `t.Skip`, TODO ni error
  silenciado; ningún test se desactiva ni pierde lo que comprobaba: a los que cambian solo se les añade lo nuevo. Si
  `dupl`, `gocyclo`, `paralleltest` o `misspell` marcan algo, se reestructura o se renombra (una función por subtest
  desde una tabla; una palabra española que `misspell` marca va a `ignore-rules`).
- **Antes de tocar una función, una constante, un tipo o un campo que usan los tests**, `grep -n` de sus usos en todos
  los `_test.go` de `internal/evals`: lo que haya que alinear entra en la misma tarea, y sus ficheros están en sus
  rutas.
- **Proporcionalidad** (constitución, «Gates»; plan.md, «Lo que no se construye»): ninguna tarea añade `sin_servir` a
  la medida del juez (D4), un umbral que decida sobre `fuera_de_lo_grabado` o sobre los motivos, una constante
  exportada en `internal/render`, un juicio de la línea `⚠ REDACCIÓN A FECHA:` en la 11, `lecturas` o `sin_servir` de
  otro applet que `boe` o de búsquedas, ni cambia el comentario de los peores casos de `.github/workflows/evals.yml`.

---

## Fase 1: Datos — el esquema y el manifiesto

**Objetivo**: el esquema admite `lecturas` y `sin_servir`, y el manifiesto pide los cuatro bloques de `lecturas`, que
`grabar_datos` graba tras el commit de T002.

**Prueba independiente**: `TestEsquemaDeEval`, `TestEvalsDelRepositorio` y `TestManifiestoDeGrabaciones` en verde;
tras `grabar_datos`, las cuatro respuestas en el directorio de lo grabado de las evals.

- [X] T001 [datos] El esquema del formato de eval gana `lecturas` y `sin_servir`, y nada más (plan.md paso 1; contracts/formato-de-eval.md §1; data-model §1, regla 1; research D9, V8, V21). En `schemas/eval.yaml.json`: en `properties` del objeto raíz, que sigue con `additionalProperties: false`, `"lecturas": {"type": "array", "minItems": 1, "items": {"$ref": "#/$defs/lectura"}}` y `"sin_servir": {"type": "array", "minItems": 1, "items": {"$ref": "#/$defs/lectura-sin-servir"}}`; en `$defs`, `lectura` (objeto con `additionalProperties: false`, `required` `norma`, `bloque` y `motivo`; `norma` y `bloque` con `$ref` a las definiciones de norma y de bloque que ya usa el comando de bloque, y `motivo` `{"type": "string", "pattern": "\\S"}`) y `lectura-sin-servir` (`oneOf` de `{norma, bloque}` y `{norma, verbo}`, cada rama con `additionalProperties: false` y sus dos claves en `required`, `verbo` con `enum` `indice` y `metadatos`), tal como el contrato las escribe; si las definiciones de norma y de bloque se llaman de otro modo en el fichero, se usan las que hay y se anota en `gates/supuestos.md`. Ninguna otra regla del esquema cambia, ninguna eval los lleva todavía y ningún código cambia: el lector estricto de `LeerEval` sigue sin las claves, que llegan con su tipo en T003. **Verificación**: `TestEsquemaDeEval` y `TestEvalsDelRepositorio` en verde (cada eval del repositorio valida contra el esquema nuevo, que compila); `make ci` en verde — FR-009, FR-012; SC-001. Rutas: schemas/eval.yaml.json

- [X] T002 [datos] Las tres entradas del manifiesto de grabación de las evals, y nada más (plan.md paso 2, «Datos externos»; contracts/formato-de-eval.md §4; data-model §9; research V2, V11, D17). En `internal/evals/testdata/grabaciones.json`, al final de `normas`, las tres entradas del contrato tal cual —«régimen jurídico del sector público» con `titulo_empieza_por` `Ley 40/2015,` y `bloques` `a26` y `a27`; «impuesto renta personas físicas» con `Ley 35/2006,` y `a18`; «reguladora de las haciendas locales» con `Real Decreto Legislativo 2/2004,` y `a15`—, cada una con su `para`; `"fuente": "boe.legislacion-consolidada"` sigue igual. Ningún test ni código cambia: `TestGrabarEvals` ya sirve primero lo grabado de H4 y H5 y solo pide a la fuente lo que falta (research V11). Tras el commit, el paso `grabar_datos` del workflow graba los cuatro bloques en `internal/evals/testdata/boe.legislacion-consolidada/`; ninguna tarea los escribe a mano, toca la red ni escribe en la carpeta de la evidencia, y `docs/SOURCES.md` no cambia (D17). **Verificación**: `TestManifiestoDeGrabaciones` (su subtest del repositorio lee el manifiesto real sin error) y `TestGrabacionesSinSolape` en verde; `make ci` en verde — FR-010; SC-018. Rutas: internal/evals/testdata/grabaciones.json

---

## Fase 2: Fundación — el formato lee `lecturas` y `sin_servir`

**Objetivo**: `LeerEval` lee las dos claves con su tipo, y rechaza lo mal formado con el nombre del fichero y la
entrada. Lo necesitan US1 (la derivación suma su acumulador a las consultas de las evals), US2 y US3.

**Prueba independiente**: `TestFormatoDeLecturas` (quickstart.md §2).

- [X] T003 El formato de eval gana `lecturas` y `sin_servir`, con los cruces de `LeerEval` detrás del esquema (plan.md paso 3; contracts/formato-de-eval.md §1 y §2; data-model §1, reglas 1 a 3; research D9, V21). **`internal/evals/formato.go`**: `LecturaDeLaEval{Norma, Bloque, Motivo string}` (`norma`, `bloque`, `motivo`) y `LecturaSinServir{Norma, Bloque, Verbo string}` (`norma`, `bloque`, `verbo`, los dos últimos opcionales), con su comentario de documentación; `Eval` gana `Lecturas []LecturaDeLaEval` (`lecturas`) y `SinServir []LecturaSinServir` (`sin_servir`). `LeerEval`, detrás de la validación contra el esquema, rechaza como fichero mal formado, con un error que empieza por el nombre del fichero como los demás: una entrada de `lecturas` cuyo bloque —misma norma y mismo bloque— es un comando de bloque de la eval, con o sin `fecha` (`lecturas: <norma> <bloque>: ya es un comando esperado de la eval`), o una de sus `citas` (`… ya es una cita esperada de la eval`); y un bloque de `sin_servir` que es una de sus `citas` (`sin_servir: <norma> <bloque>: es una cita esperada de la eval`). Un bloque de `sin_servir` que es un comando de bloque de la misma eval se admite (FR-016). **`TestFormatoDeLecturas`** (`internal/evals/formato_test.go`), con evals sintéticas escritas en `t.TempDir()` y leídas con `LeerEval` y con `LeerConjunto`: lee una eval con `lecturas` y con `sin_servir` de las dos formas (un bloque y una consulta de norma), con sus campos; y rechaza, cada una con el nombre del fichero y la entrada (o la ruta del esquema: `/lecturas/0/motivo`, `/sin_servir/0`…), una lectura sin `motivo`, con el motivo vacío y con solo blancos; una lectura cuyo bloque ya es un comando de bloque, con `fecha` y sin ella; una cuyo bloque es una cita; un bloque de `sin_servir` que es una cita; un `verbo` que no es `indice` ni `metadatos`; una entrada con `bloque` y `verbo`; y `lecturas` o `sin_servir` vacías; y admite un bloque de `sin_servir` que es un comando de bloque. Con el lector de antes (sin los tipos), el caso que lee las dos claves sale en rojo. **Verificación**: `TestFormatoDeLecturas`, `TestLeerEval`, `TestEsquemaDeEval` y `TestEvalsDelRepositorio` en verde; `make ci` en verde — FR-008, FR-009, FR-012, FR-014; SC-004, SC-005. Rutas: internal/evals/formato.go, internal/evals/formato_test.go

---

## Fase 3: US1 — La caché de la sesión sirve lo que la skill lleva a leer (P1) 🎯 MVP

**Objetivo**: toda sesión, de cualquier skill, sirve cada bloque de `data/plazos.yaml` y de `data/cobertura.yaml`, sin
fecha y con `9999-12-31`, y el índice y los metadatos de cada norma de esos dos ficheros y de `data/normas.yaml`; lo
mismo la base de la medida del juez; y lo que no está grabado es una falta que nombra la lectura.

**Prueba independiente**: quickstart.md §1 (subtest `lo-que-lleva-a-leer`, `TestLoQueLlevaALeerSinGrabar`,
`TestBaseDeLaReconstruccion`) y §7.a.

- [X] T004 [US1] Lo que la skill lleva a leer, derivado de `data/` y servido en la caché de cada sesión y en la base de la medida del juez (plan.md paso 4; contracts/cache-de-la-sesion.md §1 a §3; data-model §2; research D1, D4, V1, V19). **`internal/evals/consultas.go`**: `PuntoLoQueLlevaALeer Punto = "lo que la skill lleva a leer"`; `datosDelRepositorio = "../../data"`; `lecturasDeLosDatos(dir) ([]Consulta, error)`, que lee `dir/plazos.yaml` con `plazos.LeerTipos`, `dir/cobertura.yaml` con `skills.LeerCobertura` y `dir/normas.yaml` con `skills.LeerNormas`, y anota con `anotarBloque`, en este orden y sin repetir, cada bloque de los tipos (sus `Preceptos`, sus `Unidades` en el orden de `plazos.Unidades()` y sus `PreceptosDelComputo`), cada `Precepto` de la cobertura y el `Aplicable` de su `Norma`, con el origen del fichero de que sale; y, de cada norma de esos bloques, de la cobertura y de `normas.yaml`, `indice` y `metadatos`; un fichero que no se lee o que su lector rechaza es un error; ningún identificador de norma ni de bloque escrito en el código (FR-002). `consultasDeLaCache(evals, dir) ([]Consulta, error)`: las de `ConsultasNecesarias(evals)` y, detrás, las de `lecturasDeLosDatos(dir)`, con el mismo acumulador (la invocación que ya estaba suma su origen). **`internal/evals/preparar.go`**: `prepararLaCache` prepara con `consultasDeLaCache(evals, datosDelRepositorio)` y devuelve su error como error de la preparación (la sesión no se abre); `queFaltaDesde` nombra la falta del punto nuevo `data/<fichero>.yaml: lo que la skill lleva a leer <orden>: «…» terminó con código N: …` (data-model §2). **`internal/evals/medida.go`**: `leerLasEvalsDeHoy` usa `consultasDeLaCache` con `datosDelRepositorio` para la base de la medida (FR-004), sin `sin_servir` (D4). **Tests**: subtest nuevo `lo-que-lleva-a-leer` de `TestEvalsDelRepositorio` (`conjunto_test.go`): desde la caché preparada de una sesión de una eval de cada skill, cada bloque de los dos ficheros, sin fecha y con `--fecha 9999-12-31`, y el índice y los metadatos de cada norma de los tres, salen con 0 con `--offline`, enumerados por el oráculo de los tests de hoy (`lecturasDeLosTipos` y `lecturasDeLaCobertura` de `grabaciones_test.go`, y `skills.LeerNormas` para las normas), no por el código que prueba; `TestLoQueLlevaALeerSinGrabar` (`preparar_test.go`): con una copia de `data/` en `t.TempDir()` con `bloque: a128,` cambiado por `bloque: a129,` en `plazos.yaml`, la preparación da una falta que empieza por `data/plazos.yaml: lo que la skill lleva a leer` y nombra la orden, y la sesión no se abre (para ello `prepararLaCache` o la función que la llama recibe el directorio de `data/`, que en producción es `datosDelRepositorio`); `TestCacheDeLasSesiones` (`preparar_test.go`, nuevo), subtest `sin-servir` con su primera mitad: desde la caché de una sesión de la 07 de `legal-core`, `boe articulo BOE-A-2015-10565 daprimera --offline` sale con 0 sin fecha, con `--fecha 2026-09-09` y con `--fecha 9999-12-31`, y desde la de la 05, `a124` con 0 en esas tres (US1, escenario 1); caso nuevo de `TestBaseDeLaReconstruccion` (`medida_test.go`): la base de la medida de `boe-legislacion` sirve `daprimera`, que no nombra ninguna de sus evals; y los tests de la base o de la preparación que cuentan sus consultas, con las de `data/` añadidas. Con el mutante de FR-002 (`consultasDeLaCache` devuelve solo `ConsultasNecesarias(evals)`), `lo-que-lleva-a-leer` y `TestCacheDeLasSesiones/sin-servir` salen en rojo: se aplica en un árbol aparte bajo `$TMPDIR`, se apunta en `specs/025-h26-1-las-evals-sirven/mutantes.md` con su sustitución, su comprobación de que se aplicó y los tests que fallan, y se retira; el de FR-003 (quickstart.md §7.a) se apunta igual. **Verificación**: los cuatro tests en verde; `make ci` en verde — FR-001 a FR-005; SC-003, SC-008, SC-011. Rutas: internal/evals/consultas.go, internal/evals/preparar.go, internal/evals/medida.go, internal/evals/conjunto_test.go, internal/evals/grabaciones_test.go, internal/evals/job_test.go, specs/025-h26-1-las-evals-sirven/mutantes.md

---

## Fase 4: US2 — Una eval hace fallar a propósito la lectura que quiere probar (P1)

**Objetivo**: `sin_servir` quita sus lecturas de la caché de las sesiones de su eval; el juicio las reparte en
`lecturas_sin_servir`; el comando de bloque que la eval hace fallar admite 4 y 5; la serie cuenta las sesiones que lo
intentaron; `lectura_sin_intentar` decide; y las evals 11 y 12 de `legal-core` lo usan.

**Prueba independiente**: quickstart.md §1 (segunda mitad de `TestCacheDeLasSesiones/sin-servir`), §3
(`TestJuzgarLasLecturasSinServir`), §4 (`TestUmbralesDeLasLecturas`) y §5.

- [ ] T005 [US2] `sin_servir` quita sus consultas de la caché de las sesiones de su eval (plan.md paso 4; contracts/cache-de-la-sesion.md §1; data-model §1 y §2; research D3, V3). **`internal/evals/formato.go`**: `LecturaSinServir.consulta(Invocacion) bool` —la invocación es del applet `boe` y consulta esa lectura: para un bloque, verbo `articulo` o `articulos`, primer argumento de posición la norma y el bloque entre los demás, con o sin `--fecha` (con `lecturaDeLosArgumentos`); para una consulta de norma, ese verbo y esa norma (con `esDeLaNorma`)— y `Eval.sinServir(Invocacion) bool`, alguna entrada la consulta. **`internal/evals/consultas.go`**: `sinLoQueNoSeSirve(consultas, eval) []Consulta`, las consultas menos las que `eval.sinServir` reconoce, convertida cada una en la `Invocacion` que la ejecuta (applet, verbo, argumentos, `Consulta: true`). **`internal/evals/preparar.go`**: `prepararLaCache` prepara la caché de la sesión de una eval con `sinLoQueNoSeSirve(consultasDeLaCache(evals, datos), eval)`; la base de la medida del juez sigue sin `sin_servir` (D4). **Tests** (`internal/evals/consultas_test.go`, casos nuevos de `TestConsultasNecesarias`): con una eval con `sin_servir` de un bloque y de un `indice`, `sinLoQueNoSeSirve` quita `articulo <norma> <bloque>`, la misma con `--fecha 9999-12-31` y con otra fecha, `articulos` con ese bloque entre otros, e `indice <norma>`, y deja las de otros bloques y otras normas y `metadatos <norma>`; con una eval sin `sin_servir`, las deja todas. La comprobación desde la caché preparada de una sesión, con 4 en tres fechas, la hace T008 con la 11 y la 12 del repositorio, que todavía no existen. **Verificación**: `TestConsultasNecesarias`, `TestCacheDeLasSesiones` y `TestEvalsDelRepositorio` en verde (ninguna eval del repositorio lleva aún `sin_servir`, así que ninguna caché cambia); `make ci` en verde — FR-013; SC-005. Rutas: internal/evals/formato.go, internal/evals/consultas.go, internal/evals/preparar.go, internal/evals/consultas_test.go

- [ ] T006 [US2] El juicio reparte en `lecturas_sin_servir` lo que consulta una lectura de `sin_servir`, y el comando de bloque que la eval hace fallar admite 4 y 5 (plan.md paso 6; contracts/juicio-e-informe.md §1 y §2; data-model §4 y §5, filas 1, 3, 4, 5 y 7 del reparto; research D5, D6). **`internal/evals/formato.go`**: `Eval.comandoSinServir(ComandoEsperado) bool`, el comando es de la forma bloque y su norma y su bloque son una entrada de bloque de `SinServir`. **`internal/evals/juzgar.go`**: `InvocacionFallida.Codigo` pasa a `*int` (`codigo`, `null` sin código); `ResultadoDeEval.LecturasSinServir []InvocacionFallida` (`lecturas_sin_servir`, detrás de `otras_fallidas`, `[]` si no hay); en `informarLaJuzgada`, la invocación —de la traza o por una llamada— que consulta una lectura del `sin_servir` de su eval (`eval.sinServir`), termine como termine (también sin código o con 0), va a `lecturas_sin_servir` con su orden y su código, y no a `fuera_de_lo_grabado` ni a `otras_fallidas`; las demás, como hoy; `satisface` y `consultoComoPide` admiten también el 4 y el 5 en la invocación que lee el bloque del comando en su `fecha` cuando `eval.comandoSinServir(comando)` (FR-016), y nada más cambia en lo que satisface o ejecuta una invocación. **`internal/evals/informe.go`**: el `fuera_de_lo_grabado` de la raíz, que junta el de cada sesión, ya no lleva lecturas de `sin_servir` (FR-022), y `FueraDeLoGrabadoDelInforme.Codigo` sigue siendo un entero (desde el `*int`, que en esa lista siempre tiene 4 o 5). **`TestJuzgarLasLecturasSinServir`** (`internal/evals/juzgar_test.go`, nuevo), con sesiones sintéticas en `t.TempDir()` y evals construidas en el test con el contenido de contracts/formato-de-eval.md §3.1 y §3.2: la sesión de la 11 que resuelve Leganés, pide `plazos calcular reposicion`, lee `a124` con `--fecha 2026-09-09`, sale con 5 y da `⚠ PLAZO: 2026-10-13` pasa, con la lectura en `lecturas_sin_servir` y no en `fuera_de_lo_grabado` ni en `otras_fallidas`, y no cuenta en `plazo_sin_fecha`; la misma por la herramienta `boe_articulo` con `fecha` y un sobre `fuente-no-disponible`, también; la que la lee sin `--fecha`, la que la lee en otra fecha y la que no la lee llevan el motivo del comando ausente; la que no tiene la línea `⚠ PLAZO: 2026-10-13`, no pasa; la de la 12 con `⚠ NO CUBIERTO:` y `BOE-A-2015-10565` y sin línea del plazo pasa y no cuenta en `cobertura_sin_declarar`, y la que ejecuta `plazos calcular`, no; una sesión cortada que consultó la lectura sin código la lleva en `lecturas_sin_servir` con `codigo: null`; y la 16 de `boe-legislacion` del repositorio, con `lecturas` puestas en el test y sin ellas, juzga igual una sesión que no lee `a26` ni `a27` (FR-007). Los tests que comparan `codigo` o la lista de fallidas se alinean con el `*int`, sin perder lo que comprobaban. **Verificación**: `TestJuzgarLasLecturasSinServir`, `TestJuzgar`, `TestJuzgarPlazos`, `TestInforme` y los del sondeo en verde; `make ci` en verde — FR-007, FR-013, FR-015, FR-016, FR-017, FR-018, FR-022; SC-004, SC-005. Rutas: internal/evals/formato.go, internal/evals/juzgar.go, internal/evals/informe.go, internal/evals/juzgar_test.go, internal/evals/informe_test.go, internal/evals/plazos_test.go, internal/evals/sondeo_test.go

- [ ] T007 [US2] La serie cuenta las sesiones con una lectura de `sin_servir`, y el umbral `lectura_sin_intentar:<modelo>:<modo>` decide (plan.md paso 7; contracts/juicio-e-informe.md §4 y §5; data-model §6 y §7; research D11, D12). **`internal/evals/informe.go`**: `TasaDelInforme.SesionesConLecturaSinServir *int` (`sesiones_con_lectura_sin_servir`): en la serie de una eval con `sin_servir`, cuántas de sus sesiones legibles tienen algo en `lecturas_sin_servir`, también 0; `null` en las demás; no decide. **`internal/evals/umbrales.go`**: detrás de los umbrales de hoy, por cada modo del plan y, en cada modo, por cada modelo del plan en su orden, `lectura_sin_intentar:<modelo>:<modo>`, solo en el informe de una skill con al menos una eval con un comando de bloque cuya lectura está en su `sin_servir` (`comandoSinServir`), como `plazo_sin_fecha` solo en el de la que tiene evals de plazo: total, las sesiones medidas de ese modelo y modo de esas evals (research D11: legibles, sin `sin_medir`, de las series del plan); medida, las de esas que no tienen en `lecturas_sin_servir` ninguna invocación que consulte el bloque de cada uno de esos comandos, con o sin `--fecha` y termine como termine (también la que el tope cortó antes de intentarla); `<=` 0; `decide: true` solo con el modelo que decide (ADR 0031); también con medida 0; la descripción de contracts §4; y su motivo, si decide y no se cumple, en la forma del de `plazo_sin_fecha`, con sus sesiones separadas por «; ». **`internal/evals/doc.go`**: el contrato del paquete nombra la serie y el umbral. **`TestUmbralesDeLasLecturas`** (`internal/evals/umbrales_test.go`, nuevo), con sesiones sintéticas e informe escritos en `t.TempDir()`: tres sesiones medidas de una eval con un comando de bloque en su `sin_servir`, del modelo que decide en el modo orden, que intentan la lectura dan `lectura_sin_intentar:<modelo>:orden` con medida 0, total 3 y `cumple: true`; con una que no la intenta, medida 1, su motivo con la sesión y el veredicto `fallo`; el del modelo informativo, con `decide: false` y el veredicto igual; ausente en el informe de una skill sin una eval así; y la serie de esa eval con `sesiones_con_lectura_sin_servir` 3, 0 si ninguna la intenta, y `null` en la de otra eval (US2, escenario 5). Con el mutante de FR-033 (el umbral retirado, o con `decide: false`), el test sale en rojo: se aplica en un árbol aparte bajo `$TMPDIR`, se apunta en `specs/025-h26-1-las-evals-sirven/mutantes.md` y se retira. Los tests que comparan una serie o el JSON de un informe ganan `"sesiones_con_lectura_sin_servir": null`; ninguna eval del repositorio lleva aún `sin_servir`, así que ninguna lista de umbrales del plan cambia todavía. **Verificación**: `TestUmbralesDeLasLecturas`, `TestUmbralesDelInforme`, `TestInforme`, `TestInformeEnDosModos`, `TestInformeConElJuez` y `TestUmbralesDeLasTresSkills` en verde; `make ci` en verde — FR-015, FR-033; SC-005, SC-008, SC-013. Rutas: internal/evals/informe.go, internal/evals/umbrales.go, internal/evals/doc.go, internal/evals/informe_test.go, internal/evals/umbrales_test.go, internal/evals/serie_test.go, internal/evals/sesiones_test.go, internal/evals/sondeo_test.go, internal/evals/cierre_test.go, internal/evals/ejecucion_test.go, specs/025-h26-1-las-evals-sirven/mutantes.md

- [ ] T008 [US2] Las evals 11 y 12 de `legal-core`, con las cifras que su llegada cambia (plan.md paso 8; contracts/formato-de-eval.md §3.1 y §3.2; data-model §8; research D13, V10, V20). **`evals/legal-core/11-plazo-reposicion-sin-leer-el-precepto.yaml`**, nueva, con el comentario y las claves de contracts §3.1: la pregunta de la 05, `activa: true`, los tres comandos de la 05 (`territorio resolver Leganés`, `plazos calcular reposicion` y el bloque `a124` de `BOE-A-2015-10565` con `fecha: "2026-09-09"`), ese bloque en `sin_servir`, y el `plazo` de la 05 (`fecha: "2026-10-13"`, `avisos: []` y su `razonamiento` copiado tal cual de `evals/legal-core/05-plazo-reposicion-leganes.yaml`); sin `citas` ni `redaccion_fechada`. **`evals/legal-core/12-cobertura-seguridad-social-sin-leer-la-da-primera.yaml`**, nueva, con el comentario y las claves de contracts §3.2: la pregunta, `activa: true`, `daprimera` de `BOE-A-2015-10565` en `sin_servir`, `prohibidos` con `plazos calcular` y `no_cubierto` con `norma: BOE-A-2015-10565`; sin `citas` ni `comandos`. **Tests**: segunda mitad de `TestCacheDeLasSesiones/sin-servir` (`preparar_test.go`): desde la caché de una sesión de la 11, `boe articulo BOE-A-2015-10565 a124 --offline` sale con 4 sin fecha, con `--fecha 2026-09-09` y con `--fecha 9999-12-31`, y desde la de la 12, `daprimera` con 4 en las tres (US2, escenario 1; FR-013); `TestJuzgarLasLecturasSinServir` (`juzgar_test.go`) juzga además las sesiones de T006 con la 11 y la 12 leídas del repositorio, con el mismo resultado; y los literales que la llegada cambia: `TestEvalsDelRepositorio` (doce evals de `legal-core`, `evalsDeLegalCoreConLaDeVigencia`, y `conjunto-legal-core` y `grabado` en verde), `TestDefinicionDelJob` (`peorCasoDeLegalCore` de `16261 s = …` a `19525 s = 485 s + (⌈67 / 2⌉ + ⌈66 / 2⌉ + ⌈6 / 2⌉) × (22 s + 240 s + 10 s)`, con `⌈` y `⌉` escritos `\xe2\x8c\x88` y `\xe2\x8c\x89` como hoy, y el caso del tope en 288 minutos, que ahora deja fuera también a `legal-core`; el tope de 352 minutos no cambia), `probarElPlanDelRepositorio` (`plan_test.go`, series y sesiones de `legal-core` con las dos evals nuevas en los dos modos), `TestUmbralesDeLasTresSkills` (`legal-core` con los cuatro `lectura_sin_intentar` detrás de los de hoy: 14 umbrales y 12 que deciden), `TestUmbralesDeLegalCore` (los totales de 30 respuestas que activan, 12 de plazo y 9 de cobertura por modo) y los que arman sesiones de las evals de `legal-core` del repositorio y fijan sus totales (`informe_test.go`, `plazos_test.go`). Con el mutante de FR-013 (`sinLoQueNoSeSirve` aplica el `sin_servir` de cualquier eval de la skill a todas), `TestCacheDeLasSesiones/sin-servir` sale en rojo (la 05 deja de servir `a124`): se aplica en un árbol aparte bajo `$TMPDIR`, se apunta en `specs/025-h26-1-las-evals-sirven/mutantes.md` y se retira. **Verificación**: los tests nombrados en verde; `make ci` en verde — FR-013, FR-016 a FR-020, FR-033; SC-005, SC-008, SC-009, SC-012 a SC-014. Rutas: evals/legal-core/11-plazo-reposicion-sin-leer-el-precepto.yaml, evals/legal-core/12-cobertura-seguridad-social-sin-leer-la-da-primera.yaml, internal/evals/preparar_test.go, internal/evals/juzgar_test.go, internal/evals/conjunto_test.go, internal/evals/definicion_test.go, internal/evals/plan_test.go, internal/evals/informe_test.go, internal/evals/plazos_test.go, internal/evals/umbrales_test.go, internal/evals/job_test.go, internal/evals/medida_test.go, internal/evals/sesiones_test.go, internal/evals/sondeo_test.go, specs/025-h26-1-las-evals-sirven/mutantes.md

---

## Fase 5: US3 — Una eval sirve la exploración razonable que la provoca (P2)

**Objetivo**: `lecturas` es el cuarto punto de las consultas necesarias, y las evals 14, 15 y 16 de
`boe-legislacion` llevan los cuatro bloques que grabó T002.

**Prueba independiente**: quickstart.md §1 (`TestCacheDeLasSesiones/lecturas`, `TestConsultasNecesarias`, `grabado`).

- [ ] T009 [US3] `lecturas` en las consultas necesarias y en las evals 14, 15 y 16 de `boe-legislacion` (plan.md paso 4; contracts/formato-de-eval.md §3.3; contracts/cache-de-la-sesion.md §1 y §2; data-model §2 y §8; research D2, V2). **`internal/evals/consultas.go`**: `PuntoLecturaDeLaEval Punto = "lectura de la eval"`; `ConsultasNecesarias` anota, detrás de las citas de cada eval, cada entrada de sus `lecturas` con `anotarBloque` (sin fecha y con `--fecha 9999-12-31`), con el origen del fichero de la eval. **`internal/evals/preparar.go`**: `queFaltaDesde` nombra la falta del punto nuevo `<fichero de la eval>: la lectura <norma> <bloque> sin su bloque: «…» terminó con código N: …` (data-model §2). **Evals**, detrás de `citas` y sin cambiar nada más, con los motivos de contracts §3.3 tal cual: `evals/boe-legislacion/16-lrjsp-legalidad-por-materia.yaml`, `a26` y `a27` de `BOE-A-2015-10566`; `evals/boe-legislacion/15-irpf-rendimientos-por-materia.yaml`, `a18` de `BOE-A-2006-20764`; `evals/boe-legislacion/14-trlrhl-impuestos-por-materia.yaml`, `a15` de `BOE-A-2004-4214`. **Tests**: caso nuevo de `TestConsultasNecesarias` (`consultas_test.go`): cada `lecturas` da su bloque sin fecha y con `9999-12-31`, con el punto `lectura de la eval`, y un bloque que ya pedía otra eval suma su origen; subtest nuevo `lecturas` de `TestCacheDeLasSesiones` (`preparar_test.go`): desde la caché de una sesión de una eval de `boe-legislacion` que no es la 14, la 15 ni la 16, los cuatro bloques salen con 0 sin fecha y con `--fecha 9999-12-31` (US3, escenario 1; FR-006); `TestEvalsDelRepositorio` lee las tres con sus `lecturas` y su subtest `grabado` sigue en verde con las cuatro respuestas que grabó `grabar_datos` tras T002 (FR-010, FR-011); `TestJuzgarLasLecturasSinServir` (`juzgar_test.go`) juzga la sesión de la 16 que no lee `a26` ni `a27` con la 16 del repositorio, con el mismo juicio que sin `lecturas` (US3, escenario 2; FR-007); y los tests de la base de la medida o de la preparación que cuentan sus consultas, con las de `lecturas` añadidas. **Verificación**: los tests nombrados en verde; `make ci` en verde — FR-006, FR-007, FR-010, FR-011; SC-004, SC-018. Rutas: internal/evals/consultas.go, internal/evals/preparar.go, evals/boe-legislacion/14-trlrhl-impuestos-por-materia.yaml, evals/boe-legislacion/15-irpf-rendimientos-por-materia.yaml, evals/boe-legislacion/16-lrjsp-legalidad-por-materia.yaml, internal/evals/conjunto_test.go, internal/evals/juzgar_test.go, internal/evals/medida_test.go, internal/evals/job_test.go

---

## Fase 6: US4 — El informe cuenta las lecturas fuera de lo grabado (P2)

**Objetivo**: el informe de cada skill publica `fuera_de_lo_grabado:<modelo>:<modo>` con `decide: false`, sin las
lecturas de `sin_servir`.

**Prueba independiente**: quickstart.md §4 (`TestUmbralesDeLasLecturas`, `TestUmbralesDeLasTresSkills`).

- [ ] T010 [US4] El umbral publicado `fuera_de_lo_grabado:<modelo>:<modo>` (plan.md paso 7; contracts/juicio-e-informe.md §4; data-model §7; research D11, D12). **`internal/evals/umbrales.go`**: en el informe de toda skill, detrás de los umbrales de hoy, por cada modo del plan y, en cada modo, por cada modelo del plan en su orden, `fuera_de_lo_grabado:<modelo>:<modo>`, delante del `lectura_sin_intentar` del mismo modelo y modo donde lo haya (contracts §4): medida, la suma de las lecturas del `fuera_de_lo_grabado` de las sesiones medidas de ese modelo y modo, que ya no lleva las de `sin_servir` (T006); total, esas sesiones; `<=` 0; `decide: false` siempre, con su `cumple` como el de cualquier umbral, sin motivo ni cambio del veredicto; también con medida 0; la descripción de contracts §4. Las sesiones de las evals sin binario ni servidor no cuentan. **`internal/evals/doc.go`**: el contrato del paquete lo nombra. **Tests**: casos nuevos de `TestUmbralesDeLasLecturas` (`umbrales_test.go`): con tres sesiones medidas de `claude-sonnet-5-5` en el modo orden, dos lecturas con 5 fuera de la caché en una y una de `sin_servir` con 5 en otra, `fuera_de_lo_grabado:claude-sonnet-5-5:orden` con medida 2, total 3, `decide: false`, `cumple: false` y el veredicto igual (US4, escenario 1); un modelo y un modo sin ninguna, con medida 0 (escenario 2); y una sesión sin binario ni servidor que no cuenta; `TestUmbralesDeLasTresSkills` (`informe_test.go`), con el plan del job y en su orden: `legal-core` 18 umbrales y 12 que deciden; `boe-legislacion` y `jurisprudencia` 19 y 11, sin ningún `lectura_sin_intentar`, con los que ya decidían iguales (SC-019, SC-020); y los que comparan la lista entera de umbrales (`TestUmbralesDelInforme`, `TestUmbralesDeBoeLegislacion`, `TestUmbralesDeJurisprudencia`, `TestUmbralesDeLegalCore`, `TestUmbralesConJuezYConSentencias`, `TestInformeEnDosModos`, `TestInformeConElJuez`, `TestInforme`), con los nuevos añadidos y nada quitado. **Verificación**: los tests nombrados en verde; `make ci` en verde — FR-021, FR-022; SC-006, SC-015. Rutas: internal/evals/umbrales.go, internal/evals/doc.go, internal/evals/umbrales_test.go, internal/evals/informe_test.go, internal/evals/serie_test.go, internal/evals/sesiones_test.go, internal/evals/sondeo_test.go, internal/evals/cierre_test.go, internal/evals/ejecucion_test.go

---

## Fase 7: US6 — `legal-core` v1.2.3 dice qué hacer con una lectura fallida (P2)

**Objetivo**: los pasos 7 y 10 de la skill, en 299 líneas como mucho; las evals que los miden ya están (T008).

**Prueba independiente**: quickstart.md §6.

- [ ] T011 [US6] `legal-core` v1.2.3: los pasos 7 y 10 (plan.md paso 9; contracts/skill-legal-core.md §1 a §4; research D14, V9). En `skills/legal-core/SKILL.md`: el paso 7 dice, en el mismo párrafo que ya dice que la línea `⚠ NO CUBIERTO:` va aunque falle una lectura, que tras esa línea la respuesta no expone la norma que rige el procedimiento —sus plazos, sus recursos—, que nombra la línea con lo que trae `references/cobertura_por_materia.md` (FR-030); el paso 10 gana un punto: si la lectura de un precepto de `data.preceptos` en la fecha de inicio sale con 4 o 5, la línea `⚠ PLAZO:` se mantiene con la fecha de `plazos` y la cita de cada precepto de `data.preceptos`, y en lugar de su línea `⚠ REDACCIÓN A FECHA:` la respuesta dice aparte que no pudo leer esa redacción, como `boe-legislacion` dice cualquier lectura fallida, sin afirmar lo que dice el precepto (FR-029). Lo que entra se compensa reescribiendo con menos líneas los dos párrafos que ya tratan la lectura fallida —el del paso 7 y el primer punto del paso 10—, sin quitar ninguna regla ni ningún caso, ni tocar el frontmatter salvo lo que pida la versión, la tabla de comandos generada ni `references/`. **Verificación**: `wc -l skills/legal-core/SKILL.md` da 299 o menos; `make skills-check` (`TestSkillsDelRepositorio`) en verde; `make ci` en verde — FR-029, FR-030, FR-031; SC-021. Rutas: skills/legal-core/SKILL.md

---

## Fase 8: US5 — El motivo de un 1 que no es de kitlegal (P3)

**Objetivo**: la llamada sin sobre sale con `codigo: null` y `motivo: "sin_sobre"`, y la orden que no pudo escribir su
salida, enlazada con su uso de Bash, con `motivo: "salida_no_escrita"` y `bash`; nada decide por ellos.

**Prueba independiente**: quickstart.md §3 (`TestJuzgarElMotivoDeLasFallidas`, `TestLineaDeLaSalidaNoEscrita`,
`TestLeerLasLlamadas`, `TestTextosDeLaSesion`) y §7.b.

- [ ] T012 [US5] Lo que se lee de una sesión para el motivo: `Llamada.SinSobre` y `Sesion.OrdenesDeBash` (plan.md paso 5; data-model §3; research D7, D10). **`internal/evals/sesion.go`**: `Llamada.SinSobre bool`, verdadero si el transcript trae su `tool_result`, es un error (`is_error` verdadero) y ninguno de los textos de su contenido es un objeto JSON con `ok`; una llamada sin resultado, falso. `Sesion.OrdenesDeBash []Texto`: cada uso de Bash cuyo `input.command` lleva la palabra `kitlegal` y cuyo `tool_result` trae el transcript, con su orden tal cual y su salida, en el orden de sus `tool_use`, leídos de lo que ya se lee para el juez. `Sesion.Textos`, `Llamada.Error` y `Llamada.Clase` no cambian. **Tests** (`internal/evals/sesion_test.go`): casos nuevos de `TestLeerLasLlamadas` —la llamada a `boe_indice` cuyo resultado es un error con el aviso del agente de que la herramienta no existe, `SinSobre`; la de un error con un sobre de kitlegal, la que termina bien y la que no tiene resultado, no— y de `TestTextosDeLaSesion` —dos usos de Bash con `kitlegal`, uno con `| python3 -c …` y la línea de la escritura fallida en su salida, en su orden y tal cual; un uso de Bash sin `kitlegal` y uno sin `tool_result`, fuera—. **Verificación**: `TestLeerLasLlamadas` y `TestTextosDeLaSesion` en verde; `make ci` en verde — FR-023, FR-025, FR-026; SC-007. Rutas: internal/evals/sesion.go, internal/evals/sesion_test.go

- [ ] T013 [US5] El juicio publica el motivo de la llamada sin sobre y de la orden de salida no escrita, sin que decida nada (plan.md paso 6; contracts/juicio-e-informe.md §1, §3 y §6; data-model §4, filas 2 y 6, el enlace, y §5; research D7, D8, D10). **`internal/evals/juzgar.go`**: `InvocacionInformada.Motivo` (`motivo`, omitido vacío); `InvocacionFallida.Motivo` (`sin_sobre` o `salida_no_escrita`, omitido vacío) y `InvocacionFallida.Bash` (la orden tal cual, omitida vacía); el prefijo de la escritura fallida se pide una vez a `render.Nuevo` con una salida estándar que falla, tomando del error que devuelve lo que va delante de la causa, nunca una copia literal del mensaje (FR-025); en `informarLaJuzgada`, fila 2 antes que la 3: la llamada `SinSobre` se informa con `codigo: null` y `motivo: "sin_sobre"` y va a `otras_fallidas` con ellos, también si consulta una lectura de `sin_servir` (que entonces no es una lectura intentada); fila 6: una invocación de la traza con código 1 se enlaza con un uso de `Sesion.OrdenesDeBash` si su orden contiene como palabras (`contieneComoPalabra`) el applet, el verbo y cada argumento de la invocación y alguna línea de su salida empieza por el prefijo, y si es el único uso que lo cumple; enlazada, lleva `codigo: 1`, `motivo: "salida_no_escrita"` y `bash`, en `otras_fallidas` o, si consulta una lectura de `sin_servir`, en `lecturas_sin_servir`; sin enlace, con dos usos que lo cumplen o con otra salida de error, como hoy (FR-027). Si la sesión pasa, sus motivos y lo que satisface o ejecuta cada invocación no cambian (FR-028). **`internal/evals/informe.go`**: en la tabla de invocaciones de `informe.md`, la fila de una invocación con `codigo: null` y `motivo` dice «sin sobre» en la columna del código, en lugar de «sin código (sesión cortada)». **Tests** (`internal/evals/juzgar_test.go`): `TestJuzgarElMotivoDeLasFallidas`, con sesiones sintéticas: la llamada a `boe_indice` sin sobre del modo orden, con `codigo: null` y `motivo: "sin_sobre"` en `invocaciones` y en `otras_fallidas`, y también si consulta una lectura de `sin_servir`; la orden `kitlegal boe indice BOE-A-2015-10565 --json | python3 -c …` cuya salida trae la línea y la invocación `boe indice BOE-A-2015-10565` de la traza con 1, en `otras_fallidas` con `codigo: 1`, `motivo: "salida_no_escrita"` y `bash` tal cual, y en `lecturas_sin_servir` si consulta una lectura de `sin_servir`; el 1 con otra salida de error, sin uso que la enlace y con dos usos que lo cumplen, como hoy; y en todos, `pasa` y `motivos` iguales a los de la misma sesión juzgada sin los campos nuevos (FR-028); `TestLineaDeLaSalidaNoEscrita`: la salida de error de `app.Main` con una salida estándar que falla trae una línea que empieza por el prefijo que reconoce el arnés. Casos nuevos de `TestUmbralesDeLasLecturas` (`umbrales_test.go`): la sesión cuya única invocación de la lectura es una llamada sin sobre cuenta en `lectura_sin_intentar` (medida 1, veredicto `fallo`), y la que la intenta con una orden de salida no escrita, no (medida 0). Con el mutante de FR-025 (`render: no se pudo escribir en %s` cambiado por `render: no se ha podido escribir en %s` en `internal/render/render.go`, en un árbol aparte bajo `$TMPDIR`), `TestTuberiaCerrada` sale en rojo y `TestLineaDeLaSalidaNoEscrita` sigue en verde (quickstart.md §7.b): se apunta en `specs/025-h26-1-las-evals-sirven/mutantes.md` y se retira; `internal/render` no cambia en el árbol. **Verificación**: los tests nombrados, `TestJuzgar`, `TestInforme` y los del sondeo en verde; `make ci` en verde — FR-015, FR-023 a FR-028, FR-033; SC-007, SC-017. Rutas: internal/evals/juzgar.go, internal/evals/informe.go, internal/evals/juzgar_test.go, internal/evals/umbrales_test.go, internal/evals/informe_test.go, internal/evals/sondeo_test.go, specs/025-h26-1-las-evals-sirven/mutantes.md

---

## Fase 9: Pulido — documentos, mutantes y cierre

- [ ] T014 Lo que los documentos dicen del formato de las evals, de la caché de una sesión y de los umbrales nuevos, con cada cifra, fichero, test y umbral sacados del árbol y no de memoria (plan.md paso 10; research V18). **`CHANGELOG.md`**, dentro de `## [Unreleased]`: `lecturas` y `sin_servir` en el formato de las evals; lo que sirve la caché de cada sesión (lo que la skill lleva a leer desde `data/`, y `lecturas`) y lo que `sin_servir` le quita; `lecturas_sin_servir`, `sesiones_con_lectura_sin_servir`, los umbrales `fuera_de_lo_grabado:<modelo>:<modo>` (publicado, `decide: false`) y `lectura_sin_intentar:<modelo>:<modo>` (decide con 0, solo donde hay una eval con un comando de bloque en su `sin_servir`); el `motivo` y el `bash` de las invocaciones fallidas; las evals 11 y 12 de `legal-core` y las `lecturas` de la 14, la 15 y la 16 de `boe-legislacion`, con los cuatro bloques grabados; y `legal-core` v1.2.3, con lo que dicen los pasos 7 y 10 (quickstart.md §6 busca `legal-core. v1.2.3`); ninguna sección de una versión publicada cambia. **`CONTRIBUTING.md`**: la tabla del formato de las evals gana `lecturas` y `sin_servir`, con sus reglas y rechazos; y en «Job de evals», qué sirve la caché de una sesión, los dos umbrales nuevos con lo que cuentan y si deciden, el número de umbrales de cada skill (18 y 12 que deciden en `legal-core`; 19 y 11 en las otras dos), el peor caso de `legal-core` (el literal de `TestDefinicionDelJob`) y los campos nuevos del informe; solo lo falso o incompleto se reescribe. `CLAUDE.md` y `docs/WORKFLOW.md` no dicen nada de la caché de las sesiones (research V18, `grep -n caché` en los dos) y no cambian. **Verificación**: cada test, fichero, umbral y cifra nombrados existen en el árbol (`grep -n`); `make ci` en verde — FR-032; SC-010. Rutas: CHANGELOG.md, CONTRIBUTING.md

- [ ] T015 Los mutantes del hito, aplicados uno cada vez en un árbol aparte fuera del repositorio sobre la cabeza de la rama, y su registro completo (plan.md paso 11; research D16; quickstart.md §7). Con `git worktree add --detach "$TMPDIR/h261-copia" HEAD` y `go -C "$TMPDIR/h261-copia"`, sin tocar el árbol de trabajo, cada uno por separado y con su comprobación de que se aplicó (`git diff --no-index --quiet` contra el `.orig`) y el fichero restaurado tras él: (1) FR-003, `bloque: a128,` por `bloque: a129,` en `data/plazos.yaml`: `TestEvalsDelRepositorio` en rojo con la falta `data/plazos.yaml: lo que la skill lleva a leer …a129…`; (2) FR-002, `consultasDeLaCache` sin la derivación de `data/`: `TestEvalsDelRepositorio/lo-que-lleva-a-leer` y `TestCacheDeLasSesiones/sin-servir` en rojo; (3) FR-013, el `sin_servir` de cualquier eval aplicado a todas las de la skill en `sinLoQueNoSeSirve`: `TestCacheDeLasSesiones/sin-servir` en rojo; (4) FR-033, `lectura_sin_intentar` con `decide: false`, y retirado: `TestUmbralesDeLasLecturas` y `TestUmbralesDeLasTresSkills` en rojo; (5) FR-025, el mensaje de `internal/render/render.go` cambiado: `TestTuberiaCerrada` en rojo y `TestLineaDeLaSalidaNoEscrita` en verde; sin mutantes, todos pasan. **`specs/025-h26-1-las-evals-sirven/mutantes.md`**: por mutante, el cambio en una línea, la orden ejecutada y cada test que falla, con su subtest y su mensaje, tal como se vio en esta sesión, completando lo que apuntaron T004, T007, T008 y T013. Esta tarea no cambia nada bajo `internal/`; si un mutante no hace fallar ningún test, lo anota con su salida y la tarea no se da por terminada, porque el refuerzo es del test de la tarea que construyó el control. El árbol aparte se retira con `git worktree remove --force "$TMPDIR/h261-copia"`, y `git worktree list` ya no lo nombra. **Verificación**: los cinco mutantes con al menos un test en rojo (y el 5 con el del arnés en verde) en `mutantes.md`; `make ci` en verde; `git status` sin nada fuera de `mutantes.md` — FR-002, FR-003, FR-013, FR-025, FR-033; SC-007, SC-008. Rutas: specs/025-h26-1-las-evals-sirven/mutantes.md

- [ ] T016 Cierre de la Definition of Done, sin código de producto (`docs/ROADMAP.md` §1; quickstart.md §0 a §9; plan.md, «Controles de umbral»): `specs/025-h26-1-las-evals-sirven/cierre.md` con (1) la lista, sacada de `git diff --name-status main`, de lo creado o modificado, y la salida vacía de quickstart.md §9 (`git diff --name-only main -- cmd internal/app internal/cli internal/render internal/source internal/core internal/httpx internal/cache internal/graph internal/mcp .github testdata` y la carpeta de la evidencia), con la comprobación de que no cambia nada de «Lo que no cambia» de la batería (FR-028); (2) la cobertura sobre `coverage.out` y `coverage-integration.out` de `make ci` con `go tool cover -func` —global ≥ 70 % e `internal/core/**` ≥ 85 % (Definition of Done §1.9)— con la cifra de cada una y la de `internal/evals`, y, si alguna queda por debajo, el hueco (fichero y funciones sin cubrir) para la revisión final; esta tarea no escribe tests; (3) el resultado de quickstart.md §0 a §9, orden a orden, con lo esperado de cada una, y la constancia de que §10 lo hace el workflow en el cierre; (4) la tabla de plan.md, «Controles de umbral», con el test de cada fila `ci:` encontrado como `^func <Test>\(` en su ruta y verde, el objetivo `skills-check` en verde, y cada fila `evals:` con el umbral que publican T007, T008 y T010 o el de hoy que sigue con `decide: true` en el informe de un plan sintético (`TestUmbralesDeLasTresSkills`); y (5) los dos límites conocidos que el spec manda enseñar en el informe final («Fuera de alcance»: la línea `⚠ REDACCIÓN A FECHA:` de memoria en la 11 y la exposición de la norma del procedimiento tras `⚠ NO CUBIERTO:`), y que SC-017 se confirma con lo que publique el job de cierre. **Verificación**: `make ci` en verde y `git status` sin nada fuera de `cierre.md` — FR-028, FR-032; SC-001, SC-002, SC-016, SC-019, SC-020. Rutas: specs/025-h26-1-las-evals-sirven/cierre.md

---

## Dependencias y orden de ejecución

El bucle ejecuta las tareas en su orden; ninguna depende de una posterior.

| Tarea | Depende de | Por qué |
|---|---|---|
| T001 | — | el esquema, antes que el tipo que lo lee |
| T002 | — | el manifiesto; `grabar_datos` graba tras su commit |
| T003 | T001 | `LeerEval` valida contra el esquema que admite las dos claves |
| T004 | T003 | la derivación suma su acumulador a `ConsultasNecesarias`, que lee evals con las claves nuevas |
| T005 | T003, T004 | `LecturaSinServir`; `sinLoQueNoSeSirve` sobre `consultasDeLaCache` |
| T006 | T003, T005 | `SinServir` y `sinServir` en el reparto; `comandoSinServir` |
| T007 | T006 | `lecturas_sin_servir` de cada sesión; `comandoSinServir` |
| T008 | T004 a T007 | la caché sin `sin_servir`, su juicio y su umbral, para las dos evals nuevas |
| T009 | T002, T003, T006 | los cuatro bloques grabados; `Lecturas`; la sesión de la 16 de T006 |
| T010 | T006, T007, T008 | `fuera_de_lo_grabado` sin `sin_servir`; delante de `lectura_sin_intentar`; las listas con la 11 y la 12 |
| T011 | T008 | las evals que miden la skill, antes que la skill (Definition of Done §1.10) |
| T012 | — | solo la lectura de la sesión |
| T013 | T006, T007, T012 | el reparto y el umbral que los motivos matizan; `SinSobre` y `OrdenesDeBash` |
| T014 | T001 a T013 | lo que documenta ya está en el árbol, con sus cifras |
| T015 | T004, T007, T008, T013 | los tests que deben fallar con cada mutante |
| T016 | todas | el cierre sobre el diff entero |

### Historias

- **US1** (P1): T004 (con T001 a T003). **US2** (P1): T005 a T008. **US3** (P2): T009 (con T002). **US4** (P2):
  T010. **US6** (P2): T011 (con T008). **US5** (P3): T012 y T013.
- Cada historia se puede probar sola tras sus tareas: US1 con `lo-que-lleva-a-leer` y `TestLoQueLlevaALeerSinGrabar`;
  US2 con `TestCacheDeLasSesiones/sin-servir`, `TestJuzgarLasLecturasSinServir` y `TestUmbralesDeLasLecturas`; US3 con
  `TestCacheDeLasSesiones/lecturas`; US4 con los casos de `fuera_de_lo_grabado`; US5 con
  `TestJuzgarElMotivoDeLasFallidas` y `TestLineaDeLaSalidaNoEscrita`; US6 con `make skills-check`.

### Oportunidades de paralelismo

Ninguna tarea lleva `[P]`: el workflow las ejecuta en secuencia, con un guardián por tarea. Fuera del bucle, T001 y T002
son independientes entre sí, y T012 solo depende del árbol de hoy.

---

## Trazabilidad

### Requisitos → tareas

| Requisito | Tareas |
|---|---|
| FR-001, FR-002, FR-003, FR-004, FR-005 | T004 (T015: mutantes de FR-002 y FR-003) |
| FR-006 | T009 |
| FR-007 | T006, T009 |
| FR-008 | T003 |
| FR-009 | T001, T003 |
| FR-010 | T002, T009 |
| FR-011 | T009 |
| FR-012 | T001, T003 |
| FR-013 | T005, T008 (T015: mutante) |
| FR-014 | T003 |
| FR-015 | T006, T007, T013 |
| FR-016 | T006, T008 |
| FR-017, FR-018 | T006, T008 |
| FR-019, FR-020 | T008 |
| FR-021 | T010 |
| FR-022 | T006, T010 |
| FR-023 | T012, T013 |
| FR-024 | T013 |
| FR-025 | T012, T013 (T015: mutante) |
| FR-026 | T012, T013 |
| FR-027, FR-028 | T013 (T016: FR-028 sobre el diff) |
| FR-029, FR-030, FR-031 | T011 |
| FR-032 | T014 |
| FR-033 | T007, T008, T013 (T015: mutante) |
| SC-001 | todas; T016 |
| SC-002 | plan.md; T016 |
| SC-003 | T004 |
| SC-004 | T003, T006, T009 |
| SC-005 | T003, T005, T006, T007, T008 |
| SC-006 | T010 |
| SC-007 | T012, T013, T015 |
| SC-008 | T004, T007, T008, T015 |
| SC-009 | T008 |
| SC-010 | T014 (el job, el workflow en el cierre) |
| SC-011 | T004 (el job de cierre lo mide) |
| SC-012, SC-013, SC-014 | T007, T008 (el job de cierre lo mide) |
| SC-015 | T010 |
| SC-016 | `TestInforme`, sin cambiar; T016 |
| SC-017 | T013 (el job de cierre lo confirma) |
| SC-018 | T002, T009 |
| SC-019, SC-020 | T010 (los que deciden, iguales en `TestUmbralesDeLasTresSkills`); T016 |
| SC-021 | T011 |

### Controles de umbral (plan.md) → tarea que construye el control

| Fila de plan.md | Control | Tarea | El test que lo ve fallar |
|---|---|---|---|
| FR-033 (orden y herramienta) | `evals:legal-core:lectura_sin_intentar:claude-sonnet-5-5:<modo>` | T007 (umbral), T008 (la 11 en su total), T013 (sin sobre y salida no escrita) | `TestUmbralesDeLasLecturas`: una sesión sin la lectura intentada, medida 1 y `fallo`; mutante de FR-033 |
| FR-029 (plazo) | `evals:legal-core:plazo_sin_fecha:claude-sonnet-5-5:<modo>` | T008 (la 11 cuenta en él) | `TestJuzgarLasLecturasSinServir` (la 11 sin la línea no pasa) y `TestUmbralesDeLegalCore` con 12 respuestas de plazo |
| FR-030 (cobertura) | `evals:legal-core:cobertura_sin_declarar:claude-sonnet-5-5:<modo>` | T008 (la 12 cuenta en él) | `TestJuzgarLasLecturasSinServir` (la 12 con `plazos calcular` no pasa) y `TestUmbralesDeLegalCore` con 9 de cobertura |
| FR-019 (activación) | `evals:legal-core:sin_activar:claude-sonnet-5-5:<modo>` | T008 | `TestUmbralesDeLegalCore` con 30 que activan |
| SC-014 (vigencia, redacción fechada) | `evals:legal-core:vigencia_sin_declarar:…`, `evals:legal-core:redaccion_sin_fechar:…` | T008 (sus totales no cambian: 3 y 9) | `TestUmbralesDeLegalCore`, sin cambiar lo que comprueba |
| SC-019, SC-002 (`boe-legislacion`) | `afirma_lo_no_leido` y los otros 10 que deciden | T010 (siguen con `decide: true` en la lista) | `TestUmbralesDeLasTresSkills` y `TestUmbralesDeBoeLegislacion` |
| SC-020, SC-002 (`jurisprudencia`) | sus 11 que deciden | T010 | `TestUmbralesDeLasTresSkills` y `TestUmbralesDeJurisprudencia` |
| SC-016 | `ci:internal/evals/informe_test.go:TestInforme` | existe; T016 lo comprueba | `TestInforme` (una llegada a la red, `fallo`) |
| SC-011, FR-001, FR-003, SC-003 | `ci:…/conjunto_test.go:TestEvalsDelRepositorio`, `ci:…/preparar_test.go:TestLoQueLlevaALeerSinGrabar` | T004 | `lo-que-lleva-a-leer`; la copia de `data/` con `a129` |
| FR-002, FR-005, SC-008 | `TestEvalsDelRepositorio`, `TestCacheDeLasSesiones` | T004, T015 | mutante de FR-002 |
| FR-004 | `ci:…/medida_test.go:TestBaseDeLaReconstruccion` | T004 | la base sin `daprimera` |
| SC-018, FR-006, FR-010, FR-011, SC-004 | `TestCacheDeLasSesiones`, `TestEvalsDelRepositorio` | T009 (con T002) | `lecturas` sin servir o sin grabar |
| FR-013, SC-005, SC-008 | `ci:…/preparar_test.go:TestCacheDeLasSesiones` | T005, T008, T015 | mutante de FR-013 |
| FR-008, FR-009, FR-012, FR-014 | `ci:…/formato_test.go:TestFormatoDeLecturas`, `TestEvalsDelRepositorio` | T003 (con T001) | cada rechazo |
| FR-007, FR-015 a FR-018, SC-005 | `ci:…/juzgar_test.go:TestJuzgarLasLecturasSinServir` | T006, T008, T009 | cada sesión sintética |
| FR-033, FR-015, SC-008, SC-013 (`make ci`) | `ci:…/umbrales_test.go:TestUmbralesDeLasLecturas` | T007, T013, T015 | mutante de FR-033 |
| FR-021, FR-022, SC-006, SC-015 | `TestUmbralesDeLasLecturas`, `ci:…/informe_test.go:TestUmbralesDeLasTresSkills` | T010 | medida 2 de 3 con `decide: false` |
| FR-023, FR-024, FR-026 a FR-028, SC-007, SC-017 | `ci:…/juzgar_test.go:TestJuzgarElMotivoDeLasFallidas`, `ci:…/sesion_test.go:TestLeerLasLlamadas` | T012, T013 | cada sesión sintética |
| FR-025, SC-007 | `ci:…/juzgar_test.go:TestLineaDeLaSalidaNoEscrita`, `ci:internal/app/tuberia_unix_test.go:TestTuberiaCerrada` | T013, T015 | mutante de FR-025 |
| FR-020, SC-009 | `ci:…/definicion_test.go:TestDefinicionDelJob` | T008 | el tope en 288 minutos |
| FR-031, SC-021 | `ci:Makefile:skills-check`, `ci:internal/app/skills_test.go:TestSkillsDelRepositorio` | T011 | 300 líneas o más |

### Definition of Done (`docs/ROADMAP.md` §1) → tareas

| Punto | Aplica | Tarea |
|---|---|---|
| 1 · `make ci` en verde | sí | todas; T016 |
| 2 · tests offline; fixtures grabados | sí | cada tarea con sus tests; T002 y `grabar_datos` para los cuatro bloques |
| 3 · `net/http`, `os.Exit`, `fmt.Print*` | sí, sin cambios | ningún import ni escritura nueva (plan.md, R2, R4, R5); `make ci` |
| 4 · salida de applet contra `schemas/` | no: ningún applet cambia | — |
| 5 · errores tipados | no: ninguno cambia (FR-028) | T016 |
| 6 · e2e y `CHANGELOG.md` | `CHANGELOG.md` sí; e2e no aplica (plan.md) | T014 |
| 7 · ADR | no: el ADR 0043 ya está aceptado en `main` (#179) | — |
| 8 · `docs/SOURCES.md` | no: la fila no cambia (research D17) | — |
| 9 · cobertura | sí | T016 |
| 10 · skill: evals antes que el código, < 300 líneas | sí | T008 (evals), T011 (skill) |
| 11 · territorio | no: ningún caso especial; Leganés solo en las evals | — |
| 12 · grafo | no: nada del grafo cambia | — |
| 13 · plazos: común y regulado aparte | sí: la 11 y la 12 | T008 |
| `schemas/` | sí | T001 |

---

## Estrategia de implementación

1. **Datos y fundación** (T001 a T003): el esquema, el manifiesto (y su grabación) y el formato.
2. **MVP, US1** (T004): con la derivación de `data/`, el rojo de `cobertura_sin_declarar` de la #175 ya no depende de la
   caché. Se valida con quickstart.md §1 y §7.a.
3. **US2** (T005 a T008): la robustez de `legal-core` medida a propósito, con su umbral que decide.
4. **US3, US4, US6** (T009 a T011): `lecturas`, la señal publicada y el texto de la skill.
5. **US5** (T012, T013): el motivo de los 1 que no son de kitlegal.
6. **Pulido** (T014 a T016): documentos, mutantes y cierre. El job de cierre lo lanza el workflow tras la revisión final.

---

## Comprobación contra la rúbrica de `juez_tasks` y `precheck.sh tasks`

- precheck: cada tarea con el formato `- [ ] Tnnn`, ids T001 a T016 únicos y correlativos; cada una con rutas tras
  «Rutas:»; las que nombran el directorio de esquemas o uno de fixtures llevan `[datos]` (T001 y T002) y ninguna otra
  los nombra en su línea; ninguna nombra la carpeta de la evidencia con su ruta, una grabación con la variable de
  grabar, la plataforma ni a una persona; ninguna `[aceptacion]`, porque el plan dice «Aceptación e2e: no aplica».
- a · `analisis_critico`: lo comprueba el juez contra `gates/analyze.md`.
- b · `trazabilidad`: cada tarea cita sus FR/SC; «Requisitos → tareas» cubre FR-001 a FR-033 y SC-001 a SC-021; cada
  fila de «Controles de umbral» tiene su tarea y el test que la ve fallar.
- c · `rebanadas_verdes`: test e implementación en el mismo diff, salvo las excepciones declaradas arriba (T001, T002,
  T011, T014 a T016), cada una con su verificación; el orden es el de «Dependencias».
- d · `rutas_declaradas`: cada tarea nombra sus ficheros, ninguno genérico; los tests que pueden cambiar por una lista
  literal están declarados (lo comprobado arriba).
- e · `datos_separados`: T001 y T002, solas y sin código; los cuatro bloques salen de `grabar_datos`.
- f · `dod`: la tabla de la Definition of Done.
- g · `checklist_veraz`: `checklists/requirements.md` no cambia.
- h · `aceptacion_primero`: el plan justifica «no aplica» (plan.md, «Aceptación e2e» y Complexity Tracking).
- i · `autonomia`: ninguna tarea exige a una persona, publica la rama ni mide CI o evals remotas.
