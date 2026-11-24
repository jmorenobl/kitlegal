# Research: H45.1 · Los tests del molde, sin el estado de hoy del repositorio

Decisiones del plan (D), comprobaciones de herramientas en local (V) y el prototipo que las midió (P). Todo lo que se cita
como visto se ejecutó y se vio terminar en la sesión del plan, el 2026-11-24, o está en el repositorio; lo que no se pudo
medir lo dice su entrada.

## P1 · El prototipo

Un clon del repositorio en el temporal del sandbox (`git clone --no-hardlinks -q`, cabeza `b2daefa4`, con `origin/main`
en `e1555a7d`, el `main` del repositorio), con los cuatro tests como dicen D1 a D4 y el guion como dice D5, fuera del
árbol de trabajo. Lo que se vio:

- **Con los datos de `main`**: `TestCoberturaDelRepositorio`, `TestControlesSobreLasReferenciasAMano/forma-en-una-referencia-con-otros-bloques`
  y `/forma-en-ninguna-referencia`, `TestDefinicionDelJob/cuatro-evals-mas` y `TestEvalsDelRepositorio/precepto-de-la-cobertura`,
  en `PASS`; con el guion de D5 aún sin su línea `cmp`, `go test -count=1 ./internal/skills/ ./internal/evals/
  ./internal/app/` → `ok` en los tres (0,663 s, 22,282 s y 14,473 s); `go vet ./internal/evals/ ./internal/skills/` sin
  salida; `golangci-lint run ./internal/skills/... ./internal/evals/...` (el de `tools/golangci-lint/go.mod`) →
  `0 issues.`
- **La copia con un paquete más** (`-v -run '^TestEntregaDelHito$/^h45-paquete-de-mas$' ./internal/app/`): `PASS` en
  7,87 s, con `copia: hecha`, `sincronía: 0`, `orden del paso 2 de normativa, líneas más abajo: 1`, `tests enteros: 0` y
  `subtests: 0` (109 bytes); con la línea `cmp` de D5 y la copia congelada sin copiar, `FAIL` con «… differ»; copiada,
  `PASS` en 8,03 s.
- **Cada test devuelto a `main`, en la copia** (SC-005): `TestCoberturaDelRepositorio` → `tests enteros: 1`, `--- FAIL:
  TestCoberturaDelRepositorio`; los dos casos con «línea 36» → `tests enteros: 1`, `--- FAIL:
  TestControlesSobreLasReferenciasAMano/forma-en-una-referencia-con-otros-bloques`; `cuatro-evals-mas` de `main` →
  `subtests: 1`, «"5h52m0s" is not greater than or equal to "6h44m21s"» (21 evals); `llevaNoCubierto` de `main` →
  `subtests: 1`, «95-fuera-de-lo-declarado.yaml: no_cubierto.precepto , bloque  no es el precepto de ninguna fila de
  data/cobertura.yaml», el rojo del prototipo de H42.
- **Los mutantes de hoy, en el árbol del clon**: L06 → `TestCoberturaDelRepositorio` en rojo, «solo las 5 primeras»
  (mensaje de 751 bytes); L07 → en rojo, `expected: 9`, `actual: 0`; T1 → `cuatro-evals-mas` en rojo, «"5h52m0s" is not
  greater than or equal to "7h47m49s"» (también `/tope-con-las-evals-del-repositorio` y `/sinteticas`); L29 →
  `cuatro-evals-mas` en rojo; P6 → los dos casos en rojo; un control que da otra línea (`linea: numero + 1` en
  `tramosDeCodigo`) → los dos casos en rojo (y tres casos más, con defectos en `a_mano.md`); L33 → «12-cobertura-seguridad-social-sin-leer-la-da-primera.yaml:
  no_cubierto.precepto BOE-A-2015-10565, bloque a999 no es el precepto de ninguna fila de data/cobertura.yaml»; las tres
  evals con `no_cubierto` (07, 08, 12) pasadas a lo no declarado → «… tiene al menos una eval con no_cubierto»
  (`Should NOT be empty`); ocho copias de más de la 01 en `evals/plazos-y-territorio/` → `/tope-con-las-evals-del-repositorio`
  en rojo y `/cuatro-evals-mas` en verde (FR-008, US4.1).
- **L04 y L05** (FR-004): con la fila quitada (o el texto cambiado) en `data/cobertura.yaml` **y la sincronía hecha**
  (`go test -count=1 -run '^TestSkillsDelRepositorio$' ./internal/app/ -args -regenerar-skills`), `h32-paquetes` en
  rojo, «sin sus datos: c991808/cobertura_por_materia.md: | procedimientos sancionadores en el orden social | …» y
  `h32-paquetes.txtar:37: no match for ^filas: 33 de 33$`; `TestCoberturaDelRepositorio`, en verde (D1). Sin la
  sincronía, `h32-paquetes` sigue en verde —lee las referencias generadas, no `data/`— y quien cae es
  `TestSkillsDelRepositorio` (`make skills-check`), por la deriva.

Con cada mutante deshecho, `git status --short` del clon solo daba los cinco ficheros del cambio. El clon y su diff
(`prototipo-h451.diff`, 442 líneas) viven en el temporal y no entran en el repositorio.

## Decisiones

### D1 · `TestCoberturaDelRepositorio`: las nueve tuplas a mano y, en su orden, las filas de la base leídas (FR-001 a FR-004)

- **Decisión**: lo esperado son las nueve tuplas `(norma del precepto, apartado, norma que la rige)` escritas a mano, sin
  el texto de ningún procedimiento. Lo leído son las filas de `LeerCobertura` que son de lo administrativo por el filtro
  de hoy, escrito en el test: el `vertical` de la norma que la rige en `data/normas.yaml` y, si no lleva norma o la suya
  no tiene `vertical`, el de la norma de su precepto. La comprobación: cada fila leída de la base es una de las nueve,
  en su orden —una subsecuencia—; una de las nueve puede no estar entre las leídas. El filtro ya no se aplica a lo
  escrito.
- **Por qué**: una fila que deja la base porque gana la norma de otro paquete (la del art. 112.4 en H42, la del orden
  social en la copia) sale de lo leído y no deja hueco en lo escrito que haya que adivinar; una fila de la base que
  pierde o cambia su norma (L06) o un `LeerCobertura` que pierde `Norma` (L07) da una tupla leída que no está en su
  sitio de las nueve, y el test cae (P1). Con los datos de `main`, las nueve leídas son las nueve escritas. La
  subsecuencia es también la única forma correcta con tuplas repetidas: hay cuatro de la DA 1ª.2.c, tres iguales.
- **Alternativas descartadas**:
  - Lo esperado leído de `data/cobertura.yaml` en crudo: con L06, el YAML y `LeerCobertura` dicen lo mismo y el test
    sigue en verde; FR-003 lo exige en rojo.
  - El filtro aplicado también a lo escrito con su propia norma y las dos listas iguales (lo de `main`): la fila que gana
    norma sigue en lo escrito con la norma vacía y en lo leído no está; es el rojo del prototipo de H42 y el de SC-005
    (P1).
  - Quitar de lo escrito tantas tuplas iguales como falten en lo leído: con tuplas repetidas no se sabe cuál, y el orden
    sale mal cuando la que se va es la primera de la DA 1ª.2.c y no la última (`[T, T, U, T]` contra `[T, U, T]`): un
    rojo que no es de nadie.
  - Emparejar por posición: un paquete puede añadir una fila en medio (la de H42 que parte el sancionador tributario).
  - La subsecuencia y, además, exigir que cada tupla que falta tenga una fila leída de otro paquete con el mismo
    precepto: volvería a ver L04 en este test, pero FR-004 deja L04 y L05 a `h32-paquetes`; es un mecanismo más sin
    requisito (constitución V).
- **Mensaje**: «`<data/cobertura.yaml>`: de sus filas de administrativo, `<leídas>`, solo las `<n>` primeras son, en su
  orden, de las nueve tuplas de contracts/datos-de-plazos.md §3 de H9, `<nueve>`» (contracts/tests-genericos.md §1).

### D2 · La línea del paso 2 sale del `SKILL.md` que el caso copia (FR-005)

- **Decisión**: `mutantesDeLasOrdenes` gana `t *testing.T` y escribe el defecto de los dos casos con
  `lineaDeLaOrdenDelPasoDos(t)`: el número de la única línea de `skills/normativa/SKILL.md` del módulo
  (`raizDeLasSkills`) que es, sin los blancos de sus extremos, `bashDelPasoDos`. Cuenta con `strings.Lines` y un
  contador propio, no con `tramosDeCodigo` ni `ordenesDeLectura`. Sin una línea así, o con dos, `require` cae con una
  premisa.
- **Por qué**: `TestControlesSobreLasReferenciasAMano` copia la skill del módulo (`copiaDeLaSkill`) y los dos mutantes
  solo tocan `references/con_herramientas.md`: la línea del `SKILL.md` de la copia es la del módulo. Con `main`, 36; en
  la copia con un paquete más, 37 (P1). Contar aparte hace que un control que da otra línea caiga (P1: el mutante
  `numero + 1`).
- **Alternativas descartadas**: calcularla dentro de cada mutante desde la copia (un campo nuevo en
  `mutanteDeUnControl` para un solo uso); usar `tramosDeCodigo` (el mutante de «otra línea» lo arrastraría y el test no
  caería).

### D3 · `cuatro-evals-mas` mide una carpeta de 16 evals de modo (FR-006 a FR-008)

- **Decisión**: el subtest copia las evals del módulo al temporal, como hoy; en la carpeta de `plazos-y-territorio` de
  la copia quita cada eval que da `LeerConjunto` (`os.Root.Remove`) y escribe 16 copias de la primera eval de modo que
  decide (`!SinBinarioNiServidor && !Informativa`, el predicado de hoy), `01-copia-<descripción>` a `16-copia-…`; exige
  16 evals al releerla; y compara, como hoy, `peorCaso` con `peorCasoEsperado`, el tope de la definición leída
  (`TopeEnMinutos`) con el peor caso a la concurrencia del repositorio (cabe) y con `plazos-y-territorio` a dos (no
  cabe). Sin `topeEnMinutos` ni otra duración escrita; la constante es `evalsDeLaCarpeta = 16`.
- **Por qué**: 16 es el número del hito, medido en el spec y repetido en P1 (T1: 7 h 47 min 49 s a dos sobre 5 h 52
  min); no depende de cuántas evals tenga el módulo (P1: con ocho de más, verde). Quitar solo las evals deja en su sitio
  lo que no es una eval (una carpeta `juez/` o la lista de expresiones, si la skill las tuviera), que el peor caso lee.
- **Alternativas descartadas**: vaciar la carpeta y crearla otra vez (perdería lo que no es una eval); sumar evals a
  las del módulo (lo de `main`: cae con las 16 de H42).

### D4 · `llevaNoCubierto` deja fuera lo no declarado (FR-009)

- **Decisión**: `return eval.NoCubierto != NoCubiertoEsperado{} && !eval.NoCubierto.FueraDeLoDeclarado`, y el mensaje del
  `require.NotEmpty` de `probarLosPreceptosDeLaCobertura` dice «al menos una eval con no_cubierto de lo regulado aparte».
  `preceptosFueraDeLaCobertura` y el mutante `a999` no cambian.
- **Por qué**: es la letra de FR-009; las dos funciones que lo usan ganan el filtro a la vez (P1: L33 da su mensaje de
  hoy; las tres pasadas a lo no declarado dan el rojo del `NotEmpty`).
- **Alternativa descartada**: mirar si `Precepto` está vacío: filtraría por un síntoma, y una eval de lo regulado aparte
  sin precepto, que hoy es un defecto, dejaría de serlo.

### D5 · La copia con un paquete más ejerce los cuatro casos (FR-010 a FR-013)

- **Decisión** (contracts/copia-con-un-paquete-mas.md): el programa del guion, además de lo de hoy,
  - estrecha la fila «actuaciones y procedimientos en materia de extranjería y asilo» (DA 1ª.2.d, sin norma) a «…, salvo
    los del paquete de prueba», con su texto de `c991808` dentro (FR-010 a);
  - da a «procedimientos sancionadores en el orden social» (DA 1ª.2.c, sin norma) la norma del paquete,
    `BOE-A-2015-11722` con su precepto aplicable `a1`, el de la fila del paquete de hoy (FR-010 b);
  - pone `paquete/prueba-de-normativa.md` en `skills/normativa/references/prueba.md` (FR-010 c);
  - copia la primera eval de `evals/plazos-y-territorio/` con los prefijos `96-` a `99-` y añade
    `95-fuera-de-lo-declarado.yaml` (FR-010 d);
  - tras la sincronía, imprime cuántas líneas más abajo está la orden del paso 2 en el `SKILL.md` de `normativa` de la
    copia que en el del módulo (US6.4);
  - corre `TestControlesSobreLasReferenciasAMano` entero entre los tests enteros, `TestDefinicionDelJob` sale de ellos y
    de él solo corre `cuatro-evals-mas`, entre los subtests (FR-011).
  Los dos cambios en `data/cobertura.yaml` son sustituciones de un texto que tiene que estar una sola vez; si no está,
  la instalación falla y el guion cae en su `exec`. El guion comprueba con `grep` la fila estrechada en
  `administrativo.md`, la del orden social en `prueba.md` y no en `administrativo.md`, la línea del paquete en el
  `SKILL.md` de `normativa`, y con `cmp` que es igual a su copia congelada (FR-013).
- **Por qué esas filas**: FR-010 deja fuera las tres que cambia H42 (los tributos, el sancionador tributario y aduanero,
  el art. 112.4); la de tráfico ya lleva norma y ya pasa a `prueba.md`; la de la Seguridad Social la nombra la eval 12.
  La que gana norma es una de la DA 1ª.2.c a propósito: es el caso de tuplas repetidas de D1, el que haría una materia
  laboral con el orden social, más difícil que el de H42 (art. 112.4, sin repetir). Las dos que quedan (extranjería y
  asilo, art. 112.2) sirven para estrechar; la de extranjería y asilo, por ser de otra letra.
- **La referencia en `normativa`**: prosa sin ninguna expresión de `evals/normativa/expresiones-prohibidas.yaml`, sin
  órdenes, sin líneas ⚠ ni fechas con cifras: los controles de `TestControlesSobreLasReferenciasAMano` no dan ningún
  defecto con la copia sin mutar (P1, `tests enteros: 0`). La línea que la sincronía escribe es la de
  `internal/skills/paquetes.go:296-297`: «- Si la pregunta es de lo prueba, lee [prueba](references/prueba.md).»
- **La eval de lo no declarado**: la forma de `evals/consulta-juridica/04-despido-plazo.yaml` sin lecturas: `pregunta`,
  `activa: true`, `prohibidos` con `plazos calcular` y `no_cubierto: {fuera_de_lo_declarado: true}`. `ComprobarConjunto`
  la admite (P1: `conjunto-plazos-y-territorio` en verde en la copia).
- **La línea de la orden del paso 2 en la salida**: el guion no tiene aritmética; el programa ya imprime una línea por
  paso, y esta es una más, «orden del paso 2 de normativa, líneas más abajo: `<n>`». Las cuatro de hoy no cambian.
- **El `cmp`**: FR-013 pide las dos copias iguales byte a byte y hoy nada de `make ci` lo mira; la línea del guion
  (`KITLEGAL_SKILLS` es absoluta, `internal/app/e2e_test.go:379`) lo hace cumplir en cada `make ci` (ADR 0029).
- **Alternativas descartadas**: un guion de aceptación nuevo en `specs/043-…/aceptacion/` (D6); poner los cambios de la
  base como ficheros del archivo (`paquete/*.yaml` con filas enteras): `anadirEntradas` solo añade al final de un bloque,
  y sustituir una fila entera obligaría a otra función más larga para lo mismo; comprobar la línea con `awk` en el
  guion (otro programa en el archivo para una cuenta que el de Go ya puede hacer).

### D6 · Aceptación e2e: no aplica, y la aceptación es la copia de H45 (constitución III)

- **Decisión**: «Aceptación e2e: no aplica». El hito no cambia nada visible del binario (FR-014): no hay guion nuevo que
  describa una entrega del binario. Su aceptación, la del spec, es la copia `h45-paquete-de-mas` con los cuatro casos
  (FR-010 a FR-013), que cambia una tarea `[datos]` con su copia congelada de H45, y `make ci` en verde con `main`. Que
  la copia los ejerce lo prueba SC-005 en el registro de mutantes: con cada test devuelto a `main`, el guion en rojo (P1).
- **Por qué**: la tarea `[aceptacion]` escribe en `specs/043-h45-1-los-tests-del/aceptacion/` y `aceptacion.sh activar`
  la activa como `h45-1-<nombre>.txtar` (`scripts/workflow/aceptacion.sh:77-80`): sería una segunda copia con un paquete
  más, que el spec no pide y que doblaría su tiempo en `make ci`. El spec fija que el cambio va en una tarea `[datos]`
  («Assumptions») y la persona lo revisa al leer el informe (capa 3).
- **Orden**: cada tarea deja `make ci` en verde (`scripts/workflow/verificar.sh`, que corre el guardián y `make ci`):
  primero los cuatro tests, que con los datos de `main` comprueban lo mismo; después el guion; al final el registro.
- **Precedente**: H24, H31 y H7.3 declararon «no aplica» sin cambio del binario (`specs/017-…`, `specs/027-…`,
  `specs/013-…/plan.md`).

### D7 · El registro de mutantes, en un clon del temporal (SC-004, SC-005, SC-007)

- **Decisión**: la última tarea escribe `specs/043-h45-1-los-tests-del/mutantes.md` con la orden, lo esperado y lo visto
  de cada mutante de la tabla «Mutantes» del plan, cada uno en un clon del repositorio en `$TMPDIR` (`git clone
  --no-hardlinks`), nunca en el árbol de trabajo, como D13 de H45. L04 y L05 se aplican con la sincronía hecha (P1).
- **Por qué**: SC-004, SC-005 y SC-007 los pide el spec con ese registro como control; aplicarlos en el árbol dejaría
  cambios que el guardián rechaza.

### D8 · Ni `CHANGELOG.md`, ni ADR, ni documentación

- **Decisión**: ninguna entrada en `CHANGELOG.md` ni en `docs/`.
- **Por qué**: la Definition of Done pide `CHANGELOG.md` si cambia un comportamiento visible (`docs/ROADMAP.md` §1,
  punto 6), y aquí no cambia ninguno (FR-014); el spec deja fuera cambiar el ADR 0062 o escribir otro.

### D9 · Skills de Go

`golang-how-to` orquesta: `golang-testing` y `golang-stretchr-testify` (premisas con `require`, comparaciones con
`assert`; subtests paralelos como hoy), `golang-lint` (la configuración de hoy, sin `//nolint`; P1: `0 issues`) y
`golang-code-style` (mensajes en español, líneas de ≤ 120). Ningún paquete, tipo exportado ni dependencia nuevos
(`golang-project-layout` no aplica).

## Verificaciones en local

| # | Afirmación | Dónde se comprobó |
|---|---|---|
| V1 | El toolchain es Go 1.27 | `go version` → `go1.27.4 linux/amd64` |
| V2 | `os.Root` tiene `Remove`, `ReadFile` y `WriteFile` | `go doc os.Root` (`func (r *Root) Remove(name string) error`, entre otras) |
| V3 | `-run` parte la expresión por las barras sin corchetes y cada parte casa con un nivel del nombre | `go help testflag`, líneas 132-135 |
| V4 | testscript v1.16.0 tiene `cmp file1 file2`, `grep [-count=N] pattern file` y `stdout` | `go.mod:11`; `go doc github.com/rogpeppe/go-internal/testscript`, líneas 130, 171 y 200 |
| V5 | `KITLEGAL_SKILLS` es una ruta absoluta en los guiones | `internal/app/e2e_test.go:379` (`filepath.Abs`) y `:396` |
| V6 | `make skills-check` solo compara (sin `-regenerar-skills`) y `make ci` lo corre | `Makefile:136-137` y `:314` |
| V7 | `make test` y `make test-integration` dejan `coverage.out` y `coverage-integration.out`, que Git ignora; los demás objetivos de `make ci` no escriben en el árbol, y `vuln` consulta su base por red | `Makefile:97`, `:101`, `:124-133`, `:209-226` y `:127` («requiere red»); `.gitignore` (`/*.out`, `/coverage.*`) |
| V8 | Cada tarea del run pasa el guardián de diff y `make ci` | `scripts/workflow/verificar.sh` (cabecera y modo `tarea`) |
| V9 | `testdata/` solo con `[datos]`; la suite congelada que se protege es la del hito en curso (`$d/gates/aceptacion-congelada.json`) | `scripts/workflow/guardian-diff.sh:7-21`, `:75` y `:108-117` |
| V10 | La sincronía escribe una línea por paquete con `references/<paquete>.md` en la skill | `internal/skills/paquetes.go:151-204` y `:291-301` |
| V11 | `golangci-lint` se ejecuta con `go tool -modfile=tools/golangci-lint/go.mod` | `Makefile:41` y `:112-113` |
| V12 | `git diff --name-only <rango> -- '*.go' ':!*_test.go'` lista los `.go` de cualquier carpeta salvo los de test (quickstart Q3) | ejecutado sobre `e714d74f~1..e714d74f` (H45): solo `internal/evals/dependencias.go`, `grabaciones.go`, `medida.go`, `sesiones.go` e `internal/skills/regiones.go` |
| V13 | Un `git clone --no-hardlinks` del repositorio tiene `origin/main` en el `main` local (quickstart Q8) | `git -C <clon> rev-parse origin/main` → `e1555a7d…`, igual que `git rev-parse main` |

## Datos externos

Ninguno. El hito no graba nada ni deriva nada a `data/`: no hay manifiesto `grabaciones.json` ni test `TestGrabar*`, y
el paso `grabar_datos` no tiene qué hacer. La copia usa `BOE-A-2015-11722`, bloque `a1`, el precepto aplicable que ya
usan la fila de tráfico de `data/cobertura.yaml:39` y la fila del paquete de hoy.

## Supuestos no verificados

- Que el prototipo de H42 pase con estos cuatro tests y sin ningún `.go`: sus datos no están en el repositorio. Lo
  comprueba la persona antes de lanzar H42 (spec, «Fuera de alcance»).
- Que `make ci` entero (con `vuln`, `goreleaser-check` y `secrets`) quede en verde: en P1 se corrieron `go test` de los
  tres paquetes que el hito toca y `golangci-lint` de los dos con código; `make ci` lo corre cada tarea del run.
