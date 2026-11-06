---

description: "Tareas de H41 · Las skills del producto mandan leer sus referencias"
---

# Tasks: H41 · Las skills del producto mandan leer sus referencias

**Input**: `specs/034-h41-las-skills-del/` — [spec.md](spec.md), [plan.md](plan.md), [research.md](research.md),
[data-model.md](data-model.md), [quickstart.md](quickstart.md) y [contracts/](contracts/) (`remisiones.md`,
`variantes.md`, `arnes.md`, `tabla.md`, `aceptacion.md`).

**Tests**: obligatorios (constitución III). La primera tarea escribe la suite de aceptación, que el workflow congela;
cada tarea posterior trae su test y la implementación que lo pone en verde, y deja `make ci` en verde por sí sola.

**Organización**: una fase por historia, en el orden en que se ejecutan una tras otra. La línea de cada tarea declara
todas las rutas que crea o modifica (el guardián rechaza cualquier otra; `go.mod`, `go.sum`, `CHANGELOG.md` y el
directorio del feature siempre se admiten, y declarar `x.go` admite `x_test.go`).

## Format: `[ID] [P?] [Story] Descripción`

- **[P]**: puede ir en paralelo (ficheros distintos, sin dependencia de una tarea sin terminar). Ninguna lo lleva.
- **[Story]**: la historia de [spec.md](spec.md) a la que sirve (US1 a US5; US6 no tiene tareas: la hace la persona
  fuera del run, FR-052 y FR-053).
- **[aceptacion]**: la suite congelada (solo T001). **[datos]**: la tarea toca ficheros bajo `testdata/` (solo T008).

## Convenciones de todas las tareas

- `make ci` en verde al terminar cada una, en primer plano y por objetivos si hace falta (`make fmt-check`,
  `make lint`, `make test`, `make test-integration`, `make vuln`, `make schema-check`, `make skills-check`…). Ninguna
  usa la red más que la de las herramientas de Go, ninguna graba (no hay datos externos: plan.md «Datos externos»;
  ni manifiesto `grabaciones.json` ni `TestGrabar*`), ninguna escribe en `evidencias/` ni en `data/`, ninguna abre una
  sesión con modelo ni ejecuta `TestSondeoDelLaboratorio`, `make evals-laboratorio`, `make evals` ni los guiones
  `scripts/evals*.sh` desde la terminal (FR-031, FR-052).
- Punto de partida (research D20): lo que vuelve tal cual de la #245 se toma de la rama `033-h41-las-skills-del` (local
  y en `origin`; `git show 033-h41-las-skills-del:<ruta>`) —`laboratorio_familias.go`, el arnés, la tabla, la lectura
  con `cd`, `remisiones_test.go`, `.golangci.yml`, los fixtures de la tabla— con los cambios de este plan («Lo que
  cambia respecto a la #245»). **No** vuelven: los cambios de `skills/plazos-y-territorio` y de sus referencias; los
  guiones de la #245; y, en `referenciasDeLasVariantes` de `laboratorio_tabla.go`, la comprobación de un `sondeo.json`
  sin `referencias_del_producto` y su mensaje (research D17, D20; gates/supuestos.md).
- Los cambios de las skills son, byte a byte, los `diff -u` de [contracts/remisiones.md](contracts/remisiones.md) §2 a
  §4; nada más de `skills/` cambia (FR-007, FR-009): ni el frontmatter ni la `description`, ni la tabla de comandos, ni
  la región de los paquetes, ni ninguna referencia salvo las dos de §3. `skills/plazos-y-territorio/` y
  `skills/consulta-juridica/` no se tocan (FR-010). No se tocan `evals/`, `data/`, `internal/mcp/`, el generador de
  `internal/skills` (`sincronia.go`, `regiones.go`, `markdown.go`…), `laboratorio/variantes.yaml`,
  `laboratorio/variantes/`, `laboratorio/evals/`, las variantes de H40 (`laboratorio/normativa/`), las recetas del
  `Makefile` (salvo la expresión `-run` de `skills-check`, T004) ni ningún fichero de `testdata/` que ya exista
  (Fuera de alcance; FR-021, FR-042). Ningún argumento, bandera ni orden nueva de `kitlegal` ni de `cmd/laboratorio`;
  ningún umbral cambia.
- Los mensajes del arnés son los de [contracts/arnes.md](contracts/arnes.md) §2; la frase, la de
  [contracts/remisiones.md](contracts/remisiones.md) §6; la regla y la lista, las de §5 y §7; el cálculo desde control,
  el de §8; las variantes de `plazos-y-territorio`, las de [contracts/variantes.md](contracts/variantes.md) §2 y su
  control, el de §4; la tabla y la lectura con `cd`, las de [contracts/tabla.md](contracts/tabla.md) §1 a §3. Nada se
  inventa fuera de ellos.
- `internal/evals` no nombra `os.Stdin`, `os.Stdout`, `os.Stderr` ni llama a `os.Exit` (R4, R5); ningún `panic`; lo que
  llega de fuera manipulado es el defecto `inesperado` (1) de hoy, sin comprobación, mensaje ni test propios (research
  D17). Ningún `//nolint`. Tests de tabla con `t.Parallel()`.
- Cada control se ve fallar con un mutante (la medida pasada del umbral) antes de darlo por bueno, y el mutante se
  revierte (ADR 0029). Los prototipos y guiones de un solo uso van bajo `$TMPDIR` y no se versionan.
- Las suites congeladas ya activas (las de H39, H40 y `h32-normativa.txtar`) siguen en verde sin cambiar un fichero suyo
  (FR-042; research V4). Los fixtures nuevos no llevan nada con forma de secreto (gitleaks corre en `make ci`).

---

## Phase 1: Setup — la suite de aceptación

**Propósito**: fijar, antes de cualquier código, lo que instala el binario y lo que la tabla tiene que dar con las
sesiones de las tres familias (ADR 0018; plan.md «Aceptación e2e»).

- [X] T001 [aceptacion] Escribir, a partir del spec y de [contracts/aceptacion.md](contracts/aceptacion.md), sin código de producto y sin evals nuevas (Fuera de alcance: «Evals nuevas, y cambiar las que ya existen»), los dos guiones testscript specs/034-h41-las-skills-del/aceptacion/skills-instaladas.txtar y specs/034-h41-las-skills-del/aceptacion/laboratorio-tabla.txtar (el workflow los activa como `h41-skills-instaladas.txtar` y `h41-laboratorio-tabla.txtar` en la carpeta de guiones de internal/app al terminar el bucle, y los ejecuta `TestEntregaDelHito`; el arnés e2e no cambia: `kitlegal` en el `PATH` y `KITLEGAL_LABORATORIO_BIN`, de H40). Cada guion abre con un comentario que dice lo que comprueba y una línea «Cubre:» con sus historias, FR y SC, y no contiene los textos `usage: `, `unknown command`, `unexpected command` ni `cannot parse`. **skills-instaladas.txtar** («Cubre: US1, US2, US3 · FR-001 a FR-006, FR-010, FR-025 · SC-001, SC-004, SC-005, SC-009»): `exec kitlegal skills install --dir destino --json` y, sobre lo instalado, las órdenes `grep`, `! grep` y `! exists` de contracts/aceptacion.md §2, en su orden y byte a byte (las siete órdenes de leer de `normativa` en su sitio; ninguna de `tras leer [con_herramientas]`, `las dudas, en [como_se_cita]`, «Nada de otra conversación» ni «Si una lectura falla…» en el cuerpo; los dos textos movidos al final de `redaccion_modificada.md` y `lecturas_que_fallan.md`; las dos órdenes de `jurisprudencia` y ningún «Antes de dar la consulta»; la lista de control de `plazos-y-territorio`, «Si no lo dice, **pregúntalo** y espera» en el paso 1 y ningún «lee/Lee [municipios|calculo_de_plazos.md]»; ninguna variante del laboratorio instalada). **laboratorio-tabla.txtar** («Cubre: US5 · FR-032, FR-040, FR-041 · SC-006»): las órdenes de contracts/aceptacion.md §1, en su orden y byte a byte (cada `exec sh -c '… || true'` delante de su orden con código; tablas de `normativa`, `plazos-y-territorio` y `jurisprudencia` con `cmp`, `! stderr .`; al final `! exec $KITLEGAL_LABORATORIO_BIN tabla mezcla`, `! stdout .` y la expresión de `stderr` de §1); ficheros del archivo: los sondeos `normativa/h41/`, `plazos-y-territorio/h41/`, `jurisprudencia/h41/` y `mezcla/h41/`, cada uno con su `sondeo.json` en una línea (`{"modelo":"claude-sonnet-5-5","claude_code":"2.1.317"}` y, donde §1 lo dice, `referencias_del_producto`: en `normativa`, las ocho de la familia y `nueva.md`, en orden de nombre; en `plazos-y-territorio`, sus cuatro; en `jurisprudencia`, sus dos; en `mezcla`, sin lista) y las sesiones de las tablas de §1 (`codigo-de-la-sesion`, `sesion.err` y `sesion.jsonl` con la forma de las de H40: `system/init` con el modelo y la versión, un `assistant` con su `message.id` por respuesta, los resultados en mensajes `user` y `result`; rutas de las skills bajo `/s/claude/skills/`; turnos y órdenes de `Bash` exactamente los de §1, incluidas las cinco variantes de `plazos-y-territorio`, una sesión cada una); y `normativa-tabla.md`, `plazos-y-territorio-tabla.md` y `jurisprudencia-tabla.md`, la tabla y el detalle de contracts/tabla.md §2 y §3 byte a byte, con la cabecera, las columnas, el formato de recuento, Wilson y turno de la tabla de H40 (specs/032-h40-por-qu-e/aceptacion/laboratorio-tabla.txtar): la de `plazos-y-territorio`, la del ejemplo de contracts/tabla.md §3; la de `normativa`, con la columna `nueva.md` vacía en las filas de `normativa-control` y `normativa-remisiones-en-orden` y «0 de 1» en las de `normativa-producto`, «Cumplen la regla» como en H40 y las aperturas de §1 (en orden, `redaccion_en_una_fecha.md` en el turno 4 y no `normas.md`; en herramienta, `con_herramientas.md` y `normas.md` en el turno 2); la de `jurisprudencia`, sin «Cumplen la regla». Antes de terminar, con un prototipo (worktree bajo `$TMPDIR` con los dos guiones copiados en la carpeta de guiones de internal/app con su nombre `h41-…`, `go -C <worktree> test -count=1 ./internal/app -run 'TestEntregaDelHito/^h41-'`), comprobar que cada guion falla hoy por una aserción y no por no poder leerse: `skills-instaladas` en su primer `grep`; `laboratorio-tabla` en `cmp normativa.out normativa-tabla.md` (hoy `laboratorio tabla normativa` sale con 1, `…/normativa-producto no es ninguna variante de normativa`).

**Checkpoint**: la suite queda congelada; ninguna tarea posterior la toca.

---

## Phase 2: User Story 3 (base) — Las familias del producto y sus `<skill>-control` congeladas (Priority: P1)

**Objetivo**: las tres familias declaradas en Go con su campo `producto`, y las dos `<skill>-control` nuevas copiadas
de `skills/` antes de cambiar ninguna skill, comprobadas por sus huellas. Va primero porque las tareas de US1 y US2
miden las skills contra estas copias (contracts/remisiones.md §7.3 y §8) y `normativa-control` deja de compararse con
`skills/normativa` (FR-020).

**Independent Test**: `TestVariantesDeControl`, `TestVariantesDeNormativa` y `TestLaboratorioFueraDelProducto` en verde,
y en rojo con cada mutante de sus controles (SC-007, SC-009).

- [X] T002 [US3] Las tres familias del producto, las dos `<skill>-control` nuevas y sus controles (FR-020, FR-021, FR-022, FR-025; SC-007, SC-009; contracts/variantes.md §1 y §4; data-model §1 y §2; research D9, D10, V1, V13). **Laboratorio**, lo primero de la tarea y antes de cambiar ninguna skill: laboratorio/plazos-y-territorio/plazos-y-territorio-control/skills/plazos-y-territorio/ (5 ficheros) y laboratorio/jurisprudencia/jurisprudencia-control/skills/jurisprudencia/ (3 ficheros), con `cp -R skills/<skill>/. laboratorio/<skill>/<skill>-control/skills/<skill>/` (comprobar antes que `git diff --quiet v0.9.0 HEAD -- skills/` sale con 0). **Código**: internal/evals/laboratorio_normativa.go pasa a llamarse internal/evals/laboratorio_familias.go (`git mv`; research D10), y su tipo `familiaDelLaboratorio` gana el campo `producto` (vacío en la de H39); las tres familias de data-model §1, en este orden, como `familiasDelProducto`: `normativa` (`normativa-control` y `normativa-producto` en orden y herramienta, y las tres de H40 en herramienta; las ocho referencias de H40; evals `evals/normativa`, las 04, 05, 15 y 22; `conLaRegla` `normativa-testigo-en-referencia`), `plazos-y-territorio` (por ahora solo `plazos-y-territorio-control` y `plazos-y-territorio-producto`, en los dos modos: las tres formas llegan con sus directorios en T005; `administrativo.md`, `calculo_de_plazos.md`, `jerarquia_normativa.md`, `municipios.md`; `evals/plazos-y-territorio` con las 02, 05, 10 y 11 de data-model §1) y `jurisprudencia` (`jurisprudencia-control` y `jurisprudencia-producto` en los dos modos; `consulta_preparada.md`, `documento_aportado.md`; `evals/jurisprudencia` con las 01, 03, 04 y 06); las variantes versionadas de una familia son las suyas menos `producto`. `familiaDeNormativa` sigue nombrando la familia de `normativa` (la primera de `familiasDelProducto`), de modo que el arnés y la tabla de hoy dan lo mismo para las variantes de H40 sin tocar su código —`TestArnesDeNormativa`, `TestArnesDelLaboratorio`, `TestTablaDeNormativa` y `TestTablaDelLaboratorio` siguen en verde sin cambiar ningún valor esperado—; aceptar `<skill>-producto` en el arnés y en la tabla llega con T006 y T007. En .golangci.yml, `producto` entre las palabras que ignora misspell, con el comentario de la rama 033 (research V13). **Tests**: en internal/evals/laboratorio_familias_test.go, nuevo, `TestVariantesDeControl` (contracts/variantes.md §4), por familia del producto: `laboratorio/<skill>/` tiene un directorio por variante versionada y nada más, y ninguno es `<skill>-producto`; los ficheros de cada `<skill>-control` son, ruta a ruta, los de una tabla `ruta → sha256` escrita en el test al copiarlos (17: 9 de `normativa`, 5 de `plazos-y-territorio`, 3 de `jurisprudencia`; calculadas con `shasum -a 256` sobre las copias), y falla con la ruta del que difiere, falta o sobra; y los nombres de las variantes de las cuatro familias (H39 de `laboratorio/variantes.yaml` y las tres) no se repiten. En internal/evals/laboratorio_normativa_test.go, `TestVariantesDeNormativa` sin los subtests «directorios de la familia» y «control es skills/normativa», que cubre `TestVariantesDeControl`; los demás, sin cambiar ningún valor esperado (FR-021). En internal/evals/laboratorio_test.go, `TestLaboratorioFueraDelProducto` generalizado (contracts/variantes.md §4): los nombres que no pueden ser un tramo de una ruta de lo empotrado (`kitlegal.Skills()`), de lo que deja `kitlegal skills install --dir`, de una skill que ejecute el job ni de una carpeta de `evals/` son los de H39 y los de los directorios de variante de las carpetas de las tres familias, leídos del disco; ninguna ruta de `laboratorio/` tiene un tramo `<skill>-producto` ni mide (`rutaQueMide`); lo demás como hoy. Mutantes que se ven en rojo y se revierten: un byte cambiado en `laboratorio/jurisprudencia/jurisprudencia-control/skills/jurisprudencia/SKILL.md`; un fichero de más en `plazos-y-territorio-control`; una referencia borrada de `normativa-control`; un directorio `laboratorio/normativa/normativa-producto/`; un nombre de variante de una familia repetido en otra; y una copia de `plazos-y-territorio-control` bajo un nombre que el recorrido de lo empotrado ve. `TestSkillsDelRepositorio` y `make skills-check` siguen en verde con `skills/` sin cambios.

**Checkpoint**: las tres `<skill>-control` existen y no pueden cambiar sin que `make ci` lo diga; `skills/` aún no ha
cambiado.

---

## Phase 3: User Story 2 y User Story 1 — Lo que sale del cuerpo está en su referencia, y las remisiones son órdenes de leer (Priority: P1) 🎯 MVP

**Objetivo**: las siete remisiones de `normativa` y las tres de `jurisprudencia` reescritas, y los dos movimientos de
`normativa`, atados a sus `<skill>-control` por el cálculo desde control y a la regla y la lista de las remisiones por
un test del repositorio; `skills/plazos-y-territorio` atada byte a byte a su control.

**Independent Test**: `TestSkillsDelProductoDesdeControl`, `TestFrases`, `TestRemisionesDelRepositorio` y
`TestSkillsDelRepositorio` en verde con `make skills-check` y `make ci`, y en rojo con cada mutante (SC-001 a SC-005);
el prototipo de `h41-skills-instaladas.txtar` en verde.

- [X] T003 [US2] `normativa` manda leer sus referencias y lleva a `redaccion_modificada.md` y `lecturas_que_fallan.md` lo que solo usa su caso, con el control que calcula el producto desde control y el que ata `plazos-y-territorio` a su control (FR-004, FR-006, FR-007, FR-008, FR-009, FR-010; SC-002, SC-003, SC-004, SC-005; contracts/remisiones.md §1, §2, §3 y §8; data-model §5; research D1, D2, D4, D8, V2 a V5). **Skill**: skills/normativa/SKILL.md con los tramos del `diff -u` de contracts/remisiones.md §2 (R1 a R6, con la «o» de D1 en la orden de `redaccion_modificada.md`, y la salida de M-a y M-b); skills/normativa/references/redaccion_modificada.md y skills/normativa/references/lecturas_que_fallan.md con los tramos de §3 al final, detrás de una línea en blanco (M-a partida tras «la»). El cuerpo queda en 10 926 caracteres (`awk 'f; NR>1 && /^---$/ && !f {f=1}' skills/normativa/SKILL.md | LC_ALL=en_US.UTF-8 wc -m`; research V3); las cinco formas fijas siguen en sus bloques `text` y la frase del paso 2 y los pasos 3 y 6, como exige `h32-normativa.txtar`. **Test**: en internal/evals/laboratorio_familias_test.go, `TestSkillsDelProductoDesdeControl` (contracts/remisiones.md §8), con la declaración de data-model §5 (cambios `{antes, despues}` del cuerpo y movimientos `{texto, referencia, enLaReferencia}`) escrita con los bytes de §2 y §3 para `normativa` y vacía para `jurisprudencia`, que hoy es su control: subtests `normativa` y `jurisprudencia`, que leen control y `skills/<skill>` sin lo que genera `make skills-sync` (referencias cuya primera línea empieza por `<!-- generado desde `; regiones entre `<!-- paquetes: …` y `<!-- fin de los paquetes -->` y entre `<!-- inicio de la tabla de comandos…` y `<!-- fin de la tabla de comandos -->` reducidas a sus marcas), aplican cada cambio, cuyo `antes` tiene que estar una sola vez, añaden a cada referencia su `enLaReferencia`, y exigen el resultado ruta a ruta y byte a byte en `skills/<skill>`, fallando con la ruta del primer fichero que difiere, falta o sobra, o con el cambio cuyo `antes` ya no está una vez; por cada movimiento, texto y `enLaReferencia` con los blancos juntados son iguales, el texto no está en el `SKILL.md` de `skills/<skill>` y sí en su referencia; el frontmatter de `skills/<skill>/SKILL.md`, de la primera línea a la que lo cierra, es el de control byte a byte; y el subtest `plazos-y-territorio` (§8.4; research D4), con un comentario que dice que es así hasta H41.1: `skills/plazos-y-territorio` y `plazos-y-territorio-control` tienen los mismos ficheros ruta a ruta y byte a byte, **también lo generado**, y falla con la ruta del que difiere, falta o sobra. Mutantes que se ven en rojo y se revierten: un byte fuera de lo declarado en `skills/normativa/SKILL.md`; «Nada de otra conversación» devuelto al cuerpo; una letra de la `description` de `normativa` cambiada; un byte de más en `redaccion_modificada.md`; un `antes` que no está en control; un byte cambiado en `skills/plazos-y-territorio/references/municipios.md` y un fichero de más en `skills/plazos-y-territorio/references/` (SC-005); y, para `TestSkillsDelRepositorio` (SC-002, SC-003, de hoy), el cuerpo de `skills/normativa/SKILL.md` llevado a 11 001 caracteres, que da `cuerpo de 11001 caracteres (máximo 11000)`, y un byte cambiado dentro de la tabla de comandos generada de `normativa`, que da su deriva tras `make skills-sync`. `TestSkillsDelRepositorio`, `TestVerticalDeEnsayo` (FR-008) y las suites activas siguen en verde.
- [ ] T004 [US1] `jurisprudencia` manda leer sus referencias, y la frase, la regla y la lista de las remisiones como test del repositorio (FR-001, FR-002, FR-003, FR-005, FR-007, FR-008, FR-009, FR-010; SC-001, SC-002, SC-003, SC-004; contracts/remisiones.md §4 a §7; data-model §5 y §7; research D3, D5, D6, D7, V4, V10). **Skill**: skills/jurisprudencia/SKILL.md con el `diff -u` de contracts/remisiones.md §4 (J1 y J2; cuerpo de 10 842 caracteres); ninguna referencia suya cambia. **Frase**: internal/skills/frases.go, nuevo, `func Frases(cuerpo string) []string` con los segmentos de contracts/remisiones.md §6 (líneas de bloques de código, vallas incluidas, como `lineasFueraDeBloques`; encabezados y comentarios HTML enteros como segmentos solos; frases que abre una viñeta o un paso y que cierra un punto seguido de un blanco, sin que lo cierre el `.md` de un enlace; blancos juntados en un espacio), sin escribir nada y sin cambiar ningún otro fichero de `internal/skills`; y su test de tabla `TestFrases` en internal/skills/frases_test.go con los seis casos de §6. **Regla**: internal/skills/remisiones_test.go, nuevo, en el paquete y con `raizDelRepositorio = "../.."`, `TestRemisionesDelRepositorio` (contracts/remisiones.md §7) con `skills.Frases`: por cada skill de `skills/` con `references/`, en orden de nombre, salvo las de `fueraDeLasRemisiones = map[string]string{"plazos-y-territorio": "H41.1"}` (con el comentario que dice que el valor es el hito que la incluye; FR-010): (1) la regla, con los tres motivos y los textos exactos de §7.1, uno por referencia y motivo, en orden de nombre de las referencias, `<skill>: references/<f>: <motivo>`; (2) la lista, `remisionesDelPlan` con las frases de §5 (siete de `normativa`, tres de `jurisprudencia`) comparadas, enteras y en orden, con los segmentos con enlace fuera de la región de los paquetes, fallando con `<skill>: la remisión <n> no es la del plan` o `<skill>: no tiene lista de remisiones del plan`; (3) el subtest «control», la regla sobre el `SKILL.md` de `normativa-control` y `jurisprudencia-control` con exactamente los fallos de §7.3 (en `normativa`, dos en `con_herramientas.md` y el del verbo en `buscar_la_norma.md`, `como_se_cita.md`, `lecturas_que_fallan.md`, `normas.md`, `redaccion_en_una_fecha.md` y `varios_bloques.md`; en `jurisprudencia`, «antes de dar la consulta» en `consulta_preparada.md`). En el Makefile, solo la expresión `-run` de la receta de `skills-check`, que gana `TestRemisionesDelRepositorio` junto a `TestNormasDelRepositorio`; nada más de su receta ni de otra cambia. **Producto desde control**: en internal/evals/laboratorio_familias_test.go, la declaración de `jurisprudencia` de `TestSkillsDelProductoDesdeControl`, con los bytes de §4, en lugar de la vacía. Mutantes que se ven en rojo y se revierten: el «lee» de una remisión de `normativa` cambiado por «leer»; «si la tienes» añadido a la frase de un enlace de `normativa`; «Antes de dar la consulta, lee» al principio de la frase de `consulta_preparada.md`; el enlace a `normas.md` quitado del cuerpo; una palabra de una frase de la lista cambiada en `skills/jurisprudencia/SKILL.md` (la lista y `TestSkillsDelProductoDesdeControl`); `plazos-y-territorio` quitada de `fueraDeLasRemisiones`, que da los fallos de su control y el de la lista que no tiene; y una regla de `Frases` rota (el `.md` de un enlace que cierra la frase), que `TestFrases` ve. `make skills-check` corre el test nuevo y sale en verde. Antes de terminar, con un prototipo (worktree bajo `$TMPDIR` con specs/034-h41-las-skills-del/aceptacion/skills-instaladas.txtar copiado en la carpeta de guiones de internal/app como `h41-skills-instaladas.txtar`, `go -C <worktree> test -count=1 ./internal/app -run 'TestEntregaDelHito/^h41-skills-instaladas$'`), comprobar que el guion congelado pasa.

**Checkpoint**: `normativa` y `jurisprudencia` llevan sus órdenes de leer y lo movido; `plazos-y-territorio` sigue
siendo su control; `make skills-check` lo comprueba; es la entrega que recibe quien instala (historias 1, 2 y 3).

---

## Phase 4: User Story 3 — Las tres formas de `plazos-y-territorio` en el laboratorio (Priority: P1)

**Objetivo**: las tres variantes de `plazos-y-territorio` versionadas junto a su control, declaradas en su familia y
atadas a lo que declaran, sin tocar nada más que las órdenes de `municipios.md` y `calculo_de_plazos.md`.

**Independent Test**: `TestVariantesDePlazosYTerritorio`, `TestVariantesDeControl`, `TestVariantesDeNormativa` y
`TestLaboratorioFueraDelProducto` en verde, y en rojo con cada mutante (SC-009).

- [ ] T005 [US3] Las tres variantes de `plazos-y-territorio` y su control (FR-021, FR-023, FR-024, FR-025; SC-009; contracts/variantes.md §1 a §4; data-model §1 y §6; research D11, D12, D13, D15). **Laboratorio**: laboratorio/plazos-y-territorio/plazos-y-territorio-sin-condicion/skills/plazos-y-territorio/, laboratorio/plazos-y-territorio/plazos-y-territorio-condicion-por-orden/skills/plazos-y-territorio/ y laboratorio/plazos-y-territorio/plazos-y-territorio-lista-con-verbo/skills/plazos-y-territorio/, cada una copia de `plazos-y-territorio-control` (`cp -R`) con, en su `SKILL.md`, exactamente los tramos del `diff -u` de contracts/variantes.md §2.1, §2.2 y §2.3, y sus cuatro referencias byte a byte las de control. Sus cuerpos, medidos con la orden de research V3, son 10 991, 11 245 y 10 938 caracteres (contracts/variantes.md §3); si uno no da su cifra, el `SKILL.md` no es el del contrato. El tope de 11 000 no se les aplica y ninguna comprobación sale en rojo por encima de 10 930 (FR-024); el run no aplica M1, M2 ni M3. **Código**: en internal/evals/laboratorio_familias.go, la familia `plazos-y-territorio` gana las tres variantes, en los dos modos y en ese orden tras la del producto (data-model §1). **Tests**: en internal/evals/laboratorio_familias_test.go, `TestVariantesDePlazosYTerritorio` (contracts/variantes.md §4) con `declaracionesDePlazosYTerritorio` (data-model §6: el tipo `declaracionDeLaVariante` de H40, solo con `sustituciones` `{fichero, antes, despues, hito}`, en el orden de las variantes), por variante: (1) el `despues` de cada sustitución, normalizado con `textoDelHito`, es el texto de §2; (2) la declaración aplicada sobre `plazos-y-territorio-control` da, ruta a ruta y byte a byte, el directorio de la variante, y falla con la ruta del primer fichero que difiere, falta o sobra; (3) cada sustitución es del `SKILL.md`, y, quitados de los dos cuerpos los segmentos de `skills.Frases` con un enlace a `references/municipios.md` o `references/calculo_de_plazos.md`, el de la variante y el de control tienen los mismos segmentos en el mismo orden, y falla con el primero que difiere. En internal/evals/laboratorio_normativa_test.go, los ayudantes de H40 (`comprobarLaVarianteDeclarada`, `calcularLaVariante` y los que los usan) pasan a recibir la familia, sin cambiar ningún valor esperado de `TestVariantesDeNormativa` (FR-021). `TestVariantesDeControl` ve los cinco directorios de la familia y `TestLaboratorioFueraDelProducto` sus nombres, sin cambiar su código. Mutantes que se ven en rojo y se revierten: un byte cambiado en `references/calculo_de_plazos.md` de `plazos-y-territorio-sin-condicion`; «Si no lo dice, **pregúntalo** y espera» cambiado en la variante y en su declaración a la vez (lo ve el punto 3); la línea de `jerarquia_normativa.md` cambiada en `plazos-y-territorio-lista-con-verbo`; un `despues` declarado que no es el texto de §2; un directorio de variante de la familia borrado (`TestVariantesDeControl`); y una copia de `plazos-y-territorio-condicion-por-orden` bajo un nombre que el recorrido de lo empotrado ve (`TestLaboratorioFueraDelProducto`). `TestSkillsDelProductoDesdeControl` y `make skills-check` siguen en verde: `skills/plazos-y-territorio` no cambia.

**Checkpoint**: las cinco variantes de `plazos-y-territorio` —control, producto y las tres formas— están declaradas; las
tres versionadas no pueden cambiar fuera de lo que declaran sin que `make ci` lo diga.

---

## Phase 5: User Story 4 — El arnés mide `control`, `producto` y las variantes de las tres familias (Priority: P1)

**Objetivo**: `make evals-laboratorio` acepta una familia del producto, instala `<skill>-producto` desde `skills/` y deja
la lista de sus referencias en `sondeo.json`, con las diecinueve filas de FR-033 en cuatro invocaciones.

**Independent Test**: `TestArnesDeLasFamilias`, sin modelo y con el sustituto de `claude`: 0 sesiones con cada petición
inválida y n/4 por eval en cada variante y modo con las válidas, cada una con la skill de su variante (SC-007, SC-008).

- [ ] T006 [US4] El arnés con las cuatro familias y su control sin modelo (FR-022, FR-030, FR-032, FR-033; SC-007, SC-008; contracts/arnes.md §2 a §5; data-model §1, §3 y §4; research D14, D15, V9). **Código**: internal/evals/laboratorio_sondeo.go, internal/evals/laboratorio_tabla.go (solo el tipo `sondeoLeido`) y, si lo pide, internal/evals/laboratorio_familias.go: las familias de la petición son la de H39 y las tres del producto, en ese orden; los mensajes de contracts/arnes.md §2 (el de la mezcla, una por familia pedida en el orden de las familias, con la de H39 nombrada «del laboratorio de H39», de modo que `VARIANTES=control,normativa-control` da byte a byte el de H40; los demás, los de H39 y H40), todos juntos y sin abrir sesiones, construir nada ni crear `DIRECTORIO`; de una variante versionada, la copia de `laboratorio/<skill>/<variante>/skills/` como en H40; de `<skill>-producto`, la copia de `skills/<skill>` del árbol en `<temporal>/skills/<skill>-producto/<skill>/` con `os.CopyFS`; el directorio de evals del reparto `evals/<skill>` y el plan de `SESIONES/4` sesiones de cada una de las cuatro evals de la familia, en su orden, de nombre `<eval sin .yaml>-<nn>`; y `sondeo.json` con `referencias_del_producto` (los nombres de `references/` de la copia, en orden de nombre) solo si se pide `<skill>-producto`, y sin ella el de hoy byte a byte. La clave es el campo `ReferenciasDelProducto []string` con la etiqueta `json:"referencias_del_producto,omitempty"`, tercero tras `Modelo` y `ClaudeCode`, en `sondeoLeido`, el tipo con el que el arnés escribe hoy `sondeo.json` y con el que la tabla lo lee: un solo tipo para el fichero (data-model §4); su comentario de tipo nombra la clave nueva. **Tests**: en internal/evals/laboratorio_sondeo_test.go, `TestArnesDeLasFamilias` (contracts/arnes.md §5) con la preparación sin construir y el sustituto de `claude` de `TestArnesDelLaboratorio`, que anota sus invocaciones: con cada petición inválida (`plazos-y-territorio-otra`; `normativa-remisiones-en-orden` con `MODOS=orden`; `normativa-producto` y `jurisprudencia-control`; `plazos-y-territorio-control` y `control`; `SESIONES=6`; `SESIONES=0`), 0 sesiones, nada construido, `DIRECTORIO` sin crear y su mensaje de §2; con las cuatro invocaciones de FR-033 (`VARIANTES=normativa-control,normativa-producto MODOS=orden,herramienta`; `VARIANTES=normativa-remisiones-en-orden MODOS=herramienta`; las cinco de `plazos-y-territorio` con `MODOS=orden,herramienta`; `VARIANTES=jurisprudencia-control,jurisprudencia-producto MODOS=orden,herramienta`) y `SESIONES=4`, una sesión de cada una de las cuatro evals de la familia en cada variante y modo pedidos, en `DIRECTORIO/<variante>/<modo>/<eval>-01/`, y ninguna más (16, 4, 40 y 16); con `VARIANTES=normativa-remisiones-en-orden MODOS=herramienta SESIONES=8`, dos de cada eval; en cada sesión, `claude/skills/` con solo la skill de la familia, ruta a ruta y byte a byte la de `laboratorio/<skill>/<variante>/skills/<skill>/` en una variante versionada o la de `skills/<skill>` en `<skill>-producto` (en `plazos-y-territorio-producto`, también la de `plazos-y-territorio-control`), sin `servidor.json` en modo orden y con `kitlegal mcp serve` en herramienta; y `DIRECTORIO/sondeo.json` con `referencias_del_producto` igual a los nombres de `skills/<skill>/references` en orden de nombre si se pidió `<skill>-producto`, y sin la clave si no. `TestArnesDeNormativa`, `TestArnesDelLaboratorio`, `TestRecetasDelSondeo` y `TestServidorDelLaboratorio`, sin cambiar ningún valor esperado. Mutantes que se ven en rojo y se revierten: la copia de `<skill>-producto` tomada de `<skill>-control`; la clave `referencias_del_producto` escrita siempre; la mezcla de dos familias del producto aceptada; `SESIONES=0` aceptado; una de las tres formas de `plazos-y-territorio` sin modo orden; y una eval de la familia quitada del plan.

**Checkpoint**: el sondeo de las diecinueve filas de FR-033 se puede lanzar con las órdenes de quickstart.md §9, fuera
del run.

---

## Phase 6: User Story 5 — La tabla cuenta la lectura con `cd` al directorio de las referencias (Priority: P1)

**Objetivo**: `make laboratorio-tabla` da la tabla de cualquiera de las tres familias del producto, con las referencias
de lo medido y la lectura con `cd`, sin cambiar las de H39 y H40.

**Independent Test**: `TestLecturaConCd`, `TestTablaDeLasFamilias`, `TestTablaDeNormativa` y `TestTablaDelLaboratorio`
sin cambios, y el guion congelado `h41-laboratorio-tabla.txtar` en verde (SC-006, SC-009).

- [ ] T007 [US5] La tabla de una familia del producto y la lectura con `cd` (FR-032, FR-040, FR-041, FR-042; SC-006, SC-009; contracts/tabla.md §1 a §3 y §5; data-model §8; research D16, D17). **Código**: internal/evals/laboratorio_tabla.go (y internal/evals/laboratorio_familias.go si lo pide): la familia es la de la primera variante que lee en orden de nombre, una de las tres del producto si es suya y si no la de H39 como hoy; en la de una familia del producto, un directorio de variante de otra familia, también de H39, da `laboratorio: <ruta> no es ninguna variante de <skill>`, código 1 y sin tabla; la tabla usa el campo `ReferenciasDelProducto` de `sondeoLeido` que añade T006, sin tocar el tipo y sin ninguna comprobación ni mensaje propios para un `sondeo.json` sin la lista (research D17, D20: lo que no sabe leer es el defecto `inesperado` de hoy); las referencias de la tabla son la unión, en orden de nombre, de las de la familia por cada directorio de variante versionada que lee y de la lista de cada sondeo con un directorio `<skill>-producto`, y la celda de una referencia que la variante de la fila no tiene va vacía, en la fila y en el detalle; filas en el orden de la declaración de la familia, `orden` antes que `herramienta`; «Cumplen la regla» y «Cumple», solo en la de `normativa`; `cita_preparar`, `cita_cotejar`, `territorio_resolver`, `plazos_calcular` y las de `boe` cuentan como primera herramienta como en H40; y, solo en las familias del producto, la lectura con `cd` de contracts/tabla.md §2 sobre las órdenes simples de hoy (`ordenesSimples`): una orden simple cuya primera palabra que no es una asignación es `cd` y cuya siguiente, sin barra final, es `<skill>/references` o termina en `/<skill>/references`, y en una orden simple posterior de la misma orden una palabra que es exactamente `<f>` o `./<f>`; la regla de H40 que cuenta `references/<f>` sigue aparte. **Test**: en internal/evals/laboratorio_tabla_test.go, `TestLecturaConCd` (contracts/tabla.md §5), tabla de casos con `t.Parallel()`: las trece órdenes de la tabla de contracts/tabla.md §2, orden a orden, con la familia de `normativa` (incluido el `cd` en un `Bash` y `cat normas.md` en otro, que no abre), y la familia de H39 con `cd …/articulos-del-boe/references && cat cierre.md`, que no la abre. `TestTablaDeNormativa`, `TestTablaDelLaboratorio`, `TestReglaTestigo` y `TestOrdenesDeBash`, sin cambiar ningún valor esperado. Mutantes que se ven en rojo y se revierten: la lectura con `cd` aplicada en la familia de H39; el `cd` aceptado en otra orden de `Bash`; `cat otra/normas.md` aceptado; `cat normas.md.bak` aceptado; y `<skill>` de otra familia aceptado en la ruta del `cd`. Antes de terminar, con un prototipo (worktree bajo `$TMPDIR` con specs/034-h41-las-skills-del/aceptacion/laboratorio-tabla.txtar copiado en la carpeta de guiones de internal/app como `h41-laboratorio-tabla.txtar`, `go -C <worktree> test -count=1 ./internal/app -run 'TestEntregaDelHito/^h41-laboratorio-tabla$'`), comprobar que el guion congelado pasa, y que las suites activas de H39 y H40 siguen en verde (`go test -count=1 -run '^TestEntregaDelHito$/^h(39|40)-' ./internal/app/`; FR-042).
- [ ] T008 [datos] [US5] Las sesiones fijas de las tres familias y su tabla esperada, con el test que las lee, en una sola tarea porque el control de SC-006 compara la salida del guion de la tabla, ya escrito en T007, byte a byte con esas tablas y no hay test sin los fixtures (FR-032, FR-040, FR-041; SC-006, SC-007; contracts/tabla.md §2, §3 y §5; data-model §4 y §8). **Fixtures**: internal/evals/testdata/laboratorio/familias/normativa/, internal/evals/testdata/laboratorio/familias/plazos-y-territorio/ e internal/evals/testdata/laboratorio/familias/jurisprudencia/, cada una con `sondeo/` (sondeos con su `sondeo.json` y su lista `referencias_del_producto` donde hay `<skill>-producto`; sesiones fijas con la forma de las de H40, en los dos modos; en `plazos-y-territorio`, de las cinco variantes de la familia; lecturas con `cd` que cuentan y que no; en `normativa`, una referencia de la lista del producto que la familia no declara; una sesión sin terminar, código 124) y `tabla.md`, la salida esperada byte a byte, con los recuentos, turnos y extremos de Wilson de H40, celdas vacías donde la variante no tiene la referencia y «Cumplen la regla» solo en la de `normativa`; e internal/evals/testdata/laboratorio/familias/mezcla/sondeo/h41/, sin `tabla.md`: un sondeo de `jurisprudencia` con un directorio de `normativa` (plan.md «Gates»): su `sondeo.json` sin lista, `jurisprudencia-control/orden/01-existe-con-numero-y-fecha-01`, una sesión terminada que es la primera variante en orden de nombre y hace de `jurisprudencia` la familia de la tabla, y `normativa-control/orden/04-lgt-prescripcion-01`, que es de otra familia. Ningún fichero de `testdata/` existente cambia; nada se graba. **Test**: en internal/evals/laboratorio_tabla_test.go, `TestTablaDeLasFamilias` (contracts/tabla.md §5), con `t.Parallel()`: por cada una de las tres familias, nombradas en el test y no sacadas de leer `familias/`, la salida de la tabla sobre `familias/<skill>/sondeo/` es su `tabla.md` byte a byte, con stderr vacío y código 0; y sobre `familias/mezcla/sondeo/`, código 1, salida estándar vacía y `laboratorio: <ruta>/h41/normativa-control no es ninguna variante de jurisprudencia` en la salida de error. Mutantes que se ven en rojo y se revierten: un turno cambiado en una sesión de `plazos-y-territorio-condicion-por-orden`; la orden de una lectura con `cd` de `jurisprudencia` partida en dos `Bash`; la referencia de la lista del producto quitada de `sondeo.json` de `normativa`; un extremo de Wilson cambiado en un `tabla.md`; y el directorio `normativa-control` de `familias/mezcla/sondeo/h41/` renombrado a `jurisprudencia-producto`.

**Checkpoint**: la medida de la aceptación se puede contar sin modelo; las tablas de H39 y H40 no cambian.

---

## Phase 7: Polish & cross-cutting

- [ ] T009 La documentación de lo que entra en el repositorio y la línea de lo que queda para antes de fusionar (FR-020, FR-030, FR-040, FR-050, FR-051, FR-052, FR-053; DoD §1; research D19; plan.md «Cierre y pendientes»): CHANGELOG.md, en *Unreleased*, el cambio de `normativa` y `jurisprudencia` (cada remisión, una orden de leer en su sitio; «Nada de otra conversación» a `redaccion_modificada.md` y «Si una lectura falla…» a `lecturas_que_fallan.md`), las tres formas de `plazos-y-territorio` en el laboratorio, el arnés con las tres familias del producto y la tabla con la lectura con `cd`; docs/SKILLS.md, en §1 tras «Cómo se remite», lo de research D19 (la orden cuya condición la trae lo que devuelve una llamada va en el paso de esa llamada, detrás de ella; el sitio para las órdenes se hace llevando a la referencia que se manda leer en ese caso lo que solo ese caso necesita, con el límite del ADR 0052, punto 4; `make ci` y `make skills-check` comprueban la frase de cada remisión), sin el punto 2 del ADR 0052 (FR-050); CONTRIBUTING.md, en la tabla de objetivos y en «El sondeo del laboratorio», las familias del producto, `<skill>-control` congelada en `v0.9.0` (ya no es `skills/normativa`), `<skill>-producto` instalada desde `skills/`, las tres formas de `plazos-y-territorio` y las cuatro invocaciones del sondeo de la aceptación; en el Makefile, solo las líneas `##` de ayuda de `evals-laboratorio` y `laboratorio-tabla`, que nombran las cuatro familias (las recetas no cambian: `TestRecetasDelSondeo` sigue en verde); el comentario de paquete de cmd/laboratorio/main.go y la cabecera de comentario de scripts/evals-laboratorio.sh, las cuatro familias, sin cambiar código ni órdenes; y specs/034-h41-las-skills-del/gates/supuestos.md, una línea `- [alcance] tasks: FR-052 y FR-053 …`: el run no lanza el sondeo de las diecinueve filas ni escribe los ADR de FR-052 y FR-053; la fusión queda pendiente de las tres tablas con el modelo y la versión anotados y de su lectura, fuera del run, y si `normativa` o `jurisprudencia` no cumplen, lo medido va a un ADR que dice qué medir después; FR-051 lo mide el job de evals del cierre con los umbrales de hoy. Ninguna fila nueva en docs/SOURCES.md (no hay fuente nueva). `make ci` en verde.

---

## User Story 6 — La persona mide y decide antes de fusionar (P2)

Sin tareas: FR-052, FR-053, SC-011 y SC-012 se hacen fuera del run (ADR 0032) con lo que entregan T002 a T008 y las
órdenes de quickstart.md §9; el informe final lo deja pendiente por la línea `[alcance]` que escribe T009 en
`gates/supuestos.md`.

---

## Dependencies & Execution Order

### Orden de las fases

- **Fase 1 (T001)**: primero; congela la suite.
- **Fase 2 (T002)**: las `<skill>-control` nuevas se copian antes de cambiar ninguna skill, y `normativa-control` deja
  de compararse con `skills/normativa` antes de que T003 la cambie.
- **Fase 3 (T003 → T004)**: T003 crea `TestSkillsDelProductoDesdeControl` sobre las `<skill>-control` de T002; T004
  completa su declaración, crea `skills.Frases` y `TestRemisionesDelRepositorio`, que necesita las dos skills reescritas
  (la de `normativa`, de T003) y las `<skill>-control` (subtest «control»). Tras T004, `h41-skills-instaladas.txtar`
  pasa.
- **Fase 4 (T005)**: necesita las familias (T002) y `skills.Frases` (T004) para el punto 3 de su control.
- **Fase 5 (T006)**: necesita las familias con `producto` (T002) y las cinco variantes de `plazos-y-territorio` (T005)
  para la tercera invocación de FR-033; copia `skills/<skill>` tal como esté.
- **Fase 6 (T007 → T008)**: T007 necesita las familias (T002, T005) y lee la clave que escribe T006; T008 fija con
  fixtures lo que cuenta el código de T007. Tras T007, `h41-laboratorio-tabla.txtar` pasa.
- **Fase 7 (T009)**: al final; documenta lo de T002 a T008.

### Dentro de cada tarea

Test y la implementación mínima que lo pone en verde, en la misma tarea (T008, el test sobre el código de T007, con sus
fixtures); el mutante que ve fallar el control, revertido; `make ci` en verde antes de marcarla.

### Oportunidades de paralelismo

Ninguna tarea lleva [P]: T002 a T008 comparten `internal/evals/laboratorio_familias*.go` o dependen de la anterior
(T005 de `skills.Frases` de T004; T006 de las variantes de T005; T007 de la clave de `sondeo.json` de T006; T008 del
código de T007), y T004 y T009 comparten el `Makefile`. Se ejecutan en serie, en el orden de los ids.

## Parallel Example

```text
Sin tareas en paralelo: el bucle las ejecuta una a una, T001 → T009.
```

## Implementation Strategy

### MVP

1. Fase 1: la suite.
2. Fase 2: las familias y las `<skill>-control` (T002). **Validar**: `TestVariantesDeControl`.
3. Fase 3: las dos skills (T003, T004). **Validar**: `make skills-check` y el prototipo de
   `h41-skills-instaladas.txtar`. Es lo que recibe quien instala.

### Entrega incremental

4. Fase 4: las tres formas de `plazos-y-territorio` (T005). **Validar**: `TestVariantesDePlazosYTerritorio`.
5. Fase 5: el arnés (T006). **Validar**: `TestArnesDeLasFamilias`.
6. Fase 6: la tabla (T007, T008). **Validar**: `TestLecturaConCd`, `TestTablaDeLasFamilias` y el prototipo de
   `h41-laboratorio-tabla.txtar`.
7. Fase 7: CHANGELOG, `docs/SKILLS.md`, CONTRIBUTING, ayuda del Makefile, comentarios y la línea de supuestos.

## Notes

- 9 tareas: T001 [aceptacion]; US1: T004; US2: T003; US3: T002, T005; US4: T006; US5: T007, T008; transversal: T009;
  US6: ninguna.
- Tarea [datos]: T008 (`internal/evals/testdata/laboratorio/familias/`), que dice en su línea por qué va con su test.
  El hito no tiene datos externos (plan.md «Datos externos»): ni manifiesto `grabaciones.json`, ni `TestGrabar*`, ni
  fila nueva en `docs/SOURCES.md`.
- Definition of Done: `make ci` y tests offline en cada tarea; R4 y R5 (convenciones); ningún sobre ni esquema cambia
  (sin tarea de `schemas/`); defecto `inesperado` y errores de uso (T006, T007); e2e en T001; CHANGELOG en T009; ADR: no
  aplica en el run, los de FR-052 y FR-053 se escriben fuera del run; `docs/SOURCES.md`: no aplica (ninguna fuente);
  skills con evals: las de hoy, que mide el job del cierre (FR-051), sin evals nuevas (Fuera de alcance); fixtures en
  T008.
- Cada fila de «Controles de umbral» de plan.md y la tarea que construye su control (y lo ve fallar con un mutante):
  SC-001/FR-001 a FR-003 (`TestRemisionesDelRepositorio`), T004; SC-002/FR-008 (`TestSkillsDelRepositorio`, de hoy),
  T003, con el mutante de 11 001 caracteres; SC-003/FR-009 (`TestSkillsDelProductoDesdeControl` y
  `TestSkillsDelRepositorio`), T003 (frontmatter y deriva de la tabla de comandos) y T004; SC-004/FR-006/FR-007
  (`TestSkillsDelProductoDesdeControl`), T003 y T004; SC-005/FR-010 (`TestSkillsDelProductoDesdeControl`, subtest
  `plazos-y-territorio`), T003; SC-001/SC-004/SC-005/SC-009/FR-001 a FR-006/FR-010/FR-025 (`h41-skills-instaladas.txtar`
  en `TestEntregaDelHito`), T001, que lo ve fallar hoy, y T004, que lo pone en verde; SC-006/FR-040/FR-041
  (`TestTablaDeLasFamilias`, `TestLecturaConCd` y `h41-laboratorio-tabla.txtar`), T007 (`TestLecturaConCd`, y el guion
  que T001 ve fallar) y T008 (`TestTablaDeLasFamilias`); SC-007/FR-022/FR-032 (`TestArnesDeLasFamilias`,
  `TestTablaDeLasFamilias`, `TestVariantesDeControl`), T006, T008 y T002; SC-008/FR-030/FR-033
  (`TestArnesDeLasFamilias`), T006; SC-009/FR-020 (`TestVariantesDeControl`), T002; SC-009/FR-021
  (`TestVariantesDeNormativa`, sin cambiar sus valores), T002 y T005; SC-009/FR-023
  (`TestVariantesDePlazosYTerritorio`), T005; SC-009/FR-025 (`TestLaboratorioFueraDelProducto`), T002 y T005;
  SC-009/FR-042 (las suites congeladas de H39 y H40 en `TestEntregaDelHito`), todas, comprobado en T007; SC-010/FR-051,
  las tres filas `evals:`, las mide el job de evals del cierre, que hace el workflow tras la revisión final con los
  umbrales de hoy, sin tarea que los toque.
- Requisitos sin tarea propia: FR-024 (no es un umbral: lo cumplen los cuerpos de T005 y lo da el plan, contracts/variantes.md
  §3); FR-031 (todas: ninguna graba, abre sesiones con modelo ni escribe fuera del repositorio y `$TMPDIR`); FR-051 y
  SC-010 (el cierre del workflow; la línea de T009); FR-052, FR-053, SC-011 y SC-012 (fuera del run; la línea
  `[alcance]` de T009).
