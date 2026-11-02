# T003 · intento 1: sin verde por una ruta que la tarea no declaraba

## Causa

La tarea pide regenerar `skills/boe-legislacion/references/normas.md` con su fila nueva (la Ley 50/1980), y el guion
de H31.1 `internal/app/testdata/script/h31-1-h31-1-sitio-en-el-cuerpo.txtar`, que T004 retira (FR-040), compara esa
referencia byte a byte con la de la skill medida bajo `evidencias/sondeos-20261026/brazo-W/skill/`, que no se puede
escribir. Con la fila nueva, `TestEntregaDelHito` sale en rojo en `make test` y en `make test-integration`. La línea de
T003 no declaraba el guion, así que este intento no podía tocarlo (las rutas quedan congeladas en
`gates/tarea-actual.json`): la tarea queda `[ ]` y su línea de `tasks.md` gana la ruta y el arreglo («Guion de H31.1»).

Salida de `go test -count=1 -run '^TestEntregaDelHito$' ./internal/app/` en el árbol del intento:

```text
--- FAIL: TestEntregaDelHito (0.01s)
    --- FAIL: TestEntregaDelHito/h31-1-h31-1-sitio-en-el-cuerpo (0.06s)
            > cmp $RAMA/references/normas.md $MEDIDA/references/normas.md
            FAIL: testdata/script/h31-1-h31-1-sitio-en-el-cuerpo.txtar:48: /Users/jorge/Projects/kitlegal/skills/boe-legislacion/references/normas.md and /Users/jorge/Projects/kitlegal/skills/../evidencias/sondeos-20261026/brazo-W/skill/references/normas.md differ
```

## Lo que ya está hecho en el árbol (sin commit) y en verde

- `schemas/normas.yaml.json`: `vertical`, cadena con el patrón `^[a-z][a-z0-9-]*$`.
- `internal/skills/normas.go`: `Norma.Vertical` (`yaml:"vertical"`); los ejemplos de los comentarios pasan a «campo no
  declarado: estado».
- `internal/skills/normas_test.go`: `con-vertical` es válida (LPAC con `vertical: fiscal` y la LCSP sin ella);
  `vertical-con-otra-forma` (`vertical: Fiscal`) da `BOE-A-2015-10565: vertical: 'Fiscal' does not match pattern
  '^[a-z][a-z0-9-]*$'`; `defectos-en-dos-normas-en-orden-de-linea` usa `estado: vigente`; `TestNormasDelRepositorio`
  gana el subtest `paquete-de-lo-administrativo` (las diez de research D5, exactamente).
- `internal/skills/sincronia_test.go`: `norma-con-vertical` sale de `probarDefectos` y es un subtest de
  `TestRegenerarYComparar` sin defectos; los otros tres casos que usaban `vertical: fiscal` como campo no declarado
  usan `estado: vigente` (`norma-con-campo-no-declarado`, `vertebrales-sin-normas-con-campo-no-declarado`, y los dos
  de orden).
- `data/normas.yaml`: cabecera de `vertical`, `vertical: administrativo` en las diez normas y `BOE-A-1980-22501` con
  el título «Ley 50/1980, de 8 de octubre, de Contrato de Seguro.» y el rango «Ley» de la respuesta grabada
  (`GET_…_q_limit_10_query_7B_22query_22_3A_20_-41c36196.json`), `materias: [seguros]`.
- `internal/evals/grabaciones_test.go`: la excepción de T002 retirada (`git diff c991808` no toca
  `reproducirBusqueda` ni `verificarLasNormasDeLaCobertura`); `verificarIdentificadores` lee los dos manifiestos y
  resuelve cada norma con el de H5 o, si ninguna entrada suya la resuelve, con el de H9; subtest
  `norma-de-h9-sin-entrada`; `norma-sin-entrada` quita la LPAC de las dos copias (el de H9 también la resuelve).
- `skills/boe-legislacion/references/normas.md`: regenerada con `make skills-sync` (una fila nueva).
- Dos supuestos `[interno] T003` en `gates/supuestos.md`.

Medido en este intento: `go test -count=1 ./internal/skills/ ./internal/evals/` en verde (con
`TestPreceptosDeLosPlazos` y `TestIdentificadoresDeLasNormas`), `make skills-check` en verde, `make fmt-check lint`
con 0 issues, y `make test-tiempos vuln schema-check goreleaser-check secrets mod-verify mod-tidy-check` en verde. Solo
`TestEntregaDelHito` falla, por el guion de arriba.

## El arreglo para el intento 2, comprobado en una copia desechable

En `internal/app/testdata/script/h31-1-h31-1-sitio-en-el-cuerpo.txtar`, tras el párrafo de cabecera que termina en
«no nombra las normas de las evals por materia.», un párrafo nuevo:

```text
#
# references/normas.md se genera desde data/normas.yaml, que gana en H32 la Ley
# 50/1980 (FR-035 de H32): es la de la skill medida sin otra diferencia que su
# fila, que se quita antes de comparar. H32 retira este guion (FR-040 de H32).
```

y la línea `cmp $RAMA/references/normas.md $MEDIDA/references/normas.md` pasa a ser:

```text
exec grep -v -F '| `BOE-A-1980-22501` |' $RAMA/references/normas.md
cp stdout normas-sin-la-ley-50-1980.md
cmp normas-sin-la-ley-50-1980.md $MEDIDA/references/normas.md
```

## Intento 2: en verde

Con el árbol del intento 1 intacto, el guion sale en rojo en su línea 48 (reproducido en esta sesión); con el cambio de
arriba aplicado en el árbol, `TestEntregaDelHito` da `ok` (70 tests de `internal/app`), y el mutante (el título de la
LOPJ sin su punto final en `normas.md`) lo pone en rojo en la línea 54, la del `cmp` nuevo; `make skills-sync` devuelve
`normas.md` al mismo blob (`8c38e46e`). Medido en primer plano, los doce objetivos de `make ci` en cuatro órdenes
consecutivas, por el tope de diez minutos de cada orden de la sesión: `make fmt-check lint` (0 issues), `make test`,
`make test-integration test-tiempos` y `make vuln schema-check skills-check goreleaser-check secrets mod-verify
mod-tidy-check`, todos en verde. `go test -count=1 ./internal/skills/ ./internal/evals/` en verde (3293 tests), con
`TestPreceptosDeLosPlazos` y los cinco subtests de `TestIdentificadoresDeLasNormas`. Sin supuestos nuevos.

## Comprobación del arreglo en el intento 1

Sobre una copia del árbol del intento (`rsync -a --exclude .git` a `$TMPDIR`) con ese cambio,
`go -C <copia> test -count=1 -run '^TestEntregaDelHito$' ./internal/app/` da `ok`; y con un mutante en la copia (el
título de la LOPJ sin su punto final en `normas.md`) el guion sale en rojo en la línea del `cmp`: «normas-sin-la-ley-50-1980.md
and …/brazo-W/skill/references/normas.md differ». Falta, con el guion cambiado en el árbol, el `make ci` completo.
