---
description: "Tareas de H39 · Cómo se organiza una skill para que se active y se lea en los dos modos"
---

# Tasks: H39 · Cómo se organiza una skill para que se active y se lea en los dos modos

**Input**: `specs/031-h39-c-omo-se/` — [spec.md](spec.md), [plan.md](plan.md), [research.md](research.md),
[data-model.md](data-model.md), [quickstart.md](quickstart.md) y [contracts/](contracts/) (`esquemas.md`,
`laboratorio.md`, `arnes.md`, `tabla.md`).

**Tests**: obligatorios (constitución III). La primera tarea escribe la suite de aceptación, que el workflow congela;
cada tarea posterior trae su test y la implementación que lo pone en verde, y deja `make ci` en verde por sí sola.

**Organización**: una fase por historia, en el orden en que se ejecutan una tras otra. La línea de cada tarea declara
todas las rutas que crea o modifica (el guardián rechaza cualquier otra; `go.mod`, `go.sum`, `CHANGELOG.md` y el
directorio del feature siempre se admiten, y declarar `x.go` admite `x_test.go`).

## Format: `[ID] [P?] [Story] Descripción`

- **[P]**: puede ir en paralelo (ficheros distintos, sin dependencia de una tarea sin terminar).
- **[Story]**: la historia de [spec.md](spec.md) a la que sirve (US1 a US3; US4 no tiene tareas).
- **[aceptacion]**: la suite congelada (solo T001). **[datos]**: la tarea toca ficheros bajo `testdata/` o `schemas/`.

## Convenciones de todas las tareas

- `make ci` en verde al terminar cada una, en primer plano y por objetivos si hace falta (`make fmt-check`,
  `make lint`, `make test`, `make test-integration`, `make vuln`, `make schema-check`, `make skills-check`…). Ninguna
  usa la red más que la de las herramientas de Go, ninguna graba, ninguna abre una sesión con modelo ni ejecuta
  `TestSondeoDelLaboratorio`, `make evals-laboratorio` ni los guiones `scripts/evals*.sh` desde la terminal (los
  ejecutan los tests de Go con sustitutos, como hoy `TestGuionDelSondeo`).
- `skills/` no se toca en ninguna tarea (FR-004, SC-002), ni `internal/mcp/` (FR-029: las instrucciones de hoy).
- Los textos de la skill de laboratorio, de sus variantes y de la regla testigo son los de
  [contracts/laboratorio.md](contracts/laboratorio.md); los mensajes y la cadena del arnés, los de
  [contracts/arnes.md](contracts/arnes.md); el formato de la tabla, el de [contracts/tabla.md](contracts/tabla.md).
  Nada se inventa fuera de ellos.
- `internal/evals` no nombra `os.Stdin`, `os.Stdout`, `os.Stderr` ni llama a `os.Exit` (R4, R5): recibe los
  `io.Reader`/`io.Writer` y devuelve el código que `cmd/laboratorio/main.go` pasa a `os.Exit`, como `cmd/empaquetar`.
  Ningún `panic`; lo que llega de fuera manipulado es defecto `inesperado` (1).
- Los fixtures nuevos no llevan nada con forma de secreto (gitleaks corre en `make ci`).

---

## Phase 1: Setup — la suite de aceptación

**Propósito**: fijar, antes de cualquier código, lo que la entrega visible del binario tiene que cumplir (ADR 0018).

- [X] T001 [aceptacion] [US1] Escribir, a partir del spec y de [contracts/esquemas.md](contracts/esquemas.md) §1 y §2 y sin código de producto, el guion testscript specs/031-h39-c-omo-se/aceptacion/describe.txtar (el workflow lo activa como `h39-describe.txtar` al terminar el bucle), con un comentario de cabecera «Cubre: US1.1, US1.3 · FR-001, FR-003 · SC-001 (en `--describe`); FR-002 lo comprueba `TestAyudasDeLosArgumentos` (research D12)», con la convención de invocación de los guiones `h32-*.txtar` y estas aserciones: (1) `exec kitlegal boe articulo BOE-A-2015-10565 a21 --describe` sale con 0 y su stdout lleva, una vez cada una, las líneas `"description": "<ayuda>"` de `norma`, `bloque` y `fecha` con su ayuda carácter a carácter (las de las etiquetas `help` de los argumentos de `boe articulo`, con `<año>` y `<número>` sin escapar, research V4: «Id del bloque, del campo id de kitlegal boe indice (a21, da-3, a1-30…): no se compone con el número del artículo, que no siempre es su id.» en `bloque`), escapando en el patrón los metacaracteres de la expresión regular; (2) en esa misma salida, la propiedad de cada una de las ocho banderas globales (`json`, `timeout`…) sale con los mismos bytes que hoy, sin `description` (copiar su forma de la salida de esa orden con el binario construido del árbol de hoy, FR-003); (3) `exec kitlegal plazos calcular x x x --describe` sale con 0 y lleva la `description` de `tipo`, `inicio`, `municipios`, `duracion` y `unidad` con las ayudas de las etiquetas `help` de `plazos calcular`; (4) `exec kitlegal skills install --describe` sale con 0 y lleva en `global` «Actúa en el ámbito global: $HOME/.agents/skills y los hosts de HOME.» con `$HOME` literal (escapado para que testscript no lo expanda, comprobado contra la salida). Antes de terminar, comprobar con un prototipo (un worktree bajo `$TMPDIR` con el guion activado, ejecutado con `go -C`) que el guion falla hoy por una aserción de (1), (3) o (4) y no por un `exec`; no escribe evals de ninguna skill del producto, porque ninguna cambia (FR-004)

**Checkpoint**: la suite queda congelada; ninguna tarea posterior la toca.

---

## Phase 2: User Story 1 — La ayuda de cada argumento llega a su herramienta (Priority: P1) 🎯 MVP

**Goal**: la ayuda de cada argumento o bandera propia de un verbo es la `description` de su propiedad en `--describe`
y en el `inputSchema` de su herramienta MCP; nada más cambia.

**Independent Test**: `go run ./cmd/kitlegal boe articulo BOE-A-2015-10565 a21 --describe` lleva las tres
`description` (quickstart §1); `TestAyudasDeLosArgumentos`, `make schema-check` y `make skills-check` en verde con
`skills/` sin cambios (quickstart §2).

- [X] T002 [datos] [US1] `describirArgumentos` y los ficheros que fijan los bytes de `--describe`, en una sola tarea porque en cuanto `internal/cli` emite la `description`, `TestEsquemasPublicados` (`make schema-check`), el golden de plazos y la sección `main-articulos.describe` de `h20-fecha-main.txtar` salen en rojo hasta regenerarse, y `make ci` tiene que quedar en verde (FR-001, FR-002, FR-003, FR-004, FR-005; SC-001, SC-002; contracts/esquemas.md §1, §2 y §4; research D1, D3, V1, V3, V8). **Código**: en internal/cli/describe.go, `describirArgumentos(entrada, campos)` pone en la propiedad de cada campo **del verbo** con etiqueta `help` no vacía su valor tal cual, detrás de las claves que ya lleva; la llama `esquema()` tras `anotarBanderas` (las ocho globales no pasan por ella) y la llama `EsquemasDeHerramienta` en internal/cli/herramienta.go. **Test**: `TestAyudaComoDescripcion` en internal/cli/describe_test.go, con structs de prueba (un argumento con ayuda que lleva `<`, `>` y `$`, uno sin ayuda, uno con ayuda vacía y las globales) en los dos esquemas: la ayuda sin reescribir, sin `description` donde no hay ayuda ni en las globales, y el resto del esquema igual; actualizar las expectativas de los tests de internal/cli que fijan el esquema de un struct con `help` (internal/cli/describe_test.go, internal/cli/herramienta_test.go y, si alguno lo fija, internal/cli/parse_test.go, internal/cli/literal_test.go o internal/cli/supuestos_kong_test.go), y la de `compruebaEntradaDeCheck` en internal/app/grafo_test.go, que fija sin `description` las propiedades `norma` y `bloques` de la entrada de `graph check --describe` y deja `TestAppletGrafo` en rojo (redelimitada en el intento 1, gates/tarea-T002.md). **Datos**: antes de regenerar, comprobar que `make schema-check` sale en rojo (el control ve la diferencia); regenerar schemas/norma.json, schemas/bloque.json, schemas/municipio.json, schemas/instalacion.json, schemas/grafo.json, schemas/cita.json y schemas/plazo.json con la bandera `-actualizar-esquemas` de `TestEsquemasPublicados` (contracts/esquemas.md §4; `servidor.json` no cambia), internal/app/testdata/plazos/describe.json con `-actualizar-golden-de-plazos`, y la sección `main-articulos.describe` de internal/app/testdata/script/h20-fecha-main.txtar con la salida del binario del árbol; `make skills-check` en verde y `git status --short` sin ningún fichero de las skills del producto. **CHANGELOG.md**: en *Unreleased*, *Changed*: «Los parámetros de `--describe` y de las herramientas del servidor MCP llevan como `description` la ayuda de su argumento» (FR-005)
- [X] T003 [US1] El control de SC-001 en internal/app/ayudas_test.go, `TestAyudasDeLosArgumentos`, sobre el código de T002 (FR-001, FR-002, FR-003; SC-001; contracts/esquemas.md §5; research D2, V5, V6): recorre los verbos de `RegistroDeProduccion`; de cada uno toma la ayuda de cada argumento y bandera propios del modelo que Kong construye con la gramática del applet (la que imprime `--help`), y la compara con la `description` de su propiedad en el documento de `--describe` (orden en proceso, con un valor por cada argumento obligatorio) y, en los verbos con herramienta, en el `inputSchema` de `tools/list` del servidor en proceso con `mcptest`; falla con el primer argumento cuya `description` falta o difiere, con una propiedad sin ayuda que la lleva, con una global que la lleva, y si cuenta menos de las 28 ayudas de hoy (22 en verbos con herramienta). Comprobar con un mutante temporal (quitar la llamada a `describirArgumentos` de `EsquemasDeHerramienta`, y aparte la de `esquema()`) que el test falla en cada forma, y revertirlo antes de `make ci`

**Checkpoint**: US1 entregada; `h39-describe.txtar` pasaría con el binario del árbol.

---

## Phase 3: User Story 2 — Una skill de laboratorio con su regla testigo y sus variantes (Priority: P1)

**Goal**: la skill `articulos-del-boe`, su skill de entrada `consulta-legal`, las ocho variantes y las cuatro evals, en
`laboratorio/`, fuera del producto, con la regla testigo y su guion.

**Independent Test**: `go test -count=1 -run '^(TestReglaTestigo|TestVariantesDelLaboratorio|TestEvalsDelLaboratorio|TestLaboratorioFueraDelProducto)$' ./internal/evals/`
en verde (quickstart §2 y §4).

- [X] T004 [P] [datos] [US2] La regla testigo y su guion con sus respuestas fijas, en una sola tarea porque el test es el que lee esas respuestas y sin él no hay control (FR-013, FR-014; SC-003; contracts/laboratorio.md §3; data-model §5; research D6, V20). **Código**: internal/evals/laboratorio.go, nuevo, con la línea `Sello de lectura: KL-7Q4` y el párrafo de la regla como constantes y `reglaTestigo(respuesta string) bool`: cumple si la última línea no vacía, sin espacios ni `\r` a los lados, es exactamente esa línea. **Datos**: en internal/evals/testdata/laboratorio/respuestas/, `cumple.txt`, `cumple-con-lineas-en-blanco.txt`, `no-cumple-sin-sello.txt`, `no-cumple-sello-no-al-final.txt`, `no-cumple-sello-en-negrita.txt` (`**Sello de lectura: KL-7Q4**`) y `no-cumple-sello-con-texto.txt` (`Sello de lectura: KL-7Q4.`). **Test**: `TestReglaTestigo` en internal/evals/laboratorio_test.go: cada respuesta fija con el veredicto de su nombre; la respuesta del producto, la de la sesión `09-constitucion-articulo-140-claude-sonnet-5-5-01` leída del informe de evals de `normativa` que versionó H32 (research V20), no cumple; y la línea del sello no aparece en ningún `SKILL.md` ni referencia de las skills del producto ni en `mcp.Instrucciones`. Ver fallar el test con un mutante temporal de `reglaTestigo` (que acepte el sello en cualquier línea) y revertirlo
- [X] T005 [US2] La declaración, la skill de control y las siete variantes, con su control (FR-011, FR-020, FR-021, FR-022, FR-023, FR-024, FR-025, FR-026, FR-027, FR-028; SC-004; contracts/laboratorio.md §1, §2, §4 y §5; data-model §2 y §3; research D4, D5, V11). **Laboratorio**: laboratorio/variantes.yaml con el texto de contracts/laboratorio.md §2; laboratorio/variantes/control/skills/articulos-del-boe/SKILL.md (frontmatter con `name` y la `description` de §4, sin `metadata`; el cuerpo de §4 con el párrafo de la regla tras la frase de la cita), y en laboratorio/variantes/ las variantes `instrucciones` (`instrucciones.txt`: el texto de `mcp.Instrucciones`, un espacio, «Antes de la primera herramienta de kitlegal, activa la skill articulos-del-boe.» y un salto de línea), `entrada` (con el `SKILL.md` de la skill `consulta-legal` de §5), `referencia-secundaria`, `parametro` (`parametro.txt`: la ayuda de `bloque`, un espacio, el párrafo de la regla y un salto de línea), `referencia-paso`, `referencia-en-el-punto` (las tres con `cierre.md` en `references`: «# Cómo termina la respuesta», una línea vacía y el párrafo de la regla) y `descripcion-sin-herramientas`, con la estructura de contracts/laboratorio.md §1, cada una copia completa de la de control salvo lo que §5 cambia. **Código**: en internal/evals/laboratorio.go, la lectura estricta de la declaración (`skills.ValidarDocumentoYAML`: claves desconocidas y repetidas rechazadas) y las partes de data-model §3 (`descripcion`, `cuerpo`, `referencia`, `entrada`, `instrucciones`, `parametro`). **Test**: `TestVariantesDelLaboratorio` en internal/evals/laboratorio_test.go: la declaración con las ocho variantes en orden, sus modos y sus partes; `skill`, `entrada` y `herramienta` (que está en `app.NombresDeHerramientas`) y `parametro` (propiedad de su entrada); cada parte no declarada idéntica byte a byte a la de control, cada declarada distinta y ningún fichero fuera de una parte; la regla donde la pone cada variante y no en el cuerpo de las que la sacan; la `description` de control nombra alguna herramienta y la de `descripcion-sin-herramientas` ninguna (ni `kitlegal <applet>`); `instrucciones.txt` nombra la skill y el cuerpo de `consulta-legal` también; y la forma del producto en cada variante con `skills.Regenerar` y `skills.Comparar` sobre el directorio de la variante (cuerpo ≤ 11 000 caracteres, descripción ≤ 900, `references/` plano y enlazado, sin defectos ni deriva). Ver fallar el test cambiando un byte de una parte no declarada de una variante y, aparte, igualando una declarada, y revertirlo
- [X] T006 [US2] Las cuatro evals de laboratorio y su control (FR-012; contracts/laboratorio.md §6; data-model §4; research V19): laboratorio/evals/01-constitucion-articulo-140.yaml, laboratorio/evals/02-lgt-prescripcion.yaml, laboratorio/evals/03-lrjsp-principio-de-legalidad.yaml y laboratorio/evals/04-et-vacaciones.yaml, con la pregunta, `activa: true`, `comandos` y `citas` de las evals 09, 04, 07 y 10 de `normativa` (research V19) y el comentario de cabecera de §6; y `TestEvalsDelLaboratorio` en internal/evals/laboratorio_test.go (con lo que necesite de internal/evals/laboratorio.go): `LeerConjunto` las lee bien formadas, son cuatro, cada comando es una lectura de la `herramienta` de la declaración, y `Preparar` con la grabación de hoy (`UnionDeGrabaciones()`) no da faltas y `ComprobarSinRed` no da fallos; ver fallar el test con una eval temporal que pide un bloque que la grabación no tiene, y quitarla. Sin grabar nada
- [X] T007 [US2] El laboratorio fuera del producto: `TestLaboratorioFueraDelProducto` en internal/evals/laboratorio_test.go, con lo que necesite de internal/evals/laboratorio.go, sobre el código de hoy y lo de T005 y T006 (FR-010; SC-005; contracts/laboratorio.md §7; research V9, V10): ningún fichero de `kitlegal.Skills()` tiene la ruta de una skill del laboratorio (`articulos-del-boe`, `consulta-legal`) ni los bytes de un fichero del laboratorio; `app.Arrancar` con `skills install --dir <directorio vacío del test>` no deja ninguno; la definición del job de evals no ejecuta ninguna de las dos skills y no hay carpeta de evals del producto con el nombre de ninguna; y ninguna ruta del laboratorio mide (`rutaQueMide`). Ver fallar cada comprobación con un mutante temporal (una ruta del laboratorio en la lista que se mira) y revertirlo

**Checkpoint**: US2 entregada; el laboratorio existe, cumple la forma y no llega al producto.

---

## Phase 4: User Story 3 — El arnés del sondeo y el guion que lo cuenta (Priority: P1)

**Goal**: el servidor de las sesiones de `instrucciones` y `parametro` sin tocar `kitlegal`, el guion que cuenta un
sondeo con Wilson, y el arnés que lanza las sesiones por variante y modo con sus rechazos antes de abrir nada.

**Independent Test**: `go run ./cmd/laboratorio tabla internal/evals/testdata/laboratorio/sondeo | diff - internal/evals/testdata/laboratorio/tabla.md`
sin diferencias (quickstart §3); `TestServidorDelLaboratorio`, `TestTablaDelLaboratorio` y `TestArnesDelLaboratorio`
en verde (quickstart §2).

- [X] T008 [P] [US3] El servidor de las sesiones de `instrucciones` y `parametro` y la orden de `servidor.json` (FR-021, FR-024, FR-029; contracts/arnes.md §3.4 y §5; research D7, V12, V13, V14, S3). **Código**: internal/evals/sesiones.go, `SesionesAEjecutar` gana `Servidor []string` (la orden de `servidor.json`; vacío, la de hoy, `Binario mcp serve`, byte a byte) y `escribirElServidor` la usa; internal/evals/laboratorio_servidor.go, nuevo, el filtro de `laboratorio servidor -kitlegal <ruta> [-instrucciones <texto>] [-herramienta <h> -parametro <p> -descripcion <texto>]`: arranca `<ruta> mcp serve` con el entorno del proceso (`os/exec`, sin el SDK de MCP, R7), copia su entrada al hijo sin tocarla, su salida de error a la propia, y de su salida, línea a línea, reescribe solo `result.instructions` de la respuesta a `initialize` y la `description` de `inputSchema.properties.<p>` de la herramienta `<h>` en la de `tools/list`, y deja las demás líneas byte a byte; termina con el código del hijo; argumentos sin esa forma, o una herramienta o un parámetro que la lista no trae, son defecto: lo dice en su salida de error y sale con 1; cmd/laboratorio/main.go, nuevo, con la orden `servidor` (y el uso con 1 ante una orden desconocida), que pasa `os.Stdin`, `os.Stdout`, `os.Stderr` y `os.Args` al paquete de evals y llama a `os.Exit` con su código. **Test**: `TestServidorDelLaboratorio` en internal/evals/laboratorio_servidor_test.go, con las líneas que escribe el servidor de producción en proceso (`mcptest.Sesion.Lineas`): sin opciones, la salida igual byte a byte; con `-instrucciones`, solo cambia ese campo; con `-herramienta boe_articulo -parametro bloque -descripcion …`, solo cambia esa `description` (la línea reescrita, decodificada, igual salvo ese campo); con una herramienta o un parámetro que no existen, 1 y su mensaje; y en internal/evals/sesiones_test.go, el caso de `servidor.json` con `Servidor` dado y el de hoy, igual byte a byte, con `Servidor` vacío. `TestInstrucciones` sigue en verde sin cambios (FR-029)
- [X] T009 [datos] [US3] El guion que cuenta un sondeo, con las sesiones fijas y la tabla esperada, en una sola tarea porque el control de SC-006 compara la salida del guion byte a byte con esa tabla y no hay test sin los dos (FR-040, FR-041, FR-042, FR-043; SC-006; contracts/tabla.md; data-model §6 y §7; research D9, D10, D11, V17, V18, V21; Clarifications Q1 a Q4). **Código**: internal/evals/laboratorio_tabla.go, nuevo: lee `<ruta>` como un sondeo si tiene `sondeo.json` o cada subdirectorio que lo tenga, en orden de nombre; de cada sesión, `LeerSesion` (terminada y con respuesta) y un lector propio de turnos del transcript (turno nuevo con el primer `assistant` y con cada `assistant` tras un `user` con algún `tool_result` y con `message.id` distinto del anterior); activa (`tool_use` de `Skill` con `input.skill` igual al nombre); primera herramienta en modo herramienta (nombre, o lo que sigue a su último `__`, en `app.NombresDeHerramientas`; `ToolSearch` no); primera orden en modo orden (`Bash` partida por `|`, `||`, `&&`, `;`, `&` y saltos de línea y en palabras con las comillas y la barra de sh, primera palabra que no es `NOMBRE=valor` igual a `kitlegal` o terminada en `/kitlegal`, y argv que para `InterpretarInvocacion` es de un applet, consulta y da una herramienta del registro); apertura de la referencia (`Read` cuyo `file_path` termina en `articulos-del-boe/references/cierre.md` o `Bash` que contiene `references/cierre.md`; `Grep` y `Glob` no); `reglaTestigo` sobre la respuesta; agrupa por variante, modo, modelo y versión; y escribe en el `io.Writer` dado la tabla y el detalle de contracts/tabla.md §3 y §4, con «`<k> de <n> · <p> % [<lo>–<hi> %]`», coma decimal, Wilson al 95 % (z = 1,96) recortado a [0, 100] antes de redondear, la mediana del turno entera o con «,5» y las celdas vacías de §3; sin ningún sondeo o con algo que no sabe leer, lo dice con la ruta en la salida de error y devuelve 1 sin tabla. cmd/laboratorio/main.go gana la orden `tabla <ruta>`; el Makefile gana `laboratorio-tabla` (`go run ./cmd/laboratorio tabla "$(DIRECTORIO)"`, con su línea `##` para `make help`, fuera de `ci`). **Datos**: internal/evals/testdata/laboratorio/sondeo/, uno o dos sondeos con `sondeo.json` y sesiones fijas (`sesion.jsonl`, y lo que `LeerSesion` necesite) que cubren lo de contracts/tabla.md §6 —los dos modos; `Skill` y la primera herramienta en la misma respuesta en dos mensajes `assistant`; `Skill` y `ToolSearch` en un turno y `boe_articulo` en el siguiente; una orden `kitlegal` en una tubería y otra en una secuencia; `which kitlegal`, `--describe` y `skills list` antes; una sesión que activa sin ninguna herramienta; la referencia abierta dos veces, con `Read` por tramos y con `Bash`; `Grep` sobre ella; una variante con referencia que ninguna sesión abre; una sesión sin terminar; y una fila sin ninguna sesión terminada—, y internal/evals/testdata/laboratorio/tabla.md, la salida esperada escrita a partir del contrato y de esas sesiones (los extremos de Wilson calculados aparte con `awk` en `$TMPDIR`). **Test**: `TestTablaDelLaboratorio` en internal/evals/laboratorio_tabla_test.go: la salida sobre `sondeo/`, byte a byte, es `tabla.md`; sobre sondeos que el test monta en un temporal con 40 copias de transcripciones fijas, las celdas de 0, 20 y 40 de 40 son `[0,0–8,8 %]`, `[35,2–64,8 %]` y `[91,2–100,0 %]`; un transcript truncado y un directorio sin sondeo dan 1 sin tabla. Ver fallar el test con un mutante temporal (contar como antes la activación del mismo turno) y revertirlo
- [X] T010 [US3] El arnés del sondeo, `sondearElLaboratorio`, y su control sin modelo (FR-030, FR-031, FR-033; SC-007; contracts/arnes.md §2, §3 y §4; data-model §6; research D8, V12, V22, S1). **Código**: internal/evals/laboratorio_sondeo.go, nuevo: comprueba juntos, uno por línea y con el nombre del argumento de `make`, `VARIANTES`, `MODOS`, cada par variante × modo, `SESIONES` (múltiplo de 4, ≥ 4), `MODELO` (`^claude-[a-z]+-[0-9]+(-[0-9]+)*$`), `CLAUDE_CODE`, `CONCURRENCIA` y `DIRECTORIO` (no vacío y no existe), con los mensajes de contracts/arnes.md §2; detrás, la credencial (el mensaje de hoy, `errSinSuscripcion`) y el primer campo de `claude --version` igual a `CLAUDE_CODE`; cualquiera de ellos es error de uso: ninguna sesión, nada construido y sin `DIRECTORIO`. Si valen, prepara en el temporal `kitlegal` y `laboratorio` (`go install -trimpath`, preparación inyectable para el test), `home/` y la copia de las skills de cada variante; crea `DIRECTORIO/` y `DIRECTORIO/sondeo.json` (`modelo`, `claude_code`); y por cada variante en el orden de la declaración y cada modo (`orden` antes que `herramienta`), `ejecutarSesiones` con `Sesiones` = `DIRECTORIO/<variante>/<modo>/`, `Skills` = la copia de la variante, `Evals` = las evals del laboratorio, el guion de sesión y el tope de hoy, la `CONCURRENCIA` y `SESIONES/4` sesiones de cada eval de nombre `<eval sin .yaml>-<nn>`; en modo herramienta, `Servidor` solo en `instrucciones` (`-instrucciones` con el texto de `instrucciones.txt`) y `parametro` (`-herramienta boe_articulo -parametro bloque -descripcion` con el de `parametro.txt`); tras un reparto con sesiones sin abrir por el límite de uso, no abre más; y escribe la salida de contracts/arnes.md §4. **Test**: `TestArnesDelLaboratorio` en internal/evals/laboratorio_sondeo_test.go (con los sustitutos de internal/evals/sustitutos_test.go si los reutiliza), con la preparación sustituida y un sustituto de `claude` en el `PATH` que contesta a `--version` y anota cada sesión que abre: con la versión distinta, con `MODELO=sonnet`, con `VARIANTES=instrucciones MODOS=orden` y con `SESIONES=6`, 0 sesiones, ningún `DIRECTORIO` y el mensaje de §2; con una petición válida (dos variantes, una con un modo y otra con dos, `SESIONES=8`), dos sesiones de cada eval en cada variante y modo pedidos, en `DIRECTORIO/<variante>/<modo>/<eval>-<nn>/`, ninguna más, `sondeo.json` con el modelo y la versión, y `servidor.json` con la orden del filtro solo en `instrucciones` y `parametro`. Ver fallar el test con un mutante temporal (abrir antes de comprobar la versión) y revertirlo
- [X] T011 [US3] La cadena con la que se lanza el arnés, que ni `make ci`, ni un gancho, ni el job de evals, ni un paso del workflow ejecutan (FR-030, FR-031, FR-032, FR-033; SC-007; contracts/arnes.md §1; research D8, V16). **Código**: en internal/evals/job_test.go (`//go:build evals`), `TestSondeoDelLaboratorio`, punto de entrada que lee sus argumentos tras `-args` (`-temporal` y los siete de `make`, reutilizando las banderas de `TestSondeo` cuando coinciden en nombre) y llama a `sondearElLaboratorio`, dejando un error de uso en `uso.txt` y la salida en `salida.txt` del temporal como `TestSondeo`; scripts/evals-laboratorio.sh, nuevo, como el guion del sondeo de hoy: crea su temporal `kitlegal-laboratorio.XXXXXX` y lo borra al salir, resuelve `DIRECTORIO` relativo desde la raíz, ejecuta la orden `go test -tags evals -count=1 -timeout 0 -run '^TestSondeoDelLaboratorio$'` del paquete de evals con `-args …` (contracts/arnes.md §1) con su registro en `go-test.log`, e imprime `uso.txt` en la salida de error con 1, el registro con 1, o `salida.txt` con 0; scripts/evals-sondeo-llavero.sh gana un primer argumento de una lista cerrada (`sondeo` → el guion del sondeo de hoy con sus cinco, `laboratorio` → el nuevo con los siete; otro, su uso y 1) sin cambiar cómo toma la credencial; el Makefile gana `evals-laboratorio` (`@scripts/evals-sondeo-llavero.sh laboratorio "$(VARIANTES)" "$(MODOS)" "$(SESIONES)" "$(MODELO)" "$(CLAUDE_CODE)" "$(CONCURRENCIA)" "$(DIRECTORIO)"`, con su comentario y su línea `##`, fuera de `ci`) y `evals-sondeo` llama al guion del llavero con `sondeo` delante. **Test**: `TestGuionDelLaboratorio` en internal/evals/laboratorio_sondeo_test.go, como `TestGuionDelSondeo`, con el sustituto de go delante en el `PATH` y un `TMPDIR` vacío: con siete argumentos ejecuta la orden `go test` de arriba con ellos tras `-args` y `-temporal`; con `uso.txt`, sale con 1 e imprime solo el uso; si `go test` falla, su registro con 1; si no, `salida.txt` con 0; sin siete argumentos, su uso y 1 sin ejecutar go; y el `TMPDIR` del test queda vacío. `TestRecetasDelSondeo` en internal/evals/laboratorio_sondeo_test.go, en `make ci` y sin modelo, el único control de las dos recetas: ejecuta con os/exec, con el directorio de trabajo en la raíz del repositorio y sin `MAKEFLAGS`, `MFLAGS` ni `MAKELEVEL` en su entorno (para que el `make` de `make ci` no le pase los suyos), `make -n evals-sondeo SKILL=normativa EVALS=01 MODELO=claude-sonnet-5-5 REPETICIONES=2 CONCURRENCIA=1` y `make -n evals-laboratorio VARIANTES=control,instrucciones MODOS=orden,herramienta SESIONES=8 MODELO=claude-sonnet-5-5 CLAUDE_CODE=2.1.284 CONCURRENCIA=2 DIRECTORIO=sondeo-de-prueba` (`-n` solo imprime: ni guion, ni credencial, ni sesión), y exige en cada salida una sola línea que empieza por `scripts/evals-sondeo-llavero.sh` e igual, carácter a carácter, a la orden de contracts/arnes.md §1 con esos valores: `scripts/evals-sondeo-llavero.sh sondeo` seguido de sus cinco argumentos entre comillas dobles en el orden `SKILL`, `EVALS`, `MODELO`, `REPETICIONES`, `CONCURRENCIA`, y `scripts/evals-sondeo-llavero.sh laboratorio` seguido de sus siete en el orden `VARIANTES`, `MODOS`, `SESIONES`, `MODELO`, `CLAUDE_CODE`, `CONCURRENCIA`, `DIRECTORIO`; las demás líneas (la receta de `check-tools`) no cuentan. Ver fallar `TestRecetasDelSondeo` con un mutante temporal de la receta de `evals-sondeo` en el Makefile (sin `sondeo` delante) y revertirlo; el test y su orden se escriben con Write o Edit, no desde la terminal, porque la política de las sesiones del run veda `make … evals-sondeo` como orden (scripts/workflow/politica-paso.sh). Comprobar además que `go vet -tags evals` del paquete de evals compila, que `bash -n` acepta los dos guiones y que `make -n evals-laboratorio …`, que la política admite, da la orden de arriba, y con `git grep` que ninguna definición de flujo de GitHub, ningún gancho y ningún guion del workflow `hito` nombran `evals-laboratorio` ni `TestSondeoDelLaboratorio` (FR-032)

**Checkpoint**: US3 entregada; quickstart §3 sale como dice, y §5 y §6 quedan listos para quien los lanza fuera del run.

---

## Phase 5: Polish & cross-cutting

- [ ] T012 La documentación de lo que entra en el repositorio (FR-030, FR-032, FR-040; DoD §1 punto 6; plan.md «Project Structure»): CONTRIBUTING.md, en la tabla de objetivos, `make evals-laboratorio` (sesiones con modelo que consumen la suscripción de quien lo lanza; no es un veredicto; ningún paso del workflow `hito` lo lanza) y `make laboratorio-tabla` (sin red, sin modelo y sin credencial), y en «Sondeo local» un apartado del sondeo del laboratorio con su orden, sus siete argumentos y lo que se comprueba antes de abrir ninguna sesión (contracts/arnes.md §1 y §2), la tabla (contracts/tabla.md §1) y el primer argumento nuevo del guion del llavero; y CHANGELOG.md, en *Unreleased*, *Added*: la skill de laboratorio, en su directorio de primer nivel y fuera del producto, con sus ocho variantes y cuatro evals, `make evals-laboratorio` y `make laboratorio-tabla`. Comprobar con `make help` que las dos líneas `##` salen y con `make ci` en verde

---

## User Story 4 — La persona mide y decide antes de fusionar (P2)

Sin tareas: FR-050 y FR-051 los hace una persona fuera del run (ADR 0032), y FR-052 dice que el run no lanza el sondeo
ni escribe el ADR. El informe final lo deja pendiente por `gates/supuestos.md`, que ya lo anota (FR-052, FR-053;
research D13). Lo que la persona ejecuta es quickstart §5 y §6, con lo que entregan T008 a T011.

---

## Dependencies & Execution Order

### Orden de las fases

- **Fase 1 (T001)**: primero; congela la suite.
- **Fase 2 (T002 → T003)**: T003 mide el código de T002. `h39-describe.txtar` pasa desde T002.
- **Fase 3 (T004 → T005 → T006 → T007)**: T005 usa la regla de T004 para comprobar dónde la pone cada variante; T006
  usa la `herramienta` de la declaración de T005; T007 mira lo que dejan T005 y T006.
- **Fase 4 (T008 → T009 → T010 → T011)**: T009 añade su orden a cmd/laboratorio/main.go de T008 y usa la regla (T004)
  y la declaración (T005); T010 usa `Servidor` (T008), la declaración y los textos de las variantes (T005) y las evals
  (T006); T011 lanza `sondearElLaboratorio` de T010.
- **Fase 5 (T012)**: al final; documenta los objetivos de T009 y T011.

### Dentro de cada tarea

Test y la implementación mínima que lo pone en verde, en la misma tarea; el mutante que ve fallar el control, revertido;
`make ci` en verde antes de marcarla.

### Oportunidades de paralelismo

- T004 [P] no comparte ficheros con T002 y T003 (internal/evals frente a internal/cli e internal/app): la fase 3 puede
  empezar tras T001.
- T008 [P] solo necesita T001: no toca `laboratorio/` ni lo de T004 a T007 (`laboratorio_servidor.go`, `sesiones.go` y
  cmd/laboratorio/main.go, nuevo).
- Las demás comparten `internal/evals/laboratorio*.go`, cmd/laboratorio/main.go o el Makefile y van en serie.

## Parallel Example: tras T001

```text
T002 [datos] [US1] describirArgumentos y los ficheros que fijan los bytes de --describe (internal/cli/, schemas/…)
T004 [P] [datos] [US2] La regla testigo y su guion (internal/evals/laboratorio.go, internal/evals/testdata/laboratorio/respuestas/)
T008 [P] [US3] El servidor de las sesiones de instrucciones y parametro (internal/evals/laboratorio_servidor.go, internal/evals/sesiones.go, cmd/laboratorio/main.go)
```

## Implementation Strategy

### MVP

1. Fase 1: la suite.
2. Fase 2: la ayuda de cada argumento en `--describe` y en las herramientas MCP (US1). **Validar**: quickstart §1 y
   el guion `h39-describe` con un `make ci` en verde. Es la mejora del producto que el cierre mide con las evals de
   todas las skills (FR-053).

### Entrega incremental

3. Fase 3: el laboratorio y sus controles (US2). **Validar**: quickstart §4.
4. Fase 4: servidor, tabla y arnés (US3). **Validar**: quickstart §2 y §3.
5. Fase 5: CONTRIBUTING y CHANGELOG.

## Notes

- 12 tareas: T001 [aceptacion]; US1: T002, T003; US2: T004, T005, T006, T007; US3: T008, T009, T010, T011; transversal:
  T012; US4: ninguna.
- Tareas [datos]: T002 (`schemas/` y los golden de `internal/app/testdata/`), T004 (`respuestas/`) y T009 (`sondeo/` y
  `tabla.md`); las tres mezclan código y dicen por qué en su línea («en una sola tarea porque…»). Ninguna graba ni toca
  la grabación de hoy: el hito no tiene datos externos (plan.md «Datos externos»), ni manifiesto `grabaciones.json`, ni
  `TestGrabar*`, ni fila nueva en `docs/SOURCES.md`.
- Definition of Done: punto 1, `make ci` en cada tarea; 2, tests offline en cada tarea; 3, R4 y R5 en
  `cmd/laboratorio` (convenciones); 4, `schemas/` regenerado en T002 con `make schema-check`; 5, defecto `inesperado` y
  errores de uso (T008 a T011); 6, e2e en T001 y CHANGELOG en T002 y T012; 7, no aplica en el run: el ADR que precisa
  el 0046 lo escribe quien lanza el sondeo antes de fusionar (FR-051, FR-052); 8, no aplica (ninguna fuente); 10, no
  aplica: ninguna skill del producto cambia (FR-004) y la de laboratorio cumple la forma por T005; 9, 11, 12 y 13, no
  aplican.
- Cada fila de «Controles de umbral» de plan.md y la tarea que construye su control: SC-001/FR-001/FR-002/FR-003
  (`TestAyudasDeLosArgumentos`), T003; SC-001/FR-001/FR-003 (`TestAyudaComoDescripcion`), T002; FR-004
  (`TestEsquemasPublicados`, de hoy), T002, que lo ve en rojo antes de regenerar; SC-002/FR-004
  (`TestSkillsDelRepositorio`, de hoy), T002, con `skills/` sin cambios; FR-029 (`TestInstrucciones`, de hoy), T008,
  sin tocar `internal/mcp/`; SC-003/FR-013/FR-014, T004; SC-004/FR-011/FR-020 a FR-028, T005; FR-012, T006;
  SC-005/FR-010, T007; SC-006/FR-041 a FR-043, T009; SC-007/FR-030/FR-031, T010 (y la cadena, T011, con `TestRecetasDelSondeo` para las recetas de `evals-sondeo` y `evals-laboratorio`); FR-021/FR-024/FR-029
  (`TestServidorDelLaboratorio`), T008; SC-008/FR-053, las cuatro filas `evals:`, las mide el job de evals del cierre,
  que hace el workflow tras la revisión final con los umbrales de hoy, sin tarea que los toque.
- Requisitos sin tarea propia: FR-050 y FR-051 (la persona, fuera del run), FR-052 (anotado en `gates/supuestos.md`),
  FR-053 y SC-008 (el cierre del workflow), SC-009 (capa 3).
