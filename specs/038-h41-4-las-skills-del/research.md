# Research: H41.4 · Las skills del producto, con las descripciones cortas

Todo lo verificado se ejecutó en esta sesión y se vio terminar, sobre una copia del repositorio en el temporal del
sandbox (`/tmp/claude/plan-h414/copia`, `git clone --local` de la rama, cabeza `b13886af`, que en `skills/`,
`laboratorio/` e `internal/` es la base `ab951c22`: `git diff --stat ab951c22 HEAD` solo da el paso del fichero del
hito a `docs/roadmap/done/`). En la copia se aplicaron las cuatro descripciones, el cambio de FR-005, un prototipo de
las dos comprobaciones nuevas y el guion; nada de eso está en el repositorio, y la copia no lo toca. Ninguna orden usó
red, modelo ni credencial.

## Verificado (V1-V18)

- **V1 · Las medidas.** Un programa de un solo uso en la copia (`skills.LeerFrontmatter` y `unicode/utf8`): los
  textos de `laboratorio/producto-descripciones-cortas/<skill>.txt` sin su salto final miden 417 (`consulta-juridica`),
  420 (`jurisprudencia`), 426 (`normativa`) y 427 (`plazos-y-territorio`) runas, **1690** entre los cuatro, y 431,
  424, 438 y 435 bytes (1728); ninguno tiene dos espacios seguidos, blancos en los bordes ni saltos dentro. Las
  `description` de hoy miden 705, 842, 887 y 879 runas (**3313**), en bloques `>-` de 8, 9, 9 y 9 líneas contando la
  de la clave, la más ancha de 118, 118, 119 y 118 runas.
- **V2 · La composición.** Con el corte de D4, cada `SKILL.md` compuesto da con `skills.LeerFrontmatter` el texto de
  la variante (4 de 4), y lo que no es la clave (`descripcionYCuerpo`, la regla de `lineasDelSkill`) es el del
  original, byte a byte (4 de 4). Cada bloque queda en la línea de la clave y cuatro sangradas (contracts/skills.md §2).
- **V3 · La base de la rama.** `git merge-base main HEAD` da `ab951c227e1fc8acac8bd6b23534e275e4a36a76`.
  `git ls-tree -r --name-only ab951c22 skills/consulta-juridica` da un solo fichero, `SKILL.md`. Su SHA-256 sin las
  líneas de la clave es `bd089909060d0fd82c1c6d2ee4ae148f6f0a1b2b7d7eefd476d164dc610a205a`, calculada dos veces: con la
  regla de `lineasDelSkill` en el programa de V1 y con `awk` y `sha256sum` sobre `git show ab951c22:…`.
- **V4 · Qué se pone en rojo con las descripciones nuevas.** Con solo los cuatro `SKILL.md` cambiados,
  `go test -count=1 -skip TestMedidasDeTiempo ./...` da 28 paquetes en verde y `internal/evals` en rojo solo por
  `TestSkillsDelProductoDesdeControl/{jurisprudencia,normativa,plazos-y-territorio}/{desde control,frontmatter}`;
  «movimientos», `TestConLaDescripcion`, `TestEleccionDeLosMutantes`, `TestDescripcionesCortas`,
  `TestVariantesDeControl`, `TestSkillsDelRepositorio` y los guiones de `TestEntregaDelHito` (H19 a H41.3) siguen en
  verde. Con `-tags=integration`, `internal/skills`, `internal/evals` (sin ese test) e `internal/app`, en verde.
- **V5 · Lo generado.** La receta de `make skills-sync` (`scripts/skills-sync.sh`: `go test -count=1 -run
  '^TestSkillsDelRepositorio$' ./internal/app/ -args -regenerar-skills`) sobre la copia cambiada no toca nada más:
  `git status --porcelain` sigue dando los cuatro `SKILL.md`, con el mismo diff (16 líneas añadidas, 31 quitadas).
- **V6 · El prototipo.** El cambio de contracts/controles.md §3 y las dos comprobaciones de §1 y §2, en la copia,
  salen en verde, también con `-race`; con ellas y el guion de V9, `go test -count=1 -skip TestMedidasDeTiempo ./...`
  da los 29 paquetes en verde y ninguno en rojo. `golangci-lint run ./internal/evals/...` (`go tool
  -modfile=tools/golangci-lint/go.mod golangci-lint`, la herramienta del `Makefile`) dio primero 12 avisos —10 de
  `thelper` en los mutantes escritos como cierres y 2 de `gosec` G304 por `os.ReadFile` con una ruta variable— y 0
  tras leer con `fs.ReadFile` sobre `os.DirFS`, escribir con `leerFichero` y `escribirFichero` (`internal/evals/preparar.go:524`
  y `:532`) y abrir cada mutante con `t.Helper()`: el contrato lo fija así (D10).
- **V7 · Los rojos de `TestSkillsDelProductoDesdeControl` cambiado.** Sobre la copia con el cambio, cada mutante da su
  rojo con la ruta: `## Protocolo` → `## Protocolos` en `normativa` («desde control», `skills/normativa/SKILL.md
  difiere de lo que dan su control y su declaración»); `name` de `jurisprudencia` y el orden de `kitlegal-applets` de
  `plazos-y-territorio` («desde control» y «frontmatter»); un carácter al final de
  `jurisprudencia/references/consulta_preparada.md` («desde control»); `normativa/references/otra.md` («… sobra: no
  está en lo que dan su control y su declaración»); y la `description` de `normativa` sin el punto final («desde
  control» y «frontmatter»: `la description de skills/normativa/SKILL.md es la de
  laboratorio/producto-descripciones-cortas/normativa.txt`). Con la `description` de `normativa` en un bloque `>-` de
  una sola línea, `TestSkillsDelProductoDesdeControl` sigue en verde —nada cambia fuera de la clave y el texto es el
  mismo— y la comprobación de FR-002 sale en rojo: la forma la controla FR-002, no FR-005.
- **V8 · Los mutantes de las comprobaciones nuevas.** Cada caso de contracts/controles.md §1.3 y §2.3 da, en el
  prototipo, exactamente sus líneas (el de la quinta skill, tras corregir la suma esperada a 1710: «Una skill de
  ensayo.» mide 20).
- **V9 · El guion.** `internal/app/testdata/script/h41-4-skills-instaladas.txtar` con el texto de contracts/aceptacion.md,
  en la copia: en verde con el prototipo; en rojo por una aserción (`h41-4-skills-instaladas.txtar:22: no match`) con
  `skills/` de la base; y en rojo en la línea 31 con la `description` de `normativa` en un bloque `>-` de una línea.
  Sus cuatro patrones son, carácter a carácter, los que da `regexp.QuoteMeta` sobre el texto de la variante con cada
  espacio cambiado por `( |\n  )` (`grep -o -F -f` del generador sobre el guion: 4 de 4).
- **V10 · `grep` de testscript.** `scriptMatch` compila `regexp.Compile("(?m)" + pattern)` sobre el contenido entero
  del fichero (`github.com/rogpeppe/go-internal@v1.16.0/testscript/cmd.go:690`, la versión de `go.mod`): `^` y `$`
  son de línea, `\A` es el principio del fichero y `\n` dentro de las comillas simples llega al patrón.
- **V11 · Cómo se activa el guion.** `scripts/workflow/aceptacion.sh:77-80` lo copia como
  `internal/app/testdata/script/<hito en minúsculas, con «.» → «-»>-<nombre>`: `h41-4-skills-instaladas.txtar`, que
  ejecuta `TestEntregaDelHito` (`internal/app/e2e_test.go:556`).
- **V12 · Los umbrales del cierre.** Los informes del cierre de H41.3 (`specs/037-h41-3-por-que-normativa/gates/evals/*.json`)
  publican con `decide: true` los nombres de «Controles de umbral»; `consulta-juridica` no publica ninguno con
  `decide: true` (solo `fuera_de_lo_grabado`, con `decide: false`). Con una descripción cambiada, el job juzga la de
  cada skill de la cabeza (`internal/evals/decision.go:84-117`) con la única clase del juez de las descripciones,
  `habla_en_primera_o_segunda_persona`, que decide (`internal/evals/descripciones.go:29` y `:129`), y la publica como
  umbral con `decide: true` (`specs/030-h32-arquitectura-de-skills/gates/evals/consulta-juridica.json` y
  `plazos-y-territorio.json`: `{"nombre":"habla_en_primera_o_segunda_persona","decide":true,"cumple":true}`).
- **V13 · Nada generado copia las descripciones (FR-007).** `git grep -l` en `ab951c22`, sin `specs/` ni `evidencias/`,
  de una frase de la descripción de hoy de cada skill («qué dice la jurisprudencia sobre la cláusula suelo», «dónde se
  regula una», «si un municipio es foral o qué norma prevalece», «también cuando el agente cree saber la respuesta»):
  solo `skills/`, las variantes de `laboratorio/`, `evals/jurisprudencia/03-por-materia.yaml` (la pregunta de la eval,
  que la descripción de hoy pone de ejemplo; no es una copia ni se genera), `docs/roadmap/done/H7.4-…`, el ADR 0055 y
  `CHANGELOG.md`. `web/src/pages/llms.txt.ts:4` lee las skills «de la versión publicada»; el plugin y el `.mcpb` los
  compone `internal/empaquetado` (`doc.go:3-5`) al empaquetar, y no hay `plugin/` en el repositorio.
- **V14 · Lo que instala el binario.** `go run ./cmd/kitlegal skills install --dir <temporal>/destino --json` desde la
  copia sale con `{"ok":true,…}` y las cuatro skills `instalada`; el `SKILL.md` de `normativa` abre con el bloque de
  contracts/skills.md §2.
- **V15 · El job no cambia.** `.github/workflows/evals.yml`: `MODELO_DE_EVALS: claude-sonnet-5-5` (línea 298) y
  `VERSION_DE_CLAUDE_CODE: 2.1.326` (línea 315).
- **V16 · Lo que corre `make ci`.** `make test` es `go test -race -shuffle=on … ./...` (`Makefile:96-97`), así que
  los tests nuevos de `internal/evals` entran sin tocar el `Makefile`; `make skills-check` (`Makefile:136-137`) no
  cambia. `make evals-laboratorio` toma once argumentos (`Makefile:172-173`) y `make laboratorio-tabla`, `DIRECTORIO`
  (`Makefile:189-190`), con la forma de `specs/037-h41-3-por-que-normativa/contracts/arnes.md` §1. La tabla del sondeo
  de H41.3, con 640 sesiones, mide 99 412 bytes y 653 líneas (`evidencias/sondeos-20261117/tabla-normativa.md`).
- **V17 · La tabla junta los sondeos de la misma clave.** `make laboratorio-tabla` lee la ruta como un sondeo si tiene
  `sondeo.json` o, si no, cada subdirectorio que lo tenga (`internal/evals/laboratorio_tabla.go:558-561`), y pone cada
  sesión en la fila de su clave (`anotar`, `:732-739`): variante, modo, modelo, versión de Claude Code, skills
  instaladas, instalación y presupuesto (`claveDeFila`, `:112-124`). Las dos filas del sondeo de H41.4 tienen esa clave
  igual (`normativa-producto`, herramienta, `claude-sonnet-5-5`, 2.1.326, las cuatro, `producto`, sin presupuesto): con
  el directorio de los dos, saldría una fila de 640. Por eso el quickstart pide una tabla por sondeo (escenario 8). Lo
  dice el código; no se ha ejecutado con dos sondeos así.
- **V18 · Las órdenes del quickstart.** En la copia: `go test -count=1 -run '^TestEntregaDelHito$/^h41-4-skills-instaladas$'
  -v ./internal/app/` da `--- PASS: TestEntregaDelHito/h41-4-skills-instaladas`; `git diff --name-only <base> --
  '*_test.go'` da los dos ficheros de test del hito; `git diff --quiet <base> -- laboratorio/ evals/ web/ mcp/
  internal/empaquetado/ .goreleaser.yaml .github/workflows/evals.yml && echo "sin cambios"` da `sin cambios`; y `make
  skills-check`, `ok` en sus tres paquetes. En el repositorio: el `awk` del escenario 5 saca la sección «Cambiado» de
  «[Unreleased]» de `CHANGELOG.md`, y el `jq` del escenario 7, sobre los informes del cierre de H41.3, da `skill`,
  `veredicto` y los umbrales con `decide`. `check-tools` solo escribe en la salida de error y la receta de
  `laboratorio-tabla` empieza por `@` (`Makefile:190` y `:294-308`), así que redirigir la salida deja solo la tabla.
  `git worktree add -h` da `git worktree add [-f] [--detach] … <path> [<commit-ish>]`. `TestArnesDeLasFamilias`
  (`internal/evals/laboratorio_sondeo_test.go`, sin etiqueta de compilación) entró en V4.

## Supuestos no verificados

- **S1 · El host lee el bloque `>-` como el escalar del sondeo de H41.3.** El sondeo de H41.3 instaló la variante con
  la `description` en un escalar de una línea (`skillConLaDescripcion`); el producto la lleva en un bloque `>-`. Que
  `skills.LeerFrontmatter` dé el mismo texto en los dos está verificado (V2); que Claude Code lo lea igual no se puede
  comprobar sin modelo. Lo mide el sondeo de aceptación (FR-011), que instala `skills/` tal cual.
- **S2 · El resultado del sondeo y del cierre.** Lo que den las evals del cierre (FR-010) y el sondeo de la persona
  (FR-011) no se puede saber en el run; el plan no lo afirma.

## Decisiones (D1-D11)

- **D1 · Dónde viven las comprobaciones nuevas.** **Decisión**: un fichero de test nuevo,
  `internal/evals/descripciones_del_producto_test.go`, en el paquete `evals`, con `TestDescripcionesDelProducto`
  (FR-001, FR-002, FR-004) y `TestConsultaJuridicaDesdeLaBase` (FR-003). Reutiliza lo que ya lee la variante y parte
  un `SKILL.md`: `descripcionesCortas`, `descripcionYCuerpo`, `leerLosFicheros`, `maximoDeUnaDescripcionCorta` y
  `maximoDeLasDescripcionesCortas`. **Alternativas**: (a) `internal/skills`, junto a `ValidarFrontmatter` —
  rechazada: `evals` importa `skills`, así que `skills` no puede leer la variante sin un ciclo, y los 450 y los 1700
  son límites de la variante, no del estándar que valida para toda skill y la vertical de ensayo; (b) un test más en
  `laboratorio_descripciones_test.go` o en `TestSkillsDelRepositorio` — rechazada: son suites congeladas, y FR-005 dice
  que la única que cambia es `TestSkillsDelProductoDesdeControl`; (c) una función de producción o una orden — rechazada:
  ningún consumidor fuera de `make ci` (principio V); (d) un objetivo de `make` — rechazado: `make test` ya la corre
  (V16). **Por qué**: lo más simple que lee las dos cosas que compara con el código que ya las lee.
- **D2 · `consulta-juridica`, por huellas de la base.** **Decisión**: una tabla `ruta → SHA-256` de cada fichero de
  `skills/consulta-juridica/` en `ab951c22`, la del `SKILL.md` calculada sin las líneas de la clave (V3), fijada en el
  test; la comprobación compara, en orden de ruta, y da el primer fichero que difiere, falta o sobra. **Alternativas**:
  (a) leer la base con git (`git show ab951c22:…`) — rechazada: el `make ci` del cierre no garantiza el historial (spec,
  «Assumptions»); (b) una copia de la skill en `testdata/` — rechazada: duplica el fichero en una ruta protegida (capa
  3) para dar lo mismo que una huella; (c) no comparar lo generado, como `sinLoGenerado` — rechazada: FR-003 y SC-002
  piden byte a byte todo salvo la clave, y la región de paquetes del `SKILL.md` es de la skill. **Por qué**: el
  precedente de `huellasDeLosControles` (`TestVariantesDeControl`), una línea por fichero.
- **D3 · El cambio de `TestSkillsDelProductoDesdeControl`.** **Decisión**: en «desde control» y en «frontmatter», lo
  esperado y lo obtenido se comparan con el `SKILL.md` sin las líneas de la clave (`descripcionYCuerpo`), y en los dos
  se exige aparte que la `description` que lee `skills.LeerFrontmatter` de `skills/<skill>/SKILL.md` sea el texto de la
  variante, que el test lee una vez con `descripcionesCortas`. **Alternativas**: (a) componer lo esperado con
  `skillConLaDescripcion` — rechazada: escribe un escalar de una línea, y habría que quitar igual las líneas de la clave
  para comparar; (b) injertar en lo esperado las líneas de la clave de `skills/` — rechazada: compara esas líneas
  consigo mismas y necesita igual la comprobación del texto; (c) el texto de las constantes de `TestDescripcionesCortas`
  — rechazada: FR-005 dice «el texto de la variante», y `TestDescripcionesCortas` ya fija que el fichero y la constante
  coinciden. **Por qué**: es lo que pide FR-005 («fuera de las líneas de la clave, byte a byte, y la description que
  lee `skills.LeerFrontmatter`, el texto de la variante»), sin tocar los otros ficheros, los movimientos ni los mensajes
  de rojo (V7).
- **D4 · Cómo se parte el bloque `>-`** (lo deja al plan el spec, «Assumptions»). **Decisión**: cada línea, con su
  sangría de dos espacios, de como mucho 120 caracteres, con tantas palabras como quepan y cortando solo en un espacio,
  que desaparece: la regla con que están partidos los de hoy (V1: la más ancha, 119). Da cuatro líneas sangradas por
  skill (contracts/skills.md §2). **Alternativas**: (a) una línea por frase — rechazada: unas pasan de 120 y otras
  quedan cortas, sin regla que se pueda repetir; (b) un escalar de una línea — rechazada por FR-002. **Por qué**: la
  forma de la base, que el revisor lee igual que hoy.
- **D5 · El guion no afirma dónde se corta cada línea.** **Decisión**: cada patrón es el texto de la variante con
  `regexp.QuoteMeta` y cada espacio cambiado por `( |\n  )`, anclado entre `\A---\nname: <skill>\ndescription: >-\n  `
  y `\nmetadata:\n` (o `\n---\n` en `consulta-juridica`), más otro `grep` que exige una segunda línea sangrada.
  **Alternativa**: las cuatro líneas exactas de D4 — rechazada: el spec dice que basta un bloque `>-` de más de una
  línea sangrada que se lea como el texto exacto, y los guiones afirman solo lo que fija el spec
  (`.specify/workflows/hito/workflow.yml:453`, la tarea `[aceptacion]`).
  **Por qué**: el guion queda en verde con cualquier corte que cumpla el spec y en rojo con uno que no (V9).
- **D6 · Qué rojos se prueban dentro de los tests.** **Decisión**: los mutantes que piden los escenarios del spec
  —US1.2 (un carácter añadido, quitado o cambiado), FR-002 (un escalar de una línea, un bloque de una línea), FR-004
  (una de 451; una suma de 1701), el caso límite de la quinta skill y US2.3 en `consulta-juridica` (cuerpo, `name`, un
  fichero de más, uno de menos, y solo la clave, que no es defecto)— van como casos de tabla sobre copias de `skills/` y
  de la variante en `t.TempDir()`. Los rojos de `TestSkillsDelProductoDesdeControl` no se añaden: su comportamiento en
  rojo no cambia (FR-005) y se verificó una vez (V7), como en H41.1. **Por qué**: cada caso se traza a un escenario o
  requisito; ninguno es de un estado por debajo del umbral (un `SKILL.md` sin frontmatter o una variante ilegible son el
  error de la función, defecto `inesperado`, sin caso).
- **D7 · Las medidas son de `skills/`, todas sus skills.** **Decisión**: la longitud y la forma se miden en cada
  subdirectorio de `skills/`, tenga o no texto en la variante, y la suma es la de todas; la igualdad, solo con la que lo
  tiene. **Alternativa**: sumar solo las que tienen texto, como `descripcionesCortas` — rechazada: FR-004 pide medir
  «sobre `skills/`», y con una quinta skill el listado lleva cinco textos. Con una quinta sin texto la comprobación ya
  sale en rojo; que la suma la cuente da una línea más (contracts/controles.md §1.3).
- **D8 · El orden de las tareas.** **Decisión**: las cuatro descripciones y el cambio de FR-005 en una sola tarea, porque
  con una sin la otra `make ci` queda en rojo (V4); las comprobaciones nuevas, después (con `skills/` de hoy,
  `TestDescripcionesDelProducto` sale en rojo); `CHANGELOG.md`, al final. El guion, primero, como `[aceptacion]`.
- **D9 · La entrada de `CHANGELOG.md`.** **Decisión**: una viñeta en «[Unreleased]» → «Cambiado», con el texto de
  contracts/skills.md §4: las cuatro skills, la variante de H41.3, las cifras de V1, que nada más cambia y el porqué
  (ADR 0055). **Alternativa**: «Añadido» — rechazada: FR-008 dice «Cambiado».
- **D10 · Skills de Go aplicadas.** `golang-how-to` con `golang-testing` y `golang-stretchr-testify` (tablas de casos
  con `t.Parallel()`, `require` para lo que impide seguir y `assert` para el resto; mutantes sobre `t.TempDir()`),
  `golang-project-layout` (ningún paquete nuevo; el fichero de test junto al que reutiliza), `golang-naming`
  (identificadores en español como el paquete, sin exportar), `golang-lint` (ningún `//nolint`; V6) y `golang-safety`
  (lecturas por `fs.FS`, sin rutas fuera de la raíz). `golang-cli`: ninguna orden, bandera ni código de salida cambia.
- **D11 · Fuentes externas: ninguna.** El hito no graba nada ni lee ninguna fuente: ni manifiesto `grabaciones.json`,
  ni test `TestGrabar*`, ni nada en `evidencias/` ni en `data/`. La tabla del sondeo la deja la persona en
  `evidencias/`, fuera del run.
