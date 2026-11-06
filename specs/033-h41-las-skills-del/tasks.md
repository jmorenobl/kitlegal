---
description: "Tareas de H41 · Las skills del producto mandan leer sus referencias"
---

# Tasks: H41 · Las skills del producto mandan leer sus referencias

**Input**: `specs/033-h41-las-skills-del/` — [spec.md](spec.md), [plan.md](plan.md), [research.md](research.md),
[data-model.md](data-model.md), [quickstart.md](quickstart.md) y [contracts/](contracts/) (`remisiones.md`,
`variantes.md`, `arnes.md`, `tabla.md`, `aceptacion.md`).

**Tests**: obligatorios (constitución III). La primera tarea escribe la suite de aceptación, que el workflow congela;
cada tarea posterior trae su test y la implementación que lo pone en verde, y deja `make ci` en verde por sí sola.

**Organización**: una fase por historia, en el orden en que se ejecutan una tras otra. La línea de cada tarea declara
todas las rutas que crea o modifica (el guardián rechaza cualquier otra; `go.mod`, `go.sum`, `CHANGELOG.md` y el
directorio del feature siempre se admiten, y declarar `x.go` admite `x_test.go`).

## Format: `[ID] [P?] [Story] Descripción`

- **[P]**: puede ir en paralelo (ficheros distintos, sin dependencia de una tarea sin terminar).
- **[Story]**: la historia de [spec.md](spec.md) a la que sirve (US1 a US4; US5 no tiene tareas).
- **[aceptacion]**: la suite congelada (solo T001). **[datos]**: la tarea toca ficheros bajo `testdata/` (solo T007).

## Convenciones de todas las tareas

- `make ci` en verde al terminar cada una, en primer plano y por objetivos si hace falta (`make fmt-check`,
  `make lint`, `make test`, `make test-integration`, `make vuln`, `make schema-check`, `make skills-check`…). Ninguna
  usa la red más que la de las herramientas de Go, ninguna graba (no hay datos externos: plan.md «Datos externos»),
  ninguna escribe en `evidencias/` ni en `data/`, ninguna abre una sesión con modelo ni ejecuta
  `TestSondeoDelLaboratorio`, `make evals-laboratorio`, `make evals` ni los guiones `scripts/evals*.sh` desde la terminal
  (FR-031, FR-052).
- Los cambios de las skills son, byte a byte, los `diff -u` de [contracts/remisiones.md](contracts/remisiones.md) §2 a
  §5; nada más de `skills/` cambia (FR-007): ni el frontmatter ni la `description` (FR-010), ni la tabla de comandos, ni
  el bloque de los paquetes, ni ninguna referencia salvo `skills/normativa/references/redaccion_modificada.md` (§3). No
  se tocan `skills/consulta-juridica/`, `evals/`, `data/`, `internal/mcp/`, el generador de `internal/skills`
  (`sincronia.go`, `regiones.go`, `markdown.go`…), `laboratorio/variantes.yaml`, `laboratorio/variantes/`,
  `laboratorio/evals/`, las variantes de H40 (`laboratorio/normativa/`), las recetas del `Makefile` ni ningún fichero de
  `testdata/` que ya exista (Fuera de alcance; FR-021, FR-042). Ningún argumento, bandera ni orden nueva de `kitlegal`
  ni de `cmd/laboratorio`; ningún umbral cambia.
- Los mensajes del arnés son los de [contracts/arnes.md](contracts/arnes.md) §2; la regla de las remisiones, la de
  [contracts/remisiones.md](contracts/remisiones.md) §6; el cálculo desde control, el de §7; la tabla y la lectura con
  `cd`, las de [contracts/tabla.md](contracts/tabla.md) §1 a §3. Nada se inventa fuera de ellos.
- `internal/evals` no nombra `os.Stdin`, `os.Stdout`, `os.Stderr` ni llama a `os.Exit` (R4, R5); ningún `panic`; lo que
  llega de fuera manipulado es defecto `inesperado` (1), sin mensaje ni test propios. Ningún `//nolint`. Tests de tabla
  con `t.Parallel()`.
- Cada control se ve fallar con un mutante (la medida pasada del umbral) antes de darlo por bueno, y el mutante se
  revierte (ADR 0029). Los guiones de un solo uso, si los hay, van bajo `$TMPDIR` y no se versionan.
- Las suites congeladas ya activas (las de H39, H40 y `h32-normativa.txtar`) siguen en verde sin cambiar un fichero suyo
  (FR-042; research V4). Los fixtures nuevos no llevan nada con forma de secreto (gitleaks corre en `make ci`).

---

## Phase 1: Setup — la suite de aceptación

**Propósito**: fijar, antes de cualquier código, lo que instala el binario y lo que la tabla tiene que dar con las
sesiones de las tres familias (ADR 0018; plan.md «Aceptación e2e»).

- [X] T001 [aceptacion] Escribir, a partir del spec y de [contracts/aceptacion.md](contracts/aceptacion.md), sin código de producto, los dos guiones testscript specs/033-h41-las-skills-del/aceptacion/laboratorio-tabla.txtar y specs/033-h41-las-skills-del/aceptacion/remisiones.txtar (el workflow los activa como `h41-laboratorio-tabla.txtar` y `h41-remisiones.txtar` al terminar el bucle, y los ejecuta `TestEntregaDelHito`; el arnés e2e no cambia: `kitlegal` en el `PATH` y `KITLEGAL_LABORATORIO_BIN`, de H40). Cada guion abre con un comentario que dice lo que comprueba y una línea «Cubre:» con sus historias y requisitos, sin los textos `usage: `, `unknown command`, `unexpected command` ni `cannot parse`. **laboratorio-tabla.txtar** (US4; «Cubre: US4.1, US4.2, US4.4 · FR-040, FR-041 · SC-005»): las órdenes de contracts/aceptacion.md §1, en su orden y byte a byte (cada `exec sh -c '… || true'` delante de su orden con código; al final `! exec $KITLEGAL_LABORATORIO_BIN tabla mezcla`, `! stdout .` y la expresión de `stderr` de §1); ficheros del archivo: los sondeos `normativa/h41/`, `plazos-y-territorio/h41/`, `jurisprudencia/h41/` y `mezcla/h41/`, cada uno con su `sondeo.json` en una línea (`{"modelo":"claude-sonnet-5-5","claude_code":"2.1.317"}` y, donde §1 lo dice, `referencias_del_producto`: en `normativa`, las ocho de la familia y `nueva.md`, en orden de nombre; en `plazos-y-territorio`, sus cuatro; en `jurisprudencia`, sus dos; en `mezcla`, sin lista) y las sesiones de las tablas de §1 (`codigo-de-la-sesion`, `sesion.err` y `sesion.jsonl` con la forma de las de H40: `system/init` con el modelo y la versión, un `assistant` con su `message.id` por respuesta, los resultados en mensajes `user` y `result`; las rutas de las skills bajo `/s/claude/skills/`; los turnos y las órdenes de `Bash` exactamente los de §1); y `normativa-tabla.md`, `plazos-y-territorio-tabla.md` y `jurisprudencia-tabla.md`, la tabla y el detalle de contracts/tabla.md §2 y §3, byte a byte, con la cabecera `# Sondeo de las variantes de <skill>`, los extremos de Wilson al 95 % como en H40, la columna `nueva.md` vacía en las filas y el detalle de `normativa-control` y `normativa-remisiones-en-orden` y a «0 de 1» en las de `normativa-producto`, «Cumplen la regla» solo en la de `normativa`, y la de `plazos-y-territorio` igual al ejemplo de contracts/tabla.md §3. **remisiones.txtar** (US1 y US2; «Cubre: US1.1, US1.2, US1.3, US2.1 · FR-001 a FR-006, FR-009 · SC-001, SC-004; el tope de 11 000, la regla sobre cada frase y el cálculo desde control los comprueban TestSkillsDelRepositorio, TestRemisionesDelRepositorio y TestSkillsDelProductoDesdeControl»): `exec kitlegal skills install --dir destino --json` y, sobre lo instalado, los `grep` y `! grep` de contracts/aceptacion.md §2, en su orden y byte a byte. Antes de terminar, con un prototipo (un worktree bajo `$TMPDIR` con los dos guiones copiados en la carpeta de guiones de internal/app con su nombre `h41-…` y ejecutado con `go -C <worktree> test -count=1 ./internal/app -run 'TestEntregaDelHito'`), comprobar que cada guion falla hoy por una aserción y no por un `exec` que no puede ejecutarse: `h41-laboratorio-tabla` en `cmp normativa.out normativa-tabla.md` (hoy `laboratorio tabla normativa` sale con 1, `…/normativa-producto no es ninguna variante de normativa`) y `h41-remisiones` en el primer `grep` (la orden del «Protocolo» no está en el `SKILL.md` de hoy); y que, con el árbol de hoy, las demás suites activas siguen en verde. No escribe evals: el spec deja fuera las evals nuevas y los cambios de las que existen (Fuera de alcance; FR-051 usa las de hoy)

**Checkpoint**: la suite queda congelada; ninguna tarea posterior la toca.

---

## Phase 2: User Story 3 (base) — Las familias del producto y sus `<skill>-control` congeladas (Priority: P1)

**Objetivo**: las tres familias declaradas en Go con su campo `producto`, y las dos `<skill>-control` nuevas copiadas
de `skills/` antes de cambiar ninguna skill, comprobadas por sus huellas. Va primero porque las tareas de US1 y US2
miden las skills contra estas copias (contracts/remisiones.md §6.4 y §7) y `normativa-control` deja de compararse con
`skills/normativa` (FR-020).

**Independent Test**: `TestVariantesDeControl`, `TestVariantesDeNormativa` y `TestLaboratorioFueraDelProducto` en verde,
y en rojo con cada mutante de sus controles (SC-006, SC-008).

- [X] T002 [US3] Las tres familias del producto, las dos `<skill>-control` nuevas y sus controles (FR-020, FR-021, FR-022, FR-023; SC-006, SC-008; contracts/variantes.md §1 y §2; data-model §1 y §2; research D8, D9, V1). **Laboratorio**, lo primero de la tarea y antes de cambiar ninguna skill: laboratorio/plazos-y-territorio/plazos-y-territorio-control/skills/plazos-y-territorio/ (5 ficheros) y laboratorio/jurisprudencia/jurisprudencia-control/skills/jurisprudencia/ (3 ficheros), con `cp -R skills/<skill>/. laboratorio/<skill>/<skill>-control/skills/<skill>/` (hoy `skills/` es `v0.9.0`: `git diff v0.9.0 -- skills/` vacío, comprobarlo antes de copiar). **Código**: internal/evals/laboratorio_normativa.go pasa a llamarse internal/evals/laboratorio_familias.go (`git mv`; research D9), y su tipo `familiaDelLaboratorio` gana el campo `producto` (vacío en la de H39); las tres familias de data-model §1, en este orden, como `familiasDelProducto`: `normativa` (variantes `normativa-control` y `normativa-producto` en orden y herramienta, y las tres de H40 en herramienta; las ocho referencias de H40; evals `evals/normativa`, las 04, 05, 15 y 22 de H40; `conLaRegla` `normativa-testigo-en-referencia`), `plazos-y-territorio` (`plazos-y-territorio-control` y `plazos-y-territorio-producto` en los dos modos; `administrativo.md`, `calculo_de_plazos.md`, `jerarquia_normativa.md`, `municipios.md`; `evals/plazos-y-territorio` con `02-territorio-municipio-no-cubierto.yaml`, `05-plazo-reposicion-leganes.yaml`, `10-plazo-reposicion-anterior-a-la-lpac.yaml`, `11-plazo-reposicion-sin-leer-el-precepto.yaml`) y `jurisprudencia` (`jurisprudencia-control` y `jurisprudencia-producto` en los dos modos; `consulta_preparada.md`, `documento_aportado.md`; `evals/jurisprudencia` con `01-existe-con-numero-y-fecha.yaml`, `03-por-materia.yaml`, `04-documento-pegado.yaml`, `06-documento-que-no-es-el-pedido.yaml`); las variantes versionadas de una familia son las suyas menos `producto`. Lo que el arnés y la tabla usan hoy (`familiaDeNormativa`, `tiene`, `directorioDeLaVariante`…) da lo mismo que hoy para las variantes de H40, de modo que `TestArnesDeNormativa`, `TestArnesDelLaboratorio`, `TestTablaDeNormativa` y `TestTablaDelLaboratorio` siguen en verde sin cambiar ningún valor esperado; aceptar `<skill>-producto` en el arnés y en la tabla llega con T005 y T006. **Tests**: en internal/evals/laboratorio_familias_test.go, nuevo, `TestVariantesDeControl` (contracts/variantes.md §2), por familia del producto: `laboratorio/<skill>/` tiene un directorio por variante versionada y nada más, y ninguno es `<skill>-producto`; los ficheros de cada `<skill>-control` son, ruta a ruta, los de una tabla `ruta → sha256` escrita en el test al copiarlos (17: 9 de `normativa`, 5 de `plazos-y-territorio`, 3 de `jurisprudencia`; calculadas con `shasum -a 256` sobre las copias), y falla con la ruta del que difiere, falta o sobra; y los nombres de las variantes de las cuatro familias (H39 de `laboratorio/variantes.yaml` y las tres) no se repiten. En internal/evals/laboratorio_normativa_test.go, `TestVariantesDeNormativa` sin los subtests «directorios de la familia» y «control es skills/normativa», que pasan a `TestVariantesDeControl`; los demás, sin cambiar ningún valor esperado (FR-021). En internal/evals/laboratorio_test.go, `TestLaboratorioFueraDelProducto` generalizado (contracts/variantes.md §2): los nombres que no pueden ser un tramo de una ruta de lo empotrado (`kitlegal.Skills()`), de lo que deja `kitlegal skills install --dir`, de una skill que ejecute el job ni de una carpeta de `evals/` son los de H39 y los de los directorios de variante de las carpetas de las tres familias; ninguna ruta de `laboratorio/` tiene un tramo `<skill>-producto` ni mide (`rutaQueMide`); lo demás como hoy. Mutantes que se ven en rojo y se revierten: un byte cambiado en `laboratorio/jurisprudencia/jurisprudencia-control/skills/jurisprudencia/SKILL.md`; un fichero de más en `plazos-y-territorio-control`; una referencia borrada de `normativa-control`; un directorio `laboratorio/normativa/normativa-producto/`; un nombre de variante de una familia repetido en otra; y una copia de `plazos-y-territorio-control` bajo un nombre que el recorrido de lo empotrado ve. `TestSkillsDelRepositorio` y `make skills-check` siguen en verde con `skills/` sin cambios

**Checkpoint**: las tres `<skill>-control` existen y no pueden cambiar sin que `make ci` lo diga; `skills/` aún no ha
cambiado.

---

## Phase 3: User Story 1 y User Story 2 — Las remisiones son órdenes de leer, y lo que sale del cuerpo está en su referencia (Priority: P1) 🎯 MVP

**Objetivo**: las trece remisiones reescritas en las tres skills y el movimiento de `normativa`, atados a sus
`<skill>-control` por el cálculo desde control y a la regla de las remisiones por un test del repositorio.

**Independent Test**: `TestSkillsDelProductoDesdeControl`, `TestRemisionesDelRepositorio` y `TestSkillsDelRepositorio`
en verde con `make skills-check` y `make ci`, y en rojo con cada mutante (SC-001 a SC-004); el prototipo de
`h41-remisiones.txtar` en verde.

- [X] T003 [US2] `normativa` manda leer sus referencias y deja en `redaccion_modificada.md` lo que solo usa su caso, con el control que calcula el producto desde control (FR-004, FR-007, FR-008, FR-009, FR-010; SC-002, SC-003, SC-004; contracts/remisiones.md §1, §2, §3 y §7; data-model §5; research D1, D2, D7, V2 a V4). **Skill**: skills/normativa/SKILL.md con los tramos del `diff -u` de contracts/remisiones.md §2 (R1 a R6 y la salida de las dos líneas de «Nada de otra conversación» del paso 4), y skills/normativa/references/redaccion_modificada.md con las tres líneas de §3 al final, detrás de una línea en blanco; el cuerpo queda en 10 981 caracteres; las cinco formas fijas siguen en sus bloques `text` y los pasos 3 y 6 y la primera viñeta del paso 4, línea a línea, como exige `h32-normativa.txtar`. **Test**: en internal/evals/laboratorio_familias_test.go (y, si la declaración necesita un tipo, internal/evals/laboratorio_familias.go), `TestSkillsDelProductoDesdeControl` (contracts/remisiones.md §7), por skill de las tres, con la declaración de data-model §5 escrita con los bytes de §2 y §3 para `normativa` —cambios `{antes, despues}` del cuerpo y el movimiento `{texto, referencia, enLaReferencia}`— y vacía para `plazos-y-territorio` y `jurisprudencia` (que hoy son su control): lee control y `skills/<skill>` sin lo que genera `make skills-sync` (referencias cuya primera línea empieza por `<!-- generado desde `; las regiones entre `<!-- paquetes: …` y `<!-- fin de los paquetes -->` y entre `<!-- inicio de la tabla de comandos…` y `<!-- fin de la tabla de comandos -->` reducidas a sus marcas); aplica cada cambio, cuyo `antes` tiene que estar una sola vez, y añade a cada referencia su `enLaReferencia`; exige el resultado ruta a ruta y byte a byte en `skills/<skill>` y falla con la ruta del primer fichero que difiere, falta o sobra, o con el cambio cuyo `antes` ya no está una vez; por cada movimiento, su texto y su `enLaReferencia`, con los blancos juntados y `references/<x>` escrito `<x>`, son iguales y el texto no está en el `SKILL.md` de `skills/<skill>`; y el frontmatter de `skills/<skill>/SKILL.md`, de la primera línea a la que lo cierra, es el de control byte a byte. Mutantes que se ven en rojo y se revierten: un byte fuera de lo declarado en `skills/normativa/SKILL.md`; «Nada de otra conversación» devuelto al cuerpo; una letra de la `description` cambiada; un byte de más en `redaccion_modificada.md`; un `antes` que no está en control; y, para `TestSkillsDelRepositorio` (SC-002, de hoy), el cuerpo de `skills/normativa/SKILL.md` llevado a 11 001 caracteres, que da `cuerpo de 11001 caracteres (máximo 11000)`. `TestSkillsDelRepositorio` (sin deriva tras `make skills-sync`, FR-010) y las suites activas siguen en verde
- [X] T004 [US1] `plazos-y-territorio` y `jurisprudencia` mandan leer sus referencias, y la regla de las remisiones como test del repositorio (FR-001, FR-002, FR-003, FR-005, FR-006, FR-007, FR-008, FR-010; SC-001, SC-002, SC-003, SC-004; contracts/remisiones.md §4, §5, §6 y §7; research D3 a D6, V5, V6, V11). **Skills**: skills/plazos-y-territorio/SKILL.md con el `diff -u` de §4 (P1 a P3; cuerpo de 10 982 caracteres; la orden generada de `administrativo.md` no cambia) y skills/jurisprudencia/SKILL.md con el de §5 (J1 y J2; cuerpo de 10 842 caracteres); ninguna referencia de las dos cambia. **Tests**: en internal/evals/laboratorio_familias_test.go, la declaración de `TestSkillsDelProductoDesdeControl` de las dos skills, con los bytes de §4 y §5, en lugar de la vacía. En internal/skills/remisiones_test.go, nuevo, en el paquete y con `raizDelRepositorio = "../.."`, `TestRemisionesDelRepositorio` (contracts/remisiones.md §6), con el lector de líneas fuera de bloques y la expresión del enlace de `internal/skills/markdown.go` sin cambiar código del paquete: sobre el `SKILL.md` de cada skill de `skills/` que tiene `references/`, las líneas de bloques de código cuentan como en blanco; la frase de cada enlace `[…](references/<f>)` o `[…](references/<f>#…)` con los límites de §6.2; y los fallos `<skill>: references/<f>: <regla>`, uno por referencia y regla, en orden de nombre de las referencias, con los tres textos de §6.3 (la palabra entera `lee` o `Lee` antes del enlace; cada una de «si la tienes», «antes de escribir la respuesta» y «antes de dar la consulta» con los blancos y saltos juntados en un espacio y sin distinguir mayúsculas, el mensaje con la expresión tal como está en la lista; una referencia sin ningún enlace del cuerpo): con las tres skills de `skills/`, ninguno. Subtest «control»: la misma regla sobre el `SKILL.md` de cada `laboratorio/<skill>/<skill>-control/skills/<skill>/`, con exactamente los fallos de §6.4 (en `normativa`, dos en `con_herramientas.md` y uno en `normas.md`, `buscar_la_norma.md`, `redaccion_en_una_fecha.md`, `varios_bloques.md`, `lecturas_que_fallan.md` y `como_se_cita.md`; en `plazos-y-territorio`, el del verbo en `jerarquia_normativa.md`, `municipios.md` y `calculo_de_plazos.md`; en `jurisprudencia`, el de «antes de dar la consulta» en `consulta_preparada.md`, que abre su frase con «Antes»). Subtest «frases»: dos cuerpos escritos en el test, uno con una orden de leer antes de un bloque de código para un enlace escrito después y otro con «si la» y «tienes» en dos líneas de la frase de un enlace, que fallan los dos. En el Makefile, solo la expresión de tests de la receta de `skills-check`, que gana `TestRemisionesDelRepositorio` junto a `TestNormasDelRepositorio`; nada más de su receta ni de otra cambia. Mutantes que se ven en rojo y se revierten: el «lee» de una remisión de `plazos-y-territorio` cambiado por «leer»; «si la tienes» añadido a la frase de un enlace de `normativa`; «Antes de dar la consulta, lee» al principio de la frase de `consulta_preparada.md`; el enlace a `municipios.md` quitado del cuerpo; y un byte fuera de lo declarado en `skills/jurisprudencia/SKILL.md` (`TestSkillsDelProductoDesdeControl`). `make skills-check` corre el test nuevo y sale en verde. Antes de terminar, con un prototipo (worktree bajo `$TMPDIR` con specs/033-h41-las-skills-del/aceptacion/remisiones.txtar copiado en la carpeta de guiones de internal/app como `h41-remisiones.txtar`, `go -C <worktree> test -count=1 ./internal/app -run 'TestEntregaDelHito/^h41-remisiones$'`), comprobar que el guion congelado pasa

**Checkpoint**: las tres skills llevan sus órdenes de leer y lo movido; `make skills-check` lo comprueba; es la entrega
que recibe quien instala (historias 1 y 2).

---

## Phase 4: User Story 3 — El arnés mide `control` frente a `producto` en las tres familias (Priority: P1)

**Objetivo**: `make evals-laboratorio` acepta una familia del producto, instala `<skill>-producto` desde `skills/` y deja
la lista de sus referencias en `sondeo.json`.

**Independent Test**: `TestArnesDeLasFamilias`, sin modelo y con el sustituto de `claude`: 0 sesiones con cada petición
inválida y n/4 por eval en cada variante y modo con las válidas, cada una con la skill de su variante (SC-006, SC-007).

- [X] T005 [US3] El arnés con las cuatro familias y su control sin modelo (FR-022, FR-030, FR-032, FR-033; SC-006, SC-007; contracts/arnes.md §2 a §5; data-model §1, §3 y §4; research D10, D11, D12, V7, V8). **Código**: internal/evals/laboratorio_sondeo.go, internal/evals/laboratorio_tabla.go (solo el tipo `sondeoLeido`) y, si lo pide, internal/evals/laboratorio_familias.go: las familias de la petición son la de H39 y las tres del producto, en ese orden; los mensajes de contracts/arnes.md §2 (el de la mezcla, una por familia pedida en el orden de las familias, con la de H39 nombrada «del laboratorio de H39», de modo que `VARIANTES=control,normativa-control` da byte a byte el de H40; los demás, los de H39 y H40), todos juntos y sin abrir sesiones, construir nada ni crear `DIRECTORIO`; de una variante versionada, la copia de `laboratorio/<skill>/<variante>/skills/` como en H40; de `<skill>-producto`, la copia de `skills/<skill>` del árbol en `<temporal>/skills/<skill>-producto/<skill>/` con `os.CopyFS`; el directorio de evals del reparto `evals/<skill>` y el plan de `SESIONES/4` sesiones de cada una de las cuatro evals de la familia, en su orden, de nombre `<eval sin .yaml>-<nn>`; y `sondeo.json` con `referencias_del_producto` (los nombres de `references/` de la copia, en orden de nombre) solo si se pide `<skill>-producto`, y sin ella el de hoy byte a byte. La clave es el campo `ReferenciasDelProducto []string` con la etiqueta `json:"referencias_del_producto,omitempty"`, tercero tras `Modelo` y `ClaudeCode`, en `sondeoLeido` (internal/evals/laboratorio_tabla.go), el tipo con el que el arnés escribe hoy `sondeo.json` (`json.Marshal(sondeoLeido{…})` en `crearElSondeoDelLaboratorio`, internal/evals/laboratorio_sondeo.go) y con el que la tabla lo lee: el arnés y la tabla siguen compartiendo un solo tipo para el fichero, sin un segundo tipo que duplique data-model §4; su comentario de tipo nombra la clave nueva. **Tests**: en internal/evals/laboratorio_sondeo_test.go, `TestArnesDeLasFamilias` (contracts/arnes.md §5) con la preparación sin construir y el sustituto de `claude` de `TestArnesDelLaboratorio`: con cada petición inválida (`plazos-y-territorio-otra`; `normativa-producto` y `jurisprudencia-control`; `plazos-y-territorio-control` y `control`; `SESIONES=6`; `normativa-remisiones-en-orden` en modo orden), 0 sesiones, nada construido, `DIRECTORIO` sin crear y su mensaje; con `VARIANTES=<skill>-control,<skill>-producto MODOS=orden,herramienta SESIONES=4`, una por familia, una sesión de cada una de sus cuatro evals en cada variante y modo en `DIRECTORIO/<variante>/<modo>/<eval>-01/` y ninguna más; con `VARIANTES=normativa-remisiones-en-orden MODOS=herramienta SESIONES=8`, dos de cada eval; en cada sesión, `claude/skills/` con solo la skill de la familia, ruta a ruta y byte a byte la de `laboratorio/<skill>/<variante>/skills/<skill>/` o la de `skills/<skill>` en `<skill>-producto`, sin `servidor.json` en modo orden y con `kitlegal mcp serve` en herramienta; y `DIRECTORIO/sondeo.json` con `referencias_del_producto` igual a los nombres de `skills/<skill>/references` en orden de nombre si se pidió `<skill>-producto`, y sin la clave si no. `TestArnesDeNormativa`, `TestArnesDelLaboratorio`, `TestRecetasDelSondeo` y `TestServidorDelLaboratorio`, sin cambiar ningún valor esperado. Mutantes que se ven en rojo y se revierten: la copia de `<skill>-producto` tomada de `<skill>-control`; la clave `referencias_del_producto` escrita siempre; la mezcla de dos familias del producto aceptada; y una eval de la familia quitada del plan

**Checkpoint**: el sondeo de las trece filas de FR-032 se puede lanzar con las órdenes de quickstart.md, fuera del run.

---

## Phase 5: User Story 4 — La tabla cuenta la lectura con `cd` al directorio de las referencias (Priority: P1)

**Objetivo**: `make laboratorio-tabla` da la tabla de cualquiera de las tres familias del producto, con las referencias
de lo medido y la lectura con `cd`, sin cambiar las de H39 y H40.

**Independent Test**: `TestLecturaConCd`, `TestTablaDeLasFamilias`, `TestTablaDeNormativa` y `TestTablaDelLaboratorio`
sin cambios, y el guion congelado `h41-laboratorio-tabla.txtar` en verde (SC-005, SC-008).

- [X] T006 [US4] La tabla de una familia del producto y la lectura con `cd` (FR-040, FR-041, FR-042; SC-005, SC-008; contracts/tabla.md §1 a §3; data-model §6 y §7; research D13, D14). **Código**: internal/evals/laboratorio_tabla.go (y internal/evals/laboratorio_familias.go si lo pide): la familia es la de la primera variante que lee, una de las tres del producto si es suya y si no la de H39 como hoy; en la de una familia del producto, un directorio de variante de otra familia, también de H39, da `laboratorio: <ruta> no es ninguna variante de <skill>`, código 1 y sin tabla; la tabla usa el campo `ReferenciasDelProducto` de `sondeoLeido` que añade T005, sin tocar el tipo; las referencias de la tabla son la unión, en orden de nombre, de las de la familia por cada variante versionada que lee y de la lista de cada sondeo con un directorio `<skill>-producto`, y la celda de una referencia que la variante de la fila no tiene va vacía, en la fila y en el detalle; «Cumplen la regla» y «Cumple», solo en la de `normativa`; `cita_preparar`, `cita_cotejar`, `territorio_resolver`, `plazos_calcular` y las de `boe` cuentan como primera herramienta como en H40; y, solo en las familias del producto, la lectura con `cd` de contracts/tabla.md §2 sobre las órdenes simples de hoy (`ordenesSimples`): una orden simple cuya primera palabra que no es una asignación es `cd` y cuya siguiente, sin barra final, es `<skill>/references` o termina en `/<skill>/references`, y en una orden simple posterior de la misma orden una palabra que es exactamente `<f>` o `./<f>`. **Test**: en internal/evals/laboratorio_tabla_test.go, `TestLecturaConCd` (contracts/tabla.md §5), tabla de casos con `t.Parallel()`: las trece órdenes de la tabla de contracts/tabla.md §2, orden a orden, con la familia de `normativa`, y los casos de US4.2 (`cd …/normativa/references` en un `Bash` y `cat normas.md` en otro, que no abre), y la familia de H39 con `cd …/articulos-del-boe/references && cat cierre.md`, que no la abre. `TestTablaDeNormativa`, `TestTablaDelLaboratorio`, `TestReglaTestigo` y `TestOrdenesDeBash`, sin cambiar ningún valor esperado. Mutantes que se ven en rojo y se revierten: la lectura con `cd` aplicada en la familia de H39; el `cd` aceptado en otra orden de `Bash`; `cat otra/normas.md` aceptado; y `<skill>` de otra familia aceptado en la ruta del `cd`. Antes de terminar, con un prototipo (worktree bajo `$TMPDIR` con specs/033-h41-las-skills-del/aceptacion/laboratorio-tabla.txtar copiado en la carpeta de guiones de internal/app como `h41-laboratorio-tabla.txtar`, `go -C <worktree> test -count=1 ./internal/app -run 'TestEntregaDelHito/^h41-laboratorio-tabla$'`), comprobar que el guion congelado pasa, y que las suites activas de H39 y H40 siguen en verde (FR-042)
- [X] T007 [datos] [US4] Las sesiones fijas de las tres familias y su tabla esperada, con el test que las lee, en una sola tarea porque el control de SC-005 compara la salida del guion de la tabla, ya escrito en T006, byte a byte con esas tablas y no hay test sin los fixtures (FR-040, FR-041; SC-005; contracts/tabla.md §3 y §5; data-model §4, §6 y §7). **Fixtures**: internal/evals/testdata/laboratorio/familias/normativa/, internal/evals/testdata/laboratorio/familias/plazos-y-territorio/ e internal/evals/testdata/laboratorio/familias/jurisprudencia/, cada una con `sondeo/` (sondeos con su `sondeo.json` y su lista `referencias_del_producto` donde hay `<skill>-producto`; sesiones fijas con la forma de las de H40, en los dos modos; lecturas con `cd` que cuentan y que no; en `normativa`, una referencia de la lista del producto que la familia no declara; una sesión sin terminar, código 124) y `tabla.md`, la salida esperada byte a byte, con los recuentos, turnos y extremos de Wilson de H40, celdas vacías donde la variante no tiene la referencia y «Cumplen la regla» solo en la de `normativa`; e internal/evals/testdata/laboratorio/familias/mezcla/sondeo/h41/, el sondeo de una familia del producto con un directorio de variante de otra familia, sin `tabla.md`: su `sondeo.json` sin lista (no tiene directorio `<skill>-producto`), `normativa-control/orden/04-lgt-prescripcion-01`, una sesión terminada que es la primera variante en orden de nombre y hace de `normativa` la familia de la tabla, y `plazos-y-territorio-control/orden/02-territorio-municipio-no-cubierto-01`, que es de otra familia (el caso inverso del de `mezcla/h41/` del guion congelado, donde la familia es `jurisprudencia`). Ningún fichero de `testdata/` existente cambia; nada se graba. **Test**: en internal/evals/laboratorio_tabla_test.go, `TestTablaDeLasFamilias` (contracts/tabla.md §5), con `t.Parallel()`: por cada una de las tres familias, nombradas en el test y no sacadas de leer `familias/`, la salida de la tabla sobre `familias/<skill>/sondeo/` es su `tabla.md` byte a byte, con stderr vacío y código 0; y sobre `familias/mezcla/sondeo/`, código 1, salida estándar vacía y `laboratorio: <ruta>/h41/plazos-y-territorio-control no es ninguna variante de normativa` en la salida de error. Mutantes que se ven en rojo y se revierten: un turno cambiado en una sesión de `plazos-y-territorio`; la orden de una lectura con `cd` de `jurisprudencia` partida en dos `Bash`; la referencia de la lista del producto quitada de `sondeo.json` de `normativa`; un extremo de Wilson cambiado en un `tabla.md`; y el directorio `plazos-y-territorio-control` de `familias/mezcla/sondeo/h41/` renombrado a `normativa-remisiones-en-orden`

**Checkpoint**: la medida de la aceptación se puede contar sin modelo; las tablas de H39 y H40 no cambian.

---

## Phase 6: Polish & cross-cutting

- [X] T008 La documentación de lo que entra en el repositorio y la línea de lo que queda para antes de fusionar (FR-020, FR-030, FR-040, FR-050, FR-051, FR-052; DoD §1, punto 6; research D16; plan.md «Cierre y pendientes»): CHANGELOG.md, en *Unreleased*, el cambio de las tres skills (cada remisión, una orden de leer en su sitio; «Nada de otra conversación» a `redaccion_modificada.md`), el arnés con las tres familias del producto y la tabla con la lectura con `cd`; docs/SKILLS.md, en §1 «Cómo se remite», lo de research D16 (la orden cuya condición la da lo que devuelve una llamada va en el paso de esa llamada, detrás de ella, «antes de seguir»; el sitio para las órdenes se hace llevando a la referencia que se manda leer en ese caso lo que solo necesita ese caso; `make ci` y `make skills-check` comprueban la frase de cada remisión); CONTRIBUTING.md, en «El sondeo del laboratorio», las familias del producto, `<skill>-control` congelada en `v0.9.0` (ya no es `skills/normativa`), `<skill>-producto` instalada desde `skills/` y el sondeo de la aceptación; en el Makefile, solo las líneas `##` de ayuda de `evals-laboratorio` y `laboratorio-tabla`, que nombran las cuatro familias (las recetas no cambian: `TestRecetasDelSondeo` sigue en verde); el comentario de paquete de cmd/laboratorio/main.go y la cabecera de comentario de scripts/evals-laboratorio.sh, las cuatro familias, sin cambiar código ni órdenes; y specs/033-h41-las-skills-del/gates/supuestos.md, una línea `- [alcance] tasks: FR-052 …`: el run no lanza el sondeo de las trece filas ni escribe su ADR; la fusión queda pendiente de las tres tablas con el modelo y la versión anotados y de su lectura, y si no cumplen lo medido va a un ADR que dice qué medir después; y FR-051 lo mide el job de evals del cierre con los umbrales de hoy. `make ci` en verde

---

## User Story 5 — La persona mide y decide antes de fusionar (P2)

Sin tareas: FR-052 y SC-010 se hacen fuera del run (ADR 0032) con lo que entregan T002 a T007 y las órdenes de
quickstart.md; el informe final lo deja pendiente por la línea `[alcance]` de FR-052 que escribe T008 en
`gates/supuestos.md`.

---

## Dependencies & Execution Order

### Orden de las fases

- **Fase 1 (T001)**: primero; congela la suite.
- **Fase 2 (T002)**: las `<skill>-control` nuevas se copian antes de cambiar ninguna skill, y `normativa-control` deja
  de compararse con `skills/normativa` antes de que T003 la cambie.
- **Fase 3 (T003 → T004)**: T003 crea `TestSkillsDelProductoDesdeControl` sobre las tres `<skill>-control` de T002;
  T004 completa su declaración y crea `TestRemisionesDelRepositorio`, que necesita las tres skills reescritas (la de
  `normativa`, de T003) y las tres `<skill>-control` (subtest «control»). Tras T004, `h41-remisiones.txtar` pasa.
- **Fase 4 (T005)**: necesita las familias con `producto` (T002); copia `skills/<skill>` tal como esté.
- **Fase 5 (T006 → T007)**: T006 necesita las familias (T002) y lee la clave que escribe T005; T007 fija con fixtures lo
  que cuenta el código de T006. Tras T006, `h41-laboratorio-tabla.txtar` pasa.
- **Fase 6 (T008)**: al final; documenta lo de T003 a T007.

### Dentro de cada tarea

Test y la implementación mínima que lo pone en verde, en la misma tarea (T007, el test sobre el código de T006, con sus
fixtures); el mutante que ve fallar el control, revertido; `make ci` en verde antes de marcarla.

### Oportunidades de paralelismo

Ninguna tarea lleva [P]: T003 a T007 comparten `internal/evals/laboratorio_familias*.go` o dependen de la anterior
(T006 de la clave de `sondeo.json` de T005; T007 del código de T006), y T004 y T008 comparten el `Makefile`. Se
ejecutan en serie, en el orden de los ids.

## Parallel Example

```text
Sin tareas en paralelo: el bucle las ejecuta una a una, T001 → T008.
```

## Implementation Strategy

### MVP

1. Fase 1: la suite.
2. Fase 2: las familias y las `<skill>-control` (T002). **Validar**: `TestVariantesDeControl`.
3. Fase 3: las tres skills (T003, T004). **Validar**: `make skills-check` y el prototipo de `h41-remisiones.txtar`. Es
   lo que recibe quien instala.

### Entrega incremental

4. Fase 4: el arnés (T005). **Validar**: `TestArnesDeLasFamilias`.
5. Fase 5: la tabla (T006, T007). **Validar**: `TestLecturaConCd`, `TestTablaDeLasFamilias` y el prototipo de
   `h41-laboratorio-tabla.txtar`.
6. Fase 6: CHANGELOG, `docs/SKILLS.md`, CONTRIBUTING, ayuda del Makefile, comentarios y la línea de supuestos.

## Notes

- 8 tareas: T001 [aceptacion]; US3: T002, T005; US2: T003; US1: T004; US4: T006, T007; transversal: T008; US5: ninguna.
- Tarea [datos]: T007 (`internal/evals/testdata/laboratorio/familias/`), que dice en su línea por qué va con su test. El
  hito no tiene datos externos (plan.md «Datos externos»): ni manifiesto `grabaciones.json`, ni `TestGrabar*`, ni fila
  nueva en `docs/SOURCES.md`.
- Definition of Done: `make ci` y tests offline en cada tarea; R4 y R5 (convenciones); ningún sobre ni esquema cambia
  (sin tarea de `schemas/`); defecto `inesperado` y errores de uso (T005, T006); e2e en T001; CHANGELOG en T008; ADR:
  no aplica en el run, el de un sondeo que no cumple se escribe fuera del run (FR-052); `docs/SOURCES.md`: no aplica
  (ninguna fuente); skills con evals: las de hoy, que mide el job del cierre (FR-051); fixtures en T007.
- Cada fila de «Controles de umbral» de plan.md y la tarea que construye su control (y lo ve fallar con un mutante):
  SC-001/FR-001 a FR-003 (`TestRemisionesDelRepositorio`), T004; SC-002/FR-008 (`TestSkillsDelRepositorio`, de hoy),
  T003, con el mutante de 11 001 caracteres; SC-003/FR-010 (`TestSkillsDelProductoDesdeControl` y
  `TestSkillsDelRepositorio`), T003 y T004; SC-004/FR-007/FR-009 (`TestSkillsDelProductoDesdeControl`), T003 y T004;
  SC-001/SC-004/FR-001 a FR-006/FR-009 (`h41-remisiones.txtar` en `TestEntregaDelHito`), T001, que lo ve fallar hoy, y
  T004, que lo pone en verde; SC-005/FR-040/FR-041 (`TestTablaDeLasFamilias`, `TestLecturaConCd` y
  `h41-laboratorio-tabla.txtar`), T006 (`TestLecturaConCd`, y el guion que T001 ve fallar) y T007
  (`TestTablaDeLasFamilias`); SC-006/FR-022/FR-033 (`TestArnesDeLasFamilias`, `TestVariantesDeControl`,
  `TestLaboratorioFueraDelProducto`), T005 y T002; SC-007/FR-030 (`TestArnesDeLasFamilias`), T005; SC-008/FR-020
  (`TestVariantesDeControl`), T002; SC-008/FR-021 (`TestVariantesDeNormativa`, sin cambiar sus valores), T002;
  SC-008/FR-023 (`TestLaboratorioFueraDelProducto`), T002; SC-008/FR-042 (las suites congeladas de H39 y H40 en
  `TestEntregaDelHito`), todas, comprobado en T006; SC-009/FR-051, las tres filas `evals:`, las mide el job de evals
  del cierre, que hace el workflow tras la revisión final con los umbrales de hoy, sin tarea que los toque.
- Requisitos sin tarea propia: FR-031 (todas: ninguna graba, abre sesiones con modelo ni escribe fuera del repositorio
  y `$TMPDIR`); FR-032 (lo hace posible T005; las trece filas y las 520 sesiones se miden fuera del run); FR-051 y
  SC-009 (el cierre del workflow; la línea de T008); FR-052 y SC-010 (fuera del run; la línea `[alcance]` de T008).

---

## Phase 7: Convergence

- [X] T009 [US2] El paso 10 de `plazos-y-territorio` manda leer la referencia a la que salió su caso, detrás de la lectura (FR-009 y US2/AC1, «el cuerpo manda leer esa referencia en el paso que lo usa»; FR-003; la regla que T008 escribió en docs/SKILLS.md §1, «Cómo se remite»: la orden cuya condición la da lo que devuelve una llamada va en el paso de esa llamada, detrás de ella, «antes de seguir») (partial). Hoy, la corrección de T004 (contracts/remisiones.md §4; gates/supuestos.md) sacó del paso 10 de skills/plazos-y-territorio/SKILL.md «Si sale con 4 o 5, **`⚠ PLAZO:` se mantiene**: di aparte que no has podido leerlo, sin afirmar qué dice.» y lo llevó al final de `references/calculo_de_plazos.md`, pero el paso 10 no manda leer esa referencia: la única orden es P3, en «Referencias», antes del «Protocolo» y de la llamada que trae el 4 o el 5; en `normativa`, en cambio, las dos referencias a las que se movió texto (`redaccion_modificada.md`, `lecturas_que_fallan.md`) se mandan leer en el paso 2, detrás de la orden que trae su caso. **Skill**: en skills/plazos-y-territorio/SKILL.md, en el paso 10, detrás de la lectura de cada precepto, una orden que cumpla FR-001 a FR-003 —«lee» antes del enlace a `references/calculo_de_plazos.md`, la condición del caso (la lectura sale con 4 o 5), «antes de seguir» y lo que trae— sin tocar P1 a P3, la orden generada de `administrativo.md`, el frontmatter ni la tabla de comandos (FR-005, FR-010). El cuerpo sigue en 11 000 caracteres como mucho también con la línea de 70 de la vertical de ensayo de H32 (internal/skills/vertical_de_ensayo_test.go: hoy 10 873 sin paquetes, 10 930 como mucho); si la orden no cabe, el sitio se hace con otro movimiento de FR-009 del cuerpo de `plazos-y-territorio`, entero y palabra por palabra al final de la referencia escrita a mano —skills/plazos-y-territorio/references/calculo_de_plazos.md, skills/plazos-y-territorio/references/municipios.md o skills/plazos-y-territorio/references/jerarquia_normativa.md— que el cuerpo manda leer en el paso que lo usa, detrás de una línea en blanco, sin quitar ni cambiar de sentido ninguna regla (FR-007) y sin que ninguna línea de la referencia empiece por `⚠` (control 7 de H31). **Contrato y test**: specs/033-h41-las-skills-del/contracts/remisiones.md §4, el `diff -u` del paso 10 (y el del movimiento nuevo, si lo hay, con su fila «De | A | Texto | Por qué»), y en internal/evals/laboratorio_familias_test.go la declaración de `plazos-y-territorio` de `TestSkillsDelProductoDesdeControl` con esos bytes. **Docs**: CHANGELOG.md (*Unreleased*, la entrada de H41 de las tres skills: en `plazos-y-territorio`, también `calculo_de_plazos.md` en el paso 10 si la lectura sale con 4 o 5) y, si hay un movimiento nuevo, docs/SKILLS.md §1, en la lista de lo que pasó a una referencia. Mutantes que se ven en rojo y se revierten: la orden del paso 10 quitada (`TestSkillsDelProductoDesdeControl`); su «lee» cambiado por «leer» (`TestRemisionesDelRepositorio`); y el cuerpo llevado a 10 931 caracteres sin paquetes (`TestVerticalDeEnsayo`). Antes de terminar, con un prototipo (worktree bajo `$TMPDIR` con specs/033-h41-las-skills-del/aceptacion/remisiones.txtar copiado en la carpeta de guiones de internal/app como `h41-remisiones.txtar`, `go -C <worktree> test -count=1 ./internal/app -run 'TestEntregaDelHito/^(h41-remisiones|h32-normativa)$'`), comprobar que el guion congelado sigue en verde; `make skills-check` y `make ci` en verde
