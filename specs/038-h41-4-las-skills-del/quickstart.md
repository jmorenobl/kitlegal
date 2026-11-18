# Quickstart: validar H41.4

Desde la raíz del repositorio, en la cabeza de la rama `038-h41-4-las-skills-del` con todo confirmado, y con la
toolchain de `go.mod`. Los escenarios 1 a 6 no usan red (salvo `govulncheck` en el 6), ni modelo, ni credencial, y no
dejan nada en el árbol, el índice ni el historial salvo lo que dice cada uno. El 7 lo hace el workflow y el 8, la
persona, fuera del run. La base de la rama es `ab951c227e1fc8acac8bd6b23534e275e4a36a76`.

## 1. Las descripciones de `skills/` son las de la variante, con sus límites (US1; FR-001, FR-002, FR-004; SC-001, SC-003)

```sh
go test -count=1 -run '^TestDescripcionesDelProducto$' -v ./internal/evals/
go test -count=1 -run '^TestSkillsDelRepositorio$' ./internal/app/
```

Verde: `--- PASS: TestDescripcionesDelProducto` con sus nueve subtests, las nueve filas de
contracts/controles.md §1.3 —la rama, sin ninguna línea, y los ocho mutantes, cada uno con sus líneas— y `ok` en los
dos paquetes. Las medidas, en
contracts/skills.md §3.

## 2. Lo que instala el binario (US1; FR-001, FR-002; SC-001)

```sh
go test -count=1 -run '^TestEntregaDelHito$/^h41-4-skills-instaladas$' -v ./internal/app/
destino="$(mktemp -d)"
go run ./cmd/kitlegal skills install --dir "$destino/skills" --json
sed -n '1,8p' "$destino/skills/normativa/SKILL.md"
rm -rf "$destino"
```

Verde: `--- PASS: TestEntregaDelHito/h41-4-skills-instaladas` (el guion de contracts/aceptacion.md, que el workflow
activa al terminar las tareas). La instalación sale con `{"ok":true,…}` y las cuatro skills `instalada`, y el
`SKILL.md` de `normativa` abre con `---`, `name: normativa` y el bloque de contracts/skills.md §2. Lo único que se
escribe es el directorio temporal, que la última orden borra.

## 3. Nada más de las skills cambia (US2; FR-003, FR-005; SC-002)

```sh
go test -count=1 -run '^(TestSkillsDelProductoDesdeControl|TestConsultaJuridicaDesdeLaBase)$' -v ./internal/evals/
git diff --stat ab951c227e1fc8acac8bd6b23534e275e4a36a76 -- skills/
```

Verde: los dos tests, con «desde control», «movimientos» y «frontmatter» de las tres familias y los seis casos de
contracts/controles.md §2.3. El diff: los cuatro `SKILL.md` y nada más, `4 files changed, 16 insertions(+), 31
deletions(-)` (research V5).

## 4. Las suites congeladas, el laboratorio y lo generado siguen como estaban (US3; FR-006, FR-007, FR-009)

```sh
go test -count=1 -run '^(TestDescripcionesCortas|TestDescripcionesCortasPorLaEstructura|TestConLaDescripcion|TestVariantesDeControl|TestEleccionDeLosMutantes)$' ./internal/evals/
git diff --quiet ab951c227e1fc8acac8bd6b23534e275e4a36a76 -- laboratorio/ evals/ web/ mcp/ internal/empaquetado/ .goreleaser.yaml .github/workflows/evals.yml && echo "sin cambios"
git diff --name-only ab951c227e1fc8acac8bd6b23534e275e4a36a76 -- '*_test.go'
make skills-check
make skills-sync
git status --porcelain -- skills/
```

Verde: `ok`; `sin cambios`; como únicos tests tocados, `internal/evals/descripciones_del_producto_test.go` (nuevo) e
`internal/evals/laboratorio_familias_test.go`; `make skills-check` en verde; y `git status` sin ninguna línea tras
`make skills-sync`. Si `make skills-sync` cambiara algo, ese es el defecto de FR-007; `git restore -- skills/` deja el
árbol como estaba.

## 5. `CHANGELOG.md` (FR-008)

```sh
awk '/^## \[Unreleased\]/ {u = 1; next} /^## \[/ {u = 0} u && /^### / {s = $0} u && s == "### Cambiado"' CHANGELOG.md | grep -A5 'Las cuatro skills del producto llevan'
```

Sale la viñeta de contracts/skills.md §4: las cuatro skills, la variante `producto-descripciones-cortas`, que nada más
cambia y el porqué. Sin control en `make ci` (FR-008): lo lee la revisión.

## 6. `make ci` (SC-004)

```sh
make ci
```

Termina con `ci: todos los controles en verde`. Necesita red para `govulncheck` (`make vuln`); el resto, no.

## 7. La medición del cierre (FR-010, SC-005)

No se lanza a mano: tras la revisión final, el workflow empuja la rama, abre la propuesta de cambio y mide CI y evals
sobre la cabeza. Con el cierre hecho, cada umbral de «Controles de umbral» se lee sin modelo:

```sh
jq -c '{skill, veredicto, decide: [.umbrales[] | select(.decide) | {nombre, cumple}]}' specs/038-h41-4-las-skills-del/gates/evals/*.json
```

Verde: `veredicto` aprobado en las cuatro skills, y `cumple: true` en cada umbral con `decide`, también
`habla_en_primera_o_segunda_persona` en las cuatro. Si algo cae, la reparación no cambia ninguna descripción para que
pase y lo deja dicho en el informe (FR-010).

## 8. El sondeo de aceptación de la persona (FR-011, SC-006)

Fuera del run y antes de fusionar, en una terminal interactiva, con `claude --version` en 2.1.326 (la de
`VERSION_DE_CLAUDE_CODE`) y la credencial del sondeo (`scripts/evals-sondeo-llavero.sh`). Las dos filas, el mismo día,
una detrás de otra. Crea el árbol de trabajo `../kitlegal-h41-4-base`, que la última orden quita, y los sondeos en
`../kitlegal-sondeos/h41-4/`, que no pueden existir antes; no toca el árbol de la rama.

```sh
git worktree add --detach ../kitlegal-h41-4-base ab951c227e1fc8acac8bd6b23534e275e4a36a76
(cd ../kitlegal-h41-4-base && make evals-laboratorio VARIANTES=normativa-producto MODOS=herramienta SESIONES=320 \
  MODELO=claude-sonnet-5-5 CLAUDE_CODE=2.1.326 CONCURRENCIA=5 DIRECTORIO=../kitlegal-sondeos/h41-4/control EVAL=04 \
  INSTALACION=producto)
make evals-laboratorio VARIANTES=normativa-producto MODOS=herramienta SESIONES=320 MODELO=claude-sonnet-5-5 \
  CLAUDE_CODE=2.1.326 CONCURRENCIA=5 DIRECTORIO=../kitlegal-sondeos/h41-4/producto EVAL=04 INSTALACION=producto
make laboratorio-tabla DIRECTORIO=../kitlegal-sondeos/h41-4/control > ../kitlegal-sondeos/h41-4/tabla-control.md
make laboratorio-tabla DIRECTORIO=../kitlegal-sondeos/h41-4/producto > ../kitlegal-sondeos/h41-4/tabla-producto.md
git worktree remove ../kitlegal-h41-4-base
```

`DIRECTORIO` es relativo a la raíz de cada árbol, y los dos árboles son hermanos, así que los dos sondeos quedan en el
mismo `../kitlegal-sondeos/h41-4/`. Una tabla por sondeo: las dos filas tienen la misma clave de fila y, con el
directorio de los dos, la tabla las juntaría en una de 640 (research V17).

**Lectura.** En cada tabla, «No activan `normativa` antes de la primera herramienta» es 320 menos «Activan antes», y su
intervalo de Wilson al 95 % es el de «Activan antes» restado de 100 % por los dos extremos (el intervalo de Wilson de
1 − p es el de p reflejado). El hito se acepta solo si el de `producto` queda por debajo del de `control` sin tocarse;
«Activan otra skill» se da y no decide. La persona deja las dos tablas y su lectura en `evidencias/` y en la propuesta
de cambio; el sondeo no se repite para ver si sale en verde (ADR 0047, punto 4).
