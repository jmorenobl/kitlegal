# Quickstart: validar H20

Escenarios que prueban la entrega, de dentro afuera. Todos se ejecutan desde la raíz del repositorio, sin red ni
credenciales: los tests usan lo grabado de H4 (`internal/source/boe/testdata/`) y directorios temporales, y los guiones
e2e, el arnés de `internal/app/e2e_test.go`, que construye sus binarios en un directorio temporal. Ninguno escribe en
el árbol de trabajo, el índice ni el historial de git, salvo el 7, que lo dice. Cada uno dice qué da **hoy**, en
`main` (`0b2a21d`), y qué da **tras el hito**.

Formas que se usan: un test, `go test -count=1 -run '^<Test>$' ./<paquete>/`; varios, `-run '^(<A>|<B>)$'`; un subtest,
`-run '^<Test>$/^<subtest>$'` (el `-run` separa por `/`, así que la alternancia va dentro de un nivel).

## 0. Precondiciones

```bash
go version                       # go1.27.2 (go.mod: go 1.27.0, toolchain go1.27.2)
git rev-parse --abbrev-ref HEAD  # 024-h20-boe-articulo-fecha
```

## 1. Lo que ya vale hoy

```bash
go test -count=1 -run '^TestEntregaDelHito$/^boe-codigos$' ./internal/app/
make skills-check
```

**Hoy y tras el hito**: `ok` (el primero tardó 2,7 s en esta sesión, research M7). Es la base: la lectura sin fecha y
la skill de hoy.

## 2. La lectura fechada en el adaptador

```bash
go test -count=1 -run '^(TestValidarFecha|TestLeerBloque|TestArticuloEnUnaFecha|TestArticuloEnUnaFechaYLaCache|TestFallosNoSeGuardan|TestObservadoDeBoe|TestGolden|TestArticulo)$' ./internal/source/boe/
```

**Hoy**: `ok` con los tests de hoy (`TestLeerBloque`, `TestFallosNoSeGuardan`, `TestObservadoDeBoe`, `TestGolden`,
`TestArticulo`); los nuevos no existen. **Tras el hito**: `ok`, con los siete casos del golden de
contracts/articulo-fecha.md §2, los de FR-004 con respuestas construidas en el test, el de los avisos (FR-011), los de
la caché (FR-020 a FR-022) y los tres fallos de la fuente con fecha, iguales que sin ella y sin nada guardado (FR-007);
y los golden de hoy, sin cambiar ni un byte (FR-012).

## 3. El grafo

```bash
go test -count=1 -run '^(TestEntregaDelMontador)$' ./internal/cli/
go test -count=1 -run '^(TestApplySoloLoNuevo|TestIntegracionLecturaFechada)$' ./internal/graph/
go test -count=1 -run '^(TestLecturaFechadaEnElGrafo|TestLaSalidaDeBoeNoCambiaConElGrafo|TestLaEntregaLlevaLaProcedenciaDelSobre)$' ./internal/app/
```

**Hoy**: `ok` con los de hoy. **Tras el hito**: `ok`; el lote lleva `SoloLoNuevo`; con él, lo que `world.db` ya guarda
no cambia y lo nuevo entra con la procedencia del sobre; la fila de lecturas y `graph check` quedan como sin la lectura
fechada, también en un bloque sin fila (FR-030 a FR-032, FR-086). Escriben solo en `t.TempDir()`.

## 4. La suite de aceptación

```bash
go test -count=1 -run '^TestEntregaDelHito$/^h20-' ./internal/app/
```

**Hoy**: `ok … [no tests to run]` (comprobado en esta sesión): la suite está en
`specs/024-h20-boe-articulo-fecha/aceptacion/` y el workflow la activa como `internal/app/testdata/script/h20-*.txtar`
al terminar las tareas. Para ver el rojo de hoy de un guion antes de activarlo, el paso del workflow
`scripts/workflow/aceptacion.sh rojo-primero` lo copia como `zz-rojo-<nombre>.txtar` y lo retira. **Tras el hito**: los
seis guiones en verde.

## 5. Esquemas, `--describe` y la skill

```bash
make schema-check
make skills-check
go test -count=1 -run '^(TestHerramientasDelServidor|TestSalidaDeBoeContraSchemas)$' ./internal/app/
```

**Hoy**: `ok`. **Tras el hito**: `ok` con `schemas/bloque.json` regenerado —la entrada de `articulo` con `fecha` en
`x-banderas` y su salida con `fecha_pedida` y `redaccion_posterior`; la parte de `articulos`, igual (FR-012, FR-085)—,
la herramienta `boe_articulo` con el parámetro `fecha`, y `SKILL.md` por debajo de 300 líneas con su tabla regenerada
(FR-048). Para ver el `--describe` sin red ni caché (necesita la norma y el bloque, como en `boe-verbos.txtar`):

```bash
go run ./cmd/kitlegal boe articulo BOE-A-2015-10565 a21 --describe
```

**Hoy** (ejecutado en esta sesión): el esquema con `boe.Articulo` y sin `fecha` ni `x-banderas`, y el árbol de trabajo
sin cambios. **Tras el hito**: con `"fecha"` en la entrada y en `x-banderas`, y la salida en `boe.ArticuloFechado`.
`--describe` no abre la caché ni pide nada.

## 6. Las evals y el umbral

```bash
go test -count=1 -run '^(TestExtraerRedaccionesFechadas|TestFormatoDeLasRedaccionesFechadas|TestJuzgarLecturasFechadas)$' ./internal/evals/
go test -count=1 -run '^(TestUmbralesDeBoeLegislacion|TestUmbralesDeLasTresSkills|TestDefinicionDelJob|TestPlanEnDosModos|TestConsultasNecesarias|TestConjuntoDeEvals)$' ./internal/evals/
go test -count=1 -run '^TestEvalsDelRepositorio$' ./internal/evals/
```

**Hoy**: `ok` con los de hoy (los nuevos no existen; `TestUmbralesDeLasTresSkills` espera trece en
`boe-legislacion`). **Tras el hito**: `ok`; los quince umbrales de `boe-legislacion` con once que deciden; los casos
de FR-089; el peor caso del job, 18 977 s, por debajo de 352 minutos (FR-050); y las veintitrés evals válidas, con su
caché preparada y comprobada sin red, también la lectura de cada bloque con `--fecha 9999-12-31` (FR-053). Ninguno abre
una sesión con modelo (FR-072).

## 7. Los mutantes (FR-090)

En un árbol aparte, fuera del repositorio, para no dejar ningún mutante en el de trabajo:

```bash
git worktree add --detach "$TMPDIR/h20-mutantes" HEAD
# en "$TMPDIR/h20-mutantes", cada mutante de contracts/evals-redaccion-fechada.md §5, de uno en uno:
go -C "$TMPDIR/h20-mutantes" test -count=1 -run '^TestArticuloEnUnaFecha$' ./internal/source/boe/
go -C "$TMPDIR/h20-mutantes" test -count=1 -run '^TestObservadoDeBoe$' ./internal/source/boe/
go -C "$TMPDIR/h20-mutantes" test -count=1 -run '^TestLecturaFechadaEnElGrafo$' ./internal/app/
go -C "$TMPDIR/h20-mutantes" test -count=1 -run '^(TestUmbralesDeBoeLegislacion|TestUmbralesDeLasTresSkills)$' ./internal/evals/
git worktree remove --force "$TMPDIR/h20-mutantes"
```

**Deja**: mientras existe, la entrada del árbol en `.git/worktrees/`, que `git worktree remove` retira; nada en el
árbol de trabajo, el índice ni el historial. **Tras el hito**: con cada mutante, el test de su fila falla; sin
mutantes, todos pasan. Lo que da cada uno queda en `specs/024-h20-boe-articulo-fecha/mutantes.md`.

## 8. Todo

```bash
make ci
```

**Tras el hito**: verde, con `schema-check`, el drift de `references/` y las reglas del conjunto (FR-080).

## 9. El cierre (lo hace el workflow, no una tarea)

Tras la revisión final, el workflow empuja la rama, abre la propuesta de cambio y mide el job de evals sobre la cabeza.
En el `informe.json` de `boe-legislacion`: los dos `redaccion_sin_fechar:claude-sonnet-5-5:<modo>` con `decide: true`,
0 de 6; `sin_activar`, 0 de 60 en cada modo; `afirma_lo_no_leido:claude-sonnet-5-5`, como mucho 3 de 120; la medida del
juez y las duraciones, cumplidas; ninguna respuesta sin juzgar; `red` vacío; veredicto aprobado (SC-001). El informe
final da el umbral nuevo «comprobado por su control» (SC-002).

## 10. Una persona, después del run (SC-003)

Con la release publicada, en la app de escritorio de Claude con el plugin: «¿qué exigía el artículo 118 de la LCSP para
un contrato menor el 1 de junio de 2019?». La respuesta cita la redacción vigente desde el 9 de marzo de 2018 y avisa
de que hoy es la de `BOE-A-2020-1651` (el Real Decreto-ley 3/2020), desde el 6 de febrero de 2020. Nadie lo hace dentro
del run.
