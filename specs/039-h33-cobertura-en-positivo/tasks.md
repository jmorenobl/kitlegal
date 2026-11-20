---

description: "Tareas de H33 · Cobertura en positivo por capacidad y materia"
---

# Tasks: H33 · Cobertura en positivo por capacidad y materia

**Input**: `specs/039-h33-cobertura-en-positivo/` — [spec.md](spec.md), [plan.md](plan.md), [research.md](research.md),
[data-model.md](data-model.md), [quickstart.md](quickstart.md) y [contracts/](contracts/) (`datos-y-regiones.md`,
`skills.md`, `evals.md`, `controles.md`, `aceptacion.md`).

**Tests**: obligatorios (constitución III). La primera tarea escribe la suite de aceptación, que el workflow congela;
cada tarea posterior trae su test y el cambio que lo pone en verde, y deja `make ci` en verde por sí sola.

**Organización**: las tareas van en el orden en que se ejecutan una tras otra, el de plan.md «Orden de implementación»
(de dentro afuera), y cada fase lleva la historia a la que sirve sobre todo. La línea de cada tarea declara todas las
rutas que crea o modifica: el guardián rechaza cualquier otra; `go.mod`, `go.sum`, `CHANGELOG.md` y el directorio del
feature siempre se admiten, y declarar `x.go` admite `x_test.go`.

## Format: `[ID] [P?] [Story] Descripción`

- **[P]**: ninguna tarea lo lleva. Cada una parte de la anterior: T003 deja el tipo que leen T004 y T005; T004 y T005
  regeneran el mismo `SKILL.md` de `plazos-y-territorio` y tocan `internal/skills/regiones.go` y
  `internal/evals/laboratorio_familias_test.go`; T006 deja el formato que necesitan las evals de T007, y T006 y T007
  tocan `internal/evals/informe_test.go` y `internal/evals/conjunto.go`.
- **[Story]**: la historia de [spec.md](spec.md) a la que sirve: US1 (la línea y sus evals), US2 (lo cubierto en datos y
  su región), US3 (lo regulado aparte con la plantilla), US4 (las evals miden la frontera sin modelo), US5 (lo instalado).
  T002 (lo grabado) es fundacional y no lleva historia; T008 y T009 son de pulido.
- **[aceptacion]**: la suite congelada, solo T001.
- **[datos]**: T002 a T006, cada una porque crea o modifica ficheros bajo `testdata/` o `schemas/`. T002 solo toca el
  manifiesto. T003 a T006 juntan los datos con el código y el test que los leen porque, por separado, `make ci` queda en
  rojo; cada línea dice por qué (plan.md «Orden de implementación», pasos 3 a 6).

## Convenciones de todas las tareas

- `make ci` en verde al terminar cada una, en primer plano, y si hace falta por objetivos (`make fmt-check`, `make lint`,
  `make test`, `make test-integration`, `make vuln`, `make schema-check`, `make skills-check`). Ninguna usa la red más que
  la de las herramientas de Go; ninguna graba ni pide a una fuente: lo único que se graba, el art. 59 del ET, lo graba el
  paso `grabar_datos` del workflow, sin modelo, tras el commit de T002 (plan.md «Datos externos»). Ninguna escribe fuera
  del repositorio y de `$TMPDIR`; ninguna abre una sesión con modelo ni ejecuta `make evals`, `make evals-sondeo`,
  `make evals-laboratorio`, `TestSondeo*` ni los guiones `scripts/evals*.sh` (FR-026; ADR 0032).
- `data/cobertura.yaml` es el único fichero de `data/` que cambia, en T003, con los textos literales de FR-001: no son
  datos de una fuente externa sino la declaración de lo que kitlegal cubre, que el hito fija palabra por palabra, y no
  hay nada grabado de lo que salgan (research D12). Ningún otro fichero de `data/` se toca, y ninguno sale de lo grabado.
- No se tocan `.golangci.yml`, `Makefile`, `.github/workflows/ci.yml`, `scripts/`, `cmd/`, `internal/core/`,
  `laboratorio/`, `skills/normativa/`, `skills/jurisprudencia/`, `skills/plazos-y-territorio/references/calculo_de_plazos.md`,
  el frontmatter ni la `description` de ninguna skill (contracts/skills.md; contracts/controles.md §6). Ningún
  `//nolint`, `t.Skip`, `TODO` ni test desactivado. De `.github/workflows/evals.yml` solo cambia un comentario, en T007.
- Las suites congeladas de los hitos anteriores (`specs/0*/aceptacion/` y sus copias activadas en
  `internal/app/testdata/script/`) siguen en verde sin cambiar un fichero suyo, salvo los cuatro cambios de FR-020: los
  dos guiones de H32 de `internal/app/testdata/script/` y la declaración de `plazos-y-territorio` en T004, y la huella de
  `consulta-juridica` en T005. Las copias de `specs/030-h32-arquitectura-de-skills/aceptacion/` no cambian.
- Los textos son los de los contratos: las skills, la referencia y la documentación, los de
  [contracts/skills.md](contracts/skills.md); los datos, el esquema, la región de lo cubierto, el generador y sus
  defectos, los de [contracts/datos-y-regiones.md](contracts/datos-y-regiones.md); el formato, el juicio, los umbrales,
  la regla, las evals, la grabación y el job, los de [contracts/evals.md](contracts/evals.md); los tests y sus mutantes,
  los de [contracts/controles.md](contracts/controles.md); el guion, el de [contracts/aceptacion.md](contracts/aceptacion.md).
  Nada se inventa fuera de ellos, y ningún texto de la línea `⚠ NO CUBIERTO:` vive en Go (FR-006, FR-013).
- Arquitectura (plan.md «Constitution Check»): solo cambian `internal/skills` e `internal/evals`; tipos y funciones no
  exportados salvo `FilaCubierta` y los campos nuevos de `Cobertura`, `Eval` y `NoCubiertoEsperado`. Ningún import,
  dependencia, código de salida, error de producción ni `panic` nuevo; nada nuevo escribe en stdout ni llama a `os.Exit`.
  Un dato ilegible es un defecto de la sincronía; una entrada manipulada, el defecto `inesperado` (ADR 0023). Tests de
  tabla con `t.Parallel()` y `testify`, como los de cada paquete; los mutantes, sobre copias en `t.TempDir()`.
- Cada control se ve fallar con un mutante —la medida pasada de su umbral— antes de darlo por bueno (ADR 0029): los de
  contracts/controles.md son subtests permanentes; los que se ven a mano se aplican sobre una copia bajo `$TMPDIR`, nunca
  sobre el árbol versionado, y no se versionan.

---

## Phase 1: Setup — la suite de aceptación (US5, prioridad P2, escrita primero)

**Goal**: fijar, antes de cualquier cambio, lo que `kitlegal skills install` tiene que instalar de la cobertura en
positivo (ADR 0018; plan.md «Aceptación e2e»).

**Independent Test**: el workflow ve el guion en rojo por su primera aserción (rojo primero) y lo congela; al terminar el
bucle lo activa como `h33-cobertura-instalada.txtar` y lo ejecuta `TestEntregaDelHito` en `make ci`.

- [X] T001 [aceptacion] [US5] Escribir, a partir del spec y de contracts/aceptacion.md §2, sin código de producto, el guion testscript specs/039-h33-cobertura-en-positivo/aceptacion/cobertura-instalada.txtar, byte a byte el bloque `text` de §2 con su salto de línea final: la cabecera de comentarios con la línea `# Cubre: US5 · FR-004, FR-007, FR-010, FR-025 · SC-001, SC-002, SC-003, SC-005`; `exec kitlegal skills install --dir destino --json`; los `grep` literales de la plantilla sola en su bloque en `plazos-y-territorio` (con `-count=1` y el `! grep` de la marca), de la plantilla seguida de la marca en `consulta-juridica` (con `-count=1`), los dos `! grep` de «procedimiento común» en la línea de la plantilla, el de la región de lo cubierto con sus dos marcas y sus cuatro líneas, el `exec awk -f filas.awk` con `stdout '^filas: 9 de 9, 2 con norma$'`, el de la fila de los tributos y el `! grep` del tope de 10 860 caracteres del cuerpo (diez `(?s:.){1000}` y un `(?s:.){861}`); y la sección `-- filas.awk --` tal cual. Cubre la historia 5 entera (FR-025) y lo instalado de FR-004, FR-007 y FR-010 (SC-001, SC-002, SC-003, SC-005); afirma solo lo que el spec fija byte a byte (contracts/aceptacion.md §1). Sin evals en esta tarea: las de la frontera (01, 03 y 04 de `consulta-juridica`, 17 de `normativa`) llevan `fuera_de_lo_declarado` y `tambien_la_resuelven`, claves que el esquema de eval de hoy rechaza, y dejarían `make ci` en rojo; las escribe T007, tras el formato de T006 (research D18). Hoy falla por su primera aserción, el `grep` de la plantilla de `plazos-y-territorio` (research V7b); lo comprueba el workflow (rojo primero), y la tarea no deja ningún otro fichero cambiado.

**Checkpoint**: la suite está escrita y congelada; nada del producto ha cambiado.

---

## Phase 2: Foundational — lo grabado

**Purpose**: el art. 59 del ET, que lee la eval 04 de T007 (FR-018), grabado por el paso `grabar_datos` antes de que lo
use ninguna tarea. La entrada sin grabar no pone nada en rojo (research V6).

- [X] T002 [datos] La entrada del art. 59 del Estatuto de los Trabajadores en el manifiesto de grabación de H9, sin código ni test nuevos (FR-018; contracts/evals.md §6; research D11, V5, V6, V14; plan.md «Datos externos»). **Manifiesto**: internal/evals/testdata/grabaciones.json (`"fuente": "boe.legislacion-consolidada"`, con fila «Revisado» 2026-09-13 en `docs/SOURCES.md` de `main`), al final de su lista `normas`, la entrada de contracts/evals.md §6 byte a byte: `{"busqueda": "estatuto de los trabajadores", "titulo_empieza_por": "Real Decreto Legislativo 2/2015,", "bloques": ["a59"], "para": "…"}` con el `para` de §6; nada más del fichero cambia. **Test de grabación**: el de hoy, `TestGrabarEvals` (`//go:build grabacion`, internal/evals/grabacion_test.go), sin cambios: sirve primero desde lo grabado en H4 y H5 y solo pide a la fuente, por el cliente HTTP único del proyecto y con GET, el bloque `a59`. Tras el commit, el paso `grabar_datos` del workflow lo graba en la carpeta `boe.legislacion-consolidada` del `testdata` del paquete; la tarea no graba nada. **Verificación**, sin cambiar otro fichero: `make ci` en verde con la entrada sin grabar (`TestGrabacionesSinSolape` y la verificación de las normas de la cobertura siguen en verde, research V6); `grep -c '"titulo_empieza_por": "Real Decreto Legislativo 2/2015,", "bloques": \["a59"\]' internal/evals/testdata/grabaciones.json` da `1` (quickstart, escenario 7).

**Checkpoint**: tras el paso `grabar_datos`, el bloque `a59` del ET está grabado y las tareas siguientes parten de él.

---

## Phase 3: User Story 2 — lo cubierto se declara en datos (Priority: P1)

**Goal**: `data/cobertura.yaml` lleva el motivo de lo no declarado y las cuatro filas de lo cubierto, con su esquema y su
lector (FR-001, FR-002).

**Independent Test**: `TestLeerCobertura` (subtest `repositorio` y los seis mutantes del esquema).

- [X] T003 [datos] [US2] El motivo y lo cubierto en datos, con su esquema, su lector y su test, en una sola tarea porque el lector estricto y el esquema rechazan hoy las dos claves nuevas, `TestLeerCobertura` lee el fichero del repositorio, y sin el motivo obligatorio el fichero o los árboles sintéticos de `TestRegiones` dejan `make ci` en rojo (FR-001, FR-002; SC-008; contracts/datos-y-regiones.md §1 y §2; data-model §1; research D5, D12, V4, V15; plan.md «Orden de implementación», paso 3). **Esquema**: schemas/cobertura.yaml.json gana, sin reformatear nada de hoy, `motivo_de_lo_no_declarado` en `required` y en `properties` (`{"type": "string", "minLength": 1}`), `cubierto` en `properties` (`{"type": "array", "items": {"$ref": "#/$defs/cubierto"}}`) y, en `$defs`, `cubierto`: objeto con `additionalProperties: false`, `required: ["capacidad", "alcance"]`, `capacidad` con `enum` `["normativa", "jurisprudencia", "plazos", "territorio"]`, `alcance` con `minLength: 1` y `materia` `{"type": "string"}` sin patrón; ni `cubierto` obligatoria ni `minItems`; `regulados_aparte` no cambia un byte. **Lector**: internal/skills/cobertura.go: `Cobertura` gana `MotivoDeLoNoDeclarado string` (`yaml:"motivo_de_lo_no_declarado"`) y `Cubierto []FilaCubierta` (`yaml:"cubierto"`), y el tipo exportado `FilaCubierta{Capacidad, Alcance, Materia string}` con sus claves YAML; `LeerCobertura` no cambia su contrato. **Datos**: data/cobertura.yaml: la cabecera deja de decir que solo lleva lo regulado aparte y dice qué lleva cada parte y quién la lee; delante de `regulados_aparte`, `motivo_de_lo_no_declarado: fuera de lo que kitlegal cubre` y `cubierto` con las cuatro filas de FR-001, en su orden y con sus textos literales, cada `alcance` en una línea como escalar sin comillas (el de los plazos con su comilla invertida) y `materia: administrativo` solo en la de los plazos; las 9 filas de `regulados_aparte`, sin cambiar un byte. **Test**: internal/skills/cobertura_test.go, `TestLeerCobertura`: el subtest `repositorio` lee el fichero del repositorio sin defecto y exige el motivo y las cuatro filas, campo a campo, con los textos de FR-001 (en rojo con un carácter cambiado en uno de ellos); `coberturaDePrueba` gana las dos claves; los seis mutantes de contracts/datos-y-regiones.md §2 (fila sin `capacidad`, sin `alcance`, `capacidad: documentos`, fila con `normas`, sin `motivo_de_lo_no_declarado` y con él vacío), cada uno con su defecto exacto y su línea fijada; los casos de hoy de `regulados_aparte`, con sus mensajes de hoy. **Fixtures de otros tests**: internal/skills/regiones_test.go, solo su constante `coberturaDeLasRegiones` gana la línea `motivo_de_lo_no_declarado: fuera de lo que kitlegal cubre` delante de `regulados_aparte`; nada más del fichero cambia en esta tarea. **Verificación**: el `go test` de `TestLeerCobertura`, `TestRegiones`, `TestRegionDeLaCobertura` y `TestVerticalDeEnsayo` (quickstart, escenario 3, sin `TestRegionDeLoCubierto`, que llega en T005) y `make ci` en verde; `make skills-sync` no cambia ningún fichero (nada del motivo ni de lo cubierto llega todavía a una skill).

**Checkpoint**: lo cubierto está en datos y validado; ninguna skill ha cambiado.

---

## Phase 4: User Story 3 — lo regulado aparte usa la misma plantilla (Priority: P1)

**Goal**: la tabla `cobertura_por_materia` de `administrativo.md` sale de la plantilla única del `SKILL.md` de
`plazos-y-territorio` y de los dos motivos de su referencia (FR-004 a FR-006, FR-009).

**Independent Test**: `TestRegionDeLaCobertura` y `h32-paquetes.txtar`, con la fila de los tributos byte a byte.

- [X] T004 [datos] [US3] La plantilla y los motivos: el generador, la skill de `plazos-y-territorio` con su referencia regenerada, los motivos de la vertical de ensayo, los dos guiones de H32 y la declaración de su suite congelada, en una sola tarea porque, con el generador nuevo y la skill de hoy, la región de `administrativo.md` da su defecto; con la skill nueva y sin los guiones o sin la declaración, `TestEntregaDelHito/h32-*` y `TestSkillsDelProductoDesdeControl` salen en rojo; y sin los motivos de la vertical, `TestVerticalDeEnsayo` (FR-004, FR-005, FR-006, FR-009, FR-010, FR-011, FR-020, FR-021, FR-023; SC-002, SC-005, SC-008, SC-009; contracts/datos-y-regiones.md §5 y §6; contracts/skills.md §2 y §3; contracts/controles.md §1 a §3; data-model §2; research D2, D13, D17, D19; plan.md «Orden de implementación», paso 4). **Generador**: internal/skills/regiones.go: `formasDeNoCubierto` pasa a ser `lineaDeNoCubierto{plantilla, motivoConNorma, motivoSinNorma}`; `plantillaDelCuerpo` da la única línea de los bloques `text` del `SKILL.md` que, sin sus blancos del principio, empieza por `⚠ NO CUBIERTO: ` y lleva `<motivo>` (con otra cantidad, la skill no tiene plantilla); `lineaDeNoCubierto.conLosMotivosDe`, las dos líneas de los bloques `text` de la referencia que llevan `<precepto>` y no empiezan por `⚠`, una con `<norma>` y otra sin ella; cada fila, `| <procedimiento> | <línea> |`, sustituye primero `<motivo>` por el motivo de su caso y después, con un solo `strings.NewReplacer`, todos los marcadores conocidos (los de hoy más `<asunto>`, por el procedimiento); los cuatro defectos de §5 con el prefijo `<skill>: references/<ref>.md: `, los marcadores desconocidos una vez cada uno y en su orden, y sin motivos buscados en una skill sin plantilla; `conFormas` pasa a `conPlantilla`; `filasSinRegion` no cambia; internal/skills/sincronia.go, solo lo que esa llamada pida; internal/skills/export_test.go, si un test externo lo necesita. **Tests**: internal/skills/regiones_test.go, `TestRegiones`: sus árboles sintéticos con la plantilla y los dos motivos; su `SKILL.md` sintético (hoy `skillMdConLasFormas`, que pasa a `skillMdConLaPlantilla`) lleva la plantilla sola en su bloque `text` y, en su prosa, fuera del bloque, la línea `Lo no declarado está «fuera de lo que kitlegal cubre».`, el `motivo_de_lo_no_declarado` que `coberturaDeLasRegiones` gana en T003, para que el defecto del motivo en la prosa que añade T005 no salga en ninguno de sus casos (los defectos esperados de los casos de hoy no cambian, ni con T004 ni con T005); y un mutante por defecto (la plantilla quitada, una segunda plantilla, un motivo quitado, un marcador desconocido en la plantilla y otro en un motivo), cada uno con su defecto exacto; la fila sin región, como hoy; internal/skills/referencias_test.go, `TestRegionDeLaCobertura`: `filaDeLosTributos`, byte a byte la de la historia 3, escenario 1, del spec; las 9 filas del paquete; la referencia del repositorio es la generada; su mutante, la plantilla cambiada sin regenerar, da la deriva; las constantes de los dos tests con la línea vieja pasan a la plantilla. **Skill**: skills/plazos-y-territorio/SKILL.md, los tres cambios de líneas enteras de contracts/skills.md §2 (paso 7, «Cómo se presenta un plazo» con la plantilla sola en su bloque y sin la marca, regla 9) y nada más; skills/plazos-y-territorio/references/administrativo.md, la frase y el bloque `text` de los dos motivos de §3, delante de la marca de inicio de la región, y la región regenerada con `make skills-sync`, sin escribirla a mano. **Vertical de ensayo**: internal/skills/testdata/vertical-de-ensayo/skills/plazos-y-territorio/references/ensayo.md gana, delante de su región, la frase y el bloque de los dos motivos con los mismos bytes que `administrativo.md`. **Guiones de H32**: internal/app/testdata/script/h32-paquetes.txtar, solo su `-- tributos.txt --`, que pasa a ser la fila de la tabla con la línea de la historia 3, escenario 1; internal/app/testdata/script/h32-capacidades.txtar, en `-- formas.txt --` la fila `plazos-y-territorio|⚠ NO CUBIERTO:|1` y el comentario de cabecera que dice una forma, la plantilla, con `formas: 16 de 16` sin cambiar; las copias de H32 bajo `specs` no se tocan. **Suite congelada**: internal/evals/laboratorio_familias_test.go: `declaracionDelProducto` gana `enLasReferencias []cambioEnUnaReferencia` con `cambioEnUnaReferencia{referencia, antes, despues string}`; `conLaDeclaracion` los aplica a `references/<referencia>` de la base sin lo generado con la regla de los del cuerpo (el `antes`, una sola vez; si no, el error con su número y su referencia); `sinLoGenerado` reduce también a sus marcas las regiones de cada referencia escrita a mano (desde la línea que empieza por `<!-- inicio de ` y termina en ` con make skills-sync, no editar -->` hasta la `<!-- fin de <generador> -->` de su generador; sin cerrar, el error de hoy con la ruta); la declaración de `plazos-y-territorio` gana los tres cambios de §2 en `cambios` y el de §3 en `enLasReferencias`; los subtests de `normativa` y `jurisprudencia` no cambian. **Verificación**: `make skills-sync` y después `git status --porcelain` sin ficheros nuevos cambiados; el cuerpo de `plazos-y-territorio` mide 10 686 caracteres (quickstart, escenario 6); los `go test` de quickstart, escenarios 3 y 5 (sin `TestRegionDeLoCubierto`), y `make ci` en verde, con `TestSkillsDelRepositorio` (`deriva`, `presupuestos`, `formas-compartidas`, `forma-de-cita`, `ordenes-de-ejemplo`) sin cambiar; a mano, sobre una copia del repositorio en `$TMPDIR`: un byte cambiado en `skills/plazos-y-territorio/SKILL.md` fuera de la declaración, y otro en lo escrito a mano de `administrativo.md`, ponen en rojo el subtest de `plazos-y-territorio` de `TestSkillsDelProductoDesdeControl`.

**Checkpoint**: la tabla de lo regulado aparte es la plantilla con su motivo, y las suites de H32 y de
`plazos-y-territorio` siguen en verde con sus cambios declarados.

---

## Phase 5: User Story 2 — lo cubierto llega generado a la skill de entrada (Priority: P1)

**Goal**: `make skills-sync` escribe en `consulta-juridica` la región de lo cubierto, y su §4 lleva la plantilla y la
marca (FR-003, FR-007, FR-008, FR-011). Da también el texto de la historia 1.

**Independent Test**: `make skills-sync` sin diff; `TestRegionDeLoCubierto` y `TestVerticalDeEnsayo`, con la fila de la
vertical en la región.

- [X] T005 [datos] [US2] La región de lo cubierto y la skill de entrada: el generador, el motivo en la prosa, el `SKILL.md` de `consulta-juridica`, la huella de su suite congelada y la fila de lo cubierto de la vertical de ensayo, en una sola tarea porque el `SKILL.md` nuevo cambia la huella de `TestConsultaJuridicaDesdeLaBase`, sus marcas sin generador dejan la región sin llenar y `deriva` en rojo, `partirElCuerpo` partiría sus líneas, y la vertical con su fila y sin el generador ni los tres cambios de su test sale en rojo (FR-003, FR-004, FR-007, FR-008, FR-010, FR-011, FR-020, FR-021, FR-023; SC-001, SC-003, SC-005, SC-008, SC-009; contracts/datos-y-regiones.md §3, §4 y §6; contracts/skills.md §1; contracts/controles.md §1, §2 y §4; data-model §3; research D1, D3, D4, D14, D16, D17; plan.md «Orden de implementación», paso 5). **Región**: internal/skills/cubierto.go, nuevo: `regionDeLoCubierto`, un `regionDeMarcas` con las marcas `<!-- lo que cubre la base: lo genera make skills-sync -->` y `<!-- fin de lo que cubre la base -->`, y su generador como método de `datosDeLasRegiones`, que lee `data/cobertura.yaml` y `data/normas.yaml` con los lectores perezosos de las regiones (`leerUnDato`, `usar`), una vez por sincronía: una línea por fila de `cubierto`, en su orden, `- **<Capacidad con su primera letra en mayúscula>**: <alcance>`, sin escribir `materia`; los defectos de §3 (sin `data/cobertura.yaml`, con defectos en él una vez cada uno, una `materia` que no es el `vertical` de ninguna norma, marcas que no van en pareja), con los que la región queda como está; internal/skills/sincronia.go, `regenerarSkill` la llama justo detrás de `paquetes.regenerar` y antes de medir el cuerpo; internal/skills/regiones.go, en toda skill con plantilla y sobre el `SKILL.md` regenerado, el defecto `<skill>: SKILL.md: la prosa no dice el motivo de lo no declarado, «fuera de lo que kitlegal cubre»` si el texto, con cada tramo de blancos reducido a un espacio, no contiene `«` + el motivo + `»`, y sin `data/cobertura.yaml` el de §4; internal/skills/export_test.go, si un test externo lo necesita. **Test**: internal/skills/cubierto_test.go, nuevo, `TestRegionDeLoCubierto`, sobre una copia en `t.TempDir()` de `consulta-juridica`, `plazos-y-territorio` y los datos del repositorio: la región da las 4 líneas en su orden y la sincronía no cambia nada; y sus tres mutantes, cada uno con los defectos exactos de cada skill de la copia: `materia: Laboral` (el de la materia, de la sincronía y no del esquema), sin `data/cobertura.yaml` (en `consulta-juridica`, el de la región y el de la plantilla; en `plazos-y-territorio`, el de la plantilla y el de hoy de su región `cobertura_por_materia`; nada regenerado) y el motivo cambiado en la prosa de §4 (el del motivo). **Skill**: skills/consulta-juridica/SKILL.md, los cambios de contracts/skills.md §1 con sus cortes de línea: en §3, las cuatro viñetas de las capacidades pasan a ser las dos marcas, que llena `make skills-sync` (la viñeta «Paquetes de materia» con su región, la de «Sin herramienta, no cubiertos», el párrafo de entrada y el frontmatter no cambian); §4, su título, su párrafo y el bloque `text` con la plantilla y la marca, y detrás el párrafo «Por qué» sin cambios; el cuerpo, 4481 caracteres. **Recorte de los presupuestos**: internal/app/skills_test.go, `partirElCuerpo` no parte las líneas entre las marcas de lo cubierto, como no parte las de la tabla de comandos; nada más del fichero cambia. **Suite congelada**: internal/evals/descripciones_del_producto_test.go, `huellasDeConsultaJuridicaEnLaBase[SKILL.md]` pasa a la SHA-256 del `SKILL.md` de la entrega sin las líneas de la clave `description` (`descripcionYCuerpo`) y con sus dos regiones reducidas a sus marcas (`conLasRegionesGeneradasReducidas`), calculada dos veces, con el código del test y con `awk` y `sha256sum`, y las dos iguales; los mensajes, `… falta: está en la entrega de H33`, `… sobra: no está en la entrega de H33` y `… difiere de la entrega de H33`, y sus cinco casos de hoy dan su defecto exacto con esos textos («solo la clave», ninguno); internal/evals/laboratorio_familias_test.go, `regionesGeneradas` gana `{inicio: "<!-- lo que cubre la base: ", fin: "<!-- fin de lo que cubre la base -->"}` y `{inicio: "<!-- paquetes de la base: ", fin: "<!-- fin de los paquetes de la base -->"}`. **Vertical de ensayo**: internal/skills/testdata/vertical-de-ensayo/data/cobertura.yaml gana, delante de `regulados_aparte`, `cubierto` con la fila `{capacidad: plazos, alcance: los de ensayo del molde de los paquetes., materia: ensayo}` de §6; internal/skills/vertical_de_ensayo_test.go, los tres cambios de §6: `anadirEntradas` inserta las entradas de cada clave al final del bloque de esa clave en la copia; `partirPorLasMarcas` conoce las dos marcas de lo cubierto; `comprobarLoSincronizado` exige en la región de `consulta-juridica` la línea `- **Plazos**: los de ensayo del molde de los paquetes.`, la última, detrás de la del territorio; cada cuerpo, como mucho 11 000 caracteres (10 756 el de `plazos-y-territorio`). **Verificación**: `make skills-sync` llena la región con las cuatro líneas de contracts/skills.md §1 y después no cambia ningún fichero; el cuerpo de `consulta-juridica` mide 4481 caracteres y el de `plazos-y-territorio` 10 686 (quickstart, escenario 6); los `go test` de quickstart, escenarios 3 y 5, y `make ci` en verde, con `TestRegiones` en verde sin tocar internal/skills/regiones_test.go (su `SKILL.md` sintético dice el motivo en la prosa desde T004) y con `formas-compartidas` comparando ya la plantilla de las dos skills; a mano, sobre una copia del repositorio en `$TMPDIR`: un carácter cambiado en la plantilla de `consulta-juridica` pone en rojo `formas-compartidas`, y una línea de la región de lo cubierto cambiada a mano pone en rojo `deriva`.

**Checkpoint**: lo cubierto llega generado a `consulta-juridica`, que declara con la misma plantilla que
`plazos-y-territorio` y marca lo leído para completar.

---

## Phase 6: User Story 4 — las evals miden la frontera sin modelo (Priority: P1)

**Goal**: el formato admite la eval fuera de lo declarado y `tambien_la_resuelven`, el juicio la reconoce y los umbrales
nacen sin caso especial (FR-012 a FR-015).

**Independent Test**: `TestFormatoDeLaCobertura`, `TestJuzgarLaDeclaracionDeNoCubierto`, `TestActivadaPor` y
`TestSkillsNombradasEnTambienLaResuelven`.

- [X] T006 [datos] [US4] El formato y el juicio de la frontera, sin evals nuevas, en una sola tarea porque el esquema de eval y su lector estricto tienen que admitir a la vez `fuera_de_lo_declarado` y `tambien_la_resuelven`, y el juicio y los umbrales leen esos campos (FR-012, FR-013, FR-014, FR-015, FR-021; SC-008; contracts/evals.md §1 a §3; contracts/controles.md §5; data-model §4; research D6, D7, D8, D9, V4, V12; plan.md «Orden de implementación», paso 6). **Esquema**: schemas/eval.yaml.json: `no_cubierto` con `additionalProperties: false`, `if: {required: [fuera_de_lo_declarado]}`, `then: {not: {anyOf: [{required: [norma]}, {required: [precepto]}]}}`, `else: {required: [norma, precepto]}` y `fuera_de_lo_declarado` `{const: true}`, sin `oneOf`, y `norma` y `precepto` sin cambios; en la raíz, `tambien_la_resuelven` con la forma de `se_activan` (lista no vacía, sin repetir, de nombres `^[a-z0-9]+(-[a-z0-9]+)*$`); las demás reglas de `no_cubierto` no cambian. **Formato**: internal/evals/formato.go: `NoCubiertoEsperado` gana `FueraDeLoDeclarado bool` (`yaml:"fuera_de_lo_declarado"`) y `Eval`, `TambienLaResuelven []string` (`yaml:"tambien_la_resuelven"`); `esDeCobertura` es `Norma != "" || FueraDeLoDeclarado`; `Eval.activadaPor(sesion Sesion, skill string) bool`, verdadero si `sesion.Activada(skill)` o la de alguna de `TambienLaResuelven`. **Juicio**: internal/evals/plazos.go: `declarada`, con `FueraDeLoDeclarado`, verdadero si `ExtraerNoCubierto` da alguna declaración y, si no, la regla de hoy (norma y cita del precepto en la misma línea); la constante `motivoFueraDeLoDeclarado` (`fuera de lo declarado`); `juzgarElPlazo` da `declaración de no cubierto ausente: ` seguido de la norma esperada o de esa constante, y cada línea del plazo sigue dando `línea del plazo en un asunto no cubierto: <fecha>`; ninguna constante del paquete de las evals lleva texto de la plantilla, del motivo ni de la cola; internal/evals/juzgar.go (`juzgarEnModo`) e internal/evals/informe.go (`juzgarSesion`) fijan `ResultadoDeEval.Activada` con `activadaPor`; `se_activan`, `no_se_activan` y el sondeo del laboratorio no cambian. **Umbral**: internal/evals/umbrales.go, solo `descripcionDelUmbralDeCoberturaSinDeclarar`, palabra por palabra la de FR-014; ningún otro código de los umbrales cambia. **Conjunto**: internal/evals/conjunto.go, `ComprobarLasSkillsNombradas` mira también `tambien_la_resuelven` con su mensaje de hoy (`… tambien_la_resuelven nombra <nombre> …`). **Tests**: internal/evals/formato_test.go, `TestFormatoDeLaCobertura`, nuevo, con los tres casos de contracts/evals.md §1 (`{norma, precepto}` se lee y es de cobertura; `{fuera_de_lo_declarado: true}` se lee y es de cobertura; las dos variantes a la vez, rechazada con un defecto en `no_cubierto`), y `TestFormatoDePlazos` sin cambiar ni sus casos ni sus mensajes; internal/evals/plazos_test.go, `TestJuzgarLaDeclaracionDeNoCubierto`, nuevo, `juzgarElPlazo` sobre las once respuestas constantes de la tabla de §2, cada una con su `NoCubiertoDeclarado` y sus motivos exactos, y en `TestJuzgarPlazos` solo sus constantes con la línea vieja pasan a la plantilla con sus datos; internal/evals/tambien_la_resuelven_test.go, nuevo, `TestActivadaPor` (`activadaPor` con y sin la otra skill declarada; en rojo si la activación de la otra skill no cuenta) y `TestSkillsNombradasEnTambienLaResuelven` (un nombre que no es una skill del producto es un defecto del conjunto); internal/evals/juzgar_test.go (`TestJuzgarLasLecturasSinServir`) e internal/evals/informe_test.go (`TestUmbralesDePlazosYTerritorio` y los demás del fichero), solo sus constantes con la línea vieja pasan a la plantilla y la descripción del umbral pasa a la nueva; lo que comprueban no cambia. **Verificación**: el `go test` de quickstart, escenario 4, más `TestJuzgarLasLecturasSinServir` y `TestUmbralesDePlazosYTerritorio`, y `make ci` en verde; ninguna eval de `evals` cambia en esta tarea.

**Checkpoint**: el arnés reconoce las dos variantes de la eval de cobertura y la activación por otra skill; aún no hay
evals que las usen.

---

## Phase 7: User Story 1 — las evals de la frontera en dos materias (Priority: P1)

**Goal**: las cuatro evals de `consulta-juridica`, una a cada lado de la frontera en dos materias sin paquete, la regla
que lo exige, la 17 de `normativa` con `tambien_la_resuelven`, y lo que el conjunto cuenta (FR-014 a FR-019).

**Independent Test**: las evals 01 y 04 de `consulta-juridica` con `no_cubierto: {fuera_de_lo_declarado: true}`, con
`TestEvalsDelRepositorio`, `TestUmbralesDeLasSkills`, `TestCacheDeLasSesiones`, `TestDefinicionDelJob` y
`TestMatrizDeLasPropuestas` en verde.

- [X] T007 [US1] Las evals de la frontera y lo que cuenta el conjunto, en una sola tarea porque la regla «frontera en dos materias» solo se cumple con las cuatro evals, y la matriz, el peor caso, la caché y los umbrales de `consulta-juridica` cambian una sola vez con ellas; parte de lo grabado por el paso `grabar_datos` tras T002 (FR-014, FR-015, FR-016, FR-017, FR-018, FR-019, FR-021; SC-006, SC-007, SC-008; contracts/evals.md §3 a §7; contracts/controles.md §5; data-model §5; research D10, D18, V10, V12; plan.md «Orden de implementación», paso 7; DoD §1, puntos 10 y 13). **Evals**: evals/consulta-juridica/01-poliza-tras-el-fallecimiento.yaml gana al final `no_cubierto: {fuera_de_lo_declarado: true}` en bloque, y su comentario deja de decir «No exige ninguna línea fija: lo que no cubre lo dice en prosa» y dice que la respuesta declara con la plantilla lo que no cubre (a quién reclamar, qué vía), con el motivo fuera de lo declarado; lo demás no cambia; evals/consulta-juridica/03-seguro-prorroga.yaml, nueva, y evals/consulta-juridica/04-despido-plazo.yaml, nueva, byte a byte las de contracts/evals.md §5, sin `lecturas` (ADR 0043, punto 2); evals/normativa/17-ltaibg-plazo-por-materia.yaml gana tras su comentario de hoy el comentario y `tambien_la_resuelven: [plazos-y-territorio]` de §5, en bloque, y sus comandos y sus citas no cambian; la 02 de `consulta-juridica` y las 07, 08 y 12 de `plazos-y-territorio` no cambian. **Regla**: internal/evals/conjunto.go, `incumplimientoDeLaFronteraEnDosMaterias`, la cuarta de `ReglasDeConsultaJuridica`, detrás de «una sola capacidad», con nombre `frontera en dos materias`: cuenta las normas de `data/normas.yaml` sin `vertical` que cita (`normasCitadas`) una eval activa que no es de cobertura y otra activa de cobertura con `fuera_de_lo_declarado` que prohíbe `plazos calcular`; con menos de 2, el defecto de §4 (`hay <n> normas de data/normas.yaml sin vertical con una eval activa a cada lado de la frontera y el conjunto lleva al menos 2: …; evals activas: …`); sin nombrar ninguna norma ni materia; las tres reglas de hoy no cambian. **Tests**: internal/evals/conjunto_test.go, `TestEvalsDelRepositorio`: «formato» con las cuatro evals; el mutante de «no activación de cada capacidad» quita `plazos-y-territorio` de `no_se_activan` en la 02 y en la 03, cada uno con su defecto; el conjunto de la rama cumple la regla nueva con 2 (`BOE-A-1980-22501`, `BOE-A-2015-11430`) y su mutante sin la 03 da el defecto con 1 (`BOE-A-2015-11430`); internal/evals/informe_test.go, `TestUmbralesDeLasSkills`, subtest `consulta-juridica`: `sin_activar:claude-sonnet-5-5:orden` (total 12), `cobertura_sin_declarar:claude-sonnet-5-5:orden` (total 6), los dos de `herramienta` (12 y 6), en ese orden, los cuatro con `Decide: true` y umbral 0, delante de los cuatro `fuera_de_lo_grabado` de hoy, y su comentario deja de decir «sin `sin_activar`»; internal/evals/preparar_test.go, `TestCacheDeLasSesiones`, subtest `consulta-juridica`: desde la caché de una sesión de la 04, `boe articulo BOE-A-2015-11430 a59 --fecha 2026-11-02 --offline` sale con 0, y desde la de la 01, los arts. 22 y 4 de la Ley 50/1980, como hoy; internal/evals/definicion_test.go, `TestDefinicionDelJob`: el peor caso de `consulta-juridica`, `SesionesPorGrupo: []int{25, 24}`, concurrencia 1, 15 013 s con su texto (`15013 s` … `(⌈25 / 1⌉ + ⌈24 / 1⌉)` …), bajo el tope de 21 120 s, y la concurrencia de las demás skills sin cambiar; internal/evals/eleccion_test.go, `goldenDeLaMatriz` de `TestMatrizDeLasPropuestas` con las evals nuevas de cada ejecución, regenerada desde la salida del test y revisada línea a línea contra las evals del conjunto. **Job**: .github/workflows/evals.yml, solo el comentario del tope: «15 013 s en consulta-juridica, 250,2 minutos, con sus cuatro evals» donde dice «8485 s …, 141,4 minutos, con sus dos evals»; ni `VERSION_DE_CLAUDE_CODE`, ni `MODELO_DE_EVALS`, ni la concurrencia, ni el tope. **Verificación**: el `go test` de quickstart, escenario 4, con `TestEvalsDelRepositorio`, `TestUmbralesDeLasSkills`, `TestCacheDeLasSesiones`, `TestDefinicionDelJob` y `TestMatrizDeLasPropuestas`, y `make ci` en verde; a mano, sobre una copia del repositorio en `$TMPDIR`: quitar `no_cubierto` de la 04 deja a `consulta-juridica` sin `cobertura_sin_declarar` de cada modo y pone en rojo `TestUmbralesDeLasSkills`, y una eval más con seis sesiones por grupo de más en un trabajo hasta pasar de su tope pone en rojo `TestDefinicionDelJob`.

**Checkpoint**: las cuatro skills tienen las evals y los umbrales que mide el cierre del workflow (FR-024, SC-004).

---

## Phase 8: Polish — documentación, changelog y verificación final

**Purpose**: FR-022 y la Definition of Done; la verificación de la rama entera (SC-009).

- [X] T008 La documentación de FR-022, sin código ni control en `make ci` (FR-022; contracts/skills.md §4; contracts/datos-y-regiones.md §5; quickstart, escenario 9). **Ficheros**: docs/SKILLS.md: la plantilla única y la marca `⚠ NO VERIFICADO:`; los dos motivos en la referencia del paquete; la región de lo cubierto y sus marcas; en la tabla de los defectos de la región, «la plantilla» donde dice «las dos formas» y la fila nueva de los motivos, con los mensajes de contracts/datos-y-regiones.md §5; la fila de `consulta-juridica` de la tabla de las skills sin «en prosa» ni «sin… líneas `⚠`»; CONTRIBUTING.md: la fila de `no_cubierto` con `fuera_de_lo_declarado`, la fila nueva de `tambien_la_resuelven`, la de `cobertura_sin_declarar` con la descripción de FR-014 y los umbrales de `consulta-juridica` con `sin_activar` y `cobertura_sin_declarar`; docs/HISTORIAL.md: la viñeta del hito, con lo entregado y sus cifras medidas en el run (4 filas de lo cubierto, 9 de la tabla, 4 evals de `consulta-juridica`, cuerpos de 4481 y 10 686 caracteres, peor caso de 15 013 s) y sin las de la medición del cierre, que no se mide en el run (research S3); docs/USO.md: la entrada de la bitácora de FR-022, con la fecha del día de la tarea: un criterio de sondeo o de umbral es un defecto que vería quien usa kitlegal, no la ruta entre skills, y los `⚠ NO CUBIERTO:` falsos los cazó leer las respuestas, no un umbral; web/src/pages/como-se-evalua/index.astro: la frase que anuncia H33 pasa a decir lo hecho, sin ninguna cita escrita a mano. Ningún ADR: el 0058 ya está (FR-022). **Verificación**: los `grep` de quickstart, escenario 9, dan al menos `2`, `0`, `0`, `1` y al menos `1`; `make ci` en verde.
- [X] T009 La entrada del changelog y la verificación final de la rama, sin cambiar otro fichero (FR-020, FR-021, FR-022, FR-023, FR-025; SC-009; DoD §1, puntos 1 y 6; quickstart, escenarios 3 a 8). **Fichero**: CHANGELOG.md, al final de «## [Unreleased]» → «### Cambiado», una viñeta que dice la plantilla única `⚠ NO CUBIERTO:` en todas las materias con la marca `⚠ NO VERIFICADO:` de `consulta-juridica`, lo cubierto declarado en `data/cobertura.yaml` por capacidad y materia y generado en `consulta-juridica`, y las evals de la frontera en dos materias con sus umbrales; nada más del fichero cambia. **Verificación**: `make ci` en verde, con `git status --porcelain` igual antes y después (quickstart, escenario 8); `make skills-sync` seguido de `git status --porcelain` sin cambios; los `go test` de quickstart, escenarios 3, 4 y 5, en verde, y `git diff --stat main` sin ninguna línea de las copias de la suite de H32 bajo `specs`; los cuerpos, 4481 y 10 686 caracteres (escenario 6); la grabación del art. 59 presente en su commit de `grabar_datos` (escenario 7); y `go test ./...` y `go test -tags=integration` en verde en los paquetes que nombra SC-009.

---

## Dependencies & Execution Order

Estrictamente en secuencia, de T001 a T009: ninguna depende de una posterior.

- **T001** no depende de nada: la suite se escribe antes que el código y queda congelada.
- **T002** no depende de código; el paso `grabar_datos` graba tras su commit, y **T007** es la primera que lo lee.
- **T003** deja `Cobertura.MotivoDeLoNoDeclarado` y `Cobertura.Cubierto`, que leen **T004** (la plantilla y los motivos,
  sin el motivo de los datos todavía) y **T005** (la región y el motivo en la prosa).
- **T004** deja la plantilla en `plazos-y-territorio`; **T005** la pone en `consulta-juridica`, y desde ahí
  `formas-compartidas` compara las dos.
- **T006** deja el formato y el juicio; **T007** escribe las evals que los usan.
- **T008** documenta lo de T003 a T007; **T009** cierra con el changelog y la verificación entera.

Historias: US5 abre (T001) y la comprueba entera la verificación de T009; US2 (T003, T005), US3 (T004), US4 (T006) y US1 (T005 en el texto, T007 en
sus evals) se suceden en el orden de dentro afuera del plan.

## Parallel Opportunities

Ninguna: cada tarea parte de la anterior y varias tocan los mismos ficheros (ver «Format»). El workflow las ejecuta una
a una.

## Implementation Strategy

- **MVP**: T001 a T005: la plantilla única en las dos skills, lo cubierto en datos y generado, y la tabla de lo regulado
  aparte con su motivo; con ello pasa el guion de aceptación de US5.
- **Incremento 2**: T006 y T007: la frontera medida sin modelo y los umbrales que hace cumplir el cierre del workflow.
- **Cierre**: T008 y T009. La medición de las evals en los dos modos (FR-024, SC-004) la hace el workflow tras la
  revisión final, no una tarea.

## Controles de umbral → tarea que construye el control (ADR 0029)

| Fila de plan.md «Controles de umbral» | Control | Tarea | Cómo se ve fallar |
|---|---|---|---|
| FR-001 · 0 diferencias con los textos de FR-001 | `TestLeerCobertura`, subtest `repositorio` | T003 | un carácter cambiado en el motivo o en una fila |
| FR-002, SC-008 · 0 de 6 mutantes del esquema aceptados | `TestLeerCobertura` | T003 | los seis mutantes de contracts/datos-y-regiones.md §2 |
| FR-003, SC-008 · 0 de 3 mutantes de la sincronía sin su defecto | `TestRegionDeLoCubierto` (nuevo) | T005 | la materia, sin `data/cobertura.yaml` y el motivo cambiado |
| FR-007, SC-001 · 4 de 4 filas en la región, en su orden | `TestRegionDeLoCubierto` (nuevo) | T005 | las líneas entre las marcas tras regenerar la copia |
| FR-007, FR-023, SC-001 · 0 ficheros que cambie `make skills-sync` | `TestSkillsDelRepositorio`, `deriva` | T004 y T005 | el mutante de `TestRegionDeLaCobertura` (T004) y una línea de la región de lo cubierto cambiada en una copia en `$TMPDIR` (T005) |
| FR-004, SC-003 · 0 diferencias entre las dos plantillas | `TestSkillsDelRepositorio`, `formas-compartidas` | T005 | un carácter cambiado en la plantilla de `consulta-juridica`, en una copia en `$TMPDIR` |
| FR-006, SC-002 · la fila de los tributos byte a byte; 9 de 9 filas | `TestRegionDeLaCobertura` | T004 | la plantilla cambiada sin regenerar |
| FR-006, SC-008 · 0 de 5 mutantes del generador sin su defecto | `TestRegiones` | T004 | los cinco mutantes de contracts/controles.md §1 |
| FR-005, FR-009, FR-020 · 0 bytes fuera de lo generado y de la declaración | `TestSkillsDelProductoDesdeControl`, `plazos-y-territorio` | T004 | un byte fuera de la declaración en el `SKILL.md` y otro en `administrativo.md`, en una copia en `$TMPDIR` |
| FR-008, FR-020 · 0 bytes con la huella de la entrega | `TestConsultaJuridicaDesdeLaBase` | T005 | sus cinco mutantes de hoy, con los mensajes nuevos |
| FR-010, SC-005 · como mucho 11 000 caracteres por cuerpo | `TestSkillsDelRepositorio`, `presupuestos` (`defectosDelCuerpo`) | T004 y T005 los mantienen; `partirElCuerpo` en T005 | sus mutantes de hoy, que parten el cuerpo hasta pasar del tope |
| FR-007, FR-010, FR-011, SC-005 · 11 000 con la vertical; su línea, la última | `TestVerticalDeEnsayo` | T004 (motivos) y T005 (fila y los tres cambios) | la línea de la vertical ausente de la región o fuera de su sitio |
| FR-012, SC-008 · 0 evals con las dos variantes aceptadas | `TestFormatoDeLaCobertura` (nuevo), `TestFormatoDePlazos` | T006 | el mutante con las dos variantes a la vez |
| FR-013, SC-008 · 0 de 11 respuestas al revés o con otro motivo | `TestJuzgarLaDeclaracionDeNoCubierto` (nuevo) | T006 | las once respuestas constantes de contracts/evals.md §2 |
| FR-014 · los cuatro umbrales de `consulta-juridica` con `Decide: true` | `TestUmbralesDeLasSkills`, `consulta-juridica` | T007 (con la descripción de T006) | sin `no_cubierto` en la 04, en una copia en `$TMPDIR` |
| FR-015 · 0 activaciones sin contar; 0 nombres ajenos aceptados | `TestActivadaPor`, `TestSkillsNombradasEnTambienLaResuelven` (nuevos) | T006 | la otra skill sin declarar; un nombre que no es una skill de `skills/` |
| FR-016, FR-017, SC-007, SC-008 · cuatro evals; 2 normas a cada lado | `TestEvalsDelRepositorio` | T007 | el mutante sin la 03 y el de «no activación» sobre la 02 y la 03 |
| FR-018 · 0 consultas sin servir desde lo grabado | `TestCacheDeLasSesiones`, `consulta-juridica` | T007 (lo grabado, T002) | la consulta del art. 59 sin grabar sale con un código distinto de 0 |
| FR-019, SC-006 · peor caso de 15 013 s, como mucho 21 120 s | `TestDefinicionDelJob` | T007 | un trabajo con sesiones de más hasta pasar su tope, en una copia en `$TMPDIR` |
| FR-019 · 0 bytes con la matriz de cada ejecución | `TestMatrizDeLasPropuestas` | T007 | un byte de diferencia con `goldenDeLaMatriz` |
| FR-020, FR-021, SC-009 · 0 tests en rojo, con y sin `integration` | `make test`, `make test-integration` | todas; T009 lo comprueba entero | cualquier test en rojo |
| FR-004, FR-025, SC-001, SC-002, SC-003, SC-005 · 100 % de las aserciones del guion | el guion congelado, en `TestEntregaDelHito` | T001 | rojo primero, que comprueba el workflow sobre el código de hoy |
| FR-014, FR-024, SC-004 · los umbrales `evals:` de `consulta-juridica`, `plazos-y-territorio`, `normativa` y `jurisprudencia` | el job de evals del cierre del workflow | T006 (juicio y descripción) y T007 (las evals que hacen nacer los de `consulta-juridica`); los demás ya existen | lo mide el cierre del workflow; ninguna tarea lo ejecuta |

## Definition of Done (docs/ROADMAP.md §1) → tarea

| Punto | Aplica | Tarea |
|---|---|---|
| 1 · `make ci` en verde | sí | todas; T009, la rama entera |
| 2 · tests offline; fixtures grabados | sí: el art. 59 | T002 (manifiesto; lo graba `grabar_datos`); los tests de T003 a T007 |
| 3 · `net/http`, `os.Exit`, `fmt.Print*` | sin nada nuevo | convenciones; `depguard` y `forbidigo` en `make ci` |
| 4 · salidas de applet contra `schemas/` | no: ningún applet ni sobre cambia | — |
| 5 · errores tipados, ningún `panic` | sin errores de producción nuevos | convenciones |
| 6 · e2e y `CHANGELOG.md` | sí | T001 (guion), T004 (guiones de H32), T009 (changelog) |
| 7 · ADR | no: el 0058 ya está (FR-022) | — |
| 8 · `docs/SOURCES.md` | no cambia: la fuente tiene fila «Revisado» 2026-09-13 en `main` y el hito no cambia su uso | — |
| 9 · cobertura de `internal/core` | `internal/core` no cambia | — |
| 10 · evals de la skill | sí | T007, tras el formato de T006 (research D18) |
| 11 · dimensión territorial | no | — |
| 12 · operaciones de grafo | no | — |
| 13 · un asunto cubierto y otro de fuera de lo declarado | sí | T007 (la regla «frontera en dos materias» y las 01 a 04); lo regulado aparte, las 07, 08 y 12, sin cambios |
| esquemas (`schemas/`) | sí | T003 (`cobertura.yaml.json`), T006 (`eval.yaml.json`) |
