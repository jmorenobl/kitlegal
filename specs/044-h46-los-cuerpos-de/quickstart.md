# Quickstart: validar H46

Escenarios para la cabeza de la rama `044-h46-los-cuerpos-de` con todas las tareas hechas y la suite activada. Todas las
órdenes van desde la raíz del repositorio. Q1 a Q8 solo leen el árbol, salvo lo que dicen; Q9 trabaja en un clon en
`$TMPDIR` y lo borra al terminar. Ninguno toca el índice ni el historial.

## Q1 · `make ci` en verde, con `h45-paquete-de-mas` (SC-001; FR-008)

```sh
make ci
```

Esperado: termina con `ci: todos los controles en verde`. Deja `coverage.out` y `coverage-integration.out`, que Git
ignora (research V13); `vuln` consulta su base por red.

## Q2 · La sincronía no cambia nada y `skills-check` no da defectos (US1.2; FR-004, FR-007; SC-002)

```sh
make skills-sync && git status --porcelain -- skills/
make skills-check
```

Esperado: la primera no imprime ninguna ruta (con `skills/` sincronizado, la sincronía no deja ninguna diferencia que
Git vea); la segunda, `ok` en `./internal/app/`, `./internal/skills/` y `./internal/evals/`.

## Q3 · Los cuerpos, medidos como el gate (SC-002; FR-007; las cifras de FR-017)

```sh
for s in normativa plazos-y-territorio jurisprudencia; do
  printf '%s ' "$s"
  awk 'n>=2{print} /^---$/{n++}' "skills/$s/SKILL.md" | wc -m | tr '\n' ' '
  awk 'n>=2{print} /^---$/{n++}' "skills/$s/SKILL.md" \
    | awk '/^<!-- fin de los paquetes -->$/{d=0} !d{print} /^<!-- paquetes: lo genera make skills-sync -->$/{d=1}' | wc -m
done
```

Esperado, con una locale UTF-8 (research V15): `normativa 9498 9498`, `plazos-y-territorio 9577 9483` y
`jurisprudencia 9473 9473`: el cuerpo y el cuerpo sin las líneas de los paquetes, todos ≤ 9600 y ≤ 11 000. Son las
cifras que `docs/SKILLS.md` §3 da al cerrar.

## Q4 · `skills/` es el del prototipo (US3.7; FR-010)

```sh
git diff --quiet f89093dd -- skills/ && echo igual
```

Esperado: `igual`.

## Q5 · La copia congelada de H41.1 es igual a su guion (US4.4; FR-016)

```sh
cmp specs/035-h41-1-plazos-y-territorio/aceptacion/skills-instaladas.txtar \
  internal/app/testdata/script/h41-1-skills-instaladas.txtar && echo igual
```

Esperado: `igual`.

## Q6 · La suite y los guiones que fijan el texto (US1.1, US3.1-US3.6, US4.4; FR-016, FR-021; SC-006)

```sh
go test -count=1 -v -run '^TestEntregaDelHito$/^(h46-tabla-de-tres-columnas|h46-textos-en-su-referencia|h41-skills-instaladas|h41-1-skills-instaladas|h45-paquete-de-mas)$' ./internal/app/ | grep -E -- '--- (PASS|FAIL)|^(ok|FAIL)'
```

Esperado: los cinco `--- PASS` y `ok`.

## Q7 · Los tests que cambian (FR-001 a FR-003, FR-008, FR-009, FR-014, FR-015; SC-003 a SC-005)

```sh
go test -count=1 -run '^(TestRenderizarTabla|TestDescripcionDeVerbo|TestRegenerarYComparar|TestCuerpoDeLaSkill|TestVerticalDeEnsayo|TestRemisionesDelRepositorio)$' ./internal/skills/
go test -count=1 -run '^TestSkillsDelProductoDesdeControl$' ./internal/evals/
go test -count=1 -run '^TestSkillsDelRepositorio$/^presupuestos$|^TestOrdenesDeLasSkillsEmpotradas$/^control$' ./internal/app/
```

Esperado: `ok` en los tres paquetes. En la tercera, cada alternativa de primer nivel va con su barra fuera de
paréntesis: `-run` no parte una barra que va entre paréntesis (research V17).

## Q8 · Qué tests y guiones cambian, y qué Go de producción (SC-007; FR-019)

```sh
git diff --name-only main HEAD -- '*_test.go' '*.txtar'
git diff --name-only main HEAD -- '*.go' ':!*_test.go'
git diff --name-only main HEAD -- schemas/ data/ evals/ laboratorio/ .github/ cmd/
```

Esperado, la primera, exactamente estas trece rutas:

```text
internal/app/skills_test.go
internal/app/testdata/script/h41-1-skills-instaladas.txtar
internal/app/testdata/script/h41-skills-instaladas.txtar
internal/app/testdata/script/h46-tabla-de-tres-columnas.txtar
internal/app/testdata/script/h46-textos-en-su-referencia.txtar
internal/evals/laboratorio_familias_test.go
internal/skills/comandos_test.go
internal/skills/remisiones_test.go
internal/skills/sincronia_test.go
internal/skills/vertical_de_ensayo_test.go
specs/035-h41-1-plazos-y-territorio/aceptacion/skills-instaladas.txtar
specs/044-h46-los-cuerpos-de/aceptacion/tabla-de-tres-columnas.txtar
specs/044-h46-los-cuerpos-de/aceptacion/textos-en-su-referencia.txtar
```

la segunda, `internal/skills/comandos.go` e `internal/skills/sincronia.go`; la tercera, nada.

## Q9 · Los mutantes, en un clon del temporal (SC-003 a SC-006; FR-001, FR-003, FR-005)

```sh
git clone -q --no-hardlinks . "$TMPDIR/h46-mut" && cd "$TMPDIR/h46-mut"

# El tope de FR-005, a 9601 y a 9599 (SC-003)
sed -i 's/maximoSinLosPaquetes\( *\)= 9600/maximoSinLosPaquetes\1= 9601/' internal/skills/sincronia.go
go test -count=1 -run '^TestSkillsDelRepositorio$/^presupuestos$' ./internal/app/ | grep -c -- '--- FAIL: .*presupuestos/[a-z-]*/'
git checkout -q -- internal/skills/sincronia.go
sed -i 's/maximoSinLosPaquetes\( *\)= 9600/maximoSinLosPaquetes\1= 9599/' internal/skills/sincronia.go
go test -count=1 -run '^TestSkillsDelRepositorio$/^presupuestos$' ./internal/app/ | grep -c -- '--- FAIL: .*presupuestos/[a-z-]*/'
git checkout -q -- internal/skills/sincronia.go

# La regla que no se aplica: sincronia.go de main, con los tests del hito (SC-003)
git show origin/main:internal/skills/sincronia.go > internal/skills/sincronia.go
go test -count=1 -run '^TestSkillsDelRepositorio$/^presupuestos$' ./internal/app/ | grep -c -- '--- FAIL: .*presupuestos/[a-z-]*/'
git checkout -q -- internal/skills/sincronia.go

# La tabla de cuatro columnas: comandos.go de main (FR-001, FR-003)
git show origin/main:internal/skills/comandos.go > internal/skills/comandos.go
go test -count=1 -v -run '^(TestRenderizarTabla|TestRegenerarYComparar|TestCuerpoDeLaSkill)$' ./internal/skills/ | grep -E -- '^--- FAIL'
git checkout -q -- internal/skills/comandos.go

# Un cambio declarado deshecho: el SKILL.md de normativa de main (SC-004)
git show origin/main:skills/normativa/SKILL.md > skills/normativa/SKILL.md
go test -count=1 -run '^TestSkillsDelProductoDesdeControl$' ./internal/evals/ | tail -1
git checkout -q -- skills/

# Las frases de main en las remisiones del plan (SC-005)
git show origin/main:internal/skills/remisiones_test.go > internal/skills/remisiones_test.go
go test -count=1 -run '^TestRemisionesDelRepositorio$' ./internal/skills/ | tail -1
git checkout -q -- internal/skills/remisiones_test.go

# La suite sobre los skills/ de main (SC-006)
git checkout -q origin/main -- skills/
go test -count=1 -v -run '^TestEntregaDelHito$/^h46-' ./internal/app/ | grep -E -- '--- FAIL: TestEntregaDelHito/'
git checkout -q HEAD -- skills/

cd - >/dev/null && rm -rf "$TMPDIR/h46-mut"
```

Esperado: `6`, `9` y `6` fallos en `presupuestos` (research P1: con 9601, los casos de 9601 y de 11 001 de las tres
skills con marcas; con 9599, además los de 9600; sin la regla, los de 9601 y de 11 001); `--- FAIL` en
`TestRenderizarTabla`, `TestRegenerarYComparar` y `TestCuerpoDeLaSkill`; `FAIL` en `TestSkillsDelProductoDesdeControl`
y en `TestRemisionesDelRepositorio`; y los dos guiones `h46-` en `--- FAIL`. En el clon, `main` es `origin/main`
(research V14). El registro de cada mutante de la tabla del plan, uno por entrada de FR-014 y una frase cada vez en
SC-005, lo escribe T006 en `mutantes.md`.
