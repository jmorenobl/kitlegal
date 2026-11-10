---

description: "Tareas de H41.1 · plazos-y-territorio manda leer sus referencias sin condición"
---

# Tasks: H41.1 · `plazos-y-territorio` manda leer sus referencias sin condición

**Input**: `specs/035-h41-1-plazos-y-territorio/` — [spec.md](spec.md), [plan.md](plan.md), [research.md](research.md),
[data-model.md](data-model.md), [quickstart.md](quickstart.md) y [contracts/](contracts/) (`skill.md`, `controles.md`,
`aceptacion.md`).

**Tests**: obligatorios (constitución III). La primera tarea escribe la suite de aceptación, que el workflow congela;
cada tarea posterior trae su test y la implementación que lo pone en verde, y deja `make ci` en verde por sí sola.

**Organización**: una fase por historia, en el orden en que se ejecutan una tras otra (plan.md «Orden de
implementación»). La línea de cada tarea declara todas las rutas que crea o modifica (el guardián rechaza cualquier
otra; `go.mod`, `go.sum`, `CHANGELOG.md` y el directorio del feature siempre se admiten, y declarar `x.go` admite
`x_test.go`).

## Format: `[ID] [P?] [Story] Descripción`

- **[P]**: puede ir en paralelo (ficheros distintos, sin dependencia de una tarea sin terminar): T004 y T005.
- **[Story]**: la historia de [spec.md](spec.md) a la que sirve (US1, US2, US3, US5; US4 no tiene tareas: el sondeo se
  hace fuera del run, FR-008 y SC-003).
- **[aceptacion]**: la suite congelada (solo T001). **[datos]**: la tarea toca un fichero bajo `testdata/` (solo T002,
  que dice en su línea por qué va con la skill y los tests).

## Convenciones de todas las tareas

- `make ci` en verde al terminar cada una, en primer plano y, si hace falta, por objetivos (`make fmt-check`,
  `make lint`, `make test`, `make test-integration`, `make vuln`, `make schema-check`, `make skills-check`). Ninguna usa
  la red más que la de las herramientas de Go, ninguna graba (no hay datos externos: plan.md «Datos externos»; ni
  manifiesto `grabaciones.json` ni `TestGrabar*`), ninguna escribe en `data/` ni fuera del repositorio y `$TMPDIR`,
  ninguna abre una sesión con modelo ni ejecuta `TestSondeoDelLaboratorio`, `make evals-laboratorio`, `make evals`,
  `make evals-sondeo` ni los guiones `scripts/evals*.sh` (FR-008: el run no lanza el sondeo).
- Los bytes de la skill son los de [contracts/skill.md](contracts/skill.md) §2 y §3; los tests, los de
  [contracts/controles.md](contracts/controles.md) §1 a §3; los guiones, los de
  [contracts/aceptacion.md](contracts/aceptacion.md) §1 y §2; la guía y el registro de cambios, los de research D8 y D9.
  Nada se inventa fuera de ellos.
- No cambian un byte (FR-006, FR-008, FR-012; SC-005): `laboratorio/` (las cuatro variantes de `plazos-y-territorio` y
  las de `normativa`), `skills/normativa`, `skills/jurisprudencia`, `skills/consulta-juridica`, la descripción y el
  frontmatter de `skills/plazos-y-territorio`, sus referencias `administrativo.md`, `jerarquia_normativa.md` y
  `municipios.md`, `evals/`, `data/`, `schemas/`, el arnés (`scripts/evals-*.sh`, `cmd/laboratorio`,
  `internal/evals/laboratorio*.go` que no son `_test.go`), el generador de `internal/skills`, `Makefile`,
  `.golangci.yml`, `CONTRIBUTING.md` (research D10) y el ejemplo de remisión de `docs/SKILLS.md` §1 (Fuera de alcance).
  Ningún umbral, límite ni tope cambia; ninguna orden, bandera ni variable nueva de `kitlegal`.
- Ningún `//nolint`, ningún `t.Skip`, ningún paquete ni fichero de Go nuevo; tests de tabla con `testify`, como hoy.
- Cada control se ve fallar con un mutante de un solo cambio (la medida pasada del umbral) antes de darlo por bueno
  (ADR 0029; contracts/controles.md §5; quickstart.md §5): siempre sobre una copia en `$TMPDIR` (la de quickstart.md
  §5, y para los tests que leen `Makefile`, `scripts/` o `testdata/`, esos también, copiados o enlazados para leerlos),
  nunca sobre el árbol; la copia se borra al terminar. Los prototipos y guiones de un solo uso van bajo `$TMPDIR` y no
  se versionan.
- Las suites congeladas ya activas en internal/app/testdata/script/ (las de H32, H39, H40 y la de laboratorio-tabla de
  H41) siguen en verde sin cambiar un fichero suyo; de las de H41, solo cambia `h41-skills-instaladas.txtar`, y solo en
  T002 (FR-013).

---

## Phase 1: Setup — la suite de aceptación

**Propósito**: fijar, antes de cualquier cambio, lo que instala el binario de `plazos-y-territorio` (ADR 0018; plan.md
«Aceptación e2e»).

- [ ] T001 [aceptacion] Escribir, a partir del spec y de [contracts/aceptacion.md](contracts/aceptacion.md) §1, sin código de producto y sin evals nuevas (FR-011 y FR-012: las evals no cambian; Fuera de alcance: «Evals nuevas, y cambiar las que ya existen»), el guion testscript specs/035-h41-1-plazos-y-territorio/aceptacion/skills-instaladas.txtar, byte a byte el bloque de §1 (39 líneas): la cabecera de comentario con lo que comprueba y la línea «# Cubre: US1, US2 · FR-001, FR-002, FR-003 · SC-001»; `exec kitlegal skills install --dir destino --json`; y, sobre `destino/plazos-y-territorio/SKILL.md` y `destino/plazos-y-territorio/references/calculo_de_plazos.md`, las órdenes `grep`, `grep -count=1` y `! grep` de §1 en su orden (la orden de `municipios.md` en un párrafo propio al principio del «Protocolo», delante de «Territorio: del 1 al 4…»; la de `calculo_de_plazos.md` al final del paso 7, delante de «8. **Situar el plazo**»; una vez cada enlace; ninguna viñeta `- [municipios.md]` ni `- [calculo_de_plazos.md]`; «Referencias» terminada en la viñeta de `jerarquia_normativa.md` y seguida de «## Protocolo»; «Si no lo dice, **pregúntalo** y espera» en el paso 1; la frase del 4 o 5 fuera del cuerpo, el paso 10 terminado en ``⚠ REDACCIÓN A FECHA:`.`` y seguido del 11, y la frase al final de `calculo_de_plazos.md` detrás de una línea en blanco, con `\z`). El guion no contiene `usage: `, `unknown command`, `unexpected command` ni `cannot parse`, y no afirma nada que §1 no afirme (las otras tres referencias y la región de los paquetes los comprueban los tests de T002; research D6). Lo activa el workflow como `h41-1-skills-instaladas.txtar` en la carpeta de guiones de internal/app y lo ejecuta `TestEntregaDelHito` (el arnés e2e no cambia). Antes de terminar, con un prototipo (worktree bajo `$TMPDIR` con el guion copiado como `h41-1-skills-instaladas.txtar` en la carpeta de guiones de internal/app del worktree, `go -C <worktree> test -count=1 ./internal/app -run '^TestEntregaDelHito$/^h41-1-skills-instaladas$'`; el worktree se retira al terminar), comprobar que falla hoy por una aserción y no por no poder leerse: en su línea 21, el primer `grep`, con `no match for "^## Protocolo\n\nAntes de la primera orden o llamada, lee …"` (research V6).

**Checkpoint**: la suite queda congelada; ninguna tarea posterior la toca.

---

## Phase 2: User Story 1, User Story 2 y la base de User Story 3 — La skill es `-sin-condicion` con M1, y `make ci` lo comprueba (Priority: P1) 🎯 MVP

**Goal**: `skills/plazos-y-territorio` manda leer `municipios.md` y `calculo_de_plazos.md` sin condición (US1), con la
frase del 4 o 5 al final de `calculo_de_plazos.md` (US2), y `make ci` comprueba que es la forma sin condición con M1 y
nada más, el tope y que las variantes no cambian (US3, escenarios 1 a 3 y 5).

**Independent Test**: `go test -count=1 -run '^TestEntregaDelHito$/^h41-skills-instaladas$' ./internal/app/`,
quickstart.md §3 (los tests de `internal/evals` y `TestVerticalDeEnsayo`) y §4 (los `diff` y `cmp` frente a
`-sin-condicion` y el cuerpo de 10 882 caracteres); el prototipo del guion de T001, en verde.

- [ ] T002 [datos] [US1] La skill con M1, los dos tests de `internal/evals` que la protegen y el bloque de FR-013 del guion de H41, en una sola tarea porque con la skill cambiada cada parte que falte deja `make ci` en rojo, y con la skill de hoy el bloque nuevo no se cumple (research D7, D12, V5, V13; plan.md «Orden de implementación», paso 2) — US1, US2 y la base de US3; FR-001, FR-002, FR-003, FR-004, FR-005, FR-006, FR-008, FR-012, FR-013; SC-001, SC-002, SC-005. **Skill**: skills/plazos-y-territorio/SKILL.md, el de laboratorio/plazos-y-territorio/plazos-y-territorio-sin-condicion/skills/plazos-y-territorio/SKILL.md con el único cambio de [contracts/skill.md](contracts/skill.md) §2 (sus líneas 68-69 pasan a la línea ``    `⚠ REDACCIÓN A FECHA:`.``), y después `make skills-sync`, que deja la región de los paquetes como la genera `data/` (sin deriva en `make skills-check`); skills/plazos-y-territorio/references/calculo_de_plazos.md, el de hoy con, al final, una línea en blanco y la línea «Si sale con 4 o 5, **`⚠ PLAZO:` se mantiene**: di aparte que no has podido leerlo, sin afirmar qué dice.» con su salto (§3: 2 869 bytes); las otras tres referencias y el frontmatter, sin tocar. **Tests**: internal/evals/laboratorio_familias_test.go, `TestSkillsDelProductoDesdeControl` según [contracts/controles.md](contracts/controles.md) §1 y data-model.md §3: el campo `desde` en `declaracionDelProducto` (vacío, la de control; `normativa` y `jurisprudencia` no lo llevan y sus declaraciones no cambian), la constante `siSaleConCuatroOCinco` y la declaración de `plazos-y-territorio` de §1.1 en `declaracionesDelProducto`; los subtests «desde control», «movimientos» y «frontmatter» sobre la base elegida, con los mensajes de §1.2 (`skills/plazos-y-territorio/<ruta> difiere de lo que dan plazos-y-territorio-sin-condicion y su declaración`, `falta: está en …`, `sobra: no está en …`; «su control» en `normativa` y `jurisprudencia`, como hoy); sale el subtest «hasta H41.1» y los comentarios que dicen que la skill es su control; internal/evals/laboratorio_sondeo_test.go, solo `origenesDeLaVariante` y su comentario, según §3 (en toda `<skill>-producto`, solo `skills/`). **Guion**: internal/app/testdata/script/h41-skills-instaladas.txtar, sus líneas 11-13 y 17-20 de cabecera y su bloque de las líneas 48-53 por los de [contracts/aceptacion.md](contracts/aceptacion.md) §2, byte a byte; `normativa`, `jurisprudencia`, «Cubre:» y lo de las variantes no instaladas, como están. **Verde**: quickstart.md §3 y §4 (el primer `diff` da solo `68,69c68`, el segundo solo `36a37,38`, los tres `cmp` con 0 y el cuerpo 10 882), los dos guiones `h41-skills-instaladas` y el prototipo de `h41-1-skills-instaladas` de T001, `TestVariantesDeControl`, `TestVariantesDePlazosYTerritorio`, `TestArnesDeLasFamilias`, `TestRecetasDelSondeo`, `TestLaboratorioFueraDelProducto`, `TestSkillsDelRepositorio`, `TestVerticalDeEnsayo`, `TestRemisionesDelRepositorio` (aún con `plazos-y-territorio` fuera, hasta T003) y los de la tabla, sin cambiar ningún valor esperado, y `make lint` sin `issues`. **Mutantes**, sobre la copia y de uno en uno (contracts/controles.md §5; research V5, V7, V13): un byte añadido a `calculo_de_plazos.md` de la skill y un carácter quitado de su cuerpo (10 881) → rojo `TestSkillsDelProductoDesdeControl` con su ruta; un byte de `municipios.md` de `-sin-condicion` → rojo `TestVariantesDePlazosYTerritorio` y `TestSkillsDelProductoDesdeControl`; por la rama con `desde` vacío, que esta tarea reescribe, un byte al principio de `skills/normativa/references/con_herramientas.md` → rojo `TestSkillsDelProductoDesdeControl/normativa/desde_control` con `skills/normativa/references/con_herramientas.md difiere de lo que dan su control y su declaración`, y otro al principio de `skills/jurisprudencia/references/documento_aportado.md` → rojo `TestSkillsDelProductoDesdeControl/jurisprudencia/desde_control` con `skills/jurisprudencia/references/documento_aportado.md difiere de lo que dan su control y su declaración` (dos ficheros escritos a mano, fuera de lo generado; al principio, para que «movimientos» no cambie); un byte de `municipios.md` de `plazos-y-territorio-control` → rojo `TestVariantesDeControl` («difiere: su huella es …»); un byte del `SKILL.md` de `-lista-con-verbo` → rojo `TestVariantesDePlazosYTerritorio`; 49 caracteres más en el cuerpo (10 931) → rojo `TestVerticalDeEnsayo` con `plazos-y-territorio: SKILL.md: cuerpo de 11001 caracteres (máximo 11000)` y verde `TestSkillsDelRepositorio`, y 119 más (11 001) → rojo también `TestSkillsDelRepositorio` con el mismo defecto; un carácter cambiado en la receta de `evals-laboratorio` del `Makefile` de la copia → rojo `TestRecetasDelSondeo`; el segundo origen de `plazos-y-territorio-producto` repuesto en `origenesDeLaVariante` → rojo `TestArnesDeLasFamilias` («la skill de la sesión … es, ruta a ruta y byte a byte, la de …»); y, para el bloque de FR-013, mutaciones de la skill de la copia con el guion de H41 nuevo intacto (cabecera de seis líneas: el bloque empieza en la línea 51 y sus seis `grep` son las líneas 54 a 59; contracts/aceptacion.md §2), cada una → rojo `TestEntregaDelHito/h41-skills-instaladas` en la línea del `grep` que deja de cumplirse: `skills/plazos-y-territorio` repuesta a la de hoy (la `skills/plazos-y-territorio` de `plazos-y-territorio-control`, que hoy es igual byte a byte) → línea 54, lo que research D7 da por hecho; «lee » quitado de la orden de `municipios.md` (`llamada, lee [municipios.md]` → `llamada, [municipios.md]`) → línea 54; «lee» quitado de la orden de `calculo_de_plazos.md` (`Antes del 8, lee` → `Antes del 8,`) → línea 55; la viñeta `- [municipios.md](references/municipios.md): …` de control repuesta en «Referencias» → línea 56 (`! grep`); «**pregúntalo**» → «pregúntalo» en el paso 1 → línea 57; la frase «Si sale con 4 o 5, **`⚠ PLAZO:` se mantiene**: …» añadida al cuerpo, detrás del paso 10 → línea 58 (`! grep`); y un byte añadido al final de `references/calculo_de_plazos.md` → línea 59 (`\z`). `make ci` en verde.

**Checkpoint**: lo que recibe quien instala ya es la forma aplicada, y `make ci` la protege; falta solo la frase de cada
remisión (T003).

---

## Phase 3: User Story 3 — `TestRemisionesDelRepositorio` mira también `plazos-y-territorio` (Priority: P1)

**Goal**: la regla del verbo, de las expresiones y de la referencia sin orden y la lista del plan se aplican a
`plazos-y-territorio`, con el único fallo admitido de la viñeta de `jerarquia_normativa.md` (US3, escenario 4).

**Independent Test**: quickstart.md §1 (`make skills-check`) y `go test -count=1 -run '^(TestRemisionesDelRepositorio|TestFrases)$' ./internal/skills/`.

- [ ] T003 [US3] `TestRemisionesDelRepositorio` sin exclusiones y con los fallos que admite cada skill del producto, en internal/skills/remisiones_test.go, según [contracts/controles.md](contracts/controles.md) §2 y data-model.md §4 (necesita la skill de T002: con la de hoy, la lista del plan no se cumple; research V4) — US3; FR-007; SC-001. Sale `fueraDeLasRemisiones` con su `continue` y su comentario; entra `fallosAdmitidosDelProducto` de §2.1, con la única entrada `"plazos-y-territorio: references/jerarquia_normativa.md: " + remisionSinElVerbo` y el comentario que dice por qué (la viñeta queda como en v0.9.0, FR-002; ADR 0053, punto 1 y «Lo que cuesta»: ninguna eval llega a su condición); por cada skill de `skills/` con `references/`, en orden de nombre, `assert.Equal(t, fallosAdmitidosDelProducto[skill], fallosDeLaRegla(skill, segmentos, referencias), "%s: la regla da exactamente los fallos que admite", skill)` sobre todo el cuerpo, región de los paquetes incluida; y en `remisionesDelPlan`, la lista de `plazos-y-territorio` de §2.2, sus tres frases en ese orden (la de `jerarquia_normativa.md`, la de `municipios.md` y la de `calculo_de_plazos.md`), con el defecto de hoy (`plazos-y-territorio: la remisión <n> no es la del plan`, con las dos listas). `fallosDeControl` y el subtest «control» no cambian; el comentario del test deja de nombrar exclusiones. **Verde**: quickstart.md §1 y la orden de `internal/skills` de §3, y `make lint` sin `issues`. **Mutantes**, sobre la copia y de uno en uno (contracts/controles.md §5; research V7): «sin calcular» → «sin contar» en la remisión de `calculo_de_plazos.md` → rojo con `plazos-y-territorio: la remisión 3 no es la del plan`; «Lee » delante del enlace de la viñeta de `jerarquia_normativa.md` → rojo con `plazos-y-territorio: la regla da exactamente los fallos que admite` (deja de dar el admitido) y `la remisión 1 no es la del plan`; y «lee» quitado de la orden de `municipios.md` → rojo con `plazos-y-territorio: la regla da exactamente los fallos que admite` (da uno no admitido). `make ci` en verde.

**Checkpoint**: los controles de US3 están todos; lo que queda es documentación.

---

## Phase 4: User Story 5 — La guía y el registro de cambios dicen lo medido (Priority: P3)

**Goal**: `docs/SKILLS.md` §1 dice lo medido en `normativa` y `plazos-y-territorio` y describe el control tal como
queda; `CHANGELOG.md` anuncia el cambio de la skill.

**Independent Test**: quickstart.md §6: `grep -n 'hasta H41.1' docs/SKILLS.md` sale con 1 sin nada; `grep -n 'H41.1'
CHANGELOG.md` da la entrada de *Unreleased*.

- [ ] T004 [P] [US5] docs/SKILLS.md, en §1 «Cómo se remite», los dos cambios de research D8 y ninguno más — US5; FR-009: la frase «`plazos-y-territorio` sigue como estaba hasta H41.1, que le aplica la forma de sus órdenes que mida el laboratorio (ADR 0052, punto 3).» pasa a las dos frases de D8, punto 1 (la forma que midió el laboratorio, ADR 0053, con sus dos sitios y las cifras de 39 y 40 de 40 sesiones frente a 3 y 11, y la línea de FR-009: lo que la skill necesita en las preguntas a las que llega se manda leer sin condición, antes de la primera orden o llamada o del paso que lo usa; ADR 0053, punto 3), sin el punto 2 del ADR 0052; y en la viñeta de `TestRemisionesDelRepositorio` sale «—salvo `plazos-y-territorio`, que el test declara fuera hasta H41.1—» y entra, detrás de «Cada una lleva `lee` o `Lee` antes del enlace», el inciso de D8, punto 2 (`fallosAdmitidosDelProducto`, hoy uno: la viñeta de `jerarquia_normativa.md` de `plazos-y-territorio`). «De ahí, tres reglas más:» y el ejemplo de remisión del apartado no cambian. **Verde**: los dos primeros `grep` de quickstart.md §6 y el `awk` del apartado, leído contra FR-009; `make ci` en verde (la guía no entra en ningún test que cambie su valor esperado).
- [ ] T005 [P] [US5] CHANGELOG.md, en *Unreleased* → «Cambiado», detrás de la entrada de H41, la entrada de research D9, palabra por palabra — US5; FR-010: `plazos-y-territorio` manda leer `municipios.md` y `calculo_de_plazos.md` sin condición (H41.1; ADR 0053), sus dos viñetas fuera de «Referencias», la frase «`⚠ PLAZO:` se mantiene» al final de `calculo_de_plazos.md` para que el cuerpo quepa en el tope, nada más de la skill cambia, y `make ci` comprueba la forma con M1 y la frase de cada remisión. **Verde**: el segundo `grep` de quickstart.md §6 da la línea de la entrada; quickstart.md §1 a §4; y `make ci` en verde (quickstart.md §7).

---

## User Story 4 — La persona mide la forma aplicada con el sondeo de H41.1 (P2)

Sin tareas: FR-008 y SC-003 se miden fuera del run (ADR 0032), con el arnés de H41 sin cambiar y lo que entrega T002, y
las órdenes de quickstart.md §8. El run no lanza el sondeo ni abre sesiones con modelo; la decisión de fusionar queda
pendiente de su tabla y su lectura (plan.md «Cierre y pendientes»; gates/supuestos.md).

---

## Dependencies & Execution Order

### Orden de las fases

- **Fase 1 (T001)**: primero; congela la suite, que hoy falla en su línea 21.
- **Fase 2 (T002)**: la skill, la declaración `desde`, `origenesDeLaVariante` y el bloque de FR-013 juntos (research
  D7). Tras T002, el guion de T001 pasa.
- **Fase 3 (T003)**: necesita la skill de T002 para la lista del plan y el único fallo admitido.
- **Fase 4 (T004, T005)**: al final; documentan el control tal como queda tras T003. Independientes entre sí.

### Dentro de cada tarea

Test y la implementación mínima que lo pone en verde, en la misma tarea; el mutante que ve fallar el control, sobre la
copia; `make ci` en verde antes de marcarla.

### Oportunidades de paralelismo

Solo T004 y T005 (ficheros distintos, sin dependencia entre ellas). T002 y T003 dependen de la anterior. El bucle las
ejecuta una a una, en el orden de los ids.

## Parallel Example

```text
# Tras T003, las dos de la fase 4 pueden ir a la vez:
T004 docs/SKILLS.md §1 «Cómo se remite» (FR-009)
T005 CHANGELOG.md, Unreleased (FR-010)
```

## Implementation Strategy

### MVP

1. Fase 1: la suite (T001).
2. Fase 2: la skill con M1 y sus controles (T002). **Validar**: quickstart.md §2 a §4. Es lo que recibe quien instala.

### Entrega incremental

3. Fase 3: la frase de cada remisión (T003). **Validar**: quickstart.md §1.
4. Fase 4: guía y registro de cambios (T004, T005). **Validar**: quickstart.md §6 y §7.

## Notes

- 5 tareas: T001 [aceptacion]; US1, US2 y la base de US3: T002; US3: T003; US5: T004, T005; US4: ninguna.
- Tarea [datos]: T002, por `internal/app/testdata/script/h41-skills-instaladas.txtar`, un fichero existente de
  `testdata/` (FR-013; research V10); dice en su línea por qué va con la skill y los dos tests. El hito no tiene datos
  externos (plan.md «Datos externos»): ni manifiesto `grabaciones.json`, ni `TestGrabar*`, ni grabación.
- Definition of Done (docs/ROADMAP.md §1): `make ci` y tests offline en cada tarea; ningún código de producción, sobre
  ni esquema cambia (sin tarea de `schemas/`; R4 y R5, por las convenciones); e2e: T001 (el guion nuevo) y T002 (el de
  H41, FR-013); CHANGELOG: T005; ADR: no aplica (la decisión es el ADR 0053, que el hito aplica; el que pediría un
  SC-003 que no cumple se escribe fuera del run, Fuera de alcance); `docs/SOURCES.md`: no aplica (ninguna fuente);
  fixtures: ninguno nuevo; skills con evals: las de hoy, sin evals nuevas (FR-011, FR-012), que mide el job del cierre;
  dimensión territorial: ninguna orden ni texto nombra un municipio, y las evals 02 y 05 siguen siendo la matriz
  cubierto/no cubierto.
- Cada fila de «Controles de umbral» de plan.md y la tarea que construye o toca su control y lo ve fallar con un
  mutante:
  - FR-004/SC-002, `TestSkillsDelRepositorio` (de hoy, sin cambiar): T002, cuerpo de 11 001 sin la vertical.
  - FR-004/SC-002, `TestVerticalDeEnsayo` (de hoy, sin cambiar): T002, 49 caracteres más (11 001 con la vertical).
  - FR-004/SC-002/FR-005, `TestSkillsDelProductoDesdeControl` (10 882): T002, un carácter quitado del cuerpo.
  - FR-001/FR-002/FR-003/FR-005/SC-001, `TestSkillsDelProductoDesdeControl` («desde control», «movimientos»,
    «frontmatter»): T002, un byte de `calculo_de_plazos.md` y uno de `municipios.md` de `-sin-condicion`.
  - FR-006/SC-005, `TestVariantesDeControl` (de hoy, sin cambiar): T002, un byte de `plazos-y-territorio-control`.
  - FR-006/SC-005, `TestVariantesDePlazosYTerritorio` (de hoy, sin cambiar): T002, un byte de `-lista-con-verbo`.
  - FR-007/SC-001, `TestRemisionesDelRepositorio`: T003, sus tres mutantes.
  - FR-012/SC-005, `TestSkillsDelProductoDesdeControl` en `normativa` y `jurisprudencia` (sin cambiar sus
    declaraciones): T002, que añade `desde` sin tocarlas, reescribe el código que elige la base y comprueba sus
    subtests en verde; sus mutantes pasan por la rama con `desde` vacío: un byte al principio de
    `skills/normativa/references/con_herramientas.md` y otro al de `skills/jurisprudencia/references/documento_aportado.md`,
    cada uno en rojo en `TestSkillsDelProductoDesdeControl/<skill>/desde_control` con «skills/<skill>/<ruta> difiere de
    lo que dan su control y su declaración».
  - FR-008, `TestRecetasDelSondeo` (de hoy, sin cambiar): T002, un carácter de la receta de `evals-laboratorio`.
  - FR-008, `TestArnesDeLasFamilias`: T002, el segundo origen repuesto.
  - FR-001/FR-002/FR-003/SC-001, `h41-1-skills-instaladas.txtar` en `TestEntregaDelHito`: T001, que lo ve fallar hoy en
    su línea 21, y T002, que lo pone en verde.
  - FR-013/SC-001, `h41-skills-instaladas.txtar` en `TestEntregaDelHito`: T002, con el guion nuevo intacto y la skill
    de la copia cambiada, una mutación por cada uno de los seis `grep` del bloque (líneas 54 a 59 del fichero con la
    cabecera nueva, dos de ellos `! grep`), y la skill de hoy repuesta (rojo en la línea 54).
  - FR-011/SC-004, las filas `evals:` de `plazos-y-territorio`, `normativa` y `jurisprudencia`: las mide el job de
    evals del cierre, que hace el workflow tras la revisión final con los umbrales de hoy, sin tarea que los toque.
- Requisitos sin tarea propia: FR-011 y SC-004 (el cierre del workflow); FR-012 (todas: ninguna declara una ruta de lo
  que no cambia, y lo guarda el guardián de diff); FR-008 y SC-003 en su parte de medida (fuera del run; US4).
