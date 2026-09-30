# H7.3 · Cierre de la Definition of Done

Cierre de T015 (FR-008, FR-068, FR-070, FR-080, FR-090, FR-099; SC-010) sobre `b9438b2` (T014, el último commit antes
de T015). La rama está al día con `main`: `git merge-base HEAD main` es la cabeza de `main`, `f56c827`, así que
`git diff --name-status main` da exactamente lo que cambia el hito. Las cuatro partes son las de la tarea: los esquemas,
datos de prueba y evals que toca el hito, la cobertura, el quickstart y lo que el hito no cambia.

## 1. Esquemas, datos de prueba y evals tocados en el hito

Lista para la capa 3 del informe final: cada fichero de `schemas/`, de cualquier `testdata/` o de `evals/` que el hito
crea, modifica o retira, con la tarea que lo tocó y el motivo. Sale de `git diff --name-status main`, filtrado a esos
tres árboles, y la tarea de cada fila sale de `git log --format=%s main..HEAD -- <fichero>`. Son dos filas, las dos `M` y
de T001: el esquema de la lista y la lista. No hay ninguna `A`, `D` ni `R`.

### Esquemas (`schemas/`)

| Estado | Fichero | Tarea | Motivo |
|---|---|---|---|
| M | `schemas/expresiones-prohibidas.yaml.json` | T001 `[datos]` | La tercera familia de la lista (FR-020, FR-023; contracts/lista-de-expresiones.md §2): `anuncio` entra en `required`, detrás de `maquinaria` y `otra_conversacion`, y en `properties` con el mismo `$ref` a `#/$defs/expresiones` (lista con `minItems: 1` de cadenas con el patrón de H7.2). `additionalProperties: false` sigue igual. Diff: 3 líneas añadidas y 2 quitadas. No sale de `--describe`, así que `schema-check` no lo compara. Lo fijan los casos de `internal/evals/formato_test.go`: una lista con las tres familias valida, y una sin `anuncio`, con `anuncio: []` o con una expresión con `*` no. |

`schemas/eval.yaml.json`, el formato común de eval, no cambia, y el resto de esquemas de `main` tampoco.

### Datos de prueba (`testdata/`)

Ninguno. `git diff --name-status main -- testdata '*/testdata/*'` no imprime nada, y las dos rutas no pasan en vacío
porque nombran 56 y 480 ficheros versionados. La búsqueda más amplia, `git diff --name-status main -- '*testdata*'
'*.golden' '*.txtar' '*/fuzz/*' '*grabaciones.json'`, tampoco imprime nada: no cambia ninguna grabación, ninguna
derivada, ningún guion `testscript`, ningún manifiesto `grabaciones.json` ni ningún corpus de fuzz (plan.md, «Datos
externos»; research D22). Los sustitutos de `claude`, `strace` y `go`, y también las sesiones y los transcripts
sintéticos, son constantes de los tests escritas en `t.TempDir()`, como pide tasks.md.

### Evals (`evals/`)

| Estado | Fichero | Tarea | Motivo |
|---|---|---|---|
| M | `evals/boe-legislacion/expresiones-prohibidas.yaml` | T001 `[datos]` | La lista de `boe-legislacion` gana la familia `anuncio`, con las 14 expresiones de contracts/lista-de-expresiones.md §1 («que trasladar», «ya puedo responder», «tengo todo lo necesario», «sin redacciones cambiadas»…), detrás de las otras dos, y su comentario pasa a nombrar las tres familias (FR-020). Las otras dos familias no cambian: `maquinaria` sigue con 22 expresiones y `otra_conversacion` con 16. En total, 52 expresiones, frente a las 38 de `main`. Diff: 19 líneas añadidas y 3 quitadas, que son las del comentario. La lista queda calibrada, eval por eval y familia por familia, en la subprueba `expresiones-calibradas` (T002): casa en 35 de las 93 respuestas del cierre de H7.1 (maquinaria 34, otra conversación 1, anuncio 24) y en 10 de las 93 del de H7.2 (10, 0 y 9) (FR-021, FR-095; SC-003). No casa en ningún bloque grabado ni en ninguna forma que enseña la skill (`expresiones-en-los-bloques`, `expresiones-de-la-skill`; FR-022; SC-004). Lo que se hace más estricto es la lista, no el umbral (FR-008, §4). |

El hito no añade, retira ni edita ninguna eval (spec, «Fuera de alcance»): `boe-legislacion` se queda con sus 19 y
`legal-core` con sus 3.

### Lo que no está en la lista

- **`skills/boe-legislacion/SKILL.md`** (M, T013) está fuera de esos árboles. Es la skill v0.1.3, con 270 líneas
  (266 en `main`; 51 añadidas y 47 quitadas), y su prosa la comprueba la subprueba `prosa-de-la-skill` (§3).
- **Sondas y mutantes momentáneos** (T002, T004, T005, T007, T013): se crearon y se retiraron dentro de su tarea.
  `git log --name-only main..HEAD` no nombra ningún fichero con `zz-`, `.orig`, `mutante` ni `sonda` (0 coincidencias).
- **`data/`, `evidencias/`, `web/` y `docs/`** no cambian: `git diff --name-status main -- data evidencias web docs
  README.md` no imprime nada.

## 2. Cobertura (Definition of Done, punto 9)

Medida con las órdenes de quickstart.md §9.3, sobre los dos perfiles que deja el `make ci` de esta tarea:
`coverage.out` (`make test`) y `coverage-integration.out` (`make test-integration`). Cada subtotal sale de
`go tool cover -func` sobre el perfil filtrado a las líneas del árbol, con su cabecera `mode:`. «Unión» es la
concatenación de los dos perfiles, que `go tool cover` funde bloque a bloque, como hace Codecov con los dos que publica
CI. Los recuentos de sentencias salen de los bloques del perfil, contando cada bloque una vez. Incluyen las funciones
anónimas de nivel de paquete, que `-func` no cuenta; por eso el global unitario da 96,85 % por recuento y 96,8 % por
`-func`.

| Árbol | `coverage.out` | `coverage-integration.out` | Unión | Umbral (`codecov.yml`) |
|---|---|---|---|---|
| Global | **96,8 %** (11 263 de 11 629 sentencias) | **97,4 %** (11 325 de 11 629) | 97,4 % | ≥ 70 % |
| `internal/core/**` (dominio) | **98,6 %** (2 545 de 2 581) | **98,6 %** (2 545 de 2 581) | 98,6 % | ≥ 85 % |
| `internal/cli/**` | 98,7 % (372 de 377) | 98,7 % | 98,7 % | ≥ 90 % |
| `internal/evals` | 97,2 % (3 004 de 3 089) | 98,2 % (3 035 de 3 089) | 98,2 % | — |

Los tres umbrales de proyecto se cumplen, así que no hace falta ningún test y T015 no toca ningún fichero de test. El
binario no cambia (§4) y `internal/core` da las mismas 2 545 de 2 581 sentencias que en el cierre de H7.2. El perfil de
integración sube `internal/evals` porque solo en él corre `TestPrepararElArbolDelSondeo`, que construye el árbol del
sondeo de verdad.

**Los ficheros nuevos del paquete de evals**, en la unión (y, entre paréntesis, en `coverage.out` si cambia):

| Fichero | Sentencias | Funciones por debajo del 100 % en `coverage.out` (`go tool cover -func`) |
|---|---|---|
| `internal/evals/umbrales.go` | 50 de 50 (100 %) | ninguna: las nueve funciones, al 100 % |
| `internal/evals/limites.go` | 20 de 20 (100 %) | ninguna: las tres, al 100 % |
| `internal/evals/definicion.go` | 43 de 43 (100 %) | ninguna: las siete, al 100 % |
| `internal/evals/sesiones.go` | 263 de 282 (93,3 %) | `detenido` 87,5 %, `desenlace` 90,0 %, `abrirSesion` 84,2 %, `crearDirectorioDeSesion` 90,0 %, `enlazarLasSkills` 70,0 %, `ejecutarElGuion` 94,4 %, `esperarConElTope` 95,8 %, `esperarConElMargen` 88,9 %, `matarTrasElMargen` 90,0 %, `senalAlGrupo` 71,4 %, `codigoDelProceso` 87,5 %, `crearFicheroDeSesion` 75,0 %, `cerrarFicheroDeSesion` 66,7 %; las otras ocho, al 100 % |
| `internal/evals/sondeo.go` | 322 de 333 (96,7 %) (291 de 333, 87,4 %) | `juzgarElSondeo` 96,8 %, `comprobarElSondeo` 88,9 %, `sondear` 77,3 %, y las cuatro que construyen el árbol (`prepararElArbol`, `ordenDeConstruirElBinario`, `ordenDeInstalarLasSkills`, `ejecutarLaOrdenDelArbol`) al 0,0 %, porque solo las ejecuta el test con la etiqueta `integration`; en la unión, `prepararElArbol` 87,5 % y `ejecutarLaOrdenDelArbol` 75,0 % |

**El diff.** En H7.3 nada pide una cifra del diff, pero Codecov la mide en la propuesta de cambio con el estado `patch`
(`target: auto`, la cobertura de la base, bloqueante). He hecho una estimación local con el criterio de los cierres de
H7.1 y H7.2: cuento cada línea añadida de los `.go` de producto de `git diff -U0 main` que cae en un bloque de la unión
de los dos perfiles. Da **93,98 %: 890 de 947 líneas**, sin líneas parciales. La base es la cobertura que el cierre de
H7.2 midió antes de su revisión final: 97,5 % en sentencias, y desde la fusión de H7.2 `main` no cambia ningún `.go`.
Queda por encima de la estimación, así que es probable que `codecov/patch` salga por debajo de su objetivo. Las 57
líneas que ningún perfil ejecuta están todas en `internal/evals` y todas son ramas de error de la regla genérica (un
error con el fichero o la orden delante, que el job y el sondeo convierten en código 1):

| Fichero | Líneas | Qué son |
|---|---|---|
| `internal/evals/sesiones.go` | 384-385, 459-460 | `filepath.Abs` que falla, lo que solo pasa si falla `os.Getwd`. |
| `internal/evals/sesiones.go` | 400-401, 424-425 | La preparación de la sesión que falla (la de `PrepararSesion`, que tiene sus propios tests) y `codigo-de-la-sesion` que no se puede escribir. |
| `internal/evals/sesiones.go` | 436-437, 464-465, 469-470, 552-553, 559-560, 717-718, 726-727 | El sistema de ficheros que falla al crear el directorio de la sesión, al listar las skills instaladas, al enlazar una skill, o al crear o cerrar `sesion.jsonl` y `sesion.err`, todo en un directorio que el propio repartidor acaba de crear en 0700. |
| `internal/evals/sesiones.go` | 604-605, 701-702 | El guion de la sesión que no arranca, y una espera que no devuelve un `*exec.ExitError`. El guion es `scripts/evals-sesion.sh` o, en los tests, un sustituto ejecutable. |
| `internal/evals/sesiones.go` | 641-642, 669-670, 683, 685 | Las señales al grupo: un `kill` que falla por algo distinto de que el grupo ya no exista, y ese mismo caso (`ESRCH`, que se devuelve como `os.ErrProcessDone`), que depende de la carrera entre la señal y la salida del grupo. |
| `internal/evals/sesiones.go` | 284, 346 | Dos ramas de sincronización: el cierre del reparto visto en `detenido` antes que en ningún otro sitio, y la interrupción sin ninguna sesión abierta. Los tests de cancelación las cubren por el otro camino, que depende del orden de las gorrutinas. |
| `internal/evals/sondeo.go` | 219-221, 519-520, 794-795, 803-804 | El directorio de sesiones que no se puede listar o crear, la definición del job que no se puede leer y el `filepath.Abs` del temporal. |
| `internal/evals/sondeo.go` | 695-696, 700-701, 707-708, 733-734, 798-799, 823-824, 837-838 | La construcción del árbol que falla (`go install`, el `HOME` del sondeo, `skills install`) y los errores del repartidor o del juicio que `sondear` devuelve tal cual. |

Queda como supuesto `[alcance]` de T015 en `gates/supuestos.md`. La tarea solo pide tests si la cobertura global o la
del dominio quedan por debajo de su umbral, y no quedan. Si `codecov/patch` sale rojo, lo decide quien lea el informe,
como en H7.1 y H7.2, y ningún umbral se toca.

## 3. Quickstart (§1 a §5, §8 y §9)

Ejecutados sobre `b9438b2`, tal como los escribe `quickstart.md`, desde la raíz del repositorio. Las salidas de
`go test` se leyeron con `rtk proxy`, sin el resumen del proxy de la sesión, y los `--- PASS` y `--- FAIL` se
contaron con `grep -c`, para que ningún escenario pase en vacío. Los bloques de §4, §5, §9.1, §9.2 y §9.3 no se
reescribieron: los extrajo del propio `quickstart.md` un programa fuera del repositorio, que los pasó a `bash` y comprobó
al terminar que no quedaba ningún `kitlegal-h73-*` en el `TMPDIR` (0 antes y 0 después en todos). Se hizo así porque
la sesión no deja escribir en el `TMPDIR` con órdenes de shell. Del §5
no se ejecutó la última orden, y tampoco el §6, porque lanzan el sondeo y ninguna tarea lo ejecuta (FR-068). El §7 lee el
informe del job de cierre, que llega después del run. Todos los escenarios dan lo esperado.

| Escenario | Resultado |
|---|---|
| §1. La lista, la prosa de la skill y el calibrado | Primera orden: **16 `--- PASS`, contados, ningún `--- FAIL`**, que son `TestEvalsDelRepositorio` y sus 15 subpruebas, entre ellas las que nombra el quickstart (`expresiones-calibradas`, `expresiones-en-los-bloques`, `expresiones-de-la-skill` y `prosa-de-la-skill`), y además `normas-conocidas`, `avisos-del-esquema`, `conjunto`, `cobertura-del-esquema`, `avisos-de-la-skill`, `hallazgos-de-la-skill`, `hallazgos-del-esquema`, `conjunto-legal-core`, `formato`, `grafo-previo` y `grabado`. Segunda orden: **53 `--- PASS`, ningún `--- FAIL`**: `TestProsaDeLaSkill` con 17 subpruebas (18), `TestExtraerExpresionesProhibidas` con 20 (21, entre ellas `maquinaria-otra-conversacion-y-anuncio` y `nada-que-trasladar`) y `TestJuzgarLasExpresionesProhibidas` con 13 (14, entre ellas `positiva/anuncio`). `make skills-check` da `ok` en `internal/app`, `internal/skills` e `internal/evals`. `wc -l`: **270** (< 300). |
| §2. El informe: umbrales, sin medir, reintentos y duración | **52 `--- PASS`, contados, ningún `--- FAIL`**: `TestUmbralesDelInforme` con 9 subpruebas (10), con los casos del quickstart: `tres-de-51-de-sonnet-5` → `fallo`, `dos-de-51-de-sonnet-5` sin cambio, `dos-de-30-de-haiku-4-5` (incumplido) sin cambio, `sin-lista-ni-objetivo` → `[]`, `901-s-con-objetivo-900` → `fallo` y `900-s-con-objetivo-900` sin cambio. Además, `TestInformeMarkdownDeLosUmbrales` con 3 (4), `TestInformeConSesionesSinMedir` con 5 (6: `mensaje-del-limite-de-uso`, `reintentos-agotados`, `cortada-durante-reintentos`, que son (a), (b) y (c), más `sin-abrir-tras-el-limite-de-uso` y `reintentos-de-los-que-se-recupera`, el 429 recuperado), `TestLeerSesionConReintentos` con 12 (13) y `TestClasificarElLimite` con 18 (19). |
| §3. El repartidor con los sustitutos | **14 `--- PASS`, contados, ningún `--- FAIL`**: `TestEjecutarSesionesEnParalelo`, `TestEjecutarSesionesTrasElLimiteDeUso` con `mensaje-del-limite-de-uso` y `reintentos-agotados`, `TestEjecutarSesionesConElContextoCancelado` y `TestTopeDeLaSesion` con `termina-antes`, `basta-term` (124), `hace-falta-kill` (137) y `muere-por-otra-senal`. El patrón del quickstart no lleva `$`, así que ejecuta además `TestEjecutarSesionesConUnError` y `TestEjecutarSesionesSinConcurrencia` (`0` y `-1`), que también pasan. En macOS, sin strace, sin sudo y sin `timeout`. |
| §4. La definición del job y lo que la hace fallar | Primera orden: `ok`. En la copia de `git archive HEAD` con `cancel-in-progress: true`: `--- FAIL: TestDefinicionDelJob/del-repositorio`, con `jobs.evals.concurrency.cancel-in-progress: vale true, y lo esperado es false`, luego `FAIL` y **`código: 1`**. `rm -rf` borra la copia, y ni el árbol ni el índice cambian: `git status --porcelain` no nombra nada de la copia, solo los dos ficheros de `gates/` que el workflow modificó antes de la tarea y los dos de esta tarea (`quickstart.md` y `cierre.md`). |
| §5. El sondeo sin modelo (sin la última orden) | Primera orden: **77 `--- PASS`, contados, ningún `--- FAIL`**: `TestJuicioDelSondeo` con 24 subpruebas (25), `TestSalidaDelSondeo` con 2 (3), `TestComprobarElSondeo` con 37 (38), `TestGuionDelSondeo` con 4 (5: `sin-cinco-argumentos` y `go-test-sale-con-0`, `-1` y `-2`) y `TestSondear` con 5 (6, entre ellas `sin-la-credencial` y `limite-de-uso-en-la-primera`). Segunda orden: `ok`. Con `-v`, `--- PASS: TestPrepararElArbolDelSondeo`, así que la etiqueta `integration` sí lo ejecuta. Tercera: `evals-sondeo       sondeo local de unas evals de una skill con Claude Code, sin strace ni veredicto (macOS o Linux; consume la suscripción; CLAUDE_CODE_OAUTH_TOKEN)`. |
| §6. El sondeo con modelo | **No se ejecuta.** Abre 30 sesiones de Sonnet 5 con la suscripción de quien lo lanza, así que queda fuera de `make ci` y del run y lo ejecuta la persona al leer el informe final (FR-068, FR-070; SC-002). Sus órdenes nombran lo que existe (abajo). |
| §7. El job de cierre | **No se ejecuta**: lee `gates/evals/*.json`, que deja el cierre del workflow después del run (FR-098). |
| §8. `make ci` | En primer plano: `ci: todos los controles en verde`, con `skills-check` y `schema-check` sin drift. Es el del que salen los perfiles del §2 y también la verificación de esta tarea. |
| §9.1. Los puntos de entrada de la etiqueta `evals` | `go vet -tags evals` no imprime nada y sale con 0. La lista con la etiqueta nombra `TestEjecucionDelJob`, `TestSondeo` y `TestComprobarConsultaRepetida`, y ninguno de `TestPlanDeSesiones`, `TestPrepararSesion` ni `TestInformeDelJob`: `grep -x` da exactamente esos tres. La lista sin etiqueta nombra todos los tests de «Tests nuevos» de plan.md salvo `TestPrepararElArbolDelSondeo`, que es de la etiqueta `integration` (§5). `git grep` no imprime nada y da **`código: 1`**. |
| §9.2 y §9.3 | En §4 y §2. |

### Las órdenes de §5 y §6 nombran lo que existe (FR-070)

- **El objetivo `evals-sondeo`** está en `make help` (§5, tercera orden). El `Makefile` le pasa al guion
  `"$(SKILL)" "$(EVALS)" "$(MODELO)" "$(REPETICIONES)" "$(CONCURRENCIA)"`.
- **Sus cinco argumentos en el guion del sondeo**: `scripts/evals-sondeo.sh` sale con uso si no recibe exactamente
  cinco (`[[ $# -ne 5 ]]`, línea 18) y los pasa a `TestSondeo` como `-skill "$1" -evals "$2" -modelo "$3"
  -repeticiones "$4" -concurrencia "$5"` (línea 37). Las cinco banderas, más `-temporal`, están registradas en
  `internal/evals/job_test.go` (`//go:build evals`), y la subprueba `sin-cinco-argumentos` de `TestGuionDelSondeo` lo
  fija.
- **Las evals 03, 06, 13, 14 y 15 de `boe-legislacion`** existen: `03-lrbrl-atribuciones-del-pleno.yaml`,
  `06-irpf-rendimientos-del-trabajo.yaml`, `13-lrbrl-atribuciones-por-materia.yaml`,
  `14-trlrhl-impuestos-por-materia.yaml` y `15-irpf-rendimientos-por-materia.yaml`, todas en `evals/boe-legislacion/`.
- **`TestSondeo`** está en la lista con la etiqueta `evals` del §9.1.
- **Lo demás que usa el §6**: `git show main:skills/boe-legislacion/SKILL.md` lee la `SKILL.md` de `main` (266
  líneas, la v0.1.2), `git archive HEAD` lleva el `Makefile` para `make -C "$copia"`, y `claude-sonnet-5` es el
  `MODELO_DE_EVALS` de `.github/workflows/evals.yml`, que el sondeo acepta como id (`TestComprobarElSondeo`).

## 4. Lo que el hito no cambia (quickstart.md §9.2) y los umbrales

Salida del §9.2: `H7.2: 0`; la segunda orden, sin salida; la tercera, `0`; la cuarta, sin salida; la quinta, solo
`M	evals/boe-legislacion/expresiones-prohibidas.yaml` y `M	schemas/expresiones-prohibidas.yaml.json`; y `270`.

- **ADR nuevo: no hay** (FR-080). La cuarta orden del §9.2 buscaba los ADR en `docs/adr`, pero el directorio es
  `docs/ADR/`, y una ruta de git distingue mayúsculas aunque `core.ignorecase` esté activado en macOS
  (`git ls-files docs/adr` nombra 0 ficheros y `git ls-files docs/ADR`, 29). Tal como estaba, la orden no podía
  encontrar ningún ADR. T015 corrige esa ruta en `quickstart.md` (supuesto `[interno]` de T015), y con `docs/ADR` la
  orden sigue sin imprimir nada. `docs/ADR/` tiene los mismos 29 ficheros en `HEAD` y en `main`, y `git diff
  --name-status main -- docs` tampoco imprime nada. Ninguna decisión de arquitectura cambia: el hito aplica el ADR 0016
  (Sonnet 5 decide, tres repeticiones, regla por serie, Haiku 4.5 informativo) y el contrato de `umbrales` del ADR 0029,
  sin reabrir ninguno.
- **Los artefactos de H7.2 no cambian** (FR-080): `git diff --quiet main -- specs/012-h7-2-la-consulta-repetida` sale
  con 0 (`H7.2: 0`), sobre 52 ficheros versionados.
- **El binario no cambia ni enlaza el paquete de evals.** `git diff --name-only main -- cmd internal ':!internal/evals'`
  no imprime nada (sobre 460 ficheros versionados), y `go list -deps ./cmd/kitlegal | grep -c '/internal/evals$'` da
  `0`: ninguno de los 17 paquetes del módulo que enlaza el binario (entre ellos el propio `cmd/kitlegal`) es
  `internal/evals`.
- **No hay fuente ni grabación nueva para la tabla de fuentes.** `docs/SOURCES.md` no cambia y no hay ningún manifiesto
  `grabaciones.json` nuevo, ningún test `//go:build grabacion`, nada para el paso `grabar_datos` y nada en
  `evidencias/` (§1; plan.md, «Datos externos»).
- **`skills/boe-legislacion/SKILL.md` tiene 270 líneas**, menos de 300 (FR-017; SC-010). La comprueba también
  `TestSkillsDelRepositorio` en `skills-check`.
- **Ningún umbral de FR-002 ni de FR-051 se rebaja, se retira ni queda sin decidir** (FR-008):
  - El 5 % de Sonnet 5 es `umbralDeExpresionesProhibidas = 0.05` (`internal/evals/umbrales.go:13`). Entró con T005 y
    ningún commit posterior lo cambia: `git log -G 'umbralDeExpresionesProhibidas = ' main..HEAD` solo nombra T005.
    Decide en el modelo que decide (`Decide: delModelo.Modelo == e.ModeloQueDecide`), y ese modelo sigue siendo
    `MODELO_DE_EVALS: claude-sonnet-5`, que ningún commit de la rama toca. El total son todas las respuestas medidas
    del modelo en las evals que activan la skill, sin sacar las informativas: `recontarExpresiones` solo deja fuera
    las sesiones ilegibles, las sin medir y las evals que no activan la skill. Por eso el caso del quickstart es
    «3 de 51». La lista no se recorta: gana 14 expresiones y no pierde ninguna (§1).
  - Los 900 s son `objetivo_de_duracion: 900` en la entrada `boe-legislacion` del `include` de
    `.github/workflows/evals.yml`, que llega al informe como `OBJETIVO_DE_DURACION_DE_EVALS`. Entraron con T008 y
    ningún commit posterior los cambia. Su umbral lleva `Decide: true` (`umbrales.go:104`), y `TestDefinicionDelJob`
    falla si el objetivo de `boe-legislacion` no es 900.
  - `TestUmbralesDelInforme` ve fallar los dos: `tres-de-51-de-sonnet-5` y `901-s-con-objetivo-900` dan `fallo` (§3).
  - La regla por serie (FR-007) no cambia: `REPETICIONES_DE_EVALS: 3` y `UMBRAL_DE_EVALS: 2` siguen en
    `evals.yml`.
- **plan.md tiene las dos filas del cierre de FR-099** en «Controles de umbral»: `FR-002, FR-003, SC-001 (Sonnet 5)`,
  con ≤ 5 % (≤ 2 de 51), en `evals:boe-legislacion:expresiones_prohibidas:claude-sonnet-5`, y `FR-051, SC-001
  (duración)`, con ≤ 900 s, en `evals:boe-legislacion:duracion_de_las_sesiones`. El umbral de Haiku 4.5 (FR-004) no
  tiene fila, y el texto de la sección lo dice. El resto de filas son los umbrales que se miden en `make ci`, cada uno
  con su test (`ci:…`).
- **SC-010**: `make ci` en verde, con `schema-check` y `skills-check` sin drift y con las reglas del conjunto; la skill
  con 270 líneas; y las entradas del hito en `CHANGELOG.md`, bajo *Unreleased* (T014): `boe-legislacion` v0.1.3,
  `umbrales` con el de Sonnet 5 que decide, la familia `anuncio`, el job en paralelo, las sesiones sin medir, los
  reintentos, la duración y el sondeo `make evals-sondeo`.
