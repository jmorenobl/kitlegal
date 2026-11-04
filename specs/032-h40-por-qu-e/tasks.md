---
description: "Tareas de H40 · Por qué normativa no lee sus referencias en modo herramienta"
---

# Tasks: H40 · Por qué `normativa` no lee sus referencias en modo herramienta

**Input**: `specs/032-h40-por-qu-e/` — [spec.md](spec.md), [plan.md](plan.md), [research.md](research.md),
[data-model.md](data-model.md), [quickstart.md](quickstart.md) y [contracts/](contracts/) (`variantes.md`, `arnes.md`,
`tabla.md`, `arnes-e2e.md`).

**Tests**: obligatorios (constitución III). La primera tarea escribe la suite de aceptación, que el workflow congela;
cada tarea posterior trae su test y la implementación que lo pone en verde, y deja `make ci` en verde por sí sola.

**Organización**: una fase por historia, en el orden en que se ejecutan una tras otra. La línea de cada tarea declara
todas las rutas que crea o modifica (el guardián rechaza cualquier otra; `go.mod`, `go.sum`, `CHANGELOG.md` y el
directorio del feature siempre se admiten, y declarar `x.go` admite `x_test.go`).

## Format: `[ID] [P?] [Story] Descripción`

- **[P]**: puede ir en paralelo (ficheros distintos, sin dependencia de una tarea sin terminar).
- **[Story]**: la historia de [spec.md](spec.md) a la que sirve (US1 a US3; US4 no tiene tareas).
- **[aceptacion]**: la suite congelada (solo T001). **[datos]**: la tarea toca ficheros bajo `testdata/`.

## Convenciones de todas las tareas

- `make ci` en verde al terminar cada una, en primer plano y por objetivos si hace falta (`make fmt-check`,
  `make lint`, `make test`, `make test-integration`, `make vuln`, `make schema-check`, `make skills-check`…). Ninguna
  usa la red más que la de las herramientas de Go, ninguna graba, ninguna abre una sesión con modelo ni ejecuta
  `TestSondeoDelLaboratorio`, `make evals-laboratorio` ni los guiones `scripts/evals*.sh` desde la terminal.
- No se tocan `skills/` (FR-002, FR-007, FR-043), `internal/mcp/` (FR-008: las instrucciones y descripciones de hoy),
  `evals/`, `laboratorio/variantes.yaml`, `laboratorio/variantes/`, `laboratorio/evals/`, las recetas del `Makefile`
  ni ningún fichero de `testdata/` que ya exista (FR-013, FR-023). Ningún argumento, bandera ni orden nueva de
  `kitlegal` ni de `cmd/laboratorio`.
- Los textos de las variantes son, byte a byte, los de [contracts/variantes.md](contracts/variantes.md); los mensajes
  del arnés, los de [contracts/arnes.md](contracts/arnes.md) §2; el formato de la tabla, el de
  [contracts/tabla.md](contracts/tabla.md) §3 y §4. Nada se inventa fuera de ellos.
- `internal/evals` no nombra `os.Stdin`, `os.Stdout`, `os.Stderr` ni llama a `os.Exit` (R4, R5); ningún `panic`; lo
  que llega de fuera manipulado es defecto `inesperado` (1). Ningún `//nolint`.
- Cada control se ve fallar con un mutante (la medida pasada del umbral) antes de darlo por bueno, y el mutante se
  revierte (ADR 0029).
- Los fixtures nuevos no llevan nada con forma de secreto (gitleaks corre en `make ci`).

---

## Phase 1: Setup — la suite de aceptación

**Propósito**: fijar, antes de cualquier código, lo que la tabla tiene que dar con las sesiones de las dos familias
(ADR 0018; plan.md «Aceptación e2e»).

- [X] T001 [aceptacion] [US3] Escribir, a partir del spec y de [contracts/arnes-e2e.md](contracts/arnes-e2e.md) y [contracts/tabla.md](contracts/tabla.md) §3 y §4, sin código de producto, el guion testscript specs/032-h40-por-qu-e/aceptacion/laboratorio-tabla.txtar (el workflow lo activa como `h40-laboratorio-tabla.txtar` al terminar el bucle) y ampliar el arnés e2e en internal/app/e2e_test.go para que pueda ejecutarlo (contracts/arnes-e2e.md §1; research D13, V13 a V17). **Arnés**: `TestMain` construye también `./cmd/laboratorio` con `go install` (paquete `github.com/jmorenobl/kitlegal/cmd/laboratorio`, `GOBIN` en una carpeta propia del temporal del arnés, nombre `laboratorio`) y cada guion recibe su ruta absoluta en `KITLEGAL_LABORATORIO_BIN`, junto a las constantes de las demás variables; no va en el `PATH` y no cambia ninguna otra variable, el binario de e2e, las órdenes del arnés ni ningún guion de hoy; si no se construye, el error va a `entorno.err`; `TestVariablesDeLosGuiones` la cuenta entre las rutas absolutas que fija y `TestBinariosDelArnes` no cambia. **Guion**: comentario de cabecera «Cubre: US3.1, US3.2, US3.3, US3.4 · FR-020, FR-021, FR-023, FR-030 · SC-005, SC-007; los extremos de Wilson de 0, 20 y 40 de 40 los comprueba TestTablaDeNormativa» (sin los textos `usage: `, `unknown command`, `unexpected command` ni `cannot parse`); órdenes, en este orden: `exec $KITLEGAL_LABORATORIO_BIN tabla h39`, `cmp stdout h39-tabla.md`, `! stderr .`, `exec $KITLEGAL_LABORATORIO_BIN tabla normativa`, `cmp stdout normativa-tabla.md`, `! stderr .`; ficheros del archivo: `laboratorio/variantes.yaml` copiado byte a byte de `main`; bajo `h39/`, los 63 ficheros de las sesiones fijas del laboratorio de H39 de `main` (las carpetas `h39-a` y `h39-b` del sondeo fijo que lee `TestTablaDelLaboratorio`), con sus rutas relativas y byte a byte, vacíos incluidos; `h39-tabla.md`, la tabla esperada de H39 de `main` byte a byte; `normativa/h40-a/sondeo.json` con `{"modelo":"claude-sonnet-5-5","claude_code":"2.1.317"}` y su salto; las cinco sesiones de la tabla de contracts/arnes-e2e.md §2 (`codigo-de-la-sesion`, `sesion.err` y `sesion.jsonl` con la forma de las de H39: `system/init` con modelo y versión, un `assistant` con su `message.id` por respuesta, resultados en mensajes `user` y `result`; la de `normativa-cuerpo-corto` sin terminar, código 124, como `h39-a/control/herramienta/04-et-vacaciones-01`); y `normativa-tabla.md`, la tabla y el detalle de contracts/tabla.md §3 y §4, byte a byte. Antes de terminar, con un prototipo (un worktree bajo `$TMPDIR` con el guion copiado en la carpeta de guiones de internal/app y ejecutado con `go -C <worktree> test ./internal/app -run 'TestEntregaDelHito'`), comprobar que la parte de H39 pasa y que el guion falla hoy por la aserción de `exec $KITLEGAL_LABORATORIO_BIN tabla normativa` (sale con 1, `… no es ninguna variante de la declaración del laboratorio`), no por un `exec` que no puede ejecutarse; `make ci` en verde con el arnés ampliado. No escribe evals de ninguna skill del producto, porque ninguna cambia (FR-043; Fuera de alcance)

**Checkpoint**: la suite queda congelada; ninguna tarea posterior la toca.

---

## Phase 2: User Story 1 — Cuatro variantes de `normativa`, cada una con una sola diferencia (Priority: P1) 🎯 MVP

**Objetivo**: las cuatro variantes en `laboratorio/normativa/`, fuera del producto, comprobadas por `make ci` contra
`skills/normativa` y contra su declaración.

**Independent Test**: `TestVariantesDeNormativa` y `TestLaboratorioFueraDelProducto` en verde, y en rojo con cada
mutante de sus controles (SC-001, SC-002, SC-003).

- [X] T002 [US1] La familia de `normativa`, la variante de control y los dos controles que la ven (FR-001, FR-002, FR-007; SC-001, SC-003; contracts/variantes.md §1, §2 y §7; data-model §1 y §2; research D1 a D4, D12, V1, V5, V10). **Código**: internal/evals/laboratorio_normativa.go, nuevo, con la familia declarada en Go (data-model §1): los nombres `normativa-control`, `normativa-remisiones-en-orden`, `normativa-cuerpo-corto` y `normativa-testigo-en-referencia`, en ese orden; la skill `normativa`; el directorio `laboratorio/normativa/<variante>/`; y las ocho referencias en orden de nombre de fichero (`buscar_la_norma.md` … `varios_bloques.md`); solo los campos que usan esta tarea y sus tests (los demás llegan con la tarea que los usa). **Laboratorio**: laboratorio/normativa/normativa-control/skills/normativa/ con los nueve ficheros de `skills/normativa`, copiados byte a byte. **Tests**: en internal/evals/laboratorio_normativa_test.go, nuevo, `TestVariantesDeNormativa` (con `t.Parallel()`): `normativa-control` frente a `skills/normativa`, mismo conjunto de rutas y mismos bytes, falla con la ruta del primer fichero que difiere, falta o sobra; las ocho referencias declaradas son exactamente los ficheros de `references/` de `normativa-control`; y ningún nombre de la familia es el de una variante de `laboratorio/variantes.yaml`. En internal/evals/laboratorio_test.go, `TestLaboratorioFueraDelProducto` generalizado (research D12): recorre todo `laboratorio/`; deja fuera de la comprobación de bytes los ficheros `…/skills/<s>/<r>` con los bytes de `skills/<s>/<r>`; suma los nombres de los directorios de `laboratorio/normativa/` a los que no pueden ser un tramo de una ruta de `kitlegal.Skills()` ni de lo que instala `kitlegal skills install --dir`, ni una skill que ejecute el job, ni una carpeta de `evals/`; y con los ficheros de H39 hace lo mismo que hoy (sin cambiar ningún valor esperado de `TestReglaTestigo`, `TestVariantesDelLaboratorio` ni `TestEvalsDelLaboratorio`). Mutantes que se ven en rojo y se revierten: un byte cambiado en `normativa-control/skills/normativa/SKILL.md`; un fichero de más en `normativa-control`; una referencia de `normativa-control` borrada; y una copia de `normativa-control` bajo un nombre que el recorrido de lo empotrado ve. `TestSkillsDelRepositorio` y `make skills-check` siguen en verde con `skills/` sin cambios (FR-007)
- [X] T003 [US1] Las tres variantes declaradas y su control (FR-003, FR-004, FR-005, FR-006, FR-007; SC-002; contracts/variantes.md §3, §4, §5 y §7; data-model §3; research D5, D6, V2, V3, V4). **Laboratorio**: laboratorio/normativa/normativa-remisiones-en-orden/skills/normativa/, laboratorio/normativa/normativa-cuerpo-corto/skills/normativa/ y laboratorio/normativa/normativa-testigo-en-referencia/skills/normativa/, cada una con los nueve ficheros: los de `normativa-control` y, en los declarados, el resultado de aplicarles su declaración con los bytes exactos de contracts/variantes.md (las cinco sustituciones de §3; la regla por secciones de §4, con `### 4. Responder` escrito `### 3. Responder`; la inserción en el «Protocolo» y la sección `## Cómo termina la respuesta` al final de `con_herramientas.md` de §5, con el párrafo leído de `laboratorio/variantes/referencia-paso/skills/articulos-del-boe/references/cierre.md`); el guion que las calcula, si lo hay, de un solo uso bajo `$TMPDIR` y sin versionar. **Tests**: en internal/evals/laboratorio_normativa_test.go (y, si la declaración necesita un tipo, internal/evals/laboratorio_normativa.go), `TestVariantesDeNormativa` amplía: `laboratorio/normativa/` tiene exactamente los cuatro directorios de la familia; cada una de las tres se compara byte a byte, fichero a fichero, con la que se calcula desde `normativa-control` con su declaración (data-model §3: sustituciones `{fichero, antes, después, hito}`, la regla del cuerpo corto y el añadido); falla con la ruta del primer fichero que difiere, con la sustitución, el encabezado o la viñeta que ya no se aplica y cuántas veces está, o con el «después» cuyo texto normalizado (sin la marca `- ` y con cada salto y la sangría que lo sigue como un espacio) no es el `hito`; y en las tres el cuerpo tiene una remisión `](references/<f>.md)` a cada una de las ocho y ninguna más. Mutantes que se ven en rojo y se revierten: un byte fuera de lo declarado en cada variante; un «hito» con una letra cambiada; un «antes» que no está en control; la sección de `con_herramientas.md` sin su última línea; y una línea de más en el cuerpo corto. `TestLaboratorioFueraDelProducto` sigue en verde (el `SKILL.md` y el `con_herramientas.md` que difieren entran en la comprobación de bytes) y `TestSkillsDelRepositorio` también: `normativa-remisiones-en-orden` pasa de 11 000 caracteres fuera de `skills/` (FR-007)

**Checkpoint**: las cuatro variantes existen y `make ci` las ata a `skills/normativa` y a su declaración.

---

## Phase 3: User Story 2 — El arnés de H39 lanza las variantes con las evals reales de `normativa` (Priority: P1)

**Objetivo**: `make evals-laboratorio` acepta una familia de `normativa` y la prepara con sus skills, sus evals y el
servidor de hoy.

**Independent Test**: `TestArnesDeNormativa`, sin modelo y con el sustituto de `claude`: 0 sesiones con cada petición
inválida y n/4 por eval en cada variante y modo con las válidas (SC-006).

- [X] T004 [US2] El arnés con las dos familias y su control sin modelo (FR-008, FR-010, FR-011, FR-012, FR-013; SC-006; contracts/arnes.md §2, §3 y §5; data-model §1, §4 y §5; research D2, D7, D8, D14, S3). **Código**: internal/evals/laboratorio_normativa.go gana, en la familia, los modos (`normativa-control`: orden y herramienta; las otras tres: herramienta), el directorio de evals del reparto `evals/normativa` y las cuatro evals del sondeo, en orden (`04-lgt-prescripcion.yaml`, `05-trlrhl-impuestos-municipales.yaml`, `15-irpf-rendimientos-por-materia.yaml`, `22-lcsp-contrato-menor-en-2019.yaml`); internal/evals/laboratorio_sondeo.go (y, si hace falta, internal/evals/laboratorio.go): `comprobar` busca cada nombre de `VARIANTES` en las dos familias, con el mensaje de hoy para el que no es de ninguna, `VARIANTES: se mezclan variantes del laboratorio de H39 y variantes de normativa` en una línea tras las de los nombres, y `VARIANTES y MODOS: <v> no se mide en modo <m>` por cada par que no es de su variante, en el orden de su declaración y primero los de H39; todos los errores juntos, uno por línea, sin abrir sesión, construir nada ni crear `DIRECTORIO`; con una familia de `normativa`, copia `laboratorio/normativa/<variante>/skills/` como directorio de skills de la variante, prepara las sesiones con `evals/normativa`, planifica `SESIONES/4` sesiones de cada una de las cuatro evals (`<eval sin .yaml>-<nn>`), falla antes de construir nada, con su nombre, si una de las cuatro falta, y en modo herramienta deja el `servidor.json` de hoy (`kitlegal mcp serve` del temporal, sin `laboratorio servidor`). **Test**: `TestArnesDeNormativa` en internal/evals/laboratorio_sondeo_test.go, con la preparación sin construir y el sustituto de `claude` de `TestArnesDelLaboratorio` (contracts/arnes.md §5): con `VARIANTES=normativa-otra`; con las tres de solo herramienta y `MODOS=orden` (tres líneas, una por variante en su orden); con `SESIONES=6`; con `VARIANTES=control,normativa-control`; con la versión de Claude Code distinta; y con `MODELO=sonnet`: 0 sesiones, nada construido, `DIRECTORIO` sin crear y su mensaje; con `VARIANTES=normativa-control MODOS=orden,herramienta SESIONES=4`, una sesión `<eval>-01` de cada una de las cuatro evals en cada modo y ninguna más; con las otras tres, `MODOS=herramienta SESIONES=8`, dos de cada eval en cada variante y ninguna más; en cada sesión válida, `claude/skills/` tiene solo `normativa` con el `SKILL.md` de su variante byte a byte, sin `servidor.json` en orden y con `kitlegal mcp serve` en herramienta. `TestArnesDelLaboratorio`, `TestGuionDelLaboratorio`, `TestRecetasDelSondeo` y `TestServidorDelLaboratorio`, sin cambiar ningún valor esperado. Mutantes que se ven en rojo y se revierten: `normativa-cuerpo-corto` con el modo orden; sin la comprobación de la mezcla; y el directorio de skills de otra variante en una sesión

**Checkpoint**: el sondeo de las cinco filas se puede lanzar con dos órdenes (quickstart).

---

## Phase 4: User Story 3 — La tabla dice, por variante y modo, qué se activa, qué se abre y qué se cumple (Priority: P1)

**Objetivo**: la regla testigo comprobada con respuestas a las evals de `normativa`, y `make laboratorio-tabla` con la
tabla de la familia de `normativa` sin cambiar la de H39.

**Independent Test**: `TestReglaTestigoDeNormativa`, `TestTablaDeNormativa`, `TestTablaDelLaboratorio` sin cambios y
el guion congelado `h40-laboratorio-tabla.txtar` en verde (SC-004, SC-005, SC-007).

- [X] T005 [P] [datos] [US3] Las respuestas fijas de la regla testigo y su control, en una sola tarea porque el test es el que lee esas respuestas y el guion que comprueba, `reglaTestigo`, ya existe y no cambia (FR-030; SC-004; contracts/tabla.md §6 y §7; research D11, V18). **Datos**: internal/evals/testdata/laboratorio/normativa/respuestas/ con las seis respuestas de contracts/tabla.md §6 —`cumple.txt` (eval 04), `no-cumple-sin-sello.txt` (05), `no-cumple-sello-en-negrita.txt` (15), `no-cumple-sello-entre-comillas.txt` (22), `no-cumple-sello-en-codigo.txt` (04) y `no-cumple-sello-no-al-final.txt` (05)—, cada una una respuesta con citas a su eval y la última línea no vacía de la tabla. **Test**: `TestReglaTestigoDeNormativa` en internal/evals/laboratorio_normativa_test.go, con los ayudantes de `TestReglaTestigo` sin cambiarlos: el directorio tiene exactamente esas seis, cada una con el veredicto de su nombre; y la respuesta de la sesión `05-trlrhl-impuestos-municipales-claude-sonnet-5-5-01` del informe versionado `specs/030-h32-arquitectura-de-skills/gates/evals/normativa.json`, leída del informe, que activó `normativa` y tiene respuesta, no cumple. `TestReglaTestigo` y sus respuestas de H39, sin cambios. Mutantes que se ven en rojo y se revierten: `cumple.txt` con el sello en negrita; y una séptima respuesta en el directorio
- [X] T006 [datos] [US3] La tabla de la familia de `normativa`, con sus sesiones fijas y la tabla esperada, en una sola tarea porque el control de SC-005 compara la salida del guion byte a byte con esa tabla y no hay test sin los dos (FR-020, FR-021, FR-022, FR-023, FR-030; SC-005, SC-007; contracts/tabla.md §1 a §4 y §7; data-model §1, §5 y §6; research D9, D10, V21). **Código**: internal/evals/laboratorio_tabla.go (y internal/evals/laboratorio_normativa.go, que gana en la familia la variante con regla testigo, `normativa-testigo-en-referencia`): la familia de la tabla es la de la primera variante que lee (sondeos y variantes en orden de nombre); con una de `normativa`, cuenta por sesión la activación de `normativa` antes de la primera herramienta de kitlegal de su modo y, por cada una de las ocho referencias, el turno de su primera apertura (`Read` con `file_path` `normativa/references/<f>` o que termina en `/normativa/references/<f>`, entero o por tramos, o `Bash` cuya orden contiene `references/<f>`; `Grep` y `Glob` no), y escribe las 16 columnas de §3 y el detalle de §4 con el recuento, Wilson y la mediana de hoy; «Cumplen la regla» y «Cumple» solo en `normativa-testigo-en-referencia`, vacías en las demás; con `n` = 0, las celdas vacías de §3; un directorio de variante que no es de `normativa` en una tabla de `normativa` es `laboratorio: <ruta> no es ninguna variante de normativa`, código 1 y sin tabla; con una familia de H39, la salida de hoy byte a byte. **Datos**: internal/evals/testdata/laboratorio/normativa/sondeo/ con sesiones fijas de los dos modos que cubren lo de contracts/tabla.md §7 —una referencia abierta dos veces con `Read` por tramos y otra con `Bash`; `Grep` sobre una referencia; referencias que ninguna sesión abre; `Skill` y la primera herramienta en el mismo turno; una sesión sin ninguna herramienta que activa; una sin terminar; una fila sin ninguna sesión terminada; una sesión de `normativa-testigo-en-referencia` que cumple la regla y otra que no— e internal/evals/testdata/laboratorio/normativa/tabla.md, su tabla y su detalle. **Test**: `TestTablaDeNormativa` en internal/evals/laboratorio_tabla_test.go: la salida sobre ese sondeo es, byte a byte, `tabla.md`; sobre sondeos que monta en un temporal con 40 copias de sesiones fijas, las celdas de 0, 20 y 40 de 40 son `[0,0–8,8 %]`, `[35,2–64,8 %]` y `[91,2–100,0 %]`; y con un directorio de una variante de H39 dentro de un sondeo de `normativa`, el defecto, código 1 y sin tabla. `TestTablaDelLaboratorio`, `TestOrdenesDeBash` y las sesiones fijas y la tabla de H39, sin cambiar ningún valor ni byte (FR-023). Mutantes que se ven en rojo y se revierten: contar dos veces una referencia abierta dos veces; `Skill` en el mismo turno contado como antes; y la celda de la regla a cero fuera de `normativa-testigo-en-referencia`. Antes de terminar, con un prototipo (un worktree bajo `$TMPDIR` con specs/032-h40-por-qu-e/aceptacion/laboratorio-tabla.txtar copiado en la carpeta de guiones de internal/app como `h40-laboratorio-tabla.txtar`, ejecutado con `go -C <worktree> test ./internal/app -run 'TestEntregaDelHito'`), comprobar que el guion congelado pasa sin tocarlo

**Checkpoint**: la medida de la aceptación se puede contar sin modelo; la de H39 no cambia.

---

## Phase 5: Polish & cross-cutting

- [ ] T007 La documentación de lo que entra en el repositorio (FR-010, FR-012, FR-020, FR-040; DoD §1 punto 6; research D15; plan.md «Project Structure»): en el Makefile, solo las líneas `##` de ayuda de `evals-laboratorio` y `laboratorio-tabla`, que dicen que también lanzan y cuentan las variantes de `normativa`, sin tocar ninguna receta; en CONTRIBUTING.md, «El sondeo del laboratorio»: las cuatro variantes de `normativa` con sus nombres y sus modos, las evals 04, 05, 15 y 22 de `evals/normativa`, que una invocación es de una sola familia, y los dos lanzamientos y la tabla de contracts/arnes.md §1; en cmd/laboratorio/main.go, el comentario de paquete, que su tabla cuenta las dos familias, sin cambiar código; y en CHANGELOG.md, *Unreleased*, *Added*: las variantes de `normativa` en `laboratorio/normativa/`, fuera del producto, y que `make evals-laboratorio` y `make laboratorio-tabla` las lanzan y las cuentan. Comprobar con `make help` que las dos líneas `##` salen, que `TestRecetasDelSondeo` sigue en verde con las recetas de hoy, y `make ci` en verde

---

## User Story 4 — La persona mide y decide antes de fusionar (P2)

Sin tareas: FR-040 y FR-041 se hacen fuera del run (ADR 0032), y FR-042 dice que el run no lanza el sondeo ni escribe
el ADR. El informe final lo deja pendiente por la línea `[alcance]` de FR-042 en `gates/supuestos.md`: la propuesta no
se fusiona sin el sondeo de FR-040 ni el ADR de FR-041 (research D16). Lo que se ejecuta entonces es
quickstart.md, con lo que entregan T002 a T006.

---

## Dependencies & Execution Order

### Orden de las fases

- **Fase 1 (T001)**: primero; congela la suite y da al arnés e2e `KITLEGAL_LABORATORIO_BIN`.
- **Fase 2 (T002 → T003)**: T003 calcula las tres variantes desde la `normativa-control` de T002 y amplía su test;
  `TestLaboratorioFueraDelProducto`, generalizado en T002, ve las tres en T003.
- **Fase 3 (T004)**: necesita las cuatro variantes (T003), cuyo `SKILL.md` compara en cada sesión, y la familia (T002).
- **Fase 4 (T005, T006)**: T005 solo necesita `laboratorio_normativa_test.go` de T002; T006 usa la familia (T002, T004)
  y la regla de hoy, y es la que pone en verde el guion congelado de T001.
- **Fase 5 (T007)**: al final; documenta lo de T003, T004 y T006.

### Dentro de cada tarea

Test y la implementación mínima que lo pone en verde, en la misma tarea; el mutante que ve fallar el control,
revertido; `make ci` en verde antes de marcarla.

### Oportunidades de paralelismo

- T005 [P] no comparte ficheros con T004 (`laboratorio_normativa_test.go` y `respuestas/` frente a
  `laboratorio_sondeo.go`, `laboratorio_normativa.go` y `laboratorio_sondeo_test.go`): puede ir tras T003, en paralelo
  con T004.
- Las demás comparten `internal/evals/laboratorio_normativa*.go` o dependen de lo anterior y van en serie.

## Parallel Example: tras T003

```text
T004 [US2] El arnés con las dos familias (internal/evals/laboratorio_sondeo.go, internal/evals/laboratorio_normativa.go)
T005 [P] [datos] [US3] Las respuestas fijas de la regla testigo (internal/evals/testdata/laboratorio/normativa/respuestas/, internal/evals/laboratorio_normativa_test.go)
```

## Implementation Strategy

### MVP

1. Fase 1: la suite.
2. Fase 2: las cuatro variantes y sus controles (US1). **Validar**: `TestVariantesDeNormativa` y
   `TestLaboratorioFueraDelProducto` en verde con `make ci`.

### Entrega incremental

3. Fase 3: el arnés (US2). **Validar**: `TestArnesDeNormativa`.
4. Fase 4: la regla testigo y la tabla (US3). **Validar**: `TestReglaTestigoDeNormativa`, `TestTablaDeNormativa` y el
   prototipo del guion congelado.
5. Fase 5: Makefile (ayuda), CONTRIBUTING, comentario de `cmd/laboratorio` y CHANGELOG.

## Notes

- 7 tareas: T001 [aceptacion]; US1: T002, T003; US2: T004; US3: T005, T006; transversal: T007; US4: ninguna.
- Tareas [datos]: T005 (`respuestas/`) y T006 (`sondeo/` y `tabla.md` de `normativa`); las dos dicen en su línea por qué
  van con su test (y T006 con su código). Ninguna graba ni toca una grabación: el hito no tiene datos externos (plan.md
  «Datos externos»), ni manifiesto `grabaciones.json`, ni `TestGrabar*`, ni fila nueva en `docs/SOURCES.md`.
- Definition of Done: punto 1, `make ci` en cada tarea; 2, tests offline en cada tarea; 3, R4 y R5 (convenciones); 4,
  no aplica (ningún sobre ni esquema cambia); 5, defecto `inesperado` y errores de uso (T004, T006); 6, e2e en T001 y
  CHANGELOG en T007; 7, no aplica en el run: el ADR de FR-041 se escribe fuera del run, antes de fusionar (FR-042); 8,
  no aplica (ninguna fuente); 10, no aplica: ninguna skill del producto cambia (FR-043) y las variantes no son del
  producto (FR-007); 9, 11, 12 y 13, no aplican.
- Cada fila de «Controles de umbral» de plan.md y la tarea que construye su control: SC-001/FR-002
  (`TestVariantesDeNormativa`), T002; SC-002/FR-003 a FR-006 (`TestVariantesDeNormativa`), T003; SC-003/FR-001
  (`TestLaboratorioFueraDelProducto`), T002, que T003 mantiene; FR-007 (`TestSkillsDelRepositorio`, de hoy), T002 y
  T003, con `skills/` sin cambios; SC-006/FR-008/FR-010/FR-011 (`TestArnesDeNormativa`), T004; SC-004/FR-030
  (`TestReglaTestigoDeNormativa`), T005; SC-005/FR-020 a FR-022 (`TestTablaDeNormativa`), T006; SC-005/SC-007 y
  FR-020/FR-021/FR-023/FR-030 (el guion `h40-laboratorio-tabla.txtar` en `TestEntregaDelHito`), T001, que lo ve fallar
  hoy, y T006, que lo pone en verde; SC-007/FR-023 (`TestTablaDelLaboratorio`, de hoy), T006, sin cambiar ningún valor;
  SC-008/FR-043, las tres filas `evals:`, las mide el job de evals del cierre, que hace el workflow tras la revisión
  final con los umbrales de hoy, sin tarea que los toque.
- Requisitos sin tarea propia: FR-040 y FR-041 (fuera del run), FR-042 (su línea `[alcance]` en `gates/supuestos.md`),
  FR-043 (su línea `[alcance]` en `gates/supuestos.md`, y el cierre del workflow), SC-008 (el cierre del workflow) y
  SC-009 (capa 3). FR-012 lo cumplen T004 (el arnés no lo ejecuta `make ci`: solo el sustituto) y T007 (la cadena documentada, sin recetas nuevas); FR-013, todas (ninguna graba ni escribe fuera del
  repositorio y `$TMPDIR`).
