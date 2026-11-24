# Quickstart: comprobar H45.1

Escenarios para comprobar la entrega con el hito implementado y commiteado en la rama `043-h45-1-los-tests-del`. Se
ejecutan desde la raíz del repositorio, sin red ni modelo. Q1 a Q7 solo leen el árbol, salvo lo que declara cada uno; Q8
trabaja en un clon del temporal y no toca el árbol de trabajo, el índice ni el historial. Contratos:
[tests-genericos.md](contracts/tests-genericos.md) y [copia-con-un-paquete-mas.md](contracts/copia-con-un-paquete-mas.md).

## Q1 · `make ci` en verde (SC-001; US1.1)

```bash
make ci
```

Esperado: termina con `ci: todos los controles en verde`. Deja `coverage.out` y `coverage-integration.out` en la raíz,
que `.gitignore` ignora (`/*.out`, `/coverage.*`); la caché de Go y la de `golangci-lint`, fuera del repositorio. El
objetivo `vuln` consulta la base de vulnerabilidades por red (`Makefile:127`): sin ella, cae por la fuente, no por el
hito, y los demás se comprueban con `make fmt-check lint test test-integration test-tiempos schema-check skills-check`.

## Q2 · `make skills-check` sin diferencias (SC-002; US1.2)

```bash
make skills-check
```

Esperado: `ok` en `./internal/app/`, `./internal/skills/` y `./internal/evals/`. No regenera nada: compara.

## Q3 · El diff contra `main`, solo tests y los dos guiones (US1.3; FR-014)

```bash
git diff --name-only main...HEAD -- data evals skills .github
git diff --name-only main...HEAD -- '*.go' ':!*_test.go'
git diff --name-only main...HEAD
```

Esperado: las dos primeras no imprimen nada; la tercera, solo `internal/skills/cobertura_test.go`,
`internal/evals/conjunto_test.go`, `internal/evals/definicion_test.go`,
`internal/app/testdata/script/h45-paquete-de-mas.txtar`, `specs/042-h45-el-molde-admite/aceptacion/paquete-de-mas.txtar`,
`docs/roadmap/done/H45.1-los-tests-del-molde-sin-el-estado-de-hoy.md` (el commit con que se lanzó el hito) y ficheros de
`specs/043-h45-1-los-tests-del/`.

## Q4 · La copia con un paquete más, en verde con los cuatro casos (SC-003; US6.1, US6.3, US6.4; FR-010, FR-011)

```bash
go test -count=1 -v -run '^TestEntregaDelHito$/^h45-paquete-de-mas$' ./internal/app/
```

Esperado: `--- PASS: TestEntregaDelHito/h45-paquete-de-mas`, y en su registro `copia: hecha`, `sincronía: 0`,
`orden del paso 2 de normativa, líneas más abajo: 1`, `tests enteros: 0` y `subtests: 0`, y los `grep` de la fila
estrechada, la del orden social y la línea del paquete en `normativa` sin fallo. Escribe solo en el `$WORK` de
testscript, en el temporal, y en la caché de Go.

## Q5 · Los dos guiones, iguales byte a byte (FR-013; US6.5)

```bash
cmp internal/app/testdata/script/h45-paquete-de-mas.txtar specs/042-h45-el-molde-admite/aceptacion/paquete-de-mas.txtar
```

Esperado: sin salida y código 0.

## Q6 · Con `main`, el defecto en la línea 36 (US1.4; FR-005)

```bash
grep -n 'kitlegal boe articulo BOE-A-1978-31229 a140 --json && kitlegal graph check' skills/normativa/SKILL.md
go test -count=1 -v -run '^TestControlesSobreLasReferenciasAMano$/^(forma-en-una-referencia-con-otros-bloques|forma-en-ninguna-referencia)$' ./internal/evals/
```

Esperado: la primera da `36:` y la línea; la segunda, los dos casos en `PASS` (su defecto esperado es exacto, con la línea
que da el `SKILL.md`).

## Q7 · Los otros tres, en verde en el árbol (US2, US4, US5)

```bash
go test -count=1 -v -run '^TestCoberturaDelRepositorio$' ./internal/skills/
go test -count=1 -v -run '^TestDefinicionDelJob$/^cuatro-evals-mas$' ./internal/evals/
go test -count=1 -v -run '^TestEvalsDelRepositorio$/^precepto-de-la-cobertura$' ./internal/evals/
```

Esperado: `PASS` en los tres.

## Q8 · Los mutantes, en un clon del temporal (SC-004, SC-005, SC-007; FR-003 a FR-009, FR-012, FR-013)

Preparación (lo único que crea es `$TMPDIR/h451`; el clon tiene `origin/main`, el `main` del repositorio):

```bash
git clone --no-hardlinks -q . "$TMPDIR/h451"
```

Cada mutante: aplicarlo, correr su control y deshacerlo; tras deshacerlo, `git -C "$TMPDIR/h451" status --short` no da
nada.

**L06** (FR-003): rojo, «solo las 5 primeras».

```bash
sed -i 39d "$TMPDIR/h451/data/cobertura.yaml"
go -C "$TMPDIR/h451" test -count=1 -run '^TestCoberturaDelRepositorio$' ./internal/skills/
git -C "$TMPDIR/h451" checkout -- data/
```

**L07** (FR-003): rojo, `expected: 9`, `actual: 0`.

```bash
sed -i 's|^\treturn ValidarDocumentoYAML\[Cobertura\](contenido, esquema)$|\tc, err := ValidarDocumentoYAML[Cobertura](contenido, esquema)\n\tfor i := range c.ReguladosAparte {\n\t\tc.ReguladosAparte[i].Norma = nil\n\t}\n\n\treturn c, err|' "$TMPDIR/h451/internal/skills/cobertura.go"
go -C "$TMPDIR/h451" test -count=1 -run '^TestCoberturaDelRepositorio$' ./internal/skills/
git -C "$TMPDIR/h451" checkout -- internal/skills/cobertura.go
```

**L04 y L05, con la sincronía** (FR-004; SC-007): `h32-paquetes` en rojo, `no match for ^filas: 33 de 33$`; y
`TestCoberturaDelRepositorio`, en verde.

```bash
sed -i '35,36d' "$TMPDIR/h451/data/cobertura.yaml"
go -C "$TMPDIR/h451" test -count=1 -run '^TestSkillsDelRepositorio$' ./internal/app/ -args -regenerar-skills
go -C "$TMPDIR/h451" test -count=1 -v -run '^TestEntregaDelHito$/^h32-paquetes$' ./internal/app/
git -C "$TMPDIR/h451" checkout -- data/ skills/
sed -i 's/procedimientos sancionadores en el orden social$/procedimientos sancionadores en el orden laboral/' "$TMPDIR/h451/data/cobertura.yaml"
go -C "$TMPDIR/h451" test -count=1 -run '^TestSkillsDelRepositorio$' ./internal/app/ -args -regenerar-skills
go -C "$TMPDIR/h451" test -count=1 -v -run '^TestEntregaDelHito$/^h32-paquetes$' ./internal/app/
git -C "$TMPDIR/h451" checkout -- data/ skills/
```

**P6 y otra línea** (FR-005): los dos casos en rojo.

```bash
sed -i 's/defectosDeLasOrdenes(delSkillMd, deLasReferencias)/defectosDeLasOrdenes(delSkillMd, nil)/' "$TMPDIR/h451/internal/evals/conjunto_test.go"
go -C "$TMPDIR/h451" test -count=1 -v -run '^TestControlesSobreLasReferenciasAMano$/^(forma-en-una-referencia-con-otros-bloques|forma-en-ninguna-referencia)$' ./internal/evals/
git -C "$TMPDIR/h451" checkout -- internal/evals/conjunto_test.go
sed -i 's/tramos = append(tramos, tramoDeCodigo{linea: numero, texto: recortada})/tramos = append(tramos, tramoDeCodigo{linea: numero + 1, texto: recortada})/' "$TMPDIR/h451/internal/evals/conjunto_test.go"
go -C "$TMPDIR/h451" test -count=1 -v -run '^TestControlesSobreLasReferenciasAMano$/^(forma-en-una-referencia-con-otros-bloques|forma-en-ninguna-referencia)$' ./internal/evals/
git -C "$TMPDIR/h451" checkout -- internal/evals/conjunto_test.go
```

**T1 y L29** (FR-006, FR-007): `cuatro-evals-mas` en rojo; con T1, «"5h52m0s" is not greater than or equal to …».

```bash
sed -i '260s/concurrencia: 3/concurrencia: 2/' "$TMPDIR/h451/.github/workflows/evals.yml"
sed -i 's/"plazos-y-territorio": {Concurrencia: 3, ObjetivoDeDuracion: 0},/"plazos-y-territorio": {Concurrencia: 2, ObjetivoDeDuracion: 0},/' "$TMPDIR/h451/internal/evals/definicion_test.go"
go -C "$TMPDIR/h451" test -count=1 -v -run '^TestDefinicionDelJob$/^cuatro-evals-mas$' ./internal/evals/
git -C "$TMPDIR/h451" checkout -- .github/ internal/evals/definicion_test.go
sed -i 's/Modos:               \[\]Modo{ModoOrden, ModoHerramienta},/Modos:               []Modo{ModoOrden},/' "$TMPDIR/h451/internal/evals/definicion.go"
go -C "$TMPDIR/h451" test -count=1 -v -run '^TestDefinicionDelJob$/^cuatro-evals-mas$' ./internal/evals/
git -C "$TMPDIR/h451" checkout -- internal/evals/definicion.go
```

**FR-008**: con ocho evals de más, `/tope-con-las-evals-del-repositorio` en rojo y `/cuatro-evals-mas` en verde.

```bash
for n in 90 91 92 93 94 95 96 97; do cp "$TMPDIR/h451/evals/plazos-y-territorio/01-territorio-municipio-cubierto.yaml" "$TMPDIR/h451/evals/plazos-y-territorio/$n-copia.yaml"; done
go -C "$TMPDIR/h451" test -count=1 -v -run '^TestDefinicionDelJob$/^(tope-con-las-evals-del-repositorio|cuatro-evals-mas)$' ./internal/evals/
git -C "$TMPDIR/h451" clean -fdq -- evals/
```

**L33 y todas de lo no declarado** (FR-009): rojo, con el mensaje de `a999` de la 12; y con `Should NOT be empty`.

```bash
sed -i '15s/bloque: daprimera}/bloque: a999}/' "$TMPDIR/h451/evals/plazos-y-territorio/12-cobertura-seguridad-social-sin-leer-la-da-primera.yaml"
go -C "$TMPDIR/h451" test -count=1 -v -run '^TestEvalsDelRepositorio$/^precepto-de-la-cobertura$' ./internal/evals/
git -C "$TMPDIR/h451" checkout -- evals/
sed -i -e 's/^  norma: .*$/  fuera_de_lo_declarado: true/' -e '/^  precepto: /d' "$TMPDIR/h451/evals/plazos-y-territorio/07-cobertura-liquidacion-tributaria.yaml" "$TMPDIR/h451/evals/plazos-y-territorio/08-cobertura-multa-de-trafico.yaml" "$TMPDIR/h451/evals/plazos-y-territorio/12-cobertura-seguridad-social-sin-leer-la-da-primera.yaml"
go -C "$TMPDIR/h451" test -count=1 -v -run '^TestEvalsDelRepositorio$/^precepto-de-la-cobertura$' ./internal/evals/
git -C "$TMPDIR/h451" checkout -- evals/
```

**Cada test, devuelto a `main`** (SC-005; FR-012): el guion en rojo; `tests enteros: 1` con los dos primeros y
`subtests: 1` con los dos últimos, con el test nombrado en su registro.

```bash
git -C "$TMPDIR/h451" checkout origin/main -- internal/skills/cobertura_test.go
go -C "$TMPDIR/h451" test -count=1 -v -run '^TestEntregaDelHito$/^h45-paquete-de-mas$' ./internal/app/
git -C "$TMPDIR/h451" checkout HEAD -- internal/skills/cobertura_test.go
sed -i 's/lineaDeLaOrdenDelPasoDos(t), bashDelPasoDos, sinSuForma/36, bashDelPasoDos, sinSuForma/' "$TMPDIR/h451/internal/evals/conjunto_test.go"
go -C "$TMPDIR/h451" test -count=1 -v -run '^TestEntregaDelHito$/^h45-paquete-de-mas$' ./internal/app/
git -C "$TMPDIR/h451" checkout -- internal/evals/conjunto_test.go
git -C "$TMPDIR/h451" checkout origin/main -- internal/evals/definicion_test.go
go -C "$TMPDIR/h451" test -count=1 -v -run '^TestEntregaDelHito$/^h45-paquete-de-mas$' ./internal/app/
git -C "$TMPDIR/h451" checkout HEAD -- internal/evals/definicion_test.go
sed -i 's/ && !eval.NoCubierto.FueraDeLoDeclarado$//' "$TMPDIR/h451/internal/evals/conjunto_test.go"
go -C "$TMPDIR/h451" test -count=1 -v -run '^TestEntregaDelHito$/^h45-paquete-de-mas$' ./internal/app/
git -C "$TMPDIR/h451" checkout -- internal/evals/conjunto_test.go
```

**La copia congelada distinta** (FR-013): el guion en rojo, «… differ».

```bash
printf '\n' >> "$TMPDIR/h451/specs/042-h45-el-molde-admite/aceptacion/paquete-de-mas.txtar"
go -C "$TMPDIR/h451" test -count=1 -v -run '^TestEntregaDelHito$/^h45-paquete-de-mas$' ./internal/app/
git -C "$TMPDIR/h451" checkout -- specs/
```

Al terminar:

```bash
rm -rf "$TMPDIR/h451"
```
