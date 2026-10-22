# Tasks: H30 · Las evals miden solo lo que cambia: dependencias de cada eval y resultados por huella

**Input**: `specs/026-h30-las-evals-miden/` (spec.md, plan.md, research.md, data-model.md, contracts/, quickstart.md)

**Prerrequisitos**: spec.md y plan.md leídos; `.specify/memory/constitution.md` 2.14.0 aplicada (umbral de
materialidad, criterio de uso, proporcionalidad y convergencia en «Gates»; controles de umbral del ADR 0029; ADR 0018,
0028, 0031, 0032, 0037 y 0038). **Ninguna tarea crea el esqueleto ni los gates: `make ci` existe y pasa**, y cada tarea
lo deja en verde al terminar.

**Aceptación**: **no aplica** (plan.md, «Aceptación e2e», y Complexity Tracking; research D20). El hito no cambia nada
que vea quien usa `kitlegal`: el binario no enlaza `internal/evals`, ningún `SKILL.md` ni eval cambia (FR-026) y las dos
funciones nuevas de `internal/skills` solo leen. Ninguna tarea `[aceptacion]`. La aceptación del spec (SC-006 y SC-007)
son tests de `make ci` sobre la decisión del job (`TestDecisionDelJob`, T010), y la medición de cierre la lanza el
workflow tras la revisión final, ya con el job nuevo (SC-014).

**Tests**: obligatorios (constitución §III; plan.md, «Controles mecánicos»). Cada tarea trae sus tests y la
implementación que los hace pasar en el mismo diff; los nombres de test y sus ficheros son los de plan.md, «Tests
nuevos» y «Tests existentes que cambian lo que esperan», porque el informe final busca cada control `ci:<ruta>:<Test>`
como `^func <Test>\(` en esa ruta.

**Rebanadas verticales y excepciones declaradas.** Son rebanada completa —test e implementación en el mismo diff,
`make ci` en verde por sí solas— T002 a T015. Las demás, cada una por su razón:

- **T001** es un test sobre código que ya existe (el golden del informe de la suite completa, fijado antes de tocar
  `informe.go`; plan.md, «Orden de implementación», paso 1): no hay implementación que añadir, y es lo que hace
  comprobable FR-026 en las tareas siguientes.
- **T016** es el ADR (FR-024, SC-013): documentación; su control, `TestLimiteDeLaTasa`, entra antes, en T002, y el ADR
  escribe las cifras que ese test y `scripts/workflow/serie.sh` dan al ejecutarlos.
- **T017** es documentación (SC-016); su verificación es que cada disparador, test y fichero que nombra existe en el
  árbol, y `make ci`.
- **T018** es el cierre de la Definition of Done, sin código de producto.

**Tareas `[datos]`**: ninguna (plan.md, «Fixtures, `testdata/` y `schemas/`»; «Datos externos»: ninguno). Las respuestas
de la plataforma, las medidas de `main`, los informes de prueba y los goldens los construye cada test en memoria o en
`t.TempDir()` (research D15, D16). Ningún manifiesto `grabaciones.json`, ningún paso `grabar_datos`, ninguna fuente de
`docs/SOURCES.md` y ningún fichero de `data/` cambian.

**Lo que este fichero añade al plan, comprobado en el árbol en esta sesión** (lectura y `grep`, sin ejecutar ningún
test):

- `internal/evals/conjunto.go` ya define `type Conjunto` (el conjunto de evals de una skill y sus reglas, H5): el
  conjunto de dependencias de una eval de FR-001 no puede llamarse así; T004 le da otro nombre (`dependenciasDeLaEval`
  o equivalente, sin exportar).
- `internal/evals/tanda.go` ya tiene la ejecución de `gh` con texto constante y cada valor en una variable del entorno
  (`ghDeLaTanda.ejecutar`); `plataforma.go` (T008) la reutiliza o la extrae a un punto común, sin duplicarla
  (constitución V).
- `TestEjecucionDelJob` y `TestTandaDelCommit` están en `internal/evals/job_test.go` (etiqueta `evals`);
  `TestDefinicionDelJob` en `definicion_test.go`; `TestFilaDeLaSerie` y los demás de la serie en `serie_test.go`.
- `internal/skills/comandos.go` y `internal/skills/referencias.go` ya existen: T003 añade a cada uno su función.
- Ningún test de `make ci` ejecuta hoy `scripts/evals.sh` (solo `scripts/evals-sesion.sh`, en `sesiones_test.go`); el
  `claude` sustituto vive en `sustitutos_test.go`. T013 añade el primero, que comprueba la obligación de
  `DECISION_DE_EVALS`; plan.md, «Tests existentes que cambian lo que esperan», corregido en ese sentido.
- `evals/boe-legislacion/juez/clases.yaml` da `umbral: 0.028` a `afirma_lo_no_leido` y
  `evals/jurisprudencia/juez/clases.yaml`, `umbral: 0.017`: de ahí salen *p*ref y k(*N*) en T002, nunca de una cifra
  escrita en el test.
- El último ADR es el 0044: el de FR-024 es el **0045**.

---

## Formato: `- [ ] Tnnn [Story?] Descripción — FR/SC. Rutas: …`

- **[P]**: ninguna tarea lo lleva; en el bucle del workflow todo va en secuencia.
- **[aceptacion]**: ninguna (Aceptación e2e: no aplica).
- **[datos]**: ninguna.
- **[Story]**: US1 una propuesta mide las evals que ha podido cambiar · US2 lo que no cambia reutiliza el resultado de
  `main` por su huella · US3 la release exige la suite completa · US4 la suite completa a petición · US5 el ADR del
  umbral en tasa y la serie. Las fases de fundación y pulido no llevan historia; una tarea que sirve a varias lleva la
  principal y nombra las demás por sus FR.

## Batería de verificación por tarea

- **`make ci` en primer plano** y la tarea marcada `[X]` en el mismo turno en que termina en verde: formato · lint
  (`depguard`, `forbidigo`, `gosec`, `misspell`, `revive`, `dupl`, `gocyclo`, `paralleltest`, `unused`…) · tests con
  `-race` y `-shuffle=on` · `test-integration` · `test-tiempos` · `vuln` · `schema-check` · `skills-check` ·
  `goreleaser-check` · secretos · módulos. Ningún registro de la verificación en la raíz del repositorio. Las salidas de
  `go test`, de `git diff` y de `make` se leen con `rtk proxy` o con sondas positivas, porque el proxy de la sesión las
  resume. `go test -run` parte el patrón por `/`: un `-run` por test, y se cuentan los `PASS`.
- **Los puntos de entrada con la etiqueta `evals` compilan**: `go vet -tags evals ./internal/evals/` en verde en cada
  tarea que toca `job_test.go` o lo que usa (T010, T012, T014), porque `make ci` no compila esa etiqueta.
- **Sin red ni modelo** (ADR 0032): ninguna tarea usa la red salvo la de las herramientas de Go (`make vuln`), ni graba
  nada, ni escribe en la carpeta de la evidencia; ninguna ejecuta `claude`, `gh`, `make evals`, `make evals-sondeo`,
  `make evals-medir-juez`, `TestEjecucionDelJob`, `TestEleccionDeLaPropuesta`, `TestDecisionDelTrabajo`,
  `TestSuiteCompletaDelCommit`, `TestTandaDelCommit`, `TestMedidaDelJuez`, `TestSondeo` ni `TestGrabar*`. En los tests
  de `make ci` la plataforma va en memoria o con el `gh` sustituto que ya usa la tanda (research D15); `go list` y
  `git clone` son locales, sobre el repositorio o sobre un clon en `t.TempDir()`.
- **Sin personas a mitad**: ninguna tarea la hace, la revisa ni la desbloquea una persona.
- **Ningún umbral se cumple cambiándolo** (ADR 0029; FR-026): ninguna tarea cumple un umbral de plan.md, «Controles de
  umbral», cambiando su valor, retirando su test o un caso de una tabla, ni cambia las evals de `evals/`, las clases del
  juez y sus umbrales (`clases.yaml`), las repeticiones, la regla de 2 de 3, los modelos, la concurrencia (5, 4 y 2) ni
  el tope de 352 minutos. Cada control nuevo se ve en rojo antes de darlo por bueno, con el mutante de su fila aplicado
  en un árbol aparte bajo `$TMPDIR` (`git worktree add --detach`, nunca dentro del repositorio) o en memoria dentro del
  propio test, y retirado al terminar; ninguno queda en el diff.
- **Lo que no cambia** (FR-026, FR-027; plan.md, «Constraints»): `scripts/workflow/`, `.specify/`,
  `internal/workflow/` (SC-015: ninguna tarea declara esas rutas, y el guardián de diff lo hace cumplir), `cmd/`,
  `internal/app`, `internal/cli`, `internal/core`, `internal/source`, `internal/render`, `internal/mcp`, `skills/`,
  `evals/`, `data/`, el `Makefile`, los fixtures y esquemas del repositorio, `docs/SOURCES.md`, `docs/ROADMAP.md`,
  `CLAUDE.md`, `web/` y los specs de otros hitos. En `internal/skills` solo cambian `comandos.go` y `referencias.go`
  (T003).
- **Ningún test escribe fuera de `t.TempDir()`**; lo temporal de una tarea va bajo `$TMPDIR`, nunca bajo `/tmp`. El
  cálculo con `go list` fija `GOOS=linux GOARCH=amd64 CGO_ENABLED=0` (research D3, V2) para dar lo mismo en macOS y en
  el runner.
- **Sin atajos** (plan.md, «Comprobación contra la rúbrica», g): ningún `//nolint`, `t.Skip`, TODO ni error
  silenciado; ningún test se desactiva ni pierde lo que comprobaba: a los que cambian solo se les añade lo nuevo, salvo
  lo que plan.md, «Tests existentes que cambian lo que esperan», dice que cambia. Si `dupl`, `gocyclo`, `paralleltest` o
  `misspell` marcan algo, se reestructura o se renombra (una función por subtest desde una tabla; una palabra española
  que `misspell` marca va a `ignore-rules`).
- **Antes de tocar una función, una constante, un tipo o un campo que usan los tests**, `grep -n` de sus usos en todos
  los `_test.go` de `internal/evals` y de la raíz: lo que haya que alinear entra en la misma tarea, y sus ficheros están
  en sus rutas.
- **Nada exportado nuevo en `internal/evals`** (plan.md, principio V); en `internal/skills`, solo `AppletsDeLaTabla` y
  `DatosDeLaReferencia`. Ningún import de `net/http`, `database/sql` ni del SDK de MCP; ningún `os.Exit` ni escritura en
  `os.Stdout` (reglas R2 a R5 de `internal/arch_test.go`).

---

## Fase 1: Setup

Ninguna tarea: el módulo, el `Makefile`, los gates y el paquete `internal/evals` existen y `make ci` pasa.

---

## Fase 2: Fundación — lo que fija FR-026 y la regla de la tasa

**Propósito**: fijar el informe de hoy antes de tocar `informe.go`, y tener la regla k(*n*) que usan el informe con
reutilizadas (T007) y el ADR (T016).

- [X] T001 Golden del informe de la suite completa, antes de tocar `informe.go` — FR-026, SC-012 (plan.md, «Orden de implementación», paso 1; contracts/informe-del-job.md §1). **`internal/evals/informe_test.go`**: `TestInformeDeLaSuiteCompletaNoCambia` compone, con los datos de prueba con que ya componen el informe los tests de este fichero (evals, sesiones sintéticas y juicio del juez en `t.TempDir()`, con todas las sesiones de un plan entero de una skill con juez, en sus dos modos), el `informe.json` y el `informe.md` de una ejecución completa con el código de hoy, y los compara byte a byte con dos constantes del test (el golden, sacado de la salida de hoy y pegado una vez; sin fichero en disco). La comparación quita antes, de la salida, el bloque `origen` del JSON y la sección «Medidas y reutilizadas» del Markdown (no existen todavía: hoy no quita nada); con cualquier otra diferencia sale en rojo diciendo la primera línea distinta. Mutante (árbol aparte): cambiar un carácter de una descripción de umbral de `umbrales.go` → rojo. Rutas: internal/evals/informe_test.go
- [X] T002 La regla de la tasa: *p*ref derivada de `clases.yaml` y el límite k(*n*) para cualquier número de respuestas medidas — FR-024, SC-010, SC-013 (data-model §9; contracts/informe-del-job.md §4; research D12). **`internal/evals/tasa.go`**: *N* de cada skill con juez (las respuestas del modelo que decide en sus dos modos que juzga una ejecución completa, del plan de la skill), *K* = el mayor entero con *K*/*N* ≤ `umbral` de su clase en `evals/<skill>/juez/clases.yaml`, *p*ref = la mayor tasa, redondeada hacia abajo a centésimas de punto porcentual, con P(Bin(*N*, *p*ref) > *K*) ≤ 0,02, y k(*n*) = el *k* más estricto con P(Bin(*n*, *p*ref) > *k*) ≤ 0,02. **`internal/evals/umbrales.go`**: el umbral `<clase>:<modelo>` de una clase que decide se cumple si `medida ≤ k(n)`, con *n* las respuestas medidas del modelo que decide en los dos modos, y publica como `umbral` el de `clases.yaml` si el mayor *m* con *m*/*n* ≤ `umbral` es k(*n*), y si no k(*n*)/*n* (contracts/informe-del-job.md §4, para que el informe final del workflow (`informe.sh`, sin cambios) no lo marque INCOHERENTE); con todas las respuestas, lo de hoy. **`internal/evals/informe.go`** (redelimitada en el intento 1, gates/tarea-T002.md): el juicio del juez lleva *N*, las respuestas que el plan de las evals del informe da a juzgar en sus modos, porque `umbrales.go` no recibe las evals y sin *N* no hay límite; nada más cambia en ese fichero. **Tests**: `internal/evals/tasa_test.go`, `TestLimiteDeLaTasa`: para cada skill con juez leída del repositorio y cada *n* de 1 a *N*, P(Bin(*n*, *p*ref) > k(*n*)) ≤ 0,02 calculada con racionales exactos (`math/big`), y k(*N*) igual al límite de hoy (3 de 120 en `boe-legislacion`, 1 de 60 en `jurisprudencia`, que el test saca de `clases.yaml` y del plan, no de una constante), con `t.Log` de *p*ref, *N* y k(*N*) de cada skill (las cifras que escribe T016); mutante en memoria, sin tocar el fichero: el `clases.yaml` de `boe-legislacion` con `umbral: 0.017` → k(120) = 2, y el test sale en rojo si no lo es; los tests del umbral conjunto de una clase en `umbrales_test.go` e `informe_test.go` con menos respuestas que las de su plan entero esperan k(*n*) (plan.md, «Tests existentes que cambian»), con un caso de 60 respuestas de `boe-legislacion` (k(60) = 2: 1 marcada cumple con `umbral` 1/30). `TestInformeDeLaSuiteCompletaNoCambia` (T001) sigue en verde sin tocar su golden. Rutas: internal/evals/tasa.go, internal/evals/umbrales.go, internal/evals/informe.go, internal/evals/informe_test.go

**Checkpoint**: `make ci` en verde; el informe de la suite completa, fijado; la tasa, con su regla y su test.

---

## Fase 3: US1 — Una propuesta mide las evals que ha podido cambiar, y nada más (P1) 🎯 MVP

**Objetivo**: cada eval tiene su conjunto de dependencias calculado sin listas a mano, y la elección de una propuesta
da exactamente las evals cuyo conjunto toca, con su motivo.

**Test independiente**: `TestConjuntoDeCadaEval` y `TestEleccionDeLosMutantes` en verde; los mutantes de SC-003 dan
eval por eval lo esperado (quickstart.md §1 y §2).

- [X] T003 [US1] Las dos lecturas de `internal/skills` que usa el conjunto, sin duplicar sus formatos — FR-001(c)(d), FR-005 (research D11; plan.md, principio V). **`internal/skills/comandos.go`**: `AppletsDeLaTabla(skillMD []byte) ([]string, error)`, los applets (`kitlegal <applet> …`) de la tabla de comandos de un `SKILL.md`, entre las marcas con que `make skills-sync` la genera, sin repetir y en orden; error con una tabla sin marcas. **`internal/skills/referencias.go`**: `DatosDeLaReferencia(referencia []byte) (string, bool)`, el `data/*.yaml` de la cabecera `<!-- generado desde data/<fichero>, no editar -->` de una referencia generada, y `false` sin esa cabecera. **Tests**: `TestAppletsDeLaTabla` en `internal/skills/comandos_test.go`, con el `SKILL.md` de cada una de las tres skills del repositorio (los applets que su tabla nombra, sacados de la propia tabla con una sonda independiente, no escritos a mano) y con una tabla sin marcas (error); `TestDatosDeLaReferencia` en `internal/skills/referencias_test.go`, con cada referencia de `skills/*/references/` (las generadas dan su `data/*.yaml`, que existe; las demás, `false`) y con un texto sin cabecera. Ninguna otra función de `internal/skills` cambia, y nada de red entra en sus importaciones. Rutas: internal/skills/comandos.go, internal/skills/referencias.go
- [X] T004 [US1] El conjunto de dependencias de cada eval, lo común y las rutas que miden — FR-001 a FR-005, SC-002 (data-model §1 a §4; contracts/eleccion-y-huella.md §1 y §2; research D2 a D5). **`internal/evals/dependencias.go`** (nada exportado; ningún tipo llamado `Conjunto`, que ya existe): las rutas que miden las evals (FR-004), como el predicado fijo de data-model §1 escrito en `dependencias.go` —una lista fija de rutas en el código (research D4), sin leer el disco ni la definición del job: el `case` del trabajo `cambios` de `evals.yml` que lista hoy esas rutas lo sustituye T013 por `TestEleccionDeLaPropuesta`, que llama a este predicado—; lo común (FR-002, data-model §2); el conjunto de cada eval por entradas (data-model §3): (a) su fichero y lo de su skill que la eval usa, (b) las skills que nombra y las suyas, (c) las referencias y los `data/*.yaml` de que salen (`skills.DatosDeLaReferencia`), (d) los paquetes Go de los applets de su tabla de comandos (`skills.AppletsDeLaTabla` y `go list -deps` con `GOOS=linux GOARCH=amd64 CGO_ENABLED=0`, sin sus ficheros de test ni sus directorios de fixtures, research D4), (e) sus grabaciones y (f) su juez; el árbol de la cabeza y el de la base (research D5: base local con `git`, con el cálculo sobre el árbol de cada uno). Si una parte no se puede calcular —una skill nombrada que no existe, un applet de la tabla sin paquete— el error nombra la eval y la causa (FR-005), y quien lo llama elige todas con `sin_conjunto`. **Tests**: `internal/evals/dependencias_test.go`, `TestConjuntoDeCadaEval`: cada eval de `evals/*/` del repositorio con conjunto calculable, no vacío y con sus seis partes; lo que FR-001 pide y excluye para una eval de `legal-core` y una de `jurisprudencia` (las rutas concretas, sacadas del repositorio por el test, no escritas de memoria); mutante en memoria con un applet inexistente en la tabla de una skill → error de esa eval, y el test sale en rojo si el cálculo lo acepta; el predicado de data-model §1 con una ruta de cada clase (una que mide, `internal/workflow/…`, el `_test.go` y el directorio de fixtures de un paquete de un applet, una raíz de grabaciones de §3(e) y un fichero de `docs/`), cada una con lo que el predicado da. Rutas: internal/evals/dependencias.go
- [X] T005 [US1] La elección de una propuesta y su motivo — FR-002 a FR-008, SC-003 (data-model §5; contracts/eleccion-y-huella.md §3 y §5; research D19). **`internal/evals/eleccion.go`**: dadas las rutas cambiadas de una propuesta (FR-007: los ficheros de la propuesta con sus nombres anteriores en un renombrado, cada ruta con las dos) y los conjuntos de cabeza y base, elige: las evals de cuyo conjunto cambia alguna ruta (`conjunto`), todas con lo común (`comun`), con un cambio del `description` de la cabecera de cualquier `SKILL.md` entre base y cabeza (`descripcion`, FR-003; el cuerpo no cuenta como descripción), con una ruta de las que miden que no está en ningún conjunto ni en lo común (`sin_asignar`, FR-004) o con un conjunto incalculable (`sin_conjunto`, FR-005), y con los ficheros de la propuesta sin leer (`plataforma`); cada eval elegida con su motivo y las rutas que la eligen, en el orden y la forma de contracts/eleccion-y-huella.md §3 (FR-008); una propuesta que no toca nada de lo que miden las evals no elige ninguna. **Tests**: `internal/evals/eleccion_test.go`, `TestEleccionDeLosMutantes`, un subtest por mutante de SC-003, FR-002, FR-003, FR-004, FR-006 y FR-007 sobre el repositorio, con los nombres de quickstart.md §2: `referencia-de-legal-core` (la referencia de `legal-core` → sus 12 evals y 0 de las demás); `cuerpo-de-boe-legislacion` (el cuerpo de `skills/boe-legislacion/SKILL.md` → sus 23 más las 12 de `legal-core` y las 10 de `jurisprudencia`); `cuerpo-de-jurisprudencia` (FR-003: solo el cuerpo de `skills/jurisprudencia/SKILL.md` → sus 10 y ninguna más); `internal-cli` (FR-002: un fichero de `internal/cli` → las 45); `sqlite-en-go-mod` (FR-002 y regla 2 de data-model §5: la versión de `modernc.org/sqlite` en `go.mod` → las 45); `descripcion-de-legal-core` (FR-003: solo la descripción de `legal-core` → las 45); `paquete-sin-asignar` (FR-004: un paquete nuevo bajo `internal/` del que no depende ningún applet ni lo común → las 45, y el resultado nombra la ruta); `solo-docs` (FR-004: un cambio solo en `docs/` → 0); `una-eval` (FR-006: `evals/jurisprudencia/03-por-materia.yaml` → esa eval y ninguna más); y `referencia-renombrada` (FR-007: un renombrado de la referencia de `legal-core`, con las dos rutas → sus 12). El esperado sale del listado de `evals/<skill>/`, no de los conjuntos; cada subtest compara eval por eval y sale en rojo con una sola de diferencia. **Corrección de un test ajeno que `make ci` deja en rojo bajo carga** (gates/tarea-T005.md): `TestEjecutarSesionesConUnError` (`internal/evals/sesiones_test.go`, de H7.3) exige que la primera sesión esté abierta cuando el fallo de la tercera cierra el reparto, pero la tercera se abre en cuanto la segunda termina su segundo de espera, y bajo la carga de `test-integration` la primera, que se abre a la vez, puede no haber llegado aún a su `claude`; el sustituto de `claude` de `internal/evals/sustitutos_test.go` gana `<sesión>.tras` en el directorio de transcripts —la ruta que tiene que existir para que empiece su espera— y el método `empezarLaEsperaCuandoExista`, y el test deja a la segunda sesión esperando a la marca `espera-empezada` de la primera, que duerme 5 s, con la espera de la base a 0; con 2 s de retraso de la primera sesión antes de `ejecutarElGuion` (se restaura antes de `make ci`), el test de hoy falla por «la sesion …-01 se abre» y el corregido pasa. Rutas: internal/evals/eleccion.go, internal/evals/sesiones_test.go, internal/evals/sustitutos_test.go

**Checkpoint**: la elección es calculable y comprobada; el job aún no la usa (T013).

---

## Fase 4: US2 — Lo que no cambia reutiliza el resultado de `main` por su huella (P1)

**Objetivo**: cada eval tiene una huella; lo no elegido busca en las medidas de `main` un resultado con su huella y lo
reutiliza sin abrir sesión; el informe separa lo medido de lo reutilizado sin cambiar el veredicto ni la serie.

**Test independiente**: `TestDecisionDelJob` (SC-006, SC-007) y `TestInformeConReutilizadas` en verde (quickstart.md
§3 a §6).

- [ ] T006 [US2] La huella de cada eval — FR-009, SC-008 (data-model §6; contracts/eleccion-y-huella.md §4; research D6, D7). **`internal/evals/huella.go`**: el texto canónico de contracts/eleccion-y-huella.md §4 —el modelo que decide, la versión de Claude Code de las sesiones, el modelo del juez y su versión, y el sha256 de cada ruta de su conjunto y de lo común, en orden; de `go.mod` y `go.sum`, las versiones de los módulos de los paquetes de su conjunto o de lo común, y el resto como una ruta de lo común (research D7)— y su sha256 en 64 caracteres hexadecimales; mismo resultado en cualquier directorio (rutas relativas a la raíz, sin fechas ni permisos). **Tests**: `internal/evals/huella_test.go`, `TestHuellaDeLasEvals`, un subtest por clase de ruta, sobre una copia del repositorio en `t.TempDir()`: cada clase (fichero de la eval, referencia, `data/*.yaml`, paquete de un applet, grabación, juez, descripción, lo común) cambia la huella de las evals que elige y no la de las demás; la versión en `go.mod` de un módulo de lo común, todas; un `replace` y una línea de `go.sum` de un módulo de ningún conjunto, todas; un fichero de `docs/`, ninguna; el modelo que decide o el del juez, todas; dos clones locales (`git clone` en `t.TempDir()`), las mismas huellas. Sale en rojo si una ruta que elige no cambia la huella o una que no elige la cambia. Rutas: internal/evals/huella.go
- [ ] T007 [US2] El bloque `origen`, los aportes, los umbrales `:reutilizadas` y el veredicto con lo reutilizado — FR-011, FR-015, FR-022, FR-023, FR-026, SC-009, SC-012 (data-model §8 y §9; contracts/informe-del-job.md §1 a §6; research D9, D10). **`internal/evals/origen.go`**: el bloque `origen` (`modo`, `sin_conjunto`, `error_de_la_plataforma`, `evals_medidas`, `evals_reutilizadas`, `sesiones_abiertas`, y por eval `eval`, `huella`, `medida`, `motivo` y `rutas` o `procedencia`, `aportes` y, de una reutilizada, sus `tasas`), siempre presente, en orden de fichero; los aportes de cada eval a cada umbral de las respuestas (todos menos duración y medida del juez, los de total mayor que 0). **`internal/evals/informe.go`** y **`internal/evals/umbrales.go`**: componen el informe con `origen`; los campos de hoy dicen solo lo medido (FR-023); junto a cada umbral de las respuestas al que aportan las reutilizadas, `<nombre>:reutilizadas` con la suma de sus aportes (`decide` el del original, salvo `<clase>:<modelo>:reutilizadas`, que no decide); el veredicto es `fallo` además con una serie que decide de una reutilizada con `pasa: false` o un `:reutilizadas` que decide sin cumplir, cada uno su motivo; `<clase>:<modelo>` cuenta solo lo medido con k(*n*) de T002; `informe.md` gana la sección «Medidas y reutilizadas» de §6 con su tabla de series reutilizadas. **Tests**: `internal/evals/informe_test.go`, `TestInformeConReutilizadas`: veredicto `fallo` con una serie reutilizada en fallo y con una respuesta reutilizada sin activar; los `:reutilizadas` y su `decide`; la suma de los aportes de las reutilizadas es la `medida` y el `total` de su `:reutilizadas`, y la de todas las evals, la del umbral medido más la del reutilizado; la tasa con solo lo medido; `TestInformeDeLaSuiteCompletaNoCambia` (T001) sigue en verde, ahora quitando de verdad `origen` y la sección nueva; `internal/evals/serie_test.go`, `TestFilaConReutilizadas`: ejecuta tal cual el guion de la serie del workflow (`serie.sh fila`, en su sitio y sin copiarlo) sobre un informe con reutilizadas escrito en `t.TempDir()`, que sale con 0 y da byte a byte la fila del mismo informe con solo lo medido. Rutas: internal/evals/origen.go, internal/evals/informe.go, internal/evals/umbrales.go, internal/evals/serie_test.go
- [ ] T008 [US2] La plataforma: interfaz, su implementación con `gh` y el informe de un registro — FR-007, FR-010, FR-025 (contracts/job-de-evals.md §5; research D8, D15, V7). **`internal/evals/plataforma.go`**: una interfaz sin exportar con las lecturas que necesitan la elección, la búsqueda y la release —los ficheros de la propuesta (con su nombre anterior), las propuestas cerradas sobre `main` fusionadas (número, cabeza y fecha de fusión), las ejecuciones del flujo `evals` (todas, y las de un commit), los trabajos de una ejecución con sus pasos y conclusiones, y el registro de un trabajo—, implementada con las órdenes de contracts/job-de-evals.md §5, todas de lectura, con texto constante y cada valor en una variable del entorno, reutilizando la ejecución de `gh` de `tanda.go` sin duplicarla; del registro de un trabajo, quita trabajo, paso, hora y marca de orden de bytes y toma lo que hay entre `--- inicio de informe.json ---` y `--- fin de informe.json ---`, como `informe_del_trabajo` del guion común del workflow (`comun.sh`, sin cambios); un error de una orden lleva la orden y su salida de error. **Tests**: `internal/evals/plataforma_test.go`, `TestOrdenesDeLaPlataforma` (con el `gh` sustituto de la tanda: cada orden con sus argumentos y variables exactas, y el error con la orden y su salida de error) y `TestInformeDelRegistro` (un registro sintético con prefijos de trabajo, paso y hora y con marca de orden de bytes → el informe; sin marcas o con dos informes → error). Los tests de la tanda siguen en verde sin cambiar lo que esperan. Rutas: internal/evals/plataforma.go, internal/evals/tanda.go, internal/evals/sustitutos_test.go
- [ ] T009 [US2] La búsqueda de resultados en las medidas de `main` — FR-010 a FR-013 (data-model §7 y §8; research D8, D9, D17). **`internal/evals/reutilizacion.go`**: las medidas de `main` —ejecuciones del flujo `evals` terminadas, no canceladas, sobre la cabeza de una propuesta fusionada en `main` (el `headSha` de la ejecución es la cabeza con que se fusionó), con el trabajo de la skill con la marca «El informe del trabajo está escrito» en `success`— de la más reciente a la más antigua, como mucho 10; de cada una, el informe de la skill por el registro de su trabajo (T008), y de cada eval no elegida que no tenga aún resultado, el de la primera medida con su misma huella; no se reutiliza una eval con una sesión sin medir o una respuesta sin juzgar; la procedencia de un resultado que ya era reutilizado en esa medida es la original, y una procedencia de 90 días o más antes del instante de la decisión deja la eval sin resultado (`sin_resultado_de_main`); la búsqueda para en cuanto cada eval tiene resultado; la primera orden de la plataforma que falla, o el plazo de 10 minutos agotado (con la orden en curso), la para, y lo que queda se mide con `sin_resultado_de_main` y el error en `origen.error_de_la_plataforma`. **Tests**: `internal/evals/reutilizacion_test.go`, `TestBusquedaDeMedidasDeMain`, con la plataforma en memoria y un instante fijo, un subtest por caso: una ejecución cancelada, una de un commit que no es la cabeza fusionada, una sin la marca, una de otro commit de la misma propuesta y una procedencia de 90 días o más no cuentan; la más reciente gana; la undécima no se recorre; una sesión sin medir o una respuesta sin juzgar no se reutiliza; la procedencia de una reutilizada en `main` es la original; una orden que falla y el plazo agotado paran la búsqueda, lo que queda se mide y la orden y su error quedan en el resultado. Rutas: internal/evals/reutilizacion.go
- [ ] T010 [US2] La decisión de cada trabajo y de la apertura, el plan con solo lo medido y sus puntos de entrada — FR-012 a FR-014, FR-016, FR-017, FR-019, SC-006, SC-007 (data-model §5 a §8; contracts/job-de-evals.md §2 a §4; research D13, D17, D18). **`internal/evals/decision.go`**: para un trabajo de una skill, el modo (`completa`: todas medidas con `suite_completa`, sin búsqueda; `reutilizacion`: las elegidas se miden y no se buscan, FR-012, y las demás se buscan con T009 y, sin resultado, se miden con `sin_resultado_de_main`, FR-014), con los ficheros de la propuesta de T008 (si `gh` no los da, todas las de la skill medidas con `plataforma` y sin búsqueda) y las huellas de T006; escribe el fichero de la decisión (el bloque `origen` sin `sesiones_abiertas`) y, para el registro, el error de la plataforma y una línea por eval —medida con su motivo o reutilizada con su procedencia— (contracts/eleccion-y-huella.md §3); la decisión de la apertura, `elige=si|no` (FR-016: sin ninguna eval elegida, `no`; con un conjunto incalculable o sin los ficheros de la propuesta, `si`). **`internal/evals/plan.go`** y **`internal/evals/ejecucion.go`**: el plan y la ejecución abren solo las sesiones de las evals medidas de la decisión y componen el informe (T007) con las reutilizadas de la decisión. **`internal/evals/job_test.go`** (etiqueta `evals`): `TestEleccionDeLaPropuesta` (`-propuesta`, `-base`, `-salida`), `TestDecisionDelTrabajo` (`-skill`, `-modo`, `-commit`, `-propuesta`, `-base`, `-modelo-que-decide`, `-version-de-claude-code`, `-modelo-del-juez`, `-version-del-juez`, `-salida`, con plazo de 10 minutos para la plataforma) y `-decision` en `TestEjecucionDelJob`, como en contracts/job-de-evals.md §4; sin `-decision`, `TestEjecucionDelJob` lo mide todo como hoy (la obligación la pone `scripts/evals.sh` en T013). **Tests**: `internal/evals/decision_test.go`, `TestDecisionDelJob`, con la plataforma en memoria y medidas de `main` sintéticas con la huella de cada eval, un subtest por caso: `solo-el-cuerpo-de-jurisprudencia` (SC-006: 0 sesiones planificadas de `boe-legislacion` y de `legal-core`, todas sus evals reutilizadas; sin esa medida, FR-014 las mide), `etiqueta-evals-sin-cambios` (SC-007: 0 sesiones y las 45 evals reutilizadas con su huella), solo la eval sin resultado medida con su motivo (FR-014), una apertura sin elección no planifica nada (FR-016), la suite completa sin ninguna búsqueda, y los ficheros de la propuesta que `gh` no da (todas las de la skill medidas con `plataforma` y el error en `origen`). Cada subtest cuenta las sesiones del plan de cada trabajo y sale en rojo con una sola de más. `go vet -tags evals ./internal/evals/` en verde. Rutas: internal/evals/decision.go, internal/evals/plan.go, internal/evals/ejecucion.go, internal/evals/job_test.go
- [ ] T011 [US2] El golden de la matriz: lo que mide y reutiliza cada skill — FR-008, FR-019, SC-004 (contracts/eleccion-y-huella.md §3 y §5; research D16). **`internal/evals/eleccion.go`** (y **`decision.go`** si la composición lo pide): la matriz de una ejecución en texto, por skill, con las evals que mide (y su motivo) y las que reutiliza, sin sesiones, en un orden estable. **Tests**: `internal/evals/eleccion_test.go`, `TestMatrizDeLasPropuestas`: con una medida de `main` sintética con la huella de cada eval, la matriz de cada mutante de T005, de una propuesta sin nada elegido y de la suite completa, comparada byte a byte con una constante del test (el golden, sin fichero en disco); sale en rojo con cualquier diferencia. Rutas: internal/evals/eleccion.go, internal/evals/decision.go

**Checkpoint**: el arnés decide y compone el informe con lo reutilizado; el job aún lo mide todo (T013 lo conecta).

---

## Fase 5: US4 — La suite completa a petición, y el job que lo junta todo (P2)

**Objetivo**: la tanda distingue el modo; `evals-completa` y el despacho miden la suite completa; el job elige,
decide, reutiliza y deja la marca del informe sin abrir más sesiones a la vez que hoy. Va antes de US3 porque la
release exige una ejecución en modo `completa`, que solo existe tras T013.

**Test independiente**: `TestTandaConLaSuiteCompleta` y `TestDefinicionDelJob` en verde (quickstart.md §8).

- [ ] T012 [US4] El modo de cada tanda — FR-018, FR-021 (data-model §10; contracts/job-de-evals.md §2, `tanda`; research D13). **`internal/evals/tanda.go`**: la tanda recibe el modo (`completa` o `reutilizacion`), escribe `medir=si|no` y `modo=<modo>`, y lee de cada ejecución anterior sin terminar sobre el mismo commit sus dos marcas («Esta ejecución mide el commit» y «Esta ejecución mide la suite completa»): no mide si una de ellas mide con un modo igual o mayor (`completa` > `reutilizacion`). **`internal/evals/job_test.go`**: `-modo` en `TestTandaDelCommit`. **Tests**: `internal/evals/tanda_test.go`, `TestTandaConLaSuiteCompleta`: una con reutilización no mide con una completa en curso; una completa sí mide con una con reutilización en curso; y los tests de la tanda que ya existen (`TestOrdenesDeLaTanda`, `TestSalidaDeLaTanda`, `TestRegistroDeLaTanda`) ganan el modo sin perder lo que comprobaban; `internal/evals/definicion_test.go`: los subtests `segundo-disparo` y `estado-de-la-tanda` de `TestDefinicionDelJob`, que usan lo que esta tarea cambia en `tanda.go` (`esperarLaDecision`, `leerLaTandaDeLaEjecucion`, `decisionDeLaTanda`, `decisionTrasLaEspera`, `ejecucionDelCommit`, `estadoDeLaTanda` y sus valores `tandaQueMide`, `tandaQueNoMide` y `tandaSinDecidir`), ganan el modo (data-model §10) sin perder lo que comprobaban, para que `make ci` compile y siga en verde en esta tarea. `go vet -tags evals ./internal/evals/` en verde. Rutas: internal/evals/tanda.go, internal/evals/job_test.go, internal/evals/definicion_test.go
- [ ] T013 [US4] El job de evals con elección, decisión y reutilización, y `scripts/evals.sh` con la decisión obligatoria — FR-010, FR-012, FR-016 a FR-020, SC-011 (y SC-016 en la cabecera del job) (contracts/job-de-evals.md §1 a §3; research D13, D17, D18). **`.github/workflows/evals.yml`**: el trabajo `cambios` con `fetch-depth: 0`, `permissions: contents: read, pull-requests: read`, `timeout-minutes: 15`, salida `elige` y el paso «Elegir las evals que la propuesta ha podido cambiar» con `TestEleccionDeLaPropuesta`; la `tanda` con `MODO_DE_EVALS`, la condición de §2 y las salidas `medir` y `modo`; en `evals (<skill>)`, `permissions: contents: read, actions: read, pull-requests: read`, `MODO_DE_EVALS`, `PROPUESTA` y `BASE_DE_LA_PROPUESTA` en el `env`, `fetch-depth: 0`, el paso «Decidir qué mide y qué reutiliza» tras instalar Go, con `GH_TOKEN` y `GH_REPO` solo en él, «Ejecutar las evals» con `DECISION_DE_EVALS`, y el último, «El informe del trabajo está escrito», con `if: ${{ !cancelled() }}`; la etiqueta `evals-completa` entre los disparadores; matriz, concurrencia (5, 4 y 2), modelos, repeticiones y tope (352 min), como hoy; la cabecera dice qué mide cada disparador y que la release exige la suite completa. **`internal/evals/definicion.go`**: lee lo nuevo (el trabajo `cambios`, el modo de la tanda, los permisos, el `fetch-depth`, el paso de la decisión y su plazo, la marca final) y el peor caso de cada trabajo suma los 900 s de la decisión (`boe-legislacion`: 19 877 s, en los 352 minutos). **`scripts/evals.sh`**: sin `DECISION_DE_EVALS`, o si no es un fichero legible, termina con 1 antes de la primera sesión con «evals: falta DECISION_DE_EVALS» (la comprobación va tras leer sus argumentos y antes de cualquier otra cosa que dependa del entorno); con ella, la pasa como `-decision` a `TestEjecucionDelJob`. **Tests**: `internal/evals/definicion_test.go`, `TestDefinicionDelJob`: el trabajo `cambios` con `TestEleccionDeLaPropuesta`; modo `completa` solo con el despacho, `evals-completa` y `evals-prueba-de-red`; la tanda solo con uno de ellos, `evals` o `elige=si`; los permisos, el `fetch-depth`, el paso de la decisión antes de «Ejecutar las evals» con `GH_TOKEN` solo en él, `DECISION_DE_EVALS` y la marca final con `!cancelled()`; una combinación por skill con 5, 4 y 2 y 11 en total (SC-011: rojo si la definición admite más); el peor caso con la decisión dentro del tope; sus subtests `segundo-disparo` y `estado-de-la-tanda` ya vienen con el modo de T012, y esta tarea solo les añade lo que lea de nuevo `definicion.go`; `internal/evals/sustitutos_test.go`: un test nuevo (hoy ningún test de `make ci` ejecuta `scripts/evals.sh`) que lo ejecuta con el `claude` sustituto delante en el `PATH` y comprueba que sin `DECISION_DE_EVALS`, y con una ruta que no es un fichero legible, sale con 1 y «evals: falta DECISION_DE_EVALS» sin que llegue ninguna sesión al sustituto; mutante (árbol aparte): sin la comprobación en el guion → rojo. Rutas: .github/workflows/evals.yml, internal/evals/definicion.go, scripts/evals.sh, internal/evals/sustitutos_test.go

**Checkpoint**: el job entero funciona con elección y reutilización; US1, US2 y US4 completas.

---

## Fase 6: US3 — La release exige la suite completa del commit etiquetado (P1)

**Objetivo**: `release.yml` no publica nada sin una ejecución de la suite completa del commit etiquetado en verde.

**Test independiente**: `TestComprobarLaSuiteCompleta` y el subtest `suite-completa` de `TestConfiguracionDeLaRelease`
en verde, con cada mutante en rojo (quickstart.md §9).

- [ ] T014 [US3] La comprobación de la suite completa de un commit — FR-025, SC-005 (data-model §11; contracts/release.md §2 y §3; research D14). **`internal/evals/suite_completa.go`**: para el commit *c*, de las ejecuciones del flujo `evals` sobre *c* (T008), terminadas y de la más reciente a la más antigua, vale la primera con `headSha` *c*, todos los trabajos `evals (<skill>)` de las skills con evals en *c* en `success`, y en el informe de cada una `origen.modo: completa`, cada eval de `evals/<skill>/` en *c* medida y ninguna reutilizada; si ninguna vale, el error con el mensaje de contracts/release.md §2 (el commit, una línea por ejecución mirada con su primer motivo o «ninguna ejecución del flujo evals sobre el commit», y cómo lanzar la suite completa sobre la etiqueta). **`internal/evals/job_test.go`**: `TestSuiteCompletaDelCommit` (`-commit`, `-etiqueta`; etiqueta `evals`). **Tests**: `internal/evals/suite_completa_test.go`, `TestComprobarLaSuiteCompleta`, con la plataforma en memoria: falla sin ejecución, con otro `headSha`, con un trabajo en `failure`, con un informe que falta, con una eval reutilizada, con modo `reutilizacion` y con una eval del commit que no está en el informe, cada uno con su motivo en el mensaje; sale bien con una completa y en verde. `go vet -tags evals ./internal/evals/` en verde. Rutas: internal/evals/suite_completa.go, internal/evals/job_test.go
- [ ] T015 [US3] El trabajo `suite-completa` de la release, del que depende `publicar` — FR-025, SC-005 (contracts/release.md §1 y §3). **`.github/workflows/release.yml`**: el trabajo `suite-completa` de contracts/release.md §1 (`ubuntu-24.04`, `permissions: contents: read, actions: read`, `GH_TOKEN`, `GH_REPO`, `COMMIT_ETIQUETADO: ${{ github.sha }}`, `ETIQUETA`, checkout sin credenciales persistidas, `setup-go` con `go.mod` y el paso con `TestSuiteCompletaDelCommit`), y `publicar` con `needs: [suite-completa]`; `humo` y `catalogo` siguen detrás de `publicar`. **Tests**: `release_test.go`, subtest `suite-completa` de `TestConfiguracionDeLaRelease`: el trabajo existe con su paso y sus permisos, el paso pasa `${{ github.sha }}` como commit, `publicar` depende de él, y el conjunto de trabajos y su orden lo incluyen; mutantes (en memoria, sobre el YAML leído): sin el trabajo, sin la dependencia, con otro commit en el paso → rojo cada uno. Los demás subtests de `TestConfiguracionDeLaRelease` siguen en verde. Rutas: .github/workflows/release.yml, release_test.go

**Checkpoint**: la release no publica sin la suite completa; US3 completa.

---

## Fase 7: US5 — Un ADR decide cómo cuenta lo reutilizado en el umbral en tasa y en la serie (P2)

**Objetivo**: el ADR 0045 escribe la decisión que T002 y T007 ya aplican, con la derivación de *p*ref y sus cifras.

**Test independiente**: el ADR existe y sus cifras son las que dan `TestLimiteDeLaTasa` y `scripts/workflow/serie.sh`
al ejecutarlos (quickstart.md §7 y §10).

- [ ] T016 [US5] El ADR 0045, lo reutilizado no cuenta en la tasa — FR-024, SC-013 (spec FR-024; data-model §9; research D12). **`docs/ADR/0045-lo-reutilizado-no-cuenta-en-la-tasa.md`**, con la cabecera de los ADR de `docs/ADR/` (estado, fecha, hito H30 y lo que no toca: el umbral de ninguna clase, el modelo que decide, los guiones del workflow): el contexto (la tasa del ADR 0038 con todas las respuestas, y lo que pasa con menos); la decisión de FR-024 (lo reutilizado no cuenta en `<clase>:<modelo>` ni en `<clase>:<modelo>:<modo>`, que se aplican a lo medido del modelo que decide con los dos modos juntos; el límite k(*n*); la serie de cierres cuenta solo lo medido y el guion de la serie del workflow (`serie.sh`) no cambia; ningún umbral sube); la derivación de *p*ref desde el `umbral` de `clases.yaml` y, como su resultado, las cifras de hoy de `boe-legislacion` y `jurisprudencia` (*N*, k(*N*), *p*ref), copiadas de la salida de `go test -count=1 -v -run '^TestLimiteDeLaTasa$' ./internal/evals/` ejecutada en esta tarea; que estrechar el `umbral` de `clases.yaml` baja *p*ref y estrecha el límite para todo *n*; la tasa acumulada de la serie de `boe-legislacion`, sacada de `serie.sh mostrar boe-legislacion` (el guion de la serie del workflow, sin cambios) ejecutado en esta tarea (no copiada del spec), por encima de *p*ref, y que subir un umbral pediría otro ADR; las alternativas rechazadas de research D12. Ninguna cifra de memoria: si la salida difiere del spec, manda la salida, y se anota en `specs/026-h30-las-evals-miden/gates/supuestos.md`. Rutas: docs/ADR/0045-lo-reutilizado-no-cuenta-en-la-tasa.md, specs/026-h30-las-evals-miden/gates/supuestos.md

---

## Fase 8: Pulido — documentación y cierre de la Definition of Done

- [ ] T017 Documentación del job, de sus disparadores y de la release — SC-016 (y FR-016 a FR-018, FR-025, que describe) (`docs/ROADMAP.md` §1). **`CHANGELOG.md`**, bajo *Unreleased*: el job mide solo las evals que una propuesta ha podido cambiar y reutiliza lo demás de `main` por su huella; qué mide cada disparador (apertura, etiqueta `evals`, `evals-completa`, despacho, `evals-prueba-de-red`, `evals-medir-juez`); el informe con `origen` y los umbrales `:reutilizadas`; el límite k(*n*) del ADR 0045; la release exige la suite completa del commit etiquetado; `scripts/evals.sh` exige `DECISION_DE_EVALS`. **`docs/WORKFLOW.md`**: donde nombra la etiqueta `evals` del cierre (`medir_cierre`), qué mide con el job nuevo (las elegidas y lo que no tiene resultado reutilizable; la suite completa, con `evals-completa`). **`CONTRIBUTING.md`**: la tabla de cuándo corre el job (línea de «Por etiqueta» y las demás) con los disparadores y modos de contracts/job-de-evals.md §1, y que antes de etiquetar una release el commit necesita su suite completa en verde (`gh workflow run evals.yml --ref <etiqueta>`). **`internal/evals/doc.go`**: el paquete, con la elección, la huella, la reutilización, la decisión y la comprobación de la release. Cada test, fichero, disparador y cifra que nombran existe en el árbol (`grep -n` de cada uno). Rutas: CHANGELOG.md, docs/WORKFLOW.md, CONTRIBUTING.md, internal/evals/doc.go
- [ ] T018 Cierre de la Definition of Done, sin código de producto — SC-001, SC-015 (`docs/ROADMAP.md` §1; quickstart.md §1 a §11; plan.md, «Controles de umbral»). **`specs/026-h30-las-evals-miden/cierre.md`** con: (1) la lista, sacada de `git diff --name-status main`, de lo creado o modificado, y la salida vacía de la primera orden de quickstart.md §10 (el diff respecto de `main` en los tres directorios que FR-027 excluye; FR-027, SC-015), con la comprobación de que nada de «Lo que no cambia» de la batería está en el diff; (2) cada fila de plan.md, «Controles de umbral», con su test encontrado como `^func <Test>\(` en su ruta y su `PASS` en una ejecución de `go test -count=1 -run '^<Test>$'` de esta tarea (un `-run` por test); (3) los escenarios de quickstart.md §1 a §10 ejecutados, con su resultado; (4) la cobertura sobre `coverage.out` y `coverage-integration.out` de `make ci` con `go tool cover -func` —global ≥ 70 % e `internal/core/**` ≥ 85 % (Definition of Done §1.9)— con la cifra de cada una y la de `internal/evals`, y, si alguna queda por debajo, el hueco (fichero y función); (5) `make ci` en verde (quickstart.md §11). Lo que no se puede medir sin red ni plataforma (SC-014, la medición de cierre) se dice tal cual: lo hace el workflow tras la revisión final. Rutas: specs/026-h30-las-evals-miden/cierre.md

---

## Dependencias y orden de ejecución

Todo va en secuencia, T001 → T018; ninguna tarea depende de una posterior.

| Tarea | Necesita | Por qué |
|---|---|---|
| T001 | — | El golden se fija con el `informe.go` de hoy, antes de que T002 y T007 lo toquen |
| T002 | T001 | Cambia `umbrales.go`; el golden comprueba que con todas las respuestas nada cambia |
| T003 | — | Lecturas puras de `internal/skills` |
| T004 | T003 | El conjunto usa `AppletsDeLaTabla` y `DatosDeLaReferencia` |
| T005 | T004 | La elección compara conjuntos de cabeza y base |
| T006 | T004 | La huella es la del conjunto de cada eval y de lo común |
| T007 | T002, T006 | `origen` lleva huellas; `<clase>:<modelo>` con k(*n*); el golden de T001 quita `origen` |
| T008 | — (en orden, tras T007) | Interfaz de la plataforma, que usan T009, T010 y T014 |
| T009 | T006, T007, T008 | Busca por huella en informes con `origen`, leídos por la plataforma |
| T010 | T005, T006, T008, T009 | La decisión junta elección, huella, ficheros de la propuesta y búsqueda |
| T011 | T005, T010 | La matriz es lo que mide y reutiliza cada skill según la decisión |
| T012 | — (en orden, tras T011) | Modo de la tanda, que usa el job de T013; adapta los subtests de `TestDefinicionDelJob` que usan la tanda |
| T013 | T010, T012 | El job llama a los puntos de entrada de T010 y T012; `scripts/evals.sh` exige la decisión que T010 escribe |
| T014 | T007, T008, T013 | La release busca ejecuciones en modo `completa`, que existen desde T013, y lee `origen` |
| T015 | T014 | El trabajo de la release llama a `TestSuiteCompletaDelCommit` |
| T016 | T002, T007 | El ADR escribe las cifras de `TestLimiteDeLaTasa` y la decisión que aplica T007 |
| T017 | T013, T015, T016 | Documenta el job, la release y el ADR ya hechos |
| T018 | todas | Cierre |

**Historias**: US1 (T003 a T005) y la fundación (T001, T002) primero; US2 (T006 a T011) sobre US1; US4 (T012, T013)
conecta US1 y US2 en el job; US3 (T014, T015) sobre US4, por el modo `completa`; US5 (T016) sobre la fundación y US2.

### Oportunidades de paralelismo

El bucle del workflow ejecuta una tarea cada vez. Fuera de él, serían independientes entre sí: T001 y T003; T006 y
T008 (tras T004); T016 y T017 en sus partes de
documentación. Por eso ninguna lleva `[P]`.

---

## Trazabilidad

| Requisito | Tareas |
|---|---|
| FR-001 | T003, T004 |
| FR-002, FR-003, FR-004 | T004, T005 |
| FR-005 | T003, T004, T005 |
| FR-006 | T005 |
| FR-007 | T005, T008 |
| FR-008 | T005, T011 |
| FR-009 | T006 |
| FR-010 | T008, T009, T013 |
| FR-011 | T007, T009 |
| FR-012 | T009, T010, T013 |
| FR-013 | T009, T010 |
| FR-014 | T010 |
| FR-015 | T007 |
| FR-016, FR-017 | T010, T013 |
| FR-018 | T012, T013 |
| FR-019 | T010, T011, T013 |
| FR-020 | T013 |
| FR-021 | T012 |
| FR-022 | T007 |
| FR-023 | T007 |
| FR-024 | T002, T007, T016 |
| FR-025 | T008, T014, T015 |
| FR-026 | T001, T007 (y la batería: ninguna tarea toca `evals/` ni `clases.yaml`) |
| FR-027 | ninguna tarea declara `scripts/workflow/`, `.specify/` ni `internal/workflow/`; T018 lo comprueba |
| SC-001 | todas (`make ci` en cada una), T018 |
| SC-002 | T004 |
| SC-003 | T005 |
| SC-004 | T011 |
| SC-005 | T014, T015 |
| SC-006, SC-007 | T010 |
| SC-008 | T006 |
| SC-009 | T007 |
| SC-010 | T002 |
| SC-011 | T013 |
| SC-012 | T001, T007 |
| SC-013 | T002, T016 |
| SC-014 | la medición de cierre del workflow, con el job de T013 (sin tarea: no la hace una tarea) |
| SC-015 | T018 (y el guardián de diff) |
| SC-016 | T013 (cabecera de `evals.yml`), T017 |

### Controles de umbral de plan.md y la tarea que construye cada uno

| Requisito (fila) | Control | Tarea | Lo ve en rojo |
|---|---|---|---|
| SC-001 | `make ci` | todas | cualquier test o control en rojo |
| FR-001, FR-005, SC-002 | `TestConjuntoDeCadaEval` | T004 | el mutante del applet inexistente |
| FR-002 a FR-004, FR-006, FR-007, SC-003 | `TestEleccionDeLosMutantes` | T005 | una eval de diferencia en un mutante |
| FR-008, FR-019, SC-004 | `TestMatrizDeLasPropuestas` | T011 | un byte de diferencia con el golden |
| FR-009, SC-008 | `TestHuellaDeLasEvals` | T006 | una huella cambiada fuera de las elegidas, o sin cambiar entre ellas, o distinta entre clones |
| FR-010 a FR-013 | `TestBusquedaDeMedidasDeMain` | T009 | una medida inválida usada, la undécima recorrida, un resultado incompleto reutilizado, una procedencia que no es la original o de 90 días, un error sin escribir |
| FR-013, FR-014, FR-016, FR-017, SC-006, SC-007 | `TestDecisionDelJob` | T010 | una sesión de más en el plan de un trabajo |
| FR-015, FR-022 | `TestInformeConReutilizadas` | T007 | una serie reutilizada en fallo o una respuesta reutilizada sin activar que no dan `fallo` |
| FR-023, SC-009 | `TestFilaConReutilizadas` | T007 | `serie.sh fila` sin 0 o una fila distinta de la de solo lo medido |
| FR-024, SC-010, SC-013 | `TestLimiteDeLaTasa` | T002 | un *n* con P > 0,02, k(*N*) distinto del de hoy, o el mutante 2 de 120 sin k(120) = 2 |
| FR-016, FR-018, FR-020, SC-011 | `TestDefinicionDelJob` | T013 | un modo o una condición de la tanda que no son los del contrato, o más sesiones a la vez |
| FR-021 | `TestTandaConLaSuiteCompleta` | T012 | una combinación que mide cuando no debe, o no mide cuando debe |
| FR-025, SC-005 (`release.yml`) | `TestConfiguracionDeLaRelease` | T015 | sin el trabajo, sin la dependencia, con otro commit |
| FR-025, SC-005 (comprobación) | `TestComprobarLaSuiteCompleta` | T014 | aceptar otro commit, una ejecución en rojo, una eval reutilizada o un informe que falta |
| FR-026, SC-012 | `TestInformeDeLaSuiteCompletaNoCambia` | T001 (y T007) | un byte de diferencia, sin `origen`, con el informe de antes del hito |

---

## Estrategia de implementación

1. **MVP (US1)**: T001 a T005 —el informe fijado, la regla de la tasa, los conjuntos y la elección—. Comprobable
   con quickstart.md §1 y §2, sin cambiar todavía el job.
2. **US2**: T006 a T011 —huella, informe con lo reutilizado, plataforma, búsqueda, decisión y matriz—. Comprobable con
   `TestDecisionDelJob` (SC-006, SC-007).
3. **US4**: T012 y T013 —el modo de la tanda y el job conectado—. Desde aquí el job elige, decide y reutiliza.
4. **US3**: T014 y T015 —la release exige la suite completa—.
5. **US5 y pulido**: T016 a T018 —ADR, documentación y cierre—.

Cada paso deja `make ci` en verde; la medición de cierre del workflow (SC-014) corre con el job de T013.

---

## Comprobación contra la rúbrica de `juez_tasks` y `precheck.sh tasks`

- `precheck.sh tasks`: cada línea `- [ ] Tnnn` con ids únicos y correlativos y con sus rutas; ninguna nombra los
  directorios de fixtures ni de esquemas (no hay `[datos]`) ni la carpeta de la evidencia; ninguna graba, pide a una
  persona ni toca la plataforma; ninguna `[aceptacion]`, porque plan.md dice «Aceptación e2e: no aplica».
- a (análisis crítico): lo resuelve `gates/analyze.md`, que se escribe después de este fichero.
- b (trazabilidad): cada tarea cita sus FR/SC; cada requisito tiene tarea («Trazabilidad»); cada fila de «Controles de
  umbral» tiene la tarea que construye su control y el mutante o caso que lo ve en rojo (tabla de arriba).
- c (rebanadas verdes): T002 a T015 llevan test e implementación en el mismo diff; T001, T016, T017 y T018, con su
  razón al principio; el orden, en «Dependencias».
- d (rutas declaradas): ficheros concretos en cada tarea, sin directorios genéricos.
- e (datos separados): no hay datos externos ni fixtures; todo en memoria o en `t.TempDir()`.
- f (DoD): CHANGELOG (T017), ADR (T016), `docs/SOURCES.md` (no aplica: ninguna fuente), esquemas (ninguno nuevo),
  fixtures (ninguno), cobertura y cierre (T018).
- g (checklists): no se tocan.
- h (aceptación primero): no aplica, justificado por plan.md (Complexity Tracking; research D20).
- i (autonomía): ninguna tarea exige a una persona, publica la rama ni mide CI o evals remotas; la medición de cierre
  la hace el workflow.
