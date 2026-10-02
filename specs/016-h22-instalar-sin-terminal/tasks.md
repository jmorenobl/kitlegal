# Tasks: H22 · Instalar sin terminal: la extensión de escritorio con el servidor y el plugin de Claude con las skills

**Input**: `specs/016-h22-instalar-sin-terminal/` (spec.md, plan.md, research.md, data-model.md, contracts/, quickstart.md)

**Prerrequisitos**: spec.md y plan.md leídos; `.specify/memory/constitution.md` 2.10.0 aplicada (umbral de materialidad,
criterio de uso, proporcionalidad y convergencia en «Gates»; controles de umbral del ADR 0029; ADR 0030, 0032 y 0035);
H0-H7.4, H19 y H21 en `main` (el kernel, el registro de applets, el servidor `kitlegal mcp serve` con sus diez
herramientas, las dos skills empotradas, goreleaser con sus seis archivos, `TestSnapshot`, `TestConfiguracionDeLaRelease`
y los flujos `ci.yml` y `release.yml`). **Ninguna tarea crea el esqueleto ni los gates: `make ci` existe y pasa**, y
cada tarea lo deja en verde al terminar.

**Aceptación**: plan.md dice «Aceptación e2e: no aplica» (research D13): el hito no añade ni cambia ningún
comportamiento del binario (FR-001, FR-080). **No hay tarea de aceptación, ni guiones `testscript` nuevos, ni evals
nuevas**: las skills no cambian. Lo que hace de aceptación en el run es la tabla de plan.md, «Aceptación e2e»: los tests
de `make ci` que traen T001 a T005, las seis subpruebas nuevas de `TestSnapshot` (T003), `TestPluginValido` (T004) y la
revisión final de la documentación (T006). El trabajo `snapshot` de la CI y el job de evals los mide el workflow en el
cierre, no una tarea (SC-001). SC-002 lo mide una persona tras publicar una release, fuera del run.

**Datos**: ninguno (plan.md, «Datos externos»; research D18). **No hay tarea de datos**: ninguna tarea crea ni modifica
ficheros de los directorios `testdata` ni `schemas`, ni de `data`; no hay manifiesto `grabaciones.json` nuevo ni test de
grabación, y el paso `grabar_datos` no tiene nada que grabar. El esquema oficial del manifiesto de MCP Bundle no se
versiona (FR-017; Clarifications, pregunta 2). Los tests crean en `t.TempDir()` sus binarios de prueba y un icono de
otro tamaño, y leen el icono versionado del árbol.

**Tests**: obligatorios (constitución §III; FR-066 a FR-069; plan.md, «Controles mecánicos»). Cada tarea de código trae
su test y la implementación mínima que lo hace pasar, en el mismo diff; los nombres de test son los de plan.md.

**Rebanadas verticales y excepciones declaradas.** T001 a T005 son rebanadas completas: su test y su implementación en
el mismo diff, `make ci` en verde por sí solas. Dos no llevan código con test propio, cada una por su razón y con su
verificación:

- **T006** es documentación: su verificación es que cada orden, fichero, test, objetivo y dirección que nombra existe
  en el árbol, que las cadenas literales de contracts/documentacion.md están tal cual, y `make ci`. La comprueba la
  revisión final contra FR-050 a FR-053 (SC-012, sin umbral numérico).
- **T007** es el cierre de la Definition of Done, sin código de producto: escribe la lista de lo tocado, mide la
  cobertura y ejecuta quickstart.md §1 a §9.

**Tamaño de T002.** El paquete del paso llega entero en una tarea (plan.md, «Orden de implementación», paso 2): sin el
`main`, nada usa `Ejecutar`, y lo que escribe las piezas es interno al paquete, así que por partes `unused` dejaría
`make ci` en rojo. Es un paquete nuevo, sin nada alrededor que cambie salvo las dos listas de la regla R1.

**Modo**: desatendido (ADR 0018). El workflow `hito` ejecuta y verifica las tareas **una a una**, con guardián de diff:
cada tarea declara en su propia línea, tras «Rutas:», **todas** las rutas que crea o modifica, y solo toca esas (más
`go.mod`, `go.sum`, `CHANGELOG.md`, el directorio del feature y el `x_test.go` de cada `x.go` declarado). Los demás
nombres que aparecen en una línea son contexto: referencias a los documentos del feature (contracts/, plan.md…) o
nombres de ficheros que el paso escribe en la salida de goreleaser, que no existen en el árbol versionado; lo que una
tarea solo lee se nombra sin su directorio. El cierre en la plataforma —publicar la rama, abrir la propuesta de cambio
y medir la CI, el trabajo `snapshot` y el job de evals— lo hace el workflow tras la revisión final; quickstart.md §10
queda para el workflow y para quien lea el informe final.

## Formato: `- [ ] Tnnn [Story?] Descripción — FR/SC. Rutas: …`

- **[P]**: ninguna tarea lo lleva; en el bucle del workflow todo va en secuencia.
- **[Story]**: US1 quien no usa una terminal instala kitlegal con dos ficheros · US2 el marketplace · US3 quien
  mantiene kitlegal publica las dos piezas con cada etiqueta · US4 el README · US5 un verbo o una skill nuevos llegan
  sin tocar el paso. La fundación (T001, T002) sirve a US1, US3 y US5 a la vez y no lleva historia; el cierre tampoco.
- Ninguna tarea lleva etiqueta de aceptación ni de datos (arriba, «Aceptación» y «Datos»).

## Batería de verificación por tarea

- **`make ci` en primer plano** y la tarea marcada `[X]` en el mismo turno en que termina en verde: formato · lint
  (`depguard`, `forbidigo`, `gosec`, `misspell`, `revive`, `dupl`, `gocyclo`…, también sobre los ficheros con la
  etiqueta `snapshot`) · tests con `-race` · `test-integration` · `test-tiempos` · `vuln` · `schema-check` ·
  `skills-check` · `goreleaser-check` · secretos · módulos. Ningún registro de la verificación en la raíz del
  repositorio. Las salidas de `go test` y de `make` se leen con `rtk proxy` o con sondas positivas, porque el proxy de
  la sesión las resume.
- **Sin red ni modelo**: ninguna tarea usa la red salvo la de las herramientas de Go (`make vuln`); ninguna ejecuta
  `claude`, `make plugin-check`, `make evals`, `make evals-sondeo` ni `release.yml`; ninguna etiqueta ni publica.
- **Sin personas a mitad**: ninguna tarea la hace, la revisa ni la desbloquea una persona.
- **Ningún test escribe fuera de `t.TempDir()`**; lo temporal de una tarea va bajo `$TMPDIR`, nunca bajo `/tmp`. El
  `dist` que deja `make release` en T003 y T007 se retira antes de terminar la tarea.
- **Sin atajos** (plan.md, «Comprobación contra la rúbrica», g): ningún `//nolint`, ningún `t.Skip`, TODO ni error
  silenciado; ninguna exclusión de lint nueva salvo la entrada de la lista `core` de `depguard` de T002. Si `dupl`,
  `gocyclo`, `paralleltest` o `misspell` marcan algo, se reestructura o se renombra (research V21: `Autoria`, no
  `Autor`; `os.OpenRoot` y `Root.WriteFile`, no `os.WriteFile`).
- **Los umbrales no se rebajan** (ADR 0029): ninguna tarea ni corrección cumple un umbral de plan.md, «Controles de
  umbral», rebajándolo, retirando su test o quitando una fila de una tabla de subpruebas. Cada control nuevo se ve en
  rojo antes de darlo por bueno: con sus casos negativos (T001, T002), con el árbol sin el cambio (T003) o con un
  mutante temporal del fichero que fija, que no queda en el diff (T003 a T005).
- **Antes de tocar un tipo o una lista que construyen los tests**, `grep` de sus usos en todos los `_test.go` del
  paquete: `release_test.go` lee `.goreleaser.yaml`, el `Makefile` y los flujos de forma estricta, y cambia en la misma
  tarea que el fichero que fija (research V24).
- **Proporcionalidad** (constitución, «Gates»): ninguna tarea añade un caso, un mensaje ni un test para un estado sin
  vía real —un `dist` tocado a mano entre goreleaser y el paso, un zip alterado tras publicarse, unos binarios que no
  son lo que su nombre dice, una carpeta de salida a medio escribir— (plan.md, «Trazabilidad», último párrafo).
- **Lo que no cambia** (FR-006, FR-053, FR-080): los applets, los verbos, `--describe`, las herramientas y las
  `instructions` del servidor; las dos `SKILL.md` y sus `references`; las evals; los seis archivos, los paquetes,
  `install.sh`, el cask y el bucket; `SOURCES.md`; la web; los specs, los planes y las suites congeladas de los hitos
  anteriores; el workflow `hito` y sus guiones; los ADR.

---

## Phase 1: Setup

No hay tareas: el repositorio, el `Makefile` y los gates existen desde H0, y el hito no añade ninguna dependencia
(plan.md, «Primary Dependencies»).

---

## Phase 2: Fundación — el paso que empaqueta (bloquea todas las historias)

**Propósito**: el programa del repositorio que escribe `kitlegal.mcpb`, `kitlegal-plugin.zip` y el catálogo, con
`tools` sacado del registro y `skills` de lo empotrado. Sin él no hay nada que goreleaser pueda ejecutar ni que el
snapshot pueda comprobar.

- [X] T001 Las herramientas anunciadas, desde el registro (contracts/paso.md §6 y §7; data-model §7; research D2, D15, V27; plan.md, «Orden de implementación», paso 1): en el fichero de herramientas del paquete `app`, el tipo exportado `HerramientaAnunciada{Nombre, Descripcion}` y la función `HerramientasAnunciadas(registro *Registro) []HerramientaAnunciada`, que da el nombre `<applet>_<verbo>` y la descripción de cada verbo en el orden del registro —applets por nombre, verbos en el orden de su catálogo, sin los de `mcp` ni los de `skills`— sobre el mismo recorrido (`verbosAnunciados`) del que el servidor saca lo que anuncia, sin cambiar `verbosAnunciados`, `NombresDeHerramientas` ni nada de lo que el servidor anuncia (FR-080). Con su test en el mismo diff, `TestHerramientasAnunciadas`: con el servidor en proceso y el cliente `mcptest`, sobre los applets de producción y con los de ejemplo añadidos, el nombre y la descripción de cada herramienta que el servidor lista son, comparados como conjunto y no por posición, los que da la función; el test se ve en rojo si la función omite una herramienta o cambia una descripción (se comprueba con un mutante temporal que no queda en el diff). `TestHerramientasDelServidor` no se toca y sigue en verde. Control de umbral de la fila «herramientas» de plan.md (FR-014, FR-061, SC-004), en su parte de `make ci`. — FR-014, FR-069, FR-080; SC-004. Rutas: internal/app/herramientas.go (y su `herramientas_test.go`).
- [X] T002 El paso entero, su `main` y su regla de arquitectura (contracts/paso.md §1 a §7; data-model §1 a §6; research D1, D3 a D5, D8, D14, D16, D17, D20, V21; plan.md, «Orden de implementación», paso 2). **Los textos** (`textos.go`): las constantes exportadas `NombreVisible`, `Descripcion`, `DescripcionLarga` y `Autoria` con los valores de data-model §5, el único sitio donde están escritas (FR-015), y el límite de 120 caracteres contados como caracteres y no como bytes. **Las piezas** (`piezas.go`): los dos zips reproducibles —orden fijo de entradas, sin entradas de directorio, Deflate de `archive/zip`, fecha 1980-01-01T00:00:00Z, modo `0755` en los dos binarios y `0644` en lo demás, sin comentario—; el `.mcpb` con exactamente cuatro entradas en el orden de data-model §2 (el manifiesto, el icono copiado byte a byte y los dos binarios copiados sin mirarlos); el manifiesto de un tipo con los campos de data-model §3 en ese orden y ninguno más, sin `user_config`, con `manifest_version` `0.3`, la versión tal cual la recibe y `tools` de las herramientas que se le dan; el plugin con `plugin.json` (los campos de data-model §4, sin `mcpServers`) y cada fichero del árbol de skills que se le da, con su ruta bajo `skills`, en el orden de `fs.WalkDir`, y nada más; los JSON con sangría de dos espacios, sin escapar `<`, `>` ni `&` y con salto de línea final; el rechazo de una descripción de más de 120 caracteres y de un icono que no es un PNG de 512 × 512 px (`image/png`, `DecodeConfig`); la escritura con `os.OpenRoot` y `Root.WriteFile` en una carpeta que tiene que existir. **El catálogo** (`catalogo.go`): el documento de data-model §6 y contracts/paso.md §4 a partir de la versión y de la huella —una sola entrada `kitlegal` de fuente `archive`, la dirección de la release de esa versión y no la de la última, los textos de `textos.go`—, con error si la versión está vacía o la huella no casa con `^[0-9a-f]{64}$`. **`Ejecutar`** (`ejecutar.go`): `Ejecutar(args []string, errores io.Writer) int` atiende las órdenes `piezas` y `catalogo` de contracts/paso.md §1 con `flag`, todas las banderas obligatorias; compone `piezas` con `app.RegistroDeProduccion`, `app.HerramientasAnunciadas` y `kitlegal.Skills()`; devuelve 0 o 1; en un fallo escribe en `errores` una línea que empieza por `empaquetar: ` con el texto de la tabla de contracts/paso.md §1; no escribe nada en la salida estándar ni pide nada a la red; ningún `panic`. **`doc.go`** con el comentario del paquete (un programa de construcción que el binario distribuido no enlaza) y **`export_test.go`** con lo interno que los tests necesitan. **El `main`**: `os.Exit(empaquetado.Ejecutar(os.Args[1:], os.Stderr))` y nada más. **La regla R1**: el paquete nuevo entra en la lista `core` de `depguard` y en `paquetesInternos` de `TestArquitectura`, que pasa de diez a once, con su comentario al día; y `TestElBinarioNoEnlazaElPaso`, junto a `TestElBinarioNoEnlazaLosEjemplos`: el cierre de `go list -deps` del binario distribuido no contiene el paquete del paso. **Tests del paquete, en el mismo diff, offline, con binarios de prueba y carpetas de `t.TempDir()`, y cada uno con el caso que lo pone en rojo** (contracts/paso.md §7): `TestPiezas` (las cuatro entradas exactas, en orden, con los bytes de los dos binarios y del icono y sus modos; el manifiesto y `plugin.json` leídos de forma estricta, campo a campo, con `mcp_config` con la orden y los argumentos del contrato; el plugin con cada fichero del árbol byte a byte y nada más; con una herramienta y una skill más en la entrada, aparecen sin tocar el paso, US5), `TestPiezasReproducibles` (dos ejecuciones en dos carpetas: los dos `.mcpb` iguales byte a byte, y los dos plugins también), `TestPiezasSinEntrada` (sin el binario de macOS, sin el de Windows, sin el icono y con una carpeta de salida que no existe: error que nombra lo que falta o lo que falló), `TestIcono` (el icono versionado del árbol es un PNG de 512 × 512 px y el paso lo acepta; uno de 256 × 256 y unos bytes que no son un PNG, creados en el test, lo hacen fallar), `TestDescripcionCorta` (`Descripcion` tiene 120 caracteres como mucho; una de 120 pasa y una de 121 falla, con caracteres de más de un byte), `TestCatalogo` (lectura estricta del documento; sin versión, con una huella de 63 dígitos o en mayúsculas, error) y `TestEjecutar` (`piezas` con la composición de producción: código 0, nada en la salida de error, `tools` igual a `app.HerramientasAnunciadas` del registro de producción —que contiene las diez de hoy, escritas en el test— y el árbol `skills` del plugin igual a `kitlegal.Skills()`, sin un fichero de más ni de menos; `catalogo`: el documento en el fichero de salida y código 0; sin orden, con una desconocida, sin una bandera y sin un binario: código 1 y una línea `empaquetar: …`). Son los controles de `make ci` de las filas «descripción», «icono», «0 bytes de diferencia», «entradas», «herramientas», «`skills` del plugin», «manifiesto», «binarios» (el paso copia sin cambiar un byte), «2 de 2 ficheros» (el paso escribe los dos), «servidor» (`mcp_config`) y «FR-001» de plan.md, «Controles de umbral». Antes de terminar, quickstart.md §2 a §4 a mano, con el temporal bajo `$TMPDIR` y retirado al acabar. `make schema-check` y `make skills-check` siguen sin drift: el binario distribuido no cambia. — FR-001, FR-004, FR-005, FR-010 a FR-017, FR-020 a FR-022, FR-030, FR-031, FR-066, FR-067, FR-069, FR-070; SC-004, SC-005, SC-007, SC-009, SC-010; US3 (escenarios 2, 3 y 6), US5. Rutas: cmd/empaquetar/main.go, internal/empaquetado/ (doc.go, textos.go, piezas.go, catalogo.go, ejecutar.go, export_test.go y sus `_test.go`), internal/arch_test.go, .golangci.yml.

**Checkpoint**: `go run` del paso escribe las dos piezas y el catálogo a partir de dos binarios de prueba; `make ci`
en verde; el binario distribuido es el de antes.

---

## Phase 3: User Story 1 — Quien no usa una terminal instala kitlegal con dos ficheros (P1) 🎯 MVP

**Goal**: el snapshot y la release dejan `kitlegal.mcpb` y `kitlegal-plugin.zip`, con sus huellas, junto a los seis
archivos de hoy, y `make snapshot-check` comprueba las dos piezas contra el binario real (US1, escenarios 1 a 6; US3,
escenario 1; US5).

**Independent Test**: `make release` y `make snapshot-check` (quickstart.md §5 y §6), con las diez subpruebas de
`TestSnapshot` en verde.

- [X] T003 [US1] goreleaser ejecuta el paso y el snapshot lo comprueba (contracts/release.md §1, §3 y §7; research D6, D7, D9, D11, D19, V1 a V11, V17 a V19, V23, V24, V32; plan.md, «Orden de implementación», paso 3). **Primero las seis subpruebas**, en un fichero nuevo de la raíz con la etiqueta `snapshot`, registradas en la tabla de `TestSnapshot` junto a las cuatro de hoy, que no cambian: `dos-piezas` (los dos ficheros en el directorio del snapshot y una línea `<sha256>  <nombre>` de cada uno en `checksums.txt`, con la huella del fichero; FR-060), `manifiesto-de-la-extension` (lectura estricta: `0.3`, cada campo de data-model §3 con su valor y ninguno más, `user_config` incluido; los textos, los del paquete del paso, importado como `paso`; `version` igual a la que imprime `version` el binario del snapshot de la plataforma del test, sin su `v`; `description` de 120 caracteres como mucho; FR-061, FR-066), `binarios-de-la-extension` (exactamente las cuatro entradas; el bit de ejecución del binario de macOS; con `debug/macho`, exactamente dos arquitecturas, una `amd64` y una `arm64` por su tipo de CPU y no por su posición, cada una byte a byte el `kitlegal` del archivo de macOS de su arquitectura; y el `.exe` byte a byte el del archivo de Windows `amd64`; FR-062), `icono-de-la-extension` (byte a byte el icono versionado, un PNG de 512 × 512 px; FR-066), `servidor-de-la-extension` (el `.mcpb` extraído en un temporal cuyo nombre lleva espacios, el binario de la plataforma del test puesto en el sitio del de macOS con un enlace simbólico y no copiándolo, la orden de `mcp_config` con `${__dirname}` resuelto, sus argumentos —que tienen que ser `mcp` y `serve`—, `/` como directorio de trabajo y la caché y el `HOME` en temporales: el cliente `mcptest` completa el saludo y lista, como conjunto de nombres y descripciones, exactamente `tools` del manifiesto, que contiene las diez de hoy escritas en el test; FR-061, FR-064) y `skills-del-plugin` (las entradas del plugin son exactamente `plugin.json` y, bajo `skills`, los ficheros que deja `kitlegal skills install` del binario del snapshot en un directorio vacío, sin su manifiesto, byte a byte; `plugin.json` leído de forma estricta, sin `mcpServers`, con la versión y los textos del manifiesto; FR-063). Cada una falla nombrando lo que falta, sobra o difiere. **Se ven en rojo** con `make release` y `make snapshot-check` sobre el árbol sin el cambio de goreleaser: faltan las dos piezas. **Después, `.goreleaser.yaml`**, con lo de contracts/release.md §1 y nada más: `id: kitlegal` en la construcción; `universal_binaries` con una entrada —`id: kitlegal-universal`, `ids: [kitlegal]`, `replace: false` y un solo gancho `post` con la orden de ese contrato, carácter a carácter, y `output: true`—; `ids: [kitlegal]` en `archives`; y los dos ficheros, en ese orden y detrás de `install.sh`, en `checksum.extra_files` y en `release.extra_files`. **Y, en la misma tarea, `TestConfiguracionDeLaRelease`**, que lee esa configuración de forma estricta: sus tipos ganan `builds[].id`, `universal_binaries` y `archives[].ids`; `plataformas` fija el `id`; la subprueba nueva `universal` fija la entrada única, `replace` escrito y falso, ningún gancho `pre` y la orden del gancho `post`; `archivos` fija `ids`; `checksums` y `publicacion`, los tres `extra_files` en orden; `probarSecretoDelPublicador` no cambia. Cada subprueba tocada se ve en rojo con un mutante temporal de `.goreleaser.yaml` que no queda en el diff: `replace: true`, el universal sin `id` propio, `archives` sin `ids`, un `extra_files` de menos. **Verificación**: `make goreleaser-check`; `make release` y `make snapshot-check` en verde, con las diez subpruebas de `TestSnapshot` —`seis-archivos` incluida: exactamente seis, sin ningún archivo `darwin_all`— y los guiones `instalador-`; se retira el directorio del snapshot; `make ci`. Son los controles de las filas «binarios», «2 de 2 ficheros» y «FR-006» de plan.md, «Controles de umbral», en `make ci`, y las medidas sobre el snapshot real de las filas «descripción», «icono», «entradas», «herramientas», «`skills` del plugin», «manifiesto» y «servidor», que el cierre cuenta. — FR-002, FR-005, FR-006, FR-060 a FR-064, FR-066, FR-068, FR-069; SC-003 a SC-007, SC-009, SC-011; US1 (escenarios 1 a 6), US3 (escenario 1), US5. Rutas: snapshot_piezas_test.go, snapshot_test.go, .goreleaser.yaml, release_test.go.

**Checkpoint**: `make release` deja las dos piezas con su línea en `checksums.txt` y los seis archivos de hoy;
`make snapshot-check` las comprueba.

---

## Phase 4: User Story 2 — El marketplace (P2)

**Goal**: el plugin del snapshot y el catálogo de su versión se validan con `claude plugin validate` en el trabajo
`snapshot` de la CI, con la versión de Claude Code de `evals.yml` (US2, escenario 1). Que el catálogo se publique con
cada etiqueta (US2, escenarios 2 y 3) llega con T005.

**Independent Test**: `TestConfiguracionDeLaRelease` en `make ci`; `make plugin-check` lo ejecuta el trabajo `snapshot`
en el cierre (SC-008), no una tarea.

- [X] T004 [US2] `make plugin-check` y sus dos pasos en el trabajo `snapshot` (contracts/release.md §2, §4, §5 y §7; research D10, V12, V26, S3, S4; plan.md, «Orden de implementación», paso 4). **`TestPluginValido`**, en el fichero de las subpruebas del snapshot, con la etiqueta `snapshot` y fuera de `TestSnapshot`: extrae el plugin del snapshot en un directorio temporal; escribe en otro el catálogo que da la orden `catalogo` del paso —por su función, importando el paquete como `paso`— para la versión de `metadata.json` del snapshot y la huella que `checksums.txt` da para el plugin; ejecuta `claude plugin validate .` en cada uno de los dos directorios, y falla, con lo que la orden escribió, si `claude` no está en el `PATH` o sale con un código distinto de 0; no abre ninguna sesión con modelo ni usa ninguna credencial; sin `t.Skip`. **`Makefile`**: el objetivo `plugin-check` con la receta de contracts/release.md §2, en `.PHONY` y con su línea de ayuda, fuera de los prerrequisitos de `ci`; la línea de ayuda de `release` nombra las dos piezas; el `Makefile` sigue sin nombrar `PUBLISHER_TOKEN`. **El trabajo `snapshot` de `ci.yml`**: detrás de sus cuatro pasos de hoy, los dos de contracts/release.md §5 —instalar Claude Code con `VERSION_DE_CLAUDE_CODE` y `make plugin-check`—, sin cambiar sus permisos, su runner ni la ausencia de secretos; el trabajo `ci` no se toca. **Y, en la misma tarea, `TestConfiguracionDeLaRelease`**: `objetivos-del-makefile` y `ayuda` fijan la receta y la línea de ayuda de `plugin-check` y que `ci` sigue sin él; `trabajo-de-snapshot`, los seis pasos en orden y que `VERSION_DE_CLAUDE_CODE` es igual a la de `jobs.evals.env` de `evals.yml`, que el test lee y la tarea no modifica. Cada subprueba tocada se ve en rojo con un mutante temporal que no queda en el diff: otra versión de Claude Code en `ci.yml`, el paso de validar retirado, `plugin-check` entre los prerrequisitos de `ci`. **La tarea no ejecuta `make plugin-check` ni `claude`**: comprueba que el fichero con la etiqueta `snapshot` compila (`go vet -tags=snapshot .`) y pasa el lint, y que `make ci` queda en verde. Es el control de la fila «`claude plugin validate`» de plan.md, «Controles de umbral»: en `make ci`, que ese trabajo lo ejecuta con la versión de `evals.yml`; la medida, en el cierre. — FR-023, FR-030, FR-065, FR-068; SC-008; US2 (escenario 1). Rutas: snapshot_piezas_test.go, Makefile, .github/workflows/ci.yml, release_test.go.

**Checkpoint**: el trabajo `snapshot` tiene seis pasos y `make ci` falla si su versión de Claude Code difiere de la de
`evals.yml`.

---

## Phase 5: User Story 3 — Quien mantiene kitlegal publica las dos piezas con cada etiqueta (P1)

**Goal**: `release.yml` atesta los dos ficheros, `humo` los comprueba desde fuera y un trabajo nuevo, `catalogo`,
escribe el catálogo de la etiqueta cuando `humo` sale en verde (US3, escenario 5; US2, escenarios 2 y 3). En el run se
comprueba su definición; su primera ejecución real es la release que una persona publique.

**Independent Test**: `TestConfiguracionDeLaRelease` en `make ci` (FR-068).

- [ ] T005 [US3] `release.yml`: la atestación, las seis comprobaciones nuevas de `humo` y el trabajo `catalogo` (contracts/release.md §6 y §7; research D11, D12, V20, V22, V30, S1, S2, S5, S11; plan.md, «Orden de implementación», paso 5). **`publicar`**: `subject-path` de la atestación gana las dos piezas, nueve sujetos; sus pasos, sus permisos y sus tokens no cambian (FR-003). **`humo`**: los seis pasos de hoy, sin tocar, y detrás los seis de contracts/release.md §6.2, con sus cuerpos tal cual —descargar la extensión, el plugin y los tres archivos de sus binarios; comprobar las cinco huellas contra `checksums.txt`; verificar la atestación de las dos piezas; comprobar con `jq` la versión y los campos fijos del manifiesto, `user_config` ausente; separar las dos arquitecturas del universal leyendo su cabecera con `od`, `head` y `tail`, reconocidas por su tipo de CPU, y comparar por su huella cada una con el binario de su archivo de macOS, y el `.exe` con el del archivo de Windows; y arrancar `kitlegal mcp serve` desde el binario de Linux de la release con el cliente de la tubería con nombre y fallar si los nombres que lista no son los de `tools` del manifiesto—, con las reglas de los de hoy: ninguna acción, sin obtener el código, `shell: bash`, el directorio temporal del runner, `set -euo pipefail` como primera orden, los permisos de solo lectura y sin secretos (FR-040 a FR-043). **`catalogo`**, nuevo, como en contracts/release.md §6.3: `needs: [humo]`, `ubuntu-latest`, `permissions` exactamente `contents: read`, y cuatro pasos —obtener el código sin dejar la credencial en el clon, instalar Go con `go.mod` y sin caché, componer el catálogo con la orden `catalogo` del paso para la etiqueta sin su `v` y la huella que el `checksums.txt` publicado da para el plugin, y escribirlo con la API de contenidos en el repositorio del catálogo, con `PUBLISHER_TOKEN` solo en el entorno de ese paso—; solo escribe el `marketplace.json` del catálogo y lo sustituye entero; si no puede escribir, el paso falla y su trabajo sale en rojo sin afectar a los otros dos (FR-031, FR-032). **Y, en la misma tarea, `TestConfiguracionDeLaRelease`**: `flujo-de-la-release` fija los tres trabajos; `atestacion`, los nueve sujetos; `humo`, doce pasos, con `comprobacionesDelHumo` de seis a doce entradas, las órdenes de cada paso nuevo como líneas enteras y `bash -n` sobre cada cuerpo; la subprueba nueva `catalogo`, de qué depende, su runner, sus permisos, que no tiene entorno de trabajo, sus cuatro pasos en orden, el entorno de componer (el token del flujo) y el de publicar (`secrets.PUBLISHER_TOKEN`), sus órdenes como líneas enteras y `bash -n`; y `tokens-de-la-publicacion`, que de todos los flujos `PUBLISHER_TOKEN` se nombra en dos sitios —el paso de goreleaser en `publicar` y el que publica el catálogo— y que `release.yml` no lee ningún secreto fuera de esos dos entornos. Cada subprueba tocada se ve en rojo con un mutante temporal de `release.yml` que no queda en el diff: un sujeto de menos, un paso de `humo` retirado, `catalogo` sin `needs`, con `contents: write` o con el secreto en el entorno del trabajo. **La tarea no ejecuta el flujo, no etiqueta y no publica**: los cuerpos de los pasos no se ejecutan contra ninguna release (research S1 y S2 siguen siendo supuestos, y el informe lo dice); lo que se comprueba es su definición, y `make ci` en verde. Es el control de la fila «3 trabajos, 2 pasos con el token, 9 sujetos, 12 pasos de `humo`» de plan.md, «Controles de umbral». — FR-003, FR-031, FR-032, FR-040 a FR-043, FR-068; SC-011; US3 (escenario 5), US2 (escenarios 2 y 3). Rutas: .github/workflows/release.yml, release_test.go.

**Checkpoint**: la definición de la release, entera, está fijada por `make ci`.

---

## Phase 6: User Story 4 — El README dice qué instalar, qué significa el aviso rojo y qué no está probado (P2)

**Goal**: la documentación deja de ser falsa (FR-052) y cuenta la instalación oficial (US4).

**Independent Test**: lectura del README, de CONTRIBUTING y de `CHANGELOG.md` contra contracts/documentacion.md
(SC-012), en la revisión final.

- [ ] T006 [US4] La documentación del hito (contracts/documentacion.md §1 a §4; plan.md, «Orden de implementación», paso 6). **README**: el apartado «Instalar» de hoy se parte en tres, en este orden y delante de «El servidor MCP»: `## Instalar sin terminal`, con los ocho contenidos de FR-050 en el orden de contracts/documentacion.md §1.1 y sus cadenas literales —que es la instalación oficial y dónde está probada, la app de escritorio de Claude en macOS, y que en Windows los pasos son los mismos y nadie los ha probado; los dos pasos, con lo que se ve en cada uno descrito en texto y sin imágenes, las dos direcciones fijas de «la última release» y el marketplace del catálogo como alternativa al zip; qué dice el aviso rojo de la app y qué hace kitlegal de verdad; que hacen falta las dos piezas y qué pasa con una sola, con la cita de ejemplo; cómo se actualiza cada una, dicho como lo previsto y sin probar; que con `kitlegal skills install` y el plugin las skills están dos veces en Claude Code; que en Linux no hay app de escritorio y se usa el binario con Claude Code, con enlace al apartado siguiente; y que en la web y en el móvil no funciona, con la línea `⚠ SIN CONSULTA AL BOE:`—; `## Instalar con la terminal`, el «Instalar» de hoy con otro título; y `## Otras instalaciones, sin probar`, con la extensión en Windows, lo que el README dice hoy de la app de ChatGPT, de Codex y de Antigravity, traído de «Instalar» y de «El servidor MCP», y dónde contar si funciona o qué falla (las incidencias del repositorio o `info@kitlegal.es`) (FR-051). Deja de decir que Claude Cowork y el chat de Claude «todavía no» y que llegarán con un plugin, y de anunciar el plugin en «Lo que viene»; «oficial» y «soportada» solo se dicen de la app de escritorio de Claude en macOS; nada de cómo se actualiza cada pieza ni de si la app admite el marketplace se afirma como probado (FR-052). **CONTRIBUTING**, en «La release» y en «Los controles», lo de contracts/documentacion.md §2: el paso que empaqueta —qué es, qué lee y qué escribe cada orden, de dónde salen la versión, `tools`, las skills y los textos, quién lo ejecuta y cómo probarlo en local—, `make release`, `make snapshot-check` con sus seis subpruebas nuevas, `make plugin-check`, `publicar`, `humo`, `catalogo` y `PUBLISHER_TOKEN` (qué escribe, qué dos pasos lo ven, y que su permiso sobre el catálogo se da antes de la primera release y fuera del run); y el enlace que hoy apunta al apartado «Instalar» del README pasa al de «Instalar con la terminal». **`CHANGELOG.md`**, *Unreleased*, «Añadido»: las tres entradas de contracts/documentacion.md §3; ninguna en «Cambiado». No se tocan `SOURCES.md`, la web, los specs anteriores, los ADR ni las dos `SKILL.md`. **Verificación**: cada orden, objetivo, test, fichero, trabajo y dirección que la documentación nombra existe tal cual en el árbol tras T001 a T005 (se comprueba con `grep` uno a uno), ningún enlace interno queda roto, y `make ci` en verde. — FR-050 a FR-053; SC-012; US4 (escenarios 1 a 3). Rutas: README.md, CONTRIBUTING.md, CHANGELOG.md.

**Checkpoint**: las cinco historias están entregadas hasta donde el run llega.

---

## Phase 7: Cierre — Definition of Done

- [ ] T007 Cierre de la Definition of Done, sin código de producto (`ROADMAP.md` §1; plan.md, «Gates» y «Obligaciones para tasks.md»): specs/016-h22-instalar-sin-terminal/cierre.md con (1) la lista, sacada de `git diff --name-status main`, de lo creado y lo modificado en el hito, y la comprobación de que no hay ningún fichero de los directorios `testdata`, `schemas`, `data`, `skills`, `evals`, `web` ni `docs`, ni de un spec anterior, ni del workflow `hito` (FR-053, FR-080): ni fuente nueva, ni fila nueva en `SOURCES.md`, ni esquema, ni fixture, ni ADR —el hito no cambia ninguna decisión de arquitectura que no cambiara ya el ADR 0035—; (2) la cobertura sobre `coverage.out` y `coverage-integration.out` de `make ci` con `go tool cover -func`: global ≥ 70 % y el dominio ≥ 85 % (Definition of Done §1.9), con la del paquete del paso; (3) quickstart.md §1 a §9 ejecutados tal cual, en orden y en primer plano, con la salida de cada uno frente a lo esperado —§5 y §6 son `make release` y `make snapshot-check`, y §9 retira el directorio del snapshot—, y §10 anotado como no ejecutado: `make plugin-check`, `release.yml`, la prueba de SC-002 y el cierre son de la persona o del workflow; (4) por cada fila de plan.md, «Controles de umbral», el test que la controla, que existe en la cabeza y que `make ci` lo ejecuta, y cuáles de las cinco medidas sobre el snapshot real quedan para el trabajo `snapshot` del cierre; (5) los supuestos que siguen sin medir (research S1 a S11), dichos como tales; y (6) lo que queda para la persona tras el run: que `PUBLISHER_TOKEN` pueda escribir en el repositorio del catálogo, la etiqueta, la release, la prueba de SC-002 con sus cuatro anotaciones y versionar el esquema oficial del manifiesto. Si algo de (1) a (4) no cuadra, se dice tal cual en cierre.md: esta tarea no corrige código. `make ci` en verde. — FR-053, FR-068, FR-069, FR-070, FR-080; SC-001, SC-011. Rutas: specs/016-h22-instalar-sin-terminal/cierre.md.

---

## Dependencias y orden de ejecución

Secuencia estricta, la del plan (de dentro afuera): **T001 → T002 → T003 → T004 → T005 → T006 → T007**. Ninguna tarea
depende de una posterior.

- **T002** necesita `app.HerramientasAnunciadas` (T001).
- **T003** necesita el paso (T002): el gancho de goreleaser lo ejecuta, y las subpruebas importan sus textos.
- **T004** necesita el fichero de subpruebas del snapshot (T003) y la orden `catalogo` (T002).
- **T005** necesita la orden `catalogo` (T002), que el trabajo nuevo ejecuta, y las dos piezas en la release (T003).
  Va detrás de T004 porque las dos cambian `release_test.go`.
- **T006** documenta lo que T001 a T005 dejan; **T007** mide el árbol entero.

Historias: US1 y US5 quedan entregadas con T003; US2, con T004 (la validación) y T005 (la publicación); US3, con T002
(escenarios 2, 3 y 6), T003 (1), T005 (5) y `make ci` en cada tarea (4); US4, con T006.

### Oportunidades de paralelismo

Ninguna dentro del run: el workflow ejecuta una tarea cada vez, y T003, T004 y T005 comparten `release_test.go`. Por
eso ninguna tarea lleva `[P]`.

## Trazabilidad

### Requisitos → tareas

| Requisito | Tareas |
|---|---|
| FR-001 | T002 (el paso, su `main`, `TestElBinarioNoEnlazaElPaso`), T003 (el gancho) |
| FR-002 | T003 |
| FR-003 | T005 |
| FR-004, FR-067 | T002 (`TestPiezasReproducibles`) |
| FR-005 | T002 (`TestPiezasSinEntrada`, `TestEjecutar`), T003 (el gancho hace fallar a goreleaser) |
| FR-006 | T003 |
| FR-010 a FR-013, FR-017 | T002 (`TestPiezas`), T003 (subpruebas del snapshot) |
| FR-014 | T001, T002 (`TestEjecutar`), T003 (`servidor-de-la-extension`) |
| FR-015, FR-016, FR-066 | T002 (`TestDescripcionCorta`, `TestIcono`), T003 |
| FR-020 a FR-022 | T002, T003 (`skills-del-plugin`) |
| FR-023 | T004 |
| FR-030 | T002 (`TestCatalogo`), T004 |
| FR-031, FR-032 | T002 (la orden `catalogo`), T005 |
| FR-040 a FR-043 | T005 |
| FR-050 a FR-052 | T006 |
| FR-053 | T006, T007 (`SOURCES.md` sin cambios) |
| FR-060 a FR-064 | T003 |
| FR-065 | T004 |
| FR-068 | T003, T004, T005, T007 |
| FR-069 | T001, T002, T003, T007 (cobertura) |
| FR-070 | T002; la tabla de abajo; T007 |
| FR-080 | T001 (no cambia lo anunciado), T007 |
| SC-001 | lo mide el cierre del workflow; T007 deja preparado lo que el run puede medir |
| SC-002 | humana, fuera del run: ninguna tarea |
| SC-003 a SC-007, SC-009 | T002 (en `make ci`), T003 (sobre el snapshot) |
| SC-008 | T004 (el control); la medida, en el cierre |
| SC-010 | T002 |
| SC-011 | T003, T004, T005, T007 |
| SC-012 | T006; la comprueba la revisión final |

### Controles de umbral (plan.md) → tarea que construye el control

| Fila de plan.md | Test de `make ci` | Tarea | Cómo se ve en rojo |
|---|---|---|---|
| Descripción, 120 caracteres | `TestDescripcionCorta` | T002 | una de 121 falla |
| Icono, PNG de 512 × 512 px | `TestIcono`, `TestPiezas` | T002 | uno de 256 × 256 y unos bytes que no son un PNG fallan |
| 0 bytes entre dos ejecuciones | `TestPiezasReproducibles` | T002 | compara byte a byte dos ejecuciones |
| 4 entradas en el `.mcpb` | `TestPiezas` | T002 | la lista exacta de entradas, con sus modos |
| `tools`, el conjunto exacto | `TestEjecutar`, `TestHerramientasAnunciadas` | T002, T001 | las diez de hoy escritas en el test; mutante en T001 |
| 0 ficheros distintos en `skills` | `TestEjecutar`, `TestPiezas` | T002 | ni uno de más ni de menos frente a lo empotrado |
| Manifiesto `0.3`, campos exactos | `TestPiezas` | T002 | lectura estricta, campo a campo |
| Binarios, byte a byte | `TestPiezas`, `TestConfiguracionDeLaRelease` | T002, T003 | mutante de `universal_binaries` |
| 2 de 2 ficheros y huellas | `TestPiezas`, `TestConfiguracionDeLaRelease` | T002, T003 | un `extra_files` de menos |
| 6 archivos, ninguno universal | `TestConfiguracionDeLaRelease` | T003 | `archives` sin `ids`, `replace: true` |
| El servidor desde el `.mcpb` | `TestPiezas`, `TestEntregaDelHito` (ya existe, sin tocar) | T002 | `mcp_config` con otra orden u otros argumentos |
| `claude plugin validate`, 2 de 2 | `TestConfiguracionDeLaRelease` | T004 | otra versión de Claude Code, el paso retirado |
| 0 paquetes del paso en el binario | `TestElBinarioNoEnlazaElPaso` | T002 | el cierre de `go list -deps` |
| 3 trabajos, 2 pasos con el token, 9 sujetos, 12 pasos | `TestConfiguracionDeLaRelease` | T005 | mutantes de `release.yml` |

Las cinco medidas que solo existen sobre el snapshot real o con Claude Code —los bytes de las arquitecturas, los dos
ficheros en el snapshot, los seis archivos, el servidor arrancado desde el `.mcpb` y `claude plugin validate`— tienen su
control en `make snapshot-check` (T003) y en `make plugin-check` (T004), que ejecuta el trabajo `snapshot` y cuenta el
cierre del workflow (plan.md, «Lo que solo se mide sobre el snapshot real»).

### Definition of Done (`ROADMAP.md` §1) → tareas

| Punto | Aplica | Tarea |
|---|---|---|
| 1 · `make ci` en verde | sí | todas |
| 2 · tests unitarios offline; fixtures si toca red | tests sí; fixtures no (no toca red) | T001 a T005; T007 lo constata |
| 3 · sin `net/http`, `os.Exit`, `fmt.Print*` fuera de lo autorizado | sí | T002 (`os.Exit` solo en el `main`; R1) |
| 4 · esquemas y `schema-check` | sin esquema nuevo: el binario no cambia | T002 y T007 comprueban que no hay drift |
| 5 · errores con código estable, sin `panic` | el paso sale con 0 o con 1 (research D14) | T002 |
| 6 · e2e y `CHANGELOG.md` si cambia un comportamiento visible | e2e no aplica; `CHANGELOG.md` sí (FR-053) | T006 |
| 7 · ADR | no: ninguna decisión nueva (spec, «Relación con H19 y H21») | T007 lo constata |
| 8 · `SOURCES.md` | no: ninguna fuente nueva; queda sin cambios (FR-053) | T007 lo constata |
| 9 · cobertura | sí | T007 |
| 10 · evals y skills | no: ninguna skill se entrega ni cambia | T007 lo constata |
| 11, 12, 13 · territorio, grafo, plazos | no | — |

## Estrategia de implementación

1. **Fundación** (T001, T002): el paso existe y se prueba entero en `make ci`, sin construir ninguna plataforma.
2. **MVP — US1** (T003): el snapshot deja las dos piezas y las comprueba contra el binario real. Con esto, quien
   construye una release ya tiene los dos ficheros con su huella.
3. **US2 y US3** (T004, T005): la validación del plugin y del catálogo en la CI, y la release que atesta, comprueba
   desde fuera y publica el catálogo.
4. **US4** (T006): la documentación.
5. **Cierre** (T007): lo que el run puede medir, medido, y lo que no, dicho.

Tras el bucle, el workflow hace la revisión final y el cierre: publica la rama, abre la propuesta de cambio y mide la
CI —con el trabajo `snapshot`: `make release`, `make snapshot-check` y `make plugin-check`— y el job de evals (SC-001).
La etiqueta, la release y la prueba de SC-002 son de una persona (ADR 0020).

## Comprobación contra la rúbrica de `juez_tasks` y `precheck.sh tasks`

- **precheck**: siete tareas con el formato `- [ ] Tnnn`, ids únicos y correlativos; cada línea declara rutas; ninguna
  nombra los directorios protegidos de fixtures ni de esquemas, ni el de evidencias; ninguna es de plataforma ni exige
  a una persona; ninguna tarea de aceptación, como pide «Aceptación e2e: no aplica». `gates/analyze.md` lo escribe el
  paso siguiente del workflow.
- **b · trazabilidad**: las tres tablas de «Trazabilidad».
- **c · rebanadas verdes**: T001 a T005, test e implementación en el mismo diff; T006 y T007, declaradas arriba.
- **d · rutas declaradas**: tras «Rutas:» en cada línea, ficheros o el directorio del paquete nuevo.
- **e · datos separados**: no hay tareas de datos; ninguna toca fixtures ni esquemas.
- **f · Definition of Done**: la tabla de arriba.
- **h · aceptación primero**: el plan justifica «no aplica» (research D13).
- **i · autonomía**: ninguna tarea exige a una persona, publica ni mide en la plataforma.
