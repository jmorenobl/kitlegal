---

description: "Tareas de H45 · El molde admite un paquete con sus plazos y su norma, sin código (ADR 0062)"
---

# Tasks: H45 · El molde admite un paquete con sus plazos y su norma, sin código (ADR 0062)

**Input**: `specs/042-h45-el-molde-admite/` — spec.md, plan.md, research.md (P1-P3, D1-D14), data-model.md,
contracts/ (sincronia-de-los-paquetes.md, skills-y-laboratorio.md, evals.md, tests-sin-listas.md, aceptacion.md),
quickstart.md.

**Tests**: cada tarea lleva su test y la implementación mínima que lo pone en verde, y deja `make ci` en verde al
terminar. La suite de aceptación (T001) se escribe primero, falla hoy por una aserción y queda congelada; el workflow la
activa al final del bucle y `make ci` tiene que pasarla.

**Reglas que valen para todas las tareas**: sin red (salvo la de las herramientas de Go), sin `KITLEGAL_RECORD`, sin
sesiones con modelo; nada se escribe en el directorio `evidencias` ni en `data/` (el hito no cambia `data/` ni graba
nada: plan, «Datos externos»); ninguna eval cambia (FR-025); ningún `//nolint` ni `t.Skip` nuevo; cada mutante se
aplica en una copia del módulo en `$TMPDIR`, nunca en el árbol de trabajo (research D13). «La copia con un paquete
más» es la de data-model §6 (la que monta `paquete-de-mas.txtar`), en `$TMPDIR`.

## Format: `[ID] [P?] [Story] Description`

- **[P]**: puede ir en paralelo con las demás [P] de su fase (ficheros distintos, sin dependencias entre ellas).
- **[USn]**: la historia del spec a la que sirve.
- **[aceptacion]**: la suite congelada (solo T001). **[datos]**: toca ficheros de prueba bajo `testdata/` y no mezcla
  otro trabajo salvo lo que tiene que cambiar en el mismo commit (se dice en la tarea).

---

## Phase 1: Setup — la suite de aceptación

**Propósito**: los dos guiones de la entrega (contracts/aceptacion.md), rojos hoy por una aserción.

- [X] T001 [aceptacion] Escribir, a partir del spec y sin código de producto, los dos guiones testscript de la entrega en `specs/042-h45-el-molde-admite/aceptacion/paquete-de-mas.txtar` y `specs/042-h45-el-molde-admite/aceptacion/skills-instaladas.txtar`, cada uno con un comentario que nombra lo que cubre, y dar a los guiones del arnés `GOCACHE` y `GOMODCACHE` en `internal/app/e2e_test.go` (research D4: la preparación del arnés lee una vez `go env -json GOCACHE GOMODCACHE` y el `Setup` de `ejecutarGuiones` pasa las dos variables a cada guion, como `prepararInstalacion` de los tests de instalación del paquete Go de las skills, que no se tocan; ninguna orden ni variable más). `paquete-de-mas.txtar` (US4, US2.1-US2.2 · FR-001, FR-002, FR-005, FR-006, FR-016, FR-017 · SC-004): exactamente los pasos 1-4 de contracts/aceptacion.md §1 —el `env` de Go (`HOME=$WORK/casa`, `GOPATH`, `GOPROXY=off`, `GOENV=off`, `GOFLAGS=-mod=readonly`); `exec go run paquete-de-mas.go $KITLEGAL_SKILLS/.. repo`, con el programa de solo biblioteca estándar dentro del archivo, que copia el módulo sin `.git`, `web`, `dist`, `bin`, `.codegraph`, `specs`, `.specify`, `.agents`, `.claude` ni enlaces simbólicos, instala el paquete `prueba` de data-model §6 (entradas al final del bloque de `normas`, `tipos`, `cubierto` y `regulados_aparte`; `prueba.md` en `plazos-y-territorio`; el primer fichero del directorio de evals de `plazos-y-territorio` copiado como `99-<nombre>`), sincroniza con `TestSkillsDelRepositorio -args -regenerar-skills` y corre las dos órdenes `go test -count=1` con los filtros literales del contrato, imprime `copia: hecha`, `sincronía: <código>`, `tests enteros: <código>` y `subtests: <código>`, escribe el principio de la salida de una orden que falle, hasta 1200 bytes, en la salida de errores y sale con 0 salvo que no pueda copiar o instalar; las aserciones `stdout` de las cuatro líneas con código 0; y los `grep` de `prueba.md` (el tipo `prueba-de-un-paquete-mas`, la fila «procedimientos del paquete de prueba de un paquete más» y la de «procedimientos sancionadores en materia de tráfico y seguridad vial») y los `! grep` de `administrativo.md` (ni `BOE-A-2015-11722` ni el tipo)—, con los ficheros `paquete-de-mas.go`, `paquete/normas.yaml`, `paquete/plazos.yaml`, `paquete/cobertura.yaml` y `paquete/prueba.md` dentro del archivo. `skills-instaladas.txtar` (US5.1, US9.1 · FR-009, FR-023 · SC-007): contracts/aceptacion.md §2 —`kitlegal skills install --dir destino --json`; `! grep 'En PowerShell 5\.1:'` en `destino/normativa/SKILL.md`; las cuatro líneas de §C.2 al final de `destino/normativa/references/con_herramientas.md`, tras una línea en blanco y ancladas con `\z`; y la `description` de FR-023 en el frontmatter instalado de `plazos-y-territorio` con `( |\n  )` entre palabras, como `h41-4-skills-instaladas.txtar`—. Los guiones afirman solo lo que fija el spec. Comprobar que cada uno falla hoy por una aserción (`paquete-de-mas`: `sincronía: 1`; `skills-instaladas`: la frase de PowerShell en el cuerpo y la descripción que nombra la LPAC) y que `make ci` sigue en verde con el cambio del arnés.

**Checkpoint**: la suite existe, falla hoy por una aserción y queda congelada.

---

## Phase 2: Foundational — el `plazos.yaml` de prueba de `TestRegiones`

- [X] T002 [datos] Crear `internal/skills/testdata/regiones/plazos.yaml`, copia de los tipos y cómputos de hoy del `plazos.yaml` del repositorio, que no se toca (contracts/tests-sin-listas.md, fila `TestRegiones`; research D10; FR-016, SC-003), con una cabecera que dice que es un dato de prueba de las reglas de las regiones y no el repositorio. Ningún test lo lee todavía: `make ci` sigue en verde.

**Checkpoint**: T003 puede pasar `TestRegiones` a datos de prueba.

---

## Phase 3: User Story 3 — el paquete de un tipo y sus dos defectos (Priority: P1) 🎯 MVP

**Goal**: cada tipo va solo a la referencia de su paquete; un tipo sin paquete o de dos es un defecto de la sincronía
(FR-001 a FR-004), y con los datos de hoy nada cambia (US1.1, FR-008).

**Independent Test**: `go test -run '^TestRegiones$' ./internal/skills/` con sus tres subtests de defectos, y
`make skills-check` sin diferencias.

- [X] T003 [US3] El paquete de un tipo en `internal/skills/regiones.go` y sus tests en `internal/skills/regiones_test.go` (FR-001, FR-002, FR-003, FR-004, FR-008, FR-016; US1.1, US2.1, US3.1-US3.4; SC-001, SC-003; contracts/sincronia-de-los-paquetes.md §1, §2, §6; research D1, D12): `paquetesDelTipo(tipo, normas)` da los `vertical` distintos, en orden alfabético, de las normas de `Preceptos` y de `Unidades` (sin las normas que no están o no llevan `vertical`, y sin mirar `PreceptosDelComputo`); `filasDeLosRecursosYPlazos` usa el contexto de su región: junta en el orden de los tipos del `plazos.yaml` leído un defecto por tipo con 0 paquetes o con 2 o más, con los textos literales de spec §C.1 (FR-003 y FR-004), y con alguno devuelve `&defectoDeLosDatos{fichero: "plazos.yaml", causa: errors.Join(…)}`; sin defectos, `filasDeLosTipos` sin cambiar sobre los tipos cuyo paquete es el de la región. `TestRegiones` deja de copiar el `plazos.yaml` real del repositorio (`arbolDeRegiones`) y lee el `plazos.yaml` de prueba de T002; `normasDeLasRegiones` gana una norma de prueba con `vertical: ensayo`; y `TestRegiones/defectos` gana `tipo-sin-paquete` (el defecto con su código y ninguna referencia regenerada), `tipo-de-dos-paquetes` («administrativo, ensayo») y `computo-de-otro-paquete` (0 defectos y el tipo en `administrativo.md`). Ver en rojo cada subtest nuevo con la regla quitada y el tercero contando los preceptos del cómputo, y `TestRegiones/sincronia` con una región distinta de la generada; comprobar que `make skills-sync` no deja diferencias en las referencias de las skills (`make skills-check` en verde) y `make ci` en verde.

**Checkpoint**: el molde de los tipos funciona y lo administrativo sale igual.

---

## Phase 4: User Story 5 — `normativa` hace sitio y el control de PowerShell lo admite (Priority: P1)

**Goal**: la forma para PowerShell 5.1 del paso 2 pasa a `references/con_herramientas.md`; el control de las órdenes y
la declaración del laboratorio lo admiten (FR-009 a FR-013). Va antes de la vertical crecida porque el sitio de FR-020
(T006) lo necesita.

**Independent Test**: `go test -run '^(TestEvalsDelRepositorio|TestControlesSobreLasReferenciasAMano|TestSkillsDelProductoDesdeControl)$' ./internal/evals/`.

- [X] T004 [US5] El movimiento de `normativa` con el control y la declaración que lo admiten, en una sola tarea porque en cuanto el bloque sale del cuerpo `ordenes-para-powershell` y `TestSkillsDelProductoDesdeControl` lo ven a la vez (FR-009, FR-010, FR-011, FR-012, FR-013; US5.1-US5.4; SC-008; contracts/skills-y-laboratorio.md §1-§3; research D6, D7): quitar de `skills/normativa/SKILL.md` sus líneas 38 a 41 («En PowerShell 5.1:», las dos vallas y la orden, con su sangría) sin tocar nada más del cuerpo; añadir al final de `skills/normativa/references/con_herramientas.md`, tras una línea en blanco, esas cuatro líneas sin la sangría, byte a byte como §C.2; en `internal/evals/conjunto_test.go`, `controlDeLasOrdenes` da por cubierta una orden de Bash del `SKILL.md` sin su forma detrás si alguna referencia de `skills.ReferenciasAMano` de la misma skill lleva una orden PowerShell con el mismo verbo, la misma norma y los mismos bloques (`esSuFormaParaPowerShell`, sin cambiar), sin ninguna condición sobre cuándo se lee la referencia, y las órdenes de una referencia siguen exigiendo la suya detrás en su fichero, sin cambiar ningún mensaje; `mutantesDeLasOrdenes` de `TestControlesSobreLasReferenciasAMano` gana los casos 1 a 4 de §C.3 (`orden-del-skill-md-con-su-forma-en-una-referencia` sin defecto; `forma-en-una-referencia-con-otros-bloques`, `forma-en-ninguna-referencia` y `orden-de-una-referencia-con-su-forma-en-el-skill-md` con su defecto), conserva `orden-con-su-forma-en-otra-referencia` y `sin-las-ordenes-de-articulos` y no gana ninguno más; en `internal/evals/laboratorio_familias_test.go`, la declaración de `normativa` gana el `cambio` (las cuatro líneas con sangría → vacío) y el `movimiento` a `con_herramientas.md` (`"\n"` + las cuatro líneas sin sangría + `"\n"`); y la fila de `ordenes-para-powershell` de `docs/SKILLS.md` dice el texto de contracts/skills-y-laboratorio.md §2. Ver en rojo los cinco mutantes de §C.3 y el subtest `movimientos` con el texto movido de vuelta al cuerpo; `make ci` en verde.

**Checkpoint**: el cuerpo de `normativa` tiene sitio para un paquete real más.

---

## Phase 5: User Story 2 — la vertical de ensayo entra con un tipo y una fila con norma (Priority: P1)

**Goal**: el tipo y la fila de la vertical salen solo en `ensayo.md`, sin Go ni `SKILL.md` a mano, y cada cuerpo de
skill de capacidad tiene sitio para otra línea de paquete (FR-005 a FR-007, FR-019, FR-020); el mutante de la regla de
hoy la pone en rojo (US1.4).

**Independent Test**: `go test -run '^TestVerticalDeEnsayo$' ./internal/skills/`.

- [X] T005 [datos] [US2] La vertical de ensayo crece (FR-019; data-model §7; research D11): nuevo `internal/skills/testdata/vertical-de-ensayo/data/plazos.yaml` con `tipos:` y el tipo `ensayo` (cómputo `lpac`, 10 días hábiles, `da: vencimiento`, precepto `BOE-A-1900-1` `a2` `"2"` `fija: duracion`); una fila más de `regulados_aparte` en `internal/skills/testdata/vertical-de-ensayo/data/cobertura.yaml` con su precepto en la DA 1ª.2.c de la LPAC y `norma` `BOE-A-1900-2`, y su cabecera al día; y la sección y las marcas vacías de `recursos_y_plazos` en `internal/skills/testdata/vertical-de-ensayo/skills/plazos-y-territorio/references/ensayo.md`. `instalarLaVertical` todavía no lee los tipos, así que con el molde de T003 la región nueva sale vacía. Con la regla de hoy (la fila es del paquete de la norma de su precepto) la fila nueva sale en la `administrativo.md` de la copia de `TestVerticalDeEnsayo`, que pasa de 100 a 101 líneas sin índice y la pone en rojo (intento 1, `gates/tarea-T005.md`): por eso la tarea trae también, en `internal/skills/regiones.go`, `paqueteDeLaFila(entrada, normas)` (contracts/sincronia-de-los-paquetes.md §3): el `vertical` de la `norma` de la fila si está en la tabla de normas con `vertical` y si no el de la norma de su precepto, y la usan `filasDeLaCobertura` y `filasSinRegion` (defecto con el texto de hoy si su paquete es vacío o no está entre los cubiertos); la línea de cada fila no cambia; el test primero, en `internal/skills/regiones_test.go`, y `TestVerticalDeEnsayo`, hoy en rojo por esta causa, en verde sin tocarlo. `make skills-check` sin diferencias y `make ci` en verde.
- [ ] T006 [US2] El paquete de una fila y `TestVerticalDeEnsayo` crecida (FR-002, FR-005, FR-006, FR-007, FR-016, FR-020; US1.4, US2.1-US2.4; SC-002; contracts/sincronia-de-los-paquetes.md §3, §5; research D2, D11): en `internal/skills/regiones.go`, `paqueteDeLaFila(entrada, normas)` da el `vertical` de la `norma` de la fila si está en la tabla de normas del repositorio con `vertical` y si no el de la norma de su precepto, y la usan `filasDeLaCobertura` y `filasSinRegion` (defecto con el texto de hoy si su paquete es vacío o no tiene la región); la línea de cada fila no cambia. En `internal/skills/vertical_de_ensayo_test.go`, `instalarLaVertical` añade también las entradas de `tipos` del `plazos.yaml` de la vertical de T005, y el test exige: la fila del tipo `ensayo` en `recursos_y_plazos` de `plazos-y-territorio/references/ensayo.md` y en ninguna otra referencia de la copia; la fila con norma de ensayo en `ensayo.md` y en ninguna línea de `administrativo.md`; un subtest `sin-la-region` en el que, con la `ensayo.md` de `plazos-y-territorio` sin la región de la cobertura, `skills.Regenerar` da el defecto de FR-006 para esa fila; tras sincronizar, con una línea de 69 caracteres y su salto añadida al final del cuerpo de `jurisprudencia`, `normativa` y `plazos-y-territorio`, `skills.Regenerar` sin defectos (sin exportar nada); y la fila de lo cubierto de la vertical como última de la región y detrás de la que era la última en la copia antes de instalar, en lugar de «detrás de la del territorio». Ver en rojo los cuatro mutantes de US2.4 (un tipo asignado a todas las referencias, que es también la regla de hoy de US1.4; una fila asignada por su precepto; `filasSinRegion` con la regla vieja; el cuerpo de `normativa` sin el movimiento de T004); `make skills-sync` sin diferencias en las referencias de las skills (`make skills-check` en verde) y `make ci` en verde.

**Checkpoint**: el molde admite un paquete con sus plazos y su norma sin Go (aceptación 1 del hito).

---

## Phase 6: User Story 7 — las grabaciones de un hito entran sin Go (Priority: P2)

**Goal**: `UnionDeGrabaciones` suma sola las carpetas de grabaciones de cada hito (FR-015, §C.4).

**Independent Test**: `go test -run '^(TestGrabacionesSinSolape|TestManifiestoDeGrabaciones)$' ./internal/evals/`.

- [ ] T007 [P] [US7] Las raíces de grabaciones de los hitos (FR-015; US7.1-US7.3; SC-005; contracts/evals.md §1; research D5): en `internal/evals/grabaciones.go`, `UnionDeGrabaciones() ([]string, error)` sobre `unionDeGrabacionesDe(raiz)`, que da `GrabacionesDeH4`, `GrabacionesDeH5`, `GrabacionesDeH9` y detrás `boe.legislacion-consolidada` de cada carpeta de un hito del directorio `evidencias` del repositorio con `manifiesto.json` y esa carpeta, por número de hito y después de subhito, en una lista nueva en cada llamada y con el error de `os.ReadDir` envuelto con su ruta; `carpetaDeUnHito(nombre)` con `strings.CutPrefix`, `strings.Cut` y `strconv.Atoi`, sin `regexp` (`h36` → 36, 0; `h9-1` → 9, 1; `adr-*` y `sondeos-*` no); `GrabacionesDeH36` desaparece. En `internal/evals/dependencias.go`, el predicado puro `esDeUnaRaizDeGrabaciones(ruta)` sustituye a la enumeración de raíces en `rutaQueMide`; `internal/evals/medida.go`, `internal/evals/sesiones.go` y `internal/evals/dependencias.go` propagan el error. Los tests que llaman a la unión o a la constante lo reciben: `internal/evals/conjunto_test.go`, `internal/evals/dependencias_test.go`, `internal/evals/grabacion_test.go`, `internal/evals/juzgar_test.go`, `internal/evals/laboratorio_test.go`, `internal/evals/preparar_test.go` e `internal/skills/normas_test.go`. En `internal/evals/grabaciones_test.go`, `TestGrabacionesSinSolape` exige sobre el repositorio las tres fijas delante, la de h36, ninguna de h9, 3 + tantas como carpetas de un hito con manifiesto y carpeta del BOE, y que dos raíces solo compartan `robots.txt`; y `TestManifiestoDeGrabaciones/raices-de-los-hitos` gana `con-h99` y `orden-y-carpetas-que-no-entran` sobre árboles en `t.TempDir()`, como la tabla de contracts/evals.md §1. Ver en rojo el orden alfabético, `h9` dentro y un fichero grabado copiado entre dos raíces (en una copia en `$TMPDIR`); `make ci` en verde.

---

## Phase 7: User Story 8 — el trabajo de evals de `plazos-y-territorio` admite las evals de H42 (Priority: P2)

**Goal**: `plazos-y-territorio` a tres sesiones a la vez, doce en total, y el peor caso derivado de `evals/` (FR-021,
FR-022).

**Independent Test**: `go test -run '^(TestDefinicionDelJob|TestPlanEnDosModos)$' ./internal/evals/`.

- [ ] T008 [P] [US8] El tope del trabajo de evals (FR-021, FR-022, FR-016, FR-024; US8.1-US8.3; SC-006; contracts/evals.md §2; research D9): en `.github/workflows/evals.yml`, `concurrencia: 3` para `plazos-y-territorio` y su comentario con la cuenta a tres y «doce sesiones entre los cuatro», sin tocar el tope de 352 min, los modelos ni los demás ajustes; en `internal/evals/definicion_test.go`, `sesionesALaVezComoMucho = 12` con su comentario, `ajustesDelTrabajo["plazos-y-territorio"]` con `Concurrencia: 3` y la definición sintética del contrato con la del repositorio, `probarElPeorCasoDelTrabajo` con lo esperado derivado de `LeerConjunto` sobre el directorio de evals de cada skill (series por modo con el modelo que decide y los informativos, por repeticiones, más la sesión de la prueba de red en `orden`, y la concurrencia de la definición leída) y cada trabajo dentro de su tope, sin `[]int{67, 66, 6}`, `20725 * time.Second` ni sus textos; un subtest `cuatro-evals-mas` con el directorio de evals del repositorio copiado a `t.TempDir()` y cuatro copias de más de una eval de modo de `plazos-y-territorio`, que exige que su peor caso quepa en 352 min con la definición del repositorio y no quepa con `plazos-y-territorio` a dos; y `tope-con-las-evals-del-repositorio` con sus líneas calculadas desde el peor caso de la definición que monta; en `internal/evals/plan_test.go`, `probarElPlanDelRepositorio` con las series y sesiones por modo derivadas de `LeerConjunto`. En `CHANGELOG.md`, *Unreleased* «Cambiado», la concurrencia nueva del trabajo de evals de `plazos-y-territorio`. Ver en rojo `plazos-y-territorio` a dos y una sesión más a la vez (13); comprobar en la copia con un paquete más que los dos tests siguen en verde; `make ci` en verde.

---

## Phase 8: User Story 6 — un paquete entra sin tocar el laboratorio de H41 (Priority: P2)

**Goal**: `TestSkillsDelProductoDesdeControl` no pide declarar la `references/<vertical>.md` que su base no tiene
(FR-014).

**Independent Test**: `go test -run '^TestSkillsDelProductoDesdeControl$' ./internal/evals/`.

- [ ] T009 [P] [US6] El laboratorio sin las referencias de un paquete (FR-014, FR-016, FR-008; US6.1-US6.3; SC-001, SC-009; contracts/skills-y-laboratorio.md §3; research D7): en `internal/evals/laboratorio_familias_test.go`, la comparación pasa a `diferenciasDeLosFicheros`, que devuelve las diferencias y que `compararLosFicheros` pasa a `assert`, y no da «sobra» para la referencia `<v>.md` de una skill del producto si `<v>` es un `vertical` de la tabla de normas del repositorio (leída con `skills.LeerNormas` de `tablaDeNormasDelRepositorio`) y la base (la variante `desde` o el control) no tiene ese fichero; si la base lo tiene, se compara como hoy. Subtest nuevo `referencias-de-un-paquete` sobre los ficheros leídos del producto con un fichero añadido o cambiado en memoria: la `administrativo.md` de las referencias de `normativa` → 0 diferencias; la misma como `materia-inventada.md` → 1, «sobra»; un byte cambiado fuera de las regiones de la `administrativo.md` de las referencias de `plazos-y-territorio` → 1. Comprobar en la copia con un paquete más que `TestSkillsDelProductoDesdeControl` sigue en verde; ver en rojo los dos mutantes de US6.2 y US6.3 y un byte no declarado en una skill del producto (SC-001); `make ci` en verde.

---

## Phase 9: User Story 4 — un paquete más no rompe ningún test ajeno (Priority: P1)

**Goal**: los tests de la lista de §C.5 (contracts/tests-sin-listas.md) derivan lo esperado de `data/` y `evals/` o
usan datos de prueba, y cada uno sigue en rojo con su mutante de `main` (FR-016, FR-017). `TestRegiones` (T003),
`TestVerticalDeEnsayo` (T006), `TestDefinicionDelJob` y `TestPlanEnDosModos` (T008), `TestSkillsDelProductoDesdeControl`
(T009) y las descripciones (T010, T011) ya están; aquí, el resto. «Del paquete» es un filtro escrito en el test (las
normas con `vertical: administrativo`), no la función de producción.

**Independent Test**: la copia con un paquete más corre las dos órdenes de `paquete-de-mas.txtar` con código 0.

- [ ] T010 [P] [US4] Los tests del paquete Go de las skills sin listas (FR-016; US4.1-US4.3; SC-004; contracts/tests-sin-listas.md, la tabla de la sección del paquete de las skills, la primera): en `internal/skills/cobertura_test.go`, `TestLeerCobertura/repositorio` compara las filas de la base (sin `materia` o de `administrativo`) en su orden y exige que las demás tengan por `materia` un `vertical` de la tabla de normas del repositorio, y `TestCoberturaDelRepositorio` fija sus 9 tuplas sobre las filas del paquete (por su norma si la lleva con `vertical`, o por su precepto); en `internal/skills/cubierto_test.go`, `TestRegionDeLoCubierto/repositorio` espera una línea por fila de `cubierto` con la forma de `lineasDeLoCubierto` y su copia lleva cada skill que tiene una referencia de un paquete, y `/sin-data-cobertura` espera un defecto por `references/*.md` de la copia con la marca de `cobertura_por_materia`, en orden de nombre, tras el de la plantilla; en `internal/skills/referencias_test.go`, `TestRenderizarRecursosYPlazos/desde-data` pinta los tipos del paquete y espera `2 +` una fila por tipo y una por unidad de cada tipo con unidades, y `TestRegionDeLaCobertura` espera la fila de cada fila del paquete en `administrativo.md` con la línea de la plantilla, `2 + n` y `n` con `n` las del paquete y una deriva por referencia con la marca. Ver en rojo cada test con su mutante de la columna «Rojo con» y, en la copia con un paquete más, todos en verde; `make ci` en verde.
- [ ] T011 [P] [US4] Los subtests de `TestEvalsDelRepositorio` sin listas (FR-016, FR-018; US4.1-US4.3; SC-004; contracts/tests-sin-listas.md, la tabla de la sección del paquete de las evals, la tercera): en `internal/evals/conjunto_test.go`, `precepto-de-la-cobertura` (`probarLosPreceptosDeLaCobertura`) toma las evals con `no_cubierto` de `LeerConjunto`, exige al menos una y que el precepto de cada una sea el de una fila de la cobertura del repositorio (`preceptosFueraDeLaCobertura`), y `conjunto-plazos-y-territorio` espera tantas evals como `*.yaml` del directorio y ninguna mal formada, sin «las doce». Las reglas del conjunto de `normativa`, en el código de producción del conjunto de evals, no se tocan. Ver en rojo su mutante de hoy (la 12 con el bloque `a999`) y `LeerConjunto` que pierde un fichero; en la copia con un paquete más, en verde; `make ci` en verde.
- [ ] T012 [P] [US4] Los umbrales y el informe sin listas (FR-016; US4.1-US4.3; SC-004; contracts/tests-sin-listas.md, la tabla de la sección del paquete de las evals, la tercera): en `internal/evals/informe_test.go`, `TestUmbralesDePlazosYTerritorio` deriva totales, nombres de eval y modelos de `LeerConjunto` sobre las evals de `plazos-y-territorio` por 3 repeticiones con los predicados de producción de cada clase (`esDePlazo`, `esDeCobertura`, `esDeVigencia`, `esperaUnaRedaccionFechada`, `comandosSinServir`) y los modelos por eval de modo; `TestUmbralesDeNormativa`, de `LeerConjunto` sobre las evals de `normativa` y el `ObjetivoDeDuracion` de la definición leída; `TestUmbralesDeLasSkills`, del conjunto de cada skill; y `TestInformeEnDosModos/plazos-y-territorio`, de las evals de `LeerConjunto` sin `SinBinarioNiServidor`, en su orden. Ver en rojo los mutantes de la tabla (`if r.conVigencia` quitado; `lectura_sin_intentar` que decide para todo modelo; `esperaUnaRedaccionFechada` cambiada; el umbral de duración fuera; `umbralesDelInforme` en otro orden; `fuera_de_lo_grabado` que decide; una serie en otro orden) y, en la copia con un paquete más, todos en verde; `make ci` en verde.
- [ ] T013 [P] [US4] La matriz de las propuestas sin lista (FR-016; US4.1-US4.3; SC-004; contracts/tests-sin-listas.md, fila `TestMatrizDeLasPropuestas`): en `internal/evals/eleccion_test.go`, la parte de los ficheros de `goldenDeLaMatriz` sale de `repositorio.evals[skill]` y la forma de cada sección se queda literal. Ver en rojo el formato de `matrizDeLaEjecucion` cambiado y la regla de elección cambiada, y en la copia con un paquete más, en verde; `make ci` en verde.
- [ ] T014 [P] [US4] El caso de los tributos de `TestCasosDePlazos` sin el texto de su fila (FR-016, FR-026; US1.3, US4.1-US4.3; SC-004): en `internal/app/plazos_casos_test.go`, `19-eval-c-tributos` busca la única fila con precepto DA 1ª.2.a de la LPAC y `norma` la LGT en lugar de su procedimiento, y los demás casos no cambian. Ver en rojo la fila sin `norma`, con otro `aplicable` u otro precepto, o repetida; en la copia con un paquete más, en verde; `make ci` en verde.
- [ ] T015 [datos] [US4] Los guiones de la app sin listas, con sus copias congeladas (FR-016, FR-017, FR-026; US1.3, US4.1-US4.3; SC-004; contracts/tests-sin-listas.md, la tabla de la sección de la app, la segunda): `internal/app/testdata/script/h32-paquetes.txtar` y `specs/030-h32-arquitectura-de-skills/aceptacion/paquetes.txtar` buscan las filas del generador y la de tributos dentro de alguna referencia de un paquete de `plazos-y-territorio` (`references/*.md` con una región); `internal/app/testdata/script/h33-cobertura-instalada.txtar` y `specs/039-h33-cobertura-en-positivo/aceptacion/cobertura-instalada.txtar` exigen las 4 líneas de la base al principio de la región de lo cubierto y, en todas las referencias de paquetes instaladas, tantas filas como `- procedimiento:` y con norma tantas como `norma:` del `cobertura.yaml` del directorio de datos que está junto a `$KITLEGAL_SKILLS`, como dice el contrato (que el guion solo lee); `internal/app/testdata/script/h36-plazos-de-hoy.txtar` y `specs/041-h36-el-motor-de/aceptacion/plazos-de-hoy.txtar` cambian `error-tipo` a código 2, `clase` `argumentos` y el mensaje con su forma (`el tipo \"inventado\" no es ` seguido de códigos separados por «, » y « ni » el último), con `stdout` y expresiones, sin el `hash` literal. Cada guion igual a su copia congelada. Ver en rojo cada guion con su mutante de la columna «Rojo con» (en una copia en `$TMPDIR`) y, en la copia con un paquete más, `h32-paquetes`, `h33-cobertura-instalada` y `h36-plazos-de-hoy` en verde; `make ci` en verde.

**Checkpoint**: con un paquete más, ningún test de la lista cae (aceptación 3 del hito), salvo los de la descripción, que cambian en la fase siguiente.

---

## Phase 10: User Story 9 — la descripción de `plazos-y-territorio` nombra su capacidad (Priority: P2)

**Goal**: el texto de FR-023 en el `SKILL.md` y en la variante, con las sumas de los tests derivadas (FR-023, FR-024).

**Independent Test**: `go test -run '^(TestDescripcionesDelProducto|TestDescripcionesCortas|TestSkillsDelProductoDesdeControl)$' ./internal/evals/` y `TestEntregaDelHito/h41-4-skills-instaladas`.

- [ ] T016 [US9] Las sumas de los tests de descripciones derivadas de las descripciones de hoy (FR-023, FR-016; SC-007; contracts/skills-y-laboratorio.md §4; research D8): en `internal/evals/descripciones_del_producto_test.go`, «una de 451» y «una quinta skill sin texto» suman lo que añaden (25 y 20) a la suma de las cuatro descripciones leídas al empezar el test, y «una suma de 1701» añade a `plazos-y-territorio` `1701 − suma` letras; en `internal/evals/laboratorio_descripciones_test.go`, «una que empieza por <ejemplo>» espera el defecto de la suma solo si la suma de hoy más 10 pasa de 1 700. Con la descripción de hoy, todo sigue en verde. Ver en rojo una descripción por encima de 450 y la suma por encima de 1 700; `make ci` en verde.
- [ ] T017 [datos] [US9] La descripción nueva de `plazos-y-territorio`, en una sola tarea con el guion que la fija porque en cuanto cambia el binario empotra otra y `h41-4-skills-instaladas` la ve (FR-023, FR-024; US9.1, US9.2; SC-007; contracts/skills-y-laboratorio.md §4; research D8): la `description` del frontmatter de `skills/plazos-y-territorio/SKILL.md` y `laboratorio/producto-descripciones-cortas/plazos-y-territorio.txt` pasan al texto de FR-023 byte a byte (429 caracteres, `wc -m` sin el salto), sin tocar nada más de la skill; `descripcionCortaDePlazosYTerritorio` de `internal/evals/laboratorio_descripciones_test.go` la sigue; la línea 34 de `internal/app/testdata/script/h41-4-skills-instaladas.txtar` y de su copia congelada `specs/038-h41-4-las-skills-del/aceptacion/skills-instaladas.txtar` la exige con `( |\n  )` entre palabras; si la declaración de `plazos-y-territorio` de `internal/evals/laboratorio_familias_test.go` fija la descripción, la sigue; y `CHANGELOG.md`, *Unreleased* «Cambiado», dice la descripción nueva, que deja de nombrar una materia. Comprobar con `wc -m` los 429 caracteres y la suma de la variante (1 692 ≤ 1 700); ver en rojo, con la descripción de hoy en el `SKILL.md` o en la variante, `TestSkillsDelProductoDesdeControl`, `TestDescripcionesCortas` y `h41-4-skills-instaladas`; `make ci` en verde.

**Checkpoint**: las dos skills cambian solo en lo declarado (aceptación 2 del hito); `skills-instaladas.txtar` de T001 ya pasaría.

---

## Phase 11: Polish — el registro de los controles

- [ ] T018 Escribir `specs/042-h45-el-molde-admite/mutantes.md` (FR-027; FR-008, FR-018, FR-025, FR-026; SC-001 a SC-009, SC-011; plan, tabla de mutantes; research D13): una entrada por control de «Criterios del hito (literales)» del spec, con la orden exacta, lo esperado y lo visto, cada mutante aplicado en una copia del módulo en `$TMPDIR` y nunca en el árbol de trabajo, y las comparaciones con `main` hechas con `git show main:<ruta>` o `git diff main -- <rutas>`, que solo leen. Incluye: lo de hoy igual (`make skills-sync` sin diferencias; el diff de las skills frente a `main` solo en las cuatro líneas movidas de `normativa`, su `con_herramientas.md` y la `description` de `plazos-y-territorio`; la regla de hoy con un tipo de ensayo en rojo en `TestVerticalDeEnsayo`; `TestCasosDePlazos`, `TestAppletPlazos` y los guiones `h9-plazos-*` y `h36-plazos-de-hoy` en verde); el molde (los cuatro mutantes de US2.4); los defectos (la regla quitada; el cómputo contado); los tests sin listas (con los dos guiones de la suite de aceptación del hito, que no se tocan, activados en una copia, `h45-paquete-de-mas` en verde, y en rojo con la lista literal de hoy devuelta a `TestCoberturaDelRepositorio`, `tests enteros: 1`; y el mutante de cada test de la lista); las grabaciones; el tope del trabajo; la descripción; PowerShell (los cinco de §C.3); el laboratorio (sus dos mutantes); el diff de las evals frente a `main`, vacío (FR-025), y las reglas del conjunto de `normativa`, en el código de producción del conjunto de evals, sin cambiar frente a `main` (FR-018). Recorrer los escenarios del quickstart del hito y apuntar en el registro lo que se haya podido medir y lo que no; `make ci` en verde.

---

## Controles de umbral → tarea que construye el control

Cada fila de «## Controles de umbral» de plan.md, con la tarea que construye o adapta su control y lo ve en rojo:

| Requisito | Control | Tarea |
|---|---|---|
| SC-001, FR-008 | `TestSkillsDelRepositorio`, `TestRegiones/sincronia`, `TestSkillsDelProductoDesdeControl` | T003 (región distinta de la generada), T009 (byte no declarado), T018 (contra `main`) |
| SC-002, FR-002, FR-005, FR-006, FR-020 | `TestVerticalDeEnsayo` | T006 (los cuatro mutantes de US2.4) |
| SC-003, FR-003, FR-004 | `TestRegiones/defectos` | T003 |
| SC-004, FR-016, FR-017 | `h45-paquete-de-mas` en `TestEntregaDelHito` | T001 (el guion), T010-T015 (cada test, con su mutante), T018 (la lista literal devuelta) |
| SC-005, FR-015 | `TestGrabacionesSinSolape`, `TestManifiestoDeGrabaciones` | T007 |
| SC-006, FR-021, FR-022 | `TestDefinicionDelJob`, `TestPlanEnDosModos` | T008 |
| SC-007, FR-023 | `TestDescripcionesCortas`, `TestDescripcionesDelProducto`, `TestSkillsDelProductoDesdeControl`, `h41-4-skills-instaladas`, `h45-skills-instaladas` | T016 (topes), T017 (la de hoy en rojo), T001 (`skills-instaladas.txtar`) |
| SC-007, SC-010 (juez de las descripciones) | job de evals del cierre, umbral 0 | control existente que no cambia; lo mide el cierre del workflow, no una tarea; T017 deja la descripción que juzga |
| SC-008, FR-011, FR-012 | `ordenes-para-powershell`, `TestControlesSobreLasReferenciasAMano` | T004 |
| SC-009, FR-014 | `TestSkillsDelProductoDesdeControl/referencias-de-un-paquete` | T009 |
| FR-018 | `TestEvalsDelRepositorio` (conjunto de `normativa`) | control existente sin cambiar; T011 no toca `conjunto.go`; T018 lo registra contra `main` |
| FR-026 | `TestCasosDePlazos`, `TestAppletPlazos`, `h9-plazos-*`, `h36-plazos-de-hoy` | T014, T015, T018 |
| SC-010, FR-025 (las cuatro skills) | job de evals del cierre, umbrales con `decide: true` | controles existentes que no cambian (ninguna eval cambia); los mide el cierre del workflow; T008 hace que el trabajo de `plazos-y-territorio` quepa en su tope |
| SC-011 | `make ci` | todas las tareas |

## Definition of Done (docs/ROADMAP.md §1) que aplica

- `make ci` en verde: cada tarea; los guiones de T001, activados por el workflow al final.
- e2e y `CHANGELOG.md` (comportamiento visible: la skill instalada): T001, T015, T017; `CHANGELOG.md` en T008 y T017
  (FR-024).
- ADR: no aplica (el hito aplica el ADR 0062 y no escribe otro: spec, «Fuera de alcance»).
- `docs/SOURCES.md`: no aplica (ninguna fuente nueva ni petición HTTP; plan, «Datos externos»).
- `schemas/`: no aplica (ningún sobre ni esquema cambia; FR-026).
- Fixtures: T002, T005, T015, T017 (`[datos]`); ninguna grabación.
- Skills (§1.10): cuerpos ≤ 11 000 (T004, T006), descripciones ≤ 900 y variante ≤ 450 / 1 700 (T016, T017),
  referencias regeneradas sin deriva (T003, T006); ninguna eval cambia (FR-025).

---

## Dependencies & Execution Order

- **T001** primero (la suite congelada). **T002** antes de T003.
- **T003** (molde de los tipos) antes de T005 y T006.
- **T004** (movimiento de `normativa`) antes de T006 (el sitio de FR-020) y de T009 (misma declaración).
- **T005** (datos de la vertical) antes de T006.
- **T007**, **T008**, **T009**: independientes entre sí tras T004 (T007 y T011 tocan `conjunto_test.go`; T011 va después).
- **T010-T015**: tras T003-T009 (sus derivaciones se comprueban con el molde entero en la copia con un paquete más);
  [P] entre sí, ficheros distintos.
- **T016** antes de T017 (las sumas derivadas antes de cambiar la descripción).
- **T018** al final.

Historias: US3 (T003) → US5 (T004) → US2 (T005, T006) → US7, US8, US6 (T007, T008, T009) → US4 (T010-T015) → US9
(T016, T017). US1 no tiene tarea propia: sus escenarios los comprueban T003 (US1.1), T017 y T018 (US1.2), T014, T015 y
T018 (US1.3) y T006 (US1.4).

## Parallel Example

```text
# Tras T006, en la fase 6-8:
T007 internal/evals/grabaciones.go …
T008 .github/workflows/evals.yml, internal/evals/definicion_test.go, internal/evals/plan_test.go
T009 internal/evals/laboratorio_familias_test.go

# Fase 9:
T010 internal/skills/{cobertura,cubierto,referencias}_test.go
T012 internal/evals/informe_test.go
T013 internal/evals/eleccion_test.go
T014 internal/app/plazos_casos_test.go
```

El bucle del workflow ejecuta las tareas una a una en el orden de sus ids; las [P] solo dicen que no dependen entre sí.

## Implementation Strategy

- **MVP**: T001-T006 — el molde con la vertical de ensayo (US3, US5, US2): lo que H42 necesita para entrar con datos.
- Después, lo que deja entrar a H42 sin Go: grabaciones (T007), trabajo de evals (T008), laboratorio (T009) y los tests
  sin listas (T010-T015).
- Al final, la descripción (T016, T017) y el registro de los controles (T018).
