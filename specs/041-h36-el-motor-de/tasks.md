---

description: "Tareas de H36 · El motor de cómputo descrito por datos (ADR 0061)"
---

# Tasks: H36 · El motor de cómputo descrito por datos (ADR 0061)

**Input**: `specs/041-h36-el-motor-de/` — [spec.md](spec.md), [plan.md](plan.md), [research.md](research.md),
[data-model.md](data-model.md), [quickstart.md](quickstart.md) y [contracts/](contracts/) (`datos-de-plazos.md`,
`api-del-dominio.md`, `applet-calcular.md`, `aceptacion.md`, `grabacion.md`).

**Tests**: obligatorios (constitución III). La primera tarea escribe la suite de aceptación, que el workflow congela;
cada tarea posterior trae su test y el cambio que lo pone en verde, y deja `make ci` en verde por sí sola.

**Organización**: las tareas van en el orden en que se ejecutan una tras otra, de dentro afuera, con las piezas
indivisibles de plan.md «Diseño» («Orden y piezas indivisibles»). Cada fase lleva la historia a la que sirve sobre todo.
La línea de cada tarea declara todas las rutas que crea o modifica: el guardián rechaza cualquier otra; `go.mod`,
`go.sum`, `CHANGELOG.md` y el directorio del feature siempre se admiten, y declarar `x.go` admite `x_test.go`.

## Format: `[ID] [P?] [Story] Descripción`

- **[P]**: ninguna tarea lo lleva. Cada una parte de la anterior: T004 deja la forma nueva que leen T005 a T011; T007
  deja el motor que miden T010 y T011; T008 cambia firmas que usan los tests de T004 y T007; T009 parte de lo que graba
  el paso `grabar_datos` tras T002; y varias tocan los mismos tests de `internal/core/plazos` y `internal/app/plazos_test.go`.
- **[Story]**: la historia de [spec.md](spec.md) a la que sirve: US1 (con los datos de hoy, lo mismo, byte a byte), US2
  (el civil de la LEC), US3 (el cómputo de ensayo sin tocar Go), US4 (el esquema sin datos), US5 (`LeerTipos`), US6 (los
  preceptos del civil contra lo grabado). T002 (lo que se graba) es fundacional y no lleva historia; T012 y T013 son de
  pulido.
- **[aceptacion]**: la suite congelada, solo T001.
- **[datos]**: T002, T003 y T006, cada una porque crea o modifica ficheros bajo `testdata/` o `schemas/`. T002 solo
  toca el manifiesto y el test de grabación; T006 solo los dos ficheros de prueba. T003 junta las etiquetas del esquema
  con sus goldens y los tests que comparan unos con otros porque, por separado, `make ci` queda en rojo (plan.md, «Orden y
  piezas indivisibles»: «El esquema y sus goldens […] van en una sola tarea `[datos]`»); su línea dice por qué.

## Convenciones de todas las tareas

- `make ci` en verde al terminar cada una, en primer plano, y si hace falta por objetivos (`make fmt-check`, `make lint`,
  `make test`, `make test-integration`, `make vuln`, `make schema-check`, `make skills-check`). Ninguna usa la red más que
  la de las herramientas de Go; ninguna graba ni pide a una fuente: lo único que se graba, los bloques `a130` y `a133` de
  la LEC, lo graba el paso `grabar_datos` del workflow, sin modelo, tras el commit de T002 (plan.md «Datos externos»).
  Ninguna escribe fuera del repositorio y de `$TMPDIR`; ninguna abre una sesión con modelo ni ejecuta `make evals`,
  `make evals-sondeo`, `make evals-laboratorio`, `TestSondeo*` ni los guiones `scripts/evals*.sh` (ADR 0032).
- `data/plazos.yaml` es el único fichero de `data/` que cambia, en T004 (la forma de §C.1) y en T009 (el cómputo
  `civil`), con el contenido literal de contracts/datos-de-plazos.md §1: no copia texto de ninguna norma, nombra cada
  precepto por su identificador, su bloque y su apartado, y lo comprueba contra lo grabado `TestPreceptosDeLosPlazos`
  (research D14, precedente de H9). Ningún otro fichero de `data/` se toca (`data/festivos/`, `data/cobertura.yaml`,
  `data/normas.yaml` no cambian; spec, «Fuera de alcance»), y ninguno sale de lo grabado.
- No se tocan `.golangci.yml`, `Makefile`, `.github/`, `scripts/`, `cmd/`, `skills/`, `evals/`, `docs/`, ningún fichero
  de `specs/023-h9-internal-core-plazos/` (FR-018) ni `specs/031-h39-c-omo-se/gates/aceptacion-congelada.json` (research
  D12). Ningún `//nolint`, `t.Skip`, `TODO` ni test desactivado. Ningún ADR nuevo ni cambio de los ADR 0061 y 0042, y
  `docs/SOURCES.md` no cambia: la única fuente, el BOE consolidado, ya tiene su fila revisada en `main` (spec, «Fuera de
  alcance»; contracts/grabacion.md §4).
- Los tests de H9 que cambian son solo los de FR-017, y en ninguno cambian las entradas ni lo esperado de un caso de
  cálculo; fuera de esa lista cambian solo las tres desviaciones de research D10 y plan.md «Desviaciones declaradas»
  —la raíz del cliente de `TestGrabarEvals` (T002), la lista esperada de `TestGrabacionesSinSolape` (T009) y la
  exigencia de una búsqueda grabada a la unión y no a cada conjunto en `rangosDeLasBusquedasGrabadas` de
  `TestEsquemaDeNormas/rangos-grabados` (T009)— y la entrada de `probarLosDatosQueNoSeLeen` de H32 (T004;
  `gates/supuestos.md`). Las suites congeladas de los hitos anteriores y sus
  copias activadas siguen en verde sin cambiar un fichero suyo, salvo el guion de H39 y su copia en T003 (FR-017).
- Los textos son los de los contratos y del spec: los datos, las reglas y los ficheros de prueba, los de
  [contracts/datos-de-plazos.md](contracts/datos-de-plazos.md); la API, el cómputo, los avisos, los errores y lo que toma
  `internal/evals`, los de [contracts/api-del-dominio.md](contracts/api-del-dominio.md) y spec §C.3 a §C.6; la ayuda, el
  esquema, `TestDescribeDePlazosSinDatos` y `TestNingunCodigoDePlazosEnGo`, los de
  [contracts/applet-calcular.md](contracts/applet-calcular.md) y spec §C.7 y §C.8; la grabación, los de
  [contracts/grabacion.md](contracts/grabacion.md); los guiones, los de [contracts/aceptacion.md](contracts/aceptacion.md);
  los goldens del civil y del ensayo, los de spec §C.9 y §C.10. Nada se inventa fuera de ellos.
- Arquitectura (plan.md «Constitution Check»): ningún paquete, dependencia, orden, bandera ni código de salida nuevo;
  `internal/core/plazos` y sus tests solo usan la biblioteca estándar sin `os` ni `io`, `go.yaml.in/yaml/v3` e
  `internal/core/schema` (R1): sus ficheros de prueba entran con `//go:embed`, y lo que lee ficheros (`os`, `go/parser`)
  vive en tests de `internal/app`. El error nuevo de FR-007 es un `errorDeArgumentos` (clase `argumentos`, código 2); los
  defectos de `LeerTipos`, sin clase de usuario (`inesperado`, código 1). Ningún `panic`; nada nuevo escribe en stdout ni
  llama a `os.Exit`. Tests de tabla con `t.Parallel()`, `require` para las premisas y `assert` para lo comparado
  (research D15). Cobertura de `internal/core/**` ≥ 85 % (DoD 9).
- Cada control se ve fallar con un mutante —la medida pasada de su umbral— antes de darlo por bueno (ADR 0029): cada
  tarea que construye un control lo ve en rojo con el mutante de plan.md «Mutantes» sobre una copia bajo `$TMPDIR`, nunca
  sobre el árbol versionado, y T013 los registra todos en `mutantes.md`.

---

## Phase 1: Setup — la suite de aceptación (US1 y US4, escrita primero)

**Goal**: fijar, antes de cualquier cambio de producto y con el binario de `main`, lo que `kitlegal plazos calcular`
tiene que seguir dando con los datos de hoy y lo que su esquema deja de nombrar (ADR 0018; plan.md «Aceptación e2e»).

**Independent Test**: el workflow ve cada guion en rojo por su primera aserción (rojo primero) y los congela; al terminar
el bucle los activa como `h36-plazos-de-hoy.txtar` y `h36-describe.txtar` y los ejecuta `TestEntregaDelHito` en `make ci`.

- [X] T001 [aceptacion] [US1] Escribir, a partir del spec y de contracts/aceptacion.md, sin código de producto, los dos guiones testscript specs/041-h36-el-motor-de/aceptacion/plazos-de-hoy.txtar y specs/041-h36-el-motor-de/aceptacion/describe.txtar. **Precondición** (contracts/aceptacion.md §1; quickstart §0): `git diff --name-only cd3ad148 HEAD | grep -v '^docs/' | grep -v '^specs/'` no imprime nada; si imprime algo, no se captura. **Captura** con el binario de `main` (FR-021): `make build`, que lo deja en `./bin/kitlegal` (ignorado por git), y, tras escribir el guion de §2 sin sus ficheros, una sola orden por golden, `./bin/kitlegal plazos calcular <orden> --json | sed -E -e 's/"fecha_consulta":"[^"]*"/"fecha_consulta":""/' -e '1i -- <golden> --' >> specs/041-h36-el-motor-de/aceptacion/plazos-de-hoy.txtar`, que lleva el sobre del binario al guion sin pasar por el contexto de la sesión. **plazos-de-hoy.txtar** (§2; cabecera `# Cubre: US1 · FR-005, FR-006, FR-008, FR-011, FR-012, FR-021 · SC-001`): primero la aserción que falla hoy, `exec kitlegal plazos calcular x x x --describe` con `! stdout 'alzada'` y `! stdout 'reposicion'`; después los tres casos (`caso-11.json`, `caso-15.json`, `caso-16.json`) y los nueve errores de la tabla de §2.3 (`error-tipo.json` … `error-fuera-del-calendario.json`), cada orden con `exec sh -c`, la salida estándar por la misma sustitución de `fecha_consulta` (sin el `1i`), `test $? -eq <código>` sobre el código de `kitlegal` y no el de la tubería (0 en los casos; 2, o 3 en `error-inexistente.json`, en los errores) y `cmp stdout <golden>`; la fila «no cargan» y la nueva de FR-007 no se capturan (§2.3). **describe.txtar** (§3; cabecera `# Cubre: US4.1 (datos de hoy), US4.2, US4.3 · FR-012 · SC-005`): `exec kitlegal plazos calcular x x x --describe` con código 0 y, por argumento, la ayuda de spec §C.7 byte a byte con la forma de las líneas 39 a 43 de `h39-describe.txtar` (falla hoy en la primera); `stdout -count=2` del `"pattern"` de §C.8; `! stdout '"dias-del-art-183-lopj"'` y `! stdout '"alzada"'`; los enumerados `"nacional-y-autonomico"`, `"sin-festivos-locales"` y `"autonomico"`; la descripción del verbo, `stdout 'Devuelve la fecha de un plazo —un recurso, el acceso a la información pública, el silencio o uno genérico—'`; y `exec kitlegal plazos calcular --help` con código 0, sin mirar su texto (research S5). Los guiones afirman solo lo que el spec fija byte a byte; US2, US3, US5 y US6 no se ven desde el binario y las cubren los tests Go de T004, T007 a T011 (§4). Sin evals: el hito no cambia ninguna skill, protocolo, referencia ni eval (FR-020). La tarea no deja ningún otro fichero cambiado (`bin/` está ignorado).

**Checkpoint**: la suite está escrita y congelada con los sobres de `main`; nada del producto ha cambiado.

---

## Phase 2: Foundational — lo que se graba

**Purpose**: los bloques `a130` y `a133` de la LEC, que lee `TestPreceptosDeLosPlazos` en T009 (FR-015, FR-016),
grabados por el paso `grabar_datos` en el manifiesto nuevo de H36 antes de que los use ninguna tarea. La entrada sin
grabar no pone nada en rojo: la raíz de H36 no entra en la unión hasta T009 (research D9, V8).

- [X] T002 [datos] La entrada de la Ley 1/2000 en el manifiesto de grabación de `internal/evals` y la raíz de grabación de `TestGrabarEvals`, sin otro código ni test (FR-016; contracts/grabacion.md §1 y §2; research D9, D10, V7, V16; plan.md «Datos externos»). **Manifiesto**: internal/evals/testdata/grabaciones.json (`"fuente": "boe.legislacion-consolidada"`, con fila revisada el 2026-09-13 en `docs/SOURCES.md` de `main`), al final de su lista `normas`, la entrada de contracts/grabacion.md §1 byte a byte: `{"busqueda": "enjuiciamiento civil", "titulo_empieza_por": "Ley 1/2000,", "bloques": ["a130", "a133"], "para": "…"}` con el `para` de §1; nada más del fichero cambia. **Test de grabación**: internal/evals/grabacion_test.go (`//go:build grabacion`, `TestGrabarEvals`), desviación declarada de research D10: el cliente que graba se construye con `httpx.ConRaizDeGrabacion` sobre el valor de la variable `KITLEGAL_EVIDENCIAS` cuando no está vacía, y con la raíz de su conjunto si lo está, como hoy (línea 94); nada más del test cambia: lee los dos manifiestos, siembra cada consulta desde `UnionDeGrabaciones` con `Preparar` y solo pide a la fuente, por el cliente HTTP único del proyecto y con GET, lo que falta, que son `boe articulo BOE-A-2000-323 a130` y `a133` (V16). Tras el commit, el paso `grabar_datos` del workflow lo graba en el manifiesto nuevo de H36 que nombra contracts/grabacion.md §2; la tarea no graba nada. **Verificación**, sin cambiar otro fichero: `go vet -tags grabacion ./internal/evals/` sin defectos; `make ci` en verde con la entrada sin grabar (`TestManifiestoDeGrabaciones`, `TestGrabacionesSinSolape` y `TestIdentificadoresDeLasNormas` siguen en verde: la LEC la sigue resolviendo la entrada de H5); `grep -c '"titulo_empieza_por": "Ley 1/2000,", "bloques": \["a130", "a133"\]' internal/evals/testdata/grabaciones.json` da `1`.

**Checkpoint**: tras el paso `grabar_datos`, los bloques `a130` y `a133` de la LEC están grabados en el conjunto de H36 y
las tareas siguientes parten de ellos.

---

## Phase 3: User Story 4 — el esquema de `calcular` no depende de los datos (Priority: P2)

**Goal**: la ayuda de los cinco argumentos es la de §C.7 y `tipo` y `avisos[].codigo` llevan el `pattern` de §C.8 en
lugar de un `enum` (FR-012). Va antes del motor porque las etiquetas no dependen de la forma de los datos, y deja de usar
`CodigosDeTipo` y `CodigosDeAviso` en los tests del esquema antes de que T004 a T008 los cambien.

**Independent Test**: `TestEsquemasPublicados`, el subtest `los-enumerados-publicados` de `TestAppletPlazos`, el golden
`describe.json` y `h39-describe`; tras la activación, `h36-describe`.

- [X] T003 [datos] [US4] Las etiquetas del esquema de `calcular` y sus goldens, en una sola tarea porque con las etiquetas nuevas y los goldens de hoy `TestAppletPlazos`, `TestEsquemasPublicados` y `TestEntregaDelHito/h39-describe` salen en rojo, y con los goldens nuevos y las etiquetas de hoy, también (FR-012, FR-017; SC-005; spec §C.7 y §C.8; contracts/applet-calcular.md §1 a §3; research D5, D12; plan.md «Diseño», paso 4). **Etiquetas**: internal/app/plazos.go, las `help` de `Tipo`, `Inicio`, `Municipios`, `Duracion` y `Unidad` de `argumentosDeCalcular`, byte a byte las cinco de §C.7; la descripción del verbo no cambia. internal/core/plazos/salida.go: `Plazo.Tipo` y `Aviso.Codigo` con `jsonschema:"pattern=^[a-z0-9]+(-[a-z0-9]+)*$"` y ningún `enum`; el comentario de cabecera deja de nombrar `CodigosDeTipo` y `CodigosDeAviso` entre los vocabularios que repiten los enumerados, y el de `Aviso.Codigo` deja de decir que es uno de `CodigosDeAviso`; los enumerados de `festivos[].ambito`, `cobertura[].ambito` y `cobertura[].falta` siguen. **Tests**: internal/app/plazos_test.go, `compruebaDeclaracionDePlazos` deja de exigir que la ayuda del tipo nombre cada código de `CodigosDeTipo` y sigue exigiendo que la de la unidad nombre las tres unidades; `compruebaEnumeradosDePlazos` exige el `pattern` de §C.8 y ningún `enum` en `tipo` de `plazos.Plazo` y `codigo` de `plazos.Aviso` y los otros tres enumerados como hoy. internal/core/plazos/salida_test.go: su test de los enumerados deja `tipo` y `codigo`, que ya no tienen, y sigue con los tres de hoy. **Goldens**, regenerados con las banderas del repositorio: internal/app/testdata/plazos/describe.json con `go test -count=1 -run '^TestAppletPlazos$' ./internal/app/ -actualizar-golden-de-plazos`, y schemas/plazo.json (solo la parte de `plazos calcular`) con `go test -count=1 -run '^TestEsquemasPublicados$' ./internal/app/ -actualizar-esquemas`; a mano e idénticas, las líneas 39 a 43 de internal/app/testdata/script/h39-describe.txtar y de su copia specs/031-h39-c-omo-se/aceptacion/describe.txtar, con las ayudas de §C.7. **Verificación**: `git diff --exit-code` sobre `caso-a.json`, `caso-b.json` y `caso-e.json`, que la bandera reescribe con los mismos bytes (FR-021; si cambiaran, se arregla el código, no el golden); `grep -c '"pattern"'` de `describe.json` cuenta los dos nuevos más los que ya tuviera; `make ci` en verde; en una copia bajo `$TMPDIR`, el enumerado de los tipos de hoy devuelto a la etiqueta de `Plazo.Tipo` pone en rojo `TestEsquemasPublicados` y `los-enumerados-publicados` (mutante de SC-005).

**Checkpoint**: `--describe` de `calcular` ya no nombra tipos ni avisos; la salida de cada plazo no ha cambiado.

---

## Phase 4: User Story 5 — `LeerTipos` lee y comprueba la forma nueva (Priority: P2)

**Goal**: `data/plazos.yaml` en la forma de §C.1 —`da` en cada tipo y los tres ejes en cada cómputo— y `LeerTipos`
con las diez reglas de §C.2, que copia en cada tipo su cómputo (FR-001 a FR-003, FR-013). El motor de hoy sigue
contando igual con lo que lee.

**Independent Test**: `TestLeerTipos` (un subtest por regla, el de dos defectos y el de la copia del contrato), y todos
los de H9 de `internal/core/plazos` e `internal/app` en verde sin cambiar un caso.

- [X] T004 [US5] Los datos en la forma nueva y su lector, en una sola tarea porque el `LeerTipos` de hoy rechaza las claves nuevas (`KnownFields`) y el nuevo las exige (FR-001, FR-002, FR-003, FR-011, FR-013, FR-017; SC-005; spec §C.1 y §C.2; contracts/datos-de-plazos.md §1 y §2; contracts/api-del-dominio.md §1; research D1, D3; plan.md «Diseño», pasos 1 y 2). **Datos**: data/plazos.yaml con el contenido de contracts/datos-de-plazos.md §1 **sin el cómputo `civil`**, que entra en T009 (§1, «El civil entra cuando hay con qué comprobarlo»): la cabecera de §1, `da` en cada uno de los siete tipos (detrás de `duracion` si la lleva y delante de `preceptos` o `unidades`; `silencio` con `dia-siguiente`, los demás con `vencimiento`), y en `lpac` y `contencioso` sus `municipios`, sus `unidades` (`contencioso` solo `meses`) y sus `periodos` (el de agosto `no-corre` y el del art. 183 `aviso` de `contencioso`, con el texto de `textoDelArt183` byte a byte), con los preceptos, sus bloques, sus `fija` y sus `vigente_desde` de hoy intactos y el precepto de un periodo sin `bloque`. **Lector**: internal/core/plazos/tipos.go: `Tipo` gana `Da string` (`yaml:"da"`) y un campo sin exportar con su cómputo descrito (municipios con su texto, unidades, periodos con su efecto, su precepto y su aviso, preceptos) y la posición de este en el fichero; desaparece `PreceptosDeUnComputo` (research D1); `LeerTipos`, con la misma firma, decodifica con `KnownFields(true)`, acumula con `errors.Join`, en el orden del fichero, los defectos de las diez reglas de §C.2 con la forma de la tabla de contracts/datos-de-plazos.md §2 y los de hoy que no dependen de un código, y deja de dar los que dependen de los siete códigos y de los dos cómputos («no es uno de los siete», «falta el tipo», «no es lpac ni contencioso», «falta el cómputo», «fija su duración y no la lleva», «no fija su duración y lleva una», «lleva unidades, que solo lleva generico»); admite un cómputo sin tipos. Desaparece `fijaSuDuracion` (tipos.go:36-41), que decide por los códigos `silencio` y `generico` y cuyos únicos usos son las dos reglas de la duración que se quitan (tipos.go:240 y :246): sin ellas queda sin uso y `unused` (`.golangci.yml`, `default: standard`) pone `make lint` en rojo; lo que decidía lo dicen ahora los datos, con las reglas 2 y 3 de §C.2 y los defectos de la duración que no dependen de un código (contracts/datos-de-plazos.md §2). Las constantes `Tipo…`, `Computo…`, `CodigosDeTipo`, `Computos` y `AvisoDiasDelArt183` siguen hasta T008, sin que el lector las use. internal/core/plazos/calcular.go, solo para tomar del cómputo descrito los preceptos del cómputo que hoy llegan por `PreceptosDeUnComputo` (ADR 0042 sobre los preceptos del tipo, de su unidad y de su cómputo, como hoy); el cálculo no cambia. internal/core/plazos/doc.go, solo si su comentario nombra las reglas que desaparecen. **Tests**: internal/core/plazos/calcular_test.go, `plazosDelContrato` pasa a ser data/plazos.yaml de contracts/datos-de-plazos.md §1 **con** el civil y sin la cabecera (FR-017); internal/core/plazos/tipos_test.go, `TestLeerTipos` con un subtest por regla sobre la copia del contrato rota como dice la tercera columna de la tabla de §2 (que exige que el defecto nombre el tipo o el cómputo y lo que falla, no sus bytes), el de la copia del contrato (los siete tipos sin defecto, cada uno con su cómputo copiado, y el civil sin tipos; US5.1), el de una copia con los defectos de las reglas 1 y 10 a la vez, cuyo error da los dos (US5.3), y sin los subtests de las reglas que se quitan. internal/skills/regiones_test.go, solo la entrada de `probarLosDatosQueNoSeLeen` (H32, no es de H9): con las reglas de §C.2 `tipos: []` ya no tiene defecto (los suyos eran «falta el tipo» y «falta el cómputo»), así que pasa a `tipos: [{codigo: inventado}]`, que da los de las reglas 1 y 10 y los de nombre y preceptos; sus aserciones no cambian (`gates/supuestos.md`). **Verificación**: `make ci` en verde con los tests de H9 sin cambiar un caso (`TestCalcular`, `TestComprobarPeticion`, `TestPropiedadesDelComputo`, `TestCasosDePlazos`, `TestAppletPlazos` y los guiones `h9-plazos-*`), `TestRegiones` sin diferencias (FR-020) y `TestPreceptosDeLosPlazos` en verde (sus fragmentos sustituidos siguen casando, spec §C.1); en una copia bajo `$TMPDIR`, quitar cada una de las diez reglas pone en rojo su subtest (mutantes de SC-005).

**Checkpoint**: los cómputos y los tipos son datos con sus tres ejes; `calcular` da lo mismo.

- [X] T005 [US4] `TestDescribeDePlazosSinDatos`, nuevo, en internal/app/plazos_describe_test.go (FR-012; US4.1; SC-005; contracts/applet-calcular.md §3; research D7). Ejecuta `kitlegal plazos calcular x x x --describe` por `registroDePlazos` sobre tres juegos de `FicherosDePlazos`: `fuentesDePlazos(t)`; los mismos con `Tipos` = data/plazos.yaml más un tipo de `lpac` y un periodo `aviso` de `lpac` con su código, su etiqueta y su texto, añadidos por sustitución de texto como `sustituida` de `grabaciones_test.go` (un `require` exige que `LeerTipos` lea esa copia sin defecto y que dé un tipo más que la real, que es lo que fija contracts/applet-calcular.md §3: el periodo `aviso` mal formado lo rechazan las reglas 7 a 9, y el código de aviso de más no se puede contar todavía, porque `CodigosDeAviso()` no recibe los tipos hasta T008 y el cómputo descrito no se exporta, research D1; esa comprobación la añade T008); y los mismos con `Tipos` vacío, que no carga. Exige en los tres código 0 y la salida igual, byte a byte, al golden `describe.json` de T003, que solo lee. Sin código de producto: `--help` y `--describe` ya no cargan los datos (research D5), y `compruebaDefectoDeComposicionDePlazos` y el caso «sin-los-tipos» no cambian (US4.2). **Verificación**: `go test -count=1 -run '^TestDescribeDePlazosSinDatos$' ./internal/app/` y `make ci` en verde; en una copia bajo `$TMPDIR`, la ayuda de hoy devuelta a `internal/app/plazos.go` pone el test en rojo (mutante de SC-005).

**Checkpoint**: el esquema de `plazos_calcular` es el mismo con cualquier `data/plazos.yaml`.

---

## Phase 5: User Story 3 — un cómputo de ensayo entra sin tocar Go (Priority: P1)

**Goal**: un solo algoritmo cuenta cualquier cómputo con sus ejes como datos (FR-004 a FR-008), probado con un cómputo
inventado que entra sin tocar Go (FR-022), y ningún Go del binario nombra un tipo, un cómputo ni un aviso de periodo
(FR-009, FR-010).

**Independent Test**: `TestComputoDeEnsayo` (E1 a E13) y `TestNingunCodigoDePlazosEnGo`, con todos los de H9 en verde.

- [X] T006 [datos] [US3] Los dos ficheros de prueba del dominio, sin código ni test (FR-022, FR-023; spec §C.9 y §C.10; contracts/datos-de-plazos.md §3 y §4; research D7, D8): internal/core/plazos/testdata/computo-de-ensayo.yaml, byte a byte el de §3 (cómputo `ensayo` con `municipios: [el del lugar del ensayo]`, `unidades: [dias-habiles, dias-naturales]`, el periodo `inhabil` del `04-06` al `04-10` y el `aviso` del `10-19` al `10-23` con `periodo-de-ensayo`; tipos `ensayo-fijo` y `ensayo-libre`; preceptos inventados de `BOE-A-0000-1` con `vigente_desde` `2000-01-01`), e internal/core/plazos/testdata/tipos-de-prueba-civil.yaml, byte a byte el de §4 (solo `tipos:` con `prueba-civil-dias`, con el 133.2, y `prueba-civil-meses`, con el 133.3, sin duración y con `da: vencimiento`, con la misma sangría que los tipos de data/plazos.yaml). Ningún test los lee todavía: los empotran T007, T008, T010 y T011. **Verificación**: `make ci` en verde; cada fichero empieza por su cabecera de comentarios de §3 y §4.

- [X] T007 [US3] El motor con los ejes como datos y `TestComputoDeEnsayo`, en una sola tarea porque el test es la prueba de que el motor no decide por códigos (FR-004, FR-005, FR-006, FR-007, FR-008, FR-011, FR-022; SC-001, SC-003; spec §C.3 a §C.6 y §C.10; contracts/api-del-dominio.md §2 a §4; research D2, D3, D13; plan.md «Diseño», paso 3). **Motor**: internal/core/plazos/computo.go, un solo algoritmo con lo que el ADR 0061 deja fijo (día inicial, meses de fecha a fecha con el último del mes, prórroga, inhábiles de base de los municipios que cuentan) y los periodos del cómputo del tipo: `inhabil` suma sus días a los inhábiles (y sus festivos que no son sábado ni domingo, a los saltados en días hábiles), `no-corre` mueve la partida y el vencimiento de los meses como research D9 de H9, y `da` decide la fecha (`dia-siguiente`, el día siguiente al vencimiento sin prórroga; un día después del 9999-12-31 es el error de hoy); sin `TipoSilencio` ni `ComputoContencioso`. internal/core/plazos/avisos.go: `pasaPor(desde, hasta, periodo)` con el periodo como dato, que puede cruzar el año; los avisos de periodo de los datos, `{codigo, texto: "⚠ <etiqueta>: <texto>"}`, una vez por plazo, detrás de los dos de los festivos que faltan y delante de los de un festivo de una parte, en el orden de los periodos de su cómputo (FR-008); desaparece `textoDelArt183` (su texto ya está en data/plazos.yaml desde T004); `CodigosDeAviso()`, `EtiquetaDeAviso(codigo)` y `AvisoDiasDelArt183` siguen hasta T008. internal/core/plazos/calcular.go: el número de municipios es el de los textos de su cómputo, y la unidad dada a un tipo sin duración que su cómputo no admite es el error de FR-007, tras los de la unidad y delante del número de municipios (research D13). internal/core/plazos/errores.go: el mensaje de municipios compuesto con los textos del cómputo y el de FR-007, con las formas de contracts/api-del-dominio.md §4 (unidades en el orden de `Unidades()`, unidas con `alternativas`). internal/core/plazos/tipos.go, solo si el motor necesita del cómputo descrito un dato que T004 no le deja al alcance; ninguna función suya decide ya por un código (`fijaSuDuracion` la retira T004, que deja de usar en el lector `CodigosDeTipo` y `Computos`, y estas dos las retira T008). **Test**: internal/core/plazos/ensayo_test.go, nuevo, `TestComputoDeEnsayo`: el fichero de ensayo de T006 empotrado con `//go:embed` (como `internal/core/ids/errores_test.go:156`), leído con `LeerTipos`, sobre el calendario de la rejilla de las propiedades, con un subtest por fila E1 a E13 de spec §C.10 (fecha, avisos en su orden, o código, clase y mensaje). internal/core/plazos/calcular_test.go e internal/core/plazos/avisos_test.go, solo lo que el cambio de firmas internas exija, sin cambiar ninguna entrada ni lo esperado de un caso (FR-017). **Verificación**: `go test -count=1 -run '^(TestComputoDeEnsayo|TestCalcular|TestComprobarPeticion|TestPropiedadesDelComputo)$' ./internal/core/plazos/` y `go test -count=1 -run '^(TestCasosDePlazos|TestAppletPlazos)$' ./internal/app/` en verde; `make ci` en verde (`h9-plazos-*` y los goldens `caso-a/b/e.json` sin cambiar); en una copia bajo `$TMPDIR`, los mutantes de plan.md (un motor que solo reconoce agosto, que ignora `dia-siguiente` fuera del silencio, que no comprueba las unidades del cómputo, que da los avisos de periodo con un texto fijo) ponen en rojo E1, E5, E12 y E7.

- [X] T008 [US3] Los avisos desde los tipos leídos y ningún código en Go, con `TestNingunCodigoDePlazosEnGo`, en una sola tarea porque cambiar las dos firmas y retirar las constantes obliga a cambiar a la vez todos sus usos (FR-009, FR-010; SC-003; contracts/api-del-dominio.md §1 y §5; contracts/applet-calcular.md §4; research D1, D4, D6). **Dominio**: internal/core/plazos/avisos.go, `CodigosDeAviso(tipos []Tipo) []string` (una lista nueva en cada llamada: los dos de los festivos que faltan, los de los periodos de los cómputos que nombran `tipos` en el orden de esos cómputos en el fichero y de sus periodos, y `festivo-de-una-parte`; con `nil`, los tres del motor) y `EtiquetaDeAviso(tipos []Tipo, codigo string) (string, bool)`; desaparece `AvisoDiasDelArt183`; siguen las tres constantes del motor. internal/core/plazos/tipos.go: desaparecen `TipoAlzada` … `TipoGenerico`, `CodigosDeTipo`, `ComputoLPAC`, `ComputoContencioso` y `Computos` (el tipo `Computo string` se queda). **`internal/evals`**: internal/evals/plazos.go lee `plazos.LeerTipos(data.Plazos)` una sola vez y saca de ahí, con `CodigosDeAviso(tipos)` y `EtiquetaDeAviso(tipos, …)`, las formas fijas de `formasDeAvisoDelPlazo`; `ExtraerAvisosDelPlazo` y `ComprobarCodigosDeAvisoDelPlazo` conservan su firma, y la segunda devuelve el error de lectura de los tipos si lo hay y sigue exigiendo que el esquema de eval publicado (`eval.yaml.json`, que no cambia) enumere exactamente esos códigos, los cuatro de hoy en su orden. **Tests de FR-017**, solo las constantes y las firmas: internal/core/plazos/avisos_test.go, internal/core/plazos/calcular_test.go, internal/core/plazos/propiedades_test.go, internal/core/plazos/tipos_test.go, internal/core/plazos/salida_test.go, internal/app/plazos_test.go, internal/skills/referencias_test.go, internal/evals/plazos_test.go (los de `CodigosDeAviso`) e internal/evals/formato_test.go (el subtest `enumerado-de-los-avisos` de `TestFormatoDePlazos`, con `ComprobarCodigosDeAvisoDelPlazo`). **`TestDescribeDePlazosSinDatos`** (internal/app/plazos_describe_test.go, de T005, no de H9): con `CodigosDeAviso(tipos)` ya posible, el `require` sobre la copia con un tipo y un periodo `aviso` más gana el que T005 no podía escribir, que `CodigosDeAviso` de los tipos de la copia dé exactamente un código más que el de los tipos de data/plazos.yaml, el del aviso añadido, en el lugar de su periodo (contracts/applet-calcular.md §3); lo demás del test no cambia. **Test nuevo**: internal/app/plazos_codigos_test.go, `TestNingunCodigoDePlazosEnGo` de contracts/applet-calcular.md §4: los códigos `tipos[].codigo`, `computos[].codigo` y `computos[].periodos[].aviso.codigo` de `../../data/plazos.yaml` y del fichero de ensayo de T006 (por su ruta relativa desde `internal/app`), con al menos uno de cada clase en cada fichero; cada `.go` que no acaba en `_test.go` de `../core/plazos`, `.`, `../skills` y `../evals`, sin subdirectorios, con `go/parser.ParseFile`; un defecto, con fichero, línea y código, por cada `*ast.BasicLit` de cadena cuyo valor (`strconv.Unquote`) es entero un código y por cada etiqueta de campo que lo lleva como palabra entera (`\b<código>\b`). **Verificación**: `go test -count=1 -run '^(TestNingunCodigoDePlazosEnGo|TestDescribeDePlazosSinDatos)$' ./internal/app/` y `go test -count=1 -run '^TestFormatoDePlazos$' ./internal/evals/` en verde; `make ci` en verde; en una copia bajo `$TMPDIR`, `if tipo.Codigo == "silencio"` en `internal/core/plazos` y la ayuda de hoy en `internal/app/plazos.go` ponen en rojo `TestNingunCodigoDePlazosEnGo` (mutantes de SC-003), y un periodo `aviso` más en `lpac` de data/plazos.yaml pone en rojo `enumerado-de-los-avisos` (control de FR-009); y la sustitución de `TestDescribeDePlazosSinDatos` sin su periodo `aviso` pone en rojo su `require` del código de más.

**Checkpoint**: el motor no nombra ningún tipo, cómputo ni aviso de periodo; con los datos de hoy, todo sale igual.

---

## Phase 6: User Story 6 — los preceptos del cómputo civil se comprueban contra lo grabado (Priority: P3)

**Goal**: el cómputo `civil` entra en `data/plazos.yaml`, sin tipos, cuando hay con qué comprobarlo: la raíz de H36 en
la unión y `TestPreceptosDeLosPlazos` sobre los cómputos sin tipos (FR-014 a FR-016). Va antes de US2 porque sus tests
leen el civil del `data/plazos.yaml` real.

**Independent Test**: `TestPreceptosDeLosPlazos` con los cinco preceptos del civil, y `TestGrabacionesSinSolape` con la
unión de cuatro conjuntos.

- [X] T009 [US6] La raíz de H36 en la unión, el civil en los datos y su comprobación, en una sola tarea porque cada pieza sin las otras deja `make ci` en rojo o el civil sin comprobar (FR-014, FR-015, FR-016; SC-006; contracts/grabacion.md §3 y §4; contracts/datos-de-plazos.md §1; research D9, D10, D11; plan.md «Desviaciones declaradas»). Parte de lo grabado por el paso `grabar_datos` tras T002. **Unión**: internal/evals/grabaciones.go, `GrabacionesDeH36` con la ruta de contracts/grabacion.md §3, y `UnionDeGrabaciones()` devuelve H4, H5, H9 y H36, en ese orden; `GrafosPrevios` sigue fuera. **Datos**: data/plazos.yaml gana el cómputo `civil` de contracts/datos-de-plazos.md §1 byte a byte (un municipio, `el de la sede del tribunal`; `dias-habiles` y `meses`; los periodos `inhabil` de agosto y del `12-24` al `01-06`, con el 130.2; los preceptos 130.2 —2022-12-23—, 133.1 —`fija: expiracion`, 2001-01-08—, 133.2 —2010-05-04—, 133.3 —2001-01-08— y 133.4 —2010-05-04— de `BOE-A-2000-323`), sin ningún tipo que cuente con él. **Tests**: internal/evals/grabaciones_test.go, solo tres cosas: la lista esperada de `TestGrabacionesSinSolape` gana `GrabacionesDeH36` (desviación de research D10); `verificarPreceptos` añade, detrás de `lecturasDeLosTipos(tipos)`, las lecturas de los preceptos de cada cómputo que no nombra ningún tipo, en el orden del fichero y nombradas «el cómputo `<código>`», con una estructura YAML propia del test sobre los mismos bytes (research D11); y `palabrasClaveDeLosPreceptos` gana en `a130` «24 de diciembre», «6 de enero» y «agosto», y en `a133` «día siguiente», «veinticuatro horas», «se excluirán los inhábiles», «de fecha a fecha» y «prorrogados». `lecturasDeLosTipos`, `loQueLlevaALeerElRepositorio`, `internal/evals/consultas.go` y su oráculo no cambian: el civil no entra en lo que la skill lleva a leer (FR-014). internal/skills/normas_test.go (de H5, no de H9), solo `rangosDeLasBusquedasGrabadas`, tercera desviación de research D10: el `require.NotEmpty` de «el conjunto de grabaciones %s tiene alguna búsqueda» (normas_test.go:642) pasa de cada conjunto a la unión —que entre todos los conjuntos haya alguna búsqueda grabada—, porque el conjunto de H36 solo guarda `robots.txt` y los bloques `a130` y `a133` (contracts/grabacion.md §2: la búsqueda «enjuiciamiento civil» la sirve el de H5) y sin el cambio `TestEsquemaDeNormas/rangos-grabados` sale en rojo; el comentario de la función dice lo mismo, y nada más del test cambia (el enum de rango sigue siendo el conjunto de rangos de todas las búsquedas grabadas de la unión). **Verificación**: `go test -count=1 -run '^(TestPreceptosDeLosPlazos|TestGrabacionesSinSolape|TestManifiestoDeGrabaciones|TestLecturasDeLosDatos|TestDirectorioDeLosGrafosPrevios|TestIdentificadoresDeLasNormas)$' ./internal/evals/` y `go test -count=1 -run '^TestEsquemaDeNormas$' ./internal/skills/` en verde (quickstart §6); `TestRegiones` sin diferencias (SC-007) y `make ci` en verde; en una copia bajo `$TMPDIR`, el `vigente_desde` del 133.1 con el 2010-05-04 y el del 130.2 con el 2015-10-07 ponen en rojo `TestPreceptosDeLosPlazos` (mutantes de SC-006), y `rutaDeLaBusqueda` con otra ruta, que deja la unión sin búsquedas, pone en rojo `TestEsquemaDeNormas/rangos-grabados` por el `require` de la unión (mutante de la desviación).

**Checkpoint**: el civil está en los datos, comprobado contra lo grabado del BOE, y no lo pide ningún tipo.

---

## Phase 7: User Story 2 — un cómputo de otra materia entra como datos: el civil de la LEC (Priority: P1)

**Goal**: el motor cuenta el civil de la LEC con sus datos: los goldens C1 a C10 y las propiedades sobre la rejilla
(FR-023, FR-024).

**Independent Test**: `TestComputoCivil` (C1 a C10 y que ningún tipo lleva el civil) y `TestPropiedadesDelComputo`
con la rejilla del civil.

- [X] T010 [US2] `TestComputoCivil`, nuevo, en internal/app/plazos_civil_test.go (FR-014, FR-023; US2.1 a US2.5, US5.1; SC-002; spec §C.9; research D7, D8). Combina el `data/plazos.yaml` real de `fuentesDePlazos(t)` con el fichero de tipos de prueba del civil de T006 (por su ruta relativa desde `internal/app`) sustituyendo la única línea `tipos:` del base por el contenido del de prueba, tras exigir con `require` que el base tiene exactamente una; lo lee `LeerTipos`; resuelve Madrid (`28079`), y Leganés en C8, con el registro de territorio como `comoLoCalculaElDominio`, y calcula con `plazos.Calcular` y los festivos reales de `data/festivos/`. Un subtest por fila C1 a C10 de §C.9 (fecha, festivos saltados, códigos de aviso, o código, clase y mensaje: C6 con código 3 y solo el 130.2 en el mensaje; C7 y C8 con código 2 y los mensajes literales de §C.9), y uno que exige que ningún tipo del `data/plazos.yaml` real cuenta con `civil` y que lo lee sin defecto (US2.5, US5.1). Sin código de producto: el motor es el de T007 y el civil, el de T009. **Verificación**: `go test -count=1 -run '^TestComputoCivil$' ./internal/app/` y `make ci` en verde; en una copia bajo `$TMPDIR`, los mutantes de plan.md (agosto `aviso` con su aviso completo; agosto `no-corre`; sin el periodo del 24 de diciembre; agosto desde el `08-02`; el calendario sin los locales de Madrid) ponen en rojo C1, C3, C2, C9 y C10 (mutantes de SC-002).

- [ ] T011 [US2] Las propiedades del cómputo civil en `TestPropiedadesDelComputo`, internal/core/plazos/propiedades_test.go (FR-024; SC-004; research D8). Sin cambiar la rejilla ni las propiedades de H9, gana las del civil sobre la misma rejilla con un municipio cada vez y los dos tipos de prueba de FR-023: la copia del contrato (`plazosDelContrato`, con el civil desde T004) combinada, como en T010, con el fichero de tipos de prueba de T006 empotrado con `//go:embed`; para cada inicio, duración y municipio de la rejilla, la fecha nunca es inhábil para la sede ni cae en un periodo `inhabil`; en días hábiles, entre el día siguiente al inicio y la fecha hay exactamente N días que no son inhábiles de base ni de un periodo; en meses, la fecha es la de fecha a fecha o, si esa es inhábil, el primer hábil siguiente. Sin código de producto. **Verificación**: `go test -count=1 -run '^TestPropiedadesDelComputo$' ./internal/core/plazos/` y `make ci` en verde; en una copia bajo `$TMPDIR`, el motor que cuenta como hábiles los días de un periodo `inhabil` lo pone en rojo (mutante de SC-004).

**Checkpoint**: el civil de la LEC se cuenta como datos, sin una línea de Go suya.

---

## Phase 8: Polish & Cross-Cutting Concerns

- [ ] T012 La entrada de *Unreleased* > «Cambiado» de CHANGELOG.md con las tres frases de FR-019: que los tipos y los cómputos de plazos son datos (`data/plazos.yaml`, ADR 0061); que el esquema de `plazos calcular` ya no enumera los tipos ni los avisos (lleva un `pattern`); y que el cómputo civil de la LEC está en el motor, pero no se pide todavía desde `kitlegal plazos` (FR-019; DoD 6). Nada más del fichero cambia. **Verificación**: `git diff -- CHANGELOG.md` solo añade esas líneas bajo *Unreleased*; `make ci` en verde.

- [ ] T013 El registro de los mutantes, specs/041-h36-el-motor-de/mutantes.md, y la comprobación final del diff (spec, «Controles»; plan.md «Mutantes» y «Controles de umbral»; ADR 0029; FR-017, FR-018, FR-020, FR-025; SC-007, SC-009). **Mutantes**: sobre una copia del árbol bajo `$TMPDIR`, nunca sobre el versionado, con los dos guiones de `specs/041-h36-el-motor-de/aceptacion/` copiados como `h36-plazos-de-hoy.txtar` y `h36-describe.txtar` donde los busca `TestEntregaDelHito` en la copia, se aplica cada mutante de la tabla «Mutantes» de plan.md, uno a uno, se ejecuta el control de su fila, se anota qué test y qué subtest cayó y se deshace; y, para que cada fila `ci:` de «Controles de umbral» tenga el suyo, también: un periodo `aviso` más en `lpac` de data/plazos.yaml de la copia (`enumerado-de-los-avisos`, FR-009) y el nombre de un tipo cambiado en data/plazos.yaml de la copia (`TestRegiones`, SC-007), y `make ci` de la copia con uno de ellos (SC-009). Cada fila del registro: control, mutante, orden ejecutada y test en rojo; un mutante que no pone su control en rojo es un defecto del control, que se arregla en el test de su tarea antes de cerrar esta. **Diff**: `git diff --name-only main -- specs/023-h9-internal-core-plazos/ skills/ evals/ data/festivos/ data/cobertura.yaml data/normas.yaml docs/ ':!docs/roadmap/in-progress/H36-el-motor-de-computo-descrito.md' ':!docs/roadmap/done/H36-el-motor-de-computo-descrito.md'` vacío (FR-018, FR-020, FR-025: ningún texto de skill cambia), y `git diff -M --name-status main -- docs/roadmap/` solo da una línea, el renombrado `R100` del fichero del hito de `docs/roadmap/in-progress/` a `docs/roadmap/done/`, que hace el workflow y no una tarea (commit 3c80c3f6); `git diff --name-only main`, sin pathspec, contrastado fichero a fichero con lo admitido, que es: las rutas que declaran T001 a T012; lo que el guardián admite sin declararlo (`go.mod`, `go.sum`, `CHANGELOG.md` y el `x_test.go` de un `x.go` declarado; «Format»); este directorio del feature; el fichero del hito que mueve el workflow; y lo que commitea el paso `grabar_datos` tras T002, que ninguna tarea declara porque ninguna puede escribirlo (el directorio de evidencias de H36, con lo grabado y su manifiesto, y el registro y el log de la grabación en gates/ de este directorio; contracts/grabacion.md §2), cuyos ficheros da `git show --name-only --format= <commit>` del commit que devuelve `git log --format=%h --grep='^chore(H36): grabaciones de' main..HEAD`, una lista que no nombra nada fuera de eso y del manifiesto de T002; un fichero fuera de lo admitido es un defecto de la tarea que lo cambió; y, en cada test y en cada fichero de datos de un test (los goldens y el esquema publicado de T003, los ficheros de prueba de T006, el manifiesto de T002, los guiones de aceptación de H39 en specs/031-h39-c-omo-se/aceptacion/), cambio a cambio con `git diff main -- <fichero>`, solo nombra lo de FR-017, los tests nuevos de «Trazabilidad» (`TestDescribeDePlazosSinDatos`, `TestComputoDeEnsayo`, `TestNingunCodigoDePlazosEnGo`, `TestComputoCivil`) con los ficheros de prueba de T006, y lo declarado fuera de FR-017: la raíz del cliente de `TestGrabarEvals` en internal/evals/grabacion_test.go (T002), la lista esperada de `TestGrabacionesSinSolape` en internal/evals/grabaciones_test.go (T009) y el `require` de la unión de `rangosDeLasBusquedasGrabadas` en internal/skills/normas_test.go (T009), las tres de research D10 y plan.md «Desviaciones declaradas», y la entrada de `probarLosDatosQueNoSeLeen` en internal/skills/regiones_test.go (T004, `gates/supuestos.md`); cualquier otro cambio en un test es un defecto que se arregla en la tarea que lo hizo, y en ningún caso cambian las entradas ni lo esperado de un caso de cálculo de H9; `make skills-check` y `make ci` en verde en el árbol (quickstart §7 y §8). Sin código de producto ni tests.

---

## Dependencies & Execution Order

- **T001** primero y antes de cualquier cambio de producto: captura con el binario del árbol, que es el de `main`.
- **T002** antes de T009: el paso `grabar_datos` graba tras su commit, y T009 añade la raíz de H36 a la unión, que no
  existe hasta entonces (research D9, V8).
- **T003** antes de T004 a T008: deja de usar `CodigosDeTipo` y `CodigosDeAviso` en los tests del esquema.
- **T004** antes de T005 a T011: la forma nueva de los datos y su lector.
- **T006** antes de T007, T008, T010 y T011: los ficheros de prueba que empotran o leen.
- **T007** antes de T008, T010 y T011: el motor por ejes. **T008** cuando ya no queda ninguna constante ni la ayuda de
  hoy (T003, T007) y existe el fichero de ensayo (T006).
- **T009** antes de T010: el civil en el `data/plazos.yaml` real.
- **T012** y **T013** al final; T013 la última, cuando todos los controles existen.

## Parallel Opportunities

Ninguna: el bucle ejecuta una tarea cada vez, y cada una parte del árbol que deja la anterior. Por ficheros, T005 y T006
no comparten ninguno, ni T010 y T011, pero T005 lee el golden de T003 y la forma de T004, y T010 y T011 el motor de
T007 y los ficheros de T006.

## Implementation Strategy

- **MVP**: T001 a T008 —el esquema sin datos, los datos con sus ejes, el motor genérico probado con el ensayo y ningún
  código en Go—, que entregan US1, US3, US4 y US5: con los datos de hoy todo sale igual y una materia entra como datos.
- **Incremento**: T009 a T011, el civil de la LEC como segunda implementación real, comprobado contra lo grabado (US6) y
  contado con sus goldens y sus propiedades (US2).
- **Cierre**: T012 y T013. Publicar la rama, medir CI y las evals de todas las skills (FR-025, SC-008) lo hace el
  workflow tras la revisión final; ninguna tarea lo hace.

## Trazabilidad

| Requisito | Tareas |
|---|---|
| FR-001, FR-002, FR-003 | T004 |
| FR-004, FR-005, FR-006, FR-007, FR-008 | T007 (y T010 con el civil) |
| FR-009, FR-010 | T008 |
| FR-011 | T001 (la captura), T004, T007 (lo mismo con los datos de hoy) |
| FR-012 | T001, T003, T005 |
| FR-013 | T004 |
| FR-014 | T009, T010 |
| FR-015, FR-016 | T002, T009 |
| FR-017 | T003, T004, T007, T008, T009; T013 lo comprueba |
| FR-018, FR-020 | T013 |
| FR-019 | T012 |
| FR-021 | T001 |
| FR-022 | T006, T007 |
| FR-023 | T006, T010 |
| FR-024 | T011 |
| FR-025 | el cierre del workflow; T013 comprueba que ningún texto de skill cambia |
| SC-001 | T001, T007, T013 |
| SC-002 | T010, T013 |
| SC-003 | T007, T008, T013 |
| SC-004 | T011, T013 |
| SC-005 | T003, T004, T005, T013 |
| SC-006 | T009, T013 |
| SC-007 | T004, T009, T013 |
| SC-008 | el cierre del workflow (job de evals) |
| SC-009 | cada tarea (`make ci`), T013 |

| Fila de «Controles de umbral» (plan.md) | Tarea que construye el control | Mutante que lo ve en rojo |
|---|---|---|
| SC-001, FR-011, FR-021 | T001 (`h36-plazos-de-hoy`); los de H9 ya existen y T004 y T007 los mantienen | T013 (art. 183, silencio con `vencimiento`, texto del municipio de `contencioso`) |
| SC-002, FR-023 | T010 | T010, T013 |
| SC-003, FR-022 | T007 | T007, T013 |
| SC-003, FR-010 | T008 | T008, T013 |
| SC-004, FR-024 | T011 | T011, T013 |
| SC-005, FR-012 | T003, T005; T001 (`h36-describe`) | T003, T005, T013 |
| SC-005, FR-013 | T004 | T004, T013 |
| SC-006, FR-015, FR-016 | T009 | T009, T013 |
| FR-009 | T008 | T008, T013 |
| SC-007, FR-020 | `TestRegiones`, que ya existe; T004 y T009 lo mantienen | T013 |
| SC-009, FR-017 | `make ci` en cada tarea | T013 |
| SC-008, FR-025 (evals) | el job de evals del cierre, sin tarea | — |

## Definition of Done (docs/ROADMAP.md §1)

| Punto | Cómo se cumple |
|---|---|
| 1 · `make ci` en verde | cada tarea; T013 en el árbol final |
| 2 · tests offline y fixtures grabados | los tests de T004, T005, T007 a T011, sin red; lo único que se graba lo graba `grabar_datos` desde el manifiesto de T002; los ficheros de prueba, en T006 |
| 3 · `net/http`, `os.Exit`, `fmt.Print*` | nada nuevo los usa (convenciones; `make lint`) |
| 4 · salida contra `schemas/` y `schema-check` | T003 regenera la parte de `plazos calcular` del esquema publicado; `TestAppletPlazos` sigue validando cada salida |
| 5 · errores tipados, sin `panic` | el error de FR-007, `errorDeArgumentos`, en T007; los defectos de `LeerTipos`, `inesperado`, en T004 |
| 6 · e2e y `CHANGELOG.md` | T001 (los dos guiones), T003 (el de H39 y su copia), T012 |
| 7 · ADR | no aplica: el ADR 0061 ya está en `main` y el spec excluye ADR nuevos y cambios de los ADR 0061 y 0042 («Fuera de alcance») |
| 8 · `docs/SOURCES.md` y `verify-sources.sh` | no aplica: la única fuente, el BOE consolidado, ya tiene su fila revisada y su caso; no cambia (contracts/grabacion.md §4) |
| 9 · cobertura de `internal/core/**` ≥ 85 % | `TestLeerTipos` (T004), `TestComputoDeEnsayo` (T007), `TestCalcular` y las propiedades (T011) cubren el código nuevo del dominio |
| 10 · evals de una skill que cambia | no aplica: ninguna skill cambia (FR-020); el cierre mide las de todas (FR-025) |
| 11 · municipio cubierto y no cubierto | Leganés y Madrid en T001 y T010; sin locales de 2027 en C2 y C4 (T010) y `82001` sin fichero local en E11 (T007) |
| 12 · grafo | no aplica: `plazos` no observa identidades |
| 13 · lo declarado y lo de fuera | no cambia: ningún tipo pide el civil (US2.5) y la cobertura de H33 sigue igual |
