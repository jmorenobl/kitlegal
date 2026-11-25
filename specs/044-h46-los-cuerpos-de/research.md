# Research: H46 · Los cuerpos de las skills de capacidad, con sitio para las materias

Fase 0 del plan. No quedó ninguna duda abierta en el contexto técnico: el spec y el hito fijan qué cambia, y el prototipo
`f89093dd` (rama `cuerpos-prototipo`) da los textos byte a byte. Lo que sigue es lo medido en esta sesión, las decisiones
con su alternativa descartada y dónde se comprobó cada afirmación sobre una herramienta.

## P1 · El prototipo, medido en esta sesión

En un clon del repositorio en el directorio temporal (`git clone -q /home/jorge/code/kitlegal /tmp/claude/h46-plan`,
`git checkout -q f89093dd`), nunca en el árbol de trabajo. La base del prototipo es `7aaad57f` (H45.1 fusionado, `git
merge-base main f89093dd`); desde ella, `main` solo cambia `docs/`, `.specify/workflows/hito/workflow.yml` y `web/`
(`git log --stat` de `223278b0`, `78bb6f76` y `6693f744`), así que el código y `skills/` de la base son los de `main`.

**Los cuerpos**, en runas, como el gate (`git show <rev>:skills/<skill>/SKILL.md | awk 'n>=2{print} /^---$/{n++}' |
wc -m`, con `en_US.UTF-8`; V15):

| | `normativa` | `plazos-y-territorio` | `jurisprudencia` | `consulta-juridica` |
|---|---|---|---|---|
| `main` | 10 750 | 10 686 | 10 842 | 4481 |
| `f89093dd` | 9498 | 9577 | 9473 | 4481 |
| `f89093dd`, sin las líneas entre las marcas de los paquetes | 9498 | 9483 | 9473 | (sin marcas) |
| Margen hasta 9600 | 102 | 117 | 127 | — |

Las cifras de `main` y del prototipo son las del hito y del ADR 0063. La única línea de paquete hoy es la de lo
administrativo en `plazos-y-territorio`, 94 caracteres con su salto (52 + 3 × 14).

**Las tablas de comandos**, la región entre sus marcas con ellas (`awk` de la marca de inicio a la de fin, `wc -m`):
`normativa` 2673 en `main` y 1436 en el prototipo (−1237); `plazos-y-territorio` 1498 → 929 (−569); `jurisprudencia`
1302 → 796 (−506). **Las referencias que ganan texto** (`wc -m`): `municipios.md` 3873 → 4666 (+793),
`consulta_preparada.md` 1997 → 2698 (+701), `documento_aportado.md` 2756 → 3208 (+452).

**Los tests sobre el prototipo**, en el clon:

- la lista de `make skills-check` (`go test -count=1 -run '^(TestSkillsDelRepositorio|…|TestIdentificadoresDeLasNormas)$'
  ./internal/app/ ./internal/skills/ ./internal/evals/`): `371 passed in 3 packages`;
- `go test -count=1 ./internal/skills/ ./internal/evals/`: `3945 passed in 2 packages`;
- `TestEntregaDelHito` con `h41-skills-instaladas`, `h41-1-skills-instaladas`, `h45-skills-instaladas` y
  `h45-paquete-de-mas`: los cuatro `PASS` (el último, 8,21 s);
- `go tool -modfile=tools/golangci-lint/go.mod golangci-lint run ./internal/skills/ ./internal/app/ ./internal/evals/`:
  `0 issues.`

**La suite de aceptación**, los dos guiones de contracts/aceptacion.md tal cual, copiados al clon con el prefijo `zz-`:
los dos `PASS` sobre `f89093dd`; con los `skills/` de `main` (`git checkout origin/main -- skills/`), los dos `FAIL` por
una aserción, `tabla-de-tres-columnas.txtar:20` y `textos-en-su-referencia.txtar:25` (`no match for …`). Antes, sus
mismas aserciones en un solo guion, con cada una invertida salvo la del bloque de la cobertura —que vale en los dos,
porque es lo que se queda—, `PASS` sobre los `skills/` de `main`: cada una de las otras 31 falla en `main`.

**Los mutantes del defecto nuevo**, sobre el prototipo, con `go test -count=1 -run
'^TestSkillsDelRepositorio$/^presupuestos$' ./internal/app/`:

| Mutante en `internal/skills/sincronia.go` | Casos en rojo |
|---|---|
| `maximoSinLosPaquetes = 9601` | 6: `nueve-mil-seiscientos-uno-sin-los-paquetes` y `once-mil-y-un-caracteres` de las tres skills con marcas |
| `maximoSinLosPaquetes = 9599` | 9: los 6 anteriores y `nueve-mil-seiscientos-sin-los-paquetes` de las tres |
| la regla desactivada (`… > maximoSinLosPaquetes && false`) | 6, los del primero |
| sin restar las líneas de los paquetes (`- 0*utf8.RuneCount(…)`) | 3, los de `plazos-y-territorio`, la única con una línea |
| la regla también sin marcas (`err == nil && (conRegion \|\| !conRegion)`) | 2: `once-mil-caracteres` y `once-mil-y-un-caracteres` de `consulta-juridica` |

Y tres más, cada uno restaurado después con `git checkout -q --`: `git show origin/main:internal/skills/comandos.go >
internal/skills/comandos.go` (la tabla de cuatro columnas) da `--- FAIL` en `TestRenderizarTabla`,
`TestRegenerarYComparar` y `TestCuerpoDeLaSkill`; `git show origin/main:skills/normativa/SKILL.md >
skills/normativa/SKILL.md` (su única entrada de FR-014 deshecha), `FAIL` en `TestSkillsDelProductoDesdeControl`; y `git
show origin/main:internal/skills/remisiones_test.go > internal/skills/remisiones_test.go` (las dos frases de `main`),
`FAIL` en `TestRemisionesDelRepositorio`. La orden de Q7 con dos alternativas de subtest selecciona
`TestOrdenesDeLasSkillsEmpotradas/control` y `TestSkillsDelRepositorio/presupuestos`, los dos `PASS` (V17).

**No medido en esta sesión**: `make ci` entero (`vuln`, `goreleaser-check`, `secrets`, `test-tiempos`, la cobertura) y las
evals, que solo corre el cierre. Lo corre cada tarea del run (`scripts/workflow/verificar.sh`) y el cierre del workflow.

## Decisiones

### D1 · La tabla, sin lo que da la llamada (FR-001 a FR-004)

- **Decisión**: en `internal/skills/comandos.go`, `columnasDeLaTabla` pasa a `| Orden | Herramienta | Qué hace |` con
  `|---|---|---|`; `filaDelVerbo` escribe tres celdas, con el mismo escape; y `RenderizarTabla` cierra con una línea en
  blanco, sin la del sobre ni la de las banderas. `RenderizarTabla` sigue llamando a `mismoSobreYBanderas` por cada
  verbo, con los mismos errores (FR-002), y `LeerDescripcionDeVerbo` y los campos `Devuelve`, `Sobre` y `Banderas` de
  `DescripcionDeVerbo` no cambian. Lo que queda sin uso se retira: el método `Devuelve.texto`, `lineaDelSobre`,
  `lineaDeLasBanderas`, `comoCodigo` y las constantes `textoDeObjeto`, `textoDeListaDeObjetos`, `textoSinFormaDeclarada`
  y `valorDeUnaBandera`. Se quedan `enumerar` (`paquetes.go:121`, `comandos.go:885`) y `Bandera.escrita`, que usa
  `sintaxisDeLaOrden`. El comentario de `RenderizarTabla` dice que lo que devuelve en `data`, el sobre y las banderas
  comunes los da el esquema de salida de la herramienta y `--describe` de la orden (ADR 0063).
- **Alternativas**: (a) dejar de leer `Devuelve` del documento: rechazada, una descripción con la salida mal formada
  dejaría de dar su defecto (FR-002); (b) dejar la línea del sobre o la de las banderas: rechazada, FR-001; (c) conservar
  las funciones sin uso: rechazada, `unused` las da (V3) y sería código muerto (constitución V).
- **Por qué**: es el cambio mínimo que deja los bytes de FR-001 y los defectos de FR-002; el prototipo lo hizo así y su
  lint da 0 issues (P1).

### D2 · El sitio de las materias, en `defectosDelCuerpo` (FR-005, FR-006)

- **Decisión**: una constante `maximoSinLosPaquetes = 9600` junto a `maximoDeLineasDelCuerpo` y
  `maximoDeCaracteresDelCuerpo` (`sincronia.go:27-31`), y en `defectosDelCuerpo`, detrás del defecto de 11 000: con
  `regionDeLosPaquetes.buscar(cuerpo)` sin error y con la región, `sinLosPaquetes := utf8.RuneCount(cuerpo) -
  utf8.RuneCount(cuerpo[region.desde:region.hasta])`; si pasa de 9600, el defecto `SKILL.md: cuerpo sin las líneas de
  los paquetes de <n> caracteres (máximo 9600)`. Sin las marcas, o con unas marcas que no van en pareja, nada: el
  defecto de las marcas ya lo da `paquetes.regenerar` (V7). `defectosDelCuerpo` recibe el contenido regenerado (V8), así
  que mide lo que deja `make skills-sync`. Sus comentarios, y el de `Regenerar` (`sincronia.go:186`), dicen la regla.
- **Alternativas**: (a) contarlo en `paquetes.regenerar`, que ya busca la región: rechazada, mezcla la sincronía de la
  región con los presupuestos y cambia dónde va el defecto (los del cuerpo van antes que los de las regiones,
  `sincronia.go:364`); (b) restar también las dos marcas: rechazada, el ADR 0063 (punto 1) y el supuesto del spec las
  cuentan; (c) aplicarla también a la región de la base de `consulta-juridica`, o a toda skill: rechazada, FR-006 y
  «Fuera de alcance».
- **Por qué**: FR-005 la pone en «la comprobación del cuerpo», detrás del defecto de 11 000; los mutantes de P1 la
  ponen en rojo por cada lado.

### D3 · `casosDePresupuestos`, con y sin las marcas (FR-008)

- **Decisión**: la función lee el `SKILL.md` de la skill en el árbol (`rutaEnLaSkill(raizDelRepositorio, …)`), busca la
  línea de la marca de inicio y la de fin, y cuenta en runas `p`, lo que hay entre ellas. Con las marcas, sus casos de
  caracteres son: el cuerpo ensanchado a `9600 + p`, sin defectos; a `9601 + p`, con solo el de FR-005 (9601); y a
  11 001, con el de 11 001 y, detrás, el de FR-005 con `11001 − p`. Sin ellas, los de hoy: 11 000 sin defectos y 11 001
  con el suyo. Los de 499 y 500 líneas no cambian. `ensancharElCuerpo` sigue añadiendo su línea de eñes al final del
  fichero, fuera de las marcas. Las dos marcas van como constantes del test en `internal/app`, como ya va
  `inicioDeLaTablaDeComandos` (`skills_test.go:53-55`).
- **Alternativas**: (a) leer la región en la copia, dentro de `alterar`: rechazada, el defecto esperado necesita `p` al
  construir el caso; (b) un campo nuevo en `skillDelRecorrido`: rechazada, fontanería para un solo subtest; (c) exportar
  las marcas desde `internal/skills`: rechazada, API nueva sin otro consumidor.
- **Por qué**: casa con el guion `h45-paquete-de-mas`, cuyo paso de sincronía corre el subtest en su copia (V9): ahí
  `raizDelRepositorio` es la copia, y `normativa` lleva la línea del paquete de prueba (70 caracteres), que el caso
  cuenta (P1: el guion en verde).

### D4 · `TestVerticalDeEnsayo` sin «otra línea de su misma longitud» (FR-009)

- **Decisión**: se retiran `comprobarElSitioDeOtroPaquete` y su llamada de `internal/skills/vertical_de_ensayo_test.go`,
  el import `unicode/utf8` que solo usaba ella, y la frase del comentario del test que la describe, que pasa a decir que
  el sitio de los demás paquetes lo comprueba el presupuesto del cuerpo sin sus líneas. Lo demás del test no cambia.
- **Alternativa**: dejarla: rechazada, añadía su línea fuera de las marcas y bajaba el tope real a unos 9530 (hito).

### D5 · Los textos de `skills/`, los de `f89093dd` tal cual (FR-010 a FR-013)

- **Decisión**: la tarea `[datos]` trae `skills/` con `git checkout f89093dd -- skills/`, un objeto del repositorio
  local (`git rev-parse --verify f89093dd^{commit}` en esta sesión), sin red. Va detrás de la de la tabla: con la tabla de
  tres columnas ya generada, las tablas del prototipo son las que da `make skills-sync` (P1: `skills-check` en verde
  sobre `f89093dd`), y no hay deriva.
- **Alternativa**: reescribir los textos a mano siguiendo el diff: rechazada, FR-010 pide los bytes del prototipo, y a
  mano es donde se cuela un byte distinto.

### D6 · Las declaraciones del producto: cambios, no movimientos (FR-014)

- **Decisión**: `declaracionesDelProducto` (`internal/evals/laboratorio_familias_test.go`) gana 15 entradas con los bytes
  del prototipo: en `normativa`, 1 `cambioDelCuerpo`; en `jurisprudencia`, 5 del cuerpo y 4 `cambioEnUnaReferencia` (2
  en `consulta_preparada.md`, 2 en `documento_aportado.md`); en `plazos-y-territorio`, 3 del cuerpo y 2 en
  `municipios.md` (contracts/textos-y-controles.md §2). Lo que el test comprueba no cambia.
- **Alternativa**: `movimientoALaReferencia`: rechazada, es para un texto que llega igual al final de la referencia,
  con los blancos juntados detrás de una línea en blanco (`laboratorio_familias_test.go:318-325`); aquí el texto movido
  cambia palabras («con esta forma» pasa a «con la forma del paso 4 de `SKILL.md`») y llega bajo un apartado nuevo.
- **Por qué**: es como lo declaró el prototipo, y su test pasa (P1).

### D7 · Las remisiones del plan y los guiones de H41 (FR-015, FR-016)

- **Decisión**: en `remisionesDelPlan`, las dos frases nuevas del cuerpo, completas (contracts/textos-y-controles.md
  §3); en `h41-skills-instaladas.txtar`, el `grep` de la orden con la tabla detrás pasa a dos, la orden con su frase nueva
  en el `SKILL.md` y la cabecera de la tabla en `references/consulta_preparada.md`; en `h41-1-skills-instaladas.txtar` y
  en su copia congelada, el `grep` del «Protocolo» con la orden nueva de `municipios.md`, partida en sus líneas de 120
  columnas. Los dos ficheros de H41.1 quedan iguales byte a byte.
- **Alternativa**: retirar las aserciones que fijan el texto de hoy: rechazada, el siguiente cambio de texto quedaría sin
  control (US4).

### D8 · La suite de aceptación: dos guiones, acotados a la región o al apartado (FR-021)

- **Decisión**: `specs/044-h46-los-cuerpos-de/aceptacion/tabla-de-tres-columnas.txtar` (US1) y
  `…/textos-en-su-referencia.txtar` (US3), con los patrones de contracts/aceptacion.md, medidos en P1. Lo que el spec dice
  de una región o de un apartado se busca ahí: la tabla, entre sus marcas (`(?s)<!-- inicio de la tabla de comandos: …
  .*? … .*?<!-- fin de la tabla de comandos -->`); un apartado, desde su `## ` sin otro `#` por medio (`[^#]*`); «termina
  con», con `\z`. Cada frase citada, con `\s+` entre sus palabras (spec, US3).
- **Alternativas**: (a) negativas sobre el fichero entero: rechazada, el spec dice «la región»; (b) contar las tablas o
  las filas: rechazada, fijaría los applets de hoy, lo que H45.1 quitó de los tests (ADR 0062); (c) un solo guion:
  rechazada, uno por historia dice mejor qué cubre cada uno, y cuestan 0,02 s.
- **Por qué US1 y US3 y no las demás**: US2 es un defecto que no da ninguna orden de `kitlegal` (FR-021), US4 son tests y
  guiones que ya existen, y US5, las evals del cierre.

### D9 · El orden de las tareas, cada una en verde (FR-020)

- **Decisión**: T001 la suite (`[aceptacion]`); T002 la tabla (FR-001 a FR-004, y lo de §1 de `docs/SKILLS.md` que la
  describe); T003 `[datos]`, los textos con lo que los fija (FR-010 a FR-016); T004 el sitio de las materias (FR-005 a
  FR-009, y §3 —con las cifras de los tres cuerpos, que ya no cambian tras T003—, §5 y «La vertical de ensayo» de
  `docs/SKILLS.md`); T005 `CHANGELOG.md` (FR-018), que describe el hito entero; T006 el registro de mutantes, que vuelve
  a medir los cuerpos (quickstart Q3).
- **Por qué en verde cada una**: tras T002, los cuerpos son los de `main` menos la tabla, 9513, 10 117 y 10 336, bajo
  11 000, y la regla de 9600 aún no existe; tras T003, los de P1; T004 entra con los cuerpos ya bajo 9600. Con la regla
  antes que los textos, `plazos-y-territorio` y `jurisprudencia` quedarían por encima (hito). La tarea `[datos]` lleva
  Go, `skills/` y guiones a la vez porque cada parte sin las otras deja `make ci` en rojo (hito; precedente: T002 de
  H41.1).

### D10 · El registro de mutantes, en un clon del temporal (SC-003 a SC-006)

- **Decisión**: T006 escribe `specs/044-h46-los-cuerpos-de/mutantes.md` con la orden, lo esperado y lo visto de cada
  mutante de la tabla «Mutantes» del plan, en un clon en `$TMPDIR` (`git clone -q --no-hardlinks`), nunca en el árbol
  de trabajo, como D7 de H45.1. En el clon, `main` es `origin/main` (V14).
- **Por qué**: SC-003 a SC-005 piden ese registro como control junto al test, y un mutante en el árbol lo rechaza el
  guardián.

### D11 · El fixture del subtest «control» (FR-003)

- **Decisión**: en `TestOrdenesDeLasSkillsEmpotradas/control` (`internal/app/skills_test.go:394-429`), la cabecera pasa a
  la de tres columnas y cada fila pierde su cuarta celda (`| objeto |`, `| nada |`), para que el fixture copie una tabla
  como la de hoy; lo que espera no cambia. Ese subtest lee la orden y la herramienta de cada fila (`filasDeLaTabla`,
  `:2408-2428`) y no mira la cabecera, así que no sale en rojo con la de cuatro: el rojo de la tabla de cuatro columnas
  es el de los tres tests que fijan sus bytes (FR-003), que el registro de mutantes deja.
- **Alternativa**: una aserción nueva sobre la cabecera del fixture: rechazada, no la pide el spec y no la produce el
  binario.

### D12 · La guía y el registro (FR-017, FR-018)

- **Decisión**: `docs/SKILLS.md` §1 (línea 78), la tabla de comandos con sus tres columnas; «La vertical de ensayo»
  (líneas 370-371), sin «sitio para otra línea de paquete de su misma longitud»; §3, una fila «Caracteres del cuerpo sin
  las líneas de los paquetes, en una skill con sus marcas | Como mucho 9600» con su porqué (1400 para las líneas, unas
  veinte materias de nombre corto a 52 + 3n), y las cifras de los tres cuerpos al cerrar, medidas como el gate; §5, la
  fila del defecto con su mensaje y los 9600 en «En los límites, en verde» (línea 620). `CHANGELOG.md`, bajo
  *Unreleased* → *Cambiado*, una entrada del hito (H46; ADR 0063) que nombra `normativa`, `plazos-y-territorio` y
  `jurisprudencia` y lo que cambia en cada una, y la regla nueva de `skills-check`, como la de H41.1
  (`CHANGELOG.md:246`).

### D13 · Las evals: ninguna cambia, y el cierre las mide todas (US5, SC-008)

- **Decisión**: ninguna eval, umbral ni el job cambian («Fuera de alcance»). Que el cierre no reutilice nada de `main`
  lo da su regla de hoy: con la etiqueta `evals` (`cierre.sh:170`) el job reutiliza solo la medida de `main` con la
  misma huella (`evals.yml:20-26`), y el conjunto de cada eval lleva su skill y cada skill que nombran su `SKILL.md` o
  sus referencias a mano (`dependencias.go:917-921`); `consulta-juridica` nombra las tres de capacidad, que cambian, así
  que ninguna eval de ninguna skill conserva su huella (V10).
- **Alternativa**: una tarea que pida la suite completa: rechazada, el cierre en la plataforma lo hace el workflow, no
  una tarea, y no hace falta.

### D14 · Skills de Go

`golang-how-to` orquesta: `golang-testing` y `golang-stretchr-testify` (los casos de `casosDePresupuestos` como tabla,
premisas con `require`, comparaciones con `assert.Equal` de los defectos en su orden), `golang-lint` (la configuración de
hoy, sin `//nolint` ni `t.Skip`; `unused` obliga a retirar lo que la tabla deja sin uso, D1), `golang-code-style`
(mensajes en español, líneas de ≤ 120) y `golang-naming` (la constante como sus vecinas, `maximoSinLosPaquetes`). Ningún
paquete, tipo exportado ni dependencia nuevos: `golang-project-layout` no aplica.

## Verificaciones en local

| # | Afirmación | Dónde se comprobó |
|---|---|---|
| V1 | El toolchain es Go 1.27 | `go version` → `go1.27.4 linux/amd64`; `go.mod:3` (`go 1.27.0`) y `:5` (`toolchain go1.27.4`) |
| V2 | El `grep` de testscript v1.16.0 compila `(?m)` + el patrón con `regexp` de Go y lo busca en el fichero entero: valen `\A`, `\z`, `(?s)` y `\s+` a través de saltos | `/home/jorge/go/pkg/mod/github.com/rogpeppe/go-internal@v1.16.0/testscript/cmd.go:689-698` (`regexp.Compile(`(?m)` + pattern)`, `os.ReadFile`) |
| V3 | `unused` («Checks Go code for unused constants, variables, functions and types») va activado por defecto en golangci-lint 2.13.2, y el repositorio parte de `default: standard` | `go tool -modfile=tools/golangci-lint/go.mod golangci-lint help linters` y `… version`; `.golangci.yml:47-48` |
| V4 | `aceptacion.sh activar` copia cada guion a `internal/app/testdata/script/<hito en minúsculas, con - por .>-<nombre>`, y `rojo-primero` exige que cada uno falle por una aserción en `TestEntregaDelHito` | `scripts/workflow/aceptacion.sh:36-62` y `:74-89` |
| V5 | `testdata/` solo con `[datos]`; cada tarea pasa el guardián y `make ci` | `scripts/workflow/guardian-diff.sh:7-11`, `:75` y `:95`; `scripts/workflow/verificar.sh:2` y `:47` |
| V6 | `make skills-check` solo compara y `make ci` lo corre; `make skills-sync` corre `TestSkillsDelRepositorio` con `-regenerar-skills` | `Makefile:136-137`, `:193-194` y `:314`; `scripts/skills-sync.sh:7` |
| V7 | `regionDeMarcas.buscar` da lo que hay entre las dos líneas de las marcas, sin ellas; sin ninguna, ni región ni error; unas que no van en pareja, error, cuyo defecto ya da `paquetes.regenerar` | `internal/skills/paquetes.go:62-91` y `:258-263` |
| V8 | `defectosDelCuerpo` mide el contenido regenerado | `internal/skills/sincronia.go:364` |
| V9 | El paso de sincronía de `h45-paquete-de-mas` corre `TestSkillsDelRepositorio` con `-regenerar-skills` en la copia, cuya raíz es `raizDelRepositorio` (`../..`) | `internal/app/testdata/script/h45-paquete-de-mas.txtar:175`; `internal/app/skills_test.go:44` y `:91-97` |
| V10 | El conjunto de una eval lleva su skill y, por cierre, cada skill que nombra un `SKILL.md` o una referencia a mano del conjunto; el cierre pone la etiqueta `evals`, que mide lo elegido y reutiliza solo lo que tiene la misma huella; `consulta-juridica` nombra las tres skills de capacidad | `internal/evals/dependencias.go:917-921`; `scripts/workflow/cierre.sh:170`; `.github/workflows/evals.yml:20-26`; `grep -o` en `skills/consulta-juridica/SKILL.md` (`normativa` 1, `jurisprudencia` 1, `plazos-y-territorio` 4) |
| V11 | El veredicto del job falla por una serie que no pasa y por un umbral que decide y no se cumple | `scripts/evals.sh:183-186`; `internal/evals/doc.go:156-161` |
| V12 | Los umbrales con `decide: true` de cada skill son los de los informes del cierre de H45.1, y su configuración no ha cambiado desde entonces | `jq` sobre `specs/043-h45-1-los-tests-del/gates/evals/*.json` (`cierre.json`: `52af770e`); `git diff --stat 52af770e HEAD -- .github/workflows/evals.yml evals/ internal/evals/ ':!internal/evals/*_test.go'`, vacío |
| V13 | `make test` deja `coverage.out`, que Git ignora | `Makefile:97`; `.gitignore:5-6` (`/*.out`, `/coverage.*`) |
| V14 | En un clon, `main` no es una rama local y `origin/main` sí | `git -C /tmp/claude/h46-plan checkout -q main -- skills/` → `fatal: invalid reference: main`; con `origin/main`, hecho |
| V15 | `wc -m` cuenta caracteres, no bytes, en la locale de la sesión | `locale` → `LC_CTYPE="en_US.UTF-8"`; `printf 'ñ\n' \| wc -m` → `2` |
| V16 | `golangci-lint` se invoca con `go tool -modfile=tools/golangci-lint/go.mod` | `Makefile:41` y `:112-113` |
| V17 | `-run` parte la expresión por cada `/` y cada `\|` que no van entre corchetes ni paréntesis: `^A$/^(b\|c)$` son dos niveles con una alternancia en el segundo, y `^A$/^b$\|^C$/^d$` dos alternativas de dos niveles | `/home/jorge/.local/go/src/testing/match.go:168-209` (`splitRegexp`; `go env GOROOT`) |
| V18 | El servidor MCP anuncia cada herramienta con un esquema de salida | `internal/mcp/servir.go:104-110` (`OutputSchema: herramienta.Salida`) |
| V19 | Los pathspec `'*_test.go'` y `'*.txtar'` de `git diff --name-only` casan en cualquier carpeta, y `':!*_test.go'` los excluye (quickstart Q8) | ejecutado sobre `7aaad57f~1..7aaad57f` (H45.1): los cinco ficheros de test y guiones, en `internal/…` y `specs/…`; con `'*.go' ':!*_test.go'`, ninguno |

## Datos externos

Ninguno. El hito no graba nada ni deriva nada a `data/`: no hay manifiesto `grabaciones.json` ni test `TestGrabar*`, y
el paso `grabar_datos` no tiene qué hacer. Los textos salen de `f89093dd`, un objeto del repositorio (D5).

## Supuestos no verificados

- Que `make ci` entero quede en verde: en P1 se corrieron los tests de los tres paquetes que el hito toca, los cuatro
  guiones y el lint; `vuln`, `goreleaser-check`, `secrets` y `test-tiempos` los corre cada tarea del run.
- Que las respuestas no cambien (US5, SC-008): lo decide el cierre de evals, que el run no corre en sus tareas.
- Que el prototipo de H42, rebasado sobre `main` con este hito, pase sin ningún `.go`: sus datos no están en el
  repositorio; lo comprueba la persona antes de lanzar H42 (spec, «Fuera de alcance»).
